package graph

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"github.com/jmorenobl/kitlegal/internal/core"
	"github.com/jmorenobl/kitlegal/internal/core/grafo"
)

const (
	// patronDelTemporal es el nombre del fichero en que se construye world.db
	// antes de publicarlo: os.CreateTemp cambia el asterisco por un número, y
	// ningún lector ni ninguna entrega mira un fichero con ese nombre.
	patronDelTemporal = ficheroDelGrafo + "-nuevo-*"

	// permisosDeDirectorio son los de cada directorio que crea la entrega: el
	// grafo es de la cuenta, como la caché (research.md D8).
	permisosDeDirectorio fs.FileMode = 0o700
)

// errTemporalIncompleto es el temporal que, cerrado, sigue teniendo su -wal: lo
// confirmado en él no estaría en el fichero que se publica.
var errTemporalIncompleto = errors.New("el temporal no quedó completo: su registro de escritura sigue junto a él")

// publicar es el paso 3 de la entrega, sobre un world.db que no existe
// (contracts/almacen-world-db.md §4; research.md D11, V36 C): el lote tiene que
// entrar en un grafo vacío, y world.db se construye entero —en WAL, con el
// esquema y el lote— en un temporal del mismo directorio, que se publica de una
// vez con un enlace duro. Así world.db aparece completo o no aparece (FR-003,
// FR-013), y nunca sustituye a uno que otra invocación haya publicado antes: en
// ese caso el lote se aplica sobre ese world.db, en su sitio.
//
// Si la entrega falla, no deja nada (FR-033; §4.1): ni world.db, ni el
// temporal ni sus auxiliares, ni ningún directorio que creó, salvo uno que otra
// invocación ya usa. Un fallo de esa limpieza se une a la causa, sin callarlo.
func (a *Almacen) publicar(ctx context.Context, ruta string, lote core.Lote, consolidado grafo.Consolidado) (err error) {
	if err := grafo.ValidarContraGrafoVacio(lote); err != nil {
		return errorDeLoteRechazado(ruta, err)
	}

	temporal, creados, err := crearElTemporal(ctx, ruta)

	defer func() {
		if err != nil {
			err = errors.Join(err, retirarDirectorios(ruta, creados))
		}
	}()

	if err != nil {
		return err
	}

	publicado, err := a.publicarElTemporal(ctx, ruta, temporal, consolidado)
	if err != nil || publicado {
		return err
	}

	return aplicarEnSuSitio(ctx, ruta, lote, consolidado)
}

// publicarElTemporal hace los pasos 3.4 a 3.7 con el temporal ya creado: lo
// cierra, construye en él world.db, mira el contexto y lo enlaza con el nombre
// de world.db. Dice si lo publicó; si otra invocación publicó world.db antes
// —os.Link da fs.ErrExist—, no es un fallo y no lo publica. Pase lo que pase,
// retira el temporal y sus auxiliares; world.db, si se publicó, es otro nombre
// del mismo fichero y se queda.
func (a *Almacen) publicarElTemporal(
	ctx context.Context, ruta string, temporal *os.File, consolidado grafo.Consolidado,
) (publicado bool, err error) {
	nombre := temporal.Name()

	defer func() { err = errors.Join(err, retirarElTemporal(ruta, nombre)) }()

	if err := temporal.Close(); err != nil {
		return false, errorDeEntradaSalida(operacionEscribir, ruta, err)
	}

	if err := construirEnElTemporal(ctx, ruta, nombre, consolidado); err != nil {
		return false, err
	}

	if err := ctx.Err(); err != nil {
		return false, errorDePlazo(operacionEscribir, ruta, err)
	}

	err = a.enlazador()(nombre, ruta)

	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, fs.ErrExist):
		return false, nil
	default:
		return false, errorDePublicacion(filepath.Dir(ruta), err)
	}
}

// crearElTemporal hace los pasos 3.2 y 3.3: crea cada directorio que falta y,
// en el de world.db, el temporal, en 0600. Si el temporal no se puede crear
// porque el directorio ya no existe —otra invocación que falló lo retiró
// vacío—, vuelve a crear lo que falte mientras el contexto siga vivo. Devuelve
// el temporal abierto y cada directorio que creó, en el orden en que los creó,
// también si falla, para que se retiren.
func crearElTemporal(ctx context.Context, ruta string) (*os.File, []string, error) {
	directorio := filepath.Dir(ruta)

	var creados []string

	for {
		nuevos, err := crearDirectorios(directorio)
		creados = append(creados, nuevos...)

		if err == nil {
			var temporal *os.File
			if temporal, err = os.CreateTemp(directorio, patronDelTemporal); err == nil {
				return temporal, creados, nil
			}
		}

		switch {
		case !errors.Is(err, fs.ErrNotExist) || !puedeRehacerse(directorio):
			return nil, creados, errorDeDirectorioNoEscribible(directorio, err)
		case ctx.Err() != nil:
			return nil, creados, errorDePlazo(operacionEscribir, ruta, conElErrorDelContexto(ctx, err))
		}
	}
}

// crearDirectorios crea cada directorio que falta hasta el dado, de uno en uno
// desde el primer antecesor que no existe, en 0700, y devuelve los que ha
// creado en el orden en que los creó. Uno que ya existe cuando va a crearlo
// —otra invocación lo creó entre medias— no se anota, porque retirarlo no le
// corresponde. Devuelve el primer fallo tal cual, junto con lo que llegó a crear.
func crearDirectorios(directorio string) ([]string, error) {
	var faltan []string

	for actual := directorio; ; {
		_, err := os.Stat(actual)
		if err == nil {
			break
		}

		padre := filepath.Dir(actual)
		if !ausente(err) || padre == actual {
			return nil, err
		}

		faltan = append(faltan, actual)
		actual = padre
	}

	var creados []string

	for _, falta := range slices.Backward(faltan) {
		err := os.Mkdir(falta, permisosDeDirectorio)

		switch {
		case err == nil:
			creados = append(creados, falta)
		case !errors.Is(err, fs.ErrExist):
			return creados, err
		}
	}

	return creados, nil
}

// puedeRehacerse dice si volver al paso 3.2 puede arreglar un «no existe» al
// crear el temporal: sí si el antecesor más cercano que existe —el
// directorio mismo incluido— es un directorio, porque lo que falta se puede
// crear en él; no si es un enlace simbólico sin destino o algo que no es un
// directorio, donde volver a intentarlo daría siempre lo mismo.
func puedeRehacerse(directorio string) bool {
	for actual := directorio; ; {
		_, err := os.Lstat(actual)
		if err == nil {
			info, err := os.Stat(actual)

			return err == nil && info.IsDir()
		}

		padre := filepath.Dir(actual)
		if !errors.Is(err, fs.ErrNotExist) || padre == actual {
			return false
		}

		actual = padre
	}
}

// construirEnElTemporal hace el paso 3.4: abre el temporal con la cadena de
// escritura, lo pone en WAL fuera de toda transacción y, en una sola, crea el
// esquema y aplica el lote sobre el grafo vacío; lo cierra y comprueba que el
// cierre llevó todo al fichero, sin dejar su -wal. Nadie más conoce el temporal,
// así que no espera a nadie. Los mensajes nombran world.db, que es lo que se
// está construyendo.
func construirEnElTemporal(ctx context.Context, ruta, temporal string, consolidado grafo.Consolidado) error {
	base, err := abrirConexion(operacionEscribir, ruta, cadenaDeEscritura(temporal))
	if err != nil {
		return err
	}

	err = fijarWAL(ctx, base, ruta)
	if err == nil {
		err = escribirLote(ctx, base, ruta, consolidado)
	}

	if err := errors.Join(err, cerrarTrasElFallo(base, operacionEscribir, ruta)); err != nil {
		return err
	}

	_, err = os.Stat(temporal + sufijoWAL)

	switch {
	case err == nil:
		return errorDeEntradaSalida(operacionEscribir, ruta, errTemporalIncompleto)
	case !errors.Is(err, fs.ErrNotExist):
		return errorDeEntradaSalida(operacionEscribir, ruta, err)
	}

	return nil
}

// retirarElTemporal retira el temporal y el -wal, el -shm o el diario que
// pudieran quedar junto a él (paso 3.7). Lo que ya no está no es un fallo;
// cualquier otro se devuelve nombrando world.db.
func retirarElTemporal(ruta, temporal string) error {
	var fallos []error

	for _, sufijo := range []string{"", sufijoWAL, sufijoMemoriaCompartida, sufijoDiario} {
		if err := os.Remove(temporal + sufijo); err != nil && !errors.Is(err, fs.ErrNotExist) {
			fallos = append(fallos, errorDeEntradaSalida(operacionEscribir, ruta, err))
		}
	}

	return errors.Join(fallos...)
}

// retirarDirectorios retira cada directorio que creó una entrega que falla, del
// más profundo al menos (paso 3.7). Uno que no está vacío —os.Remove da
// fs.ErrExist: ENOTEMPTY en Unix, ERROR_DIR_NOT_EMPTY en Windows— lo usa otra
// invocación y se queda; uno que ya no está tampoco es un fallo. Cualquier otro
// fallo se devuelve nombrando world.db.
func retirarDirectorios(ruta string, creados []string) error {
	var fallos []error

	for _, creado := range slices.Backward(creados) {
		err := os.Remove(creado)
		if err != nil && !errors.Is(err, fs.ErrExist) && !errors.Is(err, fs.ErrNotExist) {
			fallos = append(fallos, errorDeEntradaSalida(operacionEscribir, ruta, err))
		}
	}

	return errors.Join(fallos...)
}
