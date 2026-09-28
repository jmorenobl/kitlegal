package graph

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/grafo"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// errSinEnlaces es el fallo con que la costura hace fallar la publicación por
// algo que no es un world.db ya publicado: un sistema de ficheros sin enlaces
// duros, por ejemplo.
var errSinEnlaces = errors.New("este sistema de ficheros no admite enlaces duros")

// sinCostura hace fallar la prueba si la entrega llega a publicar: es la costura
// de las entregas que tienen que fallar antes.
func sinCostura(t *testing.T) func(origen, destino string) error {
	t.Helper()

	return func(origen, destino string) error {
		t.Errorf("la entrega no debía llegar a publicar %s como %s", origen, destino)

		return errSinEnlaces
	}
}

// TestPublicar fija la entrega sobre world.db ausente (contracts/almacen-world-db.md
// §4, paso 3; FR-001, FR-003, FR-013): crea cada directorio que falta, en 0700,
// y publica de una vez un world.db en 0600, en WAL, con el esquema en
// la versión 1 y el lote; en el directorio no queda nada más, ni el temporal ni
// un auxiliar.
func TestPublicar(t *testing.T) {
	t.Parallel()

	raiz := t.TempDir()
	directorio := filepath.Join(raiz, "a", "b")
	ruta := filepath.Join(directorio, "world.db")
	lote := loteDelBOE(urlDelBloque, fechaDelBloque, operacionesDelBloque()...)

	require.NoError(t, Nuevo(ConDirectorio(directorio)).Apply(t.Context(), lote))

	for _, cada := range []struct {
		ruta     string
		permisos fs.FileMode
	}{
		{ruta: filepath.Join(raiz, "a"), permisos: fs.ModeDir | permisosDeDirectorio},
		{ruta: directorio, permisos: fs.ModeDir | permisosDeDirectorio},
		{ruta: ruta, permisos: 0o600},
	} {
		info, err := os.Stat(cada.ruta)
		require.NoError(t, err)
		assert.Equal(t, cada.permisos, info.Mode(), "los permisos de %s", cada.ruta)
	}

	assert.Equal(t, []string{"world.db"}, nombresEn(t, directorio))
	assert.Equal(t, []byte{2, 2}, leerFichero(t, ruta)[18:20], "world.db nace en WAL")
	assert.Equal(t, registrosDe(t, lote), grafoGuardado(t, ruta))

	base := abrirBaseDePrueba(t, ruta, "mode=rw&"+pragmaSoloConsultas)
	compruebaEsquema(t, base)
	assert.Equal(t, int64(1), versionDe(t, base))
	require.NoError(t, base.Close())

	assert.Equal(t, []string{"world.db"}, nombresEn(t, directorio), "leer no deja nada")
}

// TestPublicarCuandoOtraInvocacionPublicoAntes fija la primera salida de la
// costura (contracts/almacen-world-db.md §4, paso 3.6, y §7): si os.Link da
// fs.ErrExist porque otro almacén publicó world.db entre medias, el lote se
// aplica sobre ese world.db y quedan las dos observaciones; del temporal no
// queda nada.
func TestPublicarCuandoOtraInvocacionPublicoAntes(t *testing.T) {
	t.Parallel()

	directorio := filepath.Join(t.TempDir(), "cache")
	antigua := loteDelBOE(urlDelBloque, fechaDelBloque, operacionesDelBloque()...)
	reciente := loteDelBOE(urlDelBloque, fechaSiguiente, operacionesDelBloque()...)

	var delEnlace error

	almacen := Nuevo(ConDirectorio(directorio))
	almacen.enlazar = func(origen, destino string) error {
		require.NoError(t, Nuevo(ConDirectorio(directorio)).Apply(t.Context(), reciente), "el otro almacén publica")

		delEnlace = os.Link(origen, destino)

		return delEnlace
	}

	require.NoError(t, almacen.Apply(t.Context(), antigua))
	require.ErrorIs(t, delEnlace, fs.ErrExist, "premisa: el nombre ya estaba ocupado")

	assert.Equal(t, primeraYUltima(registrosDe(t, antigua), registrosDe(t, reciente)),
		grafoGuardado(t, filepath.Join(directorio, "world.db")), "quedan las dos observaciones")
	assert.Equal(t, []string{"world.db"}, nombresEn(t, directorio))
}

// TestPublicarConOtroError fija la segunda salida de la costura
// (contracts/almacen-world-db.md §4, pasos 3.6 y 3.7, §4.1 y §7): con cualquier
// otro error, la entrega falla con «no se puede publicar» y no queda world.db,
// ni ningún world.db-nuevo-* ni sus auxiliares, ni ningún directorio que creó;
// sí el que ya existía y el que otra invocación usa, porque tiene algo dentro.
func TestPublicarConOtroError(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string
		// directorio es el de world.db, relativo a la raíz.
		directorio string
		// preparar deja lo que hay bajo la raíz antes de entregar; nil si nada.
		preparar func(t *testing.T, raiz string) string
		// ocupado, si no está vacío, es el directorio, relativo a la raíz, que
		// crea la entrega y que otra invocación empieza a usar antes de que la
		// entrega falle.
		ocupado string
	}{
		{nombre: "el directorio no existía", directorio: filepath.Join("a", "b", "cache")},
		{nombre: "el directorio ya existía", directorio: "cache", preparar: cacheVacia},
		{
			nombre:     "otra invocación usa un directorio que creó la entrega",
			directorio: filepath.Join("a", "cache"),
			ocupado:    "a",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			raiz := t.TempDir()
			directorio := filepath.Join(raiz, caso.directorio)

			if caso.preparar != nil {
				caso.preparar(t, raiz)
			}

			esperado := huellasDelArbol(t, raiz)
			ocupante := path.Join(caso.ocupado, "de-otra-invocacion")

			if caso.ocupado != "" {
				esperado[caso.ocupado] = (fs.ModeDir | permisosDeDirectorio).String()
				esperado[ocupante] = fs.FileMode(0o600).String() + " " + strings.TrimPrefix(huella(""), "sha256:")
			}

			almacen := Nuevo(ConDirectorio(directorio))
			almacen.enlazar = func(origen, _ string) error {
				require.FileExists(t, origen, "premisa: el temporal está completo al publicarlo")

				if caso.ocupado != "" {
					escribirFichero(t, filepath.Join(raiz, filepath.FromSlash(ocupante)), nil)
				}

				return errSinEnlaces
			}

			err := almacen.Apply(t.Context(), loteDelBOE(urlDelBloque, fechaDelBloque, operacionesDelBloque()...))
			compruebaFalloDeEntrega(t, err, schema.ClaseInesperado, directorio,
				"grafo: no se puede publicar world.db en "+strconv.Quote(directorio)+": "+errSinEnlaces.Error())
			require.ErrorIs(t, err, errSinEnlaces)
			assert.Equal(t, esperado, huellasDelArbol(t, raiz), "no queda nada de lo que creó la entrega")
		})
	}
}

// TestPublicarSinCrearNada fija las entregas sobre world.db ausente que fallan
// antes de crear nada (contracts/almacen-world-db.md §4, pasos 1 y 3.1, §4.1 y
// §7; FR-033): un lote que el grafo vacío rechaza y un contexto ya terminado no
// crean el directorio ni el temporal, y no llegan a publicar.
func TestPublicarSinCrearNada(t *testing.T) {
	t.Parallel()

	terminado, cancela := context.WithCancel(t.Context())
	cancela()

	casos := []struct {
		nombre    string
		ctx       context.Context
		operacion schema.Operacion
		clase     schema.Clase
		// ruta es la de world.db en el mensaje; vacía si el fallo llega antes
		// de resolverla.
		ruta    func(directorio string) string
		mensaje func(ruta string) string
	}{
		{
			nombre:    "un lote que el grafo vacío rechaza",
			ctx:       t.Context(),
			operacion: schema.Arista{Origen: idNorma, Relacion: grafo.RelacionTieneParte, Destino: idBloque},
			clase:     schema.ClaseInesperado,
			ruta:      func(directorio string) string { return filepath.Join(directorio, "world.db") },
			mensaje: func(ruta string) string {
				return "grafo: el lote no entra en " + strconv.Quote(ruta) + `: la arista de "` + idNorma + `" a "` +
					idBloque + `" por "eli:has_part": su origen no es un nodo del lote y el grafo est` +
					"\xc3\xa1 vac\xc3\xado"
			},
		},
		{
			nombre:    "el contexto ya terminado",
			ctx:       terminado,
			operacion: laNorma(nil),
			clase:     schema.ClaseFuenteNoDisponible,
			ruta:      func(string) string { return "" },
			mensaje: func(string) string {
				return "grafo: el plazo termin\xc3\xb3 antes de escribir world.db"
			},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			raiz := t.TempDir()
			directorio := filepath.Join(raiz, "no-existe", "cache")
			antes := huellasDelArbol(t, raiz)

			almacen := Nuevo(ConDirectorio(directorio))
			almacen.enlazar = sinCostura(t)

			err := almacen.Apply(caso.ctx, loteDelBOE(urlDeLaNorma, fechaDelBloque, caso.operacion))

			ruta := caso.ruta(directorio)
			compruebaFalloDeEntrega(t, err, caso.clase, ruta, caso.mensaje(ruta))
			assert.Equal(t, antes, huellasDelArbol(t, raiz), "no se crea nada")
		})
	}
}

// TestPublicarBajoUnEnlaceSinDestino fija un directorio de world.db que es un
// enlace simbólico sin destino: el temporal no se puede crear a través de él, y
// volver al paso 3.2 no lo arreglaría, así que la entrega falla
// enseguida con «no se puede escribir world.db en…», sin reintentarlo hasta el
// plazo y sin crear el destino ni nada más.
func TestPublicarBajoUnEnlaceSinDestino(t *testing.T) {
	t.Parallel()

	raiz := t.TempDir()
	directorio := filepath.Join(raiz, "cache")
	require.NoError(t, os.Symlink(filepath.Join(raiz, "no-existe"), directorio))

	antes := huellasDelArbol(t, raiz)

	conPlazo, cancela := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancela()

	almacen := Nuevo(ConDirectorio(directorio))
	almacen.enlazar = sinCostura(t)

	err := almacen.Apply(conPlazo, loteDelBOE(urlDeLaNorma, fechaDelBloque, laNorma(nil)))

	var fallo *Error

	require.ErrorAs(t, err, &fallo)
	assert.Equal(t, schema.ClaseInesperado, fallo.Clase())
	assert.Equal(t, directorio, fallo.Ruta)
	assert.True(t, strings.HasPrefix(err.Error(), "grafo: no se puede escribir world.db en "+strconv.Quote(directorio)+": "),
		err.Error())
	require.ErrorIs(t, err, fs.ErrNotExist)
	assert.Equal(t, antes, huellasDelArbol(t, raiz))
}

// TestCrearDirectorios fija el paso 3.2 (contracts/almacen-world-db.md §4): lo
// que falta se crea de uno en uno desde el primer antecesor que no existe, en
// 0700, y se anota en el orden en que se crea; un directorio que ya existe
// —también como enlace— no se anota; y bajo un fichero no se crea nada.
func TestCrearDirectorios(t *testing.T) {
	t.Parallel()

	t.Run("desde el primer antecesor que no existe", func(t *testing.T) {
		t.Parallel()

		raiz := t.TempDir()
		esperados := []string{
			filepath.Join(raiz, "a"), filepath.Join(raiz, "a", "b"), filepath.Join(raiz, "a", "b", "c"),
		}

		creados, err := crearDirectorios(esperados[2])
		require.NoError(t, err)
		assert.Equal(t, esperados, creados)

		for _, creado := range creados {
			info, err := os.Stat(creado)
			require.NoError(t, err)
			assert.Equal(t, fs.ModeDir|permisosDeDirectorio, info.Mode(), creado)
		}
	})

	t.Run("uno que ya existe no se anota", func(t *testing.T) {
		t.Parallel()

		creados, err := crearDirectorios(t.TempDir())
		require.NoError(t, err)
		assert.Empty(t, creados)
	})

	t.Run("bajo un fichero no se crea nada", func(t *testing.T) {
		t.Parallel()

		raiz := t.TempDir()
		fichero := filepath.Join(raiz, "fichero")
		escribirFichero(t, fichero, nil)
		antes := huellasDelArbol(t, raiz)

		creados, err := crearDirectorios(filepath.Join(fichero, "a", "b"))
		require.ErrorIs(t, err, syscall.ENOTDIR)
		assert.Empty(t, creados)
		assert.Equal(t, antes, huellasDelArbol(t, raiz))
	})

	t.Run("un enlace sin destino ocupa el nombre y no se anota", func(t *testing.T) {
		t.Parallel()

		raiz := t.TempDir()
		enlace := filepath.Join(raiz, "enlace")
		require.NoError(t, os.Symlink(filepath.Join(raiz, "no-existe"), enlace))

		creados, err := crearDirectorios(filepath.Join(enlace, "a"))
		require.ErrorIs(t, err, fs.ErrNotExist)
		assert.Empty(t, creados)
		assert.NoFileExists(t, filepath.Join(raiz, "no-existe"))
	})
}

// TestPuedeRehacerse fija cuándo volver al paso 3.2 puede arreglar un «no
// existe» (contracts/almacen-world-db.md §4, paso 3.3): si el antecesor más
// cercano que existe es un directorio —otra invocación retiró los que faltan—,
// sí; si es un enlace sin destino o un fichero, no.
func TestPuedeRehacerse(t *testing.T) {
	t.Parallel()

	raiz := t.TempDir()

	enlace := filepath.Join(raiz, "enlace")
	require.NoError(t, os.Symlink(filepath.Join(raiz, "no-existe"), enlace))

	enlaceADirectorio := filepath.Join(raiz, "enlace-a-directorio")
	require.NoError(t, os.Symlink(raiz, enlaceADirectorio))

	fichero := filepath.Join(raiz, "fichero")
	escribirFichero(t, fichero, nil)

	for ruta, puede := range map[string]bool{
		filepath.Join(raiz, "a", "b"): true,
		raiz:                          true,
		filepath.Join(enlaceADirectorio, "a", "b"): true,
		enlace:                      false,
		filepath.Join(enlace, "a"):  false,
		filepath.Join(fichero, "a"): false,
	} {
		assert.Equal(t, puede, puedeRehacerse(ruta), ruta)
	}
}

// TestRetirar fija la limpieza del paso 3.7 (contracts/almacen-world-db.md §4):
// el temporal y sus auxiliares se retiran, y lo que ya no está no es un fallo;
// cada directorio anotado se retira, del más profundo al menos, y uno que no
// está vacío —otra invocación lo usa— o que ya no está tampoco es un fallo;
// cualquier otro fallo se devuelve, nombrando world.db.
func TestRetirar(t *testing.T) {
	t.Parallel()

	t.Run("el temporal y sus auxiliares", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		temporal := filepath.Join(directorio, "world.db-nuevo-1")

		for _, sufijo := range []string{"", sufijoWAL, sufijoMemoriaCompartida, sufijoDiario} {
			escribirFichero(t, temporal+sufijo, nil)
		}

		require.NoError(t, retirarElTemporal(filepath.Join(directorio, "world.db"), temporal))
		assert.Empty(t, nombresEn(t, directorio))
		require.NoError(t, retirarElTemporal(filepath.Join(directorio, "world.db"), temporal), "lo que ya no está no falla")
	})

	t.Run("cada directorio anotado", func(t *testing.T) {
		t.Parallel()

		raiz := t.TempDir()
		creados := []string{filepath.Join(raiz, "a"), filepath.Join(raiz, "a", "b"), filepath.Join(raiz, "c")}

		for _, creado := range creados {
			require.NoError(t, os.Mkdir(creado, 0o700))
		}

		escribirFichero(t, filepath.Join(raiz, "c", "de-otra-invocacion"), nil)

		require.NoError(t, retirarDirectorios(filepath.Join(raiz, "a", "b", "world.db"),
			append(creados, filepath.Join(raiz, "ya-no-esta"))))
		assert.Equal(t, []string{"c"}, nombresEn(t, raiz))
	})

	t.Run("un fallo que no es de otra invocación", func(t *testing.T) {
		t.Parallel()

		raiz := t.TempDir()
		dentro := filepath.Join(raiz, "dentro")
		require.NoError(t, os.Mkdir(dentro, 0o700))
		temporal := filepath.Join(raiz, "world.db-nuevo-1")
		escribirFichero(t, temporal, nil)
		directorioSinEscritura(t, raiz)

		ruta := filepath.Join(raiz, "world.db")

		for nombre, err := range map[string]error{
			"el directorio": retirarDirectorios(ruta, []string{dentro}),
			"el temporal":   retirarElTemporal(ruta, temporal),
		} {
			var fallo *Error

			require.ErrorAs(t, err, &fallo, nombre)
			require.ErrorIs(t, err, fs.ErrPermission, nombre)
			assert.Equal(t, ruta, fallo.Ruta, nombre)
			assert.Contains(t, err.Error(), "world.db", nombre)
		}
	})
}
