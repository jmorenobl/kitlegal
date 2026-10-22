package workflow

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

const (
	// rutaDelRegistro es el registro de los workflows instalados de spec-kit,
	// lo que enseña `specify workflow list`.
	rutaDelRegistro = ".specify/workflows/workflow-registry.json"

	// fuenteDelWorkflow es desde donde lo instala make workflow-registro: su
	// propio directorio, relativo a la raíz.
	fuenteDelWorkflow = ".specify/workflows/hito"

	// comoSeArregla es lo que dice cada fallo de TestElRegistroDeSpecKitEsElDelWorkflow.
	comoSeArregla = "make workflow-registro lo vuelve a instalar desde su definición"
)

// cabeceraDelWorkflow es la cabecera de workflow.yml, lo que spec-kit copia al
// registro al instalarlo.
type cabeceraDelWorkflow struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
}

// workflowInstalado es la entrada de un workflow en el registro de spec-kit.
type workflowInstalado struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Source      string `json:"source"`
}

// TestElRegistroDeSpecKitEsElDelWorkflow: el registro de spec-kit dice del
// workflow hito lo que dice la cabecera de workflow.yml —nombre, versión y
// descripción—, y que su fuente es su propio directorio. Hasta el
// 2026-10-22 decía 1.4.1, con la descripción de septiembre y una fuente en un
// directorio temporal que ya no existía, porque nada lo volvía a instalar al
// cambiar la versión. Lanzar un run no lo necesita —`specify workflow run`
// lee workflow.yml—, pero `specify workflow list` lo enseña como verdad.
func TestElRegistroDeSpecKitEsElDelWorkflow(t *testing.T) {
	t.Parallel()

	var delFichero struct {
		Workflow cabeceraDelWorkflow `yaml:"workflow"`
	}

	require.NoError(t, yaml.Unmarshal([]byte(leerDelRepositorio(t, rutaDelWorkflow)), &delFichero))
	require.NotEmpty(t, delFichero.Workflow.ID, "workflow.yml tiene su cabecera")

	var registro struct {
		Workflows map[string]workflowInstalado `json:"workflows"`
	}

	require.NoError(t, json.Unmarshal([]byte(leerDelRepositorio(t, rutaDelRegistro)), &registro))

	instalado, ok := registro.Workflows[delFichero.Workflow.ID]
	require.Truef(t, ok, "el registro no tiene el workflow %s: %s", delFichero.Workflow.ID, comoSeArregla)

	assert.Equalf(t, delFichero.Workflow.Version, instalado.Version, "la versión del registro: %s", comoSeArregla)
	assert.Equalf(t, delFichero.Workflow.Name, instalado.Name, "el nombre del registro: %s", comoSeArregla)
	assert.Equalf(t, delFichero.Workflow.Description, instalado.Description, "la descripción del registro: %s", comoSeArregla)
	assert.Equalf(t, fuenteDelWorkflow, instalado.Source, "la fuente del registro: %s", comoSeArregla)
}
