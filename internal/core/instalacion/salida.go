package instalacion

// Los tipos de la salida de install, el data del verbo, con las claves de
// data-model §10 en español (contracts/applet-skills.md §4.1; FR-051).
// Ninguno lleva omitempty y una lista vacía es una lista, nunca nil, así que
// se emite [] y no null.
//
// Las etiquetas jsonschema son la descripción formal de los valores, de la que
// --describe genera el esquema del verbo y de la que sale el publicado
// (research.md D15, D16). Una etiqueta no puede nombrar una constante, así que
// los enumerados repiten los vocabularios de Estado, Modo y el host, y TestPlan
// exige que digan lo mismo.

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
