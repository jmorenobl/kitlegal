package ejemplo

import (
	"context"
	"log/slog"
	"strings"
	"unicode/utf8"

	"github.com/jmorenobl/kitlegal/internal/app"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// La procedencia con la que firma `contar`, del mismo espacio de nombres
// reservado que `echo` y por el mismo motivo: lo que devuelve es un cálculo
// sobre sus argumentos y no la cita de ninguna fuente (FR-016).
const (
	fuenteContar = "kitlegal.contar"
	urlContar    = "kitlegal:applet/contar"
)

// Contar es el segundo applet de ejemplo, y existe para demostrar dos cosas que
// `echo` no puede demostrar:
//
//   - que un applet puede declarar **dos** verbos y **ninguno** por omisión, de
//     modo que invocarlo sin nombrar verbo termina en código 2
//     (contracts/registro-y-describe.md §2 bis);
//   - que un applet que no declara más que su nombre, sus verbos y el contenido
//     de su `data` hereda exactamente lo mismo que el primero: las ocho
//     banderas globales, el sobre, la huella, los códigos de salida, las dos
//     formas de presentación, --describe y la ayuda (SC-010).
func Contar() app.Applet { return appletContar{} }

// appletContar declara lo mismo que el otro applet y en la misma forma: si
// hiciera falta algo más para heredar el kernel, SC-010 no se estaría
// cumpliendo.
type appletContar struct{}

func (appletContar) Nombre() string { return "contar" }

//nolint:misspell // «Descripcion» es español y lo fija el contrato; el diccionario de misspell es solo inglés (research.md D9).
func (appletContar) Descripcion() string { return "Cuenta lo que hay en un texto." }

func (appletContar) Verbos() []app.Verbo {
	return []app.Verbo{
		{
			Nombre: "letras",
			//nolint:misspell // «Descripcion» es español y lo fija el contrato; el diccionario de misspell es solo inglés (research.md D9).
			Descripcion: "Cuenta las letras del texto.",
			Argumentos:  func() app.Argumentos { return &argumentosLetras{} },
			Salida:      cuenta{},
		},
		{
			Nombre: "palabras",
			//nolint:misspell // «Descripcion» es español y lo fija el contrato; el diccionario de misspell es solo inglés (research.md D9).
			Descripcion: "Cuenta las palabras del texto.",
			Argumentos:  func() app.Argumentos { return &argumentosPalabras{} },
			Salida:      cuenta{},
		},
	}
}

// cuenta es el contenido de `data` de los dos verbos: el texto sobre el que se
// contó y el total. Es el mismo tipo para ambos porque es la misma forma; lo
// que cambia entre un verbo y otro es qué se cuenta, no cómo se devuelve.
type cuenta struct {
	Texto string `json:"texto"`
	Total int    `json:"total"`
}

// argumentosLetras y argumentosPalabras son dos tipos distintos y no uno
// compartido, y no es una duplicación evitable: el kernel arma la gramática con
// un campo por verbo y recupera los argumentos del verbo seleccionado por ese
// campo, así que es el tipo el que distingue una operación de la otra.
type argumentosLetras struct {
	Texto string `arg:"" optional:"" help:"Texto cuyas letras se cuentan."`
}

type argumentosPalabras struct {
	Texto string `arg:"" optional:"" help:"Texto cuyas palabras se cuentan."`
}

// Ejecutar cuenta las letras en runas y no en octetos: «canción» son siete
// letras y nueve octetos, y lo que se pidió contar son letras.
func (a *argumentosLetras) Ejecutar(
	_ context.Context, _ schema.Contexto, _ *slog.Logger,
) (schema.Resultado, error) {
	return resultadoDeContar(a.Texto, utf8.RuneCountInString(a.Texto)), nil
}

// Ejecutar cuenta las palabras como bloques separados por espacio en blanco, que
// es lo que hace strings.Fields: un texto vacío o todo espacios son cero
// palabras.
func (a *argumentosPalabras) Ejecutar(
	_ context.Context, _ schema.Contexto, _ *slog.Logger,
) (schema.Resultado, error) {
	return resultadoDeContar(a.Texto, len(strings.Fields(a.Texto))), nil
}

// resultadoDeContar es lo común a los dos verbos: la procedencia del applet y la
// forma del contenido. Escrito una sola vez, los dos verbos no pueden divergir
// en cómo citan lo que devuelven.
func resultadoDeContar(texto string, total int) schema.Resultado {
	return schema.Resultado{
		Procedencia: schema.Procedencia{Fuente: fuenteContar, URL: urlContar},
		Datos:       cuenta{Texto: texto, Total: total},
	}
}

// Las tres comprobaciones en tiempo de compilación del contrato.
var (
	_ app.Applet     = appletContar{}
	_ app.Argumentos = (*argumentosLetras)(nil)
	_ app.Argumentos = (*argumentosPalabras)(nil)
)
