package evals

import "path"

// familiaDelLaboratorio es un conjunto de variantes del laboratorio, cada una
// con la misma skill y en un directorio con su nombre (data-model §1 de H40).
type familiaDelLaboratorio struct {
	// skill es el nombre de la skill de cada variante, el name de su
	// frontmatter.
	skill string
	// carpeta es el directorio de las variantes, desde la raíz del
	// repositorio.
	carpeta string
	// variantes son los nombres de las variantes, en orden, empezando por la de
	// control; el de cada una es el de su directorio.
	variantes []string
	// referencias son los ficheros de references/ de la skill, en orden de
	// nombre.
	referencias []string
}

// varianteDeControlDeNormativa es la variante de normativa que es
// skills/normativa, ruta a ruta y byte a byte (FR-002 de H40;
// contracts/variantes.md §2).
const varianteDeControlDeNormativa = "normativa-control"

// carpetaDeNormativa es la de las variantes de normativa dentro de la del
// laboratorio, fuera de skills/, de evals/ y de lo que mide el job (research D1
// de H40).
const carpetaDeNormativa = "normativa"

// familiaDeNormativa es la familia de las variantes de normativa, declarada en
// Go porque la fija entera el hito (research D2 y D3 de H40; data-model §1 de
// H40; contracts/variantes.md §1).
var familiaDeNormativa = familiaDelLaboratorio{
	skill:   "normativa",
	carpeta: path.Join(carpetaDelLaboratorio, carpetaDeNormativa),
	variantes: []string{
		varianteDeControlDeNormativa,
		"normativa-remisiones-en-orden",
		"normativa-cuerpo-corto",
		"normativa-testigo-en-referencia",
	},
	referencias: []string{
		"buscar_la_norma.md",
		"como_se_cita.md",
		"con_herramientas.md",
		"lecturas_que_fallan.md",
		"normas.md",
		"redaccion_en_una_fecha.md",
		"redaccion_modificada.md",
		"varios_bloques.md",
	},
}

// directorioDeLaVariante es el directorio de la variante de ese nombre, desde
// la raíz del repositorio: el que se copia tal cual como directorio de skills
// de sus sesiones (data-model §2 de H40).
func (f familiaDelLaboratorio) directorioDeLaVariante(nombre string) string {
	return path.Join(f.carpeta, nombre)
}

// directorioDeLaSkill es el directorio de la skill de la variante de ese
// nombre, desde la raíz del repositorio.
func (f familiaDelLaboratorio) directorioDeLaSkill(nombre string) string {
	return path.Join(f.directorioDeLaVariante(nombre), carpetaDeLasSkillsDeLaVariante, f.skill)
}
