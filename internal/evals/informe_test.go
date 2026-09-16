package evals

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// casosDeInforme es el directorio de los casos de TestInforme: cada
// subdirectorio es una ejecución entera, con sus evals en evals/ y un directorio
// por sesión en sesiones/, y sin-python.txt es el de todos (contrato job-de-evals
// §9.1).
const casosDeInforme = "testdata/sesiones/informe"

// Lo que TestInforme y TestEscribirInformeSinSusEntradas pasan a EscribirInforme
// y lo que buscan en lo que escribe (contrato job-de-evals §9).
const (
	// modeloQueDecide y commitEvaluado son el modelo que el job pide y el commit
	// evaluado. El modelo es a propósito distinto del que declaran los
	// transcripts sintéticos, modeloDeLosTranscripts, del que es prefijo: así no
	// pasan ni el modelo tomado de las sesiones ni los modelos de las sesiones
	// tomados del job, y el alias que el proveedor resuelve a una versión con
	// fecha sigue siendo el modelo que se pidió (contrato job-de-evals §4).
	modeloQueDecide        = "claude-haiku"
	commitEvaluado         = "0123456789abcdef0123456789abcdef01234567"
	modeloDeLosTranscripts = "claude-haiku-4-5"

	// modeloInformativoDelCaso es el de los modelos informativos de los casos que
	// los llevan, y el que se pide en el caso del modelo distinto.
	modeloInformativoDelCaso = "claude-opus-5"

	// ficheroDeLaEvalInformativa es la eval informativa del caso que la lleva.
	ficheroDeLaEvalInformativa = "13-por-materia.yaml"

	// ficheroSinPython es el de la comprobación sin Python de todos los casos.
	ficheroSinPython = "sin-python.txt"

	// casoAprobado es el caso cuyas entradas cambia, una a una,
	// TestEscribirInformeSinSusEntradas.
	casoAprobado = "aprobado"

	// Las sesiones de las ejecuciones sintéticas.
	sesionDelArticulo21   = "01-lpac-articulo-21"
	sesionDeLaPruebaDeRed = "01-lpac-articulo-21-prueba-de-red"
	sesionDeNoActivacion  = "11-no-activa-programacion"

	// Las dos invocaciones del bloque a9998, que no está en ninguna grabación.
	ordenDeA9998        = "boe articulo BOE-A-2015-10565 a9998 --json"
	ordenDeA9998Offline = "boe articulo BOE-A-2015-10565 a9998 --offline --json"
)

// informeLeido es lo que EscribirInforme dejó en su destino: informe.json leído
// como Informe y, en crudo, lo que se compara tal como está escrito, e
// informe.md. caso es el directorio del caso, vacío si las entradas no son las de
// un caso.
type informeLeido struct {
	caso    string
	informe Informe
	crudo   informeCrudo
	md      string
}

// informeCrudo es lo que se lee de informe.json sin convertirlo a un tipo de Go,
// para distinguir null de una lista vacía: los motivos de la raíz y, de cada
// invocación de cada sesión, su código y sus conexiones.
type informeCrudo struct {
	Motivos jsontext.Value `json:"motivos"`
	Evals   []struct {
		Sesion       string            `json:"sesion"`
		Invocaciones []invocacionCruda `json:"invocaciones"`
	} `json:"evals"`
}

// invocacionCruda es una invocación de informe.json con su código y sus
// conexiones tal como están escritos.
type invocacionCruda struct {
	Orden      string         `json:"orden"`
	Codigo     jsontext.Value `json:"codigo"`
	Conexiones jsontext.Value `json:"conexiones"`
}

// TestInforme fija EscribirInforme sobre cada ejecución sintética del contrato
// job-de-evals §9.1 (§3.3, §5 y §9; data-model §10.3): la cabecera con el modelo
// y el commit recibidos, los modelos y las versiones de los transcripts y la
// comprobación sin Python byte a byte; cada sesión juzgada con la eval que nombra
// su eval.txt, o sin pasar con un motivo por cada fichero que falta o no se puede
// leer; los motivos de la raíz en su orden; y el veredicto, que falla con un
// fichero mal formado, una eval que no pasa, una eval sin ninguna sesión o una
// petición llegada a la red, y no con una invocación fuera de lo grabado (FR-071,
// FR-073, FR-076, SC-003, SC-012).
func TestInforme(t *testing.T) {
	t.Parallel()

	// tresRepeticiones y conModeloInformativo son los ajustes de los casos que no
	// se miden con una sola sesión por serie.
	tresRepeticiones := func(entradas *InformeAEscribir) { entradas.Repeticiones, entradas.Umbral = 3, 2 }
	conModeloInformativo := func(entradas *InformeAEscribir) {
		entradas.ModelosInformativos = []string{modeloInformativoDelCaso}
	}

	casos := []struct {
		nombre string

		// ajustar cambia las entradas del caso; nil deja las de entradasDelCaso.
		ajustar   func(entradas *InformeAEscribir)
		comprobar func(t *testing.T, leido informeLeido)
	}{
		{nombre: casoAprobado, comprobar: comprobarAprobado},
		{nombre: "fuera-de-lo-grabado-no-cambia-el-veredicto", comprobar: comprobarFueraDeLoGrabado},
		{nombre: "fichero-mal-formado", comprobar: comprobarFicheroMalFormado},
		{nombre: "eval-que-no-pasa", comprobar: comprobarEvalQueNoPasa},
		{nombre: "sesion-sin-terminar", comprobar: comprobarSesionSinTerminar},
		{nombre: "sesion-ilegible", comprobar: comprobarSesionIlegible},
		{nombre: "llegada-a-la-red", comprobar: comprobarLlegadaALaRed},
		{nombre: "sesion-cortada-con-invocaciones", comprobar: comprobarSesionCortada},
		{nombre: "eval-sin-sesion", comprobar: comprobarEvalSinSesion},
		{nombre: "sin-eval-txt", comprobar: comprobarSinEvalTxt},
		{nombre: "eval-desconocida", comprobar: comprobarEvalDesconocida},
		{nombre: "sin-pregunta-txt", comprobar: comprobarSinPreguntaTxt},
		{nombre: "sin-modelo-txt", comprobar: comprobarSinModeloTxt},
		{nombre: "traza-ilegible", comprobar: comprobarTrazaIlegible},
		{nombre: "umbral-alcanzado", ajustar: tresRepeticiones, comprobar: comprobarUmbralAlcanzado},
		{nombre: "umbral-no-alcanzado", ajustar: tresRepeticiones, comprobar: comprobarUmbralNoAlcanzado},
		{nombre: "eval-informativa-no-decide", comprobar: comprobarEvalInformativa},
		{
			nombre:    "modelos-informativos-no-deciden",
			ajustar:   conModeloInformativo,
			comprobar: comprobarModeloInformativo,
		},
		{
			nombre:    "faltan-sesiones",
			ajustar:   func(entradas *InformeAEscribir) { entradas.Repeticiones = 2 },
			comprobar: comprobarFaltanSesiones,
		},
		{
			nombre:    "otro-modelo-en-la-sesion",
			ajustar:   func(entradas *InformeAEscribir) { entradas.ModeloQueDecide = modeloInformativoDelCaso },
			comprobar: comprobarOtroModelo,
		},
	}

	nombres := make([]string, 0, len(casos))
	for _, caso := range casos {
		nombres = append(nombres, caso.nombre)
	}

	assert.Equal(t, ejecucionesDeInforme(t), slices.Sorted(slices.Values(nombres)),
		"cada directorio de %s es un caso de TestInforme, y cada caso tiene el suyo", casosDeInforme)

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			entradas := entradasDelCaso(caso.nombre, t.TempDir())
			if caso.ajustar != nil {
				caso.ajustar(&entradas)
			}

			informe, err := EscribirInforme(entradas)
			require.NoError(t, err)

			leido := leerInformeEscrito(t, entradas.Destino, informe)
			leido.caso = filepath.Join(casosDeInforme, caso.nombre)

			caso.comprobar(t, leido)
		})
	}
}

// TestEscribirInformeSinSusEntradas fija el contrato de error de EscribirInforme
// (contrato job-de-evals §3.3 y §9): con las entradas del caso aprobado de
// TestInforme y una sola cambiada, la comprobación sin Python, el directorio de
// sesiones o el de evals que no se pueden leer, o un destino en el que no se puede
// escribir —tampoco solo informe.json—, dan un error que nombra la ruta y ningún
// informe escrito, ni a medias (FR-081); y
// sin ninguna eval ni sesión que juzgar, el informe se escribe con veredicto
// fallo.
func TestEscribirInformeSinSusEntradas(t *testing.T) {
	t.Parallel()

	// contenidoDelDestino es el del fichero regular que ocupa el lugar del
	// directorio de destino.
	const contenidoDelDestino = "fichero que ocupa el lugar del directorio del informe\n"

	casos := []struct {
		nombre string

		// cambiar cambia una de las entradas y devuelve la ruta que el error tiene
		// que nombrar; vacía si EscribirInforme no devuelve ningún error.
		cambiar func(t *testing.T, entradas *InformeAEscribir) string

		// destinoEsUnFichero dice si el destino es un fichero regular, que tiene
		// que quedar igual, en lugar de un directorio sin informe.
		destinoEsUnFichero bool
	}{
		{
			nombre: "sin-python-inexistente",
			cambiar: func(t *testing.T, entradas *InformeAEscribir) string {
				t.Helper()

				entradas.SinPython = filepath.Join(t.TempDir(), ficheroSinPython)

				return entradas.SinPython
			},
		},
		{
			nombre: "sesiones-inexistente",
			cambiar: func(t *testing.T, entradas *InformeAEscribir) string {
				t.Helper()

				entradas.Sesiones = filepath.Join(t.TempDir(), "sesiones")

				return entradas.Sesiones
			},
		},
		{
			nombre: "evals-inexistente",
			cambiar: func(t *testing.T, entradas *InformeAEscribir) string {
				t.Helper()

				entradas.Evals = filepath.Join(t.TempDir(), "evals")

				return entradas.Evals
			},
		},
		{
			nombre: "destino-es-un-fichero",
			cambiar: func(t *testing.T, entradas *InformeAEscribir) string {
				t.Helper()

				entradas.Destino = filepath.Join(t.TempDir(), "informe")
				require.NoError(t, os.WriteFile(entradas.Destino, []byte(contenidoDelDestino), 0o600))

				return entradas.Destino
			},
			destinoEsUnFichero: true,
		},
		{
			// informe.md se puede escribir, pero un directorio ocupa el lugar de
			// informe.json: el informe.md ya escrito no se queda a medias.
			nombre: "informe-json-no-se-puede-escribir",
			cambiar: func(t *testing.T, entradas *InformeAEscribir) string {
				t.Helper()

				require.NoError(t, os.Mkdir(filepath.Join(entradas.Destino, "informe.json"), 0o750))

				return entradas.Destino
			},
		},
		{
			nombre: "evals-y-sesiones-vacios",
			cambiar: func(t *testing.T, entradas *InformeAEscribir) string {
				t.Helper()

				vacios := t.TempDir()
				entradas.Evals = filepath.Join(vacios, "evals")
				entradas.Sesiones = filepath.Join(vacios, "sesiones")

				require.NoError(t, os.Mkdir(entradas.Evals, 0o750))
				require.NoError(t, os.Mkdir(entradas.Sesiones, 0o750))

				return ""
			},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			entradas := entradasDelCaso(casoAprobado, t.TempDir())
			ruta := caso.cambiar(t, &entradas)

			informe, err := EscribirInforme(entradas)

			if ruta == "" {
				require.NoError(t, err)

				leido := leerInformeEscrito(t, entradas.Destino, informe)
				assert.Empty(t, leido.informe.Evals)
				exigirMotivosDeLaRaiz(t, leido, "ninguna eval bien formada que juzgar")
				assert.Equal(t, VeredictoFallo, leido.informe.Veredicto)

				return
			}

			require.Error(t, err)
			require.ErrorContains(t, err, ruta)
			assert.Zero(t, informe, "sin informe, EscribirInforme no devuelve ningún veredicto")

			if caso.destinoEsUnFichero {
				assert.Equal(t, contenidoDelDestino, contenidoDeLaSesion(t, filepath.Dir(entradas.Destino),
					filepath.Base(entradas.Destino)), "el fichero que ocupa el destino queda igual")

				return
			}

			assert.NoFileExists(t, filepath.Join(entradas.Destino, "informe.md"))
			assert.NoFileExists(t, filepath.Join(entradas.Destino, "informe.json"))
		})
	}
}

// comprobarAprobado exige el veredicto aprobado sin motivos ni nada que informar,
// la cabecera con los modelos y las versiones de los transcripts, cada uno una
// vez y en el orden de las sesiones, y la respuesta y el fin de la sesión del
// art. 21, con su pregunta y su respuesta en su sección de informe.md.
func comprobarAprobado(t *testing.T, leido informeLeido) {
	t.Helper()

	assert.Equal(t, VeredictoAprobado, leido.informe.Veredicto)
	exigirMotivosDeLaRaiz(t, leido)
	assert.Equal(t, "[]", string(leido.crudo.Motivos), "motivos es una lista vacía, no null")
	assert.Equal(t, "ninguno", seccionDelInforme(t, leido.md, "Ficheros mal formados"))
	assert.Equal(t, "ninguna", seccionDelInforme(t, leido.md, "Invocaciones fuera de lo grabado"))
	assert.Equal(t, "ninguna petición llegó a la red de una fuente",
		seccionDelInforme(t, leido.md, "Peticiones llegadas a la red"))

	assert.Equal(t, []string{modeloDeLosTranscripts}, leido.informe.ModelosDeSesion)
	assert.Equal(t, []string{"2.1.270", "2.1.269"}, leido.informe.VersionesDeClaudeCode)
	exigirLineas(t, seccionDelInforme(t, leido.md, "Cabecera"),
		"Modelos de las sesiones: "+modeloDeLosTranscripts, "Versiones de Claude Code: 2.1.270, 2.1.269")

	resultado := resultadoDeLaSesion(t, leido.informe, sesionDelArticulo21)
	assert.True(t, resultado.Pasa)
	assert.Equal(t, respuestaConCita, resultado.Respuesta)
	assert.Equal(t, "result success", resultado.FinDeLaSesion)
	assert.True(t, resultadoDeLaSesion(t, leido.informe, sesionDeNoActivacion).Pasa)

	seccion := seccionDelInforme(t, leido.md, "Sesión "+sesionDelArticulo21)
	assert.Contains(t, seccion, contenidoDeLaSesion(t, directorioDeSesion(leido, sesionDelArticulo21), "pregunta.txt"))
	assert.Contains(t, seccion, respuestaConCita)
}

// comprobarFueraDeLoGrabado exige las dos invocaciones de a9998 de la sesión de
// prueba de red fuera de lo grabado, con su sesión y su eval, sus conexiones en
// informe.json y en su sección de informe.md, y el veredicto aprobado.
func comprobarFueraDeLoGrabado(t *testing.T, leido informeLeido) {
	t.Helper()

	assert.Equal(t, VeredictoAprobado, leido.informe.Veredicto)

	for _, sesion := range []string{sesionDelArticulo21, sesionDeLaPruebaDeRed} {
		resultado := resultadoDeLaSesion(t, leido.informe, sesion)
		assert.Equal(t, ficheroDeLaEval01, resultado.Eval, "la sesión %s se juzga con la eval 01", sesion)
		assert.True(t, resultado.Pasa, "la sesión %s pasa", sesion)
	}

	assert.Equal(t, []FueraDeLoGrabadoDelInforme{
		{Sesion: sesionDeLaPruebaDeRed, Eval: ficheroDeLaEval01, Orden: ordenDeA9998, Codigo: 5},
		{Sesion: sesionDeLaPruebaDeRed, Eval: ficheroDeLaEval01, Orden: ordenDeA9998Offline, Codigo: 4},
	}, leido.informe.FueraDeLoGrabado)
	exigirLineas(t, seccionDelInforme(t, leido.md, "Invocaciones fuera de lo grabado"),
		filaDeTabla(sesionDeLaPruebaDeRed, ficheroDeLaEval01, ordenDeA9998, "5"),
		filaDeTabla(sesionDeLaPruebaDeRed, ficheroDeLaEval01, ordenDeA9998Offline, "4"))

	resultado := resultadoDeLaSesion(t, leido.informe, sesionDeLaPruebaDeRed)
	assert.Equal(t, []ConexionInformada{{Destino: destinoDeBucle, Clase: ConexionLocal}},
		invocacionInformada(t, resultado, ordenDeA9998).Conexiones, "la pareja de las tres conexiones, una vez")
	assert.Equal(t, "[]", string(invocacionEscrita(t, leido, sesionDeLaPruebaDeRed, ordenDeA9998Offline).Conexiones))

	seccion := seccionDelInforme(t, leido.md, "Sesión "+sesionDeLaPruebaDeRed)
	conRed := filaQueEmpiezaPor(t, seccion, ordenDeA9998)
	assert.Contains(t, conRed, destinoDeBucle)
	assert.Contains(t, conRed, string(ConexionLocal))
	assert.Contains(t, filaQueEmpiezaPor(t, seccion, ordenDeA9998Offline), "sin conexiones")
}

// comprobarFicheroMalFormado exige el fichero mal formado con el error que da
// LeerConjunto, en informe.json, en informe.md y como único motivo de la raíz, la
// sesión juzgada con la eval bien formada y el veredicto fallo.
func comprobarFicheroMalFormado(t *testing.T, leido informeLeido) {
	t.Helper()

	const malFormado = "02-sin-pregunta.yaml"

	conjunto, err := LeerConjunto(filepath.Join(leido.caso, "evals"))
	require.NoError(t, err)
	require.Len(t, conjunto.MalFormados, 1)

	errorDelFichero := conjunto.MalFormados[0].Error.Error()

	assert.Equal(t, []FicheroMalFormadoDelInforme{{Fichero: malFormado, Error: errorDelFichero}},
		leido.informe.FicherosMalFormados)
	exigirLineas(t, seccionDelInforme(t, leido.md, "Ficheros mal formados"), "- "+malFormado+": "+errorDelFichero)
	assert.True(t, resultadoDeLaSesion(t, leido.informe, sesionDelArticulo21).Pasa)
	exigirMotivosDeLaRaiz(t, leido, malFormado+": mal formado: "+errorDelFichero)
	assert.Equal(t, VeredictoFallo, leido.informe.Veredicto)
}

// comprobarEvalQueNoPasa exige la cita ausente como único motivo de la raíz,
// precedida de la sesión, y el veredicto fallo.
func comprobarEvalQueNoPasa(t *testing.T, leido informeLeido) {
	t.Helper()

	assert.False(t, resultadoDeLaSesion(t, leido.informe, sesionDelArticulo21).Pasa)
	exigirMotivosDeLaRaiz(t, leido,
		motivoDeLaTasa(ficheroDeLaEval01, modeloQueDecide, 0, 1, 1),
		sesionDelArticulo21+": cita ausente: "+textoDeLaCita21)
	assert.Equal(t, VeredictoFallo, leido.informe.Veredicto)
}

// comprobarSesionSinTerminar exige que la eval de no activación cuya sesión cortó
// el tope no pase, con el motivo del tope y ninguno de sesión ilegible, sin
// respuesta y con el fin de su último mensaje, y su motivo y su salida de error en
// su sección de informe.md.
func comprobarSesionSinTerminar(t *testing.T, leido informeLeido) {
	t.Helper()

	const motivo = "la sesión no terminó: tope de 240 s agotado (código 124)"

	resultado := resultadoDeLaSesion(t, leido.informe, sesionDeNoActivacion)
	assert.False(t, resultado.Pasa)
	assert.Empty(t, resultado.Respuesta)
	assert.Equal(t, "system", resultado.FinDeLaSesion)
	exigirMotivosDeLaRaiz(t, leido,
		motivoDeLaTasa(ficheroDeNoActivacion, modeloQueDecide, 0, 1, 1),
		sesionDeNoActivacion+": "+motivo)
	assert.Equal(t, VeredictoFallo, leido.informe.Veredicto)

	seccion := seccionDelInforme(t, leido.md, "Sesión "+sesionDeNoActivacion)
	assert.Contains(t, seccion, motivo)
	assert.Contains(t, seccion, contenidoDeLaSesion(t, directorioDeSesion(leido, sesionDeNoActivacion), "sesion.err"))
}

// comprobarSesionIlegible exige la sesión sin codigo-de-la-sesion ilegible y la
// cabecera sin ningún modelo ni versión, porque no se pudo leer ninguna sesión.
func comprobarSesionIlegible(t *testing.T, leido informeLeido) {
	t.Helper()

	motivo := exigirSesionIlegible(t, leido, "codigo-de-la-sesion")
	exigirMotivosDeLaRaiz(t, leido,
		motivoDeLaTasa(ficheroDeLaEval01, modeloQueDecide, 0, 1, 1),
		sesionDelArticulo21+": "+motivo)
	assert.Empty(t, leido.informe.ModelosDeSesion)
	assert.Empty(t, leido.informe.VersionesDeClaudeCode)
	exigirLineas(t, seccionDelInforme(t, leido.md, "Cabecera"),
		"Modelos de las sesiones: ninguno", "Versiones de Claude Code: ninguna")
}

// comprobarLlegadaALaRed exige la conexión de clase red en red, en su tabla de
// informe.md y como único motivo de la raíz, y el veredicto fallo aunque la eval
// pase.
func comprobarLlegadaALaRed(t *testing.T, leido informeLeido) {
	t.Helper()

	assert.True(t, resultadoDeLaSesion(t, leido.informe, sesionDelArticulo21).Pasa)
	assert.Equal(t, []RedDelInforme{
		{Sesion: sesionDelArticulo21, Eval: ficheroDeLaEval01, Orden: ordenDeA9998, Destino: destinoPublico},
	}, leido.informe.Red)
	exigirLineas(t, seccionDelInforme(t, leido.md, "Peticiones llegadas a la red"),
		filaDeTabla(sesionDelArticulo21, ficheroDeLaEval01, ordenDeA9998, destinoPublico))
	exigirMotivosDeLaRaiz(t, leido,
		sesionDelArticulo21+": petición llegada a la red: "+ordenDeA9998+" → "+destinoPublico)
	assert.Equal(t, VeredictoFallo, leido.informe.Veredicto)
}

// comprobarSesionCortada exige que la sesión que el tope cortó con código 137 no
// pase por el tope, el comando y la cita, sin ser ilegible; que sus dos
// invocaciones se informen, la que no terminó sin código y con su conexión; y que
// la terminada con 4 vaya a fuera de lo grabado y la conexión a red.
func comprobarSesionCortada(t *testing.T, leido informeLeido) {
	t.Helper()

	resultado := resultadoDeLaSesion(t, leido.informe, sesionDelArticulo21)
	assert.False(t, resultado.Pasa)
	assert.Equal(t, []string{
		"la sesión no terminó: terminada por señal tras el tope (código 137)",
		"comando ausente: " + textoDelComando21,
		"cita ausente: " + textoDeLaCita21,
	}, resultado.Motivos)
	assert.Empty(t, resultado.OtrasFallidas)

	require.Len(t, resultado.Invocaciones, 2)
	assert.Equal(t, ordenDeA9998Offline, resultado.Invocaciones[0].Orden, "primero la del proceso 2000")
	assert.Equal(t, codigoDeSalida(4), resultado.Invocaciones[0].Codigo)
	assert.Equal(t, "[]", string(invocacionEscrita(t, leido, sesionDelArticulo21, ordenDeA9998Offline).Conexiones))
	assert.Equal(t, ordenDeA9998, resultado.Invocaciones[1].Orden)
	assert.Nil(t, resultado.Invocaciones[1].Codigo)
	assert.Equal(t, "null", string(invocacionEscrita(t, leido, sesionDelArticulo21, ordenDeA9998).Codigo))
	assert.Equal(t, []ConexionInformada{{Destino: destinoPublico, Clase: ConexionRed}}, resultado.Invocaciones[1].Conexiones)

	assert.Equal(t, []FueraDeLoGrabadoDelInforme{
		{Sesion: sesionDelArticulo21, Eval: ficheroDeLaEval01, Orden: ordenDeA9998Offline, Codigo: 4},
	}, leido.informe.FueraDeLoGrabado)
	assert.Equal(t, []RedDelInforme{
		{Sesion: sesionDelArticulo21, Eval: ficheroDeLaEval01, Orden: ordenDeA9998, Destino: destinoPublico},
	}, leido.informe.Red)

	assert.Contains(t, filaQueEmpiezaPor(t, seccionDelInforme(t, leido.md, "Sesión "+sesionDelArticulo21), ordenDeA9998),
		"sin código (sesión cortada)")
	exigirLineas(t, seccionDelInforme(t, leido.md, "Invocaciones fuera de lo grabado"),
		filaDeTabla(sesionDelArticulo21, ficheroDeLaEval01, ordenDeA9998Offline, "4"))
	exigirLineas(t, seccionDelInforme(t, leido.md, "Peticiones llegadas a la red"),
		filaDeTabla(sesionDelArticulo21, ficheroDeLaEval01, ordenDeA9998, destinoPublico))

	motivosDeLaSesion := make([]string, 0, len(resultado.Motivos))
	for _, motivo := range resultado.Motivos {
		motivosDeLaSesion = append(motivosDeLaSesion, sesionDelArticulo21+": "+motivo)
	}

	exigirMotivosDeLaRaiz(t, leido, slices.Concat(
		[]string{motivoDeLaTasa(ficheroDeLaEval01, modeloQueDecide, 0, 1, 1)},
		motivosDeLaSesion,
		[]string{sesionDelArticulo21 + ": petición llegada a la red: " + ordenDeA9998 + " → " + destinoPublico},
	)...)
	assert.Equal(t, VeredictoFallo, leido.informe.Veredicto)
}

// comprobarEvalSinSesion exige un único resultado, que pasa, y el veredicto fallo
// porque la serie de la eval bien formada que ningún eval.txt nombra no tiene
// ninguna sesión.
func comprobarEvalSinSesion(t *testing.T, leido informeLeido) {
	t.Helper()

	require.Len(t, leido.informe.Evals, 1)
	assert.True(t, resultadoDeLaSesion(t, leido.informe, sesionDelArticulo21).Pasa)
	exigirMotivosDeLaRaiz(t, leido,
		motivoDeLasSesionesQueFaltan(ficheroDeNoActivacion, modeloQueDecide, 0, 1),
		motivoDeLaTasa(ficheroDeNoActivacion, modeloQueDecide, 0, 0, 1))
	assert.Equal(t, VeredictoFallo, leido.informe.Veredicto)
}

// comprobarSinEvalTxt exige la sesión sin eval.txt ilegible, con la eval vacía, y
// la eval 01 sin ninguna sesión que la juzgue.
func comprobarSinEvalTxt(t *testing.T, leido informeLeido) {
	t.Helper()

	motivo := exigirSesionIlegible(t, leido, "eval.txt")
	exigirMotivosDeLaRaiz(t, leido,
		slices.Concat(motivosDeLaSerieSinSesiones(), []string{sesionDelArticulo21 + ": " + motivo})...)
	assert.Empty(t, resultadoDeLaSesion(t, leido.informe, sesionDelArticulo21).Eval)
}

// comprobarEvalDesconocida exige la sesión cuyo eval.txt no nombra ninguna eval
// ilegible, con la eval que nombra en el resultado y en el motivo, y la eval 01 sin
// ninguna sesión que la juzgue.
func comprobarEvalDesconocida(t *testing.T, leido informeLeido) {
	t.Helper()

	const desconocida = "03-inexistente.yaml"

	motivo := exigirSesionIlegible(t, leido, "eval.txt")
	exigirMotivosDeLaRaiz(t, leido,
		slices.Concat(motivosDeLaSerieSinSesiones(), []string{sesionDelArticulo21 + ": " + motivo})...)
	assert.Contains(t, motivo, desconocida)
	assert.Equal(t, desconocida, resultadoDeLaSesion(t, leido.informe, sesionDelArticulo21).Eval)
}

// comprobarSinPreguntaTxt exige la sesión sin pregunta.txt ilegible y ningún
// motivo de eval sin sesión, porque su eval.txt nombra la eval 01.
func comprobarSinPreguntaTxt(t *testing.T, leido informeLeido) {
	t.Helper()

	motivo := exigirSesionIlegible(t, leido, "pregunta.txt")
	exigirMotivosDeLaRaiz(t, leido,
		motivoDeLaTasa(ficheroDeLaEval01, modeloQueDecide, 0, 1, 1),
		sesionDelArticulo21+": "+motivo)
}

// comprobarTrazaIlegible exige la sesión con la línea execve cortada ilegible por
// su traza, con el fichero, el número de línea y su texto en el motivo.
func comprobarTrazaIlegible(t *testing.T, leido informeLeido) {
	t.Helper()

	traza := filepath.Join(directorioDeSesion(leido, sesionDelArticulo21), "traza", "t.2000")

	motivo := exigirSesionIlegible(t, leido, "traza")
	assert.Contains(t, motivo, traza)
	assert.Contains(t, motivo, "línea 1")
	assert.Contains(t, motivo, lineaDeLaTraza(t, traza, 1))
	exigirMotivosDeLaRaiz(t, leido,
		motivoDeLaTasa(ficheroDeLaEval01, modeloQueDecide, 0, 1, 1),
		sesionDelArticulo21+": "+motivo)
}

// exigirSesionIlegible exige que la sesión del art. 21 no pase con un único
// motivo, el de sesión ilegible por el fichero, y el veredicto fallo. Los motivos
// de la raíz los exige cada caso, porque dependen de en qué serie cae la sesión:
// la que pierde su eval o su modelo deja además la serie planificada sin
// sesiones. Devuelve el motivo.
func exigirSesionIlegible(t *testing.T, leido informeLeido, fichero string) string {
	t.Helper()

	resultado := resultadoDeLaSesion(t, leido.informe, sesionDelArticulo21)
	assert.False(t, resultado.Pasa)
	require.Len(t, resultado.Motivos, 1, "un único motivo: %q", resultado.Motivos)

	motivo := resultado.Motivos[0]
	assert.True(t, strings.HasPrefix(motivo, "sesión ilegible: "+fichero+": "),
		"el motivo %q es el de sesión ilegible por %s", motivo, fichero)

	assert.Equal(t, VeredictoFallo, leido.informe.Veredicto)

	return motivo
}

// motivosDeLaSerieSinSesiones son los dos motivos de la serie del art. 21 con el
// modelo que decide cuando ninguna sesión cae en ella, con una repetición y umbral
// 1.
func motivosDeLaSerieSinSesiones() []string {
	return []string{
		motivoDeLasSesionesQueFaltan(ficheroDeLaEval01, modeloQueDecide, 0, 1),
		motivoDeLaTasa(ficheroDeLaEval01, modeloQueDecide, 0, 0, 1),
	}
}

// comprobarSinModeloTxt exige la sesión sin modelo.txt ilegible: sin el modelo no
// cae en la serie que el plan pide, que queda sin sesiones.
func comprobarSinModeloTxt(t *testing.T, leido informeLeido) {
	t.Helper()

	motivo := exigirSesionIlegible(t, leido, "modelo.txt")
	assert.Empty(t, resultadoDeLaSesion(t, leido.informe, sesionDelArticulo21).Modelo)
	exigirMotivosDeLaRaiz(t, leido,
		slices.Concat(motivosDeLaSerieSinSesiones(), []string{sesionDelArticulo21 + ": " + motivo})...)
}

// comprobarUmbralAlcanzado exige que la serie del art. 21 con tres sesiones, dos
// de las cuales pasan, llegue al umbral de 2 y que el veredicto sea aprobado sin
// ningún motivo, aunque una sesión no pase: eso es lo que el umbral absorbe
// (ADR 0016).
func comprobarUmbralAlcanzado(t *testing.T, leido informeLeido) {
	t.Helper()

	tasa := tasaDeLaSerie(t, leido.informe, ficheroDeLaEval01, modeloQueDecide)
	assert.Equal(t, TasaDelInforme{
		Eval: ficheroDeLaEval01, Modelo: modeloQueDecide, Planificada: true, Decide: true,
		Sesiones: 3, Pasan: 2, Pasa: true,
	}, tasa)

	assert.False(t, resultadoDeLaSesion(t, leido.informe, sesionDelArticulo21+"-claude-haiku-03").Pasa,
		"la tercera sesión no pasa y aun así la serie llega al umbral")
	exigirMotivosDeLaRaiz(t, leido)
	assert.Equal(t, VeredictoAprobado, leido.informe.Veredicto)

	exigirLineas(t, seccionDelInforme(t, leido.md, "Tasas por eval"),
		filaDeTabla(ficheroDeLaEval01, modeloQueDecide, "sí", "sí", "2 de 3", "llega al umbral"))
}

// comprobarUmbralNoAlcanzado exige que la serie con solo una sesión que pasa de
// tres no llegue al umbral, con su tasa y los motivos de sus dos sesiones que no
// pasan, y el veredicto fallo.
func comprobarUmbralNoAlcanzado(t *testing.T, leido informeLeido) {
	t.Helper()

	tasa := tasaDeLaSerie(t, leido.informe, ficheroDeLaEval01, modeloQueDecide)
	assert.Equal(t, 1, tasa.Pasan)
	assert.False(t, tasa.Pasa)

	motivoDeLaCita := "cita ausente: " + textoDeLaCita21
	exigirMotivosDeLaRaiz(t, leido,
		motivoDeLaTasa(ficheroDeLaEval01, modeloQueDecide, 1, 3, 2),
		sesionDelArticulo21+"-claude-haiku-02: "+motivoDeLaCita,
		sesionDelArticulo21+"-claude-haiku-03: "+motivoDeLaCita)
	assert.Equal(t, VeredictoFallo, leido.informe.Veredicto)
}

// comprobarEvalInformativa exige que la serie de una eval informativa no decida,
// que su tasa se publique igual y que el veredicto sea aprobado aunque su sesión
// no pase (ADR 0016).
func comprobarEvalInformativa(t *testing.T, leido informeLeido) {
	t.Helper()

	assert.Equal(t, TasaDelInforme{
		Eval: ficheroDeLaEvalInformativa, Modelo: modeloQueDecide, Planificada: true,
		Sesiones: 1, Pasan: 0,
	}, tasaDeLaSerie(t, leido.informe, ficheroDeLaEvalInformativa, modeloQueDecide))

	assert.True(t, tasaDeLaSerie(t, leido.informe, ficheroDeLaEval01, modeloQueDecide).Decide)
	exigirMotivosDeLaRaiz(t, leido)
	assert.Equal(t, VeredictoAprobado, leido.informe.Veredicto)
}

// comprobarModeloInformativo exige que la serie de uno de los modelos
// informativos no decida, que su tasa se publique y que el veredicto sea aprobado aunque su sesión
// no pase (ADR 0016).
func comprobarModeloInformativo(t *testing.T, leido informeLeido) {
	t.Helper()

	assert.Equal(t, TasaDelInforme{
		Eval: ficheroDeLaEval01, Modelo: modeloInformativoDelCaso, Planificada: true,
		Sesiones: 1, Pasan: 0,
	}, tasaDeLaSerie(t, leido.informe, ficheroDeLaEval01, modeloInformativoDelCaso))

	assert.Equal(t, []string{modeloInformativoDelCaso}, leido.informe.ModelosInformativos)
	exigirLineas(t, seccionDelInforme(t, leido.md, "Cabecera"),
		"Modelos informativos: "+modeloInformativoDelCaso)
	exigirMotivosDeLaRaiz(t, leido)
	assert.Equal(t, VeredictoAprobado, leido.informe.Veredicto)
}

// comprobarFaltanSesiones exige que una serie con menos sesiones de las que pide
// el plan haga fallar el veredicto, aunque las que hay pasen: si no, una sesión
// que el guion no llegó a abrir se perdería en silencio.
func comprobarFaltanSesiones(t *testing.T, leido informeLeido) {
	t.Helper()

	tasa := tasaDeLaSerie(t, leido.informe, ficheroDeLaEval01, modeloQueDecide)
	assert.Equal(t, 1, tasa.Sesiones)
	assert.True(t, tasa.Pasa, "la sesión que hay pasa y llega al umbral")

	exigirMotivosDeLaRaiz(t, leido, motivoDeLasSesionesQueFaltan(ficheroDeLaEval01, modeloQueDecide, 1, 2))
	assert.Equal(t, VeredictoFallo, leido.informe.Veredicto)
}

// comprobarOtroModelo exige que una sesión que declara un modelo que no es el que
// se le pidió no pase, con su motivo, y que el informe publique los dos ids.
func comprobarOtroModelo(t *testing.T, leido informeLeido) {
	t.Helper()

	sesion := sesionDelArticulo21 + "-claude-opus-5-01"

	resultado := resultadoDeLaSesion(t, leido.informe, sesion)
	assert.False(t, resultado.Pasa)
	assert.Equal(t, modeloInformativoDelCaso, resultado.Modelo)
	assert.Equal(t, modeloDeLosTranscripts, resultado.ModeloDeLaSesion)

	motivo := "la sesión no declara el modelo que se le pidió: " + modeloDeLosTranscripts +
		", y se pidió " + modeloInformativoDelCaso
	exigirMotivosDeLaRaiz(t, leido,
		motivoDeLaTasa(ficheroDeLaEval01, modeloInformativoDelCaso, 0, 1, 1),
		sesion+": "+motivo)
	assert.Equal(t, VeredictoFallo, leido.informe.Veredicto)

	exigirLineas(t, seccionDelInforme(t, leido.md, "Sesión "+sesion),
		"Modelo pedido: "+modeloInformativoDelCaso, "Modelo de la sesión: "+modeloDeLosTranscripts)
}

// ejecucionesDeInforme son los nombres de los casos de casosDeInforme, cada uno
// un directorio, en orden; la comprobación sin Python es el único fichero que no
// es un caso.
func ejecucionesDeInforme(t *testing.T) []string {
	t.Helper()

	entradas, err := os.ReadDir(casosDeInforme)
	require.NoError(t, err)

	nombres := make([]string, 0, len(entradas))

	for _, entrada := range entradas {
		if entrada.Name() == ficheroSinPython {
			continue
		}

		require.True(t, entrada.IsDir(), "en %s cada caso es el directorio de una ejecución", casosDeInforme)

		nombres = append(nombres, entrada.Name())
	}

	return nombres
}

// entradasDelCaso son las entradas de EscribirInforme de una ejecución de
// casosDeInforme, con el destino dado: una repetición por serie y umbral 1, que
// es lo que tiene la mayoría de los casos; los que miden el umbral, las
// repeticiones o los modelos informativos las cambian con su propio ajustar.
func entradasDelCaso(caso, destino string) InformeAEscribir {
	return InformeAEscribir{
		Skill:           skillDeLasSesiones,
		Evals:           filepath.Join(casosDeInforme, caso, "evals"),
		Sesiones:        filepath.Join(casosDeInforme, caso, "sesiones"),
		Destino:         destino,
		ModeloQueDecide: modeloQueDecide,
		Repeticiones:    1,
		Umbral:          1,
		Commit:          commitEvaluado,
		SinPython:       filepath.Join(casosDeInforme, ficheroSinPython),
	}
}

// motivoDeLasSesionesQueFaltan y motivoDeLaTasa son los dos motivos de la raíz que
// da una serie (data-model §10.3).
func motivoDeLasSesionesQueFaltan(eval, modelo string, hay, pide int) string {
	return fmt.Sprintf("%s con %s: hay %d sesiones y el plan pide %d", eval, modelo, hay, pide)
}

func motivoDeLaTasa(eval, modelo string, pasan, sesiones, umbral int) string {
	return fmt.Sprintf("%s con %s: pasan %d de %d, y el umbral es %d", eval, modelo, pasan, sesiones, umbral)
}

// tasaDeLaSerie es la tasa de la serie de esa eval con ese modelo en el informe.
func tasaDeLaSerie(t *testing.T, informe Informe, eval, modelo string) TasaDelInforme {
	t.Helper()

	posicion := slices.IndexFunc(informe.Tasas, func(tasa TasaDelInforme) bool {
		return tasa.Eval == eval && tasa.Modelo == modelo && !tasa.PreguntaAmpliada
	})
	require.GreaterOrEqual(t, posicion, 0, "el informe tiene la tasa de %s con %s", eval, modelo)

	return informe.Tasas[posicion]
}

// leerInformeEscrito lee informe.json e informe.md del destino y exige lo que
// todo informe cumple: los dos ficheros con permisos 0o600 y un salto de línea
// final; informe.json igual al Informe devuelto; y la cabecera, con la skill, el
// modelo y el commit recibidos y la comprobación sin Python byte a byte, en
// informe.json y en informe.md, que empieza por su título y lleva el veredicto de
// informe.json.
func leerInformeEscrito(t *testing.T, destino string, devuelto Informe) informeLeido {
	t.Helper()

	escrito := contenidoDelInforme(t, destino, "informe.json")

	codificado, err := json.Marshal(devuelto)
	require.NoError(t, err)
	assert.JSONEq(t, string(codificado), escrito, "informe.json es el Informe que devuelve EscribirInforme")

	leido := informeLeido{md: contenidoDelInforme(t, destino, "informe.md")}
	require.NoError(t, json.Unmarshal([]byte(escrito), &leido.informe))
	require.NoError(t, json.Unmarshal([]byte(escrito), &leido.crudo))

	sinPython := contenidoDeLaSesion(t, casosDeInforme, ficheroSinPython)

	assert.Equal(t, skillDeLasSesiones, leido.informe.Skill)
	assert.Equal(t, commitEvaluado, leido.informe.Commit)
	assert.Equal(t, sinPython, leido.informe.SinPython, "sin_python es sin-python.txt byte a byte")

	titulo, _, _ := strings.Cut(leido.md, "\n")
	assert.Equal(t, "# Informe de evals de "+skillDeLasSesiones, titulo)
	exigirLineas(t, seccionDelInforme(t, leido.md, "Veredicto"), "Veredicto: "+string(leido.informe.Veredicto))
	exigirLineas(t, seccionDelInforme(t, leido.md, "Cabecera"),
		"Modelo que decide: "+leido.informe.ModeloQueDecide,
		"Repeticiones por eval: "+strconv.Itoa(leido.informe.Repeticiones),
		"Umbral: "+strconv.Itoa(leido.informe.Umbral),
		"Commit: "+commitEvaluado)
	assert.Equal(t, "```text\n"+sinPython+"```", seccionDelInforme(t, leido.md, "Comprobación sin Python"))

	return leido
}

// contenidoDelInforme es el contenido de un fichero del informe, que tiene que
// estar en el destino con permisos 0o600 y terminar en un salto de línea.
func contenidoDelInforme(t *testing.T, destino, fichero string) string {
	t.Helper()

	info, err := os.Stat(filepath.Join(destino, fichero))
	require.NoError(t, err, "%s escrito en el destino", fichero)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm(), "%s escrito con permisos 0o600", fichero)

	contenido := contenidoDeLaSesion(t, destino, fichero)
	assert.True(t, strings.HasSuffix(contenido, "\n"), "%s termina en un salto de línea", fichero)

	return contenido
}

// exigirMotivosDeLaRaiz exige que los motivos de la raíz sean exactamente los
// esperados, en su orden, y que la sección del veredicto de informe.md lleve una
// línea por cada uno, o «Motivos: ninguno».
func exigirMotivosDeLaRaiz(t *testing.T, leido informeLeido, esperados ...string) {
	t.Helper()

	veredicto := seccionDelInforme(t, leido.md, "Veredicto")

	if len(esperados) == 0 {
		assert.Empty(t, leido.informe.Motivos)
		exigirLineas(t, veredicto, "Motivos: ninguno")

		return
	}

	assert.Equal(t, esperados, leido.informe.Motivos)

	for _, motivo := range esperados {
		exigirLineas(t, veredicto, "- "+motivo)
	}
}

// resultadoDeLaSesion es el resultado de la sesión en el informe.
func resultadoDeLaSesion(t *testing.T, informe Informe, sesion string) ResultadoDeEval {
	t.Helper()

	posicion := slices.IndexFunc(informe.Evals, func(resultado ResultadoDeEval) bool { return resultado.Sesion == sesion })
	require.GreaterOrEqual(t, posicion, 0, "el informe tiene el resultado de la sesión %s", sesion)

	return informe.Evals[posicion]
}

// invocacionInformada es la invocación de la orden en el resultado.
func invocacionInformada(t *testing.T, resultado ResultadoDeEval, orden string) InvocacionInformada {
	t.Helper()

	posicion := slices.IndexFunc(resultado.Invocaciones, func(invocacion InvocacionInformada) bool {
		return invocacion.Orden == orden
	})
	require.GreaterOrEqual(t, posicion, 0, "la sesión %s tiene la invocación %s", resultado.Sesion, orden)

	return resultado.Invocaciones[posicion]
}

// invocacionEscrita es, tal como está escrita en informe.json, la invocación de
// la orden en el resultado de la sesión.
func invocacionEscrita(t *testing.T, leido informeLeido, sesion, orden string) invocacionCruda {
	t.Helper()

	for _, resultado := range leido.crudo.Evals {
		if resultado.Sesion != sesion {
			continue
		}

		for _, invocacion := range resultado.Invocaciones {
			if invocacion.Orden == orden {
				return invocacion
			}
		}
	}

	require.FailNow(t, "informe.json no tiene la invocación", "sesión %s, orden %s", sesion, orden)

	return invocacionCruda{}
}

// directorioDeSesion es el directorio de la sesión en el caso leído.
func directorioDeSesion(leido informeLeido, sesion string) string {
	return filepath.Join(leido.caso, "sesiones", sesion)
}

// seccionDelInforme es el cuerpo de la sección ## titulo de informe.md, sin los
// blancos de los extremos: lo que hay hasta la sección siguiente.
func seccionDelInforme(t *testing.T, md, titulo string) string {
	t.Helper()

	_, resto, encontrada := strings.Cut(md, "\n## "+titulo+"\n")
	require.True(t, encontrada, "informe.md tiene la sección ## %s", titulo)

	cuerpo, _, _ := strings.Cut(resto, "\n## ")

	return strings.TrimSpace(cuerpo)
}

// exigirLineas exige que el texto tenga cada una de las líneas, enteras.
func exigirLineas(t *testing.T, texto string, lineas ...string) {
	t.Helper()

	todas := strings.Split(texto, "\n")

	for _, linea := range lineas {
		assert.True(t, slices.Contains(todas, linea), "falta la línea %q en:\n%s", linea, texto)
	}
}

// filaDeTabla es la fila de una tabla de informe.md con las celdas dadas.
func filaDeTabla(celdas ...string) string {
	return "| " + strings.Join(celdas, " | ") + " |"
}

// filaQueEmpiezaPor es la fila de una tabla del texto cuya primera celda es la
// dada.
func filaQueEmpiezaPor(t *testing.T, texto, celda string) string {
	t.Helper()

	for linea := range strings.Lines(texto) {
		if strings.HasPrefix(linea, "| "+celda+" |") {
			return strings.TrimSuffix(linea, "\n")
		}
	}

	require.FailNow(t, "falta la fila", "ninguna fila empieza por la celda %q en:\n%s", celda, texto)

	return ""
}
