package evals

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/app"
	"github.com/jmorenobl/kitlegal/internal/cache"
	"github.com/jmorenobl/kitlegal/internal/core/grafo"
	"github.com/jmorenobl/kitlegal/internal/graph"
	"github.com/jmorenobl/kitlegal/internal/skills"
	"github.com/jmorenobl/kitlegal/internal/source/boe"
)

// Contenidos de las evals sintéticas de TestLeerConjunto. La del art. 21 se
// escribe con los trozos de TestLeerEval.
const (
	contenidoDelArticulo21     = preguntaDelArticulo21 + "activa: true\n" + comandoDelArticulo21 + citaDelArticulo21
	contenidoDeProgramacion    = "pregunta: \"¿Cómo invierto una lista enlazada en Go?\"\nactiva: false\n"
	contenidoSinPregunta       = "activa: false\n"
	contenidoConActivaRepetida = preguntaDelArticulo21 + "activa: true\nactiva: false\n"

	// ficheroDeDosBloquesDeLaLCSP es el nombre con el que TestLeerConjunto lee
	// evalDeDosBloquesDeLaLCSP, el de la eval de contracts/evals-y-juicio.md §3
	// de H7.4.
	ficheroDeDosBloquesDeLaLCSP = "20-lcsp-dos-bloques-redaccion-cambiada.yaml"
)

// Los ficheros de la carpeta juez sintética de TestLeerConjunto
// (contracts/juez-y-voto.md §1 de H24). La declaración lleva dos clases: la que
// decide, con un umbral que no es el valor cero, y la que solo se publica, con
// el mayor que admite su esquema; y el esquema de la respuesta, una propiedad
// por clase.
const (
	claseQueDecide = "  - nombre: afirma_lo_no_leido\n" +
		"    decide: true\n" +
		"    umbral: 0.05\n"
	claseQueSePublica = "  - nombre: clase_2\n" +
		"    decide: false\n" +
		"    umbral: 1\n"
	clasesDelJuez = "# Las clases del juez de una skill sintética.\n" +
		"clases:\n" + claseQueDecide + claseQueSePublica
	rubricaDelJuez = "# Rúbrica\n\nResponde a las dos preguntas sobre la respuesta.\n"
	esquemaDelJuez = `{"type":"object","properties":{"afirma_lo_no_leido":{"type":"object"},"clase_2":{"type":"object"}}}` + "\n"
	casosDelJuez   = "clase: afirma_lo_no_leido\ncasos: []\n"
	medidaDelJuez  = "{\"clase\":\"afirma_lo_no_leido\"}\n"
)

// Los nombres, dentro de la carpeta de evals, de los cinco ficheros de la
// carpeta juez, que son los que llevan delante sus errores.
const (
	casosEnElJuez   = "juez/casos.yaml"
	clasesEnElJuez  = "juez/clases.yaml"
	esquemaEnElJuez = "juez/esquema.json"
	medidaEnElJuez  = "juez/medida.json"
	rubricaEnElJuez = "juez/rubrica.md"
)

// entradasDelJuez son las entradas de la carpeta juez de un conjunto sintético:
// sus cinco ficheros bien formados, en orden de nombre, salvo los de cambios,
// cada uno en lugar del que lleva su nombre.
func entradasDelJuez(t *testing.T, cambios ...entradaDeConjunto) []entradaDeConjunto {
	t.Helper()

	entradas := []entradaDeConjunto{
		{nombre: casosEnElJuez, contenido: casosDelJuez},
		{nombre: clasesEnElJuez, contenido: clasesDelJuez},
		{nombre: esquemaEnElJuez, contenido: esquemaDelJuez},
		{nombre: medidaEnElJuez, contenido: medidaDelJuez},
		{nombre: rubricaEnElJuez, contenido: rubricaDelJuez},
	}

	for _, cambio := range cambios {
		posicion := slices.IndexFunc(entradas, func(entrada entradaDeConjunto) bool { return entrada.nombre == cambio.nombre })
		require.GreaterOrEqual(t, posicion, 0, "%s es uno de los cinco ficheros de la carpeta juez", cambio.nombre)

		entradas[posicion] = cambio
	}

	return entradas
}

// entradaDeConjunto es una entrada que un test crea en el directorio de un
// conjunto de evals: un fichero con su contenido o, con carpeta, un
// subdirectorio vacío. Su nombre puede llevar delante la carpeta en la que va,
// que se crea con ella.
type entradaDeConjunto struct {
	nombre    string
	contenido string
	carpeta   bool
}

// malFormadoEsperado es un fichero mal formado que LeerConjunto tiene que
// devolver y un fragmento del motivo que su error tiene que decir.
type malFormadoEsperado struct {
	fichero   string
	fragmento string

	// exacto dice que el fragmento es el error entero.
	exacto bool
}

// TestLeerConjunto fija la lectura de un directorio de evals del contrato
// evals-y-grabaciones §1: todas sus entradas se leen antes de devolver, las
// evals bien formadas van a Evals y todo lo demás a MalFormados, en orden de
// nombre y cada una con un error que empieza por su nombre, sin que ninguna se
// salte ni sea el error de la lectura, que queda para el directorio que no se
// puede listar.
//
// Desde H7.2, fija también la lista de expresiones prohibidas de la carpeta
// (contrato lista-y-juicio §1; FR-050, FR-055): la entrada que se llama
// exactamente expresiones-prohibidas.yaml no es un fichero de eval; bien
// formada, queda en Conjunto.Prohibidas, con sus
// tres familias desde H7.3 (FR-023, FR-024); si no es un fichero regular o no
// valida —también la de dos familias de H7.2, sin anuncio—, es un fichero mal
// formado que la nombra y el conjunto queda sin lista; y una carpeta sin ella se
// lee como antes del hito, sin lista en el conjunto. Desde H24 la lista no va a
// ninguna eval: ya no juzga sus respuestas (data-model §5 de H24; FR-070).
//
// Desde H7.4 (contracts/lista-de-expresiones.md §1 y §2; FR-030, FR-031), la
// lista bien formada tiene sus cinco claves, y la familia redaccion_no_leida y
// las formas_fijas quedan también en Conjunto.Prohibidas; y cada lista mal
// formada lo está solo por su defecto. Desde H24
// (contracts/skill-boe-legislacion.md §4; FR-085), tiene seis: la clave
// salida_de_las_herramientas queda también en Conjunto.Prohibidas, y la lista de
// cinco claves de H7.4, sin ella, es un fichero mal formado. Y cada eval se lee
// con sus claves de H7.4 (contracts/evals-y-juicio.md §1 y §5; FR-003, FR-004,
// FR-053): las de plazos-y-territorio, con la skill que no se activa en
// NoSeActivan, y la positiva de los dos bloques de la LCSP, con sus redacciones
// modificadas en RedaccionesModificadas.
//
// Desde H24 (contracts/juez-y-voto.md §1; FR-020), fija también la carpeta del
// juez, con los casos de casosDeLaCarpetaJuez: la entrada que se llama
// exactamente juez no es un fichero de eval. Si es una carpeta bien formada,
// queda en Conjunto.Juez, con sus clases en su orden y la rúbrica, el esquema,
// los casos y la medida enteros. Cada fichero suyo que falta,
// que no se puede leer o que no tiene su forma —la declaración que incumple su
// esquema o repite un nombre, y el esquema cuyas propiedades no son exactamente
// las clases declaradas— es un fichero mal formado con su nombre,
// juez/<fichero>, y el conjunto se queda sin juez; y si juez no es una carpeta,
// el mal formado es ella. Un conjunto sin esa entrada no tiene juez: es lo que
// se exige de todos los casos anteriores.
func TestLeerConjunto(t *testing.T) {
	t.Parallel()

	leidaDelArticulo21 := Eval{
		Fichero:  nombreDeEval,
		Pregunta: "¿qué dice el art. 21 de la Ley 39/2015?",
		Activa:   true,
		Comandos: []ComandoEsperado{{Applet: "boe", Norma: "BOE-A-2015-10565", Bloque: "a21"}},
		Citas:    []CitaEsperada{{Norma: "BOE-A-2015-10565", Bloque: "a21"}},
	}
	leidaDeProgramacion := Eval{
		Fichero:  "02-no-activa-programacion.yaml",
		Pregunta: "¿Cómo invierto una lista enlazada en Go?",
	}
	sinLaForma := "no tiene la forma <nn>-<descripción>.yaml"

	// Lo que se lee de listaBienFormada, la del test del esquema con sus seis
	// claves; las dos claves de H7.4 bien formadas; y esas dos con la de H24, para
	// que la lista sin una familia de H7.3 lo esté solo por eso.
	lista := ExpresionesProhibidas{
		Maquinaria:              []string{"memoria de consultas", "hallazgos", "c\xc3\xb3digo de salida"},
		OtraConversacion:        []string{"te dije", "conversaci\xc3\xb3n anterior"},
		Anuncio:                 []string{"que trasladar", "ya puedo responder", "as\xc3\xad que respondo"},
		RedaccionNoLeida:        []string{"ya no exige", "se elimin\xc3\xb3"},
		SalidaDeLasHerramientas: []string{"el sobre", "fecha_vigencia"},
		FormasFijas: []string{
			"\xe2\x9a\xa0 REDACCI\xc3\x93N MODIFICADA: <cita>: vigente desde <fecha>.",
			"No se ha podido comprobar si la redacci\xc3\xb3n ha cambiado",
		},
	}
	clavesDeH74 := redaccionNoLeidaBienFormada + formasFijasBienFormadas
	demasClaves := clavesDeH74 + salidaDeLasHerramientasBienFormada

	casos := []casoDeLeerConjunto{
		{
			nombre: "bien-formadas",
			entradas: []entradaDeConjunto{
				{nombre: "02-no-activa-programacion.yaml", contenido: contenidoDeProgramacion},
				{nombre: nombreDeEval, contenido: contenidoDelArticulo21},
			},
			evals: []Eval{leidaDelArticulo21, leidaDeProgramacion},
		},
		{
			nombre: "con-mal-formadas",
			entradas: []entradaDeConjunto{
				{nombre: nombreDeEval, contenido: contenidoDelArticulo21},
				{nombre: "02-sin-pregunta.yaml", contenido: contenidoSinPregunta},
				{nombre: "03-activa-repetida.yaml", contenido: contenidoConActivaRepetida},
			},
			evals: []Eval{leidaDelArticulo21},
			malFormados: []malFormadoEsperado{
				{fichero: "02-sin-pregunta.yaml", fragmento: "missing property 'pregunta'"},
				{fichero: "03-activa-repetida.yaml", fragmento: "activa repetido en las líneas 2 y 3"},
			},
		},
		{
			nombre: "entradas-que-no-son-evals",
			entradas: []entradaDeConjunto{
				{nombre: nombreDeEval, contenido: contenidoDelArticulo21},
				{nombre: "02-subdirectorio.yaml", carpeta: true},
				{nombre: "notas.txt", contenido: "notas sueltas\n"},
			},
			evals: []Eval{leidaDelArticulo21},
			malFormados: []malFormadoEsperado{
				{fichero: "02-subdirectorio.yaml", fragmento: "no es un fichero regular"},
				{fichero: "notas.txt", fragmento: sinLaForma},
			},
		},
		{
			// Cada nombre incumple una parte de ^[0-9]{2}-[a-z0-9]+(-[a-z0-9]+)*\.yaml$
			// con el contenido de una eval válida: el nombre basta para que no lo sea.
			nombre: "nombres-sin-la-forma",
			entradas: []entradaDeConjunto{
				{nombre: "1-una-cifra.yaml", contenido: contenidoDeProgramacion},
				{nombre: "001-tres-cifras.yaml", contenido: contenidoDeProgramacion},
				{nombre: "01_guion-bajo.yaml", contenido: contenidoDeProgramacion},
				{nombre: "01-.yaml", contenido: contenidoDeProgramacion},
				{nombre: "01-Mayuscula.yaml", contenido: contenidoDeProgramacion},
				{nombre: "01-doble--guion.yaml", contenido: contenidoDeProgramacion},
				{nombre: "01-guion-final-.yaml", contenido: contenidoDeProgramacion},
				{nombre: "01-extension.yml", contenido: contenidoDeProgramacion},
				{nombre: "02-no-activa-programacion.yaml", contenido: contenidoDeProgramacion},
			},
			evals: []Eval{leidaDeProgramacion},
			malFormados: []malFormadoEsperado{
				{fichero: "001-tres-cifras.yaml", fragmento: sinLaForma},
				{fichero: "01-.yaml", fragmento: sinLaForma},
				{fichero: "01-Mayuscula.yaml", fragmento: sinLaForma},
				{fichero: "01-doble--guion.yaml", fragmento: sinLaForma},
				{fichero: "01-extension.yml", fragmento: sinLaForma},
				{fichero: "01-guion-final-.yaml", fragmento: sinLaForma},
				{fichero: "01_guion-bajo.yaml", fragmento: sinLaForma},
				{fichero: "1-una-cifra.yaml", fragmento: sinLaForma},
			},
		},
		{
			nombre: "con-lista",
			entradas: []entradaDeConjunto{
				{nombre: "02-no-activa-programacion.yaml", contenido: contenidoDeProgramacion},
				{nombre: ficheroDeExpresionesProhibidas, contenido: listaBienFormada},
				{nombre: nombreDeEval, contenido: contenidoDelArticulo21},
			},
			evals:      []Eval{leidaDelArticulo21, leidaDeProgramacion},
			prohibidas: lista,
		},
		{
			nombre: "lista-sin-una-familia",
			entradas: []entradaDeConjunto{
				{nombre: "02-no-activa-programacion.yaml", contenido: contenidoDeProgramacion},
				{
					nombre:    ficheroDeExpresionesProhibidas,
					contenido: maquinariaBienFormada + anuncioBienFormado + demasClaves,
				},
				{nombre: nombreDeEval, contenido: contenidoDelArticulo21},
			},
			evals: []Eval{leidaDelArticulo21, leidaDeProgramacion},
			malFormados: []malFormadoEsperado{
				{fichero: ficheroDeExpresionesProhibidas, fragmento: "missing property 'otra_conversacion'"},
			},
		},
		{
			// La lista de dos familias de H7.2, sin la de anuncio, está mal
			// formada: las evals se leen sin lista (FR-023).
			nombre: "lista-sin-anuncio",
			entradas: []entradaDeConjunto{
				{
					nombre:    ficheroDeExpresionesProhibidas,
					contenido: maquinariaBienFormada + otraConversacionBienFormada + demasClaves,
				},
				{nombre: nombreDeEval, contenido: contenidoDelArticulo21},
			},
			evals: []Eval{leidaDelArticulo21},
			malFormados: []malFormadoEsperado{
				{fichero: ficheroDeExpresionesProhibidas, fragmento: "missing property 'anuncio'"},
			},
		},
		{
			// La lista de cinco claves de H7.4, sin la de H24, está mal formada:
			// el conjunto queda sin lista (FR-085 de H24).
			nombre: "lista-sin-salida-de-las-herramientas",
			entradas: []entradaDeConjunto{
				{
					nombre: ficheroDeExpresionesProhibidas,
					contenido: maquinariaBienFormada + otraConversacionBienFormada + anuncioBienFormado +
						clavesDeH74,
				},
				{nombre: nombreDeEval, contenido: contenidoDelArticulo21},
			},
			evals: []Eval{leidaDelArticulo21},
			malFormados: []malFormadoEsperado{{
				fichero:   ficheroDeExpresionesProhibidas,
				fragmento: "missing property 'salida_de_las_herramientas'",
			}},
		},
		{
			nombre: "lista-con-una-familia-repetida",
			entradas: []entradaDeConjunto{
				{
					nombre:    ficheroDeExpresionesProhibidas,
					contenido: listaBienFormada + "maquinaria:\n  - json\n",
				},
				{nombre: nombreDeEval, contenido: contenidoDelArticulo21},
			},
			evals: []Eval{leidaDelArticulo21},
			malFormados: []malFormadoEsperado{
				{fichero: ficheroDeExpresionesProhibidas, fragmento: "maquinaria repetido en las l\xc3\xadneas 1 y 21"},
			},
		},
		{
			// Con su nombre, una carpeta no es un fichero de eval sin la forma de
			// nombre: es la lista, que no es un fichero regular.
			nombre: "lista-que-no-es-un-fichero",
			entradas: []entradaDeConjunto{
				{nombre: ficheroDeExpresionesProhibidas, carpeta: true},
				{nombre: nombreDeEval, contenido: contenidoDelArticulo21},
			},
			evals: []Eval{leidaDelArticulo21},
			malFormados: []malFormadoEsperado{{
				fichero:   ficheroDeExpresionesProhibidas,
				fragmento: ficheroDeExpresionesProhibidas + ": no es un fichero regular",
				exacto:    true,
			}},
		},
		{
			// Solo el nombre exacto es el de la lista: con otra extensión, es un
			// fichero de eval sin la forma de nombre.
			nombre: "lista-con-otro-nombre",
			entradas: []entradaDeConjunto{
				{nombre: "expresiones-prohibidas.yml", contenido: listaBienFormada},
				{nombre: nombreDeEval, contenido: contenidoDelArticulo21},
			},
			evals:       []Eval{leidaDelArticulo21},
			malFormados: []malFormadoEsperado{{fichero: "expresiones-prohibidas.yml", fragmento: sinLaForma}},
		},
		{
			// Las dos de plazos-y-territorio, cada una con la skill que no se activa
			// detrás de activa, como en el repositorio.
			nombre: "no-se-activan-en-plazos-y-territorio",
			entradas: []entradaDeConjunto{
				{
					nombre: plazosYTerritorioCubierto,
					contenido: preguntaDelMunicipio + "activa: true\n" + noSeActivaNormativa + comandoDelMunicipio +
						"territorio:\n  comunidad: Comunidad de Madrid\n",
				},
				{nombre: plazosYTerritorioNoActivacion, contenido: evalDeLaReceta + noSeActivaNormativa},
			},
			evals: []Eval{
				{
					Fichero:     plazosYTerritorioCubierto,
					Pregunta:    "¿En qué boletines se publican las normas que afectan a Leganés?",
					Activa:      true,
					NoSeActivan: []string{"normativa"},
					Comandos:    []ComandoEsperado{{Applet: "territorio", Verbo: "resolver", Municipio: "Leganés"}},
					Territorio:  TerritorioEsperado{Comunidad: "Comunidad de Madrid"},
				},
				{
					Fichero:     plazosYTerritorioNoActivacion,
					Pregunta:    "¿Cómo se hace una tortilla de patatas jugosa?",
					NoSeActivan: []string{"normativa"},
				},
			},
		},
		{
			nombre: "redacciones-modificadas-de-una-positiva",
			entradas: []entradaDeConjunto{
				{nombre: ficheroDeDosBloquesDeLaLCSP, contenido: evalDeDosBloquesDeLaLCSP},
				{nombre: nombreDeEval, contenido: contenidoDelArticulo21},
			},
			evals: []Eval{leidaDelArticulo21, leidaDeDosBloquesDeLaLCSP(ficheroDeDosBloquesDeLaLCSP)},
		},
	}

	casos = append(casos, casosDeLaCarpetaJuez(t,
		entradaDeConjunto{nombre: nombreDeEval, contenido: contenidoDelArticulo21}, leidaDelArticulo21)...)

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			dir := crearConjunto(t, caso.entradas)

			conjunto, err := LeerConjunto(dir)
			require.NoError(t, err, "un fichero mal formado nunca es el error de la lectura")
			assert.Equal(t, caso.evals, conjunto.Evals)
			assert.Equal(t, caso.prohibidas, conjunto.Prohibidas)
			assert.Equal(t, caso.juez, conjunto.Juez)

			ficheros := make([]string, 0, len(conjunto.MalFormados))
			for _, malFormado := range conjunto.MalFormados {
				ficheros = append(ficheros, malFormado.Fichero)
			}

			esperados := make([]string, 0, len(caso.malFormados))
			for _, esperado := range caso.malFormados {
				esperados = append(esperados, esperado.fichero)
			}

			require.Equal(t, esperados, ficheros, "cada entrada que no es una eval está, en orden de nombre")

			for indice, esperado := range caso.malFormados {
				motivo := conjunto.MalFormados[indice].Error
				require.ErrorContains(t, motivo, esperado.fragmento)
				assert.True(t, strings.HasPrefix(motivo.Error(), esperado.fichero+": "),
					"el error empieza por el nombre del fichero: %q", motivo.Error())

				if esperado.exacto {
					require.EqualError(t, motivo, esperado.fragmento)
				}
			}
		})
	}

	t.Run("directorio-inexistente", func(t *testing.T) {
		t.Parallel()

		dir := filepath.Join(t.TempDir(), "no-existe")

		conjunto, err := LeerConjunto(dir)
		require.ErrorIs(t, err, fs.ErrNotExist)
		require.ErrorContains(t, err, dir)
		assert.Zero(t, conjunto, "sin directorio no hay conjunto, ni a medias")
	})
}

// crearConjunto crea en un directorio temporal del test las entradas dadas y
// devuelve su ruta.
func crearConjunto(t *testing.T, entradas []entradaDeConjunto) string {
	t.Helper()

	dir := t.TempDir()
	crearEntradas(t, dir, entradas)

	return dir
}

// crearEntradas crea en el directorio las entradas dadas, cada una con la
// carpeta que lleve delante su nombre.
func crearEntradas(t *testing.T, dir string, entradas []entradaDeConjunto) {
	t.Helper()

	for _, entrada := range entradas {
		ruta := filepath.Join(dir, entrada.nombre)
		if entrada.carpeta {
			require.NoError(t, os.MkdirAll(ruta, 0o750))

			continue
		}

		require.NoError(t, os.MkdirAll(filepath.Dir(ruta), 0o750))
		require.NoError(t, os.WriteFile(ruta, []byte(entrada.contenido), 0o600))
	}
}

// escribirLaCarpetaDelJuez escribe en la carpeta de evals dada, que es de un
// directorio temporal del test, la carpeta juez de una skill sintética, con sus
// cinco ficheros bien formados (entradasDelJuez): con ella, la skill de esa
// carpeta tiene juez, y el informe, los umbrales de sus respuestas (research D13
// de H24).
func escribirLaCarpetaDelJuez(t *testing.T, evals string) {
	t.Helper()

	crearEntradas(t, evals, entradasDelJuez(t))

	conjunto, err := LeerConjunto(evals)
	require.NoError(t, err)
	require.NotNil(t, conjunto.Juez, "%s tiene la carpeta del juez, bien formada", evals)
}

// casoDeLeerConjunto es un caso de TestLeerConjunto: las entradas de un
// directorio de evals y lo que LeerConjunto tiene que leer de él.
type casoDeLeerConjunto struct {
	nombre      string
	entradas    []entradaDeConjunto
	evals       []Eval
	malFormados []malFormadoEsperado

	// prohibidas es la lista que tiene que quedar en el conjunto.
	prohibidas ExpresionesProhibidas

	// juez es el juez que tiene que quedar en el conjunto; nil, ninguno.
	juez *Juez
}

// casosDeLaCarpetaJuez son los casos de TestLeerConjunto de la carpeta del juez
// (contracts/juez-y-voto.md §1 de H24; FR-020). El conjunto de cada uno lleva
// además la eval dada, que se lee igual con el juez bien o mal formado: la
// carpeta bien formada; la declaración de clases que incumple su esquema, de
// cinco maneras —una clave de más, en la raíz y en una clase, una clase sin
// nombre, un nombre con mayúscula, un umbral de 2 y ninguna clase—, y la que
// repite un nombre; el esquema de la respuesta sin la propiedad de una clase,
// con una propiedad que no es de ninguna y que no es JSON; cada uno de los
// cinco ficheros que falta y que no se puede leer, con una carpeta en su lugar;
// y la entrada juez que no es una carpeta. Cada carpeta mal formada lo está
// solo por su defecto, así que da solo ese fichero.
func casosDeLaCarpetaJuez(t *testing.T, eval entradaDeConjunto, leida Eval) []casoDeLeerConjunto {
	t.Helper()

	conJuez := func(cambios ...entradaDeConjunto) []entradaDeConjunto {
		return append([]entradaDeConjunto{eval}, entradasDelJuez(t, cambios...)...)
	}
	conClases := func(clases string) []entradaDeConjunto {
		return conJuez(entradaDeConjunto{nombre: clasesEnElJuez, contenido: clases})
	}
	conEsquema := func(esquema string) []entradaDeConjunto {
		return conJuez(entradaDeConjunto{nombre: esquemaEnElJuez, contenido: esquema})
	}
	malFormado := func(nombre string, entradas []entradaDeConjunto, fichero, fragmento string) casoDeLeerConjunto {
		return casoDeLeerConjunto{
			nombre:      nombre,
			entradas:    entradas,
			evals:       []Eval{leida},
			malFormados: []malFormadoEsperado{{fichero: fichero, fragmento: fragmento}},
		}
	}

	const (
		sinNombre      = "  - decide: true\n    umbral: 0.05\n"
		conMayuscula   = "  - nombre: Afirma_lo_no_leido\n    decide: true\n    umbral: 0.05\n"
		conUmbralDe2   = "  - nombre: afirma_lo_no_leido\n    decide: true\n    umbral: 2\n"
		noSonLasClases = "no son exactamente las clases declaradas (afirma_lo_no_leido, clase_2)"
	)

	casos := []casoDeLeerConjunto{
		{
			nombre:   "juez-bien-formado",
			entradas: conJuez(),
			evals:    []Eval{leida},
			juez: &Juez{
				Clases: []ClaseDelJuez{
					{Nombre: "afirma_lo_no_leido", Decide: true, Umbral: 0.05},
					{Nombre: "clase_2", Decide: false, Umbral: 1},
				},
				Rubrica: rubricaDelJuez,
				Esquema: esquemaDelJuez,
				Casos:   casosDelJuez,
				Medida:  medidaDelJuez,
			},
		},
		malFormado("juez-con-una-clave-de-mas", conClases(clasesDelJuez+"rubrica: rubrica.md\n"),
			clasesEnElJuez, "additional properties 'rubrica' not allowed"),
		malFormado("juez-con-una-clave-de-mas-en-una-clase",
			conClases("clases:\n"+claseQueDecide+"    descripcion: inventa\n"+claseQueSePublica),
			clasesEnElJuez, "additional properties 'descripcion' not allowed"),
		malFormado("juez-con-una-clase-sin-nombre", conClases("clases:\n"+sinNombre+claseQueSePublica),
			clasesEnElJuez, "missing property 'nombre'"),
		malFormado("juez-con-mayuscula-en-un-nombre", conClases("clases:\n"+conMayuscula+claseQueSePublica),
			clasesEnElJuez, "'Afirma_lo_no_leido' does not match pattern"),
		malFormado("juez-con-un-umbral-de-2", conClases("clases:\n"+conUmbralDe2+claseQueSePublica),
			clasesEnElJuez, "maximum: got 2, want 1"),
		malFormado("juez-sin-ninguna-clase", conClases("clases: []\n"),
			clasesEnElJuez, "minItems: got 0, want 1"),
		malFormado("juez-con-una-clase-repetida", conClases(clasesDelJuez+claseQueDecide),
			clasesEnElJuez, clasesEnElJuez+": la clase afirma_lo_no_leido está repetida"),
		malFormado("juez-con-un-esquema-sin-una-clase",
			conEsquema(`{"type":"object","properties":{"afirma_lo_no_leido":{"type":"object"}}}`),
			esquemaEnElJuez, esquemaEnElJuez+": sus propiedades (afirma_lo_no_leido) "+noSonLasClases),
		malFormado("juez-con-un-esquema-con-otra-propiedad",
			conEsquema(`{"properties":{"afirma_lo_no_leido":{},"clase_2":{},"clase_3":{}}}`),
			esquemaEnElJuez, esquemaEnElJuez+": sus propiedades (afirma_lo_no_leido, clase_2, clase_3) "+noSonLasClases),
		malFormado("juez-con-un-esquema-que-no-es-json", conEsquema("no es JSON\n"),
			esquemaEnElJuez, esquemaEnElJuez+": no se puede leer como JSON: "),
		{
			// Con su nombre, un fichero no es una eval sin la forma de nombre: es
			// la carpeta del juez, que no es un directorio.
			nombre:   "juez-que-no-es-una-carpeta",
			entradas: []entradaDeConjunto{eval, {nombre: "juez", contenido: clasesDelJuez}},
			evals:    []Eval{leida},
			malFormados: []malFormadoEsperado{{
				fichero:   "juez",
				fragmento: "juez: no es un directorio",
				exacto:    true,
			}},
		},
	}

	for _, fichero := range []string{casosEnElJuez, clasesEnElJuez, esquemaEnElJuez, medidaEnElJuez, rubricaEnElJuez} {
		sinElFichero := slices.DeleteFunc(conJuez(), func(entrada entradaDeConjunto) bool { return entrada.nombre == fichero })

		casos = append(casos,
			malFormado("juez-sin-"+filepath.Base(fichero), sinElFichero, fichero, fichero+": no se puede leer: "),
			malFormado("juez-con-una-carpeta-por-"+filepath.Base(fichero),
				conJuez(entradaDeConjunto{nombre: fichero, carpeta: true}), fichero, fichero+": no se puede leer: "))
	}

	return casos
}

// Identificadores de las normas sintéticas de TestConjuntoDeEvals. Salvo el de
// la LPAC, que fija la regla del art. 21, son inventados —ninguna norma es del
// año 0000—: las reglas del conjunto solo los comparan entre sí y con las normas
// conocidas.
const (
	normaLPAC           = "BOE-A-2015-10565"
	normaLCSP           = "BOE-A-0000-2"
	normaLRBRL          = "BOE-A-0000-3"
	normaLGT            = "BOE-A-0000-4"
	normaTRLRHL         = "BOE-A-0000-5"
	normaIRPF           = "BOE-A-0000-6"
	normaLRJSP          = "BOE-A-0000-7"
	normaLTAIBG         = "BOE-A-0000-8"
	normaConstitucion   = "BOE-A-0000-9"
	normaET             = "BOE-A-0000-10"
	normaComun          = "BOE-A-0000-11"
	normaSinAbreviatura = "BOE-A-0000-12"
	normaModificadora   = "BOE-A-0000-13"
	normaDesconocida    = "BOE-A-0000-99"
)

// evalConFecha es el fichero de la eval con fecha del conjunto sintético de
// TestConjuntoDeEvals (contracts/evals-redaccion-fechada.md §1.4 de H20).
const evalConFecha = "15-lcsp-contrato-menor-en-2019.yaml"

// conjuntoSintetico son las evals y las normas conocidas con las que
// TestConjuntoDeEvals llama a ComprobarConjunto con las reglas de
// normativa.
type conjuntoSintetico struct {
	evals  []Eval
	normas map[string]NormaConocida
}

// conjuntoQueCumple devuelve, nuevo en cada llamada, un conjunto que cumple
// todas las reglas de data-model §6.3: el de evals/normativa/ del contrato
// de evals §2, con diez positivas de una norma cada una y dos de no activación,
// la informativa por materia, desde H21, la eval sin binario ni servidor
// (contracts/evals-en-dos-modos.md §1 de H21) y, desde H20, una eval con fecha
// (contracts/evals-redaccion-fechada.md §1.4 de H20). La 07 y la 08 citan además
// una norma común, porque una positiva que cita dos normas cuenta como materia
// distinta si al menos una no la cita ninguna otra (spec, casos límite).
func conjuntoQueCumple(t *testing.T) conjuntoSintetico {
	t.Helper()

	positivas := []struct{ fichero, pregunta, norma, bloque string }{
		{"01-lpac-articulo-21.yaml", "¿qué dice el art. 21 de la Ley 39/2015?", normaLPAC, "a21"},
		{"02-lcsp-contrato-menor.yaml", "¿Qué debe incluir el expediente de un contrato menor?", normaLCSP, "a1-30"},
		{"03-lrbrl-atribuciones-del-pleno.yaml", "¿Qué atribuciones tiene el Pleno?", normaLRBRL, "a22"},
		{"04-lgt-prescripcion.yaml", "¿Qué plazo de prescripción tiene una deuda tributaria?", normaLGT, "a66"},
		{"05-trlrhl-impuestos-municipales.yaml", "¿Qué impuestos exigen los ayuntamientos?", normaTRLRHL, "a59"},
		{"06-irpf-rendimientos-del-trabajo.yaml", "¿Qué rendimientos del trabajo grava el IRPF?", normaIRPF, "a17"},
		{"07-lrjsp-principio-de-legalidad.yaml", "¿Qué dice la Ley 40/2015 de la legalidad?", normaLRJSP, "a25"},
		{"08-ltaibg-plazo-de-resolucion.yaml", "¿En qué plazo se resuelve una solicitud de acceso?", normaLTAIBG, "a20"},
		{"09-constitucion-articulo-140.yaml", "¿Qué dice el artículo 140 de la Constitución?", normaConstitucion, "a140"},
		{"10-et-vacaciones.yaml", "¿Qué vacaciones reconoce el Estatuto de los Trabajadores?", normaET, "a38"},
	}

	conjunto := conjuntoSintetico{
		evals: make([]Eval, 0, len(positivas)+5),
		normas: map[string]NormaConocida{
			normaLPAC:         {Abreviatura: "LPAC", Materias: []string{"procedimiento administrativo"}},
			normaLCSP:         {Abreviatura: "LCSP", Materias: []string{"contratación pública"}},
			normaLRBRL:        {Abreviatura: "LRBRL", Materias: []string{"régimen local"}},
			normaLGT:          {Abreviatura: "LGT", Materias: []string{"tributos"}},
			normaTRLRHL:       {Abreviatura: "TRLRHL", Materias: []string{"haciendas locales", "tributos"}},
			normaIRPF:         {Abreviatura: "LIRPF", Materias: []string{"tributos"}},
			normaLRJSP:        {Abreviatura: "LRJSP", Materias: []string{"sector público"}},
			normaLTAIBG:       {Abreviatura: "LTAIBG", Materias: []string{"transparencia"}},
			normaConstitucion: {Materias: []string{"constitución"}},
			normaET:           {Abreviatura: "ET", Materias: []string{"trabajo"}},
			normaComun:        {Materias: []string{"sector público", "transparencia"}},
		},
	}

	for _, positiva := range positivas {
		conjunto.evals = append(conjunto.evals, Eval{
			Fichero:  positiva.fichero,
			Pregunta: positiva.pregunta,
			Activa:   true,
			Comandos: []ComandoEsperado{{Applet: "boe", Norma: positiva.norma, Bloque: positiva.bloque}},
			Citas:    []CitaEsperada{{Norma: positiva.norma, Bloque: positiva.bloque}},
		})
	}

	conjunto.evals = append(conjunto.evals,
		Eval{Fichero: "11-no-activa-programacion.yaml", Pregunta: "¿Cómo invierto una lista enlazada en Go?"},
		Eval{Fichero: "12-no-activa-acuerdo-entre-amigos.yaml", Pregunta: "Reescribe en un tono cercano esta frase."},
		// La informativa por materia: no cuenta entre las positivas ni en la regla
		// de materias, así que repite la norma de la 03 (ADR 0016).
		Eval{
			Fichero: "13-lrbrl-atribuciones-por-materia.yaml", Pregunta: "¿Qué atribuciones tiene el Pleno?",
			Activa: true, Informativa: true,
			Comandos: []ComandoEsperado{{Applet: "boe", Norma: normaLRBRL, Bloque: "a22"}},
			Citas:    []CitaEsperada{{Norma: normaLRBRL, Bloque: "a22"}},
		},
		// La eval sin binario ni servidor: activa la skill sin comandos ni citas, y
		// no cuenta entre las positivas ni en la regla de materias (FR-046 de H21).
		Eval{
			Fichero: "14-sin-binario-ni-servidor.yaml", Pregunta: "\xc2\xbfQu\xc3\xa9 dice el art. 53 de la Ley 39/2015?",
			Activa: true, SinBinarioNiServidor: true,
		},
		// La eval con fecha: activa la skill, decide y repite la norma y el bloque
		// de la 02, y no cuenta entre las positivas ni en la regla de materias; la
		// norma de su redacción de hoy no es ninguna conocida, porque las normas
		// conocidas no miran redaccion_fechada (research D15 de H20).
		Eval{
			Fichero:  evalConFecha,
			Pregunta: "\xc2\xbfQu\xc3\xa9 exig\xc3\xada el expediente de un contrato menor el 1 de junio de 2019?",
			Activa:   true,
			Comandos: []ComandoEsperado{{Applet: "boe", Norma: normaLCSP, Bloque: "a1-30", Fecha: "2019-06-01"}},
			Citas:    []CitaEsperada{{Norma: normaLCSP, Bloque: "a1-30"}},
			RedaccionFechada: &RedaccionFechadaEsperada{
				Norma: normaLCSP, Bloque: "a1-30", Fecha: "2019-06-01", FechaVigencia: "20180309",
				Posterior: &RedaccionPosteriorEsperada{FechaVigencia: "20200206", NormaModificadora: normaModificadora},
			},
		},
	)

	evalDe(t, conjunto.evals, "06-irpf-rendimientos-del-trabajo.yaml").Reproduce = "boe-fiscal"

	for _, fichero := range []string{"07-lrjsp-principio-de-legalidad.yaml", "08-ltaibg-plazo-de-resolucion.yaml"} {
		eval := evalDe(t, conjunto.evals, fichero)
		eval.Citas = append(eval.Citas, CitaEsperada{Norma: normaComun, Bloque: "a3"})
	}

	return conjunto
}

// evalDe es la eval del conjunto con ese fichero, para modificarla.
func evalDe(t *testing.T, evals []Eval, fichero string) *Eval {
	t.Helper()

	indice := slices.IndexFunc(evals, func(eval Eval) bool { return eval.Fichero == fichero })
	require.GreaterOrEqual(t, indice, 0, "el conjunto sintético tiene %s", fichero)

	return &evals[indice]
}

// sinEvals quita del conjunto las evals con esos ficheros, que tienen que estar.
func sinEvals(t *testing.T, conjunto *conjuntoSintetico, ficheros ...string) {
	t.Helper()

	for _, fichero := range ficheros {
		evalDe(t, conjunto.evals, fichero)
	}

	conjunto.evals = slices.DeleteFunc(conjunto.evals, func(eval Eval) bool {
		return slices.Contains(ficheros, eval.Fichero)
	})
}

// sinMateria quita la materia de todas las normas conocidas del conjunto.
func sinMateria(conjunto *conjuntoSintetico, materia string) {
	for identificador, norma := range conjunto.normas {
		norma.Materias = slices.DeleteFunc(norma.Materias, func(otra string) bool { return otra == materia })
		conjunto.normas[identificador] = norma
	}
}

// TestConjuntoDeEvals fija ComprobarConjunto con sus dos juegos de reglas
// (contrato de evals §3 de H6; research D22): las de normativa y las de
// plazos-y-territorio, cada juego sobre evals sintéticas. Desde H21, el tamaño
// de normativa llega a 21 y los dos juegos llevan la regla sin binario ni
// servidor, la última de cada tabla: exactamente una eval que lo declara, que no
// cuenta como positiva ni necesita un esperado verificable
// (contracts/evals-en-dos-modos.md §1 de H21; FR-046). Desde H23, el tercer
// juego, el de jurisprudencia (contracts/evals-jurisprudencia.md §5 de H23;
// FR-056), que desde H25 cuenta diez evals (contracts/evals-jurisprudencia.md
// §3 de H25; FR-065). Desde H20, el tamaño de normativa llega a 23, y una
// eval con fecha se aparta como la eval sin binario ni servidor: no es positiva
// (contracts/evals-redaccion-fechada.md §1.4 de H20; research D15 de H20).
func TestConjuntoDeEvals(t *testing.T) {
	t.Parallel()

	t.Run("normativa", probarReglasDeNormativa)
	t.Run("plazos-y-territorio", probarReglasDePlazosYTerritorio)
	t.Run("jurisprudencia", probarReglasDeJurisprudencia)
}

// probarReglasDeNormativa fija ComprobarConjunto con ReglasDeNormativa
// (contrato de evals §2 de H5): el conjunto que cumple todas las reglas de
// data-model §6.3 no da ningún defecto; cada copia que incumple solo una regla da
// exactamente un defecto, con el nombre de esa regla y un mensaje que nombra los
// ficheros y las normas implicados; y un conjunto que las incumple todas da un
// defecto por regla, en el orden de la tabla. Desde H21, el conjunto que cumple
// lleva la eval sin binario ni servidor sin que cuente como positiva; y sin
// ninguna eval sin binario ni servidor, o con dos, incumple solo la regla nueva
// (contracts/evals-en-dos-modos.md §1 de H21). Desde H20, lleva además una eval
// con fecha, que se aparta como la eval sin binario ni servidor: no cuenta como
// positiva ni en la regla de materias, aunque repite la norma de la 02, y la
// norma de su redacción de hoy no se mira; con una norma desconocida en su
// comando o en su cita incumple solo la regla de las normas conocidas; y con 23
// evals el conjunto sigue cumpliendo y con 24 incumple el tamaño
// (contracts/evals-redaccion-fechada.md §1.4 de H20; research D15 de H20).
func probarReglasDeNormativa(t *testing.T) {
	t.Parallel()

	todos := []string{
		"01-lpac-articulo-21.yaml", "02-lcsp-contrato-menor.yaml", "03-lrbrl-atribuciones-del-pleno.yaml",
		"04-lgt-prescripcion.yaml", "05-trlrhl-impuestos-municipales.yaml", "06-irpf-rendimientos-del-trabajo.yaml",
		"07-lrjsp-principio-de-legalidad.yaml", "08-ltaibg-plazo-de-resolucion.yaml",
		"09-constitucion-articulo-140.yaml", "10-et-vacaciones.yaml",
		"11-no-activa-programacion.yaml", "12-no-activa-acuerdo-entre-amigos.yaml",
		"13-lrbrl-atribuciones-por-materia.yaml", "14-sin-binario-ni-servidor.yaml", evalConFecha,
	}
	positivas := todos[:10]
	informativa := todos[12]
	sinBinario := todos[13]
	otraSinBinario := "16-sin-binario-ni-servidor-otra.yaml"

	// Con las ocho primeras, el conjunto tiene las 23 evals del máximo; con las
	// nueve, una más.
	deMas := []string{
		"16-no-activa-16.yaml", "17-no-activa-17.yaml", "18-no-activa-18.yaml", "19-no-activa-19.yaml",
		"20-no-activa-20.yaml", "21-no-activa-21.yaml", "22-no-activa-22.yaml", "23-no-activa-23.yaml",
		"24-no-activa-24.yaml",
	}
	conDeMas := func(conjunto *conjuntoSintetico, ficheros []string) {
		for _, fichero := range ficheros {
			conjunto.evals = append(conjunto.evals, Eval{Fichero: fichero, Pregunta: "¿Qué hora es?"})
		}
	}

	t.Run("cumple-todas", func(t *testing.T) {
		t.Parallel()

		conjunto := conjuntoQueCumple(t)
		require.Equal(t, todos, ficherosDelConjunto(conjunto), "el conjunto sintético es el del contrato")

		assert.Empty(t, ComprobarConjunto(conjunto.evals, conjunto.normas, ReglasDeNormativa()))

		conDeMas(&conjunto, deMas[:len(deMas)-1])
		require.Len(t, conjunto.evals, 23, "el conjunto sint\xc3\xa9tico llega al m\xc3\xa1ximo de evals")

		assert.Empty(t, ComprobarConjunto(conjunto.evals, conjunto.normas, ReglasDeNormativa()),
			"con 23 evals, el conjunto cumple el tama\xc3\xb1o")
	})

	casos := []struct {
		nombre    string
		modificar func(t *testing.T, conjunto *conjuntoSintetico)
		regla     string
		ficheros  []string
		normas    []string

		// dice es lo que el mensaje dice además de nombrar ficheros y normas.
		dice []string
	}{
		{
			nombre: "tamaño",
			modificar: func(_ *testing.T, conjunto *conjuntoSintetico) {
				conDeMas(conjunto, deMas)
			},
			regla:    "tamaño",
			ficheros: slices.Concat(todos, deMas),
			dice:     []string{"hay 24 evals y el conjunto lleva entre 10 y 23"},
		},
		{
			nombre: "positivas",
			modificar: func(t *testing.T, conjunto *conjuntoSintetico) {
				t.Helper()
				sinEvals(t, conjunto, "10-et-vacaciones.yaml")
			},
			regla:    "positivas",
			ficheros: positivas[:9],
		},
		{
			nombre: "no-activación",
			modificar: func(t *testing.T, conjunto *conjuntoSintetico) {
				t.Helper()
				sinEvals(t, conjunto, "11-no-activa-programacion.yaml", "12-no-activa-acuerdo-entre-amigos.yaml")
			},
			regla:    "no activación",
			ficheros: positivas,
		},
		{
			// Sin ninguna informativa, las preguntas por materia habrían
			// desaparecido del conjunto sin que nada lo dijera (ADR 0016).
			nombre: "informativas",
			modificar: func(t *testing.T, conjunto *conjuntoSintetico) {
				t.Helper()
				sinEvals(t, conjunto, informativa)
			},
			regla:    "informativas",
			ficheros: slices.Concat(positivas, todos[10:12]),
		},
		{
			nombre: "informativas-sin-activar",
			modificar: func(t *testing.T, conjunto *conjuntoSintetico) {
				t.Helper()
				eval := evalDe(t, conjunto.evals, informativa)
				eval.Activa, eval.Comandos, eval.Citas = false, nil, nil
			},
			regla:    "informativas",
			ficheros: []string{informativa},
		},
		{
			nombre: "materias-distintas",
			modificar: func(t *testing.T, conjunto *conjuntoSintetico) {
				t.Helper()
				et := evalDe(t, conjunto.evals, "10-et-vacaciones.yaml")
				et.Comandos = []ComandoEsperado{{Applet: "boe", Norma: normaConstitucion, Bloque: "a35"}}
				et.Citas = []CitaEsperada{{Norma: normaConstitucion, Bloque: "a35"}}
			},
			regla:    "materias distintas",
			ficheros: []string{"09-constitucion-articulo-140.yaml", "10-et-vacaciones.yaml"},
			normas:   []string{normaConstitucion},
		},
		{
			nombre: "normas-del-hito",
			modificar: func(t *testing.T, conjunto *conjuntoSintetico) {
				t.Helper()
				conjunto.normas[normaSinAbreviatura] = NormaConocida{Materias: []string{"haciendas locales"}}
				trlrhl := evalDe(t, conjunto.evals, "05-trlrhl-impuestos-municipales.yaml")
				trlrhl.Comandos = []ComandoEsperado{{Applet: "boe", Norma: normaSinAbreviatura, Bloque: "a1"}}
				trlrhl.Citas = []CitaEsperada{{Norma: normaSinAbreviatura, Bloque: "a1"}}
			},
			regla:    "normas del hito",
			ficheros: positivas,
			normas:   []string{"TRLRHL", normaTRLRHL},
		},
		{
			nombre: "normas-del-hito-sin-la-abreviatura",
			modificar: func(_ *testing.T, conjunto *conjuntoSintetico) {
				trlrhl := conjunto.normas[normaTRLRHL]
				trlrhl.Abreviatura = ""
				conjunto.normas[normaTRLRHL] = trlrhl
			},
			regla:    "normas del hito",
			ficheros: positivas,
			normas:   []string{"TRLRHL"},
		},
		{
			nombre: "normas-del-hito-con-la-abreviatura-repetida",
			modificar: func(_ *testing.T, conjunto *conjuntoSintetico) {
				conjunto.normas[normaSinAbreviatura] = NormaConocida{Abreviatura: "LGT", Materias: []string{"tributos"}}
			},
			regla:    "normas del hito",
			ficheros: positivas,
			normas:   []string{"LGT", normaLGT, normaSinAbreviatura},
		},
		{
			nombre: "art-21",
			modificar: func(t *testing.T, conjunto *conjuntoSintetico) {
				t.Helper()
				evalDe(t, conjunto.evals, "01-lpac-articulo-21.yaml").Citas[0].Bloque = "a22"
			},
			regla:    "art. 21",
			ficheros: []string{"01-lpac-articulo-21.yaml"},
			normas:   []string{normaLPAC},
		},
		{
			nombre: "art-21-con-otra-pregunta",
			modificar: func(t *testing.T, conjunto *conjuntoSintetico) {
				t.Helper()
				evalDe(t, conjunto.evals, "01-lpac-articulo-21.yaml").Pregunta = "¿Qué dice el art. 21 de la Ley 39/2015?"
			},
			regla:    "art. 21",
			ficheros: []string{"01-lpac-articulo-21.yaml"},
			normas:   []string{normaLPAC},
		},
		{
			nombre: "fiscal",
			modificar: func(_ *testing.T, conjunto *conjuntoSintetico) {
				sinMateria(conjunto, "tributos")
			},
			regla:    "fiscal",
			ficheros: positivas[1:],
		},
		{
			// La única eval del art. 21 no cuenta como la fiscal: FR-063 pide otra.
			nombre: "fiscal-solo-en-la-del-art-21",
			modificar: func(_ *testing.T, conjunto *conjuntoSintetico) {
				sinMateria(conjunto, "tributos")
				lpac := conjunto.normas[normaLPAC]
				lpac.Materias = append(lpac.Materias, "tributos")
				conjunto.normas[normaLPAC] = lpac
			},
			regla:    "fiscal",
			ficheros: positivas[1:],
			normas:   []string{normaLPAC},
		},
		{
			nombre: "boe-fiscal",
			modificar: func(t *testing.T, conjunto *conjuntoSintetico) {
				t.Helper()
				evalDe(t, conjunto.evals, "06-irpf-rendimientos-del-trabajo.yaml").Reproduce = ""
			},
			regla:    "boe-fiscal",
			ficheros: todos,
		},
		{
			nombre: "normas-conocidas",
			modificar: func(t *testing.T, conjunto *conjuntoSintetico) {
				t.Helper()
				lrbrl := evalDe(t, conjunto.evals, "03-lrbrl-atribuciones-del-pleno.yaml")
				lrbrl.Comandos = append(lrbrl.Comandos, ComandoEsperado{Applet: "boe", Verbo: "indice", Norma: normaDesconocida})
				ce := evalDe(t, conjunto.evals, "09-constitucion-articulo-140.yaml")
				ce.Citas = append(ce.Citas, CitaEsperada{Norma: normaDesconocida, Bloque: "a1"})
			},
			regla:    "normas conocidas",
			ficheros: []string{"03-lrbrl-atribuciones-del-pleno.yaml", "09-constitucion-articulo-140.yaml"},
			normas:   []string{normaDesconocida},
		},
		{
			// La eval con fecha no es positiva, pero la regla de las normas
			// conocidas sí la mira (research D15 de H20).
			nombre: "normas-conocidas-en-la-eval-con-fecha",
			modificar: func(t *testing.T, conjunto *conjuntoSintetico) {
				t.Helper()
				conFecha := evalDe(t, conjunto.evals, evalConFecha)
				conFecha.Comandos[0].Norma = normaDesconocida
				conFecha.Citas[0].Norma = normaDesconocida
			},
			regla:    "normas conocidas",
			ficheros: []string{evalConFecha},
			normas:   []string{normaDesconocida},
		},
		{
			nombre: "sin-binario-ni-servidor-ninguna",
			modificar: func(t *testing.T, conjunto *conjuntoSintetico) {
				t.Helper()
				sinEvals(t, conjunto, sinBinario)
			},
			regla: "sin binario ni servidor",
			dice:  []string{"hay 0 evals sin binario ni servidor", "exactamente 1: ning\xc3\xban fichero"},
		},
		{
			// La segunda tampoco cuenta como positiva: el conjunto sigue teniendo
			// sus diez, y solo incumple la regla nueva.
			nombre: "sin-binario-ni-servidor-dos",
			modificar: func(_ *testing.T, conjunto *conjuntoSintetico) {
				conjunto.evals = append(conjunto.evals, Eval{
					Fichero: otraSinBinario, Pregunta: "\xc2\xbfQu\xc3\xa9 dice el art. 54 de la Ley 39/2015?",
					Activa: true, SinBinarioNiServidor: true,
				})
			},
			regla:    "sin binario ni servidor",
			ficheros: []string{sinBinario, otraSinBinario},
			dice:     []string{"hay 2 evals sin binario ni servidor"},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			conjunto := conjuntoQueCumple(t)
			caso.modificar(t, &conjunto)

			defectos := ComprobarConjunto(conjunto.evals, conjunto.normas, ReglasDeNormativa())
			require.Len(t, defectos, 1, "la copia solo incumple la regla %s: %v", caso.regla, defectos)
			assert.Equal(t, caso.regla, defectos[0].Regla)

			for _, nombrado := range slices.Concat(caso.ficheros, caso.normas, caso.dice) {
				assert.Contains(t, defectos[0].Mensaje, nombrado, "el mensaje nombra lo implicado")
			}
		})
	}

	t.Run("orden-de-la-tabla", func(t *testing.T) {
		t.Parallel()

		conjunto := conjuntoQueCumple(t)
		conjunto.evals = nil

		for numero := 1; numero <= 24; numero++ {
			conjunto.evals = append(conjunto.evals, Eval{
				Fichero:  fmt.Sprintf("%02d-positiva.yaml", numero),
				Pregunta: "¿Qué dice esta norma?",
				Activa:   true,
				Comandos: []ComandoEsperado{{Applet: "boe", Norma: normaDesconocida, Bloque: "a1"}},
				Citas:    []CitaEsperada{{Norma: normaDesconocida, Bloque: "a1"}},
			})
		}

		assert.Equal(t, []string{
			"tamaño", "positivas", "no activación", "informativas", "materias distintas", "normas del hito",
			"art. 21", "fiscal", "boe-fiscal", "normas conocidas", "sin binario ni servidor",
		}, reglasIncumplidas(ComprobarConjunto(conjunto.evals, conjunto.normas, ReglasDeNormativa())))
	})
}

// Las evals sintéticas del conjunto de plazos-y-territorio de
// probarReglasDePlazosYTerritorio: la del municipio cubierto, la del no
// cubierto y la de no activación (contrato de evals §4 de H6), desde H21, la
// eval sin binario ni servidor (contracts/evals-en-dos-modos.md §1 de H21) y,
// desde H9, la de plazo y la de cobertura (contracts/evals-legal-core.md §4 de
// H9); y las que los casos le añaden.
const (
	plazosYTerritorioCubierto       = "01-territorio-municipio-cubierto.yaml"
	plazosYTerritorioNoCubierto     = "02-territorio-municipio-no-cubierto.yaml"
	plazosYTerritorioNoActivacion   = "03-no-activa-receta-de-cocina.yaml"
	plazosYTerritorioSinBinario     = "04-sin-binario-ni-servidor.yaml"
	plazosYTerritorioDePlazo        = "05-plazo-reposicion-leganes.yaml"
	plazosYTerritorioDeCobertura    = "07-cobertura-liquidacion-tributaria.yaml"
	plazosYTerritorioSinVerificable = "05-territorio-sin-esperado.yaml"
	plazosYTerritorioConCitas       = "05-articulo-21-con-citas.yaml"
	plazosYTerritorioOtraSinBinario = "05-sin-binario-ni-servidor-otra.yaml"
)

// conjuntoDePlazosYTerritorio devuelve, nuevo en cada llamada, un conjunto que
// cumple las reglas de plazos-y-territorio: una eval activa que resuelve un
// municipio del territorio configurado y espera su boletín, otra que resuelve
// uno de una comunidad sin configuración y espera no configurado cada aspecto
// de boletín, una de no activación, la eval sin binario ni servidor, activa y
// sin nada que consultar ni que esperar, y, desde H9, una de plazo que solo
// espera el plazo y una de cobertura que solo espera la declaración de no
// cubierto, sin comandos ni citas: las dos tienen un esperado verificable.
func conjuntoDePlazosYTerritorio() []Eval {
	return []Eval{
		{
			Fichero:  plazosYTerritorioCubierto,
			Pregunta: "¿En qué boletines se publican las normas que afectan a Leganés?",
			Activa:   true,
			Comandos: []ComandoEsperado{{Applet: "territorio", Verbo: "resolver", Municipio: "Leganés"}},
			Territorio: TerritorioEsperado{
				Comunidad: "Comunidad de Madrid", Provincia: "Madrid", Boletines: []string{"BOCM"},
			},
		},
		{
			Fichero:  plazosYTerritorioNoCubierto,
			Pregunta: "¿En qué boletines se publican las normas que afectan a Tordesillas?",
			Activa:   true,
			Comandos: []ComandoEsperado{{Applet: "territorio", Verbo: "resolver", Municipio: "Tordesillas"}},
			Territorio: TerritorioEsperado{
				Comunidad: "Castilla y León",
				Provincia: "Valladolid",
				Cobertura: []string{"boletin_autonomico: no-configurado", "boletin_provincial: no-configurado"},
			},
		},
		{Fichero: plazosYTerritorioNoActivacion, Pregunta: "¿Cómo hago una tortilla de patatas?"},
		{
			Fichero:              plazosYTerritorioSinBinario,
			Pregunta:             "\xc2\xbfQu\xc3\xa9 comunidad y qu\xc3\xa9 boletines corresponden al Ayuntamiento de Getafe?",
			Activa:               true,
			NoSeActivan:          []string{"normativa"},
			SinBinarioNiServidor: true,
		},
		evalDePlazoSintetica(),
		evalDeCoberturaSintetica(),
	}
}

// probarReglasDePlazosYTerritorio fija ComprobarConjunto con
// ReglasDePlazosYTerritorio (contrato de evals §3 de H6; FR-081, FR-082): el
// conjunto que las cumple no da ningún defecto, y tampoco con una eval activa
// más que solo espera citas; cada copia que incumple una regla da el defecto de
// esa regla, con su nombre y un mensaje que nombra los ficheros implicados; y
// un conjunto que las incumple todas da un defecto por regla, en el orden de la
// tabla. Desde H21, el conjunto lleva cuatro evals, con la que es sin binario
// ni servidor, que no necesita un esperado verificable: sin una de las cuatro
// incumple solo la regla de esa eval, sin dos incumple también el tamaño, y con
// dos evals sin binario ni servidor, solo la regla nueva
// (contracts/evals-en-dos-modos.md §1 de H21; FR-046).
//
// Desde H9, el conjunto lleva además una eval de plazo y una de cobertura, que
// solo esperan el plazo y la declaración de no cubierto y tienen con ellos un
// esperado verificable; sin una de las dos incumple solo la regla de la
// cobertura por materia, que va detrás de la del esperado verificable, y su
// mensaje nombra las que hay de cada clase; y sin cuatro, también el tamaño
// (contracts/evals-legal-core.md §4 de H9; FR-051, FR-088).
func probarReglasDePlazosYTerritorio(t *testing.T) {
	t.Parallel()

	t.Run("cumple-todas", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, ComprobarConjunto(conjuntoDePlazosYTerritorio(), nil, ReglasDePlazosYTerritorio()))

		conCitas := append(conjuntoDePlazosYTerritorio(), Eval{
			Fichero:  plazosYTerritorioConCitas,
			Pregunta: preguntaDelArticulo21Eval,
			Activa:   true,
			Comandos: []ComandoEsperado{{Applet: "boe", Norma: normaLPAC, Bloque: "a21"}},
			Citas:    []CitaEsperada{{Norma: normaLPAC, Bloque: "a21"}},
		})
		assert.Empty(t, ComprobarConjunto(conCitas, nil, ReglasDePlazosYTerritorio()),
			"una eval que solo espera citas tiene un esperado verificable")
	})

	casos := []struct {
		nombre    string
		modificar func(t *testing.T, evals []Eval) []Eval
		reglas    []string
		ficheros  []string

		// dice es lo que el mensaje del último defecto dice además de nombrar
		// ficheros.
		dice []string
	}{
		{
			nombre:    "sin-la-del-cubierto",
			modificar: sinLaEval(plazosYTerritorioCubierto),
			reglas:    []string{"cubierto"},
			ficheros:  []string{plazosYTerritorioNoCubierto},
		},
		{
			nombre:    "sin-la-del-no-cubierto",
			modificar: sinLaEval(plazosYTerritorioNoCubierto),
			reglas:    []string{"no cubierto"},
			ficheros:  []string{plazosYTerritorioCubierto},
		},
		{
			nombre:    "sin-la-de-no-activación",
			modificar: sinLaEval(plazosYTerritorioNoActivacion),
			reglas:    []string{"no activación"},
			ficheros:  []string{plazosYTerritorioCubierto, plazosYTerritorioNoCubierto, plazosYTerritorioSinBinario},
		},
		{
			nombre:    "sin-la-eval-sin-binario-ni-servidor",
			modificar: sinLaEval(plazosYTerritorioSinBinario),
			reglas:    []string{"sin binario ni servidor"},
			dice:      []string{"hay 0 evals sin binario ni servidor", "exactamente 1: ning\xc3\xban fichero"},
		},
		{
			nombre: "dos-evals-sin-binario-ni-servidor",
			modificar: func(_ *testing.T, evals []Eval) []Eval {
				return append(evals, Eval{
					Fichero:  plazosYTerritorioOtraSinBinario,
					Pregunta: "\xc2\xbfQu\xc3\xa9 boletines corresponden al Ayuntamiento de M\xc3\xb3stoles?",
					Activa:   true, SinBinarioNiServidor: true,
				})
			},
			reglas:   []string{"sin binario ni servidor"},
			ficheros: []string{plazosYTerritorioSinBinario, plazosYTerritorioOtraSinBinario},
			dice:     []string{"hay 2 evals sin binario ni servidor"},
		},
		{
			nombre:    "sin-la-de-cobertura",
			modificar: sinLaEval(plazosYTerritorioDeCobertura),
			reglas:    []string{"cobertura por materia"},
			ficheros:  []string{plazosYTerritorioDePlazo},
			dice:      []string{"de cobertura, ning\xc3\xban fichero"},
		},
		{
			nombre:    "sin-la-de-plazo",
			modificar: sinLaEval(plazosYTerritorioDePlazo),
			reglas:    []string{"cobertura por materia"},
			ficheros:  []string{plazosYTerritorioDeCobertura},
			dice:      []string{"de plazo, ning\xc3\xban fichero"},
		},
		{
			// Con cuatro evals menos, el conjunto ya no llega a las tres del mínimo,
			// y el defecto del tamaño va el primero, delante de los de las evals que
			// faltan.
			nombre: "tama\xc3\xb1o",
			modificar: func(t *testing.T, evals []Eval) []Eval {
				t.Helper()

				for _, fichero := range []string{plazosYTerritorioSinBinario, plazosYTerritorioNoActivacion, plazosYTerritorioDePlazo, plazosYTerritorioDeCobertura} {
					evals = sinLaEval(fichero)(t, evals)
				}

				return evals
			},
			reglas: []string{"tama\xc3\xb1o", "no activaci\xc3\xb3n", "cobertura por materia", "sin binario ni servidor"},
			dice:   []string{"hay 0 evals sin binario ni servidor"},
		},
		{
			nombre: "cubierto-sin-boletines",
			modificar: func(t *testing.T, evals []Eval) []Eval {
				t.Helper()
				evalDe(t, evals, plazosYTerritorioCubierto).Territorio.Boletines = nil

				return evals
			},
			reglas:   []string{"cubierto"},
			ficheros: []string{plazosYTerritorioCubierto, plazosYTerritorioNoCubierto},
		},
		{
			nombre: "cubierto-sin-comando-de-territorio",
			modificar: func(t *testing.T, evals []Eval) []Eval {
				t.Helper()
				evalDe(t, evals, plazosYTerritorioCubierto).Comandos = []ComandoEsperado{
					{Applet: "boe", Verbo: "buscar", Terminos: []string{"Leganés"}},
				}

				return evals
			},
			reglas:   []string{"cubierto"},
			ficheros: []string{plazosYTerritorioNoCubierto},
		},
		{
			nombre: "cubierto-sin-activar",
			modificar: func(t *testing.T, evals []Eval) []Eval {
				t.Helper()
				evalDe(t, evals, plazosYTerritorioCubierto).Activa = false

				return evals
			},
			reglas:   []string{"cubierto"},
			ficheros: []string{plazosYTerritorioNoCubierto},
		},
		{
			// Con boletines, la eval de una comunidad sin configuración sigue sin
			// ser la de un municipio del territorio configurado.
			nombre: "no-cubierto-con-boletines-no-es-cubierto",
			modificar: func(t *testing.T, evals []Eval) []Eval {
				t.Helper()
				evalDe(t, evals, plazosYTerritorioNoCubierto).Territorio.Boletines = []string{"BOE"}

				return sinLaEval(plazosYTerritorioCubierto)(t, evals)
			},
			reglas:   []string{"cubierto"},
			ficheros: []string{plazosYTerritorioNoCubierto},
		},
		{
			nombre: "no-cubierto-con-un-solo-aspecto",
			modificar: func(t *testing.T, evals []Eval) []Eval {
				t.Helper()
				evalDe(t, evals, plazosYTerritorioNoCubierto).Territorio.Cobertura = []string{"boletin_autonomico: no-configurado"}

				return evals
			},
			reglas:   []string{"no cubierto"},
			ficheros: []string{plazosYTerritorioCubierto, plazosYTerritorioNoCubierto},
		},
		{
			nombre: "esperado-verificable",
			modificar: func(_ *testing.T, evals []Eval) []Eval {
				return append(evals, Eval{
					Fichero:  plazosYTerritorioSinVerificable,
					Pregunta: "¿Qué boletín publica las ordenanzas de Leganés?",
					Activa:   true,
					Comandos: []ComandoEsperado{{Applet: "territorio", Verbo: "resolver", Municipio: "Leganés"}},
				})
			},
			reglas:   []string{"esperado verificable"},
			ficheros: []string{plazosYTerritorioSinVerificable},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			defectos := ComprobarConjunto(caso.modificar(t, conjuntoDePlazosYTerritorio()), nil, ReglasDePlazosYTerritorio())
			require.Equal(t, caso.reglas, reglasIncumplidas(defectos), "la copia incumple solo esas reglas: %v", defectos)

			for _, nombrado := range slices.Concat(caso.ficheros, caso.dice) {
				assert.Contains(t, defectos[len(defectos)-1].Mensaje, nombrado, "el mensaje nombra lo implicado")
			}
		})
	}

	t.Run("orden-de-la-tabla", func(t *testing.T) {
		t.Parallel()

		sinEsperado := []Eval{{
			Fichero:  plazosYTerritorioSinVerificable,
			Pregunta: "¿Qué boletín publica las ordenanzas de Leganés?",
			Activa:   true,
			Comandos: []ComandoEsperado{{Applet: "territorio", Verbo: "resolver", Municipio: "Leganés"}},
		}}

		assert.Equal(t, []string{
			"tamaño", "cubierto", "no cubierto", "no activación", "esperado verificable", "cobertura por materia",
			"sin binario ni servidor",
		}, reglasIncumplidas(ComprobarConjunto(sinEsperado, nil, ReglasDePlazosYTerritorio())))
	})

	t.Run("sin-configuración-ningún-boletín", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, []string{"boletin_autonomico: no-configurado", "boletin_provincial: no-configurado"},
			aspectosNoConfigurados(), "una comunidad sin configuración tiene sin configurar sus dos boletines")
	})
}

// Las evals sintéticas del conjunto de jurisprudencia de
// probarReglasDeJurisprudencia, con los nombres de las seis de
// contracts/evals-jurisprudencia.md §5 de H23 y de las cuatro de
// contracts/evals-jurisprudencia.md §1 de H25, y la que los casos le añaden.
const (
	jurisprudenciaExiste       = "01-existe-con-numero-y-fecha.yaml"
	jurisprudenciaResumen      = "02-resumen-sin-documento.yaml"
	jurisprudenciaPorMateria   = "03-por-materia.yaml"
	jurisprudenciaDocumento    = "04-documento-pegado.yaml"
	jurisprudenciaDelTC        = "05-tribunal-constitucional.yaml"
	jurisprudenciaNoEsElPedido = "06-documento-que-no-es-el-pedido.yaml"
	jurisprudenciaConocida     = "07-resumen-de-una-conocida.yaml"
	jurisprudenciaDoctrina     = "08-doctrina-con-el-fallo-delante.yaml"
	jurisprudenciaFicha        = "09-de-que-trata-con-la-ficha-sola.yaml"
	jurisprudenciaDadaPorHecha = "10-doctrina-dada-por-hecha.yaml"
	jurisprudenciaDeMas        = "11-una-de-mas.yaml"
)

// conjuntoDeJurisprudencia devuelve, nuevo en cada llamada, un conjunto que
// cumple las reglas de jurisprudencia: las diez evals de los dos contratos, con
// lo que cada una espera y una pregunta corta en las que pegan el documento.
func conjuntoDeJurisprudencia() []Eval {
	preparar := ComandoEsperado{Applet: "cita", Verbo: "preparar"}
	prepararConTexto := ComandoEsperado{Applet: "cita", Verbo: "preparar", ConTexto: true}

	// conElDocumento es una eval cuya pregunta lleva pegado un documento, que
	// espera cita cotejar y la cita de ese documento.
	conElDocumento := func(fichero, entrada string) Eval {
		return Eval{
			Fichero:  fichero,
			Pregunta: entrada + "\n\nRoj: STS 3144/2023 - ECLI:ES:TS:2023:3144\n",
			Activa:   true,
			Comandos: []ComandoEsperado{{Applet: "cita", Verbo: "cotejar"}},
			Sentencias: SentenciasEsperadas{
				Citas: []CitaDeSentenciaEsperada{{ECLI: ecliDelFragmento, ROJ: rojDelFragmento}},
			},
		}
	}

	// porMateria es una eval de una pregunta por materia, que espera cita
	// preparar con texto, la dirección de búsqueda y ninguna cita.
	porMateria := func(fichero, pregunta string) Eval {
		return Eval{
			Fichero:    fichero,
			Pregunta:   pregunta,
			Activa:     true,
			Comandos:   []ComandoEsperado{prepararConTexto},
			Sentencias: SentenciasEsperadas{DireccionDeBusqueda: true, NingunaCita: true},
		}
	}

	porNumeroYFecha := func(fichero, pregunta, numero, fecha string) Eval {
		return Eval{
			Fichero:  fichero,
			Pregunta: pregunta,
			Activa:   true,
			Comandos: []ComandoEsperado{preparar},
			Sentencias: SentenciasEsperadas{
				NoComprobada: true,
				Direcciones:  []string{direccionDelCendoj},
				Casillas: []CasillaEsperada{
					{Nombre: casillaDeResolucion, Valor: numero}, {Nombre: casillaDeFecha, Valor: fecha},
				},
				NingunaCita: true,
			},
		}
	}

	return []Eval{
		porNumeroYFecha(jurisprudenciaExiste, "\xc2\xbfexiste la STS 1088/2023, de 4 de julio?", "1088/2023", "04/07/2023"),
		porNumeroYFecha(jurisprudenciaResumen, "res\xc3\xbameme la STS 9999/2023, de 1 de enero", "9999/2023", "01/01/2023"),
		porMateria(jurisprudenciaPorMateria, "\xc2\xbfqu\xc3\xa9 dice la jurisprudencia sobre la cl\xc3\xa1usula suelo?"),
		conElDocumento(jurisprudenciaDocumento, "C\xc3\xadtame esta sentencia."),
		{
			Fichero:    jurisprudenciaDelTC,
			Pregunta:   "\xc2\xbfQu\xc3\xa9 resolvi\xc3\xb3 el Tribunal Constitucional en su sentencia 79/2024 (ECLI:ES:TC:2024:79)?",
			Activa:     true,
			Sentencias: SentenciasEsperadas{Direcciones: []string{direccionDelTC}, NingunaCita: true},
		},
		{
			Fichero:  jurisprudenciaNoEsElPedido,
			Pregunta: "Traigo la sentencia de ROJ STS 1088/2023.\n\nRoj: STS 3144/2023 - ECLI:ES:TS:2023:3144\n",
			Activa:   true,
			Comandos: []ComandoEsperado{
				{Applet: "cita", Verbo: "cotejar", ROJ: rojQueNoEsElSuyo},
				{Applet: "cita", Verbo: "preparar", ROJ: rojQueNoEsElSuyo},
			},
			Sentencias: SentenciasEsperadas{
				NoComprobada:  true,
				Direcciones:   []string{direccionDelCendoj},
				Casillas:      []CasillaEsperada{{Nombre: casillaDeROJ, Valor: rojQueNoEsElSuyo}},
				SinCitaDelROJ: []string{rojQueNoEsElSuyo},
			},
		},
		porNumeroYFecha(jurisprudenciaConocida, "res\xc3\xbameme la STS 241/2013, de 9 de mayo", "241/2013", "09/05/2013"),
		conElDocumento(jurisprudenciaDoctrina, "\xc2\xbfQu\xc3\xa9 doctrina fija esta sentencia y en qu\xc3\xa9 fundamentos se apoya?"),
		conElDocumento(jurisprudenciaFicha, "\xc2\xbfDe qu\xc3\xa9 trata esta sentencia?"),
		porMateria(jurisprudenciaDadaPorHecha, "\xc2\xbfEs verdad que el Tribunal Supremo declar\xc3\xb3 nulas las cl\xc3\xa1usulas "+
			"suelo por falta de transparencia? Dime en qu\xc3\xa9 sentencia."),
	}
}

// probarReglasDeJurisprudencia fija ComprobarConjunto con ReglasDeJurisprudencia
// (contracts/evals-jurisprudencia.md §5 de H23; FR-056; y, con las diez,
// contracts/evals-jurisprudencia.md §3 de H25; FR-065; SC-010): el conjunto de
// las diez evals no da ningún defecto; cada copia a la que le falta una de las
// diez —nueve—, o que lleva una más —once—, da el defecto del tamaño; cada copia
// en la que una deja de declarar algo de lo que su clase exige —la línea, la
// dirección, una casilla, ninguna cita, la cita, la dirección de búsqueda o su
// comando— da el defecto de la regla de esa clase, con su nombre y un mensaje
// que dice cuántas hay, cuántas lleva el conjunto y qué ficheros la cumplen; y
// un conjunto que las incumple todas da un defecto por regla, en el orden de la
// tabla.
func probarReglasDeJurisprudencia(t *testing.T) {
	t.Parallel()

	t.Run("cumple-todas", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, ComprobarConjunto(conjuntoDeJurisprudencia(), nil, ReglasDeJurisprudencia()))
	})

	const (
		tamanio           = "tama\xc3\xb1o"
		activanLaSkill    = "activaci\xc3\xb3n"
		sentencias        = "sentencias"
		numeroYFecha      = "n\xc3\xbamero y fecha"
		materia           = "materia"
		documento         = "documento"
		noCubierta        = "no cubierta"
		documentoDistinto = "documento distinto"
	)

	// cambiar devuelve la modificación que aplica el cambio a la eval con ese
	// fichero, que tiene que estar.
	cambiar := func(fichero string, cambio func(eval *Eval)) func(t *testing.T, evals []Eval) []Eval {
		return func(t *testing.T, evals []Eval) []Eval {
			t.Helper()
			cambio(evalDe(t, evals, fichero))

			return evals
		}
	}

	// deMas es una undécima eval, que activa la skill y declara sentencias sin
	// ser de ninguna de las cinco clases.
	deMas := Eval{
		Fichero:    jurisprudenciaDeMas,
		Pregunta:   "\xc2\xbfD\xc3\xb3nde se buscan las sentencias del Tribunal Supremo?",
		Activa:     true,
		Comandos:   []ComandoEsperado{{Applet: "cita", Verbo: "preparar"}},
		Sentencias: SentenciasEsperadas{Direcciones: []string{direccionDelCendoj}},
	}

	casos := []struct {
		nombre    string
		modificar func(t *testing.T, evals []Eval) []Eval
		reglas    []string

		// dice es lo que dicen, entre todos, los mensajes de los defectos.
		dice []string
	}{
		{
			nombre:    "sin-la-01",
			modificar: sinLaEval(jurisprudenciaExiste),
			reglas:    []string{tamanio, numeroYFecha},
			dice: []string{
				"hay 9 evals y el conjunto lleva exactamente 10",
				"hay 2 evals con la l\xc3\xadnea, ninguna cita, una direcci\xc3\xb3n y las casillas \xc2\xabN\xc2\xba Resoluci\xc3\xb3n\xc2\xbb y " +
					"\xc2\xabFecha resoluci\xc3\xb3n\xc2\xbb, con cita preparar, y el conjunto lleva exactamente 3: " +
					jurisprudenciaResumen + ", " + jurisprudenciaConocida,
			},
		},
		{
			nombre:    "sin-la-03",
			modificar: sinLaEval(jurisprudenciaPorMateria),
			reglas:    []string{tamanio, materia},
			dice: []string{
				"hay 1 evals con la direcci\xc3\xb3n de b\xc3\xbasqueda y ninguna cita, con cita preparar con texto, " +
					"y el conjunto lleva exactamente 2: " + jurisprudenciaDadaPorHecha,
			},
		},
		{
			nombre:    "sin-la-04",
			modificar: sinLaEval(jurisprudenciaDocumento),
			reglas:    []string{tamanio, documento},
			dice: []string{"hay 2 evals con una cita, con cita cotejar, y el conjunto lleva exactamente 3: " +
				jurisprudenciaDoctrina + ", " + jurisprudenciaFicha},
		},
		{
			nombre:    "sin-la-05",
			modificar: sinLaEval(jurisprudenciaDelTC),
			reglas:    []string{tamanio, noCubierta},
			dice: []string{"hay 0 evals sin comandos, con una direcci\xc3\xb3n y ninguna cita, y el conjunto lleva exactamente 1: " +
				"ning\xc3\xban fichero"},
		},
		{
			nombre:    "sin-la-06",
			modificar: sinLaEval(jurisprudenciaNoEsElPedido),
			reglas:    []string{tamanio, documentoDistinto},
			dice: []string{
				"hay 0 evals con la l\xc3\xadnea, un ROJ que no se cita, y cita cotejar y cita preparar con ese ROJ, " +
					"y el conjunto lleva exactamente 1",
			},
		},
		{
			nombre:    "sin-la-07",
			modificar: sinLaEval(jurisprudenciaConocida),
			reglas:    []string{tamanio, numeroYFecha},
			dice: []string{
				"hay 9 evals y el conjunto lleva exactamente 10",
				"hay 2 evals", "exactamente 3: " + jurisprudenciaExiste + ", " + jurisprudenciaResumen,
			},
		},
		{
			nombre:    "sin-la-08",
			modificar: sinLaEval(jurisprudenciaDoctrina),
			reglas:    []string{tamanio, documento},
			dice:      []string{"hay 2 evals", "exactamente 3: " + jurisprudenciaDocumento + ", " + jurisprudenciaFicha},
		},
		{
			nombre:    "sin-la-09",
			modificar: sinLaEval(jurisprudenciaFicha),
			reglas:    []string{tamanio, documento},
			dice:      []string{"hay 2 evals", "exactamente 3: " + jurisprudenciaDocumento + ", " + jurisprudenciaDoctrina},
		},
		{
			nombre:    "sin-la-10",
			modificar: sinLaEval(jurisprudenciaDadaPorHecha),
			reglas:    []string{tamanio, materia},
			dice:      []string{"hay 1 evals", "exactamente 2: " + jurisprudenciaPorMateria},
		},
		{
			nombre:    "con-una-de-mas",
			modificar: func(_ *testing.T, evals []Eval) []Eval { return append(evals, deMas) },
			reglas:    []string{tamanio},
			dice: []string{
				"hay 11 evals y el conjunto lleva exactamente 10", jurisprudenciaExiste, jurisprudenciaDadaPorHecha,
				jurisprudenciaDeMas,
			},
		},
		{
			nombre: "con-dos-del-tribunal-constitucional",
			modificar: func(t *testing.T, evals []Eval) []Eval {
				t.Helper()

				otra := *evalDe(t, evals, jurisprudenciaDelTC)
				otra.Fichero = jurisprudenciaDeMas

				return append(evals, otra)
			},
			reglas: []string{tamanio, noCubierta},
			dice: []string{
				"hay 2 evals sin comandos, con una direcci\xc3\xb3n y ninguna cita, y el conjunto lleva exactamente 1: " +
					jurisprudenciaDelTC + ", " + jurisprudenciaDeMas,
			},
		},
		{
			nombre:    "una-que-no-activa-la-skill",
			modificar: cambiar(jurisprudenciaDelTC, func(eval *Eval) { eval.Activa = false }),
			reglas:    []string{activanLaSkill},
			dice:      []string{"evals que no activan la skill (activa: false): " + jurisprudenciaDelTC},
		},
		{
			nombre: "una-que-no-declara-sentencias",
			modificar: cambiar(jurisprudenciaDocumento, func(eval *Eval) {
				eval.Sentencias = SentenciasEsperadas{}
				eval.Citas = []CitaEsperada{{Norma: normaLPAC, Bloque: "a21"}}
			}),
			reglas: []string{sentencias, documento},
			dice:   []string{"evals que no declaran sentencias: " + jurisprudenciaDocumento},
		},
		{
			nombre:    "01-sin-ninguna-cita",
			modificar: cambiar(jurisprudenciaExiste, func(eval *Eval) { eval.Sentencias.NingunaCita = false }),
			reglas:    []string{numeroYFecha},
			dice:      []string{"hay 2 evals", "exactamente 3: " + jurisprudenciaResumen + ", " + jurisprudenciaConocida},
		},
		{
			nombre:    "02-sin-la-linea",
			modificar: cambiar(jurisprudenciaResumen, func(eval *Eval) { eval.Sentencias.NoComprobada = false }),
			reglas:    []string{numeroYFecha},
			dice:      []string{"hay 2 evals", "exactamente 3: " + jurisprudenciaExiste + ", " + jurisprudenciaConocida},
		},
		{
			// Desde H25 la clase exige además una dirección (FR-061 de H25).
			nombre:    "02-sin-la-direccion",
			modificar: cambiar(jurisprudenciaResumen, func(eval *Eval) { eval.Sentencias.Direcciones = nil }),
			reglas:    []string{numeroYFecha},
			dice:      []string{"hay 2 evals", "exactamente 3: " + jurisprudenciaExiste + ", " + jurisprudenciaConocida},
		},
		{
			nombre: "01-sin-la-casilla-de-la-fecha",
			modificar: cambiar(jurisprudenciaExiste, func(eval *Eval) {
				eval.Sentencias.Casillas = eval.Sentencias.Casillas[:1]
			}),
			reglas: []string{numeroYFecha},
		},
		{
			nombre:    "01-con-cotejar-y-no-preparar",
			modificar: cambiar(jurisprudenciaExiste, func(eval *Eval) { eval.Comandos[0].Verbo = "cotejar" }),
			reglas:    []string{numeroYFecha},
		},
		{
			nombre:    "07-sin-la-linea",
			modificar: cambiar(jurisprudenciaConocida, func(eval *Eval) { eval.Sentencias.NoComprobada = false }),
			reglas:    []string{numeroYFecha},
			dice:      []string{"hay 2 evals", "exactamente 3: " + jurisprudenciaExiste + ", " + jurisprudenciaResumen},
		},
		{
			nombre:    "07-sin-la-direccion",
			modificar: cambiar(jurisprudenciaConocida, func(eval *Eval) { eval.Sentencias.Direcciones = nil }),
			reglas:    []string{numeroYFecha},
		},
		{
			nombre: "07-sin-la-casilla-del-numero",
			modificar: cambiar(jurisprudenciaConocida, func(eval *Eval) {
				eval.Sentencias.Casillas = eval.Sentencias.Casillas[1:]
			}),
			reglas: []string{numeroYFecha},
		},
		{
			nombre:    "07-sin-ninguna-cita",
			modificar: cambiar(jurisprudenciaConocida, func(eval *Eval) { eval.Sentencias.NingunaCita = false }),
			reglas:    []string{numeroYFecha},
		},
		{
			nombre:    "07-con-cotejar-y-no-preparar",
			modificar: cambiar(jurisprudenciaConocida, func(eval *Eval) { eval.Comandos[0].Verbo = "cotejar" }),
			reglas:    []string{numeroYFecha},
		},
		{
			nombre:    "03-con-preparar-sin-texto",
			modificar: cambiar(jurisprudenciaPorMateria, func(eval *Eval) { eval.Comandos[0].ConTexto = false }),
			reglas:    []string{materia},
		},
		{
			nombre: "03-sin-la-direccion-de-busqueda",
			modificar: cambiar(jurisprudenciaPorMateria, func(eval *Eval) {
				eval.Sentencias.DireccionDeBusqueda = false
			}),
			reglas: []string{materia},
		},
		{
			nombre: "03-sin-ninguna-cita",
			modificar: cambiar(jurisprudenciaPorMateria, func(eval *Eval) {
				eval.Sentencias.NingunaCita = false
			}),
			reglas: []string{materia},
		},
		{
			nombre:    "10-con-preparar-sin-texto",
			modificar: cambiar(jurisprudenciaDadaPorHecha, func(eval *Eval) { eval.Comandos[0].ConTexto = false }),
			reglas:    []string{materia},
			dice:      []string{"hay 1 evals", "exactamente 2: " + jurisprudenciaPorMateria},
		},
		{
			nombre: "10-sin-la-direccion-de-busqueda",
			modificar: cambiar(jurisprudenciaDadaPorHecha, func(eval *Eval) {
				eval.Sentencias.DireccionDeBusqueda = false
			}),
			reglas: []string{materia},
		},
		{
			nombre: "10-sin-ninguna-cita",
			modificar: cambiar(jurisprudenciaDadaPorHecha, func(eval *Eval) {
				eval.Sentencias.NingunaCita = false
			}),
			reglas: []string{materia},
		},
		{
			nombre:    "04-con-preparar-y-no-cotejar",
			modificar: cambiar(jurisprudenciaDocumento, func(eval *Eval) { eval.Comandos[0].Verbo = "preparar" }),
			reglas:    []string{documento},
		},
		{
			nombre: "04-sin-la-cita",
			modificar: cambiar(jurisprudenciaDocumento, func(eval *Eval) {
				eval.Sentencias = SentenciasEsperadas{NingunaCita: true}
			}),
			reglas: []string{documento},
		},
		{
			nombre:    "08-con-preparar-y-no-cotejar",
			modificar: cambiar(jurisprudenciaDoctrina, func(eval *Eval) { eval.Comandos[0].Verbo = "preparar" }),
			reglas:    []string{documento},
			dice:      []string{"hay 2 evals", "exactamente 3: " + jurisprudenciaDocumento + ", " + jurisprudenciaFicha},
		},
		{
			nombre: "08-sin-la-cita",
			modificar: cambiar(jurisprudenciaDoctrina, func(eval *Eval) {
				eval.Sentencias = SentenciasEsperadas{NingunaCita: true}
			}),
			reglas: []string{documento},
		},
		{
			nombre:    "09-con-preparar-y-no-cotejar",
			modificar: cambiar(jurisprudenciaFicha, func(eval *Eval) { eval.Comandos[0].Verbo = "preparar" }),
			reglas:    []string{documento},
		},
		{
			nombre: "09-sin-la-cita",
			modificar: cambiar(jurisprudenciaFicha, func(eval *Eval) {
				eval.Sentencias = SentenciasEsperadas{NingunaCita: true}
			}),
			reglas: []string{documento},
		},
		{
			nombre: "05-con-un-comando",
			modificar: cambiar(jurisprudenciaDelTC, func(eval *Eval) {
				eval.Comandos = []ComandoEsperado{{Applet: "cita", Verbo: "preparar"}}
			}),
			reglas: []string{noCubierta},
		},
		{
			nombre: "05-sin-direccion",
			modificar: cambiar(jurisprudenciaDelTC, func(eval *Eval) {
				eval.Sentencias.Direcciones = nil
			}),
			reglas: []string{noCubierta},
		},
		{
			nombre:    "06-que-prepara-otro-roj",
			modificar: cambiar(jurisprudenciaNoEsElPedido, func(eval *Eval) { eval.Comandos[1].ROJ = rojDelFragmento }),
			reglas:    []string{documentoDistinto},
		},
		{
			nombre:    "06-que-coteja-sin-el-roj",
			modificar: cambiar(jurisprudenciaNoEsElPedido, func(eval *Eval) { eval.Comandos[0].ROJ = "" }),
			reglas:    []string{documentoDistinto},
		},
		{
			nombre: "06-sin-la-linea",
			modificar: cambiar(jurisprudenciaNoEsElPedido, func(eval *Eval) {
				eval.Sentencias.NoComprobada = false
			}),
			reglas: []string{documentoDistinto},
		},
		{
			nombre: "06-sin-el-roj-que-no-se-cita",
			modificar: cambiar(jurisprudenciaNoEsElPedido, func(eval *Eval) {
				eval.Sentencias.SinCitaDelROJ = nil
			}),
			reglas: []string{documentoDistinto},
		},
		{
			// El applet es parte del comando: el de otro applet con el mismo verbo
			// no es cita cotejar.
			nombre:    "04-con-cotejar-de-otro-applet",
			modificar: cambiar(jurisprudenciaDocumento, func(eval *Eval) { eval.Comandos[0].Applet = "boe" }),
			reglas:    []string{documento},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			defectos := ComprobarConjunto(caso.modificar(t, conjuntoDeJurisprudencia()), nil, ReglasDeJurisprudencia())
			require.Equal(t, caso.reglas, reglasIncumplidas(defectos), "la copia incumple solo esas reglas: %v", defectos)

			mensajes := presentarDefectos(defectos)
			for _, dicho := range caso.dice {
				assert.Contains(t, mensajes, dicho, "los mensajes dicen cu\xc3\xa1l falta")
			}
		})
	}

	t.Run("orden-de-la-tabla", func(t *testing.T) {
		t.Parallel()

		ninguna := []Eval{{Fichero: "01-no-activa-receta-de-cocina.yaml", Pregunta: "\xc2\xbfC\xc3\xb3mo hago una tortilla?"}}

		assert.Equal(t, []string{
			tamanio, activanLaSkill, sentencias, numeroYFecha, materia, documento, noCubierta, documentoDistinto,
		}, reglasIncumplidas(ComprobarConjunto(ninguna, nil, ReglasDeJurisprudencia())))
	})
}

// sinLaEval devuelve la modificación de probarReglasDePlazosYTerritorio que
// quita del conjunto la eval con ese fichero, que tiene que estar.
func sinLaEval(fichero string) func(t *testing.T, evals []Eval) []Eval {
	return func(t *testing.T, evals []Eval) []Eval {
		t.Helper()
		evalDe(t, evals, fichero)

		return slices.DeleteFunc(evals, func(eval Eval) bool { return eval.Fichero == fichero })
	}
}

// reglasIncumplidas son las reglas de los defectos, en su orden.
func reglasIncumplidas(defectos []DefectoDelConjunto) []string {
	reglas := []string{}
	for _, defecto := range defectos {
		reglas = append(reglas, defecto.Regla)
	}

	return reglas
}

// ficherosDelConjunto son los ficheros de las evals del conjunto, en su orden.
func ficherosDelConjunto(conjunto conjuntoSintetico) []string {
	ficheros := make([]string, 0, len(conjunto.evals))
	for _, eval := range conjunto.evals {
		ficheros = append(ficheros, eval.Fichero)
	}

	return ficheros
}

// Lo que TestEvalsDelRepositorio lee del repositorio y lo que le exige (contrato
// evals-y-grabaciones §2).
const (
	// evalsDeNormativa es evals/normativa/, relativo al directorio de
	// este paquete, que es donde go test ejecuta sus tests (research.md V46).
	evalsDeNormativa = "../../evals/normativa"

	// evalsMinimasDelRepositorio son las doce del contrato de evals §2, diez
	// positivas y dos de no activación: con menos, el subtest grabado podría
	// pasar sin comprobar lo que necesita alguna de ellas (plan.md, obligación
	// 12).
	evalsMinimasDelRepositorio = 12

	// reglaDeNormasConocidas es la regla de data-model §6.3 cuyos defectos
	// presenta su propio subtest, normas-conocidas, y no el subtest conjunto.
	reglaDeNormasConocidas = "normas conocidas"

	// directorioDeEvals es evals/, relativo al directorio de este paquete: en
	// cada una de sus carpetas están las evals de una skill (FR-060).
	directorioDeEvals = "../../evals"

	// evalsDePlazosYTerritorio es evals/plazos-y-territorio/, relativo al
	// directorio de este paquete: las evals de la skill plazos-y-territorio, a
	// las que se aplican las reglas de su juego (contrato de evals §3 y §4 de
	// H6).
	evalsDePlazosYTerritorio = "../../evals/plazos-y-territorio"

	// evalsDeJurisprudencia es evals/jurisprudencia/, relativo al directorio de
	// este paquete: las evals de la skill jurisprudencia, a las que se aplican
	// las reglas de su juego (contracts/evals-jurisprudencia.md §5 de H23 y §3 de
	// H25).
	evalsDeJurisprudencia = "../../evals/jurisprudencia"

	// evalsDePlazosYTerritorioConLaDeVigencia son las de
	// evals/plazos-y-territorio/: desde H9, las cuatro de antes, las tres de
	// plazo y las dos de cobertura (contracts/evals-legal-core.md §5 de H9;
	// FR-050), desde el ADR 0042, la de vigencia, y desde H26.1, la de plazo y la
	// de cobertura con una lectura sin servir (FR-017, FR-018 de H26.1).
	evalsDePlazosYTerritorioConLaDeVigencia = 12
)

// TestEvalsDelRepositorio comprueba sin red las evals del repositorio (contrato
// evals-y-grabaciones §2 y §5.2; FR-060, FR-062 a FR-066, FR-074, FR-075,
// SC-010): cada fichero de cada directorio evals/<skill>/ se lee como eval, y las
// de evals/normativa/ cumplen las reglas del conjunto de data-model §6.3,
// esperan solo normas de data/normas.yaml y tienen en las grabaciones de H4 y de
// H5 lo que necesitan para servir sin red cada consulta; el esquema publicado
// del formato admite en avisos exactamente los códigos de aviso del binario
// (FR-013 de H5.1); el SKILL.md de normativa lleva la forma fija de cada
// uno de esos códigos, reconocida con la misma función que usa Juzgar (FR-014 de
// H5.1); el esquema publicado admite en hallazgos exactamente las clases de
// hallazgo que el binario etiqueta (contrato evals-y-skill §3 de H7.1; FR-054);
// el SKILL.md de normativa lleva la forma fija de cada una de esas clases,
// reconocida con la misma función que usa Juzgar (FR-045, FR-093 de H7.1);
// el esquema publicado admite en la cobertura del territorio esperado
// exactamente las combinaciones del vocabulario del applet territorio (contrato de
// evals §1.2 de H6); las de evals/plazos-y-territorio/ cumplen las reglas del
// conjunto de plazos-y-territorio (contrato de evals §3 de H6; FR-080 a FR-082,
// SC-011); y el grafo previo de cada eval que lo lleva existe y se prepara sin
// faltas, con un BloqueVersion por comando (contrato evals-y-skill §3 de H7;
// FR-085). Desde H7.2, sobre
// cada grafo previo, la lectura de los bloques de la eval y graph check dan los
// hallazgos que la eval espera, de la redacción que dejó el grafo previo a la
// leída (contrato eval-y-derivada §4; FR-002 de H7.2). Desde H7.3, ninguna
// expresión de la lista de expresiones prohibidas de evals/normativa/ va
// en la prosa del SKILL.md de normativa —fuera
// del código y de la región generada— ni el fichero lleva ninguna fecha AAAAMMDD
// escrita con cifras (contracts/skill-boe-legislacion.md §5; FR-091, SC-005).
// Desde H7.4, cada orden de lectura y
// comprobación de la skill lleva detrás su forma para PowerShell
// (contracts/skill-boe-legislacion.md
// §4 de H7.4; FR-095; SC-005). Desde
// H21, las 21 evals de normativa y las 4 de plazos-y-territorio cumplen además
// la regla sin binario ni servidor de su juego: cada carpeta lleva exactamente
// una eval que lo declara (contracts/evals-en-dos-modos.md §1 de H21; FR-046).
// Desde H23, las evals de evals/jurisprudencia/ cumplen las reglas de su juego
// (contracts/evals-jurisprudencia.md §5 de H23; FR-056); desde H25 son diez, con
// las cuatro que llevan una pregunta del sondeo de la validación de su juez
// (contracts/evals-jurisprudencia.md §3 de H25; FR-065; SC-010), y su carpeta
// tiene juez, que declara sus dos clases, afirma_lo_no_leido, que decide, y
// afirma_que_existe, que solo se publica (contracts/juez-de-jurisprudencia.md
// §1 de H25; FR-002). Desde
// H24, la carpeta del juez de cada skill que la tiene se lee con sus evals y no
// da ningún fichero mal formado: la de normativa declara sus dos clases,
// afirma_lo_no_leido, que decide, y cuenta_su_proceso, que solo se publica, y
// plazos-y-territorio no tiene juez (contracts/juez-y-voto.md §1 de H24;
// FR-020, FR-021); y la lista deja de compararse con respuestas, con los
// bloques grabados y con las formas de la skill, porque ya no juzga ninguna
// respuesta: del calibrado de H7.4 quedan sus tres totales, como premisa de
// TestJuzgarSinLaLista (FR-071 de H24). Desde H26, la clase que decide lleva en
// cada declaración el umbral del ADR 0038, una tasa sobre las respuestas de
// todos los modos juntas —0.028 en normativa y 0.017 en jurisprudencia—,
// y la que solo se publica sigue con 0: los dos valores van aquí escritos a
// mano y se comparan con igualdad exacta, de modo que otro en cualquiera de las
// dos declaraciones no pasa (contracts/clases-yaml.diff de H26; data-model §1
// de H26; research D9 de H26; FR-020, FR-022, FR-077; SC-007). Desde H9, las de
// evals/plazos-y-territorio/ son nueve, bien formadas, cumplen además la regla
// de la cobertura por materia, y la unión de grabaciones sirve sin red lo que
// leen del BOE, como las de normativa (contracts/evals-legal-core.md §4 y §5 de
// H9; FR-050, FR-052, FR-088). Desde H20, las consultas de las dos carpetas
// llevan, junto a cada bloque, su lectura con --fecha 9999-12-31, y la caché se
// prepara y se comprueba sin red también con ellas
// (contracts/evals-redaccion-fechada.md §4 de H20; FR-053); y las de normativa
// son veintitrés, con las dos con fecha, bien formadas, con las reglas de su
// conjunto y con su caché preparada y comprobada sin red
// (contracts/evals-redaccion-fechada.md §1.3 y §1.4 de H20; FR-050 a FR-053); y
// el SKILL.md de normativa lleva, en sus bloques de código text, la forma fija
// de cada una de las tres líneas de una redacción fechada
// (contracts/skill-boe-legislacion.md §1 de H20; FR-048). Desde H26.1, la 14,
// la 15 y la 16 de normativa llevan sus lecturas, y desde H31.1 también la 17,
// y ninguna otra eval lleva ninguna; sus bloques son consultas necesarias, que
// el subtest grabado exige grabadas y servidas sin red como las demás
// (contracts/formato-de-eval.md §3.3 de H26.1; FR-010, FR-011;
// contracts/lecturas-y-grabaciones.md §1 de H31.1). Desde H31, los controles de
// contenido de las skills —la prosa, las órdenes para PowerShell, la línea sin
// consulta, las formas de la redacción fechada, las de aviso y las de hallazgo—
// leen la skill entera, su SKILL.md y sus referencias a mano, cada fichero por
// separado y con su ruta en cada defecto; las formas fijas se siguen exigiendo
// en SKILL.md, y cada línea ⚠ de una referencia a mano de normativa y de
// plazos-y-territorio es una de su SKILL.md
// (contracts/controles-de-contenido.md de H31; research D16 de H31; FR-016 a
// FR-018); y cada orden de ejemplo concreta de kitlegal boe de la skill entera
// de cada skill con evals sale con 0 con --offline sobre la caché de una sesión
// de su skill (contracts/controles-de-contenido.md §10 de H31; research D17 de
// H31; FR-019). Desde H32, el no_cubierto.precepto de cada eval de cobertura
// —la 07, la 08 y la 12 de plazos-y-territorio, las tres con la DA 1ª de la
// LPAC— es el precepto de una fila de data/cobertura.yaml, y un bloque que
// ninguna fila tiene da el defecto (contracts/forma-de-cita.md §4 de H32;
// data-model §8 de H32; research D16 de H32; FR-014). Lee las carpetas enteras.
func TestEvalsDelRepositorio(t *testing.T) {
	t.Parallel()

	conjunto, errDeLaLectura := LeerConjunto(evalsDeNormativa)

	normas, err := skills.LeerNormas(contenidoDelFichero(t, tablaDeNormasDelRepositorio))
	require.NoError(t, err, "la tabla de normas %s", tablaDeNormasDelRepositorio)

	// Una sola llamada: conjunto y normas-conocidas presentan cada uno una parte
	// de sus defectos.
	defectos := ComprobarConjunto(conjunto.Evals, normasConocidasDe(normas), ReglasDeNormativa())
	esDeNormasConocidas := func(defecto DefectoDelConjunto) bool { return defecto.Regla == reglaDeNormasConocidas }

	t.Run("formato", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, errDeLaLectura)
		assert.GreaterOrEqual(t, len(conjunto.Evals), evalsMinimasDelRepositorio,
			"%s tiene al menos las doce evals bien formadas del contrato", evalsDeNormativa)

		// El formato común es el de las evals de cualquier skill (FR-060): se leen
		// todas las carpetas de evals/, la de normativa entre ellas.
		carpetas, malFormados := malFormadosDeCadaSkill(t, directorioDeEvals)

		require.Contains(t, carpetas, evalsDeNormativa)
		assert.Empty(t, malFormados, "ficheros mal formados en %s:\n%s",
			directorioDeEvals, strings.Join(malFormados, "\n"))

		// La carpeta del juez se lee con las evals: normativa declara sus
		// dos clases y plazos-y-territorio no tiene juez (FR-020 y FR-021 de H24).
		// La que decide lleva el umbral del ADR 0038, como mucho 3 marcadas de 108,
		// y la que solo se publica, 0 (FR-020 y FR-022 de H26).
		require.NotNil(t, conjunto.Juez, "%s tiene la carpeta del juez", evalsDeNormativa)
		assert.Equal(t, []ClaseDelJuez{
			{Nombre: "afirma_lo_no_leido", Decide: true, Umbral: 0.028},
			{Nombre: "cuenta_su_proceso", Decide: false, Umbral: 0},
		}, conjunto.Juez.Clases)

		dePlazosYTerritorio, err := LeerConjunto(evalsDePlazosYTerritorio)
		require.NoError(t, err)
		assert.Nil(t, dePlazosYTerritorio.Juez, "%s no tiene juez", evalsDePlazosYTerritorio)
	})

	t.Run("conjunto", func(t *testing.T) {
		t.Parallel()

		deOtrasReglas := slices.DeleteFunc(slices.Clone(defectos), esDeNormasConocidas)
		assert.Empty(t, deOtrasReglas, "defectos del conjunto de %s:\n%s",
			evalsDeNormativa, presentarDefectos(deOtrasReglas))
	})

	t.Run("conjunto-plazos-y-territorio", func(t *testing.T) {
		t.Parallel()

		dePlazosYTerritorio, err := LeerConjunto(evalsDePlazosYTerritorio)
		require.NoError(t, err)

		defectosDePlazosYTerritorio := ComprobarConjunto(dePlazosYTerritorio.Evals, normasConocidasDe(normas), ReglasDePlazosYTerritorio())
		assert.Empty(t, defectosDePlazosYTerritorio, "defectos del conjunto de %s:\n%s",
			evalsDePlazosYTerritorio, presentarDefectos(defectosDePlazosYTerritorio))

		// Las doce de plazos-y-territorio, bien formadas (FR-050 de H9; ADR 0042;
		// FR-017 y FR-018 de H26.1), y lo que leen del BOE sus sesiones, servido
		// sin red por la unión de grabaciones (FR-052 de H9).
		assert.Empty(t, dePlazosYTerritorio.MalFormados, "cada fichero de %s es una eval bien formada", evalsDePlazosYTerritorio)

		ficheros := make([]string, 0, len(dePlazosYTerritorio.Evals))
		for _, eval := range dePlazosYTerritorio.Evals {
			ficheros = append(ficheros, eval.Fichero)
		}

		assert.Len(t, ficheros, evalsDePlazosYTerritorioConLaDeVigencia, "%s tiene las doce evals de plazos-y-territorio",
			evalsDePlazosYTerritorio)

		consultas := ConsultasNecesarias(dePlazosYTerritorio.Evals)
		require.NotEmpty(t, consultas, "las evals de %s leen algún bloque del BOE", evalsDePlazosYTerritorio)

		// Desde H20, cada bloque con su lectura con --fecha 9999-12-31 (FR-053).
		exigeLecturasFechadas(t, consultas, evalsDePlazosYTerritorio)

		dirCache := t.TempDir()

		preparadas, err := Preparar(dirCache, UnionDeGrabaciones(), consultas)
		require.NoError(t, err)
		assert.Empty(t, preparadas, "consultas de las evals de %s sin su respuesta en las grabaciones %s:\n%s",
			evalsDePlazosYTerritorio, strings.Join(UnionDeGrabaciones(), " y "), presentarFaltas(preparadas))

		comprobadas, err := ComprobarSinRed(dirCache, consultas)
		require.NoError(t, err)
		assert.Empty(t, comprobadas, "consultas de las evals de %s que la caché preparada no sirve sin red:\n%s",
			evalsDePlazosYTerritorio, presentarFaltas(comprobadas))
	})

	t.Run("precepto-de-la-cobertura", probarLosPreceptosDeLaCobertura)

	t.Run("conjunto-jurisprudencia", func(t *testing.T) {
		t.Parallel()

		deJurisprudencia, err := LeerConjunto(evalsDeJurisprudencia)
		require.NoError(t, err)

		// Desde H25 tiene juez: su carpeta se lee con sus evals y declara sus dos
		// clases, la que decide y la que solo se publica, y no cuenta_su_proceso
		// (contracts/juez-de-jurisprudencia.md §1 de H25; FR-002). La que decide
		// lleva el umbral del ADR 0038, como mucho 1 marcada de 60, y la que solo
		// se publica, 0 (FR-020 y FR-022 de H26).
		require.NotNil(t, deJurisprudencia.Juez, "%s tiene la carpeta del juez", evalsDeJurisprudencia)
		assert.Equal(t, []ClaseDelJuez{
			{Nombre: "afirma_lo_no_leido", Decide: true, Umbral: 0.017},
			{Nombre: "afirma_que_existe", Decide: false, Umbral: 0},
		}, deJurisprudencia.Juez.Clases)

		delConjunto := ComprobarConjunto(deJurisprudencia.Evals, nil, ReglasDeJurisprudencia())
		assert.Empty(t, delConjunto, "defectos del conjunto de %s:\n%s",
			evalsDeJurisprudencia, presentarDefectos(delConjunto))

		// El applet cita no consulta nada: sus evals no necesitan grabaciones
		// (FR-054 de H23).
		assert.Empty(t, ConsultasNecesarias(deJurisprudencia.Evals),
			"las evals de %s no necesitan ninguna respuesta grabada", evalsDeJurisprudencia)
	})

	t.Run("piden-el-cotejo", func(t *testing.T) {
		t.Parallel()

		// Piden el cotejo las evals que activan la skill y esperan algún cita
		// cotejar, lleve o no ROJ, y lo dicen sus comandos: en jurisprudencia, la
		// 04, la 06, la 08 y la 09; en normativa y en plazos-y-territorio, ninguna,
		// y sus umbrales no llevan documento_sin_cotejar (FR-001, FR-007 y FR-063
		// de H23.1; SC-005).
		carpetas := []struct {
			carpeta string
			piden   []string
		}{
			{carpeta: evalsDeNormativa},
			{carpeta: evalsDePlazosYTerritorio},
			{
				carpeta: evalsDeJurisprudencia,
				piden: []string{
					"04-documento-pegado.yaml", "06-documento-que-no-es-el-pedido.yaml",
					"08-doctrina-con-el-fallo-delante.yaml", "09-de-que-trata-con-la-ficha-sola.yaml",
				},
			},
		}

		for _, deLaCarpeta := range carpetas {
			leido, err := LeerConjunto(deLaCarpeta.carpeta)
			require.NoError(t, err)
			require.NotEmpty(t, leido.Evals, "%s tiene evals bien formadas", deLaCarpeta.carpeta)

			var piden []string

			for _, eval := range leido.Evals {
				if pideElCotejo(eval) {
					piden = append(piden, eval.Fichero)
				}
			}

			assert.Equal(t, deLaCarpeta.piden, piden, "las evals de %s que piden el cotejo", deLaCarpeta.carpeta)
		}
	})

	t.Run("normas-conocidas", func(t *testing.T) {
		t.Parallel()

		// Con la regla nombrada de otra forma, este subtest no vería sus defectos
		// y pasaría en vacío.
		require.True(t, slices.ContainsFunc(ReglasDeNormativa(), func(regla ReglaDelConjunto) bool {
			return regla.nombre == reglaDeNormasConocidas
		}), "la regla %q es una de las del conjunto", reglaDeNormasConocidas)

		desconocidas := slices.DeleteFunc(slices.Clone(defectos), func(defecto DefectoDelConjunto) bool {
			return !esDeNormasConocidas(defecto)
		})
		assert.Empty(t, desconocidas, "normas de las evals de %s que no están en %s:\n%s",
			evalsDeNormativa, tablaDeNormasDelRepositorio, presentarDefectos(desconocidas))
	})

	t.Run("grabado", func(t *testing.T) {
		t.Parallel()

		consultas := ConsultasNecesarias(conjunto.Evals)
		require.NotEmpty(t, consultas, "las evals de %s necesitan alguna consulta", evalsDeNormativa)

		// Desde H20, la caché preparada sirve también la lectura con --fecha
		// 9999-12-31 de cada bloque, y la unión de grabaciones la sirve sin red
		// (contracts/evals-redaccion-fechada.md §4 de H20; FR-053).
		exigeLecturasFechadas(t, consultas, evalsDeNormativa)

		dirCache := t.TempDir()

		preparadas, err := Preparar(dirCache, UnionDeGrabaciones(), consultas)
		require.NoError(t, err)
		assert.Empty(t, preparadas, "consultas de las evals de %s sin su respuesta en las grabaciones %s:\n%s",
			evalsDeNormativa, strings.Join(UnionDeGrabaciones(), " y "), presentarFaltas(preparadas))

		comprobadas, err := ComprobarSinRed(dirCache, consultas)
		require.NoError(t, err)
		assert.Empty(t, comprobadas, "consultas de las evals de %s que la caché preparada no sirve sin red:\n%s",
			evalsDeNormativa, presentarFaltas(comprobadas))
	})

	t.Run("avisos-del-esquema", func(t *testing.T) {
		t.Parallel()

		esquema, err := esquemaDeEval()
		require.NoError(t, err)

		assert.NoError(t, ComprobarCodigosDeAviso(esquema),
			"el esquema publicado %s admite en avisos exactamente los códigos de aviso del binario", rutaDelEsquemaDeEval)
	})

	t.Run("avisos-de-la-skill", func(t *testing.T) {
		t.Parallel()

		probarElControlDeContenido(t, controlDeLosAvisos)
	})

	t.Run("hallazgos-del-esquema", func(t *testing.T) {
		t.Parallel()

		esquema, err := esquemaDeEval()
		require.NoError(t, err)

		assert.NoError(t, ComprobarClasesDeHallazgo(esquema),
			"el esquema publicado %s admite en hallazgos exactamente las clases de hallazgo etiquetadas por el binario",
			rutaDelEsquemaDeEval)
	})

	t.Run("hallazgos-de-la-skill", func(t *testing.T) {
		t.Parallel()

		probarElControlDeContenido(t, controlDeLosHallazgos)
	})

	t.Run("formas-de-la-redaccion-fechada", func(t *testing.T) {
		t.Parallel()

		probarElControlDeContenido(t, controlDeLasFormasFechadas)
	})

	t.Run("cobertura-del-esquema", func(t *testing.T) {
		t.Parallel()

		// Con el vocabulario vacío, un esquema sin enumerado pasaría en vacío.
		require.NotEmpty(t, aspectosDeCobertura(), "el applet territorio tiene vocabulario de cobertura")

		esquema, err := esquemaDeEval()
		require.NoError(t, err)

		assert.NoError(t, ComprobarAspectosDeCobertura(esquema),
			"el esquema publicado %s admite en la cobertura del territorio exactamente el vocabulario del applet",
			rutaDelEsquemaDeEval)
	})

	t.Run("grafo-previo", probarGrafosPrevios)

	t.Run("prosa-de-la-skill", func(t *testing.T) {
		t.Parallel()

		probarElControlDeContenido(t, controlDeLaProsa(listaDelRepositorio(t, conjunto)))
	})

	t.Run("ordenes-para-powershell", func(t *testing.T) {
		t.Parallel()

		probarElControlDeContenido(t, controlDeLasOrdenes)
	})

	t.Run("linea-sin-consulta", func(t *testing.T) {
		t.Parallel()

		probarElControlDeContenido(t, controlDeLaLineaSinConsulta)
	})

	t.Run("lineas-de-aviso-de-las-referencias", func(t *testing.T) {
		t.Parallel()

		probarElControlDeContenido(t, controlDeLasLineasDeAviso)
	})

	t.Run("ordenes-de-ejemplo", probarLasOrdenesDeEjemplo)

	t.Run("lo-que-lleva-a-leer", probarLoQueLlevaALeer)

	t.Run("lecturas", probarLasLecturasDeLasEvalsDelRepositorio)
}

// probarLosPreceptosDeLaCobertura es el subtest precepto-de-la-cobertura de
// TestEvalsDelRepositorio (contracts/forma-de-cita.md §4 de H32; data-model §8
// de H32; research D16 de H32; FR-014): las evals de cobertura de
// plazos-y-territorio son la 07, la de los tributos, la 08, la del tráfico, y la
// 12, la de la Seguridad Social, cada una con el precepto de su fila, escrito
// aquí a mano —en las tres, la DA 1ª de la LPAC—; ese precepto es el de una fila
// de data/cobertura.yaml; y su mutante, la 12 con un bloque que ninguna fila
// tiene, da el defecto con su fichero y su precepto.
func probarLosPreceptosDeLaCobertura(t *testing.T) {
	t.Parallel()

	dePlazosYTerritorio, err := LeerConjunto(evalsDePlazosYTerritorio)
	require.NoError(t, err)

	cobertura, err := skills.LeerCobertura(contenidoDelFichero(t, coberturaDelRepositorio))
	require.NoError(t, err, "la cobertura por materia %s", coberturaDelRepositorio)

	daPrimera := CitaEsperada{Norma: "BOE-A-2015-10565", Bloque: "daprimera"}
	preceptos := map[string]CitaEsperada{}

	for _, eval := range dePlazosYTerritorio.Evals {
		if eval.esDeCobertura() {
			preceptos[eval.Fichero] = eval.NoCubierto.Precepto
		}
	}

	assert.Equal(t, map[string]CitaEsperada{
		evalDeLaLiquidacion: daPrimera,
		evalDeLaMulta:       daPrimera,
		ficheroDeLaEval12:   daPrimera,
	}, preceptos, "las evals de cobertura de %s, con el precepto de su fila", evalsDePlazosYTerritorio)

	assert.Empty(t, preceptosFueraDeLaCobertura(dePlazosYTerritorio.Evals, cobertura),
		"cada no_cubierto.precepto de %s es el de una fila de %s", evalsDePlazosYTerritorio, coberturaDelRepositorio)

	const bloqueDeNingunaFila = "a999"

	require.False(t, slices.ContainsFunc(cobertura.ReguladosAparte, func(fila skills.ReguladoAparte) bool {
		return fila.Precepto.Bloque == bloqueDeNingunaFila
	}), "premisa: ninguna fila de %s tiene el bloque %s", coberturaDelRepositorio, bloqueDeNingunaFila)

	mutadas := slices.Clone(dePlazosYTerritorio.Evals)
	posicion := slices.IndexFunc(mutadas, func(eval Eval) bool { return eval.Fichero == ficheroDeLaEval12 })
	require.NotEqual(t, -1, posicion, "%s está en %s", ficheroDeLaEval12, evalsDePlazosYTerritorio)

	mutadas[posicion].NoCubierto.Precepto.Bloque = bloqueDeNingunaFila

	assert.Equal(t, []string{
		ficheroDeLaEval12 + ": no_cubierto.precepto BOE-A-2015-10565, bloque a999 no es el precepto de ninguna fila " +
			"de data/cobertura.yaml",
	}, preceptosFueraDeLaCobertura(mutadas, cobertura))
}

// preceptosFueraDeLaCobertura son los defectos de las evals de cobertura cuyo
// no_cubierto.precepto no es el precepto —su norma y su bloque— de ninguna fila
// de la cobertura por materia, uno por eval, en su orden, con su fichero y su
// precepto (data-model §8 de H32; FR-014). Las demás evals no lo llevan y no se
// miran.
func preceptosFueraDeLaCobertura(evals []Eval, cobertura skills.Cobertura) []string {
	var defectos []string

	for _, eval := range evals {
		if !eval.esDeCobertura() {
			continue
		}

		precepto := eval.NoCubierto.Precepto
		deUnaFila := slices.ContainsFunc(cobertura.ReguladosAparte, func(fila skills.ReguladoAparte) bool {
			return fila.Precepto.Norma == precepto.Norma && fila.Precepto.Bloque == precepto.Bloque
		})

		if !deUnaFila {
			defectos = append(defectos, fmt.Sprintf("%s: no_cubierto.precepto %s, bloque %s no es el precepto de "+
				"ninguna fila de data/cobertura.yaml", eval.Fichero, precepto.Norma, precepto.Bloque))
		}
	}

	return defectos
}

// probarLasLecturasDeLasEvalsDelRepositorio es el subtest lecturas de
// TestEvalsDelRepositorio (contracts/formato-de-eval.md §3.3 de H26.1;
// data-model §8 de H26.1; FR-011): de las evals del repositorio, llevan
// lecturas la 14, la 15, la 16 y la 17 de normativa, cada una con sus
// bloques y los motivos del contrato, escritos aquí a mano, y ninguna otra; que
// la unión de grabaciones los sirve sin red lo exige el subtest grabado, porque
// son consultas necesarias (FR-010). Desde H31.1, la 14 y la 15 ganan cuatro
// cada una y la 17 la primera, todas con el mismo motivo
// (contracts/lecturas-y-grabaciones.md §1 de H31.1; FR-011).
func probarLasLecturasDeLasEvalsDelRepositorio(t *testing.T) {
	t.Parallel()

	const (
		trlrhl = "BOE-A-2004-4214"
		lirpf  = "BOE-A-2006-20764"
		ltaibg = "BOE-A-2013-12887"

		motivoDeLa14 = "De la misma norma que el art. 59 que pide la eval: la exploración que provoca la pregunta, " +
			"vista en los jobs de la #175 y la #178 (ADR 0043, clase b)."
		motivoDeLa15 = "Junto al art. 17 que pide la eval: la exploración que provoca la pregunta, vista en los jobs " +
			"de la #175 y la #178 (ADR 0043, clase b)."
		// motivoDeLoQueSeNombra es el de las nueve lecturas de H31.1, el mismo
		// para todas (contracts/lecturas-y-grabaciones.md §1 de H31.1; FR-011).
		motivoDeLoQueSeNombra = "Lo nombran, sin haberlo leído, respuestas de los sondeos de " +
			"evidencias/sondeos-20261025/ o de los jobs de las evals, y la regla de leer lo que se nombra lo pide " +
			"(H31.1; ADR 0043, clase b)."
	)

	carpetas := []struct {
		carpeta  string
		lecturas map[string][]LecturaDeLaEval
	}{
		{
			carpeta: evalsDeNormativa,
			lecturas: map[string][]LecturaDeLaEval{
				"14-trlrhl-impuestos-por-materia.yaml": {
					{Norma: trlrhl, Bloque: "a15", Motivo: motivoDeLa14},
					{Norma: trlrhl, Bloque: "a60", Motivo: motivoDeLoQueSeNombra},
					{Norma: trlrhl, Bloque: "a92", Motivo: motivoDeLoQueSeNombra},
					{Norma: trlrhl, Bloque: "a100", Motivo: motivoDeLoQueSeNombra},
					{Norma: trlrhl, Bloque: "a104", Motivo: motivoDeLoQueSeNombra},
				},
				"15-irpf-rendimientos-por-materia.yaml": {
					{Norma: lirpf, Bloque: "a18", Motivo: motivoDeLa15},
					{Norma: lirpf, Bloque: "a7", Motivo: motivoDeLoQueSeNombra},
					{Norma: lirpf, Bloque: "a19", Motivo: motivoDeLoQueSeNombra},
					{Norma: lirpf, Bloque: "a20", Motivo: motivoDeLoQueSeNombra},
					{Norma: lirpf, Bloque: "dadecimoctava", Motivo: motivoDeLoQueSeNombra},
				},
				ficheroDeLaEval16: {
					{Norma: normaDeLaLRJSP, Bloque: "a26", Motivo: motivoDeLasLecturasDeLa16},
					{Norma: normaDeLaLRJSP, Bloque: "a27", Motivo: motivoDeLasLecturasDeLa16},
				},
				"17-ltaibg-plazo-por-materia.yaml": {{Norma: ltaibg, Bloque: "a24", Motivo: motivoDeLoQueSeNombra}},
			},
		},
		{carpeta: evalsDePlazosYTerritorio, lecturas: map[string][]LecturaDeLaEval{}},
		{carpeta: evalsDeJurisprudencia, lecturas: map[string][]LecturaDeLaEval{}},
	}

	for _, deLaCarpeta := range carpetas {
		leido, err := LeerConjunto(deLaCarpeta.carpeta)
		require.NoError(t, err)
		require.NotEmpty(t, leido.Evals, "%s tiene evals bien formadas", deLaCarpeta.carpeta)

		lecturas := map[string][]LecturaDeLaEval{}

		for _, eval := range leido.Evals {
			if len(eval.Lecturas) > 0 {
				lecturas[eval.Fichero] = eval.Lecturas
			}
		}

		assert.Equal(t, deLaCarpeta.lecturas, lecturas, "las lecturas de las evals de %s", deLaCarpeta.carpeta)
	}
}

// probarLoQueLlevaALeer es el subtest lo-que-lleva-a-leer de
// TestEvalsDelRepositorio (contracts/cache-de-la-sesion.md §1 y §2 de H26.1;
// FR-001, FR-002, FR-005; SC-003, SC-008, SC-011): desde la caché preparada de
// una sesión de la primera eval de cada skill, cada bloque de data/plazos.yaml
// y de data/cobertura.yaml, sin fecha y con --fecha 9999-12-31, y el índice y
// los metadatos de cada norma de esos dos ficheros y de data/normas.yaml salen
// con 0 con --offline. Los enumera loQueLlevaALeerElRepositorio, el oráculo de
// los tests de los preceptos, no el código que prepara la caché.
func probarLoQueLlevaALeer(t *testing.T) {
	t.Parallel()

	lecturas := loQueLlevaALeerElRepositorio(t)
	require.NotEmpty(t, lecturas, "data/ lleva a leer algún bloque del BOE")

	for _, carpeta := range []string{evalsDeNormativa, evalsDePlazosYTerritorio, evalsDeJurisprudencia} {
		t.Run(filepath.Base(carpeta), func(t *testing.T) {
			t.Parallel()

			conjunto, err := LeerConjunto(carpeta)
			require.NoError(t, err)
			require.NotEmpty(t, conjunto.Evals, "%s tiene evals bien formadas", carpeta)

			fichero := conjunto.Evals[0].Fichero
			dirCache := cacheDeUnaSesion(t, carpeta, fichero)

			comprobadas, err := ComprobarSinRed(dirCache, lecturas)
			require.NoError(t, err)
			assert.Empty(t, comprobadas, "lo que la skill lleva a leer de data/ que la caché de una sesión de %s de %s "+
				"no sirve sin red:\n%s", fichero, carpeta, presentarFaltas(comprobadas))
		})
	}
}

const (
	// raizDeLasSkills es la raíz del repositorio, relativa al directorio de este
	// paquete: en su directorio skills/ está cada skill con su SKILL.md.
	raizDeLasSkills = "../.."

	// lineaSinConsultaDeLasSkills es la línea que cada SKILL.md enseña para la
	// respuesta de quien no tiene ni la herramienta ni el binario, con el
	// marcador de la causa, carácter a carácter (contracts/skills.md §3 de H21,
	// con lo que se dice a quien ya tiene kitlegal desde el 2026-10-20).
	lineaSinConsultaDeLasSkills = "⚠ SIN CONSULTA AL BOE: <causa>. Si ya tienes kitlegal, cierra Claude del todo, " +
		"ábrelo y pregunta en un chat nuevo; si no: https://kitlegal.es/instalar/"
)

// skillsConLineaSinConsulta son las skills que llevan la regla, las dos que
// consultan el BOE, y las que skills/ tiene que tener para que la subprueba
// linea-sin-consulta no pase en vacío para ninguna (FR-035 de H21).
// jurisprudencia no está: no consulta el BOE y, sin herramienta ni binario,
// lleva su propia línea (research.md D16 de H23).
var skillsConLineaSinConsulta = []string{skillNormativa, "plazos-y-territorio"}

// defectosDeLaLineaSinConsulta da los defectos de un SKILL.md respecto de la
// línea de contracts/skills.md §3 de H21, o nil si no tiene ninguno: que ningún
// bloque de código text sea, sin su sangría, exactamente esa línea; o que
// ExtraerSinConsulta no la reconozca, en el bloque tal como está escrito, con su
// dirección.
func defectosDeLaLineaSinConsulta(markdown string) []string {
	bloques := bloquesDeTexto(markdown)

	indice := slices.IndexFunc(bloques, func(bloque string) bool {
		return strings.TrimSpace(bloque) == lineaSinConsultaDeLasSkills
	})
	if indice < 0 {
		return []string{"SKILL.md no dice en un bloque text la línea «" + lineaSinConsultaDeLasSkills + "»"}
	}

	conLinea, conDireccion := ExtraerSinConsulta(bloques[indice])
	if !conLinea || !conDireccion {
		return []string{"el juicio no reconoce con su dirección la línea «" + lineaSinConsultaDeLasSkills +
			"» del bloque text de SKILL.md"}
	}

	return nil
}

// TestLineaSinConsultaDeUnaSkill fija la comprobación de la subprueba
// linea-sin-consulta, defectosDeLaLineaSinConsulta, sobre Markdown escrito aquí
// (contracts/skills.md §5 de H21; FR-035): la línea sola en un bloque text, con
// la sangría de un elemento de lista o sin ella, no tiene defectos; y sí los
// tiene sin la línea, con una palabra cambiada, sin su dirección, con otra línea
// en su mismo bloque, en la prosa o en un bloque que no es text.
func TestLineaSinConsultaDeUnaSkill(t *testing.T) {
	t.Parallel()

	const (
		regla     = "8. **Sin herramienta y sin binario, la respuesta lo dice.** La respuesta lleva esta línea:\n"
		noLaDice  = "SKILL.md no dice en un bloque text la línea «" + lineaSinConsultaDeLasSkills + "»"
		otraFrase = "No se ha podido comprobar si la redacción ha cambiado desde una consulta anterior."
	)

	enUnBloque := func(lenguaje, sangria string, lineas ...string) string {
		bloque := sangria + "```" + lenguaje + "\n"
		for _, linea := range lineas {
			bloque += sangria + linea + "\n"
		}

		return bloque + sangria + "```\n"
	}

	casos := []struct {
		nombre   string
		markdown string
		defectos []string
	}{
		{
			nombre:   "en-un-bloque-de-una-lista",
			markdown: enUnBloque("text", "   ", otraFrase) + regla + enUnBloque("text", "   ", lineaSinConsultaDeLasSkills),
		},
		{
			nombre:   "en-un-bloque-sin-sangria",
			markdown: regla + "\n" + enUnBloque("text", "", lineaSinConsultaDeLasSkills),
		},
		{
			nombre:   "sin-la-linea",
			markdown: regla + enUnBloque("text", "   ", otraFrase),
			defectos: []string{noLaDice},
		},
		{
			nombre: "con-una-palabra-cambiada",
			markdown: regla + enUnBloque("text", "   ",
				strings.Replace(lineaSinConsultaDeLasSkills, "un chat nuevo", "este chat", 1)),
			defectos: []string{noLaDice},
		},
		{
			nombre: "sin-su-direccion",
			markdown: regla + enUnBloque("text", "   ",
				strings.TrimSuffix(lineaSinConsultaDeLasSkills, " https://kitlegal.es/instalar/")),
			defectos: []string{noLaDice},
		},
		{
			nombre:   "con-otra-linea-en-su-bloque",
			markdown: regla + enUnBloque("text", "   ", lineaSinConsultaDeLasSkills, otraFrase),
			defectos: []string{noLaDice},
		},
		{
			nombre:   "en-la-prosa",
			markdown: regla + "   " + lineaSinConsultaDeLasSkills + "\n",
			defectos: []string{noLaDice},
		},
		{
			nombre:   "en-un-bloque-que-no-es-text",
			markdown: regla + enUnBloque("bash", "   ", lineaSinConsultaDeLasSkills),
			defectos: []string{noLaDice},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.defectos, defectosDeLaLineaSinConsulta(caso.markdown))
		})
	}
}

// Los informes del job de evals de normativa en el cierre de H7.1, de
// H7.2 y de H7.3 (el de 196ee05), relativos al directorio de este paquete: sus
// respuestas son con las que se calibró la lista de expresiones prohibidas
// (FR-021 de H7.3; FR-032 de H7.4; contracts/lista-de-expresiones.md §4 de
// H7.4). Desde H24 la lista no juzga ninguna respuesta y su reparto por eval y
// por familia ya no se exige (FR-071 de H24): de ese calibrado quedan las que la
// lista marca en cada informe, que son las que juzga TestJuzgarSinLaLista
// (FR-111 de H24). Están versionados, solo se leen y no se editan (FR-080 de
// H7.3; FR-090 de H7.4).
const (
	informeDeH71 = "../../specs/011-h7-1-graph-check-acotado/gates/evals/boe-legislacion.json"
	informeDeH72 = "../../specs/012-h7-2-la-consulta-repetida/gates/evals/boe-legislacion.json"
	informeDeH73 = "../../specs/013-h7-3-el-umbral-de/gates/evals/boe-legislacion.json"
)

// respuestasDeCadaInforme son las sesiones de cada uno de esos informes, una
// respuesta cada una.
const respuestasDeCadaInforme = 93

// entradaDelInforme es lo que TestJuzgarSinLaLista lee de cada sesión de uno de
// esos informes: su nombre, el fichero de su eval, si la eval espera que la
// skill se active, si la sesión la activó y su respuesta.
type entradaDelInforme struct {
	Sesion    string `json:"sesion"`
	Eval      string `json:"eval"`
	Activa    bool   `json:"activa"`
	Activada  bool   `json:"activada"`
	Respuesta string `json:"respuesta"`
}

// entradasDelInforme son las sesiones del informe versionado de la ruta, en su
// orden. El informe tiene sus 93 respuestas.
func entradasDelInforme(t *testing.T, ruta string) []entradaDelInforme {
	t.Helper()

	var informe struct {
		Evals []entradaDelInforme `json:"evals"`
	}
	require.NoError(t, json.Unmarshal(contenidoDelFichero(t, ruta), &informe), "%s es un informe del job de evals", ruta)
	require.Len(t, informe.Evals, respuestasDeCadaInforme, "el informe %s tiene sus respuestas", ruta)

	return informe.Evals
}

// listaDelRepositorio es la lista de expresiones prohibidas de
// evals/normativa/, la que deja en el conjunto el LeerConjunto de
// TestEvalsDelRepositorio. Tiene expresiones en sus cuatro familias, las de lo
// que devuelven las herramientas y formas fijas: sin ellas, los tests que la
// aplican pasarían en vacío.
func listaDelRepositorio(t *testing.T, conjunto Conjunto) ExpresionesProhibidas {
	t.Helper()

	lista := conjunto.Prohibidas
	require.NotEmpty(t, lista.Maquinaria, "%s tiene lista de expresiones prohibidas, con las de la maquinaria",
		evalsDeNormativa)
	require.NotEmpty(t, lista.OtraConversacion, "%s tiene lista de expresiones prohibidas, con las de otra conversación",
		evalsDeNormativa)
	require.NotEmpty(t, lista.Anuncio, "%s tiene lista de expresiones prohibidas, con las del anuncio",
		evalsDeNormativa)
	require.NotEmpty(t, lista.RedaccionNoLeida, "%s tiene lista de expresiones prohibidas, con las de la redacción "+
		"no leída", evalsDeNormativa)
	require.NotEmpty(t, lista.SalidaDeLasHerramientas, "%s tiene lista de expresiones prohibidas, con las de lo que "+
		"devuelven las herramientas", evalsDeNormativa)
	require.NotEmpty(t, lista.FormasFijas, "%s tiene lista de expresiones prohibidas, con sus formas fijas",
		evalsDeNormativa)

	return lista
}

// textoAMirar es un texto que no puede llevar ninguna expresión prohibida, con
// lo que lo nombra en un fallo: un párrafo de la prosa de un SKILL.md.
type textoAMirar struct {
	nombre string

	// texto es el párrafo con cada tramo de código en línea cambiado por un
	// espacio: en él se buscan las cuatro familias de la lista.
	texto string

	// conCodigo es el párrafo con cada tramo de código en línea sin sus acentos
	// graves y con su texto: en él se busca lo que devuelven las herramientas,
	// que un SKILL.md escribe como código (contracts/skill-boe-legislacion.md §4
	// de H24).
	conCodigo string
}

// expresionesEn da una línea por cada texto que lleva alguna expresión de la
// lista, con su nombre y las expresiones que lleva, en el orden de los textos; o
// nil si ninguno lleva ninguna. Las de las cuatro familias se buscan en el texto
// sin su código en línea, con ExtraerExpresionesProhibidas, y van delante; las
// de lo que devuelven las herramientas, en el texto con él, con
// salidaDeLasHerramientasEn.
func expresionesEn(textos []textoAMirar, lista ExpresionesProhibidas) []string {
	var lineas []string

	for _, texto := range textos {
		encontradas := slices.Concat(ExtraerExpresionesProhibidas(texto.texto, lista),
			salidaDeLasHerramientasEn(texto.conCodigo, lista))
		if len(encontradas) > 0 {
			lineas = append(lineas, texto.nombre+": "+strings.Join(encontradas, ", "))
		}
	}

	return lineas
}

// salidaDeLasHerramientasEn son las expresiones de la clave
// salida_de_las_herramientas de la lista que lleva el texto, en el orden de la
// lista y sin repetir, o nil si no lleva ninguna. La comparación es la de las
// familias, formaDeExpresion: las palabras de la expresión en su orden, sin
// distinguir mayúsculas y sin letra ni cifra a los lados, de modo que «del
// sobre» no es «el sobre» y fecha_vigencia_reciente sí lleva fecha_vigencia
// (contracts/skill-boe-legislacion.md §4 de H24; FR-085). No quita las formas
// fijas: se aplica a la prosa de un SKILL.md, no a una respuesta.
func salidaDeLasHerramientasEn(texto string, lista ExpresionesProhibidas) []string {
	var encontradas []string

	for _, expresion := range lista.SalidaDeLasHerramientas {
		if !slices.Contains(encontradas, expresion) && formasDeExpresiones.forma(expresion).MatchString(texto) {
			encontradas = append(encontradas, expresion)
		}
	}

	return encontradas
}

// articuloLeido es lo que estas pruebas miran de la data de boe articulo --json:
// el texto del bloque y su fecha de vigencia.
type articuloLeido struct {
	Texto         string `json:"texto"`
	FechaVigencia string `json:"fecha_vigencia"`
}

// leerBloque ejecuta con el registro boe articulo --json de la consulta y
// devuelve su data y lo que escribió en la salida de error, donde el kernel
// avisa de una entrega al grafo que falló. La lectura termina en 0 y su salida es
// un sobre con un bloque.
func leerBloque(t *testing.T, registro *app.Registro, consulta Consulta) (articuloLeido, string) {
	t.Helper()

	var salida, errores bytes.Buffer

	argv := slices.Concat([]string{programaDeLasConsultas, consulta.Applet, consulta.Verbo}, consulta.Argumentos,
		[]string{"--json"})
	codigo := app.Main(argv, registro, &salida, &errores,
		sinDatosDeConstruccion, sinDatosDeConstruccion, sinDatosDeConstruccion)
	require.Zero(t, codigo, "«%s» lee el bloque: %s", ordenDe(consulta), errores.String())

	var sobre struct {
		Data articuloLeido `json:"data"`
	}
	require.NoError(t, json.Unmarshal(salida.Bytes(), &sobre), "la salida de «%s» es un sobre con un bloque",
		ordenDe(consulta))

	return sobre.Data, errores.String()
}

// Las líneas que abren y cierran un bloque de código text de Markdown, sin la
// sangría con la que van dentro de una lista.
const (
	aperturaDeBloqueDeTexto = "```text"
	cierreDeBloqueDeCodigo  = "```"
)

// bloquesDeTexto son los contenidos de los bloques de código text del Markdown,
// en su orden: las líneas entre la que lo abre y la que lo cierra, sin sus saltos
// de línea finales y unidas por uno. Como en CommonMark, un bloque que no se
// cierra llega hasta el final del documento.
func bloquesDeTexto(markdown string) []string {
	var (
		bloques []string
		lineas  []string
		dentro  bool
	)

	for linea := range strings.Lines(markdown) {
		recortada := strings.TrimSpace(linea)

		switch {
		case !dentro && recortada == aperturaDeBloqueDeTexto:
			dentro, lineas = true, []string{}
		case dentro && recortada == cierreDeBloqueDeCodigo:
			bloques = append(bloques, strings.Join(lineas, "\n"))
			dentro = false
		case dentro:
			lineas = append(lineas, strings.TrimSuffix(linea, "\n"))
		}
	}

	if dentro {
		bloques = append(bloques, strings.Join(lineas, "\n"))
	}

	return bloques
}

// defectosDeLasFormasFechadas da los defectos de un SKILL.md respecto de las
// tres formas fijas de una redacción fechada (contracts/skill-boe-legislacion.md
// §1 de H20; FR-048), o nil si no tiene ninguno: uno por cada forma de
// formasFechadas, en su orden, que ningún bloque de código text lleva con la
// gramática con la que se juzga la respuesta (expresionesFechadas). Como la de
// los hallazgos, la comprobación es de la etiqueta: decirla con otras palabras,
// o en la prosa y no en un bloque text, no la enseña.
func defectosDeLasFormasFechadas(markdown string) []string {
	bloques := bloquesDeTexto(markdown)

	var defectos []string

	for _, forma := range formasFechadas {
		if !slices.ContainsFunc(bloques, expresionesFechadas()[forma].MatchString) {
			defectos = append(defectos, "ningún bloque text de SKILL.md lleva la forma fija "+formaEscrita(string(forma)))
		}
	}

	return defectos
}

// Lo que la prosa de un SKILL.md no es (contracts/skill-boe-legislacion.md §5):
// cada bloque delimitado, de la línea que empieza, tras la sangría, por su
// delimitador a la siguiente que empieza por él, y la región generada, de la
// línea que empieza por su marca de inicio a la que es su marca de fin. Esas
// líneas, marcas incluidas, parten la prosa en párrafos.
const (
	delimitadorDeBloque      = "```"
	inicioDeLaRegionGenerada = "<!-- inicio de la tabla de comandos"
	finDeLaRegionGenerada    = "<!-- fin de la tabla de comandos -->"
)

var (
	// elementoDeLista casa con la línea que abre un elemento de lista, `- `,
	// `* ` o `<n>. ` tras la sangría: empieza otro párrafo.
	elementoDeLista = regexp.MustCompile(`^[ \t]*(?:[-*]|[0-9]+\.) `)

	// codigoEnLinea casa con un tramo de código en línea, que en la prosa se
	// cambia por un espacio para que no junte las palabras de sus lados. El
	// grupo es su texto, sin los acentos graves: lo que queda de él donde se
	// busca lo que devuelven las herramientas.
	codigoEnLinea = regexp.MustCompile("`([^`]*)`")

	// fechaConCifras casa con una fecha AAAAMMDD escrita con cifras: ocho
	// cifras, sin otra delante ni detrás, con un mes de 01 a 12 y un día de 01 a
	// 31 (FR-013). El grupo es la fecha.
	fechaConCifras = regexp.MustCompile(`(?:^|[^0-9])([0-9]{4}(?:0[1-9]|1[0-2])(?:0[1-9]|[12][0-9]|3[01]))(?:$|[^0-9])`)
)

// defectosDeLaProsa da una línea por cada párrafo de la prosa del Markdown
// (parrafosDeLaProsa) que lleva alguna expresión de la lista, con el número de
// su primera línea y las expresiones, en el orden del fichero, y detrás una por
// cada línea del fichero entero, sin quitar nada, que lleva una fecha AAAAMMDD
// escrita con cifras, con su número y sus fechas; o nil si no hay ninguna. Las
// expresiones de un párrafo son las de expresionesEn: las de las cuatro
// familias, que no miran su código en línea, y las de lo que devuelven las
// herramientas, que sí.
func defectosDeLaProsa(markdown string, lista ExpresionesProhibidas) []string {
	defectos := expresionesEn(parrafosDeLaProsa(markdown), lista)

	numero := 0
	for linea := range strings.Lines(markdown) {
		numero++

		var fechas []string
		for _, casada := range fechaConCifras.FindAllStringSubmatch(linea, -1) {
			fechas = append(fechas, casada[1])
		}

		if len(fechas) > 0 {
			defectos = append(defectos, fmt.Sprintf("línea %d: fecha escrita con cifras %s", numero,
				strings.Join(fechas, ", ")))
		}
	}

	return defectos
}

// parrafosDeLaProsa son los párrafos de la prosa del Markdown entero,
// frontmatter incluido, en su orden, cada uno nombrado por el número de su
// primera línea (contracts/skill-boe-legislacion.md §5): sin los bloques
// delimitados ni la región generada, partido en párrafos por las líneas en
// blanco, por las de esos bloques y de esa región y por cada línea que abre un
// elemento de lista, con las líneas de cada párrafo juntas con un espacio, sin su
// sangría, y cada tramo de código en línea cambiado por un espacio. Cada párrafo
// va además con su código en línea (conCodigo): cada tramo, sin sus acentos
// graves y con su texto (contracts/skill-boe-legislacion.md §4 de H24). Como en
// CommonMark, un bloque o una región que no se cierran llegan hasta el final.
func parrafosDeLaProsa(markdown string) []textoAMirar {
	var (
		parrafos           []textoAMirar
		lineas             []string
		primera, numero    int
		enBloque, enRegion bool
	)

	cerrarElParrafo := func() {
		if len(lineas) > 0 {
			parrafo := strings.Join(lineas, " ")
			parrafos = append(parrafos, textoAMirar{
				nombre:    fmt.Sprintf("párrafo de la línea %d", primera),
				texto:     codigoEnLinea.ReplaceAllString(parrafo, " "),
				conCodigo: codigoEnLinea.ReplaceAllString(parrafo, "${1}"),
			})
		}

		lineas = nil
	}

	for linea := range strings.Lines(markdown) {
		numero++
		recortada := strings.TrimSpace(linea)

		switch {
		case enRegion:
			enRegion = recortada != finDeLaRegionGenerada
		case enBloque:
			enBloque = !strings.HasPrefix(recortada, delimitadorDeBloque)
		case strings.HasPrefix(recortada, inicioDeLaRegionGenerada):
			cerrarElParrafo()
			enRegion = true
		case strings.HasPrefix(recortada, delimitadorDeBloque):
			cerrarElParrafo()
			enBloque = true
		case recortada == "":
			cerrarElParrafo()
		default:
			if elementoDeLista.MatchString(linea) {
				cerrarElParrafo()
			}

			if len(lineas) == 0 {
				primera = numero
			}

			lineas = append(lineas, recortada)
		}
	}

	cerrarElParrafo()

	return parrafos
}

// TestProsaDeLaSkill fija la extracción de la subprueba prosa-de-la-skill,
// defectosDeLaProsa, sobre Markdown escrito aquí (contracts/skill-boe-legislacion.md
// §5; FR-091, SC-005): una expresión en un tramo de código, en un bloque
// delimitado —también con sangría— o en la región generada no cuenta; partida
// por un salto de línea dentro de un párrafo o de un elemento de lista, sí; en el
// frontmatter, sí; dos párrafos no se juntan, y los parten una línea en blanco,
// la que abre un elemento de lista, un bloque delimitado y las marcas de la
// región; cada párrafo se nombra por su primera línea y lleva sus expresiones en
// el orden de la lista. Una fecha AAAAMMDD con cifras cuenta en cualquier parte
// del fichero, también dentro de un bloque; AAAAMMDD, un identificador BOE-A-…,
// nueve cifras o un mes o un día imposibles, no.
//
// Desde H24 (contracts/skill-boe-legislacion.md §4; FR-085, FR-110, SC-010),
// fija también lo que devuelven las herramientas, la clave
// salida_de_las_herramientas de la lista, que se busca en toda la prosa con su
// código en línea: los dos párrafos de SKILL.md v0.1.6 que señala el hito, tal
// cual, dan su defecto con sus expresiones; el nombre de un campo en un tramo de
// código cuenta, y graph check en otro del mismo párrafo sigue sin contar; en un
// bloque delimitado y en la región generada no cuenta ninguna; «del sobre», «el
// mismo sobre» y «sobre el texto» no son «el sobre», y fecha_vigencia_reciente sí
// lleva fecha_vigencia; no distingue mayúsculas; y en un párrafo van detrás de
// las de las familias.
func TestProsaDeLaSkill(t *testing.T) {
	t.Parallel()

	// Los dos párrafos de SKILL.md v0.1.6 con las frases que señala FR-085 de
	// H24, con sus líneas y su sangría: el de la vigencia, del paso 5, y el de la
	// redacción superada, de «Redacción modificada».
	const (
		vigenciaDeLaVersionAnterior = "- Traslada cada aviso de vigencia del sobre con su forma fija: `⚠`, la " +
			"etiqueta del aviso tal como la da el binario y\n" +
			"  dos puntos, seguidos de la frase del binario o de una explicación (más en «Cómo se cita»). De la " +
			"vigencia del bloque,\n" +
			"  la respuesta dice lo que trae el sobre de `kitlegal boe`: sus avisos y, de la redacción leída, qué " +
			"norma la dio\n" +
			"  (`norma_modificadora`) y desde cuándo rige (`fecha_vigencia`). Hasta cuándo, nunca: ningún sobre " +
			"trae el fin de una\n" +
			"  redacción, tampoco en una norma derogada, cuyo aviso no lleva fecha, y darlo sería texto legal sin " +
			"fuente. El de\n" +
			"  `kitlegal graph check` no dice nada de ella. Recuerda que los textos consolidados del BOE tienen " +
			"carácter informativo.\n"
		redaccionSuperadaDeLaVersionAnterior = "- **La redacción superada no la has leído.** `kitlegal boe " +
			"articulo` da solo la redacción vigente, y\n" +
			"  `kitlegal graph check`, dos fechas: nada de lo que devuelven dice qué decía la redacción superada " +
			"ni en qué se\n" +
			"  diferencia de la vigente. La respuesta no lo dice, ni lo resume, ni lo compara, aunque creas " +
			"saberlo: sería texto\n" +
			"  legal sin fuente. Sí dice lo que da la lectura: el texto vigente con su cita, qué norma le dio esa " +
			"redacción\n" +
			"  (`norma_modificadora`) y desde cuándo rige (`fecha_vigencia`); hasta cuándo, nunca (paso 5).\n"
	)

	lista := ExpresionesProhibidas{
		Maquinaria:              []string{"memoria de consultas", "hallazgos", "graph check"},
		OtraConversacion:        []string{"te dije"},
		Anuncio:                 []string{"redacto la respuesta"},
		SalidaDeLasHerramientas: []string{"el sobre", "fecha_vigencia", "norma_modificadora"},
	}
	region := "<!-- inicio de la tabla de comandos: generada desde --describe con make skills-sync, no editar -->\n"
	finDeLaRegion := "<!-- fin de la tabla de comandos -->\n"

	casos := []struct {
		nombre   string
		markdown string
		defectos []string
	}{
		{
			nombre:   "en-un-tramo-de-codigo",
			markdown: "Comprueba con `kitlegal graph check` y lee `data.hallazgos`.\n",
		},
		{
			nombre: "en-un-bloque-delimitado",
			markdown: "Por ejemplo:\n\n```text\nSin hallazgos en la memoria de consultas.\n```\n\n" +
				"Y nada más.\n",
		},
		{
			nombre:   "en-un-bloque-delimitado-con-sangria",
			markdown: "- Comprueba:\n\n  ```bash\n  kitlegal graph check BOE-A-2015-10565 a21 --json\n  ```\n",
		},
		{
			nombre:   "en-la-region-generada",
			markdown: "Antes.\n\n" + region + "\n| kitlegal graph check | hallazgos |\n\n" + finDeLaRegion + "\nDespués.\n",
		},
		{
			nombre:   "partida-dentro-de-un-parrafo",
			markdown: "Una segunda lectura apagaría lo que la memoria de\nconsultas tiene que decirte.\n",
			defectos: []string{"párrafo de la línea 1: memoria de consultas"},
		},
		{
			nombre:   "partida-dentro-de-un-elemento-de-lista",
			markdown: "Lee cada bloque:\n\n- una segunda lectura apagaría la memoria de\n  consultas.\n",
			defectos: []string{"párrafo de la línea 3: memoria de consultas"},
		},
		{
			nombre: "en-el-frontmatter",
			markdown: "---\nname: normativa\ndescription: >-\n  Traslada los hallazgos.\n---\n\n" +
				"# Consultar\n",
			defectos: []string{"párrafo de la línea 1: hallazgos"},
		},
		{
			nombre:   "dos-parrafos-no-se-juntan",
			markdown: "Lo que la memoria de\n\nconsultas dice.\n",
		},
		{
			nombre:   "dos-elementos-de-lista-no-se-juntan",
			markdown: "- lo que la memoria de\n* consultas dice\n7. redacto la\n  8. respuesta\n",
		},
		{
			nombre:   "un-bloque-delimitado-parte-el-parrafo",
			markdown: "Lo que la memoria de\n```text\nnada\n```\nconsultas dice.\n",
		},
		{
			nombre:   "las-marcas-de-la-region-parten-el-parrafo",
			markdown: "Lo que la memoria de\n" + region + finDeLaRegion + "consultas dice.\n",
		},
		{
			nombre: "cada-parrafo-con-sus-expresiones",
			markdown: "Primero.\n\nSin hallazgos en la memoria de consultas; te dije. Redacto la\nrespuesta.\n\n" +
				"Luego, graph check.\n",
			defectos: []string{
				"párrafo de la línea 3: memoria de consultas, hallazgos, te dije, redacto la respuesta",
				"párrafo de la línea 6: graph check",
			},
		},
		{
			nombre:   "el-codigo-deja-un-espacio",
			markdown: "Sin`--json`hallazgos.\n",
			defectos: []string{"párrafo de la línea 1: hallazgos"},
		},
		{
			nombre:   "una-fecha-con-cifras",
			markdown: "Sustituida por la de 20250101, que es la que se cita.\n",
			defectos: []string{"línea 1: fecha escrita con cifras 20250101"},
		},
		{
			nombre:   "una-fecha-en-un-bloque-o-en-codigo",
			markdown: "Por ejemplo:\n\n```text\nla de 20161002 y la de 20250101\n```\n\nO `20180309`.\n",
			defectos: []string{
				"línea 4: fecha escrita con cifras 20161002, 20250101",
				"línea 7: fecha escrita con cifras 20180309",
			},
		},
		{
			nombre:   "expresiones-y-fechas",
			markdown: "Los hallazgos de\n\n20200206.\n",
			defectos: []string{"párrafo de la línea 1: hallazgos", "línea 3: fecha escrita con cifras 20200206"},
		},
		{
			nombre: "no-son-fechas",
			markdown: "Las fechas van como `AAAAMMDD` o AAAAMMDD, en BOE-A-2015-10565, no 123456789, " +
				"20251301 ni 20250132.\n",
		},
		{
			nombre:   "la-vigencia-de-la-version-anterior",
			markdown: vigenciaDeLaVersionAnterior,
			defectos: []string{"párrafo de la línea 1: el sobre, fecha_vigencia, norma_modificadora"},
		},
		{
			nombre:   "la-redaccion-superada-de-la-version-anterior",
			markdown: redaccionSuperadaDeLaVersionAnterior,
			defectos: []string{"párrafo de la línea 1: fecha_vigencia, norma_modificadora"},
		},
		{
			nombre:   "un-campo-en-un-tramo-de-codigo",
			markdown: "Di desde cuándo rige (`fecha_vigencia`) y comprueba con `kitlegal graph check`.\n",
			defectos: []string{"párrafo de la línea 1: fecha_vigencia"},
		},
		{
			nombre: "la-salida-en-un-bloque-delimitado",
			markdown: "Por ejemplo:\n\n```text\nel sobre lleva fecha_vigencia y norma_modificadora\n```\n\n" +
				"Y nada más.\n",
		},
		{
			nombre: "la-salida-en-la-region-generada",
			markdown: "Antes.\n\n" + region + "\n| `boe_articulo` | el sobre, con `fecha_vigencia` y " +
				"`norma_modificadora` |\n\n" + finDeLaRegion + "\nDespués.\n",
		},
		{
			nombre:   "no-es-el-sobre",
			markdown: "Cada aviso del sobre, el mismo sobre y una afirmación sobre el texto.\n",
		},
		{
			nombre:   "un-campo-dentro-de-otro",
			markdown: "La fecha de la que acabas de leer (`fecha_vigencia_reciente`).\n",
			defectos: []string{"párrafo de la línea 1: fecha_vigencia"},
		},
		{
			nombre:   "la-salida-sin-distinguir-mayusculas",
			markdown: "El sobre trae los avisos.\n",
			defectos: []string{"párrafo de la línea 1: el sobre"},
		},
		{
			nombre:   "la-salida-detras-de-las-familias",
			markdown: "Si el sobre trae hallazgos, lee `norma_modificadora`.\n",
			defectos: []string{"párrafo de la línea 1: hallazgos, el sobre, norma_modificadora"},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.defectos, defectosDeLaProsa(caso.markdown, lista))
		})
	}
}

// Las palabras de una orden de lectura y comprobación de SKILL.md
// (contracts/skill-boe-legislacion.md §4 de H7.4): la lectura con kitlegal boe
// y uno de sus dos verbos, la comprobación con kitlegal graph check, la bandera
// que cierra cada una, lo que las une en Bash y, en PowerShell, lo que cierra la
// lectura, la condición y el cierre de su bloque.
var (
	lecturaDeLaOrden      = []string{"kitlegal", "boe"}
	verbosDeLectura       = []string{verboArticulo, "articulos"}
	comprobacionDeLaOrden = []string{"kitlegal", "graph", "check"}
	condicionDePowerShell = []string{"if", "($LASTEXITCODE", "-eq", "0)", "{"}
)

const (
	banderaDeLaOrden           = "--json"
	unionDeBash                = "&&"
	finDeLecturaDePowerShell   = banderaDeLaOrden + ";"
	cierreDeBloqueDePowerShell = "}"
)

// La orden de lectura y comprobación de un bloque, en sus dos formas, y lo que
// dice el defecto de la de Bash a la que no sigue la suya: las escriben
// TestOrdenesParaPowerShell y los mutantes de TestControlesSobreLasReferenciasAMano.
const (
	bashDeArticulo = "kitlegal boe articulo BOE-A-2015-10565 a21 --json && " +
		"kitlegal graph check BOE-A-2015-10565 a21 --json"
	powershellDeArticulo = "kitlegal boe articulo BOE-A-2015-10565 a21 --json; if ($LASTEXITCODE -eq 0) { " +
		"kitlegal graph check BOE-A-2015-10565 a21 --json }"

	sinSuForma = ": no la sigue su forma para PowerShell, «…; if ($LASTEXITCODE -eq 0) { … }», con el mismo " +
		"verbo, la misma norma y los mismos bloques"
)

// ordenDeLectura es una orden de lectura y comprobación escrita en el código de
// un SKILL.md, en cualquiera de sus dos formas (contracts/skill-boe-legislacion.md
// §4 de H7.4).
type ordenDeLectura struct {
	// linea es el número de la línea del Markdown en que está, y texto, la orden
	// tal como la escribe, sin la sangría.
	linea int
	texto string

	// powershell dice que es la forma para PowerShell, con la condición en lugar
	// de &&.
	powershell bool

	// verbo, norma y bloques son los de la lectura; normaComprobada y
	// bloquesComprobados, los de kitlegal graph check, vacíos si no los lleva.
	verbo              string
	norma              string
	bloques            []string
	normaComprobada    string
	bloquesComprobados []string
}

// nombre es como la nombra un defecto: su línea y su texto.
func (o ordenDeLectura) nombre() string {
	return fmt.Sprintf("línea %d, «%s»", o.linea, o.texto)
}

// leerOrdenDeLectura lee como orden de lectura y comprobación un tramo de código:
// sus palabras, separadas por blancos, son kitlegal boe, articulo o articulos, la
// norma, al menos un bloque y --json, y detrás, o && y kitlegal graph check con
// sus argumentos y --json —la forma de Bash—, o, con --json; en lugar de --json,
// if ($LASTEXITCODE -eq 0) { y kitlegal graph check con sus argumentos, --json y }
// —la de PowerShell—. Cualquier otro tramo no es una orden de lectura, y conForma
// es falso.
func leerOrdenDeLectura(linea int, tramo string) (orden ordenDeLectura, conForma bool) {
	palabras := strings.Fields(tramo)

	resto, conLectura := sinPrefijo(palabras, lecturaDeLaOrden)
	if !conLectura || len(resto) == 0 || !slices.Contains(verbosDeLectura, resto[0]) {
		return ordenDeLectura{}, false
	}

	partes, conPartes := partirLaOrden(resto[1:])
	argumentos, conComprobacion := argumentosDeLaComprobacion(partes.comprobacion)

	// La lectura lleva la norma y al menos un bloque.
	if !conPartes || !conComprobacion || len(partes.lectura) < 2 {
		return ordenDeLectura{}, false
	}

	orden = ordenDeLectura{
		linea: linea, texto: strings.Join(palabras, " "), powershell: partes.powershell,
		verbo: resto[0], norma: partes.lectura[0], bloques: partes.lectura[1:],
	}

	if len(argumentos) > 0 {
		orden.normaComprobada, orden.bloquesComprobados = argumentos[0], argumentos[1:]
	}

	return orden, true
}

// partesDeLaOrden son las palabras de una orden de lectura y comprobación
// detrás del verbo, partidas: las de la lectura, sin su --json, y las de la
// comprobación, sin lo que las une a la lectura ni, en PowerShell, el cierre de su
// bloque; y si es la forma para PowerShell.
type partesDeLaOrden struct {
	lectura      []string
	comprobacion []string
	powershell   bool
}

// partirLaOrden parte las palabras que siguen al verbo por lo que une la lectura
// a la comprobación: --json y && en Bash; en PowerShell, --json; y la condición,
// con el cierre de su bloque al final. conPartes es falso si no las une ninguna
// de las dos formas.
func partirLaOrden(palabras []string) (partes partesDeLaOrden, conPartes bool) {
	if fin := slices.Index(palabras, unionDeBash); fin > 0 && palabras[fin-1] == banderaDeLaOrden {
		return partesDeLaOrden{lectura: palabras[:fin-1], comprobacion: palabras[fin+1:]}, true
	}

	fin := slices.Index(palabras, finDeLecturaDePowerShell)
	if fin < 0 {
		return partesDeLaOrden{}, false
	}

	enElBloque, conCondicion := sinPrefijo(palabras[fin+1:], condicionDePowerShell)
	if !conCondicion || len(enElBloque) == 0 || enElBloque[len(enElBloque)-1] != cierreDeBloqueDePowerShell {
		return partesDeLaOrden{}, false
	}

	return partesDeLaOrden{lectura: palabras[:fin], comprobacion: enElBloque[:len(enElBloque)-1], powershell: true}, true
}

// argumentosDeLaComprobacion son los argumentos de kitlegal graph check en las
// palabras de la comprobación, sin su --json; conComprobacion es falso si no son
// kitlegal graph check con sus argumentos y --json al final.
func argumentosDeLaComprobacion(palabras []string) (argumentos []string, conComprobacion bool) {
	argumentos, conComprobacion = sinPrefijo(palabras, comprobacionDeLaOrden)
	if !conComprobacion || len(argumentos) == 0 || argumentos[len(argumentos)-1] != banderaDeLaOrden {
		return nil, false
	}

	return argumentos[:len(argumentos)-1], true
}

// sinPrefijo son las palabras sin el prefijo, y conPrefijo dice si empezaban por
// él.
func sinPrefijo(palabras, prefijo []string) (resto []string, conPrefijo bool) {
	if len(palabras) < len(prefijo) || !slices.Equal(palabras[:len(prefijo)], prefijo) {
		return nil, false
	}

	return palabras[len(prefijo):], true
}

// ordenesDeLectura son las órdenes de lectura y comprobación del Markdown, en su
// orden: cada tramo de código (tramosDeCodigo) que leerOrdenDeLectura lee como
// una.
func ordenesDeLectura(markdown string) []ordenDeLectura {
	var ordenes []ordenDeLectura

	for _, tramo := range tramosDeCodigo(markdown) {
		if orden, conForma := leerOrdenDeLectura(tramo.linea, tramo.texto); conForma {
			ordenes = append(ordenes, orden)
		}
	}

	return ordenes
}

// tramoDeCodigo es un tramo de código del Markdown: el número de la línea en
// que está y su texto.
type tramoDeCodigo struct {
	linea int
	texto string
}

// tramosDeCodigo son los tramos de código del Markdown en que se buscan sus
// órdenes, en su orden: cada línea de un bloque delimitado, sin la sangría, y
// cada tramo de código en línea fuera de ellos, sin sus acentos graves. Como en
// CommonMark, un bloque que no se cierra llega hasta el final.
func tramosDeCodigo(markdown string) []tramoDeCodigo {
	var (
		tramos   []tramoDeCodigo
		numero   int
		enBloque bool
	)

	for linea := range strings.Lines(markdown) {
		numero++
		recortada := strings.TrimSpace(linea)

		switch {
		case strings.HasPrefix(recortada, delimitadorDeBloque):
			enBloque = !enBloque
		case enBloque:
			tramos = append(tramos, tramoDeCodigo{linea: numero, texto: recortada})
		default:
			for _, codigo := range codigoEnLinea.FindAllString(recortada, -1) {
				tramos = append(tramos, tramoDeCodigo{linea: numero, texto: strings.Trim(codigo, "`")})
			}
		}
	}

	return tramos
}

// separadoresDeOrdenes son lo que termina una orden de ejemplo en su tramo de
// código, también pegado a una palabra: lo que une dos órdenes en Bash, lo que
// las separa en PowerShell y la tubería (contracts/controles-de-contenido.md §10
// de H31).
var separadoresDeOrdenes = []string{unionDeBash, ";", "|"}

// marcadorDeEjemplo casa con lo que hace de una orden un ejemplo por concretar:
// un marcador <…>, tres puntos o unos puntos suspensivos
// (contracts/controles-de-contenido.md §10 de H31).
var marcadorDeEjemplo = regexp.MustCompile(`<[^<>]*>|\.\.\.|…`)

// ordenDeEjemplo es una orden de ejemplo concreta de kitlegal boe del Markdown
// (contracts/controles-de-contenido.md §10 de H31): el número de la línea en que
// está y la consulta que hace, sin --json, que le añade ComprobarSinRed.
type ordenDeEjemplo struct {
	linea    int
	consulta Consulta
}

// ordenesDeEjemplo son las órdenes de ejemplo concretas de kitlegal boe del
// Markdown, en su orden (contracts/controles-de-contenido.md §10 de H31; research
// D17 de H31; FR-019): en cada tramo de código de los que lee ordenesDeLectura
// (tramosDeCodigo), cada aparición de las palabras kitlegal boe con las que la
// siguen hasta el primer &&, ; o | (palabrasHastaElSeparador), sin --json, si
// son un verbo y al menos un argumento y no llevan ningún marcador
// (marcadorDeEjemplo). La mención del verbo sin argumentos, la orden con un
// marcador y la que está en la prosa no son órdenes concretas, y no se ejecutan.
func ordenesDeEjemplo(markdown string) []ordenDeEjemplo {
	var ordenes []ordenDeEjemplo

	for _, tramo := range tramosDeCodigo(markdown) {
		palabras := strings.Fields(tramo.texto)

		for posicion := range palabras {
			resto, conLectura := sinPrefijo(palabras[posicion:], lecturaDeLaOrden)
			if !conLectura {
				continue
			}

			orden := palabrasHastaElSeparador(resto)
			if len(orden) < 2 || marcadorDeEjemplo.MatchString(strings.Join(orden, " ")) {
				continue
			}

			ordenes = append(ordenes, ordenDeEjemplo{
				linea:    tramo.linea,
				consulta: Consulta{Applet: appletDeLasNormas, Verbo: orden[0], Argumentos: orden[1:]},
			})
		}
	}

	return ordenes
}

// palabrasHastaElSeparador son las palabras hasta el primer separador de
// órdenes (separadoresDeOrdenes), sin --json: de la palabra a la que va pegado,
// lo que tiene delante.
func palabrasHastaElSeparador(palabras []string) []string {
	var hasta []string

	for _, palabra := range palabras {
		fin := len(palabra)
		for _, separador := range separadoresDeOrdenes {
			if posicion := strings.Index(palabra, separador); posicion >= 0 {
				fin = min(fin, posicion)
			}
		}

		if antes := palabra[:fin]; antes != "" && antes != banderaDeLaOrden {
			hasta = append(hasta, antes)
		}

		if fin < len(palabra) {
			return hasta
		}
	}

	return hasta
}

// defectosDeLasOrdenesParaPowerShell da una línea por cada defecto de las
// órdenes de lectura y comprobación del Markdown de un SKILL.md
// (contracts/skill-boe-legislacion.md §4 de H7.4; FR-024, FR-095, SC-005): primero,
// por cada verbo de lectura, si ninguna orden de Bash lo usa; después, en el orden
// del fichero, por cada orden de Bash cuya siguiente orden no es la de PowerShell
// con su mismo verbo, su misma norma y sus mismos bloques —porque falta, porque es
// otra de Bash o porque lee otra cosa—, y por cada orden, de las dos formas, cuya
// comprobación no lleva la norma y los bloques de su lectura. Una orden sin la
// condición de PowerShell o con otra no es ninguna de las dos formas: la de Bash
// que la precede no tiene la suya. Nil si no hay ningún defecto. Desde H31, las
// dos partes las aplica por separado el control de la skill entera
// (controlDeLasOrdenes): la primera, a las órdenes de todos sus ficheros, y la
// segunda, a las de cada uno.
func defectosDeLasOrdenesParaPowerShell(markdown string) []string {
	ordenes := ordenesDeLectura(markdown)

	return slices.Concat(verbosSinOrdenDeBash(ordenes), defectosDeLasOrdenes(ordenes))
}

// verbosSinOrdenDeBash da una línea por cada verbo de lectura, en su orden, que
// ninguna orden de Bash de las órdenes usa, o nil si las hay de todos.
func verbosSinOrdenDeBash(ordenes []ordenDeLectura) []string {
	var defectos []string

	for _, verbo := range verbosDeLectura {
		if !slices.ContainsFunc(ordenes, func(orden ordenDeLectura) bool {
			return !orden.powershell && orden.verbo == verbo
		}) {
			defectos = append(defectos, fmt.Sprintf("ninguna orden de Bash lee con «kitlegal boe %s» y comprueba "+
				"detrás con kitlegal graph check", verbo))
		}
	}

	return defectos
}

// defectosDeLasOrdenes da, en su orden, una línea por cada orden de Bash de las
// órdenes de un fichero a la que no sigue su forma para PowerShell y otra por
// cada orden cuya comprobación no lleva la norma y los bloques de su lectura, o
// nil si no hay ninguna.
func defectosDeLasOrdenes(ordenes []ordenDeLectura) []string {
	var defectos []string

	for posicion, orden := range ordenes {
		if !orden.powershell && (posicion+1 == len(ordenes) || !esSuFormaParaPowerShell(orden, ordenes[posicion+1])) {
			defectos = append(defectos, orden.nombre()+": no la sigue su forma para PowerShell, «…; if "+
				"($LASTEXITCODE -eq 0) { … }», con el mismo verbo, la misma norma y los mismos bloques")
		}

		if orden.normaComprobada != orden.norma || !slices.Equal(orden.bloquesComprobados, orden.bloques) {
			defectos = append(defectos, orden.nombre()+": kitlegal graph check no lleva la norma y los bloques "+
				"de la lectura")
		}
	}

	return defectos
}

// esSuFormaParaPowerShell dice si siguiente es la forma para PowerShell de la
// orden de Bash: de PowerShell, con su mismo verbo, su misma norma y sus mismos
// bloques en su orden.
func esSuFormaParaPowerShell(bash, siguiente ordenDeLectura) bool {
	return siguiente.powershell && siguiente.verbo == bash.verbo && siguiente.norma == bash.norma &&
		slices.Equal(siguiente.bloques, bash.bloques)
}

// TestOrdenesParaPowerShell fija defectosDeLasOrdenesParaPowerShell, la
// comprobación de la subprueba ordenes-para-powershell, sobre Markdown escrito
// aquí (contracts/skill-boe-legislacion.md §4 de H7.4; FR-024, FR-095, SC-005):
// las dos órdenes con sus dos formas, en bloques o en código en línea, no tienen
// defectos; sin la forma de PowerShell, con && en ella, sin el if, con otra
// condición, con otra norma o con otros bloques en la de PowerShell, cada una
// nombra su orden de Bash; con otros bloques o sin argumentos en la comprobación,
// nombra la orden que la lleva; y sin la orden de articulos, lo dice. El texto de
// kitlegal boe articulo en la prosa no es una orden.
func TestOrdenesParaPowerShell(t *testing.T) {
	t.Parallel()

	const (
		bashDeArticulos = "kitlegal boe articulos <norma> <bloques>... --json && " +
			"kitlegal graph check <norma> <bloques>... --json"
		powershellDeArticulos = "kitlegal boe articulos <norma> <bloques>... --json; if ($LASTEXITCODE -eq 0) { " +
			"kitlegal graph check <norma> <bloques>... --json }"

		otraComprobacion = ": kitlegal graph check no lleva la norma y los bloques de la lectura"
	)

	// skill es un SKILL.md con la orden de articulo y la de articulos, cada una
	// en un bloque bash seguido de uno powershell y en una lista, como las enseña
	// la skill; las líneas de las órdenes son la 6, la 12, la 18 y la 22.
	skill := func(articulo, powershellDeArticulo, articulos, powershellDeArticulos string) string {
		return "## Protocolo\n\n- Lee los bloques de uno en uno con `kitlegal boe articulo`:\n\n" +
			"  ```bash\n  " + articulo + "\n  ```\n\n" +
			"  En PowerShell, la misma orden es:\n\n" +
			"  ```powershell\n  " + powershellDeArticulo + "\n  ```\n\n" +
			"  Varios bloques a la vez:\n\n" +
			"  ```bash\n  " + articulos + "\n  ```\n\n" +
			"  ```powershell\n  " + powershellDeArticulos + "\n  ```\n"
	}
	nombrada := func(linea int, orden string) string { return fmt.Sprintf("línea %d, «%s»", linea, orden) }

	// La de PowerShell escrita con && es la de Bash.
	conAmpersands := bashDeArticulo
	sinElIf := "kitlegal boe articulo BOE-A-2015-10565 a21 --json; kitlegal graph check BOE-A-2015-10565 a21 --json"
	otraCondicion := strings.Replace(powershellDeArticulo, "($LASTEXITCODE -eq 0)", "($?)", 1)
	otraNorma := strings.ReplaceAll(powershellDeArticulo, "BOE-A-2015-10565", "BOE-A-2017-12902")
	otrosBloques := strings.ReplaceAll(powershellDeArticulo, "a21", "a22")
	otraComprobacionEnBash := strings.Replace(bashDeArticulo, "check BOE-A-2015-10565 a21",
		"check BOE-A-2015-10565 a22", 1)
	otraComprobacionEnPowerShell := strings.Replace(powershellDeArticulo, "check BOE-A-2015-10565 a21",
		"check BOE-A-2015-10565 a21 a22", 1)
	sinArgumentos := strings.Replace(bashDeArticulos, "check <norma> <bloques>... --json", "check --json", 1)

	casos := []struct {
		nombre   string
		markdown string
		defectos []string
	}{
		{
			nombre:   "las-dos-ordenes-con-sus-dos-formas",
			markdown: skill(bashDeArticulo, powershellDeArticulo, bashDeArticulos, powershellDeArticulos),
		},
		{
			nombre: "en-codigo-en-linea",
			markdown: "Lee con `kitlegal boe articulo` (`" + bashDeArticulo + "`; en PowerShell, `" +
				powershellDeArticulo + "`) o con `" + bashDeArticulos + "`, en PowerShell `" +
				powershellDeArticulos + "`.\n",
		},
		{
			nombre:   "sin-la-de-powershell",
			markdown: skill(bashDeArticulo, "", bashDeArticulos, powershellDeArticulos),
			defectos: []string{nombrada(6, bashDeArticulo) + sinSuForma},
		},
		{
			nombre:   "sin-la-ultima-de-powershell",
			markdown: skill(bashDeArticulo, powershellDeArticulo, bashDeArticulos, ""),
			defectos: []string{nombrada(18, bashDeArticulos) + sinSuForma},
		},
		{
			// La de PowerShell con && es otra de Bash: tampoco la sigue la suya.
			nombre:   "con-ampersands-en-la-de-powershell",
			markdown: skill(bashDeArticulo, conAmpersands, bashDeArticulos, powershellDeArticulos),
			defectos: []string{nombrada(6, bashDeArticulo) + sinSuForma, nombrada(12, conAmpersands) + sinSuForma},
		},
		{
			nombre:   "sin-el-if",
			markdown: skill(bashDeArticulo, sinElIf, bashDeArticulos, powershellDeArticulos),
			defectos: []string{nombrada(6, bashDeArticulo) + sinSuForma},
		},
		{
			nombre:   "con-otra-condicion",
			markdown: skill(bashDeArticulo, otraCondicion, bashDeArticulos, powershellDeArticulos),
			defectos: []string{nombrada(6, bashDeArticulo) + sinSuForma},
		},
		{
			nombre:   "con-otra-norma-en-la-de-powershell",
			markdown: skill(bashDeArticulo, otraNorma, bashDeArticulos, powershellDeArticulos),
			defectos: []string{nombrada(6, bashDeArticulo) + sinSuForma},
		},
		{
			nombre:   "con-otros-bloques-en-la-de-powershell",
			markdown: skill(bashDeArticulo, otrosBloques, bashDeArticulos, powershellDeArticulos),
			defectos: []string{nombrada(6, bashDeArticulo) + sinSuForma},
		},
		{
			nombre:   "con-otros-bloques-en-la-comprobacion-de-bash",
			markdown: skill(otraComprobacionEnBash, powershellDeArticulo, bashDeArticulos, powershellDeArticulos),
			defectos: []string{nombrada(6, otraComprobacionEnBash) + otraComprobacion},
		},
		{
			nombre:   "con-otros-bloques-en-la-comprobacion-de-powershell",
			markdown: skill(bashDeArticulo, otraComprobacionEnPowerShell, bashDeArticulos, powershellDeArticulos),
			defectos: []string{nombrada(12, otraComprobacionEnPowerShell) + otraComprobacion},
		},
		{
			nombre:   "sin-argumentos-en-la-comprobacion",
			markdown: skill(bashDeArticulo, powershellDeArticulo, sinArgumentos, powershellDeArticulos),
			defectos: []string{nombrada(18, sinArgumentos) + otraComprobacion},
		},
		{
			nombre:   "sin-la-de-articulos",
			markdown: skill(bashDeArticulo, powershellDeArticulo, "", ""),
			defectos: []string{
				"ninguna orden de Bash lee con «kitlegal boe articulos» y comprueba detrás con kitlegal graph check",
			},
		},
		{
			nombre:   "sin-ninguna-orden",
			markdown: "Lee los bloques con `kitlegal boe articulo` y `kitlegal boe articulos`.\n",
			defectos: []string{
				"ninguna orden de Bash lee con «kitlegal boe articulo» y comprueba detrás con kitlegal graph check",
				"ninguna orden de Bash lee con «kitlegal boe articulos» y comprueba detrás con kitlegal graph check",
			},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.defectos, defectosDeLasOrdenesParaPowerShell(caso.markdown))
		})
	}
}

// TestOrdenesDeEjemplo fija ordenesDeEjemplo, el extractor del control 10, el de
// la subprueba ordenes-de-ejemplo (contracts/controles-de-contenido.md §10 de
// H31; research D17 de H31; FR-019, SC-005), sobre Markdown escrito aquí: la
// orden de un bloque, en su forma de Bash, hasta &&, también con la sangría de
// una lista, y en la de PowerShell, hasta ;, y la de código en línea, cada una
// con su línea, sin --json y con su --fecha y su día; la que sigue a una tubería,
// hasta |; y las dos de un mismo tramo. Ninguna de la que lleva un marcador, <…>,
// ... o …, de la mención del verbo sin argumentos, también con solo --json, ni de
// la que está en la prosa.
func TestOrdenesDeEjemplo(t *testing.T) {
	t.Parallel()

	const (
		enSuFechaEnBash = "kitlegal boe articulo BOE-A-1985-5392 a22 --fecha 2005-06-15 --json && " +
			"kitlegal graph check BOE-A-1985-5392 a22 --json"
		enSuFechaEnPowerShell = "kitlegal boe articulo BOE-A-1985-5392 a22 --fecha 2005-06-15 --json; " +
			"if ($LASTEXITCODE -eq 0) { kitlegal graph check BOE-A-1985-5392 a22 --json }"
	)

	// enBloque es un Markdown con la orden en la línea 6, en un bloque del
	// lenguaje con la sangría.
	enBloque := func(sangria, lenguaje, orden string) string {
		return "## Lectura\n\n- Lee el bloque en su fecha:\n\n" +
			sangria + "```" + lenguaje + "\n" + sangria + orden + "\n" + sangria + "```\n"
	}
	deLaLinea := func(linea int, verbo string, argumentos ...string) ordenDeEjemplo {
		return ordenDeEjemplo{
			linea:    linea,
			consulta: Consulta{Applet: appletDeLasNormas, Verbo: verbo, Argumentos: argumentos},
		}
	}
	enSuFecha := []string{"BOE-A-1985-5392", "a22", "--fecha", "2005-06-15"}

	casos := []struct {
		nombre   string
		markdown string
		ordenes  []ordenDeEjemplo
	}{
		{
			nombre:   "en-un-bloque-en-su-forma-de-bash",
			markdown: enBloque("", "bash", enSuFechaEnBash),
			ordenes:  []ordenDeEjemplo{deLaLinea(6, verboArticulo, enSuFecha...)},
		},
		{
			nombre:   "en-un-bloque-de-una-lista",
			markdown: enBloque("  ", "bash", enSuFechaEnBash),
			ordenes:  []ordenDeEjemplo{deLaLinea(6, verboArticulo, enSuFecha...)},
		},
		{
			nombre:   "en-un-bloque-en-su-forma-de-powershell",
			markdown: enBloque("", "powershell", enSuFechaEnPowerShell),
			ordenes:  []ordenDeEjemplo{deLaLinea(6, verboArticulo, enSuFecha...)},
		},
		{
			nombre: "en-codigo-en-linea",
			markdown: "# Fechas\n\nLee el bloque en su fecha con " +
				"`kitlegal boe articulo BOE-A-1985-5392 a22 --fecha 2005-06-15 --json`, o con `boe_articulo`.\n",
			ordenes: []ordenDeEjemplo{deLaLinea(3, verboArticulo, enSuFecha...)},
		},
		{
			nombre:   "tras-una-tuberia",
			markdown: enBloque("", "bash", "kitlegal boe indice BOE-A-2015-10565 --json | jq '.data.bloques'"),
			ordenes:  []ordenDeEjemplo{deLaLinea(6, verboIndice, "BOE-A-2015-10565")},
		},
		{
			nombre: "dos-en-un-tramo",
			markdown: enBloque("", "bash", "kitlegal boe indice BOE-A-2015-10565 --json && "+
				"kitlegal boe articulo BOE-A-2015-10565 a21 --json"),
			ordenes: []ordenDeEjemplo{
				deLaLinea(6, verboIndice, "BOE-A-2015-10565"),
				deLaLinea(6, verboArticulo, "BOE-A-2015-10565", "a21"),
			},
		},
		{
			nombre:   "con-un-marcador-entre-angulos",
			markdown: enBloque("", "bash", "kitlegal boe buscar <términos> --json"),
		},
		{
			nombre: "con-un-marcador-en-la-fecha",
			markdown: "Léelo con `kitlegal boe articulo BOE-A-1985-5392 a22 --fecha <inicio> --json`, sin " +
				"`kitlegal graph check`.\n",
		},
		{
			nombre:   "con-tres-puntos",
			markdown: enBloque("", "bash", "kitlegal boe articulos BOE-A-2015-10565 a21 a22... --json"),
		},
		{
			nombre:   "con-puntos-suspensivos",
			markdown: enBloque("", "bash", "kitlegal boe articulos BOE-A-2015-10565 a21 … --json"),
		},
		{
			nombre: "la-mencion-del-verbo",
			markdown: "Lo citado sale de `kitlegal boe articulo` o `articulos`. Si `kitlegal boe` falla, " +
				"dilo.\n",
		},
		{
			nombre:   "el-verbo-con-solo-json",
			markdown: enBloque("", "bash", "kitlegal boe articulo --json"),
		},
		{
			nombre:   "en-la-prosa",
			markdown: "Lee el artículo con kitlegal boe articulo BOE-A-2015-10565 a21 --json y cítalo.\n",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.ordenes, ordenesDeEjemplo(caso.markdown))
		})
	}
}

// ficheroDeUnaSkill es un fichero de la skill entera: su ruta en el directorio
// de la skill, SKILL.md o references/<fichero>.md, como la nombra un defecto, y
// su texto.
type ficheroDeUnaSkill struct {
	ruta  string
	texto string
}

// skillEntera es lo que leen de una skill los controles de contenido desde H31
// (contracts/controles-de-contenido.md de H31; research D16 de H31; FR-016): su
// nombre, su SKILL.md y sus referencias a mano (skills.ReferenciasAMano), en
// orden de nombre, cada fichero por separado. Las referencias generadas no se
// leen.
type skillEntera struct {
	nombre      string
	skillMd     ficheroDeUnaSkill
	referencias []ficheroDeUnaSkill
}

// leerLaSkillEntera lee la skill entera del nombre en el repositorio de raiz.
func leerLaSkillEntera(t *testing.T, raiz, nombre string) skillEntera {
	t.Helper()

	cargada, err := skills.Cargar(raiz, nombre)
	require.NoError(t, err)

	aMano, err := skills.ReferenciasAMano(os.DirFS(raiz), nombre)
	require.NoError(t, err, "las referencias a mano de %s", nombre)

	skill := skillEntera{
		nombre:  nombre,
		skillMd: ficheroDeUnaSkill{ruta: ficheroDeLaSkill, texto: string(cargada.Contenido)},
	}

	for _, referencia := range aMano {
		skill.referencias = append(skill.referencias, ficheroDeUnaSkill{
			ruta:  path.Join(directorioDeLasReferencias, referencia.Fichero),
			texto: string(referencia.Contenido),
		})
	}

	return skill
}

// ficheros son el SKILL.md y, detrás, las referencias a mano.
func (s skillEntera) ficheros() []ficheroDeUnaSkill {
	return slices.Concat([]ficheroDeUnaSkill{s.skillMd}, s.referencias)
}

// defectosDe son los defectos de un fichero de la skill, cada uno con la skill y
// la ruta del fichero delante; nil sin ninguno.
func (s skillEntera) defectosDe(fichero ficheroDeUnaSkill, defectos []string) []string {
	return conPrefijo(s.nombre+": "+fichero.ruta+": ", defectos)
}

// defectosDeLaSkill son los defectos de la skill entera, o los que ya nombran su
// fichero, cada uno con la skill delante; nil sin ninguno.
func (s skillEntera) defectosDeLaSkill(defectos []string) []string {
	return conPrefijo(s.nombre+": ", defectos)
}

// conPrefijo son los defectos, cada uno con el prefijo delante; nil sin ninguno.
func conPrefijo(prefijo string, defectos []string) []string {
	var conElPrefijo []string

	for _, defecto := range defectos {
		conElPrefijo = append(conElPrefijo, prefijo+defecto)
	}

	return conElPrefijo
}

// controlDeContenido es uno de los controles 1 a 7 de
// contracts/controles-de-contenido.md de H31 (research D16 de H31; FR-016 a
// FR-018): lo que exige, como lo dice su fallo; las skills en que se aplica, las
// de antes de H31; y sus defectos sobre la skill entera de una de ellas, una
// línea cada uno con la skill y, si es de un fichero, el fichero; nil sin
// ninguno.
type controlDeContenido struct {
	exige    string
	skills   []string
	defectos func(skill skillEntera) []string
}

// probarElControlDeContenido aplica el control a la skill entera de cada una de
// sus skills del repositorio, que skills/ tiene que tener para que no pase en
// vacío en ninguna, y exige que no dé ningún defecto.
func probarElControlDeContenido(t *testing.T, control controlDeContenido) {
	t.Helper()

	nombres, err := skills.Listar(raizDeLasSkills)
	require.NoError(t, err)
	require.Subset(t, nombres, control.skills, "skills/ tiene las skills en que se aplica el control")

	var defectos []string

	for _, nombre := range control.skills {
		defectos = append(defectos, control.defectos(leerLaSkillEntera(t, raizDeLasSkills, nombre))...)
	}

	assert.Empty(t, defectos, "en %s, %s:\n%s", strings.Join(control.skills, " y "), control.exige,
		strings.Join(defectos, "\n"))
}

// controlDeLaProsa es el control 1, el de la subprueba prosa-de-la-skill
// (contracts/skill-boe-legislacion.md §5 de H7.3; research D4 de H7.3; FR-010,
// FR-013, FR-091, SC-005 de H7.3): con defectosDeLaProsa, ningún párrafo de la
// prosa de ningún fichero, frontmatter incluido, lleva ninguna expresión de la
// lista, tampoco de lo que devuelven las herramientas escrita como código en
// línea (contracts/skill-boe-legislacion.md §4 de H24; FR-085, FR-110, SC-010
// de H24), y ninguna línea lleva una fecha AAAAMMDD escrita con cifras. SKILL.md
// tiene prosa: sin ella, el control pasaría en vacío.
func controlDeLaProsa(lista ExpresionesProhibidas) controlDeContenido {
	return controlDeContenido{
		exige: "ningún fichero enseña con su prosa expresiones que la respuesta no lleva ni escribe con cifras una " +
			"fecha AAAAMMDD, y SKILL.md tiene prosa",
		skills: []string{skillNormativa},
		defectos: func(skill skillEntera) []string {
			var defectos []string

			if len(parrafosDeLaProsa(skill.skillMd.texto)) == 0 {
				defectos = skill.defectosDe(skill.skillMd,
					[]string{"no tiene prosa fuera del código y de la región generada"})
			}

			for _, fichero := range skill.ficheros() {
				defectos = append(defectos, skill.defectosDe(fichero, defectosDeLaProsa(fichero.texto, lista))...)
			}

			return defectos
		},
	}
}

// controlDeLasOrdenes es el control 3, el de la subprueba
// ordenes-para-powershell (contracts/skill-boe-legislacion.md §4 de H7.4;
// FR-024, FR-095, SC-005 de H7.4): cada orden de lectura y comprobación de un
// fichero lleva detrás, en ese mismo fichero, su forma para PowerShell —una orden
// de un fichero no se empareja con la de otro—, y hay una orden de Bash de cada
// verbo de lectura en la skill entera (FR-018 de H31). Los defectos de los
// verbos, que son de la skill entera, van delante.
var controlDeLasOrdenes = controlDeContenido{
	exige: "cada orden de lectura y comprobación lleva detrás, en su fichero, su forma para PowerShell, y la skill " +
		"tiene una orden de Bash de cada verbo de lectura",
	skills: []string{skillNormativa},
	defectos: func(skill skillEntera) []string {
		var (
			todas         []ordenDeLectura
			deLosFicheros []string
		)

		for _, fichero := range skill.ficheros() {
			ordenes := ordenesDeLectura(fichero.texto)
			todas = append(todas, ordenes...)
			deLosFicheros = append(deLosFicheros, skill.defectosDe(fichero, defectosDeLasOrdenes(ordenes))...)
		}

		return append(skill.defectosDeLaSkill(verbosSinOrdenDeBash(todas)), deLosFicheros...)
	},
}

// controlDeLaLineaSinConsulta es el control 2, el de la subprueba
// linea-sin-consulta (contracts/skills.md §5 de H21; FR-035, FR-077 de H21): el
// SKILL.md de cada skill que lleva la regla dice la línea de contracts/skills.md
// §3 de H21 tal cual en un bloque de código text, y esa línea casa con lo que el
// juicio reconoce, ExtraerSinConsulta, con su dirección. Es una forma fija que
// las evals comprueban sin modelo: en una referencia sola no vale (FR-018 de
// H31).
var controlDeLaLineaSinConsulta = controlDeContenido{
	exige:  "SKILL.md enseña la línea " + formaEscrita(etiquetaSinConsulta) + " en un bloque text",
	skills: skillsConLineaSinConsulta,
	defectos: func(skill skillEntera) []string {
		return skill.defectosDeLaSkill(defectosDeLaLineaSinConsulta(skill.skillMd.texto))
	},
}

// controlDeLasFormasFechadas es el control 4, el de la subprueba
// formas-de-la-redaccion-fechada (contracts/skill-boe-legislacion.md §1 de H20;
// FR-048 de H20): con defectosDeLasFormasFechadas, el SKILL.md enseña en un
// bloque text cada una de las tres formas fijas de una redacción fechada; en una
// referencia sola no vale (FR-018 de H31).
var controlDeLasFormasFechadas = controlDeContenido{
	exige:  "SKILL.md enseña en un bloque text la forma fija de cada línea de una redacción fechada",
	skills: []string{skillNormativa},
	defectos: func(skill skillEntera) []string {
		return skill.defectosDeLaSkill(defectosDeLasFormasFechadas(skill.skillMd.texto))
	},
}

// controlDeLosAvisos es el control 5, el de la subprueba avisos-de-la-skill
// (FR-014 de H5.1): el SKILL.md lleva la forma fija de cada código de aviso del
// binario, reconocida con la misma función que usa Juzgar; en una referencia sola
// no vale (FR-018 de H31).
var controlDeLosAvisos = controlDeContenido{
	exige:  "SKILL.md enseña la forma fija de cada código de aviso del binario",
	skills: []string{skillNormativa},
	defectos: func(skill skillEntera) []string {
		return skill.defectosDe(skill.skillMd, lineasDelError(ComprobarFormasDeAviso(skill.skillMd.texto)))
	},
}

// controlDeLosHallazgos es el control 6, el de la subprueba
// hallazgos-de-la-skill (FR-045, FR-093 de H7.1): el SKILL.md lleva la forma fija
// de cada clase de hallazgo etiquetada por el binario, reconocida con la misma
// función que usa Juzgar; en una referencia sola no vale (FR-018 de H31).
var controlDeLosHallazgos = controlDeContenido{
	exige:  "SKILL.md enseña la forma fija de cada clase de hallazgo etiquetada por el binario",
	skills: []string{skillNormativa},
	defectos: func(skill skillEntera) []string {
		return skill.defectosDe(skill.skillMd, lineasDelError(ComprobarFormasDeHallazgo(skill.skillMd.texto)))
	},
}

// lineasDelError son las líneas del mensaje del error, una por cada error que
// junta errors.Join, o nil sin error.
func lineasDelError(err error) []string {
	if err == nil {
		return nil
	}

	return strings.Split(err.Error(), "\n")
}

// skillsConLineasDeAviso son las skills en que se aplica el control 7: las dos
// cuyas formas fijas, las líneas ⚠, comprueban sin modelo sus evals (research
// D16 de H31).
var skillsConLineasDeAviso = []string{skillNormativa, "plazos-y-territorio"}

// inicioDeLineaDeAviso casa con la línea que, tras su sangría, un marcador de
// lista (-, * o <n>.) con sus blancos y el énfasis (*, _ o `), cada uno si lo
// lleva, empieza por la marca ⚠ (contracts/controles-de-contenido.md §7 de H31).
var inicioDeLineaDeAviso = regexp.MustCompile("^[ \t]*(?:(?:[-*]|[0-9]+\\.)[ \t]+)?[*_`]*⚠")

// controlDeLasLineasDeAviso es el control 7, el de la subprueba
// lineas-de-aviso-de-las-referencias (contracts/controles-de-contenido.md §7 de
// H31; research D16 de H31; FR-018), con defectosDeLasLineasDeAviso.
var controlDeLasLineasDeAviso = controlDeContenido{
	exige:    "cada línea ⚠ de una referencia a mano es una de SKILL.md",
	skills:   skillsConLineasDeAviso,
	defectos: defectosDeLasLineasDeAviso,
}

// defectosDeLasLineasDeAviso da, en el orden de las referencias a mano y de sus
// líneas, una línea por cada línea ⚠ de una referencia (inicioDeLineaDeAviso)
// que no es, sin los blancos de los extremos, ninguna de SKILL.md, también sin
// ellos, con su número y su texto; o nil si no hay ninguna. Una variante mal
// escrita de una forma fija —una errata, otra puntuación, otro marcador— no lo
// es, aunque la gramática con la que se juzga la respuesta la reconociera.
func defectosDeLasLineasDeAviso(skill skillEntera) []string {
	delSkillMd := map[string]bool{}
	for linea := range strings.Lines(skill.skillMd.texto) {
		delSkillMd[strings.TrimSpace(linea)] = true
	}

	var defectos []string

	for _, referencia := range skill.referencias {
		numero := 0

		for linea := range strings.Lines(referencia.texto) {
			numero++

			if recortada := strings.TrimSpace(linea); inicioDeLineaDeAviso.MatchString(linea) && !delSkillMd[recortada] {
				defectos = append(defectos, fmt.Sprintf("%s: %s:%d: la línea ⚠ no es ninguna de SKILL.md: «%s»",
					skill.nombre, referencia.ruta, numero, recortada))
			}
		}
	}

	return defectos
}

// controlDeLasOrdenesDeEjemplo es el control 10, el de la subprueba
// ordenes-de-ejemplo (contracts/controles-de-contenido.md §10 de H31; research
// D17 de H31; FR-019, SC-005): en cada una de las skills, cada orden de ejemplo
// concreta de cada fichero (ordenesDeEjemplo) sale con 0 con --offline,
// ComprobarSinRed, sobre la caché de una sesión de su skill, la que da cacheDe,
// que solo se pide para la skill que tiene alguna. Cada Falta es un defecto con
// la skill y el fichero, y el error que impide comprobar las del fichero,
// también.
func controlDeLasOrdenesDeEjemplo(conEvals []string, cacheDe func(skill string) string) controlDeContenido {
	return controlDeContenido{
		exige: "cada orden de ejemplo concreta de kitlegal boe sale con 0 con --offline sobre la caché de una sesión " +
			"de su skill",
		skills: conEvals,
		defectos: func(skill skillEntera) []string {
			var (
				defectos []string
				dirCache string
			)

			for _, fichero := range skill.ficheros() {
				ordenes := ordenesDeEjemplo(fichero.texto)
				if len(ordenes) == 0 {
					continue
				}

				if dirCache == "" {
					dirCache = cacheDe(skill.nombre)
				}

				consultas := make([]Consulta, 0, len(ordenes))
				for _, orden := range ordenes {
					consultas = append(consultas, orden.consulta)
				}

				faltas, err := ComprobarSinRed(dirCache, consultas)
				if err != nil {
					defectos = append(defectos, skill.defectosDe(fichero, []string{err.Error()})...)

					continue
				}

				for _, falta := range faltas {
					defectos = append(defectos, skill.defectosDe(fichero, []string{falta.String()})...)
				}
			}

			return defectos
		},
	}
}

// skillsConEvals son las skills de skills/ del repositorio, en su orden, que
// tienen su carpeta en evals/: en ellas se aplica el control 10.
func skillsConEvals(t *testing.T) []string {
	t.Helper()

	nombres, err := skills.Listar(raizDeLasSkills)
	require.NoError(t, err)

	var conEvals []string

	for _, nombre := range nombres {
		info, err := os.Stat(filepath.Join(directorioDeEvals, nombre))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}

		require.NoError(t, err, "la carpeta de evals de %s", nombre)

		if info.IsDir() {
			conEvals = append(conEvals, nombre)
		}
	}

	return conEvals
}

// cacheDeLaPrimeraEvalServida es la caché de una sesión de la skill de la
// carpeta de evals contra la que se ejecutan sus órdenes de ejemplo
// (contracts/controles-de-contenido.md §10 de H31): la de cacheDeUnaSesion para
// la primera eval de la carpeta, por nombre, que no declara sin_servir.
func cacheDeLaPrimeraEvalServida(t *testing.T, carpeta string) string {
	t.Helper()

	conjunto, err := LeerConjunto(carpeta)
	require.NoError(t, err)

	var primera string

	for _, eval := range conjunto.Evals {
		if len(eval.SinServir) == 0 && (primera == "" || eval.Fichero < primera) {
			primera = eval.Fichero
		}
	}

	require.NotEmpty(t, primera, "%s tiene una eval bien formada que no declara sin_servir", carpeta)

	return cacheDeUnaSesion(t, carpeta, primera)
}

// probarLasOrdenesDeEjemplo es el subtest ordenes-de-ejemplo de
// TestEvalsDelRepositorio (contracts/controles-de-contenido.md §10 de H31;
// research D17 de H31; FR-019, SC-005): el control 10 en cada skill con evals,
// sobre la caché de la primera eval de su carpeta que no declara sin_servir; y
// el repositorio tiene al menos una orden de ejemplo concreta, para que el
// control no pase en vacío.
func probarLasOrdenesDeEjemplo(t *testing.T) {
	t.Parallel()

	conEvals := skillsConEvals(t)
	probarElControlDeContenido(t, controlDeLasOrdenesDeEjemplo(conEvals, func(skill string) string {
		return cacheDeLaPrimeraEvalServida(t, filepath.Join(directorioDeEvals, skill))
	}))

	ordenes := 0

	for _, nombre := range conEvals {
		for _, fichero := range leerLaSkillEntera(t, raizDeLasSkills, nombre).ficheros() {
			ordenes += len(ordenesDeEjemplo(fichero.texto))
		}
	}

	assert.Positive(t, ordenes, "las skills con evals, %s, tienen alguna orden de ejemplo concreta de kitlegal boe",
		strings.Join(conEvals, ", "))
}

// mutanteDeUnControl es un caso de TestControlesSobreLasReferenciasAMano: el
// control, la skill de la copia, lo que le hace a la copia, que recibe el
// directorio de la skill, y los defectos que el control tiene que dar con ella:
// con exactos, esos y ninguno más, en su orden; con prefijos, tantos como esos,
// en su orden, y cada uno empieza por el suyo, para el defecto que termina con
// lo que la invocación escribió en la salida de error, con la hora de su registro;
// sin ninguno de los dos, al menos esos; y sin ninguno, ninguno.
type mutanteDeUnControl struct {
	nombre   string
	control  controlDeContenido
	skill    string
	mutar    func(t *testing.T, dir string)
	defectos []string
	exactos  bool
	prefijos bool
}

// TestControlesSobreLasReferenciasAMano es el de los mutantes de los controles
// 1 a 7 y 10 de contracts/controles-de-contenido.md de H31 (research D16 y D17 de
// H31; FR-016 a FR-019; SC-005): sobre una copia de la skill en un temporal, cada
// mutante lleva la regla de un control del SKILL.md a una referencia a mano, o
// la escribe en ella, y el control da su defecto, con la skill y el fichero.
// Las de ausencia, en la referencia, como en SKILL.md: una expresión prohibida y
// una fecha AAAAMMDD con cifras en su prosa, y una orden de lectura sin su forma
// para PowerShell, también cuando esa forma está en otro fichero; la orden de
// Bash de un verbo se busca en la skill entera, así que llevar las de articulos a
// una referencia no es un defecto y quitarlas sí. Las formas fijas, en SKILL.md:
// llevar a una referencia la línea sin consulta, cada forma de aviso, cada una
// de hallazgo o cada una de la redacción fechada da el defecto de su control.
// Y una línea ⚠ de una referencia que no es, sin los blancos de los extremos,
// una de SKILL.md —una errata, un marcador de lista o un énfasis— es un defecto
// con su línea; la que sí lo es, o la que no empieza por la marca, no. Una
// orden de ejemplo de una referencia que la caché de una sesión no sirve es un
// defecto, y la que lleva un marcador no se ejecuta. Sin mutar, la copia no da
// ningún defecto del control: si no, el mutante pasaría por lo que ya estaba.
func TestControlesSobreLasReferenciasAMano(t *testing.T) {
	t.Parallel()

	conjunto, err := LeerConjunto(evalsDeNormativa)
	require.NoError(t, err)

	mutantes := slices.Concat(mutantesDeLaProsa(listaDelRepositorio(t, conjunto)), mutantesDeLasOrdenes(),
		mutantesDeLaLineaSinConsulta(), mutantesDeLasFormasMovidas(), mutantesDeLasLineasDeAviso(t),
		mutantesDeLasOrdenesDeEjemplo(t))

	for _, mutante := range mutantes {
		t.Run(mutante.nombre, func(t *testing.T) {
			t.Parallel()

			require.Contains(t, mutante.control.skills, mutante.skill, "el control se aplica en %s", mutante.skill)

			raiz := copiaDeLaSkill(t, mutante.skill)
			require.Empty(t, mutante.control.defectos(leerLaSkillEntera(t, raiz, mutante.skill)),
				"sin mutar, la copia de %s no da ningún defecto del control", mutante.skill)

			mutante.mutar(t, filepath.Join(raiz, directorioDeLasSkills, mutante.skill))
			defectos := mutante.control.defectos(leerLaSkillEntera(t, raiz, mutante.skill))

			switch {
			case len(mutante.defectos) == 0:
				assert.Empty(t, defectos, "el control no da ningún defecto con la copia")
			case mutante.exactos:
				assert.Equal(t, mutante.defectos, defectos, "los defectos del control con la copia")
			case mutante.prefijos:
				exigeLosPrefijos(t, mutante.defectos, defectos)
			default:
				assert.Subset(t, defectos, mutante.defectos, "los defectos del control con la copia")
			}
		})
	}
}

// exigeLosPrefijos exige que haya tantos defectos como prefijos y que cada uno,
// en su orden, empiece por el suyo.
func exigeLosPrefijos(t *testing.T, prefijos, defectos []string) {
	t.Helper()

	require.Len(t, defectos, len(prefijos), "los defectos del control con la copia:\n%s", strings.Join(defectos, "\n"))

	for posicion, prefijo := range prefijos {
		assert.True(t, strings.HasPrefix(defectos[posicion], prefijo),
			"el defecto %d del control con la copia empieza por «%s»: %s", posicion+1, prefijo, defectos[posicion])
	}
}

// skillNormativa es la skill en la que se aplican todos los controles de
// contenido.
const skillNormativa = "normativa"

// referenciaAMano es el fichero de references/ de la copia en el que escriben
// los mutantes: ninguna skill lo declara en su kitlegal-referencias, y no abre
// con la cabecera de una referencia generada.
const referenciaAMano = "a_mano.md"

// copiaDeLaSkill copia la skill del repositorio, su directorio entero, en
// skills/ de un temporal, y lo devuelve: es la raíz de la copia.
func copiaDeLaSkill(t *testing.T, nombre string) string {
	t.Helper()

	raiz := t.TempDir()
	require.NoError(t, os.CopyFS(filepath.Join(raiz, directorioDeLasSkills, nombre),
		os.DirFS(filepath.Join(raizDeLasSkills, directorioDeLasSkills, nombre))))

	return raiz
}

// escribirLaReferencia escribe el contenido en el fichero de references/ de la
// skill de dir.
func escribirLaReferencia(t *testing.T, dir, fichero, contenido string) {
	t.Helper()

	carpeta := filepath.Join(dir, directorioDeLasReferencias)
	require.NoError(t, os.MkdirAll(carpeta, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(carpeta, fichero), []byte(contenido), 0o600))
}

// conLasReferencias es el mutante que escribe cada fichero de references/ con
// su contenido.
func conLasReferencias(contenidos map[string]string) func(t *testing.T, dir string) {
	return func(t *testing.T, dir string) {
		t.Helper()

		for fichero, contenido := range contenidos {
			escribirLaReferencia(t, dir, fichero, contenido)
		}
	}
}

// quitarDelSkillMd quita del SKILL.md de la skill de dir cada línea con la que
// casa, que tiene que haber al menos una, y devuelve las quitadas, juntas.
func quitarDelSkillMd(t *testing.T, dir string, casa func(linea string) bool) string {
	t.Helper()

	return quitarDelFichero(t, filepath.Join(dir, ficheroDeLaSkill), casa)
}

// quitarDelFichero quita del fichero de la ruta cada línea con la que casa, que
// tiene que haber al menos una, y devuelve las quitadas, juntas.
func quitarDelFichero(t *testing.T, ruta string, casa func(linea string) bool) string {
	t.Helper()

	var quedan, quitadas strings.Builder

	for linea := range strings.Lines(string(contenidoDelFichero(t, ruta))) {
		if casa(linea) {
			quitadas.WriteString(linea)
		} else {
			quedan.WriteString(linea)
		}
	}

	require.NotZero(t, quitadas.Len(), "%s lleva lo que el mutante quita", ruta)
	require.NoError(t, os.WriteFile(ruta, []byte(quedan.String()), 0o600))

	return quitadas.String()
}

// enBloqueDeTexto son las líneas, que terminan en un salto de línea, en un
// bloque de código text, como se enseña una forma fija.
func enBloqueDeTexto(lineas string) string {
	return aperturaDeBloqueDeTexto + "\n" + lineas + cierreDeBloqueDeCodigo + "\n"
}

// sinLasLineas es el mutante que quita del fichero de la skill, su ruta dentro
// de ella, cada línea con la que casa.
func sinLasLineas(fichero string, casa func(linea string) bool) func(t *testing.T, dir string) {
	return func(t *testing.T, dir string) {
		t.Helper()

		quitarDelFichero(t, filepath.Join(dir, fichero), casa)
	}
}

// conLasLineasMovidas es el mutante que lleva del SKILL.md a la referencia a
// mano cada línea con la que casa, en un bloque de código text.
func conLasLineasMovidas(casa func(linea string) bool) func(t *testing.T, dir string) {
	return func(t *testing.T, dir string) {
		t.Helper()

		escribirLaReferencia(t, dir, referenciaAMano, "# Lo que decía SKILL.md\n\n"+
			enBloqueDeTexto(quitarDelSkillMd(t, dir, casa)))
	}
}

// conLasLineasLlevadasAlSkillMd es el mutante que lleva cada línea con la que
// casa del fichero de la skill, su ruta dentro de ella, al final de su SKILL.md,
// en un bloque de código text.
func conLasLineasLlevadasAlSkillMd(fichero string, casa func(linea string) bool) func(t *testing.T, dir string) {
	return func(t *testing.T, dir string) {
		t.Helper()

		llevadas := quitarDelFichero(t, filepath.Join(dir, fichero), casa)
		ruta := filepath.Join(dir, ficheroDeLaSkill)

		require.NoError(t, os.WriteFile(ruta, []byte(string(contenidoDelFichero(t, ruta))+"\n"+
			enBloqueDeTexto(llevadas)), 0o600))
	}
}

// mutantesDeLaProsa son los del control 1: una expresión de la lista, la
// primera de la maquinaria, y una fecha AAAAMMDD con cifras en la prosa de una
// referencia a mano, cada una con el defecto de defectosDeLaProsa y la ruta.
func mutantesDeLaProsa(lista ExpresionesProhibidas) []mutanteDeUnControl {
	control := controlDeLaProsa(lista)
	expresion := lista.Maquinaria[0]

	return []mutanteDeUnControl{
		{
			nombre: "prosa-con-una-expresion-prohibida", control: control, skill: skillNormativa,
			mutar: conLasReferencias(map[string]string{
				referenciaAMano: "# Notas\n\nEsta referencia nombra la " + expresion + " en su prosa.\n",
			}),
			defectos: []string{"normativa: references/a_mano.md: párrafo de la línea 3: " + expresion},
			exactos:  true,
		},
		{
			nombre: "prosa-con-una-fecha-con-cifras", control: control, skill: skillNormativa,
			mutar: conLasReferencias(map[string]string{
				referenciaAMano: "# Notas\n\nLa redacción de 20151002 es la primera.\n",
			}),
			defectos: []string{"normativa: references/a_mano.md: línea 3: fecha escrita con cifras 20151002"},
			exactos:  true,
		},
	}
}

// mutantesDeLasOrdenes son los del control 3: en una referencia a mano, la
// orden de Bash de articulo sin su forma para PowerShell, con ella en otra
// referencia y con ella detrás; y las órdenes de articulos, que desde H31 están
// en references/varios_bloques.md (contracts/skills-reescritas.md de H31,
// «normativa»), llevadas de ahí a SKILL.md, o quitadas.
func mutantesDeLasOrdenes() []mutanteDeUnControl {
	const (
		enBash       = "# Lecturas\n\n```bash\n" + bashDeArticulo + "\n```\n"
		enPowerShell = "# Lecturas en PowerShell\n\n```powershell\n" + powershellDeArticulo + "\n```\n"
	)

	sinSuFormaEnLaReferencia := []string{"normativa: references/a_mano.md: línea 4, «" + bashDeArticulo + "»" +
		sinSuForma}
	conLasDeArticulos := func(linea string) bool {
		return strings.Contains(linea, "kitlegal boe articulos <norma> <bloques>... --json")
	}
	referenciaDeArticulos := path.Join(directorioDeLasReferencias, "varios_bloques.md")

	return []mutanteDeUnControl{
		{
			nombre: "orden-sin-su-forma-para-powershell", control: controlDeLasOrdenes, skill: skillNormativa,
			mutar:    conLasReferencias(map[string]string{referenciaAMano: enBash}),
			defectos: sinSuFormaEnLaReferencia, exactos: true,
		},
		{
			nombre: "orden-con-su-forma-en-otra-referencia", control: controlDeLasOrdenes, skill: skillNormativa,
			mutar:    conLasReferencias(map[string]string{referenciaAMano: enBash, "otra_a_mano.md": enPowerShell}),
			defectos: sinSuFormaEnLaReferencia, exactos: true,
		},
		{
			nombre: "orden-con-su-forma-detras", control: controlDeLasOrdenes, skill: skillNormativa,
			mutar: conLasReferencias(map[string]string{
				referenciaAMano: enBash + "\n```powershell\n" + powershellDeArticulo + "\n```\n",
			}),
		},
		{
			nombre: "ordenes-de-articulos-en-skill-md", control: controlDeLasOrdenes, skill: skillNormativa,
			mutar: conLasLineasLlevadasAlSkillMd(referenciaDeArticulos, conLasDeArticulos),
		},
		{
			nombre: "sin-las-ordenes-de-articulos", control: controlDeLasOrdenes, skill: skillNormativa,
			mutar: sinLasLineas(referenciaDeArticulos, conLasDeArticulos),
			defectos: []string{"normativa: ninguna orden de Bash lee con «kitlegal boe articulos» y comprueba " +
				"detrás con kitlegal graph check"},
			exactos: true,
		},
	}
}

// mutantesDeLaLineaSinConsulta son los del control 2: en cada skill que la
// lleva, la línea sin consulta llevada del bloque text de SKILL.md al de una
// referencia a mano.
func mutantesDeLaLineaSinConsulta() []mutanteDeUnControl {
	mutantes := make([]mutanteDeUnControl, 0, len(skillsConLineaSinConsulta))

	for _, skill := range skillsConLineaSinConsulta {
		mutantes = append(mutantes, mutanteDeUnControl{
			nombre: "linea-sin-consulta-movida-en-" + skill, control: controlDeLaLineaSinConsulta, skill: skill,
			mutar: conLasLineasMovidas(func(linea string) bool {
				return strings.TrimSpace(linea) == lineaSinConsultaDeLasSkills
			}),
			defectos: []string{skill + ": SKILL.md no dice en un bloque text la línea «" + lineaSinConsultaDeLasSkills +
				"»"},
			exactos: true,
		})
	}

	return mutantes
}

// mutantesDeLasFormasMovidas son los de los controles 4, 5 y 6: cada línea de
// SKILL.md que lleva una forma de la redacción fechada, un aviso o un hallazgo,
// con la gramática con la que se juzga la respuesta, llevada al bloque text de
// una referencia a mano, forma a forma; el control da al menos el defecto de la
// movida, porque una línea puede llevar más de una.
func mutantesDeLasFormasMovidas() []mutanteDeUnControl {
	var mutantes []mutanteDeUnControl

	movida := func(nombre string, control controlDeContenido, forma *regexp.Regexp, defecto string) {
		mutantes = append(mutantes, mutanteDeUnControl{
			nombre: nombre, control: control, skill: skillNormativa,
			mutar: conLasLineasMovidas(forma.MatchString), defectos: []string{defecto},
		})
	}

	for _, forma := range formasFechadas {
		movida("forma-fechada-movida-"+strings.ToLower(strings.ReplaceAll(string(forma), " ", "-")),
			controlDeLasFormasFechadas, expresionesFechadas()[forma],
			"normativa: ningún bloque text de SKILL.md lleva la forma fija "+formaEscrita(string(forma)))
	}

	etiquetasDeAviso := boe.EtiquetasDeAviso()
	for _, codigo := range boe.CodigosDeAviso() {
		movida("aviso-movido-"+codigo, controlDeLosAvisos, formasDeAviso()[codigo],
			"normativa: SKILL.md: falta la forma fija del aviso "+codigo+": "+
				formaEscrita(etiquetasDeAviso[codigo]))
	}

	etiquetasDeHallazgo := grafo.EtiquetasDeHallazgo()
	for _, clase := range clasesEtiquetadas() {
		movida("hallazgo-movido-"+clase, controlDeLosHallazgos, formasDeHallazgo()[clase],
			"normativa: SKILL.md: falta la forma fija del hallazgo "+clase+": "+
				formaEscrita(etiquetasDeHallazgo[grafo.ClaseDeHallazgo(clase)]))
	}

	return mutantes
}

// mutantesDeLasLineasDeAviso son los del control 7, uno por skill en que se
// aplica, con una referencia a mano que lleva líneas ⚠ de SKILL.md, con otra
// sangría, y variantes de ellas: en normativa, la de la redacción
// modificada sin la tilde, el ejemplo del contrato; en plazos-y-territorio, la
// del plazo con un marcador de lista y énfasis, una forma con un número de
// lista, otra con los énfasis de subrayado y de código y otra con el marcador *
// delante. Una línea con la marca en medio no es una línea ⚠.
func mutantesDeLasLineasDeAviso(t *testing.T) []mutanteDeUnControl {
	t.Helper()

	modificada := lineaDeAvisoDeLaSkill(t, skillNormativa, "⚠ REDACCIÓN MODIFICADA:")
	malEscrita := strings.Replace(modificada, "REDACCIÓN", "REDACCION", 1)

	plazo := lineaDeAvisoDeLaSkill(t, "plazos-y-territorio", "⚠ PLAZO: <AAAA-MM-DD> — <tipo>")
	festivos := lineaDeAvisoDeLaSkill(t, "plazos-y-territorio", "⚠ FALTAN FESTIVOS LOCALES:")
	conEnfasis := "- **⚠ PLAZO:**" + strings.TrimPrefix(plazo, "⚠ PLAZO:")
	conNumero := "1. " + festivos
	conSubrayado := "_" + festivos + "_"
	conCodigo := "`" + festivos + "`"
	conAsterisco := "* " + festivos

	return []mutanteDeUnControl{
		{
			nombre: "linea-de-aviso-mal-escrita-en-normativa", control: controlDeLasLineasDeAviso,
			skill: skillNormativa,
			mutar: conLasReferencias(map[string]string{
				"redaccion_modificada.md": "# Redacción modificada\n\nLa respuesta lo dice con su línea:\n\n```text\n  " +
					modificada + "\n" + malEscrita + "\n```\n",
			}),
			defectos: []string{"normativa: references/redaccion_modificada.md:7: la línea ⚠ no es ninguna de " +
				"SKILL.md: «" + malEscrita + "»"},
			exactos: true,
		},
		{
			nombre: "lineas-de-aviso-mal-escritas-en-plazos-y-territorio", control: controlDeLasLineasDeAviso,
			skill: "plazos-y-territorio",
			mutar: conLasReferencias(map[string]string{
				referenciaAMano: "# Plazos\n\n- La línea `⚠ PLAZO:` va sola.\n" + conEnfasis + "\n" + conNumero + "\n" +
					conSubrayado + "\n" + conCodigo + "\n   " + festivos + "\t\n" + conAsterisco + "\n",
			}),
			defectos: []string{
				"plazos-y-territorio: references/a_mano.md:4: la línea ⚠ no es ninguna de SKILL.md: «" + conEnfasis + "»",
				"plazos-y-territorio: references/a_mano.md:5: la línea ⚠ no es ninguna de SKILL.md: «" + conNumero + "»",
				"plazos-y-territorio: references/a_mano.md:6: la línea ⚠ no es ninguna de SKILL.md: «" + conSubrayado + "»",
				"plazos-y-territorio: references/a_mano.md:7: la línea ⚠ no es ninguna de SKILL.md: «" + conCodigo + "»",
				"plazos-y-territorio: references/a_mano.md:9: la línea ⚠ no es ninguna de SKILL.md: «" + conAsterisco + "»",
			},
			exactos: true,
		},
	}
}

// lineaDeAvisoDeLaSkill es la línea del SKILL.md de la skill del repositorio que
// empieza, sin su sangría, por el principio, sin los blancos de los extremos;
// tiene que haber una.
func lineaDeAvisoDeLaSkill(t *testing.T, skill, principio string) string {
	t.Helper()

	cargada, err := skills.Cargar(raizDeLasSkills, skill)
	require.NoError(t, err)

	for linea := range strings.Lines(string(cargada.Contenido)) {
		if recortada := strings.TrimSpace(linea); strings.HasPrefix(recortada, principio) {
			return recortada
		}
	}

	require.Failf(t, "falta la línea", "el SKILL.md de %s no lleva ninguna línea que empiece por «%s»", skill, principio)

	return ""
}

// mutantesDeLasOrdenesDeEjemplo son los del control 10, sobre la caché de la
// primera eval de evals/normativa/ que no declara sin_servir, preparada
// aquí, para todos: en una referencia a mano, la orden de buscar de antes de H31,
// sin marcador, que la caché de una sesión no sirve, da su defecto con la skill
// y el fichero, el texto de la Falta; y con un marcador, no se ejecuta y no da
// ninguno.
func mutantesDeLasOrdenesDeEjemplo(t *testing.T) []mutanteDeUnControl {
	t.Helper()

	dirCache := cacheDeLaPrimeraEvalServida(t, evalsDeNormativa)
	control := controlDeLasOrdenesDeEjemplo([]string{skillNormativa}, func(string) string { return dirCache })

	return []mutanteDeUnControl{
		{
			nombre: "orden-de-ejemplo-sin-servir", control: control, skill: skillNormativa,
			mutar: conLasReferencias(map[string]string{
				referenciaAMano: "# Búsquedas\n\n```bash\nkitlegal boe buscar régimen jurídico del sector público " +
					"--json\n```\n",
			}),
			defectos: []string{"normativa: references/a_mano.md: «boe buscar régimen jurídico del sector " +
				"público» terminó con código 4: "},
			prefijos: true,
		},
		{
			nombre: "orden-de-ejemplo-con-un-marcador", control: control, skill: skillNormativa,
			mutar: conLasReferencias(map[string]string{
				referenciaAMano: "# Búsquedas\n\nBusca con `kitlegal boe buscar <términos> --json`.\n",
			}),
		},
	}
}

// conjuntoDeUnaSkill es el conjunto de evals leído de una carpeta de evals/, con
// la carpeta.
type conjuntoDeUnaSkill struct {
	carpeta  string
	conjunto Conjunto
}

// conjuntosDeCadaSkill lee como conjunto de evals cada carpeta de raiz, una por
// skill, en orden de nombre. Lo que en raiz no es una carpeta no son las evals
// de ninguna skill y no se lee.
func conjuntosDeCadaSkill(t *testing.T, raiz string) []conjuntoDeUnaSkill {
	t.Helper()

	entradas, err := os.ReadDir(raiz)
	require.NoError(t, err, "el directorio de evals %s", raiz)

	var conjuntos []conjuntoDeUnaSkill

	for _, entrada := range entradas {
		if !entrada.IsDir() {
			continue
		}

		dir := filepath.Join(raiz, entrada.Name())

		leido, err := LeerConjunto(dir)
		require.NoError(t, err)

		conjuntos = append(conjuntos, conjuntoDeUnaSkill{carpeta: dir, conjunto: leido})
	}

	return conjuntos
}

// malFormadosDeCadaSkill lee como conjunto de evals cada carpeta de raiz, una por
// skill, y devuelve las carpetas, en orden de nombre, y cada fichero mal formado
// de cualquiera de ellas con la carpeta delante de su error. Lo que en raiz no es
// una carpeta no son las evals de ninguna skill y no se lee.
func malFormadosDeCadaSkill(t *testing.T, raiz string) (carpetas, malFormados []string) {
	t.Helper()

	for _, deUnaSkill := range conjuntosDeCadaSkill(t, raiz) {
		carpetas = append(carpetas, deUnaSkill.carpeta)

		for _, malFormado := range deUnaSkill.conjunto.MalFormados {
			malFormados = append(malFormados, deUnaSkill.carpeta+": "+malFormado.Error.Error())
		}
	}

	return carpetas, malFormados
}

// probarGrafosPrevios es la subprueba grafo-previo de TestEvalsDelRepositorio
// (contrato evals-y-skill §3 de H7; FR-085): el grafo previo de cada eval de
// cada carpeta de evals/ que lo lleva nombra un conjunto de GrafosPrevios que
// existe, y prepararlo como lo prepara el job, con UnionDeGrabaciones y en
// temporales, no da ninguna falta ni error y deja en el grafo del mundo un
// BloqueVersion por comando, el de su norma y su bloque; y, desde H7.2, lo que
// hace después la sesión da los hallazgos que la eval espera (contrato
// eval-y-derivada §4; FR-002). Alguna eval lo lleva: sin ninguna, la subprueba
// pasaría en vacío.
func probarGrafosPrevios(t *testing.T) {
	t.Parallel()

	var comprobadas []string

	for _, deUnaSkill := range conjuntosDeCadaSkill(t, directorioDeEvals) {
		for _, eval := range deUnaSkill.conjunto.Evals {
			if eval.GrafoPrevio.Grabaciones == "" {
				continue
			}

			ruta := filepath.Join(deUnaSkill.carpeta, eval.Fichero)
			comprobadas = append(comprobadas, ruta)

			compruebaElGrafoPrevio(t, ruta, eval)
		}
	}

	assert.NotEmpty(t, comprobadas, "alguna eval de %s lleva grafo_previo", directorioDeEvals)
}

// compruebaElGrafoPrevio exige que el grafo previo de la eval de esa ruta nombre
// un conjunto de GrafosPrevios que existe y que prepararlo en una caché temporal,
// con UnionDeGrabaciones, no dé ninguna falta ni error y deje en el grafo del
// mundo un BloqueVersion por comando, el de su norma y su bloque. Desde H7.2
// exige además, sobre ese grafo, lo que verá la sesión (compruebaLaSesion).
func compruebaElGrafoPrevio(t *testing.T, ruta string, eval Eval) {
	t.Helper()

	previo := eval.GrafoPrevio
	require.DirExists(t, filepath.Join(GrafosPrevios, previo.Grabaciones),
		"%s: el grafo previo que nombra está en %s", ruta, GrafosPrevios)

	dirCache := t.TempDir()

	faltas, err := prepararGrafoPrevio(dirCache, UnionDeGrabaciones(), GrafosPrevios, eval)
	require.NoError(t, err, ruta)
	assert.Empty(t, faltas, "%s: comandos del grafo previo que no se preparan:\n%s", ruta, presentarFaltas(faltas))

	porComando := make([]CitaEsperada, 0, len(previo.Comandos))
	for _, comando := range previo.Comandos {
		porComando = append(porComando, CitaEsperada{Norma: comando.Norma, Bloque: comando.Bloque})
	}

	versionados := bloquesVersionados(t, dirCache)

	citas := make([]CitaEsperada, 0, len(versionados))
	anotados := make(map[string]bloqueVersionado, len(versionados))

	for _, versionado := range versionados {
		citas = append(citas, versionado.cita)
		anotados[versionado.id] = versionado
	}

	assert.ElementsMatch(t, porComando, citas,
		"%s: el grafo previo deja un BloqueVersion por comando, el de su norma y su bloque", ruta)

	compruebaLaSesion(t, ruta, eval, dirCache, anotados)
}

// compruebaLaSesion hace sobre el grafo del mundo de dirCache, con el grafo
// previo de la eval ya preparado, lo que hace su sesión, en proceso y sin red
// (contrato eval-y-derivada §4 de H7.2; research D14; FR-002): lee cada bloque de
// los comandos de la eval con boe articulo --json sobre UnionDeGrabaciones, con
// una caché temporal y entregando a ese grafo, y comprueba con graph check --json
// cada norma de sus comandos, con los bloques que ha leído de ella. Las dos
// terminan en 0; las clases de los hallazgos son exactamente las de hallazgos de
// la eval; y cada version-obsoleta es de un BloqueVersion que dejó el grafo
// previo —de los anotados, por su id—, con la fecha de vigencia de esa versión
// como la superada y la del bloque leído como la reciente. Desde H7.4, cada
// redacción de redacciones_modificadas de la eval es uno de esos version-obsoleta,
// de su norma y su bloque, con su fecha_vigencia como la superada y su
// fecha_vigencia_reciente como la leída: la eval no puede esperar unas fechas que
// las grabaciones no dan (contracts/evals-y-juicio.md §4 de H7.4; research D11).
func compruebaLaSesion(t *testing.T, ruta string, eval Eval, dirCache string, anotados map[string]bloqueVersionado) {
	t.Helper()

	sesion, err := registroDeBoe(copiaDeLaUnionDeGrabaciones(t), cache.ConDirectorio(t.TempDir()))
	require.NoError(t, err)
	require.NoError(t, sesion.Registrar(app.AppletGrafo(app.DependenciasDeGrafo{
		Reloj: time.Now, Almacen: []graph.Opcion{graph.ConDirectorio(dirCache)},
	})))
	sesion.EntregarAlGrafo(graph.Nuevo(graph.ConDirectorio(dirCache)))

	var normas []string

	bloquesDe := map[string][]string{}
	leidas := map[CitaEsperada]string{}

	for _, comando := range eval.Comandos {
		if comando.Norma != "" && !slices.Contains(normas, comando.Norma) {
			normas = append(normas, comando.Norma)
		}

		if formaDelComando(comando) != formaBloque {
			continue
		}

		consulta := Consulta{Applet: comando.Applet, Verbo: verboArticulo, Argumentos: []string{comando.Norma, comando.Bloque}}
		leido, errores := leerBloque(t, sesion, consulta)
		assert.Empty(t, errores, "%s: «%s» llega al grafo de la sesión", ruta, ordenDe(consulta))

		bloquesDe[comando.Norma] = append(bloquesDe[comando.Norma], comando.Bloque)
		leidas[CitaEsperada{Norma: comando.Norma, Bloque: comando.Bloque}] = leido.FechaVigencia
	}

	var (
		clases    []string
		obsoletas []RedaccionEsperada
	)

	for _, norma := range normas {
		for _, hallazgo := range comprobacionDeLaSesion(t, sesion, norma, bloquesDe[norma]).Hallazgos {
			clases = append(clases, string(hallazgo.Clase))

			if hallazgo.Clase != grafo.ClaseVersionObsoleta {
				continue
			}

			anotado, esDelGrafoPrevio := anotados[hallazgo.ID]
			if !assert.True(t, esDelGrafoPrevio, "%s: el version-obsoleta de %s es de una redacción que dejó el grafo "+
				"previo: %s", ruta, norma, hallazgo.ID) {
				continue
			}

			assert.Equal(t, anotado.fechaVigencia, hallazgo.FechaVigencia,
				"%s: la redacción superada de %s es la que dejó el grafo previo", ruta, hallazgo.ID)
			assert.Equal(t, leidas[anotado.cita], hallazgo.FechaVigenciaReciente,
				"%s: la redacción reciente de %s es la que ha leído la sesión", ruta, hallazgo.ID)

			obsoletas = append(obsoletas, RedaccionEsperada{
				Norma:                 anotado.cita.Norma,
				Bloque:                anotado.cita.Bloque,
				FechaVigencia:         hallazgo.FechaVigencia,
				FechaVigenciaReciente: hallazgo.FechaVigenciaReciente,
			})
		}
	}

	slices.Sort(clases)
	assert.Equal(t, slices.Sorted(slices.Values(eval.Hallazgos)), slices.Compact(clases),
		"%s: graph check da en la sesión exactamente las clases de hallazgo que la eval espera", ruta)

	for _, esperada := range eval.RedaccionesModificadas {
		assert.Contains(t, obsoletas, esperada, "%s: la redacción modificada %s que la eval espera es un "+
			"version-obsoleta que graph check da en la sesión, de su norma y su bloque y con sus dos fechas",
			ruta, esperada.texto())
	}
}

// comprobacionDeLaSesion es la data de graph check <norma> <bloques> --json con
// el registro de la sesión. La comprobación termina en 0.
func comprobacionDeLaSesion(t *testing.T, sesion *app.Registro, norma string, bloques []string) grafo.Comprobacion {
	t.Helper()

	var salida, errores bytes.Buffer

	argv := slices.Concat([]string{programaDeLasConsultas, "graph", verboCheck, norma}, bloques, []string{"--json"})
	codigo := app.Main(argv, sesion, &salida, &errores, sinDatosDeConstruccion, sinDatosDeConstruccion,
		sinDatosDeConstruccion)
	require.Zero(t, codigo, "«%s» comprueba la memoria: %s", strings.Join(argv[1:], " "), errores.String())

	var sobre struct {
		Data grafo.Comprobacion `json:"data"`
	}
	require.NoError(t, json.Unmarshal(salida.Bytes(), &sobre), "la salida de «%s» es un sobre con una comprobación",
		strings.Join(argv[1:], " "))

	return sobre.Data
}

// bloqueVersionado es un BloqueVersion del grafo del mundo: su id, la norma y el
// bloque de los que cuelga y su fecha de vigencia.
type bloqueVersionado struct {
	id            string
	cita          CitaEsperada
	fechaVigencia string
}

// bloquesVersionados son los BloqueVersion del grafo del mundo de dirCache, en
// el orden de su instantánea, cada uno con el identificador de la Norma y el
// bloque del Bloque de los que cuelga, por las aristas eli:has_part y
// eli:has_version. Una versión o un Bloque a los que no llega una sola de esas
// aristas hace fallar la prueba.
func bloquesVersionados(t *testing.T, dirCache string) []bloqueVersionado {
	t.Helper()

	lectura, err := graph.Leer(t.Context(), graph.ConDirectorio(dirCache))
	require.NoError(t, err)

	instantanea, err := lectura.Instantanea(t.Context(), grafo.Ambito{})
	require.NoError(t, errors.Join(err, lectura.Close()))

	nodos := make(map[string]grafo.NodoDeInstantanea, len(instantanea.Nodos))
	for _, nodo := range instantanea.Nodos {
		nodos[nodo.ID] = nodo
	}

	// deQuienCuelga es el nodo del que sale la única arista de la relación que
	// llega al de ese id.
	deQuienCuelga := func(relacion, id string) grafo.NodoDeInstantanea {
		var origenes []string

		for _, arista := range instantanea.Aristas {
			if arista.Relacion == relacion && arista.Destino == id {
				origenes = append(origenes, arista.Origen)
			}
		}

		require.Len(t, origenes, 1, "una sola arista %s llega a %s", relacion, id)

		return nodos[origenes[0]]
	}

	var versionados []bloqueVersionado

	for _, nodo := range instantanea.Nodos {
		if nodo.Tipo != grafo.TipoBloqueVersion {
			continue
		}

		bloque := deQuienCuelga(grafo.RelacionTieneVersion, nodo.ID)
		norma := deQuienCuelga(grafo.RelacionTieneParte, bloque.ID)

		versionados = append(versionados, bloqueVersionado{
			id: nodo.ID,
			cita: CitaEsperada{
				Norma:  fmt.Sprint(norma.Datos[grafo.DatoIdentificador]),
				Bloque: fmt.Sprint(bloque.Datos[grafo.DatoBloque]),
			},
			fechaVigencia: fmt.Sprint(nodo.Datos[grafo.DatoFechaVigencia]),
		})
	}

	return versionados
}

// TestFormatoDeLasEvalsDeCadaSkill fija, sobre un evals/ que el propio test
// escribe en t.TempDir(), que el subtest formato de TestEvalsDelRepositorio lee
// las evals de cada skill y no solo las de normativa (FR-060): con una eval
// bien formada en alfa/, otra sin pregunta en beta/ y un fichero suelto en la
// raíz, lee alfa y beta y da solo el fichero de beta, con su directorio y su
// error.
func TestFormatoDeLasEvalsDeCadaSkill(t *testing.T) {
	t.Parallel()

	raiz := t.TempDir()
	alfa, beta := filepath.Join(raiz, "alfa"), filepath.Join(raiz, "beta")

	require.NoError(t, os.Mkdir(alfa, 0o700))
	require.NoError(t, os.Mkdir(beta, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(alfa, nombreDeEval), []byte(contenidoDelArticulo21), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(beta, "01-sin-pregunta.yaml"), []byte(contenidoSinPregunta), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(raiz, "notas.txt"), []byte(contenidoSinPregunta), 0o600))

	carpetas, malFormados := malFormadosDeCadaSkill(t, raiz)

	assert.Equal(t, []string{alfa, beta}, carpetas)
	require.Len(t, malFormados, 1)
	assert.True(t, strings.HasPrefix(malFormados[0], beta+": 01-sin-pregunta.yaml: "), malFormados[0])
	assert.Contains(t, malFormados[0], "missing property 'pregunta'")
}

// El fragmento que la persona pega en las evals de jurisprudencia y lo que lo
// fija (spec de H23, FR-053 y SC-008): su ruta, relativa al directorio de este
// paquete, su tamaño y su SHA-256, que son los de su manifiesto. Y el tamaño de
// su ficha, que son sus líneas hasta la primera en blanco
// (contracts/evals-jurisprudencia.md §2 de H25).
const (
	fragmentoDelRepositorio = "../../evidencias/adr-0036/ecli-es-ts-2023-3144-fragmento.txt"
	bytesDelFragmento       = 2353
	huellaDelFragmento      = "4886e0c8ca9836527ec08d8732b50315640a76803033374af5287b2a8321ee27"
	bytesDeLaFicha          = 316
)

// evalsConElFragmento son las evals de jurisprudencia cuya pregunta lleva
// pegado el fragmento entero: la que pide que se cite y la que dice traer otra
// sentencia (contracts/evals-jurisprudencia.md §5 de H23) y, desde H25, la que
// pregunta por su doctrina (contracts/evals-jurisprudencia.md §1 de H25;
// FR-062, FR-066).
var evalsConElFragmento = []string{jurisprudenciaDocumento, jurisprudenciaNoEsElPedido, jurisprudenciaDoctrina}

// evalConLaFicha es la eval de jurisprudencia cuya pregunta lleva pegada la
// ficha del fragmento y nada más de él (contracts/evals-jurisprudencia.md §1 de
// H25; FR-063, FR-066).
const evalConLaFicha = jurisprudenciaFicha

// preguntasSinElFragmento son, de las evals de evalsConElFragmento leídas de
// dir, las que no llevan el fragmento byte a byte en su pregunta, tal como la
// lee LeerEval.
func preguntasSinElFragmento(t *testing.T, dir string, fragmento []byte) []string {
	t.Helper()

	var sinEl []string

	for _, fichero := range evalsConElFragmento {
		eval, err := LeerEval(fichero, contenidoDelFichero(t, filepath.Join(dir, fichero)))
		require.NoError(t, err)

		if !strings.Contains(eval.Pregunta, string(fragmento)) {
			sinEl = append(sinEl, fichero)
		}
	}

	return sinEl
}

// llevaSoloLaFicha dice si la eval evalConLaFicha, leída de dir, lleva pegada en
// su pregunta, tal como la lee LeerEval, la ficha byte a byte y nada más: lo que
// sigue a la primera línea en blanco de la pregunta es la ficha.
func llevaSoloLaFicha(t *testing.T, dir, ficha string) bool {
	t.Helper()

	eval, err := LeerEval(evalConLaFicha, contenidoDelFichero(t, filepath.Join(dir, evalConLaFicha)))
	require.NoError(t, err)

	_, pegado, _ := strings.Cut(eval.Pregunta, "\n\n")

	return pegado == ficha
}

// TestPreguntasConElFragmento fija que el texto que la persona pega en las
// evals 04 y 06 de jurisprudencia es el fragmento del repositorio, byte a byte
// (spec de H23, FR-053 y SC-008): el fichero es el de su manifiesto, por su
// tamaño y su huella; la pregunta de cada una, tal como la lee LeerEval, lo
// contiene entero; y con un solo byte del texto pegado cambiado en una copia de
// las evals, la de ese byte deja de contenerlo. Desde H25
// (contracts/evals-jurisprudencia.md §4 de H25; FR-066, FR-110; SC-010), también
// la eval (h), la 08, lleva el fragmento byte a byte, y la (i), la 09, su ficha
// byte a byte y nada más del fragmento: con un byte de la ficha cambiado en una
// copia, o con la línea que la sigue en el fragmento puesta detrás de ella, deja
// de llevar solo la ficha.
func TestPreguntasConElFragmento(t *testing.T) {
	t.Parallel()

	fragmento := contenidoDelFichero(t, fragmentoDelRepositorio)
	require.Len(t, fragmento, bytesDelFragmento, "%s no ha cambiado de tama\xc3\xb1o", fragmentoDelRepositorio)

	huella := sha256.Sum256(fragmento)
	require.Equal(t, huellaDelFragmento, hex.EncodeToString(huella[:]), "%s no ha cambiado", fragmentoDelRepositorio)

	assert.Empty(t, preguntasSinElFragmento(t, evalsDeJurisprudencia, fragmento),
		"evals de %s cuya pregunta no lleva %s byte a byte", evalsDeJurisprudencia, fragmentoDelRepositorio)

	ficha := leerTextosDelFragmento(t).ficha
	require.Len(t, ficha, bytesDeLaFicha, "la ficha de %s no ha cambiado de tama\xc3\xb1o", fragmentoDelRepositorio)

	assert.True(t, llevaSoloLaFicha(t, evalsDeJurisprudencia, ficha),
		"%s lleva pegada la ficha de %s byte a byte, y nada m\xc3\xa1s del fragmento", evalConLaFicha, fragmentoDelRepositorio)

	// La última línea del fragmento, que en la eval va con la sangría del bloque:
	// su último byte es el que se cambia en la copia.
	lineas := strings.Split(strings.TrimRight(string(fragmento), "\n"), "\n")
	ultima := []byte("  " + lineas[len(lineas)-1] + "\n")

	for _, cambiada := range evalsConElFragmento {
		t.Run("con-un-byte-cambiado-en-la-"+cambiada[:2], func(t *testing.T) {
			t.Parallel()

			copia := t.TempDir()

			for _, fichero := range evalsConElFragmento {
				contenido := contenidoDelFichero(t, filepath.Join(evalsDeJurisprudencia, fichero))

				if fichero == cambiada {
					posicion := bytes.LastIndex(contenido, ultima)
					require.GreaterOrEqual(t, posicion, 0, "%s lleva la \xc3\xbaltima l\xc3\xadnea del fragmento", fichero)

					// El byte anterior al salto de línea, que deja de ser el suyo.
					contenido[posicion+len(ultima)-2]++
				}

				require.NoError(t, os.WriteFile(filepath.Join(copia, fichero), contenido, 0o600))
			}

			assert.Equal(t, []string{cambiada}, preguntasSinElFragmento(t, copia, fragmento))
		})
	}

	// La última línea de la ficha, con la sangría del bloque, y la línea que la
	// sigue en el fragmento tras su línea en blanco: lo primero que la eval (i)
	// no lleva.
	deLaFicha := strings.Split(strings.TrimRight(ficha, "\n"), "\n")
	ultimaDeLaFicha := []byte("  " + deLaFicha[len(deLaFicha)-1] + "\n")

	siguiente, _, _ := strings.Cut(strings.TrimPrefix(string(fragmento), ficha+"\n"), "\n")
	require.NotEmpty(t, siguiente, "premisa: %s sigue detr\xc3\xa1s de su ficha", fragmentoDelRepositorio)

	copiasDeLaFicha := []struct {
		nombre string

		// cambiar devuelve el contenido de la eval con el cambio, que va donde
		// está la última línea de la ficha.
		cambiar func(contenido []byte, posicion int) []byte
	}{
		{
			nombre: "con-un-byte-cambiado-en-la-ficha",
			cambiar: func(contenido []byte, posicion int) []byte {
				// El byte anterior al salto de línea, que deja de ser el suyo.
				contenido[posicion+len(ultimaDeLaFicha)-2]++

				return contenido
			},
		},
		{
			nombre: "con-mas-del-fragmento-que-la-ficha",
			cambiar: func(contenido []byte, posicion int) []byte {
				return slices.Insert(contenido, posicion+len(ultimaDeLaFicha), []byte("\n  "+siguiente+"\n")...)
			},
		},
	}

	for _, caso := range copiasDeLaFicha {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			contenido := contenidoDelFichero(t, filepath.Join(evalsDeJurisprudencia, evalConLaFicha))

			posicion := bytes.LastIndex(contenido, ultimaDeLaFicha)
			require.GreaterOrEqual(t, posicion, 0, "%s lleva la \xc3\xbaltima l\xc3\xadnea de la ficha", evalConLaFicha)

			copia := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(copia, evalConLaFicha), caso.cambiar(contenido, posicion), 0o600))

			assert.False(t, llevaSoloLaFicha(t, copia, ficha))
		})
	}
}

// preguntasDeLosSondeos es, desde la raíz del repositorio, el fichero con las
// preguntas de los dos sondeos de la validación del juez de jurisprudencia
// (contracts/evals-jurisprudencia.md §1 de H25).
const preguntasDeLosSondeos = "evidencias/adr-0037-jurisprudencia/preguntas.json"

// evalsDelSondeo une cada una de las cuatro evals de H25 con la pregunta del
// sondeo que lleva y con los bytes de esa pregunta
// (contracts/evals-jurisprudencia.md §1 de H25; research D14 de H25): las evals
// se numeran seguidas, y las preguntas conservan el número de su sondeo.
var evalsDelSondeo = []struct {
	fichero  string
	pregunta string
	bytes    int
}{
	{fichero: jurisprudenciaConocida, pregunta: "07-resumen-de-una-conocida", bytes: 39},
	{fichero: jurisprudenciaDoctrina, pregunta: "09-doctrina-con-el-fallo-delante", bytes: 2496},
	{fichero: jurisprudenciaFicha, pregunta: "11-de-que-trata-con-la-ficha-sola", bytes: 423},
	{fichero: jurisprudenciaDadaPorHecha, pregunta: "12-doctrina-dada-por-hecha", bytes: 123},
}

// preguntasQueNoSonLasDelSondeo son, de las evals de evalsDelSondeo leídas de
// dir, las que no llevan como pregunta, tal como la lee LeerEval, la de su
// entrada del sondeo, byte a byte: delSondeo las da por el fichero de su eval.
func preguntasQueNoSonLasDelSondeo(t *testing.T, dir string, delSondeo map[string]string) []string {
	t.Helper()

	var otras []string

	for _, caso := range evalsDelSondeo {
		eval, err := LeerEval(caso.fichero, contenidoDelFichero(t, filepath.Join(dir, caso.fichero)))
		require.NoError(t, err)

		if eval.Pregunta != delSondeo[caso.fichero] {
			otras = append(otras, caso.fichero)
		}
	}

	return otras
}

// cambiarUnByteDeLaPregunta cambia, en el contenido de una eval, el último byte
// de la última línea de su pregunta, que el fichero lleva tal cual detrás de su
// clave pregunta.
func cambiarUnByteDeLaPregunta(t *testing.T, fichero string, contenido []byte) {
	t.Helper()

	eval, err := LeerEval(fichero, contenido)
	require.NoError(t, err)

	lineas := strings.Split(strings.TrimRight(eval.Pregunta, "\n"), "\n")
	ultima := []byte(lineas[len(lineas)-1])

	clave := bytes.Index(contenido, []byte("\npregunta:"))
	require.GreaterOrEqual(t, clave, 0, "%s lleva la clave pregunta detr\xc3\xa1s de su comentario", fichero)

	posicion := bytes.LastIndex(contenido, ultima)
	require.Greater(t, posicion, clave,
		"%s lleva tal cual, detr\xc3\xa1s de su clave, la \xc3\xbaltima l\xc3\xadnea de su pregunta", fichero)

	contenido[posicion+len(ultima)-1]++
}

// TestPreguntasDelSondeo fija que las cuatro evals de H25 preguntan lo que se
// preguntó en el sondeo con el que se validó el juez de jurisprudencia
// (contracts/evals-jurisprudencia.md §1 y §4 de H25; research D14 de H25;
// FR-060, FR-110; SC-010): la pregunta de cada una, tal como la lee LeerEval, es
// igual, byte a byte, a la de su entrada de preguntas.json, compuesta con el
// fragmento y su ficha como la compone la reconstrucción de un caso de la
// medida, y tiene los bytes del contrato; y con un solo byte de la pregunta
// cambiado en una copia de las cuatro evals, la de ese byte deja de serlo.
func TestPreguntasDelSondeo(t *testing.T) {
	t.Parallel()

	preguntas, err := leerPreguntasDelSondeo(raizDelRepositorio, preguntasDeLosSondeos)
	require.NoError(t, err)

	quienCompone := nuevoReconstructor(evalsDeJurisprudencia)
	delSondeo := make(map[string]string, len(evalsDelSondeo))

	for _, caso := range evalsDelSondeo {
		compuesta, err := quienCompone.preguntaDelSondeo(preguntas, caso.pregunta)
		require.NoError(t, err, "la pregunta %s de %s", caso.pregunta, preguntasDeLosSondeos)
		require.Len(t, compuesta, caso.bytes, "la pregunta %s de %s, compuesta", caso.pregunta, preguntasDeLosSondeos)

		delSondeo[caso.fichero] = compuesta
	}

	assert.Empty(t, preguntasQueNoSonLasDelSondeo(t, evalsDeJurisprudencia, delSondeo),
		"evals de %s cuya pregunta no es, byte a byte, la de su entrada de %s", evalsDeJurisprudencia, preguntasDeLosSondeos)

	for _, cambiada := range evalsDelSondeo {
		t.Run("con-un-byte-cambiado-en-la-"+cambiada.fichero[:2], func(t *testing.T) {
			t.Parallel()

			copia := t.TempDir()

			for _, caso := range evalsDelSondeo {
				contenido := contenidoDelFichero(t, filepath.Join(evalsDeJurisprudencia, caso.fichero))

				if caso.fichero == cambiada.fichero {
					cambiarUnByteDeLaPregunta(t, caso.fichero, contenido)
				}

				require.NoError(t, os.WriteFile(filepath.Join(copia, caso.fichero), contenido, 0o600))
			}

			assert.Equal(t, []string{cambiada.fichero}, preguntasQueNoSonLasDelSondeo(t, copia, delSondeo))
		})
	}
}

// normasConocidasDe es, por identificador, lo que las reglas del conjunto
// necesitan de cada norma de la tabla.
func normasConocidasDe(normas []skills.Norma) map[string]NormaConocida {
	conocidas := make(map[string]NormaConocida, len(normas))
	for _, norma := range normas {
		conocidas[norma.Identificador] = NormaConocida{Abreviatura: norma.Abreviatura, Materias: norma.Materias}
	}

	return conocidas
}

// presentarDefectos escribe cada defecto del conjunto en su línea, con su regla
// delante.
func presentarDefectos(defectos []DefectoDelConjunto) string {
	lineas := make([]string, 0, len(defectos))
	for _, defecto := range defectos {
		lineas = append(lineas, defecto.Regla+": "+defecto.Mensaje)
	}

	return strings.Join(lineas, "\n")
}

// presentarFaltas escribe cada falta en sus líneas, una por origen.
func presentarFaltas(faltas []Falta) string {
	textos := make([]string, 0, len(faltas))
	for _, falta := range faltas {
		textos = append(textos, falta.String())
	}

	return strings.Join(textos, "\n")
}

// exigeLecturasFechadas exige que las consultas de las evals de la carpeta lean
// algún bloque y que cada lectura de un bloque sin fecha vaya acompañada de la
// misma con --fecha 9999-12-31: así, preparar y comprobar sin red su caché es
// también prepararla y comprobarla para la lectura fechada de cada bloque
// (contracts/evals-redaccion-fechada.md §4 de H20; research D18; FR-053).
func exigeLecturasFechadas(t *testing.T, consultas []Consulta, carpeta string) {
	t.Helper()

	esLecturaDe := func(consulta Consulta, argumentos []string) bool {
		return consulta.Applet == "boe" && consulta.Verbo == "articulo" && slices.Equal(consulta.Argumentos, argumentos)
	}

	var bloques, sinLaFechada []string

	for _, consulta := range consultas {
		if consulta.Applet != "boe" || consulta.Verbo != "articulo" || slices.Contains(consulta.Argumentos, "--fecha") {
			continue
		}

		bloque := strings.Join(consulta.Argumentos, " ")
		bloques = append(bloques, bloque)

		fechada := slices.Concat(consulta.Argumentos, []string{"--fecha", "9999-12-31"})
		if !slices.ContainsFunc(consultas, func(otra Consulta) bool { return esLecturaDe(otra, fechada) }) {
			sinLaFechada = append(sinLaFechada, bloque)
		}
	}

	require.NotEmpty(t, bloques, "las evals de %s leen algún bloque sin fecha", carpeta)
	assert.Empty(t, sinLaFechada, "bloques de las evals de %s sin su lectura con --fecha 9999-12-31", carpeta)
}
