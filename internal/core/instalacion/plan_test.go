package instalacion_test

import (
	"path"
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
//     clase y en orden de ruta, en un error que declara la clase «inesperado»
//     y cuyo mensaje es la cabecera de contracts/applet-skills.md §5 y una
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
	assert.Equal(t, schema.ClaseInesperado, conClase.Clase())

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
