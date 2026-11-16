package evals

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"unicode/utf8"

	"github.com/jmorenobl/kitlegal/internal/skills"
)

// La variante producto-descripciones-cortas (contracts/variante.md §1; data-model
// §1; research D1): un fichero <skill>.txt por skill de skills/ con su
// description, fuera de skills/, de evals/ y de lo que empotra el binario.
const (
	// carpetaDeLasDescripcionesCortas es el directorio de la variante, relativo a
	// la raíz del repositorio, la ruta con que empieza cada uno de sus defectos.
	carpetaDeLasDescripcionesCortas = carpetaDelLaboratorio + "/producto-descripciones-cortas"

	// extensionDeLaDescripcionCorta es la del fichero de la description de una
	// skill en la variante: <skill>.txt.
	extensionDeLaDescripcionCorta = ".txt"

	// maximoDeUnaDescripcionCorta y maximoDeLasDescripcionesCortas son los límites
	// en runas de cada description de la variante y de la suma de todas (FR-001,
	// FR-002).
	maximoDeUnaDescripcionCorta    = 450
	maximoDeLasDescripcionesCortas = 1700
)

// descripcionesCortas lee la variante producto-descripciones-cortas de la raíz
// (FR-002; contracts/variante.md §3; research D3) y devuelve el texto de la
// description de cada skill de skills/ que la tiene, por su nombre, y sus
// defectos, uno por línea y cada uno con la ruta de la variante delante. Las
// skills son los subdirectorios de skills/ (la regla de
// copiarLasDemasSkillsDelArbol) y el texto de cada una es su fichero <skill>.txt
// sin el salto de línea final (textoDeLaVariante). Los defectos van en este
// orden: por cada skill, en orden de nombre, el de que le falte su fichero o, si
// lo tiene, el de su longitud y el de su etiqueta XML, la primera que da
// skills.PrimeraEtiquetaXML; después, por cada entrada de la variante que no es
// <skill>.txt de una skill, en orden de nombre; y al final, el de la suma de las
// que hay. Los caracteres se cuentan como runas, no como bytes (ADR 0046, punto
// 2). El error, sin textos ni defectos, es el de un directorio o un fichero que
// no se puede leer, también el de uno que no existe: no es un defecto de la
// variante (ADR 0023).
func descripcionesCortas(raiz string) (map[string]string, []string, error) {
	arbol := filepath.Join(raiz, directorioDeLasSkills)

	deSkills, err := os.ReadDir(arbol)
	if err != nil {
		return nil, nil, fmt.Errorf("las skills del árbol %s no se pueden leer: %w", arbol, err)
	}

	directorio := filepath.Join(raiz, filepath.FromSlash(carpetaDeLasDescripcionesCortas))

	deLaVariante, err := os.ReadDir(directorio)
	if err != nil {
		return nil, nil, fmt.Errorf("la variante %s no se puede leer: %w", directorio, err)
	}

	entradas := make([]string, 0, len(deLaVariante))
	for _, entrada := range deLaVariante {
		entradas = append(entradas, entrada.Name())
	}

	var (
		ficheros []string
		defectos []string
		suma     int
	)

	textos := map[string]string{}

	for _, entrada := range deSkills {
		if !entrada.IsDir() {
			continue
		}

		skill := entrada.Name()
		fichero := skill + extensionDeLaDescripcionCorta
		ficheros = append(ficheros, fichero)

		if !slices.Contains(entradas, fichero) {
			defectos = append(defectos, defectoDeLasDescripcionesCortas("falta la descripción de %s", skill))

			continue
		}

		texto, err := textoDeLaVariante(directorio, fichero)
		if err != nil {
			return nil, nil, err
		}

		textos[skill] = texto

		longitud := utf8.RuneCountInString(texto)
		suma += longitud

		if longitud > maximoDeUnaDescripcionCorta {
			defectos = append(defectos, defectoDeLasDescripcionesCortas("la descripción de %s tiene %d caracteres (máximo %d)",
				skill, longitud, maximoDeUnaDescripcionCorta))
		}

		if etiqueta := skills.PrimeraEtiquetaXML(texto); etiqueta != "" {
			defectos = append(defectos, defectoDeLasDescripcionesCortas(
				"la descripción de %s tiene la etiqueta XML «%s» (no admite ninguna)", skill, etiqueta))
		}
	}

	for _, entrada := range entradas {
		if !slices.Contains(ficheros, entrada) {
			defectos = append(defectos, defectoDeLasDescripcionesCortas(
				"«%s» no es la descripción de ninguna skill de %s/", entrada, directorioDeLasSkills))
		}
	}

	if suma > maximoDeLasDescripcionesCortas {
		defectos = append(defectos, defectoDeLasDescripcionesCortas("las descripciones suman %d caracteres (máximo %d)",
			suma, maximoDeLasDescripcionesCortas))
	}

	return textos, defectos, nil
}

// defectoDeLasDescripcionesCortas es la línea de un defecto de la variante: su
// ruta, «: » y el texto del defecto con esos argumentos.
func defectoDeLasDescripcionesCortas(formato string, argumentos ...any) string {
	return carpetaDeLasDescripcionesCortas + ": " + fmt.Sprintf(formato, argumentos...)
}
