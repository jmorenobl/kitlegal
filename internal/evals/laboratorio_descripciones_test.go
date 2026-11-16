package evals

import (
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				"<ejemplo> "+descripcionCortaDeNormativa+"\n"),
			defectos: []string{
				"laboratorio/producto-descripciones-cortas: la descripción de normativa tiene la etiqueta XML " +
					"«<ejemplo>» (no admite ninguna)",
			},
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
