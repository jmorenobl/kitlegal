package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// actualizarEsquemas es la bandera con la que TestEsquemasPublicados escribe,
// antes de compararlos, los ficheros publicados de la tabla desde lo que emite
// --describe. Solo la usa la tarea [datos] de los esquemas, con la orden del
// contrato esquemas-fixtures-y-controles §1, y lo escrito lo revisa una persona;
// make schema-check nunca la pasa (FR-110).
var actualizarEsquemas = flag.Bool("actualizar-esquemas", false,
	"escribe en schemas/ los ficheros de la tabla desde --describe de sus verbos antes de compararlos")

const (
	// carpetaDeLosEsquemas es schemas/, relativa al directorio de este paquete.
	carpetaDeLosEsquemas = "../../schemas"
	// raizDeLosEsquemas es el principio del $id de cada fichero publicado y del
	// de cada una de sus partes.
	raizDeLosEsquemas = "https://kitlegal.es/schemas/"
	// borradorDeLosEsquemas es el $schema de la raíz de cada fichero: el mismo
	// borrador que declara cada parte.
	borradorDeLosEsquemas = "https://json-schema.org/draft/2020-12/schema"
)

// ficheroDeEsquemas es uno de los ficheros publicados en schemas/: el applet cuyos
// verbos publica, su nombre, la entidad que nombra su título y los verbos cuyas
// salidas agrupa, en orden alfabético (contrato esquemas-fixtures-y-controles §1).
// El applet compone el título, «<applet> · <entidad>», y el nombre de cada parte,
// «<applet> <verbo>», que es como se invoca el verbo (contrato del applet
// territorio §6).
type ficheroDeEsquemas struct {
	applet  string
	nombre  string
	entidad string
	verbos  []string
}

// ficherosDeEsquemas son los ficheros publicados, con su applet y sus verbos
// (FR-110). Los de territorio, skills, graph y mcp se llaman como su entidad,
// igual que los de boe, y no como el applet (contrato del applet territorio §6;
// research.md D25; research.md D16 de H19; contracts/applet-graph.md §6 de H7;
// research.md D11 de H21).
var ficherosDeEsquemas = []ficheroDeEsquemas{
	{applet: "boe", nombre: "norma.json", entidad: "norma", verbos: []string{"analisis", "buscar", "indice", "metadatos"}},
	{applet: "boe", nombre: "bloque.json", entidad: "bloque", verbos: []string{"articulo", "articulos"}},
	{applet: "territorio", nombre: "municipio.json", entidad: "municipio", verbos: []string{"resolver"}},
	{applet: "skills", nombre: "instalacion.json", entidad: "instalacion", verbos: []string{"doctor", "install", "list"}},
	{applet: "graph", nombre: "grafo.json", entidad: "grafo", verbos: []string{"check", "show", "stats"}},
	{applet: "mcp", nombre: "servidor.json", entidad: "servidor", verbos: []string{"serve"}},
}

// TestEsquemasPublicados es lo que vigila make schema-check (FR-110, SC-006;
// contrato esquemas-fixtures-y-controles §1; research.md D11): regenera en
// memoria los ficheros de la tabla desde lo que emite --describe cada verbo del
// registro de producción, cada parte en su forma canónica con su $id, y compara
// cada fichero publicado que exista, parte a parte y entero. Una parte distinta
// falla nombrando el fichero y el verbo; un fichero que no es la serialización
// canónica de sus partes —la raíz, una parte de más, el orden, el sangrado o el
// salto final— falla nombrando el fichero. Mientras no hay ninguno publicado no
// se compara nada. Con -actualizar-esquemas escribe antes todos.
//
// El primer subtest compara los ficheros publicados; el resto fija la forma
// canónica y la del contrato, y demuestra sobre carpetas temporales, con las
// partes reales, que el comparador no pasa en vacío.
func TestEsquemasPublicados(t *testing.T) {
	t.Parallel()

	emitidas := partesDeProduccion(t)

	if *actualizarEsquemas {
		for _, fichero := range ficherosDeEsquemas {
			escribeEsquema(t, carpetaDeLosEsquemas, fichero, esquemaCanonico(t, fichero, emitidas[fichero.nombre]))
		}
	}

	t.Run("schemas", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, comprobarEsquemas(carpetaDeLosEsquemas, ficherosDeEsquemas, emitidas))
	})

	t.Run("forma-canonica", func(t *testing.T) {
		t.Parallel()

		documento, err := objetoJSON([]byte(`{"zeta": "<b>«&»</b>", "alfa": {"entero": 18446744073709551615, "b": [1.50, true, null]}}`))
		require.NoError(t, err)

		contenido, err := formaCanonicaDeEsquema(documento)
		require.NoError(t, err)

		const esperada = `{
  "alfa": {
    "b": [
      1.50,
      true,
      null
    ],
    "entero": 18446744073709551615
  },
  "zeta": "<b>«&»</b>"
}
`

		// Línea a línea y no como JSON equivalente: lo que se fija es la forma
		// exacta —orden de claves, sangrado, cifras tal cual y salto final—.
		assert.Equal(t, strings.Split(esperada, "\n"), strings.Split(string(contenido), "\n"))
	})

	t.Run("forma-del-contrato", func(t *testing.T) {
		t.Parallel()

		descripciones := map[string]string{
			"norma.json": "Salidas de los verbos analisis, buscar, indice y metadatos del applet boe." +
				" Generado desde --describe con make schema-check; no editar.",
			"bloque.json": "Salidas de los verbos articulo y articulos del applet boe." +
				" Generado desde --describe con make schema-check; no editar.",
			"municipio.json": "Salida del verbo resolver del applet territorio." +
				" Generado desde --describe con make schema-check; no editar.",
			"instalacion.json": "Salidas de los verbos doctor, install y list del applet skills." +
				" Generado desde --describe con make schema-check; no editar.",
			"grafo.json": "Salidas de los verbos check, show y stats del applet graph." +
				" Generado desde --describe con make schema-check; no editar.",
			"servidor.json": "Salida del verbo serve del applet mcp." +
				" Generado desde --describe con make schema-check; no editar.",
		}

		for _, fichero := range ficherosDeEsquemas {
			documento, err := objetoJSON(esquemaCanonico(t, fichero, emitidas[fichero.nombre]))
			require.NoError(t, err)

			assert.ElementsMatch(t, []string{"$defs", "$id", "$schema", "description", "title"},
				slices.Collect(maps.Keys(documento)), "la raíz de %s", fichero.nombre)
			assert.Equal(t, "https://kitlegal.es/schemas/"+fichero.nombre, documento["$id"])
			assert.Equal(t, "https://json-schema.org/draft/2020-12/schema", documento["$schema"])
			assert.Equal(t, descripciones[fichero.nombre], documento["description"])
			assert.Equal(t, fichero.applet+" · "+fichero.entidad, documento["title"])

			partes, esObjeto := documento["$defs"].(map[string]any)
			require.True(t, esObjeto, "$defs de %s es un objeto", fichero.nombre)
			require.ElementsMatch(t, fichero.verbos, slices.Collect(maps.Keys(partes)))

			for _, verbo := range fichero.verbos {
				parte, esObjeto := partes[verbo].(map[string]any)
				require.True(t, esObjeto, "la parte de %s en %s es un objeto", verbo, fichero.nombre)

				assert.Equal(t, "https://kitlegal.es/schemas/"+fichero.nombre+"/"+verbo, parte["$id"])
				assert.Equal(t, fichero.applet+" "+verbo, parte["title"], "la parte es lo que emite el verbo con --describe")
				assert.Contains(t, parte, "properties")
			}
		}
	})

	t.Run("sin-ficheros-no-compara-nada", func(t *testing.T) {
		t.Parallel()

		// Sin partes emitidas: un comparador que las consultara sin ningún fichero
		// publicado fallaría por todas.
		require.NoError(t, comprobarEsquemas(t.TempDir(), ficherosDeEsquemas, nil))
	})

	t.Run("el-applet-sale-de-la-tabla", func(t *testing.T) {
		t.Parallel()

		// Una fila de otro applet con sus partes: el título, la descripción y la
		// invocación que nombra el fallo de una parte salen de la fila.
		otra := ficheroDeEsquemas{applet: "otro", nombre: "cosa.json", entidad: "cosa", verbos: []string{"deshacer", "hacer"}}
		partes := map[string]any{
			"deshacer": map[string]any{"$id": raizDeLosEsquemas + "cosa.json/deshacer", "title": "otro deshacer"},
			"hacer":    map[string]any{"$id": raizDeLosEsquemas + "cosa.json/hacer", "title": "otro hacer"},
		}

		documento := documentoDeEsquemas(otra, partes)
		assert.Equal(t, "otro · cosa", documento["title"])
		assert.Equal(t, "Salidas de los verbos deshacer y hacer del applet otro."+
			" Generado desde --describe con make schema-check; no editar.", documento["description"])

		// Con un solo verbo, la descripción lo nombra en singular.
		sola := ficheroDeEsquemas{applet: "otro", nombre: "sola.json", entidad: "sola", verbos: []string{"hacer"}}
		assert.Equal(t, "Salida del verbo hacer del applet otro."+
			" Generado desde --describe con make schema-check; no editar.",
			documentoDeEsquemas(sola, map[string]any{"hacer": partes["hacer"]})["description"])

		carpeta := t.TempDir()
		escribeEsquema(t, carpeta, otra, esquemaCanonico(t, otra, partes))

		// La salida de hacer cambiada sin regenerar.
		emitidasDeOtra := maps.Clone(partes)
		emitidasDeOtra["hacer"] = partes["deshacer"]

		assert.Equal(t,
			[]string{"schemas/cosa.json: la parte de «hacer» no coincide con lo que emite `kitlegal otro hacer --describe`"},
			fallosDe(comprobarEsquemas(carpeta, []ficheroDeEsquemas{otra}, map[string]map[string]any{otra.nombre: emitidasDeOtra})))
	})

	t.Run("escribe-solo-para-su-propietario", func(t *testing.T) {
		t.Parallel()

		carpeta := filepath.Join(t.TempDir(), "schemas")
		bloque := ficheroDeLaTabla(t, "bloque.json")

		escribeEsquema(t, carpeta, bloque, esquemaCanonico(t, bloque, emitidas[bloque.nombre]))

		escrito, err := os.Stat(filepath.Join(carpeta, bloque.nombre))
		require.NoError(t, err)
		assert.Equal(t, fs.FileMode(0o600), escrito.Mode().Perm())
		require.NoError(t, comprobarEsquemas(carpeta, ficherosDeEsquemas, emitidas))
	})

	compruebaElComparadorDeEsquemas(t, emitidas)
}

// compruebaElComparadorDeEsquemas publica en carpetas temporales otras formas de
// los ficheros regenerados con las partes reales y exige, para cada una, la lista
// exacta de fallos del comparador, en su orden: ninguno si coinciden, el de la
// parte y verbo distintos si difiere una parte, y el del fichero si difiere fuera
// de ellas.
func compruebaElComparadorDeEsquemas(t *testing.T, emitidas map[string]map[string]any) {
	t.Helper()

	norma, bloque := ficheroDeLaTabla(t, "norma.json"), ficheroDeLaTabla(t, "bloque.json")
	regenerados := esquemasRegenerados(t, emitidas)
	normaRegenerada, bloqueRegenerado := regenerados[norma.nombre], regenerados[bloque.nombre]

	const (
		parteDeArticulo  = "schemas/bloque.json: la parte de «articulo» no coincide con lo que emite `kitlegal boe articulo --describe`"
		parteDeArticulos = "schemas/bloque.json: la parte de «articulos» no coincide con lo que emite `kitlegal boe articulos --describe`"
		parteDeMetadatos = "schemas/norma.json: la parte de «metadatos» no coincide con lo que emite `kitlegal boe metadatos --describe`"
		bloqueNoCanonico = "schemas/bloque.json: el fichero no es la serialización canónica de sus partes"
	)

	// La salida de metadatos cambiada sin regenerar: el fichero publicado es
	// canónico, pero su parte de metadatos ya no es lo que el verbo emite.
	metadatosSinRegenerar := maps.Clone(emitidas[norma.nombre])
	metadatosSinRegenerar["metadatos"] = emitidas[norma.nombre]["indice"]

	sinArticulos := maps.Clone(emitidas[bloque.nombre])
	delete(sinArticulos, "articulos")

	conParteDeMas := maps.Clone(emitidas[bloque.nombre])
	conParteDeMas["ajena"] = emitidas[bloque.nombre]["articulo"]

	noEsJSON := []byte("no es JSON")
	_, errDeLectura := objetoJSON(noEsJSON)
	require.Error(t, errDeLectura)

	casos := []struct {
		nombre     string
		publicados map[string][]byte
		fallos     []string
	}{
		{
			nombre:     "regenerados-coinciden",
			publicados: regenerados,
		},
		{
			nombre: "parte-editada-a-mano",
			publicados: map[string][]byte{
				norma.nombre:  normaRegenerada,
				bloque.nombre: reemplazaUnaVez(t, bloqueRegenerado, "\"title\": \"boe articulo\"\n", "\"title\": \"boe articulo a mano\"\n"),
			},
			fallos: []string{parteDeArticulo},
		},
		{
			nombre:     "compara-solo-los-que-existen",
			publicados: map[string][]byte{bloque.nombre: reemplazaUnaVez(t, bloqueRegenerado, `"minLength": 1`, `"minLength": 2`)},
			fallos:     []string{parteDeArticulo},
		},
		{
			nombre:     "salida-cambiada-sin-regenerar",
			publicados: map[string][]byte{norma.nombre: esquemaCanonico(t, norma, metadatosSinRegenerar)},
			fallos:     []string{parteDeMetadatos},
		},
		{
			nombre:     "parte-que-falta",
			publicados: map[string][]byte{bloque.nombre: esquemaCanonico(t, bloque, sinArticulos)},
			fallos:     []string{parteDeArticulos},
		},
		{
			nombre:     "parte-de-mas",
			publicados: map[string][]byte{bloque.nombre: esquemaCanonico(t, bloque, conParteDeMas)},
			fallos:     []string{bloqueNoCanonico},
		},
		{
			nombre:     "raiz-editada-a-mano",
			publicados: map[string][]byte{bloque.nombre: reemplazaUnaVez(t, bloqueRegenerado, `"title": "boe · bloque"`, `"title": "boe · otro"`)},
			fallos:     []string{bloqueNoCanonico},
		},
		{
			nombre:     "otro-sangrado",
			publicados: map[string][]byte{bloque.nombre: bytes.ReplaceAll(bloqueRegenerado, []byte("  "), []byte("\t"))},
			fallos:     []string{bloqueNoCanonico},
		},
		{
			nombre:     "sin-salto-final",
			publicados: map[string][]byte{bloque.nombre: bytes.TrimSuffix(bloqueRegenerado, []byte("\n"))},
			fallos:     []string{bloqueNoCanonico},
		},
		{
			nombre:     "no-es-json",
			publicados: map[string][]byte{bloque.nombre: noEsJSON},
			fallos:     []string{parteDeArticulo, parteDeArticulos, bloqueNoCanonico + ": " + errDeLectura.Error()},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			carpeta := publicaEnUnaCarpeta(t, caso.publicados)

			assert.Equal(t, caso.fallos, fallosDe(comprobarEsquemas(carpeta, ficherosDeEsquemas, emitidas)))
		})
	}
}

// TestEsquemasCubrenTodosLosVerbos completa lo que vigila TestEsquemasPublicados,
// que no compara un fichero que no existe (FR-110, FR-111, SC-006; contrato
// esquemas-fixtures-y-controles §1): los ficheros publicados existen y cada verbo
// del registro de producción tiene su parte en exactamente uno de ellos. Sin
// esto, borrar un fichero o registrar un verbo sin publicar su parte dejaría su
// salida sin el contrato contra el que la valida TestSalidaDeBoeContraSchemas.
//
// El primer subtest lo comprueba sobre schemas/; el resto demuestra sobre
// carpetas temporales, con las partes reales, que la comprobación no pasa en
// vacío. Cada caso publica todos los ficheros de la tabla regenerados salvo lo
// que cambia, de modo que una fila nueva no altera los fallos que espera.
func TestEsquemasCubrenTodosLosVerbos(t *testing.T) {
	t.Parallel()

	registrados := verbosDeProduccion(t)
	require.NotEmpty(t, registrados, "sin verbos registrados no habría nada que cubrir")

	t.Run("schemas", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, comprobarCobertura(carpetaDeLosEsquemas, ficherosDeEsquemas, registrados))
	})

	emitidas := partesDeProduccion(t)
	norma, bloque := ficheroDeLaTabla(t, "norma.json"), ficheroDeLaTabla(t, "bloque.json")
	regenerados := esquemasRegenerados(t, emitidas)

	sinArticulos := maps.Clone(emitidas[bloque.nombre])
	delete(sinArticulos, "articulos")

	normaConArticulo := maps.Clone(emitidas[norma.nombre])
	normaConArticulo["articulo"] = emitidas[bloque.nombre]["articulo"]

	t.Run("la-parte-lleva-el-applet-de-su-fichero", func(t *testing.T) {
		t.Parallel()

		// Otro applet con un verbo que boe también tiene: su parte cubre «otro
		// articulo» y no es una segunda parte de «boe articulo».
		otra := ficheroDeEsquemas{applet: "otro", nombre: "otra.json", entidad: "otra", verbos: []string{"articulo"}}

		carpeta := publicaEnUnaCarpeta(t, regenerados)
		escribeEsquema(t, carpeta, otra,
			esquemaCanonico(t, otra, map[string]any{"articulo": emitidas[bloque.nombre]["articulo"]}))

		require.NoError(t, comprobarCobertura(carpeta,
			append(slices.Clone(ficherosDeEsquemas), otra), append(slices.Clone(registrados), "otro articulo")))
	})

	casos := []struct {
		nombre      string
		publicados  map[string][]byte
		registrados []string
		fallos      []string
	}{
		{
			nombre:      "regenerados-cubren-los-registrados",
			publicados:  regenerados,
			registrados: registrados,
		},
		{
			nombre:      "falta-un-fichero",
			publicados:  sinFichero(regenerados, bloque.nombre),
			registrados: registrados,
			fallos: []string{
				"schemas/bloque.json: el fichero no está publicado",
				"«boe articulo» no tiene su parte en ningún fichero de schemas/",
				"«boe articulos» no tiene su parte en ningún fichero de schemas/",
			},
		},
		{
			nombre:      "verbo-sin-parte",
			publicados:  conFichero(regenerados, bloque.nombre, esquemaCanonico(t, bloque, sinArticulos)),
			registrados: registrados,
			fallos:      []string{"«boe articulos» no tiene su parte en ningún fichero de schemas/"},
		},
		{
			nombre:      "verbo-en-dos-ficheros",
			publicados:  conFichero(regenerados, norma.nombre, esquemaCanonico(t, norma, normaConArticulo)),
			registrados: registrados,
			fallos:      []string{"«boe articulo» tiene su parte en más de un fichero de schemas/: norma.json, bloque.json"},
		},
		{
			nombre:      "verbo-registrado-sin-publicar",
			publicados:  regenerados,
			registrados: append(slices.Clone(registrados), "boe nuevo"),
			fallos:      []string{"«boe nuevo» no tiene su parte en ningún fichero de schemas/"},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			carpeta := publicaEnUnaCarpeta(t, caso.publicados)

			assert.Equal(t, caso.fallos, fallosDe(comprobarCobertura(carpeta, ficherosDeEsquemas, caso.registrados)))
		})
	}
}

// verbosDeProduccion son los verbos del registro de producción como se nombran
// detrás del binario, «<applet> <verbo>», en el orden de los applets y de sus
// verbos.
func verbosDeProduccion(t *testing.T) []string {
	t.Helper()

	registro, err := RegistroDeProduccion("")
	require.NoError(t, err)

	var verbos []string

	for _, nombre := range registro.Nombres() {
		applet, registrado := registro.Buscar(nombre)
		require.True(t, registrado, "%q está en el registro", nombre)

		for _, verbo := range applet.Verbos() {
			verbos = append(verbos, nombre+" "+verbo.Nombre)
		}
	}

	return verbos
}

// comprobarCobertura exige que cada fichero de la lista exista en la carpeta y
// que cada verbo registrado, «<applet> <verbo>», tenga su parte en exactamente
// uno de ellos, y reúne todos los fallos. Una parte es una clave del $defs de la
// raíz y nombra un verbo del applet de su fichero (contrato
// esquemas-fixtures-y-controles §1).
func comprobarCobertura(carpeta string, ficheros []ficheroDeEsquemas, registrados []string) error {
	var fallos []error

	publicadaEn := make(map[string][]string)

	for _, fichero := range ficheros {
		partes, err := partesPublicadas(filepath.Join(carpeta, fichero.nombre))
		if err != nil {
			fallos = append(fallos, fmt.Errorf("schemas/%s: %w", fichero.nombre, err))

			continue
		}

		for verbo := range partes {
			invocado := fichero.applet + " " + verbo
			publicadaEn[invocado] = append(publicadaEn[invocado], fichero.nombre)
		}
	}

	for _, verbo := range registrados {
		switch donde := publicadaEn[verbo]; {
		case len(donde) == 0:
			fallos = append(fallos, fmt.Errorf("«%s» no tiene su parte en ningún fichero de schemas/", verbo))
		case len(donde) > 1:
			fallos = append(fallos, fmt.Errorf("«%s» tiene su parte en más de un fichero de schemas/: %s",
				verbo, strings.Join(donde, ", ")))
		}
	}

	return errors.Join(fallos...)
}

// partesPublicadas son las partes del fichero publicado en la ruta, por verbo.
// Un fichero que no existe, que no es un objeto JSON o cuyo $defs no es un
// objeto no publica ninguna, y es un fallo.
func partesPublicadas(ruta string) (map[string]any, error) {
	contenido, err := leerEsquema(ruta)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, errors.New("el fichero no está publicado")
	}

	if err != nil {
		return nil, err
	}

	documento, err := objetoJSON(contenido)
	if err != nil {
		return nil, err
	}

	partes, esObjeto := documento["$defs"].(map[string]any)
	if !esObjeto {
		return nil, errors.New("el $defs de la raíz no es un objeto")
	}

	return partes, nil
}

// publicaEnUnaCarpeta escribe en una carpeta temporal los ficheros publicados,
// por nombre, y la devuelve: los que no están en el mapa no existen en ella.
func publicaEnUnaCarpeta(t *testing.T, publicados map[string][]byte) string {
	t.Helper()

	carpeta := t.TempDir()

	for _, fichero := range ficherosDeEsquemas {
		if contenido, publicado := publicados[fichero.nombre]; publicado {
			escribeEsquema(t, carpeta, fichero, contenido)
		}
	}

	return carpeta
}

// ficheroDeLaTabla es la fila de la tabla con ese nombre, que tiene que estar:
// las pruebas nombran los ficheros que usan en lugar de depender del orden de
// las filas.
func ficheroDeLaTabla(t *testing.T, nombre string) ficheroDeEsquemas {
	t.Helper()

	i := slices.IndexFunc(ficherosDeEsquemas, func(fichero ficheroDeEsquemas) bool {
		return fichero.nombre == nombre
	})
	require.NotEqual(t, -1, i, "%s está en la tabla de ficheros publicados", nombre)

	return ficherosDeEsquemas[i]
}

// esquemasRegenerados son todos los ficheros de la tabla en su forma canónica con
// las partes emitidas, por nombre: lo que escribiría -actualizar-esquemas.
func esquemasRegenerados(t *testing.T, emitidas map[string]map[string]any) map[string][]byte {
	t.Helper()

	regenerados := make(map[string][]byte, len(ficherosDeEsquemas))

	for _, fichero := range ficherosDeEsquemas {
		regenerados[fichero.nombre] = esquemaCanonico(t, fichero, emitidas[fichero.nombre])
	}

	return regenerados
}

// conFichero son los ficheros publicados con el de ese nombre cambiado por el
// contenido, sin tocar los demás ni el mapa de partida.
func conFichero(publicados map[string][]byte, nombre string, contenido []byte) map[string][]byte {
	variante := maps.Clone(publicados)
	variante[nombre] = contenido

	return variante
}

// sinFichero son los ficheros publicados sin el de ese nombre, sin tocar el mapa
// de partida.
func sinFichero(publicados map[string][]byte, nombre string) map[string][]byte {
	variante := maps.Clone(publicados)
	delete(variante, nombre)

	return variante
}

// partesDeProduccion regenera en memoria las partes de los ficheros de la tabla:
// pide --describe de cada verbo, con el applet de su fichero, al registro de
// producción, por la raíz de composición entera y con los argumentos del
// contrato, y añade a cada documento su $id. El resultado, por nombre de fichero
// y de verbo, solo se lee.
func partesDeProduccion(t *testing.T) map[string]map[string]any {
	t.Helper()

	registro, err := RegistroDeProduccion("")
	require.NoError(t, err)

	emitidas := make(map[string]map[string]any, len(ficherosDeEsquemas))

	for _, fichero := range ficherosDeEsquemas {
		partes := make(map[string]any, len(fichero.verbos))

		for _, verbo := range fichero.verbos {
			invocacion := slices.Concat([]string{"kitlegal", fichero.applet},
				argumentoDelContrato(t, fichero.applet, verbo), []string{"--describe"})

			res := invocar(t, registro, invocacion...)
			require.Equal(t, 0, res.codigo, "kitlegal %s %s --describe: %s", fichero.applet, verbo, res.errores)

			parte, err := objetoJSON([]byte(res.salida))
			require.NoError(t, err, "kitlegal %s %s --describe emite un único objeto JSON", fichero.applet, verbo)

			parte["$id"] = raizDeLosEsquemas + fichero.nombre + "/" + verbo
			partes[verbo] = parte
		}

		emitidas[fichero.nombre] = partes
	}

	return emitidas
}

// contratosDeLosApplets son, por applet, los verbos de su contrato con la
// invocación que usan sus pruebas: de ahí salen los argumentos con los que cada
// verbo de la tabla se describe.
var contratosDeLosApplets = map[string]func() []verboDelContrato{
	"boe":        verbosDelContrato,
	"graph":      verbosDelContratoDeGrafo,
	"mcp":        verbosDelContratoDeMCP,
	"skills":     verbosDelContratoDeSkills,
	"territorio": verbosDelContratoDeTerritorio,
}

// verbosDelContratoDeMCP es el único verbo de mcp con la invocación con que se
// describe: no tiene argumentos propios, así que basta el verbo
// (contracts/servidor-mcp.md §1 de H21). Describir no sirve ni lee de la
// entrada estándar.
func verbosDelContratoDeMCP() []verboDelContrato {
	return []verboDelContrato{{nombre: "serve", argumento: []string{"serve"}}}
}

// verbosDelContratoDeGrafo son los tres verbos de graph con la invocación con
// que se describen: show exige su id por su posición y el análisis de la
// invocación va antes que la descripción, así que sin él --describe termina en 2;
// stats no tiene argumentos y los de check son opcionales (contracts/applet-graph.md
// §1 de H7 y de H7.1). Describir no
// valida el id ni abre world.db, de modo que el de show es el mismo con que lo
// describe el guion grafo-applet de la suite de aceptación.
func verbosDelContratoDeGrafo() []verboDelContrato {
	return []verboDelContrato{
		{nombre: "show", argumento: []string{"show", "x"}},
		{nombre: "stats", argumento: []string{"stats"}},
		{nombre: "check", argumento: []string{"check"}},
	}
}

// verbosDelContratoDeSkills son los tres verbos de skills con la invocación con
// que se describen: ninguno tiene argumentos obligatorios, así que basta el verbo
// (contracts/applet-skills.md §1 de H19). Describir no examina el disco ni lee lo
// empotrado.
func verbosDelContratoDeSkills() []verboDelContrato {
	return []verboDelContrato{
		{nombre: verboInstall, argumento: []string{verboInstall}},
		{nombre: verboList, argumento: []string{verboList}},
		{nombre: verboDoctor, argumento: []string{verboDoctor}},
	}
}

// verbosDelContratoDeTerritorio es el único verbo de territorio con la
// invocación con que se describe: su consulta es obligatoria y el análisis de la
// invocación va antes que la descripción, así que sin ella --describe termina en
// 2 (contrato del applet territorio §1). Describir no resuelve nada, de modo que
// la consulta es la del municipio cubierto del registro local de sus pruebas, que
// no es ningún municipio real.
func verbosDelContratoDeTerritorio() []verboDelContrato {
	return []verboDelContrato{{nombre: "resolver", argumento: []string{"resolver", "Villaconfigurada"}}}
}

// argumentoDelContrato es la invocación del verbo del applet que usan sus
// pruebas, verbo incluido: --describe no ejecuta nada, pero la gramática exige los
// argumentos antes de describirse.
func argumentoDelContrato(t *testing.T, applet, verbo string) []string {
	t.Helper()

	contrato, conContrato := contratosDeLosApplets[applet]
	require.True(t, conContrato, "el applet %q no tiene contrato con las invocaciones de sus verbos", applet)

	for _, delContrato := range contrato() {
		if delContrato.nombre == verbo {
			return slices.Clone(delContrato.argumento)
		}
	}

	require.Failf(t, "verbo sin invocación", "el verbo %q no está en el contrato del applet %q", verbo, applet)

	return nil
}

// esquemaCanonico es el fichero con esas partes en su forma canónica.
func esquemaCanonico(t *testing.T, fichero ficheroDeEsquemas, partes map[string]any) []byte {
	t.Helper()

	contenido, err := formaCanonicaDeEsquema(documentoDeEsquemas(fichero, partes))
	require.NoError(t, err)

	return contenido
}

// documentoDeEsquemas es el fichero publicado con esas partes, en la forma del
// contrato esquemas-fixtures-y-controles §1: un recurso embebido por verbo bajo
// $defs, y en la raíz su $id, el borrador, la descripción y el título.
func documentoDeEsquemas(fichero ficheroDeEsquemas, partes map[string]any) map[string]any {
	return map[string]any{
		"$defs":   partes,
		"$id":     raizDeLosEsquemas + fichero.nombre,
		"$schema": borradorDeLosEsquemas,
		"description": salidasDe(fichero.verbos) + " del applet " + fichero.applet + "." +
			" Generado desde --describe con make schema-check; no editar.",
		"title": fichero.applet + " · " + fichero.entidad,
	}
}

// salidasDe nombra lo que publica un fichero como lo dice su descripción: la
// salida de su verbo, en singular, si publica uno, y si publica varios, las de
// todos enumerados.
func salidasDe(verbos []string) string {
	if len(verbos) == 1 {
		return "Salida del verbo " + verbos[0]
	}

	return "Salidas de los verbos " + enumeracion(verbos)
}

// enumeracion escribe los verbos como se enumeran en una frase: separados por
// comas y el último detrás de «y».
func enumeracion(verbos []string) string {
	if len(verbos) < 2 {
		return strings.Join(verbos, "")
	}

	return strings.Join(verbos[:len(verbos)-1], ", ") + " y " + verbos[len(verbos)-1]
}

// formaCanonicaDeEsquema escribe un documento como se publica: claves ordenadas
// —las de un mapa las ordena el codificador—, dos espacios de sangrado, sin
// escapar HTML y con salto final (contrato esquemas-fixtures-y-controles §1).
func formaCanonicaDeEsquema(documento any) ([]byte, error) {
	var contenido bytes.Buffer

	codificador := json.NewEncoder(&contenido)
	codificador.SetEscapeHTML(false)
	codificador.SetIndent("", "  ")

	if err := codificador.Encode(documento); err != nil {
		return nil, fmt.Errorf("el esquema no se puede serializar: %w", err)
	}

	return contenido.Bytes(), nil
}

// objetoJSON lee un único objeto JSON con las cifras conservadas como literales,
// que es la representación de la que sale la forma canónica: leer un esquema y
// volver a escribirlo no cambia ni una cifra.
func objetoJSON(contenido []byte) (map[string]any, error) {
	decodificador := json.NewDecoder(bytes.NewReader(contenido))
	decodificador.UseNumber()

	var objeto map[string]any

	if err := decodificador.Decode(&objeto); err != nil {
		return nil, fmt.Errorf("no es un objeto JSON: %w", err)
	}

	if objeto == nil {
		return nil, errors.New("no es un objeto JSON: es null")
	}

	if err := decodificador.Decode(new(json.RawMessage)); !errors.Is(err, io.EOF) {
		return nil, errors.New("lleva algo más detrás del objeto JSON")
	}

	return objeto, nil
}

// comprobarEsquemas compara cada fichero de la lista que exista en la carpeta con
// las partes emitidas de sus verbos y reúne todos los fallos. Un fichero que no
// existe no se compara, así que sin ninguno no se compara nada.
func comprobarEsquemas(carpeta string, ficheros []ficheroDeEsquemas, emitidas map[string]map[string]any) error {
	var fallos []error

	for _, fichero := range ficheros {
		publicado, err := leerEsquema(filepath.Join(carpeta, fichero.nombre))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}

		if err != nil {
			fallos = append(fallos, err)

			continue
		}

		fallos = append(fallos, compararEsquema(fichero, publicado, emitidas[fichero.nombre])...)
	}

	return errors.Join(fallos...)
}

// compararEsquema compara un fichero publicado con las partes emitidas de sus
// verbos: primero cada parte, que falla nombrando fichero y verbo, y después el
// fichero entero contra la serialización canónica de las partes que publica, que
// es lo que ve una diferencia fuera de ellas —la raíz, una parte de otro verbo, el
// orden de las claves, el sangrado o el salto final— y falla nombrando el fichero.
// Un contenido que no es un objeto JSON no publica ninguna parte.
func compararEsquema(fichero ficheroDeEsquemas, publicado []byte, emitidas map[string]any) []error {
	noCanonico := fmt.Errorf("schemas/%s: el fichero no es la serialización canónica de sus partes", fichero.nombre)

	documento, err := objetoJSON(publicado)
	if err != nil {
		return append(partesDistintas(fichero, nil, emitidas), fmt.Errorf("%w: %w", noCanonico, err))
	}

	publicadas := make(map[string]any, len(fichero.verbos))

	if definiciones, esObjeto := documento["$defs"].(map[string]any); esObjeto {
		for _, verbo := range fichero.verbos {
			if parte, publica := definiciones[verbo]; publica {
				publicadas[verbo] = parte
			}
		}
	}

	fallos := partesDistintas(fichero, publicadas, emitidas)

	canonico, err := formaCanonicaDeEsquema(documentoDeEsquemas(fichero, publicadas))
	if err != nil {
		return append(fallos, fmt.Errorf("%w: %w", noCanonico, err))
	}

	if !bytes.Equal(canonico, publicado) {
		fallos = append(fallos, noCanonico)
	}

	return fallos
}

// partesDistintas es un fallo por cada verbo del fichero cuya parte publicada no
// es la emitida, o que falta en uno de los dos lados. Las dos salen de leer JSON
// con las cifras como literales, así que ser iguales como valores es tener la
// misma forma canónica.
func partesDistintas(fichero ficheroDeEsquemas, publicadas, emitidas map[string]any) []error {
	var fallos []error

	for _, verbo := range fichero.verbos {
		publicada, publica := publicadas[verbo]
		emitida, emite := emitidas[verbo]

		if !publica || !emite || !reflect.DeepEqual(publicada, emitida) {
			fallos = append(fallos, fmt.Errorf(
				"schemas/%s: la parte de «%s» no coincide con lo que emite `kitlegal %s %s --describe`",
				fichero.nombre, verbo, fichero.applet, verbo))
		}
	}

	return fallos
}

// fallosDe son los mensajes de los fallos reunidos en el error, uno por línea, o
// ninguno.
func fallosDe(err error) []string {
	if err == nil {
		return nil
	}

	return strings.Split(err.Error(), "\n")
}

// reemplazaUnaVez cambia la primera aparición de viejo en el contenido, que tiene
// que estar: una edición que no cambia nada no probaría nada.
func reemplazaUnaVez(t *testing.T, contenido []byte, viejo, nuevo string) []byte {
	t.Helper()

	require.Contains(t, string(contenido), viejo)

	return bytes.Replace(contenido, []byte(viejo), []byte(nuevo), 1)
}

// leerEsquema lee un fichero publicado por filepath.Clean (gosec G304), desde un
// auxiliar distinto del que escribe (G703). El error conserva fs.ErrNotExist.
func leerEsquema(ruta string) ([]byte, error) {
	contenido, err := os.ReadFile(filepath.Clean(ruta))
	if err != nil {
		return nil, fmt.Errorf("%s no se puede leer: %w", ruta, err)
	}

	return contenido, nil
}

// escribeEsquema deja el fichero en la carpeta, y la carpeta, solo para su
// propietario (gosec G301 y G306), desde un auxiliar distinto del que lee (G703).
func escribeEsquema(t *testing.T, carpeta string, fichero ficheroDeEsquemas, contenido []byte) {
	t.Helper()

	require.NoError(t, os.MkdirAll(carpeta, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(carpeta, fichero.nombre), contenido, 0o600))
}
