package evals

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/jmorenobl/kitlegal/internal/skills"
	"github.com/jmorenobl/kitlegal/internal/source/boe"
)

// ficheroDeExpresionesProhibidas es el nombre exacto de la lista de expresiones
// prohibidas en la carpeta de evals de una skill: LeerConjunto la reconoce por él
// y no la lee como eval (contrato lista-y-juicio §1; research D1).
const ficheroDeExpresionesProhibidas = "expresiones-prohibidas.yaml"

// rutaDelEsquemaDeExpresionesProhibidas es el esquema publicado de la lista,
// relativo al directorio de este paquete, como el de eval.
const rutaDelEsquemaDeExpresionesProhibidas = "../../schemas/expresiones-prohibidas.yaml.json"

// Los extremos de la expresión regular de una expresión prohibida: delante y
// detrás de sus palabras, el principio o el final del texto o algo que no es
// letra ni cifra (research D3).
const (
	antesDeLaExpresion   = `(?:^|[^\p{L}\p{N}])`
	despuesDeLaExpresion = `(?:$|[^\p{L}\p{N}])`
)

// Lo que se escribe en una forma fija de la lista y las piezas de su expresión
// regular (contracts/lista-de-expresiones.md §3 de H7.4; research D7 de H7.4).
const (
	// marcaDeFormaFija es la marca con la que empieza una forma fija que lleva
	// la etiqueta de un aviso o de un hallazgo.
	marcaDeFormaFija = "⚠"

	// marcadorDeCita y marcadorDeFecha son donde la respuesta lleva la cita de
	// un bloque y una fecha de vigencia.
	marcadorDeCita  = "<cita>"
	marcadorDeFecha = "<fecha>"

	// blancoDeFormaFija es cada blanco de la forma: uno o más espacios o
	// tabuladores.
	blancoDeFormaFija = `[ \t]+`

	// fechaDeFormaFija es una fecha de vigencia: ocho cifras.
	fechaDeFormaFija = `[0-9]{8}`

	// citaDeFormaFija es la cita de un bloque: lo que hay tras el texto anterior
	// hasta el corchete de cierre de una cita [<identificador>, bloque <id>], en
	// la misma línea.
	citaDeFormaFija = `[^\[\n]*\[[^\]\n]*,[ \t]*bloque[ \t]+[^\]\n]+\]`

	// saltoDeFormaFija es lo que deja en el texto cada tramo quitado: un salto de
	// línea, a través del cual no casa ninguna expresión.
	saltoDeFormaFija = "\n"
)

// piezaDeFormaFija es cada pieza de una forma fija, en su orden: un blanco, uno
// de los dos marcadores, un tramo de texto sin blancos ni marcadores o un «<» que
// no abre un marcador.
var piezaDeFormaFija = regexp.MustCompile(`[ \t]+|<cita>|<fecha>|[^ \t<]+|<`)

// ExpresionesProhibidas es la lista de expresiones prohibidas de una skill, las
// que no lleva la respuesta de una eval que la activa, en sus cuatro familias,
// con las formas fijas que se quitan de la respuesta antes de buscarlas (FR-050
// de H7.2, FR-020 de H7.3 y FR-030 y FR-031 de H7.4; data-model §1 de H7.4). La
// maquinaria, lo dicho en otra conversación y el anuncio son de la clase A; la
// redacción no leída, de la clase B. Se lee de
// evals/<skill>/expresiones-prohibidas.yaml, validada contra
// schemas/expresiones-prohibidas.yaml.json, que exige las cinco claves; su valor
// cero es el de una skill sin lista. Desde H24 no juzga ninguna respuesta: es
// el vocabulario que la prosa de la skill no usa (FR-070 y FR-071 de H24).
type ExpresionesProhibidas struct {
	// Maquinaria son las de la maquinaria interna —la memoria de consultas,
	// kitlegal graph y sus verbos, los códigos de salida, los hallazgos y sus
	// clases, el JSON y el sobre—, en el orden del fichero.
	Maquinaria []string `yaml:"maquinaria"`

	// OtraConversacion son las que atribuyen a la skill algo dicho a quien
	// pregunta en otra conversación, en el orden del fichero.
	OtraConversacion []string `yaml:"otra_conversacion"`

	// Anuncio son las que anuncian a quien pregunta la respuesta que viene o el
	// estado de lo comprobado —que no hay nada que trasladar, que ya se puede
	// responder, que se tiene lo necesario—, en el orden del fichero. Ninguna
	// dice que la norma no está derogada o que no tiene avisos (FR-020).
	Anuncio []string `yaml:"anuncio"`

	// RedaccionNoLeida son las de la clase B: las que cuentan de una redacción
	// que ninguna orden devolvió qué decía, hasta cuándo rigió o qué cambió
	// respecto de ella, en el orden del fichero (FR-011 y FR-030 de H7.4).
	RedaccionNoLeida []string `yaml:"redaccion_no_leida"`

	// FormasFijas son los textos que enseña la skill y que se quitan de la
	// respuesta antes de buscar las expresiones, en el orden del fichero: cada
	// uno, de una línea, con los marcadores <cita> y <fecha> donde la respuesta
	// lleva una cita y una fecha de vigencia (FR-031 de H7.4;
	// contracts/lista-de-expresiones.md §3 de H7.4).
	FormasFijas []string `yaml:"formas_fijas"`
}

// esquemaDeExpresionesProhibidas compila una sola vez el esquema publicado de la
// lista, que no cambia mientras se ejecutan los tests.
var esquemaDeExpresionesProhibidas = sync.OnceValues(func() (*jsonschema.Schema, error) {
	return compilarEsquemaPublicado(rutaDelEsquemaDeExpresionesProhibidas, "de la lista de expresiones prohibidas")
})

// leerExpresionesProhibidas lee el contenido de expresiones-prohibidas.yaml con
// el lector común de documentos YAML de internal/skills —una clave repetida es un
// defecto con sus dos líneas, nunca la última que gana— y lo valida contra
// expresiones-prohibidas.yaml.json. Todo error empieza por el nombre del fichero
// y va con la lista vacía (FR-055).
func leerExpresionesProhibidas(contenido []byte) (ExpresionesProhibidas, error) {
	esquema, err := esquemaDeExpresionesProhibidas()
	if err != nil {
		return ExpresionesProhibidas{}, fmt.Errorf("%s: %w", ficheroDeExpresionesProhibidas, err)
	}

	lista, err := skills.ValidarDocumentoYAML[ExpresionesProhibidas](contenido, esquema)
	if err != nil {
		return ExpresionesProhibidas{}, fmt.Errorf("%s: %w", ficheroDeExpresionesProhibidas, err)
	}

	return lista, nil
}

// ExtraerExpresionesProhibidas devuelve las expresiones de la lista que lleva el
// texto, en el orden de la lista —la maquinaria, lo dicho en otra conversación,
// el anuncio y la redacción no leída— y sin repetir, o nil si no lleva ninguna.
// Antes de buscar quita del texto las formas fijas de la lista, con
// sinFormasFijas. Una expresión se encuentra si su forma casa en algún punto de
// lo que queda: sus palabras en su orden, cada una sin distinguir mayúsculas,
// con los blancos y el énfasis de Markdown entre dos que tolera la forma fija de
// los avisos (H5.1), y sin letra ni cifra a los lados. No pliega tildes ni
// admite un salto de línea entre dos palabras (FR-051 de H7.2, FR-024 de H7.3 y
// FR-030 y FR-031 de H7.4; contratos lista-y-juicio §3 de H7.2 y
// lista-de-expresiones §3 de H7.3 y de H7.4; research D3 de H7.2 y D7 de H7.4).
// Una lista sin familias, la de una skill sin lista, no quita ni encuentra nada.
func ExtraerExpresionesProhibidas(texto string, lista ExpresionesProhibidas) []string {
	expresiones := lista.expresiones()
	if len(expresiones) == 0 {
		return nil
	}

	texto = lista.sinFormasFijas(texto)

	var encontradas []string

	for _, expresion := range expresiones {
		if !slices.Contains(encontradas, expresion) && formasDeExpresiones.forma(expresion).MatchString(texto) {
			encontradas = append(encontradas, expresion)
		}
	}

	return encontradas
}

// expresiones son las de las cuatro familias de la lista, en su orden —la
// maquinaria, lo dicho en otra conversación, el anuncio y la redacción no
// leída— y con sus repeticiones; ninguna si la lista no tiene familias, como la
// de una skill sin lista.
func (l ExpresionesProhibidas) expresiones() []string {
	return slices.Concat(l.Maquinaria, l.OtraConversacion, l.Anuncio, l.RedaccionNoLeida)
}

// sinFormasFijas es el texto sin las formas fijas de la skill
// (contracts/lista-de-expresiones.md §3 de H7.4; FR-031 de H7.4): primero cada
// forma de FormasFijas, en su orden y en todas sus apariciones, con la
// expresión de formaDeFormaFija; después la marca, la etiqueta y los dos puntos
// de cada aviso de vigencia, con las expresiones de ExtraerAvisos, y lo que les
// sigue en la línea se queda. Cada tramo quitado se cambia por un salto de
// línea, para que las palabras de sus dos lados no formen una expresión. Una
// lista sin formas fijas no quita nada: se busca en el texto entero, como antes
// de H7.4.
func (l ExpresionesProhibidas) sinFormasFijas(texto string) string {
	if len(l.FormasFijas) == 0 {
		return texto
	}

	for _, forma := range l.FormasFijas {
		texto = formasDeLasFormasFijas.forma(forma).ReplaceAllLiteralString(texto, saltoDeFormaFija)
	}

	avisos := formasDeAviso()
	for _, codigo := range boe.CodigosDeAviso() {
		if aviso, conForma := avisos[codigo]; conForma {
			texto = aviso.ReplaceAllLiteralString(texto, saltoDeFormaFija)
		}
	}

	return texto
}

// formasDeExpresiones son las expresiones regulares ya compiladas de las
// expresiones prohibidas, por expresión, y formasDeLasFormasFijas, las de las
// formas fijas de las listas, por forma: cada una se compila una sola vez, la
// primera vez que se pide, aunque la pidan a la vez tests en paralelo, y no por
// cada respuesta. Solo crecen con las expresiones y las formas distintas que se
// piden, las de listas fijas.
var (
	formasDeExpresiones    = &formasCompiladas{compilar: formaDeExpresion, porTexto: map[string]*regexp.Regexp{}}
	formasDeLasFormasFijas = &formasCompiladas{compilar: formaDeFormaFija, porTexto: map[string]*regexp.Regexp{}}
)

// formasCompiladas guarda, con su cerrojo, la expresión regular ya compilada de
// cada texto, la que da compilar.
type formasCompiladas struct {
	compilar func(texto string) *regexp.Regexp
	cerrojo  sync.Mutex
	porTexto map[string]*regexp.Regexp
}

// forma es la expresión regular del texto: la de compilar, que se compila la
// primera vez que se pide y se guarda para las siguientes.
func (f *formasCompiladas) forma(texto string) *regexp.Regexp {
	f.cerrojo.Lock()
	defer f.cerrojo.Unlock()

	forma, compilada := f.porTexto[texto]
	if !compilada {
		forma = f.compilar(texto)
		f.porTexto[texto] = forma
	}

	return forma
}

// formaDeFormaFija compila la expresión regular de una forma fija de la lista
// (contracts/lista-de-expresiones.md §3 de H7.4): la cabeza de la forma, si
// tiene una (cabezaDeFormaFija), con la expresión de la forma fija de los avisos
// y de los hallazgos (patronDeEtiqueta) —la marca, sus blancos y su énfasis de
// Markdown (H5.1)— y el énfasis que la cierra; y cada pieza de lo que queda en su
// orden: cada blanco, blancoDeFormaFija; cada <fecha>, fechaDeFormaFija; cada
// <cita>, un grupo opcional con citaDeFormaFija y lo que la sigue en la forma
// hasta el blanco siguiente, ese blanco incluido; y el resto, su texto con
// regexp.QuoteMeta. Así la forma quita la línea con la cita y sin ella. Las
// piezas son fijas o pasan por regexp.QuoteMeta, así que la compilación no puede
// fallar.
func formaDeFormaFija(forma string) *regexp.Regexp {
	var patron strings.Builder

	etiqueta, resto, conCabeza := cabezaDeFormaFija(forma)
	if conCabeza {
		patron.WriteString(patronDeEtiqueta(etiqueta) + separadorDeAviso)
	} else {
		resto = forma
	}

	citaAbierta := false

	for _, pieza := range piezaDeFormaFija.FindAllString(resto, -1) {
		switch {
		case pieza[0] == ' ' || pieza[0] == '\t':
			patron.WriteString(blancoDeFormaFija)

			if citaAbierta {
				patron.WriteString(`)?`)
				citaAbierta = false
			}
		case pieza == marcadorDeFecha:
			patron.WriteString(fechaDeFormaFija)
		case pieza == marcadorDeCita && !citaAbierta:
			patron.WriteString(`(?:` + citaDeFormaFija)
			citaAbierta = true
		default:
			patron.WriteString(regexp.QuoteMeta(pieza))
		}
	}

	if citaAbierta {
		patron.WriteString(`)?`)
	}

	return regexp.MustCompile(patron.String())
}

// cabezaDeFormaFija separa de una forma fija que empieza por la marca su
// etiqueta —lo que va de la marca a los primeros dos puntos— y lo que sigue a
// esos dos puntos. conCabeza es falso, y la forma no tiene cabeza, si no empieza
// por la marca, si no tiene dos puntos o si su etiqueta no tiene palabras.
func cabezaDeFormaFija(forma string) (etiqueta, resto string, conCabeza bool) {
	sinMarca, conMarca := strings.CutPrefix(forma, marcaDeFormaFija)
	etiqueta, resto, conDosPuntos := strings.Cut(sinMarca, finalDeAviso)

	return etiqueta, resto, conMarca && conDosPuntos && strings.TrimSpace(etiqueta) != ""
}

// formaDeExpresion compila la expresión regular de una expresión prohibida: las
// palabras de la expresión en su orden, cada una sin distinguir mayúsculas y
// unidas por entrePalabrasDeAviso, entre antesDeLaExpresion y
// despuesDeLaExpresion (contrato lista-y-juicio §3). Las piezas son fijas y cada
// palabra pasa por regexp.QuoteMeta, así que la compilación no puede fallar.
func formaDeExpresion(expresion string) *regexp.Regexp {
	palabras := strings.Fields(expresion)
	for i, palabra := range palabras {
		palabras[i] = `(?i:` + regexp.QuoteMeta(palabra) + `)`
	}

	return regexp.MustCompile(antesDeLaExpresion + strings.Join(palabras, entrePalabrasDeAviso) + despuesDeLaExpresion)
}
