package evals

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Lo que estos tests leen del repositorio (contrato evals-y-grabaciones §3.1 y
// §3.4).
const (
	// nombreDelManifiestoDelRepositorio es el nombre del manifiesto de
	// grabación real, en el directorio que contiene GrabacionesDeH5.
	nombreDelManifiestoDelRepositorio = "grabaciones.json"

	// entradasMinimasDelManifiesto son las diez normas de H5: el manifiesto real
	// tiene al menos una entrada por cada una (plan.md, obligación 12).
	entradasMinimasDelManifiesto = 10

	// grabacionDeRobots es el único nombre de fichero que comparten los
	// conjuntos de grabaciones de H4 y de H5: la grabación de robots.txt, que la
	// reproducción no usa (data-model §7.3).
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

	t.Run("repositorio", func(t *testing.T) {
		t.Parallel()

		ruta := filepath.Join(filepath.Dir(GrabacionesDeH5), nombreDelManifiestoDelRepositorio)

		contenido, err := os.ReadFile(filepath.Clean(ruta))
		require.NoError(t, err, "el manifiesto de grabación del repositorio no se puede leer")

		leido, err := LeerManifiesto(contenido)
		require.NoError(t, err, "el manifiesto de grabación del repositorio, %s", ruta)
		assert.GreaterOrEqual(t, len(leido.Normas), entradasMinimasDelManifiesto,
			"el manifiesto de grabación del repositorio, %s, tiene una entrada por cada norma de H5", ruta)
	})
}

// TestGrabacionesSinSolape comprueba que los conjuntos de grabaciones de H4 y de
// H5 solo comparten el nombre de la grabación de robots.txt (contrato
// evals-y-grabaciones §3.4; data-model §7.3). Cualquier otro nombre común haría
// que, en la unión que copia Preparar, la grabación de H5 sustituyera sin avisar
// a la de H4, que es de H4 y que H5 solo lee.
func TestGrabacionesSinSolape(t *testing.T) {
	t.Parallel()

	deH4 := nombresDeLasGrabaciones(t, GrabacionesDeH4)
	deH5 := nombresDeLasGrabaciones(t, GrabacionesDeH5)

	var comunes []string

	for _, nombre := range deH5 {
		if slices.Contains(deH4, nombre) {
			comunes = append(comunes, nombre)
		}
	}

	assert.Equal(t, []string{grabacionDeRobots}, comunes,
		"los conjuntos de grabaciones %s y %s solo coinciden en %s", GrabacionesDeH4, GrabacionesDeH5, grabacionDeRobots)
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
