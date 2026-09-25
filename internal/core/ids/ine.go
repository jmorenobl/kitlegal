package ids

import (
	"fmt"
	"unicode/utf8"
)

// Las formas del código INE de municipio y sus límites (data-model §1.1,
// contrato de identificadores §2). Los mensajes de rechazo nombran la forma,
// así que la comprobación y el mensaje no pueden decir cosas distintas.
const (
	// formaINE es el código: PP la provincia y MMM el municipio dentro de ella.
	formaINE = "PPMMM"
	// formaINEConDigito es el código seguido del dígito de control que declara
	// quien lo escribe.
	formaINEConDigito = "PPMMMD"
	// cifrasDeProvincia es la longitud de PP.
	cifrasDeProvincia = 2
	// cifrasINE es la longitud del código, sin el dígito.
	cifrasINE = len(formaINE)
	// primeraProvincia y ultimaProvincia acotan PP. Se comparan como texto, lo
	// que equivale a compararlas como números porque las dos partes tienen dos
	// cifras ASCII.
	primeraProvincia = "01"
	ultimaProvincia  = "52"
	// primerMunicipio y ultimoMunicipio acotan MMM, que con tres cifras no
	// puede pasar del último: solo queda fuera el 000.
	primerMunicipio = "001"
	ultimoMunicipio = "999"
)

// CodigoINE es el código INE de un municipio, PPMMM. Es inmutable y solo se
// construye analizando, con AnalizarCodigoINE o AnalizarCodigoINEConDigito, o
// desde un DIR3 analizado. El valor cero no es ningún código y se escribe como
// la cadena vacía.
type CodigoINE struct {
	// provincia son las dos cifras de PP.
	provincia string
	// municipio son las tres cifras de MMM.
	municipio string
}

// AnalizarCodigoINE analiza un código INE de cinco cifras ASCII, con la
// provincia entre 01 y 52 y el municipio entre 001 y 999. La entrada no se
// recorta ni se completa: los ceros por delante son parte del código, y un
// espacio o un salto de línea la hacen inválida. Fuera de esa forma devuelve
// un error de clase «argumentos» que nombra la entrada y dice qué tiene de
// malo (FR-030, FR-033).
func AnalizarCodigoINE(entrada string) (CodigoINE, error) {
	if err := comprobarForma(identificadorINE, entrada, formaINE); err != nil {
		return CodigoINE{}, err
	}

	return codigoINE(identificadorINE, entrada, entrada)
}

// AnalizarCodigoINEConDigito analiza un código INE de seis cifras ASCII: las
// cinco del código, con las mismas reglas que AnalizarCodigoINE, y el dígito de
// control que declara quien lo escribe, que se devuelve como el carácter de su
// cifra, de '0' a '9'. Que ese dígito sea el oficial no lo decide el análisis:
// lo compara ComprobarDigito con el de la relación del INE (research.md D9).
func AnalizarCodigoINEConDigito(entrada string) (codigo CodigoINE, digito byte, err error) {
	err = comprobarForma(identificadorINEConDigito, entrada, formaINEConDigito)
	if err != nil {
		return CodigoINE{}, 0, err
	}

	codigo, err = codigoINE(identificadorINEConDigito, entrada, entrada[:cifrasINE])
	if err != nil {
		return CodigoINE{}, 0, err
	}

	return codigo, entrada[cifrasINE], nil
}

// String devuelve las cinco cifras del código, con sus ceros por delante:
// analizar lo que devuelve da el mismo código (FR-032).
func (c CodigoINE) String() string {
	return c.provincia + c.municipio
}

// Provincia devuelve las dos cifras de la provincia del código.
func (c CodigoINE) Provincia() string {
	return c.provincia
}

// ComprobarDigito compara el dígito de control declarado para el código con el
// oficial de la relación del INE, los dos como el carácter de su cifra. Si
// coinciden devuelve nil; si no, un error de clase «argumentos» que nombra el
// código con el dígito recibido y dice cuál es cada uno de los dos. No calcula
// nada: quien resuelve comprueba antes que el municipio está en la relación,
// porque sin él no hay dígito oficial con el que comparar (contrato de
// identificadores §3).
func (c CodigoINE) ComprobarDigito(declarado, oficial byte) error {
	if declarado == oficial {
		return nil
	}

	return entradaInvalida(identificadorINEConDigito, c.String()+caracter(declarado), fmt.Sprintf(
		"el dígito de control recibido es %q y el oficial es %q", caracter(declarado), caracter(oficial)))
}

// comprobarForma exige que la entrada sea solo cifras ASCII y tantas como
// letras tiene la forma.
func comprobarForma(identificador, entrada, forma string) error {
	if err := comprobarCifras(identificador, entrada, entrada); err != nil {
		return err
	}

	if len(entrada) != len(forma) {
		return entradaInvalida(identificador, entrada, fmt.Sprintf(
			"tiene %d cifras y la forma %s tiene %d", len(entrada), forma, len(forma)))
	}

	return nil
}

// comprobarCifras exige que todos los bytes de cifras, el tramo de la entrada
// que tiene que serlo, sean cifras ASCII. Mira bytes y no runas a propósito:
// una cifra de otra escritura no forma parte de ningún identificador. El
// rechazo nombra el primer carácter que no lo es, entero aunque ocupe varios
// bytes. Lo comparten el análisis del código INE y el del DIR3.
func comprobarCifras(identificador, entrada, cifras string) error {
	for indice := range len(cifras) {
		if esCifra(cifras[indice]) {
			continue
		}

		// Los bytes anteriores son cifras ASCII, así que el índice está al
		// principio de un carácter.
		return entradaInvalida(identificador, entrada, fmt.Sprintf(
			"%q no es una cifra", primerCaracter(cifras[indice:])))
	}

	return nil
}

// primerCaracter devuelve el primer carácter del texto con todos sus bytes, o
// solo el primero si no empieza por un carácter UTF-8 válido, para que el
// mensaje lo muestre tal cual.
func primerCaracter(texto string) string {
	_, medida := utf8.DecodeRuneInString(texto)

	return texto[:medida]
}

// codigoINE construye el código de las cinco cifras ASCII de cifras, que ya
// tienen esa longitud, si la provincia y el municipio están en su rango. Los
// rechazos nombran la entrada entera, de la que cifras es un tramo.
func codigoINE(identificador, entrada, cifras string) (CodigoINE, error) {
	provincia, municipio := cifras[:cifrasDeProvincia], cifras[cifrasDeProvincia:]

	if provincia < primeraProvincia || provincia > ultimaProvincia {
		return CodigoINE{}, entradaInvalida(identificador, entrada, fmt.Sprintf(
			"la provincia %q no está entre %s y %s", provincia, primeraProvincia, ultimaProvincia))
	}

	if municipio < primerMunicipio {
		return CodigoINE{}, entradaInvalida(identificador, entrada, fmt.Sprintf(
			"el municipio %q no está entre %s y %s", municipio, primerMunicipio, ultimoMunicipio))
	}

	return CodigoINE{provincia: provincia, municipio: municipio}, nil
}

// esCifra dice si el byte es una cifra ASCII, del 0 al 9.
func esCifra(b byte) bool {
	return '0' <= b && b <= '9'
}

// caracter escribe un byte tal cual, sin interpretarlo como una runa: el byte
// 0xff sigue siendo 0xff, no la ÿ que daría string(rune(b)).
func caracter(b byte) string {
	return string([]byte{b})
}
