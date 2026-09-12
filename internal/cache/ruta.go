package cache

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
)

// ficheroDeLaBase es el nombre de la base de datos dentro del directorio
// efectivo (FR-020). No se exporta —nada de la ruta se exporta— y el contrato
// lo fija, de modo que los tests externos escriben el literal (D2).
const ficheroDeLaBase = "cache.db"

// origenDeLaRuta es de dónde salió el directorio de la caché: las tres únicas
// procedencias posibles, y cada constante es ya la frase con la que los
// mensajes la nombran, que es el vocabulario cerrado del campo Origen de Error
// (data-model §8).
//
// Nombrarla no es un adorno: es lo que distingue «la opción que trae el código
// está mal» de «lo que hay en el entorno está mal», y sin eso quien lee el
// fallo no sabe dónde corregirlo (FR-022, FR-035).
type origenDeLaRuta string

const (
	origenOpcion   origenDeLaRuta = "opción ConDirectorio"
	origenVariable origenDeLaRuta = "variable " + VariableDirectorio
	origenOmision  origenDeLaRuta = "ruta por omisión"
)

// rutaEfectiva resuelve el directorio de la caché y dice de dónde salió, en el
// orden que fija FR-023: la opción si se declaró, la variable de entorno si
// está presente y, si no hay ninguna de las dos, el directorio de la cuenta
// (FR-019, al pie de la letra: $HOME en Unix y macOS, %USERPROFILE% en
// Windows).
//
// La variable se consulta **una sola vez**, aquí, al construir el cliente, y
// quien llama pasa cómo consultarla: así esa única lectura es comprobable y la
// opción, cuando está, no llega siquiera a mirar el entorno. Presente y vacía
// devuelve la cadena vacía con su origen y nunca la ruta por omisión, porque la
// caída silenciosa está prohibida (FR-022); rechazarla es cosa de compruebaRuta,
// que es la validación que vale en cualquier modo.
//
// Un directorio vacío significa aquí «no se declaró la opción»: ConDirectorio
// rechaza la cadena vacía antes, al aplicarse.
func rutaEfectiva(directorio string, buscaEnElEntorno func(string) (string, bool)) (string, origenDeLaRuta, error) {
	if directorio != "" {
		return directorio, origenOpcion, nil
	}

	if valor, declarada := buscaEnElEntorno(VariableDirectorio); declarada {
		return valor, origenVariable, nil
	}

	cuenta, err := os.UserHomeDir()
	if err != nil {
		return "", origenOmision, errorDeRutaPorOmisionIndeterminable(err)
	}

	return filepath.Join(cuenta, ".cache", "kitlegal"), origenOmision, nil
}

// compruebaRuta valida el valor del directorio efectivo, y lo hace igual en
// cualquier modo: un valor vacío y una ruta que existe y no es un directorio
// son «argumentos» (2) tanto si se va a escribir como si solo se va a leer
// (FR-015, FR-022).
//
// El os.Stat de aquí es el único que este paquete hace sobre el directorio
// efectivo (D3). Que falle no es un fallo de validación: o el directorio no
// está —y en modo normal se crea, y en solo lectura es una ausencia— o no se
// sabe si está; las dos ramas dependen del modo y las decide quien abre la base
// de datos, clasificando el fallo con esInexistente.
func compruebaRuta(directorio string, de origenDeLaRuta) error {
	if directorio == "" {
		return errorDeRutaVacia("construir", de)
	}

	info, err := os.Stat(directorio)
	if err == nil && !info.IsDir() {
		return errorDeRutaInservible("construir", string(de), directorio, nil)
	}

	return nil
}

// esInexistente es la regla «inexistente» del contrato del puerto §3, la única
// que decide qué fallo de os.Stat significa que algo no está: lo significa si y
// solo si el error equivale a fs.ErrNotExist —no hay nada en la ruta— o a
// syscall.ENOTDIR —un componente de la ruta, el padre, es un fichero, de modo
// que el directorio no existe *como directorio*—.
//
// Hacen falta las dos condiciones: en Go 1.27 syscall.Errno.Is solo reconoce
// ENOENT como fs.ErrNotExist, así que sin nombrar ENOTDIR el directorio cuyo
// padre es un fichero caería en «otro error» y terminaría en un fallo (1) donde
// FR-015 exige tratarlo exactamente como un cache.db inexistente (4). En
// Windows ENOTDIR es ERROR_PATH_NOT_FOUND, que Errno.Is ya trata como
// fs.ErrNotExist, y la regla vale igual en las tres plataformas del binario.
//
// Todo lo demás —el acceso denegado, un fallo de entrada y salida— es «otro»:
// no se sabe si lo que se busca está, y qué hacer con eso depende del modo. No
// hay una tercera rama, y por eso ninguna forma de fallo queda sin clase
// (FR-033, D3).
//
// Se define una vez, aquí, y es lo único para lo que este paquete importa
// syscall.
func esInexistente(err error) bool {
	return errors.Is(err, fs.ErrNotExist) || errors.Is(err, syscall.ENOTDIR)
}

// errorDeRutaVacia es la fila 4 de la tabla del contrato de errores §3, y la
// mitad de la 3: el valor que declara dónde vive la caché no lleva nada dentro.
// El mensaje nombra el origen —la opción o la variable—, que es lo único que
// hay que corregir, y la ruta no se nombra porque no la hay (FR-022).
func errorDeRutaVacia(operacion string, de origenDeLaRuta) *Error {
	fallo := errorDeArgumentos(operacion, fmt.Sprintf("la %s no lleva ninguna ruta", de), nil)
	fallo.Origen = string(de)

	return fallo
}

// errorDeRutaPorOmisionIndeterminable es la fila 7: no hay opción ni variable y
// tampoco se sabe cuál es el directorio de la cuenta. El mensaje nombra las dos
// maneras de arreglarlo, que son las dos que quien invoca tiene a mano (D3).
func errorDeRutaPorOmisionIndeterminable(causa error) *Error {
	fallo := errorDeArgumentos("construir",
		"no se sabe cuál es el directorio de la cuenta: declara HOME o "+VariableDirectorio, causa)
	fallo.Origen = string(origenOmision)

	return fallo
}
