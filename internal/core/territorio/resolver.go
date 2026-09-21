package territorio

import (
	"strings"
	"unicode"

	"github.com/jmorenobl/kitlegal/internal/core/ids"
)

// cifrasConDigito es la longitud del código INE seguido de su dígito de
// control, PPMMMD.
const cifrasConDigito = len("PPMMMD")

// Resolver lee una consulta y devuelve el territorio del municipio que nombra
// (data-model §2.6). La forma de la entrada decide qué es:
//
//   - vacía, o hecha solo de espacios y separadores, no nombra nada: error de
//     clase «argumentos»;
//   - sin ninguna letra, es un código INE de cinco cifras, o de seis con su
//     dígito de control, y la gramática de internal/core/ids rechaza con clase
//     «argumentos» lo que no llega a serlo —cifras de más o de menos, la
//     provincia fuera de 01-52, el municipio 000, algo que no es una cifra—.
//     Un código bien formado se busca primero en la relación —si no está, es
//     «no encontrado»— y solo después se compara su dígito con el oficial
//     —si no coincide, «argumentos»—;
//   - con alguna letra, es un nombre, y se busca por su forma plegada: si no
//     es la de ningún municipio, «no encontrado»; si es la de varios,
//     «argumentos» con todos los candidatos, sin elegir ninguno.
//
// Todo error nombra la entrada y es de una de esas dos clases; ninguno más
// (FR-010 a FR-016). Resolver no lee nada: todo está en el registro.
func (r *Registro) Resolver(entrada string) (Territorio, error) {
	plegada := Plegar(entrada)

	switch {
	case plegada == "":
		return Territorio{}, consultaVacia(entrada)
	case !strings.ContainsFunc(plegada, unicode.IsLetter):
		return r.resolverCodigo(entrada)
	default:
		return r.resolverNombre(entrada, plegada)
	}
}

// resolverCodigo resuelve una entrada que se lee como código INE: primero «¿está
// en la relación?» y después el dígito, que sin municipio no hay dígito oficial
// con el que comparar (research.md D9).
func (r *Registro) resolverCodigo(entrada string) (Territorio, error) {
	codigo, digito, conDigito, err := analizarCodigo(entrada)
	if err != nil {
		return Territorio{}, err
	}

	municipio, esta := r.municipio(codigo)
	if !esta {
		return Territorio{}, codigoFueraDeLaRelacion(entrada)
	}

	if conDigito {
		if err := codigo.ComprobarDigito(digito, municipio.digito[0]); err != nil {
			return Territorio{}, err
		}
	}

	return r.territorioDe(municipio), nil
}

// analizarCodigo analiza la entrada con la gramática del código INE: con seis
// caracteres, la del código con su dígito; con cualquier otra longitud, la del
// código de cinco cifras, cuyo rechazo dice cuántas tiene la entrada y cuántas
// tiene la forma.
func analizarCodigo(entrada string) (codigo ids.CodigoINE, digito byte, conDigito bool, err error) {
	if len(entrada) == cifrasConDigito {
		codigo, digito, err = ids.AnalizarCodigoINEConDigito(entrada)

		return codigo, digito, true, err
	}

	codigo, err = ids.AnalizarCodigoINE(entrada)

	return codigo, 0, false, err
}

// resolverNombre resuelve una entrada que se lee como nombre por su forma
// plegada.
func (r *Registro) resolverNombre(entrada, plegada string) (Territorio, error) {
	candidatos := r.candidatos(plegada)

	switch len(candidatos) {
	case 0:
		return Territorio{}, nombreFueraDeLaRelacion(entrada)
	case 1:
		return r.territorioDe(candidatos[0]), nil
	default:
		return Territorio{}, nombreAmbiguo(entrada, candidatos)
	}
}
