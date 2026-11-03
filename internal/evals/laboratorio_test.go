package evals

import (
	"encoding/json"
	"io/fs"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/mcp"
)

// respuestasDeLaReglaTestigo es el directorio de las respuestas fijas con las
// que TestReglaTestigo comprueba el guion de la regla testigo, cada una con su
// veredicto en el nombre (contracts/laboratorio.md §3 de H39).
const respuestasDeLaReglaTestigo = "testdata/laboratorio/respuestas"

// Los principios del nombre de una respuesta fija, que dicen su veredicto: la
// que empieza por el de no cumplir no cumple la regla y la que empieza por el de
// cumplir la cumple.
const (
	principioDeLaQueCumple   = "cumple"
	principioDeLaQueNoCumple = "no-cumple"
)

// nombresDeLasRespuestasFijas son las respuestas fijas de la regla testigo: una
// que la cumple con el sello al final, otra con líneas vacías detrás, y una por
// cada forma de no cumplirla que la regla distingue —el sello no aparece,
// aparece y no es la última línea, es la última con formato y es la última con
// otro texto— (FR-014 de H39; data-model §5 de H39).
var nombresDeLasRespuestasFijas = []string{
	"cumple.txt",
	"cumple-con-lineas-en-blanco.txt",
	"no-cumple-sin-sello.txt",
	"no-cumple-sello-no-al-final.txt",
	"no-cumple-sello-en-negrita.txt",
	"no-cumple-sello-con-texto.txt",
}

// La respuesta de una skill de kitlegal a una pregunta de las evals de
// laboratorio: la de la primera sesión de Sonnet en el modo orden de la eval
// del artículo 140 de la Constitución, la 01 del laboratorio, en el informe de
// evals de normativa que versionó H32 (research V20 de H39). El informe está
// versionado, solo se lee y no se edita.
const (
	informeDeNormativaDeH32 = "../../specs/030-h32-arquitectura-de-skills/gates/evals/normativa.json"
	evalDeLaSesionDeH32     = "09-constitucion-articulo-140.yaml"
	modeloDeLaSesionDeH32   = "claude-sonnet-5-5"
	numeroDeLaSesionDeH32   = 1
)

// skillsDeKitlegal es el directorio de las skills de kitlegal, desde el del
// paquete: las que empotra el binario, con su SKILL.md y sus referencias.
const skillsDeKitlegal = "../../skills"

// TestReglaTestigo comprueba el guion de la regla testigo (FR-013, FR-014 y
// SC-003 de H39; contracts/laboratorio.md §3 de H39): acierta el veredicto de
// cada respuesta fija y el de la respuesta de una skill de kitlegal, que no la
// cumple, y la línea del sello es inventada para el laboratorio, así que no
// aparece en ninguna skill ni referencia de kitlegal ni en las instrucciones del
// servidor MCP.
func TestReglaTestigo(t *testing.T) {
	t.Parallel()

	t.Run("respuestas fijas", comprobarLasRespuestasFijas)
	t.Run("el párrafo de la regla", comprobarElParrafoDeLaRegla)
	t.Run("respuesta de una skill de kitlegal", comprobarLaRespuestaDeKitlegal)
	t.Run("sello fuera de kitlegal", comprobarElSelloFueraDeKitlegal)
}

// comprobarLasRespuestasFijas: el directorio tiene exactamente las respuestas
// fijas, y el guion da a cada una el veredicto de su nombre.
func comprobarLasRespuestasFijas(t *testing.T) {
	t.Parallel()

	entradas, err := os.ReadDir(respuestasDeLaReglaTestigo)
	require.NoError(t, err, "el directorio %s de las respuestas fijas se lee", respuestasDeLaReglaTestigo)

	nombres := make([]string, 0, len(entradas))
	for _, entrada := range entradas {
		nombres = append(nombres, entrada.Name())
	}

	require.ElementsMatch(t, nombresDeLasRespuestasFijas, nombres,
		"%s tiene las respuestas fijas de la regla testigo y ninguna más", respuestasDeLaReglaTestigo)

	for _, nombre := range nombresDeLasRespuestasFijas {
		cumple := veredictoDelNombre(t, nombre)
		respuesta := string(contenidoDelFichero(t, path.Join(respuestasDeLaReglaTestigo, nombre)))

		require.NotEmpty(t, strings.TrimSpace(respuesta), "la respuesta fija %s tiene texto", nombre)
		assert.Equal(t, cumple, reglaTestigo(respuesta), "el guion da a %s el veredicto de su nombre", nombre)
	}
}

// veredictoDelNombre es el veredicto que dice el nombre de una respuesta fija:
// cumple si empieza por el principio de la que cumple y no cumple si empieza
// por el de la que no. Un nombre sin ninguno de los dos no dice veredicto.
func veredictoDelNombre(t *testing.T, nombre string) bool {
	t.Helper()

	switch {
	case strings.HasPrefix(nombre, principioDeLaQueNoCumple+"-"):
		return false
	case strings.HasPrefix(nombre, principioDeLaQueCumple):
		return true
	default:
		require.Failf(t, "respuesta fija sin veredicto",
			"el nombre %s no empieza por %q ni por %q", nombre, principioDeLaQueCumple, principioDeLaQueNoCumple)

		return false
	}
}

// comprobarElParrafoDeLaRegla: el párrafo de la regla termina con la línea del
// sello que pide, y el párrafo mismo, copiado como última línea, no la cumple.
func comprobarElParrafoDeLaRegla(t *testing.T) {
	t.Parallel()

	assert.True(t, strings.HasSuffix(parrafoDeLaRegla, ": "+lineaDelSello),
		"el párrafo de la regla dice exactamente la línea del sello: %q", parrafoDeLaRegla)
	assert.False(t, reglaTestigo(parrafoDeLaRegla+"\n"),
		"la regla copiada como última línea lleva el sello con otro texto y no la cumple")
}

// comprobarLaRespuestaDeKitlegal: la respuesta de la sesión de H32, leída del
// informe versionado, es la de una sesión de la eval del artículo 140 que activó
// la skill normativa, y no cumple la regla.
func comprobarLaRespuestaDeKitlegal(t *testing.T) {
	t.Parallel()

	var informe struct {
		Evals []entradaDelInforme `json:"evals"`
	}
	require.NoError(t, json.Unmarshal(contenidoDelFichero(t, informeDeNormativaDeH32), &informe),
		"%s es un informe del job de evals", informeDeNormativaDeH32)

	nombre := nombreDeSesion(evalDeLaSesionDeH32, modeloDeLaSesionDeH32, ModoOrden, numeroDeLaSesionDeH32, false)

	var sesiones []entradaDelInforme
	for _, entrada := range informe.Evals {
		if entrada.Sesion == nombre {
			sesiones = append(sesiones, entrada)
		}
	}

	require.Len(t, sesiones, 1, "el informe %s tiene una sesión %s", informeDeNormativaDeH32, nombre)
	sesion := sesiones[0]
	require.Equal(t, evalDeLaSesionDeH32, sesion.Eval, "la sesión %s es de la eval del artículo 140", nombre)
	require.True(t, sesion.Activada, "la sesión %s activó la skill normativa", nombre)
	require.NotEmpty(t, strings.TrimSpace(sesion.Respuesta), "la sesión %s tiene respuesta", nombre)

	assert.False(t, reglaTestigo(sesion.Respuesta),
		"la respuesta de la skill normativa en la sesión %s no cumple la regla testigo", nombre)
}

// comprobarElSelloFueraDeKitlegal: la línea del sello no aparece en ningún
// fichero de las skills de kitlegal —ni en un SKILL.md ni en una referencia— ni
// en las instrucciones que da el servidor MCP. Lee al menos un SKILL.md y una
// referencia, para no pasar en vacío.
func comprobarElSelloFueraDeKitlegal(t *testing.T) {
	t.Parallel()

	skills := os.DirFS(skillsDeKitlegal)
	skillsLeidas, referenciasLeidas := 0, 0

	err := fs.WalkDir(skills, ".", func(ruta string, entrada fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entrada.IsDir() {
			return nil
		}

		contenido, err := fs.ReadFile(skills, ruta)
		if err != nil {
			return err
		}

		assert.NotContains(t, string(contenido), lineaDelSello,
			"el fichero %s de las skills de kitlegal no lleva la línea del sello", path.Join(skillsDeKitlegal, ruta))

		switch {
		case path.Base(ruta) == "SKILL.md":
			skillsLeidas++
		case path.Base(path.Dir(ruta)) == "references":
			referenciasLeidas++
		}

		return nil
	})
	require.NoError(t, err, "las skills de kitlegal en %s se leen", skillsDeKitlegal)

	assert.Positive(t, skillsLeidas, "se lee al menos un SKILL.md de %s", skillsDeKitlegal)
	assert.Positive(t, referenciasLeidas, "se lee al menos una referencia de %s", skillsDeKitlegal)
	assert.NotContains(t, mcp.Instrucciones, lineaDelSello,
		"las instrucciones del servidor MCP no llevan la línea del sello")
}
