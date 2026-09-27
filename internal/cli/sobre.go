package cli

import (
	"errors"
	"time"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// El espacio de nombres reservado con el que firma el propio kernel cuando
// emite un sobre sin haber llegado a consultar ninguna fuente: los dos fallos
// anteriores a la ejecución del applet —bandera desconocida y applet no
// registrado— y cualquier otro en el que no se conozca la procedencia
// (FR-016, FR-045, contracts/sobre-de-salida.md §3).
const (
	fuenteDelKernel = "kitlegal.cli"
	urlDelKernel    = "kitlegal:cli"
)

// sinSobre dice que una emisión no lleva sobre. Se nombra porque el único sitio
// que lo usa es el camino en el que la salida estándar acaba de fallar, y ahí
// un `false` suelto no contaría por qué.
const sinSobre = false

// errSinFallo lo devuelve Fallo cuando se le pide el sobre de un error que no
// existe. No es un fallo de quien invoca sino un defecto de quien llama, así
// que no lleva ninguno de los seis sentinelas y sale, como todo lo que nadie
// previó, con el código del fallo inesperado (FR-031).
var errSinFallo = errors.New("cli: un sobre de fallo necesita un fallo")

// ProcedenciaKernel es la procedencia de un sobre que emite el kernel sin haber
// llegado a consultar nada. Un consumidor que la vea sabe que está ante un
// resultado calculado y no ante una cita de fuente pública
// (contracts/sobre-de-salida.md §3).
//
// Es una función y no una variable de paquete para que nadie pueda cambiar
// desde fuera la procedencia con la que el kernel firma.
func ProcedenciaKernel() schema.Procedencia {
	return schema.Procedencia{Fuente: fuenteDelKernel, URL: urlDelKernel}
}

// Reloj es la fuente del instante que va en fecha_consulta.
type Reloj func() time.Time

// Montador es el único punto del proyecto que monta un sobre y el único que lo
// emite. Un applet devuelve un schema.Resultado o un error y no conoce ni `ok`,
// ni la huella, ni fecha_consulta, ni los códigos de salida: así la forma del
// sobre es idéntica para todos los applets y también para los fallos anteriores
// a la ejecución de cualquiera de ellos (FR-015, FR-045).
type Montador struct {
	// Ahora es el reloj con el que se fecha el sobre cuando la procedencia no
	// declara la fecha de consulta: la de un applet calculado, la del kernel o
	// la de una fuente que no la conoce (FR-096). Nulo significa el reloj del
	// sistema, de modo que el valor cero del tipo es el montador de producción
	// y solo un test necesita fijarlo: con el instante bajo control se puede
	// comprobar que la huella no depende de él (FR-013, SC-005).
	Ahora Reloj
}

// Emitir convierte el desenlace de una invocación en su código de salida y
// escribe lo que ese desenlace deba escribir. Es el punto único del que habla
// FR-045: la traducción del error a código de salida y la emisión del sobre de
// fallo ocurren aquí, juntas, y nunca en un applet.
//
// res es lo que devolvió el applet —o el valor cero cuando el fallo es anterior
// a su ejecución, en cuyo caso su procedencia no se conoce y firma el kernel—,
// y err es el fallo, o nulo si no lo hubo. enJSON elige la forma del sobre; sin
// él un fallo no lleva sobre, porque la tabla mínima es la forma de un
// resultado y no la de un fallo (contracts/sobre-de-salida.md §5 y §7).
func (m Montador) Emitir(p Presentador, enJSON bool, res schema.Resultado, err error) int {
	if err != nil {
		return m.emitirFallo(p, enJSON, res.Procedencia, err)
	}

	sobre, errMontaje := m.Exito(res)
	if errMontaje != nil {
		// Montar el sobre falla antes de que nada llegue a la salida estándar,
		// así que el descriptor sigue sano y el fallo sale por el camino normal,
		// con su propio sobre (contracts/sobre-de-salida.md §4).
		return m.emitirFallo(p, enJSON, res.Procedencia, errMontaje)
	}

	if errEscritura := p.Presentar(sobre, enJSON); errEscritura != nil {
		// Aquí la salida estándar ya ha fallado. El fallo sale por el mismo
		// camino que cualquier otro pero sin sobre: no se intenta un segundo
		// por el descriptor que acaba de romperse
		// (contracts/banderas-y-exit-codes.md §4).
		return m.emitirFallo(p, sinSobre, res.Procedencia, errEscritura)
	}

	return codigoCorrecto
}

// emitirFallo escribe el sobre de fallo —solo si se pidió la forma legible por
// máquina— y el mensaje para la persona, y devuelve el código de salida de la
// clase del error.
//
// El sobre no sustituye al código de salida ni al mensaje: los duplica en forma
// estructurada (contracts/sobre-de-salida.md §5).
func (m Montador) emitirFallo(
	p Presentador, enJSON bool, proc schema.Procedencia, err error,
) int {
	codigo := CodigoSalida(err)

	if enJSON {
		if errEmision := m.escribirFallo(p, proc, err); errEmision != nil {
			// La emisión del sobre ha fallado, así que el desenlace pasa a ser
			// el inesperado. No se reintenta: queda el mensaje para la persona,
			// que sigue siendo el del fallo y no el de la escritura
			// (contracts/banderas-y-exit-codes.md §4).
			codigo = codigoInesperado
		}
	}

	// Si también falla la salida de error ya no queda descriptor por el que
	// contarlo, y el contrato deja únicamente el código.
	if errAviso := p.Aviso(err.Error()); errAviso != nil {
		return codigoInesperado
	}

	return codigo
}

// escribirFallo monta el sobre de fallo y lo escribe. Devuelve el error de la
// única escritura que hace, sin reintentarla y sin intentar ninguna otra.
func (m Montador) escribirFallo(p Presentador, proc schema.Procedencia, err error) error {
	sobre, errMontaje := m.Fallo(proc, err)
	if errMontaje != nil {
		return errMontaje
	}

	return p.Presentar(sobre, true)
}

// Exito monta el sobre de un resultado correcto: la procedencia que declaró el
// applet, la fecha de la consulta y la huella del contenido en su forma
// canónica.
//
// Devuelve error —y no un sobre a medias— cuando la procedencia no sostendría
// una cita o cuando el contenido no se puede serializar. Las dos cosas las
// provoca el applet y no quien invoca, y las dos ocurren antes de que nada
// llegue a la salida estándar (FR-016, FR-017, contracts/sobre-de-salida.md §4).
func (m Montador) Exito(res schema.Resultado) (schema.Sobre, error) {
	if err := res.Procedencia.Validar(); err != nil {
		return schema.Sobre{}, err
	}

	return m.montar(codigoCorrecto, res.Procedencia, res.Datos)
}

// Fallo monta el sobre de un fallo: las mismas seis claves, `data` con
// exactamente la clase y el mensaje, y la procedencia de la fuente que se
// estaba consultando cuando se conoce (FR-045).
//
// La clase, el mensaje y el código salen todos del mismo error y por la misma
// traducción con la que termina el proceso, que es lo que hace cierto por
// construcción que `ok` sea falso si y solo si el código de salida no es 0
// (FR-014): ninguno de los seis sentinelas ni el fallo inesperado se traducen
// al código del éxito.
func (m Montador) Fallo(proc schema.Procedencia, err error) (schema.Sobre, error) {
	if err == nil {
		return schema.Sobre{}, errSinFallo
	}

	datos := schema.DatosError{Clase: Clasificar(err), Mensaje: err.Error()}

	return m.montar(CodigoSalida(err), procedenciaConocida(proc), datos)
}

// montar es el único sitio donde se escribe la forma del sobre, y por eso el
// de éxito y el de fallo no pueden divergir: la fecha sale de la misma regla y
// la huella, de la misma forma canónica (ADR 0006: «se calculan igual»).
//
// Recibe el código de salida y no un booleano porque la invariante de FR-014 no
// es que `ok` acompañe al desenlace sino que valga exactamente `código == 0`;
// escrita así, no hay forma de emitir un sobre que diga lo contrario de lo que
// dirá el proceso al terminar.
func (m Montador) montar(codigo int, proc schema.Procedencia, datos any) (schema.Sobre, error) {
	huella, err := schema.Huella(datos)
	if err != nil {
		return schema.Sobre{}, err
	}

	return schema.Sobre{
		Ok:            codigo == codigoCorrecto,
		Fuente:        proc.Fuente,
		URL:           proc.URL,
		FechaConsulta: m.fechaDeConsulta(proc),
		Hash:          huella,
		Data:          datos,
	}, nil
}

// fechaDeConsulta es el instante con el que se fecha el sobre: el que declara
// la procedencia cuando quien consultó lo conoce, y el del reloj del montador
// cuando no (FR-096, docs/ADR/0015). Recibe la procedencia ya resuelta —la
// validada del éxito o la conocida del fallo—, así que una procedencia que no
// sostiene una cita llega aquí sustituida por la del kernel, que no trae fecha,
// y el sobre queda fechado con el reloj aunque la rechazada trajera una.
//
// La huella no pasa por aquí: se calcula sobre data, que no lleva la fecha, y
// por eso sigue detectando que el contenido cambió (ADR 0006).
func (m Montador) fechaDeConsulta(proc schema.Procedencia) time.Time {
	if proc.FechaConsulta.IsZero() {
		return m.instante()
	}

	return proc.FechaConsulta
}

// procedenciaConocida devuelve la de la fuente que se estaba consultando cuando
// la hay y la del kernel cuando no. Conocida quiere decir aquí que sostendría
// una cita: un fallo anterior al applet no trae ninguna, y una que no valida no
// podría citarse aunque venga rellena. Es lo que garantiza que el sobre de
// fallo nunca lleve la procedencia vacía, ni siquiera cuando el defecto está en
// el applet (FR-016, FR-045).
func procedenciaConocida(proc schema.Procedencia) schema.Procedencia {
	if proc.Validar() != nil {
		return ProcedenciaKernel()
	}

	return proc
}

// instante lee el reloj del montador, o el del sistema si no se le inyectó
// ninguno. Que el valor cero del tipo funcione es lo que evita un constructor
// al que la raíz de composición tendría que pasarle un nulo.
func (m Montador) instante() time.Time {
	if m.Ahora == nil {
		return time.Now()
	}

	return m.Ahora()
}
