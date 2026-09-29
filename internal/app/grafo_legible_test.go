package app

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/grafo"
)

// Los tests del texto que show, stats y check cuentan a una persona sin
// --json (H7.1 FR-060 a FR-063; contracts/applet-graph.md §5): se comparan
// byte a byte, porque la salida es determinista y esa es la garantía. Los
// valores de entrada son los tipos de data, no world.db: lo que aquí se prueba
// es cómo se cuenta, no qué se lee, que ya lo prueban TestAppletGrafo y la
// suite de aceptación. Los ids, las fuentes, las urls y las fechas son los
// inventados de grafo_test.go.

// paraAcotar es la línea con la que termina check sin argumentos
// (contracts/applet-graph.md §5.3).
const paraAcotar = "Para acotar la comprobación a una norma y a sus bloques: kitlegal graph check <norma> [<bloque>...]"

// normaDePrueba es una norma con la forma BOE-A-<año>-<número> que no es
// ninguna real.
const normaDePrueba = "BOE-A-2099-99999"

// procedenciaDeLaNorma es la de las observaciones del primer lote de la muestra.
func procedenciaDeLaNorma() grafo.Procedencia {
	return grafo.Procedencia{Fuente: fuenteDeLaNorma, URL: urlDeLaNorma, FechaConsulta: fechaDeLaNorma}
}

// observacionDeLaNorma es como se cuenta esa procedencia: la fecha, la fuente
// y la url, separadas por un punto medio.
const observacionDeLaNorma = fechaDeLaNorma + " · " + fuenteDeLaNorma + " · " + urlDeLaNorma

// caducadaDePrueba es un hallazgo fuente-caducada del id con la explicación de
// la plantilla de H7 para esa cita.
func caducadaDePrueba(id, cita string) grafo.Hallazgo {
	return grafo.Hallazgo{
		Clase: grafo.ClaseFuenteCaducada,
		ID:    id,
		Explicacion: "La consulta de " + cita + " a " + fuenteDeLaNorma + " en " + urlDeLaNorma + " del " +
			fechaDeLaNorma + " tenía una vigencia de 604800 s y caducó el " + caducidadDeLaNorma + ".",
		Procedencia:      procedenciaDeLaNorma(),
		VigenciaSegundos: 604800,
	}
}

// obsoletaDePrueba es un hallazgo version-obsoleta de la versión del id con la
// explicación de la plantilla de H7.
func obsoletaDePrueba(id string) grafo.Hallazgo {
	return grafo.Hallazgo{
		Clase: grafo.ClaseVersionObsoleta,
		ID:    id,
		Explicacion: "La versión de [" + identificadorDeLaNorma + ", bloque a1] con fecha de vigencia " +
			fechaDeVigenciaDePrueba + " está superada por la de fecha de vigencia 20270101, observada en " +
			urlDeLaNorma + " el " + fechaDeLaNorma + ".",
		Procedencia:           procedenciaDeLaNorma(),
		FechaVigencia:         fechaDeVigenciaDePrueba,
		FechaVigenciaReciente: "20270101",
	}
}

// lineas une las líneas de un texto esperado, cada una con su salto.
func lineas(texto ...string) string {
	return strings.Join(texto, "\n") + "\n"
}

// TestLegibleDeStats fija el texto de stats (FR-061; §5.1): la frase de los
// totales con cada número en singular o en plural, y una sección por grupo con
// algún par, alineado en columna por runas; el grafo vacío, solo la frase.
func TestLegibleDeStats(t *testing.T) {
	t.Parallel()

	t.Run("la plantilla, con sus dos secciones", func(t *testing.T) {
		t.Parallel()

		recuento := grafo.Recuento{
			Nodos: 3, Aristas: 2, Textos: 1,
			NodosPorTipo: []grafo.RecuentoDeNodos{
				{Tipo: grafo.TipoBloque, Fuente: fuenteDeLaNorma, Nodos: 1},
				{Tipo: grafo.TipoBloqueVersion, Fuente: fuenteDeLaNorma, Nodos: 1},
				{Tipo: grafo.TipoNorma, Fuente: fuenteDeLaNorma, Nodos: 1},
			},
			AristasPorRelacion: []grafo.RecuentoDeAristas{
				{Relacion: grafo.RelacionTieneParte, Fuente: fuenteDeLaNorma, Aristas: 1},
				{Relacion: grafo.RelacionTieneVersion, Fuente: fuenteDeLaNorma, Aristas: 1},
			},
		}

		assert.Equal(t, lineas(
			"El grafo del mundo tiene 3 nodos, 2 aristas y 1 texto.",
			"",
			"Nodos por tipo y fuente:",
			"  Bloque         prueba.legislacion  1",
			"  BloqueVersion  prueba.legislacion  1",
			"  Norma          prueba.legislacion  1",
			"",
			"Aristas por relación y fuente:",
			"  eli:has_part     prueba.legislacion  1",
			"  eli:has_version  prueba.legislacion  1",
		), legibleDeStats(recuento))
	})

	t.Run("singular y plural por el número", func(t *testing.T) {
		t.Parallel()

		for _, caso := range []struct {
			nodos, aristas, textos int
			frase                  string
		}{
			{1, 1, 1, "El grafo del mundo tiene 1 nodo, 1 arista y 1 texto."},
			{2, 0, 12, "El grafo del mundo tiene 2 nodos, 0 aristas y 12 textos."},
			{0, 21, 0, "El grafo del mundo tiene 0 nodos, 21 aristas y 0 textos."},
		} {
			texto := legibleDeStats(grafo.Recuento{Nodos: caso.nodos, Aristas: caso.aristas, Textos: caso.textos})

			assert.Equal(t, lineas(caso.frase), texto, "sin pares, ninguna sección")
		}
	})

	t.Run("el grafo vacío es una línea con tres ceros", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "El grafo del mundo tiene 0 nodos, 0 aristas y 0 textos.\n", legibleDeStats(grafo.Recuento{}))
	})

	t.Run("una sección sin pares no se escribe", func(t *testing.T) {
		t.Parallel()

		recuento := grafo.Recuento{
			Nodos: 1, Textos: 0,
			NodosPorTipo: []grafo.RecuentoDeNodos{{Tipo: grafo.TipoMunicipio, Fuente: fuenteDelMunicipio, Nodos: 1}},
		}

		assert.Equal(t, lineas(
			"El grafo del mundo tiene 1 nodo, 0 aristas y 0 textos.",
			"",
			"Nodos por tipo y fuente:",
			"  Municipio  kitlegal.prueba  1",
		), legibleDeStats(recuento))
	})

	t.Run("se alinea en columna por runas y no por bytes", func(t *testing.T) {
		t.Parallel()

		recuento := grafo.Recuento{
			Nodos: 15, Aristas: 3,
			NodosPorTipo: []grafo.RecuentoDeNodos{
				{Tipo: "Órgano", Fuente: "prueba.legislación", Nodos: 12},
				{Tipo: "Norma", Fuente: "kitlegal.prueba", Nodos: 3},
			},
			AristasPorRelacion: []grafo.RecuentoDeAristas{
				{Relacion: "lb:pertenece_a", Fuente: "prueba.legislación", Aristas: 1},
				{Relacion: "eli:has_part", Fuente: "kitlegal.prueba", Aristas: 2},
			},
		}

		assert.Equal(t, lineas(
			"El grafo del mundo tiene 15 nodos, 3 aristas y 0 textos.",
			"",
			"Nodos por tipo y fuente:",
			"  Órgano  prueba.legislación  12",
			"  Norma   kitlegal.prueba     3",
			"",
			"Aristas por relación y fuente:",
			"  lb:pertenece_a  prueba.legislación  1",
			"  eli:has_part    kitlegal.prueba     2",
		), legibleDeStats(recuento))
	})
}

// fichaDePruebaDelBloque es la ficha del Bloque de la muestra: su dato, su
// primera y su última observación, la arista saliente hacia su versión y la
// entrante desde su Norma.
func fichaDePruebaDelBloque() grafo.Ficha {
	return grafo.Ficha{
		Nodo: grafo.NodoDeFicha{
			ID: idDelBloque, Tipo: grafo.TipoBloque, Datos: map[string]any{grafo.DatoBloque: "a1"},
			PrimeraObservacion: fechaDeLaNorma, UltimaObservacion: procedenciaDeLaNorma(),
		},
		Salientes: []grafo.AristaDeFicha{{
			Relacion: grafo.RelacionTieneVersion, ID: idDeLaVersion(),
			PrimeraObservacion: fechaDeLaNorma, UltimaObservacion: procedenciaDeLaNorma(),
		}},
		Entrantes: []grafo.AristaDeFicha{{
			Relacion: grafo.RelacionTieneParte, ID: idDeLaNorma,
			PrimeraObservacion: fechaDeLaNorma, UltimaObservacion: procedenciaDeLaNorma(),
		}},
	}
}

// showLegibleDePruebaDelBloque es lo que cuenta show de esa ficha, con la
// plantilla de §5.2.
func showLegibleDePruebaDelBloque() string {
	return lineas(
		"Bloque "+idDelBloque,
		"  bloque: a1",
		"Primera observación: "+fechaDeLaNorma,
		"Última observación: "+observacionDeLaNorma,
		"",
		"Aristas salientes:",
		"  eli:has_version → "+idDeLaVersion(),
		"    última observación: "+observacionDeLaNorma,
		"",
		"Aristas entrantes:",
		"  eli:has_part ← "+idDeLaNorma,
		"    última observación: "+observacionDeLaNorma,
	)
}

// TestLegibleDeShow fija el texto de show (FR-062; §5.2): el tipo y el id, un
// dato por línea en el orden de sus claves comparando bytes —el valor que no es
// una cadena, en JSON—, la primera y la última observación, y cada arista
// saliente y entrante con su relación, el otro extremo y su última
// observación, o «ninguna».
func TestLegibleDeShow(t *testing.T) {
	t.Parallel()

	t.Run("la plantilla, con sus dos aristas", func(t *testing.T) {
		t.Parallel()

		texto, err := legibleDeShow(fichaDePruebaDelBloque())
		require.NoError(t, err)

		assert.Equal(t, showLegibleDePruebaDelBloque(), texto)
	})

	t.Run("sin aristas en un sentido, ninguna", func(t *testing.T) {
		t.Parallel()

		ficha := grafo.Ficha{Nodo: grafo.NodoDeFicha{
			ID: idDelOrgano, Tipo: grafo.TipoOrgano,
			PrimeraObservacion: fechaDelMunicipio,
			UltimaObservacion: grafo.Procedencia{
				Fuente: fuenteDelMunicipio, URL: urlDelMunicipio, FechaConsulta: "2026-09-30T10:00:00+02:00",
			},
		}}

		texto, err := legibleDeShow(ficha)
		require.NoError(t, err)

		assert.Equal(t, lineas(
			"Organo "+idDelOrgano,
			"Primera observación: "+fechaDelMunicipio,
			"Última observación: 2026-09-30T10:00:00+02:00 · "+fuenteDelMunicipio+" · "+urlDelMunicipio,
			"",
			"Aristas salientes: ninguna.",
			"",
			"Aristas entrantes: ninguna.",
		), texto, "sin datos, ninguna línea de datos")
	})

	t.Run("los datos por clave y en JSON lo que no es una cadena", func(t *testing.T) {
		t.Parallel()

		ficha := grafo.Ficha{Nodo: grafo.NodoDeFicha{
			ID: idDelMunicipio, Tipo: grafo.TipoMunicipio,
			Datos: map[string]any{
				grafo.DatoNombre:    "Villaprueba",
				grafo.DatoCodigoINE: "99001",
				"Zona":              float64(3),
				"activo":            true,
				"nada":              nil,
				"lista":             []any{"a", float64(1)},
				"objeto":            map[string]any{"b": "<&>", "a": 1.5},
				"ñandú":             "después de todas las ASCII",
			},
			PrimeraObservacion: fechaDelMunicipio,
			UltimaObservacion: grafo.Procedencia{
				Fuente: fuenteDelMunicipio, URL: urlDelMunicipio, FechaConsulta: fechaDelMunicipio,
			},
		}}

		texto, err := legibleDeShow(ficha)
		require.NoError(t, err)

		assert.True(t, strings.HasPrefix(texto, lineas(
			"Municipio "+idDelMunicipio,
			"  Zona: 3",
			"  activo: true",
			"  codigo_ine: 99001",
			"  lista: [\"a\",1]",
			"  nada: null",
			"  nombre: Villaprueba",
			`  objeto: {"a":1.5,"b":"<&>"}`,
			"  ñandú: después de todas las ASCII",
			"Primera observación: "+fechaDelMunicipio,
		)), "%s", texto)
	})

	t.Run("las aristas en el orden de data", func(t *testing.T) {
		t.Parallel()

		otra := grafo.Procedencia{Fuente: fuenteDelMunicipio, URL: urlDelMunicipio, FechaConsulta: fechaDelMunicipio}
		ficha := grafo.Ficha{
			Nodo: grafo.NodoDeFicha{
				ID: idDeLaNorma, Tipo: grafo.TipoNorma,
				Datos:              map[string]any{grafo.DatoIdentificador: identificadorDeLaNorma},
				PrimeraObservacion: fechaDeLaNorma, UltimaObservacion: procedenciaDeLaNorma(),
			},
			Salientes: []grafo.AristaDeFicha{
				{Relacion: grafo.RelacionTieneParte, ID: idDelBloque, UltimaObservacion: procedenciaDeLaNorma()},
				{Relacion: grafo.RelacionTieneParte, ID: idDeLaNorma + "#a2", UltimaObservacion: otra},
			},
		}

		texto, err := legibleDeShow(ficha)
		require.NoError(t, err)

		assert.True(t, strings.HasSuffix(texto, lineas(
			"",
			"Aristas salientes:",
			"  eli:has_part → "+idDelBloque,
			"    última observación: "+observacionDeLaNorma,
			"  eli:has_part → "+idDeLaNorma+"#a2",
			"    última observación: "+fechaDelMunicipio+" · "+fuenteDelMunicipio+" · "+urlDelMunicipio,
			"",
			"Aristas entrantes: ninguna.",
		)), "%s", texto)
	})
}

// comprobacionDePrueba es una comprobación de la norma y los bloques con esos
// totales y esos hallazgos listados; los omitidos, los que no se listan.
func comprobacionDePrueba(norma string, bloques []string, obsoletas, caducadas int, listados ...grafo.Hallazgo) grafo.Comprobacion {
	return grafo.Comprobacion{
		Norma: norma, Bloques: bloques, VersionObsoleta: obsoletas, FuenteCaducada: caducadas,
		Omitidos: obsoletas + caducadas - len(listados), Hallazgos: listados,
	}
}

// hallazgosDePruebaDeLasDosClases son, en el orden de check, una versión
// obsoleta y las tres consultas caducadas de la muestra.
func hallazgosDePruebaDeLasDosClases() []grafo.Hallazgo {
	citaDelBloque := "[" + identificadorDeLaNorma + ", bloque a1]"

	return []grafo.Hallazgo{
		obsoletaDePrueba(idDeLaVersion()),
		caducadaDePrueba(idDeLaNorma, identificadorDeLaNorma),
		caducadaDePrueba(idDelBloque, citaDelBloque),
		caducadaDePrueba(idDeLaVersion(), citaDelBloque),
	}
}

// TestLegibleDeCheck fija el texto de check (FR-063; §5.3): la cabecera con su
// ámbito, el total de cada clase y cuántos se listan y se omiten, también
// ninguno; un grupo por clase con algún listado, version-obsoleta primero, con
// el total de la clase, y cada hallazgo con su explicación y su id debajo; sin
// hallazgos, una frase con su ámbito; y, solo sin argumentos, cómo acotar.
func TestLegibleDeCheck(t *testing.T) {
	t.Parallel()

	hallazgos := hallazgosDePruebaDeLasDosClases()

	t.Run("la plantilla, sin argumentos y con las dos clases", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, lineas(
			"Hallazgos en todo lo consultado: 1 version-obsoleta y 3 fuente-caducada; se listan 4 y se omiten 0.",
			"",
			"version-obsoleta (1):",
			"  - "+hallazgos[0].Explicacion,
			"    "+idDeLaVersion(),
			"fuente-caducada (3):",
			"  - "+hallazgos[1].Explicacion,
			"    "+idDeLaNorma,
			"  - "+hallazgos[2].Explicacion,
			"    "+idDelBloque,
			"  - "+hallazgos[3].Explicacion,
			"    "+idDeLaVersion(),
			"",
			paraAcotar,
		), legibleDeCheck(comprobacionDePrueba("", nil, 1, 3, hallazgos...)))
	})

	t.Run("con la norma no dice cómo acotar", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, lineas(
			"Hallazgos de "+normaDePrueba+": 0 version-obsoleta y 1 fuente-caducada; se lista 1 y se omiten 0.",
			"",
			"fuente-caducada (1):",
			"  - "+hallazgos[1].Explicacion,
			"    "+idDeLaNorma,
		), legibleDeCheck(comprobacionDePrueba(normaDePrueba, nil, 0, 1, hallazgos[1])))
	})

	t.Run("con la norma y sus bloques, el ámbito en la cabecera", func(t *testing.T) {
		t.Parallel()

		texto := legibleDeCheck(comprobacionDePrueba(normaDePrueba, []string{"a1", "a2"}, 1, 0, hallazgos[0]))

		assert.Equal(t, lineas(
			"Hallazgos de "+normaDePrueba+", bloques a1 y a2: 1 version-obsoleta y 0 fuente-caducada; se lista 1 y se"+
				" omiten 0.",
			"",
			"version-obsoleta (1):",
			"  - "+hallazgos[0].Explicacion,
			"    "+idDeLaVersion(),
		), texto)
	})

	t.Run("los omitidos cuentan en el total de su clase y un grupo sin listados no se escribe", func(t *testing.T) {
		t.Parallel()

		texto := legibleDeCheck(comprobacionDePrueba("", nil, 240, 4590, hallazgos[0]))

		assert.Equal(t, lineas(
			"Hallazgos en todo lo consultado: 240 version-obsoleta y 4590 fuente-caducada; se lista 1 y se omiten"+
				" 4829.",
			"",
			"version-obsoleta (240):",
			"  - "+hallazgos[0].Explicacion,
			"    "+idDeLaVersion(),
			"",
			paraAcotar,
		), texto)
	})

	t.Run("singular y plural de los listados y los omitidos", func(t *testing.T) {
		t.Parallel()

		for _, caso := range []struct {
			comprobacion grafo.Comprobacion
			recuento     string
		}{
			{comprobacionDePrueba("", nil, 2, 0, hallazgos[0]), "se lista 1 y se omite 1."},
			{comprobacionDePrueba("", nil, 1, 3, hallazgos[0], hallazgos[1]), "se listan 2 y se omiten 2."},
			{comprobacionDePrueba("", nil, 0, 3, hallazgos[1:]...), "se listan 3 y se omiten 0."},
		} {
			cabecera, _, _ := strings.Cut(legibleDeCheck(caso.comprobacion), "\n")

			assert.True(t, strings.HasSuffix(cabecera, "; "+caso.recuento), "%q", cabecera)
		}
	})

	t.Run("sin hallazgos, una frase con cada uno de los cinco ámbitos", func(t *testing.T) {
		t.Parallel()

		for _, caso := range []struct {
			norma   string
			bloques []string
			texto   string
		}{
			{"", nil, lineas("No hay nada que volver a comprobar en todo lo consultado.", "", paraAcotar)},
			{"", []string{}, lineas("No hay nada que volver a comprobar en todo lo consultado.", "", paraAcotar)},
			{normaDePrueba, nil, lineas("No hay nada que volver a comprobar de " + normaDePrueba + ".")},
			{normaDePrueba, []string{"a21"}, lineas("No hay nada que volver a comprobar de " + normaDePrueba +
				", bloque a21.")},
			{normaDePrueba, []string{"a21", "a22"}, lineas("No hay nada que volver a comprobar de " + normaDePrueba +
				", bloques a21 y a22.")},
			{normaDePrueba, []string{"a21", "a22", "a23"}, lineas("No hay nada que volver a comprobar de " +
				normaDePrueba + ", bloques a21, a22 y a23.")},
			{normaDePrueba, []string{"a99", "a1", "a99", "da1"}, lineas("No hay nada que volver a comprobar de " +
				normaDePrueba + ", bloques a99, a1, a99 y da1.")},
		} {
			assert.Equal(t, caso.texto, legibleDeCheck(comprobacionDePrueba(caso.norma, caso.bloques, 0, 0)),
				"%q %q", caso.norma, caso.bloques)
		}
	})
}

// TestLegibleDelGrafoSinTabuladoresNiEscapes exige que ningún texto de los tres
// verbos lleve tabuladores ni secuencias de escape y que termine en un salto de
// línea (§5): tampoco cuando un valor de data los lleva —un bloque pedido, que
// solo se rechaza vacío o en blanco, o un dato de un nodo—, que entonces se
// escribe entre comillas y con sus caracteres de control escapados, como los
// escriben los mensajes de error.
func TestLegibleDelGrafoSinTabuladoresNiEscapes(t *testing.T) {
	t.Parallel()

	conControles := grafo.Ficha{Nodo: grafo.NodoDeFicha{
		ID: idDelMunicipio, Tipo: grafo.TipoMunicipio,
		Datos: map[string]any{
			grafo.DatoNombre: "Villa\tprueba",
			"color":          "\x1b[31mrojo\x1b[0m",
			"borrado":        "a\x7fb",
			"c1":             "a\u009bb",
			"lista":          []any{"x\x7f"},
			"sin controles":  "tal cual",
		},
		PrimeraObservacion: fechaDelMunicipio,
	}}

	show, err := legibleDeShow(conControles)
	require.NoError(t, err)

	for _, linea := range []string{
		`  borrado: "a\x7fb"`,
		`  c1: "a\u009bb"`,
		`  color: "\x1b[31mrojo\x1b[0m"`,
		`  lista: "[\"x\x7f\"]"`,
		`  nombre: "Villa\tprueba"`,
		"  sin controles: tal cual",
	} {
		assert.Contains(t, show, "\n"+linea+"\n")
	}

	check := legibleDeCheck(comprobacionDePrueba(normaDePrueba, []string{"a\t1", "a\x1b[31m1", "a1"}, 0, 0))
	assert.Equal(t, "No hay nada que volver a comprobar de "+normaDePrueba+`, bloques "a\t1", "a\x1b[31m1" y a1.`+"\n",
		check)

	fichaDelBloque, err := legibleDeShow(fichaDePruebaDelBloque())
	require.NoError(t, err)

	for _, texto := range []string{
		show,
		check,
		fichaDelBloque,
		legibleDeStats(grafo.Recuento{
			Nodos:        1,
			NodosPorTipo: []grafo.RecuentoDeNodos{{Tipo: "Tipo\tcon\x1bcontroles", Fuente: fuenteDeLaNorma, Nodos: 1}},
		}),
		legibleDeStats(grafo.Recuento{}),
		legibleDeCheck(comprobacionDePrueba("", nil, 1, 3, hallazgosDePruebaDeLasDosClases()...)),
		legibleDeCheck(comprobacionDePrueba("", nil, 0, 0)),
	} {
		assert.NotContains(t, texto, "\t")
		assert.NotContains(t, texto, "\x1b")
		assert.NotContains(t, texto, "\x7f")
		assert.NotContains(t, texto, "\u009b")
		assert.True(t, strings.HasSuffix(texto, "\n"), "termina en salto de línea: %q", texto)
	}
}

// TestLegibleDelGrafoDeterminista exige que los mismos datos den el mismo texto
// dos veces (FR-060), también con los datos de un nodo, que son un mapa y se
// recorren en otro orden en cada vuelta.
func TestLegibleDelGrafoDeterminista(t *testing.T) {
	t.Parallel()

	datos := map[string]any{}
	for _, clave := range []string{"k", "c", "x", "a", "q", "m", "b", "z", "e", "t", "h", "o"} {
		datos[clave] = clave + clave
	}

	ficha := fichaDePruebaDelBloque()
	ficha.Nodo.Datos = datos

	recuento := grafo.Recuento{
		Nodos: 2, NodosPorTipo: []grafo.RecuentoDeNodos{{Tipo: grafo.TipoNorma, Fuente: fuenteDeLaNorma, Nodos: 2}},
	}
	comprobacion := comprobacionDePrueba("", nil, 1, 3, hallazgosDePruebaDeLasDosClases()...)

	primera, err := legibleDeShow(ficha)
	require.NoError(t, err)

	segunda, err := legibleDeShow(ficha)
	require.NoError(t, err)

	assert.Equal(t, primera, segunda, "show")

	stats, otraVez := legibleDeStats(recuento), legibleDeStats(recuento)
	assert.Equal(t, stats, otraVez, "stats")

	check, otraVez := legibleDeCheck(comprobacion), legibleDeCheck(comprobacion)
	assert.Equal(t, check, otraVez, "check")
}
