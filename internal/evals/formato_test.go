package evals

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/skills"
	"github.com/jmorenobl/kitlegal/internal/source/boe"
)

// nombreDeEval es el nombre de fichero con el que los tests leen cada eval
// sintética, y por el que tiene que empezar todo error (US4, escenario 4).
const nombreDeEval = "01-lpac-articulo-21.yaml"

// Trozos de las evals sintéticas de TestLeerEval, cada uno con sus líneas
// completas: la pregunta del contrato evals-y-grabaciones §1 y el comando y la
// cita del artículo 21 de la LPAC.
const (
	preguntaDelArticulo21 = "pregunta: \"¿qué dice el art. 21 de la Ley 39/2015?\"\n"
	comandoDelArticulo21  = "comandos:\n" +
		"  - applet: boe\n" +
		"    norma: BOE-A-2015-10565\n" +
		"    bloque: a21\n"
	citaDelArticulo21 = "citas:\n" +
		"  - norma: BOE-A-2015-10565\n" +
		"    bloque: a21\n"
)

// TestLeerEval fija la lectura de una eval del contrato evals-y-grabaciones §1:
// las tres formas de comando, con y sin reproduce, y la de no activación se leen
// enteras y con su Fichero; cada fichero inválido, con una clave repetida
// incluida, da un error que empieza por su nombre y dice qué falla y dónde.
func TestLeerEval(t *testing.T) {
	t.Parallel()

	citaDeLaLRBRL := []CitaEsperada{{Norma: "BOE-A-1985-5392", Bloque: "a85bis."}}

	casos := []struct {
		nombre     string
		documento  string
		leida      Eval
		error      string
		fragmentos []string
	}{
		{
			nombre: "comando-de-bloque",
			documento: "# Eval positiva: debe activar la skill, leer el bloque y citarlo.\n" +
				preguntaDelArticulo21 + "activa: true\n" + comandoDelArticulo21 + citaDelArticulo21,
			leida: Eval{
				Fichero:  nombreDeEval,
				Pregunta: "¿qué dice el art. 21 de la Ley 39/2015?",
				Activa:   true,
				Comandos: []ComandoEsperado{{Applet: "boe", Norma: "BOE-A-2015-10565", Bloque: "a21"}},
				Citas:    []CitaEsperada{{Norma: "BOE-A-2015-10565", Bloque: "a21"}},
			},
		},
		{
			nombre: "comandos-de-consulta-de-norma",
			documento: "pregunta: \"¿Qué servicios presta un municipio de más de 5000 habitantes?\"\n" +
				"activa: true\n" +
				"comandos:\n" +
				"  - applet: boe\n    verbo: indice\n    norma: BOE-A-1985-5392\n" +
				"  - applet: boe\n    verbo: metadatos\n    norma: BOE-A-1985-5392\n" +
				"  - applet: boe\n    verbo: analisis\n    norma: BOE-A-1985-5392\n" +
				"citas:\n  - norma: BOE-A-1985-5392\n    bloque: a85bis.\n",
			leida: Eval{
				Fichero:  nombreDeEval,
				Pregunta: "¿Qué servicios presta un municipio de más de 5000 habitantes?",
				Activa:   true,
				Comandos: []ComandoEsperado{
					{Applet: "boe", Verbo: "indice", Norma: "BOE-A-1985-5392"},
					{Applet: "boe", Verbo: "metadatos", Norma: "BOE-A-1985-5392"},
					{Applet: "boe", Verbo: "analisis", Norma: "BOE-A-1985-5392"},
				},
				Citas: citaDeLaLRBRL,
			},
		},
		{
			nombre: "comando-de-busqueda-con-reproduce",
			documento: "pregunta: \"¿Qué ley regula las bases del régimen local?\"\n" +
				"activa: true\n" +
				"reproduce: boe-fiscal\n" +
				"comandos:\n  - applet: boe\n    verbo: buscar\n    terminos: [bases, régimen, local]\n" +
				"citas:\n  - norma: BOE-A-1985-5392\n    bloque: a85bis.\n",
			leida: Eval{
				Fichero:   nombreDeEval,
				Pregunta:  "¿Qué ley regula las bases del régimen local?",
				Activa:    true,
				Reproduce: "boe-fiscal",
				Comandos:  []ComandoEsperado{{Applet: "boe", Verbo: "buscar", Terminos: []string{"bases", "régimen", "local"}}},
				Citas:     citaDeLaLRBRL,
			},
		},
		{
			nombre: "no-activa",
			documento: "# Eval de no activación: pregunta ajena a la normativa del BOE.\n" +
				"pregunta: \"¿Cómo invierto una lista enlazada en Go?\"\n" +
				"activa: false\n",
			leida: Eval{Fichero: nombreDeEval, Pregunta: "¿Cómo invierto una lista enlazada en Go?"},
		},
		{
			nombre:    "sin-pregunta",
			documento: "activa: false\n",
			error:     nombreDeEval + ": línea 1: missing property 'pregunta'",
		},
		{
			nombre:    "sin-activa",
			documento: preguntaDelArticulo21,
			error:     nombreDeEval + ": línea 1: missing property 'activa'",
		},
		{
			nombre:    "positiva-sin-citas",
			documento: preguntaDelArticulo21 + "activa: true\n" + comandoDelArticulo21,
			error:     nombreDeEval + ": línea 1: missing property 'citas'",
		},
		{
			nombre:    "no-activa-con-comandos",
			documento: preguntaDelArticulo21 + "activa: false\n" + comandoDelArticulo21,
			error:     nombreDeEval + ": línea 1: 'not' failed",
		},
		{
			nombre: "comando-con-verbo-y-bloque",
			documento: preguntaDelArticulo21 + "activa: true\n" +
				"comandos:\n" +
				"  - applet: boe\n" +
				"    verbo: indice\n" +
				"    norma: BOE-A-2015-10565\n" +
				"    bloque: a21\n" +
				citaDelArticulo21,
			fragmentos: []string{
				"comandos/0, línea 4: additional properties 'verbo' not allowed",
				"comandos/0, línea 4: additional properties 'bloque' not allowed",
			},
		},
		{
			nombre: "cita-sin-identificador-valido",
			documento: preguntaDelArticulo21 + "activa: true\n" + comandoDelArticulo21 +
				"citas:\n  - norma: BOE-A-15-1\n    bloque: a21\n",
			error: nombreDeEval + ": citas/0/norma, línea 8: " +
				"'BOE-A-15-1' does not match pattern '^BOE-A-[0-9]{4}-[0-9]{1,9}$'",
		},
		{
			nombre: "bloque-con-punto-inicial",
			documento: preguntaDelArticulo21 + "activa: true\n" + comandoDelArticulo21 +
				"citas:\n  - norma: BOE-A-2015-10565\n    bloque: .a1\n",
			error: nombreDeEval + ": citas/0/bloque, línea 9: " +
				"'.a1' does not match pattern '^[A-Za-z0-9][A-Za-z0-9.-]{0,63}$'",
		},
		{
			nombre:    "clave-desconocida",
			documento: preguntaDelArticulo21 + "activa: false\nmateria: programación\n",
			error:     nombreDeEval + ": línea 1: additional properties 'materia' not allowed",
		},
		{
			nombre:    "activa-repetida",
			documento: preguntaDelArticulo21 + "activa: true\nactiva: false\n",
			error:     nombreDeEval + ": activa repetido en las líneas 2 y 3",
		},
		{
			nombre: "bloque-repetido-en-una-cita",
			documento: preguntaDelArticulo21 + "activa: true\n" + comandoDelArticulo21 +
				"citas:\n  - norma: BOE-A-2015-10565\n    bloque: a21\n    bloque: a22\n",
			error: nombreDeEval + ": citas/0: bloque repetido en las líneas 9 y 10",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			leida, err := LeerEval(nombreDeEval, []byte(caso.documento))
			if caso.error == "" && len(caso.fragmentos) == 0 {
				require.NoError(t, err)
				assert.Equal(t, caso.leida, leida)

				return
			}

			require.Error(t, err)
			assert.Zero(t, leida, "una eval mal formada no se entrega a medias")
			assert.True(t, strings.HasPrefix(err.Error(), nombreDeEval+": "),
				"el error empieza por el nombre del fichero: %q", err.Error())

			if caso.error != "" {
				require.EqualError(t, err, caso.error)
			}

			for _, fragmento := range caso.fragmentos {
				assert.ErrorContains(t, err, fragmento)
			}
		})
	}
}

// TestEsquemaDeEval comprueba que el esquema publicado del formato común de eval
// compila con las aserciones de formato activas. Lo que dicen sus patrones de
// norma y de bloque lo fija TestGramaticasCoincidenConBoe.
func TestEsquemaDeEval(t *testing.T) {
	t.Parallel()

	esquema, err := esquemaDeEval()
	require.NoError(t, err)
	assert.NotNil(t, esquema)
}

// TestCompilarEsquemaDeEvalQueNoSirve fija los dos errores con los que no se
// obtiene el esquema del formato de eval, sobre la ruta de un directorio temporal
// del test: una carpeta donde se espera el fichero del esquema y un fichero JSON
// que no es un JSON Schema válido. Los dos nombran la ruta.
func TestCompilarEsquemaDeEvalQueNoSirve(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string

		// contenido es el del fichero del esquema; vacío, en su lugar hay una
		// carpeta.
		contenido string

		motivo string
	}{
		{nombre: "carpeta", motivo: "no se puede leer el esquema del formato de eval"},
		{
			nombre:    "tipo-que-no-es-de-json-schema",
			contenido: `{"$schema": "https://json-schema.org/draft/2020-12/schema", "type": "entero"}`,
			motivo:    "el esquema no compila",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			ruta := filepath.Join(t.TempDir(), "eval.yaml.json")
			if caso.contenido == "" {
				require.NoError(t, os.Mkdir(ruta, 0o750))
			} else {
				require.NoError(t, os.WriteFile(ruta, []byte(caso.contenido), 0o600))
			}

			esquema, err := compilarEsquemaDeEval(ruta)

			require.ErrorContains(t, err, ruta)
			require.ErrorContains(t, err, caso.motivo)
			assert.Nil(t, esquema)
		})
	}
}

// casoDeGramatica es un valor límite de una gramática de identificador y si
// boe lo acepta (data-model, cabecera).
type casoDeGramatica struct {
	nombre string
	valor  string
	valido bool
}

// casosDeNorma son los casos límite de NORMA, ^BOE-A-[0-9]{4}-[0-9]{1,9}$.
var casosDeNorma = []casoDeGramatica{
	{nombre: "lpac", valor: "BOE-A-2015-10565", valido: true},
	{nombre: "número-de-una-cifra", valor: "BOE-A-2015-1", valido: true},
	{nombre: "número-de-nueve-cifras", valor: "BOE-A-2015-123456789", valido: true},
	{nombre: "número-de-diez-cifras", valor: "BOE-A-2015-1234567890"},
	{nombre: "año-de-dos-cifras", valor: "BOE-A-15-1"},
	{nombre: "año-de-tres-cifras", valor: "BOE-A-201-10565"},
	{nombre: "año-de-cinco-cifras", valor: "BOE-A-20150-10565"},
	{nombre: "sin-número", valor: "BOE-A-2015-"},
	{nombre: "vacía", valor: ""},
	{nombre: "minúsculas", valor: "boe-a-2015-10565"},
	{nombre: "otra-serie", valor: "BOE-B-2015-10565"},
	{nombre: "espacio-inicial", valor: " BOE-A-2015-10565"},
	{nombre: "salto-de-línea-final", valor: "BOE-A-2015-10565\n"},
	{nombre: "cifras-de-otra-escritura", valor: "BOE-A-２０１５-10565"},
}

// casosDeBloque son los casos límite de BLOQUE, ^[A-Za-z0-9][A-Za-z0-9.-]{0,63}$.
var casosDeBloque = []casoDeGramatica{
	{nombre: "artículo", valor: "a21", valido: true},
	{nombre: "punto-final", valor: "a85bis.", valido: true},
	{nombre: "guion", valor: "a1-30", valido: true},
	{nombre: "preámbulo", valor: "preambulo", valido: true},
	{nombre: "mayúscula-inicial", valor: "A21", valido: true},
	{nombre: "una-cifra", valor: "1", valido: true},
	{nombre: "sesenta-y-cuatro-caracteres", valor: "a" + strings.Repeat("1", 63), valido: true},
	{nombre: "sesenta-y-cinco-caracteres", valor: "a" + strings.Repeat("1", 64)},
	{nombre: "vacío", valor: ""},
	{nombre: "punto-inicial", valor: ".a1"},
	{nombre: "guion-inicial", valor: "-a1"},
	{nombre: "dos-puntos", valor: ".."},
	{nombre: "barra", valor: "a/1"},
	{nombre: "espacio", valor: "a 1"},
	{nombre: "porcentaje", valor: "a%201"},
	{nombre: "salto-de-línea-final", valor: "a21\n"},
	{nombre: "letra-no-ascii", valor: "á21"},
	{nombre: "guion-bajo", valor: "a_1"},
}

// ubicacionDeGramatica es un sitio del formato de eval donde va un valor de la
// gramática: el documento de una eval positiva con el valor ahí y valores
// válidos en todo lo demás.
type ubicacionDeGramatica struct {
	nombre    string
	documento func(valor string) map[string]any
}

// evalPositiva es una eval positiva con un único comando y una única cita.
func evalPositiva(comando, cita map[string]any) map[string]any {
	return map[string]any{
		"pregunta": "¿qué dice el art. 21 de la Ley 39/2015?",
		"activa":   true,
		"comandos": []any{comando},
		"citas":    []any{cita},
	}
}

// TestGramaticasCoincidenConBoe comprueba que los patrones de NORMA y BLOQUE
// escritos en los esquemas aceptan y rechazan exactamente lo mismo que
// boe.ValidarNorma y boe.ValidarBloque sobre los casos límite de data-model
// (cabecera; control 10 del plan). En eval.yaml.json lo comprueba en cada sitio
// del formato donde va el valor —el comando de bloque, el de consulta de norma y
// la cita—, con una eval que solo puede fallar por ese valor: la misma eval con
// un valor válido se lee sin error. En normas.yaml.json lo comprueba en el único
// sitio donde va una norma, el nombre de cada entrada de normas, con una tabla
// que solo puede fallar por ese nombre, y exige que cada rechazo sea el defecto
// «identificador con otra forma» de esa entrada: el del patrón de
// propertyNames, y no otro.
func TestGramaticasCoincidenConBoe(t *testing.T) {
	t.Parallel()

	t.Run("eval.yaml.json", func(t *testing.T) {
		t.Parallel()

		const normaValida, bloqueValido = "BOE-A-2015-10565", "a21"

		comandoDeBloque := func(norma, bloque string) map[string]any {
			return map[string]any{"applet": "boe", "norma": norma, "bloque": bloque}
		}
		cita := func(norma, bloque string) map[string]any {
			return map[string]any{"norma": norma, "bloque": bloque}
		}

		gramaticas := []struct {
			nombre      string
			casos       []casoDeGramatica
			valida      func(string) error
			valorValido string
			ubicaciones []ubicacionDeGramatica
		}{
			{
				nombre:      "norma",
				casos:       casosDeNorma,
				valida:      boe.ValidarNorma,
				valorValido: normaValida,
				ubicaciones: []ubicacionDeGramatica{
					{nombre: "comando de bloque", documento: func(valor string) map[string]any {
						return evalPositiva(comandoDeBloque(valor, bloqueValido), cita(normaValida, bloqueValido))
					}},
					{nombre: "comando de consulta de norma", documento: func(valor string) map[string]any {
						return evalPositiva(map[string]any{"applet": "boe", "verbo": "metadatos", "norma": valor},
							cita(normaValida, bloqueValido))
					}},
					{nombre: "cita", documento: func(valor string) map[string]any {
						return evalPositiva(comandoDeBloque(normaValida, bloqueValido), cita(valor, bloqueValido))
					}},
				},
			},
			{
				nombre:      "bloque",
				casos:       casosDeBloque,
				valida:      boe.ValidarBloque,
				valorValido: bloqueValido,
				ubicaciones: []ubicacionDeGramatica{
					{nombre: "comando de bloque", documento: func(valor string) map[string]any {
						return evalPositiva(comandoDeBloque(normaValida, valor), cita(normaValida, bloqueValido))
					}},
					{nombre: "cita", documento: func(valor string) map[string]any {
						return evalPositiva(comandoDeBloque(normaValida, bloqueValido), cita(normaValida, valor))
					}},
				},
			},
		}

		for _, gramatica := range gramaticas {
			t.Run(gramatica.nombre, func(t *testing.T) {
				t.Parallel()

				for _, ubicacion := range gramatica.ubicaciones {
					require.True(t, aceptaLaEval(t, ubicacion.documento(gramatica.valorValido)),
						"con %q en %s la eval es válida: sin eso, un rechazo no diría nada del patrón",
						gramatica.valorValido, ubicacion.nombre)
				}

				for _, caso := range gramatica.casos {
					t.Run(caso.nombre, func(t *testing.T) {
						t.Parallel()

						aceptaBoe := gramatica.valida(caso.valor) == nil
						require.Equal(t, caso.valido, aceptaBoe,
							"la tabla dice de %q lo mismo que boe", caso.valor)

						for _, ubicacion := range gramatica.ubicaciones {
							assert.Equal(t, aceptaBoe, aceptaLaEval(t, ubicacion.documento(caso.valor)),
								"los patrones de %s del esquema en %s aceptan %q si y solo si boe lo acepta",
								gramatica.nombre, ubicacion.nombre, caso.valor)
						}
					})
				}
			})
		}
	})

	t.Run("normas.yaml.json", func(t *testing.T) {
		t.Parallel()

		const normaValida = "BOE-A-2015-10565"

		require.NoError(t, leerNormasConIdentificador(t, normaValida),
			"con %q la tabla de normas es válida: sin eso, un rechazo no diría nada del patrón", normaValida)

		for _, caso := range casosDeNorma {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				aceptaBoe := boe.ValidarNorma(caso.valor) == nil
				require.Equal(t, caso.valido, aceptaBoe, "la tabla dice de %q lo mismo que boe", caso.valor)

				err := leerNormasConIdentificador(t, caso.valor)
				if aceptaBoe {
					require.NoError(t, err, "el patrón de propertyNames de normas.yaml.json acepta %q, como boe", caso.valor)

					return
				}

				var defecto *skills.DefectoDeNorma
				require.ErrorAs(t, err, &defecto,
					"el patrón de propertyNames de normas.yaml.json rechaza %q, como boe", caso.valor)
				assert.Equal(t, &skills.DefectoDeNorma{Norma: caso.valor, Defecto: "identificador con otra forma"}, defecto,
					"el rechazo de %q es el del patrón de propertyNames", caso.valor)
			})
		}
	})
}

// leerNormasConIdentificador lee con skills.LeerNormas una tabla de normas,
// escrita en JSON, que es YAML válido y conserva cada texto tal cual, con una
// sola norma cuyo nombre es el identificador y que es válida en todo lo demás, y
// devuelve su error.
func leerNormasConIdentificador(t *testing.T, identificador string) error {
	t.Helper()

	contenido, err := json.Marshal(map[string]any{
		"normas": map[string]any{
			identificador: map[string]any{
				"titulo":   "Ley 39/2015, de 1 de octubre, del Procedimiento Administrativo Común de las Administraciones Públicas.",
				"rango":    "Ley",
				"materias": []any{"procedimiento administrativo"},
			},
		},
	})
	require.NoError(t, err)

	_, err = skills.LeerNormas(contenido)

	return err
}

// aceptaLaEval dice si LeerEval acepta el documento, escrito en JSON, que es
// YAML válido y conserva cada texto tal cual.
func aceptaLaEval(t *testing.T, documento map[string]any) bool {
	t.Helper()

	contenido, err := json.Marshal(documento)
	require.NoError(t, err)

	_, err = LeerEval(nombreDeEval, contenido)

	return err == nil
}
