package disco

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"

	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
)

// La sonda de Disponible (research.md D9): un enlace de nombre
// .kitlegal-sonda-<16 hexadecimales aleatorios> con un destino literal que no
// resuelve y que nunca se sigue.
const (
	prefijoDeLaSonda = ".kitlegal-sonda-"
	destinoDeLaSonda = "kitlegal-sonda"
	// intentosDeSonda son los nombres que se prueban, uno tras otro, mientras
	// el que se prueba ya existe, antes de darse por vencido.
	intentosDeSonda = 8
)

// Enlazador es el creador de enlaces del sistema: el puerto Enlazador del
// dominio sobre os.Symlink (FR-024; research.md D9). Su valor cero es el del
// binario; los dos campos solo los sustituye el test del paquete, para forzar
// un nombre de sonda que ya existe y una sonda que no se puede retirar.
type Enlazador struct {
	// aleatorio da la parte aleatoria del nombre de cada sonda; sin él, la de
	// crypto/rand.
	aleatorio func() string

	// retirar retira la sonda; sin él, os.Remove.
	retirar func(ruta string) error
}

// El Enlazador es un Enlazador.
var _ instalacion.Enlazador = Enlazador{}

// Disponible dice si en directorio se puede crear un enlace simbólico, y lo
// dice creando uno en ese mismo directorio —el sistema de ficheros donde
// vivirá la entrada de host, no el de TMPDIR— y retirándolo después, así que
// lo deja con las mismas entradas y los mismos bytes (FR-048, FR-068).
//
// Si el enlace no se puede crear, por la razón que sea —el sistema de ficheros
// no los admite, no hay permiso o el directorio no existe, que la sonda no
// crea—, la respuesta es no. Si su nombre ya existe, prueba otro sin tocar lo
// que había, hasta intentosDeSonda, y sin ninguno libre es un error. Si no se
// puede retirar, el directorio ya no está como estaba, y es un error que
// nombra la sonda: «retirar la sonda <ruta>: <error del sistema>»
// (contracts/applet-skills.md §5).
func (e Enlazador) Disponible(directorio string) (bool, error) {
	for range intentosDeSonda {
		sonda := path.Join(directorio, prefijoDeLaSonda+e.parteAleatoria())

		creada, existia := crearLaSonda(sonda)

		switch {
		case existia:
			continue
		case !creada:
			return false, nil
		}

		if err := e.retirarLaSonda(sonda); err != nil {
			return false, fallo(operacionRetirar, sonda, delSistema(err))
		}

		return true, nil
	}

	return false, fmt.Errorf("crear la sonda en %s: los %d nombres probados ya existían", directorio, intentosDeSonda)
}

// crearLaSonda crea el enlace de la sonda y dice si lo creó y, si no, si fue
// porque en su ruta ya había algo. Que no se pueda crear por otra razón no es
// un error sino la respuesta de Disponible, así que no se devuelve.
func crearLaSonda(sonda string) (creada, existia bool) {
	err := os.Symlink(destinoDeLaSonda, sonda)

	return err == nil, errors.Is(err, fs.ErrExist)
}

// parteAleatoria son los dieciséis hexadecimales del nombre de una sonda.
//
// Salen de crypto/rand y no de math/rand, que gosec marca con G404. El
// resultado de Read no se comprueba porque la función «never returns an error,
// and always fills b entirely» (go doc crypto/rand.Read); ante un fallo del
// lector del sistema el proceso termina, como en internal/httpx.
func (e Enlazador) parteAleatoria() string {
	if e.aleatorio != nil {
		return e.aleatorio()
	}

	var octetos [8]byte

	rand.Read(octetos[:])

	return hex.EncodeToString(octetos[:])
}

// retirarLaSonda retira el enlace de la sonda, sin seguirlo.
func (e Enlazador) retirarLaSonda(sonda string) error {
	if e.retirar != nil {
		return e.retirar(sonda)
	}

	return os.Remove(sonda)
}

// Enlazar crea en ruta un enlace simbólico con ese destino literal, sin
// limpiarlo ni comprobar que resuelve (FR-021). Si en ruta ya hay algo, no lo
// toca.
func (Enlazador) Enlazar(destino, ruta string) error {
	return delSistema(os.Symlink(destino, ruta))
}
