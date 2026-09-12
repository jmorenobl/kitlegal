package cache

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Las rutas y la clave de estas tablas son ficticias y fijas: el mensaje tiene
// que quedar igual en cualquier máquina, así que ninguna sale de un t.TempDir()
// ni del entorno. La clave imita la que un adaptador de fuente usará —prefijo
// del applet y dirección pedida—, contra el único host ficticio que el hito
// admite (plan.md, obligación 12).
const (
	directorioDePrueba = "/ruta"
	baseDePrueba       = "/ruta/cache.db"
	claveDePrueba      = "prueba:http://fuente.prueba/norma"
)

// Los dos orígenes que la tabla del contrato de errores §6 nombra en sus dos
// primeras formas. El vocabulario completo del campo Origen está en el
// comentario de Error; aquí se escriben tal como el mensaje los muestra, que es
// lo que este test fija.
const (
	origenDeLaVariable = "variable KITLEGAL_CACHE_DIR"
	origenDeLaOpcion   = "opción ConDirectorio"
)

// TestErrorMensajes fija la forma exacta de cada mensaje del contrato de
// errores §6 —una fila por forma— y, además, comprueba lo que la columna
// «Mensaje nombra» de la tabla cerrada del §3 exige de cada situación: la ruta
// y el origen cuando el fallo es de ruta, las dos versiones cuando el esquema
// es de otro binario, la clave y que la invocación es de solo lectura cuando no
// hay entrada, la operación tras cerrar, y la ruta junto con los dos ficheros
// auxiliares en el fallo de la fila 13 (FR-035, D9).
//
// La igualdad literal es lo que fija la forma; la lista de lo que el mensaje
// nombra es lo que fija el contrato, y se comprueba aparte para que reescribir
// una frase no pueda hacer desaparecer en silencio un dato que hace falta para
// actuar.
func TestErrorMensajes(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre   string
		fallo    *Error
		esperado string
		nombra   []string
	}{
		{
			nombre: "ruta inservible",
			fallo:  errorDeRutaInservible("construir", origenDeLaVariable, directorioDePrueba, nil),
			esperado: `caché: la variable KITLEGAL_CACHE_DIR apunta a "/ruta", ` +
				`que existe y no es un directorio`,
			nombra: []string{directorioDePrueba, origenDeLaVariable},
		},
		{
			nombre:   "directorio no escribible",
			fallo:    errorDeDirectorioNoEscribible("construir", origenDeLaOpcion, directorioDePrueba, nil),
			esperado: `caché: no se puede escribir en el directorio "/ruta" (opción ConDirectorio)`,
			nombra:   []string{directorioDePrueba, origenDeLaOpcion},
		},
		{
			nombre: "fichero no escribible por acceso denegado",
			fallo:  errorDeFicheroNoEscribible("construir", origenDeLaOpcion, baseDePrueba, fs.ErrPermission),
			esperado: `caché: no se puede abrir "/ruta/cache.db" para escribir (opción ConDirectorio): ` +
				`acceso denegado`,
			nombra: []string{baseDePrueba, origenDeLaOpcion, "acceso denegado"},
		},
		{
			nombre:   "fichero no escribible por otra causa",
			fallo:    errorDeFicheroNoEscribible("construir", origenDeLaVariable, baseDePrueba, errors.New("is a directory")),
			esperado: `caché: no se puede abrir "/ruta/cache.db" para escribir (variable KITLEGAL_CACHE_DIR)`,
			nombra:   []string{baseDePrueba, origenDeLaVariable},
		},
		{
			nombre:   "fichero ilegible en solo lectura",
			fallo:    errorDeFicheroIlegible(baseDePrueba, fs.ErrPermission),
			esperado: `caché: no se puede leer "/ruta/cache.db" en solo lectura: acceso denegado`,
			nombra:   []string{baseDePrueba, "solo lectura", "acceso denegado"},
		},
		{
			nombre:   "bloqueo agotado al construir",
			fallo:    errorDeBloqueo("construir", baseDePrueba, "", 5*time.Second, nil),
			esperado: `caché: "/ruta/cache.db" está bloqueada por otra invocación y la espera de 5s se agotó`,
			nombra:   []string{baseDePrueba, "5s"},
		},
		{
			nombre: "bloqueo agotado al escribir",
			fallo:  errorDeBloqueo("escribir", baseDePrueba, claveDePrueba, 5*time.Second, nil),
			esperado: `caché: no se pudo escribir "prueba:http://fuente.prueba/norma": "/ruta/cache.db" ` +
				`está bloqueada por otra invocación y la espera de 5s se agotó`,
			nombra: []string{"escribir", claveDePrueba, baseDePrueba, "5s"},
		},
		{
			nombre: "version ajena",
			fallo:  errorDeVersionAjena("migrar", baseDePrueba, 7, 1),
			esperado: `caché: "/ruta/cache.db" tiene el esquema en la versión 7 ` +
				`y este binario conoce la 1: no se modifica`,
			nombra: []string{baseDePrueba, "7", "1"},
		},
		{
			nombre:   "fichero inutilizable",
			fallo:    errorDeFicheroInutilizable("construir", baseDePrueba, nil),
			esperado: `caché: "/ruta/cache.db" no es una base de datos utilizable; no se borra`,
			nombra:   []string{baseDePrueba},
		},
		{
			nombre: "WAL sin memoria compartida",
			fallo:  errorDeWALSinMemoriaCompartida("leer", baseDePrueba, nil),
			esperado: `caché: "/ruta/cache.db" tiene un registro de escritura "/ruta/cache.db-wal" ` +
				`y el directorio no permite crear "/ruta/cache.db-shm": no se puede leer en solo lectura`,
			nombra: []string{baseDePrueba, "/ruta/cache.db-wal", "/ruta/cache.db-shm"},
		},
		{
			nombre: "ausencia en solo lectura",
			fallo:  errorDeAusenciaEnSoloLectura(baseDePrueba, claveDePrueba),
			esperado: `caché: no hay entrada vigente para "prueba:http://fuente.prueba/norma" ` +
				`y la invocación es de solo lectura (--offline)`,
			nombra: []string{claveDePrueba, "solo lectura"},
		},
		{
			nombre: "escritura en solo lectura",
			fallo:  errorDeEscrituraEnSoloLectura(baseDePrueba, claveDePrueba),
			esperado: `caché: no se puede escribir "prueba:http://fuente.prueba/norma" ` +
				`en una caché de solo lectura`,
			nombra: []string{claveDePrueba, "solo lectura"},
		},
		{
			nombre:   "tras cerrar",
			fallo:    errorTrasCierre("leer", baseDePrueba, claveDePrueba),
			esperado: `caché: cliente cerrado; no se puede leer "prueba:http://fuente.prueba/norma"`,
			nombra:   []string{"leer", claveDePrueba},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			mensaje := caso.fallo.Error()
			assert.Equal(t, caso.esperado, mensaje)

			for _, dato := range caso.nombra {
				assert.Contains(t, mensaje, dato,
					"el mensaje tiene que nombrar lo que la tabla del contrato de errores §3 exige")
			}
		})
	}
}

// TestErrorEnvueltoConservaLaClase comprueba el mecanismo de D9: el error
// declara su clase y el kernel la reconoce con errors.As sin que internal/cache
// importe internal/cli —es el test, que sí puede, quien llama a cli.Clasificar
// (contracts/errores-y-codigos.md §4)—, y envolverlo con fmt.Errorf("%w") no
// cambia ni la clase que el kernel devuelve ni el código de salida que produce,
// por muchas capas de contexto que se le añadan.
//
// Comprueba además las dos direcciones del envoltorio: hacia fuera, que
// errors.As recupera el *Error con sus campos intactos, que es como un
// adaptador distingue una ausencia en solo lectura de un fallo real sin
// analizar ninguna cadena de texto (contrato §2); y hacia dentro, que Unwrap
// expone la causa, de modo que errors.Is la alcance a través del error del
// paquete.
func TestErrorEnvueltoConservaLaClase(t *testing.T) {
	t.Parallel()

	causa := errors.New("el disco está lleno")

	casos := []struct {
		nombre string
		fallo  *Error
		clase  schema.Clase
		codigo int
	}{
		{
			nombre: "argumentos: la ruta existe y no es un directorio",
			fallo:  errorDeRutaInservible("construir", origenDeLaVariable, directorioDePrueba, causa),
			clase:  schema.ClaseArgumentos,
			codigo: 2,
		},
		{
			nombre: "argumentos: el directorio no se puede escribir",
			fallo:  errorDeDirectorioNoEscribible("construir", origenDeLaOpcion, directorioDePrueba, causa),
			clase:  schema.ClaseArgumentos,
			codigo: 2,
		},
		{
			nombre: "argumentos: el fichero no se puede abrir para escribir",
			fallo:  errorDeFicheroNoEscribible("construir", origenDeLaOpcion, baseDePrueba, causa),
			clase:  schema.ClaseArgumentos,
			codigo: 2,
		},
		{
			nombre: "argumentos: la vigencia recibida no es válida",
			fallo:  errorDeArgumentos("escribir", "la vigencia tiene que ser mayor que cero", nil),
			clase:  schema.ClaseArgumentos,
			codigo: 2,
		},
		{
			nombre: "fuente no disponible: no hay entrada y solo se lee",
			fallo:  errorDeAusenciaEnSoloLectura(baseDePrueba, claveDePrueba),
			clase:  schema.ClaseFuenteNoDisponible,
			codigo: 4,
		},
		{
			nombre: "fuente no disponible: el contexto se canceló",
			fallo:  errorDeFuenteNoDisponible("leer", "el contexto terminó antes que la operación", causa),
			clase:  schema.ClaseFuenteNoDisponible,
			codigo: 4,
		},
		{
			nombre: "inesperado: el esquema es de otra versión",
			fallo:  errorDeVersionAjena("migrar", baseDePrueba, 7, 1),
			clase:  schema.ClaseInesperado,
			codigo: 1,
		},
		{
			nombre: "inesperado: el fichero no es una base utilizable",
			fallo:  errorDeFicheroInutilizable("construir", baseDePrueba, causa),
			clase:  schema.ClaseInesperado,
			codigo: 1,
		},
		{
			nombre: "inesperado: hay registro de escritura y no se puede crear la memoria compartida",
			fallo:  errorDeWALSinMemoriaCompartida("leer", baseDePrueba, causa),
			clase:  schema.ClaseInesperado,
			codigo: 1,
		},
		{
			nombre: "inesperado: el fichero no se deja leer en solo lectura",
			fallo:  errorDeFicheroIlegible(baseDePrueba, causa),
			clase:  schema.ClaseInesperado,
			codigo: 1,
		},
		{
			nombre: "inesperado: la base sigue bloqueada tras la espera",
			fallo:  errorDeBloqueo("escribir", baseDePrueba, claveDePrueba, 5*time.Second, causa),
			clase:  schema.ClaseInesperado,
			codigo: 1,
		},
		{
			nombre: "inesperado: se intenta escribir en una caché de solo lectura",
			fallo:  errorDeEscrituraEnSoloLectura(baseDePrueba, claveDePrueba),
			clase:  schema.ClaseInesperado,
			codigo: 1,
		},
		{
			nombre: "inesperado: la operación llega después de cerrar",
			fallo:  errorTrasCierre("escribir", baseDePrueba, claveDePrueba),
			clase:  schema.ClaseInesperado,
			codigo: 1,
		},
		{
			nombre: "inesperado: falla el cierre de la base",
			fallo:  errorInesperado("cerrar", "no se pudo cerrar la base de datos", causa),
			clase:  schema.ClaseInesperado,
			codigo: 1,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.clase, caso.fallo.Clase())
			assert.Equal(t, caso.clase, cli.Clasificar(caso.fallo))
			assert.Equal(t, caso.codigo, cli.CodigoSalida(caso.fallo))

			envuelto := fmt.Errorf("al consultar la caché: %w",
				fmt.Errorf("al abrir la base de datos: %w", caso.fallo))

			assert.Equal(t, caso.clase, cli.Clasificar(envuelto),
				"envolver con %%w no puede cambiar la clase que el kernel reconoce")
			assert.Equal(t, caso.codigo, cli.CodigoSalida(envuelto))

			var recuperado *Error
			require.ErrorAs(t, envuelto, &recuperado)
			assert.Equal(t, caso.fallo, recuperado)

			if caso.fallo.Causa != nil {
				assert.ErrorIs(t, envuelto, caso.fallo.Causa,
					"Unwrap tiene que exponer la causa a través de todas las capas")
			}
		})
	}
}

// TestErrorNuloOCero comprueba que leer un *Error que no construyó ningún
// constructor del paquete —uno nulo, o uno a cero desde fuera— ni entra en
// panic ni devuelve un mensaje vacío, y que declara la clase de lo que nadie
// previó: un sobre de fallo sin mensaje no le diría nada a nadie, y ninguna
// ruta puede terminar en una clase que el sobre no pueda llevar (FR-033,
// FR-034, contracts/errores-y-codigos.md §2).
func TestErrorNuloOCero(t *testing.T) {
	t.Parallel()

	// Error es un tipo exportado: nada impide que alguien de fuera lo construya
	// a cero, ni que una función devuelva un *Error nulo como error. Ninguno de
	// los dos casos lo produce este paquete, y ninguno puede acabar en panic ni
	// en un sobre con el mensaje vacío.
	aCero := &Error{}

	var nulo *Error

	require.NotPanics(t, func() {
		assert.NotEmpty(t, aCero.Error())
		assert.Equal(t, schema.ClaseInesperado, aCero.Clase())
		assert.NoError(t, aCero.Unwrap())

		assert.NotEmpty(t, nulo.Error())
		assert.Equal(t, schema.ClaseInesperado, nulo.Clase())
		assert.NoError(t, nulo.Unwrap())
	})

	// Un error así, si llegara al kernel, sale con el código de lo no previsto
	// y nunca con el del éxito.
	assert.Equal(t, schema.ClaseInesperado, cli.Clasificar(aCero))
	assert.Equal(t, 1, cli.CodigoSalida(aCero))
}

// filasProvocables son las filas de la tabla cerrada del contrato de errores §3
// que se pueden provocar sin que el sistema de ficheros haga valer ningún
// permiso: todas menos la 12 y la 13, que miden con las mismas dos funciones
// del kernel TestIntegracionDirectorioDenegado y
// TestIntegracionWALSinMemoriaCompartida.
var filasProvocables = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 14, 15}

// clasesQueLaCacheNoProduce y codigosQueLaCacheNoProduce son las tres clases del
// vocabulario que ninguna ruta de fallo de la caché puede declarar, con sus
// códigos de salida: la caché no sabe si algo existe en la fuente, no habla con
// nadie que limite peticiones y no actúa con ninguna identidad (FR-033).
var (
	clasesQueLaCacheNoProduce = []schema.Clase{
		schema.ClaseNoEncontrado,
		schema.ClaseLimiteOTos,
		schema.ClaseIdentidadHumana,
	}
	codigosQueLaCacheNoProduce = []int{3, 5, 6}
)

// provocado es un error tal como lo devolvió el paquete y cómo se provocó. Una
// misma fila de la tabla se da de varias formas —en Get y en Put, en modo normal
// y en solo lectura, con el contexto cancelado y vencido— y cada una se
// comprueba por separado, con su nombre en el mensaje, sin abrir por eso una
// subprueba anidada.
type provocado struct {
	como string
	err  error
}

// TestClasesDeError fija FR-033 y FR-034 sobre el error real y no sobre uno
// construido a mano: cada fila de la tabla cerrada del contrato de errores §3
// que se puede provocar sin permisos —de la 1 a la 11, la 14 y la 15— se provoca
// de verdad, con toda base en un t.TempDir(), y lo que el paquete devuelve se
// entrega al kernel. La clase que reconoce cli.Clasificar y el código que produce
// cli.CodigoSalida tienen que ser los de la fila, tanto con el error tal como sale
// del paquete como envuelto con %w; ninguna forma puede terminar en las tres
// clases que la caché no produce —«no encontrado» (3), «límite o TOS» (5) e
// «identidad humana» (6)— y ninguna termina en panic (SC-011, D9).
//
// Es una subprueba por fila y ninguna anidada: el escenario 8 del quickstart
// cuenta las subpruebas de primer nivel y espera trece. Las formas de una misma
// fila se comprueban dentro de su subprueba. Qué filas hay no se deja a esa
// cuenta: la cabecera exige exactamente las del contrato, y en su orden, antes de
// provocar ninguna.
//
// No declara t.Parallel(): las filas 4 y 7 declaran el entorno con t.Setenv, que
// no se puede usar en pruebas paralelas. Por eso cada subprueba fija antes de
// nada las tres fuentes de la ruta —la cuenta en un directorio temporal y la
// variable sin declarar—, de modo que ni el entorno de quien ejecuta la prueba
// ni lo que declare otra fila cambien un resultado ni saquen ninguna base de un
// t.TempDir().
func TestClasesDeError(t *testing.T) {
	casos := []struct {
		fila      int
		situacion string
		clase     schema.Clase
		codigo    int
		provoca   func(t *testing.T) []provocado
	}{
		{
			fila: 1, situacion: "clave-vacia",
			clase: schema.ClaseArgumentos, codigo: 2, provoca: provocaClaveVacia,
		},
		{
			fila: 2, situacion: "vigencia-no-positiva",
			clase: schema.ClaseArgumentos, codigo: 2, provoca: provocaVigenciaNoPositiva,
		},
		{
			fila: 3, situacion: "opcion-invalida",
			clase: schema.ClaseArgumentos, codigo: 2, provoca: provocaOpcionInvalida,
		},
		{
			fila: 4, situacion: "variable-presente-y-vacia",
			clase: schema.ClaseArgumentos, codigo: 2, provoca: provocaVariableVacia,
		},
		{
			fila: 5, situacion: "ruta-que-no-es-un-directorio",
			clase: schema.ClaseArgumentos, codigo: 2, provoca: provocaRutaQueNoEsUnDirectorio,
		},
		{
			fila: 6, situacion: "directorio-no-creable-en-modo-normal",
			clase: schema.ClaseArgumentos, codigo: 2, provoca: provocaDirectorioNoCreable,
		},
		{
			fila: 7, situacion: "sin-directorio-de-la-cuenta",
			clase: schema.ClaseArgumentos, codigo: 2, provoca: provocaSinDirectorioDeLaCuenta,
		},
		{
			fila: 8, situacion: "ausencia-en-solo-lectura",
			clase: schema.ClaseFuenteNoDisponible, codigo: 4, provoca: provocaAusenciaEnSoloLectura,
		},
		{
			fila: 9, situacion: "version-de-esquema-desconocida",
			clase: schema.ClaseInesperado, codigo: 1, provoca: provocaVersionDesconocida,
		},
		{
			fila: 10, situacion: "fichero-inutilizable",
			clase: schema.ClaseInesperado, codigo: 1, provoca: provocaFicheroInutilizable,
		},
		{
			fila: 11, situacion: "escritura-en-solo-lectura",
			clase: schema.ClaseInesperado, codigo: 1, provoca: provocaEscrituraEnSoloLectura,
		},
		{
			fila: 14, situacion: "operacion-tras-cerrar",
			clase: schema.ClaseInesperado, codigo: 1, provoca: provocaOperacionTrasCerrar,
		},
		{
			fila: 15, situacion: "contexto-terminado",
			clase: schema.ClaseFuenteNoDisponible, codigo: 4, provoca: provocaContextoTerminado,
		},
	}

	filas := make([]int, 0, len(casos))
	for _, caso := range casos {
		filas = append(filas, caso.fila)
	}

	require.Equal(t, filasProvocables, filas,
		"una subprueba por cada fila del contrato de errores §3 que se provoca sin permisos, y ninguna más")

	for _, caso := range casos {
		t.Run(fmt.Sprintf("fila-%02d-%s", caso.fila, caso.situacion), func(t *testing.T) {
			cuenta := t.TempDir()
			t.Setenv("HOME", cuenta)
			t.Setenv("USERPROFILE", cuenta)
			sinVariable(t)

			var provocados []provocado

			require.NotPanics(t, func() { provocados = caso.provoca(t) },
				"ninguna ruta de fallo termina en panic (FR-034)")
			require.NotEmpty(t, provocados, "la fila se provoca al menos de una forma")

			for _, hecho := range provocados {
				compruebaLaClase(t, hecho, caso.clase, caso.codigo)
			}
		})
	}
}

// compruebaLaClase entrega un error provocado al kernel en las dos formas en que
// le puede llegar —tal como sale del paquete y envuelto con %w por quien lo
// propaga— y exige en las dos la clase y el código de la fila, que no sean los de
// ninguna de las tres clases que la caché no produce, un mensaje que no esté
// vacío y que el error sea de verdad un *Error que declara esa clase él mismo: sin
// eso último la fila pasaría igual con un error cualquiera que el kernel
// clasificara en la misma clase por otra vía (contrato de errores §1 y §4).
func compruebaLaClase(t *testing.T, hecho provocado, clase schema.Clase, codigo int) {
	t.Helper()

	require.Error(t, hecho.err, "%s: la situación tiene que fallar", hecho.como)

	formas := []struct {
		nombre string
		err    error
	}{
		{nombre: "tal como sale del paquete", err: hecho.err},
		{nombre: "envuelto con %w", err: fmt.Errorf("al consultar la caché: %w", hecho.err)},
	}

	for _, forma := range formas {
		donde := hecho.como + ", " + forma.nombre

		var (
			declarada schema.Clase
			salida    int
			mensaje   string
		)

		require.NotPanics(t, func() {
			declarada = cli.Clasificar(forma.err)
			salida = cli.CodigoSalida(forma.err)
			mensaje = forma.err.Error()
		}, "%s: clasificar el error y leer su mensaje no entra en panic (FR-034)", donde)

		assert.Equal(t, clase, declarada, "%s: la clase que reconoce el kernel", donde)
		assert.Equal(t, codigo, salida, "%s: el código de salida", donde)
		assert.NotContains(t, clasesQueLaCacheNoProduce, declarada,
			"%s: la caché nunca declara esta clase (FR-033)", donde)
		assert.NotContains(t, codigosQueLaCacheNoProduce, salida,
			"%s: la caché nunca termina con este código (FR-033)", donde)
		assert.NotEmpty(t, mensaje, "%s: un sobre de fallo sin mensaje no dice nada", donde)

		var fallo *Error
		require.ErrorAs(t, forma.err, &fallo, "%s: el error es el del paquete", donde)
		assert.Equal(t, clase, fallo.Clase(), "%s: y es él quien declara la clase (schema.ConClase)", donde)
	}
}

// construye intenta construir un cliente con las opciones y devuelve el error de
// New. Si New no fallara donde la fila espera que falle, el cliente se cierra al
// terminar la prueba: la comprobación de la clase ya dirá que faltaba el error, y
// no queda ninguna base abierta.
func construye(t *testing.T, opciones ...Opcion) error {
	t.Helper()

	cliente, err := New(t.Context(), opciones...)
	cierraAlTerminar(t, cliente)

	return err
}

// cierraAlTerminar registra el cierre del cliente que devolvió New, si devolvió
// alguno.
func cierraAlTerminar(t *testing.T, cliente *Cliente) {
	t.Helper()

	if cliente != nil {
		t.Cleanup(func() { require.NoError(t, cliente.Close()) })
	}
}

// construyeEnLosDosModos es construye en modo normal y en solo lectura, para las
// filas que valen igual se vaya a escribir o solo a leer (FR-015, FR-022).
func construyeEnLosDosModos(t *testing.T, como string, opciones ...Opcion) []provocado {
	t.Helper()

	return []provocado{
		{como: como + ", en modo normal", err: construye(t, opciones...)},
		{como: como + ", en solo lectura", err: construye(t, append(slices.Clone(opciones), SoloLectura())...)},
	}
}

// leeEnSoloLectura abre el directorio en solo lectura y devuelve el error de
// leer la clave.
func leeEnSoloLectura(t *testing.T, directorio, clave string, opciones ...Opcion) error {
	t.Helper()

	lector := clienteAbierto(t, directorio, append([]Opcion{SoloLectura()}, opciones...)...)
	_, _, err := lector.Get(t.Context(), clave)

	return err
}

// provocaClaveVacia es la fila 1: la clave vacía al leer y al escribir, sobre una
// base que existe y funciona.
func provocaClaveVacia(t *testing.T) []provocado {
	t.Helper()

	cliente := clienteAbierto(t, t.TempDir())

	_, _, alLeer := cliente.Get(t.Context(), "")
	alEscribir := cliente.Put(t.Context(), "", []byte("<norma>contenido</norma>"), vigenciaDePrueba)

	return []provocado{
		{como: "Get con la clave vacía", err: alLeer},
		{como: "Put con la clave vacía", err: alEscribir},
	}
}

// provocaVigenciaNoPositiva es la fila 2: una vigencia de cero y una negativa.
func provocaVigenciaNoPositiva(t *testing.T) []provocado {
	t.Helper()

	cliente := clienteAbierto(t, t.TempDir())
	contenido := []byte("<norma>contenido</norma>")

	return []provocado{
		{como: "Put con la vigencia cero", err: cliente.Put(t.Context(), claveDePrueba, contenido, 0)},
		{
			como: "Put con una vigencia negativa",
			err:  cliente.Put(t.Context(), claveDePrueba, contenido, -vigenciaDePrueba),
		},
	}
}

// provocaOpcionInvalida es la fila 3: cada una de las tres opciones que pueden
// llevar un valor que no sirve. Las dos que no declaran el directorio van detrás
// de uno temporal, para que un New que no las rechazara tampoco saliera de él.
func provocaOpcionInvalida(t *testing.T) []provocado {
	t.Helper()

	return []provocado{
		{como: "ConDirectorio sin ruta", err: construye(t, ConDirectorio(""))},
		{como: "ConReloj sin reloj", err: construye(t, ConDirectorio(t.TempDir()), ConReloj(nil))},
		{
			como: "ConRegistrador sin registrador",
			err:  construye(t, ConDirectorio(t.TempDir()), ConRegistrador(nil)),
		},
	}
}

// provocaVariableVacia es la fila 4: KITLEGAL_CACHE_DIR declarada y vacía, que no
// puede caer en silencio a la ruta por omisión, en los dos modos.
func provocaVariableVacia(t *testing.T) []provocado {
	t.Helper()

	t.Setenv(VariableDirectorio, "")

	return construyeEnLosDosModos(t, "la variable declarada y vacía")
}

// provocaRutaQueNoEsUnDirectorio es la fila 5: la ruta existe y es un fichero,
// llegue por la opción o por la variable, y en los dos modos.
func provocaRutaQueNoEsUnDirectorio(t *testing.T) []provocado {
	t.Helper()

	fichero := filepath.Join(t.TempDir(), "fichero")
	require.NoError(t, os.WriteFile(fichero, []byte("no soy un directorio"), 0o600))

	provocados := construyeEnLosDosModos(t, "un fichero por la opción", ConDirectorio(fichero))

	t.Setenv(VariableDirectorio, fichero)

	return append(provocados, construyeEnLosDosModos(t, "un fichero por la variable")...)
}

// provocaDirectorioNoCreable es la fila 6 sin permisos: en modo normal el
// directorio cuelga de un fichero y no se puede crear, llegue la ruta por la
// opción o por la variable. En solo lectura ese mismo directorio no es un fallo
// de ruta sino uno inexistente, y es una de las formas de la fila 8.
func provocaDirectorioNoCreable(t *testing.T) []provocado {
	t.Helper()

	_, porLaOpcion := directorioBajoUnFichero(t)
	_, porLaVariable := directorioBajoUnFichero(t)

	alDeclararlo := construye(t, ConDirectorio(porLaOpcion))

	t.Setenv(VariableDirectorio, porLaVariable)

	return []provocado{
		{como: "bajo un fichero, por la opción", err: alDeclararlo},
		{como: "bajo un fichero, por la variable", err: construye(t)},
	}
}

// provocaSinDirectorioDeLaCuenta es la fila 7: no hay opción ni variable y
// tampoco se sabe cuál es el directorio de la cuenta, en los dos modos.
func provocaSinDirectorioDeLaCuenta(t *testing.T) []provocado {
	t.Helper()

	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")

	return construyeEnLosDosModos(t, "sin HOME ni "+VariableDirectorio)
}

// provocaAusenciaEnSoloLectura es la fila 8 con todas sus formas: en solo
// lectura no hay entrada vigente porque la clave no está, porque la entrada
// expiró, porque la base no tiene esquema o porque no existe cache.db, ni el
// directorio, ni como directorio —su padre es un fichero— (FR-015, FR-016,
// FR-018).
func provocaAusenciaEnSoloLectura(t *testing.T) []provocado {
	t.Helper()

	conEntrada := t.TempDir()
	siembra(t, conEntrada, claveDePrueba, []byte("<norma>contenido</norma>"),
		ConReloj(relojEn(instanteDePrueba).Ahora))

	sinEsquema := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(sinEsquema, ficheroDeLaBase), nil, 0o600))

	_, bajoUnFichero := directorioBajoUnFichero(t)
	enLaExpiracion := ConReloj(relojEn(instanteDePrueba.Add(vigenciaDePrueba)).Ahora)

	return []provocado{
		{como: "una clave que no está", err: leeEnSoloLectura(t, conEntrada, claveDePrueba+"/ausente")},
		{
			como: "una entrada en su instante de expiración",
			err:  leeEnSoloLectura(t, conEntrada, claveDePrueba, enLaExpiracion),
		},
		{como: "una base sin esquema", err: leeEnSoloLectura(t, sinEsquema, claveDePrueba)},
		{como: "sin cache.db", err: leeEnSoloLectura(t, t.TempDir(), claveDePrueba)},
		{
			como: "sin el directorio",
			err:  leeEnSoloLectura(t, filepath.Join(t.TempDir(), "no-esta"), claveDePrueba),
		},
		{como: "con el directorio bajo un fichero", err: leeEnSoloLectura(t, bajoUnFichero, claveDePrueba)},
	}
}

// provocaVersionDesconocida es la fila 9: la base registra la versión 99, que este
// binario no conoce, y ni el modo normal la migra ni el de solo lectura la lee.
func provocaVersionDesconocida(t *testing.T) []provocado {
	t.Helper()

	directorio := t.TempDir()
	siembra(t, directorio, claveDePrueba, []byte("<norma>contenido</norma>"))
	ejecutaEnLaBase(t, filepath.Join(directorio, ficheroDeLaBase),
		`INSERT INTO schema_version(version, aplicada_en) VALUES (99, '2026-01-01T00:00:00Z')`)

	return construyeEnLosDosModos(t, "el esquema en la versión 99", ConDirectorio(directorio))
}

// provocaFicheroInutilizable es la fila 10: en la ruta de la base hay un fichero
// que no es una base de datos.
func provocaFicheroInutilizable(t *testing.T) []provocado {
	t.Helper()

	directorio := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(directorio, ficheroDeLaBase),
		[]byte("esto no es una base de datos de SQLite"), 0o600))

	return construyeEnLosDosModos(t, "un cache.db que no es una base", ConDirectorio(directorio))
}

// provocaEscrituraEnSoloLectura es la fila 11: escribir en una caché de solo
// lectura, sobre una base que existe y sobre un cliente sin base.
func provocaEscrituraEnSoloLectura(t *testing.T) []provocado {
	t.Helper()

	contenido := []byte("<norma>contenido</norma>")

	conBase := t.TempDir()
	siembra(t, conBase, claveDePrueba, contenido)

	sobreLaBase := clienteAbierto(t, conBase, SoloLectura())
	sinBase := clienteAbierto(t, t.TempDir(), SoloLectura())

	return []provocado{
		{
			como: "Put sobre una base que existe",
			err:  sobreLaBase.Put(t.Context(), claveDePrueba, contenido, vigenciaDePrueba),
		},
		{como: "Put sin base", err: sinBase.Put(t.Context(), claveDePrueba, contenido, vigenciaDePrueba)},
	}
}

// provocaOperacionTrasCerrar es la fila 14 en las dos formas que se pueden
// provocar de forma determinista: una operación después de Close —sobre un
// cliente con base y sobre uno sin ella— y la base bloqueada por otra invocación
// más tiempo que la espera, con la espera acortada a milisegundos, al construir
// sobre una base sin migrar y al escribir. Un fallo de entrada y salida que
// sobreviene no se puede provocar sin estropear el disco, y su clase la fija el
// mismo constructor, errorInesperado.
func provocaOperacionTrasCerrar(t *testing.T) []provocado {
	t.Helper()

	conBase := clienteAbierto(t, t.TempDir())
	require.NoError(t, conBase.Close())

	sinBase := clienteAbierto(t, t.TempDir(), SoloLectura())
	require.NoError(t, sinBase.Close())

	_, _, alLeer := conBase.Get(t.Context(), claveDePrueba)
	alEscribir := conBase.Put(t.Context(), claveDePrueba, []byte("<norma>contenido</norma>"), vigenciaDePrueba)
	_, _, alLeerSinBase := sinBase.Get(t.Context(), claveDePrueba)

	sinMigrar := t.TempDir()
	retieneElBloqueo(t, filepath.Join(sinMigrar, ficheroDeLaBase))

	bloqueada := t.TempDir()
	escritor := clienteAbierto(t, bloqueada, conEsperaAnteBloqueo(esperaCorta))
	retieneElBloqueo(t, filepath.Join(bloqueada, ficheroDeLaBase))

	return []provocado{
		{como: "Get tras cerrar", err: alLeer},
		{como: "Put tras cerrar", err: alEscribir},
		{como: "Get tras cerrar un cliente sin base", err: alLeerSinBase},
		{
			como: "New con la base bloqueada más tiempo que la espera",
			err:  construye(t, ConDirectorio(sinMigrar), conEsperaAnteBloqueo(esperaCorta)),
		},
		{
			como: "Put con la base bloqueada más tiempo que la espera",
			err:  escritor.Put(t.Context(), claveDePrueba, []byte("<norma>contenido</norma>"), vigenciaDePrueba),
		},
	}
}

// provocaContextoTerminado es la fila 15: New, Get y Put con el contexto
// cancelado y con el contexto vencido, y además New y Put con un contexto que
// vence **durante** la espera ante el bloqueo que otra invocación retiene, que
// es donde la espera del motor no mira el contexto y el cliente tiene que
// mirarlo por él. Get y Put operan sobre una base con la entrada guardada y
// vigente, para que el 4 no se pueda confundir con ninguna ausencia. Todo se
// prepara con el contexto de la prueba y solo la llamada al paquete recibe el
// terminado.
func provocaContextoTerminado(t *testing.T) []provocado {
	t.Helper()

	cancelado, cancela := context.WithCancel(t.Context())
	cancela()

	vencido, suelta := context.WithDeadline(t.Context(), time.Unix(0, 0))
	t.Cleanup(suelta)

	contenido := []byte("<norma>contenido</norma>")
	directorio := t.TempDir()
	siembra(t, directorio, claveDePrueba, contenido)
	cliente := clienteAbierto(t, directorio)

	terminados := []struct {
		nombre string
		ctx    context.Context
	}{
		{nombre: "cancelado", ctx: cancelado},
		{nombre: "vencido", ctx: vencido},
	}

	provocados := make([]provocado, 0, 3*len(terminados)+2)

	for _, terminado := range terminados {
		nuevo, alConstruir := New(terminado.ctx, ConDirectorio(t.TempDir()))
		cierraAlTerminar(t, nuevo)

		_, _, alLeer := cliente.Get(terminado.ctx, claveDePrueba)
		alEscribir := cliente.Put(terminado.ctx, claveDePrueba, contenido, vigenciaDePrueba)

		provocados = append(provocados,
			provocado{como: "New con el contexto " + terminado.nombre, err: alConstruir},
			provocado{como: "Get con el contexto " + terminado.nombre, err: alLeer},
			provocado{como: "Put con el contexto " + terminado.nombre, err: alEscribir},
		)
	}

	return append(provocados, provocaVencimientoDuranteLaEspera(t)...)
}

// provocaVencimientoDuranteLaEspera es la forma de la fila 15 que vive en la
// espera ante bloqueo: otra invocación retiene el bloqueo de escritura y el
// plazo de quien llama vence mientras New —sobre una base sin migrar— y Put
// esperan.
func provocaVencimientoDuranteLaEspera(t *testing.T) []provocado {
	t.Helper()

	sinMigrar := t.TempDir()
	retieneElBloqueo(t, filepath.Join(sinMigrar, ficheroDeLaBase))

	bloqueada := t.TempDir()
	escritor := clienteAbierto(t, bloqueada)
	retieneElBloqueo(t, filepath.Join(bloqueada, ficheroDeLaBase))

	ctx, cancela := context.WithTimeout(t.Context(), plazoCorto)
	t.Cleanup(cancela)

	nuevo, alConstruir := New(ctx, ConDirectorio(sinMigrar))
	cierraAlTerminar(t, nuevo)

	return []provocado{
		{como: "New con el contexto vencido durante la espera ante bloqueo", err: alConstruir},
		{
			como: "Put con el contexto vencido durante la espera ante bloqueo",
			err:  escritor.Put(ctx, claveDePrueba, []byte("<norma>contenido</norma>"), vigenciaDePrueba),
		},
	}
}
