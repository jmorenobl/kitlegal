package evals

import (
	"fmt"
	"slices"
)

// Lo que data-model §7.1 fija de las consultas necesarias que no copian el
// verbo de un comando esperado. Las normas de las evals tienen la gramática del
// BOE, así que su índice, sus metadatos y sus bloques se leen con el applet boe,
// y un bloque, con su verbo articulo.
const (
	appletDeLasNormas = "boe"
	verboIndice       = "indice"
	verboMetadatos    = "metadatos"
	verboArticulo     = "articulo"
	verboBuscar       = "buscar"
)

// fechaDeTodaVigencia es la fecha con la que la caché preparada lee cada bloque
// además de sin fecha: la última que admite --fecha, posterior a toda vigencia,
// así que la lectura sale con 0 en todo bloque grabado y deja guardada la
// entrada de sus versiones, de la que se sirve después la lectura del bloque en
// cualquier fecha (research D18 de H20; FR-053).
const fechaDeTodaVigencia = "9999-12-31"

// Punto es de cuál de los tres puntos de data-model §7.1 sale una consulta
// necesaria de una eval.
type Punto string

const (
	// PuntoComandoEsperado es el punto 1: la consulta es un comando esperado de
	// la eval.
	PuntoComandoEsperado Punto = "comando esperado"

	// PuntoNormaDeLaEval es el punto 2: la consulta es el índice o los metadatos
	// de una norma de las citas o de los comandos esperados de la eval, que el
	// protocolo consulta aunque no sean comandos esperados (FR-007).
	PuntoNormaDeLaEval Punto = "norma de la eval"

	// PuntoCitaEsperada es el punto 3: la consulta es la lectura del bloque de
	// una cita esperada de la eval, porque lo citado sale de lo leído (FR-008).
	PuntoCitaEsperada Punto = "cita esperada"
)

// Origen es un lugar del conjunto de evals del que sale una consulta necesaria:
// la eval y el punto, para nombrarlos en un fallo (data-model §7.1).
type Origen struct {
	// Eval es el Fichero de la eval.
	Eval string

	// Punto es el punto de data-model §7.1 del que sale la consulta.
	Punto Punto
}

// Consulta es una consulta necesaria: la invocación kitlegal <applet> <verbo>
// <argumentos…> que la caché preparada tiene que poder servir (data-model §7.1;
// FR-074, FR-075).
type Consulta struct {
	// Applet es el applet que se invoca, como boe.
	Applet string

	// Verbo es el verbo de la invocación, como articulo o indice.
	Verbo string

	// Argumentos son los que siguen al verbo, en su orden.
	Argumentos []string

	// Origenes son todos los lugares del conjunto de los que sale la consulta,
	// sin repetir y en el orden en que aparecen: la misma invocación puede ser a
	// la vez un comando esperado y la lectura de una cita, o necesitarla varias
	// evals.
	Origenes []Origen
}

// ConsultasNecesarias son las consultas que necesitan las evals del conjunto
// (data-model §7.1; contrato evals-y-grabaciones §4), recorriendo cada eval en su
// orden:
//
//  1. por cada comando esperado, según su forma: la forma bloque, articulo
//     <norma> <bloque>, lleve o no fecha, y nunca con la suya, que en un bloque
//     sin redacción vigente ese día sale con 3 y sería una falta
//     (contracts/evals-redaccion-fechada.md §4 de H20; research D18); la
//     consulta de norma, <verbo> <norma>; la búsqueda,
//     buscar <terminos…>; el comando de territorio, ninguna, porque el applet
//     territorio no pide nada por red ni usa la caché y no hay nada que grabar
//     (data-model §6.3 de H6; FR-043); la comprobación, tampoco, porque lee el
//     grafo de la sesión, no la caché ni la red (contrato evals-y-skill §1 de
//     H7); los dos de cita, tampoco, porque el applet cita no consulta
//     ninguna fuente (contracts/evals-jurisprudencia.md §2 de H23; FR-054); y
//     el comando calcular, tampoco, porque el applet plazos no consulta nada
//     (contracts/evals-legal-core.md §1 de H9);
//  2. por cada norma de sus comandos y de sus citas, sin repetir: indice <norma>
//     y metadatos <norma>;
//  3. por cada cita esperada: articulo <norma> <bloque>.
//
// Desde H20, cada articulo <norma> <bloque> de los puntos 1 y 3 va seguido de
// la misma lectura con --fecha 9999-12-31, con los mismos orígenes: la caché
// preparada sirve así el bloque en cualquier fecha (research D18 de H20;
// FR-053).
//
// Ninguna invocación se repite: la que ya estaba solo suma su eval y su punto a
// los Origenes de la primera, si no los tenía. Una eval de no activación no
// añade nada.
func ConsultasNecesarias(conjunto []Eval) []Consulta {
	necesarias := consultasNecesarias{posiciones: map[string]int{}}

	for _, eval := range conjunto {
		comando := Origen{Eval: eval.Fichero, Punto: PuntoComandoEsperado}
		for _, esperado := range eval.Comandos {
			switch formaDelComando(esperado) {
			case formaBloque:
				necesarias.anotarBloque(comando, esperado.Applet, esperado.Norma, esperado.Bloque)
			case formaConsultaDeNorma:
				necesarias.anotar(comando, esperado.Applet, esperado.Verbo, esperado.Norma)
			case formaBusqueda:
				necesarias.anotar(comando, esperado.Applet, esperado.Verbo, esperado.Terminos...)
			case formaTerritorio:
				// Lo que resuelve viaja dentro del binario: no hay consulta que la
				// caché preparada tenga que servir.
			case formaComprobacion:
				// Lo que comprueba es el grafo de la sesión, que no sale de la
				// caché preparada: tampoco hay consulta que servir.
			case formaPreparar, formaCotejar:
				// El applet cita no consulta nada: prepara una consulta y coteja
				// el texto que recibe, sin red ni caché. No hay nada que grabar.
			case formaCalcular:
				// El applet plazos calcula con los festivos y los tipos que viajan
				// dentro del binario, sin red ni caché: tampoco hay nada que grabar.
			}
		}

		norma := Origen{Eval: eval.Fichero, Punto: PuntoNormaDeLaEval}
		for _, identificador := range normasEsperadas(eval) {
			necesarias.anotar(norma, appletDeLasNormas, verboIndice, identificador)
			necesarias.anotar(norma, appletDeLasNormas, verboMetadatos, identificador)
		}

		cita := Origen{Eval: eval.Fichero, Punto: PuntoCitaEsperada}
		for _, esperada := range eval.Citas {
			necesarias.anotarBloque(cita, appletDeLasNormas, esperada.Norma, esperada.Bloque)
		}
	}

	return necesarias.consultas
}

// consultasNecesarias acumula las consultas sin repetir ninguna invocación.
type consultasNecesarias struct {
	consultas []Consulta

	// posiciones da, por la clave de una invocación, su posición en consultas.
	posiciones map[string]int
}

// anotar añade la invocación con su origen o, si ya estaba, suma el origen a los
// suyos cuando no lo tenía. Los argumentos se copian: cambiar después los
// términos de la eval de la que sale no cambia la consulta.
func (c *consultasNecesarias) anotar(origen Origen, applet, verbo string, argumentos ...string) {
	// %q de la invocación entera es una clave sin ambigüedad: un término con un
	// espacio no se confunde con dos términos.
	clave := fmt.Sprintf("%q", append([]string{applet, verbo}, argumentos...))

	posicion, anotada := c.posiciones[clave]
	if !anotada {
		posicion = len(c.consultas)
		c.posiciones[clave] = posicion
		c.consultas = append(c.consultas, Consulta{Applet: applet, Verbo: verbo, Argumentos: slices.Clone(argumentos)})
	}

	if consulta := &c.consultas[posicion]; !slices.Contains(consulta.Origenes, origen) {
		consulta.Origenes = append(consulta.Origenes, origen)
	}
}

// anotarBloque anota la lectura del bloque de la norma sin fecha y, detrás, la
// misma con --fecha 9999-12-31, las dos con el origen (research D18 de H20).
func (c *consultasNecesarias) anotarBloque(origen Origen, applet, norma, bloque string) {
	c.anotar(origen, applet, verboArticulo, norma, bloque)
	c.anotar(origen, applet, verboArticulo, norma, bloque, "--"+banderaDeLaFecha, fechaDeTodaVigencia)
}
