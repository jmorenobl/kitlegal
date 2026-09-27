package instalacion_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// rutasDelAmbito es todo lo que un Ámbito dice de sí mismo, con la skill
// legal-core como ejemplo de las rutas que dependen de una skill, para
// comparar un ámbito entero con una sola aserción.
type rutasDelAmbito struct {
	clase             instalacion.ClaseDeAmbito
	raiz              string
	neutro            string
	guardas           []string
	conHosts          bool
	banderas          string
	manifiesto        string
	skill             string
	directorioDelHost string
	skillsDelHost     string
	host              string
	hosts             []string
	configAntigravity string
	skillsAntigravity string
	hostAntigravity   string
}

// rutasDe lee de ambito todo lo que dice de sí mismo.
func rutasDe(ambito instalacion.Ambito) rutasDelAmbito {
	return rutasDelAmbito{
		clase:             ambito.Clase(),
		raiz:              ambito.Raiz(),
		neutro:            ambito.Neutro(),
		guardas:           ambito.Guardas(),
		conHosts:          ambito.ConHosts(),
		banderas:          ambito.Banderas(),
		manifiesto:        ambito.RutaDelManifiesto(),
		skill:             ambito.RutaDeSkill("legal-core"),
		directorioDelHost: ambito.DirectorioDelHost("claude"),
		skillsDelHost:     ambito.SkillsDelHost("claude"),
		host:              ambito.RutaDeHost("claude", "legal-core"),
		hosts:             ambito.Hosts(),
		configAntigravity: ambito.DirectorioDelHost("antigravity"),
		skillsAntigravity: ambito.SkillsDelHost("antigravity"),
		hostAntigravity:   ambito.RutaDeHost("antigravity", "legal-core"),
	}
}

// ambitoGlobal es el ámbito global de home, que tiene que existir.
func ambitoGlobal(t *testing.T, home string) instalacion.Ambito {
	t.Helper()

	ambito, err := instalacion.NuevoAmbitoGlobal(home)
	require.NoError(t, err, "el ámbito global de HOME=%q", home)

	return ambito
}

// TestAmbito fija los tres ámbitos de data-model §2 y las rutas que presentan
// (contracts/applet-skills.md §3; research.md D11): todas con /, limpias, sin
// barra final y como se alcanzan desde el directorio de trabajo —relativas en
// local, colgando de HOME con -g y de la ruta tal como se pasó con --dir—; las
// guardas, de las que cada una tiene que ser un directorio real o no existir
// (FR-027); los hosts, que solo tienen el local —claude— y el global —claude y
// antigravity— (FR-013, FR-021; ADR 0025);
// y las banderas que repite cada orden de doctor, con la ruta de --dir tal
// como se pasó, entre comillas simples y con cada comilla simple escapada como
// la escapa el shell, en la palabra siguiente o, si empieza por «-», en la
// misma palabra que --dir (FR-066).
//
// El valor cero es el ámbito local, el de por omisión (FR-011), y sin HOME no
// hay ámbito global (FR-012).
func TestAmbito(t *testing.T) {
	t.Parallel()

	local := rutasDelAmbito{
		clase:             instalacion.AmbitoLocal,
		raiz:              "",
		neutro:            ".agents/skills",
		guardas:           []string{".agents", ".agents/skills"},
		conHosts:          true,
		banderas:          "",
		manifiesto:        ".agents/skills/kitlegal.json",
		skill:             ".agents/skills/legal-core",
		directorioDelHost: ".claude",
		skillsDelHost:     ".claude/skills",
		host:              ".claude/skills/legal-core",
		hosts:             []string{"claude"},
	}
	globalDeAna := rutasDelAmbito{
		clase:             instalacion.AmbitoGlobal,
		raiz:              "/home/ana",
		neutro:            "/home/ana/.agents/skills",
		guardas:           []string{"/home/ana/.agents", "/home/ana/.agents/skills"},
		conHosts:          true,
		banderas:          "-g",
		manifiesto:        "/home/ana/.agents/skills/kitlegal.json",
		skill:             "/home/ana/.agents/skills/legal-core",
		directorioDelHost: "/home/ana/.claude",
		skillsDelHost:     "/home/ana/.claude/skills",
		host:              "/home/ana/.claude/skills/legal-core",
		hosts:             []string{"claude", "antigravity"},
		configAntigravity: "/home/ana/.gemini/config",
		skillsAntigravity: "/home/ana/.gemini/config/skills",
		hostAntigravity:   "/home/ana/.gemini/config/skills/legal-core",
	}

	// dir es el ámbito de --dir, sin raíz ni hosts, con su directorio neutro
	// ya limpio y sus banderas.
	dir := func(neutro, banderas, manifiesto, skill string) rutasDelAmbito {
		return rutasDelAmbito{
			clase:      instalacion.AmbitoDir,
			neutro:     neutro,
			guardas:    []string{neutro},
			banderas:   banderas,
			manifiesto: manifiesto,
			skill:      skill,
			hosts:      []string{},
		}
	}

	casos := []struct {
		nombre  string
		ambito  instalacion.Ambito
		esperan rutasDelAmbito
	}{
		// Local: el directorio de trabajo, sin buscar ninguna raíz de proyecto.
		{nombre: "local", ambito: instalacion.NuevoAmbitoLocal(), esperan: local},
		{nombre: "el valor cero es el local", ambito: instalacion.Ambito{}, esperan: local},

		// Global: HOME, ya expandido, limpio.
		{nombre: "global", ambito: ambitoGlobal(t, "/home/ana"), esperan: globalDeAna},
		{nombre: "global con HOME con barra final", ambito: ambitoGlobal(t, "/home/ana/"), esperan: globalDeAna},
		{nombre: "global con HOME que se limpia", ambito: ambitoGlobal(t, "/home//ana/./"), esperan: globalDeAna},
		{
			nombre: "global con HOME en la raíz del sistema",
			ambito: ambitoGlobal(t, "/"),
			esperan: rutasDelAmbito{
				clase:             instalacion.AmbitoGlobal,
				raiz:              "/",
				neutro:            "/.agents/skills",
				guardas:           []string{"/.agents", "/.agents/skills"},
				conHosts:          true,
				banderas:          "-g",
				manifiesto:        "/.agents/skills/kitlegal.json",
				skill:             "/.agents/skills/legal-core",
				directorioDelHost: "/.claude",
				skillsDelHost:     "/.claude/skills",
				host:              "/.claude/skills/legal-core",
				hosts:             []string{"claude", "antigravity"},
				configAntigravity: "/.gemini/config",
				skillsAntigravity: "/.gemini/config/skills",
				hostAntigravity:   "/.gemini/config/skills/legal-core",
			},
		},

		// --dir: la ruta tal como se pasó, limpia en las rutas y literal en
		// las banderas.
		{
			nombre:  "dir relativo",
			ambito:  instalacion.NuevoAmbitoDir("destino"),
			esperan: dir("destino", "--dir 'destino'", "destino/kitlegal.json", "destino/legal-core"),
		},
		{
			nombre: "dir con barra final",
			ambito: instalacion.NuevoAmbitoDir("otro/destino/"),
			esperan: dir("otro/destino", "--dir 'otro/destino/'", "otro/destino/kitlegal.json",
				"otro/destino/legal-core"),
		},
		{
			nombre: "dir absoluto",
			ambito: instalacion.NuevoAmbitoDir("/tmp/skills"),
			esperan: dir("/tmp/skills", "--dir '/tmp/skills'", "/tmp/skills/kitlegal.json",
				"/tmp/skills/legal-core"),
		},
		{
			nombre:  "dir que se limpia",
			ambito:  instalacion.NuevoAmbitoDir("./a/../b"),
			esperan: dir("b", "--dir './a/../b'", "b/kitlegal.json", "b/legal-core"),
		},
		{
			nombre:  "dir que es el directorio de trabajo",
			ambito:  instalacion.NuevoAmbitoDir("."),
			esperan: dir(".", "--dir '.'", "kitlegal.json", "legal-core"),
		},
		{
			nombre:  "dir vacío, que se resuelve contra el directorio de trabajo",
			ambito:  instalacion.NuevoAmbitoDir(""),
			esperan: dir(".", "--dir ''", "kitlegal.json", "legal-core"),
		},
		{
			nombre: "dir con una comilla simple",
			ambito: instalacion.NuevoAmbitoDir("it's/x"),
			esperan: dir("it's/x", `--dir 'it'\''s/x'`, "it's/x/kitlegal.json",
				"it's/x/legal-core"),
		},
		{
			// Separada, la palabra «'-raro'» se leería como otra bandera: la
			// ruta va en la misma palabra que --dir.
			nombre:  "dir que empieza por un guion",
			ambito:  instalacion.NuevoAmbitoDir("-raro"),
			esperan: dir("-raro", "--dir='-raro'", "-raro/kitlegal.json", "-raro/legal-core"),
		},
		{
			nombre:  "dir que es un guion suelto",
			ambito:  instalacion.NuevoAmbitoDir("-"),
			esperan: dir("-", "--dir='-'", "-/kitlegal.json", "-/legal-core"),
		},
		{
			nombre: "dir que empieza por un guion y lleva una comilla simple",
			ambito: instalacion.NuevoAmbitoDir("--it's"),
			esperan: dir("--it's", `--dir='--it'\''s'`, "--it's/kitlegal.json",
				"--it's/legal-core"),
		},
		{
			nombre:  "dir con un guion que no va delante",
			ambito:  instalacion.NuevoAmbitoDir("a/-raro"),
			esperan: dir("a/-raro", "--dir 'a/-raro'", "a/-raro/kitlegal.json", "a/-raro/legal-core"),
		},
		{
			nombre: "dir con blancos y lo que el shell expandiría sin comillas",
			ambito: instalacion.NuevoAmbitoDir("mis skills/$HOME/*"),
			esperan: dir("mis skills/$HOME/*", "--dir 'mis skills/$HOME/*'",
				"mis skills/$HOME/*/kitlegal.json", "mis skills/$HOME/*/legal-core"),
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.esperan, rutasDe(caso.ambito))
		})
	}

	t.Run("las guardas son una lista nueva en cada llamada", func(t *testing.T) {
		t.Parallel()

		ambito := instalacion.NuevoAmbitoLocal()
		guardas := ambito.Guardas()
		guardas[0] = "otra"

		assert.Equal(t, []string{".agents", ".agents/skills"}, ambito.Guardas(),
			"cambiar la lista devuelta no puede cambiar el ámbito")
	})

	t.Run("sin HOME no hay ámbito global", func(t *testing.T) {
		t.Parallel()

		ambito, err := instalacion.NuevoAmbitoGlobal("")
		exigirRechazo(t, err, schema.ClaseConflicto, fraseSinHome)
		assert.Zero(t, ambito, "sin HOME se devuelve el valor cero, nunca un ámbito global a medias")
	})
}
