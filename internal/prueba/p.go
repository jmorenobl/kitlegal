// Package prueba es un fixture desechable del escenario 11.
package prueba

import "fmt"

// P escribe directamente en stdout, que es lo que forbidigo debe bloquear.
func P() { fmt.Println("no") }
