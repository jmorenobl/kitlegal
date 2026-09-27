package instalacion

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"maps"
	"path"
	"slices"
	"strings"
	"unicode"
)

// Manifiesto es kitlegal.json, el único de cada ámbito, en la raíz de su
// directorio neutro (FR-030, FR-031; contracts/manifiesto.md): la versión del
// binario que lo escribió por última vez y cada skill declarada, con sus
// ficheros y sus entradas de host. No lleva rutas absolutas, fechas, usuarios ni
// máquinas, así que se versiona con el proyecto (FR-032).
type Manifiesto struct {
	// Version es la del binario que lo escribió por última vez.
	Version string

	// Skills son las declaradas, por nombre, también las que el binario ya no
	// empotra (FR-036).
	Skills map[string]SkillDeclarada
}

// SkillDeclarada es la entrada de una skill en el manifiesto.
type SkillDeclarada struct {
	// Version es la del binario de su instalación.
	Version string

	// Ficheros son los instalados: ruta relativa al directorio neutro, que
	// empieza por el nombre de la skill y una barra, → huella de sus bytes.
	Ficheros map[string]string

	// Hosts son sus entradas de host, por nombre del host; sin ninguna, nil o
	// vacío, que el manifiesto no distingue.
	Hosts map[string]EntradaDeHost
}

// EntradaDeHost es la entrada de una skill en el directorio de skills de un
// host, <raíz del ámbito>/<directorio de skills del host>/<skill> (FR-021,
// FR-024; ADR 0025).
type EntradaDeHost struct {
	// Ruta es la de la entrada relativa a la raíz del ámbito, exactamente
	// .claude/skills/<skill> en claude y .gemini/config/skills/<skill> en
	// antigravity.
	Ruta string

	// Modo dice si la entrada es un enlace o una copia.
	Modo Modo

	// Ficheros son, solo en copia, los copiados: ruta relativa a la raíz del
	// ámbito, que empieza por la de la entrada y una barra, → huella.
	Ficheros map[string]string
}

// Modo es cómo está una entrada de host: enlazada al directorio neutro o, si
// el creador de enlaces falló, copiada (FR-024, FR-025).
type Modo string

// Los dos modos de una entrada de host.
const (
	// ModoEnlace es el enlace simbólico relativo de FR-021.
	ModoEnlace Modo = "enlace"
	// ModoCopia es el recurso de copia de FR-024.
	ModoCopia Modo = "copia"
)

// LeerManifiesto lee el contenido de un kitlegal.json y lo devuelve si respeta
// la forma de contracts/manifiesto.md §1. Si no, devuelve un
// *ManifiestoIlegible y un Manifiesto vacío (§3, reglas 3 y 4; FR-035):
//
//  1. el documento tiene que ser un único valor JSON con la forma del
//     manifiesto, sin nada detrás, en UTF-8 válido, sin miembros que no sean
//     de la forma y sin nombres repetidos en ningún objeto;
//  2. los objetos del contrato no pueden ser null, skills y los ficheros de
//     una skill no pueden faltar, hosts se omite si no tiene entradas y los
//     ficheros de una entrada de host van solo, y siempre, en copia;
//  3. y cada valor cumple su regla: versiones no vacías y sin caracteres de
//     control; nombres de skill con su forma; rutas relativas, limpias y con
//     /, bajo el directorio de su skill o de su copia; huellas sha256; y las
//     claves de hosts, cada una la de un host conocido, con la ruta de FR-021
//     de ese host y un modo conocido (ADR 0025).
//
// Que exista, sea un fichero regular y se pueda leer lo comprueba quien lo
// lee del disco, que solo pasa aquí sus bytes.
func LeerManifiesto(contenido []byte) (Manifiesto, error) {
	var documento documentoDelManifiesto

	err := json.Unmarshal(contenido, &documento,
		json.RejectUnknownMembers(true), jsontext.AllowDuplicateNames(false), jsontext.AllowInvalidUTF8(false))
	if err != nil {
		return Manifiesto{}, &ManifiestoIlegible{
			causa: fmt.Errorf("no es un documento JSON con la forma del manifiesto: %w", err),
		}
	}

	if err := comprobarDocumento(documento); err != nil {
		return Manifiesto{}, &ManifiestoIlegible{causa: err}
	}

	return documento.manifiesto(), nil
}

// Bytes es el manifiesto en su forma canónica (contracts/manifiesto.md §2):
// encoding/json/v2 con las claves de todo objeto en orden, sangrado de dos
// espacios, sin escapar HTML y con un salto de línea final. Los mismos datos
// dan siempre los mismos bytes, en cualquier ruta y en cualquier máquina
// (FR-032).
//
// Lo que LeerManifiesto no podría volver a leer no se escribe: un manifiesto
// que no respeta la forma —una ruta absoluta, una huella que no es sha256, un
// enlace con ficheros…— da un error y ningún byte.
func (m Manifiesto) Bytes() ([]byte, error) {
	documento := documentoDe(m)

	if err := comprobarDocumento(documento); err != nil {
		return nil, fmt.Errorf("el manifiesto no respeta su forma y no se escribe: %w", err)
	}

	contenido, err := json.Marshal(documento,
		json.Deterministic(true), jsontext.WithIndent("  "), jsontext.EscapeForHTML(false))
	if err != nil {
		return nil, fmt.Errorf("serializar el manifiesto: %w", err)
	}

	return append(contenido, '\n'), nil
}

// documentoDelManifiesto es la forma JSON del manifiesto. Los campos van en
// orden alfabético de su clave porque encoding/json/v2 escribe los de una
// estructura en el orden en que se declaran, y así cada objeto sale con sus
// claves en orden, como los mapas con json.Deterministic.
type documentoDelManifiesto struct {
	Skills  objeto[documentoDeSkill] `json:"skills"`
	Version string                   `json:"version"`
}

// documentoDeSkill es la forma JSON de la entrada de una skill.
type documentoDeSkill struct {
	Ficheros objeto[string]          `json:"ficheros"`
	Hosts    objeto[documentoDeHost] `json:"hosts,omitzero"`
	Version  string                  `json:"version"`
}

// documentoDeHost es la forma JSON de una entrada de host.
type documentoDeHost struct {
	Ficheros objeto[string] `json:"ficheros,omitzero"`
	Modo     string         `json:"modo"`
	Ruta     string         `json:"ruta"`
}

// objeto es un objeto JSON del contrato leído como mapa: nil si la clave no
// está en el documento, y un mapa, aunque vacío, si está. Así se distingue la
// clave ausente de la presente sin entradas. null no es un objeto, y lo
// rechaza.
type objeto[V any] map[string]V

// errNoEsUnObjeto es el valor que no es un objeto JSON donde el contrato pide
// uno.
var errNoEsUnObjeto = errors.New("no es un objeto JSON")

// UnmarshalJSONFrom lee un objeto JSON con las opciones de la lectura en
// curso, que el decodificador arrastra: miembros desconocidos y nombres
// repetidos siguen siendo errores dentro de él. encoding/json/v2 llama a este
// método también con null, que no es un objeto.
func (o *objeto[V]) UnmarshalJSONFrom(decodificador *jsontext.Decoder) error {
	if decodificador.PeekKind() != '{' {
		return errNoEsUnObjeto
	}

	leido := map[string]V{}
	if err := json.UnmarshalDecode(decodificador, &leido); err != nil {
		return err
	}

	*o = leido

	return nil
}

// manifiesto es el Manifiesto de un documento ya comprobado, con los mismos
// mapas: nil los ficheros de un enlace, que no los tiene, y un mapa todos los
// demás objetos.
func (d documentoDelManifiesto) manifiesto() Manifiesto {
	skills := make(map[string]SkillDeclarada, len(d.Skills))

	for nombre, entrada := range d.Skills {
		skill := SkillDeclarada{Version: entrada.Version, Ficheros: entrada.Ficheros}

		for nombreDelHost, host := range entrada.Hosts {
			if skill.Hosts == nil {
				skill.Hosts = map[string]EntradaDeHost{}
			}

			skill.Hosts[nombreDelHost] = EntradaDeHost{Ruta: host.Ruta, Modo: Modo(host.Modo), Ficheros: host.Ficheros}
		}

		skills[nombre] = skill
	}

	return Manifiesto{Version: d.Version, Skills: skills}
}

// documentoDe es la forma JSON de m tal como se escribe: skills y los ficheros
// de cada skill van siempre, aunque estén vacíos; hosts, solo con alguna
// entrada de host; y los ficheros de cada entrada van en copia aunque estén
// vacíos y faltan en enlace si no tiene ninguno, de modo que un enlace con
// ficheros llega a la comprobación y no se escribe.
func documentoDe(m Manifiesto) documentoDelManifiesto {
	skills := make(objeto[documentoDeSkill], len(m.Skills))

	for nombre, skill := range m.Skills {
		entrada := documentoDeSkill{Version: skill.Version, Ficheros: presente(skill.Ficheros)}

		for nombreDelHost, host := range skill.Hosts {
			documento := documentoDeHost{Modo: string(host.Modo), Ruta: host.Ruta}
			if host.Modo == ModoCopia || len(host.Ficheros) > 0 {
				documento.Ficheros = presente(host.Ficheros)
			}

			if entrada.Hosts == nil {
				entrada.Hosts = objeto[documentoDeHost]{}
			}

			entrada.Hosts[nombreDelHost] = documento
		}

		skills[nombre] = entrada
	}

	return documentoDelManifiesto{Version: m.Version, Skills: skills}
}

// presente es el objeto de ficheros que se escribe aunque no tenga entradas.
func presente(ficheros map[string]string) objeto[string] {
	if ficheros == nil {
		return objeto[string]{}
	}

	return ficheros
}

// comprobarDocumento aplica las reglas de la forma que la lectura de
// encoding/json/v2 no aplica, en el orden del documento y con las claves de
// cada objeto en orden, de modo que el mismo documento dé siempre el mismo
// motivo. Es la misma comprobación al leer y al escribir.
func comprobarDocumento(documento documentoDelManifiesto) error {
	if err := comprobarVersion(documento.Version); err != nil {
		return fmt.Errorf("version: %w", err)
	}

	if documento.Skills == nil {
		return errors.New("falta skills")
	}

	for _, nombre := range slices.Sorted(maps.Keys(documento.Skills)) {
		if err := comprobarSkill(nombre, documento.Skills[nombre]); err != nil {
			return fmt.Errorf("skill %q: %w", nombre, err)
		}
	}

	return nil
}

// comprobarVersion aplica la regla de toda versión del manifiesto: no vacía y
// sin caracteres de control, que es lo que garantiza que el aviso sea una
// sola línea (research.md D10). No exige la forma de SemVer.
func comprobarVersion(version string) error {
	if version == "" {
		return errors.New("está vacía")
	}

	if strings.ContainsFunc(version, unicode.IsControl) {
		return fmt.Errorf("%q lleva un carácter de control", version)
	}

	return nil
}

// comprobarSkill aplica las reglas de la entrada de la skill nombre: su
// nombre, su versión, sus ficheros y, si tiene, sus entradas de host.
func comprobarSkill(nombre string, entrada documentoDeSkill) error {
	if !esNombreDeSkill(nombre) {
		return fmt.Errorf("el nombre no tiene la forma de un nombre de skill: a-z, 0-9 y -, sin guion al "+
			"principio, al final ni dos seguidos, de 1 a %d caracteres", maximoDelNombreDeSkill)
	}

	if err := comprobarVersion(entrada.Version); err != nil {
		return fmt.Errorf("version: %w", err)
	}

	if entrada.Ficheros == nil {
		return errors.New("falta ficheros")
	}

	if err := comprobarFicheros(entrada.Ficheros, nombre+"/"); err != nil {
		return fmt.Errorf("ficheros: %w", err)
	}

	if entrada.Hosts == nil {
		return nil
	}

	if len(entrada.Hosts) == 0 {
		return errors.New("hosts no tiene ninguna entrada: sin entradas se omite")
	}

	for _, nombreDelHost := range slices.Sorted(maps.Keys(entrada.Hosts)) {
		registrado, conocido := definicionDe(nombreDelHost)
		if !conocido {
			return fmt.Errorf("hosts: el host %q no se admite: los admitidos son %s", nombreDelHost,
				enumeracion(nombresDeHosts()))
		}

		if err := comprobarHost(nombre, registrado, entrada.Hosts[nombreDelHost]); err != nil {
			return fmt.Errorf("hosts: %s: %w", nombreDelHost, err)
		}
	}

	return nil
}

// comprobarHost aplica las reglas de la entrada de la skill nombre en el host
// registrado: su ruta exacta, la del directorio de skills del host y el
// nombre, su modo y los ficheros, que van solo, y siempre, en copia.
func comprobarHost(nombre string, registrado definicionDeHost, host documentoDeHost) error {
	if ruta := path.Join(registrado.skills, nombre); host.Ruta != ruta {
		return fmt.Errorf("la ruta es %q y tiene que ser %q", host.Ruta, ruta)
	}

	switch Modo(host.Modo) {
	case ModoEnlace:
		if host.Ficheros != nil {
			return fmt.Errorf("ficheros solo va en modo %s", ModoCopia)
		}
	case ModoCopia:
		if host.Ficheros == nil {
			return fmt.Errorf("falta ficheros, obligatorio en modo %s", ModoCopia)
		}

		if err := comprobarFicheros(host.Ficheros, host.Ruta+"/"); err != nil {
			return fmt.Errorf("ficheros: %w", err)
		}
	default:
		return fmt.Errorf("el modo es %q y tiene que ser %q o %q", host.Modo, ModoEnlace, ModoCopia)
	}

	return nil
}

// comprobarFicheros aplica a cada fichero declarado las reglas de su ruta, que
// además tiene que empezar por prefijo, y de su huella.
func comprobarFicheros(ficheros objeto[string], prefijo string) error {
	for _, ruta := range slices.Sorted(maps.Keys(ficheros)) {
		if err := comprobarRuta(ruta); err != nil {
			return err
		}

		if !strings.HasPrefix(ruta, prefijo) {
			return fmt.Errorf("la ruta %q no empieza por %q", ruta, prefijo)
		}

		if huella := ficheros[ruta]; !esHuella(huella) {
			return fmt.Errorf("la huella de %q no es sha256: seguido de 64 hexadecimales en minúscula: %q",
				ruta, huella)
		}
	}

	return nil
}

// comprobarRuta aplica la regla de toda ruta del manifiesto (FR-032): relativa,
// con / como único separador —una barra invertida también lo es en Windows, y
// ahí nadie la convertiría— y limpia, sin elementos vacíos, . ni .., igual a
// su forma limpia.
func comprobarRuta(ruta string) error {
	switch {
	case ruta == "":
		return errors.New("una ruta está vacía")
	case strings.HasPrefix(ruta, "/"):
		return fmt.Errorf("la ruta %q es absoluta", ruta)
	case strings.Contains(ruta, `\`):
		return fmt.Errorf("la ruta %q lleva una barra invertida: el separador es /", ruta)
	case !limpia(ruta):
		return fmt.Errorf("la ruta %q tiene elementos vacíos, . o .. y no está limpia", ruta)
	}

	return nil
}

// limpia dice si una ruta relativa no vacía es igual a su forma limpia y no
// tiene ningún elemento . ni ..; path.Clean conserva los .. del principio.
func limpia(ruta string) bool {
	if path.Clean(ruta) != ruta {
		return false
	}

	for elemento := range strings.SplitSeq(ruta, "/") {
		if elemento == "." || elemento == ".." {
			return false
		}
	}

	return true
}
