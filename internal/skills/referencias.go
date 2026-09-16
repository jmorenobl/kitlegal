package skills

import (
	"bytes"
	"cmp"
	"slices"
	"strings"
)

// Lo que forma references/normas.md, la referencia generada desde
// data/normas.yaml (contrato normas-y-referencias §5; data-model §4.2).
const (
	// nombreDeLasNormas es el nombre de la referencia de las normas y el de su
	// YAML de datos.
	nombreDeLasNormas = "normas"

	// cabeceraDeLasNormas es la primera línea de la referencia: la marca de
	// fichero generado que nombra el YAML de datos del que sale (FR-031). La
	// ruta es texto de la marca y no una ruta del sistema, así que va siempre
	// con barra.
	cabeceraDeLasNormas = "<!-- generado desde " + carpetaDeLosDatos + "/" + nombreDeLasNormas + extensionDeLosDatos +
		", no editar -->"

	// encabezadoDeLasNormas es el título de la referencia.
	encabezadoDeLasNormas = "# Normas de referencia"

	// columnasDeLasNormas son la fila de títulos de la tabla y su fila de
	// separación.
	columnasDeLasNormas = "| Norma | Abreviatura | Identificador | Rango | Materias |\n|---|---|---|---|---|"

	// separadorDeMaterias separa las materias de una norma dentro de su celda.
	separadorDeMaterias = ", "

	// prefijoDelIdentificador precede al año en el identificador
	// BOE-A-<año>-<número> de una norma.
	prefijoDelIdentificador = "BOE-A-"
)

// escapeDeCelda deja el texto de una celda en su celda y en su línea: la barra,
// que la cerraría, se escribe \|, y cada salto de línea —\r\n, \r o \n, los tres
// finales de línea de CommonMark— se sustituye por un espacio. \r\n va antes que
// \r para que cuente como un solo salto.
var escapeDeCelda = strings.NewReplacer("\r\n", " ", "\r", " ", "\n", " ", "|", `\|`)

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
	ordenadas := slices.Clone(normas)
	slices.SortStableFunc(ordenadas, compararNormas)

	var referencia bytes.Buffer

	referencia.WriteString(cabeceraDeLasNormas + "\n\n" + encabezadoDeLasNormas + "\n\n" + columnasDeLasNormas + "\n")

	for _, norma := range ordenadas {
		celdas := []string{
			escapeDeCelda.Replace(norma.Titulo),
			escapeDeCelda.Replace(norma.Abreviatura),
			"`" + escapeDeCelda.Replace(norma.Identificador) + "`",
			escapeDeCelda.Replace(norma.Rango),
			escapeDeCelda.Replace(strings.Join(norma.Materias, separadorDeMaterias)),
		}

		referencia.WriteString("| " + strings.Join(celdas, " | ") + " |\n")
	}

	return referencia.Bytes()
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
