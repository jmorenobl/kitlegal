package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"sync"

	"github.com/jmorenobl/kitlegal"
	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/disco"
)

// La procedencia con la que firma skills. Es un applet calculado: no consulta
// ninguna fuente, sino lo que el binario lleva dentro y el disco del ámbito, así
// que firma en el espacio de nombres reservado (FR-050; ADR 0006). La fecha de
// consulta la pone el reloj del kernel.
const (
	fuenteDeSkills = "kitlegal.skills"
	urlDeSkills    = "kitlegal:applet/skills"
)

// variableHome es la del entorno de la que sale la raíz del ámbito global
// (FR-012). Se lee en cada invocación, y sin definir o vacía es lo mismo.
const variableHome = "HOME"

// Los tres verbos, como se escriben en la invocación y como los nombra el
// mensaje de su fallo (contracts/applet-skills.md §1 y §5).
const (
	verboInstall = "install"
	verboList    = "list"
	verboDoctor  = "doctor"
)

// Los defectos de composición: unas dependencias con las que el applet no puede
// atender ninguna invocación. No declaran clase, así que salen como inesperado,
// sin procedencia, que firma y fecha el kernel.
var (
	// errSinEnlazador son unas dependencias sin creador de enlaces.
	errSinEnlazador = errors.New("skills: el applet se compuso sin el Enlazador de sus dependencias")
	// errSinLoEmpotrado son unas dependencias sin lo empotrado.
	errSinLoEmpotrado = errors.New("las dependencias no lo traen")
)

// DependenciasDeSkills es lo que el applet skills recibe de la raíz de
// composición: la versión del binario, lo que lleva empotrado y el creador de
// enlaces. El binario distribuido da las de DependenciasDeSkillsDelSistema; el de
// e2e sustituye en ellas el Enlazador por uno que siempre falla, y las pruebas,
// además, lo empotrado (FR-024; research.md D9).
//
// El applet no compone el Enlazador dentro a propósito: es lo único del disco
// cuya respuesta depende del sistema de ficheros, y sustituirlo es como se
// ejerce el recurso de copia. El lector y el escritor del disco sí los compone
// él, con internal/disco.
type DependenciasDeSkills struct {
	// Version es la del binario: la que declara el manifiesto de cada
	// instalación (FR-031) y con la que doctor compara (FR-077).
	Version string
	// Skills es lo empotrado, con la forma de kitlegal.Skills(): una carpeta
	// skills/<skill>/ por skill, con su SKILL.md y su references/
	// (contracts/skills-e-invocacion.md §1). El applet lo lee la primera vez
	// que se ejecuta uno de sus verbos.
	Skills fs.FS
	// Enlazador crea las entradas de host y dice si se pueden crear en un
	// directorio (FR-024).
	Enlazador instalacion.Enlazador
}

// DependenciasDeSkillsDelSistema son las del binario distribuido: la versión
// que recibe, lo empotrado en el binario y el creador de enlaces del sistema de
// internal/disco, que sondea en el propio sistema de ficheros del ámbito
// (research.md D9). Componerlas no lee nada.
func DependenciasDeSkillsDelSistema(version string) DependenciasDeSkills {
	return DependenciasDeSkills{Version: version, Skills: kitlegal.Skills(), Enlazador: disco.Enlazador{}}
}

// AppletSkills es el applet que instala en un ámbito las skills que el binario
// lleva dentro y dice qué hay instalado y si sigue como lo dejó, con tres
// verbos, install, list y doctor, y ninguno por omisión: nombrar el verbo es
// obligatorio (contracts/applet-skills.md §1). Ninguno abre una conexión
// (FR-016) ni emite operaciones de grafo (FR-054).
//
// Componerlo no lee nada. Lo empotrado se lee una sola vez por applet, la
// primera vez que se ejecuta un verbo, y lo que resulte —las skills o su
// defecto— sirve a todas las invocaciones siguientes: la ayuda y --describe no
// lo pagan, y la lectura vive en el valor del applet y no en el paquete.
func AppletSkills(dependencias DependenciasDeSkills) Applet {
	return appletSkills{
		dependencias: dependencias,
		empotradas: sync.OnceValues(func() ([]instalacion.SkillEmpotrada, error) {
			if dependencias.Skills == nil {
				return nil, errSinLoEmpotrado
			}

			return skillsEmpotradasDe(dependencias.Skills)
		}),
	}
}

// appletSkills declara lo que declara un applet y nada más: su nombre, su línea
// de ayuda y su catálogo de verbos, que llevan consigo las dependencias y la
// lectura de lo empotrado.
type appletSkills struct {
	dependencias DependenciasDeSkills

	// empotradas lee lo empotrado la primera vez que se le llama y devuelve lo
	// mismo en todas las siguientes.
	empotradas func() ([]instalacion.SkillEmpotrada, error)
}

func (appletSkills) Nombre() string { return "skills" }

func (appletSkills) Descripcion() string {
	return "Instala en el proyecto, o en la cuenta, las skills que lleva el binario, y dice qué hay instalado" +
		" y si sigue como lo dejó."
}

// Verbos son los tres del contrato, en su orden y sin ninguno por omisión, con
// sus banderas y el valor cero del tipo de su data (contracts/applet-skills.md
// §1 y §4).
func (a appletSkills) Verbos() []Verbo {
	return []Verbo{
		{
			Nombre: verboInstall,
			Descripcion: "Instala las skills del binario, todas o las nombradas, en el directorio de skills del" +
				" ámbito y las enlaza en sus hosts.",
			Argumentos: func() Argumentos { return &argumentosDeInstall{applet: a} },
			Salida:     []instalacion.SkillInstalada(nil),
		},
		{
			Nombre:      verboList,
			Descripcion: "Lista las skills que declara el manifiesto del ámbito, con su versión y sus enlaces.",
			Argumentos:  a.deUnAmbito(a.dependencias.listar),
			Salida:      instalacion.Listado{},
		},
		{
			Nombre: verboDoctor,
			Descripcion: "Comprueba lo instalado en el ámbito contra su manifiesto y da la orden que arregla cada" +
				" hallazgo.",
			Argumentos: a.deUnAmbito(a.dependencias.diagnosticar),
			Salida:     instalacion.Diagnostico{},
		},
	}
}

// deUnAmbito es la fábrica de argumentos de list y doctor, que solo aceptan las
// banderas del ámbito y se distinguen por la operación que hacen en él.
func (a appletSkills) deUnAmbito(operacion operacionDeSkills) func() Argumentos {
	return func() Argumentos { return &argumentosDeAmbito{applet: a, operacion: operacion} }
}

// argumentosDeInstall son los de install: las skills que se piden, por su
// posición, y las banderas del ámbito y del host (contracts/applet-skills.md
// §1). Host y Dir son punteros porque pasar la bandera con un valor vacío
// también es pasarla (FR-052). Qué combinaciones se admiten no lo decide la
// gramática sino la validación del dominio, con sus mensajes (research.md D14).
type argumentosDeInstall struct {
	Skills []string `arg:"" optional:"" name:"skill" help:"Skills que se instalan; sin ninguna, todas las del binario."`
	Global bool     `short:"g" help:"Actúa en el ámbito global: $HOME/.agents/skills y los hosts de HOME."`
	Host   *string  `placeholder:"claude" help:"Enlaza en el host aunque su directorio no exista; el único es claude."`
	Dir    *string  `placeholder:"<ruta>" help:"Directorio de skills en lugar de .agents/skills: el ámbito es esa ruta, sin hosts."`

	applet appletSkills
}

// Ejecutar instala las skills pedidas, o con --dry-run describe lo que haría.
func (a *argumentosDeInstall) Ejecutar(
	_ context.Context, ec schema.Contexto, _ *slog.Logger,
) (schema.Resultado, error) {
	invocacion := instalacion.Invocacion{Skills: a.Skills, Global: a.Global, Host: a.Host, Dir: a.Dir}

	return a.applet.ejecutar(invocacion, ec.DryRun, a.applet.dependencias.instalar)
}

// argumentosDeAmbito son los de list y doctor: las banderas del ámbito, y la
// operación del verbo en un campo no exportado que la gramática no ve
// (contracts/applet-skills.md §1).
type argumentosDeAmbito struct {
	Global bool    `short:"g" help:"Actúa en el ámbito global: $HOME/.agents/skills y los hosts de HOME."`
	Dir    *string `placeholder:"<ruta>" help:"Directorio de skills en lugar de .agents/skills: el ámbito es esa ruta, sin hosts."`

	applet    appletSkills
	operacion operacionDeSkills
}

// Ejecutar hace la operación del verbo en el ámbito. Ni list ni doctor tienen
// nada que describir con --dry-run: no cambian nada en el disco (FR-062,
// FR-068).
func (a *argumentosDeAmbito) Ejecutar(
	_ context.Context, ec schema.Contexto, _ *slog.Logger,
) (schema.Resultado, error) {
	return a.applet.ejecutar(instalacion.Invocacion{Global: a.Global, Dir: a.Dir}, ec.DryRun, a.operacion)
}

// Las comprobaciones en tiempo de compilación del contrato del applet.
var (
	_ Applet     = appletSkills{}
	_ Argumentos = (*argumentosDeInstall)(nil)
	_ Argumentos = (*argumentosDeAmbito)(nil)
)

// operacionDeSkills es lo que hace un verbo con la invocación ya validada: su
// data, o con --dry-run su descripción, sin procedencia, o su fallo.
type operacionDeSkills func(
	pedido instalacion.Pedido, empotradas []instalacion.SkillEmpotrada, ensayo bool,
) (schema.Resultado, error)

// ejecutar atiende una invocación de cualquiera de los tres verbos, en el orden
// que fija FR-052:
//
//  1. unas dependencias que no sirven —sin Enlazador, o con lo empotrado que
//     falta o no se puede leer— son un defecto de composición, que no es algo
//     que quien invoca pueda corregir: sale como inesperado, sin procedencia;
//  2. la invocación se valida antes de tocar el disco, con HOME tal como está en
//     el entorno: un error de argumentos sale con 2 y, solo sin ninguno, -g sin
//     HOME sale con 7, un conflicto con el entorno (contracts/applet-skills.md
//     §2; ADR 0023);
//  3. y solo entonces el verbo examina el ámbito.
//
// Todo lo que decide el applet, correcto o fallido, lo firma él (FR-050): el
// fallo va en el sobre de fallo del kernel, sin cambios, con su mensaje en data y
// en la salida de error (FR-052).
func (a appletSkills) ejecutar(
	invocacion instalacion.Invocacion, ensayo bool, operacion operacionDeSkills,
) (schema.Resultado, error) {
	if a.dependencias.Enlazador == nil {
		return schema.Resultado{}, errSinEnlazador
	}

	empotradas, err := a.empotradas()
	if err != nil {
		return schema.Resultado{}, defectoDeLoEmpotrado(err)
	}

	firmado := schema.Resultado{Procedencia: schema.Procedencia{Fuente: fuenteDeSkills, URL: urlDeSkills}}

	pedido, err := instalacion.ValidarInvocacion(invocacion, os.Getenv(variableHome), empotradas)
	if err != nil {
		return firmado, err
	}

	resultado, err := operacion(pedido, empotradas, ensayo)
	resultado.Procedencia = firmado.Procedencia

	return resultado, err
}

// instalar planifica la instalación de las skills del pedido, comprobando cada
// conflicto antes de escribir nada, y la aplica con el disco real y el Enlazador
// de las dependencias (FR-040 a FR-047); su data es cada skill pedida tal como
// queda (FR-051). Con --dry-run no aplica nada: describe el plan, una línea por
// skill, que el kernel presenta en la salida de error, y con algún conflicto
// falla igual que sin la bandera, con código 7: el conflicto se conoce sin
// efectos, así que el ensayo predice el código de la orden real (FR-048;
// research.md D13; ADR 0023). La sonda del Enlazador sí
// se hace, porque es de donde sale el modo que se predice, y deja el disco como
// estaba (research.md D9).
func (d DependenciasDeSkills) instalar(
	pedido instalacion.Pedido, empotradas []instalacion.SkillEmpotrada, ensayo bool,
) (schema.Resultado, error) {
	plan, err := instalacion.Planificar(disco.Lector{}, d.Enlazador, pedido, empotradas, d.Version)
	if err != nil {
		return schema.Resultado{}, delVerbo(verboInstall, err)
	}

	if ensayo {
		return schema.Resultado{Ensayo: ensayoDeInstall(plan.Skills)}, nil
	}

	instaladas, err := instalacion.Aplicar(plan, disco.NuevoEscritor(d.Enlazador))
	if err != nil {
		// Aplicar ya lo nombra: «skills install: <operación> <ruta>: <error del
		// sistema>» (FR-044).
		return schema.Resultado{}, err
	}

	return schema.Resultado{Datos: instaladas}, nil
}

// listar da lo que el manifiesto del ámbito declara, sin examinar nada más ni
// cambiar nada (FR-060 a FR-062).
func (DependenciasDeSkills) listar(
	pedido instalacion.Pedido, empotradas []instalacion.SkillEmpotrada, _ bool,
) (schema.Resultado, error) {
	listado, err := instalacion.Listar(disco.Lector{}, pedido.Ambito, empotradas)
	if err != nil {
		return schema.Resultado{}, delVerbo(verboList, err)
	}

	return schema.Resultado{Datos: listado}, nil
}

// diagnosticar compara el disco del ámbito con su manifiesto y con la versión
// del binario, sin cambiar nada, con el Enlazador de las dependencias para
// decidir si una copia ya podría ser un enlace (FR-065 a FR-069). Su data es el
// diagnóstico, con los hallazgos que haya: encontrar algo es el resultado de una
// verificación que ha funcionado, y sale con código 0 (ADR 0023).
func (d DependenciasDeSkills) diagnosticar(
	pedido instalacion.Pedido, empotradas []instalacion.SkillEmpotrada, _ bool,
) (schema.Resultado, error) {
	diagnostico, err := instalacion.Diagnosticar(disco.Lector{}, d.Enlazador, pedido.Ambito, empotradas, d.Version)
	if err != nil {
		return schema.Resultado{}, delVerbo(verboDoctor, err)
	}

	return schema.Resultado{Datos: diagnostico}, nil
}

// ensayoDeInstall es lo que --dry-run describe de install: una línea por skill
// pedida, con su ruta, el estado que resultaría y cada entrada de host con el
// modo que predice la sonda (contracts/applet-skills.md §7). El prefijo
// «--dry-run: se habría pedido » lo pone el kernel.
func ensayoDeInstall(skills []instalacion.SkillInstalada) []string {
	lineas := make([]string, 0, len(skills))

	for _, skill := range skills {
		linea := "instalar " + skill.Nombre + " en " + skill.Ruta + ": " + string(skill.Estado)
		for _, enlace := range skill.Enlaces {
			linea += "; enlace " + enlace.Ruta + " (" + string(enlace.Modo) + ")"
		}

		lineas = append(lineas, linea)
	}

	return lineas
}

// delVerbo pone «skills <verbo>: » delante de un fallo que el dominio devuelve
// tal cual y que no lo nombra: el del disco al examinar o leer, el de la sonda
// del creador de enlaces que no se pudo retirar y el de una versión del binario
// que el manifiesto no admitiría (contracts/applet-skills.md §5). Lo envuelve,
// porque no declara ninguna clase que pudiera decidir otro código. Los que sí la
// declaran —el de un conflicto de install y el del ámbito ilegible— ya empiezan
// por él y pasan tal cual.
func delVerbo(verbo string, err error) error {
	var conClase schema.ConClase
	if errors.As(err, &conClase) {
		return err
	}

	return fmt.Errorf("skills %s: %w", verbo, err)
}

// defectoDeLoEmpotrado es el fallo de lo empotrado que no se puede leer.
// Incorpora el defecto como texto y **no** con %w a propósito, como el de
// territorio: el código de salida tiene que ser el del fallo inesperado sea cual
// sea el error que lo causó.
func defectoDeLoEmpotrado(err error) error {
	return fmt.Errorf("skills: lo empotrado con que se compuso el applet no se puede leer: %s", err.Error())
}
