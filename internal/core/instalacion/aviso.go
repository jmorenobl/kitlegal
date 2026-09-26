package instalacion

import "path"

// Aviso es la línea que avisa, sin red, de que las skills instaladas son de
// otra versión que este binario, y dice si hay que darla (data-model §7;
// contracts/aviso.md §2-§4; FR-070 a FR-073, FR-077). La piden las
// invocaciones de otro applet (contracts/aviso.md §1), que la escriben en la
// salida de error; home es el valor de HOME con /, vacío si no está definido.
//
// Primero la forma: si versionDelBinario no tiene la de SemVer 2.0.0, es un
// binario de desarrollo y no hay aviso, sin examinar el disco (FR-073).
//
// Después el manifiesto, sin seguir ningún enlace (FR-028). El local,
// ./.agents/skills/kitlegal.json, si ./.agents y ./.agents/skills son, cada
// uno, un directorio real; si falta cualquiera de los tres, el global, con las
// mismas reglas bajo HOME, salvo que HOME esté vacío, sea relativo o sea la
// raíz del sistema. Lo que existe y no se puede usar —una de esas rutas que no
// es un directorio real, un manifiesto ilegible, también un enlace o lo que no
// es un fichero regular, o un fallo del Disco— no tiene ningún efecto, y si es
// del local el global no se mira: el aviso nunca es un error (FR-072).
//
// Hay aviso si la versión del manifiesto, o la de alguna skill que declara y
// el binario empotra, es distinta de la del binario según MismaVersion; la de
// una skill no empotrada no se compara (FR-036, FR-071). La línea nombra la
// versión que difiere —la del manifiesto si difiere; si no, la de la primera
// de esas skills por nombre—, la del binario y la orden que lo arregla, con -g
// si el manifiesto es el global. Las dos versiones van tal como están, y como
// las del manifiesto no llevan caracteres de control, es siempre una sola
// línea.
func Aviso(disco Disco, home, versionDelBinario string, empotradas []SkillEmpotrada) (string, bool) {
	if !FormaSemVer(versionDelBinario) {
		return "", false
	}

	manifiesto, ambito, hay := manifiestoDelAviso(disco, home)
	if !hay {
		return "", false
	}

	instalada, difiere := versionQueDifiere(manifiesto, versionDelBinario, empotradas)
	if !difiere {
		return "", false
	}

	orden := "kitlegal skills install"
	if banderas := ambito.Banderas(); banderas != "" {
		orden += " " + banderas
	}

	return "aviso: las skills instaladas son de kitlegal " + instalada +
		" y este binario es kitlegal " + versionDelBinario + "; ejecuta: " + orden, true
}

// busquedaDelAviso es lo que el aviso encuentra en un ámbito.
type busquedaDelAviso int

// Las tres salidas de la búsqueda en un ámbito (contracts/aviso.md §2).
const (
	// nadaQueLeer es que falta .agents, .agents/skills o kitlegal.json: desde
	// el local, se sigue con el global.
	nadaQueLeer busquedaDelAviso = iota
	// manifiestoEncontrado es un kitlegal.json legible, el que se compara.
	manifiestoEncontrado
	// sinEfecto es lo que existe y no se puede usar: fin de la búsqueda.
	sinEfecto
)

// manifiestoDelAviso es el manifiesto que compara el aviso y el ámbito en el
// que está, con la regla de búsqueda de FR-070: el local y, solo si en él no
// hay nada que leer, el global. Si no hay ninguno que comparar, devuelve
// falso.
func manifiestoDelAviso(disco Disco, home string) (Manifiesto, Ambito, bool) {
	local := NuevoAmbitoLocal()

	manifiesto, busqueda := buscarElManifiesto(disco, local)

	switch busqueda {
	case manifiestoEncontrado:
		return manifiesto, local, true
	case sinEfecto:
		return Manifiesto{}, Ambito{}, false
	case nadaQueLeer:
		// La regla de búsqueda sigue con el global.
	}

	global, hay := ambitoGlobalDelAviso(home)
	if !hay {
		return Manifiesto{}, Ambito{}, false
	}

	manifiesto, busqueda = buscarElManifiesto(disco, global)
	if busqueda != manifiestoEncontrado {
		return Manifiesto{}, Ambito{}, false
	}

	return manifiesto, global, true
}

// ambitoGlobalDelAviso es el ámbito global en el que el aviso puede buscar,
// el de HOME, si HOME es una ruta absoluta que no es la raíz del sistema. Sin
// HOME, que es el error de NuevoAmbitoGlobal, no se busca en ninguna parte: ni
// en /.agents, que es lo que daría HOME vacío pegado a una ruta, ni en una
// ruta relativa al directorio de trabajo (contracts/aviso.md §2, paso 4); y
// lo mismo con un HOME relativo o que es la raíz del sistema, que llevarían a
// esos mismos sitios.
func ambitoGlobalDelAviso(home string) (Ambito, bool) {
	ambito, err := NuevoAmbitoGlobal(home)
	if err != nil || !path.IsAbs(ambito.Raiz()) || ambito.Raiz() == "/" {
		return Ambito{}, false
	}

	return ambito, true
}

// buscarElManifiesto aplica los pasos 1 a 3 de contracts/aviso.md §2 en el
// ámbito: cada una de sus guardas, .agents y .agents/skills, examinada sin
// seguirla, y después kitlegal.json, que se lee solo si es un fichero regular.
// Un fallo del Disco no permite decidir y, como lo que existe y no se puede
// usar, es sinEfecto: el aviso no lo nombra ni lo devuelve (data-model §7).
func buscarElManifiesto(disco Disco, ambito Ambito) (Manifiesto, busquedaDelAviso) {
	for _, guarda := range ambito.Guardas() {
		entrada, err := disco.Examinar(guarda)
		if err != nil {
			return Manifiesto{}, sinEfecto
		}

		switch entrada.Tipo {
		case EntradaDirectorio:
		case EntradaAusente:
			return Manifiesto{}, nadaQueLeer
		case EntradaFichero, EntradaEnlace, EntradaOtra:
			return Manifiesto{}, sinEfecto
		}
	}

	manifiesto, err := leerManifiestoDe(disco, ambito.RutaDelManifiesto())

	switch {
	case err != nil:
		return Manifiesto{}, sinEfecto
	case manifiesto.Version == "":
		// Un manifiesto leído lleva siempre versión, que LeerManifiesto no
		// admite vacía: sin ella, no había kitlegal.json.
		return Manifiesto{}, nadaQueLeer
	}

	return manifiesto, manifiestoEncontrado
}

// versionQueDifiere es la versión instalada que el aviso nombra, si alguna
// difiere de la del binario según FR-077: la del manifiesto si difiere; si no,
// la de la primera skill por nombre de las que declara y el binario empotra.
func versionQueDifiere(manifiesto Manifiesto, version string, empotradas []SkillEmpotrada) (string, bool) {
	if !MismaVersion(manifiesto.Version, version) {
		return manifiesto.Version, true
	}

	for _, nombre := range declaradasYEmpotradas(manifiesto, empotradas) {
		if instalada := manifiesto.Skills[nombre].Version; !MismaVersion(instalada, version) {
			return instalada, true
		}
	}

	return "", false
}
