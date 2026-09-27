// Package render es el adaptador de presentación: el único paquete del
// proyecto que escribe en la salida estándar y el que lleva a la salida de
// error los mensajes dirigidos a una persona (R5 de
// contracts/reglas-de-arquitectura.md, FR-040).
//
// Implementa la interfaz que el kernel declara y consume, de modo que
// internal/cli no lo importa: la raíz de composición construye el presentador
// con los dos descriptores y se lo inyecta. Así os.Stdout y os.Stderr solo se
// nombran en las raíces de composición y añadir una forma de presentación nueva
// no obliga a tocar ningún applet (FR-044, research.md D2, D15).
//
// La regla que gobierna todo el paquete: **toda escritura comprueba y propaga
// su error**, incluido el vaciado final de la tabla. Un descriptor roto —la
// tubería cerrada— es un fallo como cualquier otro y sube hasta el único punto
// que traduce errores a código de salida, en lugar de convertirse en un pánico
// o en silencio (contracts/banderas-y-exit-codes.md §4, research.md D27).
package render

import (
	"fmt"
	"io"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Los dos descriptores, nombrados para el mensaje del error de escritura: quien
// lo lea sabe cuál de los dos falló sin tener que deducirlo.
const (
	nombreSalida = "la salida estándar"
	nombreError  = "la salida de error"
)

// Presentador escribe por los dos descriptores que recibe, y por ninguno más.
// No conoce applets, ni errores tipados, ni códigos de salida: recibe un sobre
// ya montado o un texto ya compuesto y lo hace llegar a su descriptor.
type Presentador struct {
	salida  io.Writer
	errores io.Writer
}

// Nuevo construye el presentador con los dos descriptores de la invocación. Los
// entrega la raíz de composición —os.Stdout y os.Stderr en el binario, dos
// buffers en un test—, que es el único sitio donde se nombran.
func Nuevo(salida, errores io.Writer) *Presentador {
	return &Presentador{salida: salida, errores: errores}
}

// Presentar escribe el sobre en la salida estándar en la forma que pide enJSON:
// el documento JSON cuando es verdadero y la tabla mínima cuando no. Es el
// único camino por el que un sobre llega a la salida estándar, tanto en éxito
// como en fallo (FR-041, FR-042, FR-045). Qué sobre y en qué forma lo decide
// el kernel: este paquete no conoce applets, y un applet que cuenta su
// resultado para una persona llega aquí como texto, por Texto (docs/ADR/0026).
func (p *Presentador) Presentar(sobre schema.Sobre, enJSON bool) error {
	if enJSON {
		return enDescriptor(nombreSalida, escribirJSON(p.salida, sobre))
	}

	return enDescriptor(nombreSalida, escribirTabla(p.salida, sobre))
}

// Texto escribe en la salida estándar el texto dirigido a una persona: las tres
// líneas de version y la ayuda derivada del registro, que --json no altera
// porque solo elige la forma del sobre (FR-026, FR-042), y el contenido de un
// resultado contado para una persona, que sin --json ocupa el lugar de la tabla
// mínima (docs/ADR/0026).
func (p *Presentador) Texto(texto string) error {
	return enDescriptor(nombreSalida, escribirTexto(p.salida, texto))
}

// Aviso escribe en la salida de error el mensaje dirigido a una persona: la
// causa de un fallo, la lista de applets disponibles, la descripción de
// --dry-run y el aviso de un KITLEGAL_LOG inválido. No pasa por el registro de
// eventos, y por eso ningún nivel puede ocultarlo
// (contracts/banderas-y-exit-codes.md §6).
func (p *Presentador) Aviso(texto string) error {
	return enDescriptor(nombreError, escribirTexto(p.errores, texto))
}

// Salida es el escritor crudo de la salida estándar, que existe solo para
// entregárselo a quien escribe por su cuenta y no puede recibir un presentador:
// el analizador de la línea de órdenes, que emite ahí su ayuda (research.md
// D11).
func (p *Presentador) Salida() io.Writer { return p.salida }

// Error es el escritor crudo de la salida de error. Además de lo anterior, es
// el destino del registro de eventos, que se monta sobre él y nunca sobre la
// salida estándar (FR-036).
func (p *Presentador) Error() io.Writer { return p.errores }

// enDescriptor añade al fallo de una escritura el descriptor en el que ocurrió
// y lo devuelve envuelto, de modo que quien lo clasifique siga viendo la causa
// original con errors.Is. Un acierto no se envuelve: sigue siendo nil.
func enDescriptor(nombre string, err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("render: no se pudo escribir en %s: %w", nombre, err)
}
