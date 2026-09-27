package skills

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"go.yaml.in/yaml/v3"
)

// delimitadorDelFrontmatter es la línea que abre y la que cierra el frontmatter
// de SKILL.md (data-model §1.1).
const delimitadorDelFrontmatter = "---"

// Las claves de primer nivel que admite el estándar Agent Skills (data-model
// §1.1; research.md D3 y V15).
const (
	claveName          = "name"
	claveDescription   = "description"
	claveLicense       = "license"
	claveAllowedTools  = "allowed-tools"
	claveCompatibility = "compatibility"
	claveMetadata      = "metadata"
)

// clavesAdmitidas son las claves de primer nivel del estándar, en el orden de la
// tabla de data-model §1.1.
var clavesAdmitidas = []string{
	claveName, claveDescription, claveLicense, claveAllowedTools, claveCompatibility, claveMetadata,
}

// Las claves de metadata con las que una skill declara lo que la sincronización
// genera para ella (data-model §1.3; research.md D4).
const (
	claveKitlegalApplets     = "kitlegal-applets"
	claveKitlegalReferencias = "kitlegal-referencias"

	// separadorDeNombres separa los nombres de applet o de referencia dentro del
	// valor de cada una.
	separadorDeNombres = " "
)

// Los límites del estándar, en caracteres y no en bytes (research.md D3).
const (
	maximoDelNombre          = 64
	maximoDeLaDescripcion    = 1024
	maximoDeLaCompatibilidad = 500
)

// carpetaDeLosDatos es el directorio de los YAML de datos, relativo a la raíz
// del repositorio: donde la validación de las referencias busca el que la tabla
// de generadores declara para cada una (data-model §1.3; research.md D20).
const carpetaDeLosDatos = "data"

// etiquetaDeCadena es la etiqueta de YAML de un escalar de texto.
const etiquetaDeCadena = "!!str"

// Los defectos de un contenido que no tiene el bloque del frontmatter.
var (
	errSinFrontmatter = errors.New("sin frontmatter: la primera línea no es " + delimitadorDelFrontmatter)
	errSinCierre      = errors.New("frontmatter sin cierre: ninguna línea " + delimitadorDelFrontmatter +
		" después de la primera")
)

// Frontmatter es el frontmatter de un SKILL.md leído (data-model §1.1): lo que
// las reglas y la sincronización usan de él. license y allowed-tools solo se
// leen para comprobar su tipo, porque ninguna regla ni ninguna generación usa su
// valor.
type Frontmatter struct {
	// Claves son las claves de primer nivel en el orden del documento, las del
	// estándar y las que no lo son: con ellas se sabe si name y description
	// están y qué clave no se admite.
	Claves []string

	// Name es el valor de name, o vacío si no está.
	Name string

	// Description es el valor de description, o vacío si no está.
	Description string

	// Compatibility es el valor de compatibility, o vacío si no está.
	Compatibility string

	// Metadata es el mapa de metadata, o nil si no está.
	Metadata map[string]string
}

// DeclaracionDeKitlegal es lo que una skill declara en metadata para que la
// sincronización le genere tabla y referencias (data-model §1.3; research.md D4).
type DeclaracionDeKitlegal struct {
	// Applets son los nombres de kitlegal-applets, en el orden en que la tabla
	// los presenta, o nil si la skill no la declara.
	Applets []string

	// Referencias son los nombres de kitlegal-referencias, o nil si la skill no
	// la declara. Cada nombre es el de una fila de la tabla de generadores, que
	// declara de qué YAML de datos sale la referencia.
	Referencias []string
}

// DeclaracionDeKitlegal es la declaración de kitlegal del frontmatter: cada
// clave declarada, partida por el espacio que separa sus nombres, sin descartar
// ninguno, también el vacío que dejan dos espacios seguidos.
func (f Frontmatter) DeclaracionDeKitlegal() DeclaracionDeKitlegal {
	return DeclaracionDeKitlegal{
		Applets:     nombresDeclarados(f.Metadata, claveKitlegalApplets),
		Referencias: nombresDeclarados(f.Metadata, claveKitlegalReferencias),
	}
}

// nombresDeclarados son los nombres del valor de una clave de metadata, o nil si
// la clave no está.
func nombresDeclarados(metadata map[string]string, clave string) []string {
	valor, declarada := metadata[clave]
	if !declarada {
		return nil
	}

	return strings.Split(valor, separadorDeNombres)
}

// LeerFrontmatter lee el frontmatter del contenido de un SKILL.md (data-model
// §1.1 y §1.3):
//
//  1. el frontmatter es el bloque YAML entre la primera línea, que tiene que ser
//     ---, y la siguiente línea ---; sin la primera o sin la segunda, es un
//     defecto;
//  2. el bloque, con su línea de apertura, se lee con el lector común,
//     ValidarDocumentoYAML: --- abre el único documento YAML del bloque, así que
//     cada línea que nombra un defecto es la de SKILL.md, y toda clave repetida
//     es una ClaveRepetida con sus dos líneas;
//  3. el documento es un mapa; una clave que repite otra a través de un alias,
//     que el lector común compara como la escribe el documento, también es una
//     ClaveRepetida, en el primer nivel y dentro de metadata;
//  4. name, description, license y compatibility son cadenas; allowed-tools,
//     una cadena o una lista de cadenas; y metadata, un mapa de cadena a cadena:
//     cada valor que no lo es, un DefectoEnElDocumento con su ruta y su línea.
//
// Una clave que no es del estándar no es un defecto de la lectura: va a Claves y
// la rechaza ValidarFrontmatter. Los defectos de los pasos 3 y 4 van todos, en el
// orden del documento y unidos con errors.Join, y con ellos no se devuelve
// ningún frontmatter.
func LeerFrontmatter(contenido []byte) (Frontmatter, error) {
	bloque, err := bloqueDelFrontmatter(contenido)
	if err != nil {
		return Frontmatter{}, err
	}

	documento, err := ValidarDocumentoYAML[yaml.Node](bloque, nil)
	if err != nil {
		return Frontmatter{}, err
	}

	var lectura lecturaDelFrontmatter

	frontmatter := lectura.frontmatter(&documento)
	if err := errors.Join(lectura.defectos...); err != nil {
		return Frontmatter{}, err
	}

	return frontmatter, nil
}

// bloqueDelFrontmatter es el principio del contenido hasta la línea que cierra
// el frontmatter, sin ella: la línea que lo abre y las del bloque. Una línea es
// delimitador solo si es exactamente ---.
func bloqueDelFrontmatter(contenido []byte) ([]byte, error) {
	leido := 0

	for linea := range bytes.Lines(contenido) {
		esDelimitador := string(bytes.TrimSuffix(linea, []byte("\n"))) == delimitadorDelFrontmatter

		switch {
		case leido == 0 && !esDelimitador:
			return nil, errSinFrontmatter
		case leido > 0 && esDelimitador:
			return contenido[:leido], nil
		}

		leido += len(linea)
	}

	if leido == 0 {
		return nil, errSinFrontmatter
	}

	return nil, errSinCierre
}

// lecturaDelFrontmatter lee, clave a clave, el documento del frontmatter que ya
// leyó el lector común, y acumula sus defectos en el orden del documento.
type lecturaDelFrontmatter struct {
	defectos []error
}

// claveDelMapa es una clave de un mapa del documento, con su texto resuelto si
// está escrita con un alias, y su valor.
type claveDelMapa struct {
	texto string
	valor *yaml.Node
}

// frontmatter lee las claves del primer nivel del documento.
func (l *lecturaDelFrontmatter) frontmatter(documento *yaml.Node) Frontmatter {
	mapa := sinEnvoltorio(documento)
	if mapa.Kind != yaml.MappingNode {
		l.defectos = append(l.defectos, &DefectoEnElDocumento{Linea: mapa.Line, Motivo: "el frontmatter no es un mapa"})

		return Frontmatter{}
	}

	var frontmatter Frontmatter

	for _, clave := range l.claves(mapa, nil) {
		frontmatter.Claves = append(frontmatter.Claves, clave.texto)
		ruta := []string{clave.texto}

		switch clave.texto {
		case claveName:
			frontmatter.Name, _ = l.cadena(ruta, clave.valor)
		case claveDescription:
			frontmatter.Description, _ = l.cadena(ruta, clave.valor)
		case claveLicense:
			l.cadena(ruta, clave.valor)
		case claveAllowedTools:
			l.cadenaOListaDeCadenas(ruta, clave.valor)
		case claveCompatibility:
			frontmatter.Compatibility, _ = l.cadena(ruta, clave.valor)
		case claveMetadata:
			frontmatter.Metadata = l.metadata(ruta, clave.valor)
		default:
			// Una clave que no es del estándar la rechaza ValidarFrontmatter.
		}
	}

	return frontmatter
}

// claves devuelve las claves del mapa en su orden, cada una con su texto
// resuelto, y apunta como ClaveRepetida cada una cuyo texto ya había aparecido,
// que no entra en la lista. El lector común ya rechazó las que se repiten
// escritas igual; aquí quedan las que se repiten a través de un alias, que al
// convertir el documento habrían sustituido a la primera sin decir nada.
func (l *lecturaDelFrontmatter) claves(mapa *yaml.Node, ruta []string) []claveDelMapa {
	lineas := make(map[string]int, len(mapa.Content)/2)
	claves := make([]claveDelMapa, 0, len(mapa.Content)/2)

	for indice := 0; indice+1 < len(mapa.Content); indice += 2 {
		nodo := mapa.Content[indice]
		texto := sinEnvoltorio(nodo).Value

		if primera, repetida := lineas[texto]; repetida {
			l.defectos = append(l.defectos, &ClaveRepetida{Mapa: ruta, Clave: texto, Lineas: [2]int{primera, nodo.Line}})

			continue
		}

		lineas[texto] = nodo.Line
		claves = append(claves, claveDelMapa{texto: texto, valor: mapa.Content[indice+1]})
	}

	return claves
}

// cadena devuelve el texto del valor si es una cadena y, si no, apunta el
// defecto con la ruta y la línea del valor tal como lo escribe el documento.
func (l *lecturaDelFrontmatter) cadena(ruta []string, valor *yaml.Node) (string, bool) {
	if !esCadena(valor) {
		l.defectoEn(ruta, valor, "no es una cadena")

		return "", false
	}

	return sinEnvoltorio(valor).Value, true
}

// cadenaOListaDeCadenas comprueba que el valor es una cadena o una lista de
// cadenas y apunta el defecto de cada elemento de la lista que no lo es.
func (l *lecturaDelFrontmatter) cadenaOListaDeCadenas(ruta []string, valor *yaml.Node) {
	lista := sinEnvoltorio(valor)
	if lista.Kind != yaml.SequenceNode {
		if !esCadena(valor) {
			l.defectoEn(ruta, valor, "no es una cadena ni una lista de cadenas")
		}

		return
	}

	for indice, elemento := range lista.Content {
		l.cadena(rutaHija(ruta, strconv.Itoa(indice)), elemento)
	}
}

// metadata devuelve el mapa de cadena a cadena del valor, o nil si no es un
// mapa, y apunta el defecto de cada valor que no es una cadena.
func (l *lecturaDelFrontmatter) metadata(ruta []string, valor *yaml.Node) map[string]string {
	mapa := sinEnvoltorio(valor)
	if mapa.Kind != yaml.MappingNode {
		l.defectoEn(ruta, valor, "no es un mapa de cadena a cadena")

		return nil
	}

	claves := l.claves(mapa, ruta)
	metadata := make(map[string]string, len(claves))

	for _, clave := range claves {
		if texto, esTexto := l.cadena(rutaHija(ruta, clave.texto), clave.valor); esTexto {
			metadata[clave.texto] = texto
		}
	}

	return metadata
}

// defectoEn apunta un DefectoEnElDocumento en la ruta, con la línea del nodo.
func (l *lecturaDelFrontmatter) defectoEn(ruta []string, nodo *yaml.Node, motivo string) {
	l.defectos = append(l.defectos, &DefectoEnElDocumento{Ruta: ruta, Linea: nodo.Line, Motivo: motivo})
}

// esCadena dice si el nodo, o aquel al que apunta si es un alias, es un escalar
// de texto: con yaml.v3, un número o un booleano se decodifican sin error en un
// string, así que el tipo se mira en la etiqueta y no al decodificar.
func esCadena(nodo *yaml.Node) bool {
	escalar := sinEnvoltorio(nodo)

	return escalar.Kind == yaml.ScalarNode && escalar.ShortTag() == etiquetaDeCadena
}

// ValidarFrontmatter aplica al frontmatter de la skill las reglas de data-model
// §1.1 y §1.3 (research.md D3 y D4; FR-040) y devuelve un DefectoDeSkill por
// cada incumplimiento, en este orden:
//
//   - name está, no está vacío, solo tiene a-z, 0-9 y guiones, sin guion al
//     principio ni al final ni dos seguidos —juntas, la forma
//     ^[a-z0-9]+(-[a-z0-9]+)*$—, tiene hasta 64 caracteres y es igual al nombre
//     del directorio de la skill;
//   - description está, no está vacía tras recortar espacios, tiene hasta 1024
//     caracteres y no tiene < ni >;
//   - compatibility tiene hasta 500 caracteres;
//   - cada clave de primer nivel es del estándar;
//   - cada applet de kitlegal-applets está entre los registrados y no se repite;
//   - cada referencia de kitlegal-referencias tiene fila en la tabla de
//     generadores —un YAML de datos de su nombre no se la da—, y el YAML de
//     datos que esa fila declara, que no tiene por qué llamarse como ella, es un
//     fichero regular del directorio de datos del repositorio de raíz
//     (research.md D20).
//
// Los caracteres se cuentan como runas, no como bytes. El error queda para un
// directorio de datos que existe y no se puede listar, que solo se lista si la
// skill declara referencias, y va sin ningún defecto.
func ValidarFrontmatter(raiz, skill string, frontmatter Frontmatter, applets []string) ([]*DefectoDeSkill, error) {
	deKitlegal := frontmatter.DeclaracionDeKitlegal()

	deLasReferencias, err := defectosDeLasReferencias(raiz, deKitlegal.Referencias)
	if err != nil {
		return nil, err
	}

	textos := slices.Concat(
		defectosDelNombre(frontmatter, skill),
		defectosDeLaDescripcion(frontmatter),
		defectosDeLaCompatibilidad(frontmatter),
		clavesNoAdmitidas(frontmatter),
		defectosDeLosApplets(deKitlegal.Applets, applets),
		deLasReferencias,
	)

	defectos := make([]*DefectoDeSkill, 0, len(textos))
	for _, texto := range textos {
		defectos = append(defectos, &DefectoDeSkill{Skill: skill, Defecto: texto})
	}

	return defectos, nil
}

// defectosDelNombre son los defectos de name: cada regla que incumple.
func defectosDelNombre(frontmatter Frontmatter, skill string) []string {
	if !slices.Contains(frontmatter.Claves, claveName) {
		return []string{"falta " + claveName}
	}

	nombre := frontmatter.Name

	var defectos []string

	if nombre == "" {
		defectos = append(defectos, claveName+" vacío")
	}

	if strings.ContainsFunc(nombre, noAdmitidaEnElNombre) {
		defectos = append(defectos, fmt.Sprintf("%s %q con caracteres que no son a-z, 0-9 ni -", claveName, nombre))
	}

	if strings.HasPrefix(nombre, "-") || strings.HasSuffix(nombre, "-") || strings.Contains(nombre, "--") {
		defectos = append(defectos, fmt.Sprintf("%s %q con un guion al principio, al final o dos seguidos", claveName, nombre))
	}

	if longitud := utf8.RuneCountInString(nombre); longitud > maximoDelNombre {
		defectos = append(defectos, demasiadoLargo(claveName, longitud, maximoDelNombre))
	}

	if nombre != skill {
		defectos = append(defectos, fmt.Sprintf("%s %q distinto del nombre del directorio", claveName, nombre))
	}

	return defectos
}

// noAdmitidaEnElNombre dice si el carácter no es uno de los que admite el
// estándar en name: a-z, 0-9 y el guion.
func noAdmitidaEnElNombre(caracter rune) bool {
	return (caracter < 'a' || caracter > 'z') && (caracter < '0' || caracter > '9') && caracter != '-'
}

// defectosDeLaDescripcion son los defectos de description: cada regla que
// incumple.
func defectosDeLaDescripcion(frontmatter Frontmatter) []string {
	if !slices.Contains(frontmatter.Claves, claveDescription) {
		return []string{"falta " + claveDescription}
	}

	descripcion := frontmatter.Description

	var defectos []string

	if strings.TrimSpace(descripcion) == "" {
		defectos = append(defectos, claveDescription+" vacía")
	}

	if longitud := utf8.RuneCountInString(descripcion); longitud > maximoDeLaDescripcion {
		defectos = append(defectos, demasiadoLargo(claveDescription, longitud, maximoDeLaDescripcion))
	}

	if strings.ContainsAny(descripcion, "<>") {
		defectos = append(defectos, claveDescription+" con < o >")
	}

	return defectos
}

// defectosDeLaCompatibilidad es el defecto de compatibility, si es más larga de
// lo que admite el estándar.
func defectosDeLaCompatibilidad(frontmatter Frontmatter) []string {
	if longitud := utf8.RuneCountInString(frontmatter.Compatibility); longitud > maximoDeLaCompatibilidad {
		return []string{demasiadoLargo(claveCompatibility, longitud, maximoDeLaCompatibilidad)}
	}

	return nil
}

// demasiadoLargo es el defecto de un valor con más caracteres que su máximo,
// como «name de 65 caracteres (máximo 64)».
func demasiadoLargo(clave string, longitud, maximo int) string {
	return fmt.Sprintf("%s de %d caracteres (máximo %d)", clave, longitud, maximo)
}

// clavesNoAdmitidas es un defecto por cada clave de primer nivel que no es del
// estándar, en el orden del documento.
func clavesNoAdmitidas(frontmatter Frontmatter) []string {
	var defectos []string

	for _, clave := range frontmatter.Claves {
		if !slices.Contains(clavesAdmitidas, clave) {
			defectos = append(defectos, "clave no admitida: "+clave)
		}
	}

	return defectos
}

// defectosDeLosApplets son los defectos de kitlegal-applets: cada repetición de
// un applet y cada applet que no está entre los registrados, en el orden de la
// declaración y una sola vez por nombre.
func defectosDeLosApplets(declarados, registrados []string) []string {
	ruta := presentarRuta([]string{claveMetadata, claveKitlegalApplets})
	vistos := make(map[string]bool, len(declarados))

	var defectos []string

	for _, applet := range declarados {
		switch {
		case vistos[applet]:
			defectos = append(defectos, fmt.Sprintf("%s: applet %q repetido", ruta, applet))
		case !slices.Contains(registrados, applet):
			defectos = append(defectos, fmt.Sprintf("%s: applet %q no registrado", ruta, applet))
		}

		vistos[applet] = true
	}

	return defectos
}

// defectosDeLasReferencias son los defectos de kitlegal-referencias, en el orden
// de la declaración: uno por cada referencia sin fila en la tabla de generadores
// y uno por cada referencia sin el YAML de datos que su fila declara. Sin
// referencias declaradas no lista el directorio de datos.
func defectosDeLasReferencias(raiz string, declaradas []string) ([]string, error) {
	if len(declaradas) == 0 {
		return nil, nil
	}

	datos, err := yamlDeDatos(raiz)
	if err != nil {
		return nil, err
	}

	ruta := presentarRuta([]string{claveMetadata, claveKitlegalReferencias})

	var defectos []string

	for _, referencia := range declaradas {
		generador, conocido := generadoresDeReferencias[referencia]

		switch {
		case !conocido:
			defectos = append(defectos, fmt.Sprintf("%s: %q sin generador conocido", ruta, referencia))
		case !slices.Contains(datos, generador.datos):
			defectos = append(defectos, fmt.Sprintf("%s: %q sin %s", ruta, referencia,
				carpetaDeLosDatos+"/"+generador.datos))
		}
	}

	return defectos, nil
}

// yamlDeDatos son los nombres de los ficheros regulares del directorio de datos
// del repositorio de raíz. Buscar entre ellos el YAML de datos de una referencia,
// y no componer una ruta con él, hace que solo cuente un fichero de ese
// directorio. Un repositorio sin directorio de datos no tiene ninguno.
func yamlDeDatos(raiz string) ([]string, error) {
	directorio := filepath.Join(raiz, carpetaDeLosDatos)

	entradas, err := os.ReadDir(directorio)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("el directorio de datos %s no se puede listar: %w", directorio, err)
	}

	var ficheros []string

	for _, entrada := range entradas {
		if entrada.Type().IsRegular() {
			ficheros = append(ficheros, entrada.Name())
		}
	}

	return ficheros, nil
}
