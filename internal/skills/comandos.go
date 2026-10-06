package skills

import (
	"bytes"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// Las dos líneas que delimitan la región generada de SKILL.md, en este orden y
// una sola vez cada una (data-model §2): todo lo que hay entre ellas es la tabla
// de comandos, y el resto de SKILL.md queda byte a byte igual al regenerarla
// (FR-032).
const (
	marcaDeInicioDeLaTabla = "<!-- inicio de la tabla de comandos: generada desde --describe con make skills-sync, " +
		"no editar -->"
	marcaDeFinDeLaTabla = "<!-- fin de la tabla de comandos -->"
)

// Lo que forma la tabla de comandos (contrato sincronizacion-y-comprobacion §3).
const (
	// programaDeLasOrdenes es el binario tal como lo invoca una skill, desde el
	// PATH: cada applet de la tabla se titula `kitlegal <applet>` y cada orden se
	// escribe `kitlegal <applet> <verbo> …` (ADR 0019;
	// contracts/skills-e-invocacion.md §2 de H19; FR-080).
	programaDeLasOrdenes = "kitlegal"

	// separadorDeLaHerramienta une el applet y el verbo en el nombre de la
	// herramienta del servidor MCP de una fila: `<applet>_<verbo>` (H21 FR-003,
	// FR-030).
	separadorDeLaHerramienta = "_"

	// columnasDeLaTabla son la fila de títulos de la tabla de un applet y su fila
	// de separación (contracts/skills.md §4 de H21).
	columnasDeLaTabla = "| Orden | Herramienta | Qué hace | Qué devuelve en `data` |\n|---|---|---|---|"

	// Lo que devuelve un verbo en data, según su forma.
	textoDeObjeto          = "objeto"
	textoDeListaDeObjetos  = "lista de objetos"
	textoSinFormaDeclarada = "sin forma declarada"

	// valorDeUnaBandera es lo que sigue a una bandera que lleva valor.
	valorDeUnaBandera = "<valor>"
)

// Lo que la tabla lee del documento de --describe (data-model §2.1).
const (
	claveDeLaEntrada = "entrada"
	claveDeLaSalida  = "salida"
	claveDeLosDatos  = "data"

	// tipoBooleano es el tipo de una bandera que no lleva valor, y tipoLista el
	// de un argumento que admite varios valores.
	tipoBooleano = "boolean"
	tipoLista    = "array"

	// prefijoDeLasDefiniciones precede al nombre de una definición de $defs en
	// una referencia del mismo documento.
	prefijoDeLasDefiniciones = "#/$defs/"
)

// Las partes del documento de --describe que nombra un defecto de su lectura.
const (
	rutaDeLaEntrada       = "properties." + claveDeLaEntrada
	rutaDeLasBanderas     = rutaDeLaEntrada + ".properties"
	rutaDeLaSalida        = "properties." + claveDeLaSalida
	rutaDelSobre          = rutaDeLaSalida + ".required"
	rutaDeLosDatos        = rutaDeLaSalida + ".then.properties." + claveDeLosDatos
	rutaDeLosElementos    = rutaDeLosDatos + ".items"
	rutaDeLosDatosDeFallo = rutaDeLaSalida + ".else.properties." + claveDeLosDatos
)

// DescripcionDeVerbo es lo que la tabla de comandos presenta de un verbo, leído
// del documento que emite --describe para él (data-model §2.1).
type DescripcionDeVerbo struct {
	// Applet y Verbo son las dos palabras del título del documento.
	Applet, Verbo string

	// Hace es la ayuda del verbo: la descripción del documento.
	Hace string

	// Argumentos son las propiedades de la entrada, en su orden, sin las
	// banderas globales: los que van por su posición y las banderas propias del
	// verbo; nil si no declara ninguno.
	Argumentos []Argumento

	// Banderas son las banderas globales, en el orden del documento de un verbo
	// sin argumentos.
	Banderas []Bandera

	// Devuelve es la forma de data en un sobre correcto.
	Devuelve Devuelve

	// Sobre son las claves del sobre y las de data en un sobre de fallo.
	Sobre Sobre
}

// Argumento es un argumento de un verbo: uno de posición o una bandera propia.
type Argumento struct {
	// Nombre es el de la propiedad de la entrada.
	Nombre string

	// Escritura es cómo se escribe en la orden un argumento que es una bandera
	// propia del verbo, `--roj`: el title de su propiedad. Está vacía en el que
	// va por su posición, que no lo lleva (research.md D16 de H23).
	Escritura string

	// Obligatorio dice si la entrada lo exige: si está en su required.
	Obligatorio bool

	// Varios dice si admite varios valores: si su tipo es array.
	Varios bool
}

// Bandera es una bandera global.
type Bandera struct {
	// Nombre es el de la propiedad de la entrada, sin los guiones con que se
	// escribe.
	Nombre string

	// ConValor dice si la bandera lleva un valor: si su tipo no es boolean.
	ConValor bool
}

// FormaDeLosDatos es la forma que declara el documento para data en un sobre
// correcto.
type FormaDeLosDatos int

const (
	// FormaSinDeclarar es la de data sin $ref, ni en su sitio ni en sus
	// elementos: el documento no dice qué claves lleva.
	FormaSinDeclarar FormaDeLosDatos = iota

	// FormaObjeto es la de data con $ref: un objeto con las claves de esa
	// definición.
	FormaObjeto

	// FormaListaDeObjetos es la de data con items.$ref: una lista de objetos con
	// las claves de esa definición.
	FormaListaDeObjetos
)

// Devuelve es lo que devuelve un verbo en data.
type Devuelve struct {
	// Forma es la forma de data.
	Forma FormaDeLosDatos

	// Claves son las propiedades de la definición de un objeto o de los objetos
	// de una lista, en el orden en que las declara esa definición; nil sin forma
	// declarada o si la definición no declara ninguna.
	Claves []string
}

// Sobre son las claves del sobre de salida de un verbo.
type Sobre struct {
	// Claves son las obligatorias del sobre, en el orden de su required.
	Claves []string

	// ClavesDeFallo son las obligatorias de data en un sobre de fallo, en el
	// orden del required de su definición.
	ClavesDeFallo []string
}

// LeerDescripcionDeVerbo lee la descripción de un verbo del documento que emite
// --describe para él y del documento de un verbo sin argumentos, el de las
// banderas globales (data-model §2.1; research.md D6):
//
//   - el applet y el verbo son las dos palabras del título, separadas por un
//     espacio, y lo que hace, la descripción;
//   - las banderas son las propiedades de la entrada del documento de las
//     banderas globales, en su orden, y llevan valor si su tipo no es boolean;
//   - los argumentos son las propiedades de la entrada del verbo, en su orden,
//     salvo las que se llaman como una bandera; cada uno es obligatorio si está
//     en el required de la entrada, admite varios valores si su tipo es array y
//     es una bandera propia del verbo si su propiedad lleva title, que es cómo
//     se escribe (research.md D16 de H23);
//   - lo que devuelve sale del data de la rama then de la salida: con $ref, un
//     objeto; con items.$ref, una lista de objetos; con las claves de la
//     definición de $defs a la que apunta, en el orden de sus propiedades; y sin
//     $ref, sin forma declarada;
//   - el sobre son las claves del required de la salida y, para el fallo, las del
//     required de la definición a la que apunta el data de la rama else.
//
// Los documentos se leen sin perder el orden de sus propiedades y sin admitir
// una clave repetida en ningún objeto. Es un defecto, sin descripción, un
// documento que no se puede leer, un título que no son dos palabras, la falta de
// la entrada o de la salida, un sobre sin claves, una referencia que no es la de
// una definición de $defs, un data de fallo sin referencia o cuya definición no
// declara ninguna clave obligatoria, y un documento de las banderas globales sin
// entrada o sin ninguna bandera. Cada error nombra el documento —el del verbo por
// su título— y la parte que falla.
func LeerDescripcionDeVerbo(documento, globales []byte) (DescripcionDeVerbo, error) {
	var verbo esquemaDescrito
	if err := json.Unmarshal(documento, &verbo); err != nil {
		return DescripcionDeVerbo{}, fmt.Errorf("documento de --describe: no se puede leer: %w", err)
	}

	banderas, err := leerBanderas(globales)
	if err != nil {
		return DescripcionDeVerbo{}, fmt.Errorf("documento de --describe de las banderas globales: %w", err)
	}

	descripcion, err := describirVerbo(verbo, banderas)
	if err != nil {
		return DescripcionDeVerbo{}, fmt.Errorf("documento de --describe «%s»: %w", verbo.Titulo, err)
	}

	return descripcion, nil
}

// leerBanderas lee las banderas globales del documento de un verbo sin
// argumentos.
func leerBanderas(globales []byte) ([]Bandera, error) {
	var documento esquemaDescrito
	if err := json.Unmarshal(globales, &documento); err != nil {
		return nil, fmt.Errorf("no se puede leer: %w", err)
	}

	entrada := documento.Propiedades.buscar(claveDeLaEntrada)
	if entrada == nil {
		return nil, errors.New("falta " + rutaDeLaEntrada)
	}

	if len(entrada.Propiedades) == 0 {
		return nil, errors.New(rutaDeLasBanderas + " no declara ninguna bandera")
	}

	banderas := make([]Bandera, 0, len(entrada.Propiedades))
	for _, bandera := range entrada.Propiedades {
		banderas = append(banderas, Bandera{Nombre: bandera.nombre, ConValor: bandera.esquema.Tipo != tipoBooleano})
	}

	return banderas, nil
}

// describirVerbo lee del documento de un verbo su descripción, con las banderas
// globales ya leídas.
func describirVerbo(verbo esquemaDescrito, banderas []Bandera) (DescripcionDeVerbo, error) {
	partes := strings.Split(verbo.Titulo, " ")
	if len(partes) != 2 || partes[0] == "" || partes[1] == "" {
		return DescripcionDeVerbo{}, errors.New("el título no tiene la forma «<applet> <verbo>»")
	}

	entrada := verbo.Propiedades.buscar(claveDeLaEntrada)
	if entrada == nil {
		return DescripcionDeVerbo{}, errors.New("falta " + rutaDeLaEntrada)
	}

	salida := verbo.Propiedades.buscar(claveDeLaSalida)
	if salida == nil {
		return DescripcionDeVerbo{}, errors.New("falta " + rutaDeLaSalida)
	}

	devuelve, err := loQueDevuelve(verbo.Definiciones, datosDeLaRama(salida.Entonces))
	if err != nil {
		return DescripcionDeVerbo{}, err
	}

	sobre, err := sobreDeLaSalida(verbo.Definiciones, salida)
	if err != nil {
		return DescripcionDeVerbo{}, err
	}

	return DescripcionDeVerbo{
		Applet:     partes[0],
		Verbo:      partes[1],
		Hace:       verbo.Ayuda,
		Argumentos: argumentosDeLaEntrada(entrada, banderas),
		Banderas:   banderas,
		Devuelve:   devuelve,
		Sobre:      sobre,
	}, nil
}

// argumentosDeLaEntrada son las propiedades de la entrada de un verbo que no se
// llaman como ninguna bandera global, en su orden, cada una con la escritura que
// declara su title.
func argumentosDeLaEntrada(entrada *esquemaDescrito, banderas []Bandera) []Argumento {
	var argumentos []Argumento

	for _, propiedad := range entrada.Propiedades {
		esBandera := slices.ContainsFunc(banderas, func(bandera Bandera) bool {
			return bandera.Nombre == propiedad.nombre
		})
		if esBandera {
			continue
		}

		argumentos = append(argumentos, Argumento{
			Nombre:      propiedad.nombre,
			Escritura:   propiedad.esquema.Titulo,
			Obligatorio: slices.Contains(entrada.Obligatorias, propiedad.nombre),
			Varios:      propiedad.esquema.Tipo == tipoLista,
		})
	}

	return argumentos
}

// loQueDevuelve es la forma de data en un sobre correcto, desde su esquema, o
// sin forma declarada si no lo hay.
func loQueDevuelve(definiciones map[string]*esquemaDescrito, datos *esquemaDescrito) (Devuelve, error) {
	switch {
	case datos == nil:
		return Devuelve{Forma: FormaSinDeclarar}, nil
	case datos.Referencia != "":
		referida, err := definicionReferida(definiciones, rutaDeLosDatos, datos.Referencia)
		if err != nil {
			return Devuelve{}, err
		}

		return Devuelve{Forma: FormaObjeto, Claves: referida.Propiedades.nombres()}, nil
	case datos.Elementos != nil && datos.Elementos.Referencia != "":
		referida, err := definicionReferida(definiciones, rutaDeLosElementos, datos.Elementos.Referencia)
		if err != nil {
			return Devuelve{}, err
		}

		return Devuelve{Forma: FormaListaDeObjetos, Claves: referida.Propiedades.nombres()}, nil
	default:
		return Devuelve{Forma: FormaSinDeclarar}, nil
	}
}

// sobreDeLaSalida son las claves del sobre y las de data en un sobre de fallo.
func sobreDeLaSalida(definiciones map[string]*esquemaDescrito, salida *esquemaDescrito) (Sobre, error) {
	if len(salida.Obligatorias) == 0 {
		return Sobre{}, errors.New(rutaDelSobre + " no declara ninguna clave del sobre")
	}

	fallo := datosDeLaRama(salida.SiNo)
	if fallo == nil || fallo.Referencia == "" {
		return Sobre{}, errors.New(rutaDeLosDatosDeFallo + " no apunta a ninguna definición de $defs")
	}

	referida, err := definicionReferida(definiciones, rutaDeLosDatosDeFallo, fallo.Referencia)
	if err != nil {
		return Sobre{}, err
	}

	if len(referida.Obligatorias) == 0 {
		return Sobre{}, fmt.Errorf("la definición %s de %s no declara ninguna clave obligatoria",
			strings.TrimPrefix(fallo.Referencia, prefijoDeLasDefiniciones), rutaDeLosDatosDeFallo)
	}

	return Sobre{Claves: salida.Obligatorias, ClavesDeFallo: referida.Obligatorias}, nil
}

// datosDeLaRama es el esquema de data de una rama de la condición de la salida,
// o nulo si la rama no está o no lo declara.
func datosDeLaRama(rama *esquemaDescrito) *esquemaDescrito {
	if rama == nil {
		return nil
	}

	return rama.Propiedades.buscar(claveDeLosDatos)
}

// definicionReferida es la definición de $defs a la que apunta la referencia del
// esquema de ruta. Una referencia a otro documento, o a un nombre que $defs no
// tiene, no apunta a ninguna.
func definicionReferida(definiciones map[string]*esquemaDescrito, ruta, referencia string) (*esquemaDescrito, error) {
	nombre, delMismoDocumento := strings.CutPrefix(referencia, prefijoDeLasDefiniciones)

	referida := definiciones[nombre]
	if !delMismoDocumento || referida == nil {
		return nil, fmt.Errorf("%s apunta a %s, que no es ninguna definición de $defs", ruta, referencia)
	}

	return referida, nil
}

// esquemaDescrito es un esquema del documento de --describe con lo único que la
// tabla lee de él; el resto de sus claves se ignora. Un esquema booleano, como
// el data sin restringir de un verbo que no declara su salida, no declara nada
// de eso y se lee vacío.
type esquemaDescrito struct {
	Titulo       string                      `json:"title"`
	Ayuda        string                      `json:"description"`
	Tipo         string                      `json:"type"`
	Referencia   string                      `json:"$ref"`
	Obligatorias []string                    `json:"required"`
	Propiedades  propiedadesEnOrden          `json:"properties"`
	Elementos    *esquemaDescrito            `json:"items"`
	Entonces     *esquemaDescrito            `json:"then"`
	SiNo         *esquemaDescrito            `json:"else"`
	Definiciones map[string]*esquemaDescrito `json:"$defs"`
}

// UnmarshalJSONFrom lee un esquema booleano como un esquema vacío y cualquier
// otro valor como el objeto de un esquema.
func (e *esquemaDescrito) UnmarshalJSONFrom(decodificador *jsontext.Decoder) error {
	if tipo := decodificador.PeekKind(); tipo == jsontext.KindTrue || tipo == jsontext.KindFalse {
		*e = esquemaDescrito{}
		_, err := decodificador.ReadToken()

		return err
	}

	// sinMetodos tiene los campos de esquemaDescrito y no este método, así que
	// leerlo no vuelve a llamarlo.
	type sinMetodos esquemaDescrito

	return json.UnmarshalDecode(decodificador, (*sinMetodos)(e))
}

// propiedad es un miembro de properties: su nombre y su esquema.
type propiedad struct {
	nombre  string
	esquema esquemaDescrito
}

// propiedadesEnOrden son los miembros de properties en el orden del documento,
// que es el que presenta la tabla y el que un mapa perdería.
type propiedadesEnOrden []propiedad

// UnmarshalJSONFrom lee los miembros de un objeto en su orden: en cada vuelta, el
// nombre de un miembro, con su esquema detrás, o el cierre del objeto. El
// decodificador rechaza un nombre repetido y todo lo que no es ni un nombre ni el
// cierre.
func (p *propiedadesEnOrden) UnmarshalJSONFrom(decodificador *jsontext.Decoder) error {
	apertura, err := decodificador.ReadToken()
	if err != nil {
		return err
	}

	if apertura.Kind() != jsontext.KindBeginObject {
		return errors.New("properties no es un objeto")
	}

	var propiedades propiedadesEnOrden

	for {
		nombre, err := decodificador.ReadToken()
		if err != nil {
			return err
		}

		if nombre.Kind() == jsontext.KindEndObject {
			break
		}

		// El nombre deja de ser válido en la lectura siguiente: se copia antes.
		leida := propiedad{nombre: nombre.String()}
		if err := json.UnmarshalDecode(decodificador, &leida.esquema); err != nil {
			return err
		}

		propiedades = append(propiedades, leida)
	}

	*p = propiedades

	return nil
}

// buscar es el esquema de la propiedad que se llama así, o nulo si no la hay.
func (p propiedadesEnOrden) buscar(nombre string) *esquemaDescrito {
	for indice := range p {
		if p[indice].nombre == nombre {
			return &p[indice].esquema
		}
	}

	return nil
}

// nombres son los nombres de las propiedades en su orden, o nil si no hay
// ninguna.
func (p propiedadesEnOrden) nombres() []string {
	var nombres []string

	for _, propiedad := range p {
		nombres = append(nombres, propiedad.nombre)
	}

	return nombres
}

// RenderizarTabla da la región de la tabla de comandos de SKILL.md, lo que va
// entre sus dos marcas, para los applets que la skill declara y las descripciones
// de los verbos del registro (contrato sincronizacion-y-comprobacion §3;
// data-model §2.2; FR-032, FR-034; contracts/skills.md §4 y FR-030 de H21):
//
//   - una sección ### por applet, en el orden de applets, con `kitlegal
//     <applet>` como título y una tabla con una fila por verbo del applet, en el
//     orden de las descripciones; las de un applet que no se declara no se
//     presentan;
//   - cada fila da la sintaxis de la orden —<nombre> por cada argumento de
//     posición, <nombre>... si admite varios valores, y con [ delante si es
//     opcional, con los corchetes cerrados detrás del último, como la ayuda de
//     Kong: [<norma> [<bloques>...]]; y, detrás, [--<nombre>=<nombre>] por cada
//     bandera propia del verbo—, la herramienta del servidor MCP que hace lo mismo,
//     `<applet>_<verbo>`, lo que hace y lo que devuelve en data; en cada celda la
//     barra se escribe \| y un salto de línea es un espacio;
//   - detrás de la última sección, la línea del sobre, que es el mismo por la
//     orden y por la herramienta de cada fila, y la de las banderas comunes, que
//     valen para todas las órdenes de la tabla.
//
// Empieza y termina con una línea en blanco y usa \n como fin de línea; las
// mismas entradas dan siempre los mismos bytes. Es un defecto, sin tabla, no
// declarar ningún applet, declarar uno sin ningún verbo descrito y que dos verbos
// de la tabla declaren banderas globales o sobres distintos: la tabla afirma que
// son los de todas sus órdenes.
func RenderizarTabla(applets []string, descripciones []DescripcionDeVerbo) ([]byte, error) {
	if len(applets) == 0 {
		return nil, errors.New("la tabla de comandos no declara ningún applet")
	}

	secciones := make([][]DescripcionDeVerbo, 0, len(applets))

	for _, applet := range applets {
		verbos := verbosDelApplet(descripciones, applet)
		if len(verbos) == 0 {
			return nil, fmt.Errorf("el applet %s no tiene ningún verbo descrito", applet)
		}

		secciones = append(secciones, verbos)
	}

	comun := secciones[0][0]

	var tabla bytes.Buffer

	for indice, verbos := range secciones {
		tabla.WriteString("\n### `" + programaDeLasOrdenes + " " + applets[indice] + "`\n\n" + columnasDeLaTabla + "\n")

		for _, verbo := range verbos {
			if err := mismoSobreYBanderas(comun, verbo); err != nil {
				return nil, err
			}

			tabla.WriteString(filaDelVerbo(verbo))
		}
	}

	tabla.WriteString("\n" + lineaDelSobre(comun.Sobre) + "\n\n" + lineaDeLasBanderas(comun.Banderas) + "\n\n")

	return tabla.Bytes(), nil
}

// verbosDelApplet son las descripciones de los verbos del applet, en su orden.
func verbosDelApplet(descripciones []DescripcionDeVerbo, applet string) []DescripcionDeVerbo {
	var verbos []DescripcionDeVerbo

	for _, descripcion := range descripciones {
		if descripcion.Applet == applet {
			verbos = append(verbos, descripcion)
		}
	}

	return verbos
}

// mismoSobreYBanderas comprueba que un verbo declara las mismas banderas
// globales y el mismo sobre que el primero de la tabla.
func mismoSobreYBanderas(comun, verbo DescripcionDeVerbo) error {
	if !slices.Equal(comun.Banderas, verbo.Banderas) {
		return fmt.Errorf("%s no declara las mismas banderas globales que %s", nombreDeLaOrden(verbo),
			nombreDeLaOrden(comun))
	}

	if !slices.Equal(comun.Sobre.Claves, verbo.Sobre.Claves) ||
		!slices.Equal(comun.Sobre.ClavesDeFallo, verbo.Sobre.ClavesDeFallo) {
		return fmt.Errorf("%s no declara el mismo sobre que %s", nombreDeLaOrden(verbo), nombreDeLaOrden(comun))
	}

	return nil
}

// nombreDeLaOrden es el applet y el verbo de una descripción, como su título.
func nombreDeLaOrden(descripcion DescripcionDeVerbo) string {
	return descripcion.Applet + " " + descripcion.Verbo
}

// filaDelVerbo es la fila de la tabla de un verbo: su sintaxis, su herramienta,
// lo que hace y lo que devuelve en data, cada celda con su texto escapado.
func filaDelVerbo(verbo DescripcionDeVerbo) string {
	celdas := []string{
		"`" + sintaxisDeLaOrden(verbo) + "`",
		"`" + nombreDeLaHerramienta(verbo) + "`",
		verbo.Hace,
		verbo.Devuelve.texto(),
	}

	for indice, celda := range celdas {
		celdas[indice] = escapeDeCelda.Replace(celda)
	}

	return "| " + strings.Join(celdas, " | ") + " |\n"
}

// sintaxisDeLaOrden es la orden de un verbo: kitlegal, su applet y el verbo, con
// cada argumento de posición en su orden y, detrás, cada bandera propia en el
// suyo (data-model §2.2; research.md D7; research.md D16 de H23). Un argumento
// de posición opcional abre un corchete que se cierra detrás del último, así que
// los opcionales quedan anidados como en la ayuda de Kong: `[<norma>
// [<bloques>...]]`; Kong no admite un obligatorio detrás de un opcional. Una
// bandera propia se escribe con su escritura y su nombre como valor,
// `[--roj=<roj>]`: no hay otra forma para la que no lleva valor ni para la que
// se repite, que ningún verbo de una tabla tiene.
func sintaxisDeLaOrden(verbo DescripcionDeVerbo) string {
	partes := []string{programaDeLasOrdenes, verbo.Applet, verbo.Verbo}
	abiertos := 0

	var banderas []string

	for _, argumento := range verbo.Argumentos {
		if argumento.Escritura != "" {
			banderas = append(banderas, "["+argumento.Escritura+"=<"+argumento.Nombre+">]")

			continue
		}

		parte := "<" + argumento.Nombre + ">"
		if argumento.Varios {
			parte += "..."
		}

		if !argumento.Obligatorio {
			parte = "[" + parte
			abiertos++
		}

		partes = append(partes, parte)
	}

	dePosicion := strings.Join(partes, " ") + strings.Repeat("]", abiertos)

	return strings.Join(slices.Concat([]string{dePosicion}, banderas), " ")
}

// nombreDeLaHerramienta es el de la herramienta del servidor MCP de un verbo:
// su applet y su verbo, los mismos de la orden, unidos por un guion bajo (H21
// FR-003). Que el servidor la anuncia lo comprueba
// TestOrdenesDeLasSkillsEmpotradas contra el registro de producción (H21
// FR-031).
func nombreDeLaHerramienta(verbo DescripcionDeVerbo) string {
	return verbo.Applet + separadorDeLaHerramienta + verbo.Verbo
}

// texto es lo que devuelve un verbo como lo escribe la tabla: la forma y, si la
// definición declara claves, «con» y sus claves separadas por «, ».
func (d Devuelve) texto() string {
	var forma string

	switch d.Forma {
	case FormaObjeto:
		forma = textoDeObjeto
	case FormaListaDeObjetos:
		forma = textoDeListaDeObjetos
	default:
		return textoSinFormaDeclarada
	}

	if len(d.Claves) == 0 {
		return forma
	}

	return forma + " con " + strings.Join(comoCodigo(d.Claves), ", ")
}

// lineaDelSobre es la línea que da las claves del sobre y las de data en un
// sobre de fallo, y que dice que la orden y la herramienta de cada fila
// devuelven el mismo (H21 FR-030).
func lineaDelSobre(sobre Sobre) string {
	return "La orden y la herramienta de cada fila devuelven el mismo sobre: " +
		strings.Join(comoCodigo(sobre.Claves), ", ") +
		"; con `ok` falso, `" + claveDeLosDatos + "` lleva " + enumerar(comoCodigo(sobre.ClavesDeFallo)) + "."
}

// lineaDeLasBanderas es la línea de las banderas comunes a todas las órdenes,
// cada una con --, y con <valor> detrás la que lleva valor.
func lineaDeLasBanderas(banderas []Bandera) string {
	escritas := make([]string, 0, len(banderas))

	for _, bandera := range banderas {
		escrita := "--" + bandera.Nombre
		if bandera.ConValor {
			escrita += " " + valorDeUnaBandera
		}

		escritas = append(escritas, escrita)
	}

	return "Banderas comunes: " + strings.Join(comoCodigo(escritas), ", ") + "."
}

// comoCodigo escribe cada texto como código de Markdown.
func comoCodigo(textos []string) []string {
	codigos := make([]string, 0, len(textos))

	for _, texto := range textos {
		codigos = append(codigos, "`"+texto+"`")
	}

	return codigos
}

// enumerar une los textos como una enumeración: separados por «, » salvo el
// último, que va detrás de « y ».
func enumerar(textos []string) string {
	if len(textos) < 2 {
		return strings.Join(textos, "")
	}

	return strings.Join(textos[:len(textos)-1], ", ") + " y " + textos[len(textos)-1]
}

// lineaDeMarca es una línea del contenido de SKILL.md que es exactamente una de
// las dos marcas de la región.
type lineaDeMarca struct {
	// numero es el de la línea, desde 1.
	numero int

	// comienzo es la posición de su primer byte, y final la que sigue a su salto
	// de línea o, en la última línea sin salto, el final del contenido.
	comienzo, final int
}

// SustituirRegion devuelve el contenido de un SKILL.md con region en lugar de lo
// que hay entre sus dos marcas (data-model §2; FR-032): las dos líneas de las
// marcas y todo lo que hay antes y después de ellas quedan byte a byte igual. Es
// un defecto, sin contenido, que falten las dos marcas o una de ellas, que una
// esté en más de una línea o que el fin vaya antes del inicio. Una línea es una
// marca solo si es exactamente la marca, sin espacios ni retorno de carro.
func SustituirRegion(contenido, region []byte) ([]byte, error) {
	inicios, fines := lineasDeLasMarcas(contenido)

	if err := comprobarMarcas(inicios, fines); err != nil {
		return nil, err
	}

	inicio, fin := inicios[0], fines[0]

	sustituido := make([]byte, 0, inicio.final+len(region)+len(contenido)-fin.comienzo)
	sustituido = append(sustituido, contenido[:inicio.final]...)
	sustituido = append(sustituido, region...)
	sustituido = append(sustituido, contenido[fin.comienzo:]...)

	return sustituido, nil
}

// lineasDeLasMarcas son las líneas del contenido que son la marca de inicio y
// las que son la de fin, cada una en su orden.
func lineasDeLasMarcas(contenido []byte) (inicios, fines []lineaDeMarca) {
	numero, comienzo := 0, 0

	for linea := range bytes.Lines(contenido) {
		numero++
		marca := lineaDeMarca{numero: numero, comienzo: comienzo, final: comienzo + len(linea)}

		switch string(bytes.TrimSuffix(linea, []byte("\n"))) {
		case marcaDeInicioDeLaTabla:
			inicios = append(inicios, marca)
		case marcaDeFinDeLaTabla:
			fines = append(fines, marca)
		}

		comienzo = marca.final
	}

	return inicios, fines
}

// comprobarMarcas da el primer defecto de las marcas: sin las dos, sin una, una
// en más de una línea o el fin antes del inicio.
func comprobarMarcas(inicios, fines []lineaDeMarca) error {
	switch {
	case len(inicios) == 0 && len(fines) == 0:
		return errors.New("sin las marcas de la tabla de comandos")
	case len(inicios) == 0:
		return errors.New("sin la marca de inicio de la tabla de comandos")
	case len(fines) == 0:
		return errors.New("sin la marca de fin de la tabla de comandos")
	case len(inicios) > 1:
		return marcaRepetida("inicio", inicios)
	case len(fines) > 1:
		return marcaRepetida("fin", fines)
	case fines[0].numero < inicios[0].numero:
		return fmt.Errorf("la marca de fin de la tabla de comandos, en la línea %d, está antes que la de inicio, "+
			"en la línea %d", fines[0].numero, inicios[0].numero)
	default:
		return nil
	}
}

// marcaRepetida es el defecto de una marca que está en más de una línea,
// nombrando cada una.
func marcaRepetida(marca string, lineas []lineaDeMarca) error {
	numeros := make([]string, 0, len(lineas))

	for _, linea := range lineas {
		numeros = append(numeros, strconv.Itoa(linea.numero))
	}

	return fmt.Errorf("la marca de %s de la tabla de comandos aparece %d veces, en las líneas %s", marca,
		len(lineas), enumerar(numeros))
}
