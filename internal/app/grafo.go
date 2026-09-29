package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/grafo"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/graph"
	"github.com/jmorenobl/kitlegal/internal/source/boe"
)

// La procedencia con la que firma graph. Lee el grafo del mundo, que es local y
// que escribe el propio binario, así que no consulta ninguna fuente pública y
// firma en el espacio de nombres reservado; la procedencia de cada observación
// va, dato a dato, dentro de data (FR-051; contracts/applet-graph.md §2).
const (
	fuenteDelGrafo = "kitlegal.graph"
	urlDelGrafo    = "kitlegal:applet/graph"
)

// errSinReloj es el defecto de composición de unas dependencias sin reloj: el
// applet no puede fechar su respuesta ni comprobar nada, y no es algo que quien
// pregunta pueda corregir. Sale como inesperado, sin procedencia, que firma y
// fecha el kernel.
var errSinReloj = errors.New("graph: el applet se compuso sin reloj y no puede fechar ninguna respuesta")

// DependenciasDeGrafo es lo que el applet graph recibe de la raíz de
// composición (contracts/applet-graph.md §1; research.md D19). El binario
// distribuido da las de DependenciasDelGrafoDelSistema; las pruebas y el binario
// de e2e, un reloj fijo y, las pruebas, world.db en su carpeta.
type DependenciasDeGrafo struct {
	// Reloj da el instante de la invocación, que es la fecha_consulta del sobre
	// y el instante con el que check compara (FR-051, FR-067). Se lee una sola
	// vez por invocación; nulo es un defecto de composición.
	Reloj func() time.Time
	// Almacen son las opciones con las que se ubica world.db. Vacío: la regla de
	// ubicación de la caché —KITLEGAL_CACHE_DIR o el directorio de la cuenta—,
	// resuelta en cada invocación (FR-001, FR-011).
	Almacen []graph.Opcion
}

// DependenciasDelGrafoDelSistema son las del binario distribuido: el reloj del
// sistema y world.db donde lo pone la regla de ubicación de la caché.
func DependenciasDelGrafoDelSistema() DependenciasDeGrafo {
	return DependenciasDeGrafo{Reloj: time.Now}
}

// AppletGrafo es el applet que lee el grafo del mundo, con tres verbos y
// ninguno por omisión: show, stats y check (FR-050; contracts/applet-graph.md
// §1). Ninguno escribe nada ni entrega operaciones —su Resultado.Grafo es el
// valor cero (FR-004, FR-005, FR-046)—, así que --no-graph y --offline no
// cambian lo que leen, y con --dry-run leen igual y el kernel escribe su
// descripción, sin sobre (contracts/applet-graph.md §2).
//
// Componerlo no abre nada: cada invocación resuelve la ruta de world.db, lo lee
// y lo cierra.
func AppletGrafo(dependencias DependenciasDeGrafo) Applet {
	return appletGrafo{dependencias: dependencias}
}

// appletGrafo declara lo que declara un applet y nada más: su nombre, su línea
// de ayuda y su catálogo de verbos, que llevan consigo las dependencias.
type appletGrafo struct {
	dependencias DependenciasDeGrafo
}

func (appletGrafo) Nombre() string { return "graph" }

func (appletGrafo) Descripcion() string {
	return "Lee el grafo del mundo: lo que el binario ha observado de las fuentes, con su procedencia."
}

// Verbos son los tres del contrato, en su orden, con sus argumentos y el valor
// cero del tipo de su data (contracts/applet-graph.md §1 y §3).
func (a appletGrafo) Verbos() []Verbo {
	return []Verbo{
		{
			Nombre:      "show",
			Descripcion: "Devuelve un nodo del grafo del mundo con sus aristas y su procedencia, sin texto legal.",
			Argumentos:  func() Argumentos { return &argumentosDeShow{dependencias: a.dependencias} },
			Salida:      grafo.Ficha{},
		},
		{
			Nombre:      "stats",
			Descripcion: "Cuenta los nodos, las aristas y los textos del grafo del mundo por tipo, relación y fuente.",
			Argumentos:  func() Argumentos { return &argumentosDeStats{dependencias: a.dependencias} },
			Salida:      grafo.Recuento{},
		},
		{
			Nombre: "check",
			// La cota, escrita desde grafo.MaximoDeHallazgos: la ayuda dice la
			// que aplica Comprobar (H7.1 contracts/applet-graph.md §1).
			Descripcion: fmt.Sprintf("Comprueba lo consultado de una norma, de algunos de sus bloques o, sin"+
				" argumentos, todo lo consultado, y lista como mucho %d hallazgos: redacciones que han cambiado"+
				" desde la lectura anterior y consultas caducadas.", grafo.MaximoDeHallazgos),
			Argumentos: func() Argumentos { return &argumentosDeCheck{dependencias: a.dependencias} },
			Salida:     grafo.Comprobacion{},
		},
	}
}

// argumentosDeShow son los de show: exactamente un id, por su posición
// (FR-052). Ninguna bandera propia: el applet hereda las ocho globales.
//
// El id es un cli.Literal y no un string porque se busca tal cual, también con
// bytes que no son UTF-8, y el analizador los cambiaría por U+FFFD antes de
// llegar aquí en un string (FR-052, FR-053).
type argumentosDeShow struct {
	ID cli.Literal `arg:"" name:"id" help:"Id del nodo: un ELI, «ine:<código>», un DIR3…"`

	dependencias DependenciasDeGrafo
}

// Ejecutar devuelve la ficha del nodo del id: sus datos, sus observaciones y sus
// aristas, sin el cuerpo de ningún texto (FR-053, FR-070). Un id que no puede
// ser el de ningún nodo es «argumentos», sin abrir world.db (FR-052); uno que
// no está, también con el grafo ausente o sin esquema, «no encontrado», con un
// mensaje que lo nombra con sus bytes, escritos con %q (FR-053).
func (a *argumentosDeShow) Ejecutar(ctx context.Context, _ schema.Contexto, _ *slog.Logger) (schema.Resultado, error) {
	id := string(a.ID)

	return a.dependencias.responder(ctx, grafo.ValidarID(id),
		func(ctx context.Context, lectura *graph.Lectura, _ time.Time) (any, error) {
			ficha, encontrada, err := lectura.Ficha(ctx, id)
			if err != nil {
				return nil, err
			}

			if !encontrada {
				return nil, fmt.Errorf("%w: el id %q no está en el grafo del mundo", cli.ErrNoEncontrado, id)
			}

			return ficha, nil
		})
}

// argumentosDeStats son los de stats: ninguno, ni por su posición ni con una
// bandera propia (FR-054).
type argumentosDeStats struct {
	dependencias DependenciasDeGrafo
}

// Ejecutar cuenta los nodos, las aristas y los textos del grafo, por tipo,
// relación y fuente; el grafo ausente o vacío son tres ceros (FR-054).
func (a *argumentosDeStats) Ejecutar(ctx context.Context, _ schema.Contexto, _ *slog.Logger) (schema.Resultado, error) {
	return a.dependencias.responder(ctx, nil,
		func(ctx context.Context, lectura *graph.Lectura, _ time.Time) (any, error) {
			return lectura.Recuento(ctx)
		})
}

// argumentosDeCheck son los de check: la norma y sus bloques, por su posición y
// los dos opcionales, como los de boe articulos, y ninguna bandera propia
// (H7.1 FR-001, FR-015; research.md D6). La norma es un puntero para que una
// dada vacía —la de una variable sin valor— se distinga de no darla, que es
// comprobar todo lo consultado.
type argumentosDeCheck struct {
	Norma   *string  `arg:"" optional:"" name:"norma" help:"Identificador BOE de la norma, BOE-A-<año>-<número>; sin él, todo lo consultado."`
	Bloques []string `arg:"" optional:"" name:"bloques" help:"Ids de bloque de esa norma, como a21; sin ellos, todos los suyos."`

	dependencias DependenciasDeGrafo
}

// Ejecutar comprueba el ámbito de los argumentos —una norma y, si se nombran,
// bloques suyos; sin norma, todo lo consultado— en el instante del reloj de la
// invocación y devuelve su comprobación: los totales de cada clase y como mucho
// grafo.MaximoDeHallazgos hallazgos, una lista vacía y no nula si no hay
// ninguno. Encontrar algo es un resultado correcto (FR-060, FR-067; H7.1 FR-002,
// FR-010 a FR-012; ADR 0023). Unos argumentos que no valen son «argumentos», sin
// abrir world.db; una norma o un bloque bien formados que el grafo no conoce no
// traen nada (H7.1 FR-003, FR-004).
//
// Una instantánea con lo que ninguna entrega guarda —una fecha de consulta que
// no es RFC 3339— no se puede comprobar: es un world.db dañado, que sale como
// inesperado y no como una lista de hallazgos inventada ni vacía.
func (a *argumentosDeCheck) Ejecutar(ctx context.Context, _ schema.Contexto, _ *slog.Logger) (schema.Resultado, error) {
	ambito, argumentos := a.ambito()

	return a.dependencias.responder(ctx, argumentos,
		func(ctx context.Context, lectura *graph.Lectura, ahora time.Time) (any, error) {
			instantanea, err := lectura.Instantanea(ctx, ambito)
			if err != nil {
				return nil, err
			}

			comprobacion, err := grafo.Comprobar(instantanea, ambito, ahora)
			if err != nil {
				return nil, fmt.Errorf("grafo: world.db guarda lo que ninguna entrega escribe y no se puede"+
					" comprobar: %w", err)
			}

			return comprobacion, nil
		})
}

// ambito es el que piden los argumentos, o el fallo «argumentos» que los
// rechaza (contracts/applet-graph.md §2 de H7.1): una norma dada, también
// vacía, fuera de la gramática de boe.ValidarNorma —la misma de boe articulo—,
// o un bloque vacío o de solo espacio en blanco en el sentido de
// unicode.IsSpace. Cada mensaje nombra el valor.
func (a *argumentosDeCheck) ambito() (grafo.Ambito, error) {
	if a.Norma == nil {
		return grafo.Ambito{}, nil
	}

	if err := boe.ValidarNorma(*a.Norma); err != nil {
		return grafo.Ambito{}, fmt.Errorf("%w: %w", cli.ErrArgumentos, err)
	}

	for _, bloque := range a.Bloques {
		if strings.TrimFunc(bloque, unicode.IsSpace) == "" {
			return grafo.Ambito{}, fmt.Errorf("%w: el bloque %q está vacío o solo tiene espacio en blanco",
				cli.ErrArgumentos, bloque)
		}
	}

	return grafo.Ambito{Norma: *a.Norma, Bloques: a.Bloques}, nil
}

// Las comprobaciones en tiempo de compilación del contrato del applet.
var (
	_ Applet     = appletGrafo{}
	_ Argumentos = (*argumentosDeShow)(nil)
	_ Argumentos = (*argumentosDeStats)(nil)
	_ Argumentos = (*argumentosDeCheck)(nil)
)

// responder es el esqueleto de los tres verbos: lee el reloj una sola vez, firma
// con ese instante, rechaza los argumentos que el verbo ya sabe que no valen sin
// abrir nada, abre la lectura de world.db, lee con ella y la cierra.
//
// Todo fallo que decide el applet lleva su firma: el de los argumentos y el de
// internal/graph, que declara su clase —la ruta que no se puede ubicar es
// «argumentos»; el plazo de --timeout agotado, «fuente no disponible»; world.db
// inutilizable o bloqueado más de la espera propia, «inesperado»— y el kernel la
// traduce a 2, 4 o 1 (contracts/applet-graph.md §4). El reloj nulo es un
// defecto de composición y no firma nada. El resultado no lleva operaciones de
// grafo: ningún verbo observa el mundo (FR-046).
func (d DependenciasDeGrafo) responder(
	ctx context.Context,
	argumentos error,
	leer func(ctx context.Context, lectura *graph.Lectura, ahora time.Time) (any, error),
) (schema.Resultado, error) {
	if d.Reloj == nil {
		return schema.Resultado{}, errSinReloj
	}

	ahora := d.Reloj()
	firmado := schema.Resultado{
		Procedencia: schema.Procedencia{Fuente: fuenteDelGrafo, URL: urlDelGrafo, FechaConsulta: ahora},
	}

	if argumentos != nil {
		return firmado, argumentos
	}

	lectura, err := graph.Leer(ctx, d.Almacen...)
	if err != nil {
		return firmado, err
	}

	datos, err := leer(ctx, lectura, ahora)
	if cierre := lectura.Close(); cierre != nil {
		err = errors.Join(err, cierre)
	}

	if err != nil {
		return firmado, err
	}

	firmado.Datos = datos

	return firmado, nil
}
