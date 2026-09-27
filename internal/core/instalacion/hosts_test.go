package instalacion_test

import (
	"slices"
	"testing"

	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
)

// Las rutas de antigravity en el HOME de prueba (ADR 0025).
const (
	geminiDePrueba         = homeDePrueba + "/.gemini"
	configDeGeminiDePrueba = geminiDePrueba + "/config"
	skillsDeGeminiDePrueba = configDeGeminiDePrueba + "/skills"
)

// Invocaciones de install con -g que repiten las filas de antigravity.
var (
	globalConAntigravity = instalacion.Invocacion{Global: true, Hosts: []string{"antigravity"}}
	localConAntigravity  = instalacion.Invocacion{Hosts: []string{"antigravity"}}
)

// TestHostAntigravity fija el host antigravity (ADR 0025), con los mismos
// arneses que TestConflictos y TestPlan: solo lo tiene el ámbito global, donde
// lee .gemini/config/skills y no ~/.agents/skills; en local lee el directorio
// neutro y no tiene entradas. Se enlaza sin --host si .gemini y .gemini/config
// son, cada uno, un directorio real, y con --host antigravity aunque falten; cada
// directorio de la cadena tiene que ser un directorio real o no existir, y su
// enlace sube tres niveles hasta el directorio neutro.
func TestHostAntigravity(t *testing.T) {
	t.Parallel()

	t.Run("conflicto a conflicto", func(t *testing.T) {
		t.Parallel()

		probarCasosDeConflictos(t, casosDeConflictosDeAntigravity())
	})

	t.Run("plan", func(t *testing.T) {
		t.Parallel()

		probarCasosDePlan(t, casosDelPlanDeAntigravity(t))
	})
}

// casosDeConflictosDeAntigravity son las filas de data-model §4.1 y §4.3 en
// la cadena de antigravity y en sus entradas, y el manifiesto local que
// declara una entrada suya.
func casosDeConflictosDeAntigravity() []casoDeConflictos {
	return []casoDeConflictos{
		{
			nombre: ".gemini es un enlace a un directorio, con --host antigravity",
			preparar: func(d *discoEnMemoria) {
				d.directorio("/dotfiles/gemini/config/skills")
				d.enlace(geminiDePrueba, "/dotfiles/gemini")
			},
			invocacion:   globalConAntigravity,
			esperados:    []instalacion.Conflicto{conflicto(noEsDirectorio, geminiDePrueba)},
			noExaminadas: []string{"/dotfiles"},
		},
		{
			nombre: ".gemini es un enlace, sin --host: cuenta como ausente",
			preparar: func(d *discoEnMemoria) {
				d.directorio("/dotfiles/gemini/config/skills")
				d.enlace(geminiDePrueba, "/dotfiles/gemini")
			},
			invocacion:   global,
			noExaminadas: []string{"/dotfiles"},
		},
		{
			nombre:     ".gemini/config es un fichero, con --host antigravity",
			preparar:   func(d *discoEnMemoria) { d.fichero(configDeGeminiDePrueba, "no soy un directorio") },
			invocacion: globalConAntigravity,
			esperados:  []instalacion.Conflicto{conflicto(noEsDirectorio, configDeGeminiDePrueba)},
		},
		{
			nombre:     ".gemini/config/skills es un fichero, también sin --host",
			preparar:   func(d *discoEnMemoria) { d.fichero(skillsDeGeminiDePrueba, "no soy un directorio") },
			invocacion: global,
			esperados:  []instalacion.Conflicto{conflicto(noEsDirectorio, skillsDeGeminiDePrueba)},
		},
		{
			nombre:     "carpeta ajena con el nombre de una skill",
			preparar:   func(d *discoEnMemoria) { d.fichero(skillsDeGeminiDePrueba+"/legal-core/mio.md", "mío") },
			invocacion: global,
			esperados:  []instalacion.Conflicto{conflicto(carpetaAjena, skillsDeGeminiDePrueba+"/legal-core")},
		},
		{
			// El destino del enlace de claude sube dos niveles: desde
			// .gemini/config/skills llega a .gemini/.agents/skills, que no
			// existe, aunque la skill esté instalada en el directorio neutro.
			nombre: "el enlace con el destino de claude cuelga",
			preparar: func(d *discoEnMemoria) {
				instalarEn(d, ambitoGlobal(d.t, homeDePrueba), "legal-core").escribir()
				d.enlace(skillsDeGeminiDePrueba+"/legal-core", "../../.agents/skills/legal-core")
			},
			invocacion: global,
			esperados:  []instalacion.Conflicto{conflicto(enlaceRoto, skillsDeGeminiDePrueba+"/legal-core")},
		},
		{
			nombre: "un manifiesto local que declara una entrada de antigravity",
			preparar: func(d *discoEnMemoria) {
				i := instalarLocal(d, "legal-core")
				i.manifiesto.Skills["legal-core"] = conEntrada(i.manifiesto.Skills["legal-core"], "antigravity",
					instalacion.EntradaDeHost{Ruta: ".gemini/config/skills/legal-core", Modo: instalacion.ModoEnlace})
				i.escribir()
			},
			esperados: []instalacion.Conflicto{conflicto(conEntradasDeHost, ".agents/skills/kitlegal.json")},
		},
	}
}

// casosDelPlanDeAntigravity son los de un ámbito global con antigravity y el
// local con --host antigravity.
func casosDelPlanDeAntigravity(t *testing.T) []casoDePlan {
	t.Helper()

	deHome := ambitoGlobal(t, homeDePrueba)
	nuevo := manifiestoNuevo(deHome, homeDePrueba+"/.agents", homeDePrueba+"/.agents/skills")
	escribirLasDos := escribirEnteras(t, lasDosEn(deHome)...)
	antigravity := []string{"antigravity"}

	return []casoDePlan{
		{
			nombre:     "-g con .gemini/config: un enlace por skill que sube tres niveles",
			preparar:   func(d *discoEnMemoria) { d.directorio(configDeGeminiDePrueba) },
			invocacion: global,
			skills:     lasDosEnHosts(deHome, instalacion.EstadoInstalada, instalacion.ModoEnlace, antigravity...),
			operaciones: slices.Concat(enlazarLasDosEnHosts(deHome, antigravity, skillsDeGeminiDePrueba),
				nuevo, escribirLasDos),
			sondas: []string{configDeGeminiDePrueba},
		},
		{
			nombre: "-g con .claude y .gemini/config: los dos hosts, claude primero",
			preparar: func(d *discoEnMemoria) {
				d.directorio(homeDePrueba + "/.claude")
				d.directorio(configDeGeminiDePrueba)
			},
			invocacion: global,
			skills: lasDosEnHosts(deHome, instalacion.EstadoInstalada, instalacion.ModoEnlace,
				"claude", "antigravity"),
			operaciones: slices.Concat(
				enlazarLasDosEnHosts(deHome, []string{"claude", "antigravity"},
					homeDePrueba+"/.claude/skills", skillsDeGeminiDePrueba),
				nuevo, escribirLasDos),
			sondas: []string{homeDePrueba + "/.claude", configDeGeminiDePrueba},
		},
		{
			nombre:       "-g con .gemini y sin .gemini/config, el de Gemini CLI: no se enlaza",
			preparar:     func(d *discoEnMemoria) { d.directorio(geminiDePrueba) },
			invocacion:   global,
			skills:       lasDos(deHome, instalacion.EstadoInstalada),
			operaciones:  slices.Concat(nuevo, escribirLasDos),
			noExaminadas: []string{skillsDeGeminiDePrueba},
		},
		{
			nombre:     "-g --host antigravity sin .gemini: se crea la cadena entera, la sonda en HOME",
			preparar:   func(d *discoEnMemoria) { d.directorio(homeDePrueba) },
			invocacion: globalConAntigravity,
			skills:     lasDosEnHosts(deHome, instalacion.EstadoInstalada, instalacion.ModoEnlace, antigravity...),
			operaciones: slices.Concat(
				enlazarLasDosEnHosts(deHome, antigravity, geminiDePrueba, configDeGeminiDePrueba, skillsDeGeminiDePrueba),
				nuevo, escribirLasDos),
			sondas: []string{homeDePrueba},
		},
		{
			nombre:     "-g --host antigravity con .gemini: la sonda en .gemini",
			preparar:   func(d *discoEnMemoria) { d.directorio(geminiDePrueba) },
			invocacion: globalConAntigravity,
			skills:     lasDosEnHosts(deHome, instalacion.EstadoInstalada, instalacion.ModoEnlace, antigravity...),
			operaciones: slices.Concat(
				enlazarLasDosEnHosts(deHome, antigravity, configDeGeminiDePrueba, skillsDeGeminiDePrueba),
				nuevo, escribirLasDos),
			sondas: []string{geminiDePrueba},
		},
		{
			nombre:     "-g con .gemini/config y un creador de enlaces que no funciona, copias (FR-024)",
			preparar:   func(d *discoEnMemoria) { d.directorio(configDeGeminiDePrueba) },
			invocacion: global,
			admite:     admiteNunca,
			skills:     lasDosEnHosts(deHome, instalacion.EstadoInstalada, instalacion.ModoCopia, antigravity...),
			operaciones: slices.Concat([]string{"2 crear " + skillsDeGeminiDePrueba}, nuevo,
				escribirEnteras(t,
					deHome.RutaDeSkill("boe-legislacion"), deHome.RutaDeHost("antigravity", "boe-legislacion"),
					deHome.RutaDeSkill("legal-core"), deHome.RutaDeHost("antigravity", "legal-core"))),
			sondas: []string{configDeGeminiDePrueba},
		},
		{
			nombre: "-g con los enlaces de antigravity ya declarados: sin cambios",
			preparar: func(d *discoEnMemoria) {
				instalarEn(d, deHome, "boe-legislacion", "legal-core").
					enlazarEn("antigravity", "boe-legislacion").enlazarEn("antigravity", "legal-core").escribir()
			},
			invocacion: global,
			skills:     lasDosEnHosts(deHome, instalacion.EstadoSinCambios, instalacion.ModoEnlace, antigravity...),
		},
		{
			nombre:       "--host antigravity en local: lee el directorio neutro y no hay nada que enlazar",
			preparar:     func(d *discoEnMemoria) { d.directorio(".gemini/config") },
			invocacion:   localConAntigravity,
			skills:       lasDos(ambitoLocal, instalacion.EstadoInstalada),
			operaciones:  slices.Concat(manifiestoNuevo(ambitoLocal, ".agents", ".agents/skills"), escribirEnteras(t, lasDosEn(ambitoLocal)...)),
			noExaminadas: []string{".gemini"},
		},
	}
}

// lasDosEnHosts es la salida de las dos skills empotradas en el ámbito, en
// orden de nombre, con el mismo estado y una entrada con modo en cada host, en
// el orden dado.
func lasDosEnHosts(
	ambito instalacion.Ambito, estado instalacion.Estado, modo instalacion.Modo, hosts ...string,
) []instalacion.SkillInstalada {
	skills := make([]instalacion.SkillInstalada, 0, 2)

	for _, nombre := range []string{"boe-legislacion", "legal-core"} {
		enlaces := []instalacion.Enlace{}
		for _, host := range hosts {
			enlaces = append(enlaces, instalacion.Enlace{Host: host, Ruta: ambito.RutaDeHost(host, nombre), Modo: modo})
		}

		skills = append(skills, instalacion.SkillInstalada{
			Nombre: nombre, Ruta: ambito.RutaDeSkill(nombre), Estado: estado, Enlaces: enlaces,
		})
	}

	return skills
}

// enlazarLasDosEnHosts son las operaciones de la fase 2: cada directorio que
// falta y, por cada skill empotrada y en ella por cada host, en su orden, el
// enlace de FR-021 con el destino de ese host.
func enlazarLasDosEnHosts(ambito instalacion.Ambito, hosts []string, faltan ...string) []string {
	operaciones := make([]string, 0, len(faltan)+2*len(hosts))
	for _, dir := range faltan {
		operaciones = append(operaciones, "2 crear "+dir)
	}

	for _, nombre := range []string{"boe-legislacion", "legal-core"} {
		for _, host := range hosts {
			operaciones = append(operaciones,
				"2 enlazar "+ambito.RutaDeHost(host, nombre)+" -> "+subidaDelHostDePrueba[host]+nombre)
		}
	}

	return operaciones
}

// casosDoctorDeAntigravity son los de doctor en el ámbito global con entradas
// de antigravity: su orden fuerza el host con --host antigravity, y con
// --host claude --host antigravity si la skill tiene entrada en los dos, de
// modo que reinstala aunque falte .gemini/config (ADR 0025).
func casosDoctorDeAntigravity() []casoDeDoctor {
	enAntigravity := func(d *discoEnMemoria) {
		instalarEn(d, ambitoGlobal(d.t, homeDePrueba), "boe-legislacion", "legal-core").
			enlazarEn("antigravity", "boe-legislacion").enlazarEn("antigravity", "legal-core").escribir()
	}
	enLosDos := func(d *discoEnMemoria) {
		instalarEn(d, ambitoGlobal(d.t, homeDePrueba), "boe-legislacion", "legal-core").
			enlazarEn("claude", "boe-legislacion").enlazarEn("antigravity", "boe-legislacion").
			enlazarEn("claude", "legal-core").enlazarEn("antigravity", "legal-core").escribir()
	}
	hostLcDeGemini := skillsDeGeminiDePrueba + "/legal-core"
	hostBoeDeGemini := skillsDeGeminiDePrueba + "/boe-legislacion"
	lc := instala("legal-core") + " -g --host antigravity"

	return []casoDeDoctor{
		{nombre: "-g, los enlaces de antigravity sin tocar", preparar: enAntigravity, invocacion: global},
		{
			nombre: "-g, falta un enlace de antigravity",
			preparar: func(d *discoEnMemoria) {
				enAntigravity(d)
				d.retirar(hostLcDeGemini)
			},
			invocacion: global,
			esperados:  []instalacion.Hallazgo{hallazgo(aOtroSitio, hostLcDeGemini, lc)},
		},
		{
			nombre: "-g con los dos hosts y sin .gemini/config: cada orden fuerza los dos",
			preparar: func(d *discoEnMemoria) {
				enLosDos(d)
				d.retirar(configDeGeminiDePrueba)
			},
			invocacion: global,
			esperados: []instalacion.Hallazgo{
				hallazgo(aOtroSitio, hostBoeDeGemini, instala("boe-legislacion")+" -g --host claude --host antigravity"),
				hallazgo(aOtroSitio, hostLcDeGemini, instala("legal-core")+" -g --host claude --host antigravity"),
			},
			noExaminadas: []string{skillsDeGeminiDePrueba},
		},
		{
			nombre: "-g, un enlace de antigravity con el destino de claude",
			preparar: func(d *discoEnMemoria) {
				enAntigravity(d)
				d.enlace(hostLcDeGemini, "../../.agents/skills/legal-core")
			},
			invocacion: global,
			esperados:  []instalacion.Hallazgo{hallazgo(aOtroSitio, hostLcDeGemini, rmY(hostLcDeGemini, lc))},
		},
		{
			nombre: "-g, una copia de antigravity donde ya se puede enlazar",
			preparar: func(d *discoEnMemoria) {
				instalarEn(d, ambitoGlobal(d.t, homeDePrueba), "legal-core").copiarEn("antigravity", "legal-core").escribir()
			},
			invocacion: global,
			esperados:  []instalacion.Hallazgo{hallazgo(esCopia, hostLcDeGemini, lc)},
			sondas:     []string{skillsDeGeminiDePrueba},
		},
	}
}
