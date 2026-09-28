package boe

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Lo que comparten las pruebas de lo que observan articulo y articulos
// (contracts/emision.md §1). Los ids, los tipos, las relaciones y las claves de
// los datos van escritos enteros, y no con las constantes de internal/core/grafo
// ni con las funciones de grafo.go, para que un cambio en ellos no pase por aquí
// en silencio.
const (
	// eliDeLaLey39 y eliDeLaLey30 son los ids de la Norma de la Ley 39/2015 y de
	// la Ley 30/1992: la ruta de su url_eli desde el segmento eli.
	eliDeLaLey39 = "eli/es/l/2015/10/01/39"
	eliDeLaLey30 = "eli/es/l/1992/11/26/30"
	// huellaDelArticulo21 es el hash_texto del artículo 21 de la Ley 39/2015 en
	// sus grabaciones, el mismo que fijan los guiones de aceptación del hito.
	huellaDelArticulo21 = "sha256:98d3b9d4686a3155f48641841df7023e2c17a693fcd0e11d6beef2615abcec7c"
	// lineaMinimaDelCuerpo es la longitud, en caracteres, desde la que una línea
	// del cuerpo de un texto no puede aparecer en ningún nodo (FR-041, FR-070).
	lineaMinimaDelCuerpo = 20
)

// TestObservadoDeBoe fija lo que observan del mundo los verbos de boe
// (contracts/emision.md §1; research.md D20; FR-040, FR-041, FR-065, FR-071):
//
//   - articulo emite, con la vigencia de su entrada —siete días—, la Norma por
//     el ELI de su url_eli, el Bloque y la BloqueVersion con sus datos, las
//     aristas eli:has_part y eli:has_version y el texto por su hash_texto, lo
//     pida a la fuente o lo sirva de la caché;
//   - articulos emite eso mismo una vez por cada bloque distinto, en el orden de
//     su primera aparición;
//   - el id de la Norma es la ruta de url_eli desde su primer segmento que es
//     exactamente eli hasta el final, sin consulta ni fragmento; sin ese
//     segmento, con url_eli vacío o que no se puede analizar, el artículo no
//     emite nada, ni con ningún id alternativo, y su data no cambia;
//   - ningún nodo lleva el cuerpo del bloque, que solo viaja en su Texto;
//   - y buscar, indice, metadatos y analisis no emiten nada.
func TestObservadoDeBoe(t *testing.T) {
	t.Parallel()

	subtests := []struct {
		nombre string
		probar func(*testing.T)
	}{
		{nombre: "articulo", probar: probarObservadoDeArticulo},
		{nombre: "articulos-con-un-bloque-repetido", probar: probarObservadoDeArticulos},
	}

	for _, subtest := range subtests {
		t.Run(subtest.nombre, func(t *testing.T) {
			t.Parallel()

			subtest.probar(t)
		})
	}

	for _, caso := range casosDelELI() {
		t.Run("url-eli-"+caso.nombre, func(t *testing.T) {
			t.Parallel()

			probarELIDeLaNorma(t, caso.urlELI, caso.eli)
		})
	}

	// Los bloques distintos de cada verbo que emite; los demás, ninguno.
	bloquesDistintos := map[string]int{"articulo": 1, "articulos": 3}

	for _, verbo := range consultasDeLosSeisVerbos() {
		t.Run("verbo-"+verbo.nombre, func(t *testing.T) {
			t.Parallel()

			resultado := verbo.resuelvePidiendo(t, nuevoBanco(t, reproduce(carpetaDeLasGrabaciones)))

			compruebaLoObservadoPorElVerbo(t, resultado.Grafo, bloquesDistintos[verbo.nombre])
		})
	}
}

// probarObservadoDeArticulo fija lo que observa articulo BOE-A-2015-10565 a21
// sobre sus grabaciones: sus seis operaciones con los ids, los datos y las
// relaciones escritos enteros, la vigencia de siete días y el cuerpo solo en el
// texto; y lo mismo servido de su entrada con --offline.
func probarObservadoDeArticulo(t *testing.T) {
	t.Helper()

	banco := nuevoBanco(t, reproduce(carpetaDeLasGrabaciones))
	consulta := ConsultaArticulo{Norma: normaVigente, Bloque: bloqueDelArticulo21}
	articulo := articuloDelArticulo21(t)

	require.Equal(t, huellaDelArticulo21, articulo.HashTexto)

	const (
		bloque  = eliDeLaLey39 + "#a21"
		version = bloque + "@20161002:" + huellaDelArticulo21
	)

	esperado := schema.Observado{
		Vigencia: 604_800 * time.Second,
		Operaciones: []schema.Operacion{
			schema.Nodo{ID: eliDeLaLey39, Tipo: "Norma", Datos: map[string]any{"identificador": "BOE-A-2015-10565"}},
			schema.Nodo{ID: bloque, Tipo: "Bloque", Datos: map[string]any{"bloque": "a21"}},
			schema.Nodo{ID: version, Tipo: "BloqueVersion", Datos: map[string]any{
				"fecha_vigencia":     "20161002",
				"fecha_version":      "20151002",
				"norma_modificadora": "BOE-A-2015-10565",
				"hash_texto":         huellaDelArticulo21,
			}},
			schema.Arista{Origen: eliDeLaLey39, Relacion: "eli:has_part", Destino: bloque},
			schema.Arista{Origen: bloque, Relacion: "eli:has_version", Destino: version},
			schema.Texto{Huella: huellaDelArticulo21, Cuerpo: articulo.Texto},
		},
	}

	for _, ec := range []schema.Contexto{{}, {Offline: true}} {
		resultado, err := banco.resuelve(t, ec, consulta)

		require.NoError(t, err, "con %+v", ec)
		assert.Equal(t, esperado, resultado.Grafo, "con %+v", ec)
		assert.Equal(t, articulo, resultado.Datos, "con %+v", ec)
		compruebaSinElCuerpo(t, resultado.Grafo)
	}
}

// probarObservadoDeArticulos fija lo que observa articulos BOE-A-2015-10565 a21
// a22 a21: las seis operaciones de a21 y después las de a22, las de a21 una sola
// vez aunque el bloque se pida dos, con la vigencia de siete días; y data, en el
// orden pedido y con la repetición.
func probarObservadoDeArticulos(t *testing.T) {
	t.Helper()

	banco := nuevoBanco(t, reproduce(carpetaDeLasGrabaciones))
	bloques := []string{bloqueDelArticulo21, bloqueDelArticulo22, bloqueDelArticulo21}
	datos := articulosComoArticulo(t, bloques...)

	resultado, err := banco.resuelve(t, schema.Contexto{}, ConsultaArticulos{Norma: normaVigente, Bloques: bloques})

	require.NoError(t, err)
	assert.Equal(t, observadoDeLosBloques(eliDeLaLey39, datos[0], datos[1]), resultado.Grafo)
	assert.Len(t, resultado.Grafo.Operaciones, 12, "seis operaciones por cada uno de los dos bloques distintos")
	assert.Equal(t, datos, resultado.Datos)
	compruebaSinElCuerpo(t, resultado.Grafo)
}

// casoDelELI es un url_eli de los metadatos de la Ley 39/2015 y el id de la
// Norma que sale de él; vacío, ninguno, y entonces el artículo no emite nada.
type casoDelELI struct {
	nombre string
	urlELI string
	eli    string
}

// casosDelELI son los url_eli de los que se prueba de dónde sale el id de la
// Norma (FR-040): de la ruta, desde su primer segmento que es exactamente eli
// hasta el final, sin consulta ni fragmento, con los segmentos tal como van
// escritos en la dirección —una barra escapada no separa dos segmentos—; y los
// que no dan ningún id.
func casosDelELI() []casoDelELI {
	return []casoDelELI{
		{nombre: "vacio", urlELI: ""},
		{nombre: "sin-segmento-eli", urlELI: "https://www.boe.es/buscar/act.php?id=BOE-A-2015-10565"},
		{nombre: "que-no-se-puede-analizar", urlELI: "https://www.boe.es/eli/es/l/2015/10/01/39%zz"},
		{nombre: "segmento-en-mayusculas", urlELI: "https://www.boe.es/ELI/es/l/2015/10/01/39"},
		{nombre: "segmento-que-empieza-por-eli", urlELI: "https://www.boe.es/elis/es/l/2015/10/01/39"},
		{nombre: "eli-solo-en-la-consulta-y-el-fragmento", urlELI: "https://www.boe.es/buscar/act.php?ruta=/eli/es#/eli/es"},
		{nombre: "barra-escapada", urlELI: "https://www.boe.es/eli%2Fes/l/2015/10/01/39"},
		{nombre: "de-la-ley-39", urlELI: "https://www.boe.es/eli/es/l/2015/10/01/39", eli: eliDeLaLey39},
		{nombre: "con-consulta-y-fragmento", urlELI: "https://www.boe.es/eli/es/l/2015/10/01/39?idioma=es#a21", eli: eliDeLaLey39},
		{nombre: "desde-el-primer-segmento-eli", urlELI: "https://www.boe.es/datos/eli/es/l/eli/39", eli: "eli/es/l/eli/39"},
		{nombre: "hasta-el-final-de-la-ruta", urlELI: "https://www.boe.es/eli/es/l/2015/10/01/39/con/", eli: eliDeLaLey39 + "/con/"},
	}
}

// probarELIDeLaNorma resuelve articulo BOE-A-2015-10565 a21 con los metadatos de
// sus grabaciones cambiados por unos cuyo único campo es ese url_eli, y exige
// que data lleve el url_eli tal cual y que lo observado sea el del bloque con ese
// id de la Norma o, si no hay id, nada.
func probarELIDeLaNorma(t *testing.T, urlELI, eli string) {
	t.Helper()

	cuerpo, err := json.Marshal(map[string]map[string]string{"data": {"url_eli": urlELI}})
	require.NoError(t, err)

	banco := nuevoBanco(t, salvoEn(metadatosVigente, respondeConElCuerpo(string(cuerpo))))

	resultado, err := banco.resuelve(t, schema.Contexto{}, ConsultaArticulo{Norma: normaVigente, Bloque: bloqueDelArticulo21})
	require.NoError(t, err)

	datos, esArticulo := resultado.Datos.(Articulo)
	require.Truef(t, esArticulo, "el data de articulo es de tipo %T", resultado.Datos)
	assert.Equal(t, urlELI, datos.URLELI)

	if eli == "" {
		assert.Equal(t, schema.Observado{}, resultado.Grafo, "sin ELI no se emite ninguna operación")

		return
	}

	assert.Equal(t, observadoDeLosBloques(eli, datos), resultado.Grafo)
}

// compruebaLoObservadoPorElVerbo exige lo que observa un verbo sobre las
// grabaciones: nada si no emite, y si emite, la vigencia de siete días y seis
// operaciones por cada bloque distinto, sin el cuerpo en ningún nodo.
func compruebaLoObservadoPorElVerbo(t *testing.T, observado schema.Observado, bloquesDistintos int) {
	t.Helper()

	if bloquesDistintos == 0 {
		assert.Equal(t, schema.Observado{}, observado, "el verbo no emite ninguna operación")

		return
	}

	assert.Equal(t, vigenciaDeLosArticulos, observado.Vigencia)
	assert.Len(t, observado.Operaciones, 6*bloquesDistintos)
	compruebaSinElCuerpo(t, observado)
}

// compruebaSinElCuerpo exige que el cuerpo de cada texto observado viaje solo en
// su Texto: ninguna de sus líneas de lineaMinimaDelCuerpo caracteres o más
// aparece en el id ni en ningún dato de ningún nodo (FR-041). Falla si no hay
// ninguna línea así que comprobar, para no pasar en vacío.
func compruebaSinElCuerpo(t *testing.T, observado schema.Observado) {
	t.Helper()

	var (
		lineas []string
		nodos  []schema.Nodo
	)

	for _, operacion := range observado.Operaciones {
		switch op := operacion.(type) {
		case schema.Texto:
			lineas = append(lineas, lineasLargas(op.Cuerpo)...)
		case schema.Nodo:
			nodos = append(nodos, op)
		}
	}

	require.NotEmpty(t, lineas, "hay líneas del cuerpo que buscar en los nodos")
	require.NotEmpty(t, nodos)

	for _, nodo := range nodos {
		for _, linea := range lineas {
			assert.NotContains(t, nodo.ID, linea, "el id del nodo no lleva el cuerpo")

			for clave, valor := range nodo.Datos {
				assert.NotContains(t, fmt.Sprint(valor), linea, "el dato %s del nodo %s no lleva el cuerpo", clave, nodo.ID)
			}
		}
	}
}

// lineasLargas son las líneas del cuerpo de lineaMinimaDelCuerpo caracteres o
// más.
func lineasLargas(cuerpo string) []string {
	var largas []string

	for linea := range strings.SplitSeq(cuerpo, "\n") {
		if utf8.RuneCountInString(linea) >= lineaMinimaDelCuerpo {
			largas = append(largas, linea)
		}
	}

	return largas
}

// operacionesDelBloque son las seis operaciones que emite un artículo de la
// norma con ese ELI, compuestas aquí con las formas de contracts/emision.md §1:
// la Norma, el Bloque y la BloqueVersion con sus datos, la arista eli:has_part de
// la Norma al Bloque, la eli:has_version del Bloque a la BloqueVersion y el texto
// por su hash_texto.
func operacionesDelBloque(eli string, articulo Articulo) []schema.Operacion {
	bloque := eli + "#" + articulo.Bloque
	version := bloque + "@" + articulo.FechaVigencia + ":" + articulo.HashTexto

	return []schema.Operacion{
		schema.Nodo{ID: eli, Tipo: "Norma", Datos: map[string]any{"identificador": articulo.Norma}},
		schema.Nodo{ID: bloque, Tipo: "Bloque", Datos: map[string]any{"bloque": articulo.Bloque}},
		schema.Nodo{ID: version, Tipo: "BloqueVersion", Datos: map[string]any{
			"fecha_vigencia":     articulo.FechaVigencia,
			"fecha_version":      articulo.FechaVersion,
			"norma_modificadora": articulo.NormaModificadora,
			"hash_texto":         articulo.HashTexto,
		}},
		schema.Arista{Origen: eli, Relacion: "eli:has_part", Destino: bloque},
		schema.Arista{Origen: bloque, Relacion: "eli:has_version", Destino: version},
		schema.Texto{Huella: articulo.HashTexto, Cuerpo: articulo.Texto},
	}
}

// observadoDeLosBloques es lo que observan los artículos de la norma con ese
// ELI, cada uno de un bloque distinto: sus operaciones de operacionesDelBloque,
// en su orden, con la vigencia de la entrada de articulo, siete días (FR-065).
func observadoDeLosBloques(eli string, articulos ...Articulo) schema.Observado {
	var operaciones []schema.Operacion

	for _, articulo := range articulos {
		operaciones = append(operaciones, operacionesDelBloque(eli, articulo)...)
	}

	return schema.Observado{Vigencia: vigenciaDeLosArticulos, Operaciones: operaciones}
}
