package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"time"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/mcp"
	"github.com/jmorenobl/kitlegal/internal/render"
)

// separadorDeHerramienta une el applet y el verbo en el nombre de una
// herramienta: `<applet>_<verbo>` (H21 FR-003).
const separadorDeHerramienta = "_"

// Las banderas globales con las que cada llamada invoca su verbo, escritas como
// las escribe quien da la orden. --json va siempre: el resultado de una llamada
// es el sobre. Las otras tres son las que se dan a `mcp serve` y valen para
// todas sus llamadas (H21 FR-020; contracts/servidor-mcp.md §3 de H21).
const (
	banderaJSON     = "--json"
	banderaTimeout  = "--timeout="
	banderaOffline  = "--offline"
	banderaSinGrafo = "--no-graph"
)

// appletsSinHerramientas son los applets cuyos verbos el servidor no anuncia:
// skills, que instala en el equipo de quien lo usa y no consulta nada, y mcp,
// que es el propio servidor (H21 FR-002). Salen de los propios applets, de modo
// que cada nombre esté escrito en un solo sitio.
var appletsSinHerramientas = []string{appletMCP{}.Nombre(), appletSkills{}.Nombre()}

// verboAnunciado es un verbo del registro del que sale una herramienta: el
// applet que lo ofrece, el verbo y el nombre de la herramienta.
type verboAnunciado struct {
	herramienta string
	applet      Applet
	verbo       Verbo
}

// verbosAnunciados son los verbos del registro de los que sale una herramienta
// del servidor MCP: los de cada applet, en el orden de sus nombres y en el de
// su catálogo, salvo los de appletsSinHerramientas, cada uno con el nombre
// `<applet>_<verbo>`. Es el único sitio que dice qué verbos dan herramienta y
// cómo se llama cada una (H21 FR-002, FR-003).
func verbosAnunciados(registro *Registro) []verboAnunciado {
	var anunciados []verboAnunciado

	for _, nombre := range registro.Nombres() {
		if slices.Contains(appletsSinHerramientas, nombre) {
			continue
		}

		applet, _ := registro.Buscar(nombre)

		for _, verbo := range applet.Verbos() {
			anunciados = append(anunciados, verboAnunciado{
				herramienta: nombre + separadorDeHerramienta + verbo.Nombre,
				applet:      applet,
				verbo:       verbo,
			})
		}
	}

	return anunciados
}

// NombresDeHerramientas son los nombres de las herramientas que el servidor MCP
// anuncia con ese registro: las de herramientasDe, en su orden. Es lo que lee
// quien tiene que reconocer una llamada a una de ellas sin arrancar el servidor
// —el job de evals, en el transcript de una sesión— y sin repetir qué applets no
// dan herramientas (H21 FR-002, FR-042; contracts/evals-en-dos-modos.md §3 de
// H21).
func NombresDeHerramientas(registro *Registro) []string {
	var nombres []string

	for _, anunciado := range verbosAnunciados(registro) {
		nombres = append(nombres, anunciado.herramienta)
	}

	return nombres
}

// herramientasDe da las herramientas del servidor MCP: una por cada verbo de
// cada applet del registro, salvo los de appletsSinHerramientas, con el nombre
// `<applet>_<verbo>`, la descripción del verbo y los dos esquemas que salen del
// generador de --describe. No hay ninguna lista escrita a mano: un applet o un
// verbo nuevo en el registro es una herramienta más sin tocar nada aquí (H21
// FR-002 a FR-004; data-model §1 de H21).
//
// ec es el contexto de ejecución de `mcp serve`, del que salen las banderas de
// cada llamada; registrador, el del servidor, con el nivel ya resuelto, que
// registra un evento por llamada; y errores, la salida de error del servidor,
// adonde va el mensaje de cada llamada que falla.
//
// Un verbo que no se puede describir —sin fábrica de argumentos, o cuyos
// esquemas no se pueden construir— es un defecto de quien escribió el applet:
// el servidor no arranca con una herramienta a medias.
func herramientasDe(
	registro *Registro, ec schema.Contexto, registrador *slog.Logger, errores io.Writer,
) ([]mcp.Herramienta, error) {
	// El aviso de versión lo da el kernel una vez, al analizar `mcp serve`: las
	// llamadas resuelven con un registro que no lo da (H21 FR-023).
	deLasLlamadas := registro.sinAvisador()
	banderas := banderasDeCadaLlamada(ec)

	var herramientas []mcp.Herramienta

	for _, anunciado := range verbosAnunciados(registro) {
		argumentos, err := argumentosDelVerbo(anunciado.applet, anunciado.verbo)
		if err != nil {
			return nil, err
		}

		// Los argumentos de la fábrica solo se reflejan, aquí y en cada
		// llamada: quien los rellena es la gramática de cada invocación, que
		// pide los suyos.
		descrito := verboDescrito(anunciado.applet, anunciado.verbo, argumentos.Addr().Interface())

		entrada, salida, err := cli.EsquemasDeHerramienta(descrito)
		if err != nil {
			return nil, fmt.Errorf("app: los esquemas de la herramienta %s no se pueden construir: %w",
				anunciado.herramienta, err)
		}

		llamada := llamadaDeHerramienta{
			applet:      anunciado.applet,
			verbo:       descrito,
			banderas:    banderas,
			registro:    deLasLlamadas,
			registrador: registrador,
			errores:     errores,
		}

		herramientas = append(herramientas, mcp.Herramienta{
			Nombre:      anunciado.herramienta,
			Descripcion: anunciado.verbo.Descripcion,
			Entrada:     entrada,
			Salida:      salida,
			Llamar:      llamada.atender,
		})
	}

	return herramientas, nil
}

// banderasDeCadaLlamada son las banderas globales con las que el servidor
// invoca el verbo de cada llamada: --json, el plazo de `mcp serve` como plazo
// de la llamada y, si se le dieron, --offline y --no-graph. Ninguna es un
// parámetro de una herramienta (H21 FR-020).
func banderasDeCadaLlamada(ec schema.Contexto) []string {
	banderas := []string{banderaJSON, banderaTimeout + ec.Timeout.String()}

	if ec.Offline {
		banderas = append(banderas, banderaOffline)
	}

	if ec.SinGrafo {
		banderas = append(banderas, banderaSinGrafo)
	}

	return banderas
}

// llamadaDeHerramienta es lo que la herramienta de un verbo necesita para
// atender cada llamada como una invocación del kernel (research.md D3 de H21).
// No guarda nada de una llamada a otra: el servidor las atiende a la vez, cada
// una en su gorrutina.
type llamadaDeHerramienta struct {
	// applet es el applet del verbo, ya resuelto: el despacho no se repite.
	applet Applet
	// verbo es el verbo como el kernel lo describe: de él sale la línea de
	// órdenes de cada llamada.
	verbo cli.Verbo
	// banderas son las del servidor que valen para cada llamada.
	banderas []string
	// registro es la copia sin avisador del registro del servidor.
	registro *Registro
	// registrador es el del servidor.
	registrador *slog.Logger
	// errores es la salida de error del servidor.
	errores io.Writer
}

// atender atiende una llamada: ejecuta el verbo como el kernel ejecuta la orden
// `kitlegal <applet> <verbo> --json --timeout=<plazo> [--offline] [--no-graph]
// -- <posicionales…>`, con un presentador cuyo escritor de salida es un búfer
// de esta llamada, registra su evento y termina con el mismo final que Main,
// que monta el sobre —de éxito o de fallo— en ese búfer y, si terminó bien,
// entrega al grafo del mundo lo que el verbo observó antes de volver: la
// llamada que el cliente envíe después lo ve (H21 FR-010, FR-013;
// contracts/servidor-mcp.md §3 de H21).
//
// El sobre son los bytes del búfer sin el salto final: los que la orden escribe
// en la salida estándar. Es un error de herramienta si y solo si el código de
// la invocación no es 0, que es cuando el sobre lleva `ok` falso (H21 FR-011).
// El mensaje de una llamada que falla y la línea de una entrega que no llega al
// grafo van a la salida de error del servidor, como en la orden.
func (l llamadaDeHerramienta) atender(argumentos json.RawMessage) mcp.Resultado {
	var sobre bytes.Buffer

	p := render.Nuevo(&sobre, l.errores)

	inicio := time.Now()
	fin, invocacion := l.resolver(p, argumentos)

	// Un evento por llamada, con el applet y el verbo de la herramienta también
	// cuando sus argumentos no llegan a analizarse, y antes del mensaje del
	// fallo, como en la orden.
	cli.RegistrarEvento(context.Background(), l.registrador, cli.Evento{
		Applet:     l.verbo.Applet,
		Verbo:      l.verbo.Verbo,
		Duracion:   time.Since(inicio),
		Clase:      claseDe(fin.err),
		Argumentos: invocacion,
	})

	codigo := emitir(p, l.registro, fin)

	return mcp.Resultado{
		Sobre: bytes.TrimSuffix(sobre.Bytes(), []byte("\n")),
		Fallo: codigo != cli.CodigoSalida(nil),
	}
}

// resolver convierte los argumentos de la llamada en la línea de órdenes del
// verbo y la resuelve como el kernel resuelve la de una invocación: el mismo
// análisis, el plazo de la llamada, Ejecutar y el plazo vencido. Devuelve
// además la línea, que es lo que el evento registra en depuración.
//
// Unos argumentos que no valen no ejecutan nada: su desenlace es el fallo de la
// clase `argumentos`, con la forma de un fallo anterior al applet.
func (l llamadaDeHerramienta) resolver(p cli.Presentador, argumentos json.RawMessage) (desenlace, []string) {
	// Una llamada devuelve siempre el sobre: es la forma con la que se presenta
	// también el fallo anterior al análisis.
	previo := cli.Preliminar{JSON: true}

	linea, err := cli.LineaDeLlamada(l.verbo, argumentos)
	if err != nil {
		return desenlace{enJSON: previo.JSON, err: err}, nil
	}

	// El verbo, las banderas del servidor y la línea de la llamada, cuyos
	// argumentos de posición van detrás del terminador: ninguno puede ser una
	// bandera ni nombrar otro verbo.
	invocacion := slices.Concat([]string{l.verbo.Verbo}, l.banderas, linea)

	fin, _ := resolverApplet(p, l.registrador, l.registro,
		Despacho{Destino: DestinoApplet, Applet: l.applet, Args: invocacion}, previo)

	return fin, invocacion
}
