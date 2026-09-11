package render

import (
	"bytes"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// El presentador de este paquete es la implementación de la interfaz que el
// kernel declara y consume (research.md D2, D15). La comprobación es de
// compilación y está aquí, en el test, para que la dirección de dependencia del
// código de producción siga siendo render → core/schema y el kernel no llegue a
// importar este paquete nunca.
var _ cli.Presentador = (*Presentador)(nil)

// errTuberiaCerrada es el fallo del descriptor que ya no admite nada: el caso
// que el contrato obliga a traducir a código de salida y nunca a un pánico
// (contracts/banderas-y-exit-codes.md §4).
var errTuberiaCerrada = errors.New("la tubería está cerrada")

// escritorRoto es el io.Writer que siempre devuelve error y cuenta cuántas
// veces lo han intentado. La cuenta es lo que permite comprobar la segunda
// mitad de la regla: que tras una escritura fallida nadie vuelve a escribir por
// ese descriptor.
type escritorRoto struct {
	escrituras int
}

func (e *escritorRoto) Write(_ []byte) (int, error) {
	e.escrituras++

	return 0, errTuberiaCerrada
}

// instanteDeEjemplo es la fecha de consulta de los sobres de este test, fija
// para que la forma emitida se pueda comparar carácter a carácter. Lleva
// desplazamiento horario explícito, como exige el contrato del sobre.
func instanteDeEjemplo() time.Time {
	return time.Date(2026, 9, 11, 10, 12, 0, 0, time.FixedZone("CEST", 2*60*60))
}

// hashDeEjemplo es una huella con la forma del contrato: el prefijo del
// algoritmo seguido de los 64 dígitos hexadecimales.
const hashDeEjemplo = "sha256:2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"

// sobreDeEjemplo es el sobre de la entrega literal del hito —`kitlegal echo
// hola`—, el mismo que ilustra el contrato en su sección de presentación
// legible (contracts/sobre-de-salida.md §7).
func sobreDeEjemplo() schema.Sobre {
	return schema.Sobre{
		Ok:            true,
		Fuente:        "kitlegal.echo",
		URL:           "kitlegal:applet/echo",
		FechaConsulta: instanteDeEjemplo(),
		Hash:          hashDeEjemplo,
		Data:          map[string]any{"mensaje": "hola"},
	}
}

// TestPresentador comprueba el reparto entre los dos descriptores: el sobre, en
// cualquiera de sus dos formas, y el texto para personas salen por la salida
// estándar; el aviso, por la de error. Ninguno de los dos escribe en el
// descriptor del otro (FR-040 … FR-042, SC-011).
func TestPresentador(t *testing.T) {
	t.Parallel()

	t.Run("los escritores crudos son los que se le entregaron", func(t *testing.T) {
		t.Parallel()

		var salida, errores bytes.Buffer
		presentador := Nuevo(&salida, &errores)

		assert.Same(t, &salida, presentador.Salida())
		assert.Same(t, &errores, presentador.Error())
	})

	casos := []struct {
		nombre   string
		escribir func(p *Presentador) error
		enSalida string
		enError  string
	}{
		{
			nombre:   "el sobre en la forma legible por máquina va a la salida estándar",
			escribir: func(p *Presentador) error { return p.Presentar(sobreDeEjemplo(), true) },
			enSalida: `"fuente":"kitlegal.echo"`,
		},
		{
			nombre:   "la tabla mínima también va a la salida estándar",
			escribir: func(p *Presentador) error { return p.Presentar(sobreDeEjemplo(), false) },
			enSalida: "fuente          kitlegal.echo\n",
		},
		{
			nombre:   "el texto para la persona va a la salida estándar",
			escribir: func(p *Presentador) error { return p.Texto("kitlegal 1.2.3") },
			enSalida: "kitlegal 1.2.3\n",
		},
		{
			nombre:   "el aviso va a la salida de error",
			escribir: func(p *Presentador) error { return p.Aviso("applets disponibles: (ninguno)") },
			enError:  "applets disponibles: (ninguno)\n",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			var salida, errores bytes.Buffer
			require.NoError(t, caso.escribir(Nuevo(&salida, &errores)))

			if caso.enSalida != "" {
				assert.Contains(t, salida.String(), caso.enSalida)
				assert.Empty(t, errores.String(), "la salida de error no recibe nada del resultado")
			} else {
				assert.Contains(t, errores.String(), caso.enError)
				assert.Empty(t, salida.String(), "un aviso nunca toca la salida estándar")
			}
		})
	}
}

// TestEscrituraFallida es el caso de la tubería cerrada, que ningún guion puede
// simular de forma portable: escribir en un descriptor puede fallar, y eso es
// un error que se propaga —para que el kernel lo traduzca a código de salida—,
// nunca un pánico ni un silencio (FR-040, contracts/banderas-y-exit-codes.md
// §4, quickstart.md escenario 5).
func TestEscrituraFallida(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre   string
		escribir func(p *Presentador) error
	}{
		{
			nombre:   "el sobre en la forma legible por máquina",
			escribir: func(p *Presentador) error { return p.Presentar(sobreDeEjemplo(), true) },
		},
		{
			nombre:   "la tabla mínima, cuyo fallo aparece en el vaciado final",
			escribir: func(p *Presentador) error { return p.Presentar(sobreDeEjemplo(), false) },
		},
		{
			nombre:   "el texto para la persona",
			escribir: func(p *Presentador) error { return p.Texto("kitlegal 1.2.3") },
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			roto := &escritorRoto{}
			var errores bytes.Buffer
			presentador := Nuevo(roto, &errores)

			var err error
			require.NotPanics(t, func() { err = caso.escribir(presentador) },
				"ningún camino de usuario termina en pánico (FR-033)")
			require.ErrorIs(t, err, errTuberiaCerrada, "el fallo de la escritura se propaga")
			assert.Contains(t, err.Error(), "la salida estándar",
				"el mensaje nombra el descriptor que falló")

			intentos := roto.escrituras
			require.Positive(t, intentos, "el presentador llegó a intentar la escritura")

			// Así reacciona el kernel a una escritura fallida: el mensaje para la
			// persona por la salida de error, sin volver a tocar el descriptor
			// roto. El presentador no reintenta por su cuenta ni emite un segundo
			// sobre por el descriptor que acaba de romperse.
			require.NoError(t, presentador.Aviso("la salida estándar falló"))
			assert.Equal(t, intentos, roto.escrituras,
				"no se intenta un segundo sobre por el descriptor roto")
			assert.Contains(t, errores.String(), "la salida estándar falló")
		})
	}

	t.Run("el aviso por la salida de error rota", func(t *testing.T) {
		t.Parallel()

		roto := &escritorRoto{}
		presentador := Nuevo(io.Discard, roto)

		var err error
		require.NotPanics(t, func() { err = presentador.Aviso("applet desconocido") })
		require.ErrorIs(t, err, errTuberiaCerrada)
		assert.Contains(t, err.Error(), "la salida de error")
	})

	t.Run("los dos descriptores rotos dejan solo el código de salida", func(t *testing.T) {
		t.Parallel()

		salida, errores := &escritorRoto{}, &escritorRoto{}
		presentador := Nuevo(salida, errores)

		require.NotPanics(t, func() {
			require.ErrorIs(t, presentador.Presentar(sobreDeEjemplo(), true), errTuberiaCerrada)
			require.ErrorIs(t, presentador.Aviso("la salida estándar falló"), errTuberiaCerrada)
		}, "cuando ya no queda descriptor por el que contarlo, no queda nada que hacer salvo devolver el error")
	})
}
