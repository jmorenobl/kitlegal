package cache

import (
	"errors"
	"fmt"
	"testing"

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
