// adaptador_test.go es el adaptador de prueba del hito y la demostración, con el
// kernel entero en proceso, de lo que la caché promete a quien la usará de
// verdad: que la segunda consulta idéntica no vuelve a pedir nada y que
// --offline, que hasta este hito se aceptaba sin cambiar nada, sirve lo guardado
// o termina con «fuente no disponible» sin tocar el fichero (FR-045 a FR-047,
// SC-001, SC-003).
//
// Es **material de test** y nada más: vive en un fichero _test.go, no se crea en
// el directorio reservado a los adaptadores de fuente —que este hito no abre— y
// no se registra en ningún binario, de modo que el conjunto de verbos del binario
// que se publica queda como lo dejó H2 (FR-044). El único registro en el que
// aparece lo construye este test.
//
// Va en el paquete externo cache_test porque compila solo contra la superficie
// exportada de la caché, que es la que tendrá cada adaptador de fuente (research
// D12). Y pide siempre por httpx.Replay, nunca contra un servidor local: la regla
// R2 deniega la biblioteca HTTP en todo fichero fuera de internal/httpx, los de
// test incluidos y por prefijo. La reproducción es además estricta —una petición
// sin grabación termina en un fallo que la nombra—, y eso es lo que convierte «no
// se pidió nada» en algo que se puede comprobar (FR-046).
package cache_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/app"
	"github.com/jmorenobl/kitlegal/internal/cache"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/httpx"
)

// fuenteDePrueba es a la vez el nombre del applet, el de la fuente que cita el
// sobre de lo consultado y el prefijo de la clave de caché: el adaptador de una
// fuente se invoca por ella y guarda bajo su nombre.
const fuenteDePrueba = "prueba"

// La única dirección que el adaptador consulta —en el host ficticio de la
// grabación escrita a mano— y el contenido que esa grabación entrega.
const (
	direccionDeLaNorma = "http://fuente.prueba/norma"
	contenidoDeLaNorma = "<norma>contenido</norma>"
)

// vigenciaDeLaFuente es la vigencia con la que el adaptador guarda lo que la
// fuente le entrega (D12). Ningún subtest la espera: el reloj inyectado salta al
// instante que haga falta, de modo que la duración de la prueba no depende de
// ella (SC-002).
const vigenciaDeLaFuente = time.Hour

// Los orígenes con los que el adaptador declara de dónde salió el cuerpo que
// devuelve: de la fuente, de la caché o —en `guardar`— de sus propios argumentos.
const (
	origenFuente     = "fuente"
	origenCache      = "cache"
	origenArgumentos = "argumentos"
)

// La procedencia de lo que devuelve `guardar`. No consulta ninguna fuente, así
// que cita el espacio de nombres reservado a lo que el binario calcula, como el
// applet de ejemplo `echo` (FR-016 de H1).
const (
	fuenteDeLoGuardado = "kitlegal.prueba"
	urlDeLoGuardado    = "kitlegal:applet/prueba/guardar"
)

// ficheroDeLaBase es el fichero que la caché crea en su directorio (FR-019), y
// los otros dos son los auxiliares del registro de escritura, que un lector de
// solo lectura puede dejar junto a la base sin haber escrito nada en ella: por
// eso las comparaciones de antes y después los excluyen.
const (
	ficheroDeLaBase     = "cache.db"
	registroDeEscritura = ficheroDeLaBase + "-wal"
	memoriaCompartida   = ficheroDeLaBase + "-shm"
)

// Los datos de construcción con los que se invoca a la raíz de composición. Lo
// que importa es que el kernel los reciba y no los interprete.
const (
	nombreDelBinario = "kitlegal"
	versionDePrueba  = "dev"
	commitDePrueba   = "none"
	fechaDePrueba    = "unknown"
)

// instanteDeLaPrueba es el «ahora» de todo reloj inyectado en estas pruebas, y
// los otros dos son las dos posiciones del borde de la vigencia de lo guardado
// en él: el último instante en que sigue vigente y el primero en que ya no
// (FR-008).
var (
	instanteDeLaPrueba    = time.Date(2026, time.September, 12, 0, 0, 0, 0, time.UTC)
	ultimoInstanteVigente = instanteDeLaPrueba.Add(vigenciaDeLaFuente - time.Nanosecond)
	instanteDeExpiracion  = instanteDeLaPrueba.Add(vigenciaDeLaFuente)
)

// adaptadorDePrueba es el applet `prueba`: el patrón del contrato del puerto §8
// que copiará cada adaptador de fuente, con lo que un test necesita inyectar y
// nada más. Declara su identidad y sus dos verbos; ninguna bandera, ningún código
// de salida y ninguna forma de presentación, que hereda del kernel.
type adaptadorDePrueba struct {
	// directorio es el de la caché. Vacío, la ruta sale del entorno
	// —KITLEGAL_CACHE_DIR o el directorio de la cuenta—, que es lo que
	// TestAdaptadorConVariableDeEntorno ejercita.
	directorio string
	// reloj es el de la caché. Nulo, la caché usa el suyo; con él se siembra lo
	// caducado y se lee en el borde sin esperar tiempo real (FR-009).
	reloj func() time.Time
	// fuente es el cliente con el que se pide lo que la caché no tiene, y lo da
	// el test: la reproducción de la grabación o la estricta sobre un directorio
	// vacío.
	fuente *httpx.Cliente
	// peticiones cuenta las que el adaptador llegó a pedir a la fuente. Es la
	// medida directa de «cero peticiones»; la reproducción estricta es la otra,
	// porque cualquier petición que se emita sin grabación hace fallar la
	// invocación nombrándola.
	peticiones atomic.Int64
}

func (*adaptadorDePrueba) Nombre() string { return fuenteDePrueba }

func (*adaptadorDePrueba) Descripcion() string {
	return "Consulta una dirección pasando por la caché de internal/cache."
}

func (a *adaptadorDePrueba) Verbos() []app.Verbo {
	return []app.Verbo{
		{
			Nombre:      "consultar",
			Descripcion: "Devuelve el cuerpo de la dirección: de la caché si está vigente y, si no, de la fuente, que se guarda.",
			Argumentos:  func() app.Argumentos { return &argumentosConsultar{adaptador: a} },
			Salida:      cuerpoDeLaFuente{},
		},
		{
			Nombre:      "guardar",
			Descripcion: "Guarda un contenido bajo una clave con la vigencia de la fuente.",
			Argumentos:  func() app.Argumentos { return &argumentosGuardar{adaptador: a} },
			Salida:      cuerpoDeLaFuente{},
		},
	}
}

// abreLaCache construye la caché como la construirá cada adaptador de fuente:
// con el registrador que el kernel le entrega y en modo de solo lectura si y solo
// si la invocación lleva --offline. Es el único sitio del hito donde la bandera
// se convierte en modo, y la caché ya no deja nada que decidir después: bajo ese
// modo la ausencia llega como «fuente no disponible» y escribir es «inesperado»
// (FR-015 a FR-017, FR-047).
func (a *adaptadorDePrueba) abreLaCache(
	ctx context.Context, ejecucion schema.Contexto, registrador *slog.Logger,
) (*cache.Cliente, error) {
	opciones := []cache.Opcion{cache.ConRegistrador(registrador)}

	if a.directorio != "" {
		opciones = append(opciones, cache.ConDirectorio(a.directorio))
	}

	if a.reloj != nil {
		opciones = append(opciones, cache.ConReloj(a.reloj))
	}

	if ejecucion.Offline {
		opciones = append(opciones, cache.SoloLectura())
	}

	return cache.New(ctx, opciones...)
}

// claveDe es la clave de caché de una dirección: el nombre de la fuente delante,
// para que dos fuentes no se respondan la una por la otra (SC-010). La caché no la
// interpreta.
func claveDe(direccion string) string { return fuenteDePrueba + ":" + direccion }

// cuerpoDeLaFuente es el contenido de data de los dos verbos: el cuerpo y de
// dónde salió. Se declara como tipo y no como un mapa suelto porque de él sale la
// mitad de salida del esquema que emite --describe.
type cuerpoDeLaFuente struct {
	Cuerpo string `json:"cuerpo"`
	Origen string `json:"origen"`
}

// argumentosConsultar son los argumentos del verbo `consultar <dirección>`, con
// las etiquetas de las que el kernel construye la gramática. El adaptador no
// lleva etiqueta y no es un argumento: el analizador no alcanza los campos sin
// exportar, y la fábrica del verbo lo pone en cada valor nuevo.
type argumentosConsultar struct {
	URL string `arg:"" help:"Dirección que se consulta."`

	adaptador *adaptadorDePrueba
}

// Ejecutar es el patrón del contrato del puerto §8: mira la caché antes de pedir,
// pide solo lo que no tiene y guarda lo pedido con la vigencia de la fuente. El
// cierre de la caché se une al error de retorno, porque un cierre fallido también
// es un fallo de la invocación.
func (a *argumentosConsultar) Ejecutar(
	ctx context.Context, ejecucion schema.Contexto, registrador *slog.Logger,
) (resultado schema.Resultado, err error) {
	guardada, err := a.adaptador.abreLaCache(ctx, ejecucion, registrador)
	if err != nil {
		return schema.Resultado{}, err // ya lleva su clase
	}

	defer func() { err = errors.Join(err, guardada.Close()) }()

	clave := claveDe(a.URL)

	contenido, presente, err := guardada.Get(ctx, clave)
	if err != nil {
		return schema.Resultado{}, err // bajo --offline, la ausencia llega aquí con la clase 4
	}

	if presente {
		return schema.Resultado{
			Procedencia: schema.Procedencia{Fuente: fuenteDePrueba, URL: a.URL},
			Datos:       cuerpoDeLaFuente{Cuerpo: string(contenido), Origen: origenCache},
		}, nil
	}

	a.adaptador.peticiones.Add(1)

	respuesta, err := a.adaptador.fuente.Pedir(ctx, ejecucion, httpx.Peticion{Metodo: "GET", URL: a.URL})
	if err != nil {
		return schema.Resultado{}, fmt.Errorf("consultando %s: %w", a.URL, err) // la clase no cambia
	}

	// Bajo --dry-run la petición no se emitió y la respuesta no trae ningún
	// cuerpo: guardarla dejaría en la caché un contenido que la fuente no dio. Lo
	// que se habría pedido va al ensayo, como en el adaptador de H2 (contrato del
	// cliente de H2 §6).
	if respuesta.Ensayo {
		return schema.Resultado{
			Procedencia: schema.Procedencia{Fuente: fuenteDePrueba, URL: respuesta.URL},
			Ensayo:      []string{respuesta.Descripcion()},
		}, nil
	}

	err = guardada.Put(ctx, clave, respuesta.Cuerpo, vigenciaDeLaFuente)
	if err != nil {
		return schema.Resultado{}, err
	}

	return schema.Resultado{
		Procedencia: schema.Procedencia{Fuente: fuenteDePrueba, URL: respuesta.URL},
		Datos:       cuerpoDeLaFuente{Cuerpo: string(respuesta.Cuerpo), Origen: origenFuente},
	}, nil
}

// argumentosGuardar son los argumentos del verbo `guardar <clave> <contenido>`.
// El verbo existe para acreditar por el kernel que escribir en la caché de una
// invocación con --offline falla (FR-017, FR-047) y para sembrar entradas desde
// una invocación sin ella.
type argumentosGuardar struct {
	Clave     string `arg:"" help:"Clave bajo la que se guarda."`
	Contenido string `arg:"" help:"Contenido que se guarda."`

	adaptador *adaptadorDePrueba
}

// Ejecutar construye la caché igual que `consultar` y guarda el contenido con la
// vigencia de la fuente. En solo lectura, Put falla con «inesperado» nombrando la
// clave y sin escribir nada.
func (a *argumentosGuardar) Ejecutar(
	ctx context.Context, ejecucion schema.Contexto, registrador *slog.Logger,
) (resultado schema.Resultado, err error) {
	guardada, err := a.adaptador.abreLaCache(ctx, ejecucion, registrador)
	if err != nil {
		return schema.Resultado{}, err
	}

	defer func() { err = errors.Join(err, guardada.Close()) }()

	err = guardada.Put(ctx, a.Clave, []byte(a.Contenido), vigenciaDeLaFuente)
	if err != nil {
		return schema.Resultado{}, err
	}

	return schema.Resultado{
		Procedencia: schema.Procedencia{Fuente: fuenteDeLoGuardado, URL: urlDeLoGuardado},
		Datos:       cuerpoDeLaFuente{Cuerpo: a.Contenido, Origen: origenArgumentos},
	}, nil
}

// Las comprobaciones en tiempo de compilación del contrato del applet.
var (
	_ app.Applet     = (*adaptadorDePrueba)(nil)
	_ app.Argumentos = (*argumentosConsultar)(nil)
	_ app.Argumentos = (*argumentosGuardar)(nil)
)

// TestAdaptadorDePruebaConElKernel ejercita la caché como la ejercitará cada
// adaptador de fuente: invocando al kernel en proceso —app.Main sobre un registro
// construido aquí— con el adaptador de prueba. Es la comprobación de extremo a
// extremo que ningún guion testscript puede hacer, porque solo observa binarios y
// ninguno enlaza este applet (FR-044).
//
// Los nueve subtests son de un solo nivel y cada uno parte de su propio
// directorio, porque los escenarios 2 y 4 del quickstart cuentan sus líneas. Seis,
// y solo seis, empiezan por «offline-», que es el prefijo que filtra el escenario
// 4: ningún subtest nuevo puede llevarlo sin entrar antes en el inventario de
// tests del plan y en el esperado de ese escenario.
func TestAdaptadorDePruebaConElKernel(t *testing.T) {
	t.Parallel()

	t.Run("primera-consulta", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		adaptador := &adaptadorDePrueba{
			directorio: directorio,
			reloj:      relojEn(instanteDeLaPrueba),
			fuente:     reproduccionDeLaGrabacion(t),
		}

		invocacion := invocarAlKernel(t, adaptador, "consultar", direccionDeLaNorma, "--json")

		assert.Equal(t, cuerpoDeLaFuente{Cuerpo: contenidoDeLaNorma, Origen: origenFuente}, cuerpoServido(t, invocacion),
			"sin --offline, la caché vacía no es un fallo: la consulta sigue su curso hasta la fuente, el código es 0 "+
				"y el sobre trae lo que la fuente entregó (US1 escenario 1, US3 escenario 5, FR-014)")
		assert.Equal(t, int64(1), adaptador.peticiones.Load(), "se pidió a la fuente una vez")

		contenido, vigente := leeLaNorma(t, directorio, ultimoInstanteVigente)
		assert.True(t, vigente, "lo pedido queda guardado y vigente hasta el último instante de su vigencia (US1 escenario 1)")
		assert.Equal(t, []byte(contenidoDeLaNorma), contenido, "con el cuerpo que entregó la fuente, byte a byte (FR-012)")

		_, vigente = leeLaNorma(t, directorio, instanteDeExpiracion)
		assert.False(t, vigente,
			"y deja de estarlo en el instante en que vence la vigencia con la que se guardó, la de la fuente (FR-006, FR-008)")
	})

	t.Run("segunda-consulta-sin-red", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		siembraConLaFuente(t, directorio)

		adaptador := adaptadorSinRed(t, directorio, ultimoInstanteVigente)
		segunda := cuerpoServido(t, invocarAlKernel(t, adaptador, "consultar", direccionDeLaNorma, "--json"))

		assert.Equal(t, origenCache, segunda.Origen,
			"la misma clave dentro de su vigencia se sirve de la caché (US1 escenario 2)")
		assert.Equal(t, []byte(contenidoDeLaNorma), []byte(segunda.Cuerpo),
			"con el cuerpo idéntico byte a byte al de la primera consulta (SC-001)")
		assert.Zero(t, adaptador.peticiones.Load(),
			"y sin emitir ninguna petición: con la reproducción estricta sobre un directorio vacío, "+
				"cualquiera habría terminado en un fallo que la nombra (SC-001, FR-046)")
	})

	t.Run("sin-cache-la-reproduccion-falla", func(t *testing.T) {
		t.Parallel()

		adaptador := adaptadorSinRed(t, t.TempDir(), instanteDeLaPrueba)
		invocacion := invocarAlKernel(t, adaptador, "consultar", direccionDeLaNorma, "--json")

		// El control negativo del subtest anterior: la misma reproducción estricta
		// con la caché vacía sí recibe la petición y la hace fallar nombrándola, así
		// que aquel pasa porque no se pidió nada y no porque la reproducción lo
		// dejara pasar (US1 escenario 3).
		compruebaFallo(t, invocacion, 1, schema.ClaseInesperado, "GET "+direccionDeLaNorma)
		assert.Equal(t, int64(1), adaptador.peticiones.Load(), "la consulta se emitió: sin entrada no hay otro camino")
	})

	t.Run("offline-presente", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		siembraConLaFuente(t, directorio)
		antes := estadoDe(t, directorio)

		adaptador := adaptadorSinRed(t, directorio, ultimoInstanteVigente)
		invocacion := invocarAlKernel(t, adaptador, "consultar", direccionDeLaNorma, "--json", "--offline")

		assert.Equal(t, cuerpoDeLaFuente{Cuerpo: contenidoDeLaNorma, Origen: origenCache}, cuerpoServido(t, invocacion),
			"bajo --offline lo guardado y vigente se sirve con código 0 (US3 escenario 1)")
		compruebaSinPeticionesNiCambios(t, adaptador, directorio, antes)
	})

	t.Run("offline-ausente", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		siembra := adaptadorSinRed(t, directorio, instanteDeLaPrueba)
		sembrada := invocarAlKernel(t, siembra, "guardar", claveDe("http://fuente.prueba/otra"), "otro contenido", "--json")
		require.Equal(t, 0, sembrada.codigo, "la base existe y tiene otra entrada; salida de error: %s", sembrada.errores)

		antes := estadoDe(t, directorio)

		adaptador := adaptadorSinRed(t, directorio, instanteDeLaPrueba)
		invocacion := invocarAlKernel(t, adaptador, "consultar", direccionDeLaNorma, "--json", "--offline")

		compruebaFallo(t, invocacion, 4, schema.ClaseFuenteNoDisponible, claveDe(direccionDeLaNorma), "solo lectura")
		compruebaSinPeticionesNiCambios(t, adaptador, directorio, antes)
	})

	t.Run("offline-expirada", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		siembraConLaFuente(t, directorio)
		antes := estadoDe(t, directorio)

		adaptador := adaptadorSinRed(t, directorio, instanteDeExpiracion)
		invocacion := invocarAlKernel(t, adaptador, "consultar", direccionDeLaNorma, "--json", "--offline")

		// Lo caducado no se sirve ni sin red: se sembró con el reloj inyectado y se
		// lee en el instante exacto en que vence, sin esperar (US3 escenario 3,
		// FR-018).
		compruebaFallo(t, invocacion, 4, schema.ClaseFuenteNoDisponible, claveDe(direccionDeLaNorma))
		compruebaSinPeticionesNiCambios(t, adaptador, directorio, antes)
	})

	t.Run("offline-sin-base", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		antes := estadoDe(t, directorio)

		adaptador := adaptadorSinRed(t, directorio, instanteDeLaPrueba)
		invocacion := invocarAlKernel(t, adaptador, "consultar", direccionDeLaNorma, "--json", "--offline")

		// Un directorio sin base es una ausencia y no una ruta mal declarada: 4 y
		// nunca 2, y sin crear ni migrar nada (US3 escenario 4, FR-015).
		compruebaFallo(t, invocacion, 4, schema.ClaseFuenteNoDisponible, claveDe(direccionDeLaNorma))
		compruebaSinPeticionesNiCambios(t, adaptador, directorio, antes)
		assert.NoFileExists(t, filepath.Join(directorio, registroDeEscritura), "ni el registro de escritura")
		assert.NoFileExists(t, filepath.Join(directorio, memoriaCompartida), "ni la memoria compartida")
	})

	t.Run("offline-directorio-inexistente", func(t *testing.T) {
		t.Parallel()

		directorio := filepath.Join(t.TempDir(), "sin-crear")
		antes := estadoDe(t, directorio)

		adaptador := adaptadorSinRed(t, directorio, instanteDeLaPrueba)
		invocacion := invocarAlKernel(t, adaptador, "consultar", direccionDeLaNorma, "--json", "--offline")

		// Tampoco es 2 un directorio que no existe: sin red y sin base no hay nada
		// que servir, y el modo de solo lectura no lo crea (US3 escenario 4, FR-015).
		compruebaFallo(t, invocacion, 4, schema.ClaseFuenteNoDisponible, claveDe(direccionDeLaNorma))
		compruebaSinPeticionesNiCambios(t, adaptador, directorio, antes)
		assert.NoDirExists(t, directorio, "el directorio sigue sin existir")
	})

	t.Run("offline-guardar", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		siembraConLaFuente(t, directorio)
		antes := estadoDe(t, directorio)

		adaptador := adaptadorSinRed(t, directorio, instanteDeLaPrueba)
		invocacion := invocarAlKernel(t, adaptador, "guardar", claveDe(direccionDeLaNorma), "otro contenido",
			"--json", "--offline")

		// Escribir en la caché de una invocación de solo lectura es un defecto del
		// código que lo intenta, no de la invocación: 1 y no 2 (US3 escenario 2,
		// FR-017).
		compruebaFallo(t, invocacion, 1, schema.ClaseInesperado, claveDe(direccionDeLaNorma), "solo lectura")
		compruebaSinPeticionesNiCambios(t, adaptador, directorio, antes)

		contenido, _ := leeLaNorma(t, directorio, instanteDeLaPrueba)
		assert.Equal(t, []byte(contenidoDeLaNorma), contenido, "la entrada sigue siendo la que había: no se escribió nada")
	})
}

// TestAdaptadorConVariableDeEntorno comprueba desde el kernel que la caché vive
// donde la persona usuaria la declara con KITLEGAL_CACHE_DIR, sin tocar ninguna
// invocación, y que un valor inservible es un error de argumentos y nunca una
// caída silenciosa al directorio de la cuenta (US4, FR-022, FR-023, SC-004).
//
// No declara t.Parallel(), ni él ni sus subtests, y no es un descuido: usa
// t.Setenv, que no se puede usar en pruebas paralelas, y el entorno es justo lo
// que mide. HOME y USERPROFILE apuntan a un directorio temporal, de modo que
// ningún subtest escribe bajo la cuenta real y todos pueden comprobar que bajo
// ella no apareció nada.
func TestAdaptadorConVariableDeEntorno(t *testing.T) {
	cuenta := t.TempDir()
	t.Setenv("HOME", cuenta)
	t.Setenv("USERPROFILE", cuenta)

	t.Run("la-base-vive-bajo-la-variable", func(t *testing.T) {
		declarado := filepath.Join(t.TempDir(), "por-la-variable")
		t.Setenv(cache.VariableDirectorio, declarado)

		adaptador := &adaptadorDePrueba{fuente: reproduccionDeLaGrabacion(t)}
		invocacion := invocarAlKernel(t, adaptador, "consultar", direccionDeLaNorma, "--json")

		assert.Equal(t, origenFuente, cuerpoServido(t, invocacion).Origen)
		assert.FileExists(t, filepath.Join(declarado, ficheroDeLaBase),
			"la base se crea en el directorio que declara la variable (US4 escenario 2)")
		assert.Empty(t, estadoDe(t, cuenta), "y no bajo HOME: la variable manda sobre el directorio de la cuenta (FR-023)")
	})

	t.Run("la-variable-sin-valor", func(t *testing.T) {
		t.Setenv(cache.VariableDirectorio, "")

		adaptador := adaptadorSinRed(t, "", instanteDeLaPrueba)
		invocacion := invocarAlKernel(t, adaptador, "consultar", direccionDeLaNorma, "--json")

		compruebaFallo(t, invocacion, 2, schema.ClaseArgumentos, cache.VariableDirectorio)
		assert.Zero(t, adaptador.peticiones.Load(), "la invocación no llega a pedir nada")
		assert.Empty(t, estadoDe(t, cuenta),
			"y no cae al directorio de la cuenta: declarar la variable sin nada no es no declararla (FR-022)")
	})

	t.Run("la-variable-apunta-a-un-fichero", func(t *testing.T) {
		fichero := filepath.Join(t.TempDir(), "fichero")
		require.NoError(t, os.WriteFile(fichero, []byte("no es un directorio"), 0o600))
		t.Setenv(cache.VariableDirectorio, fichero)

		adaptador := adaptadorSinRed(t, "", instanteDeLaPrueba)
		invocacion := invocarAlKernel(t, adaptador, "consultar", direccionDeLaNorma, "--json")

		compruebaFallo(t, invocacion, 2, schema.ClaseArgumentos, cache.VariableDirectorio, fichero)
		assert.Zero(t, adaptador.peticiones.Load(), "la invocación no llega a pedir nada")
		assert.Empty(t, estadoDe(t, cuenta), "y no cae al directorio de la cuenta (FR-022)")
		assert.Equal(t, []byte("no es un directorio"), leeElFichero(t, fichero), "el fichero queda intacto")
	})
}

// relojEn es un reloj parado en un instante.
func relojEn(instante time.Time) func() time.Time {
	return func() time.Time { return instante }
}

// reproduccionDeLaGrabacion es el cliente que responde desde la única grabación
// del hito, escrita a mano contra el host ficticio: el directorio de grabaciones
// de la fuente, <raíz>/<fuente>, como lo deja la grabación de H2 (contrato de
// grabación de H2 §2). Ninguna otra petición tiene respuesta.
func reproduccionDeLaGrabacion(t *testing.T) *httpx.Cliente {
	t.Helper()

	cliente, err := httpx.Replay(filepath.Join("testdata", "reproduccion", fuenteDePrueba), httpx.ConFuente(fuenteDePrueba))
	require.NoError(t, err, "el directorio de la grabación existe y se puede reproducir")

	return cliente
}

// adaptadorSinRed es el adaptador cuyo cliente reproduce un directorio vacío: la
// reproducción estricta, en la que toda petición que se emita termina en un fallo
// de clase «inesperado» que la nombra. Una invocación con él que acaba con
// cualquier otro código no pidió nada (FR-046).
func adaptadorSinRed(t *testing.T, directorio string, ahora time.Time) *adaptadorDePrueba {
	t.Helper()

	cliente, err := httpx.Replay(t.TempDir(), httpx.ConFuente(fuenteDePrueba))
	require.NoError(t, err, "un directorio vacío se puede reproducir")

	return &adaptadorDePrueba{directorio: directorio, reloj: relojEn(ahora), fuente: cliente}
}

// siembraConLaFuente deja la norma en la caché como la deja una consulta sin
// --offline: pedida a la grabación y guardada en instanteDeLaPrueba con la
// vigencia de la fuente. Es el estado del que parten los subtests que encuentran
// algo guardado.
func siembraConLaFuente(t *testing.T, directorio string) {
	t.Helper()

	adaptador := &adaptadorDePrueba{
		directorio: directorio,
		reloj:      relojEn(instanteDeLaPrueba),
		fuente:     reproduccionDeLaGrabacion(t),
	}

	invocacion := invocarAlKernel(t, adaptador, "consultar", direccionDeLaNorma, "--json")
	require.Equal(t, 0, invocacion.codigo, "la siembra termina bien; salida de error: %s", invocacion.errores)
}

// invocacionDelKernel es en qué queda una invocación: lo único que un consumidor
// del binario puede mirar.
type invocacionDelKernel struct {
	codigo  int
	salida  string
	errores string
}

// invocarAlKernel ejecuta la raíz de composición entera sobre un registro que
// solo existe aquí dentro, con el applet de prueba delante de los argumentos. Es
// el único lugar del proyecto donde el adaptador de prueba se registra: ningún
// binario lo enlaza (FR-044).
func invocarAlKernel(t *testing.T, adaptador *adaptadorDePrueba, argv ...string) invocacionDelKernel {
	t.Helper()

	var registro app.Registro

	require.NoError(t, registro.Registrar(adaptador), "el adaptador de prueba declara un applet válido")

	var salida, errores bytes.Buffer

	codigo := app.Main(append([]string{nombreDelBinario, fuenteDePrueba}, argv...), &registro, &salida, &errores,
		versionDePrueba, commitDePrueba, fechaDePrueba)

	return invocacionDelKernel{codigo: codigo, salida: salida.String(), errores: errores.String()}
}

// sobreRecibido es el sobre tal como sale por la salida estándar con --json, con
// data sin interpretar: cada comprobación la lee con la forma que le toca, el
// cuerpo del verbo o los datos del fallo.
type sobreRecibido struct {
	Ok     bool            `json:"ok"`
	Fuente string          `json:"fuente"`
	URL    string          `json:"url"`
	Data   json.RawMessage `json:"data"`
}

// sobreDe analiza la salida estándar de una invocación con --json y comprueba de
// paso que lleva un único documento y nada más.
func sobreDe(t *testing.T, invocacion invocacionDelKernel) sobreRecibido {
	t.Helper()

	decodificador := json.NewDecoder(strings.NewReader(invocacion.salida))

	var sobre sobreRecibido

	require.NoError(t, decodificador.Decode(&sobre),
		"la salida estándar es el sobre: %q; salida de error: %s", invocacion.salida, invocacion.errores)
	require.ErrorIs(t, decodificador.Decode(new(json.RawMessage)), io.EOF, "y no hay nada más después del sobre")

	return sobre
}

// datosDe lee data con la forma que se espera, sin admitir ninguna clave de más.
func datosDe[T any](t *testing.T, data json.RawMessage) T {
	t.Helper()

	decodificador := json.NewDecoder(bytes.NewReader(data))
	decodificador.DisallowUnknownFields()

	var datos T

	require.NoError(t, decodificador.Decode(&datos), "data tiene la forma esperada: %s", data)

	return datos
}

// cuerpoServido comprueba que la invocación terminó bien citando la dirección
// consultada —venga el cuerpo de la fuente o de la caché— y devuelve el contenido
// de data.
func cuerpoServido(t *testing.T, invocacion invocacionDelKernel) cuerpoDeLaFuente {
	t.Helper()

	require.Equal(t, 0, invocacion.codigo, "la consulta termina bien; salida de error: %s", invocacion.errores)

	sobre := sobreDe(t, invocacion)
	assert.True(t, sobre.Ok)
	assert.Equal(t, fuenteDePrueba, sobre.Fuente)
	assert.Equal(t, direccionDeLaNorma, sobre.URL, "el sobre cita la dirección consultada")

	return datosDe[cuerpoDeLaFuente](t, sobre.Data)
}

// compruebaFallo comprueba el código de salida, que el sobre es de fallo, la
// clase que llega en data y que el mensaje nombra cada una de las cosas dadas.
func compruebaFallo(
	t *testing.T, invocacion invocacionDelKernel, codigo int, clase schema.Clase, nombra ...string,
) {
	t.Helper()

	require.Equal(t, codigo, invocacion.codigo, "salida de error: %s", invocacion.errores)

	sobre := sobreDe(t, invocacion)
	assert.False(t, sobre.Ok)

	fallo := datosDe[schema.DatosError](t, sobre.Data)
	assert.Equal(t, clase, fallo.Clase, "la clase que declara el error llega intacta al sobre")

	for _, nombrado := range nombra {
		assert.Contains(t, fallo.Mensaje, nombrado, "el mensaje nombra lo que explica el fallo")
	}
}

// compruebaSinPeticionesNiCambios es lo que toda invocación con --offline deja
// igual: el adaptador no pidió nada a la fuente y el directorio de la caché
// —cache.db byte a byte, o su ausencia— está como estaba (FR-015, FR-047, SC-003).
func compruebaSinPeticionesNiCambios(t *testing.T, adaptador *adaptadorDePrueba, directorio string, antes []string) {
	t.Helper()

	assert.Zero(t, adaptador.peticiones.Load(), "bajo --offline no se pide nada a la fuente (FR-047)")
	assert.Equal(t, antes, estadoDe(t, directorio),
		"y el directorio de la caché queda como estaba: ni fichero nuevo, ni cache.db distinto (SC-003)")
}

// estadoDe describe el directorio de la caché para compararlo antes y después:
// lo que contiene, con cache.db por su huella SHA-256 y sin los dos auxiliares
// del registro de escritura, o que no existe.
func estadoDe(t *testing.T, directorio string) []string {
	t.Helper()

	contenido, err := os.ReadDir(directorio)
	if errors.Is(err, fs.ErrNotExist) {
		return []string{directorio + " no existe"}
	}

	require.NoError(t, err)

	estado := []string{}

	for _, entrada := range contenido {
		switch nombre := entrada.Name(); nombre {
		case registroDeEscritura, memoriaCompartida:
			continue
		case ficheroDeLaBase:
			suma := sha256.Sum256(leeElFichero(t, filepath.Join(directorio, nombre)))
			estado = append(estado, nombre+" sha256:"+hex.EncodeToString(suma[:]))
		default:
			estado = append(estado, nombre)
		}
	}

	return estado
}

// leeElFichero devuelve el contenido entero de un fichero de la prueba.
func leeElFichero(t *testing.T, ruta string) []byte {
	t.Helper()

	contenido, err := os.ReadFile(filepath.Clean(ruta))
	require.NoError(t, err)

	return contenido
}

// leeLaNorma mira la caché por debajo del kernel, con un cliente propio cuyo
// «ahora» es el instante dado: devuelve lo que hay bajo la clave de la norma y si
// está vigente en ese instante.
func leeLaNorma(t *testing.T, directorio string, ahora time.Time) ([]byte, bool) {
	t.Helper()

	lectora, err := cache.New(t.Context(), cache.ConDirectorio(directorio), cache.ConReloj(relojEn(ahora)))
	require.NoError(t, err)

	contenido, presente, err := lectora.Get(t.Context(), claveDe(direccionDeLaNorma))
	require.NoError(t, errors.Join(err, lectora.Close()))

	return contenido, presente
}
