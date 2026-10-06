package cendoj

import (
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/jmorenobl/kitlegal/internal/core"
)

// verboResolver es el verbo del applet cita que la consulta resuelve.
const verboResolver = "resolver"

// vigenciaDeResolver es la vigencia con la que se guarda lo que responde una
// consulta de resolver, sea una entrega o un «no encontrado»: 30 días (FR-040;
// contrato cita-resolver §5).
const vigenciaDeResolver = 30 * 24 * time.Hour

// ConsultaResolver es resolver <referencia>: la resolución identificada por un
// ECLI, por un ROJ o por un número de resolución con su fecha (FR-002).
type ConsultaResolver struct {
	// Referencia es la de la resolución, ya validada: la construye
	// NuevaReferencia.
	Referencia Referencia
}

// Verbo es resolver.
func (ConsultaResolver) Verbo() string {
	return verboResolver
}

// La consulta implementa el puerto del dominio por valor, y con ello también
// por puntero, y que lo siga haciendo no depende de que alguien lo recuerde.
var _ core.Consulta = ConsultaResolver{}

// Los dos campos del intervalo de fechas de resolución del formulario, que solo
// lleva la consulta por número de resolución, y cómo se escribe en ellos una
// fecha (docs/JURISPRUDENCIA.md §3; research D9).
const (
	campoFechaDesde = "FECHARESOLUCIONDESDE"
	campoFechaHasta = "FECHARESOLUCIONHASTA"
	// separadorDeLaFecha es el de una fecha AAAA-MM-DD, y barraDeLaFecha, el de
	// la misma fecha escrita dd/mm/aaaa, que es como la lleva el formulario.
	separadorDeLaFecha = "-"
	barraDeLaFecha     = "/"
)

// camposFijos son los cinco campos que lleva todo envío del formulario, los de
// la consulta que se probó a mano: una búsqueda en la base de jurisprudencia,
// con la primera página de diez resultados ordenados por fecha
// (docs/JURISPRUDENCIA.md §3; FR-020). Nunca se pide otra página. Cada llamada
// da un mapa nuevo, que quien lo recibe puede completar.
func camposFijos() map[string]string {
	return map[string]string{
		"action":         "query",
		"databasematch":  "AN",
		"recordsPerPage": "10",
		"sort":           "IN_FECHARESOLUCION:decreasing",
		"start":          "1",
	}
}

// camposDe son los campos del envío del formulario para una referencia: los
// cinco fijos y los de su forma, que es un solo campo de consulta (FR-006;
// contrato fuente-cendoj-y-grabacion §2):
//
//   - un ECLI, el campo ECLI;
//   - un ROJ, el campo ROJ, sin fechas también cuando la referencia lleva
//     fecha: esa fecha la compara después quien filtra lo encontrado;
//   - un número de resolución, el campo NUMERORESOLUCION con su fecha como
//     principio y como fin del intervalo de fechas de resolución.
//
// Una referencia que no construyó NuevaReferencia no tiene forma y no da ningún
// campo: sin campos, internal/httpx no envía nada, de modo que del adaptador no
// sale una consulta sin el campo de una referencia.
func camposDe(referencia Referencia) map[string]string {
	var propios map[string]string

	switch referencia.forma {
	case formaECLI:
		propios = map[string]string{campoECLI: referencia.valor}
	case formaROJ:
		propios = map[string]string{campoROJ: referencia.valor}
	case formaResolucion:
		fecha := fechaDelFormulario(referencia.fecha)
		propios = map[string]string{
			campoNumeroDeResolucion: referencia.valor,
			campoFechaDesde:         fecha,
			campoFechaHasta:         fecha,
		}
	default:
		return nil
	}

	campos := camposFijos()
	maps.Copy(campos, propios)

	return campos
}

// fechaDelFormulario escribe dd/mm/aaaa, como la lleva el formulario, una fecha
// AAAA-MM-DD ya validada: 2023-07-04 es 04/07/2023.
func fechaDelFormulario(fecha string) string {
	partes := strings.Split(fecha, separadorDeLaFecha)
	slices.Reverse(partes)

	return strings.Join(partes, barraDeLaFecha)
}
