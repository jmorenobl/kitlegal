package instalacion_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
)

// versionNueva es la del binario de casi todas las filas del aviso: posterior
// a versionDePrueba, la de las instalaciones de las pruebas.
const versionNueva = "v0.2.0"

// Lo que el aviso puede pedir al disco en una invocación: los tres pasos del
// §2 de contracts/aviso.md en el ámbito local y en el global, y la lectura de
// un solo manifiesto (plan.md, Performance Goals).
const (
	maximoDeExamenesDelAviso = 6
	maximoDeLecturasDelAviso = 1
)

// casoDeAviso es una fila de TestAviso: el disco, el HOME y la versión del
// binario de una invocación de otro applet, y la línea que tiene que dar el
// aviso.
type casoDeAviso struct {
	nombre string
	// preparar deja el disco del caso, que empieza con el directorio de
	// trabajo y el temporal vacíos.
	preparar func(d *discoEnMemoria)
	// home es el valor de HOME; sin él, homeDePrueba, salvo con sinHome.
	home string
	// sinHome es HOME sin definir o vacío, que para el aviso es lo mismo.
	sinHome bool
	// version es la del binario; sin ella, versionNueva.
	version string
	// esperada es la línea del aviso; vacía si no tiene que haber ninguna.
	esperada string
	// noExaminadas son rutas de las que no se puede examinar nada, ni ellas ni
	// lo que cuelga de ellas.
	noExaminadas []string
}

// homeDelCaso es el HOME del caso.
func (c casoDeAviso) homeDelCaso() string {
	switch {
	case c.sinHome:
		return ""
	case c.home == "":
		return homeDePrueba
	}

	return c.home
}

// versionDelCaso es la versión del binario del caso.
func (c casoDeAviso) versionDelCaso() string {
	if c.version == "" {
		return versionNueva
	}

	return c.version
}

// avisoLocal es la línea de contracts/aviso.md §4 con la versión instalada que
// difiere y la del binario, y avisoGlobal, la misma con -g detrás, la del
// manifiesto global.
func avisoLocal(instalada, binario string) string {
	return "aviso: las skills instaladas son de kitlegal " + instalada + " y este binario es kitlegal " + binario +
		"; ejecuta: kitlegal skills install"
}

func avisoGlobal(instalada, binario string) string {
	return avisoLocal(instalada, binario) + " -g"
}

// enLocal, enGlobal y enNeutro son las dos skills de las pruebas instaladas
// por un binario de version, sin hosts: en el ámbito local, en el global de
// homeDePrueba y en el directorio neutro que se nombre. Cada fila las cambia
// antes de escribir su manifiesto.
func enLocal(d *discoEnMemoria, version string) *instalada {
	return instalarLocal(d, "boe-legislacion", "legal-core").deVersion(version)
}

func enGlobal(d *discoEnMemoria, version string) *instalada {
	return instalarEn(d, ambitoGlobal(d.t, homeDePrueba), "boe-legislacion", "legal-core").deVersion(version)
}

func enNeutro(d *discoEnMemoria, neutro, version string) *instalada {
	return instalarEn(d, instalacion.NuevoAmbitoDir(neutro), "boe-legislacion", "legal-core").deVersion(version)
}

// globalDeOtraVersion deja en el ámbito global el manifiesto de un binario
// anterior, el que avisaría si se llegara a mirar.
func globalDeOtraVersion(d *discoEnMemoria) {
	enGlobal(d, versionDePrueba).escribir()
}

// TestAviso fija el aviso sin red (data-model §7; contracts/aviso.md §2-§4;
// FR-027, FR-036, FR-070 a FR-073, FR-077; SC-013): qué manifiesto mira, con
// una fila por cada salida de cada paso de la búsqueda, cuándo avisa y la
// línea que da. En cada fila:
//
//   - la línea es exactamente la esperada, una sola, y hay aviso si y solo si
//     hay línea; nunca un error;
//   - nada cambia en el disco;
//   - no se abre nada que no sea un fichero regular, no se lista nada y no se
//     examina nada por debajo de una entrada que no es un directorio real, ni
//     a través de un enlace (FR-028);
//   - no se examina nada de lo que la búsqueda no mira: el global cuando el
//     local decide, y lo que hay al otro lado de un enlace;
//   - como mucho seis exámenes y una lectura.
func TestAviso(t *testing.T) {
	t.Parallel()

	grupos := []struct {
		nombre string
		casos  []casoDeAviso
	}{
		{nombre: "paso 1, .agents", casos: casosDelAvisoEnAgents()},
		{nombre: "paso 2, skills en .agents", casos: casosDelAvisoEnSkills()},
		{nombre: "paso 3, kitlegal.json", casos: casosDelAvisoEnElManifiesto()},
		{nombre: "paso 4, HOME y el global", casos: casosDelAvisoEnElGlobal()},
		{nombre: "cuándo avisa", casos: casosDeLasVersionesDelAviso()},
		{nombre: "la línea", casos: casosDeLaLineaDelAviso()},
	}

	for _, grupo := range grupos {
		t.Run(grupo.nombre, func(t *testing.T) {
			t.Parallel()

			for _, caso := range grupo.casos {
				t.Run(caso.nombre, func(t *testing.T) {
					t.Parallel()

					probarCasoDeAviso(t, caso)
				})
			}
		})
	}

	t.Run("binario de desarrollo", probarAvisoDeUnBinarioDeDesarrollo)
}

// probarCasoDeAviso pide el aviso sobre el disco del caso y lo compara con lo
// esperado.
func probarCasoDeAviso(t *testing.T, caso casoDeAviso) {
	t.Helper()

	d := nuevoDiscoEnMemoria(t)
	d.directorio(temporal)
	caso.preparar(d)
	antes := d.instantanea()

	linea, hay := instalacion.Aviso(d, caso.homeDelCaso(), caso.versionDelCaso(), empotradasDePrueba())

	assert.Equal(t, caso.esperada, linea)
	assert.Equal(t, caso.esperada != "", hay, "hay aviso si y solo si hay línea")
	assert.NotContains(t, linea, "\n", "una sola línea (contracts/aviso.md §4)")
	assert.Equal(t, antes, d.instantanea(), "el aviso no cambia nada en el disco")
	exigirDiscoRespetado(t, d, directorioDeTrabajo)
	exigirDiscoRespetado(t, d, homeDePrueba)
	exigirNoExaminadas(t, d, caso.noExaminadas...)
	assert.LessOrEqual(t, d.llamadas(opExaminar), maximoDeExamenesDelAviso, "tres exámenes por ámbito como mucho")
	assert.LessOrEqual(t, d.llamadas(opLeer), maximoDeLecturasDelAviso, "una lectura, la del manifiesto")
	assert.Zero(t, d.llamadas(opHuella), "ninguna huella")
	assert.Zero(t, d.llamadas(opNombres), "ningún listado")
}

// casosDelAvisoEnAgents son los de ./.agents (paso 1): si no existe, se busca
// el global; si existe y no es un directorio real, o no se puede examinar, no
// hay efecto, nada se lee a través de él y el global no se mira.
func casosDelAvisoEnAgents() []casoDeAviso {
	return []casoDeAviso{
		{
			nombre:   "sin .agents, el global de otra versión",
			preparar: globalDeOtraVersion,
			esperada: avisoGlobal(versionDePrueba, versionNueva),
		},
		{
			nombre: ".agents es un fichero, con un global de otra versión",
			preparar: func(d *discoEnMemoria) {
				globalDeOtraVersion(d)
				d.fichero(".agents", "no es un directorio")
			},
			noExaminadas: []string{homeDePrueba},
		},
		{
			nombre: ".agents es un enlace a un directorio con un manifiesto de otra versión",
			preparar: func(d *discoEnMemoria) {
				globalDeOtraVersion(d)
				enNeutro(d, temporal+"/agentes/skills", versionDePrueba).escribir()
				d.enlace(".agents", temporal+"/agentes")
			},
			noExaminadas: []string{temporal, homeDePrueba},
		},
		{
			nombre: ".agents es un enlace colgando",
			preparar: func(d *discoEnMemoria) {
				globalDeOtraVersion(d)
				d.enlace(".agents", "no-existe")
			},
			noExaminadas: []string{homeDePrueba},
		},
		{
			nombre: ".agents es una tubería",
			preparar: func(d *discoEnMemoria) {
				globalDeOtraVersion(d)
				d.tuberia(".agents")
			},
			noExaminadas: []string{homeDePrueba},
		},
		{
			nombre: ".agents no se puede examinar",
			preparar: func(d *discoEnMemoria) {
				globalDeOtraVersion(d)
				enLocal(d, versionDePrueba).escribir()
				d.fallar(opExaminar, ".agents", errInyectado)
			},
			noExaminadas: []string{homeDePrueba},
		},
	}
}

// casosDelAvisoEnSkills son los de ./.agents/skills dentro de un ./.agents
// real (paso 2), con las mismas salidas que el paso 1.
func casosDelAvisoEnSkills() []casoDeAviso {
	return []casoDeAviso{
		{
			nombre: "sin skills en .agents, el global de otra versión",
			preparar: func(d *discoEnMemoria) {
				globalDeOtraVersion(d)
				d.directorio(".agents")
			},
			esperada: avisoGlobal(versionDePrueba, versionNueva),
		},
		{
			nombre: "skills es un enlace colgando, con un global de otra versión",
			preparar: func(d *discoEnMemoria) {
				globalDeOtraVersion(d)
				d.enlace(".agents/skills", "no-existe")
			},
			noExaminadas: []string{homeDePrueba},
		},
		{
			nombre: "skills es un enlace a un directorio con un manifiesto de otra versión",
			preparar: func(d *discoEnMemoria) {
				globalDeOtraVersion(d)
				enNeutro(d, temporal+"/skills", versionDePrueba).escribir()
				d.enlace(".agents/skills", temporal+"/skills")
			},
			noExaminadas: []string{temporal, homeDePrueba},
		},
		{
			nombre: "skills es un fichero",
			preparar: func(d *discoEnMemoria) {
				globalDeOtraVersion(d)
				d.fichero(".agents/skills", "no es un directorio")
			},
			noExaminadas: []string{homeDePrueba},
		},
		{
			nombre: "skills no se puede examinar",
			preparar: func(d *discoEnMemoria) {
				globalDeOtraVersion(d)
				enLocal(d, versionDePrueba).escribir()
				d.fallar(opExaminar, ".agents/skills", errInyectado)
			},
			noExaminadas: []string{homeDePrueba},
		},
	}
}

// casosDelAvisoEnElManifiesto son los de ./.agents/skills/kitlegal.json con
// ./.agents y ./.agents/skills, cada uno, un directorio real (paso 3): si no
// existe, se busca el global; si existe, es el manifiesto y el global no se
// mira, y si es ilegible —también un enlace o lo que no es un fichero
// regular, que no se abre— no hay efecto.
func casosDelAvisoEnElManifiesto() []casoDeAviso {
	const manifiesto = ".agents/skills/kitlegal.json"

	return []casoDeAviso{
		{
			nombre: "sin kitlegal.json, el global de otra versión",
			preparar: func(d *discoEnMemoria) {
				globalDeOtraVersion(d)
				d.fichero(".agents/skills/legal-core/SKILL.md", "# legal-core\n")
			},
			esperada: avisoGlobal(versionDePrueba, versionNueva),
		},
		{
			nombre: "el local de otra versión: sin -g, aunque el global también lo sea",
			preparar: func(d *discoEnMemoria) {
				enGlobal(d, versionVieja).escribir()
				enLocal(d, versionDePrueba).escribir()
			},
			esperada:     avisoLocal(versionDePrueba, versionNueva),
			noExaminadas: []string{homeDePrueba},
		},
		{
			nombre: "el local al día, con un global de otra versión",
			preparar: func(d *discoEnMemoria) {
				globalDeOtraVersion(d)
				enLocal(d, versionNueva).escribir()
			},
			noExaminadas: []string{homeDePrueba},
		},
		{
			nombre: "el local ilegible",
			preparar: func(d *discoEnMemoria) {
				globalDeOtraVersion(d)
				d.fichero(manifiesto, "{")
			},
			noExaminadas: []string{homeDePrueba},
		},
		{
			nombre: "el local sin versión",
			preparar: func(d *discoEnMemoria) {
				globalDeOtraVersion(d)
				d.fichero(manifiesto, `{"skills":{},"version":""}`)
			},
			noExaminadas: []string{homeDePrueba},
		},
		{
			nombre: "kitlegal.json es un enlace a un manifiesto de otra versión",
			preparar: func(d *discoEnMemoria) {
				globalDeOtraVersion(d)
				enNeutro(d, temporal+"/skills", versionDePrueba).escribir()
				d.directorio(".agents/skills")
				d.enlace(manifiesto, temporal+"/skills/kitlegal.json")
			},
			noExaminadas: []string{temporal, homeDePrueba},
		},
		{
			nombre: "kitlegal.json es un directorio",
			preparar: func(d *discoEnMemoria) {
				globalDeOtraVersion(d)
				d.directorio(manifiesto)
			},
			noExaminadas: []string{homeDePrueba},
		},
		{
			nombre: "kitlegal.json es una tubería",
			preparar: func(d *discoEnMemoria) {
				globalDeOtraVersion(d)
				d.tuberia(manifiesto)
			},
			noExaminadas: []string{homeDePrueba},
		},
		{
			nombre: "kitlegal.json no se puede examinar",
			preparar: func(d *discoEnMemoria) {
				globalDeOtraVersion(d)
				enLocal(d, versionDePrueba).escribir()
				d.fallar(opExaminar, manifiesto, errInyectado)
			},
			noExaminadas: []string{homeDePrueba},
		},
		{
			nombre: "kitlegal.json no se puede leer",
			preparar: func(d *discoEnMemoria) {
				globalDeOtraVersion(d)
				enLocal(d, versionDePrueba).escribir()
				d.fallar(opLeer, manifiesto, errInyectado)
			},
			noExaminadas: []string{homeDePrueba},
		},
	}
}

// casosDelAvisoEnElGlobal son los de HOME y el ámbito global (paso 4): sin
// HOME, con HOME en la raíz del sistema o relativo, no se busca el global en
// ninguna parte; con HOME, los pasos 1 a 3 sobre ~/.agents, y todo lo que no
// es el manifiesto legible no tiene efecto.
func casosDelAvisoEnElGlobal() []casoDeAviso {
	return []casoDeAviso{
		{
			nombre:       "sin HOME y sin nada local: sin efecto, tampoco en /.agents",
			preparar:     func(d *discoEnMemoria) { enNeutro(d, "/.agents/skills", versionDePrueba).escribir() },
			sinHome:      true,
			noExaminadas: []string{"/.agents"},
		},
		{
			nombre:   "sin HOME y un local de otra versión: sin -g",
			preparar: func(d *discoEnMemoria) { enLocal(d, versionDePrueba).escribir() },
			sinHome:  true,
			esperada: avisoLocal(versionDePrueba, versionNueva),
		},
		{
			nombre:       "HOME en la raíz del sistema: sin efecto",
			preparar:     func(d *discoEnMemoria) { enNeutro(d, "/.agents/skills", versionDePrueba).escribir() },
			home:         "/",
			noExaminadas: []string{"/.agents"},
		},
		{
			nombre:       "HOME en la raíz del sistema, sin limpiar: sin efecto",
			preparar:     func(d *discoEnMemoria) { enNeutro(d, "/.agents/skills", versionDePrueba).escribir() },
			home:         "//.",
			noExaminadas: []string{"/.agents"},
		},
		{
			nombre:       "HOME relativo: sin efecto",
			preparar:     func(d *discoEnMemoria) { enNeutro(d, "casa/.agents/skills", versionDePrueba).escribir() },
			home:         "casa",
			noExaminadas: []string{"casa"},
		},
		{
			nombre:   "HOME con barra final, el global de otra versión",
			preparar: globalDeOtraVersion,
			home:     homeDePrueba + "/",
			esperada: avisoGlobal(versionDePrueba, versionNueva),
		},
		{
			nombre:   "el global al día",
			preparar: func(d *discoEnMemoria) { enGlobal(d, versionNueva).escribir() },
		},
		{
			nombre:   "sin HOME en el disco",
			preparar: func(*discoEnMemoria) {},
		},
		{
			nombre:   "sin .agents en HOME",
			preparar: func(d *discoEnMemoria) { d.directorio(homeDePrueba) },
		},
		{
			nombre:   ".agents de HOME es un fichero",
			preparar: func(d *discoEnMemoria) { d.fichero(homeDePrueba+"/.agents", "no es un directorio") },
		},
		{
			nombre: ".agents de HOME es un enlace a un directorio con un manifiesto de otra versión",
			preparar: func(d *discoEnMemoria) {
				enNeutro(d, temporal+"/agentes/skills", versionDePrueba).escribir()
				d.directorio(homeDePrueba)
				d.enlace(homeDePrueba+"/.agents", temporal+"/agentes")
			},
			noExaminadas: []string{temporal},
		},
		{
			nombre:   "sin skills en .agents de HOME",
			preparar: func(d *discoEnMemoria) { d.directorio(homeDePrueba + "/.agents") },
		},
		{
			nombre:   "skills de HOME es un enlace colgando",
			preparar: func(d *discoEnMemoria) { d.enlace(homeDePrueba+"/.agents/skills", "no-existe") },
		},
		{
			nombre: "sin kitlegal.json en HOME",
			preparar: func(d *discoEnMemoria) {
				d.fichero(homeDePrueba+"/.agents/skills/legal-core/SKILL.md", "# legal-core\n")
			},
		},
		{
			nombre:   "el global ilegible",
			preparar: func(d *discoEnMemoria) { d.fichero(homeDePrueba+"/.agents/skills/kitlegal.json", "{") },
		},
		{
			nombre: "el kitlegal.json de HOME es un enlace a un manifiesto de otra versión",
			preparar: func(d *discoEnMemoria) {
				enNeutro(d, temporal+"/skills", versionDePrueba).escribir()
				d.directorio(homeDePrueba + "/.agents/skills")
				d.enlace(homeDePrueba+"/.agents/skills/kitlegal.json", temporal+"/skills/kitlegal.json")
			},
			noExaminadas: []string{temporal},
		},
	}
}

// casosDeLasVersionesDelAviso son los de cuándo avisa (§3): la versión del
// manifiesto o la de alguna skill declarada y empotrada que es distinta de la
// del binario según FR-077, nombrando la del manifiesto si difiere y si no la
// de la primera de esas skills por nombre; la de una skill no empotrada no se
// compara (FR-036).
func casosDeLasVersionesDelAviso() []casoDeAviso {
	return []casoDeAviso{
		{
			nombre:   "todo en la versión del binario",
			preparar: func(d *discoEnMemoria) { enLocal(d, versionNueva).escribir() },
		},
		{
			nombre:   "lo instalado sin la v inicial",
			preparar: func(d *discoEnMemoria) { enLocal(d, "0.2.0").escribir() },
		},
		{
			nombre:   "el binario sin la v inicial",
			preparar: func(d *discoEnMemoria) { enLocal(d, versionNueva).escribir() },
			version:  "0.2.0",
		},
		{
			nombre:   "lo instalado con metadatos de construcción",
			preparar: func(d *discoEnMemoria) { enLocal(d, "v0.2.0+abc").escribir() },
			esperada: avisoLocal("v0.2.0+abc", versionNueva),
		},
		{
			nombre:   "el binario con metadatos de construcción",
			preparar: func(d *discoEnMemoria) { enLocal(d, versionNueva).escribir() },
			version:  "v0.2.0+abc",
			esperada: avisoLocal(versionNueva, "v0.2.0+abc"),
		},
		{
			nombre:   "una pre-release",
			preparar: func(d *discoEnMemoria) { enLocal(d, "0.1.1-SNAPSHOT-abc1234").escribir() },
			version:  "0.1.0",
			esperada: avisoLocal("0.1.1-SNAPSHOT-abc1234", "0.1.0"),
		},
		{
			nombre: "el manifiesto al día y una skill de otra versión",
			preparar: func(d *discoEnMemoria) {
				enLocal(d, versionNueva).skillDeVersion("legal-core", versionDePrueba).escribir()
			},
			esperada: avisoLocal(versionDePrueba, versionNueva),
		},
		{
			nombre: "dos skills de otra versión: la primera por nombre",
			preparar: func(d *discoEnMemoria) {
				enLocal(d, versionNueva).
					skillDeVersion("legal-core", versionVieja).
					skillDeVersion("boe-legislacion", versionDePrueba).
					escribir()
			},
			esperada: avisoLocal(versionDePrueba, versionNueva),
		},
		{
			nombre: "el manifiesto y una skill de otra versión: la del manifiesto",
			preparar: func(d *discoEnMemoria) {
				enLocal(d, versionDePrueba).skillDeVersion("boe-legislacion", versionVieja).escribir()
			},
			esperada: avisoLocal(versionDePrueba, versionNueva),
		},
		{
			nombre: "una skill que solo difiere en la v inicial",
			preparar: func(d *discoEnMemoria) {
				enLocal(d, versionNueva).skillDeVersion("legal-core", "0.2.0").escribir()
			},
		},
		{
			nombre: "una skill no empotrada de otra versión",
			preparar: func(d *discoEnMemoria) {
				enLocal(d, versionNueva).noEmpotrada("antigua", versionVieja).escribir()
			},
		},
		{
			nombre: "una no empotrada antes por nombre y una empotrada de otra versión",
			preparar: func(d *discoEnMemoria) {
				enLocal(d, versionNueva).
					noEmpotrada("antigua", versionVieja).
					skillDeVersion("legal-core", versionDePrueba).
					escribir()
			},
			esperada: avisoLocal(versionDePrueba, versionNueva),
		},
		{
			nombre: "el global al día y una skill de otra versión",
			preparar: func(d *discoEnMemoria) {
				enGlobal(d, versionNueva).skillDeVersion("legal-core", versionDePrueba).escribir()
			},
			esperada: avisoGlobal(versionDePrueba, versionNueva),
		},
	}
}

// casosDeLaLineaDelAviso fijan la línea de §4 con su texto literal: las dos
// versiones tal como están, sin quitar ni poner la v, y -g solo con el
// manifiesto global.
func casosDeLaLineaDelAviso() []casoDeAviso {
	return []casoDeAviso{
		{
			nombre:   "con el manifiesto local",
			preparar: func(d *discoEnMemoria) { enLocal(d, "v0.1.0").escribir() },
			version:  "v0.2.0",
			esperada: "aviso: las skills instaladas son de kitlegal v0.1.0 y este binario es kitlegal v0.2.0; " +
				"ejecuta: kitlegal skills install",
		},
		{
			nombre:   "con el manifiesto global",
			preparar: func(d *discoEnMemoria) { enGlobal(d, "0.1.0").escribir() },
			version:  "0.2.0",
			esperada: "aviso: las skills instaladas son de kitlegal 0.1.0 y este binario es kitlegal 0.2.0; " +
				"ejecuta: kitlegal skills install -g",
		},
	}
}

// probarAvisoDeUnBinarioDeDesarrollo fija FR-073: con una versión sin la forma
// de SemVer 2.0.0 no hay aviso, y la forma se comprueba antes que nada, de
// modo que el disco ni se examina.
func probarAvisoDeUnBinarioDeDesarrollo(t *testing.T) {
	t.Parallel()

	for _, version := range []string{"dev", "", "ade1540", "ade1540-dirty", "v0.2", "V0.2.0", "vv0.2.0", "v0.2.0\n"} {
		t.Run(fmt.Sprintf("%q", version), func(t *testing.T) {
			t.Parallel()

			d := nuevoDiscoEnMemoria(t)
			enLocal(d, versionDePrueba).escribir()

			linea, hay := instalacion.Aviso(d, homeDePrueba, version, empotradasDePrueba())

			assert.Empty(t, linea)
			assert.False(t, hay)
			assert.Empty(t, d.accesos, "sin la forma de SemVer no se examina el disco")
		})
	}
}
