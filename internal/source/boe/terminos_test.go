package boe

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core"
)

// rutaDeSources es docs/SOURCES.md visto desde este paquete: la tabla de fuentes
// que revisa una persona y a la que este test ata las constantes de terminos.go.
const rutaDeSources = "../../../docs/SOURCES.md"

// Lo que el test lee de la tabla de fuentes (contrato esquemas-fixtures-y-controles
// §8).
const (
	// columnaRitmo, columnaTerminos y columnaRevisado son las tres celdas de la
	// fila que tienen que coincidir con terminos.go.
	columnaRitmo    = "Ritmo"
	columnaTerminos = "Términos de uso"
	columnaRevisado = "Revisado"
	// revisionPendiente es la celda «Revisado» de la fila propuesta que dejó la
	// tarea del arnés de grabación. Desde que una persona revisó la fila en la
	// pausa del manifiesto ya no se admite: solo vale la fecha real de la
	// revisión.
	revisionPendiente = "pendiente"
)

// cabeceraDeSources es la cabecera de la tabla del contrato §8, celda a celda y en
// su orden.
var cabeceraDeSources = []string{
	"Fuente", "Applet", "Base", "Licencia", columnaTerminos, "robots.txt", columnaRitmo, "Formato", columnaRevisado,
}

// separadorDeCelda es una celda de la fila que separa la cabecera de los datos en
// una tabla de Markdown.
var separadorDeCelda = regexp.MustCompile(`^:?-+:?$`)

// direccionEntreAngulos es una dirección escrita como <…>, la forma en que la
// celda «Términos de uso» lleva la suya.
var direccionEntreAngulos = regexp.MustCompile(`<([^<>\s]+)>`)

// TestFuenteCoincideConSources ata la fila boe.legislacion-consolidada de
// docs/SOURCES.md a las constantes de terminos.go (FR-121, FR-122; research.md
// D15): el ritmo de la fila es IntervaloEntrePeticiones, la dirección de sus
// términos de uso es terminosDeUso.URL y su «Revisado» es terminosDeUso.Revisados
// como fecha AAAA-MM-DD, que no es cero. La fila propuesta, «Revisado» pendiente
// con Revisados cero, dejó de admitirse en la tarea que sigue a la pausa del
// manifiesto, en la que una persona revisó la fila (contrato
// esquemas-fixtures-y-controles §8).
//
// El primer subtest compara la tabla real; el resto demuestra sobre tablas en
// memoria que la comparación no pasa en vacío.
func TestFuenteCoincideConSources(t *testing.T) {
	t.Parallel()

	t.Run("docs-SOURCES", func(t *testing.T) {
		t.Parallel()

		contenido, err := os.ReadFile(rutaDeSources)
		require.NoError(t, err)

		require.NoError(t, comprobarFilaDeLaFuente(
			string(contenido), NombreDeLaFuente, IntervaloEntrePeticiones, terminosDeUso))
	})

	const (
		terminos        = "https://www.boe.es/terminos"
		fuente          = "`boe.legislacion-consolidada`"
		base            = "<https://www.boe.es/datosabiertos/api/legislacion-consolidada>"
		fechaDeRevision = "2026-09-14"
	)

	revisados := time.Date(2026, time.September, 14, 0, 0, 0, 0, time.UTC)

	fila := func(ritmo, deLosTerminos, revisado string) string {
		return "| " + strings.Join([]string{
			fuente, "`boe`", base, "por revisar", deLosTerminos, "por revisar", ritmo,
			"XML (bloque) y JSON; sin autenticación", revisado,
		}, " | ") + " |"
	}

	casos := []struct {
		nombre    string
		filas     []string
		sinTabla  bool
		revisados time.Time
		// mensaje es un fragmento del error esperado; vacío, la fila coincide.
		mensaje string
	}{
		{
			nombre:    "revisada-con-su-fecha",
			filas:     []string{fila("`1s`", "<"+terminos+">", fechaDeRevision)},
			revisados: revisados,
		},
		{
			nombre:   "sin-tabla",
			sinTabla: true,
			mensaje:  "no tiene la tabla de fuentes",
		},
		{
			nombre:  "sin-fila",
			mensaje: "ninguna fila",
		},
		{
			nombre: "fila-repetida",
			filas: []string{
				fila("`1s`", "<"+terminos+">", fechaDeRevision),
				fila("`1s`", "<"+terminos+">", fechaDeRevision),
			},
			revisados: revisados,
			mensaje:   "2 filas",
		},
		{
			nombre:  "celdas-de-menos",
			filas:   []string{"| " + fuente + " | `boe` |"},
			mensaje: "2 celdas",
		},
		{
			nombre:    "ritmo-distinto",
			filas:     []string{fila("`2s`", "<"+terminos+">", fechaDeRevision)},
			revisados: revisados,
			mensaje:   "«Ritmo» es 2s",
		},
		{
			nombre:    "ritmo-ilegible",
			filas:     []string{fila("un segundo", "<"+terminos+">", fechaDeRevision)},
			revisados: revisados,
			mensaje:   "«Ritmo» no es una duración",
		},
		{
			nombre:    "terminos-distintos",
			filas:     []string{fila("`1s`", "<https://www.boe.es/otros>", fechaDeRevision)},
			revisados: revisados,
			mensaje:   "«Términos de uso» es https://www.boe.es/otros",
		},
		{
			nombre:    "terminos-sin-direccion",
			filas:     []string{fila("`1s`", terminos, fechaDeRevision)},
			revisados: revisados,
			mensaje:   "«Términos de uso» lleva 0 direcciones",
		},
		{
			// La fila propuesta, que se admitía hasta la pausa del manifiesto.
			nombre:  "pendiente-con-cero",
			filas:   []string{fila("`1s`", "<"+terminos+">", revisionPendiente)},
			mensaje: "«Revisado» no es una fecha AAAA-MM-DD",
		},
		{
			nombre:    "pendiente-con-fecha",
			filas:     []string{fila("`1s`", "<"+terminos+">", revisionPendiente)},
			revisados: revisados,
			mensaje:   "«Revisado» no es una fecha AAAA-MM-DD",
		},
		{
			nombre:  "fecha-con-cero",
			filas:   []string{fila("`1s`", "<"+terminos+">", fechaDeRevision)},
			mensaje: "«Revisado» es 2026-09-14 y terminosDeUso.Revisados, cero",
		},
		{
			nombre:    "fecha-distinta",
			filas:     []string{fila("`1s`", "<"+terminos+">", "2026-09-15")},
			revisados: revisados,
			mensaje:   "«Revisado» es 2026-09-15",
		},
		{
			nombre:    "fecha-ilegible",
			filas:     []string{fila("`1s`", "<"+terminos+">", "14/09/2026")},
			revisados: revisados,
			mensaje:   "«Revisado» no es una fecha AAAA-MM-DD",
		},
		{
			nombre:  "fecha-cero",
			filas:   []string{fila("`1s`", "<"+terminos+">", "0001-01-01")},
			mensaje: "«Revisado» es la fecha cero",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			lineas := []string{"# Fuentes", ""}
			if !caso.sinTabla {
				lineas = append(lineas,
					"| "+strings.Join(cabeceraDeSources, " | ")+" |",
					"|"+strings.Repeat("---|", len(cabeceraDeSources)))
			}

			lineas = append(lineas, caso.filas...)
			lineas = append(lineas, "", "Texto que sigue a la tabla.")

			err := comprobarFilaDeLaFuente(strings.Join(lineas, "\n"), NombreDeLaFuente, time.Second,
				core.Terminos{URL: terminos, Revisados: caso.revisados})

			if caso.mensaje == "" {
				require.NoError(t, err)

				return
			}

			require.ErrorContains(t, err, caso.mensaje)
		})
	}
}

// comprobarFilaDeLaFuente localiza en la tabla de fuentes la fila cuya primera
// celda es el nombre de la fuente y comprueba que su ritmo, la dirección de sus
// términos de uso y su fecha de revisión son los que declara el adaptador.
func comprobarFilaDeLaFuente(tabla, fuente string, intervalo time.Duration, terminos core.Terminos) error {
	celdas, err := filaDeLaFuente(tabla, fuente)
	if err != nil {
		return err
	}

	return errors.Join(
		comprobarRitmo(celdas[slices.Index(cabeceraDeSources, columnaRitmo)], intervalo),
		comprobarTerminos(celdas[slices.Index(cabeceraDeSources, columnaTerminos)], terminos.URL),
		comprobarRevision(celdas[slices.Index(cabeceraDeSources, columnaRevisado)], terminos.Revisados),
	)
}

// filaDeLaFuente devuelve las celdas de la única fila de la tabla de fuentes cuya
// primera celda, sin comillas invertidas, es el nombre de la fuente. La tabla es
// la que empieza en la cabecera del contrato §8 y llega hasta la primera línea que
// no es de tabla.
func filaDeLaFuente(tabla, fuente string) ([]string, error) {
	lineas := strings.Split(tabla, "\n")

	inicio := slices.IndexFunc(lineas, func(linea string) bool {
		celdas, esDeTabla := celdasDe(linea)

		return esDeTabla && slices.Equal(celdas, cabeceraDeSources)
	})
	if inicio < 0 {
		return nil, fmt.Errorf("docs/SOURCES.md no tiene la tabla de fuentes con la cabecera %s",
			strings.Join(cabeceraDeSources, ", "))
	}

	var encontradas [][]string

	for numero, linea := range lineas[inicio+1:] {
		celdas, esDeTabla := celdasDe(linea)
		if !esDeTabla {
			break
		}

		if slices.ContainsFunc(celdas, separadorDeCelda.MatchString) {
			continue
		}

		if strings.Trim(celdas[0], "`") != fuente {
			continue
		}

		if len(celdas) != len(cabeceraDeSources) {
			return nil, fmt.Errorf("docs/SOURCES.md, línea %d: la fila de %s tiene %d celdas y la tabla, %d",
				inicio+numero+2, fuente, len(celdas), len(cabeceraDeSources))
		}

		encontradas = append(encontradas, celdas)
	}

	switch len(encontradas) {
	case 1:
		return encontradas[0], nil
	case 0:
		return nil, fmt.Errorf("docs/SOURCES.md no tiene ninguna fila de %s", fuente)
	default:
		return nil, fmt.Errorf("docs/SOURCES.md tiene %d filas de %s y tiene que tener una", len(encontradas), fuente)
	}
}

// celdasDe parte una línea de tabla de Markdown en sus celdas, sin el espacio que
// las rodea, y dice si la línea es de tabla: empieza y termina por «|».
func celdasDe(linea string) ([]string, bool) {
	linea = strings.TrimSpace(linea)
	if len(linea) < 2 || !strings.HasPrefix(linea, "|") || !strings.HasSuffix(linea, "|") {
		return nil, false
	}

	celdas := strings.Split(linea[1:len(linea)-1], "|")
	for indice, celda := range celdas {
		celdas[indice] = strings.TrimSpace(celda)
	}

	return celdas, true
}

// comprobarRitmo exige que la celda «Ritmo», un literal de duración de Go entre
// comillas invertidas, sea el intervalo entre peticiones de la fuente.
func comprobarRitmo(celda string, intervalo time.Duration) error {
	ritmo, err := time.ParseDuration(strings.Trim(celda, "`"))
	if err != nil {
		return fmt.Errorf("la celda «%s» no es una duración de Go: %w", columnaRitmo, err)
	}

	if ritmo != intervalo {
		return fmt.Errorf("la celda «%s» es %s y IntervaloEntrePeticiones, %s", columnaRitmo, ritmo, intervalo)
	}

	return nil
}

// comprobarTerminos exige que la celda «Términos de uso» lleve una sola dirección
// entre <…> y que sea la de los términos de uso de la fuente.
func comprobarTerminos(celda, direccion string) error {
	encontradas := direccionEntreAngulos.FindAllStringSubmatch(celda, -1)
	if len(encontradas) != 1 {
		return fmt.Errorf("la celda «%s» lleva %d direcciones entre <…> y tiene que llevar una: %s",
			columnaTerminos, len(encontradas), celda)
	}

	if encontradas[0][1] != direccion {
		return fmt.Errorf("la celda «%s» es %s y terminosDeUso.URL, %s", columnaTerminos, encontradas[0][1], direccion)
	}

	return nil
}

// comprobarRevision exige que la celda «Revisado» sea una fecha AAAA-MM-DD que no
// es cero e igual a la fecha de revisión de los términos. Cualquier otra celda,
// también pendiente, la de la fila propuesta, es un error.
func comprobarRevision(celda string, revisados time.Time) error {
	fecha, err := time.Parse(time.DateOnly, celda)
	if err != nil {
		return fmt.Errorf("la celda «%s» no es una fecha AAAA-MM-DD: %w", columnaRevisado, err)
	}

	if fecha.IsZero() {
		return fmt.Errorf("la celda «%s» es la fecha cero, que no es ninguna revisión", columnaRevisado)
	}

	if !fecha.Equal(revisados) {
		return fmt.Errorf("la celda «%s» es %s y terminosDeUso.Revisados, %s",
			columnaRevisado, celda, formatoDeRevisados(revisados))
	}

	return nil
}

// formatoDeRevisados escribe la fecha de revisión como la celda «Revisado», o dice
// que es cero.
func formatoDeRevisados(revisados time.Time) string {
	if revisados.IsZero() {
		return "cero"
	}

	return revisados.Format(time.DateOnly)
}
