package boe

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"
	"unicode/utf8"
)

// Las claves de las respuestas JSON que la lectura abre. Los campos de texto de
// cada verbo los nombra quien los lee.
const (
	// claveDeDatos es la del envoltorio de toda respuesta, la que trae lo pedido
	// (refs/boe.py 188, 278, 358, 455 y 514).
	claveDeDatos = "data"
	// claveDeBloques es la del elemento que anida los bloques del índice
	// (refs/boe.py 368-369).
	claveDeBloques = "bloque"
	// claveDeMaterias y envoltorioDeMateria son las materias del análisis y el
	// envoltorio de cada una (refs/boe.py 527 y 531).
	claveDeMaterias     = "materias"
	envoltorioDeMateria = "materia"
	// claveDeNotas y envoltorioDeNota son las notas del análisis y su envoltorio
	// (refs/boe.py 538 y 540).
	claveDeNotas     = "notas"
	envoltorioDeNota = "nota"
	// claveDeReferencias es la del objeto que agrupa las dos listas de
	// referencias del análisis (refs/boe.py 544-545).
	claveDeReferencias = "referencias"
	// claveDeAnteriores y envoltorioDeAnterior son las referencias a las normas
	// que esta modifica o deroga y el envoltorio de cada grupo (refs/boe.py 546
	// y 551).
	claveDeAnteriores    = "anteriores"
	envoltorioDeAnterior = "anterior"
	// claveDePosteriores y envoltorioDePosterior son las de las normas que la
	// modifican y el envoltorio de cada grupo (refs/boe.py 566 y 571).
	claveDePosteriores    = "posteriores"
	envoltorioDePosterior = "posterior"
)

// Los motivos de lo que no se puede leer: dicen qué no se pudo interpretar
// —el cuerpo, la raíz, el elemento o el campo— y nunca llevan el cuerpo ni un
// recorte suyo (contrato errores-y-codigos §2, filas 16 y 17).
const (
	motivoDelCuerpoNoUTF8   = "el cuerpo no es UTF-8 válido"
	motivoDelCuerpoIlegible = "el cuerpo no es JSON legible"
	motivoDeLaRaiz          = "la raíz del JSON no es un objeto, sino %s"
	motivoDelPrimerElemento = "el primer elemento de data no es un objeto, sino %s"
	motivoDelCampo          = "el campo %q no es texto, sino %s"
)

// leerEnvoltorio lee el cuerpo de una respuesta JSON de buscar, indice,
// metadatos o analisis y devuelve su data (data-model.md §3.1, research.md D7):
//
//   - J1: el cuerpo es JSON en UTF-8 —un único valor, con solo espacio en blanco
//     alrededor— y su raíz, un objeto. Un octeto que no forma UTF-8 lo hace
//     ilegible aunque vaya dentro de una cadena, donde encoding/json lo cambiaría
//     en silencio por U+FFFD y el texto presentado ya no sería el de la fuente;
//     refs/boe.py tampoco lo lee (decode("utf-8"), línea 82).
//   - J2: data ausente o nula es la lista vacía, como get("data", []). Si lo que
//     devuelve está vacío lo dice esVacio, y qué significa lo decide cada verbo
//     (J3).
//
// Los números llegan como json.Number, con su literal: ni se redondean en coma
// flotante ni se confunden con una cadena.
//
// Lo que no se puede leer es el error de errorDeLectura, con el detalle técnico
// del analizador como causa, que va al registro y no al mensaje.
func leerEnvoltorio(cuerpo []byte) (any, error) {
	if !utf8.Valid(cuerpo) {
		return nil, errorDeLectura(nil, motivoDelCuerpoNoUTF8)
	}

	decodificador := json.NewDecoder(bytes.NewReader(cuerpo))
	decodificador.UseNumber()

	var raiz any
	if err := decodificador.Decode(&raiz); err != nil {
		return nil, errorDeLectura(err, motivoDelCuerpoIlegible)
	}

	// Decode se detiene al final del primer valor. Lo que haya detrás, salvo
	// espacio en blanco, hace ilegible el cuerpo, como en json.loads.
	if _, err := decodificador.Token(); !errors.Is(err, io.EOF) {
		return nil, errorDeLectura(err, motivoDelCuerpoIlegible)
	}

	envoltorio, esObjeto := raiz.(map[string]any)
	if !esObjeto {
		return nil, errorDeLectura(nil, fmt.Sprintf(motivoDeLaRaiz, describirTipo(raiz)))
	}

	datos := envoltorio[claveDeDatos]
	if datos == nil {
		return []any{}, nil
	}

	return datos, nil
}

// esVacio dice si un valor de la respuesta está vacío en el sentido de J2: la
// lista vacía, el objeto vacío, la cadena vacía o nulo, lo que el if not de
// refs/boe.py descarta en data (282, 362, 459 y 518) y en las materias, las
// notas y las referencias (528, 539, 547 y 567). El cero y el falso no lo están:
// data-model.md §3.1 no los cuenta.
func esVacio(valor any) bool {
	switch valor := valor.(type) {
	case nil:
		return true
	case string:
		return valor == ""
	case []any:
		return len(valor) == 0
	case map[string]any:
		return len(valor) == 0
	default:
		return false
	}
}

// primerElemento es el objeto de los metadatos o del análisis: el primer
// elemento de data si es una lista, o data si llega suelto, como item = data[0]
// if isinstance(data, list) and data else data (refs/boe.py 195, 463 y 522; J5).
// Si lo que resulta no es un objeto, la respuesta es ilegible. Quien la llama ya
// ha descartado data vacío, que en esos dos verbos es «no encontrado» (J3).
func primerElemento(datos any) (map[string]any, error) {
	elemento := datos
	if lista, esLista := datos.([]any); esLista && len(lista) > 0 {
		elemento = lista[0]
	}

	objeto, esObjeto := elemento.(map[string]any)
	if !esObjeto {
		return nil, errorDeLectura(nil, fmt.Sprintf(motivoDelPrimerElemento, describirTipo(elemento)))
	}

	return objeto, nil
}

// comoLista es _ensure_list (refs/boe.py 218-225), con la que un valor suelto
// donde la API suele entregar una lista se lee como lista de un elemento (J4,
// FR-070): nulo es la lista vacía, una lista va tal cual —la misma, sin
// copiarla— y cualquier otro valor es una lista de uno. No descarta los valores
// vacíos: donde refs/boe.py lo hace, lo hace quien la llama (desenvolver).
func comoLista(valor any) []any {
	switch valor := valor.(type) {
	case nil:
		return []any{}
	case []any:
		return valor
	default:
		return []any{valor}
	}
}

// resultadosDeBusqueda son los resultados de buscar en el orden de la fuente:
// data como lista, o un resultado suelto como lista de uno (J4, FR-070:
// refs/boe.py 286 toma cada clave del objeto como un elemento), sin los
// elementos que no son objeto (J8, 287). Quien la llama ya ha resuelto data
// vacío, que en buscar es la lista vacía (J3).
func resultadosDeBusqueda(datos any) []map[string]any {
	return objetosDe(comoLista(datos))
}

// bloquesDelIndice son los bloques del índice en el orden de la fuente (J6,
// refs/boe.py 365-369): los que anida el primer elemento de data si es un objeto
// con la clave bloque, y data en otro caso. data, el elemento que los anida y
// su bloque pueden llegar sueltos y se leen como lista de uno (J4, FR-070:
// refs/boe.py 385 toma cada clave del objeto como un elemento); y los elementos
// que no son objeto se saltan (J8, 386).
func bloquesDelIndice(datos any) []map[string]any {
	elementos := comoLista(datos)
	if len(elementos) == 0 {
		return []map[string]any{}
	}

	primero, esObjeto := elementos[0].(map[string]any)
	if anidados, anida := primero[claveDeBloques]; esObjeto && anida {
		elementos = comoLista(anidados)
	}

	return objetosDe(elementos)
}

// materiasDe son las materias de un análisis (refs/boe.py 527-535), abiertas de
// su envoltorio materia con desenvolver (J7, 531). Las que no son objeto se
// conservan, porque aportan su texto (535).
func materiasDe(analisis map[string]any) []any {
	return desenvolver(analisis[claveDeMaterias], envoltorioDeMateria)
}

// notasDe son las notas de un análisis (refs/boe.py 538-541): una cadena, una
// lista de cadenas o su envoltorio nota (J7, 540), abiertas con desenvolver. El
// envoltorio se abre también dentro de la lista, que es como las entrega la API
// —notas: [{nota: [...]}] en el análisis grabado de la LPAC— igual que las
// materias y las referencias. Leer cada nota como texto es de quien la llama,
// con textoDelValor.
func notasDe(analisis map[string]any) []any {
	return desenvolver(analisis[claveDeNotas], envoltorioDeNota)
}

// referenciasAnterioresDe son las referencias de un análisis a las normas que
// esta modifica o deroga (refs/boe.py 544-564).
func referenciasAnterioresDe(analisis map[string]any) []map[string]any {
	return referenciasDe(analisis, claveDeAnteriores, envoltorioDeAnterior)
}

// referenciasPosterioresDe son las referencias de un análisis a las normas que
// la modifican (refs/boe.py 566-581).
func referenciasPosterioresDe(analisis map[string]any) []map[string]any {
	return referenciasDe(analisis, claveDePosteriores, envoltorioDePosterior)
}

// referenciasDe es una de las dos listas de referencias de un análisis: ninguna
// si referencias no es un objeto (refs/boe.py 545); si lo es, los grupos de la
// lista abiertos de su envoltorio con desenvolver (J7, 550-554 y 570-574), sin
// las referencias que no son objeto (J8, 556 y 576).
func referenciasDe(analisis map[string]any, lista, envoltorio string) []map[string]any {
	referencias, esObjeto := analisis[claveDeReferencias].(map[string]any)
	if !esObjeto {
		return []map[string]any{}
	}

	return objetosDe(desenvolver(referencias[lista], envoltorio))
}

// desenvolver abre una lista cuyos elementos pueden llegar envueltos, como
// m.get("materia", m) (refs/boe.py 531, 540, 551 y 571; J7): un elemento que es
// un objeto con la clave del envoltorio aporta su contenido, y cualquier otro se
// aporta a sí mismo, en el orden de la fuente. La lista y el contenido de cada
// envoltorio pueden llegar sueltos y se leen como lista de uno (J4, FR-070); y
// una lista vacía no aporta nada, como los if materias, if notas, if anteriores
// e if posteriores de refs/boe.py (528, 539, 547 y 567).
func desenvolver(valor any, envoltorio string) []any {
	elementos := []any{}
	if esVacio(valor) {
		return elementos
	}

	for _, elemento := range comoLista(valor) {
		objeto, esObjeto := elemento.(map[string]any)
		if contenido, envuelto := objeto[envoltorio]; esObjeto && envuelto {
			elementos = append(elementos, comoLista(contenido)...)

			continue
		}

		elementos = append(elementos, elemento)
	}

	return elementos
}

// objetosDe son los elementos de una lista que son objetos, en su orden: los
// demás se saltan, como los if isinstance(item, dict) de refs/boe.py (287, 386,
// 556 y 576; J8).
func objetosDe(elementos []any) []map[string]any {
	objetos := make([]map[string]any, 0, len(elementos))
	for _, elemento := range elementos {
		if objeto, esObjeto := elemento.(map[string]any); esObjeto {
			objetos = append(objetos, objeto)
		}
	}

	return objetos
}

// textoDe es el texto de un campo de un objeto de la respuesta: textoDelValor
// de lo que haya en el campo, que si falta es nulo (J9, FR-016).
func textoDe(objeto map[string]any, campo string) (string, error) {
	return textoDelValor(objeto[campo], campo)
}

// textoDelValor lee como texto el valor que la respuesta da en el campo que
// nombra (J9, FR-016, research.md D7): una cadena es su valor, y nulo —también
// el de un campo ausente— es la cadena vacía, nunca el marcador ? que imprime
// refs/boe.py (289-297, 387, 473-482, 533 y 559-561). Cualquier otro tipo
// —número, booleano, lista u objeto— es ilegible nombrando el campo: el repr de
// Python que imprimiría refs/boe.py no es texto de la fuente, y convertirlo a su
// literal enmascararía un cambio de forma de la API.
func textoDelValor(valor any, campo string) (string, error) {
	switch valor := valor.(type) {
	case string:
		return valor, nil
	case nil:
		return "", nil
	default:
		return "", errorDeLectura(nil, fmt.Sprintf(motivoDelCampo, campo, describirTipo(valor)))
	}
}

// textoDelObjetoDe es el texto de un subcampo del objeto que la respuesta da en
// un campo, y la cadena vacía si en el campo no hay un objeto: rango.texto y
// estado_consolidacion.texto en buscar, donde lo que no es objeto da ? en
// refs/boe.py (288-291; data-model.md §2.3), y estado_consolidacion.codigo en
// metadatos (467). El subcampo se lee con textoDelValor y el error lo nombra
// como campo.subcampo.
func textoDelObjetoDe(objeto map[string]any, campo, subcampo string) (string, error) {
	interior, esObjeto := objeto[campo].(map[string]any)
	if !esObjeto {
		return "", nil
	}

	return textoDelValor(interior[subcampo], campo+"."+subcampo)
}

// textoDelObjetoOCadenaDe es textoDelObjetoDe si en el campo hay un objeto y,
// si no, textoDe del propio campo, de modo que una cadena es la cadena y nulo o
// ausente, la cadena vacía: rango y estado_consolidacion en metadatos y relacion
// en las referencias del análisis, el str(x) de lo que no es objeto en
// refs/boe.py (465-469, 557-560 y 577-580; data-model.md §2.5 y §2.6).
func textoDelObjetoOCadenaDe(objeto map[string]any, campo, subcampo string) (string, error) {
	if _, esObjeto := objeto[campo].(map[string]any); esObjeto {
		return textoDelObjetoDe(objeto, campo, subcampo)
	}

	return textoDe(objeto, campo)
}

// describirTipo nombra, para el mensaje, el tipo de un valor de la respuesta tal
// como lo entrega el decodificador.
func describirTipo(valor any) string {
	switch valor.(type) {
	case nil:
		return "nulo"
	case string:
		return "una cadena"
	case json.Number:
		return "un número"
	case bool:
		return "un booleano"
	case []any:
		return "una lista"
	case map[string]any:
		return "un objeto"
	default:
		return fmt.Sprintf("un valor de tipo %T", valor)
	}
}

// errorDeLectura es el fallo de una lectura: «fuente no disponible», código 4
// (contrato errores-y-codigos, filas 16 y 17), sin dirección ni instante,
// porque la lectura no los conoce. Quien pidió lo envuelve con
// errorDeFuenteNoDisponible, su dirección y su instante: como este error declara
// su clase, su mensaje sigue en el del envoltorio, detrás de la dirección, y el
// sobre de fallo toma la dirección y el instante del de fuera.
func errorDeLectura(causa error, motivo string) *Error {
	return errorDeFuenteNoDisponible("", time.Time{}, causa, motivo)
}
