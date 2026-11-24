package skills

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// La cobertura sintética de TestLeerCobertura: la clave de lo regulado aparte en
// la línea 1, una entrada con norma en las líneas 2 a 4 y una sin ella en las 5
// y 6; el motivo de lo no declarado en la 7, y la clave de lo cubierto en la 8,
// con una fila sin materia en las 9 y 10 y otra con ella, y con una comilla
// invertida en su alcance, en las 11 a 13. Lo regulado aparte va delante para
// que los casos que ya fijaba el test den los mismos mensajes, con sus mismas
// líneas. Solo la forma es la del esquema: los procedimientos, los
// identificadores, los bloques, los apartados, el nombre, el motivo, los
// alcances y la materia son inventados.
const (
	inicioDeLaCobertura = "regulados_aparte:\n"

	entradaConNorma = "  - procedimiento: procedimiento de prueba con norma propia\n" +
		"    precepto: {norma: BOE-A-2099-1, bloque: daprimera, apartado: \"DA 1ª.2.a\"}\n" +
		"    norma: {identificador: BOE-A-2099-2, nombre: \"Ley de Prueba, General\", " +
		"aplicable: {bloque: a9, apartado: \"9\"}}\n"
	entradaSinNorma = "  - procedimiento: procedimiento de prueba sin norma\n" +
		"    precepto: {norma: BOE-A-2099-1, bloque: a7, apartado: \"7.4\"}\n"

	reguladosDePrueba = inicioDeLaCobertura + entradaConNorma + entradaSinNorma

	motivoDePrueba = "motivo_de_lo_no_declarado: fuera de lo que la prueba cubre\n"

	cubiertoDePrueba = "cubierto:\n" +
		"  - capacidad: normativa\n" +
		"    alcance: leer las normas de prueba, de cualquier materia.\n" +
		"  - capacidad: plazos\n" +
		"    alcance: solo los plazos de prueba, que declara `la-skill-de-prueba` con su línea.\n" +
		"    materia: prueba\n"

	coberturaDePrueba = reguladosDePrueba + motivoDePrueba + cubiertoDePrueba
)

// coberturaLeidaDePrueba es lo que LeerCobertura lee de coberturaDePrueba.
func coberturaLeidaDePrueba() Cobertura {
	return Cobertura{
		MotivoDeLoNoDeclarado: "fuera de lo que la prueba cubre",
		Cubierto: []FilaCubierta{
			{Capacidad: "normativa", Alcance: "leer las normas de prueba, de cualquier materia."},
			{
				Capacidad: "plazos",
				Alcance:   "solo los plazos de prueba, que declara `la-skill-de-prueba` con su línea.",
				Materia:   "prueba",
			},
		},
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
// (contracts/datos-de-plazos.md §3 de H9; FR-031): el documento válido da su
// motivo, sus filas de lo cubierto y sus entradas de lo regulado aparte en su
// orden, con la materia y la norma que lo rige solo en la que las lleva; y cada
// forma que el esquema rechaza —una clave que no declara, un campo que falta, un
// texto vacío, un identificador o un bloque con otra forma, una lista vacía, un
// valor de otro tipo, y en lo cubierto (FR-002; contracts/datos-y-regiones.md §2
// de H33) una fila sin capacidad o sin alcance, una capacidad fuera del
// vocabulario, una fila con normas y un motivo que falta o va vacío— y la clave
// repetida que rechaza el lector son un error que nombra el punto del documento
// y su línea, y con él no se entrega nada. El subtest repositorio fija el
// motivo y lo cubierto de la base de data/cobertura.yaml (FR-001 de H33): sus
// filas sin materia o de lo administrativo, en su orden; las demás son de otro
// paquete (FR-016 de H45) y solo se exige que su materia sea el vertical de
// alguna norma de data/normas.yaml.
func TestLeerCobertura(t *testing.T) {
	t.Parallel()

	t.Run("repositorio", func(t *testing.T) {
		t.Parallel()

		contenido, err := os.ReadFile(coberturaDelRepositorio)
		require.NoError(t, err, "la cobertura por materia del repositorio no se puede leer")

		cobertura, err := LeerCobertura(contenido)
		require.NoError(t, err, "la cobertura por materia del repositorio, %s", coberturaDelRepositorio)

		verticales := verticalesDelRepositorio(t)

		var deLaBase []FilaCubierta

		for _, fila := range cobertura.Cubierto {
			if fila.Materia == "" || fila.Materia == paqueteDeLaBase {
				deLaBase = append(deLaBase, fila)

				continue
			}

			assert.Contains(t, verticales, fila.Materia,
				"la fila de lo cubierto de la capacidad %q de %s lleva por materia el vertical de una norma de %s",
				fila.Capacidad, coberturaDelRepositorio, normasDelRepositorio)
		}

		assert.Equal(t, "fuera de lo que kitlegal cubre", cobertura.MotivoDeLoNoDeclarado,
			"%s trae el motivo de lo no declarado de FR-001", coberturaDelRepositorio)
		assert.Equal(t, []FilaCubierta{
			{
				Capacidad: "normativa",
				Alcance: "leer y citar cualquier norma consolidada del BOE, estatal o autonómica, de cualquier " +
					"materia y en cualquier fecha.",
			},
			{
				Capacidad: "jurisprudencia",
				Alcance: "preparar la consulta exacta de una sentencia, que la persona busca con su navegador, y " +
					"cotejar el documento que trae. No consulta ningún buscador ni dice si una sentencia existe.",
			},
			{
				Capacidad: "plazos",
				Alcance: "solo los del procedimiento administrativo común, los del acceso a la información pública " +
					"y la interposición del recurso contencioso-administrativo, con los festivos de los municipios " +
					"que cuentan. Lo que la ley regula aparte de ese procedimiento lo declara `plazos-y-territorio` " +
					"con su línea.",
				Materia: "administrativo",
			},
			{
				Capacidad: "territorio",
				Alcance: "la provincia, la comunidad, el régimen, el DIR3 del ayuntamiento y los boletines de un " +
					"municipio.",
			},
		}, deLaBase, "%s trae las cuatro filas de lo cubierto de la base de FR-001, en su orden", coberturaDelRepositorio)
	})

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
			documento: inicioDeLaCobertura[:len(inicioDeLaCobertura)-1] + " []\n" + motivoDePrueba + cubiertoDePrueba,
			error:     "regulados_aparte, línea 1: minItems: got 0, want 1",
		},
		{
			nombre:    "valor-de-otro-tipo",
			documento: cambiadaEnLaCobertura(t, "apartado: \"7.4\"", "apartado: [\"7.4\"]"),
			error:     "regulados_aparte/1/precepto/apartado, línea 6: got array, want string",
		},
		{
			nombre: "clave-repetida",
			documento: reguladosDePrueba + inicioDeLaCobertura + entradaSinNorma +
				motivoDePrueba + cubiertoDePrueba,
			error:    "regulados_aparte repetido en las líneas 1 y 7",
			repetida: &ClaveRepetida{Clave: "regulados_aparte", Lineas: [2]int{1, 7}},
		},
		{
			nombre:    "fila-sin-capacidad",
			documento: cambiadaEnLaCobertura(t, "  - capacidad: normativa\n    alcance:", "  - alcance:"),
			error:     "cubierto/0, línea 9: missing property 'capacidad'",
		},
		{
			nombre:    "fila-sin-alcance",
			documento: cambiadaEnLaCobertura(t, "\n    alcance: leer las normas de prueba, de cualquier materia.", ""),
			error:     "cubierto/0, línea 9: missing property 'alcance'",
		},
		{
			nombre:    "capacidad-fuera-del-vocabulario",
			documento: cambiadaEnLaCobertura(t, "capacidad: normativa", "capacidad: documentos"),
			error: "cubierto/0/capacidad, línea 9: " +
				"value must be one of 'normativa', 'jurisprudencia', 'plazos', 'territorio'",
		},
		{
			nombre: "fila-con-normas",
			documento: cambiadaEnLaCobertura(t, "de cualquier materia.\n",
				"de cualquier materia.\n    normas: [BOE-A-2015-10565]\n"),
			error: "cubierto/0, línea 9: additional properties 'normas' not allowed",
		},
		{
			nombre:    "sin-motivo",
			documento: cambiadaEnLaCobertura(t, motivoDePrueba, ""),
			error:     "línea 1: missing property 'motivo_de_lo_no_declarado'",
		},
		{
			nombre: "motivo-vacio",
			documento: cambiadaEnLaCobertura(t, "motivo_de_lo_no_declarado: fuera de lo que la prueba cubre",
				"motivo_de_lo_no_declarado: \"\""),
			error: "motivo_de_lo_no_declarado, línea 7: minLength: got 0, want 1",
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

// normasDelRepositorio es data/normas.yaml, relativo al directorio de este
// paquete: la tabla de normas de la que salen los paquetes, los valores de
// vertical de sus normas (data-model §1 de H45).
const normasDelRepositorio = "../../data/normas.yaml"

// paqueteDeLaBase es el paquete de lo administrativo, el de la base: el
// vertical de sus normas en data/normas.yaml (contracts/paquetes.md §1 de H32).
const paqueteDeLaBase = "administrativo"

// normasDelRepositorioPorIdentificador son las normas de data/normas.yaml por
// su identificador.
func normasDelRepositorioPorIdentificador(t *testing.T) map[string]Norma {
	t.Helper()

	contenido, err := os.ReadFile(normasDelRepositorio)
	require.NoError(t, err, "la tabla de normas del repositorio no se puede leer")

	normas, err := LeerNormas(contenido)
	require.NoError(t, err, "la tabla de normas del repositorio, %s", normasDelRepositorio)

	porIdentificador := make(map[string]Norma, len(normas))
	for _, norma := range normas {
		porIdentificador[norma.Identificador] = norma
	}

	return porIdentificador
}

// verticalesDelRepositorio son los valores de vertical de las normas de
// data/normas.yaml, los paquetes del repositorio, cada uno una vez.
func verticalesDelRepositorio(t *testing.T) []string {
	t.Helper()

	var verticales []string

	for _, norma := range normasDelRepositorioPorIdentificador(t) {
		if norma.Vertical != "" && !slices.Contains(verticales, norma.Vertical) {
			verticales = append(verticales, norma.Vertical)
		}
	}

	return verticales
}

// TestCoberturaDelRepositorio comprueba la cobertura por materia del
// repositorio (contracts/datos-de-plazos.md §3 de H9; FR-031; research D17 de
// H9; FR-001 a FR-004 y research D1 de H45.1): LeerCobertura la lee sin ningún
// defecto y cada una de sus filas de lo regulado aparte que es de lo
// administrativo es, en su orden, una de las nueve tuplas escritas aquí —la
// norma de su precepto, su apartado y la norma que la rige, sin el texto de su
// procedimiento—: las cuatro letras de la DA 1ª.2 LPAC, con la c repartida en
// cuatro filas, y los arts. 112.2 y 112.4, y solo la aplicación de los tributos
// y los sancionadores de tráfico con la norma que los rige. Una fila leída es
// de lo administrativo por la regla escrita aquí: la norma que la rige está en
// data/normas.yaml con ese vertical o, si no lleva norma o la suya no tiene
// vertical, lo tiene la de su precepto. Las leídas son una subsecuencia de las
// nueve: una fila que pasa a otro paquete, porque gana la norma de otro, deja
// de estar entre las leídas sin dejar un hueco que adivinar entre las nueve, y
// una fila de la base que pierde o cambia la norma que la rige no está en su
// sitio entre ellas. Que no falte ninguna fila de la base ni cambie el texto de
// su procedimiento lo comprueba el guion h32-paquetes, contra las líneas de la
// región de la cobertura de cada skill; que cada bloque esté grabado y que cada
// identificador sea el de la búsqueda grabada, TestPreceptosDeLosPlazos, en
// internal/evals.
func TestCoberturaDelRepositorio(t *testing.T) {
	t.Parallel()

	contenido, err := os.ReadFile(coberturaDelRepositorio)
	require.NoError(t, err, "la cobertura por materia del repositorio no se puede leer")

	cobertura, err := LeerCobertura(contenido)
	require.NoError(t, err, "la cobertura por materia del repositorio, %s", coberturaDelRepositorio)

	normas := normasDelRepositorioPorIdentificador(t)

	type tupla struct {
		delPrecepto, apartado, queLaRige string
	}

	var deLaBase []tupla

	for _, regulado := range cobertura.ReguladosAparte {
		leida := tupla{delPrecepto: regulado.Precepto.Norma, apartado: regulado.Precepto.Apartado}
		if regulado.Norma != nil {
			leida.queLaRige = regulado.Norma.Identificador
		}

		vertical := normas[leida.queLaRige].Vertical
		if vertical == "" {
			vertical = normas[leida.delPrecepto].Vertical
		}

		if vertical == paqueteDeLaBase {
			deLaBase = append(deLaBase, leida)
		}
	}

	const lpac = "BOE-A-2015-10565"

	nueve := []tupla{
		{lpac, "DA 1ª.2.a", "BOE-A-2003-23186"},
		{lpac, "DA 1ª.2.b", ""},
		{lpac, "DA 1ª.2.c", ""},
		{lpac, "DA 1ª.2.c", ""},
		{lpac, "DA 1ª.2.c", "BOE-A-2015-11722"},
		{lpac, "DA 1ª.2.c", ""},
		{lpac, "DA 1ª.2.d", ""},
		{lpac, "112.2", ""},
		{lpac, "112.4", ""},
	}

	casadas := 0

	for _, una := range nueve {
		if casadas < len(deLaBase) && deLaBase[casadas] == una {
			casadas++
		}
	}

	assert.Equal(t, len(deLaBase), casadas,
		"%s: de sus filas de %s, %v, solo las %d primeras son, en su orden, de las nueve tuplas de "+
			"contracts/datos-de-plazos.md §3 de H9, %v",
		coberturaDelRepositorio, paqueteDeLaBase, deLaBase, casadas, nueve)
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
