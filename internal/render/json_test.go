package render

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// sobreEmitido es el documento que produce el sobre de ejemplo: las seis claves
// con sus valores, la fecha en RFC 3339 con desplazamiento horario explícito y
// nada más. Lo que el contrato exige de su forma —una sola línea, un único
// salto al final y sin escapar los caracteres HTML— se comprueba aparte, porque
// una comparación entre documentos JSON no lo vería (FR-042, research.md D15).
const sobreEmitido = `{"ok":true,"fuente":"kitlegal.echo","url":"kitlegal:applet/echo",` +
	`"fecha_consulta":"2026-09-11T10:12:00+02:00","hash":"` + hashDeEjemplo + `",` +
	`"data":{"mensaje":"hola"}}`

// TestJSON comprueba la forma legible por máquina: un documento y nada más en
// la salida estándar (FR-041, FR-042).
func TestJSON(t *testing.T) {
	t.Parallel()

	t.Run("un único documento, sin sangrado y con salto de línea final", func(t *testing.T) {
		t.Parallel()

		var salida bytes.Buffer
		require.NoError(t, escribirJSON(&salida, sobreDeEjemplo()))

		emitido := salida.String()
		assert.JSONEq(t, sobreEmitido, emitido)
		assert.True(t, strings.HasSuffix(emitido, "\n"), "el documento cierra su línea")
		assert.Equal(t, 1, strings.Count(emitido, "\n"),
			"el documento va en una sola línea y termina en un único salto")
	})

	t.Run("y nada más: lo que sigue al documento es el fin de la entrada", func(t *testing.T) {
		t.Parallel()

		var salida bytes.Buffer
		require.NoError(t, escribirJSON(&salida, sobreDeEjemplo()))

		decodificador := json.NewDecoder(&salida)

		var documento map[string]any
		require.NoError(t, decodificador.Decode(&documento))
		assert.Len(t, documento, 6, "las seis claves del sobre, ni una más ni una menos")

		_, err := decodificador.Token()
		assert.ErrorIs(t, err, io.EOF, "tras el documento no queda nada en la salida estándar")
	})

	t.Run("no escapa los caracteres HTML", func(t *testing.T) {
		t.Parallel()

		sobre := sobreDeEjemplo()
		sobre.URL = "https://www.boe.es/buscar/act.php?id=BOE-A-2015-10565&b=1"
		sobre.Data = map[string]any{"titulo": "<abogados & procuradores>"}

		var salida bytes.Buffer
		require.NoError(t, escribirJSON(&salida, sobre))

		emitido := salida.String()
		assert.Contains(t, emitido, "?id=BOE-A-2015-10565&b=1")
		assert.Contains(t, emitido, "<abogados & procuradores>")
		// Con el escape activado —que es el de json.Marshal por omisión— estos
		// tres caracteres saldrían como secuencias \uXXXX y la url del sobre
		// dejaría de poder copiarse tal cual a un navegador.
		for _, secuencia := range []string{"\\u0026", "\\u003c", "\\u003e"} {
			assert.NotContains(t, emitido, secuencia,
				"el sobre se serializa sin escapar los caracteres HTML")
		}
	})

	t.Run("el sobre de fallo se serializa igual que el de éxito", func(t *testing.T) {
		t.Parallel()

		sobre := sobreDeEjemplo()
		sobre.Ok = false
		sobre.Fuente, sobre.URL = "kitlegal.cli", "kitlegal:cli"
		sobre.Data = schema.DatosError{
			Clase:   schema.ClaseArgumentos,
			Mensaje: "bandera desconocida: --ruidoso",
		}

		var salida bytes.Buffer
		require.NoError(t, escribirJSON(&salida, sobre))

		emitido := salida.String()
		assert.JSONEq(t,
			`{"ok":false,"fuente":"kitlegal.cli","url":"kitlegal:cli",`+
				`"fecha_consulta":"2026-09-11T10:12:00+02:00","hash":"`+hashDeEjemplo+`",`+
				`"data":{"clase":"argumentos","mensaje":"bandera desconocida: --ruidoso"}}`,
			emitido)
		assert.Equal(t, 1, strings.Count(emitido, "\n"),
			"el sobre de fallo se emite igual que el de éxito: un documento y un salto")
	})

	t.Run("el fallo de escritura se propaga", func(t *testing.T) {
		t.Parallel()

		roto := &escritorRoto{}
		require.ErrorIs(t, escribirJSON(roto, sobreDeEjemplo()), errTuberiaCerrada)
		assert.Equal(t, 1, roto.escrituras, "el documento se escribe de una vez")
	})

	t.Run("un contenido que no se puede serializar es un error, no un pánico", func(t *testing.T) {
		t.Parallel()

		sobre := sobreDeEjemplo()
		sobre.Data = make(chan int)

		var salida bytes.Buffer
		var err error
		require.NotPanics(t, func() { err = escribirJSON(&salida, sobre) })
		require.Error(t, err)
		assert.Empty(t, salida.String(),
			"un sobre que no se puede serializar no deja medio documento en la salida estándar")
	})
}
