package instalacion

import (
	"path"
	"strings"
)

// Las rutas que un ámbito define por debajo de su raíz (FR-011, FR-012,
// FR-021, FR-030), relativas a ella y con /.
const (
	// directorioDeAgentes es el que contiene el directorio neutro.
	directorioDeAgentes = ".agents"
	// directorioNeutro es donde se instalan las skills en local y en global.
	directorioNeutro = directorioDeAgentes + "/skills"
	// directorioDelHostClaude es el de configuración del host claude.
	directorioDelHostClaude = ".claude"
	// directorioDeSkillsDelHostClaude es donde viven sus entradas de host.
	directorioDeSkillsDelHostClaude = directorioDelHostClaude + "/skills"
	// nombreDelManifiesto es el del único manifiesto del ámbito, en la raíz de
	// su directorio neutro.
	nombreDelManifiesto = "kitlegal.json"
)

// ClaseDeAmbito dice dónde actúa una invocación de skills (data-model §2).
type ClaseDeAmbito int

// Las tres clases de ámbito. La local es el valor cero porque es la de por
// omisión, sin -g ni --dir (FR-011).
const (
	// AmbitoLocal es el directorio de trabajo, sin buscar ninguna raíz de
	// proyecto ni subir de directorio (FR-011).
	AmbitoLocal ClaseDeAmbito = iota
	// AmbitoGlobal es HOME, con -g/--global (FR-012).
	AmbitoGlobal
	// AmbitoDir es solo el directorio neutro que nombra --dir, sin raíz y sin
	// hosts (FR-013).
	AmbitoDir
)

// Ambito es dónde actúa una invocación de skills install, list o doctor: su
// raíz, su directorio neutro con el manifiesto, las rutas de las que cada una
// tiene que ser un directorio real o no existir, si tiene hosts y las
// banderas que lo repiten en las órdenes de doctor (data-model §2;
// contracts/applet-skills.md §3).
//
// Toda ruta que da va con /, limpia (path.Clean), sin barra final y como se
// alcanza desde el directorio de trabajo de la invocación: relativa en el
// ámbito local, colgando de HOME con -g y de la ruta tal como se pasó con
// --dir (research.md D11). Es la ruta que se presenta en la salida y la que
// retira la orden de un hallazgo de doctor; el adaptador la convierte al
// separador del sistema al tocar el disco. Lo que hay por encima de la raíz o
// de la ruta de --dir no se comprueba y se usa tal cual (FR-027).
//
// El valor cero es el ámbito local. Los otros dos los construyen
// NuevoAmbitoGlobal y NuevoAmbitoDir, o ValidarInvocacion a partir de la
// invocación.
type Ambito struct {
	// clase es la del ámbito.
	clase ClaseDeAmbito
	// raiz es, en global, HOME limpio; vacía en los otros dos.
	raiz string
	// dir es, con --dir, la ruta tal como se pasó; vacía en los otros dos.
	dir string
}

// NuevoAmbitoLocal es el ámbito del directorio de trabajo, el de por omisión
// (FR-011): su directorio neutro es .agents/skills y tiene el host claude en
// .claude/skills.
func NuevoAmbitoLocal() Ambito {
	return Ambito{clase: AmbitoLocal}
}

// NuevoAmbitoGlobal es el ámbito de -g, con raíz en home, el valor de HOME
// con / (FR-012). Sin HOME —sin definir o vacío, que para quien lo lee del
// entorno es lo mismo— no hay ámbito global: devuelve el valor cero y el
// error de clase «inesperado» que dice que falta HOME (contracts/applet-skills.md
// §2, fila 5; data-model §9). HOME no se comprueba de ninguna otra forma: es
// lo que hay por encima de la raíz (FR-027).
func NuevoAmbitoGlobal(home string) (Ambito, error) {
	if home == "" {
		return Ambito{}, homeSinDefinir()
	}

	return Ambito{clase: AmbitoGlobal, raiz: path.Clean(home)}, nil
}

// NuevoAmbitoDir es el ámbito de --dir con la ruta tal como se pasó, con /
// (FR-013): esa ruta, limpia, es su directorio neutro, y no tiene raíz ni
// hosts. Una ruta relativa cuelga del directorio de trabajo, y una vacía es
// el propio directorio de trabajo, como su forma limpia, «.».
func NuevoAmbitoDir(ruta string) Ambito {
	return Ambito{clase: AmbitoDir, dir: ruta}
}

// Clase es la del ámbito.
func (a Ambito) Clase() ClaseDeAmbito {
	return a.clase
}

// Raiz es la raíz del ámbito, de la que cuelgan su directorio neutro y el
// directorio de su host (FR-021): vacía en el local, que es el directorio de
// trabajo, y HOME limpio en el global. Con --dir no hay raíz y también es
// vacía: ese ámbito es solo su directorio neutro.
func (a Ambito) Raiz() string {
	return a.raiz
}

// Neutro es el directorio neutro, donde se instala cada skill y vive el
// manifiesto: .agents/skills bajo la raíz en local y en global, y la ruta de
// --dir, limpia.
func (a Ambito) Neutro() string {
	if a.clase == AmbitoDir {
		return path.Clean(a.dir)
	}

	return path.Join(a.raiz, directorioNeutro)
}

// Guardas son las rutas del ámbito de las que cada una tiene que ser un
// directorio real o no existir, comprobado sin seguir enlaces (FR-027):
// .agents y .agents/skills bajo la raíz en local y en global, y la ruta de
// --dir, en ese orden. Es una lista nueva en cada llamada.
func (a Ambito) Guardas() []string {
	if a.clase == AmbitoDir {
		return []string{a.Neutro()}
	}

	return []string{path.Join(a.raiz, directorioDeAgentes), a.Neutro()}
}

// ConHosts dice si el ámbito tiene el host claude: el local y el global sí;
// el de --dir, no (FR-013).
func (a Ambito) ConHosts() bool {
	return a.clase != AmbitoDir
}

// Banderas son las que repite cada orden de doctor para actuar en este mismo
// ámbito (FR-066): ninguna en local, -g en global y, con --dir, la ruta tal
// como se pasó, entre comillas simples.
func (a Ambito) Banderas() string {
	switch a.clase {
	case AmbitoGlobal:
		return "-g"
	case AmbitoDir:
		return "--dir " + entreComillas(a.dir)
	case AmbitoLocal:
		// El de por omisión no lleva ninguna.
	}

	return ""
}

// RutaDelManifiesto es la de kitlegal.json, en la raíz del directorio neutro
// (FR-030).
func (a Ambito) RutaDelManifiesto() string {
	return path.Join(a.Neutro(), nombreDelManifiesto)
}

// RutaDeSkill es la del directorio de la skill nombre en el directorio neutro,
// la que se presenta en la salida (FR-051). nombre tiene que ser el de una
// skill: no se comprueba.
func (a Ambito) RutaDeSkill(nombre string) string {
	return path.Join(a.Neutro(), nombre)
}

// DirectorioDelHost es el de configuración del host claude, .claude bajo la
// raíz, cuya existencia decide si se enlaza sin --host (FR-022). Con --dir,
// que no tiene hosts, es vacía.
func (a Ambito) DirectorioDelHost() string {
	if !a.ConHosts() {
		return ""
	}

	return path.Join(a.raiz, directorioDelHostClaude)
}

// SkillsDelHost es el directorio de las entradas del host claude,
// .claude/skills bajo la raíz (FR-026). Con --dir, que no tiene hosts, es
// vacía.
func (a Ambito) SkillsDelHost() string {
	if !a.ConHosts() {
		return ""
	}

	return path.Join(a.raiz, directorioDeSkillsDelHostClaude)
}

// RutaDeHost es la de la entrada de la skill nombre en el host claude,
// .claude/skills/<nombre> bajo la raíz, la que se presenta en la salida
// (FR-021, FR-051). Con --dir, que no tiene hosts, es vacía. nombre tiene que
// ser el de una skill: no se comprueba.
func (a Ambito) RutaDeHost(nombre string) string {
	if !a.ConHosts() {
		return ""
	}

	return path.Join(a.raiz, directorioDeSkillsDelHostClaude, nombre)
}

// entreComillas es texto como una sola palabra de shell POSIX: entre comillas
// simples, dentro de las cuales nada se expande, y con cada comilla simple
// escrita así, que cierra las comillas, pone una comilla escapada y las
// vuelve a abrir (FR-066):
//
//	'\''
func entreComillas(texto string) string {
	return "'" + strings.ReplaceAll(texto, "'", `'\''`) + "'"
}
