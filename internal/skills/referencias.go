package skills

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strconv"
	"strings"

	"github.com/jmorenobl/kitlegal/internal/core/plazos"
)

// Lo que forma una celda de la tabla de una referencia generada (contrato
// normas-y-referencias §5; contrato skill-legal-core §2).
const (
	// separadorDeMaterias separa las materias de una norma dentro de su celda, y
	// los tipos de norma de un nivel dentro de la suya.
	separadorDeMaterias = ", "

	// prefijoDelIdentificador precede al año en el identificador
	// BOE-A-<año>-<número> de una norma.
	prefijoDelIdentificador = "BOE-A-"

	// separadorDePreceptos separa los preceptos de un tipo de plazo dentro de su
	// celda.
	separadorDePreceptos = "; "
)

// nombresDeLasUnidades da el nombre de cada unidad de plazos.Unidades, en
// singular y en plural, para la duración que fija un tipo de plazo.
var nombresDeLasUnidades = map[plazos.Unidad][2]string{
	plazos.UnidadMeses:         {"mes", "meses"},
	plazos.UnidadDiasHabiles:   {"día hábil", "días hábiles"},
	plazos.UnidadDiasNaturales: {"día natural", "días naturales"},
}

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
// contrato skill-legal-core §4): lo que declara de una referencia
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

// columnasDeUnTipo son los encabezados de la tabla de los recursos y plazos:
// nombre, código, duración, cómputo y preceptos de cada tipo.
var columnasDeUnTipo = []string{"Tipo", "Código", "Duración", "Cómputo", "Precepto"}

// generadoresDeReferencias es la tabla de generadores (research.md D20; contrato
// skill-legal-core §4): la fila de cada referencia por su nombre, que es el
// que declara kitlegal-referencias y el de references/<nombre>.md. Una referencia
// sin fila no tiene generador, y dos pueden salir del mismo YAML de datos. Los
// recursos y plazos y la cobertura por materia no son referencias enteras sino
// la tabla de una región de la referencia de un paquete (generadoresDeRegiones; contracts/paquetes.md
// §3 de H32): la cita de sus preceptos en la forma única lleva la abreviatura de
// la norma, de data/normas.yaml, y la línea de cada fila de la cobertura, la forma
// del cuerpo de la skill que la tiene.
var generadoresDeReferencias = map[string]generadorDeReferencia{
	"normas":              generadorDeLasNormas,
	"leyes_vertebrales":   generadorDeLasLeyesVertebrales,
	"jerarquia_normativa": generadorDeLaJerarquia,
}

// cabecera es la primera línea de la referencia: la marca de fichero generado
// que nombra el YAML de datos del que sale (FR-031, FR-065, FR-066). La ruta es
// texto de la marca y no una ruta del sistema, así que va siempre con barra.
func (g generadorDeReferencia) cabecera() string {
	return inicioDeLaCabecera + g.datos + finDeLaCabecera
}

// inicioDeLaCabecera y finDeLaCabecera rodean, en la cabecera de una referencia
// generada, al nombre del YAML de datos dentro de data/: la escribe cabecera y
// la lee DatosDeLaReferencia.
const (
	inicioDeLaCabecera = "<!-- generado desde " + carpetaDeLosDatos + "/"
	finDeLaCabecera    = ", no editar -->"
)

// DatosDeLaReferencia es el YAML de datos del que sale una referencia generada,
// como lo nombra su cabecera, data/<fichero> (research D11 y FR-001(c) de H30;
// FR-031): la primera línea es exactamente la cabecera que escribe cabecera, con
// un fichero. Sin esa cabecera la referencia no es generada, y da false.
func DatosDeLaReferencia(referencia []byte) (string, bool) {
	linea, _, _ := bytes.Cut(referencia, []byte("\n"))

	fichero, conInicio := strings.CutPrefix(string(linea), inicioDeLaCabecera)
	fichero, conFin := strings.CutSuffix(fichero, finDeLaCabecera)

	if !conInicio || !conFin || fichero == "" {
		return "", false
	}

	return carpetaDeLosDatos + "/" + fichero, true
}

// ReferenciasAMano son las referencias a mano de la skill del árbol, que es la
// raíz de un repositorio (contracts/gates-de-skills-check.md §3; research D13;
// data-model §3; FR-012): los ficheros regulares .md de
// skills/<skill>/references/ que su kitlegal-referencias no declara y cuya
// primera línea no es la cabecera de una referencia generada
// (DatosDeLaReferencia), con sus bytes y en orden de nombre. No lo son lo que no
// es un fichero regular .md —un subdirectorio, un enlace, otra extensión—, la
// referencia declarada, con su cabecera o sin ella, ni la no declarada que abre
// con la cabecera, que es sobrante. Sin references/ no hay ninguna, y no se lee
// SKILL.md. El error queda para lo que no se puede leer: references/, el
// frontmatter de SKILL.md o una referencia.
func ReferenciasAMano(arbol fs.FS, skill string) ([]Referencia, error) {
	directorio := path.Join(carpetaDeLasSkills, skill)
	carpeta := path.Join(directorio, carpetaDeLasReferencias)

	entradas, err := fs.ReadDir(arbol, carpeta)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("%s no se puede listar: %w", carpeta, err)
	}

	declaradas, err := ficherosDeclarados(arbol, path.Join(directorio, ficheroDeLaSkill))
	if err != nil {
		return nil, err
	}

	var aMano []Referencia

	for _, entrada := range entradas {
		nombre := entrada.Name()
		if !esFicheroMd(entrada) || declaradas[nombre] {
			continue
		}

		ruta := path.Join(carpeta, nombre)

		contenido, err := fs.ReadFile(arbol, ruta)
		if err != nil {
			return nil, fmt.Errorf("%s no se puede leer: %w", ruta, err)
		}

		if _, generada := DatosDeLaReferencia(contenido); generada {
			continue
		}

		aMano = append(aMano, Referencia{Fichero: nombre, Contenido: contenido})
	}

	return aMano, nil
}

// esFicheroMd dice si una entrada de references/, vista sin seguirla, es lo
// único que references/ admite (FR-006 de H31): un fichero regular .md.
func esFicheroMd(entrada fs.DirEntry) bool {
	return entrada.Type().IsRegular() && path.Ext(entrada.Name()) == extensionDeLasReferencias
}

// entradasQueNoSonFicherosMd son los nombres de las entradas de
// skills/<skill>/references/ del árbol, que es la raíz de un repositorio, que no
// son un fichero regular .md —un subdirectorio, un enlace, un fichero de otra
// extensión o sin ella—, declaradas o no, en orden de nombre
// (contracts/gates-de-skills-check.md §3; research D13; FR-006 de H31): cada una
// es un defecto, no una deriva, y nadie la retira. Sin references/ no hay
// ninguna. El error queda para un references/ que no se puede listar.
func entradasQueNoSonFicherosMd(arbol fs.FS, skill string) ([]string, error) {
	carpeta := path.Join(carpetaDeLasSkills, skill, carpetaDeLasReferencias)

	entradas, err := fs.ReadDir(arbol, carpeta)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("%s no se puede listar: %w", carpeta, err)
	}

	var nombres []string

	for _, entrada := range entradas {
		if !esFicheroMd(entrada) {
			nombres = append(nombres, entrada.Name())
		}
	}

	return nombres, nil
}

// defectosDeEnlacesEIndices son los defectos de las referencias de la skill —las
// generadas, como las regenera su fila, y las a mano, como están— con el cuerpo
// de su SKILL.md regenerado (contracts/gates-de-skills-check.md §2 y §2.3;
// research D13 y D14; FR-007 a FR-009 de H31), referencia a referencia en el
// orden de sus nombres y, en cada una, en este orden:
//
//   - si ningún enlace del cuerpo, fuera de un bloque de código, tiene por
//     destino references/<fichero>, con o sin #apartado (FR-007);
//   - cada otra referencia de la skill a la que lleva un enlace suyo, resuelto
//     desde references/, una vez y en el orden en que aparece (FR-008);
//   - lo que le falta para abrir con su índice si tiene más de 100 líneas
//     (FR-009).
//
// Cada defecto nombra el fichero, references/<fichero>, y la regla.
func defectosDeEnlacesEIndices(skill string, cuerpo []byte, referencias []Referencia) []*DefectoDeSkill {
	ordenadas := slices.SortedFunc(slices.Values(referencias), func(a, b Referencia) int {
		return cmp.Compare(a.Fichero, b.Fichero)
	})

	deLaSkill := make(map[string]bool, len(ordenadas))
	for _, referencia := range ordenadas {
		deLaSkill[rutaDeLaReferencia(referencia)] = true
	}

	enlazadas := enlacesDeMarkdown(cuerpo)

	var defectos []*DefectoDeSkill

	defecto := func(ruta, regla string) {
		defectos = append(defectos, &DefectoDeSkill{Skill: skill, Defecto: ruta + ": " + regla})
	}

	for _, referencia := range ordenadas {
		ruta := rutaDeLaReferencia(referencia)

		if !slices.Contains(enlazadas, ruta) {
			defecto(ruta, ficheroDeLaSkill+" no la enlaza con []("+ruta+")")
		}

		var otras []string

		for _, enlazada := range ficherosEnlazados(referencia.Contenido, carpetaDeLasReferencias) {
			if enlazada != ruta && deLaSkill[enlazada] && !slices.Contains(otras, enlazada) {
				otras = append(otras, enlazada)
				defecto(ruta, "enlaza otra referencia, "+enlazada)
			}
		}

		if falta := faltaDelIndice(referencia.Contenido); falta != "" {
			defecto(ruta, falta)
		}
	}

	return defectos
}

// rutaDeLaReferencia es la ruta de la referencia dentro del directorio de su
// skill, con barra, como la nombran los defectos y los enlaces de SKILL.md:
// references/<fichero>.
func rutaDeLaReferencia(referencia Referencia) string {
	return carpetaDeLasReferencias + "/" + referencia.Fichero
}

// ficherosDeclarados son los ficheros de references/ que declara el
// kitlegal-referencias del SKILL.md de la ruta del árbol: <nombre>.md por cada
// nombre. Es un error que SKILL.md no se pueda leer o que su frontmatter no se
// lea.
func ficherosDeclarados(arbol fs.FS, skillMd string) (map[string]bool, error) {
	contenido, err := fs.ReadFile(arbol, skillMd)
	if err != nil {
		return nil, fmt.Errorf("%s no se puede leer: %w", skillMd, err)
	}

	frontmatter, err := LeerFrontmatter(contenido)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", skillMd, err)
	}

	nombres := frontmatter.DeclaracionDeKitlegal().Referencias
	declarados := make(map[string]bool, len(nombres))

	for _, nombre := range nombres {
		declarados[nombre+extensionDeLasReferencias] = true
	}

	return declarados, nil
}

// componer es la referencia con el cuerpo dado: la cabecera, el título, la fila
// de títulos de la tabla y su fila de separación, y detrás el cuerpo, con \n como
// fin de línea.
func (g generadorDeReferencia) componer(cuerpo []byte) []byte {
	var referencia bytes.Buffer

	referencia.WriteString(g.cabecera() + "\n\n# " + g.titulo + "\n\n")
	referencia.WriteString(encabezadoDeLaTabla(g.encabezados))
	referencia.Write(cuerpo)

	return referencia.Bytes()
}

// encabezadoDeLaTabla es la fila de títulos de una tabla y su fila de
// separación, cada una con su salto de línea.
func encabezadoDeLaTabla(encabezados []string) string {
	return filaDeLaTabla(encabezados) + "|" + strings.Repeat("---|", len(encabezados)) + "\n"
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
// de data/normas.yaml (contrato skill-legal-core §2; FR-065, FR-067): como
// RenderizarNormas, con su propio título y solo con las normas marcadas
// vertebral; si no hay ninguna, la tabla va sin filas. Las normas que recibe no
// cambian.
func RenderizarLeyesVertebrales(normas []Norma) []byte {
	return generadorDeLasLeyesVertebrales.componer(filasDeLasNormas(vertebrales(normas)))
}

// RenderizarJerarquia da references/jerarquia_normativa.md desde la jerarquía de
// data/jerarquia.yaml (contrato skill-legal-core §2; FR-066, FR-067): la
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

// RenderizarRecursosYPlazos da la tabla de la región recursos_y_plazos desde los
// tipos de plazo de data/plazos.yaml y las normas de data/normas.yaml
// (contracts/paquetes.md §3; research D7 de H32; FR-010, FR-019): la fila de
// títulos, su separación y una fila por tipo —su nombre, su código, su duración,
// su cómputo y sus preceptos— y, en el que lleva unidades, el genérico, una por
// unidad, en el orden de plazos.Unidades y con el precepto que da su cómputo.
//
// La duración que fija un tipo va con el nombre de su unidad en singular o en
// plural; la del que no la fija dice que la recibe, en la unidad de su fila o con
// la suya. Los preceptos van separados por «; », cada uno en la forma única
// (citaDelPrecepto), y ninguna celda lleva texto de ninguna norma. Los tipos van
// en el orden en que llegan, que es el del fichero, y ni ellos ni las normas
// cambian. Es un error, que la nombra, un precepto de una norma que no está entre
// las normas: sin ella no hay forma legible.
func RenderizarRecursosYPlazos(tipos []plazos.Tipo, normas []Norma) ([]byte, error) {
	filas, err := filasDeLosTipos(tipos, normasPorIdentificador(normas))
	if err != nil {
		return nil, err
	}

	return slices.Concat([]byte(encabezadoDeLaTabla(columnasDeUnTipo)), filas), nil
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

// filasDeLosTipos es una fila por tipo y, en el que lleva unidades, una por
// unidad, en el orden de plazos.Unidades: el mapa de unidades no tiene orden. La
// cita de cada precepto lleva la forma legible de su norma, de las normas por su
// identificador; un precepto de una norma que no está es un error.
func filasDeLosTipos(tipos []plazos.Tipo, normas map[string]Norma) ([]byte, error) {
	var filas bytes.Buffer

	for _, tipo := range tipos {
		if len(tipo.Unidades) == 0 {
			fila, err := filaDelTipo(tipo, duracionDelTipo(tipo, ""), tipo.Preceptos, normas)
			if err != nil {
				return nil, err
			}

			filas.WriteString(fila)

			continue
		}

		for _, unidad := range plazos.Unidades() {
			precepto, tiene := tipo.Unidades[unidad]
			if !tiene {
				continue
			}

			fila, err := filaDelTipo(tipo, duracionDelTipo(tipo, unidad), append(slices.Clone(tipo.Preceptos), precepto),
				normas)
			if err != nil {
				return nil, err
			}

			filas.WriteString(fila)
		}
	}

	return filas.Bytes(), nil
}

// filaDelTipo es la fila de un tipo con su duración y sus preceptos, cada uno
// citado con la norma de su identificador entre las normas.
func filaDelTipo(tipo plazos.Tipo, duracion string, preceptos []plazos.Precepto, normas map[string]Norma) (
	string, error,
) {
	citas := make([]string, 0, len(preceptos))

	for _, precepto := range preceptos {
		norma, esta := normas[precepto.Norma]
		if !esta {
			return "", fmt.Errorf("el precepto %s del tipo %s es de %s, que no está en %s/%s", precepto.Apartado,
				tipo.Codigo, precepto.Norma, carpetaDeLosDatos, ficheroDeLasNormas)
		}

		citas = append(citas, escapeDeCelda.Replace(citaDelPrecepto(norma, precepto.Bloque, precepto.Apartado)))
	}

	return filaDeLaTabla([]string{
		escapeDeCelda.Replace(tipo.Nombre),
		"`" + escapeDeCelda.Replace(tipo.Codigo) + "`",
		duracion,
		"`" + escapeDeCelda.Replace(string(tipo.Computo)) + "`",
		strings.Join(citas, separadorDePreceptos),
	}), nil
}

// duracionDelTipo es la celda de la duración de un tipo: la que fija, con el
// nombre de su unidad; o, si no la fija, la que recibe, en la unidad dada —la de
// una fila del genérico— o con la suya.
func duracionDelTipo(tipo plazos.Tipo, unidad plazos.Unidad) string {
	switch {
	case tipo.Duracion != nil:
		return duracionEnPalabras(*tipo.Duracion)
	case unidad != "":
		return "la que recibe, en `" + escapeDeCelda.Replace(string(unidad)) + "`"
	default:
		return "la que recibe, con su unidad"
	}
}

// duracionEnPalabras es la cantidad con el nombre de su unidad, en singular si
// es 1 y en plural si no. Una unidad sin nombre en nombresDeLasUnidades va tal
// cual: LeerTipos no la deja pasar, y TestRenderizarRecursosYPlazos exige un
// nombre para cada unidad de plazos.Unidades.
func duracionEnPalabras(duracion plazos.Duracion) string {
	nombre := escapeDeCelda.Replace(string(duracion.Unidad))

	if nombres, tiene := nombresDeLasUnidades[duracion.Unidad]; tiene {
		nombre = nombres[1]
		if duracion.Cantidad == 1 {
			nombre = nombres[0]
		}
	}

	return strconv.Itoa(duracion.Cantidad) + " " + nombre
}

// citaDelPrecepto es la cita de un apartado de una norma consolidada del BOE en
// la forma única (contracts/paquetes.md §3; research D7 de H32; FR-010): su forma
// legible y, detrás, entre corchetes, el identificador de la norma y el bloque de
// su índice, sin texto de la norma: «art. 124.1 LPAC [BOE-A-2015-10565, bloque
// a124]», «DA 1ª.2.a LPAC [BOE-A-2015-10565, bloque daprimera]».
func citaDelPrecepto(norma Norma, bloque, apartado string) string {
	return formaLegible(norma, apartado) + " [" + norma.Identificador + ", bloque " + bloque + "]"
}

// formaLegible es el apartado de la norma como se lee: «art.» delante del
// apartado de un artículo, que empieza por una cifra —el de una disposición, como
// «DA 1ª.2.a», ya dice qué es—, un espacio y la abreviatura de la norma o, si no
// la tiene, su título hasta la primera coma, como «Ley 29/1998».
func formaLegible(norma Norma, apartado string) string {
	if strings.IndexAny(apartado, "0123456789") == 0 {
		apartado = "art. " + apartado
	}

	nombre := norma.Abreviatura
	if nombre == "" {
		nombre, _, _ = strings.Cut(norma.Titulo, ",")
	}

	return apartado + " " + nombre
}

// normasPorIdentificador son las normas por su identificador.
func normasPorIdentificador(normas []Norma) map[string]Norma {
	porIdentificador := make(map[string]Norma, len(normas))

	for _, norma := range normas {
		porIdentificador[norma.Identificador] = norma
	}

	return porIdentificador
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
