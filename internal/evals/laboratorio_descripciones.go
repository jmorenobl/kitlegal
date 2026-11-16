package evals

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"unicode/utf8"

	"go.yaml.in/yaml/v3"

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

// errSinLaClaveDeLaDescripcion es el error de componer un SKILL.md sin
// frontmatter, o sin la clave description en él: no hay líneas en cuyo sitio
// poner el texto. Es un defecto inesperado (ADR 0023), porque las skills de
// skills/ pasan la comprobación de su forma.
var errSinLaClaveDeLaDescripcion = errors.New("el SKILL.md no tiene frontmatter o no tiene en él la clave " +
	nombreDeLaClaveDeLaDescripcion)

// skillConLaDescripcion es el contenido de un SKILL.md con el texto como
// description de su frontmatter (FR-011; contracts/arnes.md §3; data-model §3;
// research D5): las líneas de la clave, las que lineasDelSkill da como de la
// description, dejan su sitio a las que da el emisor de YAML para la clave con
// ese valor, que elige la forma del escalar —plano si puede, entrecomillado si
// no— con que se lee el texto exacto; todas las demás quedan tal cual y en su
// orden. El error, sin contenido, es errSinLaClaveDeLaDescripcion o el del
// emisor.
func skillConLaDescripcion(contenido []byte, texto string) ([]byte, error) {
	deLaClave, err := yaml.Marshal(map[string]string{nombreDeLaClaveDeLaDescripcion: texto})
	if err != nil {
		return nil, fmt.Errorf("la %s «%s» no se puede escribir en YAML: %w", nombreDeLaClaveDeLaDescripcion, texto, err)
	}

	var compuesto bytes.Buffer

	sustituida := false

	for linea, deLaDescripcion := range lineasDelSkill(string(contenido)) {
		switch {
		case !deLaDescripcion:
			compuesto.WriteString(linea)
		case !sustituida:
			compuesto.Write(deLaClave)

			sustituida = true
		}
	}

	if !sustituida {
		return nil, errSinLaClaveDeLaDescripcion
	}

	return compuesto.Bytes(), nil
}

// componerLasDescripciones reescribe, en la copia de las skills de una variante
// en el temporal, la de normativa-producto con INSTALACION=producto-descripciones-cortas,
// el SKILL.md de cada skill con la description de su texto de la variante,
// por su nombre (skillConLaDescripcion; FR-011; contracts/arnes.md §3; data-model
// §3; research D5): lo demás de la copia queda como lo dejó os.CopyFS, byte a
// byte, y la copia no se guarda en DIRECTORIO. Cada skill de la copia, una de
// skills/, tiene su texto, porque la variante que se compone no tiene ningún
// defecto; una sin él es un error, como el de un SKILL.md que no se puede leer,
// componer o escribir (defecto inesperado, ADR 0023).
func componerLasDescripciones(copia string, descripciones map[string]string) error {
	entradas, err := os.ReadDir(copia)
	if err != nil {
		return fmt.Errorf("las skills de la copia %s no se pueden leer: %w", copia, err)
	}

	for _, entrada := range entradas {
		texto, esta := descripciones[entrada.Name()]
		if !esta {
			return fmt.Errorf("la skill %s de la copia %s no tiene descripción en %s", entrada.Name(), copia,
				carpetaDeLasDescripcionesCortas)
		}

		ruta := filepath.Join(copia, entrada.Name(), ficheroDeLaSkillDeLaVariante)

		contenido, err := leerFichero(ruta)
		if err != nil {
			return fmt.Errorf("el %s %s no se puede leer: %w", ficheroDeLaSkillDeLaVariante, ruta, err)
		}

		compuesto, err := skillConLaDescripcion(contenido, texto)
		if err != nil {
			return fmt.Errorf("el %s %s no se puede componer: %w", ficheroDeLaSkillDeLaVariante, ruta, err)
		}

		if err := escribirFichero(ruta, compuesto); err != nil {
			return fmt.Errorf("el %s %s no se puede escribir: %w", ficheroDeLaSkillDeLaVariante, ruta, err)
		}
	}

	return nil
}

// defectoDeLasDescripcionesCortas es la línea de un defecto de la variante: su
// ruta, «: » y el texto del defecto con esos argumentos.
func defectoDeLasDescripcionesCortas(formato string, argumentos ...any) string {
	return carpetaDeLasDescripcionesCortas + ": " + fmt.Sprintf(formato, argumentos...)
}
