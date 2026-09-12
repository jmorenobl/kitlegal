package cache

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Las dos tablas que resuelven o validan la ruta no declaran t.Parallel(), y no
// es un descuido: usan t.Setenv, que «cannot be used in parallel tests», y el
// entorno es justamente lo que miden. Declaran además siempre el estado de las
// tres fuentes de la ruta —la opción, la variable y el directorio de la
// cuenta—, de modo que lo que haya en la máquina de quien ejecuta el test no
// pueda cambiar ningún resultado.

// TestEsInexistente fija la regla «inexistente» del contrato del puerto §3: de
// todo lo que os.Stat puede devolver, solo dos cosas significan que el
// directorio no está —no hay nada en la ruta, o el padre es un fichero y por
// tanto no existe *como directorio*— y todo lo demás es un fallo que no se sabe
// interpretar como ausencia. Es la única regla que decide eso, y de ella
// dependen tanto el trato del directorio en solo lectura como el de cache.db
// (FR-015, D3).
//
// La fila del padre que es un fichero comprueba además que ese error **no**
// equivale a fs.ErrNotExist: es la razón exacta por la que la regla necesita
// las dos condiciones y no una, y sin fijarla nada impediría simplificarla a
// una sola y devolver un fallo (1) donde el contrato exige una ausencia (4).
func TestEsInexistente(t *testing.T) {
	t.Parallel()

	base := t.TempDir()

	fichero := filepath.Join(base, "fichero")
	require.NoError(t, os.WriteFile(fichero, []byte("no soy un directorio"), 0o600))

	_, sinNada := os.Stat(filepath.Join(base, "no-esta"))
	_, padreFichero := os.Stat(filepath.Join(fichero, "sub"))

	// El acceso denegado no se provoca aquí: haría falta que el sistema de
	// ficheros hiciera valer unos permisos, que es una precondición del entorno
	// y la comprueban los tests de integración del hito, los únicos que pueden
	// dejar de medir cuando no se cumple. Lo que esta fila fija es el
	// predicado, así que el error se construye con la forma exacta que os.Stat
	// devuelve en ese caso —un *fs.PathError con syscall.EACCES, que errors.Is
	// reconoce como fs.ErrPermission— y la fila comprueba esa equivalencia
	// antes de mirar lo que el predicado responde.
	denegado := &fs.PathError{Op: "stat", Path: base, Err: syscall.EACCES}

	casos := []struct {
		nombre      string
		err         error
		equivaleA   error
		noEquivaleA error
		esperado    bool
	}{
		{
			nombre:    "no hay nada en la ruta",
			err:       sinNada,
			equivaleA: fs.ErrNotExist,
			esperado:  true,
		},
		{
			nombre:      "el padre de la ruta es un fichero",
			err:         padreFichero,
			equivaleA:   syscall.ENOTDIR,
			noEquivaleA: fs.ErrNotExist,
			esperado:    true,
		},
		{
			nombre:    "el acceso está denegado",
			err:       denegado,
			equivaleA: fs.ErrPermission,
			esperado:  false,
		},
		{
			nombre:   "no hay ningún error",
			err:      nil,
			esperado: false,
		},
		{
			nombre:   "un error ajeno al sistema de ficheros",
			err:      errors.New("la base de datos está ocupada"),
			esperado: false,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			if caso.equivaleA != nil {
				require.ErrorIs(t, caso.err, caso.equivaleA,
					"la fila no mide lo que dice si el error no es el que anuncia")
			}

			if caso.noEquivaleA != nil {
				require.NotErrorIs(t, caso.err, caso.noEquivaleA)
			}

			assert.Equal(t, caso.esperado, esInexistente(caso.err))
		})
	}
}

// TestRutaEfectivaPrecedencia fija el orden en que sale la ruta —la opción, la
// variable de entorno y, si no hay ninguna, el directorio de la cuenta—, el
// origen con el que cada una viaja a los mensajes y que la variable se lee
// **una sola vez**, al construir (FR-019, FR-023, D3).
//
// Las lecturas se cuentan envolviendo la consulta al entorno: es lo que
// permite comprobar que la opción, cuando está, no llega siquiera a mirar la
// variable, y que cuando no está la variable se consulta una vez y no una por
// cada uso posterior de la ruta.
//
// La fila de la variable declarada y vacía es FR-022 al pie de la letra: el
// valor vacío viaja con su origen y no se convierte nunca en la ruta por
// omisión, que es la caída silenciosa que el requisito prohíbe. Quien lo
// rechaza es compruebaRuta, y eso lo mide TestRutaInservible.
func TestRutaEfectivaPrecedencia(t *testing.T) {
	cuenta := t.TempDir()
	t.Setenv("HOME", cuenta)
	t.Setenv("USERPROFILE", cuenta)

	declarado := filepath.Join(t.TempDir(), "por-la-opcion")
	delEntorno := filepath.Join(t.TempDir(), "por-la-variable")

	casos := []struct {
		nombre     string
		directorio string
		variable   string
		declarada  bool
		sinCuenta  bool
		esperada   string
		de         origenDeLaRuta
		lecturas   int
		nombra     []string
	}{
		{
			nombre:     "la opción gana a la variable y al directorio de la cuenta",
			directorio: declarado,
			variable:   delEntorno,
			declarada:  true,
			esperada:   declarado,
			de:         origenOpcion,
			lecturas:   0,
		},
		{
			nombre:    "la variable gana al directorio de la cuenta",
			variable:  delEntorno,
			declarada: true,
			esperada:  delEntorno,
			de:        origenVariable,
			lecturas:  1,
		},
		{
			nombre:    "la variable declarada y vacía no cae al directorio de la cuenta",
			variable:  "",
			declarada: true,
			esperada:  "",
			de:        origenVariable,
			lecturas:  1,
		},
		{
			nombre:   "sin opción ni variable, el directorio de la cuenta",
			esperada: filepath.Join(cuenta, ".cache", "kitlegal"),
			de:       origenOmision,
			lecturas: 1,
		},
		{
			nombre:    "sin cuenta, la ruta por omisión no se puede determinar",
			sinCuenta: true,
			de:        origenOmision,
			lecturas:  1,
			nombra:    []string{"HOME", VariableDirectorio},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			if caso.declarada {
				t.Setenv(VariableDirectorio, caso.variable)
			} else {
				sinVariable(t)
			}

			if caso.sinCuenta {
				t.Setenv("HOME", "")
				t.Setenv("USERPROFILE", "")
			}

			lecturas := 0
			busca := func(nombre string) (string, bool) {
				lecturas++
				assert.Equal(t, VariableDirectorio, nombre,
					"la única variable de entorno de la caché es la del contrato")

				return os.LookupEnv(nombre)
			}

			ruta, de, err := rutaEfectiva(caso.directorio, busca)

			if len(caso.nombra) > 0 {
				require.Error(t, err)
				assert.Equal(t, schema.ClaseArgumentos, cli.Clasificar(err))
				assert.Equal(t, 2, cli.CodigoSalida(err))

				for _, texto := range caso.nombra {
					assert.Contains(t, err.Error(), texto,
						"el mensaje tiene que decir qué hay que declarar para arreglarlo")
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, caso.esperada, ruta)
			}

			assert.Equal(t, caso.de, de, "el origen viaja con la ruta hasta los mensajes")
			assert.Equal(t, caso.lecturas, lecturas,
				"la variable se lee una sola vez, al construir, y solo si no hay opción")
		})
	}
}

// TestRutaInservible fija lo que vale en **cualquier** modo: un valor vacío
// —venga de la opción o de la variable— y una ruta que existe y no es un
// directorio son «argumentos» (2) tanto si se va a escribir como si solo se va
// a leer, y el mensaje nombra de dónde salió la ruta y cuál era, que es lo
// único con lo que quien invoca puede corregirlo (FR-015, FR-022, US4
// escenarios 3 y 4).
//
// Cada caso comprueba además que el directorio de la cuenta sigue sin existir
// después del fallo: un valor inservible no puede terminar en una caída
// silenciosa a la ruta por omisión (FR-022).
func TestRutaInservible(t *testing.T) {
	cuenta := t.TempDir()
	t.Setenv("HOME", cuenta)
	t.Setenv("USERPROFILE", cuenta)

	fichero := filepath.Join(t.TempDir(), "fichero")
	require.NoError(t, os.WriteFile(fichero, []byte("no soy un directorio"), 0o600))

	casos := []struct {
		nombre    string
		opciones  []Opcion
		variable  string
		declarada bool
		nombra    []string
	}{
		{
			nombre:   "vacía por la opción",
			opciones: []Opcion{ConDirectorio("")},
			nombra:   []string{origenDeLaOpcion},
		},
		{
			nombre:    "vacía por la variable",
			variable:  "",
			declarada: true,
			nombra:    []string{origenDeLaVariable},
		},
		{
			nombre:   "un fichero como directorio",
			opciones: []Opcion{ConDirectorio(fichero)},
			nombra:   []string{origenDeLaOpcion, fichero},
		},
	}

	modos := []struct {
		nombre string
		opcion []Opcion
	}{
		{nombre: "normal"},
		{nombre: "solo lectura", opcion: []Opcion{SoloLectura()}},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			if caso.declarada {
				t.Setenv(VariableDirectorio, caso.variable)
			} else {
				sinVariable(t)
			}

			for _, modo := range modos {
				t.Run(modo.nombre, func(t *testing.T) {
					opciones := append(slices.Clone(caso.opciones), modo.opcion...)

					cliente, err := New(t.Context(), opciones...)

					require.Error(t, err)
					assert.Nil(t, cliente)
					assert.Equal(t, schema.ClaseArgumentos, cli.Clasificar(err))
					assert.Equal(t, 2, cli.CodigoSalida(err))

					for _, texto := range caso.nombra {
						assert.Contains(t, err.Error(), texto,
							"el mensaje nombra de dónde salió la ruta y cuál era")
					}

					_, err = os.Stat(filepath.Join(cuenta, ".cache"))
					assert.ErrorIs(t, err, fs.ErrNotExist,
						"una ruta inservible no puede caer en silencio a la ruta por omisión")
				})
			}
		})
	}
}

// sinVariable deja KITLEGAL_CACHE_DIR sin declarar mientras dura la subprueba.
// t.Setenv no sabe borrar una variable, así que se declara primero —que es lo
// que registra la restauración de lo que hubiera en el entorno al terminar— y
// se borra después.
func sinVariable(t *testing.T) {
	t.Helper()

	t.Setenv(VariableDirectorio, "")
	require.NoError(t, os.Unsetenv(VariableDirectorio))
}
