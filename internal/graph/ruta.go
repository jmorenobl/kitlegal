package graph

import (
	"errors"
	"path/filepath"

	"github.com/jmorenobl/kitlegal/internal/cache"
)

// ficheroDelGrafo es el nombre de la base del grafo del mundo dentro de su
// directorio (FR-001).
const ficheroDelGrafo = "world.db"

// Opcion ajusta dónde vive world.db. Se aplica al leer o al entregar, no al
// construir: un fallo de una opción llega como el de la ubicación, con la
// operación que la resolvía.
type Opcion func(*ajustes) error

// ajustes es lo que declaran las opciones.
type ajustes struct {
	// directorio es el de ConDirectorio; vacío si ninguna opción lo declaró.
	directorio string
}

// errDirectorioVacio es el fallo de ConDirectorio(""): una opción que declara
// el directorio sin ninguna ruta, que es «argumentos» y nunca la regla de la
// caché en silencio (contracts/almacen-world-db.md §1).
var errDirectorioVacio = errors.New("la opción ConDirectorio no lleva ninguna ruta")

// ConDirectorio sustituye a la regla de ubicación de la caché: world.db vive en
// el directorio dado, que se usa tal cual y sin consultar el entorno. La usan
// los tests, el applet graph y la preparación de evals. Con varias, gana la
// última; la cadena vacía es un fallo de clase «argumentos» al resolver.
func ConDirectorio(directorio string) Opcion {
	return func(a *ajustes) error {
		if directorio == "" {
			return errDirectorioVacio
		}

		a.directorio = directorio

		return nil
	}
}

// ubicar resuelve la ruta de world.db para una operación: dentro del directorio
// de ConDirectorio o, sin la opción, del de cache.Directorio(), con su misma
// regla y sus mismos errores (FR-001, FR-011; contracts/almacen-world-db.md
// §2). Consulta el entorno en cada llamada y no crea ni abre nada: quien lee o
// entrega la llama en ese instante, nunca al construir.
func ubicar(operacion string, opciones []Opcion) (string, error) {
	var declarados ajustes

	for _, opcion := range opciones {
		if err := opcion(&declarados); err != nil {
			return "", errorDeUbicacion(operacion, err)
		}
	}

	directorio := declarados.directorio
	if directorio == "" {
		deLaCache, err := cache.Directorio()
		if err != nil {
			return "", errorDeUbicacion(operacion, err)
		}

		directorio = deLaCache
	}

	return filepath.Join(directorio, ficheroDelGrafo), nil
}
