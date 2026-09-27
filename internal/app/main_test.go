package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// timeoutDelContrato es el plazo por omisión que promete
// contracts/banderas-y-exit-codes.md §1, escrito aquí y no leído del kernel:
// si el kernel lo cambiara, el contrato seguiría diciendo esto y el test lo
// notaría.
const timeoutDelContrato = 30 * time.Second

// appletDeContexto es el applet que **anota** lo que el kernel le entrega en
// cada ejecución, para poder comprobar desde fuera del applet dos cosas que
// ninguna otra tabla puede ver: si el applet llegó a ejecutarse y con qué
// contexto. Declara lo mismo que cualquier applet y nada más (SC-010).
type appletDeContexto struct {
	ejecuciones *[]schema.Contexto
}

func (a appletDeContexto) Nombre() string { return "contexto" }

func (a appletDeContexto) Descripcion() string { return "applet que anota su contexto de ejecución" }

func (a appletDeContexto) Verbos() []Verbo {
	return []Verbo{{
		Nombre:      "anotar",
		Descripcion: "anota el contexto con el que se ejecuta",
		Argumentos:  func() Argumentos { return &argumentosDeContexto{ejecuciones: a.ejecuciones} },
		Salida:      map[string]any{},
		PorOmision:  true,
	}}
}

// argumentosDeContexto son los argumentos del verbo: un mensaje opcional, y el
// cuaderno en un campo no exportado que la gramática no ve.
type argumentosDeContexto struct {
	Mensaje string `arg:"" optional:"" help:"Mensaje que se devuelve."`

	ejecuciones *[]schema.Contexto
}

func (a *argumentosDeContexto) Ejecutar(
	_ context.Context, ec schema.Contexto, _ *slog.Logger,
) (schema.Resultado, error) {
	*a.ejecuciones = append(*a.ejecuciones, ec)

	return schema.Resultado{
		Procedencia: schema.Procedencia{Fuente: "kitlegal.contexto", URL: "kitlegal:applet/contexto"},
		Datos:       map[string]any{"mensaje": a.Mensaje},
	}, nil
}

var (
	_ Applet     = appletDeContexto{}
	_ Argumentos = (*argumentosDeContexto)(nil)
)

// registroDeContexto construye el registro con el applet que anota, y devuelve
// además el cuaderno en el que anota.
func registroDeContexto(t *testing.T) (*Registro, *[]schema.Contexto) {
	t.Helper()

	ejecuciones := &[]schema.Contexto{}

	var registro Registro

	require.NoError(t, registro.Registrar(appletDeContexto{ejecuciones: ejecuciones}))

	return &registro, ejecuciones
}

// TestContextoDeEjecucion comprueba, desde el applet, lo que el kernel promete
// entregarle y cuándo: que las seis opciones globales llegan tal cual en el
// contexto de ejecución —también --offline, --no-graph y --asunto, que en H1 no
// tienen objeto pero sí tienen que llegar (FR-021, FR-023, FR-024)—, que
// --dry-run **no corta antes del applet** sino que viaja en el contexto
// (FR-022, US4.7) y que --describe y la ayuda no lo ejecutan (FR-049, US5.3).
//
// Es la comprobación que ninguna otra tabla hace: las demás miran los
// descriptores y el código, y esas dos cosas no cambiarían si el kernel cortara
// en --dry-run o entregara un contexto vacío.
func TestContextoDeEjecucion(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre      string
		argv        []string
		ejecuciones int
		contexto    schema.Contexto
		codigo      int
	}{
		{
			nombre:      "sin banderas, el contexto lleva los valores por omisión",
			argv:        []string{"kitlegal", "contexto", "hola"},
			ejecuciones: 1,
			contexto:    schema.Contexto{Timeout: timeoutDelContrato},
			codigo:      0,
		},
		{
			nombre: "las seis opciones globales llegan tal cual al applet",
			argv: []string{
				"kitlegal", "contexto", "hola",
				"--json", "--timeout", "5s", "--offline", "--dry-run", "--no-graph", "--asunto", "demo",
			},
			ejecuciones: 1,
			contexto: schema.Contexto{
				JSON:     true,
				Timeout:  5 * time.Second,
				Offline:  true,
				DryRun:   true,
				SinGrafo: true,
				Asunto:   "demo",
			},
			codigo: 0,
		},
		{
			nombre:      "--dry-run no corta antes del applet: la bandera viaja en el contexto",
			argv:        []string{"kitlegal", "contexto", "hola", "--dry-run"},
			ejecuciones: 1,
			contexto:    schema.Contexto{Timeout: timeoutDelContrato, DryRun: true},
			codigo:      0,
		},
		{
			nombre:      "--verbose no viaja al applet: solo fija el nivel del registro",
			argv:        []string{"kitlegal", "contexto", "hola", "--verbose"},
			ejecuciones: 1,
			contexto:    schema.Contexto{Timeout: timeoutDelContrato},
			codigo:      0,
		},
		{
			nombre:      "--describe no ejecuta el applet: describirse y actuar son excluyentes",
			argv:        []string{"kitlegal", "contexto", "hola", "--describe"},
			ejecuciones: 0,
			codigo:      0,
		},
		{
			nombre:      "--describe con --dry-run tampoco lo ejecuta",
			argv:        []string{"kitlegal", "contexto", "hola", "--describe", "--dry-run"},
			ejecuciones: 0,
			codigo:      0,
		},
		{
			nombre:      "la ayuda del applet no lo ejecuta",
			argv:        []string{"kitlegal", "contexto", "--help"},
			ejecuciones: 0,
			codigo:      0,
		},
		{
			nombre:      "la ayuda del verbo no lo ejecuta",
			argv:        []string{"kitlegal", "contexto", "anotar", "--help"},
			ejecuciones: 0,
			codigo:      0,
		},
		{
			nombre:      "una invocación mal formada no llega al applet",
			argv:        []string{"kitlegal", "contexto", "hola", "--timeout", "abc"},
			ejecuciones: 0,
			codigo:      2,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			registro, ejecuciones := registroDeContexto(t)
			res := invocar(t, registro, caso.argv...)

			assert.Equal(t, caso.codigo, res.codigo, res.errores)
			require.Len(t, *ejecuciones, caso.ejecuciones,
				"el applet se ejecuta exactamente cuando la invocación lo pide")

			if caso.ejecuciones > 0 {
				assert.Equal(t, caso.contexto, (*ejecuciones)[0],
					"el contexto de ejecución es el que las banderas describen, y nada más")
			}
		})
	}
}

// arrancar ejecuta la raíz de arranque entera, con el registro que construya
// construir, contra dos buffers.
func arrancar(t *testing.T, construir func(string) (*Registro, error), argv ...string) invocacionDePrueba {
	t.Helper()

	var salida, errores bytes.Buffer

	codigo := Arrancar(argv, construir, &salida, &errores, versionDePrueba, commitDePrueba, fechaDePrueba)

	return invocacionDePrueba{codigo: codigo, salida: salida.String(), errores: errores.String()}
}

// TestArrancar fija la raíz de arranque de los binarios (contrato puerto-y-applet
// §5 de H4; research.md D16): construye el registro una vez por arranque, con la
// versión del binario que recibe (research.md D4 de H19), y, si se construye, la
// invocación es exactamente la de Main; si no, el fallo es un defecto de
// composición y nunca de quien invoca —el sobre del kernel de clase inesperada si
// se pidió --json, el mensaje en la salida de error y el código 1, también cuando
// el error del registro declara otra clase—, sin ningún pánico.
func TestArrancar(t *testing.T) {
	t.Parallel()

	t.Run("registro-valido", func(t *testing.T) {
		t.Parallel()

		construcciones := 0
		construir := func(_ string) (*Registro, error) {
			construcciones++

			return registroDeCodigos(t, resultadoCorrecto), nil
		}

		version := arrancar(t, construir, "kitlegal", "version")
		assert.Equal(t, invocar(t, registroDeCodigos(t, resultadoCorrecto), "kitlegal", "version"), version,
			"con el registro construido, la invocación es la de Main")
		assert.Equal(t, version, arrancar(t, construir, "kitlegal", "--version"),
			"--version es «version»: el mismo código, la misma salida y nada en la de error")

		res := arrancar(t, construir, "kitlegal", "prueba", "hola", "--json")
		assert.Equal(t, 0, res.codigo, res.errores)

		sobre := sobreDelJSON(t, res.salida)
		assert.Equal(t, true, sobre["ok"])
		assert.Equal(t, procedenciaDePrueba.Fuente, sobre["fuente"])

		assert.Equal(t, 3, construcciones, "cada arranque construye su registro una sola vez")
	})

	t.Run("construir-recibe-la-version", func(t *testing.T) {
		t.Parallel()

		// La versión llega tal cual, también la que no tiene forma SemVer y la
		// vacía de quien no tiene ninguna: Arrancar no la interpreta ni la
		// sustituye, y es la misma que atiende «version» (FR-073, FR-091).
		for _, version := range []string{versionDePrueba, "v0.2.0", "dev", ""} {
			var recibidas []string

			construir := func(recibida string) (*Registro, error) {
				recibidas = append(recibidas, recibida)

				return registroDeCodigos(t, resultadoCorrecto), nil
			}

			var salida, errores bytes.Buffer

			codigo := Arrancar([]string{"kitlegal", "version"}, construir, &salida, &errores,
				version, commitDePrueba, fechaDePrueba)

			assert.Equal(t, 0, codigo, errores.String())
			assert.Equal(t, []string{version}, recibidas,
				"construir recibe, una sola vez, exactamente la versión que recibe Arrancar")
			assert.Equal(t, fmt.Sprintf(formatoDeVersion, version, commitDePrueba, fechaDePrueba), salida.String(),
				"la versión que recibe construir es la que atiende «version»")
		}
	})

	t.Run("registro-invalido", func(t *testing.T) {
		t.Parallel()

		var registro Registro

		errDelRegistro := registro.Registrar(appletConVerbos("version", verboDePrueba("mostrar", true)))
		require.ErrorIs(t, errDelRegistro, ErrNombreReservado)

		fallos := []error{
			errDelRegistro,
			// Un error que declara otra clase sigue siendo un registro que no se
			// construyó: no lo corrige quien invoca.
			fmt.Errorf("registro a medias: %w", cli.ErrArgumentos),
		}

		for _, fallo := range fallos {
			construir := func(_ string) (*Registro, error) { return nil, fallo }

			enJSON := arrancar(t, construir, "kitlegal", "prueba", "hola", "--json")
			exigirSobreDeFallo(t, enJSON, schema.ClaseInesperado, 1, cli.ProcedenciaKernel())
			assert.Contains(t, enJSON.errores, fallo.Error(), "el mensaje nombra el error del registro")

			sinJSON := arrancar(t, construir, "kitlegal", "prueba", "hola")
			assert.Equal(t, 1, sinJSON.codigo)
			assert.Empty(t, sinJSON.salida, "sin --json un fallo deja la salida estándar vacía")
			assert.Contains(t, sinJSON.errores, fallo.Error(), "el mensaje nombra el error del registro")
		}
	})
}

// errTuberiaCerrada es el fallo del descriptor que ya no admite nada.
var errTuberiaCerrada = errors.New("la tubería está cerrada")

// escritorRoto es el io.Writer que siempre falla.
type escritorRoto struct{}

func (escritorRoto) Write([]byte) (int, error) { return 0, errTuberiaCerrada }

// TestSalidaEstandarRota comprueba extremo a extremo que la salida estándar rota
// termina siempre con el código del fallo inesperado, sea lo que sea lo que se
// estaba escribiendo: el sobre, la tabla, el esquema, «version», la ayuda del
// binario, la del applet o la del verbo —que escribe Kong y que sin vigilar el
// escritor saldría con el código de argumentos inválidos— (FR-031,
// contracts/banderas-y-exit-codes.md §4).
//
// No es paralelo porque fija KITLEGAL_LOG, que es estado del proceso entero.
func TestSalidaEstandarRota(t *testing.T) {
	t.Setenv(cli.VariableNivel, "")

	escriben := [][]string{
		{"kitlegal", "version"},
		{"kitlegal", "--version"},
		{"kitlegal", "--help"},
		{"kitlegal", "prueba", "--help"},
		{"kitlegal", "prueba", "probar", "--help"},
		{"kitlegal", "prueba", "probar", "-h"},
		{"kitlegal", "prueba", "hola"},
		{"kitlegal", "prueba", "hola", "--json"},
		{"kitlegal", "prueba", "hola", "--describe"},
	}

	for _, argv := range escriben {
		t.Run(strings.Join(argv[1:], " "), func(t *testing.T) {
			t.Setenv(cli.VariableNivel, "")

			var errores strings.Builder

			codigo := Main(argv, registroDeCodigos(t, resultadoCorrecto), escritorRoto{}, &errores,
				versionDePrueba, commitDePrueba, fechaDePrueba)

			assert.Equal(t, 1, codigo, "una escritura fallida es el fallo inesperado, nunca un error de argumentos")
			assert.Contains(t, errores.String(), errTuberiaCerrada.Error(),
				"el mensaje para la persona sale por la salida de error, que sigue sana")
		})
	}

	t.Run("--dry-run no escribe en la salida estándar y no la echa en falta", func(t *testing.T) {
		t.Setenv(cli.VariableNivel, "")

		var errores strings.Builder

		codigo := Main([]string{"kitlegal", "prueba", "hola", "--dry-run", "--json"},
			registroDeCodigos(t, resultadoCorrecto), escritorRoto{}, &errores,
			versionDePrueba, commitDePrueba, fechaDePrueba)

		assert.Equal(t, 0, codigo)
		assert.Contains(t, errores.String(), "--dry-run")
	})
}

// TestDryRun comprueba las tres reglas de FR-022 sobre la misma invocación: la
// salida estándar queda vacía —también con --json—, la descripción del applet,
// el verbo y los argumentos que se habrían ejecutado van a la salida de error, y
// el código es 0.
//
// El segundo caso es el que da sentido a la decisión de research.md D10: con el
// registro de eventos apagado hasta el máximo la descripción **sigue estando**,
// porque no viaja por un canal filtrable sino por el presentador. Un requisito
// de visibilidad incondicional no se implementa sobre un canal que se filtra.
//
// No es paralelo, y no es un descuido: el nivel del registro es estado del
// proceso entero, así que fijarlo obliga a que este test tenga el proceso para
// él solo.
func TestDryRun(t *testing.T) {
	t.Setenv(cli.VariableNivel, "")

	casos := []struct {
		nombre string
		nivel  string
		argv   []string
	}{
		{
			nombre: "la salida estándar queda vacía también con --json",
			nivel:  "",
			argv:   []string{"kitlegal", "prueba", "hola", "--dry-run", "--json"},
		},
		{
			nombre: "la descripción se ve con el registro de eventos apagado",
			nivel:  "error",
			argv:   []string{"kitlegal", "prueba", "hola", "--dry-run"},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Setenv(cli.VariableNivel, caso.nivel)

			res := invocar(t, registroDeCodigos(t, resultadoCorrecto), caso.argv...)

			assert.Equal(t, 0, res.codigo, "--dry-run termina con código 0")
			assert.Empty(t, res.salida,
				"una operación no realizada no tiene nada que citar")

			assert.Contains(t, res.errores, "prueba", "la descripción nombra el applet")
			assert.Contains(t, res.errores, "probar", "la descripción nombra el verbo")
			assert.Contains(t, res.errores, "hola",
				"la descripción lleva los argumentos que se habrían ejecutado")
		})
	}
}

// TestDryRunPresentaElEnsayo comprueba lo que H2 añade a la bandera: que las
// líneas que el applet deja en schema.Resultado.Ensayo —lo que cada capa con
// efectos habría hecho en lugar de hacerlo— llegan a la salida de error detrás
// de la descripción de H1, que la salida estándar sigue vacía y que el código
// sigue siendo 0, porque estar en ensayo no es un fallo (FR-051, FR-065,
// ADR 0011). Y la otra mitad del contrato de errores §5.3: si el applet
// describió y después falló, se presenta igual lo que alcanzó a dejar en Ensayo
// y manda el fallo, con su código.
//
// El nivel del registro de eventos se fija al máximo a propósito, por la misma
// razón que en TestDryRun: un requisito de visibilidad incondicional no se
// implementa sobre un canal que se filtra, así que estas líneas tienen que
// seguir estando con el registro apagado (research.md D6). Y por eso mismo no es
// paralelo: el nivel es estado del proceso entero.
func TestDryRunPresentaElEnsayo(t *testing.T) {
	t.Setenv(cli.VariableNivel, "error")

	// Dos líneas y no una: lo que el campo promete es «una por operación», así
	// que el orden en que el applet las dejó también es observable.
	ensayo := []string{
		"GET https://fuente.prueba/norma",
		"GET https://fuente.prueba/norma/a21",
	}

	t.Run("el applet describe y termina bien", func(t *testing.T) {
		t.Setenv(cli.VariableNivel, "error")

		registro := registroDeCodigos(t, func(_ context.Context) (schema.Resultado, error) {
			return schema.Resultado{Procedencia: procedenciaDePrueba, Ensayo: ensayo}, nil
		})

		res := invocar(t, registro, "kitlegal", "prueba", "hola", "--dry-run")

		assert.Equal(t, 0, res.codigo, "el ensayo no es un fallo: --dry-run sigue terminando con 0")
		assert.Empty(t, res.salida,
			"la salida estándar no lleva la descripción del ensayo: una operación no realizada no cita nada")

		require.Contains(t, res.errores,
			"--dry-run: se habría pedido "+ensayo[0]+"\n--dry-run: se habría pedido "+ensayo[1],
			"cada línea del ensayo sale por la salida de error, y en el orden en que el applet las dejó")

		assert.Less(t,
			strings.Index(res.errores, "no se ha ejecutado nada"),
			strings.Index(res.errores, "se habría pedido"),
			"la descripción de H1 sigue encabezando lo que el ensayo añade")
	})

	t.Run("el applet describe y después falla: se presenta el ensayo y manda el fallo", func(t *testing.T) {
		t.Setenv(cli.VariableNivel, "error")

		// Un applet que alcanzó a describir una operación antes de fallar con
		// una clase conocida —la fuente que no responde, código 4—, como haría
		// un adaptador cuya segunda petición no se pudiera ni describir.
		registro := registroDeCodigos(t, func(_ context.Context) (schema.Resultado, error) {
			return schema.Resultado{Procedencia: procedenciaDePrueba, Ensayo: ensayo[:1]},
				fmt.Errorf("la fuente no ha respondido: %w", cli.ErrFuenteNoDisponible)
		})

		res := invocar(t, registro, "kitlegal", "prueba", "hola", "--dry-run")

		assert.Equal(t, 4, res.codigo,
			"manda el fallo del applet, con su clase, y no el 0 del ensayo (contrato de errores §5.3)")
		assert.Empty(t, res.salida,
			"sin --json un fallo no lleva sobre, y el ensayo tampoco escribe en la salida estándar")
		assert.Contains(t, res.errores, "--dry-run: se habría pedido "+ensayo[0],
			"lo que el applet alcanzó a dejar en Ensayo se presenta aunque haya fallado")
		assert.Contains(t, res.errores, "la fuente no ha respondido", "y el mensaje del fallo sale igualmente")
		assert.Less(t,
			strings.Index(res.errores, "se habría pedido"),
			strings.Index(res.errores, "la fuente no ha respondido"),
			"la descripción del ensayo va antes que el fallo, que es lo último que la persona lee")
	})
}

// TestPlazoAgotado comprueba la mitad de FR-020 que ninguna otra tabla verifica:
// que vencer --timeout produce el código 4 **sin depender de que el applet
// devuelva el error tipado**.
//
// El applet de este caso espera a que el plazo venza y devuelve entonces un
// resultado correcto: si el kernel se fiara de lo que el applet devuelve, la
// invocación terminaría en 0. Lo que clasifica es el plazo agotado, que es
// indistinguible desde fuera de una fuente que no responde (research.md D9).
func TestPlazoAgotado(t *testing.T) {
	t.Parallel()

	registro := registroDeCodigos(t, func(ctx context.Context) (schema.Resultado, error) {
		<-ctx.Done()

		return schema.Resultado{
			Procedencia: procedenciaDePrueba,
			Datos:       map[string]any{"mensaje": "tarde"},
		}, nil
	})

	res := invocar(t, registro, "kitlegal", "prueba", "hola", "--json", "--timeout", "10ms")

	exigirSobreDeFallo(t, res, schema.ClaseFuenteNoDisponible, 4, procedenciaDePrueba)
	assert.Contains(t, res.errores, "plazo",
		"el mensaje para la persona dice que lo que falló fue el plazo")
}
