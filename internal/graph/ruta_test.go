package graph

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cache"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Las tablas que resuelven la ruta por el entorno no declaran t.Parallel(), y
// no es un descuido: usan t.Setenv, que no se puede usar en tests paralelos, y
// el entorno es justo lo que miden. Declaran siempre las tres fuentes de la
// ruta —la opción, la variable y el directorio de la cuenta—, de modo que lo que
// haya en la máquina de quien ejecuta el test no cambie ningún resultado.

// casoDeUbicacion es una fila de TestUbicar: el entorno y las opciones con que
// se resuelve, y la ruta esperada o, si falla, lo que el mensaje nombra.
type casoDeUbicacion struct {
	nombre    string
	opciones  []Opcion
	variable  string
	declarada bool
	sinCuenta bool
	// esperada es la ruta de world.db; vacía si la resolución falla.
	esperada string
	// nombra es lo que el mensaje del fallo contiene además del encabezado.
	nombra []string
	// deLaCache dice si el fallo es el de la regla de la caché, que el mensaje
	// repite tras el encabezado.
	deLaCache bool
}

// TestUbicar fija la ubicación de world.db (FR-001, FR-011;
// contracts/almacen-world-db.md §2 y §6): dentro del directorio de ConDirectorio
// o, sin la opción, del de la regla de la caché —la variable KITLEGAL_CACHE_DIR
// o el directorio de la cuenta—, con los mismos errores de clase «argumentos»
// que la caché y un mensaje que empieza por «grafo: no se puede ubicar
// world.db: ». Resolver no es crear: bajo la raíz temporal no aparece, no
// desaparece y no cambia nada.
func TestUbicar(t *testing.T) {
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
	declarado := filepath.Join(raiz, "declarado")

	casos := []casoDeUbicacion{
		{
			nombre:    "la opción gana a la variable y a la cuenta",
			opciones:  []Opcion{ConDirectorio(declarado)},
			variable:  existente,
			declarada: true,
			esperada:  filepath.Join(declarado, "world.db"),
		},
		{
			nombre:    "la opción no consulta el entorno aunque el entorno no dé ningún directorio",
			opciones:  []Opcion{ConDirectorio(declarado)},
			declarada: true,
			sinCuenta: true,
			esperada:  filepath.Join(declarado, "world.db"),
		},
		{
			nombre:   "la última opción gana",
			opciones: []Opcion{ConDirectorio(existente), ConDirectorio(declarado)},
			esperada: filepath.Join(declarado, "world.db"),
		},
		{
			nombre:    "la opción vacía es un fallo aunque haya variable",
			opciones:  []Opcion{ConDirectorio("")},
			variable:  existente,
			declarada: true,
			nombra:    []string{"ConDirectorio"},
		},
		{
			nombre:    "la variable que nombra un directorio gana a la cuenta",
			variable:  existente,
			declarada: true,
			esperada:  filepath.Join(existente, "world.db"),
		},
		{
			nombre:    "la variable que nombra un directorio que aún no existe",
			variable:  inexistente,
			declarada: true,
			esperada:  filepath.Join(inexistente, "world.db"),
		},
		{
			nombre:   "sin variable, el directorio de la cuenta",
			esperada: filepath.Join(cuenta, ".cache", "kitlegal", "world.db"),
		},
		{
			nombre:    "la variable declarada y vacía",
			declarada: true,
			nombra:    []string{cache.VariableDirectorio},
			deLaCache: true,
		},
		{
			nombre:    "la variable que nombra un fichero",
			variable:  fichero,
			declarada: true,
			nombra:    []string{cache.VariableDirectorio, fichero},
			deLaCache: true,
		},
		{
			nombre:    "sin variable ni directorio de la cuenta",
			sinCuenta: true,
			nombra:    []string{"declara HOME o " + cache.VariableDirectorio},
			deLaCache: true,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			if caso.declarada {
				t.Setenv(cache.VariableDirectorio, caso.variable)
			} else {
				sinVariable(t)
			}

			if caso.sinCuenta {
				t.Setenv("HOME", "")
				t.Setenv("USERPROFILE", "")
			}

			compruebaUbicacion(t, raiz, caso)
		})
	}
}

// compruebaUbicacion es el resto de cada fila de TestUbicar, con el entorno ya
// preparado.
func compruebaUbicacion(t *testing.T, raiz string, caso casoDeUbicacion) {
	t.Helper()

	antes := huellasDelArbol(t, raiz)

	ruta, err := ubicar(operacionLeer, caso.opciones)

	assert.Equal(t, antes, huellasDelArbol(t, raiz), "resolver la ruta no crea ni toca nada")

	if caso.esperada != "" {
		require.NoError(t, err)
		assert.Equal(t, caso.esperada, ruta)

		if len(caso.opciones) == 0 {
			directorio, err := cache.Directorio()
			require.NoError(t, err)
			assert.Equal(t, filepath.Join(directorio, "world.db"), ruta,
				"sin la opción, world.db vive en el directorio de la caché, con su misma regla")
		}

		return
	}

	require.Error(t, err)
	assert.Empty(t, ruta, "un fallo no devuelve ninguna ruta")

	var fallo *Error

	require.ErrorAs(t, err, &fallo)
	assert.Equal(t, schema.ClaseArgumentos, fallo.Clase())
	assert.Equal(t, operacionLeer, fallo.Operacion)
	assert.Empty(t, fallo.Ruta, "sin ubicación no hay ruta que nombrar")
	assert.Contains(t, err.Error(), "grafo: no se puede ubicar world.db: ")

	for _, texto := range caso.nombra {
		assert.Contains(t, err.Error(), texto, "el mensaje dice qué hay que corregir")
	}

	if caso.deLaCache {
		_, deLaCache := cache.Directorio()
		require.Error(t, deLaCache)
		assert.Equal(t, "grafo: no se puede ubicar world.db: "+deLaCache.Error(), err.Error(),
			"el mensaje repite el de la regla de la caché tras el encabezado")

		var causa *cache.Error

		require.ErrorAs(t, err, &causa, "la causa es el error de la caché")
		assert.Equal(t, schema.ClaseArgumentos, causa.Clase(), "con su misma clase")
	}
}

// TestUbicarResuelveAlUsarse fija que la ruta se resuelve cuando se lee o se
// entrega, nunca al construir (contracts/almacen-world-db.md §2): construir una
// opción vacía no falla ni consulta nada, y el entorno que cuenta es el del
// instante en que se resuelve.
func TestUbicarResuelveAlUsarse(t *testing.T) {
	raiz := t.TempDir()
	uno := filepath.Join(raiz, "uno")
	otro := filepath.Join(raiz, "otro")

	t.Setenv(cache.VariableDirectorio, uno)

	vacia := ConDirectorio("")
	require.NotNil(t, vacia, "construir la opción vacía no falla: el fallo llega al resolver")

	ruta, err := ubicar(operacionEscribir, nil)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(uno, "world.db"), ruta)

	t.Setenv(cache.VariableDirectorio, otro)

	ruta, err = ubicar(operacionEscribir, nil)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(otro, "world.db"), ruta, "cada resolución consulta el entorno de ese instante")

	_, err = ubicar(operacionEscribir, []Opcion{vacia})

	var fallo *Error

	require.ErrorAs(t, err, &fallo)
	assert.Equal(t, operacionEscribir, fallo.Operacion, "el fallo lleva la operación que resolvía")
	assert.Equal(t, schema.ClaseArgumentos, fallo.Clase())
	require.ErrorIs(t, err, errDirectorioVacio)
}

// sinVariable deja KITLEGAL_CACHE_DIR sin declarar durante el test y la
// restaura al terminar: t.Setenv guarda el valor que tenía y os.Unsetenv la
// retira.
func sinVariable(t *testing.T) {
	t.Helper()

	t.Setenv(cache.VariableDirectorio, "")
	require.NoError(t, os.Unsetenv(cache.VariableDirectorio))
}

// huellasDelArbol describe todo lo que hay bajo la raíz: cada ruta relativa con
// su tipo, sus permisos y la huella SHA-256 de su contenido si es un fichero. Dos
// descripciones iguales dicen que no apareció, no desapareció y no cambió nada.
func huellasDelArbol(t *testing.T, raiz string) map[string]string {
	t.Helper()

	sistema := os.DirFS(raiz)
	huellas := map[string]string{}

	err := fs.WalkDir(sistema, ".", func(nombre string, entrada fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		info, err := entrada.Info()
		if err != nil {
			return err
		}

		descripcion := info.Mode().String()

		if info.Mode().IsRegular() {
			contenido, err := fs.ReadFile(sistema, nombre)
			if err != nil {
				return err
			}

			suma := sha256.Sum256(contenido)
			descripcion += " " + hex.EncodeToString(suma[:])
		}

		huellas[nombre] = descripcion

		return nil
	})
	if errors.Is(err, fs.ErrNotExist) {
		return map[string]string{}
	}

	require.NoError(t, err, "no se pudo recorrer %s", raiz)

	return huellas
}
