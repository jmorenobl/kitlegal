package evals

import (
	"cmp"
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
