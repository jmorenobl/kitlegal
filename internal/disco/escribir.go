package disco

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
)

// permisosDeDirectorio son los de cada directorio que se crea, los mismos que
// usa internal/skills (research.md D8). Los ficheros quedan con los de
// os.CreateTemp, 0o600: ninguno lleva permiso de ejecución (FR-015).
const permisosDeDirectorio = 0o750

// patronDelTemporal es el nombre del temporal de cada escritura, en el mismo
// directorio que el fichero: oculto y sin forma de fichero de una skill.
const patronDelTemporal = ".kitlegal-*.tmp"

// Escritor es el puerto Escritor del dominio sobre el sistema de ficheros:
// las operaciones sueltas con las que se aplica un plan, sin decidir nada
// (research.md D7, D8). Los enlaces los crea con el Enlazador que recibe, así
// que su valor cero no sirve: se construye con NuevoEscritor.
type Escritor struct {
	enlazador instalacion.Enlazador
}

// El Escritor es un Escritor.
var _ instalacion.Escritor = Escritor{}

// NuevoEscritor es el Escritor que crea los enlaces con enlazador, el de la
// invocación: el del sistema en el binario distribuido y otro en test y en el
// binario de e2e (FR-024; research.md D9).
func NuevoEscritor(enlazador instalacion.Enlazador) Escritor {
	return Escritor{enlazador: enlazador}
}

// CrearDirectorio crea el directorio real de ruta, cuyo padre existe. Si en
// ruta ya hay algo, un enlace colgando incluido, no crea nada.
func (Escritor) CrearDirectorio(ruta string) error {
	return delSistema(os.Mkdir(filepath.Clean(ruta), permisosDeDirectorio))
}

// EscribirFichero deja en ruta un fichero regular con contenido, de forma
// atómica: lo escribe en un temporal del mismo directorio, lo vuelca a disco y
// lo renombra encima, así que un corte nunca deja un fichero a medias con el
// nombre final; y el temporal se retira si algo falla (research.md D8).
//
// Solo sustituye un fichero regular: el rename sustituiría sin quejarse un
// enlace, una tubería o un socket, que no son lo que el plan escribe, así que
// en esos no escribe nada y devuelve «no es un fichero regular» (FR-028,
// FR-047). Un directorio no hace falta mirarlo antes: el rename no puede
// sustituirlo, y falla.
func (Escritor) EscribirFichero(ruta string, contenido []byte) error {
	destino := filepath.Clean(ruta)

	if err := comprobarDestino(destino); err != nil {
		return err
	}

	temporal, err := os.CreateTemp(filepath.Dir(destino), patronDelTemporal)
	if err != nil {
		return delSistema(err)
	}

	if err := volcarYCerrar(temporal, contenido); err != nil {
		return errors.Join(err, retirarElTemporal(temporal.Name()))
	}

	if err := os.Rename(filepath.Clean(temporal.Name()), destino); err != nil {
		return errors.Join(delSistema(err), retirarElTemporal(temporal.Name()))
	}

	return nil
}

// comprobarDestino comprueba, sin seguirlo, que en destino no hay nada, o hay
// lo que el rename sustituye con razón —un fichero regular— o rechaza por sí
// mismo —un directorio—; en otro caso, «no es un fichero regular».
func comprobarDestino(destino string) error {
	examinado, err := os.Lstat(destino)

	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil
	case err != nil:
		return delSistema(err)
	case examinado.Mode().IsRegular(), examinado.IsDir():
		return nil
	default:
		return errNoEsFicheroRegular
	}
}

// volcarYCerrar escribe contenido en el temporal, lo vuelca a disco y lo
// cierra. El cierre se hace pase lo que pase, y ningún error se pierde.
func volcarYCerrar(temporal *os.File, contenido []byte) error {
	_, errAlEscribir := temporal.Write(contenido)

	var errAlVolcar error
	if errAlEscribir == nil {
		errAlVolcar = temporal.Sync()
	}

	errAlCerrar := temporal.Close()

	return errors.Join(delSistema(errAlEscribir), delSistema(errAlVolcar), delSistema(errAlCerrar))
}

// retirarElTemporal retira el temporal de una escritura que falló. Si tampoco
// se puede, se dice junto al fallo que trajo hasta aquí: el temporal se queda
// en el directorio.
func retirarElTemporal(nombre string) error {
	return delSistema(os.Remove(filepath.Clean(nombre)))
}

// Retirar retira la entrada de ruta: un fichero, un enlace, sin seguirlo, o un
// directorio vacío. Un directorio con algo dentro no se retira.
func (Escritor) Retirar(ruta string) error {
	return delSistema(os.Remove(filepath.Clean(ruta)))
}

// Enlazar crea en ruta un enlace simbólico con ese destino literal, con el
// Enlazador de la invocación.
func (e Escritor) Enlazar(destino, ruta string) error {
	return e.enlazador.Enlazar(destino, ruta)
}
