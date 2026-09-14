package app

import (
	"context"
	"log/slog"

	"github.com/jmorenobl/kitlegal/internal/cache"
	"github.com/jmorenobl/kitlegal/internal/core"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/httpx"
	"github.com/jmorenobl/kitlegal/internal/source/boe"
)

// DependenciasDeBoe es lo que el applet boe recibe de la raíz de composición:
// cómo se construye el cliente HTTP de la fuente y con qué opciones adicionales
// se abre la caché. El binario distribuido da las de DependenciasDeRed; el de
// e2e y las pruebas, un cliente en reproducción y una caché en su carpeta
// (contrato puerto-y-applet §4 y §5).
//
// Son una función y unas opciones, y no un cliente y una caché ya abiertos, a
// propósito: el applet compone la fuente en cada invocación, de modo que lo que
// se sirve de la caché no construye ningún cliente, lo que se invoca mal no abre
// ninguna caché y --offline y --dry-run la abren en solo lectura (FR-092,
// FR-094; research.md D5).
type DependenciasDeBoe struct {
	// Cliente construye, por invocación y solo si hace falta pedir, el cliente
	// HTTP de la fuente, con el registrador de eventos que el kernel entrega al
	// applet. Sin él, la invocación que llega a la fuente es un defecto de
	// composición (contrato errores-y-codigos, fila 22).
	Cliente func(registrador *slog.Logger) (*httpx.Cliente, error)
	// Cache son opciones adicionales para abrir la caché. Vacío: la de la cuenta
	// o la de KITLEGAL_CACHE_DIR.
	Cache []cache.Opcion
}

// DependenciasDeRed son las del binario distribuido: el cliente contra la red,
// con el nombre de la fuente, el intervalo entre peticiones que fija su fila de
// SOURCES.md y el registrador que el kernel entrega al applet, y la caché de la
// cuenta o la de KITLEGAL_CACHE_DIR (FR-122; contrato puerto-y-applet §4).
// Componerlas no construye nada: el cliente lo construye la invocación que pide
// algo.
func DependenciasDeRed() DependenciasDeBoe {
	return DependenciasDeBoe{
		Cliente: func(registrador *slog.Logger) (*httpx.Cliente, error) {
			return httpx.New(
				httpx.ConFuente(boe.NombreDeLaFuente),
				httpx.ConIntervalo(boe.IntervaloEntrePeticiones),
				httpx.ConRegistrador(registrador),
			)
		},
	}
}

// AppletBoe es el applet de la API de Legislación Consolidada del BOE, con sus
// seis verbos y ninguno por omisión: nombrar el verbo es obligatorio (FR-001).
// Cada verbo construye su consulta con los argumentos que recibe, compone la
// fuente con las dependencias y devuelve tal cual lo que responde su Fetch; no
// lee ninguna bandera ni escribe nada (contrato puerto-y-applet §4).
func AppletBoe(dependencias DependenciasDeBoe) Applet {
	return appletBoe{dependencias: dependencias}
}

// appletBoe declara lo que declara un applet y nada más: su nombre, su línea de
// ayuda y su catálogo de verbos, que llevan consigo las dependencias.
type appletBoe struct {
	dependencias DependenciasDeBoe
}

func (appletBoe) Nombre() string { return "boe" }

func (appletBoe) Descripcion() string {
	return "Consulta la legislación consolidada del BOE y la devuelve lista para citar."
}

// Verbos son los seis del contrato, en su orden, cada uno con sus argumentos
// posicionales, todos obligatorios, y el valor cero del tipo de su data
// (contrato puerto-y-applet §4).
func (a appletBoe) Verbos() []Verbo {
	return []Verbo{
		{
			Nombre:      "buscar",
			Descripcion: "Busca normas consolidadas por las palabras de su título o con una consulta de la fuente.",
			Argumentos:  func() Argumentos { return &argumentosDeBuscar{dependencias: a.dependencias} },
			Salida:      []boe.ResultadoDeBusqueda(nil),
		},
		{
			Nombre:      "indice",
			Descripcion: "Devuelve los bloques de una norma consolidada, en el orden de la fuente.",
			Argumentos:  a.deUnaNorma(consultaDelIndice),
			Salida:      boe.Indice{},
		},
		{
			Nombre:      "articulo",
			Descripcion: "Devuelve el texto vigente de un bloque de una norma, con los avisos de su vigencia.",
			Argumentos:  func() Argumentos { return &argumentosDeArticulo{dependencias: a.dependencias} },
			Salida:      boe.Articulo{},
		},
		{
			Nombre:      "articulos",
			Descripcion: "Devuelve el texto vigente de varios bloques de una norma, en el orden pedido.",
			Argumentos:  func() Argumentos { return &argumentosDeArticulos{dependencias: a.dependencias} },
			Salida:      []boe.Articulo(nil),
		},
		{
			Nombre:      "metadatos",
			Descripcion: "Devuelve los datos de una norma y los avisos de su vigencia.",
			Argumentos:  a.deUnaNorma(consultaDeLosMetadatos),
			Salida:      boe.Metadatos{},
		},
		{
			Nombre:      "analisis",
			Descripcion: "Devuelve las materias, las notas y las referencias de una norma.",
			Argumentos:  a.deUnaNorma(consultaDelAnalisis),
			Salida:      boe.Analisis{},
		},
	}
}

// deUnaNorma es la fábrica de argumentos de un verbo que consulta un recurso de
// una norma —indice, metadatos y analisis—, que solo se distinguen por la
// consulta que hacen con ella.
func (a appletBoe) deUnaNorma(consulta func(norma string) core.Consulta) func() Argumentos {
	return func() Argumentos {
		return &argumentosDeUnaNorma{consulta: consulta, dependencias: a.dependencias}
	}
}

// consultaDelIndice, consultaDeLosMetadatos y consultaDelAnalisis son las
// consultas de los verbos de una norma.
func consultaDelIndice(norma string) core.Consulta { return boe.ConsultaIndice{Norma: norma} }

func consultaDeLosMetadatos(norma string) core.Consulta { return boe.ConsultaMetadatos{Norma: norma} }

func consultaDelAnalisis(norma string) core.Consulta { return boe.ConsultaAnalisis{Norma: norma} }

// argumentosDeBuscar son los de buscar: una o más palabras, que la fuente une
// con un espacio (contrato verbos-y-salidas §1).
type argumentosDeBuscar struct {
	Texto []string `arg:"" name:"texto" help:"Palabras del título, o una consulta con AND, OR, NOT, titulo:, materia: o comillas."`

	dependencias DependenciasDeBoe
}

func (a *argumentosDeBuscar) Ejecutar(
	ctx context.Context, ec schema.Contexto, registrador *slog.Logger,
) (schema.Resultado, error) {
	return consultarBoe(ctx, ec, registrador, a.dependencias, boe.ConsultaBuscar{Texto: a.Texto})
}

// argumentosDeUnaNorma son los de indice, metadatos y analisis: la norma, y la
// consulta que el verbo hace con ella en un campo no exportado que la gramática
// no ve (contrato verbos-y-salidas §2, §5 y §6).
type argumentosDeUnaNorma struct {
	Norma string `arg:"" name:"norma" help:"Identificador de la norma en el BOE: BOE-A-<año>-<número>."`

	consulta     func(norma string) core.Consulta
	dependencias DependenciasDeBoe
}

func (a *argumentosDeUnaNorma) Ejecutar(
	ctx context.Context, ec schema.Contexto, registrador *slog.Logger,
) (schema.Resultado, error) {
	return consultarBoe(ctx, ec, registrador, a.dependencias, a.consulta(a.Norma))
}

// argumentosDeArticulo son los de articulo: la norma y el bloque (contrato
// verbos-y-salidas §3).
type argumentosDeArticulo struct {
	Norma  string `arg:"" name:"norma" help:"Identificador de la norma en el BOE: BOE-A-<año>-<número>."`
	Bloque string `arg:"" name:"bloque" help:"Identificador del bloque en el índice de la norma: a21, da-3…"`

	dependencias DependenciasDeBoe
}

func (a *argumentosDeArticulo) Ejecutar(
	ctx context.Context, ec schema.Contexto, registrador *slog.Logger,
) (schema.Resultado, error) {
	return consultarBoe(ctx, ec, registrador, a.dependencias, boe.ConsultaArticulo{Norma: a.Norma, Bloque: a.Bloque})
}

// argumentosDeArticulos son los de articulos: la norma y uno o más bloques, en
// el orden pedido y con repeticiones (contrato verbos-y-salidas §4).
type argumentosDeArticulos struct {
	Norma   string   `arg:"" name:"norma" help:"Identificador de la norma en el BOE: BOE-A-<año>-<número>."`
	Bloques []string `arg:"" name:"bloques" help:"Identificadores de los bloques, en el orden en que se quieren."`

	dependencias DependenciasDeBoe
}

func (a *argumentosDeArticulos) Ejecutar(
	ctx context.Context, ec schema.Contexto, registrador *slog.Logger,
) (schema.Resultado, error) {
	return consultarBoe(ctx, ec, registrador, a.dependencias, boe.ConsultaArticulos{Norma: a.Norma, Bloques: a.Bloques})
}

// Las comprobaciones en tiempo de compilación del contrato del applet.
var (
	_ Applet     = appletBoe{}
	_ Argumentos = (*argumentosDeBuscar)(nil)
	_ Argumentos = (*argumentosDeUnaNorma)(nil)
	_ Argumentos = (*argumentosDeArticulo)(nil)
	_ Argumentos = (*argumentosDeArticulos)(nil)
)

// consultarBoe compone la fuente del BOE para esta invocación, con el cliente y
// la caché de las dependencias y el registrador que el kernel entrega al applet,
// y devuelve tal cual lo que responde a la consulta: su resultado y su error,
// que ya lleva su clase y la procedencia de lo que falló (contrato
// puerto-y-applet §4). La fuente valida la consulta antes de abrir la caché, la
// abre en solo lectura con --offline o --dry-run, construye el cliente solo si
// pide algo y cierra lo que abrió antes de volver.
//
// Componerla solo falla por un defecto de composición —unas dependencias sin
// cliente—, y ese fallo es «inesperado» sin procedencia, que firma y fecha el
// kernel (contrato errores-y-codigos, fila 22).
func consultarBoe(
	ctx context.Context,
	ec schema.Contexto,
	registrador *slog.Logger,
	dependencias DependenciasDeBoe,
	consulta core.Consulta,
) (schema.Resultado, error) {
	fuente, err := boe.Nueva(
		boe.ConCliente(dependencias.construccionDelCliente(registrador)),
		boe.ConCache(dependencias.aperturaDeLaCache(registrador)),
		boe.ConRegistrador(registrador),
	)
	if err != nil {
		return schema.Resultado{}, err
	}

	return fuente.Fetch(ctx, ec, consulta)
}

// construccionDelCliente es la función de boe.ConCliente: construye el cliente
// con el registrador de la invocación. Sin Cliente no hay función, y la fuente
// lo rechaza al componerse como el defecto que es.
func (d DependenciasDeBoe) construccionDelCliente(registrador *slog.Logger) func() (boe.Pedidor, error) {
	if d.Cliente == nil {
		return nil
	}

	return func() (boe.Pedidor, error) {
		cliente, err := d.Cliente(registrador)
		if err != nil || cliente == nil {
			// Un *httpx.Cliente nulo dentro de la interfaz no sería nulo para la
			// fuente: se devuelve la interfaz nula, que la fuente reconoce como
			// la construcción que no da ni cliente ni error.
			return nil, err
		}

		return cliente, nil
	}
}

// aperturaDeLaCache es la función de boe.ConCache: abre la caché con las
// opciones de las dependencias, sus eventos en el registrador de la invocación
// y, cuando la fuente lo pide —con --offline o --dry-run—, en solo lectura
// (FR-092, FR-094). Cada apertura compone su propia lista de opciones, de modo
// que las de las dependencias no cambian nunca.
func (d DependenciasDeBoe) aperturaDeLaCache(registrador *slog.Logger) boe.AperturaDeCache {
	return func(ctx context.Context, soloLectura bool) (boe.CacheAbierta, error) {
		opciones := make([]cache.Opcion, 0, len(d.Cache)+2)
		opciones = append(opciones, d.Cache...)
		opciones = append(opciones, cache.ConRegistrador(registrador))

		if soloLectura {
			opciones = append(opciones, cache.SoloLectura())
		}

		abierta, err := cache.New(ctx, opciones...)
		if err != nil {
			return nil, err
		}

		return abierta, nil
	}
}
