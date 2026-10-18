package skills

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// La cobertura sintética de TestLeerCobertura: la clave de la lista en la línea
// 1, una entrada con norma en las líneas 2 a 4 y una sin ella en las 5 y 6.
// Solo la forma es la del esquema: los procedimientos, los identificadores, los
// bloques, los apartados y el nombre son inventados.
const (
	inicioDeLaCobertura = "regulados_aparte:\n"

	entradaConNorma = "  - procedimiento: procedimiento de prueba con norma propia\n" +
		"    precepto: {norma: BOE-A-2099-1, bloque: daprimera, apartado: \"DA 1ª.2.a\"}\n" +
		"    norma: {identificador: BOE-A-2099-2, nombre: \"Ley de Prueba, General\", " +
		"aplicable: {bloque: a9, apartado: \"9\"}}\n"
	entradaSinNorma = "  - procedimiento: procedimiento de prueba sin norma\n" +
		"    precepto: {norma: BOE-A-2099-1, bloque: a7, apartado: \"7.4\"}\n"

	coberturaDePrueba = inicioDeLaCobertura + entradaConNorma + entradaSinNorma
)

// coberturaLeidaDePrueba es lo que LeerCobertura lee de coberturaDePrueba.
func coberturaLeidaDePrueba() Cobertura {
	return Cobertura{
		ReguladosAparte: []ReguladoAparte{
			{
				Procedimiento: "procedimiento de prueba con norma propia",
				Precepto:      PreceptoDeLaCobertura{Norma: "BOE-A-2099-1", Bloque: "daprimera", Apartado: "DA 1ª.2.a"},
				Norma: &NormaQueLoRige{
					Identificador: "BOE-A-2099-2",
					Nombre:        "Ley de Prueba, General",
					Aplicable:     PreceptoAplicable{Bloque: "a9", Apartado: "9"},
				},
			},
			{
				Procedimiento: "procedimiento de prueba sin norma",
				Precepto:      PreceptoDeLaCobertura{Norma: "BOE-A-2099-1", Bloque: "a7", Apartado: "7.4"},
			},
		},
	}
}

// TestLeerCobertura fija LeerCobertura sobre la cobertura sintética
// (contracts/datos-de-plazos.md §3 de H9; FR-031): el documento válido da sus
// entradas en su orden, con la norma que lo rige solo en la que la lleva; y cada
// forma que el esquema rechaza —una clave que no declara, un campo que falta, un
// texto vacío, un identificador o un bloque con otra forma, una lista vacía, un
// valor de otro tipo— y la clave repetida que rechaza el lector son un error
// que nombra el punto del documento y su línea, y con él no se entrega nada.
func TestLeerCobertura(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		documento string
		error     string
		// repetida es la clave repetida que el error lleva dentro, si la hay.
		repetida *ClaveRepetida
	}{
		{nombre: "valido", documento: coberturaDePrueba},
		{
			nombre:    "clave-no-declarada",
			documento: cambiadaEnLaCobertura(t, "bloque: a7,", "bloque: a7, vigencia: siempre,"),
			error:     "regulados_aparte/1/precepto, línea 6: additional properties 'vigencia' not allowed",
		},
		{
			nombre:    "falta-un-campo",
			documento: cambiadaEnLaCobertura(t, ", aplicable: {bloque: a9, apartado: \"9\"}", ""),
			error:     "regulados_aparte/0/norma, línea 4: missing property 'aplicable'",
		},
		{
			nombre:    "texto-vacio",
			documento: cambiadaEnLaCobertura(t, "procedimiento: procedimiento de prueba sin norma", "procedimiento: \"\""),
			error:     "regulados_aparte/1/procedimiento, línea 5: minLength: got 0, want 1",
		},
		{
			nombre:    "identificador-con-otra-forma",
			documento: cambiadaEnLaCobertura(t, "identificador: BOE-A-2099-2", "identificador: BOE-A-99-2"),
			error: "regulados_aparte/0/norma/identificador, línea 4: " +
				"'BOE-A-99-2' does not match pattern '^BOE-A-[0-9]{4}-[0-9]{1,9}$'",
		},
		{
			nombre:    "bloque-con-otra-forma",
			documento: cambiadaEnLaCobertura(t, "bloque: a9,", "bloque: \"a 9\","),
			error: "regulados_aparte/0/norma/aplicable/bloque, línea 4: " +
				"'a 9' does not match pattern '^[A-Za-z0-9][A-Za-z0-9.-]{0,63}$'",
		},
		{
			nombre:    "lista-vacia",
			documento: inicioDeLaCobertura[:len(inicioDeLaCobertura)-1] + " []\n",
			error:     "regulados_aparte, línea 1: minItems: got 0, want 1",
		},
		{
			nombre:    "valor-de-otro-tipo",
			documento: cambiadaEnLaCobertura(t, "apartado: \"7.4\"", "apartado: [\"7.4\"]"),
			error:     "regulados_aparte/1/precepto/apartado, línea 6: got array, want string",
		},
		{
			nombre:    "clave-repetida",
			documento: coberturaDePrueba + inicioDeLaCobertura + entradaSinNorma,
			error:     "regulados_aparte repetido en las líneas 1 y 7",
			repetida:  &ClaveRepetida{Clave: "regulados_aparte", Lineas: [2]int{1, 7}},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			cobertura, err := LeerCobertura([]byte(caso.documento))
			if caso.error == "" {
				require.NoError(t, err)
				assert.Equal(t, coberturaLeidaDePrueba(), cobertura)

				return
			}

			require.EqualError(t, err, caso.error)
			assert.Zero(t, cobertura, "una cobertura con defectos no se entrega a medias")

			if caso.repetida != nil {
				var repetida *ClaveRepetida
				require.ErrorAs(t, err, &repetida)
				assert.Equal(t, caso.repetida, repetida)
			}
		})
	}
}

// cambiadaEnLaCobertura es la cobertura sintética con viejo, que tiene que
// estar en ella exactamente una vez, sustituido por nuevo: un caso que no
// cambiara nada sería el documento válido con otro nombre.
func cambiadaEnLaCobertura(t *testing.T, viejo, nuevo string) string {
	t.Helper()

	require.Equal(t, 1, strings.Count(coberturaDePrueba, viejo), "%q está una sola vez en la cobertura sintética", viejo)

	return strings.Replace(coberturaDePrueba, viejo, nuevo, 1)
}

// coberturaDelRepositorio es data/cobertura.yaml, relativo al directorio de
// este paquete, que es donde go test ejecuta sus tests.
const coberturaDelRepositorio = "../../data/cobertura.yaml"

// TestCoberturaDelRepositorio comprueba la cobertura por materia del
// repositorio (contracts/datos-de-plazos.md §3 de H9; FR-031; research D17 de
// H9): LeerCobertura la lee sin ningún defecto y trae las nueve entradas en su
// orden —las cuatro letras de la DA 1ª.2 LPAC, con la c repartida por materias,
// y los arts. 112.2 y 112.4—, y solo la aplicación de los tributos y los
// sancionadores de tráfico llevan la norma que los rige. Que cada bloque esté
// grabado y que cada identificador sea el de la búsqueda grabada lo comprueba
// TestPreceptosDeLosPlazos, en internal/evals.
func TestCoberturaDelRepositorio(t *testing.T) {
	t.Parallel()

	contenido, err := os.ReadFile(coberturaDelRepositorio)
	require.NoError(t, err, "la cobertura por materia del repositorio no se puede leer")

	cobertura, err := LeerCobertura(contenido)
	require.NoError(t, err, "la cobertura por materia del repositorio, %s", coberturaDelRepositorio)

	type entrada struct {
		procedimiento, apartado string
		conNorma                bool
	}

	leidas := make([]entrada, 0, len(cobertura.ReguladosAparte))
	for _, regulado := range cobertura.ReguladosAparte {
		leidas = append(leidas, entrada{regulado.Procedimiento, regulado.Precepto.Apartado, regulado.Norma != nil})
	}

	assert.Equal(t, []entrada{
		{"aplicación de los tributos en materia tributaria y aduanera y su revisión en vía administrativa", "DA 1ª.2.a", true},
		{"gestión, inspección, liquidación, recaudación, impugnación y revisión en materia de Seguridad Social y " +
			"desempleo", "DA 1ª.2.b", false},
		{"procedimientos sancionadores en materia tributaria y aduanera", "DA 1ª.2.c", false},
		{"procedimientos sancionadores en el orden social", "DA 1ª.2.c", false},
		{"procedimientos sancionadores en materia de tráfico y seguridad vial", "DA 1ª.2.c", true},
		{"procedimientos sancionadores en materia de extranjería", "DA 1ª.2.c", false},
		{"actuaciones y procedimientos en materia de extranjería y asilo", "DA 1ª.2.d", false},
		{"procedimientos de impugnación que una ley establece en sustitución del recurso de alzada", "112.2", false},
		{"reclamaciones económico-administrativas", "112.4", false},
	}, leidas, "%s trae las nueve entradas de contracts/datos-de-plazos.md §3, en su orden", coberturaDelRepositorio)
}

// TestCompilarEsquemaDeCoberturaDesdeUnaRuta fija la lectura y la compilación
// del esquema de data/cobertura.yaml desde su ruta, como la de los otros
// esquemas publicados de data/: la del esquema publicado lo da; una carpeta en
// su lugar es un error de lectura, y un JSON que no es un JSON Schema válido,
// uno de compilación que nombra la ruta.
func TestCompilarEsquemaDeCoberturaDesdeUnaRuta(t *testing.T) {
	t.Parallel()

	publicado, err := compilarEsquemaDeCobertura(rutaDelEsquemaDeCobertura)
	require.NoError(t, err)
	assert.NotNil(t, publicado)

	directorio := t.TempDir()
	carpeta := filepath.Join(directorio, "carpeta.json")
	sinCompilar := filepath.Join(directorio, "procedimiento-sin-tipo.json")

	require.NoError(t, os.Mkdir(carpeta, 0o750))
	require.NoError(t, os.WriteFile(sinCompilar, []byte(`{"$schema": "https://json-schema.org/draft/2020-12/schema", `+
		`"properties": {"procedimiento": {"type": "no"}}}`), 0o600))

	for ruta, defecto := range map[string]string{
		carpeta:     "no se puede leer el esquema de data/cobertura.yaml: ",
		sinCompilar: "el esquema de data/cobertura.yaml " + sinCompilar + ": el esquema no compila: ",
	} {
		esquema, err := compilarEsquemaDeCobertura(ruta)
		require.ErrorContains(t, err, defecto)
		assert.Nil(t, esquema)

		if ruta == carpeta {
			require.ErrorIs(t, err, syscall.EISDIR)
		}
	}
}
