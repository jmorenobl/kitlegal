package app_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cache"
	"github.com/jmorenobl/kitlegal/internal/core"
	"github.com/jmorenobl/kitlegal/internal/core/grafo"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/graph"
	"github.com/jmorenobl/kitlegal/internal/source/boe"
)

// Las cotas de SC-007 y SC-008 y cuántas veces se mide cada una.
const (
	// maximoDeLaEntrega es lo que, como mucho, puede añadir la entrega al grafo
	// a la mediana de boe articulo servido desde la caché (SC-007).
	maximoDeLaEntrega       = 150 * time.Millisecond
	invocacionesDeLaCache   = 20
	maximoDeCheck           = 3 * time.Second
	maximoDeStats           = time.Second
	invocacionesDelGrafo    = 5
	nodosDelGrafoGrande     = 10_000
	aristasDelGrafoGrande   = 10_000
	archivoDelGrafoDelMundo = "world.db"
)

// El grafo grande de SC-008 tiene la forma de lo que emiten boe y territorio
// (contracts/emision.md): normas con sus bloques y las versiones de cada
// bloque, con el texto de cada versión por su huella, y municipios con órganos
// que les pertenecen. Todo es inventado, como en los tests del applet, para que
// nada pueda tomarse por un dato real.
//
// Un bosque de normas, bloques y versiones tiene una arista menos por norma que
// nodos; las que faltan hasta que aristas y nodos sean los mismos las ponen
// órganos que pertenecen a dos municipios, como una mancomunidad. De ahí salen
// cuántos órganos y municipios hay (organosDelGrafoGrande y
// municipiosDelGrafoGrande).
//
// Las normas se consultaron hace meses con la vigencia de una semana de boe, y
// sus bloques tienen cuatro versiones: check tiene trabajo de las dos clases,
// una fuente caducada por nodo de norma y una versión obsoleta por cada versión
// que no es la última de su bloque.
const (
	normasDelGrafoGrande = 70
	bloquesPorNorma      = 25
	versionesPorBloque   = 4

	fuenteDelGrafoGrande    = "prueba.legislacion"
	fuenteTerritorialGrande = "kitlegal.prueba"
	urlTerritorialGrande    = "kitlegal:applet/prueba"
	fechaDelGrafoGrande     = "2026-01-15T10:00:00Z"
	vigenciaDelGrafoGrande  = 7 * 24 * time.Hour
	primerAnoDeVigencia     = 2016
	primerCodigoINEDePrueba = 90_000
	primerDIR3DePrueba      = 9_000_000
)

// Lo que resulta de esa forma: los nodos y las aristas de las normas, y los
// órganos y municipios que completan el grafo.
const (
	bloquesDelGrafoGrande   = normasDelGrafoGrande * bloquesPorNorma
	versionesDelGrafoGrande = bloquesDelGrafoGrande * versionesPorBloque
	nodosDeLasNormas        = normasDelGrafoGrande + bloquesDelGrafoGrande + versionesDelGrafoGrande
	aristasDeLasNormas      = bloquesDelGrafoGrande + versionesDelGrafoGrande

	organosDelGrafoGrande    = (aristasDelGrafoGrande - aristasDeLasNormas) / 2
	municipiosDelGrafoGrande = nodosDelGrafoGrande - nodosDeLasNormas - organosDelGrafoGrande
)

// TestCosteDelGrafo mide lo que cuesta el grafo del mundo con el binario de
// e2e y el reloj de pared: SC-007, lo que la entrega añade a boe articulo
// servido desde la caché con un world.db que ya existe, y SC-008, lo que tardan
// graph check y graph stats sobre un grafo de 10 000 nodos y 10 000 aristas
// (research.md D28 de H7). Publica las medianas medidas, que es lo que dice lo
// cerca que quedan de su cota en cada máquina (supuestos S1 y S2 del plan).
//
// Como TestMedidasDeTiempo, no es paralela y make ci la ejecuta en su propio
// paso (test-tiempos), después de test y test-integration, que la saltan: una
// medida de reloj con todo el módulo en marcha mide la carga, no el programa.
//
//nolint:paralleltest // Mide con el reloj de pared, como TestMedidasDeTiempo: con otra prueba a la vez, mediría también lo suyo.
func TestCosteDelGrafo(t *testing.T) {
	require.NoError(t, entorno.err)

	t.Run("la-mediana", compruebaLaMediana)
	t.Run("SC-007-entrega-desde-la-cache", compruebaElCosteDeLaEntrega)
	t.Run("SC-008-check-y-stats-sobre-el-grafo-grande", compruebaElCosteDelGrafoGrande)
}

// compruebaLaMediana fija la mediana con la que se comparan las cotas: la
// central con un número impar de medidas y la media de las dos centrales con
// uno par, sin depender del orden en que llegan.
func compruebaLaMediana(t *testing.T) {
	t.Helper()

	ms := time.Millisecond

	assert.Equal(t, 3*ms, mediana([]time.Duration{5 * ms, 1 * ms, 3 * ms, 9 * ms, 2 * ms}))
	assert.Equal(t, 4*ms, mediana([]time.Duration{9 * ms, 1 * ms, 5 * ms, 3 * ms}))
}

// compruebaElCosteDeLaEntrega siembra la caché y world.db con una consulta de
// boe articulo, vacía la reproducción —una invocación que pidiera algo no
// encontraría ninguna grabación y fallaría— y mide las mismas invocaciones
// servidas desde la caché con la entrega y con --no-graph, alternando cuál va
// primero para que ninguna de las dos se lleve siempre la máquina más fría. Cada
// una da la salida de la que sembró, sin nada en la salida de error: una
// entrega que fallara avisaría ahí, y lo medido no sería entregar (SC-007).
func compruebaElCosteDeLaEntrega(t *testing.T) {
	t.Helper()

	trabajo := t.TempDir()
	require.NoError(t, copiarGrabaciones(trabajo))

	carpetaDeLaCache := filepath.Join(trabajo, directorioDeLaCache)

	_, sembrada := medirOrden(t, ordenDeBoe(t.Context(), entorno.binario, false), trabajo, carpetaDeLaCache)
	require.FileExists(t, filepath.Join(carpetaDeLaCache, archivoDelGrafoDelMundo),
		"premisa: la entrega de la consulta que siembra la cach\xc3\xa9 deja world.db")

	reproduccion := filepath.Join(trabajo, directorioDeReproduccion)
	require.NoError(t, os.RemoveAll(reproduccion))
	require.NoError(t, os.MkdirAll(filepath.Join(reproduccion, boe.NombreDeLaFuente), 0o700))

	var conGrafo, sinGrafo []time.Duration

	for i := range invocacionesDeLaCache {
		for _, sinEntrega := range []bool{i%2 == 1, i%2 == 0} {
			duracion, salida := medirOrden(t, ordenDeBoe(t.Context(), entorno.binario, sinEntrega), trabajo,
				carpetaDeLaCache)
			require.Equal(t, string(sembrada), string(salida),
				"servida desde la cach\xc3\xa9, la misma salida que la sembrada")

			if sinEntrega {
				sinGrafo = append(sinGrafo, duracion)
			} else {
				conGrafo = append(conGrafo, duracion)
			}
		}
	}

	medianaCon, medianaSin := mediana(conGrafo), mediana(sinGrafo)
	diferencia := medianaCon - medianaSin

	t.Logf("SC-007: mediana de %d boe articulo desde la cach\xc3\xa9: %s con la entrega y %s con --no-graph;"+
		" diferencia %s (m\xc3\xa1ximo %s)", invocacionesDeLaCache, medianaCon, medianaSin, diferencia, maximoDeLaEntrega)

	assert.LessOrEqualf(t, diferencia, maximoDeLaEntrega,
		"SC-007: la entrega a\xc3\xb1ade %s a la mediana (%s con ella, %s sin ella) y el m\xc3\xa1ximo es %s",
		diferencia, medianaCon, medianaSin, maximoDeLaEntrega)
}

// compruebaElCosteDelGrafoGrande construye el grafo grande con la API pública
// de internal/graph y mide graph check y graph stats sobre él, alternándolos.
// Cada respuesta es la esperada —el recuento del grafo entero y los hallazgos
// de las dos clases—, de modo que lo medido es el trabajo de verdad y no, por
// ejemplo, un grafo que no se pudo leer (SC-008).
func compruebaElCosteDelGrafoGrande(t *testing.T) {
	t.Helper()

	directorio := t.TempDir()
	construirElGrafoGrande(t, directorio)

	var deCheck, deStats []time.Duration

	for range invocacionesDelGrafo {
		duracion, salida := medirOrden(t, ordenDelGrafo(t.Context(), entorno.binario, "check"), directorio, directorio)
		compruebaLosHallazgosDelGrafoGrande(t, salida)

		deCheck = append(deCheck, duracion)

		duracion, salida = medirOrden(t, ordenDelGrafo(t.Context(), entorno.binario, "stats"), directorio, directorio)
		compruebaElRecuentoDelGrafoGrande(t, salida)

		deStats = append(deStats, duracion)
	}

	medianaDeCheck, medianaDeStats := mediana(deCheck), mediana(deStats)

	t.Logf("SC-008: mediana de %d sobre %d nodos y %d aristas: graph check %s (m\xc3\xa1ximo %s), graph stats %s"+
		" (m\xc3\xa1ximo %s)", invocacionesDelGrafo, nodosDelGrafoGrande, aristasDelGrafoGrande, medianaDeCheck,
		maximoDeCheck, medianaDeStats, maximoDeStats)

	assert.Lessf(t, medianaDeCheck, maximoDeCheck, "SC-008: graph check tarda %s de mediana y el m\xc3\xa1ximo es %s",
		medianaDeCheck, maximoDeCheck)
	assert.Lessf(t, medianaDeStats, maximoDeStats, "SC-008: graph stats tarda %s de mediana y el m\xc3\xa1ximo es %s",
		medianaDeStats, maximoDeStats)
}

// construirElGrafoGrande entrega al world.db del directorio los lotes del grafo
// grande: uno por norma, como una consulta de boe cada una, y uno con los
// municipios y sus órganos, como territorio.
func construirElGrafoGrande(t *testing.T, directorio string) {
	t.Helper()

	almacen := graph.Nuevo(graph.ConDirectorio(directorio))

	for norma := range normasDelGrafoGrande {
		require.NoError(t, almacen.Apply(t.Context(), loteDeUnaNorma(norma)))
	}

	require.NoError(t, almacen.Apply(t.Context(), loteTerritorialGrande()))
}

// loteDeUnaNorma es el de una norma del grafo grande: la Norma, sus bloques y
// las versiones de cada bloque, con sus aristas y el texto de cada versión.
func loteDeUnaNorma(norma int) core.Lote {
	idDeLaNorma := "eli/prueba/l/2026/" + strconv.Itoa(norma)
	operaciones := []schema.Operacion{schema.Nodo{
		ID: idDeLaNorma, Tipo: grafo.TipoNorma,
		Datos: map[string]any{grafo.DatoIdentificador: "PRUEBA-2026-" + strconv.Itoa(norma)},
	}}

	for bloque := range bloquesPorNorma {
		nombre := "a" + strconv.Itoa(bloque+1)
		idDelBloque := idDeLaNorma + "#" + nombre
		operaciones = append(operaciones,
			schema.Nodo{ID: idDelBloque, Tipo: grafo.TipoBloque, Datos: map[string]any{grafo.DatoBloque: nombre}},
			schema.Arista{Origen: idDeLaNorma, Relacion: grafo.RelacionTieneParte, Destino: idDelBloque},
		)

		for version := range versionesPorBloque {
			cuerpo := fmt.Sprintf("Texto inventado %d del bloque %s.", version, idDelBloque)
			huella := "sha256:" + huellaSHA256([]byte(cuerpo))
			fechaDeVigencia := strconv.Itoa(primerAnoDeVigencia+version) + "0101"
			idDeLaVersion := idDelBloque + "@" + fechaDeVigencia + ":" + huella
			operaciones = append(operaciones,
				schema.Nodo{ID: idDeLaVersion, Tipo: grafo.TipoBloqueVersion, Datos: map[string]any{
					grafo.DatoFechaVigencia: fechaDeVigencia, grafo.DatoHashTexto: huella,
				}},
				schema.Arista{Origen: idDelBloque, Relacion: grafo.RelacionTieneVersion, Destino: idDeLaVersion},
				schema.Texto{Huella: huella, Cuerpo: cuerpo},
			)
		}
	}

	return core.Lote{
		Fuente:        fuenteDelGrafoGrande,
		URL:           "https://legislacion.example/prueba/2026/" + strconv.Itoa(norma),
		FechaConsulta: fechaDelGrafoGrande,
		Vigencia:      vigenciaDelGrafoGrande,
		Operaciones:   operaciones,
	}
}

// loteTerritorialGrande es el de los municipios y los órganos del grafo grande,
// sin vigencia: cada órgano pertenece a dos municipios consecutivos.
func loteTerritorialGrande() core.Lote {
	idDelMunicipio := func(municipio int) string {
		return "ine:" + strconv.Itoa(primerCodigoINEDePrueba+(municipio%municipiosDelGrafoGrande))
	}

	var operaciones []schema.Operacion

	for municipio := range municipiosDelGrafoGrande {
		codigo := strconv.Itoa(primerCodigoINEDePrueba + municipio)
		operaciones = append(operaciones, schema.Nodo{ID: "ine:" + codigo, Tipo: grafo.TipoMunicipio, Datos: map[string]any{
			grafo.DatoCodigoINE: codigo, grafo.DatoNombre: "Municipio de prueba " + codigo,
		}})
	}

	for organo := range organosDelGrafoGrande {
		dir3 := "L" + strconv.Itoa(primerDIR3DePrueba+organo)
		operaciones = append(operaciones,
			schema.Nodo{ID: dir3, Tipo: grafo.TipoOrgano, Datos: map[string]any{grafo.DatoDIR3: dir3}},
			schema.Arista{Origen: dir3, Relacion: grafo.RelacionPerteneceA, Destino: idDelMunicipio(organo)},
			schema.Arista{Origen: dir3, Relacion: grafo.RelacionPerteneceA, Destino: idDelMunicipio(organo + 1)},
		)
	}

	return core.Lote{
		Fuente: fuenteTerritorialGrande, URL: urlTerritorialGrande, FechaConsulta: fechaDelGrafoGrande,
		Operaciones: operaciones,
	}
}

// compruebaElRecuentoDelGrafoGrande exige que graph stats cuente el grafo
// grande entero: sus nodos, sus aristas y un texto por versión.
func compruebaElRecuentoDelGrafoGrande(t *testing.T, salida []byte) {
	t.Helper()

	var recuento grafo.Recuento

	require.NoError(t, json.Unmarshal(datosDelSobreDeExito(t, salida), &recuento))
	assert.Equal(t, nodosDelGrafoGrande, recuento.Nodos)
	assert.Equal(t, aristasDelGrafoGrande, recuento.Aristas)
	assert.Equal(t, versionesDelGrafoGrande, recuento.Textos)
}

// compruebaLosHallazgosDelGrafoGrande exige que graph check encuentre en el
// grafo grande lo que tiene: una fuente caducada por cada nodo de las normas,
// que se consultaron con vigencia hace meses, y ninguna de los municipios ni
// de los órganos, que no la declaran; y una versión obsoleta por cada versión
// que no es la última de su bloque.
func compruebaLosHallazgosDelGrafoGrande(t *testing.T, salida []byte) {
	t.Helper()

	var hallazgos []grafo.Hallazgo

	require.NoError(t, json.Unmarshal(datosDelSobreDeExito(t, salida), &hallazgos))

	porClase := make(map[grafo.ClaseDeHallazgo]int)
	for _, hallazgo := range hallazgos {
		porClase[hallazgo.Clase]++
	}

	assert.Equal(t, map[grafo.ClaseDeHallazgo]int{
		grafo.ClaseFuenteCaducada:  nodosDeLasNormas,
		grafo.ClaseVersionObsoleta: bloquesDelGrafoGrande * (versionesPorBloque - 1),
	}, porClase)
}

// datosDelSobreDeExito es el data de un sobre de éxito de graph.
func datosDelSobreDeExito(t *testing.T, salida []byte) json.RawMessage {
	t.Helper()

	var sobre struct {
		OK     bool            `json:"ok"`
		Fuente string          `json:"fuente"`
		Data   json.RawMessage `json:"data"`
	}

	require.NoError(t, json.Unmarshal(salida, &sobre))
	require.True(t, sobre.OK)
	require.Equal(t, "kitlegal.graph", sobre.Fuente)

	return sobre.Data
}

// ordenDeBoe es boe articulo del bloque a21 de la LPAC con --json, con
// --no-graph o sin él. Los argumentos van escritos enteros con constantes, que
// es lo que el análisis de seguridad exige de un subproceso.
func ordenDeBoe(ctx context.Context, binario string, sinEntrega bool) *exec.Cmd {
	if sinEntrega {
		return exec.CommandContext(ctx, binario, "boe", "articulo", "BOE-A-2015-10565", "a21", "--no-graph", "--json")
	}

	return exec.CommandContext(ctx, binario, "boe", "articulo", "BOE-A-2015-10565", "a21", "--json")
}

// ordenDelGrafo es graph check o graph stats con --json.
func ordenDelGrafo(ctx context.Context, binario, verbo string) *exec.Cmd {
	if verbo == "check" {
		return exec.CommandContext(ctx, binario, "graph", "check", "--json")
	}

	return exec.CommandContext(ctx, binario, "graph", "stats", "--json")
}

// medirOrden ejecuta la orden desde el directorio de trabajo con la carpeta de
// la caché como único entorno, exige que salga con 0 sin escribir nada en la
// salida de error y devuelve lo que tardó —arranque del proceso incluido, como
// cronometra— y su salida estándar.
func medirOrden(t *testing.T, orden *exec.Cmd, trabajo, carpetaDeLaCache string) (time.Duration, []byte) {
	t.Helper()

	var salida, errores bytes.Buffer

	orden.Dir = trabajo
	orden.Env = []string{cache.VariableDirectorio + "=" + carpetaDeLaCache}
	orden.Stdout = &salida
	orden.Stderr = &errores

	inicio := time.Now()
	err := orden.Run()
	duracion := time.Since(inicio)

	require.NoError(t, err, "%q: %s", orden.Args, errores.String())
	require.Empty(t, errores.String(), "%q no escribe nada en la salida de error", orden.Args)

	return duracion, salida.Bytes()
}

// mediana es la de las medidas: la central o, con un número par, la media de
// las dos centrales.
func mediana(medidas []time.Duration) time.Duration {
	ordenadas := slices.Sorted(slices.Values(medidas))
	mitad := len(ordenadas) / 2

	if len(ordenadas)%2 == 1 {
		return ordenadas[mitad]
	}

	return (ordenadas[mitad-1] + ordenadas[mitad]) / 2
}
