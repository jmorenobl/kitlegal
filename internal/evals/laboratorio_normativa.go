package evals

import (
	"errors"
	"fmt"
	"path"
	"slices"
)

// familiaDelLaboratorio es un conjunto de variantes del laboratorio, cada una
// con la misma skill y en un directorio con su nombre, y lo que el arnés del
// sondeo necesita de ellas (data-model §1 de H40): una invocación del arnés es
// de una sola familia (FR-010 de H40).
type familiaDelLaboratorio struct {
	// skill es el nombre de la skill de cada variante, el name de su
	// frontmatter.
	skill string
	// carpeta es el directorio de las variantes, desde la raíz del
	// repositorio.
	carpeta string
	// variantes son las variantes, en el orden de su declaración, empezando por
	// la de control, cada una con los modos en que se mide; el nombre de cada
	// una es el de su directorio.
	variantes []varianteDelLaboratorio
	// referencias son los ficheros de references/ de la skill, en orden de
	// nombre.
	referencias []string
	// evals es el directorio de evals de los repartos, desde la raíz del
	// repositorio: cada sesión se prepara con las consultas de todas las suyas.
	evals string
	// evalsDelSondeo son los ficheros de las evals de evals que planifica el
	// sondeo, en su orden; sin ninguno, las planifica todas, en orden de nombre.
	evalsDelSondeo []string
}

// varianteDeControlDeNormativa es la variante de normativa que es
// skills/normativa, ruta a ruta y byte a byte (FR-002 de H40;
// contracts/variantes.md §2).
const varianteDeControlDeNormativa = "normativa-control"

// skillDeNormativa es la skill de las variantes de normativa, la de
// skills/normativa, cuyas evals son las de evals/normativa.
const skillDeNormativa = "normativa"

// carpetaDeNormativa es la de las variantes de normativa dentro de la del
// laboratorio, fuera de skills/, de evals/ y de lo que mide el job (research D1
// de H40).
const carpetaDeNormativa = "normativa"

// familiaDeNormativa es la familia de las variantes de normativa, declarada en
// Go porque la fija entera el hito (research D2, D3 y D7 de H40; data-model §1 de
// H40; contracts/variantes.md §1; contracts/arnes.md §3 de H40): la de control,
// en los dos modos, y las otras tres, solo en el modo herramienta; las evals de
// sus repartos, las de evals/normativa, que son las del job; y las cuatro que
// planifica el sondeo.
var familiaDeNormativa = familiaDelLaboratorio{
	skill:   skillDeNormativa,
	carpeta: path.Join(carpetaDelLaboratorio, carpetaDeNormativa),
	variantes: []varianteDelLaboratorio{
		{Nombre: varianteDeControlDeNormativa, Modos: []Modo{ModoOrden, ModoHerramienta}},
		{Nombre: "normativa-remisiones-en-orden", Modos: []Modo{ModoHerramienta}},
		{Nombre: "normativa-cuerpo-corto", Modos: []Modo{ModoHerramienta}},
		{Nombre: "normativa-testigo-en-referencia", Modos: []Modo{ModoHerramienta}},
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
	evals: path.Join(directorioDeLasEvals, skillDeNormativa),
	evalsDelSondeo: []string{
		"04-lgt-prescripcion.yaml",
		"05-trlrhl-impuestos-municipales.yaml",
		"15-irpf-rendimientos-por-materia.yaml",
		"22-lcsp-contrato-menor-en-2019.yaml",
	},
}

// nombres son los nombres de las variantes de la familia, en su orden.
func (f familiaDelLaboratorio) nombres() []string {
	nombres := make([]string, 0, len(f.variantes))
	for _, variante := range f.variantes {
		nombres = append(nombres, variante.Nombre)
	}

	return nombres
}

// tiene dice si la familia tiene una variante con ese nombre.
func (f familiaDelLaboratorio) tiene(nombre string) bool {
	return slices.Contains(f.nombres(), nombre)
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

// evalsDelPlan son las evals que planifica el sondeo de la familia, de las
// leídas de su directorio de evals (contracts/arnes.md §3.2 de H40): las de
// evalsDelSondeo, en su orden, o, sin ninguna, todas las leídas, en su orden. El
// error nombra cada eval del sondeo que no está entre las leídas, una por línea.
func (f familiaDelLaboratorio) evalsDelPlan(leidas []Eval) ([]Eval, error) {
	if len(f.evalsDelSondeo) == 0 {
		return leidas, nil
	}

	var (
		plan   []Eval
		faltan []error
	)

	for _, fichero := range f.evalsDelSondeo {
		posicion := slices.IndexFunc(leidas, func(eval Eval) bool { return eval.Fichero == fichero })
		if posicion < 0 {
			faltan = append(faltan, fmt.Errorf("%s no tiene la eval %s del sondeo", f.evals, fichero))

			continue
		}

		plan = append(plan, leidas[posicion])
	}

	if err := errors.Join(faltan...); err != nil {
		return nil, err
	}

	return plan, nil
}
