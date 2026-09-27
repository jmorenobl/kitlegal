package instalacion_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Las frases de contracts/applet-skills.md §2, una por fila de la tabla: lo que
// contiene el mensaje de cada rechazo de la invocación.
const (
	fraseGlobalConDir     = "-g y --dir se excluyen"
	fraseHostConDir       = "--host no se combina con --dir"
	fraseHostNoAdmitido   = "el único host admitido es claude"
	fraseSkillDesconocida = "no es ninguna skill de este binario; skills disponibles: boe-legislacion, legal-core"
	fraseSinHome          = "HOME no está definido o está vacío"
)

// frasesDeRechazo son las cinco, en el orden de la tabla.
var frasesDeRechazo = []string{
	fraseGlobalConDir,
	fraseHostConDir,
	fraseHostNoAdmitido,
	fraseSkillDesconocida,
	fraseSinHome,
}

// exigirRechazo exige que err sea el rechazo de una fila de
// contracts/applet-skills.md §2: un error que declara la clase esperada
// —buscada como la busca el kernel, con errors.As a schema.ConClase—, cuyo
// mensaje contiene la frase de esa fila y ninguna de las otras cuatro, de modo
// que decide una sola fila. El código de salida que corresponde a cada clase
// no se comprueba aquí: el dominio no importa internal/cli ni en sus tests.
func exigirRechazo(t *testing.T, err error, clase schema.Clase, frase string) {
	t.Helper()

	require.Error(t, err)

	var conClase schema.ConClase
	require.ErrorAs(t, err, &conClase, "el rechazo no declara su clase")
	assert.Equal(t, clase, conClase.Clase(), "la clase del rechazo %q", err)

	for _, otra := range frasesDeRechazo {
		if otra == frase {
			assert.Contains(t, err.Error(), otra, "el mensaje del rechazo")
		} else {
			assert.NotContains(t, err.Error(), otra, "el mensaje nombra otra fila además de la que decide")
		}
	}
}

// texto es el valor de una bandera de cadena que se pasó, aunque vacío.
func texto(valor string) *string {
	return &valor
}

// skillsDelBinario son las skills empotradas de las pruebas, las dos del
// binario, en una lista nueva en cada llamada. Sus ficheros no importan para
// validar una invocación.
func skillsDelBinario() []instalacion.SkillEmpotrada {
	return []instalacion.SkillEmpotrada{{Nombre: "boe-legislacion"}, {Nombre: "legal-core"}}
}

// TestValidarInvocacion fija la validación de contracts/applet-skills.md §2 y
// su precedencia (FR-052): (1) -g con --dir, en los tres verbos; (2) --host
// con --dir; (3) --host con un valor distinto de claude; (4) un nombre que no
// es de una skill empotrada, con la lista de las disponibles; las cuatro de
// clase «argumentos», código 2. Solo sin ninguna de ellas, (5) -g con HOME sin
// definir o vacío, de clase «inesperado», código 1. La primera que se cumple
// decide, aunque la invocación caiga además en las siguientes, y un error de
// argumentos gana siempre a -g sin HOME (SC-005).
//
// list y doctor no declaran nombres ni --host, así que su invocación es la de
// install sin ellos: la fila 1 vale para los tres verbos con las mismas filas.
//
// Ninguna fila examina el disco: ValidarInvocacion no recibe el puerto Disco,
// y HOME le llega como un valor (FR-010, FR-012, FR-013, FR-020). Sin
// rechazo, devuelve el ámbito y las skills pedidas: sin nombres, todas las
// empotradas; con nombres, esas, en orden de nombre y un repetido una vez.
func TestValidarInvocacion(t *testing.T) {
	t.Parallel()

	t.Run("rechazos", probarRechazosDeLaInvocacion)
	t.Run("invocaciones válidas", probarInvocacionesValidas)

	t.Run("nombra la primera skill desconocida, una vez", func(t *testing.T) {
		t.Parallel()

		invocacion := instalacion.Invocacion{Skills: []string{"otra", "legal-core", "desconocida", "otra"}}
		_, err := instalacion.ValidarInvocacion(invocacion, "", skillsDelBinario())

		exigirRechazo(t, err, schema.ClaseArgumentos, fraseSkillDesconocida)
		assert.Equal(t, 1, strings.Count(err.Error(), `"otra"`), "la desconocida, entre comillas y una vez: %q", err)
		assert.NotContains(t, err.Error(), "desconocida", "solo la primera desconocida")
	})

	t.Run("nombra el host que no se admite", func(t *testing.T) {
		t.Parallel()

		_, err := instalacion.ValidarInvocacion(instalacion.Invocacion{Host: texto("codex")}, "", skillsDelBinario())

		exigirRechazo(t, err, schema.ClaseArgumentos, fraseHostNoAdmitido)
		assert.Contains(t, err.Error(), `"codex"`)
	})

	t.Run("las disponibles en orden de nombre aunque lleguen en otro", func(t *testing.T) {
		t.Parallel()

		empotradas := skillsDelBinario()
		slices.Reverse(empotradas)

		invocacion := instalacion.Invocacion{Skills: []string{"desconocida"}}
		_, err := instalacion.ValidarInvocacion(invocacion, "", empotradas)

		exigirRechazo(t, err, schema.ClaseArgumentos, fraseSkillDesconocida)
	})

	t.Run("no cambia los nombres que recibe", func(t *testing.T) {
		t.Parallel()

		nombres := []string{"legal-core", "boe-legislacion", "legal-core"}
		pedido, err := instalacion.ValidarInvocacion(instalacion.Invocacion{Skills: nombres}, "", skillsDelBinario())

		require.NoError(t, err)
		assert.Equal(t, []string{"boe-legislacion", "legal-core"}, pedido.Skills)
		assert.Equal(t, []string{"legal-core", "boe-legislacion", "legal-core"}, nombres,
			"ordenar y quitar repetidos se hace sobre una copia")
	})
}

// probarRechazosDeLaInvocacion pasa por las cinco filas de la tabla, cada una
// sola y junto a las que la siguen, que no llegan a comprobarse.
func probarRechazosDeLaInvocacion(t *testing.T) {
	t.Parallel()

	const home = "/home/ana"

	casos := []struct {
		nombre     string
		invocacion instalacion.Invocacion
		home       string
		clase      schema.Clase
		frase      string
	}{
		// Fila 1: -g con --dir, en install, list y doctor.
		{
			nombre:     "-g con --dir",
			invocacion: instalacion.Invocacion{Global: true, Dir: texto("destino")},
			home:       home, clase: schema.ClaseArgumentos, frase: fraseGlobalConDir,
		},
		{
			nombre:     "-g con --dir vacío, que también se pasó",
			invocacion: instalacion.Invocacion{Global: true, Dir: texto("")},
			home:       home, clase: schema.ClaseArgumentos, frase: fraseGlobalConDir,
		},
		{
			nombre:     "-g con --dir y --host claude",
			invocacion: instalacion.Invocacion{Global: true, Dir: texto("destino"), Host: texto("claude")},
			home:       home, clase: schema.ClaseArgumentos, frase: fraseGlobalConDir,
		},
		{
			nombre:     "-g con --dir sin HOME",
			invocacion: instalacion.Invocacion{Global: true, Dir: texto("destino")},
			home:       "", clase: schema.ClaseArgumentos, frase: fraseGlobalConDir,
		},
		{
			nombre: "-g con --dir, un host que no es claude y una skill desconocida, sin HOME",
			invocacion: instalacion.Invocacion{
				Global: true, Dir: texto("destino"), Host: texto("codex"), Skills: []string{"desconocida"},
			},
			home: "", clase: schema.ClaseArgumentos, frase: fraseGlobalConDir,
		},

		// Fila 2: --host con --dir, sea cual sea el host.
		{
			nombre:     "--host claude con --dir",
			invocacion: instalacion.Invocacion{Host: texto("claude"), Dir: texto("destino")},
			home:       home, clase: schema.ClaseArgumentos, frase: fraseHostConDir,
		},
		{
			nombre:     "--host vacío con --dir",
			invocacion: instalacion.Invocacion{Host: texto(""), Dir: texto(".agents/skills")},
			home:       home, clase: schema.ClaseArgumentos, frase: fraseHostConDir,
		},
		{
			nombre: "--host que no es claude con --dir y una skill desconocida",
			invocacion: instalacion.Invocacion{
				Host: texto("codex"), Dir: texto("destino"), Skills: []string{"desconocida"},
			},
			home: home, clase: schema.ClaseArgumentos, frase: fraseHostConDir,
		},

		// Fila 3: --host con un valor distinto de claude.
		{
			nombre:     "--host codex",
			invocacion: instalacion.Invocacion{Host: texto("codex")},
			home:       home, clase: schema.ClaseArgumentos, frase: fraseHostNoAdmitido,
		},
		{
			nombre:     "--host antigravity",
			invocacion: instalacion.Invocacion{Host: texto("antigravity")},
			home:       home, clase: schema.ClaseArgumentos, frase: fraseHostNoAdmitido,
		},
		{
			nombre:     "--host vacío",
			invocacion: instalacion.Invocacion{Host: texto("")},
			home:       home, clase: schema.ClaseArgumentos, frase: fraseHostNoAdmitido,
		},
		{
			nombre:     "--host Claude, con mayúscula",
			invocacion: instalacion.Invocacion{Host: texto("Claude")},
			home:       home, clase: schema.ClaseArgumentos, frase: fraseHostNoAdmitido,
		},
		{
			nombre:     "--host que no es claude y una skill desconocida",
			invocacion: instalacion.Invocacion{Host: texto("codex"), Skills: []string{"desconocida"}},
			home:       home, clase: schema.ClaseArgumentos, frase: fraseHostNoAdmitido,
		},
		{
			nombre:     "-g con un host que no es claude, sin HOME",
			invocacion: instalacion.Invocacion{Global: true, Host: texto("codex")},
			home:       "", clase: schema.ClaseArgumentos, frase: fraseHostNoAdmitido,
		},

		// Fila 4: un nombre que no es de una skill empotrada.
		{
			nombre:     "una skill desconocida",
			invocacion: instalacion.Invocacion{Skills: []string{"desconocida"}},
			home:       home, clase: schema.ClaseArgumentos, frase: fraseSkillDesconocida,
		},
		{
			nombre:     "una skill desconocida junto a una empotrada",
			invocacion: instalacion.Invocacion{Skills: []string{"legal-core", "desconocida"}},
			home:       home, clase: schema.ClaseArgumentos, frase: fraseSkillDesconocida,
		},
		{
			nombre:     "un nombre vacío",
			invocacion: instalacion.Invocacion{Skills: []string{""}},
			home:       home, clase: schema.ClaseArgumentos, frase: fraseSkillDesconocida,
		},
		{
			nombre:     "el nombre de una empotrada con mayúsculas",
			invocacion: instalacion.Invocacion{Skills: []string{"Legal-Core"}},
			home:       home, clase: schema.ClaseArgumentos, frase: fraseSkillDesconocida,
		},
		{
			nombre:     "una ruta en lugar de un nombre",
			invocacion: instalacion.Invocacion{Skills: []string{"../legal-core"}},
			home:       home, clase: schema.ClaseArgumentos, frase: fraseSkillDesconocida,
		},
		{
			nombre:     "-g con una skill desconocida, sin HOME",
			invocacion: instalacion.Invocacion{Global: true, Skills: []string{"desconocida"}},
			home:       "", clase: schema.ClaseArgumentos, frase: fraseSkillDesconocida,
		},
		{
			nombre: "-g con --host claude y una skill desconocida, sin HOME",
			invocacion: instalacion.Invocacion{
				Global: true, Host: texto("claude"), Skills: []string{"legal-core", "desconocida"},
			},
			home: "", clase: schema.ClaseArgumentos, frase: fraseSkillDesconocida,
		},

		// Fila 5: -g sin HOME, solo sin ninguna de las anteriores.
		{
			nombre:     "-g sin HOME",
			invocacion: instalacion.Invocacion{Global: true},
			home:       "", clase: schema.ClaseInesperado, frase: fraseSinHome,
		},
		{
			nombre: "-g con --host claude y skills empotradas, una repetida, sin HOME",
			invocacion: instalacion.Invocacion{
				Global: true, Host: texto("claude"), Skills: []string{"legal-core", "legal-core"},
			},
			home: "", clase: schema.ClaseInesperado, frase: fraseSinHome,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			pedido, err := instalacion.ValidarInvocacion(caso.invocacion, caso.home, skillsDelBinario())

			exigirRechazo(t, err, caso.clase, caso.frase)
			assert.Zero(t, pedido, "un rechazo no devuelve ningún pedido")
		})
	}
}

// probarInvocacionesValidas fija el pedido de cada invocación que no cae en
// ninguna fila: su ámbito (FR-011 a FR-013), las skills pedidas (FR-010) y si
// se pidió el host claude (FR-023). El ámbito local y el de --dir no miran
// HOME.
func probarInvocacionesValidas(t *testing.T) {
	t.Parallel()

	todas := []string{"boe-legislacion", "legal-core"}

	casos := []struct {
		nombre     string
		invocacion instalacion.Invocacion
		home       string
		esperado   instalacion.Pedido
	}{
		{
			nombre:   "sin nada: el ámbito local y todas las empotradas",
			home:     "/home/ana",
			esperado: instalacion.Pedido{Ambito: instalacion.NuevoAmbitoLocal(), Skills: todas},
		},
		{
			nombre:   "el ámbito local no mira HOME",
			home:     "",
			esperado: instalacion.Pedido{Ambito: instalacion.NuevoAmbitoLocal(), Skills: todas},
		},
		{
			nombre:     "nombres desordenados y repetidos: en orden de nombre y una vez",
			invocacion: instalacion.Invocacion{Skills: []string{"legal-core", "boe-legislacion", "legal-core"}},
			esperado:   instalacion.Pedido{Ambito: instalacion.NuevoAmbitoLocal(), Skills: todas},
		},
		{
			nombre:     "un nombre repetido cuenta una vez",
			invocacion: instalacion.Invocacion{Skills: []string{"legal-core", "legal-core"}},
			esperado:   instalacion.Pedido{Ambito: instalacion.NuevoAmbitoLocal(), Skills: []string{"legal-core"}},
		},
		{
			nombre:     "--host claude",
			invocacion: instalacion.Invocacion{Host: texto("claude")},
			esperado:   instalacion.Pedido{Ambito: instalacion.NuevoAmbitoLocal(), Skills: todas, HostClaude: true},
		},
		{
			nombre:     "-g con HOME",
			invocacion: instalacion.Invocacion{Global: true},
			home:       "/home/ana",
			esperado:   instalacion.Pedido{Ambito: ambitoGlobal(t, "/home/ana"), Skills: todas},
		},
		{
			nombre:     "-g con --host claude, una skill y HOME con barra final",
			invocacion: instalacion.Invocacion{Global: true, Host: texto("claude"), Skills: []string{"legal-core"}},
			home:       "/home/ana/",
			esperado: instalacion.Pedido{
				Ambito: ambitoGlobal(t, "/home/ana"), Skills: []string{"legal-core"}, HostClaude: true,
			},
		},
		{
			nombre:     "--dir no mira HOME",
			invocacion: instalacion.Invocacion{Dir: texto("destino")},
			home:       "",
			esperado:   instalacion.Pedido{Ambito: instalacion.NuevoAmbitoDir("destino"), Skills: todas},
		},
		{
			nombre:     "--dir vacío, con una skill",
			invocacion: instalacion.Invocacion{Dir: texto(""), Skills: []string{"boe-legislacion"}},
			home:       "/home/ana",
			esperado: instalacion.Pedido{
				Ambito: instalacion.NuevoAmbitoDir(""), Skills: []string{"boe-legislacion"},
			},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			pedido, err := instalacion.ValidarInvocacion(caso.invocacion, caso.home, skillsDelBinario())

			require.NoError(t, err)
			assert.Equal(t, caso.esperado, pedido)
		})
	}
}
