package workflow

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	guionDelWrapper   = "scripts/claude-modelo.sh"
	guionDeLosModelos = "scripts/workflow/modelos.sh"
	guionDelInforme   = "scripts/workflow/informe.sh"
	registroDelHito   = gatesDelHito + "/modelos.jsonl"

	// claudeSustituto es el Claude Code que encuentra el wrapper en el test
	// (KITLEGAL_CLAUDE_BIN). No abre ninguna sesión: apunta sus argumentos,
	// escribe el transcript de la sesión que le fija --session-id con una
	// respuesta por cada modelo de KITLEGAL_TEST_MODELOS (vacío: ningún
	// transcript), escribe en sus dos salidas y sale con KITLEGAL_TEST_SALIDA.
	claudeSustituto = `#!/bin/sh
printf '%s\n' "$@" > "$KITLEGAL_TEST_ARGUMENTOS"
sesion=""; anterior=""
for a in "$@"; do [ "$anterior" = --session-id ] && sesion="$a"; anterior="$a"; done
if [ -n "${KITLEGAL_TEST_MODELOS:-}" ] && [ -n "$sesion" ]; then
  mkdir -p "$CLAUDE_CONFIG_DIR/projects/arbol"
  f="$CLAUDE_CONFIG_DIR/projects/arbol/$sesion.jsonl"
  printf '%s\n' '{"type":"user","message":{"role":"user","content":"el prompt"}}' > "$f"
  for m in $KITLEGAL_TEST_MODELOS; do
    printf '{"type":"assistant","sessionId":"%s","message":{"model":"%s","content":[]}}\n' "$sesion" "$m" >> "$f"
  done
fi
printf 'la respuesta de la sesión\n'
printf 'lo que la sesión escribe en su salida de error\n' >&2
exit "${KITLEGAL_TEST_SALIDA:-0}"
`
	salidaDelSustituto  = "la respuesta de la sesión\n"
	erroresDelSustituto = "lo que la sesión escribe en su salida de error\n"

	modeloDelJuezA = "claude-opus-5-5"
	modeloDelJuezB = "claude-fable-5-1"
	modeloDeSonnet = "claude-sonnet-5-5"
)

// uuid es la forma del id que el wrapper fija a la sesión.
var uuid = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

// sesionRegistrada es una línea de gates/modelos.jsonl.
type sesionRegistrada struct {
	Sesion        *string        `json:"sesion"`
	Paso          *string        `json:"paso"`
	Rol           *string        `json:"rol"`
	Pedido        string         `json:"pedido"`
	Esfuerzo      *string        `json:"esfuerzo"`
	Origen        string         `json:"origen"`
	AMano         bool           `json:"a_mano"`
	Repuesto      *string        `json:"repuesto"`
	Inicio        string         `json:"inicio"`
	Real          map[string]int `json:"real"`
	SinModeloReal *string        `json:"sin_modelo_real"`
}

// sesionDelTest es una sesión que el test abre por el wrapper: el modelo que
// pide, con su sufijo de paso y rol, las respuestas que escribe en su
// transcript y las variables con que la abre quien la lanza.
type sesionDelTest struct {
	modelo    string
	modelos   string
	salida    string
	variables []string
}

// abrirSesion lanza el wrapper con el Claude Code sustituto y devuelve lo que
// deja y los argumentos que recibió el sustituto.
func (a *arbol) abrirSesion(sesion sesionDelTest) (salidaDeOrden, []string) {
	a.t.Helper()

	sustituto := filepath.Join(a.fuera, "claude")

	fuera, err := os.OpenRoot(a.fuera)
	require.NoError(a.t, err)

	defer func() { require.NoError(a.t, fuera.Close()) }()

	if _, err := fuera.Stat("claude"); err != nil {
		escribirEjecutable(a.t, fuera, "claude", claudeSustituto)
	}

	argumentos := filepath.Join(a.fuera, "argumentos")
	variables := append([]string{
		"KITLEGAL_CLAUDE_BIN=" + sustituto,
		"KITLEGAL_TEST_ARGUMENTOS=" + argumentos,
		"KITLEGAL_TEST_MODELOS=" + sesion.modelos,
		"KITLEGAL_TEST_SALIDA=" + sesion.salida,
	}, sesion.variables...)

	salida := a.ejecutar(variables, guionDelWrapper, "-p", "el prompt", "--model", sesion.modelo)

	recibidos, err := fuera.ReadFile("argumentos")
	require.NoError(a.t, err)

	return salida, strings.Split(strings.TrimSuffix(string(recibidos), "\n"), "\n")
}

// registradas lee gates/modelos.jsonl.
func (a *arbol) registradas() []sesionRegistrada {
	a.t.Helper()

	var sesiones []sesionRegistrada

	for linea := range strings.Lines(a.leer(registroDelHito)) {
		var sesion sesionRegistrada
		require.NoError(a.t, json.Unmarshal([]byte(linea), &sesion))

		sesiones = append(sesiones, sesion)
	}

	return sesiones
}

// valorDe devuelve el argumento que sigue a una bandera, o vacío.
func valorDe(argumentos []string, bandera string) string {
	for i, argumento := range argumentos[:max(len(argumentos)-1, 0)] {
		if argumento == bandera {
			return argumentos[i+1]
		}
	}

	return ""
}

// casoDelWrapper es una sesión abierta por el wrapper y lo que tiene que quedar
// de ella en el registro.
type casoDelWrapper struct {
	nombre  string
	modelos string
	salida  string
	codigo  int
	real    map[string]int
	// sinModelo es lo que dice el registro cuando no se sabe el modelo real.
	sinModelo string
}

// TestElWrapperRegistraLaSesion: scripts/claude-modelo.sh fija un id a la
// sesión y la apunta en gates/modelos.jsonl con su paso, su rol y el modelo y
// el esfuerzo pedidos, sin el sufijo #<paso>:<rol>, que no llega a Claude
// Code; `modelos.sh resolver` saca el modelo real del transcript de esa
// sesión, sin contar los mensajes que Claude Code escribe sin modelo, y si no
// puede saberlo lo dice. La sesión sale con su código y su salida, también
// cuando no es 0.
func TestElWrapperRegistraLaSesion(t *testing.T) {
	t.Parallel()

	casos := []casoDelWrapper{
		{
			nombre: "con el modelo en el transcript", modelos: modeloDelJuezA + " <synthetic> " + modeloDelJuezA,
			real: map[string]int{modeloDelJuezA: 2},
		},
		{
			nombre: "con dos modelos en el transcript", modelos: modeloDelJuezA + " " + modeloDeSonnet,
			real: map[string]int{modeloDelJuezA: 1, modeloDeSonnet: 1},
		},
		{
			nombre: "la sesión sale con un código distinto de 0", modelos: modeloDeSonnet, salida: "7", codigo: 7,
			real: map[string]int{modeloDeSonnet: 1},
		},
		{
			nombre: "sin transcript", sinModelo: "no está el transcript de la sesión",
		},
		{
			nombre: "solo con mensajes sin modelo", modelos: "<synthetic>", salida: "1", codigo: 1,
			sinModelo: "la sesión no trae ninguna respuesta del modelo, solo 1 mensajes de Claude Code sin modelo",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			probarElWrapper(t, caso)
		})
	}
}

func probarElWrapper(t *testing.T, caso casoDelWrapper) {
	t.Helper()

	a := nuevoArbol(t)

	salida, argumentos := a.abrirSesion(sesionDelTest{
		modelo: modeloDelJuezA + "@xhigh#revision_juez_a:modelo_revisor", modelos: caso.modelos, salida: caso.salida,
	})
	assert.Equal(t, caso.codigo, salida.codigo, "el wrapper sale con el código de la sesión")
	assert.Equal(t, salidaDelSustituto, salida.estandar, "la salida estándar es la de la sesión, sin nada más")
	assert.Equal(t, erroresDelSustituto, salida.errores, "la salida de error es la de la sesión, sin nada más")

	assert.Equal(t, modeloDelJuezA, valorDe(argumentos, "--model"), "Claude Code recibe el modelo sin esfuerzo ni sufijo")
	assert.Equal(t, "xhigh", valorDe(argumentos, "--effort"))

	sesion := valorDe(argumentos, "--session-id")
	assert.Regexp(t, uuid, sesion, "el wrapper fija a la sesión un id con forma de UUID")

	for _, argumento := range argumentos {
		assert.NotContains(t, argumento, "#revision_juez_a", "el sufijo de paso y rol no llega a Claude Code")
	}

	a.debeSalirBien(nil, guionDeLosModelos, "resolver")

	registradas := a.registradas()
	require.Len(t, registradas, 1)

	registrada := registradas[0]
	require.NotNil(t, registrada.Sesion)
	assert.Equal(t, sesion, *registrada.Sesion)
	assert.Equal(t, "revision_juez_a", deref(registrada.Paso))
	assert.Equal(t, "modelo_revisor", deref(registrada.Rol))
	assert.Equal(t, modeloDelJuezA, registrada.Pedido)
	assert.Equal(t, "xhigh", deref(registrada.Esfuerzo))
	assert.Equal(t, "workflow", registrada.Origen)
	assert.False(t, registrada.AMano)
	assert.Nil(t, registrada.Repuesto)
	assert.Equal(t, caso.real, registrada.Real, "el modelo real es el del transcript, sin los mensajes sin modelo")

	if caso.sinModelo == "" {
		assert.Nil(t, registrada.SinModeloReal)
	} else {
		assert.Contains(t, deref(registrada.SinModeloReal), caso.sinModelo, "sin modelo real, el registro dice por qué")
	}
}

func deref(valor *string) string {
	if valor == nil {
		return ""
	}

	return *valor
}

// TestElRegistroNoCambiaLaSesion: si el wrapper no puede escribir el registro
// —aquí, sin .specify/feature.json—, la sesión se abre igual, con su código de
// salida y su salida estándar, y el wrapper solo lo dice en la de error.
func TestElRegistroNoCambiaLaSesion(t *testing.T) {
	t.Parallel()

	a := nuevoArbol(t)
	require.NoError(t, a.raiz.Remove(".specify/feature.json"))

	salida, argumentos := a.abrirSesion(sesionDelTest{
		modelo: modeloDelJuezA + "@xhigh#juez_spec:modelo_juez", modelos: modeloDelJuezA, salida: "3",
	})
	assert.Equal(t, 3, salida.codigo, "el wrapper sale con el código de la sesión")
	assert.Equal(t, salidaDelSustituto, salida.estandar, "la salida estándar es la de la sesión")
	assert.Contains(t, salida.errores, "claude-modelo: no se ha podido registrar la sesión")
	assert.True(t, strings.HasSuffix(salida.errores, erroresDelSustituto), "y después, la salida de error de la sesión")
	assert.Equal(t, modeloDelJuezA, valorDe(argumentos, "--model"))
	assert.True(t, a.noExiste(registroDelHito))
}

// TestUnTranscriptQueYaNoEstaNoDeshaceLoResuelto: Claude Code borra los
// transcripts con el tiempo; lo que se resolvió cuando estaba se conserva.
func TestUnTranscriptQueYaNoEstaNoDeshaceLoResuelto(t *testing.T) {
	t.Parallel()

	a := nuevoArbol(t)
	a.abrirSesion(sesionDelTest{modelo: modeloDelJuezA + "#converge:modelo_juez", modelos: modeloDelJuezA})
	a.debeSalirBien(nil, guionDeLosModelos, "resolver")
	require.NoError(t, a.raiz.RemoveAll(".config-claude/projects"))
	a.debeSalirBien(nil, guionDeLosModelos, "resolver")

	assert.Equal(t, map[string]int{modeloDelJuezA: 1}, a.registradas()[0].Real)
}

// rondaConSusJueces abre las sesiones de los dos jueces de una ronda, escribe
// sus veredictos y la lee con `gate.sh leer revision`.
func rondaConSusJueces(a *arbol, juezA, juezB sesionDelTest) {
	a.t.Helper()

	if juezA.modelo != "" {
		a.abrirSesion(juezA)
	}

	if juezB.modelo != "" {
		a.abrirSesion(juezB)
	}

	a.escribirJSON(gatesDelHito+"/revision-a.json", veredicto(true, "juez A"))
	a.escribirJSON(gatesDelHito+"/revision-b.json", veredicto(true, "juez B"))
	a.debeSalirBien(nil, guionDelGate, "leer", "revision", "4")
}

func juezA(modelos string, variables ...string) sesionDelTest {
	return sesionDelTest{modelo: modeloDelJuezA + "@xhigh#revision_juez_a:modelo_revisor", modelos: modelos, variables: variables}
}

func juezB(pedido, modelos string, variables ...string) sesionDelTest {
	return sesionDelTest{modelo: pedido + "@xhigh#revision_juez_b:modelo_adversario", modelos: modelos, variables: variables}
}

// TestRevisionJuzgadoLlevaElModeloDeCadaJuez: `gate.sh leer revision` pone en
// cada ronda de revision-juzgado.json el modelo real de cada juez, el de la
// última sesión de su paso registrada después de la lectura anterior; si
// ninguna lo es —un veredicto que no escribió una sesión del workflow ni de
// scripts/paso.sh, como en H26—, lo dice.
func TestRevisionJuzgadoLlevaElModeloDeCadaJuez(t *testing.T) {
	t.Parallel()

	a := nuevoArbol(t)
	rondaConSusJueces(a, juezA(strings.Repeat(modeloDelJuezA+" ", 3)), juezB(modeloDelJuezB, modeloDeSonnet+" "+modeloDeSonnet))
	rondaConSusJueces(a, sesionDelTest{}, sesionDelTest{})

	var rondas juzgado
	a.leerJSON(gatesDelHito+"/revision-juzgado.json", &rondas)
	require.Len(t, rondas.Rondas, 2)

	primera := rondas.Rondas[0]
	assert.Equal(t, map[string]int{modeloDelJuezA: 3}, primera.Modelos["a"].Real, "el juez A de la ronda 1")
	assert.Equal(t, map[string]int{modeloDeSonnet: 2}, primera.Modelos["b"].Real, "el juez B de la ronda 1 corrió con otro modelo que el pedido")
	assert.Equal(t, modeloDelJuezB, deref(primera.Modelos["b"].Pedido))
	assert.NotEmpty(t, primera.Leido)

	segunda := rondas.Rondas[1]
	for _, asiento := range []string{"a", "b"} {
		assert.Nil(t, segunda.Modelos[asiento].Real, "en la ronda 2 ninguna sesión juzgó el asiento %s", asiento)
		assert.Contains(t, deref(segunda.Modelos[asiento].SinModeloReal), "ninguna sesión de revision_juez_"+asiento+" registrada desde")
	}
}

// seccionUnoDelInforme ejecuta el informe final con --solo-ver y devuelve su
// sección 1.
func seccionUnoDelInforme(a *arbol) string {
	a.t.Helper()

	informe := a.debeSalirBien(nil, guionDelInforme, hitoDelTest, "--solo-ver").estandar
	_, despues, encontrada := strings.Cut(informe, "## 1. Estado\n")
	require.True(a.t, encontrada, "el informe tiene sección 1: %s", informe)

	seccion, _, _ := strings.Cut(despues, "\n## 2.")

	return seccion
}

// casoDelInforme son los jueces de una ronda y lo que la sección 1 del informe
// final dice de ellos.
type casoDelInforme struct {
	nombre       string
	juezA, juezB sesionDelTest
	contiene     []string
	noContiene   []string
}

// TestElInformeFinalNombraLasSesionesConOtroModelo: la sección 1 dice con qué
// modelo juzgó cada asiento y, en una línea propia, cada sesión que no corrió
// con el modelo que workflow.yml da a su rol, o de la que no se sabe, con su
// causa; en un run sano, que no hay ninguna.
func TestElInformeFinalNombraLasSesionesConOtroModelo(t *testing.T) {
	t.Parallel()

	sinOtro := "- **Sesiones con otro modelo que el de su rol**: ninguna (2 sesiones registradas"

	casos := []casoDelInforme{
		{
			nombre: "un run sano",
			juezA:  juezA(modeloDelJuezA), juezB: juezB(modeloDelJuezB, modeloDelJuezB),
			contiene: []string{
				"juez A aprobado (con `" + modeloDelJuezA + "`), juez B aprobado (con `" + modeloDelJuezB + "`)",
				sinOtro,
			},
		},
		{
			nombre: "el supervisor bajó de modelo por el límite de uso",
			juezA:  juezA(modeloDelJuezA),
			juezB: juezB(modeloDelJuezA, modeloDelJuezA,
				"KITLEGAL_MODELO_REPUESTO=fable → claude-opus-5-5 (límite de uso en revision_jueces, 2026-10-13T13:50Z)"),
			contiene: []string{
				"juez B aprobado (con `" + modeloDelJuezA + "`)",
				"- **Sesiones con otro modelo que el de su rol, o sin saber cuál: 1** de 2 (ADR 0033):",
				"  - `revision_juez_b` (",
				"corrió con `" + modeloDelJuezA + "` (1 respuestas); `workflow.yml` da `" + modeloDelJuezB + "` a `modelo_adversario`. " +
					"El supervisor había cambiado de modelo por el límite de uso (fable → claude-opus-5-5",
			},
			noContiene: []string{sinOtro, "`revision_juez_a` ("},
		},
		{
			nombre: "se pidió a mano con paso.sh",
			juezA:  juezA(modeloDelJuezA),
			juezB:  juezB(modeloDeSonnet, modeloDeSonnet, "KITLEGAL_SESION_ORIGEN=paso.sh", "KITLEGAL_MODELO_A_MANO=1"),
			contiene: []string{
				"corrió con `" + modeloDeSonnet + "` (1 respuestas); `workflow.yml` da `" + modeloDelJuezB + "` a `modelo_adversario`. " +
					"Se pidió a mano con scripts/paso.sh (" + modeloDeSonnet + ").",
			},
			noContiene: []string{sinOtro},
		},
		{
			nombre: "no se sabe el modelo de una sesión",
			juezA:  juezA(""), juezB: juezB(modeloDelJuezB, modeloDelJuezB),
			contiene: []string{
				"juez A aprobado (modelo desconocido: no está el transcript de la sesión",
				"  - `revision_juez_a` (",
				"modelo real desconocido; `workflow.yml` da `" + modeloDelJuezA + "` a `modelo_revisor`. No se sabe con qué modelo corrió",
			},
			noContiene: []string{sinOtro},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			probarElInforme(t, caso)
		})
	}
}

func probarElInforme(t *testing.T, caso casoDelInforme) {
	t.Helper()

	a := nuevoArbol(t)
	rondaConSusJueces(a, caso.juezA, caso.juezB)

	registro := a.leer(registroDelHito)
	seccion := seccionUnoDelInforme(a)

	for _, texto := range caso.contiene {
		assert.Contains(t, seccion, texto)
	}

	for _, texto := range caso.noContiene {
		assert.NotContains(t, seccion, texto)
	}

	assert.Equal(t, registro, a.leer(registroDelHito), "con --solo-ver, el informe no toca el registro")
}

// TestElInformeFinalDeUnRejuicio: la sección 1 cuenta el rejuicio del hito
// entero con su motivo y los veredictos que sustituye, que no desaparecen; y
// la sección 7 da por vistos por él todos los commits hasta su cabeza, también
// uno que antes no había visto ningún juez.
func TestElInformeFinalDeUnRejuicio(t *testing.T) {
	t.Parallel()

	a := nuevoArbol(t)
	cicloCerradoComoEnH26(a)

	a.escribir(ficheroDelProducto, "un cambio que ningún juez vio\n")
	a.commit("docs: un commit a mano después del último veredicto")
	sinVer := a.informeCompleto()
	assert.Contains(t, sinVer, "un commit a mano después del último veredicto: **ningún juez lo vio**")

	a.debeSalirBien(nil, guionDeLaRevision, "rejuzgar", hitoDelTest, motivoDelRejuicio)
	assert.Contains(t, seccionUnoDelInforme(a), "- **Rejuicio del hito entero abierto** (ciclo 2), sin veredicto todavía: "+motivoDelRejuicio+".")

	a.debeSalirBien(nil, guionDeLaRevision, "cambios", hitoDelTest)
	rondaConSusJueces(a, juezA(modeloDelJuezA), juezB(modeloDelJuezB, modeloDelJuezB))

	informe := a.informeCompleto()
	assert.Contains(t, informe, "- **Rejuicio del hito entero** (ciclo 2, desde la ronda 2): "+motivoDelRejuicio+
		". Sustituye a los veredictos anteriores: `revision-a-sustituido-c2.json` (aprobado), `revision-pendiente-sustituido-c2.md`, "+
		"los de las rondas 1 a 1 (`revision-{a,b}-r<n>.json`).")
	assert.Contains(t, informe, "un commit a mano después del último veredicto: lo vio la ronda 2 (juez A: aprobado; juez B: aprobado)")
	assert.Contains(t, informe, "- **Cambios que ningún juez vio**: ninguno.")
}

// TestElInformeFinalDiceQueElRejuicioVioUnCommit: un commit que vio una ronda
// anterior lo vuelve a ver el rejuicio del hito entero, y la sección 7 lo dice.
func TestElInformeFinalDiceQueElRejuicioVioUnCommit(t *testing.T) {
	t.Parallel()

	a := nuevoArbol(t)
	primera := a.cabeza()

	a.escribir(ficheroDelProducto, "lo que corrigió el corrector\n")
	a.commit("fix(H0): motivos de la revisión final")

	segunda := a.cabeza()

	a.escribirJSON(gatesDelHito+"/revision-a-r2.json", veredicto(true, "juez A de Sonnet"))
	a.escribirJSON(gatesDelHito+"/revision-a.json", veredicto(true, "juez A de Sonnet"))
	a.escribirJSON(gatesDelHito+"/revision-b-r2.json", veredicto(true, "juez B de Sonnet"))
	a.escribirJSON(gatesDelHito+"/revision-b.json", veredicto(true, "juez B de Sonnet"))
	a.escribirJSON(gatesDelHito+"/revision-juzgado.json", juzgado{Rondas: []rondaJuzgada{
		{Ronda: 1, Ciclo: 1, Sha: primera, JuezA: "aprobado", JuezB: "rechazado"},
		{Ronda: 2, Ciclo: 1, Sha: segunda, JuezA: "aprobado", JuezB: "aprobado"},
	}})
	a.escribirJSON(gatesDelHito+"/revision-ciclo.json", map[string]any{"ciclo": 1, "desde_ronda": 1, "cerrado": true})
	a.escribir(gatesDelHito+"/revision-rondas", "2\n")
	a.commit("docs(H0): veredictos de la revisión final")

	a.debeSalirBien(nil, guionDeLaRevision, "rejuzgar", hitoDelTest, motivoDelRejuicio)
	a.debeSalirBien(nil, guionDeLaRevision, "cambios", hitoDelTest)
	rondaConSusJueces(a, juezA(modeloDelJuezA), juezB(modeloDelJuezB, modeloDelJuezB))

	assert.Contains(t, a.informeCompleto(), "fix(H0): motivos de la revisión final: lo vio la ronda 2 (juez A: aprobado; juez B: aprobado); "+
		"y el rejuicio del hito entero en la ronda 3 (juez A: aprobado; juez B: aprobado).")
}

// informeCompleto ejecuta el informe final con --solo-ver.
func (a *arbol) informeCompleto() string {
	a.t.Helper()

	return a.debeSalirBien(nil, guionDelInforme, hitoDelTest, "--solo-ver").estandar
}

// TestElInformeFinalSoloVerNoTocaElRegistro: con --solo-ver, el informe
// resuelve el modelo real de las sesiones en una copia y dice cuál fue, sin
// escribir nada en el registro.
func TestElInformeFinalSoloVerNoTocaElRegistro(t *testing.T) {
	t.Parallel()

	a := nuevoArbol(t)
	a.abrirSesion(sesionDelTest{modelo: modeloDeSonnet + "@high#analyze:modelo_analisis", modelos: modeloDelJuezA})

	registro := a.leer(registroDelHito)
	require.Contains(t, registro, `"sin_modelo_real":"sin resolver"`)

	assert.Contains(t, seccionUnoDelInforme(a), "  - `analyze` (")
	assert.Contains(t, seccionUnoDelInforme(a), "corrió con `"+modeloDelJuezA+"` (1 respuestas); `workflow.yml` da `"+modeloDeSonnet+"` a `modelo_analisis`.")
	assert.Equal(t, registro, a.leer(registroDelHito), "el registro sigue sin resolver")
}
