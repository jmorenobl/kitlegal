package cita

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/jmorenobl/kitlegal/internal/core/ids"
)

// Las etiquetas de los ocho datos de la ficha, escritas como en el documento
// (H23, FR-021). El ROJ y el ECLI van los dos en la línea «Roj:»: etiquetaECLI
// no encabeza ninguna línea, y es el nombre con el que un error nombra ese
// dato.
const (
	etiquetaROJ        = "Roj"
	etiquetaECLI       = "ECLI"
	etiquetaOrgano     = "Órgano"
	etiquetaFecha      = "Fecha"
	etiquetaRecurso    = "Nº de Recurso"
	etiquetaResolucion = "Nº de Resolución"
	etiquetaPonente    = "Ponente"
	etiquetaTipo       = "Tipo de Resolución"
)

// Lo que separa las partes de la ficha.
const (
	// finDeLinea parte el texto en líneas; el \r de un final \r\n se va con
	// los blancos de los extremos de cada una.
	finDeLinea = "\n"
	// finDeEtiqueta va pegado a la etiqueta, entera, y delante del valor.
	finDeEtiqueta = ":"
	// separadorDelROJ va entre el ROJ y el ECLI en la línea «Roj:».
	separadorDelROJ = " - "
)

// Ficha son los ocho datos con los que el CENDOJ encabeza un documento, con
// las claves de contracts/applet-cita.md §4: como en el documento, sin los
// blancos de alrededor, salvo la fecha, que va AAAA-MM-DD. Del documento no
// lleva nada más: ni su texto, ni su fallo, ni las demás líneas de la ficha
// (H23, FR-022). Solo se construye con LeerFicha, que la comprueba; su valor
// cero no es ninguna ficha.
type Ficha struct {
	// ROJ es el del documento, con la forma de un ROJ.
	ROJ string `json:"roj" jsonschema:"minLength=1"`
	// ECLI es el del documento, con la forma de un ECLI español.
	ECLI string `json:"ecli" jsonschema:"minLength=1"`
	// Organo es el que dictó la resolución.
	Organo string `json:"organo" jsonschema:"minLength=1"`
	// Fecha es la de la resolución, AAAA-MM-DD: un día que existe.
	Fecha string `json:"fecha" jsonschema:"pattern=^[0-9]{4}-[0-9]{2}-[0-9]{2}$"`
	// Recurso es el número de recurso.
	Recurso string `json:"recurso" jsonschema:"minLength=1"`
	// Resolucion es el número de resolución, <número>/<año>.
	Resolucion string `json:"resolucion" jsonschema:"pattern=^[0-9]+/[0-9]{4}$"`
	// Ponente es quien redactó la resolución.
	Ponente string `json:"ponente" jsonschema:"minLength=1"`
	// Tipo es el de la resolución: sentencia, auto…
	Tipo string `json:"tipo" jsonschema:"minLength=1"`

	// roj y ecli son los dos identificadores analizados: de ellos sale la
	// correspondencia y el número que se cruza con el de resolución.
	roj  ids.ROJ
	ecli ids.ECLI
}

// LeerFicha lee la primera ficha del texto de un documento del CENDOJ: la
// línea «Roj: <ROJ> - <ECLI>» y, a partir de ella, una línea «<etiqueta>:
// <valor>» por cada uno de los otros seis datos, de la primera que lleva su
// etiqueta entera. Admite los finales de línea \n y \r\n y blancos en los
// extremos de cada línea; lo que hay antes de la línea «Roj:» no se lee, y las
// líneas sin una de esas etiquetas —las otras de la ficha y el texto de la
// sentencia— no dan ningún dato.
//
// La ficha es reconocible si están los ocho datos y los cuatro que se comparan
// tienen su forma: el ROJ, el ECLI español, la fecha dd/mm/aaaa de un día que
// existe y el número de resolución. Si no, devuelve un error de clase
// «argumentos» que dice que el texto no lleva ficha o que nombra, con su
// etiqueta, el primer dato que falta o que no tiene su forma, en el orden de
// la ficha (H23, FR-021, FR-026).
func LeerFicha(texto string) (Ficha, error) {
	lineas := lineasDesdeLaFicha(texto)
	if len(lineas) == 0 {
		return Ficha{}, argumentosInvalidos("el texto no lleva la ficha de un documento del CENDOJ: " +
			"falta la línea «Roj:», la primera de su encabezamiento")
	}

	ficha, err := identificadoresDe(lineas[0])
	if err != nil {
		return Ficha{}, err
	}

	resto := lineas[1:]

	if ficha.Organo, err = datoDe(resto, etiquetaOrgano); err != nil {
		return Ficha{}, err
	}

	if ficha.Fecha, err = fechaDe(resto); err != nil {
		return Ficha{}, err
	}

	if ficha.Recurso, err = datoDe(resto, etiquetaRecurso); err != nil {
		return Ficha{}, err
	}

	if ficha.Resolucion, err = resolucionDe(resto); err != nil {
		return Ficha{}, err
	}

	if ficha.Ponente, err = datoDe(resto, etiquetaPonente); err != nil {
		return Ficha{}, err
	}

	if ficha.Tipo, err = datoDe(resto, etiquetaTipo); err != nil {
		return Ficha{}, err
	}

	return ficha, nil
}

// lineasDesdeLaFicha son las líneas del texto desde la primera que empieza por
// «Roj:», cada una sin los blancos de sus extremos, o ninguna si el texto no
// lleva esa línea.
func lineasDesdeLaFicha(texto string) []string {
	lineas := strings.Split(texto, finDeLinea)
	for indice, linea := range lineas {
		lineas[indice] = strings.TrimSpace(linea)
	}

	primera := slices.IndexFunc(lineas, func(linea string) bool {
		return strings.HasPrefix(linea, etiquetaROJ+finDeEtiqueta)
	})
	if primera < 0 {
		return nil
	}

	return lineas[primera:]
}

// identificadoresDe lee de la línea «Roj:» el ROJ y el ECLI, que van separados
// por « - », y los comprueba con su forma. El corte se hace antes de quitar
// los blancos de cada parte, para que una línea sin su ROJ lo diga de él y no
// del ECLI.
func identificadoresDe(lineaDelROJ string) (Ficha, error) {
	valor := strings.TrimPrefix(lineaDelROJ, etiquetaROJ+finDeEtiqueta)
	antes, despues, _ := strings.Cut(valor, separadorDelROJ)
	delROJ, delECLI := strings.TrimSpace(antes), strings.TrimSpace(despues)

	if delROJ == "" {
		return Ficha{}, faltaElDato(etiquetaROJ)
	}

	if delECLI == "" {
		return Ficha{}, argumentosInvalidos("la ficha no es reconocible: falta el dato «%s»: la línea «Roj:» lleva %q "+
			"y su forma es «Roj: <ROJ> - <ECLI>»", etiquetaECLI, strings.TrimSpace(valor))
	}

	roj, err := ids.AnalizarROJ(delROJ)
	if err != nil {
		return Ficha{}, sinSuFormaDeIdentificador(etiquetaROJ, err)
	}

	ecli, err := ids.AnalizarECLI(delECLI)
	if err != nil {
		return Ficha{}, sinSuFormaDeIdentificador(etiquetaECLI, err)
	}

	return Ficha{ROJ: delROJ, ECLI: delECLI, roj: roj, ecli: ecli}, nil
}

// datoDe lee un dato de texto: el valor de la primera línea que lleva su
// etiqueta entera, seguida de los dos puntos, sin los blancos de alrededor. No
// puede faltar ni ir vacío. Como la etiqueta va entera, «Nº de Recurso» y
// «Nº de Resolución» no se confunden.
func datoDe(lineas []string, etiqueta string) (string, error) {
	for _, linea := range lineas {
		valor, laLleva := strings.CutPrefix(linea, etiqueta+finDeEtiqueta)
		if !laLleva {
			continue
		}

		if valor = strings.TrimSpace(valor); valor != "" {
			return valor, nil
		}

		break
	}

	return "", faltaElDato(etiqueta)
}

// fechaDe lee la fecha de la resolución, que en el documento va dd/mm/aaaa y
// tiene que ser un día que existe, y la devuelve AAAA-MM-DD (research.md D8).
func fechaDe(lineas []string) (string, error) {
	valor, err := datoDe(lineas, etiquetaFecha)
	if err != nil {
		return "", err
	}

	dia, err := time.Parse(formaDeLaFechaDelCENDOJ, valor)
	if err != nil {
		return "", sinSuForma(etiquetaFecha, valor, "dd/mm/aaaa, un día que existe")
	}

	return dia.Format(formaDeLaFecha), nil
}

// resolucionDe lee el número de resolución, <número>/<año>.
func resolucionDe(lineas []string) (string, error) {
	valor, err := datoDe(lineas, etiquetaResolucion)
	if err != nil {
		return "", err
	}

	if !esNumeroConAnio(valor) {
		return "", sinSuForma(etiquetaResolucion, valor, "<número>/<año>, en cifras y con el año de cuatro")
	}

	return valor, nil
}

// faltaElDato es la ficha a la que le falta uno de sus ocho datos: no está su
// línea, o está sin valor.
func faltaElDato(etiqueta string) error {
	return argumentosInvalidos("la ficha no es reconocible: falta el dato «%s»", etiqueta)
}

// sinSuForma es la ficha con uno de los datos que se comparan mal escrito: lo
// nombra con su etiqueta, dice lo que el documento lleva y cuál es su forma.
func sinSuForma(etiqueta, valor, forma string) error {
	return argumentosInvalidos("la ficha no es reconocible: el dato «%s» lleva %q y su forma es %s",
		etiqueta, valor, forma)
}

// sinSuFormaDeIdentificador es lo mismo para el ROJ y el ECLI, cuya forma
// explica el rechazo de su analizador. Lo envuelve con %w: la clase del error
// sigue siendo la suya, «argumentos».
func sinSuFormaDeIdentificador(etiqueta string, rechazo error) error {
	return fmt.Errorf("la ficha no es reconocible: el dato «%s» no tiene su forma: %w", etiqueta, rechazo)
}
