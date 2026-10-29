package evals

import (
	"bytes"
	"cmp"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/app"
	"github.com/jmorenobl/kitlegal/internal/cache"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
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
	skillDeTerritorio            = "plazos-y-territorio"
	territorioDeLaSkillInstalada = "/home/runner/.claude/skills/plazos-y-territorio/scripts/territorio"

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

// Lo que juzgan la eval de los dos bloques, sintética con las dos redacciones
// modificadas de la eval 20 de normativa, y sus sesiones de TestJuzgar,
// construidas en memoria (contracts/evals-y-juicio.md §2 y §3 de H7.4).
const (
	ficheroDeLosDosBloques = "20-lcsp-dos-bloques-redaccion-cambiada.yaml"

	// bloqueDeLaDA3 es el otro bloque que lee y cita esa eval, la disposición
	// adicional tercera de la LCSP; textoDelComandoDA3 y textoDeLaCitaDA3, su
	// comando esperado y su cita esperada, con el texto con el que los presenta el
	// informe.
	bloqueDeLaDA3      = "da-3"
	textoDelComandoDA3 = "bloque boe BOE-A-2017-12902 da-3"
	textoDeLaCitaDA3   = "BOE-A-2017-12902 da-3"

	// ordenDeLosDosBloques es la orden de la skill que lee los dos bloques de una
	// vez, y ordenDeLaComprobacionDeLosDosBloques, la que comprueba la memoria de
	// la norma con los dos.
	ordenDeLosDosBloques                 = "boe articulos BOE-A-2017-12902 a1-30 da-3 --json"
	ordenDeLaComprobacionDeLosDosBloques = "graph check BOE-A-2017-12902 a1-30 da-3 --json"

	// fechaDeLaRedaccionOriginal es la fecha de vigencia de la redacción original
	// de los dos bloques, la superada; fechaDelArticulo118Vigente y
	// fechaDeLaDA3Vigente, las de la redacción vigente de cada uno, la leída.
	fechaDeLaRedaccionOriginal = "20180309"
	fechaDelArticulo118Vigente = "20200206"
	fechaDeLaDA3Vigente        = "20230101"

	// redaccionDelArticulo118 y redaccionDeLaDA3 son las dos redacciones
	// modificadas que espera la eval, con el texto con el que las presentan el
	// informe y los motivos: la norma, el bloque y las dos fechas.
	redaccionDelArticulo118 = "BOE-A-2017-12902 a1-30 20180309 20200206"
	redaccionDeLaDA3        = "BOE-A-2017-12902 da-3 20180309 20230101"

	// citaDelArticulo118 y citaDeLaDA3 son las citas de los dos bloques con la
	// forma de «Cómo se cita», la que va detrás de la etiqueta de la línea
	// (research D4 de H7.4).
	citaDelArticulo118 = "art. 118 de la Ley 9/2017 [BOE-A-2017-12902, bloque a1-30]"
	citaDeLaDA3        = "disposici\xc3\xb3n adicional tercera de la Ley 9/2017 [BOE-A-2017-12902, bloque da-3]"

	// cuerpoConLasDosCitas es el cuerpo de la respuesta, con la cita de los dos
	// bloques fuera de toda línea ⚠ REDACCIÓN MODIFICADA:.
	cuerpoConLasDosCitas = "El art\xc3\xadculo 118 de la LCSP regula el expediente del contrato menor " +
		"[BOE-A-2017-12902, bloque a1-30], y su disposici\xc3\xb3n adicional tercera, las normas espec\xc3\xadficas " +
		"de contrataci\xc3\xb3n en las entidades locales [BOE-A-2017-12902, bloque da-3]."

	// principioDeRedaccionAusente es el principio del motivo de una redacción
	// modificada ausente, al que sigue su texto, escrito a mano.
	principioDeRedaccionAusente = "redacci\xc3\xb3n modificada ausente: "
)

// Lo que juzgan la eval de plazos-y-territorio que declara que normativa no se
// activa y sus sesiones de TestJuzgar (contracts/evals-y-juicio.md §2 y §5 de
// H7.4; FR-003, FR-004, FR-094).
const (
	// skillQueNoSeActiva es la skill que declaran las evals de
	// plazos-y-territorio en no_se_activan, y motivoDeNormativaActivada, el
	// motivo de la sesión que la activa, escrito a mano.
	skillQueNoSeActiva        = "normativa"
	motivoDeNormativaActivada = "se activ\xc3\xb3 la skill normativa, que la eval dice que no se activa"
)

// Lo que juzgan las evals del municipio cubierto que declaran las skills que se
// activan con plazos-y-territorio y sus sesiones de TestJuzgar
// (contracts/evals.md §2 de H32; research D14 de H32; FR-036).
const (
	// motivoDeNormativaSinActivar y motivoDeJurisprudenciaSinActivar son los
	// motivos de la sesión que no activa una skill que la eval declara en
	// se_activan, escritos a mano.
	motivoDeNormativaSinActivar      = "no se activó la skill normativa, que la eval dice que se activa"
	motivoDeJurisprudenciaSinActivar = "no se activó la skill jurisprudencia, que la eval dice que se activa"

	// motivoDeJurisprudenciaActivada es el motivo de la sesión que activa
	// jurisprudencia cuando la eval la declara en no_se_activan, escrito a mano.
	motivoDeJurisprudenciaActivada = "se activó la skill jurisprudencia, que la eval dice que no se activa"
)

// Respuestas con expresiones de la lista de normativa (contrato
// lista-y-juicio §3 y §4 de H7.2; contracts/lista-de-expresiones.md §6 de H7.3),
// con las que los tests del juicio, del informe y del sondeo fijan que la lista
// ya no juzga ninguna (FR-070 de H24).
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
// es la skill con la que se juzga; vacía, la de las sesiones de normativa.
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
//
// Desde H7.4, con una eval sintética con las dos redacciones modificadas de la
// eval 20, cada redacción esperada está solo si una línea con la forma fija de
// version-obsoleta lleva, como primera cita, la de su bloque y, como primeras
// fechas, las suyas en su orden: con una línea por bloque, leídos en una orden o
// en dos, pasa; con una sola, con las dos sin cita, con la cita de un bloque y las
// fechas del otro, con las fechas en otro orden o con las dos en una sola línea,
// no pasa, y cada ausente va a las ausentes con su motivo, detrás del de los
// hallazgos; una eval sin redacciones esperadas
// se juzga como antes aunque la respuesta lleve las líneas. Y una sesión que
// activa una skill de no_se_activan no pasa, con un motivo que la nombra detrás
// del de la activación y delante del de los comandos, en una eval de
// plazos-y-territorio que activa la skill y en una de no activación; sin
// activarla, o si la eval no la declara, se juzga como antes
// (contracts/evals-y-juicio.md §2 de H7.4; FR-003, FR-004, FR-023, FR-052,
// FR-053, FR-094; SC-007).
//
// Desde H32, una sesión que no activa una skill de se_activan no pasa, con un
// motivo que la nombra por cada una que falta, en el orden de la eval, detrás
// del de la activación y delante de los de no_se_activan; activándolas todas,
// en cualquier orden, pasa (contracts/evals.md §2 de H32; research D14 de H32;
// FR-036; SC-015).
func TestJuzgar(t *testing.T) {
	t.Parallel()

	derogada := []string{"derogada"}
	lasDosRedacciones := []string{redaccionDelArticulo118, redaccionDeLaDA3}

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
						"la activación no coincide: se esperaba que la skill normativa se activara y no se activó",
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
					Modo:             ModoOrden,
					Activada:         true,
					Respuesta:        respuestaSinSkill,
					CodigoDeLaSesion: codigoDeSalida(0),
					FinDeLaSesion:    "result success",
					SesionTerminada:  true,
					Motivos: []string{
						"la activación no coincide: se esperaba que la skill normativa no se activara y se activó",
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
					Modo:             ModoOrden,
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
						{Orden: "boe articulo BOE-A-2015-10565 a21 --offline --json", Codigo: codigoDeSalida(4)},
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
					SkillsActivadas:     []string{skillNormativa},
					Codigo:              124,
					Fin:                 "assistant",
					Cortada:             true,
					MotivoSinTerminar:   "tope de 240 s agotado (código 124)",
					Invocaciones:        []Invocacion{invocada(t, nil, deLaSkill("articulo", normaDeLasTrazas, "a21", "--json"))},
				},
				esperado: ResultadoDeEval{
					Eval:             ficheroDeLaEval01,
					Modo:             ModoOrden,
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
						{Orden: "boe articulo BOE-A-2015-10565 a9998 --json", Codigo: codigoDeSalida(5)},
						{Orden: "boe articulo BOE-A-2015-10565 a9998 --offline --json", Codigo: codigoDeSalida(4)},
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
						{Orden: "boe articulo BOE-A-15-1 a21 --json", Codigo: codigoDeSalida(2)},
						{Orden: "boe articulo BOE-A-2015-10565 a9999 --json", Codigo: codigoDeSalida(3)},
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
			nombre:  "redacciones-modificadas-una-por-bloque",
			juicios: []juicio{conLasLineas(t, []string{lineaConLaCitaDel118(), lineaConLaCitaDeLaDA3()}, lasDosRedacciones, nil)},
		},
		{
			// Leer los dos bloques en dos órdenes satisface los dos, como en una.
			nombre:  "redacciones-modificadas-leidas-en-dos-ordenes",
			juicios: []juicio{losDosBloquesEnDosOrdenes(t)},
		},
		{
			nombre: "redaccion-modificada-una-sola",
			juicios: []juicio{conLasLineas(t, []string{lineaConLaCitaDel118()},
				[]string{redaccionDelArticulo118}, []string{redaccionDeLaDA3})},
		},
		{
			nombre: "redacciones-modificadas-sin-cita",
			juicios: []juicio{conLasLineas(t, []string{
				lineaDeRedaccion("", fechaDeLaRedaccionOriginal, fechaDelArticulo118Vigente),
				lineaDeRedaccion("", fechaDeLaRedaccionOriginal, fechaDeLaDA3Vigente),
			}, nil, lasDosRedacciones)},
		},
		{
			nombre: "redacciones-modificadas-con-la-cita-de-un-bloque-y-las-fechas-del-otro",
			juicios: []juicio{conLasLineas(t, []string{
				lineaDeRedaccion(citaDelArticulo118, fechaDeLaRedaccionOriginal, fechaDeLaDA3Vigente),
				lineaDeRedaccion(citaDeLaDA3, fechaDeLaRedaccionOriginal, fechaDelArticulo118Vigente),
			}, nil, lasDosRedacciones)},
		},
		{
			nombre: "redacciones-modificadas-con-las-fechas-en-otro-orden",
			juicios: []juicio{conLasLineas(t, []string{
				lineaDeRedaccion(citaDelArticulo118, fechaDelArticulo118Vigente, fechaDeLaRedaccionOriginal),
				lineaDeRedaccion(citaDeLaDA3, fechaDeLaDA3Vigente, fechaDeLaRedaccionOriginal),
			}, nil, lasDosRedacciones)},
		},
		{
			nombre: "redacciones-modificadas-en-una-sola-linea",
			juicios: []juicio{conLasLineas(t, []string{lineaConLaCitaDel118() + " " + lineaConLaCitaDeLaDA3()},
				[]string{redaccionDelArticulo118}, []string{redaccionDeLaDA3})},
		},
		{
			nombre:  "redacciones-modificadas-ausentes-detras-del-hallazgo",
			juicios: []juicio{redaccionesDetrasDelHallazgo(t)},
		},
		{
			// La eval 19 no espera redacciones modificadas: las dos líneas, con su
			// cita y sus fechas, no cambian su juicio.
			nombre: "sin-redacciones-modificadas-el-juicio-de-antes",
			juicios: []juicio{conHallazgo(t,
				lineaConLaCitaDel118()+"\n\n"+lineaConLaCitaDeLaDA3()+"\n\n"+respuestaConLaCita118, func(r *ResultadoDeEval) {
					r.HallazgosEncontrados = []string{versionObsoleta}
				})},
		},
		{
			// La sesión de la eval de plazos-y-territorio activa también normativa:
			// no pasa, con el motivo que la nombra; sin activarla, pasa.
			nombre:  "no-se-activan",
			juicios: seActivaLaQueNoSeActiva(t),
		},
		{
			nombre:  "no-se-activan-detras-del-motivo-de-plazos-y-territorio",
			juicios: []juicio{laQueNoSeActivaDetrasDeLaActivacion(t)},
		},
		{
			// La sesión de la eval que declara que normativa y jurisprudencia se
			// activan con plazos-y-territorio no pasa sin activar alguna, con un
			// motivo que nombra cada una que falta; activándolas todas, pasa.
			nombre:  "se-activan",
			juicios: seActivanLasQueSeActivan(t),
		},
		{
			nombre:  "se-activan-delante-de-no-se-activan",
			juicios: []juicio{lasQueSeActivanDelanteDeLasQueNo()},
		},
		{
			// La eval 01 no declara no_se_activan: otra skill activada no cambia su
			// juicio.
			nombre: "sin-no-se-activan-el-juicio-de-antes",
			juicios: []juicio{{
				eval: evalDelArticulo21(),
				sesion: cambiada(sesionQuePasa(t), func(s *Sesion) {
					s.SkillsActivadas = []string{skillDeTerritorio, skillNormativa}
				}),
				esperado: resultadoQuePasa(),
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
				assert.Equal(t, j.esperado, Juzgar(j.eval, j.sesion, cmp.Or(j.skill, skillNormativa)),
					"la eval %s con la sesión del caso", j.eval.Fichero)
			}
		})
	}
}

// TestJuzgarLasEvalsSinHallazgos fija que las evals del repositorio de
// normativa y de plazos-y-territorio que no esperan hallazgos se juzgan igual
// que antes de H7.1 (contrato evals-y-skill §1 de H7.1; FR-054): con una sesión
// terminada, activada si la eval espera que se active, cuya respuesta lleva la
// cita del art. 21 de la Ley 39/2015, el resultado es el mismo con la forma fija
// de version-obsoleta delante de la respuesta que sin ella, salvo la respuesta, y
// no reparte ningún hallazgo.
func TestJuzgarLasEvalsSinHallazgos(t *testing.T) {
	t.Parallel()

	for dir, skill := range map[string]string{evalsDeNormativa: skillNormativa, evalsDePlazosYTerritorio: skillDeTerritorio} {
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

			codificado, err := json.Marshal(Juzgar(caso.juicio.eval, caso.juicio.sesion, skillNormativa))
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

// TestPorMateriaDelResultadoEnJSON fija que el resultado de una sesión lleva la
// clase de pregunta de su eval en por_materia, siempre escrito, codificado como
// lo codifica EscribirInforme: la serie de cierres reparte con él las
// respuestas sin leer las evals (ADR 0038, punto 6). Que pase o no la sesión no
// lo cambia.
func TestPorMateriaDelResultadoEnJSON(t *testing.T) {
	t.Parallel()

	for _, porMateria := range []bool{false, true} {
		t.Run(strconv.FormatBool(porMateria), func(t *testing.T) {
			t.Parallel()

			eval := evalDelArticulo21()
			eval.PorMateria = porMateria

			codificado, err := json.Marshal(Juzgar(eval, sesionQuePasa(t), skillNormativa))
			require.NoError(t, err)

			var crudo map[string]jsontext.Value

			require.NoError(t, json.Unmarshal(codificado, &crudo))
			require.Contains(t, crudo, "por_materia", "la clave se escribe también con false")
			assert.Equal(t, strconv.FormatBool(porMateria), string(crudo["por_materia"]))
		})
	}
}

// TestRedaccionesDelResultadoEnJSON fija que el resultado de una sesión escribe,
// detrás de hallazgos_ausentes, redacciones_modificadas_encontradas y
// redacciones_modificadas_ausentes como listas, nunca null, codificado como lo
// codifica EscribirInforme (data-model §4 de H7.4; contracts/evals-y-juicio.md §2;
// FR-053): vacías con una eval que no espera redacciones, y con el texto de cada
// una en la que corresponde con la eval de los dos bloques y una sola línea.
func TestRedaccionesDelResultadoEnJSON(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre      string
		juicio      juicio
		encontradas string
		ausentes    string
	}{
		{
			nombre:      "sin-redacciones",
			juicio:      juicio{eval: evalDelArticulo21(), sesion: sesionQuePasa(t)},
			encontradas: "[]",
			ausentes:    "[]",
		},
		{
			nombre: "una-encontrada-y-una-ausente",
			juicio: conLasLineas(t, []string{lineaConLaCitaDel118()},
				[]string{redaccionDelArticulo118}, []string{redaccionDeLaDA3}),
			encontradas: `["` + redaccionDelArticulo118 + `"]`,
			ausentes:    `["` + redaccionDeLaDA3 + `"]`,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			codificado, err := json.Marshal(Juzgar(caso.juicio.eval, caso.juicio.sesion, skillNormativa))
			require.NoError(t, err)

			var crudo struct {
				Encontradas jsontext.Value `json:"redacciones_modificadas_encontradas"`
				Ausentes    jsontext.Value `json:"redacciones_modificadas_ausentes"`
			}

			require.NoError(t, json.Unmarshal(codificado, &crudo))
			assert.Equal(t, caso.encontradas, string(crudo.Encontradas))
			assert.Equal(t, caso.ausentes, string(crudo.Ausentes))

			assert.Equal(t, "redacciones_modificadas_encontradas", claveDetras(t, string(codificado), "hallazgos_ausentes"))
			assert.Equal(t, "redacciones_modificadas_ausentes",
				claveDetras(t, string(codificado), "redacciones_modificadas_encontradas"))
		})
	}
}

// Lo que la lista de expresiones dejaba en el resultado de una sesión hasta H24,
// escrito a mano: la clave con las expresiones que llevaba su respuesta y el
// principio del motivo de cada una, al que seguía la expresión
// (contracts/informe-del-job.md §6 de H24).
const (
	claveDeLasExpresiones         = "expresiones_prohibidas"
	principioDeExpresionProhibida = "expresión prohibida: "
)

// Las premisas de TestJuzgarSinLaLista sobre el calibrado de H7.4 (FR-111 de
// H24; H7.4 FR 032): las respuestas que la lista del repositorio marca en los
// tres informes y, de ellas, las de una eval cuyo fichero ya no está en la
// carpeta —las de la que retiró H7.2, que el informe de H7.1 nombra y aquí se
// nombra solo por sus dos cifras (FR-020 de H7.2)—.
const (
	marcadasEnElCalibrado     = 56
	marcadasDeLaEvalRetirada  = 2
	cifrasDeLaEvalRetirada    = "19"
	informeDeLaEvalRetirada   = informeDeH71
	cifrasDelFicheroDeUnaEval = 2
)

// marcadasDeCadaInforme son los tres informes del calibrado de H7.4, en su
// orden, con las respuestas que la lista del repositorio marca en cada uno. Son
// lo que queda de ese calibrado: su reparto por eval y por familia ya no se
// exige (FR-071 de H24).
var marcadasDeCadaInforme = []struct {
	informe  string
	marcadas int
}{
	{informe: informeDeH71, marcadas: 36},
	{informe: informeDeH72, marcadas: 11},
	{informe: informeDeH73, marcadas: 9},
}

// TestJuzgarSinLaLista fija que la lista de expresiones ya no juzga ninguna
// respuesta (contracts/informe-del-job.md §6 y §9 de H24; FR-070, FR-111;
// SC-011), con las evals que LeerConjunto lee de la carpeta de normativa y
// la lista de esa carpeta, que sigue en ella para la prosa de la skill:
//
//   - la respuesta de la eval sin binario ni servidor que ofrece repetir la
//     consulta, la de las dos sesiones del job que la lista daba por un anuncio
//     (docs/USO.md, 2026-10-04), en una sesión que cumple todo lo demás de su
//     eval, pasa;
//   - cada respuesta que la lista marca en los tres informes del calibrado de
//     H7.4 —36, 11 y 9, exigidas como premisa—, juzgada con su eval y una sesión
//     terminada hecha de su entrada, deja un resultado sin la clave de las
//     expresiones y sin ningún motivo de una: 0 sesiones marcadas por la lista.
//
// No se busca la expresión dentro de los motivos: graph check es una expresión
// de la lista y también el texto de un comando ausente.
func TestJuzgarSinLaLista(t *testing.T) {
	t.Parallel()

	conjunto, err := LeerConjunto(evalsDeNormativa)
	require.NoError(t, err)

	lista := listaDelRepositorio(t, conjunto)

	t.Run("sin-binario-ni-servidor", func(t *testing.T) {
		t.Parallel()
		juzgarLaOfertaSinLaLista(t, conjunto.Evals, lista)
	})

	t.Run("calibrado", func(t *testing.T) {
		t.Parallel()
		juzgarElCalibradoSinLaLista(t, conjunto.Evals, lista)
	})
}

// juzgarLaOfertaSinLaLista juzga la eval sin binario ni servidor de la carpeta
// con la sesión que la cumple y la respuesta que, detrás de su línea, ofrece
// repetir la consulta: la oferta casa con una expresión de la lista, y la sesión
// pasa sin nada de la lista en su resultado.
func juzgarLaOfertaSinLaLista(t *testing.T, evals []Eval, lista ExpresionesProhibidas) {
	t.Helper()

	sinConsulta := *evalDe(t, evals, ficheroSinBinarioDeBoe)
	require.True(t, sinConsulta.SinBinarioNiServidor, "%s es la eval sin binario ni servidor", ficheroSinBinarioDeBoe)
	require.True(t, sinConsulta.Activa, "%s espera que la skill se active", ficheroSinBinarioDeBoe)
	require.Equal(t, []string{"respondo con el texto"}, ExtraerExpresionesProhibidas(ofertaDeConsultarDespues, lista),
		"la oferta de consultar después casa con una expresión de la lista")

	respuesta := lineaSinConsultaAlBOE + "\n\n" + ofertaDeConsultarDespues
	resultado := Juzgar(sinConsulta, sesionSinBinarioNiServidor(skillNormativa, respuesta), skillNormativa)

	assert.True(t, resultado.Pasa, "%s pasa con la oferta de consultar después: %v", ficheroSinBinarioDeBoe,
		resultado.Motivos)
	assert.Empty(t, deLaLista(t, resultado), "%s no lleva nada de la lista en su resultado", ficheroSinBinarioDeBoe)
}

// respuestaDelCalibrado es una respuesta de un informe del calibrado de H7.4 que
// la lista del repositorio marca, con el informe del que sale y la eval con la
// que se juzga.
type respuestaDelCalibrado struct {
	informe string
	entrada entradaDelInforme
	eval    Eval

	// enLaCarpeta dice si el fichero de su eval está en la carpeta de evals.
	enLaCarpeta bool
}

// juzgarElCalibradoSinLaLista juzga cada respuesta del calibrado de H7.4
// (respuestasDelCalibrado) con una sesión terminada hecha de su entrada —su
// respuesta tal cual y la skill activada si la entrada lo dice, sin
// invocaciones—: lo que falte de lo que su eval espera da sus motivos de
// siempre, y ningún resultado lleva nada de la lista.
func juzgarElCalibradoSinLaLista(t *testing.T, evals []Eval, lista ExpresionesProhibidas) {
	t.Helper()

	var marcadasPorLaLista []string

	for _, caso := range respuestasDelCalibrado(t, evals, lista) {
		sesion := sesionTerminada(caso.entrada.Activada, caso.entrada.Respuesta)

		for _, rastro := range deLaLista(t, Juzgar(caso.eval, sesion, skillNormativa)) {
			marcadasPorLaLista = append(marcadasPorLaLista, caso.informe+", "+caso.entrada.Sesion+": "+rastro)
		}
	}

	assert.Empty(t, marcadasPorLaLista, "sesiones del calibrado de H7.4 que Juzgar marca por la lista de %s:\n%s",
		evalsDeNormativa, strings.Join(marcadasPorLaLista, "\n"))
}

// respuestasDelCalibrado son las respuestas de los tres informes del calibrado
// de H7.4 que marca la lista, aplicada con ExtraerExpresionesProhibidas a la
// respuesta de cada entrada, en el orden de los informes y de sus entradas, con
// sus premisas: son las de marcadasDeCadaInforme, 56 en total; la eval de todas
// espera que la skill se active, porque la lista solo juzgaba esas; y el fichero
// de su eval está en la carpeta en todas menos en dos, las del informe de H7.1 y
// de la eval que retiró H7.2. Esas dos se juzgan con una eval que solo declara
// lo que su entrada dice de ella, su fichero y que activa la skill; las demás,
// con la de la carpeta.
func respuestasDelCalibrado(t *testing.T, evals []Eval, lista ExpresionesProhibidas) []respuestaDelCalibrado {
	t.Helper()

	var respuestas []respuestaDelCalibrado

	for _, calibrado := range marcadasDeCadaInforme {
		delInforme := 0

		for _, entrada := range entradasDelInforme(t, calibrado.informe) {
			if len(ExtraerExpresionesProhibidas(entrada.Respuesta, lista)) == 0 {
				continue
			}

			delInforme++

			respuestas = append(respuestas, respuestaDelCalibradoDe(t, evals, calibrado.informe, entrada))
		}

		require.Equal(t, calibrado.marcadas, delInforme, "respuestas del informe %s que marca la lista de %s",
			calibrado.informe, evalsDeNormativa)
	}

	require.Len(t, respuestas, marcadasEnElCalibrado, "respuestas de los tres informes que marca la lista de %s",
		evalsDeNormativa)

	fueraDeLaCarpeta := slices.DeleteFunc(slices.Clone(respuestas), func(respuesta respuestaDelCalibrado) bool {
		return respuesta.enLaCarpeta
	})
	require.Len(t, fueraDeLaCarpeta, marcadasDeLaEvalRetirada, "respuestas marcadas cuya eval no está en %s",
		evalsDeNormativa)

	for _, respuesta := range fueraDeLaCarpeta {
		require.Equal(t, informeDeLaEvalRetirada, respuesta.informe, "el informe de la eval retirada")
		require.Equal(t, cifrasDeLaEvalRetirada, respuesta.entrada.Eval[:cifrasDelFicheroDeUnaEval],
			"las dos cifras de la eval retirada")
	}

	return respuestas
}

// respuestaDelCalibradoDe es la respuesta marcada de la entrada del informe con
// su eval: la de la carpeta con ese fichero, como la deja LeerConjunto, o, si no
// está, una que solo declara lo que la entrada dice de ella. La entrada dice que
// su eval espera que la skill se active, y su fichero empieza por sus dos
// cifras.
func respuestaDelCalibradoDe(t *testing.T, evals []Eval, informe string, entrada entradaDelInforme) respuestaDelCalibrado {
	t.Helper()

	require.True(t, entrada.Activa, "la eval de %s, del informe %s, espera que la skill se active", entrada.Sesion,
		informe)
	require.Regexp(t, `^[0-9]{2}-`, entrada.Eval, "el fichero de la eval de %s, del informe %s, empieza por sus dos "+
		"cifras", entrada.Sesion, informe)

	respuesta := respuestaDelCalibrado{
		informe: informe,
		entrada: entrada,
		eval:    Eval{Fichero: entrada.Eval, Activa: entrada.Activa},
	}

	posicion := slices.IndexFunc(evals, func(eval Eval) bool { return eval.Fichero == entrada.Eval })
	if posicion >= 0 {
		respuesta.eval, respuesta.enLaCarpeta = evals[posicion], true
	}

	return respuesta
}

// deLaLista es lo que el resultado lleva de la lista de expresiones, una línea
// por cosa: la clave de las expresiones, si está en el resultado codificado como
// lo codifica EscribirInforme, y cada motivo que empieza como el de una
// expresión. Nil si no lleva nada.
func deLaLista(t *testing.T, resultado ResultadoDeEval) []string {
	t.Helper()

	codificado, err := json.Marshal(resultado)
	require.NoError(t, err)

	var claves map[string]jsontext.Value
	require.NoError(t, json.Unmarshal(codificado, &claves))

	var rastros []string

	if valor, conLaClave := claves[claveDeLasExpresiones]; conLaClave {
		rastros = append(rastros, "la clave "+claveDeLasExpresiones+", con "+string(valor))
	}

	for _, motivo := range resultado.Motivos {
		if strings.HasPrefix(motivo, principioDeExpresionProhibida) {
			rastros = append(rastros, "el motivo «"+motivo+"»")
		}
	}

	return rastros
}

// ofertaDeConsultarDespues es lo que la respuesta de una sesión sin binario ni
// servidor añade a su línea cuando ofrece repetir la consulta, como la escribió
// el modelo en el job.
const ofertaDeConsultarDespues = "Cuando kitlegal esté disponible, lo consulto y te respondo con el texto y su cita."

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
		{argv: dePlazosYTerritorio("resolver", "Leganés", "--json"), orden: ordenDeLeganes},
		{argv: dePlazosYTerritorio("resolver", "leganes", "--json"), orden: "territorio resolver leganes --json"},
		{argv: dePlazosYTerritorio("resolver", "LEGANÉS", "--json"), orden: "territorio resolver LEGANÉS --json"},
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
			invocada(t, codigoDeSalida(0), dePlazosYTerritorio("resolver", "Getafe", "--json")),
			invocada(t, codigoDeSalida(0), dePlazosYTerritorio("resolver", "Leganés", "--describe")),
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
			Modo:                 ModoOrden,
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
		Modo:                 ModoOrden,
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
// activó plazos-y-territorio y tiene la respuesta y las invocaciones dadas.
func sesionDeTerritorio(respuesta string, invocaciones ...Invocacion) Sesion {
	return cambiada(sesionTerminada(false, respuesta, invocaciones...), func(s *Sesion) {
		s.SkillsActivadas = []string{skillDeTerritorio}
	})
}

// resuelveLeganes es la invocación de la skill que resuelve el municipio de la
// eval del municipio cubierto y termina con código 0.
func resuelveLeganes(t *testing.T) Invocacion {
	t.Helper()

	return invocada(t, codigoDeSalida(0), dePlazosYTerritorio("resolver", "Leganés", "--json"))
}

// dePlazosYTerritorio es el argv con el que la skill instalada invoca el applet
// territorio.
func dePlazosYTerritorio(tokens ...string) []string {
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
			Modo:               ModoOrden,
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
			Modo:               ModoOrden,
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
			r.FueraDeLoGrabado = []InvocacionFallida{{Orden: "boe articulo BOE-A-2015-10565 a9998 --json", Codigo: codigoDeSalida(5)}}
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
			r.OtrasFallidas = []InvocacionFallida{{Orden: ordenDeLaComprobacionDeLaNorma, Codigo: codigoDeSalida(7)}}
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
				r.OtrasFallidas = []InvocacionFallida{{Orden: ordenDeGraphShow, Codigo: codigoDeSalida(3)}}
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
			r.OtrasFallidas = []InvocacionFallida{{Orden: otroApplet, Codigo: codigoDeSalida(2)}}
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
		Modo:               ModoOrden,
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

// evalDeLosDosBloques es la eval 20 de contracts/evals-y-juicio.md §3 de H7.4: el
// grafo previo con la redacción original del art. 118 y de la disposición
// adicional tercera de la LCSP, la lectura de los dos bloques y la comprobación
// de la memoria con su norma, graph show prohibido, la cita de los dos bloques, el
// hallazgo version-obsoleta y las dos redacciones modificadas.
func evalDeLosDosBloques() Eval {
	articulo118 := ComandoEsperado{Applet: "boe", Norma: normaDeLaLCSP, Bloque: bloqueDelArticulo118}
	da3 := ComandoEsperado{Applet: "boe", Norma: normaDeLaLCSP, Bloque: bloqueDeLaDA3}

	return Eval{
		Fichero: ficheroDeLosDosBloques,
		Pregunta: "Hace tiempo te pregunt\xc3\xa9 qu\xc3\xa9 exige la LCSP para el expediente de un contrato menor, " +
			"en su art\xc3\xadculo 118, y qu\xc3\xa9 a\xc3\xb1ade su disposici\xc3\xb3n adicional tercera para los " +
			"ayuntamientos. \xc2\xbfQu\xc3\xa9 dicen ahora?",
		Activa:      true,
		Informativa: true,
		GrafoPrevio: GrafoPrevio{
			Grabaciones: "lcsp-a1-30-y-da-3-redaccion-original", Comandos: []ComandoEsperado{articulo118, da3},
		},
		Comandos:   []ComandoEsperado{articulo118, da3, {Applet: "graph", Verbo: "check", Norma: normaDeLaLCSP}},
		Prohibidos: []ComandoProhibido{{Applet: "graph", Verbo: "show"}},
		Citas: []CitaEsperada{
			{Norma: normaDeLaLCSP, Bloque: bloqueDelArticulo118}, {Norma: normaDeLaLCSP, Bloque: bloqueDeLaDA3},
		},
		Hallazgos: []string{versionObsoleta},
		RedaccionesModificadas: []RedaccionEsperada{
			{
				Norma: normaDeLaLCSP, Bloque: bloqueDelArticulo118,
				FechaVigencia: fechaDeLaRedaccionOriginal, FechaVigenciaReciente: fechaDelArticulo118Vigente,
			},
			{
				Norma: normaDeLaLCSP, Bloque: bloqueDeLaDA3,
				FechaVigencia: fechaDeLaRedaccionOriginal, FechaVigenciaReciente: fechaDeLaDA3Vigente,
			},
		},
	}
}

// resultadoDeLosDosBloques es el resultado de la eval de los dos bloques con una
// sesión terminada y activada que lee los dos bloques en una orden, comprueba la
// memoria de la norma con los dos y responde con la respuesta dada, con las dos
// citas, la forma de version-obsoleta y las dos redacciones: pasa.
func resultadoDeLosDosBloques(respuesta string) ResultadoDeEval {
	return ResultadoDeEval{
		Eval:                   ficheroDeLosDosBloques,
		Modo:                   ModoOrden,
		Activa:                 true,
		Activada:               true,
		ComandosEjecutados:     []string{textoDelComando118, textoDelComandoDA3, textoDeLaComprobacionDeLaNorma},
		CitasEncontradas:       []string{textoDeLaCita118, textoDeLaCitaDA3},
		HallazgosEncontrados:   []string{versionObsoleta},
		RedaccionesEncontradas: []string{redaccionDelArticulo118, redaccionDeLaDA3},
		Invocaciones: []InvocacionInformada{
			{Orden: ordenDeLosDosBloques, Codigo: codigoDeSalida(0)},
			{Orden: ordenDeLaComprobacionDeLosDosBloques, Codigo: codigoDeSalida(0)},
		},
		Respuesta:        respuesta,
		CodigoDeLaSesion: codigoDeSalida(0),
		FinDeLaSesion:    "result success",
		SesionTerminada:  true,
		Pasa:             true,
	}
}

// conLasLineas es el juicio de la eval de los dos bloques con la sesión de
// resultadoDeLosDosBloques cuya respuesta son las líneas dadas, cada una en su
// párrafo, y detrás el cuerpo con las dos citas: su resultado esperado reparte
// las redacciones entre las encontradas y las ausentes dadas y, con alguna
// ausente, lleva su motivo por cada una y no pasa.
func conLasLineas(t *testing.T, lineas, encontradas, ausentes []string) juicio {
	t.Helper()

	respuesta := strings.Join(slices.Concat(lineas, []string{cuerpoConLasDosCitas}), "\n\n")

	return juicio{
		eval:   evalDeLosDosBloques(),
		sesion: sesionTerminada(true, respuesta, leeLosDosBloques(t), compruebaLosDosBloques(t)),
		esperado: cambiado(resultadoDeLosDosBloques(respuesta), func(r *ResultadoDeEval) {
			r.RedaccionesEncontradas, r.RedaccionesAusentes = encontradas, ausentes

			if len(ausentes) > 0 {
				r.Motivos = prefijados(principioDeRedaccionAusente, ausentes)
				r.Pasa = false
			}
		}),
	}
}

// losDosBloquesEnDosOrdenes es el juicio de la eval de los dos bloques con una
// sesión que lee cada bloque en su orden y responde con una línea por bloque:
// los dos comandos de bloque quedan ejecutados, como con una orden, y pasa.
func losDosBloquesEnDosOrdenes(t *testing.T) juicio {
	t.Helper()

	const ordenDeLaDA3 = "boe articulo BOE-A-2017-12902 da-3 --json"

	respuesta := lineaConLaCitaDel118() + "\n\n" + lineaConLaCitaDeLaDA3() + "\n\n" + cuerpoConLasDosCitas

	return juicio{
		eval: evalDeLosDosBloques(),
		sesion: sesionTerminada(true, respuesta,
			leeElArticulo118(t),
			invocada(t, codigoDeSalida(0), deLaSkill(strings.Fields(ordenDeLaDA3)[1:]...)),
			compruebaLosDosBloques(t)),
		esperado: cambiado(resultadoDeLosDosBloques(respuesta), func(r *ResultadoDeEval) {
			r.Invocaciones = []InvocacionInformada{
				{Orden: ordenDelArticulo118, Codigo: codigoDeSalida(0)},
				{Orden: ordenDeLaDA3, Codigo: codigoDeSalida(0)},
				{Orden: ordenDeLaComprobacionDeLosDosBloques, Codigo: codigoDeSalida(0)},
			}
		}),
	}
}

// redaccionesDetrasDelHallazgo es el juicio de la eval de los dos bloques con la
// sesión de conLasLineas cuya respuesta anuncia que ya puede responder, sin
// ninguna línea: los motivos van en su orden, el del hallazgo ausente y el de
// cada redacción ausente, en el orden de la eval, y ninguno por el anuncio, que
// lleva una expresión de la lista de normativa.
func redaccionesDetrasDelHallazgo(t *testing.T) juicio {
	t.Helper()

	caso := conLasLineas(t, []string{elAnuncioDeLaRespuesta}, nil, nil)
	caso.esperado = cambiado(caso.esperado, func(r *ResultadoDeEval) {
		r.HallazgosEncontrados = nil
		r.HallazgosAusentes = []string{versionObsoleta}
		r.RedaccionesAusentes = []string{redaccionDelArticulo118, redaccionDeLaDA3}
		r.Motivos = []string{
			motivoDeVersionObsoleta,
			principioDeRedaccionAusente + redaccionDelArticulo118,
			principioDeRedaccionAusente + redaccionDeLaDA3,
		}
		r.Pasa = false
	})

	return caso
}

// lineaDeRedaccion es la línea ⚠ REDACCIÓN MODIFICADA: de research D4 de H7.4 con
// la cita dada detrás de la etiqueta y las dos fechas dadas; sin cita, si es
// vacía, la línea de H7.3.
func lineaDeRedaccion(cita, superada, leida string) string {
	etiqueta := "\xe2\x9a\xa0 REDACCI\xc3\x93N MODIFICADA: "
	if cita != "" {
		etiqueta += cita + ": "
	}

	return etiqueta + "la redacci\xc3\xb3n con fecha de vigencia " + superada + ", la que se consult\xc3\xb3 antes, " +
		"ha sido sustituida por la de " + leida + ", que es la que se cita."
}

// lineaConLaCitaDel118 y lineaConLaCitaDeLaDA3 son las líneas ⚠ REDACCIÓN
// MODIFICADA: de cada bloque de la eval de los dos bloques, con su cita y sus dos
// fechas.
func lineaConLaCitaDel118() string {
	return lineaDeRedaccion(citaDelArticulo118, fechaDeLaRedaccionOriginal, fechaDelArticulo118Vigente)
}

func lineaConLaCitaDeLaDA3() string {
	return lineaDeRedaccion(citaDeLaDA3, fechaDeLaRedaccionOriginal, fechaDeLaDA3Vigente)
}

// leeLosDosBloques es la invocación de la skill que lee en una orden los dos
// bloques de la eval de los dos bloques y termina con código 0, y
// compruebaLosDosBloques, la que comprueba con los dos la memoria de su norma.
func leeLosDosBloques(t *testing.T) Invocacion {
	t.Helper()

	return invocada(t, codigoDeSalida(0), deLaSkill(strings.Fields(ordenDeLosDosBloques)[1:]...))
}

func compruebaLosDosBloques(t *testing.T) Invocacion {
	t.Helper()

	return invocada(t, codigoDeSalida(0), deKitlegal(strings.Fields(ordenDeLaComprobacionDeLosDosBloques)...))
}

// evalDePlazosYTerritorioSinNormativa es la eval del municipio cubierto que
// declara, como las de plazos-y-territorio del repositorio, que normativa no se
// activa.
func evalDePlazosYTerritorioSinNormativa() Eval {
	eval := evalDelMunicipioCubierto()
	eval.NoSeActivan = []string{skillQueNoSeActiva}

	return eval
}

// seActivaLaQueNoSeActiva son los juicios de las evals de plazos-y-territorio
// que declaran que normativa no se activa: la del municipio cubierto, con la
// sesión que lo resuelve, declara su territorio y activa plazos-y-territorio y
// normativa, no pasa, con el motivo que nombra normativa, y con la misma sin
// activar normativa, pasa; y la de no activación, con una sesión que activa
// solo normativa —la activación de plazos-y-territorio coincide—, no pasa, con
// ese motivo.
func seActivaLaQueNoSeActiva(t *testing.T) []juicio {
	t.Helper()

	sinActivarla := sesionDeTerritorio(respuestaDelMunicipio, resuelveLeganes(t))
	activandola := cambiada(sinActivarla, func(s *Sesion) {
		s.SkillsActivadas = []string{skillDeTerritorio, skillQueNoSeActiva}
	})

	deNoActivacion := evalDeNoActivacion()
	deNoActivacion.NoSeActivan = []string{skillQueNoSeActiva}

	return []juicio{
		{
			eval:   evalDePlazosYTerritorioSinNormativa(),
			sesion: activandola,
			skill:  skillDeTerritorio,
			esperado: cambiado(resultadoDelMunicipioQuePasa(), func(r *ResultadoDeEval) {
				r.Motivos = []string{motivoDeNormativaActivada}
				r.Pasa = false
			}),
		},
		{
			eval:     evalDePlazosYTerritorioSinNormativa(),
			sesion:   sinActivarla,
			skill:    skillDeTerritorio,
			esperado: resultadoDelMunicipioQuePasa(),
		},
		{
			eval:   deNoActivacion,
			sesion: sesionTerminada(true, respuestaSinSkill),
			skill:  skillDeTerritorio,
			esperado: ResultadoDeEval{
				Eval:             ficheroDeNoActivacion,
				Modo:             ModoOrden,
				Respuesta:        respuestaSinSkill,
				CodigoDeLaSesion: codigoDeSalida(0),
				FinDeLaSesion:    "result success",
				SesionTerminada:  true,
				Motivos:          []string{motivoDeNormativaActivada},
			},
		},
	}
}

// laQueNoSeActivaDetrasDeLaActivacion es el juicio de la eval del municipio
// cubierto que declara que normativa no se activa con una sesión que activa
// solo normativa, no resuelve el municipio y declara su territorio: los
// motivos van en su orden, el de la activación de plazos-y-territorio, el de
// normativa y el del comando ausente.
func laQueNoSeActivaDetrasDeLaActivacion(t *testing.T) juicio {
	t.Helper()

	return juicio{
		eval:   evalDePlazosYTerritorioSinNormativa(),
		sesion: sesionTerminada(true, respuestaDelMunicipio),
		skill:  skillDeTerritorio,
		esperado: cambiado(resultadoDelMunicipioQuePasa(), func(r *ResultadoDeEval) {
			r.Activada = false
			r.ComandosEjecutados = nil
			r.ComandosAusentes = []string{textoDelComandoDeLeganes}
			r.Invocaciones = nil
			r.Motivos = []string{
				"la activaci\xc3\xb3n no coincide: se esperaba que la skill plazos-y-territorio se activara y no se activ\xc3\xb3",
				motivoDeNormativaActivada,
				"comando ausente: " + textoDelComandoDeLeganes,
			}
			r.Pasa = false
		}),
	}
}

// evalDelMunicipioQueActivaDos es la eval del municipio cubierto que declara
// que, con plazos-y-territorio, se activan normativa y jurisprudencia, en ese
// orden, como la 02 de consulta-juridica declara la skill de su capacidad
// (contracts/evals.md §2 de H32).
func evalDelMunicipioQueActivaDos() Eval {
	eval := evalDelMunicipioCubierto()
	eval.SeActivan = []string{skillNormativa, skillDeJurisprudencia}

	return eval
}

// seActivanLasQueSeActivan son los juicios de la eval del municipio cubierto que
// declara que se activan normativa y jurisprudencia, con la sesión que lo
// resuelve y declara su territorio: activando solo plazos-y-territorio, no
// pasa, con un motivo por cada una, en el orden de la eval; activando además
// normativa, no pasa, con el motivo que nombra jurisprudencia; y activándolas
// todas, en otro orden que el de la eval, pasa.
func seActivanLasQueSeActivan(t *testing.T) []juicio {
	t.Helper()

	sinActivarlas := sesionDeTerritorio(respuestaDelMunicipio, resuelveLeganes(t))
	activando := func(skills ...string) Sesion {
		return cambiada(sinActivarlas, func(s *Sesion) {
			s.SkillsActivadas = append([]string{skillDeTerritorio}, skills...)
		})
	}
	queNoPasa := func(motivos ...string) ResultadoDeEval {
		return cambiado(resultadoDelMunicipioQuePasa(), func(r *ResultadoDeEval) {
			r.Motivos = motivos
			r.Pasa = false
		})
	}

	return []juicio{
		{
			eval:     evalDelMunicipioQueActivaDos(),
			sesion:   sinActivarlas,
			skill:    skillDeTerritorio,
			esperado: queNoPasa(motivoDeNormativaSinActivar, motivoDeJurisprudenciaSinActivar),
		},
		{
			eval:     evalDelMunicipioQueActivaDos(),
			sesion:   activando(skillNormativa),
			skill:    skillDeTerritorio,
			esperado: queNoPasa(motivoDeJurisprudenciaSinActivar),
		},
		{
			eval:     evalDelMunicipioQueActivaDos(),
			sesion:   activando(skillDeJurisprudencia, skillNormativa),
			skill:    skillDeTerritorio,
			esperado: resultadoDelMunicipioQuePasa(),
		},
	}
}

// lasQueSeActivanDelanteDeLasQueNo es el juicio de la eval del municipio
// cubierto que declara que normativa se activa y jurisprudencia no, con una
// sesión que activa solo jurisprudencia, no resuelve el municipio y declara su
// territorio: los motivos van en su orden, el de la activación de
// plazos-y-territorio, el de normativa sin activar, el de jurisprudencia
// activada y el del comando ausente.
func lasQueSeActivanDelanteDeLasQueNo() juicio {
	eval := evalDelMunicipioCubierto()
	eval.SeActivan = []string{skillNormativa}
	eval.NoSeActivan = []string{skillDeJurisprudencia}

	return juicio{
		eval: eval,
		sesion: cambiada(sesionTerminada(false, respuestaDelMunicipio), func(s *Sesion) {
			s.SkillsActivadas = []string{skillDeJurisprudencia}
		}),
		skill: skillDeTerritorio,
		esperado: cambiado(resultadoDelMunicipioQuePasa(), func(r *ResultadoDeEval) {
			r.Activada = false
			r.ComandosEjecutados = nil
			r.ComandosAusentes = []string{textoDelComandoDeLeganes}
			r.Invocaciones = nil
			r.Motivos = []string{
				"la activación no coincide: se esperaba que la skill plazos-y-territorio se activara y no se activó",
				motivoDeNormativaSinActivar,
				motivoDeJurisprudenciaActivada,
				"comando ausente: " + textoDelComandoDeLeganes,
			}
			r.Pasa = false
		}),
	}
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
		sesion.SkillsActivadas = []string{skillNormativa}
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
		Modo:               ModoOrden,
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
			r.FueraDeLoGrabado = []InvocacionFallida{{Orden: "boe " + fuera, Codigo: codigoDeSalida(4)}}
		}),
	}
}

// TestExigirElModeloPedido comprueba que una sesión pasa el control del modelo si
// declara el id pedido, tal cual o con la fecha de su versión detrás, o si no
// declaró ninguno; y que un id más largo que empieza por el pedido es otro modelo
// —claude-sonnet-5-5 no es claude-sonnet-5 (ADR 0031)—, que no pasa y da su
// motivo.
func TestExigirElModeloPedido(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre, pedido, declarado string
		pasa                      bool
	}{
		{nombre: "el-mismo-id", pedido: "claude-sonnet-5-5", declarado: "claude-sonnet-5-5", pasa: true},
		{nombre: "con-la-fecha", pedido: "claude-haiku-4-5", declarado: "claude-haiku-4-5-20251001", pasa: true},
		{nombre: "sin-declarar", pedido: "claude-sonnet-5-5", declarado: "", pasa: true},
		{nombre: "otra-version-del-mismo-nombre", pedido: "claude-sonnet-5", declarado: "claude-sonnet-5-5"},
		{nombre: "la-anterior", pedido: "claude-sonnet-5-5", declarado: "claude-sonnet-5"},
		{nombre: "fecha-corta", pedido: "claude-haiku-4-5", declarado: "claude-haiku-4-5-2025100"},
		{nombre: "fecha-no-numerica", pedido: "claude-haiku-4-5", declarado: "claude-haiku-4-5-2025100a"},
		{nombre: "fecha-sin-guion", pedido: "claude-haiku-4-5", declarado: "claude-haiku-4-520251001"},
		{nombre: "otro-modelo", pedido: "claude-sonnet-5-5", declarado: "claude-opus-5-5"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			resultado := ResultadoDeEval{Modelo: caso.pedido, ModeloDeLaSesion: caso.declarado, Pasa: true}
			resultado.exigirElModeloPedido()

			assert.Equal(t, caso.pasa, resultado.Pasa)
			if caso.pasa {
				assert.Empty(t, resultado.Motivos)

				return
			}

			assert.Equal(t, []string{motivoDeOtroModelo + caso.declarado + ", y se pidió " + caso.pedido},
				resultado.Motivos)
		})
	}
}

// Lo que juzgan TestJuzgarLasLlamadas y TestJuzgarSinBinarioNiServidor, con
// sesiones construidas en memoria (contracts/evals-en-dos-modos.md §4 de H21).
const (
	// kitlegalDelServidor es la ruta absoluta del binario que servidor.json
	// declara en una sesión del modo herramienta: con ella arranca el agente el
	// servidor, y con ella puede una sesión ejecutar una orden sin el PATH
	// (research.md D18 de H21).
	kitlegalDelServidor = "/home/runner/go/bin/kitlegal"

	// ordenDelServidor es la orden de la invocación de la traza del proceso del
	// servidor, como la presenta el informe.
	ordenDelServidor = "mcp serve"

	// herramientaDeArticulo es la herramienta que lee un bloque, y
	// llamadaDelArticulo21, la orden de la llamada que lee el de la eval 01 como
	// la presenta el informe: la herramienta seguida de sus argumentos.
	herramientaDeArticulo = "boe_articulo"
	llamadaDelArticulo21  = "boe_articulo BOE-A-2015-10565 a21"

	// principioDeOrdenSinKitlegal es el principio del motivo de una orden de
	// kitlegal en una sesión que no lo tiene en el PATH, al que sigue la orden,
	// escrito a mano.
	principioDeOrdenSinKitlegal = "orden de kitlegal en una sesión sin kitlegal en el PATH: "

	// sobreCorrectoDeLaLlamada y sobreDeNoEncontrado son el sobre de éxito y el de
	// fallo que el servidor devuelve como resultado de una llamada, reducidos a lo
	// que se lee de ellos.
	sobreCorrectoDeLaLlamada = `{"ok":true,"fuente":"boe.legislacion-consolidada","data":{"bloque":"a21"}}`
	sobreDeNoEncontrado      = `{"ok":false,"fuente":"boe.legislacion-consolidada",` +
		`"data":{"clase":"no-encontrado","mensaje":"el bloque no existe"}}`
)

// Lo que juzga TestJuzgarSinBinarioNiServidor: los ficheros de las dos evals sin
// binario ni servidor, la línea con su forma fija y los tres motivos de
// contracts/evals-en-dos-modos.md §4 de H21, escritos a mano.
const (
	ficheroSinBinarioDeBoe               = "21-sin-binario-ni-servidor.yaml"
	ficheroSinBinarioDePlazosYTerritorio = "04-sin-binario-ni-servidor.yaml"

	// lineaSinConsultaAlBOE es la línea de contracts/skills.md §3 de H21 con una
	// causa en lugar del marcador y la dirección en la misma línea, y
	// lineaSinSuDireccion, la misma sin la dirección.
	lineaSinSuDireccion   = "⚠ SIN CONSULTA AL BOE: este agente no tiene la herramienta ni el binario de kitlegal."
	lineaSinConsultaAlBOE = lineaSinSuDireccion +
		" Para consultarlo hace falta instalar kitlegal: https://kitlegal.es/instalar/"

	motivoDeLaLineaAusente      = "línea ⚠ SIN CONSULTA AL BOE: ausente"
	motivoDeLaLineaSinDireccion = "la línea ⚠ SIN CONSULTA AL BOE: no lleva https://kitlegal.es/instalar/"
	principioDeCitaSinConsulta  = "cita en una respuesta sin consulta: "

	// citaDelArticulo53 es una cita con la forma fija, la del artículo por el que
	// pregunta la eval de normativa, y textoDeLaCita53, su texto en el
	// informe y en los motivos.
	citaDelArticulo53 = "[BOE-A-2015-10565, bloque a53]"
	textoDeLaCita53   = "BOE-A-2015-10565 a53"
)

// juicioDeModo es un juicio de una sesión del modo dado, con juzgarEnModo.
type juicioDeModo struct {
	juicio

	modo Modo
}

// TestJuzgarLasLlamadas fija el juicio de las llamadas a las herramientas del
// servidor MCP, con sesiones construidas en memoria
// (contracts/evals-en-dos-modos.md §4 de H21; research.md D17 y D18 de H21;
// FR-041, FR-042, FR-081; SC-012; US5-3, US5-4): una llamada cuenta como una
// invocación de su applet y su verbo, con los argumentos de la orden equivalente,
// de modo que el comando de bloque lo satisface boe_articulo —o boe_articulos—
// con su norma y su bloque, y no con otra norma ni con otro bloque, ni con la
// norma y el bloque de la eval si su resultado es un error o si no tiene
// resultado; el nombre con el prefijo del agente es la misma herramienta; el
// prohibido lo ejecuta graph_show con cualquier resultado; y los comandos de
// términos y de municipio los satisfacen boe_buscar y territorio_resolver. Cada
// llamada se informa con las invocaciones, detrás de las de la traza, con su
// herramienta y sus argumentos como orden y con el código de la orden
// equivalente: 0, el de la clase de su error —4 y 5, fuera de lo grabado; los
// demás, otras fallidas; 1 si la clase no se lee— o ninguno sin resultado.
//
// El proceso del servidor, mcp serve en la traza, no es una consulta: no
// satisface ni falla nada, termine como termine, y sus conexiones se informan con
// él. Y en una sesión del modo herramienta, una invocación de la traza de otro
// applet no pasa, con un motivo que nombra la orden, aunque satisfaga el comando;
// la misma sesión en el modo orden pasa.
func TestJuzgarLasLlamadas(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre  string
		juicios []juicioDeModo
	}{
		{nombre: "bloque-satisfecho", juicios: llamadasQueLeenElBloque(t)},
		{nombre: "otra-norma-u-otro-bloque", juicios: llamadasQueNoLeenElBloque(t)},
		{nombre: "resultado-de-error", juicios: llamadasConResultadoDeError(t)},
		{nombre: "sin-resultado", juicios: []juicioDeModo{llamadaSinResultado(t)}},
		{nombre: "prefijo-del-agente", juicios: llamadasConElPrefijoDelAgente(t)},
		{nombre: "prohibido-por-herramienta", juicios: prohibidoPorHerramienta(t)},
		{nombre: "terminos-por-boe-buscar", juicios: terminosPorHerramienta(t)},
		{nombre: "municipio-por-territorio-resolver", juicios: municipioPorHerramienta(t)},
		{nombre: "orden-en-el-modo-herramienta", juicios: ordenEnElModoHerramienta(t)},
		{nombre: "el-servidor-no-es-una-consulta", juicios: elServidorNoEsUnaConsulta(t)},
		{nombre: "argumentos-que-no-convierten", juicios: llamadasQueNoConvierten(t)},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			for posicion, j := range caso.juicios {
				assert.Equal(t, j.esperado, juzgarEnModo(j.eval, j.sesion, cmp.Or(j.skill, skillNormativa), j.modo),
					"el juicio %d: la eval %s con la sesión del caso", posicion, j.eval.Fichero)
			}
		})
	}
}

// llamadasQueLeenElBloque son los juicios de la eval 01 con una sesión del modo
// herramienta que lee su bloque con una llamada correcta y lo cita: con
// boe_articulo, con sus argumentos en el otro orden —la orden equivalente los
// lleva en el de la orden, no en el del objeto— y con boe_articulos con el bloque
// entre los que pide. En todos, el comando queda ejecutado y la eval pasa.
func llamadasQueLeenElBloque(t *testing.T) []juicioDeModo {
	t.Helper()

	const (
		enOtroOrden   = `{"bloque":"a21","norma":"BOE-A-2015-10565"}`
		variosBloques = `{"norma":"BOE-A-2015-10565","bloques":["a24","a21"]}`
	)

	return []juicioDeModo{
		{modo: ModoHerramienta, juicio: juicio{
			eval: evalDelArticulo21(),
			sesion: sesionDelModoHerramienta(t, respuestaConCita,
				llamadaCorrecta(herramientaDeArticulo, argumentosDelArticulo21)),
			esperado: resultadoDeLasLlamadas(llamadaInformada(llamadaDelArticulo21, codigoDeSalida(0))),
		}},
		{modo: ModoHerramienta, juicio: juicio{
			eval:     evalDelArticulo21(),
			sesion:   sesionDelModoHerramienta(t, respuestaConCita, llamadaCorrecta(herramientaDeArticulo, enOtroOrden)),
			esperado: resultadoDeLasLlamadas(llamadaInformada(llamadaDelArticulo21, codigoDeSalida(0))),
		}},
		{modo: ModoHerramienta, juicio: juicio{
			eval:   evalDelArticulo21(),
			sesion: sesionDelModoHerramienta(t, respuestaConCita, llamadaCorrecta("boe_articulos", variosBloques)),
			esperado: resultadoDeLasLlamadas(
				llamadaInformada("boe_articulos BOE-A-2015-10565 a24 a21", codigoDeSalida(0))),
		}},
	}
}

// llamadasQueNoLeenElBloque son los juicios de la eval 01 con una sesión del modo
// herramienta cuya llamada correcta a boe_articulo es de otra norma, de otro
// bloque o lleva la norma y el bloque cambiados de sitio: el comando queda
// ausente y la eval no pasa, sin que ninguna vaya a las fallidas.
func llamadasQueNoLeenElBloque(t *testing.T) []juicioDeModo {
	t.Helper()

	dadas := []struct {
		argumentos string
		orden      string
	}{
		{argumentos: `{"norma":"BOE-A-2017-12902","bloque":"a21"}`, orden: "boe_articulo BOE-A-2017-12902 a21"},
		{argumentos: argumentosDelArticulo24, orden: "boe_articulo BOE-A-2015-10565 a24"},
		{argumentos: `{"norma":"a21","bloque":"BOE-A-2015-10565"}`, orden: "boe_articulo a21 BOE-A-2015-10565"},
	}

	juicios := make([]juicioDeModo, 0, len(dadas))

	for _, dada := range dadas {
		juicios = append(juicios, juicioDeModo{modo: ModoHerramienta, juicio: juicio{
			eval: evalDelArticulo21(),
			sesion: sesionDelModoHerramienta(t, respuestaConCita,
				llamadaCorrecta(herramientaDeArticulo, dada.argumentos)),
			esperado: cambiado(resultadoDeLasLlamadas(llamadaInformada(dada.orden, codigoDeSalida(0))),
				sinElBloqueDeLaEval01),
		}})
	}

	return juicios
}

// llamadasConResultadoDeError son los juicios de la eval 01 con una sesión del
// modo herramienta que llama a boe_articulo con la norma y el bloque de la eval
// y recibe un error, uno por cada clase del sobre de fallo y dos más con una
// clase que no se lee, vacía o de fuera del vocabulario: el comando queda
// ausente y la eval no pasa, y la llamada va, con el código de su clase, a fuera
// de lo grabado con fuente-no-disponible y con limite-o-tos, y a las otras
// fallidas con las demás, con el 1 si la clase no se lee.
func llamadasConResultadoDeError(t *testing.T) []juicioDeModo {
	t.Helper()

	errores := []struct {
		clase  schema.Clase
		codigo int
		fuera  bool
	}{
		{clase: schema.ClaseArgumentos, codigo: 2},
		{clase: schema.ClaseNoEncontrado, codigo: 3},
		{clase: schema.ClaseFuenteNoDisponible, codigo: 4, fuera: true},
		{clase: schema.ClaseLimiteOTos, codigo: 5, fuera: true},
		{clase: schema.ClaseIdentidadHumana, codigo: 6},
		{clase: schema.ClaseConflicto, codigo: 7},
		{clase: schema.ClaseInesperado, codigo: 1},
		{clase: "", codigo: 1},
		{clase: "otra-clase", codigo: 1},
	}

	juicios := make([]juicioDeModo, 0, len(errores))

	for _, fallo := range errores {
		fallida := []InvocacionFallida{{Orden: llamadaDelArticulo21, Codigo: codigoDeSalida(fallo.codigo)}}

		juicios = append(juicios, juicioDeModo{modo: ModoHerramienta, juicio: juicio{
			eval: evalDelArticulo21(),
			sesion: sesionDelModoHerramienta(t, respuestaConCita,
				llamadaFallida(herramientaDeArticulo, argumentosDelArticulo21, fallo.clase)),
			esperado: cambiado(resultadoDeLasLlamadas(llamadaInformada(llamadaDelArticulo21, codigoDeSalida(fallo.codigo))),
				func(r *ResultadoDeEval) {
					sinElBloqueDeLaEval01(r)

					if fallo.fuera {
						r.FueraDeLoGrabado = fallida
					} else {
						r.OtrasFallidas = fallida
					}
				}),
		}})
	}

	return juicios
}

// llamadaSinResultado es el juicio de la eval 01 con una sesión del modo
// herramienta cuya llamada a boe_articulo, con la norma y el bloque de la eval,
// no tiene resultado: queda sin código, no satisface el comando y no va a
// ninguna de las fallidas.
func llamadaSinResultado(t *testing.T) juicioDeModo {
	t.Helper()

	return juicioDeModo{modo: ModoHerramienta, juicio: juicio{
		eval: evalDelArticulo21(),
		sesion: sesionDelModoHerramienta(t, respuestaConCita,
			Llamada{Herramienta: herramientaDeArticulo, Argumentos: []byte(argumentosDelArticulo21)}),
		esperado: cambiado(resultadoDeLasLlamadas(llamadaInformada(llamadaDelArticulo21, nil)), sinElBloqueDeLaEval01),
	}}
}

// llamadasConElPrefijoDelAgente son los juicios de la eval 01 con una sesión del
// modo herramienta cuyas llamadas son las que LeerSesion lee de un transcript en
// el que el agente nombra la herramienta con su prefijo,
// mcp__kitlegal__boe_articulo: con el sobre de éxito, satisface el comando y la
// eval pasa; con el de fallo, sin is_error, no lo satisface.
func llamadasConElPrefijoDelAgente(t *testing.T) []juicioDeModo {
	t.Helper()

	conPrefijo := func(sobre string) []Llamada {
		return llamadasDelTranscript(t,
			mensajeDeLlamadas(t, usoDeHerramienta{
				id: "toolu_01", nombre: "mcp__kitlegal__boe_articulo", entrada: argumentosDelArticulo21,
			})+mensajeDeResultados(t, resultadoDeHerramienta{id: "toolu_01", contenido: cadenaJSON(t, sobre)}))
	}

	return []juicioDeModo{
		{modo: ModoHerramienta, juicio: juicio{
			eval:     evalDelArticulo21(),
			sesion:   sesionDelModoHerramienta(t, respuestaConCita, conPrefijo(sobreCorrectoDeLaLlamada)...),
			esperado: resultadoDeLasLlamadas(llamadaInformada(llamadaDelArticulo21, codigoDeSalida(0))),
		}},
		{modo: ModoHerramienta, juicio: juicio{
			eval:   evalDelArticulo21(),
			sesion: sesionDelModoHerramienta(t, respuestaConCita, conPrefijo(sobreDeNoEncontrado)...),
			esperado: cambiado(resultadoDeLasLlamadas(llamadaInformada(llamadaDelArticulo21, codigoDeSalida(3))),
				func(r *ResultadoDeEval) {
					sinElBloqueDeLaEval01(r)

					r.OtrasFallidas = []InvocacionFallida{{Orden: llamadaDelArticulo21, Codigo: codigoDeSalida(3)}}
				}),
		}},
	}
}

// prohibidoPorHerramienta son los juicios de la eval de la consulta repetida,
// que prohíbe graph show, con una sesión del modo herramienta que comprueba la
// memoria de la norma, lee el bloque y lo cita con llamadas correctas: sin más,
// pasa; y con una llamada a graph_show, con un resultado correcto, con un error
// o sin resultado, el prohibido queda ejecutado con su motivo y la eval no pasa.
func prohibidoPorHerramienta(t *testing.T) []juicioDeModo {
	t.Helper()

	const (
		laFicha        = `{"id":"eli/es/l/2015/10/01/39#a21"}`
		llamadaDeFicha = "graph_show eli/es/l/2015/10/01/39#a21"
	)

	lasEsperadas := []Llamada{
		llamadaCorrecta("graph_check", `{"norma":"BOE-A-2017-12902","bloques":["a1-30"]}`),
		llamadaCorrecta(herramientaDeArticulo, `{"norma":"BOE-A-2017-12902","bloque":"a1-30"}`),
	}

	resultado := func(deLaFicha ...InvocacionInformada) ResultadoDeEval {
		informadas := slices.Concat([]InvocacionInformada{
			{Orden: ordenDelServidor, Codigo: codigoDeSalida(0)},
			llamadaInformada("graph_check BOE-A-2017-12902 a1-30", codigoDeSalida(0)),
			llamadaInformada("boe_articulo BOE-A-2017-12902 a1-30", codigoDeSalida(0)),
		}, deLaFicha)

		return cambiado(resultadoDeLaConsultaRepetida(informadas...), func(r *ResultadoDeEval) {
			r.Modo = ModoHerramienta
		})
	}

	ejecutado := func(r *ResultadoDeEval) {
		r.ComandosProhibidosEjecutados = []string{textoDeGraphShow}
		r.Motivos = []string{"comando prohibido ejecutado: " + textoDeGraphShow}
		r.Pasa = false
	}

	conLaFicha := func(ficha Llamada, esperado ResultadoDeEval) juicioDeModo {
		return juicioDeModo{modo: ModoHerramienta, juicio: juicio{
			eval:     evalDeLaConsultaRepetida(),
			sesion:   sesionDelModoHerramienta(t, respuestaConLaCita118, slices.Concat(lasEsperadas, []Llamada{ficha})...),
			esperado: esperado,
		}}
	}

	return []juicioDeModo{
		{modo: ModoHerramienta, juicio: juicio{
			eval:     evalDeLaConsultaRepetida(),
			sesion:   sesionDelModoHerramienta(t, respuestaConLaCita118, lasEsperadas...),
			esperado: resultado(),
		}},
		conLaFicha(llamadaCorrecta("graph_show", laFicha),
			cambiado(resultado(llamadaInformada(llamadaDeFicha, codigoDeSalida(0))), ejecutado)),
		conLaFicha(llamadaFallida("graph_show", laFicha, schema.ClaseNoEncontrado),
			cambiado(resultado(llamadaInformada(llamadaDeFicha, codigoDeSalida(3))), func(r *ResultadoDeEval) {
				r.OtrasFallidas = []InvocacionFallida{{Orden: llamadaDeFicha, Codigo: codigoDeSalida(3)}}
				ejecutado(r)
			})),
		conLaFicha(Llamada{Herramienta: "graph_show", Argumentos: []byte(laFicha)},
			cambiado(resultado(llamadaInformada(llamadaDeFicha, nil)), ejecutado)),
	}
}

// terminosPorHerramienta son los juicios de una eval con un comando de términos
// y una sesión del modo herramienta que busca con boe_buscar: con cada término
// como palabra de sus argumentos, el comando queda ejecutado y la eval pasa; con
// otros, queda ausente.
func terminosPorHerramienta(t *testing.T) []juicioDeModo {
	t.Helper()

	const (
		comandoDeTerminos = "boe buscar Régimen local 7/1985"
		respuesta         = "Es la Ley 7/1985, reguladora de las Bases del Régimen Local [BOE-A-1985-5392, bloque a1]."
	)

	eval := Eval{
		Fichero:  "03-lrbrl-busqueda.yaml",
		Pregunta: "¿Qué ley regula las bases del régimen local?",
		Activa:   true,
		Comandos: []ComandoEsperado{{Applet: "boe", Verbo: "buscar", Terminos: []string{"Régimen", "local", "7/1985"}}},
		Citas:    []CitaEsperada{{Norma: normaDeLaLRBRL, Bloque: "a1"}},
	}

	resultado := func(orden string) ResultadoDeEval {
		return cambiado(resultadoDeLasLlamadas(llamadaInformada(orden, codigoDeSalida(0))), func(r *ResultadoDeEval) {
			r.Eval = eval.Fichero
			r.ComandosEjecutados = []string{comandoDeTerminos}
			r.CitasEncontradas = []string{"BOE-A-1985-5392 a1"}
			r.Respuesta = respuesta
		})
	}

	return []juicioDeModo{
		{modo: ModoHerramienta, juicio: juicio{
			eval: eval,
			sesion: sesionDelModoHerramienta(t, respuesta,
				llamadaCorrecta("boe_buscar", `{"texto":["bases del RÉGIMEN local","7/1985"]}`)),
			esperado: resultado("boe_buscar bases del RÉGIMEN local 7/1985"),
		}},
		{modo: ModoHerramienta, juicio: juicio{
			eval: eval,
			sesion: sesionDelModoHerramienta(t, respuesta,
				llamadaCorrecta("boe_buscar", `{"texto":["régimen localmente","7/1985"]}`)),
			esperado: cambiado(resultado("boe_buscar régimen localmente 7/1985"), func(r *ResultadoDeEval) {
				r.ComandosEjecutados = nil
				r.ComandosAusentes = []string{comandoDeTerminos}
				r.Motivos = []string{"comando ausente: " + comandoDeTerminos}
				r.Pasa = false
			}),
		}},
	}
}

// municipioPorHerramienta son los juicios de la eval del municipio cubierto con
// una sesión del modo herramienta de plazos-y-territorio que lo resuelve con
// territorio_resolver y declara su territorio: con el municipio, escrito de otra
// forma que plegada es la suya, el comando queda ejecutado y la eval pasa; con
// su código INE, queda ausente.
func municipioPorHerramienta(t *testing.T) []juicioDeModo {
	t.Helper()

	sesion := func(consulta string) Sesion {
		return cambiada(sesionDelModoHerramienta(t, respuestaDelMunicipio,
			llamadaCorrecta("territorio_resolver", `{"consulta":`+cadenaJSON(t, consulta)+`}`)), func(s *Sesion) {
			s.SkillsActivadas = []string{skillDeTerritorio}
		})
	}

	resultado := func(consulta string) ResultadoDeEval {
		return cambiado(resultadoDelMunicipioQuePasa(), func(r *ResultadoDeEval) {
			r.Modo = ModoHerramienta
			r.Invocaciones = []InvocacionInformada{
				{Orden: ordenDelServidor, Codigo: codigoDeSalida(0)},
				llamadaInformada("territorio_resolver "+consulta, codigoDeSalida(0)),
			}
		})
	}

	return []juicioDeModo{
		{modo: ModoHerramienta, juicio: juicio{
			eval: evalDelMunicipioCubierto(), skill: skillDeTerritorio, sesion: sesion("LEGANES"), esperado: resultado("LEGANES"),
		}},
		{modo: ModoHerramienta, juicio: juicio{
			eval: evalDelMunicipioCubierto(), skill: skillDeTerritorio, sesion: sesion("28074"),
			esperado: cambiado(resultado("28074"), func(r *ResultadoDeEval) {
				r.ComandosEjecutados = nil
				r.ComandosAusentes = []string{textoDelComandoDeLeganes}
				r.Motivos = []string{"comando ausente: " + textoDelComandoDeLeganes}
				r.Pasa = false
			}),
		}},
	}
}

// ordenEnElModoHerramienta son los juicios de la eval 01 con una sesión que
// ejecuta kitlegal por la ruta absoluta de su binario, la que declara
// servidor.json (research.md D18 de H21): en el modo herramienta no pasa, con un
// motivo por cada orden, en su orden y consulte o no, aunque la orden satisfaga
// el comando o lo satisfaga además una llamada; la misma sesión en el modo orden
// pasa.
func ordenEnElModoHerramienta(t *testing.T) []juicioDeModo {
	t.Helper()

	const laAyuda = "boe --help"

	porLaRuta := func(tokens ...string) Invocacion {
		return invocada(t, codigoDeSalida(0), slices.Concat([]string{kitlegalDelServidor}, tokens))
	}

	laLectura := porLaRuta("boe", "articulo", normaDeLasTrazas, "a21", "--json")
	leida := InvocacionInformada{Orden: ordenDelArticulo21, Codigo: codigoDeSalida(0)}
	servida := InvocacionInformada{Orden: ordenDelServidor, Codigo: codigoDeSalida(0)}

	conLaOrden := sesionTerminada(true, respuestaConCita, sirveElServidor(t, codigoDeSalida(0)), laLectura)

	conLaLlamadaYDosOrdenes := cambiada(
		sesionTerminada(true, respuestaConCita, porLaRuta("boe", "--help"), sirveElServidor(t, codigoDeSalida(0)), laLectura),
		func(s *Sesion) {
			s.Llamadas = []Llamada{llamadaCorrecta(herramientaDeArticulo, argumentosDelArticulo21)}
		})

	return []juicioDeModo{
		{modo: ModoHerramienta, juicio: juicio{
			eval:   evalDelArticulo21(),
			sesion: conLaOrden,
			esperado: cambiado(resultadoQuePasa(servida), func(r *ResultadoDeEval) {
				r.Modo = ModoHerramienta
				r.Motivos = []string{principioDeOrdenSinKitlegal + ordenDelArticulo21}
				r.Pasa = false
			}),
		}},
		{modo: ModoHerramienta, juicio: juicio{
			eval:   evalDelArticulo21(),
			sesion: conLaLlamadaYDosOrdenes,
			esperado: cambiado(resultadoQuePasa(), func(r *ResultadoDeEval) {
				r.Modo = ModoHerramienta
				r.Invocaciones = []InvocacionInformada{
					{Orden: laAyuda, Codigo: codigoDeSalida(0)},
					servida, leida,
					llamadaInformada(llamadaDelArticulo21, codigoDeSalida(0)),
				}
				r.Motivos = []string{principioDeOrdenSinKitlegal + laAyuda, principioDeOrdenSinKitlegal + ordenDelArticulo21}
				r.Pasa = false
			}),
		}},
		{modo: ModoOrden, juicio: juicio{
			eval: evalDelArticulo21(), sesion: conLaOrden, esperado: resultadoQuePasa(servida),
		}},
	}
}

// elServidorNoEsUnaConsulta son los juicios de la eval 01 con una sesión que la
// pasa y en cuya traza el proceso del servidor termina con un código distinto de
// 0 —el de una señal, el de la fuente no disponible o el de unos argumentos
// inválidos— y, en el primero, conecta con la red: no va a fuera de lo grabado ni
// a las otras fallidas, su conexión se informa con él y cuenta como una llegada a
// la red, y la eval pasa; tampoco ejecuta un prohibido que lo nombre. En el modo
// orden tampoco es una consulta.
func elServidorNoEsUnaConsulta(t *testing.T) []juicioDeModo {
	t.Helper()

	conLaLlamada := func(servidor Invocacion) Sesion {
		return cambiada(sesionTerminada(true, respuestaConCita, servidor), func(s *Sesion) {
			s.Llamadas = []Llamada{llamadaCorrecta(herramientaDeArticulo, argumentosDelArticulo21)}
		})
	}

	esperado := func(servido InvocacionInformada) ResultadoDeEval {
		return cambiado(resultadoQuePasa(), func(r *ResultadoDeEval) {
			r.Modo = ModoHerramienta
			r.Invocaciones = []InvocacionInformada{servido, llamadaInformada(llamadaDelArticulo21, codigoDeSalida(0))}
		})
	}

	queProhibeElServidor := evalDelArticulo21()
	queProhibeElServidor.Prohibidos = []ComandoProhibido{{Applet: "mcp", Verbo: "serve"}}

	return []juicioDeModo{
		{modo: ModoHerramienta, juicio: juicio{
			eval: evalDelArticulo21(),
			sesion: conLaLlamada(sirveElServidor(t, codigoDeSalida(-1),
				conexionInet(destinoPublico, resultadoEnCurso, ConexionRed))),
			esperado: cambiado(esperado(InvocacionInformada{
				Orden:      ordenDelServidor,
				Codigo:     codigoDeSalida(-1),
				Conexiones: []ConexionInformada{{Destino: destinoPublico, Clase: ConexionRed}},
			}), func(r *ResultadoDeEval) {
				r.LlegadasALaRed = []LlegadaALaRed{{Orden: ordenDelServidor, Destino: destinoPublico}}
			}),
		}},
		{modo: ModoHerramienta, juicio: juicio{
			eval:     queProhibeElServidor,
			sesion:   conLaLlamada(sirveElServidor(t, codigoDeSalida(4))),
			esperado: esperado(InvocacionInformada{Orden: ordenDelServidor, Codigo: codigoDeSalida(4)}),
		}},
		{modo: ModoOrden, juicio: juicio{
			eval:     queProhibeElServidor,
			sesion:   sesionQuePasa(t, sirveElServidor(t, codigoDeSalida(2))),
			esperado: resultadoQuePasa(InvocacionInformada{Orden: ordenDelServidor, Codigo: codigoDeSalida(2)}),
		}},
	}
}

// llamadasQueNoConvierten son los juicios de la eval 01 con una sesión del modo
// herramienta cuya llamada a boe_articulo lleva unos argumentos que no tienen
// orden equivalente —una propiedad de más, que nombra una bandera global, un
// valor de otro tipo, o algo que no es un objeto— y recibe el error de la clase
// argumentos: cuenta como una invocación sin argumentos, que se informa con la
// herramienta sola, no satisface el comando y va a las otras fallidas.
func llamadasQueNoConvierten(t *testing.T) []juicioDeModo {
	t.Helper()

	argumentos := []string{
		`{"norma":"BOE-A-2015-10565","bloque":"a21","json":true}`,
		`{"norma":"BOE-A-2015-10565","bloque":21}`,
		`["BOE-A-2015-10565","a21"]`,
	}

	juicios := make([]juicioDeModo, 0, len(argumentos))

	for _, dados := range argumentos {
		juicios = append(juicios, juicioDeModo{modo: ModoHerramienta, juicio: juicio{
			eval: evalDelArticulo21(),
			sesion: sesionDelModoHerramienta(t, respuestaConCita,
				llamadaFallida(herramientaDeArticulo, dados, schema.ClaseArgumentos)),
			esperado: cambiado(resultadoDeLasLlamadas(llamadaInformada(herramientaDeArticulo, codigoDeSalida(2))),
				func(r *ResultadoDeEval) {
					sinElBloqueDeLaEval01(r)

					r.OtrasFallidas = []InvocacionFallida{{Orden: herramientaDeArticulo, Codigo: codigoDeSalida(2)}}
				}),
		}})
	}

	return juicios
}

// TestVerbosDeLasHerramientas fija de dónde salen el applet, el verbo y los
// argumentos con los que una llamada cuenta como una invocación (H21 FR-003,
// FR-042): cada herramienta que el servidor anuncia con el registro de
// producción, y ninguna más, tiene su verbo, cuyo applet y cuyo verbo componen su
// nombre, <applet>_<verbo>.
func TestVerbosDeLasHerramientas(t *testing.T) {
	t.Parallel()

	registro, err := app.RegistroDeProduccion("")
	require.NoError(t, err)

	verbos, err := verbosDeLasHerramientas()
	require.NoError(t, err)

	anunciadas := app.NombresDeHerramientas(registro)
	require.NotEmpty(t, anunciadas, "el servidor anuncia herramientas")

	assert.ElementsMatch(t, anunciadas, slices.Collect(maps.Keys(verbos)),
		"cada herramienta anunciada, y ninguna más, tiene su verbo")

	for herramienta, verbo := range verbos {
		assert.Equal(t, herramienta, verbo.Applet+"_"+verbo.Verbo, "el nombre de %s es su applet, _ y su verbo", herramienta)
		assert.NotNil(t, verbo.Argumentos, "%s tiene los argumentos de su verbo", herramienta)
	}
}

// TestJuzgarSinBinarioNiServidor fija el juicio de una eval sin binario ni
// servidor, con las dos del contrato y sesiones construidas en memoria
// (contracts/evals-en-dos-modos.md §4 de H21; research.md D21 de H21; FR-046,
// FR-047, FR-081; SC-012; US3-1 a US3-4): de las cuatro respuestas —con la línea
// ⚠ SIN CONSULTA AL BOE: y su dirección y sin citas; sin la línea; con la
// etiqueta y sin la dirección en su línea; y con la línea y una cita— pasa solo
// la primera, y cada una de las otras lleva su motivo, carácter a carácter. El
// resultado dice si la línea está con su dirección y qué citas lleva la
// respuesta, y no tiene modo.
func TestJuzgarSinBinarioNiServidor(t *testing.T) {
	t.Parallel()

	const (
		sinLaLinea = "No he podido consultar el BOE. Instala kitlegal desde https://kitlegal.es/instalar/ " +
			"y vuelve a preguntar."
		sinLaDireccionEnSuLinea = lineaSinSuDireccion + "\n\n" +
			"Para consultarlo hace falta instalar kitlegal: https://kitlegal.es/instalar/"
		conUnaCita = lineaSinConsultaAlBOE + "\n\nDe memoria, el artículo 53 de la Ley 39/2015 reconoce los derechos " +
			"del interesado " + citaDelArticulo53 + "."
	)

	respuestas := []struct {
		nombre    string
		respuesta string
		cambiar   func(*ResultadoDeEval)
	}{
		{nombre: "con-la-linea-y-sin-citas", respuesta: lineaSinConsultaAlBOE, cambiar: func(*ResultadoDeEval) {}},
		{nombre: "sin-la-linea", respuesta: sinLaLinea, cambiar: func(r *ResultadoDeEval) {
			r.LineaSinConsulta = false
			r.Motivos = []string{motivoDeLaLineaAusente}
			r.Pasa = false
		}},
		{nombre: "con-la-etiqueta-y-sin-la-direccion-en-su-linea", respuesta: sinLaDireccionEnSuLinea, cambiar: func(r *ResultadoDeEval) {
			r.LineaSinConsulta = false
			r.Motivos = []string{motivoDeLaLineaSinDireccion}
			r.Pasa = false
		}},
		{nombre: "con-la-linea-y-una-cita", respuesta: conUnaCita, cambiar: func(r *ResultadoDeEval) {
			r.CitasSinConsulta = []string{textoDeLaCita53}
			r.Motivos = []string{principioDeCitaSinConsulta + textoDeLaCita53}
			r.Pasa = false
		}},
	}

	evals := []struct {
		eval  Eval
		skill string
	}{
		{eval: sinBinarioDeNormativa(), skill: skillNormativa},
		{eval: sinBinarioDePlazosYTerritorio(), skill: skillDeTerritorio},
	}

	for _, caso := range respuestas {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			for _, medida := range evals {
				sesion := sesionSinBinarioNiServidor(medida.skill, caso.respuesta)
				esperado := cambiado(resultadoSinConsulta(medida.eval, caso.respuesta), caso.cambiar)

				assert.Equal(t, esperado, Juzgar(medida.eval, sesion, medida.skill),
					"la eval %s con la respuesta del caso", medida.eval.Fichero)
			}
		})
	}
}

// TestJuzgarSinBinarioNiServidorConLoDeSiempre fija lo que rodea a las tres
// comprobaciones de una eval sin binario ni servidor
// (contracts/evals-en-dos-modos.md §4 de H21; FR-047; research.md D18 de H21):
// la línea vale con la tolerancia de las etiquetas de los avisos; sin la línea,
// la dirección no se echa en falta, y cada cita de la respuesta lleva su motivo,
// en su orden y con sus repeticiones, detrás del de la línea; la sesión que no
// termina o no activa la skill no pasa aunque la respuesta cumpla, con el motivo
// de siempre; una orden de kitlegal en su traza no pasa, con su motivo delante
// de los de la respuesta; y una eval de modo se juzga como antes aunque su
// respuesta lleve la línea y una cita, sin línea ni citas en su resultado.
func TestJuzgarSinBinarioNiServidorConLoDeSiempre(t *testing.T) {
	t.Parallel()

	const (
		conEnfasis    = "**⚠ SIN CONSULTA AL BOE:** no hay binario. Instálalo: https://kitlegal.es/instalar/"
		otraCita      = "[Ley 39/2015, BOE-A-2015-10565, bloque a13]"
		sinLineaYCon3 = "El artículo 53 " + citaDelArticulo53 + " remite al 13 " + otraCita + " y vuelve al 53 " +
			citaDelArticulo53 + "."
		laOrden = "boe articulo BOE-A-2015-10565 a53 --json"
	)

	eval := sinBinarioDeNormativa()

	conLaOrden := cambiada(sesionSinBinarioNiServidor(skillNormativa, lineaSinSuDireccion), func(s *Sesion) {
		s.Invocaciones = []Invocacion{
			invocada(t, codigoDeSalida(4), []string{kitlegalDelServidor, "boe", "articulo", normaDeLasTrazas, "a53", "--json"}),
		}
	})

	casos := []struct {
		nombre string
		juicio juicio
	}{
		{nombre: "con-enfasis", juicio: juicio{
			eval:     eval,
			sesion:   sesionSinBinarioNiServidor(skillNormativa, conEnfasis),
			esperado: resultadoSinConsulta(eval, conEnfasis),
		}},
		{nombre: "sin-la-linea-y-con-tres-citas", juicio: juicio{
			eval:   eval,
			sesion: sesionSinBinarioNiServidor(skillNormativa, sinLineaYCon3),
			esperado: cambiado(resultadoSinConsulta(eval, sinLineaYCon3), func(r *ResultadoDeEval) {
				r.LineaSinConsulta = false
				r.CitasSinConsulta = []string{textoDeLaCita53, "BOE-A-2015-10565 a13", textoDeLaCita53}
				r.Motivos = slices.Concat([]string{motivoDeLaLineaAusente},
					prefijados(principioDeCitaSinConsulta, r.CitasSinConsulta))
				r.Pasa = false
			}),
		}},
		{nombre: "sin-terminar", juicio: juicio{
			eval: eval,
			sesion: cambiada(sesionSinBinarioNiServidor(skillNormativa, lineaSinConsultaAlBOE), func(s *Sesion) {
				s.Fin, s.Terminada, s.MotivoSinTerminar = "result error_max_turns", false, "result con subtype error_max_turns"
			}),
			esperado: cambiado(resultadoSinConsulta(eval, lineaSinConsultaAlBOE), func(r *ResultadoDeEval) {
				r.FinDeLaSesion, r.SesionTerminada = "result error_max_turns", false
				r.Motivos = []string{"la sesión no terminó: result con subtype error_max_turns"}
				r.Pasa = false
			}),
		}},
		{nombre: "sin-activar", juicio: juicio{
			eval:   eval,
			sesion: sesionTerminada(false, lineaSinConsultaAlBOE),
			esperado: cambiado(resultadoSinConsulta(eval, lineaSinConsultaAlBOE), func(r *ResultadoDeEval) {
				r.Activada = false
				r.Motivos = []string{
					"la activación no coincide: se esperaba que la skill normativa se activara y no se activó",
				}
				r.Pasa = false
			}),
		}},
		{nombre: "con-una-orden-de-kitlegal", juicio: juicio{
			eval:   eval,
			sesion: conLaOrden,
			esperado: cambiado(resultadoSinConsulta(eval, lineaSinSuDireccion), func(r *ResultadoDeEval) {
				r.LineaSinConsulta = false
				r.Invocaciones = []InvocacionInformada{{Orden: laOrden, Codigo: codigoDeSalida(4)}}
				r.FueraDeLoGrabado = []InvocacionFallida{{Orden: laOrden, Codigo: codigoDeSalida(4)}}
				r.Motivos = []string{principioDeOrdenSinKitlegal + laOrden, motivoDeLaLineaSinDireccion}
				r.Pasa = false
			}),
		}},
		{nombre: "una-eval-de-modo", juicio: juicio{
			eval: evalDelArticulo21(),
			sesion: cambiada(sesionQuePasa(t), func(s *Sesion) {
				s.Respuesta = lineaSinConsultaAlBOE + "\n\n" + respuestaConCita
			}),
			esperado: cambiado(resultadoQuePasa(), func(r *ResultadoDeEval) {
				r.Respuesta = lineaSinConsultaAlBOE + "\n\n" + respuestaConCita
			}),
		}},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.juicio.esperado, Juzgar(caso.juicio.eval, caso.juicio.sesion, skillNormativa))
		})
	}
}

// TestElModoYLasLlamadasDelResultadoEnJSON fija las claves que el resultado de
// una sesión gana en H21, codificado como lo codifica EscribirInforme
// (contracts/evals-en-dos-modos.md §4 de H21; FR-042, FR-047): modo, detrás de
// modelo_de_la_sesion, con el texto del modo y vacío en una eval sin binario ni
// servidor; linea_sin_consulta y citas_sin_consulta, detrás de
// territorio_ausente desde H24, que quita expresiones_prohibidas de entre ellas
// (contracts/informe-del-job.md §6 de H24), la segunda una lista vacía, nunca
// null, en las evals de modo; y llamada en cada elemento de invocaciones, detrás
// de conexiones, verdadero solo en las que son llamadas a una herramienta.
func TestElModoYLasLlamadasDelResultadoEnJSON(t *testing.T) {
	t.Parallel()

	conUnaCita := lineaSinConsultaAlBOE + "\n\n" + citaDelArticulo53

	casos := []struct {
		nombre    string
		resultado ResultadoDeEval
		claves    []string
	}{
		{
			nombre:    "modo-orden",
			resultado: Juzgar(evalDelArticulo21(), sesionQuePasa(t), skillNormativa),
			claves: []string{
				`"modelo_de_la_sesion":"","modo":"orden","decide":false,`,
				`"territorio_ausente":[],"linea_sin_consulta":false,"citas_sin_consulta":[],"invocaciones":[` +
					`{"orden":"boe articulo BOE-A-2015-10565 a21 --json","codigo":0,"conexiones":[],"llamada":false}],`,
			},
		},
		{
			nombre: "modo-herramienta",
			resultado: juzgarEnModo(evalDelArticulo21(), sesionDelModoHerramienta(t, respuestaConCita,
				llamadaCorrecta(herramientaDeArticulo, argumentosDelArticulo21)), skillNormativa, ModoHerramienta),
			claves: []string{
				`"modelo_de_la_sesion":"","modo":"herramienta","decide":false,`,
				`"linea_sin_consulta":false,"citas_sin_consulta":[],"invocaciones":[` +
					`{"orden":"mcp serve","codigo":0,"conexiones":[],"llamada":false},` +
					`{"orden":"boe_articulo BOE-A-2015-10565 a21","codigo":0,"conexiones":[],"llamada":true}],`,
			},
		},
		{
			nombre: "sin-binario-ni-servidor",
			resultado: Juzgar(sinBinarioDeNormativa(), sesionSinBinarioNiServidor(skillNormativa, conUnaCita),
				skillNormativa),
			claves: []string{
				`"modelo_de_la_sesion":"","modo":"","decide":false,`,
				`"territorio_ausente":[],"linea_sin_consulta":true,` +
					`"citas_sin_consulta":["BOE-A-2015-10565 a53"],"invocaciones":[],`,
			},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			codificado, err := json.Marshal(caso.resultado)
			require.NoError(t, err)

			for _, clave := range caso.claves {
				assert.Contains(t, string(codificado), clave)
			}
		})
	}
}

// sinBinarioDeNormativa es la eval sin binario ni servidor de
// normativa (contracts/evals-en-dos-modos.md §1 de H21), como se lee de
// evalSinBinarioDeNormativa.
func sinBinarioDeNormativa() Eval {
	return Eval{
		Fichero:              ficheroSinBinarioDeBoe,
		Pregunta:             preguntaSinBinarioDeNormativa,
		Activa:               true,
		SinBinarioNiServidor: true,
	}
}

// sinBinarioDePlazosYTerritorio es la eval sin binario ni servidor de
// plazos-y-territorio, que declara además que normativa no se activa
// (contracts/evals-en-dos-modos.md §1 de H21), como se lee de
// evalSinBinarioDePlazosYTerritorio.
func sinBinarioDePlazosYTerritorio() Eval {
	return Eval{
		Fichero:              ficheroSinBinarioDePlazosYTerritorio,
		Pregunta:             preguntaSinBinarioDePlazosYTerritorio,
		Activa:               true,
		NoSeActivan:          []string{skillQueNoSeActiva},
		SinBinarioNiServidor: true,
	}
}

// sesionSinBinarioNiServidor es la sesión que terminó con código 0 y result
// success, activó la skill dada y tiene esa respuesta, sin invocaciones ni
// llamadas: la de una eval sin binario ni servidor.
func sesionSinBinarioNiServidor(skill, respuesta string) Sesion {
	return cambiada(sesionTerminada(false, respuesta), func(s *Sesion) { s.SkillsActivadas = []string{skill} })
}

// resultadoSinConsulta es el resultado de una eval sin binario ni servidor con
// una sesión terminada y activada cuya respuesta lleva la línea con su dirección
// y ninguna cita: sin modo, y pasa.
func resultadoSinConsulta(eval Eval, respuesta string) ResultadoDeEval {
	return ResultadoDeEval{
		Eval:             eval.Fichero,
		Activa:           true,
		Activada:         true,
		LineaSinConsulta: true,
		Respuesta:        respuesta,
		CodigoDeLaSesion: codigoDeSalida(0),
		FinDeLaSesion:    "result success",
		SesionTerminada:  true,
		Pasa:             true,
	}
}

// llamadaCorrecta es la llamada a la herramienta con esos argumentos cuyo
// resultado no es un error: la que devuelve un sobre con ok verdadero.
func llamadaCorrecta(herramienta, argumentos string) Llamada {
	return Llamada{Herramienta: herramienta, Argumentos: []byte(argumentos), ConResultado: true}
}

// llamadaFallida es la llamada a la herramienta con esos argumentos cuyo
// resultado es un error de la clase dada.
func llamadaFallida(herramienta, argumentos string, clase schema.Clase) Llamada {
	return Llamada{Herramienta: herramienta, Argumentos: []byte(argumentos), ConResultado: true, Error: true, Clase: clase}
}

// llamadaInformada es una llamada a una herramienta como la presenta el
// informe entre las invocaciones: con la herramienta y sus argumentos como
// orden, el código dado y la marca de llamada.
func llamadaInformada(orden string, codigo *int) InvocacionInformada {
	return InvocacionInformada{Orden: orden, Codigo: codigo, Llamada: true}
}

// llamadasDelTranscript son las llamadas que LeerSesion lee de los mensajes de
// un transcript, dados en memoria con su salto de línea: las de las herramientas
// del registro de producción, con el nombre sin el prefijo del agente.
func llamadasDelTranscript(t *testing.T, transcript string) []Llamada {
	t.Helper()

	herramientas, err := herramientasDelRegistro()
	require.NoError(t, err)

	leido := transcriptLeido{herramientas: herramientas}

	for linea := range strings.Lines(transcript) {
		require.NoError(t, leido.leerMensaje(strings.TrimSuffix(linea, "\n")))
	}

	require.NotEmpty(t, leido.llamadas, "el transcript tiene llamadas a herramientas del registro")

	return leido.llamadas
}

// sirveElServidor es la invocación de la traza del proceso del servidor MCP,
// que el agente arranca con la orden de servidor.json, con el código y las
// conexiones dados.
func sirveElServidor(t *testing.T, codigo *int, conexiones ...Conexion) Invocacion {
	t.Helper()

	return invocada(t, codigo, []string{kitlegalDelServidor, "mcp", "serve"}, conexiones...)
}

// sesionDelModoHerramienta es la sesión del modo herramienta que terminó con
// código 0 y result success y activó normativa, con la respuesta y las
// llamadas dadas y, en su traza, solo el proceso del servidor, que terminó con 0.
func sesionDelModoHerramienta(t *testing.T, respuesta string, llamadas ...Llamada) Sesion {
	t.Helper()

	return cambiada(sesionTerminada(true, respuesta, sirveElServidor(t, codigoDeSalida(0))), func(s *Sesion) {
		s.Llamadas = llamadas
	})
}

// resultadoDeLasLlamadas es el resultado de la eval 01 con una sesión del modo
// herramienta que la pasa: como resultadoQuePasa, con el modo herramienta y, por
// invocaciones, el proceso del servidor y, detrás, las llamadas dadas.
func resultadoDeLasLlamadas(llamadas ...InvocacionInformada) ResultadoDeEval {
	return cambiado(resultadoQuePasa(), func(r *ResultadoDeEval) {
		r.Modo = ModoHerramienta
		r.Invocaciones = slices.Concat(
			[]InvocacionInformada{{Orden: ordenDelServidor, Codigo: codigoDeSalida(0)}}, llamadas)
	})
}

// sinElBloqueDeLaEval01 deja el resultado de la eval 01 sin su comando: ausente,
// con su motivo, y sin pasar.
func sinElBloqueDeLaEval01(r *ResultadoDeEval) {
	r.ComandosEjecutados = nil
	r.ComandosAusentes = []string{textoDelComando21}
	r.Motivos = []string{"comando ausente: " + textoDelComando21}
	r.Pasa = false
}

// Lo que juzgan las sesiones de TestJuzgarLecturasFechadas: el día de la eval
// (a) de H20 y otro, y el comando de bloque del art. 118 de la LCSP con ese día
// con el texto con el que lo presentan el informe y los motivos, el de bloque
// con la fecha detrás (contracts/evals-redaccion-fechada.md §1.1 de H20).
const (
	diaDeLaEvalConFecha     = "2019-06-01"
	otroDia                 = "2019-06-02"
	textoDelComandoConFecha = textoDelComando118 + " " + diaDeLaEvalConFecha
)

// TestJuzgarLecturasFechadas fija qué invocación cumple un comando de bloque con
// fecha y cuál uno sin ella (contracts/evals-redaccion-fechada.md §1.1 de H20;
// data-model §7.1; research D12; FR-051, FR-089). Los argumentos de articulo se
// leen separando --fecha, con su valor detrás o pegado con «=», en cualquier
// sitio —también tras el terminador, que no es un argumento de posición—, de los
// de posición, la norma y sus bloques. El comando con fecha lo cumple articulo de
// la norma y el bloque con esa fecha, como orden o como llamada a boe_articulo
// con fecha, que termina con 0 o con 3; no lo cumplen la lectura sin fecha, la de
// otra fecha, articulos, la de otra norma o de otro bloque, la que no consulta, la
// que queda sin código ni la que termina con 2, 4 o 5. El comando sin fecha lo
// cumple, como en H5, articulo o articulos de la norma con el bloque que terminan
// con 0, y ninguna lectura que lleva --fecha. El comando que no se cumple queda
// ausente con su motivo, con la fecha en su texto si la lleva. Y el caso 9 de
// FR-089: el comando sin fecha de la eval 02 del repositorio no lo cumple una
// lectura fechada de su bloque, como orden ni como herramienta, que deja su
// motivo, y sí la lectura sin fecha.
func TestJuzgarLecturasFechadas(t *testing.T) {
	t.Parallel()

	conFecha := evalDelArticulo118(ComandoEsperado{
		Applet: "boe", Norma: normaDeLaLCSP, Bloque: bloqueDelArticulo118, Fecha: diaDeLaEvalConFecha,
	})
	sinFecha := evalDelArticulo118(ComandoEsperado{Applet: "boe", Norma: normaDeLaLCSP, Bloque: bloqueDelArticulo118})

	for _, caso := range lecturasDelArticulo118(t) {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			comandos := []struct {
				eval   Eval
				texto  string
				cumple bool
			}{
				{eval: conFecha, texto: textoDelComandoConFecha, cumple: caso.cumpleElConFecha},
				{eval: sinFecha, texto: textoDelComando118, cumple: caso.cumpleElSinFecha},
			}

			for _, comando := range comandos {
				resultado := juzgarEnModo(comando.eval, caso.sesion, skillNormativa, caso.modo)

				if comando.cumple {
					assert.Equal(t, []string{comando.texto}, resultado.ComandosEjecutados, "%s se cumple", comando.texto)
					assert.Empty(t, resultado.ComandosAusentes, "%s no queda ausente", comando.texto)

					continue
				}

				assert.Empty(t, resultado.ComandosEjecutados, "%s no se cumple", comando.texto)
				assert.Equal(t, []string{comando.texto}, resultado.ComandosAusentes, "%s queda ausente", comando.texto)
				assert.Contains(t, resultado.Motivos, "comando ausente: "+comando.texto, "con su motivo")
			}
		})
	}

	t.Run("caso-9-de-FR-089", probarElComandoSinFechaDeLaEval02)
}

// lecturaDelArticulo118 es una sesión que lee, o no, el art. 118 de la LCSP, en
// su modo, y si cumple el comando con la fecha de la eval (a) y el comando sin
// fecha.
type lecturaDelArticulo118 struct {
	nombre           string
	sesion           Sesion
	modo             Modo
	cumpleElConFecha bool
	cumpleElSinFecha bool
}

// lecturasDelArticulo118 son las sesiones de TestJuzgarLecturasFechadas: las de
// la tabla de contracts/evals-redaccion-fechada.md §1.1 de H20, como orden y como
// herramienta.
func lecturasDelArticulo118(t *testing.T) []lecturaDelArticulo118 {
	t.Helper()

	// lee es la sesión del modo orden con la invocación de la skill de esos
	// argumentos que termina con ese código, o sin ninguno si es nil.
	lee := func(codigo *int, argumentos ...string) Sesion {
		return sesionTerminada(true, respuestaConLaCita118, invocada(t, codigo, deLaSkill(argumentos...)))
	}

	// llama es la sesión del modo herramienta con esa llamada.
	llama := func(llamada Llamada) Sesion {
		return sesionDelModoHerramienta(t, respuestaConLaCita118, llamada)
	}

	conLaFecha := `{"norma":"BOE-A-2017-12902","bloque":"a1-30","fecha":"2019-06-01"}`
	conOtraFecha := `{"norma":"BOE-A-2017-12902","bloque":"a1-30","fecha":"2019-06-02"}`

	const norma, bloque, dia = normaDeLaLCSP, bloqueDelArticulo118, diaDeLaEvalConFecha

	lecturas := []lecturaDelArticulo118{
		{nombre: "fecha-delante", sesion: lee(codigoDeSalida(0), "articulo", "--fecha", dia, norma, bloque, "--json")},
		{nombre: "fecha-detras", sesion: lee(codigoDeSalida(0), "articulo", norma, bloque, "--fecha", dia)},
		{nombre: "fecha-delante-con-igual", sesion: lee(codigoDeSalida(0), "articulo", "--fecha="+dia, norma, bloque)},
		{nombre: "fecha-detras-con-igual", sesion: lee(codigoDeSalida(0), "articulo", norma, bloque, "--fecha="+dia)},
		{nombre: "fecha-en-medio", sesion: lee(codigoDeSalida(0), "articulo", norma, "--fecha", dia, bloque)},
		{nombre: "fecha-y-terminador", sesion: lee(codigoDeSalida(0), "articulo", "--fecha", dia, "--", norma, bloque)},
		{nombre: "fecha-tras-el-terminador", sesion: lee(codigoDeSalida(0), "articulo", "--", norma, bloque, "--fecha", dia)},
		{nombre: "fecha-con-el-3", sesion: lee(codigoDeSalida(3), "articulo", norma, bloque, "--fecha", dia)},
		{nombre: "herramienta-con-la-fecha", sesion: llama(llamadaCorrecta(herramientaDeArticulo, conLaFecha))},
		{
			nombre: "herramienta-con-la-fecha-y-el-3",
			sesion: llama(llamadaFallida(herramientaDeArticulo, conLaFecha, schema.ClaseNoEncontrado)),
		},
	}

	for posicion := range lecturas {
		lecturas[posicion].cumpleElConFecha = true
	}

	lecturas = append(lecturas,
		lecturaDelArticulo118{
			nombre: "sin-fecha", sesion: lee(codigoDeSalida(0), "articulo", norma, bloque, "--json"),
			cumpleElSinFecha: true,
		},
		lecturaDelArticulo118{
			nombre: "articulos-sin-fecha", sesion: lee(codigoDeSalida(0), "articulos", norma, "da-3", bloque),
			cumpleElSinFecha: true,
		},
		lecturaDelArticulo118{
			nombre: "sin-fecha-tras-el-terminador", sesion: lee(codigoDeSalida(0), "articulo", "--", norma, bloque),
			cumpleElSinFecha: true,
		},
		lecturaDelArticulo118{
			nombre:           "herramienta-sin-la-fecha",
			sesion:           llama(llamadaCorrecta(herramientaDeArticulo, `{"norma":"BOE-A-2017-12902","bloque":"a1-30"}`)),
			cumpleElSinFecha: true,
		},
		lecturaDelArticulo118{nombre: "sin-fecha-con-el-3", sesion: lee(codigoDeSalida(3), "articulo", norma, bloque)},
		lecturaDelArticulo118{
			nombre: "otra-fecha", sesion: lee(codigoDeSalida(0), "articulo", norma, bloque, "--fecha", otroDia),
		},
		lecturaDelArticulo118{
			nombre: "otra-fecha-con-igual", sesion: lee(codigoDeSalida(0), "articulo", norma, bloque, "--fecha="+otroDia),
		},
		lecturaDelArticulo118{
			nombre: "fecha-sin-valor", sesion: lee(codigoDeSalida(0), "articulo", norma, bloque, "--fecha"),
		},
		lecturaDelArticulo118{
			nombre: "articulos-con-la-fecha", sesion: lee(codigoDeSalida(0), "articulos", norma, bloque, "--fecha", dia),
		},
		lecturaDelArticulo118{
			nombre: "otra-norma", sesion: lee(codigoDeSalida(0), "articulo", normaDeLasTrazas, bloque, "--fecha", dia),
		},
		lecturaDelArticulo118{
			nombre: "otro-bloque", sesion: lee(codigoDeSalida(0), "articulo", norma, bloqueDeLaDA3, "--fecha", dia),
		},
		lecturaDelArticulo118{
			nombre: "sin-consulta", sesion: lee(codigoDeSalida(0), "articulo", "--dry-run", norma, bloque, "--fecha", dia),
		},
		lecturaDelArticulo118{nombre: "sin-codigo", sesion: lee(nil, "articulo", norma, bloque, "--fecha", dia)},
		lecturaDelArticulo118{
			nombre: "herramienta-con-otra-fecha", sesion: llama(llamadaCorrecta(herramientaDeArticulo, conOtraFecha)),
		},
		lecturaDelArticulo118{
			nombre: "herramienta-con-la-fecha-y-el-4",
			sesion: llama(llamadaFallida(herramientaDeArticulo, conLaFecha, schema.ClaseFuenteNoDisponible)),
		},
	)

	for _, codigo := range []int{2, 4, 5} {
		lecturas = append(lecturas, lecturaDelArticulo118{
			nombre: "fecha-con-el-" + strconv.Itoa(codigo),
			sesion: lee(codigoDeSalida(codigo), "articulo", norma, bloque, "--fecha", dia),
		})
	}

	for posicion, lectura := range lecturas {
		lecturas[posicion].modo = ModoOrden
		if len(lectura.sesion.Llamadas) > 0 {
			lecturas[posicion].modo = ModoHerramienta
		}
	}

	return lecturas
}

// probarElComandoSinFechaDeLaEval02 es el caso 9 de FR-089 de H20: la eval 02
// de normativa, leída del repositorio, espera el bloque a1-30 de la LCSP
// sin fecha; una sesión que lo cita y solo lo lee con fecha, como orden o como
// herramienta, no lo cumple, y su único motivo es el comando ausente, el de
// bloque sin fecha. La misma sesión con la lectura sin fecha pasa.
func probarElComandoSinFechaDeLaEval02(t *testing.T) {
	t.Parallel()

	conjunto, err := LeerConjunto(evalsDeNormativa)
	require.NoError(t, err)

	eval02 := *evalDe(t, conjunto.Evals, "02-lcsp-contrato-menor.yaml")
	require.Equal(t, []ComandoEsperado{{Applet: "boe", Norma: normaDeLaLCSP, Bloque: bloqueDelArticulo118}},
		eval02.Comandos, "premisa: la eval 02 espera la lectura sin fecha del a1-30")

	fechadas := []struct {
		modo   Modo
		sesion Sesion
	}{
		{modo: ModoOrden, sesion: sesionTerminada(true, respuestaConLaCita118, invocada(t, codigoDeSalida(0),
			deLaSkill("articulo", normaDeLaLCSP, bloqueDelArticulo118, "--fecha", diaDeLaEvalConFecha, "--json")))},
		{modo: ModoHerramienta, sesion: sesionDelModoHerramienta(t, respuestaConLaCita118, llamadaCorrecta(
			herramientaDeArticulo, `{"norma":"BOE-A-2017-12902","bloque":"a1-30","fecha":"2019-06-01"}`))},
	}

	for _, fechada := range fechadas {
		resultado := juzgarEnModo(eval02, fechada.sesion, skillNormativa, fechada.modo)

		assert.False(t, resultado.Pasa, "en el modo %s, la lectura fechada no cumple la eval 02", fechada.modo)
		assert.Equal(t, []string{"comando ausente: bloque boe BOE-A-2017-12902 a1-30"}, resultado.Motivos,
			"en el modo %s, el único motivo es el del comando sin fecha", fechada.modo)
	}

	sinFecha := sesionTerminada(true, respuestaConLaCita118, invocada(t, codigoDeSalida(0),
		deLaSkill("articulo", normaDeLaLCSP, bloqueDelArticulo118, "--json")))
	resultado := juzgarEnModo(eval02, sinFecha, skillNormativa, ModoOrden)

	assert.True(t, resultado.Pasa, "la lectura sin fecha cumple la eval 02: %v", resultado.Motivos)
}

// evalDelArticulo118 es una eval positiva sintética con ese comando esperado y la
// cita del art. 118 de la LCSP.
func evalDelArticulo118(comando ComandoEsperado) Eval {
	return Eval{
		Fichero:  "22-lcsp-contrato-menor-en-2019.yaml",
		Pregunta: "¿qué exigía el artículo 118 de la LCSP para el expediente de un contrato menor el 1 de junio de 2019?",
		Activa:   true,
		Comandos: []ComandoEsperado{comando},
		Citas:    []CitaEsperada{{Norma: normaDeLaLCSP, Bloque: bloqueDelArticulo118}},
	}
}

// Las evals 11 y 12 de plazos-y-territorio que juzgan las sesiones de
// TestJuzgarLasLecturasSinServir, con el contenido de
// contracts/formato-de-eval.md §3.1 y §3.2 de H26.1 y el razonamiento de la 05,
// que las juzgan también con su fichero del repositorio, y la 16 de
// normativa del repositorio.
const (
	ficheroDeLaEval11 = "11-plazo-reposicion-sin-leer-el-precepto.yaml"
	ficheroDeLaEval12 = "12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml"
	ficheroDeLaEval16 = "16-lrjsp-legalidad-por-materia.yaml"

	contenidoDeLaEval11 = `pregunta: "Me notificaron el 9 de septiembre de 2026 una resolución del Ayuntamiento de Leganés y vivo en Leganés, ¿hasta cuándo puedo presentar el recurso de reposición?"
activa: true
comandos:
  - applet: territorio
    verbo: resolver
    municipio: Leganés
  - applet: plazos
    verbo: calcular
    tipo: reposicion
  - applet: boe
    norma: BOE-A-2015-10565
    bloque: a124
    fecha: "2026-09-09"
sin_servir:
  - norma: BOE-A-2015-10565
    bloque: a124
plazo:
  fecha: "2026-10-13"
  avisos: []
  razonamiento: >-
    Un mes (art. 124.1 LPAC) contado desde el día siguiente al de la notificación (art. 30.4): vence el mismo día del
    mes siguiente, el viernes 9 de octubre de 2026, fiesta local de Leganés (BOCM-20251212-34). El último día inhábil
    se prorroga al primer día hábil siguiente (art. 30.5): el 10 y el 11 son sábado y domingo, y el 12 es fiesta
    nacional (BOE-A-2025-21667), así que el plazo vence el martes 13 de octubre de 2026. Los festivos saltados son el
    9, local de Leganés, y el 12, nacional.
`

	contenidoDeLaEval12 = `pregunta: "Vivo en Leganés y el 9 de septiembre de 2026 me notificaron una reclamación de cuotas de la Seguridad Social. ¿Hasta cuándo puedo recurrirla?"
activa: true
sin_servir:
  - norma: BOE-A-2015-10565
    bloque: daprimera
prohibidos:
  - applet: plazos
    verbo: calcular
no_cubierto:
  norma: BOE-A-2015-10565
  precepto: {norma: BOE-A-2015-10565, bloque: daprimera}
`
)

// Lo que leen y responden las sesiones de TestJuzgarLasLecturasSinServir: el día
// de la notificación; la lectura del art. 124 de la LPAC ese día, la orden que
// la pide y el texto de su comando en la 11; los argumentos de la llamada que la
// pide y su orden; las respuestas de la 11 y de la 12, con la línea del plazo y con la
// declaración de no cubierto de la entrada sin norma de data/cobertura.yaml; la
// norma de la 16 y la cita de su art. 25; y el motivo de sus dos lecturas.
const (
	diaDeLaNotificacion = "2026-09-09"

	ordenDelArticulo124      = "boe articulo BOE-A-2015-10565 a124 --fecha 2026-09-09 --json"
	textoDelComando124       = "bloque boe BOE-A-2015-10565 a124 2026-09-09"
	argumentosDelArticulo124 = `{"norma":"BOE-A-2015-10565","bloque":"a124","fecha":"2026-09-09"}`
	llamadaDelArticulo124    = "boe_articulo --fecha=2026-09-09 BOE-A-2015-10565 a124"

	respuestaDeLa11 = "Hasta el martes 13 de octubre de 2026.\n\n" + lineaDelPlazo + "\n\nNo he podido leer la " +
		"redacción del art. 124 vigente el 2026-09-09: el BOE no respondió."
	respuestaDeLa12 = "⚠ NO CUBIERTO: la gestión recaudatoria de la Seguridad Social lo regula aparte la ley, no el " +
		"procedimiento común: kitlegal no calcula sus plazos ni sus recursos (DA 1ª.2.b LPAC " + citaDeLaDAPrimera + ")."

	normaDeLaLRJSP = "BOE-A-2015-10566"
	citaDel25      = "[Ley 40/2015, art. 25, BOE-A-2015-10566, bloque a25]"

	motivoDeLasLecturasDeLa16 = "Junto al art. 25 que pide la eval, en su mismo capítulo: la exploración que " +
		"provoca la pregunta, vista en los jobs de la #175 y la #178 (ADR 0043, clase b)."
)

// TestJuzgarLasLecturasSinServir fija lo que el juicio hace con lo que una eval
// deja sin servir y con lo que lleva a leer (data-model §4 y §5 de H26.1;
// contracts/juicio-e-informe.md §1 y §2 de H26.1; research D5 y D6 de H26.1;
// FR-007, FR-013, FR-015 a FR-018 y FR-022 de H26.1; SC-004, SC-005):
//
//   - el reparto de cada invocación, por las filas 1, 3, 4, 5 y 7 de la tabla:
//     la que consulta una lectura de sin_servir va a lecturas_sin_servir termine
//     como termine, y no a fuera_de_lo_grabado ni a otras_fallidas, que siguen
//     como hoy con las demás;
//   - la 11, cuyo comando de bloque satisface la lectura en su fecha que sale con
//     4 o 5, como orden o como llamada, y ninguna otra: ni sin fecha, ni en otra,
//     ni con otro código; y que pasa con la línea del plazo, sin contar en
//     plazo_sin_fecha, y no sin ella;
//   - la 12, que pasa con la declaración de no cubierto de la LPAC, sin contar en
//     cobertura_sin_declarar, y no con plazos calcular;
//   - la sesión cortada, con la lectura sin código en lecturas_sin_servir y su
//     código null;
//   - y la 16 de normativa del repositorio, que juzga igual con sus
//     lecturas y sin ellas una sesión que no las lee.
func TestJuzgarLasLecturasSinServir(t *testing.T) {
	t.Parallel()

	t.Run("reparto", probarElRepartoDeLasLecturasSinServir)
	t.Run("eval-11", probarLaEvalSinServirDelPlazo)
	t.Run("eval-11-por-la-herramienta", probarLaEvalSinServirDelPlazoPorLaHerramienta)
	t.Run("eval-12", probarLaEvalSinServirDeLaCobertura)
	t.Run("sesion-cortada", probarLaSesionCortadaConLaLecturaSinServir)
	t.Run("eval-16-con-y-sin-lecturas", probarLasLecturasDeLaEval16)
}

// evalLeida es la eval con ese contenido, leída por LeerEval con su nombre de
// fichero: con el esquema y los cruces de lecturas y sin_servir.
func evalLeida(t *testing.T, fichero, contenido string) Eval {
	t.Helper()

	eval, err := LeerEval(fichero, []byte(contenido))
	require.NoError(t, err)

	return eval
}

// evalDeUnOrigen es una eval y de dónde sale: del contrato, con el contenido de
// este test, o del repositorio, con el de su fichero de
// evals/plazos-y-territorio/.
type evalDeUnOrigen struct {
	origen string
	eval   Eval
}

// lasDosEvals es la eval de ese fichero con el contenido del contrato y la del
// repositorio, leídas las dos por LeerEval: cada sesión se juzga con las dos y
// tiene que salir igual (T008 de H26.1).
func lasDosEvals(t *testing.T, fichero, contenido string) []evalDeUnOrigen {
	t.Helper()

	delRepositorio := contenidoDelFichero(t, filepath.Join(evalsDePlazosYTerritorio, fichero))

	return []evalDeUnOrigen{
		{origen: "del-contrato", eval: evalLeida(t, fichero, contenido)},
		{origen: "del-repositorio", eval: evalLeida(t, fichero, string(delRepositorio))},
	}
}

// calculaLaReposicion es la orden plazos calcular de la reposición desde el día
// de la notificación, con Leganés de los dos municipios, que termina con 0.
func calculaLaReposicion(t *testing.T) Invocacion {
	t.Helper()

	return calculaElPlazo(t, codigoDeSalida(0), "reposicion", diaDeLaNotificacion, "Leganés", "Leganés", "--json")
}

// deLaLPAC es la orden del applet boe con el verbo dado, la LPAC, y los demás
// argumentos, con el código dado.
func deLaLPAC(t *testing.T, codigo *int, verbo string, argumentos ...string) Invocacion {
	t.Helper()

	return invocada(t, codigo, slices.Concat([]string{kitlegalDelPath, "boe", verbo, normaDeLasTrazas}, argumentos))
}

// enSuUmbralSinModelo dice, de la respuesta de una sesión juzgada con una eval de
// plazo o de cobertura, si está en el total del umbral sin modelo de las suyas
// en su modo —plazo_sin_fecha o cobertura_sin_declarar— y si cuenta en su
// medida, con lo que el informe mide de ella (medirLosPlazos).
func enSuUmbralSinModelo(eval Eval, resultado ResultadoDeEval) (enElTotal, cuenta bool) {
	medidas := respuestasMedidas{grupos: []grupoJuzgado{
		{modo: resultado.Modo, respuestas: []respuestaDelJuez{{sesion: resultado.Sesion}}},
	}}
	medidas.medirLosPlazos([]Eval{eval}, []sesionJuzgada{{resultado: resultado}})

	grupo := medidas.grupos[0]
	if eval.esDeCobertura() {
		return len(grupo.deLasEvalsDeCobertura()) == 1, len(grupo.sinLaDeclaracion()) == 1
	}

	return len(grupo.deLasEvalsDePlazo()) == 1, len(grupo.sinLaFecha()) == 1
}

// probarElRepartoDeLasLecturasSinServir fija el reparto de data-model §4 de
// H26.1, filas 1, 3, 4, 5 y 7, con la 11 y, además del art. 124 de la LPAC, su
// índice sin servir: la invocación que no consulta no va a ninguna lista,
// aunque pida una lectura sin servir; la que consulta una va a
// lecturas_sin_servir termine como termine —con 0, con 4, sin código o con 5—,
// por articulo o por articulos con el bloque entre los que pide; y de las demás,
// la que quedó sin código o terminó con 0 no va a ninguna, la del 4 o el 5 va a
// fuera_de_lo_grabado —el metadatos de la misma norma o el mismo bloque de otra
// norma, también— y la de otro código, a otras_fallidas. La sesión pasa, y cada
// invocación está en invocaciones. Igual con la 11 del contrato y con la del
// repositorio.
func probarElRepartoDeLasLecturasSinServir(t *testing.T) {
	t.Parallel()

	for _, deUnOrigen := range lasDosEvals(t, ficheroDeLaEval11, contenidoDeLaEval11) {
		t.Run(deUnOrigen.origen, func(t *testing.T) {
			t.Parallel()

			eval := deUnOrigen.eval
			eval.SinServir = slices.Concat(eval.SinServir,
				[]LecturaSinServir{{Norma: normaDeLasTrazas, Verbo: "indice"}})

			repartirLasLecturasSinServir(t, eval)
		})
	}
}

// repartirLasLecturasSinServir juzga con la eval dada, la 11 con el índice de
// la LPAC también sin servir, la sesión del reparto de
// probarElRepartoDeLasLecturasSinServir y comprueba cada lista.
func repartirLasLecturasSinServir(t *testing.T, eval Eval) {
	t.Helper()

	lecturas := []Invocacion{
		deLaLPAC(t, codigoDeSalida(0), "articulo", "a124", "--fecha", diaDeLaNotificacion, "--describe"),
		deLaLPAC(t, codigoDeSalida(0), "indice", "--json"),
		deLaLPAC(t, codigoDeSalida(4), "articulos", "a123", "a124", "--json"),
		deLaLPAC(t, nil, "articulo", "a124", "--json"),
		deLaLPAC(t, codigoDeSalida(5), "articulo", "a124", "--fecha", diaDeLaNotificacion, "--json"),
		deLaLPAC(t, nil, "articulo", "a125", "--json"),
		deLaLPAC(t, codigoDeSalida(0), "articulo", "a125", "--json"),
		deLaLPAC(t, codigoDeSalida(5), "articulo", "a30", "--json"),
		deLaLPAC(t, codigoDeSalida(4), "metadatos", "--json"),
		invocada(t, codigoDeSalida(4), []string{kitlegalDelPath, "boe", "articulo", normaDeLaLRJSP, "a124", "--json"}),
		deLaLPAC(t, codigoDeSalida(2), "articulo", "a31", "--json"),
	}

	sesion := sesionDePlazosYTerritorio(respuestaDeLa11,
		slices.Concat([]Invocacion{resuelveLeganes(t), calculaLaReposicion(t)}, lecturas)...)
	resultado := Juzgar(eval, sesion, skillDeTerritorio)

	assert.True(t, resultado.Pasa, "la sesión pasa: %v", resultado.Motivos)
	assert.Len(t, resultado.Invocaciones, 2+len(lecturas), "cada invocación está en invocaciones")
	assert.Equal(t, []InvocacionFallida{
		{Orden: "boe indice BOE-A-2015-10565 --json", Codigo: codigoDeSalida(0)},
		{Orden: "boe articulos BOE-A-2015-10565 a123 a124 --json", Codigo: codigoDeSalida(4)},
		{Orden: "boe articulo BOE-A-2015-10565 a124 --json"},
		{Orden: ordenDelArticulo124, Codigo: codigoDeSalida(5)},
	}, resultado.LecturasSinServir)
	assert.Equal(t, []InvocacionFallida{
		{Orden: "boe articulo BOE-A-2015-10565 a30 --json", Codigo: codigoDeSalida(5)},
		{Orden: "boe metadatos BOE-A-2015-10565 --json", Codigo: codigoDeSalida(4)},
		{Orden: "boe articulo BOE-A-2015-10566 a124 --json", Codigo: codigoDeSalida(4)},
	}, resultado.FueraDeLoGrabado)
	assert.Equal(t, []InvocacionFallida{
		{Orden: "boe articulo BOE-A-2015-10565 a31 --json", Codigo: codigoDeSalida(2)},
	}, resultado.OtrasFallidas)
}

// probarLaEvalSinServirDelPlazo fija el juicio de la 11 en el modo orden con la
// sesión que resuelve Leganés, pide plazos calcular reposicion y lee el art. 124
// de una forma u otra: su comando de bloque lo satisface la lectura en su fecha
// que sale con 4 o 5, y no la que sale con 1, la que lo lee sin --fecha o en
// otra fecha, ni la sesión que no lo lee, que llevan el motivo del comando
// ausente; la lectura, si la hay, está en lecturas_sin_servir con su código, y
// nada en fuera_de_lo_grabado ni en otras_fallidas; con la línea del plazo y el
// comando satisfecho pasa sin contar en plazo_sin_fecha, y sin la línea cuenta y
// no pasa. Igual con la 11 del contrato y con la del repositorio.
func probarLaEvalSinServirDelPlazo(t *testing.T) {
	t.Parallel()

	ausente := "comando ausente: " + textoDelComando124
	enSuFecha := []string{"--fecha", diaDeLaNotificacion, "--json"}

	casos := []struct {
		nombre    string
		codigo    int
		lectura   []string
		respuesta string
		lecturas  []InvocacionFallida
		motivos   []string
		cuenta    bool
	}{
		{
			nombre: "en-su-fecha-con-el-5", codigo: 5, lectura: enSuFecha, respuesta: respuestaDeLa11,
			lecturas: []InvocacionFallida{{Orden: ordenDelArticulo124, Codigo: codigoDeSalida(5)}},
		},
		{
			nombre: "en-su-fecha-con-el-4", codigo: 4, lectura: enSuFecha, respuesta: respuestaDeLa11,
			lecturas: []InvocacionFallida{{Orden: ordenDelArticulo124, Codigo: codigoDeSalida(4)}},
		},
		{
			nombre: "en-su-fecha-con-el-1", codigo: 1, lectura: enSuFecha, respuesta: respuestaDeLa11,
			lecturas: []InvocacionFallida{{Orden: ordenDelArticulo124, Codigo: codigoDeSalida(1)}},
			motivos:  []string{ausente},
		},
		{
			nombre: "sin-fecha", codigo: 5, lectura: []string{"--json"}, respuesta: respuestaDeLa11,
			lecturas: []InvocacionFallida{{Orden: "boe articulo BOE-A-2015-10565 a124 --json", Codigo: codigoDeSalida(5)}},
			motivos:  []string{ausente},
		},
		{
			nombre: "en-otra-fecha", codigo: 5, lectura: []string{"--fecha", "2026-09-10", "--json"}, respuesta: respuestaDeLa11,
			lecturas: []InvocacionFallida{
				{Orden: "boe articulo BOE-A-2015-10565 a124 --fecha 2026-09-10 --json", Codigo: codigoDeSalida(5)},
			},
			motivos: []string{ausente},
		},
		{nombre: "sin-leerla", respuesta: respuestaDeLa11, motivos: []string{ausente}},
		{
			nombre: "sin-la-linea-del-plazo", codigo: 5, lectura: enSuFecha,
			respuesta: "Hasta el 13 de octubre de 2026, un mes desde el día siguiente a la notificación.",
			lecturas:  []InvocacionFallida{{Orden: ordenDelArticulo124, Codigo: codigoDeSalida(5)}},
			motivos:   []string{"plazo ausente: " + fechaDelPlazo}, cuenta: true,
		},
	}

	for _, deUnOrigen := range lasDosEvals(t, ficheroDeLaEval11, contenidoDeLaEval11) {
		for _, caso := range casos {
			t.Run(caso.nombre+"-"+deUnOrigen.origen, func(t *testing.T) {
				t.Parallel()

				invocaciones := []Invocacion{resuelveLeganes(t), calculaLaReposicion(t)}
				if caso.lectura != nil {
					invocaciones = append(invocaciones,
						deLaLPAC(t, codigoDeSalida(caso.codigo), "articulo", slices.Concat([]string{"a124"}, caso.lectura)...))
				}

				eval := deUnOrigen.eval
				resultado := Juzgar(eval, sesionDePlazosYTerritorio(caso.respuesta, invocaciones...), skillDeTerritorio)

				assert.Equal(t, caso.motivos, resultado.Motivos)
				assert.Equal(t, caso.motivos == nil, resultado.Pasa)
				assert.Equal(t, caso.lecturas, resultado.LecturasSinServir)
				assert.Empty(t, resultado.FueraDeLoGrabado)
				assert.Empty(t, resultado.OtrasFallidas)

				enElTotal, cuenta := enSuUmbralSinModelo(eval, resultado)
				require.True(t, enElTotal, "premisa: la 11 es de las del total de plazo_sin_fecha")
				assert.Equal(t, caso.cuenta, cuenta, "cuenta en plazo_sin_fecha")
			})
		}
	}
}

// probarLaEvalSinServirDelPlazoPorLaHerramienta fija el juicio de la 11 en el
// modo herramienta: la llamada a boe_articulo del art. 124 con su fecha y un
// sobre de la clase fuente-no-disponible o limite-o-tos satisface su comando de
// bloque y está en lecturas_sin_servir, y la sesión pasa sin contar en
// plazo_sin_fecha; sin fecha, con otra fecha o sin la llamada, el comando queda
// ausente con su motivo. Igual con la 11 del contrato y con la del repositorio.
func probarLaEvalSinServirDelPlazoPorLaHerramienta(t *testing.T) {
	t.Parallel()

	ausente := []string{"comando ausente: " + textoDelComando124}

	const (
		sinFecha    = `{"norma":"BOE-A-2015-10565","bloque":"a124"}`
		enOtraFecha = `{"norma":"BOE-A-2015-10565","bloque":"a124","fecha":"2026-09-10"}`
	)

	casos := []struct {
		nombre   string
		lectura  Llamada
		lecturas []InvocacionFallida
		motivos  []string
	}{
		{
			nombre:   "en-su-fecha-con-fuente-no-disponible",
			lectura:  llamadaFallida(herramientaDeArticulo, argumentosDelArticulo124, schema.ClaseFuenteNoDisponible),
			lecturas: []InvocacionFallida{{Orden: llamadaDelArticulo124, Codigo: codigoDeSalida(4)}},
		},
		{
			nombre:   "en-su-fecha-con-limite-o-tos",
			lectura:  llamadaFallida(herramientaDeArticulo, argumentosDelArticulo124, schema.ClaseLimiteOTos),
			lecturas: []InvocacionFallida{{Orden: llamadaDelArticulo124, Codigo: codigoDeSalida(5)}},
		},
		{
			nombre:   "sin-fecha",
			lectura:  llamadaFallida(herramientaDeArticulo, sinFecha, schema.ClaseFuenteNoDisponible),
			lecturas: []InvocacionFallida{{Orden: "boe_articulo BOE-A-2015-10565 a124", Codigo: codigoDeSalida(4)}},
			motivos:  ausente,
		},
		{
			nombre:  "en-otra-fecha",
			lectura: llamadaFallida(herramientaDeArticulo, enOtraFecha, schema.ClaseFuenteNoDisponible),
			lecturas: []InvocacionFallida{
				{Orden: "boe_articulo --fecha=2026-09-10 BOE-A-2015-10565 a124", Codigo: codigoDeSalida(4)},
			},
			motivos: ausente,
		},
		{nombre: "sin-leerla", motivos: ausente},
	}

	for _, deUnOrigen := range lasDosEvals(t, ficheroDeLaEval11, contenidoDeLaEval11) {
		for _, caso := range casos {
			t.Run(caso.nombre+"-"+deUnOrigen.origen, func(t *testing.T) {
				t.Parallel()

				llamadas := []Llamada{
					llamadaCorrecta("territorio_resolver", `{"consulta":"Leganés"}`),
					llamadaCorrecta("plazos_calcular", argumentosDeReposicion),
				}
				if caso.lectura.Herramienta != "" {
					llamadas = append(llamadas, caso.lectura)
				}

				eval := deUnOrigen.eval
				sesion := cambiada(sesionDelModoHerramienta(t, respuestaDeLa11, llamadas...),
					func(s *Sesion) { s.SkillsActivadas = []string{skillDeTerritorio} })
				resultado := juzgarEnModo(eval, sesion, skillDeTerritorio, ModoHerramienta)

				assert.Equal(t, caso.motivos, resultado.Motivos)
				assert.Equal(t, caso.motivos == nil, resultado.Pasa)
				assert.Equal(t, caso.lecturas, resultado.LecturasSinServir)
				assert.Empty(t, resultado.FueraDeLoGrabado)
				assert.Empty(t, resultado.OtrasFallidas)

				_, cuenta := enSuUmbralSinModelo(eval, resultado)
				assert.False(t, cuenta, "no cuenta en plazo_sin_fecha")
			})
		}
	}
}

// probarLaEvalSinServirDeLaCobertura fija el juicio de la 12: la sesión que
// resuelve Leganés, intenta leer la DA 1ª, que sale con 5 y está en
// lecturas_sin_servir, y responde con la declaración de no cubierto de la LPAC
// y sin línea del plazo, pasa sin contar en cobertura_sin_declarar; sin la
// declaración, cuenta y no pasa; y la que ejecuta plazos calcular no pasa, con
// el motivo del comando prohibido. Igual con la 12 del contrato y con la del
// repositorio.
func probarLaEvalSinServirDeLaCobertura(t *testing.T) {
	t.Parallel()

	laDA := []InvocacionFallida{{Orden: "boe articulo BOE-A-2015-10565 daprimera --json", Codigo: codigoDeSalida(5)}}

	casos := []struct {
		nombre    string
		calcula   bool
		respuesta string
		motivos   []string
		cuenta    bool
	}{
		{nombre: "con-la-declaración", respuesta: respuestaDeLa12},
		{
			nombre: "sin-la-declaración", respuesta: "Tienes un mes desde la notificación.",
			motivos: []string{"declaración de no cubierto ausente: " + normaDeLasTrazas}, cuenta: true,
		},
		{
			nombre: "con-plazos-calcular", calcula: true, respuesta: respuestaDeLa12,
			motivos: []string{"comando prohibido ejecutado: plazos calcular"},
		},
	}

	for _, deUnOrigen := range lasDosEvals(t, ficheroDeLaEval12, contenidoDeLaEval12) {
		for _, caso := range casos {
			t.Run(caso.nombre+"-"+deUnOrigen.origen, func(t *testing.T) {
				t.Parallel()

				invocaciones := []Invocacion{
					resuelveLeganes(t), deLaLPAC(t, codigoDeSalida(5), "articulo", "daprimera", "--json"),
				}
				if caso.calcula {
					invocaciones = append(invocaciones, calculaLaReposicion(t))
				}

				eval := deUnOrigen.eval
				resultado := Juzgar(eval, sesionDePlazosYTerritorio(caso.respuesta, invocaciones...), skillDeTerritorio)

				assert.Equal(t, caso.motivos, resultado.Motivos)
				assert.Equal(t, caso.motivos == nil, resultado.Pasa)
				assert.Equal(t, laDA, resultado.LecturasSinServir)
				assert.Empty(t, resultado.FueraDeLoGrabado)
				assert.Empty(t, resultado.OtrasFallidas)

				enElTotal, cuenta := enSuUmbralSinModelo(eval, resultado)
				require.True(t, enElTotal, "premisa: la 12 es de las del total de cobertura_sin_declarar")
				assert.Equal(t, caso.cuenta, cuenta, "cuenta en cobertura_sin_declarar")
			})
		}
	}
}

// probarLaSesionCortadaConLaLecturaSinServir fija que la sesión de la 11 que el
// tope cortó mientras leía el art. 124 en su fecha lleva esa lectura, sin
// código, en lecturas_sin_servir, que el informe escribe con codigo null, y
// nada en fuera_de_lo_grabado ni en otras_fallidas; la lectura sin código no
// satisface el comando, que queda ausente. Una sesión que no consulta ninguna
// lectura sin servir escribe lecturas_sin_servir vacía, detrás de
// otras_fallidas. Igual con la 11 del contrato y con la del repositorio.
func probarLaSesionCortadaConLaLecturaSinServir(t *testing.T) {
	t.Parallel()

	for _, deUnOrigen := range lasDosEvals(t, ficheroDeLaEval11, contenidoDeLaEval11) {
		t.Run(deUnOrigen.origen, func(t *testing.T) {
			t.Parallel()

			juzgarLaSesionCortada(t, deUnOrigen.eval)
		})
	}
}

// juzgarLaSesionCortada juzga con la eval dada, la 11, la sesión cortada y la
// que no consulta ninguna lectura sin servir de
// probarLaSesionCortadaConLaLecturaSinServir y comprueba lo que publican.
func juzgarLaSesionCortada(t *testing.T, eval Eval) {
	t.Helper()

	cortada := cambiada(sesionDePlazosYTerritorio(respuestaDeLa11, resuelveLeganes(t), calculaLaReposicion(t),
		deLaLPAC(t, nil, "articulo", "a124", "--fecha", diaDeLaNotificacion, "--json")), func(s *Sesion) {
		s.Codigo, s.Fin, s.Terminada, s.Cortada = 124, "assistant", false, true
		s.MotivoSinTerminar = "tope de 240 s agotado (código 124)"
	})

	resultado := Juzgar(eval, cortada, skillDeTerritorio)

	assert.False(t, resultado.Pasa)
	assert.Contains(t, resultado.Motivos, "comando ausente: "+textoDelComando124)
	assert.Equal(t, []InvocacionFallida{{Orden: ordenDelArticulo124}}, resultado.LecturasSinServir)
	assert.Empty(t, resultado.FueraDeLoGrabado)
	assert.Empty(t, resultado.OtrasFallidas)

	escrita, err := json.Marshal(resultado.LecturasSinServir)
	require.NoError(t, err)
	assert.JSONEq(t, `[{"orden":"`+ordenDelArticulo124+`","codigo":null}]`, string(escrita))

	sinNinguna := Juzgar(eval, sesionDePlazosYTerritorio(respuestaDeLa11, resuelveLeganes(t), calculaLaReposicion(t)),
		skillDeTerritorio)
	escrito, err := json.Marshal(sinNinguna)
	require.NoError(t, err)
	assert.Contains(t, string(escrito), `"otras_fallidas":[],"lecturas_sin_servir":[],"llegadas_a_la_red":`,
		"lecturas_sin_servir va vacía, detrás de otras_fallidas")
}

// probarLasLecturasDeLaEval16 fija que la 16 de normativa del repositorio,
// que lleva a leer a26 y a27, juzga una sesión que no los lee igual que la misma
// eval sin sus lecturas (US3, escenario 2; FR-007): la que lee el art. 25 y lo
// cita, que pasa; la que no lee nada, que no; y la que lee el 25 y el 28, que
// sale con 4, y lo cita, que pasa con el 28 fuera de lo grabado.
func probarLasLecturasDeLaEval16(t *testing.T) {
	t.Parallel()

	contenido, err := os.ReadFile(filepath.Join(evalsDeNormativa, ficheroDeLaEval16))
	require.NoError(t, err)

	conLecturas := evalLeida(t, ficheroDeLaEval16, string(contenido))
	require.Equal(t, []LecturaDeLaEval{
		{Norma: normaDeLaLRJSP, Bloque: "a26", Motivo: motivoDeLasLecturasDeLa16},
		{Norma: normaDeLaLRJSP, Bloque: "a27", Motivo: motivoDeLasLecturasDeLa16},
	}, conLecturas.Lecturas, "premisa: la 16 del repositorio lleva a leer a26 y a27")

	sinLecturas := conLecturas
	sinLecturas.Lecturas = nil

	leeElBloqueDeLaLRJSP := func(codigo int, bloque string) Invocacion {
		return invocada(t, codigoDeSalida(codigo), deLaSkill("articulo", normaDeLaLRJSP, bloque, "--json"))
	}

	sesiones := []struct {
		nombre string
		sesion Sesion
		pasa   bool
	}{
		{
			nombre: "lee-el-25-y-lo-cita",
			sesion: sesionTerminada(true, "El art. 25 lo regula "+citaDel25+".", leeElBloqueDeLaLRJSP(0, "a25")),
			pasa:   true,
		},
		{nombre: "no-lee-nada", sesion: sesionTerminada(true, "No lo sé.")},
		{
			nombre: "lee-el-25-y-el-28",
			sesion: sesionTerminada(true, "El art. 25 lo regula "+citaDel25+".",
				leeElBloqueDeLaLRJSP(0, "a25"), leeElBloqueDeLaLRJSP(4, "a28")),
			pasa: true,
		},
	}

	for _, caso := range sesiones {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			sin := Juzgar(sinLecturas, caso.sesion, skillNormativa)
			require.Equal(t, caso.pasa, sin.Pasa, "premisa: sin lecturas, %v", sin.Motivos)

			assert.Equal(t, sin, Juzgar(conLecturas, caso.sesion, skillNormativa))
		})
	}
}

// Lo que ven las sesiones de TestJuzgarElMotivoDeLasFallidas
// (contracts/juicio-e-informe.md §1 y §3 de H26.1): la llamada a boe_indice de
// la LPAC como la presenta el informe; la lectura del índice de la traza, como
// la presenta el informe; la orden de Bash que la filtra con python3, que el
// runner no tiene, y el principio de su salida; y las otras dos órdenes de Bash
// de los casos: la del índice de otra norma, que no nombra la invocación, y otra
// del mismo índice, que también la nombra.
const (
	llamadaDelIndice        = "boe_indice BOE-A-2015-10565"
	ordenDeLaTrazaDelIndice = "boe indice BOE-A-2015-10565 --json"

	bashDelIndiceConPython = ordenDelIndice +
		` | python3 -c "import json,sys; d=json.load(sys.stdin); print(len(d['data']['bloques']))"`
	salidaSinPython = "Exit code 127\n/bin/bash: line 1: python3: command not found\n"

	bashDelIndiceDeOtraNorma = "kitlegal boe indice BOE-A-2015-10566 --json | python3 -m json.tool"
	bashDelMismoIndice       = "kitlegal boe indice BOE-A-2015-10565 | python3 -m json.tool"
)

// TestJuzgarElMotivoDeLasFallidas fija el motivo con el que el juicio publica un
// fallo que no es de kitlegal, sin que decida nada (data-model §4, filas 2 y 6,
// y §5 de H26.1; contracts/juicio-e-informe.md §1, §3 y §6 de H26.1; research
// D7, D8 y D10 de H26.1; FR-015, FR-023 a FR-028 de H26.1; SC-007):
//
//   - la llamada sin sobre va con codigo null y motivo sin_sobre en invocaciones
//     y en otras_fallidas, también si consulta una lectura de sin_servir, que
//     entonces no está en lecturas_sin_servir; y su fila de informe.md dice «sin
//     sobre» en la columna del código;
//   - la invocación de la traza que terminó con 1 y que enlaza un único uso de
//     Bash cuya salida trae la línea de la escritura fallida va con codigo 1,
//     motivo salida_no_escrita y la orden de Bash tal cual en otras_fallidas o,
//     si consulta una lectura de sin_servir, en lecturas_sin_servir;
//   - la de otra salida de error, sin un uso que la enlace, con dos o con otro
//     código, como hoy;
//   - y en todas, la sesión pasa o no, con los mismos motivos y los mismos
//     comandos, igual que la misma sesión sin lo que se lee para el motivo
//     (FR-028).
func TestJuzgarElMotivoDeLasFallidas(t *testing.T) {
	t.Parallel()

	t.Run("sin-sobre", probarElMotivoDeLaLlamadaSinSobre)
	t.Run("salida-no-escrita", probarElMotivoDeLaSalidaNoEscrita)
}

// probarElMotivoDeLaLlamadaSinSobre fija la fila 2 del reparto con la sesión de
// la eval 01 que pasa y, detrás, una llamada a boe_indice que el agente no tiene
// y que le devuelve su aviso como un error: con la eval tal cual y con el índice
// de la LPAC en su sin_servir, que la llamada consulta.
func probarElMotivoDeLaLlamadaSinSobre(t *testing.T) {
	t.Parallel()

	indice := func(sinSobre bool) Llamada {
		return Llamada{
			Herramienta: "boe_indice", Argumentos: []byte(argumentosDelIndice), ConResultado: true, Error: true,
			SinSobre: sinSobre,
		}
	}

	casos := []struct {
		nombre    string
		sinServir []LecturaSinServir
	}{
		{nombre: "sin-sin-servir"},
		{nombre: "con-el-indice-sin-servir", sinServir: []LecturaSinServir{{Norma: normaDeLasTrazas, Verbo: "indice"}}},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			eval := evalDelArticulo21()
			eval.SinServir = caso.sinServir

			conSobre := cambiada(sesionQuePasa(t), func(s *Sesion) { s.Llamadas = []Llamada{indice(false)} })
			sinSobre := cambiada(sesionQuePasa(t), func(s *Sesion) { s.Llamadas = []Llamada{indice(true)} })

			if caso.sinServir != nil {
				require.Equal(t, []InvocacionFallida{{Orden: llamadaDelIndice, Codigo: codigoDeSalida(1)}},
					Juzgar(eval, conSobre, skillNormativa).LecturasSinServir,
					"premisa: sin la marca, la llamada es una lectura de sin_servir")
			}

			resultado := Juzgar(eval, sinSobre, skillNormativa)

			assert.Equal(t, []InvocacionInformada{
				{Orden: ordenDelArticulo21, Codigo: codigoDeSalida(0)},
				{Orden: llamadaDelIndice, Motivo: "sin_sobre", Llamada: true},
			}, resultado.Invocaciones)
			assert.Equal(t, []InvocacionFallida{{Orden: llamadaDelIndice, Motivo: "sin_sobre"}}, resultado.OtrasFallidas)
			assert.Empty(t, resultado.LecturasSinServir, "la llamada sin sobre no llega a la lectura")
			assert.Empty(t, resultado.FueraDeLoGrabado)

			exigirLoEscrito(t, resultado.Invocaciones[1],
				`{"orden":"`+llamadaDelIndice+`","codigo":null,"motivo":"sin_sobre","conexiones":[],"llamada":true}`)
			exigirLoEscrito(t, resultado.OtrasFallidas,
				`[{"orden":"`+llamadaDelIndice+`","codigo":null,"motivo":"sin_sobre"}]`)
			assert.Equal(t, []string{llamadaDelIndice, "sin sobre", "sin conexiones", "sí"},
				filasDeInvocaciones(resultado.Invocaciones)[1], "su fila de informe.md dice «sin sobre»")

			exigirElMismoJuicio(t, Juzgar(eval, conSobre, skillNormativa), resultado)
		})
	}
}

// probarElMotivoDeLaSalidaNoEscrita fija la fila 6 del reparto con la sesión de
// la eval 01 que pasa y, delante de su lectura, la del índice de la LPAC, que
// termina con el código del caso, con las órdenes de Bash del caso: solo la que
// la enlaza sola y trae la línea de la escritura fallida le da el motivo, en
// otras_fallidas o, con el índice en su sin_servir, en lecturas_sin_servir.
func probarElMotivoDeLaSalidaNoEscrita(t *testing.T) {
	t.Parallel()

	linea := lineaDeLaEscrituraFallida(t)
	sinPython := Texto{Orden: bashDelIndiceConPython, Salida: salidaSinPython + linea}
	indiceSinServir := []LecturaSinServir{{Norma: normaDeLasTrazas, Verbo: "indice"}}

	enlazada := []InvocacionFallida{{
		Orden: ordenDeLaTrazaDelIndice, Codigo: codigoDeSalida(1), Motivo: "salida_no_escrita",
		Bash: bashDelIndiceConPython,
	}}
	comoHoy := []InvocacionFallida{{Orden: ordenDeLaTrazaDelIndice, Codigo: codigoDeSalida(1)}}

	casos := []struct {
		nombre    string
		codigo    int
		sinServir []LecturaSinServir
		ordenes   []Texto

		otras, lecturas, fuera []InvocacionFallida

		// escritas son las tres listas, juntas, como las escribe el informe.
		escritas string
	}{
		{
			nombre: "enlazada", codigo: 1, ordenes: []Texto{sinPython}, otras: enlazada,
			escritas: `[{"orden":"` + ordenDeLaTrazaDelIndice + `","codigo":1,"motivo":"salida_no_escrita","bash":` +
				cadenaJSON(t, bashDelIndiceConPython) + `}]`,
		},
		{
			nombre: "enlazada-con-el-indice-sin-servir", codigo: 1, sinServir: indiceSinServir,
			ordenes: []Texto{sinPython}, lecturas: enlazada,
			escritas: `[{"orden":"` + ordenDeLaTrazaDelIndice + `","codigo":1,"motivo":"salida_no_escrita","bash":` +
				cadenaJSON(t, bashDelIndiceConPython) + `}]`,
		},
		{
			nombre: "con-otra-salida-de-error", codigo: 1,
			ordenes: []Texto{{Orden: bashDelIndiceConPython, Salida: salidaSinPython}}, otras: comoHoy,
			escritas: `[{"orden":"` + ordenDeLaTrazaDelIndice + `","codigo":1}]`,
		},
		{
			nombre: "sin-uso-que-la-enlace", codigo: 1,
			ordenes: []Texto{{Orden: bashDelIndiceDeOtraNorma, Salida: salidaSinPython + linea}}, otras: comoHoy,
			escritas: `[{"orden":"` + ordenDeLaTrazaDelIndice + `","codigo":1}]`,
		},
		{
			nombre: "con-dos-usos-que-la-enlazan", codigo: 1,
			ordenes: []Texto{sinPython, {Orden: bashDelMismoIndice, Salida: linea}}, otras: comoHoy,
			escritas: `[{"orden":"` + ordenDeLaTrazaDelIndice + `","codigo":1}]`,
		},
		{
			nombre: "con-otro-codigo", codigo: codigoFuenteNoDisponible, ordenes: []Texto{sinPython},
			fuera:    []InvocacionFallida{{Orden: ordenDeLaTrazaDelIndice, Codigo: codigoDeSalida(codigoFuenteNoDisponible)}},
			escritas: `[{"orden":"` + ordenDeLaTrazaDelIndice + `","codigo":4}]`,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			eval := evalDelArticulo21()
			eval.SinServir = caso.sinServir

			sinLasOrdenes := sesionQuePasa(t,
				invocada(t, codigoDeSalida(caso.codigo), deKitlegal("boe", "indice", normaDeLasTrazas, "--json")))
			conLasOrdenes := cambiada(sinLasOrdenes, func(s *Sesion) { s.OrdenesDeBash = caso.ordenes })

			resultado := Juzgar(eval, conLasOrdenes, skillNormativa)

			assert.Equal(t, caso.otras, resultado.OtrasFallidas)
			assert.Equal(t, caso.lecturas, resultado.LecturasSinServir)
			assert.Equal(t, caso.fuera, resultado.FueraDeLoGrabado)
			assert.Equal(t, InvocacionInformada{Orden: ordenDeLaTrazaDelIndice, Codigo: codigoDeSalida(caso.codigo)},
				resultado.Invocaciones[0], "en invocaciones, como hoy")
			exigirLoEscrito(t, slices.Concat(resultado.OtrasFallidas, resultado.LecturasSinServir,
				resultado.FueraDeLoGrabado), caso.escritas)

			exigirElMismoJuicio(t, Juzgar(eval, sinLasOrdenes, skillNormativa), resultado)
		})
	}
}

// exigirLoEscrito exige que el valor se escriba en JSON como lo escrito.
func exigirLoEscrito(t *testing.T, valor any, escrito string) {
	t.Helper()

	codificado, err := json.Marshal(valor)
	require.NoError(t, err)
	assert.JSONEq(t, escrito, string(codificado))
}

// exigirElMismoJuicio exige que el resultado juzgue la sesión como el de la misma
// sesión sin lo que se lee de ella para el motivo de una fallida: si pasa, por
// qué no pasa y qué comandos esperados y prohibidos ejecuta (FR-028 de H26.1).
func exigirElMismoJuicio(t *testing.T, sinElMotivo, conElMotivo ResultadoDeEval) {
	t.Helper()

	assert.Equal(t, sinElMotivo.Pasa, conElMotivo.Pasa, "pasa igual")
	assert.Equal(t, sinElMotivo.Motivos, conElMotivo.Motivos, "con los mismos motivos")
	assert.Equal(t, sinElMotivo.ComandosEjecutados, conElMotivo.ComandosEjecutados)
	assert.Equal(t, sinElMotivo.ComandosAusentes, conElMotivo.ComandosAusentes)
	assert.Equal(t, sinElMotivo.ComandosProhibidosEjecutados, conElMotivo.ComandosProhibidosEjecutados)
}

// TestLineaDeLaSalidaNoEscrita ata el prefijo con el que el arnés reconoce la
// línea de la escritura fallida al binario que la escribe (research D8 de H26.1;
// FR-025 de H26.1; SC-007): app.Main lee el índice de la LPAC de una caché
// sembrada desde las grabaciones y, con una salida estándar que falla como una
// tubería sin lector, termina con 1, el código de lo inesperado, y deja en su
// salida de error una línea que empieza por el prefijo que el arnés pide al
// presentador, sin ninguna copia de su mensaje. Si el mensaje cambia, el prefijo
// cambia con él y este test sigue en verde: lo que exige el mensaje al binario
// es TestTuberiaCerrada (quickstart.md §7.b de H26.1).
func TestLineaDeLaSalidaNoEscrita(t *testing.T) {
	t.Parallel()

	dirCache := t.TempDir()

	faltas, err := Preparar(dirCache, UnionDeGrabaciones(),
		[]Consulta{{Applet: "boe", Verbo: "indice", Argumentos: []string{normaDeLasTrazas}}})
	require.NoError(t, err)
	require.Empty(t, faltas, "premisa: las grabaciones sirven el índice de la LPAC: %s", presentarFaltas(faltas))

	registro, err := registroDeBoe(t.TempDir(), cache.ConDirectorio(dirCache))
	require.NoError(t, err)

	var errores bytes.Buffer

	codigo := app.Main([]string{programaDeLasConsultas, "boe", "indice", normaDeLasTrazas, "--offline", "--json"},
		registro, salidaSinLector{}, &errores, sinDatosDeConstruccion, sinDatosDeConstruccion, sinDatosDeConstruccion)

	assert.Equal(t, 1, codigo, "la salida que no se escribe es lo inesperado: %s", errores.String())
	require.NotEmpty(t, prefijoDeLaSalidaNoEscrita(), "el presentador da el prefijo de la escritura fallida")
	assert.True(t, traeLaLineaDeLaSalidaNoEscrita(errores.String()),
		"la salida de error lleva una línea que empieza por %q: %q", prefijoDeLaSalidaNoEscrita(), errores.String())
}
