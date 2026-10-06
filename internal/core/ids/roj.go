package ids

import (
	"fmt"
	"strings"
)

// La forma del ROJ, el identificador nacional de una resolución en el CENDOJ
// (H23, FR-004; data-model §1): unas siglas, un espacio, el número, una barra
// y el año. Los mensajes de rechazo nombran la forma.
const (
	// identificadorROJ es lo que el error nombra antes de la entrada.
	identificadorROJ = "el ROJ"
	// formaROJ es la forma entera, la que nombran los mensajes de rechazo.
	formaROJ = "<siglas> <número>/<año>"
	// espacioROJ va entre las palabras de las siglas y entre las siglas y el
	// número: un solo carácter U+0020, ningún otro blanco.
	espacioROJ = " "
	// barraROJ va entre el número y el año.
	barraROJ = "/"
)

// ROJ es el identificador nacional de una resolución, <siglas>
// <número>/<año>: STS 3144/2023. Es inmutable y solo se construye analizando,
// con AnalizarROJ. El valor cero no es ningún ROJ y se escribe como la cadena
// vacía.
type ROJ struct {
	// siglas son la palabra o las palabras que van delante del número.
	siglas string
	// numero son las cifras del número, con sus ceros por delante.
	numero string
	// anio son las cuatro cifras del año.
	anio string
}

// AnalizarROJ analiza un ROJ: las siglas —una o más palabras de letras
// mayúsculas ASCII separadas por un espacio—, un espacio, el número —una o
// más cifras ASCII—, una barra y el año de cuatro cifras, sin nada delante ni
// detrás. La entrada no se recorta ni se pasa a mayúsculas, y las siglas no se
// buscan en ninguna lista de órganos. Fuera de esa forma devuelve un error de
// clase «argumentos» que nombra la entrada y dice qué tiene de malo (H23,
// FR-004).
func AnalizarROJ(entrada string) (ROJ, error) {
	if entrada == "" {
		return ROJ{}, entradaInvalida(identificadorROJ, entrada, "está vacío y la forma es "+formaROJ)
	}

	corte := strings.LastIndex(entrada, espacioROJ)
	if corte < 0 {
		return ROJ{}, entradaInvalida(identificadorROJ, entrada, fmt.Sprintf(
			"no lleva ningún espacio y la forma %s separa con uno las siglas del número", formaROJ))
	}

	siglas, resto := entrada[:corte], entrada[corte+len(espacioROJ):]

	numero, anio, conBarra := strings.Cut(resto, barraROJ)
	if !conBarra {
		return ROJ{}, entradaInvalida(identificadorROJ, entrada, fmt.Sprintf(
			"tras el último espacio va %q y la forma %s lleva ahí el número, una barra y el año", resto, formaROJ))
	}

	if err := comprobarSiglas(entrada, siglas); err != nil {
		return ROJ{}, err
	}

	if err := comprobarNumeroDelROJ(entrada, numero); err != nil {
		return ROJ{}, err
	}

	if err := comprobarAnio(identificadorROJ, entrada, anio); err != nil {
		return ROJ{}, err
	}

	return ROJ{siglas: siglas, numero: numero, anio: anio}, nil
}

// String devuelve el ROJ como se escribió: analizar lo que devuelve da el
// mismo ROJ.
func (r ROJ) String() string {
	if r == (ROJ{}) {
		return ""
	}

	return r.siglas + espacioROJ + r.numero + barraROJ + r.anio
}

// comprobarSiglas exige que las siglas sean una o más palabras de letras
// mayúsculas ASCII, separada cada una de la siguiente por un solo espacio.
func comprobarSiglas(entrada, siglas string) error {
	for palabra := range strings.SplitSeq(siglas, espacioROJ) {
		if palabra == "" {
			return entradaInvalida(identificadorROJ, entrada, fmt.Sprintf(
				"las siglas %q no son una o más palabras separadas por un solo espacio", siglas))
		}

		if ajeno, hayAjeno := caracterAjeno(palabra, esMayuscula); hayAjeno {
			return entradaInvalida(identificadorROJ, entrada, fmt.Sprintf(
				"las siglas %q llevan %q, que no es una letra mayúscula de la A a la Z", siglas, ajeno))
		}
	}

	return nil
}

// comprobarNumeroDelROJ exige que el número sean una o más cifras ASCII.
func comprobarNumeroDelROJ(entrada, numero string) error {
	if numero == "" {
		return entradaInvalida(identificadorROJ, entrada, "el número está vacío y son una o más cifras")
	}

	if ajeno, hayAjeno := caracterAjeno(numero, esCifra); hayAjeno {
		return entradaInvalida(identificadorROJ, entrada, fmt.Sprintf(
			"el número %q lleva %q, que no es una cifra", numero, ajeno))
	}

	return nil
}
