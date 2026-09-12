//go:build integration

// Las pruebas de este fichero son lo que solo se puede comprobar con permisos
// reales del sistema de ficheros y con dos procesos: la base que dos invocaciones usan a la
// vez, el directorio en el que no se puede escribir, el que ni siquiera se puede
// recorrer y el registro de escritura que llega sin su memoria compartida
// (FR-015, FR-022, FR-032, SC-004, SC-007, filas 12 y 13 del contrato de errores).
//
// Lleva la etiqueta integration porque depende del entorno y lanza procesos, y
// por eso make ci la ejecuta con test-integration y el lint la alcanza con
// run.build-tags: un fichero etiquetado no entra sin las dos puertas (FR-041,
// D11). Toda base vive en t.TempDir(), también la del otro proceso (SC-009).
//
// El segundo proceso es el propio binario de test relanzado con
// -test.run=^TestProcesoAuxiliar$ y un papel en el entorno, de modo que corre
// con el mismo detector de carreras y sin ningún package main (clarificación Q5).
//
// Va en el paquete externo cache_test, como el adaptador de prueba: lo que aquí
// se mide es lo que ve quien usa la superficie exportada.
package cache_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"

	"github.com/jmorenobl/kitlegal/internal/cache"
	"github.com/jmorenobl/kitlegal/internal/cli"
)

const (
	// variableDelPapel es la que convierte TestProcesoAuxiliar en el otro
	// proceso, y variableDelInforme la ruta en la que ese proceso deja cuántas
	// entradas confirmó o encontró. El informe es lo que impide un verde en
	// vacío: un hijo que no llegara a ejecutar la prueba terminaría también con 0.
	variableDelPapel   = "KITLEGAL_PRUEBA_PAPEL"
	variableDelInforme = "KITLEGAL_PRUEBA_INFORME"

	// Los dos papeles del otro proceso.
	papelEscritor          = "escritor"
	papelLectorSoloLectura = "lector-solo-lectura"

	// entradasEntreProcesos es cuántas entradas confirma el escritor y tiene que
	// encontrar el lector, y vigenciaDeIntegracion la vigencia con la que se
	// guardan: con el reloj real de los dos procesos, de sobra para que ninguna
	// caduque mientras dura la prueba.
	entradasEntreProcesos = 16
	vigenciaDeIntegracion = time.Hour

	// repeticionesPorEntrada hace que cada contenido ocupe varias páginas de
	// SQLite, que es lo que haría visible una entrada leída a medias.
	repeticionesPorEntrada = 1024

	// Los permisos con los que se restringe un directorio: sin escritura, y sin
	// nada.
	permisosDeSoloLectura fs.FileMode = 0o500
	permisosDenegados     fs.FileMode = 0o000
)

// TestIntegracionDosProcesos fija FR-032 y SC-007 entre dos procesos, que es
// como se encuentran dos invocaciones reales: el lector encuentra todas las
// entradas que el escritor confirmó, enteras y sin fallar por bloqueo, mientras
// el escritor sigue con su cliente abierto —y lo confirmado, en su registro de
// escritura—.
//
// Cada subprueba pone a un proceso en cada papel. En la primera el lector es de
// solo lectura, que es el que fallaría con una ausencia falsa (4) si no viera el
// registro de otra invocación (FR-015), y el padre sigue escribiendo mientras el
// hijo lee. En la segunda escribe el hijo, con la ruta de la caché tomada de
// KITLEGAL_CACHE_DIR (FR-020), y el padre lee sin parar mientras tanto.
func TestIntegracionDosProcesos(t *testing.T) {
	t.Parallel()

	t.Run("escritor-padre-lector-solo-lectura-hijo", func(t *testing.T) {
		t.Parallel()

		directorio, informe := t.TempDir(), filepath.Join(t.TempDir(), "informe")
		escritor := cacheAbierta(t, cache.ConDirectorio(directorio))

		require.NoError(t, confirmaEntradas(t.Context(), escritor))
		compruebaRegistroConEscrituras(t, directorio)

		terminado := make(chan struct{})
		escrituras := make(chan error, 1)

		go func() { escrituras <- escribeHastaQueTermine(t.Context(), escritor, terminado) }()

		lector := procesoAuxiliar(t.Context(), binarioDeLaPrueba(t), directorio, papelLectorSoloLectura, informe)
		salida, err := lector.CombinedOutput()

		close(terminado)

		require.NoError(t, <-escrituras, "el escritor no falla por bloqueo mientras el otro proceso lee (FR-032)")
		require.NoError(t, err, "el lector de solo lectura del otro proceso termina bien; su salida:\n%s", salida)
		require.Equal(t, entradasEntreProcesos, leeElInforme(t, informe),
			"el lector del otro proceso encontró enteras todas las entradas confirmadas (SC-007)")
	})

	t.Run("escritor-hijo-lector-padre", func(t *testing.T) {
		t.Parallel()

		directorio, informe := t.TempDir(), filepath.Join(t.TempDir(), "informe")
		lector := cacheAbierta(t, cache.ConDirectorio(directorio))

		escritor := procesoAuxiliar(t.Context(), binarioDeLaPrueba(t), directorio, papelEscritor, informe)

		// Cerrar la entrada estándar es lo que deja al escritor cerrar su cliente:
		// hasta entonces lo confirmado sigue en su registro de escritura.
		entradaDelEscritor, err := escritor.StdinPipe()
		require.NoError(t, err)

		var salida bytes.Buffer

		escritor.Stdout, escritor.Stderr = &salida, &salida

		require.NoError(t, escritor.Start())

		var (
			grupo sync.WaitGroup
			final error
		)

		terminado := make(chan struct{})

		grupo.Go(func() {
			defer close(terminado)

			final = escritor.Wait()
		})

		// Si una comprobación falla a mitad, el contexto de la prueba se cancela
		// antes de las limpiezas y eso termina el otro proceso; esta espera, que
		// corre antes que la del directorio temporal, no deja que se borre con el
		// escritor todavía dentro.
		t.Cleanup(grupo.Wait)

		for !existe(t, informe) {
			select {
			case <-terminado:
				t.Fatalf("el escritor del otro proceso terminó sin confirmar sus entradas: %v; su salida:\n%s",
					final, salida.String())
			default:
			}

			compruebaLoQueHaya(t, lector)
		}

		require.Equal(t, entradasEntreProcesos, leeElInforme(t, informe), "el escritor del otro proceso las confirmó todas")

		for indice := range entradasEntreProcesos {
			compruebaEntradaConfirmada(t, lector, indice)
		}

		require.NoError(t, entradaDelEscritor.Close())
		<-terminado
		require.NoError(t, final, "el escritor del otro proceso termina bien; su salida:\n%s", salida.String())
	})
}

// TestProcesoAuxiliar es el otro proceso de TestIntegracionDosProcesos, y solo
// lo es cuando la variable del papel está en el entorno: sin ella retorna sin
// hacer nada, que es lo que ocurre en toda ejecución normal de la suite (el
// patrón TestHelperProcess de la biblioteca estándar). No es un salto: no hay
// nada que medir en él fuera de su papel.
func TestProcesoAuxiliar(t *testing.T) {
	t.Parallel()

	papel := os.Getenv(variableDelPapel)
	informe := os.Getenv(variableDelInforme)

	switch papel {
	case "":
		return
	case papelEscritor:
		actuaComoEscritor(t, informe)
	case papelLectorSoloLectura:
		actuaComoLectorDeSoloLectura(t, informe)
	default:
		t.Fatalf("%s trae un papel desconocido: %q", variableDelPapel, papel)
	}
}

// TestIntegracionDirectorioNoEscribible fija FR-015 y FR-022 sobre un directorio
// en el que no se puede escribir (0500) y que guarda una base con entradas y sin
// registro de escritura: el contenedor de solo lectura de US3.
//
// En solo lectura se lee con normalidad —código 0, todas las entradas y ningún
// fichero nuevo—, porque SQLite no puede crear la memoria compartida, la primera
// consulta falla con SQLITE_READONLY_DIRECTORY y la apertura se rehace como
// inmutable, que es correcto precisamente porque no hay registro que leer (D5). En
// modo normal una caché que no puede escribir no sirve, y es «argumentos» (2)
// nombrando el directorio y de dónde salió (fila 6).
func TestIntegracionDirectorioNoEscribible(t *testing.T) {
	t.Parallel()

	exigeQueLosPermisosSeHaganValer(t)

	t.Run("solo-lectura-lee", func(t *testing.T) {
		t.Parallel()

		directorio := baseConfirmadaYCerrada(t)
		restringe(t, directorio, permisosDeSoloLectura)

		antes := huellasDe(t, directorio)

		lector, err := cache.New(t.Context(), cache.ConDirectorio(directorio), cache.SoloLectura())
		require.Equal(t, 0, cli.CodigoSalida(err), "el directorio no escribible se lee con normalidad (FR-015): %v", err)
		t.Cleanup(func() { require.NoError(t, lector.Close()) })

		for indice := range entradasEntreProcesos {
			compruebaEntradaConfirmada(t, lector, indice)
		}

		assert.Equal(t, antes, huellasDe(t, directorio), "leer no crea ni cambia ningún fichero (FR-015)")
	})

	t.Run("normal-argumentos", func(t *testing.T) {
		t.Parallel()

		directorio := baseConfirmadaYCerrada(t)
		restringe(t, directorio, permisosDeSoloLectura)

		antes := huellasDe(t, directorio)

		err := intentaAbrir(t, cache.ConDirectorio(directorio))
		compruebaCodigo(t, err, 2, directorio, "opción ConDirectorio")

		assert.Equal(t, antes, huellasDe(t, directorio), "el intento no crea ni cambia ningún fichero")
	})
}

// TestIntegracionDirectorioDenegado fija la fila 12 sobre un directorio que no se
// puede ni recorrer (0000) y que guarda una base con entradas. En solo lectura no
// se puede saber si la base está, y contestar que no hay nada sería mentir: es
// «inesperado» (1) nombrando la base, y nunca la ausencia (4) que un cliente sin
// base daría. En modo normal es «argumentos» (2), como todo directorio en el que
// no se puede escribir (fila 6).
func TestIntegracionDirectorioDenegado(t *testing.T) {
	t.Parallel()

	exigeQueLosPermisosSeHaganValer(t)

	t.Run("solo-lectura-inesperado", func(t *testing.T) {
		t.Parallel()

		directorio := baseConfirmadaYCerrada(t)
		restringe(t, directorio, permisosDenegados)

		err := intentaAbrir(t, cache.ConDirectorio(directorio), cache.SoloLectura())
		compruebaCodigo(t, err, 1, filepath.Join(directorio, ficheroDeLaBase))
		require.ErrorIs(t, err, fs.ErrPermission, "la causa es el acceso denegado, no una ausencia (fila 12)")
	})

	t.Run("normal-argumentos", func(t *testing.T) {
		t.Parallel()

		directorio := baseConfirmadaYCerrada(t)
		restringe(t, directorio, permisosDenegados)

		err := intentaAbrir(t, cache.ConDirectorio(directorio))
		compruebaCodigo(t, err, 2, directorio, "opción ConDirectorio")
	})
}

// TestIntegracionWALSinMemoriaCompartida fija la fila 13: una base cuyo registro
// de escritura trae entradas confirmadas llega, sin su memoria compartida, a un
// directorio en el que no se puede crear (0500). SQLite no puede leer ese
// registro, la primera consulta falla con SQLITE_CANTOPEN y abrir como inmutable
// ignoraría lo que contiene; en solo lectura es «inesperado» (1) nombrando la
// base y los dos auxiliares, y no se crea ni se cambia nada.
//
// El código 14 es lo que hace valer esta prueba: la comprobación del registro se
// dispara ante él y ante 1544, y si solo se disparara ante 1544 este caso caería
// en «fichero inutilizable» con un mensaje que no nombra los auxiliares (D5).
func TestIntegracionWALSinMemoriaCompartida(t *testing.T) {
	t.Parallel()

	exigeQueLosPermisosSeHaganValer(t)

	origen := t.TempDir()
	escritor := cacheAbierta(t, cache.ConDirectorio(origen))

	require.NoError(t, confirmaEntradas(t.Context(), escritor))
	compruebaRegistroConEscrituras(t, origen)

	copia := t.TempDir()

	for _, nombre := range []string{ficheroDeLaBase, registroDeEscritura} {
		copiaElFichero(t, filepath.Join(origen, nombre), filepath.Join(copia, nombre))
	}

	restringe(t, copia, permisosDeSoloLectura)

	antes := huellasDe(t, copia)
	ruta := filepath.Join(copia, ficheroDeLaBase)

	err := intentaAbrir(t, cache.ConDirectorio(copia), cache.SoloLectura())
	compruebaCodigo(t, err, 1, ruta, ruta+"-wal", ruta+"-shm")

	var delControlador *sqlite.Error

	require.ErrorAs(t, err, &delControlador, "la causa es la del controlador de SQLite")
	assert.Equal(t, sqlite3.SQLITE_CANTOPEN, delControlador.Code(),
		"con el registro presente y sin memoria compartida, la primera consulta falla con SQLITE_CANTOPEN (D5)")
	assert.Equal(t, antes, huellasDe(t, copia), "ni se crea la memoria compartida ni cambia ningún fichero")
}

// exigeQueLosPermisosSeHaganValer es la precondición de las pruebas de
// permisos: crea un directorio sin permisos y exige que os.Stat dentro devuelva
// fs.ErrPermission. Donde no lo devuelve —un proceso con privilegios, que no
// respeta los permisos— esas pruebas no medirían nada (SC-004).
//
// Qué hacer entonces depende de dónde corre el proceso. En la integración
// continua (CI no vacía, que GitHub Actions exporta como CI=true) el entorno
// tiene que hacer valer los permisos (supuesto S3), y el fallo lo pone en rojo:
// la receta de test-integration corre sin -v y un salto no dejaría rastro. Fuera
// de ella se salta nombrando la misma causa, porque un puesto inadecuado no es un
// rojo del hito (Complexity Tracking del plan).
func exigeQueLosPermisosSeHaganValer(t *testing.T) {
	t.Helper()

	cerrado := t.TempDir()
	restringe(t, cerrado, permisosDenegados)

	_, err := os.Stat(filepath.Join(cerrado, ficheroDeLaBase))
	if errors.Is(err, fs.ErrPermission) {
		return
	}

	causa := fmt.Sprintf("el sistema de ficheros no hace valer los permisos: os.Stat dentro de un directorio "+
		"0000 devolvió %v y no fs.ErrPermission (¿el proceso corre como root?)", err)

	if os.Getenv("CI") != "" {
		t.Fatalf("%s; en la integración continua los permisos tienen que hacerse valer (supuesto S3)", causa)
	}

	t.Skip(causa)
}

// restringe deja el directorio con los permisos dados y registra la restauración
// de los que tenía. Quien llama lo hace después de t.TempDir(), así que la
// restauración corre antes que la limpieza del directorio temporal, que es un
// RemoveAll y fallaría la prueba sin poder recorrerlo (obligación 6 del plan).
func restringe(t *testing.T, directorio string, permisos fs.FileMode) {
	t.Helper()

	info, err := os.Stat(directorio)
	require.NoError(t, err)

	permisosDeAntes := info.Mode().Perm()

	require.NoError(t, os.Chmod(directorio, permisos))
	t.Cleanup(func() {
		require.NoError(t, os.Chmod(directorio, permisosDeAntes), "se restauran los permisos de %s", directorio)
	})
}

// binarioDeLaPrueba es la ruta del binario de test que se está ejecutando, tal
// como la da el sistema operativo. No sale de os.Args[0], que es un argumento
// más que quien lanza el proceso fija a su gusto y que gosec sigue como dato no
// confiable hasta exec.CommandContext (G702); os.Executable es también lo que usa
// la biblioteca estándar para relanzar sus binarios de test
// (internal/testenv.Executable).
func binarioDeLaPrueba(t *testing.T) string {
	t.Helper()

	ejecutable, err := os.Executable()
	require.NoError(t, err, "el sistema operativo dice qué binario se está ejecutando")

	return ejecutable
}

// procesoAuxiliar prepara el relanzamiento del binario de test en un papel. El
// ejecutable llega como parámetro y el único argumento es un literal, de modo
// que no hay ninguna orden compuesta a partir de datos (G204). El papel, la ruta
// de la caché y la del informe viajan en el entorno.
func procesoAuxiliar(ctx context.Context, ejecutable, directorio, papel, informe string) *exec.Cmd {
	proceso := exec.CommandContext(ctx, ejecutable, "-test.run=^TestProcesoAuxiliar$")
	proceso.Env = append(os.Environ(),
		variableDelPapel+"="+papel,
		cache.VariableDirectorio+"="+directorio,
		variableDelInforme+"="+informe,
	)

	return proceso
}

// actuaComoEscritor es el papel del escritor en el otro proceso: construye la
// caché con la ruta de KITLEGAL_CACHE_DIR, confirma las entradas, lo anuncia en
// el informe y mantiene el cliente abierto hasta que el padre cierra la entrada
// estándar, para que el padre lea mientras lo confirmado sigue en el registro.
func actuaComoEscritor(t *testing.T, informe string) {
	t.Helper()

	escritor := cacheAbierta(t)

	require.NoError(t, confirmaEntradas(t.Context(), escritor))
	escribeElInforme(t, informe, entradasEntreProcesos)

	_, err := io.Copy(io.Discard, os.Stdin)
	require.NoError(t, err, "el padre libera al escritor cerrando la entrada estándar")
}

// actuaComoLectorDeSoloLectura es el papel del lector en el otro proceso: con la
// ruta de KITLEGAL_CACHE_DIR y en solo lectura, encuentra enteras todas las
// entradas y deja en el informe cuántas comprobó.
func actuaComoLectorDeSoloLectura(t *testing.T, informe string) {
	t.Helper()

	lector := cacheAbierta(t, cache.SoloLectura())

	for indice := range entradasEntreProcesos {
		compruebaEntradaConfirmada(t, lector, indice)
	}

	escribeElInforme(t, informe, entradasEntreProcesos)
}

// cacheAbierta construye la caché con las opciones dadas y registra su cierre.
func cacheAbierta(t *testing.T, opciones ...cache.Opcion) *cache.Cliente {
	t.Helper()

	cliente, err := cache.New(t.Context(), opciones...)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, cliente.Close()) })

	return cliente
}

// intentaAbrir construye la caché donde se espera un fallo y lo devuelve. Si la
// construcción no fallara, el cliente se cierra igualmente al terminar.
func intentaAbrir(t *testing.T, opciones ...cache.Opcion) error {
	t.Helper()

	cliente, err := cache.New(t.Context(), opciones...)
	if cliente != nil {
		t.Cleanup(func() { require.NoError(t, cliente.Close()) })
	}

	return err
}

// compruebaCodigo exige el código de salida que el kernel daría al fallo y que el
// mensaje nombre cada una de las cosas dadas.
func compruebaCodigo(t *testing.T, err error, codigo int, nombra ...string) {
	t.Helper()

	require.Equal(t, codigo, cli.CodigoSalida(err), "el fallo sale con el código de su clase: %v", err)

	for _, nombrado := range nombra {
		assert.ErrorContains(t, err, nombrado, "el mensaje nombra lo que explica el fallo")
	}
}

// baseConfirmadaYCerrada deja en un directorio temporal una base con las
// entradas confirmadas y el cliente ya cerrado. Al cerrarse el último cliente
// SQLite vuelca el registro de escritura en la base y retira los dos auxiliares,
// así que todo lo confirmado queda en cache.db, que es el estado del que parten
// las pruebas de permisos que no tratan del registro.
func baseConfirmadaYCerrada(t *testing.T) string {
	t.Helper()

	directorio := t.TempDir()

	escritor, err := cache.New(t.Context(), cache.ConDirectorio(directorio))
	require.NoError(t, err)
	require.NoError(t, errors.Join(confirmaEntradas(t.Context(), escritor), escritor.Close()))

	for _, auxiliar := range []string{registroDeEscritura, memoriaCompartida} {
		_, err := os.Stat(filepath.Join(directorio, auxiliar))
		require.ErrorIs(t, err, fs.ErrNotExist, "al cerrar el último cliente no queda %s", auxiliar)
	}

	return directorio
}

// compruebaRegistroConEscrituras exige que el registro de escritura de la base
// tenga contenido, que es lo que dice que lo confirmado todavía no se volcó en
// cache.db y que el lector va a tener que leerlo de ahí.
func compruebaRegistroConEscrituras(t *testing.T, directorio string) {
	t.Helper()

	info, err := os.Stat(filepath.Join(directorio, registroDeEscritura))
	require.NoError(t, err, "con el escritor abierto hay registro de escritura")
	require.Positive(t, info.Size(), "y lo confirmado está en él")
}

// confirmaEntradas guarda las entradas numeradas; cada una está confirmada en
// cuanto Put vuelve sin error.
func confirmaEntradas(ctx context.Context, escritor *cache.Cliente) error {
	for indice := range entradasEntreProcesos {
		err := escritor.Put(ctx, claveEntreProcesos(indice), cuerpoEntreProcesos(indice), vigenciaDeIntegracion)
		if err != nil {
			return fmt.Errorf("confirmar la entrada %d: %w", indice, err)
		}
	}

	return nil
}

// escribeHastaQueTermine sigue escribiendo una clave aparte hasta que se cierra
// el canal, para que el otro proceso lea mientras hay escrituras en curso.
func escribeHastaQueTermine(ctx context.Context, escritor *cache.Cliente, terminado <-chan struct{}) error {
	for ronda := 0; ; ronda++ {
		select {
		case <-terminado:
			return nil
		default:
		}

		if err := escritor.Put(ctx, "dos-procesos/en-curso", fmt.Appendf(nil, "ronda %d", ronda),
			vigenciaDeIntegracion); err != nil {
			return fmt.Errorf("escritura %d mientras el otro proceso lee: %w", ronda, err)
		}
	}
}

// compruebaLoQueHaya lee todas las entradas mientras el otro proceso las
// escribe: una puede no estar todavía, pero la que está, está entera, y ninguna
// lectura falla por bloqueo.
func compruebaLoQueHaya(t *testing.T, lector *cache.Cliente) {
	t.Helper()

	for indice := range entradasEntreProcesos {
		contenido, presente, err := lector.Get(t.Context(), claveEntreProcesos(indice))
		require.NoError(t, err, "leer mientras otro proceso escribe no falla por bloqueo (FR-032)")

		if presente {
			compruebaEntera(t, indice, contenido)
		}
	}
}

// compruebaEntradaConfirmada exige encontrar entera una entrada que el escritor
// ya confirmó: sin fallo —ni por bloqueo ni por una ausencia falsa, que en solo
// lectura sería el código 4— y byte a byte como se guardó.
func compruebaEntradaConfirmada(t *testing.T, lector *cache.Cliente, indice int) {
	t.Helper()

	contenido, presente, err := lector.Get(t.Context(), claveEntreProcesos(indice))
	require.NoError(t, err, "una entrada confirmada no falla por bloqueo ni se declara ausente (FR-032, SC-007)")
	require.True(t, presente, "la entrada %d está confirmada y tiene que estar", indice)
	compruebaEntera(t, indice, contenido)
}

// compruebaEntera exige que el contenido leído sea exactamente el de la entrada.
func compruebaEntera(t *testing.T, indice int, contenido []byte) {
	t.Helper()

	if esperado := cuerpoEntreProcesos(indice); !bytes.Equal(esperado, contenido) {
		t.Fatalf("la entrada %d llega a medias o cambiada: %d bytes de %d (FR-032)",
			indice, len(contenido), len(esperado))
	}
}

// escribeElInforme deja cuántas entradas se confirmaron o se encontraron. Se
// escribe aparte y se renombra, para que el otro proceso no lo lea a medias.
func escribeElInforme(t *testing.T, informe string, cuantas int) {
	t.Helper()

	require.NotEmpty(t, informe, "%s dice dónde dejar el informe", variableDelInforme)

	provisional := filepath.Clean(informe + ".provisional")
	require.NoError(t, os.WriteFile(provisional, []byte(strconv.Itoa(cuantas)), 0o600))
	require.NoError(t, os.Rename(provisional, filepath.Clean(informe)))
}

// leeElInforme devuelve cuántas entradas dice el informe del otro proceso.
func leeElInforme(t *testing.T, informe string) int {
	t.Helper()

	cuantas, err := strconv.Atoi(string(leeElFichero(t, informe)))
	require.NoError(t, err, "el informe es un número")

	return cuantas
}

// existe dice si hay algo en la ruta, y falla ante cualquier otra respuesta que
// no sea «no está».
func existe(t *testing.T, ruta string) bool {
	t.Helper()

	_, err := os.Stat(ruta)
	if errors.Is(err, fs.ErrNotExist) {
		return false
	}

	require.NoError(t, err)

	return true
}

// copiaElFichero copia un fichero de la prueba con acceso reservado a la cuenta.
func copiaElFichero(t *testing.T, desde, hasta string) {
	t.Helper()

	require.NoError(t, os.WriteFile(filepath.Clean(hasta), leeElFichero(t, desde), 0o600))
}

// huellasDe describe todo lo que hay en un directorio, auxiliares incluidos,
// con la huella SHA-256 de cada fichero: lo que no puede cambiar cuando nada se
// escribe.
func huellasDe(t *testing.T, directorio string) []string {
	t.Helper()

	entradas, err := os.ReadDir(directorio)
	require.NoError(t, err)

	huellas := make([]string, 0, len(entradas))

	for _, entrada := range entradas {
		suma := sha256.Sum256(leeElFichero(t, filepath.Join(directorio, entrada.Name())))
		huellas = append(huellas, entrada.Name()+" sha256:"+hex.EncodeToString(suma[:]))
	}

	return huellas
}

// claveEntreProcesos y cuerpoEntreProcesos son la clave y el contenido de la
// entrada que ocupa una posición. Cada contenido repite su número a lo largo de
// varias páginas, así que ninguno responde por otro y uno a medias no coincide.
func claveEntreProcesos(indice int) string {
	return fmt.Sprintf("dos-procesos/%d", indice)
}

func cuerpoEntreProcesos(indice int) []byte {
	return bytes.Repeat(fmt.Appendf(nil, "<entrada %04d>", indice), repeticionesPorEntrada)
}
