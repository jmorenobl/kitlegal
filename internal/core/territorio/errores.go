package territorio

import (
	"fmt"
	"strings"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// errorDeConsulta es la respuesta negativa a una consulta que el registro
// decide: una entrada que no nombra nada, un código o un nombre que no está en
// la relación o un nombre que lleva a varios municipios. No se exporta: quien
// lo recibe lo reconoce por su clase, con errors.As a schema.ConClase, que es
// como lo reconoce el kernel. Los rechazos que decide la gramática del código
// INE los da internal/core/ids con su propio tipo, de la misma clase.
type errorDeConsulta struct {
	// clase es schema.ClaseArgumentos o schema.ClaseNoEncontrado: una
	// consulta no puede fallar de ninguna otra forma (FR-016).
	clase schema.Clase
	// mensaje nombra la entrada con %q y dice qué le pasa.
	mensaje string
}

// Un fallo de consulta declara su clase él mismo.
var _ schema.ConClase = (*errorDeConsulta)(nil)

// Error devuelve el mensaje, que llega literal al sobre de fallo y a la salida
// de error (research.md V17).
func (e *errorDeConsulta) Error() string {
	return e.mensaje
}

// Clase es la del fallo, que el kernel traduce a código 2 o 3.
func (e *errorDeConsulta) Clase() schema.Clase {
	return e.clase
}

// consultaVacia es la entrada sin ninguna letra ni ninguna cifra una vez
// plegada: vacía o hecha solo de espacios y separadores, que no nombra nada.
func consultaVacia(entrada string) error {
	return &errorDeConsulta{
		clase: schema.ClaseArgumentos,
		mensaje: fmt.Sprintf("la consulta %q no nombra ningún municipio: no tiene ninguna letra ni ninguna cifra",
			entrada),
	}
}

// codigoFueraDeLaRelacion es el código INE bien formado —provincia de 01 a 52
// y municipio de 001 a 999— que no es el de ningún municipio de la relación
// (FR-011).
func codigoFueraDeLaRelacion(entrada string) error {
	return &errorDeConsulta{
		clase:   schema.ClaseNoEncontrado,
		mensaje: fmt.Sprintf("ningún municipio de la relación tiene el código INE %q", entrada),
	}
}

// nombreFueraDeLaRelacion es el nombre que, plegado, no es ninguna forma
// conocida de ningún municipio de la relación (FR-011).
func nombreFueraDeLaRelacion(entrada string) error {
	return &errorDeConsulta{
		clase:   schema.ClaseNoEncontrado,
		mensaje: fmt.Sprintf("ningún municipio de la relación se llama %q", entrada),
	}
}

// nombreAmbiguo es el nombre que lleva a más de un municipio. El sobre de
// fallo no tiene dónde poner una lista, así que los candidatos viajan en el
// mensaje, todos, en el orden en que llegan —el de su código INE— y en la
// forma fija «<código INE> <nombre> (<provincia>)», separados por «; »
// (FR-013, FR-014, research.md D14).
func nombreAmbiguo(entrada string, candidatos []*municipioRegistrado) error {
	formas := make([]string, 0, len(candidatos))
	for _, candidato := range candidatos {
		formas = append(formas, fmt.Sprintf("%s %s (%s)", candidato.codigo, candidato.nombre, candidato.provincia.nombre))
	}

	return &errorDeConsulta{
		clase: schema.ClaseArgumentos,
		mensaje: fmt.Sprintf("el nombre %q es el de %d municipios de la relación; consulta uno por su código INE: %s",
			entrada, len(candidatos), strings.Join(formas, "; ")),
	}
}

// defectoDeCarga es un defecto de las fuentes: un fichero que no se lee, un
// dato que falta o una incoherencia entre ficheros. No declara ninguna clase a
// propósito —el kernel lo trata como inesperado, código 1—: unas fuentes que no
// cargan son un defecto de composición y no algo que quien pregunta pueda
// corregir (contrato del applet §7). Por eso guarda el motivo como texto y no
// envuelve el error que lo causó: un código mal formado en un fichero lo
// rechaza internal/core/ids con clase «argumentos», y envolverlo con %w
// convertiría el defecto de los datos en un código 2.
type defectoDeCarga struct {
	// fichero es la ruta del fichero, relativa a la raíz del repositorio.
	fichero string
	// motivo dice qué falla y nombra el dato.
	motivo string
}

// Error nombra el fichero y el defecto, como «data/territorio/dir3.yaml: el
// municipio PPMMM no está en data/territorio/municipios.yaml».
func (d *defectoDeCarga) Error() string {
	return d.fichero + ": " + d.motivo
}
