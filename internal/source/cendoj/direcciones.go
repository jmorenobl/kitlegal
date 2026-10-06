package cendoj

import "github.com/jmorenobl/kitlegal/internal/httpx"

// Las tres direcciones de la fuente, que salen de su fila de docs/SOURCES.md:
// la base es su celda «Base», y la página y el formulario, las dos rutas que
// nombra su celda «Formato». Las tres son https y del sitio del buscador, y son
// las únicas que la fuente pide: nunca el documento de una resolución ni
// ninguna otra (FR-020, FR-021; contrato fuente-cendoj-y-grabacion §1 y §2).
// TestFuenteCoincideConSources falla si divergen de la fila.
const (
	// baseDelBuscador es la base del buscador de jurisprudencia.
	baseDelBuscador = "https://www.poderjudicial.es/search"
	// paginaDelBuscador es la página que se pide primero y que da la cookie de
	// sesión con la que se envía el formulario.
	paginaDelBuscador = baseDelBuscador + "/indexAN.jsp"
	// formularioDeConsulta es la dirección a la que se envía el formulario, la
	// única a la que el cliente de la fuente puede hacer un POST, y la url del
	// sobre de una consulta.
	formularioDeConsulta = baseDelBuscador + "/search.action"
)

// Los tres campos de consulta del formulario, uno por forma de referencia: los
// que nombra la celda «Formato» de la fila de la fuente, y ninguno más
// (FR-006). Ninguna consulta lleva dos.
const (
	campoECLI               = "ECLI"
	campoROJ                = "ROJ"
	campoNumeroDeResolucion = "NUMERORESOLUCION"
)

// intentosPorPeticion son las veces que se pide cada petición a la fuente, la
// primera incluida: una. Ninguna petición al buscador se repite (FR-021;
// research D7).
const intentosPorPeticion = 1

// OpcionesDeRed son las opciones de internal/httpx con las que la fuente tiene
// que pedirse contra la red: su nombre, su formulario —sin el que el cliente no
// abre consultas ni envía nada— y un solo intento por petición. Quien compone
// añade el ritmo, con IntervaloEntrePeticiones, y el registrador (contrato
// fuente-cendoj-y-grabacion §1; data-model §7).
func OpcionesDeRed() []httpx.Opcion {
	return append(OpcionesDeReproduccion(), httpx.ConIntentos(intentosPorPeticion))
}

// OpcionesDeReproduccion son las de un cliente que reproduce grabaciones de la
// fuente: su nombre y su formulario. No llevan los intentos, que la
// reproducción no admite porque en su cadena nada reintenta.
func OpcionesDeReproduccion() []httpx.Opcion {
	return []httpx.Opcion{
		httpx.ConFuente(NombreDeLaFuente),
		httpx.ConFormulario(formularioDeConsulta),
	}
}
