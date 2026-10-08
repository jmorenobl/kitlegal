package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"path"
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

// claveDeBanderas es la anotación con la que la parte de entrada del documento
// nombra las banderas propias del verbo. Lleva el prefijo `x-` porque no es del
// vocabulario del borrador: un validador la ignora, y quien la lee es la tabla
// de comandos de las skills, que sin ella no distingue una bandera de un
// argumento de posición opcional (research.md D14 de H23).
const claveDeBanderas = "x-banderas"

// errEsquemaImposible es el fallo de quien declaró el verbo, no el de quien lo
// invoca: unos argumentos que no son un struct, un sobre que dejó de declarar
// sus claves o dos tipos distintos que se describirían con el mismo nombre. No
// lleva ninguno de los seis sentinelas, así que sale con el código de lo que
// nadie previó y nunca con el de argumentos inválidos (FR-031).
var errEsquemaImposible = errors.New("cli: el esquema del verbo no se pudo construir")

// Los tipos que la generación necesita nombrar. El sobre, porque la condición
// de la salida se decide sobre dos de sus claves; la duración y la clase de
// error, porque su representación en el documento no es la de su tipo Go; las
// globales, porque sus campos son la mitad de la entrada; y los datos de fallo,
// porque son la rama `else` de la condición.
var (
	tipoDelSobre     = reflect.TypeOf(schema.Sobre{})
	tipoDeDuracion   = reflect.TypeOf(time.Duration(0))
	tipoDeClase      = reflect.TypeOf(schema.Clase(""))
	tipoDeGlobales   = reflect.TypeOf(Globales{})
	tipoDatosDeFallo = reflect.TypeOf(schema.DatosError{})
)

// paqueteDelSobre es el paquete cuyos tipos se nombran en el documento tal y
// como el contrato los nombra —`DatosError`, sin cualificar—: son el
// vocabulario del sobre, no el de un applet.
var paqueteDelSobre = tipoDelSobre.PkgPath()

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
	campos, err := camposDeLosArgumentos(d.Argumentos)
	if err != nil {
		return nil, err
	}

	g := nuevoGenerador()

	// Los argumentos del verbo y, detrás, las ocho banderas globales que el
	// applet no declara y recibe igualmente (FR-018, SC-010): quien invoca no
	// distingue una bandera del kernel de un argumento del verbo.
	entrada := g.entrada(append(campos, camposVisibles(tipoDeGlobales)...))
	anotarBanderas(entrada, campos)

	salida, err := g.salida(d.Salida)
	if err != nil {
		return nil, err
	}

	// Se comprueba después de reflejar las dos partes y antes de emitir nada:
	// un documento cuyas definiciones no se pueden nombrar sin ambigüedad no
	// describe el verbo, y no se emite a medias (FR-047, FR-048).
	if g.colision != nil {
		return nil, g.colision
	}

	partes := jsonschema.NewProperties()
	partes.Set(claveEntrada, entrada)
	partes.Set(claveSalida, salida)

	return &jsonschema.Schema{
		Version:     jsonschema.Version,
		Title:       strings.TrimSpace(d.Applet + " " + d.Verbo),
		Description: d.Ayuda,
		Properties:  partes,
		Definitions: g.definiciones,
	}, nil
}

// generador es el estado de la construcción de **un** documento: las
// definiciones que las dos partes comparten y el nombre con el que cada tipo
// aparece en ellas.
//
// Nombrar es lo que evita la colisión que una reflexión suelta dejaría pasar en
// silencio: la biblioteca identifica cada definición por el nombre del tipo y,
// cuando dos tipos distintos se llaman igual, el segundo referencia o pisa la
// definición del primero sin decir nada. Aquí los tipos del sobre conservan el
// nombre que el contrato les da y los de cualquier otro paquete van
// cualificados con el nombre de su paquete —`ejemplo.mensaje`—, de modo que un
// applet que declare su propio `DatosError` no pise el del kernel; y si aun así
// dos tipos distintos acabaran con el mismo nombre, el documento no se
// construye (errEsquemaImposible), nunca se emite uno que describa otra cosa.
type generador struct {
	// definiciones son los `$defs` del documento, comunes a las dos partes.
	definiciones jsonschema.Definitions
	// vistos asocia cada nombre de definición ya usado con el tipo que lo
	// ocupa, para detectar que otro tipo lo reclama.
	vistos map[string]reflect.Type
	// colision es el primer choque de nombres detectado, o nulo.
	colision error
}

func nuevoGenerador() *generador {
	return &generador{
		definiciones: jsonschema.Definitions{},
		vistos:       map[string]reflect.Type{},
	}
}

// entrada describe lo que se escribe con esos campos: un objeto cerrado con una
// propiedad por campo, obligatoria si la gramática la exige.
//
// Se recorren los campos en lugar de reflejar un struct entero porque quien
// decide cuáles son es quien llama: el documento de --describe presenta como una
// sola las dos mitades de la línea de órdenes, que vienen de dos tipos distintos
// —los argumentos del verbo y las globales—, y una herramienta del servidor MCP
// solo los del verbo (herramienta.go).
func (g *generador) entrada(campos []reflect.StructField) *jsonschema.Schema {
	entrada := &jsonschema.Schema{
		Type:                 "object",
		Properties:           jsonschema.NewProperties(),
		AdditionalProperties: jsonschema.FalseSchema,
	}

	for _, campo := range campos {
		nombre := nombreEnLaInvocacion(campo)
		entrada.Properties.Set(nombre, g.reflejar(campo.Type))

		if exigidoPorLaGramatica(campo) {
			entrada.Required = append(entrada.Required, nombre)
		}
	}

	return entrada
}

// anotarBanderas deja dicho en la entrada del documento cuáles de los campos del
// verbo se escriben como banderas: los que no van por su posición, con el nombre
// de su propiedad y en el orden de sus campos. Las ocho globales no están entre
// ellos: son las mismas en todos los verbos y el documento de uno sin argumentos
// ya las da.
//
// Un verbo sin banderas propias no lleva la anotación, ni vacía: su documento es,
// byte a byte, el de antes de que existiera. Y es del documento de --describe y
// no de la entrada que construye el generador, porque dice cómo se escribe la
// orden: el esquema de entrada de una herramienta, que recibe un objeto, no la
// lleva (herramienta.go).
func anotarBanderas(entrada *jsonschema.Schema, campos []reflect.StructField) {
	var banderas []string

	for _, campo := range campos {
		if _, dePosicion := campo.Tag.Lookup("arg"); !dePosicion {
			banderas = append(banderas, nombreEnLaInvocacion(campo))
		}
	}

	if len(banderas) == 0 {
		return
	}

	entrada.Extras = map[string]any{claveDeBanderas: banderas}
}

// salida describe el sobre completo con `data` condicionado a `ok`: el del
// applet cuando la operación fue bien y la forma común de error del kernel
// cuando no. Es lo que hace que una ejecución fallida valide contra el esquema
// que ese mismo applet emite (FR-047, SC-015).
//
// Las restricciones de cada clave —`fuente` y `url` no vacías, `url` con formato
// de URI, `hash` con el patrón de la huella, `clase` dentro del vocabulario y
// `mensaje` no vacío— no se escriben aquí: salen de las etiquetas de los tipos
// del sobre y del vocabulario de clases del dominio, que es lo que hace cierto
// que el esquema emitido sea el del contrato y no una copia (FR-017, FR-048,
// contracts/sobre-de-salida.md §6).
func (g *generador) salida(salidaDelApplet any) (*jsonschema.Schema, error) {
	claveOk, err := claveDelSobre("Ok")
	if err != nil {
		return nil, err
	}

	claveDatos, err := claveDelSobre("Data")
	if err != nil {
		return nil, err
	}

	sobre := g.reflejarExpandido(tipoDelSobre)
	sobre.If = sobreCuyaClave(claveOk, &jsonschema.Schema{Const: true})
	sobre.Then = sobreCuyaClave(claveDatos, g.datosDelApplet(salidaDelApplet))
	sobre.Else = sobreCuyaClave(claveDatos, g.reflejar(tipoDatosDeFallo))

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
func (g *generador) datosDelApplet(salida any) *jsonschema.Schema {
	if salida == nil {
		return jsonschema.TrueSchema
	}

	return g.reflejar(reflect.TypeOf(salida))
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
func (g *generador) reflejar(tipo reflect.Type) *jsonschema.Schema {
	return g.trasladar(g.reflector().ReflectFromType(tipo))
}

// reflejarExpandido es lo mismo con el tipo raíz escrito en su sitio en lugar de
// referenciado, que es como el sobre tiene que aparecer: la parte de salida es el
// sobre y no un apuntador a él.
func (g *generador) reflejarExpandido(tipo reflect.Type) *jsonschema.Schema {
	generador := g.reflector()
	generador.ExpandedStruct = true

	return g.trasladar(generador.ReflectFromType(tipo))
}

// trasladar recoge en las definiciones del documento las que trae una reflexión
// suelta y deja el esquema listo para colgar de él.
//
// El `$schema` se quita porque va una sola vez, en la raíz; las definiciones
// propias, porque todas las referencias son `#/$defs/…` y solo resuelven si están
// en la raíz del documento que se compila. Que un nombre repetido entre dos
// reflexiones sea siempre el mismo tipo —y por tanto la misma definición— lo
// garantiza nombre, que es quien vigila las colisiones.
func (g *generador) trasladar(reflejado *jsonschema.Schema) *jsonschema.Schema {
	for nombre, propia := range reflejado.Definitions {
		g.definiciones[nombre] = propia
	}

	reflejado.Definitions = nil
	reflejado.Version = ""

	return reflejado
}

// reflector es la biblioteca configurada para este documento.
//
// Anonymous no es un detalle: sin él cada reflexión llevaría su propio `$id`, y
// una referencia `#/$defs/…` escrita bajo ese identificador resolvería contra esa
// base y no contra el documento emitido, de modo que el esquema no compilaría.
func (g *generador) reflector() *jsonschema.Reflector {
	return &jsonschema.Reflector{
		Anonymous: true,
		Mapper:    representacionEnElDocumento,
		Namer:     g.nombre,
	}
}

// nombre decide cómo se llama un tipo en los `$defs` del documento y anota qué
// tipo ocupa cada nombre. La biblioteca lo consulta para todo tipo con nombre,
// pero solo los struct llegan a ser definiciones: los demás se describen en su
// sitio, así que solo los struct entran en la vigilancia de colisiones.
//
// Los tipos del paquete del sobre se llaman como el contrato los llama; los de
// cualquier otro paquete llevan delante el nombre de su paquete, que es lo que
// separa el `DatosError` del kernel del `DatosError` que un applet pudiera
// declarar. Un tipo sin nombre no se nombra: la biblioteca lo describe en su
// sitio, como hace por omisión.
func (g *generador) nombre(tipo reflect.Type) string {
	if tipo.Name() == "" {
		return ""
	}

	nombre := tipo.Name()
	if tipo.PkgPath() != "" && tipo.PkgPath() != paqueteDelSobre {
		nombre = path.Base(tipo.PkgPath()) + "." + nombre
	}

	if tipo.Kind() == reflect.Struct {
		g.registrar(nombre, tipo)
	}

	return nombre
}

// registrar anota que el nombre lo ocupa ese tipo, y deja constancia del choque
// si ya lo ocupaba otro distinto. Solo se guarda el primero: es el que nombra
// los dos tipos en conflicto, que es lo que quien escribió el applet necesita
// para deshacerlo.
func (g *generador) registrar(nombre string, tipo reflect.Type) {
	ocupante, ocupado := g.vistos[nombre]
	if !ocupado {
		g.vistos[nombre] = tipo

		return
	}

	if ocupante != tipo && g.colision == nil {
		g.colision = fmt.Errorf("%w: los tipos %s y %s se describirían los dos como %q en $defs",
			errEsquemaImposible, ocupante, tipo, nombre)
	}
}

// representacionEnElDocumento describe los tipos cuya forma en el documento no
// es la de su tipo Go. Son dos: la duración de --timeout, que se escribe «30s»
// y no el número de nanosegundos con que Go la representa, y la clase de error,
// que es una cadena dentro de un vocabulario cerrado y no cualquier cadena
// (FR-017, contracts/sobre-de-salida.md §6).
func representacionEnElDocumento(tipo reflect.Type) *jsonschema.Schema {
	switch tipo {
	case tipoDeDuracion:
		return &jsonschema.Schema{Type: "string"}
	case tipoDeClase:
		return &jsonschema.Schema{Type: "string", Enum: vocabularioDeClases()}
	}

	return nil
}

// vocabularioDeClases es el `enum` de `clase`, derivado del vocabulario del
// dominio y no de una lista escrita aquí: si el dominio ganara una clase, el
// esquema la admitiría sin que nadie edite este fichero (FR-048).
func vocabularioDeClases() []any {
	clases := schema.Clases()
	valores := make([]any, 0, len(clases))

	for _, clase := range clases {
		valores = append(valores, string(clase))
	}

	return valores
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

// sangradoDelDocumento es el de cada nivel del documento de --describe.
const sangradoDelDocumento = "  "

// serializar escribe el documento sin escapar caracteres HTML —la misma regla que
// el sobre— y con sangrado, porque lo lee tanto una persona como un agente. El
// salto final lo pone quien escribe, no este texto.
func serializar(esquema *jsonschema.Schema) (string, error) {
	return codificar(esquema, sangradoDelDocumento)
}

// codificar es la única escritura de un esquema, con el sangrado que pida quien
// llama: el del documento de --describe o ninguno, que es el JSON compacto de los
// esquemas de una herramienta (herramienta.go).
func codificar(esquema *jsonschema.Schema, sangrado string) (string, error) {
	var documento strings.Builder

	codificador := json.NewEncoder(&documento)
	codificador.SetEscapeHTML(false)
	codificador.SetIndent("", sangrado)

	if err := codificador.Encode(esquema); err != nil {
		return "", fmt.Errorf("%w: %w", errEsquemaImposible, err)
	}

	return strings.TrimRight(documento.String(), "\n"), nil
}
