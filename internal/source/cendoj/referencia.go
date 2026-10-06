package cendoj

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jmorenobl/kitlegal/internal/core/ids"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// forma es la forma en que se da una referencia, que decide el campo del
// formulario por el que se consulta. Su texto es el que lleva la clave de la
// caché (contrato cita-resolver §5).
type forma string

// Las tres formas de referencia de cita resolver (FR-002).
const (
	// formaECLI es la de un ECLI, que identifica él solo la resolución.
	formaECLI forma = "ecli"
	// formaROJ es la de un ROJ, con la fecha de la resolución o sin ella.
	formaROJ forma = "roj"
	// formaResolucion es la de un número de resolución, que solo identifica con
	// su fecha: cada órgano numera las suyas.
	formaResolucion forma = "resolucion"
)

// formasAdmitidas es lo que los rechazos dicen que vale como referencia, con la
// escritura de cada forma en la orden.
const formasAdmitidas = "un ECLI, un ROJ (--roj) o un número de resolución con su fecha (--resolucion y --fecha)"

// organoDelTribunalConstitucional es el código de órgano de los ECLI del
// Tribunal Constitucional, que no está en esta fuente (FR-015).
const organoDelTribunalConstitucional = "TC"

// Las formas de un número de resolución y de una fecha (FR-005; contrato
// cita-resolver §1). Solo admiten cifras ASCII, y nada delante ni detrás: ni un
// blanco, ni un salto de línea.
var (
	// formaDelNumeroDeResolucion es <número>/<año>: una o más cifras, una barra
	// y cuatro cifras, como 1088/2023.
	formaDelNumeroDeResolucion = regexp.MustCompile(`^[0-9]+/[0-9]{4}$`)
	// formaDeLaFecha es AAAA-MM-DD. Que el día exista lo dice después el
	// calendario.
	formaDeLaFecha = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}$`)
)

// Referencia es la referencia de una resolución tal como la pide cita resolver:
// una sola forma —un ECLI, un ROJ o un número de resolución—, su valor como se
// escribió y, si la lleva, la fecha de la resolución. Es inmutable y solo se
// construye validando, con NuevaReferencia. El valor cero no es ninguna
// referencia.
type Referencia struct {
	// forma es la que se dio: decide el campo del formulario.
	forma forma
	// valor es el ECLI, el ROJ o el número de resolución, sin recortar ni pasar
	// a mayúsculas.
	valor string
	// fecha es la de la resolución, AAAA-MM-DD. Vacía si no se dio: siempre con
	// un ECLI, y con un ROJ que no la lleva.
	fecha string
	// constitucional dice si es un ECLI cuyo órgano es el Tribunal
	// Constitucional.
	constitucional bool
}

// NuevaReferencia construye la referencia de una resolución con lo que recibe
// cita resolver, donde un argumento vacío es un argumento que no se dio. Valida
// todo, sin abrir ni pedir nada, en este orden (contrato cita-resolver §1;
// FR-002 a FR-005):
//
//  1. hay exactamente una forma de referencia: ni ninguna, ni dos;
//  2. un ECLI tiene su forma y es español, y no lleva fecha, tampoco si es del
//     Tribunal Constitucional;
//  3. un ROJ tiene su forma, y su fecha, si la lleva, la suya;
//  4. un número de resolución tiene su forma y lleva su fecha, con la suya.
//
// Lo primero que falla es un error de clase «argumentos» que dice qué falla.
// Nada se recorta ni se pasa a mayúsculas, y ni el órgano ni las siglas se
// buscan en ninguna lista. El ECLI y el ROJ los reconoce internal/core/ids.
func NuevaReferencia(ecli, roj, resolucion, fecha string) (Referencia, error) {
	dadas := formasDadas(ecli, roj, resolucion)
	if len(dadas) == 0 {
		return Referencia{}, referenciaInvalida("falta la referencia de la resolución: %s", formasAdmitidas)
	}

	if len(dadas) > 1 {
		return Referencia{}, referenciaInvalida("la referencia lleva %s, y tiene que ser una sola: %s",
			enumerar(dadas), formasAdmitidas)
	}

	switch {
	case ecli != "":
		return referenciaPorECLI(ecli, fecha)
	case roj != "":
		return referenciaPorROJ(roj, fecha)
	default:
		return referenciaPorResolucion(resolucion, fecha)
	}
}

// DelTribunalConstitucional dice si la referencia es un ECLI cuyo órgano es el
// Tribunal Constitucional, que esta fuente no cubre. Solo lo dice del ECLI: una
// sentencia suya pedida por su ROJ o por su número es una referencia como
// cualquier otra (FR-015).
func (r Referencia) DelTribunalConstitucional() bool {
	return r.constitucional
}

// formasDadas nombra, en el orden de los argumentos, las formas de referencia
// que se han dado.
func formasDadas(ecli, roj, resolucion string) []string {
	var dadas []string

	if ecli != "" {
		dadas = append(dadas, "un ECLI")
	}

	if roj != "" {
		dadas = append(dadas, "un ROJ")
	}

	if resolucion != "" {
		dadas = append(dadas, "un número de resolución")
	}

	return dadas
}

// enumerar escribe dos o más nombres como una enumeración: «un ECLI y un ROJ»,
// «un ECLI, un ROJ y un número de resolución».
func enumerar(nombres []string) string {
	ultimo := len(nombres) - 1

	return strings.Join(nombres[:ultimo], ", ") + " y " + nombres[ultimo]
}

// referenciaPorECLI es la de un ECLI: con su forma, que comprueba
// ids.AnalizarECLI, y sin fecha.
func referenciaPorECLI(ecli, fecha string) (Referencia, error) {
	analizado, err := ids.AnalizarECLI(ecli)
	if err != nil {
		return Referencia{}, err
	}

	if fecha != "" {
		return Referencia{}, referenciaInvalida("un ECLI no lleva fecha (--fecha): identifica él solo la resolución")
	}

	return Referencia{
		forma:          formaECLI,
		valor:          ecli,
		constitucional: analizado.Organo() == organoDelTribunalConstitucional,
	}, nil
}

// referenciaPorROJ es la de un ROJ: con su forma, que comprueba ids.AnalizarROJ,
// y con su fecha si la lleva.
func referenciaPorROJ(roj, fecha string) (Referencia, error) {
	if _, err := ids.AnalizarROJ(roj); err != nil {
		return Referencia{}, err
	}

	if fecha == "" {
		return Referencia{forma: formaROJ, valor: roj}, nil
	}

	if err := validarFecha(fecha); err != nil {
		return Referencia{}, err
	}

	return Referencia{forma: formaROJ, valor: roj, fecha: fecha}, nil
}

// referenciaPorResolucion es la de un número de resolución: con su forma y con
// su fecha, que es obligatoria.
func referenciaPorResolucion(resolucion, fecha string) (Referencia, error) {
	if !formaDelNumeroDeResolucion.MatchString(resolucion) {
		return Referencia{}, referenciaInvalida(
			"el número de resolución %q no es válido: la forma es <número>/<año>, con el número en cifras y el año "+
				"de cuatro cifras, como 1088/2023", resolucion)
	}

	if fecha == "" {
		return Referencia{}, referenciaInvalida(
			"el número de resolución %s necesita su fecha (--fecha AAAA-MM-DD): cada órgano numera las suyas, y "+
				"solo con la fecha identifica una resolución", resolucion)
	}

	if err := validarFecha(fecha); err != nil {
		return Referencia{}, err
	}

	return Referencia{forma: formaResolucion, valor: resolucion, fecha: fecha}, nil
}

// validarFecha exige que la fecha se escriba AAAA-MM-DD y que sea un día que
// existe.
func validarFecha(fecha string) error {
	if !formaDeLaFecha.MatchString(fecha) {
		return referenciaInvalida("la fecha %q no es válida: la forma es AAAA-MM-DD, como 2023-07-04", fecha)
	}

	if _, err := time.Parse(time.DateOnly, fecha); err != nil {
		return referenciaInvalida("la fecha %q no es válida: no es un día que existe", fecha)
	}

	return nil
}

// errorDeReferencia es el rechazo de una referencia mal dada. No se exporta:
// quien lo recibe lo reconoce por su clase, con errors.As a schema.ConClase,
// que es como lo reconoce el kernel.
type errorDeReferencia struct {
	// motivo es qué falla, en español y para la persona.
	motivo string
}

// Un rechazo de la referencia declara su clase él mismo.
var _ schema.ConClase = (*errorDeReferencia)(nil)

// referenciaInvalida construye el rechazo de una referencia con lo que falla.
func referenciaInvalida(formato string, argumentos ...any) error {
	return &errorDeReferencia{motivo: fmt.Sprintf(formato, argumentos...)}
}

// Error dice qué falla de la referencia.
func (e *errorDeReferencia) Error() string {
	return e.motivo
}

// Clase es siempre «argumentos»: una referencia mal dada es un fallo de quien
// la escribe, que el kernel traduce a código 2 y que no llega a pedir nada.
func (*errorDeReferencia) Clase() schema.Clase {
	return schema.ClaseArgumentos
}
