package bocm

import (
	"bytes"
	"encoding/json/v2"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"

	"github.com/jmorenobl/kitlegal/internal/core/plazos"
)

// Lo que fija el cotejo con el catálogo CKAN (FR-084; SC-007; ADR 0039): dónde
// está la evidencia, que este test solo lee, el año que el fichero del catálogo
// y el de kitlegal tienen en común y el código de entidad de las filas del
// municipio, las únicas con código INE (ADR 0039, punto 5).
const (
	rutaDelCatalogo     = "../../../evidencias/adr-0039/festivos_locales_historicos.json"
	rutaDelLeeme        = "../../../evidencias/adr-0039/LEEME.md"
	anioDelCotejo       = 2025
	entidadDelMunicipio = 0
)

// filaDelCatalogo es una fila del fichero CKAN, con los campos que nombra el
// LEEME.md de la evidencia y ninguno más.
type filaDelCatalogo struct {
	MunicipioCodigo string `json:"municipio_codigo"`
	EntidadCodigo   int    `json:"entidad_codigo"`
	MunicipioNombre string `json:"municipio_nombre"`
	Anio            int    `json:"año"`
	EntidadNombre   string `json:"entidad_nombre"`
	FechaFestivo    string `json:"fecha_festivo"`
}

// catalogo es el fichero CKAN entero.
type catalogo struct {
	Data []filaDelCatalogo `json:"data"`
}

// claseDeDiferencia es por qué el catálogo y kitlegal no coinciden en un
// municipio, con las palabras del LEEME.md.
type claseDeDiferencia string

const (
	// faltaEnElCatalogo: el municipio no tiene filas del año en el catálogo;
	// sus festivos los dio una corrección.
	faltaEnElCatalogo claseDeDiferencia = "falta en el fichero"
	// correccionSinAplicar: el catálogo trae la fecha de la resolución
	// principal y kitlegal la de la corrección que la cambió.
	correccionSinAplicar claseDeDiferencia = "corrección sin aplicar"
	// otraFecha: el catálogo da otra fecha sin ninguna corrección detrás;
	// kitlegal lleva la de la principal.
	otraFecha claseDeDiferencia = "otra fecha sin corrección"
	// lineaSinLeer: el catálogo lee la línea que kitlegal deja sin leer.
	lineaSinLeer claseDeDiferencia = "línea que no se lee sin interpretarla"
)

// diferencia es una de las diez del LEEME.md: el municipio tal como lo nombra
// el LEEME, su clase y, cuando el LEEME las dice, la fecha del catálogo y las
// de kitlegal.
type diferencia struct {
	nombre       string
	clase        claseDeDiferencia
	enElCatalogo string
	enKitlegal   []string
}

// diferenciasDelLeeme son las diez diferencias de FR-084, las del LEEME.md.
var diferenciasDelLeeme = []diferencia{
	{nombre: "Belmonte de Tajo", clase: faltaEnElCatalogo},
	{nombre: "Cabanillas de la Sierra", clase: faltaEnElCatalogo},
	{nombre: "La Hiruela", clase: faltaEnElCatalogo},
	{nombre: "Montejo de la Sierra", clase: faltaEnElCatalogo},
	{nombre: "Patones", clase: faltaEnElCatalogo},
	{nombre: "Talamanca de Jarama", clase: faltaEnElCatalogo},
	{nombre: "Daganzo de Arriba", clase: correccionSinAplicar, enElCatalogo: "2025-09-01", enKitlegal: []string{"2025-09-08"}},
	{nombre: "Majadahonda", clase: correccionSinAplicar, enElCatalogo: "2025-09-14", enKitlegal: []string{"2025-09-15"}},
	{
		nombre: "Fuente el Saz de Jarama", clase: otraFecha, enElCatalogo: "2025-09-06",
		enKitlegal: []string{"2025-09-08", "2025-09-09"},
	},
	{nombre: "Valdeolmos-Alalpardo", clase: lineaSinLeer},
}

// festivosDelCatalogo son, por código INE, las fechas del año de las filas del
// municipio del catálogo, ordenadas. Cada fila del año tiene que ser de uno de
// los municipios de la comunidad: un municipio del catálogo que no está en el
// territorio dejaría el cotejo incompleto.
func festivosDelCatalogo(t *testing.T, filas []filaDelCatalogo, codigos []string) map[string][]string {
	t.Helper()

	porCodigo := map[string][]string{}

	for _, fila := range filas {
		if fila.Anio != anioDelCotejo || fila.EntidadCodigo != entidadDelMunicipio {
			continue
		}

		delMunicipio := []string{}

		for _, codigo := range codigos {
			if strings.HasSuffix(codigo, fila.MunicipioCodigo) {
				delMunicipio = append(delMunicipio, codigo)
			}
		}

		require.Len(t, delMunicipio, 1, "el municipio %s (%s) del catálogo no es uno solo de la comunidad",
			fila.MunicipioCodigo, fila.MunicipioNombre)

		porCodigo[delMunicipio[0]] = append(porCodigo[delMunicipio[0]], fila.FechaFestivo)
	}

	for codigo := range porCodigo {
		slices.Sort(porCodigo[codigo])
	}

	return porCodigo
}

// festivosDeKitlegal son, por código INE, las fechas del año del fichero local,
// ordenadas, y el source de cada municipio.
func festivosDeKitlegal(fichero plazos.FestivosLocales) (map[string][]string, map[string]string) {
	porCodigo, sources := map[string][]string{}, map[string]string{}

	for _, festivo := range fichero.Festivos {
		if strings.HasPrefix(festivo.Fecha, fmt.Sprintf("%d-", anioDelCotejo)) {
			porCodigo[festivo.Municipio] = append(porCodigo[festivo.Municipio], festivo.Fecha)
			sources[festivo.Municipio] = festivo.Source
		}
	}

	for codigo := range porCodigo {
		slices.Sort(porCodigo[codigo])
	}

	return porCodigo, sources
}

// comprobarDiferencia comprueba que la diferencia de un municipio es de la
// clase que el LEEME le da.
func comprobarDiferencia(t *testing.T, d diferencia, codigo, principal string, delCatalogo, deKitlegal []string,
	source string, faltas []plazos.FaltaLocal,
) {
	t.Helper()

	nombre := fmt.Sprintf("%s (%s)", d.nombre, codigo)

	switch d.clase {
	case faltaEnElCatalogo:
		assert.Empty(t, delCatalogo, "%s: el catálogo no lo tiene", nombre)
		assert.NotEmpty(t, deKitlegal, "%s: kitlegal tiene sus festivos", nombre)
		assert.NotEqual(t, principal, source, "%s: sus festivos los dio una corrección", nombre)
	case correccionSinAplicar:
		assert.Contains(t, delCatalogo, d.enElCatalogo, "%s: el catálogo trae la fecha de la principal", nombre)
		assert.NotContains(t, deKitlegal, d.enElCatalogo, "%s: kitlegal no trae la fecha de la principal", nombre)
		assert.Subset(t, deKitlegal, d.enKitlegal, "%s: kitlegal trae la de la corrección", nombre)
		assert.NotEqual(t, principal, source, "%s: el source es la corrección", nombre)
	case otraFecha:
		assert.Contains(t, delCatalogo, d.enElCatalogo, "%s: el catálogo da otra fecha", nombre)
		assert.NotContains(t, deKitlegal, d.enElCatalogo, "%s: kitlegal no la da", nombre)
		assert.Subset(t, deKitlegal, d.enKitlegal, "%s: kitlegal trae las de la principal", nombre)
		assert.Equal(t, principal, source, "%s: ninguna corrección detrás", nombre)
	case lineaSinLeer:
		assert.NotEmpty(t, delCatalogo, "%s: el catálogo lee la línea", nombre)
		assert.Empty(t, deKitlegal, "%s: kitlegal no entra en sus fechas", nombre)
		assert.Contains(t, faltas, plazos.FaltaLocal{
			Anio: anioDelCotejo, Municipio: codigo, Motivo: plazos.FaltaLineaSinLeer, Source: principal,
		}, "%s: queda fuera de cobertura con su motivo", nombre)
	default:
		t.Fatalf("%s: la clase %q no es una del LEEME", nombre, d.clase)
	}
}

// festivosEscritos son las filas de festivos de data/festivos/locales/13.yaml,
// el fichero del árbol.
func festivosEscritos(t *testing.T) []plazos.FestivoLocal {
	t.Helper()

	escrito, err := os.ReadFile(rutaDelFichero)
	require.NoError(t, err, "el fichero lo escribe TestFestivosLocalesDerivados con -actualizar-festivos")

	var fichero plazos.FestivosLocales

	lector := yaml.NewDecoder(bytes.NewReader(escrito))
	lector.KnownFields(true)
	require.NoError(t, lector.Decode(&fichero))

	return fichero.Festivos
}

// TestCotejoConElCatalogo coteja los festivos locales de 2025 de data/festivos/
// —lo que sale de lo grabado, que es lo que lleva data/festivos/locales/13.yaml
// (FR-083) y el test lo comprueba— con el fichero del catálogo CKAN de la
// evidencia del ADR 0039, por código INE y contra las filas del municipio
// (entidad_codigo 0), y exige exactamente las diez diferencias de su LEEME.md,
// ninguna más ni ninguna menos, cada una de su clase (FR-084; SC-007). Lee el
// catálogo y el LEEME.md sin cambiarlos.
func TestCotejoConElCatalogo(t *testing.T) {
	t.Parallel()

	registro := registroDelRepositorio(t)
	municipios := registro.MunicipiosDe(ComunidadDelBoletin)
	require.Len(t, municipios, numeroDeMunicipios)

	codigos := make([]string, 0, len(municipios))
	for _, municipio := range municipios {
		codigos = append(codigos, municipio.Codigo.String())
	}

	contenido, err := os.ReadFile(rutaDelCatalogo)
	require.NoError(t, err)

	var fichero catalogo
	require.NoError(t, json.Unmarshal(contenido, &fichero, json.RejectUnknownMembers(true)))
	require.NotEmpty(t, fichero.Data)

	leeme, err := os.ReadFile(rutaDelLeeme)
	require.NoError(t, err)

	locales, err := Derivar(t.Context(), clienteDeLoGrabado(t), manifiestoDelPaquete(t), registro)
	require.NoError(t, err)

	indice := slices.IndexFunc(locales.Anios, func(a plazos.AnioLocal) bool { return a.Anio == anioDelCotejo })
	require.GreaterOrEqual(t, indice, 0, "el fichero no tiene el año %d", anioDelCotejo)
	require.NotEmpty(t, locales.Anios[indice].Disposiciones)

	principal := locales.Anios[indice].Disposiciones[0]
	delCatalogo := festivosDelCatalogo(t, fichero.Data, codigos)
	deKitlegal, sources := festivosDeKitlegal(locales)

	var distintos []string

	for _, codigo := range codigos {
		if !slices.Equal(delCatalogo[codigo], deKitlegal[codigo]) {
			distintos = append(distintos, codigo)
		}
	}

	c := nuevoCruce(municipios)
	esperados := make([]string, 0, len(diferenciasDelLeeme))

	for _, d := range diferenciasDelLeeme {
		assert.Contains(t, string(leeme), d.nombre, "el LEEME.md nombra la diferencia")

		codigo, casa := c.municipioDe(d.nombre)
		require.True(t, casa, "«%s» es un municipio de la comunidad", d.nombre)

		esperados = append(esperados, codigo)
		comprobarDiferencia(t, d, codigo, principal, delCatalogo[codigo], deKitlegal[codigo], sources[codigo], locales.Faltas)
	}

	slices.Sort(esperados)
	assert.Equal(t, esperados, distintos, "las diferencias con el catálogo son exactamente las diez del LEEME.md")

	assert.ElementsMatch(t, locales.Festivos, festivosEscritos(t), "lo cotejado es lo que lleva data/festivos/")
}
