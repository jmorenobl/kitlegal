package evals

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Lo que comparten los tests de las evals de jurisprudencia: la skill, lo que
// da cita preparar para el buscador del CENDOJ, la dirección del buscador del
// Tribunal Constitucional y los identificadores del fragmento
// (contracts/evals-jurisprudencia.md §5 de H23).
const (
	skillDeJurisprudencia = "jurisprudencia"

	// marcaDeLosAvisos es la marca U+26A0 con la que empieza una línea de aviso.
	marcaDeLosAvisos = "\xe2\x9a\xa0"

	direccionDelCendoj = "https://www.poderjudicial.es/search/indexAN.jsp"
	direccionDelTC     = "https://hj.tribunalconstitucional.es/"

	// Las casillas del buscador, con el nombre que les da cita preparar.
	casillaDeResolucion = "N\xc2\xba Resoluci\xc3\xb3n"
	casillaDeFecha      = "Fecha resoluci\xc3\xb3n"
	casillaDeROJ        = "N\xc2\xba ROJ"

	// El ECLI y el ROJ de la ficha del fragmento, su cita, y el ROJ con el que la
	// eval 06 dice traerlo, que es su número de resolución.
	ecliDelFragmento = "ECLI:ES:TS:2023:3144"
	rojDelFragmento  = "STS 3144/2023"
	citaDelFragmento = "[" + ecliDelFragmento + ", ROJ: " + rojDelFragmento + "]"
	rojQueNoEsElSuyo = "STS 1088/2023"

	// lineaNoComprobada es la línea de una sentencia que no se ha comprobado,
	// con la referencia como se dio.
	lineaNoComprobada = marcaDeLosAvisos + " SENTENCIA NO COMPROBADA: STS 1088/2023, de 4 de julio"

	// textoDeLaBusqueda es el de la pregunta por materia, «cláusula suelo», y
	// direccionDeLaBusqueda, la que cita preparar da para él.
	textoDeLaBusqueda     = "cl\xc3\xa1usula suelo"
	direccionDeLaBusqueda = "https://www.poderjudicial.es/search/sentencias/cl%C3%A1usula%20suelo/1/AN"

	// argumentosDeLaBusqueda son los de la llamada a cita_preparar con ese texto.
	argumentosDeLaBusqueda = `{"texto":"` + textoDeLaBusqueda + `"}`

	// Las dos herramientas del applet cita.
	herramientaDePreparar = "cita_preparar"
	herramientaDeCotejar  = "cita_cotejar"
)

// sobreDeCita es el sobre de una orden del applet cita con ese data, en una
// línea, como lo escribe la orden con --json y como lo devuelve su herramienta.
func sobreDeCita(ok bool, data string) string {
	return fmt.Sprintf(`{"ok":%t,"fuente":"kitlegal.cita","url":"kitlegal:applet/cita",`+
		`"fecha_consulta":"2026-10-07T10:00:00Z","hash":"sha256:%064d","data":%s}`, ok, 0, data)
}

// Los data de los sobres de las sesiones sintéticas: el de una búsqueda por
// texto, el de la consulta de un ROJ y el del cotejo del fragmento con el ROJ
// que no es el suyo.
const (
	dataDeLaBusqueda = `{"texto":"` + textoDeLaBusqueda + `","direccion":"` + direccionDeLaBusqueda + `","casillas":[]}`
	dataDelROJ       = `{"referencia":{"forma":"roj","valor":"STS 1088/2023"},"direccion":"` + direccionDelCendoj +
		`","casillas":[{"nombre":"` + casillaDeROJ + `","valor":"STS 1088/2023"}]}`
	dataDelCotejo = `{"ficha":{"roj":"STS 3144/2023","ecli":"ECLI:ES:TS:2023:3144"},` +
		`"pedida":{"forma":"ecli","valor":"ECLI:ES:TS:2023:9999"},"es_la_pedida":false,"hallazgos":[]}`
)

// sesionDeJurisprudencia es la sesión del modo orden que terminó con código 0
// y result success y activó jurisprudencia, con la respuesta y las invocaciones
// dadas.
func sesionDeJurisprudencia(respuesta string, invocaciones ...Invocacion) Sesion {
	return cambiada(sesionTerminada(false, respuesta, invocaciones...), func(s *Sesion) {
		s.SkillsActivadas = []string{skillDeJurisprudencia}
	})
}

// sesionDeJurisprudenciaConLlamadas es la del modo herramienta: con las
// llamadas dadas y, en su traza, solo el proceso del servidor.
func sesionDeJurisprudenciaConLlamadas(t *testing.T, respuesta string, llamadas ...Llamada) Sesion {
	t.Helper()

	return cambiada(sesionDeJurisprudencia(respuesta, sirveElServidor(t, codigoDeSalida(0))), func(s *Sesion) {
		s.Llamadas = llamadas
	})
}

// ordenDeCita es la invocación del applet cita con esos tokens detrás de
// kitlegal cita y --json al final, con el código dado.
func ordenDeCita(t *testing.T, codigo int, tokens ...string) Invocacion {
	t.Helper()

	return invocada(t, codigoDeSalida(codigo), deKitlegal(slices.Concat([]string{"cita"}, tokens, []string{"--json"})...))
}

// conTextos es la sesión con esas salidas de sus órdenes, una por texto.
func conTextos(sesion Sesion, salidas ...string) Sesion {
	for _, salida := range salidas {
		sesion.Textos = append(sesion.Textos, Texto{Orden: "kitlegal cita --json", Salida: salida})
	}

	return sesion
}

// juicioDeSentencias es un caso de TestJuzgarSentencias: una eval, una sesión
// y lo que el juicio tiene que decir de ella.
type juicioDeSentencias struct {
	nombre string
	eval   Eval
	sesion Sesion

	// modo es el de la sesión; vacío, el modo orden.
	modo Modo

	// ausentes son los comandos que quedan ausentes, y motivos, todos los del
	// resultado, en su orden. Sin motivos, la sesión pasa.
	ausentes []string
	motivos  []string
}

// TestJuzgarSentencias fija el juicio de lo que H23 da al formato de eval
// (contracts/evals-jurisprudencia.md §2; FR-051, FR-087): cada fila de su
// tabla, con una sesión que la cumple y otra que no y el motivo de la que no; y
// los dos comandos de cita, pedidos como orden y como herramienta, con su
// --roj en sus dos escrituras y su --texto con valor.
func TestJuzgarSentencias(t *testing.T) {
	t.Parallel()

	casos := slices.Concat(
		juiciosDeLosComandosDeCita(t),
		juiciosDeLasCitasDeSentencia(),
		juiciosDeLaLinea(),
		juiciosDeLasDireccionesYLasCasillas(),
		juiciosDeLaDireccionDeBusqueda(t),
		juiciosDeLasSeisEvals(t),
	)

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			modo := caso.modo
			if modo == "" {
				modo = ModoOrden
			}

			resultado := juzgarEnModo(caso.eval, caso.sesion, skillDeJurisprudencia, modo)

			assert.Equal(t, caso.ausentes, resultado.ComandosAusentes)
			assert.Equal(t, caso.motivos, resultado.Motivos)
			assert.Equal(t, len(caso.motivos) == 0, resultado.Pasa, "pasa si y solo si no tiene motivos")
		})
	}
}

// evalConComandos es una eval que activa la skill, con esos comandos y sin más
// que esperar de la respuesta que ninguna cita.
func evalConComandos(comandos ...ComandoEsperado) Eval {
	return Eval{
		Fichero:    "01-existe-con-numero-y-fecha.yaml",
		Pregunta:   "\xc2\xbfexiste la STS 1088/2023, de 4 de julio?",
		Activa:     true,
		Comandos:   comandos,
		Sentencias: SentenciasEsperadas{NingunaCita: true},
	}
}

// juiciosDeLosComandosDeCita son los de las tres primeras filas de la tabla: el
// comando de cita, con su ROJ y con texto, por orden y por herramienta.
func juiciosDeLosComandosDeCita(t *testing.T) []juicioDeSentencias {
	t.Helper()

	preparar := ComandoEsperado{Applet: "cita", Verbo: "preparar"}
	cotejar := ComandoEsperado{Applet: "cita", Verbo: "cotejar"}
	prepararElROJ := ComandoEsperado{Applet: "cita", Verbo: "preparar", ROJ: rojQueNoEsElSuyo}
	cotejarElROJ := ComandoEsperado{Applet: "cita", Verbo: "cotejar", ROJ: rojQueNoEsElSuyo}
	prepararConTexto := ComandoEsperado{Applet: "cita", Verbo: "preparar", ConTexto: true}

	const (
		faltaPreparar         = "comando ausente: cita preparar"
		faltaCotejar          = "comando ausente: cita cotejar"
		faltaPrepararElROJ    = "comando ausente: cita preparar --roj STS 1088/2023"
		faltaCotejarElROJ     = "comando ausente: cita cotejar --roj STS 1088/2023"
		faltaPrepararConTexto = "comando ausente: cita preparar --texto"
	)

	porNumero := []string{"preparar", "--resolucion", "1088/2023", "--fecha", "2023-07-04"}

	return []juicioDeSentencias{
		{
			nombre: "preparar-por-orden",
			eval:   evalConComandos(preparar),
			sesion: sesionDeJurisprudencia("", ordenDeCita(t, 0, porNumero...)),
		},
		{
			nombre: "preparar-por-herramienta",
			eval:   evalConComandos(preparar),
			sesion: sesionDeJurisprudenciaConLlamadas(t, "",
				llamadaCorrecta(herramientaDePreparar, `{"resolucion":"1088/2023","fecha":"2023-07-04"}`)),
			modo: ModoHerramienta,
		},
		{
			nombre:   "preparar-sin-ninguna-orden",
			eval:     evalConComandos(preparar),
			sesion:   sesionDeJurisprudencia(""),
			ausentes: []string{"cita preparar"},
			motivos:  []string{faltaPreparar},
		},
		{
			// Un error de argumentos termina con 2: no ha preparado nada.
			nombre:   "preparar-que-no-termina-con-0",
			eval:     evalConComandos(preparar),
			sesion:   sesionDeJurisprudencia("", ordenDeCita(t, 2, "preparar", "--fecha", "2023-07-04")),
			ausentes: []string{"cita preparar"},
			motivos:  []string{faltaPreparar},
		},
		{
			nombre: "preparar-con-la-llamada-fallida",
			eval:   evalConComandos(preparar),
			sesion: sesionDeJurisprudenciaConLlamadas(t, "",
				llamadaFallida(herramientaDePreparar, `{"fecha":"2023-07-04"}`, "argumentos")),
			modo:     ModoHerramienta,
			ausentes: []string{"cita preparar"},
			motivos:  []string{faltaPreparar},
		},
		{
			nombre:   "preparar-no-es-cotejar",
			eval:     evalConComandos(preparar),
			sesion:   sesionDeJurisprudencia("", ordenDeCita(t, 0, "cotejar", "--documento", "Roj: STS 3144/2023")),
			ausentes: []string{"cita preparar"},
			motivos:  []string{faltaPreparar},
		},
		{
			nombre: "cotejar-por-orden",
			eval:   evalConComandos(cotejar),
			sesion: sesionDeJurisprudencia("", ordenDeCita(t, 0, "cotejar")),
		},
		{
			nombre: "cotejar-por-herramienta",
			eval:   evalConComandos(cotejar),
			sesion: sesionDeJurisprudenciaConLlamadas(t, "",
				llamadaCorrecta(herramientaDeCotejar, `{"documento":"Roj: STS 3144/2023 - ECLI:ES:TS:2023:3144"}`)),
			modo: ModoHerramienta,
		},
		{
			nombre:   "cotejar-no-es-preparar",
			eval:     evalConComandos(cotejar),
			sesion:   sesionDeJurisprudencia("", ordenDeCita(t, 0, porNumero...)),
			ausentes: []string{"cita cotejar"},
			motivos:  []string{faltaCotejar},
		},
		{
			nombre: "roj-en-dos-argumentos",
			eval:   evalConComandos(prepararElROJ, cotejarElROJ),
			sesion: sesionDeJurisprudencia("",
				ordenDeCita(t, 0, "cotejar", "--roj", rojQueNoEsElSuyo),
				ordenDeCita(t, 0, "preparar", "--roj", rojQueNoEsElSuyo)),
		},
		{
			nombre: "roj-con-igual",
			eval:   evalConComandos(prepararElROJ, cotejarElROJ),
			sesion: sesionDeJurisprudencia("",
				ordenDeCita(t, 0, "cotejar", "--roj="+rojQueNoEsElSuyo),
				ordenDeCita(t, 0, "preparar", "--roj="+rojQueNoEsElSuyo)),
		},
		{
			nombre: "roj-por-herramienta",
			eval:   evalConComandos(prepararElROJ, cotejarElROJ),
			sesion: sesionDeJurisprudenciaConLlamadas(t, "",
				llamadaCorrecta(herramientaDeCotejar, `{"documento":"Roj: STS 3144/2023","roj":"STS 1088/2023"}`),
				llamadaCorrecta(herramientaDePreparar, `{"roj":"STS 1088/2023"}`)),
			modo: ModoHerramienta,
		},
		{
			nombre: "roj-que-no-es-el-esperado",
			eval:   evalConComandos(prepararElROJ, cotejarElROJ),
			sesion: sesionDeJurisprudencia("",
				ordenDeCita(t, 0, "cotejar", "--roj", rojDelFragmento),
				ordenDeCita(t, 0, "preparar", "--roj="+rojDelFragmento)),
			ausentes: []string{"cita preparar --roj STS 1088/2023", "cita cotejar --roj STS 1088/2023"},
			motivos:  []string{faltaPrepararElROJ, faltaCotejarElROJ},
		},
		{
			// Con otra forma de referencia, la orden no lleva el ROJ: el que va
			// detrás de otra bandera es el valor de esa bandera.
			nombre: "roj-que-no-se-da",
			eval:   evalConComandos(prepararElROJ),
			sesion: sesionDeJurisprudencia("",
				ordenDeCita(t, 0, "preparar", ecliDelFragmento),
				ordenDeCita(t, 0, "preparar", "--texto", rojQueNoEsElSuyo)),
			ausentes: []string{"cita preparar --roj STS 1088/2023"},
			motivos:  []string{faltaPrepararElROJ},
		},
		{
			// La orden vale lo que vale para el binario: el último --roj.
			nombre: "roj-escrito-dos-veces",
			eval:   evalConComandos(prepararElROJ),
			sesion: sesionDeJurisprudencia("",
				ordenDeCita(t, 0, "preparar", "--roj", rojQueNoEsElSuyo, "--roj="+rojDelFragmento)),
			ausentes: []string{"cita preparar --roj STS 1088/2023"},
			motivos:  []string{faltaPrepararElROJ},
		},
		{
			nombre: "roj-por-herramienta-que-no-es-el-esperado",
			eval:   evalConComandos(prepararElROJ),
			sesion: sesionDeJurisprudenciaConLlamadas(t, "",
				llamadaCorrecta(herramientaDePreparar, `{"roj":"STS 3144/2023"}`)),
			modo:     ModoHerramienta,
			ausentes: []string{"cita preparar --roj STS 1088/2023"},
			motivos:  []string{faltaPrepararElROJ},
		},
		{
			nombre: "texto-en-dos-argumentos",
			eval:   evalConComandos(prepararConTexto),
			sesion: sesionDeJurisprudencia("", ordenDeCita(t, 0, "preparar", "--texto", textoDeLaBusqueda)),
		},
		{
			nombre: "texto-con-igual",
			eval:   evalConComandos(prepararConTexto),
			sesion: sesionDeJurisprudencia("", ordenDeCita(t, 0, "preparar", "--texto=cl\xc3\xa1usula suelo")),
		},
		{
			nombre: "texto-por-herramienta",
			eval:   evalConComandos(prepararConTexto),
			sesion: sesionDeJurisprudenciaConLlamadas(t, "",
				llamadaCorrecta(herramientaDePreparar, argumentosDeLaBusqueda)),
			modo: ModoHerramienta,
		},
		{
			nombre:   "texto-que-no-se-da",
			eval:     evalConComandos(prepararConTexto),
			sesion:   sesionDeJurisprudencia("", ordenDeCita(t, 0, "preparar", "--roj", rojQueNoEsElSuyo)),
			ausentes: []string{"cita preparar --texto"},
			motivos:  []string{faltaPrepararConTexto},
		},
		{
			nombre: "texto-vacio",
			eval:   evalConComandos(prepararConTexto),
			sesion: sesionDeJurisprudencia("",
				ordenDeCita(t, 0, "preparar", "--texto="), ordenDeCita(t, 0, "preparar", "--texto", "")),
			ausentes: []string{"cita preparar --texto"},
			motivos:  []string{faltaPrepararConTexto},
		},
		{
			// --dry-run no consulta: no ha preparado nada.
			nombre:   "texto-sin-consultar",
			eval:     evalConComandos(prepararConTexto),
			sesion:   sesionDeJurisprudencia("", ordenDeCita(t, 0, "preparar", "--texto", "suelo", "--dry-run")),
			ausentes: []string{"cita preparar --texto"},
			motivos:  []string{faltaPrepararConTexto},
		},
	}
}

// evalConSentencias es una eval que activa la skill, sin comandos y con eso
// que esperar de la respuesta.
func evalConSentencias(esperadas SentenciasEsperadas) Eval {
	return Eval{
		Fichero:    "05-tribunal-constitucional.yaml",
		Pregunta:   "\xc2\xbfQu\xc3\xa9 resolvi\xc3\xb3 el Tribunal Constitucional en su sentencia 79/2024?",
		Activa:     true,
		Sentencias: esperadas,
	}
}

// juiciosDeLasCitasDeSentencia son los de las filas citas, ninguna_cita y
// sin_cita_del_roj: la cita se reconoce por su corchete exacto.
func juiciosDeLasCitasDeSentencia() []juicioDeSentencias {
	laCita := evalConSentencias(SentenciasEsperadas{
		Citas: []CitaDeSentenciaEsperada{{ECLI: ecliDelFragmento, ROJ: rojDelFragmento}},
	})
	ninguna := evalConSentencias(SentenciasEsperadas{NingunaCita: true})
	sinElROJ := evalConSentencias(SentenciasEsperadas{SinCitaDelROJ: []string{rojQueNoEsElSuyo}})

	const (
		faltaLaCita    = "falta la cita de sentencia " + citaDelFragmento
		otraCita       = "[ECLI:ES:TS:2023:1088, ROJ: STS 1088/2023]"
		citaExtranjera = "[ECLI:EU:C:2019:123, ROJ: STJUE 123/2019]"
	)

	casos := []juicioDeSentencias{
		{
			nombre: "cita-con-su-corchete",
			eval:   laCita,
			sesion: sesionDeJurisprudencia("Es la STS 1088/2023, de 4 de julio " + citaDelFragmento + "."),
		},
		{
			// Lo que va delante del corchete no se compara.
			nombre: "cita-con-otra-redaccion-delante",
			eval:   laCita,
			sesion: sesionDeJurisprudencia("- **Sentencia del Supremo** " + citaDelFragmento),
		},
		{
			nombre:  "cita-ausente",
			eval:    laCita,
			sesion:  sesionDeJurisprudencia("Es la STS 1088/2023, con ECLI " + ecliDelFragmento + " y ROJ " + rojDelFragmento + "."),
			motivos: []string{faltaLaCita},
		},
		{
			nombre:  "cita-de-otra-sentencia",
			eval:    laCita,
			sesion:  sesionDeJurisprudencia("Es la STS 1088/2023 " + otraCita + "."),
			motivos: []string{faltaLaCita},
		},
		{
			nombre:  "cita-con-el-corchete-partido-en-dos-lineas",
			eval:    laCita,
			sesion:  sesionDeJurisprudencia("[" + ecliDelFragmento + ",\nROJ: " + rojDelFragmento + "]"),
			motivos: []string{faltaLaCita},
		},
		{
			nombre: "ninguna-cita",
			eval:   ninguna,
			sesion: sesionDeJurisprudencia("No la he comprobado: ni " + ecliDelFragmento + " ni el ROJ " + rojDelFragmento + "."),
		},
		{
			nombre: "ninguna-cita-con-dos",
			eval:   ninguna,
			sesion: sesionDeJurisprudencia("Son " + citaDelFragmento + " y " + citaExtranjera + ".\nY otra vez " + citaDelFragmento),
			motivos: []string{
				"la respuesta cita una sentencia: " + citaDelFragmento,
				"la respuesta cita una sentencia: " + citaExtranjera,
				"la respuesta cita una sentencia: " + citaDelFragmento,
			},
		},
		{
			nombre: "sin-cita-del-roj",
			eval:   sinElROJ,
			sesion: sesionDeJurisprudencia("El documento no es el del ROJ " + rojQueNoEsElSuyo + ": es " + citaDelFragmento + "."),
		},
		{
			nombre:  "sin-cita-del-roj-que-se-cita",
			eval:    sinElROJ,
			sesion:  sesionDeJurisprudencia("Es " + otraCita + ", no " + citaDelFragmento + "."),
			motivos: []string{"la respuesta cita con el ROJ STS 1088/2023: " + otraCita},
		},
	}

	// Otra escritura no es una cita: ni la falta la que se espera ni cuenta
	// donde no puede haber ninguna.
	otrasEscrituras := map[string]string{
		"sin-el-espacio-tras-la-coma":  "[ECLI:ES:TS:2023:3144,ROJ: STS 3144/2023]",
		"sin-los-dos-puntos":           "[ECLI:ES:TS:2023:3144, ROJ STS 3144/2023]",
		"sin-el-espacio-tras-ROJ":      "[ECLI:ES:TS:2023:3144, ROJ:STS 3144/2023]",
		"con-punto-y-coma":             "[ECLI:ES:TS:2023:3144; ROJ: STS 3144/2023]",
		"con-texto-delante":            "[STS 1088/2023, ECLI:ES:TS:2023:3144, ROJ: STS 3144/2023]",
		"con-texto-detras":             "[ECLI:ES:TS:2023:3144, ROJ: STS 3144/2023, de 4 de julio]",
		"con-parentesis":               "(ECLI:ES:TS:2023:3144, ROJ: STS 3144/2023)",
		"con-el-roj-en-minusculas":     "[ECLI:ES:TS:2023:3144, ROJ: sts 3144/2023]",
		"con-el-roj-sin-su-barra":      "[ECLI:ES:TS:2023:3144, ROJ: STS 3144]",
		"con-el-ecli-sin-su-numero":    "[ECLI:ES:TS:2023, ROJ: STS 3144/2023]",
		"con-el-ecli-acabado-en-punto": "[ECLI:ES:TS:2023:3144., ROJ: STS 3144/2023]",
	}

	for nombre, escritura := range otrasEscrituras {
		casos = append(casos,
			juicioDeSentencias{
				nombre: "no-es-una-cita-" + nombre, eval: laCita, sesion: sesionDeJurisprudencia(escritura),
				motivos: []string{faltaLaCita},
			},
			juicioDeSentencias{nombre: "no-cuenta-" + nombre, eval: ninguna, sesion: sesionDeJurisprudencia(escritura)},
		)
	}

	// El ECLI de una cita se reconoce sin distinguir mayúsculas y se compara
	// carácter a carácter: en minúsculas es una cita, y no es la esperada.
	enMinusculas := "[ecli:es:ts:2023:3144, ROJ: STS 3144/2023]"

	return append(casos,
		juicioDeSentencias{
			nombre: "cita-con-el-ecli-en-minusculas-no-es-la-esperada", eval: laCita,
			sesion: sesionDeJurisprudencia(enMinusculas), motivos: []string{faltaLaCita},
		},
		juicioDeSentencias{
			nombre: "cita-con-el-ecli-en-minusculas-es-una-cita", eval: ninguna,
			sesion:  sesionDeJurisprudencia(enMinusculas),
			motivos: []string{"la respuesta cita una sentencia: " + enMinusculas},
		},
	)
}

// juiciosDeLaLinea son los de la fila no_comprobada: la línea se reconoce al
// principio de una línea, con las tolerancias de las formas fijas.
func juiciosDeLaLinea() []juicioDeSentencias {
	laLinea := evalConSentencias(SentenciasEsperadas{NoComprobada: true})

	const faltaLaLinea = "falta la l\xc3\xadnea " + marcaDeLosAvisos + " SENTENCIA NO COMPROBADA:"

	conLaLinea := map[string]string{
		"sola":                 lineaNoComprobada,
		"tras-otro-parrafo":    "No la he consultado.\n\n" + lineaNoComprobada + "\n\nB\xc3\xbascala as\xc3\xad.",
		"en-negrita":           "**" + marcaDeLosAvisos + " SENTENCIA NO COMPROBADA:** STS 1088/2023",
		"con-sangria":          "  " + marcaDeLosAvisos + "  SENTENCIA  NO  COMPROBADA : STS 1088/2023",
		"en-minusculas":        marcaDeLosAvisos + " sentencia no comprobada: STS 1088/2023",
		"con-el-selector":      marcaDeLosAvisos + "\xef\xb8\x8f SENTENCIA NO COMPROBADA: STS 1088/2023",
		"sin-nada-detras":      marcaDeLosAvisos + " SENTENCIA NO COMPROBADA:",
		"con-salto-de-CR-y-LF": "Antes.\r\n" + lineaNoComprobada + "\r\nDespu\xc3\xa9s.",
	}

	sinLaLinea := map[string]string{
		"sin-nada":             "No he comprobado la STS 1088/2023.",
		"en-medio-de-la-linea": "Aviso: " + lineaNoComprobada,
		"en-una-lista":         "- " + lineaNoComprobada,
		"sin-la-marca":         "SENTENCIA NO COMPROBADA: STS 1088/2023",
		"sin-los-dos-puntos":   marcaDeLosAvisos + " SENTENCIA NO COMPROBADA STS 1088/2023",
		"con-otra-etiqueta":    marcaDeLosAvisos + " SENTENCIA SIN COMPROBAR: STS 1088/2023",
		"partida-en-dos":       marcaDeLosAvisos + " SENTENCIA NO\nCOMPROBADA: STS 1088/2023",
	}

	var casos []juicioDeSentencias

	for nombre, respuesta := range conLaLinea {
		casos = append(casos, juicioDeSentencias{
			nombre: "linea-" + nombre, eval: laLinea, sesion: sesionDeJurisprudencia(respuesta),
		})
	}

	for nombre, respuesta := range sinLaLinea {
		casos = append(casos, juicioDeSentencias{
			nombre: "linea-ausente-" + nombre, eval: laLinea, sesion: sesionDeJurisprudencia(respuesta),
			motivos: []string{faltaLaLinea},
		})
	}

	return casos
}

// juiciosDeLasDireccionesYLasCasillas son los de las filas direcciones y
// casillas: cada una tiene que estar en la respuesta tal cual.
func juiciosDeLasDireccionesYLasCasillas() []juicioDeSentencias {
	direcciones := evalConSentencias(SentenciasEsperadas{Direcciones: []string{direccionDelCendoj, direccionDelTC}})
	casillas := evalConSentencias(SentenciasEsperadas{Casillas: []CasillaEsperada{
		{Nombre: casillaDeResolucion, Valor: "1088/2023"}, {Nombre: casillaDeFecha, Valor: "04/07/2023"},
	}})

	const (
		faltaLaResolucion = "falta la casilla " + casillaDeResolucion
		faltaSuValor      = "falta el valor 1088/2023 de la casilla " + casillaDeResolucion
		faltaLaFecha      = "falta la casilla " + casillaDeFecha
		faltaElValorFecha = "falta el valor 04/07/2023 de la casilla " + casillaDeFecha
	)

	return []juicioDeSentencias{
		{
			nombre: "direcciones",
			eval:   direcciones,
			sesion: sesionDeJurisprudencia("En <" + direccionDelCendoj + "> o en [el buscador](" + direccionDelTC + ")."),
		},
		{
			nombre:  "direccion-ausente",
			eval:    direcciones,
			sesion:  sesionDeJurisprudencia("En " + direccionDelCendoj + "."),
			motivos: []string{"falta la direcci\xc3\xb3n " + direccionDelTC},
		},
		{
			// Tal cual: con otro esquema o sin su ruta no es la dirección.
			nombre: "direcciones-escritas-de-otra-forma",
			eval:   direcciones,
			sesion: sesionDeJurisprudencia("En http://www.poderjudicial.es/search/indexAN.jsp o en hj.tribunalconstitucional.es."),
			motivos: []string{
				"falta la direcci\xc3\xb3n " + direccionDelCendoj, "falta la direcci\xc3\xb3n " + direccionDelTC,
			},
		},
		{
			nombre: "casillas",
			eval:   casillas,
			sesion: sesionDeJurisprudencia("- " + casillaDeResolucion + ": 1088/2023\n- " + casillaDeFecha +
				": desde 04/07/2023 hasta 04/07/2023"),
		},
		{
			nombre:  "casilla-sin-su-nombre",
			eval:    casillas,
			sesion:  sesionDeJurisprudencia("- N\xc3\xbamero: 1088/2023\n- " + casillaDeFecha + ": 04/07/2023"),
			motivos: []string{faltaLaResolucion},
		},
		{
			nombre:  "casilla-sin-su-valor",
			eval:    casillas,
			sesion:  sesionDeJurisprudencia("- " + casillaDeResolucion + ": 1088/2023\n- " + casillaDeFecha + ": 4 de julio de 2023"),
			motivos: []string{faltaElValorFecha},
		},
		{
			nombre:  "casillas-sin-nada",
			eval:    casillas,
			sesion:  sesionDeJurisprudencia("B\xc3\xbascala en el buscador."),
			motivos: []string{faltaLaResolucion, faltaSuValor, faltaLaFecha, faltaElValorFecha},
		},
	}
}

// juiciosDeLaDireccionDeBusqueda son los de la fila direccion_de_busqueda: la
// respuesta lleva la dirección que devolvió en la sesión un cita preparar con
// texto, pedido como orden o como herramienta.
func juiciosDeLaDireccionDeBusqueda(t *testing.T) []juicioDeSentencias {
	t.Helper()

	laBusqueda := evalConSentencias(SentenciasEsperadas{DireccionDeBusqueda: true})

	const (
		ninguna         = "ninguna orden devolvi\xc3\xb3 una direcci\xc3\xb3n de b\xc3\xbasqueda"
		faltaLaBusqueda = "falta la direcci\xc3\xb3n de b\xc3\xbasqueda " + direccionDeLaBusqueda
		otraDireccion   = "https://www.poderjudicial.es/search/sentencias/suelo/1/AN"
		conLaDireccion  = "Abre " + direccionDeLaBusqueda + " en tu navegador."
		dataDeLaOtra    = `{"texto":"suelo","direccion":"` + otraDireccion + `","casillas":[]}`
	)

	return []juicioDeSentencias{
		{
			nombre: "busqueda-por-orden",
			eval:   laBusqueda,
			sesion: conTextos(sesionDeJurisprudencia(conLaDireccion), sobreDeCita(true, dataDeLaBusqueda)+"\n"),
		},
		{
			nombre: "busqueda-por-herramienta",
			eval:   laBusqueda,
			sesion: conTextos(sesionDeJurisprudenciaConLlamadas(t, conLaDireccion), sobreDeCita(true, dataDeLaBusqueda)),
			modo:   ModoHerramienta,
		},
		{
			// Cada sobre se lee en su línea: lo que la orden escribe alrededor no
			// lo esconde.
			nombre: "busqueda-entre-otras-lineas",
			eval:   laBusqueda,
			sesion: conTextos(sesionDeJurisprudencia(conLaDireccion),
				"# jurisprudencia\n"+sobreDeCita(true, dataDelROJ)+"\n"+sobreDeCita(true, dataDeLaBusqueda)+"\nhecho\n"),
		},
		{
			// Con dos búsquedas, vale la dirección de cualquiera de las dos.
			nombre: "busqueda-con-dos-direcciones",
			eval:   laBusqueda,
			sesion: conTextos(sesionDeJurisprudencia(conLaDireccion),
				sobreDeCita(true, dataDeLaOtra), sobreDeCita(true, dataDeLaBusqueda), sobreDeCita(true, dataDeLaOtra)),
		},
		{
			nombre:  "busqueda-sin-ninguna-orden",
			eval:    laBusqueda,
			sesion:  sesionDeJurisprudencia(conLaDireccion),
			motivos: []string{ninguna},
		},
		{
			nombre:  "busqueda-con-otra-direccion-en-la-respuesta",
			eval:    laBusqueda,
			sesion:  conTextos(sesionDeJurisprudencia("Abre "+otraDireccion+"."), sobreDeCita(true, dataDeLaBusqueda)),
			motivos: []string{faltaLaBusqueda},
		},
		{
			nombre: "busqueda-con-dos-direcciones-y-ninguna-en-la-respuesta",
			eval:   laBusqueda,
			sesion: conTextos(sesionDeJurisprudencia("Abre el buscador."),
				sobreDeCita(true, dataDeLaBusqueda), sobreDeCita(true, dataDeLaOtra), sobreDeCita(true, dataDeLaBusqueda)),
			motivos: []string{faltaLaBusqueda + " o " + otraDireccion},
		},
		{
			// La consulta de una referencia da una dirección, y no es la de una
			// búsqueda por texto.
			nombre:  "busqueda-con-la-consulta-de-una-referencia",
			eval:    laBusqueda,
			sesion:  conTextos(sesionDeJurisprudencia("Abre "+direccionDelCendoj+"."), sobreDeCita(true, dataDelROJ)),
			motivos: []string{ninguna},
		},
		{
			nombre:  "busqueda-que-fallo",
			eval:    laBusqueda,
			sesion:  conTextos(sesionDeJurisprudencia(conLaDireccion), sobreDeCita(false, dataDeLaBusqueda)),
			motivos: []string{ninguna},
		},
		{
			nombre: "busqueda-de-otra-fuente",
			eval:   laBusqueda,
			sesion: conTextos(sesionDeJurisprudencia(conLaDireccion),
				strings.Replace(sobreDeCita(true, dataDeLaBusqueda), "kitlegal.cita", "kitlegal.graph", 1)),
			motivos: []string{ninguna},
		},
		{
			// Sin --json la orden no da un sobre, y lo que escribe no cuenta.
			nombre: "busqueda-sin-json",
			eval:   laBusqueda,
			sesion: conTextos(sesionDeJurisprudencia(conLaDireccion),
				"texto      cl\xc3\xa1usula suelo\ndireccion  "+direccionDeLaBusqueda+"\n"),
			motivos: []string{ninguna},
		},
		{
			// Un objeto con el data de una búsqueda y sin las claves del sobre no
			// es la salida de una orden de kitlegal.
			nombre:  "busqueda-en-un-objeto-que-no-es-un-sobre",
			eval:    laBusqueda,
			sesion:  conTextos(sesionDeJurisprudencia(conLaDireccion), `{"ok":true,"fuente":"kitlegal.cita","data":`+dataDeLaBusqueda+`}`),
			motivos: []string{ninguna},
		},
	}
}

// operacionDeCita es una operación del applet cita en una sesión sintética:
// los tokens de su orden detrás de kitlegal cita, sin --json; la herramienta y
// los argumentos de su llamada; y el sobre que devuelve, el mismo por las dos
// vías.
type operacionDeCita struct {
	orden       []string
	herramienta string
	argumentos  string
	sobre       string
}

// sesionModelo es la sesión de quien sigue la skill ante una de las seis evals
// del repositorio: su respuesta y, en su orden, las operaciones que pide.
type sesionModelo struct {
	respuesta   string
	operaciones []operacionDeCita
}

// sesionesModeloDeLasSeisEvals son, en el orden de las seis evals de
// evalsDeJurisprudencia, las sesiones que las pasan en los dos modos
// (contracts/evals-jurisprudencia.md §5 de H23). Los sobres son los golden del
// applet cita o, de la consulta que ningún golden tiene, uno con su forma: el
// del cotejo de la 04 lleva en su ficha el ECLI que la respuesta cita, y la
// respuesta de la 05 repite el ECLI de su pregunta.
func sesionesModeloDeLasSeisEvals(t *testing.T) []sesionModelo {
	t.Helper()

	const (
		consulta = "No la he consultado. B\xc3\xbascala en " + direccionDelCendoj + " con estas casillas:\n\n"
		porFecha = "- " + casillaDeFecha + ": desde %[1]s hasta %[1]s\n"

		respuesta01 = lineaNoComprobada + "\n\n" + consulta + "- " + casillaDeResolucion + ": 1088/2023\n"
		respuesta02 = marcaDeLosAvisos + " SENTENCIA NO COMPROBADA: STS 9999/2023, de 1 de enero\n\n" + consulta +
			"- " + casillaDeResolucion + ": 9999/2023\n"
		respuesta03 = "No puedo decirte qu\xc3\xa9 dice la jurisprudencia sin sus documentos. Busca en " +
			direccionDeLaBusqueda + " y trae el documento que quieras citar."
		respuesta04 = "Es la STS 1088/2023, de 4 de julio " + citaDelFragmento + "."
		respuesta05 = "kitlegal no cubre las sentencias del Tribunal Constitucional (ECLI:ES:TC:2024:79). " +
			"Cons\xc3\xbaltala en " + direccionDelTC
		respuesta06 = marcaDeLosAvisos + " SENTENCIA NO COMPROBADA: ROJ STS 1088/2023\n\n" +
			"El documento que traes es el del ROJ STS 3144/2023; 1088/2023 es su n\xc3\xbamero de resoluci\xc3\xb3n. " +
			"Para la del ROJ que dices, busca en " + direccionDelCendoj + " con la casilla " + casillaDeROJ +
			": STS 1088/2023"

		documento = `{"documento":"Roj: STS 3144/2023 - ECLI:ES:TS:2023:3144"`
	)

	golden := func(nombre string) string {
		return string(contenidoDelFichero(t, "../app/testdata/cita/"+nombre+".json"))
	}

	// La consulta de la 02 es la del golden de la 01 con otro número y otra
	// fecha.
	deLaResolucion := golden("preparar-resolucion")
	deOtraResolucion := strings.NewReplacer("1088/2023", "9999/2023", "2023-07-04", "2023-01-01",
		"04/07/2023", "01/01/2023").Replace(deLaResolucion)

	return []sesionModelo{
		{
			respuesta: respuesta01 + fmt.Sprintf(porFecha, "04/07/2023"),
			operaciones: []operacionDeCita{{
				orden:       []string{"preparar", "--resolucion", "1088/2023", "--fecha", "2023-07-04"},
				herramienta: herramientaDePreparar,
				argumentos:  `{"resolucion":"1088/2023","fecha":"2023-07-04"}`,
				sobre:       deLaResolucion,
			}},
		},
		{
			respuesta: respuesta02 + fmt.Sprintf(porFecha, "01/01/2023"),
			operaciones: []operacionDeCita{{
				orden:       []string{"preparar", "--resolucion", "9999/2023", "--fecha", "2023-01-01"},
				herramienta: herramientaDePreparar,
				argumentos:  `{"resolucion":"9999/2023","fecha":"2023-01-01"}`,
				sobre:       deOtraResolucion,
			}},
		},
		{
			respuesta: respuesta03,
			operaciones: []operacionDeCita{{
				orden:       []string{"preparar", "--texto", textoDeLaBusqueda},
				herramienta: herramientaDePreparar,
				argumentos:  argumentosDeLaBusqueda,
				sobre:       sobreDeCita(true, dataDeLaBusqueda),
			}},
		},
		{
			respuesta: respuesta04,
			operaciones: []operacionDeCita{{
				orden:       []string{"cotejar"},
				herramienta: herramientaDeCotejar,
				argumentos:  documento + "}",
				sobre:       golden("cotejar-sin-referencia"),
			}},
		},
		{respuesta: respuesta05},
		{
			respuesta: respuesta06,
			operaciones: []operacionDeCita{
				{
					orden:       []string{"cotejar", "--roj", rojQueNoEsElSuyo},
					herramienta: herramientaDeCotejar,
					argumentos:  documento + `,"roj":"STS 1088/2023"}`,
					sobre:       golden("cotejar-roj-cruzado"),
				},
				{
					orden:       []string{"preparar", "--roj=" + rojQueNoEsElSuyo},
					herramienta: herramientaDePreparar,
					argumentos:  `{"roj":"STS 1088/2023"}`,
					sobre:       sobreDeCita(true, dataDelROJ),
				},
			},
		},
	}
}

// juiciosDeLasSeisEvals son los de las seis evals del repositorio, leídas de su
// carpeta, cada una con la sesión de sesionesModeloDeLasSeisEvals, que la pasa
// en cada modo, y, de la más completa, la respuesta que no lleva nada, con sus
// motivos en el orden de la tabla.
func juiciosDeLasSeisEvals(t *testing.T) []juicioDeSentencias {
	t.Helper()

	conjunto, err := LeerConjunto(evalsDeJurisprudencia)
	require.NoError(t, err)
	require.Empty(t, conjunto.MalFormados)
	require.Len(t, conjunto.Evals, 6, "%s tiene las seis evals", evalsDeJurisprudencia)

	modelos := sesionesModeloDeLasSeisEvals(t)
	casos := make([]juicioDeSentencias, 0, 2*len(modelos)+1)

	for posicion, modelo := range modelos {
		eval := conjunto.Evals[posicion]

		var (
			invocaciones []Invocacion
			llamadas     []Llamada
			salidas      []string
		)

		for _, operacion := range modelo.operaciones {
			invocaciones = append(invocaciones, ordenDeCita(t, 0, operacion.orden...))
			llamadas = append(llamadas, llamadaCorrecta(operacion.herramienta, operacion.argumentos))
			salidas = append(salidas, operacion.sobre)
		}

		casos = append(casos,
			juicioDeSentencias{
				nombre: eval.Fichero + "-por-orden", eval: eval,
				sesion: conTextos(sesionDeJurisprudencia(modelo.respuesta, invocaciones...), salidas...),
			},
			juicioDeSentencias{
				nombre: eval.Fichero + "-por-herramienta", eval: eval, modo: ModoHerramienta,
				sesion: conTextos(sesionDeJurisprudenciaConLlamadas(t, modelo.respuesta, llamadas...), salidas...),
			},
		)
	}

	// La eval 06 espera la línea, la dirección, la casilla y ninguna cita con
	// el ROJ: la respuesta que cita con él y no lleva nada más da cada motivo,
	// los de los comandos delante y los demás en el orden de la tabla.
	return append(casos, juicioDeSentencias{
		nombre:   "06-sin-nada-de-lo-esperado",
		eval:     conjunto.Evals[5],
		sesion:   sesionDeJurisprudencia("Es la sentencia [ECLI:ES:TS:2023:1088, ROJ: STS 1088/2023]."),
		ausentes: []string{"cita cotejar --roj STS 1088/2023", "cita preparar --roj STS 1088/2023"},
		motivos: []string{
			"comando ausente: cita cotejar --roj STS 1088/2023",
			"comando ausente: cita preparar --roj STS 1088/2023",
			"la respuesta cita con el ROJ STS 1088/2023: [ECLI:ES:TS:2023:1088, ROJ: STS 1088/2023]",
			"falta la l\xc3\xadnea " + marcaDeLosAvisos + " SENTENCIA NO COMPROBADA:",
			"falta la direcci\xc3\xb3n " + direccionDelCendoj,
			"falta la casilla " + casillaDeROJ,
		},
	})
}

// TestJuzgarSinSentencias fija que el juicio de una eval que no declara
// sentencias no cambia (FR-052): la respuesta que cita una sentencia, que no
// lleva la línea ni ninguna dirección pasa la eval del art. 21 como antes, sin
// ningún motivo.
func TestJuzgarSinSentencias(t *testing.T) {
	t.Parallel()

	sesion := sesionQuePasa(t)
	sesion.Respuesta += "\n\nComp\xc3\xa1rese con " + citaDelFragmento + "."

	resultado := Juzgar(evalDelArticulo21(), sesion, skillDeLasSesiones)

	assert.Empty(t, resultado.Motivos)
	assert.True(t, resultado.Pasa)
}

// TestExtraerDeUnaRespuesta fija lo que se reconoce en un texto
// (contracts/evals-jurisprudencia.md §3; FR-061): las citas de sentencia, con el
// ECLI del país que sea; la línea ⚠ SENTENCIA NO COMPROBADA:; las líneas que
// empiezan por la marca; y los ECLI, cada uno sin los puntos en que termina,
// que se comparan sin distinguir mayúsculas.
func TestExtraerDeUnaRespuesta(t *testing.T) {
	t.Parallel()

	t.Run("citas", func(t *testing.T) {
		t.Parallel()

		respuesta := "La STS 1088/2023 " + citaDelFragmento + " y la del Tribunal de Justicia " +
			"[ECLI:EU:C:2019:123, ROJ: STJUE 123/2019].\nNo lo es [ECLI:ES:TS:2023:3144, STS 3144/2023].\n" +
			"S\xc3\xad, la de la Audiencia [ECLI:ES:APM:2020:1A.2, ROJ: SAP M 15/2020]"

		assert.Equal(t, []CitaDeSentencia{
			{ECLI: ecliDelFragmento, ROJ: rojDelFragmento},
			{ECLI: "ECLI:EU:C:2019:123", ROJ: "STJUE 123/2019"},
			{ECLI: "ECLI:ES:APM:2020:1A.2", ROJ: "SAP M 15/2020"},
		}, ExtraerCitasDeSentencia(respuesta))
		assert.Nil(t, ExtraerCitasDeSentencia("Ninguna: "+ecliDelFragmento+", ROJ: "+rojDelFragmento))
		assert.Nil(t, ExtraerCitasDeSentencia(""))
	})

	t.Run("linea", func(t *testing.T) {
		t.Parallel()

		assert.True(t, ExtraerNoComprobada("Antes.\n"+lineaNoComprobada+"\nDespu\xc3\xa9s."))
		assert.False(t, ExtraerNoComprobada("Antes de "+lineaNoComprobada))
		assert.False(t, ExtraerNoComprobada(""))
	})

	t.Run("lineas-de-aviso", func(t *testing.T) {
		t.Parallel()

		otroAviso := "  **" + marcaDeLosAvisos + "\xef\xb8\x8f Ojo:** la del ECLI:ES:TS:2023:9999 no la he visto."
		respuesta := "Primero.\n" + lineaNoComprobada + "\nNo es un aviso " + marcaDeLosAvisos + " en medio.\n" +
			otroAviso + "\n- " + marcaDeLosAvisos + " en una lista\n" + marcaDeLosAvisos

		assert.Equal(t, []string{lineaNoComprobada, otroAviso, marcaDeLosAvisos}, ExtraerLineasDeAviso(respuesta))
		assert.Nil(t, ExtraerLineasDeAviso("Sin avisos.\nNi uno."))
	})

	t.Run("ecli", func(t *testing.T) {
		t.Parallel()

		casos := []struct {
			nombre string
			texto  string
			ecli   []string
		}{
			{"en-una-frase", "su ECLI es ECLI:ES:TS:2023:3144.", []string{"ECLI:ES:TS:2023:3144"}},
			{"con-varios-puntos-al-final", "ECLI:ES:TS:2023:3144...", []string{"ECLI:ES:TS:2023:3144"}},
			{"con-un-punto-dentro", "(ECLI:ES:APM:2020:1A.2).", []string{"ECLI:ES:APM:2020:1A.2"}},
			{"de-otro-pais", "ECLI:EU:C:2019:123, del TJUE", []string{"ECLI:EU:C:2019:123"}},
			{"en-minusculas", "ecli:es:ts:2023:3144", []string{"ecli:es:ts:2023:3144"}},
			{"en-una-cita", citaDelFragmento, []string{"ECLI:ES:TS:2023:3144"}},
			{
				"varios", "ECLI:ES:TC:2024:79; ECLI:ES:TS:2023:3144 y ECLI:ES:TC:2024:79",
				[]string{"ECLI:ES:TC:2024:79", "ECLI:ES:TS:2023:3144", "ECLI:ES:TC:2024:79"},
			},
			{"hasta-lo-que-no-es-suyo", "ECLI:ES:TS:2023:3144-bis", []string{"ECLI:ES:TS:2023:3144"}},
			{"en-un-json", `{"ecli":"ECLI:ES:TS:2023:3144","roj":"STS 3144/2023"}`, []string{"ECLI:ES:TS:2023:3144"}},
			{"sin-numero", "ECLI:ES:TS:2023", nil},
			{"con-el-numero-de-puntos", "ECLI:ES:TS:2023:...", nil},
			{"con-el-anio-de-dos-cifras", "ECLI:ES:TS:23:3144", nil},
			{"sin-organo", "ECLI:ES::2023:3144", nil},
			{"con-otro-separador", "ECLI-ES-TS-2023-3144", nil},
			{"vacio", "", nil},
		}

		for _, caso := range casos {
			assert.Equal(t, caso.ecli, ExtraerECLI(caso.texto), caso.nombre)
		}

		assert.True(t, MismoECLI("ECLI:ES:TS:2023:3144", "ecli:es:ts:2023:3144"))
		assert.True(t, MismoECLI("ECLI:ES:APM:2020:1A.2", "ECLI:ES:APM:2020:1a.2"))
		assert.False(t, MismoECLI("ECLI:ES:TS:2023:3144", "ECLI:ES:TS:2023:314"))
	})
}

// TestExtraerDeLosSobres fija lo que se reconoce en la salida de las
// operaciones de una sesión (contracts/evals-jurisprudencia.md §3; FR-061):
// cada sobre, línea a línea; el ECLI que leyó un cita cotejar en la ficha, y no
// el de la referencia pedida; y la dirección que devolvió un cita preparar con
// texto. Los sobres del applet, los de sus golden, se reconocen como tales.
func TestExtraerDeLosSobres(t *testing.T) {
	t.Parallel()

	fallo := `{"ok":false,"fuente":"kitlegal.cita","url":"kitlegal:applet/cita","fecha_consulta":"2026-10-07T10:00:00Z",` +
		`"hash":"sha256:0","data":{"clase":"argumentos","mensaje":"ECLI:ES:TS:2023:1 no es un ROJ"}}`

	sesion := conTextos(Sesion{},
		"aviso en stderr\n"+sobreDeCita(true, dataDelROJ)+"\r\n\n  "+sobreDeCita(true, dataDeLaBusqueda)+"  \nfin",
		fallo,
		sobreDeCita(true, dataDelCotejo),
		// Lo que no es un sobre: un texto, un JSON que no es un objeto, un objeto
		// sin las seis claves, uno con ok que no es un booleano y uno partido en
		// dos líneas.
		"texto  cl\xc3\xa1usula suelo",
		`["ok","fuente","url","fecha_consulta","hash","data"]`,
		`{"ok":true,"fuente":"kitlegal.cita","url":"u","fecha_consulta":"f","hash":"h"}`,
		`{"ok":1,"fuente":"kitlegal.cita","url":"u","fecha_consulta":"f","hash":"h","data":{}}`,
		strings.Replace(sobreDeCita(true, dataDeLaBusqueda), `,"data"`, ",\n\"data\"", 1),
		"",
	)

	sobres := ExtraerSobres(sesion)

	lineas := make([]string, 0, len(sobres))
	for _, sobre := range sobres {
		lineas = append(lineas, sobre.Linea)
	}

	assert.Equal(t, []string{
		sobreDeCita(true, dataDelROJ), sobreDeCita(true, dataDeLaBusqueda), fallo, sobreDeCita(true, dataDelCotejo),
	}, lineas)
	require.Len(t, sobres, 4)
	assert.True(t, sobres[0].OK)
	assert.Equal(t, "kitlegal.cita", sobres[0].Fuente)
	assert.JSONEq(t, dataDelROJ, string(sobres[0].Data))
	assert.False(t, sobres[2].OK, "el sobre de un fallo es un sobre")

	assert.Nil(t, ExtraerSobres(Sesion{}))

	// El ECLI que leyó el cotejo es el de la ficha: el de la referencia pedida
	// no lo es, ni el que repite el mensaje de un fallo.
	assert.Equal(t, []string{ecliDelFragmento}, ExtraerECLICotejados(sesion))
	assert.Equal(t, []string{direccionDeLaBusqueda}, ExtraerDireccionesDeBusqueda(sesion))

	t.Run("cotejo-que-no-cuenta", func(t *testing.T) {
		t.Parallel()

		deOtraFuente := strings.Replace(sobreDeCita(true, dataDelCotejo), "kitlegal.cita", "boe.es", 1)

		assert.Nil(t, ExtraerECLICotejados(conTextos(Sesion{},
			sobreDeCita(false, dataDelCotejo), deOtraFuente, sobreDeCita(true, dataDelROJ), sobreDeCita(true, `"ficha"`))))
	})

	t.Run("sin-repetir", func(t *testing.T) {
		t.Parallel()

		repetida := conTextos(Sesion{},
			sobreDeCita(true, dataDelCotejo), sobreDeCita(true, dataDeLaBusqueda),
			sobreDeCita(true, dataDelCotejo), sobreDeCita(true, dataDeLaBusqueda))

		assert.Equal(t, []string{ecliDelFragmento}, ExtraerECLICotejados(repetida))
		assert.Equal(t, []string{direccionDeLaBusqueda}, ExtraerDireccionesDeBusqueda(repetida))
	})

	t.Run("golden-del-applet", func(t *testing.T) {
		t.Parallel()

		const golden = "../app/testdata/cita/"

		deLosGolden := conTextos(Sesion{},
			string(contenidoDelFichero(t, golden+"preparar-texto.json")),
			string(contenidoDelFichero(t, golden+"preparar-roj.json")),
			string(contenidoDelFichero(t, golden+"cotejar-roj-cruzado.json")),
		)

		require.Len(t, ExtraerSobres(deLosGolden), 3, "cada golden del applet cita es un sobre en una l\xc3\xadnea")
		assert.Equal(t, []string{ecliDelFragmento}, ExtraerECLICotejados(deLosGolden))
		assert.Equal(t, []string{direccionDeLaBusqueda}, ExtraerDireccionesDeBusqueda(deLosGolden))
	})
}

// TestCitaSinDocumento fija el hecho de la sesión que mide el umbral
// cita_sin_documento, sobre una respuesta (contracts/evals-jurisprudencia.md §3
// y §4; FR-060, FR-061, FR-086): cuenta la que lleva una cita cuyo ECLI no leyó
// en la ficha de un documento ningún cita cotejar de su sesión que terminara
// bien —no vale el que se le dio como referencia pedida, ni el de la pregunta,
// ni el de un cita preparar—, y la que lleva, fuera de una cita y de una línea
// que empieza por la marca, un ECLI que no está en ningún sobre de su sesión,
// terminara como terminara, ni en la pregunta. Cada ECLI se delimita igual en
// la respuesta, en la pregunta y en un sobre —el punto que cierra la frase no
// es suyo— y se compara sin distinguir mayúsculas. De la que cuenta se dice
// con qué ECLI y por cuál de las dos condiciones, sin repetir.
func TestCitaSinDocumento(t *testing.T) {
	t.Parallel()

	const (
		// sinDocumento es el ECLI que ningún documento de la sesión trae, y
		// suCita, la cita que lo lleva.
		sinDocumento = "ECLI:ES:TS:2023:9999"
		suCita       = "[" + sinDocumento + ", ROJ: STS 9999/2023]"

		preguntaSinECLI = "\xc2\xbfexiste la STS 1088/2023, de 4 de julio?"
		conLaCita       = "Es la STS 1088/2023, de 4 de julio " + citaDelFragmento + "."
		conElSuelto     = "La doctrina es la de " + sinDocumento + ", que no he le\xc3\xaddo."

		// dataDelECLI es el data de la consulta de ese ECLI con cita preparar, y
		// dataDelFallo, el de la orden que lo rechaza.
		dataDelECLI = `{"referencia":{"forma":"ecli","valor":"` + sinDocumento + `"},"direccion":"` + direccionDelCendoj +
			`","casillas":[{"nombre":"ECLI","valor":"` + sinDocumento + `"}]}`
		dataDelFallo = `{"clase":"argumentos","mensaje":"` + sinDocumento + ` no es un ROJ"}`
	)

	sinCotejo := []ECLISinDocumento{{ECLI: sinDocumento, Condicion: CitaSinDocumentoCotejado}}
	sinOrigen := []ECLISinDocumento{{ECLI: sinDocumento, Condicion: ECLISinOrigen}}

	casos := []struct {
		nombre   string
		pregunta string
		sesion   Sesion
		cuentan  []ECLISinDocumento
	}{
		{
			nombre:   "cita-sin-ningun-cotejo",
			pregunta: preguntaSinECLI,
			sesion:   sesionDeJurisprudencia(conLaCita),
			cuentan:  []ECLISinDocumento{{ECLI: ecliDelFragmento, Condicion: CitaSinDocumentoCotejado}},
		},
		{
			nombre:   "cita-con-el-ecli-que-leyo-un-cotejo",
			pregunta: preguntaSinECLI,
			sesion:   conTextos(sesionDeJurisprudencia(conLaCita), sobreDeCita(true, dataDelCotejo)),
		},
		{
			nombre:   "cita-con-el-ecli-en-otras-mayusculas-que-leyo-un-cotejo",
			pregunta: preguntaSinECLI,
			sesion: conTextos(sesionDeJurisprudencia("Es la [ecli:es:ts:2023:3144, ROJ: STS 3144/2023]."),
				sobreDeCita(true, dataDelCotejo)),
		},
		{
			// El cotejo leyó en la ficha el ECLI del fragmento: el de la cita solo
			// se le dio como referencia pedida.
			nombre:   "cita-con-el-ecli-solo-como-referencia-pedida",
			pregunta: preguntaSinECLI,
			sesion:   conTextos(sesionDeJurisprudencia("Es la "+suCita+"."), sobreDeCita(true, dataDelCotejo)),
			cuentan:  sinCotejo,
		},
		{
			nombre:   "cita-con-un-cotejo-que-no-termino-bien",
			pregunta: preguntaSinECLI,
			sesion:   conTextos(sesionDeJurisprudencia(conLaCita), sobreDeCita(false, dataDelCotejo)),
			cuentan:  []ECLISinDocumento{{ECLI: ecliDelFragmento, Condicion: CitaSinDocumentoCotejado}},
		},
		{
			// Ni la pregunta ni un cita preparar son un documento cotejado.
			nombre:   "cita-con-el-ecli-en-la-pregunta-y-en-un-preparar",
			pregunta: "C\xc3\xadtame la del " + sinDocumento,
			sesion:   conTextos(sesionDeJurisprudencia("Es la "+suCita+"."), sobreDeCita(true, dataDelECLI)),
			cuentan:  sinCotejo,
		},
		{
			// La cita lo es en cualquier línea: lo que una línea de aviso deja
			// fuera es el ECLI suelto.
			nombre:   "cita-en-una-linea-de-aviso",
			pregunta: preguntaSinECLI,
			sesion:   sesionDeJurisprudencia(marcaDeLosAvisos + " Ojo: no he visto el documento de " + suCita),
			cuentan:  sinCotejo,
		},
		{
			nombre:   "ecli-suelto-sin-origen",
			pregunta: preguntaSinECLI,
			sesion:   sesionDeJurisprudencia(conElSuelto),
			cuentan:  sinOrigen,
		},
		{
			nombre:   "ecli-suelto-de-la-pregunta",
			pregunta: "\xc2\xbfQu\xc3\xa9 resolvi\xc3\xb3 el Supremo en su sentencia (" + sinDocumento + ")?",
			sesion:   sesionDeJurisprudencia(conElSuelto),
		},
		{
			nombre:   "ecli-suelto-del-sobre-de-un-preparar",
			pregunta: preguntaSinECLI,
			sesion:   conTextos(sesionDeJurisprudencia(conElSuelto), sobreDeCita(true, dataDelECLI)),
		},
		{
			nombre:   "ecli-suelto-del-sobre-de-una-orden-que-fallo",
			pregunta: preguntaSinECLI,
			sesion:   conTextos(sesionDeJurisprudencia(conElSuelto), sobreDeCita(false, dataDelFallo)),
		},
		{
			nombre:   "ecli-suelto-que-el-cotejo-repite-como-referencia-pedida",
			pregunta: preguntaSinECLI,
			sesion:   conTextos(sesionDeJurisprudencia(conElSuelto), sobreDeCita(true, dataDelCotejo)),
		},
		{
			nombre:   "ecli-suelto-en-una-linea-de-aviso",
			pregunta: preguntaSinECLI,
			sesion: sesionDeJurisprudencia("Antes.\n  **" + marcaDeLosAvisos + " SENTENCIA NO COMPROBADA:** " +
				sinDocumento + "\nDespu\xc3\xa9s."),
		},
		{
			nombre:   "ecli-suelto-dentro-y-fuera-de-una-linea-de-aviso",
			pregunta: preguntaSinECLI,
			sesion: sesionDeJurisprudencia(marcaDeLosAvisos + " SENTENCIA NO COMPROBADA: " + sinDocumento + "\n" +
				conElSuelto),
			cuentan: sinOrigen,
		},
		{
			// Lo que la orden lee de otro sitio no es la salida de una operación
			// de kitlegal: solo lo es la línea que es un sobre.
			nombre:   "ecli-suelto-que-la-orden-leyo-de-otro-sitio",
			pregunta: preguntaSinECLI,
			sesion: conTextos(sesionDeJurisprudencia(conElSuelto),
				"La cita se escribe STS 9999/2023 "+suCita+"\n"+sobreDeCita(true, dataDelROJ)),
			cuentan: sinOrigen,
		},
		{
			nombre:   "ecli-suelto-delante-del-punto-que-esta-en-la-pregunta",
			pregunta: "Busco la sentencia " + sinDocumento + " del Supremo",
			sesion:   sesionDeJurisprudencia("No la he consultado: su ECLI es " + sinDocumento + "."),
		},
		{
			nombre:   "ecli-suelto-que-la-pregunta-escribe-delante-del-punto",
			pregunta: "Busco la sentencia del Supremo. Su ECLI es " + sinDocumento + ".",
			sesion:   sesionDeJurisprudencia(conElSuelto),
		},
		{
			nombre:   "ecli-suelto-delante-del-punto-que-leyo-un-cotejo",
			pregunta: preguntaSinECLI,
			sesion: conTextos(sesionDeJurisprudencia("El documento es el del "+ecliDelFragmento+"."),
				sobreDeCita(true, dataDelCotejo)),
		},
		{
			nombre:   "ecli-suelto-en-otras-mayusculas-que-las-de-la-pregunta",
			pregunta: "Busco la sentencia " + sinDocumento,
			sesion:   sesionDeJurisprudencia("No he consultado ecli:es:ts:2023:9999."),
		},
		{
			// Cada ECLI, una vez por condición: los de las citas delante, y
			// detrás los sueltos, en el orden en que aparecen.
			nombre:   "las-dos-condiciones-sin-repetir",
			pregunta: preguntaSinECLI,
			sesion: sesionDeJurisprudencia("La " + suCita + " y otra vez la " + suCita + ".\n" +
				"V\xc3\xa9anse ECLI:ES:TS:2022:1, " + sinDocumento + " y ecli:es:ts:2022:1."),
			cuentan: []ECLISinDocumento{
				{ECLI: sinDocumento, Condicion: CitaSinDocumentoCotejado},
				{ECLI: "ECLI:ES:TS:2022:1", Condicion: ECLISinOrigen},
				{ECLI: sinDocumento, Condicion: ECLISinOrigen},
			},
		},
		{
			nombre:   "sin-cita-ni-ecli",
			pregunta: preguntaSinECLI,
			sesion:   sesionDeJurisprudencia(lineaNoComprobada + "\n\nB\xc3\xbascala en " + direccionDelCendoj),
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.cuentan, CitaSinDocumento(caso.pregunta, caso.sesion))
		})
	}
}
