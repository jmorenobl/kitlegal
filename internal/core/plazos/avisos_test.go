package plazos

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAvisosDePlazos comprueba los códigos de aviso de los tipos, en el orden de
// contracts/applet-plazos.md §5 y, desde H36, de los cómputos en el fichero
// (FR-009 de H36), sus etiquetas y sus textos de forma fija, una
// línea que empieza por «⚠ <ETIQUETA>:», con los años, el motivo y hasta dónde
// llegan los datos de su ámbito (FR-026, FR-048; research D6).
func TestAvisosDePlazos(t *testing.T) {
	t.Parallel()

	t.Run("los códigos, en el orden del contrato", func(t *testing.T) {
		t.Parallel()
		comprobarCodigos(t)
	})
	t.Run("las etiquetas", func(t *testing.T) {
		t.Parallel()
		comprobarEtiquetas(t)
	})
	t.Run("los años", func(t *testing.T) {
		t.Parallel()
		comprobarAnios(t)
	})
	t.Run("los motivos", func(t *testing.T) {
		t.Parallel()
		comprobarMotivos(t)
	})
	t.Run("los tramos sin festivos", func(t *testing.T) {
		t.Parallel()
		comprobarSegmentos(t)
	})
	t.Run("los textos", func(t *testing.T) {
		t.Parallel()
		comprobarTextos(t)
	})
}

// periodoDeAvisoDeLaLPAC es un periodo de aviso que la copia del contrato no
// tiene, el único de lpac, que nombra uno de los preceptos del cómputo: con él,
// los dos cómputos de los tipos del contrato dan un aviso de periodo.
const periodoDeAvisoDeLaLPAC = `    periodos:
      - desde: "03-01"
        hasta: "03-02"
        efecto: aviso
        precepto: {norma: BOE-A-2015-10565, apartado: "30.2"}
        aviso: {codigo: aviso-de-la-lpac, etiqueta: AVISO DE LA LPAC, texto: "el plazo pasa por el 1 o el 2 de marzo."}
`

// comprobarCodigos exige los códigos de CodigosDeAviso (FR-009 de H36; research
// D3): los cuatro de hoy con los tipos del contrato; los tres del motor sin
// tipos o con tipos cuyo cómputo no tiene avisos de periodo; el del ensayo con
// los suyos; con dos cómputos que avisan, sus avisos en el orden de los cómputos
// en el fichero y no en el de los tipos; y una lista nueva en cada llamada.
func comprobarCodigos(t *testing.T) {
	t.Helper()

	tipos := tiposDelContrato(t)
	delMotor := []string{"faltan-festivos-nacionales-y-autonomicos", "faltan-festivos-locales", "festivo-de-una-parte"}

	codigos := CodigosDeAviso(tipos)
	assert.Equal(t, []string{
		"faltan-festivos-nacionales-y-autonomicos", "faltan-festivos-locales", "dias-del-art-183-lopj",
		"festivo-de-una-parte",
	}, codigos)

	codigos[0] = "cambiado"
	assert.Equal(t, AvisoFaltanNacionalesYAutonomicos, CodigosDeAviso(tipos)[0], "cada llamada da una lista nueva")

	assert.Equal(t, delMotor, CodigosDeAviso(nil), "sin tipos, los tres del motor")
	assert.Equal(t, delMotor, CodigosDeAviso(tipos[:2]), "alzada y reposicion cuentan con lpac, que no avisa")

	delEnsayo, err := LeerTipos(computoDeEnsayo)
	require.NoError(t, err)
	assert.Equal(t, []string{
		"faltan-festivos-nacionales-y-autonomicos", "faltan-festivos-locales", "periodo-de-ensayo",
		"festivo-de-una-parte",
	}, CodigosDeAviso(delEnsayo))

	conDos, err := LeerTipos([]byte((&cambioDeTexto{
		de: "    unidades: [meses, dias-habiles, dias-naturales]\n",
		a:  "    unidades: [meses, dias-habiles, dias-naturales]\n" + periodoDeAvisoDeLaLPAC,
	}).aplicar(t, plazosDelContrato)))
	require.NoError(t, err)
	require.Equal(t, []Computo{"contencioso", "lpac"}, []Computo{conDos[2].Computo, conDos[0].Computo},
		"premisa: el primer tipo cuenta con contencioso, el segundo cómputo del fichero, y el segundo con lpac")
	assert.Equal(t, []string{
		"faltan-festivos-nacionales-y-autonomicos", "faltan-festivos-locales", "aviso-de-la-lpac",
		"dias-del-art-183-lopj", "festivo-de-una-parte",
	}, CodigosDeAviso([]Tipo{conDos[2], conDos[0]}), "en el orden de los cómputos en el fichero")
}

// comprobarEtiquetas exige la etiqueta de cada código de los tipos: las de los
// tres del motor, con tipos o sin ellos, y la de cada aviso de periodo de un
// cómputo de los tipos; y ninguna de lo que no es uno de sus códigos.
func comprobarEtiquetas(t *testing.T) {
	t.Helper()

	tipos := tiposDelContrato(t)

	delEnsayo, err := LeerTipos(computoDeEnsayo)
	require.NoError(t, err)

	for codigo, etiqueta := range map[string]string{
		"faltan-festivos-nacionales-y-autonomicos": "FALTAN FESTIVOS NACIONALES Y AUTONÓMICOS",
		"faltan-festivos-locales":                  "FALTAN FESTIVOS LOCALES",
		"festivo-de-una-parte":                     "FESTIVO DE UNA PARTE",
	} {
		for nombre, de := range map[string][]Tipo{"contrato": tipos, "ensayo": delEnsayo, "ninguno": nil} {
			dada, existe := EtiquetaDeAviso(de, codigo)
			assert.True(t, existe, "%s con los tipos %s", codigo, nombre)
			assert.Equal(t, etiqueta, dada, "%s con los tipos %s", codigo, nombre)
		}
	}

	dada, existe := EtiquetaDeAviso(tipos, "dias-del-art-183-lopj")
	assert.True(t, existe)
	assert.Equal(t, "DÍAS DEL ART. 183 LOPJ", dada)

	dada, existe = EtiquetaDeAviso(delEnsayo, "periodo-de-ensayo")
	assert.True(t, existe)
	assert.Equal(t, "PERIODO DE ENSAYO", dada)

	for _, caso := range []struct {
		tipos  []Tipo
		codigo string
	}{
		{tipos: tipos, codigo: "plazo"},
		{tipos: tipos, codigo: "periodo-de-ensayo"},
		{tipos: tipos[:2], codigo: "dias-del-art-183-lopj"},
		{tipos: nil, codigo: "dias-del-art-183-lopj"},
		{tipos: delEnsayo, codigo: "dias-del-art-183-lopj"},
	} {
		dada, existe := EtiquetaDeAviso(caso.tipos, caso.codigo)
		assert.False(t, existe, "%s con %d tipos", caso.codigo, len(caso.tipos))
		assert.Empty(t, dada, caso.codigo)
	}
}

// comprobarAnios exige las tres formas de los años de un aviso: un año, un
// tramo y, antes del primer año cargado de su ámbito, los anteriores a él.
func comprobarAnios(t *testing.T) {
	t.Helper()

	cargados := &Alcance{Desde: 2041, Hasta: 2043, Disposicion: "PRUEBA-1"}

	for _, caso := range []struct {
		segmento segmento
		alcance  *Alcance
		anios    string
	}{
		{segmento{desde: 2044, hasta: 2044, falta: FaltaSinCargar}, cargados, "de 2044"},
		{segmento{desde: 2044, hasta: 2046, falta: FaltaSinCargar}, cargados, "de 2044 a 2046"},
		{segmento{desde: 2040, hasta: 2040, falta: FaltaSinCargar}, cargados, "anteriores a 2041"},
		{segmento{desde: 2036, hasta: 2040, falta: FaltaSinCargar}, cargados, "anteriores a 2041"},
		{segmento{desde: 2041, hasta: 2042, falta: FaltaNoComunicado}, cargados, "de 2041 a 2042"},
		{segmento{desde: 2040, hasta: 2040, falta: FaltaSinFestivosLocales}, nil, "de 2040"},
	} {
		assert.Equal(t, caso.anios, aniosDe(caso.segmento, caso.alcance), "%+v", caso.segmento)
	}
}

// comprobarMotivos exige el motivo de cada falta de cobertura, vacío con la
// de un año sin cargar.
func comprobarMotivos(t *testing.T) {
	t.Helper()

	for falta, motivo := range map[Falta]string{
		FaltaSinCargar:          "",
		FaltaSinFestivosLocales: ": no tiene festivos locales de Comunidad Marina",
		FaltaNoComunicado:       ": el boletín los da como no comunicados por el ayuntamiento",
		FaltaLineaSinLeer:       ": su línea del boletín no se lee sin interpretarla",
		FaltaSinLinea:           ": el boletín no trae su línea",
	} {
		assert.Equal(t, motivo, motivoDe(falta, "Comunidad Marina"), falta)
	}
}

// comprobarSegmentos exige que los tramos con falta de un ámbito y municipio se
// junten en segmentos de años seguidos con el mismo motivo, y que los demás no
// cuenten.
func comprobarSegmentos(t *testing.T) {
	t.Helper()

	tramos := []Tramo{
		{Ambito: CoberturaNacionalYAutonomica, Desde: 2040, Hasta: 2040, Falta: FaltaSinCargar},
		{Ambito: CoberturaNacionalYAutonomica, Desde: 2041, Hasta: 2041, Disposicion: "PRUEBA-1"},
		{Ambito: CoberturaNacionalYAutonomica, Desde: 2042, Hasta: 2043, Falta: FaltaSinCargar},
		{Ambito: CoberturaLocal, Municipio: "71001", Desde: 2040, Hasta: 2040, Falta: FaltaSinCargar},
		{Ambito: CoberturaLocal, Municipio: "71002", Desde: 2040, Hasta: 2040, Falta: FaltaSinCargar},
		{Ambito: CoberturaLocal, Municipio: "71002", Desde: 2041, Hasta: 2041, Falta: FaltaNoComunicado},
		{Ambito: CoberturaLocal, Municipio: "71002", Desde: 2042, Hasta: 2042, Falta: FaltaNoComunicado},
		{Ambito: CoberturaLocal, Municipio: "71002", Desde: 2043, Hasta: 2044, Falta: FaltaSinCargar},
		{Ambito: CoberturaLocal, Municipio: "71003", Desde: 2041, Hasta: 2041, Disposicion: "PRUEBA-L"},
	}

	assert.Equal(t, []segmento{
		{desde: 2040, hasta: 2040, falta: FaltaSinCargar}, {desde: 2042, hasta: 2043, falta: FaltaSinCargar},
	}, segmentosSinFestivos(tramos, CoberturaNacionalYAutonomica, ""))
	assert.Equal(t, []segmento{
		{desde: 2040, hasta: 2040, falta: FaltaSinCargar},
		{desde: 2041, hasta: 2042, falta: FaltaNoComunicado},
		{desde: 2043, hasta: 2044, falta: FaltaSinCargar},
	}, segmentosSinFestivos(tramos, CoberturaLocal, "71002"))
	assert.Empty(t, segmentosSinFestivos(tramos, CoberturaLocal, "71003"))
}

// comprobarTextos exige cada texto entero, escrito a mano desde el contrato, y
// que cada uno sea una línea que empieza por «⚠ <ETIQUETA>: ».
func comprobarTextos(t *testing.T) {
	t.Helper()

	marina := resuelto("72001", "Puerto Marino", "72", "Comunidad Marina")
	sintetica := resuelto("71001", "Villa Primera", "71", "Comunidad Sintética")
	nacionales := &Alcance{Desde: 2041, Hasta: 2043, Disposicion: "PRUEBA-N-3"}
	locales := &Alcance{Desde: 2041, Hasta: 2042, Disposicion: "PRUEBA-L-9"}
	// El aviso del art. 183 LOPJ es el del periodo de aviso del contencioso, el
	// que LeerTipos lee del contrato (TestLeerTipos).
	_, contencioso := computosDelContrato()

	for _, caso := range []struct {
		aviso Aviso
		texto string
	}{
		{
			avisoDeLosNacionales([]segmento{{desde: 2044, hasta: 2044, falta: FaltaSinCargar}}, nacionales),
			"⚠ FALTAN FESTIVOS NACIONALES Y AUTONÓMICOS: kitlegal no tiene los de 2044; la fecha se ha calculado sin " +
				"ellos y puede ser anterior a la real. Los que tiene llegan a 2043 (PRUEBA-N-3).",
		},
		{
			avisoDeLosNacionales([]segmento{
				{desde: 2039, hasta: 2040, falta: FaltaSinCargar}, {desde: 2044, hasta: 2046, falta: FaltaSinCargar},
			}, nacionales),
			"⚠ FALTAN FESTIVOS NACIONALES Y AUTONÓMICOS: kitlegal no tiene los anteriores a 2041 ni de 2044 a 2046; la " +
				"fecha se ha calculado sin ellos y puede ser anterior a la real. Los que tiene llegan a 2043 (PRUEBA-N-3).",
		},
		{
			avisoDeLosNacionales([]segmento{{desde: 2044, hasta: 2044, falta: FaltaSinCargar}}, nil),
			"⚠ FALTAN FESTIVOS NACIONALES Y AUTONÓMICOS: kitlegal no tiene los de 2044; la fecha se ha calculado sin " +
				"ellos y puede ser anterior a la real.",
		},
		{
			avisoDeLosLocales(marina, []segmento{{desde: 2041, hasta: 2041, falta: FaltaSinFestivosLocales}}, nil),
			"⚠ FALTAN FESTIVOS LOCALES: kitlegal no tiene los de Puerto Marino (72001) de 2041: no tiene festivos " +
				"locales de Comunidad Marina; la fecha se ha calculado sin ellos y puede ser anterior a la real.",
		},
		{
			avisoDeLosLocales(sintetica, []segmento{
				{desde: 2042, hasta: 2042, falta: FaltaLineaSinLeer}, {desde: 2043, hasta: 2043, falta: FaltaSinCargar},
			}, locales),
			"⚠ FALTAN FESTIVOS LOCALES: kitlegal no tiene los de Villa Primera (71001) de 2042: su línea del boletín no " +
				"se lee sin interpretarla ni de 2043; la fecha se ha calculado sin ellos y puede ser anterior a la real. " +
				"Los de Comunidad Sintética llegan a 2042 (PRUEBA-L-9).",
		},
		{
			avisoDeSuPeriodo(*contencioso.Periodos[1].Aviso),
			"⚠ DÍAS DEL ART. 183 LOPJ: el plazo pasa por días del 24 de diciembre al 6 de enero, que el art. 183 LOPJ " +
				"declara inhábiles para las actuaciones judiciales; se han contado, salvo los inhábiles del art. 182.1 " +
				"LOPJ, porque el art. 128.2 LJCA no los excluye del plazo para interponer el recurso. La fecha es la más " +
				"temprana posible.",
		},
		{
			avisoDeUnaParte(dia(t, "2041-06-24"),
				FestivoDeUnaParte{Ambito: AmbitoAutonomico, Parte: "Isla Sintética", Source: "PRUEBA-N-1"}, marina),
			"⚠ FESTIVO DE UNA PARTE: el 2041-06-24 es fiesta solo en Isla Sintética (Comunidad Marina), según " +
				"PRUEBA-N-1; kitlegal no sabe si Puerto Marino está ahí y lo ha contado como hábil. La fecha es la más " +
				"temprana posible.",
		},
	} {
		assert.Equal(t, caso.texto, caso.aviso.Texto, caso.aviso.Codigo)

		etiqueta, existe := EtiquetaDeAviso(tiposDelContrato(t), caso.aviso.Codigo)
		assert.True(t, existe, caso.aviso.Codigo)
		assert.True(t, strings.HasPrefix(caso.aviso.Texto, "⚠ "+etiqueta+": "), caso.aviso.Texto)
		assert.NotContains(t, caso.aviso.Texto, "\n")
	}
}
