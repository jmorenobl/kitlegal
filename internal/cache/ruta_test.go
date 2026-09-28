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

// Las tablas que resuelven o validan la ruta no declaran t.Parallel(), y no es
// un descuido: usan t.Setenv, que «cannot be used in parallel tests», y el
// entorno es justamente lo que miden. Declaran además siempre el estado de las
// fuentes de la ruta —la opción, la variable y el directorio de la cuenta—, de
// modo que lo que haya en la máquina de quien ejecuta el test no pueda cambiar
// ningún resultado.

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

// casoDeDirectorio es una fila de TestDirectorio: el estado del entorno y lo que
// Directorio tiene que devolver con él. Una fila con nombra espera un fallo, y
// esas son las palabras que su mensaje tiene que llevar; una sin nombra espera
// el directorio de esperado.
type casoDeDirectorio struct {
	nombre    string
	variable  string
	declarada bool
	sinCuenta bool
	esperado  string
	nombra    []string
}

// TestDirectorio fija la regla de ubicación que el paquete exporta para que el
// grafo del mundo viva junto a la caché (contracts/almacen-world-db.md §2 de H7,
// research D8): la variable si está presente y, sin ella, el directorio de la
// cuenta; y, con los mismos valores inservibles que la caché, los mismos errores
// de clase «argumentos» (2), con el mensaje que pide declarar HOME o
// KITLEGAL_CACHE_DIR cuando no hay ni variable ni cuenta (H7 FR-001, FR-011).
//
// «La misma regla» no se afirma solo con las rutas esperadas: cada fila
// construye además la caché de solo lectura con el mismo entorno —sin la opción,
// que es lo que Directorio no tiene— y exige su mismo directorio o su mismo
// error, campo a campo. Así las dos no pueden divergir sin que el test lo diga.
//
// Resolver no es crear: cada fila comprueba que bajo la raíz temporal, donde
// están la cuenta, lo que nombra la variable y el fichero, no aparece, no
// desaparece y no cambia nada. Eso incluye que la variable declarada y vacía no
// cae en silencio al directorio de la cuenta (FR-022 de H3).
func TestDirectorio(t *testing.T) {
	raiz := t.TempDir()

	cuenta := filepath.Join(raiz, "cuenta")
	require.NoError(t, os.Mkdir(cuenta, 0o700))
	t.Setenv("HOME", cuenta)
	t.Setenv("USERPROFILE", cuenta)

	existente := filepath.Join(raiz, "existente")
	require.NoError(t, os.Mkdir(existente, 0o700))

	fichero := filepath.Join(raiz, "fichero")
	require.NoError(t, os.WriteFile(fichero, []byte("no soy un directorio"), 0o600))

	inexistente := filepath.Join(raiz, "inexistente")

	casos := []casoDeDirectorio{
		{
			nombre:    "la variable que nombra un directorio gana al de la cuenta",
			variable:  existente,
			declarada: true,
			esperado:  existente,
		},
		{
			nombre:    "la variable que nombra un directorio que aún no existe",
			variable:  inexistente,
			declarada: true,
			esperado:  inexistente,
		},
		{
			nombre:   "sin variable, el directorio de la cuenta",
			esperado: filepath.Join(cuenta, ".cache", "kitlegal"),
		},
		{
			nombre:    "la variable declarada y vacía",
			variable:  "",
			declarada: true,
			nombra:    []string{origenDeLaVariable},
		},
		{
			nombre:    "la variable que nombra un fichero",
			variable:  fichero,
			declarada: true,
			nombra:    []string{origenDeLaVariable, fichero},
		},
		{
			nombre:    "sin variable ni directorio de la cuenta",
			sinCuenta: true,
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

			compruebaDirectorio(t, raiz, caso)
		})
	}
}

// compruebaDirectorio es el resto de cada fila de TestDirectorio, con el
// entorno ya preparado: resuelve con Directorio y con la caché, y compara las
// dos cosas con lo que la fila espera y entre sí.
func compruebaDirectorio(t *testing.T, raiz string, caso casoDeDirectorio) {
	t.Helper()

	antes := arbolDe(t, raiz)

	directorio, err := Directorio()

	cliente, errDeLaCache := New(t.Context(), SoloLectura())
	if cliente != nil {
		t.Cleanup(func() { require.NoError(t, cliente.Close()) })
	}

	assert.Equal(t, antes, arbolDe(t, raiz), "resolver el directorio no crea ni toca nada")

	if len(caso.nombra) == 0 {
		require.NoError(t, err)
		assert.Equal(t, caso.esperado, directorio)

		require.NoError(t, errDeLaCache)
		assert.Equal(t, cliente.directorio, directorio, "la misma regla que la caché, el mismo directorio")

		return
	}

	require.Error(t, err)
	assert.Empty(t, directorio, "un fallo no devuelve ningún directorio")
	assert.Equal(t, schema.ClaseArgumentos, cli.Clasificar(err))
	assert.Equal(t, 2, cli.CodigoSalida(err))

	for _, texto := range caso.nombra {
		assert.Contains(t, err.Error(), texto,
			"el mensaje dice de dónde salió la ruta y cuál era, o qué hay que declarar")
	}

	assert.Nil(t, cliente)
	assert.Equal(t, errDeLaCache, err, "los mismos errores que la caché: el mismo mensaje, la misma clase y los mismos campos")
}

// TestDirectorioNoCreable fija la fila 6 del contrato de errores y la
// cualificación por modo de FR-022 sobre un directorio que no se puede crear
// porque su padre es un fichero. En modo normal, donde la caché tiene que crear
// la base para servir de algo, es «argumentos» (2) nombrando de dónde salió la
// ruta y cuál era. En solo lectura no es ningún fallo de ruta sino un directorio
// de caché inexistente, y leer es «fuente no disponible» (4), nunca 2 (FR-015,
// SC-011).
//
// Es el caso por el que la regla «inexistente» nombra aparte ENOTDIR: os.Stat no
// devuelve fs.ErrNotExist, y sin esa segunda condición el subtest de solo
// lectura terminaría en 1. En los dos modos el fichero que hace de padre queda
// intacto y bajo el directorio temporal no aparece nada.
//
// Declara t.Parallel() porque la ruta llega por la opción, que no consulta el
// entorno.
func TestDirectorioNoCreable(t *testing.T) {
	t.Parallel()

	t.Run("normal", func(t *testing.T) {
		t.Parallel()

		padre, directorio := directorioBajoUnFichero(t)
		antes := arbolDe(t, padre)

		cliente, err := New(t.Context(), ConDirectorio(directorio))

		require.Error(t, err)
		assert.Nil(t, cliente)
		assert.Equal(t, schema.ClaseArgumentos, cli.Clasificar(err))
		assert.Equal(t, 2, cli.CodigoSalida(err))
		assert.Contains(t, err.Error(), origenDeLaOpcion, "el mensaje nombra de dónde salió la ruta")
		assert.Contains(t, err.Error(), directorio, "…y cuál era")
		assert.Equal(t, antes, arbolDe(t, padre), "el fichero padre queda intacto y no se crea nada")
	})

	t.Run("solo-lectura", func(t *testing.T) {
		t.Parallel()

		padre, directorio := directorioBajoUnFichero(t)
		antes := arbolDe(t, padre)

		cliente, err := New(t.Context(), ConDirectorio(directorio), SoloLectura())
		require.NoError(t, err, "en solo lectura un directorio inexistente no es un fallo de ruta (FR-015)")
		t.Cleanup(func() { require.NoError(t, cliente.Close()) })
		assert.Nil(t, cliente.db, "el cliente queda sin base")

		contenido, presente, err := cliente.Get(t.Context(), claveDePrueba)
		compruebaAusenciaEnSoloLectura(t, contenido, presente, err, claveDePrueba)

		assert.Equal(t, antes, arbolDe(t, padre), "el fichero padre queda intacto y no se crea nada")
	})
}

// directorioBajoUnFichero devuelve un directorio temporal con un fichero dentro
// y una ruta de directorio que cuelga de ese fichero: no existe como directorio
// y no se puede crear.
func directorioBajoUnFichero(t *testing.T) (padre, directorio string) {
	t.Helper()

	padre = t.TempDir()
	fichero := filepath.Join(padre, "fichero")
	require.NoError(t, os.WriteFile(fichero, []byte("no soy un directorio"), 0o600))

	return padre, filepath.Join(fichero, "sub")
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
