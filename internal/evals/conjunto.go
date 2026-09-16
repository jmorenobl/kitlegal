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
)

// formaDelNombre es la forma del nombre de un fichero de eval, <nn>-<descripción>.yaml:
// dos cifras y una descripción en minúsculas con guiones (contrato
// evals-y-grabaciones §1).
var formaDelNombre = regexp.MustCompile(`^[0-9]{2}-[a-z0-9]+(-[a-z0-9]+)*\.yaml$`)

// Conjunto es lo que LeerConjunto lee de un directorio de evals: las evals bien
// formadas y, aparte, las entradas que no se pueden leer como eval (contrato
// evals-y-grabaciones §1).
type Conjunto struct {
	// Evals son las bien formadas, en orden de nombre de fichero, cada una con
	// su Fichero.
	Evals []Eval

	// MalFormados son las entradas que no se pueden leer como eval, en orden de
	// nombre de fichero.
	MalFormados []FicheroMalFormado
}

// FicheroMalFormado es una entrada del directorio de evals que no es una eval
// (data-model §6).
type FicheroMalFormado struct {
	// Fichero es el nombre de la entrada dentro del directorio.
	Fichero string

	// Error dice por qué no es una eval; su texto empieza por Fichero.
	Error error
}

// LeerConjunto lee todas las entradas del directorio antes de devolver nada y las
// separa en evals bien formadas y ficheros mal formados, los dos en orden de
// nombre (contrato evals-y-grabaciones §1, FR-071):
//
//  1. toda entrada es un fichero de eval: la que no es un fichero regular, o
//     cuyo nombre no tiene la forma <nn>-<descripción>.yaml, es un
//     FicheroMalFormado con ese motivo, nunca una entrada que se salta;
//  2. cada fichero se lee y se pasa a LeerEval con su nombre: si no se puede
//     leer o LeerEval devuelve un error, es un FicheroMalFormado con ese error;
//     si no, su Eval va a Evals;
//  3. el error queda para un directorio que no se puede listar: lo nombra y va
//     con un Conjunto vacío. Un directorio vacío da un Conjunto vacío sin error.
//
// Solo lee y comprueba el formato: las reglas del conjunto son de
// ComprobarConjuntoDeBoeLegislacion.
func LeerConjunto(dir string) (Conjunto, error) {
	entradas, err := os.ReadDir(dir)
	if err != nil {
		return Conjunto{}, fmt.Errorf("el directorio de evals %s no se puede listar: %w", dir, err)
	}

	var conjunto Conjunto

	for _, entrada := range entradas {
		eval, err := leerEntrada(dir, entrada)
		if err != nil {
			conjunto.MalFormados = append(conjunto.MalFormados, FicheroMalFormado{Fichero: entrada.Name(), Error: err})

			continue
		}

		conjunto.Evals = append(conjunto.Evals, eval)
	}

	return conjunto, nil
}

// leerEntrada lee como eval una entrada del directorio. Todo error empieza por
// el nombre de la entrada y, si la entrada no es un fichero regular con la forma
// de nombre, dice cada una de las dos cosas que le faltan.
func leerEntrada(dir string, entrada fs.DirEntry) (Eval, error) {
	nombre := entrada.Name()

	var motivos []string
	if !entrada.Type().IsRegular() {
		motivos = append(motivos, "no es un fichero regular")
	}

	if !formaDelNombre.MatchString(nombre) {
		motivos = append(motivos, "su nombre no tiene la forma <nn>-<descripción>.yaml")
	}

	if len(motivos) > 0 {
		return Eval{}, fmt.Errorf("%s: %s", nombre, strings.Join(motivos, " y "))
	}

	// filepath.Clean es lo que el control de rutas reconoce como saneado antes
	// de abrir un fichero (gosec G304); el nombre es el de una entrada del propio
	// directorio con la forma de arriba, sin separadores de ruta.
	contenido, err := os.ReadFile(filepath.Clean(filepath.Join(dir, nombre)))
	if err != nil {
		return Eval{}, fmt.Errorf("%s: no se puede leer: %w", nombre, err)
	}

	return LeerEval(nombre, contenido)
}

// NormaConocida es lo que las reglas del conjunto necesitan de una norma de
// data/normas.yaml; ComprobarConjuntoDeBoeLegislacion las recibe en un mapa por
// identificador (contrato evals-y-grabaciones §2).
type NormaConocida struct {
	// Abreviatura es la abreviatura de la norma, o vacía si no tiene.
	Abreviatura string

	// Materias son las materias de la norma.
	Materias []string
}

// DefectoDelConjunto es el incumplimiento de una regla del conjunto de evals
// (data-model §6.3).
type DefectoDelConjunto struct {
	// Regla es el nombre de la regla en la tabla de data-model §6.3, como
	// «materias distintas» o «normas conocidas».
	Regla string

	// Mensaje dice qué se incumple, nombrando los ficheros de eval y las normas
	// implicados.
	Mensaje string
}

// Lo que fijan las reglas del conjunto de evals de boe-legislacion (data-model
// §6.3; FR-062 a FR-064).
const (
	minimoDeEvals             = 10
	maximoDeEvals             = 20
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

// reglaDelConjunto es una regla de data-model §6.3 con su nombre en la tabla.
type reglaDelConjunto struct {
	nombre string

	// incumplimiento devuelve el mensaje del defecto, o vacío si el conjunto
	// cumple la regla.
	incumplimiento func(conjunto *conjuntoAComprobar) string
}

// reglasDelConjunto son las reglas de data-model §6.3 en el orden de su tabla,
// salvo la de revisión (sin municipio), que no es mecánica.
var reglasDelConjunto = []reglaDelConjunto{
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
}

// ComprobarConjuntoDeBoeLegislacion aplica a las evals bien formadas de
// evals/boe-legislacion/ las reglas de data-model §6.3, salvo la de revisión, con
// las normas de data/normas.yaml por identificador (contrato evals-y-grabaciones
// §2). Devuelve un defecto por cada regla que se incumple, en el orden de la
// tabla, y ninguno si se cumplen todas. No lee ficheros.
func ComprobarConjuntoDeBoeLegislacion(evals []Eval, normas map[string]NormaConocida) []DefectoDelConjunto {
	conjunto := nuevoConjuntoAComprobar(evals, normas)

	var defectos []DefectoDelConjunto

	for _, regla := range reglasDelConjunto {
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
	// veredicto, es decir, las que no son informativas. Las reglas que cuentan
	// materias miran solo estas: una eval informativa mide algo que la skill
	// todavía no puede hacer y repite la norma de la positiva de la que sale
	// (ADR 0016).
	positivas []int

	// informativas son las posiciones de las evals con informativa: true, y
	// informativasSinActivar, las de esas que además no son positivas.
	informativas           []int
	informativasSinActivar []int

	// citadaPor da, por norma, las posiciones de las positivas que la citan.
	citadaPor map[string][]int

	// delArticulo21 son las posiciones de las evals que deciden con la pregunta y
	// la cita del art. 21.
	delArticulo21 []int
}

// nuevoConjuntoAComprobar pasa una sola vez por las evals y deja hecho lo que
// miran las reglas.
func nuevoConjuntoAComprobar(evals []Eval, normas map[string]NormaConocida) *conjuntoAComprobar {
	conjunto := &conjuntoAComprobar{evals: evals, normas: normas, citadaPor: map[string][]int{}}

	for posicion, eval := range evals {
		if eval.Informativa {
			conjunto.informativas = append(conjunto.informativas, posicion)

			if !eval.Activa {
				conjunto.informativasSinActivar = append(conjunto.informativasSinActivar, posicion)
			}
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

// incumplimientoDelTamanio: entre 10 y 20 ficheros.
func incumplimientoDelTamanio(conjunto *conjuntoAComprobar) string {
	if len(conjunto.evals) >= minimoDeEvals && len(conjunto.evals) <= maximoDeEvals {
		return ""
	}

	return fmt.Sprintf("hay %d evals y el conjunto lleva entre %d y %d: %s",
		len(conjunto.evals), minimoDeEvals, maximoDeEvals, conjunto.todosLosFicheros())
}

// incumplimientoDePositivas: exactamente 10 con activa: true que deciden, es
// decir, sin contar las informativas.
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
// informativas son positivas. Son las preguntas por materia, que vuelven al
// conjunto sin decidir el veredicto porque la herramienta que las haría posibles
// sigue en el backlog (ADR 0016); si midieran una no activación, no medirían
// nada.
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
