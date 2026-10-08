package cita

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmorenobl/kitlegal/internal/core/ids"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Las tres formas de dar una referencia (H23, FR-005), que son los valores de
// `forma` en lo que el paquete devuelve. Las etiquetas jsonschema de los tipos
// repiten este vocabulario, porque una etiqueta no puede nombrar una
// constante, y los tests exigen que digan lo mismo.
const (
	// formaECLI es la del ECLI de una resolución española.
	formaECLI = "ecli"
	// formaROJ es la del ROJ, el identificador de la resolución en el CENDOJ.
	formaROJ = "roj"
	// formaResolucion es la del número de resolución, que va con su fecha.
	formaResolucion = "resolucion"
)

// Cómo nombra un mensaje cada forma dada.
const (
	unECLI   = "un ECLI"
	unROJ    = "un ROJ"
	unNumero = "un número de resolución"
)

// Las dos escrituras de una fecha, como disposiciones del paquete time
// (research.md D8).
const (
	// formaDeLaFecha es AAAA-MM-DD: la del argumento y la de toda fecha que el
	// paquete devuelve, también la de la ficha.
	formaDeLaFecha = "2006-01-02"
	// formaDeLaFechaDelCENDOJ es dd/mm/aaaa: la de la ficha de un documento y
	// la que escribe quien rellena una casilla del buscador.
	formaDeLaFechaDelCENDOJ = "02/01/2006"
)

// Las partes de un número de resolución, <número>/<año>: las mismas que van
// tras las siglas de un ROJ.
const (
	// barraDelNumero va entre el número y el año.
	barraDelNumero = "/"
	// cifrasDelAnio son las del año.
	cifrasDelAnio = 4
)

// Referencia es lo que identifica la sentencia que se busca o que se pidió,
// dada de una sola forma de tres: su ECLI, su ROJ o su número de resolución
// con su fecha. Solo se construye con NuevaReferencia, que la comprueba; su
// valor cero no es ninguna referencia.
type Referencia struct {
	// Forma es ecli, roj o resolucion.
	Forma string `json:"forma" jsonschema:"enum=ecli,enum=roj,enum=resolucion"`
	// Valor es el ECLI, el ROJ o el número de resolución, tal como se dio.
	Valor string `json:"valor" jsonschema:"minLength=1"`
	// Fecha es la de la resolución, AAAA-MM-DD: solo va con un número.
	Fecha string `json:"fecha,omitempty" jsonschema:"pattern=^[0-9]{4}-[0-9]{2}-[0-9]{2}$"`

	// ecli es el ECLI analizado, si la forma es esa: da su órgano y su
	// equivalente.
	ecli ids.ECLI
	// roj es el ROJ analizado, si la forma es esa: da su número y su
	// equivalente.
	roj ids.ROJ
	// dia es el de Fecha, para escribirlo como lo pide una casilla.
	dia time.Time
}

// NuevaReferencia reconoce la referencia que dan los cuatro argumentos de los
// dos verbos: un ECLI, un ROJ, un número de resolución y la fecha de ese
// número. Cada uno es nil si no se escribió; el que se escribió, también con
// valor vacío, está dado, se comprueba con su forma y cuenta como una forma
// dada (research.md D3). La entrada no se recorta ni se pasa a mayúsculas.
//
// Sin ninguna forma dada no hay referencia, y hay es falso: no es un error
// aquí, y quien llama decide si lo es. Una referencia mal dada es un error de
// clase «argumentos», en este orden: más de una forma a la vez; una fecha sin
// número; la forma dada mal escrita; un número sin fecha; y una fecha que no es
// AAAA-MM-DD o que no es un día que existe (H23, FR-005, FR-006).
func NuevaReferencia(ecli, roj, resolucion, fecha *string) (referencia Referencia, hay bool, err error) {
	if err = unaSolaForma(ecli, roj, resolucion); err != nil {
		return Referencia{}, false, err
	}

	if fecha != nil && resolucion == nil {
		return Referencia{}, false, argumentosInvalidos(
			"la fecha %q solo vale con un número de resolución, y no se ha dado ninguno", *fecha)
	}

	switch {
	case ecli != nil:
		referencia, err = referenciaPorECLI(*ecli)
	case roj != nil:
		referencia, err = referenciaPorROJ(*roj)
	case resolucion != nil:
		referencia, err = referenciaPorNumero(*resolucion, fecha)
	default:
		return Referencia{}, false, nil
	}

	if err != nil {
		return Referencia{}, false, err
	}

	return referencia, true, nil
}

// unaSolaForma rechaza la referencia dada de más de una forma, y nombra las
// que se han dado.
func unaSolaForma(ecli, roj, resolucion *string) error {
	var dadas []string

	if ecli != nil {
		dadas = append(dadas, unECLI)
	}

	if roj != nil {
		dadas = append(dadas, unROJ)
	}

	if resolucion != nil {
		dadas = append(dadas, unNumero)
	}

	if len(dadas) < 2 {
		return nil
	}

	ultima := len(dadas) - 1

	return argumentosInvalidos("la referencia se da de una sola forma y se han dado %s y %s",
		strings.Join(dadas[:ultima], ", "), dadas[ultima])
}

// referenciaPorECLI es la referencia de un ECLI español. Su rechazo es el del
// analizador, que nombra la entrada y dice qué tiene de malo, y que el de otro
// país no es español.
func referenciaPorECLI(entrada string) (Referencia, error) {
	ecli, err := ids.AnalizarECLI(entrada)
	if err != nil {
		return Referencia{}, err
	}

	return Referencia{Forma: formaECLI, Valor: entrada, ecli: ecli}, nil
}

// referenciaPorROJ es la referencia de un ROJ, con el rechazo de su
// analizador.
func referenciaPorROJ(entrada string) (Referencia, error) {
	roj, err := ids.AnalizarROJ(entrada)
	if err != nil {
		return Referencia{}, err
	}

	return Referencia{Forma: formaROJ, Valor: entrada, roj: roj}, nil
}

// referenciaPorNumero es la referencia de un número de resolución con su
// fecha: el número tiene su forma, la fecha está dada y es un día que existe.
func referenciaPorNumero(numero string, fecha *string) (Referencia, error) {
	if !esNumeroConAnio(numero) {
		return Referencia{}, argumentosInvalidos("el número de resolución %q no es válido: la forma es "+
			"<número>/<año>, en cifras y con el año de cuatro, como 1088/2023", numero)
	}

	if fecha == nil {
		return Referencia{}, argumentosInvalidos("el número de resolución %s necesita su fecha: un número solo "+
			"identifica una sentencia con el día en que se dictó, AAAA-MM-DD", numero)
	}

	dia, err := time.Parse(formaDeLaFecha, *fecha)
	if err != nil {
		return Referencia{}, argumentosInvalidos("la fecha %q no es válida: la forma es AAAA-MM-DD y tiene que ser "+
			"un día que existe, como 2023-07-04", *fecha)
	}

	return Referencia{Forma: formaResolucion, Valor: numero, Fecha: *fecha, dia: dia}, nil
}

// esNumeroConAnio dice si el texto es un número de resolución, <número>/<año>:
// una o más cifras, una barra y las cuatro cifras del año, sin nada delante ni
// detrás. Lo comparten la referencia y la lectura de la ficha.
func esNumeroConAnio(texto string) bool {
	numero, anio, conBarra := strings.Cut(texto, barraDelNumero)

	return conBarra && numero != "" && sonCifras(numero) && len(anio) == cifrasDelAnio && sonCifras(anio)
}

// sonCifras dice si todo el texto son cifras ASCII, del 0 al 9. Mira bytes y no
// runas, como internal/core/ids: las cifras de otras escrituras no valen.
func sonCifras(texto string) bool {
	for indice := range len(texto) {
		if texto[indice] < '0' || texto[indice] > '9' {
			return false
		}
	}

	return true
}

// errorDeArgumentos es el rechazo de lo que se le da al dominio de la cita:
// una referencia mal dada, un texto de búsqueda vacío o un texto sin ficha
// reconocible. No se exporta: quien lo recibe lo reconoce por su clase, con
// errors.As a schema.ConClase, que es como lo reconoce el kernel. Los rechazos
// que decide la forma de un ECLI o de un ROJ los da internal/core/ids con su
// propio tipo, de la misma clase.
type errorDeArgumentos struct {
	// mensaje nombra lo recibido y dice qué le falta.
	mensaje string
}

// Un rechazo del dominio de la cita declara su clase él mismo.
var _ schema.ConClase = (*errorDeArgumentos)(nil)

// argumentosInvalidos construye el rechazo con su mensaje.
func argumentosInvalidos(formato string, valores ...any) error {
	return &errorDeArgumentos{mensaje: fmt.Sprintf(formato, valores...)}
}

// Error devuelve el mensaje, que llega literal al sobre de fallo y a la salida
// de error.
func (e *errorDeArgumentos) Error() string {
	return e.mensaje
}

// Clase es siempre «argumentos»: lo que se rechaza lo escribió quien invoca, y
// el kernel lo traduce a código 2 (H23, FR-006, FR-015, FR-026; research.md
// D9).
func (*errorDeArgumentos) Clase() schema.Clase {
	return schema.ClaseArgumentos
}
