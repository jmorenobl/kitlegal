package cli

import (
	"fmt"

	"github.com/alecthomas/kong"
)

// Decision es lo que queda por hacer cuando el análisis de una invocación
// termina bien. Son tres y se excluyen entre sí; cuál manda lo decide una
// prelación fija que no depende del orden en que se escriban las banderas
// (FR-028, research.md D11).
type Decision string

const (
	// DecisionEjecutar entrega el control al applet: es lo que ocurre cuando no
	// se pidió ni la ayuda ni la autodescripción.
	DecisionEjecutar Decision = "ejecutar"
	// DecisionAyuda dice que la ayuda ya está escrita en la salida estándar y
	// que la invocación termina con código 0. Gana sobre todo lo demás.
	DecisionAyuda Decision = "ayuda"
	// DecisionDescribir pide el esquema JSON del verbo en lugar de ejecutarlo.
	// Gana sobre la ejecución y pierde contra la ayuda.
	DecisionDescribir Decision = "describir"
)

// Analisis es lo que el kernel entiende de una invocación ya analizada: qué hay
// que hacer, con qué opciones globales y sobre qué verbo.
//
// Los argumentos del verbo no viajan aquí: quedan escritos en la gramática que
// aportó el paquete de composición, que es quien la construyó y quien la
// entrega después al applet.
type Analisis struct {
	// Decision es la rama que gana la prelación.
	Decision Decision
	// Globales son las ocho banderas ya analizadas, con sus valores por omisión
	// aplicados.
	Globales Globales
	// Verbo es el nombre del verbo seleccionado. Va vacío cuando la decisión es
	// la ayuda, que no resuelve ningún verbo (research.md D11, D26).
	Verbo string
}

// gramatica es la forma que tiene una invocación cuando el applet ya está
// resuelto: las ocho banderas globales embebidas y, colgando de ellas, los
// verbos de ese applet.
//
// Los verbos llegan como `any` dentro de kong.Plugins —el mecanismo con el que
// la biblioteca acepta una parte de la gramática que no se conoce al compilar—,
// y por eso este paquete no necesita importar el de composición ni conocer el
// tipo Applet: la dirección de dependencia del roadmap se mantiene
// (research.md D1, D2).
type gramatica struct {
	Globales
	kong.Plugins
}

// Analizar construye y analiza la gramática de **una** invocación, con el applet
// ya resuelto y el verbo ya normalizado por el paquete de composición.
//
// nombre es el nombre con el que se invocó —el del enlace simbólico o el del
// primer argumento—, y es el que aparece en la ayuda. verbos es la parte de la
// gramática que aporta ese applet: un puntero a un struct con un mandato por
// verbo. args son los argumentos ya separados del nombre del programa.
//
// A Kong se le entregan los dos escritores del presentador y una función de
// terminación que no termina nada: sin la primera, la ayuda saldría por un
// descriptor que no es el único que escribe (FR-040); sin la segunda, --help
// llamaría a os.Exit desde dentro de este paquete, que es justo lo que la regla
// de arquitectura y FR-035 prohíben (gates/supuestos-kong.md, S2 y S3).
//
// Todo fallo del análisis es un error de argumentos, código 2: la bandera
// desconocida, el valor con formato inválido, el argumento obligatorio ausente y
// el verbo desconocido llegan como un único tipo de error y la distinción entre
// ellos la lleva el mensaje, no la clase (FR-027, gates/supuestos-kong.md, S5).
// Lo que no se puede construir, en cambio, no es un fallo de quien invoca: una
// gramática inválida es un defecto del applet y sale como error inesperado.
func Analizar(p Presentador, nombre string, verbos any, args []string) (Analisis, error) {
	gram := gramatica{Plugins: kong.Plugins{verbos}}

	// La ayuda integrada escribe en la salida estándar y pide terminar con 0.
	// Anotarlo aquí es lo que hace que la prelación de --help no dependa de lo
	// que devuelva el análisis: tras escribir la ayuda, Kong sigue analizando y
	// acaba devolviendo el error de la gramática incompleta, que en ese caso no
	// es un fallo de la invocación sino la consecuencia de haber pedido ayuda.
	// Durante el análisis no hay ningún otro camino que llame a la terminación
	// (gates/supuestos-kong.md, detalle de S3).
	ayudaEmitida := false

	analizador, err := kong.New(&gram,
		kong.Name(nombre),
		kong.Writers(p.Salida(), p.Error()),
		kong.Exit(func(int) { ayudaEmitida = true }),
	)
	if err != nil {
		return Analisis{}, fmt.Errorf("la gramática de %s no se pudo construir: %w", nombre, err)
	}

	contexto, errAnalisis := analizador.Parse(args)

	if ayudaEmitida {
		return Analisis{Decision: DecisionAyuda, Globales: gram.Globales}, nil
	}

	if errAnalisis != nil {
		return Analisis{}, fmt.Errorf("%w: %w", ErrArgumentos, errAnalisis)
	}

	if err := validarGlobales(gram.Globales); err != nil {
		return Analisis{}, err
	}

	return Analisis{
		Decision: decisionDe(gram.Globales),
		Globales: gram.Globales,
		Verbo:    verboSeleccionado(contexto),
	}, nil
}

// validarGlobales comprueba lo que la gramática no puede expresar: que el plazo
// es una duración positiva. Un plazo de cero o negativo se analiza sin problema
// —«0» y «-5s» son duraciones válidas— pero no describe ninguna operación
// posible, así que es una invocación que hay que corregir y no una operación que
// vence en el acto (FR-020, FR-027).
//
// Se comprueba también cuando se pidió --describe: una invocación mal escrita lo
// está con independencia de lo que se fuera a hacer con ella.
func validarGlobales(g Globales) error {
	if g.Timeout <= 0 {
		return fmt.Errorf("%w: --timeout espera una duración positiva y recibió %s",
			ErrArgumentos, g.Timeout)
	}

	return nil
}

// decisionDe aplica los dos escalones de la prelación que quedan una vez
// descartada la ayuda, que se resuelve antes y gana sobre todo (research.md
// D11).
func decisionDe(g Globales) Decision {
	if g.Describe {
		return DecisionDescribir
	}

	return DecisionEjecutar
}

// verboSeleccionado lee del análisis el nombre del verbo, que es lo único que el
// kernel necesita saber de la parte de la gramática que no conoce.
//
// Un applet siempre declara al menos un verbo —lo comprueba el registro al
// construirse, no la invocación de un usuario—, así que un análisis correcto
// selecciona uno: si no se nombra ninguno, el análisis falla antes de llegar
// aquí. Que este paquete no dé por supuesta esa garantía, que vive en otro, es
// lo que evita convertir una gramática degenerada en un pánico.
func verboSeleccionado(contexto *kong.Context) string {
	nodo := contexto.Selected()
	if nodo == nil {
		return ""
	}

	return nodo.Name
}
