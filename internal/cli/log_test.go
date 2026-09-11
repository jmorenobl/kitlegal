package cli

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// errEscrituraRota es el fallo de un descriptor cerrado: el único motivo por el
// que montar el registrador puede devolver error, porque avisar de un
// KITLEGAL_LOG inválido es una escritura como cualquier otra y su fallo se
// propaga en lugar de descartarse (research.md D15).
var errEscrituraRota = errors.New("la tubería está cerrada")

// presentadorDoble es el doble en memoria de la interfaz que el kernel consume.
// Guarda en dos buffers separados lo que sale por cada descriptor, que es lo que
// permite comprobar la mitad de FR-036 que ningún otro test puede ver: que el
// registro de eventos y el aviso del valor inválido van al escritor de error y
// la salida estándar queda intacta.
//
// El paquete de presentación llega en T008; hasta entonces este doble es la
// única implementación de la interfaz, y que baste para montar el registrador
// demuestra que el kernel no necesita importar aquel paquete (research.md D2).
type presentadorDoble struct {
	salida  escritor
	errores escritor
	avisos  []string
	textos  []string
	sobres  []schema.Sobre
	// fallo, si no es nulo, es lo que devuelven las tres escrituras del
	// presentador, como haría una tubería cerrada.
	fallo error
}

// El doble tiene que satisfacer la interfaz completa, no la parte que este test
// usa: si el contrato del presentador crece, esto deja de compilar.
var _ Presentador = (*presentadorDoble)(nil)

func (p *presentadorDoble) Presentar(sobre schema.Sobre, enJSON bool) error {
	p.sobres = append(p.sobres, sobre)

	return p.escribir(&p.salida, fmt.Sprintf("sobre(ok=%t, json=%t)", sobre.Ok, enJSON))
}

func (p *presentadorDoble) Texto(texto string) error {
	p.textos = append(p.textos, texto)

	return p.escribir(&p.salida, texto)
}

func (p *presentadorDoble) Aviso(texto string) error {
	p.avisos = append(p.avisos, texto)

	return p.escribir(&p.errores, texto)
}

func (p *presentadorDoble) Salida() io.Writer { return &p.salida }

func (p *presentadorDoble) Error() io.Writer { return &p.errores }

func (p *presentadorDoble) escribir(destino *escritor, texto string) error {
	if p.fallo != nil {
		return p.fallo
	}

	_, err := io.WriteString(destino, texto+"\n")

	return err
}

// escritor es un buffer de bytes mínimo. No se usa bytes.Buffer porque el doble
// entrega el escritor al manejador de slog y lo lee después desde el test, y un
// tipo propio deja claro que ese ir y venir es la única operación que necesita.
type escritor struct {
	contenido []byte
}

func (e *escritor) Write(p []byte) (int, error) {
	e.contenido = append(e.contenido, p...)

	return len(p), nil
}

func (e *escritor) String() string { return string(e.contenido) }

// nivelesConocidos son los cuatro niveles del vocabulario de KITLEGAL_LOG, de
// menor a mayor. Preguntar por los cuatro al registrador ya montado es lo que
// comprueba el nivel resuelto sin depender de cómo esté escrita la resolución.
var nivelesConocidos = []slog.Level{
	slog.LevelDebug,
	slog.LevelInfo,
	slog.LevelWarn,
	slog.LevelError,
}

// casoDeNivel es una fila de la tabla de resolución del nivel: lo que dice la
// variable de entorno, lo que dice la bandera, el nivel que sale de ahí y si la
// invocación merece un aviso.
type casoDeNivel struct {
	nombre  string
	entorno string
	verbose bool
	nivel   slog.Level
	aviso   bool
}

// casosDeNivel cubre la regla de tres escalones de research.md D14: la
// variable si está presente y es válida, luego --verbose, y warn por omisión.
func casosDeNivel() []casoDeNivel {
	return []casoDeNivel{
		{
			nombre: "sin variable y sin bandera, warn por omisión",
			nivel:  slog.LevelWarn,
		},
		{
			nombre:  "--verbose equivale a debug",
			verbose: true,
			nivel:   slog.LevelDebug,
		},
		{
			nombre:  "la variable fija debug",
			entorno: "debug",
			nivel:   slog.LevelDebug,
		},
		{
			nombre:  "la variable fija info",
			entorno: "info",
			nivel:   slog.LevelInfo,
		},
		{
			nombre:  "la variable fija warn",
			entorno: "warn",
			nivel:   slog.LevelWarn,
		},
		{
			nombre:  "la variable fija error",
			entorno: "error",
			nivel:   slog.LevelError,
		},
		{
			nombre:  "la variable manda sobre la bandera, bajando el detalle",
			entorno: "error",
			verbose: true,
			nivel:   slog.LevelError,
		},
		{
			nombre:  "la variable manda sobre la bandera, subiéndolo sin ella",
			entorno: "debug",
			nivel:   slog.LevelDebug,
		},
		{
			nombre:  "la variable vacía se trata como ausente",
			entorno: "",
			verbose: true,
			nivel:   slog.LevelDebug,
		},
		{
			nombre:  "un valor inválido no aborta: avisa y cae en el nivel por omisión",
			entorno: "ruidoso",
			nivel:   slog.LevelWarn,
			aviso:   true,
		},
		{
			nombre:  "un valor inválido con --verbose avisa y cae en debug",
			entorno: "DEBUG",
			verbose: true,
			nivel:   slog.LevelDebug,
			aviso:   true,
		},
	}
}

// TestRegistro comprueba el registrador que el kernel monta: los cuatro niveles,
// la prioridad de KITLEGAL_LOG sobre --verbose, el aviso —no el aborto— ante un
// valor inválido, y que nada de todo eso llega a la salida estándar (FR-025,
// FR-036 … FR-038, FR-040 parcial).
func TestRegistro(t *testing.T) {
	t.Parallel()

	for _, caso := range casosDeNivel() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			doble := &presentadorDoble{}
			registrador, err := NuevoRegistrador(doble, caso.entorno, caso.verbose)
			require.NoError(t, err)
			require.NotNil(t, registrador)

			for _, nivel := range nivelesConocidos {
				assert.Equal(t, nivel >= caso.nivel, registrador.Enabled(t.Context(), nivel),
					"el nivel resuelto es %s; %s debería estar %s",
					caso.nivel, nivel, siHabilitado(nivel >= caso.nivel))
			}

			if caso.aviso {
				assert.Len(t, doble.avisos, 1)
			} else {
				assert.Empty(t, doble.avisos)
			}

			assert.Empty(t, doble.salida.String(),
				"ni el registro de eventos ni su aviso tocan la salida estándar")
		})
	}

	t.Run("todo el registro va al escritor de error", func(t *testing.T) {
		t.Parallel()

		doble := &presentadorDoble{}
		registrador, err := NuevoRegistrador(doble, "debug", false)
		require.NoError(t, err)

		mensajes := []string{"depuración", "información", "aviso", "fallo"}
		registrador.Debug(mensajes[0])
		registrador.Info(mensajes[1])
		registrador.Warn(mensajes[2])
		registrador.Error(mensajes[3])

		for _, mensaje := range mensajes {
			assert.Contains(t, doble.errores.String(), mensaje)
		}
		assert.Empty(t, doble.salida.String())
	})

	t.Run("el aviso del valor inválido no viaja por el registro de eventos", func(t *testing.T) {
		t.Parallel()

		doble := &presentadorDoble{}
		_, err := NuevoRegistrador(doble, "ruidoso", false)
		require.NoError(t, err, "una variable mal escrita no es un error de la invocación")

		require.Len(t, doble.avisos, 1)
		assert.Contains(t, doble.avisos[0], VariableNivel,
			"el aviso nombra la variable para que quien la escribió sepa dónde mirar")
		assert.Contains(t, doble.avisos[0], "ruidoso", "y nombra el valor que se ignora")

		// Un registro de slog lleva siempre `level=` y `msg=`. Que el escritor de
		// error no los tenga es lo que demuestra que el aviso lo escribe el
		// presentador y que por tanto ningún nivel puede ocultarlo —ni siquiera el
		// que el propio valor inválido acabaría de fijar— (research.md D14).
		assert.NotContains(t, doble.errores.String(), "level=")
		assert.NotContains(t, doble.errores.String(), "msg=")
		assert.Empty(t, doble.salida.String())
	})

	t.Run("el fallo al escribir el aviso se propaga y deja registrador utilizable", func(t *testing.T) {
		t.Parallel()

		doble := &presentadorDoble{fallo: errEscrituraRota}
		registrador, err := NuevoRegistrador(doble, "ruidoso", true)

		require.ErrorIs(t, err, errEscrituraRota)
		require.NotNil(t, registrador)
		assert.True(t, registrador.Enabled(t.Context(), slog.LevelDebug),
			"el nivel se resuelve igual aunque el aviso no se haya podido escribir")
	})

	t.Run("una invocación correcta se registra en info y no se ve por omisión", func(t *testing.T) {
		t.Parallel()

		correcta := Evento{Applet: "echo", Verbo: "repetir", Duracion: time.Millisecond}

		porOmision := &presentadorDoble{}
		registrador, err := NuevoRegistrador(porOmision, "", false)
		require.NoError(t, err)
		RegistrarEvento(t.Context(), registrador, correcta)
		assert.Empty(t, porOmision.errores.String(),
			"con warn por omisión, una invocación correcta deja la salida de error limpia")

		conInfo := &presentadorDoble{}
		registrador, err = NuevoRegistrador(conInfo, "info", false)
		require.NoError(t, err)
		RegistrarEvento(t.Context(), registrador, correcta)
		registrado := conInfo.errores.String()
		assert.Contains(t, registrado, "applet=echo")
		assert.Contains(t, registrado, "verbo=repetir")
		assert.NotContains(t, registrado, "clase=",
			"una invocación sin fallo no tiene clase de error que registrar")
		assert.Empty(t, conInfo.salida.String())
	})
}

// siHabilitado da el texto del fallo de la tabla de niveles: sin él, un caso
// roto solo dice «false != true».
func siHabilitado(habilitado bool) string {
	if habilitado {
		return "habilitado"
	}

	return "deshabilitado"
}

// nifDeLaPersona es el argumento que identifica a una persona física. Es el dato
// que FR-039 prohíbe registrar por omisión, y el que este test persigue por la
// salida de error en cada nivel.
const nifDeLaPersona = "12345678Z"

// eventoConFallo es lo que el kernel registra de una invocación que terminó mal:
// applet, verbo, duración y clase de error —nada de ello identifica a nadie— más
// los argumentos, que sí, y que por eso solo se registran en debug.
func eventoConFallo() Evento {
	return Evento{
		Applet:     "boe",
		Verbo:      "buscar",
		Duracion:   37 * time.Millisecond,
		Clase:      schema.ClaseNoEncontrado,
		Argumentos: []string{"buscar", nifDeLaPersona},
	}
}

// casoDePrivacidad es una fila de la tabla de FR-039: cómo se resuelve el nivel
// y si los argumentos de la invocación deben aparecer en el registro.
type casoDePrivacidad struct {
	nombre        string
	entorno       string
	verbose       bool
	conArgumentos bool
}

// TestRegistroPrivacidad comprueba lo que FR-039 exige y que ningún otro caso
// verifica: que el registro emitido lleva applet, verbo, duración y clase de
// error en todos los niveles en que se emite, que los argumentos de la
// invocación —lo único que puede identificar a una persona física— quedan fuera
// con el nivel por omisión y con info, y que aparecen solo cuando el nivel
// resuelto es debug, tanto si lo pide --verbose como si lo pide KITLEGAL_LOG.
//
// Es el test que invoca el escenario 7 de quickstart.md por su nombre.
func TestRegistroPrivacidad(t *testing.T) {
	t.Parallel()

	casos := []casoDePrivacidad{
		{
			nombre: "con el nivel por omisión (warn), sin argumentos",
		},
		{
			nombre:  "con info, tampoco",
			entorno: "info",
		},
		{
			nombre:        "con --verbose, que resuelve debug, sí",
			verbose:       true,
			conArgumentos: true,
		},
		{
			nombre:        "con KITLEGAL_LOG=debug y sin bandera alguna, también",
			entorno:       "debug",
			conArgumentos: true,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			doble := &presentadorDoble{}
			registrador, err := NuevoRegistrador(doble, caso.entorno, caso.verbose)
			require.NoError(t, err)

			RegistrarEvento(t.Context(), registrador, eventoConFallo())

			registrado := doble.errores.String()
			require.NotEmpty(t, registrado,
				"una invocación fallida se registra desde warn, que es el nivel por omisión")

			assert.Contains(t, registrado, "applet=boe")
			assert.Contains(t, registrado, "verbo=buscar")
			assert.Contains(t, registrado, "duracion=")
			assert.Contains(t, registrado, "clase="+string(schema.ClaseNoEncontrado))

			if caso.conArgumentos {
				assert.Contains(t, registrado, nifDeLaPersona,
					"en debug los argumentos se registran: es el nivel que se pide a sabiendas")
			} else {
				assert.NotContains(t, registrado, "argumentos=")
				assert.NotContains(t, registrado, nifDeLaPersona,
					"FR-039: por omisión no se registra nada que identifique a una persona física")
			}

			assert.Empty(t, doble.salida.String())
		})
	}
}
