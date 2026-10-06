package cendoj

import (
	"time"

	"github.com/jmorenobl/kitlegal/internal/core"
)

// Lo que la fuente declara de quién es y de cómo se la consulta: su nombre, el
// ritmo al que se le pide y los términos de uso bajo los que se hace. Los tres
// salen de la fila cendoj.jurisprudencia de docs/SOURCES.md, que revisa y fija
// una persona, y TestFuenteCoincideConSources falla si este fichero y esa fila
// divergen (FR-023; contrato fuente-cendoj-y-grabacion §1). Viven solos en este
// fichero para que quien revisa la fila los alinee con ella sin tocar ningún
// otro código.

// NombreDeLaFuente es el nombre de la fuente de este adaptador, el buscador de
// jurisprudencia del CENDOJ: el que va en la clave fuente del sobre y el que
// identifica sus grabaciones al grabarlas y al reproducirlas. No es del espacio
// reservado de lo que se obtiene sin consultar ninguna fuente pública (ADR
// 0006).
const NombreDeLaFuente = "cendoj.jurisprudencia"

// IntervaloEntrePeticiones es la separación mínima entre dos peticiones al
// sitio del buscador: la celda «Ritmo» de la fila de la fuente, que es el
// Crawl-delay de su robots.txt (FR-020; research D8).
const IntervaloEntrePeticiones = 5 * time.Second

// terminosDeUso son los términos de uso de la fuente: la dirección de la celda
// «Términos de uso» de la fila, que es la página del buscador, donde está su
// aviso legal, y el día de su celda «Revisado», a las 00:00 UTC.
var terminosDeUso = core.Terminos{
	URL:       "https://www.poderjudicial.es/search/indexAN.jsp",
	Revisados: time.Date(2026, time.October, 3, 0, 0, 0, 0, time.UTC),
}
