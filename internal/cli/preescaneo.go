package cli

import "strings"

// Preliminar es lo que el kernel sabe de una invocación antes de que exista
// gramática que analizar: tres banderas leídas directamente de la lista de
// argumentos.
//
// Es un valor provisional. En cuanto el análisis de Kong termina bien manda lo
// analizado, y el pre-escaneo se descarta (research.md D25, regla 5).
type Preliminar struct {
	// JSON decide la forma en que se presenta un fallo ocurrido antes de que
	// las globales existan —bandera desconocida y applet no registrado, los dos
	// casos obligatorios de SC-014—: sobre de fallo en la salida estándar si se
	// pidió, y si no, mensaje para la persona en la salida de error con la
	// salida estándar vacía (FR-045).
	JSON bool
	// Verbose decide el nivel del registro de eventos mientras no hay globales,
	// sin perjuicio de que KITLEGAL_LOG tenga prioridad sobre él (research.md
	// D14).
	Verbose bool
	// Ayuda suprime la normalización del verbo por omisión: quien pide la ayuda
	// de un applet no está invocando ningún verbo (research.md D26, regla 3).
	Ayuda bool
}

// terminador separa las banderas de los argumentos: lo que va después ya no es
// una bandera aunque se escriba como tal.
const terminador = "--"

// Los tres nombres largos exactos que el pre-escaneo reconoce, y ninguno más.
// No hay formas cortas ni abreviaturas: una abreviatura que Kong aceptara y el
// pre-escaneo no —o al revés— sería una divergencia silenciosa entre lo que se
// lee aquí y lo que Kong analiza después, y lo estrecho es justamente lo que
// hace este procedimiento comprobable (research.md D25, regla 1).
const (
	banderaJSON    = "--json"
	banderaVerbose = "--verbose"
	banderaAyuda   = "--help"
)

// PreEscanear lee --json, --verbose y --help de la lista de argumentos antes de
// que exista gramática, con un procedimiento deliberadamente estrecho: solo las
// tres formas largas exactas y sus formas con valor booleano explícito
// (--json=false), se detiene en el terminador y cualquier otro token se ignora.
//
// No valida, no falla y no consume: no devuelve error porque cuando se ejecuta
// no hay todavía nada que pueda emitir un fallo, y deja la lista
// que recibe intacta, de modo que quien la pasa se la puede entregar después a
// Kong entera. Un token desconocido, una forma no soportada o un valor que Kong
// rechazaría no encienden ninguna bandera y tampoco detienen el recorrido
// (research.md D25, reglas 1 a 3).
//
// No decide qué se ejecuta. Decide cómo se presenta un fallo anterior a las
// globales, el nivel del registro de eventos y si se suprime la normalización
// del verbo por omisión (research.md D25, regla 4).
func PreEscanear(args []string) Preliminar {
	var previo Preliminar

	for _, arg := range args {
		if arg == terminador {
			break
		}

		nombre, literal, conValor := strings.Cut(arg, "=")

		valor, reconocido := valorBooleano(literal, conValor)
		if !reconocido {
			continue
		}

		switch nombre {
		case banderaJSON:
			previo.JSON = valor
		case banderaVerbose:
			previo.Verbose = valor
		case banderaAyuda:
			previo.Ayuda = valor
		}
	}

	return previo
}

// valorBooleano interpreta la parte que sigue al signo igual. Una bandera sin
// valor vale por lo que afirma, y con valor vale lo que el valor diga.
//
// Los seis literales son los que acepta el decodificador de booleanos de Kong,
// y sin distinguir mayúsculas por el mismo motivo: están copiados de su
// boolMapper (kong@v1.16.1, mapper.go, líneas 296-321), que es donde acabará
// el mismo argumento cuando exista gramática. Que la lista coincida es lo que
// permite al TestPreescaneo de T005 comprobar que el pre-escaneo y el análisis
// dicen lo mismo para toda invocación bien formada; si Kong cambiara la suya,
// ese test lo señalaría.
//
// Un literal fuera de esa lista no se reconoce y la bandera se queda como
// estaba. No es una omisión: Kong lo rechazará al analizar y el fallo saldrá
// como error de argumentos, que es de quien es esa decisión; el pre-escaneo no
// valida.
func valorBooleano(literal string, conValor bool) (valor, reconocido bool) {
	if !conValor {
		return true, true
	}

	switch strings.ToLower(literal) {
	case "true", "1", "yes":
		return true, true
	case "false", "0", "no":
		return false, true
	}

	return false, false
}
