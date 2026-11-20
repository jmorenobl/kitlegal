package skills

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/jmorenobl/kitlegal/internal/core/plazos"
)

// Lo que forma las marcas de una región generada de una referencia a mano
// (contracts/paquetes.md §3; data-model §4 de H32): la de inicio nombra el
// generador y cada YAML de data/ del que sale la tabla, y la de fin, el
// generador. Cada una va en su línea.
const (
	inicioDeLaMarcaDeInicio      = "<!-- inicio de "
	antesDeLosDatosDeLaMarca     = ": generada desde "
	separadorDeLosDatosDeLaMarca = " y "
	finDeLaMarcaDeInicio         = " con make skills-sync, no editar -->"
	inicioDeLaMarcaDeFin         = "<!-- fin de "
	finDeLaMarcaDeFin            = " -->"
)

// Los YAML de data/ de los que sale cada región, además de data/normas.yaml
// (ficheroDeLasNormas).
const (
	ficheroDeLaCobertura = "cobertura.yaml"
	ficheroDeLosPlazos   = "plazos.yaml"
)

// generadorDeLaCobertura es el nombre del generador de la región de la cobertura
// por materia: el único cuyas filas dependen del paquete de la referencia, de la
// plantilla del cuerpo de la skill y de los motivos de la referencia.
const generadorDeLaCobertura = "cobertura_por_materia"

// Lo que el generador de la cobertura lee del cuerpo de la skill y de la
// referencia de su paquete (contracts/datos-y-regiones.md §5; data-model §2 de
// H33): las líneas de un bloque text —la plantilla, que empieza por la etiqueta y
// lleva el marcador del motivo; los motivos, que llevan el del precepto y no
// empiezan por el aviso—, y los marcadores que sustituye en ellas con los datos
// de cada fila.
const (
	lenguajeDeLasFormas               = "text"
	inicioDeUnAviso                   = "⚠"
	etiquetaDeNoCubierto              = inicioDeUnAviso + " NO CUBIERTO: "
	marcadorDelProcedimiento          = "<procedimiento>"
	marcadorDelAsunto                 = "<asunto>"
	marcadorDelMotivo                 = "<motivo>"
	marcadorDeLaNorma                 = "<norma>"
	marcadorDelIdentificadorDeLaNorma = "<identificador de la norma>"
	marcadorDelPrecepto               = "<precepto>"
	marcadorDelIdentificador          = "<identificador>"
	marcadorDelID                     = "<id>"
)

// marcadoresDeLaCobertura son los marcadores que la tabla de la cobertura conoce
// (data-model §2 de H33).
var marcadoresDeLaCobertura = []string{
	marcadorDelProcedimiento, marcadorDelAsunto, marcadorDelMotivo, marcadorDeLaNorma,
	marcadorDelIdentificadorDeLaNorma, marcadorDelPrecepto, marcadorDelIdentificador, marcadorDelID,
}

// marcadorDeUnaForma es un marcador de una forma: el texto entre < y >.
var marcadorDeUnaForma = regexp.MustCompile(`<[^<>]+>`)

// generadorDeRegion es un generador de internal/skills que admite región en una
// referencia a mano (contracts/paquetes.md §3 de H32): los YAML de data/ de los
// que sale su tabla, el título de cada columna y la función que da sus filas.
type generadorDeRegion struct {
	// datos son los YAML dentro de data/ de los que sale, en el orden en que los
	// nombra su marca de inicio.
	datos []string

	// encabezados son los títulos de cada columna de su tabla.
	encabezados []string

	// filas da las filas de la tabla desde los datos y la región, o lo que lo
	// impide: un *defectoDeLosDatos o un error de lectura.
	filas func(datos *datosDeLasRegiones, region contextoDeLaRegion) ([]byte, error)
}

// contextoDeLaRegion es lo que una región toma de su sitio: el paquete que nombra
// su referencia, y la plantilla ⚠ NO CUBIERTO: del cuerpo de su skill con los
// motivos de su referencia.
type contextoDeLaRegion struct {
	// paquete es el nombre de la referencia sin .md: el de su paquete si lo es.
	paquete string

	// linea es la plantilla del SKILL.md de la skill con los motivos de la
	// referencia: lo que da la línea de cada fila de la cobertura.
	linea lineaDeNoCubierto
}

// generadoresDeRegiones son los tres generadores que admiten región, por su
// nombre (contracts/paquetes.md §3 de H32): la cobertura por materia, con las
// filas del paquete de la referencia; los recursos y plazos, con cada precepto en
// la forma única; y las leyes vertebrales.
var generadoresDeRegiones = map[string]generadorDeRegion{
	generadorDeLaCobertura: {
		datos:       []string{ficheroDeLaCobertura, ficheroDeLasNormas},
		encabezados: []string{"Procedimiento", "Línea que da la respuesta"},
		filas:       filasDeLaCobertura,
	},
	"recursos_y_plazos": {
		datos:       []string{ficheroDeLosPlazos, ficheroDeLasNormas},
		encabezados: columnasDeUnTipo,
		filas:       filasDeLosRecursosYPlazos,
	},
	"leyes_vertebrales": {
		datos:       []string{ficheroDeLasNormas},
		encabezados: columnasDeUnaNorma,
		filas:       filasDeLasLeyesVertebrales,
	},
}

// marcas son las dos marcas de la región del generador de nombre: la de inicio,
// con cada YAML de data/ del que sale, y la de fin.
func (g generadorDeRegion) marcas(nombre string) regionDeMarcas {
	datos := make([]string, 0, len(g.datos))
	for _, fichero := range g.datos {
		datos = append(datos, carpetaDeLosDatos+"/"+fichero)
	}

	return regionDeMarcas{
		nombre: "la región " + nombre,
		inicio: inicioDeLaMarcaDeInicio + nombre + antesDeLosDatosDeLaMarca +
			strings.Join(datos, separadorDeLosDatosDeLaMarca) + finDeLaMarcaDeInicio,
		fin: inicioDeLaMarcaDeFin + nombre + finDeLaMarcaDeFin,
	}
}

// DatosDeLasRegiones son los YAML de data/ que nombran las marcas de inicio de
// cada región de una referencia, data/<fichero>, cada uno una vez y en el orden
// en que aparecen (contracts/paquetes.md §3 y contracts/evals.md §5 (c) de H32):
// lo que la región carga, como la cabecera de una referencia generada
// (DatosDeLaReferencia). Una línea es una marca de inicio si tiene su forma, con
// cualquier generador; sin ninguna, no hay ninguno.
func DatosDeLasRegiones(referencia []byte) []string {
	var datos []string

	for linea := range bytes.Lines(referencia) {
		resto, esInicio := strings.CutPrefix(string(bytes.TrimSuffix(linea, []byte("\n"))), inicioDeLaMarcaDeInicio)
		_, resto, conDatos := strings.Cut(resto, antesDeLosDatosDeLaMarca)
		resto, conFin := strings.CutSuffix(resto, finDeLaMarcaDeInicio)

		if !esInicio || !conDatos || !conFin {
			continue
		}

		for _, fichero := range strings.Split(resto, separadorDeLosDatosDeLaMarca) {
			if strings.HasPrefix(fichero, carpetaDeLosDatos+"/") && !slices.Contains(datos, fichero) {
				datos = append(datos, fichero)
			}
		}
	}

	return datos
}

// regionEncontrada es una región de una referencia a mano: su generador y lo que
// hay entre sus marcas.
type regionEncontrada struct {
	generador string
	entre     entreMarcas
}

// regionesDeLaReferencia da cada región del contenido de una referencia a mano
// cuyas marcas van en pareja, en el orden en que van, y los textos de los
// defectos de sus marcas, sin el fichero delante (contracts/paquetes.md §3 de
// H32): una marca de un generador que no admite región, una vez por generador;
// una marca de inicio de uno que la admite que no es la suya, con sus datos; unas
// marcas que no van en pareja; y una región que empieza dentro de otra. Cada
// generador admite una región por referencia. Con algún defecto, ninguna región
// se regenera, pero la de la cobertura sigue siendo la de las filas de su
// paquete.
func regionesDeLaReferencia(contenido []byte) ([]regionEncontrada, []string) {
	var defectos []string

	vistos := make(map[string]bool)

	for linea := range bytes.Lines(contenido) {
		texto := string(bytes.TrimSuffix(linea, []byte("\n")))

		generador, esMarca := generadorDeLaMarca(texto)
		if !esMarca || vistos[generador] {
			continue
		}

		deLaRegion, admite := generadoresDeRegiones[generador]
		inicio := deLaRegion.marcas(generador).inicio

		switch {
		case !admite:
			vistos[generador] = true
			defectos = append(defectos, "región de "+generador+", un generador que no la admite")
		case strings.HasPrefix(texto, inicioDeLaMarcaDeInicio) && texto != inicio:
			vistos[generador] = true
			defectos = append(defectos, "la marca de inicio de la región "+generador+" no es «"+inicio+"»")
		}
	}

	var encontradas []regionEncontrada

	for _, generador := range slices.Sorted(maps.Keys(generadoresDeRegiones)) {
		if vistos[generador] {
			continue
		}

		entre, lleva, err := generadoresDeRegiones[generador].marcas(generador).buscar(contenido)

		switch {
		case err != nil:
			defectos = append(defectos, err.Error())
		case lleva:
			encontradas = append(encontradas, regionEncontrada{generador: generador, entre: entre})
		}
	}

	slices.SortFunc(encontradas, func(a, b regionEncontrada) int { return cmp.Compare(a.entre.desde, b.entre.desde) })

	for posicion := 1; posicion < len(encontradas); posicion++ {
		if anterior := encontradas[posicion-1]; encontradas[posicion].entre.desde <= anterior.entre.hasta {
			defectos = append(defectos, fmt.Sprintf("las marcas de la región %s no van en pareja: van dentro de la región %s",
				encontradas[posicion].generador, anterior.generador))
		}
	}

	return encontradas, defectos
}

// generadorDeLaMarca es el generador que nombra una línea con la forma de una
// marca de región —la de inicio, hasta sus dos puntos, o la de fin—, y si la
// tiene.
func generadorDeLaMarca(linea string) (string, bool) {
	if resto, esFin := strings.CutPrefix(linea, inicioDeLaMarcaDeFin); esFin {
		generador, cerrada := strings.CutSuffix(resto, finDeLaMarcaDeFin)

		return generador, cerrada && generador != ""
	}

	resto, esInicio := strings.CutPrefix(linea, inicioDeLaMarcaDeInicio)
	if !esInicio || !strings.HasSuffix(resto, finDeLaMarcaDeFin) {
		return "", false
	}

	generador, _, _ := strings.Cut(strings.TrimSuffix(resto, finDeLaMarcaDeFin), ":")

	return generador, generador != ""
}

// lineaDeNoCubierto es de lo que sale la línea ⚠ NO CUBIERTO: de cada fila de la
// cobertura, con sus marcadores (data-model §2; research D2 de H33): la plantilla
// del cuerpo de la skill y los dos motivos de la referencia de su paquete, el de
// la fila que trae la norma que la rige y el de la que no. Ningún texto de la
// línea vive en Go (FR-006 de H33).
type lineaDeNoCubierto struct {
	plantilla, motivoConNorma, motivoSinNorma string
}

// plantillaDelCuerpo es la plantilla ⚠ NO CUBIERTO: del SKILL.md de contenido:
// la única línea de sus bloques text que, sin los blancos del principio, empieza
// por la etiqueta y lleva <motivo>. Con otra cantidad, la skill no la tiene.
func plantillaDelCuerpo(contenido []byte) (string, bool) {
	var plantillas []string

	for _, linea := range lineasDeLosBloquesText(contenido) {
		linea = strings.TrimLeft(linea, " ")

		if strings.HasPrefix(linea, etiquetaDeNoCubierto) && strings.Contains(linea, marcadorDelMotivo) {
			plantillas = append(plantillas, linea)
		}
	}

	if len(plantillas) != 1 {
		return "", false
	}

	return plantillas[0], true
}

// conLosMotivosDe es la línea con los motivos de la referencia de contenido: las
// líneas de sus bloques text que, sin los blancos del principio, llevan
// <precepto> y no empiezan por ⚠; la que lleva <norma> y la que no. Si no son
// dos, una de cada, la referencia no los tiene y la línea queda sin ellos.
func (l lineaDeNoCubierto) conLosMotivosDe(contenido []byte) (lineaDeNoCubierto, bool) {
	var conNorma, sinNorma []string

	for _, linea := range lineasDeLosBloquesText(contenido) {
		linea = strings.TrimLeft(linea, " ")

		switch {
		case !strings.Contains(linea, marcadorDelPrecepto) || strings.HasPrefix(linea, inicioDeUnAviso):
		case strings.Contains(linea, marcadorDeLaNorma):
			conNorma = append(conNorma, linea)
		default:
			sinNorma = append(sinNorma, linea)
		}
	}

	if len(conNorma) != 1 || len(sinNorma) != 1 {
		return l, false
	}

	l.motivoConNorma, l.motivoSinNorma = conNorma[0], sinNorma[0]

	return l, true
}

// lineasDeLosBloquesText son las líneas del contenido, sin su salto, que están
// dentro de un bloque de código text, en su orden: como en lineasFueraDeBloques,
// un bloque va de la valla que lo abre a la que lo cierra, y es text si lo dice
// lo que sigue a la valla que lo abre.
func lineasDeLosBloquesText(contenido []byte) []string {
	var (
		dentro  []string
		abierta string
		deText  bool
	)

	for linea := range strings.Lines(string(contenido)) {
		linea = strings.TrimSuffix(linea, "\n")
		valla, resto := vallaDeLaLinea(linea)

		switch {
		case abierta == "" && valla != "":
			abierta, deText = valla, strings.TrimSpace(resto) == lenguajeDeLasFormas
		case abierta == "":
		case valla != "" && valla[0] == abierta[0] && len(valla) >= len(abierta) && strings.TrimSpace(resto) == "":
			abierta = ""
		case deText:
			dentro = append(dentro, linea)
		}
	}

	return dentro
}

// desconocidos son los marcadores que la tabla de la cobertura no conoce
// (contracts/datos-y-regiones.md §5 de H33), cada uno una vez y en el orden en
// que aparecen: los de la plantilla y, detrás, los de los motivos que no estaban
// en ella.
func (l lineaDeNoCubierto) desconocidos() (deLaPlantilla, deLosMotivos []string) {
	deLaPlantilla = marcadoresDesconocidos(l.plantilla, nil)
	deLosMotivos = marcadoresDesconocidos(l.motivoConNorma+"\n"+l.motivoSinNorma, deLaPlantilla)

	return deLaPlantilla, deLosMotivos
}

// marcadoresDesconocidos son los marcadores del texto que la tabla de la
// cobertura no conoce y que no están entre los vistos, cada uno una vez y en el
// orden en que aparecen.
func marcadoresDesconocidos(texto string, vistos []string) []string {
	var desconocidos []string

	for _, marcador := range marcadorDeUnaForma.FindAllString(texto, -1) {
		if !slices.Contains(marcadoresDeLaCobertura, marcador) && !slices.Contains(vistos, marcador) &&
			!slices.Contains(desconocidos, marcador) {
			desconocidos = append(desconocidos, marcador)
		}
	}

	return desconocidos
}

// linea es la línea ⚠ NO CUBIERTO: de la entrada de la cobertura cuyo precepto
// es de la norma dada (data-model §2 de H33): la plantilla con <motivo> por el
// motivo con la norma que lo rige si la entrada la trae, por el otro si no, y
// después, en una sola pasada, cada marcador sustituido por su dato —<asunto> y
// <procedimiento>, por el procedimiento— y el precepto en su forma legible. Lo
// sustituido no se vuelve a mirar.
func (l lineaDeNoCubierto) linea(entrada ReguladoAparte, norma Norma) string {
	motivo, nombre, identificador := l.motivoSinNorma, "", ""
	if entrada.Norma != nil {
		motivo, nombre, identificador = l.motivoConNorma, entrada.Norma.Nombre, entrada.Norma.Identificador
	}

	return strings.NewReplacer(
		marcadorDelAsunto, entrada.Procedimiento,
		marcadorDelProcedimiento, entrada.Procedimiento,
		marcadorDeLaNorma, nombre,
		marcadorDelIdentificadorDeLaNorma, identificador,
		marcadorDelPrecepto, formaLegible(norma, entrada.Precepto.Apartado),
		marcadorDelIdentificador, entrada.Precepto.Norma,
		marcadorDelID, entrada.Precepto.Bloque,
	).Replace(strings.ReplaceAll(l.plantilla, marcadorDelMotivo, motivo))
}

// defectoDeLosDatos es lo que impide generar una región desde un YAML de data/:
// que falte, sin causa, o los defectos de su lectura, en causa.
type defectoDeLosDatos struct {
	// fichero es el YAML, dentro de data/.
	fichero string

	// causa son los defectos de su lectura, o nil si falta.
	causa error
}

// Error dice el fichero y lo que le pasa.
func (d *defectoDeLosDatos) Error() string {
	if d.causa == nil {
		return carpetaDeLosDatos + "/" + d.fichero + " no existe"
	}

	return carpetaDeLosDatos + "/" + d.fichero + ": " + d.causa.Error()
}

// Unwrap es la causa.
func (d *defectoDeLosDatos) Unwrap() error {
	return d.causa
}

// datoLeido es un YAML de data/ leído: su valor, si existe, y los defectos de su
// lectura, con los que no hay valor.
type datoLeido[T any] struct {
	valor   T
	existe  bool
	defecto error
}

// leerUnDato lee una sola vez, la primera que se pide, el YAML de data/ del
// repositorio de raíz con su lector. Que no exista no es un error; el error queda
// para lo que no se puede leer.
func leerUnDato[T any](raiz, fichero string, leer func([]byte) (T, error)) func() (datoLeido[T], error) {
	return sync.OnceValues(func() (datoLeido[T], error) {
		contenido, err := leerFichero(filepath.Join(raiz, carpetaDeLosDatos, fichero))

		switch {
		case errors.Is(err, fs.ErrNotExist):
			return datoLeido[T]{}, nil
		case err != nil:
			return datoLeido[T]{}, fmt.Errorf("%s/%s no se puede leer: %w", carpetaDeLosDatos, fichero, err)
		}

		valor, defecto := leer(contenido)

		return datoLeido[T]{valor: valor, existe: true, defecto: defecto}, nil
	})
}

// leido dice si el YAML existe y se ha leído sin defectos: si tiene valor.
func (d datoLeido[T]) leido() bool {
	return d.existe && d.defecto == nil
}

// usar es el valor del YAML de data/ que da leer, o lo que impide usarlo: un
// *defectoDeLosDatos si falta o tiene defectos, o el error de su lectura.
func usar[T any](leer func() (datoLeido[T], error), fichero string) (T, error) {
	var cero T

	dato, err := leer()

	switch {
	case err != nil:
		return cero, err
	case !dato.existe:
		return cero, &defectoDeLosDatos{fichero: fichero}
	case dato.defecto != nil:
		return cero, &defectoDeLosDatos{fichero: fichero, causa: dato.defecto}
	}

	return dato.valor, nil
}

// datosDeLasRegiones son los YAML de data/ de los que sale cada región de un
// repositorio, cada uno leído una vez y solo si alguna región o la comprobación de
// las filas de la cobertura lo pide.
type datosDeLasRegiones struct {
	normas    func() (datoLeido[[]Norma], error)
	tipos     func() (datoLeido[[]plazos.Tipo], error)
	cobertura func() (datoLeido[Cobertura], error)
}

// leerDatosDeLasRegiones prepara los datos de cada región del repositorio de
// raíz, sin leer nada todavía.
func leerDatosDeLasRegiones(raiz string) *datosDeLasRegiones {
	return &datosDeLasRegiones{
		normas:    leerUnDato(raiz, ficheroDeLasNormas, LeerNormas),
		tipos:     leerUnDato(raiz, ficheroDeLosPlazos, plazos.LeerTipos),
		cobertura: leerUnDato(raiz, ficheroDeLaCobertura, LeerCobertura),
	}
}

// filasDeLaCobertura son las filas de la región de la cobertura: una por entrada
// de data/cobertura.yaml cuyo precepto es de una norma con el vertical del
// paquete de la región, en su orden, con su procedimiento y su línea ⚠ NO
// CUBIERTO:.
func filasDeLaCobertura(datos *datosDeLasRegiones, region contextoDeLaRegion) ([]byte, error) {
	cobertura, err := usar(datos.cobertura, ficheroDeLaCobertura)
	if err != nil {
		return nil, err
	}

	normas, err := usar(datos.normas, ficheroDeLasNormas)
	if err != nil {
		return nil, err
	}

	porIdentificador := normasPorIdentificador(normas)

	var filas bytes.Buffer

	for _, entrada := range cobertura.ReguladosAparte {
		norma, esta := porIdentificador[entrada.Precepto.Norma]
		if !esta || norma.Vertical != region.paquete {
			continue
		}

		filas.WriteString(filaDeLaTabla([]string{
			escapeDeCelda.Replace(entrada.Procedimiento),
			escapeDeCelda.Replace(region.linea.linea(entrada, norma)),
		}))
	}

	return filas.Bytes(), nil
}

// filasDeLosRecursosYPlazos son las filas de la región de los recursos y plazos:
// los tipos de data/plazos.yaml con sus preceptos citados con las normas de
// data/normas.yaml. Un precepto de una norma que no está es un defecto de
// data/plazos.yaml.
func filasDeLosRecursosYPlazos(datos *datosDeLasRegiones, _ contextoDeLaRegion) ([]byte, error) {
	tipos, err := usar(datos.tipos, ficheroDeLosPlazos)
	if err != nil {
		return nil, err
	}

	normas, err := usar(datos.normas, ficheroDeLasNormas)
	if err != nil {
		return nil, err
	}

	filas, err := filasDeLosTipos(tipos, normasPorIdentificador(normas))
	if err != nil {
		return nil, &defectoDeLosDatos{fichero: ficheroDeLosPlazos, causa: err}
	}

	return filas, nil
}

// filasDeLasLeyesVertebrales son las filas de la región de las leyes
// vertebrales: las normas de data/normas.yaml marcadas vertebral, como en
// references/leyes_vertebrales.md.
func filasDeLasLeyesVertebrales(datos *datosDeLasRegiones, _ contextoDeLaRegion) ([]byte, error) {
	normas, err := usar(datos.normas, ficheroDeLasNormas)
	if err != nil {
		return nil, err
	}

	return filasDeLasNormas(vertebrales(normas)), nil
}

// regionesDeLaSkill es lo que da cada región de las referencias a mano de una
// skill.
type regionesDeLaSkill struct {
	// referencias son las referencias a mano con alguna región, con lo que hay
	// entre sus marcas regenerado, en su orden; las que tienen un defecto no van.
	referencias []Referencia

	// deLasMarcas son los defectos de cada región, con el fichero de su
	// referencia delante.
	deLasMarcas []*DefectoDeSkill

	// deLosDatos son los de los YAML de data/ de los que salen, con su fichero
	// delante y cada uno una vez.
	deLosDatos []*DefectoDeSkill

	// cubiertos son los nombres, sin .md, de las referencias con una región de la
	// cobertura: las filas de su paquete las genera la skill.
	cubiertos []string

	// conPlantilla dice si el SKILL.md de la skill tiene la plantilla ⚠ NO
	// CUBIERTO:.
	conPlantilla bool
}

// regenerar regenera cada región de las referencias a mano de la skill, con su
// SKILL.md regenerado (contracts/paquetes.md §3; research D4 y D6 de H32;
// contracts/datos-y-regiones.md §5 de H33): entre las marcas de cada región, la
// tabla de su generador, y nada fuera de ellas (regenerarLaReferencia). Antes, en
// una skill con la plantilla ⚠ NO CUBIERTO:, comprueba que su prosa dice el
// motivo de lo no declarado (comprobarElMotivo). Es un defecto de la referencia,
// con su fichero, cada defecto de sus marcas; los de una región de la cobertura
// (defectosDeLaCobertura); y una región sin uno de sus YAML de data/. Los
// defectos de un YAML de data/ que no se lee son de ese fichero. El error queda
// para lo que no se puede leer.
func (d *datosDeLasRegiones) regenerar(skill string, skillMd []byte, aMano []Referencia) (regionesDeLaSkill, error) {
	plantilla, conPlantilla := plantillaDelCuerpo(skillMd)
	resultado := regionesDeLaSkill{conPlantilla: conPlantilla}

	if conPlantilla {
		if err := d.comprobarElMotivo(skill, skillMd, &resultado); err != nil {
			return regionesDeLaSkill{}, fmt.Errorf("%s: %s: %w", skill, ficheroDeLaSkill, err)
		}
	}

	for _, referencia := range aMano {
		if err := d.regenerarLaReferencia(skill, referencia, plantilla, &resultado); err != nil {
			return regionesDeLaSkill{}, err
		}
	}

	return resultado, nil
}

// regenerarLaReferencia regenera cada región de la referencia a mano de la skill,
// con la plantilla de su SKILL.md si resultado dice que la tiene, y anota en
// resultado si es la de un paquete con la región de la cobertura, los defectos
// de cada región suya y, sin ninguno, la referencia regenerada. El error queda
// para lo que no se puede leer.
func (d *datosDeLasRegiones) regenerarLaReferencia(skill string, referencia Referencia, plantilla string,
	resultado *regionesDeLaSkill,
) error {
	ruta := rutaDeLaReferencia(referencia)
	defectos := len(resultado.deLasMarcas) + len(resultado.deLosDatos)

	defecto := func(texto string) {
		resultado.deLasMarcas = append(resultado.deLasMarcas, &DefectoDeSkill{Skill: skill, Defecto: ruta + ": " + texto})
	}

	region := contextoDeLaRegion{paquete: strings.TrimSuffix(referencia.Fichero, extensionDeLasReferencias)}

	encontradas, deLasMarcas := regionesDeLaReferencia(referencia.Contenido)
	for _, texto := range deLasMarcas {
		defecto(texto)
	}

	if slices.ContainsFunc(encontradas, func(encontrada regionEncontrada) bool {
		return encontrada.generador == generadorDeLaCobertura
	}) {
		resultado.cubiertos = append(resultado.cubiertos, region.paquete)
	}

	if len(encontradas) == 0 || len(deLasMarcas) > 0 {
		return nil
	}

	tablas := make([][]byte, len(encontradas))

	for posicion, encontrada := range encontradas {
		if encontrada.generador == generadorDeLaCobertura {
			linea, defectosDeLaRegion, conLinea := defectosDeLaCobertura(plantilla, resultado.conPlantilla,
				referencia.Contenido)
			for _, texto := range defectosDeLaRegion {
				defecto(texto)
			}

			if !conLinea {
				continue
			}

			region.linea = linea
		}

		tabla, err := d.tabla(skill, ruta, encontrada.generador, region, resultado)
		if err != nil {
			return fmt.Errorf("%s: %s: %w", skill, ruta, err)
		}

		tablas[posicion] = tabla
	}

	if len(resultado.deLasMarcas)+len(resultado.deLosDatos) > defectos {
		return nil
	}

	contenido := referencia.Contenido
	for posicion, encontrada := range slices.Backward(encontradas) {
		contenido = encontrada.entre.sustituir(contenido, tablas[posicion])
	}

	resultado.referencias = append(resultado.referencias, Referencia{Fichero: referencia.Fichero, Contenido: contenido})

	return nil
}

// comprobarElMotivo anota en el resultado el defecto del SKILL.md de la skill,
// que tiene la plantilla ⚠ NO CUBIERTO:, cuya prosa no dice el motivo de lo no
// declarado de data/cobertura.yaml entre comillas angulares, con cada tramo de
// blancos de los dos —saltos de línea incluidos— reducido a un espacio
// (contracts/datos-y-regiones.md §4; research D3 de H33): el motivo que la
// plantilla lleva en lo que no está declarado lo dice la skill tal cual. Sin
// data/cobertura.yaml, el defecto es el de la plantilla sin él; con defectos en
// él, los suyos. El error queda para lo que no se puede leer.
func (d *datosDeLasRegiones) comprobarElMotivo(skill string, skillMd []byte, resultado *regionesDeLaSkill) error {
	cobertura, err := usar(d.cobertura, ficheroDeLaCobertura)
	if err != nil {
		return resultado.anotarLosDatos(skill,
			ficheroDeLaSkill+": la plantilla "+strings.TrimSuffix(etiquetaDeNoCubierto, " "), err)
	}

	if strings.Contains(conLosBlancosReducidos(string(skillMd)),
		"«"+conLosBlancosReducidos(cobertura.MotivoDeLoNoDeclarado)+"»") {
		return nil
	}

	resultado.deLasMarcas = append(resultado.deLasMarcas, &DefectoDeSkill{
		Skill: skill,
		Defecto: ficheroDeLaSkill + ": la prosa no dice el motivo de lo no declarado, «" +
			cobertura.MotivoDeLoNoDeclarado + "»",
	})

	return nil
}

// conLosBlancosReducidos es el texto con cada tramo de blancos, saltos de línea
// incluidos, reducido a un espacio, y sin los de sus dos extremos.
func conLosBlancosReducidos(texto string) string {
	return strings.Join(strings.Fields(texto), " ")
}

// defectosDeLaCobertura son la línea de la región de la cobertura de una
// referencia de contenido, con la plantilla del SKILL.md de su skill, y los
// defectos de la región, sin su fichero, en el orden de
// contracts/datos-y-regiones.md §5 de H33: en una skill sin plantilla, ese
// defecto y ningún motivo buscado; en una referencia sin sus dos motivos, ese; y
// cada marcador que la tabla no conoce, una vez, los de la plantilla y después
// los de los motivos. Sin plantilla o sin motivos, no hay línea (false).
func defectosDeLaCobertura(plantilla string, conPlantilla bool, contenido []byte) (lineaDeNoCubierto, []string, bool) {
	etiqueta := strings.TrimSuffix(etiquetaDeNoCubierto, " ")

	if !conPlantilla {
		return lineaDeNoCubierto{}, []string{
			"región " + generadorDeLaCobertura + " en una skill sin la plantilla " + etiqueta + " en " + ficheroDeLaSkill,
		}, false
	}

	var defectos []string

	linea, conMotivos := lineaDeNoCubierto{plantilla: plantilla}.conLosMotivosDe(contenido)
	if !conMotivos {
		defectos = append(defectos, "región "+generadorDeLaCobertura+" en una referencia sin sus dos motivos")
	}

	deLaPlantilla, deLosMotivos := linea.desconocidos()

	for _, marcador := range deLaPlantilla {
		defectos = append(defectos, "la plantilla "+etiqueta+" de "+ficheroDeLaSkill+" lleva "+marcador+
			", un marcador que la tabla no conoce")
	}

	for _, marcador := range deLosMotivos {
		defectos = append(defectos, "un motivo de la plantilla lleva "+marcador+", un marcador que la tabla no conoce")
	}

	return linea, defectos, conMotivos
}

// tabla es la tabla de la región del generador de nombre en la referencia de la
// ruta, con sus encabezados y sus filas, o nil si sus datos lo impiden: el defecto
// de la región sin uno de sus YAML va a las marcas de resultado, con la ruta, y
// los del YAML que no se lee, a sus datos. El error queda para lo que no se puede
// leer.
func (d *datosDeLasRegiones) tabla(skill, ruta, nombre string, region contextoDeLaRegion,
	resultado *regionesDeLaSkill,
) ([]byte, error) {
	generador := generadoresDeRegiones[nombre]

	filas, err := generador.filas(d, region)
	if err != nil {
		return nil, resultado.anotarLosDatos(skill, ruta+": la región "+nombre, err)
	}

	return slices.Concat([]byte(encabezadoDeLaTabla(generador.encabezados)), filas), nil
}

// anotarLosDatos anota en el resultado los defectos de la skill que da err, lo
// que impide salir de los YAML de data/ a lo que lo pide —una región o la
// plantilla, con su fichero delante— (defectosDeLosDatosQueFaltan), y devuelve el
// error que no es de los datos.
func (r *regionesDeLaSkill) anotarLosDatos(skill, loQueLoPide string, err error) error {
	deLasMarcas, deLosDatos, err := defectosDeLosDatosQueFaltan(skill, loQueLoPide, err)

	r.deLasMarcas = append(r.deLasMarcas, deLasMarcas...)
	r.deLosDatos = sinRepetir(r.deLosDatos, deLosDatos)

	return err
}

// defectosDeLosDatosQueFaltan son los defectos de la skill que da err, lo que
// impide salir de los YAML de data/ a lo que lo pide: sin uno de ellos, «<lo que
// lo pide> sin data/<fichero>», que va con los de las marcas; con defectos en él,
// los de su lector, con su fichero delante, que van con los de los datos. El
// error que no es un *defectoDeLosDatos es lo que no se puede leer, y se devuelve
// sin defectos.
func defectosDeLosDatosQueFaltan(skill, loQueLoPide string, err error) (
	deLasMarcas, deLosDatos []*DefectoDeSkill, deLectura error,
) {
	var queFalla *defectoDeLosDatos

	switch {
	case errors.As(err, &queFalla) && queFalla.causa == nil:
		return []*DefectoDeSkill{{
			Skill:   skill,
			Defecto: loQueLoPide + " sin " + carpetaDeLosDatos + "/" + queFalla.fichero,
		}}, nil, nil
	case errors.As(err, &queFalla):
		return nil, defectosDelError(skill, carpetaDeLosDatos+"/"+queFalla.fichero, queFalla.causa), nil
	default:
		return nil, nil, err
	}
}

// filasSinRegion son los defectos de las filas de data/cobertura.yaml que
// ninguna región de la cobertura genera (contracts/paquetes.md §3; FR-012 de
// H32), una por fila y en su orden, sin la skill: las de una norma sin vertical o
// que no está en data/normas.yaml, y las de un paquete que no está entre los
// cubiertos. Sin data/cobertura.yaml no hay filas; con defectos en él o en
// data/normas.yaml, no se sabe de qué paquete es cada una, y sus defectos los
// dan sus lectores. El error queda para lo que no se puede leer.
func (d *datosDeLasRegiones) filasSinRegion(cubiertos []string) ([]string, error) {
	cobertura, err := d.cobertura()
	if err != nil {
		return nil, err
	}

	normas, err := d.normas()
	if err != nil {
		return nil, err
	}

	// Sin data/normas.yaml, ninguna norma tiene vertical y ninguna fila tiene
	// región; con defectos, sus defectos son los de su lector.
	if !cobertura.leido() || (normas.existe && !normas.leido()) {
		return nil, nil
	}

	porIdentificador := normasPorIdentificador(normas.valor)

	var defectos []string

	for _, entrada := range cobertura.valor.ReguladosAparte {
		norma, esta := porIdentificador[entrada.Precepto.Norma]
		if esta && norma.Vertical != "" && slices.Contains(cubiertos, norma.Vertical) {
			continue
		}

		defectos = append(defectos, carpetaDeLosDatos+"/"+ficheroDeLaCobertura+": la fila de «"+entrada.Procedimiento+
			"» no la genera ninguna región "+generadorDeLaCobertura)
	}

	return defectos, nil
}
