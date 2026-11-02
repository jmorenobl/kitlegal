package skills

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// ficheroDeLasNormas es el YAML de datos, dentro de data/, de cuyos valores de
// vertical salen los paquetes de materia (contracts/paquetes.md §1 de H32).
const ficheroDeLasNormas = "normas.yaml"

// defectoSinEnlazar va delante del fichero de la referencia de un paquete en una
// skill cuyo SKILL.md no lleva las marcas de los paquetes (contracts/paquetes.md
// §2; FR-023 de H32).
const defectoSinEnlazar = "referencia de paquete sin enlazar: "

// regionDeMarcas es una región de un SKILL.md que la sincronía escribe entre dos
// marcas, cada una en su línea, fuera de la tabla de comandos.
type regionDeMarcas struct {
	// nombre es lo que la región lleva, como lo nombran sus defectos.
	nombre string

	// inicio y fin son las líneas de sus dos marcas, sin el salto.
	inicio, fin string
}

// regionDeLosPaquetes y regionDeLaBase son las de los paquetes de materia
// (contracts/paquetes.md §2 y §4; FR-001 punto 3 y FR-023 de H32).
var (
	// regionDeLosPaquetes es la de los enlaces a las referencias de los paquetes
	// de una skill de capacidad.
	regionDeLosPaquetes = regionDeMarcas{
		nombre: "los paquetes",
		inicio: "<!-- paquetes: lo genera make skills-sync -->",
		fin:    "<!-- fin de los paquetes -->",
	}

	// regionDeLaBase es la de la lista de los paquetes de la skill de entrada,
	// con las skills que tienen la referencia de cada uno.
	regionDeLaBase = regionDeMarcas{
		nombre: "los paquetes de la base",
		inicio: "<!-- paquetes de la base: lo genera make skills-sync -->",
		fin:    "<!-- fin de los paquetes de la base -->",
	}
)

// entreMarcas es lo que hay entre las dos marcas de una región de un SKILL.md:
// desde el byte que sigue a la línea de la de inicio hasta el primero de la línea
// de la de fin.
type entreMarcas struct {
	desde, hasta int
}

// buscar es lo que hay entre las marcas de la región en el contenido, y si las
// lleva: sin ninguna de las dos, no la lleva, y no es un defecto. Una línea es una
// marca solo si es exactamente la marca, sin espacios ni retorno de carro. Es el
// defecto de las marcas que no van en pareja que falte una de las dos, que una
// esté en más de una línea o que la de fin vaya antes que la de inicio.
func (r regionDeMarcas) buscar(contenido []byte) (entreMarcas, bool, error) {
	inicios, fines := lineasQueSon(contenido, r.inicio), lineasQueSon(contenido, r.fin)

	var falla string

	switch {
	case len(inicios) == 0 && len(fines) == 0:
		return entreMarcas{}, false, nil
	case len(inicios) == 0:
		falla = "falta la de inicio"
	case len(fines) == 0:
		falla = "falta la de fin"
	case len(inicios) > 1:
		falla = enVariasLineas("inicio", inicios)
	case len(fines) > 1:
		falla = enVariasLineas("fin", fines)
	case fines[0].numero < inicios[0].numero:
		falla = fmt.Sprintf("la de fin, en la línea %d, va antes que la de inicio, en la línea %d", fines[0].numero,
			inicios[0].numero)
	default:
		return entreMarcas{desde: inicios[0].final, hasta: fines[0].comienzo}, true, nil
	}

	return entreMarcas{}, false, fmt.Errorf("las marcas de %s no van en pareja: %s", r.nombre, falla)
}

// lineasQueSon son las líneas del contenido que, sin su salto, son exactamente el
// texto, en su orden.
func lineasQueSon(contenido []byte, texto string) []lineaDeMarca {
	var lineas []lineaDeMarca

	numero, comienzo := 0, 0

	for linea := range bytes.Lines(contenido) {
		numero++

		if string(bytes.TrimSuffix(linea, []byte("\n"))) == texto {
			lineas = append(lineas, lineaDeMarca{numero: numero, comienzo: comienzo, final: comienzo + len(linea)})
		}

		comienzo += len(linea)
	}

	return lineas
}

// enVariasLineas dice en qué líneas está la marca que aparece más de una vez.
func enVariasLineas(marca string, lineas []lineaDeMarca) string {
	numeros := make([]string, 0, len(lineas))

	for _, linea := range lineas {
		numeros = append(numeros, strconv.Itoa(linea.numero))
	}

	return fmt.Sprintf("la de %s aparece %d veces, en las líneas %s", marca, len(lineas), enumerar(numeros))
}

// sustituir es el contenido con lineas en lugar de lo que hay entre las marcas;
// las líneas de las marcas y lo de fuera quedan byte a byte igual.
func (e entreMarcas) sustituir(contenido, lineas []byte) []byte {
	return slices.Concat(contenido[:e.desde], lineas, contenido[e.hasta:])
}

// paquetes son los paquetes de materia de un repositorio y las skills que tienen
// la referencia de cada uno (contracts/paquetes.md §1; data-model §3; research
// D3 de H32).
type paquetes struct {
	// nombres son los valores de vertical de data/normas.yaml, cada uno una vez y
	// en orden.
	nombres []string

	// deLaSkill son, por skill, los paquetes con referencia en ella, en el orden de
	// sus nombres.
	deLaSkill map[string][]string

	// conReferencia son, por paquete, las skills que tienen su referencia, en el
	// orden de sus nombres.
	conReferencia map[string][]string

	// defecto es el de data/normas.yaml, que no se lee como tabla de normas: con
	// él no se sabe ningún paquete.
	defecto error
}

// leerPaquetes lee los paquetes de materia del repositorio de raíz, de skills en
// el orden de sus nombres (contracts/paquetes.md §1; research D3 de H32): los
// valores de vertical de data/normas.yaml; y la referencia de uno en una skill es
// references/<paquete>.md, un fichero regular. Sin data/normas.yaml no hay
// ninguno; con una tabla de normas que no se lee, tampoco, y sus defectos son el
// defecto de los paquetes. El error queda para lo que no se puede leer:
// data/normas.yaml o el references/ de una skill.
func leerPaquetes(raiz string, skills []string) (paquetes, error) {
	contenido, err := leerFichero(filepath.Join(raiz, carpetaDeLosDatos, ficheroDeLasNormas))

	switch {
	case errors.Is(err, fs.ErrNotExist):
		return paquetes{}, nil
	case err != nil:
		return paquetes{}, fmt.Errorf("%s/%s no se puede leer: %w", carpetaDeLosDatos, ficheroDeLasNormas, err)
	}

	// Con defectos, LeerNormas no da ninguna norma: no hay ningún paquete, y la
	// skill que lleve las marcas tiene el defecto de la tabla (regenerar).
	normas, defecto := LeerNormas(contenido)

	var nombres []string

	for _, norma := range normas {
		if norma.Vertical != "" && !slices.Contains(nombres, norma.Vertical) {
			nombres = append(nombres, norma.Vertical)
		}
	}

	slices.Sort(nombres)

	leidos := paquetes{
		nombres:       nombres,
		deLaSkill:     make(map[string][]string, len(skills)),
		conReferencia: make(map[string][]string, len(nombres)),
		defecto:       defecto,
	}

	arbol := os.DirFS(raiz)

	for _, skill := range skills {
		deLaSkill, err := paquetesConReferencia(arbol, skill, nombres)
		if err != nil {
			return paquetes{}, err
		}

		leidos.deLaSkill[skill] = deLaSkill

		for _, paquete := range deLaSkill {
			leidos.conReferencia[paquete] = append(leidos.conReferencia[paquete], skill)
		}
	}

	return leidos, nil
}

// paquetesConReferencia son los paquetes de nombres, en su orden, cuya
// referencia está en la skill del árbol, que es la raíz de un repositorio: un
// fichero regular references/<paquete>.md. Sin references/ no hay ninguno; el
// error queda para un references/ que no se puede listar.
func paquetesConReferencia(arbol fs.FS, skill string, nombres []string) ([]string, error) {
	carpeta := path.Join(carpetaDeLasSkills, skill, carpetaDeLasReferencias)

	entradas, err := fs.ReadDir(arbol, carpeta)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("%s no se puede listar: %w", carpeta, err)
	}

	var conReferencia []string

	for _, paquete := range nombres {
		if slices.ContainsFunc(entradas, func(entrada fs.DirEntry) bool {
			return esFicheroMd(entrada) && entrada.Name() == paquete+extensionDeLasReferencias
		}) {
			conReferencia = append(conReferencia, paquete)
		}
	}

	return conReferencia, nil
}

// regenerar es el SKILL.md de la skill, de contenido, con lo que la sincronía
// escribe entre las marcas de cada región de los paquetes que lleva
// (contracts/paquetes.md §2 y §4; FR-001 punto 3 y FR-023 de H32), con los
// defectos de sus marcas y los de los datos de los que salen:
//
//   - entre las de los paquetes, el enlace a la referencia de cada paquete que la
//     skill tiene, en el orden de sus nombres;
//   - entre las de la base, la lista de los paquetes con referencia en alguna
//     skill, en el orden de sus nombres;
//   - las marcas de una región que no van en pareja son un defecto, y la región
//     queda como está;
//   - sin las marcas de los paquetes, cada referencia de un paquete que la skill
//     tiene es un defecto: nada la enlaza;
//   - con data/normas.yaml que no se lee, cada región con sus marcas queda como
//     está, y el defecto es el de la tabla de normas, con su fichero delante.
//
// Los defectos de las marcas llevan SKILL.md delante.
func (p paquetes) regenerar(skill string, contenido []byte) ([]byte, []*DefectoDeSkill, []*DefectoDeSkill) {
	var deLasMarcas []*DefectoDeSkill

	conMarcas := false

	for _, region := range []regionDeMarcas{regionDeLosPaquetes, regionDeLaBase} {
		entre, lleva, err := region.buscar(contenido)

		switch {
		case err != nil:
			deLasMarcas = append(deLasMarcas, defectosDelError(skill, ficheroDeLaSkill, err)...)
		case !lleva && region == regionDeLosPaquetes:
			for _, paquete := range p.deLaSkill[skill] {
				deLasMarcas = append(deLasMarcas, &DefectoDeSkill{
					Skill:   skill,
					Defecto: ficheroDeLaSkill + ": " + defectoSinEnlazar + paquete + extensionDeLasReferencias,
				})
			}
		case lleva:
			conMarcas = true

			if p.defecto == nil {
				contenido = entre.sustituir(contenido, p.lineas(region, skill))
			}
		}
	}

	if conMarcas && p.defecto != nil {
		return contenido, deLasMarcas, defectosDelError(skill, carpetaDeLosDatos+"/"+ficheroDeLasNormas, p.defecto)
	}

	return contenido, deLasMarcas, nil
}

// lineas es lo que la sincronía escribe entre las marcas de la región en el
// SKILL.md de la skill: en la de los paquetes, una línea por cada paquete con
// referencia en ella; en la de la base, una por cada paquete con referencia en
// alguna skill, con esas skills como código en línea, separadas por «, ».
func (p paquetes) lineas(region regionDeMarcas, skill string) []byte {
	var lineas bytes.Buffer

	if region == regionDeLosPaquetes {
		for _, paquete := range p.deLaSkill[skill] {
			lineas.WriteString("- Si la pregunta es de lo " + paquete + ", lee [" + paquete + "](" +
				carpetaDeLasReferencias + "/" + paquete + extensionDeLasReferencias + ").\n")
		}

		return lineas.Bytes()
	}

	for _, paquete := range p.nombres {
		skills := p.conReferencia[paquete]
		if len(skills) == 0 {
			continue
		}

		lineas.WriteString("- lo " + paquete + ": `" + strings.Join(skills, "`, `") + "`.\n")
	}

	return lineas.Bytes()
}
