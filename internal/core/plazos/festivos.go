package plazos

import (
	"bytes"
	"cmp"
	"slices"

	"go.yaml.in/yaml/v3"
)

// FestivosNacionales es data/festivos/nacionales-y-autonomicos.yaml, los
// festivos de las resoluciones del BOE (contracts/datos-de-plazos.md §1.1).
type FestivosNacionales struct {
	// Fuente es la de las resoluciones, boe.fiestas-laborales.
	Fuente string `yaml:"fuente"`
	// Anios son los años cargados, cada uno con su resolución.
	Anios []AnioNacional `yaml:"anios"`
	// Festivos son las filas: una por fecha, ámbito y parte.
	Festivos []FestivoNacional `yaml:"festivos"`
}

// AnioNacional es un año cargado y la resolución que da sus festivos.
type AnioNacional struct {
	Anio        int    `yaml:"anio"`
	Disposicion string `yaml:"disposicion"`
}

// FestivoNacional es una fecha que la resolución de su año marca con un ámbito
// en unas comunidades, en toda ellas o, con Parte, solo en una parte.
type FestivoNacional struct {
	// Fecha es AAAA-MM-DD.
	Fecha string `yaml:"fecha"`
	// Ambito es nacional o autonomico.
	Ambito Ambito `yaml:"ambito"`
	// Comunidades son los códigos de las comunidades en que la resolución la
	// marca con ese ámbito.
	Comunidades []string `yaml:"comunidades"`
	// Parte es la parte de la comunidad a la que la resolución limita la
	// fiesta, o vacía si es de toda ella (FR-025).
	Parte string `yaml:"parte,omitempty"`
	// Source es la resolución de su año.
	Source string `yaml:"source"`
}

// FestivosLocales es data/festivos/locales/<comunidad>.yaml, los festivos
// locales de los municipios de una comunidad de las disposiciones de su
// boletín (contracts/datos-de-plazos.md §1.2).
type FestivosLocales struct {
	// Fuente es la de las disposiciones, como bocm.fiestas-locales.
	Fuente string `yaml:"fuente"`
	// Comunidad es el código de la comunidad, el que da nombre al fichero.
	Comunidad string `yaml:"comunidad"`
	// Anios son los años cargados, cada uno con sus disposiciones.
	Anios []AnioLocal `yaml:"anios"`
	// Festivos son las filas: una por fecha y municipio.
	Festivos []FestivoLocal `yaml:"festivos"`
	// Faltas son, por año, los municipios sin festivos y las líneas que no
	// casan con un municipio (FR-022).
	Faltas []FaltaLocal `yaml:"faltas"`
}

// AnioLocal es un año cargado y sus disposiciones, en el orden en que se
// publicaron: la última es hasta donde llegan los datos del año.
type AnioLocal struct {
	Anio          int      `yaml:"anio"`
	Disposiciones []string `yaml:"disposiciones"`
}

// FestivoLocal es una fecha que es fiesta local de un municipio.
type FestivoLocal struct {
	// Fecha es AAAA-MM-DD.
	Fecha string `yaml:"fecha"`
	// Municipio es su código INE.
	Municipio string `yaml:"municipio"`
	// Source es la última disposición del año que fija la línea del municipio.
	Source string `yaml:"source"`
}

// FaltaLocal es, en un año, un municipio sin festivos o una línea del boletín
// que no casa con un único municipio, y su motivo, uno de MotivosDeFalta.
type FaltaLocal struct {
	Anio int `yaml:"anio"`
	// Municipio es el código INE del municipio sin festivos, o vacío en una
	// línea sin municipio.
	Municipio string `yaml:"municipio,omitempty"`
	// Linea es el nombre tal como lo escribe el boletín, solo con
	// FaltaSinMunicipio.
	Linea  string `yaml:"linea,omitempty"`
	Motivo Falta  `yaml:"motivo"`
	// Source es la disposición de la línea, o la última del año con
	// FaltaSinLinea.
	Source string `yaml:"source"`
}

// Cabecera es lo que nombra la línea de fichero generado: el directorio de lo
// grabado del que sale el fichero y el test que lo deriva, con su paquete.
type Cabecera struct {
	// Desde es el directorio de lo grabado, como evidencias/h9/boe.fiestas-laborales/.
	Desde string
	// Test es el test de derivación que escribe el fichero.
	Test string
	// Paquete es el del test, como internal/source/boe/fiestas.
	Paquete string
}

// linea es la primera línea del fichero, un comentario de fichero generado.
func (c Cabecera) linea() string {
	return "# generado desde " + c.Desde + " por " + c.Test + " (" + c.Paquete + "), no editar\n"
}

// EscribirFestivosNacionales escribe data/festivos/nacionales-y-autonomicos.yaml
// en su forma canónica: la línea de fichero generado, las claves en el orden
// del contrato, una fila por línea, los años por año, las filas por fecha,
// ámbito —nacional antes que autonómico— y parte —sin parte antes—, y las
// comunidades de cada fila ordenadas. No cambia lo que recibe.
func EscribirFestivosNacionales(fichero FestivosNacionales, cabecera Cabecera) ([]byte, error) {
	canonico := FestivosNacionales{
		Fuente:   fichero.Fuente,
		Anios:    slices.Clone(fichero.Anios),
		Festivos: make([]FestivoNacional, 0, len(fichero.Festivos)),
	}

	slices.SortStableFunc(canonico.Anios, func(a, b AnioNacional) int { return cmp.Compare(a.Anio, b.Anio) })

	for _, festivo := range fichero.Festivos {
		festivo.Comunidades = slices.Sorted(slices.Values(festivo.Comunidades))
		canonico.Festivos = append(canonico.Festivos, festivo)
	}

	ambitos := Ambitos()

	slices.SortStableFunc(canonico.Festivos, func(a, b FestivoNacional) int {
		return cmp.Or(
			cmp.Compare(a.Fecha, b.Fecha),
			cmp.Compare(slices.Index(ambitos, a.Ambito), slices.Index(ambitos, b.Ambito)),
			cmp.Compare(a.Parte, b.Parte),
		)
	})

	return escribir(canonico, cabecera)
}

// EscribirFestivosLocales escribe data/festivos/locales/<comunidad>.yaml en su
// forma canónica: la línea de fichero generado, las claves en el orden del
// contrato, una fila por línea, los años por año con sus disposiciones en el
// orden dado, los festivos por fecha y municipio, y las faltas por año,
// municipio y línea, las de un municipio antes que las de una línea. No cambia
// lo que recibe.
func EscribirFestivosLocales(fichero FestivosLocales, cabecera Cabecera) ([]byte, error) {
	canonico := FestivosLocales{
		Fuente:    fichero.Fuente,
		Comunidad: fichero.Comunidad,
		Anios:     slices.Clone(fichero.Anios),
		Festivos:  slices.Clone(fichero.Festivos),
		Faltas:    slices.Clone(fichero.Faltas),
	}

	slices.SortStableFunc(canonico.Anios, func(a, b AnioLocal) int { return cmp.Compare(a.Anio, b.Anio) })
	slices.SortStableFunc(canonico.Festivos, func(a, b FestivoLocal) int {
		return cmp.Or(cmp.Compare(a.Fecha, b.Fecha), cmp.Compare(a.Municipio, b.Municipio))
	})
	slices.SortStableFunc(canonico.Faltas, func(a, b FaltaLocal) int {
		return cmp.Or(cmp.Compare(a.Anio, b.Anio), vaciosAlFinal(a.Municipio, b.Municipio), cmp.Compare(a.Linea, b.Linea))
	})

	return escribir(canonico, cabecera)
}

// vaciosAlFinal compara dos textos con el vacío después de cualquier otro.
func vaciosAlFinal(a, b string) int {
	switch {
	case a == b:
		return 0
	case a == "":
		return 1
	case b == "":
		return -1
	default:
		return cmp.Compare(a, b)
	}
}

// escribir codifica un fichero de data/festivos/ tras su línea de fichero
// generado, con sangría de dos espacios y cada elemento de sus listas en una
// línea, en estilo de flujo. Sin listas de más de una línea, el fichero se lee
// y se compara por filas.
func escribir(fichero any, cabecera Cabecera) ([]byte, error) {
	var raiz yaml.Node
	if err := raiz.Encode(fichero); err != nil {
		return nil, err
	}

	for indice := 1; indice < len(raiz.Content); indice += 2 {
		if lista := raiz.Content[indice]; lista.Kind == yaml.SequenceNode {
			for _, fila := range lista.Content {
				fila.Style = yaml.FlowStyle
			}
		}
	}

	var salida bytes.Buffer

	salida.WriteString(cabecera.linea())

	codificador := yaml.NewEncoder(&salida)
	codificador.SetIndent(2)

	if err := codificador.Encode(&raiz); err != nil {
		return nil, err
	}

	if err := codificador.Close(); err != nil {
		return nil, err
	}

	return salida.Bytes(), nil
}
