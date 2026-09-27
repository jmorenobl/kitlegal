package instalacion

import (
	"path"
	"strings"
)

// Las rutas que un ámbito define por debajo de su raíz (FR-011, FR-012,
// FR-030), relativas a ella y con /.
const (
	// directorioDeAgentes es el que contiene el directorio neutro.
	directorioDeAgentes = ".agents"
	// directorioNeutro es donde se instalan las skills en local y en global.
	directorioNeutro = directorioDeAgentes + "/skills"
	// nombreDelManifiesto es el del único manifiesto del ámbito, en la raíz de
	// su directorio neutro.
	nombreDelManifiesto = "kitlegal.json"
)

// Los nombres de los hosts conocidos, los valores que admite --host y las
// claves de hosts en el manifiesto (FR-020; ADR 0025).
const (
	hostClaude      = "claude"
	hostAntigravity = "antigravity"
)

// definicionDeHost es un host que no lee el directorio neutro y necesita sus
// propias entradas: su nombre; su directorio de configuración, ajustes, cuya
// existencia decide si se enlaza sin --host (FR-022); su directorio de skills, donde viven
// sus entradas (FR-021); los dos relativos a la raíz del ámbito y con /; y si
// solo lo tiene el ámbito global.
type definicionDeHost struct {
	nombre     string
	ajustes    string
	skills     string
	soloGlobal bool
}

// hostsConocidos son los hosts, en el orden en que se examinan, se enlazan y
// se presentan (ADR 0025). claude lee .claude/skills en local y en global.
// antigravity lee el directorio neutro del proyecto, así que en local no
// necesita entradas, pero en global lee .gemini/config/skills y no
// ~/.agents/skills. Los que leen el directorio neutro en los dos ámbitos, como
// Codex, no son hosts.
var hostsConocidos = []definicionDeHost{
	{nombre: hostClaude, ajustes: ".claude", skills: ".claude/skills"},
	{nombre: hostAntigravity, ajustes: ".gemini/config", skills: ".gemini/config/skills", soloGlobal: true},
}

// esHostConocido dice si nombre es el de un host conocido.
func esHostConocido(nombre string) bool {
	_, conocido := definicionDe(nombre)

	return conocido
}

// definicionDe es la definición del host nombre y si es uno conocido.
func definicionDe(nombre string) (definicionDeHost, bool) {
	for _, host := range hostsConocidos {
		if host.nombre == nombre {
			return host, true
		}
	}

	return definicionDeHost{}, false
}

// HostsConocidos son los nombres de los hosts conocidos, en su orden: los
// valores que admite --host y el vocabulario del host de cada entrada en la
// salida. Es una lista nueva en cada llamada.
func HostsConocidos() []string {
	return nombresDeHosts()
}

// nombresDeHosts son los nombres de los hosts conocidos, en su orden.
func nombresDeHosts() []string {
	nombres := make([]string, 0, len(hostsConocidos))
	for _, host := range hostsConocidos {
		nombres = append(nombres, host.nombre)
	}

	return nombres
}

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
// (FR-011): su directorio neutro es .agents/skills y su único host, claude,
// tiene sus entradas en .claude/skills.
func NuevoAmbitoLocal() Ambito {
	return Ambito{clase: AmbitoLocal}
}

// NuevoAmbitoGlobal es el ámbito de -g, con raíz en home, el valor de HOME
// con / (FR-012). Sin HOME —sin definir o vacío, que para quien lo lee del
// entorno es lo mismo— no hay ámbito global: devuelve el valor cero y el
// error de clase «conflicto» que dice que falta HOME (contracts/applet-skills.md
// §2, fila 5; data-model §9; ADR 0023). HOME no se comprueba de ninguna otra forma: es
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

// Raiz es la raíz del ámbito, de la que cuelgan su directorio neutro y el de
// cada uno de sus hosts (FR-021): vacía en el local, que es el directorio de
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

// ConHosts dice si el ámbito tiene algún host: el local y el global sí; el de
// --dir, no (FR-013).
func (a Ambito) ConHosts() bool {
	return a.clase != AmbitoDir
}

// Hosts son los nombres de los hosts del ámbito, en el orden de los hosts
// conocidos: claude en local; claude y antigravity en global; ninguno con
// --dir (FR-013; ADR 0025). Es una lista nueva en cada llamada.
func (a Ambito) Hosts() []string {
	hosts := []string{}

	for _, host := range hostsConocidos {
		if a.tieneHost(host) {
			hosts = append(hosts, host.nombre)
		}
	}

	return hosts
}

// TieneHost dice si el host nombre es uno del ámbito.
func (a Ambito) TieneHost(nombre string) bool {
	host, conocido := definicionDe(nombre)

	return conocido && a.tieneHost(host)
}

// tieneHost dice si el ámbito tiene host: con --dir, ninguno; en local, los
// que no son solo del global.
func (a Ambito) tieneHost(host definicionDeHost) bool {
	switch a.clase {
	case AmbitoDir:
		return false
	case AmbitoLocal:
		return !host.soloGlobal
	case AmbitoGlobal:
	}

	return true
}

// Banderas son las que repite cada orden de doctor para actuar en este mismo
// ámbito (FR-066): ninguna en local, -g en global y, con --dir, la ruta tal
// como se pasó, entre comillas simples.
//
// La ruta de --dir va en la palabra siguiente, «--dir '<ruta>'», salvo si
// empieza por «-»: el análisis de la invocación lee esa palabra como otra
// bandera, y la orden saldría con 2 sin reinstalar nada, así que entonces va
// en la misma palabra, «--dir='<ruta>'», que es la forma que acepta.
func (a Ambito) Banderas() string {
	switch a.clase {
	case AmbitoGlobal:
		return "-g"
	case AmbitoDir:
		if strings.HasPrefix(a.dir, "-") {
			return "--dir=" + entreComillas(a.dir)
		}

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

// DirectorioDelHost es el de configuración del host, bajo la raíz —.claude,
// o .gemini/config para antigravity—, cuya existencia decide si se enlaza sin
// --host (FR-022). Vacía si el ámbito no tiene ese host.
func (a Ambito) DirectorioDelHost(host string) string {
	registrado, hay := a.definicionDelHost(host)
	if !hay {
		return ""
	}

	return path.Join(a.raiz, registrado.ajustes)
}

// SkillsDelHost es el directorio de las entradas del host, bajo la raíz
// —.claude/skills, o .gemini/config/skills para antigravity— (FR-026). Vacía si
// el ámbito no tiene ese host.
func (a Ambito) SkillsDelHost(host string) string {
	registrado, hay := a.definicionDelHost(host)
	if !hay {
		return ""
	}

	return path.Join(a.raiz, registrado.skills)
}

// RutaDeHost es la de la entrada de la skill nombre en el host, <directorio
// de skills del host>/<nombre>, la que se presenta en la salida (FR-021,
// FR-051). Vacía si el ámbito no tiene ese host. nombre tiene que ser el de una
// skill: no se comprueba.
func (a Ambito) RutaDeHost(host, nombre string) string {
	registrado, hay := a.definicionDelHost(host)
	if !hay {
		return ""
	}

	return path.Join(a.raiz, registrado.skills, nombre)
}

// cadenaDelHost es cada directorio que lleva de la raíz al de skills del
// host, de arriba abajo y el de skills el último: .claude y .claude/skills, o
// .gemini, .gemini/config y .gemini/config/skills. Cada uno tiene que ser un
// directorio real o no existir, porque nada por debajo de la raíz se lee a
// través de un enlace (FR-027, FR-028). Vacía si el ámbito no tiene ese host.
func (a Ambito) cadenaDelHost(host string) []string {
	registrado, hay := a.definicionDelHost(host)
	if !hay {
		return nil
	}

	var cadena []string
	for _, dir := range cadenaDe(registrado.skills) {
		if dir != "." {
			cadena = append(cadena, path.Join(a.raiz, dir))
		}
	}

	return cadena
}

// definicionDelHost es la definición del host, si el ámbito lo tiene.
func (a Ambito) definicionDelHost(host string) (definicionDeHost, bool) {
	registrado, conocido := definicionDe(host)

	return registrado, conocido && a.tieneHost(registrado)
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
