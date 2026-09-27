package instalacion_test

import (
	"errors"
	"fmt"
	"maps"
	"path"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
)

// directorioDeTrabajo es el de toda invocación sobre el disco en memoria: una
// ruta relativa cuelga de él, como cuelga del directorio de trabajo real.
const directorioDeTrabajo = "/trabajo"

// homeDePrueba es el HOME de las invocaciones con -g.
const homeDePrueba = "/home/ana"

// maximoDeSaltos es cuántos enlaces sigue el disco en memoria al resolver una
// ruta antes de darla por un ciclo, los mismos que Linux antes de ELOOP.
const maximoDeSaltos = 40

// Los errores del disco en memoria, los de las llamadas que el sistema también
// rechazaría.
var (
	// errNoExiste es la ruta que no existe.
	errNoExiste = errors.New("no existe")
	// errNoEsDirectorio es la ruta que pasa por algo que no es un directorio,
	// o el directorio que se pide listar y no lo es.
	errNoEsDirectorio = errors.New("no es un directorio")
	// errCiclo es la ruta que no se resuelve sin seguir más de maximoDeSaltos
	// enlaces.
	errCiclo = errors.New("demasiados enlaces encadenados: un ciclo")
	// errNoEsRegular es el fichero que se pide abrir y no es un fichero
	// regular.
	errNoEsRegular = errors.New("no es un fichero regular")
)

// Los errores de las operaciones del Escritor y del Enlazador sobre el disco
// en memoria, los que el sistema también daría.
var (
	// errYaExiste es crear un directorio o un enlace donde ya hay algo.
	errYaExiste = errors.New("ya existe")
	// errNoVacio es retirar un directorio que tiene algo dentro.
	errNoVacio = errors.New("el directorio no está vacío")
	// errSinEnlaces es crear un enlace en un directorio cuyo sistema de
	// ficheros no los admite.
	errSinEnlaces = errors.New("este sistema de ficheros no admite enlaces simbólicos")
	// errEsDirectorio es retirar con rm, sin -r, un directorio real.
	errEsDirectorio = errors.New("es un directorio")
)

// errInyectado es el fallo de entrada y salida que se inyecta en una llamada.
var errInyectado = errors.New("fallo de entrada y salida inyectado")

// operacion es el nombre de un método del puerto Disco.
type operacion string

// Las cuatro operaciones del puerto Disco.
const (
	opExaminar operacion = "Examinar"
	opHuella   operacion = "Huella"
	opLeer     operacion = "Leer"
	opNombres  operacion = "Nombres"
)

// acceso es una llamada al puerto Disco: la operación y la ruta absoluta.
type acceso struct {
	operacion operacion
	ruta      string
}

// nodoEnMemoria es una entrada del disco en memoria: su tipo, sus bytes si es
// un fichero regular y su destino literal si es un enlace. Una tubería con
// nombre, un socket o un dispositivo es de tipo EntradaOtra.
type nodoEnMemoria struct {
	tipo      instalacion.TipoDeEntrada
	contenido []byte
	destino   string
}

// discoEnMemoria es el puerto Disco sobre un árbol en memoria, con la
// semántica de un sistema de ficheros POSIX: las entradas se guardan por su
// ruta absoluta, cada componente intermedio de una ruta se resuelve siguiendo
// los enlaces —con .. sobre el directorio físico, no sobre el texto— y el
// último no se sigue (Lstat); un enlace resuelve si seguirlo, con los que
// encadene, llega a algo que existe sin pasar de maximoDeSaltos enlaces.
//
// Además de hacer de disco, vigila al dominio: anota cada llamada con su ruta,
// deja inyectar un fallo en una operación sobre una ruta y anota como
// violación todo intento de abrir lo que no es un fichero regular —una
// tubería con nombre no terminaría nunca— o de listar lo que no es un
// directorio real (FR-028).
type discoEnMemoria struct {
	t           *testing.T
	nodos       map[string]nodoEnMemoria
	fallos      map[acceso]error
	accesos     []acceso
	violaciones []string
}

// El disco en memoria es un Disco.
var _ instalacion.Disco = (*discoEnMemoria)(nil)

// nuevoDiscoEnMemoria es un disco con la raíz y el directorio de trabajo, y
// nada más.
func nuevoDiscoEnMemoria(t *testing.T) *discoEnMemoria {
	t.Helper()

	d := &discoEnMemoria{
		t:      t,
		nodos:  map[string]nodoEnMemoria{"/": {tipo: instalacion.EntradaDirectorio}},
		fallos: map[acceso]error{},
	}
	d.directorio(directorioDeTrabajo)

	return d
}

// absoluta es ruta como la ve el disco: limpia y colgando de la raíz, o del
// directorio de trabajo si es relativa.
func absoluta(ruta string) string {
	if path.IsAbs(ruta) {
		return path.Clean(ruta)
	}

	return path.Join(directorioDeTrabajo, ruta)
}

// fichero deja en ruta un fichero regular con contenido, y crea como
// directorio real todo lo que le falte por encima.
func (d *discoEnMemoria) fichero(ruta, contenido string) {
	d.poner(ruta, nodoEnMemoria{tipo: instalacion.EntradaFichero, contenido: []byte(contenido)})
}

// tuberia deja en ruta una tubería con nombre.
func (d *discoEnMemoria) tuberia(ruta string) {
	d.poner(ruta, nodoEnMemoria{tipo: instalacion.EntradaOtra})
}

// enlace deja en ruta un enlace simbólico con ese destino literal, resuelva o
// no.
func (d *discoEnMemoria) enlace(ruta, destino string) {
	d.poner(ruta, nodoEnMemoria{tipo: instalacion.EntradaEnlace, destino: destino})
}

// directorio deja en ruta un directorio real con los que le falten por
// encima; si ya lo es, no cambia nada.
func (d *discoEnMemoria) directorio(ruta string) {
	d.crearDirectorios(absoluta(ruta))
}

// retirar quita la entrada de ruta y todo lo que cuelga de ella.
func (d *discoEnMemoria) retirar(ruta string) {
	abs := absoluta(ruta)

	for clave := range d.nodos {
		if clave == abs || strings.HasPrefix(clave, abs+"/") {
			delete(d.nodos, clave)
		}
	}
}

// fallar hace que op sobre ruta, y solo sobre ella, devuelva err.
func (d *discoEnMemoria) fallar(op operacion, ruta string, err error) {
	d.fallos[acceso{operacion: op, ruta: absoluta(ruta)}] = err
}

// poner sustituye lo que haya en ruta, con todo lo que cuelgue de ello, por
// nodo.
func (d *discoEnMemoria) poner(ruta string, nodo nodoEnMemoria) {
	abs := absoluta(ruta)

	d.crearDirectorios(path.Dir(abs))
	d.retirar(abs)
	d.nodos[abs] = nodo
}

// crearDirectorios crea abs y lo que le falte por encima, cada uno como
// directorio real. Que alguno exista y no lo sea es un error del test que
// prepara el disco, no del dominio.
func (d *discoEnMemoria) crearDirectorios(abs string) {
	if abs == "/" {
		return
	}

	d.crearDirectorios(path.Dir(abs))

	nodo, existe := d.nodos[abs]
	if !existe {
		d.nodos[abs] = nodoEnMemoria{tipo: instalacion.EntradaDirectorio}

		return
	}

	if nodo.tipo != instalacion.EntradaDirectorio {
		d.t.Fatalf("el disco en memoria no crea nada dentro de %s, que no es un directorio real", abs)
	}
}

// seguir resuelve abs como lo resuelve el sistema al abrirla, siguiendo todos
// los enlaces, también el último, y devuelve la ruta física de lo que
// encuentra; o errNoExiste, errNoEsDirectorio si pasa por algo que no es un
// directorio, o errCiclo.
func (d *discoEnMemoria) seguir(abs string) (string, error) {
	pendientes := elementos(abs)
	actual := "/"
	saltos := 0

	for len(pendientes) > 0 {
		componente := pendientes[0]
		pendientes = pendientes[1:]

		if componente == "." {
			continue
		}

		if componente == ".." {
			actual = path.Dir(actual)

			continue
		}

		siguiente := path.Join(actual, componente)

		nodo, existe := d.nodos[siguiente]

		switch {
		case !existe:
			return "", errNoExiste
		case nodo.tipo == instalacion.EntradaEnlace:
			saltos++
			if saltos > maximoDeSaltos {
				return "", errCiclo
			}

			if path.IsAbs(nodo.destino) {
				actual = "/"
			}

			pendientes = append(elementos(nodo.destino), pendientes...)
		case nodo.tipo == instalacion.EntradaDirectorio || len(pendientes) == 0:
			actual = siguiente
		default:
			return "", errNoEsDirectorio
		}
	}

	return actual, nil
}

// elementos son los de una ruta, sin los vacíos.
func elementos(ruta string) []string {
	return strings.FieldsFunc(ruta, func(r rune) bool { return r == '/' })
}

// lstat es la entrada de abs sin seguirla, con su directorio resuelto como lo
// resuelve el sistema: su ruta física y su nodo, o que no existe. Un
// directorio por encima que no existe es que abs no existe; uno que no es un
// directorio o un ciclo, un error.
func (d *discoEnMemoria) lstat(abs string) (string, nodoEnMemoria, bool, error) {
	if abs == "/" {
		return abs, d.nodos[abs], true, nil
	}

	padre, err := d.seguir(path.Dir(abs))
	if errors.Is(err, errNoExiste) {
		return "", nodoEnMemoria{}, false, nil
	}

	if err != nil {
		return "", nodoEnMemoria{}, false, err
	}

	if d.nodos[padre].tipo != instalacion.EntradaDirectorio {
		return "", nodoEnMemoria{}, false, errNoEsDirectorio
	}

	fisica := path.Join(padre, path.Base(abs))
	nodo, existe := d.nodos[fisica]

	return fisica, nodo, existe, nil
}

// acceder anota la llamada op sobre ruta y devuelve la ruta absoluta, o el
// fallo inyectado en ella.
func (d *discoEnMemoria) acceder(op operacion, ruta string) (string, error) {
	abs := absoluta(ruta)
	d.accesos = append(d.accesos, acceso{operacion: op, ruta: abs})

	if err, hay := d.fallos[acceso{operacion: op, ruta: abs}]; hay {
		return "", fmt.Errorf("%s %s: %w", op, ruta, err)
	}

	return abs, nil
}

// llamadas es cuántas veces se ha llamado a op, sobre cualquier ruta.
func (d *discoEnMemoria) llamadas(op operacion) int {
	n := 0

	for _, pedido := range d.accesos {
		if pedido.operacion == op {
			n++
		}
	}

	return n
}

// Examinar es la entrada de ruta sin seguirla: ausente, con su tipo y, si es
// un enlace, su destino literal y si resuelve.
func (d *discoEnMemoria) Examinar(ruta string) (instalacion.Entrada, error) {
	abs, err := d.acceder(opExaminar, ruta)
	if err != nil {
		return instalacion.Entrada{}, err
	}

	_, nodo, existe, err := d.lstat(abs)
	if err != nil {
		return instalacion.Entrada{}, fmt.Errorf("%s %s: %w", opExaminar, ruta, err)
	}

	if !existe {
		return instalacion.Entrada{Tipo: instalacion.EntradaAusente}, nil
	}

	entrada := instalacion.Entrada{Tipo: nodo.tipo}

	if nodo.tipo == instalacion.EntradaEnlace {
		_, err := d.seguir(abs)
		entrada.Destino = nodo.destino
		entrada.Resuelve = err == nil
	}

	return entrada, nil
}

// Huella es la de los bytes del fichero regular de ruta.
func (d *discoEnMemoria) Huella(ruta string) (string, error) {
	contenido, err := d.abrir(opHuella, ruta)
	if err != nil {
		return "", err
	}

	return instalacion.HuellaDe(contenido), nil
}

// Leer son los bytes del fichero regular de ruta, en una copia.
func (d *discoEnMemoria) Leer(ruta string) ([]byte, error) {
	return d.abrir(opLeer, ruta)
}

// abrir son los bytes de ruta si es un fichero regular. Abrir lo que no lo es
// es una violación de FR-028, además de un error.
func (d *discoEnMemoria) abrir(op operacion, ruta string) ([]byte, error) {
	abs, err := d.acceder(op, ruta)
	if err != nil {
		return nil, err
	}

	_, nodo, existe, err := d.lstat(abs)

	switch {
	case err != nil:
		return nil, fmt.Errorf("%s %s: %w", op, ruta, err)
	case !existe:
		return nil, fmt.Errorf("%s %s: %w", op, ruta, errNoExiste)
	case nodo.tipo != instalacion.EntradaFichero:
		d.violaciones = append(d.violaciones, fmt.Sprintf("%s abre %s, que no es un fichero regular", op, ruta))

		return nil, fmt.Errorf("%s %s: %w", op, ruta, errNoEsRegular)
	}

	return slices.Clone(nodo.contenido), nil
}

// Nombres son los de las entradas del directorio real de ruta, en orden.
// Listar lo que no es un directorio real es una violación de FR-028, además
// de un error.
func (d *discoEnMemoria) Nombres(ruta string) ([]string, error) {
	abs, err := d.acceder(opNombres, ruta)
	if err != nil {
		return nil, err
	}

	fisica, nodo, existe, err := d.lstat(abs)

	switch {
	case err != nil:
		return nil, fmt.Errorf("%s %s: %w", opNombres, ruta, err)
	case !existe:
		return nil, fmt.Errorf("%s %s: %w", opNombres, ruta, errNoExiste)
	case nodo.tipo != instalacion.EntradaDirectorio:
		d.violaciones = append(d.violaciones, fmt.Sprintf("%s lista %s, que no es un directorio real", opNombres, ruta))

		return nil, fmt.Errorf("%s %s: %w", opNombres, ruta, errNoEsDirectorio)
	}

	var nombres []string

	for clave := range d.nodos {
		if clave != "/" && path.Dir(clave) == fisica {
			nombres = append(nombres, path.Base(clave))
		}
	}

	slices.Sort(nombres)

	return nombres, nil
}

// noDirectorioPorEncima es una entrada por encima de abs y por debajo de raiz
// que existe y no es un directorio real, mirada sin seguir enlaces; vacía si
// no hay ninguna. Lo que hay por encima de raiz no se mira (FR-027).
func (d *discoEnMemoria) noDirectorioPorEncima(abs, raiz string) string {
	for dir := path.Dir(abs); strings.HasPrefix(dir, raiz+"/"); dir = path.Dir(dir) {
		if nodo, existe := d.nodos[dir]; existe && nodo.tipo != instalacion.EntradaDirectorio {
			return dir
		}
	}

	return ""
}

// exigirDiscoRespetado exige que el dominio no haya abierto lo que no es un
// fichero regular, ni listado lo que no es un directorio real, ni pedido nada
// por debajo de una entrada de raiz que no es un directorio real (FR-028):
// esa entrada es lo único que se nombra.
func exigirDiscoRespetado(t *testing.T, d *discoEnMemoria, raiz string) {
	t.Helper()

	assert.Empty(t, d.violaciones, "abrió o listó lo que no es un fichero regular o un directorio real (FR-028)")

	for _, pedido := range d.accesos {
		if entrada := d.noDirectorioPorEncima(pedido.ruta, raiz); entrada != "" {
			t.Errorf("%s %s pasa por %s, que no es un directorio real (FR-028)", pedido.operacion, pedido.ruta, entrada)
		}
	}
}

// exigirNoExaminadas exige que el dominio no haya pedido al disco ninguna de
// rutas ni nada de lo que cuelga de ellas.
func exigirNoExaminadas(t *testing.T, d *discoEnMemoria, rutas ...string) {
	t.Helper()

	for _, ruta := range rutas {
		abs := absoluta(ruta)

		for _, pedido := range d.accesos {
			if pedido.ruta == abs || strings.HasPrefix(pedido.ruta, abs+"/") {
				t.Errorf("%s %s: no tenía que examinarse nada de %s", pedido.operacion, pedido.ruta, ruta)
			}
		}
	}
}

// instantanea es todo lo que hay en el disco, entrada a entrada: con ella se
// comprueba que algo lo deja byte a byte igual, con las mismas entradas y los
// mismos destinos de enlace (FR-045, FR-048).
func (d *discoEnMemoria) instantanea() map[string]nodoEnMemoria {
	return maps.Clone(d.nodos)
}

// paraCrear resuelve ruta como la resuelve el sistema al crear o cambiar algo
// en ella: su directorio tiene que existir y ser un directorio, siguiendo los
// enlaces de encima, y la entrada misma no se sigue. Devuelve su ruta física y
// lo que hay en ella.
func (d *discoEnMemoria) paraCrear(op, ruta string) (string, nodoEnMemoria, bool, error) {
	abs := absoluta(ruta)

	padre, err := d.seguir(path.Dir(abs))
	if err == nil && d.nodos[padre].tipo != instalacion.EntradaDirectorio {
		err = errNoEsDirectorio
	}

	if err != nil {
		return "", nodoEnMemoria{}, false, fmt.Errorf("%s %s: %w", op, ruta, err)
	}

	fisica := path.Join(padre, path.Base(abs))
	nodo, existe := d.nodos[fisica]

	return fisica, nodo, existe, nil
}

// escritorEnMemoria es el puerto Escritor sobre el disco en memoria, con la
// semántica estricta del sistema en cada operación suelta: crear un
// directorio donde no hay nada, escribir un fichero regular donde no hay nada
// o hay otro, y retirar un fichero, un enlace sin seguirlo o un directorio
// vacío; los enlaces, con el Enlazador de la invocación. Lo que el sistema
// rechazaría es un error, así que un plan que pide algo imposible no pasa.
//
// Además anota cada operación que se le pide, en orden, y deja hacer fallar
// una de ellas, como falla un disco lleno o sin permiso: la operación número
// fallarEn, contando desde 1, no hace nada y devuelve errInyectado, el error
// del sistema; con fallarEn 0 no falla ninguna.
type escritorEnMemoria struct {
	disco     *discoEnMemoria
	enlazador *enlazadorEnMemoria
	fallarEn  int
	// operaciones son las pedidas, también la que falla: «crear <ruta>»,
	// «escribir <ruta>», «retirar <ruta>» o «enlazar <ruta> -> <destino>».
	operaciones []string
}

// El escritor en memoria es un Escritor.
var _ instalacion.Escritor = (*escritorEnMemoria)(nil)

// pedir anota operacion y devuelve errInyectado si es la que tiene que
// fallar.
func (e *escritorEnMemoria) pedir(operacion string) error {
	e.operaciones = append(e.operaciones, operacion)

	if len(e.operaciones) == e.fallarEn {
		return errInyectado
	}

	return nil
}

// CrearDirectorio crea el directorio real de ruta, donde no hay nada.
func (e *escritorEnMemoria) CrearDirectorio(ruta string) error {
	if err := e.pedir("crear " + ruta); err != nil {
		return err
	}

	fisica, _, existe, err := e.disco.paraCrear("CrearDirectorio", ruta)
	if err != nil {
		return err
	}

	if existe {
		return fmt.Errorf("CrearDirectorio %s: %w", ruta, errYaExiste)
	}

	e.disco.nodos[fisica] = nodoEnMemoria{tipo: instalacion.EntradaDirectorio}

	return nil
}

// EscribirFichero deja en ruta un fichero regular con una copia de contenido,
// donde no hay nada o sustituyendo a otro fichero regular.
func (e *escritorEnMemoria) EscribirFichero(ruta string, contenido []byte) error {
	if err := e.pedir("escribir " + ruta); err != nil {
		return err
	}

	fisica, nodo, existe, err := e.disco.paraCrear("EscribirFichero", ruta)
	if err != nil {
		return err
	}

	if existe && nodo.tipo != instalacion.EntradaFichero {
		return fmt.Errorf("EscribirFichero %s: %w", ruta, errNoEsRegular)
	}

	e.disco.nodos[fisica] = nodoEnMemoria{tipo: instalacion.EntradaFichero, contenido: slices.Clone(contenido)}

	return nil
}

// Retirar quita la entrada de ruta, sin seguirla, si es un fichero, un enlace
// o un directorio vacío.
func (e *escritorEnMemoria) Retirar(ruta string) error {
	if err := e.pedir("retirar " + ruta); err != nil {
		return err
	}

	fisica, nodo, existe, err := e.disco.paraCrear("Retirar", ruta)

	switch {
	case err != nil:
		return err
	case !existe:
		return fmt.Errorf("Retirar %s: %w", ruta, errNoExiste)
	case nodo.tipo == instalacion.EntradaDirectorio && e.disco.tieneDentro(fisica):
		return fmt.Errorf("Retirar %s: %w", ruta, errNoVacio)
	}

	delete(e.disco.nodos, fisica)

	return nil
}

// Enlazar crea el enlace con el Enlazador de la invocación.
func (e *escritorEnMemoria) Enlazar(destino, ruta string) error {
	if err := e.pedir("enlazar " + ruta + " -> " + destino); err != nil {
		return err
	}

	return e.enlazador.Enlazar(destino, ruta)
}

// rm retira la entrada de ruta como la retira rm, la orden de shell con la
// que empieza la orden de un hallazgo de doctor (FR-066): su directorio se
// resuelve siguiendo los enlaces de encima y la entrada misma no se sigue, así
// que un enlace se retira sin tocar su destino; un fichero, un enlace u otra
// entrada se retiran siempre, y un directorio real solo con recursivo (-r),
// con todo lo que cuelga de él, sin seguir ninguno de sus enlaces. Devuelve la
// ruta física de lo que retiró.
func (d *discoEnMemoria) rm(ruta string, recursivo bool) (string, error) {
	fisica, nodo, existe, err := d.paraCrear("rm", ruta)

	switch {
	case err != nil:
		return "", err
	case !existe:
		return "", fmt.Errorf("rm %s: %w", ruta, errNoExiste)
	case nodo.tipo == instalacion.EntradaDirectorio && !recursivo:
		return "", fmt.Errorf("rm %s: %w", ruta, errEsDirectorio)
	}

	d.retirar(fisica)

	return fisica, nil
}

// tieneDentro dice si hay alguna entrada dentro de la ruta física dir.
func (d *discoEnMemoria) tieneDentro(dir string) bool {
	for clave := range d.nodos {
		if clave != "/" && path.Dir(clave) == dir {
			return true
		}
	}

	return false
}

// temporal es un directorio del disco en memoria que no es de ningún ámbito
// de las pruebas, como el TMPDIR de una máquina real.
const temporal = "/tmp"

// admiteSiempre y admiteNunca son las respuestas de un creador de enlaces que
// funciona en cualquier directorio y de uno que no funciona en ninguno.
func admiteSiempre(string) bool { return true }

func admiteNunca(string) bool { return false }

// admiteFueraDe es la de un creador de enlaces que funciona en cualquier
// directorio que no está por debajo de raiz, y en ninguno que lo está: el
// ámbito vive en un sistema de ficheros sin enlaces y el temporal en otro que
// los admite (research.md D9).
func admiteFueraDe(raiz string) func(string) bool {
	return func(directorio string) bool {
		abs := absoluta(directorio)

		return abs != raiz && !strings.HasPrefix(abs, raiz+"/")
	}
}

// enlazadorEnMemoria es un Enlazador sintético sobre el disco en memoria:
// Disponible responde lo que diga sondea del directorio, sin tocar el disco,
// o err si lo tiene, y anota el directorio por el que se pregunta; Enlazar
// crea el enlace en el disco si enlaza admite su directorio, y si no falla
// como un sistema de ficheros sin enlaces. Un enlazador honesto sondea y
// enlaza con la misma respuesta.
type enlazadorEnMemoria struct {
	disco       *discoEnMemoria
	sondea      func(directorio string) bool
	enlaza      func(directorio string) bool
	err         error
	preguntados []string
}

// El enlazador en memoria es un Enlazador.
var _ instalacion.Enlazador = (*enlazadorEnMemoria)(nil)

// nuevoEnlazadorHonesto es el que sondea y enlaza con la respuesta de admite.
func nuevoEnlazadorHonesto(d *discoEnMemoria, admite func(string) bool) *enlazadorEnMemoria {
	return &enlazadorEnMemoria{disco: d, sondea: admite, enlaza: admite}
}

// Disponible anota directorio y responde lo que diga sondea, o err.
func (e *enlazadorEnMemoria) Disponible(directorio string) (bool, error) {
	e.preguntados = append(e.preguntados, directorio)

	if e.err != nil {
		return false, e.err
	}

	return e.sondea(directorio), nil
}

// Enlazar crea en ruta un enlace con ese destino literal, donde no hay nada,
// si enlaza admite su directorio.
func (e *enlazadorEnMemoria) Enlazar(destino, ruta string) error {
	fisica, _, existe, err := e.disco.paraCrear("Enlazar", ruta)

	switch {
	case err != nil:
		return err
	case !e.enlaza(path.Dir(absoluta(ruta))):
		return fmt.Errorf("Enlazar %s: %w", ruta, errSinEnlaces)
	case existe:
		return fmt.Errorf("Enlazar %s: %w", ruta, errYaExiste)
	}

	e.disco.nodos[fisica] = nodoEnMemoria{tipo: instalacion.EntradaEnlace, destino: destino}

	return nil
}

// aplicarEnMemoria lleva a cabo el plan con Aplicar sobre escritor, que no
// puede fallar, y devuelve las skills cuya entrada de host quedó en copia
// porque su enlace no se pudo crear (FR-024). Exige que la salida sea la
// prevista en el plan salvo en eso: cada una de esas entradas, que el plan
// preveía enlace, sale en copia.
func aplicarEnMemoria(t *testing.T, plan instalacion.Plan, escritor *escritorEnMemoria) []string {
	t.Helper()

	skills, err := instalacion.Aplicar(plan, escritor)
	require.NoError(t, err)
	require.Len(t, skills, len(plan.Skills), "una salida por cada skill pedida")

	var enCopia []string

	for i, skill := range skills {
		if !slices.Equal(skill.Enlaces, plan.Skills[i].Enlaces) {
			enCopia = append(enCopia, skill.Nombre)
		}
	}

	assert.Equal(t, enModoCopia(plan.Skills, enCopia...), skills,
		"la salida es la prevista, con la entrada de cada enlace que no se pudo crear en copia (FR-024)")

	return enCopia
}

// enModoCopia son las skills de la salida prevista con la entrada de host de
// cada una de nombres —un enlace previsto— en copia, en una lista nueva.
func enModoCopia(previstas []instalacion.SkillInstalada, nombres ...string) []instalacion.SkillInstalada {
	skills := make([]instalacion.SkillInstalada, 0, len(previstas))

	for _, skill := range previstas {
		enlaces := slices.Clone(skill.Enlaces)

		for i := range enlaces {
			if slices.Contains(nombres, skill.Nombre) && enlaces[i].Modo == instalacion.ModoEnlace {
				enlaces[i].Modo = instalacion.ModoCopia
			}
		}

		skill.Enlaces = enlaces
		skills = append(skills, skill)
	}

	return skills
}

// enlazadorDePrueba es un Enlazador sintético: a Disponible responde siempre
// disponible, o err si lo tiene, y anota el directorio por el que se pregunta;
// Enlazar solo anota la ruta del enlace, sin crear nada.
type enlazadorDePrueba struct {
	disponible  bool
	err         error
	preguntados []string
	enlazados   []string
}

// El enlazador de prueba es un Enlazador.
var _ instalacion.Enlazador = (*enlazadorDePrueba)(nil)

// Disponible anota directorio y responde lo que se le dijo.
func (e *enlazadorDePrueba) Disponible(directorio string) (bool, error) {
	e.preguntados = append(e.preguntados, directorio)

	return e.disponible, e.err
}

// Enlazar anota ruta.
func (e *enlazadorDePrueba) Enlazar(_, ruta string) error {
	e.enlazados = append(e.enlazados, ruta)

	return nil
}

// versionDePrueba es la del binario que hizo las instalaciones de las
// pruebas.
const versionDePrueba = "v0.1.0"

// empotradasDePrueba son las skills empotradas de las pruebas, las dos del
// binario con sus ficheros, en una lista nueva en cada llamada.
func empotradasDePrueba() []instalacion.SkillEmpotrada {
	return []instalacion.SkillEmpotrada{
		{Nombre: "boe-legislacion", Ficheros: []instalacion.FicheroEmpotrado{
			instalacion.NuevoFicheroEmpotrado("SKILL.md", []byte("# boe-legislacion\n")),
			instalacion.NuevoFicheroEmpotrado("references/normas.md", []byte("# Normas\n")),
		}},
		{Nombre: "legal-core", Ficheros: []instalacion.FicheroEmpotrado{
			instalacion.NuevoFicheroEmpotrado("SKILL.md", []byte("# legal-core\n")),
			instalacion.NuevoFicheroEmpotrado("references/jerarquia_normativa.md", []byte("# Jerarquía\n")),
			instalacion.NuevoFicheroEmpotrado("references/leyes_vertebrales.md", []byte("# Leyes\n")),
		}},
	}
}

// empotradaDePrueba es la skill empotrada de las pruebas que se llama nombre.
func empotradaDePrueba(t *testing.T, nombre string) instalacion.SkillEmpotrada {
	t.Helper()

	empotradas := empotradasDePrueba()

	i := slices.IndexFunc(empotradas, func(skill instalacion.SkillEmpotrada) bool { return skill.Nombre == nombre })
	require.GreaterOrEqual(t, i, 0, "%q no es una skill empotrada de las pruebas", nombre)

	return empotradas[i]
}

// instalada es una instalación hecha a mano en el disco en memoria, como la
// dejaría install, sobre la que cada caso cambia una sola cosa: los ficheros
// ya están en el disco y el manifiesto no se escribe hasta escribir.
type instalada struct {
	disco      *discoEnMemoria
	ambito     instalacion.Ambito
	manifiesto instalacion.Manifiesto
}

// instalarEn deja en el directorio neutro de ambito los ficheros empotrados de
// cada skill nombrada y los declara, sin entradas de host.
func instalarEn(d *discoEnMemoria, ambito instalacion.Ambito, nombres ...string) *instalada {
	i := &instalada{
		disco:      d,
		ambito:     ambito,
		manifiesto: instalacion.Manifiesto{Version: versionDePrueba, Skills: map[string]instalacion.SkillDeclarada{}},
	}

	d.directorio(ambito.Neutro())

	for _, nombre := range nombres {
		ficheros := map[string]string{}

		for _, fichero := range empotradaDePrueba(d.t, nombre).Ficheros {
			d.fichero(path.Join(ambito.RutaDeSkill(nombre), fichero.Ruta), string(fichero.Contenido))
			ficheros[nombre+"/"+fichero.Ruta] = fichero.Huella
		}

		i.manifiesto.Skills[nombre] = instalacion.SkillDeclarada{Version: versionDePrueba, Ficheros: ficheros}
	}

	return i
}

// instalarLocal es instalarEn en el ámbito local.
func instalarLocal(d *discoEnMemoria, nombres ...string) *instalada {
	return instalarEn(d, instalacion.NuevoAmbitoLocal(), nombres...)
}

// El directorio de skills de cada host, relativo a la raíz del ámbito, y el
// destino de su enlace hasta el directorio neutro (ADR 0025).
var (
	skillsDelHostDePrueba = map[string]string{
		"claude":      ".claude/skills",
		"antigravity": ".gemini/config/skills",
	}
	subidaDelHostDePrueba = map[string]string{
		"claude":      "../../.agents/skills/",
		"antigravity": "../../../.agents/skills/",
	}
)

// entradaDe es la entrada de la skill declarada en el host, o nil si no tiene.
func entradaDe(skill instalacion.SkillDeclarada, host string) *instalacion.EntradaDeHost {
	entrada, hay := skill.Hosts[host]
	if !hay {
		return nil
	}

	return &entrada
}

// conEntrada es skill con entrada en el host, en un mapa de hosts nuevo.
func conEntrada(skill instalacion.SkillDeclarada, host string, entrada instalacion.EntradaDeHost) instalacion.SkillDeclarada {
	hosts := maps.Clone(skill.Hosts)
	if hosts == nil {
		hosts = map[string]instalacion.EntradaDeHost{}
	}

	hosts[host] = entrada
	skill.Hosts = hosts

	return skill
}

// cambiarEnClaude aplica cambio a la entrada en claude de la skill del
// manifiesto, que tiene que tenerla, y la deja en su sitio.
func cambiarEnClaude(m *instalacion.Manifiesto, skill string, cambio func(*instalacion.EntradaDeHost)) {
	entrada := m.Skills[skill].Hosts["claude"]
	cambio(&entrada)
	m.Skills[skill].Hosts["claude"] = entrada
}

// enClaude son las entradas de host de una skill que solo tiene la de claude.
func enClaude(entrada instalacion.EntradaDeHost) map[string]instalacion.EntradaDeHost {
	return map[string]instalacion.EntradaDeHost{"claude": entrada}
}

// enlazar es enlazarEn en el host claude.
func (i *instalada) enlazar(nombre string) *instalada {
	return i.enlazarEn("claude", nombre)
}

// enlazarEn deja la entrada de la skill nombre en el host como el enlace de
// FR-021 y la declara en modo enlace.
func (i *instalada) enlazarEn(host, nombre string) *instalada {
	i.disco.enlace(i.ambito.RutaDeHost(host, nombre), subidaDelHostDePrueba[host]+nombre)

	i.manifiesto.Skills[nombre] = conEntrada(i.manifiesto.Skills[nombre], host, instalacion.EntradaDeHost{
		Ruta: skillsDelHostDePrueba[host] + "/" + nombre,
		Modo: instalacion.ModoEnlace,
	})

	return i
}

// copiar es copiarEn en el host claude.
func (i *instalada) copiar(nombre string) *instalada {
	return i.copiarEn("claude", nombre)
}

// copiarEn deja la entrada de la skill nombre en el host como la copia de
// FR-024, con los ficheros empotrados, y la declara en modo copia con sus
// huellas.
func (i *instalada) copiarEn(host, nombre string) *instalada {
	ficheros := map[string]string{}
	ruta := skillsDelHostDePrueba[host] + "/" + nombre

	for _, fichero := range empotradaDePrueba(i.disco.t, nombre).Ficheros {
		i.disco.fichero(path.Join(i.ambito.RutaDeHost(host, nombre), fichero.Ruta), string(fichero.Contenido))
		ficheros[ruta+"/"+fichero.Ruta] = fichero.Huella
	}

	i.manifiesto.Skills[nombre] = conEntrada(i.manifiesto.Skills[nombre], host, instalacion.EntradaDeHost{
		Ruta:     ruta,
		Modo:     instalacion.ModoCopia,
		Ficheros: ficheros,
	})

	return i
}

// declarar añade al manifiesto, en la skill nombre, el fichero ruta —relativo
// al directorio neutro— con la huella de contenido, sin tocar el disco.
func (i *instalada) declarar(nombre, ruta, contenido string) *instalada {
	i.manifiesto.Skills[nombre].Ficheros[ruta] = instalacion.HuellaDe([]byte(contenido))

	return i
}

// olvidar retira del manifiesto la declaración de ruta en la skill nombre o
// en sus copias de host, sin tocar el disco.
func (i *instalada) olvidar(nombre, ruta string) *instalada {
	skill := i.manifiesto.Skills[nombre]
	delete(skill.Ficheros, ruta)

	for _, entrada := range skill.Hosts {
		delete(entrada.Ficheros, ruta)
	}

	return i
}

// deVersion hace de version el manifiesto y cada skill que declara, como si
// las hubiera puesto ahí otro binario.
func (i *instalada) deVersion(version string) *instalada {
	i.manifiesto.Version = version

	for nombre, skill := range i.manifiesto.Skills {
		skill.Version = version
		i.manifiesto.Skills[nombre] = skill
	}

	return i
}

// skillDeVersion hace de version solo la skill nombre, como la habría dejado
// un install de esa skill sola con otro binario.
func (i *instalada) skillDeVersion(nombre, version string) *instalada {
	skill := i.manifiesto.Skills[nombre]
	skill.Version = version
	i.manifiesto.Skills[nombre] = skill

	return i
}

// noEmpotrada deja en el directorio neutro la skill nombre, que las pruebas no
// empotran, con un SKILL.md, y la declara de version con una entrada de host
// en enlace que no está en el disco: la de una skill que puso ahí un binario
// de otra versión o de otra rama (FR-036).
func (i *instalada) noEmpotrada(nombre, version string) *instalada {
	contenido := "# " + nombre + "\n"
	i.disco.fichero(path.Join(i.ambito.RutaDeSkill(nombre), "SKILL.md"), contenido)
	i.manifiesto.Skills[nombre] = instalacion.SkillDeclarada{
		Version:  version,
		Ficheros: map[string]string{nombre + "/SKILL.md": instalacion.HuellaDe([]byte(contenido))},
		Hosts:    enClaude(instalacion.EntradaDeHost{Ruta: ".claude/skills/" + nombre, Modo: instalacion.ModoEnlace}),
	}

	return i
}

// deOtroBinario deja en el directorio de la skill nombre el fichero rel con
// contenido y lo declara con esa huella, como lo habría dejado un binario que
// empotraba otro contenido.
func (i *instalada) deOtroBinario(nombre, rel, contenido string) *instalada {
	i.disco.fichero(path.Join(i.ambito.RutaDeSkill(nombre), rel), contenido)
	i.manifiesto.Skills[nombre].Ficheros[nombre+"/"+rel] = instalacion.HuellaDe([]byte(contenido))

	return i
}

// copiaDeOtroBinario hace lo mismo que deOtroBinario en la copia de host de la
// skill nombre en claude, que tiene que estar declarada.
func (i *instalada) copiaDeOtroBinario(nombre, rel, contenido string) *instalada {
	i.disco.fichero(path.Join(i.ambito.RutaDeHost("claude", nombre), rel), contenido)
	i.manifiesto.Skills[nombre].Hosts["claude"].Ficheros[".claude/skills/"+nombre+"/"+rel] = instalacion.HuellaDe([]byte(contenido))

	return i
}

// escribir deja el manifiesto en el disco, en su forma canónica.
func (i *instalada) escribir() {
	contenido, err := i.manifiesto.Bytes()
	require.NoError(i.disco.t, err, "el manifiesto de la instalación hecha a mano")

	i.disco.fichero(i.ambito.RutaDelManifiesto(), string(contenido))
}

// TestDiscoEnMemoria fija el doble de prueba del que dependen las pruebas del
// dominio: que Examinar ve cada entrada como Lstat, sin seguir el último
// componente y siguiendo los de encima; que un enlace resuelve si su cadena
// llega a algo que existe, con .. sobre el directorio físico, y no resuelve si
// cuelga o está en un ciclo; que Huella, Leer y Nombres solo abren un fichero
// regular y listan un directorio real, y anotan como violación lo demás; y que
// un fallo inyectado sale solo en su operación y su ruta.
func TestDiscoEnMemoria(t *testing.T) {
	t.Parallel()

	t.Run("examinar", probarExaminarEnMemoria)
	t.Run("abrir y listar", probarAbrirEnMemoria)
	t.Run("fallos inyectados", probarFallosEnMemoria)
	t.Run("escribir", probarEscribirEnMemoria)
	t.Run("fallos del escritor", probarFallosDelEscritorEnMemoria)
	t.Run("enlazar", probarEnlazarEnMemoria)
	t.Run("rm", probarRmEnMemoria)
}

// discoDeMuestra es un disco con una entrada de cada clase.
func discoDeMuestra(t *testing.T) *discoEnMemoria {
	t.Helper()

	d := nuevoDiscoEnMemoria(t)
	d.fichero("dir/f", "contenido")
	d.directorio("dir/sub")
	d.tuberia("tuberia")
	d.enlace("enlace-a-dir", "dir")
	d.enlace("cadena", "enlace-a-dir")
	d.enlace("absoluto", "/trabajo/dir/f")
	d.enlace("colgando", "no-existe")
	d.enlace("ciclo", "ciclo")
	d.enlace("a", "b")
	d.enlace("b", "a")
	d.enlace("profundo", "dir/sub")
	d.enlace("fisico", "profundo/../f")

	return d
}

func probarExaminarEnMemoria(t *testing.T) {
	t.Parallel()

	casos := []struct {
		ruta     string
		esperada instalacion.Entrada
	}{
		{ruta: "no-existe", esperada: instalacion.Entrada{Tipo: instalacion.EntradaAusente}},
		{ruta: "no-existe/x", esperada: instalacion.Entrada{Tipo: instalacion.EntradaAusente}},
		{ruta: "/", esperada: instalacion.Entrada{Tipo: instalacion.EntradaDirectorio}},
		{ruta: ".", esperada: instalacion.Entrada{Tipo: instalacion.EntradaDirectorio}},
		{ruta: "dir", esperada: instalacion.Entrada{Tipo: instalacion.EntradaDirectorio}},
		{ruta: "dir/f", esperada: instalacion.Entrada{Tipo: instalacion.EntradaFichero}},
		{ruta: "tuberia", esperada: instalacion.Entrada{Tipo: instalacion.EntradaOtra}},
		{ruta: "enlace-a-dir/f", esperada: instalacion.Entrada{Tipo: instalacion.EntradaFichero}},
		{ruta: "enlace-a-dir", esperada: enlaceEnMemoria("dir", true)},
		{ruta: "cadena", esperada: enlaceEnMemoria("enlace-a-dir", true)},
		{ruta: "absoluto", esperada: enlaceEnMemoria("/trabajo/dir/f", true)},
		{ruta: "fisico", esperada: enlaceEnMemoria("profundo/../f", true)},
		{ruta: "colgando", esperada: enlaceEnMemoria("no-existe", false)},
		{ruta: "ciclo", esperada: enlaceEnMemoria("ciclo", false)},
		{ruta: "a", esperada: enlaceEnMemoria("b", false)},
	}

	for _, caso := range casos {
		t.Run(caso.ruta, func(t *testing.T) {
			t.Parallel()

			entrada, err := discoDeMuestra(t).Examinar(caso.ruta)
			require.NoError(t, err)
			assert.Equal(t, caso.esperada, entrada)
		})
	}

	for ruta, esperado := range map[string]error{"dir/f/x": errNoEsDirectorio, "ciclo/x": errCiclo} {
		t.Run(ruta, func(t *testing.T) {
			t.Parallel()

			_, err := discoDeMuestra(t).Examinar(ruta)
			require.ErrorIs(t, err, esperado)
		})
	}
}

// enlaceEnMemoria es la entrada de un enlace con ese destino literal.
func enlaceEnMemoria(destino string, resuelve bool) instalacion.Entrada {
	return instalacion.Entrada{Tipo: instalacion.EntradaEnlace, Destino: destino, Resuelve: resuelve}
}

func probarAbrirEnMemoria(t *testing.T) {
	t.Parallel()

	d := discoDeMuestra(t)

	huella, err := d.Huella("dir/f")
	require.NoError(t, err)
	assert.Equal(t, instalacion.HuellaDe([]byte("contenido")), huella)

	leido, err := d.Leer("/trabajo/dir/f")
	require.NoError(t, err)
	assert.Equal(t, "contenido", string(leido))

	leido[0] = 'X'
	leido, err = d.Leer("dir/f")
	require.NoError(t, err)
	assert.Equal(t, "contenido", string(leido), "Leer da una copia")

	nombres, err := d.Nombres("dir")
	require.NoError(t, err)
	assert.Equal(t, []string{"f", "sub"}, nombres)

	_, err = d.Huella("no-existe")
	require.ErrorIs(t, err, errNoExiste)
	assert.Empty(t, d.violaciones, "lo que no existe no se abre: no es una violación")

	_, err = d.Huella("tuberia")
	require.ErrorIs(t, err, errNoEsRegular)
	_, err = d.Leer("enlace-a-dir")
	require.ErrorIs(t, err, errNoEsRegular)
	_, err = d.Nombres("enlace-a-dir")
	require.ErrorIs(t, err, errNoEsDirectorio)
	assert.Len(t, d.violaciones, 3, "abrir una tubería o un enlace y listar un enlace: %q", d.violaciones)

	assert.Equal(t, acceso{operacion: opHuella, ruta: "/trabajo/dir/f"}, d.accesos[0], "cada llamada, anotada")
	assert.Len(t, d.accesos, 8)
	assert.Equal(t, 3, d.llamadas(opHuella), "las de cada operación, sobre cualquier ruta")
	assert.Equal(t, 2, d.llamadas(opNombres))
	assert.Zero(t, d.llamadas(opExaminar))
}

func probarFallosEnMemoria(t *testing.T) {
	t.Parallel()

	d := discoDeMuestra(t)
	d.fallar(opExaminar, "dir", errInyectado)
	d.fallar(opHuella, "dir/f", errInyectado)
	d.fallar(opNombres, "/trabajo/dir", errInyectado)

	_, err := d.Examinar("dir")
	require.ErrorIs(t, err, errInyectado)
	_, err = d.Huella("dir/f")
	require.ErrorIs(t, err, errInyectado)
	_, err = d.Nombres("dir")
	require.ErrorIs(t, err, errInyectado)

	entrada, err := d.Examinar("dir/f")
	require.NoError(t, err, "el fallo es solo de su operación y su ruta")
	assert.Equal(t, instalacion.EntradaFichero, entrada.Tipo)

	_, err = d.Leer("dir/f")
	require.NoError(t, err, "el fallo es solo de su operación y su ruta")
}

// probarEscribirEnMemoria fija las operaciones del Escritor en memoria: cada
// una hace lo mismo que el sistema y rechaza lo que el sistema rechazaría.
func probarEscribirEnMemoria(t *testing.T) {
	t.Parallel()

	d := discoDeMuestra(t)
	e := &escritorEnMemoria{disco: d, enlazador: nuevoEnlazadorHonesto(d, admiteNunca)}

	require.NoError(t, e.CrearDirectorio("nuevo"))
	require.NoError(t, e.EscribirFichero("nuevo/f", []byte("uno")))
	require.NoError(t, e.EscribirFichero("nuevo/f", []byte("dos")), "sustituye a un fichero regular")
	require.NoError(t, e.EscribirFichero("enlace-a-dir/g", []byte("tres")), "los enlaces de encima se siguen")

	leido, err := d.Leer("dir/g")
	require.NoError(t, err)
	assert.Equal(t, "tres", string(leido))

	require.ErrorIs(t, e.CrearDirectorio("dir"), errYaExiste)
	require.ErrorIs(t, e.CrearDirectorio("no-existe/x"), errNoExiste, "el padre tiene que existir")
	require.ErrorIs(t, e.EscribirFichero("dir/f/x", nil), errNoEsDirectorio)
	require.ErrorIs(t, e.EscribirFichero("dir/sub", nil), errNoEsRegular, "no sustituye a un directorio")
	require.ErrorIs(t, e.EscribirFichero("colgando", nil), errNoEsRegular, "no escribe a través de un enlace")
	require.ErrorIs(t, e.Retirar("dir"), errNoVacio)
	require.ErrorIs(t, e.Retirar("no-existe"), errNoExiste)

	require.NoError(t, e.Retirar("enlace-a-dir"), "retira el enlace, no lo que hay al otro lado")
	require.NoError(t, e.Retirar("dir/sub"), "un directorio vacío")

	entrada, err := d.Examinar("dir/f")
	require.NoError(t, err)
	assert.Equal(t, instalacion.EntradaFichero, entrada.Tipo)

	entrada, err = d.Examinar("dir/sub")
	require.NoError(t, err)
	assert.Equal(t, instalacion.EntradaAusente, entrada.Tipo)
}

// probarFallosDelEscritorEnMemoria fija cómo falla el Escritor en memoria:
// anota cada operación pedida, en orden, y la que tiene que fallar devuelve
// errInyectado sin hacer nada; las demás, antes y después, se hacen.
func probarFallosDelEscritorEnMemoria(t *testing.T) {
	t.Parallel()

	d := nuevoDiscoEnMemoria(t)
	e := &escritorEnMemoria{disco: d, enlazador: nuevoEnlazadorHonesto(d, admiteSiempre), fallarEn: 2}

	require.NoError(t, e.CrearDirectorio("a"))
	require.ErrorIs(t, e.EscribirFichero("a/f", []byte("uno")), errInyectado)
	assert.Equal(t, instalacion.EntradaAusente, examinarEnMemoria(t, d, "a/f").Tipo, "la que falla no hace nada")

	require.NoError(t, e.Enlazar("f", "a/enlace"))
	require.NoError(t, e.Retirar("a/enlace"))
	assert.Equal(t, instalacion.EntradaAusente, examinarEnMemoria(t, d, "a/enlace").Tipo)

	assert.Equal(t, []string{"crear a", "escribir a/f", "enlazar a/enlace -> f", "retirar a/enlace"}, e.operaciones)
}

// probarEnlazarEnMemoria fija el Enlazador en memoria: responde lo que diga
// sondea, anotando el directorio, y solo crea el enlace donde enlaza lo
// admite.
func probarEnlazarEnMemoria(t *testing.T) {
	t.Parallel()

	d := nuevoDiscoEnMemoria(t)
	d.directorio(temporal)

	enlazador := nuevoEnlazadorHonesto(d, admiteFueraDe(directorioDeTrabajo))

	fuera, err := enlazador.Disponible(temporal)
	require.NoError(t, err)
	assert.True(t, fuera, "fuera del ámbito admite enlaces")

	dentro, err := enlazador.Disponible(".")
	require.NoError(t, err)
	assert.False(t, dentro, "dentro del ámbito, no")
	assert.Equal(t, []string{temporal, "."}, enlazador.preguntados)

	require.NoError(t, enlazador.Enlazar("../x", temporal+"/enlace"))
	assert.Equal(t, enlaceEnMemoria("../x", false), examinarEnMemoria(t, d, temporal+"/enlace"))

	require.ErrorIs(t, enlazador.Enlazar("x", "enlace"), errSinEnlaces)
	require.ErrorIs(t, enlazador.Enlazar("x", temporal+"/enlace"), errYaExiste)
	assert.Equal(t, instalacion.EntradaAusente, examinarEnMemoria(t, d, "enlace").Tipo)

	enlazador.err = errInyectado
	_, err = enlazador.Disponible(temporal)
	require.ErrorIs(t, err, errInyectado)
}

// probarRmEnMemoria fija rm sobre el disco en memoria, la de las órdenes de
// doctor: retira un fichero, una tubería o un enlace sin seguirlo, y un
// directorio real solo con -r y entero; lo que no existe es un error, y el
// destino de un enlace retirado queda como estaba.
func probarRmEnMemoria(t *testing.T) {
	t.Parallel()

	d := discoDeMuestra(t)

	fisica, err := d.rm("enlace-a-dir/f", false)
	require.NoError(t, err, "los enlaces de encima se siguen")
	assert.Equal(t, "/trabajo/dir/f", fisica)

	fisica, err = d.rm("enlace-a-dir", false)
	require.NoError(t, err)
	assert.Equal(t, "/trabajo/enlace-a-dir", fisica)
	assert.Equal(t, instalacion.EntradaAusente, examinarEnMemoria(t, d, "enlace-a-dir").Tipo)
	assert.Equal(t, instalacion.EntradaDirectorio, examinarEnMemoria(t, d, "dir/sub").Tipo, "el destino, intacto")

	_, err = d.rm("tuberia", false)
	require.NoError(t, err)

	_, err = d.rm("dir", false)
	require.ErrorIs(t, err, errEsDirectorio, "sin -r no retira un directorio")
	assert.Equal(t, instalacion.EntradaDirectorio, examinarEnMemoria(t, d, "dir/sub").Tipo)

	_, err = d.rm("dir", true)
	require.NoError(t, err)
	assert.Equal(t, instalacion.EntradaAusente, examinarEnMemoria(t, d, "dir/sub").Tipo, "con -r, entero")

	_, err = d.rm("no-existe", false)
	require.ErrorIs(t, err, errNoExiste)

	_, err = d.rm("colgando/x", true)
	require.ErrorIs(t, err, errNoExiste)
}

// examinarEnMemoria es la entrada de ruta en el disco, sin error.
func examinarEnMemoria(t *testing.T, d *discoEnMemoria, ruta string) instalacion.Entrada {
	t.Helper()

	entrada, err := d.Examinar(ruta)
	require.NoError(t, err)

	return entrada
}
