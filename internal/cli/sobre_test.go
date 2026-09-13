package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// esquemaDelContrato es la descripción formal del sobre, copiada de
// contracts/sobre-de-salida.md §6. Está escrita a mano y no generada a
// propósito: si saliera de los mismos tipos que producen el sobre, un error en
// esos tipos se cancelaría con el mismo error en el esquema y el test seguiría
// en verde. El esquema **generado** es otra cosa y lo comprueba TestDescribe,
// que valida estos mismos sobres contra él (T007, SC-015).
//
// La rama `then` deja `data` sin restringir porque el contrato la resuelve con
// el `$defs` del applet, que la descripción general del sobre no conoce: lo que
// aquí se fija es el nivel superior —seis claves y ninguna más— y el `data` del
// sobre de fallo.
const esquemaDelContrato = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "additionalProperties": false,
  "required": ["ok", "fuente", "url", "fecha_consulta", "hash", "data"],
  "properties": {
    "ok":             { "type": "boolean" },
    "fuente":         { "type": "string", "minLength": 1 },
    "url":            { "type": "string", "minLength": 1, "format": "uri" },
    "fecha_consulta": { "type": "string", "format": "date-time" },
    "hash":           { "type": "string", "pattern": "^sha256:[0-9a-f]{64}$" },
    "data":           true
  },
  "if":   { "properties": { "ok": { "const": true } } },
  "then": { "properties": { "data": true } },
  "else": { "properties": { "data": { "$ref": "#/$defs/DatosError" } } },
  "$defs": {
    "DatosError": {
      "type": "object",
      "additionalProperties": false,
      "required": ["clase", "mensaje"],
      "properties": {
        "clase": { "type": "string",
                   "enum": ["argumentos", "no-encontrado", "fuente-no-disponible",
                            "limite-o-tos", "identidad-humana", "inesperado"] },
        "mensaje": { "type": "string", "minLength": 1 }
      }
    }
  }
}`

// urlDelEsquema es lo único que el compilador necesita además del documento: el
// recurso se identifica por URL y el `$ref` interno se resuelve contra ella.
const urlDelEsquema = "https://ventanillalegal.es/schemas/sobre.json"

// clavesDelSobre son las seis del contrato, ni una más ni una menos.
var clavesDelSobre = []string{"ok", "fuente", "url", "fecha_consulta", "hash", "data"}

// instanteDePrueba es lo que devuelve el reloj inyectado en las tablas. Lleva un
// desplazamiento horario distinto de cero a propósito: con «Z» no se
// distinguiría una serialización que conserva el desplazamiento de otra que lo
// pierde al normalizar a UTC (FR-013).
var instanteDePrueba = time.Date(2026, time.September, 11, 10, 12, 0, 0,
	time.FixedZone("CEST", 2*60*60))

// procedenciaDelApplet es la de un applet que sí sabe de dónde salió su
// contenido. Sirve para la mitad de FR-045 que el espacio reservado del kernel
// no cubre: la procedencia de la fuente consultada se conserva también cuando
// la operación falla, porque una url que devolvió «no encontrado» es una cita
// negativa útil.
var procedenciaDelApplet = schema.Procedencia{
	Fuente: "boe.legislacion-consolidada",
	URL:    "https://www.boe.es/buscar/act.php?id=BOE-A-2015-10565",
}

// montadorDePrueba es el montador con el reloj fijado, que es lo que hace
// comprobable el valor de fecha_consulta sin depender del instante en que corra
// el test.
func montadorDePrueba() Montador {
	return Montador{Ahora: func() time.Time { return instanteDePrueba }}
}

// resultadoDePrueba es lo que devuelve un applet que terminó bien.
func resultadoDePrueba() schema.Resultado {
	return schema.Resultado{
		Procedencia: procedenciaDelApplet,
		Datos:       map[string]any{"mensaje": "hola", "veces": 2},
	}
}

// presentadorConJSON es el doble del presentador que además serializa el sobre,
// porque este test necesita leer de la salida estándar el documento que el
// validador consume. La serialización es la que T008 implementará en
// internal/render —sin escapar caracteres HTML y con salto de línea final—, y
// que el sobre real salga así lo comprueban los guiones de extremo a extremo de
// T015; aquí lo que se valida es el sobre que el kernel monta, no quién lo
// escribe.
type presentadorConJSON struct {
	presentadorDoble
}

var _ Presentador = (*presentadorConJSON)(nil)

func (p *presentadorConJSON) Presentar(sobre schema.Sobre, enJSON bool) error {
	if !enJSON {
		return p.presentadorDoble.Presentar(sobre, enJSON)
	}

	p.sobres = append(p.sobres, sobre)

	if p.fallo != nil {
		return p.fallo
	}

	codificador := json.NewEncoder(&p.salida)
	codificador.SetEscapeHTML(false)

	return codificador.Encode(sobre)
}

// dobleRoto es el presentador cuyas tres escrituras fallan, como una tubería
// cerrada.
func dobleRoto() *presentadorConJSON {
	return &presentadorConJSON{presentadorDoble: presentadorDoble{fallo: errEscrituraRota}}
}

// compilarContrato compila la descripción formal del sobre. Con conFormato, las
// aserciones de `format` quedan activadas: sin ellas el borrador 2020-12 trata
// `format` como una anotación y no como una aserción, de modo que una url que
// no es un URI pasaría y el control quedaría muerto con el test en verde
// (contracts/sobre-de-salida.md §6, nota para quien escriba el validador).
func compilarContrato(t *testing.T, conFormato bool) *jsonschema.Schema {
	t.Helper()

	documento, err := jsonschema.UnmarshalJSON(strings.NewReader(esquemaDelContrato))
	require.NoError(t, err)

	compilador := jsonschema.NewCompiler()
	if conFormato {
		compilador.AssertFormat()
	}
	require.NoError(t, compilador.AddResource(urlDelEsquema, documento))

	compilado, err := compilador.Compile(urlDelEsquema)
	require.NoError(t, err)

	return compilado
}

// unicoDocumento exige que la salida estándar lleve exactamente un documento
// JSON y nada más —lo que piden FR-042 y SC-014— y lo devuelve en la
// representación genérica que el validador consume, con los números conservados
// como literales.
func unicoDocumento(t *testing.T, salida string) map[string]any {
	t.Helper()

	decodificador := json.NewDecoder(strings.NewReader(salida))
	decodificador.UseNumber()

	var documento any
	require.NoError(t, decodificador.Decode(&documento))

	_, err := decodificador.Token()
	require.ErrorIs(t, err, io.EOF, "la salida estándar lleva algo más que el sobre")

	objeto, esObjeto := documento.(map[string]any)
	require.True(t, esObjeto, "el sobre es un objeto JSON")

	return objeto
}

// exigirContrato comprueba contra la descripción formal, con las aserciones de
// formato activadas, y además que las claves del nivel superior son exactamente
// las seis. Lo segundo no lo cubre `additionalProperties: false`, que impide una
// séptima pero no obliga a que las seis aparezcan todas por su nombre.
func exigirContrato(t *testing.T, documento map[string]any) {
	t.Helper()

	require.NoError(t, compilarContrato(t, true).Validate(documento))

	claves := make([]string, 0, len(documento))
	for clave := range documento {
		claves = append(claves, clave)
	}
	assert.ElementsMatch(t, clavesDelSobre, claves)
}

// emitirDePrueba ejecuta el único punto que traduce el desenlace de una
// invocación en código de salida y devuelve lo observable: el código y los dos
// descriptores.
func emitirDePrueba(
	doble *presentadorConJSON, enJSON bool, res schema.Resultado, err error,
) int {
	return montadorDePrueba().Emitir(doble, enJSON, res, err)
}

// casoDeFallo es una fila de la tabla de las seis clases: el error que se
// fuerza, la clase y el código que le corresponden, la procedencia que trae el
// applet —vacía en los fallos anteriores a su ejecución— y la que el sobre debe
// acabar llevando.
type casoDeFallo struct {
	nombre      string
	err         error
	clase       schema.Clase
	codigo      int
	delApplet   schema.Procedencia
	procedencia schema.Procedencia
}

// errorDeBanderaDesconocida es el primero de los dos fallos anteriores a la
// ejecución del applet que SC-014 obliga a cubrir. No se escribe a mano: sale
// del analizador real, de modo que el caso dejaría de valer si el análisis
// dejara de clasificarlo como error de argumentos.
func errorDeBanderaDesconocida(t *testing.T) error {
	t.Helper()

	_, err := Analizar(&presentadorDoble{}, nombreDePrueba, &verbosDeEjemplo{},
		[]string{"repetir", "hola", "--jsno"})
	require.Error(t, err)

	return err
}

// casosDeFallo cubre las seis clases del contrato con los dos fallos anteriores
// a la ejecución del applet incluidos, que es literalmente lo que pide SC-014.
func casosDeFallo(t *testing.T) []casoDeFallo {
	t.Helper()

	return []casoDeFallo{
		{
			nombre:      "bandera desconocida, antes de llegar al applet",
			err:         errorDeBanderaDesconocida(t),
			clase:       schema.ClaseArgumentos,
			codigo:      2,
			procedencia: ProcedenciaKernel(),
		},
		{
			// El despacho que produce este error llega en T010; aquí se escribe
			// con el mismo sentinela, que es lo que decide la clase y el código.
			// El mensaje exacto del despacho lo comprueba TestSobreDeFallo (T011).
			nombre:      "applet no registrado, antes de llegar al applet",
			err:         fmt.Errorf("%w: applet desconocido: boe", ErrArgumentos),
			clase:       schema.ClaseArgumentos,
			codigo:      2,
			procedencia: ProcedenciaKernel(),
		},
		{
			nombre:      "no encontrado, con la procedencia de la fuente consultada",
			err:         fmt.Errorf("el bloque a99 de BOE-A-2015-10565: %w", ErrNoEncontrado),
			clase:       schema.ClaseNoEncontrado,
			codigo:      3,
			delApplet:   procedenciaDelApplet,
			procedencia: procedenciaDelApplet,
		},
		{
			nombre:      "fuente no disponible",
			err:         fmt.Errorf("el BOE no responde: %w", ErrFuenteNoDisponible),
			clase:       schema.ClaseFuenteNoDisponible,
			codigo:      4,
			delApplet:   procedenciaDelApplet,
			procedencia: procedenciaDelApplet,
		},
		{
			nombre:      "límite de peticiones o términos de uso",
			err:         fmt.Errorf("una petición cada 2s: %w", ErrLimiteOTos),
			clase:       schema.ClaseLimiteOTos,
			codigo:      5,
			delApplet:   procedenciaDelApplet,
			procedencia: procedenciaDelApplet,
		},
		{
			nombre:      "identidad humana, sin haber hecho nada",
			err:         fmt.Errorf("presentar el escrito: %w", ErrIdentidadHumana),
			clase:       schema.ClaseIdentidadHumana,
			codigo:      6,
			procedencia: ProcedenciaKernel(),
		},
		{
			nombre:      "inesperado, el fallo que nadie declaró",
			err:         errEscrituraRota,
			clase:       schema.ClaseInesperado,
			codigo:      1,
			procedencia: ProcedenciaKernel(),
		},
	}
}

// TestContratoSobre comprueba que todo sobre que el kernel emite —de éxito y de
// fallo, y en fallo para las seis clases— cumple el contrato del sobre de
// salida, validándolo contra la descripción formal de contracts/sobre-de-salida.md
// §6 compilada con las aserciones de formato activadas (FR-014 … FR-017, FR-045,
// SC-014, SC-015 parcial).
//
// Es el test que invoca el escenario 6 de quickstart.md por su nombre.
func TestContratoSobre(t *testing.T) {
	t.Parallel()

	t.Run("el sobre de éxito", func(t *testing.T) {
		t.Parallel()

		doble := &presentadorConJSON{}
		codigo := emitirDePrueba(doble, true, resultadoDePrueba(), nil)

		require.Equal(t, 0, codigo)
		documento := unicoDocumento(t, doble.salida.String())
		exigirContrato(t, documento)

		assert.Equal(t, true, documento["ok"], "ok es verdadero si y solo si el código es 0")
		assert.Equal(t, procedenciaDelApplet.Fuente, documento["fuente"])
		assert.Equal(t, procedenciaDelApplet.URL, documento["url"])
		assert.Equal(t, instanteDePrueba.Format(time.RFC3339Nano), documento["fecha_consulta"])
		assert.Equal(t, map[string]any{"mensaje": "hola", "veces": json.Number("2")},
			documento["data"])

		huella, err := schema.Huella(resultadoDePrueba().Datos)
		require.NoError(t, err)
		assert.Equal(t, huella, documento["hash"],
			"la huella del sobre es la del contenido en su forma canónica")

		assert.Empty(t, doble.errores.String(), "un resultado correcto no avisa de nada")
	})

	t.Run("las seis clases de error con --json", func(t *testing.T) {
		t.Parallel()

		for _, caso := range casosDeFallo(t) {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				doble := &presentadorConJSON{}
				codigo := emitirDePrueba(
					doble, true, schema.Resultado{Procedencia: caso.delApplet}, caso.err)

				require.Equal(t, caso.codigo, codigo)

				documento := unicoDocumento(t, doble.salida.String())
				exigirContrato(t, documento)

				assert.Equal(t, caso.codigo == 0, documento["ok"],
					"ok es falso si y solo si el código de salida no es 0")
				assert.Equal(t, caso.procedencia.Fuente, documento["fuente"])
				assert.Equal(t, caso.procedencia.URL, documento["url"])
				assert.NotEmpty(t, documento["fuente"])
				assert.NotEmpty(t, documento["url"])
				assert.Equal(t, instanteDePrueba.Format(time.RFC3339Nano),
					documento["fecha_consulta"])

				assert.Equal(t,
					map[string]any{
						"clase":   string(caso.clase),
						"mensaje": caso.err.Error(),
					},
					documento["data"],
					"data es exactamente {clase, mensaje}")

				huella, err := schema.Huella(schema.DatosError{
					Clase: caso.clase, Mensaje: caso.err.Error(),
				})
				require.NoError(t, err)
				assert.Equal(t, huella, documento["hash"],
					"la huella del sobre de fallo se calcula igual que la del de éxito")

				assert.Equal(t, []string{caso.err.Error()}, doble.avisos,
					"el mensaje para la persona va además a la salida de error")
			})
		}
	})

	t.Run("sin --json el fallo deja la salida estándar vacía", func(t *testing.T) {
		t.Parallel()

		doble := &presentadorConJSON{}
		codigo := emitirDePrueba(doble, false, schema.Resultado{},
			fmt.Errorf("el bloque a99: %w", ErrNoEncontrado))

		assert.Equal(t, 3, codigo)
		assert.Empty(t, doble.salida.String(),
			"la tabla mínima es la forma de un resultado, no de un fallo")
		assert.Empty(t, doble.sobres, "sin --json no se monta ningún sobre de fallo")
		assert.Len(t, doble.avisos, 1)
	})

	t.Run("el reloj inyectado fija fecha_consulta y no toca la huella", func(t *testing.T) {
		t.Parallel()

		otroInstante := instanteDePrueba.Add(72 * time.Hour).In(time.UTC)
		otro := Montador{Ahora: func() time.Time { return otroInstante }}

		conReloj, err := montadorDePrueba().Exito(resultadoDePrueba())
		require.NoError(t, err)
		conOtroReloj, err := otro.Exito(resultadoDePrueba())
		require.NoError(t, err)

		assert.Equal(t, instanteDePrueba, conReloj.FechaConsulta)
		assert.Equal(t, otroInstante, conOtroReloj.FechaConsulta)
		assert.Equal(t, conReloj.Hash, conOtroReloj.Hash,
			"la huella no depende de fecha_consulta: por eso detecta que el contenido cambió")
	})

	t.Run("el valor cero del montador usa el reloj del sistema", func(t *testing.T) {
		t.Parallel()

		antes := time.Now()
		sobre, err := Montador{}.Exito(resultadoDePrueba())
		require.NoError(t, err)

		assert.False(t, sobre.FechaConsulta.Before(antes))
		assert.False(t, sobre.FechaConsulta.After(time.Now()))
	})

	t.Run("un data no serializable falla antes de escribir nada", func(t *testing.T) {
		t.Parallel()

		irrepresentable := schema.Resultado{
			Procedencia: procedenciaDelApplet,
			Datos:       make(chan int),
		}

		_, err := montadorDePrueba().Exito(irrepresentable)
		require.Error(t, err)

		doble := &presentadorConJSON{}
		codigo := emitirDePrueba(doble, true, irrepresentable, nil)

		assert.Equal(t, 1, codigo, "el contenido que no se puede serializar es inesperado")
		require.Len(t, doble.sobres, 1, "solo llega a la salida estándar el sobre de fallo")

		documento := unicoDocumento(t, doble.salida.String())
		exigirContrato(t, documento)
		assert.Equal(t, false, documento["ok"])
		assert.Equal(t, procedenciaDelApplet.Fuente, documento["fuente"])
	})

	t.Run("una procedencia que no sostiene una cita es un fallo del applet", func(t *testing.T) {
		t.Parallel()

		for _, invalida := range []schema.Procedencia{
			{},
			{Fuente: "kitlegal.echo"},
			{Fuente: "kitlegal.echo", URL: "applet/echo"},
		} {
			_, err := montadorDePrueba().Exito(schema.Resultado{Procedencia: invalida})
			require.Error(t, err, "procedencia %+v", invalida)

			doble := &presentadorConJSON{}
			codigo := emitirDePrueba(doble, true, schema.Resultado{Procedencia: invalida}, nil)

			assert.Equal(t, 1, codigo)
			documento := unicoDocumento(t, doble.salida.String())
			exigirContrato(t, documento)
			assert.Equal(t, ProcedenciaKernel().Fuente, documento["fuente"],
				"una procedencia que no se puede citar no se conoce: firma el kernel")
			assert.Equal(t, ProcedenciaKernel().URL, documento["url"])
		}
	})

	t.Run("montar un sobre de fallo sin fallo no es posible", func(t *testing.T) {
		t.Parallel()

		_, err := montadorDePrueba().Fallo(procedenciaDelApplet, nil)
		assert.ErrorIs(t, err, errSinFallo)
	})

	t.Run("un sobre de fallo que no se pudo montar no se escribe", func(t *testing.T) {
		t.Parallel()

		// escribirFallo es la única puerta por la que un sobre de fallo llega a
		// la salida estándar, y no la cruza si el montaje no salió: sin ese
		// guardián se escribiría el sobre cero —seis claves vacías— y el
		// consumidor recibiría un documento que incumple el contrato. Quien
		// llama nunca le pasa un error nulo, así que la comprobación se hace
		// aquí, directamente sobre el guardián.
		doble := &presentadorConJSON{}
		err := montadorDePrueba().escribirFallo(doble, procedenciaDelApplet, nil)

		require.ErrorIs(t, err, errSinFallo)
		assert.Empty(t, doble.sobres, "no se escribe un sobre que no se pudo montar")
		assert.Empty(t, doble.salida.String())
	})

	t.Run("no hay un segundo sobre si la escritura del primero falló", func(t *testing.T) {
		t.Parallel()

		t.Run("cuando falla la del sobre de éxito", func(t *testing.T) {
			t.Parallel()

			doble := dobleRoto()
			codigo := emitirDePrueba(doble, true, resultadoDePrueba(), nil)

			assert.Equal(t, 1, codigo, "una escritura fallida es un fallo inesperado")
			assert.Len(t, doble.sobres, 1, "no se intenta un segundo sobre por el descriptor roto")
			assert.Equal(t, []string{errEscrituraRota.Error()}, doble.avisos,
				"queda el mensaje para la persona en la salida de error")
		})

		t.Run("cuando falla la del sobre de fallo", func(t *testing.T) {
			t.Parallel()

			doble := dobleRoto()
			original := fmt.Errorf("el bloque a99: %w", ErrNoEncontrado)
			codigo := emitirDePrueba(doble, true, schema.Resultado{}, original)

			assert.Equal(t, 1, codigo,
				"el fallo de la escritura manda sobre el código de la clase")
			assert.Len(t, doble.sobres, 1)
			assert.Equal(t, []string{original.Error()}, doble.avisos,
				"el mensaje para la persona sigue siendo el del fallo, no el de la escritura")
		})
	})

	t.Run("sin AssertFormat el control de la url queda muerto", func(t *testing.T) {
		t.Parallel()

		conFormato := compilarContrato(t, true)
		sinFormato := compilarContrato(t, false)

		// Un sobre por lo demás impecable cuya url no es un URI absoluto. El
		// montador no lo produce —Procedencia.Validar lo rechaza—, así que se
		// escribe a mano: lo que se comprueba aquí es que la descripción formal
		// lo rechazaría igualmente, que es la mitad de FR-017 que no depende
		// del código.
		noEsURI := documentoConURL(t, "no-es-un-uri")
		require.Error(t, conFormato.Validate(noEsURI))
		require.NoError(t, sinFormato.Validate(noEsURI),
			"sin AssertFormat, `format` es una anotación: este es el control que quedaría muerto")

		vacia := documentoConURL(t, "")
		require.Error(t, conFormato.Validate(vacia))
		require.Error(t, sinFormato.Validate(vacia), "la url vacía la para minLength")

		require.NoError(t, conFormato.Validate(documentoConURL(t, ProcedenciaKernel().URL)),
			"el esquema de URI reservado del kernel es un URI absoluto")
	})
}

// documentoConURL construye un sobre de fallo correcto salvo por su url, que es
// la que se le dé. Es el material del caso negativo de FR-017.
func documentoConURL(t *testing.T, url string) map[string]any {
	t.Helper()

	sobre := schema.Sobre{
		Ok:            false,
		Fuente:        ProcedenciaKernel().Fuente,
		URL:           url,
		FechaConsulta: instanteDePrueba,
		Hash:          strings.Repeat("0", 64),
		Data:          schema.DatosError{Clase: schema.ClaseArgumentos, Mensaje: "bandera desconocida"},
	}
	sobre.Hash = schema.PrefijoHuella + sobre.Hash

	crudo, err := json.Marshal(sobre)
	require.NoError(t, err)

	documento, err := jsonschema.UnmarshalJSON(strings.NewReader(string(crudo)))
	require.NoError(t, err)

	objeto, esObjeto := documento.(map[string]any)
	require.True(t, esObjeto)

	return objeto
}

// TestMontadorFechaDeConsulta comprueba con qué instante fecha el montador el
// sobre ahora que la procedencia puede declarar la fecha de consulta: con esa
// fecha cuando la procedencia es válida y la declara, en éxito y en fallo por
// el mismo camino; con su reloj cuando no la declara; y con su reloj y la firma
// del kernel cuando la procedencia no sostiene una cita, aunque traiga fecha.
// La huella no depende de la fecha en ningún caso (FR-096, ADR 0006,
// contracts/puerto-y-applet.md §2).
//
// Es el test que invoca el escenario 4 de quickstart.md por su nombre.
func TestMontadorFechaDeConsulta(t *testing.T) {
	t.Parallel()

	// La consulta lleva otro desplazamiento horario que el reloj del montador:
	// un sobre que tomara el del reloj, o que normalizara el instante a otra
	// zona, no pasaría por coincidencia.
	consultada := time.Date(2026, time.September, 4, 8, 45, 30, 250_000_000,
		time.FixedZone("CET", 60*60))

	conFecha := procedenciaDelApplet
	conFecha.FechaConsulta = consultada

	noEncontrado := fmt.Errorf("el bloque a99 de BOE-A-2015-10565: %w", ErrNoEncontrado)

	casos := []struct {
		nombre      string
		recibida    schema.Procedencia
		err         error
		codigo      int
		procedencia schema.Procedencia
		fecha       time.Time
	}{
		{
			nombre:      "con fecha, en éxito: la de la consulta",
			recibida:    conFecha,
			codigo:      0,
			procedencia: procedenciaDelApplet,
			fecha:       consultada,
		},
		{
			nombre:      "con fecha, en fallo: la de la consulta",
			recibida:    conFecha,
			err:         noEncontrado,
			codigo:      3,
			procedencia: procedenciaDelApplet,
			fecha:       consultada,
		},
		{
			nombre:      "sin fecha, en éxito: el reloj del montador",
			recibida:    procedenciaDelApplet,
			codigo:      0,
			procedencia: procedenciaDelApplet,
			fecha:       instanteDePrueba,
		},
		{
			nombre:      "sin fecha, en fallo: el reloj del montador",
			recibida:    procedenciaDelApplet,
			err:         noEncontrado,
			codigo:      3,
			procedencia: procedenciaDelApplet,
			fecha:       instanteDePrueba,
		},
		{
			nombre: "inválida con fecha, en fallo: el kernel y su reloj",
			recibida: schema.Procedencia{
				Fuente:        procedenciaDelApplet.Fuente,
				URL:           "www.boe.es/buscar/act.php?id=BOE-A-2015-10565",
				FechaConsulta: consultada,
			},
			err:         noEncontrado,
			codigo:      3,
			procedencia: ProcedenciaKernel(),
			fecha:       instanteDePrueba,
		},
		{
			// El montaje del éxito rechaza la procedencia y el fallo que eso
			// provoca es del applet: sale por el sobre de fallo, con el kernel.
			nombre: "inválida con fecha, en éxito: el kernel y su reloj",
			recibida: schema.Procedencia{
				Fuente:        procedenciaDelApplet.Fuente,
				URL:           "www.boe.es/buscar/act.php?id=BOE-A-2015-10565",
				FechaConsulta: consultada,
			},
			codigo:      1,
			procedencia: ProcedenciaKernel(),
			fecha:       instanteDePrueba,
		},
		{
			nombre:      "solo con fecha, en fallo: el kernel y su reloj",
			recibida:    schema.Procedencia{FechaConsulta: consultada},
			err:         noEncontrado,
			codigo:      3,
			procedencia: ProcedenciaKernel(),
			fecha:       instanteDePrueba,
		},
		{
			// La de un fallo anterior a cualquier petición: el applet no llegó
			// a construir ninguna y devuelve la procedencia cero.
			nombre:      "cero, en fallo: el kernel y su reloj",
			recibida:    schema.Procedencia{},
			err:         noEncontrado,
			codigo:      3,
			procedencia: ProcedenciaKernel(),
			fecha:       instanteDePrueba,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			res := resultadoDePrueba()
			res.Procedencia = caso.recibida

			doble := &presentadorConJSON{}
			codigo := emitirDePrueba(doble, true, res, caso.err)

			require.Equal(t, caso.codigo, codigo)
			require.Len(t, doble.sobres, 1, "se emite un único sobre")
			assert.Equal(t, caso.fecha, doble.sobres[0].FechaConsulta)

			documento := unicoDocumento(t, doble.salida.String())
			exigirContrato(t, documento)
			assert.Equal(t, caso.procedencia.Fuente, documento["fuente"])
			assert.Equal(t, caso.procedencia.URL, documento["url"])
			assert.Equal(t, caso.fecha.Format(time.RFC3339Nano), documento["fecha_consulta"],
				"fecha_consulta sale en RFC 3339 con el desplazamiento del instante elegido")
		})
	}

	t.Run("la huella no depende de la fecha", func(t *testing.T) {
		t.Parallel()

		montador := montadorDePrueba()

		exitoConFecha, err := montador.Exito(
			schema.Resultado{Procedencia: conFecha, Datos: resultadoDePrueba().Datos})
		require.NoError(t, err)
		exitoSinFecha, err := montador.Exito(resultadoDePrueba())
		require.NoError(t, err)

		assert.Equal(t, consultada, exitoConFecha.FechaConsulta)
		assert.Equal(t, instanteDePrueba, exitoSinFecha.FechaConsulta)
		assert.Equal(t, exitoSinFecha.Hash, exitoConFecha.Hash,
			"el mismo data con dos fechas distintas tiene la misma huella")

		falloConFecha, err := montador.Fallo(conFecha, noEncontrado)
		require.NoError(t, err)
		falloSinFecha, err := montador.Fallo(procedenciaDelApplet, noEncontrado)
		require.NoError(t, err)

		assert.Equal(t, consultada, falloConFecha.FechaConsulta)
		assert.Equal(t, instanteDePrueba, falloSinFecha.FechaConsulta)
		assert.Equal(t, falloSinFecha.Hash, falloConFecha.Hash,
			"en el sobre de fallo la huella se calcula igual que en el de éxito")
	})
}
