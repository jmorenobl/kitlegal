package instalacion_test

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Las clases de hallazgo de FR-065 con un nombre corto, para que cada fila
// quepa en una línea.
const (
	editado     = instalacion.HallazgoFicheroEditado
	colgando    = instalacion.HallazgoEnlaceColgando
	aOtroSitio  = instalacion.HallazgoEnlaceAOtroSitio
	esCopia     = instalacion.HallazgoCopia
	otraVersion = instalacion.HallazgoVersionDistinta
)

// Las rutas de las pruebas de doctor en el ámbito local: el directorio de cada
// skill en el directorio neutro y su entrada en el host claude.
const (
	neutroPlazos    = ".agents/skills/plazos-y-territorio"
	neutroNormativa = ".agents/skills/normativa"
	hostPlazos      = ".claude/skills/plazos-y-territorio"
	hostNormativa   = ".claude/skills/normativa"
)

// hallazgo es el hallazgo de esa clase, en esa ruta y con esa orden.
func hallazgo(clase instalacion.ClaseDeHallazgo, ruta, orden string) instalacion.Hallazgo {
	return instalacion.Hallazgo{Clase: clase, Ruta: ruta, Orden: orden}
}

// instala es la orden de install que nombra esas skills, en el ámbito local.
func instala(skills ...string) string {
	return strings.Join(append([]string{"kitlegal skills install"}, skills...), " ")
}

// conHostClaude es orden con --host claude detrás.
func conHostClaude(orden string) string {
	return orden + " --host claude"
}

// rmY y rmRY son orden precedida de la retirada de ruta, que no lleva ninguna
// comilla simple: sin -r y con -r.
func rmY(ruta, orden string) string {
	return "rm -- '" + ruta + "' && " + orden
}

func rmRY(ruta, orden string) string {
	return "rm -r -- '" + ruta + "' && " + orden
}

// casoDeDoctor es una fila de TestHallazgos y de TestOrdenesDeDoctor: un disco
// con una instalación, una invocación de doctor y lo que tiene que encontrar
// en él. Sobre las mismas filas, TestOrdenesDeDoctorArreglan ejecuta las
// órdenes de cada hallazgo.
type casoDeDoctor struct {
	nombre string
	// preparar deja el disco del caso, que empieza con el directorio de
	// trabajo vacío y el temporal.
	preparar func(d *discoEnMemoria)
	// invocacion es la de doctor.
	invocacion instalacion.Invocacion
	// home es el valor de HOME; sin él, homeDePrueba.
	home string
	// admite dice en qué directorio funciona el creador de enlaces, que
	// sondea y enlaza con la misma respuesta; sin él, en todos.
	admite func(directorio string) bool
	// version es la del binario; sin ella, versionDePrueba.
	version string
	// esperados tiene cada hallazgo, en el orden en que se nombra; ninguno si
	// doctor sale con 0.
	esperados []instalacion.Hallazgo
	// sondas tiene cada directorio por el que se pregunta Disponible, en
	// orden.
	sondas []string
	// noExaminadas son rutas de las que no se puede examinar nada, ni ellas
	// ni lo que cuelga de ellas.
	noExaminadas []string
	// excepcion es, si la hay, la ruta de una de las excepciones de FR-066
	// que el install de la orden del primer hallazgo comprueba: con ella no
	// valen las garantías de FR-066 y ese install la nombra como conflicto.
	excepcion string
}

// homeDelCaso es el HOME del caso.
func (c casoDeDoctor) homeDelCaso() string {
	if c.home == "" {
		return homeDePrueba
	}

	return c.home
}

// versionDelCaso es la versión del binario del caso.
func (c casoDeDoctor) versionDelCaso() string {
	if c.version == "" {
		return versionDePrueba
	}

	return c.version
}

// preparado es el disco del caso, recién preparado, y el ámbito de su
// invocación.
func (c casoDeDoctor) preparado(t *testing.T) (*discoEnMemoria, instalacion.Ambito) {
	t.Helper()

	d := nuevoDiscoEnMemoria(t)
	d.directorio(temporal)
	c.preparar(d)

	pedido, err := instalacion.ValidarInvocacion(c.invocacion, c.homeDelCaso(), empotradasDePrueba())
	require.NoError(t, err, "la invocación del caso")

	return d, pedido.Ambito
}

// enlazadorEn es un creador de enlaces nuevo sobre d, el del caso.
func (c casoDeDoctor) enlazadorEn(d *discoEnMemoria) *enlazadorEnMemoria {
	if c.admite == nil {
		return nuevoEnlazadorHonesto(d, admiteSiempre)
	}

	return nuevoEnlazadorHonesto(d, c.admite)
}

// diagnosticar es una invocación nueva de doctor sobre d, en el ámbito, con
// su propio creador de enlaces, que devuelve.
func (c casoDeDoctor) diagnosticar(
	d *discoEnMemoria, ambito instalacion.Ambito,
) (instalacion.Diagnostico, *enlazadorEnMemoria, error) {
	enlazador := c.enlazadorEn(d)
	diagnostico, err := instalacion.Diagnosticar(d, enlazador, ambito, empotradasDePrueba(), c.versionDelCaso())

	return diagnostico, enlazador, err
}

// grupoDeDoctor es un grupo de filas de TestHallazgos.
type grupoDeDoctor struct {
	nombre string
	casos  []casoDeDoctor
}

// gruposDeDoctor son las filas de TestHallazgos: una por cada clase de
// hallazgo de data-model §6, en el directorio neutro, dentro de una skill, en
// los hosts claude y antigravity y en una copia de host, y por cada regla de
// las versiones.
func gruposDeDoctor() []grupoDeDoctor {
	return []grupoDeDoctor{
		{nombre: "directorio neutro", casos: casosDoctorDelNeutro()},
		{nombre: "dentro de una skill", casos: casosDoctorDentroDeUnaSkill()},
		{nombre: "host claude", casos: casosDoctorDelHost()},
		{nombre: "host antigravity", casos: casosDoctorDeAntigravity()},
		{nombre: "copia de host", casos: casosDoctorDeLaCopia()},
		{nombre: "versiones", casos: casosDoctorDeLasVersiones()},
	}
}

// TestHallazgos fija doctor (data-model §6; contracts/applet-skills.md §4.3,
// §5 y §6; FR-036, FR-065 a FR-069, FR-077; SC-011, SC-012): sobre las skills
// declaradas y empotradas, y solo sobre ellas, cada una de las cinco clases de
// hallazgo con su ruta y la orden que lo arregla, en el orden de FR-066. En
// cada fila:
//
//   - el diagnóstico llega siempre, sin error, con el manifiesto, su versión,
//     la del binario y exactamente cada hallazgo esperado, o la lista vacía:
//     encontrar algo es el resultado de la verificación, no un fallo (ADR
//     0023);
//   - nada cambia en el disco (FR-068);
//   - Disponible solo se pregunta por .claude/skills ante una copia de host
//     que sigue siendo un directorio real, una sola vez (FR-069);
//   - no se abre nada que no sea un fichero regular, no se lista nada y no se
//     examina nada por debajo de una entrada que no es un directorio real, la
//     única que se nombra (FR-028).
func TestHallazgos(t *testing.T) {
	t.Parallel()

	for _, grupo := range gruposDeDoctor() {
		t.Run(grupo.nombre, func(t *testing.T) {
			t.Parallel()

			probarCasosDeDoctor(t, grupo.casos)
		})
	}

	t.Run("sin manifiesto", probarDoctorSinManifiesto)
	t.Run("ámbito ilegible", func(t *testing.T) {
		t.Parallel()

		probarAmbitosIlegibles(t, "doctor", func(t *testing.T, d *discoEnMemoria, ambito instalacion.Ambito) error {
			t.Helper()

			enlazador := nuevoEnlazadorHonesto(d, admiteSiempre)

			diagnostico, err := instalacion.Diagnosticar(d, enlazador, ambito, empotradasDePrueba(), versionDePrueba)
			assert.Equal(t, instalacion.Diagnostico{}, diagnostico, "ningún hallazgo")
			assert.Empty(t, enlazador.preguntados, "sin sonda")

			return err
		})
	})
	t.Run("fallos de entrada y salida", probarFallosDelDoctor)
	t.Run("versión del binario que no se puede declarar", probarVersionDelDoctor)
	t.Run("la salida", probarSalidaDeDoctor)
}

// probarCasosDeDoctor comprueba cada caso en su propio disco.
func probarCasosDeDoctor(t *testing.T, casos []casoDeDoctor) {
	t.Helper()

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			probarCasoDeDoctor(t, caso)
		})
	}
}

// probarCasoDeDoctor diagnostica en el disco del caso y lo compara con lo
// esperado.
func probarCasoDeDoctor(t *testing.T, caso casoDeDoctor) {
	t.Helper()

	d, ambito := caso.preparado(t)
	antes := d.instantanea()

	diagnostico, enlazador, err := caso.diagnosticar(d, ambito)

	// Encontrar algo es el resultado de una verificación que ha funcionado: el
	// diagnóstico llega entero, con los hallazgos como datos, y nunca como
	// fallo (ADR 0023).
	require.NoError(t, err, "con hallazgos o sin ellos, doctor no falla")
	assert.Equal(t, ambito.Neutro(), diagnostico.Directorio)
	assert.True(t, diagnostico.Manifiesto)
	assert.NotNil(t, diagnostico.Version, "con manifiesto, su versión (FR-067)")
	assert.Equal(t, caso.versionDelCaso(), diagnostico.VersionDelBinario)
	assert.Equal(t, hallazgosEsperados(caso.esperados), diagnostico.Hallazgos,
		"primero los que llevan rm; en cada grupo, por ruta y por clase (FR-066)")

	assert.Equal(t, antes, d.instantanea(), "doctor no deja ningún cambio en disco (FR-068)")
	assert.Equal(t, caso.sondas, enlazador.preguntados, "Disponible, solo por .claude/skills ante una copia y una vez")
	exigirDiscoRespetado(t, d, raizVigilada(ambito))
	exigirNoExaminadas(t, d, caso.noExaminadas...)
}

// hallazgosEsperados es la lista que el diagnóstico tiene que traer: la del
// caso, o la lista vacía —nunca nula, que en JSON sería null— si el caso no
// espera ninguno.
func hallazgosEsperados(esperados []instalacion.Hallazgo) []instalacion.Hallazgo {
	if esperados == nil {
		return []instalacion.Hallazgo{}
	}

	return esperados
}

// casosDoctorDelNeutro son los de una skill del directorio neutro, en local y
// sin host: un fichero declarado editado, que falta o que ya no es un fichero
// regular, y el directorio de la skill que no es un directorio real o falta.
func casosDoctorDelNeutro() []casoDeDoctor {
	instaladas := func(d *discoEnMemoria) { instalarLocal(d, "normativa", "plazos-y-territorio").escribir() }
	plazos := instala("plazos-y-territorio")

	return []casoDeDoctor{
		{nombre: "una instalación sin tocar", preparar: instaladas},
		{
			nombre: "un fichero editado",
			preparar: func(d *discoEnMemoria) {
				instaladas(d)
				d.fichero(neutroPlazos+"/SKILL.md", "editado a mano")
			},
			esperados: []instalacion.Hallazgo{hallazgo(editado, neutroPlazos+"/SKILL.md", rmY(neutroPlazos+"/SKILL.md", plazos))},
		},
		{
			nombre: "un fichero que falta",
			preparar: func(d *discoEnMemoria) {
				instaladas(d)
				d.retirar(neutroPlazos + "/references/leyes_vertebrales.md")
			},
			esperados: []instalacion.Hallazgo{hallazgo(editado, neutroPlazos+"/references/leyes_vertebrales.md", plazos)},
		},
		{
			nombre: "un fichero que es un enlace a una copia idéntica",
			preparar: func(d *discoEnMemoria) {
				instaladas(d)
				d.fichero(temporal+"/SKILL.md", "# plazos-y-territorio\n")
				d.enlace(neutroPlazos+"/SKILL.md", temporal+"/SKILL.md")
			},
			esperados:    []instalacion.Hallazgo{hallazgo(editado, neutroPlazos+"/SKILL.md", rmY(neutroPlazos+"/SKILL.md", plazos))},
			noExaminadas: []string{temporal + "/SKILL.md"},
		},
		{
			nombre: "un fichero que es un directorio real",
			preparar: func(d *discoEnMemoria) {
				instaladas(d)
				d.retirar(neutroPlazos + "/SKILL.md")
				d.fichero(neutroPlazos+"/SKILL.md/nota.md", "no es de kitlegal")
			},
			esperados:    []instalacion.Hallazgo{hallazgo(editado, neutroPlazos+"/SKILL.md", rmRY(neutroPlazos+"/SKILL.md", plazos))},
			noExaminadas: []string{neutroPlazos + "/SKILL.md/nota.md"},
		},
		{
			nombre: "un fichero que es una tubería",
			preparar: func(d *discoEnMemoria) {
				instaladas(d)
				d.tuberia(neutroPlazos + "/SKILL.md")
			},
			esperados: []instalacion.Hallazgo{hallazgo(editado, neutroPlazos+"/SKILL.md", rmY(neutroPlazos+"/SKILL.md", plazos))},
		},
		{
			nombre: "la skill es un enlace colgando",
			preparar: func(d *discoEnMemoria) {
				instaladas(d)
				d.enlace(neutroPlazos, temporal+"/no-existe")
			},
			esperados: []instalacion.Hallazgo{hallazgo(colgando, neutroPlazos, rmY(neutroPlazos, plazos))},
		},
		{
			nombre: "la skill es un enlace a una copia idéntica fuera del proyecto",
			preparar: func(d *discoEnMemoria) {
				instalarEn(d, instalacion.NuevoAmbitoDir(temporal), "plazos-y-territorio")
				instaladas(d)
				d.enlace(neutroPlazos, temporal+"/plazos-y-territorio")
			},
			esperados:    []instalacion.Hallazgo{hallazgo(aOtroSitio, neutroPlazos, rmY(neutroPlazos, plazos))},
			noExaminadas: []string{temporal + "/plazos-y-territorio"},
		},
		{
			nombre: "la skill es un fichero",
			preparar: func(d *discoEnMemoria) {
				instaladas(d)
				d.fichero(neutroPlazos, "no soy un directorio")
			},
			esperados: []instalacion.Hallazgo{hallazgo(aOtroSitio, neutroPlazos, rmY(neutroPlazos, plazos))},
		},
		{
			nombre: "falta la skill entera",
			preparar: func(d *discoEnMemoria) {
				instaladas(d)
				d.retirar(neutroPlazos)
			},
			esperados: []instalacion.Hallazgo{
				hallazgo(editado, neutroPlazos+"/SKILL.md", plazos),
				hallazgo(editado, neutroPlazos+"/references/jerarquia_normativa.md", plazos),
				hallazgo(editado, neutroPlazos+"/references/leyes_vertebrales.md", plazos),
			},
		},
		{
			nombre: "lo que el manifiesto no declara no es un hallazgo",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "normativa", "plazos-y-territorio").
					olvidar("plazos-y-territorio", "plazos-y-territorio/references/leyes_vertebrales.md").escribir()
				d.fichero(neutroPlazos+"/mio.md", "no es de kitlegal")
				d.fichero(neutroPlazos+"/references/mio.md", "no es de kitlegal")
				d.enlace(neutroPlazos+"/otro", "no-existe")
				d.fichero(".agents/skills/ajena/SKILL.md", "no es de kitlegal")
			},
			noExaminadas: []string{
				neutroPlazos + "/mio.md", neutroPlazos + "/references/mio.md", neutroPlazos + "/otro",
				neutroPlazos + "/references/leyes_vertebrales.md", ".agents/skills/ajena",
			},
		},
	}
}

// casosDoctorDentroDeUnaSkill son los de un directorio intermedio de los
// ficheros declarados, references, que no es un directorio real o falta: el
// hallazgo es él, sin ninguno por los ficheros de debajo.
func casosDoctorDentroDeUnaSkill() []casoDeDoctor {
	instaladas := func(d *discoEnMemoria) { instalarLocal(d, "normativa", "plazos-y-territorio").escribir() }
	plazos := instala("plazos-y-territorio")
	references := neutroPlazos + "/references"

	return []casoDeDoctor{
		{
			nombre: "references es un enlace a una copia idéntica",
			preparar: func(d *discoEnMemoria) {
				instalarEn(d, instalacion.NuevoAmbitoDir(temporal), "plazos-y-territorio")
				instaladas(d)
				d.enlace(references, temporal+"/plazos-y-territorio/references")
			},
			esperados:    []instalacion.Hallazgo{hallazgo(aOtroSitio, references, rmY(references, plazos))},
			noExaminadas: []string{temporal + "/plazos-y-territorio"},
		},
		{
			nombre: "references es un enlace colgando",
			preparar: func(d *discoEnMemoria) {
				instaladas(d)
				d.enlace(references, "no-existe")
			},
			esperados: []instalacion.Hallazgo{hallazgo(colgando, references, rmY(references, plazos))},
		},
		{
			nombre: "references es un fichero",
			preparar: func(d *discoEnMemoria) {
				instaladas(d)
				d.fichero(references, "no soy un directorio")
			},
			esperados: []instalacion.Hallazgo{hallazgo(aOtroSitio, references, rmY(references, plazos))},
		},
		{
			nombre: "falta references",
			preparar: func(d *discoEnMemoria) {
				instaladas(d)
				d.retirar(references)
			},
			esperados: []instalacion.Hallazgo{
				hallazgo(editado, references+"/jerarquia_normativa.md", plazos),
				hallazgo(editado, references+"/leyes_vertebrales.md", plazos),
			},
		},
		{
			nombre: "un manifiesto legible que declara SKILL.md como fichero y como directorio de otro",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "normativa", "plazos-y-territorio").
					declarar("plazos-y-territorio", "plazos-y-territorio/SKILL.md/nota.md", "no es de kitlegal").escribir()
			},
			esperados: []instalacion.Hallazgo{hallazgo(aOtroSitio, neutroPlazos+"/SKILL.md", rmY(neutroPlazos+"/SKILL.md", plazos))},
		},
	}
}

// casosDoctorDelHost son los de una entrada de host declarada enlace, en
// local con .claude: el enlace de FR-021 que cuelga, otra cosa en su lugar, la
// entrada que falta y .claude o .claude/skills que ya no son un directorio
// real; y --host claude en las órdenes de las skills con entrada de host.
func casosDoctorDelHost() []casoDeDoctor {
	enlazadas := func(d *discoEnMemoria) {
		instalarLocal(d, "normativa", "plazos-y-territorio").enlazar("normativa").enlazar("plazos-y-territorio").escribir()
	}
	plazos := conHostClaude(instala("plazos-y-territorio"))
	normativa := conHostClaude(instala("normativa"))
	lasDosEntradas := []instalacion.Hallazgo{hallazgo(aOtroSitio, hostNormativa, normativa), hallazgo(aOtroSitio, hostPlazos, plazos)}

	return []casoDeDoctor{
		{nombre: "los enlaces sin tocar", preparar: enlazadas},
		{
			nombre: "la skill es un enlace colgando, y su enlace de host cuelga",
			preparar: func(d *discoEnMemoria) {
				enlazadas(d)
				d.enlace(neutroPlazos, temporal+"/no-existe")
			},
			esperados: []instalacion.Hallazgo{hallazgo(colgando, neutroPlazos, rmY(neutroPlazos, plazos)), hallazgo(colgando, hostPlazos, plazos)},
		},
		{
			nombre: "falta la skill, y su enlace de host cuelga",
			preparar: func(d *discoEnMemoria) {
				enlazadas(d)
				d.retirar(neutroPlazos)
			},
			esperados: []instalacion.Hallazgo{
				hallazgo(editado, neutroPlazos+"/SKILL.md", plazos),
				hallazgo(editado, neutroPlazos+"/references/jerarquia_normativa.md", plazos),
				hallazgo(editado, neutroPlazos+"/references/leyes_vertebrales.md", plazos),
				hallazgo(colgando, hostPlazos, plazos),
			},
		},
		{
			nombre: "un enlace de host con otro destino, que resuelve",
			preparar: func(d *discoEnMemoria) {
				enlazadas(d)
				d.fichero("otro-sitio/mio.md", "no es de kitlegal")
				d.enlace(hostPlazos, "../../otro-sitio")
			},
			esperados:    []instalacion.Hallazgo{hallazgo(aOtroSitio, hostPlazos, rmY(hostPlazos, plazos))},
			noExaminadas: []string{"otro-sitio"},
		},
		{
			nombre: "un enlace de host con otro destino, que cuelga",
			preparar: func(d *discoEnMemoria) {
				enlazadas(d)
				d.enlace(hostPlazos, "../../no-existe")
			},
			esperados: []instalacion.Hallazgo{hallazgo(aOtroSitio, hostPlazos, rmY(hostPlazos, plazos))},
		},
		{
			nombre: "un fichero en lugar del enlace de host",
			preparar: func(d *discoEnMemoria) {
				enlazadas(d)
				d.fichero(hostPlazos, "no es de kitlegal")
			},
			esperados: []instalacion.Hallazgo{hallazgo(aOtroSitio, hostPlazos, rmY(hostPlazos, plazos))},
		},
		{
			nombre: "una tubería en lugar del enlace de host",
			preparar: func(d *discoEnMemoria) {
				enlazadas(d)
				d.tuberia(hostPlazos)
			},
			esperados: []instalacion.Hallazgo{hallazgo(aOtroSitio, hostPlazos, rmY(hostPlazos, plazos))},
		},
		{
			nombre: "un directorio real en lugar del enlace de host",
			preparar: func(d *discoEnMemoria) {
				enlazadas(d)
				d.retirar(hostPlazos)
				d.fichero(hostPlazos+"/nota.md", "no es de kitlegal")
			},
			esperados:    []instalacion.Hallazgo{hallazgo(aOtroSitio, hostPlazos, rmRY(hostPlazos, plazos))},
			noExaminadas: []string{hostPlazos + "/nota.md"},
		},
		{
			nombre: "falta la entrada de host",
			preparar: func(d *discoEnMemoria) {
				enlazadas(d)
				d.retirar(hostPlazos)
			},
			esperados: []instalacion.Hallazgo{hallazgo(aOtroSitio, hostPlazos, plazos)},
		},
		{
			nombre: "falta .claude/skills",
			preparar: func(d *discoEnMemoria) {
				enlazadas(d)
				d.retirar(".claude/skills")
			},
			esperados: lasDosEntradas,
		},
		{
			nombre: "falta todo .claude",
			preparar: func(d *discoEnMemoria) {
				enlazadas(d)
				d.retirar(".claude")
			},
			esperados: lasDosEntradas,
		},
		{
			nombre: ".claude es un fichero",
			preparar: func(d *discoEnMemoria) {
				enlazadas(d)
				d.fichero(".claude", "no soy un directorio")
			},
			esperados: lasDosEntradas,
			excepcion: ".claude",
		},
		{
			nombre: ".claude/skills es un enlace a un directorio",
			preparar: func(d *discoEnMemoria) {
				enlazadas(d)
				d.directorio(temporal + "/skills-de-claude")
				d.enlace(".claude/skills", temporal+"/skills-de-claude")
			},
			esperados:    lasDosEntradas,
			noExaminadas: []string{temporal + "/skills-de-claude"},
			excepcion:    ".claude/skills",
		},
		{
			nombre: "--host claude solo en la orden de la skill con entrada de host",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "normativa", "plazos-y-territorio").enlazar("normativa").escribir()
				d.fichero(neutroNormativa+"/SKILL.md", "editado a mano")
				d.fichero(neutroPlazos+"/SKILL.md", "editado a mano")
			},
			esperados: []instalacion.Hallazgo{
				hallazgo(editado, neutroNormativa+"/SKILL.md", rmY(neutroNormativa+"/SKILL.md", normativa)),
				hallazgo(editado, neutroPlazos+"/SKILL.md", rmY(neutroPlazos+"/SKILL.md", instala("plazos-y-territorio"))),
			},
		},
		{
			nombre: "una skill que el binario no empotra, de otra versión, editada y sin su entrada de host",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "normativa", "plazos-y-territorio").enlazar("normativa").enlazar("plazos-y-territorio").
					noEmpotrada("otra-skill", versionVieja).escribir()
				d.fichero(".agents/skills/otra-skill/SKILL.md", "editado a mano")
			},
			noExaminadas: []string{".agents/skills/otra-skill", ".claude/skills/otra-skill"},
		},
	}
}

// casosDoctorDeLaCopia son los de una entrada de host declarada copia, en
// local con .claude: la copia que ya puede ser un enlace, según diga el mismo
// creador de enlaces que usa install (FR-069); sus ficheros, como los del
// directorio neutro; y otra cosa en su lugar.
func casosDoctorDeLaCopia() []casoDeDoctor {
	conCopia := func(d *discoEnMemoria) {
		instalarLocal(d, "normativa", "plazos-y-territorio").enlazar("normativa").copiar("plazos-y-territorio").escribir()
	}
	plazos := conHostClaude(instala("plazos-y-territorio"))
	sonda := []string{".claude/skills"}

	return []casoDeDoctor{
		{
			nombre:    "una copia donde ya se puede crear el enlace",
			preparar:  conCopia,
			esperados: []instalacion.Hallazgo{hallazgo(esCopia, hostPlazos, plazos)},
			sondas:    sonda,
		},
		{
			nombre: "dos copias, con una sola sonda",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "normativa", "plazos-y-territorio").copiar("normativa").copiar("plazos-y-territorio").escribir()
			},
			esperados: []instalacion.Hallazgo{
				hallazgo(esCopia, hostNormativa, conHostClaude(instala("normativa"))),
				hallazgo(esCopia, hostPlazos, plazos),
			},
			sondas: sonda,
		},
		{
			nombre:   "una copia con un creador de enlaces que solo los crea fuera del ámbito",
			preparar: conCopia,
			admite:   admiteFueraDe(directorioDeTrabajo),
			sondas:   sonda,
		},
		{
			nombre: "un fichero de la copia editado, sin enlaces",
			preparar: func(d *discoEnMemoria) {
				conCopia(d)
				d.fichero(hostPlazos+"/SKILL.md", "editado a mano")
			},
			admite:    admiteNunca,
			esperados: []instalacion.Hallazgo{hallazgo(editado, hostPlazos+"/SKILL.md", rmY(hostPlazos+"/SKILL.md", plazos))},
			sondas:    sonda,
		},
		{
			nombre: "un fichero de la copia que falta, sin enlaces",
			preparar: func(d *discoEnMemoria) {
				conCopia(d)
				d.retirar(hostPlazos + "/references/leyes_vertebrales.md")
			},
			admite:    admiteNunca,
			esperados: []instalacion.Hallazgo{hallazgo(editado, hostPlazos+"/references/leyes_vertebrales.md", plazos)},
			sondas:    sonda,
		},
		{
			nombre: "un fichero de la copia editado, con enlaces",
			preparar: func(d *discoEnMemoria) {
				conCopia(d)
				d.fichero(hostPlazos+"/SKILL.md", "editado a mano")
			},
			esperados: []instalacion.Hallazgo{
				hallazgo(editado, hostPlazos+"/SKILL.md", rmY(hostPlazos+"/SKILL.md", plazos)),
				hallazgo(esCopia, hostPlazos, plazos),
			},
			sondas: sonda,
		},
		{
			nombre: "references de la copia es un enlace colgando",
			preparar: func(d *discoEnMemoria) {
				conCopia(d)
				d.enlace(hostPlazos+"/references", "no-existe")
			},
			admite:    admiteNunca,
			esperados: []instalacion.Hallazgo{hallazgo(colgando, hostPlazos+"/references", rmY(hostPlazos+"/references", plazos))},
			sondas:    sonda,
		},
		{
			nombre: "references de la copia es un enlace a un directorio",
			preparar: func(d *discoEnMemoria) {
				conCopia(d)
				d.directorio(temporal + "/references")
				d.enlace(hostPlazos+"/references", temporal+"/references")
			},
			admite:       admiteNunca,
			esperados:    []instalacion.Hallazgo{hallazgo(aOtroSitio, hostPlazos+"/references", rmY(hostPlazos+"/references", plazos))},
			sondas:       sonda,
			noExaminadas: []string{temporal + "/references"},
		},
		{
			nombre: "una entrada no declarada dentro de la copia, sin enlaces",
			preparar: func(d *discoEnMemoria) {
				conCopia(d)
				d.fichero(hostPlazos+"/mio.md", "no es de kitlegal")
			},
			admite:       admiteNunca,
			sondas:       sonda,
			noExaminadas: []string{hostPlazos + "/mio.md"},
		},
		{
			nombre: "una entrada no declarada dentro de la copia, con enlaces",
			preparar: func(d *discoEnMemoria) {
				conCopia(d)
				d.fichero(hostPlazos+"/mio.md", "no es de kitlegal")
			},
			esperados:    []instalacion.Hallazgo{hallazgo(esCopia, hostPlazos, plazos)},
			sondas:       sonda,
			noExaminadas: []string{hostPlazos + "/mio.md"},
			excepcion:    hostPlazos + "/mio.md",
		},
		{
			nombre: "el enlace de FR-021 donde se declaró la copia",
			preparar: func(d *discoEnMemoria) {
				conCopia(d)
				d.retirar(hostPlazos)
				d.enlace(hostPlazos, "../../.agents/skills/plazos-y-territorio")
			},
			esperados: []instalacion.Hallazgo{hallazgo(aOtroSitio, hostPlazos, rmY(hostPlazos, plazos))},
		},
		{
			nombre: "un fichero donde se declaró la copia",
			preparar: func(d *discoEnMemoria) {
				conCopia(d)
				d.fichero(hostPlazos, "no es de kitlegal")
			},
			esperados: []instalacion.Hallazgo{hallazgo(aOtroSitio, hostPlazos, rmY(hostPlazos, plazos))},
		},
		{
			nombre: "falta la copia",
			preparar: func(d *discoEnMemoria) {
				conCopia(d)
				d.retirar(hostPlazos)
			},
			esperados: []instalacion.Hallazgo{hallazgo(aOtroSitio, hostPlazos, plazos)},
		},
	}
}

// casosDoctorDeLasVersiones son los de la versión distinta (FR-065 (5),
// FR-077): la del manifiesto, un único hallazgo cuya orden reinstala todas las
// skills declaradas y empotradas, o, si coincide, la de cada skill; la regla
// de igualdad, también con un binario de desarrollo; y el orden de la lista.
func casosDoctorDeLasVersiones() []casoDeDoctor {
	const manifiesto = ".agents/skills/kitlegal.json"

	lasDos := instala("normativa", "plazos-y-territorio")

	return []casoDeDoctor{
		{
			nombre: "el manifiesto y sus dos skills, de otra versión",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "normativa", "plazos-y-territorio").deVersion(versionVieja).escribir()
			},
			esperados: []instalacion.Hallazgo{hallazgo(otraVersion, manifiesto, lasDos)},
		},
		{
			nombre: "el manifiesto de otra versión y una sola skill con entrada de host",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "normativa", "plazos-y-territorio").enlazar("plazos-y-territorio").deVersion(versionVieja).escribir()
			},
			esperados: []instalacion.Hallazgo{hallazgo(otraVersion, manifiesto, conHostClaude(lasDos))},
		},
		{
			nombre: "una sola skill de otra versión",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "normativa", "plazos-y-territorio").skillDeVersion("normativa", versionVieja).escribir()
			},
			esperados: []instalacion.Hallazgo{hallazgo(otraVersion, neutroNormativa, instala("normativa"))},
		},
		{
			nombre: "las dos skills de otra versión y el manifiesto no",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "normativa", "plazos-y-territorio").skillDeVersion("normativa", versionVieja).
					skillDeVersion("plazos-y-territorio", "v0.2.0").escribir()
			},
			esperados: []instalacion.Hallazgo{
				hallazgo(otraVersion, neutroNormativa, instala("normativa")),
				hallazgo(otraVersion, neutroPlazos, instala("plazos-y-territorio")),
			},
		},
		{
			nombre: "el manifiesto de otra versión, que solo declara una skill que el binario no empotra",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d).noEmpotrada("otra-skill", versionVieja).deVersion(versionVieja).escribir()
			},
			esperados:    []instalacion.Hallazgo{hallazgo(otraVersion, manifiesto, instala())},
			noExaminadas: []string{".agents/skills/otra-skill", ".claude"},
		},
		{
			nombre: "0.1.0 es la misma versión que v0.1.0",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "normativa", "plazos-y-territorio").deVersion("0.1.0").escribir()
			},
		},
		{
			nombre: "0.1.0+abc no es la misma versión que v0.1.0",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "normativa", "plazos-y-territorio").deVersion("0.1.0+abc").escribir()
			},
			esperados: []instalacion.Hallazgo{hallazgo(otraVersion, manifiesto, lasDos)},
		},
		{
			nombre: "con un binario de desarrollo también se comparan",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "normativa", "plazos-y-territorio").escribir()
			},
			version:   "dev",
			esperados: []instalacion.Hallazgo{hallazgo(otraVersion, manifiesto, lasDos)},
		},
		{
			nombre: "un fichero editado y el manifiesto de otra versión: primero el que lleva rm",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "normativa", "plazos-y-territorio").deVersion(versionVieja).escribir()
				d.fichero(neutroPlazos+"/SKILL.md", "editado a mano")
			},
			esperados: []instalacion.Hallazgo{
				hallazgo(editado, neutroPlazos+"/SKILL.md", rmY(neutroPlazos+"/SKILL.md", instala("plazos-y-territorio"))),
				hallazgo(otraVersion, manifiesto, lasDos),
			},
		},
		{
			nombre: "en cada grupo, por ruta y no por clase ni por skill",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "normativa", "plazos-y-territorio").enlazar("normativa").enlazar("plazos-y-territorio").
					skillDeVersion("plazos-y-territorio", versionVieja).escribir()
				d.retirar(hostNormativa)
				d.fichero(neutroNormativa+"/references/normas.md", "editado a mano")
			},
			esperados: []instalacion.Hallazgo{
				hallazgo(editado, neutroNormativa+"/references/normas.md",
					rmY(neutroNormativa+"/references/normas.md", conHostClaude(instala("normativa")))),
				hallazgo(otraVersion, neutroPlazos, conHostClaude(instala("plazos-y-territorio"))),
				hallazgo(aOtroSitio, hostNormativa, conHostClaude(instala("normativa"))),
			},
		},
	}
}

// probarDoctorSinManifiesto exige que, sin manifiesto en el ámbito, doctor dé
// el diagnóstico sin él —su directorio neutro, manifiesto no, la versión nula,
// la del binario y ningún hallazgo— sin examinar nada más que las guardas y el
// manifiesto (FR-067).
func probarDoctorSinManifiesto(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre     string
		preparar   func(d *discoEnMemoria)
		invocacion instalacion.Invocacion
		neutro     string
	}{
		{nombre: "sin .agents", preparar: func(*discoEnMemoria) {}, neutro: ".agents/skills"},
		{nombre: "sin .agents/skills", preparar: func(d *discoEnMemoria) { d.directorio(".agents") }, neutro: ".agents/skills"},
		{
			nombre: "el directorio neutro sin kitlegal.json, con una skill y .claude",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "plazos-y-territorio").copiar("plazos-y-territorio")
			},
			neutro: ".agents/skills",
		},
		{
			nombre:     "-g con un HOME que no existe",
			preparar:   func(*discoEnMemoria) {},
			invocacion: global,
			neutro:     homeDePrueba + "/.agents/skills",
		},
		{
			nombre:     "--dir que no existe",
			preparar:   func(*discoEnMemoria) {},
			invocacion: instalacion.Invocacion{Dir: texto("otro/destino")},
			neutro:     "otro/destino",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			doctor := casoDeDoctor{preparar: caso.preparar, invocacion: caso.invocacion}
			d, ambito := doctor.preparado(t)
			antes := d.instantanea()

			diagnostico, enlazador, err := doctor.diagnosticar(d, ambito)
			require.NoError(t, err)

			assert.Equal(t, instalacion.Diagnostico{
				Directorio:        caso.neutro,
				VersionDelBinario: versionDePrueba,
				Hallazgos:         []instalacion.Hallazgo{},
			}, diagnostico)
			assert.Equal(t, antes, d.instantanea())
			assert.Empty(t, enlazador.preguntados)
			exigirNoExaminadas(t, d, neutroPlazos, ".claude")
		})
	}
}

// probarFallosDelDoctor exige que un fallo de entrada y salida al examinar,
// al obtener una huella o en la sonda del creador de enlaces se devuelva tal
// cual: no es un hallazgo ni un ámbito ilegible, y doctor sale con 1 sin
// nombrar ninguno.
func probarFallosDelDoctor(t *testing.T) {
	t.Parallel()

	enlazadas := func(d *discoEnMemoria) {
		instalarLocal(d, "normativa", "plazos-y-territorio").enlazar("plazos-y-territorio").escribir()
	}
	conCopia := func(d *discoEnMemoria) {
		instalarLocal(d, "normativa", "plazos-y-territorio").copiar("plazos-y-territorio").escribir()
	}

	casos := []struct {
		nombre   string
		preparar func(d *discoEnMemoria)
		op       operacion
		ruta     string
		sonda    error
	}{
		{nombre: "una guarda", preparar: enlazadas, op: opExaminar, ruta: ".agents/skills"},
		{nombre: "el manifiesto", preparar: enlazadas, op: opExaminar, ruta: ".agents/skills/kitlegal.json"},
		{nombre: "el directorio de una skill", preparar: enlazadas, op: opExaminar, ruta: neutroPlazos},
		{nombre: "un directorio intermedio", preparar: enlazadas, op: opExaminar, ruta: neutroPlazos + "/references"},
		{
			nombre: "el de encima de un directorio intermedio",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "plazos-y-territorio").declarar("plazos-y-territorio", "plazos-y-territorio/a/b/c.md", "no es de kitlegal").escribir()
			},
			op:   opExaminar,
			ruta: neutroPlazos + "/a",
		},
		{nombre: "un fichero declarado", preparar: enlazadas, op: opExaminar, ruta: neutroPlazos + "/SKILL.md"},
		{nombre: "la huella de un fichero declarado", preparar: enlazadas, op: opHuella, ruta: neutroPlazos + "/SKILL.md"},
		{nombre: ".claude", preparar: enlazadas, op: opExaminar, ruta: ".claude"},
		{nombre: ".claude/skills", preparar: enlazadas, op: opExaminar, ruta: ".claude/skills"},
		{nombre: "una entrada de host", preparar: enlazadas, op: opExaminar, ruta: hostPlazos},
		{nombre: "la huella de un fichero de una copia", preparar: conCopia, op: opHuella, ruta: hostPlazos + "/SKILL.md"},
		{nombre: "la sonda del creador de enlaces", preparar: conCopia, sonda: errInyectado},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			d := nuevoDiscoEnMemoria(t)
			caso.preparar(d)

			if caso.ruta != "" {
				d.fallar(caso.op, caso.ruta, errInyectado)
			}

			enlazador := nuevoEnlazadorHonesto(d, admiteSiempre)
			enlazador.err = caso.sonda

			diagnostico, err := instalacion.Diagnosticar(d, enlazador, ambitoLocal, empotradasDePrueba(), versionDePrueba)
			require.ErrorIs(t, err, errInyectado)
			assert.Equal(t, instalacion.Diagnostico{}, diagnostico)

			var conClase schema.ConClase
			assert.NotErrorAs(t, err, &conClase,
				"un fallo de entrada y salida no declara clase: es un defecto del entorno, no un conflicto (ADR 0023)")

			var ilegible *instalacion.AmbitoIlegible
			assert.NotErrorAs(t, err, &ilegible, "ni un ámbito ilegible")
		})
	}
}

// probarVersionDelDoctor exige que una versión del binario que el manifiesto
// no admitiría —la que declararía el install de cada orden— se rechace antes
// de examinar nada.
func probarVersionDelDoctor(t *testing.T) {
	t.Parallel()

	for _, version := range []string{"", "v0.1.0\n"} {
		t.Run(strconv.Quote(version), func(t *testing.T) {
			t.Parallel()

			d := nuevoDiscoEnMemoria(t)
			instalarLocal(d, "plazos-y-territorio").escribir()

			_, err := instalacion.Diagnosticar(d, nuevoEnlazadorHonesto(d, admiteSiempre), ambitoLocal,
				empotradasDePrueba(), version)
			require.Error(t, err)
			assert.Empty(t, d.accesos, "no se examina nada")
		})
	}
}

// probarSalidaDeDoctor fija la salida de doctor sin hallazgos en JSON, la de
// contracts/applet-skills.md §4.3, con la biblioteca con la que la escribe el
// kernel: sus claves en español y en su orden, la versión nula sin manifiesto
// y la lista vacía como [], nunca null. Y las etiquetas jsonschema de las que
// --describe saca el esquema: la versión, nula; la clase de un hallazgo, las
// cinco de FR-065 (research.md D15).
func probarSalidaDeDoctor(t *testing.T) {
	t.Parallel()

	version := versionDePrueba
	diagnosticos := map[string]instalacion.Diagnostico{
		`{"directorio":".agents/skills","manifiesto":false,"version":null,"version_del_binario":"v0.1.0","hallazgos":[]}`: {
			Directorio: ".agents/skills", VersionDelBinario: versionDePrueba, Hallazgos: []instalacion.Hallazgo{},
		},
		`{"directorio":".agents/skills","manifiesto":true,"version":"v0.1.0","version_del_binario":"v0.1.0","hallazgos":[]}`: {
			Directorio: ".agents/skills", Manifiesto: true, Version: &version, VersionDelBinario: versionDePrueba,
			Hallazgos: []instalacion.Hallazgo{},
		},
	}

	for esperado, diagnostico := range diagnosticos {
		salida, err := json.Marshal(diagnostico)
		require.NoError(t, err)
		assert.Equal(t, esperado, string(salida), "las claves en el orden del contrato")
	}

	salida, err := json.Marshal(hallazgo(editado, neutroPlazos+"/SKILL.md", instala("plazos-y-territorio")))
	require.NoError(t, err)
	assert.JSONEq(t, `{"clase": "fichero editado", "ruta": ".agents/skills/plazos-y-territorio/SKILL.md",
		"orden": "kitlegal skills install plazos-y-territorio"}`, string(salida))

	campo, hay := reflect.TypeFor[instalacion.Diagnostico]().FieldByName("Version")
	require.True(t, hay)
	assert.Equal(t, "nullable,minLength=1", campo.Tag.Get("jsonschema"), "la versión, nula sin manifiesto (FR-053)")

	clases := []instalacion.ClaseDeHallazgo{editado, colgando, aOtroSitio, esCopia, otraVersion}
	enumerado := make([]string, 0, len(clases))

	for _, clase := range clases {
		enumerado = append(enumerado, "enum="+string(clase))
	}

	campo, hay = reflect.TypeFor[instalacion.Hallazgo]().FieldByName("Clase")
	require.True(t, hay)
	assert.Equal(t, strings.Join(enumerado, ","), campo.Tag.Get("jsonschema"), "las cinco clases de FR-065")
}
