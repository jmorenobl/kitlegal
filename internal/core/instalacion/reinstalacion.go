package instalacion

import "slices"

// PermisoDelAviso es lo que una invocación permite hacer al aviso de las
// skills instaladas (data-model §1 de H38; contracts/aviso-y-reinstalacion.md
// §1). Lo decide el kernel con lo que sabe de la invocación —el análisis, la
// decisión, --dry-run y el verbo— y lo recibe Reinstalacion.Avisar. El valor
// cero es SoloAvisar, el que menos hace.
type PermisoDelAviso int

// Los tres permisos, del que menos hace al que más.
const (
	// SoloAvisar da la línea de hoy, la de Aviso, sin mirar nada más: el de un
	// análisis fallido, el de --describe y el de mcp serve (FR-002, FR-015).
	SoloAvisar PermisoDelAviso = iota

	// Ensayar da la línea de FR-009 si la reinstalación saldría, sin tomar el
	// candado ni escribir nada: el de una ejecución con --dry-run (FR-009).
	Ensayar

	// Reinstalar reinstala las skills globales si se cumplen FR-003 (a) a (d) y
	// da la línea de FR-006: el de cualquier otra ejecución (FR-001).
	Reinstalar
)

// finDeLasLineas es lo que cierra las dos líneas del aviso que reinstala
// (contracts/aviso-y-reinstalacion.md §3).
const finDeLasLineas = "; se aplican en la próxima conversación"

// Reinstalacion son los puertos y los datos con que el aviso de una invocación
// decide y, si puede, reinstala las skills globales (data-model §2 de H38;
// research.md D1): un valor por llamada, que no guarda nada entre llamadas.
// Disco, Enlazador, Escritor y Candado son obligatorios; los pone la
// composición del kernel, con los mismos adaptadores con que instala
// kitlegal skills install, así que reinstala lo que instalaría (FR-005).
type Reinstalacion struct {
	// Disco es la lectura del sistema de ficheros, la del aviso y la de
	// Planificar.
	Disco Disco

	// Enlazador es el de install: sondea los hosts al planificar y enlaza al
	// aplicar.
	Enlazador Enlazador

	// Escritor aplica el plan de la reinstalación.
	Escritor Escritor

	// Candado toma el candado de la reinstalación (FR-010).
	Candado TomarCandado

	// Home es el valor de HOME con /, vacío si no está definido, como lo
	// recibe Aviso.
	Home string

	// Version es la del binario.
	Version string

	// Empotradas son las skills que el binario lleva dentro.
	Empotradas []SkillEmpotrada

	// Retiradas son los nombres de skill que el binario retira, con sus
	// herederas.
	Retiradas Retiradas
}

// Avisar es la línea que la invocación escribe en la salida de error, y si hay
// que darla, con lo que permite permiso (contracts/aviso-y-reinstalacion.md
// §2; FR-003 a FR-011, FR-013, FR-014). Sigue los pasos del contrato y el
// primero que decide termina:
//
//  1. hoy: sin la línea de Aviso, ninguna línea y nada más;
//  2. con SoloAvisar, la línea de hoy;
//  3. ámbito: si el manifiesto que compara el aviso es el local, la de hoy;
//  4. versiones: si no se cumple la regla de FR-004, la de hoy;
//  5. plan: si Planificar, con el pedido de data-model §4, da un error —un
//     conflicto o un fallo del disco o de la sonda— o una skill «sin
//     retirar», la de hoy;
//  6. con Ensayar, la línea de FR-009, sin candado ni escritura;
//  7. candado: sin tomarlo, la de hoy; tomado, se libera al terminar, pase lo
//     que pase;
//  8. otra vez, bajo el candado y sobre el disco de ahora, los pasos 1, 3, 4
//     y 5: sin línea de hoy, ninguna (otra invocación ya reinstaló); si alguno
//     no se cumple, la de hoy de ahora;
//  9. y Aplicar el plan del paso 8: con un error, la línea de hoy del paso 8,
//     la del estado anterior a intentarlo; sin él, la de FR-006.
//
// Las dos líneas nuevas nombran la versión del binario y la que nombra la de
// hoy, las dos tal como están. Nada de esto devuelve un error: lo que falla se
// queda en la línea de hoy, que es lo que se ve (FR-012; research.md D11).
func (r Reinstalacion) Avisar(permiso PermisoDelAviso) (string, bool) {
	linea, hay := Aviso(r.Disco, r.Home, r.Version, r.Empotradas)
	if !hay {
		return "", false
	}

	switch permiso {
	case SoloAvisar:
		return linea, true
	case Ensayar:
		return r.ensayar(linea), true
	case Reinstalar:
		return r.reinstalar(linea)
	}

	// Un permiso que no es ninguno de los tres no permite más que avisar.
	return linea, true
}

// ensayar es la línea de una invocación con Ensayar que tiene la de hoy, linea:
// la de FR-009 si los pasos 3 a 5 se cumplen, y si no, linea (paso 6).
func (r Reinstalacion) ensayar(linea string) string {
	reinstalable, cumple := r.evaluar()
	if !cumple {
		return linea
	}

	return "aviso: kitlegal " + r.Version + " se actualizarían las skills globales, que eran de " +
		reinstalable.nombra + finDeLasLineas
}

// reinstalar es la línea de una invocación con Reinstalar que tiene la de hoy,
// linea, y si hay que darla: los pasos 3 a 9.
func (r Reinstalacion) reinstalar(linea string) (string, bool) {
	if _, cumple := r.evaluar(); !cumple {
		return linea, true
	}

	liberar, tomado := r.Candado()
	if !tomado {
		return linea, true
	}

	defer liberar()

	ahora, hay := Aviso(r.Disco, r.Home, r.Version, r.Empotradas)
	if !hay {
		return "", false
	}

	reinstalable, cumple := r.evaluar()
	if !cumple {
		return ahora, true
	}

	if _, err := Aplicar(reinstalable.plan, r.Escritor); err != nil {
		return ahora, true
	}

	return "aviso: kitlegal " + r.Version + " ha actualizado las skills globales, que eran de " +
		reinstalable.nombra + finDeLasLineas, true
}

// reinstalable es lo que sacan del disco los pasos 3 a 5 cuando se cumplen: el
// plan de la reinstalación y la versión que nombra la línea de hoy.
type reinstalable struct {
	plan   Plan
	nombra string
}

// evaluar aplica, sobre el disco de ahora, los pasos 3 a 5 de
// contracts/aviso-y-reinstalacion.md §2, FR-003 (a) a (c), y dice si se
// cumplen. Quien la llama ya tiene la línea de hoy, así que la versión del
// binario tiene la forma de SemVer (Aviso, FR 073 de H19).
func (r Reinstalacion) evaluar() (reinstalable, bool) {
	manifiesto, ambito, hay := manifiestoDelAviso(r.Disco, r.Home)
	if !hay || ambito.Clase() != AmbitoGlobal {
		return reinstalable{}, false
	}

	if !versionesQueReinstalan(manifiesto, r.Version, r.Empotradas) {
		return reinstalable{}, false
	}

	pedido := pedidoDeReinstalacion(ambito, manifiesto, r.Empotradas, r.Retiradas)
	if len(pedido.Skills) == 0 {
		// Sin ninguna skill que pedir, reinstalar solo cambiaría la versión
		// del manifiesto: callaría el aviso sin poner nada al día.
		return reinstalable{}, false
	}

	plan, err := Planificar(r.Disco, r.Enlazador, pedido, r.Empotradas, r.Retiradas, r.Version)
	if err != nil || algunaSinRetirar(plan) {
		return reinstalable{}, false
	}

	nombra, _ := versionQueDifiere(manifiesto, r.Version, r.Empotradas)

	return reinstalable{plan: plan, nombra: nombra}, true
}

// versionesQueReinstalan aplica la regla de versiones de FR-004 a cada versión
// que compara el aviso —la del manifiesto y la de cada skill que declara y el
// binario empotra— contra la del binario, que tiene la forma de SemVer: todas
// la tienen, ninguna va después de la del binario en precedencia SemVer 2.0.0
// y alguna va antes.
func versionesQueReinstalan(manifiesto Manifiesto, version string, empotradas []SkillEmpotrada) bool {
	comparadas := []string{manifiesto.Version}
	for _, nombre := range declaradasYEmpotradas(manifiesto, empotradas) {
		comparadas = append(comparadas, manifiesto.Skills[nombre].Version)
	}

	algunaAntes := false

	for _, instalada := range comparadas {
		if !FormaSemVer(instalada) {
			return false
		}

		orden := precedencia(instalada, version)
		if orden > 0 {
			return false
		}

		algunaAntes = algunaAntes || orden < 0
	}

	return algunaAntes
}

// pedidoDeReinstalacion es el pedido de la reinstalación en el ámbito
// (data-model §4 de H38): cada skill que el manifiesto declara y el binario
// empotra y, por cada declarada con un nombre de retiradas, sus herederas; en
// orden de nombre, cada una una vez y sin --host. Es el que ValidarInvocacion
// da a kitlegal skills install -g con esas skills.
func pedidoDeReinstalacion(
	ambito Ambito, manifiesto Manifiesto, empotradas []SkillEmpotrada, retiradas Retiradas,
) Pedido {
	skills := declaradasYEmpotradas(manifiesto, empotradas)
	for nombre := range manifiesto.Skills {
		skills = append(skills, retiradas[nombre]...)
	}

	slices.Sort(skills)

	return Pedido{Ambito: ambito, Skills: slices.Compact(skills)}
}

// algunaSinRetirar dice si alguna skill del plan es de un nombre retirado que
// no se retira porque se ha editado: cuenta como tocada (FR-005).
func algunaSinRetirar(plan Plan) bool {
	return slices.ContainsFunc(plan.Skills, func(skill SkillInstalada) bool {
		return skill.Estado == EstadoSinRetirar
	})
}
