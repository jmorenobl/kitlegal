package boe_test

import (
	"errors"
	"fmt"
	"go/doc/comment"
	"go/parser"
	"go/token"
	"maps"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seccionDelPorte es el encabezado de la sección del comentario del paquete que
// anota el porte de refs/boe.py (contrato esquemas-fixtures-y-controles §9).
const seccionDelPorte = "Comportamientos no obvios de refs/boe.py"

// destinoDelPorte es lo que el porte hace con un comportamiento de boe.py.
type destinoDelPorte string

const (
	sePorta   destinoDelPorte = "se porta"
	seAdapta  destinoDelPorte = "se adapta"
	noSePorta destinoDelPorte = "no se porta"
)

// anotacion es lo que una entrada de la sección fija y FR-120 exige: las líneas
// de refs/boe.py, el destino y el requisito. El comportamiento y el motivo son
// prosa de doc.go; leerAnotacion exige que estén, pero no los compara.
type anotacion struct {
	lineas    string
	destino   destinoDelPorte
	requisito string
}

// entradasDeFR120 son las 32 entradas que FR-120 exige como mínimo, por número,
// con las líneas escritas como en doc.go: la 18 («282-283 frente a la 302» en
// FR-120) queda como «282-283 y 302», y la 21 y la 26 reúnen en una sola lista
// las líneas que FR-120 reparte entre sus partes.
var entradasDeFR120 = map[int]anotacion{
	1:  {"37-46", sePorta, "FR-091"},
	2:  {"411-416 y 433", sePorta, "FR-091"},
	3:  {"56 y 66", sePorta, "FR-096"},
	4:  {"49-66", seAdapta, "FR-090"},
	5:  {"110-118", sePorta, "FR-011"},
	6:  {"132-136", sePorta, "FR-011"},
	7:  {"99-100 y 139-150", noSePorta, "FR-014"},
	8:  {"102-104", noSePorta, "FR-014"},
	9:  {"104, 150, 289-296, 387, 420, 473-482, 533, 559-561 y 579-581", seAdapta, "FR-016"},
	10: {"186-193", noSePorta, "FR-013"},
	11: {"186", seAdapta, "FR-090"},
	12: {"405-411", seAdapta, "FR-013"},
	13: {"437-443", seAdapta, "FR-020"},
	14: {"433", sePorta, "FR-021"},
	15: {"440-442", seAdapta, "FR-021"},
	16: {"99-104, 139-150, 186-193, 405-416 y 433", noSePorta, "FR-093"},
	17: {"485-498", seAdapta, "FR-050"},
	18: {"282-283 y 302", seAdapta, "FR-032"},
	19: {"362-363, 459-460 y 518-519", sePorta, "FR-041, FR-051, FR-061 y FR-093"},
	20: {"218-225", seAdapta, "FR-070"},
	21: {"195, 463, 522, 531, 540, 550-554 y 570-574", sePorta, "FR-070"},
	22: {"365-369", sePorta, "FR-040"},
	23: {"107", sePorta, "FR-010"},
	24: {"155-176 y 389", sePorta, "FR-040"},
	25: {"158-175 y 375", sePorta, "FR-040 y Fuera de alcance"},
	26: {"252, 258-271 y 784", sePorta, "FR-030"},
	27: {"264-271 y 783", seAdapta, "FR-030"},
	28: {"186, 355, 406, 452 y 511", seAdapta, "FR-080 y FR-081"},
	29: {"564", seAdapta, "FR-060"},
	30: {"77-86, 276-280, 360, 363, 457, 460, 516, 519, 777-779 y 804-806", noSePorta, "FR-100"},
	31: {"32 y 81", noSePorta, "FR-122"},
	32: {"306-346, 588-689 y 692-732", noSePorta, "Fuera de alcance"},
}

// TestDocAnotaElPorte es el control de SC-010: el comentario del paquete, leído
// con el analizador sintáctico de la biblioteca estándar y estructurado como lo
// presenta go doc, anota una sola vez cada una de las 32 entradas de FR-120 con
// las líneas de refs/boe.py, el destino y el requisito que FR-120 le fija.
// Quitar una entrada, repetirla, partirla en varias líneas o cambiarle las
// líneas, el destino o el requisito lo hace fallar (plan.md, control 13).
func TestDocAnotaElPorte(t *testing.T) {
	t.Parallel()

	anotaciones := anotacionesDelPorte(t)

	for _, numero := range slices.Sorted(maps.Keys(entradasDeFR120)) {
		esperada := entradasDeFR120[numero]

		t.Run(fmt.Sprintf("entrada-%02d", numero), func(t *testing.T) {
			t.Parallel()

			anotada, hay := anotaciones[numero]
			require.True(t, hay, "doc.go no anota la entrada [%d] de FR-120 en la sección «%s»",
				numero, seccionDelPorte)

			assert.Equal(t, esperada, anotada,
				"la entrada [%d] de doc.go lleva las líneas de refs/boe.py, el destino y el requisito que le fija FR-120",
				numero)
		})
	}
}

// anotacionesDelPorte lee doc.go con go/parser, estructura su comentario de
// paquete con go/doc/comment y devuelve por número las entradas de la sección
// del porte. Toda entrada de la sección tiene que tener la forma del contrato y
// aparecer una sola vez; las que pasen de 32 se admiten, porque FR-120 fija un
// mínimo, pero con la misma forma.
func anotacionesDelPorte(t *testing.T) map[int]anotacion {
	t.Helper()

	fichero, err := parser.ParseFile(token.NewFileSet(), "doc.go", nil, parser.ParseComments|parser.PackageClauseOnly)
	require.NoError(t, err, "el test lee el doc.go del propio paquete")
	require.NotNil(t, fichero.Doc, "doc.go lleva el comentario del paquete")

	textos := textosDeLaSeccion(t, new(comment.Parser).Parse(fichero.Doc.Text()))
	require.NotEmpty(t, textos, "la sección «%s» anota el porte en una lista", seccionDelPorte)

	anotaciones := make(map[int]anotacion, len(textos))

	for _, texto := range textos {
		numero, anotada, errDeForma := leerAnotacion(texto)
		if errDeForma != nil {
			t.Errorf("%v: «%s»", errDeForma, texto)

			continue
		}

		if _, repetida := anotaciones[numero]; repetida {
			t.Errorf("la entrada [%d] aparece más de una vez en la sección «%s»", numero, seccionDelPorte)

			continue
		}

		anotaciones[numero] = anotada
	}

	return anotaciones
}

// textosDeLaSeccion devuelve el texto de cada elemento de lista que cuelga del
// encabezado del porte, hasta el encabezado siguiente. El encabezado tiene que
// aparecer una sola vez.
func textosDeLaSeccion(t *testing.T, documento *comment.Doc) []string {
	t.Helper()

	var (
		encabezados int
		dentro      bool
		textos      []string
	)

	for _, bloque := range documento.Content {
		switch tipo := bloque.(type) {
		case *comment.Heading:
			dentro = textoPlano(tipo.Text) == seccionDelPorte
			if dentro {
				encabezados++
			}
		case *comment.List:
			if dentro {
				textos = append(textos, textosDeLaLista(t, tipo)...)
			}
		}
	}

	require.Equal(t, 1, encabezados, "el comentario del paquete lleva una sola vez el encabezado «# %s»",
		seccionDelPorte)

	return textos
}

// textosDeLaLista devuelve el texto de cada elemento de una lista de la sección.
// Cada elemento tiene que ser un único párrafo de una sola línea: es la forma
// «una por línea de lista» del contrato, que deja cada entrada entera a la vista
// de quien busca su número.
func textosDeLaLista(t *testing.T, lista *comment.List) []string {
	t.Helper()

	textos := make([]string, 0, len(lista.Items))

	for _, elemento := range lista.Items {
		if len(elemento.Content) != 1 {
			t.Errorf("un elemento de la sección «%s» tiene %d bloques y una entrada es un solo párrafo",
				seccionDelPorte, len(elemento.Content))

			continue
		}

		parrafo, esParrafo := elemento.Content[0].(*comment.Paragraph)
		if !esParrafo {
			t.Errorf("un elemento de la sección «%s» no es un párrafo", seccionDelPorte)

			continue
		}

		texto := textoPlano(parrafo.Text)
		if strings.Contains(texto, "\n") {
			t.Errorf("la entrada ocupa más de una línea y el contrato pide una por línea de lista: «%s»", texto)

			continue
		}

		textos = append(textos, texto)
	}

	return textos
}

// textoPlano junta el texto de un párrafo o de un encabezado tal como está
// escrito, con los enlaces reducidos a su texto.
func textoPlano(fragmentos []comment.Text) string {
	var texto strings.Builder

	for _, fragmento := range fragmentos {
		switch tipo := fragmento.(type) {
		case comment.Plain:
			texto.WriteString(string(tipo))
		case comment.Italic:
			texto.WriteString(string(tipo))
		case *comment.Link:
			texto.WriteString(textoPlano(tipo.Text))
		case *comment.DocLink:
			texto.WriteString(textoPlano(tipo.Text))
		}
	}

	return texto.String()
}

var (
	// formaDeLaCabecera es el primer campo, «[N] líneas <rango>»: el número de
	// la entrada y las líneas de refs/boe.py, sueltas o en rangos, separadas por
	// «, » y la última por « y ».
	formaDeLaCabecera = regexp.MustCompile(
		`^\[([1-9][0-9]*)\] (líneas?) ([1-9][0-9]*(?:-[1-9][0-9]*)?(?:(?:, | y )[1-9][0-9]*(?:-[1-9][0-9]*)?)*)$`)

	// formaDelDestino es el tercer campo: «se porta», que puede precisar qué se
	// porta tras dos puntos, o «se adapta» y «no se porta», que llevan su motivo
	// tras dos puntos. Cada forma admite el plural cuando el comportamiento lo es.
	formaDelDestino = regexp.MustCompile(`^(no se portan?|se adaptan?|se portan?)(?:: (.+))?$`)

	// formaDelRequisito es el cuarto campo: uno o varios requisitos del spec, o
	// «Fuera de alcance», separados por «, » y el último por « y ».
	formaDelRequisito = regexp.MustCompile(
		`^(?:FR-[0-9]{3}|Fuera de alcance)(?:(?:, | y )(?:FR-[0-9]{3}|Fuera de alcance))*$`)
)

// leerAnotacion separa una entrada en los cuatro campos del contrato,
// «[N] líneas <rango> · <comportamiento> · <destino> · <requisito>», y
// comprueba la forma de cada uno. El error nombra el campo que no la tiene.
func leerAnotacion(texto string) (int, anotacion, error) {
	campos := strings.Split(texto, " · ")
	if len(campos) != 4 {
		return 0, anotacion{}, fmt.Errorf("la entrada tiene %d campos separados por « · » y el contrato pide 4: "+
			"número y líneas, comportamiento, destino y requisito", len(campos))
	}

	cabecera := formaDeLaCabecera.FindStringSubmatch(campos[0])
	if cabecera == nil {
		return 0, anotacion{}, errors.New("el primer campo no es «[N] líneas <rango>»")
	}

	numero, err := strconv.Atoi(cabecera[1])
	if err != nil {
		return 0, anotacion{}, fmt.Errorf("el número de la entrada no se interpreta: %w", err)
	}

	if unaSola := !strings.ContainsAny(cabecera[3], "-, "); unaSola != (cabecera[2] == "línea") {
		return 0, anotacion{}, errors.New("el primer campo dice «línea» ante una sola línea y «líneas» ante un rango o varias")
	}

	if strings.TrimSpace(campos[1]) == "" {
		return 0, anotacion{}, errors.New("el segundo campo, el comportamiento de boe.py, está vacío")
	}

	destino := formaDelDestino.FindStringSubmatch(campos[2])
	if destino == nil {
		return 0, anotacion{}, errors.New("el tercer campo no es «se porta», «se adapta: <motivo>» ni «no se porta: <motivo>»")
	}

	tipo := destinoDelPorte(strings.TrimSuffix(destino[1], "n"))
	if tipo != sePorta && destino[2] == "" {
		return 0, anotacion{}, fmt.Errorf("el destino «%s» lleva su motivo tras dos puntos", tipo)
	}

	if !formaDelRequisito.MatchString(campos[3]) {
		return 0, anotacion{}, errors.New("el cuarto campo no es una lista de requisitos FR-xxx o «Fuera de alcance»")
	}

	return numero, anotacion{lineas: cabecera[3], destino: tipo, requisito: campos[3]}, nil
}
