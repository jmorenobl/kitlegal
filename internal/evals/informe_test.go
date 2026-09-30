package evals

import (
	"cmp"
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
	// transcripts sintéticos, modeloDeLosTranscripts, que es el pedido con la
	// fecha de su versión detrás: así no pasan ni el modelo tomado de las
	// sesiones ni los modelos de las sesiones tomados del job, y el alias que el
	// proveedor resuelve a una versión con fecha sigue siendo el modelo que se
	// pidió (contrato job-de-evals §4). Detrás del pedido va solo la fecha: con
	// otro sufijo sería otro modelo (ADR 0031).
	modeloQueDecide        = "claude-haiku"
	commitEvaluado         = "0123456789abcdef0123456789abcdef01234567"
	modeloDeLosTranscripts = "claude-haiku-20251001"

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

// Lo que TestInformeConAvisos cambia en su copia del caso aprobado (contrato de
// formato, juicio e informe §6 de H5.1; research D10).
const (
	// avisosDeLaEval01 es lo que se añade al final de la eval del art. 21: los dos
	// avisos que espera, en ese orden.
	avisosDeLaEval01 = "avisos:\n  - derogada\n  - vigencia-agotada\n"

	// prefijoDeDerogada es lo que se antepone a la respuesta de su sesión: la
	// frase del binario para derogada, con su forma fija, y un espacio.
	prefijoDeDerogada = avisoDeDerogada + " "

	// citaDeLaRespuesta es la cita del art. 21 tal como la escribe esa respuesta.
	citaDeLaRespuesta = "[BOE-A-2015-10565, bloque a21]"
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
// para distinguir null de una lista vacía: los motivos de la raíz, las formas
// exigidas de cada serie, el recuento de las expresiones prohibidas por modelo, los
// umbrales, la duración de las sesiones, los reintentos por límite de ritmo y las
// sesiones sin medir de la raíz y, de cada sesión, sus avisos encontrados y
// ausentes, sus expresiones prohibidas, sus reintentos, si quedó sin medir y, de
// cada una de sus invocaciones, su código y sus conexiones.
type informeCrudo struct {
	Motivos                        jsontext.Value   `json:"motivos"`
	Tasas                          []tasaCruda      `json:"tasas"`
	ExpresionesProhibidasPorModelo jsontext.Value   `json:"expresiones_prohibidas_por_modelo"`
	Umbrales                       jsontext.Value   `json:"umbrales"`
	DuracionDeLasSesiones          jsontext.Value   `json:"duracion_de_las_sesiones"`
	ReintentosPorLimiteDeRitmo     jsontext.Value   `json:"reintentos_por_limite_de_ritmo"`
	SesionesSinMedir               jsontext.Value   `json:"sesiones_sin_medir"`
	Evals                          []resultadoCrudo `json:"evals"`
}

// tasaCruda es una serie de informe.json con sus formas exigidas y sus sesiones
// sin medir tal como están escritas.
type tasaCruda struct {
	Eval     string         `json:"eval"`
	Modelo   string         `json:"modelo"`
	Formas   jsontext.Value `json:"formas"`
	SinMedir jsontext.Value `json:"sin_medir"`
}

// resultadoCrudo es el resultado de una sesión de informe.json con sus comandos
// prohibidos ejecutados, sus avisos, sus hallazgos, su territorio, sus expresiones
// prohibidas, sus reintentos por límite de ritmo, si quedó sin medir y sus
// invocaciones tal como están escritos.
type resultadoCrudo struct {
	Sesion                       string            `json:"sesion"`
	ComandosProhibidosEjecutados jsontext.Value    `json:"comandos_prohibidos_ejecutados"`
	AvisosEncontrados            jsontext.Value    `json:"avisos_encontrados"`
	AvisosAusentes               jsontext.Value    `json:"avisos_ausentes"`
	HallazgosEncontrados         jsontext.Value    `json:"hallazgos_encontrados"`
	HallazgosAusentes            jsontext.Value    `json:"hallazgos_ausentes"`
	RedaccionesEncontradas       jsontext.Value    `json:"redacciones_modificadas_encontradas"`
	RedaccionesAusentes          jsontext.Value    `json:"redacciones_modificadas_ausentes"`
	TerritorioEncontrado         jsontext.Value    `json:"territorio_encontrado"`
	TerritorioAusente            jsontext.Value    `json:"territorio_ausente"`
	ExpresionesProhibidas        jsontext.Value    `json:"expresiones_prohibidas"`
	ReintentosPorLimiteDeRitmo   jsontext.Value    `json:"reintentos_por_limite_de_ritmo"`
	SinMedir                     jsontext.Value    `json:"sin_medir"`
	Invocaciones                 []invocacionCruda `json:"invocaciones"`
}

// columnaDeExpresiones es la columna de las expresiones prohibidas de la tabla de
// las sesiones de informe.md (contrato lista-y-juicio §5 de H7.2).
const columnaDeExpresiones = "Expresiones prohibidas"

// encabezadosDeLaTablaDeSesiones son los de la tabla de las sesiones de
// informe.md, con los comandos prohibidos ejecutados junto a los comandos
// ausentes (contrato evals-y-skill §2 de H7), los avisos junto a las citas, los
// hallazgos junto a los avisos (contrato evals-y-skill §6 de H7.1), las
// redacciones modificadas detrás de los hallazgos (contracts/informe-del-job.md
// §4 de H7.4), el territorio detrás (contrato de evals §2 de H6), entre el
// territorio ausente y el resultado, las expresiones prohibidas (contrato
// lista-y-juicio §5 de H7.2) y, detrás de ellas, los reintentos por límite de
// ritmo y si la sesión quedó sin medir (contrato informe-del-job §4 de H7.3).
var encabezadosDeLaTablaDeSesiones = []string{
	"Sesión", "Eval", "Modelo", "Activa", "Activada", "Sesión terminada", "Comandos ausentes",
	"Comandos prohibidos ejecutados", "Citas ausentes", "Avisos encontrados", "Avisos ausentes",
	"Hallazgos encontrados", "Hallazgos ausentes", columnaDeRedaccionesEncontradas, columnaDeRedaccionesAusentes,
	"Territorio encontrado", "Territorio ausente", columnaDeExpresiones, columnaDeReintentos, columnaSinMedir,
	"Resultado",
}

// columnaDeRedaccionesEncontradas es la columna de la tabla de las sesiones de
// informe.md con las redacciones modificadas esperadas que la respuesta lleva, y
// columnaDeRedaccionesAusentes, la de las que no lleva
// (contracts/informe-del-job.md §4 de H7.4).
const (
	columnaDeRedaccionesEncontradas = "Redacciones modificadas encontradas"
	columnaDeRedaccionesAusentes    = "Redacciones modificadas ausentes"
)

// columnaDeReintentos es la columna de la tabla de las sesiones de informe.md con
// los reintentos por límite de ritmo de cada sesión, y columnaSinMedir, la que
// dice «no» o la clase por la que quedó sin medir (contrato informe-del-job §4 de
// H7.3).
const (
	columnaDeReintentos = "Reintentos por límite de ritmo"
	columnaSinMedir     = "Sin medir"
)

// encabezadosDeLaTablaSinMedir son los de la tabla de la sección «Sesiones sin
// medir» de informe.md (contrato informe-del-job §4 de H7.3).
var encabezadosDeLaTablaSinMedir = []string{"Sesión", "Eval", "Modelo", "Motivo"}

// encabezadosDeLaTablaDeExpresiones son los de la tabla de la sección «Expresiones
// prohibidas por modelo» de informe.md (contrato lista-y-juicio §5 de H7.2).
var encabezadosDeLaTablaDeExpresiones = []string{
	"Modelo", "Respuestas con alguna expresión", "Respuestas en evals que activan la skill",
}

// parrafoDeLaSkillSinLista es lo que dice esa sección cuando la skill no tiene
// lista: un recuento de cero diría que se buscó (research D7 de H7.2).
const parrafoDeLaSkillSinLista = "la skill no tiene lista de expresiones prohibidas"

// encabezadosDeLaTablaDeTasas son los de la tabla de las series de informe.md,
// con las formas exigidas junto a la tasa (contrato evals-y-skill §6 de H7.1).
var encabezadosDeLaTablaDeTasas = []string{
	"Eval", "Modelo", "Decide", "Planificada", "Formas exigidas", "Tasa", "Resultado",
}

// sinFormasExigidas son las formas exigidas de la serie de una eval que no espera
// hallazgos: una lista vacía, no nil, igual que la que se lee de informe.json.
var sinFormasExigidas = []string{}

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
// leer; el reparto en series con su tasa y su umbral; los motivos de la raíz en
// su orden, que son exactamente las causas del fallo; y el veredicto, que falla
// con un fichero mal formado, con una serie que decide y no llega al umbral o a la
// que le faltan sesiones, con una sesión ilegible o con una petición llegada a la
// red, y no con una invocación fuera de lo grabado, con una serie informativa que
// no pasa ni con una sesión que no pasa de una serie que sí llega al umbral
// (FR-071, FR-073, FR-076, SC-003, SC-012; ADR 0016).
//
// Desde H7, cada sesión de informe.json lleva la clave comandos_prohibidos_ejecutados,
// una lista vacía cuando no hay ninguno, y la tabla de las sesiones de informe.md,
// su columna detrás de los comandos ausentes (contrato evals-y-skill §2 de H7); lo
// que llevan cuando hay uno lo fija TestInformeConProhibidos.
//
// Desde H7.2, ninguna carpeta de evals de los casos tiene lista de expresiones
// prohibidas: cada informe publica lo de una skill sin lista
// (exigirSinListaDeExpresiones); lo que publica una con lista lo fijan
// TestInformeConExpresionesProhibidas, TestInformeConExpresionesEnUnaSerie y
// TestInformeConListaMalFormada.
//
// Desde H7.3, sin lista ni objetivo de duración, ningún caso tiene umbrales
// (exigirSinUmbrales); los que tiene una skill con lista o con objetivo los
// fijan TestUmbralesDelInforme y TestInformeMarkdownDeLosUmbrales.
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
			exigirSinListaDeExpresiones(t, leido)
			exigirSinUmbrales(t, leido)
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

// TestInformeConAvisos fija lo que EscribirInforme publica de los avisos de cada
// sesión (contrato de formato, juicio e informe §5 y §6 de H5.1; research D9 y
// D10): sobre una copia del caso aprobado en la que la eval del art. 21 espera
// derogada y vigencia-agotada y la respuesta de su sesión solo lleva la forma fija
// de derogada, informe.json reparte los dos avisos en el orden de la eval, con el
// motivo del ausente detrás del de la cita ausente en la sesión y en la raíz, y
// con listas vacías en la sesión de la eval que no espera avisos; e informe.md
// los pone en la tabla de las sesiones, junto a las citas, y sigue publicando la
// respuesta (FR-040 a FR-043, SC-004). Nada se escribe bajo testdata/.
func TestInformeConAvisos(t *testing.T) {
	t.Parallel()

	const motivoDeVigenciaAgotada = "aviso ausente: vigencia-agotada"

	casos := []struct {
		nombre string

		// sinLaCita dice si la copia quita además la cita de la respuesta.
		sinLaCita bool

		// respuesta es la de la sesión del art. 21 en la copia; citasAusentes, la
		// celda de su fila en la tabla de las sesiones; y motivos, los de su
		// resultado, en su orden.
		respuesta     string
		citasAusentes string
		motivos       []string
	}{
		{
			nombre:        "uno-encontrado-y-otro-ausente",
			respuesta:     prefijoDeDerogada + respuestaConCita,
			citasAusentes: "ninguna",
			motivos:       []string{motivoDeVigenciaAgotada},
		},
		{
			nombre:        "aviso-detras-de-la-cita",
			sinLaCita:     true,
			respuesta:     prefijoDeDerogada + strings.TrimSuffix(respuestaConCita, citaDeLaRespuesta),
			citasAusentes: textoDeLaCita21,
			motivos:       []string{"cita ausente: " + textoDeLaCita21, motivoDeVigenciaAgotada},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			copia := copiaDelCasoAprobadoConAvisos(t, caso.sinLaCita)

			entradas := entradasDelCaso(casoAprobado, t.TempDir())
			entradas.Evals = filepath.Join(copia, "evals")
			entradas.Sesiones = filepath.Join(copia, "sesiones")

			informe, err := EscribirInforme(entradas)
			require.NoError(t, err)

			leido := leerInformeEscrito(t, entradas.Destino, informe)
			leido.caso = copia

			resultado := resultadoDeLaSesion(t, leido.informe, sesionDelArticulo21)
			assert.Equal(t, caso.respuesta, resultado.Respuesta)
			assert.Equal(t, caso.motivos, resultado.Motivos)
			assert.False(t, resultado.Pasa)

			delArticulo21 := resultadoEscrito(t, leido, sesionDelArticulo21)
			assert.Equal(t, `["derogada"]`, compacto(t, delArticulo21.AvisosEncontrados))
			assert.Equal(t, `["vigencia-agotada"]`, compacto(t, delArticulo21.AvisosAusentes))

			deNoActivacion := resultadoEscrito(t, leido, sesionDeNoActivacion)
			assert.Equal(t, "[]", string(deNoActivacion.AvisosEncontrados))
			assert.Equal(t, "[]", string(deNoActivacion.AvisosAusentes))
			assert.True(t, resultadoDeLaSesion(t, leido.informe, sesionDeNoActivacion).Pasa)

			motivosDeLaRaiz := []string{motivoDeLaTasa(ficheroDeLaEval01, modeloQueDecide, 0, 1, 1)}
			for _, motivo := range caso.motivos {
				motivosDeLaRaiz = append(motivosDeLaRaiz, sesionDelArticulo21+": "+motivo)
			}

			exigirMotivosDeLaRaiz(t, leido, motivosDeLaRaiz...)
			assert.Equal(t, VeredictoFallo, leido.informe.Veredicto)

			exigirLineas(t, seccionDelInforme(t, leido.md, "Sesiones"),
				filaDeTabla(encabezadosDeLaTablaDeSesiones...),
				filaDeTabla(slices.Repeat([]string{"---"}, len(encabezadosDeLaTablaDeSesiones))...),
				filaDeTabla(sesionDelArticulo21, ficheroDeLaEval01, modeloQueDecide, "sí", "sí", "sí (código 0)",
					"ninguno", "ninguno", caso.citasAusentes, "derogada", "vigencia-agotada", "ninguno", "ninguno",
					"ninguna", "ninguna", "ninguno", "ninguno", "ninguna", "0", "no", "no pasa"),
				filaDeTabla(sesionDeNoActivacion, ficheroDeNoActivacion, modeloQueDecide, "no", "no", "sí (código 0)",
					"ninguno", "ninguno", "ninguna", "ninguno", "ninguno", "ninguno", "ninguno", "ninguna", "ninguna",
					"ninguno", "ninguno", "ninguna", "0", "no", "pasa"))

			assert.Contains(t, seccionDelInforme(t, leido.md, "Sesión "+sesionDelArticulo21), caso.respuesta,
				"la sección de la sesión publica la respuesta con la forma fija")
		})
	}
}

// copiaDelCasoAprobadoConAvisos copia el caso aprobado de TestInforme en un
// directorio temporal del test y devuelve su ruta. En la copia, la eval del art. 21
// espera derogada y vigencia-agotada, y a la respuesta de su sesión se le antepone
// la forma fija de derogada y, con sinLaCita, se le quita la cita.
func copiaDelCasoAprobadoConAvisos(t *testing.T, sinLaCita bool) string {
	t.Helper()

	return copiaDelCasoAprobadoConLaEval01(t, avisosDeLaEval01, prefijoDeDerogada, sinLaCita)
}

// copiaDelCasoAprobadoConLaEval01 copia el caso aprobado de TestInforme en un
// directorio temporal del test y devuelve su ruta. En la copia, a la eval del
// art. 21 se le añade al final lo dado, a la respuesta de su sesión se le antepone
// el prefijo, si no es vacío, y, con sinLaCita, se le quita la cita. La respuesta
// está dos veces en el transcript, en el mensaje del asistente y en el result, y
// cada cambio exige exactamente esas dos sustituciones (research V20).
func copiaDelCasoAprobadoConLaEval01(t *testing.T, anadidoALaEval, prefijo string, sinLaCita bool) string {
	t.Helper()

	copia := t.TempDir()
	require.NoError(t, os.CopyFS(copia, os.DirFS(filepath.Join(casosDeInforme, casoAprobado))))

	evals := filepath.Join(copia, "evals")
	eval := contenidoDeLaSesion(t, evals, ficheroDeLaEval01)
	require.True(t, strings.HasSuffix(eval, "\n"), "la eval %s termina en un salto de línea", ficheroDeLaEval01)
	escribirEnLaCopia(t, evals, ficheroDeLaEval01, eval+anadidoALaEval)

	sesion := filepath.Join(copia, "sesiones", sesionDelArticulo21)
	transcript := contenidoDeLaSesion(t, sesion, "sesion.jsonl")

	if prefijo != "" {
		transcript = sustituirDosVeces(t, transcript, cadenaJSON(t, respuestaConCita),
			cadenaJSON(t, prefijo+respuestaConCita))
	}

	if sinLaCita {
		transcript = sustituirDosVeces(t, transcript, citaDeLaRespuesta, "")
	}

	escribirEnLaCopia(t, sesion, "sesion.jsonl", transcript)

	return copia
}

// Lo que TestInformeConProhibidos cambia en su copia del caso aprobado (contrato
// evals-y-skill §2 de H7).
const (
	// prohibidoDeLaEval01 es lo que se añade al final de la eval del art. 21: graph
	// show prohibido.
	prohibidoDeLaEval01 = "prohibidos:\n  - applet: graph\n    verbo: show\n"

	// creacionDeLaLectura es la línea de la traza de claude de la sesión del
	// art. 21 que crea el proceso que lee el bloque; la copia le añade detrás la
	// misma línea con el proceso 3000.
	creacionDeLaLectura = "clone(child_stack=NULL, flags=CLONE_CHILD_CLEARTID|CLONE_CHILD_SETTID|SIGCHLD, " +
		"child_tidptr=0x7f3a9c2f5a10) = 2000\n"

	// trazaDeGraphShow es el t.3000 de la copia: el proceso que pide con graph
	// show la ficha del bloque y termina con código 3.
	trazaDeGraphShow = `execve("/usr/local/bin/kitlegal", ["kitlegal", "graph", "show", ` +
		`"eli/es/l/2015/10/01/39#a21", "--json"], 0x7ffd8f13a6c0 /* 25 vars */) = 0` + "\n" +
		"+++ exited with 3 +++\n"
)

// TestInformeConProhibidos fija la clave comandos_prohibidos_ejecutados de
// informe.json y su columna de informe.md (contrato evals-y-skill §2 de H7;
// FR-085, FR-086): sobre una copia del caso aprobado en la que la eval del art. 21
// prohíbe graph show y su sesión lo ejecuta, aunque termine con código 3,
// informe.json lo lleva en esa sesión, con su motivo en la sesión y en la raíz, y
// una lista vacía en la sesión de la eval que no prohíbe nada; la invocación va
// además a las otras fallidas; e informe.md lo pone en la tabla de las sesiones,
// detrás de los comandos ausentes, y publica la invocación con su código en la
// sección de la sesión. Nada se escribe bajo testdata/.
func TestInformeConProhibidos(t *testing.T) {
	t.Parallel()

	const motivoDeGraphShow = "comando prohibido ejecutado: " + textoDeGraphShow

	copia := copiaDelCasoAprobadoConProhibido(t)

	entradas := entradasDelCaso(casoAprobado, t.TempDir())
	entradas.Evals = filepath.Join(copia, "evals")
	entradas.Sesiones = filepath.Join(copia, "sesiones")

	informe, err := EscribirInforme(entradas)
	require.NoError(t, err)

	leido := leerInformeEscrito(t, entradas.Destino, informe)
	leido.caso = copia

	resultado := resultadoDeLaSesion(t, leido.informe, sesionDelArticulo21)
	assert.Equal(t, []string{textoDeGraphShow}, resultado.ComandosProhibidosEjecutados)
	assert.Equal(t, []string{motivoDeGraphShow}, resultado.Motivos)
	assert.Equal(t, []InvocacionFallida{{Orden: ordenDeGraphShow, Codigo: 3}}, resultado.OtrasFallidas)
	assert.False(t, resultado.Pasa)

	assert.Equal(t, `["graph show"]`,
		compacto(t, resultadoEscrito(t, leido, sesionDelArticulo21).ComandosProhibidosEjecutados))
	assert.Equal(t, "[]", string(resultadoEscrito(t, leido, sesionDeNoActivacion).ComandosProhibidosEjecutados))
	assert.True(t, resultadoDeLaSesion(t, leido.informe, sesionDeNoActivacion).Pasa)

	exigirMotivosDeLaRaiz(t, leido, motivoDeLaTasa(ficheroDeLaEval01, modeloQueDecide, 0, 1, 1),
		sesionDelArticulo21+": "+motivoDeGraphShow)
	assert.Equal(t, VeredictoFallo, leido.informe.Veredicto)

	exigirLineas(t, seccionDelInforme(t, leido.md, "Sesiones"),
		filaDeTabla(encabezadosDeLaTablaDeSesiones...),
		filaDeTabla(sesionDelArticulo21, ficheroDeLaEval01, modeloQueDecide, "sí", "sí", "sí (código 0)",
			"ninguno", textoDeGraphShow, "ninguna", "ninguno", "ninguno", "ninguno", "ninguno", "ninguna", "ninguna",
			"ninguno", "ninguno", "ninguna", "0", "no", "no pasa"),
		filaDeTabla(sesionDeNoActivacion, ficheroDeNoActivacion, modeloQueDecide, "no", "no", "sí (código 0)",
			"ninguno", "ninguno", "ninguna", "ninguno", "ninguno", "ninguno", "ninguno", "ninguna", "ninguna",
			"ninguno", "ninguno", "ninguna", "0", "no", "pasa"))

	exigirLineas(t, seccionDelInforme(t, leido.md, "Sesión "+sesionDelArticulo21),
		filaDeTabla(ordenDeGraphShow, "3", "sin conexiones"))
}

// copiaDelCasoAprobadoConProhibido copia el caso aprobado de TestInforme en un
// directorio temporal del test y devuelve su ruta. En la copia, la eval del
// art. 21 prohíbe graph show, y la traza de su sesión tiene un proceso más, el
// 3000, que claude crea detrás del que lee el bloque y que pide con graph show la
// ficha del bloque y termina con código 3.
func copiaDelCasoAprobadoConProhibido(t *testing.T) string {
	t.Helper()

	copia := copiaDelCasoAprobadoConLaEval01(t, prohibidoDeLaEval01, "", false)

	traza := filepath.Join(copia, "sesiones", sesionDelArticulo21, directorioDeLaTraza)
	deClaude := contenidoDeLaSesion(t, traza, "t.1000")
	require.Equal(t, 1, strings.Count(deClaude, creacionDeLaLectura),
		"la traza de claude crea una sola vez el proceso que lee el bloque")

	creacionDeGraphShow := strings.Replace(creacionDeLaLectura, "= 2000", "= 3000", 1)
	escribirEnLaCopia(t, traza, "t.1000",
		strings.Replace(deClaude, creacionDeLaLectura, creacionDeLaLectura+creacionDeGraphShow, 1))
	escribirEnLaCopia(t, traza, "t.3000", trazaDeGraphShow)

	return copia
}

// Lo que TestInformeConHallazgos cambia en su copia del caso aprobado y la forma
// que espera leer en el informe (contrato evals-y-skill §6 de H7.1; research D14).
const (
	// hallazgosDeLaEval01 es lo que se añade al final de la eval del art. 21: el
	// hallazgo version-obsoleta, cuya forma fija tiene que llevar la respuesta.
	hallazgosDeLaEval01 = "hallazgos:\n  - version-obsoleta\n"

	// formaDeVersionObsoleta es la forma fija que el informe declara que exige esa
	// eval: la marca, un espacio, la etiqueta del binario y los dos puntos.
	formaDeVersionObsoleta = "⚠ REDACCIÓN MODIFICADA:"

	// redaccionesModificadasDeLaEval20 es lo que se añade al final de una eval
	// para que espere las dos redacciones modificadas de la eval 20
	// (contracts/evals-y-juicio.md §3 de H7.4).
	redaccionesModificadasDeLaEval20 = "redacciones_modificadas:\n" +
		"  - norma: BOE-A-2017-12902\n    bloque: a1-30\n" +
		"    fecha_vigencia: \"20180309\"\n    fecha_vigencia_reciente: \"20200206\"\n" +
		"  - norma: BOE-A-2017-12902\n    bloque: da-3\n" +
		"    fecha_vigencia: \"20180309\"\n    fecha_vigencia_reciente: \"20230101\"\n"
)

// TestInformeConHallazgos fija lo que EscribirInforme publica de los hallazgos
// (contrato evals-y-skill §6 de H7.1; research D14; FR-055, SC-006): sobre una
// copia del caso aprobado en la que la eval del art. 21 espera version-obsoleta,
// la serie de esa eval declara en formas la forma fija literal que exige junto a
// su tasa, y la de la eval sin hallazgos, una lista vacía, no null; cada sesión
// reparte el hallazgo entre encontrados y ausentes según lleve o no la forma, con
// el motivo del ausente en la sesión y en la raíz; e informe.md pone la forma en
// la columna «Formas exigidas» de la tabla de las series, junto a la tasa, y los
// hallazgos en «Hallazgos encontrados» y «Hallazgos ausentes», en la de las
// sesiones, detrás de los avisos. Nada se escribe bajo testdata/.
func TestInformeConHallazgos(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string

		// prefijo es lo que la copia antepone a la respuesta de la sesión del
		// art. 21: el traslado del cambio con su forma fija, o nada.
		prefijo string

		// pasa dice si la sesión del art. 21 pasa, y con ella su serie; tasa y
		// umbral son las celdas de esa serie en la tabla de las series; y
		// encontrados, ausentes y resultado, las de esa sesión en la de las
		// sesiones.
		pasa        bool
		tasa        string
		umbral      string
		encontrados string
		ausentes    string
		resultado   string
	}{
		{
			nombre:      "forma-encontrada",
			prefijo:     trasladoDelCambio + " ",
			pasa:        true,
			tasa:        "1 de 1",
			umbral:      "llega al umbral",
			encontrados: versionObsoleta,
			ausentes:    "ninguno",
			resultado:   "pasa",
		},
		{
			nombre:      "forma-ausente",
			tasa:        "0 de 1",
			umbral:      "no llega al umbral",
			encontrados: "ninguno",
			ausentes:    versionObsoleta,
			resultado:   "no pasa",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			copia := copiaDelCasoAprobadoConLaEval01(t, hallazgosDeLaEval01, caso.prefijo, false)

			entradas := entradasDelCaso(casoAprobado, t.TempDir())
			entradas.Evals = filepath.Join(copia, "evals")
			entradas.Sesiones = filepath.Join(copia, "sesiones")

			informe, err := EscribirInforme(entradas)
			require.NoError(t, err)

			leido := leerInformeEscrito(t, entradas.Destino, informe)
			leido.caso = copia

			pasan := 0
			if caso.pasa {
				pasan = 1
			}

			assert.Equal(t, TasaDelInforme{
				Eval: ficheroDeLaEval01, Modelo: modeloQueDecide, Planificada: true, Decide: true,
				Formas: []string{formaDeVersionObsoleta}, Sesiones: 1, Pasan: pasan, Pasa: caso.pasa,
			}, tasaDeLaSerie(t, leido.informe, ficheroDeLaEval01, modeloQueDecide))
			assert.Equal(t, sinFormasExigidas,
				tasaDeLaSerie(t, leido.informe, ficheroDeNoActivacion, modeloQueDecide).Formas)
			assert.Equal(t, map[string]string{
				ficheroDeLaEval01:     `["` + formaDeVersionObsoleta + `"]`,
				ficheroDeNoActivacion: "[]",
			}, formasEscritas(t, leido), "formas de cada serie tal como está escrita en informe.json")

			resultado := resultadoDeLaSesion(t, leido.informe, sesionDelArticulo21)
			assert.Equal(t, caso.pasa, resultado.Pasa)

			delArticulo21 := resultadoEscrito(t, leido, sesionDelArticulo21)
			deNoActivacion := resultadoEscrito(t, leido, sesionDeNoActivacion)
			assert.Equal(t, "[]", string(deNoActivacion.HallazgosEncontrados))
			assert.Equal(t, "[]", string(deNoActivacion.HallazgosAusentes))

			if caso.pasa {
				assert.Equal(t, `["`+versionObsoleta+`"]`, compacto(t, delArticulo21.HallazgosEncontrados))
				assert.Equal(t, "[]", string(delArticulo21.HallazgosAusentes))
				assert.Empty(t, resultado.Motivos)
				exigirMotivosDeLaRaiz(t, leido)
				assert.Equal(t, VeredictoAprobado, leido.informe.Veredicto)
			} else {
				assert.Equal(t, "[]", string(delArticulo21.HallazgosEncontrados))
				assert.Equal(t, `["`+versionObsoleta+`"]`, compacto(t, delArticulo21.HallazgosAusentes))
				assert.Equal(t, []string{motivoDeVersionObsoleta}, resultado.Motivos)
				exigirMotivosDeLaRaiz(t, leido, motivoDeLaTasa(ficheroDeLaEval01, modeloQueDecide, 0, 1, 1),
					sesionDelArticulo21+": "+motivoDeVersionObsoleta)
				assert.Equal(t, VeredictoFallo, leido.informe.Veredicto)
			}

			exigirLineas(t, seccionDelInforme(t, leido.md, "Tasas por eval"),
				filaDeTabla(encabezadosDeLaTablaDeTasas...),
				filaDeTabla(slices.Repeat([]string{"---"}, len(encabezadosDeLaTablaDeTasas))...),
				filaDeTabla(ficheroDeLaEval01, modeloQueDecide, "sí", "sí", formaDeVersionObsoleta, caso.tasa,
					caso.umbral),
				filaDeTabla(ficheroDeNoActivacion, modeloQueDecide, "sí", "sí", "ninguna", "1 de 1",
					"llega al umbral"))

			exigirLineas(t, seccionDelInforme(t, leido.md, "Sesiones"),
				filaDeTabla(encabezadosDeLaTablaDeSesiones...),
				filaDeTabla(sesionDelArticulo21, ficheroDeLaEval01, modeloQueDecide, "sí", "sí", "sí (código 0)",
					"ninguno", "ninguno", "ninguna", "ninguno", "ninguno", caso.encontrados, caso.ausentes, "ninguna",
					"ninguna", "ninguno", "ninguno", "ninguna", "0", "no", caso.resultado),
				filaDeTabla(sesionDeNoActivacion, ficheroDeNoActivacion, modeloQueDecide, "no", "no", "sí (código 0)",
					"ninguno", "ninguno", "ninguna", "ninguno", "ninguno", "ninguno", "ninguno", "ninguna", "ninguna",
					"ninguno", "ninguno", "ninguna", "0", "no", "pasa"))

			assert.Contains(t, seccionDelInforme(t, leido.md, "Sesión "+sesionDelArticulo21),
				caso.prefijo+respuestaConCita, "la sección de la sesión publica la respuesta")
		})
	}
}

// formasEscritas son, por eval, las formas exigidas de cada serie tal como están
// escritas en informe.json, sin blancos.
func formasEscritas(t *testing.T, leido informeLeido) map[string]string {
	t.Helper()

	formas := map[string]string{}
	for _, tasa := range leido.crudo.Tasas {
		formas[tasa.Eval] = compacto(t, tasa.Formas)
	}

	return formas
}

// Lo que TestInformeConExpresionesProhibidas, TestInformeConExpresionesEnUnaSerie
// y TestInformeConListaMalFormada ponen en sus copias de los casos de TestInforme
// (contrato lista-y-juicio §5 de H7.2).
const (
	// sesionDelArticulo21ConOpus y sesionDeNoActivacionConOpus son las sesiones
	// de las dos evals del caso aprobado con modeloInformativoDelCaso.
	sesionDelArticulo21ConOpus  = "01-lpac-articulo-21-claude-opus-5-01"
	sesionDeNoActivacionConOpus = "11-no-activa-programacion-claude-opus-5-01"

	// jsonEnLaRespuesta es lo que se antepone a la respuesta de la sesión de la
	// eval de no activación: una frase de programación con json, que está en la
	// lista y que esa eval no juzga (FR-052).
	jsonEnLaRespuesta = "Si la lista llega en JSON, primero hay que leerla. "

	// listaSinOtraConversacion es una lista de expresiones prohibidas mal formada:
	// le falta la familia otra_conversacion, que el esquema exige; las otras dos
	// familias de H7.3 y las dos claves de H7.4 están bien formadas.
	listaSinOtraConversacion = "maquinaria:\n  - memoria de consultas\n  - hallazgos\n" +
		"anuncio:\n  - ya puedo responder\n" +
		"redaccion_no_leida:\n  - ya no exige\n" +
		"formas_fijas:\n  - No se ha podido comprobar si la redacci\xc3\xb3n ha cambiado\n"
)

// expresionesDeLaSesion son las expresiones prohibidas que el informe publica de
// una sesión, en el orden de la lista; nil si no lleva ninguna.
type expresionesDeLaSesion struct {
	sesion      string
	encontradas []string
}

// TestInformeConExpresionesProhibidas fija lo que EscribirInforme publica de las
// expresiones prohibidas (contrato lista-y-juicio §5 de H7.2; research D7;
// FR-053, SC-001; US3.7, US3.8): sobre una copia del caso aprobado con la lista
// del repositorio y con modeloInformativoDelCaso entre los modelos informativos,
// cuya sesión del art. 21 con el modelo que decide lleva las de la transición de
// la memoria —dos de la maquinaria y una del anuncio—, la del art. 21 con el
// otro modelo ninguna y las dos de la eval de
// no activación json, informe.json publica en cada sesión las que lleva —las de
// la eval de no activación no se juzgan: lista vacía— y la tabla de las sesiones
// de informe.md, en su columna; el recuento por modelo, primero el que decide,
// cuenta las respuestas de las evals que activan la skill y las que llevan
// alguna, en informe.json y en su tabla de informe.md; y la sesión con
// expresiones no pasa y su serie, que decide, da el veredicto fallo, como una
// cita ausente (FR-054). La sesión de la prueba de red, con lo dicho en otra
// conversación, se juzga con la lista y publica su expresión, pero no cambia el
// recuento ni los motivos de la raíz; y una sesión ilegible queda fuera del
// recuento. Sin la lista, las mismas respuestas no llevan ninguna: el informe
// publica lo de una skill sin lista. Nada se escribe bajo testdata/.
func TestInformeConExpresionesProhibidas(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre      string
		pruebaDeRed bool
	}{
		{nombre: "por-sesion-y-por-modelo"},
		{nombre: "con-la-prueba-de-red", pruebaDeRed: true},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()
			exigirExpresionesDelCaso(t, caso.pruebaDeRed)
		})
	}

	t.Run("con-una-sesion-ilegible", func(t *testing.T) {
		t.Parallel()
		exigirExpresionesConUnaSesionIlegible(t)
	})

	t.Run("sin-lista", func(t *testing.T) {
		t.Parallel()

		leido := informeDeLaCopia(t, copiaConExpresiones(t, false, true), conElModeloInformativo)

		exigirSinListaDeExpresiones(t, leido)
		assert.True(t, resultadoDeLaSesion(t, leido.informe, sesionDelArticulo21).Pasa,
			"sin lista, la respuesta con expresiones pasa")
		exigirMotivosDeLaRaiz(t, leido)
		assert.Equal(t, VeredictoAprobado, leido.informe.Veredicto)
	})
}

// exigirExpresionesDelCaso escribe el informe de la copia con la lista de
// TestInformeConExpresionesProhibidas, con la sesión de la prueba de red o sin
// ella, y exige lo que publica: las expresiones de cada sesión, el recuento por
// modelo, igual en los dos casos, y los motivos de la raíz, que son los de la
// serie del art. 21 con el modelo que decide y, detrás, el de su umbral de las
// expresiones, que con 1 de 1 no se cumple (contrato informe-del-job §2 de H7.3).
func exigirExpresionesDelCaso(t *testing.T, pruebaDeRed bool) {
	t.Helper()

	leido := informeDeLaCopia(t, copiaConExpresiones(t, true, pruebaDeRed), conElModeloInformativo)

	esperadas := []expresionesDeLaSesion{
		{sesion: sesionDelArticulo21, encontradas: expresionesDeLaTransicion},
		{sesion: sesionDelArticulo21ConOpus},
		{sesion: sesionDeNoActivacion},
		{sesion: sesionDeNoActivacionConOpus},
	}

	if pruebaDeRed {
		esperadas = slices.Insert(esperadas, 2, expresionesDeLaSesion{
			sesion: sesionDeLaPruebaDeRed, encontradas: []string{"te confirmé"},
		})

		deLaPruebaDeRed := resultadoDeLaSesion(t, leido.informe, sesionDeLaPruebaDeRed)
		assert.False(t, deLaPruebaDeRed.Pasa, "la sesión de la prueba de red se juzga con la lista")
		assert.Contains(t, deLaPruebaDeRed.Motivos, "expresión prohibida: te confirmé")
	}

	exigirExpresionesPorSesion(t, leido, esperadas)

	exigirRecuento(t, leido, []RecuentoDeExpresiones{
		{Modelo: modeloQueDecide, ConAlguna: 1, Respuestas: 1},
		{Modelo: modeloInformativoDelCaso, ConAlguna: 0, Respuestas: 1},
	}, `[{"modelo":"`+modeloQueDecide+`","con_alguna":1,"respuestas":1},`+
		`{"modelo":"`+modeloInformativoDelCaso+`","con_alguna":0,"respuestas":1}]`)

	exigirMotivosDeLaRaiz(t, leido, slices.Concat(
		[]string{motivoDeLaTasa(ficheroDeLaEval01, modeloQueDecide, 0, 1, 1)},
		motivosDeLaTransicion(sesionDelArticulo21),
		[]string{"umbral expresiones_prohibidas:" + modeloQueDecide + ": 1 de 1 (100,0 %), y tiene que ser ≤ 5,0 %"})...)
	assert.Equal(t, VeredictoFallo, leido.informe.Veredicto)
}

// expresionesDeLaTransicion son las expresiones de la lista del repositorio que
// lleva transicionDeLaMemoria, en el orden de la lista: dos de la maquinaria y
// una del anuncio.
var expresionesDeLaTransicion = []string{"memoria de consultas", "hallazgos", "tengo todo lo necesario"}

// motivosDeLaTransicion son los motivos de la raíz que da la sesión dada por las
// expresiones de la transición de la memoria que lleva su respuesta, en su orden.
func motivosDeLaTransicion(sesion string) []string {
	motivos := make([]string, 0, len(expresionesDeLaTransicion))
	for _, expresion := range expresionesDeLaTransicion {
		motivos = append(motivos, sesion+": expresión prohibida: "+expresion)
	}

	return motivos
}

// exigirExpresionesConUnaSesionIlegible escribe el informe de la copia con la
// lista de TestInformeConExpresionesProhibidas en la que la sesión del art. 21
// con el modelo que decide no tiene codigo-de-la-sesion, y exige que, ilegible,
// quede sin juzgar y fuera del recuento: no tiene respuesta juzgada, y ya da su
// motivo en la raíz.
func exigirExpresionesConUnaSesionIlegible(t *testing.T) {
	t.Helper()

	copia := copiaConExpresiones(t, true, false)
	require.NoError(t, os.Remove(filepath.Join(copia, "sesiones", sesionDelArticulo21, "codigo-de-la-sesion")))

	leido := informeDeLaCopia(t, copia, conElModeloInformativo)

	exigirExpresionesPorSesion(t, leido, []expresionesDeLaSesion{
		{sesion: sesionDelArticulo21},
		{sesion: sesionDelArticulo21ConOpus},
		{sesion: sesionDeNoActivacion},
		{sesion: sesionDeNoActivacionConOpus},
	})

	exigirRecuento(t, leido, []RecuentoDeExpresiones{
		{Modelo: modeloQueDecide, ConAlguna: 0, Respuestas: 0},
		{Modelo: modeloInformativoDelCaso, ConAlguna: 0, Respuestas: 1},
	}, `[{"modelo":"`+modeloQueDecide+`","con_alguna":0,"respuestas":0},`+
		`{"modelo":"`+modeloInformativoDelCaso+`","con_alguna":0,"respuestas":1}]`)

	motivo := exigirSesionIlegible(t, leido, "codigo-de-la-sesion")
	exigirMotivosDeLaRaiz(t, leido, motivoDeLaTasa(ficheroDeLaEval01, modeloQueDecide, 0, 1, 1),
		sesionDelArticulo21+": "+motivo)
}

// TestInformeConExpresionesEnUnaSerie fija que las expresiones prohibidas no
// cambian la regla por serie (FR-054; contrato lista-y-juicio §5 de H7.2;
// ADR 0016; FR-007 de H7.3): con tres repeticiones y umbral 2, en una copia del
// caso aprobado con la lista del repositorio y solo la eval del art. 21, cuya
// serie tiene tres sesiones que pasan salvo porque dos llevan alguna expresión, la
// serie de la eval que decide no llega al umbral, con su tasa y los motivos de las
// dos sesiones; y la misma serie de la eval informativa publica su tasa sin
// decidir y sin motivos. En los dos, el recuento del modelo que decide cuenta las
// tres respuestas y las dos con alguna —las evals informativas también activan la
// skill—, y su umbral de las expresiones, con 2 de 3, no se cumple: su motivo va
// detrás de los de la serie y el veredicto es fallo también con la eval
// informativa (FR-002 y FR-003 de H7.3). Nada se escribe bajo testdata/.
func TestInformeConExpresionesEnUnaSerie(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre      string
		informativa bool
	}{
		{nombre: "eval-que-decide"},
		{nombre: "eval-informativa", informativa: true},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()
			exigirLaSerieConExpresiones(t, caso.informativa)
		})
	}
}

// exigirLaSerieConExpresiones escribe el informe de la serie de
// TestInformeConExpresionesEnUnaSerie, con la eval del art. 21 informativa o no, y
// exige su tasa, las expresiones de cada sesión, el recuento por modelo, los
// motivos de la raíz y el veredicto.
func exigirLaSerieConExpresiones(t *testing.T, informativa bool) {
	t.Helper()

	leido := informeDeLaCopia(t, copiaDeLaSerieConExpresiones(t, informativa), func(entradas *InformeAEscribir) {
		entradas.Repeticiones, entradas.Umbral = 3, 2
	})

	decide := "sí"
	if informativa {
		decide = "no"
	}

	assert.Equal(t, TasaDelInforme{
		Eval: ficheroDeLaEval01, Modelo: modeloQueDecide, Planificada: true, Decide: !informativa,
		Formas: sinFormasExigidas, Sesiones: 3, Pasan: 1,
	}, tasaDeLaSerie(t, leido.informe, ficheroDeLaEval01, modeloQueDecide))
	exigirLineas(t, seccionDelInforme(t, leido.md, "Tasas por eval"),
		filaDeTabla(ficheroDeLaEval01, modeloQueDecide, decide, "sí", "ninguna", "1 de 3", "no llega al umbral"))

	exigirExpresionesPorSesion(t, leido, []expresionesDeLaSesion{
		{sesion: sesionDeLaSerie(1)},
		{sesion: sesionDeLaSerie(2), encontradas: expresionesDeLaTransicion},
		{sesion: sesionDeLaSerie(3), encontradas: []string{"te confirmé"}},
	})
	exigirRecuento(t, leido, []RecuentoDeExpresiones{{Modelo: modeloQueDecide, ConAlguna: 2, Respuestas: 3}},
		`[{"modelo":"`+modeloQueDecide+`","con_alguna":2,"respuestas":3}]`)

	delUmbral := "umbral expresiones_prohibidas:" + modeloQueDecide + ": 2 de 3 (66,7 %), y tiene que ser ≤ 5,0 %"

	if informativa {
		exigirMotivosDeLaRaiz(t, leido, delUmbral)
		assert.Equal(t, VeredictoFallo, leido.informe.Veredicto)

		return
	}

	exigirMotivosDeLaRaiz(t, leido, slices.Concat(
		[]string{motivoDeLaTasa(ficheroDeLaEval01, modeloQueDecide, 1, 3, 2)},
		motivosDeLaTransicion(sesionDeLaSerie(2)),
		[]string{sesionDeLaSerie(3) + ": expresión prohibida: te confirmé", delUmbral})...)
	assert.Equal(t, VeredictoFallo, leido.informe.Veredicto)
}

// sesionDeLaSerie es el nombre de la sesión de ese número de la serie de
// TestInformeConExpresionesEnUnaSerie: la eval del art. 21 con el modelo que
// decide, como los nombra el plan.
func sesionDeLaSerie(numero int) string {
	return fmt.Sprintf("%s-%s-%02d", sesionDelArticulo21, modeloQueDecide, numero)
}

// TestInformeConListaMalFormada fija lo que el informe hace con una lista de
// expresiones prohibidas que no valida (FR-055; research D8 de H7.2; US3.9):
// sobre una copia del caso aprobado cuya lista no tiene la familia
// otra_conversacion y cuya sesión del art. 21 lleva dos expresiones de la
// maquinaria de esa lista, la lista queda en ficheros_mal_formados con el error de
// LeerConjunto, que es el único motivo de la raíz, y el veredicto es fallo; las
// evals se juzgan sin lista —la sesión pasa— y el informe publica lo de una skill
// sin lista. Nada se escribe bajo testdata/.
func TestInformeConListaMalFormada(t *testing.T) {
	t.Parallel()

	copia := copiaDelCasoAprobadoConLaEval01(t, "", transicionDeLaMemoria+"\n\n", false)
	escribirEnLaCopia(t, filepath.Join(copia, "evals"), ficheroDeExpresionesProhibidas, listaSinOtraConversacion)

	conjunto, err := LeerConjunto(filepath.Join(copia, "evals"))
	require.NoError(t, err)
	require.Len(t, conjunto.MalFormados, 1)
	require.Equal(t, ficheroDeExpresionesProhibidas, conjunto.MalFormados[0].Fichero)

	errorDeLaLista := conjunto.MalFormados[0].Error.Error()

	leido := informeDeLaCopia(t, copia, nil)

	assert.Equal(t, []FicheroMalFormadoDelInforme{{Fichero: ficheroDeExpresionesProhibidas, Error: errorDeLaLista}},
		leido.informe.FicherosMalFormados)
	exigirLineas(t, seccionDelInforme(t, leido.md, "Ficheros mal formados"),
		"- "+ficheroDeExpresionesProhibidas+": "+errorDeLaLista)
	exigirMotivosDeLaRaiz(t, leido, ficheroDeExpresionesProhibidas+": mal formado: "+errorDeLaLista)
	assert.Equal(t, VeredictoFallo, leido.informe.Veredicto)

	assert.True(t, resultadoDeLaSesion(t, leido.informe, sesionDelArticulo21).Pasa,
		"las evals de la carpeta se juzgan sin lista")
	exigirSinListaDeExpresiones(t, leido)
}

// encabezadosDeLaTablaDeUmbrales son los de la tabla de la sección «Umbrales» de
// informe.md (contrato informe-del-job §4 de H7.3).
var encabezadosDeLaTablaDeUmbrales = []string{"Umbral", "Medida", "Condición", "Cumple", "Hace fallar el veredicto"}

// TestInformeMarkdownDeLosUmbrales fija la sección «Umbrales» de informe.md
// (contrato informe-del-job §4 de H7.3; FR-001; US2-6), sobre ejecuciones
// sintéticas de ejecucionConUmbrales: junto al recuento de las expresiones
// prohibidas por modelo —detrás de él y delante de las sesiones sin medir, lo
// que leerInformeEscrito exige en todo informe—, una tabla con una fila por
// umbral, en su orden: el nombre, la medida —«<medida> de <total> (<p> %)» con un
// decimal y coma, o la medida sola—, la condición, si se cumple y si hace fallar
// el veredicto, con «no: solo se publica» en los que no deciden. La del contrato,
// fila a fila, con 2 de 51, 0 de 30 y 544 s; la de los tres sin cumplir; y, sin
// ninguno, el párrafo «ninguno». La cabecera lleva la duración de las sesiones.
//
// Desde H7.4, fija además las formas exigidas y la columna de cada lado de las
// redacciones modificadas de contracts/informe-del-job.md §4 de H7.4: con evals
// que esperan version-obsoleta y las dos redacciones modificadas de la eval 20,
// la columna «Formas exigidas» de cada serie de «Tasas por eval» lleva, detrás
// de la forma del hallazgo, «⚠ REDACCIÓN MODIFICADA: <texto>» por cada
// redacción, en su orden; y la tabla de las sesiones, detrás de «Hallazgos
// ausentes», «Redacciones modificadas encontradas» y «Redacciones modificadas
// ausentes», con la del art. 118, que cada respuesta traslada con su línea, en
// la primera, y la de la disposición adicional tercera, que no traslada ninguna,
// en la segunda. Las líneas con su cita no cuentan como expresión prohibida. Sin
// redacciones esperadas, las tres celdas dicen «ninguna».
func TestInformeMarkdownDeLosUmbrales(t *testing.T) {
	t.Parallel()

	comoElJob := ejecucionConUmbrales{queDeciden: 10, informativas: 7, conHaiku: true}

	casos := []struct {
		nombre    string
		ejecucion ejecucionConUmbrales

		// filas son las de la tabla, sin los encabezados; nil, sin tabla.
		filas [][]string

		// formas es la celda «Formas exigidas» de cada serie de «Tasas por eval»;
		// encontradas y ausentes, las celdas de las redacciones modificadas de cada
		// sesión de «Sesiones». Vacías, «ninguna».
		formas, encontradas, ausentes string
	}{
		{
			nombre: "las-filas-del-contrato",
			ejecucion: conCambios(comoElJob, func(e *ejecucionConUmbrales) {
				e.conAlguna, e.duracion, e.objetivo = map[string]int{modeloSonnet5: 2}, 544, 900
			}),
			filas: [][]string{
				{"`expresiones_prohibidas:claude-sonnet-5`", "2 de 51 (3,9 %)", "≤ 5,0 %", "sí", "sí"},
				{"`expresiones_prohibidas:claude-haiku-4-5-20251001`", "0 de 30 (0,0 %)", "≤ 5,0 %", "sí", "no: solo se publica"},
				{"`duracion_de_las_sesiones`", "544", "≤ 900", "sí", "sí"},
			},
		},
		{
			nombre: "los-tres-sin-cumplir",
			ejecucion: conCambios(comoElJob, func(e *ejecucionConUmbrales) {
				e.conAlguna, e.duracion, e.objetivo = map[string]int{modeloSonnet5: 3, modeloHaiku45: 2}, 901, 900
			}),
			filas: [][]string{
				{"`expresiones_prohibidas:claude-sonnet-5`", "3 de 51 (5,9 %)", "≤ 5,0 %", "no", "sí"},
				{"`expresiones_prohibidas:claude-haiku-4-5-20251001`", "2 de 30 (6,7 %)", "≤ 5,0 %", "no", "no: solo se publica"},
				{"`duracion_de_las_sesiones`", "901", "≤ 900", "no", "sí"},
			},
		},
		{
			nombre:    "sin-umbrales",
			ejecucion: ejecucionConUmbrales{queDeciden: 1, sinLista: true, duracion: 544},
		},
		{
			nombre: "con-redacciones-modificadas",
			ejecucion: ejecucionConUmbrales{
				queDeciden: 1, informativas: 1, conHaiku: true, duracion: 544,
				anadidoALaEval:         hallazgosDeLaEval01 + redaccionesModificadasDeLaEval20,
				prefijoDeLasRespuestas: lineaConLaCitaDel118() + "\n\n",
			},
			filas: [][]string{
				{"`expresiones_prohibidas:claude-sonnet-5`", "0 de 6 (0,0 %)", "≤ 5,0 %", "sí", "sí"},
				{"`expresiones_prohibidas:claude-haiku-4-5-20251001`", "0 de 3 (0,0 %)", "≤ 5,0 %", "sí", "no: solo se publica"},
			},
			formas: formaDeVersionObsoleta + ", " + formaDeVersionObsoleta + " " + redaccionDelArticulo118 + ", " +
				formaDeVersionObsoleta + " " + redaccionDeLaDA3,
			encontradas: redaccionDelArticulo118,
			ausentes:    redaccionDeLaDA3,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			leido := escribirEjecucionConUmbrales(t, caso.ejecucion)

			exigirLineas(t, seccionDelInforme(t, leido.md, "Cabecera"),
				"Duración de las sesiones: "+strconv.Itoa(caso.ejecucion.duracion)+" s")

			exigirCadaCelda(t, seccionDelInforme(t, leido.md, "Tasas por eval"), encabezadosDeLaTablaDeTasas,
				"Formas exigidas", cmp.Or(caso.formas, "ninguna"))
			exigirCadaCelda(t, seccionDelInforme(t, leido.md, "Sesiones"), encabezadosDeLaTablaDeSesiones,
				columnaDeRedaccionesEncontradas, cmp.Or(caso.encontradas, "ninguna"))
			exigirCadaCelda(t, seccionDelInforme(t, leido.md, "Sesiones"), encabezadosDeLaTablaDeSesiones,
				columnaDeRedaccionesAusentes, cmp.Or(caso.ausentes, "ninguna"))

			if caso.filas == nil {
				exigirSinUmbrales(t, leido)

				return
			}

			filas := []string{
				filaDeTabla(encabezadosDeLaTablaDeUmbrales...),
				filaDeTabla(slices.Repeat([]string{"---"}, len(encabezadosDeLaTablaDeUmbrales))...),
			}
			for _, fila := range caso.filas {
				filas = append(filas, filaDeTabla(fila...))
			}

			assert.Equal(t, strings.Join(filas, "\n"), seccionDelInforme(t, leido.md, "Umbrales"))
		})
	}
}

// exigirSinUmbrales exige lo que publica el informe de una skill sin lista de
// expresiones prohibidas ni objetivo de duración (FR-006 de H7.3): umbrales es
// una lista vacía, no null, y su sección de informe.md, el párrafo «ninguno».
func exigirSinUmbrales(t *testing.T, leido informeLeido) {
	t.Helper()

	assert.Empty(t, leido.informe.Umbrales)
	assert.Equal(t, "[]", compacto(t, leido.crudo.Umbrales), "umbrales es una lista vacía, no null")
	assert.Equal(t, "ninguno", seccionDelInforme(t, leido.md, "Umbrales"))
}

// Las clases con que una sesión queda sin medir tal como las publica el informe,
// en sin_medir de la sesión y en el motivo de sesiones_sin_medir (data-model §3
// de H7.3; contrato informe-del-job §3).
const (
	sinMedirPorElMensaje = "mensaje del límite de uso: " + textoDelLimiteDeSesion
	sinMedirPorAgotados  = "reintentos por rate_limit agotados"
	sinMedirPorElTope    = "cortada por el tope durante reintentos por rate_limit"
	sinAbrirTrasElLimite = "sin abrir tras el límite de uso"
)

// TestInformeConSesionesSinMedir fija lo que EscribirInforme hace con las
// sesiones que un límite de uso de la cuenta no dejó terminar (contrato
// informe-del-job §2.2, §3, §4 y §6 de H7.3; data-model §3 y §4; research D6;
// FR-033, FR-040 a FR-044, FR-093; SC-001, SC-007; US3-2 a US3-4), sobre copias
// del caso aprobado con la lista del repositorio armadas en t.TempDir(), sin
// tocar las versionadas: una sesión con el mensaje del límite de uso (a), con los
// reintentos por rate_limit agotados (b) o cortada por el tope durante ellos (c)
// no pasa ni falla: lleva su clase en sin_medir y como único motivo «sin medir
// por límite de uso: <clase>», queda fuera del recuento de las expresiones y deja
// su serie sin medir, sin el motivo de su tasa. Las sesiones que el repartidor no
// abrió tras el límite cuentan en su serie como sin medir, sin el motivo de las
// sesiones que faltan, que queda para las que faltan por otra causa. Con alguna
// sin medir, el veredicto es fallo con un solo motivo de la ejecución, detrás de
// los de siempre, que nombra el límite y las sesiones. Una sesión que se recupera
// de sus reintentos por rate_limit se mide como cualquier otra y publica sus
// reintentos, en la sesión y, sumados, en la raíz.
func TestInformeConSesionesSinMedir(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string

		// codigo y trasElInit son el código de la sesión del art. 21 y lo que
		// sigue en su transcript al mensaje system/init; clase es aquella por la
		// que queda sin medir, y reintentos, los suyos por rate_limit.
		codigo     int
		trasElInit string
		clase      string
		reintentos int
	}{
		{
			nombre:     "mensaje-del-limite-de-uso",
			codigo:     1,
			trasElInit: mensajeResultConError(t, textoDelLimiteDeSesion),
			clase:      sinMedirPorElMensaje,
		},
		{
			nombre: "reintentos-agotados",
			codigo: 1,
			trasElInit: mensajeDeReintento(9, 10, 429, "rate_limit") + mensajeDeReintento(10, 10, 429, "rate_limit") +
				mensajeResultConError(t, textoDelError429),
			clase:      sinMedirPorAgotados,
			reintentos: 2,
		},
		{
			nombre:     "cortada-durante-reintentos",
			codigo:     124,
			trasElInit: mensajeDeReintento(1, 10, 429, "rate_limit") + mensajeDeReintento(2, 10, 429, "rate_limit"),
			clase:      sinMedirPorElTope,
			reintentos: 2,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()
			exigirUnaSesionSinMedir(t, caso.codigo, caso.trasElInit, caso.clase, caso.reintentos)
		})
	}

	t.Run("sin-abrir-tras-el-limite-de-uso", func(t *testing.T) {
		t.Parallel()
		exigirLasSesionesSinAbrir(t)
	})

	t.Run("reintentos-de-los-que-se-recupera", func(t *testing.T) {
		t.Parallel()
		exigirLosReintentosRecuperados(t)
	})
}

// exigirUnaSesionSinMedir escribe el informe de una copia del caso aprobado con
// la lista del repositorio en la que la sesión del art. 21 termina con el código
// dado y su transcript sigue a su mensaje system/init con lo dado, y exige que
// quede sin medir por la clase dada, con sus reintentos por rate_limit en la
// sesión y en la raíz; que su serie quede sin medir, sin el motivo de su tasa;
// que no entre en el recuento; y el veredicto fallo con el único motivo de la
// ejecución, que la nombra.
func exigirUnaSesionSinMedir(t *testing.T, codigo int, trasElInit, clase string, reintentos int) {
	t.Helper()

	copia := copiaDelCasoAprobadoConLaLista(t)
	cambiarElFinalDeLaSesion(t, filepath.Join(copia, "sesiones", sesionDelArticulo21), codigo, trasElInit)

	leido := informeDeLaCopia(t, copia, nil)

	exigirSesionSinMedir(t, leido, sesionDelArticulo21, clase, reintentos)
	assert.True(t, resultadoDeLaSesion(t, leido.informe, sesionDeNoActivacion).Pasa)

	exigirTasaSinMedir(t, leido, TasaDelInforme{
		Eval: ficheroDeLaEval01, Modelo: modeloQueDecide, Planificada: true, Decide: true,
		Formas: sinFormasExigidas, Sesiones: 1, Pasan: 0, SinMedir: 1,
	})

	exigirRecuento(t, leido, []RecuentoDeExpresiones{{Modelo: modeloQueDecide, ConAlguna: 0, Respuestas: 0}},
		`[{"modelo":"`+modeloQueDecide+`","con_alguna":0,"respuestas":0}]`)

	sinMedir := SesionSinMedir{Sesion: sesionDelArticulo21, Eval: ficheroDeLaEval01, Modelo: modeloQueDecide, Motivo: clase}
	exigirSesionesSinMedir(t, leido, sinMedir)
	exigirMotivosDeLaRaiz(t, leido, motivoEsperadoDelLimite(sinMedir))
	assert.Equal(t, VeredictoFallo, leido.informe.Veredicto)

	assert.Equal(t, reintentos, leido.informe.ReintentosPorLimiteDeRitmo)
	assert.Equal(t, strconv.Itoa(reintentos), string(leido.crudo.ReintentosPorLimiteDeRitmo))
}

// exigirLasSesionesSinAbrir escribe, con tres repeticiones y umbral 2, el informe
// de una copia del caso aprobado con la lista del repositorio en la que la serie
// del art. 21 tiene una sesión que pasa, otra con el mensaje del límite de uso y
// la tercera sin abrir tras el límite, y la de no activación, una sola sesión, que
// pasa, sin ninguna sin abrir. Exige que la serie del art. 21 cuente como sin
// medir la que tiene el mensaje y la que no se abrió, sin el motivo de su tasa ni
// el de las sesiones que faltan; que la de no activación dé los dos, porque sus
// sesiones faltan por otra causa; que el recuento cuente solo la sesión medida; y
// que el motivo de la ejecución nombre las dos sesiones sin medir, en orden de
// sesión, detrás de los de siempre.
func exigirLasSesionesSinAbrir(t *testing.T) {
	t.Helper()

	copia := copiaDelCasoAprobadoConLaLista(t)
	sesiones := filepath.Join(copia, "sesiones")

	require.NoError(t, os.Rename(filepath.Join(sesiones, sesionDelArticulo21), filepath.Join(sesiones, sesionDeLaSerie(1))))

	conElMensaje := filepath.Join(sesiones, sesionDeLaSerie(2))
	require.NoError(t, os.CopyFS(conElMensaje, os.DirFS(filepath.Join(casosDeInforme, casoAprobado, "sesiones",
		sesionDelArticulo21))))
	cambiarElFinalDeLaSesion(t, conElMensaje, 1, mensajeResultConError(t, textoDelLimiteDeSesion))

	leido := informeDeLaCopia(t, copia, func(entradas *InformeAEscribir) {
		entradas.Repeticiones, entradas.Umbral = 3, 2
		entradas.SinAbrir = []SesionPlanificada{
			{Nombre: sesionDeLaSerie(3), Fichero: ficheroDeLaEval01, Modelo: modeloQueDecide},
		}
	})

	assert.True(t, resultadoDeLaSesion(t, leido.informe, sesionDeLaSerie(1)).Pasa)
	exigirSesionSinMedir(t, leido, sesionDeLaSerie(2), sinMedirPorElMensaje, 0)
	assert.Len(t, leido.informe.Evals, 3, "la sesión que no se abrió no tiene resultado")

	exigirTasaSinMedir(t, leido, TasaDelInforme{
		Eval: ficheroDeLaEval01, Modelo: modeloQueDecide, Planificada: true, Decide: true,
		Formas: sinFormasExigidas, Sesiones: 2, Pasan: 1, SinMedir: 2,
	})
	assert.Equal(t, TasaDelInforme{
		Eval: ficheroDeNoActivacion, Modelo: modeloQueDecide, Planificada: true, Decide: true,
		Formas: sinFormasExigidas, Sesiones: 1, Pasan: 1,
	}, tasaDeLaSerie(t, leido.informe, ficheroDeNoActivacion, modeloQueDecide))
	exigirLineas(t, seccionDelInforme(t, leido.md, "Tasas por eval"),
		filaDeTabla(ficheroDeNoActivacion, modeloQueDecide, "sí", "sí", "ninguna", "1 de 1", "no llega al umbral"))

	exigirRecuento(t, leido, []RecuentoDeExpresiones{{Modelo: modeloQueDecide, ConAlguna: 0, Respuestas: 1}},
		`[{"modelo":"`+modeloQueDecide+`","con_alguna":0,"respuestas":1}]`)

	sinMedir := []SesionSinMedir{
		{Sesion: sesionDeLaSerie(2), Eval: ficheroDeLaEval01, Modelo: modeloQueDecide, Motivo: sinMedirPorElMensaje},
		{Sesion: sesionDeLaSerie(3), Eval: ficheroDeLaEval01, Modelo: modeloQueDecide, Motivo: sinAbrirTrasElLimite},
	}
	exigirSesionesSinMedir(t, leido, sinMedir...)
	exigirMotivosDeLaRaiz(t, leido,
		motivoDeLasSesionesQueFaltan(ficheroDeNoActivacion, modeloQueDecide, 1, 3),
		motivoDeLaTasa(ficheroDeNoActivacion, modeloQueDecide, 1, 1, 2),
		motivoEsperadoDelLimite(sinMedir...))
	assert.Equal(t, VeredictoFallo, leido.informe.Veredicto)
}

// exigirLosReintentosRecuperados escribe el informe de una copia del caso
// aprobado con la lista del repositorio en la que la sesión del art. 21 reintenta
// dos veces por rate_limit y una por sobrecarga, y la de no activación una por
// rate_limit, antes de terminar como en el caso aprobado. Exige que las dos se
// midan y pasen, cada una con sus reintentos por rate_limit en informe.json y en
// su fila de informe.md, y la raíz con su suma, sin ninguna sesión sin medir y
// con el veredicto aprobado (FR-033, FR-041).
func exigirLosReintentosRecuperados(t *testing.T) {
	t.Helper()

	copia := copiaDelCasoAprobadoConLaLista(t)
	insertarTrasElInit(t, filepath.Join(copia, "sesiones", sesionDelArticulo21),
		mensajeDeReintento(1, 10, 429, "rate_limit")+mensajeDeReintento(2, 10, 529, "overloaded")+
			mensajeDeReintento(3, 10, 429, "rate_limit"))
	insertarTrasElInit(t, filepath.Join(copia, "sesiones", sesionDeNoActivacion),
		mensajeDeReintento(1, 10, 429, "rate_limit"))

	leido := informeDeLaCopia(t, copia, nil)

	for sesion, reintentos := range map[string]int{sesionDelArticulo21: 2, sesionDeNoActivacion: 1} {
		resultado := resultadoDeLaSesion(t, leido.informe, sesion)
		assert.True(t, resultado.Pasa, "%s se mide y pasa", sesion)
		assert.Empty(t, resultado.SinMedir, "%s no queda sin medir", sesion)
		assert.Equal(t, reintentos, resultado.ReintentosPorLimiteDeRitmo, "reintentos de %s", sesion)

		escrito := resultadoEscrito(t, leido, sesion)
		assert.Equal(t, strconv.Itoa(reintentos), string(escrito.ReintentosPorLimiteDeRitmo),
			"reintentos_por_limite_de_ritmo de %s", sesion)
		assert.JSONEq(t, `""`, string(escrito.SinMedir), "sin_medir de %s", sesion)

		assert.Equal(t, strconv.Itoa(reintentos), celdaDeLaSesion(t, leido.md, sesion, columnaDeReintentos))
		assert.Equal(t, "no", celdaDeLaSesion(t, leido.md, sesion, columnaSinMedir))
		assert.Equal(t, "pasa", celdaDeLaSesion(t, leido.md, sesion, "Resultado"))
	}

	assert.Equal(t, 3, leido.informe.ReintentosPorLimiteDeRitmo)
	assert.Equal(t, "3", string(leido.crudo.ReintentosPorLimiteDeRitmo))
	exigirLineas(t, seccionDelInforme(t, leido.md, "Cabecera"), "Reintentos por límite de ritmo: 3")

	exigirRecuento(t, leido, []RecuentoDeExpresiones{{Modelo: modeloQueDecide, ConAlguna: 0, Respuestas: 1}},
		`[{"modelo":"`+modeloQueDecide+`","con_alguna":0,"respuestas":1}]`)
	exigirSesionesSinMedir(t, leido)
	exigirMotivosDeLaRaiz(t, leido)
	assert.Equal(t, VeredictoAprobado, leido.informe.Veredicto)
}

// exigirSesionSinMedir exige que la sesión haya quedado sin medir por la clase
// dada, con sus reintentos por rate_limit: no pasa y su único motivo es el del
// límite, en informe.json, en su fila de la tabla de las sesiones y en su
// sección de informe.md.
func exigirSesionSinMedir(t *testing.T, leido informeLeido, sesion, clase string, reintentos int) {
	t.Helper()

	motivo := "sin medir por límite de uso: " + clase

	resultado := resultadoDeLaSesion(t, leido.informe, sesion)
	assert.False(t, resultado.Pasa, "%s no pasa", sesion)
	assert.Equal(t, clase, resultado.SinMedir)
	assert.Equal(t, reintentos, resultado.ReintentosPorLimiteDeRitmo)
	assert.Equal(t, []string{motivo}, resultado.Motivos, "el único motivo de %s es el del límite", sesion)

	escrito := resultadoEscrito(t, leido, sesion)
	assert.JSONEq(t, cadenaJSON(t, clase), string(escrito.SinMedir), "sin_medir de %s", sesion)
	assert.Equal(t, strconv.Itoa(reintentos), string(escrito.ReintentosPorLimiteDeRitmo),
		"reintentos_por_limite_de_ritmo de %s", sesion)

	assert.Equal(t, clase, celdaDeLaSesion(t, leido.md, sesion, columnaSinMedir))
	assert.Equal(t, strconv.Itoa(reintentos), celdaDeLaSesion(t, leido.md, sesion, columnaDeReintentos))
	assert.Equal(t, "no pasa", celdaDeLaSesion(t, leido.md, sesion, "Resultado"))
	exigirLineas(t, seccionDelInforme(t, leido.md, "Sesión "+sesion), "Motivos de la sesión:", "- "+motivo)
}

// exigirTasaSinMedir exige la tasa de una serie sin medir: la del Informe, que
// no pasa, sus sesiones sin medir escritas en informe.json, y su fila de la tabla
// de las series de informe.md, que lo dice en su resultado.
func exigirTasaSinMedir(t *testing.T, leido informeLeido, esperada TasaDelInforme) {
	t.Helper()

	assert.Equal(t, esperada, tasaDeLaSerie(t, leido.informe, esperada.Eval, esperada.Modelo))

	posicion := slices.IndexFunc(leido.crudo.Tasas, func(tasa tasaCruda) bool {
		return tasa.Eval == esperada.Eval && tasa.Modelo == esperada.Modelo
	})
	require.GreaterOrEqual(t, posicion, 0, "informe.json tiene la tasa de %s con %s", esperada.Eval, esperada.Modelo)
	assert.Equal(t, strconv.Itoa(esperada.SinMedir), string(leido.crudo.Tasas[posicion].SinMedir))

	exigirLineas(t, seccionDelInforme(t, leido.md, "Tasas por eval"),
		filaDeTabla(esperada.Eval, esperada.Modelo, siONo(esperada.Decide), siONo(esperada.Planificada), "ninguna",
			fmt.Sprintf("%d de %d", esperada.Pasan, esperada.Sesiones), fmt.Sprintf("sin medir (%d)", esperada.SinMedir)))
}

// exigirSesionesSinMedir exige las sesiones sin medir de la raíz: las del
// Informe, las escritas en informe.json, con sus cuatro claves y una lista vacía,
// no null, si no hay ninguna, y la sección de informe.md, que es su tabla, con una
// fila por sesión y en su orden, o «ninguna».
func exigirSesionesSinMedir(t *testing.T, leido informeLeido, esperadas ...SesionSinMedir) {
	t.Helper()

	if len(esperadas) == 0 {
		assert.Empty(t, leido.informe.SesionesSinMedir)
		assert.Equal(t, "[]", string(leido.crudo.SesionesSinMedir), "sesiones_sin_medir es una lista vacía, no null")
		assert.Equal(t, "ninguna", seccionDelInforme(t, leido.md, "Sesiones sin medir"))

		return
	}

	assert.Equal(t, esperadas, leido.informe.SesionesSinMedir)

	escritas := make([]string, 0, len(esperadas))
	filas := []string{
		filaDeTabla(encabezadosDeLaTablaSinMedir...),
		filaDeTabla(slices.Repeat([]string{"---"}, len(encabezadosDeLaTablaSinMedir))...),
	}

	for _, sinMedir := range esperadas {
		escritas = append(escritas, fmt.Sprintf(`{"sesion":%s,"eval":%s,"modelo":%s,"motivo":%s}`,
			cadenaJSON(t, sinMedir.Sesion), cadenaJSON(t, sinMedir.Eval), cadenaJSON(t, sinMedir.Modelo),
			cadenaJSON(t, sinMedir.Motivo)))
		filas = append(filas, filaDeTabla(sinMedir.Sesion, sinMedir.Eval, sinMedir.Modelo, sinMedir.Motivo))
	}

	assert.JSONEq(t, "["+strings.Join(escritas, ",")+"]", string(leido.crudo.SesionesSinMedir))
	assert.Equal(t, strings.Join(filas, "\n"), seccionDelInforme(t, leido.md, "Sesiones sin medir"))
}

// motivoEsperadoDelLimite es el motivo de la raíz que dan las sesiones sin medir
// dadas, cada una con su motivo y en su orden, escrito como lo fija el contrato
// informe-del-job §2.2 de H7.3.
func motivoEsperadoDelLimite(sesiones ...SesionSinMedir) string {
	nombradas := make([]string, 0, len(sesiones))
	for _, sinMedir := range sesiones {
		nombradas = append(nombradas, sinMedir.Sesion+" ("+sinMedir.Motivo+")")
	}

	return fmt.Sprintf("de la ejecución, no de la skill: límite de uso de la cuenta: %d sesiones sin medir: %s",
		len(sesiones), strings.Join(nombradas, ", "))
}

// copiaDelCasoAprobadoConLaLista copia el caso aprobado de TestInforme en un
// directorio temporal del test, con la lista del repositorio en su carpeta de
// evals, y devuelve su ruta.
func copiaDelCasoAprobadoConLaLista(t *testing.T) string {
	t.Helper()

	copia := t.TempDir()
	require.NoError(t, os.CopyFS(copia, os.DirFS(filepath.Join(casosDeInforme, casoAprobado))))
	copiarLaListaDelRepositorio(t, filepath.Join(copia, "evals"))

	return copia
}

// cambiarElFinalDeLaSesion deja en el transcript de la sesión del directorio su
// mensaje system/init, que es su primera línea, seguido de lo dado, y le pone el
// código dado.
func cambiarElFinalDeLaSesion(t *testing.T, dir string, codigo int, trasElInit string) {
	t.Helper()

	init, _, hay := strings.Cut(contenidoDeLaSesion(t, dir, "sesion.jsonl"), "\n")
	require.True(t, hay, "el transcript de %s tiene más de una línea", dir)
	require.Contains(t, init, `"subtype":"init"`, "la primera línea del transcript de %s es system/init", dir)

	escribirEnLaCopia(t, dir, "sesion.jsonl", init+"\n"+trasElInit)
	escribirEnLaCopia(t, dir, "codigo-de-la-sesion", strconv.Itoa(codigo)+"\n")
}

// insertarTrasElInit mete lo dado en el transcript de la sesión del directorio
// detrás de su mensaje system/init, que es su primera línea.
func insertarTrasElInit(t *testing.T, dir, lineas string) {
	t.Helper()

	init, resto, hay := strings.Cut(contenidoDeLaSesion(t, dir, "sesion.jsonl"), "\n")
	require.True(t, hay, "el transcript de %s tiene más de una línea", dir)
	require.Contains(t, init, `"subtype":"init"`, "la primera línea del transcript de %s es system/init", dir)

	escribirEnLaCopia(t, dir, "sesion.jsonl", init+"\n"+lineas+resto)
}

// conElModeloInformativo pone en las entradas los modelos informativos de los
// casos que los llevan: solo modeloInformativoDelCaso.
func conElModeloInformativo(entradas *InformeAEscribir) {
	entradas.ModelosInformativos = []string{modeloInformativoDelCaso}
}

// informeDeLaCopia escribe el informe de la copia de un caso, con las entradas
// del caso aprobado, las evals y las sesiones de la copia y lo que cambie
// ajustar, si no es nil, y lo lee con leerInformeEscrito.
func informeDeLaCopia(t *testing.T, copia string, ajustar func(entradas *InformeAEscribir)) informeLeido {
	t.Helper()

	entradas := entradasDelCaso(casoAprobado, t.TempDir())
	entradas.Evals = filepath.Join(copia, "evals")
	entradas.Sesiones = filepath.Join(copia, "sesiones")

	if ajustar != nil {
		ajustar(&entradas)
	}

	informe, err := EscribirInforme(entradas)
	require.NoError(t, err)

	leido := leerInformeEscrito(t, entradas.Destino, informe)
	leido.caso = copia

	return leido
}

// copiaConExpresiones copia el caso aprobado de TestInforme en un directorio
// temporal del test y devuelve su ruta. En la copia, la carpeta de evals lleva la
// lista del repositorio si conLista; la respuesta de la sesión del art. 21 lleva
// delante la transición de la memoria de consultas, y la de la eval de no
// activación, json; modeloInformativoDelCaso tiene la sesión del art. 21 del caso
// de los modelos informativos, sin ninguna expresión, y la de no activación del
// caso aprobado, con su modelo; y, con pruebaDeRed, la copia tiene además la
// sesión de la prueba de red del caso de fuera de lo grabado, con lo dicho en
// otra conversación delante de su respuesta.
func copiaConExpresiones(t *testing.T, conLista, pruebaDeRed bool) string {
	t.Helper()

	copia := t.TempDir()
	require.NoError(t, os.CopyFS(copia, os.DirFS(filepath.Join(casosDeInforme, casoAprobado))))

	if conLista {
		copiarLaListaDelRepositorio(t, filepath.Join(copia, "evals"))
	}

	sesiones := filepath.Join(copia, "sesiones")
	anteponerALaRespuesta(t, filepath.Join(sesiones, sesionDelArticulo21), transicionDeLaMemoria+"\n\n")
	anteponerALaRespuesta(t, filepath.Join(sesiones, sesionDeNoActivacion), jsonEnLaRespuesta)

	require.NoError(t, os.CopyFS(filepath.Join(sesiones, sesionDelArticulo21ConOpus), os.DirFS(
		filepath.Join(casosDeInforme, "modelos-informativos-no-deciden", "sesiones", sesionDelArticulo21ConOpus))))
	copiarSesionConOtroModelo(t, filepath.Join(casosDeInforme, casoAprobado, "sesiones", sesionDeNoActivacion),
		filepath.Join(sesiones, sesionDeNoActivacionConOpus), modeloInformativoDelCaso)

	if pruebaDeRed {
		deLaPruebaDeRed := filepath.Join(sesiones, sesionDeLaPruebaDeRed)
		require.NoError(t, os.CopyFS(deLaPruebaDeRed, os.DirFS(filepath.Join(casosDeInforme,
			"fuera-de-lo-grabado-no-cambia-el-veredicto", "sesiones", sesionDeLaPruebaDeRed))))
		anteponerALaRespuesta(t, deLaPruebaDeRed, loDichoEnOtraConversacion+"\n\n")
	}

	return copia
}

// copiaDeLaSerieConExpresiones arma en un directorio temporal del test una
// ejecución con la eval del art. 21 del caso aprobado, informativa si se pide, la
// lista del repositorio y tres sesiones de esa eval con el modelo que decide,
// copias de la del caso aprobado, que pasa: la segunda lleva delante de su
// respuesta la transición de la memoria de consultas, y la tercera, lo dicho en
// otra conversación. Devuelve su ruta.
func copiaDeLaSerieConExpresiones(t *testing.T, informativa bool) string {
	t.Helper()

	copia := t.TempDir()

	evals := filepath.Join(copia, "evals")
	require.NoError(t, os.Mkdir(evals, 0o750))

	eval := contenidoDeLaSesion(t, filepath.Join(casosDeInforme, casoAprobado, "evals"), ficheroDeLaEval01)
	if informativa {
		eval += "informativa: true\n"
	}

	escribirEnLaCopia(t, evals, ficheroDeLaEval01, eval)
	copiarLaListaDelRepositorio(t, evals)

	delCasoAprobado := os.DirFS(filepath.Join(casosDeInforme, casoAprobado, "sesiones", sesionDelArticulo21))
	prefijos := []string{"", transicionDeLaMemoria + "\n\n", loDichoEnOtraConversacion + "\n\n"}

	for posicion, prefijo := range prefijos {
		sesion := filepath.Join(copia, "sesiones", sesionDeLaSerie(posicion+1))
		require.NoError(t, os.CopyFS(sesion, delCasoAprobado))

		if prefijo != "" {
			anteponerALaRespuesta(t, sesion, prefijo)
		}
	}

	return copia
}

// copiarLaListaDelRepositorio copia la lista de expresiones prohibidas de
// boe-legislacion en la carpeta de evals dada.
func copiarLaListaDelRepositorio(t *testing.T, evals string) {
	t.Helper()

	escribirEnLaCopia(t, evals, ficheroDeExpresionesProhibidas,
		contenidoDeLaSesion(t, evalsDelRepositorio, ficheroDeExpresionesProhibidas))
}

// anteponerALaRespuesta antepone el prefijo a la respuesta de la sesión del
// directorio, que está dos veces en su transcript: en el último mensaje del
// asistente y en el result.
func anteponerALaRespuesta(t *testing.T, dir, prefijo string) {
	t.Helper()

	sesion, err := LeerSesion(dir)
	require.NoError(t, err)
	require.NotEmpty(t, sesion.Respuesta, "la sesión de %s tiene respuesta", dir)

	escribirEnLaCopia(t, dir, "sesion.jsonl", sustituirDosVeces(t, contenidoDeLaSesion(t, dir, "sesion.jsonl"),
		cadenaJSON(t, sesion.Respuesta), cadenaJSON(t, prefijo+sesion.Respuesta)))
}

// copiarSesionConOtroModelo copia la sesión del directorio origen en el destino
// como una sesión del modelo dado: su modelo.txt lo pide y su transcript lo
// declara en system/init, el único de sus mensajes que lleva el modelo sin fecha.
func copiarSesionConOtroModelo(t *testing.T, origen, destino, modelo string) {
	t.Helper()

	require.NoError(t, os.CopyFS(destino, os.DirFS(origen)))
	escribirEnLaCopia(t, destino, "modelo.txt", modelo+"\n")

	declarado := `"model":"` + modeloDeLosTranscripts + `",`
	transcript := contenidoDeLaSesion(t, destino, "sesion.jsonl")
	require.Equal(t, 1, strings.Count(transcript, declarado), "el transcript de %s declara su modelo una vez", origen)

	escribirEnLaCopia(t, destino, "sesion.jsonl", strings.Replace(transcript, declarado, `"model":"`+modelo+`",`, 1))
}

// exigirExpresionesPorSesion exige que el informe tenga exactamente las sesiones
// dadas, en su orden, y que publique de cada una sus expresiones prohibidas: en
// informe.json, la lista escrita, vacía y no null si no lleva ninguna, y, en su
// columna de la tabla de las sesiones de informe.md, separadas por «, », o
// «ninguna».
func exigirExpresionesPorSesion(t *testing.T, leido informeLeido, esperadas []expresionesDeLaSesion) {
	t.Helper()

	nombres := make([]string, 0, len(esperadas))
	for _, esperada := range esperadas {
		nombres = append(nombres, esperada.sesion)
	}

	sesiones := make([]string, 0, len(leido.informe.Evals))
	for _, resultado := range leido.informe.Evals {
		sesiones = append(sesiones, resultado.Sesion)
	}

	require.Equal(t, nombres, sesiones, "el informe tiene las sesiones de la copia, en orden de nombre")

	for _, esperada := range esperadas {
		escritas, err := json.Marshal(esperada.encontradas)
		require.NoError(t, err)

		celda := strings.Join(esperada.encontradas, ", ")
		if len(esperada.encontradas) == 0 {
			celda = "ninguna"
		}

		assert.Equal(t, string(escritas), compacto(t, resultadoEscrito(t, leido, esperada.sesion).ExpresionesProhibidas),
			"expresiones_prohibidas de %s", esperada.sesion)
		assert.Equal(t, celda, celdaDeLaSesion(t, leido.md, esperada.sesion, columnaDeExpresiones),
			"celda de las expresiones prohibidas de %s", esperada.sesion)
	}
}

// exigirRecuento exige el recuento de las expresiones prohibidas por modelo: el
// del Informe, el escrito en informe.json, sin blancos, y la sección de
// informe.md, que es su tabla con una fila por modelo, en su orden.
func exigirRecuento(t *testing.T, leido informeLeido, esperado []RecuentoDeExpresiones, escrito string) {
	t.Helper()

	assert.Equal(t, esperado, leido.informe.ExpresionesProhibidasPorModelo)
	assert.Equal(t, escrito, compacto(t, leido.crudo.ExpresionesProhibidasPorModelo))

	filas := []string{
		filaDeTabla(encabezadosDeLaTablaDeExpresiones...),
		filaDeTabla(slices.Repeat([]string{"---"}, len(encabezadosDeLaTablaDeExpresiones))...),
	}
	for _, recuento := range esperado {
		filas = append(filas, filaDeTabla(recuento.Modelo, strconv.Itoa(recuento.ConAlguna),
			strconv.Itoa(recuento.Respuestas)))
	}

	assert.Equal(t, strings.Join(filas, "\n"), seccionDelInforme(t, leido.md, "Expresiones prohibidas por modelo"))
}

// exigirSinListaDeExpresiones exige lo que publica el informe de una skill sin
// lista de expresiones prohibidas (contrato lista-y-juicio §5 de H7.2; research
// D7): el recuento por modelo es una lista vacía, no null, y su sección de
// informe.md, el párrafo que lo dice; y cada sesión escribe sus expresiones
// prohibidas como una lista vacía, no null, con «ninguna» en su columna de la
// tabla de las sesiones.
func exigirSinListaDeExpresiones(t *testing.T, leido informeLeido) {
	t.Helper()

	assert.Empty(t, leido.informe.ExpresionesProhibidasPorModelo)
	assert.Equal(t, "[]", string(leido.crudo.ExpresionesProhibidasPorModelo),
		"expresiones_prohibidas_por_modelo es una lista vacía, no null")
	assert.Equal(t, parrafoDeLaSkillSinLista, seccionDelInforme(t, leido.md, "Expresiones prohibidas por modelo"))

	require.NotEmpty(t, leido.crudo.Evals, "el informe tiene alguna sesión que mirar")

	for _, resultado := range leido.crudo.Evals {
		assert.Equal(t, "[]", string(resultado.ExpresionesProhibidas),
			"expresiones_prohibidas de %s es una lista vacía, no null", resultado.Sesion)
		assert.Equal(t, "ninguna", celdaDeLaSesion(t, leido.md, resultado.Sesion, columnaDeExpresiones),
			"celda de las expresiones prohibidas de %s", resultado.Sesion)
	}
}

// escribirEnLaCopia reescribe un fichero de la copia de un caso con el contenido
// dado.
func escribirEnLaCopia(t *testing.T, dir, fichero, contenido string) {
	t.Helper()

	require.NoError(t, os.WriteFile(filepath.Join(dir, fichero), []byte(contenido), 0o600))
}

// cadenaJSON es el texto codificado como cadena JSON, comillas incluidas, tal como
// lo escribe un transcript.
func cadenaJSON(t *testing.T, texto string) string {
	t.Helper()

	codificado, err := json.Marshal(texto)
	require.NoError(t, err)

	return string(codificado)
}

// sustituirDosVeces sustituye en el texto cada aparición de viejo por nuevo y
// exige que sean exactamente dos.
func sustituirDosVeces(t *testing.T, texto, viejo, nuevo string) string {
	t.Helper()

	require.Equal(t, 2, strings.Count(texto, viejo), "%s está exactamente dos veces en el transcript", viejo)

	return strings.ReplaceAll(texto, viejo, nuevo)
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

	// Ninguna de las dos evals del caso prohíbe comandos ni espera avisos,
	// hallazgos, redacciones modificadas ni territorio: cada sesión los escribe
	// como listas vacías, no como null (FR-040 de H5.1; contrato de evals §2 de H6;
	// contrato evals-y-skill §2 de H7 y §6 de H7.1; data-model §4 de H7.4).
	require.Len(t, leido.crudo.Evals, 2, "informe.json tiene las dos sesiones del caso")

	for _, resultado := range leido.crudo.Evals {
		assert.Equal(t, "[]", string(resultado.ComandosProhibidosEjecutados),
			"comandos_prohibidos_ejecutados de %s es una lista vacía, no null", resultado.Sesion)
		assert.Equal(t, "[]", string(resultado.AvisosEncontrados),
			"avisos_encontrados de %s es una lista vacía, no null", resultado.Sesion)
		assert.Equal(t, "[]", string(resultado.AvisosAusentes),
			"avisos_ausentes de %s es una lista vacía, no null", resultado.Sesion)
		assert.Equal(t, "[]", string(resultado.HallazgosEncontrados),
			"hallazgos_encontrados de %s es una lista vacía, no null", resultado.Sesion)
		assert.Equal(t, "[]", string(resultado.HallazgosAusentes),
			"hallazgos_ausentes de %s es una lista vacía, no null", resultado.Sesion)
		assert.Equal(t, "[]", string(resultado.RedaccionesEncontradas),
			"redacciones_modificadas_encontradas de %s es una lista vacía, no null", resultado.Sesion)
		assert.Equal(t, "[]", string(resultado.RedaccionesAusentes),
			"redacciones_modificadas_ausentes de %s es una lista vacía, no null", resultado.Sesion)
		assert.Equal(t, "[]", string(resultado.TerritorioEncontrado),
			"territorio_encontrado de %s es una lista vacía, no null", resultado.Sesion)
		assert.Equal(t, "[]", string(resultado.TerritorioAusente),
			"territorio_ausente de %s es una lista vacía, no null", resultado.Sesion)
	}

	// Sin hallazgos esperados, ninguna serie exige forma: formas es una lista
	// vacía, no null, y su celda de la tabla de las series dice «ninguna».
	require.Len(t, leido.crudo.Tasas, 2, "informe.json tiene las dos series del caso")

	for _, tasa := range leido.crudo.Tasas {
		assert.Equal(t, "[]", string(tasa.Formas), "formas de %s con %s es una lista vacía, no null",
			tasa.Eval, tasa.Modelo)
		assert.Equal(t, sinFormasExigidas, tasaDeLaSerie(t, leido.informe, tasa.Eval, tasa.Modelo).Formas)
	}

	exigirLineas(t, seccionDelInforme(t, leido.md, "Tasas por eval"),
		filaDeTabla(encabezadosDeLaTablaDeTasas...),
		filaDeTabla(slices.Repeat([]string{"---"}, len(encabezadosDeLaTablaDeTasas))...),
		filaDeTabla(ficheroDeLaEval01, modeloQueDecide, "sí", "sí", "ninguna", "1 de 1", "llega al umbral"),
		filaDeTabla(ficheroDeNoActivacion, modeloQueDecide, "sí", "sí", "ninguna", "1 de 1", "llega al umbral"))

	// Los comandos prohibidos ejecutados van en la tabla de las sesiones detrás
	// de los comandos ausentes, los hallazgos encontrados y los ausentes, detrás de
	// los avisos, las redacciones modificadas encontradas y las ausentes, detrás de
	// los hallazgos, el territorio encontrado y el ausente, detrás de las
	// redacciones, y las expresiones prohibidas, detrás del territorio, vacíos en
	// las dos.
	exigirLineas(t, seccionDelInforme(t, leido.md, "Sesiones"),
		filaDeTabla(encabezadosDeLaTablaDeSesiones...),
		filaDeTabla(slices.Repeat([]string{"---"}, len(encabezadosDeLaTablaDeSesiones))...),
		filaDeTabla(sesionDelArticulo21, ficheroDeLaEval01, modeloQueDecide, "sí", "sí", "sí (código 0)",
			"ninguno", "ninguno", "ninguna", "ninguno", "ninguno", "ninguno", "ninguno", "ninguna", "ninguna",
			"ninguno", "ninguno", "ninguna", "0", "no", "pasa"),
		filaDeTabla(sesionDeNoActivacion, ficheroDeNoActivacion, modeloQueDecide, "no", "no", "sí (código 0)",
			"ninguno", "ninguno", "ninguna", "ninguno", "ninguno", "ninguno", "ninguno", "ninguna", "ninguna",
			"ninguno", "ninguno", "ninguna", "0", "no", "pasa"))

	// Ninguna sesión reintenta ni queda sin medir: cada una escribe sus
	// reintentos por rate_limit, 0, y sin_medir vacío; cada serie, sin_medir 0;
	// y la raíz, 0 reintentos y ninguna sesión sin medir, con «ninguna» en su
	// sección de informe.md (contrato informe-del-job §3 y §4 de H7.3).
	for _, resultado := range leido.crudo.Evals {
		assert.Equal(t, "0", string(resultado.ReintentosPorLimiteDeRitmo),
			"reintentos_por_limite_de_ritmo de %s", resultado.Sesion)
		assert.JSONEq(t, `""`, string(resultado.SinMedir), "sin_medir de %s", resultado.Sesion)
	}

	for _, tasa := range leido.crudo.Tasas {
		assert.Equal(t, "0", string(tasa.SinMedir), "sin_medir de %s con %s", tasa.Eval, tasa.Modelo)
	}

	exigirSesionesSinMedir(t, leido)
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
// porque a la serie de la eval bien formada que ningún eval.txt nombra le faltan
// todas sus sesiones.
func comprobarEvalSinSesion(t *testing.T, leido informeLeido) {
	t.Helper()

	require.Len(t, leido.informe.Evals, 1)
	assert.True(t, resultadoDeLaSesion(t, leido.informe, sesionDelArticulo21).Pasa)
	exigirMotivosDeLaRaiz(t, leido,
		motivoDeLasSesionesQueFaltan(ficheroDeNoActivacion, modeloQueDecide, 0, 1),
		motivoDeLaTasa(ficheroDeNoActivacion, modeloQueDecide, 0, 0, 1))
	assert.Equal(t, VeredictoFallo, leido.informe.Veredicto)
}

// comprobarSinEvalTxt exige la sesión sin eval.txt ilegible, con la eval vacía:
// sin la eval no cae en la serie que el plan pide, que queda sin sesiones.
func comprobarSinEvalTxt(t *testing.T, leido informeLeido) {
	t.Helper()

	motivo := exigirSesionIlegible(t, leido, "eval.txt")
	exigirMotivosDeLaRaiz(t, leido,
		slices.Concat(motivosDeLaSerieSinSesiones(), []string{sesionDelArticulo21 + ": " + motivo})...)
	assert.Empty(t, resultadoDeLaSesion(t, leido.informe, sesionDelArticulo21).Eval)
}

// comprobarEvalDesconocida exige la sesión cuyo eval.txt no nombra ninguna eval
// ilegible, con la eval que nombra en el resultado y en el motivo, y la serie que
// el plan pide sin ninguna sesión.
func comprobarEvalDesconocida(t *testing.T, leido informeLeido) {
	t.Helper()

	const desconocida = "03-inexistente.yaml"

	motivo := exigirSesionIlegible(t, leido, "eval.txt")
	exigirMotivosDeLaRaiz(t, leido,
		slices.Concat(motivosDeLaSerieSinSesiones(), []string{sesionDelArticulo21 + ": " + motivo})...)
	assert.Contains(t, motivo, desconocida)
	assert.Equal(t, desconocida, resultadoDeLaSesion(t, leido.informe, sesionDelArticulo21).Eval)
}

// comprobarSinPreguntaTxt exige la sesión sin pregunta.txt ilegible y, aun así,
// dentro de la serie que el plan pide: sin la pregunta no se sabe si lleva algo
// más que la de su eval, y se toma como la de su eval.
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
		Formas: sinFormasExigidas, Sesiones: 3, Pasan: 2, Pasa: true,
	}, tasa)

	assert.False(t, resultadoDeLaSesion(t, leido.informe, sesionDelArticulo21+"-claude-haiku-03").Pasa,
		"la tercera sesión no pasa y aun así la serie llega al umbral")
	exigirMotivosDeLaRaiz(t, leido)
	assert.Equal(t, VeredictoAprobado, leido.informe.Veredicto)

	exigirLineas(t, seccionDelInforme(t, leido.md, "Tasas por eval"),
		filaDeTabla(ficheroDeLaEval01, modeloQueDecide, "sí", "sí", "ninguna", "2 de 3", "llega al umbral"))
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
		Formas: sinFormasExigidas, Sesiones: 1, Pasan: 0,
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
		Formas: sinFormasExigidas, Sesiones: 1, Pasan: 0,
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
// final; informe.json igual al Informe devuelto; la cabecera, con la skill, el
// modelo y el commit recibidos y la comprobación sin Python byte a byte, en
// informe.json y en informe.md, que empieza por su título y lleva el veredicto de
// informe.json; el recuento de las expresiones prohibidas por modelo detrás de
// las tasas, en informe.json, y su sección detrás de la de las tasas, en
// informe.md (contrato lista-y-juicio §5 de H7.2); los umbrales detrás del
// recuento, seguidos de la duración de las sesiones, los reintentos por límite de
// ritmo y las sesiones sin medir, en informe.json, con los umbrales siempre como
// lista y cumpliendo los invariantes del ADR 0029, y, en informe.md, la sección
// de los umbrales detrás de la del recuento, la de las sesiones sin medir detrás
// de ella y la duración y los reintentos en la cabecera (contrato
// informe-del-job §1, §3 y §4 de H7.3).
func leerInformeEscrito(t *testing.T, destino string, devuelto Informe) informeLeido {
	t.Helper()

	escrito := contenidoDelInforme(t, destino, "informe.json")

	codificado, err := json.Marshal(devuelto)
	require.NoError(t, err)
	assert.JSONEq(t, string(codificado), escrito, "informe.json es el Informe que devuelve EscribirInforme")

	leido := informeLeido{md: contenidoDelInforme(t, destino, "informe.md")}
	require.NoError(t, json.Unmarshal([]byte(escrito), &leido.informe))
	require.NoError(t, json.Unmarshal([]byte(escrito), &leido.crudo))

	for anterior, siguiente := range map[string]string{
		"tasas":                             "expresiones_prohibidas_por_modelo",
		"expresiones_prohibidas_por_modelo": "umbrales",
		"umbrales":                          "duracion_de_las_sesiones",
		"duracion_de_las_sesiones":          "reintentos_por_limite_de_ritmo",
		"reintentos_por_limite_de_ritmo":    "sesiones_sin_medir",
	} {
		assert.Equal(t, siguiente, claveDetras(t, escrito, anterior), "en informe.json, %s va detrás de %s",
			siguiente, anterior)
	}

	for anterior, siguiente := range map[string]string{
		"Tasas por eval":                    "Expresiones prohibidas por modelo",
		"Expresiones prohibidas por modelo": "Umbrales",
		"Umbrales":                          "Sesiones sin medir",
	} {
		assert.Equal(t, siguiente, seccionDetras(t, leido.md, anterior), "en informe.md, la sección %s va detrás de %s",
			siguiente, anterior)
	}

	assert.True(t, strings.HasPrefix(compacto(t, leido.crudo.Umbrales), "["), "umbrales es una lista, nunca null")
	assert.Equal(t, strconv.Itoa(leido.informe.DuracionDeLasSesiones), string(leido.crudo.DuracionDeLasSesiones))
	exigirLosInvariantesDeLosUmbrales(t, leido)

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
		"Commit: "+commitEvaluado,
		"Duración de las sesiones: "+strconv.Itoa(leido.informe.DuracionDeLasSesiones)+" s",
		"Reintentos por límite de ritmo: "+strconv.Itoa(leido.informe.ReintentosPorLimiteDeRitmo))
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

// resultadoEscrito es, tal como está escrito en informe.json, el resultado de la
// sesión.
func resultadoEscrito(t *testing.T, leido informeLeido, sesion string) resultadoCrudo {
	t.Helper()

	posicion := slices.IndexFunc(leido.crudo.Evals, func(resultado resultadoCrudo) bool { return resultado.Sesion == sesion })
	require.GreaterOrEqual(t, posicion, 0, "informe.json tiene el resultado de la sesión %s", sesion)

	return leido.crudo.Evals[posicion]
}

// compacto es el valor escrito sin blancos, para compararlo con su forma en una
// línea; un valor que falta no se puede compactar y es un fallo del test.
func compacto(t *testing.T, valor jsontext.Value) string {
	t.Helper()

	copia := slices.Clone(valor)
	require.NoError(t, copia.Compact())

	return string(copia)
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

// seccionDetras es el título de la sección ## que sigue a la sección ## titulo de
// informe.md.
func seccionDetras(t *testing.T, md, titulo string) string {
	t.Helper()

	_, resto, encontrada := strings.Cut(md, "\n## "+titulo+"\n")
	require.True(t, encontrada, "informe.md tiene la sección ## %s", titulo)

	_, siguiente, hay := strings.Cut(resto, "\n## ")
	require.True(t, hay, "informe.md tiene una sección detrás de ## %s", titulo)

	siguiente, _, _ = strings.Cut(siguiente, "\n")

	return siguiente
}

// claveDetras es la clave de la raíz de informe.json que sigue a la clave dada,
// en el orden en que están escritas.
func claveDetras(t *testing.T, escrito, clave string) string {
	t.Helper()

	decodificador := jsontext.NewDecoder(strings.NewReader(escrito))

	inicio, err := decodificador.ReadToken()
	require.NoError(t, err)
	require.Equal(t, jsontext.KindBeginObject, inicio.Kind(), "informe.json es un objeto")

	var claves []string

	for decodificador.PeekKind() == jsontext.KindString {
		nombre, err := decodificador.ReadToken()
		require.NoError(t, err)

		claves = append(claves, nombre.String())
		require.NoError(t, decodificador.SkipValue())
	}

	posicion := slices.Index(claves, clave)
	require.GreaterOrEqual(t, posicion, 0, "informe.json tiene la clave %s en la raíz: %q", clave, claves)
	require.Less(t, posicion+1, len(claves), "informe.json tiene una clave detrás de %s: %q", clave, claves)

	return claves[posicion+1]
}

// celdaDeLaSesion es la celda de la columna dada en la fila de la sesión de la
// tabla de las sesiones de informe.md, cuyos encabezados son
// encabezadosDeLaTablaDeSesiones.
func celdaDeLaSesion(t *testing.T, md, sesion, columna string) string {
	t.Helper()

	posicion := slices.Index(encabezadosDeLaTablaDeSesiones, columna)
	require.GreaterOrEqual(t, posicion, 0, "la tabla de las sesiones tiene la columna %s", columna)

	tabla := seccionDelInforme(t, md, "Sesiones")
	exigirLineas(t, tabla, filaDeTabla(encabezadosDeLaTablaDeSesiones...))

	fila := filaQueEmpiezaPor(t, tabla, sesion)
	celdas := strings.Split(strings.TrimSuffix(strings.TrimPrefix(fila, "| "), " |"), " | ")
	require.Len(t, celdas, len(encabezadosDeLaTablaDeSesiones), "la fila de %s tiene una celda por columna: %s",
		sesion, fila)

	return celdas[posicion]
}

// exigirCadaCelda exige que la tabla de informe.md, con los encabezados dados,
// tenga alguna fila y que en cada una la celda de la columna dada sea la
// esperada.
func exigirCadaCelda(t *testing.T, tabla string, encabezados []string, columna, esperada string) {
	t.Helper()

	posicion := slices.Index(encabezados, columna)
	require.GreaterOrEqual(t, posicion, 0, "la tabla tiene la columna %s", columna)

	lineas := strings.Split(tabla, "\n")
	require.Greater(t, len(lineas), 2, "la tabla tiene alguna fila:\n%s", tabla)
	require.Equal(t, filaDeTabla(encabezados...), lineas[0], "la tabla empieza por sus encabezados")

	for _, fila := range lineas[2:] {
		celdas := strings.Split(strings.TrimSuffix(strings.TrimPrefix(fila, "| "), " |"), " | ")
		require.Len(t, celdas, len(encabezados), "la fila tiene una celda por columna: %s", fila)
		assert.Equal(t, esperada, celdas[posicion], "la celda %s de la fila %s", columna, fila)
	}
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
