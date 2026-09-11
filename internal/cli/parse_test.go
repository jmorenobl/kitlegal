package cli

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// nombreDePrueba es el nombre con el que se invoca el applet de las tablas. No
// es «kitlegal» a propósito: así la ayuda que emite el analizador demuestra que
// el nombre llega desde quien llama y no está escrito dentro del kernel.
const nombreDePrueba = "echo"

// Los dos verbos del applet de prueba, declarados con nombre propio para que las
// tablas puedan escribir el resultado esperado como un literal. Son lo que el
// paquete de composición entrega al kernel como `any`: el kernel no sabe de
// dónde salen ni conoce su tipo.
type (
	verboRepetir struct {
		Texto string `arg:"" help:"Lo que se repite."`
	}

	verboContar struct {
		Veces int `arg:"" help:"Cuántas veces se cuenta."`
	}
)

// verbosDeEjemplo tiene dos verbos y ninguno por omisión, que es la rama en la
// que no nombrar verbo es un error de argumentos. Insertar el verbo por omisión
// cuando lo hay ocurre antes y fuera de aquí (research.md D26).
type verbosDeEjemplo struct {
	Repetir verboRepetir `cmd:"" help:"Repite lo que se le da."`
	Contar  verboContar  `cmd:"" help:"Cuenta hasta donde se le diga."`
}

// analisisDelKernel recoge todo lo observable de una llamada al análisis: lo que
// devolvió, lo que quedó escrito en la gramática del paquete de composición y lo
// que salió por cada uno de los dos descriptores del presentador.
type analisisDelKernel struct {
	analisis Analisis
	err      error
	verbos   *verbosDeEjemplo
	doble    *presentadorDoble
}

// analizarDePrueba analiza una invocación con el applet de las tablas.
//
// Que esta función retorne es, igual que en el test de los supuestos, la
// comprobación de que ningún camino del análisis termina el proceso por su
// cuenta: la ayuda integrada pide terminar y el kernel la atiende devolviendo
// una decisión (gates/supuestos-kong.md, S3).
func analizarDePrueba(t *testing.T, args ...string) analisisDelKernel {
	t.Helper()

	resultado := analisisDelKernel{verbos: &verbosDeEjemplo{}, doble: &presentadorDoble{}}
	resultado.analisis, resultado.err = Analizar(
		resultado.doble, nombreDePrueba, resultado.verbos, args)

	return resultado
}

// exigirDescriptoresLimpios comprueba que el análisis no ha escrito nada por
// ningún camino. Kong no imprime el error de una invocación mal formada, y el
// kernel tampoco: el mensaje para la persona lo emite el presentador más tarde y
// una sola vez (FR-040, gates/supuestos-kong.md, detalle de S5).
func exigirDescriptoresLimpios(t *testing.T, res analisisDelKernel) {
	t.Helper()

	assert.Empty(t, res.doble.salida.String())
	assert.Empty(t, res.doble.errores.String())
	assert.Empty(t, res.doble.textos)
	assert.Empty(t, res.doble.avisos)
	assert.Empty(t, res.doble.sobres)
}

// casoDeAnalisis es una fila de la tabla de invocaciones bien formadas: lo que
// se escribe y lo que el kernel entiende de ello.
type casoDeAnalisis struct {
	nombre   string
	args     []string
	decision Decision
	verbo    string
	globales Globales
	verbos   verbosDeEjemplo
}

// casosDeAnalisis cubre las ocho banderas sobre una invocación que se ejecuta, y
// la independencia del orden en que se escriben: una bandera global significa lo
// mismo antes del verbo, después del verbo y después de sus argumentos.
func casosDeAnalisis() []casoDeAnalisis {
	return []casoDeAnalisis{
		{
			nombre:   "el verbo y su argumento, sin ninguna bandera",
			args:     []string{"repetir", "hola"},
			decision: DecisionEjecutar,
			verbo:    "repetir",
			globales: Globales{Timeout: timeoutPorOmision},
			verbos:   verbosDeEjemplo{Repetir: verboRepetir{Texto: "hola"}},
		},
		{
			nombre:   "el otro verbo del applet",
			args:     []string{"contar", "3"},
			decision: DecisionEjecutar,
			verbo:    "contar",
			globales: Globales{Timeout: timeoutPorOmision},
			verbos:   verbosDeEjemplo{Contar: verboContar{Veces: 3}},
		},
		{
			nombre:   "--json después de los argumentos del verbo",
			args:     []string{"repetir", "hola", "--json"},
			decision: DecisionEjecutar,
			verbo:    "repetir",
			globales: Globales{JSON: true, Timeout: timeoutPorOmision},
			verbos:   verbosDeEjemplo{Repetir: verboRepetir{Texto: "hola"}},
		},
		{
			nombre:   "--json antes del verbo significa lo mismo",
			args:     []string{"--json", "repetir", "hola"},
			decision: DecisionEjecutar,
			verbo:    "repetir",
			globales: Globales{JSON: true, Timeout: timeoutPorOmision},
			verbos:   verbosDeEjemplo{Repetir: verboRepetir{Texto: "hola"}},
		},
		{
			nombre:   "--json entre el verbo y su argumento también",
			args:     []string{"repetir", "--json", "hola"},
			decision: DecisionEjecutar,
			verbo:    "repetir",
			globales: Globales{JSON: true, Timeout: timeoutPorOmision},
			verbos:   verbosDeEjemplo{Repetir: verboRepetir{Texto: "hola"}},
		},
		{
			nombre:   "--json=false deja la forma para personas",
			args:     []string{"repetir", "hola", "--json=false"},
			decision: DecisionEjecutar,
			verbo:    "repetir",
			globales: Globales{Timeout: timeoutPorOmision},
			verbos:   verbosDeEjemplo{Repetir: verboRepetir{Texto: "hola"}},
		},
		{
			nombre:   "--timeout con unidad sustituye al plazo por omisión",
			args:     []string{"repetir", "hola", "--timeout=1500ms"},
			decision: DecisionEjecutar,
			verbo:    "repetir",
			globales: Globales{Timeout: 1500 * time.Millisecond},
			verbos:   verbosDeEjemplo{Repetir: verboRepetir{Texto: "hola"}},
		},
		{
			nombre:   "--timeout con valor separado del nombre",
			args:     []string{"repetir", "hola", "--timeout", "2m"},
			decision: DecisionEjecutar,
			verbo:    "repetir",
			globales: Globales{Timeout: 2 * time.Minute},
			verbos:   verbosDeEjemplo{Repetir: verboRepetir{Texto: "hola"}},
		},
		{
			nombre:   "--asunto lleva una cadena",
			args:     []string{"repetir", "hola", "--asunto=expediente-3"},
			decision: DecisionEjecutar,
			verbo:    "repetir",
			globales: Globales{Timeout: timeoutPorOmision, Asunto: "expediente-3"},
			verbos:   verbosDeEjemplo{Repetir: verboRepetir{Texto: "hola"}},
		},
		{
			nombre: "las siete que no excluyen la ejecución, a la vez",
			args: []string{
				"--offline", "--dry-run", "--no-graph", "--verbose",
				"repetir", "hola",
				"--json", "--timeout=45s", "--asunto=expediente-3",
			},
			decision: DecisionEjecutar,
			verbo:    "repetir",
			globales: Globales{
				JSON:     true,
				Timeout:  45 * time.Second,
				Offline:  true,
				DryRun:   true,
				SinGrafo: true,
				Asunto:   "expediente-3",
				Verbose:  true,
			},
			verbos: verbosDeEjemplo{Repetir: verboRepetir{Texto: "hola"}},
		},
		{
			nombre:   "tras el terminador, lo que parece una bandera es el argumento",
			args:     []string{"repetir", "--", "--json"},
			decision: DecisionEjecutar,
			verbo:    "repetir",
			globales: Globales{Timeout: timeoutPorOmision},
			verbos:   verbosDeEjemplo{Repetir: verboRepetir{Texto: "--json"}},
		},
	}
}

// TestAnalizar comprueba la gramática de una invocación bien formada: que las
// ocho globales llegan analizadas, que el verbo y sus argumentos quedan en la
// gramática que aporta el paquete de composición y que el análisis no escribe
// nada (FR-018 … FR-021, FR-023, FR-024, FR-028).
func TestAnalizar(t *testing.T) {
	t.Parallel()

	for _, caso := range casosDeAnalisis() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			res := analizarDePrueba(t, caso.args...)

			require.NoError(t, res.err)
			assert.Equal(t, caso.decision, res.analisis.Decision)
			assert.Equal(t, caso.verbo, res.analisis.Verbo)
			assert.Equal(t, caso.globales, res.analisis.Globales)
			assert.Equal(t, caso.verbos, *res.verbos)
			exigirDescriptoresLimpios(t, res)
		})
	}
}

// TestAnalizarContexto comprueba que lo analizado llega al applet convertido en
// contexto de ejecución, que es lo único que el applet ve de las ocho banderas
// (FR-018).
func TestAnalizarContexto(t *testing.T) {
	t.Parallel()

	res := analizarDePrueba(t,
		"repetir", "hola", "--offline", "--no-graph", "--asunto=expediente-3", "--timeout=45s")

	require.NoError(t, res.err)
	assert.Equal(t, schema.Contexto{
		Timeout:  45 * time.Second,
		Offline:  true,
		SinGrafo: true,
		Asunto:   "expediente-3",
	}, res.analisis.Globales.Contexto())
}

// casoDePrecedencia es una fila de la tabla de FR-028: lo que se escribe y la
// decisión que el kernel toma, que no depende del orden.
type casoDePrecedencia struct {
	nombre   string
	args     []string
	decision Decision
	verbo    string
}

// casosDePrecedencia enumera las tres ramas y sus combinaciones en los dos
// órdenes posibles: --help gana sobre todo, --describe gana sobre la ejecución y
// en otro caso se ejecuta (research.md D11).
func casosDePrecedencia() []casoDePrecedencia {
	return []casoDePrecedencia{
		{
			nombre:   "--help solo",
			args:     []string{"--help"},
			decision: DecisionAyuda,
		},
		{
			nombre:   "--help con un verbo completo",
			args:     []string{"repetir", "hola", "--help"},
			decision: DecisionAyuda,
		},
		{
			nombre:   "--help antes del verbo",
			args:     []string{"--help", "repetir", "hola"},
			decision: DecisionAyuda,
		},
		{
			nombre:   "--help con el verbo pero sin su argumento",
			args:     []string{"repetir", "--help"},
			decision: DecisionAyuda,
		},
		{
			nombre:   "la forma corta de la ayuda integrada",
			args:     []string{"-h"},
			decision: DecisionAyuda,
		},
		{
			nombre:   "--json no altera la ayuda",
			args:     []string{"--json", "--help"},
			decision: DecisionAyuda,
		},
		{
			nombre:   "--help antes de --json tampoco",
			args:     []string{"--help", "--json"},
			decision: DecisionAyuda,
		},
		{
			nombre:   "--help gana a --describe",
			args:     []string{"repetir", "hola", "--help", "--describe"},
			decision: DecisionAyuda,
		},
		{
			nombre:   "--describe antes de --help da lo mismo",
			args:     []string{"repetir", "hola", "--describe", "--help"},
			decision: DecisionAyuda,
		},
		{
			nombre:   "--describe gana a la ejecución",
			args:     []string{"repetir", "hola", "--describe"},
			decision: DecisionDescribir,
			verbo:    "repetir",
		},
		{
			nombre:   "--describe antes del verbo da lo mismo",
			args:     []string{"--describe", "repetir", "hola"},
			decision: DecisionDescribir,
			verbo:    "repetir",
		},
		{
			nombre:   "--describe con --json sigue siendo describir",
			args:     []string{"repetir", "hola", "--describe", "--json"},
			decision: DecisionDescribir,
			verbo:    "repetir",
		},
		{
			nombre:   "sin ninguna de las dos, se ejecuta",
			args:     []string{"repetir", "hola"},
			decision: DecisionEjecutar,
			verbo:    "repetir",
		},
	}
}

// TestAnalizarPrecedencia comprueba la prelación fija e independiente del orden
// de escritura entre la ayuda, la autodescripción y la ejecución (FR-026,
// FR-028, FR-049 parcial, research.md D11).
func TestAnalizarPrecedencia(t *testing.T) {
	t.Parallel()

	for _, caso := range casosDePrecedencia() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			res := analizarDePrueba(t, caso.args...)

			require.NoError(t, res.err,
				"ninguna de las tres ramas es un fallo de la invocación")
			assert.Equal(t, caso.decision, res.analisis.Decision)
			assert.Equal(t, caso.verbo, res.analisis.Verbo)
		})
	}
}

// TestAnalizarAyuda comprueba lo que la tabla de precedencia no puede ver: que
// la ayuda es texto para personas, que sale por la salida estándar del
// presentador y que --json no la altera en absoluto, se escriba donde se
// escriba (FR-026, FR-042, research.md D11).
func TestAnalizarAyuda(t *testing.T) {
	t.Parallel()

	sola := analizarDePrueba(t, "--help")

	require.NoError(t, sola.err)
	assert.Contains(t, sola.doble.salida.String(), "Usage: "+nombreDePrueba,
		"la ayuda nombra el applet tal como se invocó")
	assert.Contains(t, sola.doble.salida.String(), "repetir",
		"la ayuda del applet enumera sus verbos")
	assert.Empty(t, sola.doble.errores.String(),
		"la ayuda no es un aviso: va a la salida estándar y no a la de error")
	assert.Empty(t, sola.doble.sobres,
		"la ayuda no es un resultado: no se emite ningún sobre")

	ordenes := map[string][]string{
		"--json antes de --help":   {"--json", "--help"},
		"--json después de --help": {"--help", "--json"},
	}

	for nombre, args := range ordenes {
		t.Run(nombre, func(t *testing.T) {
			t.Parallel()

			conJSON := analizarDePrueba(t, args...)

			require.NoError(t, conJSON.err)
			assert.Equal(t, sola.doble.salida.String(), conJSON.doble.salida.String(),
				"--json no altera la ayuda: es exactamente el mismo texto")
			assert.Empty(t, conJSON.doble.errores.String())
		})
	}
}

// casoDeArgumentosInvalidos es una fila de la tabla de FR-027: lo que se escribe
// y el fragmento que el mensaje del fallo debe nombrar para que quien invoca
// sepa qué corregir.
type casoDeArgumentosInvalidos struct {
	nombre string
	args   []string
	nombra string
}

// casosDeArgumentosInvalidos cubre las cuatro causas de FR-027 —bandera
// desconocida, valor con formato inválido, argumento obligatorio ausente y
// mandato desconocido— más las dos formas en que --timeout puede venir mal:
// sin unidad, que rechaza el análisis, y no positivo, que rechaza el kernel.
func casosDeArgumentosInvalidos() []casoDeArgumentosInvalidos {
	return []casoDeArgumentosInvalidos{
		{
			nombre: "bandera desconocida",
			args:   []string{"repetir", "hola", "--bandera-que-nadie-ha-declarado"},
			nombra: "--bandera-que-nadie-ha-declarado",
		},
		{
			nombre: "bandera desconocida antes del verbo",
			args:   []string{"--bandera-que-nadie-ha-declarado", "repetir", "hola"},
			nombra: "--bandera-que-nadie-ha-declarado",
		},
		{
			nombre: "bandera desconocida junto a --help, que no llega a escribirse",
			args:   []string{"repetir", "hola", "--bandera-que-nadie-ha-declarado", "--help"},
			nombra: "--bandera-que-nadie-ha-declarado",
		},
		{
			nombre: "bandera desconocida después de --help",
			args:   []string{"--help", "repetir", "hola", "--bandera-que-nadie-ha-declarado"},
			nombra: "--bandera-que-nadie-ha-declarado",
		},
		{
			nombre: "verbo desconocido",
			args:   []string{"verbo-que-nadie-ha-declarado"},
			nombra: "verbo-que-nadie-ha-declarado",
		},
		{
			nombre: "sin verbo, cuando el applet no declara ninguno por omisión",
			args:   nil,
			nombra: "repetir",
		},
		{
			nombre: "argumento obligatorio ausente",
			args:   []string{"repetir"},
			nombra: "texto",
		},
		{
			nombre: "argumento de más",
			args:   []string{"repetir", "hola", "sobrante"},
			nombra: "sobrante",
		},
		{
			nombre: "valor con formato inválido",
			args:   []string{"contar", "tres"},
			nombra: "tres",
		},
		{
			nombre: "duración con formato inválido",
			args:   []string{"repetir", "hola", "--timeout=abc"},
			nombra: "--timeout",
		},
		{
			nombre: "duración sin unidad",
			args:   []string{"repetir", "hola", "--timeout=30"},
			nombra: "--timeout",
		},
		{
			nombre: "valor booleano que no es ninguno de los seis",
			args:   []string{"repetir", "hola", "--json=quizás"},
			nombra: "quizás",
		},
		{
			nombre: "bandera con valor cuyo valor falta",
			args:   []string{"repetir", "hola", "--asunto"},
			nombra: "--asunto",
		},
		{
			nombre: "--timeout cero",
			args:   []string{"repetir", "hola", "--timeout=0"},
			nombra: "--timeout",
		},
		{
			nombre: "--timeout negativo",
			args:   []string{"repetir", "hola", "--timeout=-5s"},
			nombra: "--timeout",
		},
		{
			nombre: "--timeout no positivo también excluye describir",
			args:   []string{"repetir", "hola", "--describe", "--timeout=0"},
			nombra: "--timeout",
		},
	}
}

// TestAnalizarArgumentosInvalidos comprueba que todo fallo del análisis es un
// error de argumentos —código 2—, que el mensaje nombra el problema concreto y
// que el análisis no lo imprime por su cuenta (FR-027, FR-029, FR-030, FR-040).
func TestAnalizarArgumentosInvalidos(t *testing.T) {
	t.Parallel()

	for _, caso := range casosDeArgumentosInvalidos() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			res := analizarDePrueba(t, caso.args...)

			require.ErrorIs(t, res.err, ErrArgumentos)
			assert.Equal(t, codigoArgumentos, CodigoSalida(res.err))
			assert.Equal(t, schema.ClaseArgumentos, Clasificar(res.err))
			assert.Contains(t, res.err.Error(), caso.nombra)
			assert.Equal(t, Analisis{}, res.analisis,
				"un análisis que falla no entrega decisión ninguna")
			exigirDescriptoresLimpios(t, res)
		})
	}
}

// TestAnalizarSinVerbos comprueba la gramática degenerada: un applet sin ningún
// verbo no lo admite el registro, que lo rechaza al construirse (FR-008), pero
// el kernel no da por supuesta esa garantía y una invocación así no acaba en
// pánico sino sin verbo que nombrar.
func TestAnalizarSinVerbos(t *testing.T) {
	t.Parallel()

	doble := &presentadorDoble{}

	analisis, err := Analizar(doble, nombreDePrueba, &struct{}{}, []string{"--json"})

	require.NoError(t, err)
	assert.Equal(t, DecisionEjecutar, analisis.Decision)
	assert.Empty(t, analisis.Verbo)
	assert.Equal(t, Globales{JSON: true, Timeout: timeoutPorOmision}, analisis.Globales)
}

// verbosQueRedefinenUnaGlobal es un applet mal escrito: declara una bandera que
// ya es global. La gramática no se puede construir, y eso no es un fallo de
// quien invoca.
type verbosQueRedefinenUnaGlobal struct {
	Repetir struct {
		JSON  bool   `help:"Una global que este applet no debería declarar."`
		Texto string `arg:""`
	} `cmd:""`
}

// TestAnalizarGramaticaInvalida comprueba la frontera entre el error de quien
// invoca y el defecto de quien programa: una gramática que no se puede
// construir no es un error de argumentos —no hay invocación que corregir— y sale
// con el código de lo inesperado, nunca con el 2 (FR-018, FR-031).
func TestAnalizarGramaticaInvalida(t *testing.T) {
	t.Parallel()

	doble := &presentadorDoble{}

	analisis, err := Analizar(doble, nombreDePrueba, &verbosQueRedefinenUnaGlobal{},
		[]string{"repetir", "hola"})

	require.Error(t, err)
	require.NotErrorIs(t, err, ErrArgumentos,
		"un applet que redefine una global no convierte la invocación en inválida")
	assert.Equal(t, schema.ClaseInesperado, Clasificar(err))
	assert.Equal(t, codigoInesperado, CodigoSalida(err))
	assert.Contains(t, err.Error(), "json",
		"el mensaje nombra la bandera que colisiona")
	assert.Equal(t, Analisis{}, analisis)
	assert.Empty(t, doble.salida.String())
	assert.Empty(t, doble.errores.String())
}
