package plazos

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// errorDeArgumentos es el rechazo de una petición que decide el dominio: un
// tipo, una fecha de inicio, una duración, una unidad o un número de
// municipios que no valen, o un plazo que acaba después del último día que se
// escribe AAAA-MM-DD (contracts/applet-plazos.md §6; research D10). No se
// exporta: quien lo recibe lo reconoce por su clase, con errors.As a
// schema.ConClase, que es como lo reconoce el kernel, igual que los de
// territorio.
type errorDeArgumentos struct {
	// mensaje nombra la entrada con %q y dice qué le pasa.
	mensaje string
}

// Un rechazo de argumentos declara su clase él mismo.
var _ schema.ConClase = (*errorDeArgumentos)(nil)

// Error devuelve el mensaje, que llega literal al sobre de fallo y a la salida
// de error.
func (e *errorDeArgumentos) Error() string {
	return e.mensaje
}

// Clase es siempre argumentos, que el kernel traduce al código 2 (FR-012).
func (e *errorDeArgumentos) Clase() schema.Clase {
	return schema.ClaseArgumentos
}

// errorNoEncontrado es el plazo que kitlegal no tiene: el de una fecha de inicio
// en la que alguno de sus preceptos no regía como lo cuenta data/plazos.yaml
// (ADR 0042). Lo pedido, el plazo de ese día, no está en los datos del applet,
// que es su fuente: la fila «no encontrado» del ADR 0023, como la lectura de un
// bloque en una fecha sin redacción vigente.
type errorNoEncontrado struct {
	// mensaje dice el plazo, por qué no se calcula y, una por línea, cada
	// precepto con el día desde el que rige.
	mensaje string
}

// Un plazo que no se tiene declara su clase él mismo.
var _ schema.ConClase = (*errorNoEncontrado)(nil)

// Error devuelve el mensaje, que llega literal al sobre de fallo y a la salida
// de error.
func (e *errorNoEncontrado) Error() string {
	return e.mensaje
}

// Clase es siempre no encontrado, que el kernel traduce al código 3.
func (e *errorNoEncontrado) Clase() schema.Clase {
	return schema.ClaseNoEncontrado
}

// argumentos construye un rechazo de argumentos con su mensaje.
func argumentos(formato string, valores ...any) error {
	return &errorDeArgumentos{mensaje: fmt.Sprintf(formato, valores...)}
}

// tipoDesconocido es el tipo que no es ninguno de los de data/plazos.yaml.
func tipoDesconocido(tipo string, tipos []Tipo) error {
	codigos := make([]string, 0, len(tipos))
	for _, conocido := range tipos {
		codigos = append(codigos, conocido.Codigo)
	}

	return argumentos("el tipo %q no es %s", tipo, enumeracion(codigos))
}

// inicioMalEscrito es la fecha de inicio que no tiene la forma AAAA-MM-DD.
func inicioMalEscrito(inicio string) error {
	return argumentos("la fecha de inicio %q no es una fecha AAAA-MM-DD", inicio)
}

// inicioFueraDelCalendario es la fecha de inicio con la forma AAAA-MM-DD que no
// es un día del calendario, como un 30 de febrero.
func inicioFueraDelCalendario(inicio string) error {
	return argumentos("la fecha de inicio %q no es un día del calendario", inicio)
}

// duracionQueFalta es la del silencio o el genérico, que no fijan su duración.
func duracionQueFalta(tipo string) error {
	return argumentos("el tipo %q no fija su duración y no se ha dado: pide una duración, un entero positivo, con su "+
		"unidad", tipo)
}

// duracionNoPositiva es la duración que no es un entero de 1 o más, o que no
// cabe en un entero (research D28).
func duracionNoPositiva(duracion string) error {
	return argumentos("la duración %q no es un entero positivo", duracion)
}

// duracionDeMas es la que se da a un tipo que fija la suya.
func duracionDeMas(tipo, duracion string) error {
	return argumentos("el tipo %q fija su duración y no admite otra: sobra la duración %q", tipo, duracion)
}

// unidadQueFalta es la de una duración dada sin unidad.
func unidadQueFalta(duracion string) error {
	return argumentos("la duración %q no lleva su unidad, que es %s", duracion, alternativas(Unidades()))
}

// unidadDesconocida es la unidad que no es del vocabulario.
func unidadDesconocida(unidad string) error {
	return argumentos("la unidad %q no es %s", unidad, enumeracion(Unidades()))
}

// unidadDeMas es la que se da a un tipo que fija su duración y su unidad.
func unidadDeMas(tipo, unidad string) error {
	return argumentos("el tipo %q fija su duración y su unidad: sobra la unidad %q", tipo, unidad)
}

// municipiosQueNoCuentan es un número de municipios distinto del de su
// cómputo: dos en la LPAC (art. 30.6; FR-007) y uno en el contencioso (FR-010).
func municipiosQueNoCuentan(tipo Tipo, dados int) error {
	if tipo.Computo == ComputoContencioso {
		return argumentos("el tipo %q se cuenta con un municipio, el de la sede del juzgado, y se han dado %d",
			tipo.Codigo, dados)
	}

	return argumentos("el tipo %q se cuenta con dos municipios, el del interesado y el de la sede del órgano, que "+
		"pueden ser el mismo, y se han dado %d", tipo.Codigo, dados)
}

// plazoFueraDelCalendario es el plazo que acaba después del 31 de diciembre de
// 9999: su fecha no se escribe AAAA-MM-DD (contracts/applet-plazos.md §4).
func plazoFueraDelCalendario(peticion Peticion) error {
	var conDuracion string
	if peticion.Duracion != "" {
		conDuracion = fmt.Sprintf(" con la duración %q %s", peticion.Duracion, peticion.Unidad)
	}

	return argumentos("el plazo de tipo %q que empieza el %q%s acaba después del %s, el último día que se escribe "+
		"AAAA-MM-DD", peticion.Tipo, peticion.Inicio, conDuracion, ultimoDia.Format(time.DateOnly))
}

// plazoConPreceptosPosteriores es el plazo que empieza antes del día desde el
// que rige alguno de sus preceptos (ADR 0042): su primera línea dice el plazo,
// su fecha de inicio y por qué no se calcula, y cada una de las siguientes, en
// el orden dado, un precepto con su norma, su bloque, su apartado y ese día.
func plazoConPreceptosPosteriores(comprobada peticionComprobada, posteriores []Precepto) error {
	var mensaje strings.Builder

	fmt.Fprintf(&mensaje, "el plazo de tipo %q que empieza el %s no se calcula: un plazo se rige por la redacción "+
		"vigente el día en que empieza, y kitlegal lo cuenta con preceptos que ese día no decían lo que dicen hoy; "+
		"qué regía lo dicen la redacción de entonces y el régimen transitorio de la norma que la cambió, y kitlegal "+
		"no calcula con ellos:", comprobada.tipo.Codigo, comprobada.inicio.Format(time.DateOnly))

	for _, precepto := range posteriores {
		fmt.Fprintf(&mensaje, "\n%s, bloque %s, apartado %s: rige así desde el %s", precepto.Norma, precepto.Bloque,
			precepto.Apartado, precepto.VigenteDesde)
	}

	return &errorNoEncontrado{mensaje: mensaje.String()}
}

// alternativas escribe un vocabulario como «a, b o c».
func alternativas[T ~string](valores []T) string {
	var texto strings.Builder

	for indice, valor := range valores {
		switch {
		case indice == 0:
		case indice == len(valores)-1:
			texto.WriteString(" o ")
		default:
			texto.WriteString(", ")
		}

		texto.WriteString(string(valor))
	}

	return texto.String()
}
