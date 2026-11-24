package evals

import (
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/skills"
)

// Las cuatro descripciones de la variante del repositorio, cada una la línea de
// su apartado de contracts/variante.md §2 sin el salto de línea final.
const (
	descripcionCortaDeConsultaJuridica = "Punto de entrada de toda pregunta legal en España, de cualquier materia: " +
		"dice qué capacidad de kitlegal la resuelve —normativa, jurisprudencia, plazos y territorio— y con qué skill, " +
		"y lo que no cubre lo dice sin dar de memoria plazos, órganos ni vías. Se usa al empezar toda pregunta " +
		"legal, antes que las demás skills de kitlegal, también si no nombra norma o pide qué hacer, a quién " +
		"reclamar o qué escrito presentar."
	descripcionCortaDeJurisprudencia = "Sentencias y jurisprudencia de los tribunales españoles, sin citar ninguna " +
		"de memoria. Se usa siempre que la pregunta nombre, pida o cite una sentencia (STS, ECLI, ROJ), pregunte si " +
		"existe o qué dice, pida jurisprudencia sobre una materia o traiga una resolución, aunque el agente crea " +
		"conocerla. kitlegal no consulta el CENDOJ: prepara la búsqueda y coteja el documento. Se activa antes de " +
		"cita_preparar o cita_cotejar."
	descripcionCortaDeNormativa = "Consulta y cita normativa consolidada del BOE de cualquier materia: lee con " +
		"kitlegal índice y artículos y cita identificador y bloque. Se usa siempre que la respuesta dependa de lo " +
		"que dice una norma —un artículo, ley o real decreto, el plazo o requisito que fija— o de lo que decía en " +
		"una fecha, aunque el agente crea saberlo. Se activa antes de boe_articulo u otra herramienta boe_…: qué " +
		"pedir y cómo citar lo dice la skill."
	descripcionCortaDePlazosYTerritorio = "Calcula con kitlegal hasta qué día se puede recurrir una resolución " +
		"administrativa, cuándo vence un plazo del procedimiento común (LPAC) o el del contencioso y cuándo hay " +
		"silencio, con los festivos que cuentan, e identifica el territorio de un municipio: INE, provincia, " +
		"comunidad, DIR3 y boletines. No cuenta plazos de otras materias. Se usa al preguntar cuándo vence un plazo " +
		"o un recurso, o qué boletines tiene un municipio."
)

// lasCuatroSkills son las de skills/ del repositorio, en orden de nombre, las de
// los árboles que monta TestDescripcionesCortas.
var lasCuatroSkills = []string{"consulta-juridica", "jurisprudencia", "normativa", "plazos-y-territorio"}

// TestDescripcionesCortas comprueba la variante producto-descripciones-cortas y
// su comprobación (FR-001, FR-002; SC-001; contracts/variante.md §2 a §4;
// research D3): la del repositorio no tiene ningún defecto y da, por cada skill
// de skills/, la línea de su apartado de §2; y en árboles montados en un
// temporal, cada fila de la tabla de §3 da su línea exacta, y el ejemplo de §3,
// sus cinco líneas en su orden. Falla con la primera línea que difiere. Un
// directorio que no se puede leer, también uno que no existe, es el error de la
// función, no un defecto (ADR 0023).
func TestDescripcionesCortas(t *testing.T) {
	t.Parallel()

	t.Run("la variante del repositorio", func(t *testing.T) {
		t.Parallel()

		textos, defectos, err := descripcionesCortas(raizDelRepositorio)
		require.NoError(t, err)

		compararLasLineasDeLosDefectos(t, nil, defectos)

		assert.Equal(t, skillsDelArbol(t, raizDelRepositorio), slices.Sorted(maps.Keys(textos)),
			"una descripción por cada skill de skills/")

		assert.Equal(t, map[string]string{
			"consulta-juridica":   descripcionCortaDeConsultaJuridica,
			"jurisprudencia":      descripcionCortaDeJurisprudencia,
			"normativa":           descripcionCortaDeNormativa,
			"plazos-y-territorio": descripcionCortaDePlazosYTerritorio,
		}, textos)
	})

	// lasDelContrato es la variante de contracts/variante.md §2, cada fichero con
	// su línea y su salto de línea.
	lasDelContrato := map[string]string{
		"consulta-juridica.txt":   descripcionCortaDeConsultaJuridica + "\n",
		"jurisprudencia.txt":      descripcionCortaDeJurisprudencia + "\n",
		"normativa.txt":           descripcionCortaDeNormativa + "\n",
		"plazos-y-territorio.txt": descripcionCortaDePlazosYTerritorio + "\n",
	}

	// defectosDelEjemplo son los del caso «una que empieza por <ejemplo>», que
	// pone etiquetaDelEjemplo delante de la de normativa: el de la etiqueta y, si
	// la suma de lasDelContrato, sin sus saltos de línea, más lo que añade,
	// conLaEtiqueta, pasa de maximoDeLasDescripcionesCortas, el de la suma
	// (FR-023 de H45; research D8 de H45).
	const etiquetaDelEjemplo = "<ejemplo> "

	conLaEtiqueta := utf8.RuneCountInString(etiquetaDelEjemplo)
	for _, contenido := range lasDelContrato {
		conLaEtiqueta += utf8.RuneCountInString(strings.TrimSuffix(contenido, "\n"))
	}

	defectosDelEjemplo := []string{
		"laboratorio/producto-descripciones-cortas: la descripción de normativa tiene la etiqueta XML " +
			"«<ejemplo>» (no admite ninguna)",
	}

	if conLaEtiqueta > maximoDeLasDescripcionesCortas {
		defectosDelEjemplo = append(defectosDelEjemplo, fmt.Sprintf(
			"laboratorio/producto-descripciones-cortas: las descripciones suman %d caracteres (máximo 1700)",
			conLaEtiqueta))
	}

	casos := []struct {
		nombre   string
		skills   []string
		variante map[string]string
		defectos []string
	}{
		{
			nombre:   "las cuatro de la variante, sin defectos",
			skills:   lasCuatroSkills,
			variante: lasDelContrato,
		},
		{
			nombre:   "falta una",
			skills:   lasCuatroSkills,
			variante: sinLaEntrada(lasDelContrato, "plazos-y-territorio.txt"),
			defectos: []string{
				"laboratorio/producto-descripciones-cortas: falta la descripción de plazos-y-territorio",
			},
		},
		{
			nombre:   "sobra otra.txt",
			skills:   lasCuatroSkills,
			variante: conLaEntrada(lasDelContrato, "otra.txt", "otra\n"),
			defectos: []string{
				"laboratorio/producto-descripciones-cortas: «otra.txt» no es la descripción de ninguna skill de skills/",
			},
		},
		{
			nombre:   "una de 451 runas",
			skills:   lasCuatroSkills,
			variante: cuatroDe(1, 1, 451, 1),
			defectos: []string{
				"laboratorio/producto-descripciones-cortas: la descripción de normativa tiene 451 caracteres (máximo 450)",
			},
		},
		{
			nombre:   "cuatro de 450, 450, 400 y 401",
			skills:   lasCuatroSkills,
			variante: cuatroDe(450, 450, 400, 401),
			defectos: []string{
				"laboratorio/producto-descripciones-cortas: las descripciones suman 1701 caracteres (máximo 1700)",
			},
		},
		{
			nombre:   "cuatro de 450, 450, 400 y 400",
			skills:   lasCuatroSkills,
			variante: cuatroDe(450, 450, 400, 400),
		},
		{
			nombre: "una que empieza por <ejemplo>",
			skills: lasCuatroSkills,
			variante: conLaEntrada(lasDelContrato, "normativa.txt",
				etiquetaDelEjemplo+descripcionCortaDeNormativa+"\n"),
			defectos: defectosDelEjemplo,
		},
		{
			nombre:   "una quinta skill sin descripción",
			skills:   append(slices.Clone(lasCuatroSkills), "quinta"),
			variante: lasDelContrato,
			defectos: []string{
				"laboratorio/producto-descripciones-cortas: falta la descripción de quinta",
			},
		},
		{
			nombre: "el ejemplo de cinco líneas",
			skills: append(slices.Clone(lasCuatroSkills), "quinta"),
			variante: conLaEntrada(
				conLaEntrada(sinLaEntrada(lasDelContrato, "plazos-y-territorio.txt"), "normativa.txt",
					"<ejemplo>"+strings.Repeat("é", 451-len("<ejemplo>"))+"\n"),
				"otra.txt", "otra\n"),
			defectos: []string{
				"laboratorio/producto-descripciones-cortas: la descripción de normativa tiene 451 caracteres (máximo 450)",
				"laboratorio/producto-descripciones-cortas: la descripción de normativa tiene la etiqueta XML " +
					"«<ejemplo>» (no admite ninguna)",
				"laboratorio/producto-descripciones-cortas: falta la descripción de plazos-y-territorio",
				"laboratorio/producto-descripciones-cortas: falta la descripción de quinta",
				"laboratorio/producto-descripciones-cortas: «otra.txt» no es la descripción de ninguna skill de skills/",
			},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			raiz := montarElArbolDeLasDescripciones(t, caso.skills, caso.variante)

			textos, defectos, err := descripcionesCortas(raiz)
			require.NoError(t, err)

			compararLasLineasDeLosDefectos(t, caso.defectos, defectos)

			for nombre, contenido := range caso.variante {
				skill, esTxt := strings.CutSuffix(nombre, ".txt")
				if esTxt && slices.Contains(caso.skills, skill) {
					assert.Equal(t, strings.TrimSuffix(contenido, "\n"), textos[skill],
						"el texto de %s es su fichero sin el salto de línea final", skill)
				}
			}
		})
	}

	errores := []struct {
		nombre string
		quitar string
	}{
		{nombre: "sin skills/", quitar: directorioDeLasSkills},
		{nombre: "sin la variante", quitar: filepath.FromSlash(carpetaDeLasDescripcionesCortas)},
	}

	for _, caso := range errores {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			raiz := montarElArbolDeLasDescripciones(t, lasCuatroSkills, lasDelContrato)
			require.NoError(t, os.RemoveAll(filepath.Join(raiz, caso.quitar)))

			textos, defectos, err := descripcionesCortas(raiz)
			require.Error(t, err, "un directorio que no existe es el error de la función")
			assert.Nil(t, textos)
			assert.Nil(t, defectos)
		})
	}
}

// TestDescripcionesCortasPorLaEstructura comprueba lo que descripcionesCortas
// hace con una entrada que no es lo que su nombre dice (contracts/variante.md
// §3; research D3): un fichero suelto en skills/ no es una skill, así que ni se
// le exige descripción ni da texto; y un <skill>.txt que es una carpeta no se
// puede leer, que es el error de la función, sin textos ni defectos, y no un
// defecto de la variante (ADR 0023).
func TestDescripcionesCortasPorLaEstructura(t *testing.T) {
	t.Parallel()

	t.Run("un fichero suelto en skills/", func(t *testing.T) {
		t.Parallel()

		raiz := montarElArbolDeLasDescripciones(t, lasCuatroSkills, cuatroDe(1, 2, 3, 4))
		require.NoError(t, os.WriteFile(filepath.Join(raiz, directorioDeLasSkills, "LEEME.md"), []byte("# Skills\n"), 0o600))

		textos, defectos, err := descripcionesCortas(raiz)
		require.NoError(t, err)

		compararLasLineasDeLosDefectos(t, nil, defectos)
		assert.Equal(t, lasCuatroSkills, slices.Sorted(maps.Keys(textos)), "solo las carpetas de skills/ son skills")
	})

	t.Run("un txt que es una carpeta", func(t *testing.T) {
		t.Parallel()

		raiz := montarElArbolDeLasDescripciones(t, lasCuatroSkills, sinLaEntrada(cuatroDe(1, 2, 3, 4), "normativa.txt"))
		require.NoError(t, os.Mkdir(filepath.Join(raiz, filepath.FromSlash(carpetaDeLasDescripcionesCortas),
			"normativa.txt"), 0o750))

		textos, defectos, err := descripcionesCortas(raiz)
		require.ErrorIs(t, err, syscall.EISDIR)
		require.ErrorContains(t, err, "el texto de la variante producto-descripciones-cortas no se puede leer")
		assert.Nil(t, textos)
		assert.Nil(t, defectos)
	})
}

// skillDePruebaDeLaComposicion es un SKILL.md mínimo con la clave description
// en su frontmatter, el de las copias de TestComponerLasDescripciones.
const skillDePruebaDeLaComposicion = "---\nname: prueba\ndescription: La de antes.\n---\n\n# Prueba\n"

// TestComponerLasDescripciones comprueba la composición de la copia de las
// skills de una variante (FR-011; contracts/arnes.md §3; research D5), sobre
// copias montadas en un temporal: con un texto por skill, cada SKILL.md queda
// con el suyo como description y con su cuerpo; y una copia que no existe, una
// skill sin texto, una sin SKILL.md y un SKILL.md sin la clave description son
// el error de la función, que nombra lo que no se pudo hacer, y el SKILL.md que
// no se pudo componer queda como estaba.
func TestComponerLasDescripciones(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		comprobar func(t *testing.T)
	}{
		{"con un texto por skill", componerConUnTextoPorSkill},
		{"una copia que no existe", componerUnaCopiaQueNoExiste},
		{"una skill sin texto", componerUnaSkillSinTexto},
		{"una skill sin SKILL.md", componerUnaSkillSinFichero},
		{"un SKILL.md sin la clave", componerUnSkillSinLaClave},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			caso.comprobar(t)
		})
	}
}

// montarLaCopiaDeLaComposicion monta en un temporal la copia de las skills de
// una variante, con una carpeta por nombre y en cada una el SKILL.md de ese
// contenido, o ninguno si el contenido es vacío; devuelve su ruta.
func montarLaCopiaDeLaComposicion(t *testing.T, contenidos map[string]string) string {
	t.Helper()

	copia := t.TempDir()

	for skill, contenido := range contenidos {
		require.NoError(t, os.Mkdir(filepath.Join(copia, skill), 0o750))

		if contenido != "" {
			require.NoError(t, os.WriteFile(filepath.Join(copia, skill, ficheroDeLaSkillDeLaVariante),
				[]byte(contenido), 0o600))
		}
	}

	return copia
}

// skillDeLaCopia es el contenido del SKILL.md de esa skill de la copia.
func skillDeLaCopia(t *testing.T, copia, skill string) string {
	t.Helper()

	contenido, err := leerFichero(filepath.Join(copia, skill, ficheroDeLaSkillDeLaVariante))
	require.NoError(t, err)

	return string(contenido)
}

func componerConUnTextoPorSkill(t *testing.T) {
	t.Helper()

	copia := montarLaCopiaDeLaComposicion(t, map[string]string{
		"primera": skillDePruebaDeLaComposicion, "segunda": skillDePruebaDeLaComposicion,
	})
	textos := map[string]string{"primera": "La primera: con dos puntos.", "segunda": "La segunda."}

	require.NoError(t, componerLasDescripciones(copia, textos))

	_, cuerpo := descripcionYCuerpo(skillDePruebaDeLaComposicion)

	for skill, texto := range textos {
		compuesto := skillDeLaCopia(t, copia, skill)

		frontmatter, err := skills.LeerFrontmatter([]byte(compuesto))
		require.NoError(t, err, "el compuesto de %s:\n%s", skill, compuesto)
		assert.Equal(t, texto, frontmatter.Description, "la description de %s", skill)

		_, delCompuesto := descripcionYCuerpo(compuesto)
		assert.Equal(t, cuerpo, delCompuesto, "el cuerpo de %s", skill)
	}
}

func componerUnaCopiaQueNoExiste(t *testing.T) {
	t.Helper()

	copia := filepath.Join(t.TempDir(), "sin-copiar")

	err := componerLasDescripciones(copia, map[string]string{"primera": "La primera."})
	require.ErrorIs(t, err, fs.ErrNotExist)
	require.ErrorContains(t, err, "las skills de la copia "+copia+" no se pueden leer")
}

func componerUnaSkillSinTexto(t *testing.T) {
	t.Helper()

	copia := montarLaCopiaDeLaComposicion(t, map[string]string{
		"primera": skillDePruebaDeLaComposicion, "segunda": skillDePruebaDeLaComposicion,
	})

	err := componerLasDescripciones(copia, map[string]string{"primera": "La primera."})
	require.EqualError(t, err, "la skill segunda de la copia "+copia+" no tiene descripción en "+
		"laboratorio/producto-descripciones-cortas")
	assert.Equal(t, skillDePruebaDeLaComposicion, skillDeLaCopia(t, copia, "segunda"),
		"el SKILL.md de la skill sin texto queda como estaba")
}

func componerUnaSkillSinFichero(t *testing.T) {
	t.Helper()

	copia := montarLaCopiaDeLaComposicion(t, map[string]string{"primera": ""})

	err := componerLasDescripciones(copia, map[string]string{"primera": "La primera."})
	require.ErrorIs(t, err, fs.ErrNotExist)
	require.ErrorContains(t, err, "el SKILL.md "+filepath.Join(copia, "primera", ficheroDeLaSkillDeLaVariante)+
		" no se puede leer")
	assert.NoFileExists(t, filepath.Join(copia, "primera", ficheroDeLaSkillDeLaVariante), "no se escribe el que no había")
}

func componerUnSkillSinLaClave(t *testing.T) {
	t.Helper()

	const sinLaClave = "---\nname: prueba\n---\n\n# Prueba\n"

	copia := montarLaCopiaDeLaComposicion(t, map[string]string{"primera": sinLaClave})

	err := componerLasDescripciones(copia, map[string]string{"primera": "La primera."})
	require.ErrorIs(t, err, errSinLaClaveDeLaDescripcion)
	require.ErrorContains(t, err, "el SKILL.md "+filepath.Join(copia, "primera", ficheroDeLaSkillDeLaVariante)+
		" no se puede componer")
	assert.Equal(t, sinLaClave, skillDeLaCopia(t, copia, "primera"), "el SKILL.md que no se compone queda como estaba")
}

// compararLasLineasDeLosDefectos falla con la primera línea de los defectos
// obtenidos que difiere de la esperada, o con la primera que sobra o falta.
func compararLasLineasDeLosDefectos(t *testing.T, esperadas, obtenidas []string) {
	t.Helper()

	for indice := range max(len(esperadas), len(obtenidas)) {
		var esperada, obtenida string

		if indice < len(esperadas) {
			esperada = esperadas[indice]
		}

		if indice < len(obtenidas) {
			obtenida = obtenidas[indice]
		}

		require.Equal(t, esperada, obtenida, "línea %d de los defectos (todos: %q)", indice+1, obtenidas)
	}
}

// skillsDelArbol son los subdirectorios de skills/ de la raíz, en orden de
// nombre.
func skillsDelArbol(t *testing.T, raiz string) []string {
	t.Helper()

	entradas, err := os.ReadDir(filepath.Join(raiz, directorioDeLasSkills))
	require.NoError(t, err)

	var skills []string

	for _, entrada := range entradas {
		if entrada.IsDir() {
			skills = append(skills, entrada.Name())
		}
	}

	require.NotEmpty(t, skills, "skills/ tiene alguna skill")

	return skills
}

// montarElArbolDeLasDescripciones monta en un temporal una raíz con skills/ y
// un subdirectorio vacío por skill, y la variante con cada fichero de su nombre
// y su contenido tal cual; devuelve la raíz.
func montarElArbolDeLasDescripciones(t *testing.T, skills []string, variante map[string]string) string {
	t.Helper()

	raiz := t.TempDir()

	for _, skill := range skills {
		require.NoError(t, os.MkdirAll(filepath.Join(raiz, directorioDeLasSkills, skill), 0o750))
	}

	directorio := filepath.Join(raiz, filepath.FromSlash(carpetaDeLasDescripcionesCortas))
	require.NoError(t, os.MkdirAll(directorio, 0o750))

	for nombre, contenido := range variante {
		require.NoError(t, os.WriteFile(filepath.Join(directorio, nombre), []byte(contenido), 0o600))
	}

	return raiz
}

// cuatroDe es una variante de las cuatro skills con descripciones de esas
// longitudes en runas, en el orden de lasCuatroSkills, hechas de «é», que ocupa
// dos bytes, para que contar bytes no dé las mismas longitudes.
func cuatroDe(longitudes ...int) map[string]string {
	variante := make(map[string]string, len(longitudes))

	for indice, longitud := range longitudes {
		variante[lasCuatroSkills[indice]+".txt"] = strings.Repeat("é", longitud) + "\n"
	}

	return variante
}

// conLaEntrada es una copia de la variante con la entrada de ese nombre y ese
// contenido, la añada o la sustituya.
func conLaEntrada(variante map[string]string, nombre, contenido string) map[string]string {
	copia := maps.Clone(variante)
	copia[nombre] = contenido

	return copia
}

// sinLaEntrada es una copia de la variante sin la entrada de ese nombre.
func sinLaEntrada(variante map[string]string, nombre string) map[string]string {
	copia := maps.Clone(variante)
	delete(copia, nombre)

	return copia
}

// textoDeLaComposicion es un texto con que se compone un SKILL.md en
// TestConLaDescripcion, con el nombre de su caso.
type textoDeLaComposicion struct {
	nombre string
	texto  string
}

// textosQueYAMLCitaOCorta son los textos de TestConLaDescripcion además del de
// cada skill en la variante (FR-011, caso límite; research D12): cada uno con lo
// que YAML no lee como el texto si se escribe sin codificar detrás de la clave
// —«: » abre otro mapa, « #» abre un comentario, una comilla al principio abre
// un escalar entrecomillado— o lo que el emisor no puede cambiar, «»; y todos
// juntos.
var textosQueYAMLCitaOCorta = []textoDeLaComposicion{
	{"con dos puntos", "Plazos del procedimiento común: los de la LPAC, art. 30: días hábiles."},
	{"con comillas simples", "'Vigente' es la versión de hoy del BOE, y la 'anterior', la de otra fecha."},
	{"con comillas dobles", `"LPAC" es la Ley 39/2015 y "LRJSP", la Ley 40/2015.`},
	{"con almohadilla", "Cita el artículo 21 #2 de la norma y su bloque #a21."},
	{"con comillas angulares", "La «norma» consolidada se cita con su «bloque» y su «versión»."},
	{"con todo", `'Art. 21': el bloque #a21 de la «LPAC», "consolidada" y 'vigente'.`},
}

// TestConLaDescripcion comprueba la composición de un SKILL.md con otra
// description (FR-011; SC-002; contracts/arnes.md §3; research D5 y D12): con
// cada SKILL.md de skills/ del repositorio —tal cual y con las líneas sangradas
// de su description empezando por un tabulador— y, por cada uno, el texto de su
// skill en la variante producto-descripciones-cortas y los de
// textosQueYAMLCitaOCorta, la description que lee skills.LeerFrontmatter del
// compuesto es el texto exacto; skills.ValidarFrontmatter, con la raíz del
// repositorio y los applets que declara el propio frontmatter como registrados,
// no le da ningún defecto; y el cuerpo de descripcionYCuerpo del compuesto es,
// byte a byte, el del original. Un SKILL.md sin frontmatter, o sin la clave
// description en él, es el error de la función, sin contenido.
func TestConLaDescripcion(t *testing.T) {
	t.Parallel()

	variante, defectos, err := descripcionesCortas(raizDelRepositorio)
	require.NoError(t, err)
	compararLasLineasDeLosDefectos(t, nil, defectos)

	arbol := os.DirFS(filepath.Join(raizDelRepositorio, directorioDeLasSkills))

	for _, skill := range skillsDelArbol(t, raizDelRepositorio) {
		contenido, err := fs.ReadFile(arbol, path.Join(skill, ficheroDeLaSkillDeLaVariante))
		require.NoError(t, err)

		// Cada SKILL.md se compone tal cual y con su description sangrada con
		// tabuladores, que la clasificación también cuenta como de la clave.
		deLaSkill := []struct {
			nombre    string
			contenido string
		}{
			{"tal cual", string(contenido)},
			{"sangrada con tabuladores", conLaDescripcionSangradaConTabuladores(t, string(contenido))},
		}

		textos := slices.Concat([]textoDeLaComposicion{{"la de la variante", variante[skill]}}, textosQueYAMLCitaOCorta)

		for _, original := range deLaSkill {
			for _, caso := range textos {
				t.Run(skill+"/"+original.nombre+"/"+caso.nombre, func(t *testing.T) {
					t.Parallel()

					require.NotEmpty(t, caso.texto, "el texto de la composición")

					compuesto, err := skillConLaDescripcion([]byte(original.contenido), caso.texto)
					require.NoError(t, err)

					frontmatter, err := skills.LeerFrontmatter(compuesto)
					require.NoError(t, err, "el compuesto:\n%s", compuesto)
					assert.Equal(t, caso.texto, frontmatter.Description, "la description del compuesto:\n%s", compuesto)

					deLaValidacion, err := skills.ValidarFrontmatter(raizDelRepositorio, skill, frontmatter,
						frontmatter.DeclaracionDeKitlegal().Applets)
					require.NoError(t, err)

					var lineas []string
					for _, defecto := range deLaValidacion {
						lineas = append(lineas, defecto.Error())
					}

					compararLasLineasDeLosDefectos(t, nil, lineas)

					_, delOriginal := descripcionYCuerpo(original.contenido)
					_, delCompuesto := descripcionYCuerpo(string(compuesto))
					assert.Equal(t, delOriginal, delCompuesto, "el cuerpo del compuesto")
				})
			}
		}
	}

	sinLaClave := []struct {
		nombre    string
		contenido string
	}{
		{"vacío", ""},
		{"sin frontmatter", "# Normativa\n\nEl cuerpo.\n"},
		{"la clave fuera del frontmatter", "name: normativa\ndescription: Consulta el BOE.\n\n# Normativa\n"},
		{"un frontmatter sin la clave", "---\nname: normativa\n---\n\n# Normativa\n"},
		{"la clave dentro de metadata", "---\nname: normativa\nmetadata:\n  description: Consulta el BOE.\n---\n"},
		{"la clave en el cuerpo", "---\nname: normativa\n---\n\ndescription: Consulta el BOE.\n"},
	}

	for _, caso := range sinLaClave {
		t.Run("sin la clave/"+caso.nombre, func(t *testing.T) {
			t.Parallel()

			compuesto, err := skillConLaDescripcion([]byte(caso.contenido), descripcionCortaDeNormativa)
			require.ErrorIs(t, err, errSinLaClaveDeLaDescripcion)
			assert.Nil(t, compuesto)
		})
	}
}

// conLaDescripcionSangradaConTabuladores es el contenido de un SKILL.md con cada
// línea sangrada de su description, las que siguen a la de la clave, empezando
// por un tabulador en lugar de por sus espacios.
func conLaDescripcionSangradaConTabuladores(t *testing.T, contenido string) string {
	t.Helper()

	descripcion, _ := descripcionYCuerpo(contenido)

	var tabulada strings.Builder

	for linea := range strings.Lines(descripcion) {
		if strings.HasPrefix(linea, " ") {
			linea = "\t" + strings.TrimLeft(linea, " ")
		}

		tabulada.WriteString(linea)
	}

	require.NotEqual(t, descripcion, tabulada.String(), "la description tiene líneas sangradas")

	return strings.Replace(contenido, descripcion, tabulada.String(), 1)
}
