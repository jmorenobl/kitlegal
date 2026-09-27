package instalacion_test

import (
	"encoding/json"
	"path"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// cabeceraDeInstall es la primera línea del mensaje con que install nombra
// cada conflicto (contracts/applet-skills.md §5). Su última palabra va en dos
// literales porque misspell, con su diccionario inglés, marca la palabra
// española entera como una errata de «conflicts».
const cabeceraDeInstall = "skills install: nada se ha creado ni cambiado; conflict" + "os:"

// Las clases de FR-041 con un nombre corto, para que cada fila quepa en una
// línea.
const (
	carpetaAjena      = instalacion.ConflictoCarpetaAjena
	fichero           = instalacion.ConflictoFichero
	enlaceAOtroSitio  = instalacion.ConflictoEnlaceAOtroSitio
	enlaceRoto        = instalacion.ConflictoEnlaceRoto
	ficheroEditado    = instalacion.ConflictoFicheroEditado
	ficheroAjeno      = instalacion.ConflictoFicheroAjeno
	noEsDirectorio    = instalacion.ConflictoRutaQueNoEsDirectorio
	ilegible          = instalacion.ConflictoManifiestoIlegible
	conEntradasDeHost = instalacion.ConflictoManifiestoConEntradasDeHost
)

// conflicto es el conflicto de esa clase en esa ruta.
func conflicto(clase instalacion.ClaseDeConflicto, ruta string) instalacion.Conflicto {
	return instalacion.Conflicto{Clase: clase, Ruta: ruta}
}

// casoDeConflictos es una fila de TestConflictos: un disco, una invocación
// y lo que install tiene que encontrar en él antes de escribir nada.
type casoDeConflictos struct {
	nombre string
	// preparar deja el disco del caso, que empieza con el directorio de
	// trabajo vacío.
	preparar func(d *discoEnMemoria)
	// invocacion es la de install; se valida con HOME=homeDePrueba.
	invocacion instalacion.Invocacion
	// disponible es lo que responde el Enlazador a Disponible.
	disponible bool
	// esperados tiene cada conflicto, en el orden en que se nombra; ninguno si
	// install puede seguir.
	esperados []instalacion.Conflicto
	// sondas tiene cada directorio por el que se pregunta Disponible, en
	// orden.
	sondas []string
	// noExaminadas son rutas de las que no se puede examinar nada, ni ellas
	// ni lo que cuelga de ellas.
	noExaminadas []string
}

// Invocaciones de install que repiten las filas.
var (
	conHost = instalacion.Invocacion{Host: texto("claude")}
	global  = instalacion.Invocacion{Global: true}
)

// TestConflictos fija cómo detecta install cada conflicto de data-model §4
// (FR-040 a FR-043, FR-047; SC-008, SC-009), una fila por caso, en el ámbito
// (§4.1), en el directorio neutro (§4.2), en el host claude (§4.3) y en lo que
// no se toca nunca (§4.4). En cada fila:
//
//   - se nombra exactamente cada conflicto esperado, cada entrada con una sola
//     clase y en orden de ruta, en un error que declara la clase «conflicto»
//     (código 7; ADR 0023) y cuyo mensaje es la cabecera de contracts/applet-skills.md §5 y una
//     línea «<clase>: <ruta>» por conflicto; sin ninguno, ningún error;
//   - no se abre nada que no sea un fichero regular, no se lista nada que no
//     sea un directorio real y no se examina nada por debajo de una entrada del
//     ámbito que no es un directorio real, la única que se nombra (FR-028);
//   - no se enlaza nada, y Disponible solo se pregunta, una vez, donde decide
//     un conflicto: por .claude/skills ante una copia de host declarada.
//
// Un fallo de entrada y salida del disco no es un conflicto: se devuelve tal
// cual y la orden sale con 1 sin nombrar ninguno.
func TestConflictos(t *testing.T) {
	t.Parallel()

	grupos := []struct {
		nombre string
		casos  []casoDeConflictos
	}{
		{nombre: "ámbito", casos: casosDelAmbito()},
		{nombre: "ámbito del host", casos: casosDelAmbitoDelHost()},
		{nombre: "directorio neutro", casos: casosDelNeutro()},
		{nombre: "dentro de una skill", casos: casosDentroDeUnaSkill()},
		{nombre: "host claude", casos: casosDelHost()},
		{nombre: "copia de host", casos: casosDeLaCopia()},
		{nombre: "lo que no se toca", casos: casosQueNoSeTocan()},
	}

	for _, grupo := range grupos {
		t.Run(grupo.nombre, func(t *testing.T) {
			t.Parallel()

			probarCasosDeConflictos(t, grupo.casos)
		})
	}

	t.Run("fallos de entrada y salida", probarFallosAlComprobar)
	t.Run("skill pedida que no está empotrada", probarPedidoNoEmpotrado)
}

// probarCasosDeConflictos comprueba cada caso en su propio disco.
func probarCasosDeConflictos(t *testing.T, casos []casoDeConflictos) {
	t.Helper()

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			d := nuevoDiscoEnMemoria(t)
			caso.preparar(d)

			pedido, err := instalacion.ValidarInvocacion(caso.invocacion, homeDePrueba, empotradasDePrueba())
			require.NoError(t, err, "la invocación del caso")

			enlazador := &enlazadorDePrueba{disponible: caso.disponible}
			err = instalacion.ComprobarConflictos(d, enlazador, pedido, empotradasDePrueba())

			exigirConflictos(t, err, caso.esperados)
			exigirDiscoRespetado(t, d, raizVigilada(pedido.Ambito))
			exigirNoExaminadas(t, d, caso.noExaminadas...)
			assert.Equal(t, caso.sondas, enlazador.preguntados, "dónde se pregunta Disponible")
			assert.Empty(t, enlazador.enlazados, "comprobar no enlaza nada")
		})
	}
}

// raizVigilada es el directorio por debajo del cual ninguna entrada que no es
// un directorio real se puede atravesar: la raíz del ámbito, o el directorio
// de la ruta de --dir. Lo de encima no se comprueba (FR-027).
func raizVigilada(ambito instalacion.Ambito) string {
	if ambito.Clase() == instalacion.AmbitoDir {
		return path.Dir(absoluta(ambito.Neutro()))
	}

	return absoluta(ambito.Raiz())
}

// exigirConflictos exige que err nombre exactamente cada conflicto esperado,
// o ningún error si no se espera ninguno.
func exigirConflictos(t *testing.T, err error, esperados []instalacion.Conflicto) {
	t.Helper()

	if len(esperados) == 0 {
		require.NoError(t, err, "sin ningún conflicto, install sigue")

		return
	}

	var rechazo *instalacion.ErrorDeConflictos
	require.ErrorAs(t, err, &rechazo, "cada conflicto llega en el error tipado")
	assert.Equal(t, esperados, rechazo.Lista(), "un conflicto por entrada, en orden de ruta")

	var conClase schema.ConClase
	require.ErrorAs(t, err, &conClase, "el rechazo declara su clase")
	assert.Equal(t, schema.ClaseConflicto, conClase.Clase())

	lineas := []string{cabeceraDeInstall}
	for _, esperado := range esperados {
		lineas = append(lineas, string(esperado.Clase)+": "+esperado.Ruta)
	}

	assert.Equal(t, strings.Join(lineas, "\n"), err.Error(), "el mensaje de contracts/applet-skills.md §5")
}

// casosDelAmbito son los de las tres primeras filas de data-model §4.1: las
// guardas del ámbito, el manifiesto ilegible y, con --dir, el manifiesto con
// entradas de host. Con uno de ellos no se sabe qué es de quién y no se
// examina ninguna skill.
func casosDelAmbito() []casoDeConflictos {
	return []casoDeConflictos{
		{
			nombre:    ".agents es un fichero",
			preparar:  func(d *discoEnMemoria) { d.fichero(".agents", "no soy un directorio") },
			esperados: []instalacion.Conflicto{conflicto(noEsDirectorio, ".agents")},
		},
		{
			nombre: ".agents es un enlace a un directorio",
			preparar: func(d *discoEnMemoria) {
				instalarEn(d, instalacion.NuevoAmbitoDir("real/skills"), "legal-core").escribir()
				d.enlace(".agents", "real")
			},
			invocacion:   conHost,
			esperados:    []instalacion.Conflicto{conflicto(noEsDirectorio, ".agents")},
			noExaminadas: []string{"real"},
		},
		{
			nombre:    ".agents/skills es un fichero",
			preparar:  func(d *discoEnMemoria) { d.fichero(".agents/skills", "no soy un directorio") },
			esperados: []instalacion.Conflicto{conflicto(noEsDirectorio, ".agents/skills")},
		},
		{
			nombre: ".agents/skills es un enlace a un directorio con un manifiesto válido",
			preparar: func(d *discoEnMemoria) {
				instalarEn(d, instalacion.NuevoAmbitoDir("real"), "legal-core").escribir()
				d.directorio(".agents")
				d.enlace(".agents/skills", "../real")
			},
			esperados:    []instalacion.Conflicto{conflicto(noEsDirectorio, ".agents/skills")},
			noExaminadas: []string{"real"},
		},
		{
			nombre:     "la ruta de --dir es un fichero",
			preparar:   func(d *discoEnMemoria) { d.fichero("destino", "no soy un directorio") },
			invocacion: instalacion.Invocacion{Dir: texto("destino")},
			esperados:  []instalacion.Conflicto{conflicto(noEsDirectorio, "destino")},
		},
		{
			nombre: "la ruta de --dir es un enlace a un directorio",
			preparar: func(d *discoEnMemoria) {
				d.fichero("real/nota.md", "lo que hay al otro lado")
				d.enlace("destino", "real")
			},
			invocacion:   instalacion.Invocacion{Dir: texto("destino")},
			esperados:    []instalacion.Conflicto{conflicto(noEsDirectorio, "destino")},
			noExaminadas: []string{"real"},
		},
		{
			nombre: "con -g, $HOME/.agents/skills es un enlace",
			preparar: func(d *discoEnMemoria) {
				d.directorio(homeDePrueba + "/.agents")
				d.directorio(homeDePrueba + "/real")
				d.enlace(homeDePrueba+"/.agents/skills", "../real")
			},
			invocacion: global,
			esperados:  []instalacion.Conflicto{conflicto(noEsDirectorio, homeDePrueba+"/.agents/skills")},
		},
		{
			nombre: "con -g, un HOME al que se llega por un enlace se usa tal cual",
			preparar: func(d *discoEnMemoria) {
				d.fichero("/usuarios/ana/.agents/skills/boe-legislacion/mio.md", "una carpeta ajena")
				d.enlace("/home", "/usuarios")
			},
			invocacion: global,
			esperados:  []instalacion.Conflicto{conflicto(carpetaAjena, homeDePrueba+"/.agents/skills/boe-legislacion")},
		},
		{
			nombre:    "manifiesto que no es JSON",
			preparar:  func(d *discoEnMemoria) { d.fichero(".agents/skills/kitlegal.json", "esto no es JSON") },
			esperados: []instalacion.Conflicto{conflicto(ilegible, ".agents/skills/kitlegal.json")},
		},
		{
			nombre: "manifiesto que es un enlace a uno válido",
			preparar: func(d *discoEnMemoria) {
				instalarEn(d, instalacion.NuevoAmbitoDir("fuera"), "legal-core").escribir()
				d.enlace(".agents/skills/kitlegal.json", "../../fuera/kitlegal.json")
			},
			esperados:    []instalacion.Conflicto{conflicto(ilegible, ".agents/skills/kitlegal.json")},
			noExaminadas: []string{"fuera"},
		},
		{
			nombre: "manifiesto que es un directorio",
			preparar: func(d *discoEnMemoria) {
				d.fichero(".agents/skills/kitlegal.json/nota.md", "no soy un manifiesto")
			},
			esperados:    []instalacion.Conflicto{conflicto(ilegible, ".agents/skills/kitlegal.json")},
			noExaminadas: []string{".agents/skills/kitlegal.json/nota.md"},
		},
		{
			nombre:    "manifiesto que es una tubería con nombre",
			preparar:  func(d *discoEnMemoria) { d.tuberia(".agents/skills/kitlegal.json") },
			esperados: []instalacion.Conflicto{conflicto(ilegible, ".agents/skills/kitlegal.json")},
		},
		{
			nombre: "manifiesto que no se puede leer",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "legal-core").escribir()
				d.fallar(opLeer, ".agents/skills/kitlegal.json", errInyectado)
			},
			esperados: []instalacion.Conflicto{conflicto(ilegible, ".agents/skills/kitlegal.json")},
		},
		{
			nombre: "con --dir, manifiesto que declara entradas de host",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "boe-legislacion", "legal-core").enlazar("legal-core").escribir()
				d.fichero(".agents/skills/boe-legislacion/SKILL.md", "editado a mano")
			},
			invocacion:   instalacion.Invocacion{Dir: texto(".agents/skills")},
			esperados:    []instalacion.Conflicto{conflicto(conEntradasDeHost, ".agents/skills/kitlegal.json")},
			noExaminadas: []string{".agents/skills/boe-legislacion", ".agents/skills/legal-core", ".claude"},
		},
		{
			nombre: "con --dir, manifiesto sin entradas de host",
			preparar: func(d *discoEnMemoria) {
				instalarEn(d, instalacion.NuevoAmbitoDir("destino"), "boe-legislacion", "legal-core").escribir()
			},
			invocacion: instalacion.Invocacion{Dir: texto("destino")},
		},
		{
			nombre: "con un conflicto de ámbito no se examina ninguna skill",
			preparar: func(d *discoEnMemoria) {
				d.fichero(".agents/skills/kitlegal.json", "esto no es JSON")
				d.fichero(".agents/skills/boe-legislacion/mio.md", "una carpeta ajena")
				d.fichero(".claude/skills", "no soy un directorio")
			},
			esperados: []instalacion.Conflicto{
				conflicto(ilegible, ".agents/skills/kitlegal.json"),
				conflicto(noEsDirectorio, ".claude/skills"),
			},
			noExaminadas: []string{".agents/skills/boe-legislacion", ".agents/skills/legal-core"},
		},
	}
}

// casosDelAmbitoDelHost son los de las dos últimas filas de data-model §4.1:
// .claude y .claude/skills, cada uno de los cuales tiene que ser un directorio
// real o no existir cuando se enlaza en el host (FR-022, FR-023, FR-026).
func casosDelAmbitoDelHost() []casoDeConflictos {
	return []casoDeConflictos{
		{
			nombre:     ".claude es un fichero, con --host claude",
			preparar:   func(d *discoEnMemoria) { d.fichero(".claude", "no soy un directorio") },
			invocacion: conHost,
			esperados:  []instalacion.Conflicto{conflicto(noEsDirectorio, ".claude")},
		},
		{
			nombre: ".claude es un enlace a un directorio, con --host claude",
			preparar: func(d *discoEnMemoria) {
				d.fichero("real-claude/skills/legal-core", "un fichero")
				d.enlace(".claude", "real-claude")
			},
			invocacion:   conHost,
			esperados:    []instalacion.Conflicto{conflicto(noEsDirectorio, ".claude")},
			noExaminadas: []string{"real-claude"},
		},
		{
			nombre:   ".claude es un fichero, sin --host: cuenta como ausente",
			preparar: func(d *discoEnMemoria) { d.fichero(".claude", "no soy un directorio") },
		},
		{
			nombre: ".claude es un enlace, sin --host: cuenta como ausente",
			preparar: func(d *discoEnMemoria) {
				d.fichero("real-claude/skills/legal-core", "un fichero")
				d.enlace(".claude", "real-claude")
			},
			noExaminadas: []string{"real-claude"},
		},
		{
			nombre:    ".claude/skills es un fichero, con .claude real",
			preparar:  func(d *discoEnMemoria) { d.fichero(".claude/skills", "no soy un directorio") },
			esperados: []instalacion.Conflicto{conflicto(noEsDirectorio, ".claude/skills")},
		},
		{
			nombre: ".claude/skills es un enlace a un directorio, con --host claude",
			preparar: func(d *discoEnMemoria) {
				d.directorio(".claude")
				d.fichero("real-skills/legal-core", "un fichero")
				d.enlace(".claude/skills", "../real-skills")
			},
			invocacion:   conHost,
			esperados:    []instalacion.Conflicto{conflicto(noEsDirectorio, ".claude/skills")},
			noExaminadas: []string{"real-skills"},
		},
		{
			nombre: "con .claude/skills en conflicto, el directorio neutro se examina igual",
			preparar: func(d *discoEnMemoria) {
				d.fichero(".claude/skills", "no soy un directorio")
				d.fichero(".agents/skills/legal-core", "un fichero")
			},
			esperados: []instalacion.Conflicto{
				conflicto(fichero, ".agents/skills/legal-core"),
				conflicto(noEsDirectorio, ".claude/skills"),
			},
		},
		{
			nombre: "con --host claude y sin .claude, ninguna entrada de host existe",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "boe-legislacion", "legal-core").escribir()
			},
			invocacion: conHost,
		},
		{
			nombre:   "con .claude real y sin .claude/skills, ninguna entrada de host existe",
			preparar: func(d *discoEnMemoria) { d.directorio(".claude") },
		},
		{
			nombre: "con --dir no se examina nada de host",
			preparar: func(d *discoEnMemoria) {
				d.fichero(".claude/skills/legal-core", "un fichero")
			},
			invocacion:   instalacion.Invocacion{Dir: texto("destino")},
			noExaminadas: []string{".claude"},
		},
		{
			nombre:     "con -g, .claude es el de HOME",
			preparar:   func(d *discoEnMemoria) { d.fichero(homeDePrueba+"/.claude", "no soy un directorio") },
			invocacion: instalacion.Invocacion{Global: true, Host: texto("claude")},
			esperados:  []instalacion.Conflicto{conflicto(noEsDirectorio, homeDePrueba+"/.claude")},
		},
	}
}

// casosDelNeutro son los de la primera tabla de data-model §4.2: la entrada
// de cada skill pedida en el directorio neutro.
func casosDelNeutro() []casoDeConflictos {
	return []casoDeConflictos{
		{
			nombre:   "carpeta ajena, sin manifiesto",
			preparar: func(d *discoEnMemoria) { d.fichero(".agents/skills/boe-legislacion/mio.md", "no es de kitlegal") },
			esperados: []instalacion.Conflicto{
				conflicto(carpetaAjena, ".agents/skills/boe-legislacion"),
			},
			noExaminadas: []string{".agents/skills/boe-legislacion/mio.md", ".agents/skills/boe-legislacion/SKILL.md"},
		},
		{
			nombre: "carpeta ajena, con un manifiesto que no la declara",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "boe-legislacion").escribir()
				d.fichero(".agents/skills/legal-core/mio.md", "no es de kitlegal")
			},
			esperados: []instalacion.Conflicto{conflicto(carpetaAjena, ".agents/skills/legal-core")},
		},
		{
			nombre:    "fichero regular",
			preparar:  func(d *discoEnMemoria) { d.fichero(".agents/skills/boe-legislacion", "un fichero") },
			esperados: []instalacion.Conflicto{conflicto(fichero, ".agents/skills/boe-legislacion")},
		},
		{
			nombre:    "tubería con nombre, que no se abre",
			preparar:  func(d *discoEnMemoria) { d.tuberia(".agents/skills/boe-legislacion") },
			esperados: []instalacion.Conflicto{conflicto(fichero, ".agents/skills/boe-legislacion")},
		},
		{
			nombre: "enlace a otro sitio",
			preparar: func(d *discoEnMemoria) {
				d.fichero("otro-sitio/mio.md", "lo que hay al otro lado")
				d.enlace(".agents/skills/boe-legislacion", "../../otro-sitio")
			},
			esperados:    []instalacion.Conflicto{conflicto(enlaceAOtroSitio, ".agents/skills/boe-legislacion")},
			noExaminadas: []string{"otro-sitio"},
		},
		{
			nombre: "skill declarada sustituida por un enlace a una copia idéntica",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "boe-legislacion", "legal-core").escribir()
				instalarEn(d, instalacion.NuevoAmbitoDir("fuera"), "legal-core")
				d.enlace(".agents/skills/legal-core", "../../fuera/legal-core")
			},
			esperados:    []instalacion.Conflicto{conflicto(enlaceAOtroSitio, ".agents/skills/legal-core")},
			noExaminadas: []string{"fuera"},
		},
		{
			nombre:    "enlace colgando",
			preparar:  func(d *discoEnMemoria) { d.enlace(".agents/skills/boe-legislacion", "../../no-existe") },
			esperados: []instalacion.Conflicto{conflicto(enlaceRoto, ".agents/skills/boe-legislacion")},
		},
		{
			nombre:    "enlace a sí mismo",
			preparar:  func(d *discoEnMemoria) { d.enlace(".agents/skills/legal-core", "legal-core") },
			esperados: []instalacion.Conflicto{conflicto(enlaceRoto, ".agents/skills/legal-core")},
		},
		{
			nombre: "dos enlaces en ciclo",
			preparar: func(d *discoEnMemoria) {
				d.enlace(".agents/skills/boe-legislacion", "legal-core")
				d.enlace(".agents/skills/legal-core", "boe-legislacion")
			},
			esperados: []instalacion.Conflicto{
				conflicto(enlaceRoto, ".agents/skills/boe-legislacion"),
				conflicto(enlaceRoto, ".agents/skills/legal-core"),
			},
		},
		{
			nombre: "skill declarada sustituida por un enlace colgando",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "boe-legislacion", "legal-core").escribir()
				d.enlace(".agents/skills/legal-core", "/no-existe")
			},
			esperados: []instalacion.Conflicto{conflicto(enlaceRoto, ".agents/skills/legal-core")},
		},
		{
			nombre: "una en cada skill: se nombran todas, en orden de ruta",
			preparar: func(d *discoEnMemoria) {
				d.directorio(".claude/skills")
				d.enlace(".agents/skills/boe-legislacion", "../../no-existe")
				d.fichero(".agents/skills/legal-core", "un fichero")
				d.fichero(".claude/skills/boe-legislacion/mio.md", "no es de kitlegal")
				d.fichero("otro-sitio/mio.md", "lo que hay al otro lado")
				d.enlace(".claude/skills/legal-core", "../../otro-sitio")
			},
			esperados: []instalacion.Conflicto{
				conflicto(enlaceRoto, ".agents/skills/boe-legislacion"),
				conflicto(fichero, ".agents/skills/legal-core"),
				conflicto(carpetaAjena, ".claude/skills/boe-legislacion"),
				conflicto(enlaceAOtroSitio, ".claude/skills/legal-core"),
			},
		},
		{
			nombre: "un conflicto en la segunda skill basta: la primera no se instala",
			preparar: func(d *discoEnMemoria) {
				d.fichero(".agents/skills/legal-core/mio.md", "no es de kitlegal")
			},
			esperados: []instalacion.Conflicto{conflicto(carpetaAjena, ".agents/skills/legal-core")},
		},
		{
			nombre: "skill declarada cuyo directorio falta",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "boe-legislacion", "legal-core").escribir()
				d.retirar(".agents/skills/legal-core")
			},
		},
	}
}

// casosDentroDeUnaSkill son los de la segunda tabla de data-model §4.2: cada
// fichero y cada directorio intermedio de una skill declarada cuyo directorio
// es real.
func casosDentroDeUnaSkill() []casoDeConflictos {
	const skill = ".agents/skills/legal-core"

	return []casoDeConflictos{
		{
			nombre:   "instalación intacta",
			preparar: func(d *discoEnMemoria) { instalarLocal(d, "boe-legislacion", "legal-core").escribir() },
		},
		{
			nombre: "fichero editado",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "boe-legislacion", "legal-core").escribir()
				d.fichero(skill+"/SKILL.md", "editado a mano")
			},
			esperados: []instalacion.Conflicto{conflicto(ficheroEditado, skill+"/SKILL.md")},
		},
		{
			nombre: "SKILL.md sustituido por un enlace a una copia idéntica",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "legal-core").escribir()
				d.fichero("fuera/SKILL.md", "# legal-core\n")
				d.enlace(skill+"/SKILL.md", "../../../fuera/SKILL.md")
			},
			esperados:    []instalacion.Conflicto{conflicto(ficheroEditado, skill+"/SKILL.md")},
			noExaminadas: []string{"fuera"},
		},
		{
			nombre: "SKILL.md sustituido por un directorio",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "legal-core").escribir()
				d.retirar(skill + "/SKILL.md")
				d.fichero(skill+"/SKILL.md/nota.md", "un fichero")
			},
			esperados:    []instalacion.Conflicto{conflicto(ficheroEditado, skill+"/SKILL.md")},
			noExaminadas: []string{skill + "/SKILL.md/nota.md"},
		},
		{
			nombre: "SKILL.md sustituido por una tubería con nombre, que no se abre",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "legal-core").escribir()
				d.tuberia(skill + "/SKILL.md")
			},
			esperados: []instalacion.Conflicto{conflicto(ficheroEditado, skill+"/SKILL.md")},
		},
		{
			nombre: "declarado que ya no se empotra, editado",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "legal-core").declarar("legal-core", "legal-core/references/antigua.md", "antigua").escribir()
				d.fichero(skill+"/references/antigua.md", "editada")
			},
			esperados: []instalacion.Conflicto{conflicto(ficheroEditado, skill+"/references/antigua.md")},
		},
		{
			nombre: "declarado que ya no se empotra, sustituido por un enlace",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "legal-core").declarar("legal-core", "legal-core/references/antigua.md", "antigua").escribir()
				d.enlace(skill+"/references/antigua.md", "../SKILL.md")
			},
			esperados: []instalacion.Conflicto{conflicto(ficheroEditado, skill+"/references/antigua.md")},
		},
		{
			nombre: "declarados que ya no se empotran, uno intacto y otro que falta",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "legal-core").
					declarar("legal-core", "legal-core/references/antigua.md", "antigua").
					declarar("legal-core", "legal-core/references/perdida.md", "perdida").
					escribir()
				d.fichero(skill+"/references/antigua.md", "antigua")
			},
		},
		{
			nombre: "fichero empotrado que el manifiesto no declara",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "legal-core").olvidar("legal-core", "legal-core/references/jerarquia_normativa.md").escribir()
			},
			esperados: []instalacion.Conflicto{conflicto(ficheroAjeno, skill+"/references/jerarquia_normativa.md")},
		},
		{
			nombre: "enlace no declarado en una ruta empotrada, que no se sigue",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "legal-core").olvidar("legal-core", "legal-core/references/jerarquia_normativa.md").escribir()
				d.enlace(skill+"/references/jerarquia_normativa.md", "../SKILL.md")
			},
			esperados: []instalacion.Conflicto{conflicto(ficheroAjeno, skill+"/references/jerarquia_normativa.md")},
		},
		{
			nombre: "directorio no declarado en una ruta empotrada",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "legal-core").olvidar("legal-core", "legal-core/references/jerarquia_normativa.md").escribir()
				d.retirar(skill + "/references/jerarquia_normativa.md")
				d.fichero(skill+"/references/jerarquia_normativa.md/nota.md", "un fichero")
			},
			esperados:    []instalacion.Conflicto{conflicto(ficheroAjeno, skill+"/references/jerarquia_normativa.md")},
			noExaminadas: []string{skill + "/references/jerarquia_normativa.md/nota.md"},
		},
		{
			nombre: "references es un fichero",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "legal-core").escribir()
				d.fichero(skill+"/references", "no soy un directorio")
			},
			esperados: []instalacion.Conflicto{conflicto(noEsDirectorio, skill+"/references")},
		},
		{
			nombre: "references es un enlace a un directorio con los mismos ficheros",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "legal-core").escribir()
				instalarEn(d, instalacion.NuevoAmbitoDir("fuera"), "legal-core")
				d.enlace(skill+"/references", "../../../fuera/legal-core/references")
			},
			esperados:    []instalacion.Conflicto{conflicto(noEsDirectorio, skill+"/references")},
			noExaminadas: []string{"fuera"},
		},
		{
			nombre: "directorio intermedio de un declarado que ya no se empotra",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "legal-core").declarar("legal-core", "legal-core/antiguas/una.md", "antigua").escribir()
				d.fichero(skill+"/antiguas", "no soy un directorio")
			},
			esperados: []instalacion.Conflicto{conflicto(noEsDirectorio, skill+"/antiguas")},
		},
		{
			nombre: "una entrada que es a la vez intermedio y fichero declarado cae en una sola clase",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "legal-core").declarar("legal-core", "legal-core/SKILL.md/x.md", "x").escribir()
				d.fichero(skill+"/SKILL.md", "editado a mano")
			},
			esperados: []instalacion.Conflicto{conflicto(noEsDirectorio, skill+"/SKILL.md")},
		},
		{
			nombre: "fichero declarado que falta y references que falta entero",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "boe-legislacion", "legal-core").escribir()
				d.retirar(".agents/skills/boe-legislacion/SKILL.md")
				d.retirar(skill + "/references")
			},
		},
		{
			nombre: "lo no declarado que no coincide con ninguna ruta empotrada se deja intacto",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "legal-core").escribir()
				d.fichero(skill+"/notas.md", "mías")
				d.fichero(skill+"/references/mias.md", "mías")
				d.fichero(skill+"/borradores/uno.md", "mío")
				d.enlace(skill+"/atajo", "../../../no-existe")
			},
			noExaminadas: []string{
				skill + "/notas.md", skill + "/references/mias.md", skill + "/borradores", skill + "/atajo",
			},
		},
	}
}

// casosDelHost son los de data-model §4.3 en la entrada de host de cada skill
// pedida, con .claude real, salvo la copia declarada.
func casosDelHost() []casoDeConflictos {
	const host = ".claude/skills/legal-core"

	return []casoDeConflictos{
		{
			nombre:       "carpeta ajena no declarada",
			preparar:     func(d *discoEnMemoria) { d.fichero(host+"/mio.md", "no es de kitlegal") },
			esperados:    []instalacion.Conflicto{conflicto(carpetaAjena, host)},
			noExaminadas: []string{host + "/mio.md"},
		},
		{
			nombre: "carpeta ajena donde se declara un enlace",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "boe-legislacion", "legal-core").enlazar("legal-core").escribir()
				d.retirar(host)
				d.fichero(host+"/SKILL.md", "# legal-core\n")
			},
			esperados:    []instalacion.Conflicto{conflicto(carpetaAjena, host)},
			noExaminadas: []string{host + "/SKILL.md"},
		},
		{
			nombre:    "fichero regular",
			preparar:  func(d *discoEnMemoria) { d.fichero(host, "un fichero") },
			esperados: []instalacion.Conflicto{conflicto(fichero, host)},
		},
		{
			nombre: "tubería con nombre, que no se abre",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "legal-core").copiar("legal-core").escribir()
				d.tuberia(host)
			},
			esperados: []instalacion.Conflicto{conflicto(fichero, host)},
		},
		{
			nombre: "enlace a otro sitio",
			preparar: func(d *discoEnMemoria) {
				d.fichero("otro-sitio/mio.md", "lo que hay al otro lado")
				d.enlace(host, "../../otro-sitio")
			},
			esperados:    []instalacion.Conflicto{conflicto(enlaceAOtroSitio, host)},
			noExaminadas: []string{"otro-sitio"},
		},
		{
			nombre: "enlace roto con otro destino",
			preparar: func(d *discoEnMemoria) {
				d.directorio(".claude/skills")
				d.enlace(host, "../../no-existe")
			},
			esperados: []instalacion.Conflicto{conflicto(enlaceRoto, host)},
		},
		{
			nombre: "enlace al mismo sitio con otro destino literal",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "legal-core").escribir()
				d.enlace(host, directorioDeTrabajo+"/.agents/skills/legal-core")
			},
			esperados: []instalacion.Conflicto{conflicto(enlaceAOtroSitio, host)},
		},
		{
			nombre: "el enlace de FR-021 colgando, sin manifiesto, se adopta",
			preparar: func(d *discoEnMemoria) {
				d.directorio(".claude/skills")
				d.enlace(".claude/skills/boe-legislacion", "../../.agents/skills/boe-legislacion")
			},
		},
		{
			nombre: "el enlace de FR-021 que resuelve y no se declara se adopta",
			preparar: func(d *discoEnMemoria) {
				i := instalarLocal(d, "boe-legislacion", "legal-core").enlazar("legal-core")
				i.manifiesto.Skills["legal-core"] = instalacion.SkillDeclarada{
					Version:  versionDePrueba,
					Ficheros: i.manifiesto.Skills["legal-core"].Ficheros,
				}
				i.escribir()
			},
		},
		{
			nombre: "el enlace de FR-021 donde se declara una copia",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "boe-legislacion", "legal-core").copiar("legal-core").escribir()
				d.retirar(host)
				d.enlace(host, "../../.agents/skills/legal-core")
			},
		},
		{
			nombre: "con -g, la entrada de host es la de HOME",
			preparar: func(d *discoEnMemoria) {
				d.fichero(homeDePrueba+"/"+host+"/mio.md", "no es de kitlegal")
			},
			invocacion: global,
			esperados:  []instalacion.Conflicto{conflicto(carpetaAjena, homeDePrueba+"/"+host)},
		},
	}
}

// casosDeLaCopia son los de la copia de host declarada de data-model §4.3:
// si el creador de enlaces está disponible, la copia se va a retirar y todo
// lo que contiene tiene que estar declarado e intacto; si no, se mantiene y se
// actualiza con las reglas del directorio neutro.
func casosDeLaCopia() []casoDeConflictos {
	const host = ".claude/skills/legal-core"

	sonda := []string{".claude/skills"}

	return []casoDeConflictos{
		{
			nombre: "copia intacta que pasa a enlace, en las dos skills: una sola sonda",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "boe-legislacion", "legal-core").copiar("boe-legislacion").copiar("legal-core").escribir()
			},
			disponible: true,
			sondas:     sonda,
		},
		{
			nombre: "copia que pasa a enlace con un fichero editado",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "legal-core").copiar("legal-core").escribir()
				d.fichero(host+"/SKILL.md", "editado a mano")
			},
			disponible: true,
			esperados:  []instalacion.Conflicto{conflicto(ficheroEditado, host+"/SKILL.md")},
			sondas:     sonda,
		},
		{
			nombre: "copia que pasa a enlace con un fichero sustituido por un enlace",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "legal-core").copiar("legal-core").escribir()
				d.enlace(host+"/SKILL.md", "../../../.agents/skills/legal-core/SKILL.md")
			},
			disponible: true,
			esperados:  []instalacion.Conflicto{conflicto(ficheroEditado, host+"/SKILL.md")},
			sondas:     sonda,
		},
		{
			nombre: "copia que pasa a enlace con un fichero no declarado",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "legal-core").copiar("legal-core").escribir()
				d.fichero(host+"/references/mias.md", "mías")
			},
			disponible: true,
			esperados:  []instalacion.Conflicto{conflicto(ficheroAjeno, host+"/references/mias.md")},
			sondas:     sonda,
		},
		{
			nombre: "copia que pasa a enlace con un directorio no declarado",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "legal-core").copiar("legal-core").escribir()
				d.fichero(host+"/borradores/uno.md", "mío")
			},
			disponible:   true,
			esperados:    []instalacion.Conflicto{conflicto(ficheroAjeno, host+"/borradores")},
			sondas:       sonda,
			noExaminadas: []string{host + "/borradores/uno.md"},
		},
		{
			nombre: "copia que pasa a enlace con un empotrado que no declara",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "legal-core").copiar("legal-core").
					olvidar("legal-core", host+"/references/leyes_vertebrales.md").escribir()
			},
			disponible: true,
			esperados:  []instalacion.Conflicto{conflicto(ficheroAjeno, host+"/references/leyes_vertebrales.md")},
			sondas:     sonda,
		},
		{
			nombre: "copia que pasa a enlace con references que es un enlace",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "legal-core").copiar("legal-core").escribir()
				d.retirar(host + "/references")
				d.enlace(host+"/references", "../../../.agents/skills/legal-core/references")
			},
			disponible: true,
			esperados:  []instalacion.Conflicto{conflicto(noEsDirectorio, host+"/references")},
			sondas:     sonda,
		},
		{
			nombre: "copia que pasa a enlace con ficheros declarados que faltan",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "legal-core").copiar("legal-core").escribir()
				d.retirar(host + "/references")
			},
			disponible: true,
			sondas:     sonda,
		},
		{
			nombre: "copia que se mantiene, intacta",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "legal-core").copiar("legal-core").escribir()
			},
			sondas: sonda,
		},
		{
			nombre: "copia que se mantiene con un fichero editado",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "legal-core").copiar("legal-core").escribir()
				d.fichero(host+"/references/leyes_vertebrales.md", "editado a mano")
			},
			esperados: []instalacion.Conflicto{conflicto(ficheroEditado, host+"/references/leyes_vertebrales.md")},
			sondas:    sonda,
		},
		{
			nombre: "copia que se mantiene con lo no declarado que no coincide con ninguna ruta empotrada",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "legal-core").copiar("legal-core").escribir()
				d.fichero(host+"/references/mias.md", "mías")
			},
			sondas:       sonda,
			noExaminadas: []string{host + "/references/mias.md"},
		},
		{
			nombre: "copia que se mantiene con un empotrado que no declara",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "legal-core").copiar("legal-core").olvidar("legal-core", host+"/SKILL.md").escribir()
			},
			esperados: []instalacion.Conflicto{conflicto(ficheroAjeno, host+"/SKILL.md")},
			sondas:    sonda,
		},
		{
			nombre: "copia que se mantiene con references que es un fichero",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "legal-core").copiar("legal-core").escribir()
				d.fichero(host+"/references", "no soy un directorio")
			},
			esperados: []instalacion.Conflicto{conflicto(noEsDirectorio, host+"/references")},
			sondas:    sonda,
		},
	}
}

// casosQueNoSeTocan son los de data-model §4.4: una skill declarada que el
// binario no empotra y una declarada que no se pide ni se examinan (FR-034,
// FR-036).
func casosQueNoSeTocan() []casoDeConflictos {
	return []casoDeConflictos{
		{
			nombre: "skill declarada que el binario no empotra",
			preparar: func(d *discoEnMemoria) {
				i := instalarLocal(d, "boe-legislacion", "legal-core")
				i.manifiesto.Skills["otra-skill"] = instalacion.SkillDeclarada{
					Version:  "v0.0.9",
					Ficheros: map[string]string{"otra-skill/SKILL.md": instalacion.HuellaDe([]byte("# otra\n"))},
					Claude:   &instalacion.EntradaDeHost{Ruta: ".claude/skills/otra-skill", Modo: instalacion.ModoEnlace},
				}
				i.escribir()
				d.enlace(".agents/skills/otra-skill", "../../no-existe")
				d.fichero(".claude/skills/otra-skill", "un fichero")
			},
			noExaminadas: []string{".agents/skills/otra-skill", ".claude/skills/otra-skill"},
		},
		{
			nombre: "skill declarada que no se pide",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "boe-legislacion", "legal-core").copiar("legal-core").escribir()
				d.fichero(".agents/skills/legal-core/SKILL.md", "editado a mano")
				d.fichero(".claude/skills/legal-core/SKILL.md", "editado a mano")
			},
			invocacion:   instalacion.Invocacion{Skills: []string{"boe-legislacion"}},
			noExaminadas: []string{".agents/skills/legal-core", ".claude/skills/legal-core"},
		},
	}
}

// probarFallosAlComprobar exige que un fallo de entrada y salida del disco, o
// de la sonda del creador de enlaces, se devuelva tal cual y no como un
// conflicto, en cada llamada que comprobar hace al disco (data-model §3).
func probarFallosAlComprobar(t *testing.T) {
	t.Parallel()

	const host = ".claude/skills/legal-core"

	instalacionConCopia := func(d *discoEnMemoria) {
		instalarLocal(d, "boe-legislacion", "legal-core").copiar("legal-core").escribir()
	}

	casos := []struct {
		nombre     string
		preparar   func(d *discoEnMemoria)
		operacion  operacion
		ruta       string
		disponible bool
	}{
		{nombre: "una guarda", operacion: opExaminar, ruta: ".agents"},
		{nombre: "el manifiesto", preparar: instalacionConCopia, operacion: opExaminar, ruta: ".agents/skills/kitlegal.json"},
		{nombre: ".claude", operacion: opExaminar, ruta: ".claude"},
		{nombre: ".claude/skills", preparar: instalacionConCopia, operacion: opExaminar, ruta: ".claude/skills"},
		{nombre: "una skill", preparar: instalacionConCopia, operacion: opExaminar, ruta: ".agents/skills/legal-core"},
		{
			nombre: "un intermedio", preparar: instalacionConCopia,
			operacion: opExaminar, ruta: ".agents/skills/legal-core/references",
		},
		{
			nombre: "un intermedio por encima de otro",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "legal-core").declarar("legal-core", "legal-core/antiguas/sub/una.md", "antigua").escribir()
				d.fichero(".agents/skills/legal-core/antiguas/sub/una.md", "antigua")
			},
			operacion: opExaminar, ruta: ".agents/skills/legal-core/antiguas",
		},
		{nombre: "un fichero", preparar: instalacionConCopia, operacion: opExaminar, ruta: ".agents/skills/legal-core/SKILL.md"},
		{nombre: "una huella", preparar: instalacionConCopia, operacion: opHuella, ruta: ".agents/skills/legal-core/SKILL.md"},
		{nombre: "una entrada de host", preparar: instalacionConCopia, operacion: opExaminar, ruta: host},
		{nombre: "una copia", preparar: instalacionConCopia, operacion: opNombres, ruta: host, disponible: true},
		{
			nombre: "una entrada de la copia", preparar: instalacionConCopia,
			operacion: opExaminar, ruta: host + "/references", disponible: true,
		},
		{
			nombre: "un directorio de la copia", preparar: instalacionConCopia,
			operacion: opNombres, ruta: host + "/references", disponible: true,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			d := nuevoDiscoEnMemoria(t)
			if caso.preparar != nil {
				caso.preparar(d)
			}

			d.fallar(caso.operacion, caso.ruta, errInyectado)

			err := instalacion.ComprobarConflictos(d, &enlazadorDePrueba{disponible: caso.disponible},
				pedidoLocalConHost(t), empotradasDePrueba())
			exigirFalloSinConflictos(t, err, errInyectado)
		})
	}

	t.Run("la sonda del creador de enlaces", func(t *testing.T) {
		t.Parallel()

		d := nuevoDiscoEnMemoria(t)
		instalacionConCopia(d)

		err := instalacion.ComprobarConflictos(d, &enlazadorDePrueba{err: errInyectado},
			pedidoLocalConHost(t), empotradasDePrueba())
		exigirFalloSinConflictos(t, err, errInyectado)
	})
}

// pedidoLocalConHost es install de todas las skills en el ámbito local con
// --host claude.
func pedidoLocalConHost(t *testing.T) instalacion.Pedido {
	t.Helper()

	pedido, err := instalacion.ValidarInvocacion(conHost, "", empotradasDePrueba())
	require.NoError(t, err)

	return pedido
}

// exigirFalloSinConflictos exige que err sea el fallo esperado, tal cual, y
// no el rechazo por conflicto.
func exigirFalloSinConflictos(t *testing.T, err, esperado error) {
	t.Helper()

	require.ErrorIs(t, err, esperado)

	var rechazo *instalacion.ErrorDeConflictos
	assert.NotErrorAs(t, err, &rechazo, "un fallo de entrada y salida no es un conflicto")
}

// probarPedidoNoEmpotrado exige que un pedido con una skill que el binario no
// empotra —que ValidarInvocacion no deja pasar nunca— sea un error antes de
// examinar nada, y no un conflicto (FR-036).
func probarPedidoNoEmpotrado(t *testing.T) {
	t.Parallel()

	d := nuevoDiscoEnMemoria(t)
	pedido := instalacion.Pedido{Ambito: instalacion.NuevoAmbitoLocal(), Skills: []string{"legal-core", "otra-skill"}}

	err := instalacion.ComprobarConflictos(d, &enlazadorDePrueba{}, pedido, empotradasDePrueba())
	require.Error(t, err)
	assert.Contains(t, err.Error(), `"otra-skill"`)

	var rechazo *instalacion.ErrorDeConflictos
	assert.NotErrorAs(t, err, &rechazo)
	assert.Empty(t, d.accesos, "no se examina nada")
}

// casoDePlan es una fila de TestPlan: un disco sin ningún conflicto, una
// invocación y lo que install va a hacer en él.
type casoDePlan struct {
	nombre string
	// preparar deja el disco del caso, que empieza con el directorio de
	// trabajo vacío y el temporal.
	preparar func(d *discoEnMemoria)
	// invocacion es la de install; se valida con HOME=homeDePrueba.
	invocacion instalacion.Invocacion
	// admite dice en qué directorio funciona el creador de enlaces, que
	// sondea y enlaza con la misma respuesta; sin él, en todos.
	admite func(directorio string) bool
	// skills es la salida esperada, la misma con --dry-run.
	skills []instalacion.SkillInstalada
	// operaciones es el plan esperado, fase a fase, como lo escribe
	// operacionesDelPlan; ninguna si no hay nada que hacer.
	operaciones []string
	// sondas tiene cada directorio por el que se pregunta Disponible, en
	// orden.
	sondas []string
	// noExaminadas son rutas de las que no se puede examinar nada, ni ellas
	// ni lo que cuelga de ellas.
	noExaminadas []string
	// comprobar, si lo hay, comprueba lo propio del caso en el disco que deja
	// la aplicación del plan.
	comprobar func(t *testing.T, d *discoEnMemoria)
}

// ambitoLocal es el ámbito de casi todas las filas, y versionVieja la de un
// binario anterior.
var (
	ambitoLocal  = instalacion.NuevoAmbitoLocal()
	versionVieja = "v0.0.9"
)

// TestPlan fija el plan de install (data-model §5; research.md D7 y D9;
// FR-014, FR-015, FR-021, FR-024, FR-025, FR-033, FR-034, FR-036, FR-045 a
// FR-048, FR-051; SC-006, SC-007, SC-010, SC-012). En cada fila:
//
//   - planificar no cambia nada en el disco, que es todo lo que hace
//     --dry-run, y da por skill pedida su ruta, su estado y sus enlaces con su
//     modo, y las cuatro fases en su orden;
//   - Disponible solo se pregunta por el directorio de la sonda del ámbito,
//     cuando hay que decidir entre enlace y copia, y una vez por directorio;
//   - aplicado, cada entrada de host queda en el modo previsto, así que la
//     salida de la orden es la de --dry-run (FR-048);
//   - y la segunda ejecución da «sin cambios» en cada skill, con el plan vacío
//     y el disco byte a byte igual (FR-045, SC-007).
//
// Ningún fichero sale con permiso de ejecución (FR-015): el plan solo escribe
// ficheros regulares con EscribirFichero, que no lo da, y no tiene forma de
// pedir otro modo.
func TestPlan(t *testing.T) {
	t.Parallel()

	grupos := []struct {
		nombre string
		casos  []casoDePlan
	}{
		{nombre: "instalación nueva", casos: casosDeInstalacionNueva(t)},
		{nombre: "global y --dir", casos: casosDeAmbitosDelPlan(t)},
		{nombre: "actualización", casos: casosDeActualizacion(t)},
		{nombre: "lo que no se toca", casos: casosDelPlanQueNoSeTocan(t)},
		{nombre: "entradas de host", casos: casosDelPlanDelHost(t)},
		{nombre: "copias de host", casos: casosDelPlanDeLasCopias()},
		{nombre: "enlaces solo fuera del ámbito", casos: casosConEnlacesSoloFueraDelAmbito(t)},
	}

	for _, grupo := range grupos {
		t.Run(grupo.nombre, func(t *testing.T) {
			t.Parallel()

			probarCasosDePlan(t, grupo.casos)
		})
	}

	t.Run("recurso de copia cuando el enlace falla al aplicar", probarRecursoDeCopia)
	t.Run("con un conflicto no hay plan", probarPlanConConflictos)
	t.Run("fallos al planificar", probarFallosAlPlanificar)
	t.Run("la salida", probarSalidaDeInstall)
}

// probarCasosDePlan comprueba cada caso en su propio disco.
func probarCasosDePlan(t *testing.T, casos []casoDePlan) {
	t.Helper()

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			probarCasoDePlan(t, caso)
		})
	}
}

// probarCasoDePlan planifica el caso sobre su disco, lo compara con lo
// esperado, lo aplica y planifica una segunda vez.
func probarCasoDePlan(t *testing.T, caso casoDePlan) {
	t.Helper()

	d := nuevoDiscoEnMemoria(t)
	d.directorio(temporal)
	caso.preparar(d)

	pedido, err := instalacion.ValidarInvocacion(caso.invocacion, homeDePrueba, empotradasDePrueba())
	require.NoError(t, err, "la invocación del caso")

	admite := caso.admite
	if admite == nil {
		admite = admiteSiempre
	}

	enlazador := nuevoEnlazadorHonesto(d, admite)
	antes := d.instantanea()

	plan, err := instalacion.Planificar(d, enlazador, pedido, empotradasDePrueba(), versionDePrueba)
	require.NoError(t, err)

	assert.Equal(t, antes, d.instantanea(), "planificar, que es todo lo que hace --dry-run, no cambia nada (FR-048)")
	assert.Equal(t, caso.skills, plan.Skills, "la salida de la orden y de --dry-run (FR-051)")
	assert.Equal(t, caso.operaciones, operacionesDelPlan(t, plan), "las cuatro fases, en orden (D7)")
	assert.Equal(t, caso.sondas, enlazador.preguntados, "Disponible, solo por el directorio de la sonda y una vez")
	exigirDiscoRespetado(t, d, raizVigilada(pedido.Ambito))
	exigirNoExaminadas(t, d, caso.noExaminadas...)

	enCopia := aplicarEnMemoria(t, plan, &escritorEnMemoria{disco: d, enlazador: enlazador})
	assert.Empty(t, enCopia, "cada entrada de host queda en el modo previsto: --dry-run dice lo que hace la orden")
	exigirHostsEnElDisco(t, d, plan.Skills)

	if caso.comprobar != nil {
		caso.comprobar(t, d)
	}

	exigirSegundaSinCambios(t, d, pedido, admite, plan.Skills)
}

// operacionesDelPlan es el plan escrito una operación por línea, con el número
// de su fase delante: lo que se retira, cada directorio que se crea, cada
// enlace, el manifiesto si cambia con los modos previstos y cada fichero.
func operacionesDelPlan(t *testing.T, plan instalacion.Plan) []string {
	t.Helper()

	var operaciones []string

	anotar := func(fase, operacion string, rutas ...string) {
		for _, ruta := range rutas {
			operaciones = append(operaciones, fase+" "+operacion+" "+ruta)
		}
	}

	anotar("1", "retirar", plan.Retirar...)
	anotar("2", "crear", plan.Enlazar.DirectoriosQueFaltan...)

	for _, enlace := range plan.Enlazar.Enlaces {
		anotar("2", "enlazar", enlace.Ruta+" -> "+enlace.Destino)
	}

	anotar("3", "crear", plan.Manifiesto.DirectoriosQueFaltan...)

	contenido, err := plan.Manifiesto.Contenido(nil)
	require.NoError(t, err, "el manifiesto final")

	if contenido != nil {
		anotar("3", "manifiesto", plan.Manifiesto.Ruta)
	}

	anotar("4", "crear", plan.Escribir.DirectoriosQueFaltan...)

	for _, fichero := range plan.Escribir.Ficheros {
		anotar("4", "escribir", fichero.Ruta)
	}

	return operaciones
}

// exigirHostsEnElDisco exige que cada entrada de host de la salida esté en el
// disco en su modo: el enlace de FR-021 o un directorio real con la copia.
func exigirHostsEnElDisco(t *testing.T, d *discoEnMemoria, skills []instalacion.SkillInstalada) {
	t.Helper()

	for _, skill := range skills {
		for _, enlace := range skill.Enlaces {
			entrada := examinarEnMemoria(t, d, enlace.Ruta)

			switch enlace.Modo {
			case instalacion.ModoEnlace:
				assert.Equal(t, enlaceEnMemoria("../../.agents/skills/"+skill.Nombre, true), entrada,
					"%s, el enlace de FR-021", enlace.Ruta)
			case instalacion.ModoCopia:
				assert.Equal(t, instalacion.EntradaDirectorio, entrada.Tipo, "%s, una copia", enlace.Ruta)
			}
		}
	}
}

// exigirSegundaSinCambios planifica de nuevo sobre lo que dejó la primera
// ejecución, con un creador de enlaces que responde lo mismo, y exige «sin
// cambios» en cada skill, con los mismos enlaces, el plan vacío, Disponible
// preguntado como mucho una vez por directorio y el disco byte a byte igual
// tras aplicarlo (FR-045, SC-007).
func exigirSegundaSinCambios(
	t *testing.T, d *discoEnMemoria, pedido instalacion.Pedido, admite func(string) bool,
	primera []instalacion.SkillInstalada,
) {
	t.Helper()

	enlazador := nuevoEnlazadorHonesto(d, admite)
	antes := d.instantanea()

	plan, err := instalacion.Planificar(d, enlazador, pedido, empotradasDePrueba(), versionDePrueba)
	require.NoError(t, err, "la segunda ejecución no encuentra ningún conflicto")

	esperadas := make([]instalacion.SkillInstalada, 0, len(primera))
	for _, skill := range primera {
		skill.Estado = instalacion.EstadoSinCambios
		esperadas = append(esperadas, skill)
	}

	assert.Equal(t, esperadas, plan.Skills, "la segunda ejecución: sin cambios en cada skill")
	assert.Empty(t, operacionesDelPlan(t, plan), "la segunda ejecución: el plan vacío")
	assert.Empty(t, aplicarEnMemoria(t, plan, &escritorEnMemoria{disco: d, enlazador: enlazador}))
	assert.Equal(t, antes, d.instantanea(), "la segunda ejecución deja el disco byte a byte igual")

	preguntados := slices.Sorted(slices.Values(enlazador.preguntados))
	assert.Len(t, enlazador.preguntados, len(slices.Compact(preguntados)), "Disponible, una vez por directorio")
}

// salidaEn es la skill nombre del ámbito con ese estado y, por cada modo, su
// entrada en el host claude con ese modo.
func salidaEn(
	ambito instalacion.Ambito, nombre string, estado instalacion.Estado, modos ...instalacion.Modo,
) instalacion.SkillInstalada {
	enlaces := []instalacion.Enlace{}
	for _, modo := range modos {
		enlaces = append(enlaces, instalacion.Enlace{Host: "claude", Ruta: ambito.RutaDeHost(nombre), Modo: modo})
	}

	return instalacion.SkillInstalada{Nombre: nombre, Ruta: ambito.RutaDeSkill(nombre), Estado: estado, Enlaces: enlaces}
}

// lasDos es la salida de las dos skills empotradas en el ámbito, en orden de
// nombre, con el mismo estado y los mismos modos.
func lasDos(ambito instalacion.Ambito, estado instalacion.Estado, modos ...instalacion.Modo) []instalacion.SkillInstalada {
	return []instalacion.SkillInstalada{
		salidaEn(ambito, "boe-legislacion", estado, modos...),
		salidaEn(ambito, "legal-core", estado, modos...),
	}
}

// escribirEnteras son las operaciones de la fase 4 que escriben entera la
// skill empotrada de cada ruta —su directorio en el neutro o su copia de
// host, cuyo último elemento es el nombre de la skill—: primero cada
// directorio de todas, de arriba abajo, y después cada fichero.
func escribirEnteras(t *testing.T, rutas ...string) []string {
	t.Helper()

	var creados, escritos []string

	for _, ruta := range rutas {
		creados = append(creados, "4 crear "+ruta)

		for _, fichero := range empotradaDePrueba(t, path.Base(ruta)).Ficheros {
			crear := "4 crear " + path.Join(ruta, path.Dir(fichero.Ruta))
			if !slices.Contains(creados, crear) {
				creados = append(creados, crear)
			}

			escritos = append(escritos, "4 escribir "+path.Join(ruta, fichero.Ruta))
		}
	}

	return append(creados, escritos...)
}

// manifiestoNuevo son las operaciones de la fase 3: cada directorio que falta
// hasta el neutro y el manifiesto del ámbito.
func manifiestoNuevo(ambito instalacion.Ambito, faltan ...string) []string {
	operaciones := make([]string, 0, len(faltan)+1)
	for _, dir := range faltan {
		operaciones = append(operaciones, "3 crear "+dir)
	}

	return append(operaciones, "3 manifiesto "+ambito.RutaDelManifiesto())
}

// enlazarLasDos son las operaciones de la fase 2: cada directorio que falta
// hasta .claude/skills y el enlace de FR-021 de cada skill empotrada.
func enlazarLasDos(ambito instalacion.Ambito, faltan ...string) []string {
	operaciones := make([]string, 0, len(faltan)+2)
	for _, dir := range faltan {
		operaciones = append(operaciones, "2 crear "+dir)
	}

	for _, nombre := range []string{"boe-legislacion", "legal-core"} {
		operaciones = append(operaciones, "2 enlazar "+ambito.RutaDeHost(nombre)+" -> ../../.agents/skills/"+nombre)
	}

	return operaciones
}

// lasDosEn son las rutas de las dos skills empotradas en el directorio neutro
// del ámbito, y lasDosConCopia, además, cada una seguida de su copia de host.
func lasDosEn(ambito instalacion.Ambito) []string {
	return []string{ambito.RutaDeSkill("boe-legislacion"), ambito.RutaDeSkill("legal-core")}
}

func lasDosConCopia(ambito instalacion.Ambito) []string {
	return []string{
		ambito.RutaDeSkill("boe-legislacion"), ambito.RutaDeHost("boe-legislacion"),
		ambito.RutaDeSkill("legal-core"), ambito.RutaDeHost("legal-core"),
	}
}

// casosDeInstalacionNueva son los de un ámbito local en el que no hay nada:
// cada skill sale «instalada», con cada directorio que falta creado (FR-014)
// y, si se enlaza en el host, con su enlace o su copia (FR-021 a FR-025).
func casosDeInstalacionNueva(t *testing.T) []casoDePlan {
	t.Helper()

	nuevo := manifiestoNuevo(ambitoLocal, ".agents", ".agents/skills")

	return []casoDePlan{
		{
			nombre:      "sin .claude, sin hosts",
			preparar:    func(*discoEnMemoria) {},
			skills:      lasDos(ambitoLocal, instalacion.EstadoInstalada),
			operaciones: slices.Concat(nuevo, escribirEnteras(t, lasDosEn(ambitoLocal)...)),
			comprobar: func(t *testing.T, d *discoEnMemoria) {
				t.Helper()

				manifiesto := leerManifiestoDelDisco(t, d)
				assert.Equal(t, versionDePrueba, manifiesto.Version)
				assert.Equal(t, huellasEmpotradas(t, "legal-core", "legal-core/"), manifiesto.Skills["legal-core"].Ficheros)
				assert.Nil(t, manifiesto.Skills["legal-core"].Claude, "sin hosts")
			},
		},
		{
			nombre:      "una sola skill: un directorio y un manifiesto que declara una (SC-006)",
			preparar:    func(*discoEnMemoria) {},
			invocacion:  instalacion.Invocacion{Skills: []string{"legal-core"}},
			skills:      []instalacion.SkillInstalada{salidaEn(ambitoLocal, "legal-core", instalacion.EstadoInstalada)},
			operaciones: slices.Concat(nuevo, escribirEnteras(t, ".agents/skills/legal-core")),
			comprobar: func(t *testing.T, d *discoEnMemoria) {
				t.Helper()

				nombres, err := d.Nombres(".agents/skills")
				require.NoError(t, err)
				assert.Equal(t, []string{"kitlegal.json", "legal-core"}, nombres)
				assert.Len(t, leerManifiestoDelDisco(t, d).Skills, 1)
			},
		},
		{
			nombre:   "con .claude, un enlace relativo por skill",
			preparar: func(d *discoEnMemoria) { d.directorio(".claude") },
			skills:   lasDos(ambitoLocal, instalacion.EstadoInstalada, instalacion.ModoEnlace),
			operaciones: slices.Concat(enlazarLasDos(ambitoLocal, ".claude/skills"), nuevo,
				escribirEnteras(t, lasDosEn(ambitoLocal)...)),
			sondas: []string{".claude"},
		},
		{
			nombre:     "--host claude sin .claude: se crea",
			preparar:   func(*discoEnMemoria) {},
			invocacion: conHost,
			skills:     lasDos(ambitoLocal, instalacion.EstadoInstalada, instalacion.ModoEnlace),
			operaciones: slices.Concat(enlazarLasDos(ambitoLocal, ".claude", ".claude/skills"), nuevo,
				escribirEnteras(t, lasDosEn(ambitoLocal)...)),
			sondas: []string{"."},
		},
		{
			nombre:   "con .claude y un creador de enlaces que no funciona, copias (FR-024)",
			preparar: func(d *discoEnMemoria) { d.directorio(".claude") },
			admite:   admiteNunca,
			skills:   lasDos(ambitoLocal, instalacion.EstadoInstalada, instalacion.ModoCopia),
			operaciones: slices.Concat([]string{"2 crear .claude/skills"}, nuevo,
				escribirEnteras(t, lasDosConCopia(ambitoLocal)...)),
			sondas: []string{".claude"},
			comprobar: func(t *testing.T, d *discoEnMemoria) {
				t.Helper()

				copia := leerManifiestoDelDisco(t, d).Skills["legal-core"].Claude
				require.NotNil(t, copia)
				assert.Equal(t, instalacion.ModoCopia, copia.Modo)
				assert.Equal(t, huellasEmpotradas(t, "legal-core", ".claude/skills/legal-core/"), copia.Ficheros,
					"la copia declara la huella de cada fichero copiado")
			},
		},
	}
}

// casosDeAmbitosDelPlan son los del ámbito global y el de --dir, con las rutas
// como se alcanzan desde el directorio de trabajo y cada directorio que falta
// hasta el neutro y hasta .claude/skills, de arriba abajo (FR-014).
func casosDeAmbitosDelPlan(t *testing.T) []casoDePlan {
	t.Helper()

	deHome := ambitoGlobal(t, homeDePrueba)
	dir := instalacion.NuevoAmbitoDir("otro/destino/")
	conHostGlobal := instalacion.Invocacion{Global: true, Host: texto("claude")}
	nuevo := manifiestoNuevo(deHome, homeDePrueba+"/.agents", homeDePrueba+"/.agents/skills")

	return []casoDePlan{
		{
			nombre:     "-g con --host claude: la sonda, en HOME",
			preparar:   func(d *discoEnMemoria) { d.directorio(homeDePrueba) },
			invocacion: conHostGlobal,
			skills:     lasDos(deHome, instalacion.EstadoInstalada, instalacion.ModoEnlace),
			operaciones: slices.Concat(enlazarLasDos(deHome, homeDePrueba+"/.claude", homeDePrueba+"/.claude/skills"),
				nuevo, escribirEnteras(t, lasDosEn(deHome)...)),
			sondas: []string{homeDePrueba},
		},
		{
			// La raíz se usa tal cual (FR-027): un HOME que es un enlace que
			// resuelve a un directorio existe, y es donde se sondea; no es el
			// HOME que no existe de la fila siguiente, que no se sondea.
			nombre: "-g con --host claude y un HOME que es un enlace a un directorio: la sonda, en HOME (D9)",
			preparar: func(d *discoEnMemoria) {
				d.directorio("/usuarios/ana")
				d.enlace(homeDePrueba, "/usuarios/ana")
			},
			invocacion: conHostGlobal,
			skills:     lasDos(deHome, instalacion.EstadoInstalada, instalacion.ModoEnlace),
			operaciones: slices.Concat(enlazarLasDos(deHome, homeDePrueba+"/.claude", homeDePrueba+"/.claude/skills"),
				nuevo, escribirEnteras(t, lasDosEn(deHome)...)),
			sondas: []string{homeDePrueba},
		},
		{
			nombre:     "-g con un HOME que no existe: se crea, sin sonda, y se predice enlace (D9)",
			preparar:   func(*discoEnMemoria) {},
			invocacion: conHostGlobal,
			skills:     lasDos(deHome, instalacion.EstadoInstalada, instalacion.ModoEnlace),
			operaciones: slices.Concat(
				enlazarLasDos(deHome, "/home", homeDePrueba, homeDePrueba+"/.claude", homeDePrueba+"/.claude/skills"),
				nuevo, escribirEnteras(t, lasDosEn(deHome)...)),
		},
		{
			nombre:     "--dir, con lo que falta por encima, y un .claude que no se mira",
			preparar:   func(d *discoEnMemoria) { d.directorio(".claude") },
			invocacion: instalacion.Invocacion{Dir: texto("otro/destino/")},
			skills:     lasDos(dir, instalacion.EstadoInstalada),
			operaciones: slices.Concat(manifiestoNuevo(dir, "otro", "otro/destino"),
				escribirEnteras(t, lasDosEn(dir)...)),
			noExaminadas: []string{".claude"},
		},
	}
}

// casosDeActualizacion son los de una instalación que ya está: nada que
// cambiar, ficheros declarados que faltan, otro binario y ficheros que se
// dejan de empotrar (FR-045, FR-046).
func casosDeActualizacion(t *testing.T) []casoDePlan {
	t.Helper()

	const skill = ".agents/skills/legal-core"

	return []casoDePlan{
		{
			nombre: "nada que cambiar: sin cambios y el plan vacío",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "boe-legislacion", "legal-core").enlazar("boe-legislacion").enlazar("legal-core").escribir()
			},
			skills: lasDos(ambitoLocal, instalacion.EstadoSinCambios, instalacion.ModoEnlace),
		},
		{
			nombre: "ficheros declarados que faltan, repuestos sin tocar el manifiesto",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "boe-legislacion", "legal-core").escribir()
				d.retirar(".agents/skills/boe-legislacion/SKILL.md")
				d.retirar(skill + "/references")
			},
			skills: lasDos(ambitoLocal, instalacion.EstadoActualizada),
			operaciones: []string{
				"4 crear " + skill + "/references",
				"4 escribir .agents/skills/boe-legislacion/SKILL.md",
				"4 escribir " + skill + "/references/jerarquia_normativa.md",
				"4 escribir " + skill + "/references/leyes_vertebrales.md",
			},
		},
		{
			nombre: "otro binario: se retira y se reescribe lo que difiere, y se declara la versión",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "boe-legislacion", "legal-core").deVersion(versionVieja).
					deOtroBinario("legal-core", "SKILL.md", "# legal-core de antes\n").escribir()
			},
			skills: lasDos(ambitoLocal, instalacion.EstadoActualizada),
			operaciones: []string{
				"1 retirar " + skill + "/SKILL.md",
				"3 manifiesto .agents/skills/kitlegal.json",
				"4 escribir " + skill + "/SKILL.md",
			},
		},
		{
			nombre: "lo que se deja de empotrar se retira si está intacto y se quita del manifiesto si falta",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "legal-core").
					deOtroBinario("legal-core", "references/antigua.md", "antigua").
					deOtroBinario("legal-core", "antiguas/una.md", "una").
					declarar("legal-core", "legal-core/references/perdida.md", "perdida").escribir()
			},
			invocacion: instalacion.Invocacion{Skills: []string{"legal-core"}},
			skills:     []instalacion.SkillInstalada{salidaEn(ambitoLocal, "legal-core", instalacion.EstadoActualizada)},
			operaciones: []string{
				"1 retirar " + skill + "/antiguas/una.md",
				"1 retirar " + skill + "/references/antigua.md",
				"3 manifiesto .agents/skills/kitlegal.json",
			},
			comprobar: func(t *testing.T, d *discoEnMemoria) {
				t.Helper()

				assert.Equal(t, huellasEmpotradas(t, "legal-core", "legal-core/"),
					leerManifiestoDelDisco(t, d).Skills["legal-core"].Ficheros, "declara solo lo empotrado")
				assert.Equal(t, instalacion.EntradaDirectorio, examinarEnMemoria(t, d, skill+"/antiguas").Tipo,
					"el directorio que se queda vacío no está declarado y no se retira (FR-047)")
			},
		},
	}
}

// casosDelPlanQueNoSeTocan son los de un subconjunto y de una skill que el
// binario no empotra: sus entradas del manifiesto se conservan byte a byte y
// no se examina nada suyo (FR-034, FR-036).
func casosDelPlanQueNoSeTocan(t *testing.T) []casoDePlan {
	t.Helper()

	var antesDelSubconjunto, antesSinNombres []byte

	otraSkill := instalacion.SkillDeclarada{
		Version:  versionVieja,
		Ficheros: map[string]string{"otra-skill/SKILL.md": instalacion.HuellaDe([]byte("# otra\n"))},
		Claude:   &instalacion.EntradaDeHost{Ruta: ".claude/skills/otra-skill", Modo: instalacion.ModoEnlace},
	}

	return []casoDePlan{
		{
			nombre: "un subconjunto conserva las demás entradas del manifiesto",
			preparar: func(d *discoEnMemoria) {
				i := instalarLocal(d, "boe-legislacion").deVersion(versionVieja)
				i.manifiesto.Skills["otra-skill"] = otraSkill
				i.escribir()
				d.fichero(".agents/skills/boe-legislacion/SKILL.md", "editado a mano")
				antesDelSubconjunto = leerDelDisco(t, d, ambitoLocal.RutaDelManifiesto())
			},
			invocacion:   instalacion.Invocacion{Skills: []string{"legal-core"}},
			skills:       []instalacion.SkillInstalada{salidaEn(ambitoLocal, "legal-core", instalacion.EstadoInstalada)},
			operaciones:  slices.Concat(manifiestoNuevo(ambitoLocal), escribirEnteras(t, ".agents/skills/legal-core")),
			noExaminadas: []string{".agents/skills/boe-legislacion", ".agents/skills/otra-skill"},
			comprobar: func(t *testing.T, d *discoEnMemoria) {
				t.Helper()

				exigirEntradasConservadas(t, antesDelSubconjunto, leerDelDisco(t, d, ambitoLocal.RutaDelManifiesto()),
					"boe-legislacion", "otra-skill")
			},
		},
		{
			nombre: "sin nombres, la skill que el binario no empotra se queda como está",
			preparar: func(d *discoEnMemoria) {
				i := instalarLocal(d, "boe-legislacion", "legal-core").deVersion(versionVieja)
				i.manifiesto.Skills["otra-skill"] = otraSkill
				i.escribir()
				d.enlace(".agents/skills/otra-skill", "../../no-existe")
				antesSinNombres = leerDelDisco(t, d, ambitoLocal.RutaDelManifiesto())
			},
			skills:       lasDos(ambitoLocal, instalacion.EstadoActualizada),
			operaciones:  []string{"3 manifiesto .agents/skills/kitlegal.json"},
			noExaminadas: []string{".agents/skills/otra-skill"},
			comprobar: func(t *testing.T, d *discoEnMemoria) {
				t.Helper()

				exigirEntradasConservadas(t, antesSinNombres, leerDelDisco(t, d, ambitoLocal.RutaDelManifiesto()),
					"otra-skill")
				assert.Equal(t, enlaceEnMemoria("../../no-existe", false),
					examinarEnMemoria(t, d, ".agents/skills/otra-skill"))
			},
		},
	}
}

// casosDelPlanDelHost son los de las entradas de host declaradas que faltan,
// que se recrean o se quitan del manifiesto, y los del enlace de FR-021 que se
// adopta (FR-021 a FR-023, FR-041, FR-046).
func casosDelPlanDelHost(t *testing.T) []casoDePlan {
	t.Helper()

	enlazadas := func(d *discoEnMemoria) *instalada {
		return instalarLocal(d, "boe-legislacion", "legal-core").enlazar("boe-legislacion").enlazar("legal-core")
	}

	boeSinCambios := salidaEn(ambitoLocal, "boe-legislacion", instalacion.EstadoSinCambios, instalacion.ModoEnlace)
	legalActualizada := salidaEn(ambitoLocal, "legal-core", instalacion.EstadoActualizada, instalacion.ModoEnlace)

	return []casoDePlan{
		{
			nombre: "la entrada de host que falta se recrea",
			preparar: func(d *discoEnMemoria) {
				enlazadas(d).escribir()
				d.retirar(".claude/skills/legal-core")
			},
			skills:      []instalacion.SkillInstalada{boeSinCambios, legalActualizada},
			operaciones: []string{"2 enlazar .claude/skills/legal-core -> ../../.agents/skills/legal-core"},
			sondas:      []string{".claude/skills"},
		},
		{
			nombre: "sin .claude y con --host claude, se recrean las dos",
			preparar: func(d *discoEnMemoria) {
				enlazadas(d).escribir()
				d.retirar(".claude")
			},
			invocacion:  conHost,
			skills:      lasDos(ambitoLocal, instalacion.EstadoActualizada, instalacion.ModoEnlace),
			operaciones: enlazarLasDos(ambitoLocal, ".claude", ".claude/skills"),
			sondas:      []string{"."},
		},
		{
			nombre: "sin .claude ni --host, se quita del manifiesto",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "boe-legislacion", "legal-core").enlazar("legal-core").escribir()
				d.retirar(".claude")
			},
			skills: []instalacion.SkillInstalada{
				salidaEn(ambitoLocal, "boe-legislacion", instalacion.EstadoSinCambios),
				salidaEn(ambitoLocal, "legal-core", instalacion.EstadoActualizada),
			},
			operaciones: []string{"3 manifiesto .agents/skills/kitlegal.json"},
			comprobar: func(t *testing.T, d *discoEnMemoria) {
				t.Helper()

				assert.Nil(t, leerManifiestoDelDisco(t, d).Skills["legal-core"].Claude)
			},
		},
		{
			nombre: "un .claude que no es un directorio real, sin --host: se quita sin mirar debajo",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "legal-core").enlazar("legal-core").escribir()
				d.fichero(".claude", "no soy un directorio")
			},
			invocacion:   instalacion.Invocacion{Skills: []string{"legal-core"}},
			skills:       []instalacion.SkillInstalada{salidaEn(ambitoLocal, "legal-core", instalacion.EstadoActualizada)},
			operaciones:  []string{"3 manifiesto .agents/skills/kitlegal.json"},
			noExaminadas: []string{".claude/skills"},
		},
		{
			nombre: "el enlace de FR-021 colgando y sin declarar se adopta, y la skill se instala",
			preparar: func(d *discoEnMemoria) {
				d.directorio(".claude/skills")
				d.enlace(".claude/skills/boe-legislacion", "../../.agents/skills/boe-legislacion")
			},
			skills: lasDos(ambitoLocal, instalacion.EstadoInstalada, instalacion.ModoEnlace),
			operaciones: slices.Concat(
				[]string{"2 enlazar .claude/skills/legal-core -> ../../.agents/skills/legal-core"},
				manifiestoNuevo(ambitoLocal, ".agents", ".agents/skills"), escribirEnteras(t, lasDosEn(ambitoLocal)...)),
			sondas: []string{".claude/skills"},
		},
		{
			nombre: "el enlace de FR-021 donde se declara una copia se declara enlace",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "boe-legislacion", "legal-core").enlazar("boe-legislacion").copiar("legal-core").escribir()
				d.retirar(".claude/skills/legal-core")
				d.enlace(".claude/skills/legal-core", "../../.agents/skills/legal-core")
			},
			skills:      []instalacion.SkillInstalada{boeSinCambios, legalActualizada},
			operaciones: []string{"3 manifiesto .agents/skills/kitlegal.json"},
		},
	}
}

// casosDelPlanDeLasCopias son los de una copia de host declarada: pasa a
// enlace si el creador de enlaces funciona en .claude/skills y, si no, se
// mantiene y se actualiza como el directorio neutro (FR-024, FR-046).
func casosDelPlanDeLasCopias() []casoDePlan {
	const host = ".claude/skills/legal-core"

	return []casoDePlan{
		{
			// La copia lleva, además de lo empotrado, un fichero declarado de
			// otro binario dos niveles por debajo de references: retirarla de
			// arriba abajo chocaría con references, que aún no está vacío.
			nombre: "la copia pasa a enlace: se retira entera, de abajo arriba, y se enlaza",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "boe-legislacion", "legal-core").copiar("legal-core").
					copiaDeOtroBinario("legal-core", "references/antiguas/una.md", "una").escribir()
			},
			skills: lasDos(ambitoLocal, instalacion.EstadoActualizada, instalacion.ModoEnlace),
			operaciones: slices.Concat([]string{
				"1 retirar " + host + "/SKILL.md",
				"1 retirar " + host + "/references/antiguas/una.md",
				"1 retirar " + host + "/references/jerarquia_normativa.md",
				"1 retirar " + host + "/references/leyes_vertebrales.md",
				"1 retirar " + host + "/references/antiguas",
				"1 retirar " + host + "/references",
				"1 retirar " + host,
			}, enlazarLasDos(ambitoLocal), []string{"3 manifiesto .agents/skills/kitlegal.json"}),
			sondas: []string{".claude/skills"},
		},
		{
			nombre: "la copia que se mantiene, sin nada que actualizar, sale sin cambios",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "boe-legislacion", "legal-core").copiar("boe-legislacion").copiar("legal-core").escribir()
			},
			admite: admiteNunca,
			skills: lasDos(ambitoLocal, instalacion.EstadoSinCambios, instalacion.ModoCopia),
			sondas: []string{".claude/skills"},
		},
		{
			nombre: "la copia que se mantiene se actualiza como el directorio neutro",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "legal-core").copiar("legal-core").
					copiaDeOtroBinario("legal-core", "SKILL.md", "# legal-core de antes\n").escribir()
				d.retirar(host + "/references")
			},
			invocacion: instalacion.Invocacion{Skills: []string{"legal-core"}},
			admite:     admiteNunca,
			skills: []instalacion.SkillInstalada{
				salidaEn(ambitoLocal, "legal-core", instalacion.EstadoActualizada, instalacion.ModoCopia),
			},
			operaciones: []string{
				"1 retirar " + host + "/SKILL.md",
				"3 manifiesto .agents/skills/kitlegal.json",
				"4 crear " + host + "/references",
				"4 escribir " + host + "/SKILL.md",
				"4 escribir " + host + "/references/jerarquia_normativa.md",
				"4 escribir " + host + "/references/leyes_vertebrales.md",
			},
			sondas: []string{".claude/skills"},
		},
	}
}

// casosConEnlacesSoloFueraDelAmbito son los de research.md D9, con el
// Enlazador sintético que admite enlaces en el temporal, fuera del ámbito, y
// en ningún directorio del ámbito. Como la sonda se hace en el directorio del
// ámbito donde se enlazaría, la primera ejecución predice copia y la deja,
// --dry-run dice lo mismo que la orden, y la segunda sale «sin cambios» con el
// plan vacío en vez de retirar la copia para volver a fallar al enlazar
// (FR-045, FR-048). Sondear en el temporal habría predicho enlace.
func casosConEnlacesSoloFueraDelAmbito(t *testing.T) []casoDePlan {
	t.Helper()

	deHome := ambitoGlobal(t, homeDePrueba)
	enElTrabajo := admiteFueraDe(directorioDeTrabajo)
	nuevo := manifiestoNuevo(ambitoLocal, ".agents", ".agents/skills")

	return []casoDePlan{
		{
			nombre:   "local con .claude",
			preparar: func(d *discoEnMemoria) { d.directorio(".claude") },
			admite:   enElTrabajo,
			skills:   lasDos(ambitoLocal, instalacion.EstadoInstalada, instalacion.ModoCopia),
			operaciones: slices.Concat([]string{"2 crear .claude/skills"}, nuevo,
				escribirEnteras(t, lasDosConCopia(ambitoLocal)...)),
			sondas: []string{".claude"},
		},
		{
			nombre:     "local con --host claude y sin .claude",
			preparar:   func(*discoEnMemoria) {},
			invocacion: conHost,
			admite:     enElTrabajo,
			skills:     lasDos(ambitoLocal, instalacion.EstadoInstalada, instalacion.ModoCopia),
			operaciones: slices.Concat([]string{"2 crear .claude", "2 crear .claude/skills"}, nuevo,
				escribirEnteras(t, lasDosConCopia(ambitoLocal)...)),
			sondas: []string{"."},
		},
		{
			nombre:     "-g con --host claude y sin .claude",
			preparar:   func(d *discoEnMemoria) { d.directorio(homeDePrueba) },
			invocacion: instalacion.Invocacion{Global: true, Host: texto("claude")},
			admite:     admiteFueraDe(homeDePrueba),
			skills:     lasDos(deHome, instalacion.EstadoInstalada, instalacion.ModoCopia),
			operaciones: slices.Concat([]string{"2 crear " + homeDePrueba + "/.claude", "2 crear " + homeDePrueba + "/.claude/skills"},
				manifiestoNuevo(deHome, homeDePrueba+"/.agents", homeDePrueba+"/.agents/skills"),
				escribirEnteras(t, lasDosConCopia(deHome)...)),
			sondas: []string{homeDePrueba},
		},
		{
			nombre: "-g con las copias ya hechas",
			preparar: func(d *discoEnMemoria) {
				instalarEn(d, deHome, "boe-legislacion", "legal-core").copiar("boe-legislacion").copiar("legal-core").escribir()
			},
			invocacion: global,
			admite:     admiteFueraDe(homeDePrueba),
			skills:     lasDos(deHome, instalacion.EstadoSinCambios, instalacion.ModoCopia),
			sondas:     []string{homeDePrueba + "/.claude/skills"},
		},
	}
}

// probarRecursoDeCopia fija el recurso de FR-024 en la aplicación: la sonda
// dice que se puede enlazar y el enlace falla igualmente, así que cada entrada
// pasa a su copia, con los mismos ficheros que el directorio neutro, y el
// manifiesto final la declara copia con sus huellas (SC-012). Después, con un
// creador de enlaces que ya dice que no, la copia se mantiene sin cambios.
func probarRecursoDeCopia(t *testing.T) {
	t.Parallel()

	d := nuevoDiscoEnMemoria(t)
	d.directorio(".claude")

	pedido, err := instalacion.ValidarInvocacion(instalacion.Invocacion{}, "", empotradasDePrueba())
	require.NoError(t, err)

	enlazador := &enlazadorEnMemoria{disco: d, sondea: admiteSiempre, enlaza: admiteNunca}

	plan, err := instalacion.Planificar(d, enlazador, pedido, empotradasDePrueba(), versionDePrueba)
	require.NoError(t, err)
	assert.Equal(t, lasDos(ambitoLocal, instalacion.EstadoInstalada, instalacion.ModoEnlace), plan.Skills,
		"la predicción, que sale de la sonda")

	require.Len(t, plan.Enlazar.Enlaces, 2)

	for _, enlace := range plan.Enlazar.Enlaces {
		exigirRecursoDeCopia(t, enlace)
	}

	previsto, err := plan.Manifiesto.Contenido(nil)
	require.NoError(t, err)

	conOtra, err := plan.Manifiesto.Contenido([]string{"otra-skill"})
	require.NoError(t, err)
	assert.Equal(t, previsto, conOtra, "una skill que no es la de ningún enlace del plan no cambia nada")

	enCopia := aplicarEnMemoria(t, plan, &escritorEnMemoria{disco: d, enlazador: enlazador})
	require.Equal(t, []string{"boe-legislacion", "legal-core"}, enCopia)

	manifiesto := leerManifiestoDelDisco(t, d)
	for _, nombre := range enCopia {
		host := manifiesto.Skills[nombre].Claude
		require.NotNil(t, host)
		assert.Equal(t, instalacion.ModoCopia, host.Modo)
		assert.Equal(t, huellasEmpotradas(t, nombre, ".claude/skills/"+nombre+"/"), host.Ficheros)
	}

	enCopias := lasDos(ambitoLocal, instalacion.EstadoInstalada, instalacion.ModoCopia)
	exigirHostsEnElDisco(t, d, enCopias)
	exigirSegundaSinCambios(t, d, pedido, admiteNunca, enCopias)
}

// exigirRecursoDeCopia exige que el enlace sea el de FR-021 y que su recurso
// de copia sea el directorio de la entrada con cada fichero empotrado, byte a
// byte.
func exigirRecursoDeCopia(t *testing.T, enlace instalacion.EnlaceNuevo) {
	t.Helper()

	assert.Equal(t, ".claude/skills/"+enlace.Skill, enlace.Ruta)
	assert.Equal(t, "../../.agents/skills/"+enlace.Skill, enlace.Destino)

	esperado := instalacion.Escrituras{DirectoriosQueFaltan: []string{enlace.Ruta, enlace.Ruta + "/references"}}
	for _, fichero := range empotradaDePrueba(t, enlace.Skill).Ficheros {
		esperado.Ficheros = append(esperado.Ficheros,
			instalacion.Escritura{Ruta: path.Join(enlace.Ruta, fichero.Ruta), Contenido: fichero.Contenido})
	}

	assert.Equal(t, esperado, enlace.Copia)
}

// probarPlanConConflictos exige que, con un conflicto, Planificar devuelva el
// mismo rechazo que ComprobarConflictos, ningún plan y ninguna pregunta a
// Disponible: no hay ninguna entrada que crear.
func probarPlanConConflictos(t *testing.T) {
	t.Parallel()

	d := nuevoDiscoEnMemoria(t)
	d.directorio(".claude")
	d.fichero(".agents/skills/legal-core/mio.md", "no es de kitlegal")

	pedido, err := instalacion.ValidarInvocacion(instalacion.Invocacion{}, "", empotradasDePrueba())
	require.NoError(t, err)

	enlazador := nuevoEnlazadorHonesto(d, admiteSiempre)

	plan, err := instalacion.Planificar(d, enlazador, pedido, empotradasDePrueba(), versionDePrueba)
	exigirConflictos(t, err, []instalacion.Conflicto{conflicto(carpetaAjena, ".agents/skills/legal-core")})
	assert.Equal(t, instalacion.Plan{}, plan)
	assert.Empty(t, enlazador.preguntados)
}

// probarFallosAlPlanificar exige que un fallo del Disco o de la sonda al
// decidir el plan se devuelva tal cual, y no como un conflicto, y que una
// versión del binario que el manifiesto no admitiría se rechace antes de
// examinar nada.
func probarFallosAlPlanificar(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre     string
		preparar   func(d *discoEnMemoria)
		invocacion instalacion.Invocacion
		sonda      error
	}{
		{
			nombre:   "la sonda de una entrada de host que falta",
			preparar: func(d *discoEnMemoria) { d.directorio(".claude") },
			sonda:    errInyectado,
		},
		{
			nombre:     "HOME, con -g",
			preparar:   func(d *discoEnMemoria) { d.fallar(opExaminar, homeDePrueba, errInyectado) },
			invocacion: instalacion.Invocacion{Global: true, Host: texto("claude")},
		},
		{
			nombre:     "lo que hay por encima de un HOME que no existe",
			preparar:   func(d *discoEnMemoria) { d.fallar(opExaminar, "/home", errInyectado) },
			invocacion: instalacion.Invocacion{Global: true, Host: texto("claude")},
		},
		{
			nombre:     "lo que hay por encima de --dir",
			preparar:   func(d *discoEnMemoria) { d.fallar(opExaminar, "otro", errInyectado) },
			invocacion: instalacion.Invocacion{Dir: texto("otro/destino")},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			d := nuevoDiscoEnMemoria(t)
			caso.preparar(d)

			pedido, err := instalacion.ValidarInvocacion(caso.invocacion, homeDePrueba, empotradasDePrueba())
			require.NoError(t, err)

			enlazador := nuevoEnlazadorHonesto(d, admiteSiempre)
			enlazador.err = caso.sonda

			_, err = instalacion.Planificar(d, enlazador, pedido, empotradasDePrueba(), versionDePrueba)
			exigirFalloSinConflictos(t, err, errInyectado)
		})
	}

	for _, version := range []string{"", "v0.1.0\n"} {
		t.Run("versión del binario "+strconv.Quote(version), func(t *testing.T) {
			t.Parallel()

			d := nuevoDiscoEnMemoria(t)

			_, err := instalacion.Planificar(d, nuevoEnlazadorHonesto(d, admiteSiempre), pedidoLocalConHost(t),
				empotradasDePrueba(), version)
			require.Error(t, err)
			assert.Empty(t, d.accesos, "no se examina nada")
		})
	}
}

// probarSalidaDeInstall fija la salida de install en JSON, la de
// contracts/applet-skills.md §4.1, con la biblioteca con la que la escribe el
// kernel: sus claves en español y una lista vacía como [], nunca null. Y que
// los enumerados de las etiquetas jsonschema, de las que --describe saca el
// esquema, son los valores del paquete.
func probarSalidaDeInstall(t *testing.T) {
	t.Parallel()

	salida, err := json.Marshal([]instalacion.SkillInstalada{
		salidaEn(ambitoLocal, "boe-legislacion", instalacion.EstadoInstalada, instalacion.ModoEnlace),
		salidaEn(ambitoLocal, "legal-core", instalacion.EstadoSinCambios),
	})
	require.NoError(t, err)

	assert.JSONEq(t, `[
		{"nombre": "boe-legislacion", "ruta": ".agents/skills/boe-legislacion", "estado": "instalada",
		 "enlaces": [{"host": "claude", "ruta": ".claude/skills/boe-legislacion", "modo": "enlace"}]},
		{"nombre": "legal-core", "ruta": ".agents/skills/legal-core", "estado": "sin cambios", "enlaces": []}
	]`, string(salida))

	enumerado := func(valores ...string) string {
		return "enum=" + strings.Join(valores, ",enum=")
	}

	etiquetas := map[string]string{
		"Estado": enumerado(string(instalacion.EstadoInstalada), string(instalacion.EstadoActualizada),
			string(instalacion.EstadoSinCambios)),
		"Host": enumerado("claude"),
		"Modo": enumerado(string(instalacion.ModoEnlace), string(instalacion.ModoCopia)),
	}

	for campo, esperada := range etiquetas {
		tipo := reflect.TypeFor[instalacion.Enlace]()
		if campo == "Estado" {
			tipo = reflect.TypeFor[instalacion.SkillInstalada]()
		}

		declarado, hay := tipo.FieldByName(campo)
		require.True(t, hay, campo)
		assert.Equal(t, esperada, declarado.Tag.Get("jsonschema"), campo)
	}
}

// leerDelDisco son los bytes del fichero regular de ruta.
func leerDelDisco(t *testing.T, d *discoEnMemoria, ruta string) []byte {
	t.Helper()

	contenido, err := d.Leer(ruta)
	require.NoError(t, err)

	return contenido
}

// leerManifiestoDelDisco es el manifiesto del ámbito local, que tiene que
// existir y ser legible.
func leerManifiestoDelDisco(t *testing.T, d *discoEnMemoria) instalacion.Manifiesto {
	t.Helper()

	manifiesto, err := instalacion.LeerManifiesto(leerDelDisco(t, d, ambitoLocal.RutaDelManifiesto()))
	require.NoError(t, err)

	return manifiesto
}

// huellasEmpotradas son los ficheros empotrados de la skill nombre, cada uno
// con prefijo delante de su ruta, con su huella: lo que declara el manifiesto
// de una instalación o de una copia al día.
func huellasEmpotradas(t *testing.T, nombre, prefijo string) map[string]string {
	t.Helper()

	huellas := map[string]string{}
	for _, fichero := range empotradaDePrueba(t, nombre).Ficheros {
		huellas[prefijo+fichero.Ruta] = fichero.Huella
	}

	return huellas
}

// exigirEntradasConservadas exige que la entrada de cada skill nombrada tenga
// en el manifiesto despues exactamente los bytes que tenía en antes (FR-034,
// FR-036), y que la versión de nivel superior sea ya la del binario que lo
// escribió el último (contracts/manifiesto.md §2).
func exigirEntradasConservadas(t *testing.T, antes, despues []byte, nombres ...string) {
	t.Helper()

	type entradas struct {
		Skills  map[string]json.RawMessage `json:"skills"`
		Version string                     `json:"version"`
	}

	var viejas, nuevas entradas

	require.NoError(t, json.Unmarshal(antes, &viejas))
	require.NoError(t, json.Unmarshal(despues, &nuevas))

	for _, nombre := range nombres {
		require.Contains(t, viejas.Skills, nombre)
		assert.Equal(t, string(viejas.Skills[nombre]), string(nuevas.Skills[nombre]), "la entrada de %s, byte a byte", nombre)
	}

	assert.Equal(t, versionDePrueba, nuevas.Version)
}
