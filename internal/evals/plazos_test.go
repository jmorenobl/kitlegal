package evals

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Lo que usan las evals y las respuestas sintéticas de TestJuzgarPlazos: el
// binario en el PATH de la sesión, las normas, la cita del art. 124 de la LPAC
// y las líneas de las tres formas fijas de contracts/evals-legal-core.md §2 de
// H9.
const (
	kitlegalDelPath = "/home/runner/go/bin/kitlegal"

	lgtDeLasCoberturas = "BOE-A-2003-23186"
	fechaDelPlazo      = "2026-10-13"

	citaDel124 = "[Ley 39/2015, art. 124, BOE-A-2015-10565, bloque a124]"

	// lineaDelPlazo es la línea del plazo en su forma fija, con la fecha justo
	// detrás de los dos puntos y la cita en la misma línea.
	lineaDelPlazo = "⚠ PLAZO: " + fechaDelPlazo + " — recurso de reposición " + citaDel124

	// declaracionDeLaLGT es la declaración de no cubierto con la norma que rige
	// el procedimiento, y declaracionSinNorma, la de un procedimiento sin norma
	// propia, con la LPAC.
	declaracionDeLaLGT = "⚠ NO CUBIERTO: la aplicación de los tributos se rige por la Ley General Tributaria " +
		"(" + lgtDeLasCoberturas + "), no por el procedimiento común: kitlegal no calcula sus plazos ni sus recursos " +
		"(DA 1ª.2.a LPAC)."
	declaracionSinNorma = "⚠ NO CUBIERTO: la aplicación de los tributos lo regula aparte la DA 1ª.2.a " +
		"(BOE-A-2015-10565): kitlegal no calcula sus plazos ni sus recursos."

	// avisoDeLosLocales es el texto del aviso faltan-festivos-locales de plazos,
	// tal como lo da data (contracts/applet-plazos.md §5 de H9).
	avisoDeLosLocales = "⚠ FALTAN FESTIVOS LOCALES: kitlegal no tiene los de Tordesillas (47165) de 2026; la fecha " +
		"se ha calculado sin ellos y puede ser anterior a la real."
)

// TestJuzgarPlazos fija las tres formas fijas de contracts/evals-legal-core.md §2
// de H9 y su juicio (FR-048, FR-051, FR-057): lo que leen ExtraerLineasDelPlazo,
// ExtraerNoCubierto y ExtraerAvisosDelPlazo, con la tolerancia de las etiquetas
// de los avisos; lo que Juzgar anota de una eval de plazo y de una de cobertura,
// con los motivos de cada esperado que falta; el comando calcular, que satisface
// una orden o una llamada con su tipo que termina bien y que consulta; y plazos
// calcular prohibido, que ejecuta la orden y la llamada.
func TestJuzgarPlazos(t *testing.T) {
	t.Parallel()

	t.Run("lineas-del-plazo", probarLasLineasDelPlazo)
	t.Run("declaraciones-de-no-cubierto", probarLasDeclaracionesDeNoCubierto)
	t.Run("avisos-del-plazo", probarLosAvisosDelPlazo)
	t.Run("eval-de-plazo", probarElJuicioDeUnPlazo)
	t.Run("eval-de-cobertura", probarElJuicioDeUnaCobertura)
	t.Run("comando-calcular", probarElComandoCalcular)
	t.Run("declaraciones-de-plazo-sin-calcular", probarLasDeclaracionesDePlazoSinCalcular)
	t.Run("eval-de-vigencia", probarElJuicioDeUnaVigencia)
}

// declaracionDelPlazoSinCalcularSintetica es la declaración de un plazo sin
// calcular con los preceptos de la LPAC (ADR 0042).
const declaracionDelPlazoSinCalcularSintetica = "⚠ PLAZO SIN CALCULAR: recurso potestativo de reposición que " +
	"empieza el 2014-06-02: kitlegal no lo calcula, porque lo cuenta con preceptos que ese día no decían lo que dicen " +
	"hoy: BOE-A-2015-10565, bloque a124, apartado 124.1, desde el 2016-10-02."

// probarLasDeclaracionesDePlazoSinCalcular fija ExtraerPlazoSinCalcular: cada
// identificador BOE-A-… entero de cada línea con la forma fija ⚠ PLAZO SIN
// CALCULAR:, en su orden, con la tolerancia de la etiqueta. El identificador de
// otra línea no cuenta, ni el de la línea del plazo ni el de la declaración de
// no cubierto, que tienen otra etiqueta (ADR 0042).
func probarLasDeclaracionesDePlazoSinCalcular(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		respuesta string
		normas    []string
	}{
		{
			nombre:    "forma-fija",
			respuesta: "No puedo darte la fecha.\n\n" + declaracionDelPlazoSinCalcularSintetica + "\n",
			normas:    []string{"BOE-A-2015-10565"},
		},
		{
			nombre:    "con-enfasis-y-dos-normas",
			respuesta: "**⚠️ Plazo sin calcular:** BOE-A-2015-10565, bloque a30; BOE-A-1985-12666, bloque acientoochentaydos.",
			normas:    []string{"BOE-A-2015-10565", "BOE-A-1985-12666"},
		},
		{nombre: "norma-en-otra-linea", respuesta: "⚠ PLAZO SIN CALCULAR: el recurso de reposición.\n(BOE-A-2015-10565)"},
		{nombre: "la-linea-del-plazo", respuesta: lineaDelPlazo},
		{nombre: "la-de-no-cubierto", respuesta: declaracionSinNorma},
		{nombre: "sin-la-marca", respuesta: "PLAZO SIN CALCULAR: BOE-A-2015-10565."},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.normas, ExtraerPlazoSinCalcular(caso.respuesta))
		})
	}
}

// evalDeVigenciaSintetica es una eval de vigencia de plazos-y-territorio, con
// su comando calcular de un plazo sin calcular (ADR 0042).
func evalDeVigenciaSintetica() Eval {
	return Eval{
		Fichero:          "10-plazo-reposicion-anterior-a-la-lpac.yaml",
		Pregunta:         "¿Hasta cuándo podía presentar el recurso de reposición?",
		Activa:           true,
		Comandos:         []ComandoEsperado{{Applet: "plazos", Verbo: "calcular", Tipo: "reposicion", SinCalcular: true}},
		PlazoSinCalcular: PlazoSinCalcularEsperado{Norma: "BOE-A-2015-10565"},
	}
}

// probarElJuicioDeUnaVigencia fija el juicio de una eval de vigencia con la
// sesión cuya orden plazos calcular termina con 3: con la declaración del plazo
// sin calcular con su norma, pasa; sin ella, o con la de otra norma, no, con el
// motivo de la declaración ausente; y con una línea del plazo, tampoco, con su
// motivo (ADR 0042).
func probarElJuicioDeUnaVigencia(t *testing.T) {
	t.Parallel()

	calcula := calculaElPlazo(t, codigoDeSalida(3), "reposicion", "2014-06-02", "Leganés", "Leganés", "--json")
	ausente := "declaración de plazo sin calcular ausente: BOE-A-2015-10565"

	casos := []struct {
		nombre    string
		respuesta string
		juicio    juicioDelPlazo
		pasa      bool
	}{
		{
			nombre: "con-la-declaración", respuesta: declaracionDelPlazoSinCalcularSintetica,
			juicio: juicioDelPlazo{plazoSinCalcularDeclarado: true}, pasa: true,
		},
		{nombre: "sin-declaración", respuesta: "No se puede calcular.", juicio: juicioDelPlazo{motivos: []string{ausente}}},
		{
			nombre:    "con-la-de-otra-norma",
			respuesta: "⚠ PLAZO SIN CALCULAR: recurso de reposición: BOE-A-1992-26318.",
			juicio:    juicioDelPlazo{motivos: []string{ausente}},
		},
		{
			nombre:    "con-una-linea-del-plazo",
			respuesta: declaracionDelPlazoSinCalcularSintetica + "\n\n" + lineaDelPlazo,
			juicio: juicioDelPlazo{
				plazoSinCalcularDeclarado: true, lineas: []string{fechaDelPlazo},
				motivos: []string{"línea del plazo en un plazo sin calcular: " + fechaDelPlazo},
			},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			resultado := Juzgar(evalDeVigenciaSintetica(), sesionDePlazosYTerritorio(caso.respuesta, calcula), skillDeTerritorio)
			require.Equal(t, []string{"plazos calcular reposicion"}, resultado.ComandosEjecutados,
				"premisa: la orden que termina con 3 satisface el comando de un plazo sin calcular")

			juicio, pasa := delJuicio(resultado)
			assert.Equal(t, caso.juicio, juicio)
			assert.Equal(t, caso.pasa, pasa)
		})
	}
}

// probarLasLineasDelPlazo fija ExtraerLineasDelPlazo: la fecha de cada línea con
// la forma fija ⚠ PLAZO:, la fecha AAAA-MM-DD justo detrás de los dos puntos y
// alguna cita en la misma línea, en su orden; con los blancos, el énfasis, el
// selector de presentación y las minúsculas que admite la etiqueta de un aviso.
// No es una línea del plazo la que lleva la fecha detrás de la cita y no delante,
// la que no lleva cita o la lleva en otra línea, la de una fecha con otra forma o
// seguida de otra cifra, ni una fecha suelta en el texto.
func probarLasLineasDelPlazo(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		respuesta string
		fechas    []string
	}{
		{nombre: "forma-fija", respuesta: "Puedes recurrir hasta el martes.\n\n" + lineaDelPlazo + "\n", fechas: []string{fechaDelPlazo}},
		{
			nombre:    "con-enfasis-y-selector",
			respuesta: "**⚠️ Plazo:** _" + fechaDelPlazo + "_ — recurso de reposición " + citaDel124,
			fechas:    []string{fechaDelPlazo},
		},
		{
			nombre:    "dos-lineas",
			respuesta: lineaDelPlazo + "\n⚠ PLAZO: 2026-11-10 — recurso de alzada " + citaDel124,
			fechas:    []string{fechaDelPlazo, "2026-11-10"},
		},
		{nombre: "fecha-detras-de-la-cita", respuesta: "⚠ PLAZO: recurso de reposición " + citaDel124 + ", el " + fechaDelPlazo},
		{nombre: "sin-cita", respuesta: "⚠ PLAZO: " + fechaDelPlazo + " — recurso de reposición"},
		{nombre: "cita-en-otra-linea", respuesta: "⚠ PLAZO: " + fechaDelPlazo + " — recurso de reposición\n" + citaDel124},
		{nombre: "fecha-con-otra-forma", respuesta: "⚠ PLAZO: 13/10/2026 — recurso de reposición " + citaDel124},
		{nombre: "fecha-seguida-de-una-cifra", respuesta: "⚠ PLAZO: " + fechaDelPlazo + "0 — recurso de reposición " + citaDel124},
		{nombre: "sin-la-marca", respuesta: "PLAZO: " + fechaDelPlazo + " — recurso de reposición " + citaDel124},
		{nombre: "fecha-suelta", respuesta: "Vence el " + fechaDelPlazo + " " + citaDel124},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.fechas, ExtraerLineasDelPlazo(caso.respuesta))
		})
	}
}

// probarLasDeclaracionesDeNoCubierto fija ExtraerNoCubierto: cada identificador
// BOE-A-… entero de cada línea con la forma fija ⚠ NO CUBIERTO:, en su orden, con
// la tolerancia de la etiqueta; el de la norma que rige el procedimiento y, sin
// ella, el de la LPAC. El identificador de otra línea no cuenta.
func probarLasDeclaracionesDeNoCubierto(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		respuesta string
		normas    []string
	}{
		{nombre: "con-la-norma", respuesta: "No puedo darte la fecha.\n\n" + declaracionDeLaLGT + "\n", normas: []string{lgtDeLasCoberturas}},
		{nombre: "sin-norma", respuesta: declaracionSinNorma, normas: []string{"BOE-A-2015-10565"}},
		{
			nombre:    "con-enfasis",
			respuesta: "**⚠ No cubierto:** se rige por la LGT (" + lgtDeLasCoberturas + ") y la LPAC (BOE-A-2015-10565).",
			normas:    []string{lgtDeLasCoberturas, "BOE-A-2015-10565"},
		},
		{nombre: "identificador-mas-largo", respuesta: "⚠ NO CUBIERTO: se rige por BOE-A-2003-231860.", normas: []string{"BOE-A-2003-231860"}},
		{nombre: "norma-en-otra-linea", respuesta: "⚠ NO CUBIERTO: se rige por la Ley General Tributaria.\n(" + lgtDeLasCoberturas + ")"},
		{nombre: "sin-la-forma", respuesta: "No está cubierto: se rige por la Ley General Tributaria (" + lgtDeLasCoberturas + ")."},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.normas, ExtraerNoCubierto(caso.respuesta))
		})
	}
}

// probarLosAvisosDelPlazo fija ExtraerAvisosDelPlazo: los códigos de
// plazos.CodigosDeAviso cuya etiqueta, en su forma fija, lleva el texto, en el
// orden de CodigosDeAviso; un aviso de vigencia del BOE no es de plazos.
func probarLosAvisosDelPlazo(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		respuesta string
		codigos   []string
	}{
		{nombre: "uno", respuesta: lineaDelPlazo + "\n\n" + avisoDeLosLocales, codigos: []string{"faltan-festivos-locales"}},
		{
			nombre: "en-el-orden-de-plazos",
			respuesta: "⚠ días del art. 183 LOPJ: el plazo pasa por días del 24 de diciembre al 6 de enero.\n" +
				"⚠ FALTAN FESTIVOS NACIONALES Y AUTONÓMICOS: kitlegal no tiene los anteriores a 2025.",
			codigos: []string{"faltan-festivos-nacionales-y-autonomicos", "dias-del-art-183-lopj"},
		},
		{nombre: "sin-la-marca", respuesta: "FALTAN FESTIVOS LOCALES: kitlegal no tiene los de Tordesillas."},
		{nombre: "de-vigencia", respuesta: "⚠ DEROGADA: la norma no está vigente."},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.codigos, ExtraerAvisosDelPlazo(caso.respuesta))
		})
	}
}

// evalDePlazoSintetica es una eval de plazo de plazos-y-territorio con el
// comando calcular y los avisos dados.
func evalDePlazoSintetica(avisos ...string) Eval {
	return Eval{
		Fichero:  "05-plazo-reposicion-leganes.yaml",
		Pregunta: "¿Hasta cuándo puedo presentar el recurso de reposición?",
		Activa:   true,
		Comandos: []ComandoEsperado{{Applet: "plazos", Verbo: "calcular", Tipo: "reposicion"}},
		Plazo:    PlazoEsperado{Fecha: fechaDelPlazo, Avisos: avisos, Razonamiento: "Un mes desde el día siguiente."},
	}
}

// evalDeCoberturaSintetica es una eval de cobertura de plazos-y-territorio, sin
// comandos y con plazos calcular prohibido.
func evalDeCoberturaSintetica() Eval {
	return Eval{
		Fichero:    "07-cobertura-liquidacion-tributaria.yaml",
		Pregunta:   "¿Hasta cuándo puedo recurrir una liquidación de la Agencia Tributaria?",
		Activa:     true,
		Prohibidos: []ComandoProhibido{{Applet: "plazos", Verbo: "calcular"}},
		NoCubierto: NoCubiertoEsperado{Norma: lgtDeLasCoberturas},
	}
}

// calculaElPlazo es la invocación de la traza de plazos calcular con esos
// argumentos tras el applet y el código dado.
func calculaElPlazo(t *testing.T, codigo *int, argumentos ...string) Invocacion {
	t.Helper()

	return invocada(t, codigo, slices.Concat([]string{kitlegalDelPath, "plazos", "calcular"}, argumentos))
}

// sesionDePlazosYTerritorio es la sesión terminada que activó
// plazos-y-territorio, con la respuesta y las invocaciones dadas.
func sesionDePlazosYTerritorio(respuesta string, invocaciones ...Invocacion) Sesion {
	return cambiada(sesionTerminada(false, respuesta, invocaciones...), func(s *Sesion) {
		s.SkillsActivadas = []string{skillDeTerritorio}
	})
}

// lo que del juicio de una eval de plazo, de cobertura o de vigencia miran los
// casos.
type juicioDelPlazo struct {
	plazoEncontrado           bool
	avisosAusentes            []string
	noCubiertoDeclarado       bool
	plazoSinCalcularDeclarado bool
	lineas                    []string
	motivos                   []string
}

// delJuicio es lo que del resultado miran los casos, y si pasa.
func delJuicio(resultado ResultadoDeEval) (juicioDelPlazo, bool) {
	return juicioDelPlazo{
		plazoEncontrado:           resultado.PlazoEncontrado,
		avisosAusentes:            resultado.AvisosDelPlazoAusentes,
		noCubiertoDeclarado:       resultado.NoCubiertoDeclarado,
		plazoSinCalcularDeclarado: resultado.PlazoSinCalcularDeclarado,
		lineas:                    resultado.LineasDelPlazo,
		motivos:                   resultado.Motivos,
	}, resultado.Pasa
}

// probarElJuicioDeUnPlazo fija el juicio de una eval de plazo con la sesión que
// satisface su comando calcular: con la línea del plazo con su fecha y cada
// aviso que espera en su forma fija, pasa; sin la fecha delante de la cita, con
// otra fecha o sin línea, no, con el motivo plazo ausente; y sin un aviso que
// espera, tampoco, con el motivo aviso del plazo ausente, aunque la línea esté.
func probarElJuicioDeUnPlazo(t *testing.T) {
	t.Parallel()

	calcula := calculaElPlazo(t, codigoDeSalida(0), "reposicion", "2026-09-09", "Leganés", "Leganés", "--json")
	ausente := "plazo ausente: " + fechaDelPlazo

	casos := []struct {
		nombre    string
		eval      Eval
		respuesta string
		juicio    juicioDelPlazo
		pasa      bool
	}{
		{
			nombre: "con-la-linea", eval: evalDePlazoSintetica(), respuesta: "Hasta el martes.\n\n" + lineaDelPlazo,
			juicio: juicioDelPlazo{plazoEncontrado: true, lineas: []string{fechaDelPlazo}}, pasa: true,
		},
		{
			nombre:    "sin-la-fecha-delante-de-la-cita",
			eval:      evalDePlazoSintetica(),
			respuesta: "⚠ PLAZO: recurso de reposición " + citaDel124 + ", hasta el " + fechaDelPlazo,
			juicio:    juicioDelPlazo{motivos: []string{ausente}},
		},
		{
			nombre:    "con-otra-fecha",
			eval:      evalDePlazoSintetica(),
			respuesta: "⚠ PLAZO: 2026-10-09 — recurso de reposición " + citaDel124,
			juicio:    juicioDelPlazo{lineas: []string{"2026-10-09"}, motivos: []string{ausente}},
		},
		{nombre: "sin-linea", eval: evalDePlazoSintetica(), respuesta: "Hasta el 13 de octubre.", juicio: juicioDelPlazo{motivos: []string{ausente}}},
		{
			nombre: "con-el-aviso", eval: evalDePlazoSintetica("faltan-festivos-locales"),
			respuesta: lineaDelPlazo + "\n\n" + avisoDeLosLocales,
			juicio:    juicioDelPlazo{plazoEncontrado: true, lineas: []string{fechaDelPlazo}}, pasa: true,
		},
		{
			nombre: "sin-el-aviso", eval: evalDePlazoSintetica("faltan-festivos-locales", "dias-del-art-183-lopj"),
			respuesta: lineaDelPlazo + "\n\n" + avisoDeLosLocales,
			juicio: juicioDelPlazo{
				plazoEncontrado: true, avisosAusentes: []string{"dias-del-art-183-lopj"}, lineas: []string{fechaDelPlazo},
				motivos: []string{"aviso del plazo ausente: dias-del-art-183-lopj"},
			},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			resultado := Juzgar(caso.eval, sesionDePlazosYTerritorio(caso.respuesta, calcula), skillDeTerritorio)
			require.Equal(t, []string{"plazos calcular reposicion"}, resultado.ComandosEjecutados,
				"premisa: la orden satisface el comando calcular")

			juicio, pasa := delJuicio(resultado)
			assert.Equal(t, caso.juicio, juicio)
			assert.Equal(t, caso.pasa, pasa)
		})
	}
}

// probarElJuicioDeUnaCobertura fija el juicio de una eval de cobertura: con la
// declaración de no cubierto con su norma, pasa, aunque la respuesta lleve una
// fecha suelta; sin declaración o con la de otra norma, no, con el motivo
// declaración de no cubierto ausente; con una línea del plazo, tampoco, con su
// motivo; y con plazos calcular ejecutado, como orden o como llamada, tampoco,
// con el motivo de un comando prohibido, termine como termine.
func probarElJuicioDeUnaCobertura(t *testing.T) {
	t.Parallel()

	ausente := "declaración de no cubierto ausente: " + lgtDeLasCoberturas
	prohibido := "comando prohibido ejecutado: plazos calcular"

	casos := []struct {
		nombre    string
		respuesta string
		sesion    func(t *testing.T, respuesta string) Sesion
		juicio    juicioDelPlazo
		pasa      bool
	}{
		{nombre: "con-la-declaración", respuesta: declaracionDeLaLGT, juicio: juicioDelPlazo{noCubiertoDeclarado: true}, pasa: true},
		{
			nombre:    "con-una-fecha-suelta",
			respuesta: declaracionDeLaLGT + "\n\nEn el procedimiento común vencería el 2026-10-09 " + citaDel124 + ".",
			juicio:    juicioDelPlazo{noCubiertoDeclarado: true}, pasa: true,
		},
		{nombre: "sin-declaración", respuesta: "Tienes un mes.", juicio: juicioDelPlazo{motivos: []string{ausente}}},
		{nombre: "con-la-de-la-lpac", respuesta: declaracionSinNorma, juicio: juicioDelPlazo{motivos: []string{ausente}}},
		{
			nombre:    "con-una-linea-del-plazo",
			respuesta: declaracionDeLaLGT + "\n\n⚠ PLAZO: 2026-10-09 — recurso de reposición " + citaDel124,
			juicio: juicioDelPlazo{
				noCubiertoDeclarado: true, lineas: []string{"2026-10-09"},
				motivos: []string{"línea del plazo en un asunto no cubierto: 2026-10-09"},
			},
		},
		{
			nombre: "con-la-orden-que-falla", respuesta: declaracionDeLaLGT,
			sesion: func(t *testing.T, respuesta string) Sesion {
				t.Helper()

				return sesionDePlazosYTerritorio(respuesta, calculaElPlazo(t, codigoDeSalida(2), "reposicion", "2026-09-09", "Leganés"))
			},
			juicio: juicioDelPlazo{noCubiertoDeclarado: true, motivos: []string{prohibido}},
		},
		{
			nombre: "con-la-llamada", respuesta: declaracionDeLaLGT,
			sesion: func(t *testing.T, respuesta string) Sesion {
				t.Helper()

				return cambiada(sesionDelModoHerramienta(t, respuesta, llamadaCorrecta("plazos_calcular", argumentosDeReposicion)),
					func(s *Sesion) { s.SkillsActivadas = []string{skillDeTerritorio} })
			},
			juicio: juicioDelPlazo{noCubiertoDeclarado: true, motivos: []string{prohibido}},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			sesion := sesionDePlazosYTerritorio(caso.respuesta)
			modo := ModoOrden

			if caso.sesion != nil {
				sesion = caso.sesion(t, caso.respuesta)

				if len(sesion.Llamadas) > 0 {
					modo = ModoHerramienta
				}
			}

			juicio, pasa := delJuicio(juzgarEnModo(evalDeCoberturaSintetica(), sesion, skillDeTerritorio, modo))
			assert.Equal(t, caso.juicio, juicio)
			assert.Equal(t, caso.pasa, pasa)
		})
	}
}

// argumentosDeReposicion son los de una llamada a plazos_calcular con el tipo
// reposicion, como los manda el agente.
const argumentosDeReposicion = `{"tipo":"reposicion","inicio":"2026-09-09","municipios":["Leganés","Leganés"]}`

// probarElComandoCalcular fija cuándo una invocación satisface el comando
// calcular de una eval: la orden plazos calcular con su tipo como primer
// argumento de posición —detrás de las banderas del verbo, con su valor pegado o
// en el argumento siguiente, o de un «--» suelto— que termina con 0, o la llamada
// a plazos_calcular con ese tipo cuyo resultado no es un error. No lo satisface
// la de otro tipo, la que termina con otro código, la que pide la ayuda,
// --describe o --dry-run, ni la llamada con un error. El de un plazo sin
// calcular admite además la orden que termina con 3 y la llamada con el error
// no-encontrado, y ninguna otra (ADR 0042).
func probarElComandoCalcular(t *testing.T) {
	t.Parallel()

	const leganes = "Leganés"

	conLaOrden := func(codigo int, argumentos ...string) func(t *testing.T) Sesion {
		return func(t *testing.T) Sesion {
			t.Helper()

			return sesionDePlazosYTerritorio(lineaDelPlazo, calculaElPlazo(t, codigoDeSalida(codigo), argumentos...))
		}
	}

	conLaLlamada := func(llamada Llamada) func(t *testing.T) Sesion {
		return func(t *testing.T) Sesion {
			t.Helper()

			return cambiada(sesionDelModoHerramienta(t, lineaDelPlazo, llamada),
				func(s *Sesion) { s.SkillsActivadas = []string{skillDeTerritorio} })
		}
	}

	casos := []struct {
		nombre      string
		tipo        string
		sinCalcular bool
		sesion      func(t *testing.T) Sesion
		satisface   bool
	}{
		{nombre: "orden", tipo: "reposicion", sesion: conLaOrden(0, "reposicion", "2026-09-09", leganes, leganes), satisface: true},
		{nombre: "orden-con-el-3", tipo: "reposicion", sesion: conLaOrden(3, "reposicion", "2014-06-02", leganes, leganes)},
		{
			nombre: "orden-sin-calcular-con-el-3", tipo: "reposicion", sinCalcular: true,
			sesion: conLaOrden(3, "reposicion", "2014-06-02", leganes, leganes), satisface: true,
		},
		{
			nombre: "orden-sin-calcular-con-el-2", tipo: "reposicion", sinCalcular: true,
			sesion: conLaOrden(2, "reposicion", "2014-06-02", leganes),
		},
		{
			nombre: "llamada-sin-calcular-con-no-encontrado", tipo: "reposicion", sinCalcular: true,
			sesion:    conLaLlamada(llamadaFallida("plazos_calcular", argumentosDeReposicion, schema.ClaseNoEncontrado)),
			satisface: true,
		},
		{
			nombre: "orden-con-las-banderas-delante", tipo: "silencio",
			sesion:    conLaOrden(0, "--duracion", "3", "--unidad=meses", "silencio", "2026-07-09", leganes, leganes),
			satisface: true,
		},
		{nombre: "orden-tras-el-terminador", tipo: "reposicion", sesion: conLaOrden(0, "--", "reposicion", "2026-09-09", leganes, leganes), satisface: true},
		{nombre: "orden-de-otro-tipo", tipo: "alzada", sesion: conLaOrden(0, "reposicion", "2026-09-09", leganes, leganes)},
		{nombre: "orden-con-otro-codigo", tipo: "reposicion", sesion: conLaOrden(2, "reposicion", "2026-09-09", leganes)},
		{nombre: "orden-con-la-ayuda", tipo: "reposicion", sesion: conLaOrden(0, "reposicion", "--help")},
		{nombre: "orden-con-describe", tipo: "reposicion", sesion: conLaOrden(0, "reposicion", "--describe")},
		{nombre: "orden-con-dry-run", tipo: "reposicion", sesion: conLaOrden(0, "--dry-run", "reposicion", "2026-09-09", leganes, leganes)},
		{nombre: "llamada", tipo: "reposicion", sesion: conLaLlamada(llamadaCorrecta("plazos_calcular", argumentosDeReposicion)), satisface: true},
		{
			nombre: "llamada-con-las-banderas", tipo: "generico",
			sesion: conLaLlamada(llamadaCorrecta("plazos_calcular",
				`{"tipo":"generico","inicio":"2026-10-07","municipios":["Leganés","Leganés"],"duracion":"5","unidad":"dias-habiles"}`)),
			satisface: true,
		},
		{nombre: "llamada-de-otro-tipo", tipo: "alzada", sesion: conLaLlamada(llamadaCorrecta("plazos_calcular", argumentosDeReposicion))},
		{
			nombre: "llamada-con-un-error", tipo: "reposicion",
			sesion: conLaLlamada(llamadaFallida("plazos_calcular", argumentosDeReposicion, schema.ClaseArgumentos)),
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			sesion := caso.sesion(t)
			modo := ModoOrden

			if len(sesion.Llamadas) > 0 {
				modo = ModoHerramienta
			}

			eval := evalDePlazoSintetica()
			eval.Comandos = []ComandoEsperado{{Applet: "plazos", Verbo: "calcular", Tipo: caso.tipo, SinCalcular: caso.sinCalcular}}
			resultado := juzgarEnModo(eval, sesion, skillDeTerritorio, modo)
			texto := "plazos calcular " + caso.tipo

			if caso.satisface {
				assert.Equal(t, []string{texto}, resultado.ComandosEjecutados)
				assert.Empty(t, resultado.ComandosAusentes)

				return
			}

			assert.Empty(t, resultado.ComandosEjecutados)
			assert.Equal(t, []string{texto}, resultado.ComandosAusentes)
			assert.Contains(t, resultado.Motivos, "comando ausente: "+texto)
		})
	}
}
