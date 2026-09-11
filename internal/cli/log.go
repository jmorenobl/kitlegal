package cli

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// VariableNivel es la variable de entorno con la que se fija el nivel del
// registro de eventos sin tocar la invocación (FR-037). Su nombre se escribe
// aquí una sola vez: la raíz de composición la lee y entrega su valor a
// NuevoRegistrador, que es lo que mantiene la resolución del nivel libre de
// entrada y salida y comprobable en paralelo.
const VariableNivel = "KITLEGAL_LOG"

// El vocabulario cerrado de VariableNivel, en minúscula y sin abreviaturas. Un
// valor fuera de esta lista no aborta la invocación: se ignora, se avisa y el
// nivel se resuelve como si la variable no estuviera (research.md D14).
const (
	nivelDebug = "debug"
	nivelInfo  = "info"
	nivelWarn  = "warn"
	nivelError = "error"
)

// NuevoRegistrador monta el registrador de eventos de una invocación: un
// manejador estructurado —de texto, que también es filtrable por un consumidor
// automático (FR-038)— apuntando siempre al escritor de error del presentador,
// nunca a la salida estándar (FR-036, FR-040).
//
// El nivel se resuelve en tres escalones: el valor de VariableNivel si está
// presente y es válido, después --verbose —que equivale a debug (FR-025)— y warn
// por omisión, que deja la salida de error limpia salvo cuando hay algo que
// decir. Se resuelve antes de que exista gramática que analizar, de modo que un
// fallo del análisis también quede registrado; el valor de verbose lo aporta
// hasta entonces el pre-escaneo (research.md D14, D25).
//
// entorno es el valor en crudo de VariableNivel, vacío si no está definida. Un
// valor inválido no aborta la invocación y no es un error de argumentos: se
// avisa por el presentador, que es el canal que ningún nivel filtra —el aviso no
// puede quedar oculto precisamente por el valor inválido que lo motiva—. El
// registrador que se devuelve es utilizable en todos los casos, también cuando
// el aviso no se ha podido escribir y el fallo de esa escritura sube como error.
func NuevoRegistrador(p Presentador, entorno string, verbose bool) (*slog.Logger, error) {
	nivel, valido := nivelResuelto(entorno, verbose)

	registrador := slog.New(slog.NewTextHandler(p.Error(), &slog.HandlerOptions{Level: nivel}))
	if valido {
		return registrador, nil
	}

	return registrador, p.Aviso(fmt.Sprintf(
		"%s=%q no nombra ningún nivel de registro (%s, %s, %s o %s); se ignora y se usa %s",
		VariableNivel, entorno,
		nivelDebug, nivelInfo, nivelWarn, nivelError,
		strings.ToLower(nivel.String())))
}

// nivelResuelto aplica los tres escalones de la resolución y dice, además, si el
// valor de la variable era admisible. La variable manda sobre la bandera en
// ambos sentidos: baja el detalle que --verbose pedía y lo sube sin que nadie lo
// pida.
//
// La comparación es exacta: el vocabulario son cuatro palabras en minúscula, y
// aceptar otras formas sería inventar un contrato que nadie ha escrito. Lo que se
// escriba de otra forma cae en el aviso, que enumera las cuatro admitidas.
func nivelResuelto(entorno string, verbose bool) (nivel slog.Level, valido bool) {
	switch entorno {
	case nivelDebug:
		return slog.LevelDebug, true
	case nivelInfo:
		return slog.LevelInfo, true
	case nivelWarn:
		return slog.LevelWarn, true
	case nivelError:
		return slog.LevelError, true
	case "":
		// La variable no está definida, que no es lo mismo que estar mal escrita.
		return nivelDeLaBandera(verbose), true
	}

	return nivelDeLaBandera(verbose), false
}

// nivelDeLaBandera son los dos escalones que quedan cuando la variable no
// decide: --verbose y el nivel por omisión.
func nivelDeLaBandera(verbose bool) slog.Level {
	if verbose {
		return slog.LevelDebug
	}

	return slog.LevelWarn
}

// Evento es lo que el kernel registra de una invocación, y está escrito para que
// no quepa en él nada que identifique a una persona física: ni el contenido de
// data, ni la fuente consultada, ni el asunto sobre el que se trabaja (FR-039).
//
// Los argumentos son la excepción, porque un argumento sí puede serlo —un NIF,
// un nombre, una dirección—. Viajan en el evento porque en depuración hacen
// falta, y RegistrarEvento decide si se emiten: solo cuando el nivel resuelto es
// debug, que es un nivel que se pide a sabiendas.
type Evento struct {
	// Applet es el applet que atendió la invocación.
	Applet string
	// Verbo es el verbo que se ejecutó, ya normalizado.
	Verbo string
	// Duracion es lo que tardó la operación completa.
	Duracion time.Duration
	// Clase es la clase del fallo con el que terminó la invocación, y va vacía
	// cuando fue correcta.
	Clase schema.Clase
	// Argumentos son los argumentos de la invocación. Solo se registran en debug.
	Argumentos []string
}

// mensajeEvento es el mensaje del único registro que el kernel emite por
// invocación. Es fijo a propósito: lo que distingue un evento de otro son sus
// atributos, que es lo que un consumidor automático filtra (FR-038).
const mensajeEvento = "invocación"

// RegistrarEvento emite el registro de una invocación por el escritor de error,
// al nivel que le corresponde: warn si terminó con un fallo —hay algo que decir,
// y por eso se ve con el nivel por omisión— e info si fue correcta, que con warn
// por omisión deja la salida de error limpia (research.md D14).
//
// Los argumentos solo se añaden cuando el registrador tiene habilitado debug, de
// modo que la decisión de FR-039 se toma en un único sitio y no en cada llamada.
func RegistrarEvento(ctx context.Context, registrador *slog.Logger, ev Evento) {
	atributos := []slog.Attr{
		slog.String("applet", ev.Applet),
		slog.String("verbo", ev.Verbo),
		slog.Duration("duracion", ev.Duracion),
	}

	nivel := slog.LevelInfo
	if ev.Clase != "" {
		nivel = slog.LevelWarn
		atributos = append(atributos, slog.String("clase", string(ev.Clase)))
	}

	if registrador.Enabled(ctx, slog.LevelDebug) {
		atributos = append(atributos, slog.Any("argumentos", ev.Argumentos))
	}

	registrador.LogAttrs(ctx, nivel, mensajeEvento, atributos...)
}
