package workflow

import (
	"bytes"
	"cmp"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

// regenerarHitos es la bandera con la que TestRoadmap escribe en workflow.yml,
// antes de comparar, la lista de hitos que admite el input hito. Solo la pasa
// scripts/roadmap-sync.sh, que es lo que ejecuta make roadmap-sync.
var regenerarHitos = flag.Bool("regenerar-hitos", false,
	"escribe en workflow.yml la lista de hitos de docs/roadmap/ antes de compararla")

const (
	// rutaDelRoadmap es el directorio de los hitos, relativo a la raíz
	// (docs/roadmap/README.md).
	rutaDelRoadmap = "docs/roadmap"

	// Las cuatro carpetas del roadmap: los hitos que se lanzaron, el que se
	// implementa, la cola de los que vienen y los candidatos.
	dirDone       = "done"
	dirInProgress = "in-progress"
	dirTodo       = "todo"
	dirBacklog    = "backlog"

	// leeme es el nombre del fichero que explica un directorio, que no es un
	// hito.
	leeme = "README.md"

	// delimitador abre y cierra la cabecera YAML de un fichero de hito.
	delimitador = "---\n"
)

var (
	// nombreDeHito es el de un fichero de done/, in-progress/ o del backlog:
	// el hito y un slug.
	nombreDeHito = regexp.MustCompile(`^(H[0-9]+(?:\.[0-9]+)?)-[a-z0-9]+(?:-[a-z0-9]+)*\.md$`)

	// nombreEnLaCola es el de un fichero de todo/: su posición, el hito
	// y un slug.
	nombreEnLaCola = regexp.MustCompile(`^([0-9]{3})-(H[0-9]+(?:\.[0-9]+)?)-[a-z0-9]+(?:-[a-z0-9]+)*\.md$`)

	// lineaDelEnum es la línea de workflow.yml con la lista de hitos que admite
	// el input hito.
	lineaDelEnum = regexp.MustCompile(`(?m)^    enum: \["H[^\]\n]*\]$`)

	// errSinCabecera es el defecto de un fichero de hito que no empieza por su
	// cabecera YAML.
	errSinCabecera = errors.New("no empieza por su cabecera YAML (---)")
)

// cabeceraDeHito es la cabecera YAML de un fichero de hito.
type cabeceraDeHito struct {
	ID        string   `yaml:"id"`
	Fase      string   `yaml:"fase"`
	DependeDe []string `yaml:"depende_de"`
}

// hitoDelRoadmap es un fichero de hito leído.
type hitoDelRoadmap struct {
	// ruta es la del fichero, relativa a docs/roadmap/.
	ruta string
	// directorio es done, in-progress, todo o backlog.
	directorio string
	// posicion es, en todo/, el prefijo del nombre; 0 en los demás.
	posicion int
	cabecera cabeceraDeHito
}

// TestRoadmap vigila docs/roadmap/ (docs/roadmap/README.md): que cada fichero
// está donde tiene que estar y con su forma, que el roadmap se sostiene —un
// hito por identificador, como mucho uno en curso, una posición por hito de la
// cola, dependencias que existen y van antes— y que la lista de hitos de
// workflow.yml es la que se genera de él. Con -regenerar-hitos escribe antes
// esa lista. El último subtest rompe árboles pequeños, uno por defecto, para
// que ninguna comprobación pase en vacío.
func TestRoadmap(t *testing.T) {
	t.Parallel()

	hitos, defectos := leerRoadmap(os.DirFS(raizDelRepositorio + "/" + rutaDelRoadmap))
	enum := enumDeHitos(hitosQueAdmite(hitos))

	if *regenerarHitos {
		escribirEnum(t, enum)
	}

	t.Run("estructura", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, append(defectos, defectosDeLosHitos(hitos)...), "docs/roadmap/ no tiene defectos")
		assert.Greater(t, len(hitos), 20, "docs/roadmap/ tiene más de veinte hitos")
	})

	t.Run("enum", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, enum, lineaDelEnum.FindString(leerDelRepositorio(t, rutaDelWorkflow)),
			"la lista de hitos de workflow.yml es la de docs/roadmap/: ejecuta make roadmap-sync")
	})

	t.Run("defectos", func(t *testing.T) {
		t.Parallel()

		for _, caso := range casosDeDefectos() {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				hitosDelCaso, defectosDelCaso := leerRoadmap(caso.arbol)
				todos := strings.Join(append(defectosDelCaso, defectosDeLosHitos(hitosDelCaso)...), "\n")

				if caso.defecto == "" {
					assert.Empty(t, todos)

					return
				}

				assert.Contains(t, todos, caso.defecto)
			})
		}
	})
}

// casoDeDefecto es un árbol de docs/roadmap/ y el defecto que tiene que dar;
// vacío, si no tiene ninguno.
type casoDeDefecto struct {
	nombre  string
	arbol   fstest.MapFS
	defecto string
}

// casosDeDefectos son un árbol sano y uno por cada defecto que vigila
// TestRoadmap.
func casosDeDefectos() []casoDeDefecto {
	hito := func(id, extra, titulo string) *fstest.MapFile {
		return &fstest.MapFile{Data: []byte("---\nid: " + id + "\nfase: \"1\"\n" + extra + "---\n\n# " + titulo + "\n\nEntrada.\n")}
	}

	sano := func() fstest.MapFS {
		return fstest.MapFS{
			"README.md":                   {Data: []byte("# Roadmap\n")},
			"done/H1-uno.md":              hito("H1", "", "H1 · Uno"),
			"in-progress/H2-dos.md":       hito("H2", "depende_de: [H1]\n", "H2 · Dos"),
			"todo/010-H3-tres.md":         hito("H3", "depende_de: [H2]\n", "H3 · Tres"),
			"todo/020-H3.1-tres-y-uno.md": hito("H3.1", "", "H3.1 · Tres y uno"),
			"backlog/fiscal.md":           {Data: []byte("# Fiscal\n\nUna línea.\n")},
			"backlog/grupo/H9-nueve.md":   hito("H9", "", "H9 · Nueve"),
		}
	}

	con := func(cambios fstest.MapFS) fstest.MapFS {
		arbol := sano()
		maps.Copy(arbol, cambios)

		return arbol
	}

	return []casoDeDefecto{
		{"sano", con(nil), ""},
		{"sin-extension", con(fstest.MapFS{"todo/030-H4-cuatro": hito("H4", "", "H4 · Cuatro")}), "no es un fichero .md"},
		{"fuera-de-las-carpetas", con(fstest.MapFS{"otros/H4-cuatro.md": hito("H4", "", "H4 · Cuatro")}), "fuera de done/"},
		{"cola-sin-posicion", con(fstest.MapFS{"todo/H4-cuatro.md": hito("H4", "", "H4 · Cuatro")}), "no tiene la forma NNN-H<n>-slug.md"},
		{"hecho-con-posicion", con(fstest.MapFS{"done/010-H4-cuatro.md": hito("H4", "", "H4 · Cuatro")}), "no tiene la forma H<n>-slug.md"},
		{"sin-cabecera", con(fstest.MapFS{"done/H1-uno.md": {Data: []byte("# H1 · Uno\n")}}), errSinCabecera.Error()},
		{"campo-desconocido", con(fstest.MapFS{"done/H1-uno.md": hito("H1", "orden: 3\n", "H1 · Uno")}), "field orden not found"},
		{"id-distinto", con(fstest.MapFS{"done/H1-uno.md": hito("H5", "", "H5 · Uno")}), "el id H5 no es el del nombre, H1"},
		{"sin-fase", con(fstest.MapFS{"done/H1-uno.md": {Data: []byte("---\nid: H1\n---\n\n# H1 · Uno\n")}}), "no tiene fase"},
		{"titulo-de-otro", con(fstest.MapFS{"done/H1-uno.md": hito("H1", "", "H5 · Uno")}), "no empieza por «# H1 · »"},
		{"titulo-vacio", con(fstest.MapFS{"done/H1-uno.md": hito("H1", "", "H1 · ")}), "no empieza por «# H1 · »"},
		{"repetido", con(fstest.MapFS{"todo/030-H1-otra-vez.md": hito("H1", "", "H1 · Otra vez")}), "el hito H1 tiene dos ficheros"},
		{"dos-en-curso", con(fstest.MapFS{"in-progress/H4-cuatro.md": hito("H4", "", "H4 · Cuatro")}), "in-progress/ tiene 2 hitos"},
		{"posicion-repetida", con(fstest.MapFS{"todo/010-H4-cuatro.md": hito("H4", "", "H4 · Cuatro")}), "la posición 010 la tienen dos hitos"},
		{"dependencia-que-no-existe", con(fstest.MapFS{"todo/010-H3-tres.md": hito("H3", "depende_de: [H7]\n", "H3 · Tres")}), "depende de H7, que no tiene fichero"},
		{"dependencia-posterior", con(fstest.MapFS{"todo/010-H3-tres.md": hito("H3", "depende_de: [H3.1]\n", "H3 · Tres")}), "depende de H3.1, que va después en la cola"},
		{"dependencia-del-backlog", con(fstest.MapFS{"todo/010-H3-tres.md": hito("H3", "depende_de: [H9]\n", "H3 · Tres")}), "depende de H9, que va después en la cola"},
	}
}

// leerRoadmap lee los ficheros de hito de un árbol de docs/roadmap/: los de
// done/, in-progress/ y todo/, y los del backlog que llevan cabecera. Los
// README.md y los demás ficheros del backlog son texto libre. Devuelve lo leído
// y los defectos de cada fichero, con su ruta.
func leerRoadmap(arbol fs.FS) ([]hitoDelRoadmap, []string) {
	var (
		hitos    []hitoDelRoadmap
		defectos []string
	)

	err := fs.WalkDir(arbol, ".", func(ruta string, entrada fs.DirEntry, err error) error {
		if err != nil || entrada.IsDir() || path.Base(ruta) == leeme {
			return err
		}

		hito, esHito, err := leerFichero(arbol, ruta)
		if err != nil {
			defectos = append(defectos, ruta+": "+err.Error())
		} else if esHito {
			hitos = append(hitos, hito)
		}

		return nil
	})
	if err != nil {
		defectos = append(defectos, "no se puede recorrer docs/roadmap/: "+err.Error())
	}

	return hitos, defectos
}

// leerFichero lee un fichero del roadmap. esHito es falso en un fichero del
// backlog sin cabecera, que es texto libre.
func leerFichero(arbol fs.FS, ruta string) (hitoDelRoadmap, bool, error) {
	if path.Ext(ruta) != ".md" {
		return hitoDelRoadmap{}, false, errors.New("no es un fichero .md")
	}

	directorio, nombre, _ := strings.Cut(ruta, "/")
	hito := hitoDelRoadmap{ruta: ruta, directorio: directorio}

	contenido, err := fs.ReadFile(arbol, ruta)
	if err != nil {
		return hitoDelRoadmap{}, false, err
	}

	var id string

	switch directorio {
	case dirDone, dirInProgress:
		m := nombreDeHito.FindStringSubmatch(nombre)
		if m == nil {
			return hitoDelRoadmap{}, false, errors.New("no tiene la forma H<n>-slug.md, sin subdirectorios")
		}

		id = m[1]
	case dirTodo:
		m := nombreEnLaCola.FindStringSubmatch(nombre)
		if m == nil {
			return hitoDelRoadmap{}, false, errors.New("no tiene la forma NNN-H<n>-slug.md, sin subdirectorios")
		}

		hito.posicion, _ = strconv.Atoi(m[1])
		id = m[2]
	case dirBacklog:
		if !bytes.HasPrefix(contenido, []byte(delimitador)) {
			return hitoDelRoadmap{}, false, nil
		}

		m := nombreDeHito.FindStringSubmatch(path.Base(nombre))
		if m == nil {
			return hitoDelRoadmap{}, false, errors.New("lleva cabecera de hito y no tiene la forma H<n>-slug.md")
		}

		id = m[1]
	default:
		return hitoDelRoadmap{}, false, errors.New("está fuera de done/, in-progress/, todo/ y backlog/")
	}

	hito.cabecera, err = leerCabecera(contenido, id)

	return hito, err == nil, err
}

// leerCabecera lee la cabecera YAML de un fichero de hito y comprueba que es la
// del hito id y que el texto empieza por «# <id> · <título>».
func leerCabecera(contenido []byte, id string) (cabeceraDeHito, error) {
	resto, ok := bytes.CutPrefix(contenido, []byte(delimitador))
	if !ok {
		return cabeceraDeHito{}, errSinCabecera
	}

	yamlDeLaCabecera, texto, ok := bytes.Cut(resto, []byte("\n"+delimitador))
	if !ok {
		return cabeceraDeHito{}, errors.New("su cabecera YAML no se cierra (---)")
	}

	var cabecera cabeceraDeHito

	decodificador := yaml.NewDecoder(bytes.NewReader(yamlDeLaCabecera))
	decodificador.KnownFields(true)

	if err := decodificador.Decode(&cabecera); err != nil {
		return cabeceraDeHito{}, fmt.Errorf("su cabecera no se lee: %w", err)
	}

	switch {
	case cabecera.ID != id:
		return cabeceraDeHito{}, fmt.Errorf("el id %s no es el del nombre, %s", cabecera.ID, id)
	case cabecera.Fase == "":
		return cabeceraDeHito{}, errors.New("no tiene fase")
	}

	primera, _, _ := strings.Cut(strings.TrimLeft(string(texto), "\n"), "\n")
	if titulo, ok := strings.CutPrefix(primera, "# "+id+" · "); !ok || strings.TrimSpace(titulo) == "" {
		return cabeceraDeHito{}, fmt.Errorf("su texto no empieza por «# %s · » y un título", id)
	}

	return cabecera, nil
}

// defectosDeLosHitos son los defectos del roadmap entero: un identificador con
// dos ficheros, más de un hito en curso, una posición de la cola repetida y una
// dependencia que no existe o que no va antes.
func defectosDeLosHitos(hitos []hitoDelRoadmap) []string {
	var defectos []string

	porID := map[string]hitoDelRoadmap{}
	porPosicion := map[int]string{}
	enCurso := 0

	for _, hito := range hitos {
		if otro, repetido := porID[hito.cabecera.ID]; repetido {
			defectos = append(defectos, fmt.Sprintf("el hito %s tiene dos ficheros: %s y %s", hito.cabecera.ID, otro.ruta, hito.ruta))
		}

		porID[hito.cabecera.ID] = hito

		if hito.directorio == dirInProgress {
			enCurso++
		}

		if hito.directorio == dirTodo {
			if otro, repetida := porPosicion[hito.posicion]; repetida {
				defectos = append(defectos, fmt.Sprintf("la posición %03d la tienen dos hitos: %s y %s", hito.posicion, otro, hito.ruta))
			}

			porPosicion[hito.posicion] = hito.ruta
		}
	}

	if enCurso > 1 {
		defectos = append(defectos, fmt.Sprintf("in-progress/ tiene %d hitos, y como mucho tiene uno", enCurso))
	}

	for _, hito := range hitos {
		defectos = append(defectos, defectosDeLasDependencias(hito, porID)...)
	}

	return defectos
}

// defectosDeLasDependencias son los de las dependencias de un hito: cada una
// tiene fichero y, si el hito está en la cola, está hecha, en curso o antes en
// la cola.
func defectosDeLasDependencias(hito hitoDelRoadmap, porID map[string]hitoDelRoadmap) []string {
	var defectos []string

	for _, id := range hito.cabecera.DependeDe {
		dependencia, existe := porID[id]

		switch {
		case !existe:
			defectos = append(defectos, fmt.Sprintf("%s depende de %s, que no tiene fichero", hito.ruta, id))
		case hito.directorio != dirTodo, dependencia.directorio == dirDone, dependencia.directorio == dirInProgress:
		case dependencia.directorio != dirTodo || dependencia.posicion >= hito.posicion:
			defectos = append(defectos, fmt.Sprintf("%s depende de %s, que va después en la cola", hito.ruta, id))
		}
	}

	return defectos
}

// hitosQueAdmite son los hitos que admite el input hito de workflow.yml: los de
// done/, in-progress/ y todo/, en orden de número. Los del backlog no se
// pueden lanzar hasta que pasan a todo/.
func hitosQueAdmite(hitos []hitoDelRoadmap) []string {
	var ids []string

	for _, hito := range hitos {
		if hito.directorio != dirBacklog {
			ids = append(ids, hito.cabecera.ID)
		}
	}

	slices.SortFunc(ids, compararHitos)

	return ids
}

// compararHitos ordena dos identificadores por sus números: H5 < H5.1 < H10.
func compararHitos(a, b string) int {
	numeros := func(id string) []int {
		var partes []int

		for parte := range strings.SplitSeq(strings.TrimPrefix(id, "H"), ".") {
			n, _ := strconv.Atoi(parte)
			partes = append(partes, n)
		}

		return partes
	}

	return cmp.Or(slices.Compare(numeros(a), numeros(b)), strings.Compare(a, b))
}

// enumDeHitos es la línea de workflow.yml con la lista de hitos.
func enumDeHitos(ids []string) string {
	citados := make([]string, 0, len(ids))
	for _, id := range ids {
		citados = append(citados, strconv.Quote(id))
	}

	return "    enum: [" + strings.Join(citados, ",") + "]"
}

// escribirEnum sustituye en workflow.yml la línea de la lista de hitos.
func escribirEnum(t *testing.T, enum string) {
	t.Helper()

	raiz, err := os.OpenRoot(raizDelRepositorio)
	require.NoError(t, err)

	defer func() { require.NoError(t, raiz.Close()) }()

	flujo, err := raiz.ReadFile(rutaDelWorkflow)
	require.NoError(t, err)
	require.Len(t, lineaDelEnum.FindAll(flujo, -1), 1, "workflow.yml tiene una sola línea con la lista de hitos")

	require.NoError(t, raiz.WriteFile(rutaDelWorkflow, lineaDelEnum.ReplaceAllLiteral(flujo, []byte(enum)), 0o600))
}
