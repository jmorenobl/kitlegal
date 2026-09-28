package app

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/core/territorio"
)

// La procedencia con la que firma territorio. Es un applet calculado: no
// consulta ninguna fuente en ejecución, sino los ficheros congelados que viajan
// en el binario, así que firma en el espacio de nombres reservado y no con el
// identificador de una fila de docs/SOURCES.md, que va, dato a dato, en el
// source de cada uno (FR-004, FR-005; ADR 0006).
const (
	fuenteDeTerritorio = "kitlegal.territorio"
	urlDeTerritorio    = "kitlegal:applet/territorio"
)

// AppletTerritorio es el applet que resuelve un municipio de España al terreno
// que lo rodea, con un único verbo, resolver, que hay que nombrar (FR-001,
// contrato del applet §1). Recibe los ficheros congelados de data/territorio/
// ya leídos: la raíz de producción y el binario de e2e le pasan los embebidos,
// que no dependen del entorno (contrato del applet §7).
//
// Componerlo no analiza nada. Las fuentes se analizan una sola vez por applet,
// la primera vez que se ejecuta el verbo, y lo que resulte —el registro o su
// defecto— sirve a todas las invocaciones siguientes: la ayuda y --describe no
// lo pagan, y el análisis vive en el valor del applet y no en el paquete, así
// que dos applets con fuentes distintas no comparten nada.
func AppletTerritorio(fuentes territorio.Fuentes) Applet {
	return appletTerritorio{
		registro: sync.OnceValues(func() (*territorio.Registro, error) {
			return territorio.Cargar(fuentes)
		}),
	}
}

// appletTerritorio declara lo que declara un applet y nada más: su nombre, su
// línea de ayuda y su verbo, que lleva consigo el análisis de las fuentes.
type appletTerritorio struct {
	// registro analiza las fuentes la primera vez que se le llama y devuelve lo
	// mismo en todas las siguientes.
	registro func() (*territorio.Registro, error)
}

func (appletTerritorio) Nombre() string { return "territorio" }

func (appletTerritorio) Descripcion() string {
	return "Resuelve un municipio de España a su provincia, su comunidad, su régimen, el DIR3 de su ayuntamiento" +
		" y sus boletines."
}

// Verbos es resolver, sin ningún otro y sin verbo por omisión, con su
// argumento, que va por su posición, y el valor cero del territorio resuelto
// (contrato del applet §1 y §3).
func (a appletTerritorio) Verbos() []Verbo {
	return []Verbo{{
		Nombre: "resolver",
		Descripcion: "Devuelve el territorio de un municipio, por su nombre o por su código INE, con la cobertura" +
			" de lo que está configurado y verificado.",
		Argumentos: func() Argumentos { return &argumentosDeResolver{registro: a.registro} },
		Salida:     territorio.Territorio{},
	}}
}

// argumentosDeResolver son los de resolver: la consulta, obligatoria, y el
// análisis de las fuentes en un campo no exportado que la gramática no ve
// (FR-002). Ninguna bandera propia: el applet hereda las ocho globales.
type argumentosDeResolver struct {
	Consulta string `arg:"" name:"consulta" help:"Nombre del municipio, o su código INE: cinco cifras, o seis con el dígito de control."`

	registro func() (*territorio.Registro, error)
}

// Ejecutar resuelve la consulta en el registro. No abre ninguna conexión ni
// toca la caché, así que --offline no cambia nada, y no describe ninguna
// operación con --dry-run, porque no tiene ninguna capa con efectos (FR-009,
// research.md D7).
//
// La respuesta la firma el applet con la fecha más antigua de los ficheros que
// la sostienen, y el fallo que decide él —un nombre de varios municipios, algo
// que no está en la relación, una entrada que no es un código—, con la de la
// relación, que es la que lo decide: la salida no depende del reloj (contrato
// del applet §2; data-model §2.8). Ese fallo lleva su clase, «argumentos» o
// «no encontrado», y su mensaje nombra la entrada (FR-010 a FR-016).
//
// El resultado de éxito lleva además lo que observa del mundo el territorio
// resuelto, que el kernel entrega al grafo del mundo con la procedencia del
// sobre; un fallo no observa nada (contracts/emision.md §2).
//
// Unas fuentes que no cargan son un defecto de composición y no algo que quien
// pregunta pueda corregir: salen como inesperado, sin procedencia, que firma y
// fecha el kernel (contrato del applet §7).
func (a *argumentosDeResolver) Ejecutar(
	_ context.Context, _ schema.Contexto, _ *slog.Logger,
) (schema.Resultado, error) {
	registro, err := a.registro()
	if err != nil {
		return schema.Resultado{}, defectoDeComposicion(err)
	}

	resuelto, err := registro.Resolver(a.Consulta)
	if err != nil {
		return schema.Resultado{Procedencia: procedenciaDeTerritorio(registro.FechaDeLaRelacion())}, err
	}

	return schema.Resultado{
		Procedencia: procedenciaDeTerritorio(resuelto.Fecha()),
		Datos:       resuelto,
		Grafo:       resuelto.Observado(),
	}, nil
}

// procedenciaDeTerritorio es la firma del applet con la fecha de los ficheros
// que sostienen la respuesta.
func procedenciaDeTerritorio(fecha time.Time) schema.Procedencia {
	return schema.Procedencia{Fuente: fuenteDeTerritorio, URL: urlDeTerritorio, FechaConsulta: fecha}
}

// defectoDeComposicion es el fallo de unas fuentes que no cargan. Incorpora el
// defecto como texto y **no** con %w a propósito: el código de salida tiene que
// ser el del fallo inesperado sea cual sea el error que lo causó, y envolverlo
// dejaría que una clase de usuario de su cadena decidiera otro.
func defectoDeComposicion(err error) error {
	return fmt.Errorf("territorio: los ficheros de data/territorio/ con los que se compuso el applet no cargan: %s",
		err.Error())
}

// Las comprobaciones en tiempo de compilación del contrato del applet.
var (
	_ Applet     = appletTerritorio{}
	_ Argumentos = (*argumentosDeResolver)(nil)
)
