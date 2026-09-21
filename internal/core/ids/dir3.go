package ids

import "fmt"

// La forma del código DIR3 de un ayuntamiento (data-model §1.2, contrato de
// identificadores §2): la letra L de la Administración local, el tipo de
// entidad 01 de un ayuntamiento, su código INE y su dígito de control.
const (
	// formaDIR3 es la forma entera, la que nombran los mensajes de rechazo.
	formaDIR3 = "L01PPMMMD"
	// prefijoDeAyuntamiento es como se escribe el principio de todo DIR3 de
	// ayuntamiento, con la letra ya normalizada a mayúscula.
	prefijoDeAyuntamiento = "L01"
	// tipoDeAyuntamiento son las dos cifras que siguen a la letra.
	tipoDeAyuntamiento = "01"
	// cifrasDIR3 son las que siguen a la letra: el tipo, el código y el dígito.
	cifrasDIR3 = len(formaDIR3) - 1
)

// DIR3 es el código DIR3 de un ayuntamiento, L01PPMMMD. Es inmutable y solo se
// construye analizando, con AnalizarDIR3, o componiéndolo desde su código INE,
// con DIR3DeAyuntamiento. El valor cero no es ningún DIR3 y se escribe como la
// cadena vacía.
type DIR3 struct {
	// codigo es el código INE del municipio, PPMMM.
	codigo CodigoINE
	// digito es el carácter de la cifra del dígito de control, D.
	digito byte
}

// AnalizarDIR3 analiza el código DIR3 de un ayuntamiento: la letra L, en
// mayúscula o en minúscula, seguida de ocho cifras ASCII, que son el tipo de
// entidad 01, un código INE con las reglas de AnalizarCodigoINE y su dígito de
// control. La letra se normaliza a mayúscula; nada más se recorta ni se
// completa. Fuera de esa forma devuelve un error de clase «argumentos» que
// nombra la entrada y dice qué tiene de malo (FR-031, FR-033).
func AnalizarDIR3(entrada string) (DIR3, error) {
	if entrada == "" {
		return DIR3{}, entradaInvalida(identificadorDIR3, entrada, "está vacío y la forma es "+formaDIR3)
	}

	if entrada[0] != 'L' && entrada[0] != 'l' {
		return DIR3{}, entradaInvalida(identificadorDIR3, entrada, fmt.Sprintf(
			"empieza por %q y la forma %s empieza por L", primerCaracter(entrada), formaDIR3))
	}

	cifras := entrada[1:]
	if err := comprobarCifras(identificadorDIR3, entrada, cifras); err != nil {
		return DIR3{}, err
	}

	if len(cifras) != cifrasDIR3 {
		return DIR3{}, entradaInvalida(identificadorDIR3, entrada, fmt.Sprintf(
			"tras la L tiene %d cifras y la forma %s tiene %d", len(cifras), formaDIR3, cifrasDIR3))
	}

	tipo, codigoConDigito := cifras[:len(tipoDeAyuntamiento)], cifras[len(tipoDeAyuntamiento):]
	if tipo != tipoDeAyuntamiento {
		return DIR3{}, entradaInvalida(identificadorDIR3, entrada, fmt.Sprintf(
			"tras la L va %q y la forma %s lleva %s, la de un ayuntamiento", tipo, formaDIR3, tipoDeAyuntamiento))
	}

	codigo, err := codigoINE(identificadorDIR3, entrada, codigoConDigito[:cifrasINE])
	if err != nil {
		return DIR3{}, err
	}

	return DIR3{codigo: codigo, digito: codigoConDigito[cifrasINE]}, nil
}

// DIR3DeAyuntamiento compone el DIR3 del ayuntamiento de un municipio con su
// código INE y su dígito de control, que es el carácter de su cifra, de '0' a
// '9', como lo dan AnalizarCodigoINEConDigito, Digito y la relación del INE.
// Es la única forma de componer uno desde el dominio, y la inversa de
// CodigoINE y Digito: el código y el dígito del resultado son los recibidos
// (FR-031). No comprueba el dígito, que no es suyo: quien lo recibe de fuera lo
// ha comparado antes con el oficial.
func DIR3DeAyuntamiento(c CodigoINE, digito byte) DIR3 {
	return DIR3{codigo: c, digito: digito}
}

// String devuelve el DIR3 con la letra en mayúscula: analizar lo que devuelve
// da el mismo DIR3 (FR-032).
func (d DIR3) String() string {
	if d == (DIR3{}) {
		return ""
	}

	return prefijoDeAyuntamiento + d.codigo.String() + caracter(d.digito)
}

// CodigoINE devuelve el código INE del municipio del ayuntamiento.
func (d DIR3) CodigoINE() CodigoINE {
	return d.codigo
}

// Digito devuelve el dígito de control del DIR3, el carácter de su cifra.
func (d DIR3) Digito() byte {
	return d.digito
}
