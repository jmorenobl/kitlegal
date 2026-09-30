package evals

import (
	"bytes"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io/fs"
	"maps"
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
	"github.com/jmorenobl/kitlegal/internal/source/boe"
)

// Contenidos de las evals sintéticas de TestLeerConjunto. La del art. 21 se
// escribe con los trozos de TestLeerEval.
const (
	contenidoDelArticulo21     = preguntaDelArticulo21 + "activa: true\n" + comandoDelArticulo21 + citaDelArticulo21
	contenidoDeProgramacion    = "pregunta: \"¿Cómo invierto una lista enlazada en Go?\"\nactiva: false\n"
	contenidoSinPregunta       = "activa: false\n"
	contenidoConActivaRepetida = preguntaDelArticulo21 + "activa: true\nactiva: false\n"
)

// entradaDeConjunto es una entrada que un test crea en el directorio de un
// conjunto de evals: un fichero con su contenido o, con carpeta, un
// subdirectorio vacío.
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
// formada, queda en Conjunto.Prohibidas y en Prohibidas de cada eval, con sus
// tres familias desde H7.3 (FR-023, FR-024); si no es un fichero regular o no
// valida —también la de dos familias de H7.2, sin anuncio—, es un fichero mal
// formado que la nombra y las evals se leen sin lista; y una carpeta sin ella se
// lee como antes del hito, sin lista en el conjunto ni en ninguna eval.
//
// Desde H7.4 (contracts/lista-de-expresiones.md §1 y §2; FR-030, FR-031), la
// lista bien formada tiene sus cinco claves, y la familia redaccion_no_leida y
// las formas_fijas quedan también en Conjunto.Prohibidas y en Prohibidas de cada
// eval; y cada lista mal formada lo está solo por su defecto.
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

	// Lo que se lee de listaBienFormada, la del test del esquema con sus cinco
	// claves; y las dos claves de H7.4 bien formadas, para que la lista sin una
	// familia de H7.3 lo esté solo por eso.
	lista := ExpresionesProhibidas{
		Maquinaria:       []string{"memoria de consultas", "hallazgos", "c\xc3\xb3digo de salida"},
		OtraConversacion: []string{"te dije", "conversaci\xc3\xb3n anterior"},
		Anuncio:          []string{"que trasladar", "ya puedo responder", "as\xc3\xad que respondo"},
		RedaccionNoLeida: []string{"ya no exige", "se elimin\xc3\xb3"},
		FormasFijas: []string{
			"\xe2\x9a\xa0 REDACCI\xc3\x93N MODIFICADA: <cita>: vigente desde <fecha>.",
			"No se ha podido comprobar si la redacci\xc3\xb3n ha cambiado",
		},
	}
	clavesDeH74 := redaccionNoLeidaBienFormada + formasFijasBienFormadas
	conLista := func(eval Eval) Eval {
		eval.Prohibidas = lista

		return eval
	}

	casos := []struct {
		nombre      string
		entradas    []entradaDeConjunto
		evals       []Eval
		malFormados []malFormadoEsperado

		// prohibidas es la lista que tiene que quedar en el conjunto.
		prohibidas ExpresionesProhibidas
	}{
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
			evals:      []Eval{conLista(leidaDelArticulo21), conLista(leidaDeProgramacion)},
			prohibidas: lista,
		},
		{
			nombre: "lista-sin-una-familia",
			entradas: []entradaDeConjunto{
				{nombre: "02-no-activa-programacion.yaml", contenido: contenidoDeProgramacion},
				{
					nombre:    ficheroDeExpresionesProhibidas,
					contenido: maquinariaBienFormada + anuncioBienFormado + clavesDeH74,
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
					contenido: maquinariaBienFormada + otraConversacionBienFormada + clavesDeH74,
				},
				{nombre: nombreDeEval, contenido: contenidoDelArticulo21},
			},
			evals: []Eval{leidaDelArticulo21},
			malFormados: []malFormadoEsperado{
				{fichero: ficheroDeExpresionesProhibidas, fragmento: "missing property 'anuncio'"},
			},
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
				{fichero: ficheroDeExpresionesProhibidas, fragmento: "maquinaria repetido en las l\xc3\xadneas 1 y 18"},
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
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			conjunto, err := LeerConjunto(crearConjunto(t, caso.entradas))
			require.NoError(t, err, "un fichero mal formado nunca es el error de la lectura")
			assert.Equal(t, caso.evals, conjunto.Evals)
			assert.Equal(t, caso.prohibidas, conjunto.Prohibidas)

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

	for _, entrada := range entradas {
		ruta := filepath.Join(dir, entrada.nombre)
		if entrada.carpeta {
			require.NoError(t, os.Mkdir(ruta, 0o750))

			continue
		}

		require.NoError(t, os.WriteFile(ruta, []byte(entrada.contenido), 0o600))
	}

	return dir
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
// de evals §2, con diez positivas de una norma cada una y dos de no activación.
// La 07 y la 08 citan además una norma común, porque una positiva que cita dos
// normas cuenta como materia distinta si al menos una no la cita ninguna otra
// (spec, casos límite).
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
		evals: make([]Eval, 0, len(positivas)+2),
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
// (contrato de evals §3 de H6; research D22): las diez de boe-legislacion, que no
// cambian, y las cinco de legal-core, cada juego sobre evals sintéticas.
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
// defecto por regla, en el orden de la tabla.
func probarReglasDeBoeLegislacion(t *testing.T) {
	t.Parallel()

	todos := []string{
		"01-lpac-articulo-21.yaml", "02-lcsp-contrato-menor.yaml", "03-lrbrl-atribuciones-del-pleno.yaml",
		"04-lgt-prescripcion.yaml", "05-trlrhl-impuestos-municipales.yaml", "06-irpf-rendimientos-del-trabajo.yaml",
		"07-lrjsp-principio-de-legalidad.yaml", "08-ltaibg-plazo-de-resolucion.yaml",
		"09-constitucion-articulo-140.yaml", "10-et-vacaciones.yaml",
		"11-no-activa-programacion.yaml", "12-no-activa-acuerdo-entre-amigos.yaml",
		"13-lrbrl-atribuciones-por-materia.yaml",
	}
	positivas := todos[:10]
	informativa := todos[12]
	deMas := []string{
		"14-no-activa-14.yaml", "15-no-activa-15.yaml", "16-no-activa-16.yaml",
		"17-no-activa-17.yaml", "18-no-activa-18.yaml", "19-no-activa-19.yaml", "20-no-activa-20.yaml",
		"21-no-activa-21.yaml", "22-no-activa-22.yaml",
	}

	t.Run("cumple-todas", func(t *testing.T) {
		t.Parallel()

		conjunto := conjuntoQueCumple(t)
		require.Equal(t, todos, ficherosDelConjunto(conjunto), "el conjunto sintético es el del contrato")

		assert.Empty(t, ComprobarConjunto(conjunto.evals, conjunto.normas, ReglasDeBoeLegislacion()))
	})

	casos := []struct {
		nombre    string
		modificar func(t *testing.T, conjunto *conjuntoSintetico)
		regla     string
		ficheros  []string
		normas    []string
	}{
		{
			nombre: "tamaño",
			modificar: func(_ *testing.T, conjunto *conjuntoSintetico) {
				for _, fichero := range deMas {
					conjunto.evals = append(conjunto.evals, Eval{Fichero: fichero, Pregunta: "¿Qué hora es?"})
				}
			},
			regla:    "tamaño",
			ficheros: slices.Concat(todos, deMas),
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
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			conjunto := conjuntoQueCumple(t)
			caso.modificar(t, &conjunto)

			defectos := ComprobarConjunto(conjunto.evals, conjunto.normas, ReglasDeBoeLegislacion())
			require.Len(t, defectos, 1, "la copia solo incumple la regla %s: %v", caso.regla, defectos)
			assert.Equal(t, caso.regla, defectos[0].Regla)

			for _, nombrado := range slices.Concat(caso.ficheros, caso.normas) {
				assert.Contains(t, defectos[0].Mensaje, nombrado, "el mensaje nombra lo implicado")
			}
		})
	}

	t.Run("orden-de-la-tabla", func(t *testing.T) {
		t.Parallel()

		conjunto := conjuntoQueCumple(t)
		conjunto.evals = nil

		for numero := 1; numero <= 21; numero++ {
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
			"art. 21", "fiscal", "boe-fiscal", "normas conocidas",
		}, reglasIncumplidas(ComprobarConjunto(conjunto.evals, conjunto.normas, ReglasDeBoeLegislacion())))
	})
}

// Las evals sintéticas del conjunto de legal-core de probarReglasDeLegalCore: la
// del municipio cubierto, la del no cubierto y la de no activación (contrato de
// evals §4 de H6).
const (
	legalCoreCubierto       = "01-territorio-municipio-cubierto.yaml"
	legalCoreNoCubierto     = "02-territorio-municipio-no-cubierto.yaml"
	legalCoreNoActivacion   = "03-no-activa-receta-de-cocina.yaml"
	legalCoreSinVerificable = "04-territorio-sin-esperado.yaml"
	legalCoreConCitas       = "04-articulo-21-con-citas.yaml"
)

// conjuntoDeLegalCore devuelve, nuevo en cada llamada, un conjunto que cumple las
// cinco reglas de legal-core: una eval activa que resuelve un municipio del
// territorio configurado y espera su boletín, otra que resuelve uno de una
// comunidad sin configuración y espera no configurado cada aspecto de boletín, y
// una de no activación.
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
	}
}

// probarReglasDeLegalCore fija ComprobarConjunto con ReglasDeLegalCore (contrato
// de evals §3 de H6; FR-081, FR-082): el conjunto que las cumple no da ningún
// defecto, y tampoco con una eval activa más que solo espera citas; cada copia que
// incumple una regla da el defecto de esa regla, con su nombre y un mensaje que
// nombra los ficheros implicados —y el de tamaño con él cuando le quita una de sus
// tres evals, porque las otras tres reglas no las cumple una misma eval—; y un
// conjunto que las incumple todas da un defecto por regla, en el orden de la
// tabla.
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
	}{
		{
			nombre:    "sin-la-del-cubierto",
			modificar: sinLaEval(legalCoreCubierto),
			reglas:    []string{"tamaño", "cubierto"},
			ficheros:  []string{legalCoreNoCubierto},
		},
		{
			nombre:    "sin-la-del-no-cubierto",
			modificar: sinLaEval(legalCoreNoCubierto),
			reglas:    []string{"tamaño", "no cubierto"},
			ficheros:  []string{legalCoreCubierto},
		},
		{
			nombre:    "sin-la-de-no-activación",
			modificar: sinLaEval(legalCoreNoActivacion),
			reglas:    []string{"tamaño", "no activación"},
			ficheros:  []string{legalCoreCubierto, legalCoreNoCubierto},
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
			reglas:   []string{"tamaño", "cubierto"},
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

			for _, fichero := range caso.ficheros {
				assert.Contains(t, defectos[len(defectos)-1].Mensaje, fichero, "el mensaje nombra lo implicado")
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

		assert.Equal(t, []string{"tamaño", "cubierto", "no cubierto", "no activación", "esperado verificable"},
			reglasIncumplidas(ComprobarConjunto(sinEsperado, nil, ReglasDeLegalCore())))
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
// H7.2, la lista de expresiones prohibidas de evals/boe-legislacion/ marca en las
// respuestas de H7.1 el reparto calibrado, y ninguna de sus expresiones casa con
// el texto de los bloques que leen las evals ni con lo que la skill enseña a
// escribir (contrato lista-y-juicio §6; FR-043, FR-084, FR-085 de H7.2); y, sobre
// cada grafo previo, la lectura de los bloques de la eval y graph check dan los
// hallazgos que la eval espera, de la redacción que dejó el grafo previo a la
// leída (contrato eval-y-derivada §4; FR-002 de H7.2). Desde H7.3, ninguna
// expresión de esa lista va en la prosa del SKILL.md de boe-legislacion —fuera
// del código y de la región generada— ni el fichero lleva ninguna fecha AAAAMMDD
// escrita con cifras (contracts/skill-boe-legislacion.md §5; FR-091, SC-005). Lee
// las carpetas enteras, así que ningún fichero de eval se nombra aquí.
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

	t.Run("expresiones-calibradas", func(t *testing.T) {
		t.Parallel()

		probarExpresionesCalibradas(t, listaDelRepositorio(t, conjunto))
	})

	t.Run("expresiones-en-los-bloques", func(t *testing.T) {
		t.Parallel()

		probarExpresionesEnLosBloques(t, conjunto.Evals, listaDelRepositorio(t, conjunto))
	})

	t.Run("expresiones-de-la-skill", func(t *testing.T) {
		t.Parallel()

		probarExpresionesDeLaSkill(t, listaDelRepositorio(t, conjunto))
	})

	t.Run("prosa-de-la-skill", func(t *testing.T) {
		t.Parallel()

		probarProsaDeLaSkill(t, listaDelRepositorio(t, conjunto))
	})
}

// Los informes del job de evals de boe-legislacion en el cierre de H7.1 y de
// H7.2, relativos al directorio de este paquete: sus respuestas son con las que
// se calibra la lista de expresiones prohibidas (FR-021 de H7.3;
// contracts/lista-de-expresiones.md §4). Están versionados y no se editan
// (FR-080).
const (
	informeDeH71 = "../../specs/011-h7-1-graph-check-acotado/gates/evals/boe-legislacion.json"
	informeDeH72 = "../../specs/012-h7-2-la-consulta-repetida/gates/evals/boe-legislacion.json"
)

// respuestasDeCadaInforme son las sesiones de cada uno de esos informes, una
// respuesta cada una.
const respuestasDeCadaInforme = 93

// listaDelRepositorio es la lista de expresiones prohibidas de
// evals/boe-legislacion/, la que deja en el conjunto el LeerConjunto de
// TestEvalsDelRepositorio. Tiene expresiones en sus tres familias: sin ellas, las
// subpruebas que la aplican pasarían en vacío.
func listaDelRepositorio(t *testing.T, conjunto Conjunto) ExpresionesProhibidas {
	t.Helper()

	lista := conjunto.Prohibidas
	require.NotEmpty(t, lista.Maquinaria, "%s tiene lista de expresiones prohibidas, con las de la maquinaria",
		evalsDelRepositorio)
	require.NotEmpty(t, lista.OtraConversacion, "%s tiene lista de expresiones prohibidas, con las de otra conversación",
		evalsDelRepositorio)
	require.NotEmpty(t, lista.Anuncio, "%s tiene lista de expresiones prohibidas, con las del anuncio",
		evalsDelRepositorio)

	return lista
}

// marcadasPorFamilia son, de las respuestas de una eval, cuántas llevan alguna
// expresión de cada familia de la lista y cuántas alguna de la lista entera.
type marcadasPorFamilia struct {
	maquinaria       int
	otraConversacion int
	anuncio          int
	alguna           int
}

// columnasDelCalibrado es el nombre de cada columna de la tabla del calibrado,
// en el orden de marcadasPorFamilia.cuentas: las tres familias, con su clave en
// la lista, y alguna.
var columnasDelCalibrado = []string{"maquinaria", "otra_conversacion", "anuncio", "alguna"}

// cuentas son las cuentas del reparto en el orden de columnasDelCalibrado.
func (m marcadasPorFamilia) cuentas() []int {
	return []int{m.maquinaria, m.otraConversacion, m.anuncio, m.alguna}
}

// informeCalibrado es un informe versionado del job de evals con el reparto
// calibrado de sus respuestas, por las dos cifras del fichero de la eval; las
// evals que no están en él, ninguna en ninguna columna.
type informeCalibrado struct {
	ruta      string
	calibrado map[string]marcadasPorFamilia
}

// informesCalibrados son los dos informes de la calibración con la tabla de
// contracts/lista-de-expresiones.md §4: en H7.1, 35 de las 93 respuestas llevan
// alguna, y en H7.2, 10.
var informesCalibrados = []informeCalibrado{
	{
		ruta: informeDeH71,
		calibrado: map[string]marcadasPorFamilia{
			"02": {maquinaria: 1, alguna: 1},
			"03": {maquinaria: 3, anuncio: 3, alguna: 3},
			"04": {maquinaria: 3, anuncio: 3, alguna: 3},
			"05": {maquinaria: 3, anuncio: 1, alguna: 3},
			"06": {maquinaria: 3, anuncio: 2, alguna: 3},
			"07": {maquinaria: 3, anuncio: 2, alguna: 3},
			"08": {maquinaria: 2, anuncio: 1, alguna: 2},
			"09": {maquinaria: 2, anuncio: 1, alguna: 2},
			"13": {maquinaria: 3, anuncio: 2, alguna: 3},
			"14": {maquinaria: 3, anuncio: 2, alguna: 3},
			"15": {maquinaria: 3, anuncio: 3, alguna: 3},
			"16": {maquinaria: 2, anuncio: 2, alguna: 2},
			"17": {maquinaria: 3, anuncio: 2, alguna: 3},
			"19": {otraConversacion: 1, alguna: 1},
		},
	},
	{
		ruta: informeDeH72,
		calibrado: map[string]marcadasPorFamilia{
			"03": {maquinaria: 1, anuncio: 1, alguna: 1},
			"06": {maquinaria: 1, anuncio: 1, alguna: 1},
			"13": {maquinaria: 2, anuncio: 2, alguna: 2},
			"14": {maquinaria: 2, anuncio: 2, alguna: 2},
			"15": {maquinaria: 3, anuncio: 3, alguna: 3},
			"19": {maquinaria: 1, alguna: 1},
		},
	},
}

// probarExpresionesCalibradas es la subprueba expresiones-calibradas de
// TestEvalsDelRepositorio (contracts/lista-de-expresiones.md §4 y §6; FR-021,
// FR-095, SC-003, US1-5): aplicada con ExtraerExpresionesProhibidas, la
// comparación de H7.2 FR 051, a las 93 respuestas de cada informe de
// informesCalibrados, la lista marca, por las dos cifras del fichero de la eval,
// en cada familia y en la lista entera, exactamente las del reparto calibrado, y
// ninguna de las demás evals. Cada diferencia nombra el informe, la eval, la
// columna, lo contado y lo calibrado. Las evals se nombran por sus dos cifras y
// nunca por su nombre: el de una eval retirada no se escribe en ningún test
// (FR-020 de H7.2).
func probarExpresionesCalibradas(t *testing.T, lista ExpresionesProhibidas) {
	t.Helper()

	var distintas []string

	for _, informe := range informesCalibrados {
		marcadas := marcadasEnElInforme(t, informe.ruta, lista)
		distintas = append(distintas, distintasDelCalibrado(informe, marcadas)...)
	}

	assert.Empty(t, distintas, "respuestas de los informes calibrados que la lista de %s marca con otro reparto "+
		"que el calibrado:\n%s", evalsDelRepositorio, strings.Join(distintas, "\n"))
}

// marcadasEnElInforme es el reparto de las respuestas del informe versionado de
// la ruta que marca la lista, por las dos cifras del fichero de la eval: por
// cada familia, las que llevan alguna expresión suya, y las que llevan alguna de
// la lista entera. El informe tiene sus 93 respuestas.
func marcadasEnElInforme(t *testing.T, ruta string, lista ExpresionesProhibidas) map[string]marcadasPorFamilia {
	t.Helper()

	// De cada sesión del informe, lo que la calibración necesita.
	var informe struct {
		Evals []struct {
			Eval      string `json:"eval"`
			Respuesta string `json:"respuesta"`
		} `json:"evals"`
	}
	require.NoError(t, json.Unmarshal(contenidoDelFichero(t, ruta), &informe), "%s es un informe del job de evals", ruta)
	require.Len(t, informe.Evals, respuestasDeCadaInforme, "el informe %s tiene sus respuestas", ruta)

	maquinaria := ExpresionesProhibidas{Maquinaria: lista.Maquinaria}
	otraConversacion := ExpresionesProhibidas{OtraConversacion: lista.OtraConversacion}
	anuncio := ExpresionesProhibidas{Anuncio: lista.Anuncio}
	marcadas := map[string]marcadasPorFamilia{}

	for _, sesion := range informe.Evals {
		require.Regexp(t, `^[0-9]{2}-`, sesion.Eval, "el fichero de cada eval de %s empieza por sus dos cifras", ruta)
		numero := sesion.Eval[:2]

		reparto := marcadas[numero]
		reparto.maquinaria += marcadaPor(sesion.Respuesta, maquinaria)
		reparto.otraConversacion += marcadaPor(sesion.Respuesta, otraConversacion)
		reparto.anuncio += marcadaPor(sesion.Respuesta, anuncio)
		reparto.alguna += marcadaPor(sesion.Respuesta, lista)
		marcadas[numero] = reparto
	}

	return marcadas
}

// marcadaPor es 1 si la respuesta lleva alguna expresión de la lista, y 0 si no.
func marcadaPor(respuesta string, lista ExpresionesProhibidas) int {
	if len(ExtraerExpresionesProhibidas(respuesta, lista)) > 0 {
		return 1
	}

	return 0
}

// distintasDelCalibrado da una línea por cada eval y columna en que el reparto
// marcado en el informe no es el calibrado, con el informe, la eval, la columna,
// lo contado y lo calibrado, en orden de eval y de columna; o nil si no hay
// ninguna.
func distintasDelCalibrado(informe informeCalibrado, marcadas map[string]marcadasPorFamilia) []string {
	numeros := slices.Concat(slices.Collect(maps.Keys(marcadas)), slices.Collect(maps.Keys(informe.calibrado)))
	slices.Sort(numeros)

	var distintas []string

	for _, numero := range slices.Compact(numeros) {
		contadas, calibradas := marcadas[numero].cuentas(), informe.calibrado[numero].cuentas()

		for posicion, columna := range columnasDelCalibrado {
			if contadas[posicion] != calibradas[posicion] {
				distintas = append(distintas, fmt.Sprintf("%s, eval %s, columna %s: marca %d respuestas, y las "+
					"calibradas son %d", informe.ruta, numero, columna, contadas[posicion], calibradas[posicion]))
			}
		}
	}

	return distintas
}

// textoAMirar es un texto que no puede llevar ninguna expresión prohibida, con
// lo que lo nombra en un fallo.
type textoAMirar struct {
	nombre string
	texto  string
}

// expresionesEn da una línea por cada texto que lleva alguna expresión de la
// lista, con su nombre y las expresiones que lleva, en el orden de los textos; o
// nil si ninguno lleva ninguna.
func expresionesEn(textos []textoAMirar, lista ExpresionesProhibidas) []string {
	var lineas []string

	for _, texto := range textos {
		if encontradas := ExtraerExpresionesProhibidas(texto.texto, lista); len(encontradas) > 0 {
			lineas = append(lineas, texto.nombre+": "+strings.Join(encontradas, ", "))
		}
	}

	return lineas
}

// probarExpresionesEnLosBloques es la subprueba expresiones-en-los-bloques de
// TestEvalsDelRepositorio (contrato lista-y-juicio §6; research D10; FR-085,
// SC-004): ninguna expresión de la lista casa con el texto que da boe articulo
// --json, lo que una respuesta puede transcribir, de cada bloque que leen las
// evals —los de sus consultas necesarias, sobre UnionDeGrabaciones— ni de cada
// bloque de los comandos de cada grafo previo, sobre las mismas grabaciones con
// las del grafo previo encima. Cada lectura es en proceso, con su caché temporal.
func probarExpresionesEnLosBloques(t *testing.T, evals []Eval, lista ExpresionesProhibidas) {
	t.Helper()

	var leidos []textoAMirar

	grabadas := registroDeLaReproduccion(t, copiaDeLaUnionDeGrabaciones(t))

	for _, consulta := range ConsultasNecesarias(evals) {
		if consulta.Applet == appletDeLasNormas && consulta.Verbo == verboArticulo {
			leidos = append(leidos, bloqueLeido(t, grabadas, consulta))
		}
	}

	for _, eval := range evals {
		if eval.GrafoPrevio.Grabaciones == "" {
			continue
		}

		reproduccion := copiaDeLaUnionDeGrabaciones(t)
		require.NoError(t, copiarGrabaciones(filepath.Join(GrafosPrevios, eval.GrafoPrevio.Grabaciones), reproduccion))
		derivadas := registroDeLaReproduccion(t, reproduccion)

		for _, comando := range eval.GrafoPrevio.Comandos {
			leido := bloqueLeido(t, derivadas, Consulta{
				Applet: comando.Applet, Verbo: verboArticulo, Argumentos: []string{comando.Norma, comando.Bloque},
			})
			leido.nombre += " con las grabaciones del grafo previo " + eval.GrafoPrevio.Grabaciones
			leidos = append(leidos, leido)
		}
	}

	require.NotEmpty(t, leidos, "las evals de %s, o el grafo previo de alguna, leen algún bloque", evalsDelRepositorio)

	conExpresiones := expresionesEn(leidos, lista)
	assert.Empty(t, conExpresiones, "expresiones prohibidas en el texto de los %d bloques leídos para las evals de %s:\n%s",
		len(leidos), evalsDelRepositorio, strings.Join(conExpresiones, "\n"))
}

// bloqueLeido es el texto del bloque de la consulta que da boe articulo --json
// con el registro, nombrado por su orden. La lectura termina en 0 y el bloque
// tiene texto: uno vacío no dejaría nada que mirar.
func bloqueLeido(t *testing.T, registro *app.Registro, consulta Consulta) textoAMirar {
	t.Helper()

	// Sin entrega al grafo, la salida de error no dice nada que mirar aquí.
	leido, _ := leerBloque(t, registro, consulta)
	require.NotEmpty(t, leido.Texto, "«%s» da el texto del bloque", ordenDe(consulta))

	return textoAMirar{nombre: ordenDe(consulta), texto: leido.Texto}
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

// probarExpresionesDeLaSkill es la subprueba expresiones-de-la-skill de
// TestEvalsDelRepositorio (contrato lista-y-juicio §6; research D11; FR-043,
// FR-051): ninguna expresión de la lista casa con lo que la skill enseña a
// escribir en la respuesta —la forma escrita de cada etiqueta de aviso y de
// hallazgo, y el contenido de cada bloque de código text del SKILL.md de
// boe-legislacion, entre ellos el de la línea de version-obsoleta—.
func probarExpresionesDeLaSkill(t *testing.T, lista ExpresionesProhibidas) {
	t.Helper()

	etiquetasDeAviso := boe.EtiquetasDeAviso()
	etiquetasDeHallazgo := grafo.EtiquetasDeHallazgo()
	bloques := bloquesDeTexto(string(contenidoDelFichero(t, skillDelRepositorio)))

	require.NotEmpty(t, etiquetasDeAviso, "el binario etiqueta algún aviso")
	require.NotEmpty(t, etiquetasDeHallazgo, "el binario etiqueta alguna clase de hallazgo")
	require.True(t, slices.ContainsFunc(bloques, func(bloque string) bool {
		return slices.Contains(ExtraerHallazgos(bloque), string(grafo.ClaseVersionObsoleta))
	}), "%s tiene bloques de código text, entre ellos el de la línea %s", skillDelRepositorio,
		formaEscrita(etiquetasDeHallazgo[grafo.ClaseVersionObsoleta]))

	var textos []textoAMirar

	for _, codigo := range slices.Sorted(maps.Keys(etiquetasDeAviso)) {
		textos = append(textos, textoAMirar{
			nombre: "la forma del aviso " + codigo, texto: formaEscrita(etiquetasDeAviso[codigo]),
		})
	}

	for _, clase := range slices.Sorted(maps.Keys(etiquetasDeHallazgo)) {
		textos = append(textos, textoAMirar{
			nombre: "la forma del hallazgo " + string(clase), texto: formaEscrita(etiquetasDeHallazgo[clase]),
		})
	}

	for posicion, bloque := range bloques {
		textos = append(textos, textoAMirar{
			nombre: fmt.Sprintf("el bloque text %d de %s", posicion+1, skillDelRepositorio), texto: bloque,
		})
	}

	conExpresiones := expresionesEn(textos, lista)
	assert.Empty(t, conExpresiones, "expresiones prohibidas en lo que la skill enseña a escribir:\n%s",
		strings.Join(conExpresiones, "\n"))
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
// vacío.
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
	// cambia por un espacio para que no junte las palabras de sus lados.
	codigoEnLinea = regexp.MustCompile("`[^`]*`")

	// fechaConCifras casa con una fecha AAAAMMDD escrita con cifras: ocho
	// cifras, sin otra delante ni detrás, con un mes de 01 a 12 y un día de 01 a
	// 31 (FR-013). El grupo es la fecha.
	fechaConCifras = regexp.MustCompile(`(?:^|[^0-9])([0-9]{4}(?:0[1-9]|1[0-2])(?:0[1-9]|[12][0-9]|3[01]))(?:$|[^0-9])`)
)

// defectosDeLaProsa da una línea por cada párrafo de la prosa del Markdown
// (parrafosDeLaProsa) que lleva alguna expresión de la lista, con el número de
// su primera línea y las expresiones, en el orden del fichero, y detrás una por
// cada línea del fichero entero, sin quitar nada, que lleva una fecha AAAAMMDD
// escrita con cifras, con su número y sus fechas; o nil si no hay ninguna.
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
// sangría, y cada tramo de código en línea cambiado por un espacio. Como en
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
			parrafos = append(parrafos, textoAMirar{
				nombre: fmt.Sprintf("párrafo de la línea %d", primera),
				texto:  codigoEnLinea.ReplaceAllString(strings.Join(lineas, " "), " "),
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
func TestProsaDeLaSkill(t *testing.T) {
	t.Parallel()

	lista := ExpresionesProhibidas{
		Maquinaria:       []string{"memoria de consultas", "hallazgos", "graph check"},
		OtraConversacion: []string{"te dije"},
		Anuncio:          []string{"redacto la respuesta"},
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
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.defectos, defectosDeLaProsa(caso.markdown, lista))
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
// como la superada y la del bloque leído como la reciente.
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

	var clases []string

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
		}
	}

	slices.Sort(clases)
	assert.Equal(t, slices.Sorted(slices.Values(eval.Hallazgos)), slices.Compact(clases),
		"%s: graph check da en la sesión exactamente las clases de hallazgo que la eval espera", ruta)
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
