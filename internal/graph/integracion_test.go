//go:build integration

// Las pruebas de este fichero son la matriz de integración del almacén del
// grafo del mundo por su API pública (FR-088; contracts/almacen-world-db.md §7
// de H7.1; research.md D33, V44 y V49 de H7): lo que ve quien usa Nuevo, Apply,
// Leer y los verbos de una Lectura sobre un world.db de verdad, con otras
// conexiones abiertas a la vez y con los auxiliares que deja un escritor
// interrumpido.
//
// Lleva la etiqueta integration, como la matriz de la caché: make ci la ejecuta
// con test-integration y el lint la alcanza con run.build-tags. Todo world.db
// vive bajo t.TempDir(), nunca en el directorio de la cuenta, y el permiso del
// directorio que una prueba quita se restaura cuando la prueba acaba, antes de
// que t.TempDir lo retire.
//
// Los estados se preparan con la propia API siempre que se puede —el world.db
// que deja una entrega, el WAL que deja una entrega mientras otra invocación
// sigue abierta— y con una conexión de SQLite de la prueba solo para lo que
// ninguna entrega completa escribe: otra invocación que retiene world.db, un
// esquema posterior o una creación interrumpida después de poner world.db en
// WAL.
//
// Va en el paquete externo graph_test: lo que se mide es lo que ve quien usa la
// superficie exportada.
package graph_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core"
	"github.com/jmorenobl/kitlegal/internal/core/grafo"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/graph"
)

const (
	// Las fuentes, las urls y los ids son los de las dos consultas que emiten:
	// boe articulo a21 de la LPAC y territorio resolver Leganés
	// (contracts/emision.md).
	fuenteDelBOE        = "boe.legislacion-consolidada"
	fuenteDelTerritorio = "kitlegal.territorio"

	urlDeLaNorma     = "https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565"
	urlDelBloque     = urlDeLaNorma + "/texto/bloque/a21"
	urlDelTerritorio = "https://www.ine.es/daco/daco42/codmun/diccionario26.xlsx"

	idNorma     = "eli/es/l/2015/10/01/39"
	idBloque    = idNorma + "#a21"
	idMunicipio = "ine:28074"
	idOrgano    = "L01280745"
	// idAusente no es el id de ningún nodo de ninguna prueba.
	idAusente = "eli/es/l/2099/01/01/1"

	// fechaAntigua y fechaEnMadrid son el mismo instante escrito de dos formas;
	// fechaReciente es posterior a los dos.
	fechaAntigua       = "2026-09-28T10:00:00Z"
	fechaEnMadrid      = "2026-09-28T12:00:00+02:00"
	fechaReciente      = "2026-09-30T08:00:00Z"
	fechaDelTerritorio = "2026-02-04T00:00:00Z"

	cuerpo2016 = "Texto del bloque a21 en su versi\xc3\xb3n de 2016"
	cuerpo2025 = "Texto del bloque a21 en su versi\xc3\xb3n de 2025"

	semana = 7 * 24 * time.Hour

	// El nombre de la base y los sufijos con que SQLite nombra sus auxiliares
	// en WAL.
	ficheroDelGrafo         = "world.db"
	sufijoWAL               = "-wal"
	sufijoMemoriaCompartida = "-shm"

	// operacionLeer y operacionEscribir son los valores de graph.Error.Operacion.
	operacionLeer     = "leer"
	operacionEscribir = "escribir"

	// tramo es lo que cada conexión de la prueba espera dentro de SQLite, lo
	// mismo que espera el almacén en cada intento.
	tramo = "_pragma=busy_timeout(100)"

	// lecturaSinWAL son los ajustes de la cadena con que la lectura abre un
	// world.db sin -wal (contracts/almacen-world-db.md §3, paso 2): lee sin
	// poder escribir y no deja ningún auxiliar al cerrar (H7.1 research.md V8).
	lecturaSinWAL = "mode=rw&" + tramo + "&_pragma=query_only(1)"

	// Las dos formas en que la otra invocación de una prueba retiene world.db:
	// con una transacción de escritura, que no deja entregar a nadie, y además
	// en locking_mode EXCLUSIVE, que tampoco deja abrirlo a ningún lector
	// (H7.1 research.md D22, V10).
	reteniendoLaEscritura = tramo + "&_txlock=immediate"
	enExclusiva           = tramo + "&_pragma=locking_mode(EXCLUSIVE)&_txlock=immediate"

	// esperaPropia es la del almacén (contracts/almacen-world-db.md §4,
	// «Esperas»), y plazoCorto, un plazo que termina mucho antes que ella.
	esperaPropia = 5 * time.Second
	plazoCorto   = 300 * time.Millisecond

	// plazoTerminado pide a contextoCon un contexto que ya terminó.
	plazoTerminado time.Duration = -1

	// Los permisos que las pruebas ponen y quitan.
	permisosDeLaCuenta   fs.FileMode = 0o600
	permisosDeDirectorio fs.FileMode = 0o700
	permisosSinEscritura fs.FileMode = 0o500
)

// codigoDeLaClase es el código de salida con que el kernel traduce cada clase
// que el almacén declara (ADR 0023; contracts/almacen-world-db.md §6): el que
// darían los verbos de graph con ese fallo.
var codigoDeLaClase = map[schema.Clase]int{
	schema.ClaseArgumentos:         2,
	schema.ClaseFuenteNoDisponible: 4,
	schema.ClaseInesperado:         1,
}

// ---------------------------------------------------------------------------
// Los lotes.

// loteDelBOE es el lote de una consulta al BOE con la url y la fecha dadas y la
// vigencia de boe articulo.
func loteDelBOE(url, fecha string, operaciones ...schema.Operacion) core.Lote {
	return core.Lote{Fuente: fuenteDelBOE, URL: url, FechaConsulta: fecha, Vigencia: semana, Operaciones: operaciones}
}

// laNorma es el nodo de la norma con los datos dados.
func laNorma(datos map[string]any) schema.Nodo {
	return schema.Nodo{ID: idNorma, Tipo: grafo.TipoNorma, Datos: datos}
}

// loteDelBloque es el de boe articulo a21 con la url y la fecha dadas: la
// norma, su bloque y la versión de 2016 del bloque, las dos aristas que los
// unen y el texto de la versión.
func loteDelBloque(url, fecha string) core.Lote {
	huella := huellaDe(cuerpo2016)
	version := idBloque + "@20161002:" + huella

	return loteDelBOE(url, fecha,
		laNorma(map[string]any{grafo.DatoIdentificador: "BOE-A-2015-10565"}),
		schema.Nodo{ID: idBloque, Tipo: grafo.TipoBloque, Datos: map[string]any{grafo.DatoBloque: "a21"}},
		schema.Nodo{ID: version, Tipo: grafo.TipoBloqueVersion, Datos: map[string]any{
			grafo.DatoFechaVigencia: "20161002", grafo.DatoHashTexto: huella,
		}},
		schema.Arista{Origen: idNorma, Relacion: grafo.RelacionTieneParte, Destino: idBloque},
		schema.Arista{Origen: idBloque, Relacion: grafo.RelacionTieneVersion, Destino: version},
		schema.Texto{Huella: huella, Cuerpo: cuerpo2016},
	)
}

// loteDelTerritorio es el de territorio resolver Leganés con el órgano y la
// fecha dados: el municipio, el órgano de su ayuntamiento y la arista que los
// une, sin vigencia.
func loteDelTerritorio(organo, fecha string) core.Lote {
	return core.Lote{
		Fuente: fuenteDelTerritorio, URL: urlDelTerritorio, FechaConsulta: fecha,
		Operaciones: []schema.Operacion{
			schema.Nodo{ID: idMunicipio, Tipo: grafo.TipoMunicipio, Datos: map[string]any{
				grafo.DatoCodigoINE: "28074", grafo.DatoNombre: "Legan\xc3\xa9s",
			}},
			schema.Nodo{ID: organo, Tipo: grafo.TipoOrgano, Datos: map[string]any{grafo.DatoDIR3: organo}},
			schema.Arista{Origen: organo, Relacion: grafo.RelacionPerteneceA, Destino: idMunicipio},
		},
	}
}

// elBloque es la consulta del bloque con la que empiezan casi todas las
// pruebas.
func elBloque() core.Lote {
	return loteDelBloque(urlDelBloque, fechaAntigua)
}

// elTerritorio es la consulta del territorio.
func elTerritorio() core.Lote {
	return loteDelTerritorio(idOrgano, fechaDelTerritorio)
}

// huellaDe es la huella del cuerpo de un texto: «sha256:» y los 64
// hexadecimales en minúscula.
func huellaDe(cuerpo string) string {
	suma := sha256.Sum256([]byte(cuerpo))

	return "sha256:" + hex.EncodeToString(suma[:])
}

// ---------------------------------------------------------------------------
// La API.

// rutaEn es la de world.db dentro del directorio.
func rutaEn(directorio string) string {
	return filepath.Join(directorio, ficheroDelGrafo)
}

// almacenEn es el almacén con world.db en el directorio.
func almacenEn(directorio string) *graph.Almacen {
	return graph.Nuevo(graph.ConDirectorio(directorio))
}

// entregar aplica los lotes, uno tras otro, y exige que entren.
func entregar(t *testing.T, directorio string, lotes ...core.Lote) {
	t.Helper()

	almacen := almacenEn(directorio)

	for _, lote := range lotes {
		require.NoError(t, almacen.Apply(t.Context(), lote))
	}
}

// grafoEntregado deja en el directorio el world.db que crean las entregas de
// los lotes —en WAL, con el esquema y sin ningún auxiliar— y devuelve su ruta.
func grafoEntregado(t *testing.T, directorio string, lotes ...core.Lote) string {
	t.Helper()

	entregar(t, directorio, lotes...)

	ruta := rutaEn(directorio)
	compruebaSinAuxiliares(t, ruta)

	return ruta
}

// contextoCon es el contexto de quien lee o entrega: el de la prueba si el
// plazo es cero, uno que ya terminó con plazoTerminado y, con cualquier otro,
// uno que termina en ese plazo.
func contextoCon(t *testing.T, plazo time.Duration) context.Context {
	t.Helper()

	switch {
	case plazo == 0:
		return t.Context()
	case plazo < 0:
		terminado, cancela := context.WithCancel(t.Context())
		cancela()

		return terminado
	}

	conPlazo, cancela := context.WithTimeout(t.Context(), plazo)
	t.Cleanup(cancela)

	return conPlazo
}

// leido es todo lo que la API deja leer de un grafo: el recuento de graph stats,
// la instantánea de graph check y la ficha de graph show de cada nodo.
type leido struct {
	recuento    grafo.Recuento
	instantanea grafo.Instantanea
	fichas      map[string]grafo.Ficha
}

// grafoVacio es lo que se lee de un grafo ausente o sin esquema: tres ceros,
// listas vacías y ninguna ficha.
func grafoVacio() leido {
	return leido{
		recuento:    grafo.Recuento{NodosPorTipo: []grafo.RecuentoDeNodos{}, AristasPorRelacion: []grafo.RecuentoDeAristas{}},
		instantanea: grafo.Instantanea{Nodos: []grafo.NodoDeInstantanea{}, Aristas: []schema.Arista{}},
		fichas:      map[string]grafo.Ficha{},
	}
}

// leerElGrafo abre la lectura del directorio, la lee con los tres verbos —la
// ficha de cada nodo de la instantánea y la de un id que no está— y la cierra.
func leerElGrafo(t *testing.T, directorio string) leido {
	t.Helper()

	lectura, err := graph.Leer(t.Context(), graph.ConDirectorio(directorio))
	require.NoError(t, err)

	recuento, err := lectura.Recuento(t.Context())
	require.NoError(t, err)

	instantanea, err := lectura.Instantanea(t.Context())
	require.NoError(t, err)

	fichas := map[string]grafo.Ficha{}

	for _, nodo := range instantanea.Nodos {
		ficha, encontrado, err := lectura.Ficha(t.Context(), nodo.ID)
		require.NoError(t, err)
		require.True(t, encontrado, "la ficha de %s, que está en la instantánea", nodo.ID)

		fichas[nodo.ID] = ficha
	}

	_, encontrado, err := lectura.Ficha(t.Context(), idAusente)
	require.NoError(t, err)
	require.False(t, encontrado, "un id que no está no tiene ficha")

	require.NoError(t, lectura.Close())

	return leido{recuento: recuento, instantanea: instantanea, fichas: fichas}
}

// leidoTras es lo que se lee tras entregar los lotes, en su orden, en un
// directorio nuevo: la referencia de lo que un grafo tiene confirmado.
func leidoTras(t *testing.T, lotes ...core.Lote) leido {
	t.Helper()

	directorio := t.TempDir()
	entregar(t, directorio, lotes...)

	return leerElGrafo(t, directorio)
}

// compruebaFallo exige que el fallo sea un *graph.Error de la operación y la
// clase dadas, con el mensaje exacto y el código de salida de esa clase.
func compruebaFallo(t *testing.T, err error, operacion string, clase schema.Clase, mensaje string) {
	t.Helper()

	fallo := falloDe(t, err)
	assert.Equal(t, operacion, fallo.Operacion)
	assert.Equal(t, clase, fallo.Clase())
	assert.Equal(t, mensaje, err.Error())
	assert.Equal(t, codigoDeLaClase[clase], cli.CodigoSalida(err), "el código de salida de la clase %s", clase)
}

// falloDe es el *graph.Error que lleva el fallo.
func falloDe(t *testing.T, err error) *graph.Error {
	t.Helper()

	var fallo *graph.Error

	require.ErrorAs(t, err, &fallo)

	return fallo
}

// Los mensajes de contracts/almacen-world-db.md §6 que las pruebas esperan.

func mensajeInutilizable(ruta string, fallo *graph.Error) string {
	return "grafo: " + strconv.Quote(ruta) + " no es una base de datos utilizable: " + fallo.Causa.Error()
}

func mensajeDeVersionPosterior(ruta string) string {
	return "grafo: " + strconv.Quote(ruta) + " tiene el esquema en la versi\xc3\xb3n 2 y este binario conoce la 1: no se modifica"
}

func mensajeDePlazo(operacion, nombre string) string {
	return "grafo: el plazo termin\xc3\xb3 antes de " + operacion + " " + nombre
}

func mensajeDeBloqueo(_, nombre string) string {
	return "grafo: " + nombre + " est\xc3\xa1 bloqueada por otra invocaci\xc3\xb3n y la espera de 5s se agot\xc3\xb3"
}

func mensajeDelLoteRechazado(ruta string, fallo *graph.Error) string {
	return "grafo: el lote no entra en " + strconv.Quote(ruta) + ": " + fallo.Causa.Error()
}

// ---------------------------------------------------------------------------
// Los ficheros.

// estadoDelArbol describe cada entrada bajo la raíz (describir): dos estados
// iguales son un árbol en el que nada apareció, cambió ni desapareció.
func estadoDelArbol(t *testing.T, raiz string) map[string]string {
	t.Helper()

	estado := map[string]string{}

	err := filepath.WalkDir(raiz, func(ruta string, entrada fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		info, err := entrada.Info()
		if err != nil {
			return err
		}

		estado[ruta] = describir(t, ruta, info)

		return nil
	})
	require.NoError(t, err)

	return estado
}

// describir es una entrada por su tipo y sus permisos y, si es un fichero, por
// la huella de sus bytes.
func describir(t *testing.T, ruta string, info fs.FileInfo) string {
	t.Helper()

	descripcion := info.Mode().String()
	if !info.Mode().IsRegular() {
		return descripcion
	}

	return descripcion + " " + huellaDelFichero(t, ruta)
}

// huellaDelFichero es la SHA-256 de los bytes del fichero.
func huellaDelFichero(t *testing.T, ruta string) string {
	t.Helper()

	suma := sha256.Sum256(leerFichero(t, ruta))

	return hex.EncodeToString(suma[:])
}

// huellasDe es la huella de cada fichero.
func huellasDe(t *testing.T, rutas ...string) map[string]string {
	t.Helper()

	huellas := map[string]string{}

	for _, ruta := range rutas {
		huellas[ruta] = huellaDelFichero(t, ruta)
	}

	return huellas
}

// leerFichero son los bytes de un fichero de la prueba.
func leerFichero(t *testing.T, ruta string) []byte {
	t.Helper()

	contenido, err := os.ReadFile(filepath.Clean(ruta))
	require.NoError(t, err)

	return contenido
}

// escribirFichero escribe un fichero de la prueba con acceso reservado a la
// cuenta.
func escribirFichero(t *testing.T, ruta string, contenido []byte) {
	t.Helper()

	require.NoError(t, os.WriteFile(filepath.Clean(ruta), contenido, permisosDeLaCuenta))
}

// copiarFicheros copia, byte a byte, el fichero de origen con cada sufijo en el
// de destino con el mismo sufijo.
func copiarFicheros(t *testing.T, origen, destino string, sufijos ...string) {
	t.Helper()

	for _, sufijo := range sufijos {
		escribirFichero(t, destino+sufijo, leerFichero(t, origen+sufijo))
	}
}

// tamano es el tamaño del fichero en bytes.
func tamano(t *testing.T, ruta string) int64 {
	t.Helper()

	info, err := os.Stat(ruta)
	require.NoError(t, err)

	return info.Size()
}

// nombresEn son los nombres de las entradas del directorio, ordenados.
func nombresEn(t *testing.T, directorio string) []string {
	t.Helper()

	entradas, err := os.ReadDir(directorio)
	require.NoError(t, err)

	nombres := make([]string, 0, len(entradas))

	for _, entrada := range entradas {
		nombres = append(nombres, entrada.Name())
	}

	return nombres
}

// compruebaSinAuxiliares exige que junto a world.db no quede ni su -wal ni su
// -shm.
func compruebaSinAuxiliares(t *testing.T, ruta string) {
	t.Helper()

	for _, sufijo := range []string{sufijoWAL, sufijoMemoriaCompartida} {
		require.NoFileExists(t, ruta+sufijo)
	}
}

// restringir cambia los permisos de la ruta y devuelve los de antes al terminar
// la prueba, antes de que t.TempDir la retire.
func restringir(t *testing.T, ruta string, permisos fs.FileMode) {
	t.Helper()

	info, err := os.Stat(ruta)
	require.NoError(t, err)

	permisosDeAntes := info.Mode().Perm()

	require.NoError(t, os.Chmod(ruta, permisos))
	t.Cleanup(func() {
		assert.NoError(t, os.Chmod(ruta, permisosDeAntes), "se restauran los permisos de %s", ruta)
	})
}

// directorioSinEscritura quita al directorio el permiso de escritura y comprueba
// la premisa: no admite ficheros nuevos.
func directorioSinEscritura(t *testing.T, directorio string) {
	t.Helper()

	restringir(t, directorio, permisosSinEscritura)

	sonda := filepath.Join(directorio, "sonda")

	err := os.WriteFile(sonda, nil, permisosDeLaCuenta)
	if err == nil {
		require.NoError(t, os.Remove(sonda))
	}

	require.ErrorIs(t, err, fs.ErrPermission, "premisa: el directorio no admite ficheros nuevos (¿corre como root?)")
}

// ---------------------------------------------------------------------------
// Las otras conexiones: otra invocación, otro programa.

// conexion abre la base de la ruta con los ajustes de la cadena de conexión que
// la prueba necesita, con una sola conexión, y la cierra al terminar. Protege en
// la ruta lo que la sintaxis de URI de SQLite interpreta.
func conexion(t *testing.T, ruta, ajustes string) *sql.DB {
	t.Helper()

	protegida := strings.NewReplacer("%", "%25", "?", "%3F", "#", "%23").Replace(filepath.ToSlash(ruta))

	base, err := sql.Open("sqlite", "file:"+protegida+"?"+ajustes)
	require.NoError(t, err)

	base.SetMaxOpenConns(1)
	t.Cleanup(func() { assert.NoError(t, base.Close()) })

	return base
}

// ejecutar ejecuta las sentencias sobre la base con otra conexión y la cierra.
func ejecutar(t *testing.T, ruta, sentencias string) {
	t.Helper()

	base := conexion(t, ruta, tramo)

	_, err := base.ExecContext(t.Context(), sentencias)
	require.NoError(t, err)
	require.NoError(t, base.Close())
}

// otraInvocacionAbierta abre sobre world.db la conexión de otra invocación, lee
// con ella —así retiene el WAL— y la deja abierta hasta el final de la prueba.
// Mientras siga abierta, ninguna entrega es la última conexión: la que cierra no
// lleva el WAL a world.db ni retira sus auxiliares (research.md V36 F).
func otraInvocacionAbierta(t *testing.T, ruta string) *sql.DB {
	t.Helper()

	otra := conexion(t, ruta, tramo+"&_pragma=wal_autocheckpoint(0)&_txlock=immediate")

	var objetos int

	require.NoError(t, otra.QueryRowContext(t.Context(), `SELECT count(*) FROM sqlite_schema`).Scan(&objetos))

	return otra
}

// retener abre la conexión de otra invocación con los ajustes dados
// —reteniendoLaEscritura o enExclusiva— y empieza en ella una transacción, en la
// que lee, hasta que se llama a lo que devuelve, que la deshace y cierra la
// conexión, o hasta que termina la prueba. En exclusiva, world.db no puede tener
// abierta ninguna otra conexión: con la base en WAL, cada una conserva su
// bloqueo compartido mientras sigue abierta.
func retener(t *testing.T, ruta, ajustes string) (suelta func()) {
	t.Helper()

	otra := conexion(t, ruta, ajustes)

	tx, err := otra.BeginTx(context.WithoutCancel(t.Context()), nil)
	require.NoError(t, err, "la otra invocación toma world.db")

	var objetos int

	require.NoError(t, tx.QueryRowContext(t.Context(), `SELECT count(*) FROM sqlite_schema`).Scan(&objetos))

	var una sync.Once

	suelta = func() {
		una.Do(func() {
			assert.NoError(t, tx.Rollback())
			assert.NoError(t, otra.Close())
		})
	}

	t.Cleanup(suelta)

	return suelta
}

// enWALSinTablas deja en la ruta lo que deja una creación de world.db que se
// interrumpe después de ponerlo en WAL (contracts/almacen-world-db.md §4, pasos
// 3 y 4; H7.1 FR-071): el fichero en WAL, sin ninguna tabla y sin auxiliares.
func enWALSinTablas(t *testing.T, ruta string) {
	t.Helper()

	escribirFichero(t, ruta, nil)

	base := conexion(t, ruta, tramo)

	var modo string

	require.NoError(t, base.QueryRowContext(t.Context(), "PRAGMA journal_mode=WAL").Scan(&modo))
	require.Equal(t, "wal", modo)
	require.NoError(t, base.Close())

	compruebaSinAuxiliares(t, ruta)
	require.Equal(t, []byte{2, 2}, leerFichero(t, ruta)[18:20], "premisa: world.db está en WAL")
}

// columna son los valores de la única columna de la consulta, en su orden.
func columna[T any](t *testing.T, base *sql.DB, consulta string) []T {
	t.Helper()

	filas, err := base.QueryContext(t.Context(), consulta)
	require.NoError(t, err)

	defer func() { require.NoError(t, filas.Close()) }()

	var valores []T

	for filas.Next() {
		var valor T

		require.NoError(t, filas.Scan(&valor))

		valores = append(valores, valor)
	}

	require.NoError(t, filas.Err())

	return valores
}

// esquemaDe son las tablas de la base, por nombre, y las versiones que registra
// su schema_version, leídas con la cadena de la lectura de un world.db sin
// -wal, que no escribe nada ni deja ningún auxiliar al cerrar. Quien la llama
// ya ha comprobado que no queda ningún -wal, así que dicen lo que está en
// world.db.
func esquemaDe(t *testing.T, ruta string) (tablas []string, versiones []int64) {
	t.Helper()

	base := conexion(t, ruta, lecturaSinWAL)
	tablas = columna[string](t, base, `SELECT name FROM sqlite_schema WHERE type = 'table' ORDER BY name`)

	if slices.Contains(tablas, "schema_version") {
		versiones = columna[int64](t, base, `SELECT version FROM schema_version ORDER BY version`)
	}

	require.NoError(t, base.Close())

	return tablas, versiones
}

// tablasDelGrafo son las de la versión 1 del esquema.
var tablasDelGrafo = []string{"edges", "nodes", "schema_version", "texts"}

// ---------------------------------------------------------------------------
// Los auxiliares de un escritor interrumpido.

// escrituraEnElWAL es lo que se confirma en el WAL de un world.db mientras otra
// invocación sigue abierta sobre él.
type escrituraEnElWAL func(t *testing.T, directorio string, otra *sql.DB)

// elTerritorioEnElWAL confirma en el WAL la entrega del territorio.
func elTerritorioEnElWAL(t *testing.T, directorio string, _ *sql.DB) {
	t.Helper()

	entregar(t, directorio, elTerritorio())
}

// unEsquemaPosteriorEnElWAL confirma en el WAL, desde la otra invocación, la
// versión 2 del esquema: la de un binario posterior.
func unEsquemaPosteriorEnElWAL(t *testing.T, _ string, otra *sql.DB) {
	t.Helper()

	_, err := otra.ExecContext(t.Context(), `INSERT INTO schema_version VALUES (2, '2030-01-01T00:00:00Z')`)
	require.NoError(t, err)
}

// copiaConElWAL deja en destino —la ruta de la base— la copia, con los sufijos
// dados, de los ficheros de un world.db con el bloque en la base y la escritura
// confirmada en el WAL mientras otra invocación sigue abierta: los de un
// escritor abierto con marcos en el WAL. Sin ninguna conexión sobre la copia,
// su -wal es un -wal huérfano (research.md V36 D, V44).
func copiaConElWAL(t *testing.T, destino string, escribir escrituraEnElWAL, sufijos ...string) {
	t.Helper()

	origen := t.TempDir()
	ruta := grafoEntregado(t, origen, elBloque())
	otra := otraInvocacionAbierta(t, ruta)
	escribir(t, origen, otra)

	require.Positive(t, tamano(t, ruta+sufijoWAL), "premisa: lo confirmado está en el WAL y no en world.db")
	require.FileExists(t, ruta+sufijoMemoriaCompartida)

	copiarFicheros(t, ruta, destino, sufijos...)
	require.NoError(t, otra.Close())
}

// ---------------------------------------------------------------------------
// La matriz.

// TestIntegracionEsquema fija la creación del esquema (H7 FR-004, FR-013; H7.1
// FR-071; contracts/almacen-world-db.md §3, paso 1, y §4, pasos 2 a 5): la
// primera entrega sobre un directorio que no existe crea world.db en WAL, con el
// esquema y el lote y sin ningún auxiliar; y lo que deja una creación
// interrumpida —un world.db de 0 bytes, o en WAL y sin tablas— se lee como el
// grafo vacío sin cambiar nada, y la entrega siguiente crea en él el esquema.
func TestIntegracionEsquema(t *testing.T) {
	t.Parallel()

	t.Run("la primera entrega crea world.db con el esquema y el lote", func(t *testing.T) {
		t.Parallel()

		directorio := filepath.Join(t.TempDir(), "no-existe", "cache")
		entregar(t, directorio, elBloque())
		compruebaEsquemaCreado(t, directorio)

		leidoDelBloque := leerElGrafo(t, directorio)
		assert.Equal(t, recuentoDelBloque(), leidoDelBloque.recuento)
		assert.Equal(t, fichaDeLaNorma(), leidoDelBloque.fichas[idNorma])
	})

	for nombre, preparar := range map[string]func(t *testing.T, ruta string){
		"un world.db de 0 bytes": func(t *testing.T, ruta string) {
			t.Helper()

			escribirFichero(t, ruta, nil)
		},
		"un world.db en WAL y sin tablas": enWALSinTablas,
	} {
		t.Run(nombre+" se lee vacío sin cambiar y la entrega crea el esquema", func(t *testing.T) {
			t.Parallel()

			directorio := t.TempDir()
			preparar(t, rutaEn(directorio))
			compruebaLecturaSinCambios(t, directorio, grafoVacio())

			entregar(t, directorio, elBloque())
			assert.Equal(t, leidoTras(t, elBloque()), leerElGrafo(t, directorio))
			compruebaEsquemaCreado(t, directorio)
		})
	}
}

// compruebaEsquemaCreado exige que en el directorio solo esté world.db, en WAL
// y con las tablas y la versión 1 del esquema.
func compruebaEsquemaCreado(t *testing.T, directorio string) {
	t.Helper()

	ruta := rutaEn(directorio)
	assert.Equal(t, []string{ficheroDelGrafo}, nombresEn(t, directorio), "ningún auxiliar queda")
	assert.Equal(t, []byte{2, 2}, leerFichero(t, ruta)[18:20], "world.db está en WAL")

	tablas, versiones := esquemaDe(t, ruta)
	for _, tabla := range tablasDelGrafo {
		assert.Contains(t, tablas, tabla)
	}

	assert.Equal(t, []int64{1}, versiones)
}

// recuentoDelBloque es lo que graph stats cuenta tras la consulta del bloque.
func recuentoDelBloque() grafo.Recuento {
	return grafo.Recuento{
		Nodos: 3, Aristas: 2, Textos: 1,
		NodosPorTipo: []grafo.RecuentoDeNodos{
			{Tipo: grafo.TipoBloque, Fuente: fuenteDelBOE, Nodos: 1},
			{Tipo: grafo.TipoBloqueVersion, Fuente: fuenteDelBOE, Nodos: 1},
			{Tipo: grafo.TipoNorma, Fuente: fuenteDelBOE, Nodos: 1},
		},
		AristasPorRelacion: []grafo.RecuentoDeAristas{
			{Relacion: grafo.RelacionTieneParte, Fuente: fuenteDelBOE, Aristas: 1},
			{Relacion: grafo.RelacionTieneVersion, Fuente: fuenteDelBOE, Aristas: 1},
		},
	}
}

// fichaDeLaNorma es lo que graph show da de la norma tras la consulta del
// bloque.
func fichaDeLaNorma() grafo.Ficha {
	observacion := grafo.Procedencia{Fuente: fuenteDelBOE, URL: urlDelBloque, FechaConsulta: fechaAntigua}

	return grafo.Ficha{
		Nodo: grafo.NodoDeFicha{
			ID: idNorma, Tipo: grafo.TipoNorma, Datos: map[string]any{grafo.DatoIdentificador: "BOE-A-2015-10565"},
			PrimeraObservacion: fechaAntigua, UltimaObservacion: observacion,
		},
		Salientes: []grafo.AristaDeFicha{{
			Relacion: grafo.RelacionTieneParte, ID: idBloque,
			PrimeraObservacion: fechaAntigua, UltimaObservacion: observacion,
		}},
		Entrantes: []grafo.AristaDeFicha{},
	}
}

// compruebaLecturaSinCambios lee el grafo del directorio con los tres verbos,
// exige lo esperado y que ningún fichero del directorio aparezca, cambie ni
// desaparezca.
func compruebaLecturaSinCambios(t *testing.T, directorio string, esperado leido) {
	t.Helper()

	antes := estadoDelArbol(t, directorio)
	assert.Equal(t, esperado, leerElGrafo(t, directorio))
	assert.Equal(t, antes, estadoDelArbol(t, directorio), "leer no crea, no cambia ni retira nada")
}

// TestIntegracionIdempotencia fija FR-022 y SC-001: entregar el mismo lote 2 y 10
// veces deja un nodo por id, una arista por terna y un texto por huella, y lo
// que se lee —el recuento de graph stats incluido— es lo mismo tras la primera
// entrega que tras la última; como una observación idéntica no cambia nada,
// ninguna entrega repetida cambia ni crea ningún fichero.
func TestIntegracionIdempotencia(t *testing.T) {
	t.Parallel()

	for _, veces := range []int{2, 10} {
		t.Run(strconv.Itoa(veces)+" veces", func(t *testing.T) {
			t.Parallel()

			directorio := t.TempDir()
			grafoEntregado(t, directorio, elBloque())

			trasLaPrimera := estadoDelArbol(t, directorio)
			leidoTrasLaPrimera := leerElGrafo(t, directorio)
			assert.Equal(t, recuentoDelBloque(), leidoTrasLaPrimera.recuento)

			for range veces - 1 {
				entregar(t, directorio, elBloque())
				assert.Equal(t, trasLaPrimera, estadoDelArbol(t, directorio), "la entrega repetida no escribe nada")
			}

			assert.Equal(t, leidoTrasLaPrimera, leerElGrafo(t, directorio))
		})
	}
}

// TestIntegracionObservaciones fija FR-023 por la API: dos observaciones de las
// mismas claves dejan lo mismo lleguen en el orden que lleguen —fuera de orden,
// la primera de la antigua y la última de la reciente; con otros datos, los de
// la última, enteros; y en un empate de instante, la observación ganadora entera,
// como si hubiera llegado sola—.
func TestIntegracionObservaciones(t *testing.T) {
	t.Parallel()

	reobservada := loteDelBOE(urlDeLaNorma, fechaReciente, laNorma(map[string]any{grafo.DatoIdentificador: "BOE-A-2015-10565"}))
	reobservada.Vigencia = 0

	sinVigencia := loteDelBOE(urlDeLaNorma, fechaAntigua, laNorma(nil))
	sinVigencia.Vigencia = 0

	casos := []struct {
		nombre           string
		primero, segundo core.Lote
		// comprueba mira lo que se lee, que es lo mismo en los dos órdenes.
		comprueba func(t *testing.T, l leido)
	}{
		{
			nombre:    "fuera de orden: la primera de la antigua y la última de la reciente",
			primero:   loteDelBloque(urlDelBloque, fechaAntigua),
			segundo:   loteDelBloque(urlDelBloque, fechaReciente),
			comprueba: compruebaPrimeraYUltima,
		},
		{
			nombre: "una reobservación con otros datos y sin vigencia",
			primero: loteDelBOE(urlDeLaNorma, fechaAntigua,
				laNorma(map[string]any{grafo.DatoIdentificador: "BOE-A-2015-10565", "anterior": true})),
			segundo:   reobservada,
			comprueba: compruebaReobservada,
		},
		{
			nombre:    "empate: la url menor, con la fecha escrita de otra forma",
			primero:   loteDelBloque(urlDeLaNorma, fechaEnMadrid),
			segundo:   loteDelBloque(urlDelBloque, fechaAntigua),
			comprueba: comoSiLlegaraSolo(loteDelBloque(urlDeLaNorma, fechaEnMadrid)),
		},
		{
			nombre:    "empate: sin vigencia antes que con ella",
			primero:   sinVigencia,
			segundo:   loteDelBOE(urlDeLaNorma, fechaAntigua, laNorma(nil)),
			comprueba: comoSiLlegaraSolo(sinVigencia),
		},
		{
			nombre: "empate: una reobservación con otros datos, los datos canónicos menores",
			primero: loteDelBOE(urlDeLaNorma, fechaAntigua,
				laNorma(map[string]any{grafo.DatoIdentificador: "A"})),
			segundo: loteDelBOE(urlDeLaNorma, fechaAntigua,
				laNorma(map[string]any{grafo.DatoIdentificador: "B"})),
			comprueba: comoSiLlegaraSolo(loteDelBOE(urlDeLaNorma, fechaAntigua,
				laNorma(map[string]any{grafo.DatoIdentificador: "A"}))),
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			enOrden := leidoTras(t, caso.primero, caso.segundo)
			assert.Equal(t, enOrden, leidoTras(t, caso.segundo, caso.primero), "el orden de llegada no cambia nada")
			caso.comprueba(t, enOrden)
		})
	}
}

// compruebaPrimeraYUltima exige que cada nodo y cada arista tengan la primera
// observación en la fecha antigua y la última en la reciente.
func compruebaPrimeraYUltima(t *testing.T, l leido) {
	t.Helper()

	require.Len(t, l.fichas, 3)

	for id, ficha := range l.fichas {
		assert.Equal(t, fechaAntigua, ficha.Nodo.PrimeraObservacion, id)
		assert.Equal(t, fechaReciente, ficha.Nodo.UltimaObservacion.FechaConsulta, id)

		for _, arista := range ficha.Salientes {
			assert.Equal(t, fechaAntigua, arista.PrimeraObservacion, id+" "+arista.Relacion)
			assert.Equal(t, fechaReciente, arista.UltimaObservacion.FechaConsulta, id+" "+arista.Relacion)
		}
	}
}

// compruebaReobservada exige que la norma tenga los datos y la vigencia de la
// última observación, enteros y sin mezclar, con la primera de la antigua.
func compruebaReobservada(t *testing.T, l leido) {
	t.Helper()

	ficha := l.fichas[idNorma]
	assert.Equal(t, map[string]any{grafo.DatoIdentificador: "BOE-A-2015-10565"}, ficha.Nodo.Datos)
	assert.Equal(t, fechaAntigua, ficha.Nodo.PrimeraObservacion)
	assert.Equal(t, fechaReciente, ficha.Nodo.UltimaObservacion.FechaConsulta)

	require.Len(t, l.instantanea.Nodos, 1)
	assert.Zero(t, l.instantanea.Nodos[0].Vigencia, "la última no declara vigencia")
}

// comoSiLlegaraSolo exige que lo leído sea lo que deja el lote llegando solo.
func comoSiLlegaraSolo(lote core.Lote) func(t *testing.T, l leido) {
	return func(t *testing.T, l leido) {
		t.Helper()

		assert.Equal(t, leidoTras(t, lote), l, "queda la observación ganadora entera")
	}
}

// TestIntegracionRechazos fija FR-024 y FR-025 por la API: un lote sin fuente,
// con un id que cambia de tipo en el grafo o en el propio lote, con un texto
// cuya huella no es la de su cuerpo o ya está guardada con otro cuerpo, con un
// extremo que no está ni en el lote ni en el grafo o con una Persona que lleva
// un documento se rechaza entero, con «inesperado» y un mensaje que nombra
// world.db y el motivo, y el grafo queda intacto: lo mismo se lee y ningún
// fichero cambia ni aparece.
func TestIntegracionRechazos(t *testing.T) {
	t.Parallel()

	sinFuente := elBloque()
	sinFuente.Fuente = ""

	municipio := schema.Nodo{ID: idMunicipio, Tipo: grafo.TipoMunicipio, Datos: map[string]any{grafo.DatoCodigoINE: "28074"}}
	persona := func(id string, datos map[string]any) schema.Nodo {
		return schema.Nodo{ID: id, Tipo: grafo.TipoPersona, Datos: datos}
	}

	casos := []struct {
		nombre string
		// preparar cambia lo guardado antes de entregar; nil si no hace falta.
		preparar func(t *testing.T, ruta string)
		lote     core.Lote
		// sinTocarElDisco dice que el rechazo llega antes de resolver la ruta,
		// así que el mensaje nombra world.db sin ella.
		sinTocarElDisco bool
	}{
		{nombre: "sin fuente", lote: sinFuente, sinTocarElDisco: true},
		{
			nombre: "un id con otro tipo que el guardado",
			lote:   loteDelBOE(urlDeLaNorma, fechaReciente, municipio, schema.Nodo{ID: idNorma, Tipo: grafo.TipoMunicipio}),
		},
		{
			nombre: "un id con dos tipos en el lote",
			lote: loteDelBOE(urlDeLaNorma, fechaReciente, municipio,
				schema.Nodo{ID: idMunicipio, Tipo: grafo.TipoOrgano}),
			sinTocarElDisco: true,
		},
		{
			nombre: "un texto cuya huella no es la de su cuerpo",
			lote: loteDelBOE(urlDelBloque, fechaReciente, municipio,
				schema.Texto{Huella: huellaDe(cuerpo2016), Cuerpo: cuerpo2025}),
			sinTocarElDisco: true,
		},
		{
			nombre:   "una huella ya guardada con otro cuerpo",
			preparar: conLaHuellaDe2025GuardadaConOtroCuerpo,
			lote: loteDelBOE(urlDelBloque, fechaReciente, municipio,
				schema.Texto{Huella: huellaDe(cuerpo2025), Cuerpo: cuerpo2025}),
		},
		{
			nombre: "un extremo que no está ni en el lote ni en el grafo",
			lote: loteDelBOE(urlDeLaNorma, fechaReciente, municipio,
				schema.Arista{Origen: idMunicipio, Relacion: grafo.RelacionPerteneceA, Destino: idAusente}),
		},
		{
			nombre:          "una Persona con un DNI en su id",
			lote:            loteDelBOE(urlDeLaNorma, fechaReciente, municipio, persona("12345678Z", nil)),
			sinTocarElDisco: true,
		},
		{
			nombre: "una Persona con un NIE dentro de una lista anidada",
			lote: loteDelBOE(urlDeLaNorma, fechaReciente, municipio, persona("ana-garcia-lopez",
				map[string]any{"contacto": map[string]any{"documentos": []any{"X-1234567-L"}}})),
			sinTocarElDisco: true,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			directorio := t.TempDir()
			ruta := grafoEntregado(t, directorio, elBloque())

			if caso.preparar != nil {
				caso.preparar(t, ruta)
			}

			antes := estadoDelArbol(t, directorio)
			leidoAntes := leerElGrafo(t, directorio)

			err := almacenEn(directorio).Apply(t.Context(), caso.lote)

			var rechazo *grafo.Rechazo

			require.ErrorAs(t, err, &rechazo)

			nombre := strconv.Quote(ruta)
			if caso.sinTocarElDisco {
				nombre = ficheroDelGrafo
			}

			compruebaFallo(t, err, operacionEscribir, schema.ClaseInesperado,
				"grafo: el lote no entra en "+nombre+": "+rechazo.Error())
			assert.NotContains(t, err.Error(), "12345678Z")
			assert.NotContains(t, err.Error(), cuerpo2025)
			assert.Equal(t, antes, estadoDelArbol(t, directorio), "ningún fichero cambia ni aparece")
			assert.Equal(t, leidoAntes, leerElGrafo(t, directorio), "no entra nada del lote")
		})
	}
}

// conLaHuellaDe2025GuardadaConOtroCuerpo guarda, con otra conexión, un texto
// con la huella del cuerpo de 2025 y otro cuerpo: lo que ninguna entrega
// escribe, pero otro programa sí.
func conLaHuellaDe2025GuardadaConOtroCuerpo(t *testing.T, ruta string) {
	t.Helper()

	ejecutar(t, ruta, `INSERT INTO texts VALUES ('`+huellaDe(cuerpo2025)+`', 'otro cuerpo', '`+fechaAntigua+
		`', '`+fuenteDelBOE+`', '`+urlDelBloque+`')`)
	compruebaSinAuxiliares(t, ruta)
}

// TestIntegracionConcurrencia fija SC-009 y FR-014 por la API: ocho almacenes
// entregan a la vez sobre un directorio que no existe, cada uno la consulta de
// un órgano distinto de Leganés con su propia fecha. Todos crean world.db en su
// sitio si no está y aplican en él: las ocho entregas salen bien, el esquema se
// crea una sola vez, no queda ningún auxiliar y las ocho observaciones quedan,
// también las ocho del municipio que comparten (la primera, la de la fecha
// menor; la última, la de la mayor).
func TestIntegracionConcurrencia(t *testing.T) {
	t.Parallel()

	const invocaciones = 8

	raiz := t.TempDir()
	directorio := filepath.Join(raiz, "no-existe", "cache")

	lotes := make([]core.Lote, invocaciones)
	entrantes := make([]grafo.AristaDeFicha, invocaciones)

	for i := range invocaciones {
		organo, fecha := "L0128074"+strconv.Itoa(i), fmt.Sprintf("2026-02-0%dT00:00:00Z", i+1)
		lotes[i] = loteDelTerritorio(organo, fecha)
		entrantes[i] = grafo.AristaDeFicha{
			Relacion: grafo.RelacionPerteneceA, ID: organo, PrimeraObservacion: fecha,
			UltimaObservacion: grafo.Procedencia{Fuente: fuenteDelTerritorio, URL: urlDelTerritorio, FechaConsulta: fecha},
		}
	}

	for i, err := range entregarALaVez(t, directorio, lotes) {
		require.NoError(t, err, "la entrega %d", i)
	}

	assert.Equal(t, []string{"cache"}, nombresEn(t, filepath.Join(raiz, "no-existe")))
	assert.Equal(t, []string{ficheroDelGrafo}, nombresEn(t, directorio), "ningún auxiliar queda")

	_, versiones := esquemaDe(t, rutaEn(directorio))
	assert.Equal(t, []int64{1}, versiones, "el esquema se creó una sola vez")

	l := leerElGrafo(t, directorio)
	assert.Equal(t, invocaciones+1, l.recuento.Nodos)
	assert.Equal(t, invocaciones, l.recuento.Aristas)

	delMunicipio := l.fichas[idMunicipio]
	assert.Equal(t, "2026-02-01T00:00:00Z", delMunicipio.Nodo.PrimeraObservacion)
	assert.Equal(t, "2026-02-08T00:00:00Z", delMunicipio.Nodo.UltimaObservacion.FechaConsulta)
	assert.Equal(t, entrantes, delMunicipio.Entrantes, "la arista de cada órgano, con su observación")
}

// entregarALaVez entrega cada lote desde su propio almacén y su propia
// goroutine, todas soltadas a la vez, y devuelve el fallo de cada una.
func entregarALaVez(t *testing.T, directorio string, lotes []core.Lote) []error {
	t.Helper()

	var (
		salida  = make(chan struct{})
		grupo   sync.WaitGroup
		fallos  = make([]error, len(lotes))
		ctx     = t.Context()
		almacen = make([]*graph.Almacen, len(lotes))
	)

	for i, lote := range lotes {
		almacen[i] = almacenEn(directorio)

		grupo.Go(func() {
			<-salida

			fallos[i] = almacen[i].Apply(ctx, lote)
		})
	}

	close(salida)
	grupo.Wait()

	return fallos
}

// TestIntegracionInutilizables fija H7.1 FR-070 y SC-011, y H7 FR-012, por la
// API: un world.db que no es una base de datos —el vehículo de la regla
// genérica— o con un esquema posterior hace fallar la lectura y la entrega con
// «inesperado» (código 1) y su mensaje, que nombra la ruta y, en la regla
// genérica, la causa. De los bytes de lo que no es una base no se promete nada.
func TestIntegracionInutilizables(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string
		// preparar deja world.db en el directorio.
		preparar func(t *testing.T, ruta string)
		// mensaje es el esperado al leer y al entregar.
		mensaje func(ruta string, fallo *graph.Error) string
	}{
		{
			nombre: "lo que no es una base de datos",
			preparar: func(t *testing.T, ruta string) {
				t.Helper()

				escribirFichero(t, ruta, []byte("no soy una base de datos\n"))
			},
			mensaje: mensajeInutilizable,
		},
		{
			nombre:   "un esquema posterior",
			preparar: conUnEsquemaPosterior,
			mensaje:  func(ruta string, _ *graph.Error) string { return mensajeDeVersionPosterior(ruta) },
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			directorio := t.TempDir()
			ruta := rutaEn(directorio)
			caso.preparar(t, ruta)

			lectura, err := graph.Leer(t.Context(), graph.ConDirectorio(directorio))
			assert.Nil(t, lectura)
			compruebaFallo(t, err, operacionLeer, schema.ClaseInesperado, caso.mensaje(ruta, falloDe(t, err)))

			err = almacenEn(directorio).Apply(t.Context(), loteDelBloque(urlDelBloque, fechaReciente))
			compruebaFallo(t, err, operacionEscribir, schema.ClaseInesperado, caso.mensaje(ruta, falloDe(t, err)))
		})
	}
}

// conUnEsquemaPosterior deja el world.db del bloque con la versión 2 del
// esquema registrada, la de un binario posterior.
func conUnEsquemaPosterior(t *testing.T, ruta string) {
	t.Helper()

	grafoEntregado(t, filepath.Dir(ruta), elBloque())
	ejecutar(t, ruta, `INSERT INTO schema_version VALUES (2, '2030-01-01T00:00:00Z')`)
	compruebaSinAuxiliares(t, ruta)
}

// TestIntegracionPlazoYBloqueo fija FR-014 por la API: mientras otra invocación
// retiene world.db —en exclusiva para la lectura (H7.1 research.md D22) y con
// su escritura para la entrega—, la lectura y la entrega esperan por tramos y
// salen, si el plazo de quien llama termina antes, con «fuente-no-disponible»
// (código 4) sin agotar la espera propia y, si no, al agotarla, con
// «inesperado» (código 1) «bloqueada por otra invocación». Con el contexto ya
// terminado, ninguna de las dos toca nada. En todos los casos el grafo queda
// como estaba.
func TestIntegracionPlazoYBloqueo(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string
		// plazo es el de quien lee o entrega; cero, sin plazo.
		plazo   time.Duration
		clase   schema.Clase
		mensaje func(operacion, nombre string) string
	}{
		{nombre: "el plazo agotado", plazo: plazoCorto, clase: schema.ClaseFuenteNoDisponible, mensaje: mensajeDePlazo},
		{nombre: "la espera propia agotada", clase: schema.ClaseInesperado, mensaje: mensajeDeBloqueo},
	}

	for _, caso := range casos {
		t.Run(caso.nombre+" al leer", func(t *testing.T) {
			t.Parallel()

			directorio := t.TempDir()
			ruta := grafoEntregado(t, directorio, elBloque())
			antes := estadoDelArbol(t, directorio)
			suelta := retener(t, ruta, enExclusiva)

			inicio := time.Now()
			lectura, err := graph.Leer(contextoCon(t, caso.plazo), graph.ConDirectorio(directorio))
			tardo := time.Since(inicio)

			suelta()
			assert.Nil(t, lectura)
			compruebaFallo(t, err, operacionLeer, caso.clase, caso.mensaje(operacionLeer, strconv.Quote(ruta)))
			compruebaLaEspera(t, caso.plazo, tardo)
			assert.Equal(t, antes, estadoDelArbol(t, directorio))
		})

		t.Run(caso.nombre+" al entregar", func(t *testing.T) {
			t.Parallel()

			directorio := t.TempDir()
			ruta := grafoEntregado(t, directorio, elBloque())
			antes, leidoAntes := estadoDelArbol(t, directorio), leerElGrafo(t, directorio)
			suelta := retener(t, ruta, reteniendoLaEscritura)

			inicio := time.Now()
			err := almacenEn(directorio).Apply(contextoCon(t, caso.plazo), loteDelBloque(urlDelBloque, fechaReciente))
			tardo := time.Since(inicio)

			suelta()
			compruebaFallo(t, err, operacionEscribir, caso.clase, caso.mensaje(operacionEscribir, strconv.Quote(ruta)))
			compruebaLaEspera(t, caso.plazo, tardo)
			assert.Equal(t, antes, estadoDelArbol(t, directorio))
			assert.Equal(t, leidoAntes, leerElGrafo(t, directorio))
		})
	}

	t.Run("el contexto ya terminado: ni la lectura ni la entrega tocan nada", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		ruta := grafoEntregado(t, directorio, elBloque())
		antes := estadoDelArbol(t, directorio)

		lectura, err := graph.Leer(contextoCon(t, plazoTerminado), graph.ConDirectorio(directorio))
		assert.Nil(t, lectura)
		compruebaFallo(t, err, operacionLeer, schema.ClaseFuenteNoDisponible, mensajeDePlazo(operacionLeer, strconv.Quote(ruta)))
		require.ErrorIs(t, err, context.Canceled)

		err = almacenEn(directorio).Apply(contextoCon(t, plazoTerminado), loteDelBloque(urlDelBloque, fechaReciente))
		compruebaFallo(t, err, operacionEscribir, schema.ClaseFuenteNoDisponible,
			mensajeDePlazo(operacionEscribir, ficheroDelGrafo))
		require.ErrorIs(t, err, context.Canceled)

		assert.Equal(t, antes, estadoDelArbol(t, directorio))
	})
}

// compruebaLaEspera exige que, con plazo, la operación saliera antes de agotar
// la espera propia y, sin él, después.
func compruebaLaEspera(t *testing.T, plazo, tardo time.Duration) {
	t.Helper()

	if plazo == 0 {
		assert.GreaterOrEqual(t, tardo, esperaPropia, "se espera la espera propia entera")

		return
	}

	assert.Less(t, tardo, esperaPropia, "el plazo llega antes que la espera propia")
}

// TestIntegracionSinResiduo fija por la API lo que deja una entrega que falla
// antes de crear nada (contracts/almacen-world-db.md §4, pasos 1 y 2; H7 FR-011,
// FR-033): sobre un directorio de caché que no existe, con el contexto ya
// terminado o bajo un antecesor en el que no se puede crear el directorio, no
// deja ni world.db ni el directorio.
func TestIntegracionSinResiduo(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string
		// preparar deja la raíz y devuelve el directorio de world.db, que no
		// existe; nil, un directorio que no existe bajo otro que tampoco.
		preparar func(t *testing.T, raiz string) string
		plazo    time.Duration
		clase    schema.Clase
		mensaje  func(ruta string, fallo *graph.Error) string
	}{
		{
			nombre: "el contexto ya terminado",
			plazo:  plazoTerminado,
			clase:  schema.ClaseFuenteNoDisponible,
			mensaje: func(string, *graph.Error) string {
				return mensajeDePlazo(operacionEscribir, ficheroDelGrafo)
			},
		},
		{
			nombre:   "un antecesor en el que no se puede escribir",
			preparar: bajoUnDirectorioSinEscritura,
			clase:    schema.ClaseInesperado,
			mensaje: func(ruta string, fallo *graph.Error) string {
				return "grafo: no se puede escribir world.db en " + strconv.Quote(filepath.Dir(ruta)) + ": " + fallo.Causa.Error()
			},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			raiz := t.TempDir()
			directorio := filepath.Join(raiz, "no-existe", "cache")

			if caso.preparar != nil {
				directorio = caso.preparar(t, raiz)
			}

			antes := estadoDelArbol(t, raiz)

			err := almacenEn(directorio).Apply(contextoCon(t, caso.plazo), elBloque())
			compruebaFallo(t, err, operacionEscribir, caso.clase, caso.mensaje(rutaEn(directorio), falloDe(t, err)))

			assert.Equal(t, antes, estadoDelArbol(t, raiz), "no queda nada")
			assert.NoDirExists(t, directorio)
		})
	}
}

// bajoUnDirectorioSinEscritura es un directorio de caché que no existe bajo otro
// que tampoco, dentro de uno que existe y no admite entradas nuevas.
func bajoUnDirectorioSinEscritura(t *testing.T, raiz string) string {
	t.Helper()

	bloqueado := filepath.Join(raiz, "bloqueado")
	require.NoError(t, os.Mkdir(bloqueado, permisosDeDirectorio))
	directorioSinEscritura(t, bloqueado)

	return filepath.Join(bloqueado, "no-existe", "cache")
}

// TestIntegracionLecturaConWAL fija la lectura con auxiliares de WAL
// (contracts/almacen-world-db.md §3, paso 2; H7.1 FR-077; research.md V36 D y F
// de H7): con un -wal huérfano —la copia de los ficheros de un escritor
// abierto, sin otra conexión— y con un escritor abierto de verdad, los tres
// verbos ven lo confirmado en el WAL, world.db y el -wal quedan con los mismos
// bytes y, al terminar, world.db-shm existe: la desviación declarada de FR-004,
// FR-031 y SC-004, y nada más.
func TestIntegracionLecturaConWAL(t *testing.T) {
	t.Parallel()

	t.Run("un -wal huérfano", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		copiaConElWAL(t, rutaEn(directorio), elTerritorioEnElWAL, "", sufijoWAL, sufijoMemoriaCompartida)
		compruebaLecturaConWAL(t, directorio)
	})

	t.Run("un escritor abierto", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		ruta := grafoEntregado(t, directorio, elBloque())
		otra := otraInvocacionAbierta(t, ruta)
		entregar(t, directorio, elTerritorio())

		tx, err := otra.BeginTx(context.WithoutCancel(t.Context()), nil)
		require.NoError(t, err, "la otra invocación empieza a escribir y sigue abierta")
		t.Cleanup(func() { assert.NoError(t, tx.Rollback()) })

		compruebaLecturaConWAL(t, directorio)
	})
}

// compruebaLecturaConWAL lee con los tres verbos el world.db del directorio, con
// el bloque en la base y el territorio confirmado en el WAL.
func compruebaLecturaConWAL(t *testing.T, directorio string) {
	t.Helper()

	ruta := rutaEn(directorio)
	require.Positive(t, tamano(t, ruta+sufijoWAL), "premisa: el WAL tiene marcos")

	antes := huellasDe(t, ruta, ruta+sufijoWAL)

	assert.Equal(t, leidoTras(t, elBloque(), elTerritorio()), leerElGrafo(t, directorio), "lo confirmado en el WAL")
	assert.Equal(t, antes, huellasDe(t, ruta, ruta+sufijoWAL), "world.db y su -wal no cambian")
	assert.Equal(t, []string{ficheroDelGrafo, ficheroDelGrafo + sufijoMemoriaCompartida, ficheroDelGrafo + sufijoWAL},
		nombresEn(t, directorio), "el -shm existe al terminar, y nada más aparece")
}

// TestIntegracionRecuperacionDeclarada fija la fila de la recuperación de
// contracts/almacen-world-db.md §4.1 de H7 (research.md V44 de H7): la conexión de una
// entrega que falla recupera lo que dejó un escritor interrumpido, como
// cualquier escritor de SQLite. Con un -wal huérfano, el checkpoint del cierre
// lleva a world.db lo confirmado en él y retira el -wal y el -shm, tanto si el
// lote se rechaza contra lo guardado como si lo confirmado es un esquema
// posterior.
func TestIntegracionRecuperacionDeclarada(t *testing.T) {
	t.Parallel()

	t.Run("un -wal huérfano y un lote que se rechaza contra lo guardado", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		ruta := rutaEn(directorio)
		copiaConElWAL(t, ruta, elTerritorioEnElWAL, "", sufijoWAL, sufijoMemoriaCompartida)
		antes := huellaDelFichero(t, ruta)

		err := almacenEn(directorio).Apply(t.Context(),
			loteDelBOE(urlDeLaNorma, fechaReciente, schema.Nodo{ID: idNorma, Tipo: grafo.TipoMunicipio}))
		compruebaFallo(t, err, operacionEscribir, schema.ClaseInesperado, mensajeDelLoteRechazado(ruta, falloDe(t, err)))

		compruebaRecuperado(t, ruta, antes)
		assert.Equal(t, leidoTras(t, elBloque(), elTerritorio()), leerElGrafo(t, directorio),
			"el grafo tiene lo confirmado, ahora en world.db")
	})

	t.Run("un -wal huérfano con un esquema posterior confirmado en él", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		ruta := rutaEn(directorio)
		copiaConElWAL(t, ruta, unEsquemaPosteriorEnElWAL, "", sufijoWAL, sufijoMemoriaCompartida)
		antes := huellaDelFichero(t, ruta)

		err := almacenEn(directorio).Apply(t.Context(), loteDelBloque(urlDelBloque, fechaReciente))
		compruebaFallo(t, err, operacionEscribir, schema.ClaseInesperado, mensajeDeVersionPosterior(ruta))

		compruebaRecuperado(t, ruta, antes)

		_, versiones := esquemaDe(t, ruta)
		assert.Equal(t, []int64{1, 2}, versiones, "la versión 2 está ahora en world.db")
	})
}

// compruebaRecuperado exige que el checkpoint del cierre haya llevado el WAL a
// world.db —cambian sus bytes— y retirado el -wal y el -shm.
func compruebaRecuperado(t *testing.T, ruta, antes string) {
	t.Helper()

	assert.NoFileExists(t, ruta+sufijoWAL)
	assert.NoFileExists(t, ruta+sufijoMemoriaCompartida)
	assert.NotEqual(t, antes, huellaDelFichero(t, ruta), "lo confirmado en el WAL está ahora en world.db")
}
