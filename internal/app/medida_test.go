//go:build integration

package app_test

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core"
	"github.com/jmorenobl/kitlegal/internal/core/grafo"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/graph"
	"github.com/jmorenobl/kitlegal/internal/source/boe"
)

// El grafo de la medida de la bitácora (contracts/arnes-e2e.md §4; research.md
// D24): 300 normas de 8 bloques con la forma de lo que emite boe. Cada norma se
// leyó una vez entera, con la redacción A de todos sus bloques: las 270
// primeras, el 90 %, más de una semana antes de T8, y las 30 últimas dentro de
// la vigencia de boe a esa hora. El bloque a1 de las 240 primeras se leyó otra
// vez, también más de una semana antes de T8, y la fuente sirvió la redacción B.
const (
	normasDeLaMedida           = 300
	bloquesPorNormaDeLaMedida  = 8
	normasLeidasHaceMasDe7Dias = 270
	normasReleidas             = 240
	bloqueReleido              = "a1"

	redaccionA = "20200101"
	redaccionB = "20250101"

	primeraLecturaAntigua  = "2026-09-01T10:00:00Z"
	segundaLectura         = "2026-09-15T10:00:00Z"
	primeraLecturaReciente = "2026-10-05T10:00:00Z"

	// primeraNormaDeLaMedida es la norma 0, la que acotan las dos comprobaciones
	// con ámbito, y eliDeLaPrimeraNorma, el id de su Norma.
	primeraNormaDeLaMedida = "BOE-A-2020-1000"
	eliDeLaPrimeraNorma    = "eli/es/l/2020/01/01/1"
)

// Lo que lee la skill en una pregunta (SC-005): los bloques a21 a a25 de la
// LPAC, con los ids y las direcciones que emite boe, leídos con una redacción y
// después con otra posterior, y leídos otra vez con la posterior; las tres
// lecturas, dentro de la vigencia de boe a la hora de T8.
const (
	normaDeLaSkill = "BOE-A-2015-10565"
	eliDeLaLPAC    = "eli/es/l/2015/10/01/39"

	redaccionAnteriorDeLaSkill  = "20161002"
	redaccionPosteriorDeLaSkill = "20250101"

	primeraLecturaDeLaSkill = "2026-10-01T10:00:00Z"
	segundaLecturaDeLaSkill = "2026-10-02T10:00:00Z"
	terceraLecturaDeLaSkill = "2026-10-03T10:00:00Z"
)

// Lo que boe declara al emitir un artículo: la vigencia de su consulta,
// 604 800 s, y la API de la que sale cada dirección.
const (
	vigenciaDeBoe    = 604_800 * time.Second
	baseDeLaAPIDeBoe = "https://www.boe.es/datosabiertos/api/legislacion-consolidada"
)

// Las cotas de bytes de la salida estándar de graph check --json: sin
// argumentos o con la norma, sobre la medida (FR-013, SC-001), y con la norma y
// los cinco bloques cambiados de la skill (SC-005).
const (
	maximoDeBytesDelCheck  = 40_000
	maximoDeBytesDeLaSkill = 3_800
)

// hallazgoDeLaMedida es lo que se afirma de cada hallazgo listado: su clase y
// el id del nodo del que se dice.
type hallazgoDeLaMedida struct {
	clase grafo.ClaseDeHallazgo
	id    string
}

// normaDeBoe es una norma con los identificadores que emite boe: su
// identificador BOE y el ELI de su url_eli, que es el id de su Norma.
type normaDeBoe struct {
	identificador string
	eli           string
}

// TestMedidaDelGrafo siembra con el propio almacén el grafo de la medida y
// comprueba con el binario de reloj T8 lo que cuesta graph check --json sobre
// él: sin argumentos, 50 hallazgos, los version-obsoleta primero, con el total
// de cada clase y lo omitido, en 40 000 bytes como mucho (SC-001); con la norma,
// sus 18 en la misma cota (FR-013); y con la norma y un bloque, todos los de ese
// ámbito y ninguno más (SC-002). Publica los bytes medidos (FR-092).
func TestMedidaDelGrafo(t *testing.T) {
	t.Parallel()

	binario := binarioDeT8(t)
	directorio := t.TempDir()
	sembrarLaMedida(t, directorio)

	compruebaLaMedidaSinArgumentos(t, binario, directorio)
	compruebaLaMedidaConLaNorma(t, binario, directorio)
	compruebaLaMedidaConLaNormaYUnBloque(t, binario, directorio)
}

// TestLoQueLeeLaSkill comprueba lo que lee boe-legislacion por pregunta
// (SC-005): con los cinco bloques cuya última lectura vio una redacción nueva,
// cinco version-obsoleta y ningún fuente-caducada en 3 800 bytes como mucho; y,
// leídos otra vez, ninguno. Publica los bytes medidos.
func TestLoQueLeeLaSkill(t *testing.T) {
	t.Parallel()

	binario := binarioDeT8(t)
	directorio := t.TempDir()
	almacen := graph.Nuevo(graph.ConDirectorio(directorio))
	lpac := normaDeBoe{identificador: normaDeLaSkill, eli: eliDeLaLPAC}

	leerCadaBloque(t, almacen, lpac, redaccionAnteriorDeLaSkill, primeraLecturaDeLaSkill)
	leerCadaBloque(t, almacen, lpac, redaccionPosteriorDeLaSkill, segundaLecturaDeLaSkill)

	salida, comprobacion := comprobarLaSkill(t, binario, directorio)

	t.Logf("SC-005: graph check --json con la norma y los cinco bloques cambiados: %d bytes (m\xc3\xa1ximo %d)",
		len(salida), maximoDeBytesDeLaSkill)
	assert.LessOrEqual(t, len(salida), maximoDeBytesDeLaSkill, "SC-005")

	var superadas []hallazgoDeLaMedida
	for _, bloque := range bloquesDeLaSkill() {
		superadas = append(superadas, hallazgoDeLaMedida{
			clase: grafo.ClaseVersionObsoleta, id: lpac.redaccion(bloque, redaccionAnteriorDeLaSkill).id,
		})
	}

	assert.Equal(t, grafo.Comprobacion{
		Norma: normaDeLaSkill, Bloques: bloquesDeLaSkill(), VersionObsoleta: len(superadas),
		Hallazgos: comprobacion.Hallazgos,
	}, comprobacion)
	assert.Equal(t, superadas, clasesEIds(comprobacion.Hallazgos))

	leerCadaBloque(t, almacen, lpac, redaccionPosteriorDeLaSkill, terceraLecturaDeLaSkill)

	salida, comprobacion = comprobarLaSkill(t, binario, directorio)

	t.Logf("SC-005: graph check --json con la norma y los cinco bloques leídos otra vez: %d bytes", len(salida))
	assert.Equal(t, grafo.Comprobacion{Norma: normaDeLaSkill, Bloques: bloquesDeLaSkill(), Hallazgos: []grafo.Hallazgo{}},
		comprobacion)
}

// compruebaLaMedidaSinArgumentos exige de graph check --json sin argumentos
// los 50 primeros hallazgos, todos version-obsoleta —hay 240, uno por cada a1
// leído dos veces, y esa clase va primero—: las redacciones A de a1 por id
// comparando bytes, cada una superada por su B. Los totales cuentan también lo
// omitido: 4 590 fuente-caducada, sobre la Norma, los 8 Bloque y sus 8
// redacciones vistas de cada una de las 270 normas leídas hace más de una
// semana, y 4 830 − 50 omitidos (SC-001).
func compruebaLaMedidaSinArgumentos(t *testing.T, binario, directorio string) {
	t.Helper()

	salida, comprobacion := comprobarConT8(t, exec.CommandContext(t.Context(), binario, "graph", "check", "--json"),
		directorio)

	t.Logf("SC-001: graph check --json sin argumentos sobre la medida: %d bytes (m\xc3\xa1ximo %d)", len(salida),
		maximoDeBytesDelCheck)
	assert.LessOrEqual(t, len(salida), maximoDeBytesDelCheck, "SC-001")
	assert.Contains(t, string(salida),
		`"data":{"norma":"","bloques":[],"version-obsoleta":240,"fuente-caducada":4590,"omitidos":4780,"hallazgos":[`)

	superadas := make([]string, 0, normasReleidas)
	for i := range normasReleidas {
		superadas = append(superadas, normaDeLaMedida(i).redaccion(bloqueReleido, redaccionA).id)
	}

	slices.Sort(superadas)

	listadas := make([]hallazgoDeLaMedida, 0, 50)
	for _, id := range superadas[:50] {
		listadas = append(listadas, hallazgoDeLaMedida{clase: grafo.ClaseVersionObsoleta, id: id})
	}

	assert.Equal(t, listadas, clasesEIds(comprobacion.Hallazgos))

	for _, hallazgo := range comprobacion.Hallazgos {
		assert.Equal(t, redaccionA, hallazgo.FechaVigencia, hallazgo.ID)
		assert.Equal(t, redaccionB, hallazgo.FechaVigenciaReciente, hallazgo.ID)
	}
}

// compruebaLaMedidaConLaNorma exige de graph check --json con la norma 0 sus
// 18 hallazgos, todos de ella y en la cota de 40 000 bytes (FR-013): el
// version-obsoleta sobre la A de a1 y 17 fuente-caducada, sobre la Norma, sus 8
// Bloque y sus 8 redacciones vistas —la B de a1 y la A de los demás—.
func compruebaLaMedidaConLaNorma(t *testing.T, binario, directorio string) {
	t.Helper()

	salida, comprobacion := comprobarConT8(t, exec.CommandContext(t.Context(), binario,
		"graph", "check", primeraNormaDeLaMedida, "--json"), directorio)

	t.Logf("FR-013: graph check --json con la norma sobre la medida: %d bytes (m\xc3\xa1ximo %d)", len(salida),
		maximoDeBytesDelCheck)
	assert.LessOrEqual(t, len(salida), maximoDeBytesDelCheck, "FR-013")

	norma := normaDeLaMedida(0)
	require.Equal(t, eliDeLaPrimeraNorma, norma.eli, "premisa: la norma 0 es la que se acota")

	caducadas := []string{norma.eli}
	for _, bloque := range bloquesDeLaMedida() {
		vista := redaccionA
		if bloque == bloqueReleido {
			vista = redaccionB
		}

		caducadas = append(caducadas, norma.eli+"#"+bloque, norma.redaccion(bloque, vista).id)
	}

	assert.Equal(t, grafo.Comprobacion{
		Norma: primeraNormaDeLaMedida, Bloques: []string{}, VersionObsoleta: 1, FuenteCaducada: 17,
		Hallazgos: comprobacion.Hallazgos,
	}, comprobacion)
	assert.Equal(t, hallazgosDeLaNorma(norma.redaccion(bloqueReleido, redaccionA).id, caducadas),
		clasesEIds(comprobacion.Hallazgos))
}

// compruebaLaMedidaConLaNormaYUnBloque exige de graph check --json con la
// norma 0 y su bloque a1 exactamente los 4 hallazgos de ese ámbito (SC-002): el
// version-obsoleta sobre la A de a1 y los fuente-caducada sobre la Norma, el
// Bloque a1 y su B, que es la redacción que vio su última lectura.
func compruebaLaMedidaConLaNormaYUnBloque(t *testing.T, binario, directorio string) {
	t.Helper()

	salida, comprobacion := comprobarConT8(t, exec.CommandContext(t.Context(), binario,
		"graph", "check", primeraNormaDeLaMedida, bloqueReleido, "--json"), directorio)

	t.Logf("SC-002: graph check --json con la norma y un bloque sobre la medida: %d bytes", len(salida))

	norma := normaDeLaMedida(0)
	caducadas := []string{norma.eli, norma.eli + "#" + bloqueReleido, norma.redaccion(bloqueReleido, redaccionB).id}

	assert.Equal(t, grafo.Comprobacion{
		Norma: primeraNormaDeLaMedida, Bloques: []string{bloqueReleido}, VersionObsoleta: 1, FuenteCaducada: 3,
		Hallazgos: comprobacion.Hallazgos,
	}, comprobacion)
	assert.Equal(t, hallazgosDeLaNorma(norma.redaccion(bloqueReleido, redaccionA).id, caducadas),
		clasesEIds(comprobacion.Hallazgos))
}

// binarioDeT8 es la ruta del binario de e2e con el reloj T8 que construye
// TestMain.
func binarioDeT8(t *testing.T) string {
	t.Helper()

	require.NoError(t, entorno.err)

	binario := entorno.variables[variableT8]
	require.NotEmpty(t, binario, "premisa: TestMain construye el binario de reloj T8")

	return binario
}

// comprobarConT8 ejecuta graph check desde el directorio sembrado, con él como
// carpeta de la caché y como único entorno, y devuelve su salida estándar y la
// data de su sobre.
func comprobarConT8(t *testing.T, orden *exec.Cmd, directorio string) ([]byte, grafo.Comprobacion) {
	t.Helper()

	_, salida := medirOrden(t, orden, directorio, directorio)

	var comprobacion grafo.Comprobacion

	require.NoError(t, json.Unmarshal(datosDelSobreDeExito(t, salida), &comprobacion))

	return salida, comprobacion
}

// comprobarLaSkill es graph check --json con la norma y los cinco bloques de la
// skill, como lo pide boe-legislacion antes de responder. Los argumentos van
// escritos enteros con constantes, que es lo que el análisis de seguridad exige
// de un subproceso.
func comprobarLaSkill(t *testing.T, binario, directorio string) ([]byte, grafo.Comprobacion) {
	t.Helper()

	return comprobarConT8(t, exec.CommandContext(t.Context(), binario,
		"graph", "check", normaDeLaSkill, "a21", "a22", "a23", "a24", "a25", "--json"), directorio)
}

// sembrarLaMedida entrega al world.db del directorio las lecturas de la medida
// en el orden en que se hicieron: la primera de las normas leídas hace más de
// una semana, un lote por norma con sus 8 bloques, como boe articulos; la
// segunda de a1 de las 240 primeras, un lote por bloque, como boe articulo; y la
// primera de las 30 leídas dentro de la vigencia.
func sembrarLaMedida(t *testing.T, directorio string) {
	t.Helper()

	almacen := graph.Nuevo(graph.ConDirectorio(directorio))
	entregar := func(lote core.Lote) {
		require.NoError(t, almacen.Apply(t.Context(), lote))
	}

	for i := range normasLeidasHaceMasDe7Dias {
		norma := normaDeLaMedida(i)
		entregar(norma.lectura(bloquesDeLaMedida(), redaccionA, direccionDeLaNorma(norma), primeraLecturaAntigua))
	}

	for i := range normasReleidas {
		norma := normaDeLaMedida(i)
		entregar(norma.lectura([]string{bloqueReleido}, redaccionB, direccionDelBloque(norma, bloqueReleido),
			segundaLectura))
	}

	for i := normasLeidasHaceMasDe7Dias; i < normasDeLaMedida; i++ {
		norma := normaDeLaMedida(i)
		entregar(norma.lectura(bloquesDeLaMedida(), redaccionA, direccionDeLaNorma(norma), primeraLecturaReciente))
	}
}

// leerCadaBloque entrega al almacén una lectura de cada bloque de la skill con
// la redacción de esa fecha de vigencia, un lote por bloque, como boe articulo.
func leerCadaBloque(t *testing.T, almacen *graph.Almacen, norma normaDeBoe, fechaVigencia, fechaConsulta string) {
	t.Helper()

	for _, bloque := range bloquesDeLaSkill() {
		require.NoError(t, almacen.Apply(t.Context(),
			norma.lectura([]string{bloque}, fechaVigencia, direccionDelBloque(norma, bloque), fechaConsulta)))
	}
}

// normaDeLaMedida es la norma i de la medida: BOE-A-2020-1 seguido de i con
// tres cifras, con el ELI eli/es/l/2020/01/01/<i+1>.
func normaDeLaMedida(i int) normaDeBoe {
	return normaDeBoe{
		identificador: fmt.Sprintf("BOE-A-2020-1%03d", i),
		eli:           "eli/es/l/2020/01/01/" + strconv.Itoa(i+1),
	}
}

// bloquesDeLaMedida son los 8 bloques de cada norma de la medida, a1 a a8.
func bloquesDeLaMedida() []string {
	bloques := make([]string, 0, bloquesPorNormaDeLaMedida)
	for k := range bloquesPorNormaDeLaMedida {
		bloques = append(bloques, "a"+strconv.Itoa(k+1))
	}

	return bloques
}

// bloquesDeLaSkill son los cinco bloques de la LPAC que lee la skill.
func bloquesDeLaSkill() []string {
	return []string{"a21", "a22", "a23", "a24", "a25"}
}

// direccionDeLaNorma es la url del sobre de boe articulos de la norma.
func direccionDeLaNorma(norma normaDeBoe) string {
	return baseDeLaAPIDeBoe + "/id/" + norma.identificador
}

// direccionDelBloque es la url del sobre de boe articulo del bloque de la
// norma.
func direccionDelBloque(norma normaDeBoe, bloque string) string {
	return direccionDeLaNorma(norma) + "/texto/bloque/" + bloque
}

// redaccionDeBoe es una redacción de un bloque: el id de su BloqueVersion, su
// cuerpo, inventado, y la huella de ese cuerpo.
type redaccionDeBoe struct {
	id     string
	cuerpo string
	huella string
}

// redaccion es la del bloque de la norma con esa fecha de vigencia: su id es
// <eli>#<bloque>@<fecha_vigencia>:<hash_texto>, como lo compone boe.
func (n normaDeBoe) redaccion(bloque, fechaVigencia string) redaccionDeBoe {
	cuerpo := fmt.Sprintf("Texto inventado de la redacci\xc3\xb3n %s del bloque %s de %s.", fechaVigencia, bloque,
		n.identificador)
	huella := "sha256:" + huellaSHA256([]byte(cuerpo))

	return redaccionDeBoe{id: n.eli + "#" + bloque + "@" + fechaVigencia + ":" + huella, cuerpo: cuerpo, huella: huella}
}

// lectura es el lote que el kernel entrega tras leer esos bloques de la norma
// cuando la fuente sirve de todos la redacción de esa fecha de vigencia: por
// cada bloque, las operaciones que emite boe (internal/source/boe/grafo.go),
// con la procedencia del sobre y la vigencia de boe.
func (n normaDeBoe) lectura(bloques []string, fechaVigencia, url, fechaConsulta string) core.Lote {
	operaciones := make([]schema.Operacion, 0, 6*len(bloques))

	for _, bloque := range bloques {
		idDelBloque := n.eli + "#" + bloque
		redaccion := n.redaccion(bloque, fechaVigencia)
		operaciones = append(operaciones,
			schema.Nodo{ID: n.eli, Tipo: grafo.TipoNorma, Datos: map[string]any{grafo.DatoIdentificador: n.identificador}},
			schema.Nodo{ID: idDelBloque, Tipo: grafo.TipoBloque, Datos: map[string]any{grafo.DatoBloque: bloque}},
			schema.Nodo{ID: redaccion.id, Tipo: grafo.TipoBloqueVersion, Datos: map[string]any{
				grafo.DatoFechaVigencia:     fechaVigencia,
				grafo.DatoFechaVersion:      fechaVigencia,
				grafo.DatoNormaModificadora: n.identificador,
				grafo.DatoHashTexto:         redaccion.huella,
			}},
			schema.Arista{Origen: n.eli, Relacion: grafo.RelacionTieneParte, Destino: idDelBloque},
			schema.Arista{Origen: idDelBloque, Relacion: grafo.RelacionTieneVersion, Destino: redaccion.id},
			schema.Texto{Huella: redaccion.huella, Cuerpo: redaccion.cuerpo},
		)
	}

	return core.Lote{
		Fuente:        boe.NombreDeLaFuente,
		URL:           url,
		FechaConsulta: fechaConsulta,
		Vigencia:      vigenciaDeBoe,
		Operaciones:   operaciones,
	}
}

// hallazgosDeLaNorma son, en el orden de graph check, el version-obsoleta
// sobre la redacción superada y los fuente-caducada sobre esos ids, por id
// comparando bytes.
func hallazgosDeLaNorma(superada string, caducadas []string) []hallazgoDeLaMedida {
	hallazgos := []hallazgoDeLaMedida{{clase: grafo.ClaseVersionObsoleta, id: superada}}

	for _, id := range slices.Sorted(slices.Values(caducadas)) {
		hallazgos = append(hallazgos, hallazgoDeLaMedida{clase: grafo.ClaseFuenteCaducada, id: id})
	}

	return hallazgos
}

// clasesEIds son la clase y el id de cada hallazgo listado, en su orden.
func clasesEIds(hallazgos []grafo.Hallazgo) []hallazgoDeLaMedida {
	resultado := make([]hallazgoDeLaMedida, 0, len(hallazgos))
	for _, hallazgo := range hallazgos {
		resultado = append(resultado, hallazgoDeLaMedida{clase: hallazgo.Clase, id: hallazgo.ID})
	}

	return resultado
}
