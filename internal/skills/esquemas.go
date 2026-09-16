package skills

import (
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/santhosh-tekuri/jsonschema/v6/kind"
	"go.yaml.in/yaml/v3"
)

// urlDelEsquema es la dirección con la que CompilarEsquema registra el esquema en
// su compilador. No se publica en ningún sitio: solo identifica el recurso dentro
// del compilador, y las referencias internas del esquema (#/$defs/…) se resuelven
// dentro de él.
const urlDelEsquema = "urn:kitlegal:esquema"

// CompilarEsquema compila un JSON Schema 2020-12 con las aserciones de formato
// activas: sin ellas, en ese borrador format es solo una anotación y un valor
// mal formado pasaría la validación (research.md D8).
func CompilarEsquema(contenido []byte) (*jsonschema.Schema, error) {
	documento, err := jsonschema.UnmarshalJSON(bytes.NewReader(contenido))
	if err != nil {
		return nil, fmt.Errorf("el esquema no es un documento JSON: %w", err)
	}

	compilador := jsonschema.NewCompiler()
	compilador.AssertFormat()

	if err := compilador.AddResource(urlDelEsquema, documento); err != nil {
		return nil, fmt.Errorf("el esquema no se puede registrar en el compilador: %w", err)
	}

	esquema, err := compilador.Compile(urlDelEsquema)
	if err != nil {
		return nil, fmt.Errorf("el esquema no compila: %w", err)
	}

	return esquema, nil
}

// ClaveRepetida es el defecto de un mapa del documento que tiene la misma clave
// más de una vez. Analizar a yaml.Node no lo comprueba y, al convertir el nodo,
// la última clave ganaría en silencio (research.md V43): por eso lo busca el
// recorrido del lector antes de convertir nada.
type ClaveRepetida struct {
	// Mapa es la ruta del mapa dentro del documento —sus claves e índices de
	// lista desde la raíz—, vacía para el mapa de la raíz.
	Mapa []string

	// Clave es la clave repetida.
	Clave string

	// Lineas son la línea de la primera aparición de la clave y la de esta
	// repetición.
	Lineas [2]int
}

// Error nombra el mapa, la clave y sus dos líneas, como «normas:
// BOE-A-2015-10565 repetido en las líneas 12 y 20» (data-model, «Lectura de
// documentos YAML»); en el mapa de la raíz, sin ruta delante.
func (c *ClaveRepetida) Error() string {
	defecto := fmt.Sprintf("%s repetido en las líneas %d y %d", c.Clave, c.Lineas[0], c.Lineas[1])
	if len(c.Mapa) == 0 {
		return defecto
	}

	return presentarRuta(c.Mapa) + ": " + defecto
}

// DefectoEnElDocumento es un defecto en un punto del documento que no es una
// clave repetida: un segundo documento YAML, un valor que JSON no puede
// representar o un incumplimiento del esquema.
type DefectoEnElDocumento struct {
	// Ruta es la ruta del punto dentro del documento —claves de mapa e índices
	// de lista desde la raíz—, vacía en la raíz.
	Ruta []string

	// Linea es la línea del nodo que ocupa la ruta, o 0 en un documento vacío,
	// que no tiene ninguna.
	Linea int

	// Motivo dice qué falla. En un incumplimiento del esquema es el texto con el
	// que lo presenta la biblioteca de validación.
	Motivo string

	// Tipo es el incumplimiento que detectó el esquema, o nil si el defecto no
	// viene del esquema. Con él, quien lee un documento puede presentar el
	// defecto con sus propias palabras.
	Tipo jsonschema.ErrorKind
}

// Error presenta el defecto con su ruta y su línea delante, como «comandos/0,
// línea 5: additional properties 'verbo' not allowed», y sin la parte que el
// defecto no tiene.
func (d *DefectoEnElDocumento) Error() string {
	var ubicacion []string
	if len(d.Ruta) > 0 {
		ubicacion = append(ubicacion, presentarRuta(d.Ruta))
	}

	if d.Linea > 0 {
		ubicacion = append(ubicacion, fmt.Sprintf("línea %d", d.Linea))
	}

	if len(ubicacion) == 0 {
		return d.Motivo
	}

	return strings.Join(ubicacion, ", ") + ": " + d.Motivo
}

// ValidarDocumentoYAML lee un documento YAML con el lector común de data-model
// («Lectura de documentos YAML»; research.md D8), el mismo para los datos, las
// evals y el frontmatter, y lo devuelve leído en un T según las etiquetas yaml
// de sus campos:
//
//  1. lo analiza con go.yaml.in/yaml/v3 a un yaml.Node, que conserva la línea de
//     cada nodo, y exige que el contenido lleve un único documento;
//  2. hace el recorrido del nodo y rechaza toda clave repetida en cualquier
//     mapa, con una ClaveRepetida por repetición;
//  3. lo convierte con (*yaml.Node).Decode y lo normaliza a tipos JSON,
//     codificándolo en JSON y leyéndolo con jsonschema.UnmarshalJSON (números
//     como json.Number, research.md V44); cada valor que JSON no puede
//     representar —un mapa con claves que no son texto, un número infinito o
//     NaN— es un DefectoEnElDocumento con su ruta;
//  4. si esquema no es nil, lo valida contra él, y cada incumplimiento es un
//     DefectoEnElDocumento con la ruta dentro del documento y la línea del nodo
//     que la ocupa.
//
// Ningún error de estos pasos se descarta: el primer paso que encuentra defectos
// los devuelve todos, en el orden del documento y unidos con errors.Join, y los
// pasos siguientes no se ejecutan. Con un error, el T devuelto es su valor cero.
func ValidarDocumentoYAML[T any](contenido []byte, esquema *jsonschema.Schema) (T, error) {
	var cero T

	raiz, err := analizar(contenido)
	if err != nil {
		return cero, err
	}

	if repetidas := clavesRepetidas(raiz, nil); len(repetidas) > 0 {
		return cero, errors.Join(repetidas...)
	}

	valor, err := normalizar(raiz)
	if err != nil {
		return cero, err
	}

	if esquema != nil {
		if err := validar(raiz, esquema, valor); err != nil {
			return cero, err
		}
	}

	var leido T
	if err := raiz.Decode(&leido); err != nil {
		return cero, fmt.Errorf("el documento YAML no se puede leer como %T: %w", leido, err)
	}

	return leido, nil
}

// analizar lee a un yaml.Node el único documento YAML del contenido. Un
// contenido vacío es un documento vacío, cuyo nodo es el valor cero.
// yaml.Unmarshal se quedaría con el primer documento y descartaría los
// siguientes sin decir nada, así que el lector pide un segundo documento y
// exige que no lo haya.
func analizar(contenido []byte) (*yaml.Node, error) {
	lector := yaml.NewDecoder(bytes.NewReader(contenido))

	var raiz yaml.Node
	if err := lector.Decode(&raiz); err != nil {
		if errors.Is(err, io.EOF) {
			return &raiz, nil
		}

		return nil, fmt.Errorf("no es YAML válido: %w", err)
	}

	var segundo yaml.Node

	err := lector.Decode(&segundo)
	switch {
	case errors.Is(err, io.EOF):
		return &raiz, nil
	case err != nil:
		return nil, fmt.Errorf("no es YAML válido: %w", err)
	default:
		return nil, &DefectoEnElDocumento{
			Linea:  segundo.Line,
			Motivo: "empieza un segundo documento YAML y solo se admite uno",
		}
	}
}

// clavesRepetidas hace el recorrido del nodo en el orden del documento y
// devuelve una ClaveRepetida por cada clave que ya había aparecido antes en su
// mapa, con la línea de la primera aparición. Compara las claves como las
// compara (*yaml.Node).Decode —mismo tipo de nodo y mismo texto—, de modo que el
// recorrido no deja pasar ninguna repetición que Decode rechazaría. No sigue los
// alias: el nodo al que apunta un alias entra en el recorrido donde el documento
// lo escribe.
func clavesRepetidas(nodo *yaml.Node, ruta []string) []error {
	var repetidas []error

	switch nodo.Kind {
	case yaml.DocumentNode:
		for _, hijo := range nodo.Content {
			repetidas = append(repetidas, clavesRepetidas(hijo, ruta)...)
		}
	case yaml.SequenceNode:
		for indice, hijo := range nodo.Content {
			repetidas = append(repetidas, clavesRepetidas(hijo, rutaHija(ruta, strconv.Itoa(indice)))...)
		}
	case yaml.MappingNode:
		for indice := 0; indice+1 < len(nodo.Content); indice += 2 {
			clave := nodo.Content[indice]
			if primera := primeraAparicion(nodo.Content[:indice], clave); primera != nil {
				repetidas = append(repetidas, &ClaveRepetida{
					Mapa:   ruta,
					Clave:  clave.Value,
					Lineas: [2]int{primera.Line, clave.Line},
				})
			}

			repetidas = append(repetidas, clavesRepetidas(nodo.Content[indice+1], rutaHija(ruta, clave.Value))...)
		}
	default:
		// Ni un escalar, ni un alias, ni un documento vacío contienen claves.
	}

	return repetidas
}

// primeraAparicion devuelve, entre las claves y valores anteriores de un mapa,
// la primera clave igual a la dada, o nil si no la hay.
func primeraAparicion(anteriores []*yaml.Node, clave *yaml.Node) *yaml.Node {
	for indice := 0; indice < len(anteriores); indice += 2 {
		if anteriores[indice].Kind == clave.Kind && anteriores[indice].Value == clave.Value {
			return anteriores[indice]
		}
	}

	return nil
}

// normalizar convierte el documento con (*yaml.Node).Decode, que también rechaza
// la clave repetida (research.md V43), comprueba que JSON puede representar
// todos sus valores y lo devuelve con los tipos de un documento JSON: lo codifica
// en JSON y lo lee con jsonschema.UnmarshalJSON, la lectura con la que la propia
// biblioteca de validación lee sus documentos (research.md V44).
func normalizar(raiz *yaml.Node) (any, error) {
	var valor any
	if err := raiz.Decode(&valor); err != nil {
		return nil, fmt.Errorf("el documento YAML no se puede leer: %w", err)
	}

	if err := unir(sinRepresentacionJSON(raiz, valor, nil)); err != nil {
		return nil, err
	}

	codificado, err := json.Marshal(valor)
	if err != nil {
		return nil, fmt.Errorf("el documento no se puede codificar en JSON: %w", err)
	}

	normalizado, err := jsonschema.UnmarshalJSON(bytes.NewReader(codificado))
	if err != nil {
		return nil, fmt.Errorf("el documento codificado en JSON no se puede leer: %w", err)
	}

	return normalizado, nil
}

// sinRepresentacionJSON devuelve un DefectoEnElDocumento por cada valor que JSON
// no puede representar: un mapa con alguna clave que no es texto, que
// (*yaml.Node).Decode entrega como map[any]any, y un número infinito o NaN. Sin
// esta comprobación, json.Marshal fallaría sin decir dónde.
func sinRepresentacionJSON(raiz *yaml.Node, valor any, ruta []string) []*DefectoEnElDocumento {
	var defectos []*DefectoEnElDocumento

	switch valor := valor.(type) {
	case map[any]any:
		defectos = append(defectos, defectoEn(raiz, ruta, "mapa con claves que no son texto"))
	case map[string]any:
		for _, clave := range slices.Sorted(maps.Keys(valor)) {
			defectos = append(defectos, sinRepresentacionJSON(raiz, valor[clave], rutaHija(ruta, clave))...)
		}
	case []any:
		for indice, elemento := range valor {
			defectos = append(defectos, sinRepresentacionJSON(raiz, elemento, rutaHija(ruta, strconv.Itoa(indice)))...)
		}
	case float64:
		if math.IsInf(valor, 0) || math.IsNaN(valor) {
			defectos = append(defectos, defectoEn(raiz, ruta, "número que JSON no puede representar"))
		}
	}

	return defectos
}

// validar valida el documento normalizado contra el esquema y devuelve un
// DefectoEnElDocumento por cada incumplimiento, con la ruta del valor dentro del
// documento, la línea de su nodo y su tipo, en el orden del documento.
func validar(raiz *yaml.Node, esquema *jsonschema.Schema, valor any) error {
	err := esquema.Validate(valor)
	if err == nil {
		return nil
	}

	var validacion *jsonschema.ValidationError
	if !errors.As(err, &validacion) {
		return fmt.Errorf("el documento no se puede validar: %w", err)
	}

	var defectos []*DefectoEnElDocumento

	for _, incumplimiento := range incumplimientosDe(validacion) {
		defectos = append(defectos, defectoDelEsquema(raiz, incumplimiento))
	}

	return unir(defectos)
}

// defectoDelEsquema es el DefectoEnElDocumento de un incumplimiento del esquema,
// con la ruta que da el validador, la línea de su nodo y su tipo.
//
// En propertyNames la ruta del validador no es de fiar:
// santhosh-tekuri/jsonschema v6.0.3 guarda como ruta del error el búfer de ruta
// del validador sin copiarlo (validator.go 296), y la validación de las
// propiedades siguientes lo sobrescribe, así que la ruta que llega puede nombrar
// otro punto del documento, y cuál depende del orden en que el validador visita
// las claves de un mapa. Solo su longitud es fiable. Por eso la ruta se busca en
// el documento: la del mapa que está a esa profundidad y tiene la clave que
// nombra el incumplimiento. Si hay más de uno, el defecto va sin ruta ni línea
// antes que con las de otro sitio.
func defectoDelEsquema(raiz *yaml.Node, incumplimiento *jsonschema.ValidationError) *DefectoEnElDocumento {
	ruta := slices.Clone(incumplimiento.InstanceLocation)

	if nombre, deNombreDeClave := incumplimiento.ErrorKind.(*kind.PropertyNames); deNombreDeClave {
		mapas := mapasConLaClave(raiz, nil, len(ruta), nombre.Property)
		if len(mapas) != 1 {
			return &DefectoEnElDocumento{Motivo: motivoDe(incumplimiento), Tipo: incumplimiento.ErrorKind}
		}

		ruta = mapas[0]
	}

	defecto := defectoEn(raiz, ruta, motivoDe(incumplimiento))
	defecto.Tipo = incumplimiento.ErrorKind

	return defecto
}

// mapasConLaClave devuelve, en el orden del documento, la ruta de cada mapa que
// está a la profundidad dada y tiene la clave.
func mapasConLaClave(nodo *yaml.Node, ruta []string, profundidad int, clave string) [][]string {
	nodo = sinEnvoltorio(nodo)

	if len(ruta) == profundidad {
		if nodo.Kind == yaml.MappingNode && hijoDe(nodo, clave) != nil {
			return [][]string{ruta}
		}

		return nil
	}

	var mapas [][]string

	switch nodo.Kind {
	case yaml.MappingNode:
		for indice := 0; indice+1 < len(nodo.Content); indice += 2 {
			hija := rutaHija(ruta, nodo.Content[indice].Value)
			mapas = append(mapas, mapasConLaClave(nodo.Content[indice+1], hija, profundidad, clave)...)
		}
	case yaml.SequenceNode:
		for indice, hijo := range nodo.Content {
			mapas = append(mapas, mapasConLaClave(hijo, rutaHija(ruta, strconv.Itoa(indice)), profundidad, clave)...)
		}
	default:
		// Ni un escalar ni un documento vacío llevan mapas dentro.
	}

	return mapas
}

// incumplimientosDe devuelve los incumplimientos concretos del árbol de errores
// del validador: los nodos sin causas, que son los que dicen qué palabra clave
// falla y dónde, sin los que solo agrupan a otros —el esquema, una referencia,
// oneOf—. propertyNames es la excepción: el validador valida el nombre de la
// clave como un valor suelto, así que la ubicación de sus causas es la de ese
// valor y no la del mapa, y el incumplimiento es el nodo que nombra la clave,
// con sus causas en el motivo.
func incumplimientosDe(nodo *jsonschema.ValidationError) []*jsonschema.ValidationError {
	if _, deNombreDeClave := nodo.ErrorKind.(*kind.PropertyNames); deNombreDeClave || len(nodo.Causes) == 0 {
		return []*jsonschema.ValidationError{nodo}
	}

	var concretos []*jsonschema.ValidationError
	for _, causa := range nodo.Causes {
		concretos = append(concretos, incumplimientosDe(causa)...)
	}

	return concretos
}

// motivoDe es el texto del incumplimiento seguido del de cada una de sus causas,
// separados por dos puntos.
func motivoDe(incumplimiento *jsonschema.ValidationError) string {
	partes := []string{textoDe(incumplimiento.ErrorKind)}

	for _, causa := range incumplimiento.Causes {
		for _, concreto := range incumplimientosDe(causa) {
			partes = append(partes, textoDe(concreto.ErrorKind))
		}
	}

	return strings.Join(partes, ": ")
}

// textoDe es el texto con el que la biblioteca de validación presenta un tipo de
// incumplimiento. La biblioteca solo lo escribe a través de un *message.Printer
// de golang.org/x/text, que el módulo no importa directamente; su salida básica
// lo escribe con el suyo, y de un error sin causas da una única unidad con ese
// texto.
func textoDe(tipo jsonschema.ErrorKind) string {
	return (&jsonschema.ValidationError{ErrorKind: tipo}).BasicOutput().Error.String()
}

// defectoEn es el DefectoEnElDocumento del punto de la ruta, con la línea de su
// nodo. Una ruta vacía queda como nil, venga de donde venga.
func defectoEn(raiz *yaml.Node, ruta []string, motivo string) *DefectoEnElDocumento {
	if len(ruta) == 0 {
		ruta = nil
	}

	return &DefectoEnElDocumento{Ruta: ruta, Linea: lineaDe(raiz, ruta), Motivo: motivo}
}

// lineaDe es la línea del nodo que ocupa la ruta dentro del documento. Sigue los
// alias hasta el nodo al que apuntan y, si un paso de la ruta no está escrito en
// el documento —una clave que llega por una fusión <<—, se queda con la línea del
// último nodo que sí lo está. Un documento vacío no tiene líneas: 0.
func lineaDe(raiz *yaml.Node, ruta []string) int {
	nodo := sinEnvoltorio(raiz)

	for _, paso := range ruta {
		hijo := hijoDe(nodo, paso)
		if hijo == nil {
			break
		}

		nodo = hijo
	}

	return nodo.Line
}

// sinEnvoltorio salta el nodo de documento y los alias hasta el nodo que lleva
// el valor.
func sinEnvoltorio(nodo *yaml.Node) *yaml.Node {
	for {
		switch {
		case nodo.Kind == yaml.DocumentNode && len(nodo.Content) == 1:
			nodo = nodo.Content[0]
		case nodo.Kind == yaml.AliasNode && nodo.Alias != nil:
			nodo = nodo.Alias
		default:
			return nodo
		}
	}
}

// hijoDe es el nodo, sin envoltorio, al que lleva un paso de la ruta desde un
// mapa —la clave— o desde una lista —el índice—, o nil si el paso no está
// escrito en el nodo.
func hijoDe(nodo *yaml.Node, paso string) *yaml.Node {
	switch nodo.Kind {
	case yaml.MappingNode:
		for indice := 0; indice+1 < len(nodo.Content); indice += 2 {
			if nodo.Content[indice].Value == paso {
				return sinEnvoltorio(nodo.Content[indice+1])
			}
		}
	case yaml.SequenceNode:
		if indice, err := strconv.Atoi(paso); err == nil && indice >= 0 && indice < len(nodo.Content) {
			return sinEnvoltorio(nodo.Content[indice])
		}
	default:
		// Ni un escalar ni un documento vacío tienen hijos.
	}

	return nil
}

// rutaHija es la ruta con un paso más, en una copia: las rutas de los defectos
// no comparten memoria entre sí.
func rutaHija(ruta []string, paso string) []string {
	return append(slices.Clone(ruta), paso)
}

// presentarRuta escribe una ruta dentro del documento con sus pasos separados
// por /, como citas/0/bloque.
func presentarRuta(ruta []string) string {
	return strings.Join(ruta, "/")
}

// unir ordena los defectos por línea —a igual línea, por su texto— y los
// devuelve unidos con errors.Join, o nil si no hay ninguno.
func unir(defectos []*DefectoEnElDocumento) error {
	slices.SortStableFunc(defectos, func(a, b *DefectoEnElDocumento) int {
		return cmp.Or(cmp.Compare(a.Linea, b.Linea), strings.Compare(a.Error(), b.Error()))
	})

	errores := make([]error, 0, len(defectos))
	for _, defecto := range defectos {
		errores = append(errores, defecto)
	}

	return errors.Join(errores...)
}
