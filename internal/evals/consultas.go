package evals

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"

	"github.com/jmorenobl/kitlegal/internal/core/plazos"
	"github.com/jmorenobl/kitlegal/internal/skills"
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

// datosDelRepositorio es el data/ del repositorio, relativo al directorio de
// este paquete, que es donde go test ejecuta sus tests y el job sus sesiones
// (research V19 de H26.1).
const datosDelRepositorio = "../../data"

// Los tres ficheros de data/ de los que sale lo que la skill lleva a leer, por
// su nombre dentro del directorio (contracts/cache-de-la-sesion.md §1, punto 2,
// de H26.1): los tipos de plazo, la cobertura por materia y la tabla de normas.
const (
	ficheroDeLosTipos    = "plazos.yaml"
	ficheroDeLaCobertura = "cobertura.yaml"
	ficheroDeLasNormas   = "normas.yaml"
)

// directorioDeLosDatos es como nombra a data/ el origen de lo que la skill lleva
// a leer, sea cual sea el directorio del que se lee: data/plazos.yaml también
// desde una copia (data-model §2 de H26.1).
const directorioDeLosDatos = "data"

// Punto es de cuál de los puntos de data-model §7.1, o de lo que la skill lleva
// a leer de data/, sale una consulta que la caché de una sesión tiene que
// servir.
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

	// PuntoLoQueLlevaALeer no sale de ninguna eval: la consulta es la lectura de
	// un bloque que nombra data/plazos.yaml o data/cobertura.yaml, o el índice o
	// los metadatos de una norma de esos dos ficheros o de data/normas.yaml, que
	// la skill lleva a leer en cualquier sesión (data-model §2 de H26.1; FR-001).
	PuntoLoQueLlevaALeer Punto = "lo que la skill lleva a leer"
)

// Origen es un lugar del conjunto de evals del que sale una consulta necesaria:
// la eval y el punto, para nombrarlos en un fallo (data-model §7.1).
type Origen struct {
	// Eval es el Fichero de la eval o, con PuntoLoQueLlevaALeer, el fichero de
	// data/ del que sale la consulta, como data/plazos.yaml.
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
	necesarias.anotarLasEvals(conjunto)

	return necesarias.consultas
}

// consultasDeLaCache son las consultas con las que se prepara la caché de una
// sesión de las evals de una skill y la base de la medida de su juez
// (contracts/cache-de-la-sesion.md §1, puntos 1 y 2, de H26.1; research D1 y D4
// de H26.1; FR-001, FR-004): las de ConsultasNecesarias de las evals y, detrás,
// las de lecturasDeLosDatos del directorio data/, con el mismo acumulador, de
// modo que una invocación que ya estaba solo suma su origen. Su error es el de
// lecturasDeLosDatos, y con él no hay consultas.
func consultasDeLaCache(conjunto []Eval, datos string) ([]Consulta, error) {
	necesarias := consultasNecesarias{posiciones: map[string]int{}}
	necesarias.anotarLasEvals(conjunto)

	if err := necesarias.anotarLosDatos(datos); err != nil {
		return nil, err
	}

	return necesarias.consultas, nil
}

// lecturasDeLosDatos es lo que la skill lleva a leer del directorio data/
// (data-model §2 de H26.1; research D1 de H26.1; FR-001, FR-002), con el punto
// PuntoLoQueLlevaALeer y, como origen, el fichero de data/ del que sale cada
// consulta: ver anotarLosDatos. Su error es el de un fichero que no se lee o que
// su lector rechaza, y con él no hay consultas.
func lecturasDeLosDatos(datos string) ([]Consulta, error) {
	necesarias := consultasNecesarias{posiciones: map[string]int{}}

	if err := necesarias.anotarLosDatos(datos); err != nil {
		return nil, err
	}

	return necesarias.consultas, nil
}

// consultasNecesarias acumula las consultas sin repetir ninguna invocación.
type consultasNecesarias struct {
	consultas []Consulta

	// posiciones da, por la clave de una invocación, su posición en consultas.
	posiciones map[string]int
}

// anotarLasEvals anota las consultas necesarias de las evals del conjunto, en su
// orden, como las describe ConsultasNecesarias.
func (c *consultasNecesarias) anotarLasEvals(conjunto []Eval) {
	for _, eval := range conjunto {
		comando := Origen{Eval: eval.Fichero, Punto: PuntoComandoEsperado}
		for _, esperado := range eval.Comandos {
			switch formaDelComando(esperado) {
			case formaBloque:
				c.anotarBloque(comando, esperado.Applet, esperado.Norma, esperado.Bloque)
			case formaConsultaDeNorma:
				c.anotar(comando, esperado.Applet, esperado.Verbo, esperado.Norma)
			case formaBusqueda:
				c.anotar(comando, esperado.Applet, esperado.Verbo, esperado.Terminos...)
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
			c.anotarNorma(norma, identificador)
		}

		cita := Origen{Eval: eval.Fichero, Punto: PuntoCitaEsperada}
		for _, esperada := range eval.Citas {
			c.anotarBloque(cita, appletDeLasNormas, esperada.Norma, esperada.Bloque)
		}
	}
}

// bloqueDeLosDatos es un bloque que nombra un fichero de data/: la norma, por su
// identificador, y el id del bloque en su índice consolidado.
type bloqueDeLosDatos struct {
	norma, bloque string
}

// anotarLosDatos anota lo que la skill lleva a leer del directorio data/
// (data-model §2 de H26.1; FR-001, FR-002): lee plazos.yaml con
// plazos.LeerTipos, cobertura.yaml con skills.LeerCobertura y normas.yaml con
// skills.LeerNormas, y anota, en este orden y sin repetir ninguna invocación:
//
//  1. con anotarBloque, cada bloque de los tipos (bloquesDeLosTipos), con el
//     origen data/plazos.yaml, y cada uno de la cobertura
//     (bloquesDeLaCobertura), con el origen data/cobertura.yaml;
//  2. de cada norma de esos bloques, con el origen del fichero que la nombra, y
//     de cada entrada de normas.yaml, con el origen data/normas.yaml, indice
//     <norma> y metadatos <norma>.
//
// Ningún identificador de norma ni de bloque sale de otro sitio que de los tres
// ficheros. Un fichero que no se lee o que su lector rechaza es el error, que
// nombra su ruta, y entonces no anota nada.
func (c *consultasNecesarias) anotarLosDatos(datos string) error {
	tipos, err := leerDeLosDatos(datos, ficheroDeLosTipos, plazos.LeerTipos)
	if err != nil {
		return err
	}

	cobertura, err := leerDeLosDatos(datos, ficheroDeLaCobertura, skills.LeerCobertura)
	if err != nil {
		return err
	}

	tabla, err := leerDeLosDatos(datos, ficheroDeLasNormas, skills.LeerNormas)
	if err != nil {
		return err
	}

	deLasNormas := origenDeLosDatos(ficheroDeLasNormas)
	ficheros := []struct {
		origen  Origen
		bloques []bloqueDeLosDatos
	}{
		{origen: origenDeLosDatos(ficheroDeLosTipos), bloques: bloquesDeLosTipos(tipos)},
		{origen: origenDeLosDatos(ficheroDeLaCobertura), bloques: bloquesDeLaCobertura(cobertura)},
	}

	for _, fichero := range ficheros {
		for _, leido := range fichero.bloques {
			c.anotarBloque(fichero.origen, appletDeLasNormas, leido.norma, leido.bloque)
		}
	}

	for _, fichero := range ficheros {
		for _, leido := range fichero.bloques {
			c.anotarNorma(fichero.origen, leido.norma)
		}
	}

	for _, norma := range tabla {
		c.anotarNorma(deLasNormas, norma.Identificador)
	}

	return nil
}

// leerDeLosDatos lee el fichero del directorio data/ con su lector. El error
// nombra la ruta del fichero: la del que no se lee o la del que el lector
// rechaza, con el error del lector.
func leerDeLosDatos[T any](datos, fichero string, leer func([]byte) (T, error)) (T, error) {
	var leido T

	ruta := filepath.Join(datos, fichero)

	contenido, err := fs.ReadFile(os.DirFS(datos), fichero)
	if err != nil {
		return leido, fmt.Errorf("lo que la skill lleva a leer: %s no se puede leer: %w", ruta, err)
	}

	leido, err = leer(contenido)
	if err != nil {
		return leido, fmt.Errorf("lo que la skill lleva a leer: %s: %w", ruta, err)
	}

	return leido, nil
}

// origenDeLosDatos es el origen de lo que la skill lleva a leer de ese fichero
// de data/.
func origenDeLosDatos(fichero string) Origen {
	return Origen{Eval: path.Join(directorioDeLosDatos, fichero), Punto: PuntoLoQueLlevaALeer}
}

// bloquesDeLosTipos son los bloques de los tipos de plazo, en su orden: de cada
// tipo, los de sus preceptos, los de sus unidades, en el orden de
// plazos.Unidades, y los de los preceptos de su cómputo.
func bloquesDeLosTipos(tipos []plazos.Tipo) []bloqueDeLosDatos {
	var bloques []bloqueDeLosDatos

	deLosPreceptos := func(preceptos ...plazos.Precepto) {
		for _, precepto := range preceptos {
			bloques = append(bloques, bloqueDeLosDatos{norma: precepto.Norma, bloque: precepto.Bloque})
		}
	}

	for _, tipo := range tipos {
		deLosPreceptos(tipo.Preceptos...)

		for _, unidad := range plazos.Unidades() {
			if precepto, esta := tipo.Unidades[unidad]; esta {
				deLosPreceptos(precepto)
			}
		}

		deLosPreceptos(tipo.PreceptosDelComputo...)
	}

	return bloques
}

// bloquesDeLaCobertura son los bloques de la cobertura por materia, en el orden
// de sus entradas: de cada una, el de su precepto y, si nombra la norma que la
// rige, el del precepto aplicable de esa norma.
func bloquesDeLaCobertura(cobertura skills.Cobertura) []bloqueDeLosDatos {
	var bloques []bloqueDeLosDatos

	for _, regulado := range cobertura.ReguladosAparte {
		bloques = append(bloques, bloqueDeLosDatos{norma: regulado.Precepto.Norma, bloque: regulado.Precepto.Bloque})

		if norma := regulado.Norma; norma != nil {
			bloques = append(bloques, bloqueDeLosDatos{norma: norma.Identificador, bloque: norma.Aplicable.Bloque})
		}
	}

	return bloques
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

// anotarNorma anota indice <norma> y, detrás, metadatos <norma>, los dos con el
// origen.
func (c *consultasNecesarias) anotarNorma(origen Origen, norma string) {
	c.anotar(origen, appletDeLasNormas, verboIndice, norma)
	c.anotar(origen, appletDeLasNormas, verboMetadatos, norma)
}
