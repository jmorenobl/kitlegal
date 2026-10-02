package empaquetado

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/jmorenobl/kitlegal"
	"github.com/jmorenobl/kitlegal/internal/app"
)

// Los dos códigos con los que termina el paso: quien lo ejecuta —goreleaser, un
// paso de un flujo— solo distingue 0 de lo demás. La tabla de códigos del
// binario es de su contrato de resultados, y el paso no es el binario
// (research.md D14 de H22).
const (
	codigoHecho = 0
	codigoFallo = 1
)

// Las dos órdenes del paso (contracts/paso.md §1 de H22).
const (
	ordenPiezas   = "piezas"
	ordenCatalogo = "catalogo"
)

// Las banderas de las dos órdenes, todas obligatorias y todas con un valor.
const (
	banderaVersion = "version"
	banderaMacOS   = "macos"
	banderaWindows = "windows"
	banderaIcono   = "icono"
	banderaHuella  = "sha256"
	banderaSalida  = "salida"
)

// prefijoDeLaLinea es con lo que empieza la línea de todo fallo: el nombre del
// programa, para quien la lee en el registro de goreleaser o de un flujo.
const prefijoDeLaLinea = "empaquetar: "

// errUso es el fallo de una invocación que no tiene la forma de ninguna de las
// dos órdenes: sin orden, con una que no existe, con una bandera que la orden
// no tiene o que llega sin su valor, o con un argumento de más.
var errUso = errors.New("uso: empaquetar piezas -version … -macos … -windows … -icono … -salida … | " +
	"empaquetar catalogo -version … -sha256 … -salida …")

// Ejecutar atiende la línea de órdenes del paso que empaqueta y devuelve su
// código (contracts/paso.md §1 de H22):
//
//	empaquetar piezas   -version <versión> -macos <binario> -windows <binario> -icono <png> -salida <carpeta>
//	empaquetar catalogo -version <versión> -sha256 <huella> -salida <fichero>
//
// piezas escribe en la carpeta kitlegal.mcpb y kitlegal-plugin.zip, con las
// herramientas que el registro de producción anuncia por MCP y las skills
// empotradas, las dos del mismo árbol del que se compila el paso; catalogo
// escribe en el fichero el marketplace.json de esa versión.
//
// Devuelve 0 si escribe lo pedido y 1 con cualquier fallo, también con una
// invocación que no vale. En un fallo escribe en errores una sola línea, que
// empieza por «empaquetar: » y nombra lo que falta o lo que falló (H22 FR-005).
// No escribe nada en la salida estándar, no pide nada a la red y no termina
// el proceso: eso lo hace cmd/empaquetar con el código que recibe.
//
// El error de escribir esa línea no se propaga, y es una decisión y no un
// silencio: el código ya dice que el paso falló, que es lo que lee quien lo
// ejecuta, y contarlo iría a la misma salida que acaba de fallar.
func Ejecutar(args []string, errores io.Writer) int {
	if err := ejecutar(args); err != nil {
		_, _ = fmt.Fprintf(errores, "%s%v\n", prefijoDeLaLinea, err)

		return codigoFallo
	}

	return codigoHecho
}

// ejecutar atiende la orden que nombra el primer argumento con los que le
// siguen.
func ejecutar(args []string) error {
	if len(args) == 0 {
		return errUso
	}

	switch args[0] {
	case ordenPiezas:
		return ejecutarPiezas(args[1:])
	case ordenCatalogo:
		return ejecutarCatalogo(args[1:])
	default:
		return errUso
	}
}

// ejecutarPiezas atiende la orden piezas con la composición de producción: lo
// que llega por sus banderas, la descripción corta del repositorio, las
// herramientas que el registro de producción anuncia y lo empotrado en el
// paquete raíz (H22 FR-014, FR-020; research.md D2 y D3).
func ejecutarPiezas(args []string) error {
	var piezas piezasAEscribir

	err := analizar(ordenPiezas, args,
		bandera{nombre: banderaVersion, valor: &piezas.Version},
		bandera{nombre: banderaMacOS, valor: &piezas.MacOS},
		bandera{nombre: banderaWindows, valor: &piezas.Windows},
		bandera{nombre: banderaIcono, valor: &piezas.Icono},
		bandera{nombre: banderaSalida, valor: &piezas.Salida},
	)
	if err != nil {
		return err
	}

	// Construir el registro no pide ni abre nada. Uno que no se construye es un
	// defecto de quien escribió un applet, que TestRegistroDeProduccion impide
	// publicar: si llegara aquí, el paso falla con su línea, como con todo lo
	// demás.
	registro, err := app.RegistroDeProduccion(piezas.Version)
	if err != nil {
		return fmt.Errorf("el registro de applets no se puede construir: %w", err)
	}

	piezas.Descripcion = Descripcion
	piezas.Herramientas = app.HerramientasAnunciadas(registro)
	piezas.Skills = kitlegal.Skills()

	return escribirPiezas(piezas)
}

// ejecutarCatalogo atiende la orden catalogo.
func ejecutarCatalogo(args []string) error {
	var version, huella, salida string

	err := analizar(ordenCatalogo, args,
		bandera{nombre: banderaVersion, valor: &version},
		bandera{nombre: banderaHuella, valor: &huella},
		bandera{nombre: banderaSalida, valor: &salida},
	)
	if err != nil {
		return err
	}

	return escribirCatalogo(version, huella, salida)
}

// bandera es una bandera de una orden: su nombre y dónde se deja su valor.
type bandera struct {
	nombre string
	valor  *string
}

// analizar lee los argumentos de una orden con sus banderas, que son todas
// obligatorias y llevan todas un valor. Unos argumentos que no tienen la forma
// de la orden —una bandera que no es suya o que llega sin su valor, un
// argumento que no es de ninguna bandera, la ayuda— son un fallo de uso, sea
// cual sea el error con que flag los rechaza: la línea de uso dice la forma
// entera. Una bandera de la orden que falta o que llega vacía se nombra, la
// primera en el orden de la orden.
func analizar(orden string, args []string, banderas ...bandera) error {
	analizador := flag.NewFlagSet(orden, flag.ContinueOnError)
	// El paso escribe una sola línea, la suya: lo que flag escribiría por su
	// cuenta ante unos argumentos que no valen no va a ningún sitio.
	analizador.SetOutput(io.Discard)

	for _, bandera := range banderas {
		analizador.StringVar(bandera.valor, bandera.nombre, "", "")
	}

	if err := analizador.Parse(args); err != nil || analizador.NArg() > 0 {
		return errUso
	}

	for _, bandera := range banderas {
		if *bandera.valor == "" {
			return fmt.Errorf("falta -%s", bandera.nombre)
		}
	}

	return nil
}
