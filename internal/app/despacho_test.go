package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cli"
)

// registroDeDespacho es el registro con el que se ejercen todos los casos del
// despacho: «echo» —un verbo, «repetir», marcado por omisión, que es lo que hace
// válida la entrega literal del hito— y «segundo» —dos verbos y ninguno por
// omisión, la rama en que nombrar el verbo es obligatorio—. Son los dos applets
// de ejemplo de contracts/registro-y-describe.md §3, declarados aquí con los
// mismos ayudantes que usa el test del registro.
func registroDeDespacho(t *testing.T) *Registro {
	t.Helper()

	var registro Registro

	require.NoError(t, registro.Registrar(appletConVerbos("echo",
		verboDePrueba("repetir", true))))
	require.NoError(t, registro.Registrar(appletConVerbos("segundo",
		verboDePrueba("uno", false), verboDePrueba("dos", false))))

	return &registro
}

// casoDeDespacho es una invocación y lo que el despacho tiene que decidir sobre
// ella. Un caso con errorContiene es un caso que termina en código 2; los demás
// fijan el destino, el applet resuelto y los argumentos que recibirá la
// gramática, que es donde se ve si el verbo por omisión se insertó o no.
type casoDeDespacho struct {
	nombre        string
	argv          []string
	previo        cli.Preliminar
	destino       Destino
	applet        string
	reservado     string
	args          []string
	ayudaContiene []string
	errorContiene []string
}

// comprobar ejerce un caso completo contra el registro de los dos applets de
// ejemplo. Vive aquí, y no repetida en cada tabla, porque las dos tablas —la del
// despacho y la del verbo por omisión— miran exactamente lo mismo: qué atiende
// la invocación y con qué argumentos.
func (caso casoDeDespacho) comprobar(t *testing.T) {
	t.Helper()

	despacho, err := Despachar(registroDeDespacho(t), caso.argv, caso.previo)

	if len(caso.errorContiene) > 0 {
		require.ErrorIs(t, err, cli.ErrArgumentos,
			"lo que no se puede despachar es una invocación mal formada")
		assert.Equal(t, 2, cli.CodigoSalida(err),
			"un applet o un verbo que no se resuelven salen con código 2 (FR-006, FR-027)")

		for _, fragmento := range caso.errorContiene {
			require.ErrorContains(t, err, fragmento)
		}

		assert.Equal(t, Despacho{}, despacho, "un despacho fallido no resuelve nada")

		return
	}

	require.NoError(t, err)
	assert.Equal(t, caso.destino, despacho.Destino)
	assert.Equal(t, caso.args, despacho.Args)
	assert.Equal(t, caso.reservado, despacho.Reservado)

	if caso.applet == "" {
		assert.Nil(t, despacho.Applet, "esta invocación no resuelve ningún applet")
	} else {
		require.NotNil(t, despacho.Applet)
		assert.Equal(t, caso.applet, despacho.Applet.Nombre())
	}

	for _, fragmento := range caso.ayudaContiene {
		assert.Contains(t, despacho.Ayuda, fragmento)
	}
}

// TestDespacho es la tabla de los cinco casos de despacho de
// contracts/registro-y-describe.md §2, con las dos consecuencias que el contrato
// exige además: que un nombre de enlace desconocido no inutilice el binario y
// que el sufijo .exe no cuente (FR-002 … FR-006, D17).
func TestDespacho(t *testing.T) {
	t.Parallel()

	casos := []casoDeDespacho{
		{
			// Caso 1: el nombre de invocación manda y los argumentos se
			// entregan íntegros al applet, así que «boe» es el mensaje que
			// repite «echo» y no otro applet.
			nombre:  "el nombre de invocación está registrado y manda él",
			argv:    []string{"/usr/local/bin/echo", "boe"},
			destino: DestinoApplet,
			applet:  "echo",
			args:    []string{"repetir", "boe"},
		},
		{
			// Caso 2: el nombre propio del binario no está registrado, así que
			// el applet sale del primer argumento.
			nombre:  "el nombre del binario no está registrado y el applet sale del primer argumento",
			argv:    []string{"/usr/local/bin/kitlegal", "echo", "hola"},
			destino: DestinoApplet,
			applet:  "echo",
			args:    []string{"repetir", "hola"},
		},
		{
			// Caso 3: sin primer argumento no hay applet que deducir, y el
			// mensaje orienta con la lista en lugar de fallar en silencio.
			nombre:        "sin primer argumento y sin nombre reconocible",
			argv:          []string{"/usr/local/bin/kitlegal"},
			errorContiene: []string{"ningún applet", "echo", "segundo"},
		},
		{
			// Caso 4: el primer argumento no corresponde a ningún applet; el
			// mensaje lo nombra y enumera los que sí existen.
			nombre:        "el primer argumento no corresponde a ningún applet",
			argv:          []string{"/usr/local/bin/kitlegal", "noexiste", "hola"},
			errorContiene: []string{"noexiste", "echo", "segundo"},
		},
		{
			// Caso 5: los verbos reservados se reconocen antes que el registro,
			// que es lo que hace inalcanzable un applet llamado «version» y por
			// lo que el registro lo rechaza al construirse.
			nombre:    "el primer argumento es un verbo reservado del binario",
			argv:      []string{"/usr/local/bin/kitlegal", "version"},
			destino:   DestinoReservado,
			reservado: "version",
			args:      []string{},
		},
		{
			nombre:  "un nombre de enlace desconocido no inutiliza el binario",
			argv:    []string{"/tmp/kitlegal-dev", "echo", "hola"},
			destino: DestinoApplet,
			applet:  "echo",
			args:    []string{"repetir", "hola"},
		},
		{
			nombre:  "el sufijo .exe no cuenta en el nombre de invocación",
			argv:    []string{"/tmp/echo.exe", "hola"},
			destino: DestinoApplet,
			applet:  "echo",
			args:    []string{"repetir", "hola"},
		},
		{
			nombre:        "la ayuda del binario cuando no hay applet que resolver",
			argv:          []string{"/usr/local/bin/kitlegal", "--help"},
			previo:        cli.Preliminar{Ayuda: true},
			destino:       DestinoAyuda,
			args:          []string{"--help"},
			ayudaContiene: []string{"kitlegal", "echo", "segundo"},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			caso.comprobar(t)
		})
	}
}

// TestVerboPorOmision es la tabla de D26: qué verbo ejecuta una invocación que
// no lo nombra, con verbo por omisión, sin él, con un argumento que se llama
// como un verbo y con la ayuda del applet, que es el único caso en que la
// normalización se suprime (contracts/registro-y-describe.md §2 bis).
func TestVerboPorOmision(t *testing.T) {
	t.Parallel()

	casos := []casoDeDespacho{
		{
			nombre:  "con verbo por omisión, se inserta a la cabeza",
			argv:    []string{"kitlegal", "echo", "hola"},
			destino: DestinoApplet,
			applet:  "echo",
			args:    []string{"repetir", "hola"},
		},
		{
			nombre:  "con verbo por omisión y sin ningún argumento",
			argv:    []string{"kitlegal", "echo"},
			destino: DestinoApplet,
			applet:  "echo",
			args:    []string{"repetir"},
		},
		{
			nombre:  "las banderas escritas antes del verbo no lo estorban",
			argv:    []string{"kitlegal", "echo", "--json", "hola"},
			destino: DestinoApplet,
			applet:  "echo",
			args:    []string{"repetir", "--json", "hola"},
		},
		{
			// La ambigüedad, resuelta por escrito y en un solo sentido: si el
			// primer argumento se llama como un verbo, es el verbo.
			nombre:  "el argumento que se llama como un verbo es el verbo",
			argv:    []string{"kitlegal", "echo", "repetir"},
			destino: DestinoApplet,
			applet:  "echo",
			args:    []string{"repetir"},
		},
		{
			// Quien necesite lo contrario dispone del terminador: lo que va
			// después ya no nombra ningún verbo.
			nombre:  "tras el terminador no hay verbo que reconocer",
			argv:    []string{"kitlegal", "echo", "--", "repetir"},
			destino: DestinoApplet,
			applet:  "echo",
			args:    []string{"repetir", "--", "repetir"},
		},
		{
			nombre:  "el verbo nombrado de un applet sin verbo por omisión no se toca",
			argv:    []string{"kitlegal", "segundo", "uno", "algo"},
			destino: DestinoApplet,
			applet:  "segundo",
			args:    []string{"uno", "algo"},
		},
		{
			nombre:        "sin verbo por omisión, código 2 con la lista de verbos",
			argv:          []string{"kitlegal", "segundo", "hola"},
			errorContiene: []string{"segundo", "uno", "dos"},
		},
		{
			nombre:        "sin verbo por omisión y sin ningún argumento",
			argv:          []string{"kitlegal", "segundo"},
			errorContiene: []string{"segundo", "uno", "dos"},
		},
		{
			// Pedir la ayuda del applet no resuelve ningún verbo: la lista de
			// argumentos llega tal cual, sin nada insertado.
			nombre:        "con la ayuda del applet no se inserta nada",
			argv:          []string{"kitlegal", "echo", "--help"},
			previo:        cli.Preliminar{Ayuda: true},
			destino:       DestinoAyuda,
			applet:        "echo",
			args:          []string{"--help"},
			ayudaContiene: []string{"echo", "repetir"},
		},
		{
			nombre:        "pedir la ayuda de un applet sin verbo por omisión no es un fallo",
			argv:          []string{"kitlegal", "segundo", "--help"},
			previo:        cli.Preliminar{Ayuda: true},
			destino:       DestinoAyuda,
			applet:        "segundo",
			args:          []string{"--help"},
			ayudaContiene: []string{"segundo", "uno", "dos"},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			caso.comprobar(t)
		})
	}
}

// TestDespachoIndistinguible comprueba en el despacho la mitad que le toca de
// FR-007 y SC-003: invocar un applet por enlace simbólico y con el applet como
// primer argumento decide exactamente lo mismo —el mismo applet y los mismos
// argumentos—, que es lo que hace que la salida de las dos invocaciones pueda
// ser idéntica byte a byte.
func TestDespachoIndistinguible(t *testing.T) {
	t.Parallel()

	registro := registroDeDespacho(t)

	porEnlace, err := Despachar(registro, []string{"/tmp/echo", "hola", "--json"}, cli.Preliminar{})
	require.NoError(t, err)

	porArgumento, err := Despachar(registro,
		[]string{"/tmp/kitlegal", "echo", "hola", "--json"}, cli.Preliminar{})
	require.NoError(t, err)

	assert.Equal(t, porEnlace, porArgumento,
		"el enlace simbólico y el primer argumento despachan lo mismo")
}

// TestDespachoNoAltera comprueba que el despacho no escribe en la lista que
// recibe: la normalización del verbo construye una lista nueva. Quien invoca
// —la raíz de composición, con os.Args— conserva la suya intacta, y dos
// despachos sobre la misma lista deciden lo mismo.
func TestDespachoNoAltera(t *testing.T) {
	t.Parallel()

	argv := []string{"kitlegal", "echo", "hola"}

	primero, err := Despachar(registroDeDespacho(t), argv, cli.Preliminar{})
	require.NoError(t, err)

	assert.Equal(t, []string{"kitlegal", "echo", "hola"}, argv,
		"la lista de argumentos que recibe el despacho no se toca")

	segundo, err := Despachar(registroDeDespacho(t), argv, cli.Preliminar{})
	require.NoError(t, err)

	assert.Equal(t, primero.Args, segundo.Args)
}

// TestDespachoConRegistroVacio comprueba lo que hace en H1 el binario que se
// publica, cuyo registro no tiene ningún applet: «kitlegal echo hola» termina
// como cualquier otro nombre desconocido, y el mensaje dice que no hay ninguno
// en lugar de enumerar una lista vacía (FR-009,
// contracts/registro-y-describe.md §3).
func TestDespachoConRegistroVacio(t *testing.T) {
	t.Parallel()

	_, err := Despachar(RegistroDeProduccion(),
		[]string{"kitlegal", "echo", "hola"}, cli.Preliminar{})

	require.ErrorIs(t, err, cli.ErrArgumentos)
	assert.Equal(t, 2, cli.CodigoSalida(err))
	require.ErrorContains(t, err, "echo", "el mensaje nombra lo que no se ha reconocido")
	require.ErrorContains(t, err, "ningún applet")
}

// TestDespachoSinArgumentos comprueba que una lista de argumentos vacía —que no
// puede llegar del sistema operativo, pero sí de quien llame a Main— no produce
// ningún pánico: es un error de argumentos como cualquier otro (FR-033).
func TestDespachoSinArgumentos(t *testing.T) {
	t.Parallel()

	_, err := Despachar(registroDeDespacho(t), nil, cli.Preliminar{})

	require.ErrorIs(t, err, cli.ErrArgumentos)
	assert.Equal(t, 2, cli.CodigoSalida(err))
}
