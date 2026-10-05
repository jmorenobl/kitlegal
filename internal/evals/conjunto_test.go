package evals

import (
	"bytes"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io/fs"
	"os"
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

// juezEn es el juez que TestLeerConjunto espera de un conjunto creado en dir: el
// del caso, con las rutas de sus casos y de su medida dentro de dir; o nil, si
// el caso no espera ninguno.
func juezEn(dir string, juez *Juez) *Juez {
	if juez == nil {
		return nil
	}

	enDir := *juez
	enDir.Casos = filepath.Join(dir, juez.Casos)
	enDir.Medida = filepath.Join(dir, juez.Medida)

	return &enDir
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
// FR-053): las de legal-core, con la skill que no se activa en NoSeActivan, y la
// positiva de los dos bloques de la LCSP, con sus redacciones modificadas en
// RedaccionesModificadas.
//
// Desde H24 (contracts/juez-y-voto.md §1; FR-020), fija también la carpeta del
// juez, con los casos de casosDeLaCarpetaJuez: la entrada que se llama
// exactamente juez no es un fichero de eval. Si es una carpeta bien formada,
// queda en Conjunto.Juez, con sus clases en su orden, la rúbrica y el esquema
// enteros y las rutas de los casos y de la medida. Cada fichero suyo que falta,
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
			// Las dos de legal-core, cada una con la skill que no se activa
			// detrás de activa, como en el repositorio.
			nombre: "no-se-activan-en-legal-core",
			entradas: []entradaDeConjunto{
				{
					nombre: legalCoreCubierto,
					contenido: preguntaDelMunicipio + "activa: true\n" + noSeActivaBoeLegislacion + comandoDelMunicipio +
						"territorio:\n  comunidad: Comunidad de Madrid\n",
				},
				{nombre: legalCoreNoActivacion, contenido: evalDeLaReceta + noSeActivaBoeLegislacion},
			},
			evals: []Eval{
				{
					Fichero:     legalCoreCubierto,
					Pregunta:    "¿En qué boletines se publican las normas que afectan a Leganés?",
					Activa:      true,
					NoSeActivan: []string{"boe-legislacion"},
					Comandos:    []ComandoEsperado{{Applet: "territorio", Verbo: "resolver", Municipio: "Leganés"}},
					Territorio:  TerritorioEsperado{Comunidad: "Comunidad de Madrid"},
				},
				{
					Fichero:     legalCoreNoActivacion,
					Pregunta:    "¿Cómo se hace una tortilla de patatas jugosa?",
					NoSeActivan: []string{"boe-legislacion"},
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
			assert.Equal(t, juezEn(dir, caso.juez), conjunto.Juez)

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

	// juez es el juez que tiene que quedar en el conjunto, con las rutas de sus
	// casos y de su medida relativas al directorio de evals; nil, ninguno.
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
				Casos:   casosEnElJuez,
				Medida:  medidaEnElJuez,
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
	normaDesconocida    = "BOE-A-0000-99"
)

// conjuntoSintetico son las evals y las normas conocidas con las que
// TestConjuntoDeEvals llama a ComprobarConjunto con las reglas de
// boe-legislacion.
type conjuntoSintetico struct {
	evals  []Eval
	normas map[string]NormaConocida
}

// conjuntoQueCumple devuelve, nuevo en cada llamada, un conjunto que cumple
// todas las reglas de data-model §6.3: el de evals/boe-legislacion/ del contrato
// de evals §2, con diez positivas de una norma cada una y dos de no activación,
// la informativa por materia y, desde H21, la eval sin binario ni servidor
// (contracts/evals-en-dos-modos.md §1 de H21). La 07 y la 08 citan además una
// norma común, porque una positiva que cita dos normas cuenta como materia
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
		evals: make([]Eval, 0, len(positivas)+4),
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
// (contrato de evals §3 de H6; research D22): las de boe-legislacion y las de
// legal-core, cada juego sobre evals sintéticas. Desde H21, el tamaño de
// boe-legislacion llega a 21 y los dos juegos llevan la regla sin binario ni
// servidor, la última de cada tabla: exactamente una eval que lo declara, que no
// cuenta como positiva ni necesita un esperado verificable
// (contracts/evals-en-dos-modos.md §1 de H21; FR-046).
func TestConjuntoDeEvals(t *testing.T) {
	t.Parallel()

	t.Run("boe-legislacion", probarReglasDeBoeLegislacion)
	t.Run("legal-core", probarReglasDeLegalCore)
}

// probarReglasDeBoeLegislacion fija ComprobarConjunto con ReglasDeBoeLegislacion
// (contrato de evals §2 de H5): el conjunto que cumple todas las reglas de
// data-model §6.3 no da ningún defecto; cada copia que incumple solo una regla da
// exactamente un defecto, con el nombre de esa regla y un mensaje que nombra los
// ficheros y las normas implicados; y un conjunto que las incumple todas da un
// defecto por regla, en el orden de la tabla. Desde H21, el conjunto que cumple
// lleva la eval sin binario ni servidor sin que cuente como positiva; con 21
// evals sigue cumpliendo y con 22 incumple el tamaño; y sin ninguna eval sin
// binario ni servidor, o con dos, incumple solo la regla nueva
// (contracts/evals-en-dos-modos.md §1 de H21).
func probarReglasDeBoeLegislacion(t *testing.T) {
	t.Parallel()

	todos := []string{
		"01-lpac-articulo-21.yaml", "02-lcsp-contrato-menor.yaml", "03-lrbrl-atribuciones-del-pleno.yaml",
		"04-lgt-prescripcion.yaml", "05-trlrhl-impuestos-municipales.yaml", "06-irpf-rendimientos-del-trabajo.yaml",
		"07-lrjsp-principio-de-legalidad.yaml", "08-ltaibg-plazo-de-resolucion.yaml",
		"09-constitucion-articulo-140.yaml", "10-et-vacaciones.yaml",
		"11-no-activa-programacion.yaml", "12-no-activa-acuerdo-entre-amigos.yaml",
		"13-lrbrl-atribuciones-por-materia.yaml", "14-sin-binario-ni-servidor.yaml",
	}
	positivas := todos[:10]
	informativa := todos[12]
	sinBinario := todos[13]
	otraSinBinario := "15-sin-binario-ni-servidor-otra.yaml"

	// Con las siete primeras, el conjunto tiene las 21 evals del máximo; con las
	// ocho, una más.
	deMas := []string{
		"15-no-activa-15.yaml", "16-no-activa-16.yaml", "17-no-activa-17.yaml", "18-no-activa-18.yaml",
		"19-no-activa-19.yaml", "20-no-activa-20.yaml", "21-no-activa-21.yaml", "22-no-activa-22.yaml",
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

		assert.Empty(t, ComprobarConjunto(conjunto.evals, conjunto.normas, ReglasDeBoeLegislacion()))

		conDeMas(&conjunto, deMas[:len(deMas)-1])
		require.Len(t, conjunto.evals, 21, "el conjunto sint\xc3\xa9tico llega al m\xc3\xa1ximo de evals")

		assert.Empty(t, ComprobarConjunto(conjunto.evals, conjunto.normas, ReglasDeBoeLegislacion()),
			"con 21 evals, el conjunto cumple el tama\xc3\xb1o")
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
			dice:     []string{"hay 22 evals y el conjunto lleva entre 10 y 21"},
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

			defectos := ComprobarConjunto(conjunto.evals, conjunto.normas, ReglasDeBoeLegislacion())
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

		for numero := 1; numero <= 22; numero++ {
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
		}, reglasIncumplidas(ComprobarConjunto(conjunto.evals, conjunto.normas, ReglasDeBoeLegislacion())))
	})
}

// Las evals sintéticas del conjunto de legal-core de probarReglasDeLegalCore: la
// del municipio cubierto, la del no cubierto y la de no activación (contrato de
// evals §4 de H6) y, desde H21, la eval sin binario ni servidor
// (contracts/evals-en-dos-modos.md §1 de H21); y las que los casos le añaden.
const (
	legalCoreCubierto       = "01-territorio-municipio-cubierto.yaml"
	legalCoreNoCubierto     = "02-territorio-municipio-no-cubierto.yaml"
	legalCoreNoActivacion   = "03-no-activa-receta-de-cocina.yaml"
	legalCoreSinBinario     = "04-sin-binario-ni-servidor.yaml"
	legalCoreSinVerificable = "05-territorio-sin-esperado.yaml"
	legalCoreConCitas       = "05-articulo-21-con-citas.yaml"
	legalCoreOtraSinBinario = "05-sin-binario-ni-servidor-otra.yaml"
)

// conjuntoDeLegalCore devuelve, nuevo en cada llamada, un conjunto que cumple las
// reglas de legal-core: una eval activa que resuelve un municipio del
// territorio configurado y espera su boletín, otra que resuelve uno de una
// comunidad sin configuración y espera no configurado cada aspecto de boletín,
// una de no activación y la eval sin binario ni servidor, activa y sin nada que
// consultar ni que esperar.
func conjuntoDeLegalCore() []Eval {
	return []Eval{
		{
			Fichero:  legalCoreCubierto,
			Pregunta: "¿En qué boletines se publican las normas que afectan a Leganés?",
			Activa:   true,
			Comandos: []ComandoEsperado{{Applet: "territorio", Verbo: "resolver", Municipio: "Leganés"}},
			Territorio: TerritorioEsperado{
				Comunidad: "Comunidad de Madrid", Provincia: "Madrid", Boletines: []string{"BOCM"},
			},
		},
		{
			Fichero:  legalCoreNoCubierto,
			Pregunta: "¿En qué boletines se publican las normas que afectan a Tordesillas?",
			Activa:   true,
			Comandos: []ComandoEsperado{{Applet: "territorio", Verbo: "resolver", Municipio: "Tordesillas"}},
			Territorio: TerritorioEsperado{
				Comunidad: "Castilla y León",
				Provincia: "Valladolid",
				Cobertura: []string{"boletin_autonomico: no-configurado", "boletin_provincial: no-configurado"},
			},
		},
		{Fichero: legalCoreNoActivacion, Pregunta: "¿Cómo hago una tortilla de patatas?"},
		{
			Fichero:              legalCoreSinBinario,
			Pregunta:             "\xc2\xbfQu\xc3\xa9 comunidad y qu\xc3\xa9 boletines corresponden al Ayuntamiento de Getafe?",
			Activa:               true,
			NoSeActivan:          []string{"boe-legislacion"},
			SinBinarioNiServidor: true,
		},
	}
}

// probarReglasDeLegalCore fija ComprobarConjunto con ReglasDeLegalCore (contrato
// de evals §3 de H6; FR-081, FR-082): el conjunto que las cumple no da ningún
// defecto, y tampoco con una eval activa más que solo espera citas; cada copia que
// incumple una regla da el defecto de esa regla, con su nombre y un mensaje que
// nombra los ficheros implicados; y un conjunto que las incumple todas da un
// defecto por regla, en el orden de la tabla. Desde H21, el conjunto lleva
// cuatro evals, con la que es sin binario ni servidor, que no necesita un
// esperado verificable: sin una de las cuatro incumple solo la regla de esa eval,
// sin dos incumple también el tamaño, y con dos evals sin binario ni servidor,
// solo la regla nueva (contracts/evals-en-dos-modos.md §1 de H21; FR-046).
func probarReglasDeLegalCore(t *testing.T) {
	t.Parallel()

	t.Run("cumple-todas", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, ComprobarConjunto(conjuntoDeLegalCore(), nil, ReglasDeLegalCore()))

		conCitas := append(conjuntoDeLegalCore(), Eval{
			Fichero:  legalCoreConCitas,
			Pregunta: preguntaDelArticulo21Eval,
			Activa:   true,
			Comandos: []ComandoEsperado{{Applet: "boe", Norma: normaLPAC, Bloque: "a21"}},
			Citas:    []CitaEsperada{{Norma: normaLPAC, Bloque: "a21"}},
		})
		assert.Empty(t, ComprobarConjunto(conCitas, nil, ReglasDeLegalCore()),
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
			modificar: sinLaEval(legalCoreCubierto),
			reglas:    []string{"cubierto"},
			ficheros:  []string{legalCoreNoCubierto},
		},
		{
			nombre:    "sin-la-del-no-cubierto",
			modificar: sinLaEval(legalCoreNoCubierto),
			reglas:    []string{"no cubierto"},
			ficheros:  []string{legalCoreCubierto},
		},
		{
			nombre:    "sin-la-de-no-activación",
			modificar: sinLaEval(legalCoreNoActivacion),
			reglas:    []string{"no activación"},
			ficheros:  []string{legalCoreCubierto, legalCoreNoCubierto, legalCoreSinBinario},
		},
		{
			nombre:    "sin-la-eval-sin-binario-ni-servidor",
			modificar: sinLaEval(legalCoreSinBinario),
			reglas:    []string{"sin binario ni servidor"},
			dice:      []string{"hay 0 evals sin binario ni servidor", "exactamente 1: ning\xc3\xban fichero"},
		},
		{
			nombre: "dos-evals-sin-binario-ni-servidor",
			modificar: func(_ *testing.T, evals []Eval) []Eval {
				return append(evals, Eval{
					Fichero:  legalCoreOtraSinBinario,
					Pregunta: "\xc2\xbfQu\xc3\xa9 boletines corresponden al Ayuntamiento de M\xc3\xb3stoles?",
					Activa:   true, SinBinarioNiServidor: true,
				})
			},
			reglas:   []string{"sin binario ni servidor"},
			ficheros: []string{legalCoreSinBinario, legalCoreOtraSinBinario},
			dice:     []string{"hay 2 evals sin binario ni servidor"},
		},
		{
			// Con dos evals menos, el conjunto ya no llega a las tres del mínimo, y el
			// defecto del tamaño va el primero, delante de los de las dos evals que
			// faltan.
			nombre: "tama\xc3\xb1o",
			modificar: func(t *testing.T, evals []Eval) []Eval {
				t.Helper()

				return sinLaEval(legalCoreSinBinario)(t, sinLaEval(legalCoreNoActivacion)(t, evals))
			},
			reglas: []string{"tama\xc3\xb1o", "no activaci\xc3\xb3n", "sin binario ni servidor"},
			dice:   []string{"hay 0 evals sin binario ni servidor"},
		},
		{
			nombre: "cubierto-sin-boletines",
			modificar: func(t *testing.T, evals []Eval) []Eval {
				t.Helper()
				evalDe(t, evals, legalCoreCubierto).Territorio.Boletines = nil

				return evals
			},
			reglas:   []string{"cubierto"},
			ficheros: []string{legalCoreCubierto, legalCoreNoCubierto},
		},
		{
			nombre: "cubierto-sin-comando-de-territorio",
			modificar: func(t *testing.T, evals []Eval) []Eval {
				t.Helper()
				evalDe(t, evals, legalCoreCubierto).Comandos = []ComandoEsperado{
					{Applet: "boe", Verbo: "buscar", Terminos: []string{"Leganés"}},
				}

				return evals
			},
			reglas:   []string{"cubierto"},
			ficheros: []string{legalCoreNoCubierto},
		},
		{
			nombre: "cubierto-sin-activar",
			modificar: func(t *testing.T, evals []Eval) []Eval {
				t.Helper()
				evalDe(t, evals, legalCoreCubierto).Activa = false

				return evals
			},
			reglas:   []string{"cubierto"},
			ficheros: []string{legalCoreNoCubierto},
		},
		{
			// Con boletines, la eval de una comunidad sin configuración sigue sin
			// ser la de un municipio del territorio configurado.
			nombre: "no-cubierto-con-boletines-no-es-cubierto",
			modificar: func(t *testing.T, evals []Eval) []Eval {
				t.Helper()
				evalDe(t, evals, legalCoreNoCubierto).Territorio.Boletines = []string{"BOE"}

				return sinLaEval(legalCoreCubierto)(t, evals)
			},
			reglas:   []string{"cubierto"},
			ficheros: []string{legalCoreNoCubierto},
		},
		{
			nombre: "no-cubierto-con-un-solo-aspecto",
			modificar: func(t *testing.T, evals []Eval) []Eval {
				t.Helper()
				evalDe(t, evals, legalCoreNoCubierto).Territorio.Cobertura = []string{"boletin_autonomico: no-configurado"}

				return evals
			},
			reglas:   []string{"no cubierto"},
			ficheros: []string{legalCoreCubierto, legalCoreNoCubierto},
		},
		{
			nombre: "esperado-verificable",
			modificar: func(_ *testing.T, evals []Eval) []Eval {
				return append(evals, Eval{
					Fichero:  legalCoreSinVerificable,
					Pregunta: "¿Qué boletín publica las ordenanzas de Leganés?",
					Activa:   true,
					Comandos: []ComandoEsperado{{Applet: "territorio", Verbo: "resolver", Municipio: "Leganés"}},
				})
			},
			reglas:   []string{"esperado verificable"},
			ficheros: []string{legalCoreSinVerificable},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			defectos := ComprobarConjunto(caso.modificar(t, conjuntoDeLegalCore()), nil, ReglasDeLegalCore())
			require.Equal(t, caso.reglas, reglasIncumplidas(defectos), "la copia incumple solo esas reglas: %v", defectos)

			for _, nombrado := range slices.Concat(caso.ficheros, caso.dice) {
				assert.Contains(t, defectos[len(defectos)-1].Mensaje, nombrado, "el mensaje nombra lo implicado")
			}
		})
	}

	t.Run("orden-de-la-tabla", func(t *testing.T) {
		t.Parallel()

		sinEsperado := []Eval{{
			Fichero:  legalCoreSinVerificable,
			Pregunta: "¿Qué boletín publica las ordenanzas de Leganés?",
			Activa:   true,
			Comandos: []ComandoEsperado{{Applet: "territorio", Verbo: "resolver", Municipio: "Leganés"}},
		}}

		assert.Equal(t, []string{
			"tamaño", "cubierto", "no cubierto", "no activación", "esperado verificable", "sin binario ni servidor",
		}, reglasIncumplidas(ComprobarConjunto(sinEsperado, nil, ReglasDeLegalCore())))
	})

	t.Run("sin-configuración-ningún-boletín", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, []string{"boletin_autonomico: no-configurado", "boletin_provincial: no-configurado"},
			aspectosNoConfigurados(), "una comunidad sin configuración tiene sin configurar sus dos boletines")
	})
}

// sinLaEval devuelve la modificación de probarReglasDeLegalCore que quita del
// conjunto la eval con ese fichero, que tiene que estar.
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
	// evalsDelRepositorio es evals/boe-legislacion/, relativo al directorio de
	// este paquete, que es donde go test ejecuta sus tests (research.md V46).
	evalsDelRepositorio = "../../evals/boe-legislacion"

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

	// skillDelRepositorio es el SKILL.md de boe-legislacion, relativo al
	// directorio de este paquete: la skill cuyas respuestas juzgan sus evals.
	skillDelRepositorio = "../../skills/boe-legislacion/SKILL.md"

	// evalsDeLegalCore es evals/legal-core/, relativo al directorio de este
	// paquete: las evals de la skill legal-core, a las que se aplican las reglas
	// de su juego (contrato de evals §3 y §4 de H6).
	evalsDeLegalCore = "../../evals/legal-core"
)

// TestEvalsDelRepositorio comprueba sin red las evals del repositorio (contrato
// evals-y-grabaciones §2 y §5.2; FR-060, FR-062 a FR-066, FR-074, FR-075,
// SC-010): cada fichero de cada directorio evals/<skill>/ se lee como eval, y las
// de evals/boe-legislacion/ cumplen las reglas del conjunto de data-model §6.3,
// esperan solo normas de data/normas.yaml y tienen en las grabaciones de H4 y de
// H5 lo que necesitan para servir sin red cada consulta; el esquema publicado
// del formato admite en avisos exactamente los códigos de aviso del binario
// (FR-013 de H5.1); el SKILL.md de boe-legislacion lleva la forma fija de cada
// uno de esos códigos, reconocida con la misma función que usa Juzgar (FR-014 de
// H5.1); el esquema publicado admite en hallazgos exactamente las clases de
// hallazgo que el binario etiqueta (contrato evals-y-skill §3 de H7.1; FR-054);
// el SKILL.md de boe-legislacion lleva la forma fija de cada una de esas clases,
// reconocida con la misma función que usa Juzgar (FR-045, FR-093 de H7.1);
// el esquema publicado admite en la cobertura del territorio esperado
// exactamente las combinaciones del vocabulario del applet territorio (contrato de
// evals §1.2 de H6); las de evals/legal-core/ cumplen las reglas del conjunto de
// legal-core (contrato de evals §3 de H6; FR-080 a FR-082, SC-011); y el grafo
// previo de cada eval que lo lleva existe y se prepara sin faltas, con un
// BloqueVersion por comando (contrato evals-y-skill §3 de H7; FR-085). Desde
// H7.2, sobre
// cada grafo previo, la lectura de los bloques de la eval y graph check dan los
// hallazgos que la eval espera, de la redacción que dejó el grafo previo a la
// leída (contrato eval-y-derivada §4; FR-002 de H7.2). Desde H7.3, ninguna
// expresión de la lista de expresiones prohibidas de evals/boe-legislacion/ va
// en la prosa del SKILL.md de boe-legislacion —fuera
// del código y de la región generada— ni el fichero lleva ninguna fecha AAAAMMDD
// escrita con cifras (contracts/skill-boe-legislacion.md §5; FR-091, SC-005).
// Desde H7.4, cada orden de lectura y
// comprobación de la skill lleva detrás su forma para PowerShell
// (contracts/skill-boe-legislacion.md
// §4 de H7.4; FR-095; SC-005). Desde
// H21, las 21 evals de boe-legislacion y las 4 de legal-core cumplen además la
// regla sin binario ni servidor de su juego: cada carpeta lleva exactamente una
// eval que lo declara (contracts/evals-en-dos-modos.md §1 de H21; FR-046). Desde
// H24, la carpeta del juez de cada skill que la tiene se lee con sus evals y no
// da ningún fichero mal formado: la de boe-legislacion declara sus dos clases,
// afirma_lo_no_leido, que decide, y cuenta_su_proceso, que solo se publica, las
// dos con umbral 0, y legal-core no tiene juez (contracts/juez-y-voto.md §1 de
// H24; FR-020, FR-021); y la lista deja de compararse con respuestas, con los
// bloques grabados y con las formas de la skill, porque ya no juzga ninguna
// respuesta: del calibrado de H7.4 quedan sus tres totales, como premisa de
// TestJuzgarSinLaLista (FR-071 de H24). Lee las carpetas enteras, así que ningún
// fichero de eval se nombra aquí.
func TestEvalsDelRepositorio(t *testing.T) {
	t.Parallel()

	conjunto, errDeLaLectura := LeerConjunto(evalsDelRepositorio)

	normas, err := skills.LeerNormas(contenidoDelFichero(t, tablaDeNormasDelRepositorio))
	require.NoError(t, err, "la tabla de normas %s", tablaDeNormasDelRepositorio)

	// Una sola llamada: conjunto y normas-conocidas presentan cada uno una parte
	// de sus defectos.
	defectos := ComprobarConjunto(conjunto.Evals, normasConocidasDe(normas), ReglasDeBoeLegislacion())
	esDeNormasConocidas := func(defecto DefectoDelConjunto) bool { return defecto.Regla == reglaDeNormasConocidas }

	t.Run("formato", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, errDeLaLectura)
		assert.GreaterOrEqual(t, len(conjunto.Evals), evalsMinimasDelRepositorio,
			"%s tiene al menos las doce evals bien formadas del contrato", evalsDelRepositorio)

		// El formato común es el de las evals de cualquier skill (FR-060): se leen
		// todas las carpetas de evals/, la de boe-legislacion entre ellas.
		carpetas, malFormados := malFormadosDeCadaSkill(t, directorioDeEvals)

		require.Contains(t, carpetas, evalsDelRepositorio)
		assert.Empty(t, malFormados, "ficheros mal formados en %s:\n%s",
			directorioDeEvals, strings.Join(malFormados, "\n"))

		// La carpeta del juez se lee con las evals: boe-legislacion declara sus
		// dos clases y legal-core no tiene juez (FR-020 y FR-021 de H24).
		require.NotNil(t, conjunto.Juez, "%s tiene la carpeta del juez", evalsDelRepositorio)
		assert.Equal(t, []ClaseDelJuez{
			{Nombre: "afirma_lo_no_leido", Decide: true, Umbral: 0},
			{Nombre: "cuenta_su_proceso", Decide: false, Umbral: 0},
		}, conjunto.Juez.Clases)

		deLegalCore, err := LeerConjunto(evalsDeLegalCore)
		require.NoError(t, err)
		assert.Nil(t, deLegalCore.Juez, "%s no tiene juez", evalsDeLegalCore)
	})

	t.Run("conjunto", func(t *testing.T) {
		t.Parallel()

		deOtrasReglas := slices.DeleteFunc(slices.Clone(defectos), esDeNormasConocidas)
		assert.Empty(t, deOtrasReglas, "defectos del conjunto de %s:\n%s",
			evalsDelRepositorio, presentarDefectos(deOtrasReglas))
	})

	t.Run("conjunto-legal-core", func(t *testing.T) {
		t.Parallel()

		deLegalCore, err := LeerConjunto(evalsDeLegalCore)
		require.NoError(t, err)

		defectosDeLegalCore := ComprobarConjunto(deLegalCore.Evals, normasConocidasDe(normas), ReglasDeLegalCore())
		assert.Empty(t, defectosDeLegalCore, "defectos del conjunto de %s:\n%s",
			evalsDeLegalCore, presentarDefectos(defectosDeLegalCore))
	})

	t.Run("normas-conocidas", func(t *testing.T) {
		t.Parallel()

		// Con la regla nombrada de otra forma, este subtest no vería sus defectos
		// y pasaría en vacío.
		require.True(t, slices.ContainsFunc(ReglasDeBoeLegislacion(), func(regla ReglaDelConjunto) bool {
			return regla.nombre == reglaDeNormasConocidas
		}), "la regla %q es una de las del conjunto", reglaDeNormasConocidas)

		desconocidas := slices.DeleteFunc(slices.Clone(defectos), func(defecto DefectoDelConjunto) bool {
			return !esDeNormasConocidas(defecto)
		})
		assert.Empty(t, desconocidas, "normas de las evals de %s que no están en %s:\n%s",
			evalsDelRepositorio, tablaDeNormasDelRepositorio, presentarDefectos(desconocidas))
	})

	t.Run("grabado", func(t *testing.T) {
		t.Parallel()

		consultas := ConsultasNecesarias(conjunto.Evals)
		require.NotEmpty(t, consultas, "las evals de %s necesitan alguna consulta", evalsDelRepositorio)

		dirCache := t.TempDir()

		preparadas, err := Preparar(dirCache, UnionDeGrabaciones(), consultas)
		require.NoError(t, err)
		assert.Empty(t, preparadas, "consultas de las evals de %s sin su respuesta en las grabaciones %s:\n%s",
			evalsDelRepositorio, strings.Join(UnionDeGrabaciones(), " y "), presentarFaltas(preparadas))

		comprobadas, err := ComprobarSinRed(dirCache, consultas)
		require.NoError(t, err)
		assert.Empty(t, comprobadas, "consultas de las evals de %s que la caché preparada no sirve sin red:\n%s",
			evalsDelRepositorio, presentarFaltas(comprobadas))
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

		skill := contenidoDelFichero(t, skillDelRepositorio)

		assert.NoError(t, ComprobarFormasDeAviso(string(skill)),
			"%s enseña la forma fija de cada código de aviso del binario", skillDelRepositorio)
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

		skill := contenidoDelFichero(t, skillDelRepositorio)

		assert.NoError(t, ComprobarFormasDeHallazgo(string(skill)),
			"%s enseña la forma fija de cada clase de hallazgo etiquetada por el binario", skillDelRepositorio)
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

		probarProsaDeLaSkill(t, listaDelRepositorio(t, conjunto))
	})

	t.Run("ordenes-para-powershell", func(t *testing.T) {
		t.Parallel()

		defectos := defectosDeLasOrdenesParaPowerShell(string(contenidoDelFichero(t, skillDelRepositorio)))
		assert.Empty(t, defectos, "%s enseña cada orden de lectura y comprobación con su forma para PowerShell:\n%s",
			skillDelRepositorio, strings.Join(defectos, "\n"))
	})

	t.Run("linea-sin-consulta", probarLineaSinConsulta)
}

const (
	// raizDeLasSkills es la raíz del repositorio, relativa al directorio de este
	// paquete: en su directorio skills/ está cada skill con su SKILL.md.
	raizDeLasSkills = "../.."

	// lineaSinConsultaDeLasSkills es la línea que cada SKILL.md enseña para la
	// respuesta de quien no tiene ni la herramienta ni el binario, con el
	// marcador de la causa, carácter a carácter (contracts/skills.md §3 de H21).
	lineaSinConsultaDeLasSkills = "⚠ SIN CONSULTA AL BOE: <causa>. Para consultarlo hace falta instalar kitlegal: " +
		"https://kitlegal.es/instalar/"
)

// skillsConLineaSinConsulta son las skills que skills/ tiene que tener para que
// la subprueba linea-sin-consulta no pase en vacío para ninguna de las dos que
// llevan la regla (FR-035 de H21).
var skillsConLineaSinConsulta = []string{"boe-legislacion", "legal-core"}

// probarLineaSinConsulta es la subprueba linea-sin-consulta de
// TestEvalsDelRepositorio (contracts/skills.md §5 de H21; FR-035, FR-077): el
// SKILL.md de cada skill de skills/ dice la línea de contracts/skills.md §3 tal
// cual en un bloque de código text, y esa línea casa con lo que el juicio
// reconoce, ExtraerSinConsulta, con su dirección. Cada defecto nombra la skill.
func probarLineaSinConsulta(t *testing.T) {
	t.Parallel()

	nombres, err := skills.Listar(raizDeLasSkills)
	require.NoError(t, err)
	require.Subset(t, nombres, skillsConLineaSinConsulta, "skills/ tiene las skills que llevan la regla")

	var defectos []string

	for _, nombre := range nombres {
		skill, err := skills.Cargar(raizDeLasSkills, nombre)
		require.NoError(t, err)

		for _, defecto := range defectosDeLaLineaSinConsulta(string(skill.Contenido)) {
			defectos = append(defectos, nombre+": "+defecto)
		}
	}

	assert.Empty(t, defectos, "cada SKILL.md de skills/ enseña la línea %s en un bloque text:\n%s",
		formaEscrita(etiquetaSinConsulta), strings.Join(defectos, "\n"))
}

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
				strings.Replace(lineaSinConsultaDeLasSkills, "hace falta", "tienes que", 1)),
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

// Los informes del job de evals de boe-legislacion en el cierre de H7.1, de
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
// evals/boe-legislacion/, la que deja en el conjunto el LeerConjunto de
// TestEvalsDelRepositorio. Tiene expresiones en sus cuatro familias, las de lo
// que devuelven las herramientas y formas fijas: sin ellas, los tests que la
// aplican pasarían en vacío.
func listaDelRepositorio(t *testing.T, conjunto Conjunto) ExpresionesProhibidas {
	t.Helper()

	lista := conjunto.Prohibidas
	require.NotEmpty(t, lista.Maquinaria, "%s tiene lista de expresiones prohibidas, con las de la maquinaria",
		evalsDelRepositorio)
	require.NotEmpty(t, lista.OtraConversacion, "%s tiene lista de expresiones prohibidas, con las de otra conversación",
		evalsDelRepositorio)
	require.NotEmpty(t, lista.Anuncio, "%s tiene lista de expresiones prohibidas, con las del anuncio",
		evalsDelRepositorio)
	require.NotEmpty(t, lista.RedaccionNoLeida, "%s tiene lista de expresiones prohibidas, con las de la redacción "+
		"no leída", evalsDelRepositorio)
	require.NotEmpty(t, lista.SalidaDeLasHerramientas, "%s tiene lista de expresiones prohibidas, con las de lo que "+
		"devuelven las herramientas", evalsDelRepositorio)
	require.NotEmpty(t, lista.FormasFijas, "%s tiene lista de expresiones prohibidas, con sus formas fijas",
		evalsDelRepositorio)

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

// probarProsaDeLaSkill es la subprueba prosa-de-la-skill de
// TestEvalsDelRepositorio (contracts/skill-boe-legislacion.md §5; research D4;
// FR-010, FR-013, FR-091, SC-005): con defectosDeLaProsa, ningún párrafo de la
// prosa del SKILL.md de boe-legislacion, frontmatter incluido, lleva ninguna
// expresión de la lista, y ninguna línea del fichero lleva una fecha AAAAMMDD
// escrita con cifras. El fichero tiene prosa: sin ella, la subprueba pasaría en
// vacío. Desde H24 (contracts/skill-boe-legislacion.md §4; FR-085, FR-110,
// SC-010), tampoco lleva ninguna de las de lo que devuelven las herramientas,
// ni escrita como código en línea: es el control de ese umbral, 0 usos.
func probarProsaDeLaSkill(t *testing.T, lista ExpresionesProhibidas) {
	t.Helper()

	skill := string(contenidoDelFichero(t, skillDelRepositorio))
	require.NotEmpty(t, parrafosDeLaProsa(skill), "%s tiene prosa fuera del código y de la región generada",
		skillDelRepositorio)

	defectos := defectosDeLaProsa(skill, lista)
	assert.Empty(t, defectos, "%s enseña con su prosa expresiones que la respuesta no lleva, o escribe con cifras "+
		"una fecha AAAAMMDD:\n%s", skillDelRepositorio, strings.Join(defectos, "\n"))
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
			"carácter informativo.\n" //nolint:misspell // «informativo» es español: el párrafo va tal cual.
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
			markdown: "---\nname: boe-legislacion\ndescription: >-\n  Traslada los hallazgos.\n---\n\n" +
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
// orden: cada línea de un bloque delimitado y cada tramo de código en línea fuera
// de ellos que leerOrdenDeLectura lee como una. Como en CommonMark, un bloque que
// no se cierra llega hasta el final.
func ordenesDeLectura(markdown string) []ordenDeLectura {
	var (
		ordenes  []ordenDeLectura
		numero   int
		enBloque bool
	)

	for linea := range strings.Lines(markdown) {
		numero++
		recortada := strings.TrimSpace(linea)

		var tramos []string

		switch {
		case strings.HasPrefix(recortada, delimitadorDeBloque):
			enBloque = !enBloque
		case enBloque:
			tramos = []string{recortada}
		default:
			for _, codigo := range codigoEnLinea.FindAllString(recortada, -1) {
				tramos = append(tramos, strings.Trim(codigo, "`"))
			}
		}

		for _, tramo := range tramos {
			if orden, conForma := leerOrdenDeLectura(numero, tramo); conForma {
				ordenes = append(ordenes, orden)
			}
		}
	}

	return ordenes
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
// que la precede no tiene la suya. Nil si no hay ningún defecto.
func defectosDeLasOrdenesParaPowerShell(markdown string) []string {
	ordenes := ordenesDeLectura(markdown)

	var defectos []string

	for _, verbo := range verbosDeLectura {
		if !slices.ContainsFunc(ordenes, func(orden ordenDeLectura) bool {
			return !orden.powershell && orden.verbo == verbo
		}) {
			defectos = append(defectos, fmt.Sprintf("ninguna orden de Bash lee con «kitlegal boe %s» y comprueba "+
				"detrás con kitlegal graph check", verbo))
		}
	}

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
		bashDeArticulo = "kitlegal boe articulo BOE-A-2015-10565 a21 --json && " +
			"kitlegal graph check BOE-A-2015-10565 a21 --json"
		powershellDeArticulo = "kitlegal boe articulo BOE-A-2015-10565 a21 --json; if ($LASTEXITCODE -eq 0) { " +
			"kitlegal graph check BOE-A-2015-10565 a21 --json }"
		bashDeArticulos = "kitlegal boe articulos <norma> <bloques>... --json && " +
			"kitlegal graph check <norma> <bloques>... --json"
		powershellDeArticulos = "kitlegal boe articulos <norma> <bloques>... --json; if ($LASTEXITCODE -eq 0) { " +
			"kitlegal graph check <norma> <bloques>... --json }"

		sinSuForma = ": no la sigue su forma para PowerShell, «…; if ($LASTEXITCODE -eq 0) { … }», con el mismo " +
			"verbo, la misma norma y los mismos bloques"
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
// las evals de cada skill y no solo las de boe-legislacion (FR-060): con una eval
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
