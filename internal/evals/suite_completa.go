package evals

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"slices"
	"strings"
)

// Los textos fijos del mensaje con el que falla la comprobación de la suite
// completa de la release (contracts/release.md §2 de H30): la cabecera, con el
// commit; la línea sin ninguna ejecución; y el final, con la orden que lanza la
// suite completa sobre la etiqueta.
const (
	cabeceraDeLaSuiteIncompleta = "release: el commit %s no tiene una ejecución del job de evals de la suite " +
		"completa, en verde y sin evals reutilizadas."
	lineaSinEjecucionesDelCommit = "- ninguna ejecución del flujo evals sobre el commit"
	lineaDeUnaEjecucionMirada    = "- ejecución %d: %s"
	comoLanzarLaSuiteCompleta    = "Lanza la suite completa sobre la etiqueta y, cuando termine en verde, vuelve a " +
		"ejecutar la release:\n  gh workflow run evals.yml --ref %s"

	// falloDeLaPlataformaEnLaRelease es el error de la comprobación cuando una
	// lectura de la plataforma falla: no dice que se lance la suite completa,
	// porque lo que falta no es la ejecución sino la lectura.
	falloDeLaPlataformaEnLaRelease = "release: la comprobación de la suite completa del commit %s no pudo leer la " +
		"plataforma: %w"
)

// evalsDeUnaSkill son los ficheros de las evals de una skill con evals en un
// commit, en orden de nombre: los nombres con que las nombran el origen y las
// series de su informe.
type evalsDeUnaSkill struct {
	skill    string
	ficheros []string
}

// evalsDelCommit son las evals de cada skill con evals en el árbol de un
// commit, en orden de skill (FR-025 de H30): cada subdirectorio de evals/ con
// algún fichero de eval, con los de evals/<skill>/ que tienen la forma de una
// eval, como los lee el conjunto de dependencias (directorioDeEvals). Es un
// error que evals/ o uno de sus subdirectorios no se pueda listar.
func evalsDelCommit(arbol fs.FS) ([]evalsDeUnaSkill, error) {
	skills, err := directoriosDe(arbol, directorioDeLasEvals)
	if err != nil {
		return nil, err
	}

	lectura := calculoDeDependencias{arbol: arbol}

	var evals []evalsDeUnaSkill

	for _, skill := range skills {
		ficheros, _, err := lectura.directorioDeEvals(skill)
		if err != nil {
			return nil, err
		}

		if len(ficheros) > 0 {
			evals = append(evals, evalsDeUnaSkill{skill: skill, ficheros: ficheros})
		}
	}

	return evals, nil
}

// comprobarLaSuiteCompleta comprueba que el commit tiene una ejecución del flujo
// evals de la suite completa en verde y sin evals reutilizadas, la que exige la
// release antes de publicar (FR-025 de H30; data-model §11 de H30;
// contracts/release.md §2 de H30), y da su número. Mira las ejecuciones del flujo
// sobre el commit de la más reciente a la más antigua, por su creación y, a la
// misma, por su número, y vale la primera a la que no descarta ningún motivo
// (motivoQueLaDescarta) con las evals del commit dadas; no lee nada de las
// anteriores. Si ninguna vale, el error es el mensaje de contracts/release.md §2,
// con una línea por ejecución mirada y su motivo, o la de ninguna ejecución, y la
// orden que lanza la suite completa sobre la etiqueta. Si una lectura de la
// plataforma falla, el error es el suyo, envuelto, sin esa orden.
func comprobarLaSuiteCompleta(
	ctx context.Context, plataforma lecturasDeLaPlataforma, commit, etiqueta string, evals []evalsDeUnaSkill,
) (int64, error) {
	ejecuciones, err := plataforma.ejecucionesDelCommit(ctx, commit)
	if err != nil {
		return 0, fmt.Errorf(falloDeLaPlataformaEnLaRelease, commit, err)
	}

	ejecuciones = slices.Clone(ejecuciones)
	slices.SortStableFunc(ejecuciones, func(una, otra ejecucionDelFlujo) int {
		return cmp.Or(otra.Creada.Compare(una.Creada), cmp.Compare(otra.ID, una.ID))
	})

	lineas := make([]string, 0, len(ejecuciones))

	for _, ejecucion := range ejecuciones {
		motivo, err := motivoQueLaDescarta(ctx, plataforma, commit, ejecucion, evals)
		if err != nil {
			return 0, fmt.Errorf(falloDeLaPlataformaEnLaRelease, commit, err)
		}

		if motivo == "" {
			return ejecucion.ID, nil
		}

		lineas = append(lineas, fmt.Sprintf(lineaDeUnaEjecucionMirada, ejecucion.ID, motivo))
	}

	if len(lineas) == 0 {
		lineas = append(lineas, lineaSinEjecucionesDelCommit)
	}

	return 0, fmt.Errorf(cabeceraDeLaSuiteIncompleta+"\n%s\n"+comoLanzarLaSuiteCompleta, commit,
		strings.Join(lineas, "\n"), etiqueta)
}

// motivoQueLaDescarta es la primera condición de data-model §11 de H30 que la
// ejecución no cumple, en palabras, o vacío si las cumple todas: su headSha es
// el commit; ha terminado; el trabajo evals (<skill>) de cada skill con evals en
// el commit está, ha terminado y en success (motivoDeLosTrabajos); y el informe
// de cada una, en el orden de las skills, es de la suite completa con cada eval
// del commit medida y ninguna reutilizada (motivoDelInforme). Lee los trabajos
// solo si las dos primeras se cumplen, y el informe de una skill solo si todos
// los trabajos y los informes anteriores lo hacen. El error es el de la lectura
// que falla: la de los trabajos, y la de un informe si su orden de gh falló o el
// contexto se agotó; un registro que no trae el informe es un motivo.
func motivoQueLaDescarta(
	ctx context.Context, plataforma lecturasDeLaPlataforma, commit string, ejecucion ejecucionDelFlujo,
	evals []evalsDeUnaSkill,
) (string, error) {
	switch {
	case ejecucion.Commit != commit:
		return fmt.Sprintf("su headSha es %s, no el commit", ejecucion.Commit), nil
	case ejecucion.Status != statusTerminado:
		return fmt.Sprintf("no ha terminado (status %s)", ejecucion.Status), nil
	}

	trabajos, err := plataforma.trabajos(ctx, ejecucion.ID)
	if err != nil {
		return "", err
	}

	deCadaSkill, motivo := motivoDeLosTrabajos(trabajos, evals)
	if motivo != "" {
		return motivo, nil
	}

	for posicion, deLaSkill := range evals {
		contenido, err := plataforma.informeDelTrabajo(ctx, ejecucion.ID, deCadaSkill[posicion])

		var deUnaOrden *errorDeUnaOrden

		switch {
		case err == nil:
		case errors.As(err, &deUnaOrden) || ctx.Err() != nil:
			return "", err
		default:
			return fmt.Sprintf("el registro del trabajo %s no trae su informe", nombreDelTrabajoDeLaSkill(deLaSkill.skill)), nil
		}

		if motivo := motivoDelInforme(contenido, deLaSkill); motivo != "" {
			return motivo, nil
		}
	}

	return "", nil
}

// motivoDeLosTrabajos da el id del trabajo evals (<skill>) de cada skill con
// evals, en su orden, o el motivo por el que los trabajos de la ejecución no
// valen: el de la primera skill, en su orden, cuyo trabajo no está, no ha
// terminado o no terminó en success.
func motivoDeLosTrabajos(trabajos []trabajoDeLaEjecucion, evals []evalsDeUnaSkill) ([]int64, string) {
	ids := make([]int64, 0, len(evals))

	for _, deLaSkill := range evals {
		nombre := nombreDelTrabajoDeLaSkill(deLaSkill.skill)

		indice := slices.IndexFunc(trabajos, func(trabajo trabajoDeLaEjecucion) bool { return trabajo.Name == nombre })
		if indice < 0 {
			return nil, "no tiene el trabajo " + nombre
		}

		switch trabajo := trabajos[indice]; {
		case trabajo.Status != statusTerminado:
			return nil, fmt.Sprintf("el trabajo %s no ha terminado (status %s)", nombre, trabajo.Status)
		case trabajo.Conclusion != conclusionDeExito:
			return nil, fmt.Sprintf("el trabajo %s terminó en %s", nombre, trabajo.Conclusion)
		default:
			ids = append(ids, trabajo.ID)
		}
	}

	return ids, ""
}

// motivoDelInforme es el motivo por el que el informe de la skill no vale para
// la release, o vacío si vale: no se puede leer; no tiene el bloque origen; su
// modo no es el de la suite completa; reutiliza alguna eval, la primera en su
// orden; o le falta alguna eval de la skill en el commit, la primera en su
// orden.
func motivoDelInforme(contenido []byte, deLaSkill evalsDeUnaSkill) string {
	var informe Informe
	if err := json.Unmarshal(contenido, &informe); err != nil {
		return fmt.Sprintf("el informe de %s no se puede leer: %v", deLaSkill.skill, err)
	}

	origen := informe.Origen

	switch {
	case origen == nil:
		return fmt.Sprintf("el informe de %s no dice cómo midió sus evals: no tiene el bloque origen", deLaSkill.skill)
	case origen.Modo == origenConReutilizacion:
		return fmt.Sprintf("el informe de %s es de una ejecución con reutilización (%s)", deLaSkill.skill,
			evalsReutilizadasEnPalabras(len(origen.reutilizadas())))
	case origen.Modo != origenDeLaSuiteCompleta:
		return fmt.Sprintf("el informe de %s es de una ejecución de modo %q, no de la suite completa", deLaSkill.skill,
			origen.Modo)
	}

	if reutilizadas := origen.reutilizadas(); len(reutilizadas) > 0 {
		return fmt.Sprintf("el informe de %s reutiliza la eval %s", deLaSkill.skill, reutilizadas[0].Eval)
	}

	for _, fichero := range deLaSkill.ficheros {
		if !slices.ContainsFunc(origen.Evals, func(eval origenDeEval) bool { return eval.Eval == fichero }) {
			return fmt.Sprintf("el informe de %s no tiene la eval %s del commit", deLaSkill.skill, fichero)
		}
	}

	return ""
}

// nombreDelTrabajoDeLaSkill es el del trabajo de la matriz del flujo evals de la skill,
// evals (<skill>).
func nombreDelTrabajoDeLaSkill(skill string) string {
	return prefijoDeUnTrabajoDeEvals + skill + sufijoDeUnTrabajoDeEvals
}

// evalsReutilizadasEnPalabras es el número de evals reutilizadas como lo dice el
// mensaje de la release: «1 eval reutilizada» o «n evals reutilizadas».
func evalsReutilizadasEnPalabras(reutilizadas int) string {
	if reutilizadas == 1 {
		return "1 eval reutilizada"
	}

	return fmt.Sprintf("%d evals reutilizadas", reutilizadas)
}
