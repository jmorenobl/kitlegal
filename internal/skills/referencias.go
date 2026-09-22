package skills

import (
	"bytes"
	"cmp"
	"slices"
	"strings"
)

// Lo que forma una celda de la tabla de una referencia generada (contrato
// normas-y-referencias §5; contrato de la skill legal-core §2).
const (
	// separadorDeMaterias separa las materias de una norma dentro de su celda, y
	// los tipos de norma de un nivel dentro de la suya.
	separadorDeMaterias = ", "

	// prefijoDelIdentificador precede al año en el identificador
	// BOE-A-<año>-<número> de una norma.
	prefijoDelIdentificador = "BOE-A-"
)

// tituloDeLasReglas es el título de la sección de las reglas de interpretación,
// detrás de la tabla de niveles de la referencia de la jerarquía.
const tituloDeLasReglas = "## Reglas de interpretación"

// escapeDeCelda deja el texto de una celda en su celda y en su línea: la barra,
// que la cerraría, se escribe \|, y cada salto de línea —\r\n, \r o \n, los tres
// finales de línea de CommonMark— se sustituye por un espacio. \r\n va antes que
// \r para que cuente como un solo salto.
var escapeDeCelda = strings.NewReplacer("\r\n", " ", "\r", " ", "\n", " ", "|", `\|`)

// enUnaLinea deja en su línea el texto de un elemento de una lista: cada salto
// de línea es un espacio, como en una celda, y la barra no cierra nada.
var enUnaLinea = strings.NewReplacer("\r\n", " ", "\r", " ", "\n", " ")

// generadorDeReferencia es una fila de la tabla de generadores (research.md D20;
// contrato de la skill legal-core §4): lo que declara de una referencia
// generada —el YAML de datos del que sale, su título y los encabezados de su
// tabla— y la función que da el resto desde el contenido de ese YAML. La
// cabecera, el título y la fila de títulos de la tabla salen de la fila y no de
// constantes de una referencia concreta.
type generadorDeReferencia struct {
	// datos es el nombre del YAML de datos dentro de data/, que no tiene por qué
	// ser el de la referencia.
	datos string

	// titulo es el título de la referencia.
	titulo string

	// encabezados son los títulos de cada columna de su tabla.
	encabezados []string

	// cuerpo da, desde el contenido del YAML de datos, lo que sigue a la fila de
	// separación de la tabla —sus filas y lo que vaya detrás—, o los defectos del
	// YAML, que no se puede leer.
	cuerpo func(datos []byte) ([]byte, error)
}

// Las filas de la tabla de generadores, una por referencia.
var (
	// generadorDeLasNormas da references/normas.md: todas las normas de
	// data/normas.yaml.
	generadorDeLasNormas = generadorDeReferencia{
		datos:       "normas.yaml",
		titulo:      "Normas de referencia",
		encabezados: columnasDeUnaNorma,
		cuerpo:      cuerpoDeLasNormas,
	}

	// generadorDeLasLeyesVertebrales da references/leyes_vertebrales.md: solo las
	// normas de data/normas.yaml marcadas vertebral, con los mismos encabezados.
	generadorDeLasLeyesVertebrales = generadorDeReferencia{
		datos:       "normas.yaml",
		titulo:      "Leyes vertebrales",
		encabezados: columnasDeUnaNorma,
		cuerpo:      cuerpoDeLasLeyesVertebrales,
	}

	// generadorDeLaJerarquia da references/jerarquia_normativa.md: los niveles y
	// las reglas de interpretación de data/jerarquia.yaml.
	generadorDeLaJerarquia = generadorDeReferencia{
		datos:       "jerarquia.yaml",
		titulo:      "Jerarquía normativa",
		encabezados: []string{"Nivel", "Boletín", "Tipos de norma, de mayor a menor rango"},
		cuerpo:      cuerpoDeLaJerarquia,
	}
)

// columnasDeUnaNorma son los encabezados de las referencias de normas: título,
// abreviatura, identificador, rango y materias.
var columnasDeUnaNorma = []string{"Norma", "Abreviatura", "Identificador", "Rango", "Materias"}

// generadoresDeReferencias es la tabla de generadores (research.md D20; contrato
// de la skill legal-core §4): la fila de cada referencia por su nombre, que es el
// que declara kitlegal-referencias y el de references/<nombre>.md. Una referencia
// sin fila no tiene generador, y dos pueden salir del mismo YAML de datos.
var generadoresDeReferencias = map[string]generadorDeReferencia{
	"normas":              generadorDeLasNormas,
	"leyes_vertebrales":   generadorDeLasLeyesVertebrales,
	"jerarquia_normativa": generadorDeLaJerarquia,
}

// cabecera es la primera línea de la referencia: la marca de fichero generado
// que nombra el YAML de datos del que sale (FR-031, FR-065, FR-066). La ruta es
// texto de la marca y no una ruta del sistema, así que va siempre con barra.
func (g generadorDeReferencia) cabecera() string {
	return "<!-- generado desde " + carpetaDeLosDatos + "/" + g.datos + ", no editar -->"
}

// componer es la referencia con el cuerpo dado: la cabecera, el título, la fila
// de títulos de la tabla y su fila de separación, y detrás el cuerpo, con \n como
// fin de línea.
func (g generadorDeReferencia) componer(cuerpo []byte) []byte {
	var referencia bytes.Buffer

	referencia.WriteString(g.cabecera() + "\n\n# " + g.titulo + "\n\n")
	referencia.WriteString(filaDeLaTabla(g.encabezados))
	referencia.WriteString("|" + strings.Repeat("---|", len(g.encabezados)) + "\n")
	referencia.Write(cuerpo)

	return referencia.Bytes()
}

// generar es la referencia desde el contenido de su YAML de datos, o los
// defectos que impiden leerlo.
func (g generadorDeReferencia) generar(datos []byte) ([]byte, error) {
	cuerpo, err := g.cuerpo(datos)
	if err != nil {
		return nil, err
	}

	return g.componer(cuerpo), nil
}

// filaDeLaTabla es la fila de una tabla con las celdas dadas, ya escapadas, y su
// salto de línea.
func filaDeLaTabla(celdas []string) string {
	return "| " + strings.Join(celdas, " | ") + " |\n"
}

// RenderizarNormas da references/normas.md desde las normas de data/normas.yaml
// (contrato normas-y-referencias §5; FR-030, FR-031, FR-034): la cabecera de
// fichero generado que nombra data/normas.yaml, el título y una tabla con una
// fila por norma —título, abreviatura, identificador, rango y materias—, con \n
// como fin de línea y un salto final.
//
// Las filas van por el año del identificador y después por su número, los dos
// por su valor, y a igual valor por el identificador tal cual, así que el orden
// en que llegan las normas no cambia la salida. Las materias van separadas por
// «, », una abreviatura ausente es una celda vacía y en cada celda la barra se
// escribe \| y un salto de línea es un espacio. Las mismas normas dan siempre
// los mismos bytes, y las que recibe no se reordenan.
func RenderizarNormas(normas []Norma) []byte {
	return generadorDeLasNormas.componer(filasDeLasNormas(normas))
}

// RenderizarLeyesVertebrales da references/leyes_vertebrales.md desde las normas
// de data/normas.yaml (contrato de la skill legal-core §2; FR-065, FR-067): como
// RenderizarNormas, con su propio título y solo con las normas marcadas
// vertebral; si no hay ninguna, la tabla va sin filas. Las normas que recibe no
// cambian.
func RenderizarLeyesVertebrales(normas []Norma) []byte {
	return generadorDeLasLeyesVertebrales.componer(filasDeLasNormas(vertebrales(normas)))
}

// RenderizarJerarquia da references/jerarquia_normativa.md desde la jerarquía de
// data/jerarquia.yaml (contrato de la skill legal-core §2; FR-066, FR-067): la
// cabecera de fichero generado que nombra data/jerarquia.yaml, el título, una
// tabla con una fila por nivel —su nombre, la clase de boletín que lo publica y
// sus tipos de norma, de mayor a menor rango y separados por «, »— y la sección
// de las reglas de interpretación, una por línea con su código y su enunciado.
//
// Los niveles y las reglas van en el orden en que llegan, que es el de su
// enumerado: el esquema lo impone al documento. En cada celda la barra se escribe
// \|, y en las celdas y en las reglas un salto de línea es un espacio. La
// jerarquía que recibe no cambia.
func RenderizarJerarquia(jerarquia Jerarquia) []byte {
	return generadorDeLaJerarquia.componer(seccionesDeLaJerarquia(jerarquia))
}

// cuerpoDeLasNormas son las filas de todas las normas del contenido de
// data/normas.yaml.
func cuerpoDeLasNormas(datos []byte) ([]byte, error) {
	normas, err := LeerNormas(datos)
	if err != nil {
		return nil, err
	}

	return filasDeLasNormas(normas), nil
}

// cuerpoDeLasLeyesVertebrales son las filas de las normas marcadas vertebral del
// contenido de data/normas.yaml.
func cuerpoDeLasLeyesVertebrales(datos []byte) ([]byte, error) {
	normas, err := LeerNormas(datos)
	if err != nil {
		return nil, err
	}

	return filasDeLasNormas(vertebrales(normas)), nil
}

// cuerpoDeLaJerarquia son los niveles y las reglas del contenido de
// data/jerarquia.yaml, leído con LeerJerarquia.
func cuerpoDeLaJerarquia(datos []byte) ([]byte, error) {
	jerarquia, err := LeerJerarquia(datos)
	if err != nil {
		return nil, err
	}

	return seccionesDeLaJerarquia(jerarquia), nil
}

// vertebrales son las normas marcadas vertebral, en su orden y en una copia.
func vertebrales(normas []Norma) []Norma {
	return slices.DeleteFunc(slices.Clone(normas), func(norma Norma) bool { return !norma.Vertebral })
}

// filasDeLasNormas es una fila por norma, por año y número del identificador,
// sin reordenar las que recibe.
func filasDeLasNormas(normas []Norma) []byte {
	ordenadas := slices.Clone(normas)
	slices.SortStableFunc(ordenadas, compararNormas)

	var filas bytes.Buffer

	for _, norma := range ordenadas {
		filas.WriteString(filaDeLaTabla([]string{
			escapeDeCelda.Replace(norma.Titulo),
			escapeDeCelda.Replace(norma.Abreviatura),
			"`" + escapeDeCelda.Replace(norma.Identificador) + "`",
			escapeDeCelda.Replace(norma.Rango),
			escapeDeCelda.Replace(strings.Join(norma.Materias, separadorDeMaterias)),
		}))
	}

	return filas.Bytes()
}

// seccionesDeLaJerarquia es una fila por nivel y, detrás, la sección de las
// reglas, con un elemento de lista por regla.
func seccionesDeLaJerarquia(jerarquia Jerarquia) []byte {
	var secciones bytes.Buffer

	for _, nivel := range jerarquia.Niveles {
		secciones.WriteString(filaDeLaTabla([]string{
			escapeDeCelda.Replace(nivel.Nombre),
			escapeDeCelda.Replace(nivel.Boletin),
			escapeDeCelda.Replace(strings.Join(nivel.Normas, separadorDeMaterias)),
		}))
	}

	secciones.WriteString("\n" + tituloDeLasReglas + "\n\n")

	for _, regla := range jerarquia.Reglas {
		secciones.WriteString("- `" + regla.Codigo + "`: " + enUnaLinea.Replace(regla.Enunciado) + "\n")
	}

	return secciones.Bytes()
}

// compararNormas ordena dos normas por el año de su identificador y después por
// su número, los dos por su valor, y a igual valor por el identificador tal
// cual.
func compararNormas(a, b Norma) int {
	anioA, numeroA := partesDelIdentificador(a.Identificador)
	anioB, numeroB := partesDelIdentificador(b.Identificador)

	return cmp.Or(
		compararCifras(anioA, anioB),
		compararCifras(numeroA, numeroB),
		cmp.Compare(a.Identificador, b.Identificador),
	)
}

// partesDelIdentificador separa de un identificador BOE-A-<año>-<número> las
// cifras de su año y las de su número. LeerNormas no deja pasar otra forma; si
// llegara, el orden seguiría siendo total con lo que diera la separación, y el
// identificador tal cual lo desempataría.
func partesDelIdentificador(identificador string) (anio, numero string) {
	anio, numero, _ = strings.Cut(strings.TrimPrefix(identificador, prefijoDelIdentificador), "-")

	return anio, numero
}

// compararCifras compara dos tiras de cifras por su valor sin convertirlas a
// enteros, así que no hay desbordamiento ni error que atender: sin sus ceros a la
// izquierda, la más larga es la mayor y, con la misma longitud, el orden de sus
// caracteres es el de su valor.
func compararCifras(a, b string) int {
	a, b = strings.TrimLeft(a, "0"), strings.TrimLeft(b, "0")

	return cmp.Or(cmp.Compare(len(a), len(b)), cmp.Compare(a, b))
}
