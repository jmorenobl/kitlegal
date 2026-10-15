package workflow

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	guionDeLaRevision = "scripts/workflow/revision.sh"
	guionDelGate      = "scripts/workflow/gate.sh"

	// motivoDelRejuicio es el que la persona da a rejuzgar, el de H26.
	motivoDelRejuicio = "los dos asientos los juzgó Sonnet porque el supervisor bajó de modelo por el límite de uso"
)

// cicloDeLaRevision es gates/revision-ciclo.json.
type cicloDeLaRevision struct {
	Ciclo      int    `json:"ciclo"`
	DesdeRonda int    `json:"desde_ronda"`
	Motivo     string `json:"motivo"`
	Entero     bool   `json:"entero"`
	Desde      any    `json:"desde"`
	Cerrado    bool   `json:"cerrado"`
}

// rondaJuzgada es una ronda de gates/revision-juzgado.json.
type rondaJuzgada struct {
	Ronda   int                      `json:"ronda"`
	Ciclo   int                      `json:"ciclo"`
	Sha     string                   `json:"sha"`
	JuezA   string                   `json:"juez_a"`
	JuezB   string                   `json:"juez_b"`
	Leido   string                   `json:"leido,omitempty"`
	Entero  bool                     `json:"entero,omitempty"`
	Motivo  string                   `json:"motivo,omitempty"`
	Modelos map[string]modeloDelJuez `json:"modelos,omitempty"`
}

// modeloDelJuez es el modelo con el que juzgó un asiento en una ronda.
type modeloDelJuez struct {
	Sesion        *string        `json:"sesion"`
	Pedido        *string        `json:"pedido"`
	Real          map[string]int `json:"real"`
	SinModeloReal *string        `json:"sin_modelo_real"`
}

// juzgado es gates/revision-juzgado.json.
type juzgado struct {
	Rondas []rondaJuzgada `json:"rondas"`
}

// veredicto es el de un juez de la revisión final, con una evidencia que lo
// distingue de otro con el mismo resultado.
func veredicto(aprobado bool, evidencia string) map[string]any {
	resultado := "aprobado"
	if !aprobado {
		resultado = "rechazado"
	}

	motivos := []string{}
	if !aprobado {
		motivos = []string{"[a] " + evidencia}
	}

	return map[string]any{
		"veredicto":     resultado,
		"criterios":     []any{map[string]any{"id": "a", "criterio": "dod_ci", "cumple": aprobado, "evidencia": evidencia}},
		"motivos":       motivos,
		"observaciones": []string{},
	}
}

// cicloCerradoComoEnH26 deja el árbol como quedó H26 tras repetir a mano su
// revisión final: el ciclo 1 cerrado con una ronda que juzgaron dos jueces de
// Sonnet (revision-{a,b}-r1.json y revision-juzgado.json); en
// revision-a.json, un veredicto de otro modelo copiado a mano sin su copia
// -r<n>; en revision-b.json, el mismo de su copia -r1; y un
// revision-pendiente.md. Todo commiteado. Devuelve la cabeza juzgada.
func cicloCerradoComoEnH26(a *arbol) string {
	a.t.Helper()

	juzgada := a.cabeza()

	a.escribirJSON(gatesDelHito+"/revision-a-r1.json", veredicto(true, "juez A de Sonnet"))
	a.escribirJSON(gatesDelHito+"/revision-a.json", veredicto(true, "juez A de Opus, copiado a mano"))
	a.escribirJSON(gatesDelHito+"/revision-b-r1.json", veredicto(true, "juez B de Sonnet"))
	a.escribirJSON(gatesDelHito+"/revision-b.json", veredicto(true, "juez B de Sonnet"))
	a.escribir(gatesDelHito+"/revision-pendiente.md", "# Motivos de un ciclo anterior\n")
	a.escribirJSON(gatesDelHito+"/revision-juzgado.json", juzgado{Rondas: []rondaJuzgada{
		{Ronda: 1, Ciclo: 1, Sha: juzgada, JuezA: "aprobado", JuezB: "aprobado"},
	}})
	a.escribirJSON(gatesDelHito+"/revision-ciclo.json", map[string]any{
		"ciclo": 1, "desde_ronda": 1, "motivo": "revisión final del hito", "desde": nil, "commits": []any{}, "cerrado": true,
	})
	a.escribir(gatesDelHito+"/revision-rondas", "1\n")
	a.commit("docs(H0): veredictos de la revisión final")

	return juzgada
}

// TestRejuzgarElHitoEnteroComoEnH26 reproduce el caso de H26: con el ciclo de
// la revisión final cerrado y veredictos de Sonnet, `revision.sh rejuzgar`
// abre un ciclo que vuelve a juzgar el hito entero. Después, `cambios` dice
// que es un rejuicio, con su motivo, sin veredicto anterior en gates/ y con
// `git diff main...HEAD`; los veredictos apartados siguen archivados y
// `pendiente` continúa ese ciclo.
//
// Existe porque en H26 `cambios` daba el rango vacío desde el último veredicto
// y cada juez leía como suyo el veredicto de otro modelo: la revisión se
// repitió desde un clon.
func TestRejuzgarElHitoEnteroComoEnH26(t *testing.T) {
	t.Parallel()

	a := nuevoArbol(t)
	cicloCerradoComoEnH26(a)

	opus := a.leer(gatesDelHito + "/revision-a.json")
	pendiente := a.leer(gatesDelHito + "/revision-pendiente.md")

	salida := a.debeSalirBien(nil, guionDeLaRevision, "rejuzgar", hitoDelTest, motivoDelRejuicio)

	var abierto cicloDeLaRevision
	require.NoError(t, json.Unmarshal([]byte(salida.estandar), &abierto))
	assert.Equal(t, cicloDeLaRevision{Ciclo: 2, DesdeRonda: 2, Motivo: motivoDelRejuicio, Entero: true}, abierto,
		"rejuzgar dice el ciclo que abre, desde qué ronda y por qué")

	var ciclo cicloDeLaRevision
	a.leerJSON(gatesDelHito+"/revision-ciclo.json", &ciclo)
	assert.Equal(t, cicloDeLaRevision{Ciclo: 2, DesdeRonda: 2, Motivo: motivoDelRejuicio, Entero: true}, ciclo,
		"el ciclo nuevo es entero, abierto, con el motivo y sin cabeza desde la que juzgar")

	assert.True(t, a.noExiste(gatesDelHito+"/revision-a.json"), "el veredicto del juez A queda apartado")
	assert.True(t, a.noExiste(gatesDelHito+"/revision-b.json"), "el veredicto del juez B queda apartado")
	assert.True(t, a.noExiste(gatesDelHito+"/revision-pendiente.md"), "el pendiente del ciclo anterior queda apartado")
	assert.Equal(t, opus, a.leer(gatesDelHito+"/revision-a-sustituido-c2.json"),
		"el veredicto copiado a mano, sin copia -r<n>, se archiva como sustituido")
	assert.Equal(t, pendiente, a.leer(gatesDelHito+"/revision-pendiente-sustituido-c2.md"),
		"el pendiente se archiva como sustituido")
	assert.True(t, a.noExiste(gatesDelHito+"/revision-b-sustituido-c2.json"),
		"el veredicto que ya tiene una copia idéntica -r1 no se duplica")

	for _, copia := range []string{"revision-a-r1.json", "revision-b-r1.json"} {
		assert.True(t, a.existe(gatesDelHito+"/"+copia), "la copia %s de la ronda 1 sigue archivada", copia)
	}

	assert.Contains(t, a.leer(gatesDelHito+"/supuestos.md"), motivoDelRejuicio, "el rejuicio se anota como supuesto, con su motivo")

	a.debeSalirBien(nil, guionDeLaRevision, "cambios", hitoDelTest)

	cambios := a.leer(gatesDelHito + "/revision-cambios.md")
	assert.Contains(t, cambios, "# Qué juzga la ronda 2 de la revisión final")
	assert.Contains(t, cambios, "Rejuicio del hito entero")
	assert.Contains(t, cambios, motivoDelRejuicio)
	assert.Contains(t, cambios, "No hay ningún veredicto anterior tuyo")
	assert.Contains(t, cambios, "Juzgas el hito entero, `git diff main...HEAD`.")
	assert.NotContains(t, cambios, "Tu último veredicto", "la ronda no juzga un rango desde los veredictos apartados")

	var continua struct {
		Revisar  bool   `json:"revisar"`
		Continua bool   `json:"continua"`
		Ciclo    int    `json:"ciclo"`
		Motivo   string `json:"motivo"`
	}

	pendienteDelBucle := a.debeSalirBien(nil, guionDeLaRevision, "pendiente", hitoDelTest)
	require.NoError(t, json.Unmarshal([]byte(pendienteDelBucle.estandar), &continua))
	assert.True(t, continua.Revisar && continua.Continua, "pendiente continúa el ciclo abierto: %s", pendienteDelBucle.estandar)
	assert.Equal(t, 2, continua.Ciclo)
	assert.Equal(t, motivoDelRejuicio, continua.Motivo)
}

// casoDeRejuzgarQueSeNiega es un caso en que `revision.sh rejuzgar` no abre
// ningún ciclo.
type casoDeRejuzgarQueSeNiega struct {
	nombre string
	// preparar deja el árbol en el estado del caso.
	preparar func(a *arbol)
	motivo   string
	codigo   int
	mensaje  string
}

// TestRejuzgarSeNiega: con un ciclo abierto, que `pendiente` ya continúa; sin
// ningún veredicto, porque la primera revisión ya juzga el hito entero; y sin
// motivo, que va al informe final. En ninguno se aparta ningún veredicto.
func TestRejuzgarSeNiega(t *testing.T) {
	t.Parallel()

	casos := []casoDeRejuzgarQueSeNiega{
		{
			nombre: "con un ciclo abierto",
			preparar: func(a *arbol) {
				cicloCerradoComoEnH26(a)
				a.escribirJSON(gatesDelHito+"/revision-ciclo.json", map[string]any{
					"ciclo": 2, "desde_ronda": 2, "motivo": "cambios fuera de gates/", "desde": a.cabeza(), "commits": []any{}, "cerrado": false,
				})
			},
			motivo: motivoDelRejuicio, codigo: 1, mensaje: "sigue abierto",
		},
		{
			nombre: "sin ningún veredicto",
			preparar: func(a *arbol) {
				a.escribirJSON(gatesDelHito+"/revision-a.json", veredicto(true, "un veredicto sin ronda"))
			},
			motivo: motivoDelRejuicio, codigo: 1, mensaje: "no tiene ningún veredicto",
		},
		{
			nombre:   "sin motivo",
			preparar: func(a *arbol) { cicloCerradoComoEnH26(a) },
			codigo:   2, mensaje: "el motivo va al informe final",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			probarRejuzgarQueSeNiega(t, caso)
		})
	}
}

func probarRejuzgarQueSeNiega(t *testing.T, caso casoDeRejuzgarQueSeNiega) {
	t.Helper()

	a := nuevoArbol(t)
	caso.preparar(a)

	ciclo := ""
	if a.existe(gatesDelHito + "/revision-ciclo.json") {
		ciclo = a.leer(gatesDelHito + "/revision-ciclo.json")
	}

	salida := a.ejecutar(nil, guionDeLaRevision, "rejuzgar", hitoDelTest, caso.motivo)
	assert.Equal(t, caso.codigo, salida.codigo, "rejuzgar se niega: %s", salida.errores)
	assert.Contains(t, salida.errores, caso.mensaje)
	assert.True(t, a.existe(gatesDelHito+"/revision-a.json"), "no aparta ningún veredicto")
	assert.False(t, a.existe(gatesDelHito+"/revision-a-sustituido-c2.json"), "no archiva nada")

	if ciclo != "" {
		assert.Equal(t, ciclo, a.leer(gatesDelHito+"/revision-ciclo.json"), "no toca el ciclo")
	}
}

// TestLaSegundaRondaDeUnRejuicioVuelveAlRango: si en un ciclo entero el
// corrector cambia algo, la ronda siguiente del mismo ciclo juzga el rango
// desde la última ronda del ciclo, como cualquier otra, y no el hito entero.
// La ronda que lee `gate.sh leer revision` lleva que es de un rejuicio, con su
// motivo.
func TestLaSegundaRondaDeUnRejuicioVuelveAlRango(t *testing.T) {
	t.Parallel()

	a := nuevoArbol(t)
	cicloCerradoComoEnH26(a)
	a.debeSalirBien(nil, guionDeLaRevision, "rejuzgar", hitoDelTest, motivoDelRejuicio)
	a.debeSalirBien(nil, guionDeLaRevision, "cambios", hitoDelTest)

	a.escribirJSON(gatesDelHito+"/revision-a.json", veredicto(true, "juez A del rejuicio"))
	a.escribirJSON(gatesDelHito+"/revision-b.json", veredicto(false, "falta un test"))

	var leida struct {
		Estado string `json:"estado"`
	}

	require.NoError(t, json.Unmarshal([]byte(a.debeSalirBien(nil, guionDelGate, "leer", "revision", "4").estandar), &leida))
	assert.Equal(t, "corregir", leida.Estado)

	var rondas juzgado
	a.leerJSON(gatesDelHito+"/revision-juzgado.json", &rondas)
	require.Len(t, rondas.Rondas, 2)

	segunda := rondas.Rondas[1]
	assert.Equal(t, 2, segunda.Ronda)
	assert.Equal(t, 2, segunda.Ciclo)
	assert.True(t, segunda.Entero, "la ronda es de un rejuicio")
	assert.Equal(t, motivoDelRejuicio, segunda.Motivo)

	a.escribir(ficheroDelProducto, "lo que entrega el hito, con el test que faltaba\n")
	a.commit("fix(H0): motivos de la revisión final")
	a.debeSalirBien(nil, guionDeLaRevision, "cambios", hitoDelTest)

	cambios := a.leer(gatesDelHito + "/revision-cambios.md")
	assert.Contains(t, cambios, "# Qué juzga la ronda 3 de la revisión final")
	assert.Contains(t, cambios, "Tu último veredicto es el de la ronda 2, sobre `"+segunda.Sha[:7]+"`")
	assert.Contains(t, cambios, "fix(H0): motivos de la revisión final — paso `corrector_revision`")
	assert.NotContains(t, cambios, "Rejuicio del hito entero", "la segunda ronda del ciclo no vuelve a juzgar el hito entero")
}
