package instalacion_test

import (
	"fmt"
	"path"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// escenarioDeAplicar es un disco, una invocación de install y un creador de
// enlaces sobre los que se planifica y se aplica el plan.
type escenarioDeAplicar struct {
	nombre string
	// preparar deja el disco del escenario, que empieza con el directorio de
	// trabajo vacío y el temporal.
	preparar func(d *discoEnMemoria)
	// invocacion es la de install; se valida con HOME=homeDePrueba.
	invocacion instalacion.Invocacion
	// admite dice en qué directorio responde Disponible que se puede enlazar;
	// sin él, en todos.
	admite func(directorio string) bool
	// enlaza dice en qué directorio se crea de verdad un enlace; sin él, en
	// los mismos que admite.
	enlaza func(directorio string) bool
}

// resultadoDeAplicar es lo que da una invocación de install sobre el disco de
// un escenario: el plan, la salida y el error de Aplicar, y cada operación
// que se pidió al Escritor, en orden.
type resultadoDeAplicar struct {
	plan        instalacion.Plan
	skills      []instalacion.SkillInstalada
	err         error
	operaciones []string
}

// nuevoDisco es el disco del escenario, recién preparado.
func (e escenarioDeAplicar) nuevoDisco(t *testing.T) *discoEnMemoria {
	t.Helper()

	d := nuevoDiscoEnMemoria(t)
	d.directorio(temporal)
	e.preparar(d)

	return d
}

// aplicar es una invocación nueva de install sobre d, con su propio creador de
// enlaces: planifica, exige que no haya ningún conflicto y aplica el plan con
// un Escritor cuya operación número fallarEn falla (0: ninguna).
func (e escenarioDeAplicar) aplicar(t *testing.T, d *discoEnMemoria, fallarEn int) resultadoDeAplicar {
	t.Helper()

	pedido, err := instalacion.ValidarInvocacion(e.invocacion, homeDePrueba, empotradasDePrueba())
	require.NoError(t, err, "la invocación del escenario")

	admite := e.admite
	if admite == nil {
		admite = admiteSiempre
	}

	enlazador := nuevoEnlazadorHonesto(d, admite)
	if e.enlaza != nil {
		enlazador.enlaza = e.enlaza
	}

	plan, err := instalacion.Planificar(d, enlazador, pedido, empotradasDePrueba(), versionDePrueba)
	require.NoError(t, err, "planificar no encuentra ningún conflicto (FR-044)")

	escritor := &escritorEnMemoria{disco: d, enlazador: enlazador, fallarEn: fallarEn}
	skills, err := instalacion.Aplicar(plan, escritor)

	return resultadoDeAplicar{plan: plan, skills: skills, err: err, operaciones: escritor.operaciones}
}

// sinFallos aplica el escenario sobre su disco sin ningún fallo y devuelve
// cada operación que se pidió al Escritor y el estado final del disco.
func (e escenarioDeAplicar) sinFallos(t *testing.T) ([]string, map[string]nodoEnMemoria) {
	t.Helper()

	d := e.nuevoDisco(t)

	r := e.aplicar(t, d, 0)
	require.NoError(t, r.err)

	return r.operaciones, d.instantanea()
}

// escenarioNuevo es una instalación nueva en local con .claude: crea lo que
// falta hasta .claude/skills y hasta el directorio neutro, enlaza, escribe el
// manifiesto y cada fichero.
func escenarioNuevo() escenarioDeAplicar {
	return escenarioDeAplicar{
		nombre:   "instalación nueva",
		preparar: func(d *discoEnMemoria) { d.directorio(".claude") },
	}
}

// escenarioDeCopiaAEnlace pasa por las cuatro fases: retira un fichero que
// otro binario dejó distinto y una copia de host que pasa a enlace, enlaza esa
// skill y otra nueva, escribe el manifiesto y, por último, los ficheros.
func escenarioDeCopiaAEnlace() escenarioDeAplicar {
	return escenarioDeAplicar{
		nombre: "copia que pasa a enlace, con otra skill nueva",
		preparar: func(d *discoEnMemoria) {
			instalarLocal(d, "boe-legislacion").copiar("boe-legislacion").
				deOtroBinario("boe-legislacion", "SKILL.md", "# boe-legislacion de otro binario\n").escribir()
		},
	}
}

// escenariosDeAplicar son los de un creador de enlaces honesto, que enlaza
// donde dice que puede: cada entrada de host queda en el modo previsto.
func escenariosDeAplicar() []escenarioDeAplicar {
	return []escenarioDeAplicar{
		escenarioNuevo(),
		{
			nombre:     "instalación nueva con -g, --host claude y un HOME que no existe",
			preparar:   func(*discoEnMemoria) {},
			invocacion: instalacion.Invocacion{Global: true, Host: texto("claude")},
		},
		{
			nombre: "actualización desde otro binario, con un fichero declarado que falta",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "boe-legislacion", "legal-core").enlazar("boe-legislacion").enlazar("legal-core").
					deOtroBinario("legal-core", "SKILL.md", "# legal-core de otro binario\n").
					deVersion(versionVieja).escribir()
				d.retirar(".agents/skills/boe-legislacion/references/normas.md")
			},
		},
		escenarioDeCopiaAEnlace(),
		{
			nombre: "retirada de lo que se deja de empotrar, en el neutro y en una copia que se mantiene",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "boe-legislacion", "legal-core").copiar("boe-legislacion").copiar("legal-core").
					deOtroBinario("legal-core", "references/antigua.md", "# Antigua\n").
					copiaDeOtroBinario("legal-core", "references/antigua.md", "# Antigua\n").escribir()
			},
			admite: admiteNunca,
		},
	}
}

// escenarioDeEnlaceQueFalla es una instalación nueva con un creador de
// enlaces que dice que puede enlazar y después no puede: cada entrada pasa a
// su recurso de copia al aplicar (FR-024).
func escenarioDeEnlaceQueFalla() escenarioDeAplicar {
	return escenarioDeAplicar{
		nombre:   "enlaces que fallan al aplicar y pasan a copia",
		preparar: func(d *discoEnMemoria) { d.directorio(".claude") },
		enlaza:   admiteNunca,
	}
}

// TestAplicar fija cómo lleva a cabo Aplicar el plan de install a través del
// Escritor (data-model §5; research.md D7): pide exactamente las operaciones
// del plan, fase a fase y en su orden —retirar, enlazar, escribir el
// manifiesto final y escribir los ficheros—; se para en la primera que falla
// con un error que nombra la operación y la ruta, sin declarar clase, que es
// lo no previsto y sale con código 1 (FR-044; contracts/applet-skills.md §5;
// data-model §9); una entrada cuyo enlace no se puede crear pasa a su copia,
// sin parar la orden, y sale en modo copia en la salida y en el manifiesto
// (FR-024); y sin nada que hacer no pide nada.
func TestAplicar(t *testing.T) {
	t.Parallel()

	t.Run("las operaciones del plan, fase a fase", probarFasesAlAplicar)
	t.Run("se para en el primer fallo y lo nombra", probarPrimerFalloAlAplicar)
	t.Run("un enlace que no se puede crear pasa a su copia", probarCopiaAlAplicar)
	t.Run("sin nada que hacer, no pide nada", probarNadaQueAplicar)
}

// probarFasesAlAplicar exige, en cada escenario, que Aplicar pida al Escritor
// exactamente las operaciones del plan en su orden y que, con cada enlace
// creado, la salida sea la del plan, la de --dry-run (FR-048).
func probarFasesAlAplicar(t *testing.T) {
	t.Parallel()

	for _, escenario := range escenariosDeAplicar() {
		t.Run(escenario.nombre, func(t *testing.T) {
			t.Parallel()

			r := escenario.aplicar(t, escenario.nuevoDisco(t), 0)
			require.NoError(t, r.err)

			assert.Equal(t, porElEscritor(operacionesDelPlan(t, r.plan)), r.operaciones, "las del plan, en su orden (D7)")
			assert.Equal(t, r.plan.Skills, r.skills, "cada entrada de host en el modo previsto")
		})
	}
}

// porElEscritor son las operaciones del plan, como las escribe
// operacionesDelPlan, tal como se piden al Escritor: sin el número de fase, y
// el manifiesto escrito como un fichero más.
func porElEscritor(operaciones []string) []string {
	var pedidas []string

	for _, operacion := range operaciones {
		_, sinFase, _ := strings.Cut(operacion, " ")
		if ruta, esElManifiesto := strings.CutPrefix(sinFase, "manifiesto "); esElManifiesto {
			sinFase = "escribir " + ruta
		}

		pedidas = append(pedidas, sinFase)
	}

	return pedidas
}

// probarPrimerFalloAlAplicar hace fallar una operación de cada tipo y exige
// que Aplicar se pare en ella, sin pedir ninguna más ni dar salida, con el
// mensaje de contracts/applet-skills.md §5, que envuelve el error del
// Escritor y no declara clase (data-model §9).
func probarPrimerFalloAlAplicar(t *testing.T) {
	t.Parallel()

	casos := []struct {
		escenario escenarioDeAplicar
		operacion string
		mensaje   string
	}{
		{
			escenario: escenarioNuevo(),
			operacion: "crear .claude/skills",
			mensaje:   "skills install: crear .claude/skills: fallo de entrada y salida inyectado",
		},
		{
			escenario: escenarioNuevo(),
			operacion: "escribir .agents/skills/kitlegal.json",
			mensaje:   "skills install: escribir .agents/skills/kitlegal.json: fallo de entrada y salida inyectado",
		},
		{
			escenario: escenarioNuevo(),
			operacion: "escribir .agents/skills/legal-core/SKILL.md",
			mensaje:   "skills install: escribir .agents/skills/legal-core/SKILL.md: fallo de entrada y salida inyectado",
		},
		{
			escenario: escenarioDeCopiaAEnlace(),
			operacion: "retirar .claude/skills/boe-legislacion/references",
			mensaje:   "skills install: retirar .claude/skills/boe-legislacion/references: fallo de entrada y salida inyectado",
		},
	}

	for _, caso := range casos {
		t.Run(caso.operacion, func(t *testing.T) {
			t.Parallel()

			operaciones, _ := caso.escenario.sinFallos(t)
			n := slices.Index(operaciones, caso.operacion) + 1
			require.Positive(t, n, "%q es una operación del plan", caso.operacion)

			r := caso.escenario.aplicar(t, caso.escenario.nuevoDisco(t), n)

			require.EqualError(t, r.err, caso.mensaje)
			require.ErrorIs(t, r.err, errInyectado, "envuelve el error del Escritor")

			var conClase schema.ConClase
			assert.NotErrorAs(t, r.err, &conClase, "no declara clase: lo no previsto, código 1")
			assert.Nil(t, r.skills)
			assert.Equal(t, operaciones[:n], r.operaciones, "hasta la que falla, y ninguna más")
		})
	}
}

// probarCopiaAlAplicar hace fallar el primer enlace de una instalación nueva y
// exige que la orden siga: el otro enlace se crea, el manifiesto final declara
// la entrada en copia con la huella de cada fichero copiado, la copia se
// escribe en la fase 4, detrás de lo del plan, y la salida la nombra en modo
// copia sin cambiar el plan (FR-024).
func probarCopiaAlAplicar(t *testing.T) {
	t.Parallel()

	d := escenarioNuevo().nuevoDisco(t)

	r := escenarioNuevo().aplicar(t, d, 2)
	require.NoError(t, r.err, "no poder crear un enlace no para la orden")

	assert.Equal(t, porElEscritor(slices.Concat(
		enlazarLasDos(ambitoLocal, ".claude/skills"),
		manifiestoNuevo(ambitoLocal, ".agents", ".agents/skills"),
		escribirEnteras(t, ".agents/skills/boe-legislacion", ".agents/skills/legal-core", ".claude/skills/boe-legislacion"),
	)), r.operaciones)

	assert.Equal(t, []instalacion.SkillInstalada{
		salidaEn(ambitoLocal, "boe-legislacion", instalacion.EstadoInstalada, instalacion.ModoCopia),
		salidaEn(ambitoLocal, "legal-core", instalacion.EstadoInstalada, instalacion.ModoEnlace),
	}, r.skills)
	assert.Equal(t, lasDos(ambitoLocal, instalacion.EstadoInstalada, instalacion.ModoEnlace), r.plan.Skills,
		"el plan no cambia")
	exigirHostsEnElDisco(t, d, r.skills)

	for _, fichero := range empotradaDePrueba(t, "boe-legislacion").Ficheros {
		assert.Equal(t, fichero.Contenido, leerDelDisco(t, d, path.Join(".claude/skills/boe-legislacion", fichero.Ruta)))
	}

	manifiesto := leerManifiestoDelDisco(t, d)

	copia := manifiesto.Skills["boe-legislacion"].Claude
	require.NotNil(t, copia)
	assert.Equal(t, instalacion.ModoCopia, copia.Modo)
	assert.Equal(t, huellasEmpotradas(t, "boe-legislacion", ".claude/skills/boe-legislacion/"), copia.Ficheros)

	enlace := manifiesto.Skills["legal-core"].Claude
	require.NotNil(t, enlace)
	assert.Equal(t, instalacion.ModoEnlace, enlace.Modo)
}

// probarNadaQueAplicar aplica dos veces el mismo escenario y exige que la
// segunda, sin nada que hacer, no pida ninguna operación —ni el manifiesto— y
// deje el disco igual, con cada skill «sin cambios» (FR-033, FR-045).
func probarNadaQueAplicar(t *testing.T) {
	t.Parallel()

	d := escenarioNuevo().nuevoDisco(t)
	require.NoError(t, escenarioNuevo().aplicar(t, d, 0).err)

	antes := d.instantanea()

	r := escenarioNuevo().aplicar(t, d, 0)
	require.NoError(t, r.err)

	assert.Empty(t, r.operaciones)
	assert.Equal(t, antes, d.instantanea())
	assert.Equal(t, lasDos(ambitoLocal, instalacion.EstadoSinCambios, instalacion.ModoEnlace), r.skills)
}

// TestFalloAMitadSeCompleta fija FR-044 (research.md D7): en cada escenario,
// para cada operación n del plan, un Escritor que falla en la operación n deja
// el disco en un estado sobre el que volver a planificar no da ningún
// conflicto y cuyo plan, aplicado, llega al mismo estado final, byte a byte,
// que la aplicación sin fallos. Un fallo en un enlace no para la orden, que
// sigue con su copia (FR-024); cualquier otro la para en esa operación,
// nombrándola.
func TestFalloAMitadSeCompleta(t *testing.T) {
	t.Parallel()

	for _, escenario := range append(escenariosDeAplicar(), escenarioDeEnlaceQueFalla()) {
		t.Run(escenario.nombre, func(t *testing.T) {
			t.Parallel()

			probarUnFalloEnCadaOperacion(t, escenario)
		})
	}
}

// probarUnFalloEnCadaOperacion aplica el escenario sin fallos para saber sus
// operaciones y su estado final, y lo repite haciendo fallar cada una.
func probarUnFalloEnCadaOperacion(t *testing.T, escenario escenarioDeAplicar) {
	t.Helper()

	operaciones, final := escenario.sinFallos(t)
	require.NotEmpty(t, operaciones)

	for i, operacion := range operaciones {
		t.Run(fmt.Sprintf("%d %s", i+1, operacion), func(t *testing.T) {
			t.Parallel()

			d := escenario.nuevoDisco(t)
			exigirFalloEn(t, escenario.aplicar(t, d, i+1), i+1, operacion)

			segunda := escenario.aplicar(t, d, 0)
			require.NoError(t, segunda.err)
			assert.Equal(t, final, d.instantanea(), "la nueva ejecución completa la instalación")
		})
	}
}

// exigirFalloEn exige lo que hace Aplicar cuando falla la operación n, que
// tiene que ser la misma que sin fallos: si es un enlace, sigue con su copia
// y no falla; si no, se para en ella, sin salida, nombrándola.
func exigirFalloEn(t *testing.T, r resultadoDeAplicar, n int, operacion string) {
	t.Helper()

	require.GreaterOrEqual(t, len(r.operaciones), n)
	assert.Equal(t, operacion, r.operaciones[n-1], "la misma operación que sin fallos")

	if strings.HasPrefix(operacion, "enlazar ") {
		require.NoError(t, r.err, "un enlace que no se puede crear pasa a su copia (FR-024)")

		return
	}

	require.EqualError(t, r.err, "skills install: "+operacion+": "+errInyectado.Error())
	assert.Nil(t, r.skills)
	assert.Len(t, r.operaciones, n, "se para en el primer fallo")
}
