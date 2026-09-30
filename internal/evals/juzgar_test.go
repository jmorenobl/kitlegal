package evals

import (
	"cmp"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Lo que juzgan las evals y las sesiones de TestJuzgar, construidas en memoria.
const (
	ficheroDeLaEval01     = "01-lpac-articulo-21.yaml"
	ficheroDeNoActivacion = "11-no-activa-programacion.yaml"
	normaDeLaLCSP         = "BOE-A-2017-12902"
	normaDeLaLRBRL        = "BOE-A-1985-5392"

	// ordenDelArticulo21 es la orden de la lectura del bloque de la eval 01, y
	// textoDelComando21 y textoDeLaCita21, el comando esperado y la cita
	// esperada de esa eval, con el texto con el que los presenta el informe
	// (contrato job-de-evals §5).
	ordenDelArticulo21 = "boe articulo BOE-A-2015-10565 a21 --json"
	textoDelComando21  = "bloque boe BOE-A-2015-10565 a21"
	textoDeLaCita21    = "BOE-A-2015-10565 a21"

	// avisoDeDerogada y avisoDeVigenciaAgotada son las frases del binario para
	// esos dos códigos, con su forma fija, como las traslada una respuesta.
	avisoDeDerogada        = "⚠ NORMA DEROGADA: esta norma ha sido derogada."
	avisoDeVigenciaAgotada = "⚠ VIGENCIA AGOTADA: esta norma ya no está en vigor."
)

// Lo que juzgan las evals de territorio y sus sesiones de TestJuzgar,
// construidas en memoria (contrato de evals §2 de H6).
const (
	// skillDeTerritorio es la skill cuyas evals esperan el territorio, y
	// territorioDeLaSkillInstalada, el nombre de invocación con el que la skill
	// instalada invoca el applet territorio.
	skillDeTerritorio            = "legal-core"
	territorioDeLaSkillInstalada = "/home/runner/.claude/skills/legal-core/scripts/territorio"

	ficheroDelMunicipioCubierto = "01-territorio-municipio-cubierto.yaml"

	// textoDelComandoDeLeganes es el comando de territorio de la eval del
	// municipio cubierto con el texto con el que lo presenta el informe, y
	// ordenDeLeganes, la orden de la invocación que lo satisface.
	textoDelComandoDeLeganes = "territorio resolver Leganés"
	ordenDeLeganes           = "territorio resolver Leganés --json"

	// respuestaDelMunicipio declara cada elemento del territorio esperado de esa
	// eval con su forma fija.
	respuestaDelMunicipio = "Leganés está en la provincia de Madrid, en la Comunidad de Madrid.\n\n" +
		"Sus normas se publican en el BOE y en el BOCM.\n\n" +
		"- boletin_autonomico: configurado\n- boletin_provincial: configurado\n- dir3: verificado"
)

// Lo que juzgan la eval de la consulta repetida y sus sesiones de TestJuzgar,
// construidas en memoria (contrato evals-y-skill §2 y §4 de H7; desde H7.2, con
// la forma de la eval 19 del repositorio, contrato eval-y-derivada §1 y §5).
const (
	ficheroDeLaConsultaRepetida = "19-lcsp-contrato-menor-redaccion-cambiada.yaml"

	// bloqueDelArticulo118 es el bloque que lee y cita esa eval, el art. 118 de
	// la LCSP; ordenDelArticulo118 es la orden de su lectura, y
	// textoDelComando118 y textoDeLaCita118, su comando esperado y su cita
	// esperada, con el texto con el que los presenta el informe.
	bloqueDelArticulo118 = "a1-30"
	ordenDelArticulo118  = "boe articulo BOE-A-2017-12902 a1-30 --json"
	textoDelComando118   = "bloque boe BOE-A-2017-12902 a1-30"
	textoDeLaCita118     = "BOE-A-2017-12902 a1-30"

	// respuestaSinLaCita118 habla del art. 118 de la LCSP sin citarlo, y
	// respuestaConLaCita118 lleva además la cita esperada de esa eval.
	respuestaSinLaCita118 = "El art\xc3\xadculo 118 de la LCSP regula el expediente de los contratos menores."
	respuestaConLaCita118 = respuestaSinLaCita118 + "\n\n[BOE-A-2017-12902, bloque a1-30]"

	// textoDeLaComprobacion es el comando de comprobación sin norma, el de la
	// eval de H7, y textoDeGraphShow, el prohibido de esa eval, con el texto con
	// el que los presentan el informe y los motivos.
	textoDeLaComprobacion = "graph check"
	textoDeGraphShow      = "graph show"

	// ordenDeLaComprobacion y ordenDeGraphShow son las órdenes de las
	// invocaciones de la skill que comprueban el grafo y que piden la ficha del
	// bloque.
	ordenDeLaComprobacion = "graph check --json"
	ordenDeGraphShow      = "graph show eli/es/l/2015/10/01/39#a21 --json"
)

// Lo que juzgan la eval de la redacción cambiada y sus sesiones de TestJuzgar,
// construidas en memoria (contrato evals-y-skill §2 y §5 de H7.1).
const (
	// versionObsoleta es el hallazgo que esa eval espera, y
	// motivoDeVersionObsoleta, el motivo con el que queda ausente.
	versionObsoleta         = "version-obsoleta"
	motivoDeVersionObsoleta = "forma de hallazgo ausente: version-obsoleta"

	// textoDeLaComprobacionDeLaNorma es su comando de comprobación, el de la
	// consulta repetida desde H7.2, con el texto con el que lo presentan el
	// informe y los motivos; y ordenDeLaComprobacionDeLaNorma, la orden de la
	// invocación de la skill que comprueba la memoria con la norma y el bloque
	// que ha leído.
	textoDeLaComprobacionDeLaNorma = "graph check BOE-A-2017-12902"
	ordenDeLaComprobacionDeLaNorma = "graph check BOE-A-2017-12902 a1-30 --json"

	// trasladoDelCambio traslada el hallazgo con su forma fija, con el ejemplo de
	// contrato evals-y-skill §4 de H7.1.
	trasladoDelCambio = "\xe2\x9a\xa0 REDACCI\xc3\x93N MODIFICADA: la redacci\xc3\xb3n con fecha de vigencia 20161002, " +
		"la que se consult\xc3\xb3 antes, ha sido sustituida por la de 20250101, que es la que se cita."
)

// Lo que juzga TestJuzgarLasExpresionesProhibidas con la lista de
// boe-legislacion (contrato lista-y-juicio §3 y §4 de H7.2;
// contracts/lista-de-expresiones.md §6 de H7.3).
const (
	// transicionDeLaMemoria es la frase de transición de las respuestas de H7.1
	// que cuentan la comprobación (research, «Causa de raíz del ruido»): lleva
	// memoria de consultas y hallazgos, de la maquinaria, y tengo todo lo
	// necesario, del anuncio.
	transicionDeLaMemoria = "Sin hallazgos en la memoria de consultas. Ya tengo todo lo necesario para responder."

	// loDichoEnOtraConversacion atribuye a la skill algo dicho en otra
	// conversación: lleva te confirmé.
	loDichoEnOtraConversacion = "Como te confirmé, el plazo máximo para resolver es de tres meses."

	// elAnuncioDeLaRespuesta anuncia la respuesta que viene: lleva ya puedo
	// responder, del anuncio, y ninguna otra expresión de la lista.
	elAnuncioDeLaRespuesta = "Ya puedo responder."

	// redaccionModificadaDeLaLCSP es la línea con la que la skill traslada
	// version-obsoleta, con sus dos fechas: la del ejemplo del contrato
	// lista-y-juicio §3, que no lleva ninguna expresión.
	redaccionModificadaDeLaLCSP = "⚠ REDACCIÓN MODIFICADA: la redacción con fecha de vigencia 20180309, " +
		"la que se consultó antes, ha sido sustituida por la de 20200206, que es la que se cita."
)

// elementosDelMunicipio son los elementos del territorio esperado de la eval del
// municipio cubierto, con el texto con el que los presentan el informe y los
// motivos, en el orden de la eval.
var elementosDelMunicipio = []string{
	"comunidad: Comunidad de Madrid", "provincia: Madrid", "boletín: BOCM", "boletin_autonomico: configurado",
}

// juicio es una llamada a Juzgar y el resultado que el caso exige de ella. skill
// es la skill con la que se juzga; vacía, la de las sesiones de boe-legislacion.
type juicio struct {
	eval     Eval
	sesion   Sesion
	skill    string
	esperado ResultadoDeEval
}

// TestJuzgar fija la comparación mecánica de una sesión con su eval de
// data-model §10.2: pasa solo si la sesión terminó, la activación coincide y
// están todos los comandos y todas las citas esperados; un comando esperado lo
// satisface solo una invocación con consulta —ni la ayuda ni --describe o
// --dry-run verdaderos— y código 0 del mismo applet: articulo o articulos de esa
// norma con ese bloque entre los pedidos para un bloque, el mismo verbo y la
// misma norma para una consulta de norma, y buscar con cada término como palabra
// para una búsqueda; una cita cuenta solo con la misma norma y el mismo bloque; y
// las invocaciones fuera de lo grabado, las otras fallidas y las llegadas a la red
// se informan sin cambiar si pasa (contrato evals-y-grabaciones §6; FR-072,
// FR-076, SC-009; US4, escenarios 2, 3, 5, 7 y 8).
//
// Desde H5.1, los avisos esperados de la eval 01 se reparten, en el orden de la
// eval y con sus repeticiones, entre los que la respuesta lleva con su forma fija
// —también con las tolerancias de la gramática— y los ausentes, cada uno con su
// motivo detrás de los de las citas; un aviso ausente impide pasar aunque estén el
// comando y la cita, la otra redacción y la negación no llevan la forma, lo que
// sigue a la forma no se lee y la forma de un aviso no esperado no cambia nada. Las
// evals sin avisos dejan las dos listas nulas y el mismo resultado que en H5
// (contrato de formato, juicio e informe §4 y §6 de H5.1; FR-030 a FR-035, SC-003;
// US2, escenarios 3 a 8; research D3).
//
// Desde H6, el comando de territorio lo satisface solo una invocación con consulta
// y código 0 de territorio resolver cuyo único argumento, plegado, es el municipio
// esperado; y los elementos del territorio esperado —comunidad, provincia, cada
// boletín y cada aspecto de cobertura— se reparten, en el orden de la eval, entre
// los que la respuesta declara con su forma fija y los ausentes, cada uno con su
// motivo detrás de los de los avisos: un elemento ausente impide pasar aunque el
// comando esté, y una eval con citas y territorio se juzga por los dos (contrato
// de evals §2 de H6; FR-084; US5).
//
// Desde H7, el comando de comprobación lo satisface solo una invocación con
// consulta y código 0 del mismo applet con el verbo check; y un comando
// prohibido lo ejecuta toda invocación con consulta del mismo applet y el mismo
// verbo, termine con 0, con otro código o sin código, pero no la ayuda,
// --describe ni --dry-run, ni otro verbo del mismo applet ni el mismo verbo de
// otro applet: cada prohibido ejecutado va, una sola vez aunque lo ejecuten
// varias invocaciones, a los comandos prohibidos ejecutados con su motivo detrás
// de los de los comandos ausentes y delante de los de las citas, y la eval no
// pasa; sin prohibidos, el juicio es el de antes aunque la sesión ejecute ese
// comando (contrato evals-y-skill §2 de H7; FR-085, FR-086).
//
// Desde H7.1, los hallazgos esperados se reparten, en el orden de la eval, entre
// los que la respuesta lleva con su forma fija —también con las tolerancias de la
// de los avisos— y los ausentes, cada uno con su motivo detrás de los de los
// avisos: un hallazgo ausente impide pasar aunque estén los comandos y la cita, y
// decir el cambio con otras palabras no lleva la forma. El comando de
// comprobación con norma lo satisface solo la invocación de check que lleva esa
// norma como primer argumento, y el que no la lleva, cualquier check, como en
// H7; una eval sin hallazgos se juzga como antes aunque la respuesta lleve la
// forma (contrato evals-y-skill §2 de H7.1; FR-050, FR-051, FR-052, FR-094;
// SC-008).
//
// Desde H7.3, el resultado lleva los reintentos por rate_limit de la sesión, que
// no cambian su juicio (data-model §4 de H7.3; FR-033, FR-041).
func TestJuzgar(t *testing.T) {
	t.Parallel()

	derogada := []string{"derogada"}

	casos := []struct {
		nombre  string
		juicios []juicio
	}{
		{
			nombre:  "pasa",
			juicios: []juicio{{eval: evalDelArticulo21(), sesion: sesionQuePasa(t), esperado: resultadoQuePasa()}},
		},
		{
			nombre: "positiva-no-activada",
			juicios: []juicio{{
				eval:   evalDelArticulo21(),
				sesion: sesionTerminada(false, respuestaConCita, leeElArticulo21(t)),
				esperado: cambiado(resultadoQuePasa(), func(r *ResultadoDeEval) {
					r.Activada = false
					r.Motivos = []string{
						"la activación no coincide: se esperaba que la skill boe-legislacion se activara y no se activó",
					}
					r.Pasa = false
				}),
			}},
		},
		{
			nombre: "no-activa-pero-activada",
			juicios: []juicio{{
				eval:   evalDeNoActivacion(),
				sesion: sesionTerminada(true, respuestaSinSkill),
				esperado: ResultadoDeEval{
					Eval:             ficheroDeNoActivacion,
					Activada:         true,
					Respuesta:        respuestaSinSkill,
					CodigoDeLaSesion: codigoDeSalida(0),
					FinDeLaSesion:    "result success",
					SesionTerminada:  true,
					Motivos: []string{
						"la activación no coincide: se esperaba que la skill boe-legislacion no se activara y se activó",
					},
				},
			}},
		},
		{
			nombre: "sesion-sin-terminar-no-activa",
			juicios: []juicio{{
				eval: evalDeNoActivacion(),
				sesion: Sesion{
					Modelo:              modeloDeLasSesiones,
					VersionDeClaudeCode: versionDeLasSesiones,
					Codigo:              124,
					Fin:                 "system",
					Cortada:             true,
					MotivoSinTerminar:   "tope de 240 s agotado (código 124)",
				},
				esperado: ResultadoDeEval{
					Eval:             ficheroDeNoActivacion,
					CodigoDeLaSesion: codigoDeSalida(124),
					FinDeLaSesion:    "system",
					Motivos:          []string{"la sesión no terminó: tope de 240 s agotado (código 124)"},
				},
			}},
		},
		{
			// La respuesta con la cita se deja en la sesión, aunque LeerSesion no
			// la da con error_max_turns, para que la única causa sea que la sesión
			// no terminó.
			nombre: "sesion-sin-terminar-positiva",
			juicios: []juicio{{
				eval: evalDelArticulo21(),
				sesion: cambiada(sesionQuePasa(t), func(s *Sesion) {
					s.Fin = "result error_max_turns"
					s.Terminada = false
					s.MotivoSinTerminar = "result con subtype error_max_turns"
				}),
				esperado: cambiado(resultadoQuePasa(), func(r *ResultadoDeEval) {
					r.FinDeLaSesion = "result error_max_turns"
					r.SesionTerminada = false
					r.Motivos = []string{"la sesión no terminó: result con subtype error_max_turns"}
					r.Pasa = false
				}),
			}},
		},
		{
			nombre: "bloque-leido-con-codigo-4",
			juicios: []juicio{{
				eval: evalDelArticulo21(),
				sesion: sesionTerminada(true, respuestaConCita,
					invocada(t, codigoDeSalida(4), deLaSkill("articulo", normaDeLasTrazas, "a21", "--offline", "--json"))),
				esperado: cambiado(resultadoQuePasa(), func(r *ResultadoDeEval) {
					r.ComandosEjecutados = nil
					r.ComandosAusentes = []string{textoDelComando21}
					r.Invocaciones = []InvocacionInformada{
						{Orden: "boe articulo BOE-A-2015-10565 a21 --offline --json", Codigo: codigoDeSalida(4)},
					}
					r.FueraDeLoGrabado = []InvocacionFallida{
						{Orden: "boe articulo BOE-A-2015-10565 a21 --offline --json", Codigo: 4},
					}
					r.Motivos = []string{"comando ausente: " + textoDelComando21}
					r.Pasa = false
				}),
			}},
		},
		{
			nombre: "bloque-leido-sin-codigo",
			juicios: []juicio{{
				eval: evalDelArticulo21(),
				sesion: Sesion{
					Modelo:              modeloDeLasSesiones,
					VersionDeClaudeCode: versionDeLasSesiones,
					SkillsActivadas:     []string{skillDeLasSesiones},
					Codigo:              124,
					Fin:                 "assistant",
					Cortada:             true,
					MotivoSinTerminar:   "tope de 240 s agotado (código 124)",
					Invocaciones:        []Invocacion{invocada(t, nil, deLaSkill("articulo", normaDeLasTrazas, "a21", "--json"))},
				},
				esperado: ResultadoDeEval{
					Eval:             ficheroDeLaEval01,
					Activa:           true,
					Activada:         true,
					ComandosAusentes: []string{textoDelComando21},
					CitasAusentes:    []string{textoDeLaCita21},
					Invocaciones:     []InvocacionInformada{{Orden: ordenDelArticulo21}},
					CodigoDeLaSesion: codigoDeSalida(124),
					FinDeLaSesion:    "assistant",
					Motivos: []string{
						"la sesión no terminó: tope de 240 s agotado (código 124)",
						"comando ausente: " + textoDelComando21,
						"cita ausente: " + textoDeLaCita21,
					},
				},
			}},
		},
		{
			nombre: "articulos-satisface-un-bloque",
			juicios: []juicio{{
				eval: evalDelArticulo21(),
				sesion: sesionTerminada(true, respuestaConCita, invocada(t, codigoDeSalida(0),
					[]string{"/ruta/kitlegal", "boe", "articulos", normaDeLasTrazas, "a20", "a21", "--json"})),
				esperado: cambiado(resultadoQuePasa(), func(r *ResultadoDeEval) {
					r.Invocaciones = []InvocacionInformada{
						{Orden: "boe articulos BOE-A-2015-10565 a20 a21 --json", Codigo: codigoDeSalida(0)},
					}
				}),
			}},
		},
		{
			nombre:  "metadatos-no-satisface-indice",
			juicios: []juicio{metadatosNoSatisfaceIndice(t)},
		},
		{
			// Leer con código 0 otros bloques de la misma norma, con articulo y con
			// articulos, no satisface el bloque esperado aunque la respuesta lo
			// cite (FR-072: «ese identificador y ese bloque»).
			nombre: "otro-bloque-no-satisface",
			juicios: []juicio{sinElBloque21(t,
				"articulo BOE-A-2015-10565 a22 --json", "articulos BOE-A-2015-10565 a20 a22 --json")},
		},
		{
			// Otro verbo con la norma y el bloque entre sus argumentos, como una
			// búsqueda de esos dos términos, no lee el bloque.
			nombre:  "otro-verbo-con-el-bloque-no-satisface",
			juicios: []juicio{sinElBloque21(t, "buscar BOE-A-2015-10565 a21 --json")},
		},
		{
			// Con --help o -h el binario imprime la ayuda y termina con 0 sin leer
			// nada: la invocación no tiene consulta (data-model §9).
			nombre: "ayuda-no-satisface",
			juicios: []juicio{sinElBloque21(t,
				"articulo BOE-A-2015-10565 a21 --help", "articulo BOE-A-2015-10565 a21 -h")},
		},
		{
			nombre:  "buscar-verbo-y-terminos-como-palabras",
			juicios: []juicio{terminosComoPalabras(t)},
		},
		{
			nombre:  "cita-de-otro-bloque",
			juicios: []juicio{citaQueNoCuenta(t, "[BOE-A-2015-10565, bloque a22]")},
		},
		{
			nombre:  "cita-de-otra-norma",
			juicios: []juicio{citaQueNoCuenta(t, "[BOE-A-2017-12902, bloque a21]")},
		},
		{
			nombre: "fuera-de-lo-grabado-con-y-sin-offline",
			juicios: []juicio{{
				eval: evalDelArticulo21(),
				sesion: sesionQuePasa(t,
					invocada(t, codigoDeSalida(5), deLaSkill("articulo", normaDeLasTrazas, "a9998", "--json")),
					invocada(t, codigoDeSalida(4), deLaSkill("articulo", normaDeLasTrazas, "a9998", "--offline", "--json"))),
				esperado: cambiado(resultadoQuePasa(
					InvocacionInformada{Orden: "boe articulo BOE-A-2015-10565 a9998 --json", Codigo: codigoDeSalida(5)},
					InvocacionInformada{Orden: "boe articulo BOE-A-2015-10565 a9998 --offline --json", Codigo: codigoDeSalida(4)},
				), func(r *ResultadoDeEval) {
					r.FueraDeLoGrabado = []InvocacionFallida{
						{Orden: "boe articulo BOE-A-2015-10565 a9998 --json", Codigo: 5},
						{Orden: "boe articulo BOE-A-2015-10565 a9998 --offline --json", Codigo: 4},
					}
				}),
			}},
		},
		{
			nombre:  "red-no-cambia-la-eval",
			juicios: []juicio{redNoCambiaLaEval(t)},
		},
		{
			nombre: "describe-y-dry-run-no-satisfacen",
			juicios: []juicio{{
				eval: evalDelArticulo21(),
				sesion: sesionTerminada(true, respuestaConCita,
					invocada(t, codigoDeSalida(0), deLaSkill("articulo", normaDeLasTrazas, "a21", "--describe")),
					invocada(t, codigoDeSalida(0), deLaSkill("articulo", normaDeLasTrazas, "a21", "--dry-run")),
					invocada(t, codigoDeSalida(2), deLaSkill("articulo", "BOE-A-15-1", "a21", "--dry-run", "--json"))),
				esperado: cambiado(resultadoQuePasa(), func(r *ResultadoDeEval) {
					r.ComandosEjecutados = nil
					r.ComandosAusentes = []string{textoDelComando21}
					r.Invocaciones = []InvocacionInformada{
						{Orden: "boe articulo BOE-A-2015-10565 a21 --describe", Codigo: codigoDeSalida(0)},
						{Orden: "boe articulo BOE-A-2015-10565 a21 --dry-run", Codigo: codigoDeSalida(0)},
						{Orden: "boe articulo BOE-A-15-1 a21 --dry-run --json", Codigo: codigoDeSalida(2)},
					}
					r.Motivos = []string{"comando ausente: " + textoDelComando21}
					r.Pasa = false
				}),
			}},
		},
		{
			// Con --describe=false o --dry-run=false el binario sí consulta
			// (data-model §9).
			nombre:  "describe-y-dry-run-falsos-consultan",
			juicios: []juicio{describeYDryRunFalsos(t)},
		},
		{
			nombre: "otra-fallida",
			juicios: []juicio{{
				eval: evalDelArticulo21(),
				sesion: sesionQuePasa(t,
					invocada(t, codigoDeSalida(2), deLaSkill("articulo", "BOE-A-15-1", "a21", "--json")),
					invocada(t, codigoDeSalida(3), deLaSkill("articulo", normaDeLasTrazas, "a9999", "--json"))),
				esperado: cambiado(resultadoQuePasa(
					InvocacionInformada{Orden: "boe articulo BOE-A-15-1 a21 --json", Codigo: codigoDeSalida(2)},
					InvocacionInformada{Orden: "boe articulo BOE-A-2015-10565 a9999 --json", Codigo: codigoDeSalida(3)},
				), func(r *ResultadoDeEval) {
					r.OtrasFallidas = []InvocacionFallida{
						{Orden: "boe articulo BOE-A-15-1 a21 --json", Codigo: 2},
						{Orden: "boe articulo BOE-A-2015-10565 a9999 --json", Codigo: 3},
					}
				}),
			}},
		},
		{
			nombre:  "sc-009-otro-bloque",
			juicios: mismaSesionConOtraCita(t, CitaEsperada{Norma: normaDeLasTrazas, Bloque: "a22"}, "BOE-A-2015-10565 a22"),
		},
		{
			nombre:  "sc-009-otra-norma",
			juicios: mismaSesionConOtraCita(t, CitaEsperada{Norma: normaDeLaLCSP, Bloque: "a21"}, "BOE-A-2017-12902 a21"),
		},
		{
			nombre: "aviso-con-su-forma-fija",
			juicios: []juicio{conAvisos(t, derogada, avisoDeDerogada+"\n\n"+respuestaConCita, func(r *ResultadoDeEval) {
				r.AvisosEncontrados = derogada
			})},
		},
		{
			// Un juicio por cada forma tolerada: con el selector de presentación
			// U+FE0F, con el énfasis envolviendo la forma, con el énfasis en la
			// etiqueta, con espacios de más y U+00A0 entre las palabras, y en
			// minúsculas.
			nombre: "aviso-con-variantes-toleradas",
			juicios: formasToleradas(t,
				"⚠️ NORMA DEROGADA: esta norma ha sido derogada.",
				"**⚠ NORMA DEROGADA:** esta norma ha sido derogada.",
				"⚠ **NORMA DEROGADA**: esta norma ha sido derogada.",
				"⚠  NORMA DEROGADA : esta norma ha sido derogada.",
				"⚠ norma derogada: esta norma ha sido derogada."),
		},
		{
			nombre: "aviso-ausente",
			juicios: []juicio{conAvisos(t, []string{"derogada", "vigencia-agotada"}, respuestaConCita,
				func(r *ResultadoDeEval) {
					r.AvisosAusentes = []string{"derogada", "vigencia-agotada"}
					r.Motivos = []string{"aviso ausente: derogada", "aviso ausente: vigencia-agotada"}
					r.Pasa = false
				})},
		},
		{
			nombre:  "aviso-con-otra-redaccion",
			juicios: []juicio{derogadaAusente(t, "Esta ley fue derogada.\n\n"+respuestaConCita)},
		},
		{
			// El comando y la cita están: solo falta el aviso.
			nombre:  "aviso-negado",
			juicios: []juicio{derogadaAusente(t, "La Ley 30/1992 sigue en vigor.\n\n"+respuestaConCita)},
		},
		{
			// La limitación declarada: lo que sigue a la forma no se lee.
			nombre: "forma-fija-y-lo-contrario",
			juicios: []juicio{conAvisos(t, derogada, "⚠ NORMA DEROGADA: pero sigue en vigor.\n\n"+respuestaConCita,
				func(r *ResultadoDeEval) {
					r.AvisosEncontrados = derogada
				})},
		},
		{
			// Con y sin la forma de vigencia-agotada, que la eval no espera, el
			// resultado es el mismo salvo la respuesta.
			nombre: "aviso-no-esperado",
			juicios: []juicio{
				conAvisos(t, derogada, avisoDeDerogada+"\n\n"+respuestaConCita, func(r *ResultadoDeEval) {
					r.AvisosEncontrados = derogada
				}),
				conAvisos(t, derogada, avisoDeDerogada+"\n"+avisoDeVigenciaAgotada+"\n\n"+respuestaConCita,
					func(r *ResultadoDeEval) {
						r.AvisosEncontrados = derogada
					}),
			},
		},
		{
			// La eval los espera en el orden contrario al de la respuesta y al de
			// boe.CodigosDeAviso.
			nombre: "avisos-en-el-orden-de-la-eval",
			juicios: []juicio{conAvisos(t, []string{"vigencia-agotada", "derogada"},
				avisoDeDerogada+"\n"+avisoDeVigenciaAgotada+"\n\n"+respuestaConCita, func(r *ResultadoDeEval) {
					r.AvisosEncontrados = []string{"vigencia-agotada", "derogada"}
				})},
		},
		{
			// Repetido en la eval, se reparte repetido: dos veces encontrado con la
			// forma y dos veces ausente, con dos motivos, sin ella.
			nombre: "aviso-repetido",
			juicios: []juicio{
				conAvisos(t, []string{"derogada", "derogada"}, avisoDeDerogada+"\n\n"+respuestaConCita,
					func(r *ResultadoDeEval) {
						r.AvisosEncontrados = []string{"derogada", "derogada"}
					}),
				conAvisos(t, []string{"derogada", "derogada"}, respuestaConCita, func(r *ResultadoDeEval) {
					r.AvisosAusentes = []string{"derogada", "derogada"}
					r.Motivos = []string{"aviso ausente: derogada", "aviso ausente: derogada"}
					r.Pasa = false
				}),
			},
		},
		{
			nombre: "cita-y-aviso-ausentes",
			juicios: []juicio{conAvisos(t, derogada, "El artículo 21 de la Ley 39/2015 regula la obligación de resolver.",
				func(r *ResultadoDeEval) {
					r.CitasEncontradas = nil
					r.CitasAusentes = []string{textoDeLaCita21}
					r.AvisosAusentes = derogada
					r.Motivos = []string{"cita ausente: " + textoDeLaCita21, "aviso ausente: derogada"}
					r.Pasa = false
				})},
		},
		{
			nombre:  "territorio-satisface",
			juicios: territorioSatisface(t),
		},
		{
			nombre:  "territorio-otro-municipio-no-satisface",
			juicios: []juicio{territorioOtroMunicipio(t)},
		},
		{
			nombre:  "territorio-ausente",
			juicios: territorioAusente(t),
		},
		{
			nombre:  "territorio-y-citas",
			juicios: territorioYCitas(t),
		},
		{
			nombre:  "comprobacion-satisfecha-y-prohibido-sin-ejecutar",
			juicios: []juicio{consultaRepetidaQuePasa(t)},
		},
		{
			nombre:  "comprobacion-sin-consulta-con-otro-codigo-o-con-otro-verbo",
			juicios: []juicio{comprobacionAusente(t)},
		},
		{
			nombre:  "prohibido-ejecutado-con-y-sin-codigo-0",
			juicios: prohibidoEjecutado(t),
		},
		{
			nombre:  "prohibido-con-ayuda-describe-o-dry-run-no-cuenta",
			juicios: []juicio{prohibidoSinConsulta(t)},
		},
		{
			nombre:  "otro-verbo-u-otro-applet-no-es-el-prohibido",
			juicios: []juicio{otroQueElProhibido(t)},
		},
		{
			nombre:  "prohibido-entre-comandos-y-citas-ausentes",
			juicios: []juicio{prohibidoEntreAusentes(t)},
		},
		{
			// La eval 01 no prohíbe nada: graph show no cambia su juicio.
			nombre: "sin-prohibidos-el-juicio-de-antes",
			juicios: []juicio{{
				eval:     evalDelArticulo21(),
				sesion:   sesionQuePasa(t, pideLaFicha(t, codigoDeSalida(0))),
				esperado: resultadoQuePasa(InvocacionInformada{Orden: ordenDeGraphShow, Codigo: codigoDeSalida(0)}),
			}},
		},
		{
			nombre: "hallazgo-con-su-forma-fija",
			juicios: []juicio{conHallazgo(t, trasladoDelCambio+"\n\n"+respuestaConLaCita118, func(r *ResultadoDeEval) {
				r.HallazgosEncontrados = []string{versionObsoleta}
			})},
		},
		{
			// Un juicio por cada forma tolerada: la de contrato evals-y-skill §2 de
			// H7.1, con el selector de presentación U+FE0F y el énfasis envolviendo
			// la etiqueta en minúsculas; el selector U+FE0E; y espacios de más y
			// U+00A0 entre las palabras.
			nombre: "hallazgo-con-variantes-toleradas",
			juicios: hallazgoTolerado(t,
				"**\xe2\x9a\xa0\xef\xb8\x8f Redacci\xc3\xb3n modificada**: la redacci\xc3\xb3n ha sido sustituida.",
				"\xe2\x9a\xa0\xef\xb8\x8e REDACCI\xc3\x93N MODIFICADA: la redacci\xc3\xb3n ha sido sustituida.",
				"\xe2\x9a\xa0  REDACCI\xc3\x93N\xc2\xa0MODIFICADA : la redacci\xc3\xb3n ha sido sustituida."),
		},
		{
			nombre:  "hallazgo-ausente",
			juicios: []juicio{conHallazgo(t, respuestaConLaCita118, versionObsoletaAusente)},
		},
		{
			// El comando y la cita están, y la respuesta dice el cambio sin la
			// forma fija: el hallazgo queda ausente.
			nombre: "hallazgo-dicho-con-otras-palabras",
			juicios: []juicio{conHallazgo(t,
				"La redacci\xc3\xb3n ha cambiado desde la consulta anterior.\n\n"+respuestaConLaCita118, versionObsoletaAusente)},
		},
		{
			nombre:  "hallazgo-ausente-detras-de-la-cita-y-del-aviso",
			juicios: []juicio{hallazgoDetrasDeLosAusentes(t)},
		},
		{
			nombre:  "comprobacion-con-otra-norma-no-satisface",
			juicios: comprobacionConOtraNorma(t),
		},
		{
			// Una sesión que se recupera de sus reintentos se juzga como cualquier
			// otra y publica los de rate_limit, no los de sobrecarga (FR-033 y
			// FR-041 de H7.3). Juzgar no la clasifica: eso lo hace EscribirInforme.
			nombre: "con-reintentos-por-limite-de-ritmo",
			juicios: []juicio{{
				eval: evalDelArticulo21(),
				sesion: cambiada(sesionQuePasa(t), func(s *Sesion) {
					s.Reintentos = []ReintentoDeLaAPI{
						{Intento: 1, Maximo: 10, Error: "rate_limit"},
						{Intento: 2, Maximo: 10, Error: "overloaded"},
						{Intento: 3, Maximo: 10, Error: "rate_limit"},
					}
				}),
				esperado: cambiado(resultadoQuePasa(), func(r *ResultadoDeEval) { r.ReintentosPorLimiteDeRitmo = 2 }),
			}},
		},
		{
			// La eval 01 no espera hallazgos: la forma en la respuesta no cambia su
			// juicio.
			nombre: "sin-hallazgos-el-juicio-de-antes",
			juicios: []juicio{{
				eval:   evalDelArticulo21(),
				sesion: cambiada(sesionQuePasa(t), func(s *Sesion) { s.Respuesta = trasladoDelCambio + "\n\n" + respuestaConCita }),
				esperado: cambiado(resultadoQuePasa(), func(r *ResultadoDeEval) {
					r.Respuesta = trasladoDelCambio + "\n\n" + respuestaConCita
				}),
			}},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			for _, j := range caso.juicios {
				assert.Equal(t, j.esperado, Juzgar(j.eval, j.sesion, cmp.Or(j.skill, skillDeLasSesiones)),
					"la eval %s con la sesión del caso", j.eval.Fichero)
			}
		})
	}
}

// TestJuzgarLasEvalsSinHallazgos fija que las evals del repositorio de
// boe-legislacion y de legal-core que no esperan hallazgos se juzgan igual que
// antes de H7.1 (contrato evals-y-skill §1 de H7.1; FR-054): con una sesión
// terminada, activada si la eval espera que se active, cuya respuesta lleva la
// cita del art. 21 de la Ley 39/2015, el resultado es el mismo con la forma fija
// de version-obsoleta delante de la respuesta que sin ella, salvo la respuesta, y
// no reparte ningún hallazgo.
func TestJuzgarLasEvalsSinHallazgos(t *testing.T) {
	t.Parallel()

	for dir, skill := range map[string]string{evalsDelRepositorio: skillDeLasSesiones, evalsDeLegalCore: skillDeTerritorio} {
		conjunto, err := LeerConjunto(dir)
		require.NoError(t, err)
		require.NotEmpty(t, conjunto.Evals, "%s tiene evals", dir)

		for _, eval := range conjunto.Evals {
			if len(eval.Hallazgos) > 0 {
				continue
			}

			sesion := func(respuesta string) Sesion {
				return cambiada(sesionTerminada(false, respuesta), func(s *Sesion) {
					if eval.Activa {
						s.SkillsActivadas = []string{skill}
					}
				})
			}

			sinLaForma := Juzgar(eval, sesion(respuestaConCita), skill)
			conLaForma := Juzgar(eval, sesion(trasladoDelCambio+"\n\n"+respuestaConCita), skill)

			assert.Nil(t, conLaForma.HallazgosEncontrados, "%s no espera hallazgos y no encuentra ninguno", eval.Fichero)
			assert.Nil(t, conLaForma.HallazgosAusentes, "%s no espera hallazgos y no le falta ninguno", eval.Fichero)

			conLaForma.Respuesta = sinLaForma.Respuesta
			assert.Equal(t, sinLaForma, conLaForma, "%s se juzga igual con la forma fija de un hallazgo en la respuesta",
				eval.Fichero)
		}
	}
}

// TestHallazgosDelResultadoEnJSON fija que el resultado de una sesión escribe
// hallazgos_encontrados y hallazgos_ausentes como listas, nunca null, codificado
// como lo codifica EscribirInforme (contrato evals-y-skill §6 de H7.1; FR-055):
// vacías con una eval que no espera hallazgos, y con version-obsoleta en la que
// corresponde con la eval de la redacción cambiada.
func TestHallazgosDelResultadoEnJSON(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre      string
		juicio      juicio
		encontrados string
		ausentes    string
	}{
		{
			nombre:      "sin-hallazgos",
			juicio:      juicio{eval: evalDelArticulo21(), sesion: sesionQuePasa(t)},
			encontrados: "[]",
			ausentes:    "[]",
		},
		{
			nombre:      "encontrado",
			juicio:      conHallazgo(t, trasladoDelCambio+"\n\n"+respuestaConLaCita118, func(*ResultadoDeEval) {}),
			encontrados: `["version-obsoleta"]`,
			ausentes:    "[]",
		},
		{
			nombre:      "ausente",
			juicio:      conHallazgo(t, respuestaConLaCita118, func(*ResultadoDeEval) {}),
			encontrados: "[]",
			ausentes:    `["version-obsoleta"]`,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			codificado, err := json.Marshal(Juzgar(caso.juicio.eval, caso.juicio.sesion, skillDeLasSesiones))
			require.NoError(t, err)

			var crudo struct {
				Encontrados jsontext.Value `json:"hallazgos_encontrados"`
				Ausentes    jsontext.Value `json:"hallazgos_ausentes"`
			}

			require.NoError(t, json.Unmarshal(codificado, &crudo))
			assert.Equal(t, caso.encontrados, string(crudo.Encontrados))
			assert.Equal(t, caso.ausentes, string(crudo.Ausentes))
		})
	}
}

// TestJuzgarLasExpresionesProhibidas fija el juicio con la lista de expresiones
// prohibidas de boe-legislacion, la del repositorio leída con LeerConjunto
// (contrato lista-y-juicio §4 de H7.2; FR-051, FR-052, FR-054, FR-083; SC-006;
// US3.1 a US3.6; y contracts/lista-de-expresiones.md §6 de H7.3, FR-024 y
// US1-3): con la eval 01 del repositorio y una sesión que cumple todo lo demás,
// cada expresión que lleva la respuesta, de cualquiera de las tres familias, en
// el orden de la lista y con las tolerancias de la forma fija de los avisos, va
// a las expresiones prohibidas con su motivo y la eval no pasa; sin ninguna, o
// con la línea
// ⚠ REDACCIÓN MODIFICADA: con sus dos fechas, pasa y la lista queda vacía. Los
// motivos van detrás de los de Juzgar y delante del del modelo; una eval de no
// activación y una de una skill sin lista se juzgan como antes del hito aunque
// la respuesta las lleve; y el resultado escribe la clave detrás de
// territorio_ausente, una lista vacía si no hay ninguna.
func TestJuzgarLasExpresionesProhibidas(t *testing.T) {
	t.Parallel()

	conjunto, err := LeerConjunto(evalsDelRepositorio)
	require.NoError(t, err)
	require.NotEmpty(t, slices.Concat(conjunto.Prohibidas.Maquinaria, conjunto.Prohibidas.OtraConversacion,
		conjunto.Prohibidas.Anuncio), "%s tiene su lista de expresiones prohibidas", evalsDelRepositorio)

	positiva := *evalDe(t, conjunto.Evals, ficheroDeLaEval01)
	require.True(t, positiva.Activa, "%s espera que la skill se active", ficheroDeLaEval01)
	require.Equal(t, conjunto.Prohibidas, positiva.Prohibidas, "%s lleva la lista de su carpeta", ficheroDeLaEval01)

	t.Run("positiva", func(t *testing.T) {
		t.Parallel()
		juzgarLaPositivaConLaLista(t, positiva)
	})

	t.Run("detras-de-los-demas-motivos", func(t *testing.T) {
		t.Parallel()
		juzgarLosMotivosEnSuOrden(t, conjunto.Prohibidas)
	})

	t.Run("como-antes-del-hito", func(t *testing.T) {
		t.Parallel()
		juzgarComoAntesDelHito(t, conjunto)
	})

	t.Run("json", func(t *testing.T) {
		t.Parallel()
		codificarLasExpresionesDelResultado(t, positiva)
	})
}

// juzgarLaPositivaConLaLista juzga la eval positiva con la lista y la sesión que
// la pasa, con cada respuesta de la tabla delante de la cita: el resultado es el
// de la sesión que pasa con las expresiones encontradas y sus motivos, y pasa solo
// si no hay ninguna.
func juzgarLaPositivaConLaLista(t *testing.T, positiva Eval) {
	t.Helper()

	encontradasEnLaTransicion := []string{"memoria de consultas", "hallazgos", "tengo todo lo necesario"}
	motivosDeLaTransicion := []string{
		"expresión prohibida: memoria de consultas",
		"expresión prohibida: hallazgos",
		"expresión prohibida: tengo todo lo necesario",
	}

	casos := []struct {
		nombre      string
		antes       string
		encontradas []string
		motivos     []string
	}{
		{nombre: "sin-expresiones"},
		{
			nombre:      "maquinaria",
			antes:       transicionDeLaMemoria,
			encontradas: encontradasEnLaTransicion,
			motivos:     motivosDeLaTransicion,
		},
		{
			nombre:      "otra-conversacion",
			antes:       loDichoEnOtraConversacion,
			encontradas: []string{"te confirmé"},
			motivos:     []string{"expresión prohibida: te confirmé"},
		},
		{
			nombre:      "anuncio",
			antes:       elAnuncioDeLaRespuesta,
			encontradas: []string{"ya puedo responder"},
			motivos:     []string{"expresión prohibida: ya puedo responder"},
		},
		{
			nombre:      "otras-mayusculas",
			antes:       "SIN HALLAZGOS EN LA MEMORIA DE CONSULTAS. Ya tengo todo lo necesario para responder.",
			encontradas: encontradasEnLaTransicion,
			motivos:     motivosDeLaTransicion,
		},
		{
			// Dos espacios, tres y un espacio sin separación U+00A0.
			nombre:      "espacios-de-mas",
			antes:       "Sin  hallazgos en la memoria   de\xc2\xa0consultas. Ya tengo todo lo necesario para responder.",
			encontradas: encontradasEnLaTransicion,
			motivos:     motivosDeLaTransicion,
		},
		{
			nombre:      "enfasis-alrededor",
			antes:       "Sin hallazgos en la **memoria de consultas**. Ya tengo todo lo necesario para responder.",
			encontradas: encontradasEnLaTransicion,
			motivos:     motivosDeLaTransicion,
		},
		{
			nombre:      "enfasis-entre-las-palabras",
			antes:       "Sin _hallazgos_ en la *memoria* de _consultas_. Ya tengo todo lo necesario para responder.",
			encontradas: encontradasEnLaTransicion,
			motivos:     motivosDeLaTransicion,
		},
		{nombre: "redaccion-modificada", antes: redaccionModificadaDeLaLCSP},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			respuesta := respuestaConCita
			if caso.antes != "" {
				respuesta = caso.antes + "\n\n" + respuestaConCita
			}

			esperado := cambiado(resultadoQuePasa(), func(r *ResultadoDeEval) {
				r.Respuesta = respuesta
				r.ExpresionesProhibidas = caso.encontradas
				r.Motivos = caso.motivos
				r.Pasa = len(caso.encontradas) == 0
			})

			sesion := cambiada(sesionQuePasa(t), func(s *Sesion) { s.Respuesta = respuesta })

			assert.Equal(t, esperado, Juzgar(positiva, sesion, skillDeLasSesiones), "%s con la respuesta del caso",
				positiva.Fichero)
		})
	}
}

// juzgarLosMotivosEnSuOrden juzga la eval del municipio cubierto con la lista
// dada y una sesión que resuelve el municipio sin declarar su territorio y
// responde con la transición de la memoria: los motivos de las expresiones van
// detrás de los del territorio ausente, que son los últimos de antes del hito, y
// delante del del modelo que la sesión declara sin ser el pedido.
func juzgarLosMotivosEnSuOrden(t *testing.T, lista ExpresionesProhibidas) {
	t.Helper()

	eval := evalDelMunicipioCubierto()
	sesion := sesionDeTerritorio(transicionDeLaMemoria, resuelveLeganes(t))

	antes := Juzgar(eval, sesion, skillDeTerritorio)
	require.Equal(t, prefijados(motivoDeTerritorioAusente, elementosDelMunicipio), antes.Motivos,
		"sin la lista, los únicos motivos son los del territorio ausente")

	eval.Prohibidas = lista
	resultado := Juzgar(eval, sesion, skillDeTerritorio)
	resultado.Modelo, resultado.ModeloDeLaSesion = modeloInformativoDelCaso, modeloDeLasSesiones
	resultado.exigirElModeloPedido()

	assert.Equal(t, []string{"memoria de consultas", "hallazgos", "tengo todo lo necesario"},
		resultado.ExpresionesProhibidas)
	assert.Equal(t, slices.Concat(antes.Motivos, []string{
		"expresión prohibida: memoria de consultas",
		"expresión prohibida: hallazgos",
		"expresión prohibida: tengo todo lo necesario",
		motivoDeOtroModelo + modeloDeLasSesiones + ", y se pidió " + modeloInformativoDelCaso,
	}), resultado.Motivos)
	assert.False(t, resultado.Pasa)
}

// juzgarComoAntesDelHito juzga la eval de no activación del conjunto, que lleva
// la lista de su carpeta, y la del municipio cubierto de legal-core, cuya carpeta
// no tiene lista, con una sesión que cumple lo que esperan y con la misma sesión
// con expresiones de las tres familias detrás de la respuesta: el resultado es el
// mismo salvo la respuesta, sin ninguna expresión, y pasa.
func juzgarComoAntesDelHito(t *testing.T, conjunto Conjunto) {
	t.Helper()

	noActivacion := *evalDe(t, conjunto.Evals, ficheroDeNoActivacion)
	require.False(t, noActivacion.Activa, "%s no espera que la skill se active", ficheroDeNoActivacion)
	require.Equal(t, conjunto.Prohibidas, noActivacion.Prohibidas, "%s lleva la lista de su carpeta",
		ficheroDeNoActivacion)

	deLegalCore, err := LeerConjunto(evalsDeLegalCore)
	require.NoError(t, err)
	require.Zero(t, deLegalCore.Prohibidas, "%s no tiene lista de expresiones prohibidas", evalsDeLegalCore)

	municipio := *evalDe(t, deLegalCore.Evals, ficheroDelMunicipioCubierto)
	require.True(t, municipio.Activa, "%s espera que la skill se active", ficheroDelMunicipioCubierto)

	casos := []struct {
		eval      Eval
		skill     string
		respuesta string
		sesion    func(respuesta string) Sesion
	}{
		{
			eval:      noActivacion,
			skill:     skillDeLasSesiones,
			respuesta: respuestaSinSkill,
			sesion:    func(respuesta string) Sesion { return sesionTerminada(false, respuesta) },
		},
		{
			eval:      municipio,
			skill:     skillDeTerritorio,
			respuesta: respuestaDelMunicipio,
			sesion: func(respuesta string) Sesion {
				return sesionDeTerritorio(respuesta, resuelveLeganes(t))
			},
		},
	}

	for _, caso := range casos {
		conExpresiones := caso.respuesta + "\n\n" + transicionDeLaMemoria + " " + loDichoEnOtraConversacion
		require.NotEmpty(t, ExtraerExpresionesProhibidas(conExpresiones, conjunto.Prohibidas),
			"la respuesta del caso de %s lleva expresiones de la lista", caso.eval.Fichero)

		antes := Juzgar(caso.eval, caso.sesion(caso.respuesta), caso.skill)
		require.True(t, antes.Pasa, "la sesión del caso de %s cumple lo que la eval espera", caso.eval.Fichero)

		ahora := Juzgar(caso.eval, caso.sesion(conExpresiones), caso.skill)
		assert.Nil(t, ahora.ExpresionesProhibidas, "%s no encuentra ninguna expresión", caso.eval.Fichero)

		ahora.Respuesta = antes.Respuesta
		assert.Equal(t, antes, ahora, "%s se juzga igual con expresiones de la lista en la respuesta", caso.eval.Fichero)
	}
}

// codificarLasExpresionesDelResultado codifica, como lo codifica
// EscribirInforme, el resultado de la eval positiva con la sesión que la pasa, sin
// expresiones y con la transición de la memoria: expresiones_prohibidas va detrás
// de territorio_ausente, una lista vacía, nunca null, si no hay ninguna.
func codificarLasExpresionesDelResultado(t *testing.T, positiva Eval) {
	t.Helper()

	casos := []struct {
		respuesta string
		clave     string
	}{
		{respuesta: respuestaConCita, clave: `"territorio_ausente":[],"expresiones_prohibidas":[],`},
		{
			respuesta: transicionDeLaMemoria + "\n\n" + respuestaConCita,
			clave: `"territorio_ausente":[],` +
				`"expresiones_prohibidas":["memoria de consultas","hallazgos","tengo todo lo necesario"],`,
		},
	}

	for _, caso := range casos {
		sesion := cambiada(sesionQuePasa(t), func(s *Sesion) { s.Respuesta = caso.respuesta })

		codificado, err := json.Marshal(Juzgar(positiva, sesion, skillDeLasSesiones))
		require.NoError(t, err)
		assert.Contains(t, string(codificado), caso.clave)
	}
}

// prefijados son los textos con el principio dado delante de cada uno.
func prefijados(principio string, textos []string) []string {
	con := make([]string, 0, len(textos))
	for _, texto := range textos {
		con = append(con, principio+texto)
	}

	return con
}

// territorioSatisface son los juicios de la eval del municipio cubierto con una
// sesión que lo resuelve y declara todo su territorio, uno por cada forma de
// escribir el municipio que, plegada, es la suya: con sus tildes, sin ellas, en
// mayúsculas y por kitlegal territorio. En todos, el comando queda ejecutado con el
// texto de la eval, los cuatro elementos encontrados y la eval pasa.
func territorioSatisface(t *testing.T) []juicio {
	t.Helper()

	invocaciones := []struct {
		argv  []string
		orden string
	}{
		{argv: deLegalCore("resolver", "Leganés", "--json"), orden: ordenDeLeganes},
		{argv: deLegalCore("resolver", "leganes", "--json"), orden: "territorio resolver leganes --json"},
		{argv: deLegalCore("resolver", "LEGANÉS", "--json"), orden: "territorio resolver LEGANÉS --json"},
		{argv: []string{"/ruta/kitlegal", "territorio", "resolver", "Leganés", "--json"}, orden: ordenDeLeganes},
	}

	juicios := make([]juicio, 0, len(invocaciones))

	for _, invocacion := range invocaciones {
		juicios = append(juicios, juicio{
			eval:   evalDelMunicipioCubierto(),
			sesion: sesionDeTerritorio(respuestaDelMunicipio, invocada(t, codigoDeSalida(0), invocacion.argv)),
			skill:  skillDeTerritorio,
			esperado: cambiado(resultadoDelMunicipioQuePasa(), func(r *ResultadoDeEval) {
				r.Invocaciones = []InvocacionInformada{{Orden: invocacion.orden, Codigo: codigoDeSalida(0)}}
			}),
		})
	}

	return juicios
}

// territorioOtroMunicipio es el juicio de la eval del municipio cubierto con una
// sesión que declara todo su territorio sin resolverlo: resuelve otro municipio,
// resuelve el suyo con --describe, que no consulta, y lo busca con otro applet,
// todas con código 0. El comando queda ausente y la eval no pasa, con los cuatro
// elementos encontrados.
func territorioOtroMunicipio(t *testing.T) juicio {
	t.Helper()

	return juicio{
		eval: evalDelMunicipioCubierto(),
		sesion: sesionDeTerritorio(respuestaDelMunicipio,
			invocada(t, codigoDeSalida(0), deLegalCore("resolver", "Getafe", "--json")),
			invocada(t, codigoDeSalida(0), deLegalCore("resolver", "Leganés", "--describe")),
			invocada(t, codigoDeSalida(0), deLaSkill("buscar", "Leganés", "--json"))),
		skill: skillDeTerritorio,
		esperado: cambiado(resultadoDelMunicipioQuePasa(), func(r *ResultadoDeEval) {
			r.ComandosEjecutados = nil
			r.ComandosAusentes = []string{textoDelComandoDeLeganes}
			r.Invocaciones = []InvocacionInformada{
				{Orden: "territorio resolver Getafe --json", Codigo: codigoDeSalida(0)},
				{Orden: "territorio resolver Leganés --describe", Codigo: codigoDeSalida(0)},
				{Orden: "boe buscar Leganés --json", Codigo: codigoDeSalida(0)},
			}
			r.Motivos = []string{"comando ausente: " + textoDelComandoDeLeganes}
			r.Pasa = false
		}),
	}
}

// territorioAusente son los juicios de la eval del municipio cubierto con una
// sesión que lo resuelve y cuya respuesta no declara parte de su territorio: sin
// nada de él, los cuatro elementos quedan ausentes, cada uno con su motivo; y con
// la comunidad —que lleva el nombre de la provincia— y el boletín, pero con el
// otro valor del aspecto, solo queda ausente el aspecto. En los dos, el comando
// queda ejecutado y la eval no pasa.
func territorioAusente(t *testing.T) []juicio {
	t.Helper()

	sinNada := "Leganés publica sus normas en el BOE."
	conOtroValor := "Leganés está en la Comunidad de Madrid y publica en el BOCM.\n\nboletin_autonomico: no-configurado"

	return []juicio{
		{
			eval:   evalDelMunicipioCubierto(),
			sesion: sesionDeTerritorio(sinNada, resuelveLeganes(t)),
			skill:  skillDeTerritorio,
			esperado: cambiado(resultadoDelMunicipioQuePasa(), func(r *ResultadoDeEval) {
				r.TerritorioEncontrado = nil
				r.TerritorioAusente = elementosDelMunicipio
				r.Respuesta = sinNada
				r.Motivos = []string{
					"territorio ausente: comunidad: Comunidad de Madrid",
					"territorio ausente: provincia: Madrid",
					"territorio ausente: boletín: BOCM",
					"territorio ausente: boletin_autonomico: configurado",
				}
				r.Pasa = false
			}),
		},
		{
			eval:   evalDelMunicipioCubierto(),
			sesion: sesionDeTerritorio(conOtroValor, resuelveLeganes(t)),
			skill:  skillDeTerritorio,
			esperado: cambiado(resultadoDelMunicipioQuePasa(), func(r *ResultadoDeEval) {
				r.TerritorioEncontrado = elementosDelMunicipio[:3]
				r.TerritorioAusente = elementosDelMunicipio[3:]
				r.Respuesta = conOtroValor
				r.Motivos = []string{"territorio ausente: boletin_autonomico: configurado"}
				r.Pasa = false
			}),
		},
	}
}

// territorioYCitas son los juicios de una eval que espera el territorio y una
// cita, con una sesión que resuelve el municipio y lee el bloque: con los dos en
// la respuesta, pasa; sin la cita, falta solo la cita; sin el boletín, falta solo
// el boletín; y sin los dos, faltan los dos, con el motivo de la cita delante del
// del territorio.
func territorioYCitas(t *testing.T) []juicio {
	t.Helper()

	eval := Eval{
		Fichero:  "04-territorio-y-articulo-21.yaml",
		Pregunta: "¿Qué dice el art. 21 de la Ley 39/2015 y en qué boletines publica Leganés?",
		Activa:   true,
		Comandos: []ComandoEsperado{
			{Applet: "territorio", Verbo: "resolver", Municipio: "Leganés"},
			{Applet: "boe", Norma: normaDeLasTrazas, Bloque: "a21"},
		},
		Citas:      []CitaEsperada{{Norma: normaDeLasTrazas, Bloque: "a21"}},
		Territorio: TerritorioEsperado{Comunidad: "Comunidad de Madrid", Boletines: []string{"BOCM"}},
	}

	const (
		elTerritorio = "Leganés está en la Comunidad de Madrid y publica en el BOCM.\n\n"
		otroBoletin  = "Leganés está en la Comunidad de Madrid.\n\n"
		sinLaCita    = "El artículo 21 de la Ley 39/2015 regula la obligación de resolver."
		motivoCita   = "cita ausente: " + textoDeLaCita21
		motivoBocm   = "territorio ausente: boletín: BOCM"
	)

	// juicioCon es el juicio de la eval con la sesión de la respuesta dada: su
	// resultado es el que pasa, con esa respuesta y los cambios aplicados.
	juicioCon := func(respuesta string, cambiar func(*ResultadoDeEval)) juicio {
		esperado := ResultadoDeEval{
			Eval:                 eval.Fichero,
			Activa:               true,
			Activada:             true,
			ComandosEjecutados:   []string{textoDelComandoDeLeganes, textoDelComando21},
			CitasEncontradas:     []string{textoDeLaCita21},
			TerritorioEncontrado: []string{"comunidad: Comunidad de Madrid", "boletín: BOCM"},
			Invocaciones: []InvocacionInformada{
				{Orden: ordenDeLeganes, Codigo: codigoDeSalida(0)},
				{Orden: ordenDelArticulo21, Codigo: codigoDeSalida(0)},
			},
			Respuesta:        respuesta,
			CodigoDeLaSesion: codigoDeSalida(0),
			FinDeLaSesion:    "result success",
			SesionTerminada:  true,
			Pasa:             true,
		}

		if cambiar != nil {
			cambiar(&esperado)
		}

		return juicio{
			eval:     eval,
			sesion:   sesionDeTerritorio(respuesta, resuelveLeganes(t), leeElArticulo21(t)),
			skill:    skillDeTerritorio,
			esperado: esperado,
		}
	}

	return []juicio{
		juicioCon(elTerritorio+respuestaConCita, nil),
		juicioCon(elTerritorio+sinLaCita, func(r *ResultadoDeEval) {
			r.CitasEncontradas, r.CitasAusentes = nil, []string{textoDeLaCita21}
			r.Motivos, r.Pasa = []string{motivoCita}, false
		}),
		juicioCon(otroBoletin+respuestaConCita, func(r *ResultadoDeEval) {
			r.TerritorioEncontrado = []string{"comunidad: Comunidad de Madrid"}
			r.TerritorioAusente = []string{"boletín: BOCM"}
			r.Motivos, r.Pasa = []string{motivoBocm}, false
		}),
		juicioCon(otroBoletin+sinLaCita, func(r *ResultadoDeEval) {
			r.CitasEncontradas, r.CitasAusentes = nil, []string{textoDeLaCita21}
			r.TerritorioEncontrado = []string{"comunidad: Comunidad de Madrid"}
			r.TerritorioAusente = []string{"boletín: BOCM"}
			r.Motivos, r.Pasa = []string{motivoCita, motivoBocm}, false
		}),
	}
}

// evalDelMunicipioCubierto es la eval de territorio del municipio cubierto: su
// comando de territorio y un esperado con la comunidad, la provincia, el boletín
// autonómico y un aspecto de cobertura.
func evalDelMunicipioCubierto() Eval {
	return Eval{
		Fichero:  ficheroDelMunicipioCubierto,
		Pregunta: "¿En qué boletines se publican las normas que afectan a Leganés?",
		Activa:   true,
		Comandos: []ComandoEsperado{{Applet: "territorio", Verbo: "resolver", Municipio: "Leganés"}},
		Territorio: TerritorioEsperado{
			Comunidad: "Comunidad de Madrid",
			Provincia: "Madrid",
			Boletines: []string{"BOCM"},
			Cobertura: []string{"boletin_autonomico: configurado"},
		},
	}
}

// resultadoDelMunicipioQuePasa es el resultado de la eval del municipio cubierto
// con la sesión que lo resuelve y declara todo su territorio.
func resultadoDelMunicipioQuePasa() ResultadoDeEval {
	return ResultadoDeEval{
		Eval:                 ficheroDelMunicipioCubierto,
		Activa:               true,
		Activada:             true,
		ComandosEjecutados:   []string{textoDelComandoDeLeganes},
		TerritorioEncontrado: elementosDelMunicipio,
		Invocaciones:         []InvocacionInformada{{Orden: ordenDeLeganes, Codigo: codigoDeSalida(0)}},
		Respuesta:            respuestaDelMunicipio,
		CodigoDeLaSesion:     codigoDeSalida(0),
		FinDeLaSesion:        "result success",
		SesionTerminada:      true,
		Pasa:                 true,
	}
}

// sesionDeTerritorio es la sesión que terminó con código 0 y result success,
// activó legal-core y tiene la respuesta y las invocaciones dadas.
func sesionDeTerritorio(respuesta string, invocaciones ...Invocacion) Sesion {
	return cambiada(sesionTerminada(false, respuesta, invocaciones...), func(s *Sesion) {
		s.SkillsActivadas = []string{skillDeTerritorio}
	})
}

// resuelveLeganes es la invocación de la skill que resuelve el municipio de la
// eval del municipio cubierto y termina con código 0.
func resuelveLeganes(t *testing.T) Invocacion {
	t.Helper()

	return invocada(t, codigoDeSalida(0), deLegalCore("resolver", "Leganés", "--json"))
}

// deLegalCore es el argv con el que la skill instalada invoca el applet
// territorio.
func deLegalCore(tokens ...string) []string {
	return slices.Concat([]string{territorioDeLaSkillInstalada}, tokens)
}

// metadatosNoSatisfaceIndice es el juicio de una eval que espera el índice de la
// LCSP con una sesión que lee sus metadatos y el índice de otra norma: ninguno de
// los dos es el mismo verbo con la misma norma (US4, escenario 7).
func metadatosNoSatisfaceIndice(t *testing.T) juicio {
	t.Helper()

	eval := Eval{
		Fichero:  "02-lcsp-contrato-menor.yaml",
		Pregunta: "¿Qué debe incluir el expediente de un contrato menor según la Ley de Contratos del Sector Público?",
		Activa:   true,
		Comandos: []ComandoEsperado{
			{Applet: "boe", Verbo: "indice", Norma: normaDeLaLCSP},
			{Applet: "boe", Norma: normaDeLaLCSP, Bloque: "a1-30"},
		},
		Citas: []CitaEsperada{{Norma: normaDeLaLCSP, Bloque: "a1-30"}},
	}

	respuesta := "El expediente del contrato menor lo regula el art. 118 de la Ley 9/2017 [BOE-A-2017-12902, bloque a1-30]."

	return juicio{
		eval: eval,
		sesion: sesionTerminada(true, respuesta,
			invocada(t, codigoDeSalida(0), deLaSkill("metadatos", normaDeLaLCSP, "--json")),
			invocada(t, codigoDeSalida(0), deLaSkill("indice", normaDeLasTrazas, "--json")),
			invocada(t, codigoDeSalida(0), deLaSkill("articulo", normaDeLaLCSP, "a1-30", "--json"))),
		esperado: ResultadoDeEval{
			Eval:               eval.Fichero,
			Activa:             true,
			Activada:           true,
			ComandosEjecutados: []string{"bloque boe BOE-A-2017-12902 a1-30"},
			ComandosAusentes:   []string{"boe indice BOE-A-2017-12902"},
			CitasEncontradas:   []string{"BOE-A-2017-12902 a1-30"},
			Invocaciones: []InvocacionInformada{
				{Orden: "boe metadatos BOE-A-2017-12902 --json", Codigo: codigoDeSalida(0)},
				{Orden: "boe indice BOE-A-2015-10565 --json", Codigo: codigoDeSalida(0)},
				{Orden: "boe articulo BOE-A-2017-12902 a1-30 --json", Codigo: codigoDeSalida(0)},
			},
			Respuesta:        respuesta,
			CodigoDeLaSesion: codigoDeSalida(0),
			FinDeLaSesion:    "result success",
			SesionTerminada:  true,
			Motivos:          []string{"comando ausente: boe indice BOE-A-2017-12902"},
		},
	}
}

// terminosComoPalabras es el juicio de una eval con cuatro búsquedas esperadas
// (data-model §6.1). La primera la satisface la invocación cuyos argumentos, en
// minúsculas, contienen cada término como palabra, aunque el término y los
// argumentos no coincidan en mayúsculas, uno de ellos esté dentro de un argumento
// con espacios y otro lleve una barra. Las otras tres no las satisface ninguna: la
// segunda, porque la búsqueda que la busca solo contiene su término dentro de
// otras palabras, detrás y delante de una letra; la tercera, porque a la búsqueda
// que contiene sus dos primeros términos le falta el último; y la cuarta, un
// identificador, porque la única invocación que lo contiene como palabra es
// metadatos y no buscar.
func terminosComoPalabras(t *testing.T) juicio {
	t.Helper()

	eval := Eval{
		Fichero:  "03-lrbrl-busqueda.yaml",
		Pregunta: "¿Qué ley regula las bases del régimen local?",
		Activa:   true,
		Comandos: []ComandoEsperado{
			{Applet: "boe", Verbo: "buscar", Terminos: []string{"Régimen", "local", "7/1985"}},
			{Applet: "boe", Verbo: "buscar", Terminos: []string{"común"}},
			{Applet: "boe", Verbo: "buscar", Terminos: []string{"régimen", "local", "municipal"}},
			{Applet: "boe", Verbo: "buscar", Terminos: []string{normaDeLaLRBRL}},
		},
		Citas: []CitaEsperada{{Norma: normaDeLaLRBRL, Bloque: "a1"}},
	}

	respuesta := "Es la Ley 7/1985, reguladora de las Bases del Régimen Local [BOE-A-1985-5392, bloque a1]."

	return juicio{
		eval: eval,
		sesion: sesionTerminada(true, respuesta,
			invocada(t, codigoDeSalida(0), deLaSkill("buscar", "bases del RÉGIMEN local", "7/1985", "--json")),
			invocada(t, codigoDeSalida(0), deLaSkill("buscar", "procedimiento", "comúnmente", "intercomún", "--json")),
			invocada(t, codigoDeSalida(0), deLaSkill("metadatos", normaDeLaLRBRL, "--json"))),
		esperado: ResultadoDeEval{
			Eval:               eval.Fichero,
			Activa:             true,
			Activada:           true,
			ComandosEjecutados: []string{"boe buscar Régimen local 7/1985"},
			ComandosAusentes: []string{
				"boe buscar común", "boe buscar régimen local municipal", "boe buscar BOE-A-1985-5392",
			},
			CitasEncontradas: []string{"BOE-A-1985-5392 a1"},
			Invocaciones: []InvocacionInformada{
				{Orden: "boe buscar bases del RÉGIMEN local 7/1985 --json", Codigo: codigoDeSalida(0)},
				{Orden: "boe buscar procedimiento comúnmente intercomún --json", Codigo: codigoDeSalida(0)},
				{Orden: "boe metadatos BOE-A-1985-5392 --json", Codigo: codigoDeSalida(0)},
			},
			Respuesta:        respuesta,
			CodigoDeLaSesion: codigoDeSalida(0),
			FinDeLaSesion:    "result success",
			SesionTerminada:  true,
			Motivos: []string{
				"comando ausente: boe buscar común",
				"comando ausente: boe buscar régimen local municipal",
				"comando ausente: boe buscar BOE-A-1985-5392",
			},
		},
	}
}

// citaQueNoCuenta es el juicio de la eval 01 con una sesión que lee su bloque y
// cuya respuesta solo tiene la cita dada, que no es la esperada (US4, escenario
// 3).
func citaQueNoCuenta(t *testing.T, cita string) juicio {
	t.Helper()

	respuesta := "El artículo 21 de la Ley 39/2015 regula la obligación de resolver.\n\n" + cita

	return juicio{
		eval:   evalDelArticulo21(),
		sesion: sesionTerminada(true, respuesta, leeElArticulo21(t)),
		esperado: cambiado(resultadoQuePasa(), func(r *ResultadoDeEval) {
			r.CitasEncontradas = nil
			r.CitasAusentes = []string{textoDeLaCita21}
			r.Respuesta = respuesta
			r.Motivos = []string{"cita ausente: " + textoDeLaCita21}
			r.Pasa = false
		}),
	}
}

// redNoCambiaLaEval es el juicio de la eval 01 con una sesión que pasa y cuyas
// invocaciones conectan: cada invocación informa una vez cada pareja de destino y
// clase, en el orden en que aparece, y una llegada a la red por cada invocación y
// destino de clase red; nada de eso cambia si pasa (data-model §10.2; FR-076).
func redNoCambiaLaEval(t *testing.T) juicio {
	t.Helper()

	socket := Conexion{Familia: "AF_UNIX", Ruta: "/var/run/nscd/socket", Resultado: resultadoSinFichero, Clase: ConexionLocal}

	const otroPublico = "203.0.113.8:443"

	return juicio{
		eval: evalDelArticulo21(),
		sesion: sesionTerminada(true, respuestaConCita,
			invocada(t, codigoDeSalida(5), deLaSkill("articulo", normaDeLasTrazas, "a9998", "--json"),
				conexionInet(destinoPublico, resultadoEnCurso, ConexionRed)),
			invocada(t, codigoDeSalida(0), deLaSkill("articulo", normaDeLasTrazas, "a21", "--json"),
				conexionInet(destinoPublico, resultadoEnCurso, ConexionRed),
				socket,
				conexionInet(destinoDeBucle, resultadoRechazada, ConexionLocal),
				conexionInet(destinoPublico, "0", ConexionRed),
				conexionInet(otroPublico, resultadoSinRuta, ConexionBloqueada),
				socket)),
		esperado: cambiado(resultadoQuePasa(), func(r *ResultadoDeEval) {
			r.Invocaciones = []InvocacionInformada{
				{
					Orden:      "boe articulo BOE-A-2015-10565 a9998 --json",
					Codigo:     codigoDeSalida(5),
					Conexiones: []ConexionInformada{{Destino: destinoPublico, Clase: ConexionRed}},
				},
				{
					Orden:  ordenDelArticulo21,
					Codigo: codigoDeSalida(0),
					Conexiones: []ConexionInformada{
						{Destino: destinoPublico, Clase: ConexionRed},
						{Destino: "unix:/var/run/nscd/socket", Clase: ConexionLocal},
						{Destino: destinoDeBucle, Clase: ConexionLocal},
						{Destino: otroPublico, Clase: ConexionBloqueada},
					},
				},
			}
			r.FueraDeLoGrabado = []InvocacionFallida{{Orden: "boe articulo BOE-A-2015-10565 a9998 --json", Codigo: 5}}
			r.LlegadasALaRed = []LlegadaALaRed{
				{Orden: "boe articulo BOE-A-2015-10565 a9998 --json", Destino: destinoPublico},
				{Orden: ordenDelArticulo21, Destino: destinoPublico},
			}
		}),
	}
}

// mismaSesionConOtraCita son los dos juicios de SC-009 con la misma sesión, la
// que pasa: con la eval 01, pasa; con la eval 01 con su cita esperada cambiada
// por otra, que el informe presenta con el texto dado, falla solo por esa cita.
func mismaSesionConOtraCita(t *testing.T, otra CitaEsperada, texto string) []juicio {
	t.Helper()

	conOtraCita := evalDelArticulo21()
	conOtraCita.Citas = []CitaEsperada{otra}

	return []juicio{
		{eval: evalDelArticulo21(), sesion: sesionQuePasa(t), esperado: resultadoQuePasa()},
		{
			eval:   conOtraCita,
			sesion: sesionQuePasa(t),
			esperado: cambiado(resultadoQuePasa(), func(r *ResultadoDeEval) {
				r.CitasEncontradas = nil
				r.CitasAusentes = []string{texto}
				r.Motivos = []string{"cita ausente: " + texto}
				r.Pasa = false
			}),
		},
	}
}

// conAvisos es el juicio de la eval 01 con los avisos esperados dados y la sesión
// que pasa con la respuesta dada en lugar de la suya: su resultado esperado es el
// de la sesión que pasa, con esa respuesta y los cambios aplicados.
func conAvisos(t *testing.T, avisos []string, respuesta string, cambiar func(*ResultadoDeEval)) juicio {
	t.Helper()

	eval := evalDelArticulo21()
	eval.Avisos = avisos

	return juicio{
		eval:   eval,
		sesion: cambiada(sesionQuePasa(t), func(s *Sesion) { s.Respuesta = respuesta }),
		esperado: cambiado(resultadoQuePasa(), func(r *ResultadoDeEval) {
			r.Respuesta = respuesta
			cambiar(r)
		}),
	}
}

// formasToleradas son los juicios de la eval 01 que espera derogada con la sesión
// que pasa y, delante de su respuesta, cada aviso dado: en todos, derogada queda
// encontrado y la eval pasa (US2, escenario 4).
func formasToleradas(t *testing.T, avisos ...string) []juicio {
	t.Helper()

	juicios := make([]juicio, 0, len(avisos))

	for _, aviso := range avisos {
		juicios = append(juicios, conAvisos(t, []string{"derogada"}, aviso+"\n\n"+respuestaConCita,
			func(r *ResultadoDeEval) {
				r.AvisosEncontrados = []string{"derogada"}
			}))
	}

	return juicios
}

// derogadaAusente es el juicio de la eval 01 que espera derogada con la sesión que
// pasa y la respuesta dada, que no lleva su forma fija: con el comando ejecutado y
// la cita encontrada, derogada queda ausente con su motivo y la eval no pasa (US2,
// escenarios 5 y 6).
func derogadaAusente(t *testing.T, respuesta string) juicio {
	t.Helper()

	return conAvisos(t, []string{"derogada"}, respuesta, func(r *ResultadoDeEval) {
		r.AvisosAusentes = []string{"derogada"}
		r.Motivos = []string{"aviso ausente: derogada"}
		r.Pasa = false
	})
}

// consultaRepetidaQuePasa es el juicio de la eval de la consulta repetida con una
// sesión que comprueba la memoria de la norma antes y después de leer el bloque y
// lo cita: los dos comandos quedan ejecutados, ningún prohibido y la eval pasa.
func consultaRepetidaQuePasa(t *testing.T) juicio {
	t.Helper()

	return juicio{
		eval: evalDeLaConsultaRepetida(),
		sesion: sesionTerminada(true, respuestaConLaCita118,
			compruebaLaNorma(t, codigoDeSalida(0)), leeElArticulo118(t), compruebaLaNorma(t, codigoDeSalida(0))),
		esperado: resultadoDeLaConsultaRepetida(
			InvocacionInformada{Orden: ordenDeLaComprobacionDeLaNorma, Codigo: codigoDeSalida(0)},
			InvocacionInformada{Orden: ordenDelArticulo118, Codigo: codigoDeSalida(0)},
			InvocacionInformada{Orden: ordenDeLaComprobacionDeLaNorma, Codigo: codigoDeSalida(0)},
		),
	}
}

// comprobacionAusente es el juicio de la eval de la consulta repetida con una
// sesión que lee y cita el bloque y en la que ninguna invocación satisface la
// comprobación: graph check con la norma y --describe o --dry-run no consulta;
// con código 7 no termina con 0, y va a las otras fallidas; y graph stats es otro
// verbo.
func comprobacionAusente(t *testing.T) juicio {
	t.Helper()

	const (
		describe = "graph check BOE-A-2017-12902 a1-30 --describe"
		ensayo   = "graph check BOE-A-2017-12902 a1-30 --dry-run"
		stats    = "graph stats --json"
	)

	return juicio{
		eval: evalDeLaConsultaRepetida(),
		sesion: sesionTerminada(true, respuestaConLaCita118,
			invocada(t, codigoDeSalida(0), deKitlegal(strings.Fields(describe)...)),
			invocada(t, codigoDeSalida(0), deKitlegal(strings.Fields(ensayo)...)),
			compruebaLaNorma(t, codigoDeSalida(7)),
			invocada(t, codigoDeSalida(0), deKitlegal(strings.Fields(stats)...)),
			leeElArticulo118(t)),
		esperado: cambiado(resultadoDeLaConsultaRepetida(
			InvocacionInformada{Orden: describe, Codigo: codigoDeSalida(0)},
			InvocacionInformada{Orden: ensayo, Codigo: codigoDeSalida(0)},
			InvocacionInformada{Orden: ordenDeLaComprobacionDeLaNorma, Codigo: codigoDeSalida(7)},
			InvocacionInformada{Orden: stats, Codigo: codigoDeSalida(0)},
			InvocacionInformada{Orden: ordenDelArticulo118, Codigo: codigoDeSalida(0)},
		), func(r *ResultadoDeEval) {
			r.ComandosEjecutados = []string{textoDelComando118}
			r.ComandosAusentes = []string{textoDeLaComprobacionDeLaNorma}
			r.OtrasFallidas = []InvocacionFallida{{Orden: ordenDeLaComprobacionDeLaNorma, Codigo: 7}}
			r.Motivos = []string{"comando ausente: " + textoDeLaComprobacionDeLaNorma}
			r.Pasa = false
		}),
	}
}

// prohibidoEjecutado son los juicios de la eval de la consulta repetida con
// sesiones que comprueban la memoria de la norma, leen y citan el bloque y además
// piden una ficha con graph show: con código 0; con código 3, que va a las otras
// fallidas, y otra vez con código 0, que no lo ejecuta dos veces; y sin código,
// porque el tope cortó la sesión. En los tres, graph show queda ejecutado una
// vez, con su motivo detrás de los de antes, y la eval no pasa.
func prohibidoEjecutado(t *testing.T) []juicio {
	t.Helper()

	ejecutado := func(r *ResultadoDeEval) {
		r.ComandosProhibidosEjecutados = []string{textoDeGraphShow}
		r.Motivos = append(r.Motivos, "comando prohibido ejecutado: "+textoDeGraphShow)
		r.Pasa = false
	}

	comprobada := InvocacionInformada{Orden: ordenDeLaComprobacionDeLaNorma, Codigo: codigoDeSalida(0)}
	leida := InvocacionInformada{Orden: ordenDelArticulo118, Codigo: codigoDeSalida(0)}

	cortada := cambiada(sesionTerminada(true, respuestaConLaCita118,
		compruebaLaNorma(t, codigoDeSalida(0)), leeElArticulo118(t), pideLaFicha(t, nil)), func(s *Sesion) {
		s.Codigo, s.Fin, s.Terminada, s.Cortada = 124, "assistant", false, true
		s.MotivoSinTerminar = "tope de 240 s agotado (código 124)"
	})

	return []juicio{
		{
			eval: evalDeLaConsultaRepetida(),
			sesion: sesionTerminada(true, respuestaConLaCita118,
				compruebaLaNorma(t, codigoDeSalida(0)), leeElArticulo118(t), pideLaFicha(t, codigoDeSalida(0))),
			esperado: cambiado(resultadoDeLaConsultaRepetida(comprobada, leida,
				InvocacionInformada{Orden: ordenDeGraphShow, Codigo: codigoDeSalida(0)}), ejecutado),
		},
		{
			eval: evalDeLaConsultaRepetida(),
			sesion: sesionTerminada(true, respuestaConLaCita118, compruebaLaNorma(t, codigoDeSalida(0)),
				pideLaFicha(t, codigoDeSalida(3)), leeElArticulo118(t), pideLaFicha(t, codigoDeSalida(0))),
			esperado: cambiado(resultadoDeLaConsultaRepetida(comprobada,
				InvocacionInformada{Orden: ordenDeGraphShow, Codigo: codigoDeSalida(3)}, leida,
				InvocacionInformada{Orden: ordenDeGraphShow, Codigo: codigoDeSalida(0)},
			), func(r *ResultadoDeEval) {
				r.OtrasFallidas = []InvocacionFallida{{Orden: ordenDeGraphShow, Codigo: 3}}
				ejecutado(r)
			}),
		},
		{
			eval:   evalDeLaConsultaRepetida(),
			sesion: cortada,
			esperado: cambiado(resultadoDeLaConsultaRepetida(comprobada, leida, InvocacionInformada{Orden: ordenDeGraphShow}),
				func(r *ResultadoDeEval) {
					r.CodigoDeLaSesion = codigoDeSalida(124)
					r.FinDeLaSesion = "assistant"
					r.SesionTerminada = false
					r.Motivos = []string{"la sesión no terminó: tope de 240 s agotado (código 124)"}
					ejecutado(r)
				}),
		},
	}
}

// prohibidoSinConsulta es el juicio de la eval de la consulta repetida con la
// sesión que la pasa y, además, graph show con la ayuda larga y la corta, con
// --describe y con --dry-run, todas con código 0: ninguna consulta, ninguna
// ejecuta el prohibido y la eval pasa.
func prohibidoSinConsulta(t *testing.T) juicio {
	t.Helper()

	invocaciones := []Invocacion{compruebaLaNorma(t, codigoDeSalida(0)), leeElArticulo118(t)}
	informadas := []InvocacionInformada{
		{Orden: ordenDeLaComprobacionDeLaNorma, Codigo: codigoDeSalida(0)},
		{Orden: ordenDelArticulo118, Codigo: codigoDeSalida(0)},
	}

	for _, bandera := range []string{"--help", "-h", "--describe", "--dry-run"} {
		orden := ordenDeGraphShow + " " + bandera
		invocaciones = append(invocaciones, invocada(t, codigoDeSalida(0), deKitlegal(strings.Fields(orden)...)))
		informadas = append(informadas, InvocacionInformada{Orden: orden, Codigo: codigoDeSalida(0)})
	}

	return juicio{
		eval:     evalDeLaConsultaRepetida(),
		sesion:   sesionTerminada(true, respuestaConLaCita118, invocaciones...),
		esperado: resultadoDeLaConsultaRepetida(informadas...),
	}
}

// otroQueElProhibido es el juicio de la eval de la consulta repetida con la
// sesión que la pasa y, además, otro verbo de graph, stats, y el verbo prohibido
// en otro applet, boe show, que el binario rechaza con código 2 y va a las otras
// fallidas: ninguno es graph show y la eval pasa.
func otroQueElProhibido(t *testing.T) juicio {
	t.Helper()

	const (
		stats      = "graph stats --json"
		otroApplet = "boe show BOE-A-2017-12902 --json"
	)

	return juicio{
		eval: evalDeLaConsultaRepetida(),
		sesion: sesionTerminada(true, respuestaConLaCita118,
			invocada(t, codigoDeSalida(0), deKitlegal(strings.Fields(stats)...)),
			invocada(t, codigoDeSalida(2), deKitlegal(strings.Fields(otroApplet)...)),
			compruebaLaNorma(t, codigoDeSalida(0)),
			leeElArticulo118(t)),
		esperado: cambiado(resultadoDeLaConsultaRepetida(
			InvocacionInformada{Orden: stats, Codigo: codigoDeSalida(0)},
			InvocacionInformada{Orden: otroApplet, Codigo: codigoDeSalida(2)},
			InvocacionInformada{Orden: ordenDeLaComprobacionDeLaNorma, Codigo: codigoDeSalida(0)},
			InvocacionInformada{Orden: ordenDelArticulo118, Codigo: codigoDeSalida(0)},
		), func(r *ResultadoDeEval) {
			r.OtrasFallidas = []InvocacionFallida{{Orden: otroApplet, Codigo: 2}}
		}),
	}
}

// prohibidoEntreAusentes es el juicio de la eval de la consulta repetida con una
// sesión que lee el bloque y pide una ficha, sin comprobar la memoria ni citar el
// bloque: el motivo del prohibido va detrás del de la comprobación ausente y
// delante del de la cita ausente.
func prohibidoEntreAusentes(t *testing.T) juicio {
	t.Helper()

	return juicio{
		eval:   evalDeLaConsultaRepetida(),
		sesion: sesionTerminada(true, respuestaSinLaCita118, leeElArticulo118(t), pideLaFicha(t, codigoDeSalida(0))),
		esperado: cambiado(resultadoDeLaConsultaRepetida(
			InvocacionInformada{Orden: ordenDelArticulo118, Codigo: codigoDeSalida(0)},
			InvocacionInformada{Orden: ordenDeGraphShow, Codigo: codigoDeSalida(0)},
		), func(r *ResultadoDeEval) {
			r.ComandosEjecutados = []string{textoDelComando118}
			r.ComandosAusentes = []string{textoDeLaComprobacionDeLaNorma}
			r.ComandosProhibidosEjecutados = []string{textoDeGraphShow}
			r.CitasEncontradas = nil
			r.CitasAusentes = []string{textoDeLaCita118}
			r.Respuesta = respuestaSinLaCita118
			r.Motivos = []string{
				"comando ausente: " + textoDeLaComprobacionDeLaNorma,
				"comando prohibido ejecutado: " + textoDeGraphShow,
				"cita ausente: " + textoDeLaCita118,
			}
			r.Pasa = false
		}),
	}
}

// evalDeLaConsultaRepetida es la eval del contrato evals-y-skill §4 de H7 con la
// forma de la eval 19 del repositorio (contrato eval-y-derivada §1 de H7.2), sin
// el hallazgo esperado: el grafo previo con el bloque a1-30 de la LCSP, la
// lectura de ese bloque y la comprobación de la memoria con su norma, graph show
// prohibido y la cita del bloque.
func evalDeLaConsultaRepetida() Eval {
	bloque := ComandoEsperado{Applet: "boe", Norma: normaDeLaLCSP, Bloque: bloqueDelArticulo118}

	return Eval{
		Fichero: ficheroDeLaConsultaRepetida,
		Pregunta: "Hace tiempo te pregunt\xc3\xa9 qu\xc3\xa9 exige el art\xc3\xadculo 118 de la LCSP para el expediente " +
			"de un contrato menor. \xc2\xbfQu\xc3\xa9 dice ahora?",
		Activa:      true,
		Informativa: true,
		GrafoPrevio: GrafoPrevio{Grabaciones: "lcsp-a1-30-redaccion-original", Comandos: []ComandoEsperado{bloque}},
		Comandos:    []ComandoEsperado{bloque, {Applet: "graph", Verbo: "check", Norma: normaDeLaLCSP}},
		Prohibidos:  []ComandoProhibido{{Applet: "graph", Verbo: "show"}},
		Citas:       []CitaEsperada{{Norma: normaDeLaLCSP, Bloque: bloqueDelArticulo118}},
	}
}

// resultadoDeLaConsultaRepetida es el resultado de la eval de la consulta
// repetida con una sesión terminada y activada, con las invocaciones informadas
// dadas y la cita en la respuesta, que ejecuta los dos comandos esperados y
// ningún prohibido: pasa.
func resultadoDeLaConsultaRepetida(informadas ...InvocacionInformada) ResultadoDeEval {
	return ResultadoDeEval{
		Eval:               ficheroDeLaConsultaRepetida,
		Activa:             true,
		Activada:           true,
		ComandosEjecutados: []string{textoDelComando118, textoDeLaComprobacionDeLaNorma},
		CitasEncontradas:   []string{textoDeLaCita118},
		Invocaciones:       informadas,
		Respuesta:          respuestaConLaCita118,
		CodigoDeLaSesion:   codigoDeSalida(0),
		FinDeLaSesion:      "result success",
		SesionTerminada:    true,
		Pasa:               true,
	}
}

// compruebaElGrafo es la invocación de la skill que comprueba el grafo sin
// norma, con el código dado.
func compruebaElGrafo(t *testing.T, codigo *int) Invocacion {
	t.Helper()

	return invocada(t, codigo, deKitlegal(strings.Fields(ordenDeLaComprobacion)...))
}

// pideLaFicha es la invocación de la skill que pide con graph show la ficha del
// bloque a21 de la Ley 39/2015, con el código dado o sin código.
func pideLaFicha(t *testing.T, codigo *int) Invocacion {
	t.Helper()

	return invocada(t, codigo, deKitlegal(strings.Fields(ordenDeGraphShow)...))
}

// leeElArticulo118 es la invocación de la skill que lee el bloque de la eval de
// la consulta repetida y termina con código 0.
func leeElArticulo118(t *testing.T) Invocacion {
	t.Helper()

	return invocada(t, codigoDeSalida(0), deLaSkill("articulo", normaDeLaLCSP, bloqueDelArticulo118, "--json"))
}

// evalDeLaRedaccionCambiada es la eval del contrato evals-y-skill §5 de H7.1,
// como la eval 19 del repositorio desde H7.2: la de la consulta repetida con el
// hallazgo version-obsoleta esperado.
func evalDeLaRedaccionCambiada() Eval {
	eval := evalDeLaConsultaRepetida()
	eval.Hallazgos = []string{versionObsoleta}

	return eval
}

// conHallazgo es el juicio de la eval de la redacción cambiada con la sesión
// terminada y activada que lee el bloque, comprueba la memoria con su norma y
// ese bloque y responde con la respuesta dada: su resultado esperado ejecuta los
// dos comandos, sin ningún prohibido, con esa respuesta y los cambios aplicados.
func conHallazgo(t *testing.T, respuesta string, cambiar func(*ResultadoDeEval)) juicio {
	t.Helper()

	return juicio{
		eval:   evalDeLaRedaccionCambiada(),
		sesion: sesionTerminada(true, respuesta, leeElArticulo118(t), compruebaLaNorma(t, codigoDeSalida(0))),
		esperado: cambiado(resultadoDeLaConsultaRepetida(
			InvocacionInformada{Orden: ordenDelArticulo118, Codigo: codigoDeSalida(0)},
			InvocacionInformada{Orden: ordenDeLaComprobacionDeLaNorma, Codigo: codigoDeSalida(0)},
		), func(r *ResultadoDeEval) {
			r.Respuesta = respuesta
			cambiar(r)
		}),
	}
}

// hallazgoTolerado son los juicios de la eval de la redacción cambiada con la
// sesión de conHallazgo y, delante de su respuesta con la cita, cada traslado
// dado: en todos, version-obsoleta queda encontrado y la eval pasa.
func hallazgoTolerado(t *testing.T, traslados ...string) []juicio {
	t.Helper()

	juicios := make([]juicio, 0, len(traslados))

	for _, traslado := range traslados {
		juicios = append(juicios, conHallazgo(t, traslado+"\n\n"+respuestaConLaCita118, func(r *ResultadoDeEval) {
			r.HallazgosEncontrados = []string{versionObsoleta}
		}))
	}

	return juicios
}

// versionObsoletaAusente deja version-obsoleta ausente, con su motivo, y la
// eval sin pasar.
func versionObsoletaAusente(r *ResultadoDeEval) {
	r.HallazgosAusentes = []string{versionObsoleta}
	r.Motivos = append(r.Motivos, motivoDeVersionObsoleta)
	r.Pasa = false
}

// hallazgoDetrasDeLosAusentes es el juicio de la eval de la redacción cambiada
// que espera además el aviso derogada, con la sesión de conHallazgo y una
// respuesta sin la cita, sin el aviso y sin el hallazgo: el motivo del hallazgo
// va detrás del de la cita y del del aviso.
func hallazgoDetrasDeLosAusentes(t *testing.T) juicio {
	t.Helper()

	caso := conHallazgo(t, respuestaSinLaCita118, func(r *ResultadoDeEval) {
		r.CitasEncontradas = nil
		r.CitasAusentes = []string{textoDeLaCita118}
		r.AvisosAusentes = []string{"derogada"}
		r.Motivos = []string{"cita ausente: " + textoDeLaCita118, "aviso ausente: derogada"}
		versionObsoletaAusente(r)
	})
	caso.eval.Avisos = []string{"derogada"}

	return caso
}

// comprobacionConOtraNorma son los juicios de la sesión que lee y cita el bloque,
// traslada el cambio con su forma fija y comprueba la memoria de otra norma y sin
// norma, sin comprobarla de la suya: con la eval de la redacción cambiada, su
// comando de comprobación queda ausente con su motivo y la eval no pasa, aunque
// el hallazgo esté; con la de la consulta repetida con la comprobación de H7, que
// no lleva norma, cualquiera de las dos la satisface y la eval pasa.
func comprobacionConOtraNorma(t *testing.T) []juicio {
	t.Helper()

	const (
		otraNorma = "graph check BOE-A-2015-10565 --json"
		respuesta = trasladoDelCambio + "\n\n" + respuestaConLaCita118
	)

	sesion := sesionTerminada(true, respuesta,
		leeElArticulo118(t),
		invocada(t, codigoDeSalida(0), deKitlegal(strings.Fields(otraNorma)...)),
		compruebaElGrafo(t, codigoDeSalida(0)))

	informadas := []InvocacionInformada{
		{Orden: ordenDelArticulo118, Codigo: codigoDeSalida(0)},
		{Orden: otraNorma, Codigo: codigoDeSalida(0)},
		{Orden: ordenDeLaComprobacion, Codigo: codigoDeSalida(0)},
	}

	// La comprobación de H7 no lleva norma (contrato evals-y-skill §4 de H7).
	conLaComprobacionDeH7 := evalDeLaConsultaRepetida()
	conLaComprobacionDeH7.Comandos = []ComandoEsperado{
		conLaComprobacionDeH7.Comandos[0], {Applet: "graph", Verbo: "check"},
	}

	return []juicio{
		{
			eval:   evalDeLaRedaccionCambiada(),
			sesion: sesion,
			esperado: cambiado(resultadoDeLaConsultaRepetida(informadas...), func(r *ResultadoDeEval) {
				r.ComandosEjecutados = []string{textoDelComando118}
				r.ComandosAusentes = []string{textoDeLaComprobacionDeLaNorma}
				r.HallazgosEncontrados = []string{versionObsoleta}
				r.Respuesta = respuesta
				r.Motivos = []string{"comando ausente: " + textoDeLaComprobacionDeLaNorma}
				r.Pasa = false
			}),
		},
		{
			eval:   conLaComprobacionDeH7,
			sesion: sesion,
			esperado: cambiado(resultadoDeLaConsultaRepetida(informadas...), func(r *ResultadoDeEval) {
				r.ComandosEjecutados = []string{textoDelComando118, textoDeLaComprobacion}
				r.Respuesta = respuesta
			}),
		},
	}
}

// compruebaLaNorma es la invocación de la skill que comprueba la memoria de la
// norma de la eval de la consulta repetida con el bloque que ha leído, con el
// código dado.
func compruebaLaNorma(t *testing.T, codigo *int) Invocacion {
	t.Helper()

	return invocada(t, codigo, deKitlegal(strings.Fields(ordenDeLaComprobacionDeLaNorma)...))
}

// deKitlegal es el argv con el que la skill invoca un applet por el binario
// kitlegal del PATH (ADR 0019).
func deKitlegal(tokens ...string) []string {
	return slices.Concat([]string{"kitlegal"}, tokens)
}

// evalDelArticulo21 es la eval 01 del contrato evals-y-grabaciones §1: bloque y
// cita a21 de la Ley 39/2015.
func evalDelArticulo21() Eval {
	return Eval{
		Fichero:  ficheroDeLaEval01,
		Pregunta: preguntaDelArticulo21Eval,
		Activa:   true,
		Comandos: []ComandoEsperado{{Applet: "boe", Norma: normaDeLasTrazas, Bloque: "a21"}},
		Citas:    []CitaEsperada{{Norma: normaDeLasTrazas, Bloque: "a21"}},
	}
}

// evalDeNoActivacion es la eval de no activación del contrato
// evals-y-grabaciones §1.
func evalDeNoActivacion() Eval {
	return Eval{Fichero: ficheroDeNoActivacion, Pregunta: "¿Cómo invierto una lista enlazada en Go?"}
}

// sesionTerminada es la sesión que terminó con código 0 y result success, con la
// respuesta y las invocaciones dadas; activada, si cargó la skill.
func sesionTerminada(activada bool, respuesta string, invocaciones ...Invocacion) Sesion {
	sesion := Sesion{
		Modelo:              modeloDeLasSesiones,
		VersionDeClaudeCode: versionDeLasSesiones,
		Respuesta:           respuesta,
		Fin:                 "result success",
		Terminada:           true,
		Invocaciones:        invocaciones,
	}

	if activada {
		sesion.SkillsActivadas = []string{skillDeLasSesiones}
	}

	return sesion
}

// sesionQuePasa es la sesión de la eval 01 que pasa: terminada, activada, con las
// otras invocaciones dadas, la lectura de su bloque con código 0 detrás y la cita
// en la respuesta.
func sesionQuePasa(t *testing.T, otras ...Invocacion) Sesion {
	t.Helper()

	return sesionTerminada(true, respuestaConCita, slices.Concat(otras, []Invocacion{leeElArticulo21(t)})...)
}

// resultadoQuePasa es el resultado de la eval 01 con sesionQuePasa: con las
// otras invocaciones dadas antes de la lectura de su bloque, y sin nada fuera de
// lo grabado, fallido ni llegado a la red.
func resultadoQuePasa(otras ...InvocacionInformada) ResultadoDeEval {
	return ResultadoDeEval{
		Eval:               ficheroDeLaEval01,
		Activa:             true,
		Activada:           true,
		ComandosEjecutados: []string{textoDelComando21},
		CitasEncontradas:   []string{textoDeLaCita21},
		Invocaciones: slices.Concat(otras, []InvocacionInformada{
			{Orden: ordenDelArticulo21, Codigo: codigoDeSalida(0)},
		}),
		Respuesta:        respuestaConCita,
		CodigoDeLaSesion: codigoDeSalida(0),
		FinDeLaSesion:    "result success",
		SesionTerminada:  true,
		Pasa:             true,
	}
}

// leeElArticulo21 es la invocación de la skill que lee el bloque de la eval 01 y
// termina con código 0.
func leeElArticulo21(t *testing.T) Invocacion {
	t.Helper()

	return invocada(t, codigoDeSalida(0), deLaSkill("articulo", normaDeLasTrazas, "a21", "--json"))
}

// invocada es la invocación de argv tal como la interpreta InterpretarInvocacion,
// con el código y las conexiones dados: la que LeerTrazas daría de una traza.
func invocada(t *testing.T, codigo *int, argv []string, conexiones ...Conexion) Invocacion {
	t.Helper()

	invocacion, deApplet, err := InterpretarInvocacion(argv)
	require.NoError(t, err)
	require.True(t, deApplet, "%q invoca un applet", argv)

	invocacion.Codigo = codigo
	invocacion.Conexiones = conexiones

	return invocacion
}

// deLaSkill es el argv con el que la skill instalada invoca el applet boe.
func deLaSkill(tokens ...string) []string {
	return slices.Concat([]string{boeDeLaSkillInstalada}, tokens)
}

// cambiado es una copia del resultado con los cambios aplicados.
func cambiado(resultado ResultadoDeEval, cambiar func(*ResultadoDeEval)) ResultadoDeEval {
	cambiar(&resultado)

	return resultado
}

// cambiada es una copia de la sesión con los cambios aplicados.
func cambiada(sesion Sesion, cambiar func(*Sesion)) Sesion {
	cambiar(&sesion)

	return sesion
}

// sinElBloque21 es el juicio de la eval 01 con una sesión activada, con la cita
// esperada en la respuesta y con una invocación de la skill por cada orden dada
// sin el applet, todas con código 0 y ninguna de las cuales satisface el bloque
// esperado: el comando queda ausente, la eval no pasa y ninguna invocación va a
// fuera de lo grabado ni a las otras fallidas.
func sinElBloque21(t *testing.T, ordenes ...string) juicio {
	t.Helper()

	invocaciones := make([]Invocacion, 0, len(ordenes))
	informadas := make([]InvocacionInformada, 0, len(ordenes))

	for _, orden := range ordenes {
		invocaciones = append(invocaciones, invocada(t, codigoDeSalida(0), deLaSkill(strings.Fields(orden)...)))
		informadas = append(informadas, InvocacionInformada{Orden: "boe " + orden, Codigo: codigoDeSalida(0)})
	}

	return juicio{
		eval:   evalDelArticulo21(),
		sesion: sesionTerminada(true, respuestaConCita, invocaciones...),
		esperado: cambiado(resultadoQuePasa(), func(r *ResultadoDeEval) {
			r.ComandosEjecutados = nil
			r.ComandosAusentes = []string{textoDelComando21}
			r.Invocaciones = informadas
			r.Motivos = []string{"comando ausente: " + textoDelComando21}
			r.Pasa = false
		}),
	}
}

// describeYDryRunFalsos es el juicio de la eval 01 con una sesión cuyas dos
// invocaciones llevan --dry-run=false o --describe=false, con las que el binario
// consulta: la de a9998 con --offline, que termina con código 4, va a fuera de lo
// grabado, y la del bloque, con código 0, satisface el comando esperado.
func describeYDryRunFalsos(t *testing.T) juicio {
	t.Helper()

	const (
		fuera   = "articulo BOE-A-2015-10565 a9998 --dry-run=false --offline --json"
		lectura = "articulo BOE-A-2015-10565 a21 --describe=false --json"
	)

	return juicio{
		eval: evalDelArticulo21(),
		sesion: sesionTerminada(true, respuestaConCita,
			invocada(t, codigoDeSalida(4), deLaSkill(strings.Fields(fuera)...)),
			invocada(t, codigoDeSalida(0), deLaSkill(strings.Fields(lectura)...))),
		esperado: cambiado(resultadoQuePasa(), func(r *ResultadoDeEval) {
			r.Invocaciones = []InvocacionInformada{
				{Orden: "boe " + fuera, Codigo: codigoDeSalida(4)},
				{Orden: "boe " + lectura, Codigo: codigoDeSalida(0)},
			}
			r.FueraDeLoGrabado = []InvocacionFallida{{Orden: "boe " + fuera, Codigo: 4}}
		}),
	}
}
