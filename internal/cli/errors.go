// Package cli es el kernel de la línea de órdenes: lo que un applet no tiene
// que declarar. Aquí viven los errores tipados que un applet devuelve y la
// única traducción de esos errores a los códigos de salida estables del
// proyecto.
//
// El vocabulario de clases de error no vive aquí sino en internal/core/schema,
// porque aparece en el contrato JSON del sobre y en el esquema que emite
// --describe. Lo que es del kernel son los sentinelas con los que un applet
// nombra la clase y la tabla que convierte la clase en código de salida: un
// applet no conoce ningún número (FR-029, research.md D8).
package cli

import (
	"errors"
	"slices"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Los cinco sentinelas con los que un applet declara la clase de su fallo. Se
// devuelven envueltos, con el contexto que haga falta para la persona —
// fmt.Errorf("el bloque a21 de %s: %w", id, cli.ErrNoEncontrado)—, porque la
// clasificación usa errors.Is y envolver no cambia la clase (FR-032).
//
// No hay sentinela para la clase inesperada, y no es un olvido: lo inesperado
// es precisamente lo que nadie declaró, así que se reconoce por no casar con
// ninguno de estos cinco ni traer clase propia (FR-031, FR-063).
var (
	// ErrArgumentos es la invocación mal formada: bandera desconocida, valor
	// con formato inválido, argumento obligatorio ausente o applet no
	// registrado.
	ErrArgumentos = errors.New("argumentos inválidos")
	// ErrNoEncontrado es lo pedido que no existe en la fuente.
	ErrNoEncontrado = errors.New("no encontrado")
	// ErrFuenteNoDisponible es la fuente que no responde, y también el plazo
	// agotado.
	ErrFuenteNoDisponible = errors.New("fuente no disponible")
	// ErrLimiteOTos es el límite de peticiones alcanzado o la restricción de
	// los términos de uso.
	ErrLimiteOTos = errors.New("límite de peticiones o términos de uso")
	// ErrIdentidadHumana es la acción que requiere identidad humana y que, por
	// tanto, no se ha realizado.
	ErrIdentidadHumana = errors.New("requiere identidad humana")
)

// Los códigos de salida del proyecto, que son una decisión cerrada anterior a
// este hito (CLAUDE.md, «Exit codes estables»). El 1 es el del fallo
// inesperado: la convención de Unix para el error general, y el único valor
// libre que un consumidor interpreta sin documentación (FR-031, research.md
// D8).
const (
	codigoCorrecto           = 0
	codigoInesperado         = 1
	codigoArgumentos         = 2
	codigoNoEncontrado       = 3
	codigoFuenteNoDisponible = 4
	codigoLimiteOTos         = 5
	codigoIdentidadHumana    = 6
)

// Clasificar decide la clase de un error por dos vías, en este orden. Primero
// lo compara con los cinco sentinelas, en el orden en que están declarados y
// con errors.Is, de modo que un applet puede envolver el sentinela con todo el
// contexto que necesite sin que la clase —ni el código de salida que sale de
// ella— cambie (FR-032). Después pregunta al propio error: quien implementa
// schema.ConClase declara su clase sin importar este paquete, y errors.As la
// encuentra aunque vaya envuelta, que es como un adaptador —el cliente HTTP—
// dice a qué código de salida corresponde su fallo sin conocer ninguno
// (FR-063, research.md D4).
//
// Los sentinelas van primero a propósito: ningún adaptador los envuelve —no
// importan este paquete—, así que en la práctica las dos vías no compiten, y
// este orden conserva intacto lo que H1 clasificaba.
//
// Una clase declarada que no está en el vocabulario no se da por buena: el
// sobre de fallo la llevaría a una clave que el esquema de --describe restringe
// a las seis, así que lo que se inventa su clase acaba donde acaba todo lo que
// nadie previó, en la clase inesperada.
//
// Un error que no casa con ninguna de las dos vías es inesperado, y ahí está el
// motivo de que la rama por defecto viva aquí y no en el switch de
// codigoDeClase: así ese switch puede cubrir las seis clases sin rama
// `default`, y el linter exhaustive falla si alguien añade una clase nueva y se
// olvida de darle código (FR-030).
//
// Un error nulo no tiene clase: la ausencia de fallo no es una de las seis. Por
// eso quien tenga un error que puede ser nulo llama a CodigoSalida, que sí
// distingue el éxito; Clasificar devuelve la clase inesperada, que nunca se
// confunde con un éxito.
func Clasificar(err error) schema.Clase {
	switch {
	case errors.Is(err, ErrArgumentos):
		return schema.ClaseArgumentos
	case errors.Is(err, ErrNoEncontrado):
		return schema.ClaseNoEncontrado
	case errors.Is(err, ErrFuenteNoDisponible):
		return schema.ClaseFuenteNoDisponible
	case errors.Is(err, ErrLimiteOTos):
		return schema.ClaseLimiteOTos
	case errors.Is(err, ErrIdentidadHumana):
		return schema.ClaseIdentidadHumana
	}

	var conClase schema.ConClase
	if errors.As(err, &conClase) {
		if clase := conClase.Clase(); slices.Contains(schema.Clases(), clase) {
			return clase
		}
	}

	return schema.ClaseInesperado
}

// CodigoSalida es el único sitio del proyecto donde un error se convierte en un
// código de salida. Lo llama la raíz de composición del kernel, una sola vez
// por invocación, y nadie más (FR-030).
func CodigoSalida(err error) int {
	if err == nil {
		return codigoCorrecto
	}

	return codigoDeClase(Clasificar(err))
}

// codigoDeClase es la tabla de FR-029, escrita como un switch sobre un tipo con
// constantes y **sin rama `default`**: el linter exhaustive —activo desde H0—
// falla si aparece una clase sin código, y una rama por defecto lo haría
// conformarse (`default-signifies-exhaustive` está activado en .golangci.yml).
//
// El retorno final no es una rama por defecto disfrazada: schema.Clase es un
// tipo con base string, así que existen valores fuera del vocabulario de las
// seis constantes. Un valor así no lo produce ninguna ruta de este paquete
// —Clasificar solo devuelve constantes—, pero si alguien lo construyera, lo que
// no se ha previsto sale con el código de lo no previsto y nunca con el del
// éxito (FR-031).
func codigoDeClase(clase schema.Clase) int {
	switch clase {
	case schema.ClaseInesperado:
		return codigoInesperado
	case schema.ClaseArgumentos:
		return codigoArgumentos
	case schema.ClaseNoEncontrado:
		return codigoNoEncontrado
	case schema.ClaseFuenteNoDisponible:
		return codigoFuenteNoDisponible
	case schema.ClaseLimiteOTos:
		return codigoLimiteOTos
	case schema.ClaseIdentidadHumana:
		return codigoIdentidadHumana
	}

	return codigoInesperado
}
