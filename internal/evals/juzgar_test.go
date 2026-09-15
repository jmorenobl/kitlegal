package evals

import (
	"slices"
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
)

// juicio es una llamada a Juzgar y el resultado que el caso exige de ella.
type juicio struct {
	eval     Eval
	sesion   Sesion
	esperado ResultadoDeEval
}

// TestJuzgar fija la comparación mecánica de una sesión con su eval de
// data-model §10.2: pasa solo si la sesión terminó, la activación coincide y
// están todos los comandos y todas las citas esperados; un comando esperado lo
// satisface solo una invocación con consulta y código 0 del mismo applet, con
// articulo o articulos para un bloque, el mismo verbo y la misma norma para una
// consulta de norma y los términos como palabras para buscar; una cita cuenta
// solo con la misma norma y el mismo bloque; y las invocaciones fuera de lo
// grabado, las otras fallidas y las llegadas a la red se informan sin cambiar si
// pasa (contrato evals-y-grabaciones §6; FR-072, FR-076, SC-009; US4, escenarios
// 2, 3, 5, 7 y 8).
func TestJuzgar(t *testing.T) {
	t.Parallel()

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
			nombre:  "buscar-terminos-como-palabras",
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
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			for _, j := range caso.juicios {
				assert.Equal(t, j.esperado, Juzgar(j.eval, j.sesion, skillDeLasSesiones),
					"la eval %s con la sesión del caso", j.eval.Fichero)
			}
		})
	}
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

// terminosComoPalabras es el juicio de una eval con dos búsquedas esperadas. La
// primera la satisface la invocación cuyos argumentos, en minúsculas, contienen
// cada término como palabra, aunque el término y los argumentos no coincidan en
// mayúsculas, uno de ellos esté dentro de un argumento con espacios y otro lleve
// una barra. La segunda no la satisface la invocación que solo contiene su
// término dentro de otras palabras, detrás y delante de una letra (data-model
// §6.1).
func terminosComoPalabras(t *testing.T) juicio {
	t.Helper()

	eval := Eval{
		Fichero:  "03-lrbrl-busqueda.yaml",
		Pregunta: "¿Qué ley regula las bases del régimen local?",
		Activa:   true,
		Comandos: []ComandoEsperado{
			{Applet: "boe", Verbo: "buscar", Terminos: []string{"Régimen", "local", "7/1985"}},
			{Applet: "boe", Verbo: "buscar", Terminos: []string{"común"}},
		},
		Citas: []CitaEsperada{{Norma: normaDeLaLRBRL, Bloque: "a1"}},
	}

	respuesta := "Es la Ley 7/1985, reguladora de las Bases del Régimen Local [BOE-A-1985-5392, bloque a1]."

	return juicio{
		eval: eval,
		sesion: sesionTerminada(true, respuesta,
			invocada(t, codigoDeSalida(0), deLaSkill("buscar", "bases del RÉGIMEN local", "7/1985", "--json")),
			invocada(t, codigoDeSalida(0), deLaSkill("buscar", "procedimiento", "comúnmente", "intercomún", "--json"))),
		esperado: ResultadoDeEval{
			Eval:               eval.Fichero,
			Activa:             true,
			Activada:           true,
			ComandosEjecutados: []string{"boe buscar Régimen local 7/1985"},
			ComandosAusentes:   []string{"boe buscar común"},
			CitasEncontradas:   []string{"BOE-A-1985-5392 a1"},
			Invocaciones: []InvocacionInformada{
				{Orden: "boe buscar bases del RÉGIMEN local 7/1985 --json", Codigo: codigoDeSalida(0)},
				{Orden: "boe buscar procedimiento comúnmente intercomún --json", Codigo: codigoDeSalida(0)},
			},
			Respuesta:        respuesta,
			CodigoDeLaSesion: codigoDeSalida(0),
			FinDeLaSesion:    "result success",
			SesionTerminada:  true,
			Motivos:          []string{"comando ausente: boe buscar común"},
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
