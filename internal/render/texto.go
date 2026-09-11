package render

import (
	"io"
	"strings"
)

// escribirTexto escribe el texto dirigido a una persona y cierra su última
// línea, sin tocar nada más de lo que se le da: las tres líneas de version y la
// ayuda derivada del registro salen tal cual por la salida estándar, y el
// mensaje de un fallo, la lista de applets disponibles y la descripción de
// --dry-run salen tal cual por la salida de error.
//
// Ninguno de ellos pasa por el registro de eventos, y por eso ningún nivel
// puede ocultarlos: lo que el contrato promete visible no viaja por un canal
// filtrable (contracts/banderas-y-exit-codes.md §6).
//
// Un texto vacío no escribe nada: no hay línea que cerrar, y una línea en
// blanco sería una línea que nadie pidió. La escritura es una sola, de modo que
// el texto no pueda llegar partido a la mitad si el descriptor falla.
func escribirTexto(destino io.Writer, texto string) error {
	if texto == "" {
		return nil
	}

	if !strings.HasSuffix(texto, "\n") {
		texto += "\n"
	}

	_, err := io.WriteString(destino, texto)

	return err
}
