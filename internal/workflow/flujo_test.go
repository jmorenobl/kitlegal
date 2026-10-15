package workflow

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

const (
	// rutaDelWorkflow es la del workflow `hito`, relativa a la raíz.
	rutaDelWorkflow = ".specify/workflows/hito/workflow.yml"

	// modeloDelFanOut es el de la plantilla de los jueces de la revisión
	// final, que toma modelo, paso y rol de cada item.
	modeloDelFanOut = "{{ item.modelo }}#revision_juez_{{ item.id }}:{{ item.rol }}"
)

var (
	// modeloDeUnPaso es la forma del modelo de un paso con un input propio:
	// el input y, detrás, el sufijo #<paso>:<rol>.
	modeloDeUnPaso = regexp.MustCompile(`^\{\{ inputs\.(modelo_[a-z]+) \}\}#([a-z0-9_]+):(modelo_[a-z]+)$`)

	// argumentoDeUnJuez y itemDeUnJuez son, en el paso que genera los items
	// de los jueces, la variable de jq que recibe el input de cada uno y el
	// item que la usa con su rol.
	argumentoDeUnJuez = regexp.MustCompile(`--arg (m[a-z]) "\{\{ inputs\.(modelo_[a-z]+) \}\}"`)
	itemDeUnJuez      = regexp.MustCompile(`modelo:\$(m[a-z]), rol:"(modelo_[a-z]+)"`)
)

// pasoDelWorkflow es lo que el test lee de cada paso de workflow.yml.
type pasoDelWorkflow struct {
	ID    string            `yaml:"id"`
	Tipo  string            `yaml:"type"`
	Model string            `yaml:"model"`
	Run   string            `yaml:"run"`
	Items string            `yaml:"items"`
	Then  []pasoDelWorkflow `yaml:"then"`
	Else  []pasoDelWorkflow `yaml:"else"`
	Steps []pasoDelWorkflow `yaml:"steps"`
	Step  *pasoDelWorkflow  `yaml:"step"`
}

// TestCadaPasoConModeloDiceSuPasoYSuRol: spec-kit no le dice al ejecutable de
// Claude qué paso es, y el registro del modelo de cada sesión
// (gates/modelos.jsonl) lo toma del sufijo #<paso>:<rol> del modelo que el
// paso le pasa. Cada paso con modelo lleva su propio id y el input de su
// plantilla como rol; los jueces de la revisión final, el id y el rol de su
// item, que son los del input que recibe cada uno.
func TestCadaPasoConModeloDiceSuPasoYSuRol(t *testing.T) {
	t.Parallel()

	var flujo struct {
		Steps []pasoDelWorkflow `yaml:"steps"`
	}

	require.NoError(t, yaml.Unmarshal([]byte(leerDelRepositorio(t, rutaDelWorkflow)), &flujo))

	pasos := map[string]pasoDelWorkflow{}
	conModelo := 0

	var recorrer func([]pasoDelWorkflow)

	recorrer = func(lista []pasoDelWorkflow) {
		for _, paso := range lista {
			pasos[paso.ID] = paso

			if paso.Model != "" {
				conModelo++

				comprobarElModeloDelPaso(t, paso)
			}

			recorrer(paso.Then)
			recorrer(paso.Else)
			recorrer(paso.Steps)

			if paso.Step != nil {
				recorrer([]pasoDelWorkflow{*paso.Step})
			}
		}
	}

	recorrer(flujo.Steps)
	assert.Greater(t, conModelo, 20, "el workflow tiene más de veinte pasos con modelo")

	generador, esta := pasos["preparar_jueces"]
	require.True(t, esta, "el paso que genera los items de los jueces existe")

	inputs := map[string]string{}
	for _, m := range argumentoDeUnJuez.FindAllStringSubmatch(generador.Run, -1) {
		inputs[m[1]] = m[2]
	}

	items := itemDeUnJuez.FindAllStringSubmatch(generador.Run, -1)
	require.Len(t, items, 2, "dos jueces, cada uno con su rol")

	for _, m := range items {
		assert.Equal(t, inputs[m[1]], m[2], "el rol del juez de $%s es el input que recibe", m[1])
	}
}

func comprobarElModeloDelPaso(t *testing.T, paso pasoDelWorkflow) {
	t.Helper()

	if paso.ID == "revision_juez" {
		assert.Equal(t, modeloDelFanOut, paso.Model, "los jueces de la revisión final toman paso y rol de su item")

		return
	}

	m := modeloDeUnPaso.FindStringSubmatch(paso.Model)
	if !assert.NotNil(t, m, "el modelo del paso %s es su input con #<paso>:<rol>: %q", paso.ID, paso.Model) {
		return
	}

	assert.Equal(t, paso.ID, m[2], "el sufijo del paso %s nombra su propio id", paso.ID)
	assert.Equal(t, m[1], m[3], "el rol del sufijo del paso %s es el input de su modelo", paso.ID)
}
