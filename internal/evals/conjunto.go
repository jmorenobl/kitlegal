package evals

import (
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/jmorenobl/kitlegal/internal/core/territorio"
)

// formaDelNombre es la forma del nombre de un fichero de eval, <nn>-<descripción>.yaml:
// dos cifras y una descripción en minúsculas con guiones (contrato
// evals-y-grabaciones §1).
var formaDelNombre = regexp.MustCompile(`^[0-9]{2}-[a-z0-9]+(-[a-z0-9]+)*\.yaml$`)

// Conjunto es lo que LeerConjunto lee de un directorio de evals: las evals bien
// formadas, la lista de expresiones prohibidas, el juez y, aparte, las entradas
// que no se pueden leer como lo que su nombre dice que son (contrato
// evals-y-grabaciones §1; contrato lista-y-juicio §1 de H7.2;
// contracts/juez-y-voto.md §1 de H24).
type Conjunto struct {
	// Evals son las bien formadas, en orden de nombre de fichero, cada una con
	// su Fichero.
	Evals []Eval

	// MalFormados son las entradas que no se pueden leer como eval, o como la
	// lista la que lleva su nombre, y los ficheros de la carpeta del juez que
	// faltan o no tienen su forma, en orden de nombre de fichero.
	MalFormados []FicheroMalFormado

	// Prohibidas es la lista de expresiones prohibidas de la carpeta, leída de
	// expresiones-prohibidas.yaml; vacía si la carpeta no la tiene o si está mal
	// formada (research D5). Desde H24 no juzga ninguna respuesta: es el
	// vocabulario que la prosa de la skill no usa (FR-070 y FR-071 de H24).
	Prohibidas ExpresionesProhibidas

	// Juez es el juez con modelo de la skill, leído de la carpeta juez; nil si
	// la carpeta de evals no la tiene, que es no tener juez, o si está mal
	// formada (contracts/juez-y-voto.md §1 de H24; FR-020).
	Juez *Juez
}

// FicheroMalFormado es una entrada del directorio de evals que no es una eval
// (data-model §6), o un fichero de su carpeta del juez que falta o no tiene su
// forma.
type FicheroMalFormado struct {
	// Fichero es el nombre de la entrada dentro del directorio; el de un fichero
	// de la carpeta del juez lleva la carpeta delante, juez/<fichero>.
	Fichero string

	// Error dice por qué no es una eval; su texto empieza por Fichero.
	Error error
}

// LeerConjunto lee todas las entradas del directorio antes de devolver nada y las
// separa en evals bien formadas y ficheros mal formados, los dos en orden de
// nombre, la lista de expresiones prohibidas y el juez (contrato
// evals-y-grabaciones §1, FR-071; contrato lista-y-juicio §1 de H7.2;
// contracts/juez-y-voto.md §1 de H24):
//
//  1. la entrada que se llama exactamente expresiones-prohibidas.yaml es la
//     lista, no un fichero de eval: si no es un fichero regular, no se puede
//     leer o no valida, es un FicheroMalFormado con ese motivo; si no, va a
//     Prohibidas;
//  2. la entrada que se llama exactamente juez es la carpeta del juez de la
//     skill, no un fichero de eval: si no es un directorio, es un
//     FicheroMalFormado con ese motivo; si lo es, se lee con leerJuez, y cada
//     fichero suyo que falta, no se puede leer o no tiene su forma es un
//     FicheroMalFormado con su nombre, juez/<fichero>; bien formada, va a Juez.
//     Sin esa entrada, Juez es nil: la skill no tiene juez (FR-020 de H24);
//  3. cualquier otra entrada es un fichero de eval: la que no es un fichero
//     regular, o cuyo nombre no tiene la forma <nn>-<descripción>.yaml, es un
//     FicheroMalFormado con ese motivo, nunca una entrada que se salta;
//  4. cada fichero de eval se lee y se pasa a LeerEval con su nombre: si no se
//     puede leer o LeerEval devuelve un error, es un FicheroMalFormado con ese
//     error; si no, su Eval va a Evals;
//  5. el error queda para un directorio que no se puede listar: lo nombra y va
//     con un Conjunto vacío. Un directorio vacío da un Conjunto vacío sin error.
//
// Solo lee y comprueba el formato: las reglas del conjunto son de
// ComprobarConjunto.
func LeerConjunto(dir string) (Conjunto, error) {
	entradas, err := os.ReadDir(dir)
	if err != nil {
		return Conjunto{}, fmt.Errorf("el directorio de evals %s no se puede listar: %w", dir, err)
	}

	var conjunto Conjunto

	for _, entrada := range entradas {
		if entrada.Name() == carpetaDelJuez {
			juez, malFormados := leerJuez(dir, entrada)

			conjunto.Juez = juez
			conjunto.MalFormados = append(conjunto.MalFormados, malFormados...)

			continue
		}

		if err := conjunto.leer(dir, entrada); err != nil {
			conjunto.MalFormados = append(conjunto.MalFormados, FicheroMalFormado{Fichero: entrada.Name(), Error: err})
		}
	}

	return conjunto, nil
}

// leer lee en el conjunto una entrada del directorio que no es la carpeta del
// juez: la de la lista, en Prohibidas, y cualquier otra, como eval, en Evals. El
// error es el de la entrada que no se puede leer como lo que su nombre dice que
// es, y empieza por su nombre.
func (c *Conjunto) leer(dir string, entrada fs.DirEntry) error {
	if entrada.Name() == ficheroDeExpresionesProhibidas {
		lista, err := leerLista(dir, entrada)
		if err != nil {
			return err
		}

		c.Prohibidas = lista

		return nil
	}

	eval, err := leerEntrada(dir, entrada)
	if err != nil {
		return err
	}

	c.Evals = append(c.Evals, eval)

	return nil
}

// motivoNoRegular es el motivo de una entrada del directorio que no es un
// fichero regular.
const motivoNoRegular = "no es un fichero regular"

// leerLista lee la lista de expresiones prohibidas de la entrada del directorio
// que lleva su nombre, con leerExpresionesProhibidas. Todo error empieza por ese
// nombre.
func leerLista(dir string, entrada fs.DirEntry) (ExpresionesProhibidas, error) {
	if !entrada.Type().IsRegular() {
		return ExpresionesProhibidas{}, fmt.Errorf("%s: %s", entrada.Name(), motivoNoRegular)
	}

	contenido, err := leerContenido(dir, entrada.Name())
	if err != nil {
		return ExpresionesProhibidas{}, err
	}

	return leerExpresionesProhibidas(contenido)
}

// leerEntrada lee como eval una entrada del directorio. Todo error empieza por
// el nombre de la entrada y, si la entrada no es un fichero regular con la forma
// de nombre, dice cada una de las dos cosas que le faltan.
func leerEntrada(dir string, entrada fs.DirEntry) (Eval, error) {
	nombre := entrada.Name()

	var motivos []string
	if !entrada.Type().IsRegular() {
		motivos = append(motivos, motivoNoRegular)
	}

	if !formaDelNombre.MatchString(nombre) {
		motivos = append(motivos, "su nombre no tiene la forma <nn>-<descripción>.yaml")
	}

	if len(motivos) > 0 {
		return Eval{}, fmt.Errorf("%s: %s", nombre, strings.Join(motivos, " y "))
	}

	contenido, err := leerContenido(dir, nombre)
	if err != nil {
		return Eval{}, err
	}

	return LeerEval(nombre, contenido)
}

// leerContenido lee entero el fichero de la entrada del directorio con ese
// nombre, que es el de una de sus entradas, sin separadores de ruta, así que la
// ruta no sale del directorio (leerFichero). El error empieza por el nombre.
func leerContenido(dir, nombre string) ([]byte, error) {
	contenido, err := leerFichero(filepath.Join(dir, nombre))
	if err != nil {
		return nil, fmt.Errorf("%s: no se puede leer: %w", nombre, err)
	}

	return contenido, nil
}

// NormaConocida es lo que las reglas del conjunto necesitan de una norma de
// data/normas.yaml; ComprobarConjunto las recibe en un mapa por identificador
// (contrato evals-y-grabaciones §2).
type NormaConocida struct {
	// Abreviatura es la abreviatura de la norma, o vacía si no tiene.
	Abreviatura string

	// Materias son las materias de la norma.
	Materias []string
}

// DefectoDelConjunto es el incumplimiento de una regla del conjunto de evals
// (data-model §6.3 de H5 y §6.4 de H6).
type DefectoDelConjunto struct {
	// Regla es el nombre de la regla en la tabla de su juego, como «materias
	// distintas» o «normas conocidas» en el de boe-legislacion y «cubierto» en el
	// de legal-core.
	Regla string

	// Mensaje dice qué se incumple, nombrando los ficheros de eval y las normas
	// implicados.
	Mensaje string
}

// Lo que fijan las reglas del conjunto de evals de boe-legislacion (data-model
// §6.3; FR-062 a FR-064). El máximo es de 21 desde H21, con la eval sin binario
// ni servidor (contracts/evals-en-dos-modos.md §1 de H21).
const (
	minimoDeEvals             = 10
	maximoDeEvals             = 21
	positivasDelConjunto      = 10
	preguntaDelArticulo21Eval = "¿qué dice el art. 21 de la Ley 39/2015?"
	normaDelArticulo21        = "BOE-A-2015-10565"
	bloqueDelArticulo21       = "a21"
	materiaFiscal             = "tributos"
	skillQueSeReproduce       = "boe-fiscal"
)

// abreviaturasDelHito son las de las normas que alguna positiva tiene que citar
// (FR-062), en el orden de data-model §6.3.
var abreviaturasDelHito = []string{"LPAC", "LCSP", "LRBRL", "LGT", "TRLRHL"}

// Lo que fijan las reglas del conjunto de evals de legal-core (contrato de evals
// §3 de H6; FR-080 a FR-082).
const (
	minimoDeEvalsDeLegalCore = 3

	// valorNoConfigurado es el valor del vocabulario de cobertura con el que el
	// applet territorio dice que la comunidad no tiene configurado un aspecto
	// (data-model §2.4 de H6).
	valorNoConfigurado = "no-configurado"
)

// sinBinarioNiServidorDelConjunto son las evals sin binario ni servidor que
// lleva el conjunto de cada skill: exactamente una (contracts/evals-en-dos-modos.md
// §1 de H21; FR-046).
const sinBinarioNiServidorDelConjunto = 1

// ReglaDelConjunto es una regla de un juego de reglas del conjunto de evals, con
// su nombre en la tabla de ese juego. Los juegos los dan ReglasDeBoeLegislacion,
// ReglasDeLegalCore y ReglasDeJurisprudencia, y los aplica ComprobarConjunto.
type ReglaDelConjunto struct {
	nombre string

	// incumplimiento devuelve el mensaje del defecto, o vacío si el conjunto
	// cumple la regla.
	incumplimiento func(conjunto *conjuntoAComprobar) string
}

// ReglasDeBoeLegislacion son las reglas del conjunto de evals de boe-legislacion
// de data-model §6.3 de H5, en el orden de su tabla, salvo la de revisión (sin
// municipio), que no es mecánica, y, detrás, la de la eval sin binario ni
// servidor de H21. Cada llamada devuelve un juego nuevo.
func ReglasDeBoeLegislacion() []ReglaDelConjunto {
	return []ReglaDelConjunto{
		{nombre: "tamaño", incumplimiento: incumplimientoDelTamanio},
		{nombre: "positivas", incumplimiento: incumplimientoDePositivas},
		{nombre: "no activación", incumplimiento: incumplimientoDeNoActivacion},
		{nombre: "informativas", incumplimiento: incumplimientoDeInformativas},
		{nombre: "materias distintas", incumplimiento: incumplimientoDeMateriasDistintas},
		{nombre: "normas del hito", incumplimiento: incumplimientoDeNormasDelHito},
		{nombre: "art. 21", incumplimiento: incumplimientoDelArticulo21},
		{nombre: "fiscal", incumplimiento: incumplimientoFiscal},
		{nombre: "boe-fiscal", incumplimiento: incumplimientoDeBoeFiscal},
		{nombre: "normas conocidas", incumplimiento: incumplimientoDeNormasConocidas},
		{nombre: reglaSinBinarioNiServidor, incumplimiento: incumplimientoDeSinBinarioNiServidor},
	}
}

// reglaSinBinarioNiServidor es el nombre, en los dos juegos, de la regla de la
// eval sin binario ni servidor (contracts/evals-en-dos-modos.md §1 de H21).
const reglaSinBinarioNiServidor = "sin binario ni servidor"

// ReglasDeLegalCore son las reglas del conjunto de evals de legal-core del
// contrato de evals §3 de H6, en el orden de su tabla, y, detrás, la de la eval
// sin binario ni servidor de H21. Cada llamada devuelve un juego nuevo.
func ReglasDeLegalCore() []ReglaDelConjunto {
	return []ReglaDelConjunto{
		{nombre: "tamaño", incumplimiento: incumplimientoDelMinimoDeLegalCore},
		{nombre: "cubierto", incumplimiento: incumplimientoDelMunicipioCubierto},
		{nombre: "no cubierto", incumplimiento: incumplimientoDelMunicipioNoCubierto},
		{nombre: "no activación", incumplimiento: incumplimientoDeNoActivacion},
		{nombre: "esperado verificable", incumplimiento: incumplimientoDelEsperadoVerificable},
		{nombre: reglaSinBinarioNiServidor, incumplimiento: incumplimientoDeSinBinarioNiServidor},
	}
}

// Lo que fijan las reglas del conjunto de evals de jurisprudencia
// (contracts/evals-jurisprudencia.md §5 de H23; FR-050, FR-056 de H23; y, desde
// H25, contracts/evals-jurisprudencia.md §3 de H25; FR-065 de H25): las diez
// evals y cuántas hay de cada una de sus cinco clases.
const (
	tamanioDeJurisprudencia = 10

	// porNumeroYFechaDelConjunto son las que preguntan por una sentencia que se
	// da por su número de resolución y su fecha; porMateriaDelConjunto, las de
	// una pregunta por materia; conElDocumentoDelConjunto, las del documento que
	// la persona trae; y unaDeCadaClase, las de cada una de las otras dos
	// clases.
	porNumeroYFechaDelConjunto = 3
	porMateriaDelConjunto      = 2
	conElDocumentoDelConjunto  = 3
	unaDeCadaClase             = 1

	// appletDeLaCita es el applet de los dos comandos de cita.
	appletDeLaCita = "cita"

	// Las dos casillas del buscador que da cita preparar para un número de
	// resolución con su fecha.
	casillaDelNumeroDeResolucion = "Nº Resolución"
	casillaDeLaFechaDeResolucion = "Fecha resolución"
)

// ReglasDeJurisprudencia son las reglas del conjunto de evals de jurisprudencia
// de contracts/evals-jurisprudencia.md §5 de H23, en su orden, con las cuentas
// de contracts/evals-jurisprudencia.md §3 de H25: son diez evals, todas activan
// la skill y todas declaran sentencias; y, por lo que espera cada una, tres son
// por número y fecha, dos por materia, tres con el documento traído, una de
// una sentencia no cubierta y una con un documento que no es el pedido. Ninguna
// regla fija los valores de una eval: una clase se reconoce por lo que la eval
// declara, no por su número, su fecha ni su cita. Cada regla que no se cumple
// dice cuántas hay, cuántas lleva el conjunto y qué ficheros la cumplen. Cada
// llamada devuelve un juego nuevo.
func ReglasDeJurisprudencia() []ReglaDelConjunto {
	return []ReglaDelConjunto{
		{nombre: "tamaño", incumplimiento: incumplimientoDelTamanioDeJurisprudencia},
		{nombre: "activación", incumplimiento: incumplimientoDeLaActivacion},
		{nombre: "sentencias", incumplimiento: incumplimientoDeLasSentencias},
		{nombre: "número y fecha", incumplimiento: incumplimientoDeLaClase(porNumeroYFechaDelConjunto,
			"con la línea, ninguna cita, una dirección y las casillas «"+casillaDelNumeroDeResolucion+"» y «"+
				casillaDeLaFechaDeResolucion+"», con cita preparar", esPorNumeroYFecha)},
		{nombre: "materia", incumplimiento: incumplimientoDeLaClase(porMateriaDelConjunto,
			"con la dirección de búsqueda y ninguna cita, con cita preparar con texto", esPorMateria)},
		{nombre: "documento", incumplimiento: incumplimientoDeLaClase(conElDocumentoDelConjunto,
			"con una cita, con cita cotejar", esConElDocumento)},
		{nombre: "no cubierta", incumplimiento: incumplimientoDeLaClase(unaDeCadaClase,
			"sin comandos, con una dirección y ninguna cita", esDeUnaNoCubierta)},
		{nombre: "documento distinto", incumplimiento: incumplimientoDeLaClase(unaDeCadaClase,
			"con la línea, un ROJ que no se cita, y cita cotejar y cita preparar con ese ROJ", esConOtroDocumento)},
	}
}

// ComprobarConjunto aplica a las evals bien formadas de una carpeta evals/<skill>/
// las reglas del juego dado, con las normas de data/normas.yaml por identificador
// (contrato evals-y-grabaciones §2 de H5; contrato de evals §3 de H6;
// contracts/evals-jurisprudencia.md §5 de H23). Devuelve un
// defecto por cada regla que se incumple, en el orden del juego, y ninguno si se
// cumplen todas. No lee ficheros.
func ComprobarConjunto(evals []Eval, normas map[string]NormaConocida, reglas []ReglaDelConjunto) []DefectoDelConjunto {
	conjunto := nuevoConjuntoAComprobar(evals, normas)

	var defectos []DefectoDelConjunto

	for _, regla := range reglas {
		if mensaje := regla.incumplimiento(conjunto); mensaje != "" {
			defectos = append(defectos, DefectoDelConjunto{Regla: regla.nombre, Mensaje: mensaje})
		}
	}

	return defectos
}

// conjuntoAComprobar son las evals y las normas que miran las reglas, con lo que
// varias de ellas necesitan calculado una sola vez. Las evals se nombran por su
// posición en evals, y no por su Fichero, para no confundir dos con el mismo.
type conjuntoAComprobar struct {
	evals  []Eval
	normas map[string]NormaConocida

	// positivas son las posiciones de las evals con activa: true que deciden el
	// veredicto, es decir, las que no son informativas, sin la eval sin binario
	// ni servidor. Las reglas que cuentan materias miran solo estas: una eval
	// informativa no decide el veredicto, y una pregunta por materia repite
	// además la norma de la positiva de la que sale (ADR 0016); y la eval sin
	// binario ni servidor decide, pero no consulta ni cita ninguna norma
	// (contracts/evals-en-dos-modos.md §1 de H21).
	positivas []int

	// sinBinarioNiServidor son las posiciones de las evals con
	// sin_binario_ni_servidor: true.
	sinBinarioNiServidor []int

	// informativas son las posiciones de las evals con informativa: true, y
	// informativasSinActivar, las de esas que además no son positivas.
	informativas           []int
	informativasSinActivar []int

	// citadaPor da, por norma, las posiciones de las positivas que la citan.
	citadaPor map[string][]int

	// delArticulo21 son las posiciones de las evals que deciden con la pregunta y
	// la cita del art. 21.
	delArticulo21 []int

	// sobreUnMunicipio son las posiciones de las evals con activa: true que
	// tienen un comando de territorio.
	sobreUnMunicipio []int
}

// nuevoConjuntoAComprobar pasa una sola vez por las evals y deja hecho lo que
// miran las reglas.
func nuevoConjuntoAComprobar(evals []Eval, normas map[string]NormaConocida) *conjuntoAComprobar {
	conjunto := &conjuntoAComprobar{evals: evals, normas: normas, citadaPor: map[string][]int{}}

	for posicion, eval := range evals {
		if eval.Activa && slices.ContainsFunc(eval.Comandos, func(comando ComandoEsperado) bool {
			return formaDelComando(comando) == formaTerritorio
		}) {
			conjunto.sobreUnMunicipio = append(conjunto.sobreUnMunicipio, posicion)
		}

		if eval.Informativa {
			conjunto.informativas = append(conjunto.informativas, posicion)

			if !eval.Activa {
				conjunto.informativasSinActivar = append(conjunto.informativasSinActivar, posicion)
			}
		}

		if eval.SinBinarioNiServidor {
			conjunto.sinBinarioNiServidor = append(conjunto.sinBinarioNiServidor, posicion)

			continue
		}

		if eval.Informativa {
			continue
		}

		if eval.Pregunta == preguntaDelArticulo21Eval && citaElBloque(eval, normaDelArticulo21, bloqueDelArticulo21) {
			conjunto.delArticulo21 = append(conjunto.delArticulo21, posicion)
		}

		if !eval.Activa {
			continue
		}

		conjunto.positivas = append(conjunto.positivas, posicion)

		for _, norma := range normasCitadas(eval) {
			conjunto.citadaPor[norma] = append(conjunto.citadaPor[norma], posicion)
		}
	}

	return conjunto
}

// incumplimientoDelTamanio: entre 10 y 21 ficheros.
func incumplimientoDelTamanio(conjunto *conjuntoAComprobar) string {
	if len(conjunto.evals) >= minimoDeEvals && len(conjunto.evals) <= maximoDeEvals {
		return ""
	}

	return fmt.Sprintf("hay %d evals y el conjunto lleva entre %d y %d: %s",
		len(conjunto.evals), minimoDeEvals, maximoDeEvals, conjunto.todosLosFicheros())
}

// incumplimientoDePositivas: exactamente 10 con activa: true que deciden, es
// decir, sin contar las informativas, ni la eval sin binario ni servidor.
func incumplimientoDePositivas(conjunto *conjuntoAComprobar) string {
	if len(conjunto.positivas) == positivasDelConjunto {
		return ""
	}

	return fmt.Sprintf("hay %d evals positivas (activa: true) que deciden y el conjunto lleva exactamente %d: %s",
		len(conjunto.positivas), positivasDelConjunto, conjunto.ficheros(conjunto.positivas))
}

// incumplimientoDeNoActivacion: al menos una con activa: false.
func incumplimientoDeNoActivacion(conjunto *conjuntoAComprobar) string {
	if slices.ContainsFunc(conjunto.evals, func(eval Eval) bool { return !eval.Activa }) {
		return ""
	}

	return "ninguna eval es de no activación (activa: false): " + conjunto.todosLosFicheros()
}

// incumplimientoDeInformativas: al menos una con informativa: true, y todas las
// informativas son positivas. Una eval informativa se mide y se publica sin
// decidir el veredicto (ADR 0016), y hoy lo es por una de tres razones: las
// preguntas por materia, porque la herramienta que las haría posibles sigue en el
// backlog; la de la norma derogada, porque promoverla a decisoria se decide con
// los datos de varias ejecuciones (H5.1); y la de la consulta repetida tras un
// cambio de versión, que se promueve también con los datos de varias ejecuciones
// (FR-087 de H7). Si midieran una no activación, no medirían nada.
func incumplimientoDeInformativas(conjunto *conjuntoAComprobar) string {
	if len(conjunto.informativasSinActivar) > 0 {
		return "evals informativas que no son positivas (activa: true): " +
			conjunto.ficheros(conjunto.informativasSinActivar)
	}

	if len(conjunto.informativas) > 0 {
		return ""
	}

	return "ninguna eval es informativa (informativa: true): " + conjunto.todosLosFicheros()
}

// incumplimientoDeMateriasDistintas: cada positiva que decide cita al menos una
// norma que no es cita esperada de ninguna otra de ellas. Nombra las positivas que no la
// tienen y, de cada norma que citan, las positivas que la citan.
func incumplimientoDeMateriasDistintas(conjunto *conjuntoAComprobar) string {
	var sinNormaPropia []int

	var compartidas []string

	for _, posicion := range conjunto.positivas {
		normas := normasCitadas(conjunto.evals[posicion])
		if slices.ContainsFunc(normas, func(norma string) bool { return len(conjunto.citadaPor[norma]) == 1 }) {
			continue
		}

		sinNormaPropia = append(sinNormaPropia, posicion)

		for _, norma := range normas {
			compartida := fmt.Sprintf("%s (%s)", norma, conjunto.ficheros(conjunto.citadaPor[norma]))
			if !slices.Contains(compartidas, compartida) {
				compartidas = append(compartidas, compartida)
			}
		}
	}

	if len(sinNormaPropia) == 0 {
		return ""
	}

	return fmt.Sprintf("positivas sin una norma que no cite otra positiva: %s; sus normas y las positivas que las citan: %s",
		conjunto.ficheros(sinNormaPropia), enumerar(compartidas, "ninguna"))
}

// incumplimientoDeNormasDelHito: cada abreviatura del hito es la de una sola
// norma conocida, y esa norma es cita esperada de alguna positiva. Nombra cada
// abreviatura que no lo cumple con las normas que la llevan.
func incumplimientoDeNormasDelHito(conjunto *conjuntoAComprobar) string {
	var incumplidas []string

	for _, abreviatura := range abreviaturasDelHito {
		conEsaAbreviatura := conjunto.normasQueCumplen(func(norma NormaConocida) bool {
			return norma.Abreviatura == abreviatura
		})
		if len(conEsaAbreviatura) == 1 && len(conjunto.citadaPor[conEsaAbreviatura[0]]) > 0 {
			continue
		}

		incumplidas = append(incumplidas,
			fmt.Sprintf("%s (normas con esa abreviatura: %s)", abreviatura, enumerar(conEsaAbreviatura, "ninguna")))
	}

	if len(incumplidas) == 0 {
		return ""
	}

	return fmt.Sprintf("cada una de %s tiene que ser la abreviatura de una sola norma de data/normas.yaml "+
		"que cite alguna positiva, y no lo es: %s; positivas: %s",
		strings.Join(abreviaturasDelHito, ", "), strings.Join(incumplidas, "; "), conjunto.ficheros(conjunto.positivas))
}

// incumplimientoDelArticulo21: una eval con la pregunta exacta del art. 21 y la
// cita BOE-A-2015-10565 a21. Nombra las evals que tienen solo una de las dos.
func incumplimientoDelArticulo21(conjunto *conjuntoAComprobar) string {
	if len(conjunto.delArticulo21) > 0 {
		return ""
	}

	var conLaPregunta, conLaCita []int

	for posicion, eval := range conjunto.evals {
		if eval.Pregunta == preguntaDelArticulo21Eval {
			conLaPregunta = append(conLaPregunta, posicion)
		}

		if citaElBloque(eval, normaDelArticulo21, bloqueDelArticulo21) {
			conLaCita = append(conLaCita, posicion)
		}
	}

	return fmt.Sprintf("ninguna eval pregunta exactamente «%s» con la cita esperada %s %s; "+
		"con esa pregunta: %s; con esa cita: %s",
		preguntaDelArticulo21Eval, normaDelArticulo21, bloqueDelArticulo21,
		enumerar(conjunto.nombres(conLaPregunta), "ninguna"), enumerar(conjunto.nombres(conLaCita), "ninguna"))
}

// incumplimientoFiscal: una positiva distinta de la eval del art. 21 cita una
// norma cuyas materias incluyen tributos (FR-063: «al menos otra»). Si hay más de
// una eval del art. 21, cualquiera de ellas es otra respecto de las demás, así que
// solo se descarta la eval del art. 21 cuando es única.
func incumplimientoFiscal(conjunto *conjuntoAComprobar) string {
	fiscales := conjunto.normasQueCumplen(func(norma NormaConocida) bool {
		return slices.Contains(norma.Materias, materiaFiscal)
	})

	var cuentan []int

	for _, posicion := range conjunto.positivas {
		if len(conjunto.delArticulo21) == 1 && conjunto.delArticulo21[0] == posicion {
			continue
		}

		if slices.ContainsFunc(normasCitadas(conjunto.evals[posicion]), func(norma string) bool {
			return slices.Contains(fiscales, norma)
		}) {
			return ""
		}

		cuentan = append(cuentan, posicion)
	}

	return fmt.Sprintf("ninguna positiva distinta de la eval del art. 21 cita una norma cuyas materias incluyen %s; "+
		"positivas distintas de la eval del art. 21: %s; normas con esa materia: %s",
		materiaFiscal, conjunto.ficheros(cuentan), enumerar(fiscales, "ninguna"))
}

// incumplimientoDeBoeFiscal: al menos una con reproduce: boe-fiscal (FR-064).
func incumplimientoDeBoeFiscal(conjunto *conjuntoAComprobar) string {
	if slices.ContainsFunc(conjunto.evals, func(eval Eval) bool { return eval.Reproduce == skillQueSeReproduce }) {
		return ""
	}

	return fmt.Sprintf("ninguna eval declara «reproduce: %s»: %s", skillQueSeReproduce, conjunto.todosLosFicheros())
}

// incumplimientoDeNormasConocidas: toda norma de una cita o de un comando
// esperado está en data/normas.yaml (FR-020). Nombra cada norma que no está, en
// el orden en que aparece, con las evals que la esperan.
func incumplimientoDeNormasConocidas(conjunto *conjuntoAComprobar) string {
	var desconocidas []string

	esperadaPor := map[string][]int{}

	for posicion, eval := range conjunto.evals {
		for _, norma := range normasEsperadas(eval) {
			if _, conocida := conjunto.normas[norma]; conocida {
				continue
			}

			if _, vista := esperadaPor[norma]; !vista {
				desconocidas = append(desconocidas, norma)
			}

			esperadaPor[norma] = append(esperadaPor[norma], posicion)
		}
	}

	if len(desconocidas) == 0 {
		return ""
	}

	partes := make([]string, 0, len(desconocidas))
	for _, norma := range desconocidas {
		partes = append(partes, fmt.Sprintf("%s (%s)", norma, conjunto.ficheros(esperadaPor[norma])))
	}

	return "normas de citas o comandos esperados que no están en data/normas.yaml: " + strings.Join(partes, "; ")
}

// incumplimientoDelMinimoDeLegalCore: al menos tres evals bien formadas.
func incumplimientoDelMinimoDeLegalCore(conjunto *conjuntoAComprobar) string {
	if len(conjunto.evals) >= minimoDeEvalsDeLegalCore {
		return ""
	}

	return fmt.Sprintf("hay %d evals y el conjunto lleva al menos %d: %s",
		len(conjunto.evals), minimoDeEvalsDeLegalCore, conjunto.todosLosFicheros())
}

// incumplimientoDelMunicipioCubierto: al menos una eval activa resuelve con su
// comando de territorio un municipio del territorio configurado y declara sus
// boletines en lo esperado. Las reglas no resuelven municipios, así que lo que
// hace cubierto al municipio se lee de lo que su eval espera: declara boletines y
// no es de una comunidad sin configuración (deComunidadSinConfiguracion), que
// también puede declarar el boletín estatal. Nombra las evals activas con un
// comando de territorio.
func incumplimientoDelMunicipioCubierto(conjunto *conjuntoAComprobar) string {
	for _, posicion := range conjunto.sobreUnMunicipio {
		eval := conjunto.evals[posicion]
		if len(eval.Territorio.Boletines) > 0 && !deComunidadSinConfiguracion(eval) {
			return ""
		}
	}

	return "ninguna eval activa resuelve con un comando de territorio un municipio del territorio configurado " +
		"y declara sus boletines en lo esperado; evals activas con un comando de territorio: " +
		conjunto.ficheros(conjunto.sobreUnMunicipio)
}

// incumplimientoDelMunicipioNoCubierto: al menos una eval activa resuelve con su
// comando de territorio un municipio de una comunidad sin configuración y declara
// en lo esperado cada aspecto de cobertura no configurado. Nombra lo que tiene que
// declarar y las evals activas con un comando de territorio.
func incumplimientoDelMunicipioNoCubierto(conjunto *conjuntoAComprobar) string {
	for _, posicion := range conjunto.sobreUnMunicipio {
		if deComunidadSinConfiguracion(conjunto.evals[posicion]) {
			return ""
		}
	}

	return fmt.Sprintf("ninguna eval activa resuelve con un comando de territorio un municipio de una comunidad "+
		"sin configuración y declara en lo esperado %s; evals activas con un comando de territorio: %s",
		strings.Join(aspectosNoConfigurados(), ", "), conjunto.ficheros(conjunto.sobreUnMunicipio))
}

// incumplimientoDelEsperadoVerificable: toda eval activa declara citas o
// territorio en lo esperado, salvo la eval sin binario ni servidor, cuyo formato
// no admite ni las unas ni el otro: lo que su respuesta tiene que llevar lo fija
// su juicio. Nombra las que no.
func incumplimientoDelEsperadoVerificable(conjunto *conjuntoAComprobar) string {
	var sinVerificable []int

	for posicion, eval := range conjunto.evals {
		if eval.Activa && !eval.SinBinarioNiServidor && len(eval.Citas) == 0 && len(eval.Territorio.elementos()) == 0 {
			sinVerificable = append(sinVerificable, posicion)
		}
	}

	if len(sinVerificable) == 0 {
		return ""
	}

	return "evals activas sin citas ni territorio en lo esperado: " + conjunto.ficheros(sinVerificable)
}

// incumplimientoDeSinBinarioNiServidor: exactamente una eval con
// sin_binario_ni_servidor: true (FR-046 de H21). Nombra las que lo declaran.
func incumplimientoDeSinBinarioNiServidor(conjunto *conjuntoAComprobar) string {
	if len(conjunto.sinBinarioNiServidor) == sinBinarioNiServidorDelConjunto {
		return ""
	}

	return fmt.Sprintf("hay %d evals sin binario ni servidor (sin_binario_ni_servidor: true) "+
		"y el conjunto lleva exactamente %d: %s", len(conjunto.sinBinarioNiServidor),
		sinBinarioNiServidorDelConjunto, conjunto.ficheros(conjunto.sinBinarioNiServidor))
}

// incumplimientoDelTamanioDeJurisprudencia: exactamente diez evals.
func incumplimientoDelTamanioDeJurisprudencia(conjunto *conjuntoAComprobar) string {
	if len(conjunto.evals) == tamanioDeJurisprudencia {
		return ""
	}

	return fmt.Sprintf("hay %d evals y el conjunto lleva exactamente %d: %s",
		len(conjunto.evals), tamanioDeJurisprudencia, conjunto.todosLosFicheros())
}

// incumplimientoDeLaActivacion: todas activan la skill. Nombra las que no.
func incumplimientoDeLaActivacion(conjunto *conjuntoAComprobar) string {
	sinActivar := conjunto.posicionesDeLasQue(func(eval Eval) bool { return !eval.Activa })
	if len(sinActivar) == 0 {
		return ""
	}

	return "evals que no activan la skill (activa: false): " + conjunto.ficheros(sinActivar)
}

// incumplimientoDeLasSentencias: todas declaran sentencias. Nombra las que no.
func incumplimientoDeLasSentencias(conjunto *conjuntoAComprobar) string {
	sinSentencias := conjunto.posicionesDeLasQue(func(eval Eval) bool { return !eval.Sentencias.declaradas() })
	if len(sinSentencias) == 0 {
		return ""
	}

	return "evals que no declaran sentencias: " + conjunto.ficheros(sinSentencias)
}

// incumplimientoDeLaClase devuelve el incumplimiento de una regla que fija
// cuántas evals del conjunto son de una clase: exactamente esas. El mensaje
// dice cuántas hay, con la descripción de la clase, cuántas lleva el conjunto
// y qué ficheros lo son.
func incumplimientoDeLaClase(cuantas int, descripcion string, esDeLaClase func(Eval) bool,
) func(conjunto *conjuntoAComprobar) string {
	return func(conjunto *conjuntoAComprobar) string {
		deLaClase := conjunto.posicionesDeLasQue(esDeLaClase)
		if len(deLaClase) == cuantas {
			return ""
		}

		return fmt.Sprintf("hay %d evals %s, y el conjunto lleva exactamente %d: %s",
			len(deLaClase), descripcion, cuantas, conjunto.ficheros(deLaClase))
	}
}

// esPorNumeroYFecha dice si la eval es de las que preguntan por una sentencia
// que se da por su número de resolución y su fecha: espera la línea, ninguna
// cita, al menos una dirección —la del buscador, desde H25 (FR-061 de H25)— y
// las dos casillas de esa consulta, con cita preparar.
func esPorNumeroYFecha(eval Eval) bool {
	esperadas := eval.Sentencias

	return esperadas.NoComprobada && esperadas.NingunaCita && len(esperadas.Direcciones) > 0 &&
		esperaLaCasilla(eval, casillaDelNumeroDeResolucion) && esperaLaCasilla(eval, casillaDeLaFechaDeResolucion) &&
		esperaElComandoDeCita(eval, func(comando ComandoEsperado) bool { return comando.Verbo == verboPreparar })
}

// esPorMateria dice si la eval es de las de una pregunta por materia: espera la
// dirección de búsqueda que devuelva la sesión y ninguna cita, con cita
// preparar con texto.
func esPorMateria(eval Eval) bool {
	return eval.Sentencias.DireccionDeBusqueda && eval.Sentencias.NingunaCita &&
		esperaElComandoDeCita(eval, func(comando ComandoEsperado) bool {
			return comando.Verbo == verboPreparar && comando.ConTexto
		})
}

// esConElDocumento dice si la eval es de las del documento que la persona
// trae: espera una cita, con cita cotejar.
func esConElDocumento(eval Eval) bool {
	return len(eval.Sentencias.Citas) > 0 &&
		esperaElComandoDeCita(eval, func(comando ComandoEsperado) bool { return comando.Verbo == verboCotejar })
}

// esDeUnaNoCubierta dice si la eval es la de una sentencia que kitlegal no
// cubre: no espera ningún comando, y sí una dirección y ninguna cita.
func esDeUnaNoCubierta(eval Eval) bool {
	return len(eval.Comandos) == 0 && len(eval.Sentencias.Direcciones) > 0 && eval.Sentencias.NingunaCita
}

// esConOtroDocumento dice si la eval es la del documento que no es el pedido:
// espera la línea y un ROJ que no se cita, con cita cotejar y cita preparar
// con ese mismo ROJ.
func esConOtroDocumento(eval Eval) bool {
	return eval.Sentencias.NoComprobada && slices.ContainsFunc(eval.Sentencias.SinCitaDelROJ, func(roj string) bool {
		conEseROJ := func(verbo string) func(ComandoEsperado) bool {
			return func(comando ComandoEsperado) bool { return comando.Verbo == verbo && comando.ROJ == roj }
		}

		return esperaElComandoDeCita(eval, conEseROJ(verboCotejar)) && esperaElComandoDeCita(eval, conEseROJ(verboPreparar))
	})
}

// esperaLaCasilla dice si la eval espera en la respuesta la casilla con ese
// nombre.
func esperaLaCasilla(eval Eval, nombre string) bool {
	return slices.ContainsFunc(eval.Sentencias.Casillas, func(casilla CasillaEsperada) bool {
		return casilla.Nombre == nombre
	})
}

// esperaElComandoDeCita dice si la eval espera algún comando del applet cita
// que cumple la condición.
func esperaElComandoDeCita(eval Eval, condicion func(ComandoEsperado) bool) bool {
	return slices.ContainsFunc(eval.Comandos, func(comando ComandoEsperado) bool {
		return comando.Applet == appletDeLaCita && condicion(comando)
	})
}

// posicionesDeLasQue son las posiciones de las evals del conjunto que cumplen
// la condición, en su orden.
func (c *conjuntoAComprobar) posicionesDeLasQue(condicion func(Eval) bool) []int {
	var posiciones []int

	for posicion, eval := range c.evals {
		if condicion(eval) {
			posiciones = append(posiciones, posicion)
		}
	}

	return posiciones
}

// deComunidadSinConfiguracion dice si la eval espera el territorio de un
// municipio de una comunidad sin configuración: su cobertura declara no
// configurado cada aspecto que puede estarlo (aspectosNoConfigurados).
func deComunidadSinConfiguracion(eval Eval) bool {
	noConfigurados := aspectosNoConfigurados()

	return len(noConfigurados) > 0 && !slices.ContainsFunc(noConfigurados, func(aspecto string) bool {
		return !slices.Contains(eval.Territorio.Cobertura, aspecto)
	})
}

// aspectosNoConfigurados son, en la forma <aspecto>: <valor> y en el orden de sus
// claves, las del vocabulario de territorio.AspectosDeCobertura que pueden valer
// no-configurado, con ese valor: los de una comunidad sin configuración.
func aspectosNoConfigurados() []string {
	var noConfigurados []string

	for _, aspecto := range territorio.AspectosDeCobertura() {
		if slices.Contains(aspecto.Valores, valorNoConfigurado) {
			noConfigurados = append(noConfigurados, aspectoDeCobertura(aspecto.Clave, valorNoConfigurado))
		}
	}

	return noConfigurados
}

// ficheros enumera los ficheros de las evals en esas posiciones.
func (c *conjuntoAComprobar) ficheros(posiciones []int) string {
	return enumerar(c.nombres(posiciones), "ningún fichero")
}

// todosLosFicheros enumera los ficheros de todas las evals del conjunto.
func (c *conjuntoAComprobar) todosLosFicheros() string {
	nombres := make([]string, 0, len(c.evals))
	for _, eval := range c.evals {
		nombres = append(nombres, eval.Fichero)
	}

	return enumerar(nombres, "ningún fichero")
}

// nombres son los ficheros de las evals en esas posiciones, en su orden.
func (c *conjuntoAComprobar) nombres(posiciones []int) []string {
	nombres := make([]string, 0, len(posiciones))
	for _, posicion := range posiciones {
		nombres = append(nombres, c.evals[posicion].Fichero)
	}

	return nombres
}

// normasQueCumplen son los identificadores de las normas conocidas que cumplen la
// condición, ordenados.
func (c *conjuntoAComprobar) normasQueCumplen(condicion func(NormaConocida) bool) []string {
	var identificadores []string

	for _, identificador := range slices.Sorted(maps.Keys(c.normas)) {
		if condicion(c.normas[identificador]) {
			identificadores = append(identificadores, identificador)
		}
	}

	return identificadores
}

// normasCitadas son las normas de las citas esperadas de la eval, en su orden y
// sin repetir.
func normasCitadas(eval Eval) []string {
	var normas []string

	for _, cita := range eval.Citas {
		if !slices.Contains(normas, cita.Norma) {
			normas = append(normas, cita.Norma)
		}
	}

	return normas
}

// normasEsperadas son las normas de los comandos esperados de la eval, que la
// búsqueda no lleva, y después las de sus citas, en su orden y sin repetir.
func normasEsperadas(eval Eval) []string {
	var normas []string

	for _, comando := range eval.Comandos {
		if comando.Norma != "" && !slices.Contains(normas, comando.Norma) {
			normas = append(normas, comando.Norma)
		}
	}

	for _, norma := range normasCitadas(eval) {
		if !slices.Contains(normas, norma) {
			normas = append(normas, norma)
		}
	}

	return normas
}

// citaElBloque dice si la eval tiene la cita esperada de esa norma y ese bloque.
func citaElBloque(eval Eval, norma, bloque string) bool {
	return slices.Contains(eval.Citas, CitaEsperada{Norma: norma, Bloque: bloque})
}

// enumerar une los elementos con comas, o da siNoHay si no hay ninguno.
func enumerar(elementos []string, siNoHay string) string {
	if len(elementos) == 0 {
		return siNoHay
	}

	return strings.Join(elementos, ", ")
}
