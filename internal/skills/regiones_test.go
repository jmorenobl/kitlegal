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
	finDeLaCobertura  = "<!-- fin de cobertura_por_materia -->\n"
	inicioDeLosPlazos = "<!-- inicio de recursos_y_plazos: generada desde data/plazos.yaml y data/normas.yaml con make skills-sync, no editar -->\n"
	finDeLosPlazos    = "<!-- fin de recursos_y_plazos -->\n"
	inicioDeLasLeyes  = "<!-- inicio de leyes_vertebrales: generada desde data/normas.yaml con make skills-sync, no editar -->\n"
	finDeLasLeyes     = "<!-- fin de leyes_vertebrales -->\n"
)

// La plantilla ⚠ NO CUBIERTO: del cuerpo de plazos-y-territorio y los dos
// motivos de su referencia de lo administrativo, con sus marcadores (FR-004,
// FR-005; contracts/datos-y-regiones.md §5; data-model §2).
const (
	plantillaDeNoCubierto = "⚠ NO CUBIERTO: <asunto>: <motivo>. kitlegal no da sus plazos, órganos, vías ni escritos."
	motivoConNorma        = "se rige por <norma> (<identificador de la norma>), no por el procedimiento común " +
		"(<precepto> [<identificador>, bloque <id>])"
	motivoSinNorma = "lo regula aparte la ley, no el procedimiento común (<precepto> [<identificador>, bloque <id>])"
)

// prosaDeLasRegiones es lo escrito a mano de la referencia de lo administrativo
// de los árboles de TestRegiones delante de la región de cada tabla: la prosa del
// paquete y la frase y el bloque text con los dos motivos de la línea
// (contracts/skills.md §3).
const prosaDeLasRegiones = "# Lo administrativo\n\nLa prosa del paquete, escrita a mano.\n\n" +
	"Los motivos de la línea; el primero es el de la fila con norma:\n" +
	"\n" +
	"```text\n" +
	motivoConNorma + "\n" +
	motivoSinNorma + "\n" +
	"```\n" +
	"\n"

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
const coberturaDeLasRegiones = "motivo_de_lo_no_declarado: fuera de lo que kitlegal cubre\n" +
	"regulados_aparte:\n" +
	"  - procedimiento: aplicación de los tributos\n" +
	"    precepto: {norma: BOE-A-2015-10565, bloque: daprimera, apartado: \"DA 1ª.2.a\"}\n" +
	"    norma: {identificador: BOE-A-2003-23186, nombre: \"Ley 58/2003, General Tributaria\", " +
	"aplicable: {bloque: a97, apartado: \"97\"}}\n" +
	"  - procedimiento: \"reclamaciones con una barra | en su nombre\"\n" +
	"    precepto: {norma: BOE-A-2015-10565, bloque: a112, apartado: \"112.4\"}\n"

// Lo que la sincronía escribe en la región de la cobertura y en la de las leyes
// vertebrales de los árboles de TestRegiones, byte a byte: la línea de cada fila
// es la plantilla del cuerpo con el motivo que le toca y sus marcadores
// sustituidos, con la barra escrita \|; y las vertebrales, por año y número.
const (
	tablaDeLaCobertura = "| Procedimiento | Línea que da la respuesta |\n" +
		"|---|---|\n" +
		"| aplicación de los tributos | ⚠ NO CUBIERTO: aplicación de los tributos: se rige por Ley 58/2003, General " +
		"Tributaria (BOE-A-2003-23186), no por el procedimiento común (DA 1ª.2.a LPAC [BOE-A-2015-10565, bloque " +
		"daprimera]). kitlegal no da sus plazos, órganos, vías ni escritos. |\n" +
		"| reclamaciones con una barra \\| en su nombre | ⚠ NO CUBIERTO: reclamaciones con una barra \\| en su nombre: " +
		"lo regula aparte la ley, no el procedimiento común (art. 112.4 LPAC [BOE-A-2015-10565, bloque a112]). " +
		"kitlegal no da sus plazos, órganos, vías ni escritos. |\n"
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

// skillMdConLaPlantilla es el SKILL.md de plazos-y-territorio de los árboles de
// TestRegiones: el enlace a la referencia de lo administrativo entre las marcas
// de los paquetes, ya sincronizado; en la prosa, fuera de un bloque, el motivo de
// lo no declarado de coberturaDeLasRegiones entre comillas angulares; y la
// plantilla, sola en un bloque text.
const skillMdConLaPlantilla = "---\n" +
	"name: plazos-y-territorio\n" +
	"description: Skill de prueba que da la línea de lo no cubierto.\n" +
	"---\n" +
	"# Plazos\n" +
	"\n" +
	inicioDeLosPaquetes + enlaceDeLoAdministrativo + finDeLosPaquetes +
	"\n" +
	"Lo no declarado está «fuera de lo que kitlegal cubre».\n" +
	"\n" +
	"```text\n" +
	plantillaDeNoCubierto + "\n" +
	"```\n"

// arbolDeRegiones es un repositorio temporal con la skill plazos-y-territorio y su
// plantilla, su referencia de lo administrativo con los dos motivos y la región de
// cada tabla vacía, y los datos de las tres: normasDeLasRegiones,
// coberturaDeLasRegiones y data/plazos.yaml del repositorio.
func arbolDeRegiones(t *testing.T) string {
	t.Helper()

	raiz := t.TempDir()
	escribirFicheroDePrueba(t, filepath.Join(raiz, "data", "normas.yaml"), normasDeLasRegiones)
	escribirFicheroDePrueba(t, filepath.Join(raiz, "data", "cobertura.yaml"), coberturaDeLasRegiones)
	escribirFicheroDePrueba(t, filepath.Join(raiz, "data", "plazos.yaml"), string(leerDatosDelRepositorio(t, "plazos.yaml")))
	escribirFicheroDePrueba(t, rutaDeSkill(raiz, "plazos-y-territorio", "SKILL.md"), skillMdConLaPlantilla)
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
// FR-019 de H32; FR-006 de H33; contracts/datos-y-regiones.md §5 de H33), sobre
// repositorios temporales: la sincronía escribe entre las marcas de cada región
// la tabla de su generador desde data/ —la cobertura, con la línea de cada fila
// del paquete que nombra el fichero, la plantilla del cuerpo con el motivo de su
// caso; los recursos y plazos, con cada precepto en la forma única; las leyes
// vertebrales— y no toca nada fuera de ellas; lo que hay entre las marcas y no es
// lo generado, y la plantilla o un motivo cambiados sin regenerar, son la deriva
// contenido-distinto de la referencia; y cada defecto del molde, con su fichero.
func TestRegiones(t *testing.T) {
	t.Parallel()

	t.Run("sincronia", probarLaSincroniaDeLasRegiones)
	t.Run("editada-a-mano", probarLaRegionEditadaAMano)
	t.Run("plantilla-cambiada-sin-regenerar", probarLaPlantillaCambiadaSinRegenerar)
	t.Run("motivo-cambiado-sin-regenerar", probarElMotivoCambiadoSinRegenerar)
	t.Run("motivos-entre-otras-lineas", probarLosMotivosEntreOtrasLineas)
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

// probarLosMotivosEntreOtrasLineas fija que los motivos de la línea son solo las
// líneas del bloque text que llevan <precepto> y no empiezan por ⚠: un aviso del
// paquete con <precepto> y una nota sin él, en el mismo bloque, no cuentan como
// motivo, y la región es la misma que sin ellos.
func probarLosMotivosEntreOtrasLineas(t *testing.T) {
	t.Parallel()

	raiz := arbolDeRegiones(t)
	cambiarFicheroDePrueba(t, rutaDeLaReferenciaConRegiones(raiz), "```text\n"+motivoConNorma,
		"```text\n  ⚠ Un aviso del paquete que cita <precepto>.\nUna nota del paquete, sin marcadores.\n"+motivoConNorma)

	regenerado := regenerar(t, raiz, nil)
	assert.Empty(t, presentarDefectos(regenerado.Defectos()))
	require.NoError(t, skills.Escribir(raiz, regenerado))

	assert.Contains(t, leerFicheroDePrueba(t, rutaDeLaReferenciaConRegiones(raiz)),
		inicioDeLaCobertura+tablaDeLaCobertura+finDeLaCobertura, "la cobertura, con los dos motivos de siempre")
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
	cambiarFicheroDePrueba(t, rutaDeLaReferenciaConRegiones(raiz), "en su nombre: lo regula aparte la ley",
		"en su nombre: lo regula aparte otra ley")

	assert.Equal(t, derivaDeLaReferenciaConRegiones, compararSinDefectos(t, raiz, nil))

	require.NoError(t, skills.Escribir(raiz, regenerar(t, raiz, nil)))
	assert.Equal(t, antes, leerFicheroDePrueba(t, rutaDeLaReferenciaConRegiones(raiz)))
}

// probarLaPlantillaCambiadaSinRegenerar es el mutante de FR-012 de H32 con la
// plantilla de H33: la plantilla del cuerpo cambiada sin regenerar da la deriva
// de la referencia, y la línea regenerada es la plantilla nueva con el motivo de
// su caso y sus marcadores sustituidos.
func probarLaPlantillaCambiadaSinRegenerar(t *testing.T) {
	t.Parallel()

	raiz := arbolDeRegiones(t)
	require.NoError(t, skills.Escribir(raiz, regenerar(t, raiz, nil)))
	require.Empty(t, compararSinDefectos(t, raiz, nil))

	cambiarFicheroDePrueba(t, rutaDeSkill(raiz, "plazos-y-territorio", "SKILL.md"),
		"kitlegal no da sus plazos, órganos, vías ni escritos.", "kitlegal no da sus plazos ni sus vías.")

	assert.Equal(t, derivaDeLaReferenciaConRegiones, compararSinDefectos(t, raiz, nil))

	require.NoError(t, skills.Escribir(raiz, regenerar(t, raiz, nil)))
	assert.Contains(t, leerFicheroDePrueba(t, rutaDeLaReferenciaConRegiones(raiz)),
		"| ⚠ NO CUBIERTO: reclamaciones con una barra \\| en su nombre: lo regula aparte la ley, no el procedimiento "+
			"común (art. 112.4 LPAC [BOE-A-2015-10565, bloque a112]). kitlegal no da sus plazos ni sus vías. |\n")
}

// probarElMotivoCambiadoSinRegenerar fija que el motivo de la referencia, escrito
// a mano fuera de la región, cambiado sin regenerar da la deriva de la
// referencia, y que la línea regenerada lleva el motivo nuevo.
func probarElMotivoCambiadoSinRegenerar(t *testing.T) {
	t.Parallel()

	raiz := arbolDeRegiones(t)
	require.NoError(t, skills.Escribir(raiz, regenerar(t, raiz, nil)))
	require.Empty(t, compararSinDefectos(t, raiz, nil))

	cambiarFicheroDePrueba(t, rutaDeLaReferenciaConRegiones(raiz), "\nlo regula aparte la ley,",
		"\nlo regula aparte una ley,")

	assert.Equal(t, derivaDeLaReferenciaConRegiones, compararSinDefectos(t, raiz, nil))

	require.NoError(t, skills.Escribir(raiz, regenerar(t, raiz, nil)))
	assert.Contains(t, leerFicheroDePrueba(t, rutaDeLaReferenciaConRegiones(raiz)),
		"| ⚠ NO CUBIERTO: reclamaciones con una barra \\| en su nombre: lo regula aparte una ley, no el procedimiento "+
			"común (art. 112.4 LPAC [BOE-A-2015-10565, bloque a112]). kitlegal no da sus plazos, órganos, vías ni "+
			"escritos. |\n")
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
	sinLaPlantilla := "plazos-y-territorio: references/administrativo.md: región cobertura_por_materia en una skill " +
		"sin la plantilla ⚠ NO CUBIERTO: en SKILL.md"
	marcadorDeLaPlantilla := func(marcador string) string {
		return "plazos-y-territorio: references/administrativo.md: la plantilla ⚠ NO CUBIERTO: de SKILL.md lleva " +
			marcador + ", un marcador que la tabla no conoce"
	}
	marcadorDeUnMotivo := func(marcador string) string {
		return "plazos-y-territorio: references/administrativo.md: un motivo de la plantilla lleva " + marcador +
			", un marcador que la tabla no conoce"
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
			nombre: "skill-sin-la-plantilla",
			cambio: func(t *testing.T, raiz string) {
				t.Helper()
				cambiarFicheroDePrueba(t, skillMd(raiz), plantillaDeNoCubierto+"\n", "")
			},
			defectos: []string{sinLaPlantilla},
		},
		{
			nombre: "skill-con-dos-plantillas",
			cambio: func(t *testing.T, raiz string) {
				t.Helper()
				cambiarFicheroDePrueba(t, skillMd(raiz), plantillaDeNoCubierto+"\n",
					plantillaDeNoCubierto+"\n"+"  ⚠ NO CUBIERTO: <asunto>: <motivo>.\n")
			},
			defectos: []string{sinLaPlantilla},
		},
		{
			nombre: "referencia-sin-un-motivo",
			cambio: func(t *testing.T, raiz string) {
				t.Helper()
				cambiarFicheroDePrueba(t, referencia(raiz), motivoSinNorma+"\n", "")
			},
			defectos: []string{
				"plazos-y-territorio: references/administrativo.md: región cobertura_por_materia en una referencia " +
					"sin sus dos motivos",
			},
		},
		{
			nombre: "marcador-de-la-plantilla-que-la-tabla-no-conoce",
			cambio: func(t *testing.T, raiz string) {
				t.Helper()
				cambiarFicheroDePrueba(t, skillMd(raiz), "<asunto>:", "<asunto del caso>:")
			},
			defectos: []string{marcadorDeLaPlantilla("<asunto del caso>")},
		},
		{
			nombre: "marcador-de-un-motivo-que-la-tabla-no-conoce",
			cambio: func(t *testing.T, raiz string) {
				t.Helper()
				cambiarFicheroDePrueba(t, referencia(raiz), "(<identificador de la norma>)", "(<identificador de la ley>)")
			},
			defectos: []string{marcadorDeUnMotivo("<identificador de la ley>")},
		},
		{
			nombre: "marcadores-de-la-plantilla-y-de-los-motivos",
			cambio: func(t *testing.T, raiz string) {
				t.Helper()
				cambiarFicheroDePrueba(t, referencia(raiz), "(<identificador de la norma>)", "(<ley> <fecha>)")
				cambiarFicheroDePrueba(t, referencia(raiz), "\nlo regula aparte la ley,", "\nlo regula aparte <ley>,")
				cambiarFicheroDePrueba(t, skillMd(raiz), "kitlegal no da", "<fecha> <tema>: kitlegal no da <tema>")
			},
			defectos: []string{
				marcadorDeLaPlantilla("<fecha>"),
				marcadorDeLaPlantilla("<tema>"),
				marcadorDeUnMotivo("<ley>"),
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
