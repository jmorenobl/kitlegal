package evals

import (
	"errors"
	"fmt"
	"path"
	"slices"
)

// familiaDelLaboratorio es un conjunto de variantes del laboratorio, cada una
// con la misma skill, y lo que el arnés del sondeo necesita de ellas (data-model
// §1 de H40 y de H41): una invocación del arnés es de una sola familia (FR-010 de
// H40). Es del producto si tiene variante del producto.
type familiaDelLaboratorio struct {
	// skill es el nombre de la skill de cada variante, el name de su
	// frontmatter.
	skill string
	// carpeta es el directorio de las variantes versionadas, desde la raíz del
	// repositorio.
	carpeta string
	// variantes son las variantes, en el orden de su declaración, empezando por
	// la de control, cada una con los modos en que se mide; el nombre de cada
	// versionada es el de su directorio.
	variantes []varianteDelLaboratorio
	// producto es el nombre de la variante que se instala desde skills/<skill>
	// en cada sondeo que la pide, una de las variantes sin directorio en la
	// carpeta (data-model §1 y §3); en la de H39 va vacío. Las variantes
	// versionadas de la familia son las suyas menos esta.
	producto string
	// referencias son los ficheros de references/ de la skill de las variantes
	// versionadas, en orden de nombre.
	referencias []string
	// evals es el directorio de evals de los repartos, desde la raíz del
	// repositorio: cada sesión se prepara con las consultas de todas las suyas.
	evals string
	// evalsDelSondeo son los ficheros de las evals de evals que planifica el
	// sondeo, en su orden; sin ninguno, las planifica todas, en orden de nombre.
	evalsDelSondeo []string
	// conLaRegla es la variante con regla testigo, la única en cuya fila y cuyo
	// detalle la tabla de la familia escribe si la respuesta la cumple
	// (contracts/tabla.md §3 y §4 de H40). En la de H39, cuya tabla la escribe en
	// todas las variantes (contracts/tabla.md §3 de H39), va vacía.
	conLaRegla string
}

// Las skills del producto que tienen familia en el laboratorio (data-model §1):
// la de cada una es la de skills/<skill>, sus variantes versionadas están en
// laboratorio/<skill>/, fuera de skills/, de evals/ y de lo que mide el job
// (research D1 de H40; contracts/variantes.md §1), y las evals de sus repartos
// son las de evals/<skill>, que son las del job.
const (
	skillDeNormativa                 = "normativa"
	skillDeLaFamiliaDePlazos         = "plazos-y-territorio"
	skillDeLaFamiliaDeJurisprudencia = "jurisprudencia"
)

// varianteConLaReglaDeNormativa es la variante de normativa cuya referencia
// con_herramientas.md pide la regla testigo (contracts/variantes.md §5 de H40).
const varianteConLaReglaDeNormativa = "normativa-testigo-en-referencia"

// varianteDeControlDeNormativa es la variante de control de normativa, la copia
// de skills/normativa en v0.9.0 (FR-020; data-model §2).
const varianteDeControlDeNormativa = "normativa-control"

// familiasDelProducto son las familias de las skills del producto, declaradas en
// Go porque las fija enteras el hito (research D9; data-model §1;
// contracts/variantes.md §1), en este orden: normativa, plazos-y-territorio y
// jurisprudencia. Cada una tiene su variante de control, la copia de
// skills/<skill> en v0.9.0, y la del producto, que se instala desde
// skills/<skill>, las dos en los dos modos; la de normativa, además, las tres de
// H40, solo en el modo herramienta, y la variante con regla testigo, la única en
// que la tabla la escribe (contracts/tabla.md §3 de H40). Cada una, también, las
// cuatro evals que planifica su sondeo (research D12).
var familiasDelProducto = []familiaDelLaboratorio{
	{
		skill:   skillDeNormativa,
		carpeta: path.Join(carpetaDelLaboratorio, skillDeNormativa),
		variantes: []varianteDelLaboratorio{
			{Nombre: varianteDeControlDeNormativa, Modos: []Modo{ModoOrden, ModoHerramienta}},
			{Nombre: "normativa-producto", Modos: []Modo{ModoOrden, ModoHerramienta}},
			{Nombre: "normativa-remisiones-en-orden", Modos: []Modo{ModoHerramienta}},
			{Nombre: "normativa-cuerpo-corto", Modos: []Modo{ModoHerramienta}},
			{Nombre: varianteConLaReglaDeNormativa, Modos: []Modo{ModoHerramienta}},
		},
		producto:   "normativa-producto",
		conLaRegla: varianteConLaReglaDeNormativa,
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
	},
	{
		skill:   skillDeLaFamiliaDePlazos,
		carpeta: path.Join(carpetaDelLaboratorio, skillDeLaFamiliaDePlazos),
		variantes: []varianteDelLaboratorio{
			{Nombre: "plazos-y-territorio-control", Modos: []Modo{ModoOrden, ModoHerramienta}},
			{Nombre: "plazos-y-territorio-producto", Modos: []Modo{ModoOrden, ModoHerramienta}},
		},
		producto: "plazos-y-territorio-producto",
		referencias: []string{
			"administrativo.md",
			"calculo_de_plazos.md",
			"jerarquia_normativa.md",
			"municipios.md",
		},
		evals: path.Join(directorioDeLasEvals, skillDeLaFamiliaDePlazos),
		evalsDelSondeo: []string{
			"02-territorio-municipio-no-cubierto.yaml",
			"05-plazo-reposicion-leganes.yaml",
			"10-plazo-reposicion-anterior-a-la-lpac.yaml",
			"11-plazo-reposicion-sin-leer-el-precepto.yaml",
		},
	},
	{
		skill:   skillDeLaFamiliaDeJurisprudencia,
		carpeta: path.Join(carpetaDelLaboratorio, skillDeLaFamiliaDeJurisprudencia),
		variantes: []varianteDelLaboratorio{
			{Nombre: "jurisprudencia-control", Modos: []Modo{ModoOrden, ModoHerramienta}},
			{Nombre: "jurisprudencia-producto", Modos: []Modo{ModoOrden, ModoHerramienta}},
		},
		producto: "jurisprudencia-producto",
		referencias: []string{
			"consulta_preparada.md",
			"documento_aportado.md",
		},
		evals: path.Join(directorioDeLasEvals, skillDeLaFamiliaDeJurisprudencia),
		evalsDelSondeo: []string{
			"01-existe-con-numero-y-fecha.yaml",
			"03-por-materia.yaml",
			"04-documento-pegado.yaml",
			"06-documento-que-no-es-el-pedido.yaml",
		},
	},
}

// versionada es la familia con solo sus variantes versionadas, las de un
// directorio de su carpeta, en su orden, y sin variante del producto; la de H39,
// que no tiene, es la misma.
func (f familiaDelLaboratorio) versionada() familiaDelLaboratorio {
	versionada := f
	versionada.variantes = slices.DeleteFunc(slices.Clone(f.variantes), func(variante varianteDelLaboratorio) bool {
		return variante.Nombre == f.producto
	})
	versionada.producto = ""

	return versionada
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
