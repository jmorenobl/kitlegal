package ejemplo

import (
	"context"
	"log/slog"

	"github.com/jmorenobl/kitlegal/internal/app"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// La procedencia con la que firma `echo`. Es el espacio de nombres reservado
// kitlegal. / kitlegal:, que señala un resultado **calculado** y no una cita de
// fuente pública: un applet que no consulta ninguna fuente no puede fingir que
// la cita (FR-016, contracts/sobre-de-salida.md §3).
const (
	fuenteEcho = "kitlegal.echo"
	urlEcho    = "kitlegal:applet/echo"
)

// Echo es el applet de la entrega literal del hito: «kitlegal echo hola --json».
// Declara un único verbo, `repetir`, marcado por omisión, que es lo que hace
// válida una invocación que no lo nombra (research.md D26).
func Echo() app.Applet { return appletEcho{} }

// appletEcho declara lo que declara un applet y nada más: su nombre, su línea de
// ayuda y su catálogo de verbos.
type appletEcho struct{}

func (appletEcho) Nombre() string { return "echo" }

func (appletEcho) Descripcion() string { return "Devuelve el mensaje que recibe." }

func (appletEcho) Verbos() []app.Verbo {
	return []app.Verbo{{
		Nombre:      "repetir",
		Descripcion: "Devuelve el mensaje tal y como se escribió.",
		Argumentos:  func() app.Argumentos { return &argumentosRepetir{} },
		Salida:      mensaje{},
		PorOmision:  true,
	}}
}

// mensaje es el contenido de `data` del verbo `repetir`. Se declara como tipo y
// no como un mapa suelto porque de él sale la mitad de salida del esquema que
// emite --describe, que se deriva de esta declaración y no se mantiene a mano
// (FR-047, FR-048).
type mensaje struct {
	Mensaje string `json:"mensaje"`
}

// argumentosRepetir son los argumentos del verbo, con las etiquetas de las que
// el kernel construye la gramática. El mensaje es opcional para que «echo
// repetir» —el primer argumento nombra el verbo, así que es el verbo— siga
// siendo una invocación válida.
type argumentosRepetir struct {
	Mensaje string `arg:"" optional:"" help:"Mensaje que se devuelve."`
}

// Ejecutar devuelve la procedencia y el contenido, y nada más: ni `ok`, ni
// huella, ni fecha de consulta, ni forma de presentación, ni código de salida.
// De todo eso se ocupa el kernel en un único punto (FR-015, FR-044).
//
// No recibe nada del contexto de ejecución ni del registrador porque no tiene
// nada que hacer con ellos: no consulta ninguna fuente, así que ni el plazo, ni
// --offline, ni el asunto cambian lo que devuelve.
func (a *argumentosRepetir) Ejecutar(
	_ context.Context, _ schema.Contexto, _ *slog.Logger,
) (schema.Resultado, error) {
	return schema.Resultado{
		Procedencia: schema.Procedencia{Fuente: fuenteEcho, URL: urlEcho},
		Datos:       mensaje{Mensaje: a.Mensaje},
	}, nil
}

// Las dos comprobaciones en tiempo de compilación del contrato: el applet
// declara los tres métodos y sus argumentos saben ejecutarse.
var (
	_ app.Applet     = appletEcho{}
	_ app.Argumentos = (*argumentosRepetir)(nil)
)
