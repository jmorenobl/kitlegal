package cita

import (
	"fmt"
	"strings"
)

// Los tres resultados de la correspondencia entre el ROJ y el ECLI de una
// ficha (H23, FR-023).
const (
	// seCorresponden y noSeCorresponden son los de una ficha a la que alcanza
	// la regla de FR-012: sus dos identificadores son de la única pareja en la
	// que uno se deduce del otro.
	seCorresponden   = "se-corresponden"
	noSeCorresponden = "no-se-corresponden"
	// noSeDeduce es el de cualquier otra: ni se afirma que se corresponden ni
	// que no.
	noSeDeduce = "no-se-deduce"
)

// El hallazgo del cotejo (FR-025; research.md D7).
const (
	// claseDocumentoDistinto es su única clase: el documento no es el pedido.
	claseDocumentoDistinto = "documento-distinto"
	// cruceNumeroDeResolucion es el de quien pide un ROJ cuyo <número>/<año> es
	// el número de resolución del documento.
	cruceNumeroDeResolucion = "numero-de-resolucion"
	// cruceNumeroDelROJ es el de quien pide un número de resolución que es el
	// <número>/<año> del ROJ del documento.
	cruceNumeroDelROJ = "numero-del-roj"
)

// datoFecha es, con las tres formas de una referencia, uno de los datos que
// pueden diferir: la fecha, que se pide con el número.
const datoFecha = "fecha"

// nombresDeLosDatos es cómo nombra la explicación de un hallazgo cada dato
// pedido y el del documento, con su género.
var nombresDeLosDatos = map[string]struct{ pedido, delDocumento string }{
	formaECLI:       {"el ECLI", "el del documento"},
	formaROJ:        {"el ROJ", "el del documento"},
	formaResolucion: {"el número de resolución", "el del documento"},
	datoFecha:       {"la fecha", "la del documento"},
}

// Cotejo es lo que se sabe del documento que trae la persona: el data de cita
// cotejar, con las claves de contracts/applet-cita.md §4. Lo que dice de si el
// documento es el pedido va solo cuando se dio una referencia.
type Cotejo struct {
	// Ficha son los ocho datos leídos del documento.
	Ficha Ficha `json:"ficha"`
	// Correspondencia dice si el ROJ y el ECLI de la ficha se corresponden, o
	// que no se deduce.
	Correspondencia string `json:"correspondencia" jsonschema:"enum=se-corresponden,enum=no-se-corresponden,enum=no-se-deduce"`
	// Pedida es la referencia que se pidió.
	Pedida *Referencia `json:"pedida,omitempty"`
	// EsLaPedida dice si el documento es el de esa referencia.
	EsLaPedida *bool `json:"es_la_pedida,omitempty"`
	// Hallazgos es siempre una lista: vacía, o con un hallazgo si el documento
	// no es el pedido.
	Hallazgos []Hallazgo `json:"hallazgos"`
}

// Hallazgo es lo que el cotejo encuentra cuando el documento no es el pedido.
// No es un fallo: la orden termina bien y lo lleva en su data (docs/ADR/0023).
type Hallazgo struct {
	// Clase es documento-distinto.
	Clase string `json:"clase" jsonschema:"enum=documento-distinto"`
	// Difiere lleva una diferencia por dato pedido que no es el del documento.
	Difiere []Diferencia `json:"difiere" jsonschema:"minItems=1"`
	// Cruce está solo cuando el número pedido es el otro número del documento.
	Cruce string `json:"cruce,omitempty" jsonschema:"enum=numero-de-resolucion,enum=numero-del-roj"`
	// Explicacion lo dice en español, para la persona.
	Explicacion string `json:"explicacion" jsonschema:"minLength=1"`
}

// Diferencia es un dato pedido que no es el del documento.
type Diferencia struct {
	// Dato es ecli, roj, resolucion o fecha.
	Dato string `json:"dato" jsonschema:"enum=ecli,enum=roj,enum=resolucion,enum=fecha"`
	// Pedido es lo que se pidió.
	Pedido string `json:"pedido" jsonschema:"minLength=1"`
	// Documento es lo que lleva la ficha.
	Documento string `json:"documento" jsonschema:"minLength=1"`
}

// Cotejar dice lo que se sabe del documento cuya ficha se ha leído: sus datos
// y si su ROJ y su ECLI se corresponden. Con la referencia que se pidió dice
// además si el documento es el pedido —por su ECLI o por su ROJ, si el de la
// ficha es ese, carácter a carácter; por su número con su fecha, si los dos
// coinciden— y, si no lo es, lleva un hallazgo con cada dato que difiere. Ni
// el órgano ni ningún otro dato de la ficha interviene. Sin referencia, pedida
// es nil, y el cotejo no dice nada de lo pedido ni lleva hallazgo (H23, FR-022
// a FR-025).
func Cotejar(ficha Ficha, pedida *Referencia) Cotejo {
	cotejo := Cotejo{Ficha: ficha, Correspondencia: correspondenciaDe(ficha), Hallazgos: []Hallazgo{}}
	if pedida == nil {
		return cotejo
	}

	// La referencia se copia: lo que después cambie quien la pidió no cambia
	// el cotejo.
	referencia := *pedida
	difiere, cruce := diferencias(ficha, referencia)
	esLaPedida := len(difiere) == 0

	cotejo.Pedida = &referencia
	cotejo.EsLaPedida = &esLaPedida

	if !esLaPedida {
		cotejo.Hallazgos = []Hallazgo{{
			Clase:       claseDocumentoDistinto,
			Difiere:     difiere,
			Cruce:       cruce,
			Explicacion: explicar(ficha, difiere, cruce),
		}}
	}

	return cotejo
}

// correspondenciaDe aplica a la ficha la regla de FR-012, que es de
// internal/core/ids. La regla la alcanza solo si de su ECLI se deduce un ROJ y
// de su ROJ, un ECLI; entonces se corresponden si el ROJ deducido es el de la
// ficha, con el número y el año comparados carácter a carácter.
func correspondenciaDe(ficha Ficha) string {
	delECLI, elECLIEsDeLaPareja := ficha.ecli.ROJ()
	_, elROJEsDeLaPareja := ficha.roj.ECLI()

	switch {
	case !elECLIEsDeLaPareja || !elROJEsDeLaPareja:
		return noSeDeduce
	case delECLI == ficha.roj:
		return seCorresponden
	default:
		return noSeCorresponden
	}
}

// diferencias son los datos pedidos que no son los del documento, y el cruce
// si lo hay. El cruce solo existe cuando el número difiere: si se pidió un ROJ
// y su <número>/<año> es el número de resolución de la ficha, o si se pidió un
// número de resolución y es el <número>/<año> del ROJ de la ficha (FR-025).
// Las formas son tres y solo las construye NuevaReferencia: la que no es ni un
// ECLI ni un ROJ es el número con su fecha.
func diferencias(ficha Ficha, pedida Referencia) (difiere []Diferencia, cruce string) {
	switch pedida.Forma {
	case formaECLI:
		return diferencia(formaECLI, pedida.Valor, ficha.ECLI), ""
	case formaROJ:
		difiere = diferencia(formaROJ, pedida.Valor, ficha.ROJ)
		if len(difiere) > 0 && pedida.roj.Numero() == ficha.Resolucion {
			cruce = cruceNumeroDeResolucion
		}

		return difiere, cruce
	}

	difiere = diferencia(formaResolucion, pedida.Valor, ficha.Resolucion)
	if len(difiere) > 0 && pedida.Valor == ficha.roj.Numero() {
		cruce = cruceNumeroDelROJ
	}

	return append(difiere, diferencia(datoFecha, pedida.Fecha, ficha.Fecha)...), cruce
}

// diferencia es la de un dato, si lo pedido no es lo del documento carácter a
// carácter, sin normalizar; si lo es, ninguna.
func diferencia(dato, pedido, documento string) []Diferencia {
	if pedido == documento {
		return nil
	}

	return []Diferencia{{Dato: dato, Pedido: pedido, Documento: documento}}
}

// explicar escribe el hallazgo para la persona: que el documento no es el
// pedido, cada dato pedido con el del documento y, con cruce, qué es en el
// documento el número que se pidió.
func explicar(ficha Ficha, difiere []Diferencia, cruce string) string {
	partes := make([]string, 0, len(difiere))
	for _, distinto := range difiere {
		nombres := nombresDeLosDatos[distinto.Dato]
		partes = append(partes, fmt.Sprintf("se pidió %s %s y %s es %s",
			nombres.pedido, distinto.Pedido, nombres.delDocumento, distinto.Documento))
	}

	explicacion := "El documento no es el pedido: " + strings.Join(partes, "; ") + "."

	switch cruce {
	case cruceNumeroDeResolucion:
		explicacion += fmt.Sprintf(" %s es el número de resolución del documento, no su ROJ.", ficha.Resolucion)
	case cruceNumeroDelROJ:
		explicacion += fmt.Sprintf(" %s es el número del ROJ del documento (%s), no su número de resolución.",
			ficha.roj.Numero(), ficha.ROJ)
	}

	return explicacion
}
