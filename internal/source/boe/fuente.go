package boe

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jmorenobl/kitlegal/internal/core"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// NombreDeLaFuente es el nombre de la fuente de este adaptador, la API de
// Legislación Consolidada del BOE: el que va en la clave fuente del sobre y el
// que identifica sus grabaciones al grabarlas y al reproducirlas. No empieza por
// «kitlegal.», el espacio reservado de lo que se obtiene sin consultar ninguna
// fuente pública (FR-002, ADR 0006).
const NombreDeLaFuente = "boe.legislacion-consolidada"

// Las vigencias con las que se guarda lo que responde cada consulta: las de
// CACHE_TTL en refs/boe.py, líneas 37-43 (FR-091).
const (
	// vigenciaCorta es la de buscar y metadatos, cinco minutos: lo que puede
	// cambiar en cualquier publicación del BOE.
	vigenciaCorta = 300 * time.Second
	// vigenciaLarga es la de indice, articulo —también la de cada bloque de
	// articulos— y analisis, siete días: la estructura y el texto consolidado de
	// una norma, y sus referencias, solo cambian con una reforma.
	vigenciaLarga = 604_800 * time.Second
)

// Los mensajes de los fallos que produce este fichero (contrato
// errores-y-codigos, filas 21 y 22).
const (
	// motivoDelDefectoAlComponer es el de la fuente a la que le falta o se le da
	// mal una dependencia, que es un defecto de quien la compone.
	motivoDelDefectoAlComponer = "la fuente del BOE no se puede componer: %s"
	// motivoDeLaConsultaDeOtroTipo y motivoSinConsulta son los de la consulta que
	// la fuente no resuelve.
	motivoDeLaConsultaDeOtroTipo = "la fuente del BOE no resuelve consultas de tipo %T"
	motivoSinConsulta            = "la fuente del BOE no ha recibido ninguna consulta"
	// motivoAlAbrirLaCache y motivoAlCerrarLaCache son los de la caché que no se
	// deja abrir o cerrar; la dirección y el texto de la caché van detrás.
	motivoAlAbrirLaCache  = "no se ha podido abrir la caché"
	motivoAlCerrarLaCache = "no se ha podido cerrar la caché"
)

// CacheAbierta es la caché de una invocación: el puerto del dominio y el cierre.
// El puerto core.Cache no tiene Close para que nadie cierre una caché que no
// abrió; la fuente sí abre la suya, con la AperturaDeCache de ConCache, y por eso
// declara aquí lo que necesita para cerrarla. En producción es *cache.Cliente.
type CacheAbierta interface {
	core.Cache
	Close() error
}

// AperturaDeCache abre la caché de una invocación, en solo lectura si
// soloLectura, que es como la pide la fuente con --offline o con --dry-run
// (FR-092, FR-094; research.md D5).
type AperturaDeCache func(ctx context.Context, soloLectura bool) (CacheAbierta, error)

// ajustes recoge lo que declaran las opciones antes de componer la fuente. Es
// privado: las opciones son el único modo de tocarlo, y Nueva comprueba que no
// falte ninguna dependencia antes de construir.
type ajustes struct {
	construirCliente func() (Pedidor, error)
	abrirCache       AperturaDeCache
	registrador      *slog.Logger
}

// Opcion declara una dependencia de la fuente antes de componerla. Devuelve
// error porque la comprobación ocurre al componer: Nueva las aplica en su orden
// y devuelve el error de la primera que no vale.
type Opcion func(*ajustes) error

// ConCliente declara cómo se construye el Pedidor de la fuente, que en
// producción es un *httpx.Cliente. La fuente llama a construir solo si una
// invocación tiene que pedir algo, y como mucho una vez por invocación: lo que se
// sirve de la caché no construye ningún cliente (research.md D5). Una función
// nula es un defecto al componer.
func ConCliente(construir func() (Pedidor, error)) Opcion {
	return func(a *ajustes) error {
		if construir == nil {
			return defectoAlComponer(nil, "la opción ConCliente no lleva ninguna función que construya el cliente")
		}

		a.construirCliente = construir

		return nil
	}
}

// ConCache declara cómo se abre la caché de la fuente. La fuente llama a abrir
// una vez por invocación, después de validar la consulta —una invocación mal
// formada no abre nada—, en solo lectura con --offline o con --dry-run, y cierra
// lo que devuelve antes de volver de Fetch (FR-092, FR-094; research.md D5). Una
// función nula es un defecto al componer.
func ConCache(abrir AperturaDeCache) Opcion {
	return func(a *ajustes) error {
		if abrir == nil {
			return defectoAlComponer(nil, "la opción ConCache no lleva ninguna función que abra la caché")
		}

		a.abrirCache = abrir

		return nil
	}
}

// ConRegistrador declara a dónde van los eventos de la fuente, que es el mismo
// registrador que el kernel entrega al applet; sin la opción se descartan. El
// dominio no importa log/slog, así que el registrador no viaja por el puerto: se
// da al componer la fuente, una por invocación (research.md D2). Un registrador
// nulo es un defecto al componer: quien lo pasa quería registrar algo y en
// silencio no registraría nada.
func ConRegistrador(registrador *slog.Logger) Opcion {
	return func(a *ajustes) error {
		if registrador == nil {
			return defectoAlComponer(nil, "la opción ConRegistrador no lleva ningún registrador")
		}

		a.registrador = registrador

		return nil
	}
}

// Fuente es el adaptador de la API de Legislación Consolidada del BOE e
// implementa el puerto core.Source (FR-121). Se compone con Nueva, una por
// invocación, con sus dependencias declaradas como funciones que la fuente llama
// cuando las necesita, de modo que componerla no pide nada ni abre nada.
type Fuente struct {
	// construirCliente es la función de ConCliente.
	construirCliente func() (Pedidor, error)
	// abrirCache es la función de ConCache.
	abrirCache AperturaDeCache
	// registrador nunca es nulo: descarta por omisión.
	registrador *slog.Logger
}

// Fuente implementa el puerto del dominio, y que lo siga implementando no
// depende de que alguien lo recuerde.
var _ core.Source = (*Fuente)(nil)

// Nueva compone la fuente con sus dependencias. ConCliente y ConCache son
// obligatorias y ConRegistrador no. Sin alguna de las obligatorias, o con una
// opción que no vale, no devuelve ninguna fuente y falla con «inesperado»: es un
// defecto de quien la compone, nunca de quien invoca, y el error no lleva
// dirección ni instante, de modo que el sobre lo firma y lo fecha el kernel
// (contrato puerto-y-applet §3.1; errores-y-codigos, fila 22). Componer no llama
// a ninguna de las dependencias.
func Nueva(opciones ...Opcion) (*Fuente, error) {
	declarados := ajustes{registrador: slog.New(slog.DiscardHandler)}

	for posicion, opcion := range opciones {
		if opcion == nil {
			return nil, defectoAlComponer(nil, fmt.Sprintf("la opción %d de %d es nula", posicion+1, len(opciones)))
		}

		if err := opcion(&declarados); err != nil {
			return nil, err
		}
	}

	if ausentes := dependenciasAusentes(declarados); len(ausentes) > 0 {
		motivo := "falta la dependencia " + ausentes[0]
		if len(ausentes) > 1 {
			motivo = "faltan las dependencias " + strings.Join(ausentes, " y ")
		}

		return nil, defectoAlComponer(nil, motivo)
	}

	return &Fuente{
		construirCliente: declarados.construirCliente,
		abrirCache:       declarados.abrirCache,
		registrador:      declarados.registrador,
	}, nil
}

// dependenciasAusentes son las opciones obligatorias que no se declararon, en el
// orden del contrato.
func dependenciasAusentes(declarados ajustes) []string {
	var ausentes []string

	if declarados.construirCliente == nil {
		ausentes = append(ausentes, "ConCliente")
	}

	if declarados.abrirCache == nil {
		ausentes = append(ausentes, "ConCache")
	}

	return ausentes
}

// Name es el nombre de la fuente, NombreDeLaFuente.
func (*Fuente) Name() string {
	return NombreDeLaFuente
}

// TTL es la vigencia con la que se guarda lo que responde la consulta, que
// depende de su verbo y no de sus argumentos (FR-091): cinco minutos para buscar
// y metadatos, siete días para indice, articulo, articulos y analisis. Una
// consulta que la fuente no declara no tiene vigencia, porque de ella nunca se
// guarda nada.
func (*Fuente) TTL(consulta core.Consulta) time.Duration {
	switch consulta.(type) {
	case ConsultaBuscar, ConsultaMetadatos:
		return vigenciaCorta
	case ConsultaIndice, ConsultaArticulo, ConsultaArticulos, ConsultaAnalisis:
		return vigenciaLarga
	default:
		return 0
	}
}

// Terms son los términos de uso de la fuente y el día en que una persona los
// revisó, los de terminos.go (FR-121).
func (*Fuente) Terms() core.Terminos {
	return terminosDeUso
}

// Fetch resuelve la consulta (core.Source). Cada verbo resuelve la suya en su
// propio fichero, con un caso en Fetch que valida la consulta antes de abrir
// nada y la resuelve dentro de invocar (research.md D2 y D5). Una consulta sin
// caso —de un tipo que la fuente no declara, nula o de un verbo que la fuente
// todavía no resuelve— es un defecto de quien la compone: «inesperado», sin
// procedencia, porque no se ha consultado nada, y sin abrir la caché ni
// construir el cliente, también con --offline y con --dry-run (contrato
// errores-y-codigos, fila 22).
func (*Fuente) Fetch(_ context.Context, _ schema.Contexto, consulta core.Consulta) (schema.Resultado, error) {
	return schema.Resultado{}, errorDeConsultaSinCaso(consulta)
}

// errorDeConsultaSinCaso es el fallo de la consulta que la fuente no resuelve,
// que nombra su tipo. No llama a Verbo: una consulta que la fuente no declara no
// ofrece ninguna garantía de que hacerlo no entre en pánico.
func errorDeConsultaSinCaso(consulta core.Consulta) *Error {
	if consulta == nil {
		return errorInesperado("", time.Time{}, nil, motivoSinConsulta)
	}

	return errorInesperado("", time.Time{}, nil, fmt.Sprintf(motivoDeLaConsultaDeOtroTipo, consulta))
}

// invocacion es lo que dura una llamada a Fetch con la consulta ya validada: el
// contexto de ejecución, la caché abierta en el modo que ese contexto pide y el
// Pedidor, que no existe hasta que hace falta pedir algo.
type invocacion struct {
	// ec es el contexto de ejecución del kernel, con --offline y --dry-run.
	ec schema.Contexto
	// cache es la caché de la invocación, que abre y cierra invocar.
	cache core.Cache
	// construirCliente es la función de ConCliente, y cliente, lo que devolvió
	// la primera vez que se llamó.
	construirCliente func() (Pedidor, error)
	cliente          Pedidor
}

// pedidor es el Pedidor de la invocación: la primera vez lo construye con la
// función de ConCliente y después devuelve el mismo, de modo que una invocación
// construye como mucho un cliente y solo si pide algo (research.md D5). Que no se
// pueda construir es un defecto al componer, con el error de la construcción
// como causa.
func (en *invocacion) pedidor() (Pedidor, error) {
	if en.cliente != nil {
		return en.cliente, nil
	}

	cliente, err := en.construirCliente()

	switch {
	case err != nil:
		return nil, defectoAlComponer(err, "no se ha podido construir el cliente HTTP")
	case cliente == nil:
		return nil, defectoAlComponer(nil, "la función de ConCliente no ha devuelto ni cliente ni error")
	}

	en.cliente = cliente

	return cliente, nil
}

// invocar abre la caché de la invocación, resuelve con ella y la cierra antes de
// volver (contrato puerto-y-applet §3.1; research.md D5). Cada verbo lo llama
// después de validar su consulta, con la dirección del recurso que consulta, que
// es la que nombra cualquier fallo de la caché (contrato errores-y-codigos,
// fila 21):
//
//   - la caché se abre en solo lectura con --offline o con --dry-run, que
//     prometen no crear ni cambiar nada, y en modo normal en otro caso (FR-092,
//     FR-094);
//   - si no se puede abrir, no se resuelve nada y el fallo lleva la procedencia
//     del recurso sin fecha, porque no hubo petición;
//   - si solo falla el cierre, la invocación falla con la clase del cierre y el
//     resultado pierde sus datos y su fecha, pero no la descripción del ensayo,
//     que el kernel presenta también en fallo;
//   - y si fallan la resolución y el cierre, prevalece el resultado de la
//     resolución y los dos errores quedan unidos con errors.Join, de modo que
//     ninguno se pierde y la clase es la de lo que falló primero.
//
// Una fuente que no se compuso con Nueva, o una apertura que no devuelve caché
// ni error, es un defecto al componer, y nada entra en pánico.
func (f *Fuente) invocar(ctx context.Context, ec schema.Contexto, direccion string,
	resolver func(ctx context.Context, en *invocacion) (schema.Resultado, error),
) (schema.Resultado, error) {
	if f == nil || f.construirCliente == nil || f.abrirCache == nil {
		return schema.Resultado{}, defectoAlComponer(nil, "la fuente no se ha construido con Nueva")
	}

	abierta, err := f.abrirCache(ctx, ec.Offline || ec.DryRun)

	switch {
	case err != nil:
		fallo := falloDeLaCache(direccion, motivoAlAbrirLaCache, err)

		return resultadoDelFallo(fallo), fallo
	case abierta == nil:
		return schema.Resultado{}, defectoAlComponer(nil, "la función de ConCache no ha devuelto ni caché ni error")
	}

	resultado, err := resolver(ctx, &invocacion{ec: ec, cache: abierta, construirCliente: f.construirCliente})

	errDelCierre := abierta.Close()
	if errDelCierre == nil {
		return resultado, err
	}

	cierre := falloDeLaCache(direccion, motivoAlCerrarLaCache, errDelCierre)
	if err != nil {
		return resultado, errors.Join(err, cierre)
	}

	fallido := resultadoDelFallo(cierre)
	fallido.Ensayo = resultado.Ensayo

	return fallido, cierre
}

// falloDeLaCache es el error de la caché visto desde la fuente: con la dirección
// del recurso consultado, sin instante, porque la caché no pide nada y el sobre
// lo fecha el montaje, y con la clase de claseDelFalloDeLaCache (contrato
// errores-y-codigos, fila 21). El error queda como causa: alcanzable con
// errors.Is y, si declara su clase, con su texto en el mensaje.
func falloDeLaCache(direccion, motivo string, causa error) *Error {
	return nuevoError(claseDelFalloDeLaCache(causa), direccion, time.Time{}, causa, motivo)
}

// claseDelFalloDeLaCache es la clase con la que la fuente entrega un fallo de su
// caché: la que el error declara si es una de las tres que produce cache.Error
// —argumentos, fuente no disponible e inesperado—, e «inesperado» en cualquier
// otro caso, de modo que ningún fallo de la caché queda sin clase ni termina con
// el código 6 (FR-100).
func claseDelFalloDeLaCache(causa error) schema.Clase {
	var conClase schema.ConClase
	if !errors.As(causa, &conClase) {
		return schema.ClaseInesperado
	}

	switch clase := conClase.Clase(); clase {
	case schema.ClaseArgumentos, schema.ClaseFuenteNoDisponible:
		return clase
	default:
		return schema.ClaseInesperado
	}
}

// resultadoDelFallo es el Resultado con el que la fuente acompaña un fallo: la
// procedencia de lo que falló —la fuente, la dirección y, si hubo petición, su
// instante— y nada más (contrato puerto-y-applet §1).
func resultadoDelFallo(fallo *Error) schema.Resultado {
	return schema.Resultado{Procedencia: schema.Procedencia{
		Fuente:        NombreDeLaFuente,
		URL:           fallo.URL,
		FechaConsulta: fallo.Instante,
	}}
}

// defectoAlComponer es el fallo de quien compone la fuente: «inesperado», sin
// dirección ni instante, con el motivo detrás de que la fuente no se puede
// componer (contrato errores-y-codigos, fila 22).
func defectoAlComponer(causa error, motivo string) *Error {
	return errorInesperado("", time.Time{}, causa, fmt.Sprintf(motivoDelDefectoAlComponer, motivo))
}
