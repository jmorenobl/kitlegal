package evals

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Lo que se escribe en los ficheros de los casos de este fichero: un mensaje
// assistant que activa normativa con la entrada que no es {skill, args?}, uno
// cuyo message.content no es una lista y un user cuyo contenido es un texto.
const (
	assistantConLaSkillSinEntrada = `{"type":"assistant","message":{"id":"m1","content":[` +
		`{"type":"tool_use","name":"Skill","input":"normativa"}]}}` + "\n"
	assistantSinLista = `{"type":"assistant","message":{"id":"m1","content":"texto"}}` + "\n"
	userConTexto      = `{"type":"user","message":{"content":"una pregunta"}}` + "\n"
	assistantConTexto = `{"type":"assistant","message":{"id":"m1","content":[{"type":"text","text":"hola"}]}}` + "\n"
)

// TestLosUsosDeUnTranscriptQueNoSeLee: leerLosUsos y activacionesDeLaSesion
// devuelven un error que empieza por el transcript cuando falta, cuando una
// línea no es JSON, cuando un mensaje assistant no trae la lista
// message.content y cuando un uso de Skill no tiene la entrada {skill, args?};
// un mensaje user con un texto por contenido no es un defecto (FR-001;
// data-model §2).
func TestLosUsosDeUnTranscriptQueNoSeLee(t *testing.T) {
	t.Parallel()

	reconocedor, err := nuevoReconocedorDeKitlegal()
	require.NoError(t, err)

	casos := []struct {
		nombre     string
		transcript *string
		contiene   string
	}{
		{nombre: "sin transcript", contiene: ficheroDelTranscript + ": "},
		{nombre: "una línea que no es JSON", transcript: new("{\n"), contiene: "línea 1: no es un mensaje JSON"},
		{
			nombre: "un assistant sin la lista message.content", transcript: new(assistantSinLista),
			contiene: "línea 1: el mensaje assistant no tiene la lista message.content",
		},
		{
			nombre: "un uso de Skill sin la entrada {skill, args?}", transcript: new(assistantConLaSkillSinEntrada),
			contiene: ficheroDelTranscript + ": el bloque tool_use de Skill del turno 1 no tiene la entrada",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			directorio := t.TempDir()
			if caso.transcript != nil {
				escribirEnLaSesion(t, directorio, ficheroDelTranscript, *caso.transcript)
			}

			activaciones, err := activacionesDeLaSesion(directorio, reconocedor, ModoHerramienta)
			require.Error(t, err)
			assert.Contains(t, err.Error(), caso.contiene)
			assert.Nil(t, activaciones)
		})
	}

	t.Run("un user con un texto por contenido", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		escribirEnLaSesion(t, directorio, ficheroDelTranscript, userConTexto+assistantConTexto)

		usos, err := leerLosUsos(directorio)
		require.NoError(t, err)
		assert.Empty(t, usos, "un texto no trae ningún tool_result ni es un defecto")
	})
}

// TestLasSkillsDelSondeoQueNoSePuedenCopiar: con INSTALACION=producto, el arnés
// copia la skill de la variante y las demás de skills/ del árbol (FR-010), y
// anota las que deja instaladas (FR-012). Si no puede leer o copiar alguna, o
// leer lo que dejó, el error nombra la ruta, y no se abre ninguna sesión con
// una instalación a medias.
func TestLasSkillsDelSondeoQueNoSePuedenCopiar(t *testing.T) {
	t.Parallel()

	normativa := familiasDelProducto[0]
	require.Equal(t, skillDeNormativa, normativa.skill)

	t.Run("sin la skill de la familia en el árbol", func(t *testing.T) {
		t.Parallel()

		sondeo := SondeoDelLaboratorio{Raiz: t.TempDir()}

		_, err := sondeo.copiarLasSkillsDeLaVariante(t.TempDir(), normativa, normativa.producto, true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no se pueden copiar")
	})

	t.Run("sin skills/ en el árbol", func(t *testing.T) {
		t.Parallel()

		sondeo := SondeoDelLaboratorio{Raiz: t.TempDir()}

		err := sondeo.copiarLasDemasSkillsDelArbol(t.TempDir(), normativa.skill, normativa.producto)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "las skills del árbol")
	})

	t.Run("otra skill que choca con lo que ya hay", func(t *testing.T) {
		t.Parallel()

		raiz := t.TempDir()
		escribirEnLaSesion(t, filepath.Join(raiz, directorioDeLasSkills, normativa.skill, "references"),
			"normas.md", "normas\n")
		escribirEnLaSesion(t, filepath.Join(raiz, directorioDeLasSkills, "otra"), "SKILL.md", "otra\n")

		temporal := t.TempDir()
		escribirEnLaSesion(t, filepath.Join(temporal, carpetaDeLasSkillsDelSondeo, normativa.producto, "otra"),
			"SKILL.md", "ya estaba\n")

		_, err := SondeoDelLaboratorio{Raiz: raiz}.copiarLasSkillsDeLaVariante(temporal, normativa,
			normativa.producto, true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "las skills de la variante "+normativa.producto+" no se pueden copiar")
	})

	t.Run("sin las skills instaladas en el temporal", func(t *testing.T) {
		t.Parallel()

		pares := []parDelLaboratorio{{variante: normativa.variantes[1], modo: ModoHerramienta}}

		instaladas, err := skillsInstaladasDelSondeo(t.TempDir(), normativa, pares)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "las skills instaladas")
		assert.Nil(t, instaladas)
	})
}

// TestLaTablaConUnSondeoQueNoSeLee: la tabla mira en el sondeo.json de cada
// sondeo si tiene la constancia de las skills instaladas (FR-022); el de un
// sondeo sin la variante del producto no lo lee nada más, y si no es JSON la
// tabla no sale.
func TestLaTablaConUnSondeoQueNoSeLee(t *testing.T) {
	t.Parallel()

	normativa := familiasDelProducto[0]

	conElProducto := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(conElProducto, normativa.producto), 0o750))
	escribirEnLaSesion(t, conElProducto, ficheroDelSondeo,
		`{"modelo":"claude-sonnet-5-5","claude_code":"2.1.284","referencias_del_producto":["normas.md"]}`)

	sinElProducto := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(sinElProducto, varianteDeControlDeNormativa), 0o750))
	escribirEnLaSesion(t, sinElProducto, ficheroDelSondeo, "{")

	_, err := familiaDeLosSondeos(t.TempDir(), []string{conElProducto, sinElProducto})
	require.Error(t, err)
	assert.Contains(t, err.Error(), filepath.Join(sinElProducto, ficheroDelSondeo)+" no es un objeto")
}

// escribirEnLaSesion escribe un fichero en un directorio de un caso, que crea si
// no existe.
func escribirEnLaSesion(t *testing.T, directorio, nombre, contenido string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(directorio, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(directorio, nombre), []byte(contenido), 0o600))
}
