package evals

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
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

// valorDeLaSuscripcion es el de CLAUDE_CODE_OAUTH_TOKEN en los entornos de los
// tests del sondeo: basta con que no esté vacío, y ninguna sesión lo usa.
const valorDeLaSuscripcion = "de-la-base"

// Las skills del repositorio que los tests del sondeo piden: las dos de la
// matriz del job, con 4 y 1 sesiones a la vez (contracts/ejecucion-del-job.md
// §7 de H7.3).
const (
	skillQueSondea     = "boe-legislacion"
	otraSkillQueSondea = "legal-core"
)

// casoDeComprobarElSondeo es un caso de TestComprobarElSondeo: cómo cambia el
// sondeo que comprueba sin error y, o bien las líneas de su error, en su orden,
// o bien, sin error, lo comprobado: las evals pedidas, en su orden, y la
// concurrencia.
type casoDeComprobarElSondeo struct {
	nombre       string
	ajustar      func(t *testing.T, s *SondeoAEjecutar)
	errores      []string
	pedidas      []string
	concurrencia int
}

// TestComprobarElSondeo fija la comprobación de los argumentos y de la
// credencial del sondeo, antes de construir nada (contracts/sondeo.md §3.1,
// §3.2 y §7 de H7.3; research D16; FR-060, FR-063, FR-067; US4-5), sobre las
// evals del repositorio y la definición real del job, o sobre carpetas de evals
// escritas en t.TempDir(): cada error de §3.1, nombrando su argumento de make,
// con los de varios argumentos a la vez, uno por línea y en el orden de la
// orden; la concurrencia por omisión, la del job para la skill, 4 y 1; la
// pedida, si la hay; y las evals pedidas en el orden en que se piden. Sin
// CLAUDE_CODE_OAUTH_TOKEN, o con la variable vacía, el error de §3.2, que no
// llega si un argumento no vale.
func TestComprobarElSondeo(t *testing.T) {
	t.Parallel()

	for _, caso := range slices.Concat(casosDeLosArgumentos(), casosDeLasEvals(), casosDeLosEnteros(),
		casosDeLaSuscripcion()) {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			sondeo := sondeoAComprobar()
			caso.ajustar(t, &sondeo)

			comprobado, err := comprobarElSondeo(sondeo)
			if len(caso.errores) > 0 {
				require.EqualError(t, err, strings.Join(caso.errores, "\n"))

				return
			}

			require.NoError(t, err)
			exigirLoComprobado(t, sondeo, caso, comprobado)
		})
	}
}

// sondeoAComprobar es el sondeo que comprueba sin error: la skill del
// repositorio con sus evals 14 y 03, en ese orden, Sonnet 5, tres repeticiones,
// la concurrencia del job y la credencial en su entorno.
func sondeoAComprobar() SondeoAEjecutar {
	return SondeoAEjecutar{
		Argumentos: ArgumentosDelSondeo{
			Skill: skillQueSondea, Evals: "14,03", Modelo: modeloSonnet5, Repeticiones: "3",
		},
		Entorno:          []string{variableDeLaSuscripcion + "=" + valorDeLaSuscripcion},
		EvalsDeLasSkills: directorioDeEvals,
	}
}

// exigirLoComprobado exige que lo comprobado sea lo del sondeo sin error: su
// skill, la carpeta de sus evals, las evals pedidas del caso, en su orden, con la
// lista de la carpeta, su modelo, sus repeticiones y la concurrencia del caso.
func exigirLoComprobado(t *testing.T, sondeo SondeoAEjecutar, caso casoDeComprobarElSondeo, comprobado sondeoComprobado) {
	t.Helper()

	evals := filepath.Join(sondeo.EvalsDeLasSkills, sondeo.Argumentos.Skill)

	conjunto, err := LeerConjunto(evals)
	require.NoError(t, err)

	pedidas := make([]Eval, 0, len(caso.pedidas))
	for _, fichero := range caso.pedidas {
		posicion := slices.IndexFunc(conjunto.Evals, func(eval Eval) bool { return eval.Fichero == fichero })
		require.GreaterOrEqualf(t, posicion, 0, "la eval %s está en %s", fichero, evals)

		pedidas = append(pedidas, conjunto.Evals[posicion])
	}

	assert.Equal(t, sondeoComprobado{
		skill: sondeo.Argumentos.Skill, evals: evals, pedidas: pedidas, prohibidas: conjunto.Prohibidas,
		modelo: sondeo.Argumentos.Modelo, repeticiones: 3, concurrencia: caso.concurrencia,
	}, comprobado)
}

// casosDeLosArgumentos son los casos de TestComprobarElSondeo de la skill y del
// modelo, los de varios argumentos a la vez y los que comprueban sin error.
func casosDeLosArgumentos() []casoDeComprobarElSondeo {
	return []casoDeComprobarElSondeo{
		{
			nombre:       "concurrencia-del-job-en-" + skillQueSondea,
			ajustar:      func(*testing.T, *SondeoAEjecutar) {},
			pedidas:      []string{"14-trlrhl-impuestos-por-materia.yaml", "03-lrbrl-atribuciones-del-pleno.yaml"},
			concurrencia: 4,
		},
		{
			nombre: "concurrencia-del-job-en-" + otraSkillQueSondea,
			ajustar: func(_ *testing.T, s *SondeoAEjecutar) {
				s.Argumentos.Skill, s.Argumentos.Evals = otraSkillQueSondea, "03,01"
			},
			pedidas:      []string{"03-no-activa-receta-de-cocina.yaml", "01-territorio-municipio-cubierto.yaml"},
			concurrencia: 1,
		},
		{
			nombre:       "concurrencia-pedida",
			ajustar:      func(_ *testing.T, s *SondeoAEjecutar) { s.Argumentos.Concurrencia = "2" },
			pedidas:      []string{"14-trlrhl-impuestos-por-materia.yaml", "03-lrbrl-atribuciones-del-pleno.yaml"},
			concurrencia: 2,
		},
		{
			nombre:  "skill-sin-la-forma-de-un-nombre",
			ajustar: func(_ *testing.T, s *SondeoAEjecutar) { s.Argumentos.Skill = "Boe-Legislacion" },
			errores: []string{"SKILL: \xc2\xabBoe-Legislacion\xc2\xbb no es ninguna skill con evals"},
		},
		{
			nombre:  "skill-sin-carpeta-de-evals",
			ajustar: func(_ *testing.T, s *SondeoAEjecutar) { s.Argumentos.Skill = "no-existe" },
			errores: []string{"SKILL: \xc2\xabno-existe\xc2\xbb no es ninguna skill con evals"},
		},
		{
			nombre: "skill-con-un-fichero-mal-formado",
			ajustar: func(t *testing.T, s *SondeoAEjecutar) {
				t.Helper()

				s.EvalsDeLasSkills = carpetasDeEvals(t, skillQueSondea,
					entradaDeConjunto{nombre: "14-bien-formada.yaml", contenido: contenidoDelArticulo21},
					entradaDeConjunto{nombre: "03-sin-pregunta.yaml", contenido: contenidoSinPregunta})
			},
			errores: []string{"SKILL: \xc2\xabboe-legislacion\xc2\xbb no es ninguna skill con evals"},
		},
		{
			nombre: "skill-que-el-job-no-ejecuta",
			ajustar: func(t *testing.T, s *SondeoAEjecutar) {
				t.Helper()

				s.EvalsDeLasSkills = carpetasDeEvals(t, "otra-skill",
					entradaDeConjunto{nombre: "14-bien-formada.yaml", contenido: contenidoDelArticulo21},
					entradaDeConjunto{nombre: "03-bien-formada.yaml", contenido: contenidoDelArticulo21})
				s.Argumentos.Skill = "otra-skill"
			},
			errores: []string{"SKILL: el job de evals no ejecuta \xc2\xabotra-skill\xc2\xbb"},
		},
		{
			nombre:  "modelo-vacio",
			ajustar: func(_ *testing.T, s *SondeoAEjecutar) { s.Argumentos.Modelo = "" },
			errores: []string{"MODELO: est\xc3\xa1 vac\xc3\xado"},
		},
		{
			nombre:  "modelo-sin-la-forma-de-un-id",
			ajustar: func(_ *testing.T, s *SondeoAEjecutar) { s.Argumentos.Modelo = "Claude Sonnet 5" },
			errores: []string{"MODELO: \xc2\xabClaude Sonnet 5\xc2\xbb no tiene la forma de un id de modelo"},
		},
		{
			nombre: "varios-argumentos-a-la-vez",
			ajustar: func(_ *testing.T, s *SondeoAEjecutar) {
				s.Argumentos = ArgumentosDelSondeo{
					Skill: skillQueSondea, Evals: "3", Modelo: "", Repeticiones: "0", Concurrencia: "0",
				}
			},
			errores: []string{
				"EVALS: \xc2\xab3\xc2\xbb no es una lista de n\xc3\xbameros de eval de dos cifras separados por comas",
				"MODELO: est\xc3\xa1 vac\xc3\xado",
				"REPETICIONES: \xc2\xab0\xc2\xbb no es un entero mayor o igual que 1",
				"CONCURRENCIA: \xc2\xab0\xc2\xbb no es un entero mayor o igual que 1",
			},
		},
		{
			nombre: "skill-y-evals-que-no-valen",
			ajustar: func(_ *testing.T, s *SondeoAEjecutar) {
				s.Argumentos.Skill, s.Argumentos.Evals = "no-existe", "3"
			},
			errores: []string{
				"SKILL: \xc2\xabno-existe\xc2\xbb no es ninguna skill con evals",
				"EVALS: \xc2\xab3\xc2\xbb no es una lista de n\xc3\xbameros de eval de dos cifras separados por comas",
			},
		},
	}
}

// casosDeLasEvals son los casos de TestComprobarElSondeo del argumento EVALS:
// las listas que no tienen su forma —también la que repite un número— y los
// números que no son de ninguna eval de la skill, uno por línea. Con una skill
// que no vale, solo se comprueba la forma: sin sus evals, sus números no se
// pueden buscar.
func casosDeLasEvals() []casoDeComprobarElSondeo {
	sinLaForma := map[string]string{
		"una-cifra":         "3",
		"tres-cifras":       "003",
		"coma-al-final":     "03,",
		"con-un-espacio":    "03, 14",
		"vacia":             "",
		"otro-separador":    "03;14",
		"numero-repetido":   "03,14,03",
		"nombre-de-la-eval": "03-lrbrl-atribuciones-del-pleno.yaml",
	}

	casos := make([]casoDeComprobarElSondeo, 0, len(sinLaForma)+2)

	for nombre, valor := range sinLaForma {
		casos = append(casos, casoDeComprobarElSondeo{
			nombre:  "evals-" + nombre,
			ajustar: func(_ *testing.T, s *SondeoAEjecutar) { s.Argumentos.Evals = valor },
			errores: []string{"EVALS: \xc2\xab" + valor + "\xc2\xbb no es una lista de n\xc3\xbameros de eval de dos cifras " +
				"separados por comas"},
		})
	}

	return append(casos,
		casoDeComprobarElSondeo{
			nombre:  "evals-que-no-son-de-la-skill",
			ajustar: func(_ *testing.T, s *SondeoAEjecutar) { s.Argumentos.Evals = "03,98,14,99" },
			errores: []string{
				"EVALS: 98 no es ninguna eval de " + skillQueSondea,
				"EVALS: 99 no es ninguna eval de " + skillQueSondea,
			},
		},
		casoDeComprobarElSondeo{
			nombre: "evals-de-otra-skill",
			ajustar: func(_ *testing.T, s *SondeoAEjecutar) {
				s.Argumentos.Skill, s.Argumentos.Evals = otraSkillQueSondea, "01,19"
			},
			errores: []string{"EVALS: 19 no es ninguna eval de " + otraSkillQueSondea},
		},
	)
}

// casosDeLosEnteros son los casos de TestComprobarElSondeo de REPETICIONES y
// CONCURRENCIA con un valor que no es un entero mayor o igual que 1; la
// concurrencia vacía no es un error, sino la del job.
func casosDeLosEnteros() []casoDeComprobarElSondeo {
	var casos []casoDeComprobarElSondeo

	for _, valor := range []string{"0", "-1", "tres", "2.5", "1e3", "99999999999999999999"} {
		casos = append(casos,
			casoDeComprobarElSondeo{
				nombre:  "repeticiones-" + valor,
				ajustar: func(_ *testing.T, s *SondeoAEjecutar) { s.Argumentos.Repeticiones = valor },
				errores: []string{"REPETICIONES: \xc2\xab" + valor + "\xc2\xbb no es un entero mayor o igual que 1"},
			},
			casoDeComprobarElSondeo{
				nombre:  "concurrencia-" + valor,
				ajustar: func(_ *testing.T, s *SondeoAEjecutar) { s.Argumentos.Concurrencia = valor },
				errores: []string{"CONCURRENCIA: \xc2\xab" + valor + "\xc2\xbb no es un entero mayor o igual que 1"},
			})
	}

	return append(casos, casoDeComprobarElSondeo{
		nombre:  "repeticiones-vacias",
		ajustar: func(_ *testing.T, s *SondeoAEjecutar) { s.Argumentos.Repeticiones = "" },
		errores: []string{"REPETICIONES: \xc2\xab\xc2\xbb no es un entero mayor o igual que 1"},
	})
}

// casosDeLaSuscripcion son los casos de TestComprobarElSondeo de la credencial:
// sin CLAUDE_CODE_OAUTH_TOKEN o con la variable vacía, el error de
// contracts/sondeo.md §3.2, que nombra la variable; y, si además un argumento no
// vale, solo el del argumento, que se comprueba antes.
func casosDeLaSuscripcion() []casoDeComprobarElSondeo {
	faltaLaSuscripcion := []string{"falta la credencial: CLAUDE_CODE_OAUTH_TOKEN, el token de la suscripci\xc3\xb3n que da " +
		"claude setup-token, no est\xc3\xa1 en el entorno o est\xc3\xa1 vac\xc3\xada"}

	return []casoDeComprobarElSondeo{
		{
			nombre:  "sin-la-variable",
			ajustar: func(_ *testing.T, s *SondeoAEjecutar) { s.Entorno = []string{"OTRA_DE_LA_BASE=si"} },
			errores: faltaLaSuscripcion,
		},
		{
			nombre:  "con-la-variable-vacia",
			ajustar: func(_ *testing.T, s *SondeoAEjecutar) { s.Entorno = []string{variableDeLaSuscripcion + "="} },
			errores: faltaLaSuscripcion,
		},
		{
			nombre: "sin-la-variable-y-con-un-argumento-que-no-vale",
			ajustar: func(_ *testing.T, s *SondeoAEjecutar) {
				s.Entorno, s.Argumentos.Modelo = nil, ""
			},
			errores: []string{"MODELO: est\xc3\xa1 vac\xc3\xado"},
		},
	}
}

// carpetasDeEvals crea en un directorio temporal del test la carpeta de evals de
// la skill, con las entradas dadas, y devuelve el directorio: el de las evals de
// las skills de un sondeo.
func carpetasDeEvals(t *testing.T, skill string, entradas ...entradaDeConjunto) string {
	t.Helper()

	raiz := t.TempDir()
	require.NoError(t, os.CopyFS(filepath.Join(raiz, skill), os.DirFS(crearConjunto(t, entradas))))

	return raiz
}

// casoDeSondear es un caso de TestSondear: cómo cambia el sondeo del test —sus
// repeticiones y su concurrencia son las del caso—, el código con el que sale
// el sustituto de claude en todas las sesiones, el transcript y la espera de
// cada una, por su posición en el plan, y, o bien su error, o bien las sesiones
// del plan que se abren, desde la primera, y su salida.
type casoDeSondear struct {
	nombre       string
	repeticiones string
	concurrencia string
	ajustar      func(s *SondeoAEjecutar)
	codigo       int
	transcripts  func(t *testing.T) map[int]string
	esperas      map[int]int
	error        string
	abiertas     int
	salida       func(plan []SesionPlanificada) string
}

// TestSondear fija el sondeo desde sus argumentos hasta su salida
// (contracts/sondeo.md §3, §5, §6 y §7 de H7.3; research D10 y D16; FR-061 a
// FR-066; SC-009; US4-3, US4-4, US4-6), con los sustitutos de claude, una eval
// sintética del art. 21 de la LPAC en la carpeta de la skill del repositorio que
// el job ejecuta y un árbol que no construye nada: un kitlegal que no hace nada
// en el bin/ del temporal y una skill en el directorio de Claude Code de su
// HOME. Si todas las sesiones terminan sin terminar por otra causa, o si la
// primera da el mensaje del límite de uso mientras la segunda sigue abierta,
// devuelve su salida sin error: en el segundo caso no abre ninguna más, la
// segunda termina y se juzga y las que faltan salen sin medir. Con un argumento
// o una credencial que no valen, devuelve el error sin preparar el árbol, sin
// abrir ninguna sesión y sin escribir nada en el temporal. Las sesiones no ven
// ANTHROPIC_API_KEY, ANTHROPIC_AUTH_TOKEN ni ninguna otra variable de la base
// que no sea de contracts/sondeo.md §5, aunque estén en ella; ven
// CLAUDE_CODE_OAUTH_TOKEN y las demás de §5 con su valor de la base, el PATH de
// la base con el bin/ del temporal delante —el primer kitlegal que resuelven es
// el de ahí— y el HOME del temporal. Las sesiones no escriben en el HOME ni en
// el TMPDIR de la base, que quedan vacíos, y el temporal lleva solo bin/, home/
// y sesiones/, sin informe ni veredicto. Lo que el proceso escribe en su propio
// TMPDIR —la preparación de cada sesión— y lo que go install escribe en el de la
// base no lo mira: ese TMPDIR lo pone el guion dentro de su temporal
// (TestGuionDelSondeo).
func TestSondear(t *testing.T) {
	t.Parallel()

	for _, caso := range casosDeSondear() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			sondeo := nuevoSondeoDelTest(t, caso)

			salida, err := sondear(t.Context().Done(), sondeo.ejecutar)

			sondeo.exigirLaBaseSinTocar(t)

			if caso.error != "" {
				require.EqualError(t, err, caso.error)
				sondeo.exigirQueNoSondea(t)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, caso.salida(sondeo.plan), salida)
			sondeo.exigirLasSesiones(t, caso.abiertas)
		})
	}
}

// casosDeSondear son los casos de TestSondear.
func casosDeSondear() []casoDeSondear {
	faltaLaSuscripcion := "falta la credencial: CLAUDE_CODE_OAUTH_TOKEN, el token de la suscripci\xc3\xb3n que da " +
		"claude setup-token, no est\xc3\xa1 en el entorno o est\xc3\xa1 vac\xc3\xada"

	return []casoDeSondear{
		{
			nombre: "todas-las-sesiones-fallan", repeticiones: "2", codigo: 1,
			transcripts: func(t *testing.T) map[int]string {
				t.Helper()

				credencial := mensajeInit + mensajeResultConError(t, textoDeLaCredencial)

				return map[int]string{0: credencial, 1: credencial}
			},
			abiertas: 2,
			salida:   salidaConTodasLasSesionesFallidas,
		},
		{
			nombre: "limite-de-uso-en-la-primera", repeticiones: "4", concurrencia: "2",
			transcripts: func(t *testing.T) map[int]string {
				t.Helper()

				return map[int]string{
					0: mensajeInit + mensajeResultConError(t, textoDelLimiteDeSesion),
					1: transcriptTerminado, 2: transcriptTerminado, 3: transcriptTerminado,
				}
			},
			esperas:  map[int]int{1: esperaDeLaAbierta},
			abiertas: 2,
			salida:   salidaTrasElLimiteDeUso,
		},
		{
			nombre: "un-argumento-que-no-vale", repeticiones: "0",
			error: "REPETICIONES: \xc2\xab0\xc2\xbb no es un entero mayor o igual que 1",
		},
		{
			nombre: "sin-la-credencial", repeticiones: "1",
			ajustar: func(s *SondeoAEjecutar) { s.Entorno = sobreLaBase(s.Entorno, nil, variableDeLaSuscripcion) },
			error:   faltaLaSuscripcion,
		},
		{
			nombre: "con-la-credencial-vacia", repeticiones: "1",
			ajustar: func(s *SondeoAEjecutar) {
				s.Entorno = sobreLaBase(s.Entorno, []string{variableDeLaSuscripcion + "="})
			},
			error: faltaLaSuscripcion,
		},
	}
}

// esperaDeLaAbierta son los segundos que duerme en TestSondear la sesión que
// sigue abierta cuando la primera da el mensaje del límite de uso, que no
// duerme: con ellos de margen, la primera termina antes aunque la máquina vaya
// cargada.
const esperaDeLaAbierta = 3

// salidaConTodasLasSesionesFallidas es la salida del sondeo de TestSondear en el
// que todas las sesiones acaban con la credencial que no sirve y código 1: su
// serie no pasa ninguna y cada una sale entre las que quedaron sin terminar, con
// su motivo del job.
func salidaConTodasLasSesionesFallidas(plan []SesionPlanificada) string {
	lineas := []string{
		primeraLineaDelSondeoEsperada, segundaLineaDelSondeoEsperada, "",
		"Tasa de cada serie (sesiones que pasan de las medidas):",
		"- " + nombreDeEval + " con " + modeloDeLaSesion + ": 0 de 2",
		"",
		"Respuestas con alguna expresi\xc3\xb3n prohibida: " + parrafoDeLaSkillSinLista + ".",
		"",
		"Sesiones sin medir por l\xc3\xadmite de uso: ninguna.",
		"",
		tituloSinTerminarEsperado + ":",
	}

	for _, sesion := range plan {
		lineas = append(lineas, "- "+sesion.Nombre+": la sesi\xc3\xb3n no termin\xc3\xb3: c\xc3\xb3digo 1: "+
			"result con is_error: "+textoDeLaCredencial)
	}

	return lineasDeLaSalida(lineas...)
}

// salidaTrasElLimiteDeUso es la salida del sondeo de TestSondear en el que la
// primera sesión da el mensaje del límite de uso: su serie queda sin medir, y
// salen sin medir la primera, con el mensaje, y las dos que no se abrieron.
func salidaTrasElLimiteDeUso(plan []SesionPlanificada) string {
	return lineasDeLaSalida(
		primeraLineaDelSondeoEsperada, segundaLineaDelSondeoEsperada, "",
		"Tasa de cada serie (sesiones que pasan de las medidas):",
		"- "+nombreDeEval+" con "+modeloDeLaSesion+": sin medir",
		"",
		"Respuestas con alguna expresi\xc3\xb3n prohibida: "+parrafoDeLaSkillSinLista+".",
		"",
		"Sesiones sin medir por l\xc3\xadmite de uso:",
		"- "+plan[0].Nombre+": "+sinMedirPorElMensaje,
		"- "+plan[2].Nombre+": "+sinAbrirTrasElLimite,
		"- "+plan[3].Nombre+": "+sinAbrirTrasElLimite,
		"",
		tituloSinTerminarEsperado+": ninguna.",
	)
}

// sondeoDelTest es un sondeo de TestSondear con lo que el test mira después:
// sus sustitutos, su plan, el PATH de su base y las variables de §5 que se le
// dan, el HOME y el TMPDIR de su base, y si se preparó su árbol.
type sondeoDelTest struct {
	ejecutar   SondeoAEjecutar
	sustitutos sustitutos
	plan       []SesionPlanificada

	rutaDeLaBase string
	queVen       []string

	personalDeLaBase string
	temporalDeLaBase string

	arbolPreparado *bool
}

// nuevoSondeoDelTest arma el sondeo de un caso de TestSondear: los sustitutos,
// con el claude del sondeo delante en el PATH de la base y el código del caso;
// la eval sintética del art. 21 en la carpeta de la skill; los argumentos del
// caso con modeloDeLaSesion; el guion de la sesión; el árbol que no construye
// nada; y la base: la del proceso sin sus credenciales de Claude Code, con un
// HOME y un TMPDIR vacíos, las variables de §5 con valores propios del test,
// CLAUDE_CODE_OAUTH_TOKEN y las que las sesiones no tienen que ver. Deja los
// transcripts y las esperas del caso, y el ajuste del caso, si lo hay.
func nuevoSondeoDelTest(t *testing.T, caso casoDeSondear) sondeoDelTest {
	t.Helper()

	s := escribirSustitutos(t)
	claude := s.escribirElClaudeDelSondeo(t, variableDeEspera+"=0", variableDeCodigo+"="+strconv.Itoa(caso.codigo))

	guion, err := filepath.Abs(guionDeLaSesion)
	require.NoError(t, err)

	sondeo := sondeoDelTest{
		sustitutos:       s,
		rutaDeLaBase:     claude + string(os.PathListSeparator) + os.Getenv("PATH"),
		personalDeLaBase: t.TempDir(),
		temporalDeLaBase: t.TempDir(),
		arbolPreparado:   new(bool),
		queVen: []string{
			"LANG=C", "LC_ALL=C", "LC_CTYPE=C", "LC_MESSAGES=C", "TERM=dumb", "USER=persona-de-la-base",
			"LOGNAME=persona-de-la-base", "SHELL=/bin/sh", "TZ=UTC", variableDeLaSuscripcion + "=" + valorDeLaSuscripcion,
		},
	}

	sondeo.ejecutar = SondeoAEjecutar{
		Argumentos: ArgumentosDelSondeo{
			Skill: skillQueSondea, Evals: "01", Modelo: modeloDeLaSesion, Repeticiones: caso.repeticiones,
			Concurrencia: caso.concurrencia,
		},
		Entorno: sobreLaBase(os.Environ(), slices.Concat([]string{
			"PATH=" + sondeo.rutaDeLaBase, "HOME=" + sondeo.personalDeLaBase, "TMPDIR=" + sondeo.temporalDeLaBase,
			"ANTHROPIC_API_KEY=de-la-base", "ANTHROPIC_AUTH_TOKEN=de-la-base", "OTRA_DE_LA_BASE=si",
		}, sondeo.queVen), accesosDeClaudeCode...),
		EvalsDeLasSkills: carpetasDeEvals(t, skillQueSondea,
			entradaDeConjunto{nombre: nombreDeEval, contenido: contenidoDelArticulo21}),
		Temporal:        t.TempDir(),
		Guion:           guion,
		PrepararElArbol: arbolSinConstruir(sondeo.arbolPreparado),
	}

	if caso.ajustar != nil {
		caso.ajustar(&sondeo.ejecutar)
	}

	sondeo.plan = planDelSondeoDelTest(t, sondeo.ejecutar, caso.repeticiones)
	sondeo.escribirLasSesiones(t, caso)

	return sondeo
}

// planDelSondeoDelTest es el plan que abre el sondeo del test si sus argumentos
// valen: la eval sintética con modeloDeLaSesion y las repeticiones del caso.
// Con unas repeticiones que no valen, ninguno.
func planDelSondeoDelTest(t *testing.T, sondeo SondeoAEjecutar, repeticiones string) []SesionPlanificada {
	t.Helper()

	veces, err := strconv.Atoi(repeticiones)
	if err != nil || veces < 1 {
		return nil
	}

	conjunto, err := LeerConjunto(filepath.Join(sondeo.EvalsDeLasSkills, skillQueSondea))
	require.NoError(t, err)

	return PlanDeEvals{Evals: conjunto.Evals, ModeloQueDecide: modeloDeLaSesion, Repeticiones: veces}.Sesiones()
}

// escribirLasSesiones deja el transcript y la espera de cada sesión del plan
// que el caso da, por su posición.
func (d sondeoDelTest) escribirLasSesiones(t *testing.T, caso casoDeSondear) {
	t.Helper()

	if caso.transcripts == nil {
		return
	}

	for posicion, transcript := range caso.transcripts(t) {
		d.sustitutos.escribirTranscript(t, d.plan[posicion].Nombre, transcript)
	}

	for posicion, segundos := range caso.esperas {
		d.sustitutos.escribirEspera(t, d.plan[posicion].Nombre, segundos)
	}
}

// arbolSinConstruir es el árbol del sondeo de TestSondear, que no construye ni
// instala nada: deja, a través de un os.Root, un kitlegal que no hace nada en el
// bin/ del temporal y una skill vacía en el directorio de skills de Claude Code
// de su HOME, y anota que se preparó.
func arbolSinConstruir(preparado *bool) func(temporal string, base []string) error {
	return func(temporal string, _ []string) (err error) {
		*preparado = true

		raiz, err := os.OpenRoot(temporal)
		if err != nil {
			return err
		}

		defer func() { err = errors.Join(err, raiz.Close()) }()

		return errors.Join(
			raiz.Mkdir("bin", 0o700),
			escribirEjecutable(raiz, filepath.Join("bin", programaDeLasConsultas), kitlegalQueNoHaceNada),
			raiz.MkdirAll(filepath.Join("home", ".claude", "skills", skillQueSondea), 0o700),
		)
	}
}

// exigirLaBaseSinTocar exige que el HOME y el TMPDIR de la base del sondeo sigan
// vacíos: las sesiones no escriben en el entorno de quien lo lanza (FR-064).
func (d sondeoDelTest) exigirLaBaseSinTocar(t *testing.T) {
	t.Helper()

	for _, dir := range []string{d.personalDeLaBase, d.temporalDeLaBase} {
		entradas, err := os.ReadDir(dir)
		require.NoError(t, err)
		assert.Emptyf(t, entradas, "nada se escribe en %s, de la base", dir)
	}
}

// exigirQueNoSondea exige que el sondeo no haya preparado el árbol, ni abierto
// ninguna sesión, ni escrito nada en el temporal.
func (d sondeoDelTest) exigirQueNoSondea(t *testing.T) {
	t.Helper()

	assert.False(t, *d.arbolPreparado, "no se prepara el árbol")

	for _, dir := range []string{filepath.Join(d.sustitutos.comun, sustitutoClaude), d.ejecutar.Temporal} {
		entradas, err := os.ReadDir(dir)
		require.NoError(t, err)
		assert.Emptyf(t, entradas, "no se abre ninguna sesión ni se escribe nada: %s", dir)
	}
}

// exigirLasSesiones exige que el sondeo haya preparado el árbol y abierto las
// primeras sesiones del plan, cada una con el entorno de contracts/sondeo.md §5
// y terminada; que no haya abierto ninguna más; y que el temporal lleve solo
// bin/, home/ y sesiones/.
func (d sondeoDelTest) exigirLasSesiones(t *testing.T, abiertas int) {
	t.Helper()

	assert.True(t, *d.arbolPreparado, "se prepara el árbol")

	for _, sesion := range d.plan[:abiertas] {
		require.Truef(t, d.sustitutos.llego(t, sesion.Nombre), "la sesión %s se abre", sesion.Nombre)
		assert.Truef(t, d.sustitutos.cumplioLaEspera(t, sesion.Nombre), "la sesión %s termina", sesion.Nombre)

		d.exigirElEntornoDeLaSesion(t, sesion.Nombre)
	}

	for _, sesion := range d.plan[abiertas:] {
		assert.Falsef(t, d.sustitutos.llego(t, sesion.Nombre), "la sesión %s no se abre", sesion.Nombre)
	}

	entradas, err := os.ReadDir(d.ejecutar.Temporal)
	require.NoError(t, err)

	nombres := make([]string, 0, len(entradas))
	for _, entrada := range entradas {
		nombres = append(nombres, entrada.Name())
	}

	assert.Equal(t, []string{"bin", "home", "sesiones"}, nombres, "el temporal, sin informe ni veredicto")
}

// exigirElEntornoDeLaSesion exige que el sustituto de claude haya visto en la
// sesión el entorno de contracts/sondeo.md §5 de H7.3: el kitlegal del bin/ del
// temporal, el HOME del temporal —donde escribe, y no en el de la base—, el PATH
// de la base con ese bin/ delante, CLAUDE_CODE_OAUTH_TOKEN y las demás variables
// de §5 con su valor de la base; y no ANTHROPIC_API_KEY, ANTHROPIC_AUTH_TOKEN
// ni OTRA_DE_LA_BASE, que están en la base.
func (d sondeoDelTest) exigirElEntornoDeLaSesion(t *testing.T, sesion string) {
	t.Helper()

	anotado := d.sustitutos.entorno(t, sesion)
	binarios := filepath.Join(d.ejecutar.Temporal, "bin")
	personal := filepath.Join(d.ejecutar.Temporal, "home")

	assert.Equal(t, filepath.Join(binarios, programaDeLasConsultas), anotado["kitlegal"],
		"el primer kitlegal del PATH es el del temporal")
	assert.Equal(t, personal, anotado["HOME"])
	assert.FileExists(t, filepath.Join(personal, prefijoDeLoEscrito+sesion), "la sesión escribe en el HOME del temporal")
	assert.Equal(t, binarios+string(os.PathListSeparator)+d.rutaDeLaBase, anotado["PATH"])

	for _, variable := range d.queVen {
		nombre, valor, _ := strings.Cut(variable, "=")
		assert.Equalf(t, valor, anotado[nombre], "la sesión ve %s con su valor de la base", nombre)
	}

	assert.Equal(t, "no", anotado["ve-ANTHROPIC_API_KEY"], "la sesión no ve ANTHROPIC_API_KEY")
	assert.Equal(t, "no", anotado["ve-ANTHROPIC_AUTH_TOKEN"], "la sesión no ve ANTHROPIC_AUTH_TOKEN")
	assert.NotContains(t, anotado, "OTRA_DE_LA_BASE", "la sesión no ve otras variables de la base")
}

// guionDelSondeo es scripts/evals-sondeo.sh, relativo al directorio de este
// paquete, que es donde go test ejecuta los tests.
const guionDelSondeo = "../../scripts/evals-sondeo.sh"

// ficheroDeLaSalidaDelSondeo es el fichero del temporal del sondeo en el que
// TestSondeo deja la salida y del que la imprime su guion (contracts/sondeo.md
// §2 y §3 de H7.3).
const ficheroDeLaSalidaDelSondeo = "salida.txt"

// nombreDelTemporalDelSondeo es el del directorio que el guion del sondeo crea
// en TMPDIR: la plantilla de mktemp, kitlegal-sondeo.XXXXXX, con sus seis
// caracteres sustituidos.
var nombreDelTemporalDelSondeo = regexp.MustCompile(`^kitlegal-sondeo\.[A-Za-z0-9]{6}$`)

// casoDelGuionDelSondeo es un caso de TestGuionDelSondeo: la orden que ejecuta
// el guion de la ruta dada, escrita entera con constantes (gosec G204), y el
// código con el que sale el sustituto de go; y lo que se espera del guion: su
// código, su salida estándar, los fragmentos de su salida de error, que sin
// ninguno queda vacía, y si no ejecuta go.
type casoDelGuionDelSondeo struct {
	nombre     string
	orden      func(ctx context.Context, guion string) *exec.Cmd
	codigoDeGo int

	codigo  int
	salida  string
	deError []string
	sinGo   bool
}

// TestGuionDelSondeo fija scripts/evals-sondeo.sh (contracts/sondeo.md §2, §6 y
// §7 de H7.3; FR-064, FR-066; SC-009; US4-4) con el sustituto de go delante en
// el PATH y un TMPDIR vacío del test. Con cinco argumentos, crea en TMPDIR su
// temporal con la plantilla kitlegal-sondeo.XXXXXX y ejecuta, en la raíz del
// repositorio y con el tmp/ de ese temporal como TMPDIR, la orden go test del
// punto de entrada del sondeo con los cinco tras -args y -temporal con ese
// temporal, con sus dos salidas en su go-test.log. Si go test sale con 0, el
// guion sale con 0, su salida estándar es salida.txt y la de error queda vacía;
// si sale con otro código, el guion sale con 1, no imprime salida.txt y su
// salida de error lleva lo que go test escribió en sus dos salidas. Sin cinco
// argumentos, el uso y el código 1, sin ejecutar go. En todos los casos, el
// TMPDIR del test queda vacío: lo que go test deja en su TMPDIR —su directorio
// de trabajo, el de go install y la preparación de cada sesión, que el
// sustituto no borra— está en el temporal, y el guion lo borra (FR-064).
func TestGuionDelSondeo(t *testing.T) {
	t.Parallel()

	guion, err := filepath.Abs(guionDelSondeo)
	require.NoError(t, err)

	raiz, err := filepath.EvalSymlinks(filepath.Dir(filepath.Dir(guion)))
	require.NoError(t, err)

	for _, caso := range casosDelGuionDelSondeo() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			comun, temporal := t.TempDir(), t.TempDir()
			orden := caso.orden(t.Context(), guion)

			codigo, salida, deError := ejecutarElGuionDelSondeo(t, orden, caso.codigoDeGo, comun, temporal)

			assert.Equalf(t, caso.codigo, codigo, "el código del guion, con esta salida de error:\n%s", deError)
			assert.Equal(t, caso.salida, salida, "la salida estándar del guion")

			for _, fragmento := range caso.deError {
				assert.Contains(t, deError, fragmento, "la salida de error del guion")
			}

			if len(caso.deError) == 0 {
				assert.Empty(t, deError, "la salida de error del guion")
			}

			entradas, err := os.ReadDir(temporal)
			require.NoError(t, err)
			assert.Empty(t, entradas, "el guion borra su temporal y el TMPDIR queda vacío")

			exigirLaOrdenDeGo(t, caso.sinGo, orden.Args[1:], filepath.Join(comun, sustitutoGo), temporal, raiz)
		})
	}
}

// casosDelGuionDelSondeo son los casos de TestGuionDelSondeo: go test sale con
// 0, con la concurrencia vacía; sale con 1 y con 2, con una concurrencia
// pedida; y el guion recibe cuatro argumentos.
func casosDelGuionDelSondeo() []casoDelGuionDelSondeo {
	registro := []string{registroDeGoEnSuSalida, registroDeGoEnLaDeError}

	return []casoDelGuionDelSondeo{
		{
			nombre: "go-test-sale-con-0",
			orden: func(ctx context.Context, guion string) *exec.Cmd {
				return exec.CommandContext(ctx, guion, skillQueSondea, "03,14", modeloSonnet5, "3", "")
			},
			codigo: 0,
			salida: salidaDelSustitutoDeGo,
		},
		{
			nombre: "go-test-sale-con-1",
			orden: func(ctx context.Context, guion string) *exec.Cmd {
				return exec.CommandContext(ctx, guion, skillQueSondea, "03", modeloSonnet5, "1", "2")
			},
			codigoDeGo: 1,
			codigo:     1,
			deError:    registro,
		},
		{
			nombre: "go-test-sale-con-2",
			orden: func(ctx context.Context, guion string) *exec.Cmd {
				return exec.CommandContext(ctx, guion, otraSkillQueSondea, "01,03", modeloSonnet5, "2", "1")
			},
			codigoDeGo: 2,
			codigo:     1,
			deError:    registro,
		},
		{
			nombre: "sin-cinco-argumentos",
			orden: func(ctx context.Context, guion string) *exec.Cmd {
				return exec.CommandContext(ctx, guion, skillQueSondea, "03", modeloSonnet5, "3")
			},
			codigo:  1,
			deError: []string{"uso: scripts/evals-sondeo.sh"},
			sinGo:   true,
		},
	}
}

// ejecutarElGuionDelSondeo ejecuta la orden del guion del sondeo con el
// sustituto de go delante en el PATH, el TMPDIR y el directorio común dados y
// el código de go dado, y devuelve su código y sus dos salidas.
func ejecutarElGuionDelSondeo(t *testing.T, orden *exec.Cmd, codigoDeGo int, comun, temporal string,
) (codigo int, salida, deError string) {
	t.Helper()

	orden.Env = sobreLaBase(os.Environ(), []string{
		"PATH=" + escribirElGoDelSondeo(t) + string(os.PathListSeparator) + os.Getenv("PATH"),
		"TMPDIR=" + temporal,
		variableDelComun + "=" + comun,
		variableDeCodigo + "=" + strconv.Itoa(codigoDeGo),
	})

	var estandar, deErr strings.Builder

	orden.Stdout, orden.Stderr = &estandar, &deErr

	var terminada *exec.ExitError

	if err := orden.Run(); errors.As(err, &terminada) {
		codigo = terminada.ExitCode()
	} else {
		require.NoError(t, err)
	}

	return codigo, estandar.String(), deErr.String()
}

// exigirLaOrdenDeGo exige, si el guion ejecuta go, que el sustituto se haya
// ejecutado en la raíz del repositorio con la orden del punto de entrada del
// sondeo (ordenDelSondeo), con los argumentos del guion y un temporal en el
// TMPDIR dado, con el nombre de la plantilla y que ya tenía su go-test.log y su
// tmp/, y con ese tmp/ como TMPDIR; y, si no, que no se haya ejecutado.
func exigirLaOrdenDeGo(t *testing.T, sinGo bool, delGuion []string, anotaciones, temporal, raiz string) {
	t.Helper()

	if sinGo {
		assert.NoDirExists(t, anotaciones, "sin cinco argumentos, el guion no ejecuta go")

		return
	}

	argumentos := argumentosAnotados(t, anotaciones)
	temporalDelGuion := argumentos[len(argumentos)-1]

	fisico, err := filepath.EvalSymlinks(temporal)
	require.NoError(t, err)

	assert.Equal(t, fisico, filepath.Dir(temporalDelGuion), "el temporal del guion está en TMPDIR")
	assert.Regexp(t, nombreDelTemporalDelSondeo, filepath.Base(temporalDelGuion))
	assert.Equal(t, ordenDelSondeo(delGuion, temporalDelGuion), argumentos)

	for anotacion, esperada := range map[string]string{
		"directorio": raiz + "\n",
		"tmpdir":     filepath.Join(temporalDelGuion, "tmp") + "\n",
		"temporal":   "go-test.log\ntmp\n",
	} {
		contenido, err := leerFichero(filepath.Join(anotaciones, anotacion))
		require.NoError(t, err)
		assert.Equalf(t, esperada, string(contenido), "lo que el sustituto de go anota en %s", anotacion)
	}
}

// ordenDelSondeo son los argumentos de go de la orden del punto de entrada del
// sondeo (contracts/sondeo.md §2 de H7.3), con los cinco del guion, en su
// orden, y su temporal.
func ordenDelSondeo(argumentos []string, temporal string) []string {
	return []string{
		"test", "-tags", "evals", "-count=1", "-timeout", "0", "-run", "^TestSondeo$", "./internal/evals/", "-args",
		"-skill", argumentos[0], "-evals", argumentos[1], "-modelo", argumentos[2], "-repeticiones", argumentos[3],
		"-concurrencia", argumentos[4], "-temporal", temporal,
	}
}
