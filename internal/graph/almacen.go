package graph

import (
	"context"
	"os"
	"slices"

	"github.com/jmorenobl/kitlegal/internal/core"
)

// Almacen es el grafo del mundo sobre world.db: el core.GraphStore con el que
// el kernel entrega lo que observa una invocación (contracts/almacen-world-db.md
// §1 y §4; research.md D11). Guarda sus opciones y nada más: construirlo no
// resuelve la ruta ni abre nada, y cada Apply abre, aplica y cierra, así que
// entre dos entregas no queda nada abierto.
//
// Su valor cero, y un *Almacen nulo, son el almacén sin opciones, como Nuevo().
// Varias entregas, de este proceso o de otros, pueden ir a la vez contra el
// mismo world.db: cada una espera su turno por tramos que miran el contexto
// (FR-014).
type Almacen struct {
	// opciones son las de Nuevo, que se resuelven en cada entrega.
	opciones []Opcion
	// enlazar publica el temporal con el nombre de world.db. Es os.Link salvo
	// en las pruebas, que lo sustituyen para provocar las dos salidas que no
	// son el éxito: un world.db que otra invocación publicó antes y cualquier
	// otro fallo (publicar.go; plan.md, Complexity Tracking).
	enlazar func(origen, destino string) error
}

// Almacen es un core.GraphStore.
var _ core.GraphStore = (*Almacen)(nil)

// Nuevo construye el almacén con las opciones dadas, de las que guarda una
// copia. No resuelve la ruta ni abre nada: una opción que no da ningún
// directorio falla al entregar, no aquí (FR-001, FR-011, FR-026).
func Nuevo(opciones ...Opcion) *Almacen {
	return &Almacen{opciones: slices.Clone(opciones), enlazar: os.Link}
}

// Apply guarda el lote entero o nada, y aplicarlo otra vez no cambia nada
// (FR-022; contracts/almacen-world-db.md §4):
//
//  1. valida y consolida el lote sin tocar el disco (FR-024, FR-025) y mira el
//     contexto: con el plazo agotado no toca nada;
//  2. resuelve la ruta de world.db en este instante (FR-001, FR-011) y mira qué
//     hay: un directorio es inutilizable, sin tocar nada;
//  3. si no hay nada, construye world.db en un temporal y lo publica de una vez
//     (publicar.go);
//  4. si hay un fichero, lo aplica en su sitio (aplicar.go).
//
// Un fallo es un *Error con su clase, y lo que deja en el disco es lo que
// declara contracts/almacen-world-db.md §4.1 (FR-033).
func (a *Almacen) Apply(ctx context.Context, lote core.Lote) error {
	consolidado, err := consolidar(lote)
	if err != nil {
		return errorDeLoteRechazado("", err)
	}

	if err := ctx.Err(); err != nil {
		return errorDePlazo(operacionEscribir, "", err)
	}

	ruta, err := ubicar(operacionEscribir, a.susOpciones())
	if err != nil {
		return err
	}

	info, err := os.Stat(ruta)

	switch {
	case ausente(err):
		return a.publicar(ctx, ruta, lote, consolidado)
	case err != nil:
		return errorDeEntradaSalida(operacionEscribir, ruta, err)
	case info.IsDir():
		return errorEsDirectorio(operacionEscribir, ruta)
	}

	return aplicarEnSuSitio(ctx, ruta, lote, consolidado)
}

// susOpciones son las opciones del almacén; ninguna si es nulo.
func (a *Almacen) susOpciones() []Opcion {
	if a == nil {
		return nil
	}

	return a.opciones
}

// enlazador es con qué se publica el temporal: os.Link si el almacén es nulo o
// no se construyó con Nuevo.
func (a *Almacen) enlazador() func(origen, destino string) error {
	if a == nil || a.enlazar == nil {
		return os.Link
	}

	return a.enlazar
}
