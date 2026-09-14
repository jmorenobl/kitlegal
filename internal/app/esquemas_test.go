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
// antes de compararlos, los dos ficheros publicados desde lo que emite
// --describe. Solo la usa la tarea [datos] de los esquemas, con la orden del
// contrato esquemas-fixtures-y-controles §1, y lo escrito lo revisa una persona;
// make schema-check nunca la pasa (FR-110).
var actualizarEsquemas = flag.Bool("actualizar-esquemas", false,
	"escribe en schemas/ norma.json y bloque.json desde --describe de sus verbos antes de compararlos")

const (
	// carpetaDeLosEsquemas es schemas/, relativa al directorio de este paquete.
	carpetaDeLosEsquemas = "../../schemas"
	// raizDeLosEsquemas es el principio del $id de cada fichero publicado y del
	// de cada una de sus partes.
	raizDeLosEsquemas = "https://ventanillalegal.es/schemas/"
	// borradorDeLosEsquemas es el $schema de la raíz de cada fichero: el mismo
	// borrador que declara cada parte.
	borradorDeLosEsquemas = "https://json-schema.org/draft/2020-12/schema"
)

// ficheroDeEsquemas es uno de los ficheros publicados en schemas/: su nombre, la
// entidad que nombra su título y los verbos de boe cuyas salidas agrupa, en orden
// alfabético (contrato esquemas-fixtures-y-controles §1).
type ficheroDeEsquemas struct {
	nombre  string
	entidad string
	verbos  []string
}

// ficherosDeEsquemas son los dos ficheros publicados, con sus verbos (FR-110).
var ficherosDeEsquemas = []ficheroDeEsquemas{
	{nombre: "norma.json", entidad: "norma", verbos: []string{"analisis", "buscar", "indice", "metadatos"}},
	{nombre: "bloque.json", entidad: "bloque", verbos: []string{"articulo", "articulos"}},
}

// TestEsquemasPublicados es lo que vigila make schema-check (FR-110, SC-006;
// contrato esquemas-fixtures-y-controles §1; research.md D11): regenera en
// memoria norma.json y bloque.json desde lo que emite --describe cada verbo del
// registro de producción, cada parte en su forma canónica con su $id, y compara
// cada fichero publicado que exista, parte a parte y entero. Una parte distinta
// falla nombrando el fichero y el verbo; un fichero que no es la serialización
// canónica de sus partes —la raíz, una parte de más, el orden, el sangrado o el
// salto final— falla nombrando el fichero. Mientras no hay ninguno publicado no
// se compara nada. Con -actualizar-esquemas escribe antes los dos.
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
		}

		for _, fichero := range ficherosDeEsquemas {
			documento, err := objetoJSON(esquemaCanonico(t, fichero, emitidas[fichero.nombre]))
			require.NoError(t, err)

			assert.ElementsMatch(t, []string{"$defs", "$id", "$schema", "description", "title"},
				slices.Collect(maps.Keys(documento)), "la raíz de %s", fichero.nombre)
			assert.Equal(t, "https://ventanillalegal.es/schemas/"+fichero.nombre, documento["$id"])
			assert.Equal(t, "https://json-schema.org/draft/2020-12/schema", documento["$schema"])
			assert.Equal(t, descripciones[fichero.nombre], documento["description"])
			assert.Equal(t, "boe · "+fichero.entidad, documento["title"])

			partes, esObjeto := documento["$defs"].(map[string]any)
			require.True(t, esObjeto, "$defs de %s es un objeto", fichero.nombre)
			require.ElementsMatch(t, fichero.verbos, slices.Collect(maps.Keys(partes)))

			for _, verbo := range fichero.verbos {
				parte, esObjeto := partes[verbo].(map[string]any)
				require.True(t, esObjeto, "la parte de %s en %s es un objeto", verbo, fichero.nombre)

				assert.Equal(t, "https://ventanillalegal.es/schemas/"+fichero.nombre+"/"+verbo, parte["$id"])
				assert.Equal(t, "boe "+verbo, parte["title"], "la parte es lo que emite el verbo con --describe")
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

	t.Run("escribe-solo-para-su-propietario", func(t *testing.T) {
		t.Parallel()

		carpeta := filepath.Join(t.TempDir(), "schemas")
		bloque := ficherosDeEsquemas[1]

		escribeEsquema(t, carpeta, bloque, esquemaCanonico(t, bloque, emitidas[bloque.nombre]))

		escrito, err := os.Stat(filepath.Join(carpeta, bloque.nombre))
		require.NoError(t, err)
		assert.Equal(t, fs.FileMode(0o600), escrito.Mode().Perm())
		require.NoError(t, comprobarEsquemas(carpeta, ficherosDeEsquemas, emitidas))
	})

	compruebaElComparadorDeEsquemas(t, emitidas)
}

// compruebaElComparadorDeEsquemas publica en carpetas temporales otras formas de
// los dos ficheros regenerados con las partes reales y exige, para cada una, la
// lista exacta de fallos del comparador, en su orden: ninguno si coinciden, el
// de la parte y verbo distintos si difiere una parte, y el del fichero si difiere
// fuera de ellas.
func compruebaElComparadorDeEsquemas(t *testing.T, emitidas map[string]map[string]any) {
	t.Helper()

	norma, bloque := ficherosDeEsquemas[0], ficherosDeEsquemas[1]
	normaRegenerada := esquemaCanonico(t, norma, emitidas[norma.nombre])
	bloqueRegenerado := esquemaCanonico(t, bloque, emitidas[bloque.nombre])

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
			publicados: map[string][]byte{norma.nombre: normaRegenerada, bloque.nombre: bloqueRegenerado},
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
// esquemas-fixtures-y-controles §1): los dos ficheros publicados existen y cada
// verbo del registro de producción tiene su parte en exactamente uno de ellos.
// Sin esto, borrar un fichero o registrar un verbo sin publicar su parte dejaría
// su salida sin el contrato contra el que la valida TestSalidaDeBoeContraSchemas.
//
// El primer subtest lo comprueba sobre schemas/; el resto demuestra sobre
// carpetas temporales, con las partes reales, que la comprobación no pasa en
// vacío.
func TestEsquemasCubrenTodosLosVerbos(t *testing.T) {
	t.Parallel()

	registrados := verbosDeProduccion(t)
	require.NotEmpty(t, registrados, "sin verbos registrados no habría nada que cubrir")

	t.Run("schemas", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, comprobarCobertura(carpetaDeLosEsquemas, ficherosDeEsquemas, registrados))
	})

	emitidas := partesDeProduccion(t)
	norma, bloque := ficherosDeEsquemas[0], ficherosDeEsquemas[1]
	normaRegenerada := esquemaCanonico(t, norma, emitidas[norma.nombre])
	bloqueRegenerado := esquemaCanonico(t, bloque, emitidas[bloque.nombre])

	sinArticulos := maps.Clone(emitidas[bloque.nombre])
	delete(sinArticulos, "articulos")

	normaConArticulo := maps.Clone(emitidas[norma.nombre])
	normaConArticulo["articulo"] = emitidas[bloque.nombre]["articulo"]

	casos := []struct {
		nombre      string
		publicados  map[string][]byte
		registrados []string
		fallos      []string
	}{
		{
			nombre:      "regenerados-cubren-los-registrados",
			publicados:  map[string][]byte{norma.nombre: normaRegenerada, bloque.nombre: bloqueRegenerado},
			registrados: registrados,
		},
		{
			nombre:      "falta-un-fichero",
			publicados:  map[string][]byte{norma.nombre: normaRegenerada},
			registrados: registrados,
			fallos: []string{
				"schemas/bloque.json: el fichero no está publicado",
				"«boe articulo» no tiene su parte en ningún fichero de schemas/",
				"«boe articulos» no tiene su parte en ningún fichero de schemas/",
			},
		},
		{
			nombre: "verbo-sin-parte",
			publicados: map[string][]byte{
				norma.nombre:  normaRegenerada,
				bloque.nombre: esquemaCanonico(t, bloque, sinArticulos),
			},
			registrados: registrados,
			fallos:      []string{"«boe articulos» no tiene su parte en ningún fichero de schemas/"},
		},
		{
			nombre: "verbo-en-dos-ficheros",
			publicados: map[string][]byte{
				norma.nombre:  esquemaCanonico(t, norma, normaConArticulo),
				bloque.nombre: bloqueRegenerado,
			},
			registrados: registrados,
			fallos:      []string{"«boe articulo» tiene su parte en más de un fichero de schemas/: norma.json, bloque.json"},
		},
		{
			nombre:      "verbo-registrado-sin-publicar",
			publicados:  map[string][]byte{norma.nombre: normaRegenerada, bloque.nombre: bloqueRegenerado},
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

	registro, err := RegistroDeProduccion()
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
// raíz, y las de los dos ficheros son verbos del applet boe (contrato
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
			publicadaEn["boe "+verbo] = append(publicadaEn["boe "+verbo], fichero.nombre)
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

// partesDeProduccion regenera en memoria las partes de los dos ficheros: pide
// --describe de cada verbo al registro de producción, por la raíz de composición
// entera y con los argumentos del contrato, y añade a cada documento su $id. El
// resultado, por nombre de fichero y de verbo, solo se lee.
func partesDeProduccion(t *testing.T) map[string]map[string]any {
	t.Helper()

	registro, err := RegistroDeProduccion()
	require.NoError(t, err)

	emitidas := make(map[string]map[string]any, len(ficherosDeEsquemas))

	for _, fichero := range ficherosDeEsquemas {
		partes := make(map[string]any, len(fichero.verbos))

		for _, verbo := range fichero.verbos {
			invocacion := argvDeBoe(append(argumentoDelContrato(t, verbo), "--describe")...)

			res := invocar(t, registro, invocacion...)
			require.Equal(t, 0, res.codigo, "kitlegal boe %s --describe: %s", verbo, res.errores)

			parte, err := objetoJSON([]byte(res.salida))
			require.NoError(t, err, "kitlegal boe %s --describe emite un único objeto JSON", verbo)

			parte["$id"] = raizDeLosEsquemas + fichero.nombre + "/" + verbo
			partes[verbo] = parte
		}

		emitidas[fichero.nombre] = partes
	}

	return emitidas
}

// argumentoDelContrato es la invocación del verbo que usan las pruebas del
// applet, verbo incluido: --describe no ejecuta nada, pero la gramática exige los
// argumentos antes de describirse.
func argumentoDelContrato(t *testing.T, verbo string) []string {
	t.Helper()

	for _, contrato := range verbosDelContrato() {
		if contrato.nombre == verbo {
			return slices.Clone(contrato.argumento)
		}
	}

	require.Failf(t, "verbo sin invocación", "el verbo %q no está en el contrato puerto-y-applet §4", verbo)

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
		"description": "Salidas de los verbos " + enumeracion(fichero.verbos) + " del applet boe." +
			" Generado desde --describe con make schema-check; no editar.",
		"title": "boe · " + fichero.entidad,
	}
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
				"schemas/%s: la parte de «%s» no coincide con lo que emite `kitlegal boe %s --describe`",
				fichero.nombre, verbo, verbo))
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
