package skills_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/skills"
)

// Las marcas de la región de cada una de las tres tablas de la referencia de lo administrativo, cada una
// con su salto de línea (contracts/paquetes.md §3 de H32).
const (
	inicioDeLaCobertura = "<!-- inicio de cobertura_por_materia: generada desde data/cobertura.yaml y " +
		"data/normas.yaml con make skills-sync, no editar -->\n"
	finDeLaCobertura   = "<!-- fin de cobertura_por_materia -->\n"
	inicioDeLosPlazos  = "<!-- inicio de recursos_y_plazos: generada desde data/plazos.yaml y data/normas.yaml con make skills-sync, no editar -->\n"
	finDeLosPlazos     = "<!-- fin de recursos_y_plazos -->\n"
	inicioDeLasLeyes   = "<!-- inicio de leyes_vertebrales: generada desde data/normas.yaml con make skills-sync, no editar -->\n"
	finDeLasLeyes      = "<!-- fin de leyes_vertebrales -->\n"
	prosaDeLasRegiones = "# Lo administrativo\n\nLa prosa del paquete, escrita a mano.\n\n"
)

// Las dos formas ⚠ NO CUBIERTO: del cuerpo de plazos-y-territorio, con sus
// marcadores (contracts/forma-de-cita.md §3 de H32).
const (
	formaConNorma = "⚠ NO CUBIERTO: <procedimiento> se rige por <norma> (<identificador de la norma>), no por el " +
		"procedimiento común: kitlegal no calcula sus plazos ni sus recursos (<precepto> [<identificador>, bloque <id>])."
	formaSinNorma = "⚠ NO CUBIERTO: <procedimiento> lo regula aparte la ley, no el procedimiento común: kitlegal no " +
		"calcula sus plazos ni sus recursos (<precepto> [<identificador>, bloque <id>])."
)

// normasDeLasRegiones es la tabla de normas de los árboles de TestRegiones: tres
// de lo administrativo, la del contencioso sin abreviatura, y una de otra materia,
// sin vertical; dos vertebrales.
const normasDeLasRegiones = "normas:\n" +
	"  BOE-A-2015-10565:\n" +
	"    titulo: \"Ley 39/2015, de 1 de octubre, del Procedimiento Administrativo Común.\"\n" +
	"    rango: Ley\n" +
	"    abreviatura: LPAC\n" +
	"    materias:\n" +
	"      - procedimiento administrativo\n" +
	"    vertebral: true\n" +
	"    vertical: administrativo\n" +
	"  BOE-A-1998-16718:\n" +
	"    titulo: \"Ley 29/1998, de 13 de julio, reguladora de la Jurisdicción Contencioso-administrativa.\"\n" +
	"    rango: Ley\n" +
	"    materias:\n" +
	"      - jurisdicción contencioso-administrativa\n" +
	"    vertical: administrativo\n" +
	"  BOE-A-2013-12887:\n" +
	"    titulo: \"Ley 19/2013, de 9 de diciembre, de transparencia.\"\n" +
	"    rango: Ley\n" +
	"    abreviatura: LTAIBG\n" +
	"    materias:\n" +
	"      - transparencia\n" +
	"    vertical: administrativo\n" +
	"  BOE-A-2003-23186:\n" +
	"    titulo: \"Ley 58/2003, de 17 de diciembre, General Tributaria.\"\n" +
	"    rango: Ley\n" +
	"    abreviatura: LGT\n" +
	"    materias:\n" +
	"      - tributos\n" +
	"    vertebral: true\n"

// coberturaDeLasRegiones es la cobertura de los árboles de TestRegiones: una fila
// con la norma que la rige, con el apartado de una disposición, y otra sin ella,
// con el de un artículo y una barra en su procedimiento.
const coberturaDeLasRegiones = "regulados_aparte:\n" +
	"  - procedimiento: aplicación de los tributos\n" +
	"    precepto: {norma: BOE-A-2015-10565, bloque: daprimera, apartado: \"DA 1ª.2.a\"}\n" +
	"    norma: {identificador: BOE-A-2003-23186, nombre: \"Ley 58/2003, General Tributaria\", " +
	"aplicable: {bloque: a97, apartado: \"97\"}}\n" +
	"  - procedimiento: \"reclamaciones con una barra | en su nombre\"\n" +
	"    precepto: {norma: BOE-A-2015-10565, bloque: a112, apartado: \"112.4\"}\n"

// Lo que la sincronía escribe en la región de la cobertura y en la de las leyes
// vertebrales de los árboles de TestRegiones, byte a byte: la línea de cada fila
// es la forma del cuerpo que le toca con sus marcadores sustituidos, con la
// barra escrita \|; y las vertebrales, por año y número.
const (
	tablaDeLaCobertura = "| Procedimiento | Línea que da la respuesta |\n" +
		"|---|---|\n" +
		"| aplicación de los tributos | ⚠ NO CUBIERTO: aplicación de los tributos se rige por Ley 58/2003, General " +
		"Tributaria (BOE-A-2003-23186), no por el procedimiento común: kitlegal no calcula sus plazos ni sus recursos " +
		"(DA 1ª.2.a LPAC [BOE-A-2015-10565, bloque daprimera]). |\n" +
		"| reclamaciones con una barra \\| en su nombre | ⚠ NO CUBIERTO: reclamaciones con una barra \\| en su nombre lo " +
		"regula aparte la ley, no el procedimiento común: kitlegal no calcula sus plazos ni sus recursos " +
		"(art. 112.4 LPAC [BOE-A-2015-10565, bloque a112]). |\n"
	tablaDeLasLeyes = "| Norma | Abreviatura | Identificador | Rango | Materias |\n" +
		"|---|---|---|---|---|\n" +
		"| Ley 58/2003, de 17 de diciembre, General Tributaria. | LGT | `BOE-A-2003-23186` | Ley | tributos |\n" +
		"| Ley 39/2015, de 1 de octubre, del Procedimiento Administrativo Común. | LPAC | `BOE-A-2015-10565` | Ley | " +
		"procedimiento administrativo |\n"
)

// Las filas de la región de los recursos y plazos que fijan la forma legible de
// un precepto: con la abreviatura de su norma y, sin ella, con su título hasta la
// primera coma (contracts/paquetes.md §3; research D7 de H32).
const (
	filaDeLaReposicion = "| recurso potestativo de reposición | `reposicion` | 1 mes | `lpac` | " +
		"art. 124.1 LPAC [BOE-A-2015-10565, bloque a124] |\n"
	filaDelContencioso = "| recurso contencioso-administrativo | `contencioso` | 2 meses | `contencioso` | " +
		"art. 46.1 Ley 29/1998 [BOE-A-1998-16718, bloque a46]; art. 128.2 Ley 29/1998 [BOE-A-1998-16718, bloque a128] |\n"
)

// referenciaConRegiones es la referencia de lo administrativo de los árboles de
// TestRegiones, con la prosa y la región de cada tabla con lo que se da entre sus marcas.
func referenciaConRegiones(cobertura, plazos, leyes string) string {
	return prosaDeLasRegiones +
		inicioDeLaCobertura + cobertura + finDeLaCobertura +
		"\n## Recursos y plazos\n\n" +
		inicioDeLosPlazos + plazos + finDeLosPlazos +
		"\n## Las leyes vertebrales\n\n" +
		inicioDeLasLeyes + leyes + finDeLasLeyes
}

// skillMdConLasFormas es el SKILL.md de plazos-y-territorio de los árboles de
// TestRegiones: el enlace a la referencia de lo administrativo entre las marcas de
// los paquetes, ya sincronizado, y las dos formas en un bloque text.
const skillMdConLasFormas = "---\n" +
	"name: plazos-y-territorio\n" +
	"description: Skill de prueba que da la línea de lo no cubierto.\n" +
	"---\n" +
	"# Plazos\n" +
	"\n" +
	inicioDeLosPaquetes + enlaceDeLoAdministrativo + finDeLosPaquetes +
	"\n" +
	"```text\n" +
	formaConNorma + "\n" +
	formaSinNorma + "\n" +
	"```\n"

// arbolDeRegiones es un repositorio temporal con la skill plazos-y-territorio, su
// referencia de lo administrativo con la región de cada tabla vacía y los datos de las
// tres: normasDeLasRegiones, coberturaDeLasRegiones y data/plazos.yaml del
// repositorio.
func arbolDeRegiones(t *testing.T) string {
	t.Helper()

	raiz := t.TempDir()
	escribirFicheroDePrueba(t, filepath.Join(raiz, "data", "normas.yaml"), normasDeLasRegiones)
	escribirFicheroDePrueba(t, filepath.Join(raiz, "data", "cobertura.yaml"), coberturaDeLasRegiones)
	escribirFicheroDePrueba(t, filepath.Join(raiz, "data", "plazos.yaml"), string(leerDatosDelRepositorio(t, "plazos.yaml")))
	escribirFicheroDePrueba(t, rutaDeSkill(raiz, "plazos-y-territorio", "SKILL.md"), skillMdConLasFormas)
	escribirFicheroDePrueba(t, rutaDeLaReferenciaConRegiones(raiz), referenciaConRegiones("", "", ""))

	return raiz
}

// rutaDeLaReferenciaConRegiones es la de la referencia de lo administrativo en el
// árbol de raíz.
func rutaDeLaReferenciaConRegiones(raiz string) string {
	return rutaDeSkill(raiz, "plazos-y-territorio", "references", "administrativo.md")
}

// derivaDeLaReferenciaConRegiones es la deriva de la referencia de lo
// administrativo cuando lo que hay entre sus marcas no es lo generado.
var derivaDeLaReferenciaConRegiones = []*skills.Deriva{
	{Skill: "plazos-y-territorio", Ruta: "references/administrativo.md", Clase: skills.DerivaContenidoDistinto},
}

// TestRegiones fija cada región generada de una referencia a mano
// (contracts/paquetes.md §3; data-model §4 y §5; research D4, D6 y D7; FR-012 y
// FR-019 de H32), sobre repositorios temporales: la sincronía escribe entre las
// marcas de cada región la tabla de su generador desde data/ —la cobertura, con
// la línea de cada fila del paquete que nombra el fichero; los recursos y
// plazos, con cada precepto en la forma única; las leyes vertebrales— y no toca
// nada fuera de ellas; lo que hay entre las marcas y no es lo generado, y la forma
// del cuerpo cambiada sin regenerar, son la deriva contenido-distinto de la
// referencia; y cada defecto nuevo del molde, con su fichero.
func TestRegiones(t *testing.T) {
	t.Parallel()

	t.Run("sincronia", probarLaSincroniaDeLasRegiones)
	t.Run("editada-a-mano", probarLaRegionEditadaAMano)
	t.Run("forma-cambiada-sin-regenerar", probarLaFormaCambiadaSinRegenerar)
	t.Run("defectos", probarLosDefectosDeLasRegiones)
	t.Run("datos-que-no-se-leen", probarLosDatosQueNoSeLeen)
	t.Run("datos-de-las-marcas", probarLosDatosDeLasMarcas)
}

// probarLaSincroniaDeLasRegiones fija lo que la sincronía escribe en cada
// región vacía: Comparar da la deriva de la referencia, Escribir la deja con
// sus tablas y la prosa intacta, y después no hay deriva.
func probarLaSincroniaDeLasRegiones(t *testing.T) {
	t.Parallel()

	raiz := arbolDeRegiones(t)

	assert.Equal(t, derivaDeLaReferenciaConRegiones, compararSinDefectos(t, raiz, nil))

	require.NoError(t, skills.Escribir(raiz, regenerar(t, raiz, nil)))
	assert.Empty(t, compararSinDefectos(t, raiz, nil))

	referencia := leerFicheroDePrueba(t, rutaDeLaReferenciaConRegiones(raiz))
	assert.True(t, strings.HasPrefix(referencia, prosaDeLasRegiones+inicioDeLaCobertura+tablaDeLaCobertura+finDeLaCobertura),
		"la prosa, intacta, y la cobertura, byte a byte: %s", referencia)
	assert.True(t, strings.HasSuffix(referencia, inicioDeLasLeyes+tablaDeLasLeyes+finDeLasLeyes),
		"las leyes vertebrales, byte a byte: %s", referencia)

	_, plazos, conPlazos := strings.Cut(referencia, inicioDeLosPlazos)
	plazos, _, conFin := strings.Cut(plazos, finDeLosPlazos)

	require.True(t, conPlazos && conFin, "la región de los recursos y plazos: %s", referencia)
	assert.True(t, strings.HasPrefix(plazos, "| Tipo | Código | Duración | Cómputo | Precepto |\n|---|---|---|---|---|\n"),
		"los encabezados de los recursos y plazos: %s", plazos)
	assert.Contains(t, plazos, filaDeLaReposicion)
	assert.Contains(t, plazos, filaDelContencioso)
}

// probarLaRegionEditadaAMano fija que lo editado a mano entre las marcas de una
// región es la deriva de la referencia, y que Escribir lo deshace.
func probarLaRegionEditadaAMano(t *testing.T) {
	t.Parallel()

	raiz := arbolDeRegiones(t)
	sincronizada := referenciaConRegiones(tablaDeLaCobertura, "", tablaDeLasLeyes)
	escribirFicheroDePrueba(t, rutaDeLaReferenciaConRegiones(raiz), sincronizada)
	require.NoError(t, skills.Escribir(raiz, regenerar(t, raiz, nil)))

	antes := leerFicheroDePrueba(t, rutaDeLaReferenciaConRegiones(raiz))
	cambiarFicheroDePrueba(t, rutaDeLaReferenciaConRegiones(raiz), "lo regula aparte la ley", "lo regula aparte otra ley")

	assert.Equal(t, derivaDeLaReferenciaConRegiones, compararSinDefectos(t, raiz, nil))

	require.NoError(t, skills.Escribir(raiz, regenerar(t, raiz, nil)))
	assert.Equal(t, antes, leerFicheroDePrueba(t, rutaDeLaReferenciaConRegiones(raiz)))
}

// probarLaFormaCambiadaSinRegenerar es el mutante de FR-012: la forma del cuerpo
// cambiada sin regenerar da la deriva de la referencia, y la línea regenerada es
// la forma nueva con sus marcadores sustituidos.
func probarLaFormaCambiadaSinRegenerar(t *testing.T) {
	t.Parallel()

	raiz := arbolDeRegiones(t)
	require.NoError(t, skills.Escribir(raiz, regenerar(t, raiz, nil)))
	require.Empty(t, compararSinDefectos(t, raiz, nil))

	cambiarFicheroDePrueba(t, rutaDeSkill(raiz, "plazos-y-territorio", "SKILL.md"), "lo regula aparte la ley,",
		"lo regula aparte una ley,")

	assert.Equal(t, derivaDeLaReferenciaConRegiones, compararSinDefectos(t, raiz, nil))

	require.NoError(t, skills.Escribir(raiz, regenerar(t, raiz, nil)))
	assert.Contains(t, leerFicheroDePrueba(t, rutaDeLaReferenciaConRegiones(raiz)),
		"| ⚠ NO CUBIERTO: reclamaciones con una barra \\| en su nombre lo regula aparte una ley, no el procedimiento común")
}

// casoDeRegion es un caso de los defectos de una región: el cambio sobre el
// árbol de arbolDeRegiones y los defectos exactos que da.
type casoDeRegion struct {
	nombre   string
	cambio   func(t *testing.T, raiz string)
	defectos []string
}

// probarLosDefectosDeLasRegiones fija cada defecto nuevo del molde, con su
// fichero, y que con ellos Escribir no escribe nada.
func probarLosDefectosDeLasRegiones(t *testing.T) {
	t.Parallel()

	for _, caso := range casosDeRegiones() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			raiz := arbolDeRegiones(t)
			caso.cambio(t, raiz)
			antes := fotografiar(t, raiz)

			regenerado := regenerar(t, raiz, nil)
			assert.Equal(t, caso.defectos, presentarDefectos(regenerado.Defectos()))
			require.Error(t, skills.Escribir(raiz, regenerado))
			assert.Equal(t, antes, fotografiar(t, raiz), "con defectos no se escribe nada")
		})
	}
}

// casosDeRegiones son los casos de probarLosDefectosDeLasRegiones.
func casosDeRegiones() []casoDeRegion {
	referencia := rutaDeLaReferenciaConRegiones
	skillMd := func(raiz string) string { return rutaDeSkill(raiz, "plazos-y-territorio", "SKILL.md") }
	filaSinRegion := func(procedimiento string) string {
		return "plazos-y-territorio: data/cobertura.yaml: la fila de «" + procedimiento +
			"» no la genera ninguna región cobertura_por_materia"
	}

	return []casoDeRegion{
		{
			nombre: "generador-que-no-la-admite",
			cambio: func(t *testing.T, raiz string) {
				t.Helper()
				cambiarFicheroDePrueba(t, referencia(raiz), "\n## Las leyes vertebrales\n",
					"\n<!-- inicio de normas: generada desde data/normas.yaml con make skills-sync, no editar -->\n"+
						"<!-- fin de normas -->\n\n## Las leyes vertebrales\n")
			},
			defectos: []string{
				"plazos-y-territorio: references/administrativo.md: región de normas, un generador que no la admite",
			},
		},
		{
			nombre: "marcas-sin-pareja",
			cambio: func(t *testing.T, raiz string) {
				t.Helper()
				cambiarFicheroDePrueba(t, referencia(raiz), finDeLosPlazos, "")
			},
			defectos: []string{
				"plazos-y-territorio: references/administrativo.md: las marcas de la región recursos_y_plazos no van " +
					"en pareja: falta la de fin",
			},
		},
		{
			nombre: "marca-de-inicio-con-otros-datos",
			cambio: func(t *testing.T, raiz string) {
				t.Helper()
				cambiarFicheroDePrueba(t, referencia(raiz), "desde data/plazos.yaml y data/normas.yaml",
					"desde data/plazos.yaml")
			},
			defectos: []string{
				"plazos-y-territorio: references/administrativo.md: la marca de inicio de la región recursos_y_plazos " +
					"no es «" + strings.TrimSuffix(inicioDeLosPlazos, "\n") + "»",
			},
		},
		{
			nombre: "skill-sin-las-dos-formas",
			cambio: func(t *testing.T, raiz string) {
				t.Helper()
				cambiarFicheroDePrueba(t, skillMd(raiz), formaSinNorma+"\n", "")
			},
			defectos: []string{
				"plazos-y-territorio: references/administrativo.md: región cobertura_por_materia en una skill sin las " +
					"dos formas ⚠ NO CUBIERTO: en SKILL.md",
			},
		},
		{
			nombre: "marcador-que-la-tabla-no-conoce",
			cambio: func(t *testing.T, raiz string) {
				t.Helper()
				cambiarFicheroDePrueba(t, skillMd(raiz), "(<identificador de la norma>)", "(<identificador de la ley>)")
			},
			defectos: []string{
				"plazos-y-territorio: references/administrativo.md: la forma ⚠ NO CUBIERTO: de SKILL.md lleva " +
					"<identificador de la ley>, un marcador que la tabla no conoce",
			},
		},
		{
			nombre: "fila-sin-region",
			cambio: func(t *testing.T, raiz string) {
				t.Helper()
				cambiarFicheroDePrueba(t, referencia(raiz), inicioDeLaCobertura+finDeLaCobertura, "")
			},
			defectos: []string{
				filaSinRegion("aplicación de los tributos"),
				filaSinRegion("reclamaciones con una barra | en su nombre"),
			},
		},
		{
			nombre: "fila-de-una-norma-sin-paquete",
			cambio: func(t *testing.T, raiz string) {
				t.Helper()
				escribirFicheroDePrueba(t, filepath.Join(raiz, "data", "cobertura.yaml"), coberturaDeLasRegiones+
					"  - procedimiento: liquidaciones\n"+
					"    precepto: {norma: BOE-A-2003-23186, bloque: a97, apartado: \"97\"}\n")
			},
			defectos: []string{filaSinRegion("liquidaciones")},
		},
		{
			nombre: "region-sin-sus-datos",
			cambio: func(t *testing.T, raiz string) {
				t.Helper()
				retirarDePrueba(t, filepath.Join(raiz, "data", "plazos.yaml"))
			},
			defectos: []string{
				"plazos-y-territorio: references/administrativo.md: la región recursos_y_plazos sin data/plazos.yaml",
			},
		},
	}
}

// probarLosDatosQueNoSeLeen fija que los defectos de un YAML de datos de una
// región que no se lee son de la skill, con su fichero delante, y que con ellos
// Escribir no escribe nada.
func probarLosDatosQueNoSeLeen(t *testing.T) {
	t.Parallel()

	raiz := arbolDeRegiones(t)
	escribirFicheroDePrueba(t, filepath.Join(raiz, "data", "plazos.yaml"), "tipos: []\n")

	regenerado := regenerar(t, raiz, nil)
	defectos := presentarDefectos(regenerado.Defectos())

	require.NotEmpty(t, defectos)

	for _, defecto := range defectos {
		assert.True(t, strings.HasPrefix(defecto, "plazos-y-territorio: data/plazos.yaml: "),
			"el defecto nombra el fichero: %q", defecto)
	}

	require.Error(t, skills.Escribir(raiz, regenerado))
}

// probarLosDatosDeLasMarcas fija DatosDeLasRegiones: los ficheros de data/ que
// nombran las marcas de inicio de cada región de una referencia, cada uno una vez
// y en el orden en que aparecen; ninguno sin ninguna región.
func probarLosDatosDeLasMarcas(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []string{"data/cobertura.yaml", "data/normas.yaml", "data/plazos.yaml"},
		skills.DatosDeLasRegiones([]byte(referenciaConRegiones("", "", ""))))
	assert.Empty(t, skills.DatosDeLasRegiones([]byte(prosaDeLasRegiones)))
}
