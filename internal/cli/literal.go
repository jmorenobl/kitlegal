package cli

import (
	"fmt"

	"github.com/alecthomas/kong"
)

// Literal es un argumento de texto que llega al applet con los bytes con los
// que se escribió en la línea de órdenes, también los que no son UTF-8.
//
// Un campo string no los conserva: el analizador (Kong v1.16.1) lleva cada
// argumento de texto por una codificación JSON de ida y vuelta, que cambia cada
// byte que no es UTF-8 por U+FFFD antes de que llegue a ningún applet, y hace
// lo mismo con un tipo que implemente encoding.TextUnmarshaler. Un Literal toma
// el valor del analizador sin codificarlo, así que el applet que lo declara
// trabaja con lo que recibió y puede nombrarlo tal cual: es lo que pide un id
// que se busca tal cual (FR-052 y FR-053 de H7).
//
// Para --describe es una cadena, como un string: quien invoca no distingue uno
// de otro. El paquete de composición lo declara sin importar Kong, que sigue
// siendo cosa de este paquete (research.md D1, D2).
type Literal string

// Decode es el decodificador con el que el analizador rellena un Literal en
// lugar del de las cadenas: toma el siguiente valor, con las mismas reglas que
// el de las cadenas para decidir qué es un valor, y lo guarda byte a byte.
//
// Todo valor que llega de la línea de órdenes es un texto. Uno que no lo fuera
// solo podría venir de una gramática que declara otra cosa, y el de las
// cadenas tampoco lo aceptaría: es un error de argumentos como los demás del
// análisis.
func (l *Literal) Decode(ctx *kong.DecodeContext) error {
	token, err := ctx.Scan.PopValue("string")
	if err != nil {
		return err
	}

	valor, esTexto := token.Value.(string)
	if !esTexto {
		return fmt.Errorf("se esperaba un texto y se recibió %v (%T)", token.Value, token.Value)
	}

	*l = Literal(valor)

	return nil
}

// La comprobación en tiempo de compilación de que el analizador usa Decode.
var _ kong.MapperValue = (*Literal)(nil)
