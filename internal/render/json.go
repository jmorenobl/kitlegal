package render

import (
	"encoding/json"
	"io"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// escribirJSON emite el sobre en la forma legible por máquina: un único
// documento JSON, sin sangrado, con un salto de línea final y **nada más** en
// la salida estándar —ni una línea de registro, ni un aviso, ni una cabecera—
// (FR-042, contracts/banderas-y-exit-codes.md §5).
//
// El escape de los caracteres HTML se desactiva: con él, la url del sobre
// saldría con secuencias \uXXXX en lugar de sus «&», «<» y «>», y dejaría de
// poder copiarse tal cual. Es la misma decisión con la que el dominio calcula
// la forma canónica de la que sale la huella (research.md D15).
//
// El salto de línea final lo pone json.Encoder.Encode, que además escribe el
// documento de una vez: si el descriptor está roto, el fallo se devuelve y no
// queda medio documento en la salida estándar.
func escribirJSON(destino io.Writer, sobre schema.Sobre) error {
	codificador := json.NewEncoder(destino)
	codificador.SetEscapeHTML(false)

	return codificador.Encode(sobre)
}
