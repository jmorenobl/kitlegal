package evals

import (
	"bytes"
	"encoding/json/v2"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/app"
	"github.com/jmorenobl/kitlegal/internal/cache"
	"github.com/jmorenobl/kitlegal/internal/core/plazos"
	"github.com/jmorenobl/kitlegal/internal/skills"
	"github.com/jmorenobl/kitlegal/internal/source/boe"
)

// Lo que estos tests leen del repositorio (contrato evals-y-grabaciones §3.1 y
// §3.4).
const (
	// nombreDelManifiestoDelRepositorio es el nombre de cada manifiesto de
	// grabación real, en el directorio que contiene su conjunto:
	// GrabacionesDeH5 y GrabacionesDeH9.
	nombreDelManifiestoDelRepositorio = "grabaciones.json"

	// entradasMinimasDelManifiesto son las diez normas de H5: el manifiesto real
	// tiene al menos una entrada por cada una (plan.md, obligación 12).
	entradasMinimasDelManifiesto = 10

	// entradasMinimasDelManifiestoDeH9 son las siete normas de los preceptos de
	// plazos: LPAC, LJCA, LOPJ, LTAIBG, LGT, Código Civil y el texto refundido
	// de la Ley sobre Tráfico (contracts/datos-de-plazos.md §4.1 de H9).
	entradasMinimasDelManifiestoDeH9 = 7

	// grabacionDeRobots es el único nombre de fichero que comparten, dos a dos,
	// los conjuntos de grabaciones de H4, de H5 y de H9: la grabación de
	// robots.txt, que la reproducción no usa (data-model §7.3).
	grabacionDeRobots = "GET_https_www.boe.es_robots.txt.json"
)

// Trozos de los manifiestos sintéticos de TestManifiestoDeGrabaciones: el
// principio del documento hasta la lista de normas, su final y las entradas
// válidas de la LPAC, con dos bloques, y de la CE, sin bloques.
const (
	inicioDelManifiesto = `{"fuente": "boe.legislacion-consolidada", "normas": [`
	finDelManifiesto    = `]}`

	entradaDeLaLPAC = `{"busqueda": "procedimiento común", "titulo_empieza_por": "Ley 39/2015,", ` +
		`"bloques": ["a21", "a22"], "para": "01-lpac-articulo-21.yaml: el art. 21 que se cita"}`
	entradaDeLaCE = `{"busqueda": "constitución española", "titulo_empieza_por": "Constitución Española", ` +
		`"para": "SKILL.md: la norma que se nombra"}`
)

// documentoNoValido es el fragmento de todo error del punto 1 del contrato
// evals-y-grabaciones §3.1, el del documento.
const documentoNoValido = "manifiesto de grabación: no es un documento válido: "

// TestManifiestoDeGrabaciones fija LeerManifiesto (contrato evals-y-grabaciones
// §3.1; research.md D11 y V55): un manifiesto válido se lee entero, con sus
// entradas en orden, y cada defecto da, con un Manifiesto vacío, un error que
// dice que es del manifiesto y nombra lo que falla —el miembro, la fuente, la
// entrada por su posición y su prefijo, el campo o el bloque, o las dos entradas
// de una norma repetida—. Los contenidos son sintéticos salvo el del subtest
// repositorio, que lee el manifiesto real y le exige, además de leerse sin
// error, al menos una entrada por cada norma de H5.
func TestManifiestoDeGrabaciones(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre     string
		documento  string
		leido      Manifiesto
		error      string
		fragmentos []string
	}{
		{
			nombre:    "valido",
			documento: inicioDelManifiesto + entradaDeLaLPAC + ", " + entradaDeLaCE + finDelManifiesto,
			leido: Manifiesto{
				Fuente: "boe.legislacion-consolidada",
				Normas: []EntradaDelManifiesto{
					{
						Busqueda:         "procedimiento común",
						TituloEmpiezaPor: "Ley 39/2015,",
						Bloques:          []string{"a21", "a22"},
						Para:             "01-lpac-articulo-21.yaml: el art. 21 que se cita",
					},
					{
						Busqueda:         "constitución española",
						TituloEmpiezaPor: "Constitución Española",
						Para:             "SKILL.md: la norma que se nombra",
					},
				},
			},
		},
		{
			nombre: "clave-desconocida",
			documento: inicioDelManifiesto +
				`{"busqueda": "procedimiento común", "titulo_empieza_por": "Ley 39/2015,", ` +
				`"norma": "BOE-A-2015-10565", "para": "01-lpac-articulo-21.yaml: el art. 21 que se cita"}` +
				finDelManifiesto,
			fragmentos: []string{documentoNoValido, `"norma"`, `"/normas/0"`},
		},
		{
			// Con encoding/json se quedaría con el segundo valor sin error (V55).
			nombre: "clave-repetida",
			documento: inicioDelManifiesto +
				`{"busqueda": "procedimiento común", "titulo_empieza_por": "Ley 39/2015,", ` +
				`"busqueda": "ley del procedimiento", "para": "01-lpac-articulo-21.yaml: el art. 21 que se cita"}` +
				finDelManifiesto,
			fragmentos: []string{documentoNoValido, `"busqueda"`, `"/normas/0"`},
		},
		{
			nombre:     "datos-tras-el-valor",
			documento:  inicioDelManifiesto + entradaDeLaLPAC + finDelManifiesto + "\n{}\n",
			fragmentos: []string{documentoNoValido},
		},
		{
			nombre:    "otra-fuente",
			documento: `{"fuente": "boe.otra", "normas": [` + entradaDeLaLPAC + finDelManifiesto,
			error:     `manifiesto de grabación: la fuente es "boe.otra" y tiene que ser "boe.legislacion-consolidada"`,
		},
		{
			nombre:    "sin-normas",
			documento: inicioDelManifiesto + finDelManifiesto,
			error:     "manifiesto de grabación: normas no tiene ninguna entrada",
		},
		{
			nombre: "busqueda-vacia",
			documento: inicioDelManifiesto + entradaDeLaLPAC + ", " +
				`{"busqueda": "", "titulo_empieza_por": "Ley 7/1985,", ` +
				`"bloques": ["a22"], "para": "03-lrbrl-pleno.yaml: el art. 22 que se cita"}` +
				finDelManifiesto,
			error: `manifiesto de grabación: entrada 2 ("Ley 7/1985,"): el campo busqueda está vacío`,
		},
		{
			// El prefijo vacío es prefijo de cualquier otro: la entrada se comprueba
			// antes que la norma repetida.
			nombre: "prefijo-vacio",
			documento: inicioDelManifiesto + entradaDeLaLPAC + ", " +
				`{"busqueda": "bases del régimen local", "titulo_empieza_por": "", ` +
				`"bloques": ["a22"], "para": "03-lrbrl-pleno.yaml: el art. 22 que se cita"}` +
				finDelManifiesto,
			error: `manifiesto de grabación: entrada 2 (""): el campo titulo_empieza_por está vacío`,
		},
		{
			nombre: "para-vacio",
			documento: inicioDelManifiesto + entradaDeLaLPAC + ", " +
				`{"busqueda": "bases del régimen local", "titulo_empieza_por": "Ley 7/1985,", ` +
				`"bloques": ["a22"], "para": ""}` +
				finDelManifiesto,
			error: `manifiesto de grabación: entrada 2 ("Ley 7/1985,"): el campo para está vacío`,
		},
		{
			nombre: "bloque-mal-formado",
			documento: inicioDelManifiesto + entradaDeLaLPAC + ", " +
				`{"busqueda": "bases del régimen local", "titulo_empieza_por": "Ley 7/1985,", ` +
				`"bloques": ["a22", ".a1"], "para": "03-lrbrl-pleno.yaml: el art. 22 que se cita"}` +
				finDelManifiesto,
			fragmentos: []string{
				`manifiesto de grabación: entrada 2 ("Ley 7/1985,"): el bloque ".a1" no tiene la forma de un id de bloque`,
			},
		},
		{
			nombre: "bloque-repetido",
			documento: inicioDelManifiesto + entradaDeLaLPAC + ", " +
				`{"busqueda": "bases del régimen local", "titulo_empieza_por": "Ley 7/1985,", ` +
				`"bloques": ["a21", "a22", "a21"], "para": "03-lrbrl-pleno.yaml: el art. 22 que se cita"}` +
				finDelManifiesto,
			error: `manifiesto de grabación: entrada 2 ("Ley 7/1985,"): el bloque "a21" está repetido`,
		},
		{
			nombre: "prefijo-repetido",
			documento: inicioDelManifiesto + entradaDeLaLPAC + ", " + entradaDeLaCE + ", " +
				`{"busqueda": "ley del procedimiento", "titulo_empieza_por": "Ley 39/2015,", ` +
				`"bloques": ["a22"], "para": "02-lpac-articulo-22.yaml: el art. 22 que se cita"}` +
				finDelManifiesto,
			error: `manifiesto de grabación: entrada 1 ("Ley 39/2015,") y entrada 3 ("Ley 39/2015,"): ` +
				"la misma norma repetida, con el mismo prefijo",
		},
		{
			nombre: "prefijo-de-otro-prefijo",
			documento: inicioDelManifiesto + entradaDeLaLPAC + ", " +
				`{"busqueda": "ley del procedimiento", "titulo_empieza_por": "Ley 39/2015, de 1 de octubre", ` +
				`"para": "SKILL.md: la norma que se nombra"}` +
				finDelManifiesto,
			error: `manifiesto de grabación: entrada 1 ("Ley 39/2015,") y entrada 2 ("Ley 39/2015, de 1 de octubre"): ` +
				"la misma norma repetida, porque el prefijo de una empieza por el de la otra",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			leido, err := LeerManifiesto([]byte(caso.documento))
			if caso.error == "" && len(caso.fragmentos) == 0 {
				require.NoError(t, err)
				assert.Equal(t, caso.leido, leido)

				return
			}

			require.Error(t, err)
			assert.Zero(t, leido, "un manifiesto con un defecto no se entrega a medias")
			assert.True(t, strings.HasPrefix(err.Error(), "manifiesto de grabación: "),
				"el error dice que es del manifiesto: %q", err.Error())

			if caso.error != "" {
				require.EqualError(t, err, caso.error)
			}

			for _, fragmento := range caso.fragmentos {
				assert.ErrorContains(t, err, fragmento)
			}
		})
	}

	// Los dos manifiestos reales, cada uno junto a su conjunto: el de H5 y el de
	// H9, que graba grabar_datos (research D15 de H9).
	repositorio := []struct {
		nombre   string
		conjunto string
		minimo   int
		normas   string
	}{
		{nombre: "repositorio", conjunto: GrabacionesDeH5, minimo: entradasMinimasDelManifiesto, normas: "H5"},
		{
			nombre:   "repositorio-h9",
			conjunto: GrabacionesDeH9,
			minimo:   entradasMinimasDelManifiestoDeH9,
			normas:   "los preceptos de plazos de H9",
		},
	}

	for _, manifiesto := range repositorio {
		t.Run(manifiesto.nombre, func(t *testing.T) {
			t.Parallel()

			ruta := rutaDelManifiestoJuntoA(manifiesto.conjunto)

			contenido, err := os.ReadFile(filepath.Clean(ruta))
			require.NoError(t, err, "el manifiesto de grabación del repositorio %s no se puede leer", ruta)

			leido, err := LeerManifiesto(contenido)
			require.NoError(t, err, "el manifiesto de grabación del repositorio, %s", ruta)
			assert.GreaterOrEqual(t, len(leido.Normas), manifiesto.minimo,
				"el manifiesto de grabación del repositorio, %s, tiene una entrada por cada norma de %s",
				ruta, manifiesto.normas)
		})
	}
}

// TestGrabacionesSinSolape comprueba que la unión son los conjuntos de
// grabaciones de H4, de H5 y de H9, en ese orden, y que, dos a dos, solo
// comparten el nombre de la grabación de robots.txt (contrato
// evals-y-grabaciones §3.4; data-model §7.3; research D15 de H9). Cualquier
// otro nombre común haría que, en la unión que copia Preparar, la grabación de
// un conjunto sustituyera sin avisar a la de uno anterior, que es de su hito y
// que los siguientes solo leen.
func TestGrabacionesSinSolape(t *testing.T) {
	t.Parallel()

	conjuntos := UnionDeGrabaciones()
	require.Equal(t, []string{GrabacionesDeH4, GrabacionesDeH5, GrabacionesDeH9}, conjuntos,
		"la unión son los conjuntos de H4, H5 y H9, cada uno encima del anterior")

	for posicion, anterior := range conjuntos {
		delAnterior := nombresDeLasGrabaciones(t, anterior)

		for _, posterior := range conjuntos[posicion+1:] {
			var comunes []string

			for _, nombre := range nombresDeLasGrabaciones(t, posterior) {
				if slices.Contains(delAnterior, nombre) {
					comunes = append(comunes, nombre)
				}
			}

			assert.Equal(t, []string{grabacionDeRobots}, comunes,
				"los conjuntos de grabaciones %s y %s solo coinciden en %s", anterior, posterior, grabacionDeRobots)
		}
	}
}

// TestDirectorioDeLosGrafosPrevios fija dónde están los conjuntos de grabaciones
// derivadas con los que se prepara cada grafo previo (contrato evals-y-skill §3
// de H7; research D26): en grafo-previo/, junto al conjunto de grabaciones de las
// evals, y fuera de la unión que reproduce Preparar, que no copia ninguno ni se
// copia desde dentro de él, para que lo derivado llegue al grafo de la sesión y
// nunca a su caché.
func TestDirectorioDeLosGrafosPrevios(t *testing.T) {
	t.Parallel()

	assert.Equal(t, GrafosPrevios, filepath.Join(filepath.Dir(GrabacionesDeH5), "grafo-previo"),
		"cada grafo previo va en grafo-previo/, junto a las grabaciones de las evals")

	for _, conjunto := range UnionDeGrabaciones() {
		desdeLosGrafosPrevios, err := filepath.Rel(GrafosPrevios, conjunto)
		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(desdeLosGrafosPrevios, ".."+string(filepath.Separator)),
			"el conjunto %s de la unión no está dentro de %s", conjunto, GrafosPrevios)

		hastaLosGrafosPrevios, err := filepath.Rel(conjunto, GrafosPrevios)
		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(hastaLosGrafosPrevios, ".."+string(filepath.Separator)),
			"%s no está dentro del conjunto %s de la unión", GrafosPrevios, conjunto)
	}
}

// rutaDelManifiestoJuntoA es la del manifiesto de grabación real que llena el
// conjunto de grabaciones: está en el directorio que contiene el conjunto.
func rutaDelManifiestoJuntoA(conjunto string) string {
	return filepath.Join(filepath.Dir(conjunto), nombreDelManifiestoDelRepositorio)
}

// Lo que TestIdentificadoresDeLasNormas lee del repositorio y escribe en sus
// copias (contrato normas-y-referencias §6).
const (
	// tablaDeNormasDelRepositorio es data/normas.yaml, relativo al directorio de
	// este paquete, que es donde go test ejecuta sus tests (research.md V46).
	tablaDeNormasDelRepositorio = "../../data/normas.yaml"

	// normasMinimasDeLaTabla son las diez normas de H5: la verificación no pasa
	// con una tabla que tenga menos (plan.md, obligación 12).
	normasMinimasDeLaTabla = 10

	// nombreDeLaCopiaDeLaTabla es el nombre de la copia temporal de la tabla de
	// normas que escriben los subtests; la del manifiesto lleva el suyo.
	nombreDeLaCopiaDeLaTabla = "normas.yaml"
)

// Las normas con las que los subtests de TestIdentificadoresDeLasNormas rompen
// sus copias.
const (
	// identificadorDeLaLPAC es el de la norma cuyo identificador o título se
	// cambia, o cuya entrada se retira: la Ley 39/2015, la que fija el hito.
	identificadorDeLaLPAC = "BOE-A-2015-10565"

	// identificadorSinBusqueda es el que toma la LPAC en identificador-cambiado:
	// tiene la forma de un identificador y no es el resultado de ninguna
	// búsqueda grabada, porque ninguna norma es del año 2099.
	identificadorSinBusqueda = "BOE-A-2099-99999"
)

// TestIdentificadoresDeLasNormas repite sin red la verificación con boe buscar
// de los identificadores de data/normas.yaml (contrato normas-y-referencias §6;
// FR-023, FR-024, FR-043, SC-008; US6, escenario 4): sobre la reproducción de las
// grabaciones de H4 y de H5, cada norma de la tabla la resuelve exactamente una
// entrada del manifiesto de grabación, con su mismo identificador y su mismo
// título, y la tabla tiene al menos las diez normas de H5. Los subtests rompen
// copias temporales de la tabla y del manifiesto y exigen que la verificación
// falle nombrando la norma: con su identificador cambiado, con su título
// cambiado, sin la entrada que la resuelve y con una norma que da la búsqueda de
// otra entrada sin ser la que esa entrada resuelve.
func TestIdentificadoresDeLasNormas(t *testing.T) {
	t.Parallel()

	reproduccion := copiaDeLaUnionDeGrabaciones(t)
	rutaDelManifiesto := rutaDelManifiestoJuntoA(GrabacionesDeH5)

	normas, err := verificarIdentificadores(t, reproduccion, rutaDelManifiesto, tablaDeNormasDelRepositorio)
	require.NoError(t, err, "los identificadores de %s contra las búsquedas grabadas del manifiesto %s",
		tablaDeNormasDelRepositorio, rutaDelManifiesto)
	require.GreaterOrEqual(t, len(normas), normasMinimasDeLaTabla,
		"%s tiene al menos las diez normas de H5", tablaDeNormasDelRepositorio)

	tabla := string(contenidoDelFichero(t, tablaDeNormasDelRepositorio))
	contenidoDelManifiesto := string(contenidoDelFichero(t, rutaDelManifiesto))

	manifiesto, err := LeerManifiesto([]byte(contenidoDelManifiesto))
	require.NoError(t, err)

	lpac := normaDeLaTabla(t, normas, identificadorDeLaLPAC)
	entradaDeLaLPAC := entradaConPrefijoDe(t, manifiesto, lpac.Titulo)
	otra := otraNormaDeLaBusqueda(t, reproduccion, manifiesto, entradaDeLaLPAC, normas)

	tituloCambiado := strings.TrimSuffix(lpac.Titulo, ".")
	require.NotEqual(t, lpac.Titulo, tituloCambiado, "el título de la LPAC termina en punto: sin él, cambia")

	casos := []struct {
		nombre     string
		tabla      string
		manifiesto string
		error      string
	}{
		{
			nombre:     "identificador-cambiado",
			tabla:      sustituida(t, tabla, "  "+identificadorDeLaLPAC+":\n", "  "+identificadorSinBusqueda+":\n"),
			manifiesto: contenidoDelManifiesto,
			error:      identificadorSinBusqueda + ": ninguna entrada del manifiesto lo resuelve",
		},
		{
			nombre:     "titulo-cambiado",
			tabla:      sustituida(t, tabla, lpac.Titulo, tituloCambiado),
			manifiesto: contenidoDelManifiesto,
			error:      identificadorDeLaLPAC + ": el título no coincide con la búsqueda grabada: " + lpac.Titulo,
		},
		{
			nombre:     "norma-sin-entrada",
			tabla:      tabla,
			manifiesto: manifiestoSinEntrada(t, manifiesto, entradaDeLaLPAC),
			error:      identificadorDeLaLPAC + ": ninguna entrada del manifiesto lo resuelve",
		},
		{
			// La búsqueda grabada de la entrada de la LPAC da también esta norma,
			// pero la entrada resuelve la LPAC: no cuenta (data-model §7.2).
			nombre:     "norma-en-la-busqueda-de-otra-entrada",
			tabla:      tabla + normaEscrita(t, otra, lpac.Materias),
			manifiesto: contenidoDelManifiesto,
			error:      otra.Identificador + ": ninguna entrada del manifiesto lo resuelve",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			copias := t.TempDir()
			copiaDeLaTabla := filepath.Join(copias, nombreDeLaCopiaDeLaTabla)
			copiaDelManifiesto := filepath.Join(copias, nombreDelManifiestoDelRepositorio)

			require.NoError(t, escribirFichero(copiaDeLaTabla, []byte(caso.tabla)))
			require.NoError(t, escribirFichero(copiaDelManifiesto, []byte(caso.manifiesto)))

			_, err := verificarIdentificadores(t, reproduccion, copiaDelManifiesto, copiaDeLaTabla)
			require.EqualError(t, err, caso.error)
		})
	}
}

// normaResuelta es la norma que resuelve una entrada del manifiesto: el único
// resultado de su búsqueda grabada cuyo título empieza por su prefijo.
type normaResuelta struct {
	// entrada nombra la entrada que la resuelve, como nombrarEntrada.
	entrada string

	// resultado es ese resultado, con el identificador y el título grabados.
	resultado boe.ResultadoDeBusqueda
}

// verificarIdentificadores comprueba sin red la tabla de normas de rutaDeLaTabla
// contra las búsquedas grabadas del manifiesto de rutaDelManifiesto,
// reproducidas desde el directorio reproduccion (contrato normas-y-referencias
// §6), y devuelve las normas de la tabla y los defectos unidos:
//
//  1. lee el manifiesto con LeerManifiesto y devuelve su error tal cual, antes
//     de reproducir ninguna búsqueda; después lee la tabla con skills.LeerNormas,
//     con su error igual;
//  2. resuelve cada entrada del manifiesto con resolverEntrada, que da un defecto
//     por cada entrada que no se resuelve;
//  3. exige a cada norma de la tabla exactamente una entrada que la resuelva, con
//     su identificador y su título, con comprobarNorma.
//
// Lo que impide verificar, un fichero que no se lee o el registro que no se
// monta, hace fallar el test.
func verificarIdentificadores(
	t *testing.T, reproduccion, rutaDelManifiesto, rutaDeLaTabla string,
) ([]skills.Norma, error) {
	t.Helper()

	manifiesto, err := LeerManifiesto(contenidoDelFichero(t, rutaDelManifiesto))
	if err != nil {
		return nil, err
	}

	normas, err := skills.LeerNormas(contenidoDelFichero(t, rutaDeLaTabla))
	if err != nil {
		return nil, err
	}

	registro := registroDeLaReproduccion(t, reproduccion)
	resueltas := make(map[string][]normaResuelta, len(manifiesto.Normas))

	var defectos []error

	for indice, entrada := range manifiesto.Normas {
		resuelta, err := resolverEntrada(registro, indice, entrada)
		if err != nil {
			defectos = append(defectos, err)

			continue
		}

		identificador := resuelta.resultado.Identificador
		resueltas[identificador] = append(resueltas[identificador], resuelta)
	}

	for _, norma := range normas {
		if err := comprobarNorma(norma, resueltas[norma.Identificador]); err != nil {
			defectos = append(defectos, err)
		}
	}

	return normas, errors.Join(defectos...)
}

// resolverEntrada reproduce boe buscar con la búsqueda de la entrada y devuelve
// la norma que resuelve: el único resultado cuyo título empieza por su prefijo,
// comparado byte a byte con strings.HasPrefix, como en TestGrabarEvals (contrato
// evals-y-grabaciones §3.2). La búsqueda que termina con un código distinto de
// 0, o cuya salida no se lee, es un defecto que nombra la entrada y la orden, y
// nunca cuenta como una búsqueda sin resultados; ninguno o más de un resultado
// con el prefijo, un defecto que nombra la entrada y los títulos.
func resolverEntrada(registro *app.Registro, indice int, entrada EntradaDelManifiesto) (normaResuelta, error) {
	nombre := nombrarEntrada(indice, entrada)
	busqueda := busquedaDe(entrada)

	resultados, err := reproducirBusqueda(registro, busqueda)
	if err != nil {
		return normaResuelta{}, fmt.Errorf("%s: %w", nombre, err)
	}

	elegidos := slices.DeleteFunc(slices.Clone(resultados), func(resultado boe.ResultadoDeBusqueda) bool {
		return !strings.HasPrefix(resultado.Titulo, entrada.TituloEmpiezaPor)
	})
	if len(elegidos) != 1 {
		titulos := make([]string, 0, len(resultados))
		for _, resultado := range resultados {
			titulos = append(titulos, resultado.Titulo)
		}

		return normaResuelta{}, fmt.Errorf("%s: %d resultados de «%s» tienen un título que empieza por el prefijo, "+
			"y tiene que ser uno solo; títulos: %q", nombre, len(elegidos), ordenDe(busqueda), titulos)
	}

	return normaResuelta{entrada: nombre, resultado: elegidos[0]}, nil
}

// reproducirBusqueda ejecuta la búsqueda con app.Main y --json sobre el registro
// y devuelve sus resultados. Un código distinto de 0 es un error con la orden, su
// código y su mensaje, como los presenta Falta.
func reproducirBusqueda(registro *app.Registro, busqueda Consulta) ([]boe.ResultadoDeBusqueda, error) {
	var salida, errores bytes.Buffer

	argv := slices.Concat([]string{programaDeLasConsultas, busqueda.Applet, busqueda.Verbo}, busqueda.Argumentos,
		[]string{"--json"})
	if codigo := app.Main(argv, registro, &salida, &errores,
		sinDatosDeConstruccion, sinDatosDeConstruccion, sinDatosDeConstruccion); codigo != 0 {
		falta := Falta{Consulta: busqueda, Codigo: codigo, Mensaje: strings.TrimSuffix(errores.String(), "\n")}

		return nil, errors.New(falta.String())
	}

	var sobre struct {
		Data []boe.ResultadoDeBusqueda `json:"data"`
	}
	if err := json.Unmarshal(salida.Bytes(), &sobre); err != nil {
		return nil, fmt.Errorf("la salida de «%s» no es un sobre con los resultados de la búsqueda: %w",
			ordenDe(busqueda), err)
	}

	return sobre.Data, nil
}

// comprobarNorma exige que exactamente una de las entradas que resuelven el
// identificador de la norma la resuelva, y con su mismo título. Que la norma
// esté entre los resultados de la búsqueda de otra entrada sin ser el que esa
// entrada resuelve no cuenta: no está en resueltas (data-model §7.2). Dos
// entradas solo pueden resolver el mismo identificador si sus búsquedas lo dan
// con títulos distintos, porque LeerManifiesto rechaza dos prefijos de un mismo
// título.
func comprobarNorma(norma skills.Norma, resueltas []normaResuelta) error {
	switch len(resueltas) {
	case 0:
		return fmt.Errorf("%s: ninguna entrada del manifiesto lo resuelve", norma.Identificador)
	case 1:
		if grabado := resueltas[0].resultado.Titulo; grabado != norma.Titulo {
			return fmt.Errorf("%s: el título no coincide con la búsqueda grabada: %s", norma.Identificador, grabado)
		}

		return nil
	default:
		entradas := make([]string, 0, len(resueltas))
		for _, resuelta := range resueltas {
			entradas = append(entradas, resuelta.entrada)
		}

		return fmt.Errorf("%s: lo resuelven %d entradas del manifiesto, y tiene que ser una sola: %s",
			norma.Identificador, len(resueltas), strings.Join(entradas, ", "))
	}
}

// busquedaDe es la invocación de boe buscar con la búsqueda de la entrada.
func busquedaDe(entrada EntradaDelManifiesto) Consulta {
	return Consulta{Applet: appletDeLasNormas, Verbo: verboBuscar, Argumentos: []string{entrada.Busqueda}}
}

// copiaDeLaUnionDeGrabaciones copia en un directorio temporal del test los
// conjuntos de UnionDeGrabaciones, cada uno encima del anterior, como Preparar,
// y devuelve su ruta: la reproducción de la unión de H4 y H5 (contrato
// evals-y-grabaciones §5.1).
func copiaDeLaUnionDeGrabaciones(t *testing.T) string {
	t.Helper()

	copia := t.TempDir()
	for _, conjunto := range UnionDeGrabaciones() {
		require.NoError(t, copiarGrabaciones(conjunto, copia))
	}

	return copia
}

// registroDeLaReproduccion es el registro con el que se reproducen las
// búsquedas: el applet boe sobre httpx.Replay del directorio reproduccion, que no
// abre ninguna conexión, con la caché en un directorio temporal del test
// (contrato evals-y-grabaciones §5.1).
func registroDeLaReproduccion(t *testing.T, reproduccion string) *app.Registro {
	t.Helper()

	registro, err := registroDeBoe(reproduccion, cache.ConDirectorio(t.TempDir()))
	require.NoError(t, err)

	return registro
}

// contenidoDelFichero es el contenido entero del fichero de la ruta, que tiene
// que poder leerse.
func contenidoDelFichero(t *testing.T, ruta string) []byte {
	t.Helper()

	contenido, err := leerFichero(ruta)
	require.NoError(t, err, "el fichero %s no se puede leer", ruta)

	return contenido
}

// normaDeLaTabla es la norma del identificador, que tiene que estar entre las
// normas.
func normaDeLaTabla(t *testing.T, normas []skills.Norma, identificador string) skills.Norma {
	t.Helper()

	posicion := slices.IndexFunc(normas, func(norma skills.Norma) bool { return norma.Identificador == identificador })
	require.GreaterOrEqual(t, posicion, 0, "la tabla de normas tiene la norma %s", identificador)

	return normas[posicion]
}

// entradaConPrefijoDe es el índice de la única entrada del manifiesto cuyo
// prefijo lo es del título: la que resuelve la norma de ese título si su
// búsqueda la da.
func entradaConPrefijoDe(t *testing.T, manifiesto Manifiesto, titulo string) int {
	t.Helper()

	var indices []int

	for indice, entrada := range manifiesto.Normas {
		if strings.HasPrefix(titulo, entrada.TituloEmpiezaPor) {
			indices = append(indices, indice)
		}
	}

	require.Len(t, indices, 1, "una sola entrada del manifiesto tiene un prefijo del título %q", titulo)

	return indices[0]
}

// otraNormaDeLaBusqueda es el primer resultado de la búsqueda grabada de la
// entrada del manifiesto en la posición indice que no es ninguna de las normas y
// cuyo título no empieza por el prefijo de ninguna entrada del manifiesto: la
// búsqueda lo da, pero la entrada resuelve otra norma y ninguna entrada lo
// resuelve a él. Sin la condición del prefijo, la entrada de una norma que
// todavía no está en la tabla lo resolvería, y el subtest que lo añade a la
// tabla fallaría por su premisa y no por lo que prueba (contrato de la eval y la
// grabación §1 de H5.1; research D12). Una búsqueda que no dé ninguno así es un
// fallo del test, que si no pasaría en vacío.
func otraNormaDeLaBusqueda(
	t *testing.T, reproduccion string, manifiesto Manifiesto, indice int, normas []skills.Norma,
) boe.ResultadoDeBusqueda {
	t.Helper()

	entrada := manifiesto.Normas[indice]

	resultados, err := reproducirBusqueda(registroDeLaReproduccion(t, reproduccion), busquedaDe(entrada))
	require.NoError(t, err)

	posicion := slices.IndexFunc(resultados, func(resultado boe.ResultadoDeBusqueda) bool {
		enLaTabla := slices.ContainsFunc(normas, func(norma skills.Norma) bool {
			return norma.Identificador == resultado.Identificador
		})
		conPrefijo := slices.ContainsFunc(manifiesto.Normas, func(otra EntradaDelManifiesto) bool {
			return strings.HasPrefix(resultado.Titulo, otra.TituloEmpiezaPor)
		})

		return !enLaTabla && !conPrefijo
	})
	require.GreaterOrEqual(t, posicion, 0, "la búsqueda grabada %q da alguna norma que no está en la tabla de normas "+
		"y cuyo título no empieza por el prefijo de ninguna entrada del manifiesto", entrada.Busqueda)

	return resultados[posicion]
}

// normaEscrita es la línea de la tabla de normas con el identificador, el
// título y el rango del resultado y las materias, con el valor escrito en JSON,
// que es YAML válido y conserva cada texto tal cual.
func normaEscrita(t *testing.T, resultado boe.ResultadoDeBusqueda, materias []string) string {
	t.Helper()

	valor, err := json.Marshal(struct {
		Titulo   string   `json:"titulo"`
		Rango    string   `json:"rango"`
		Materias []string `json:"materias"`
	}{Titulo: resultado.Titulo, Rango: resultado.Rango, Materias: materias})
	require.NoError(t, err)

	return "  " + resultado.Identificador + ": " + string(valor) + "\n"
}

// manifiestoSinEntrada es el manifiesto escrito en JSON sin la entrada del
// índice.
func manifiestoSinEntrada(t *testing.T, manifiesto Manifiesto, indice int) string {
	t.Helper()

	contenido, err := json.Marshal(Manifiesto{
		Fuente: manifiesto.Fuente,
		Normas: slices.Delete(slices.Clone(manifiesto.Normas), indice, indice+1),
	})
	require.NoError(t, err)

	return string(contenido)
}

// sustituida es el texto con viejo, que tiene que aparecer en él exactamente una
// vez, sustituido por nuevo: un subtest que no cambiara nada, o que cambiara
// algo más, no rompería solo lo que dice.
func sustituida(t *testing.T, texto, viejo, nuevo string) string {
	t.Helper()

	require.Equal(t, 1, strings.Count(texto, viejo), "%q aparece una sola vez en el texto que se cambia", viejo)

	return strings.Replace(texto, viejo, nuevo, 1)
}

// nombresDeLasGrabaciones son los nombres de las entradas del directorio de un
// conjunto de grabaciones, que tiene que poder listarse y no estar vacío: un
// conjunto que no se lee no puede dar un solape vacío.
func nombresDeLasGrabaciones(t *testing.T, conjunto string) []string {
	t.Helper()

	entradas, err := os.ReadDir(conjunto)
	require.NoError(t, err, "el conjunto de grabaciones %s no se puede listar", conjunto)
	require.NotEmpty(t, entradas, "el conjunto de grabaciones %s está vacío", conjunto)

	nombres := make([]string, 0, len(entradas))
	for _, entrada := range entradas {
		nombres = append(nombres, entrada.Name())
	}

	return nombres
}

// Lo que TestPreceptosDeLosPlazos lee del repositorio (contracts/datos-de-plazos.md
// §4.2 de H9), relativo al directorio de este paquete, que es donde go test
// ejecuta sus tests (research.md V46).
const (
	// tiposDelRepositorio es data/plazos.yaml, los tipos de plazo.
	tiposDelRepositorio = "../../data/plazos.yaml"

	// coberturaDelRepositorio es data/cobertura.yaml, la cobertura por materia.
	coberturaDelRepositorio = "../../data/cobertura.yaml"
)

// identificadorDelTextoRefundidoDeTrafico es el del texto refundido de la Ley
// sobre Tráfico, Circulación de Vehículos a Motor y Seguridad Vial, que la
// tabla de las palabras clave necesita para leer su art. 1: el que resuelve la
// búsqueda grabada de su entrada del manifiesto de H9 (research S5 de H9).
// data/cobertura.yaml no lo toma de aquí: TestPreceptosDeLosPlazos le exige el
// que resuelve esa búsqueda.
const identificadorDelTextoRefundidoDeTrafico = "BOE-A-2015-11722"

// palabrasDeUnPrecepto son las palabras clave de un precepto del cómputo o de
// la cobertura: las que tiene que llevar el texto grabado de su bloque.
type palabrasDeUnPrecepto struct {
	// precepto lo nombra en los defectos, como «art. 30 LPAC».
	precepto string

	// norma y bloque son los del bloque que lo contiene.
	norma, bloque string

	// palabras son las que lleva su texto, tal cual.
	palabras []string
}

// palabrasClaveDeLosPreceptos es la tabla de contracts/datos-de-plazos.md §4.2
// de H9, con lo que lo grabado corrigió de lo que el plan recordaba (research
// S4 de H9): el art. 30 LPAC dice de los días que «éstos son hábiles», y el
// art. 182.1 LOPJ no enumera el 24 y el 31 de diciembre. Devuelve una tabla
// nueva en cada llamada.
func palabrasClaveDeLosPreceptos() []palabrasDeUnPrecepto {
	const (
		lpac   = "BOE-A-2015-10565"
		ljca   = "BOE-A-1998-16718"
		lopj   = "BOE-A-1985-12666"
		ltaibg = "BOE-A-2013-12887"
	)

	return []palabrasDeUnPrecepto{
		{precepto: "art. 30 LPAC", norma: lpac, bloque: "a30", palabras: []string{
			"sábados", "domingos", "éstos son hábiles", "mismo día", "último día del mes", "prorrogado",
		}},
		{precepto: "art. 21 LPAC", norma: lpac, bloque: "a21", palabras: []string{"entrada en el registro"}},
		{precepto: "art. 24 LPAC", norma: lpac, bloque: "a24", palabras: []string{"vencimiento del plazo máximo"}},
		{precepto: "DA 1ª LPAC", norma: lpac, bloque: "daprimera", palabras: []string{
			"aplicación de los tributos", "Seguridad Social", "tráfico y seguridad vial", "extranjería y asilo",
		}},
		{precepto: "art. 112 LPAC", norma: lpac, bloque: "a112", palabras: []string{
			"recurso de alzada", "reclamaciones económico-administrativas",
		}},
		{precepto: "art. 128 LJCA", norma: ljca, bloque: "a128", palabras: []string{"agosto"}},
		{precepto: "art. 182 LOPJ", norma: lopj, bloque: "acientoochentaydos", palabras: []string{
			"sábados y domingos", "fiesta nacional", "festivos a efectos laborales",
		}},
		{precepto: "art. 183 LOPJ", norma: lopj, bloque: "acientoochentaytres", palabras: []string{
			"24 de diciembre", "6 de enero",
		}},
		{precepto: "art. 185 LOPJ", norma: lopj, bloque: "acientoochentaycinco", palabras: []string{
			"Código Civil", "prorrogado",
		}},
		{precepto: "art. 5 CC", norma: "BOE-A-1889-4763", bloque: "art5", palabras: []string{"de fecha a fecha"}},
		{precepto: "art. 97 LGT", norma: "BOE-A-2003-23186", bloque: "a97", palabras: []string{
			"aplicación de los tributos",
		}},
		{
			precepto: "art. 1 del texto refundido de Tráfico", norma: identificadorDelTextoRefundidoDeTrafico,
			bloque: "a1", palabras: []string{"sancion"},
		},
		{precepto: "art. 24 LTAIBG", norma: ltaibg, bloque: "a24", palabras: []string{
			"un mes", "Consejo de Transparencia y Buen Gobierno",
		}},
	}
}

// TestPreceptosDeLosPlazos comprueba sin red los preceptos de data/plazos.yaml
// y de data/cobertura.yaml contra lo grabado del BOE (contracts/datos-de-plazos.md
// §4.2 de H9; FR-008, FR-010, FR-030, FR-031; research D16 de H9), con
// verificarPreceptos: cada bloque está grabado, cada precepto que fija la
// duración de un tipo la dice en palabras en su apartado, cada norma de la
// cobertura es la que resuelve su búsqueda grabada y el texto de cada precepto
// de la tabla lleva sus palabras clave. Los subtests rompen copias temporales de
// los dos ficheros o de la tabla y exigen que la verificación falle nombrando
// el precepto y lo que falta: una duración que el texto no dice, un bloque sin
// grabar, una palabra que el texto no lleva y una norma que ninguna búsqueda
// grabada resuelve.
func TestPreceptosDeLosPlazos(t *testing.T) {
	t.Parallel()

	require.NoError(t, verificarPreceptos(t, tiposDelRepositorio, coberturaDelRepositorio, palabrasClaveDeLosPreceptos()),
		"los preceptos de %s y de %s contra lo grabado", tiposDelRepositorio, coberturaDelRepositorio)

	tipos := string(contenidoDelFichero(t, tiposDelRepositorio))
	cobertura := string(contenidoDelFichero(t, coberturaDelRepositorio))

	conOtraPalabra := palabrasClaveDeLosPreceptos()
	posicion := slices.IndexFunc(conOtraPalabra, func(fila palabrasDeUnPrecepto) bool { return fila.bloque == "a128" })
	require.GreaterOrEqual(t, posicion, 0, "la tabla tiene el art. 128 LJCA")
	conOtraPalabra[posicion].palabras = []string{"septiembre"}

	const reposicion = "codigo: reposicion\n    nombre: recurso potestativo de reposición\n    computo: lpac\n" +
		"    duracion: {cantidad: "

	casos := []struct {
		nombre    string
		tipos     string
		cobertura string
		tabla     []palabrasDeUnPrecepto
		defecto   string
	}{
		{
			nombre:    "duracion-que-el-texto-no-dice",
			tipos:     sustituida(t, tipos, reposicion+"1,", reposicion+"2,"),
			cobertura: cobertura,
			tabla:     palabrasClaveDeLosPreceptos(),
			defecto: "reposicion: BOE-A-2015-10565 a124, apartado 124.1: " +
				"el párrafo del apartado no lleva «dos meses»",
		},
		{
			nombre:    "bloque-sin-grabar",
			tipos:     sustituida(t, tipos, "bloque: a128,", "bloque: a129,"),
			cobertura: cobertura,
			tabla:     palabrasClaveDeLosPreceptos(),
			defecto: "contencioso: BOE-A-1998-16718 a129 no está grabado: " +
				"«boe articulo BOE-A-1998-16718 a129» terminó con código ",
		},
		{
			nombre:    "palabra-que-el-texto-no-lleva",
			tipos:     tipos,
			cobertura: cobertura,
			tabla:     conOtraPalabra,
			defecto:   "art. 128 LJCA (BOE-A-1998-16718 a128): el texto grabado no lleva «septiembre»",
		},
		{
			nombre: "norma-sin-busqueda",
			tipos:  tipos,
			cobertura: sustituida(t, cobertura, "identificador: "+identificadorDelTextoRefundidoDeTrafico,
				"identificador: "+identificadorSinBusqueda),
			tabla: palabrasClaveDeLosPreceptos(),
			defecto: "procedimientos sancionadores en materia de tráfico y seguridad vial: " + identificadorSinBusqueda +
				": ninguna entrada del manifiesto de H9 lo resuelve",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			copias := t.TempDir()
			copiaDeLosTipos := filepath.Join(copias, "plazos.yaml")
			copiaDeLaCobertura := filepath.Join(copias, "cobertura.yaml")

			require.NoError(t, escribirFichero(copiaDeLosTipos, []byte(caso.tipos)))
			require.NoError(t, escribirFichero(copiaDeLaCobertura, []byte(caso.cobertura)))

			require.ErrorContains(t, verificarPreceptos(t, copiaDeLosTipos, copiaDeLaCobertura, caso.tabla), caso.defecto)
		})
	}
}

// lecturaDeUnPrecepto es un bloque que nombran data/plazos.yaml,
// data/cobertura.yaml o la tabla de las palabras clave: quién lo nombra —el
// código del tipo, el procedimiento o el precepto de la tabla—, su norma, su
// bloque y su apartado y, si fija la duración del tipo, esa duración.
type lecturaDeUnPrecepto struct {
	quien, norma, bloque, apartado string
	duracion                       *plazos.Duracion
}

// verificarPreceptos comprueba sin red los tipos de rutaDeLosTipos y la
// cobertura de rutaDeLaCobertura contra lo grabado, con la tabla de las
// palabras clave (contracts/datos-de-plazos.md §4.2 de H9), y devuelve los
// defectos unidos:
//
//  1. lee los tipos con plazos.LeerTipos y la cobertura con
//     skills.LeerCobertura, y devuelve su error tal cual;
//  2. prepara una caché desde UnionDeGrabaciones con boe articulo de cada
//     bloque que nombran los dos ficheros y la tabla, y lee de ella el texto de
//     cada uno (leerLosBloques): cada bloque sin grabar es un defecto por cada
//     tipo, procedimiento o fila que lo nombra;
//  3. en cada precepto que fija la duración de un tipo, el párrafo de su
//     apartado la lleva en palabras (comprobarLaDuracion);
//  4. cada norma de la cobertura la resuelve una sola entrada del manifiesto de
//     H9 (verificarLasNormasDeLaCobertura);
//  5. el texto de cada fila de la tabla lleva cada una de sus palabras
//     (comprobarLasPalabras).
//
// Lo que impide verificar, un fichero que no se lee o una caché que no se
// prepara, hace fallar el test.
func verificarPreceptos(t *testing.T, rutaDeLosTipos, rutaDeLaCobertura string, tabla []palabrasDeUnPrecepto) error {
	t.Helper()

	tipos, err := plazos.LeerTipos(contenidoDelFichero(t, rutaDeLosTipos))
	if err != nil {
		return err
	}

	cobertura, err := skills.LeerCobertura(contenidoDelFichero(t, rutaDeLaCobertura))
	if err != nil {
		return err
	}

	lecturas := slices.Concat(lecturasDeLosTipos(tipos), lecturasDeLaCobertura(cobertura), lecturasDeLaTabla(tabla))
	textos, defectos := leerLosBloques(t, lecturas)

	for _, lectura := range lecturas {
		texto, leido := textos[claveDelBloque(lectura.norma, lectura.bloque)]
		if leido && lectura.duracion != nil {
			if err := comprobarLaDuracion(lectura, texto); err != nil {
				defectos = append(defectos, err)
			}
		}
	}

	defectos = append(defectos, verificarLasNormasDeLaCobertura(t, cobertura)...)

	for _, fila := range tabla {
		if texto, leido := textos[claveDelBloque(fila.norma, fila.bloque)]; leido {
			defectos = append(defectos, comprobarLasPalabras(fila, texto)...)
		}
	}

	return errors.Join(defectos...)
}

// lecturasDeLosTipos son los preceptos de cada tipo, en el orden del fichero:
// los de preceptos, con la duración del tipo en el que la fija, y los de sus
// unidades, en el orden de plazos.Unidades.
func lecturasDeLosTipos(tipos []plazos.Tipo) []lecturaDeUnPrecepto {
	var lecturas []lecturaDeUnPrecepto

	for _, tipo := range tipos {
		for _, precepto := range tipo.Preceptos {
			lectura := lecturaDeUnPrecepto{
				quien: tipo.Codigo, norma: precepto.Norma, bloque: precepto.Bloque, apartado: precepto.Apartado,
			}
			if precepto.Fija == plazos.FijaDuracion {
				lectura.duracion = tipo.Duracion
			}

			lecturas = append(lecturas, lectura)
		}

		for _, unidad := range plazos.Unidades() {
			if precepto, esta := tipo.Unidades[unidad]; esta {
				lecturas = append(lecturas, lecturaDeUnPrecepto{
					quien: tipo.Codigo, norma: precepto.Norma, bloque: precepto.Bloque, apartado: precepto.Apartado,
				})
			}
		}
	}

	return lecturas
}

// lecturasDeLaCobertura son, de cada entrada de la cobertura en su orden, el
// precepto de la LPAC que la regula aparte y, si la lleva, el precepto
// aplicable de la norma que la rige.
func lecturasDeLaCobertura(cobertura skills.Cobertura) []lecturaDeUnPrecepto {
	var lecturas []lecturaDeUnPrecepto

	for _, regulado := range cobertura.ReguladosAparte {
		lecturas = append(lecturas, lecturaDeUnPrecepto{
			quien: regulado.Procedimiento, norma: regulado.Precepto.Norma, bloque: regulado.Precepto.Bloque,
			apartado: regulado.Precepto.Apartado,
		})

		if norma := regulado.Norma; norma != nil {
			lecturas = append(lecturas, lecturaDeUnPrecepto{
				quien: regulado.Procedimiento, norma: norma.Identificador, bloque: norma.Aplicable.Bloque,
				apartado: norma.Aplicable.Apartado,
			})
		}
	}

	return lecturas
}

// lecturasDeLaTabla son los bloques de las filas de la tabla de las palabras
// clave, cada uno nombrado por su precepto.
func lecturasDeLaTabla(tabla []palabrasDeUnPrecepto) []lecturaDeUnPrecepto {
	lecturas := make([]lecturaDeUnPrecepto, 0, len(tabla))
	for _, fila := range tabla {
		lecturas = append(lecturas, lecturaDeUnPrecepto{quien: fila.precepto, norma: fila.norma, bloque: fila.bloque})
	}

	return lecturas
}

// consultaDelBloque es la invocación de boe articulo con la norma y el bloque.
func consultaDelBloque(norma, bloque string) Consulta {
	return Consulta{Applet: appletDeLasNormas, Verbo: verboArticulo, Argumentos: []string{norma, bloque}}
}

// claveDelBloque identifica el bloque de una norma entre los leídos: la orden
// que lo lee.
func claveDelBloque(norma, bloque string) string {
	return ordenDe(consultaDelBloque(norma, bloque))
}

// leerLosBloques prepara con Preparar una caché desde UnionDeGrabaciones con
// boe articulo de cada bloque de las lecturas, sin repetir y en su orden, y
// lee de ella el texto de cada uno que está grabado, por su claveDelBloque.
// Cada bloque que no lo está es un defecto por cada lectura que lo nombra, con
// la falta de Preparar, que dice la orden, su código y lo que no estaba grabado.
func leerLosBloques(t *testing.T, lecturas []lecturaDeUnPrecepto) (map[string]string, []error) {
	t.Helper()

	var consultas []Consulta

	for _, lectura := range lecturas {
		clave := claveDelBloque(lectura.norma, lectura.bloque)
		if !slices.ContainsFunc(consultas, func(consulta Consulta) bool { return ordenDe(consulta) == clave }) {
			consultas = append(consultas, consultaDelBloque(lectura.norma, lectura.bloque))
		}
	}

	dirCache := t.TempDir()

	faltas, err := Preparar(dirCache, UnionDeGrabaciones(), consultas)
	require.NoError(t, err, "la caché de los preceptos se prepara desde %s", strings.Join(UnionDeGrabaciones(), " y "))

	sinGrabar := make(map[string]Falta, len(faltas))
	for _, falta := range faltas {
		sinGrabar[ordenDe(falta.Consulta)] = falta
	}

	registro, err := registroDeBoe(t.TempDir(), cache.ConDirectorio(dirCache))
	require.NoError(t, err)

	textos := make(map[string]string, len(consultas))

	for _, consulta := range consultas {
		if _, falta := sinGrabar[ordenDe(consulta)]; !falta {
			textos[ordenDe(consulta)] = textoDelBloque(t, registro, consulta)
		}
	}

	var defectos []error

	for _, lectura := range lecturas {
		if falta, esta := sinGrabar[claveDelBloque(lectura.norma, lectura.bloque)]; esta {
			defectos = append(defectos, fmt.Errorf("%s: %s %s no está grabado: %s",
				lectura.quien, lectura.norma, lectura.bloque, falta))
		}
	}

	return textos, defectos
}

// textoDelBloque lee con app.Main, --json y --offline el bloque de la consulta
// de la caché del registro, que ya lo tiene, y devuelve su texto. Un código
// distinto de 0, una salida que no se lee o un texto vacío hacen fallar el test.
func textoDelBloque(t *testing.T, registro *app.Registro, consulta Consulta) string {
	t.Helper()

	var salida, errores bytes.Buffer

	argv := slices.Concat([]string{programaDeLasConsultas, consulta.Applet, consulta.Verbo}, consulta.Argumentos,
		[]string{"--json", "--offline"})
	codigo := app.Main(argv, registro, &salida, &errores,
		sinDatosDeConstruccion, sinDatosDeConstruccion, sinDatosDeConstruccion)
	require.Zero(t, codigo, "«%s» se lee de la caché preparada: %s", ordenDe(consulta), errores.String())

	var sobre struct {
		Data boe.Articulo `json:"data"`
	}
	require.NoError(t, json.Unmarshal(salida.Bytes(), &sobre), "la salida de «%s»", ordenDe(consulta))
	require.NotEmpty(t, sobre.Data.Texto, "«%s» da el texto del bloque", ordenDe(consulta))

	return sobre.Data.Texto
}

// duracionEnPalabras es la duración de un tipo como la escribe la norma: «un
// mes» con 1 mes y «dos meses» con 2 (contracts/datos-de-plazos.md §4.2 de H9,
// punto 2). Ninguna otra tiene forma fijada.
func duracionEnPalabras(duracion plazos.Duracion) (string, bool) {
	if duracion.Unidad != plazos.UnidadMeses {
		return "", false
	}

	switch duracion.Cantidad {
	case 1:
		return "un mes", true
	case 2:
		return "dos meses", true
	default:
		return "", false
	}
}

// parrafoDelApartado es el párrafo del texto de un bloque que es el de un
// apartado (punto 2): con un número tras el punto, como «124.1», la línea que
// empieza por ese número y un punto, «1.»; sin él, como «97», el bloque entero.
// Si ninguna línea empieza así, no hay párrafo.
func parrafoDelApartado(texto, apartado string) (string, bool) {
	_, numero, conNumero := strings.Cut(apartado, ".")
	if !conNumero {
		return texto, true
	}

	for linea := range strings.Lines(texto) {
		if strings.HasPrefix(linea, numero+".") {
			return linea, true
		}
	}

	return "", false
}

// comprobarLaDuracion exige que el párrafo del apartado de la lectura lleve en
// palabras la duración que fija. El defecto nombra el tipo, la norma, el
// bloque, el apartado y la duración en palabras que falta.
func comprobarLaDuracion(lectura lecturaDeUnPrecepto, texto string) error {
	precepto := fmt.Sprintf("%s: %s %s, apartado %s", lectura.quien, lectura.norma, lectura.bloque, lectura.apartado)

	palabras, conForma := duracionEnPalabras(*lectura.duracion)
	if !conForma {
		return fmt.Errorf("%s: la duración de %d %s no tiene forma en palabras fijada", precepto,
			lectura.duracion.Cantidad, lectura.duracion.Unidad)
	}

	parrafo, esta := parrafoDelApartado(texto, lectura.apartado)
	if !esta {
		return fmt.Errorf("%s: ninguna línea del texto grabado es la del apartado", precepto)
	}

	if !strings.Contains(parrafo, palabras) {
		return fmt.Errorf("%s: el párrafo del apartado no lleva «%s»", precepto, palabras)
	}

	return nil
}

// comprobarLasPalabras da un defecto por cada palabra de la fila que el texto
// grabado de su bloque no lleva, tal cual, con el precepto, la norma y el
// bloque.
func comprobarLasPalabras(fila palabrasDeUnPrecepto, texto string) []error {
	var defectos []error

	for _, palabra := range fila.palabras {
		if !strings.Contains(texto, palabra) {
			defectos = append(defectos, fmt.Errorf("%s (%s %s): el texto grabado no lleva «%s»",
				fila.precepto, fila.norma, fila.bloque, palabra))
		}
	}

	return defectos
}

// verificarLasNormasDeLaCobertura resuelve cada entrada del manifiesto de H9
// con su búsqueda grabada, como TestIdentificadoresDeLasNormas, y exige a cada
// norma de la cobertura exactamente una entrada que la resuelva (punto 3;
// FR-031): su identificador sale de lo grabado y no de memoria (research S5 de
// H9). Una entrada que no se resuelve es un defecto, y uno cada norma que
// ninguna o más de una resuelven, con su procedimiento.
func verificarLasNormasDeLaCobertura(t *testing.T, cobertura skills.Cobertura) []error {
	t.Helper()

	manifiesto, err := LeerManifiesto(contenidoDelFichero(t, rutaDelManifiestoJuntoA(GrabacionesDeH9)))
	require.NoError(t, err)

	registro := registroDeLaReproduccion(t, copiaDeLaUnionDeGrabaciones(t))
	resueltas := make(map[string]int, len(manifiesto.Normas))

	var defectos []error

	for indice, entrada := range manifiesto.Normas {
		resuelta, err := resolverEntrada(registro, indice, entrada)
		if err != nil {
			defectos = append(defectos, err)

			continue
		}

		resueltas[resuelta.resultado.Identificador]++
	}

	for _, regulado := range cobertura.ReguladosAparte {
		if regulado.Norma == nil {
			continue
		}

		switch entradas := resueltas[regulado.Norma.Identificador]; entradas {
		case 1:
		case 0:
			defectos = append(defectos, fmt.Errorf("%s: %s: ninguna entrada del manifiesto de H9 lo resuelve",
				regulado.Procedimiento, regulado.Norma.Identificador))
		default:
			defectos = append(defectos, fmt.Errorf("%s: %s: lo resuelven %d entradas del manifiesto de H9, "+
				"y tiene que ser una sola", regulado.Procedimiento, regulado.Norma.Identificador, entradas))
		}
	}

	return defectos
}
