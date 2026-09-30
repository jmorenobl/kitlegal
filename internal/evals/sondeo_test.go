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

// casoDeLaTrazaIlegible es el caso de TestInforme cuya sesión el job no juzga
// porque no puede leer su traza; su transcript es el del caso aprobado.
const casoDeLaTrazaIlegible = "traza-ilegible"

// comandoDelBloqueA22 es lo que la copia con un comando ausente pone delante de
// las citas de la eval del art. 21: un segundo comando, que su sesión no ejecuta.
const comandoDelBloqueA22 = "  - applet: boe\n    norma: BOE-A-2015-10565\n    bloque: a22\n"

// Las dos primeras líneas de la salida del sondeo, siempre las mismas
// (contracts/sondeo.md §4 de H7.3; FR-065), escritas a mano: una salida con
// otras no pasa.
const (
	primeraLineaDelSondeoEsperada = "Esto es un sondeo, no un veredicto: el veredicto de la skill lo da el job de evals."
	segundaLineaDelSondeoEsperada = "No comprueba lo que el job lee de la traza de strace \xe2\x80\x94qu\xc3\xa9 " +
		"\xc3\xb3rdenes se ejecutaron (los comandos esperados y los prohibidos de cada eval) y las llegadas a la " +
		"red\xe2\x80\x94 ni la ausencia de Python: por eso sus tasas no son las del job."
)

// tituloSinTerminarEsperado es el título del apartado de las sesiones que
// quedaron sin terminar por otra causa (contracts/sondeo.md §4 de H7.3), en dos
// literales partidos dentro de un verbo: misspell toma el pretérito de
// «terminar», suelto, por una errata inglesa.
const tituloSinTerminarEsperado = "Sesiones que no termin" + "aron por otra causa"

// casoDelJuicioDelSondeo es un caso de TestJuicioDelSondeo: las entradas del
// informe del job, con las sesiones que juzgan el job y el sondeo, y, si la hay,
// la sesión que el job no pasa por lo que saca de la traza y el sondeo sí.
type casoDelJuicioDelSondeo struct {
	nombre             string
	entradas           func(t *testing.T) InformeAEscribir
	soloPasaEnElSondeo string
}

// TestJuicioDelSondeo fija el juicio del sondeo (contracts/sondeo.md §3.5 y §7 de
// H7.3; research D17; FR-061, FR-096; US4-1): sobre las mismas sesiones
// sintéticas, con su transcript y su traza, el sondeo da, eval a eval, el juicio
// del informe que escribe EscribirInforme, salvo lo que el job saca de la traza
// —los comandos ejecutados y ausentes, los prohibidos ejecutados, las
// invocaciones, lo fuera de lo grabado, las otras fallidas y las llegadas a la
// red— y los motivos y el pasa que dependen de ellos; el mismo recuento de las
// expresiones prohibidas; y cada serie del plan con las mismas sesiones y las
// mismas sin medir, y no menos que pasen. Las sesiones son las de cada caso de
// TestInforme, con un solo modelo, y las de copias en t.TempDir() con un comando
// ausente, con un prohibido ejecutado, con expresiones prohibidas y la prueba de
// red, y sin medir y sin abrir. La del caso de la traza ilegible, que el job no
// juzga, el sondeo la juzga como el job la del caso aprobado, que tiene su mismo
// transcript: el sondeo no lee la traza.
func TestJuicioDelSondeo(t *testing.T) {
	t.Parallel()

	for _, caso := range casosDelJuicioDelSondeo(t) {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			exigirElJuicioDelJobSinLaTraza(t, caso.entradas(t), caso.soloPasaEnElSondeo)
		})
	}

	t.Run(casoDeLaTrazaIlegible, func(t *testing.T) {
		t.Parallel()

		delJob, err := EscribirInforme(entradasDelCaso(casoAprobado, t.TempDir()))
		require.NoError(t, err)

		juicio := juicioDelSondeoDe(t, entradasDelCaso(casoDeLaTrazaIlegible, t.TempDir()))

		require.Len(t, juicio.resultados, 1)
		assert.Equal(t, juicioSinLaTraza(resultadoDeLaSesion(t, delJob, sesionDelArticulo21)), juicio.resultados[0])
	})
}

// casosDelJuicioDelSondeo son los casos de TestJuicioDelSondeo: cada caso de
// TestInforme menos el de la traza ilegible, con los ajustes que les da
// TestInforme salvo los modelos informativos, que el sondeo no tiene, y las
// copias del caso aprobado.
func casosDelJuicioDelSondeo(t *testing.T) []casoDelJuicioDelSondeo {
	t.Helper()

	ajustes := map[string]func(entradas *InformeAEscribir){
		"umbral-alcanzado":         func(entradas *InformeAEscribir) { entradas.Repeticiones, entradas.Umbral = 3, 2 },
		"umbral-no-alcanzado":      func(entradas *InformeAEscribir) { entradas.Repeticiones, entradas.Umbral = 3, 2 },
		"faltan-sesiones":          func(entradas *InformeAEscribir) { entradas.Repeticiones = 2 },
		"otro-modelo-en-la-sesion": func(entradas *InformeAEscribir) { entradas.ModeloQueDecide = modeloInformativoDelCaso },
	}

	var casos []casoDelJuicioDelSondeo

	for _, nombre := range ejecucionesDeInforme(t) {
		if nombre == casoDeLaTrazaIlegible {
			continue
		}

		casos = append(casos, casoDelJuicioDelSondeo{nombre: nombre, entradas: func(t *testing.T) InformeAEscribir {
			t.Helper()

			entradas := entradasDelCaso(nombre, t.TempDir())
			if ajustar := ajustes[nombre]; ajustar != nil {
				ajustar(&entradas)
			}

			return entradas
		}})
	}

	return append(casos,
		casoDelJuicioDelSondeo{
			nombre: "con-un-comando-ausente", entradas: entradasConUnComandoAusente, soloPasaEnElSondeo: sesionDelArticulo21,
		},
		casoDelJuicioDelSondeo{
			nombre: "con-un-prohibido-ejecutado", entradas: entradasConUnProhibido, soloPasaEnElSondeo: sesionDelArticulo21,
		},
		casoDelJuicioDelSondeo{nombre: "con-expresiones-y-la-prueba-de-red", entradas: entradasConExpresiones},
		casoDelJuicioDelSondeo{nombre: "sin-medir-y-sin-abrir", entradas: entradasSinMedirYSinAbrir},
	)
}

// exigirElJuicioDelJobSinLaTraza escribe el informe del job con las entradas y
// juzga con el sondeo las mismas sesiones, con el mismo modelo y las mismas
// repeticiones, y exige que el sondeo dé, eval a eval, el juicio del job sin lo
// que sale de la traza (juicioSinLaTraza), su mismo recuento de las expresiones
// prohibidas y, en cada serie del plan, sus mismas sesiones y sin medir y no
// menos que pasen; y, si se da, que esa sesión no pase en el job y sí en el
// sondeo, con más que pasan en su serie.
func exigirElJuicioDelJobSinLaTraza(t *testing.T, entradas InformeAEscribir, soloPasaEnElSondeo string) {
	t.Helper()

	informe, err := EscribirInforme(entradas)
	require.NoError(t, err)

	juicio := juicioDelSondeoDe(t, entradas)

	require.Len(t, juicio.resultados, len(informe.Evals), "el sondeo juzga las mismas sesiones que el job")

	for posicion, delJob := range informe.Evals {
		assert.Equal(t, juicioSinLaTraza(delJob), juicio.resultados[posicion],
			"la sesión %s es la del job sin lo que sale de la traza", delJob.Sesion)
	}

	assert.Equal(t, informe.ExpresionesProhibidasPorModelo, juicio.recuento)

	var planificadas []TasaDelInforme

	for _, tasa := range informe.Tasas {
		if tasa.Planificada {
			planificadas = append(planificadas, tasa)
		}
	}

	require.Len(t, juicio.series, len(planificadas), "el sondeo da las series del plan")

	for posicion, tasa := range planificadas {
		serie := juicio.series[posicion]

		assert.Equal(t, serieDelSondeo{
			eval: tasa.Eval, modelo: tasa.Modelo, informativa: !tasa.Decide,
			sesiones: tasa.Sesiones, pasan: serie.pasan, sinMedir: tasa.SinMedir,
		}, serie)
		assert.GreaterOrEqual(t, serie.pasan, tasa.Pasan, "la tasa de %s no es menor que la del job", tasa.Eval)
	}

	if soloPasaEnElSondeo != "" {
		exigirQueSoloPasaEnElSondeo(t, informe, juicio, soloPasaEnElSondeo)
	}
}

// exigirQueSoloPasaEnElSondeo exige que la sesión no pase en el job y sí en el
// sondeo, y que en la serie de su eval pasen más en el sondeo que en el job: el
// caso no pasa en vacío.
func exigirQueSoloPasaEnElSondeo(t *testing.T, informe Informe, juicio juicioDelSondeo, sesion string) {
	t.Helper()

	delJob := resultadoDeLaSesion(t, informe, sesion)
	assert.False(t, delJob.Pasa, "el job no pasa la sesión %s", sesion)

	posicion := slices.IndexFunc(juicio.resultados, func(resultado ResultadoDeEval) bool { return resultado.Sesion == sesion })
	require.GreaterOrEqual(t, posicion, 0, "el sondeo juzga la sesión %s", sesion)
	assert.True(t, juicio.resultados[posicion].Pasa, "el sondeo pasa la sesión %s", sesion)

	serie := slices.IndexFunc(juicio.series, func(serie serieDelSondeo) bool { return serie.eval == delJob.Eval })
	require.GreaterOrEqual(t, serie, 0, "el sondeo da la serie de %s", delJob.Eval)
	assert.Greater(t, juicio.series[serie].pasan, tasaDeLaSerie(t, informe, delJob.Eval, delJob.Modelo).Pasan)
}

// juicioSinLaTraza es el juicio del job de una sesión sin lo que sale de la
// traza (FR-061; contracts/sondeo.md §7 de H7.3): sin los comandos ejecutados y
// ausentes, los prohibidos ejecutados, las invocaciones, lo fuera de lo grabado,
// las otras fallidas ni las llegadas a la red; sin los motivos de un comando
// ausente o de un prohibido ejecutado; y, sin ellos, pasa si no le queda ningún
// motivo, como toda sesión que se juzga.
func juicioSinLaTraza(resultado ResultadoDeEval) ResultadoDeEval {
	resultado.ComandosEjecutados, resultado.ComandosAusentes, resultado.ComandosProhibidosEjecutados = nil, nil, nil
	resultado.Invocaciones, resultado.FueraDeLoGrabado = nil, nil
	resultado.OtrasFallidas, resultado.LlegadasALaRed = nil, nil

	var motivos []string

	for _, motivo := range resultado.Motivos {
		if !strings.HasPrefix(motivo, "comando ausente: ") && !strings.HasPrefix(motivo, "comando prohibido ejecutado: ") {
			motivos = append(motivos, motivo)
		}
	}

	resultado.Motivos = motivos
	resultado.Pasa = len(motivos) == 0

	return resultado
}

// juicioDelSondeoDe juzga con el sondeo las sesiones de las entradas del informe
// del job: las evals bien formadas de su directorio, con su lista, su modelo que
// decide, sus repeticiones y sus sesiones sin abrir.
func juicioDelSondeoDe(t *testing.T, entradas InformeAEscribir) juicioDelSondeo {
	t.Helper()

	conjunto, err := LeerConjunto(entradas.Evals)
	require.NoError(t, err)

	juicio, err := juzgarElSondeo(SondeoAJuzgar{
		Skill:        entradas.Skill,
		Evals:        entradas.Evals,
		Pedidas:      conjunto.Evals,
		Prohibidas:   conjunto.Prohibidas,
		Sesiones:     entradas.Sesiones,
		Modelo:       entradas.ModeloQueDecide,
		Repeticiones: entradas.Repeticiones,
		SinAbrir:     entradas.SinAbrir,
	})
	require.NoError(t, err)

	return juicio
}

// entradasDeLaCopia son las entradas del informe del job del caso aprobado con
// las evals y las sesiones de la copia.
func entradasDeLaCopia(t *testing.T, copia string) InformeAEscribir {
	t.Helper()

	entradas := entradasDelCaso(casoAprobado, t.TempDir())
	entradas.Evals = filepath.Join(copia, "evals")
	entradas.Sesiones = filepath.Join(copia, "sesiones")

	return entradas
}

// entradasConUnComandoAusente son las entradas del informe del job de una copia
// del caso aprobado de TestInforme en la que la eval del art. 21 espera además la
// lectura del bloque a22, que su sesión no hace.
func entradasConUnComandoAusente(t *testing.T) InformeAEscribir {
	t.Helper()

	copia := t.TempDir()
	require.NoError(t, os.CopyFS(copia, os.DirFS(filepath.Join(casosDeInforme, casoAprobado))))

	evals := filepath.Join(copia, "evals")
	eval := contenidoDeLaSesion(t, evals, ficheroDeLaEval01)
	require.Equal(t, 1, strings.Count(eval, "citas:"), "la eval del art. 21 lleva sus citas detrás de sus comandos")
	escribirEnLaCopia(t, evals, ficheroDeLaEval01, strings.Replace(eval, "citas:", comandoDelBloqueA22+"citas:", 1))

	return entradasDeLaCopia(t, copia)
}

// entradasConUnProhibido son las entradas del informe del job de la copia del
// caso aprobado de TestInformeConProhibidos, en la que la sesión del art. 21
// ejecuta el graph show que su eval prohíbe.
func entradasConUnProhibido(t *testing.T) InformeAEscribir {
	t.Helper()

	return entradasDeLaCopia(t, copiaDelCasoAprobadoConProhibido(t))
}

// entradasConExpresiones son las entradas del informe del job de la copia del
// caso aprobado de TestInformeConExpresionesProhibidas con la lista del
// repositorio y la sesión de la prueba de red: respuestas con expresiones
// prohibidas, sesiones de otro modelo y una serie con la pregunta ampliada.
func entradasConExpresiones(t *testing.T) InformeAEscribir {
	t.Helper()

	return entradasDeLaCopia(t, copiaConExpresiones(t, true, true))
}

// entradasSinMedirYSinAbrir son las entradas del informe del job de una copia del
// caso aprobado con la lista del repositorio en la que la sesión del art. 21
// termina con el mensaje del límite de uso y la de no activación no se abrió tras
// él.
func entradasSinMedirYSinAbrir(t *testing.T) InformeAEscribir {
	t.Helper()

	copia := copiaDelCasoAprobadoConLaLista(t)
	sesiones := filepath.Join(copia, "sesiones")

	cambiarElFinalDeLaSesion(t, filepath.Join(sesiones, sesionDelArticulo21), 1,
		mensajeResultConError(t, textoDelLimiteDeSesion))
	require.NoError(t, os.RemoveAll(filepath.Join(sesiones, sesionDeNoActivacion)))

	entradas := entradasDeLaCopia(t, copia)
	entradas.SinAbrir = []SesionPlanificada{
		{Nombre: sesionDeNoActivacion, Fichero: ficheroDeNoActivacion, Modelo: modeloQueDecide},
	}

	return entradas
}

// TestSalidaDelSondeo fija la salida del sondeo (contracts/sondeo.md §4 y §7 de
// H7.3; FR-063, FR-065; US4-2) sobre sondeos sintéticos escritos en
// t.TempDir(): las dos primeras líneas; ninguna expresión prohibida de ninguna
// sesión; la tasa de cada serie, con las informativas marcadas, la que tiene
// alguna sesión sin medir sin tasa y la sesión que no terminó por otra causa
// contada como no pasada; el recuento con el 5 % de referencia; las sesiones sin
// medir por límite de uso, también la que no se abrió, y las que quedaron sin
// terminar por otra causa, cada una con su motivo del job, entre ellas la de un
// transcript que acaba en el result con is_error de una credencial que no sirve
// y sale con código 1 (research V18); y, con la skill sin lista y todas las
// sesiones terminadas, la línea de la skill sin lista y «ninguna.» en los dos
// apartados.
func TestSalidaDelSondeo(t *testing.T) {
	t.Parallel()

	t.Run("con-lista", func(t *testing.T) {
		t.Parallel()

		exigirLaSalidaConLista(t)
	})

	t.Run("sin-lista", func(t *testing.T) {
		t.Parallel()

		exigirLaSalidaSinLista(t)
	})
}

// exigirLaSalidaConLista arma un sondeo de dos evals que deciden y tres
// informativas con la lista del repositorio, como el ejemplo de
// contracts/sondeo.md §4: la primera sesión de la segunda eval lleva delante de
// su respuesta la transición de la memoria de consultas; la primera de la
// cuarta acaba con la credencial que no sirve y código 1; la segunda de la
// quinta, con el mensaje del límite de uso; y la tercera de la quinta no se
// abrió. Exige su salida entera.
func exigirLaSalidaConLista(t *testing.T) {
	t.Helper()

	copia := escribirSondeoSintetico(t, 2, 3, true)

	anteponerALaRespuesta(t, sesionDelSondeo(copia, 2, 1), transicionDeLaMemoria+"\n\n")
	cambiarElFinalDeLaSesion(t, sesionDelSondeo(copia, 4, 1), 1, mensajeResultConError(t, textoDeLaCredencial))
	cambiarElFinalDeLaSesion(t, sesionDelSondeo(copia, 5, 2), 1, mensajeResultConError(t, textoDelLimiteDeSesion))
	require.NoError(t, os.RemoveAll(sesionDelSondeo(copia, 5, 3)))

	juicio := juicioDelSondeoDe(t, entradasDelSondeo(copia, []SesionPlanificada{
		{Nombre: sesionSintetica(5, modeloSonnet5, 3), Fichero: ficheroSintetico(5), Modelo: modeloSonnet5},
	}))
	salida := juicio.salida()

	exigirLasDosPrimerasLineas(t, salida)
	exigirSinExpresiones(t, juicio, salida)
	assert.Contains(t, salida, "- "+sesionSintetica(4, modeloSonnet5, 1)+
		": la sesi\xc3\xb3n no termin\xc3\xb3: c\xc3\xb3digo 1: result con is_error: Failed to authenticate. ")

	assert.Equal(t, lineasDeLaSalida(
		primeraLineaDelSondeoEsperada,
		segundaLineaDelSondeoEsperada,
		"",
		"Tasa de cada serie (sesiones que pasan de las medidas):",
		"- 01-sintetica.yaml con claude-sonnet-5: 3 de 3",
		"- 02-sintetica.yaml con claude-sonnet-5: 2 de 3",
		"- 03-sintetica.yaml con claude-sonnet-5 (informativa): 3 de 3",
		"- 04-sintetica.yaml con claude-sonnet-5 (informativa): 2 de 3",
		"- 05-sintetica.yaml con claude-sonnet-5 (informativa): sin medir",
		"",
		"Respuestas con alguna expresi\xc3\xb3n prohibida en las evals que activan la skill: 1 de 13 (7,7 %); "+
			"referencia: como mucho el 5 %.",
		"",
		"Sesiones sin medir por l\xc3\xadmite de uso:",
		"- 05-sintetica-claude-sonnet-5-02: "+sinMedirPorElMensaje,
		"- 05-sintetica-claude-sonnet-5-03: "+sinAbrirTrasElLimite,
		"",
		tituloSinTerminarEsperado+":",
		"- 04-sintetica-claude-sonnet-5-01: la sesi\xc3\xb3n no termin\xc3\xb3: c\xc3\xb3digo 1: result con is_error: "+
			textoDeLaCredencial,
	), salida)
}

// exigirLaSalidaSinLista arma un sondeo de una eval que decide y una
// informativa, sin lista y con todas sus sesiones terminadas y pasando, y exige
// su salida entera.
func exigirLaSalidaSinLista(t *testing.T) {
	t.Helper()

	juicio := juicioDelSondeoDe(t, entradasDelSondeo(escribirSondeoSintetico(t, 1, 1, false), nil))
	salida := juicio.salida()

	exigirLasDosPrimerasLineas(t, salida)
	assert.Nil(t, juicio.recuento, "sin lista no hay recuento: uno de cero diría que se buscó")

	assert.Equal(t, lineasDeLaSalida(
		primeraLineaDelSondeoEsperada,
		segundaLineaDelSondeoEsperada,
		"",
		"Tasa de cada serie (sesiones que pasan de las medidas):",
		"- 01-sintetica.yaml con claude-sonnet-5: 3 de 3",
		"- 02-sintetica.yaml con claude-sonnet-5 (informativa): 3 de 3",
		"",
		"Respuestas con alguna expresi\xc3\xb3n prohibida: "+parrafoDeLaSkillSinLista+".",
		"",
		"Sesiones sin medir por l\xc3\xadmite de uso: ninguna.",
		"",
		tituloSinTerminarEsperado+": ninguna.",
	), salida)
}

// exigirLasDosPrimerasLineas exige que la salida empiece por las dos líneas
// fijas: que no es un veredicto y lo que no comprueba.
func exigirLasDosPrimerasLineas(t *testing.T, salida string) {
	t.Helper()

	lineas := strings.SplitN(salida, "\n", 3)
	require.Len(t, lineas, 3)
	assert.Equal(t, primeraLineaDelSondeoEsperada, lineas[0])
	assert.Equal(t, segundaLineaDelSondeoEsperada, lineas[1])
}

// exigirSinExpresiones exige que ninguna expresión prohibida que el sondeo
// encontró en alguna sesión aparezca en su salida, en ninguna forma de
// mayúsculas: la salida es solo lo agregado (FR-065).
func exigirSinExpresiones(t *testing.T, juicio juicioDelSondeo, salida string) {
	t.Helper()

	var expresiones []string
	for _, resultado := range juicio.resultados {
		expresiones = append(expresiones, resultado.ExpresionesProhibidas...)
	}

	require.NotEmpty(t, expresiones, "alguna sesión lleva expresiones prohibidas: sin ellas, no se comprobaría nada")

	for _, expresion := range expresiones {
		assert.NotContains(t, strings.ToLower(salida), expresion)
	}
}

// escribirSondeoSintetico arma en un directorio temporal del test las evals y
// las sesiones de un sondeo con modeloSonnet5 y tres repeticiones: sus evals son
// copias de la del art. 21 del caso aprobado, las que deciden y después las
// informativas, con la lista del repositorio si conLista; y cada sesión, una
// copia de la del art. 21 del caso aprobado, que pasa, con su eval y su modelo.
// Devuelve su ruta.
func escribirSondeoSintetico(t *testing.T, queDeciden, informativas int, conLista bool) string {
	t.Helper()

	copia := t.TempDir()
	evals := filepath.Join(copia, "evals")
	require.NoError(t, os.Mkdir(evals, 0o750))

	if conLista {
		copiarLaListaDelRepositorio(t, evals)
	}

	eval := contenidoDeLaSesion(t, filepath.Join(casosDeInforme, casoAprobado, "evals"), ficheroDeLaEval01)

	for numero := 1; numero <= queDeciden+informativas; numero++ {
		contenido := eval
		if numero > queDeciden {
			contenido += "informativa: true\n"
		}

		escribirEnLaCopia(t, evals, ficheroSintetico(numero), contenido)

		for vez := 1; vez <= repeticionesConUmbrales; vez++ {
			escribirSesionSintetica(t, copia, numero, modeloSonnet5, vez, false)
		}
	}

	return copia
}

// entradasDelSondeo son las entradas de un sondeo sintético de
// escribirSondeoSintetico, con las sesiones que el repartidor no abrió, como las
// del informe del job de las que juicioDelSondeoDe toma lo que el sondeo necesita.
func entradasDelSondeo(copia string, sinAbrir []SesionPlanificada) InformeAEscribir {
	return InformeAEscribir{
		Skill:           skillDeLasSesiones,
		Evals:           filepath.Join(copia, "evals"),
		Sesiones:        filepath.Join(copia, "sesiones"),
		ModeloQueDecide: modeloSonnet5,
		Repeticiones:    repeticionesConUmbrales,
		SinAbrir:        sinAbrir,
	}
}

// sesionDelSondeo es el directorio de la sesión de esa repetición de la eval de
// ese número de un sondeo sintético.
func sesionDelSondeo(copia string, numero, vez int) string {
	return filepath.Join(copia, "sesiones", sesionSintetica(numero, modeloSonnet5, vez))
}

// lineasDeLaSalida son las líneas dadas, cada una terminada en un salto de línea.
func lineasDeLaSalida(lineas ...string) string {
	return strings.Join(lineas, "\n") + "\n"
}
