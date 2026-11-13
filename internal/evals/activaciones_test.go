package evals

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Los usos de los casos de TestActivacionesDeLosUsos, con las skills
// skillNormativa y skillDeConsultaJuridica: la consulta del modo orden es la de
// evals/normativa/04-lgt-prescripcion.yaml (contracts/informe.md §1), y la del
// modo herramienta, la misma herramienta del registro con el prefijo del
// servidor o sin él.
const (
	ordenQueConsultaLaLGT    = "kitlegal boe articulo BOE-A-2003-23186 a66 --json"
	ordenQueBuscaElBinario   = "which kitlegal"
	ordenQueLeeUnaReferencia = "cat /proyecto/.claude/skills/normativa/references/normas.md"
	herramientaConElPrefijo  = "mcp__kitlegal__boe_articulo"
	herramientaSinElPrefijo  = "boe_articulo"
	herramientaToolSearch    = "ToolSearch"
)

// usoDeSkill es el uso de Skill que activa esa skill en ese turno.
func usoDeSkill(t *testing.T, turno int, skill string) usoDelTranscript {
	t.Helper()

	return usoDelTranscript{turno: turno, nombre: herramientaSkill, entrada: entradaDe(t, map[string]string{"skill": skill})}
}

// usoDeBash es el uso de Bash con esa orden en ese turno.
func usoDeBash(t *testing.T, turno int, orden string) usoDelTranscript {
	t.Helper()

	return usoDelTranscript{turno: turno, nombre: herramientaBash, entrada: entradaDe(t, map[string]string{"command": orden})}
}

// usoDe es el uso de esa herramienta, con una entrada vacía, en ese turno.
func usoDe(turno int, nombre string) usoDelTranscript {
	return usoDelTranscript{turno: turno, nombre: nombre, entrada: json.RawMessage(`{}`)}
}

// entradaDe es el input de un bloque tool_use con esos campos de texto.
func entradaDe(t *testing.T, campos map[string]string) json.RawMessage {
	t.Helper()

	entrada, err := json.Marshal(campos)
	require.NoError(t, err)

	return entrada
}

// casoDeActivaciones son los usos de una sesión de un modo y la lista de
// activaciones que dan, o el error que la impide.
type casoDeActivaciones struct {
	nombre    string
	modo      Modo
	usos      []usoDelTranscript
	esperadas []Activacion
	error     string
}

// TestActivacionesDeLosUsos fija la lista de las skills activadas de una sesión
// (FR-001; SC-001; contracts/informe.md §1; data-model §1; research D1 y D2):
// por cada uso de Skill, en orden y también si repite skill, su input.skill y si
// va en un turno anterior al de la primera herramienta de kitlegal del modo
// —la primera también cuando hay varias—, o en cualquiera si la sesión no tiene
// ninguna; en el mismo turno, no. En el modo
// herramienta lo es una del registro, con el prefijo del servidor o sin él, y no
// ToolSearch; en el modo orden, una orden de Bash que consulta con kitlegal, y no
// which kitlegal; y en el de una eval sin binario ni servidor, ninguna, tampoco
// esa orden. Sin ninguna activación, la lista es [] y no nil.
func TestActivacionesDeLosUsos(t *testing.T) {
	t.Parallel()

	casos := []casoDeActivaciones{
		{
			nombre:    "antes de la primera herramienta",
			modo:      ModoHerramienta,
			usos:      []usoDelTranscript{usoDeSkill(t, 1, skillNormativa), usoDe(2, herramientaConElPrefijo)},
			esperadas: []Activacion{{Skill: skillNormativa, Antes: true}},
		},
		{
			nombre: "en el mismo turno que la primera herramienta y en uno posterior",
			modo:   ModoHerramienta,
			usos: []usoDelTranscript{
				usoDeSkill(t, 1, skillNormativa), usoDe(1, herramientaSinElPrefijo),
				usoDeSkill(t, 2, skillDeConsultaJuridica),
			},
			esperadas: []Activacion{{Skill: skillNormativa}, {Skill: skillDeConsultaJuridica}},
		},
		{
			// La primera herramienta es la del turno 1, no la del 3.
			nombre: "entre dos herramientas de kitlegal",
			modo:   ModoHerramienta,
			usos: []usoDelTranscript{
				usoDe(1, herramientaConElPrefijo), usoDeSkill(t, 2, skillNormativa),
				usoDe(3, herramientaSinElPrefijo),
			},
			esperadas: []Activacion{{Skill: skillNormativa}},
		},
		{
			// La primera orden es la del turno 2, no la del 4.
			nombre: "cada skill antes de una orden kitlegal",
			modo:   ModoOrden,
			usos: []usoDelTranscript{
				usoDeSkill(t, 1, skillDeConsultaJuridica), usoDeBash(t, 2, ordenQueConsultaLaLGT),
				usoDeSkill(t, 3, skillNormativa), usoDeBash(t, 4, ordenQueConsultaLaLGT),
			},
			esperadas: []Activacion{{Skill: skillDeConsultaJuridica, Antes: true}, {Skill: skillNormativa}},
		},
		{
			nombre:    "sin ninguna herramienta de kitlegal",
			modo:      ModoOrden,
			usos:      []usoDelTranscript{usoDeSkill(t, 1, skillNormativa), usoDe(2, herramientaRead)},
			esperadas: []Activacion{{Skill: skillNormativa, Antes: true}},
		},
		{
			nombre: "la misma skill activada dos veces",
			modo:   ModoOrden,
			usos: []usoDelTranscript{
				usoDeSkill(t, 1, skillNormativa), usoDeBash(t, 2, ordenQueConsultaLaLGT),
				usoDeSkill(t, 3, skillNormativa),
			},
			esperadas: []Activacion{{Skill: skillNormativa, Antes: true}, {Skill: skillNormativa}},
		},
		{
			nombre:    "una orden kitlegal en una eval sin binario ni servidor",
			modo:      "",
			usos:      []usoDelTranscript{usoDeBash(t, 1, ordenQueConsultaLaLGT), usoDeSkill(t, 2, skillNormativa)},
			esperadas: []Activacion{{Skill: skillNormativa, Antes: true}},
		},
		{
			nombre:    "una herramienta del registro en una eval sin binario ni servidor",
			modo:      "",
			usos:      []usoDelTranscript{usoDe(1, herramientaConElPrefijo), usoDeSkill(t, 2, skillNormativa)},
			esperadas: []Activacion{{Skill: skillNormativa, Antes: true}},
		},
		{
			nombre: "which kitlegal no es la primera herramienta",
			modo:   ModoOrden,
			usos: []usoDelTranscript{
				usoDeBash(t, 1, ordenQueBuscaElBinario), usoDeSkill(t, 2, skillNormativa),
				usoDeBash(t, 3, ordenQueConsultaLaLGT),
			},
			esperadas: []Activacion{{Skill: skillNormativa, Antes: true}},
		},
		{
			nombre: "ToolSearch no es la primera herramienta",
			modo:   ModoHerramienta,
			usos: []usoDelTranscript{
				{
					turno: 1, nombre: herramientaToolSearch,
					entrada: entradaDe(t, map[string]string{"query": "select:" + herramientaConElPrefijo}),
				},
				usoDeSkill(t, 2, skillNormativa), usoDe(3, herramientaConElPrefijo),
			},
			esperadas: []Activacion{{Skill: skillNormativa, Antes: true}},
		},
		{
			nombre: "una orden kitlegal en el modo herramienta no es la primera herramienta",
			modo:   ModoHerramienta,
			usos: []usoDelTranscript{
				usoDeBash(t, 1, ordenQueConsultaLaLGT), usoDeSkill(t, 2, skillNormativa),
				usoDe(3, herramientaConElPrefijo),
			},
			esperadas: []Activacion{{Skill: skillNormativa, Antes: true}},
		},
		{
			nombre:    "ninguna activación",
			modo:      ModoHerramienta,
			usos:      []usoDelTranscript{usoDeBash(t, 1, ordenQueLeeUnaReferencia), usoDe(2, herramientaConElPrefijo)},
			esperadas: []Activacion{},
		},
		{
			nombre:    "ningún uso",
			modo:      ModoOrden,
			esperadas: []Activacion{},
		},
		{
			nombre: "un uso de Skill sin su entrada",
			modo:   ModoOrden,
			usos: []usoDelTranscript{
				{turno: 1, nombre: herramientaSkill, entrada: json.RawMessage(`"normativa"`)},
			},
			error: "el bloque tool_use de Skill del turno 1 no tiene la entrada {skill, args?}",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			// Un reconocedor por caso: el analizador de su intérprete guarda lo que
			// lee de cada invocación, y los casos corren a la vez.
			reconocedor, err := nuevoReconocedorDeKitlegal()
			require.NoError(t, err)

			activaciones, err := activacionesDeLosUsos(caso.usos, func(uso usoDelTranscript) bool {
				return reconocedor.esDeKitlegal(uso, caso.modo)
			})

			if caso.error != "" {
				require.ErrorContains(t, err, caso.error)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, caso.esperadas, activaciones)
			assert.NotNil(t, activaciones, "la lista sin ninguna activación es [], no nil")
		})
	}
}

// TestTextoDeLasActivaciones fija el texto de la lista de activaciones de una
// sesión en informe.md (FR-002; contracts/informe.md §3; research D4): cada
// skill con «(antes)» o «(después)», en el orden de la lista y separadas por
// «, »; y «ninguna» con la lista vacía.
func TestTextoDeLasActivaciones(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre       string
		activaciones []Activacion
		texto        string
	}{
		{nombre: "ninguna", activaciones: []Activacion{}, texto: "ninguna"},
		{nombre: "una-antes", activaciones: []Activacion{{Skill: skillNormativa, Antes: true}}, texto: "normativa (antes)"},
		{
			nombre:       "una-despues",
			activaciones: []Activacion{{Skill: skillNormativa}},
			texto:        "normativa (después)",
		},
		{
			nombre:       "dos-en-su-orden",
			activaciones: []Activacion{{Skill: skillNormativa, Antes: true}, {Skill: skillDeConsultaJuridica}},
			texto:        "normativa (antes), consulta-juridica (después)",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.texto, textoDeLasActivaciones(caso.activaciones))
		})
	}
}

// TestLaActivacionEnJSON fija las claves de una activación en JSON, skill y
// antes (data-model §1), y que la lista sin ninguna se escribe [].
func TestLaActivacionEnJSON(t *testing.T) {
	t.Parallel()

	activaciones, err := activacionesDeLosUsos(nil, func(usoDelTranscript) bool { return false })
	require.NoError(t, err)

	vacia, err := json.Marshal(activaciones)
	require.NoError(t, err)
	assert.JSONEq(t, `[]`, string(vacia))

	una, err := json.Marshal([]Activacion{{Skill: skillNormativa, Antes: true}})
	require.NoError(t, err)
	assert.JSONEq(t, `[{"skill": "normativa", "antes": true}]`, string(una))
}
