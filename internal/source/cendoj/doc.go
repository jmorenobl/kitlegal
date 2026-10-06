// Package cendoj es el adaptador de la fuente cendoj.jurisprudencia, el
// buscador de jurisprudencia del CENDOJ: resuelve una resolución judicial que
// la persona o el modelo ya han identificado, por el formulario del buscador,
// y entrega sus metadatos (ADR 0036; H23).
//
// # Qué consulta
//
// Una resolución identificada por una de tres referencias, con el campo del
// formulario que corresponde a cada una: su ECLI, su ROJ o su número de
// resolución con su fecha. Una consulta por resolución, que son dos
// peticiones: la página del buscador, que da la cookie de sesión, y el envío
// del formulario con esa cookie. De lo que el buscador responde reconoce dos
// respuestas y ninguna más —la lista de resultados y «No se ha encontrado
// ningún resultado»—, y de cada resultado lee sus metadatos: ECLI, ROJ, órgano
// y sala, fecha, número de resolución, número de recurso, ponente y la
// dirección de su documento.
//
// Todo se pide por internal/httpx, con el agente del proyecto, el robots.txt
// del sitio y el ritmo de la fila de la fuente en docs/SOURCES.md, cinco
// segundos, y sin reintentos. El envío de ese formulario es la única excepción
// a GET y HEAD del módulo (constitución, principio I).
//
// # Qué no consulta
//
//   - No busca por materia ni por texto libre: eso lo hace la persona en su
//     navegador, y vuelve con el identificador o con el texto.
//   - No pide el documento de ninguna resolución, ni una segunda página de
//     resultados, ni ninguna otra dirección, y no sigue redirecciones.
//   - No lee ni guarda el texto de la resolución, ni el resumen que el CENDOJ
//     da de ella.
//   - No usa un navegador, no cambia de agente y no sortea un CAPTCHA ni un
//     bloqueo: una respuesta que no es ninguna de las dos reconocidas termina
//     como límite o restricción de los términos de uso, y no se insiste.
//   - No resuelve las resoluciones del Tribunal Constitucional, que no están en
//     este buscador: quien compone el applet las reconoce por su ECLI antes de
//     llegar aquí.
//
// La integración continua no consulta el CENDOJ: los tests reproducen
// respuestas grabadas, y la comprobación contra la fuente real la lanza una
// persona.
package cendoj
