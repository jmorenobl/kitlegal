package evals

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json/jsontext"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
)

// PrefijoDelLaboratorio es el principio de cada línea que el programa
// laboratorio (cmd/laboratorio) escribe en su salida de error: su nombre.
const PrefijoDelLaboratorio = "laboratorio: "

// UsoDelServidorDelLaboratorio es la forma de la orden laboratorio servidor
// (contracts/arnes.md §5 de H39), la que dice su línea de uso.
const UsoDelServidorDelLaboratorio = "laboratorio servidor -kitlegal <ruta> [-instrucciones <texto>] " +
	"[-herramienta <h> -parametro <p> -descripcion <texto>]"

// Las banderas de laboratorio servidor (contracts/arnes.md §5 de H39).
const (
	banderaDelKitlegal     = "kitlegal"
	banderaDeInstrucciones = "instrucciones"
	banderaDeLaHerramienta = "herramienta"
	banderaDelParametro    = "parametro"
	banderaDeLaDescripcion = "descripcion"
)

// codigoDelDefecto es el código con el que laboratorio servidor termina ante un
// defecto, el del fallo inesperado del contrato de resultados: unos argumentos
// sin la forma de la orden, un kitlegal que no arranca, una herramienta o un
// parámetro que la lista no trae, o una salida que no se puede leer o escribir.
const codigoDelDefecto = 1

// errUsoDelServidor es el fallo de unos argumentos que no tienen la forma de la
// orden: sin -kitlegal o con él vacío, con una bandera que no es suya o sin su
// valor, con un argumento que no es de ninguna bandera, con -herramienta,
// -parametro y -descripcion sin ir los tres juntos, o con la ayuda.
var errUsoDelServidor = errors.New("uso: " + UsoDelServidorDelLaboratorio)

// saltoDeLinea es lo que termina cada mensaje en la salida del servidor: un
// mensaje JSON por línea (research.md V13 de H39).
var saltoDeLinea = []byte("\n")

// servidorDelLaboratorio es lo que laboratorio servidor arranca y lo que
// reescribe en su salida.
type servidorDelLaboratorio struct {
	// kitlegal es la ruta del binario cuyo mcp serve se arranca.
	kitlegal string

	// instrucciones es el texto que pasa a ser result.instructions de la
	// respuesta a initialize, o nulo si no se reescribe.
	instrucciones *string

	// parametro es la descripción de parámetro que se reescribe en la
	// respuesta a tools/list, o nulo si no se reescribe.
	parametro *parametroReescrito
}

// parametroReescrito es la description que pasa a tener un parámetro de una
// herramienta en la respuesta a tools/list.
type parametroReescrito struct {
	herramienta, parametro, descripcion string
}

// tramo es dónde está un valor en un documento JSON: del byte inicio al byte
// anterior a fin.
type tramo struct {
	inicio, fin int64
}

// ServidorDelLaboratorio atiende laboratorio servidor con sus argumentos, los
// que siguen a la orden (contracts/arnes.md §5 de H39; research.md D7 de H39;
// FR-021, FR-024, FR-029): arranca `<kitlegal> mcp serve` con el entorno del
// proceso; le da la entrada sin tocarla y su salida de error a la dada; y de su
// salida, línea a línea, escribe en salida la respuesta a initialize con
// -instrucciones como result.instructions, la de tools/list con -descripcion
// como description del parámetro -parametro de la herramienta -herramienta, y
// las demás líneas byte a byte. Devuelve el código con el que termina el hijo.
// Un defecto lo dice en una línea de la salida de error y devuelve 1; el hijo,
// si estaba arrancado, no sigue. No importa el SDK de MCP (R7), y kitlegal no
// gana nada para esto (FR-029).
//
// El error de escribir esa línea no se propaga, y es una decisión y no un
// silencio: el código ya dice que hubo un defecto, que es lo que lee quien lo
// arranca, y contarlo iría a la misma salida que acaba de fallar.
func ServidorDelLaboratorio(ctx context.Context, args []string, entrada io.Reader, salida, errores io.Writer) int {
	codigo, err := servirElLaboratorio(ctx, args, entrada, salida, errores)
	if err != nil {
		_, _ = fmt.Fprintf(errores, "%s%v\n", PrefijoDelLaboratorio, err)

		return codigoDelDefecto
	}

	return codigo
}

// servirElLaboratorio lee los argumentos y, si tienen la forma de la orden,
// sirve con ellos.
func servirElLaboratorio(ctx context.Context, args []string, entrada io.Reader, salida, errores io.Writer) (int, error) {
	servidor, err := leerElServidorDelLaboratorio(args)
	if err != nil {
		return 0, err
	}

	return servidor.servir(ctx, entrada, salida, errores)
}

// leerElServidorDelLaboratorio lee los argumentos de laboratorio servidor.
func leerElServidorDelLaboratorio(args []string) (servidorDelLaboratorio, error) {
	analizador := flag.NewFlagSet("laboratorio servidor", flag.ContinueOnError)
	// La orden escribe una sola línea, la suya: lo que flag escribiría por su
	// cuenta ante unos argumentos que no valen no va a ningún sitio.
	analizador.SetOutput(io.Discard)

	var kitlegal, instrucciones, herramienta, parametro, descripcion string

	analizador.StringVar(&kitlegal, banderaDelKitlegal, "", "")
	analizador.StringVar(&instrucciones, banderaDeInstrucciones, "", "")
	analizador.StringVar(&herramienta, banderaDeLaHerramienta, "", "")
	analizador.StringVar(&parametro, banderaDelParametro, "", "")
	analizador.StringVar(&descripcion, banderaDeLaDescripcion, "", "")

	if err := analizador.Parse(args); err != nil || analizador.NArg() > 0 || kitlegal == "" {
		return servidorDelLaboratorio{}, errUsoDelServidor
	}

	dadas := map[string]bool{}
	analizador.Visit(func(bandera *flag.Flag) { dadas[bandera.Name] = true })

	servidor := servidorDelLaboratorio{kitlegal: kitlegal}

	if dadas[banderaDeInstrucciones] {
		servidor.instrucciones = &instrucciones
	}

	delParametro := []bool{dadas[banderaDeLaHerramienta], dadas[banderaDelParametro], dadas[banderaDeLaDescripcion]}
	if slices.Contains(delParametro, true) {
		if slices.Contains(delParametro, false) {
			return servidorDelLaboratorio{}, errUsoDelServidor
		}

		servidor.parametro = &parametroReescrito{herramienta: herramienta, parametro: parametro, descripcion: descripcion}
	}

	return servidor, nil
}

// servir arranca el hijo, filtra su salida y espera a que termine. Si el filtro
// falla, cierra la salida del hijo, que ya nadie lee, y lo mata antes de
// esperarlo: un servidor cuya respuesta no llega no sirve a la sesión.
func (s servidorDelLaboratorio) servir(ctx context.Context, entrada io.Reader, salida, errores io.Writer) (int, error) {
	hijo := ordenDelHijo(ctx, s.kitlegal)
	hijo.Stdin = entrada
	hijo.Stderr = errores

	deLaSalida, err := hijo.StdoutPipe()
	if err != nil {
		return 0, fmt.Errorf("la salida de %s no se puede leer: %w", s.orden(), err)
	}

	if err := hijo.Start(); err != nil {
		return 0, fmt.Errorf("%s no arranca: %w", s.orden(), err)
	}

	errDelFiltro := s.filtrar(deLaSalida, salida)
	if errDelFiltro != nil {
		errDelFiltro = errors.Join(errDelFiltro, detenerElHijo(hijo, deLaSalida))
	}

	codigo, errDeLaEspera := codigoDelProceso(hijo.Wait())

	// Tras un filtro que falla, el hijo termina porque se le mata: lo que dice su
	// espera no es el defecto.
	if errDelFiltro != nil {
		return 0, errDelFiltro
	}

	if errDeLaEspera != nil {
		return 0, fmt.Errorf("%s: %w", s.orden(), errDeLaEspera)
	}

	return codigo, nil
}

// ordenDelHijo es la orden `<kitlegal> mcp serve` con el entorno del proceso:
// el programa es el parámetro y sus argumentos son constantes, que es lo que
// gosec (G204) admite como orden que no lleva nada de fuera en los argumentos.
func ordenDelHijo(ctx context.Context, kitlegal string) *exec.Cmd {
	return exec.CommandContext(ctx, kitlegal, appletDelServidor, verboDelServidor)
}

// orden es la orden del hijo, como la nombran los errores.
func (s servidorDelLaboratorio) orden() string {
	return s.kitlegal + " " + appletDelServidor + " " + verboDelServidor
}

// detenerElHijo cierra la salida del hijo y lo mata. Que ya hubiera terminado
// no es un error.
func detenerElHijo(hijo *exec.Cmd, deLaSalida io.Closer) error {
	var errDelCierre, errDeLaSenal error

	if err := deLaSalida.Close(); err != nil {
		errDelCierre = fmt.Errorf("la salida del servidor no se puede cerrar: %w", err)
	}

	if err := hijo.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
		errDeLaSenal = fmt.Errorf("el servidor no se puede detener: %w", err)
	}

	return errors.Join(errDelCierre, errDeLaSenal)
}

// filtrar escribe en salida, línea a línea y cada una en cuanto llega entera,
// lo que el hijo escribe en lee, con cada línea reescrita por reescribir. Una
// última línea sin salto se escribe igual, también sin él.
func (s servidorDelLaboratorio) filtrar(lee io.Reader, salida io.Writer) error {
	lector := bufio.NewReader(lee)

	for {
		linea, errDeLectura := lector.ReadBytes('\n')

		if len(linea) > 0 {
			reescrita, err := s.reescribir(linea)
			if err != nil {
				return err
			}

			if _, err := salida.Write(reescrita); err != nil {
				return fmt.Errorf("la salida estándar no admite la línea del servidor: %w", err)
			}
		}

		switch {
		case errors.Is(errDeLectura, io.EOF):
			return nil
		case errDeLectura != nil:
			return fmt.Errorf("la salida de %s no se puede leer: %w", s.orden(), errDeLectura)
		}
	}
}

// reescribir es la línea dada, con su salto si lo lleva, tras reescribir en ella
// lo que haya que reescribir. Una línea que no es un documento JSON, o que no es
// la respuesta a initialize ni la de tools/list, queda byte a byte; en la que
// lo es, solo cambia el valor que se reescribe.
func (s servidorDelLaboratorio) reescribir(linea []byte) ([]byte, error) {
	if s.instrucciones == nil && s.parametro == nil {
		return linea, nil
	}

	mensaje := bytes.TrimSuffix(linea, saltoDeLinea)
	salto := linea[len(mensaje):]

	if !jsontext.Value(mensaje).IsValid() {
		return linea, nil
	}

	var err error

	if s.instrucciones != nil {
		if mensaje, err = conLasInstrucciones(mensaje, *s.instrucciones); err != nil {
			return nil, err
		}
	}

	if s.parametro != nil {
		if mensaje, err = s.parametro.reescribirEn(mensaje); err != nil {
			return nil, err
		}
	}

	return slices.Concat(mensaje, salto), nil
}

// conLasInstrucciones es el mensaje con el texto dado como result.instructions
// si es la respuesta a initialize, un objeto con result.instructions
// (contracts/arnes.md §5 de H39); si no, el mensaje tal cual.
func conLasInstrucciones(mensaje []byte, texto string) ([]byte, error) {
	instrucciones, hay, err := tramoEn(mensaje, "result", "instructions")
	if err != nil || !hay {
		return mensaje, err
	}

	return conElTexto(mensaje, instrucciones, texto)
}

// reescribirEn es el mensaje con la descripción dada como description del
// parámetro de la herramienta si es la respuesta a tools/list, un objeto con
// result.tools (contracts/arnes.md §5 de H39); si no, el mensaje tal cual. Una
// lista sin la herramienta, una herramienta sin el parámetro en
// inputSchema.properties o un parámetro sin description son un error.
func (p parametroReescrito) reescribirEn(mensaje []byte) ([]byte, error) {
	lista, hay, err := tramoEn(mensaje, "result", "tools")
	if err != nil || !hay {
		return mensaje, err
	}

	enLaLista, hay, err := p.herramientaEn(mensaje[lista.inicio:lista.fin])
	if err != nil {
		return nil, err
	}

	if !hay {
		return nil, fmt.Errorf("la lista de herramientas no trae %s", p.herramienta)
	}

	desde := lista.inicio + enLaLista.inicio
	herramienta := mensaje[desde : lista.inicio+enLaLista.fin]

	if _, hay, err = tramoEn(herramienta, "inputSchema", "properties", p.parametro); err != nil {
		return nil, err
	}

	if !hay {
		return nil, fmt.Errorf("la herramienta %s no trae el parámetro %s", p.herramienta, p.parametro)
	}

	descripcion, hay, err := tramoEn(herramienta, "inputSchema", "properties", p.parametro, "description")
	if err != nil {
		return nil, err
	}

	if !hay {
		return nil, fmt.Errorf("el parámetro %s de %s no trae descripción", p.parametro, p.herramienta)
	}

	return conElTexto(mensaje, tramo{inicio: desde + descripcion.inicio, fin: desde + descripcion.fin}, p.descripcion)
}

// herramientaEn es el tramo, en la lista dada, de la herramienta cuyo name es
// el de p, y si la lista la trae.
func (p parametroReescrito) herramientaEn(lista []byte) (tramo, bool, error) {
	lector := jsontext.NewDecoder(bytes.NewReader(lista))

	if lector.PeekKind() != jsontext.KindBeginArray {
		return tramo{}, false, nil
	}

	if _, err := lector.ReadToken(); err != nil {
		return tramo{}, false, fmt.Errorf("la lista de herramientas no se puede leer: %w", err)
	}

	for lector.PeekKind() != jsontext.KindEndArray {
		valor, err := lector.ReadValue()
		if err != nil {
			return tramo{}, false, fmt.Errorf("la lista de herramientas no se puede leer: %w", err)
		}

		fin := lector.InputOffset()
		herramienta := tramo{inicio: fin - int64(len(valor)), fin: fin}

		nombre, err := nombreDe(lista[herramienta.inicio:herramienta.fin])
		if err != nil {
			return tramo{}, false, err
		}

		if nombre == p.herramienta {
			return herramienta, true, nil
		}
	}

	return tramo{}, false, nil
}

// nombreDe es el name de la herramienta dada, o vacío si no lo tiene.
func nombreDe(herramienta []byte) (string, error) {
	nombre, hay, err := tramoEn(herramienta, "name")
	if err != nil || !hay {
		return "", err
	}

	texto, err := jsontext.AppendUnquote(nil, herramienta[nombre.inicio:nombre.fin])
	if err != nil {
		return "", fmt.Errorf("el name de una herramienta no es un texto: %w", err)
	}

	return string(texto), nil
}

// tramoEn es el tramo del valor al que lleva, desde el documento dado, la ruta
// de nombres de miembros dada, y si el documento lo tiene: no lo tiene si algún
// valor de la ruta no es un objeto o no tiene el miembro siguiente.
func tramoEn(documento []byte, ruta ...string) (tramo, bool, error) {
	lector := jsontext.NewDecoder(bytes.NewReader(documento))

	for _, nombre := range ruta {
		hay, err := entrarEnElMiembro(lector, nombre)
		if err != nil || !hay {
			return tramo{}, false, err
		}
	}

	valor, err := lector.ReadValue()
	if err != nil {
		return tramo{}, false, fmt.Errorf("el mensaje del servidor no se puede leer: %w", err)
	}

	fin := lector.InputOffset()

	return tramo{inicio: fin - int64(len(valor)), fin: fin}, true, nil
}

// entrarEnElMiembro deja el lector, que está delante de un valor, delante del
// valor de su miembro del nombre dado, y dice si lo tiene: no lo tiene si el
// valor no es un objeto o si no tiene ese miembro.
func entrarEnElMiembro(lector *jsontext.Decoder, nombre string) (bool, error) {
	if lector.PeekKind() != jsontext.KindBeginObject {
		return false, nil
	}

	if _, err := lector.ReadToken(); err != nil {
		return false, fmt.Errorf("el mensaje del servidor no se puede leer: %w", err)
	}

	for lector.PeekKind() != jsontext.KindEndObject {
		clave, err := lector.ReadToken()
		if err != nil {
			return false, fmt.Errorf("el mensaje del servidor no se puede leer: %w", err)
		}

		if clave.String() == nombre {
			return true, nil
		}

		if err := lector.SkipValue(); err != nil {
			return false, fmt.Errorf("el mensaje del servidor no se puede leer: %w", err)
		}
	}

	return false, nil
}

// conElTexto es el documento con el valor del tramo dado sustituido por el
// texto dado como texto de JSON; lo demás, byte a byte.
func conElTexto(documento []byte, valor tramo, texto string) ([]byte, error) {
	codificado, err := jsontext.AppendQuote(nil, texto)
	if err != nil {
		return nil, fmt.Errorf("el texto %q no cabe en un mensaje del protocolo: %w", texto, err)
	}

	return slices.Concat(documento[:valor.inicio], codificado, documento[valor.fin:]), nil
}
