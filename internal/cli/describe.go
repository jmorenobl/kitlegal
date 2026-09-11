package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode"

	"github.com/invopop/jsonschema"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Las dos claves del documento que emite --describe. Son una sola descripción y
// no dos documentos: FR-046 pide «un esquema JSON válido que describa la entrada
// y la salida», y un consumidor MCP necesita las dos juntas (research.md D12).
const (
	claveEntrada = "entrada"
	claveSalida  = "salida"
)

// errEsquemaImposible es el fallo de quien declaró el verbo, no el de quien lo
// invoca: unos argumentos que no son un struct o un sobre que dejó de declarar
// sus claves. No lleva ninguno de los cinco sentinelas, así que sale con el
// código de lo que nadie previó y nunca con el de argumentos inválidos (FR-031).
var errEsquemaImposible = errors.New("cli: el esquema del verbo no se pudo construir")

// tipoDelSobre y tipoDeDuracion son los dos tipos que la generación necesita
// nombrar. El primero, porque la condición de la salida se decide sobre dos de
// sus claves; el segundo, porque su representación en la línea de órdenes no es
// la de su tipo Go.
var (
	tipoDelSobre     = reflect.TypeOf(schema.Sobre{})
	tipoDeDuracion   = reflect.TypeOf(time.Duration(0))
	tipoDeGlobales   = reflect.TypeOf(Globales{})
	tipoDatosDeFallo = reflect.TypeOf(schema.DatosError{})
)

// Verbo es lo que el kernel necesita saber de un verbo para describirlo: su
// nombre, el del applet que lo ofrece, una línea de descripción y los dos tipos
// de los que sale el esquema. La construye el paquete de composición desde su
// registro, de modo que este paquete no importa aquel ni conoce el tipo Verbo
// (research.md D1, D2).
//
// Argumentos y Salida son **valores** y no tipos: el primero es lo que devuelve
// la fábrica de argumentos del verbo y el segundo el valor cero del tipo de
// `data`. Ninguno de los dos se ejecuta ni se lee: solo se refleja (FR-049,
// contracts/registro-y-describe.md §1).
type Verbo struct {
	// Applet es el nombre con el que se invocó al applet, que es el que aparece
	// en el título del esquema.
	Applet string
	// Verbo es el nombre del verbo descrito. La autodescripción se refiere
	// siempre a un verbo concreto: el nombrado o el de omisión, que resuelve el
	// despacho antes de llegar aquí (contracts/registro-y-describe.md §2 bis).
	Verbo string
	// Ayuda es la línea que el verbo declara para la ayuda.
	Ayuda string
	// Argumentos es el struct de argumentos del verbo, con las etiquetas de la
	// gramática. Nulo significa un verbo que no declara ninguno: entonces la
	// entrada son solo las ocho banderas globales.
	Argumentos any
	// Salida es el valor cero del tipo de `data`. Nulo significa un verbo que no
	// lo declara: entonces el `data` de un sobre correcto queda sin restringir,
	// porque inventarle una forma describiría algo que nadie ha dicho.
	Salida any
}

// Describir emite en la salida estándar el esquema JSON de entrada y salida del
// verbo y no ejecuta nada: describirse y actuar son excluyentes (FR-046, FR-049).
//
// Devuelve error y no un código de salida porque la traducción de errores a
// códigos ocurre en un único punto, que es quien lo llama; un documento emitido
// sin fallo es el código 0 (FR-030, FR-045).
func Describir(p Presentador, def Verbo) error {
	esquema, err := def.esquema()
	if err != nil {
		return err
	}

	documento, err := serializar(esquema)
	if err != nil {
		return err
	}

	return p.Texto(documento)
}

// esquema construye el documento entero: las dos partes y las definiciones que
// ambas comparten, reunidas en la raíz porque es contra ella donde resuelven las
// referencias `#/$defs/…`.
func (d Verbo) esquema() (*jsonschema.Schema, error) {
	definiciones := jsonschema.Definitions{}

	entrada, err := esquemaDeEntrada(definiciones, d.Argumentos)
	if err != nil {
		return nil, err
	}

	salida, err := esquemaDeSalida(definiciones, d.Salida)
	if err != nil {
		return nil, err
	}

	partes := jsonschema.NewProperties()
	partes.Set(claveEntrada, entrada)
	partes.Set(claveSalida, salida)

	return &jsonschema.Schema{
		Version:     jsonschema.Version,
		Title:       strings.TrimSpace(d.Applet + " " + d.Verbo),
		Description: d.Ayuda,
		Properties:  partes,
		Definitions: definiciones,
	}, nil
}

// esquemaDeEntrada describe lo que se escribe en la línea de órdenes: los
// argumentos del verbo y, detrás, las ocho banderas globales que el applet no
// declara y recibe igualmente (FR-018, SC-010).
//
// Se recorren los campos en lugar de reflejar el struct entero porque las dos
// mitades vienen de dos tipos distintos y el documento las presenta como una
// sola: quien invoca no distingue una bandera del kernel de un argumento del
// verbo.
func esquemaDeEntrada(
	definiciones jsonschema.Definitions, argumentos any,
) (*jsonschema.Schema, error) {
	campos, err := camposDeLosArgumentos(argumentos)
	if err != nil {
		return nil, err
	}

	entrada := &jsonschema.Schema{
		Type:                 "object",
		Properties:           jsonschema.NewProperties(),
		AdditionalProperties: jsonschema.FalseSchema,
	}

	for _, campo := range append(campos, camposVisibles(tipoDeGlobales)...) {
		nombre := nombreEnLaInvocacion(campo)
		entrada.Properties.Set(nombre, reflejar(definiciones, campo.Type))

		if exigidoPorLaGramatica(campo) {
			entrada.Required = append(entrada.Required, nombre)
		}
	}

	return entrada, nil
}

// esquemaDeSalida describe el sobre completo con `data` condicionado a `ok`: el
// del applet cuando la operación fue bien y la forma común de error del kernel
// cuando no. Es lo que hace que una ejecución fallida valide contra el esquema
// que ese mismo applet emite (FR-047, SC-015).
func esquemaDeSalida(
	definiciones jsonschema.Definitions, salidaDelApplet any,
) (*jsonschema.Schema, error) {
	claveOk, err := claveDelSobre("Ok")
	if err != nil {
		return nil, err
	}

	claveDatos, err := claveDelSobre("Data")
	if err != nil {
		return nil, err
	}

	sobre := reflejarExpandido(definiciones, tipoDelSobre)
	sobre.If = sobreCuyaClave(claveOk, &jsonschema.Schema{Const: true})
	sobre.Then = sobreCuyaClave(claveDatos, datosDelApplet(definiciones, salidaDelApplet))
	sobre.Else = sobreCuyaClave(claveDatos, reflejar(definiciones, tipoDatosDeFallo))

	return sobre, nil
}

// sobreCuyaClave es una de las tres ramas de la condición: un esquema que solo
// dice qué forma tiene una de las claves del sobre y deja las otras cinco como
// estén.
func sobreCuyaClave(clave string, esquema *jsonschema.Schema) *jsonschema.Schema {
	propiedades := jsonschema.NewProperties()
	propiedades.Set(clave, esquema)

	return &jsonschema.Schema{Properties: propiedades}
}

// datosDelApplet es la forma del `data` de un sobre correcto. Un verbo que no
// declara su tipo de salida deja la clave sin restringir en lugar de recibir una
// forma inventada.
func datosDelApplet(definiciones jsonschema.Definitions, salida any) *jsonschema.Schema {
	if salida == nil {
		return jsonschema.TrueSchema
	}

	return reflejar(definiciones, reflect.TypeOf(salida))
}

// claveDelSobre lee de la etiqueta el nombre con el que un campo del sobre viaja
// en el documento. Las dos claves de las que depende la condición no se escriben
// aquí a propósito: son las mismas que serializa el sobre, así que si alguna
// cambiara de nombre la condición la seguiría sin que nadie edite este fichero
// (FR-048).
func claveDelSobre(nombre string) (string, error) {
	campo, existe := tipoDelSobre.FieldByName(nombre)
	if !existe {
		return "", fmt.Errorf("%w: el sobre no declara el campo %s", errEsquemaImposible, nombre)
	}

	clave, _, _ := strings.Cut(campo.Tag.Get("json"), ",")
	if clave == "" {
		return "", fmt.Errorf("%w: el campo %s del sobre no declara su clave json",
			errEsquemaImposible, nombre)
	}

	return clave, nil
}

// reflejar convierte un tipo en su esquema y traslada al documento las
// definiciones que la reflexión haya creado.
func reflejar(definiciones jsonschema.Definitions, tipo reflect.Type) *jsonschema.Schema {
	return trasladar(definiciones, reflector().ReflectFromType(tipo))
}

// reflejarExpandido es lo mismo con el tipo raíz escrito en su sitio en lugar de
// referenciado, que es como el sobre tiene que aparecer: la parte de salida es el
// sobre y no un apuntador a él.
func reflejarExpandido(
	definiciones jsonschema.Definitions, tipo reflect.Type,
) *jsonschema.Schema {
	generador := reflector()
	generador.ExpandedStruct = true

	return trasladar(definiciones, generador.ReflectFromType(tipo))
}

// trasladar recoge en las definiciones del documento las que trae una reflexión
// suelta y deja el esquema listo para colgar de él.
//
// El `$schema` se quita porque va una sola vez, en la raíz; las definiciones
// propias, porque todas las referencias son `#/$defs/…` y solo resuelven si están
// en la raíz del documento que se compila.
func trasladar(
	definiciones jsonschema.Definitions, reflejado *jsonschema.Schema,
) *jsonschema.Schema {
	for nombre, propia := range reflejado.Definitions {
		definiciones[nombre] = propia
	}

	reflejado.Definitions = nil
	reflejado.Version = ""

	return reflejado
}

// reflector es el generador con el que se refleja cada tipo.
//
// Anonymous no es un detalle: sin él cada reflexión llevaría su propio `$id`, y
// una referencia `#/$defs/…` escrita bajo ese identificador resolvería contra esa
// base y no contra el documento emitido, de modo que el esquema no compilaría.
func reflector() *jsonschema.Reflector {
	return &jsonschema.Reflector{
		Anonymous: true,
		Mapper:    representacionEnLaInvocacion,
	}
}

// representacionEnLaInvocacion describe los tipos cuya forma en la línea de
// órdenes no es la de su tipo Go. Solo hay uno: la duración de --timeout, que se
// escribe «30s» y no el número de nanosegundos con que Go la representa.
func representacionEnLaInvocacion(tipo reflect.Type) *jsonschema.Schema {
	if tipo == tipoDeDuracion {
		return &jsonschema.Schema{Type: "string"}
	}

	return nil
}

// camposDeLosArgumentos son los campos del struct de argumentos del verbo. Un
// verbo sin argumentos no es un error: es un verbo que solo acepta las globales.
func camposDeLosArgumentos(argumentos any) ([]reflect.StructField, error) {
	if argumentos == nil {
		return nil, nil
	}

	tipo := reflect.TypeOf(argumentos)
	for tipo.Kind() == reflect.Pointer {
		tipo = tipo.Elem()
	}

	if tipo.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%w: los argumentos del verbo son %s y no un struct",
			errEsquemaImposible, tipo)
	}

	return camposVisibles(tipo), nil
}

// camposVisibles son los que la gramática puede llegar a rellenar: los
// exportados, con los de un struct embebido ya aplanados por reflect. El propio
// campo embebido se descarta porque sus hijos ya vienen en la lista y él no nombra
// ninguna bandera.
func camposVisibles(tipo reflect.Type) []reflect.StructField {
	campos := make([]reflect.StructField, 0, tipo.NumField())

	for _, campo := range reflect.VisibleFields(tipo) {
		if !campo.IsExported() || (campo.Anonymous && campo.Type.Kind() == reflect.Struct) {
			continue
		}

		campos = append(campos, campo)
	}

	return campos
}

// nombreEnLaInvocacion es el nombre con el que el campo se escribe en la línea de
// órdenes: la etiqueta que lo renombra cuando la hay y, si no, el nombre del
// campo en minúsculas y separado por guiones, que es la regla con la que el
// analizador lo deriva. Describir otra cosa describiría un binario distinto del
// que se publica.
func nombreEnLaInvocacion(campo reflect.StructField) string {
	if nombre, renombrado := campo.Tag.Lookup("name"); renombrado && nombre != "" {
		return conGuiones(nombre)
	}

	return conGuiones(campo.Name)
}

// conGuiones convierte NombreDeCampo en nombre-de-campo y deja intacto lo que ya
// viene escrito así, de modo que aplicarlo a un nombre ya derivado no lo cambia.
func conGuiones(nombre string) string {
	letras := []rune(nombre)

	var derivado strings.Builder

	for i, letra := range letras {
		if !unicode.IsUpper(letra) {
			derivado.WriteRune(letra)

			continue
		}

		if i > 0 && (!unicode.IsUpper(letras[i-1]) ||
			(i+1 < len(letras) && !unicode.IsUpper(letras[i+1]))) {
			derivado.WriteRune('-')
		}

		derivado.WriteRune(unicode.ToLower(letra))
	}

	return derivado.String()
}

// exigidoPorLaGramatica dice si la invocación tiene que escribir ese campo. Es la
// regla del analizador: lo marcado como obligatorio lo es, lo marcado como
// opcional o con valor por omisión no lo es, y un argumento de posición lo es
// por serlo. Ninguna de las ocho globales lo es, y por eso describirlas todas como
// obligatorias —que es lo que saldría de la reflexión sin esta regla— describiría
// un binario que no existe.
func exigidoPorLaGramatica(campo reflect.StructField) bool {
	if _, obligatorio := campo.Tag.Lookup("required"); obligatorio {
		return true
	}

	if _, opcional := campo.Tag.Lookup("optional"); opcional {
		return false
	}

	if _, conOmision := campo.Tag.Lookup("default"); conOmision {
		return false
	}

	_, dePosicion := campo.Tag.Lookup("arg")

	return dePosicion
}

// serializar escribe el documento sin escapar caracteres HTML —la misma regla que
// el sobre— y con sangrado, porque lo lee tanto una persona como un agente. El
// salto final lo pone quien escribe, no este texto.
func serializar(esquema *jsonschema.Schema) (string, error) {
	var documento strings.Builder

	codificador := json.NewEncoder(&documento)
	codificador.SetEscapeHTML(false)
	codificador.SetIndent("", "  ")

	if err := codificador.Encode(esquema); err != nil {
		return "", fmt.Errorf("%w: %w", errEsquemaImposible, err)
	}

	return strings.TrimRight(documento.String(), "\n"), nil
}
