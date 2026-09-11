// Package app es la composición del binario: el registro de applets —única
// fuente de la que se derivan el despacho, la ayuda y la autodescripción—, el
// despacho multicall y la raíz que monta el kernel con el presentador.
//
// Aquí vive también el contrato hacia quien escribe un applet: qué tiene que
// declarar —nombre, verbos y contenido de su data— y qué recibe gratis —las
// ocho banderas globales, el sobre, la huella, la fecha de consulta, los
// códigos de salida, las dos formas de presentación y el registro de eventos—
// (FR-018, FR-044, SC-010, contracts/registro-y-describe.md §1).
package app

import (
	"context"
	"log/slog"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Applet es lo que declara un applet: su identidad y su catálogo de verbos. Un
// applet no declara ninguna bandera, ningún código de salida y ninguna forma de
// presentación, y los hereda igualmente del kernel (SC-010).
type Applet interface {
	// Nombre es con lo que se le invoca —«boe», «cita», «plazos»—, sea como
	// primer argumento o por el nombre de un enlace simbólico (FR-002, FR-003).
	Nombre() string
	// Descripcion es la línea que aparece en la ayuda del binario.
	Descripcion() string
	// Verbos es el catálogo del applet: al menos uno, con nombres únicos y como
	// mucho uno marcado por omisión, lo que comprueba el registro al construirse
	// (FR-008).
	Verbos() []Verbo
}

// Verbo es una operación del applet: cómo se nombra, qué argumentos acepta y de
// qué forma es el contenido de su data.
type Verbo struct {
	// Nombre es el verbo tal y como se escribe en la invocación: «articulo»,
	// «buscar», «resolver».
	Nombre string
	// Descripcion es la línea que aparece en la ayuda del applet.
	Descripcion string
	// Argumentos es una **fábrica**: devuelve un valor nuevo en cada invocación
	// y nunca una instancia compartida, de modo que dos invocaciones simultáneas
	// no se pisan los argumentos. El valor que devuelve lleva las etiquetas con
	// las que el analizador construye la gramática del verbo.
	Argumentos func() Argumentos
	// Salida es el valor cero del tipo del contenido de data, y existe **solo**
	// para reflejarlo en --describe: nunca se serializa ni se devuelve a quien
	// invoca. Nulo significa un verbo que no lo declara, y entonces el esquema
	// deja data sin restringir en lugar de inventarle una forma (FR-047).
	Salida any
	// PorOmision marca el verbo que se toma cuando la invocación no nombra
	// ninguno, que es lo que hace válida la entrega literal del hito
	// —«kitlegal echo hola»—. Un applet puede no marcar ninguno, y entonces
	// nombrar el verbo es obligatorio; marcar dos es un defecto que el registro
	// rechaza al construirse (research.md D26).
	PorOmision bool
}

// Argumentos es la operación del verbo, con sus argumentos ya analizados en el
// valor que devolvió la fábrica.
//
// El registrador de eventos llega montado y con el nivel ya resuelto: el applet
// no lee --verbose ni KITLEGAL_LOG, y por eso el registrador viaja como
// parámetro explícito y no dentro del contexto de ejecución, que es dominio puro
// y no importa log/slog (research.md D2).
//
// Lo que devuelve es un schema.Resultado —procedencia y datos—: nunca un sobre
// ya montado, nunca texto escrito en un descriptor y nunca un código de salida.
// De todo eso se ocupa el kernel, en un único punto (FR-015, FR-030, FR-044).
type Argumentos interface {
	Ejecutar(ctx context.Context, ec schema.Contexto, log *slog.Logger) (schema.Resultado, error)
}
