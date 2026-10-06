package ids

import (
	"fmt"
	"strings"
)

// La forma del ECLI de una resolución española y sus cotas (H23, FR-003;
// data-model §1): cinco partes separadas por dos puntos, sin blancos y en
// mayúsculas. Los mensajes de rechazo nombran la forma y las cotas, así que la
// comprobación y el mensaje no pueden decir cosas distintas.
const (
	// identificadorECLI es lo que el error nombra antes de la entrada.
	identificadorECLI = "el ECLI"
	// formaECLI es la forma entera, la que nombran los mensajes de rechazo.
	formaECLI = "ECLI:ES:<órgano>:<año>:<número>"
	// prefijoECLI es la primera parte de todo ECLI.
	prefijoECLI = "ECLI"
	// paisECLI es el código de país de España, el único que se reconoce.
	paisECLI = "ES"
	// separadorECLI va entre una parte y la siguiente.
	separadorECLI = ":"
	// partesECLI son las de la forma: el prefijo, el país, el órgano, el año y
	// el número.
	partesECLI = 5
	// maximoDelOrgano y maximoDelNumero acotan el código del órgano y el
	// número de orden, que tienen al menos un carácter.
	maximoDelOrgano = 7
	maximoDelNumero = 25
	// cifrasDelAnio son las del año, en el ECLI y en el ROJ.
	cifrasDelAnio = 4
)

// ECLI es el identificador europeo de jurisprudencia de una resolución
// española, ECLI:ES:<órgano>:<año>:<número>. Es inmutable y solo se construye
// analizando, con AnalizarECLI. El valor cero no es ningún ECLI y se escribe
// como la cadena vacía.
type ECLI struct {
	// organo es el código del órgano que dictó la resolución.
	organo string
	// anio son las cuatro cifras del año.
	anio string
	// numero es el número de orden, con sus letras y sus puntos.
	numero string
}

// AnalizarECLI analiza el ECLI de una resolución española: ECLI, el país ES,
// el código del órgano —de 1 a 7 letras mayúsculas o cifras ASCII, la primera
// una letra—, el año de cuatro cifras y el número de orden —de 1 a 25 letras
// mayúsculas, cifras o puntos—, separados por dos puntos. La entrada no se
// recorta ni se pasa a mayúsculas, y el órgano no se busca en ninguna lista.
// Fuera de esa forma devuelve un error de clase «argumentos» que nombra la
// entrada y dice qué tiene de malo; el de un ECLI de otro país dice que no es
// español (H23, FR-003).
func AnalizarECLI(entrada string) (ECLI, error) {
	if entrada == "" {
		return ECLI{}, entradaInvalida(identificadorECLI, entrada, "está vacío y la forma es "+formaECLI)
	}

	partes := strings.Split(entrada, separadorECLI)
	if len(partes) != partesECLI {
		return ECLI{}, entradaInvalida(identificadorECLI, entrada, fmt.Sprintf(
			"tiene %d partes separadas por dos puntos y la forma %s tiene %d", len(partes), formaECLI, partesECLI))
	}

	prefijo, pais, organo, anio, numero := partes[0], partes[1], partes[2], partes[3], partes[4]

	if prefijo != prefijoECLI {
		return ECLI{}, entradaInvalida(identificadorECLI, entrada, fmt.Sprintf(
			"empieza por %q y la forma %s empieza por %s", prefijo, formaECLI, prefijoECLI))
	}

	if err := comprobarPais(entrada, pais); err != nil {
		return ECLI{}, err
	}

	if err := comprobarOrgano(entrada, organo); err != nil {
		return ECLI{}, err
	}

	if err := comprobarAnio(identificadorECLI, entrada, anio); err != nil {
		return ECLI{}, err
	}

	if err := comprobarNumeroDeOrden(entrada, numero); err != nil {
		return ECLI{}, err
	}

	return ECLI{organo: organo, anio: anio, numero: numero}, nil
}

// String devuelve el ECLI como se escribió: analizar lo que devuelve da el
// mismo ECLI.
func (e ECLI) String() string {
	if e == (ECLI{}) {
		return ""
	}

	return strings.Join([]string{prefijoECLI, paisECLI, e.organo, e.anio, e.numero}, separadorECLI)
}

// Organo devuelve el código del órgano que dictó la resolución, la tercera
// parte del ECLI: TC en las del Tribunal Constitucional.
func (e ECLI) Organo() string {
	return e.organo
}

// comprobarPais exige el código de país de España. El de otro país bien
// formado, dos letras mayúsculas, se rechaza diciendo que el ECLI no es
// español; cualquier otra cosa, frente a la forma.
func comprobarPais(entrada, pais string) error {
	if pais == paisECLI {
		return nil
	}

	if _, hayAjeno := caracterAjeno(pais, esMayuscula); len(pais) == len(paisECLI) && !hayAjeno {
		return entradaInvalida(identificadorECLI, entrada, fmt.Sprintf(
			"el código de país es %q y no %s: no es español", pais, paisECLI))
	}

	return entradaInvalida(identificadorECLI, entrada, fmt.Sprintf(
		"el código de país es %q y la forma %s lleva %s", pais, formaECLI, paisECLI))
}

// comprobarOrgano exige que el código del órgano sean de 1 a 7 letras
// mayúsculas o cifras ASCII y que empiece por una letra.
func comprobarOrgano(entrada, organo string) error {
	if ajeno, hayAjeno := caracterAjeno(organo, esMayusculaOCifra); hayAjeno {
		return entradaInvalida(identificadorECLI, entrada, fmt.Sprintf(
			"el órgano %q lleva %q, que no es una letra mayúscula de la A a la Z ni una cifra", organo, ajeno))
	}

	if organo == "" || len(organo) > maximoDelOrgano {
		return entradaInvalida(identificadorECLI, entrada, fmt.Sprintf(
			"el órgano %q tiene %d caracteres y son de 1 a %d", organo, len(organo), maximoDelOrgano))
	}

	if esCifra(organo[0]) {
		return entradaInvalida(identificadorECLI, entrada, fmt.Sprintf(
			"el órgano %q empieza por una cifra y no por una letra", organo))
	}

	return nil
}

// comprobarNumeroDeOrden exige que el número de orden sean de 1 a 25 letras
// mayúsculas, cifras ASCII o puntos.
func comprobarNumeroDeOrden(entrada, numero string) error {
	if ajeno, hayAjeno := caracterAjeno(numero, esDelNumeroDeOrden); hayAjeno {
		return entradaInvalida(identificadorECLI, entrada, fmt.Sprintf(
			"el número %q lleva %q, que no es una letra mayúscula de la A a la Z, una cifra ni un punto", numero, ajeno))
	}

	if numero == "" || len(numero) > maximoDelNumero {
		return entradaInvalida(identificadorECLI, entrada, fmt.Sprintf(
			"el número %q tiene %d caracteres y son de 1 a %d", numero, len(numero), maximoDelNumero))
	}

	return nil
}

// comprobarAnio exige que el año sean cuatro cifras ASCII. Los rechazos
// nombran la entrada entera, de la que anio es un tramo. Lo comparten el
// análisis del ECLI y el del ROJ.
func comprobarAnio(identificador, entrada, anio string) error {
	if ajeno, hayAjeno := caracterAjeno(anio, esCifra); hayAjeno {
		return entradaInvalida(identificador, entrada, fmt.Sprintf(
			"el año %q lleva %q, que no es una cifra", anio, ajeno))
	}

	if len(anio) != cifrasDelAnio {
		return entradaInvalida(identificador, entrada, fmt.Sprintf(
			"el año %q tiene %d cifras y son %d", anio, len(anio), cifrasDelAnio))
	}

	return nil
}

// caracterAjeno devuelve el primer carácter del texto que admite no acepta, y
// si lo hay. Mira bytes y no runas, como comprobarCifras: ninguna letra ni
// cifra de otra escritura forma parte de un identificador. Como lo que se
// admite es siempre ASCII, el byte rechazado está al principio de un carácter,
// que se devuelve entero aunque ocupe varios bytes. Lo comparten el análisis
// del ECLI y el del ROJ.
func caracterAjeno(texto string, admite func(byte) bool) (ajeno string, hay bool) {
	for indice := range len(texto) {
		if !admite(texto[indice]) {
			return primerCaracter(texto[indice:]), true
		}
	}

	return "", false
}

// esMayuscula dice si el byte es una letra mayúscula ASCII, de la A a la Z.
func esMayuscula(b byte) bool {
	return 'A' <= b && b <= 'Z'
}

// esMayusculaOCifra dice si el byte puede ir en el código del órgano.
func esMayusculaOCifra(b byte) bool {
	return esMayuscula(b) || esCifra(b)
}

// esDelNumeroDeOrden dice si el byte puede ir en el número de orden: una
// letra mayúscula, una cifra o un punto.
func esDelNumeroDeOrden(b byte) bool {
	return esMayusculaOCifra(b) || b == '.'
}
