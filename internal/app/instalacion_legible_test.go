package app

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
)

// Los tests del texto que install, list y doctor cuentan a una persona
// (docs/ADR/0026): se comparan byte a byte, porque la salida es determinista y
// esa es la garantía. Los valores de entrada son los tipos de la salida, no el
// disco: lo que aquí se prueba es cómo se cuenta, no qué se encuentra, que ya
// lo prueban TestAppletSkills y los guiones h19-skills-*.

// homeDePrueba es el HOME con el que se abrevian las rutas de estos tests.
const homeDePrueba = "/home/ana"

// versionDelManifiesto es la que el manifiesto de estos tests declara, como
// puntero porque así la lleva el listado y el diagnóstico.
func versionDelManifiesto() *string {
	v := "v0.3.0"

	return &v
}

func ambitoGlobalDePrueba(t *testing.T) instalacion.Ambito {
	t.Helper()

	ambito, err := instalacion.NuevoAmbitoGlobal(homeDePrueba)
	require.NoError(t, err)

	return ambito
}

// enlace es una entrada de host de los tests, con las rutas de un ámbito.
func enlace(host, ruta string, modo instalacion.Modo) instalacion.Enlace {
	return instalacion.Enlace{Host: host, Ruta: ruta, Modo: modo}
}

// TestNombresDeMarcaDeLosHosts exige que cada host conocido tenga el nombre
// de su marca: sin él, saldría por su clave, que es lo que la salida para
// una persona no debe decir.
func TestNombresDeMarcaDeLosHosts(t *testing.T) {
	t.Parallel()

	for _, host := range instalacion.HostsConocidos() {
		nombre, hay := nombresDeMarca[host]
		assert.True(t, hay, "el host %q tiene nombre de marca", host)
		assert.NotEqual(t, host, nombre, "el nombre de marca de %q no es su clave", host)
	}

	assert.Equal(t, "otro", nombreDeMarca("otro"), "un host sin nombre de marca sale por su clave")
}

// TestAbreviarHome fija la regla de ~: solo lo que cuelga de HOME, y la propia
// raíz.
func TestAbreviarHome(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre, home, ruta, esperada string
	}{
		{"lo que cuelga de HOME", "/home/ana", "/home/ana/.agents/skills", "~/.agents/skills"},
		{"la propia raíz", "/home/ana", "/home/ana", "~"},
		{"HOME con barra final", "/home/ana/", "/home/ana/.claude/skills/x", "~/.claude/skills/x"},
		{"una ruta relativa", "/home/ana", ".agents/skills", ".agents/skills"},
		{"otra rama del disco", "/home/ana", "/home/anabel/.agents/skills", "/home/anabel/.agents/skills"},
		{"sin HOME", "", "/home/ana/.agents/skills", "/home/ana/.agents/skills"},
		{"HOME es la raíz del sistema", "/", "/srv/skills", "/srv/skills"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.esperada, hogar(caso.home).abreviar(caso.ruta))
		})
	}
}

// TestLegibleDeInstall fija el texto de install: la cabecera con la versión, el
// directorio abreviado y quién lo lee; cada skill con su estado y sus entradas
// con el nombre de la marca; el aviso del host sin entrada; y la despedida,
// que invita a probar el agente o dice que no había nada que cambiar.
func TestLegibleDeInstall(t *testing.T) {
	t.Parallel()

	t.Run("en la cuenta, con los dos hosts", func(t *testing.T) {
		t.Parallel()

		skills := []instalacion.SkillInstalada{
			{
				Nombre: "boe-legislacion", Ruta: "/home/ana/.agents/skills/boe-legislacion",
				Estado: instalacion.EstadoInstalada,
				Enlaces: []instalacion.Enlace{
					enlace("claude", "/home/ana/.claude/skills/boe-legislacion", instalacion.ModoEnlace),
					enlace("antigravity", "/home/ana/.gemini/config/skills/boe-legislacion", instalacion.ModoEnlace),
				},
			},
			{
				Nombre: "legal-core", Ruta: "/home/ana/.agents/skills/legal-core",
				Estado: instalacion.EstadoActualizada,
				Enlaces: []instalacion.Enlace{
					enlace("claude", "/home/ana/.claude/skills/legal-core", instalacion.ModoEnlace),
					enlace("antigravity", "/home/ana/.gemini/config/skills/legal-core", instalacion.ModoEnlace),
				},
			},
		}

		assert.Equal(t, strings.Join([]string{
			"Skills de kitlegal v0.3.0 en ~/.agents/skills (Codex las lee de ahí):",
			"",
			"  boe-legislacion  instalada",
			"    Claude Code  ~/.claude/skills/boe-legislacion",
			"    Antigravity  ~/.gemini/config/skills/boe-legislacion",
			"  legal-core       actualizada",
			"    Claude Code  ~/.claude/skills/legal-core",
			"    Antigravity  ~/.gemini/config/skills/legal-core",
			"",
			"Abre tu agente y pregúntale, por ejemplo: «¿Qué dice el artículo 21 de la Ley 39/2015?»",
			"",
		}, "\n"), legibleDeInstall(skills, ambitoGlobalDePrueba(t), "v0.3.0", homeDePrueba))
	})

	t.Run("en el proyecto, sin ningún host y sin cambios", func(t *testing.T) {
		t.Parallel()

		skills := []instalacion.SkillInstalada{
			instaladaSinEnlaces(".agents/skills", "boe-legislacion", instalacion.EstadoSinCambios),
			instaladaSinEnlaces(".agents/skills", "legal-core", instalacion.EstadoSinCambios),
		}

		assert.Equal(t, strings.Join([]string{
			"Skills de kitlegal dev en .agents/skills (Codex y Antigravity las leen de ahí):",
			"",
			"  boe-legislacion  sin cambios",
			"  legal-core       sin cambios",
			"",
			"Claude Code no las verá: si lo usas, ejecuta «kitlegal skills install --host claude».",
			"Ya estaban instaladas y al día: no se ha cambiado nada.",
			"",
		}, "\n"), legibleDeInstall(skills, instalacion.NuevoAmbitoLocal(), "dev", homeDePrueba))
	})

	t.Run("una copia dice por qué no es un enlace", func(t *testing.T) {
		t.Parallel()

		skills := []instalacion.SkillInstalada{{
			Nombre: "legal-core", Ruta: ".agents/skills/legal-core", Estado: instalacion.EstadoInstalada,
			Enlaces: []instalacion.Enlace{enlace("claude", ".claude/skills/legal-core", instalacion.ModoCopia)},
		}}

		assert.Equal(t, strings.Join([]string{
			"Skills de kitlegal v0.3.0 en .agents/skills (Codex y Antigravity las leen de ahí):",
			"",
			"  legal-core  instalada",
			"    Claude Code  .claude/skills/legal-core (copia, porque en esa carpeta no se pueden crear enlaces)",
			"",
			"Abre tu agente y pregúntale, por ejemplo: «¿Qué dice el artículo 21 de la Ley 39/2015?»",
			"",
		}, "\n"), legibleDeInstall(skills, instalacion.NuevoAmbitoLocal(), "v0.3.0", homeDePrueba))
	})

	t.Run("en la cuenta, solo Claude Code enlazado: Antigravity no las verá", func(t *testing.T) {
		t.Parallel()

		skills := []instalacion.SkillInstalada{{
			Nombre: "legal-core", Ruta: "/home/ana/.agents/skills/legal-core", Estado: instalacion.EstadoInstalada,
			Enlaces: []instalacion.Enlace{enlace("claude", "/home/ana/.claude/skills/legal-core", instalacion.ModoEnlace)},
		}}

		texto := legibleDeInstall(skills, ambitoGlobalDePrueba(t), "v0.3.0", homeDePrueba)
		assert.Contains(t, texto,
			"\nAntigravity no las verá: si lo usas, ejecuta «kitlegal skills install -g --host antigravity».\n")
		assert.NotContains(t, texto, "Claude Code no las verá")
	})

	t.Run("con --dir no hay hosts ni quien lo lea", func(t *testing.T) {
		t.Parallel()

		skills := []instalacion.SkillInstalada{
			instaladaSinEnlaces("destino", "legal-core", instalacion.EstadoInstalada),
		}

		assert.Equal(t, strings.Join([]string{
			"Skills de kitlegal v0.3.0 en destino:",
			"",
			"  legal-core  instalada",
			"",
			"Abre tu agente y pregúntale, por ejemplo: «¿Qué dice el artículo 21 de la Ley 39/2015?»",
			"",
		}, "\n"), legibleDeInstall(skills, instalacion.NuevoAmbitoDir("destino"), "v0.3.0", homeDePrueba))
	})
}

// TestLegibleDeList fija el texto de list: cada skill con su versión y sus
// entradas, la que este binario ya no lleva, el host sin entrada, y los dos
// ámbitos sin nada: sin manifiesto y con un manifiesto que no declara ninguna.
func TestLegibleDeList(t *testing.T) {
	t.Parallel()

	t.Run("en la cuenta, con una skill que el binario ya no lleva", func(t *testing.T) {
		t.Parallel()

		listado := instalacion.Listado{
			Directorio: "/home/ana/.agents/skills", Manifiesto: true, Version: versionDelManifiesto(),
			Skills: []instalacion.SkillListada{
				{
					Nombre: "antigua", Ruta: "/home/ana/.agents/skills/antigua", Version: "v0.1.0", Empotrada: false,
					Enlaces: []instalacion.Enlace{},
				},
				{
					Nombre: "legal-core", Ruta: "/home/ana/.agents/skills/legal-core", Version: "v0.3.0", Empotrada: true,
					Enlaces: []instalacion.Enlace{
						enlace("claude", "/home/ana/.claude/skills/legal-core", instalacion.ModoEnlace),
						enlace("antigravity", "/home/ana/.gemini/config/skills/legal-core", instalacion.ModoCopia),
					},
				},
			},
		}

		assert.Equal(t, strings.Join([]string{
			"Skills de kitlegal en ~/.agents/skills (Codex las lee de ahí):",
			"",
			"  antigua     v0.1.0 (esta versión de kitlegal ya no la lleva)",
			"  legal-core  v0.3.0",
			"    Claude Code  ~/.claude/skills/legal-core",
			"    Antigravity  ~/.gemini/config/skills/legal-core (copia, porque en esa carpeta no se pueden crear enlaces)",
			"",
		}, "\n"), legibleDeList(listado, ambitoGlobalDePrueba(t), homeDePrueba))
	})

	t.Run("en el proyecto, sin ningún host", func(t *testing.T) {
		t.Parallel()

		listado := instalacion.Listado{
			Directorio: ".agents/skills", Manifiesto: true, Version: versionDelManifiesto(),
			Skills: []instalacion.SkillListada{{
				Nombre: "legal-core", Ruta: ".agents/skills/legal-core", Version: "v0.3.0", Empotrada: true,
				Enlaces: []instalacion.Enlace{},
			}},
		}

		assert.Equal(t, strings.Join([]string{
			"Skills de kitlegal en .agents/skills (Codex y Antigravity las leen de ahí):",
			"",
			"  legal-core  v0.3.0",
			"",
			"Claude Code no las verá: si lo usas, ejecuta «kitlegal skills install --host claude».",
			"",
		}, "\n"), legibleDeList(listado, instalacion.NuevoAmbitoLocal(), homeDePrueba))
	})

	t.Run("sin manifiesto, con -g", func(t *testing.T) {
		t.Parallel()

		listado := instalacion.Listado{
			Directorio: "/home/ana/.agents/skills", Manifiesto: false, Skills: []instalacion.SkillListada{},
		}

		assert.Equal(t, "No hay skills de kitlegal en ~/.agents/skills. Para instalarlas: kitlegal skills install -g\n",
			legibleDeList(listado, ambitoGlobalDePrueba(t), homeDePrueba))
	})

	t.Run("un manifiesto que no declara ninguna, con --dir", func(t *testing.T) {
		t.Parallel()

		listado := instalacion.Listado{
			Directorio: "destino", Manifiesto: true, Version: versionDelManifiesto(), Skills: []instalacion.SkillListada{},
		}

		assert.Equal(t, "kitlegal v0.3.0 dejó preparado destino, pero no hay ninguna skill instalada."+
			" Para instalarlas: kitlegal skills install --dir 'destino'\n",
			legibleDeList(listado, instalacion.NuevoAmbitoDir("destino"), homeDePrueba))
	})
}

// TestLegibleDeDoctor fija el texto de doctor: todo en orden, sin nada
// instalado, y cada clase de hallazgo explicada en una frase con su orden tal
// cual —sin abreviar, que es la que se copia— y la despedida en singular o en
// plural.
func TestLegibleDeDoctor(t *testing.T) {
	t.Parallel()

	t.Run("todo en orden", func(t *testing.T) {
		t.Parallel()

		diagnostico := instalacion.Diagnostico{
			Directorio: "/home/ana/.agents/skills", Manifiesto: true, Version: versionDelManifiesto(),
			VersionDelBinario: "v0.3.0", Hallazgos: []instalacion.Hallazgo{},
		}

		assert.Equal(t, "Todo en orden: las skills de ~/.agents/skills son de kitlegal v0.3.0 y están como las dejó.\n",
			legibleDeDoctor(diagnostico, ambitoGlobalDePrueba(t), homeDePrueba))
	})

	t.Run("sin manifiesto", func(t *testing.T) {
		t.Parallel()

		diagnostico := instalacion.Diagnostico{
			Directorio: ".agents/skills", Manifiesto: false, VersionDelBinario: "v0.3.0",
			Hallazgos: []instalacion.Hallazgo{},
		}

		assert.Equal(t, "No hay skills de kitlegal en .agents/skills. Para instalarlas: kitlegal skills install\n",
			legibleDeDoctor(diagnostico, instalacion.NuevoAmbitoLocal(), homeDePrueba))
	})

	t.Run("un solo hallazgo, con -g: la orden no se abrevia", func(t *testing.T) {
		t.Parallel()

		diagnostico := instalacion.Diagnostico{
			Directorio: "/home/ana/.agents/skills", Manifiesto: true, Version: versionDelManifiesto(),
			VersionDelBinario: "v0.3.0",
			Hallazgos: []instalacion.Hallazgo{{
				Clase: instalacion.HallazgoFicheroEditado,
				Ruta:  "/home/ana/.agents/skills/legal-core/SKILL.md",
				Orden: "rm -- '/home/ana/.agents/skills/legal-core/SKILL.md' && kitlegal skills install legal-core -g",
			}},
		}

		assert.Equal(t, strings.Join([]string{
			"Hay 1 cosa que arreglar en las skills de ~/.agents/skills:",
			"",
			"  1. ~/.agents/skills/legal-core/SKILL.md ya no es el fichero que dejó kitlegal:" +
				" se ha editado, se ha sustituido o falta.",
			"     rm -- '/home/ana/.agents/skills/legal-core/SKILL.md' && kitlegal skills install legal-core -g",
			"",
			"Ejecuta la orden y vuelve a comprobarlo con «kitlegal skills doctor -g».",
			"",
		}, "\n"), legibleDeDoctor(diagnostico, ambitoGlobalDePrueba(t), homeDePrueba))
	})

	t.Run("las cinco clases, en el proyecto", func(t *testing.T) {
		t.Parallel()

		version := "v0.1.0"
		diagnostico := instalacion.Diagnostico{
			Directorio: ".agents/skills", Manifiesto: true, Version: &version, VersionDelBinario: "v0.3.0",
			Hallazgos: []instalacion.Hallazgo{
				{
					Clase: instalacion.HallazgoEnlaceColgando, Ruta: ".agents/skills/legal-core",
					Orden: "rm -- '.agents/skills/legal-core' && kitlegal skills install legal-core --host claude",
				},
				{
					Clase: instalacion.HallazgoEnlaceAOtroSitio, Ruta: ".claude/skills/boe-legislacion",
					Orden: "kitlegal skills install boe-legislacion --host claude",
				},
				{
					Clase: instalacion.HallazgoCopia, Ruta: ".claude/skills/legal-core",
					Orden: "kitlegal skills install legal-core --host claude",
				},
				{
					Clase: instalacion.HallazgoVersionDistinta, Ruta: ".agents/skills/kitlegal.json",
					Orden: "kitlegal skills install boe-legislacion legal-core --host claude",
				},
				{
					Clase: instalacion.HallazgoVersionDistinta, Ruta: ".agents/skills/boe-legislacion",
					Orden: "kitlegal skills install boe-legislacion --host claude",
				},
			},
		}

		assert.Equal(t, strings.Join([]string{
			"Hay 5 cosas que arreglar en las skills de .agents/skills:",
			"",
			"  1. .agents/skills/legal-core es un enlace que apunta a algo que ya no existe.",
			"     rm -- '.agents/skills/legal-core' && kitlegal skills install legal-core --host claude",
			"  2. .claude/skills/boe-legislacion no es el enlace que creó kitlegal: apunta a otro sitio, es otra cosa o falta.",
			"     kitlegal skills install boe-legislacion --host claude",
			"  3. .claude/skills/legal-core es una copia, y en esa carpeta ya se pueden crear enlaces: puede pasar a ser uno.",
			"     kitlegal skills install legal-core --host claude",
			"  4. Las skills son de kitlegal v0.1.0 y este kitlegal es v0.3.0.",
			"     kitlegal skills install boe-legislacion legal-core --host claude",
			"  5. La skill de .agents/skills/boe-legislacion es de otra versión de kitlegal, y este kitlegal es v0.3.0.",
			"     kitlegal skills install boe-legislacion --host claude",
			"",
			"Ejecuta las órdenes en ese orden y vuelve a comprobarlo con «kitlegal skills doctor».",
			"",
		}, "\n"), legibleDeDoctor(diagnostico, instalacion.NuevoAmbitoLocal(), homeDePrueba))
	})
}

// TestLegibleSinSecuenciasDeEscape exige que ningún texto lleve colores ni
// secuencias ANSI ni tabuladores: lo que se ve es lo que hay.
func TestLegibleSinSecuenciasDeEscape(t *testing.T) {
	t.Parallel()

	skills := []instalacion.SkillInstalada{{
		Nombre: "legal-core", Ruta: ".agents/skills/legal-core", Estado: instalacion.EstadoInstalada,
		Enlaces: []instalacion.Enlace{enlace("claude", ".claude/skills/legal-core", instalacion.ModoCopia)},
	}}

	for _, texto := range []string{
		legibleDeInstall(skills, instalacion.NuevoAmbitoLocal(), "v0.3.0", ""),
		legibleDeList(instalacion.Listado{Directorio: ".agents/skills"}, instalacion.NuevoAmbitoLocal(), ""),
		legibleDeDoctor(instalacion.Diagnostico{Directorio: ".agents/skills"}, instalacion.NuevoAmbitoLocal(), ""),
	} {
		assert.NotContains(t, texto, "\x1b")
		assert.NotContains(t, texto, "\t")
		assert.True(t, strings.HasSuffix(texto, "\n"), "termina en salto de línea: %q", texto)
	}
}
