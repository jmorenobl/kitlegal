package evals

import (
	"cmp"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

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
// un caso; y votos, los mensajes de los votos que informeDeLaCopia vio pedir al
// votante que da a toda ejecución, que dice no a todo: uno por respuesta juzgada.
type informeLeido struct {
	caso    string
	informe Informe
	crudo   informeCrudo
	md      string
	votos   []string
}

// informeCrudo es lo que se lee de informe.json sin convertirlo a un tipo de Go,
// para distinguir null de una lista vacía: los motivos de la raíz, las formas
// exigidas de cada serie, los
// umbrales, la duración de las sesiones, los reintentos por límite de ritmo y las
// sesiones sin medir de la raíz y, de cada sesión, sus avisos encontrados y
// ausentes, sus reintentos, si quedó sin medir y, de
// cada una de sus invocaciones, su código y sus conexiones. Desde H21, de cada
// sesión, además, su modo, si lleva la línea sin consulta y sus citas sin
// consulta, y de cada invocación, si es una llamada a una herramienta. Desde
// H24, lo que lleva del juez, que es null si la skill no lo tiene.
type informeCrudo struct {
	Motivos                    jsontext.Value   `json:"motivos"`
	Tasas                      []tasaCruda      `json:"tasas"`
	Umbrales                   jsontext.Value   `json:"umbrales"`
	Juez                       jsontext.Value   `json:"juez"`
	DuracionDeLasSesiones      jsontext.Value   `json:"duracion_de_las_sesiones"`
	ReintentosPorLimiteDeRitmo jsontext.Value   `json:"reintentos_por_limite_de_ritmo"`
	SesionesSinMedir           jsontext.Value   `json:"sesiones_sin_medir"`
	Evals                      []resultadoCrudo `json:"evals"`
}

// tasaCruda es una serie de informe.json con su modo, sus formas exigidas y sus
// sesiones sin medir tal como están escritos.
type tasaCruda struct {
	Eval     string         `json:"eval"`
	Modelo   string         `json:"modelo"`
	Modo     jsontext.Value `json:"modo"`
	Formas   jsontext.Value `json:"formas"`
	SinMedir jsontext.Value `json:"sin_medir"`
}

// resultadoCrudo es el resultado de una sesión de informe.json con sus comandos
// prohibidos ejecutados, sus avisos, sus hallazgos, su territorio, sus
// reintentos por límite de ritmo, si quedó sin medir y sus
// invocaciones tal como están escritos; y, desde H21, su modo, si lleva la línea
// sin consulta y sus citas sin consulta (contracts/evals-en-dos-modos.md §4 de
// H21).
type resultadoCrudo struct {
	Sesion                       string            `json:"sesion"`
	Modo                         jsontext.Value    `json:"modo"`
	ComandosProhibidosEjecutados jsontext.Value    `json:"comandos_prohibidos_ejecutados"`
	AvisosEncontrados            jsontext.Value    `json:"avisos_encontrados"`
	AvisosAusentes               jsontext.Value    `json:"avisos_ausentes"`
	HallazgosEncontrados         jsontext.Value    `json:"hallazgos_encontrados"`
	HallazgosAusentes            jsontext.Value    `json:"hallazgos_ausentes"`
	RedaccionesEncontradas       jsontext.Value    `json:"redacciones_modificadas_encontradas"`
	RedaccionesAusentes          jsontext.Value    `json:"redacciones_modificadas_ausentes"`
	TerritorioEncontrado         jsontext.Value    `json:"territorio_encontrado"`
	TerritorioAusente            jsontext.Value    `json:"territorio_ausente"`
	LineaSinConsulta             jsontext.Value    `json:"linea_sin_consulta"`
	CitasSinConsulta             jsontext.Value    `json:"citas_sin_consulta"`
	ReintentosPorLimiteDeRitmo   jsontext.Value    `json:"reintentos_por_limite_de_ritmo"`
	SinMedir                     jsontext.Value    `json:"sin_medir"`
	Invocaciones                 []invocacionCruda `json:"invocaciones"`
}

// Lo que la lista de expresiones prohibidas dejaba en el informe hasta H24,
// escrito a mano, y que ningún informe lleva ya (contracts/informe-del-job.md §6
// y §7 de H24; FR-062): en informe.json, la clave de la raíz con el recuento por
// modelo —la de cada sesión es claveDeLasExpresiones—; y en informe.md, la
// sección de ese recuento y la columna de la tabla de las sesiones.
const (
	claveDelRecuentoDeExpresiones   = "expresiones_prohibidas_por_modelo"
	seccionDelRecuentoDeExpresiones = "Expresiones prohibidas por modelo"
	columnaDeExpresiones            = "Expresiones prohibidas"
)

// encabezadosDeLaTablaDeSesiones son los de la tabla de las sesiones de
// informe.md, con los comandos prohibidos ejecutados junto a los comandos
// ausentes (contrato evals-y-skill §2 de H7), los avisos junto a las citas, los
// hallazgos junto a los avisos (contrato evals-y-skill §6 de H7.1), las
// redacciones modificadas detrás de los hallazgos (contracts/informe-del-job.md
// §4 de H7.4), el territorio detrás (contrato de evals §2 de H6) y, entre el
// territorio ausente y el resultado, los reintentos por límite de ritmo y si la
// sesión quedó sin medir (contrato informe-del-job §4 de H7.3). Desde H21, el
// modo va detrás del modelo (contracts/evals-en-dos-modos.md §5.3). Desde H24,
// sin la columna de las expresiones prohibidas (contracts/informe-del-job.md §7
// de H24).
var encabezadosDeLaTablaDeSesiones = []string{
	"Sesión", "Eval", "Modelo", columnaDelModo, "Activa", "Activada", "Sesión terminada", "Comandos ausentes",
	"Comandos prohibidos ejecutados", "Citas ausentes", "Avisos encontrados", "Avisos ausentes",
	"Hallazgos encontrados", "Hallazgos ausentes", columnaDeRedaccionesEncontradas, columnaDeRedaccionesAusentes,
	"Territorio encontrado", "Territorio ausente", columnaDeReintentos, columnaSinMedir,
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

// columnaDelModo es la columna de las tablas de las series y de las sesiones de
// informe.md que dice el modo —orden o
// herramienta—, y celdaSinModo, lo que lleva en una serie o una sesión de
// una eval sin binario ni servidor, que no es de ninguno
// (contracts/evals-en-dos-modos.md §5.3 de H21).
const (
	columnaDelModo = "Modo"
	celdaSinModo   = "—"
)

// encabezadosDeLaTablaDeInvocaciones son los de la tabla de las invocaciones de
// la sección de cada sesión de informe.md: detrás de las conexiones, desde H21,
// si la invocación es una llamada a una herramienta
// (contracts/evals-en-dos-modos.md §5.3 de H21).
var encabezadosDeLaTablaDeInvocaciones = []string{"Orden", "Código", "Conexiones", "Llamada"}

// encabezadosDeLaTablaDeTasas son los de la tabla de las series de informe.md,
// con las formas exigidas junto a la tasa (contrato evals-y-skill §6 de H7.1) y,
// desde H21, el modo detrás del modelo.
var encabezadosDeLaTablaDeTasas = []string{
	"Eval", "Modelo", columnaDelModo, "Decide", "Planificada", "Formas exigidas", "Tasa", "Resultado",
}

// sinFormasExigidas son las formas exigidas de la serie de una eval que no espera
// hallazgos: una lista vacía, no nil, igual que la que se lee de informe.json.
var sinFormasExigidas = []string{}

// invocacionCruda es una invocación de informe.json con su código, sus
// conexiones y, desde H21, su marca de llamada a una herramienta tal como están
// escritos.
type invocacionCruda struct {
	Orden      string         `json:"orden"`
	Codigo     jsontext.Value `json:"codigo"`
	Conexiones jsontext.Value `json:"conexiones"`
	Llamada    jsontext.Value `json:"llamada"`
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
// Desde H7.3, sin objetivo de duración, ningún caso tiene umbrales
// (exigirSinUmbrales); los que tiene una skill con juez o con objetivo los
// fijan TestUmbralesDelInforme y TestInformeMarkdownDeLosUmbrales.
//
// Desde H24, ninguna carpeta de evals de los casos tiene juez, y ningún informe
// lleva nada de la lista de expresiones prohibidas, la tenga o no su carpeta: lo
// exige de todos leerInformeEscrito (exigirElInformeSinLaLista).
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
				filaDeTabla(sesionDelArticulo21, ficheroDeLaEval01, modeloQueDecide, "orden", "sí", "sí", "sí (código 0)",
					"ninguno", "ninguno", caso.citasAusentes, "derogada", "vigencia-agotada", "ninguno", "ninguno",
					"ninguna", "ninguna", "ninguno", "ninguno", "0", "no", "no pasa"),
				filaDeTabla(sesionDeNoActivacion, ficheroDeNoActivacion, modeloQueDecide, "orden", "no", "no", "sí (código 0)",
					"ninguno", "ninguno", "ninguna", "ninguno", "ninguno", "ninguno", "ninguno", "ninguna", "ninguna",
					"ninguno", "ninguno", "0", "no", "pasa"))

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
		filaDeTabla(sesionDelArticulo21, ficheroDeLaEval01, modeloQueDecide, "orden", "sí", "sí", "sí (código 0)",
			"ninguno", textoDeGraphShow, "ninguna", "ninguno", "ninguno", "ninguno", "ninguno", "ninguna", "ninguna",
			"ninguno", "ninguno", "0", "no", "no pasa"),
		filaDeTabla(sesionDeNoActivacion, ficheroDeNoActivacion, modeloQueDecide, "orden", "no", "no", "sí (código 0)",
			"ninguno", "ninguno", "ninguna", "ninguno", "ninguno", "ninguno", "ninguno", "ninguna", "ninguna",
			"ninguno", "ninguno", "0", "no", "pasa"))

	exigirLineas(t, seccionDelInforme(t, leido.md, "Sesión "+sesionDelArticulo21),
		filaDeTabla(encabezadosDeLaTablaDeInvocaciones...),
		filaDeTabla(ordenDeGraphShow, "3", "sin conexiones", "no"))
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
				Eval: ficheroDeLaEval01, Modelo: modeloQueDecide, Modo: ModoOrden, Planificada: true, Decide: true,
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
				filaDeTabla(ficheroDeLaEval01, modeloQueDecide, "orden", "sí", "sí", formaDeVersionObsoleta, caso.tasa,
					caso.umbral),
				filaDeTabla(ficheroDeNoActivacion, modeloQueDecide, "orden", "sí", "sí", "ninguna", "1 de 1",
					"llega al umbral"))

			exigirLineas(t, seccionDelInforme(t, leido.md, "Sesiones"),
				filaDeTabla(encabezadosDeLaTablaDeSesiones...),
				filaDeTabla(sesionDelArticulo21, ficheroDeLaEval01, modeloQueDecide, "orden", "sí", "sí", "sí (código 0)",
					"ninguno", "ninguno", "ninguna", "ninguno", "ninguno", caso.encontrados, caso.ausentes, "ninguna",
					"ninguna", "ninguno", "ninguno", "0", "no", caso.resultado),
				filaDeTabla(sesionDeNoActivacion, ficheroDeNoActivacion, modeloQueDecide, "orden", "no", "no", "sí (código 0)",
					"ninguno", "ninguno", "ninguna", "ninguno", "ninguno", "ninguno", "ninguno", "ninguna", "ninguna",
					"ninguno", "ninguno", "0", "no", "pasa"))

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

// Lo que TestInformeConListaMalFormada y la copia con expresiones del juicio del
// sondeo (copiaConExpresiones) ponen en sus copias de los casos de TestInforme
// (contrato lista-y-juicio §5 de H7.2).
const (
	// sesionDelArticulo21ConOpus y sesionDeNoActivacionConOpus son las sesiones
	// de las dos evals del caso aprobado con modeloInformativoDelCaso.
	sesionDelArticulo21ConOpus  = "01-lpac-articulo-21-claude-opus-5-01"
	sesionDeNoActivacionConOpus = "11-no-activa-programacion-claude-opus-5-01"

	// jsonEnLaRespuesta es lo que se antepone a la respuesta de la sesión de la
	// eval de no activación: una frase de programación con json, que está en la
	// lista.
	jsonEnLaRespuesta = "Si la lista llega en JSON, primero hay que leerla. "

	// listaSinOtraConversacion es una lista de expresiones prohibidas mal formada:
	// le falta la familia otra_conversacion, que el esquema exige; las otras dos
	// familias de H7.3, las dos claves de H7.4 y la de H24 están bien formadas.
	listaSinOtraConversacion = "maquinaria:\n  - memoria de consultas\n  - hallazgos\n" +
		"anuncio:\n  - ya puedo responder\n" +
		"redaccion_no_leida:\n  - ya no exige\n" +
		"salida_de_las_herramientas:\n  - el sobre\n" +
		"formas_fijas:\n  - No se ha podido comprobar si la redacci\xc3\xb3n ha cambiado\n"
)

// TestInformeConListaMalFormada fija lo que el informe hace con una lista de
// expresiones prohibidas que no valida (FR-055; research D8 de H7.2; US3.9):
// sobre una copia del caso aprobado cuya lista no tiene la familia
// otra_conversacion y cuya sesión del art. 21 lleva dos expresiones de la
// maquinaria de esa lista, la lista queda en ficheros_mal_formados con el error de
// LeerConjunto, que es el único motivo de la raíz, y el veredicto es fallo. Desde
// H24 la lista no juzga ninguna respuesta, y sigue siendo un fichero de la
// carpeta que tiene que estar bien formado (FR-071 de H24): la sesión pasa, como
// con cualquier lista. Nada se escribe bajo testdata/.
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
		"la sesión del art. 21 pasa con expresiones de la lista en su respuesta")
}

// encabezadosDeLaTablaDeUmbrales son los de la tabla de la sección «Umbrales» de
// informe.md (contrato informe-del-job §4 de H7.3).
var encabezadosDeLaTablaDeUmbrales = []string{"Umbral", "Medida", "Condición", "Cumple", "Hace fallar el veredicto"}

// TestInformeMarkdownDeLosUmbrales fija la sección «Umbrales» de informe.md
// (contrato informe-del-job §4 de H7.3; FR-001; US2-6), sobre ejecuciones
// sintéticas de ejecucionConUmbrales: detrás de las tasas y delante de las
// sesiones sin medir, lo que leerInformeEscrito exige en todo informe, una tabla
// con una fila por umbral, en su orden: el nombre, la medida —«<medida> de
// <total> (<p> %)» con un decimal y coma, o la medida sola—, la condición, si se
// cumple y si hace fallar el veredicto; lo que dice de uno que no decide lo fija
// TestUmbralQueSoloSePublica. La del contrato, fila a fila, con las de
// contracts/informe-del-job.md §2 de H7.4 que siguen en H24 —0 de 54 sin activar
// y 544 s—; la de los dos sin cumplir; y, sin ninguno, el párrafo «ninguno». La
// cabecera lleva la duración de las sesiones.
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
// en la segunda. Sin redacciones esperadas, las tres celdas dicen «ninguna».
//
// Desde H21 (contracts/evals-en-dos-modos.md §5.1 y §5.3; FR-043, FR-048), cada
// fila nombra su modo: las de un plan de un modo, el modo orden; y las del plan
// del job, las cuatro que siguen en H24 —la de las respuestas sin activar de
// cada modo y las dos duraciones—, con la suma de las tres tandas en la
// cabecera, entre las doce de contracts/informe-del-job.md §2 de H24: detrás de
// la de las respuestas sin activar de cada modo, la de cada clase del juez, con
// «no: solo se publica» la que no decide; detrás de las de los dos modos, las
// dos de la medida del juez; y detrás de las de la duración de las sesiones,
// las de la duración del juez. Y las tablas de las series y de las sesiones llevan la columna
// «Modo» detrás del modelo, con «orden»,
// «herramienta» o, en una serie o una sesión de la eval sin binario ni servidor,
// «—»; la serie del modo herramienta lleva su modo también en su modelo, como en
// informe.json, y la sección de una sesión del modo herramienta marca la
// invocación que es una llamada.
//
// Desde H24 (contracts/informe-del-job.md §2, §6 y §7 de H24; FR-034, FR-062),
// ninguna fila es de las expresiones de la lista, aunque haya respuestas que las
// lleven, y la tabla está solo si la carpeta de evals tiene la carpeta juez o el
// job da un objetivo de duración.
//
// Y fija la sección «Juez» (contracts/informe-del-job.md §7 de H24; FR-060,
// FR-061), detrás de la de los umbrales, que es donde la exige
// leerInformeEscrito: con juez, su modelo, la versión de Claude Code de sus
// votos y sus dos tablas —la de los votos de cada respuesta con algún voto
// afirmativo, con una fila por voto y clase, y la de las respuestas sin
// juzgar—, cada una con «ninguna» en su lugar si no tiene filas; y sin juez,
// «La skill no tiene juez.».
//
// Desde H25 (contracts/juez-de-jurisprudencia.md §4 de H25; FR-013 de H25), la
// octava columna de la tabla de los votos, que sigue teniendo diez, toma su
// nombre de los votos que lleva: «Sentencia» con un juez cuyo esquema lleva
// sentencia, con «—» en la sentencia que un voto dejó vacía y en la de la clase
// que no la tiene; y «Precepto» con el juez de siempre, cuya sección no cambia
// en un byte (FR-005 de H25).
func TestInformeMarkdownDeLosUmbrales(t *testing.T) {
	t.Parallel()

	comoElJob := ejecucionConUmbrales{queDeciden: 10, informativas: 8, conHaiku: true}

	conVotos, seccionConVotos := votosDelMarkdown(t)
	conSentencia, seccionConSentencia := votosDelMarkdownConSentencia(t)

	casos := []struct {
		nombre    string
		ejecucion ejecucionConUmbrales

		// conSentencia dice que el juez de la ejecución es el del esquema con
		// sentencia (escribirEjecucionConSentencia), y no el de siempre.
		conSentencia bool

		// grabados son los votos grabados de algunas de sus respuestas; sin
		// ellos, el juez dice no a todo.
		grabados []votosDeUnaRespuesta

		// filas son las de la tabla, sin los encabezados; nil, sin tabla.
		filas [][]string

		// juez es la sección «Juez»; vacía, la de una skill sin juez si la
		// ejecución no lo tiene y, si lo tiene, la de un juez sin ningún voto
		// afirmativo ni ninguna respuesta sin juzgar.
		juez string

		// formas es la celda «Formas exigidas» de cada serie de «Tasas por eval»;
		// encontradas y ausentes, las celdas de las redacciones modificadas de cada
		// sesión de «Sesiones». Vacías, «ninguna».
		formas, encontradas, ausentes string

		// exigir, si no es nil, exige lo que el caso fija además de su tabla.
		exigir func(t *testing.T, leido informeLeido)
	}{
		{
			// Dos respuestas con expresiones de la lista, que hasta H24 daban la
			// primera fila del contrato: ya no dan ninguna.
			nombre: "las-filas-del-contrato",
			ejecucion: conCambios(comoElJob, func(e *ejecucionConUmbrales) {
				e.conExpresiones, e.duracion, e.objetivo = 2, 544, 900
			}),
			filas: [][]string{
				{"`sin_activar:claude-sonnet-5-5:orden`", "0 de 54 (0,0 %)", "≤ 0,0 %", "sí", "sí"},
				{"`afirma_lo_no_leido:claude-sonnet-5-5:orden`", "0 de 54 (0,0 %)", "≤ 0,0 %", "sí", "sí"},
				{"`cuenta_su_proceso:claude-sonnet-5-5:orden`", "0 de 54 (0,0 %)", "≤ 0,0 %", "sí", "no: solo se publica"},
				{"`medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`", "0 de 212 (0,0 %)", "≤ 0,0 %", "sí", "sí"},
				{"`medida_del_juez:afirma_lo_no_leido:correctos_marcados`", "0 de 47 (0,0 %)", "≤ 0,0 %", "sí", "sí"},
				{"`duracion_de_las_sesiones:orden`", "544", "≤ 900", "sí", "sí"},
				{"`duracion_del_juez:orden`", "0", "≤ 900", "sí", "sí"},
			},
			exigir: func(t *testing.T, leido informeLeido) {
				t.Helper()

				exigirCadaCelda(t, seccionDelInforme(t, leido.md, "Tasas por eval"), encabezadosDeLaTablaDeTasas,
					columnaDelModo, "orden")
				exigirCadaCelda(t, seccionDelInforme(t, leido.md, "Sesiones"), encabezadosDeLaTablaDeSesiones,
					columnaDelModo, "orden")
			},
		},
		{
			nombre: "las-doce-filas-de-los-dos-modos",
			ejecucion: conCambios(comoElJob, func(e *ejecucionConUmbrales) {
				e.dosModos, e.conLaSinBinarioNiServidor = true, true
				e.conExpresiones, e.enHerramienta.conExpresiones = 2, 1
				e.duracion, e.duracionEnHerramienta, e.duracionSinModo, e.objetivo = 544, 612, 31, 900
			}),
			filas: [][]string{
				{"`sin_activar:claude-sonnet-5-5:orden`", "0 de 54 (0,0 %)", "≤ 0,0 %", "sí", "sí"},
				{"`afirma_lo_no_leido:claude-sonnet-5-5:orden`", "0 de 54 (0,0 %)", "≤ 0,0 %", "sí", "sí"},
				{"`cuenta_su_proceso:claude-sonnet-5-5:orden`", "0 de 54 (0,0 %)", "≤ 0,0 %", "sí", "no: solo se publica"},
				{"`sin_activar:claude-sonnet-5-5:herramienta`", "0 de 54 (0,0 %)", "≤ 0,0 %", "sí", "sí"},
				{"`afirma_lo_no_leido:claude-sonnet-5-5:herramienta`", "0 de 54 (0,0 %)", "≤ 0,0 %", "sí", "sí"},
				{
					"`cuenta_su_proceso:claude-sonnet-5-5:herramienta`", "0 de 54 (0,0 %)", "≤ 0,0 %", "sí",
					"no: solo se publica",
				},
				{"`medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`", "0 de 212 (0,0 %)", "≤ 0,0 %", "sí", "sí"},
				{"`medida_del_juez:afirma_lo_no_leido:correctos_marcados`", "0 de 47 (0,0 %)", "≤ 0,0 %", "sí", "sí"},
				{"`duracion_de_las_sesiones:orden`", "544", "≤ 900", "sí", "sí"},
				{"`duracion_de_las_sesiones:herramienta`", "612", "≤ 900", "sí", "sí"},
				{"`duracion_del_juez:orden`", "0", "≤ 900", "sí", "sí"},
				{"`duracion_del_juez:herramienta`", "0", "≤ 900", "sí", "sí"},
			},
			exigir: exigirLaColumnaDelModo,
		},
		{
			nombre: "los-dos-sin-cumplir",
			ejecucion: conCambios(comoElJob, func(e *ejecucionConUmbrales) {
				e.sinActivar, e.conExpresiones, e.duracion, e.objetivo = 1, 3, 901, 900
			}),
			filas: [][]string{
				{"`sin_activar:claude-sonnet-5-5:orden`", "1 de 54 (1,9 %)", "≤ 0,0 %", "no", "sí"},
				{"`afirma_lo_no_leido:claude-sonnet-5-5:orden`", "0 de 54 (0,0 %)", "≤ 0,0 %", "sí", "sí"},
				{"`cuenta_su_proceso:claude-sonnet-5-5:orden`", "0 de 54 (0,0 %)", "≤ 0,0 %", "sí", "no: solo se publica"},
				{"`medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`", "0 de 212 (0,0 %)", "≤ 0,0 %", "sí", "sí"},
				{"`medida_del_juez:afirma_lo_no_leido:correctos_marcados`", "0 de 47 (0,0 %)", "≤ 0,0 %", "sí", "sí"},
				{"`duracion_de_las_sesiones:orden`", "901", "≤ 900", "no", "sí"},
				{"`duracion_del_juez:orden`", "0", "≤ 900", "sí", "sí"},
			},
		},
		{
			nombre:    "sin-umbrales",
			ejecucion: ejecucionConUmbrales{queDeciden: 1, sinJuez: true, duracion: 544},
		},
		{
			nombre: "con-redacciones-modificadas",
			ejecucion: ejecucionConUmbrales{
				queDeciden: 1, informativas: 1, conHaiku: true, duracion: 544,
				anadidoALaEval:         hallazgosDeLaEval01 + redaccionesModificadasDeLaEval20,
				prefijoDeLasRespuestas: lineaConLaCitaDel118() + "\n\n",
			},
			filas: [][]string{
				{"`sin_activar:claude-sonnet-5-5:orden`", "0 de 6 (0,0 %)", "≤ 0,0 %", "sí", "sí"},
				{"`afirma_lo_no_leido:claude-sonnet-5-5:orden`", "0 de 6 (0,0 %)", "≤ 0,0 %", "sí", "sí"},
				{"`cuenta_su_proceso:claude-sonnet-5-5:orden`", "0 de 6 (0,0 %)", "≤ 0,0 %", "sí", "no: solo se publica"},
				{"`medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`", "0 de 212 (0,0 %)", "≤ 0,0 %", "sí", "sí"},
				{"`medida_del_juez:afirma_lo_no_leido:correctos_marcados`", "0 de 47 (0,0 %)", "≤ 0,0 %", "sí", "sí"},
				{"`duracion_del_juez:orden`", "0", "≤ 900", "sí", "sí"},
			},
			formas: formaDeVersionObsoleta + ", " + formaDeVersionObsoleta + " " + redaccionDelArticulo118 + ", " +
				formaDeVersionObsoleta + " " + redaccionDeLaDA3,
			encontradas: redaccionDelArticulo118,
			ausentes:    redaccionDeLaDA3,
		},
		{
			// Una con sí en cuenta_su_proceso, que no cumple su umbral, y una sin
			// juzgar, que sigue en el total de los tres.
			nombre:    "con-los-votos-del-juez",
			ejecucion: ejecucionConUmbrales{queDeciden: 1, informativas: 1, conHaiku: true},
			grabados:  conVotos,
			filas: [][]string{
				{"`sin_activar:claude-sonnet-5-5:orden`", "0 de 6 (0,0 %)", "≤ 0,0 %", "sí", "sí"},
				{"`afirma_lo_no_leido:claude-sonnet-5-5:orden`", "0 de 6 (0,0 %)", "≤ 0,0 %", "sí", "sí"},
				{"`cuenta_su_proceso:claude-sonnet-5-5:orden`", "1 de 6 (16,7 %)", "≤ 0,0 %", "no", "no: solo se publica"},
				{"`medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`", "0 de 212 (0,0 %)", "≤ 0,0 %", "sí", "sí"},
				{"`medida_del_juez:afirma_lo_no_leido:correctos_marcados`", "0 de 47 (0,0 %)", "≤ 0,0 %", "sí", "sí"},
				{"`duracion_del_juez:orden`", "0", "≤ 900", "sí", "sí"},
			},
			juez: seccionConVotos,
		},
		{
			// Sí, sí y no: ninguna marcada. La clase que solo se publica es la
			// de jurisprudencia, y la tabla de los votos lleva «Sentencia».
			nombre:       "con-los-votos-de-un-juez-con-sentencia",
			ejecucion:    ejecucionConUmbrales{queDeciden: 1, informativas: 1, conHaiku: true},
			conSentencia: true,
			grabados:     conSentencia,
			filas: [][]string{
				{"`sin_activar:claude-sonnet-5-5:orden`", "0 de 6 (0,0 %)", "≤ 0,0 %", "sí", "sí"},
				{"`afirma_lo_no_leido:claude-sonnet-5-5:orden`", "0 de 6 (0,0 %)", "≤ 0,0 %", "sí", "sí"},
				{"`afirma_que_existe:claude-sonnet-5-5:orden`", "0 de 6 (0,0 %)", "≤ 0,0 %", "sí", "no: solo se publica"},
				{"`medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`", "0 de 212 (0,0 %)", "≤ 0,0 %", "sí", "sí"},
				{"`medida_del_juez:afirma_lo_no_leido:correctos_marcados`", "0 de 47 (0,0 %)", "≤ 0,0 %", "sí", "sí"},
				{"`duracion_del_juez:orden`", "0", "≤ 900", "sí", "sí"},
			},
			juez: seccionConSentencia,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			escribir := escribirEjecucionConVotos
			if caso.conSentencia {
				escribir = escribirEjecucionConSentencia
			}

			leido, _ := escribir(t, caso.ejecucion, caso.grabados)

			exigirLineas(t, seccionDelInforme(t, leido.md, "Cabecera"),
				"Duración de las sesiones: "+strconv.Itoa(caso.ejecucion.duracionDeLasTandas())+" s")

			sinVotos := seccionDelJuezSinVotos
			if caso.ejecucion.sinJuez {
				sinVotos = seccionSinJuez
			}

			assert.Equal(t, cmp.Or(caso.juez, sinVotos), seccionDelInforme(t, leido.md, "Juez"))

			exigirCadaCelda(t, seccionDelInforme(t, leido.md, "Tasas por eval"), encabezadosDeLaTablaDeTasas,
				"Formas exigidas", cmp.Or(caso.formas, "ninguna"))
			exigirCadaCelda(t, seccionDelInforme(t, leido.md, "Sesiones"), encabezadosDeLaTablaDeSesiones,
				columnaDeRedaccionesEncontradas, cmp.Or(caso.encontradas, "ninguna"))
			exigirCadaCelda(t, seccionDelInforme(t, leido.md, "Sesiones"), encabezadosDeLaTablaDeSesiones,
				columnaDeRedaccionesAusentes, cmp.Or(caso.ausentes, "ninguna"))

			if caso.exigir != nil {
				caso.exigir(t, leido)
			}

			if caso.filas == nil {
				exigirSinUmbrales(t, leido)

				return
			}

			assert.Equal(t, tablaEscrita(encabezadosDeLaTablaDeUmbrales, caso.filas...),
				seccionDelInforme(t, leido.md, "Umbrales"))
		})
	}
}

// encabezadosDeLaTablaDeVotos y encabezadosDeLaTablaSinJuzgar son los de las dos
// tablas de la sección «Juez» de informe.md (contracts/informe-del-job.md §7 de
// H24); y encabezadosDeLaTablaConSentencia, los de la de los votos cuando alguno
// lleva sentencia, que solo cambian el octavo
// (contracts/juez-de-jurisprudencia.md §4 de H25).
var (
	encabezadosDeLaTablaDeVotos = []string{
		"Sesión", "Clase", "Voto", "Nulo", "Respuesta", "Frase", "En la respuesta", "Precepto", "Motivo", "Marcada",
	}
	encabezadosDeLaTablaConSentencia = []string{
		"Sesión", "Clase", "Voto", "Nulo", "Respuesta", "Frase", "En la respuesta", "Sentencia", "Motivo", "Marcada",
	}
	encabezadosDeLaTablaSinJuzgar = []string{"Sesión", "Motivo"}
)

// votosDelMarkdown son los votos grabados del caso con-los-votos-del-juez de
// TestInformeMarkdownDeLosUmbrales, de cuatro respuestas del modo orden de una
// ejecución de dos evals, y la sección «Juez» que dan (contracts/informe-del-job.md
// §7 de H24): la de «sí, sí y no», con sus tres votos; la de un sí nulo y su
// repetición, que dice no, los dos con el número 1; la que tiene sí en
// cuenta_su_proceso, marcada en ella; y la que queda sin juzgar, que va en su
// tabla y no en la de los votos. Cada voto da una fila por clase, en el orden
// de las sesiones, de las clases y de los votos: lo que el juez dio, tal cual,
// con «—» en el campo que dejó vacío o que su clase no tiene, y «sí» o «no» en
// si el voto es nulo, si su frase está en la respuesta y si la respuesta quedó
// marcada en la clase.
func votosDelMarkdown(t *testing.T) ([]votosDeUnaRespuesta, string) {
	t.Helper()

	siSiYNo := sesionSintetica(1, modeloSonnet55, 1)
	conUnNulo := sesionSintetica(1, modeloSonnet55, 2)
	conProceso := sesionSintetica(2, modeloSonnet55, 1)
	sinJuzgar := sesionSintetica(2, modeloSonnet55, 2)

	queNo := votoDeLasDosClases(t, afirmaQueNo(), cuentaQueNo())
	marca := func(frase string) grabacion { return votoDeLasDosClases(t, afirmaQueSi(frase), cuentaQueNo()) }

	grabados := []votosDeUnaRespuesta{
		{
			sesion: siSiYNo, parrafo: parrafoDeLosArticulos22Y24,
			votos: []grabacion{marca(fraseDelArticulo22), marca(fraseDelArticulo24), queNo},
		},
		{sesion: conUnNulo, parrafo: parrafoDeLosArticulos23Y25, votos: []grabacion{marca(fraseQueNoEsta), queNo}},
		{
			sesion: conProceso, parrafo: parrafoDelProcesoUno,
			votos: []grabacion{votoDeLasDosClases(t, afirmaQueNo(), cuentaQueSi(parrafoDelProcesoUno))},
		},
		{sesion: sinJuzgar, parrafo: parrafoLentoDeOrden, votos: []grabacion{{err: errTopeDelVoto}}},
	}

	// Las filas de un voto: la de afirma_lo_no_leido cuando dice sí, con su
	// frase y si está en la respuesta, y la de cada clase cuando dice no.
	siDeAfirma := func(sesion, voto, nulo, frase, esta string) []string {
		return []string{
			sesion, "afirma_lo_no_leido", voto, nulo, "si", frase, esta, "Art. 22 de la Ley 39/2015",
			"Dice de qué trata un artículo que ninguna herramienta devolvió.", "no",
		}
	}
	noDeAfirma := func(sesion, voto, nulo string) []string {
		return []string{
			sesion, "afirma_lo_no_leido", voto, nulo, "no", "—", "no", "—", "Todo lo que expone está en el texto devuelto.", "no",
		}
	}
	noDeCuenta := func(sesion, voto, nulo string) []string {
		return []string{
			sesion, "cuenta_su_proceso", voto, nulo, "no", "—", "no", "—", "No cuenta comprobaciones ni nombra herramientas.",
			"no",
		}
	}

	seccion := "Modelo: claude-opus-5-5\n\nVersión de Claude Code: 2.1.289\n\nVotos:\n\n" +
		tablaEscrita(encabezadosDeLaTablaDeVotos,
			siDeAfirma(siSiYNo, "1", "no", "El artículo 22 regula la suspensión del plazo", "sí"),
			siDeAfirma(siSiYNo, "2", "no", "el artículo 24, el silencio administrativo", "sí"),
			noDeAfirma(siSiYNo, "3", "no"),
			noDeCuenta(siSiYNo, "1", "no"),
			noDeCuenta(siSiYNo, "2", "no"),
			noDeCuenta(siSiYNo, "3", "no"),
			siDeAfirma(conUnNulo, "1", "sí", "El artículo 23 regula la ampliación del plazo", "no"),
			noDeAfirma(conUnNulo, "1", "no"),
			noDeCuenta(conUnNulo, "1", "sí"),
			noDeCuenta(conUnNulo, "1", "no"),
			noDeAfirma(conProceso, "1", "no"),
			[]string{
				conProceso, "cuenta_su_proceso", "1", "no", "si", "He comprobado la redacción del artículo 21 y no ha cambiado.",
				"sí", "—", "Cuenta que hizo una comprobación y lo que encontró.", "sí",
			}) +
		"\n\nRespuestas sin juzgar:\n\n" +
		tablaEscrita(encabezadosDeLaTablaSinJuzgar, []string{sinJuzgar, "voto 1: tope de 35 s agotado"})

	return grabados, seccion
}

// votosDelMarkdownConSentencia son los votos grabados del caso
// con-los-votos-de-un-juez-con-sentencia de TestInformeMarkdownDeLosUmbrales, de
// una respuesta del modo orden de una ejecución de dos evals, y la sección
// «Juez» que dan con el juez cuyo esquema lleva sentencia
// (contracts/juez-de-jurisprudencia.md §4 de H25): la de «sí, sí y no», con sus
// tres votos en cada clase. La octava columna es «Sentencia»: la que el juez
// dio en cada sí, y «—» en la del voto que dice no, que va vacía, y en la de
// afirma_que_existe, que no la tiene.
func votosDelMarkdownConSentencia(t *testing.T) ([]votosDeUnaRespuesta, string) {
	t.Helper()

	siSiYNo := sesionSintetica(1, modeloSonnet55, 1)

	marca := func(frase string) grabacion {
		return votoConSentencia(t, afirmaConSentenciaQueSi(frase), existeQueNo())
	}

	grabados := []votosDeUnaRespuesta{{
		sesion: siSiYNo, parrafo: parrafoDeLaSentencia,
		votos: []grabacion{
			marca(fraseDeLaNulidad), marca(fraseDeLasClausulas),
			votoConSentencia(t, afirmaConSentenciaQueNo(), existeQueNo()),
		},
	}}

	siDeAfirma := func(voto, frase string) []string {
		return []string{
			siSiYNo, "afirma_lo_no_leido", voto, "no", "si", frase, "sí", "STS 241/2013, de 9 de mayo",
			"La respuesta dice qué declaró la sentencia, y su texto no está ni en la pregunta ni en lo que devolvió " +
				"la herramienta.",
			"no",
		}
	}
	noDeExiste := func(voto string) []string {
		return []string{
			siSiYNo, "afirma_que_existe", voto, "no", "no", "—", "no", "—",
			"La respuesta lleva la línea de la sentencia no comprobada y no dice que exista.", "no",
		}
	}

	seccion := "Modelo: claude-opus-5-5\n\nVersión de Claude Code: 2.1.289\n\nVotos:\n\n" +
		tablaEscrita(encabezadosDeLaTablaConSentencia,
			siDeAfirma("1", "declaró la nulidad de las cláusulas suelo por falta de transparencia"),
			siDeAfirma("2", "declaró la nulidad de las cláusulas suelo"),
			[]string{
				siSiYNo, "afirma_lo_no_leido", "3", "no", "no", "—", "no", "—",
				"La respuesta no dice nada de lo que declara la sentencia.", "no",
			},
			noDeExiste("1"), noDeExiste("2"), noDeExiste("3")) +
		"\n\nRespuestas sin juzgar: ninguna"

	return grabados, seccion
}

// exigirLaColumnaDelModo exige, del caso las-doce-filas-de-los-dos-modos de
// TestInformeMarkdownDeLosUmbrales, la columna «Modo» de las tablas de las
// series y de las sesiones de informe.md, con lo que cada
// una dice en informe.json (contracts/evals-en-dos-modos.md §5 y §5.3 de H21):
// en las de la eval 02, que nadie altera, «orden» y «herramienta», con el modo
// también en el modelo de la serie del modo herramienta; en las de la eval sin
// binario ni servidor, «—»; y, en la sección de una sesión del modo herramienta,
// la invocación del
// servidor sin marca y la llamada con ella, y ninguna marcada en la del modo
// orden.
func exigirLaColumnaDelModo(t *testing.T, leido informeLeido) {
	t.Helper()

	const (
		enHerramienta = modeloSonnet55 + " (herramienta)"
		sinBinario    = 19
	)

	exigirLineas(t, seccionDelInforme(t, leido.md, "Tasas por eval"),
		filaDeTabla(encabezadosDeLaTablaDeTasas...),
		filaDeTabla(ficheroSintetico(2), modeloSonnet55, "orden", "sí", "sí", "ninguna", "3 de 3", "llega al umbral"),
		filaDeTabla(ficheroSintetico(2), enHerramienta, "herramienta", "sí", "sí", "ninguna", "3 de 3", "llega al umbral"),
		filaDeTabla(ficheroSintetico(sinBinario), modeloSonnet55, celdaSinModo, "sí", "sí", "ninguna", "3 de 3",
			"llega al umbral"),
		filaDeTabla(ficheroSintetico(sinBinario), modeloHaiku45, celdaSinModo, "no", "sí", "ninguna", "3 de 3",
			"llega al umbral"))

	sesiones := map[string]Modo{
		sesionSinteticaEn(ModoOrden, 2, modeloSonnet55, 1):       ModoOrden,
		sesionSinteticaEn(ModoHerramienta, 2, modeloSonnet55, 1): ModoHerramienta,
		sesionSintetica(sinBinario, modeloSonnet55, 1):           "",
	}
	for sesion, modo := range sesiones {
		assert.Equal(t, modo, resultadoDeLaSesion(t, leido.informe, sesion).Modo, "modo de %s", sesion)
		assert.JSONEq(t, cadenaJSON(t, string(modo)), string(resultadoEscrito(t, leido, sesion).Modo), "modo de %s", sesion)
		assert.Equal(t, cmp.Or(string(modo), celdaSinModo), celdaDeLaSesion(t, leido.md, sesion, columnaDelModo),
			"celda del modo de %s", sesion)
	}

	deHerramienta := seccionDelInforme(t, leido.md, "Sesión "+sesionSinteticaEn(ModoHerramienta, 2, modeloSonnet55, 1))
	exigirLineas(t, deHerramienta,
		filaDeTabla(encabezadosDeLaTablaDeInvocaciones...),
		filaDeTabla(ordenDelServidor, "0", "sin conexiones", "no"),
		filaDeTabla(llamadaDelArticulo21, "0", "sin conexiones", "sí"))

	deOrden := seccionDelInforme(t, leido.md, "Sesión "+sesionSinteticaEn(ModoOrden, 2, modeloSonnet55, 1))
	exigirLineas(t, deOrden, filaDeTabla(encabezadosDeLaTablaDeInvocaciones...))
	assert.NotContains(t, deOrden, " | sí |", "ninguna invocación de la sesión del modo orden es una llamada")
}

// exigirSinUmbrales exige lo que publica el informe de una skill sin juez ni
// objetivo de duración (FR-006 de H7.3; FR-037 de H24): umbrales es
// una lista vacía, no null, y su sección de informe.md, el párrafo «ninguno».
func exigirSinUmbrales(t *testing.T, leido informeLeido) {
	t.Helper()

	assert.Empty(t, leido.informe.Umbrales)
	assert.Equal(t, "[]", compacto(t, leido.crudo.Umbrales), "umbrales es una lista vacía, no null")
	assert.Equal(t, "ninguno", seccionDelInforme(t, leido.md, "Umbrales"))
}

// Lo que TestInformeEnDosModos lee del repositorio sin cambiarlo (research.md
// D24, V33 de H21; FR-045, FR-048).
const (
	// guionDelInformeFinal es el guion del workflow que escribe con informe.json
	// la sección de evals del informe final, lo único que lee la persona.
	guionDelInformeFinal = "../../scripts/workflow/informe.sh"

	// Las marcas de texto entre las que está el programa jq de seccion_evals,
	// que escribe la tabla de tasas, y aquellas entre las que está el de
	// recuentos_y_umbrales, que escribe la de los recuentos de expresiones.
	principioDelProgramaDeTasas     = `jq -r --argjson cambios "$cambios" --arg cabeza "$cabeza" '` + "\n"
	finalDelProgramaDeTasas         = `' "$f"` + "\n"
	principioDelProgramaDeRecuentos = `jq -r "$jq_umbral"'` + "\n"
	finalDelProgramaDeRecuentos     = `' "$1"` + "\n"

	// skillSinUmbrales y evalsDeLaSkillSinUmbrales son la skill que no tiene
	// juez ni objetivo de duración y sus evals del repositorio.
	skillSinUmbrales          = "legal-core"
	evalsDeLaSkillSinUmbrales = "../../evals/legal-core"
)

// Las líneas del programa jq de seccion_evals y del de recuentos_y_umbrales de
// guionDelInformeFinal que aplican la regla de research.md V33 de H21, sin su
// sangrado: la tabla de tasas tiene una columna por el modelo que decide, por
// cada modelo de los informativos y por cada modelo de tasas; cada serie va a la
// fila de su eval, con un sufijo si su pregunta es ampliada o está fuera del
// plan; la fila se marca informativa si tiene una serie planificada que no
// decide cuyo modelo es exactamente el que decide; cada celda es la primera
// serie de la fila con el modelo de su columna, con ✗ si decide y no pasa; y
// cada recuento de expresiones da una fila con su modelo. Desde H24 el informe
// no lleva ese recuento, y el guion, que no cambia, lee su clave con un valor
// por omisión (research V10 de H24): la línea sigue en él y no escribe nada.
var (
	lineasDelProgramaDeTasas = []string{
		`| ([$inf.modelo_que_decide] + $inf.modelos_informativos + [$inf.tasas[].modelo]) | ` +
			`reduce .[] as $m ([]; if index([$m]) then . else . + [$m] end) | . as $modelos`,
		`| [$inf.tasas[] | .fila = (.eval + (if .pregunta_ampliada then " (prueba de red)" ` +
			`elif (.planificada | not) then " (fuera del plan)" else "" end))]`,
		`| group_by(.fila) | map({fila: .[0].fila, eval: .[0].eval, series: .}) as $filas`,
		`informativa: ([.series[] | select(.modelo == $inf.modelo_que_decide and .planificada and (.decide | not))] ` +
			`| length > 0)}] as $filas`,
		`+ ([$modelos[] as $m | ([$r.series[] | select(.modelo == $m)][0]`,
		`| if . == null then "—" else "\(.pasan)/\(.sesiones)" + (if .decide and (.pasa | not) then " ✗" else "" end) ` +
			`end)] | join(" | "))`,
	}

	lineasDelProgramaDeRecuentos = []string{
		"(.expresiones_prohibidas_por_modelo[] | \"| `\\(.modelo)` | \\(.con_alguna) | \\(.respuestas) | " +
			"\\(pct(.con_alguna; .respuestas)) |\"), \"\"",
	}
)

// TestInformeEnDosModos fija el informe de un job que mide los dos modos y la
// eval sin binario ni servidor (contracts/evals-en-dos-modos.md §5 y §8 de H21;
// data-model §6 y §10; research D19, D24, V33; FR-044, FR-045, FR-047, FR-048,
// FR-080; SC-011; US5-2, US5-5, US5-6), con ejecuciones sintéticas de
// ejecucionConUmbrales escritas con EscribirInforme en t.TempDir():
//
//   - tasas lleva una serie por eval, modelo y modo, en el orden del contrato
//     —las del modo orden, las del modo herramienta y las de la eval sin binario
//     ni servidor—, con el id como modelo en el modo orden y en la eval sin
//     binario ni servidor y con «<id> (herramienta)» en el modo herramienta;
//   - una serie que decide y no llega al umbral solo en el modo herramienta pone
//     el veredicto en fallo, con el motivo que nombra su modelo y con él su modo,
//     aunque la misma eval pase en el modo orden y todos los umbrales se cumplan;
//     y lo mismo la serie de la eval sin binario ni servidor;
//   - con la regla que aplica cada programa jq de scripts/workflow/informe.sh
//     —cuyas líneas siguen siendo las de hoy en el guion, que se lee sin
//     ejecutarlo—, ninguna pareja de fila y modelo se repite en tasas: cada serie
//     de cada modo tiene su celda, con su ✗ la que no pasa, y la fila de la eval
//     informativa conserva su marca;
//   - y legal-core, con las evals del repositorio, sin juez, y el objetivo de la
//     definición del job, que es 0, publica umbrales como [] y las series de sus
//     dos modos y de su eval sin binario ni servidor, que son lo que decide.
func TestInformeEnDosModos(t *testing.T) {
	t.Parallel()

	// enDosModos es una ejecución del job con dos evals que deciden, una
	// informativa y la eval sin binario ni servidor, con los dos modelos.
	enDosModos := ejecucionConUmbrales{
		queDeciden: 2, informativas: 1, conHaiku: true, dosModos: true, conLaSinBinarioNiServidor: true,
	}

	t.Run("una-serie-falla-solo-en-el-modo-herramienta", func(t *testing.T) {
		t.Parallel()

		copia := armarEjecucionConUmbrales(t, enDosModos)

		sinCita := []string{
			sesionSinteticaEn(ModoHerramienta, 1, modeloSonnet55, 2),
			sesionSinteticaEn(ModoHerramienta, 1, modeloSonnet55, 3),
		}
		for _, sesion := range sinCita {
			dir := filepath.Join(copia, "sesiones", sesion)
			escribirEnLaCopia(t, dir, "sesion.jsonl",
				sustituirDosVeces(t, contenidoDeLaSesion(t, dir, "sesion.jsonl"), citaDeLaRespuesta, ""))
		}

		leido := informeDeLaCopia(t, copia, enDosModos.ajustar)

		enHerramienta := modeloSonnet55 + " (herramienta)"

		assert.Equal(t, tasasDeLosDosModos(func(tasa *TasaDelInforme) {
			if tasa.Eval == ficheroSintetico(1) && tasa.Modelo == enHerramienta {
				tasa.Pasan, tasa.Pasa = 1, false
			}
		}), leido.informe.Tasas)

		exigirMotivosDeLaRaiz(t, leido,
			ficheroSintetico(1)+" con claude-sonnet-5-5 (herramienta): pasan 1 de 3, y el umbral es 2",
			sinCita[0]+": cita ausente: "+textoDeLaCita21,
			sinCita[1]+": cita ausente: "+textoDeLaCita21)
		assert.Equal(t, VeredictoFallo, leido.informe.Veredicto)

		// Ningún umbral falla: lo que pone el veredicto en fallo es la serie.
		exigirUmbrales(t, leido, umbralesCumplidosDeNueve())

		tabla := tablaDelInformeFinal(t, leido.informe)

		assert.Equal(t, []string{modeloSonnet55, modeloHaiku45, enHerramienta, modeloHaiku45 + " (herramienta)"},
			tabla.modelos, "una columna por modelo y modo")
		assert.Len(t, tabla.celdas, len(leido.informe.Tasas), "cada serie de cada modo tiene su celda")
		assert.Equal(t, "3/3", tabla.celdas[celdaDelInformeFinal{fila: ficheroSintetico(1), modelo: modeloSonnet55}],
			"la serie del modo orden de la eval pasa")
		assert.Equal(t, "1/3 ✗", tabla.celdas[celdaDelInformeFinal{fila: ficheroSintetico(1), modelo: enHerramienta}],
			"la serie del modo herramienta de la misma eval no pasa, con su ✗")
		assert.Equal(t, "3/3", tabla.celdas[celdaDelInformeFinal{fila: ficheroSintetico(4), modelo: modeloSonnet55}],
			"la serie de la eval sin binario ni servidor, en la columna del id")
		assert.Equal(t, []string{ficheroSintetico(3)}, tabla.informativas, "la eval informativa conserva su marca")
	})

	t.Run("la-serie-de-la-eval-sin-binario-ni-servidor-falla", func(t *testing.T) {
		t.Parallel()

		leido := escribirEjecucionConUmbrales(t, conCambios(enDosModos, func(e *ejecucionConUmbrales) {
			e.sinBinarioAlteradas = 2
		}))

		assert.Equal(t, tasasDeLosDosModos(func(tasa *TasaDelInforme) {
			if tasa.Eval == ficheroSintetico(4) && tasa.Modelo == modeloSonnet55 {
				tasa.Pasan, tasa.Pasa = 1, false
			}
		}), leido.informe.Tasas)

		motivos := []string{ficheroSintetico(4) + " con claude-sonnet-5-5: pasan 1 de 3, y el umbral es 2"}

		for vez := 1; vez <= 2; vez++ {
			sesion := sesionSintetica(4, modeloSonnet55, vez)

			resultado := resultadoDeLaSesion(t, leido.informe, sesion)
			require.NotEmpty(t, resultado.Motivos, "%s no pasa", sesion)

			for _, motivo := range resultado.Motivos {
				motivos = append(motivos, sesion+": "+motivo)
			}
		}

		exigirMotivosDeLaRaiz(t, leido, motivos...)
		assert.Equal(t, VeredictoFallo, leido.informe.Veredicto)

		// Sus sesiones no entran en ninguna medida: ningún umbral falla.
		exigirUmbrales(t, leido, umbralesCumplidosDeNueve())

		tabla := tablaDelInformeFinal(t, leido.informe)
		assert.Len(t, tabla.celdas, len(leido.informe.Tasas), "cada serie tiene su celda")
		assert.Equal(t, "1/3 ✗", tabla.celdas[celdaDelInformeFinal{fila: ficheroSintetico(4), modelo: modeloSonnet55}])
	})

	t.Run("legal-core", func(t *testing.T) {
		t.Parallel()
		exigirElInformeDeLaSkillSinUmbrales(t)
	})

	t.Run("sin-poder-saber-el-modo", func(t *testing.T) {
		t.Parallel()
		exigirLaSesionSinModoConocido(t)
	})

	t.Run("las-lineas-del-guion-del-informe-final", func(t *testing.T) {
		t.Parallel()

		guion := contenidoDeLaSesion(t, filepath.Dir(guionDelInformeFinal), filepath.Base(guionDelInformeFinal))

		exigirLasLineasDelPrograma(t, guion, principioDelProgramaDeTasas, finalDelProgramaDeTasas, lineasDelProgramaDeTasas)
		exigirLasLineasDelPrograma(t, guion, principioDelProgramaDeRecuentos, finalDelProgramaDeRecuentos,
			lineasDelProgramaDeRecuentos)
	})
}

// umbralesCumplidosDeNueve son los umbrales de la ejecución de
// TestInformeEnDosModos, con tres evals que activan la skill, una de ellas
// informativa, sin ninguna respuesta del modelo que decide sin activar ni
// marcada por el juez: 0 de 9 en cada modo, con los de la medida del juez y sus
// votos en 0 s. No tiene objetivo de duración.
func umbralesCumplidosDeNueve() []Umbral {
	return umbralesConJuez(modeloSonnet55, nil, enOrden(9, 0), enHerramienta(9, 0))
}

// tasasDeLosDosModos son las tasas de la ejecución de TestInformeEnDosModos con
// todas sus sesiones pasando, en el orden de contracts/evals-en-dos-modos.md §5
// de H21, escritas a mano, y con lo que cambie cambiar en cada una: las del modo
// orden —las dos evals que deciden, con los dos modelos, y la informativa, solo
// con el que decide—, las mismas del modo herramienta, con el modo en su modelo,
// y las de la eval sin binario ni servidor, sin modo.
func tasasDeLosDosModos(cambiar func(tasa *TasaDelInforme)) []TasaDelInforme {
	tasa := func(numero int, modelo string, modo Modo, decide bool) TasaDelInforme {
		return TasaDelInforme{
			Eval: ficheroSintetico(numero), Modelo: modelo, Modo: modo, Planificada: true, Decide: decide,
			Formas: sinFormasExigidas, Sesiones: 3, Pasan: 3, Pasa: true,
		}
	}

	tasas := []TasaDelInforme{
		tasa(1, "claude-sonnet-5-5", ModoOrden, true),
		tasa(1, "claude-haiku-4-5-20251001", ModoOrden, false),
		tasa(2, "claude-sonnet-5-5", ModoOrden, true),
		tasa(2, "claude-haiku-4-5-20251001", ModoOrden, false),
		tasa(3, "claude-sonnet-5-5", ModoOrden, false),
		tasa(1, "claude-sonnet-5-5 (herramienta)", ModoHerramienta, true),
		tasa(1, "claude-haiku-4-5-20251001 (herramienta)", ModoHerramienta, false),
		tasa(2, "claude-sonnet-5-5 (herramienta)", ModoHerramienta, true),
		tasa(2, "claude-haiku-4-5-20251001 (herramienta)", ModoHerramienta, false),
		tasa(3, "claude-sonnet-5-5 (herramienta)", ModoHerramienta, false),
		tasa(4, "claude-sonnet-5-5", "", true),
		tasa(4, "claude-haiku-4-5-20251001", "", false),
	}

	for posicion := range tasas {
		cambiar(&tasas[posicion])
	}

	return tasas
}

// celdaDelInformeFinal es una celda de la tabla de tasas que escribe
// scripts/workflow/informe.sh: la de una fila y un modelo.
type celdaDelInformeFinal struct {
	fila, modelo string
}

// tablaDeTasasDelInformeFinal es la tabla de tasas que escribe
// scripts/workflow/informe.sh con un informe: los modelos, uno por columna y en
// su orden; el texto de cada celda con serie; y las filas que marca informativa,
// en el orden de tasas.
type tablaDeTasasDelInformeFinal struct {
	modelos      []string
	celdas       map[celdaDelInformeFinal]string
	informativas []string
}

// tablaDelInformeFinal aplica al informe la regla del programa jq de
// seccion_evals de scripts/workflow/informe.sh (research.md V33 de H21), sin
// ejecutarlo: hay una columna por el modelo que decide, por cada uno de los
// informativos y por cada modelo de tasas, sin repetir; cada serie va a su fila
// (filaDelInformeFinal), y la celda de una fila y un modelo es la de la serie
// (textoDeLaCelda); y la fila es informativa si tiene una serie planificada que
// no decide cuyo modelo es exactamente el que decide. El guion enseña solo la
// primera serie de cada fila y modelo: una pareja repetida dejaría una serie sin
// celda, y aquí es un fallo.
func tablaDelInformeFinal(t *testing.T, informe Informe) tablaDeTasasDelInformeFinal {
	t.Helper()

	tabla := tablaDeTasasDelInformeFinal{celdas: map[celdaDelInformeFinal]string{}}

	for _, modelo := range slices.Concat([]string{informe.ModeloQueDecide}, informe.ModelosInformativos) {
		tabla.modelos = conElQueFalte(tabla.modelos, modelo)
	}

	for _, tasa := range informe.Tasas {
		tabla.modelos = conElQueFalte(tabla.modelos, tasa.Modelo)

		fila := filaDelInformeFinal(tasa)
		celda := celdaDelInformeFinal{fila: fila, modelo: tasa.Modelo}

		_, repetida := tabla.celdas[celda]
		if !assert.Falsef(t, repetida, "la pareja de fila %q y modelo %q se repite en tasas: informe.sh solo enseña "+
			"la primera serie, y la del modo %q se queda sin celda", fila, tasa.Modelo, tasa.Modo) {
			continue
		}

		tabla.celdas[celda] = textoDeLaCelda(tasa)

		if tasa.Modelo == informe.ModeloQueDecide && tasa.Planificada && !tasa.Decide {
			tabla.informativas = conElQueFalte(tabla.informativas, fila)
		}
	}

	return tabla
}

// conElQueFalte es la lista con el valor al final si no lo tenía, sea el que sea:
// como el reduce del programa jq, que no descarta ninguno.
func conElQueFalte(lista []string, valor string) []string {
	if slices.Contains(lista, valor) {
		return lista
	}

	return append(lista, valor)
}

// filaDelInformeFinal es la fila de la tabla de tasas de
// scripts/workflow/informe.sh a la que va la serie: la de su eval, con « (prueba
// de red)» si su pregunta es ampliada y « (fuera del plan)» si el plan no la
// pide.
func filaDelInformeFinal(tasa TasaDelInforme) string {
	switch {
	case tasa.PreguntaAmpliada:
		return tasa.Eval + " (prueba de red)"
	case !tasa.Planificada:
		return tasa.Eval + " (fuera del plan)"
	default:
		return tasa.Eval
	}
}

// textoDeLaCelda es lo que scripts/workflow/informe.sh escribe en la celda de
// la serie: «<pasan>/<sesiones>», con « ✗» si decide y no pasa.
func textoDeLaCelda(tasa TasaDelInforme) string {
	texto := strconv.Itoa(tasa.Pasan) + "/" + strconv.Itoa(tasa.Sesiones)
	if tasa.Decide && !tasa.Pasa {
		texto += " ✗"
	}

	return texto
}

// exigirLasLineasDelPrograma exige que el programa jq del guion que está entre
// las dos marcas de texto tenga, sin su sangrado, cada una de las líneas dadas:
// las que aplican la regla que comprueba TestInformeEnDosModos. Si una falta, el
// test la nombra: la regla que comprueba ya no sería la del informe.sh de hoy
// (research.md D24 de H21).
func exigirLasLineasDelPrograma(t *testing.T, guion, principio, final string, lineas []string) {
	t.Helper()

	_, resto, hay := strings.Cut(guion, principio)
	require.True(t, hay, "%s tiene un programa jq que empieza por %q", guionDelInformeFinal, principio)

	programa, _, hay := strings.Cut(resto, final)
	require.True(t, hay, "el programa jq de %s que empieza por %q termina en %q", guionDelInformeFinal, principio, final)

	var delPrograma []string

	for linea := range strings.Lines(programa) {
		delPrograma = append(delPrograma, strings.TrimSpace(linea))
	}

	for _, linea := range lineas {
		assert.True(t, slices.Contains(delPrograma, linea),
			"el programa jq de %s que empieza por %q ya no tiene la línea %q", guionDelInformeFinal, principio, linea)
	}
}

// exigirElInformeDeLaSkillSinUmbrales escribe el informe de legal-core con sus
// evals del repositorio, el plan del job en sus dos modos y el objetivo de
// duración que le da la definición del job, sin ninguna sesión, y exige que la
// skill siga sin umbrales (FR-045 de H21; FR-037 de H24): no tiene juez ni
// objetivo, y umbrales es [],
// también con las duraciones de sus tandas; y que sus series sean las de sus tres
// evals de modo en cada modo, con los dos modelos, y las de su eval sin binario
// ni servidor, una sola vez, con las del modelo que decide decidiendo: 14, en el
// orden del contrato.
func exigirElInformeDeLaSkillSinUmbrales(t *testing.T) {
	t.Helper()

	delJob, err := leerDefinicionDelJob(rutaDeLaDefinicionDelJob)
	require.NoError(t, err)

	ajustes, esta := delJob.PorSkill[skillSinUmbrales]
	require.True(t, esta, "la definición del job tiene los ajustes de %s", skillSinUmbrales)
	require.Zero(t, ajustes.ObjetivoDeDuracion, "%s no tiene objetivo de duración", skillSinUmbrales)

	conjunto, err := LeerConjunto(evalsDeLaSkillSinUmbrales)
	require.NoError(t, err)
	require.Empty(t, conjunto.MalFormados)
	require.Nil(t, conjunto.Juez, "%s no tiene juez", skillSinUmbrales)

	destino := t.TempDir()

	informe, err := EscribirInforme(InformeAEscribir{
		Skill:                 skillSinUmbrales,
		Evals:                 evalsDeLaSkillSinUmbrales,
		Sesiones:              t.TempDir(),
		Destino:               destino,
		ModeloQueDecide:       modeloSonnet55,
		ModelosInformativos:   []string{modeloHaiku45},
		Repeticiones:          repeticionesConUmbrales,
		Umbral:                umbralConUmbrales,
		Modos:                 []Modo{ModoOrden, ModoHerramienta},
		Commit:                commitEvaluado,
		SinPython:             filepath.Join(casosDeInforme, ficheroSinPython),
		DuracionDeLasSesiones: 2000,
		DuracionDeLosModos:    map[Modo]int{ModoOrden: 950, ModoHerramienta: 1000},
		ObjetivoDeDuracion:    ajustes.ObjetivoDeDuracion,
	})
	require.NoError(t, err)

	assert.Empty(t, informe.Umbrales)
	assert.Nil(t, informe.Juez, "%s no tiene juez en su informe", skillSinUmbrales)

	var crudo informeCrudo
	require.NoError(t, json.Unmarshal([]byte(contenidoDelInforme(t, destino, "informe.json")), &crudo))
	assert.Equal(t, "[]", compacto(t, crudo.Umbrales), "umbrales es una lista vacía, no null")
	assert.Equal(t, "null", compacto(t, crudo.Juez), "sin juez, juez es null")

	type serie struct {
		eval, modelo string
		modo         Modo
		decide       bool
	}

	deModo := []string{
		"01-territorio-municipio-cubierto.yaml", "02-territorio-municipio-no-cubierto.yaml",
		"03-no-activa-receta-de-cocina.yaml",
	}

	var esperadas []serie

	for _, modo := range []Modo{ModoOrden, ModoHerramienta} {
		sufijo := ""
		if modo == ModoHerramienta {
			sufijo = " (herramienta)"
		}

		for _, eval := range deModo {
			esperadas = append(esperadas,
				serie{eval: eval, modelo: modeloSonnet55 + sufijo, modo: modo, decide: true},
				serie{eval: eval, modelo: modeloHaiku45 + sufijo, modo: modo})
		}
	}

	esperadas = append(esperadas,
		serie{eval: ficheroSinBinarioDeLegalCore, modelo: modeloSonnet55, decide: true},
		serie{eval: ficheroSinBinarioDeLegalCore, modelo: modeloHaiku45})

	series := make([]serie, 0, len(informe.Tasas))
	for _, tasa := range informe.Tasas {
		assert.True(t, tasa.Planificada, "%s con %s es una serie del plan", tasa.Eval, tasa.Modelo)
		series = append(series, serie{eval: tasa.Eval, modelo: tasa.Modelo, modo: tasa.Modo, decide: tasa.Decide})
	}

	assert.Equal(t, esperadas, series)
	assert.Len(t, tablaDelInformeFinal(t, informe).celdas, len(esperadas), "cada serie de cada modo tiene su celda")
}

// exigirLaSesionSinModoConocido escribe el informe de una copia del caso aprobado
// con una entrada más en su directorio de sesiones que no es un directorio, y
// exige lo que el informe hace con una sesión de la que no se puede saber si
// tiene servidor.json (contracts/evals-en-dos-modos.md §3 de H21): no la toma por
// una del modo orden; queda ilegible, con el motivo de servidor.json, que nombra
// su ruta, detrás de los de eval.txt, modelo.txt y pregunta.txt y delante del de
// la sesión; y sin modo, en informe.json y en su celda de informe.md. Las demás
// sesiones conservan el suyo.
func exigirLaSesionSinModoConocido(t *testing.T) {
	t.Helper()

	const entrada = "99-no-es-un-directorio"

	copia := copiaDelCasoAprobadoConLaLista(t)
	escribirEnLaCopia(t, filepath.Join(copia, "sesiones"), entrada, "no es el directorio de una sesión\n")

	leido := informeDeLaCopia(t, copia, nil)

	resultado := resultadoDeLaSesion(t, leido.informe, entrada)
	assert.False(t, resultado.Pasa)
	assert.Empty(t, resultado.Modo, "de una sesión cuyo servidor.json no se puede mirar no se sabe el modo")
	assert.Equal(t, celdaSinModo, celdaDeLaSesion(t, leido.md, entrada, columnaDelModo))

	principios := []string{"eval.txt: ", "modelo.txt: ", "pregunta.txt: ", "servidor.json: ", ""}
	require.Len(t, resultado.Motivos, len(principios), "un motivo por fichero que no se puede leer: %q", resultado.Motivos)

	for posicion, principio := range principios {
		assert.True(t, strings.HasPrefix(resultado.Motivos[posicion], "sesión ilegible: "+principio),
			"el motivo %q empieza por %q", resultado.Motivos[posicion], "sesión ilegible: "+principio)
	}

	assert.Contains(t, resultado.Motivos[3], filepath.Join(copia, "sesiones", entrada, "servidor.json")+
		" no se puede comprobar: ")

	assert.Equal(t, ModoOrden, resultadoDeLaSesion(t, leido.informe, sesionDelArticulo21).Modo)
	assert.Equal(t, VeredictoFallo, leido.informe.Veredicto)
}

// TestEjecutarPorTandas fija cómo se reparte el plan del job, una llamada al
// repartidor por tanda, y lo que el informe recibe de ello
// (contracts/evals-en-dos-modos.md §2.2 y §5 de H21; research.md D14 de H7.3;
// FR-043, FR-047), con un repartidor de pega que no abre nada: las tandas son las
// sesiones seguidas del mismo modo, que con el plan del job son tres —la del
// modo orden, con la prueba de red, la del modo herramienta y la de la eval sin
// binario ni servidor—, y cada una se ejecuta en su llamada, en su orden; la
// duración es la suma de los segundos de las tres, cada una redondeada hacia
// arriba, y la de cada modo, la de su tanda, sin la de la eval sin binario ni
// servidor; si una tanda acaba con sesiones sin abrir por el límite de uso, las
// siguientes no se ejecutan y todas sus sesiones cuentan como sin abrir, detrás
// de las suyas; y el error de una tanda es el de la ejecución, sin ejecutar las
// siguientes.
func TestEjecutarPorTandas(t *testing.T) {
	t.Parallel()

	plan := PlanDeEvals{
		Evals:           []Eval{{Fichero: ficheroDeLaEval01}, {Fichero: ficheroSinBinarioDeBoe, SinBinarioNiServidor: true}},
		ModeloQueDecide: modeloQueDecide,
		Repeticiones:    2,
		Modos:           []Modo{ModoOrden, ModoHerramienta},
		PruebaDeRed:     true,
	}.Sesiones()

	require.Len(t, plan, 7, "el plan tiene tres sesiones del modo orden, dos del modo herramienta y dos sin modo")

	delModoOrden, delModoHerramienta, sinModo := plan[:3], plan[3:5], plan[5:]

	errDeLaTanda := errors.New("la sesión no se puede abrir")

	casos := []struct {
		nombre string

		// ejecuciones son lo que devuelve el repartidor de pega en cada llamada, y
		// errores, su error; las que faltan son la ejecución vacía, sin error.
		ejecuciones []EjecucionDeSesiones
		errores     []error

		tandas   [][]SesionPlanificada
		esperado ejecucionPorTandas
		conError bool
	}{
		{
			nombre: "tres-tandas",
			ejecuciones: []EjecucionDeSesiones{
				{Duracion: 900*time.Second + 400*time.Millisecond}, {Duracion: 12 * time.Second}, {Duracion: 3 * time.Second},
			},
			tandas: [][]SesionPlanificada{delModoOrden, delModoHerramienta, sinModo},
			esperado: ejecucionPorTandas{
				duracion: 901 + 12 + 3, porModo: map[Modo]int{ModoOrden: 901, ModoHerramienta: 12},
			},
		},
		{
			nombre: "limite-de-uso-en-la-primera-tanda",
			ejecuciones: []EjecucionDeSesiones{
				{Duracion: 40 * time.Second, SinAbrir: delModoOrden[2:]},
			},
			tandas: [][]SesionPlanificada{delModoOrden},
			esperado: ejecucionPorTandas{
				sinAbrir: plan[2:], duracion: 40, porModo: map[Modo]int{ModoOrden: 40},
			},
		},
		{
			nombre: "limite-de-uso-en-la-segunda-tanda",
			ejecuciones: []EjecucionDeSesiones{
				{Duracion: 40 * time.Second}, {Duracion: 5 * time.Second, SinAbrir: delModoHerramienta[1:]},
			},
			tandas: [][]SesionPlanificada{delModoOrden, delModoHerramienta},
			esperado: ejecucionPorTandas{
				sinAbrir: plan[4:], duracion: 45, porModo: map[Modo]int{ModoOrden: 40, ModoHerramienta: 5},
			},
		},
		{
			nombre:      "error-en-la-segunda-tanda",
			ejecuciones: []EjecucionDeSesiones{{Duracion: 40 * time.Second}},
			errores:     []error{nil, errDeLaTanda},
			tandas:      [][]SesionPlanificada{delModoOrden, delModoHerramienta},
			conError:    true,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			var tandas [][]SesionPlanificada

			ejecucion, err := ejecutarPorTandas(plan, func(tanda []SesionPlanificada) (EjecucionDeSesiones, error) {
				llamada := len(tandas)
				tandas = append(tandas, tanda)

				var (
					ejecutada EjecucionDeSesiones
					err       error
				)

				if llamada < len(caso.ejecuciones) {
					ejecutada = caso.ejecuciones[llamada]
				}

				if llamada < len(caso.errores) {
					err = caso.errores[llamada]
				}

				return ejecutada, err
			})

			assert.Equal(t, caso.tandas, tandas, "una llamada al repartidor por tanda, en su orden")

			if caso.conError {
				require.ErrorIs(t, err, errDeLaTanda)
				assert.Zero(t, ejecucion)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, caso.esperado, ejecucion)
		})
	}

	t.Run("un-solo-modo", func(t *testing.T) {
		t.Parallel()

		// El plan de un solo modo, sin eval sin binario ni servidor, es una tanda.
		deUnModo := PlanDeEvals{
			Evals: []Eval{{Fichero: ficheroDeLaEval01}}, ModeloQueDecide: modeloQueDecide, Repeticiones: 2,
		}.Sesiones()

		llamadas := 0

		ejecucion, err := ejecutarPorTandas(deUnModo, func(tanda []SesionPlanificada) (EjecucionDeSesiones, error) {
			llamadas++

			assert.Equal(t, deUnModo, tanda)

			return EjecucionDeSesiones{Duracion: time.Second}, nil
		})
		require.NoError(t, err)
		assert.Equal(t, 1, llamadas)
		assert.Equal(t, ejecucionPorTandas{duracion: 1, porModo: map[Modo]int{ModoOrden: 1}}, ejecucion)
	})

	t.Run("sin-sesiones", func(t *testing.T) {
		t.Parallel()

		ejecucion, err := ejecutarPorTandas(nil, func([]SesionPlanificada) (EjecucionDeSesiones, error) {
			assert.Fail(t, "sin sesiones no hay ninguna tanda que ejecutar")

			return EjecucionDeSesiones{}, nil
		})
		require.NoError(t, err)
		assert.Equal(t, ejecucionPorTandas{porModo: map[Modo]int{}}, ejecucion)
	})

	t.Run("segundos-hacia-arriba", func(t *testing.T) {
		t.Parallel()

		// 900,4 s son 901 y no cumplen un objetivo de 900 (research.md D14 de H7.3).
		for duracion, segundos := range map[time.Duration]int{
			0:                                      0,
			time.Nanosecond:                        1,
			900 * time.Second:                      900,
			900*time.Second + time.Nanosecond:      901,
			900*time.Second + 400*time.Millisecond: 901,
		} {
			assert.Equal(t, segundos, segundosHaciaArriba(duracion), "%s", duracion)
		}
	})
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
// del caso aprobado con la carpeta juez de una skill sintética armadas en
// t.TempDir(), sin tocar las versionadas: una sesión con el mensaje del límite de
// uso (a), con los
// reintentos por rate_limit agotados (b) o cortada por el tope durante ellos (c)
// no pasa ni falla: lleva su clase en sin_medir y como único motivo «sin medir
// por límite de uso: <clase>», queda fuera de las respuestas medidas —el total
// del umbral de las respuestas sin activar— y deja
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
// la carpeta del juez en la que la sesión del art. 21 termina con el código
// dado y su transcript sigue a su mensaje system/init con lo dado, y exige que
// quede sin medir por la clase dada, con sus reintentos por rate_limit en la
// sesión y en la raíz; que su serie quede sin medir, sin el motivo de su tasa;
// que no cuente como respuesta medida; y el veredicto fallo con el único motivo
// de la ejecución, que la nombra.
func exigirUnaSesionSinMedir(t *testing.T, codigo int, trasElInit, clase string, reintentos int) {
	t.Helper()

	copia := copiaDelCasoAprobadoConElJuez(t)
	cambiarElFinalDeLaSesion(t, filepath.Join(copia, "sesiones", sesionDelArticulo21), codigo, trasElInit)

	leido := informeDeLaCopia(t, copia, nil)

	exigirSesionSinMedir(t, leido, sesionDelArticulo21, clase, reintentos)
	assert.True(t, resultadoDeLaSesion(t, leido.informe, sesionDeNoActivacion).Pasa)

	exigirTasaSinMedir(t, leido, TasaDelInforme{
		Eval: ficheroDeLaEval01, Modelo: modeloQueDecide, Modo: ModoOrden, Planificada: true, Decide: true,
		Formas: sinFormasExigidas, Sesiones: 1, Pasan: 0, SinMedir: 1,
	})

	exigirUmbrales(t, leido, umbralesConJuez(modeloQueDecide, nil, enOrden(0, 0)))
	assert.Empty(t, leido.votos, "una sesión sin medir no se juzga")

	sinMedir := SesionSinMedir{Sesion: sesionDelArticulo21, Eval: ficheroDeLaEval01, Modelo: modeloQueDecide, Motivo: clase}
	exigirSesionesSinMedir(t, leido, sinMedir)
	exigirMotivosDeLaRaiz(t, leido, motivoEsperadoDelLimite(sinMedir))
	assert.Equal(t, VeredictoFallo, leido.informe.Veredicto)

	assert.Equal(t, reintentos, leido.informe.ReintentosPorLimiteDeRitmo)
	assert.Equal(t, strconv.Itoa(reintentos), string(leido.crudo.ReintentosPorLimiteDeRitmo))
}

// exigirLasSesionesSinAbrir escribe, con tres repeticiones y umbral 2, el informe
// de una copia del caso aprobado con la carpeta del juez en la que la serie
// del art. 21 tiene una sesión que pasa, otra con el mensaje del límite de uso y
// la tercera sin abrir tras el límite, y la de no activación, una sola sesión, que
// pasa, sin ninguna sin abrir. Exige que la serie del art. 21 cuente como sin
// medir la que tiene el mensaje y la que no se abrió, sin el motivo de su tasa ni
// el de las sesiones que faltan; que la de no activación dé los dos, porque sus
// sesiones faltan por otra causa; que solo la sesión medida cuente como
// respuesta medida; y
// que el motivo de la ejecución nombre las dos sesiones sin medir, en orden de
// sesión, detrás de los de siempre.
func exigirLasSesionesSinAbrir(t *testing.T) {
	t.Helper()

	copia := copiaDelCasoAprobadoConElJuez(t)
	sesiones := filepath.Join(copia, "sesiones")

	require.NoError(t, os.Rename(filepath.Join(sesiones, sesionDelArticulo21), filepath.Join(sesiones, sesionDeLaSerie(1))))

	conElMensaje := filepath.Join(sesiones, sesionDeLaSerie(2))
	require.NoError(t, os.CopyFS(conElMensaje, os.DirFS(filepath.Join(casosDeInforme, casoAprobado, "sesiones",
		sesionDelArticulo21))))
	cambiarElFinalDeLaSesion(t, conElMensaje, 1, mensajeResultConError(t, textoDelLimiteDeSesion))

	leido := informeDeLaCopia(t, copia, func(entradas *InformeAEscribir) {
		entradas.Repeticiones, entradas.Umbral = 3, 2
		entradas.SinAbrir = []SesionPlanificada{
			{Nombre: sesionDeLaSerie(3), Fichero: ficheroDeLaEval01, Modelo: modeloQueDecide, Modo: ModoOrden},
		}
	})

	assert.True(t, resultadoDeLaSesion(t, leido.informe, sesionDeLaSerie(1)).Pasa)
	exigirSesionSinMedir(t, leido, sesionDeLaSerie(2), sinMedirPorElMensaje, 0)
	assert.Len(t, leido.informe.Evals, 3, "la sesión que no se abrió no tiene resultado")

	exigirTasaSinMedir(t, leido, TasaDelInforme{
		Eval: ficheroDeLaEval01, Modelo: modeloQueDecide, Modo: ModoOrden, Planificada: true, Decide: true,
		Formas: sinFormasExigidas, Sesiones: 2, Pasan: 1, SinMedir: 2,
	})
	assert.Equal(t, TasaDelInforme{
		Eval: ficheroDeNoActivacion, Modelo: modeloQueDecide, Modo: ModoOrden, Planificada: true, Decide: true,
		Formas: sinFormasExigidas, Sesiones: 1, Pasan: 1,
	}, tasaDeLaSerie(t, leido.informe, ficheroDeNoActivacion, modeloQueDecide))
	exigirLineas(t, seccionDelInforme(t, leido.md, "Tasas por eval"),
		filaDeTabla(ficheroDeNoActivacion, modeloQueDecide, "orden", "sí", "sí", "ninguna", "1 de 1", "no llega al umbral"))

	exigirUmbrales(t, leido, umbralesConJuez(modeloQueDecide, nil, enOrden(1, 0)))
	assert.Len(t, leido.votos, 1, "se juzga la única respuesta medida de una eval que activa la skill")

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
// aprobado con la carpeta del juez en la que la sesión del art. 21 reintenta
// dos veces por rate_limit y una por sobrecarga, y la de no activación una por
// rate_limit, antes de terminar como en el caso aprobado. Exige que las dos se
// midan y pasen, cada una con sus reintentos por rate_limit en informe.json y en
// su fila de informe.md, y la raíz con su suma, sin ninguna sesión sin medir y
// con el veredicto aprobado (FR-033, FR-041).
func exigirLosReintentosRecuperados(t *testing.T) {
	t.Helper()

	copia := copiaDelCasoAprobadoConElJuez(t)
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

	exigirUmbrales(t, leido, umbralesConJuez(modeloQueDecide, nil, enOrden(1, 0)))
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
		filaDeTabla(esperada.Eval, esperada.Modelo, string(esperada.Modo), siONo(esperada.Decide),
			siONo(esperada.Planificada), "ninguna",
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

// copiaDelCasoAprobadoConElJuez copia el caso aprobado de TestInforme en un
// directorio temporal del test, con la carpeta juez de una skill sintética en su
// carpeta de evals (escribirElJuezDelInforme), y devuelve su ruta: con ella, su
// informe tiene los umbrales de las respuestas del modelo que decide y los del
// juez (research D13 de H24; contracts/informe-del-job.md §2 de H24).
func copiaDelCasoAprobadoConElJuez(t *testing.T) string {
	t.Helper()

	copia := t.TempDir()
	require.NoError(t, os.CopyFS(copia, os.DirFS(filepath.Join(casosDeInforme, casoAprobado))))
	escribirElJuezDelInforme(t, filepath.Join(copia, "evals"), "")

	return copia
}

// sesionDeLaSerie es el nombre de la sesión de ese número de la serie de la eval
// del art. 21 con el modelo que decide, como los nombra el plan con más de una
// repetición.
func sesionDeLaSerie(numero int) string {
	return fmt.Sprintf("%s-%s-%02d", sesionDelArticulo21, modeloQueDecide, numero)
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

// informeDeLaCopia escribe el informe de la copia de un caso, con las entradas
// del caso aprobado, las evals y las sesiones de la copia y lo que cambie
// ajustar, si no es nil, y lo lee con leerInformeEscrito. Toda ejecución recibe
// con qué votar (contracts/informe-del-job.md §1 de H24): un votante que dice no
// a todo y no mueve su reloj, que es lo que necesita la que tiene juez y lo que
// la que no lo tiene no llama; los votos que se le piden quedan en lo leído. El
// caso que juzga con otros votos pone su votante con ajustar.
func informeDeLaCopia(t *testing.T, copia string, ajustar func(entradas *InformeAEscribir)) informeLeido {
	t.Helper()

	entradas := entradasDelCaso(casoAprobado, t.TempDir())
	entradas.Evals = filepath.Join(copia, "evals")
	entradas.Sesiones = filepath.Join(copia, "sesiones")

	votante := nuevoVotanteDelInforme(t)
	votante.darA(&entradas)

	if ajustar != nil {
		ajustar(&entradas)
	}

	informe, err := EscribirInforme(entradas)
	require.NoError(t, err)

	leido := leerInformeEscrito(t, entradas.Destino, informe)
	leido.caso = copia
	leido.votos = votante.pedidos()

	return leido
}

// copiaConExpresiones copia el caso aprobado de TestInforme en un directorio
// temporal del test y devuelve su ruta. En la copia, la carpeta de evals lleva la
// lista del repositorio; la respuesta de la sesión del art. 21 lleva
// delante la transición de la memoria de consultas, y la de la eval de no
// activación, json; modeloInformativoDelCaso tiene la sesión del art. 21 del caso
// de los modelos informativos, sin ninguna expresión, y la de no activación del
// caso aprobado, con su modelo; y la copia tiene además la
// sesión de la prueba de red del caso de fuera de lo grabado, con lo dicho en
// otra conversación delante de su respuesta.
func copiaConExpresiones(t *testing.T) string {
	t.Helper()

	copia := copiaDelCasoAprobadoConLaLista(t)
	sesiones := filepath.Join(copia, "sesiones")
	anteponerALaRespuesta(t, filepath.Join(sesiones, sesionDelArticulo21), transicionDeLaMemoria+"\n\n")
	anteponerALaRespuesta(t, filepath.Join(sesiones, sesionDeNoActivacion), jsonEnLaRespuesta)

	require.NoError(t, os.CopyFS(filepath.Join(sesiones, sesionDelArticulo21ConOpus), os.DirFS(
		filepath.Join(casosDeInforme, "modelos-informativos-no-deciden", "sesiones", sesionDelArticulo21ConOpus))))
	copiarSesionConOtroModelo(t, filepath.Join(casosDeInforme, casoAprobado, "sesiones", sesionDeNoActivacion),
		filepath.Join(sesiones, sesionDeNoActivacionConOpus), modeloInformativoDelCaso)

	deLaPruebaDeRed := filepath.Join(sesiones, sesionDeLaPruebaDeRed)
	require.NoError(t, os.CopyFS(deLaPruebaDeRed, os.DirFS(filepath.Join(casosDeInforme,
		"fuera-de-lo-grabado-no-cambia-el-veredicto", "sesiones", sesionDeLaPruebaDeRed))))
	anteponerALaRespuesta(t, deLaPruebaDeRed, loDichoEnOtraConversacion+"\n\n")

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

	// Las dos sesiones son del modo orden, de evals que no son sin binario ni
	// servidor, y sus invocaciones son órdenes de la traza: cada sesión escribe su
	// modo, linea_sin_consulta falso y citas_sin_consulta como una lista vacía, no
	// null, y cada invocación, llamada falso (contracts/evals-en-dos-modos.md §4 de
	// H21).
	var invocaciones int

	for _, resultado := range leido.crudo.Evals {
		assert.JSONEq(t, `"orden"`, string(resultado.Modo), "modo de %s", resultado.Sesion)
		assert.Equal(t, "false", string(resultado.LineaSinConsulta), "linea_sin_consulta de %s", resultado.Sesion)
		assert.Equal(t, "[]", string(resultado.CitasSinConsulta),
			"citas_sin_consulta de %s es una lista vacía, no null", resultado.Sesion)

		for _, invocacion := range resultado.Invocaciones {
			invocaciones++

			assert.Equal(t, "false", string(invocacion.Llamada), "llamada de %s en %s", invocacion.Orden, resultado.Sesion)
		}
	}

	assert.Positive(t, invocaciones, "alguna sesión del caso tiene invocaciones que mirar")

	// Sin hallazgos esperados, ninguna serie exige forma: formas es una lista
	// vacía, no null, y su celda de la tabla de las series dice «ninguna».
	require.Len(t, leido.crudo.Tasas, 2, "informe.json tiene las dos series del caso")

	for _, tasa := range leido.crudo.Tasas {
		assert.Equal(t, "[]", string(tasa.Formas), "formas de %s con %s es una lista vacía, no null",
			tasa.Eval, tasa.Modelo)
		assert.Equal(t, sinFormasExigidas, tasaDeLaSerie(t, leido.informe, tasa.Eval, tasa.Modelo).Formas)

		// Las dos series son del modo orden, el único de un plan sin modos, y su
		// modelo es el id a secas (contracts/evals-en-dos-modos.md §5 de H21).
		assert.JSONEq(t, `"orden"`, string(tasa.Modo), "modo de %s con %s", tasa.Eval, tasa.Modelo)
		assert.Equal(t, modeloQueDecide, tasa.Modelo)
	}

	exigirLineas(t, seccionDelInforme(t, leido.md, "Tasas por eval"),
		filaDeTabla(encabezadosDeLaTablaDeTasas...),
		filaDeTabla(slices.Repeat([]string{"---"}, len(encabezadosDeLaTablaDeTasas))...),
		filaDeTabla(ficheroDeLaEval01, modeloQueDecide, "orden", "sí", "sí", "ninguna", "1 de 1", "llega al umbral"),
		filaDeTabla(ficheroDeNoActivacion, modeloQueDecide, "orden", "sí", "sí", "ninguna", "1 de 1", "llega al umbral"))

	// Los comandos prohibidos ejecutados van en la tabla de las sesiones detrás
	// de los comandos ausentes, los hallazgos encontrados y los ausentes, detrás de
	// los avisos, las redacciones modificadas encontradas y las ausentes, detrás de
	// los hallazgos, el territorio encontrado y el ausente, detrás de las
	// redacciones, y las expresiones prohibidas, detrás del territorio, vacíos en
	// las dos.
	exigirLineas(t, seccionDelInforme(t, leido.md, "Sesiones"),
		filaDeTabla(encabezadosDeLaTablaDeSesiones...),
		filaDeTabla(slices.Repeat([]string{"---"}, len(encabezadosDeLaTablaDeSesiones))...),
		filaDeTabla(sesionDelArticulo21, ficheroDeLaEval01, modeloQueDecide, "orden", "sí", "sí", "sí (código 0)",
			"ninguno", "ninguno", "ninguna", "ninguno", "ninguno", "ninguno", "ninguno", "ninguna", "ninguna",
			"ninguno", "ninguno", "0", "no", "pasa"),
		filaDeTabla(sesionDeNoActivacion, ficheroDeNoActivacion, modeloQueDecide, "orden", "no", "no", "sí (código 0)",
			"ninguno", "ninguno", "ninguna", "ninguno", "ninguno", "ninguno", "ninguno", "ninguna", "ninguna",
			"ninguno", "ninguno", "0", "no", "pasa"))

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
		Eval: ficheroDeLaEval01, Modelo: modeloQueDecide, Modo: ModoOrden, Planificada: true, Decide: true,
		Formas: sinFormasExigidas, Sesiones: 3, Pasan: 2, Pasa: true,
	}, tasa)

	assert.False(t, resultadoDeLaSesion(t, leido.informe, sesionDelArticulo21+"-claude-haiku-03").Pasa,
		"la tercera sesión no pasa y aun así la serie llega al umbral")
	exigirMotivosDeLaRaiz(t, leido)
	assert.Equal(t, VeredictoAprobado, leido.informe.Veredicto)

	exigirLineas(t, seccionDelInforme(t, leido.md, "Tasas por eval"),
		filaDeTabla(ficheroDeLaEval01, modeloQueDecide, "orden", "sí", "sí", "ninguna", "2 de 3", "llega al umbral"))
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
		Eval: ficheroDeLaEvalInformativa, Modelo: modeloQueDecide, Modo: ModoOrden, Planificada: true,
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
		Eval: ficheroDeLaEval01, Modelo: modeloInformativoDelCaso, Modo: ModoOrden, Planificada: true,
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
// informe.json; los umbrales detrás de las tasas, seguidos del juez, de la
// duración de las sesiones, los reintentos por límite de
// ritmo y las sesiones sin medir, en informe.json, con los umbrales siempre como
// lista y cumpliendo los invariantes del ADR 0029, y, en informe.md, la sección
// de los umbrales detrás de la de las tasas, la del juez detrás de ella, la de
// las sesiones sin medir detrás de la del juez y la duración y los reintentos en
// la cabecera (contrato informe-del-job §1, §3 y §4 de H7.3;
// contracts/informe-del-job.md §3 y §7 de H24); y, desde H24, nada de la lista
// de expresiones prohibidas (exigirElInformeSinLaLista) y el juez con su forma
// (exigirLaFormaDelJuez).
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
		"tasas":                          "umbrales",
		"umbrales":                       "juez",
		"juez":                           "duracion_de_las_sesiones",
		"duracion_de_las_sesiones":       "reintentos_por_limite_de_ritmo",
		"reintentos_por_limite_de_ritmo": "sesiones_sin_medir",
	} {
		assert.Equal(t, siguiente, claveDetras(t, escrito, anterior), "en informe.json, %s va detrás de %s",
			siguiente, anterior)
	}

	for anterior, siguiente := range map[string]string{
		"Tasas por eval": "Umbrales",
		"Umbrales":       "Juez",
		"Juez":           "Sesiones sin medir",
	} {
		assert.Equal(t, siguiente, seccionDetras(t, leido.md, anterior), "en informe.md, la sección %s va detrás de %s",
			siguiente, anterior)
	}

	exigirElInformeSinLaLista(t, escrito, leido.md)
	exigirLaFormaDelJuez(t, leido)

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

// exigirElInformeSinLaLista exige lo que todo informe cumple desde H24
// (contracts/informe-del-job.md §6 y §7 de H24; FR-062), tenga o no lista la
// carpeta de sus evals: informe.json no lleva en su raíz el recuento de las
// expresiones prohibidas por modelo ni, en ninguna sesión, sus expresiones; e
// informe.md no tiene la sección de ese recuento ni, en la tabla de las
// sesiones, la columna de las expresiones.
func exigirElInformeSinLaLista(t *testing.T, escrito, md string) {
	t.Helper()

	var raiz map[string]jsontext.Value
	require.NoError(t, json.Unmarshal([]byte(escrito), &raiz))
	assert.NotContains(t, raiz, claveDelRecuentoDeExpresiones, "informe.json no lleva el recuento de las expresiones")

	var sesiones []map[string]jsontext.Value
	require.NoError(t, json.Unmarshal(raiz["evals"], &sesiones))

	for _, sesion := range sesiones {
		assert.NotContains(t, sesion, claveDeLasExpresiones, "la sesión %s de informe.json no lleva expresiones",
			sesion["sesion"])
	}

	assert.NotContains(t, md, "\n## "+seccionDelRecuentoDeExpresiones+"\n",
		"informe.md no tiene la sección del recuento de las expresiones")

	encabezados, _, _ := strings.Cut(seccionDelInforme(t, md, "Sesiones"), "\n")
	assert.NotContains(t, encabezados, columnaDeExpresiones, "la tabla de las sesiones no tiene la columna de las expresiones")
}

// Lo que dice la sección «Juez» de informe.md, escrito a mano
// (contracts/informe-del-job.md §7 de H24): la de una skill sin juez; y la de
// un juez sin ningún voto afirmativo ni ninguna respuesta sin juzgar, con el
// modelo y la versión con los que votan los tests del informe.
const (
	seccionSinJuez = "La skill no tiene juez."

	seccionDelJuezSinVotos = "Modelo: claude-opus-5-5\n\nVersión de Claude Code: 2.1.289\n\n" +
		"Votos: ninguna\n\nRespuestas sin juzgar: ninguna"
)

// exigirLaFormaDelJuez exige lo que todo informe cumple de su juez
// (contracts/informe-del-job.md §3 y §7 de H24; FR-060, FR-061): sin él, juez
// es null en informe.json y la sección «Juez» de informe.md dice que la skill
// no lo tiene; con él, respuestas y sin_juzgar son listas, nunca null, y la
// sección lleva su modelo y la versión de Claude Code de sus votos.
func exigirLaFormaDelJuez(t *testing.T, leido informeLeido) {
	t.Helper()

	seccion := seccionDelInforme(t, leido.md, "Juez")

	if leido.informe.Juez == nil {
		assert.Equal(t, "null", compacto(t, leido.crudo.Juez), "sin juez, juez es null")
		assert.Equal(t, seccionSinJuez, seccion)

		return
	}

	var listas struct {
		Respuestas jsontext.Value `json:"respuestas"`
		SinJuzgar  jsontext.Value `json:"sin_juzgar"`
	}

	require.NoError(t, json.Unmarshal(leido.crudo.Juez, &listas))
	assert.True(t, strings.HasPrefix(compacto(t, listas.Respuestas), "["), "juez.respuestas es una lista, nunca null")
	assert.True(t, strings.HasPrefix(compacto(t, listas.SinJuzgar), "["), "juez.sin_juzgar es una lista, nunca null")

	exigirLineas(t, seccion, "Modelo: "+leido.informe.Juez.Modelo,
		"Versión de Claude Code: "+leido.informe.Juez.VersionDeClaudeCode)
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

// tablaEscrita es una tabla de informe.md con esos encabezados y esas filas,
// sin el salto de línea final.
func tablaEscrita(encabezados []string, filas ...[]string) string {
	lineas := []string{
		filaDeTabla(encabezados...),
		filaDeTabla(slices.Repeat([]string{"---"}, len(encabezados))...),
	}
	for _, fila := range filas {
		lineas = append(lineas, filaDeTabla(fila...))
	}

	return strings.Join(lineas, "\n")
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

// Lo que los tests del informe ponen en la carpeta juez de su skill sintética
// en lugar de lo que deja escribirLaCarpetaDelJuez (contracts/informe-del-job.md
// §2 de H24): las dos clases de boe-legislacion, con su umbral, y una medida
// que se puede leer. El esquema de la respuesta es esquemaDeLasDosClases, con
// el que se valida cada voto.
const (
	clasesDelJuezDelInforme = "clases:\n" +
		"  - nombre: afirma_lo_no_leido\n    decide: true\n    umbral: 0\n" +
		"  - nombre: cuenta_su_proceso\n    decide: false\n    umbral: 0\n"

	// medidaDelJuezDelInforme lleva los recuentos de la medida versionada de
	// boe-legislacion: 0 de 212 defectos sin marcar y 0 de 47 correctos marcados.
	medidaDelJuezDelInforme = `{"clase":"afirma_lo_no_leido","fecha":"2026-10-05",` +
		`"modelo_del_juez":"claude-opus-5-5","version_de_claude_code":"2.1.289",` +
		`"rubrica":{"sha256":"huella-de-la-rubrica"},"casos":{"sha256":"huella-de-los-casos"},` +
		`"defectos":{"casos":212,"sin_marcar":0},"correctos":{"casos":47,"marcados":0}}` + "\n"

	// medidaDelJuezSinCumplirse es una medida con otros totales y sin cumplirse:
	// 1 de 20 defectos sin marcar y 2 de 5 correctos marcados.
	medidaDelJuezSinCumplirse = `{"clase":"afirma_lo_no_leido","fecha":"2026-10-05",` +
		`"modelo_del_juez":"claude-opus-5-5","version_de_claude_code":"2.1.289",` +
		`"rubrica":{"sha256":"huella-de-la-rubrica"},"casos":{"sha256":"huella-de-los-casos"},` +
		`"defectos":{"casos":20,"sin_marcar":1},"correctos":{"casos":5,"marcados":2}}` + "\n"
)

// Con lo que votan los tests del informe: el modelo del juez y la versión de
// Claude Code de sus votos, los fijados para el job de boe-legislacion, y las
// respuestas que se votan a la vez, las de su concurrencia (research D6 de H24).
const (
	modeloDelJuezDelInforme       = "claude-opus-5-5"
	versionDelJuezDelInforme      = "2.1.289"
	concurrenciaDelJuezDelInforme = 4
)

// escribirElJuezDelInforme deja en la carpeta de evals dada, que es de un
// directorio temporal del test, la carpeta juez de la skill sintética de los
// tests del informe: la de escribirLaCarpetaDelJuez con las dos clases de
// boe-legislacion, el esquema de su respuesta y una medida que se puede leer,
// la dada o, sin ella, medidaDelJuezDelInforme.
func escribirElJuezDelInforme(t *testing.T, evals, medida string) {
	t.Helper()

	escribirLaCarpetaDelJuez(t, evals)
	crearEntradas(t, evals, []entradaDeConjunto{
		{nombre: clasesEnElJuez, contenido: clasesDelJuezDelInforme},
		{nombre: esquemaEnElJuez, contenido: esquemaDeLasDosClases},
		{nombre: medidaEnElJuez, contenido: cmp.Or(medida, medidaDelJuezDelInforme)},
	})

	conjunto, err := LeerConjunto(evals)
	require.NoError(t, err)
	require.NotNil(t, conjunto.Juez, "%s tiene la carpeta del juez, bien formada", evals)

	_, err = leerMedidaDelJuez([]byte(conjunto.Juez.Medida))
	require.NoError(t, err, "la medida del juez de %s se puede leer", evals)
}

// clasesDelJuezConSentencia son las clases que los tests del informe ponen en
// la carpeta juez de su skill sintética cuando su juez es el del esquema con
// sentencia (ponerElJuezConSentencia): las dos de jurisprudencia, con su umbral,
// como las declara contracts/juez-de-jurisprudencia.md §1 de H25.
const clasesDelJuezConSentencia = "clases:\n" +
	"  - nombre: afirma_lo_no_leido\n    decide: true\n    umbral: 0\n" +
	"  - nombre: afirma_que_existe\n    decide: false\n    umbral: 0\n"

// ponerElJuezConSentencia cambia, en la carpeta juez que
// escribirElJuezDelInforme dejó en la carpeta de evals dada, las clases y el
// esquema de la respuesta por los del juez cuyo esquema lleva sentencia: las
// dos clases de jurisprudencia y esquemaConSentencia. La rúbrica, los casos y la
// medida siguen siendo los de la skill sintética.
func ponerElJuezConSentencia(t *testing.T, evals string) {
	t.Helper()

	crearEntradas(t, evals, []entradaDeConjunto{
		{nombre: clasesEnElJuez, contenido: clasesDelJuezConSentencia},
		{nombre: esquemaEnElJuez, contenido: esquemaConSentencia},
	})

	conjunto, err := LeerConjunto(evals)
	require.NoError(t, err)
	require.NotNil(t, conjunto.Juez, "%s tiene la carpeta del juez, bien formada", evals)
}

// votosDeUnaRespuesta son los votos grabados de la respuesta de una sesión de
// una ejecución sintética.
type votosDeUnaRespuesta struct {
	// sesion es la sesión cuya respuesta se juzga con ellos.
	sesion string

	// parrafo es lo que el test antepone a su respuesta: lleva las frases que
	// citan sus votos y no está en ninguna otra respuesta, y por él reconoce el
	// votante el mensaje de sus votos.
	parrafo string

	// votos son las grabaciones de sus votos, en su orden.
	votos []grabacion

	// tarda es lo que avanza el reloj del votante con cada uno de sus votos.
	tarda time.Duration
}

// votanteDelInforme es el votante de salidas grabadas de los tests del informe,
// que cuenta sus llamadas: de la respuesta que lleva el párrafo de unos votos
// grabados devuelve esos votos, por orden, y de cualquier otra, el voto que
// dice no en las dos clases. EscribirInforme lo llama desde varias gorrutinas,
// y con cada respuesta en la suya los votos de una misma respuesta le llegan
// uno detrás de otro. Lleva además el reloj que da como Ahora, que solo avanza
// con los votos grabados que tardan: así los segundos de los votos de un modo
// son la suma de lo que tardan los suyos, se voten cuantos se voten a la vez.
type votanteDelInforme struct {
	t *testing.T

	// queNo es la salida del voto que dice no en las dos clases.
	queNo string

	grabados []votosDeUnaRespuesta

	// mutex protege lo que sigue.
	mutex sync.Mutex

	// dados son, de cada respuesta con votos grabados, los que ya se han pedido.
	dados []int

	// mensajes son los de los votos pedidos, en el orden en que llegaron.
	mensajes []string

	reloj time.Time
}

// nuevoVotanteDelInforme da el votante de los tests del informe con esos votos
// grabados; sin ninguno, dice no a todo y su reloj no se mueve.
func nuevoVotanteDelInforme(t *testing.T, grabados ...votosDeUnaRespuesta) *votanteDelInforme {
	t.Helper()

	return &votanteDelInforme{
		t:        t,
		queNo:    votoDeLasDosClases(t, afirmaQueNo(), cuentaQueNo()).salida,
		grabados: grabados,
		dados:    make([]int, len(grabados)),
		reloj:    time.Date(2026, time.October, 5, 12, 0, 0, 0, time.UTC),
	}
}

func (v *votanteDelInforme) votar(mensaje string) ([]byte, error) {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	v.mensajes = append(v.mensajes, mensaje)

	posicion := slices.IndexFunc(v.grabados, func(grabado votosDeUnaRespuesta) bool {
		return strings.Contains(mensaje, grabado.parrafo)
	})
	if posicion < 0 {
		return []byte(v.queNo), nil
	}

	grabado := v.grabados[posicion]
	v.reloj = v.reloj.Add(grabado.tarda)

	numero := v.dados[posicion]
	v.dados[posicion]++

	if numero >= len(grabado.votos) {
		v.t.Errorf("de %s se pide el voto %d y solo hay %d grabados", grabado.sesion, numero+1, len(grabado.votos))

		return nil, errors.New("voto sin grabar")
	}

	return []byte(grabado.votos[numero].salida), grabado.votos[numero].err
}

// ahora es el reloj del votante, el Ahora de sus ejecuciones.
func (v *votanteDelInforme) ahora() time.Time {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	return v.reloj
}

// pedidos son los mensajes de los votos que se le han pedido, ordenados: tantos
// como llamadas ha recibido, y una lista vacía si no ha recibido ninguna.
func (v *votanteDelInforme) pedidos() []string {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	pedidos := make([]string, len(v.mensajes))
	copy(pedidos, v.mensajes)
	slices.Sort(pedidos)

	return pedidos
}

// darA pone en las entradas de EscribirInforme con qué votar: el votante, su
// reloj, el modelo y la versión del juez y cuántas respuestas se votan a la vez.
func (v *votanteDelInforme) darA(entradas *InformeAEscribir) {
	entradas.Votar = v.votar
	entradas.Ahora = v.ahora
	entradas.ModeloDelJuez = modeloDelJuezDelInforme
	entradas.VersionDelJuez = versionDelJuezDelInforme
	entradas.ConcurrenciaDelJuez = concurrenciaDelJuezDelInforme
}

// Lo que TestInformeConElJuez sabe de las sesiones de sus ejecuciones
// sintéticas, escrito a mano: la pregunta de sus evals, que es la de la eval
// del art. 21 del caso aprobado, y la salida de la llamada con la que leen el
// bloque las del modo herramienta (llamadaSinteticaAlArticulo21).
const (
	preguntaDeLasSinteticas   = "¿qué dice el art. 21 de la Ley 39/2015?"
	sobreDeLaLlamadaSintetica = `{"ok":true,"fuente":"boe.legislacion-consolidada","data":{"bloque":"a21"}}`
)

// Los párrafos que TestInformeConElJuez antepone a las respuestas con votos
// grabados, cada uno distinto de los demás, y las frases que esos votos citan
// de ellos. El primero es de respuestaJuzgada, con sus tres frases.
const (
	parrafoDeLosArticulos22Y24 = "El artículo 22 regula la suspensión del plazo, y el artículo 24, el silencio administrativo."

	parrafoDeLosArticulos23Y25 = "El artículo 23 permite ampliar el plazo, y el artículo 25, declarar la caducidad."
	fraseDelArticulo23         = "El artículo 23 permite ampliar el plazo"
	fraseDeLosArticulos23Y25   = "permite ampliar el plazo, y el artículo 25"
	fraseDelArticulo25         = "el artículo 25, declarar la caducidad"

	parrafoDelProcesoUno  = "He comprobado la redacción del artículo 21 y no ha cambiado."
	parrafoDelProcesoDos  = "He comprobado la redacción dos veces y no ha cambiado."
	parrafoDelProcesoTres = "He comprobado la redacción con el índice y no ha cambiado."

	parrafoLentoDeOrden       = "Esta respuesta del modo orden tarda en juzgarse."
	parrafoLentoDeHerramienta = "Esta respuesta del modo herramienta tarda en juzgarse."
	parrafoLentoSinBinario    = "Esta respuesta sin binario ni servidor tarda en juzgarse."
)

// Las otras dos frases que los votos del juez cuyo esquema lleva sentencia
// citan de parrafoDeLaSentencia, que es el párrafo que los tests del informe
// anteponen a la respuesta con esos votos: con fraseDeLaNulidad, las tres de la
// respuesta que marcan.
const (
	fraseDeLasClausulas    = "declaró la nulidad de las cláusulas suelo"
	fraseDeLaTransparencia = "las cláusulas suelo por falta de transparencia"
)

// Las tres frases de cada uno de los dos párrafos con preceptos, como las
// escribe el motivo de la respuesta que marcan (contracts/informe-del-job.md §4
// de H24).
const (
	frasesDel22Y24 = "«" + fraseDelArticulo22 + "» · «" + fraseDeLosDosArticulos + "» · «" + fraseDelArticulo24 + "»"
	frasesDel23Y25 = "«" + fraseDelArticulo23 + "» · «" + fraseDeLosArticulos23Y25 + "» · «" + fraseDelArticulo25 + "»"
)

// casoDelInformeConElJuez es un caso de TestInformeConElJuez: una ejecución
// sintética, los votos grabados de algunas de sus respuestas y lo que el
// informe tiene que decir.
type casoDelInformeConElJuez struct {
	nombre    string
	ejecucion ejecucionConUmbrales
	grabados  []votosDeUnaRespuesta
	umbrales  []Umbral
	motivos   []string
	veredicto Veredicto

	// comoSinVotos dice que el informe tiene que llevar los mismos umbrales y
	// los mismos motivos que el de las mismas sesiones con un juez que dice no a
	// todo.
	comoSinVotos bool

	// respuestas son las entradas de juez.respuestas del informe, en su orden,
	// y sinJuzgar, las de juez.sin_juzgar; sin ninguna, su lista va vacía. De
	// una ejecución sin juez no se miran: su juez es null.
	respuestas []RespuestaConVotos
	sinJuzgar  []RespuestaSinJuzgar

	// escrita es, si no está vacía, la primera entrada de juez.respuestas tal
	// como tiene que estar escrita en informe.json.
	escrita string
}

// entradaDeSiSiYNo es la entrada de juez.respuestas de la respuesta con los
// votos sí, sí y no del caso ninguna-marcada de TestInformeConElJuez, escrita a
// mano con la forma del ejemplo de contracts/informe-del-job.md §3 de H24: sus
// tres votos en cada clase, las dos frases de los dos primeros en la que
// decide, sin precepto en la que no lo tiene y sin marcar en ninguna.
const entradaDeSiSiYNo = `{"sesion":"01-sintetica-claude-sonnet-5-5-01","clases":[` +
	`{"clase":"afirma_lo_no_leido","marcada":false,"votos":[` +
	`{"voto":1,"nulo":false,"motivo":"Dice de qué trata un artículo que ninguna herramienta devolvió.",` +
	`"respuesta":"si","frase":"El artículo 22 regula la suspensión del plazo",` +
	`"precepto":"Art. 22 de la Ley 39/2015","frase_en_la_respuesta":true},` +
	`{"voto":2,"nulo":false,"motivo":"Dice de qué trata un artículo que ninguna herramienta devolvió.",` +
	`"respuesta":"si","frase":"el artículo 24, el silencio administrativo",` +
	`"precepto":"Art. 22 de la Ley 39/2015","frase_en_la_respuesta":true},` +
	`{"voto":3,"nulo":false,"motivo":"Todo lo que expone está en el texto devuelto.",` +
	`"respuesta":"no","frase":"","precepto":"","frase_en_la_respuesta":false}]},` +
	`{"clase":"cuenta_su_proceso","marcada":false,"votos":[` +
	`{"voto":1,"nulo":false,"motivo":"No cuenta comprobaciones ni nombra herramientas.",` +
	`"respuesta":"no","frase":"","frase_en_la_respuesta":false},` +
	`{"voto":2,"nulo":false,"motivo":"No cuenta comprobaciones ni nombra herramientas.",` +
	`"respuesta":"no","frase":"","frase_en_la_respuesta":false},` +
	`{"voto":3,"nulo":false,"motivo":"No cuenta comprobaciones ni nombra herramientas.",` +
	`"respuesta":"no","frase":"","frase_en_la_respuesta":false}]}]}`

// entradaConSentencia es la entrada de juez.respuestas de la respuesta marcada
// del caso con-sentencia de TestInformeConElJuez, escrita a mano con la forma
// de contracts/juez-de-jurisprudencia.md §4 de H25, cuyo voto de ejemplo es el
// primero: sus tres votos en cada clase, con sentencia y sin precepto en la que
// decide, donde queda marcada, y sin ninguno de los dos en la que solo se
// publica.
const entradaConSentencia = `{"sesion":"01-sintetica-claude-sonnet-5-5-01","clases":[` +
	`{"clase":"afirma_lo_no_leido","marcada":true,"votos":[` +
	`{"voto":1,"nulo":false,"motivo":"La respuesta dice qué declaró la sentencia, y su texto no está ni en la ` +
	`pregunta ni en lo que devolvió la herramienta.","respuesta":"si",` +
	`"frase":"declaró la nulidad de las cláusulas suelo por falta de transparencia",` +
	`"sentencia":"STS 241/2013, de 9 de mayo","frase_en_la_respuesta":true},` +
	`{"voto":2,"nulo":false,"motivo":"La respuesta dice qué declaró la sentencia, y su texto no está ni en la ` +
	`pregunta ni en lo que devolvió la herramienta.","respuesta":"si",` +
	`"frase":"declaró la nulidad de las cláusulas suelo",` +
	`"sentencia":"STS 241/2013, de 9 de mayo","frase_en_la_respuesta":true},` +
	`{"voto":3,"nulo":false,"motivo":"La respuesta dice qué declaró la sentencia, y su texto no está ni en la ` +
	`pregunta ni en lo que devolvió la herramienta.","respuesta":"si",` +
	`"frase":"las cláusulas suelo por falta de transparencia",` +
	`"sentencia":"STS 241/2013, de 9 de mayo","frase_en_la_respuesta":true}]},` +
	`{"clase":"afirma_que_existe","marcada":false,"votos":[` +
	`{"voto":1,"nulo":false,"motivo":"La respuesta lleva la línea de la sentencia no comprobada y no dice que exista.",` +
	`"respuesta":"no","frase":"","frase_en_la_respuesta":false},` +
	`{"voto":2,"nulo":false,"motivo":"La respuesta lleva la línea de la sentencia no comprobada y no dice que exista.",` +
	`"respuesta":"no","frase":"","frase_en_la_respuesta":false},` +
	`{"voto":3,"nulo":false,"motivo":"La respuesta lleva la línea de la sentencia no comprobada y no dice que exista.",` +
	`"respuesta":"no","frase":"","frase_en_la_respuesta":false}]}]}`

// votoPublicado es un voto de una respuesta como lo publica el informe del juez
// de las dos clases: su número, si es nulo y lo que dice de cada clase.
type votoPublicado struct {
	numero         int
	nulo           bool
	afirma, cuenta dicho
}

// votoQueMarca es el voto con ese número que dice sí en afirma_lo_no_leido con
// esa frase, y no en cuenta_su_proceso.
func votoQueMarca(numero int, frase string) votoPublicado {
	return votoPublicado{numero: numero, afirma: afirmaQueSi(frase), cuenta: cuentaQueNo()}
}

// votoQueCuenta es el voto con ese número que dice sí en cuenta_su_proceso con
// esa frase, y no en afirma_lo_no_leido.
func votoQueCuenta(numero int, frase string) votoPublicado {
	return votoPublicado{numero: numero, afirma: afirmaQueNo(), cuenta: cuentaQueSi(frase)}
}

// votoQueDiceNo es el voto con ese número que dice no en las dos clases.
func votoQueDiceNo(numero int) votoPublicado {
	return votoPublicado{numero: numero, afirma: afirmaQueNo(), cuenta: cuentaQueNo()}
}

// anulado es el mismo voto, nulo.
func (v votoPublicado) anulado() votoPublicado {
	v.nulo = true

	return v
}

// respuestaPublicada es la entrada de juez.respuestas de la sesión con esos
// votos, que son todos los suyos, y con si quedó marcada en cada una de las dos
// clases. De lo que el test sabe de sus frases: la de un voto que dice sí está
// en la respuesta, salvo fraseQueNoEsta, y la de uno que dice no, vacía, no
// está.
func respuestaPublicada(sesion string, enAfirma, enCuenta bool, votos ...votoPublicado) RespuestaConVotos {
	afirma := JuicioDeClase{Clase: claseAfirmaLoNoLeido, Marcada: enAfirma}
	cuenta := JuicioDeClase{Clase: claseCuentaSuProceso, Marcada: enCuenta}

	esta := func(dicho dicho) bool { return dicho.respuesta == "si" && dicho.frase != fraseQueNoEsta }

	for _, voto := range votos {
		afirma.Votos = append(afirma.Votos, voto.afirma.voto(voto.numero, voto.nulo, esta(voto.afirma)))
		cuenta.Votos = append(cuenta.Votos, voto.cuenta.voto(voto.numero, voto.nulo, esta(voto.cuenta)))
	}

	return RespuestaConVotos{Sesion: sesion, Clases: []JuicioDeClase{afirma, cuenta}}
}

// TestInformeConElJuez fija lo que el juez con modelo hace en el informe
// (contracts/informe-del-job.md §1 a §4 y §9 de H24; data-model §3 y §5;
// research D6, D7, D9, D10 y D13; FR-007, FR-012 a FR-014, FR-030, FR-031,
// FR-033, FR-035, FR-037, FR-060, FR-061, FR-104; SC-004), con ejecuciones
// sintéticas de
// ejecucionConUmbrales armadas en t.TempDir() —los dos modelos, una eval que
// decide y una informativa en los dos modos, con sus textos en el transcript,
// la eval sin binario ni servidor y la sesión de la prueba de red— y un
// votante de salidas grabadas que cuenta sus llamadas:
//
//   - se juzgan las respuestas del modelo que decide de las evals que activan
//     la skill, cada una con el mensaje del voto de la pregunta de su eval, su
//     respuesta y sus textos —los de su orden, los de su llamada o ninguno—, y
//     ninguna más: ni las de un modelo de los informativos ni la de la prueba
//     de red;
//   - una respuesta marcada en un modo pone el veredicto en fallo, con el
//     umbral de ese modo sin cumplir, el del otro cumplido y un motivo que la
//     nombra por su sesión con sus tres frases; dos, las dos, en orden de
//     sesión; y de un voto nulo cuenta la frase de su repetición;
//   - sin ninguna marcada —tampoco la de «sí, sí y no»—, los umbrales se
//     cumplen y el veredicto es aprobado;
//   - tres con sí en cuenta_su_proceso dan su umbral sin cumplir, que no
//     decide, y el mismo veredicto;
//   - un voto que no llega deja la respuesta sin juzgar: en el total y no en la
//     medida, tampoco en la de la clase que solo se publica, con el veredicto
//     fallo por el motivo de la ejecución que la nombra;
//   - la respuesta de la eval sin binario ni servidor marcada con tres síes
//     deja los umbrales, los motivos y su sesión como sin esa marca, y lo que
//     tardan sus votos no entra en ninguna duración;
//   - 901 s de votos en un modo, con el reloj del votante, dan fallo con el
//     motivo de la ejecución, y 900, no;
//   - los motivos van en su orden: el del umbral con los de los umbrales, el de
//     las respuestas sin juzgar detrás del de las sesiones sin medir, y el de la
//     duración del juez con los de la duración;
//   - y una skill sin juez no tiene umbrales ni pide ningún voto.
//
// En todos, el juicio sin modelo no cambia con los votos (FR-014): cada sesión
// y cada tasa son las del informe de las mismas sesiones con un juez que dice
// no a todo.
//
// Y fija lo que el informe publica del juez en su clave juez
// (contracts/informe-del-job.md §3 de H24; FR-013, FR-060, FR-061), con el
// modelo y la versión recibidos:
//
//   - en respuestas, una entrada por respuesta juzgada con algún voto
//     afirmativo, en orden de sesión, con todos sus votos en cada clase y con si
//     quedó marcada: la de «sí, sí y no», sin marcar, con sus tres votos y sus
//     dos frases, escrita con la forma del contrato; las tres con sí en
//     cuenta_su_proceso, cada una con su frase y marcada en esa clase; la que
//     queda marcada tras un voto nulo, con sus cuatro votos y el primero nulo;
//     la que solo tiene un sí nulo, con él y su repetición, y la de un nulo
//     cuya repetición también lo es, con los dos nulos; y la de la eval sin
//     binario ni servidor marcada, con sus tres votos y sus tres frases, como
//     las demás;
//   - de una respuesta cuyos votos dicen todos no, nada;
//   - en sin_juzgar, una entrada por respuesta sin juzgar, con su sesión y el
//     motivo de su voto, en orden de sesión como en el motivo de la raíz, y esa
//     respuesta no está en respuestas aunque un voto anterior dijera sí;
//   - y con una skill sin juez, juez es null.
//
// Desde H25 (contracts/juez-de-jurisprudencia.md §4 y §9 de H25; data-model §2
// de H25; FR-005, FR-013 y FR-108 de H25; SC-008 de H25), fija además el campo
// propio de la clase en cada voto publicado, con sus claves en el orden del
// contrato: precepto con el juez de siempre, cuya entrada no cambia en un byte,
// y, con un juez cuyo esquema lleva sentencia y una respuesta marcada,
// sentencia y no precepto en cada voto de afirma_lo_no_leido, y ninguno de los
// dos en los de afirma_que_existe.
func TestInformeConElJuez(t *testing.T) {
	t.Parallel()

	for _, caso := range casosDelInformeConElJuez(t) {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()
			exigirElInformeConElJuez(t, caso)
		})
	}

	t.Run("con-sentencia", func(t *testing.T) {
		t.Parallel()
		exigirElInformeConSentencia(t)
	})

	t.Run("sin-con-que-votar", func(t *testing.T) {
		t.Parallel()
		exigirQueSinVotanteNoHayInforme(t)
	})
}

// casosDelInformeConElJuez son los casos de TestInformeConElJuez.
func casosDelInformeConElJuez(t *testing.T) []casoDelInformeConElJuez {
	t.Helper()

	conJuez := ejecucionConUmbrales{
		queDeciden: 1, informativas: 1, conHaiku: true, dosModos: true, conLaSinBinarioNiServidor: true,
		conTextos: true, conLaPruebaDeRed: true,
		duracion: 500, duracionEnHerramienta: 600, duracionSinModo: 40, objetivo: 900,
	}
	enUnModo := ejecucionConUmbrales{
		queDeciden: 1, informativas: 1, conHaiku: true, conTextos: true, conLaPruebaDeRed: true,
	}

	deLasSesiones := []Umbral{umbralDeDuracion(ModoOrden, 500, 900, true), umbralDeDuracion(ModoHerramienta, 600, 900, true)}

	deOrden := sesionSinteticaEn(ModoOrden, 1, modeloSonnet55, 1)
	otraDeOrden := sesionSinteticaEn(ModoOrden, 2, modeloSonnet55, 2)
	deHerramienta := sesionSinteticaEn(ModoHerramienta, 1, modeloSonnet55, 2)
	otraDeHerramienta := sesionSinteticaEn(ModoHerramienta, 2, modeloSonnet55, 3)
	sinBinario := sesionSintetica(3, modeloSonnet55, 1)

	queNo := votoDeLasDosClases(t, afirmaQueNo(), cuentaQueNo())
	marca := func(frase string) grabacion { return votoDeLasDosClases(t, afirmaQueSi(frase), cuentaQueNo()) }
	cuenta := func(frase string) grabacion { return votoDeLasDosClases(t, afirmaQueNo(), cuentaQueSi(frase)) }

	tresSies := []grabacion{marca(fraseDelArticulo22), marca(fraseDeLosDosArticulos), marca(fraseDelArticulo24)}

	// El primero de estos cuatro es nulo, que su frase no está en la respuesta,
	// y se repite: las tres frases que marcan son las de los otros tres.
	tresSiesTrasUnNulo := []grabacion{
		marca(fraseQueNoEsta), marca(fraseDelArticulo23), marca(fraseDeLosArticulos23Y25), marca(fraseDelArticulo25),
	}

	// Esos mismos votos, como los publica el informe: el nulo y su repetición
	// llevan el mismo número.
	deTresSies := []votoPublicado{
		votoQueMarca(1, fraseDelArticulo22), votoQueMarca(2, fraseDeLosDosArticulos), votoQueMarca(3, fraseDelArticulo24),
	}
	deTresSiesTrasUnNulo := []votoPublicado{
		votoQueMarca(1, fraseQueNoEsta).anulado(), votoQueMarca(1, fraseDelArticulo23),
		votoQueMarca(2, fraseDeLosArticulos23Y25), votoQueMarca(3, fraseDelArticulo25),
	}

	return []casoDelInformeConElJuez{
		{
			nombre:    "una-marcada-en-orden",
			ejecucion: conJuez,
			grabados:  []votosDeUnaRespuesta{{sesion: deOrden, parrafo: parrafoDeLosArticulos22Y24, votos: tresSies}},
			umbrales: umbralesConJuez(modeloSonnet55, deLasSesiones,
				medidasDeUnModo{modo: ModoOrden, respuestas: 6, marcadas: 1}, enHerramienta(6, 0)),
			motivos: []string{
				"umbral afirma_lo_no_leido:claude-sonnet-5-5:orden: 1 de 6 (16,7 %), y tiene que ser ≤ 0,0 %: " +
					"01-sintetica-claude-sonnet-5-5-01: " + frasesDel22Y24,
			},
			veredicto:  VeredictoFallo,
			respuestas: []RespuestaConVotos{respuestaPublicada(deOrden, true, false, deTresSies...)},
		},
		{
			// Los votos van en otro orden que sus sesiones: el motivo las nombra
			// en el de las sesiones, y en él van sus entradas. La segunda lleva
			// sus cuatro votos, con el nulo delante de su repetición.
			nombre:    "dos-marcadas-en-herramienta",
			ejecucion: conJuez,
			grabados: []votosDeUnaRespuesta{
				{sesion: otraDeHerramienta, parrafo: parrafoDeLosArticulos23Y25, votos: tresSiesTrasUnNulo},
				{sesion: deHerramienta, parrafo: parrafoDeLosArticulos22Y24, votos: tresSies},
			},
			umbrales: umbralesConJuez(modeloSonnet55, deLasSesiones,
				enOrden(6, 0), medidasDeUnModo{modo: ModoHerramienta, respuestas: 6, marcadas: 2}),
			motivos: []string{
				"umbral afirma_lo_no_leido:claude-sonnet-5-5:herramienta: 2 de 6 (33,3 %), y tiene que ser ≤ 0,0 %: " +
					"01-sintetica-herramienta-claude-sonnet-5-5-02: " + frasesDel22Y24 +
					"; 02-sintetica-herramienta-claude-sonnet-5-5-03: " + frasesDel23Y25,
			},
			veredicto: VeredictoFallo,
			respuestas: []RespuestaConVotos{
				respuestaPublicada(deHerramienta, true, false, deTresSies...),
				respuestaPublicada(otraDeHerramienta, true, false, deTresSiesTrasUnNulo...),
			},
		},
		{
			// Sí, sí y no: tres votos, dos frases y ninguna marca.
			nombre:    "ninguna-marcada",
			ejecucion: conJuez,
			grabados: []votosDeUnaRespuesta{{
				sesion: deOrden, parrafo: parrafoDeLosArticulos22Y24,
				votos: []grabacion{marca(fraseDelArticulo22), marca(fraseDelArticulo24), queNo},
			}},
			umbrales:     umbralesConJuez(modeloSonnet55, deLasSesiones, enOrden(6, 0), enHerramienta(6, 0)),
			veredicto:    VeredictoAprobado,
			comoSinVotos: true,
			respuestas: []RespuestaConVotos{respuestaPublicada(deOrden, false, false,
				votoQueMarca(1, fraseDelArticulo22), votoQueMarca(2, fraseDelArticulo24), votoQueDiceNo(3))},
			escrita: entradaDeSiSiYNo,
		},
		{
			// Un sí cuya frase no está, nulo, y su repetición, que dice no; y
			// otro cuya repetición es también nula, y entonces su sí no cuenta.
			// Ninguna de las dos respuestas queda marcada ni se vota más, y las
			// dos se publican, con sus dos votos. La del modo herramienta va
			// delante de la del modo orden, que se vota antes: su sesión es
			// anterior.
			nombre:    "un-si-nulo-repetido",
			ejecucion: conJuez,
			grabados: []votosDeUnaRespuesta{
				{sesion: otraDeOrden, parrafo: parrafoDeLosArticulos22Y24, votos: []grabacion{marca(fraseQueNoEsta), queNo}},
				{
					sesion: deHerramienta, parrafo: parrafoDeLosArticulos23Y25,
					votos: []grabacion{marca(fraseQueNoEsta), marca(fraseQueNoEsta)},
				},
			},
			umbrales:     umbralesConJuez(modeloSonnet55, deLasSesiones, enOrden(6, 0), enHerramienta(6, 0)),
			veredicto:    VeredictoAprobado,
			comoSinVotos: true,
			respuestas: []RespuestaConVotos{
				respuestaPublicada(deHerramienta, false, false,
					votoQueMarca(1, fraseQueNoEsta).anulado(), votoQueMarca(1, fraseQueNoEsta).anulado()),
				respuestaPublicada(otraDeOrden, false, false, votoQueMarca(1, fraseQueNoEsta).anulado(), votoQueDiceNo(1)),
			},
		},
		{
			nombre:    "cuenta-su-proceso-con-tres",
			ejecucion: conJuez,
			grabados: []votosDeUnaRespuesta{
				{sesion: deOrden, parrafo: parrafoDelProcesoUno, votos: []grabacion{cuenta(parrafoDelProcesoUno)}},
				{
					sesion: sesionSinteticaEn(ModoOrden, 1, modeloSonnet55, 3), parrafo: parrafoDelProcesoDos,
					votos: []grabacion{cuenta(parrafoDelProcesoDos)},
				},
				{sesion: otraDeOrden, parrafo: parrafoDelProcesoTres, votos: []grabacion{cuenta(parrafoDelProcesoTres)}},
			},
			umbrales: umbralesConJuez(modeloSonnet55, deLasSesiones,
				medidasDeUnModo{modo: ModoOrden, respuestas: 6, conSi: 3}, enHerramienta(6, 0)),
			veredicto: VeredictoAprobado,
			respuestas: []RespuestaConVotos{
				respuestaPublicada(deOrden, false, true, votoQueCuenta(1, parrafoDelProcesoUno)),
				respuestaPublicada(sesionSinteticaEn(ModoOrden, 1, modeloSonnet55, 3), false, true,
					votoQueCuenta(1, parrafoDelProcesoDos)),
				respuestaPublicada(otraDeOrden, false, true, votoQueCuenta(1, parrafoDelProcesoTres)),
			},
		},
		{
			// El primer voto dice sí en las dos clases y el segundo no llega: la
			// respuesta no cuenta en ninguna de las dos medidas, y sí en su total.
			// Va en sin_juzgar, y no en respuestas, aunque su primer voto dijera
			// sí.
			nombre:    "un-voto-que-no-llega",
			ejecucion: conJuez,
			grabados: []votosDeUnaRespuesta{{
				sesion: deHerramienta, parrafo: parrafoDeLosArticulos22Y24 + " " + parrafoDelProcesoUno,
				votos: []grabacion{
					votoDeLasDosClases(t, afirmaQueSi(fraseDelArticulo22), cuentaQueSi(parrafoDelProcesoUno)),
					{err: errTopeDelVoto},
				},
			}},
			umbrales: umbralesConJuez(modeloSonnet55, deLasSesiones, enOrden(6, 0), enHerramienta(6, 0)),
			motivos: []string{
				"de la ejecución, no de la skill: el juez dejó 1 respuestas sin juzgar: " +
					"01-sintetica-herramienta-claude-sonnet-5-5-02 (voto 2: tope de 35 s agotado)",
			},
			veredicto: VeredictoFallo,
			sinJuzgar: []RespuestaSinJuzgar{{Sesion: deHerramienta, Motivo: "voto 2: tope de 35 s agotado"}},
		},
		{
			// Dos sin juzgar: la del modo orden se vota antes y su sesión es
			// posterior, y van en orden de sesión, en el motivo y en sin_juzgar.
			nombre:    "dos-sin-juzgar",
			ejecucion: conJuez,
			grabados: []votosDeUnaRespuesta{
				{sesion: otraDeOrden, parrafo: parrafoLentoDeOrden, votos: []grabacion{{err: errTopeDelVoto}}},
				{sesion: deHerramienta, parrafo: parrafoLentoDeHerramienta, votos: []grabacion{{err: errTopeDelVoto}}},
			},
			umbrales: umbralesConJuez(modeloSonnet55, deLasSesiones, enOrden(6, 0), enHerramienta(6, 0)),
			motivos: []string{
				"de la ejecución, no de la skill: el juez dejó 2 respuestas sin juzgar: " +
					"01-sintetica-herramienta-claude-sonnet-5-5-02 (voto 1: tope de 35 s agotado), " +
					"02-sintetica-claude-sonnet-5-5-02 (voto 1: tope de 35 s agotado)",
			},
			veredicto: VeredictoFallo,
			sinJuzgar: []RespuestaSinJuzgar{
				{Sesion: deHerramienta, Motivo: "voto 1: tope de 35 s agotado"},
				{Sesion: otraDeOrden, Motivo: "voto 1: tope de 35 s agotado"},
			},
		},
		{
			// Sus tres votos tardan 15 000 s, que no son de ningún modo. Se
			// publica como las demás, con sus tres votos y sus tres frases.
			nombre:    "la-sin-binario-ni-servidor-marcada",
			ejecucion: conJuez,
			grabados: []votosDeUnaRespuesta{{
				sesion: sinBinario, parrafo: parrafoDeLosArticulos22Y24, votos: tresSies, tarda: 5000 * time.Second,
			}},
			umbrales:     umbralesConJuez(modeloSonnet55, deLasSesiones, enOrden(6, 0), enHerramienta(6, 0)),
			veredicto:    VeredictoAprobado,
			comoSinVotos: true,
			respuestas:   []RespuestaConVotos{respuestaPublicada(sinBinario, true, false, deTresSies...)},
		},
		{
			// 900 s y un milisegundo son 901 s: se redondea hacia arriba.
			nombre:    "901-s-del-juez-en-orden-y-900-en-herramienta",
			ejecucion: conJuez,
			grabados: []votosDeUnaRespuesta{
				{
					sesion: deOrden, parrafo: parrafoLentoDeOrden, votos: []grabacion{queNo},
					tarda: 900*time.Second + time.Millisecond,
				},
				{sesion: deHerramienta, parrafo: parrafoLentoDeHerramienta, votos: []grabacion{queNo}, tarda: 900 * time.Second},
				{sesion: sinBinario, parrafo: parrafoLentoSinBinario, votos: []grabacion{queNo}, tarda: 5000 * time.Second},
			},
			umbrales: umbralesConJuez(modeloSonnet55, deLasSesiones,
				medidasDeUnModo{modo: ModoOrden, respuestas: 6, segundosDelJuez: 901},
				medidasDeUnModo{modo: ModoHerramienta, respuestas: 6, segundosDelJuez: 900}),
			motivos:   []string{"de la ejecución, no de la skill: duracion_del_juez:orden: 901 s, y tiene que ser ≤ 900 s"},
			veredicto: VeredictoFallo,
		},
		{
			nombre:    "900-s-del-juez-en-los-dos-modos",
			ejecucion: conJuez,
			grabados: []votosDeUnaRespuesta{
				{sesion: deOrden, parrafo: parrafoLentoDeOrden, votos: []grabacion{queNo}, tarda: 900 * time.Second},
				{sesion: deHerramienta, parrafo: parrafoLentoDeHerramienta, votos: []grabacion{queNo}, tarda: 900 * time.Second},
			},
			umbrales: umbralesConJuez(modeloSonnet55, deLasSesiones,
				medidasDeUnModo{modo: ModoOrden, respuestas: 6, segundosDelJuez: 900},
				medidasDeUnModo{modo: ModoHerramienta, respuestas: 6, segundosDelJuez: 900}),
			veredicto: VeredictoAprobado,
		},
		{
			// Una marcada, las sesiones de modeloHaiku45 sin medir, una respuesta
			// sin juzgar cuyo voto agota 901 s, y 901 s de sesiones.
			nombre: "los-motivos-en-su-orden",
			ejecucion: conCambios(enUnModo, func(e *ejecucionConUmbrales) {
				e.sinMedir, e.duracion, e.objetivo = modeloHaiku45, 901, 900
			}),
			grabados: []votosDeUnaRespuesta{
				{sesion: deOrden, parrafo: parrafoDeLosArticulos22Y24, votos: tresSies},
				{
					sesion: otraDeOrden, parrafo: parrafoLentoDeOrden, votos: []grabacion{{err: errTopeDelVoto}},
					tarda: 901 * time.Second,
				},
			},
			umbrales: umbralesConJuez(modeloSonnet55, []Umbral{umbralDeDuracion(ModoOrden, 901, 900, false)},
				medidasDeUnModo{modo: ModoOrden, respuestas: 6, marcadas: 1, segundosDelJuez: 901}),
			motivos: []string{
				"umbral afirma_lo_no_leido:claude-sonnet-5-5:orden: 1 de 6 (16,7 %), y tiene que ser ≤ 0,0 %: " +
					"01-sintetica-claude-sonnet-5-5-01: " + frasesDel22Y24,
				motivoEsperadoDelLimite(sesionesSinMedirDe(modeloHaiku45, 1)...),
				"de la ejecución, no de la skill: el juez dejó 1 respuestas sin juzgar: " +
					"02-sintetica-claude-sonnet-5-5-02 (voto 1: tope de 35 s agotado)",
				"de la ejecución, no de la skill: duracion_de_las_sesiones:orden: 901 s, y tiene que ser ≤ 900 s",
				"de la ejecución, no de la skill: duracion_del_juez:orden: 901 s, y tiene que ser ≤ 900 s",
			},
			veredicto:  VeredictoFallo,
			respuestas: []RespuestaConVotos{respuestaPublicada(deOrden, true, false, deTresSies...)},
			sinJuzgar:  []RespuestaSinJuzgar{{Sesion: otraDeOrden, Motivo: "voto 1: tope de 35 s agotado"}},
		},
		{
			// Sin juez, ni un umbral ni un voto, tenga o no con qué votar
			// (FR-037), y juez es null.
			nombre: "una-skill-sin-juez",
			ejecucion: conCambios(conJuez, func(e *ejecucionConUmbrales) {
				e.sinJuez, e.objetivo = true, 0
			}),
			umbrales:     []Umbral{},
			veredicto:    VeredictoAprobado,
			comoSinVotos: true,
		},
	}
}

// exigirElInformeConElJuez arma la ejecución del caso, antepone a cada
// respuesta con votos grabados su párrafo, escribe su informe con el votante de
// esos votos y exige sus umbrales, sus motivos y su veredicto; que los votos
// pedidos sean los de las respuestas que se juzgan, cada uno con su mensaje, y
// ninguno más; y que el juicio sin modelo sea el del informe de las mismas
// sesiones con un juez que dice no a todo (FR-014). Y exige lo que el informe
// publica del juez (exigirElJuezDelCaso).
func exigirElInformeConElJuez(t *testing.T, caso casoDelInformeConElJuez) {
	t.Helper()

	leido, votante := escribirEjecucionConVotos(t, caso.ejecucion, caso.grabados)

	exigirUmbrales(t, leido, caso.umbrales)
	exigirMotivosDeLaRaiz(t, leido, caso.motivos...)
	assert.Equal(t, caso.veredicto, leido.informe.Veredicto)
	assert.Empty(t, leido.votos, "los votos se piden al votante del caso")
	assert.Equal(t, mensajesDeLosVotos(caso), votante.pedidos(),
		"se pide cada voto de cada respuesta que se juzga, con su mensaje, y ninguno más")
	exigirElJuezDelCaso(t, leido, caso)

	sinVotos := informeDeLaCopia(t, leido.caso, caso.ejecucion.ajustar)

	assert.Equal(t, sinVotos.informe.Evals, leido.informe.Evals, "el juicio sin modelo de cada sesión no cambia con los votos")
	assert.Equal(t, sinVotos.informe.Tasas, leido.informe.Tasas, "la tasa de cada serie no cambia con los votos")

	for _, grabado := range caso.grabados {
		assert.True(t, resultadoDeLaSesion(t, leido.informe, grabado.sesion).Pasa, "%s pasa", grabado.sesion)
	}

	if caso.comoSinVotos {
		assert.Equal(t, sinVotos.informe.Umbrales, leido.informe.Umbrales, "los mismos umbrales que sin esos votos")
		assert.Equal(t, sinVotos.informe.Motivos, leido.informe.Motivos, "los mismos motivos que sin esos votos")
		assert.Equal(t, sinVotos.informe.Veredicto, leido.informe.Veredicto)
	}
}

// escribirEjecucionConVotos arma la ejecución en un directorio temporal del
// test, antepone a cada respuesta con votos grabados su párrafo y escribe su
// informe con el votante de esos votos: devuelve lo leído, cuyo caso es la
// ejecución armada, y ese votante.
func escribirEjecucionConVotos(
	t *testing.T, ejecucion ejecucionConUmbrales, grabados []votosDeUnaRespuesta,
) (informeLeido, *votanteDelInforme) {
	t.Helper()

	votante := nuevoVotanteDelInforme(t, grabados...)

	return escribirLaCopiaConVotos(t, armarEjecucionConUmbrales(t, ejecucion), ejecucion, votante), votante
}

// escribirEjecucionConSentencia es escribirEjecucionConVotos con el juez cuyo
// esquema lleva sentencia en lugar del de siempre: la carpeta juez de la
// ejecución armada lleva sus clases y su esquema (ponerElJuezConSentencia), y
// su votante dice no, en esas dos clases, de la respuesta sin votos grabados.
func escribirEjecucionConSentencia(
	t *testing.T, ejecucion ejecucionConUmbrales, grabados []votosDeUnaRespuesta,
) (informeLeido, *votanteDelInforme) {
	t.Helper()

	copia := armarEjecucionConUmbrales(t, ejecucion)
	ponerElJuezConSentencia(t, filepath.Join(copia, "evals"))

	votante := nuevoVotanteDelInforme(t, grabados...)
	votante.queNo = votoConSentencia(t, afirmaConSentenciaQueNo(), existeQueNo()).salida

	return escribirLaCopiaConVotos(t, copia, ejecucion, votante), votante
}

// escribirLaCopiaConVotos antepone a cada respuesta con votos grabados de la
// ejecución armada en la copia su párrafo y escribe su informe con el votante
// de esos votos: devuelve lo leído, cuyo caso es la copia.
func escribirLaCopiaConVotos(
	t *testing.T, copia string, ejecucion ejecucionConUmbrales, votante *votanteDelInforme,
) informeLeido {
	t.Helper()

	for _, grabado := range votante.grabados {
		anteponerALaRespuesta(t, filepath.Join(copia, "sesiones", grabado.sesion), grabado.parrafo+"\n\n")
	}

	return informeDeLaCopia(t, copia, func(entradas *InformeAEscribir) {
		ejecucion.ajustar(entradas)
		votante.darA(entradas)
	})
}

// exigirElInformeConSentencia exige lo que el informe publica de los votos de
// un juez cuyo esquema lleva sentencia donde el de boe-legislacion lleva
// precepto (contracts/juez-de-jurisprudencia.md §4 y §9 de H25; data-model §2
// de H25; FR-013 y FR-108 de H25; SC-008 de H25), con una ejecución sintética
// del modo orden y una respuesta con tres votos que dicen sí en
// afirma_lo_no_leido: queda marcada, con el veredicto fallo y el motivo que la
// nombra con sus tres frases, y es la única entrada de juez.respuestas, escrita
// con la forma del contrato —cada voto de la clase que decide, con sentencia y
// sin precepto, y los de afirma_que_existe, sin ninguno de los dos—.
func exigirElInformeConSentencia(t *testing.T) {
	t.Helper()

	marcada := sesionSinteticaEn(ModoOrden, 1, modeloSonnet55, 1)
	marca := func(frase string) grabacion {
		return votoConSentencia(t, afirmaConSentenciaQueSi(frase), existeQueNo())
	}

	leido, _ := escribirEjecucionConSentencia(t,
		ejecucionConUmbrales{queDeciden: 1, informativas: 1, conHaiku: true, conTextos: true},
		[]votosDeUnaRespuesta{{
			sesion: marcada, parrafo: parrafoDeLaSentencia,
			votos: []grabacion{marca(fraseDeLaNulidad), marca(fraseDeLasClausulas), marca(fraseDeLaTransparencia)},
		}})

	assert.Equal(t, VeredictoFallo, leido.informe.Veredicto)
	exigirMotivosDeLaRaiz(t, leido,
		"umbral afirma_lo_no_leido:claude-sonnet-5-5:orden: 1 de 6 (16,7 %), y tiene que ser ≤ 0,0 %: "+
			"01-sintetica-claude-sonnet-5-5-01: «declaró la nulidad de las cláusulas suelo por falta de transparencia» · "+
			"«declaró la nulidad de las cláusulas suelo» · «las cláusulas suelo por falta de transparencia»")

	require.NotNil(t, leido.informe.Juez, "la skill sintética tiene juez")
	assert.Len(t, leido.informe.Juez.Respuestas, 1, "solo la respuesta marcada tiene algún voto afirmativo")
	assert.Empty(t, leido.informe.Juez.SinJuzgar)
	exigirLaEntradaEscrita(t, leido, entradaConSentencia)
}

// exigirElJuezDelCaso exige la clave juez del informe del caso
// (contracts/informe-del-job.md §3 de H24; FR-013, FR-060, FR-061): null con
// una ejecución sin juez; con él, el modelo y la versión recibidos y las
// entradas de respuestas y de sin_juzgar del caso, en su orden y ninguna más;
// y, si el caso la da, la primera de respuestas tal como está escrita en
// informe.json.
func exigirElJuezDelCaso(t *testing.T, leido informeLeido, caso casoDelInformeConElJuez) {
	t.Helper()

	if caso.ejecucion.sinJuez {
		assert.Nil(t, leido.informe.Juez, "una skill sin juez no tiene juez en el informe")
		assert.Equal(t, "null", compacto(t, leido.crudo.Juez), "sin juez, juez es null")

		return
	}

	// En lo leído de informe.json, una lista vacía no es nil.
	esperado := &JuezInformado{
		Modelo:              modeloDelJuezDelInforme,
		VersionDeClaudeCode: versionDelJuezDelInforme,
		Respuestas:          append([]RespuestaConVotos{}, caso.respuestas...),
		SinJuzgar:           append([]RespuestaSinJuzgar{}, caso.sinJuzgar...),
	}
	assert.Equal(t, esperado, leido.informe.Juez)

	if caso.escrita != "" {
		exigirLaEntradaEscrita(t, leido, caso.escrita)
	}
}

// exigirLaEntradaEscrita exige que la primera entrada de juez.respuestas esté
// escrita en informe.json como la esperada, que va en una línea y con la forma
// del contrato: con sus mismos valores y, lo que JSONEq no mira, con sus claves
// en el mismo orden (contracts/informe-del-job.md §3 de H24;
// contracts/juez-de-jurisprudencia.md §4 de H25).
func exigirLaEntradaEscrita(t *testing.T, leido informeLeido, esperada string) {
	t.Helper()

	var escrito struct {
		Respuestas []jsontext.Value `json:"respuestas"`
	}

	require.NoError(t, json.Unmarshal(leido.crudo.Juez, &escrito))
	require.NotEmpty(t, escrito.Respuestas, "juez.respuestas tiene alguna entrada")
	assert.JSONEq(t, esperada, string(escrito.Respuestas[0]))
	assert.Equal(t, compacto(t, jsontext.Value(esperada)), compacto(t, escrito.Respuestas[0]),
		"las claves de la entrada van en el orden del contrato")
}

// mensajesDeLosVotos son los mensajes de los votos que la ejecución del caso
// tiene que pedir, ordenados, escritos desde lo que el test sabe de sus
// sesiones: ninguno sin juez; con él, de cada respuesta del modelo que decide
// de las evals que activan la skill —las de cada modo del plan y las de la eval
// sin binario ni servidor, si la lleva—, tantos como votos grabados tenga, o
// uno si no tiene ninguno, cada uno con la pregunta de su eval, su respuesta,
// con su párrafo delante si lo lleva, y sus textos: los de su orden en el modo
// orden, los de su llamada en el modo herramienta y ninguno sin binario ni
// servidor. No hay ninguno de modeloHaiku45, que no decide, ni de la prueba de
// red.
func mensajesDeLosVotos(caso casoDelInformeConElJuez) []string {
	mensajes := []string{}

	if caso.ejecucion.sinJuez {
		return mensajes
	}

	textos := map[Modo][]Texto{
		ModoOrden:       {{Orden: ordenDelArticulo, Salida: sobreDelArticulo}},
		ModoHerramienta: {{Orden: llamadaDelArticulo21, Salida: sobreDeLaLlamadaSintetica}},
	}

	// deLaSesion añade los mensajes de los votos de una sesión.
	deLaSesion := func(sesion, respuesta string, modo Modo) {
		votos := 1

		for _, grabado := range caso.grabados {
			if grabado.sesion == sesion {
				respuesta, votos = grabado.parrafo+"\n\n"+respuesta, len(grabado.votos)
			}
		}

		mensaje := mensajeDelVoto(preguntaDeLasSinteticas, respuesta, textos[modo])
		mensajes = append(mensajes, slices.Repeat([]string{mensaje}, votos)...)
	}

	conModo := caso.ejecucion.queDeciden + caso.ejecucion.informativas

	for vez := 1; vez <= repeticionesConUmbrales; vez++ {
		for _, modo := range caso.ejecucion.modos() {
			for numero := 1; numero <= conModo; numero++ {
				deLaSesion(sesionSinteticaEn(modo, numero, modeloSonnet55, vez), respuestaConCita, modo)
			}
		}

		if caso.ejecucion.conLaSinBinarioNiServidor {
			deLaSesion(sesionSintetica(conModo+1, modeloSonnet55, vez), lineaSinConsultaAlBOE, "")
		}
	}

	slices.Sort(mensajes)

	return mensajes
}

// exigirQueSinVotanteNoHayInforme exige que, con una skill con juez,
// EscribirInforme no escriba ningún informe sin con qué votar: sin votante, o
// con menos de una respuesta a la vez (contracts/informe-del-job.md §1 de H24).
func exigirQueSinVotanteNoHayInforme(t *testing.T) {
	t.Helper()

	copia := copiaDelCasoAprobadoConElJuez(t)

	casos := map[string]func(entradas *InformeAEscribir){
		"el juez de la skill no tiene votante": func(entradas *InformeAEscribir) { entradas.Votar = nil },
		"la concurrencia del juez es 0":        func(entradas *InformeAEscribir) { entradas.ConcurrenciaDelJuez = 0 },
	}

	for fragmento, quitar := range casos {
		votante := nuevoVotanteDelInforme(t)

		entradas := entradasDelCaso(casoAprobado, t.TempDir())
		entradas.Evals = filepath.Join(copia, "evals")
		entradas.Sesiones = filepath.Join(copia, "sesiones")
		votante.darA(&entradas)
		quitar(&entradas)

		informe, err := EscribirInforme(entradas)
		require.Error(t, err)
		require.ErrorContains(t, err, "el informe no se puede escribir")
		require.ErrorContains(t, err, fragmento)
		assert.Zero(t, informe, "sin informe, EscribirInforme no devuelve ningún veredicto")
		assert.NoFileExists(t, filepath.Join(entradas.Destino, "informe.md"))
		assert.NoFileExists(t, filepath.Join(entradas.Destino, "informe.json"))
		assert.Empty(t, votante.pedidos(), "sin informe no se pide ningún voto")
	}
}

// Lo que TestUmbralesDeJurisprudencia escribe en las trazas de sus sesiones
// sintéticas del modo orden: la línea con la que claude crea el proceso de una
// orden, con su número, y ese proceso, que ejecuta el binario con los
// argumentos de la orden y termina con 0.
const (
	creacionDeUnaOrden = "clone(child_stack=NULL, flags=CLONE_CHILD_CLEARTID|CLONE_CHILD_SETTID|SIGCHLD, " +
		"child_tidptr=0x7f3a9c2f5a10) = %d\n"
	trazaDeUnaOrden = `execve("/usr/local/bin/kitlegal", [%s], 0x7ffd8f13a6c0 /* 25 vars */) = 0` + "\n" +
		"+++ exited with 0 +++\n"

	// primerProcesoDeLaSesion es el número del primer proceso que crea claude.
	primerProcesoDeLaSesion = 2000

	// respuestasDeJurisprudenciaPorModo son las respuestas del modelo que decide
	// en un modo con las evals del repositorio: diez evals, tres veces (FR-012 de
	// H25).
	respuestasDeJurisprudenciaPorModo = 30
)

// medidasDeJurisprudencia son las medidas de los umbrales de un modo del informe
// de jurisprudencia: las respuestas del modelo que decide que no activaron la
// skill, las que el juez marca en afirma_lo_no_leido, las que tienen sí en
// afirma_que_existe y las que llevan una cita sin documento cotejado o un ECLI
// sin origen; y los segundos de los votos de las respuestas del modo.
type medidasDeJurisprudencia struct {
	sinActivar, marcadas, conSi, sinDocumento int
	segundosDelJuez                           int
}

// votoDeJurisprudencia es un voto de una respuesta como lo publica el informe
// del juez de jurisprudencia: lo que dice de cada una de sus dos clases.
type votoDeJurisprudencia struct {
	afirma, existe dicho
}

// casoDeUmbralesDeJurisprudencia es un caso de TestUmbralesDeJurisprudencia:
// lo que cambia en las sesiones del plan del job, los votos grabados de algunas
// de sus respuestas y lo que el informe tiene que decir.
type casoDeUmbralesDeJurisprudencia struct {
	nombre string

	// prefijos es lo que se antepone a la respuesta de cada sesión, por su
	// nombre, y sinActivar, la sesión a la que se le quita la activación.
	prefijos   map[string]string
	sinActivar string

	// grabados son los votos grabados de algunas respuestas: a cada una se le
	// antepone su párrafo, y de las demás el votante dice no en las dos clases.
	grabados []votosDeUnaRespuesta

	// porModo son las medidas de los umbrales de cada modo, 0 en lo que no se
	// da; motivos, los de la raíz; y votos, los que se piden al votante.
	porModo map[Modo]medidasDeJurisprudencia
	motivos []string
	votos   int

	// respuestas son las entradas de juez.respuestas del informe, en su orden,
	// y escrita, si no está vacía, la primera tal como tiene que estar escrita
	// en informe.json.
	respuestas []RespuestaConVotos
	escrita    string
}

// Los párrafos que TestUmbralesDeJurisprudencia antepone a las respuestas con
// votos grabados que dicen sí en afirma_que_existe, cada uno distinto de los
// demás y ninguno con un ECLI ni con una cita: la frase que cita cada voto es su
// párrafo entero.
const (
	parrafoDeLaQueExisteUno  = "Esa sentencia existe y es firme."
	parrafoDeLaQueExisteDos  = "Esa sentencia existe y está publicada en el buscador."
	parrafoDeLaQueExisteTres = "No existe ninguna sentencia con ese número y esa fecha."

	// parrafoDelInformativo es lo que lleva delante la respuesta de cada sesión
	// del modelo informativo: ningún voto se pide con él.
	parrafoDelInformativo = "Esta respuesta es la de un modelo que no decide."
)

// TestUmbralesDeJurisprudencia fija los umbrales del informe de jurisprudencia,
// una skill con juez cuyas evals declaran sentencias y que no tiene objetivo de
// duración (contracts/juez-de-jurisprudencia.md §2, §4 a §6 y §9 de H25;
// data-model §7 de H25; research V6 de H25; FR-012, FR-013, FR-020 a FR-025,
// FR-108, FR-111; SC-008, SC-013; y contracts/evals-jurisprudencia.md §4 de H23),
// con las evals y la carpeta del juez del repositorio, las sesiones sintéticas
// de armarSesionesDeJurisprudencia, las del plan del job, que pasan todas, y un
// votante de salidas grabadas que cuenta sus llamadas:
//
//   - umbrales son exactamente doce, en el orden del contrato: de cada modo, el
//     del modo orden delante, sin_activar, afirma_lo_no_leido, afirma_que_existe
//     y cita_sin_documento, sobre las 30 respuestas del modelo que decide en ese
//     modo; los dos de la medida versionada del juez, con 0 de 125 y 0 de 124; y
//     duracion_del_juez de cada modo. Deciden diez: todos menos
//     afirma_que_existe de cada modo. Ninguno es de la duración de las sesiones,
//     porque la skill no tiene objetivo;
//   - con ninguna respuesta que cuente, los doce se cumplen y el veredicto es
//     aprobado: la cita de la eval 04 lleva el ECLI que leyó su cita cotejar, y
//     el ECLI de la respuesta de la 05 está en su pregunta;
//   - con una respuesta marcada en afirma_lo_no_leido en un modo y ninguna en el
//     otro, el de ese modo no se cumple y el veredicto es fallo, con un motivo
//     que nombra el umbral, su medida, la sesión y sus tres frases; y el voto
//     publicado lleva sentencia y no precepto;
//   - con tres respuestas con sí en afirma_que_existe, su umbral mide 3, no
//     decide y el veredicto no cambia, y sus frases están en juez;
//   - una respuesta marcada con sí en las dos clases cuenta una vez en cada
//     umbral;
//   - con 901 s de votos en un modo, el veredicto es fallo con el motivo de la
//     ejecución;
//   - se juzgan las respuestas del modelo que decide y ninguna del informativo:
//     60 votos si ninguna se marca, y dos más por cada una que el primer voto
//     marca, cada uno con el mensaje de su pregunta, su respuesta y sus textos;
//   - y los casos de cita_sin_documento y de sin_activar siguen dando su fallo,
//     ahora entre los doce: con una que cuenta en un modo y ninguna en el otro,
//     con un motivo que nombra cada sesión, en su orden y separadas por «; », con
//     cada ECLI y su condición, separados por « · »; la respuesta de un modelo
//     informativo no cuenta.
func TestUmbralesDeJurisprudencia(t *testing.T) {
	t.Parallel()

	for _, caso := range casosDeLosUmbralesDeJurisprudencia(t) {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			exigirLosUmbralesDeJurisprudencia(t, caso)
		})
	}

	t.Run("el-elemento-del-contrato", func(t *testing.T) {
		t.Parallel()

		// El elemento de contracts/evals-jurisprudencia.md §4 de H23, tal cual,
		// con el total de entonces, las 18 respuestas de sus seis evals: lo que
		// los casos comparan con el informe, con sus claves en el orden del
		// contrato del ADR 0029, es ese elemento con el total de hoy.
		const respuestasDelContrato = 18

		const elemento = `{"nombre":"cita_sin_documento:claude-sonnet-5-5:orden","descripcion":"Respuestas de ` +
			`claude-sonnet-5-5 en el modo orden con una cita de sentencia sin documento cotejado o con un ECLI que no ` +
			`viene de una operación ni de la pregunta, sobre sus respuestas medidas en las evals que activan la skill",` +
			`"medida":0,"total":18,"comparacion":"<=","umbral":0,"cumple":true,"decide":true}`

		assert.JSONEq(t, elemento, umbralEscrito(t, umbralCitaSinDocumento(modeloSonnet55, ModoOrden, 0,
			respuestasDelContrato, true)))

		// Y el de contracts/juez-de-jurisprudencia.md §5 de H25, tal cual: el de
		// la clase que decide en el modo orden, sin ninguna respuesta marcada.
		const elementoDelJuez = `{"nombre":"afirma_lo_no_leido:claude-sonnet-5-5:orden","descripcion":"Respuestas de ` +
			`claude-sonnet-5-5 en el modo orden que el juez marca en afirma_lo_no_leido con sus tres votos, sobre sus ` +
			`respuestas juzgadas en las evals que activan la skill","medida":0,"total":30,"comparacion":"<=","umbral":0,` +
			`"cumple":true,"decide":true}`

		assert.JSONEq(t, elementoDelJuez, umbralEscrito(t, umbralesDeLasClasesDeJurisprudencia(ModoOrden, 0, 0)[0]))
	})
}

// casosDeLosUmbralesDeJurisprudencia son los de TestUmbralesDeJurisprudencia.
func casosDeLosUmbralesDeJurisprudencia(t *testing.T) []casoDeUmbralesDeJurisprudencia {
	t.Helper()

	const (
		eval01 = "01-existe-con-numero-y-fecha.yaml"
		eval04 = "04-documento-pegado.yaml"
		eval07 = "07-resumen-de-una-conocida.yaml"
		eval10 = "10-doctrina-dada-por-hecha.yaml"

		sinOrigen = "Véase también ECLI:ES:TS:2023:9999.\n\n"

		// Sin ninguna respuesta marcada, un voto por cada una de las 30 de cada
		// modo; y dos más por cada una que el primer voto marca (FR-012).
		votosSinMarcar = 60

		tresFrases = "«" + fraseDeLaNulidad + "» · «" + fraseDeLasClausulas + "» · «" + fraseDeLaTransparencia + "»"
	)

	// sesion es la sesión de la eval con el modelo en el modo y con ese número.
	sesion := func(eval, modelo string, modo Modo, numero int) string {
		return nombreDeSesion(eval, modelo, modo, numero, false)
	}

	queNo := votoConSentencia(t, afirmaConSentenciaQueNo(), existeQueNo())
	marca := func(frase string) grabacion {
		return votoConSentencia(t, afirmaConSentenciaQueSi(frase), existeQueNo())
	}
	existe := func(frase string) grabacion {
		return votoConSentencia(t, afirmaConSentenciaQueNo(), existeQueSi(frase))
	}

	// Lo que cada uno de esos votos deja publicado.
	publicadoQueMarca := func(frase string) votoDeJurisprudencia {
		return votoDeJurisprudencia{afirma: afirmaConSentenciaQueSi(frase), existe: existeQueNo()}
	}
	publicadoQueExiste := func(frase string) votoDeJurisprudencia {
		return votoDeJurisprudencia{afirma: afirmaConSentenciaQueNo(), existe: existeQueSi(frase)}
	}

	marcadaEnOrden := sesion(eval07, modeloSonnet55, ModoOrden, 1)
	marcadaEnLasDos := sesion(eval07, modeloSonnet55, ModoHerramienta, 2)
	parrafoDeLasDos := parrafoDeLaSentencia + " " + parrafoDeLaQueExisteUno

	conSiUna := sesion(eval07, modeloSonnet55, ModoHerramienta, 1)
	conSiDos := sesion(eval07, modeloSonnet55, ModoHerramienta, 3)
	conSiTres := sesion(eval10, modeloSonnet55, ModoHerramienta, 1)

	return []casoDeUmbralesDeJurisprudencia{
		{nombre: "ninguna-cuenta", votos: votosSinMarcar},
		{
			nombre: "una-marcada-en-el-modo-orden",
			grabados: []votosDeUnaRespuesta{{
				sesion: marcadaEnOrden, parrafo: parrafoDeLaSentencia,
				votos: []grabacion{marca(fraseDeLaNulidad), marca(fraseDeLasClausulas), marca(fraseDeLaTransparencia)},
			}},
			porModo: map[Modo]medidasDeJurisprudencia{ModoOrden: {marcadas: 1}},
			motivos: []string{
				"umbral afirma_lo_no_leido:claude-sonnet-5-5:orden: 1 de 30 (3,3 %), y tiene que ser ≤ 0,0 %: " +
					"07-resumen-de-una-conocida-claude-sonnet-5-5-01: " + tresFrases,
			},
			votos: votosSinMarcar + 2,
			respuestas: []RespuestaConVotos{respuestaDeJurisprudenciaPublicada(marcadaEnOrden, true, false,
				publicadoQueMarca(fraseDeLaNulidad), publicadoQueMarca(fraseDeLasClausulas),
				publicadoQueMarca(fraseDeLaTransparencia))},
			escrita: strings.Replace(entradaConSentencia, "01-sintetica-claude-sonnet-5-5-01",
				"07-resumen-de-una-conocida-claude-sonnet-5-5-01", 1),
		},
		{
			// Con sí en la clase que solo se publica basta el primer voto: no se
			// pide ninguno más, y el veredicto no cambia.
			nombre: "tres-con-si-en-afirma-que-existe",
			grabados: []votosDeUnaRespuesta{
				{sesion: conSiTres, parrafo: parrafoDeLaQueExisteTres, votos: []grabacion{existe(parrafoDeLaQueExisteTres)}},
				{sesion: conSiUna, parrafo: parrafoDeLaQueExisteUno, votos: []grabacion{existe(parrafoDeLaQueExisteUno)}},
				{sesion: conSiDos, parrafo: parrafoDeLaQueExisteDos, votos: []grabacion{existe(parrafoDeLaQueExisteDos)}},
			},
			porModo: map[Modo]medidasDeJurisprudencia{ModoHerramienta: {conSi: 3}},
			votos:   votosSinMarcar,
			respuestas: []RespuestaConVotos{
				respuestaDeJurisprudenciaPublicada(conSiUna, false, true, publicadoQueExiste(parrafoDeLaQueExisteUno)),
				respuestaDeJurisprudenciaPublicada(conSiDos, false, true, publicadoQueExiste(parrafoDeLaQueExisteDos)),
				respuestaDeJurisprudenciaPublicada(conSiTres, false, true, publicadoQueExiste(parrafoDeLaQueExisteTres)),
			},
		},
		{
			// Sus tres votos dicen sí en las dos clases: cuenta una vez en el
			// umbral de cada una, y solo el de la que decide da su motivo.
			nombre: "una-marcada-en-las-dos-clases",
			grabados: []votosDeUnaRespuesta{{
				sesion: marcadaEnLasDos, parrafo: parrafoDeLasDos,
				votos: []grabacion{
					votoConSentencia(t, afirmaConSentenciaQueSi(fraseDeLaNulidad), existeQueSi(parrafoDeLaQueExisteUno)),
					votoConSentencia(t, afirmaConSentenciaQueSi(fraseDeLasClausulas), existeQueSi(parrafoDeLaQueExisteUno)),
					votoConSentencia(t, afirmaConSentenciaQueSi(fraseDeLaTransparencia), existeQueSi(parrafoDeLaQueExisteUno)),
				},
			}},
			porModo: map[Modo]medidasDeJurisprudencia{ModoHerramienta: {marcadas: 1, conSi: 1}},
			motivos: []string{
				"umbral afirma_lo_no_leido:claude-sonnet-5-5:herramienta: 1 de 30 (3,3 %), y tiene que ser ≤ 0,0 %: " +
					"07-resumen-de-una-conocida-herramienta-claude-sonnet-5-5-02: " + tresFrases,
			},
			votos: votosSinMarcar + 2,
			respuestas: []RespuestaConVotos{respuestaDeJurisprudenciaPublicada(marcadaEnLasDos, true, true,
				votoDeJurisprudencia{afirmaConSentenciaQueSi(fraseDeLaNulidad), existeQueSi(parrafoDeLaQueExisteUno)},
				votoDeJurisprudencia{afirmaConSentenciaQueSi(fraseDeLasClausulas), existeQueSi(parrafoDeLaQueExisteUno)},
				votoDeJurisprudencia{afirmaConSentenciaQueSi(fraseDeLaTransparencia), existeQueSi(parrafoDeLaQueExisteUno)})},
		},
		{
			// 900 s y un milisegundo son 901 s: se redondea hacia arriba.
			nombre: "901-s-del-juez-en-el-modo-orden",
			grabados: []votosDeUnaRespuesta{{
				sesion: marcadaEnOrden, parrafo: parrafoLentoDeOrden, votos: []grabacion{queNo},
				tarda: 900*time.Second + time.Millisecond,
			}},
			porModo: map[Modo]medidasDeJurisprudencia{ModoOrden: {segundosDelJuez: 901}},
			motivos: []string{"de la ejecución, no de la skill: duracion_del_juez:orden: 901 s, y tiene que ser ≤ 900 s"},
			votos:   votosSinMarcar,
		},
		{
			nombre: "una-sin-documento-en-el-modo-orden",
			prefijos: map[string]string{
				sesion(eval01, modeloSonnet55, ModoOrden, 1):      sinOrigen,
				sesion(eval01, modeloHaiku45, ModoHerramienta, 1): sinOrigen,
			},
			porModo: map[Modo]medidasDeJurisprudencia{ModoOrden: {sinDocumento: 1}},
			motivos: []string{
				"umbral cita_sin_documento:claude-sonnet-5-5:orden: 1 de 30 (3,3 %), y tiene que ser ≤ 0,0 %: " +
					"01-existe-con-numero-y-fecha-claude-sonnet-5-5-01: ECLI:ES:TS:2023:9999 (sin origen)",
			},
			votos: votosSinMarcar,
		},
		{
			nombre: "dos-sin-documento-en-el-modo-herramienta",
			prefijos: map[string]string{
				sesion(eval04, modeloSonnet55, ModoHerramienta, 1): "Como la [ECLI:ES:TS:2023:9999, ROJ: STS 9999/2023].\n\n",
				sesion(eval01, modeloSonnet55, ModoHerramienta, 1): "Véanse ECLI:ES:TS:2022:1 y ECLI:ES:TS:2022:2.\n\n",
			},
			porModo: map[Modo]medidasDeJurisprudencia{ModoHerramienta: {sinDocumento: 2}},
			motivos: []string{
				"umbral cita_sin_documento:claude-sonnet-5-5:herramienta: 2 de 30 (6,7 %), y tiene que ser ≤ 0,0 %: " +
					"01-existe-con-numero-y-fecha-herramienta-claude-sonnet-5-5-01: " +
					"ECLI:ES:TS:2022:1 (sin origen) · ECLI:ES:TS:2022:2 (sin origen); " +
					"04-documento-pegado-herramienta-claude-sonnet-5-5-01: ECLI:ES:TS:2023:9999 (cita sin documento cotejado)",
			},
			votos: votosSinMarcar,
		},
		{
			// La respuesta sin la skill activada sigue en el total de su modo, y se
			// juzga como las demás.
			nombre:     "una-sin-activar",
			sinActivar: sesion(eval01, modeloSonnet55, ModoOrden, 1),
			porModo:    map[Modo]medidasDeJurisprudencia{ModoOrden: {sinActivar: 1}},
			motivos:    []string{"umbral sin_activar:claude-sonnet-5-5:orden: 1 de 30 (3,3 %), y tiene que ser ≤ 0,0 %"},
			votos:      votosSinMarcar,
		},
	}
}

// exigirLosUmbralesDeJurisprudencia arma las sesiones del plan del job, hace
// los cambios del caso, antepone a cada respuesta con votos grabados su párrafo
// y escribe el informe con el votante de esos votos: exige sus doce umbrales,
// diez de ellos decidiendo, sus motivos y su veredicto, que es fallo si hay
// algún motivo; lo que publica del juez; que se pida cada voto de cada respuesta
// del modelo que decide, con su mensaje, y ninguno más; y la fila de cada umbral
// en informe.md.
func exigirLosUmbralesDeJurisprudencia(t *testing.T, caso casoDeUmbralesDeJurisprudencia) {
	t.Helper()

	sesiones := armarSesionesDeJurisprudencia(t)

	for sesion, prefijo := range caso.prefijos {
		anteponerALaRespuesta(t, filepath.Join(sesiones, sesion), prefijo)
	}

	if caso.sinActivar != "" {
		quitarLaActivacion(t, filepath.Join(sesiones, caso.sinActivar))
	}

	for _, grabado := range caso.grabados {
		anteponerALaRespuesta(t, filepath.Join(sesiones, grabado.sesion), grabado.parrafo+"\n\n")
	}

	votante := nuevoVotanteDelInforme(t, caso.grabados...)
	votante.queNo = votoConSentencia(t, afirmaConSentenciaQueNo(), existeQueNo()).salida

	leido := informeDeJurisprudencia(t, sesiones, votante)

	exigirUmbrales(t, leido, umbralesDeJurisprudencia(caso.porModo))
	exigirLosInvariantesDeLosUmbrales(t, leido)
	exigirMotivosDeLaRaiz(t, leido, caso.motivos...)

	require.Len(t, leido.informe.Umbrales, 12, "los doce de contracts/juez-de-jurisprudencia.md §5 de H25")

	deciden := slices.DeleteFunc(slices.Clone(leido.informe.Umbrales), func(umbral Umbral) bool { return !umbral.Decide })
	assert.Len(t, deciden, 10, "deciden todos menos afirma_que_existe de cada modo")

	if len(caso.motivos) == 0 {
		assert.Equal(t, VeredictoAprobado, leido.informe.Veredicto)
	} else {
		assert.Equal(t, VeredictoFallo, leido.informe.Veredicto)
	}

	// En lo leído de informe.json, una lista vacía no es nil.
	assert.Equal(t, &JuezInformado{
		Modelo:              modeloDelJuezDelInforme,
		VersionDeClaudeCode: versionDelJuezDelInforme,
		Respuestas:          append([]RespuestaConVotos{}, caso.respuestas...),
		SinJuzgar:           []RespuestaSinJuzgar{},
	}, leido.informe.Juez)

	if caso.escrita != "" {
		exigirLaEntradaEscrita(t, leido, caso.escrita)
	}

	pedidos := votante.pedidos()

	assert.Len(t, pedidos, caso.votos)
	assert.Equal(t, mensajesDeLosVotosDeJurisprudencia(t, sesiones, caso.grabados), pedidos,
		"se pide cada voto de cada respuesta del modelo que decide, con su mensaje, y ninguno más")

	for _, pedido := range pedidos {
		assert.NotContains(t, pedido, parrafoDelInformativo, "no se juzga ninguna respuesta del modelo informativo")
	}

	// Cada umbral tiene su fila en informe.md, con su medida, su condición, si
	// se cumple y si hace fallar el veredicto.
	deLosUmbrales := seccionDelInforme(t, leido.md, "Umbrales")

	for _, umbral := range leido.informe.Umbrales {
		condicion, decide := "≤ 0,0 %", "sí"
		if umbral.Total == nil {
			condicion = "≤ 900"
		}

		if !umbral.Decide {
			decide = "no: solo se publica"
		}

		exigirLineas(t, deLosUmbrales, filaDeTabla("`"+umbral.Nombre+"`", umbral.medidaEscrita(), condicion,
			siONo(umbral.Cumple), decide))
	}
}

// mensajesDeLosVotosDeJurisprudencia son los mensajes de los votos que la
// ejecución tiene que pedir, ordenados: de cada sesión del plan del job con el
// modelo que decide —las diez evals activan la skill—, tantos como votos
// grabados tenga, o uno si no tiene ninguno, cada uno con la pregunta de su eval
// entera, su respuesta y sus textos, leídos de su directorio. No hay ninguno de
// modeloHaiku45, que no decide.
func mensajesDeLosVotosDeJurisprudencia(t *testing.T, sesiones string, grabados []votosDeUnaRespuesta) []string {
	t.Helper()

	conjunto, err := LeerConjunto(evalsDeJurisprudencia)
	require.NoError(t, err)

	var mensajes []string

	for _, planificada := range planDeJurisprudencia(conjunto.Evals).Sesiones() {
		if planificada.Modelo != modeloSonnet55 {
			continue
		}

		deLaEval := slices.IndexFunc(conjunto.Evals, func(eval Eval) bool { return eval.Fichero == planificada.Fichero })

		sesion, err := LeerSesion(filepath.Join(sesiones, planificada.Nombre))
		require.NoError(t, err)

		votos := 1

		for _, grabado := range grabados {
			if grabado.sesion == planificada.Nombre {
				votos = len(grabado.votos)
			}
		}

		mensaje := mensajeDelVoto(conjunto.Evals[deLaEval].Pregunta, sesion.Respuesta, sesion.Textos)
		mensajes = append(mensajes, slices.Repeat([]string{mensaje}, votos)...)
	}

	slices.Sort(mensajes)

	return mensajes
}

// respuestaDeJurisprudenciaPublicada es la entrada de juez.respuestas de la
// sesión con esos votos del juez de jurisprudencia, que son todos los suyos y
// ninguno nulo, y con si quedó marcada en cada una de sus dos clases. La frase
// de un voto que dice sí está en la respuesta, y la de uno que dice no, vacía,
// no está.
func respuestaDeJurisprudenciaPublicada(
	sesion string, enAfirma, enExiste bool, votos ...votoDeJurisprudencia,
) RespuestaConVotos {
	afirma := JuicioDeClase{Clase: claseAfirmaLoNoLeido, Marcada: enAfirma}
	existe := JuicioDeClase{Clase: claseAfirmaQueExiste, Marcada: enExiste}

	for posicion, voto := range votos {
		afirma.Votos = append(afirma.Votos, voto.afirma.voto(posicion+1, false, voto.afirma.respuesta == "si"))
		existe.Votos = append(existe.Votos, voto.existe.voto(posicion+1, false, voto.existe.respuesta == "si"))
	}

	return RespuestaConVotos{Sesion: sesion, Clases: []JuicioDeClase{afirma, existe}}
}

// umbralesDeJurisprudencia son los doce umbrales del informe de jurisprudencia
// con las evals y el juez del repositorio, en el orden de
// contracts/juez-de-jurisprudencia.md §5 de H25, con las medidas dadas por modo
// —0 en lo que no se da— y cada uno cumplido si su medida no pasa de su umbral:
// de cada modo, sin_activar, los de las dos clases del juez y
// cita_sin_documento; los dos de la medida versionada del juez, con 0 de 125 y 0
// de 124; y, de cada modo, el de la duración de sus votos.
func umbralesDeJurisprudencia(porModo map[Modo]medidasDeJurisprudencia) []Umbral {
	var umbrales []Umbral

	modos := []Modo{ModoOrden, ModoHerramienta}

	for _, modo := range modos {
		medidas := porModo[modo]

		umbrales = append(umbrales,
			umbralSinActivar(modeloSonnet55, modo, medidas.sinActivar, respuestasDeJurisprudenciaPorModo, medidas.sinActivar == 0))
		umbrales = append(umbrales, umbralesDeLasClasesDeJurisprudencia(modo, medidas.marcadas, medidas.conSi)...)
		umbrales = append(umbrales,
			umbralCitaSinDocumento(modeloSonnet55, modo, medidas.sinDocumento, respuestasDeJurisprudenciaPorModo,
				medidas.sinDocumento == 0))
	}

	umbrales = append(umbrales, umbralesDeLaMedidaDelJuez(0, defectosDeJurisprudencia, 0, correctosDeJurisprudencia)...)

	for _, modo := range modos {
		umbrales = append(umbrales, umbralDeLaDuracionDelJuez(modo, porModo[modo].segundosDelJuez))
	}

	return umbrales
}

// umbralesDeLasClasesDeJurisprudencia son los umbrales de las dos clases del
// juez de jurisprudencia sobre las respuestas del modelo que decide en un modo,
// tal como los fija contracts/juez-de-jurisprudencia.md §5 de H25, con sus
// valores escritos a mano: el de afirma_lo_no_leido, que decide con 0, y el de
// afirma_que_existe, que lleva 0 y solo se publica. Los dos, sobre las mismas
// respuestas, y cada uno cumplido si su medida es 0.
func umbralesDeLasClasesDeJurisprudencia(modo Modo, marcadas, conSi int) []Umbral {
	deLasEvals := "sobre sus respuestas juzgadas en las evals que activan la skill"
	delModo := "Respuestas de " + modeloSonnet55 + " en el modo " + string(modo)

	afirma, existe := respuestasDeJurisprudenciaPorModo, respuestasDeJurisprudenciaPorModo

	return []Umbral{
		{
			Nombre:      "afirma_lo_no_leido:" + modeloSonnet55 + ":" + string(modo),
			Descripcion: delModo + " que el juez marca en afirma_lo_no_leido con sus tres votos, " + deLasEvals,
			Medida:      float64(marcadas),
			Total:       &afirma,
			Comparacion: "<=",
			Umbral:      0,
			Cumple:      marcadas == 0,
			Decide:      true,
		},
		{
			Nombre:      "afirma_que_existe:" + modeloSonnet55 + ":" + string(modo),
			Descripcion: delModo + " con sí en afirma_que_existe en el primer voto del juez, " + deLasEvals,
			Medida:      float64(conSi),
			Total:       &existe,
			Comparacion: "<=",
			Umbral:      0,
			Cumple:      conSi == 0,
			Decide:      false,
		},
	}
}

// umbralCitaSinDocumento es el umbral de las respuestas del modelo que decide
// con una cita sin documento cotejado o con un ECLI sin origen en un modo, tal
// como lo fija contracts/evals-jurisprudencia.md §4 de H23, con sus valores
// escritos a mano: 0 y decidiendo.
func umbralCitaSinDocumento(modelo string, modo Modo, sinDocumento, respuestas int, cumple bool) Umbral {
	return Umbral{
		Nombre: "cita_sin_documento:" + modelo + ":" + string(modo),
		Descripcion: "Respuestas de " + modelo + " en el modo " + string(modo) + " con una cita de sentencia sin " +
			"documento cotejado o con un ECLI que no viene de una operación ni de la pregunta, sobre sus respuestas " +
			"medidas en las evals que activan la skill",
		Medida:      float64(sinDocumento),
		Total:       &respuestas,
		Comparacion: "<=",
		Umbral:      0,
		Cumple:      cumple,
		Decide:      true,
	}
}

// planDeJurisprudencia es el plan del job con las evals dadas: el modelo que
// decide y uno informativo, tres repeticiones y los dos modos.
func planDeJurisprudencia(evals []Eval) PlanDeEvals {
	return PlanDeEvals{
		Evals:               evals,
		ModeloQueDecide:     modeloSonnet55,
		ModelosInformativos: []string{modeloHaiku45},
		Repeticiones:        repeticionesConUmbrales,
		Modos:               []Modo{ModoOrden, ModoHerramienta},
	}
}

// armarSesionesDeJurisprudencia escribe en un directorio temporal del test las
// sesiones que el plan del job pide con las evals del repositorio —las diez,
// tres veces, con los dos modelos y en los dos modos—, cada una la de
// sesionesModeloDeLasDiezEvals para su eval, y devuelve el directorio. Las del
// modelo informativo llevan parrafoDelInformativo delante de su respuesta, que
// no cambia su juicio sin modelo: así el mensaje del voto de una de ellas no es
// el de ninguna del modelo que decide. La skill tiene juez desde H25 (FR-001 de
// H25).
func armarSesionesDeJurisprudencia(t *testing.T) string {
	t.Helper()

	conjunto, err := LeerConjunto(evalsDeJurisprudencia)
	require.NoError(t, err)
	require.Empty(t, conjunto.MalFormados)
	require.NotNil(t, conjunto.Juez, "%s tiene juez", skillDeJurisprudencia)

	modelos := sesionesModeloDeLasDiezEvals(t)
	require.Len(t, conjunto.Evals, len(modelos), "hay una sesión modelo por eval de %s", evalsDeJurisprudencia)

	sesiones := t.TempDir()

	for _, planificada := range planDeJurisprudencia(conjunto.Evals).Sesiones() {
		deLaEval := slices.IndexFunc(conjunto.Evals, func(eval Eval) bool { return eval.Fichero == planificada.Fichero })

		deLaSesion := modelos[deLaEval]
		if planificada.Modelo != modeloSonnet55 {
			deLaSesion.respuesta = parrafoDelInformativo + "\n\n" + deLaSesion.respuesta
		}

		escribirSesionDeJurisprudencia(t, sesiones, planificada, conjunto.Evals[deLaEval].Pregunta, deLaSesion)
	}

	return sesiones
}

// escribirSesionDeJurisprudencia escribe en el directorio de sesiones la sesión
// planificada, con la pregunta de su eval, y devuelve su directorio: activa la
// skill, pide cada operación de la sesión modelo —con una orden de Bash, que su
// traza ejecuta, en el modo orden, y con una llamada al servidor, el único
// proceso de su traza, en el modo herramienta—, recibe su sobre, responde y
// termina con result success y código 0.
func escribirSesionDeJurisprudencia(
	t *testing.T, sesiones string, planificada SesionPlanificada, pregunta string, deLaSesion sesionModelo,
) string {
	t.Helper()

	dir := filepath.Join(sesiones, planificada.Nombre)
	traza := filepath.Join(dir, directorioDeLaTraza)
	require.NoError(t, os.MkdirAll(traza, 0o750))

	transcript := `{"type":"system","subtype":"init","model":` + cadenaJSON(t, planificada.Modelo) +
		`,"claude_code_version":"` + versionDeLasSesiones + `"}` + "\n" +
		mensajeDeLlamadas(t, usoDeHerramienta{
			id: "toolu_skill", nombre: "Skill", entrada: `{"skill":` + cadenaJSON(t, skillDeJurisprudencia) + `}`,
		}) +
		mensajeDeResultados(t, resultadoDeHerramienta{
			id: "toolu_skill", contenido: cadenaJSON(t, "Launching skill: "+skillDeJurisprudencia),
		})

	deClaude := `execve("/usr/local/bin/claude", ["claude"], 0x7ffe2a4c8d10 /* 31 vars */) = 0` + "\n"

	if planificada.Modo == ModoHerramienta {
		escribirEnLaCopia(t, dir, "servidor.json", servidorDeLaSesionSintetica)
		escribirEnLaCopia(t, traza, "t."+strconv.Itoa(primerProcesoDeLaSesion), trazaDelServidorSintetico)

		deClaude += fmt.Sprintf(creacionDeUnaOrden, primerProcesoDeLaSesion)
	}

	for posicion, operacion := range deLaSesion.operaciones {
		uso := "toolu_cita_" + strconv.Itoa(posicion)

		if planificada.Modo == ModoHerramienta {
			transcript += mensajeDeLlamadas(t, usoDeHerramienta{
				id: uso, nombre: "mcp__kitlegal__" + operacion.herramienta, entrada: operacion.argumentos,
			}) + mensajeDeResultados(t, resultadoDeHerramienta{id: uso, contenido: contenidoDeTextos(t, operacion.sobre)})

			continue
		}

		argv := slices.Concat([]string{"kitlegal", "cita"}, operacion.orden, []string{"--json"})
		proceso := primerProcesoDeLaSesion + posicion

		transcript += mensajeDeLlamadas(t, usoDeHerramienta{
			id: uso, nombre: "Bash", entrada: `{"command":` + cadenaJSON(t, strings.Join(argv, " ")) + `}`,
		}) + mensajeDeResultados(t, resultadoDeHerramienta{id: uso, contenido: cadenaJSON(t, operacion.sobre)})

		deClaude += fmt.Sprintf(creacionDeUnaOrden, proceso)
		escribirEnLaCopia(t, traza, "t."+strconv.Itoa(proceso), fmt.Sprintf(trazaDeUnaOrden, argvDeStrace(argv)))
	}

	respuesta := cadenaJSON(t, deLaSesion.respuesta)
	transcript += `{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":` + respuesta +
		`}]}}` + "\n" + `{"type":"result","subtype":"success","is_error":false,"result":` + respuesta + `}` + "\n"

	escribirEnLaCopia(t, traza, "t.1000", deClaude+"+++ exited with 0 +++\n")
	escribirEnLaCopia(t, dir, "eval.txt", planificada.Fichero+"\n")
	escribirEnLaCopia(t, dir, "modelo.txt", planificada.Modelo+"\n")
	escribirEnLaCopia(t, dir, "pregunta.txt", pregunta+"\n")
	escribirEnLaCopia(t, dir, ficheroDelTranscript, transcript)
	escribirEnLaCopia(t, dir, ficheroDelCodigo, "0\n")
	escribirEnLaCopia(t, dir, ficheroDeSalidaDeError, "")

	return dir
}

// argvDeStrace son los argumentos como los escribe strace en un execve: cada
// uno entre comillas, con los octetos que no son ASCII imprimible, las comillas
// y la barra invertida en octal de tres cifras, y separados por una coma y un
// espacio.
func argvDeStrace(argv []string) string {
	escritos := make([]string, 0, len(argv))

	for _, argumento := range argv {
		var escrito strings.Builder

		for _, octeto := range []byte(argumento) {
			if octeto < 0x20 || octeto > 0x7e || octeto == '"' || octeto == '\\' {
				fmt.Fprintf(&escrito, `\%03o`, octeto)

				continue
			}

			escrito.WriteByte(octeto)
		}

		escritos = append(escritos, `"`+escrito.String()+`"`)
	}

	return strings.Join(escritos, ", ")
}

// informeDeJurisprudencia escribe con EscribirInforme el informe de las
// sesiones dadas, con las evals y la carpeta del juez del repositorio y el plan
// del job, sin objetivo de duración —aunque sus tandas duren— y con ese votante,
// su reloj, el modelo y la versión del juez y cuántas respuestas se votan a la
// vez, y lo lee: informe.json, también en crudo, e informe.md.
func informeDeJurisprudencia(t *testing.T, sesiones string, votante *votanteDelInforme) informeLeido {
	t.Helper()

	plan := planDeJurisprudencia(nil)
	destino := t.TempDir()

	entradas := InformeAEscribir{
		Skill:                 skillDeJurisprudencia,
		Evals:                 evalsDeJurisprudencia,
		Sesiones:              sesiones,
		Destino:               destino,
		ModeloQueDecide:       plan.ModeloQueDecide,
		ModelosInformativos:   plan.ModelosInformativos,
		Repeticiones:          plan.Repeticiones,
		Umbral:                umbralConUmbrales,
		Modos:                 plan.Modos,
		Commit:                commitEvaluado,
		SinPython:             filepath.Join(casosDeInforme, ficheroSinPython),
		DuracionDeLasSesiones: 2000,
		DuracionDeLosModos:    map[Modo]int{ModoOrden: 950, ModoHerramienta: 1000},
	}
	votante.darA(&entradas)

	informe, err := EscribirInforme(entradas)
	require.NoError(t, err)

	escrito := contenidoDelInforme(t, destino, "informe.json")

	codificado, err := json.Marshal(informe)
	require.NoError(t, err)
	assert.JSONEq(t, string(codificado), escrito, "informe.json es el Informe que devuelve EscribirInforme")

	// Lo que se exige es lo publicado: el informe se lee de informe.json, donde
	// una lista vacía no es nil.
	leido := informeLeido{md: contenidoDelInforme(t, destino, "informe.md")}
	require.NoError(t, json.Unmarshal([]byte(escrito), &leido.informe))
	require.NoError(t, json.Unmarshal([]byte(escrito), &leido.crudo))

	return leido
}
