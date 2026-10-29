package instalacion_test

import (
	"encoding/json"
	"path"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
)

// Las pruebas de la retirada al instalar (FR-025 a FR-030) llevan sus propias
// skills, con nombres que no son los de ninguna entrega: el binario empotra
// heredera y otra, y antigua y vieja son nombres retirados cuyas preguntas
// hereda heredera. El dominio no conoce ningún nombre: la lista le llega como
// argumento de Planificar.

// retiradasDePrueba es la lista de los nombres retirados de las pruebas, en un
// mapa nuevo en cada llamada.
func retiradasDePrueba() instalacion.Retiradas {
	return instalacion.Retiradas{"antigua": {"heredera"}, "vieja": {"heredera"}}
}

// empotradasDeLaRetirada son las skills que empotra el binario de las pruebas
// de la retirada, en una lista nueva en cada llamada.
func empotradasDeLaRetirada() []instalacion.SkillEmpotrada {
	return []instalacion.SkillEmpotrada{
		{Nombre: "heredera", Ficheros: []instalacion.FicheroEmpotrado{
			instalacion.NuevoFicheroEmpotrado("SKILL.md", []byte("# heredera\n")),
		}},
		{Nombre: "otra", Ficheros: []instalacion.FicheroEmpotrado{
			instalacion.NuevoFicheroEmpotrado("SKILL.md", []byte("# otra\n")),
		}},
	}
}

// ficherosDeAntes son los ficheros que un binario anterior dejó en una skill
// cuyo nombre se retiró después, por su ruta dentro de ella.
var ficherosDeAntes = map[string]string{
	"SKILL.md":             "# de un binario anterior\n",
	"references/normas.md": "# Normas de entonces\n",
}

// retiradaEnteraDe es el orden en que se retira entero base, el directorio de
// una skill con ficherosDeAntes o su copia de host: cada fichero en orden de
// ruta, después su directorio de referencias y él el último.
func retiradaEnteraDe(base string) []string {
	return []string{base + "/SKILL.md", base + "/references/normas.md", base + "/references", base}
}

// deAntes deja en el directorio neutro la skill nombre con ficherosDeAntes y la
// declara con sus huellas, de versionVieja y sin entradas de host: como la dejó
// install antes de que su nombre se retirara.
func (i *instalada) deAntes(nombre string) *instalada {
	ficheros := map[string]string{}

	for rel, contenido := range ficherosDeAntes {
		i.disco.fichero(path.Join(i.ambito.RutaDeSkill(nombre), rel), contenido)
		ficheros[nombre+"/"+rel] = instalacion.HuellaDe([]byte(contenido))
	}

	i.manifiesto.Skills[nombre] = instalacion.SkillDeclarada{Version: versionVieja, Ficheros: ficheros}

	return i
}

// copiaDeAntesEn deja la entrada de la skill nombre en el host como una copia
// con ficherosDeAntes y la declara en copia con sus huellas.
func (i *instalada) copiaDeAntesEn(host, nombre string) *instalada {
	ruta := skillsDelHostDePrueba[host] + "/" + nombre
	ficheros := map[string]string{}

	for rel, contenido := range ficherosDeAntes {
		i.disco.fichero(path.Join(i.ambito.RutaDeHost(host, nombre), rel), contenido)
		ficheros[ruta+"/"+rel] = instalacion.HuellaDe([]byte(contenido))
	}

	i.manifiesto.Skills[nombre] = conEntrada(i.manifiesto.Skills[nombre], host,
		instalacion.EntradaDeHost{Ruta: ruta, Modo: instalacion.ModoCopia, Ficheros: ficheros})

	return i
}

// localConClaude es una instalación en el ámbito local con el directorio de skills
// de claude, en el que install enlaza, todavía sin ninguna skill.
func localConClaude(d *discoEnMemoria) *instalada {
	d.directorio(".claude/skills")

	return instalarLocal(d)
}

// retiradaEn es la salida de la skill nombre del ámbito que se retira, con cada
// entrada de host que se quita.
func retiradaEn(ambito instalacion.Ambito, nombre string, enlaces ...instalacion.Enlace) instalacion.SkillInstalada {
	return instalacion.SkillInstalada{
		Nombre:  nombre,
		Ruta:    ambito.RutaDeSkill(nombre),
		Estado:  instalacion.EstadoRetirada,
		Enlaces: append([]instalacion.Enlace{}, enlaces...),
	}
}

// avisoDeVieja es el aviso de contracts/instalacion-retiradas.md §2 detrás de
// la ruta de una skill de nombre retirado que hereda heredera. Su verbo va en
// dos literales porque misspell, con su diccionario inglés, lee su raíz como
// una errata de «install».
const avisoDeVieja = " no es lo que insta" + "ló kitlegal: su protocolo puede activarse junto al de heredera, " +
	"que la sustituye; retirarla es cosa tuya."

// sinRetirarEn es la salida de la skill nombre del ámbito que no se retira, con
// su aviso.
func sinRetirarEn(ambito instalacion.Ambito, nombre string) instalacion.SkillInstalada {
	ruta := ambito.RutaDeSkill(nombre)

	return instalacion.SkillInstalada{
		Nombre:  nombre,
		Ruta:    ruta,
		Estado:  instalacion.EstadoSinRetirar,
		Enlaces: []instalacion.Enlace{},
		Aviso:   ruta + avisoDeVieja,
	}
}

// quitada es la entrada de host de la skill nombre en el host del ámbito, en
// ese modo, tal como la nombra la salida de la que se retira.
func quitada(ambito instalacion.Ambito, host, nombre string, modo instalacion.Modo) instalacion.Enlace {
	return instalacion.Enlace{Host: host, Ruta: ambito.RutaDeHost(host, nombre), Modo: modo}
}

// casoDeRetirada es una fila de TestRetiradas: un disco con skills que el
// manifiesto declara y el binario no empotra, una invocación de install y lo
// que install hace con ellas.
type casoDeRetirada struct {
	nombre string
	// invocacion es la de install; se valida con HOME=homeDePrueba.
	invocacion instalacion.Invocacion
	// preparar deja el disco del caso.
	preparar func(d *discoEnMemoria)
	// salida son las entradas de la salida de las skills que el binario no
	// empotra, en su orden; ninguna si no se nombra ninguna.
	salida []instalacion.SkillInstalada
	// retirar es lo que el plan retira, en orden; nada de las skills
	// empotradas, que se instalan de nuevo.
	retirar []string
	// quitadas son las skills cuya entrada sale del manifiesto, y conservadas,
	// las que no se tocan: su entrada sigue en él byte a byte y sus rutas en
	// el disco no cambian.
	quitadas, conservadas []string
}

// TestRetiradas fija lo que install hace, en cada ámbito, con cada skill que
// el manifiesto declara con un nombre retirado, la pida o no la invocación
// (FR-026 a FR-030; contracts/instalacion-retiradas.md §2):
//
//   - si su directorio tiene exactamente los ficheros que el manifiesto le
//     declara, cada uno con su huella, y cada entrada de host es la declarada,
//     la retira —sus ficheros, cada directorio suyo de abajo arriba, sus
//     entradas de host y su entrada del manifiesto— y sale «retirada» con las
//     entradas quitadas;
//   - si no, no toca nada de ella, conserva su entrada y sale «sin retirar»,
//     con su aviso, en esta ejecución y en las siguientes;
//   - si no queda ni su directorio ni ninguna entrada de host, quita su
//     entrada sin nombrarla;
//   - planificar, que es todo lo que hace --dry-run, no cambia nada (FR-028);
//   - y una skill declarada que no se empotra y no es un nombre retirado sigue
//     como hoy: no se retira, no se cambia y no se nombra (FR-030).
func TestRetiradas(t *testing.T) {
	t.Parallel()

	grupos := []struct {
		nombre string
		casos  []casoDeRetirada
	}{
		{nombre: "se retira", casos: casosQueSeRetiran(t)},
		{nombre: "no se retira", casos: casosQueNoSeRetiran()},
		{nombre: "sin nada en el disco", casos: casosSinNadaEnElDisco()},
	}

	for _, grupo := range grupos {
		t.Run(grupo.nombre, func(t *testing.T) {
			t.Parallel()

			for _, caso := range grupo.casos {
				t.Run(caso.nombre, func(t *testing.T) {
					t.Parallel()

					probarCasoDeRetirada(t, caso)
				})
			}
		})
	}

	t.Run("fallos del disco", probarFallosAlExaminarLaRetirada)
	t.Run("un borrado que falla", probarFalloAlRetirar)
	t.Run("la salida", probarSalidaDeLaRetirada)
}

// probarCasoDeRetirada planifica el caso sobre su disco, lo compara con lo
// esperado, lo aplica y planifica una segunda vez.
func probarCasoDeRetirada(t *testing.T, caso casoDeRetirada) {
	t.Helper()

	d := nuevoDiscoEnMemoria(t)
	d.directorio(temporal)
	caso.preparar(d)

	empotradas := empotradasDeLaRetirada()

	pedido, err := instalacion.ValidarInvocacion(caso.invocacion, homeDePrueba, empotradas)
	require.NoError(t, err, "la invocación del caso")

	ambito := pedido.Ambito
	conservadas := rutasDeLasSkills(ambito, caso.conservadas...)
	manifiesto := leerDelDisco(t, d, ambito.RutaDelManifiesto())
	enlazador := nuevoEnlazadorHonesto(d, admiteSiempre)
	antes := d.instantanea()

	plan, err := instalacion.Planificar(d, enlazador, pedido, empotradas, retiradasDePrueba(), versionDePrueba)
	require.NoError(t, err)

	assert.Equal(t, antes, d.instantanea(), "planificar, que es todo lo que hace --dry-run, no cambia nada (FR-028)")
	assert.Equal(t, caso.salida, noEmpotradas(plan.Skills), "la salida de las skills que no se empotran")
	assert.Equal(t, caso.retirar, plan.Retirar, "lo que se retira, en orden")
	exigirDiscoRespetado(t, d, raizVigilada(ambito))

	aplicadas, err := instalacion.Aplicar(plan, &escritorEnMemoria{disco: d, enlazador: enlazador})
	require.NoError(t, err)
	assert.Equal(t, plan.Skills, aplicadas, "aplicado, la salida es la prevista")

	for _, ruta := range caso.retirar {
		assert.Equal(t, instalacion.EntradaAusente, examinarEnMemoria(t, d, ruta).Tipo, "%s, retirada", ruta)
	}

	assert.Equal(t, bajo(antes, conservadas...), bajo(d.instantanea(), conservadas...),
		"lo que no se retira queda en el disco tal como estaba")

	final := leerManifiestoDe(t, d, ambito)
	for _, nombre := range caso.quitadas {
		assert.NotContains(t, final.Skills, nombre, "su entrada sale del manifiesto")
	}

	exigirEntradasConservadas(t, manifiesto, leerDelDisco(t, d, ambito.RutaDelManifiesto()), caso.conservadas...)

	plan, err = instalacion.Planificar(d, enlazador, pedido, empotradas, retiradasDePrueba(), versionDePrueba)
	require.NoError(t, err, "la segunda ejecución")
	assert.Equal(t, sinRetirar(caso.salida), noEmpotradas(plan.Skills),
		"la segunda ejecución avisa otra vez de lo que no se retiró y no nombra nada más")
	assert.Empty(t, plan.Retirar, "la segunda ejecución no retira nada")
}

// rutasDeLasSkills son, de cada skill nombrada, su directorio en el neutro del ámbito y
// su entrada en cada host del ámbito.
func rutasDeLasSkills(ambito instalacion.Ambito, nombres ...string) []string {
	var rutas []string

	for _, nombre := range nombres {
		rutas = append(rutas, ambito.RutaDeSkill(nombre))
		for _, host := range ambito.Hosts() {
			rutas = append(rutas, ambito.RutaDeHost(host, nombre))
		}
	}

	return rutas
}

// bajo son las entradas de instantanea que son una de rutas o cuelgan de ella.
func bajo(instantanea map[string]nodoEnMemoria, rutas ...string) map[string]nodoEnMemoria {
	dentro := map[string]nodoEnMemoria{}

	for clave, nodo := range instantanea {
		for _, ruta := range rutas {
			if abs := absoluta(ruta); clave == abs || strings.HasPrefix(clave, abs+"/") {
				dentro[clave] = nodo
			}
		}
	}

	return dentro
}

// noEmpotradas son las de skills que el binario de las pruebas de la retirada
// no empotra, en su orden; nil si no hay ninguna.
func noEmpotradas(skills []instalacion.SkillInstalada) []instalacion.SkillInstalada {
	var fuera []instalacion.SkillInstalada

	for _, skill := range skills {
		if !slices.ContainsFunc(empotradasDeLaRetirada(), func(empotrada instalacion.SkillEmpotrada) bool {
			return empotrada.Nombre == skill.Nombre
		}) {
			fuera = append(fuera, skill)
		}
	}

	return fuera
}

// sinRetirar son las de skills que salen «sin retirar», en su orden; nil si no
// hay ninguna.
func sinRetirar(skills []instalacion.SkillInstalada) []instalacion.SkillInstalada {
	var avisadas []instalacion.SkillInstalada

	for _, skill := range skills {
		if skill.Estado == instalacion.EstadoSinRetirar {
			avisadas = append(avisadas, skill)
		}
	}

	return avisadas
}

// leerManifiestoDe es el manifiesto del ámbito, que tiene que existir y ser
// legible.
func leerManifiestoDe(t *testing.T, d *discoEnMemoria, ambito instalacion.Ambito) instalacion.Manifiesto {
	t.Helper()

	manifiesto, err := instalacion.LeerManifiesto(leerDelDisco(t, d, ambito.RutaDelManifiesto()))
	require.NoError(t, err)

	return manifiesto
}

// casosQueSeRetiran son los de una skill de nombre retirado con las huellas
// que coinciden, en cada ámbito y con cada clase de entrada de host: se retira
// entera, aunque la invocación no pida la skill que la hereda (FR-026).
func casosQueSeRetiran(t *testing.T) []casoDeRetirada {
	t.Helper()

	local := ambitoLocal
	cuenta := ambitoGlobal(t, homeDePrueba)
	dir := instalacion.NuevoAmbitoDir("destino")

	return []casoDeRetirada{
		{
			nombre:   "con su enlace",
			preparar: func(d *discoEnMemoria) { localConClaude(d).deAntes("vieja").enlazar("vieja").escribir() },
			salida: []instalacion.SkillInstalada{
				retiradaEn(local, "vieja", quitada(local, "claude", "vieja", instalacion.ModoEnlace)),
			},
			retirar:  append(retiradaEnteraDe(".agents/skills/vieja"), ".claude/skills/vieja"),
			quitadas: []string{"vieja"},
		},
		{
			nombre: "con su copia de host",
			preparar: func(d *discoEnMemoria) {
				localConClaude(d).deAntes("vieja").copiaDeAntesEn("claude", "vieja").escribir()
			},
			salida: []instalacion.SkillInstalada{
				retiradaEn(local, "vieja", quitada(local, "claude", "vieja", instalacion.ModoCopia)),
			},
			retirar:  slices.Concat(retiradaEnteraDe(".agents/skills/vieja"), retiradaEnteraDe(".claude/skills/vieja")),
			quitadas: []string{"vieja"},
		},
		{
			nombre:   "sin entradas de host",
			preparar: func(d *discoEnMemoria) { localConClaude(d).deAntes("vieja").escribir() },
			salida:   []instalacion.SkillInstalada{retiradaEn(local, "vieja")},
			retirar:  retiradaEnteraDe(".agents/skills/vieja"),
			quitadas: []string{"vieja"},
		},
		{
			nombre:     "con la invocación que pide otra skill que no es la que la hereda",
			invocacion: instalacion.Invocacion{Skills: []string{"otra"}},
			preparar:   func(d *discoEnMemoria) { localConClaude(d).deAntes("vieja").enlazar("vieja").escribir() },
			salida: []instalacion.SkillInstalada{
				retiradaEn(local, "vieja", quitada(local, "claude", "vieja", instalacion.ModoEnlace)),
			},
			retirar:  append(retiradaEnteraDe(".agents/skills/vieja"), ".claude/skills/vieja"),
			quitadas: []string{"vieja"},
		},
		{
			nombre:     "en -g, con sus entradas en los dos hosts",
			invocacion: global,
			preparar: func(d *discoEnMemoria) {
				instalarEn(d, cuenta).deAntes("vieja").enlazarEn("claude", "vieja").enlazarEn("antigravity", "vieja").escribir()
			},
			salida: []instalacion.SkillInstalada{retiradaEn(cuenta, "vieja",
				quitada(cuenta, "claude", "vieja", instalacion.ModoEnlace),
				quitada(cuenta, "antigravity", "vieja", instalacion.ModoEnlace))},
			retirar: append(retiradaEnteraDe(homeDePrueba+"/.agents/skills/vieja"),
				homeDePrueba+"/.claude/skills/vieja", homeDePrueba+"/.gemini/config/skills/vieja"),
			quitadas: []string{"vieja"},
		},
		{
			nombre:     "con --dir",
			invocacion: instalacion.Invocacion{Dir: texto("destino")},
			preparar: func(d *discoEnMemoria) {
				instalarEn(d, instalacion.NuevoAmbitoDir("destino")).deAntes("vieja").escribir()
			},
			salida:   []instalacion.SkillInstalada{retiradaEn(dir, "vieja")},
			retirar:  retiradaEnteraDe("destino/vieja"),
			quitadas: []string{"vieja"},
		},
	}
}

// casosQueNoSeRetiran son los de una skill de nombre retirado cuyas huellas no
// coinciden: un fichero editado, añadido o que falta, lo que no es un
// directorio real o una entrada de host que no es la declarada. Nada de ella
// cambia, su entrada se conserva y sale «sin retirar» con su aviso (FR-027).
// También, la skill no empotrada que no es un nombre retirado (FR-030).
func casosQueNoSeRetiran() []casoDeRetirada {
	local := ambitoLocal
	aviso := []instalacion.SkillInstalada{sinRetirarEn(local, "vieja")}

	// conservada es el caso de vieja, enlazada en claude y cambiada después
	// por preparar.
	conservada := func(nombre string, cambiar func(d *discoEnMemoria)) casoDeRetirada {
		return casoDeRetirada{
			nombre: nombre,
			preparar: func(d *discoEnMemoria) {
				localConClaude(d).deAntes("vieja").enlazar("vieja").escribir()
				cambiar(d)
			},
			salida:      aviso,
			conservadas: []string{"vieja"},
		}
	}

	return []casoDeRetirada{
		conservada("un fichero editado", func(d *discoEnMemoria) {
			d.fichero(".agents/skills/vieja/SKILL.md", "Una línea que la persona cambió.\n")
		}),
		conservada("un fichero añadido", func(d *discoEnMemoria) {
			d.fichero(".agents/skills/vieja/mio.md", "de la persona\n")
		}),
		conservada("un directorio añadido", func(d *discoEnMemoria) { d.directorio(".agents/skills/vieja/mio") }),
		conservada("un fichero que falta", func(d *discoEnMemoria) {
			d.retirar(".agents/skills/vieja/references/normas.md")
		}),
		conservada("un fichero declarado que es un enlace", func(d *discoEnMemoria) {
			d.enlace(".agents/skills/vieja/SKILL.md", "../../../SKILL.md")
		}),
		conservada("su directorio de referencias es un enlace", func(d *discoEnMemoria) {
			d.enlace(".agents/skills/vieja/references", "../../../references")
		}),
		conservada("su directorio es un enlace", func(d *discoEnMemoria) {
			d.enlace(".agents/skills/vieja", "../../otra-parte/vieja")
		}),
		conservada("su enlace de host falta", func(d *discoEnMemoria) { d.retirar(".claude/skills/vieja") }),
		conservada("su enlace de host va a otro sitio", func(d *discoEnMemoria) {
			d.enlace(".claude/skills/vieja", "../../otra-parte/vieja")
		}),
		conservada("su entrada de host es un directorio", func(d *discoEnMemoria) {
			d.retirar(".claude/skills/vieja")
			d.directorio(".claude/skills/vieja")
		}),
		conservada("sin su directorio, con su enlace de host", func(d *discoEnMemoria) {
			d.retirar(".agents/skills/vieja")
		}),
		{
			nombre: "una entrada de host que no declara",
			preparar: func(d *discoEnMemoria) {
				localConClaude(d).deAntes("vieja").escribir()
				d.enlace(".claude/skills/vieja", "../../.agents/skills/vieja")
			},
			salida:      aviso,
			conservadas: []string{"vieja"},
		},
		{
			nombre: "su copia de host editada",
			preparar: func(d *discoEnMemoria) {
				localConClaude(d).deAntes("vieja").copiaDeAntesEn("claude", "vieja").escribir()
				d.fichero(".claude/skills/vieja/references/normas.md", "Una línea que la persona cambió.\n")
			},
			salida:      aviso,
			conservadas: []string{"vieja"},
		},
		{
			nombre: "una que se retira y otra que no, en orden de nombre",
			preparar: func(d *discoEnMemoria) {
				localConClaude(d).deAntes("antigua").deAntes("vieja").escribir()
				d.fichero(".agents/skills/vieja/SKILL.md", "Una línea que la persona cambió.\n")
			},
			salida:      []instalacion.SkillInstalada{retiradaEn(local, "antigua"), sinRetirarEn(local, "vieja")},
			retirar:     retiradaEnteraDe(".agents/skills/antigua"),
			quitadas:    []string{"antigua"},
			conservadas: []string{"vieja"},
		},
		{
			nombre:      "una skill no empotrada que no es un nombre retirado (FR-030)",
			preparar:    func(d *discoEnMemoria) { localConClaude(d).noEmpotrada("ajena", versionVieja).escribir() },
			conservadas: []string{"ajena"},
		},
	}
}

// casosSinNadaEnElDisco son los de una skill de nombre retirado de la que no
// queda ni su directorio ni ninguna entrada de host: su entrada sale del
// manifiesto sin que se nombre (FR-027).
func casosSinNadaEnElDisco() []casoDeRetirada {
	return []casoDeRetirada{
		{
			nombre: "declarada con su enlace",
			preparar: func(d *discoEnMemoria) {
				localConClaude(d).deAntes("vieja").enlazar("vieja").escribir()
				d.retirar(".agents/skills/vieja")
				d.retirar(".claude/skills/vieja")
			},
			quitadas: []string{"vieja"},
		},
		{
			nombre: "declarada con su enlace, sin el directorio de skills de claude",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d).deAntes("vieja").enlazar("vieja").escribir()
				d.retirar(".agents/skills/vieja")
				d.retirar(".claude")
			},
			quitadas: []string{"vieja"},
		},
	}
}

// probarFallosAlExaminarLaRetirada exige que un fallo del Disco al examinar
// una skill de nombre retirado se devuelva tal cual, sin plan y sin nombrarse
// como conflicto: es el fallo de hoy, un defecto del entorno.
func probarFallosAlExaminarLaRetirada(t *testing.T) {
	t.Parallel()

	// En cada caso, vieja está en claude en copia o, con enlazada, enlazada.
	casos := []struct {
		nombre   string
		op       operacion
		ruta     string
		enlazada bool
	}{
		{nombre: "examinar su directorio", op: opExaminar, ruta: ".agents/skills/vieja"},
		{nombre: "listar su directorio", op: opNombres, ruta: ".agents/skills/vieja"},
		{nombre: "examinar uno de sus ficheros", op: opExaminar, ruta: ".agents/skills/vieja/SKILL.md"},
		{nombre: "la huella de uno de sus ficheros", op: opHuella, ruta: ".agents/skills/vieja/SKILL.md"},
		{nombre: "examinar su copia de host", op: opExaminar, ruta: ".claude/skills/vieja"},
		{nombre: "listar su copia de host", op: opNombres, ruta: ".claude/skills/vieja"},
		{nombre: "examinar su enlace de host", op: opExaminar, ruta: ".claude/skills/vieja", enlazada: true},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			d := nuevoDiscoEnMemoria(t)
			instalada := localConClaude(d).deAntes("vieja")

			if caso.enlazada {
				instalada.enlazar("vieja").escribir()
			} else {
				instalada.copiaDeAntesEn("claude", "vieja").escribir()
			}

			d.fallar(caso.op, caso.ruta, errInyectado)

			empotradas := empotradasDeLaRetirada()

			pedido, err := instalacion.ValidarInvocacion(instalacion.Invocacion{}, homeDePrueba, empotradas)
			require.NoError(t, err)

			plan, err := instalacion.Planificar(d, nuevoEnlazadorHonesto(d, admiteSiempre), pedido, empotradas,
				retiradasDePrueba(), versionDePrueba)
			exigirFalloSinConflictos(t, err, errInyectado)
			assert.Equal(t, instalacion.Plan{}, plan)
		})
	}
}

// probarFalloAlRetirar exige que un borrado que falla al retirar una skill de
// nombre retirado sea el fallo de hoy de Aplicar, que nombra la operación y la
// ruta, sin pedir nada más y sin salida.
func probarFalloAlRetirar(t *testing.T) {
	t.Parallel()

	d := nuevoDiscoEnMemoria(t)
	localConClaude(d).deAntes("vieja").enlazar("vieja").escribir()

	empotradas := empotradasDeLaRetirada()

	pedido, err := instalacion.ValidarInvocacion(instalacion.Invocacion{}, homeDePrueba, empotradas)
	require.NoError(t, err)

	enlazador := nuevoEnlazadorHonesto(d, admiteSiempre)

	plan, err := instalacion.Planificar(d, enlazador, pedido, empotradas, retiradasDePrueba(), versionDePrueba)
	require.NoError(t, err)

	escritor := &escritorEnMemoria{disco: d, enlazador: enlazador, fallarEn: 1}

	skills, err := instalacion.Aplicar(plan, escritor)
	require.ErrorIs(t, err, errInyectado)
	assert.Equal(t, "skills install: retirar .agents/skills/vieja/SKILL.md: "+errInyectado.Error(), err.Error())
	assert.Nil(t, skills)
	assert.Equal(t, []string{"retirar .agents/skills/vieja/SKILL.md"}, escritor.operaciones)
}

// probarSalidaDeLaRetirada fija la forma en JSON de las dos salidas de una
// skill de nombre retirado (contracts/instalacion-retiradas.md §3): la que se
// retira, con las entradas quitadas, y la que no, con su aviso, el único campo
// que se omite en las demás.
func probarSalidaDeLaRetirada(t *testing.T) {
	t.Parallel()

	salida, err := json.Marshal([]instalacion.SkillInstalada{
		retiradaEn(ambitoLocal, "antigua", quitada(ambitoLocal, "claude", "antigua", instalacion.ModoEnlace)),
		sinRetirarEn(ambitoLocal, "vieja"),
	})
	require.NoError(t, err)

	assert.JSONEq(t, `[
		{"nombre": "antigua", "ruta": ".agents/skills/antigua", "estado": "retirada",
		 "enlaces": [{"host": "claude", "ruta": ".claude/skills/antigua", "modo": "enlace"}]},
		{"nombre": "vieja", "ruta": ".agents/skills/vieja", "estado": "sin retirar", "enlaces": [],
		 "aviso": ".agents/skills/vieja`+avisoDeVieja+`"}
	]`, string(salida))
}

// TestHerederas fija cómo se nombran las herederas de un nombre retirado en el
// aviso y en la salida legible: una, dos con «y», más con comas y «y» delante
// de la última; nada si el nombre no es retirado.
func TestHerederas(t *testing.T) {
	t.Parallel()

	retiradas := instalacion.Retiradas{"una": {"a"}, "dos": {"a", "b"}, "tres": {"a", "b", "c"}}

	assert.Equal(t, "a", retiradas.Herederas("una"))
	assert.Equal(t, "a y b", retiradas.Herederas("dos"))
	assert.Equal(t, "a, b y c", retiradas.Herederas("tres"))
	assert.Empty(t, retiradas.Herederas("ninguna"))
}
