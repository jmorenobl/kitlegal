package evals

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
}

// TestLeerConjunto fija la lectura de un directorio de evals del contrato
// evals-y-grabaciones §1: todas sus entradas se leen antes de devolver, las
// evals bien formadas van a Evals y todo lo demás a MalFormados, en orden de
// nombre y cada una con un error que empieza por su nombre, sin que ninguna se
// salte ni sea el error de la lectura, que queda para el directorio que no se
// puede listar.
func TestLeerConjunto(t *testing.T) {
	t.Parallel()

	leidaDelArticulo21 := Eval{
		Fichero:  nombreDeEval,
		Pregunta: "¿qué dice el art. 21 de la Ley 39/2015?",
		Activa:   true,
		Comandos: []ComandoEsperado{{Applet: "boe", Norma: "BOE-A-2015-10565", Bloque: "a21"}},
		Citas:    []CitaEsperada{{Norma: "BOE-A-2015-10565", Bloque: "a21"}},
	}
	pregunta := "¿Cómo invierto una lista enlazada en Go?"
	sinLaForma := "no tiene la forma <nn>-<descripción>.yaml"

	casos := []struct {
		nombre      string
		entradas    []entradaDeConjunto
		evals       []Eval
		malFormados []malFormadoEsperado
	}{
		{
			nombre: "bien-formadas",
			entradas: []entradaDeConjunto{
				{nombre: "02-no-activa-programacion.yaml", contenido: contenidoDeProgramacion},
				{nombre: nombreDeEval, contenido: contenidoDelArticulo21},
			},
			evals: []Eval{leidaDelArticulo21, {Fichero: "02-no-activa-programacion.yaml", Pregunta: pregunta}},
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
			evals: []Eval{{Fichero: "02-no-activa-programacion.yaml", Pregunta: pregunta}},
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
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			conjunto, err := LeerConjunto(crearConjunto(t, caso.entradas))
			require.NoError(t, err, "un fichero mal formado nunca es el error de la lectura")
			assert.Equal(t, caso.evals, conjunto.Evals)

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
// TestConjuntoDeEvals llama a ComprobarConjuntoDeBoeLegislacion.
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

// TestConjuntoDeEvals fija ComprobarConjuntoDeBoeLegislacion (contrato de evals
// §2): el conjunto que cumple todas las reglas de data-model §6.3 no da ningún
// defecto; cada copia que incumple solo una regla da exactamente un defecto, con
// el nombre de esa regla y un mensaje que nombra los ficheros y las normas
// implicados; y un conjunto que las incumple todas da un defecto por regla, en
// el orden de la tabla.
func TestConjuntoDeEvals(t *testing.T) {
	t.Parallel()

	todos := []string{
		"01-lpac-articulo-21.yaml", "02-lcsp-contrato-menor.yaml", "03-lrbrl-atribuciones-del-pleno.yaml",
		"04-lgt-prescripcion.yaml", "05-trlrhl-impuestos-municipales.yaml", "06-irpf-rendimientos-del-trabajo.yaml",
		"07-lrjsp-principio-de-legalidad.yaml", "08-ltaibg-plazo-de-resolucion.yaml",
		"09-constitucion-articulo-140.yaml", "10-et-vacaciones.yaml",
		"11-no-activa-programacion.yaml", "12-no-activa-acuerdo-entre-amigos.yaml",
	}
	positivas := todos[:10]
	deMas := []string{
		"13-no-activa-13.yaml", "14-no-activa-14.yaml", "15-no-activa-15.yaml", "16-no-activa-16.yaml",
		"17-no-activa-17.yaml", "18-no-activa-18.yaml", "19-no-activa-19.yaml", "20-no-activa-20.yaml",
		"21-no-activa-21.yaml",
	}

	t.Run("cumple-todas", func(t *testing.T) {
		t.Parallel()

		conjunto := conjuntoQueCumple(t)
		require.Equal(t, todos, ficherosDelConjunto(conjunto), "el conjunto sintético es el del contrato")

		assert.Empty(t, ComprobarConjuntoDeBoeLegislacion(conjunto.evals, conjunto.normas))
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

			defectos := ComprobarConjuntoDeBoeLegislacion(conjunto.evals, conjunto.normas)
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

		reglas := []string{}
		for _, defecto := range ComprobarConjuntoDeBoeLegislacion(conjunto.evals, conjunto.normas) {
			reglas = append(reglas, defecto.Regla)
		}

		assert.Equal(t, []string{
			"tamaño", "positivas", "no activación", "materias distintas", "normas del hito",
			"art. 21", "fiscal", "boe-fiscal", "normas conocidas",
		}, reglas)
	})
}

// ficherosDelConjunto son los ficheros de las evals del conjunto, en su orden.
func ficherosDelConjunto(conjunto conjuntoSintetico) []string {
	ficheros := make([]string, 0, len(conjunto.evals))
	for _, eval := range conjunto.evals {
		ficheros = append(ficheros, eval.Fichero)
	}

	return ficheros
}
