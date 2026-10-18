package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/jmorenobl/kitlegal/internal/core/plazos"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/core/territorio"
)

// La procedencia con la que firma plazos. Es un applet calculado: no consulta
// ninguna fuente en ejecución, sino los ficheros congelados que viajan en el
// binario, así que firma en el espacio de nombres reservado, y cada festivo que
// salta lleva en su source la disposición que lo fija (H9 FR-013, FR-028;
// research.md D11 de H9; ADR 0006).
const (
	fuenteDePlazos = "kitlegal.plazos"
	urlDePlazos    = "kitlegal:applet/plazos"
)

// procedenciaDePlazos es la firma de todo sobre del applet, de éxito o de
// fallo. No declara fecha: la pone el reloj del kernel, porque el plazo no
// depende de cuándo se consultó ningún fichero (contracts/applet-plazos.md §3
// de H9).
var procedenciaDePlazos = schema.Procedencia{Fuente: fuenteDePlazos, URL: urlDePlazos}

// FicherosDePlazos son los ficheros con los que se compone el applet plazos,
// tal como están escritos: los de data/territorio/, con los que resuelve los
// municipios, data/plazos.yaml y los dos de data/festivos/ (research.md D1 de
// H9). La raíz de producción y el binario de e2e le pasan los del binario
// (FuentesDePlazos), que no dependen del entorno.
type FicherosDePlazos struct {
	// Territorio son los de data/territorio/, como los recibe
	// AppletTerritorio.
	Territorio territorio.Fuentes
	// Tipos es data/plazos.yaml.
	Tipos []byte
	// Nacionales es data/festivos/nacionales-y-autonomicos.yaml.
	Nacionales []byte
	// Locales son los de data/festivos/locales/, por el código de la comunidad
	// que les da nombre.
	Locales map[string][]byte
}

// AppletPlazos es el applet que calcula la fecha de un plazo administrativo o
// procesal, con un único verbo, calcular, que hay que nombrar (H9 FR-001;
// contracts/applet-plazos.md §1). Recibe los ficheros ya leídos.
//
// Componerlo no analiza nada. Las fuentes se analizan una sola vez por applet,
// la primera vez que se ejecuta el verbo, y lo que resulte —lo cargado o su
// defecto— sirve a todas las invocaciones siguientes: la ayuda y --describe no
// lo pagan, y el análisis vive en el valor del applet y no en el paquete, así
// que dos applets con fuentes distintas no comparten nada.
func AppletPlazos(fuentes FicherosDePlazos) Applet {
	return appletPlazos{
		cargado: sync.OnceValues(func() (*plazosCargados, error) {
			return cargarPlazos(fuentes)
		}),
	}
}

// plazosCargados es lo que el applet analiza de sus fuentes: el registro de
// territorio, los tipos y el calendario de festivos.
type plazosCargados struct {
	registro   *territorio.Registro
	tipos      []plazos.Tipo
	calendario *plazos.Calendario
}

// cargarPlazos analiza los ficheros del applet. Unos que no cargan son un
// defecto de composición y no algo que quien pregunta pueda corregir: el error
// nombra cada fichero que falla y sale como inesperado, sin procedencia, que
// firma y fecha el kernel (contracts/applet-plazos.md §6 de H9). Incorpora los
// defectos como texto y **no** con %w a propósito, como territorio: el código
// de salida tiene que ser el del fallo inesperado sea cual sea el error que lo
// causó, y envolverlo dejaría que una clase de usuario de su cadena —la de un
// código mal formado en un fichero de territorio— decidiera otro.
func cargarPlazos(fuentes FicherosDePlazos) (*plazosCargados, error) {
	registro, errDelTerritorio := territorio.Cargar(fuentes.Territorio)
	tipos, errDeLosTipos := plazos.LeerTipos(fuentes.Tipos)
	calendario, errDelCalendario := plazos.CargarCalendario(fuentes.Nacionales, fuentes.Locales)

	if err := errors.Join(errDelTerritorio, errDeLosTipos, errDelCalendario); err != nil {
		return nil, fmt.Errorf("plazos: los ficheros de data/ con los que se compuso el applet no cargan: %s",
			err.Error())
	}

	return &plazosCargados{registro: registro, tipos: tipos, calendario: calendario}, nil
}

// appletPlazos declara lo que declara un applet y nada más: su nombre, su
// línea de ayuda y su verbo, que lleva consigo el análisis de las fuentes.
type appletPlazos struct {
	// cargado analiza las fuentes la primera vez que se le llama y devuelve lo
	// mismo en todas las siguientes.
	cargado func() (*plazosCargados, error)
}

func (appletPlazos) Nombre() string { return "plazos" }

func (appletPlazos) Descripcion() string {
	return "Calcula la fecha de un plazo administrativo o procesal con los festivos de los municipios que cuentan," +
		" y declara hasta dónde llegan los festivos que tiene."
}

// Verbos es calcular, sin ningún otro y sin verbo por omisión, con sus
// argumentos y el valor cero del plazo calculado (contracts/applet-plazos.md
// §1 y §4 de H9).
func (a appletPlazos) Verbos() []Verbo {
	return []Verbo{{
		Nombre: "calcular",
		Descripcion: "Devuelve la fecha de un plazo —un recurso, el acceso a la información pública, el silencio o uno" +
			" genérico— desde la fecha de inicio que se da, con los preceptos que lo fijan, los festivos saltados," +
			" los avisos y la cobertura de los festivos.",
		Argumentos: func() Argumentos { return &argumentosDeCalcular{cargado: a.cargado} },
		Salida:     plazos.Plazo{},
	}}
}

// argumentosDeCalcular son los de calcular: el tipo, la fecha de inicio y los
// municipios, por su posición, y la duración y su unidad, como banderas. Todos
// son cadenas: qué vale lo decide el dominio, con su clase, y no la gramática,
// que con un enumerado o un entero rompería la tabla de comandos generada
// (research.md D2 y V12 de H9). Ninguna otra bandera propia: el applet hereda
// las ocho globales.
type argumentosDeCalcular struct {
	Tipo       string   `arg:"" name:"tipo" help:"Tipo de plazo: alzada, reposicion, contencioso, acceso-resolver, acceso-reclamar, silencio o generico."`
	Inicio     string   `arg:"" name:"inicio" help:"Fecha de inicio, AAAA-MM-DD: la de la notificación o la publicación; en silencio, la del acuerdo de iniciación o la de entrada de la solicitud."`
	Municipios []string `arg:"" name:"municipios" help:"Municipios que cuentan, por su nombre o su código INE: el del interesado y el de la sede del órgano, o en contencioso el de la sede del juzgado."`
	Duracion   string   `placeholder:"<n>" help:"Duración, un entero positivo; solo, y siempre, en silencio y generico."`
	Unidad     string   `placeholder:"<unidad>" help:"Unidad de la duración: meses, dias-habiles o dias-naturales; solo, y siempre, en silencio y generico."`

	cargado func() (*plazosCargados, error)
}

// Ejecutar calcula el plazo pedido. No abre ninguna conexión ni toca la caché,
// así que --offline no cambia nada, y no describe ninguna operación con
// --dry-run, porque no tiene ninguna capa con efectos; tampoco observa nada del
// mundo, así que no lleva ninguna operación de grafo (H9 FR-013; research.md
// D11 de H9).
//
// Comprueba primero, con el dominio, lo que no depende de ningún municipio —el
// tipo, la fecha, la duración, la unidad y cuántos municipios son—; después
// resuelve cada municipio con territorio, y solo entonces calcula: un argumento
// que no vale sale con 2 aunque algún municipio no exista (FR-012; research.md
// D10 de H9).
func (a *argumentosDeCalcular) Ejecutar(
	_ context.Context, _ schema.Contexto, _ *slog.Logger,
) (schema.Resultado, error) {
	cargado, err := a.cargado()
	if err != nil {
		return schema.Resultado{}, err
	}

	peticion := plazos.Peticion{
		Tipo: a.Tipo, Inicio: a.Inicio, Duracion: a.Duracion, Unidad: a.Unidad,
		Municipios: make([]plazos.MunicipioResuelto, len(a.Municipios)),
	}

	if err := plazos.ComprobarPeticion(cargado.tipos, peticion); err != nil {
		return falloDePlazos(err)
	}

	peticion.Municipios, err = resolverMunicipios(cargado.registro, a.Municipios)
	if err != nil {
		return falloDePlazos(err)
	}

	plazo, err := plazos.Calcular(cargado.calendario, cargado.tipos, peticion)
	if err != nil {
		return falloDePlazos(err)
	}

	return schema.Resultado{Procedencia: procedenciaDePlazos, Datos: plazo}, nil
}

// falloDePlazos es el resultado de un fallo que decide el applet, con su firma.
func falloDePlazos(err error) (schema.Resultado, error) {
	return schema.Resultado{Procedencia: procedenciaDePlazos}, err
}

// resolverMunicipios resuelve cada municipio pedido con territorio, como
// territorio resolver, y da los que cuentan en el orden dado. Si alguno no se
// resuelve, ninguno cuenta: los que territorio daría con 2 —un nombre de varios
// municipios, una entrada que no llega a ser consulta— deciden el fallo, con su
// clase, antes que los que no están en la relación, que solo dan el 3 si todos
// los demás valen (FR-012; research.md D10 y V17 de H9). El mensaje lleva el de
// territorio de cada uno de los que deciden, tras «municipio «<pedido>»: »,
// separados por « | ».
//
// Territorio no falla de ninguna otra forma (H6 FR-016): otro fallo es un
// defecto, que sale como inesperado.
func resolverMunicipios(registro *territorio.Registro, pedidos []string) ([]plazos.MunicipioResuelto, error) {
	resueltos := make([]plazos.MunicipioResuelto, 0, len(pedidos))
	rechazos := map[schema.Clase][]string{}

	for _, pedido := range pedidos {
		resuelto, err := registro.Resolver(pedido)
		if err == nil {
			resueltos = append(resueltos, municipioQueCuenta(resuelto))

			continue
		}

		var conClase schema.ConClase
		if !errors.As(err, &conClase) ||
			(conClase.Clase() != schema.ClaseArgumentos && conClase.Clase() != schema.ClaseNoEncontrado) {
			return nil, fmt.Errorf("plazos: territorio no resuelve el municipio «%s» por un defecto: %s", pedido,
				err.Error())
		}

		rechazos[conClase.Clase()] = append(rechazos[conClase.Clase()],
			fmt.Sprintf("municipio «%s»: %s", pedido, err.Error()))
	}

	for _, clase := range []schema.Clase{schema.ClaseArgumentos, schema.ClaseNoEncontrado} {
		if mensajes := rechazos[clase]; len(mensajes) > 0 {
			return nil, &errorDeMunicipios{clase: clase, mensaje: strings.Join(mensajes, " | ")}
		}
	}

	return resueltos, nil
}

// municipioQueCuenta es el municipio de un territorio resuelto como lo recibe
// el dominio: su código, su nombre, el código de su comunidad y el nombre de
// esta, que nombran los avisos.
func municipioQueCuenta(resuelto territorio.Territorio) plazos.MunicipioResuelto {
	return plazos.MunicipioResuelto{
		Municipio: plazos.Municipio{
			CodigoINE: resuelto.CodigoINE.Codigo,
			Nombre:    resuelto.Municipio.Nombre,
			Comunidad: resuelto.Comunidad.Codigo,
		},
		NombreDeLaComunidad: resuelto.Comunidad.Nombre,
	}
}

// errorDeMunicipios es el rechazo de los municipios pedidos que territorio no
// resuelve: de clase argumentos o no encontrado, como el de territorio, con el
// mensaje de cada uno. Declara su clase él mismo, como los del dominio.
type errorDeMunicipios struct {
	clase   schema.Clase
	mensaje string
}

// Error devuelve el mensaje, que llega literal al sobre de fallo y a la salida
// de error.
func (e *errorDeMunicipios) Error() string {
	return e.mensaje
}

// Clase es la del fallo, que el kernel traduce a código 2 o 3.
func (e *errorDeMunicipios) Clase() schema.Clase {
	return e.clase
}

// Las comprobaciones en tiempo de compilación del contrato del applet.
var (
	_ Applet          = appletPlazos{}
	_ Argumentos      = (*argumentosDeCalcular)(nil)
	_ schema.ConClase = (*errorDeMunicipios)(nil)
)
