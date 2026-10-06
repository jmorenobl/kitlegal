package cendoj

import (
	"bytes"

	"github.com/jmorenobl/kitlegal/internal/httpx"
)

// claseDeRespuesta es lo que el adaptador reconoce en la respuesta del
// formulario: dos respuestas y ninguna más, y todo lo demás (FR-022; contrato
// fuente-cendoj-y-grabacion §3).
type claseDeRespuesta string

// Las tres clases de una respuesta del formulario.
const (
	// respuestaSinResultados es la que dice que la consulta no encuentra nada:
	// de ella, y solo de ella o del filtro por fecha, sale el «no encontrado».
	respuestaSinResultados claseDeRespuesta = "sin resultados"
	// respuestaConLista es la lista de resultados, de la que se leen las
	// resoluciones.
	respuestaConLista claseDeRespuesta = "lista"
	// respuestaNoReconocida es cualquier otra, con el estado que sea: un
	// bloqueo, un CAPTCHA, una redirección o una página que ha cambiado de
	// forma. No se interpreta ni se insiste.
	respuestaNoReconocida claseDeRespuesta = "no reconocida"
)

// estadoCorrecto es el único estado con el que una respuesta del buscador se
// reconoce. Se escribe aquí y no con net/http, que solo importa internal/httpx
// (R2).
const estadoCorrecto = 200

// Las dos marcas con las que se reconoce la respuesta del formulario, que son
// las que documenta docs/JURISPRUDENCIA.md §3 de lo que respondió el buscador
// (research D10, V29): el texto de una consulta sin resultados y el principio
// de la ruta del documento de una resolución, que lleva el enlace de cada
// resultado de la lista.
const (
	marcaSinResultados = "No se ha encontrado ningún resultado"
	marcaDeResultado   = "/search/AN/openDocument/"
)

// clasificar dice qué respuesta del formulario es la que recibe (contrato
// fuente-cendoj-y-grabacion §3): con estado 200 y el texto de la consulta sin
// resultados, «sin resultados»; con estado 200, sin ese texto y con algún
// enlace al documento de una resolución, «lista»; y cualquier otra, con el
// estado que sea, «no reconocida». Ninguna clase sale de un estado distinto de
// 200: un 404 no es un «no encontrado» (FR-022).
func clasificar(respuesta httpx.Respuesta) claseDeRespuesta {
	switch {
	case respuesta.Estado != estadoCorrecto:
		return respuestaNoReconocida
	case bytes.Contains(respuesta.Cuerpo, []byte(marcaSinResultados)):
		return respuestaSinResultados
	case bytes.Contains(respuesta.Cuerpo, []byte(marcaDeResultado)):
		return respuestaConLista
	default:
		return respuestaNoReconocida
	}
}
