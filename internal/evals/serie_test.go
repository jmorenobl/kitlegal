package evals

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json/v2"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// La serie de cierres del ADR 0038, punto 6 (scripts/workflow/serie.sh): lo que
// el test copia, lo que versiona el repositorio y lo que el guion lee de él.
const (
	guionDeLaSerie = "serie.sh"

	rutaDelEsquemaDeLaSerie = "../../schemas/serie-de-cierres.json"
	rutaDeLaSeriePrevia     = "../../evidencias/adr-0038/serie-previa.jsonl"
	rutaDeLasFuentes        = "../../evidencias/adr-0038/serie-previa-fuentes.tsv"
	patronDeLasSeriesDeHito = "../../specs/*/gates/serie.jsonl"

	// La skill del repositorio del test, con su juez y sus dos evals que activan
	// la skill: una que nombra el precepto y otra por materia.
	skillDeLaSerie      = "boe-legislacion"
	evalQueNombra       = "01-nombra.yaml"
	evalPorMateria      = "02-por-materia.yaml"
	evalSinBinario      = "03-sin-binario.yaml"
	evalQueNoActiva     = "04-no-activa.yaml"
	juezDeLaSerie       = "claude-opus-5-5"
	versionDelJuezSerie = "2.1.289"
	rubricaDeLaSerie    = "# Rúbrica del test\n"
	casosDeLaSerie      = "casos: []\n"

	claseQueDecideDeLaSerie  = "afirma_lo_no_leido"
	claseQuePublicaDeLaSerie = "cuenta_su_proceso"
	umbralDeBoe              = "0.028"
	umbralDeJuris            = "0.017"
	hitoDeLaSerie            = "H0"
	fechaDeLaSerie           = "2026-10-13T12:33:08Z"
	propuestaDeSerie         = 140
)

// filaDeLaSerie es una fila de la serie, con las claves de
// schemas/serie-de-cierres.json.
type filaDeLaSerie struct {
	Skill            string              `json:"skill"`
	Procedencia      procedenciaDeLaFila `json:"procedencia"`
	ArbolDeLaSkill   string              `json:"arbol_de_la_skill"`
	ModeloQueDecide  *string             `json:"modelo_que_decide"`
	Juez             *huellaDelJuez      `json:"juez"`
	PorMateriaDe     *string             `json:"por_materia_de"`
	EvalsPorMateria  []string            `json:"evals_por_materia"`
	SinJuzgar        *int                `json:"sin_juzgar"`
	SesionesSinMedir *int                `json:"sesiones_sin_medir"`
	Cuenta           bool                `json:"cuenta"`
	NoCuentaPorque   *string             `json:"no_cuenta_porque"`
	Clases           []claseDeLaFila     `json:"clases"`
}

type procedenciaDeLaFila struct {
	Hito      string `json:"hito"`
	Medicion  int    `json:"medicion"`
	Fecha     string `json:"fecha"`
	Commit    string `json:"commit"`
	Propuesta *int   `json:"propuesta"`
	Ejecucion int    `json:"ejecucion"`
	Trabajo   int    `json:"trabajo"`
}

type huellaDelJuez struct {
	Modelo              string `json:"modelo"`
	VersionDeClaudeCode string `json:"version_de_claude_code"`
	RubricaSHA256       string `json:"rubrica_sha256"`
	CasosSHA256         string `json:"casos_sha256"`
}

type claseDeLaFila struct {
	Clase    string          `json:"clase"`
	Marcadas int             `json:"marcadas"`
	Total    int             `json:"total"`
	Celdas   []celdaDeLaFila `json:"celdas"`
}

type celdaDeLaFila struct {
	Modo       Modo `json:"modo"`
	PorMateria bool `json:"por_materia"`
	Marcadas   int  `json:"marcadas"`
	Total      int  `json:"total"`
}

// sesionDeCierre es una sesión del informe del job que el test escribe: su
// eval, su modelo, su modo, si su eval activa la skill y es por materia, si el
// juez la marca en la clase que decide o solo en la que se publica, y si quedó
// sin medir.
type sesionDeCierre struct {
	eval, modelo     string
	modo             Modo
	activa           bool
	porMateria       bool
	marcada, publica bool
	sinMedir         bool
}

// nombre es el de su sesión: su eval sin la extensión, su modo, su modelo y su
// número, sin que nada lo parta.
func (s sesionDeCierre) nombre(n int) string {
	return fmt.Sprintf("%s-%s-%s-%02d", strings.TrimSuffix(s.eval, ".yaml"), s.modo, s.modelo, n)
}

// sesionesDeCierre son las del informe del caso normal: en cada modo, dos de
// la eval que nombra el precepto y una de la eval por materia del modelo que
// decide —en el modo orden, el juez marca una de cada; en el modo herramienta,
// ninguna, y una solo en la clase que se publica—, y además cuatro que no
// cuentan aunque el juez las marque: una del modelo informativo, una de la eval
// sin binario ni servidor y una de una eval que no activa la skill.
func sesionesDeCierre() []sesionDeCierre {
	return []sesionDeCierre{
		{eval: evalQueNombra, modelo: modeloSonnet55, modo: ModoOrden, activa: true, marcada: true},
		{eval: evalQueNombra, modelo: modeloSonnet55, modo: ModoOrden, activa: true},
		{eval: evalPorMateria, modelo: modeloSonnet55, modo: ModoOrden, activa: true, porMateria: true, marcada: true},
		{eval: evalQueNombra, modelo: modeloSonnet55, modo: ModoHerramienta, activa: true, publica: true},
		{eval: evalQueNombra, modelo: modeloSonnet55, modo: ModoHerramienta, activa: true},
		{eval: evalPorMateria, modelo: modeloSonnet55, modo: ModoHerramienta, activa: true, porMateria: true},
		{eval: evalQueNombra, modelo: modeloHaiku45, modo: ModoOrden, activa: true, marcada: true},
		{eval: evalSinBinario, modelo: modeloSonnet55, activa: true, marcada: true},
		{eval: evalQueNoActiva, modelo: modeloSonnet55, modo: ModoOrden, marcada: true},
	}
}

// celdasDeLaSerie son las cuatro celdas de la fila de sesionesDeCierre.
func celdasDeLaSerie() []celdaDeLaFila {
	return []celdaDeLaFila{
		{Modo: ModoOrden, PorMateria: false, Marcadas: 1, Total: 2},
		{Modo: ModoOrden, PorMateria: true, Marcadas: 1, Total: 1},
		{Modo: ModoHerramienta, PorMateria: false, Marcadas: 0, Total: 2},
		{Modo: ModoHerramienta, PorMateria: true, Marcadas: 0, Total: 1},
	}
}

// informeDeLaSerie es el informe del job de skillDeLaSerie con esas sesiones,
// escrito con el tipo Informe: sus resultados, las respuestas con algún voto
// afirmativo y, de la clase que decide, el recuento de cada modo y el umbral
// conjunto, contados como los cuenta el job; con conjunto falso, sin el umbral
// conjunto, como un informe anterior a H26.
func informeDeLaSerie(commit string, sesiones []sesionDeCierre, conjunto bool) Informe {
	informe := Informe{
		Skill: skillDeLaSerie, ModeloQueDecide: modeloSonnet55, ModelosInformativos: []string{modeloHaiku45},
		Commit: commit, Veredicto: VeredictoAprobado, Motivos: []string{}, SesionesSinMedir: []SesionSinMedir{},
		Juez: &JuezInformado{
			Modelo: juezDeLaSerie, VersionDeClaudeCode: versionDelJuezSerie,
			Respuestas: []RespuestaConVotos{}, SinJuzgar: []RespuestaSinJuzgar{}, VotosCortados: []VotoCortado{},
		},
	}

	marcadas := map[Modo]int{}
	total := map[Modo]int{}

	for n, s := range sesiones {
		resultado := ResultadoDeEval{
			Sesion: s.nombre(n), Eval: s.eval, Modelo: s.modelo, Modo: s.modo, Activa: s.activa, Activada: s.activa,
			PorMateria: s.porMateria, SesionTerminada: true,
		}

		if s.sinMedir {
			resultado.SinMedir = "límite de uso"
			informe.SesionesSinMedir = append(informe.SesionesSinMedir,
				SesionSinMedir{Sesion: resultado.Sesion, Eval: s.eval, Modelo: s.modelo, Motivo: resultado.SinMedir})
		}

		informe.Evals = append(informe.Evals, resultado)

		if s.marcada || s.publica {
			informe.Juez.Respuestas = append(informe.Juez.Respuestas, RespuestaConVotos{
				Sesion: resultado.Sesion,
				Clases: []JuicioDeClase{
					{Clase: claseQueDecideDeLaSerie, Marcada: s.marcada},
					{Clase: claseQuePublicaDeLaSerie, Marcada: s.publica},
				},
			})
		}

		if s.modelo != modeloSonnet55 || s.modo == "" || !s.activa || s.sinMedir {
			continue
		}

		total[s.modo]++

		if s.marcada {
			marcadas[s.modo]++
		}
	}

	for _, modo := range []Modo{ModoOrden, ModoHerramienta} {
		informe.Umbrales = append(informe.Umbrales, umbralDeLaSerie(claseQueDecideDeLaSerie+":"+modeloSonnet55+":"+string(modo),
			marcadas[modo], total[modo], false))
	}

	if conjunto {
		informe.Umbrales = append(informe.Umbrales, umbralDeLaSerie(claseQueDecideDeLaSerie+":"+modeloSonnet55,
			marcadas[ModoOrden]+marcadas[ModoHerramienta], total[ModoOrden]+total[ModoHerramienta], true))
	}

	return informe
}

// umbralDeLaSerie es un umbral de la clase que decide con su medida y su total.
func umbralDeLaSerie(nombre string, medida, total int, decide bool) Umbral {
	return Umbral{
		Nombre: nombre, Medida: float64(medida), Total: &total, Comparacion: comparacionMenorOIgual,
		Umbral: 0.028, Cumple: true, Decide: decide,
	}
}

// repositorioDeLaSerie deja en un temporal un repositorio de git con los guiones
// de la serie y del informe final, la skill del test con su juez y sus cuatro
// evals, el modelo que decide del job, lo que el informe final necesita del hito
// y los ficheros extra; lo commitea en main y da su directorio y su commit. El
// umbral es el de la clase que decide en clases.yaml.
func repositorioDeLaSerie(t *testing.T, umbral string, extra map[string]string) (directorio, commit string) {
	t.Helper()

	directorio = t.TempDir()

	raiz, err := os.OpenRoot(directorio)
	require.NoError(t, err)

	defer func() { require.NoError(t, raiz.Close()) }()

	ficheros := ficherosDeLaSerie(umbral)

	for nombre, contenido := range ficheros {
		require.NoError(t, raiz.MkdirAll(filepath.Dir(nombre), 0o750))
		require.NoError(t, raiz.WriteFile(nombre, []byte(contenido), 0o600))
	}

	require.NoError(t, raiz.MkdirAll("scripts/workflow", 0o750))
	copiarGuiones(t, raiz, guionDeLaSerie, comunDelWorkflow, filepath.Base(guionDelInformeFinal))

	ejecutarGit(t, directorio, "init", "-q", "-b", "main")
	ejecutarGit(t, directorio, "add", "-A")
	ejecutarGit(t, directorio, "commit", "-q", "-m", "el repositorio de la serie")

	for nombre, contenido := range extra {
		require.NoError(t, raiz.MkdirAll(filepath.Dir(nombre), 0o750))
		require.NoError(t, raiz.WriteFile(nombre, []byte(contenido), 0o600))
	}

	return directorio, ejecutarGit(t, directorio, "rev-parse", "HEAD")
}

// ficherosDeLaSerie son los de la skill del test con su juez, cuya clase que
// decide lleva el umbral dado, y sus cuatro evals; el modelo que decide del job;
// y lo que el informe final necesita del hito.
func ficherosDeLaSerie(umbral string) map[string]string {
	juez := "evals/" + skillDeLaSerie + "/juez/"
	medida := fmt.Sprintf(`{"modelo_del_juez": %q, "version_de_claude_code": %q, "rubrica": {"sha256": %q}, "casos": {"sha256": %q}}`,
		juezDeLaSerie, versionDelJuezSerie, sha256DeTexto(rubricaDeLaSerie), sha256DeTexto(casosDeLaSerie))

	return map[string]string{
		"skills/" + skillDeLaSerie + "/SKILL.md": "---\nname: " + skillDeLaSerie + "\n---\n",
		juez + "clases.yaml": "# Clases del juez del test.\nclases:\n  - nombre: " + claseQueDecideDeLaSerie + "\n    decide: true\n" +
			"    umbral: " + umbral + "\n  - nombre: " + claseQuePublicaDeLaSerie + "\n    decide: false\n    umbral: 0\n",
		juez + "rubrica.md":                               rubricaDeLaSerie,
		juez + "casos.yaml":                               casosDeLaSerie,
		juez + "medida.json":                              medida,
		"evals/" + skillDeLaSerie + "/" + evalQueNombra:   "pregunta: \"¿qué dice el art. 21 de la Ley 39/2015?\"\nactiva: true\n",
		"evals/" + skillDeLaSerie + "/" + evalPorMateria:  "pregunta: \"¿qué dice la Ley 39/2015 del plazo?\"\nactiva: true\npor_materia: true\n",
		"evals/" + skillDeLaSerie + "/" + evalSinBinario:  "pregunta: \"¿qué dice el art. 53?\"\nactiva: true\nsin_binario_ni_servidor: true\n",
		"evals/" + skillDeLaSerie + "/" + evalQueNoActiva: "pregunta: \"¿cómo se hace una tortilla?\"\nactiva: false\n",
		".github/workflows/evals.yml":                     "env:\n      MODELO_DE_EVALS: " + modeloSonnet55 + "\n",
		".specify/feature.json":                           `{"feature_directory": "` + directorioDelHitoDelInformeFinal + `"}`,
		ficheroDelHitoDelInformeFinal:                     hitoDelInformeFinal,
		directorioDelHitoDelInformeFinal + "/spec.md":     "",
		directorioDelHitoDelInformeFinal + "/plan.md":     "",
		directorioDelHitoDelInformeFinal + "/tasks.md":    "",
	}
}

// copiarGuiones copia al árbol los guiones del workflow dados, ejecutables.
func copiarGuiones(t *testing.T, raiz *os.Root, nombres ...string) {
	t.Helper()

	guiones, err := os.OpenRoot(guionesDelWorkflow)
	require.NoError(t, err)

	defer func() { require.NoError(t, guiones.Close()) }()

	for _, nombre := range nombres {
		guion, err := guiones.ReadFile(nombre)
		require.NoError(t, err)
		require.NoError(t, escribirEjecutable(raiz, "scripts/workflow/"+nombre, string(guion)))
	}
}

// ejecutarGit ejecuta git en el directorio, con el entorno de git del test, y
// da su salida sin el salto de línea final.
func ejecutarGit(t *testing.T, directorio string, argumentos ...string) string {
	t.Helper()

	// Los argumentos de git los compone el test, y el directorio es su temporal: no
	// hay entrada externa en la orden.
	//nolint:gosec // los argumentos de git los compone el test; no hay entrada externa.
	orden := exec.CommandContext(t.Context(), "git", argumentos...)
	orden.Dir, orden.Env = directorio, entornoDeGitDelTest()

	salida, err := orden.CombinedOutput()
	require.NoErrorf(t, err, "git %v: %s", argumentos, salida)

	return strings.TrimSpace(string(salida))
}

// ejecutarLaSerie ejecuta el serie.sh del repositorio con los argumentos dados y
// da su salida estándar, su salida de error y su código.
func ejecutarLaSerie(t *testing.T, directorio string, argumentos ...string) (salida, errores string, codigo int) {
	t.Helper()

	// El guion es la copia que el test acaba de escribir en su temporal y los
	// argumentos los compone el test: no hay entrada externa en la orden.
	//nolint:gosec // el guion lo escribe el test en su temporal y los argumentos los compone el test.
	orden := exec.CommandContext(t.Context(), filepath.Join(directorio, "scripts", "workflow", guionDeLaSerie), argumentos...)
	orden.Env = entornoDeGitDelTest()

	var deSalida, deError strings.Builder

	orden.Stdout, orden.Stderr = &deSalida, &deError

	if err := orden.Run(); err != nil {
		var fallo *exec.ExitError

		require.ErrorAs(t, err, &fallo, "el guion se ejecuta: %s", deError.String())

		return deSalida.String(), deError.String(), fallo.ExitCode()
	}

	return deSalida.String(), deError.String(), 0
}

// sha256DeTexto es la huella SHA-256 de un texto, en hexadecimal.
func sha256DeTexto(texto string) string {
	huella := sha256.Sum256([]byte(texto))

	return hex.EncodeToString(huella[:])
}

// procedenciaDeLaSerie es la procedencia que el test da a una fila.
func procedenciaDeLaSerie(medicion int, commit string) string {
	return fmt.Sprintf(`{"hito": %q, "medicion": %d, "fecha": %q, "commit": %q, "propuesta": %d, "ejecucion": 1, "trabajo": 2}`,
		hitoDeLaSerie, medicion, fechaDeLaSerie, commit, propuestaDeSerie)
}

// escribirInformeDeLaSerie escribe el informe en el repositorio, fuera de lo
// commiteado, y da su ruta. Con sinPorMateria, las sesiones no llevan la clave
// por_materia, como un informe anterior a ella.
func escribirInformeDeLaSerie(t *testing.T, directorio string, informe Informe, sinPorMateria bool) string {
	t.Helper()

	escrito, err := json.Marshal(informe)
	require.NoError(t, err)

	if sinPorMateria {
		var crudo map[string]any

		require.NoError(t, json.Unmarshal(escrito, &crudo))

		evals, esLista := crudo["evals"].([]any)
		require.True(t, esLista)

		for _, resultado := range evals {
			entrada, esObjeto := resultado.(map[string]any)
			require.True(t, esObjeto)
			delete(entrada, "por_materia")
		}

		escrito, err = json.Marshal(crudo)
		require.NoError(t, err)
	}

	ruta := filepath.Join(directorio, "informe.json")
	require.NoError(t, os.WriteFile(ruta, escrito, 0o600))

	return ruta
}

// validarContraElEsquemaDeLaSerie valida una fila contra el esquema publicado.
func validarContraElEsquemaDeLaSerie(t *testing.T, fila []byte) {
	t.Helper()

	esquema, err := compilarEsquemaPublicado(rutaDelEsquemaDeLaSerie, "de la serie de cierres")
	require.NoError(t, err)

	valor, err := jsonschema.UnmarshalJSON(bytes.NewReader(fila))
	require.NoError(t, err)
	require.NoErrorf(t, esquema.Validate(valor), "la fila tiene la forma del esquema: %s", fila)
}

// TestFilaDeLaSerie ejecuta `serie.sh fila` sobre un repositorio temporal con
// informes del job escritos con el tipo Informe y fija la fila de una medición
// (ADR 0038, punto 6): las respuestas medidas son las del modelo que decide, en
// un modo, de una eval que activa la skill; las marcadas, las que el juez marca
// en la clase que decide; se reparten por modo y por clase de pregunta, que da el
// campo por_materia de cada sesión o, en un informe anterior a él, los ficheros
// de la eval; una medición con una respuesta medida sin juzgar o una sesión del
// modelo que decide sin medir se registra y no cuenta; un trabajo sin informe
// también; y una fila que no casa con los umbrales del propio informe no sale.
func TestFilaDeLaSerie(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("jq"); err != nil {
		t.Skip("el guion de la serie lee con jq, que no está en el PATH")
	}

	completa := func(f filaDeLaSerie) {
		t.Helper()
		assert.True(t, f.Cuenta, "la medición completa cuenta")
		assert.Nil(t, f.NoCuentaPorque)
		require.Len(t, f.Clases, 1, "una clase por cada clase del juez que decide")
		assert.Equal(t, claseDeLaFila{Clase: claseQueDecideDeLaSerie, Marcadas: 2, Total: 6, Celdas: celdasDeLaSerie()}, f.Clases[0])
	}

	casos := []struct {
		nombre        string
		informe       func(commit string) Informe
		sinPorMateria bool
		sinInforme    bool
		exigir        func(f filaDeLaSerie)
		deError       []string
	}{
		{
			nombre:  "completa",
			informe: func(commit string) Informe { return informeDeLaSerie(commit, sesionesDeCierre(), true) },
			exigir: func(f filaDeLaSerie) {
				completa(f)
				assert.Equal(t, "informe", *f.PorMateriaDe)
				assert.Equal(t, []string{evalPorMateria}, f.EvalsPorMateria)
			},
		},
		{
			nombre:  "anterior-al-umbral-conjunto",
			informe: func(commit string) Informe { return informeDeLaSerie(commit, sesionesDeCierre(), false) },
			exigir:  completa,
		},
		{
			nombre:        "anterior-al-campo-por-materia",
			informe:       func(commit string) Informe { return informeDeLaSerie(commit, sesionesDeCierre(), true) },
			sinPorMateria: true,
			exigir: func(f filaDeLaSerie) {
				completa(f)
				assert.Equal(t, "yaml", *f.PorMateriaDe, "la fila dice de dónde sale la clase de pregunta")
				assert.Equal(t, []string{evalPorMateria}, f.EvalsPorMateria)
			},
		},
		{
			nombre: "una-respuesta-medida-sin-juzgar",
			informe: func(commit string) Informe {
				informe := informeDeLaSerie(commit, sesionesDeCierre(), true)
				informe.Juez.SinJuzgar = append(informe.Juez.SinJuzgar,
					RespuestaSinJuzgar{Sesion: sesionesDeCierre()[1].nombre(1), Motivo: "voto 2: tope agotado"})

				return informe
			},
			exigir: func(f filaDeLaSerie) {
				assert.False(t, f.Cuenta, "una medición incompleta se registra y no cuenta")
				assert.Equal(t, "1 respuesta sin juzgar", *f.NoCuentaPorque)
				assert.Equal(t, 1, *f.SinJuzgar)
				require.Len(t, f.Clases, 1, "lo medido se registra igual")
			},
		},
		{
			// La eval sin binario ni servidor no es de ningún modo: su respuesta
			// sin juzgar no deja la medición incompleta.
			nombre: "sin-juzgar-fuera-de-los-modos",
			informe: func(commit string) Informe {
				informe := informeDeLaSerie(commit, sesionesDeCierre(), true)
				informe.Juez.SinJuzgar = append(informe.Juez.SinJuzgar,
					RespuestaSinJuzgar{Sesion: sesionesDeCierre()[7].nombre(7), Motivo: "voto 1: tope agotado"})

				return informe
			},
			exigir: completa,
		},
		{
			nombre: "una-sesion-del-modelo-que-decide-sin-medir",
			informe: func(commit string) Informe {
				sesiones := append(sesionesDeCierre(), sesionDeCierre{
					eval: evalQueNombra, modelo: modeloSonnet55, modo: ModoOrden, activa: true, sinMedir: true,
				})

				return informeDeLaSerie(commit, sesiones, true)
			},
			exigir: func(f filaDeLaSerie) {
				assert.False(t, f.Cuenta)
				assert.Equal(t, "1 sesión del modelo que decide sin medir", *f.NoCuentaPorque)
			},
		},
		{
			nombre: "una-sesion-del-modelo-informativo-sin-medir",
			informe: func(commit string) Informe {
				sesiones := append(sesionesDeCierre(), sesionDeCierre{
					eval: evalQueNombra, modelo: modeloHaiku45, modo: ModoOrden, activa: true, sinMedir: true,
				})

				return informeDeLaSerie(commit, sesiones, true)
			},
			exigir: completa,
		},
		{
			nombre:     "sin-informe",
			sinInforme: true,
			exigir: func(f filaDeLaSerie) {
				assert.False(t, f.Cuenta)
				assert.Equal(t, "el registro del trabajo no trae un informe.json legible", *f.NoCuentaPorque)
				assert.Empty(t, f.Clases)
				assert.Nil(t, f.Juez, "lo que no se puede leer no se rellena")
				assert.Nil(t, f.ModeloQueDecide)
			},
		},
		{
			nombre: "no-casa-con-el-recuento-de-un-modo",
			informe: func(commit string) Informe {
				informe := informeDeLaSerie(commit, sesionesDeCierre(), false)
				informe.Umbrales[1].Medida = 1

				return informe
			},
			deError: []string{
				"no casa con los umbrales de su informe",
				claseQueDecideDeLaSerie + " en el modo herramienta: la serie cuenta 0 de 3 y el umbral `" +
					claseQueDecideDeLaSerie + ":" + modeloSonnet55 + ":herramienta` del informe, 1 de 3",
			},
		},
		{
			nombre: "no-casa-con-el-umbral-conjunto",
			informe: func(commit string) Informe {
				informe := informeDeLaSerie(commit, sesionesDeCierre(), true)
				total := 7
				informe.Umbrales[2].Total = &total

				return informe
			},
			deError: []string{
				claseQueDecideDeLaSerie + ": la serie cuenta 2 de 6 y el umbral conjunto `" + claseQueDecideDeLaSerie + ":" +
					modeloSonnet55 + "` del informe, 2 de 7",
			},
		},
		{
			nombre: "sin-ningun-umbral-de-la-clase",
			informe: func(commit string) Informe {
				informe := informeDeLaSerie(commit, sesionesDeCierre(), true)
				informe.Umbrales = nil

				return informe
			},
			deError: []string{"el informe no trae ni el umbral conjunto"},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			directorio, commit := repositorioDeLaSerie(t, umbralDeBoe, nil)
			arbol := ejecutarGit(t, directorio, "rev-parse", "HEAD:skills/"+skillDeLaSerie)

			argumentos := []string{"fila", skillDeLaSerie, procedenciaDeLaSerie(1, commit)}
			if !caso.sinInforme {
				argumentos = append(argumentos, escribirInformeDeLaSerie(t, directorio, caso.informe(commit), caso.sinPorMateria))
			}

			salida, deError, codigo := ejecutarLaSerie(t, directorio, argumentos...)
			exigirLaFila(t, caso.deError, caso.exigir, salida, deError, codigo, commit, arbol)
		})
	}
}

// exigirLaFila exige lo que el caso de TestFilaDeLaSerie espera de la salida de
// `serie.sh fila`: con errores, que salga con 1, los diga y no dé ninguna fila;
// sin ellos, una sola fila con la forma del esquema, la procedencia, el árbol de
// la skill y la huella del juez, y lo que el caso exige.
func exigirLaFila(t *testing.T, errores []string, exigir func(filaDeLaSerie), salida, deError string, codigo int,
	commit, arbol string,
) {
	t.Helper()

	if len(errores) > 0 {
		assert.Equal(t, 1, codigo, "una fila que no casa con su informe para el guion")
		assert.Empty(t, salida, "y no sale")

		for _, fragmento := range errores {
			assert.Contains(t, deError, fragmento)
		}

		return
	}

	require.Equalf(t, 0, codigo, "la fila sale: %s", deError)
	require.Equal(t, 1, strings.Count(salida, "\n"), "una fila, en una línea: %q", salida)

	validarContraElEsquemaDeLaSerie(t, []byte(salida))

	var fila filaDeLaSerie

	require.NoError(t, json.Unmarshal([]byte(salida), &fila))

	propuesta := propuestaDeSerie
	assert.Equal(t, procedenciaDeLaFila{
		Hito: hitoDeLaSerie, Medicion: 1, Fecha: fechaDeLaSerie, Commit: commit, Propuesta: &propuesta, Ejecucion: 1, Trabajo: 2,
	}, fila.Procedencia)
	assert.Equal(t, skillDeLaSerie, fila.Skill)
	assert.Equal(t, arbol, fila.ArbolDeLaSkill, "la skill como estaba: el árbol de skills/<skill>")

	if fila.Juez != nil {
		assert.Equal(t, huellaDelJuez{
			Modelo: juezDeLaSerie, VersionDeClaudeCode: versionDelJuezSerie,
			RubricaSHA256: sha256DeTexto(rubricaDeLaSerie), CasosSHA256: sha256DeTexto(casosDeLaSerie),
		}, *fila.Juez)
		assert.Equal(t, modeloSonnet55, *fila.ModeloQueDecide)
	}

	exigir(fila)
}

// filaContada es una fila de la serie que cuenta, con la clave dada, con las
// marcadas y el total dados en la celda del modo orden de las preguntas que
// nombran el precepto.
func filaContada(hito string, medicion int, clave filaDeLaSerie, marcadas, total int) filaDeLaSerie {
	fila := clave
	fila.Procedencia = procedenciaDeLaFila{
		Hito: hito, Medicion: medicion, Fecha: fmt.Sprintf("2026-10-%02dT10:00:00Z", 1+medicion), Commit: strings.Repeat("a", 40),
		Ejecucion: 1, Trabajo: 2,
	}
	cero := 0
	informe := "informe"
	fila.PorMateriaDe, fila.SinJuzgar, fila.SesionesSinMedir, fila.Cuenta = &informe, &cero, &cero, true
	fila.EvalsPorMateria = []string{}
	fila.Clases = []claseDeLaFila{{Clase: claseQueDecideDeLaSerie, Marcadas: marcadas, Total: total, Celdas: []celdaDeLaFila{
		{Modo: ModoOrden, Marcadas: marcadas, Total: total},
		{Modo: ModoOrden, PorMateria: true},
		{Modo: ModoHerramienta},
		{Modo: ModoHerramienta, PorMateria: true},
	}}}

	return fila
}

// claveDelRepositorio es la fila sin procedencia ni clases con la clave de la
// cabeza del repositorio de la serie: su árbol de la skill, el modelo que decide
// y la huella del juez.
func claveDelRepositorio(t *testing.T, directorio string) filaDeLaSerie {
	t.Helper()

	modelo := modeloSonnet55

	return filaDeLaSerie{
		Skill:           skillDeLaSerie,
		ArbolDeLaSkill:  ejecutarGit(t, directorio, "rev-parse", "HEAD:skills/"+skillDeLaSerie),
		ModeloQueDecide: &modelo,
		Juez: &huellaDelJuez{
			Modelo: juezDeLaSerie, VersionDeClaudeCode: versionDelJuezSerie,
			RubricaSHA256: sha256DeTexto(rubricaDeLaSerie), CasosSHA256: sha256DeTexto(casosDeLaSerie),
		},
	}
}

// lineasDeLaSerie son las filas en JSONL.
func lineasDeLaSerie(t *testing.T, filas ...filaDeLaSerie) string {
	t.Helper()

	var jsonl strings.Builder

	for _, fila := range filas {
		escrita, err := json.Marshal(fila)
		require.NoError(t, err)
		validarContraElEsquemaDeLaSerie(t, escrita)
		jsonl.Write(escrita)
		jsonl.WriteByte('\n')
	}

	return jsonl.String()
}

// mostrarLaSerie escribe las filas en el repositorio de la serie, la mitad en la
// semilla y la otra mitad en la serie de un hito, y da lo que escribe
// `serie.sh mostrar` de la skill del test.
func mostrarLaSerie(t *testing.T, umbral string, filas func(clave filaDeLaSerie) []filaDeLaSerie) string {
	t.Helper()

	directorio, _ := repositorioDeLaSerie(t, umbral, nil)
	todas := filas(claveDelRepositorio(t, directorio))
	mitad := len(todas) / 2

	for ruta, deLaRuta := range map[string][]filaDeLaSerie{
		"evidencias/adr-0038/serie-previa.jsonl": todas[:mitad],
		"specs/001-h1-hito/gates/serie.jsonl":    todas[mitad:],
	} {
		require.NoError(t, os.MkdirAll(filepath.Join(directorio, filepath.Dir(ruta)), 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(directorio, ruta), []byte(lineasDeLaSerie(t, deLaRuta...)), 0o600))
	}

	salida, deError, codigo := ejecutarLaSerie(t, directorio, "mostrar", skillDeLaSerie)
	require.Equalf(t, 0, codigo, "mostrar no falla: %s", deError)

	return salida
}

// TestMostrarLaSerie fija lo que `serie.sh mostrar` enseña de una skill, sin
// decidir nada (ADR 0038, puntos 3, 4 y 6): una fila por medición de la semilla y
// de las series de los hitos, con su sitio en la tasa; la tasa acumulada solo de
// las filas que cuentan con la clave de la cabeza; por debajo de 500 respuestas,
// «insuficiente»; con 500 o más, el escalón de la regla 3 con el total de hoy, y
// si permite estrechar el umbral, que lo baja una persona; si la tasa pasa del
// escalón del umbral de hoy, un aviso; y nunca que se suba.
func TestMostrarLaSerie(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("jq"); err != nil {
		t.Skip("el guion de la serie lee con jq, que no está en el PATH")
	}

	t.Run("solo-la-clave-de-la-cabeza", func(t *testing.T) {
		t.Parallel()

		salida := mostrarLaSerie(t, umbralDeBoe, func(clave filaDeLaSerie) []filaDeLaSerie {
			otraSkill := filaContada("H1", 3, clave, 5, 108)
			otraSkill.ArbolDeLaSkill = strings.Repeat("b", 40)
			otroJuez := filaContada("H1", 4, clave, 4, 108)
			otroJuez.Juez = &huellaDelJuez{
				Modelo: "otro", VersionDeClaudeCode: "1", RubricaSHA256: strings.Repeat("c", 64),
				CasosSHA256: strings.Repeat("d", 64),
			}
			incompleta := filaContada("H1", 5, clave, 3, 108)
			uno, porque := 1, "1 respuesta sin juzgar"
			incompleta.Cuenta, incompleta.SinJuzgar, incompleta.NoCuentaPorque = false, &uno, &porque
			otroModelo := filaContada("H1", 6, clave, 1, 108)
			modelo := "otro-modelo"
			otroModelo.ModeloQueDecide = &modelo

			return []filaDeLaSerie{
				filaContada("H1", 1, clave, 2, 108), otraSkill, otroJuez,
				filaContada("H1", 2, clave, 1, 108), incompleta, otroModelo,
			}
		})

		exigirLineas(t, salida,
			"| H1-1 | 2026-10-02 | `aaaaaaa` | `"+claveDelArbol(salida)+"` | 2/108 | — | 2/108 | — | 2/108 | sí |",
			"Tasa acumulada con la skill como está: 3 de 216 (1,39 %), de H1-1 y H1-2.",
			"Insuficiente: 216 de 500 respuestas. La regla 3 del ADR 0038 pide la tasa de 500 respuestas al menos: con menos no da escalón.")

		for _, enLaTasa := range []string{
			"| sí |", "| no: otra skill |", "| no: otro juez |", "| no: 1 respuesta sin juzgar |",
			"| no: otro modelo que decide |",
		} {
			assert.Contains(t, salida, enLaTasa)
		}

		assert.NotContains(t, salida, "Escalón", "por debajo de 500 respuestas no hay escalón")
	})

	casos := []struct {
		nombre     string
		umbral     string
		total      int
		filas      int
		marcadas   int
		escalon    string
		conclusion string
	}{
		{
			nombre: "tasa-del-escalon-3-con-108", umbral: umbralDeBoe, total: 108, filas: 5, marcadas: 5,
			escalon:    "como mucho 3 marcadas (vale hasta una tasa del 0,95 %). El umbral de hoy, 0.028 en `evals/boe-legislacion/juez/clases.yaml`, admite 3 de 108.",
			conclusion: "El umbral de hoy es el de la regla.",
		},
		{
			nombre: "tasa-del-escalon-2-con-108", umbral: umbralDeBoe, total: 108, filas: 5, marcadas: 2,
			escalon:    "como mucho 2 marcadas (vale hasta una tasa del 0,53 %)",
			conclusion: "Permite estrechar el umbral a 2 de 108. Lo baja una persona, con un diff de `evals/boe-legislacion/juez/clases.yaml` y estas cifras (ADR 0038, punto 4); ningún run lo cambia.",
		},
		{
			nombre: "tasa-del-escalon-1-con-108", umbral: umbralDeBoe, total: 108, filas: 5, marcadas: 1,
			escalon:    "como mucho 1 marcada (vale hasta una tasa del 0,20 %)",
			conclusion: "Permite estrechar el umbral a 1 de 108.",
		},
		{
			nombre: "tasa-por-encima-del-umbral-de-hoy", umbral: umbralDeBoe, total: 108, filas: 5, marcadas: 6,
			escalon:    "como mucho 4 marcadas",
			conclusion: "**Aviso**: la tasa acumulada, 1,11 %, pasa de la que vale para el umbral de hoy (3 de 108, hasta el 0,95 %)",
		},
		{
			nombre: "tasa-del-escalon-1-con-60", umbral: umbralDeJuris, total: 60, filas: 9, marcadas: 1,
			escalon:    "como mucho 1 marcada (vale hasta una tasa del 0,36 %). El umbral de hoy, 0.017 en `evals/boe-legislacion/juez/clases.yaml`, admite 1 de 60.",
			conclusion: "El umbral de hoy es el de la regla.",
		},
		{
			nombre: "tasa-del-escalon-2-con-60", umbral: umbralDeJuris, total: 60, filas: 9, marcadas: 5,
			escalon:    "como mucho 2 marcadas (vale hasta una tasa del 0,96 %)",
			conclusion: "**Aviso**: la tasa acumulada, 0,93 %, pasa de la que vale para el umbral de hoy (1 de 60, hasta el 0,36 %)",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			salida := mostrarLaSerie(t, caso.umbral, func(clave filaDeLaSerie) []filaDeLaSerie {
				filas := make([]filaDeLaSerie, 0, caso.filas)

				for n := range caso.filas {
					marcadas := 0
					if n < caso.marcadas {
						marcadas = 1
					}

					if n == caso.filas-1 {
						marcadas = caso.marcadas - min(caso.marcadas, caso.filas-1)
					}

					filas = append(filas, filaContada("H1", n+1, clave, marcadas, caso.total))
				}

				return filas
			})

			respuestas := caso.filas * caso.total
			assert.Contains(t, salida, fmt.Sprintf("Tasa acumulada con la skill como está: %d de %d", caso.marcadas, respuestas))
			assert.Contains(t, salida, fmt.Sprintf("Escalón de la regla 3 con %d respuestas, el total de hoy: %s", caso.total, caso.escalon))
			assert.Contains(t, salida, caso.conclusion)
			assert.NotContains(t, salida, "Insuficiente")
			exigirQueNuncaPropongaSubir(t, salida, caso.umbral, caso.total)
		})
	}
}

// claveDelArbol es el árbol de la skill, abreviado, de la línea de la clave de
// la cabeza que escribe mostrar.
func claveDelArbol(salida string) string {
	_, despues, _ := strings.Cut(salida, "Clave de la cabeza: skill `")
	arbol, _, _ := strings.Cut(despues, "`")

	return arbol
}

// exigirQueNuncaPropongaSubir exige que lo que escribe mostrar no proponga un
// umbral más alto que el de hoy: ni la palabra «subir» ni un estrechamiento a
// un número de marcadas que no sea menor que el que admite el umbral de hoy.
func exigirQueNuncaPropongaSubir(t *testing.T, salida, umbral string, total int) {
	t.Helper()

	assert.NotContains(t, strings.ToLower(salida), "subir")
	assert.NotContains(t, strings.ToLower(salida), "subirlo")

	valor, err := strconv.ParseFloat(umbral, 64)
	require.NoError(t, err)

	hoy := int(math.Floor(valor*float64(total) + 1e-9))

	for k := hoy; k <= total; k++ {
		assert.NotContains(t, salida, fmt.Sprintf("Permite estrechar el umbral a %d de %d", k, total))
	}
}

// TestEscalonesDeLaRegla fija la tasa más alta con la que vale cada escalón de la
// regla 3 del ADR 0038, como la da `serie.sh escalones`: con 108 respuestas, 3
// hasta el 0,95 %, 2 hasta el 0,53 % y 1 hasta el 0,20 %, las del ADR; con 60,
// la de jurisprudencia, 1 hasta el 0,36 % y 2 hasta el 0,96 %. Y que cada una es
// el borde de verdad: a esa tasa, una ejecución pasa del escalón como mucho 1 de
// cada 50 veces, y un poco por encima, más.
func TestEscalonesDeLaRegla(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("jq"); err != nil {
		t.Skip("el guion de la serie lee con jq, que no está en el PATH")
	}

	directorio, _ := repositorioDeLaSerie(t, umbralDeBoe, nil)

	for total, esperados := range map[int]map[int]string{
		108: {1: "0.20", 2: "0.53", 3: "0.95"},
		60:  {1: "0.36", 2: "0.96"},
	} {
		salida, deError, codigo := ejecutarLaSerie(t, directorio, "escalones", strconv.Itoa(total))
		require.Equalf(t, 0, codigo, "escalones: %s", deError)

		var escalones []struct {
			Escalon int     `json:"escalon"`
			Hasta   float64 `json:"hasta"`
		}

		require.NoError(t, json.Unmarshal([]byte(salida), &escalones))

		for _, escalon := range escalones {
			assert.LessOrEqual(t, colaBinomial(total, escalon.Hasta, escalon.Escalon), 0.02,
				"con %d, a la tasa del escalón %d, pasar de él como mucho 1 de cada 50 veces", total, escalon.Escalon)
			assert.Greater(t, colaBinomial(total, escalon.Hasta+1e-5, escalon.Escalon), 0.02,
				"con %d, por encima de la tasa del escalón %d, más", total, escalon.Escalon)

			if esperado, esta := esperados[escalon.Escalon]; esta {
				assert.Equal(t, esperado, strconv.FormatFloat(escalon.Hasta*100, 'f', 2, 64),
					"con %d respuestas, el escalón %d vale hasta el %s %%", total, escalon.Escalon, esperado)
			}
		}
	}
}

// colaBinomial es la probabilidad de que n respuestas, cada una marcada con
// probabilidad p, den más de k marcadas.
func colaBinomial(n int, p float64, k int) float64 {
	hasta := 0.0
	termino := math.Pow(1-p, float64(n))

	for i := 0; i <= k; i++ {
		hasta += termino
		termino *= float64(n-i) / float64(i+1) * p / (1 - p)
	}

	return 1 - hasta
}

// TestLaSerieVersionada comprueba en make ci la serie que versiona el
// repositorio: cada fila de la semilla y de la serie de cada hito tiene la forma
// de schemas/serie-de-cierres.json, ninguna medición de una skill está dos veces
// y la semilla tiene, en su orden, una fila por cada trabajo de su lista de
// fuentes, con su hito, su medición, su propuesta, su ejecución, su trabajo y su
// skill: la produce `serie.sh semilla` de esa lista, y nadie la escribe a mano.
func TestLaSerieVersionada(t *testing.T) {
	t.Parallel()

	rutas, err := filepath.Glob(patronDeLasSeriesDeHito)
	require.NoError(t, err)

	vistas := map[string]string{}

	var semilla []filaDeLaSerie

	for _, ruta := range append([]string{rutaDeLaSeriePrevia}, rutas...) {
		contenido, err := leerFichero(ruta)
		require.NoError(t, err)

		lineas := bufio.NewScanner(bytes.NewReader(contenido))
		lineas.Buffer(nil, 1<<20)

		for n := 1; lineas.Scan(); n++ {
			validarContraElEsquemaDeLaSerie(t, lineas.Bytes())

			var fila filaDeLaSerie

			require.NoError(t, json.Unmarshal(lineas.Bytes(), &fila))

			clave := fmt.Sprintf("%s %s-%d", fila.Skill, fila.Procedencia.Hito, fila.Procedencia.Medicion)
			if anterior, esta := vistas[clave]; esta {
				assert.Failf(t, "medición repetida", "%s está en %s y en %s:%d", clave, anterior, ruta, n)
			}

			vistas[clave] = fmt.Sprintf("%s:%d", ruta, n)

			if ruta == rutaDeLaSeriePrevia {
				semilla = append(semilla, fila)
			}
		}

		require.NoError(t, lineas.Err())
	}

	fuentes, err := leerFichero(rutaDeLasFuentes)
	require.NoError(t, err)

	var esperadas []string

	for linea := range strings.Lines(string(fuentes)) {
		if linea = strings.TrimSuffix(linea, "\n"); linea != "" && !strings.HasPrefix(linea, "#") {
			esperadas = append(esperadas, linea)
		}
	}

	leidas := make([]string, 0, len(semilla))
	for _, fila := range semilla {
		p := fila.Procedencia
		require.NotNil(t, p.Propuesta)
		leidas = append(leidas, strings.Join([]string{
			p.Hito, strconv.Itoa(p.Medicion), strconv.Itoa(*p.Propuesta), strconv.Itoa(p.Ejecucion), strconv.Itoa(p.Trabajo), fila.Skill,
		}, "\t"))
	}

	assert.Equal(t, esperadas, leidas, "una fila de la semilla por cada trabajo de sus fuentes, en su orden")
}

// TestLaSerieLeeLasEvalsComoElJob comprueba que lo que serie.sh lee de las evals
// del repositorio sin un lector de YAML es lo que leen el job y el juez: las
// clases de cada juez (`serie.sh clases`, contra leerClasesDelJuez) y las evals
// por materia de cada skill (`serie.sh por-materia`, contra LeerConjunto).
func TestLaSerieLeeLasEvalsComoElJob(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("jq"); err != nil {
		t.Skip("el guion de la serie lee con jq, que no está en el PATH")
	}

	raiz, err := filepath.Abs("../..")
	require.NoError(t, err)

	skills, err := os.ReadDir(filepath.Join(raiz, "evals"))
	require.NoError(t, err)

	conJuez := 0

	for _, entrada := range skills {
		if !entrada.IsDir() {
			continue
		}

		skill := entrada.Name()

		conjunto, err := LeerConjunto(filepath.Join(raiz, "evals", skill))
		require.NoError(t, err)

		porMateria := []string{}

		for _, eval := range conjunto.Evals {
			if eval.PorMateria {
				porMateria = append(porMateria, eval.Fichero)
			}
		}

		salida, deError, codigo := ejecutarLaSerie(t, raiz, "por-materia", skill)
		require.Equalf(t, 0, codigo, "por-materia %s: %s", skill, deError)

		var leidas []string

		require.NoError(t, json.Unmarshal([]byte(salida), &leidas))
		assert.Equal(t, porMateria, leidas, "las evals por materia de %s", skill)

		if conjunto.Juez == nil {
			continue
		}

		conJuez++

		contenido, err := leerFichero(filepath.Join(raiz, "evals", skill, "juez", "clases.yaml"))
		require.NoError(t, err)

		clases, err := leerClasesDelJuez(contenido)
		require.NoError(t, err)

		salida, deError, codigo = ejecutarLaSerie(t, raiz, "clases", skill)
		require.Equalf(t, 0, codigo, "clases %s: %s", skill, deError)

		var delGuion []struct {
			Nombre string  `json:"nombre"`
			Decide bool    `json:"decide"`
			Umbral float64 `json:"umbral"`
		}

		require.NoError(t, json.Unmarshal([]byte(salida), &delGuion))

		leidasDelGuion := make([]ClaseDelJuez, 0, len(delGuion))
		for _, clase := range delGuion {
			leidasDelGuion = append(leidasDelGuion, ClaseDelJuez(clase))
		}

		assert.Equal(t, clases, leidasDelGuion, "las clases del juez de %s", skill)
	}

	assert.Positive(t, conJuez, "hay skills con juez que comprobar")
}

// TestInformeFinalDeLaSerie ejecuta el informe final con --solo-ver sobre el
// repositorio de la serie y fija que su sección 3 enseña, tal cual, lo que
// `serie.sh mostrar` escribe de cada skill con juez: el mismo código dentro y
// fuera de un run.
func TestInformeFinalDeLaSerie(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("jq"); err != nil {
		t.Skip("el guion de la serie lee con jq, que no está en el PATH")
	}

	directorio, _ := repositorioDeLaSerie(t, umbralDeBoe, nil)
	clave := claveDelRepositorio(t, directorio)
	serie := filepath.Join(directorio, directorioDelHitoDelInformeFinal, "gates", "serie.jsonl")
	require.NoError(t, os.MkdirAll(filepath.Dir(serie), 0o750))
	require.NoError(t, os.WriteFile(serie, []byte(lineasDeLaSerie(t, filaContada("H0", 1, clave, 1, 108))), 0o600))

	mostrada, deError, codigo := ejecutarLaSerie(t, directorio, "mostrar")
	require.Equalf(t, 0, codigo, "mostrar: %s", deError)
	require.Contains(t, mostrada, "| H0-1 |")

	//nolint:gosec // el guion lo escribe el test en su temporal y los argumentos son literales.
	orden := exec.CommandContext(t.Context(),
		filepath.Join(directorio, "scripts", "workflow", filepath.Base(guionDelInformeFinal)), "H0", "--solo-ver")
	orden.Env = entornoDeGitDelTest()

	var salida, errores strings.Builder

	orden.Stdout, orden.Stderr = &salida, &errores

	require.NoErrorf(t, orden.Run(), "el guion escribe el informe final: %s", errores.String())

	_, tercera, esta := strings.Cut(salida.String(), "## 3. Evals sobre la cabeza\n")
	require.True(t, esta, "el informe final tiene su sección 3")

	tercera, _, esta = strings.Cut(tercera, "## 4. ")
	require.True(t, esta)
	assert.Contains(t, tercera, mostrada, "la sección 3 enseña la serie como mostrar")
}

// TestFilaConReutilizadas ejecuta tal cual el guion de la serie del workflow
// —scripts/workflow/serie.sh fila, en su sitio y sin copiarlo— sobre el informe
// de una ejecución con una eval reutilizada y sobre el de la misma ejecución con
// solo lo medido, sin la eval reutilizada ni decisión, los dos escritos con
// EscribirInforme y copiados en t.TempDir() con el commit de la cabeza del
// repositorio, del que el guion lee la skill y su juez (FR-023 y SC-009 de
// H30): los dos salen con 0 y dan, byte a byte, la misma fila, que cuenta solo
// las respuestas medidas. La reutilizada aporta una respuesta marcada en la
// clase que decide en cada modo, que la fila no ve.
func TestFilaConReutilizadas(t *testing.T) {
	t.Parallel()

	_, err := exec.LookPath("jq")
	require.NoError(t, err, "el guion de la serie lee con jq")

	raiz, err := filepath.Abs(filepath.Join(guionesDelWorkflow, "..", ".."))
	require.NoError(t, err)

	commit := ejecutarGit(t, raiz, "rev-parse", "HEAD")
	ejecucion := ejecucionDeDosEvals()

	copia := armarEjecucionConUmbrales(t, ejecucion)
	completa := informeDeLaCopia(t, copia, ejecucion.ajustar)

	decision := decisionQueReutilizaLaPrimera(t, completa.informe)
	marcarUnaRespuestaReutilizadaEnCadaModo(decision.Evals[0].Aportes)
	quitarLasSesionesDeLaEval(t, copia, evalQueSeReutiliza)

	conReutilizada := informeDeLaCopia(t, copia, func(entradas *InformeAEscribir) {
		ejecucion.ajustar(entradas)
		entradas.decision = &decision
	})
	require.Contains(t, umbralesPorNombre(conReutilizada.informe), conjuntoDelInforme+sufijoDeLasReutilizadas,
		"el informe lleva lo reutilizado")

	require.NoError(t, os.Remove(filepath.Join(copia, "evals", evalQueSeReutiliza)))

	soloLoMedido := informeDeLaCopia(t, copia, ejecucion.ajustar)

	filas := make([]string, 0, 2)

	for _, leido := range []informeLeido{conReutilizada, soloLoMedido} {
		evaluado := `"commit": "` + commitEvaluado + `"`
		require.Equal(t, 1, strings.Count(leido.escrito, evaluado), "el informe nombra su commit una vez")

		informe := filepath.Join(t.TempDir(), ficheroDelInformeJSON)
		require.NoError(t, os.WriteFile(informe,
			[]byte(strings.Replace(leido.escrito, evaluado, `"commit": "`+commit+`"`, 1)), 0o600))

		salida, deError, codigo := ejecutarLaSerie(t, raiz, "fila", skillDeLasSesiones, procedenciaDeLaSerie(1, commit),
			informe)
		require.Equalf(t, 0, codigo, "fila: %s", deError)

		filas = append(filas, salida)
	}

	assert.Equal(t, filas[1], filas[0], "la fila del informe con la reutilizada es la del de solo lo medido")

	var fila filaDeLaSerie

	require.NoError(t, json.Unmarshal([]byte(filas[0]), &fila))
	assert.True(t, fila.Cuenta)
	require.Len(t, fila.Clases, 1)
	assert.Equal(t, 0, fila.Clases[0].Marcadas)
	assert.Equal(t, 6, fila.Clases[0].Total, "las respuestas medidas del modelo que decide en los dos modos")
}

// marcarUnaRespuestaReutilizadaEnCadaModo da como marcada una respuesta de la
// eval reutilizada en el recuento de cada modo de la clase que decide y, con
// ellas, dos en su conjunto.
func marcarUnaRespuestaReutilizadaEnCadaModo(aportes []aporteDeLaEval) {
	for posicion, aporte := range aportes {
		switch aporte.Umbral {
		case conjuntoDelInforme:
			aportes[posicion].Medida = 2
		case conjuntoDelInforme + ":" + string(ModoOrden), conjuntoDelInforme + ":" + string(ModoHerramienta):
			aportes[posicion].Medida = 1
		}
	}
}
