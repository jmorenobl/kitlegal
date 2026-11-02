package skills_test

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/skills"
)

// Las marcas de los paquetes de un SKILL.md de capacidad y las de la lista de la
// skill de entrada, cada una con su salto de línea (contracts/paquetes.md §2 y
// §4).
const (
	inicioDeLosPaquetes         = "<!-- paquetes: lo genera make skills-sync -->\n"
	finDeLosPaquetes            = "<!-- fin de los paquetes -->\n"
	inicioDeLosPaquetesDeLaBase = "<!-- paquetes de la base: lo genera make skills-sync -->\n"
	finDeLosPaquetesDeLaBase    = "<!-- fin de los paquetes de la base -->\n"
)

// normasDePaquetes es la tabla de normas de los árboles de TestPaquetes: dos
// paquetes, fiscal antes que administrativo en el documento para que el orden de
// los nombres no sea el suyo, lo administrativo con dos normas, y una norma de
// la base, sin vertical.
const normasDePaquetes = "normas:\n" +
	"  BOE-A-2003-23186:\n" +
	"    titulo: Norma de prueba de lo fiscal.\n" +
	"    rango: Ley\n" +
	"    materias:\n" +
	"      - tributos\n" +
	"    vertical: fiscal\n" +
	"  BOE-A-2015-10565:\n" +
	"    titulo: Norma de prueba de lo administrativo.\n" +
	"    rango: Ley\n" +
	"    materias:\n" +
	"      - procedimiento\n" +
	"    vertical: administrativo\n" +
	"  BOE-A-1998-16718:\n" +
	"    titulo: Otra norma de prueba de lo administrativo.\n" +
	"    rango: Ley\n" +
	"    materias:\n" +
	"      - contencioso\n" +
	"    vertical: administrativo\n" +
	"  BOE-A-1978-31229:\n" +
	"    titulo: Norma de prueba de la base.\n" +
	"    rango: Constitución\n" +
	"    materias:\n" +
	"      - constitucional\n"

// Las líneas que la sincronía escribe entre las marcas, byte a byte: el enlace de
// lo administrativo, el de contracts/paquetes.md §2 (93 caracteres y su salto),
// el de lo fiscal y la lista de la entrada de §4.
const (
	enlaceDeLoAdministrativo = "- Si la pregunta es de lo administrativo, lee [administrativo](references/administrativo.md).\n"
	enlaceDeLoFiscal         = "- Si la pregunta es de lo fiscal, lee [fiscal](references/fiscal.md).\n"
	listaDeLoAdministrativo  = "- lo administrativo: `plazos-y-territorio`.\n"
)

// referenciaDeUnPaquete es una referencia de un paquete, escrita a mano.
const referenciaDeUnPaquete = "# La materia del paquete\n\nLo que la skill lee si la pregunta es de su materia.\n"

// skillMdDeCapacidad es el SKILL.md de una skill de capacidad de nombre, con las
// marcas de los paquetes y entreMarcas entre ellas: la marca de inicio va en la
// línea 7.
func skillMdDeCapacidad(nombre, entreMarcas string) string {
	return "---\n" +
		"name: " + nombre + "\n" +
		"description: Skill de prueba que recibe las referencias de los paquetes.\n" +
		"---\n" +
		"# Capacidad\n" +
		"\n" +
		inicioDeLosPaquetes + entreMarcas + finDeLosPaquetes +
		"\n" +
		"## Reglas\n"
}

// skillMdDeLaEntrada es el SKILL.md de la skill de entrada, con las marcas de la
// lista de los paquetes y entreMarcas entre ellas: la marca de inicio va en la
// línea 9.
func skillMdDeLaEntrada(entreMarcas string) string {
	return "---\n" +
		"name: consulta-juridica\n" +
		"description: Skill de prueba de entrada que nombra las de capacidad.\n" +
		"---\n" +
		"# Entrada\n" +
		"\n" +
		"## 3. Qué cubre la base\n" +
		"\n" +
		inicioDeLosPaquetesDeLaBase + entreMarcas + finDeLosPaquetesDeLaBase +
		"\n" +
		"## 4. Lo que no cubre\n"
}

// arbolDePaquetes es un repositorio temporal con normasDePaquetes y tres skills
// sin nada generado, con las marcas vacías: consulta-juridica, la de entrada;
// normativa, sin la referencia de ningún paquete, como en el repositorio; y
// plazos-y-territorio, con la de lo administrativo y una a mano que no es de
// ningún paquete, enlazada desde su cuerpo.
func arbolDePaquetes(t *testing.T) string {
	t.Helper()

	raiz := t.TempDir()
	escribirFicheroDePrueba(t, filepath.Join(raiz, "data", "normas.yaml"), normasDePaquetes)
	escribirFicheroDePrueba(t, rutaDeSkill(raiz, "consulta-juridica", "SKILL.md"), skillMdDeLaEntrada(""))
	escribirFicheroDePrueba(t, rutaDeSkill(raiz, "normativa", "SKILL.md"), skillMdDeCapacidad("normativa", ""))
	escribirFicheroDePrueba(t, rutaDeSkill(raiz, "plazos-y-territorio", "SKILL.md"),
		skillMdDeCapacidad("plazos-y-territorio", ""))
	escribirReferenciaDeUnPaquete(t, raiz, "plazos-y-territorio", "administrativo")
	escribirFicheroDePrueba(t, rutaDeSkill(raiz, "plazos-y-territorio", "references", "como_se_cita.md"),
		referenciaAMano)
	enlazarDesdeSkillMd(t, raiz, "plazos-y-territorio", "como_se_cita.md")

	return raiz
}

// escribirReferenciaDeUnPaquete escribe la referencia del paquete en la skill del
// repositorio de raíz: references/<paquete>.md, a mano.
func escribirReferenciaDeUnPaquete(t *testing.T, raiz, skill, paquete string) {
	t.Helper()

	escribirFicheroDePrueba(t, rutaDeSkill(raiz, skill, "references", paquete+".md"), referenciaDeUnPaquete)
}

// TestPaquetes fija lo que la sincronía genera entre las marcas de los paquetes
// (contracts/paquetes.md §1, §2 y §4; data-model §3; research D3 y D10; FR-001
// punto 3, FR-022 y FR-023 de H32), sobre repositorios temporales: los paquetes
// son los valores de vertical de data/normas.yaml y la referencia de uno en una
// skill es references/<paquete>.md; entre las marcas de un SKILL.md de capacidad
// va un enlace por cada paquete con referencia en la skill, y entre las de la
// entrada, una línea por paquete con las skills que tienen su referencia, los dos
// en el orden de los nombres; el texto entre las marcas que no es el generado es
// la deriva contenido-distinto de SKILL.md; y la referencia de un paquete en una
// skill sin las marcas, las marcas que no van en pareja y una tabla de normas con
// defectos son defectos de la skill.
func TestPaquetes(t *testing.T) {
	t.Parallel()

	t.Run("un-paquete", probarUnPaquete)
	t.Run("dos-paquetes-en-orden", probarDosPaquetesEnOrden)
	t.Run("lista-de-la-entrada", probarLaListaDeLaEntrada)
	t.Run("texto-editado-a-mano", probarElTextoEditadoAMano)
	t.Run("referencia-sin-marcas", probarLaReferenciaSinMarcas)
	t.Run("normas-con-defectos", probarLasNormasConDefectos)
	t.Run("sin-tabla-de-normas", probarSinTablaDeNormas)
	t.Run("marcas-sin-pareja", probarLasMarcasSinPareja)
	t.Run("sin-poder-leer", probarSinPoderLeerLosPaquetes)
}

// probarUnPaquete fija la línea de lo administrativo byte a byte en
// plazos-y-territorio, la lista de la entrada con ella y las marcas vacías de
// normativa, que no tiene la referencia de ningún paquete; lo fiscal, sin
// referencia en ninguna skill, no tiene línea. Comparar da la deriva de los dos
// SKILL.md que cambian, y Escribir deja el árbol sin derivas.
func probarUnPaquete(t *testing.T) {
	t.Parallel()

	linea, conSalto := strings.CutSuffix(enlaceDeLoAdministrativo, "\n")
	require.True(t, conSalto)
	assert.Equal(t, 93, utf8.RuneCountInString(linea), "la línea de lo administrativo de contracts/paquetes.md §2")

	raiz := arbolDePaquetes(t)
	antes := fotografiar(t, raiz)

	regenerado := regenerar(t, raiz, nil)
	require.Empty(t, presentarDefectos(regenerado.Defectos()))
	assert.Equal(t, map[string]string{
		"consulta-juridica":   skillMdDeLaEntrada(listaDeLoAdministrativo),
		"normativa":           skillMdDeCapacidad("normativa", ""),
		"plazos-y-territorio": skillMdDeCapacidad("plazos-y-territorio", enlaceDeLoAdministrativo) + enlaceACitar,
	}, contenidosRegenerados(regenerado))

	derivas, err := skills.Comparar(raiz, regenerado)
	require.NoError(t, err)
	assert.Equal(t, []*skills.Deriva{
		{Skill: "consulta-juridica", Ruta: "SKILL.md", Clase: skills.DerivaContenidoDistinto},
		{Skill: "plazos-y-territorio", Ruta: "SKILL.md", Clase: skills.DerivaContenidoDistinto},
	}, derivas)
	assert.Equal(t, antes, fotografiar(t, raiz), "Regenerar y Comparar no escriben nada")

	require.NoError(t, skills.Escribir(raiz, regenerado))
	assert.Empty(t, compararSinDefectos(t, raiz, nil))
	assert.Equal(t, skillMdDeCapacidad("plazos-y-territorio", enlaceDeLoAdministrativo)+enlaceACitar,
		leerFicheroDePrueba(t, rutaDeSkill(raiz, "plazos-y-territorio", "SKILL.md")))
	assert.Equal(t, referenciaDeUnPaquete,
		leerFicheroDePrueba(t, rutaDeSkill(raiz, "plazos-y-territorio", "references", "administrativo.md")),
		"la referencia del paquete es a mano: Escribir no la toca")
}

// enlaceACitar es lo que enlazarDesdeSkillMd añade al final del SKILL.md de
// plazos-y-territorio en arbolDePaquetes: el enlace a su referencia a mano que no
// es de ningún paquete.
const enlaceACitar = "- [como_se_cita](references/como_se_cita.md)\n"

// probarDosPaquetesEnOrden fija los enlaces de dos paquetes en una skill, en el
// orden de sus nombres y no en el de data/normas.yaml, y la lista de la entrada
// con los dos.
func probarDosPaquetesEnOrden(t *testing.T) {
	t.Parallel()

	raiz := arbolDePaquetes(t)
	escribirReferenciaDeUnPaquete(t, raiz, "plazos-y-territorio", "fiscal")

	regenerado := regenerar(t, raiz, nil)
	require.Empty(t, presentarDefectos(regenerado.Defectos()))

	contenidos := contenidosRegenerados(regenerado)
	assert.Equal(t, skillMdDeCapacidad("plazos-y-territorio", enlaceDeLoAdministrativo+enlaceDeLoFiscal)+enlaceACitar,
		contenidos["plazos-y-territorio"])
	assert.Equal(t, skillMdDeLaEntrada(listaDeLoAdministrativo+"- lo fiscal: `plazos-y-territorio`.\n"),
		contenidos["consulta-juridica"])
}

// probarLaListaDeLaEntrada fija la lista de la entrada: la línea de lo
// administrativo de contracts/paquetes.md §4 con una skill, y con dos, que van en
// el orden de sus nombres, separadas por «, »; cada skill enlaza solo las
// referencias que tiene.
func probarLaListaDeLaEntrada(t *testing.T) {
	t.Parallel()

	raiz := arbolDePaquetes(t)

	regenerado := regenerar(t, raiz, nil)
	require.Empty(t, presentarDefectos(regenerado.Defectos()))
	assert.Equal(t, skillMdDeLaEntrada("- lo administrativo: `plazos-y-territorio`.\n"),
		contenidosRegenerados(regenerado)["consulta-juridica"])

	escribirReferenciaDeUnPaquete(t, raiz, "normativa", "administrativo")

	regenerado = regenerar(t, raiz, nil)
	require.Empty(t, presentarDefectos(regenerado.Defectos()))

	contenidos := contenidosRegenerados(regenerado)
	assert.Equal(t, skillMdDeLaEntrada("- lo administrativo: `normativa`, `plazos-y-territorio`.\n"),
		contenidos["consulta-juridica"])
	assert.Equal(t, skillMdDeCapacidad("normativa", enlaceDeLoAdministrativo), contenidos["normativa"])
}

// probarElTextoEditadoAMano fija que el texto entre las marcas que no es el
// generado es una deriva de SKILL.md, sobre un árbol sincronizado: el enlace de
// plazos-y-territorio cambiado, la lista de la entrada sin su línea y una línea
// entre las marcas de normativa, que no tiene la referencia de ningún paquete.
// Escribir deshace las tres.
func probarElTextoEditadoAMano(t *testing.T) {
	t.Parallel()

	raiz := arbolDePaquetes(t)
	require.NoError(t, skills.Escribir(raiz, regenerar(t, raiz, nil)))
	require.Empty(t, compararSinDefectos(t, raiz, nil))

	cambiarFicheroDePrueba(t, rutaDeSkill(raiz, "plazos-y-territorio", "SKILL.md"), ", lee [", ", consulta [")
	cambiarFicheroDePrueba(t, rutaDeSkill(raiz, "consulta-juridica", "SKILL.md"), listaDeLoAdministrativo, "")
	cambiarFicheroDePrueba(t, rutaDeSkill(raiz, "normativa", "SKILL.md"), inicioDeLosPaquetes,
		inicioDeLosPaquetes+enlaceDeLoFiscal)

	assert.Equal(t, []*skills.Deriva{
		{Skill: "consulta-juridica", Ruta: "SKILL.md", Clase: skills.DerivaContenidoDistinto},
		{Skill: "normativa", Ruta: "SKILL.md", Clase: skills.DerivaContenidoDistinto},
		{Skill: "plazos-y-territorio", Ruta: "SKILL.md", Clase: skills.DerivaContenidoDistinto},
	}, compararSinDefectos(t, raiz, nil))

	require.NoError(t, skills.Escribir(raiz, regenerar(t, raiz, nil)))
	assert.Empty(t, compararSinDefectos(t, raiz, nil))
	assert.Equal(t, skillMdDeLaEntrada(listaDeLoAdministrativo),
		leerFicheroDePrueba(t, rutaDeSkill(raiz, "consulta-juridica", "SKILL.md")))
	assert.Equal(t, skillMdDeCapacidad("normativa", ""), leerFicheroDePrueba(t, rutaDeSkill(raiz, "normativa", "SKILL.md")))
	assert.Equal(t, skillMdDeCapacidad("plazos-y-territorio", enlaceDeLoAdministrativo)+enlaceACitar,
		leerFicheroDePrueba(t, rutaDeSkill(raiz, "plazos-y-territorio", "SKILL.md")))
}

// probarLaReferenciaSinMarcas fija el defecto de una skill con la referencia de
// un paquete y sin las marcas, uno por referencia en el orden de los nombres,
// aunque su cuerpo la enlace a mano, como haría un paquete sin sincronía; una
// referencia a mano que no es de ningún paquete no lo da. Con el defecto, Escribir
// no escribe nada.
func probarLaReferenciaSinMarcas(t *testing.T) {
	t.Parallel()

	raiz := arbolDePaquetes(t)
	escribirReferenciaDeUnPaquete(t, raiz, "plazos-y-territorio", "fiscal")
	cambiarFicheroDePrueba(t, rutaDeSkill(raiz, "plazos-y-territorio", "SKILL.md"),
		inicioDeLosPaquetes+finDeLosPaquetes, "")
	enlazarDesdeSkillMd(t, raiz, "plazos-y-territorio", "fiscal.md", "administrativo.md")

	escribirFicheroDePrueba(t, rutaDeSkill(raiz, "normativa", "SKILL.md"),
		strings.Replace(skillMdDeCapacidad("normativa", ""), inicioDeLosPaquetes+finDeLosPaquetes, "", 1))
	escribirFicheroDePrueba(t, rutaDeSkill(raiz, "normativa", "references", "como_se_cita.md"), referenciaAMano)
	enlazarDesdeSkillMd(t, raiz, "normativa", "como_se_cita.md")

	regenerado := regenerar(t, raiz, nil)
	assert.Equal(t, []string{
		"plazos-y-territorio: SKILL.md: referencia de paquete sin enlazar: administrativo.md",
		"plazos-y-territorio: SKILL.md: referencia de paquete sin enlazar: fiscal.md",
	}, presentarDefectos(regenerado.Defectos()))

	envejecer(t, raiz)
	antes := fotografiar(t, raiz)

	err := skills.Escribir(raiz, regenerado)
	require.ErrorContains(t, err, "referencia de paquete sin enlazar: administrativo.md")
	assert.Equal(t, antes, fotografiar(t, raiz), "con defectos, Escribir no escribe nada")
}

// probarLasNormasConDefectos fija que, con data/normas.yaml sin leer, cada skill
// con marcas tiene el defecto de la tabla, una sola vez aunque declare además una
// referencia generada desde ella, y la que no lleva marcas no lo tiene.
func probarLasNormasConDefectos(t *testing.T) {
	t.Parallel()

	raiz := arbolDePaquetes(t)
	cambiarFicheroDePrueba(t, filepath.Join(raiz, "data", "normas.yaml"),
		"      - procedimiento\n", "      - procedimiento\n"+campoNoDeclaradoDeNorma)
	cambiarFicheroDePrueba(t, rutaDeSkill(raiz, "normativa", "SKILL.md"), "---\n# Capacidad\n",
		"metadata:\n  kitlegal-referencias: normas\n---\n# Capacidad\n")
	escribirFicheroDePrueba(t, rutaDeSkill(raiz, "zeta", "SKILL.md"), skillMdDeBetaSinNadaQueGenerar)
	cambiarFicheroDePrueba(t, rutaDeSkill(raiz, "zeta", "SKILL.md"), "name: beta\n", "name: zeta\n")

	defecto := ": data/normas.yaml: BOE-A-2015-10565: campo no declarado: estado"

	assert.Equal(t, []string{
		"consulta-juridica" + defecto,
		"normativa" + defecto,
		"plazos-y-territorio" + defecto,
	}, presentarDefectos(regenerar(t, raiz, nil).Defectos()))
}

// probarSinTablaDeNormas fija que, sin data/normas.yaml, no hay paquetes: las
// marcas quedan vacías, sin defecto.
func probarSinTablaDeNormas(t *testing.T) {
	t.Parallel()

	raiz := t.TempDir()
	escribirFicheroDePrueba(t, rutaDeSkill(raiz, "normativa", "SKILL.md"),
		skillMdDeCapacidad("normativa", enlaceDeLoAdministrativo))

	regenerado := regenerar(t, raiz, nil)
	require.Empty(t, presentarDefectos(regenerado.Defectos()))
	assert.Equal(t, map[string]string{"normativa": skillMdDeCapacidad("normativa", "")},
		contenidosRegenerados(regenerado))
}

// probarLasMarcasSinPareja fija el defecto de las marcas que no van en pareja, de
// los paquetes y de la lista de la entrada, con lo que falla: sin una de las dos,
// una en más de una línea o la de fin antes que la de inicio. Con él, Escribir no
// escribe nada.
func probarLasMarcasSinPareja(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre       string
		skill        string
		viejo, nuevo string
		defecto      string
	}{
		{
			nombre:  "sin-la-de-fin",
			skill:   "normativa",
			viejo:   finDeLosPaquetes,
			defecto: "normativa: SKILL.md: las marcas de los paquetes no van en pareja: falta la de fin",
		},
		{
			nombre:  "sin-la-de-inicio",
			skill:   "normativa",
			viejo:   inicioDeLosPaquetes,
			defecto: "normativa: SKILL.md: las marcas de los paquetes no van en pareja: falta la de inicio",
		},
		{
			nombre: "inicio-repetido",
			skill:  "normativa",
			viejo:  inicioDeLosPaquetes,
			nuevo:  inicioDeLosPaquetes + inicioDeLosPaquetes,
			defecto: "normativa: SKILL.md: las marcas de los paquetes no van en pareja: la de inicio aparece 2 veces, " +
				"en las líneas 7 y 8",
		},
		{
			nombre: "fin-repetido",
			skill:  "normativa",
			viejo:  finDeLosPaquetes,
			nuevo:  finDeLosPaquetes + "\n" + finDeLosPaquetes,
			defecto: "normativa: SKILL.md: las marcas de los paquetes no van en pareja: la de fin aparece 2 veces, " +
				"en las líneas 8 y 10",
		},
		{
			nombre: "fin-antes-que-inicio",
			skill:  "normativa",
			viejo:  inicioDeLosPaquetes + finDeLosPaquetes,
			nuevo:  finDeLosPaquetes + inicioDeLosPaquetes,
			defecto: "normativa: SKILL.md: las marcas de los paquetes no van en pareja: la de fin, en la línea 7, " +
				"va antes que la de inicio, en la línea 8",
		},
		{
			nombre:  "de-la-entrada-sin-la-de-fin",
			skill:   "consulta-juridica",
			viejo:   finDeLosPaquetesDeLaBase,
			defecto: "consulta-juridica: SKILL.md: las marcas de los paquetes de la base no van en pareja: falta la de fin",
		},
		{
			nombre: "de-la-entrada-fin-antes-que-inicio",
			skill:  "consulta-juridica",
			viejo:  inicioDeLosPaquetesDeLaBase + finDeLosPaquetesDeLaBase,
			nuevo:  finDeLosPaquetesDeLaBase + inicioDeLosPaquetesDeLaBase,
			defecto: "consulta-juridica: SKILL.md: las marcas de los paquetes de la base no van en pareja: la de fin, " +
				"en la línea 9, va antes que la de inicio, en la línea 10",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			raiz := arbolDePaquetes(t)
			cambiarFicheroDePrueba(t, rutaDeSkill(raiz, caso.skill, "SKILL.md"), caso.viejo, caso.nuevo)

			regenerado := regenerar(t, raiz, nil)
			assert.Equal(t, []string{caso.defecto}, presentarDefectos(regenerado.Defectos()))

			envejecer(t, raiz)
			antes := fotografiar(t, raiz)

			require.ErrorContains(t, skills.Escribir(raiz, regenerado), caso.defecto)
			assert.Equal(t, antes, fotografiar(t, raiz), "con defectos, Escribir no escribe nada")
		})
	}
}

// probarSinPoderLeerLosPaquetes fija los errores con los que Regenerar no
// regenera nada porque no puede saber los paquetes: data/normas.yaml, que no se
// puede leer, y el references/ de una skill, que no se puede listar. Cada error
// nombra lo que no se puede leer.
func probarSinPoderLeerLosPaquetes(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string
		ruta   []string

		// directorio dice si en la ruta va un directorio vacío; si no, un fichero.
		directorio bool

		prefijo string
		causa   error
	}{
		{
			nombre:     "tabla-de-normas-que-es-un-directorio",
			ruta:       []string{"data", "normas.yaml"},
			directorio: true,
			prefijo:    "data/normas.yaml no se puede leer: ",
			causa:      syscall.EISDIR,
		},
		{
			nombre:  "references-que-es-un-fichero",
			ruta:    []string{"skills", "normativa", "references"},
			prefijo: "skills/normativa/references no se puede listar: ",
			causa:   syscall.ENOTDIR,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			raiz := arbolDePaquetes(t)
			ruta := filepath.Join(append([]string{raiz}, caso.ruta...)...)

			if caso.directorio {
				require.NoError(t, os.Remove(ruta))
				require.NoError(t, os.MkdirAll(ruta, 0o750))
			} else {
				escribirFicheroDePrueba(t, ruta, "no es un directorio\n")
			}

			regenerado, err := skills.Regenerar(raiz, nil)
			require.ErrorIs(t, err, caso.causa)
			assert.True(t, strings.HasPrefix(err.Error(), caso.prefijo), "error: %v", err)
			assert.Zero(t, regenerado)
		})
	}
}

// contenidosRegenerados son los SKILL.md regenerados de cada skill, por nombre.
func contenidosRegenerados(regenerado skills.Regenerado) map[string]string {
	contenidos := make(map[string]string, len(regenerado.Skills))

	for _, skill := range regenerado.Skills {
		contenidos[skill.Nombre] = string(skill.Contenido)
	}

	return contenidos
}
