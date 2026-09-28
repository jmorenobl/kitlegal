package schema

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSobre ejerce la forma del sobre y la validación de su procedencia, tal y
// como las fija contracts/sobre-de-salida.md §1 a §3: seis claves exactas en
// éxito y en fallo, ninguna de ellas omitida cuando lleva el valor cero,
// fecha_consulta en RFC 3339 con desplazamiento horario explícito, y fuente y
// url nunca vacías con url siempre un URI absoluto (FR-010, FR-013, FR-016,
// FR-017).
func TestSobre(t *testing.T) {
	t.Parallel()

	// Instante fijo con desplazamiento explícito: es lo que FR-013 exige y lo
	// que un reloj inyectado hará reproducible en el kernel (research.md D5).
	instante := time.Date(2026, time.September, 11, 10, 12, 0, 0, time.FixedZone("CEST", 2*60*60))
	huella := "sha256:f2a2800485840f313f12e331f6627b274d7cc6ae006f2d79e821832990a53fac"

	casos := []struct {
		nombre   string
		sobre    Sobre
		esperado string
	}{
		{
			nombre: "sobre de éxito de un applet calculado",
			sobre: Sobre{
				Ok:            true,
				Fuente:        "kitlegal.echo",
				URL:           "kitlegal:applet/echo",
				FechaConsulta: instante,
				Hash:          huella,
				Data:          map[string]any{"mensaje": "hola"},
			},
			esperado: `{"ok":true,"fuente":"kitlegal.echo","url":"kitlegal:applet/echo",` +
				`"fecha_consulta":"2026-09-11T10:12:00+02:00","hash":"` + huella + `",` +
				`"data":{"mensaje":"hola"}}`,
		},
		{
			nombre: "sobre de fallo del kernel",
			sobre: Sobre{
				Ok:            false,
				Fuente:        "kitlegal.cli",
				URL:           "kitlegal:cli",
				FechaConsulta: instante,
				Hash:          huella,
				Data:          DatosError{Clase: ClaseArgumentos, Mensaje: "bandera desconocida: --jsno"},
			},
			esperado: `{"ok":false,"fuente":"kitlegal.cli","url":"kitlegal:cli",` +
				`"fecha_consulta":"2026-09-11T10:12:00+02:00","hash":"` + huella + `",` +
				`"data":{"clase":"argumentos","mensaje":"bandera desconocida: --jsno"}}`,
		},
		{
			nombre: "el valor cero emite las seis claves: el sobre no lleva omitempty",
			sobre:  Sobre{},
			esperado: `{"ok":false,"fuente":"","url":"","fecha_consulta":"0001-01-01T00:00:00Z",` +
				`"hash":"","data":null}`,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			emitido, err := json.Marshal(caso.sobre)
			require.NoError(t, err)
			assert.Equal(t, caso.esperado, string(emitido),
				"el sobre se emite con las seis claves del contrato y ninguna más")
		})
	}

	t.Run("el sobre no admite una séptima clave", func(t *testing.T) {
		t.Parallel()

		emitido, err := json.Marshal(Sobre{})
		require.NoError(t, err)

		var claves map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(emitido, &claves))

		emitidas := make([]string, 0, len(claves))
		for clave := range claves {
			emitidas = append(emitidas, clave)
		}
		assert.ElementsMatch(t,
			[]string{"ok", "fuente", "url", "fecha_consulta", "hash", "data"}, emitidas)
	})

	casosProcedencia := []struct {
		nombre      string
		procedencia Procedencia
		esperado    error
	}{
		{
			nombre:      "procedencia del kernel",
			procedencia: Procedencia{Fuente: "kitlegal.cli", URL: "kitlegal:cli"},
		},
		{
			nombre:      "procedencia de un applet calculado",
			procedencia: Procedencia{Fuente: "kitlegal.echo", URL: "kitlegal:applet/echo"},
		},
		{
			nombre: "procedencia de una fuente pública",
			procedencia: Procedencia{
				Fuente: "boe.legislacion-consolidada",
				URL:    "https://www.boe.es/buscar/act.php?id=BOE-A-2015-10565",
			},
		},
		{
			nombre:      "esquema con dígitos, más, menos y punto",
			procedencia: Procedencia{Fuente: "fuente", URL: "esquema+raro-1.0:recurso"},
		},
		{
			nombre:      "fuente vacía",
			procedencia: Procedencia{Fuente: "", URL: "kitlegal:cli"},
			esperado:    ErrFuenteVacia,
		},
		{
			nombre:      "url vacía",
			procedencia: Procedencia{Fuente: "kitlegal.cli", URL: ""},
			esperado:    ErrURLVacia,
		},
		{
			nombre:      "url sin esquema",
			procedencia: Procedencia{Fuente: "kitlegal.cli", URL: "www.boe.es/buscar"},
			esperado:    ErrURLNoAbsoluta,
		},
		{
			nombre:      "url que es solo un nombre de máquina",
			procedencia: Procedencia{Fuente: "kitlegal.cli", URL: "www.boe.es"},
			esperado:    ErrURLNoAbsoluta,
		},
		{
			nombre:      "url relativa al esquema",
			procedencia: Procedencia{Fuente: "kitlegal.cli", URL: "//www.boe.es/buscar"},
			esperado:    ErrURLNoAbsoluta,
		},
		{
			nombre:      "esquema que empieza por dígito",
			procedencia: Procedencia{Fuente: "kitlegal.cli", URL: "1boe:recurso"},
			esperado:    ErrURLNoAbsoluta,
		},
		{
			nombre:      "url que empieza por los dos puntos",
			procedencia: Procedencia{Fuente: "kitlegal.cli", URL: ":cli"},
			esperado:    ErrURLNoAbsoluta,
		},
		{
			nombre:      "url con un carácter de control",
			procedencia: Procedencia{Fuente: "kitlegal.cli", URL: "kitlegal:c\x01li"},
			esperado:    ErrURLNoAbsoluta,
		},
	}

	for _, caso := range casosProcedencia {
		t.Run("procedencia: "+caso.nombre, func(t *testing.T) {
			t.Parallel()

			err := caso.procedencia.Validar()
			if caso.esperado == nil {
				require.NoError(t, err)

				return
			}
			require.ErrorIs(t, err, caso.esperado)
			assert.ErrorContains(t, err, "schema:",
				"el error del dominio nombra su origen")
		})
	}

	t.Run("el resultado lleva procedencia, datos, forma legible, ensayo y grafo, y nada más", func(t *testing.T) {
		t.Parallel()

		// Lo que un applet devuelve no tiene por dónde llevar `ok`, la huella
		// ni un código de salida: todo eso lo pone el kernel, y la forma de
		// garantizarlo es que el tipo no tenga más campos (FR-015, FR-044,
		// contracts/registro-y-describe.md §1).
		//
		// Ensayo es el que H2 añade: la descripción de lo que una capa con
		// efectos no llegó a hacer bajo --dry-run, que el kernel presenta y que
		// no entra en el sobre (docs/ADR/0011). Legible es el contenido contado
		// para una persona, que sin --json sustituye a la tabla mínima y tampoco
		// entra en el sobre (docs/ADR/0026). Grafo es el que añade H7: lo que la
		// invocación observó del mundo, sin fuente propia, que el kernel entrega
		// al grafo después de presentar, con la procedencia del sobre, y que
		// tampoco entra en él; su valor cero no emite nada (docs/ADR/0014,
		// FR-020, FR-021).
		//
		// La fecha de consulta no es un campo del resultado sino de su
		// procedencia, y H4 la añade como tercero: quien consultó la declara si
		// la conoce, y el kernel sigue fechando cuando no (docs/ADR/0015).
		assert.Equal(t, []string{"Procedencia", "Datos", "Legible", "Ensayo", "Grafo"},
			camposDe(reflect.TypeFor[Resultado]()))
		assert.Equal(t, []string{"Fuente", "URL", "FechaConsulta"},
			camposDe(reflect.TypeFor[Procedencia]()))
	})

	t.Run("las restricciones del contrato viajan en las etiquetas del sobre", func(t *testing.T) {
		t.Parallel()

		// De estas etiquetas deriva el kernel el esquema de --describe (FR-017,
		// FR-048). Que el esquema resultante rechace de verdad un sobre inválido
		// lo comprueba internal/cli; aquí se fija que la declaración está donde
		// el contrato dice que está y que dice lo que el contrato dice.
		etiquetas := map[string]string{
			"Fuente": "minLength=1",
			"URL":    "minLength=1,format=uri",
			"Hash":   "pattern=" + PatronHuella,
		}

		for campo, esperada := range etiquetas {
			declarado, existe := reflect.TypeFor[Sobre]().FieldByName(campo)
			require.True(t, existe, "el sobre declara el campo %s", campo)
			assert.Equal(t, esperada, declarado.Tag.Get("jsonschema"), campo)
		}

		mensaje, existe := reflect.TypeFor[DatosError]().FieldByName("Mensaje")
		require.True(t, existe)
		assert.Equal(t, "minLength=1", mensaje.Tag.Get("jsonschema"))

		huella, err := Huella(map[string]any{"mensaje": "hola"})
		require.NoError(t, err)
		assert.Regexp(t, PatronHuella, huella,
			"el patrón que declara la etiqueta es el que cumplen las huellas que produce el dominio")
	})
}

// TestProcedenciaFechaDeConsultaOpcional fija que la fecha de consulta es un
// dato opcional de la procedencia: quien consulta la declara cuando la conoce,
// su valor cero significa «no la declara quien consulta» y Validar ni la exige
// ni la tiene en cuenta. Así la ampliación es retrocompatible: una procedencia
// escrita antes de H4, sin fecha, valida igual que antes, y una fecha no hace
// citable una procedencia que no lo era (FR-096, research.md D3,
// contracts/puerto-y-applet.md §2).
func TestProcedenciaFechaDeConsultaOpcional(t *testing.T) {
	t.Parallel()

	// Un instante con desplazamiento distinto del de TestSobre, para que
	// ningún caso pase por coincidir con otro instante del paquete.
	consultada := time.Date(2026, time.September, 4, 8, 45, 30, 250_000_000, time.FixedZone("CET", 60*60))

	casos := []struct {
		nombre      string
		procedencia Procedencia
		esperado    error
	}{
		{
			nombre: "fuente pública",
			procedencia: Procedencia{
				Fuente: "boe.legislacion-consolidada",
				URL:    "https://www.boe.es/buscar/act.php?id=BOE-A-2015-10565",
			},
		},
		{
			nombre:      "applet calculado",
			procedencia: Procedencia{Fuente: "kitlegal.echo", URL: "kitlegal:applet/echo"},
		},
		{
			nombre:      "fuente vacía",
			procedencia: Procedencia{URL: "https://www.boe.es/buscar/act.php?id=BOE-A-2015-10565"},
			esperado:    ErrFuenteVacia,
		},
		{
			nombre:      "url vacía",
			procedencia: Procedencia{Fuente: "boe.legislacion-consolidada"},
			esperado:    ErrURLVacia,
		},
		{
			nombre:      "url que no es un uri absoluto",
			procedencia: Procedencia{Fuente: "boe.legislacion-consolidada", URL: "www.boe.es/buscar"},
			esperado:    ErrURLNoAbsoluta,
		},
		{
			nombre:   "valor cero",
			esperado: ErrFuenteVacia,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.True(t, caso.procedencia.FechaConsulta.IsZero(),
				"una procedencia que no nombra la fecha no la declara")

			conFecha := caso.procedencia
			conFecha.FechaConsulta = consultada

			for _, variante := range []Procedencia{caso.procedencia, conFecha} {
				err := variante.Validar()
				if caso.esperado == nil {
					require.NoError(t, err,
						"la fecha no es obligatoria: con ella o sin ella la procedencia valida (%+v)", variante)

					continue
				}
				require.ErrorIs(t, err, caso.esperado,
					"la fecha no hace citable lo que no lo es (%+v)", variante)
			}
		})
	}

	t.Run("la fecha es un instante con su zona, no un texto", func(t *testing.T) {
		t.Parallel()

		// Un time.Time conserva el desplazamiento con el que se consultó y el
		// kernel lo serializa en RFC 3339 como el resto de fecha_consulta; un
		// texto obligaría a quien consulta a elegir el formato del sobre.
		campo, existe := reflect.TypeFor[Procedencia]().FieldByName("FechaConsulta")
		require.True(t, existe, "la procedencia declara el campo FechaConsulta")
		assert.Equal(t, reflect.TypeFor[time.Time](), campo.Type)
	})
}

// camposDe enumera los campos de un struct en el orden en que se declaran.
func camposDe(tipo reflect.Type) []string {
	nombres := make([]string, 0, tipo.NumField())
	for i := range tipo.NumField() {
		nombres = append(nombres, tipo.Field(i).Name)
	}

	return nombres
}
