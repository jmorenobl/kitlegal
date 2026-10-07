package app

import (
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"
	"unicode"

	"github.com/jmorenobl/kitlegal/data"
	"github.com/jmorenobl/kitlegal/internal/core"
	"github.com/jmorenobl/kitlegal/internal/core/territorio"
	"github.com/jmorenobl/kitlegal/internal/graph"
)

// Los cinco motivos por los que el registro rechaza un applet, uno por cada
// regla de validación de contracts/registro-y-describe.md §1. Se devuelven
// envueltos con el nombre del applet y del verbo que lo causa, porque quien lee
// el fallo necesita saber cuál de todos los registrados está mal escrito.
//
// Ninguno lleva los sentinelas del kernel, y no es un olvido: un registro mal
// construido es un defecto de compilación que se detecta al arrancar y que
// nunca se convierte en un código de salida de usuario (FR-008, research.md
// D17).
var (
	// ErrNombreInvalido es el nombre vacío, con espacios o con prefijo «-»,
	// que colisionaría con una bandera.
	ErrNombreInvalido = errors.New("app: nombre de applet inválido")
	// ErrNombreDuplicado es el nombre que ya tiene otro applet registrado.
	ErrNombreDuplicado = errors.New("app: applet ya registrado")
	// ErrNombreReservado es el nombre que el binario atiende por su cuenta.
	ErrNombreReservado = errors.New("app: nombre reservado del binario")
	// ErrVerbosInvalidos es el applet sin verbos o con dos verbos que se llaman
	// igual.
	ErrVerbosInvalidos = errors.New("app: verbos del applet inválidos")
	// ErrVerbosPorOmision es el applet que marca más de un verbo por omisión.
	ErrVerbosPorOmision = errors.New("app: más de un verbo por omisión")
)

// verbosReservados son los verbos que atiende el binario y que, por tanto,
// ningún applet puede tomar como nombre: el despacho los reconoce antes de
// mirar el registro, de modo que un applet llamado así sería inalcanzable
// (FR-006, research.md D17).
var verbosReservados = []string{"version"}

// Registro es la colección de applets de un binario y la **única** fuente de la
// que se derivan el despacho, el verbo por omisión, la ayuda y la
// autodescripción: no existe ninguna lista paralela mantenida a mano (FR-001).
//
// Su valor cero es un registro vacío listo para usar. Se construye una sola vez
// al arrancar, en la raíz de composición y antes de que exista ninguna
// invocación, así que no se protege para usarse desde varias gorrutinas: quien
// lo consume después solo lee.
type Registro struct {
	applets map[string]Applet

	// avisador es el aviso de versión del binario, o nulo si no tiene
	// ninguno.
	avisador Avisador

	// almacen es el grafo del mundo al que el kernel entrega lo que observan
	// las invocaciones, o nulo si el registro no entrega nada.
	almacen core.GraphStore

	// entrada es la entrada estándar de las órdenes, o nula si el registro no
	// tiene ninguna.
	entrada io.Reader
}

// Registrar añade un applet al registro después de comprobar las cinco reglas
// de validación del contrato. Devuelve error —que la raíz de composición
// convierte en un fallo de arranque— y nunca deja a medias el registro: un
// applet rechazado no entra (FR-008).
func (r *Registro) Registrar(applet Applet) error {
	nombre := applet.Nombre()

	if err := r.validarNombre(nombre); err != nil {
		return err
	}

	if err := validarVerbos(nombre, applet.Verbos()); err != nil {
		return err
	}

	if r.applets == nil {
		r.applets = make(map[string]Applet)
	}

	r.applets[nombre] = applet

	return nil
}

// Buscar devuelve el applet registrado con ese nombre exacto y si existe. Es lo
// que consultan el despacho —con el nombre de invocación y con el primer
// argumento— y la autodescripción.
func (r *Registro) Buscar(nombre string) (Applet, bool) {
	applet, existe := r.applets[nombre]

	return applet, existe
}

// Nombres son los nombres registrados, ordenados. El orden importa porque de
// aquí salen la ayuda y la lista de applets disponibles que acompaña a un
// nombre desconocido: dos invocaciones iguales tienen que producir la misma
// salida, y el recorrido de un mapa en Go no lo garantiza (FR-006, FR-026).
func (r *Registro) Nombres() []string {
	return slices.Sorted(maps.Keys(r.applets))
}

// Avisar registra el aviso de versión del binario, que el kernel busca en las
// invocaciones que resuelven un applet distinto de skills con un verbo, una
// vez analizado, y escribe en la salida de error (contracts/aviso.md §1;
// research.md D5 de H19). Lo registra la raíz de composición al construir el
// registro; uno nuevo sustituye al anterior y uno nulo deja el registro sin
// aviso, como el valor cero.
func (r *Registro) Avisar(avisador Avisador) {
	r.avisador = avisador
}

// EntregarAlGrafo registra el almacén del grafo del mundo al que el kernel
// entrega, después de presentar, lo que observa cada invocación que termina
// bien sin --no-graph (contracts/resultado-y-entrega.md §5; research.md D6).
// Uno nuevo sustituye al anterior y uno nulo deja el registro sin entrega, como
// el valor cero: el de los tests y el de la preparación de las evals, de modo
// que ninguno escriba en el world.db de la cuenta de quien los ejecuta.
func (r *Registro) EntregarAlGrafo(almacen core.GraphStore) {
	r.almacen = almacen
}

// LeerDe registra la entrada estándar de las órdenes: la que el kernel da, en
// una orden, al verbo que la lee, y a nadie más (H23 FR-020; research.md D12 de
// H23). La registra la raíz de composición al construir el registro, que es
// quien nombra la entrada del proceso; una nueva sustituye a la anterior y una
// nula deja el registro sin entrada, como el valor cero. Registrarla no lee
// nada: de ella lee solo ese verbo, cuando se ejecuta.
//
// No es la entrada de las llamadas de herramienta, que no tienen ninguna: en el
// servidor MCP, la entrada del proceso es el protocolo (H23 FR-026).
func (r *Registro) LeerDe(entrada io.Reader) {
	r.entrada = entrada
}

// sinAvisador es una copia del registro que no da el aviso de versión: los
// mismos applets, el mismo almacén y la misma entrada, sin avisador. Con ella
// resuelve el servidor MCP cada llamada a una herramienta, que es una
// invocación del kernel: el aviso lo da el kernel una vez, al arrancar `mcp
// serve`, y no una por llamada (H21 FR-023; research.md D3 de H21). La entrada
// tampoco llega a ninguna llamada, y no porque la copia la pierda: al verbo se
// le da la del despacho, y una llamada construye el suyo sin ella. El registro
// del que sale no cambia.
func (r *Registro) sinAvisador() *Registro {
	copia := *r
	copia.avisador = nil

	return &copia
}

// aviso es la línea del aviso registrado y si hay que darla. Sin avisador no
// hay ninguna.
func (r *Registro) aviso() (string, bool) {
	if r.avisador == nil {
		return "", false
	}

	return r.avisador()
}

// validarNombre comprueba las tres primeras reglas del contrato: el nombre es
// utilizable, no lo tiene otro applet y no es de los que atiende el binario.
func (r *Registro) validarNombre(nombre string) error {
	switch {
	case nombre == "":
		return fmt.Errorf("%w: un applet no puede registrarse sin nombre", ErrNombreInvalido)
	case strings.ContainsFunc(nombre, unicode.IsSpace):
		return fmt.Errorf("%w: %q lleva espacios y nunca podría escribirse como un solo"+
			" argumento", ErrNombreInvalido, nombre)
	case strings.HasPrefix(nombre, "-"):
		return fmt.Errorf("%w: %q empieza por «-» y colisionaría con una bandera",
			ErrNombreInvalido, nombre)
	case slices.Contains(verbosReservados, nombre):
		return fmt.Errorf("%w: %q lo atiende el propio binario", ErrNombreReservado, nombre)
	}

	if _, duplicado := r.applets[nombre]; duplicado {
		return fmt.Errorf("%w: %q ya está en el registro", ErrNombreDuplicado, nombre)
	}

	return nil
}

// validarVerbos comprueba las dos últimas reglas del contrato: el applet ofrece
// al menos un verbo, con nombres únicos, y no marca dos por omisión —que
// dejaría sin decidir qué ejecuta «kitlegal echo hola»—.
func validarVerbos(applet string, verbos []Verbo) error {
	if len(verbos) == 0 {
		return fmt.Errorf("%w: el applet %q no declara ningún verbo",
			ErrVerbosInvalidos, applet)
	}

	vistos := make(map[string]struct{}, len(verbos))
	porOmision := 0

	for _, verbo := range verbos {
		if _, repetido := vistos[verbo.Nombre]; repetido {
			return fmt.Errorf("%w: el applet %q declara dos veces el verbo %q",
				ErrVerbosInvalidos, applet, verbo.Nombre)
		}

		vistos[verbo.Nombre] = struct{}{}

		if verbo.PorOmision {
			porOmision++
		}
	}

	if porOmision > 1 {
		return fmt.Errorf("%w: el applet %q marca %d verbos por omisión y solo puede marcar uno",
			ErrVerbosPorOmision, applet, porOmision)
	}

	return nil
}

// RegistroDeProduccion es el registro del binario que se publica: el applet boe
// con las dependencias de la red (DependenciasDeRed), el applet graph con las
// del sistema (DependenciasDelGrafoDelSistema), el applet mcp con las del
// sistema (DependenciasDeMCPDelSistema), el applet skills con las del sistema
// (DependenciasDeSkillsDelSistema) y el applet territorio con los ficheros que
// viajan en el binario (FuentesEmbebidas); y, como almacén al que el kernel
// entrega lo que observa cada invocación, el grafo del mundo en world.db, junto
// a la caché y con su misma regla de ubicación (graph.Nuevo sin opciones;
// FR-001, FR-030). Los applets de ejemplo no se registran nunca aquí, sino en
// el binario que compila el test e2e, que usa exactamente este mismo mecanismo
// (FR-001, FR-009, contracts/registro-y-describe.md §3 de H1).
//
// Esta función es la raíz de composición del registro distribuido, y devuelve
// el error del registro en lugar de ocultarlo: un registro inválido es un
// defecto de quien escribió el applet, que TestRegistroDeProduccion impide
// publicar, y si llegara al binario, Arrancar lo convierte en el fallo
// inesperado antes de atender ninguna invocación, nunca en un código de salida
// de usuario ni en un pánico (FR-008; research.md D16 de H4). Construirlo no pide
// nada ni abre nada: tampoco world.db, cuya ruta se resuelve al leer o al
// entregar, ni la entrada estándar, de la que mcp solo lee cuando sirve.
//
// Recibe la versión del binario que le pasa Arrancar —la cadena vacía quien no
// tiene ninguna, que no tiene forma SemVer (FR-073)—, y la firma es la de
// construir en Arrancar (research.md D4 de H19). La lleva a skills, que la
// declara en el manifiesto de cada instalación y compara con ella en doctor,
// junto con lo empotrado en el binario y el creador de enlaces del sistema de
// internal/disco (FR-024, FR-031, FR-077); con esas mismas dependencias compone
// el aviso de versión, que registra (AvisoDeVersion; research.md D5 de H19); y
// la lleva a mcp, que la da como la del servidor a quien se conecta (FR-001 de
// H21; contracts/servidor-mcp.md §1 de H21).
func RegistroDeProduccion(version string) (*Registro, error) {
	fuentes, err := FuentesEmbebidas()
	if err != nil {
		return nil, err
	}

	var registro Registro

	skills := DependenciasDeSkillsDelSistema(version)

	applets := []Applet{
		AppletBoe(DependenciasDeRed()),
		AppletGrafo(DependenciasDelGrafoDelSistema()),
		AppletMCP(DependenciasDeMCPDelSistema(version)),
		AppletSkills(skills),
		AppletTerritorio(fuentes),
	}

	for _, applet := range applets {
		if err := registro.Registrar(applet); err != nil {
			return nil, err
		}
	}

	registro.Avisar(AvisoDeVersion(skills))
	registro.EntregarAlGrafo(graph.Nuevo())

	return &registro, nil
}

// FuentesEmbebidas son los cuatro ficheros de data/territorio/ que viajan en el
// binario, tal como están escritos, en la forma en que los recibe
// AppletTerritorio. Son las que componen la raíz de producción y el binario de
// e2e: no dependen del entorno ni del directorio de trabajo (FR-056, contrato del
// applet territorio §7). Leerlas no analiza nada —eso lo hace el applet la
// primera vez que se ejecuta su verbo—, y un subárbol de comunidades que no se
// puede leer es un defecto de composición, que el error nombra.
func FuentesEmbebidas() (territorio.Fuentes, error) {
	comunidades, err := data.Comunidades()
	if err != nil {
		return territorio.Fuentes{}, fmt.Errorf("app: los ficheros de territorio del binario no se pueden leer: %w", err)
	}

	return territorio.Fuentes{
		Municipios:  data.Municipios,
		DIR3:        data.DIR3,
		Estado:      data.Estado,
		Comunidades: comunidades,
	}, nil
}
