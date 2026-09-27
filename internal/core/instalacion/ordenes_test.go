package instalacion_test

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
)

// especialesDeShell son los caracteres que sh interpretaría fuera de comillas
// y que ninguna orden de doctor lleva sin ellas.
const especialesDeShell = "\"$`;|<>()*?[#~\t\n\r"

// lectorDeShell parte una línea en órdenes y palabras como las parte sh, en el
// subconjunto que puede escribir doctor (FR-066).
type lectorDeShell struct {
	runas    []rune
	i        int
	ordenes  [][]string
	orden    []string
	palabra  strings.Builder
	empezada bool
}

// ordenesDeShell parte linea en las órdenes que encadena && y cada una en sus
// palabras, como lo hace sh: fuera de comillas, un espacio separa palabras y
// una barra invertida seguida de una comilla simple es esa comilla; entre
// comillas simples, todo es literal hasta la siguiente, y dos comillas simples
// seguidas son una palabra vacía. Cualquier otra cosa que sh interpretaría
// fuera de comillas —comillas dobles, $, `, ;, |, un & que no es de &&,
// redirecciones, comodines, #, ~, tabuladores o saltos de línea—, una comilla
// sin cerrar o una orden vacía es un error: una orden de doctor es una sola
// línea y no lleva nada de eso.
func ordenesDeShell(linea string) ([][]string, error) {
	l := &lectorDeShell{runas: []rune(linea)}

	for l.i < len(l.runas) {
		if err := l.leer(); err != nil {
			return nil, err
		}
	}

	if err := l.cerrarOrden(); err != nil {
		return nil, err
	}

	return l.ordenes, nil
}

// leer lee lo que empieza en la posición actual, fuera de comillas.
func (l *lectorDeShell) leer() error {
	r := l.runas[l.i]
	l.i++

	switch {
	case r == '\'':
		return l.leerEntreComillas()
	case r == '\\':
		return l.leerComillaEscapada()
	case r == '&':
		return l.leerY()
	case r == ' ':
		l.cerrarPalabra()
	case strings.ContainsRune(especialesDeShell, r):
		return fmt.Errorf("%q fuera de comillas: sh lo interpretaría", r)
	default:
		l.palabra.WriteRune(r)
		l.empezada = true
	}

	return nil
}

// leerEntreComillas lee hasta la comilla simple que cierra.
func (l *lectorDeShell) leerEntreComillas() error {
	l.empezada = true

	for l.i < len(l.runas) {
		r := l.runas[l.i]
		l.i++

		if r == '\'' {
			return nil
		}

		l.palabra.WriteRune(r)
	}

	return errors.New("una comilla simple sin cerrar")
}

// leerComillaEscapada lee la comilla simple que escapa una barra invertida.
func (l *lectorDeShell) leerComillaEscapada() error {
	if l.i >= len(l.runas) || l.runas[l.i] != '\'' {
		return errors.New("una barra invertida que no escapa una comilla simple")
	}

	l.i++
	l.palabra.WriteRune('\'')
	l.empezada = true

	return nil
}

// leerY lee el && que separa dos órdenes, suelto entre espacios.
func (l *lectorDeShell) leerY() error {
	suelto := !l.empezada && l.i < len(l.runas) && l.runas[l.i] == '&' &&
		(l.i+1 == len(l.runas) || l.runas[l.i+1] == ' ')
	if !suelto {
		return errors.New("un & que no es un && suelto")
	}

	l.i++

	return l.cerrarOrden()
}

// cerrarPalabra añade a la orden la palabra que se estaba leyendo, si hay
// alguna, aunque sea vacía.
func (l *lectorDeShell) cerrarPalabra() {
	if l.empezada {
		l.orden = append(l.orden, l.palabra.String())
	}

	l.palabra.Reset()
	l.empezada = false
}

// cerrarOrden añade la orden que se estaba leyendo, que no puede estar vacía.
func (l *lectorDeShell) cerrarOrden() error {
	l.cerrarPalabra()

	if len(l.orden) == 0 {
		return errors.New("una orden vacía")
	}

	l.ordenes = append(l.ordenes, l.orden)
	l.orden = nil

	return nil
}

// ordenDeInstall es la invocación de install cuyos argumentos, tras
// «kitlegal skills install», son args: las skills, -g, --dir y --host, estas
// dos con su valor en la palabra siguiente o, tras un =, en la misma. Lee como
// el análisis de la invocación del kernel: una palabra siguiente que empieza
// por «-» es otra bandera y no el valor, así que la orden que la escribiera
// saldría con 2 sin reinstalar nada, y es un fallo del test.
func ordenDeInstall(t *testing.T, args []string) instalacion.Invocacion {
	t.Helper()

	var invocacion instalacion.Invocacion

	for i := 0; i < len(args); i++ {
		bandera, valor, conIgual := strings.Cut(args[i], "=")

		switch {
		case args[i] == "-g":
			invocacion.Global = true
		case bandera == "--dir" || bandera == "--host":
			if !conIgual {
				require.Less(t, i+1, len(args), "%s sin valor", args[i])
				require.False(t, strings.HasPrefix(args[i+1], "-"),
					"%s %q: el análisis de la invocación lee el valor como otra bandera", args[i], args[i+1])

				valor = args[i+1]
				i++
			}

			if bandera == "--dir" {
				invocacion.Dir = &valor
			} else {
				invocacion.Hosts = append(invocacion.Hosts, valor)
			}
		default:
			require.False(t, strings.HasPrefix(args[i], "-"), "una bandera que install no tiene: %s", args[i])

			invocacion.Skills = append(invocacion.Skills, args[i])
		}
	}

	return invocacion
}

// ejecutarOrden ejecuta en d la orden de un hallazgo como la ejecutaría sh -c
// en el directorio de trabajo: cada orden que encadena &&, hasta la primera
// que falla, con rm sobre el disco en memoria y con un install del mismo
// binario, con el mismo HOME y con el mismo creador de enlaces que el doctor
// del caso. Devuelve el error de la que falla.
func (c casoDeDoctor) ejecutarOrden(t *testing.T, d *discoEnMemoria, linea string) error {
	t.Helper()

	ordenes, err := ordenesDeShell(linea)
	require.NoError(t, err, "una sola línea de shell POSIX: %s", linea)

	for _, argv := range ordenes {
		if err := c.ejecutar(t, d, argv); err != nil {
			return err
		}
	}

	return nil
}

// ejecutar ejecuta una sola orden, rm o install; cualquier otra es un error del
// test.
func (c casoDeDoctor) ejecutar(t *testing.T, d *discoEnMemoria, argv []string) error {
	t.Helper()

	switch {
	case len(argv) == 3 && argv[0] == "rm" && argv[1] == "--":
		_, err := d.rm(argv[2], false)

		return err
	case len(argv) == 4 && argv[0] == "rm" && argv[1] == "-r" && argv[2] == "--":
		_, err := d.rm(argv[3], true)

		return err
	case len(argv) >= 3 && slices.Equal(argv[:3], []string{"kitlegal", "skills", "install"}):
		return c.instalar(t, d, ordenDeInstall(t, argv[3:]))
	}

	require.Failf(t, "una orden que doctor no escribe", "%q", argv)

	return nil
}

// instalar hace lo que haría install con la invocación: validarla —que la
// orden nombra solo skills empotradas y banderas válidas (FR-066)—,
// planificar y aplicar el plan.
func (c casoDeDoctor) instalar(t *testing.T, d *discoEnMemoria, invocacion instalacion.Invocacion) error {
	t.Helper()

	pedido, err := instalacion.ValidarInvocacion(invocacion, c.homeDelCaso(), empotradasDePrueba())
	require.NoError(t, err, "la orden nombra solo skills de este binario y banderas válidas (FR-066)")

	enlazador := c.enlazadorEn(d)

	plan, err := instalacion.Planificar(d, enlazador, pedido, empotradasDePrueba(), c.versionDelCaso())
	if err != nil {
		return err
	}

	_, err = instalacion.Aplicar(plan, &escritorEnMemoria{disco: d, enlazador: enlazador})

	return err
}

// raizDelAmbito es el directorio del que no sale ninguna orden de doctor: la
// raíz del ámbito o, con --dir, su directorio neutro.
func raizDelAmbito(ambito instalacion.Ambito) string {
	if ambito.Clase() == instalacion.AmbitoDir {
		return absoluta(ambito.Neutro())
	}

	return absoluta(ambito.Raiz())
}

// fueraDe es todo lo que hay en el disco fuera de raiz, entrada a entrada.
func fueraDe(d *discoEnMemoria, raiz string) map[string]nodoEnMemoria {
	fuera := maps.Clone(d.nodos)
	maps.DeleteFunc(fuera, func(ruta string, _ nodoEnMemoria) bool {
		return ruta == raiz || strings.HasPrefix(ruta, raiz+"/")
	})

	return fuera
}

// exigirOrdenDelHallazgo exige que la orden del hallazgo sea una sola línea de
// shell POSIX: una retirada de la propia ruta del hallazgo, dentro del ámbito,
// si la lleva, y un install que nombra solo skills de este binario en el mismo
// ámbito de la invocación, con sus rutas y sus banderas tal como se pasaron.
func exigirOrdenDelHallazgo(t *testing.T, caso casoDeDoctor, ambito instalacion.Ambito, h instalacion.Hallazgo) {
	t.Helper()

	ordenes, err := ordenesDeShell(h.Orden)
	require.NoError(t, err, "una sola línea de shell POSIX: %s", h.Orden)
	require.LessOrEqual(t, len(ordenes), 2, h.Orden)

	install := ordenes[len(ordenes)-1]
	require.GreaterOrEqual(t, len(install), 3, h.Orden)
	require.Equal(t, []string{"kitlegal", "skills", "install"}, install[:3], h.Orden)

	if len(ordenes) == 2 {
		rm := ordenes[0]
		assert.True(t, slices.Equal(rm[:len(rm)-1], []string{"rm", "--"}) ||
			slices.Equal(rm[:len(rm)-1], []string{"rm", "-r", "--"}), "rm -- o rm -r --, nunca -f: %s", h.Orden)
		assert.Equal(t, h.Ruta, rm[len(rm)-1], "rm retira la ruta del hallazgo: %s", h.Orden)
		assert.True(t, strings.HasPrefix(absoluta(h.Ruta), raizDelAmbito(ambito)+"/"),
			"rm no retira nada fuera del ámbito: %s", h.Orden)
	}

	pedido, err := instalacion.ValidarInvocacion(ordenDeInstall(t, install[3:]), caso.homeDelCaso(), empotradasDePrueba())
	require.NoError(t, err, "la orden nombra solo skills de este binario y banderas válidas (FR-066): %s", h.Orden)
	assert.Equal(t, rutasDe(ambito), rutasDe(pedido.Ambito), "la orden actúa en el mismo ámbito: %s", h.Orden)
}

// casosDeOrdenes son las filas de TestOrdenesDeDoctor: cada forma de la orden
// de FR-066 —con rm, con rm -r o sin rm—, con las banderas de cada ámbito,
// escritas tal como se pasaron, y cada ruta entre comillas simples con la
// comilla simple escapada.
func casosDeOrdenes(t *testing.T) []casoDeDoctor {
	t.Helper()

	deHome := ambitoGlobal(t, homeDePrueba)
	deOtroHome := ambitoGlobal(t, "/home/o'neil")
	editadoEn := func(ambito instalacion.Ambito) func(d *discoEnMemoria) {
		return func(d *discoEnMemoria) {
			instalarEn(d, ambito, "boe-legislacion", "legal-core").escribir()
			d.fichero(ambito.RutaDeSkill("legal-core")+"/SKILL.md", "editado a mano")
		}
	}

	return []casoDeDoctor{
		{
			nombre:     "-g: rutas absolutas y la bandera",
			preparar:   editadoEn(deHome),
			invocacion: global,
			esperados: []instalacion.Hallazgo{hallazgo(editado, "/home/ana/.agents/skills/legal-core/SKILL.md",
				"rm -- '/home/ana/.agents/skills/legal-core/SKILL.md' && kitlegal skills install legal-core -g")},
		},
		{
			nombre:     "-g con una comilla simple en HOME",
			preparar:   editadoEn(deOtroHome),
			invocacion: global,
			home:       "/home/o'neil",
			esperados: []instalacion.Hallazgo{hallazgo(editado, "/home/o'neil/.agents/skills/legal-core/SKILL.md",
				`rm -- '/home/o'\''neil/.agents/skills/legal-core/SKILL.md' && kitlegal skills install legal-core -g`)},
		},
		{
			nombre: "-g: una copia, sin rm y con --host claude",
			preparar: func(d *discoEnMemoria) {
				instalarEn(d, deHome, "boe-legislacion", "legal-core").enlazar("boe-legislacion").
					copiar("legal-core").escribir()
			},
			invocacion: global,
			esperados: []instalacion.Hallazgo{hallazgo(esCopia, "/home/ana/.claude/skills/legal-core",
				"kitlegal skills install legal-core -g --host claude")},
			sondas: []string{"/home/ana/.claude/skills"},
		},
		{
			nombre: "-g: un directorio real en lugar del enlace de host, con rm -r",
			preparar: func(d *discoEnMemoria) {
				instalarEn(d, deHome, "legal-core").enlazar("legal-core").escribir()
				d.retirar("/home/ana/.claude/skills/legal-core")
				d.fichero("/home/ana/.claude/skills/legal-core/nota.md", "no es de kitlegal")
			},
			invocacion: global,
			esperados: []instalacion.Hallazgo{hallazgo(aOtroSitio, "/home/ana/.claude/skills/legal-core",
				"rm -r -- '/home/ana/.claude/skills/legal-core' && kitlegal skills install legal-core -g --host claude")},
		},
		{
			nombre:     "--dir relativo, sin --host",
			preparar:   editadoEn(instalacion.NuevoAmbitoDir("destino")),
			invocacion: instalacion.Invocacion{Dir: texto("destino")},
			esperados: []instalacion.Hallazgo{hallazgo(editado, "destino/legal-core/SKILL.md",
				"rm -- 'destino/legal-core/SKILL.md' && kitlegal skills install legal-core --dir 'destino'")},
		},
		{
			nombre:     "--dir con barra final: la ruta limpia y la bandera tal como se pasó",
			preparar:   editadoEn(instalacion.NuevoAmbitoDir("destino")),
			invocacion: instalacion.Invocacion{Dir: texto("destino/")},
			esperados: []instalacion.Hallazgo{hallazgo(editado, "destino/legal-core/SKILL.md",
				"rm -- 'destino/legal-core/SKILL.md' && kitlegal skills install legal-core --dir 'destino/'")},
		},
		{
			nombre:     "--dir absoluto",
			preparar:   editadoEn(instalacion.NuevoAmbitoDir("/srv/skills")),
			invocacion: instalacion.Invocacion{Dir: texto("/srv/skills")},
			esperados: []instalacion.Hallazgo{hallazgo(editado, "/srv/skills/legal-core/SKILL.md",
				"rm -- '/srv/skills/legal-core/SKILL.md' && kitlegal skills install legal-core --dir '/srv/skills'")},
		},
		{
			nombre:     "--dir con una comilla simple",
			preparar:   editadoEn(instalacion.NuevoAmbitoDir("o'tro")),
			invocacion: instalacion.Invocacion{Dir: texto("o'tro")},
			esperados: []instalacion.Hallazgo{hallazgo(editado, "o'tro/legal-core/SKILL.md",
				`rm -- 'o'\''tro/legal-core/SKILL.md' && kitlegal skills install legal-core --dir 'o'\''tro'`)},
		},
		{
			nombre:     "--dir vacío, el directorio de trabajo",
			preparar:   editadoEn(instalacion.NuevoAmbitoDir("")),
			invocacion: instalacion.Invocacion{Dir: texto("")},
			esperados: []instalacion.Hallazgo{hallazgo(editado, "legal-core/SKILL.md",
				"rm -- 'legal-core/SKILL.md' && kitlegal skills install legal-core --dir ''")},
		},
		{
			// Con la ruta en la palabra siguiente, «--dir '-raro'», el install
			// de la orden saldría con 2 después de que rm retirara el fichero.
			nombre:     "--dir que empieza por un guion: la ruta en la misma palabra que la bandera",
			preparar:   editadoEn(instalacion.NuevoAmbitoDir("-raro")),
			invocacion: instalacion.Invocacion{Dir: texto("-raro")},
			esperados: []instalacion.Hallazgo{hallazgo(editado, "-raro/legal-core/SKILL.md",
				"rm -- '-raro/legal-core/SKILL.md' && kitlegal skills install legal-core --dir='-raro'")},
		},
		{
			nombre: "--dir: el manifiesto de otra versión, con las dos skills",
			preparar: func(d *discoEnMemoria) {
				instalarEn(d, instalacion.NuevoAmbitoDir("destino"), "boe-legislacion", "legal-core").
					deVersion(versionVieja).escribir()
			},
			invocacion: instalacion.Invocacion{Dir: texto("destino")},
			esperados: []instalacion.Hallazgo{hallazgo(otraVersion, "destino/kitlegal.json",
				"kitlegal skills install boe-legislacion legal-core --dir 'destino'")},
		},
		{
			nombre: "una ruta con espacios y un && dentro de las comillas",
			preparar: func(d *discoEnMemoria) {
				instalarEn(d, instalacion.NuevoAmbitoDir("mis skills && más"), "legal-core").escribir()
				d.retirar("mis skills && más/legal-core/references")
			},
			invocacion: instalacion.Invocacion{Dir: texto("mis skills && más")},
			esperados: []instalacion.Hallazgo{
				hallazgo(editado, "mis skills && más/legal-core/references/jerarquia_normativa.md",
					"kitlegal skills install legal-core --dir 'mis skills && más'"),
				hallazgo(editado, "mis skills && más/legal-core/references/leyes_vertebrales.md",
					"kitlegal skills install legal-core --dir 'mis skills && más'"),
			},
		},
	}
}

// TestOrdenesDeDoctor fija la orden que arregla cada hallazgo (FR-066;
// contracts/applet-skills.md §6): una sola línea de shell POSIX, «[rm [-r] --
// '<ruta>' && ]kitlegal skills install <skill>… [-g | --dir '<ruta>'] [--host
// claude]», con rm solo donde el contrato lo pone, -r solo ante un directorio
// real y nunca -f; la ruta de rm, la del hallazgo, como se alcanza desde el
// directorio de trabajo; las banderas del ámbito, con la ruta de --dir tal como
// se pasó; --host claude si alguna de las skills que nombra tiene entrada de
// host, y nunca con --dir; y cada ruta entre comillas simples, con cada
// comilla simple cerrando las comillas, escapada y abriéndolas de nuevo. Cada
// orden de cada fila de TestHallazgos y de estas, leída como la lee sh, retira
// la ruta del hallazgo y actúa en el mismo ámbito.
func TestOrdenesDeDoctor(t *testing.T) {
	t.Parallel()

	t.Run("cada forma", func(t *testing.T) {
		t.Parallel()

		probarCasosDeDoctor(t, casosDeOrdenes(t))
	})

	for _, grupo := range todosLosGruposDeDoctor(t) {
		t.Run("leídas como las lee sh: "+grupo.nombre, func(t *testing.T) {
			t.Parallel()

			for _, caso := range grupo.casos {
				d, ambito := caso.preparado(t)

				diagnostico, _, err := caso.diagnosticar(d, ambito)
				require.NoError(t, err, "%s: con hallazgos o sin ellos, doctor no falla (ADR 0023)", caso.nombre)

				for _, h := range diagnostico.Hallazgos {
					exigirOrdenDelHallazgo(t, caso, ambito, h)
				}
			}
		})
	}

	t.Run("el lector de las pruebas", probarLectorDeShell)
}

// todosLosGruposDeDoctor son las filas de TestHallazgos y las de
// TestOrdenesDeDoctor.
func todosLosGruposDeDoctor(t *testing.T) []grupoDeDoctor {
	t.Helper()

	return append(gruposDeDoctor(), grupoDeDoctor{nombre: "cada forma de la orden", casos: casosDeOrdenes(t)})
}

// probarLectorDeShell fija el lector de órdenes de las pruebas, del que
// dependen TestOrdenesDeDoctor y TestOrdenesDeDoctorArreglan: lee lo que sh
// leería de una orden de doctor y rechaza todo lo que sh interpretaría de otra
// forma.
func probarLectorDeShell(t *testing.T) {
	t.Parallel()

	leidas := map[string][][]string{
		"kitlegal skills install legal-core": {{"kitlegal", "skills", "install", "legal-core"}},
		`rm -- 'a b/c'\''d' && kitlegal skills install x --dir ''`: {
			{"rm", "--", "a b/c'd"}, {"kitlegal", "skills", "install", "x", "--dir", ""},
		},
		"rm -- 'x && y' && z": {{"rm", "--", "x && y"}, {"z"}},
		"a'b'c":               {{"abc"}},
		`kitlegal skills install x --dir='-a'\''b'`: {
			{"kitlegal", "skills", "install", "x", "--dir=-a'b"},
		},
	}

	for linea, esperadas := range leidas {
		ordenes, err := ordenesDeShell(linea)
		require.NoError(t, err, linea)
		assert.Equal(t, esperadas, ordenes, linea)
	}

	for _, linea := range []string{
		`rm -- "x"`, "rm -- $HOME", "a; b", "a | b", "a & b", "a &&b", "a&& b", "a && && b", "a > b", "rm -- *",
		"rm -- ~/x", "a\nb", "a\tb", `rm -- 'x`, `a\b`, "", " && a", "a && ", "# a",
	} {
		_, err := ordenesDeShell(linea)
		assert.Error(t, err, "%q", linea)
	}
}

// TestOrdenesDeDoctorArreglan fija las dos garantías de FR-066 (SC-011): en
// cada fila de TestHallazgos y de TestOrdenesDeDoctor con hallazgos, ejecutar
// sus órdenes con sh -c en el mismo directorio de trabajo, una tras otra y en
// el orden en que doctor las lista —rm y un install del mismo binario, con el
// mismo creador de enlaces, sobre el disco en memoria— deja un doctor
// posterior sin hallazgos: (i) con uno solo, su orden lo hace desaparecer sin
// producir otro, y (ii) con varios, las de todos. Ninguna orden toca nada
// fuera del ámbito: el destino de un enlace retirado queda intacto.
//
// En las filas con una de las excepciones de FR-066 —una entrada que el
// manifiesto no declara o una ruta del ámbito que no es un directorio real—,
// doctor no propone retirarla y el install de la orden la nombra como
// conflicto sin cambiar nada (FR-042).
func TestOrdenesDeDoctorArreglan(t *testing.T) {
	t.Parallel()

	for _, grupo := range todosLosGruposDeDoctor(t) {
		t.Run(grupo.nombre, func(t *testing.T) {
			t.Parallel()

			for _, caso := range grupo.casos {
				if len(caso.esperados) == 0 {
					continue
				}

				t.Run(caso.nombre, func(t *testing.T) {
					t.Parallel()

					probarQueLasOrdenesArreglan(t, caso)
				})
			}
		})
	}
}

// probarQueLasOrdenesArreglan ejecuta en el disco del caso las órdenes de sus
// hallazgos, en su orden, y diagnostica de nuevo.
func probarQueLasOrdenesArreglan(t *testing.T, caso casoDeDoctor) {
	t.Helper()

	d, ambito := caso.preparado(t)

	primero, _, err := caso.diagnosticar(d, ambito)
	require.NoError(t, err)

	hallazgos := primero.Hallazgos
	require.NotEmpty(t, hallazgos, "el caso tiene hallazgos que arreglar")

	if caso.excepcion != "" {
		probarExcepcion(t, d, caso, hallazgos)

		return
	}

	raiz := raizDelAmbito(ambito)
	fuera := fueraDe(d, raiz)

	for _, h := range hallazgos {
		require.NoError(t, caso.ejecutarOrden(t, d, h.Orden), "la orden de %s %s: %s", h.Clase, h.Ruta, h.Orden)
	}

	diagnostico, _, err := caso.diagnosticar(d, ambito)
	require.NoError(t, err, "tras %d órdenes, en su orden, doctor no encuentra nada (FR-066 (i) y (ii))", len(hallazgos))
	assert.Equal(t, []instalacion.Hallazgo{}, diagnostico.Hallazgos)
	assert.Equal(t, fuera, fueraDe(d, raiz), "ninguna orden toca nada fuera del ámbito")
}

// probarExcepcion exige que, con la excepción de FR-066 del caso, ninguna
// orden la retire y que el install de la primera la nombre como conflicto sin
// cambiar nada.
func probarExcepcion(t *testing.T, d *discoEnMemoria, caso casoDeDoctor, hallazgos []instalacion.Hallazgo) {
	t.Helper()

	for _, h := range hallazgos {
		assert.NotContains(t, h.Orden, "'"+caso.excepcion+"'", "doctor no propone retirar lo que no es suyo (FR-066)")
	}

	require.NotContains(t, hallazgos[0].Orden, "rm ", "la primera orden solo reinstala")

	antes := d.instantanea()

	var rechazo *instalacion.ErrorDeConflictos
	require.ErrorAs(t, caso.ejecutarOrden(t, d, hallazgos[0].Orden), &rechazo,
		"el install de la orden la nombra como conflicto (FR-042)")
	assert.True(t, slices.ContainsFunc(rechazo.Lista(), func(c instalacion.Conflicto) bool {
		return c.Ruta == caso.excepcion
	}), "el conflicto nombra %s: %v", caso.excepcion, rechazo.Lista())
	assert.Equal(t, antes, d.instantanea(), "y no cambia nada")
}
