package instalacion_test

import (
	"errors"
	"fmt"
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

// enlazar deja la entrada de host de la skill nombre como el enlace de FR-021
// y la declara en modo enlace.
func (i *instalada) enlazar(nombre string) *instalada {
	i.disco.enlace(i.ambito.RutaDeHost(nombre), "../../.agents/skills/"+nombre)

	skill := i.manifiesto.Skills[nombre]
	skill.Claude = &instalacion.EntradaDeHost{Ruta: ".claude/skills/" + nombre, Modo: instalacion.ModoEnlace}
	i.manifiesto.Skills[nombre] = skill

	return i
}

// copiar deja la entrada de host de la skill nombre como la copia de FR-024,
// con los ficheros empotrados, y la declara en modo copia con sus huellas.
func (i *instalada) copiar(nombre string) *instalada {
	ficheros := map[string]string{}

	for _, fichero := range empotradaDePrueba(i.disco.t, nombre).Ficheros {
		i.disco.fichero(path.Join(i.ambito.RutaDeHost(nombre), fichero.Ruta), string(fichero.Contenido))
		ficheros[".claude/skills/"+nombre+"/"+fichero.Ruta] = fichero.Huella
	}

	skill := i.manifiesto.Skills[nombre]
	skill.Claude = &instalacion.EntradaDeHost{
		Ruta:     ".claude/skills/" + nombre,
		Modo:     instalacion.ModoCopia,
		Ficheros: ficheros,
	}
	i.manifiesto.Skills[nombre] = skill

	return i
}

// declarar añade al manifiesto, en la skill nombre, el fichero ruta —relativo
// al directorio neutro— con la huella de contenido, sin tocar el disco.
func (i *instalada) declarar(nombre, ruta, contenido string) *instalada {
	i.manifiesto.Skills[nombre].Ficheros[ruta] = instalacion.HuellaDe([]byte(contenido))

	return i
}

// olvidar retira del manifiesto la declaración de ruta en la skill nombre o
// en su copia de host, sin tocar el disco.
func (i *instalada) olvidar(nombre, ruta string) *instalada {
	skill := i.manifiesto.Skills[nombre]
	delete(skill.Ficheros, ruta)

	if skill.Claude != nil {
		delete(skill.Claude.Ficheros, ruta)
	}

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
