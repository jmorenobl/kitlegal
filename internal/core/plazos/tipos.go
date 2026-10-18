package plazos

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
	"slices"

	"go.yaml.in/yaml/v3"
)

// rutaDeLosTipos es la del fichero de los tipos, relativa a la raíz del
// repositorio, la que nombran sus defectos.
const rutaDeLosTipos = "data/plazos.yaml"

// Los siete códigos de tipo de FR-008 (contracts/datos-de-plazos.md §2).
const (
	TipoAlzada         = "alzada"
	TipoReposicion     = "reposicion"
	TipoContencioso    = "contencioso"
	TipoAccesoResolver = "acceso-resolver"
	TipoAccesoReclamar = "acceso-reclamar"
	TipoSilencio       = "silencio"
	TipoGenerico       = "generico"
)

// CodigosDeTipo son los siete códigos de tipo, en el orden del contrato.
// Devuelve una lista nueva en cada llamada.
func CodigosDeTipo() []string {
	return []string{
		TipoAlzada, TipoReposicion, TipoContencioso, TipoAccesoResolver, TipoAccesoReclamar, TipoSilencio, TipoGenerico,
	}
}

// fijaSuDuracion dice si un tipo fija su duración: todos menos el silencio y
// el genérico, que la reciben con su unidad (research D3).
func fijaSuDuracion(codigo string) bool {
	return codigo != TipoSilencio && codigo != TipoGenerico
}

// Computo es la regla con que se cuenta un plazo.
type Computo string

// Los dos cómputos: el de la LPAC (art. 30) y el del recurso
// contencioso-administrativo (arts. 46 y 128 LJCA, 182 a 185 LOPJ).
const (
	ComputoLPAC        Computo = "lpac"
	ComputoContencioso Computo = "contencioso"
)

// Computos es el vocabulario de cómputos. Devuelve una lista nueva en cada
// llamada.
func Computos() []Computo {
	return []Computo{ComputoLPAC, ComputoContencioso}
}

// Unidad es la unidad en que se expresa una duración.
type Unidad string

// Las tres unidades de una duración (art. 30 LPAC).
const (
	UnidadMeses         Unidad = "meses"
	UnidadDiasHabiles   Unidad = "dias-habiles"
	UnidadDiasNaturales Unidad = "dias-naturales"
)

// Unidades es el vocabulario de unidades. Devuelve una lista nueva en cada
// llamada.
func Unidades() []Unidad {
	return []Unidad{UnidadMeses, UnidadDiasHabiles, UnidadDiasNaturales}
}

// FijaDuracion es lo que fija el precepto que da la duración de un tipo.
const FijaDuracion = "duracion"

// Tipo es un tipo de plazo de data/plazos.yaml (data-model §1).
type Tipo struct {
	// Codigo es uno de CodigosDeTipo.
	Codigo string `yaml:"codigo"`
	// Nombre es el que lo describe.
	Nombre string `yaml:"nombre"`
	// Computo es la regla con que se cuenta.
	Computo Computo `yaml:"computo"`
	// Duracion es la que fija el tipo, o nil si la recibe con su unidad.
	Duracion *Duracion `yaml:"duracion"`
	// Preceptos son los que lo fundan, cada uno con lo que fija.
	Preceptos []Precepto `yaml:"preceptos"`
	// Unidades lleva, solo en el genérico, de cada unidad al precepto que da su
	// cómputo.
	Unidades map[Unidad]Precepto `yaml:"unidades"`
}

// Duracion es la duración que fija un tipo.
type Duracion struct {
	// Cantidad es un entero de 1 o más.
	Cantidad int `yaml:"cantidad"`
	// Unidad es una de Unidades.
	Unidad Unidad `yaml:"unidad"`
}

// Precepto es un apartado de una norma consolidada del BOE y lo que fija.
type Precepto struct {
	// Norma es el identificador BOE-A-… de la norma.
	Norma string `yaml:"norma"`
	// Bloque es el id del bloque en el índice consolidado de la norma.
	Bloque string `yaml:"bloque"`
	// Apartado es el que se cita, como «122.1».
	Apartado string `yaml:"apartado"`
	// Fija es lo que fija el precepto en el tipo: su duración (FijaDuracion),
	// su inicio, su cómputo…
	Fija string `yaml:"fija"`
}

// ficheroDeTipos es data/plazos.yaml.
type ficheroDeTipos struct {
	Tipos []Tipo `yaml:"tipos"`
}

// LeerTipos lee data/plazos.yaml (contracts/datos-de-plazos.md §2). Exige los
// siete códigos de CodigosDeTipo y ninguno más, cada uno una vez, con su nombre
// y un cómputo de Computos; una duración de 1 o más en una de Unidades en los
// cinco tipos que la fijan y en ningún otro, y en cada uno de ellos un precepto
// que fija la duración; las tres unidades, cada una con su precepto, en el
// genérico y en ningún otro; y que cada tipo nombre algún precepto, cada uno
// con su norma, su bloque, su apartado y lo que fija.
//
// Los tipos se devuelven en el orden del fichero. Con cualquier defecto no hay
// tipos: el error los dice todos, cada uno con su tipo, y no lleva clase de
// usuario; el kernel lo da como inesperado.
func LeerTipos(contenido []byte) ([]Tipo, error) {
	var fichero ficheroDeTipos
	if err := decodificar(rutaDeLosTipos, contenido, &fichero); err != nil {
		return nil, err
	}

	var c comprobacionDeTipos

	c.comprobar(fichero.Tipos)

	if err := errors.Join(c.defectos...); err != nil {
		return nil, err
	}

	return fichero.Tipos, nil
}

// comprobacionDeTipos acumula los defectos de los tipos, en el orden del
// fichero y, al final, los códigos que faltan.
type comprobacionDeTipos struct {
	defectos []error
}

// defecto anota un defecto de data/plazos.yaml.
func (c *comprobacionDeTipos) defecto(formato string, argumentos ...any) {
	c.defectos = append(c.defectos, &defectoDeDatos{fichero: rutaDeLosTipos, motivo: fmt.Sprintf(formato, argumentos...)})
}

// comprobar comprueba cada tipo y que estén los siete, cada uno una vez.
func (c *comprobacionDeTipos) comprobar(tipos []Tipo) {
	codigos := CodigosDeTipo()
	vistos := make(map[string]bool, len(tipos))

	for indice, tipo := range tipos {
		switch {
		case tipo.Codigo == "":
			c.defecto("el tipo %d no tiene código", indice+1)
		case !slices.Contains(codigos, tipo.Codigo):
			c.defecto("el tipo %q no es uno de los siete", tipo.Codigo)
		case vistos[tipo.Codigo]:
			c.defecto("el tipo %q está repetido", tipo.Codigo)
		}

		vistos[tipo.Codigo] = true

		c.tipo(tipo)
	}

	for _, codigo := range codigos {
		if !vistos[codigo] {
			c.defecto("falta el tipo %q", codigo)
		}
	}
}

// tipo comprueba el nombre, el cómputo, la duración, las unidades y los
// preceptos de un tipo.
func (c *comprobacionDeTipos) tipo(tipo Tipo) {
	if tipo.Nombre == "" {
		c.defecto("el tipo %q no tiene nombre", tipo.Codigo)
	}

	if !slices.Contains(Computos(), tipo.Computo) {
		c.defecto("el cómputo %q del tipo %q no es %s ni %s", tipo.Computo, tipo.Codigo, ComputoLPAC, ComputoContencioso)
	}

	c.duracion(tipo)
	c.unidades(tipo)

	if len(tipo.Preceptos) == 0 && len(tipo.Unidades) == 0 {
		c.defecto("el tipo %q no nombra ningún precepto", tipo.Codigo)
	}

	for indice, precepto := range tipo.Preceptos {
		c.precepto(fmt.Sprintf("el precepto %d del tipo %q", indice+1, tipo.Codigo), precepto)
	}
}

// duracion comprueba que el tipo lleva duración si y solo si la fija, y que la
// que lleva es válida y la fija uno de sus preceptos.
func (c *comprobacionDeTipos) duracion(tipo Tipo) {
	duracion := tipo.Duracion

	switch {
	case duracion == nil && fijaSuDuracion(tipo.Codigo):
		c.defecto("el tipo %q fija su duración y no la lleva", tipo.Codigo)

		return
	case duracion == nil:
		return
	case !fijaSuDuracion(tipo.Codigo):
		c.defecto("el tipo %q no fija su duración y lleva una", tipo.Codigo)
	}

	if duracion.Cantidad < 1 {
		c.defecto("la duración del tipo %q no es un entero de 1 o más", tipo.Codigo)
	}

	if !slices.Contains(Unidades(), duracion.Unidad) {
		c.defecto("la unidad %q del tipo %q no es %s", duracion.Unidad, tipo.Codigo, enumeracion(Unidades()))
	}

	fijada := slices.ContainsFunc(tipo.Preceptos, func(p Precepto) bool { return p.Fija == FijaDuracion })
	if !fijada {
		c.defecto("ningún precepto del tipo %q fija su duración", tipo.Codigo)
	}
}

// unidades comprueba que solo el genérico lleva unidades, y que lleva las
// tres, cada una con su precepto, y ninguna más.
func (c *comprobacionDeTipos) unidades(tipo Tipo) {
	if tipo.Codigo != TipoGenerico {
		if tipo.Unidades != nil {
			c.defecto("el tipo %q lleva unidades, que solo lleva %s", tipo.Codigo, TipoGenerico)
		}

		return
	}

	if tipo.Unidades == nil {
		c.defecto("el tipo %q no lleva unidades", tipo.Codigo)

		return
	}

	for _, unidad := range Unidades() {
		precepto, esta := tipo.Unidades[unidad]
		if !esta {
			c.defecto("el tipo %q no lleva la unidad %q", tipo.Codigo, unidad)

			continue
		}

		c.precepto(fmt.Sprintf("el precepto de la unidad %q del tipo %q", unidad, tipo.Codigo), precepto)
	}

	for _, unidad := range slices.Sorted(maps.Keys(tipo.Unidades)) {
		if !slices.Contains(Unidades(), unidad) {
			c.defecto("el tipo %q lleva la unidad %q, que no es %s", tipo.Codigo, unidad, enumeracion(Unidades()))
		}
	}
}

// precepto comprueba que un precepto lleva sus cuatro datos.
func (c *comprobacionDeTipos) precepto(quien string, precepto Precepto) {
	for _, dato := range []struct{ nombre, valor string }{
		{"norma", precepto.Norma},
		{"bloque", precepto.Bloque},
		{"apartado", precepto.Apartado},
		{"fija", precepto.Fija},
	} {
		if dato.valor == "" {
			c.defecto("%s no tiene %s", quien, dato.nombre)
		}
	}
}

// defectoDeDatos es un defecto de un fichero de data/ empotrado: uno que no se
// lee o un dato que no vale. No declara ninguna clase a propósito —el kernel lo
// trata como inesperado, código 1—: un fichero empotrado que no se lee es un
// defecto del binario y no algo que quien pregunta pueda corregir.
type defectoDeDatos struct {
	// fichero es la ruta del fichero, relativa a la raíz del repositorio.
	fichero string
	// motivo dice qué falla y nombra el dato.
	motivo string
}

// Error nombra el fichero y el defecto, como «data/plazos.yaml: falta el tipo
// "silencio"».
func (d *defectoDeDatos) Error() string {
	return d.fichero + ": " + d.motivo
}

// decodificar lee un fichero YAML en su tipo. Una clave que el tipo no declara
// es un defecto, no algo que se descarta en silencio, y la clave repetida la
// rechaza el propio lector.
func decodificar(ruta string, contenido []byte, destino any) error {
	if len(bytes.TrimSpace(contenido)) == 0 {
		return &defectoDeDatos{fichero: ruta, motivo: "está vacío"}
	}

	lector := yaml.NewDecoder(bytes.NewReader(contenido))
	lector.KnownFields(true)

	if err := lector.Decode(destino); err != nil {
		return &defectoDeDatos{fichero: ruta, motivo: "no se puede leer: " + err.Error()}
	}

	return nil
}

// enumeracion escribe un vocabulario como «a, b ni c».
func enumeracion[T ~string](valores []T) string {
	var texto string

	for indice, valor := range valores {
		switch {
		case indice == 0:
		case indice == len(valores)-1:
			texto += " ni "
		default:
			texto += ", "
		}

		texto += string(valor)
	}

	return texto
}
