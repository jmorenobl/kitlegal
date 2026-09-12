package cache

import (
	"bytes"
	"context"
	"encoding/binary"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// instanteDePrueba es el «ahora» del que parte todo reloj inyectado de estas
// tablas. Es fijo, y no el de la máquina, para que el instante de expiración que
// se escribe sea siempre el mismo y se pueda comparar con el de la fila.
var instanteDePrueba = time.Date(2026, time.September, 12, 0, 0, 0, 0, time.UTC)

// vigenciaDePrueba es una hora, la del adaptador de prueba. Ninguna tabla la
// espera: el reloj inyectado salta al instante que haga falta, de modo que la
// duración de una prueba no depende de la vigencia que use (SC-002).
const vigenciaDePrueba = time.Hour

// posicionesDelBorde son las tres posiciones de SC-002 alrededor del instante en
// que caduca una entrada escrita en instanteDePrueba con vigenciaDePrueba: un
// nanosegundo antes, el instante exacto y un nanosegundo después. Un nanosegundo
// es la resolución de expira_en, así que ninguna comparación más gruesa que la
// del contrato —ni la de «menor o igual», ni la de segundos— puede dar las tres
// bien (FR-008).
var posicionesDelBorde = []struct {
	nombre  string
	lectura time.Time
	vigente bool
}{
	{
		nombre:  "antes del instante de expiración",
		lectura: instanteDePrueba.Add(vigenciaDePrueba - time.Nanosecond),
		vigente: true,
	},
	{
		nombre:  "en el instante exacto de expiración",
		lectura: instanteDePrueba.Add(vigenciaDePrueba),
		vigente: false,
	},
	{
		nombre:  "después del instante de expiración",
		lectura: instanteDePrueba.Add(vigenciaDePrueba + time.Nanosecond),
		vigente: false,
	},
}

// Las dos consultas con las que estas tablas miran la tabla de entradas por
// debajo del cliente, que no deja ejecutar ninguna sentencia desde fuera
// (FR-005).
const (
	entradasDeLaBase = `SELECT count(*) FROM entradas`
	filaDeLaClave    = `SELECT count(*), COALESCE(MAX(expira_en), 0) FROM entradas WHERE clave = ?`
)

// TestPutYGetIntegros fija FR-012 y la primera mitad de SC-010: lo guardado se
// devuelve byte a byte igual, sea texto, binario que no es UTF-8 y lleva bytes
// nulos, grande o de cero bytes, y bajo una clave de cualquier longitud.
//
// Las filas de cero bytes y de contenido nulo miden además lo que «byte a byte»
// no alcanza: bytes.Equal da por iguales un nil y un []byte{}, y por eso cada
// fila exige también que lo devuelto no sea nulo. Un contenido vacío guardado es
// un valor presente y vacío, no una ausencia, aunque el controlador lo entregue
// como nil (research D8, sonda 1 C).
func TestPutYGetIntegros(t *testing.T) {
	t.Parallel()

	const mebibyte = 1 << 20

	patron := []byte("<norma>contenido</norma>\x00\xff\x80")
	grande := bytes.Repeat(patron, mebibyte/len(patron)+1)[:mebibyte]

	casos := []struct {
		nombre    string
		clave     string
		contenido []byte
	}{
		{nombre: "texto", clave: claveDePrueba, contenido: []byte("<norma>contenido</norma>")},
		{
			nombre:    "binario con bytes altos",
			clave:     claveDePrueba,
			contenido: []byte{0x00, 0x01, 0x7f, 0x80, 0xc3, 0x28, 0xfe, 0xff, 0x00},
		},
		{nombre: "cero bytes", clave: claveDePrueba, contenido: []byte{}},
		{nombre: "nulo, que se guarda como cero bytes", clave: claveDePrueba, contenido: nil},
		{nombre: "un mebibyte", clave: claveDePrueba, contenido: grande},
		{
			nombre:    "clave larga",
			clave:     claveDePrueba + "/" + strings.Repeat("norma/", 8192),
			contenido: []byte("<norma>contenido</norma>"),
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			cliente := clienteAbierto(t, t.TempDir())
			require.NoError(t, cliente.Put(t.Context(), caso.clave, caso.contenido, vigenciaDePrueba))

			leido, presente, err := cliente.Get(t.Context(), caso.clave)

			require.NoError(t, err)
			require.True(t, presente, "lo recién guardado y vigente está presente")
			require.NotNil(t, leido, "presente y vacío no es ausente: lo leído nunca es nulo (FR-012)")
			assert.Len(t, leido, len(caso.contenido))
			assert.True(t, bytes.Equal(caso.contenido, leido), "lo leído es byte a byte lo guardado")
		})
	}
}

// TestGetAusente fija FR-013 fuera del modo de solo lectura: una clave que nunca
// se guardó es una ausencia —nil, false, nil— y no un fallo, que es lo que lleva
// a quien llama a pedirla a la fuente sin comparar ningún error. El registro la
// anota como «ausencia», que es lo que la separa de una entrada que estaba y
// caducó (D14).
func TestGetAusente(t *testing.T) {
	t.Parallel()

	registrador, registro := registroEnMemoria()
	cliente := clienteAbierto(t, t.TempDir(), ConRegistrador(registrador))

	contenido, presente, err := cliente.Get(t.Context(), claveDePrueba)

	require.NoError(t, err, "fuera de solo lectura la ausencia no es un fallo (FR-013, FR-014)")
	assert.False(t, presente)
	assert.Nil(t, contenido, "la ausencia no trae contenido")
	assert.Contains(t, registro.String(), "resultado=ausencia")
}

// TestPutSustituyeLaEntrada fija FR-007 y el escenario 5 de US1: escribir una
// clave que ya está sustituye por completo contenido **y** vigencia, sin dejar
// rastro de la anterior ni repetir la clave.
//
// Las dos mitades se miden por separado porque cada una tiene su forma de
// fallar. El contenido nuevo es más corto que el anterior, de modo que una
// sustitución a medias dejaría la cola del primero. Y la vigencia nueva es más
// corta, de modo que con el reloj en medio de las dos una vigencia que no se
// hubiera sustituido seguiría sirviendo la entrada.
func TestPutSustituyeLaEntrada(t *testing.T) {
	t.Parallel()

	reloj := relojEn(instanteDePrueba)
	directorio := t.TempDir()
	ruta := filepath.Join(directorio, ficheroDeLaBase)
	cliente := clienteAbierto(t, directorio, ConReloj(reloj.Ahora))

	primera := []byte("<norma>la primera versión, que es más larga que la segunda</norma>")
	segunda := []byte("<norma>la segunda</norma>")

	require.NoError(t, cliente.Put(t.Context(), claveDePrueba, primera, 2*vigenciaDePrueba))
	require.NoError(t, cliente.Put(t.Context(), claveDePrueba, segunda, vigenciaDePrueba/2))

	leido, presente, err := cliente.Get(t.Context(), claveDePrueba)
	require.NoError(t, err)
	require.True(t, presente)
	assert.Equal(t, segunda, leido, "la lectura devuelve la nueva, sin restos de la anterior")

	filas, expiraEn := entradaGuardada(t, ruta, claveDePrueba)
	assert.Equal(t, int64(1), filas, "la clave no se repite")
	assert.Equal(t, instanteDePrueba.Add(vigenciaDePrueba/2).UnixNano(), expiraEn,
		"la expiración es la del reloj al escribir más la vigencia nueva (FR-006)")

	reloj.Pon(instanteDePrueba.Add(vigenciaDePrueba))

	leido, presente, err = cliente.Get(t.Context(), claveDePrueba)
	require.NoError(t, err)
	assert.False(t, presente, "la vigencia también se sustituye: la anterior aún no habría caducado")
	assert.Nil(t, leido)
}

// TestClavesIndependientes fija la segunda mitad de SC-010 y el escenario 4 de
// US1: dos claves distintas se guardan y se sirven por separado, guardar una no
// responde por la otra y caducar una no afecta a la otra.
//
// Las claves se parecen tanto como pueden sin ser la misma —una barra de más,
// otra caja, un espacio al final, los dos comodines de LIKE— y las que se leen
// sin haberlas guardado son prefijos u otras formas de las guardadas: cualquier
// comparación que no fuera la igualdad exacta de la clave opaca confundiría
// alguna (FR-011).
func TestClavesIndependientes(t *testing.T) {
	t.Parallel()

	reloj := relojEn(instanteDePrueba)
	cliente := clienteAbierto(t, t.TempDir(), ConReloj(reloj.Ahora))

	guardadas := []string{
		claveDePrueba,
		claveDePrueba + "/",
		claveDePrueba + " ",
		"prueba:http://fuente.prueba/Norma",
		"prueba:http://fuente.prueba/nor_a",
		"prueba:http://fuente.prueba/nor%",
	}

	// La primera caduca antes que las demás.
	for indice, clave := range guardadas {
		vigencia := vigenciaDePrueba
		if indice == 0 {
			vigencia = vigenciaDePrueba / 4
		}

		require.NoError(t, cliente.Put(t.Context(), clave, []byte("contenido de "+clave), vigencia))
	}

	for _, clave := range guardadas {
		leido, presente, err := cliente.Get(t.Context(), clave)
		require.NoError(t, err)
		require.True(t, presente, "%q está guardada", clave)
		assert.Equal(t, []byte("contenido de "+clave), leido, "cada clave responde con lo suyo (SC-010)")
	}

	for _, clave := range []string{
		"prueba:http://fuente.prueba/nor",
		"prueba:http://fuente.prueba/NORMA",
		"prueba:http://fuente.prueba/norma/x",
	} {
		leido, presente, err := cliente.Get(t.Context(), clave)
		require.NoError(t, err)
		assert.False(t, presente, "%q no se guardó y no responde por ninguna parecida", clave)
		assert.Nil(t, leido)
	}

	reloj.Pon(instanteDePrueba.Add(vigenciaDePrueba / 2))

	_, presente, err := cliente.Get(t.Context(), guardadas[0])
	require.NoError(t, err)
	assert.False(t, presente, "la primera ha caducado")

	for _, clave := range guardadas[1:] {
		_, presente, err := cliente.Get(t.Context(), clave)
		require.NoError(t, err)
		assert.True(t, presente, "caducar %q no afecta a %q", guardadas[0], clave)
	}
}

// TestClaveVacia fija FR-011 y la fila 1 del contrato de errores: la clave vacía
// es «argumentos» (2) al leer y al escribir, en los dos modos, y el mensaje
// nombra la operación. Va por delante de todo lo demás, también del modo: en
// solo lectura una clave vacía sigue siendo un error de quien llama y no una
// ausencia (4) ni una escritura prohibida (1).
//
// La base se siembra con una entrada antes de abrirla en cada modo, de modo que
// «no se escribe nada» se mida contando filas sobre una base con esquema.
func TestClaveVacia(t *testing.T) {
	t.Parallel()

	for _, modo := range modosDeApertura {
		t.Run(modo.nombre, func(t *testing.T) {
			t.Parallel()

			directorio := t.TempDir()
			siembra(t, directorio, claveDePrueba, []byte("<norma>contenido</norma>"))

			cliente := clienteAbierto(t, directorio, modo.opciones...)

			contenido, presente, err := cliente.Get(t.Context(), "")
			require.Error(t, err)
			assert.False(t, presente)
			assert.Nil(t, contenido)
			assert.Equal(t, schema.ClaseArgumentos, cli.Clasificar(err))
			assert.Equal(t, 2, cli.CodigoSalida(err))
			assert.Contains(t, err.Error(), "leer", "el mensaje nombra la operación")

			err = cliente.Put(t.Context(), "", []byte("<norma>otra</norma>"), vigenciaDePrueba)
			require.Error(t, err)
			assert.Equal(t, schema.ClaseArgumentos, cli.Clasificar(err))
			assert.Equal(t, 2, cli.CodigoSalida(err))
			assert.Contains(t, err.Error(), "escribir", "el mensaje nombra la operación")

			assert.Equal(t, int64(1),
				consultaEntero(t, filepath.Join(directorio, ficheroDeLaBase), entradasDeLaBase),
				"no se escribe nada: solo está la entrada sembrada")
		})
	}
}

// TestVigenciaInvalida fija FR-010 y la fila 2 del contrato de errores: una
// vigencia menor o igual que cero es «argumentos» (2), el mensaje nombra la
// vigencia recibida y no se escribe nada. «No guardar esto» se expresa no
// escribiendo, no con una vigencia nula.
//
// La fila negativa es un nanosegundo por debajo de cero, el valor más cercano a
// la fila de cero que sigue siendo inválido.
func TestVigenciaInvalida(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre   string
		vigencia time.Duration
	}{
		{nombre: "cero", vigencia: 0},
		{nombre: "negativa", vigencia: -time.Nanosecond},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			directorio := t.TempDir()
			cliente := clienteAbierto(t, directorio)

			err := cliente.Put(t.Context(), claveDePrueba, []byte("<norma>contenido</norma>"), caso.vigencia)

			require.Error(t, err)
			assert.Equal(t, schema.ClaseArgumentos, cli.Clasificar(err))
			assert.Equal(t, 2, cli.CodigoSalida(err))
			assert.Contains(t, err.Error(), caso.vigencia.String(), "el mensaje nombra la vigencia recibida")

			assert.Equal(t, int64(0),
				consultaEntero(t, filepath.Join(directorio, ficheroDeLaBase), entradasDeLaBase),
				"una vigencia inválida no escribe nada")

			_, presente, err := cliente.Get(t.Context(), claveDePrueba)
			require.NoError(t, err)
			assert.False(t, presente)
		})
	}
}

// TestExpiracionConRelojInyectado es el control literal del hito sobre la
// vigencia (SC-002, FR-008, FR-042): una entrada guardada con una hora de
// vigencia se lee en las tres posiciones del borde moviendo el reloj inyectado,
// sin esperar tiempo real, y solo la anterior al instante de expiración la
// sirve.
//
// Cada posición comprueba además, por debajo del cliente, que la fila sigue en
// la tabla con la expiración que se escribió: la ausencia en el instante exacto
// es la de una entrada caducada, no la de una que desapareció. El registro lo
// dice con «expirada», que es lo que distingue una cosa de la otra desde fuera
// (D6, D14).
func TestExpiracionConRelojInyectado(t *testing.T) {
	t.Parallel()

	contenido := []byte("<norma>contenido</norma>")

	for _, posicion := range posicionesDelBorde {
		t.Run(posicion.nombre, func(t *testing.T) {
			t.Parallel()

			reloj := relojEn(instanteDePrueba)
			registrador, registro := registroEnMemoria()
			directorio := t.TempDir()
			cliente := clienteAbierto(t, directorio, ConReloj(reloj.Ahora), ConRegistrador(registrador))

			require.NoError(t, cliente.Put(t.Context(), claveDePrueba, contenido, vigenciaDePrueba))

			reloj.Pon(posicion.lectura)

			leido, presente, err := cliente.Get(t.Context(), claveDePrueba)

			require.NoError(t, err, "fuera de solo lectura lo caducado es ausencia y no un fallo (FR-013)")
			assert.Equal(t, posicion.vigente, presente)

			if posicion.vigente {
				assert.Equal(t, contenido, leido)
				assert.Contains(t, registro.String(), "resultado=acierto")
			} else {
				assert.Nil(t, leido, "lo caducado no se sirve")
				assert.Contains(t, registro.String(), "resultado=expirada")
			}

			filas, expiraEn := entradaGuardada(t, filepath.Join(directorio, ficheroDeLaBase), claveDePrueba)
			assert.Equal(t, int64(1), filas, "leer lo caducado no borra la fila (D6)")
			assert.Equal(t, instanteDePrueba.Add(vigenciaDePrueba).UnixNano(), expiraEn)
		})
	}
}

// TestPutSobreEntradaCaducada fija que una entrada caducada se sustituye como
// cualquier otra: la nueva escritura es vigente con su propia vigencia, y cuando
// esta caduca a su vez lo que se lee es una ausencia y no la entrada anterior,
// que no reaparece (FR-007, FR-008, D6).
func TestPutSobreEntradaCaducada(t *testing.T) {
	t.Parallel()

	reloj := relojEn(instanteDePrueba)
	directorio := t.TempDir()
	ruta := filepath.Join(directorio, ficheroDeLaBase)
	cliente := clienteAbierto(t, directorio, ConReloj(reloj.Ahora))

	caducada := []byte("<norma>la versión que caduca</norma>")
	nueva := []byte("<norma>la versión nueva</norma>")

	require.NoError(t, cliente.Put(t.Context(), claveDePrueba, caducada, vigenciaDePrueba))
	reloj.Pon(instanteDePrueba.Add(2 * vigenciaDePrueba))

	_, presente, err := cliente.Get(t.Context(), claveDePrueba)
	require.NoError(t, err)
	require.False(t, presente, "la primera ha caducado antes de escribir la nueva")

	filas, _ := entradaGuardada(t, ruta, claveDePrueba)
	require.Equal(t, int64(1), filas, "la caducada sigue en la tabla hasta que se sustituye (D6)")

	require.NoError(t, cliente.Put(t.Context(), claveDePrueba, nueva, vigenciaDePrueba))

	leido, presente, err := cliente.Get(t.Context(), claveDePrueba)
	require.NoError(t, err)
	require.True(t, presente, "la nueva es vigente")
	assert.Equal(t, nueva, leido)

	filas, expiraEn := entradaGuardada(t, ruta, claveDePrueba)
	assert.Equal(t, int64(1), filas)
	assert.Equal(t, instanteDePrueba.Add(3*vigenciaDePrueba).UnixNano(), expiraEn,
		"la vigencia cuenta desde la nueva escritura, no desde la caducada")

	reloj.Pon(instanteDePrueba.Add(3 * vigenciaDePrueba))

	leido, presente, err = cliente.Get(t.Context(), claveDePrueba)
	require.NoError(t, err)
	assert.False(t, presente, "cuando la nueva caduca, la anterior no reaparece")
	assert.Nil(t, leido)
}

// TestContextoCancelado fija FR-003 y la fila 15 del contrato de errores: la
// caché no crea ningún contexto por su cuenta y respeta el de quien llama, y un
// contexto cancelado o vencido es «fuente no disponible» (4) al construir, al
// leer y al escribir, que es la clase con la que el kernel trata el plazo
// agotado. La causa del contexto sigue alcanzable con errors.Is.
//
// La lectura se hace sobre una entrada que está y es vigente, para que el 4 no
// pueda confundirse con ninguna ausencia. En las tres operaciones el directorio
// queda como estaba, con la huella de cada fichero: con el contexto terminado no
// se crea, no se lee y no se escribe nada.
//
// Cada operación se prepara con el contexto de la prueba y devuelve la llamada
// que recibe el contexto terminado, de modo que nada de lo que se prepara use
// un contexto distinto del que la llamada recibe.
func TestContextoCancelado(t *testing.T) {
	t.Parallel()

	terminados := []struct {
		nombre string
		causa  error
		crea   func(t *testing.T) context.Context
	}{
		{
			nombre: "cancelado",
			causa:  context.Canceled,
			crea: func(t *testing.T) context.Context {
				t.Helper()

				ctx, cancela := context.WithCancel(t.Context())
				cancela()

				return ctx
			},
		},
		{
			nombre: "vencido",
			causa:  context.DeadlineExceeded,
			crea: func(t *testing.T) context.Context {
				t.Helper()

				ctx, cancela := context.WithDeadline(t.Context(), time.Unix(0, 0))
				t.Cleanup(cancela)

				return ctx
			},
		},
	}

	contenido := []byte("<norma>contenido</norma>")

	operaciones := []struct {
		nombre string
		// prepara deja el directorio como la operación lo necesita y devuelve
		// la llamada que se hace con el contexto terminado.
		prepara func(t *testing.T, directorio string) func(ctx context.Context) error
	}{
		{
			nombre: "New",
			prepara: func(t *testing.T, directorio string) func(ctx context.Context) error {
				t.Helper()

				return func(ctx context.Context) error {
					cliente, err := New(ctx, ConDirectorio(directorio))
					assert.Nil(t, cliente, "un fallo al construir no devuelve ningún cliente")

					return err
				}
			},
		},
		{
			nombre: "Get",
			prepara: func(t *testing.T, directorio string) func(ctx context.Context) error {
				t.Helper()

				siembra(t, directorio, claveDePrueba, contenido)
				cliente := clienteAbierto(t, directorio)

				return func(ctx context.Context) error {
					leido, presente, err := cliente.Get(ctx, claveDePrueba)
					assert.False(t, presente, "con el contexto terminado no se sirve ni lo vigente")
					assert.Nil(t, leido)

					return err
				}
			},
		},
		{
			nombre: "Put",
			prepara: func(t *testing.T, directorio string) func(ctx context.Context) error {
				t.Helper()

				cliente := clienteAbierto(t, directorio)

				return func(ctx context.Context) error {
					return cliente.Put(ctx, claveDePrueba, contenido, vigenciaDePrueba)
				}
			},
		},
	}

	for _, contexto := range terminados {
		for _, operacion := range operaciones {
			t.Run(operacion.nombre+" con el contexto "+contexto.nombre, func(t *testing.T) {
				t.Parallel()

				directorio := t.TempDir()
				opera := operacion.prepara(t, directorio)
				antes := arbolDe(t, directorio)

				err := opera(contexto.crea(t))

				require.ErrorIs(t, err, contexto.causa, "la causa del contexto sigue alcanzable")
				assert.Equal(t, schema.ClaseFuenteNoDisponible, cli.Clasificar(err))
				assert.Equal(t, 4, cli.CodigoSalida(err))
				assert.Equal(t, antes, arbolDe(t, directorio),
					"con el contexto terminado no se crea, no se lee y no se escribe nada")
			})
		}
	}
}

// TestSoloLecturaAusenciaEsFuenteNoDisponible fija FR-016 y la fila 8 del
// contrato de errores: en solo lectura una clave que no está no es la ausencia
// normal de FR-013, sino «fuente no disponible» (4), porque sin red no hay a
// dónde ir a buscarla. El control positivo es la otra clave, sembrada en la
// misma base, que el mismo lector sí sirve.
func TestSoloLecturaAusenciaEsFuenteNoDisponible(t *testing.T) {
	t.Parallel()

	const sembrada = claveDePrueba + "/sembrada"

	directorio := t.TempDir()
	contenido := []byte("<norma>contenido</norma>")
	siembra(t, directorio, sembrada, contenido)

	lector := clienteAbierto(t, directorio, SoloLectura())

	leido, presente, err := lector.Get(t.Context(), claveDePrueba)
	compruebaAusenciaEnSoloLectura(t, leido, presente, err, claveDePrueba)

	leido, presente, err = lector.Get(t.Context(), sembrada)
	require.NoError(t, err, "lo que está y es vigente se sirve también en solo lectura")
	assert.True(t, presente)
	assert.Equal(t, contenido, leido)
}

// TestSoloLecturaExpiradaEsFuenteNoDisponible fija FR-018: en solo lectura la
// vigencia se aplica igual, y una entrada caducada no se sirve por el hecho de
// no haber red, ni con aviso. En las dos posiciones del borde que no son
// vigentes el resultado es la misma «fuente no disponible» (4) que la de una
// clave que nunca se guardó; en la anterior, la entrada se sirve.
func TestSoloLecturaExpiradaEsFuenteNoDisponible(t *testing.T) {
	t.Parallel()

	contenido := []byte("<norma>contenido</norma>")

	for _, posicion := range posicionesDelBorde {
		t.Run(posicion.nombre, func(t *testing.T) {
			t.Parallel()

			directorio := t.TempDir()
			siembra(t, directorio, claveDePrueba, contenido, ConReloj(relojEn(instanteDePrueba).Ahora))

			lector := clienteAbierto(t, directorio, SoloLectura(), ConReloj(relojEn(posicion.lectura).Ahora))

			leido, presente, err := lector.Get(t.Context(), claveDePrueba)

			if !posicion.vigente {
				compruebaAusenciaEnSoloLectura(t, leido, presente, err, claveDePrueba)

				return
			}

			require.NoError(t, err)
			assert.True(t, presente)
			assert.Equal(t, contenido, leido)
		})
	}
}

// TestSoloLecturaPutFalla fija FR-017 y la fila 11 del contrato de errores:
// escribir en una caché de solo lectura es «inesperado» (1) —lo que está mal es
// el código que lo intenta, no la invocación—, el mensaje nombra la clave y no
// se escribe nada: el SHA-256 de cache.db es el mismo antes y después, y la
// entrada que había se sigue leyendo tal cual.
func TestSoloLecturaPutFalla(t *testing.T) {
	t.Parallel()

	directorio := t.TempDir()
	ruta := filepath.Join(directorio, ficheroDeLaBase)
	original := []byte("<norma>la que había</norma>")
	siembra(t, directorio, claveDePrueba, original)

	antes := huella(t, ruta)
	lector := clienteAbierto(t, directorio, SoloLectura())

	err := lector.Put(t.Context(), claveDePrueba, []byte("<norma>la que no se escribe</norma>"), vigenciaDePrueba)

	require.Error(t, err)
	assert.Equal(t, schema.ClaseInesperado, cli.Clasificar(err))
	assert.Equal(t, 1, cli.CodigoSalida(err))
	assert.Contains(t, err.Error(), claveDePrueba, "el mensaje nombra la clave")
	assert.Contains(t, err.Error(), "solo lectura")

	leido, presente, err := lector.Get(t.Context(), claveDePrueba)
	require.NoError(t, err)
	assert.True(t, presente)
	assert.Equal(t, original, leido, "la entrada que había sigue siendo la que había")

	require.NoError(t, lector.Close())
	assert.Equal(t, antes, huella(t, ruta), "una escritura en solo lectura no cambia ni un byte (SC-003)")
}

// TestFalloDelControladorAlOperar fija la fila 14 del contrato de errores en su
// forma sobrevenida: la base se abre y se migra sin fallo, pero las páginas de
// la tabla de entradas están estropeadas, y leer o escribir una clave falla en
// el controlador con SQLITE_CORRUPT. Es «inesperado» (1) nombrando la
// operación, la clave y la ruta —nunca «fuente no disponible», porque no hay
// plazo ni ausencia, ni un bloqueo, porque nadie retiene nada—, la causa del
// controlador sigue alcanzable con errors.As, y el fichero no se borra ni se
// rehace (FR-028, FR-033, FR-035).
//
// Se estropea desde la tercera página: las dos primeras —sqlite_master y
// schema_version— siguen enteras, así que New abre y comprueba la versión sin
// tropezar, y es la operación la que encuentra lo que hay más allá. Un fallo
// del controlador que sobreviene no exige estropear ningún disco: basta un
// cache.db sobrescrito en su t.TempDir(), como el de TestFicheroInutilizable.
func TestFalloDelControladorAlOperar(t *testing.T) {
	t.Parallel()

	directorio := t.TempDir()
	ruta := filepath.Join(directorio, ficheroDeLaBase)
	siembra(t, directorio, claveDePrueba, []byte("<norma>contenido</norma>"))
	corrompeDesdeLaPagina(t, ruta, primeraPaginaDeEntradas)

	antes := huella(t, ruta)
	cliente := clienteAbierto(t, directorio)

	operaciones := []struct {
		nombre string
		opera  func() error
	}{
		{
			nombre: "leer",
			opera: func() error {
				leido, presente, err := cliente.Get(t.Context(), claveDePrueba)
				assert.False(t, presente, "una lectura que falla no presenta nada")
				assert.Nil(t, leido)

				return err
			},
		},
		{
			nombre: "escribir",
			opera: func() error {
				return cliente.Put(t.Context(), claveDePrueba, []byte("<norma>otra</norma>"), vigenciaDePrueba)
			},
		},
	}

	for _, operacion := range operaciones {
		err := operacion.opera()

		require.Error(t, err, "%s sobre páginas estropeadas falla", operacion.nombre)
		assert.Equal(t, schema.ClaseInesperado, cli.Clasificar(err), "%s: la clase", operacion.nombre)
		assert.Equal(t, 1, cli.CodigoSalida(err), "%s: un fallo del controlador es la fila 14: %v", operacion.nombre, err)
		assert.NotContains(t, err.Error(), "bloqueada", "%s: nadie retiene ningún bloqueo", operacion.nombre)

		for _, dato := range []string{operacion.nombre, claveDePrueba, ruta} {
			assert.Contains(t, err.Error(), dato, "%s: el mensaje nombra la operación, la clave y la ruta", operacion.nombre)
		}

		var delControlador *sqlite.Error

		require.ErrorAs(t, err, &delControlador, "%s: la causa es la del controlador", operacion.nombre)
		assert.Equal(t, sqlite3.SQLITE_CORRUPT, delControlador.Code()&0xff, "%s: que no pudo leer las páginas", operacion.nombre)

		var fallo *Error

		require.ErrorAs(t, err, &fallo)
		assert.Equal(t, claveDePrueba, fallo.Clave, "%s: la clave viaja también en el campo", operacion.nombre)
		assert.Equal(t, ruta, fallo.Ruta, "%s: y la ruta", operacion.nombre)
	}

	require.NoError(t, cliente.Close())
	require.FileExists(t, ruta)
	assert.Equal(t, antes, huella(t, ruta), "un fichero estropeado no se borra ni se rehace (FR-028)")
}

// compruebaAusenciaEnSoloLectura es lo que toda ausencia en solo lectura tiene
// que cumplir, venga de donde venga —sin fila, caducada, sin esquema o sin
// base—: ni contenido ni presencia, «fuente no disponible» (4) y nunca el 2 de
// una ruta inservible, y un mensaje que nombra la clave y dice que la invocación
// solo lee (FR-016, FR-018, fila 8 del contrato de errores). La clave viaja
// también en el campo del error, que es como la recupera quien no analiza
// cadenas.
func compruebaAusenciaEnSoloLectura(t *testing.T, contenido []byte, presente bool, err error, clave string) {
	t.Helper()

	require.Error(t, err, "en solo lectura la ausencia es un fallo: no hay fuente a la que ir (FR-016)")
	assert.False(t, presente)
	assert.Nil(t, contenido)
	assert.Equal(t, schema.ClaseFuenteNoDisponible, cli.Clasificar(err))
	assert.Equal(t, 4, cli.CodigoSalida(err), "código 4, y nunca el 2 de una ruta inservible")
	assert.Contains(t, err.Error(), clave, "el mensaje nombra la clave")
	assert.Contains(t, err.Error(), "solo lectura", "…y que la invocación solo lee")

	var fallo *Error
	require.ErrorAs(t, err, &fallo)
	assert.Equal(t, clave, fallo.Clave)
}

// clienteAbierto construye un cliente sobre el directorio y registra su cierre
// al terminar la prueba. Cerrar es idempotente, así que la prueba puede cerrarlo
// antes por su cuenta sin que la limpieza falle (FR-004).
func clienteAbierto(t *testing.T, directorio string, opciones ...Opcion) *Cliente {
	t.Helper()

	cliente, err := New(t.Context(), append([]Opcion{ConDirectorio(directorio)}, opciones...)...)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, cliente.Close()) })

	return cliente
}

// siembra deja una entrada guardada, con vigenciaDePrueba, en la base del
// directorio, y cierra el cliente antes de volver: al cerrarse el último cliente
// normal, SQLite consolida el registro de escritura en cache.db, de modo que lo
// que venga después —un lector de solo lectura, una huella— encuentre la base
// como la deja una invocación que ya terminó.
func siembra(t *testing.T, directorio, clave string, contenido []byte, opciones ...Opcion) {
	t.Helper()

	escritor := clienteAbierto(t, directorio, opciones...)
	require.NoError(t, escritor.Put(t.Context(), clave, contenido, vigenciaDePrueba))
	require.NoError(t, escritor.Close())
}

// primeraPaginaDeEntradas es la primera página de cache.db que no pertenece ni
// a sqlite_master ni a schema_version: la migración v1 crea schema_version
// antes que entradas, así que sqlite_master ocupa la primera página, la
// versión la segunda, y la tabla de entradas y el índice de su clave primaria
// vienen después. Estropear desde ella deja la apertura y la comprobación de la
// versión intactas y hace fallar solo las operaciones.
const primeraPaginaDeEntradas = 3

// corrompeDesdeLaPagina sobrescribe con 0xFF todo cache.db desde la página dada
// hasta el final, con el tamaño de página que declara la propia cabecera del
// fichero —los bytes 16 y 17, en big-endian—, de modo que la prueba no dependa
// del tamaño por omisión del controlador. Lo anterior sigue intacto, cabecera
// incluida: el fichero se sigue reconociendo como una base de datos y se deja
// leer; lo que no se puede es usar lo que hay en esas páginas.
func corrompeDesdeLaPagina(t *testing.T, ruta string, pagina int) {
	t.Helper()

	const (
		cabecera            = 100
		posicionDelTamano   = 16
		bytesDelTamano      = 2
		primeraPaginaValida = 1
		relleno             = 0xFF
	)

	contenido := leeLaBase(t, ruta)
	require.Greater(t, len(contenido), cabecera, "la cabecera de SQLite ocupa cien bytes")
	require.GreaterOrEqual(t, pagina, primeraPaginaValida)

	bytesPorPagina := int(binary.BigEndian.Uint16(contenido[posicionDelTamano : posicionDelTamano+bytesDelTamano]))
	desde := (pagina - 1) * bytesPorPagina
	require.Less(t, desde, len(contenido), "la página %d existe en el fichero", pagina)

	estropeado := bytes.Repeat([]byte{relleno}, len(contenido))
	copy(estropeado, contenido[:desde])

	require.NoError(t, os.WriteFile(filepath.Clean(ruta), estropeado, permisosDelFichero))
}

// entradaGuardada mira la fila de una clave por debajo del cliente: cuántas hay
// —ninguna o una, porque la clave es la clave primaria— y con qué instante de
// expiración, en nanosegundos Unix. La clave va como parámetro de la consulta y
// no dentro de ella.
func entradaGuardada(t *testing.T, ruta, clave string) (filas, expiraEn int64) {
	t.Helper()

	require.NoError(t, paraLeer(t, ruta).QueryRowContext(t.Context(), filaDeLaClave, clave).
		Scan(&filas, &expiraEn))

	return filas, expiraEn
}

// relojDePrueba es un reloj que la prueba mueve a mano: la vigencia se decide
// sin esperar tiempo real (FR-009, SC-002). Lleva cerrojo porque quien lo mueve
// y quien lo lee no tienen por qué ser la misma goroutine.
type relojDePrueba struct {
	mu    sync.Mutex
	ahora time.Time
}

func relojEn(instante time.Time) *relojDePrueba {
	return &relojDePrueba{ahora: instante}
}

// Ahora es la función que se inyecta con ConReloj.
func (r *relojDePrueba) Ahora() time.Time {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.ahora
}

// Pon lleva el reloj a un instante.
func (r *relojDePrueba) Pon(instante time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.ahora = instante
}

// registroEnMemoria devuelve un registrador de nivel debug que escribe en un
// búfer, para mirar qué resultado anotó la caché (D14). El búfer no es seguro
// entre goroutines, y no hace falta: cada prueba que lo usa opera desde una.
func registroEnMemoria() (*slog.Logger, *bytes.Buffer) {
	registro := &bytes.Buffer{}

	return slog.New(slog.NewTextHandler(registro, &slog.HandlerOptions{Level: slog.LevelDebug})), registro
}
