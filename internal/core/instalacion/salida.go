package instalacion

// Los tipos de la salida de install, list y doctor, el data de cada verbo, con
// las claves de data-model §10 en español y en el orden del contrato
// (contracts/applet-skills.md §4; FR-051, FR-060, FR-067). Ninguno lleva
// omitempty y una lista vacía es una lista, nunca nil, así que se emite [] y
// no null; la versión de un ámbito sin manifiesto es nil y se emite null.
//
// Las etiquetas jsonschema son la descripción formal de los valores, de la que
// --describe genera el esquema del verbo y de la que sale el publicado
// (research.md D15, D16). Una etiqueta no puede nombrar una constante, así que
// los enumerados repiten los vocabularios de Estado, Modo, el host y
// ClaseDeHallazgo, y TestPlan y TestHallazgos exigen que digan lo mismo.

// Estado es cómo queda una skill pedida tras install (FR-046, FR-051).
type Estado string

// Los tres estados de una skill pedida.
const (
	// EstadoInstalada es el de la skill que el manifiesto no declaraba.
	EstadoInstalada Estado = "instalada"
	// EstadoActualizada es el de la skill declarada cuyos ficheros, entradas
	// de host o entrada del manifiesto, versión incluida, cambian.
	EstadoActualizada Estado = "actualizada"
	// EstadoSinCambios es el de la skill declarada en la que no cambia nada.
	EstadoSinCambios Estado = "sin cambios"
)

// SkillInstalada es una skill pedida tal como queda tras install: su nombre,
// la ruta de su directorio en el directorio neutro, su estado y sus entradas
// de host, como se presentan (FR-051; research.md D11).
type SkillInstalada struct {
	// Nombre es el de la skill.
	Nombre string `json:"nombre" jsonschema:"minLength=1"`
	// Ruta es la de su directorio, como se alcanza desde el directorio de
	// trabajo.
	Ruta string `json:"ruta" jsonschema:"minLength=1"`
	// Estado es el que resulta de la orden.
	Estado Estado `json:"estado" jsonschema:"enum=instalada,enum=actualizada,enum=sin cambios"`
	// Enlaces son sus entradas de host tras la orden; ninguna sin hosts.
	Enlaces []Enlace `json:"enlaces"`
}

// Enlace es una entrada de host de una skill: el host, la ruta de la entrada
// como se alcanza desde el directorio de trabajo y su modo, enlace o, si el
// creador de enlaces falló, copia (FR-024, FR-051).
type Enlace struct {
	// Host es el del directorio de skills de la entrada.
	Host string `json:"host" jsonschema:"enum=claude"`
	// Ruta es la de la entrada.
	Ruta string `json:"ruta" jsonschema:"minLength=1"`
	// Modo es cómo está la entrada.
	Modo Modo `json:"modo" jsonschema:"enum=enlace,enum=copia"`
}

// Listado es lo que da list: el directorio neutro del ámbito, si hay
// manifiesto, su versión y cada skill que declara, en orden de nombre (FR-060,
// FR-061). Sin manifiesto, la versión es nula y la lista está vacía; con él, la
// versión nunca es nula, aunque no declare ninguna skill.
type Listado struct {
	// Directorio es el directorio neutro, como se alcanza desde el directorio
	// de trabajo.
	Directorio string `json:"directorio" jsonschema:"minLength=1"`
	// Manifiesto dice si el ámbito tiene manifiesto.
	Manifiesto bool `json:"manifiesto"`
	// Version es la del manifiesto, o nil sin él.
	Version *string `json:"version" jsonschema:"nullable,minLength=1"`
	// Skills son las declaradas, también las que el binario no empotra.
	Skills []SkillListada `json:"skills"`
}

// SkillListada es una skill que declara el manifiesto, tal como la da list: su
// nombre, la ruta de su directorio, la versión del binario de su instalación,
// si este binario la empotra y sus entradas de host con su modo (FR-036,
// FR-060).
type SkillListada struct {
	// Nombre es el de la skill.
	Nombre string `json:"nombre" jsonschema:"minLength=1"`
	// Ruta es la de su directorio, como se alcanza desde el directorio de
	// trabajo.
	Ruta string `json:"ruta" jsonschema:"minLength=1"`
	// Version es la que declara el manifiesto para ella.
	Version string `json:"version" jsonschema:"minLength=1"`
	// Empotrada dice si este binario la lleva.
	Empotrada bool `json:"empotrada"`
	// Enlaces son sus entradas de host declaradas; ninguna con --dir.
	Enlaces []Enlace `json:"enlaces"`
}

// Diagnostico es lo que da doctor: el directorio neutro del ámbito, si hay
// manifiesto, su versión, la del binario y la lista de hallazgos, vacía si no
// hay ninguno (FR-067). Los hallazgos son datos de una verificación que ha
// funcionado, no un fallo: doctor sale con código 0 con o sin ellos (ADR 0023).
type Diagnostico struct {
	// Directorio es el directorio neutro, como se alcanza desde el directorio
	// de trabajo.
	Directorio string `json:"directorio" jsonschema:"minLength=1"`
	// Manifiesto dice si el ámbito tiene manifiesto.
	Manifiesto bool `json:"manifiesto"`
	// Version es la del manifiesto, o nil sin él.
	Version *string `json:"version" jsonschema:"nullable,minLength=1"`
	// VersionDelBinario es la del binario que diagnostica.
	VersionDelBinario string `json:"version_del_binario" jsonschema:"minLength=1"`
	// Hallazgos es la lista de hallazgos, en el orden de FR-066; vacía si no hay
	// ninguno.
	Hallazgos []Hallazgo `json:"hallazgos"`
}

// Hallazgo es algo que doctor encuentra fuera de su sitio en una skill
// declarada y empotrada: su clase, su ruta, como se alcanza desde el
// directorio de trabajo, y la orden que lo arregla, una sola línea de shell
// POSIX (FR-065, FR-066).
type Hallazgo struct {
	// Clase es la del hallazgo.
	Clase ClaseDeHallazgo `json:"clase" jsonschema:"enum=fichero editado,enum=enlace colgando,enum=enlace a otro sitio,enum=copia,enum=versión distinta"`
	// Ruta es la de lo que está fuera de su sitio.
	Ruta string `json:"ruta" jsonschema:"minLength=1"`
	// Orden es la que lo arregla.
	Orden string `json:"orden" jsonschema:"minLength=1"`
}
