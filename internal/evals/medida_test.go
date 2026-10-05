package evals

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

// medidaSintetica es una medida del juez con las diez claves que se leen, cada
// una con un valor distinto, y con las que no se leen: skill, votos, origen y
// el fichero de cada huella (contracts/medida-del-juez.md §1 de H24).
const medidaSintetica = `{
  "skill": "una-skill",
  "clase": "una_clase",
  "fecha": "2026-01-02",
  "modelo_del_juez": "claude-juez-1-2",
  "version_de_claude_code": "1.2.3",
  "rubrica": {"fichero": "rubrica.md", "sha256": "huella-de-la-rubrica"},
  "casos": {"fichero": "casos.yaml", "sha256": "huella-de-los-casos"},
  "defectos": {"casos": 5, "sin_marcar": 1},
  "correctos": {"casos": 4, "marcados": 2},
  "votos": ["votos.jsonl"],
  "origen": "una medida sintética"
}
`

// clavesLeidasDeLaMedida son las diez claves de la medida que se leen, en el
// orden de contracts/medida-del-juez.md §1 de H24, con la de un objeto unida a
// la suya por un punto.
var clavesLeidasDeLaMedida = []string{
	"clase", "fecha", "modelo_del_juez", "version_de_claude_code", "rubrica.sha256", "casos.sha256",
	"defectos.casos", "defectos.sin_marcar", "correctos.casos", "correctos.marcados",
}

// Los principios de los errores de la lectura de una medida: el de la que no
// tiene alguna de las claves que se leen, que sigue con esas claves, y el de la
// que no es un documento JSON con su forma, que sigue con el error de quien la
// lee.
const (
	faltanClavesDeLaMedida = "claves que faltan: "
	medidaQueNoEsJSON      = "no se puede leer como JSON: "
)

// medidaSinCorresponder es el principio de la línea de la comprobación de una
// medida que no se puede leer o que no tiene alguna de sus claves: sigue con el
// error de su lectura.
const medidaSinCorresponder = "la medida versionada (juez/medida.json) no corresponde: "

// claseQueDecideEnElRepositorio es la clase que decide del juez de la skill de
// la primera pareja de la tabla de las copias, boe-legislacion, que es la de su
// medida (FR-032 de H24).
const claseQueDecideEnElRepositorio = "afirma_lo_no_leido"

// TestLeerMedidaDelJuez fija la lectura de la medida del juez
// (contracts/medida-del-juez.md §1 de H24; FR-040): se leen diez claves, cada
// una a su campo, y las demás no se miran; y la que no tiene alguna de las diez,
// o la tiene con null, o no es un documento JSON con esa forma, no se lee, con
// un error que dice qué claves faltan o por qué no es ese documento. Una clave
// repetida es un error, nunca la última que gana.
func TestLeerMedidaDelJuez(t *testing.T) {
	t.Parallel()

	t.Run("entera", func(t *testing.T) {
		t.Parallel()

		medida, err := leerMedidaDelJuez(escribirMedida(t, []byte(medidaSintetica)))
		require.NoError(t, err)

		assert.Equal(t, MedidaDelJuez{
			Clase:               "una_clase",
			Fecha:               "2026-01-02",
			ModeloDelJuez:       "claude-juez-1-2",
			VersionDeClaudeCode: "1.2.3",
			HuellaDeLaRubrica:   "huella-de-la-rubrica",
			HuellaDeLosCasos:    "huella-de-los-casos",
			Defectos:            DefectosDeLaMedida{Casos: 5, SinMarcar: 1},
			Correctos:           CorrectosDeLaMedida{Casos: 4, Marcados: 2},
		}, medida)
	})

	for _, caso := range medidasSinClaves(t) {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			_, err := leerMedidaDelJuez(escribirMedida(t, caso.contenido))
			require.EqualError(t, err, faltanClavesDeLaMedida+strings.Join(caso.faltan, ", "))
		})
	}

	for _, caso := range medidasQueNoSonJSON(t) {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			_, err := leerMedidaDelJuez(escribirMedida(t, caso.contenido))
			require.Error(t, err)
			assert.True(t, strings.HasPrefix(err.Error(), medidaQueNoEsJSON), err.Error())
		})
	}

	t.Run("sin-fichero", func(t *testing.T) {
		t.Parallel()

		_, err := leerMedidaDelJuez(filepath.Join(t.TempDir(), ficheroDeMedidaDelJuez))
		require.ErrorIs(t, err, fs.ErrNotExist)
		assert.True(t, strings.HasPrefix(err.Error(), "no se puede leer: "), err.Error())
	})
}

// medidaIlegible es un caso de TestLeerMedidaDelJuez: el contenido de una
// medida que no se lee y, si es por las claves que no tiene, esas claves.
type medidaIlegible struct {
	nombre    string
	contenido []byte
	faltan    []string
}

// medidasSinClaves son las medidas a las que les falta alguna de las claves que
// se leen: la sintética sin cada una de las diez; con una a null, que es no
// tenerla; sin un objeto entero, que es no tener ninguna de las suyas; y la
// vacía, a la que le faltan las diez, en su orden.
func medidasSinClaves(t *testing.T) []medidaIlegible {
	t.Helper()

	sintetica := []byte(medidaSintetica)

	casos := make([]medidaIlegible, 0, len(clavesLeidasDeLaMedida)+3)
	for _, clave := range clavesLeidasDeLaMedida {
		casos = append(casos, medidaIlegible{
			nombre:    "sin-" + clave,
			contenido: medidaCambiada(t, sintetica, clave, quitarLaClave),
			faltan:    []string{clave},
		})
	}

	return append(casos,
		medidaIlegible{
			nombre:    "con-null",
			contenido: medidaCambiada(t, sintetica, "version_de_claude_code", ponerEnLaClave(nil)),
			faltan:    []string{"version_de_claude_code"},
		},
		medidaIlegible{
			nombre:    "sin-un-objeto",
			contenido: medidaCambiada(t, sintetica, "defectos", quitarLaClave),
			faltan:    []string{"defectos.casos", "defectos.sin_marcar"},
		},
		medidaIlegible{nombre: "vacia", contenido: []byte("{}\n"), faltan: clavesLeidasDeLaMedida},
	)
}

// medidasQueNoSonJSON son las medidas que no son un documento JSON con la forma
// de una medida: la que no es JSON, la que no es un objeto, la que tiene un
// recuento que no es un entero —un texto o un número con decimales—, la que
// tiene un objeto donde va un texto y la que repite una clave.
func medidasQueNoSonJSON(t *testing.T) []medidaIlegible {
	t.Helper()

	sintetica := []byte(medidaSintetica)

	const clase = `"clase": "una_clase",`

	require.Contains(t, medidaSintetica, clase, "premisa: la medida sintética tiene la clave que el caso repite")

	return []medidaIlegible{
		{nombre: "no-es-json", contenido: []byte("{\n")},
		{nombre: "no-es-un-objeto", contenido: []byte("[]\n")},
		{
			nombre:    "recuento-que-es-un-texto",
			contenido: medidaCambiada(t, sintetica, "defectos.sin_marcar", ponerEnLaClave("0")),
		},
		{
			nombre:    "recuento-con-decimales",
			contenido: medidaCambiada(t, sintetica, "correctos.marcados", ponerEnLaClave(0.5)),
		},
		{
			nombre:    "objeto-que-no-es-un-texto",
			contenido: medidaCambiada(t, sintetica, "modelo_del_juez", ponerEnLaClave(map[string]any{})),
		},
		{
			nombre:    "clave-repetida",
			contenido: []byte(strings.Replace(medidaSintetica, clase, clase+` "clase": "otra_clase",`, 1)),
		},
	}
}

// escribirMedida escribe ese contenido como la medida de un directorio temporal
// del test y devuelve su ruta.
func escribirMedida(t *testing.T, contenido []byte) string {
	t.Helper()

	ruta := filepath.Join(t.TempDir(), ficheroDeMedidaDelJuez)
	require.NoError(t, os.WriteFile(ruta, contenido, 0o600))

	return ruta
}

// medidaCambiada es la medida de ese contenido con el cambio hecho en la clave
// de esa ruta, que tiene que estar: las claves de la ruta van separadas por un
// punto, como en rubrica.sha256, y todas menos la última son objetos.
func medidaCambiada(t *testing.T, contenido []byte, ruta string, cambio func(objeto map[string]any, clave string)) []byte {
	t.Helper()

	var medida map[string]any

	require.NoError(t, json.Unmarshal(contenido, &medida), "premisa: la medida que se cambia es un objeto JSON")

	objeto := medida
	claves := strings.Split(ruta, ".")

	for _, clave := range claves[:len(claves)-1] {
		anidado, esObjeto := objeto[clave].(map[string]any)
		require.True(t, esObjeto, "premisa: %s es un objeto de la medida que se cambia", clave)

		objeto = anidado
	}

	ultima := claves[len(claves)-1]
	require.Contains(t, objeto, ultima, "premisa: la medida que se cambia tiene la clave %s", ruta)

	cambio(objeto, ultima)

	cambiada, err := json.Marshal(medida)
	require.NoError(t, err)

	return cambiada
}

// quitarLaClave es el cambio de medidaCambiada que deja la medida sin la clave.
func quitarLaClave(objeto map[string]any, clave string) {
	delete(objeto, clave)
}

// ponerEnLaClave da el cambio de medidaCambiada que pone ese valor en la clave.
func ponerEnLaClave(valor any) func(objeto map[string]any, clave string) {
	return func(objeto map[string]any, clave string) { objeto[clave] = valor }
}

// TestMedidaVersionada es el control de umbral de FR-042, FR-105 y SC-005 de
// H24, y el de los dos umbrales medida_del_juez:afirma_lo_no_leido:… en make ci
// (contracts/medida-del-juez.md §2 y §8 de H24; FR-041): la medida versionada
// del juez de cada skill que lo tiene corresponde —su rúbrica, sus casos, su
// modelo y su versión de Claude Code son los de lo que hay— y se cumple, con sus
// dos recuentos a 0, con el modelo del juez y la versión de Claude Code de sus
// votos que fija la definición del job.
//
// Sobre una copia de la carpeta del juez en t.TempDir(), que antes no daba
// ninguna línea, lo ve fallar: con la rúbrica, los casos, el modelo fijado o la
// versión fijada cambiados, uno cada vez, y con cada uno de los dos recuentos
// distinto de 0, la única línea es la que lo nombra, seis de seis. Y ve lo demás
// que la comprobación dice: la medida de una clase que no es la que decide; todo
// a la vez, una línea por lo que falla y en el orden del contrato; la medida sin
// una clave y la que no es JSON, que no corresponden; y lo que deja de poder
// leerse con el juez ya leído. Las copias del repositorio solo se leen (FR-045).
func TestMedidaVersionada(t *testing.T) {
	t.Parallel()

	modelo, version := fijadosParaElJuez(t)

	t.Run("del-repositorio", func(t *testing.T) {
		t.Parallel()

		for _, pareja := range copiasDelJuez {
			evals := filepath.Join(raizDelRepositorio, filepath.FromSlash(path.Dir(pareja.copias)))

			lineas := comprobarLaMedida(juezDe(t, evals), modelo, version)
			assert.Empty(t, lineas, "la medida versionada de %s, con el modelo %s y la versión %s de Claude Code que fija %s:\n%s",
				pareja.copias, modelo, version, rutaDeLaDefinicionDelJob, strings.Join(lineas, "\n"))
		}
	})

	mutaciones := mutacionesDeLaMedida(modelo, version)
	require.Len(t, mutaciones, 6, "las seis mutaciones de SC-005 de H24")

	for _, mutacion := range slices.Concat(mutaciones, otrasMedidasSinCorresponder(modelo, version)) {
		t.Run(mutacion.nombre, func(t *testing.T) {
			t.Parallel()

			probarMutacionDeLaMedida(t, modelo, version, mutacion)
		})
	}

	t.Run("clases-que-deciden", func(t *testing.T) {
		t.Parallel()

		probarLasClasesQueDeciden(t, modelo, version)
	})

	for _, fichero := range []string{ficheroDeMedidaDelJuez, ficheroDeCasosDelJuez} {
		t.Run("sin-"+fichero+"-tras-leer-el-juez", func(t *testing.T) {
			t.Parallel()

			probarLoQueDejaDePoderLeerse(t, modelo, version, fichero)
		})
	}
}

// fijadosParaElJuez son el id del modelo del juez y la versión de Claude Code de
// sus votos que fija la definición del job del repositorio. Tiene que fijar los
// dos: sin ellos, compararlos con los de una medida que tampoco los tuviera
// pasaría en vacío.
func fijadosParaElJuez(t *testing.T) (modelo, version string) {
	t.Helper()

	delJob, err := leerDefinicionDelJob(rutaDeLaDefinicionDelJob)
	require.NoError(t, err)

	require.NotEmpty(t, delJob.ModeloDelJuez, "jobs.evals.env.%s de %s fija el modelo del juez",
		variableDelModeloDelJuez, rutaDeLaDefinicionDelJob)
	require.NotEmpty(t, delJob.VersionDelJuez, "jobs.evals.env.%s de %s fija la versión de Claude Code de los votos del juez",
		variableDeLaVersionDelJuez, rutaDeLaDefinicionDelJob)

	return delJob.ModeloDelJuez, delJob.VersionDelJuez
}

// cambioDeLaCarpetaDelJuez es un cambio que un caso de TestMedidaVersionada
// hace en la copia de la carpeta del juez de esa ruta.
type cambioDeLaCarpetaDelJuez func(t *testing.T, carpeta string)

// mutacionDeLaMedida es un caso de TestMedidaVersionada: lo que cambia en la
// copia de la carpeta del juez antes de leer el juez de ella, el modelo y la
// versión fijados con los que se comprueba su medida, y las líneas que la
// comprobación da, en su orden.
type mutacionDeLaMedida struct {
	nombre  string
	cambios []cambioDeLaCarpetaDelJuez
	modelo  string
	version string
	lineas  []string

	// soloElPrincipio dice que de la única línea solo se fija cómo empieza:
	// sigue con el error de quien lee la medida.
	soloElPrincipio bool
}

// Las líneas de la comprobación de la medida que no llevan ningún valor
// (contracts/medida-del-juez.md §2 de H24), y los recuentos que los casos ponen
// en la medida de la copia para darlas.
const (
	lineaDeLaRubrica = "la rúbrica (juez/rubrica.md) no es la de la medida versionada"
	lineaDeLosCasos  = "los casos (juez/casos.yaml) no son los de la medida versionada"

	defectosSinMarcarDeLaMutacion = 1
	lineaDeLosDefectosSinMarcar   = "la medida versionada no se cumple: 1 defectos sin marcar"

	correctosMarcadosDeLaMutacion = 2
	lineaDeLosCorrectosMarcados   = "la medida versionada no se cumple: 2 correctos marcados"
)

// otroFijado es un valor fijado para el juez distinto del de la medida.
func otroFijado(deLaMedida string) string {
	return deLaMedida + "-cambiado"
}

// lineaDelModelo y lineaDeLaVersion son las líneas de la comprobación de una
// medida de ese modelo, o de esa versión de Claude Code, con otro fijado.
func lineaDelModelo(deLaMedida, fijado string) string {
	return "la medida versionada es del modelo " + deLaMedida + " y el fijado para el juez es " + fijado
}

func lineaDeLaVersion(deLaMedida, fijada string) string {
	return "la medida versionada es de la versión " + deLaMedida + " de Claude Code y la fijada para los votos del juez es " + fijada
}

// mutacionesDeLaMedida son las seis de SC-005 de H24, con el modelo y la versión
// que fija la definición del job, que son los de la medida versionada: cada una
// de sus cuatro claves cambiada —la rúbrica y los casos, en la copia; el modelo
// y la versión, en lo fijado— y cada uno de sus dos recuentos distinto de 0,
// con la línea que la nombra.
func mutacionesDeLaMedida(modelo, version string) []mutacionDeLaMedida {
	return []mutacionDeLaMedida{
		{
			nombre:  "rubrica-cambiada",
			cambios: []cambioDeLaCarpetaDelJuez{cambiarUnByteDe(ficheroDeRubricaDelJuez)},
			modelo:  modelo, version: version,
			lineas: []string{lineaDeLaRubrica},
		},
		{
			nombre:  "casos-cambiados",
			cambios: []cambioDeLaCarpetaDelJuez{cambiarUnByteDe(ficheroDeCasosDelJuez)},
			modelo:  modelo, version: version,
			lineas: []string{lineaDeLosCasos},
		},
		{
			nombre: "modelo-fijado-cambiado",
			modelo: otroFijado(modelo), version: version,
			lineas: []string{lineaDelModelo(modelo, otroFijado(modelo))},
		},
		{
			nombre: "version-fijada-cambiada",
			modelo: modelo, version: otroFijado(version),
			lineas: []string{lineaDeLaVersion(version, otroFijado(version))},
		},
		{
			nombre: "defectos-sin-marcar",
			cambios: []cambioDeLaCarpetaDelJuez{
				cambiarLaMedidaEn("defectos.sin_marcar", ponerEnLaClave(defectosSinMarcarDeLaMutacion)),
			},
			modelo: modelo, version: version,
			lineas: []string{lineaDeLosDefectosSinMarcar},
		},
		{
			nombre: "correctos-marcados",
			cambios: []cambioDeLaCarpetaDelJuez{
				cambiarLaMedidaEn("correctos.marcados", ponerEnLaClave(correctosMarcadosDeLaMutacion)),
			},
			modelo: modelo, version: version,
			lineas: []string{lineaDeLosCorrectosMarcados},
		},
	}
}

// otrasMedidasSinCorresponder son los casos de TestMedidaVersionada que no son
// una de las seis mutaciones: la medida de otra clase, que da la línea de la
// clase que decide; todo cambiado a la vez, que da una línea por lo que falla
// en el orden de contracts/medida-del-juez.md §2 de H24; la medida sin una clave
// que ninguna comparación usa, que no corresponde; y la que no es JSON, que
// tampoco.
func otrasMedidasSinCorresponder(modelo, version string) []mutacionDeLaMedida {
	const (
		otraClase          = "otra_clase"
		lineaDeLaOtraClase = "la clase " + claseQueDecideEnElRepositorio + " decide y la medida versionada es de " + otraClase
	)

	return []mutacionDeLaMedida{
		{
			nombre:  "medida-de-otra-clase",
			cambios: []cambioDeLaCarpetaDelJuez{cambiarLaMedidaEn("clase", ponerEnLaClave(otraClase))},
			modelo:  modelo, version: version,
			lineas: []string{lineaDeLaOtraClase},
		},
		{
			nombre: "todo-a-la-vez",
			cambios: []cambioDeLaCarpetaDelJuez{
				cambiarLaMedidaEn("correctos.marcados", ponerEnLaClave(correctosMarcadosDeLaMutacion)),
				cambiarLaMedidaEn("defectos.sin_marcar", ponerEnLaClave(defectosSinMarcarDeLaMutacion)),
				cambiarLaMedidaEn("clase", ponerEnLaClave(otraClase)),
				cambiarUnByteDe(ficheroDeCasosDelJuez),
				cambiarUnByteDe(ficheroDeRubricaDelJuez),
			},
			modelo: otroFijado(modelo), version: otroFijado(version),
			lineas: []string{
				lineaDeLaRubrica,
				lineaDeLosCasos,
				lineaDelModelo(modelo, otroFijado(modelo)),
				lineaDeLaVersion(version, otroFijado(version)),
				lineaDeLaOtraClase,
				lineaDeLosDefectosSinMarcar,
				lineaDeLosCorrectosMarcados,
			},
		},
		{
			nombre:  "medida-sin-una-clave",
			cambios: []cambioDeLaCarpetaDelJuez{cambiarLaMedidaEn("fecha", quitarLaClave)},
			modelo:  modelo, version: version,
			lineas: []string{medidaSinCorresponder + faltanClavesDeLaMedida + "fecha"},
		},
		{
			nombre:  "medida-que-no-es-json",
			cambios: []cambioDeLaCarpetaDelJuez{escribirEnLaCarpeta(ficheroDeMedidaDelJuez, "{\n")},
			modelo:  modelo, version: version,
			lineas:          []string{medidaSinCorresponder + medidaQueNoEsJSON},
			soloElPrincipio: true,
		},
	}
}

// probarMutacionDeLaMedida copia la carpeta del juez, exige que la copia sin
// tocar no dé ninguna línea con el modelo y la versión que fija la definición
// del job, hace los cambios del caso, lee el juez de la copia y exige que la
// comprobación, con el modelo y la versión del caso, dé exactamente sus líneas.
func probarMutacionDeLaMedida(t *testing.T, modelo, version string, mutacion mutacionDeLaMedida) {
	t.Helper()

	evals := copiarLaCarpetaDelJuez(t)
	require.Empty(t, comprobarLaMedida(juezDe(t, evals), modelo, version),
		"premisa: la medida de la copia sin tocar corresponde y se cumple")

	for _, cambio := range mutacion.cambios {
		cambio(t, filepath.Join(evals, carpetaDelJuez))
	}

	lineas := comprobarLaMedida(juezDe(t, evals), mutacion.modelo, mutacion.version)

	if !mutacion.soloElPrincipio {
		assert.Equal(t, mutacion.lineas, lineas)

		return
	}

	require.Len(t, lineas, 1, strings.Join(lineas, "\n"))
	assert.True(t, strings.HasPrefix(lineas[0], mutacion.lineas[0]), lineas[0])
}

// probarLasClasesQueDeciden fija con qué clases se compara la de la medida
// (contracts/medida-del-juez.md §2 de H24; research D19 de H24): con cada una de
// las que deciden, en su orden, y con ninguna de las que solo se publican. Una
// segunda clase que decide no tiene medida y da su línea; sin ninguna que
// decida, la comparación no da ninguna.
func probarLasClasesQueDeciden(t *testing.T, modelo, version string) {
	t.Helper()

	juez := juezDe(t, copiarLaCarpetaDelJuez(t))
	require.Empty(t, comprobarLaMedida(juez, modelo, version),
		"premisa: la medida de la copia sin tocar corresponde y se cumple")

	juez.Clases = []ClaseDelJuez{
		{Nombre: "clase_1", Decide: true},
		{Nombre: claseQueDecideEnElRepositorio, Decide: true},
		{Nombre: "clase_3"},
		{Nombre: "clase_4", Decide: true},
	}

	assert.Equal(t, []string{
		"la clase clase_1 decide y la medida versionada es de " + claseQueDecideEnElRepositorio,
		"la clase clase_4 decide y la medida versionada es de " + claseQueDecideEnElRepositorio,
	}, comprobarLaMedida(juez, modelo, version))

	juez.Clases = []ClaseDelJuez{{Nombre: "clase_1"}, {Nombre: "clase_2"}}

	assert.Empty(t, comprobarLaMedida(juez, modelo, version))
}

// probarLoQueDejaDePoderLeerse quita de la copia de la carpeta del juez, con el
// juez ya leído de ella, el fichero de ese nombre, que la comprobación lee por
// su ruta —la medida o los casos—, y exige una sola línea, que empieza diciendo
// que no se puede leer y sigue con el error: ni se calla ni pasa por una huella
// que no coincide.
func probarLoQueDejaDePoderLeerse(t *testing.T, modelo, version, fichero string) {
	t.Helper()

	principios := map[string]string{
		ficheroDeMedidaDelJuez: medidaSinCorresponder + "no se puede leer: ",
		ficheroDeCasosDelJuez:  "los casos (juez/casos.yaml) no se pueden leer: ",
	}

	evals := copiarLaCarpetaDelJuez(t)
	juez := juezDe(t, evals)

	require.NoError(t, os.Remove(filepath.Join(evals, carpetaDelJuez, fichero)))

	lineas := comprobarLaMedida(juez, modelo, version)
	require.Len(t, lineas, 1, strings.Join(lineas, "\n"))
	assert.True(t, strings.HasPrefix(lineas[0], principios[fichero]), lineas[0])
}

// copiarLaCarpetaDelJuez copia en un directorio temporal del test, dentro de su
// carpeta juez, todos los ficheros de la carpeta del juez de la primera pareja
// de la tabla de las copias, y devuelve ese directorio, que es el de las evals
// de la copia. La carpeta del repositorio solo se lee (FR-045 de H24).
func copiarLaCarpetaDelJuez(t *testing.T) string {
	t.Helper()

	origen := filepath.Join(raizDelRepositorio, filepath.FromSlash(copiasDelJuez[0].copias))

	entradas, err := os.ReadDir(origen)
	require.NoError(t, err, "la carpeta del juez %s", origen)

	evals := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(evals, carpetaDelJuez), 0o750))

	for _, entrada := range entradas {
		contenido := contenidoDelFichero(t, filepath.Join(origen, entrada.Name()))
		require.NoError(t, os.WriteFile(filepath.Join(evals, carpetaDelJuez, entrada.Name()), contenido, 0o600))
	}

	return evals
}

// juezDe es el juez del directorio de evals de una skill, que tiene que tener
// la carpeta del juez bien formada.
func juezDe(t *testing.T, evals string) *Juez {
	t.Helper()

	conjunto, err := LeerConjunto(evals)
	require.NoError(t, err)

	malFormados := make([]string, 0, len(conjunto.MalFormados))
	for _, malFormado := range conjunto.MalFormados {
		malFormados = append(malFormados, malFormado.Error.Error())
	}

	require.NotNil(t, conjunto.Juez, "%s tiene la carpeta del juez bien formada:\n%s", evals, strings.Join(malFormados, "\n"))

	return conjunto.Juez
}

// cambiarUnByteDe da el cambio que cambia un byte, el último, del fichero de ese
// nombre de la copia de la carpeta del juez.
func cambiarUnByteDe(fichero string) cambioDeLaCarpetaDelJuez {
	return func(t *testing.T, carpeta string) {
		t.Helper()

		ruta := filepath.Join(carpeta, fichero)
		contenido := contenidoDelFichero(t, ruta)
		require.NotEmpty(t, contenido, "%s tiene algún byte que cambiar", ruta)

		contenido[len(contenido)-1] ^= 1
		require.NoError(t, os.WriteFile(ruta, contenido, 0o600))
	}
}

// cambiarLaMedidaEn da el cambio que hace, en la medida de la copia de la
// carpeta del juez, ese cambio de medidaCambiada en la clave de esa ruta.
func cambiarLaMedidaEn(ruta string, cambio func(objeto map[string]any, clave string)) cambioDeLaCarpetaDelJuez {
	return func(t *testing.T, carpeta string) {
		t.Helper()

		fichero := filepath.Join(carpeta, ficheroDeMedidaDelJuez)
		cambiada := medidaCambiada(t, contenidoDelFichero(t, fichero), ruta, cambio)
		require.NoError(t, os.WriteFile(fichero, cambiada, 0o600))
	}
}

// escribirEnLaCarpeta da el cambio que deja ese contenido en el fichero de ese
// nombre de la copia de la carpeta del juez.
func escribirEnLaCarpeta(fichero, contenido string) cambioDeLaCarpetaDelJuez {
	return func(t *testing.T, carpeta string) {
		t.Helper()

		require.NoError(t, os.WriteFile(filepath.Join(carpeta, fichero), []byte(contenido), 0o600))
	}
}

// parejaDeCarpetas es una fila de la tabla de TestCopiasDelJuez: la carpeta del
// juez de una skill y la carpeta de donde se copian sus cuatro ficheros, las dos
// relativas a la raíz del repositorio y con la barra de separador.
type parejaDeCarpetas struct {
	copias string
	origen string
}

// copiasDelJuez es la tabla de TestCopiasDelJuez: la carpeta del juez de cada
// skill que lo tiene, con la de la evidencia que la validó
// (contracts/medida-del-juez.md §3 de H24; FR-023). Una skill con juez que no
// esté aquí hace fallar el test.
var copiasDelJuez = []parejaDeCarpetas{
	{copias: "evals/boe-legislacion/juez", origen: "evidencias/adr-0037"},
}

// ficherosCopiadosDelJuez son los cuatro ficheros de la carpeta del juez que son
// copia de su original: la rúbrica, el esquema de la respuesta, los casos y la
// medida (FR-023). La declaración de clases no lo es: se escribe en la carpeta.
var ficherosCopiadosDelJuez = []string{"rubrica.md", "esquema.json", "casos.yaml", "medida.json"}

// TestCopiasDelJuez es el control de umbral de FR-023, FR-108 y SC-008 de H24
// (contracts/medida-del-juez.md §3 y §8): las cuatro copias de la carpeta del
// juez de cada skill de la tabla son idénticas, byte a byte, a su original, y
// ninguna skill tiene juez sin estar en la tabla. Cada defecto nombra su fichero.
//
// Sobre una copia de las dos carpetas en t.TempDir(), lo ve fallar: con un byte
// cambiado en cada una de las cuatro copias, y sin cada una de ellas, el único
// defecto es el de ese fichero, cuatro de cuatro; y la carpeta del juez de una
// skill que no está en la tabla da el suyo.
func TestCopiasDelJuez(t *testing.T) {
	t.Parallel()

	t.Run("del-repositorio", func(t *testing.T) {
		t.Parallel()

		defectos := defectosDeLasCopias(t, raizDelRepositorio, copiasDelJuez)
		assert.Empty(t, defectos, "copias del juez que no son las de su original:\n%s", strings.Join(defectos, "\n"))
	})

	for _, fichero := range ficherosCopiadosDelJuez {
		t.Run("cambiada-"+fichero, func(t *testing.T) {
			t.Parallel()

			probarCopiaCambiada(t, fichero)
		})

		t.Run("sin-"+fichero, func(t *testing.T) {
			t.Parallel()

			probarCopiaQueFalta(t, fichero)
		})
	}

	t.Run("skill-fuera-de-la-tabla", func(t *testing.T) {
		t.Parallel()

		raiz := copiarLasParejas(t, copiasDelJuez)
		require.NoError(t, os.MkdirAll(filepath.Join(raiz, "evals", "otra-skill", "juez"), 0o750))

		assert.Equal(t,
			[]string{"evals/otra-skill/juez: la skill tiene juez y su carpeta no está en la tabla de las copias"},
			defectosDeLasCopias(t, raiz, copiasDelJuez))
	})
}

// probarCopiaCambiada cambia un byte, el último, de la copia de ese fichero en
// una copia de las carpetas de la tabla, que antes no tenía ningún defecto, y
// exige el defecto que la nombra y ningún otro.
func probarCopiaCambiada(t *testing.T, fichero string) {
	t.Helper()

	pareja := copiasDelJuez[0]
	raiz := copiarLasParejas(t, copiasDelJuez)
	require.Empty(t, defectosDeLasCopias(t, raiz, copiasDelJuez), "la copia sin tocar no tiene ningún defecto")

	ruta := filepath.Join(raiz, pareja.copias, fichero)
	contenido := contenidoDelFichero(t, ruta)
	require.NotEmpty(t, contenido, "%s tiene algún byte que cambiar", ruta)

	contenido[len(contenido)-1] ^= 1
	require.NoError(t, os.WriteFile(ruta, contenido, 0o600))

	assert.Equal(t,
		[]string{pareja.copias + "/" + fichero + ": no es idéntico a " + pareja.origen + "/" + fichero},
		defectosDeLasCopias(t, raiz, copiasDelJuez))
}

// probarCopiaQueFalta quita la copia de ese fichero de una copia de las carpetas
// de la tabla y exige el defecto que la nombra y ningún otro.
func probarCopiaQueFalta(t *testing.T, fichero string) {
	t.Helper()

	pareja := copiasDelJuez[0]
	raiz := copiarLasParejas(t, copiasDelJuez)
	require.NoError(t, os.Remove(filepath.Join(raiz, pareja.copias, fichero)))

	defectos := defectosDeLasCopias(t, raiz, copiasDelJuez)
	require.Len(t, defectos, 1, "solo falta %s:\n%s", fichero, strings.Join(defectos, "\n"))
	assert.True(t, strings.HasPrefix(defectos[0], pareja.copias+"/"+fichero+": falta o no se puede leer: "), defectos[0])
}

// defectosDeLasCopias compara byte a byte, bajo raiz, cada fichero de
// ficherosCopiadosDelJuez de la carpeta de copias de cada pareja con el de su
// carpeta de origen, y devuelve un defecto por cada uno que difiere o que no
// se puede leer, con el nombre del fichero delante; y, detrás, uno por cada
// carpeta evals/<skill>/juez de raiz que no es la de ninguna pareja.
func defectosDeLasCopias(t *testing.T, raiz string, parejas []parejaDeCarpetas) []string {
	t.Helper()

	var defectos []string

	for _, pareja := range parejas {
		for _, fichero := range ficherosCopiadosDelJuez {
			copia, original := pareja.copias+"/"+fichero, pareja.origen+"/"+fichero

			contenidoDelOriginal, err := leerFichero(filepath.Join(raiz, original))
			if err != nil {
				defectos = append(defectos, original+": falta o no se puede leer: "+err.Error())

				continue
			}

			contenidoDeLaCopia, err := leerFichero(filepath.Join(raiz, copia))
			if err != nil {
				defectos = append(defectos, copia+": falta o no se puede leer: "+err.Error())

				continue
			}

			if !slices.Equal(contenidoDeLaCopia, contenidoDelOriginal) {
				defectos = append(defectos, copia+": no es idéntico a "+original)
			}
		}
	}

	return append(defectos, juecesFueraDeLaTabla(t, raiz, parejas)...)
}

// juecesFueraDeLaTabla devuelve un defecto por cada skill de raiz/evals que
// tiene la entrada juez sin que su carpeta sea la de copias de ninguna pareja, en
// orden de nombre. Lo que en raiz/evals no es una carpeta no son las evals de
// ninguna skill y no se mira.
func juecesFueraDeLaTabla(t *testing.T, raiz string, parejas []parejaDeCarpetas) []string {
	t.Helper()

	entradas, err := os.ReadDir(filepath.Join(raiz, "evals"))
	require.NoError(t, err, "el directorio de evals de %s", raiz)

	var defectos []string

	for _, entrada := range entradas {
		if !entrada.IsDir() {
			continue
		}

		carpeta := "evals/" + entrada.Name() + "/juez"

		_, err := os.Lstat(filepath.Join(raiz, carpeta))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}

		require.NoError(t, err, "la carpeta del juez %s", carpeta)

		if !slices.ContainsFunc(parejas, func(pareja parejaDeCarpetas) bool { return pareja.copias == carpeta }) {
			defectos = append(defectos, carpeta+": la skill tiene juez y su carpeta no está en la tabla de las copias")
		}
	}

	return defectos
}

// copiarLasParejas copia en un directorio temporal del test, con su ruta desde
// la raíz del repositorio, cada fichero de ficherosCopiadosDelJuez de las dos
// carpetas de cada pareja, y devuelve ese directorio.
func copiarLasParejas(t *testing.T, parejas []parejaDeCarpetas) string {
	t.Helper()

	raiz := t.TempDir()

	for _, pareja := range parejas {
		for _, carpeta := range []string{pareja.copias, pareja.origen} {
			require.NoError(t, os.MkdirAll(filepath.Join(raiz, carpeta), 0o750))

			for _, fichero := range ficherosCopiadosDelJuez {
				contenido := contenidoDelFichero(t, filepath.Join(raizDelRepositorio, carpeta, fichero))
				require.NoError(t, os.WriteFile(filepath.Join(raiz, carpeta, fichero), contenido, 0o600))
			}
		}
	}

	return raiz
}

// casosSinteticos son unos casos etiquetados con las tres formas de un caso
// (contracts/medida-del-juez.md §4 de H24): el de una lectura, con su frase; el
// derivado, con el bloque que se quita; y el que no lleva ninguna de las dos.
const casosSinteticos = `# Un comentario, como el de la cabecera de los casos del repositorio.
clase: una_clase
casos:
  - informe: informes/uno.json
    sesion: una-sesion-01
    grupo: ajuste
    etiqueta: defecto
    procedencia: lectura
    frase: "una frase de la respuesta"
  - informe: informes/uno.json
    sesion: una-sesion-01
    quitado: {norma: BOE-A-2015-10565, bloque: a21}
    grupo: medida
    etiqueta: defecto
    procedencia: derivado
  - informe: informes/dos.json
    sesion: otra-sesion-02
    grupo: medida
    etiqueta: correcto
    procedencia: bitacora
`

// TestLeerCasosEtiquetados fija la lectura de los casos etiquetados del juez
// (contracts/medida-del-juez.md §4 de H24; data-model §4; FR-024): la clase y
// cada caso con su informe, su sesión, su grupo, su etiqueta, su procedencia y,
// según el caso, el bloque que se quita o la frase, sin nada resuelto. Se leen
// con el lector común de YAML —una clave repetida es un error con sus dos
// líneas— y sin esquema publicado; lo único que la lectura exige de un caso es
// que su etiqueta sea una de las dos, porque con otra no se podría contar.
func TestLeerCasosEtiquetados(t *testing.T) {
	t.Parallel()

	t.Run("leidos", func(t *testing.T) {
		t.Parallel()

		leidos, err := leerCasosEtiquetados(escribirCasos(t, casosSinteticos))
		require.NoError(t, err)

		assert.Equal(t, CasosEtiquetados{
			Clase: "una_clase",
			Casos: []CasoEtiquetado{
				{
					Informe: "informes/uno.json", Sesion: "una-sesion-01", Grupo: "ajuste",
					Etiqueta: "defecto", Procedencia: "lectura", Frase: "una frase de la respuesta",
				},
				{
					Informe: "informes/uno.json", Sesion: "una-sesion-01", Grupo: "medida",
					Quitado:  &BloqueQuitado{Norma: "BOE-A-2015-10565", Bloque: "a21"},
					Etiqueta: "defecto", Procedencia: "derivado",
				},
				{
					Informe: "informes/dos.json", Sesion: "otra-sesion-02", Grupo: "medida",
					Etiqueta: "correcto", Procedencia: "bitacora",
				},
			},
		}, leidos)
	})

	ilegibles := []struct {
		nombre    string
		contenido string
		dice      string
	}{
		{nombre: "no-es-yaml", contenido: "casos: [\n", dice: "no es YAML válido"},
		{nombre: "clave-repetida", contenido: "clase: una\nclase: otra\n", dice: "clase repetido en las líneas 1 y 2"},
		{nombre: "casos-que-no-son-una-lista", contenido: "casos: ninguno\n", dice: "no se puede leer como"},
		{
			nombre:    "otra-etiqueta",
			contenido: strings.Replace(casosSinteticos, "etiqueta: correcto", "etiqueta: dudoso", 1),
			dice:      `el caso 3 (informes/dos.json otra-sesion-02) tiene la etiqueta "dudoso"`,
		},
		{
			nombre:    "sin-etiqueta",
			contenido: strings.Replace(casosSinteticos, "    etiqueta: defecto\n    procedencia: derivado\n", "", 1),
			dice:      `el caso 2 (informes/uno.json una-sesion-01 sin BOE-A-2015-10565 a21) tiene la etiqueta ""`,
		},
	}

	for _, ilegible := range ilegibles {
		t.Run(ilegible.nombre, func(t *testing.T) {
			t.Parallel()

			leidos, err := leerCasosEtiquetados(escribirCasos(t, ilegible.contenido))

			require.ErrorContains(t, err, ilegible.dice)
			assert.Zero(t, leidos, "sin los casos enteros no hay ninguno")
		})
	}

	t.Run("sin-fichero", func(t *testing.T) {
		t.Parallel()

		leidos, err := leerCasosEtiquetados(filepath.Join(t.TempDir(), ficheroDeCasosDelJuez))

		require.ErrorIs(t, err, fs.ErrNotExist)
		require.ErrorContains(t, err, "no se pueden leer: ")
		assert.Zero(t, leidos)
	})
}

// escribirCasos escribe ese contenido en el fichero de los casos de un
// directorio temporal del test y devuelve su ruta.
func escribirCasos(t *testing.T, contenido string) string {
	t.Helper()

	ruta := filepath.Join(t.TempDir(), ficheroDeCasosDelJuez)
	require.NoError(t, os.WriteFile(ruta, []byte(contenido), 0o600))

	return ruta
}

// TestArgumentosDeLaInvocacion fija con qué argumentos repite la reconstrucción
// cada invocación de una sesión (contracts/medida-del-juez.md §5, paso 3, de
// H24): las palabras de su orden; en una llamada a una herramienta,
// <applet>_<verbo> pasa a <applet> <verbo> y gana --json, que es lo que hace
// que su texto sea el sobre; y todas ganan --offline, también la que ya lo
// llevaba.
func TestArgumentosDeLaInvocacion(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre     string
		invocacion invocacionDelInforme
		argumentos []string
	}{
		{
			nombre:     "orden",
			invocacion: invocacionDelInforme{Orden: "boe articulo BOE-A-2015-10565 a21 --json"},
			argumentos: []string{"boe", "articulo", "BOE-A-2015-10565", "a21", "--json", "--offline"},
		},
		{
			nombre:     "orden-de-varias-palabras",
			invocacion: invocacionDelInforme{Orden: "boe buscar Estatuto de los  Trabajadores --json"},
			argumentos: []string{"boe", "buscar", "Estatuto", "de", "los", "Trabajadores", "--json", "--offline"},
		},
		{
			nombre:     "orden-que-ya-lo-llevaba",
			invocacion: invocacionDelInforme{Orden: "graph check BOE-A-2006-20764 a17 --json --offline"},
			argumentos: []string{"graph", "check", "BOE-A-2006-20764", "a17", "--json", "--offline", "--offline"},
		},
		{
			nombre:     "llamada",
			invocacion: invocacionDelInforme{Orden: "boe_articulos BOE-A-2015-10565 a21 a22", Llamada: true},
			argumentos: []string{"boe", "articulos", "BOE-A-2015-10565", "a21", "a22", "--json", "--offline"},
		},
		{
			nombre:     "llamada-sin-argumentos",
			invocacion: invocacionDelInforme{Orden: "boe_indice", Llamada: true},
			argumentos: []string{"boe", "indice", "--json", "--offline"},
		},
		{
			nombre:     "llamada-del-grafo",
			invocacion: invocacionDelInforme{Orden: "graph_check BOE-A-2015-10565", Llamada: true},
			argumentos: []string{"graph", "check", "BOE-A-2015-10565", "--json", "--offline"},
		},
		{
			nombre:     "llamada-sin-orden",
			invocacion: invocacionDelInforme{Llamada: true},
			argumentos: []string{"--offline"},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.argumentos, argumentosDeLaInvocacion(caso.invocacion))
		})
	}
}

// Las piezas de los sobres sintéticos de TestSinElBloque: lo que va delante de
// data, con sus cinco claves en su orden, y los elementos de data de una lectura
// de varios bloques. El texto del art. 21 lleva lo que un codificador podría
// escribir de otra manera —el signo «et» y el «menor que», que con el escape de
// HTML serían secuencias, un espacio de no separación literal y un separador de
// línea escrito como secuencia, que es como lo escribe el binario—, para que se
// vea que lo que queda del sobre queda byte a byte.
const (
	cabeceraDelSobre = `{"ok":true,"fuente":"boe.legislacion-consolidada","url":"https://www.boe.es/x?a=1&b=2",` +
		`"fecha_consulta":"2026-10-05T19:46:03.274991+02:00","hash":"sha256:0cfd96dce7ca98fb",`
	elementoDelA21 = `{"norma":"BOE-A-2015-10565","bloque":"a21","texto":"Artículo` + "\xc2\xa0" + `21. A & B <c> \` +
		`u2028 fin","avisos":[]}`
	elementoDelA22 = `{"norma":"BOE-A-2015-10565","bloque":"a22","texto":"Artículo 22.","avisos":["derogada"]}`
)

// Los sobres sintéticos de TestSinElBloque, como los escribe el binario: en una
// línea y con su salto final.
const (
	sobreDelA21       = cabeceraDelSobre + `"data":` + elementoDelA21 + "}\n"
	sobreDelA22       = cabeceraDelSobre + `"data":` + elementoDelA22 + "}\n"
	sobreDeDosBloques = cabeceraDelSobre + `"data":[` + elementoDelA21 + "," + elementoDelA22 + "]}\n"
	sobreSoloConElA21 = cabeceraDelSobre + `"data":[` + elementoDelA21 + "]}\n"
	sobreSoloConElA22 = cabeceraDelSobre + `"data":[` + elementoDelA22 + "]}\n"
	sobreDeUnFallo    = `{"ok":false,"fuente":"boe.legislacion-consolidada","url":"https://www.boe.es/x",` +
		`"fecha_consulta":"2026-10-05T19:46:03.274991+02:00","hash":"sha256:71af736cf5db217d",` +
		`"data":{"clase":"fuente-no-disponible","mensaje":"con --offline no se pide nada"}}` + "\n"
)

// textoDeLaInvocacion es el texto reconstruido de una invocación con esa orden
// —de una llamada a una herramienta, si llamada— que terminó con ese código y
// escribió esa salida: lleva los argumentos con los que la reconstrucción la
// habría repetido.
func textoDeLaInvocacion(orden string, llamada bool, codigo int, salida string) textoReconstruido {
	return textoReconstruido{
		Texto:      Texto{Orden: orden, Salida: salida},
		argumentos: argumentosDeLaInvocacion(invocacionDelInforme{Orden: orden, Llamada: llamada}),
		codigo:     codigo,
	}
}

// TestSinElBloque fija qué se quita de los textos de una sesión en un caso
// derivado (contracts/medida-del-juez.md §5, paso 4, de H24): fuera el texto de
// cada boe articulo de la norma y el bloque quitados que terminó con 0, por
// orden o por llamada; del de cada boe articulos que lo pidió, los elementos de
// data con ese bloque, con el sobre vuelto a escribir con sus seis claves en su
// orden y lo demás byte a byte, y fuera el texto si no queda ninguno; y nada
// más: ni la lectura que no terminó con 0, ni la de otra norma u otro bloque,
// ni lo que no es una lectura de bloques. Si no se quita nada, no hay derivado:
// es un error que nombra el bloque.
func TestSinElBloque(t *testing.T) {
	t.Parallel()

	const (
		leerElA21   = "boe articulo BOE-A-2015-10565 a21 --json"
		leerLosDos  = "boe articulos BOE-A-2015-10565 a21 a22 --json"
		llamarAlA21 = "boe_articulo BOE-A-2015-10565 a21"
		llamarALos2 = "boe_articulos BOE-A-2015-10565 a22 a21"
		leerOtraLey = "boe articulo BOE-A-2015-10566 a21 --json"
		leerOtros2  = "boe articulos BOE-A-2015-10565 a22 a23 --json"
	)

	quitado := BloqueQuitado{Norma: "BOE-A-2015-10565", Bloque: "a21"}

	elIndice := textoDeLaInvocacion("boe indice BOE-A-2015-10565 --json", false, 0, `{"ok":true}`+"\n")
	laComprobacion := textoDeLaInvocacion("graph check BOE-A-2015-10565 a21 --json", false, 0, `{"ok":true}`+"\n")
	elA21 := textoDeLaInvocacion(leerElA21, false, 0, sobreDelA21)
	elA22 := textoDeLaInvocacion("boe articulo BOE-A-2015-10565 a22 --json", false, 0, sobreDelA22)

	casos := []struct {
		nombre string
		textos []textoReconstruido
		quedan []Texto
	}{
		{
			nombre: "articulo-por-orden",
			textos: []textoReconstruido{elIndice, elA21, laComprobacion},
			quedan: []Texto{elIndice.Texto, laComprobacion.Texto},
		},
		{
			nombre: "articulo-por-llamada",
			textos: []textoReconstruido{textoDeLaInvocacion(llamarAlA21, true, 0, sobreDelA21), laComprobacion},
			quedan: []Texto{laComprobacion.Texto},
		},
		{
			nombre: "todas-sus-lecturas",
			textos: []textoReconstruido{elA21, elA22, elA21},
			quedan: []Texto{elA22.Texto},
		},
		{
			nombre: "la-unica-lectura",
			textos: []textoReconstruido{elA21},
			quedan: []Texto{},
		},
		{
			nombre: "articulos-con-otro-bloque",
			textos: []textoReconstruido{elIndice, textoDeLaInvocacion(leerLosDos, false, 0, sobreDeDosBloques)},
			quedan: []Texto{elIndice.Texto, {Orden: leerLosDos, Salida: sobreSoloConElA22}},
		},
		{
			nombre: "articulos-por-llamada",
			textos: []textoReconstruido{textoDeLaInvocacion(llamarALos2, true, 0, sobreDeDosBloques)},
			quedan: []Texto{{Orden: llamarALos2, Salida: sobreSoloConElA22}},
		},
		{
			nombre: "articulos-solo-con-el-bloque",
			textos: []textoReconstruido{
				textoDeLaInvocacion("boe articulos BOE-A-2015-10565 a21 --json", false, 0, sobreSoloConElA21),
				laComprobacion,
			},
			quedan: []Texto{laComprobacion.Texto},
		},
		{
			nombre: "lo-que-no-es-su-lectura",
			textos: []textoReconstruido{
				textoDeLaInvocacion(leerElA21, false, codigoFuenteNoDisponible, sobreDeUnFallo),
				textoDeLaInvocacion(leerLosDos, false, codigoFuenteNoDisponible, sobreDeUnFallo),
				textoDeLaInvocacion(leerOtraLey, false, 0, sobreDelA21),
				textoDeLaInvocacion(leerOtros2, false, 0, sobreSoloConElA22),
				elA22, elIndice, laComprobacion, elA21,
			},
			quedan: []Texto{
				{Orden: leerElA21, Salida: sobreDeUnFallo},
				{Orden: leerLosDos, Salida: sobreDeUnFallo},
				{Orden: leerOtraLey, Salida: sobreDelA21},
				{Orden: leerOtros2, Salida: sobreSoloConElA22},
				elA22.Texto, elIndice.Texto, laComprobacion.Texto,
			},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			quedan, err := sinElBloque(caso.textos, quitado)

			require.NoError(t, err)
			assert.Equal(t, caso.quedan, quedan)
		})
	}

	sinDerivado := []struct {
		nombre string
		textos []textoReconstruido
		dice   string
	}{
		{nombre: "sin-textos", dice: "ninguna lectura de BOE-A-2015-10565 a21 terminó con 0"},
		{
			nombre: "sin-ninguna-lectura-suya",
			textos: []textoReconstruido{
				elIndice, elA22, laComprobacion,
				textoDeLaInvocacion(leerElA21, false, codigoFuenteNoDisponible, sobreDeUnFallo),
			},
			dice: "ninguna lectura de BOE-A-2015-10565 a21 terminó con 0",
		},
		{
			nombre: "articulos-sin-sobre",
			textos: []textoReconstruido{textoDeLaInvocacion(leerLosDos, false, 0, "Artículo 21.\n")},
			dice:   "la salida de «" + leerLosDos + "» no es el sobre de una lectura de bloques",
		},
		{
			nombre: "articulos-con-un-elemento-que-no-es-un-bloque",
			textos: []textoReconstruido{
				textoDeLaInvocacion(leerLosDos, false, 0, cabeceraDelSobre+`"data":["a21"]}`+"\n"),
			},
			dice: "la salida de «" + leerLosDos + "» no es el sobre de una lectura de bloques",
		},
	}

	for _, caso := range sinDerivado {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			quedan, err := sinElBloque(caso.textos, quitado)

			require.ErrorContains(t, err, caso.dice)
			assert.Nil(t, quedan)
		})
	}
}

// El mundo sintético de TestResolverCasos: un informe con cuatro sesiones, que
// el test escribe bajo una raíz temporal, los nombres de esas sesiones y las
// evals que nombran: la sintética del art. 21 de la LPAC, la retirada del
// repositorio y una que no está en ningún sitio.
const (
	informeSintetico = "informes/uno.json"

	// sesionDeOrdenes es la de una sesión del modo orden: sus invocaciones son
	// órdenes, con la clave llamada a false. La última leyó un bloque que no
	// está en ninguna grabación.
	sesionDeOrdenes = "01-lpac-articulo-21-un-modelo-01"

	// sesionDeLlamadas es la de una sesión del modo herramienta: el proceso del
	// servidor, que no da texto, y tres llamadas.
	sesionDeLlamadas = "01-lpac-articulo-21-un-modelo-02"

	// sesionDeLaRetirada es la de una sesión de la eval retirada, de un informe
	// anterior a los dos modos: sus invocaciones no llevan la clave llamada.
	sesionDeLaRetirada = "19-lpac-articulo-21-redaccion-cambiada-un-modelo-01"

	// sesionSinEval es la de una sesión cuya eval no está entre las de hoy ni
	// entre las retiradas del repositorio.
	sesionSinEval = "09-no-esta-un-modelo-01"

	evalRetiradaDelRepositorio = "19-lpac-articulo-21-redaccion-cambiada.yaml"
	evalQueNoEsta              = "09-no-esta.yaml"

	respuestaDeLasOrdenes = "El artículo 21 obliga a resolver.\n\n  Con dos espacios delante y un salto detrás.\n"
)

// contenidoDelInformeSintetico es el informe del mundo sintético, con la forma
// de un informe versionado del job de evals (research V2 de H24) y con claves
// que la reconstrucción no lee.
const contenidoDelInformeSintetico = `{
  "commit": "0000000000000000000000000000000000000000",
  "modelo_que_decide": "un-modelo",
  "evals": [
    {
      "sesion": "01-lpac-articulo-21-un-modelo-01",
      "eval": "01-lpac-articulo-21.yaml",
      "modo": "orden",
      "pasa": true,
      "respuesta": "El artículo 21 obliga a resolver.\n\n  Con dos espacios delante y un salto detrás.\n",
      "invocaciones": [
        {"orden": "boe indice BOE-A-2015-10565 --json", "codigo": 0, "conexiones": [], "llamada": false},
        {"orden": "boe articulo BOE-A-2015-10565 a21 --json", "codigo": 0, "conexiones": [], "llamada": false},
        {"orden": "graph check BOE-A-2015-10565 a21 --json", "codigo": 0, "conexiones": [], "llamada": false},
        {"orden": "boe articulo BOE-A-2015-10565 a9998 --json", "codigo": 5, "conexiones": [], "llamada": false}
      ]
    },
    {
      "sesion": "01-lpac-articulo-21-un-modelo-02",
      "eval": "01-lpac-articulo-21.yaml",
      "modo": "herramienta",
      "pasa": true,
      "respuesta": "Otra respuesta.",
      "invocaciones": [
        {"orden": "mcp serve", "codigo": null, "conexiones": [], "llamada": false},
        {"orden": "boe_articulos BOE-A-2015-10565 a21 a22", "codigo": 0, "conexiones": [], "llamada": true},
        {"orden": "boe_articulo BOE-A-2015-10565 a21", "codigo": 0, "conexiones": [], "llamada": true},
        {"orden": "graph_check BOE-A-2015-10565", "codigo": 0, "conexiones": [], "llamada": true}
      ]
    },
    {
      "sesion": "19-lpac-articulo-21-redaccion-cambiada-un-modelo-01",
      "eval": "19-lpac-articulo-21-redaccion-cambiada.yaml",
      "pasa": false,
      "respuesta": "La respuesta de la eval retirada.",
      "invocaciones": [
        {"orden": "boe articulo BOE-A-2015-10565 a21 --json", "codigo": 0, "conexiones": []},
        {"orden": "graph check BOE-A-2015-10565 a21 --json", "codigo": 0, "conexiones": []}
      ]
    },
    {
      "sesion": "09-no-esta-un-modelo-01",
      "eval": "09-no-esta.yaml",
      "respuesta": "La respuesta de una eval que no está.",
      "invocaciones": []
    }
  ]
}
`

// evalsDelMundoSintetico son las evals de hoy del mundo sintético: la del art. 21 y la
// del art. 22 de la LPAC, cuyas consultas llenan la base.
func evalsDelMundoSintetico() []entradaDeConjunto {
	return []entradaDeConjunto{
		{nombre: nombreDeEval, contenido: contenidoDelArticulo21},
		{nombre: "02-lpac-articulo-22.yaml", contenido: contenidoDelArticulo22},
	}
}

// reconstructorSintetico es el reconstructor de un mundo sintético: la raíz es
// un directorio temporal del test con esos informes, cada uno en su ruta desde
// ella, y las evals de hoy son esas entradas. Las grabaciones, el grafo previo
// de cada eval y las evals retiradas son los del repositorio, que solo se leen.
func reconstructorSintetico(t *testing.T, informes map[string]string, evals []entradaDeConjunto) *reconstructor {
	t.Helper()

	raiz := t.TempDir()

	for ruta, contenido := range informes {
		crearEntradas(t, raiz, []entradaDeConjunto{{nombre: filepath.FromSlash(ruta), contenido: contenido}})
	}

	sintetico := nuevoReconstructor(crearConjunto(t, evals))
	sintetico.raiz = raiz

	return sintetico
}

// elInformeSintetico es el mundo de un solo informe, el sintético.
func elInformeSintetico() map[string]string {
	return map[string]string{informeSintetico: contenidoDelInformeSintetico}
}

// casoDe es un caso etiquetado de esa sesión del informe sintético, sin
// resolver: con quitado, el derivado que quita ese bloque de la LPAC.
func casoDe(sesion, quitado string) CasoEtiquetado {
	caso := CasoEtiquetado{
		Informe: informeSintetico, Sesion: sesion, Grupo: "ajuste",
		Etiqueta: etiquetaCorrecto, Procedencia: "lectura", Frase: "una frase",
	}

	if quitado != "" {
		caso.Quitado = &BloqueQuitado{Norma: "BOE-A-2015-10565", Bloque: quitado}
		caso.Etiqueta, caso.Procedencia, caso.Frase = etiquetaDefecto, "derivado", ""
	}

	return caso
}

// casoDelInforme es el caso de la sesión de órdenes del mundo sintético, pero
// de ese otro informe.
func casoDelInforme(informe string) CasoEtiquetado {
	caso := casoDe(sesionDeOrdenes, "")
	caso.Informe = informe

	return caso
}

// TestResolverCasos fija la resolución de un caso etiquetado sobre un mundo
// sintético (contracts/medida-del-juez.md §4 y §5 de H24; FR-024, FR-051): la
// respuesta es la de su sesión en su informe, tal cual; la pregunta, la del
// fichero de eval que nombra esa sesión, de las de hoy o de las retiradas; y los
// textos, los de repetir en proceso cada invocación de la sesión, en su orden y
// salvo mcp serve, con --offline —y con --json en una llamada— sobre la caché
// de la base y, si la eval lo tiene, sobre su grafo previo. En un derivado, sin
// el texto del bloque quitado. Lo que ya se ha reconstruido de una sesión se
// recuerda. Y un caso que no se puede resolver es un error que lo nombra.
func TestResolverCasos(t *testing.T) {
	t.Parallel()

	sintetico := reconstructorSintetico(t, elInformeSintetico(), evalsDelMundoSintetico())

	casos := []CasoEtiquetado{
		casoDe(sesionDeOrdenes, ""),
		casoDe(sesionDeOrdenes, "a21"),
		casoDe(sesionDeLlamadas, ""),
		casoDe(sesionDeLlamadas, "a21"),
		casoDe(sesionDeLlamadas, "a22"),
		casoDe(sesionDeLaRetirada, ""),
	}

	resueltos, err := sintetico.resolver(casos)
	require.NoError(t, err)
	require.Len(t, resueltos, len(casos))

	deOrdenes, sinElA21DeOrdenes := resueltos[0], resueltos[1]
	deLlamadas, sinElA21DeLlamadas, sinElA22DeLlamadas := resueltos[2], resueltos[3], resueltos[4]
	deLaRetirada := resueltos[5]

	t.Run("lo-que-el-caso-ya-decia", func(t *testing.T) {
		t.Parallel()

		for posicion, resuelto := range resueltos {
			resuelto.Pregunta, resuelto.Respuesta, resuelto.Textos = "", "", nil
			assert.Equal(t, casos[posicion], resuelto, "el caso %d resuelto es el mismo caso", posicion+1)
		}
	})

	t.Run("pregunta-y-respuesta", func(t *testing.T) {
		t.Parallel()

		for _, resuelto := range []CasoEtiquetado{deOrdenes, sinElA21DeOrdenes} {
			assert.Equal(t, "¿qué dice el art. 21 de la Ley 39/2015?", resuelto.Pregunta)
			assert.Equal(t, respuestaDeLasOrdenes, resuelto.Respuesta, "la respuesta va tal cual, con sus blancos")
		}

		assert.Equal(t, "Otra respuesta.", deLlamadas.Respuesta)
		assert.Equal(t, preguntaLeidaAparte(t, evalRetiradaDelRepositorio), deLaRetirada.Pregunta,
			"la pregunta de una eval retirada es la de su fichero de %s", EvalsRetiradas)
		assert.Equal(t, "La respuesta de la eval retirada.", deLaRetirada.Respuesta)
	})

	t.Run("ordenes", func(t *testing.T) {
		t.Parallel()

		exigirLosTextosDeLasOrdenes(t, deOrdenes.Textos)
	})

	t.Run("llamadas", func(t *testing.T) {
		t.Parallel()

		exigirLosTextosDeLasLlamadas(t, deLlamadas.Textos)
	})

	t.Run("grafo-previo-de-la-retirada", func(t *testing.T) {
		t.Parallel()

		require.Len(t, deLaRetirada.Textos, 2)
		assert.Equal(t, 1, datosDelSobre[comprobacionLeida](t, deLaRetirada.Textos[1].Salida).Obsoletas,
			"el grafo de la sesión ya tenía la redacción anterior, la del grafo previo de la eval retirada")

		require.Len(t, deOrdenes.Textos, 4)
		assert.Zero(t, datosDelSobre[comprobacionLeida](t, deOrdenes.Textos[2].Salida).Obsoletas,
			"la sesión de una eval sin grafo previo empieza sin grafo")
	})

	t.Run("derivados", func(t *testing.T) {
		t.Parallel()

		require.Len(t, deOrdenes.Textos, 4)
		assert.Equal(t, []Texto{deOrdenes.Textos[0], deOrdenes.Textos[2], deOrdenes.Textos[3]}, sinElA21DeOrdenes.Textos,
			"sin la lectura del a21, que terminó con 0, y con la del a9998, que no")

		require.Len(t, deLlamadas.Textos, 3)

		require.Len(t, sinElA21DeLlamadas.Textos, 2, "sin la llamada que leyó el a21")
		assert.Equal(t, []string{"a22"}, bloquesDelSobre(t, sinElA21DeLlamadas.Textos[0].Salida))
		assert.Equal(t, deLlamadas.Textos[2], sinElA21DeLlamadas.Textos[1])

		require.Len(t, sinElA22DeLlamadas.Textos, 3)
		assert.Equal(t, []string{"a21"}, bloquesDelSobre(t, sinElA22DeLlamadas.Textos[0].Salida))
		assert.Equal(t, deLlamadas.Textos[1:], sinElA22DeLlamadas.Textos[1:])

		for _, derivado := range []CasoEtiquetado{sinElA21DeLlamadas, sinElA22DeLlamadas} {
			quitados := exigirElSobreSinSusElementos(t, deLlamadas.Textos[0], derivado.Textos[0], derivado.Quitado.Bloque)
			assert.Equal(t, 1, quitados)
		}
	})

	t.Run("lo-reconstruido-se-recuerda", func(t *testing.T) {
		t.Parallel()

		// El sobre de graph check lleva el instante de su invocación: si la
		// sesión se reconstruyera otra vez, no sería el mismo.
		otraVez, err := sintetico.resolver(casos)

		require.NoError(t, err)
		assert.Equal(t, resueltos, otraVez)
	})

	for _, caso := range casosSinResolver() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			resueltos, err := sintetico.resolver([]CasoEtiquetado{casoDe(sesionDeOrdenes, ""), caso.caso})

			for _, fragmento := range caso.dice {
				require.ErrorContains(t, err, fragmento)
			}

			if caso.es != nil {
				require.ErrorIs(t, err, caso.es)
			}

			assert.Nil(t, resueltos, "con un caso sin resolver no hay ninguno")
		})
	}
}

// casoSinResolver es un caso del informe sintético que no se puede resolver, con
// lo que su error tiene que decir y, si lo hay, el error que envuelve.
type casoSinResolver struct {
	nombre string
	caso   CasoEtiquetado
	dice   []string
	es     error
}

// casosSinResolver son los casos que la reconstrucción del mundo sintético de
// TestResolverCasos no puede resolver, cada uno por lo suyo: su error empieza
// por el caso, con su informe, su sesión y, en un derivado, su bloque.
func casosSinResolver() []casoSinResolver {
	return []casoSinResolver{
		{
			nombre: "informe-que-no-esta",
			caso:   casoDelInforme("informes/no-esta.json"),
			dice: []string{
				"el caso informes/no-esta.json " + sesionDeOrdenes + ": el informe informes/no-esta.json no se puede leer: ",
			},
			es: fs.ErrNotExist,
		},
		{
			nombre: "informe-fuera-de-la-raiz",
			caso:   casoDelInforme("../fuera.json"),
			dice:   []string{"el informe ../fuera.json no se puede leer: "},
			es:     fs.ErrInvalid,
		},
		{
			nombre: "sesion-que-no-esta",
			caso:   casoDe("una-sesion-que-no-esta", ""),
			dice: []string{
				"el caso " + informeSintetico + " una-sesion-que-no-esta: el informe " + informeSintetico +
					" no tiene la sesión una-sesion-que-no-esta",
			},
		},
		{
			nombre: "eval-que-no-esta",
			caso:   casoDe(sesionSinEval, ""),
			dice:   []string{"el caso " + informeSintetico + " " + sesionSinEval + ": la eval " + evalQueNoEsta, EvalsRetiradas},
			es:     fs.ErrNotExist,
		},
		{
			nombre: "derivado-sin-lectura",
			caso:   casoDe(sesionDeOrdenes, "a22"),
			dice: []string{
				"el caso " + informeSintetico + " " + sesionDeOrdenes + " sin BOE-A-2015-10565 a22: " +
					"ninguna lectura de BOE-A-2015-10565 a22 terminó con 0",
			},
		},
		{
			nombre: "derivado-de-una-lectura-fallida",
			caso:   casoDe(sesionDeOrdenes, "a9998"),
			dice:   []string{"ninguna lectura de BOE-A-2015-10565 a9998 terminó con 0"},
		},
	}
}

// Lo que los tests de la reconstrucción leen del sobre de una salida, por su
// cuenta: sus dos primeras claves y su data, que es, según la orden, un bloque,
// una lista de bloques, una comprobación del grafo o un fallo.
type (
	sobreLeido struct {
		Ok     bool           `json:"ok"`
		Fuente string         `json:"fuente"`
		Data   jsontext.Value `json:"data"`
	}

	bloqueLeido struct {
		Bloque string `json:"bloque"`
		Texto  string `json:"texto"`
	}

	comprobacionLeida struct {
		Bloques   []string `json:"bloques"`
		Obsoletas int      `json:"version-obsoleta"`
	}

	falloLeido struct {
		Clase   string `json:"clase"`
		Mensaje string `json:"mensaje"`
	}
)

// Las fuentes y la clase de error con las que los tests de la reconstrucción
// reconocen cada sobre.
const (
	fuenteDelBoeEnElSobre   = "boe.legislacion-consolidada"
	fuenteDelGrafoEnElSobre = "kitlegal.graph"
	claseFuenteNoDisponible = "fuente-no-disponible"
)

// clavesDelSobre son las seis claves del sobre de salida, en su orden.
var clavesDelSobre = []string{"ok", "fuente", "url", "fecha_consulta", "hash", "data"}

// leerSobre lee la salida de un texto como un sobre.
func leerSobre(t *testing.T, salida string) sobreLeido {
	t.Helper()

	var sobre sobreLeido
	require.NoError(t, json.Unmarshal([]byte(salida), &sobre), "no es un sobre: %s", salida)

	return sobre
}

// datosDelSobre lee como un T el data del sobre de una salida.
func datosDelSobre[T any](t *testing.T, salida string) T {
	t.Helper()

	var datos T
	require.NoError(t, json.Unmarshal(leerSobre(t, salida).Data, &datos), "data no tiene esa forma en %s", salida)

	return datos
}

// bloquesDelSobre son los bloques de los elementos de data del sobre de una
// lectura de varios bloques, en su orden.
func bloquesDelSobre(t *testing.T, salida string) []string {
	t.Helper()

	bloques := []string{}
	for _, elemento := range datosDelSobre[[]bloqueLeido](t, salida) {
		bloques = append(bloques, elemento.Bloque)
	}

	return bloques
}

// ordenesDe son las órdenes de los textos, en su orden.
func ordenesDe(textos []Texto) []string {
	ordenes := make([]string, 0, len(textos))
	for _, texto := range textos {
		ordenes = append(ordenes, texto.Orden)
	}

	return ordenes
}

// clavesEnSuOrden son las claves del objeto JSON del texto, en el orden en que
// están escritas.
func clavesEnSuOrden(t *testing.T, texto string) []string {
	t.Helper()

	lector := jsontext.NewDecoder(strings.NewReader(texto))

	inicio, err := lector.ReadToken()
	require.NoError(t, err, "no es JSON: %s", texto)
	require.Equal(t, jsontext.KindBeginObject, inicio.Kind(), "no es un objeto JSON: %s", texto)

	var claves []string

	for lector.PeekKind() == jsontext.KindString {
		clave, err := lector.ReadToken()
		require.NoError(t, err)

		claves = append(claves, clave.String())

		require.NoError(t, lector.SkipValue())
	}

	return claves
}

// valorJSON lee el texto como un valor JSON cualquiera, con sus objetos como
// mapas y sus listas como listas, que es como se comparan dos valores sin mirar
// cómo están escritos.
func valorJSON(t *testing.T, texto []byte) any {
	t.Helper()

	var valor any
	require.NoError(t, json.Unmarshal(texto, &valor), "no es JSON: %s", texto)

	return valor
}

// exigirLosTextosDeLasOrdenes exige los textos de la sesión de órdenes del mundo
// sintético: uno por orden, con la orden del informe tal cual y, de salida, el
// sobre de repetirla con --offline sobre la caché de la base, como lo escribe el
// binario: lo grabado termina con 0 y lo que no está en ninguna grabación, sin
// red, con el sobre de su fallo.
func exigirLosTextosDeLasOrdenes(t *testing.T, textos []Texto) {
	t.Helper()

	require.Equal(t, []string{
		"boe indice BOE-A-2015-10565 --json",
		"boe articulo BOE-A-2015-10565 a21 --json",
		"graph check BOE-A-2015-10565 a21 --json",
		"boe articulo BOE-A-2015-10565 a9998 --json",
	}, ordenesDe(textos))

	for _, texto := range textos {
		assert.Equal(t, clavesDelSobre, clavesEnSuOrden(t, texto.Salida), "la salida de «%s» es su sobre", texto.Orden)
		assert.True(t, strings.HasSuffix(texto.Salida, "}\n"), "la salida de «%s» va como la escribe el binario", texto.Orden)
	}

	indice, articulo, comprobacion, fallo := textos[0], textos[1], textos[2], textos[3]

	assert.True(t, leerSobre(t, indice.Salida).Ok)
	assert.Equal(t, fuenteDelBoeEnElSobre, leerSobre(t, indice.Salida).Fuente)

	assert.True(t, leerSobre(t, articulo.Salida).Ok)
	assert.Equal(t, "a21", datosDelSobre[bloqueLeido](t, articulo.Salida).Bloque)
	assert.Contains(t, datosDelSobre[bloqueLeido](t, articulo.Salida).Texto, "Obligación de resolver")

	assert.True(t, leerSobre(t, comprobacion.Salida).Ok)
	assert.Equal(t, fuenteDelGrafoEnElSobre, leerSobre(t, comprobacion.Salida).Fuente)
	assert.Equal(t, []string{"a21"}, datosDelSobre[comprobacionLeida](t, comprobacion.Salida).Bloques)

	assert.False(t, leerSobre(t, fallo.Salida).Ok)
	assert.Equal(t, claseFuenteNoDisponible, datosDelSobre[falloLeido](t, fallo.Salida).Clase)
	assert.Contains(t, datosDelSobre[falloLeido](t, fallo.Salida).Mensaje, "con --offline no se pide nada a la fuente",
		"la orden se repite con --offline")
}

// exigirLosTextosDeLasLlamadas exige los textos de la sesión de llamadas del
// mundo sintético: ninguno del proceso del servidor y uno por llamada, con la
// orden que publica el informe tal cual —la herramienta y sus argumentos— y, de
// salida, el sobre de su orden equivalente, que es lo que da --json.
func exigirLosTextosDeLasLlamadas(t *testing.T, textos []Texto) {
	t.Helper()

	require.Equal(t, []string{
		"boe_articulos BOE-A-2015-10565 a21 a22",
		"boe_articulo BOE-A-2015-10565 a21",
		"graph_check BOE-A-2015-10565",
	}, ordenesDe(textos), "mcp serve no da ningún texto")

	for _, texto := range textos {
		assert.Equal(t, clavesDelSobre, clavesEnSuOrden(t, texto.Salida), "la salida de «%s» es su sobre", texto.Orden)
		assert.True(t, leerSobre(t, texto.Salida).Ok, "«%s» termina con 0", texto.Orden)
	}

	assert.Equal(t, []string{"a21", "a22"}, bloquesDelSobre(t, textos[0].Salida))
	assert.Equal(t, "a21", datosDelSobre[bloqueLeido](t, textos[1].Salida).Bloque)
	assert.Equal(t, fuenteDelGrafoEnElSobre, leerSobre(t, textos[2].Salida).Fuente)
}

// exigirElSobreSinSusElementos exige que el texto del derivado sea el del boe
// articulos de la sesión sin los elementos de data de ese bloque y con todo lo
// demás igual: la misma orden, las seis claves en su orden y los mismos valores
// JSON. Devuelve cuántos elementos se quitan, que es al menos uno.
func exigirElSobreSinSusElementos(t *testing.T, deLaSesion, delDerivado Texto, bloque string) int {
	t.Helper()

	esperado, esObjeto := valorJSON(t, []byte(deLaSesion.Salida)).(map[string]any)
	require.True(t, esObjeto, "la salida de «%s» es un objeto JSON", deLaSesion.Orden)

	elementos, esLista := esperado["data"].([]any)
	require.True(t, esLista, "data es una lista en la salida de «%s»", deLaSesion.Orden)

	quedan := make([]any, 0, len(elementos))

	for posicion, bloqueDelElemento := range bloquesDelSobre(t, deLaSesion.Salida) {
		if bloqueDelElemento != bloque {
			quedan = append(quedan, elementos[posicion])
		}
	}

	esperado["data"] = quedan
	quitados := len(elementos) - len(quedan)

	assert.Equal(t, deLaSesion.Orden, delDerivado.Orden)
	assert.Equal(t, clavesDelSobre, clavesEnSuOrden(t, delDerivado.Salida),
		"el sobre de «%s» se vuelve a escribir con sus seis claves en su orden", delDerivado.Orden)
	assert.Equal(t, esperado, valorJSON(t, []byte(delDerivado.Salida)),
		"el sobre de «%s» lleva los mismos valores JSON menos los elementos de %s", delDerivado.Orden, bloque)
	assert.Positive(t, quitados, "el sobre de «%s» trae el bloque %s que pidió", deLaSesion.Orden, bloque)

	return quitados
}

// TestResolverCasosSinPoderPreparar fija los errores de la resolución de un caso
// cuando lo que la reconstrucción lee o prepara no sirve
// (contracts/medida-del-juez.md §4 y §5 de H24): las evals de hoy que no se
// pueden listar o que tienen un fichero mal formado, la base a la que le falta
// una consulta de esas evals, el informe que no es un informe o que repite una
// sesión, la eval retirada que no se puede leer o que no tiene pregunta, y el
// grafo previo que no se puede copiar o cuyo comando no termina con 0. Ninguno
// resuelve nada, y cada error nombra el caso y lo que falla.
func TestResolverCasosSinPoderPreparar(t *testing.T) {
	t.Parallel()

	const (
		informeRoto     = "informes/roto.json"
		informeRepetido = "informes/repetido.json"
		grafoDeLaLCSP   = "lcsp-a1-30-redaccion-original"
	)

	conGrafoPrevio := func(grabaciones, bloque string) []entradaDeConjunto {
		return []entradaDeConjunto{
			{nombre: nombreDeEval, contenido: contenidoDelArticulo21 + grafoPrevioDe(grabaciones, bloque)},
		}
	}

	retiradas := crearConjunto(t, []entradaDeConjunto{
		{nombre: evalRetiradaDelRepositorio, contenido: "pregunta: [\n"},
		{nombre: evalQueNoEsta, contenido: "activa: true\n"},
	})

	casos := []struct {
		nombre    string
		informes  map[string]string
		evals     []entradaDeConjunto
		retiradas string
		caso      CasoEtiquetado
		dice      []string
	}{
		{
			nombre: "evals-mal-formadas",
			evals:  append(evalsDelMundoSintetico(), entradaDeConjunto{nombre: "03-sin-pregunta.yaml", contenido: contenidoSinPregunta}),
			caso:   casoDe(sesionDeOrdenes, ""),
			dice:   []string{"tiene ficheros mal formados", "03-sin-pregunta.yaml"},
		},
		{
			nombre: "base-sin-lo-grabado",
			evals: append(evalsDelMundoSintetico(),
				entradaDeConjunto{nombre: "03-lpac-articulo-9998.yaml", contenido: contenidoDelArticulo9998}),
			caso: casoDe(sesionDeOrdenes, ""),
			dice: []string{"la base no sirve las consultas de las evals de ", "03-lpac-articulo-9998.yaml", "a9998"},
		},
		{
			nombre:   "informe-que-no-es-json",
			informes: map[string]string{informeRoto: "{"},
			evals:    evalsDelMundoSintetico(),
			caso:     casoDelInforme(informeRoto),
			dice:     []string{"el informe " + informeRoto + " no es un informe del job de evals: "},
		},
		{
			nombre: "informe-con-la-sesion-repetida",
			informes: map[string]string{
				informeRepetido: `{"evals": [{"sesion": "una", "respuesta": "a"}, {"sesion": "una", "respuesta": "b"}]}`,
			},
			evals: evalsDelMundoSintetico(),
			caso:  casoDelInforme(informeRepetido),
			dice:  []string{"el informe " + informeRepetido + " tiene repetida la sesión una"},
		},
		{
			nombre:    "retirada-ilegible",
			evals:     evalsDelMundoSintetico(),
			retiradas: retiradas,
			caso:      casoDe(sesionDeLaRetirada, ""),
			dice:      []string{"la eval retirada " + evalRetiradaDelRepositorio + " de " + retiradas + ": no es YAML válido"},
		},
		{
			nombre:    "retirada-sin-pregunta",
			evals:     evalsDelMundoSintetico(),
			retiradas: retiradas,
			caso:      casoDe(sesionSinEval, ""),
			dice:      []string{"la eval retirada " + evalQueNoEsta + " de " + retiradas + " no tiene pregunta"},
		},
		{
			nombre: "grafo-previo-que-no-se-copia",
			evals:  conGrafoPrevio(grafoPrevioQueNoEsta, "a21"),
			caso:   casoDe(sesionDeOrdenes, ""),
			dice: []string{
				"preparar el grafo previo " + grafoPrevioQueNoEsta + " de la eval " + nombreDeEval,
				filepath.Join(GrafosPrevios, grafoPrevioQueNoEsta),
			},
		},
		{
			nombre: "grafo-previo-con-una-falta",
			evals:  conGrafoPrevio(grafoDeLaLCSP, "a9998"),
			caso:   casoDe(sesionDeOrdenes, ""),
			dice: []string{
				"el grafo previo " + grafoDeLaLCSP + " de la eval " + nombreDeEval + " no se puede preparar:",
				"el comando del grafo previo boe articulo BOE-A-2015-10565 a9998",
			},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			informes := elInformeSintetico()
			for ruta, contenido := range caso.informes {
				informes[ruta] = contenido
			}

			sintetico := reconstructorSintetico(t, informes, caso.evals)
			if caso.retiradas != "" {
				sintetico.retiradas = caso.retiradas
			}

			resueltos, err := sintetico.resolver([]CasoEtiquetado{caso.caso})

			require.ErrorContains(t, err, "el caso "+caso.caso.nombre()+": ")

			for _, fragmento := range caso.dice {
				require.ErrorContains(t, err, fragmento)
			}

			assert.Nil(t, resueltos)
		})
	}

	t.Run("evals-que-no-se-listan", func(t *testing.T) {
		t.Parallel()

		sintetico := reconstructorSintetico(t, elInformeSintetico(), nil)
		sintetico.evals = filepath.Join(t.TempDir(), "no-esta")

		resueltos, err := sintetico.resolver([]CasoEtiquetado{casoDe(sesionDeOrdenes, "")})

		require.ErrorIs(t, err, fs.ErrNotExist)
		require.ErrorContains(t, err, "el directorio de evals "+sintetico.evals+" no se puede listar")
		assert.Nil(t, resueltos)
	})
}

// TestBaseDeLaReconstruccion fija la vigencia de la base de una reconstrucción
// (contracts/medida-del-juez.md §5, paso 1, de H24), con el reloj en la mano del
// test: la caché que llenan las consultas de las evals de hoy sirve a todas las
// sesiones mientras no tenga más de 100 s; con más, se rehace en otro
// directorio y la anterior se retira, porque lo que guarda de buscar y de
// metadatos caduca a los 300 s; y al retirarla no queda nada.
func TestBaseDeLaReconstruccion(t *testing.T) {
	t.Parallel()

	ahora := time.Date(2026, time.October, 5, 12, 0, 0, 0, time.UTC)

	sintetico := reconstructorSintetico(t, nil, evalsDelMundoSintetico())
	sintetico.ahora = func() time.Time { return ahora }

	primera, err := sintetico.baseVigente()
	require.NoError(t, err)

	t.Cleanup(func() { assert.NoError(t, sintetico.retirarLaBase()) })

	faltas, err := ComprobarSinRed(primera, ConsultasNecesarias(evalsDe(t, sintetico.evals)))
	require.NoError(t, err)
	assert.Empty(t, faltas, "la base sirve sin red las consultas de las evals de hoy")

	ahora = ahora.Add(vigenciaDeLaBase)

	todavia, err := sintetico.baseVigente()
	require.NoError(t, err)
	assert.Equal(t, primera, todavia, "con 100 s la base sigue sirviendo")

	ahora = ahora.Add(time.Nanosecond)

	rehecha, err := sintetico.baseVigente()
	require.NoError(t, err)
	assert.NotEqual(t, primera, rehecha, "con más de 100 s la base se rehace")
	assert.NoDirExists(t, primera, "la base anterior se retira")
	assert.DirExists(t, rehecha)

	misma, err := sintetico.baseVigente()
	require.NoError(t, err)
	assert.Equal(t, rehecha, misma, "la rehecha cuenta su edad desde que se rehízo")

	require.NoError(t, sintetico.retirarLaBase())
	assert.NoDirExists(t, rehecha)
	require.NoError(t, sintetico.retirarLaBase(), "retirarla otra vez no es un error")
}

// evalsDe son las evals del directorio, que están todas bien formadas.
func evalsDe(t *testing.T, dir string) []Eval {
	t.Helper()

	conjunto, err := LeerConjunto(dir)
	require.NoError(t, err)
	require.Empty(t, conjunto.MalFormados)

	return conjunto.Evals
}

// TestReconstruccionEnElDirectorioTemporal fija lo que la reconstrucción deja en
// el directorio temporal y lo que hace sin él (contracts/medida-del-juez.md §5
// de H24): al terminar de resolver no queda nada —ni la base, ni la caché ni el
// grafo de ninguna sesión, que se descartan—, y si el directorio de la base no
// se puede crear, el error lo dice. No es paralelo, porque t.Setenv cambia el
// entorno de todo el proceso.
func TestReconstruccionEnElDirectorioTemporal(t *testing.T) {
	sintetico := reconstructorSintetico(t, elInformeSintetico(), evalsDelMundoSintetico())
	casos := []CasoEtiquetado{casoDe(sesionDeOrdenes, "a21"), casoDe(sesionDeLaRetirada, "")}

	temporal := t.TempDir()
	noEsUnDirectorio := filepath.Join(t.TempDir(), "tmp")
	require.NoError(t, os.WriteFile(noEsUnDirectorio, nil, 0o600))

	t.Setenv("TMPDIR", noEsUnDirectorio)

	resueltos, err := sintetico.resolver(casos)

	require.ErrorIs(t, err, syscall.ENOTDIR)
	require.ErrorContains(t, err, "el caso "+casos[0].nombre()+": el directorio temporal de la base no se puede crear")
	assert.Nil(t, resueltos)

	t.Setenv("TMPDIR", temporal)

	resueltos, err = sintetico.resolver(casos)

	require.NoError(t, err)
	assert.Len(t, resueltos, len(casos))

	entradas, err := os.ReadDir(temporal)
	require.NoError(t, err)
	assert.Empty(t, entradas, "la base y el directorio de cada sesión se retiran")
}

// preguntaLeidaAparte es la pregunta del fichero de eval de ese nombre, leída
// por su cuenta, sin el lector del paquete: el de las evals de boe-legislacion
// del repositorio o, si no está en ellas, el de las retiradas.
func preguntaLeidaAparte(t *testing.T, eval string) string {
	t.Helper()

	for _, carpeta := range []string{evalsDelRepositorio, EvalsRetiradas} {
		contenido, err := leerFichero(filepath.Join(carpeta, eval))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}

		require.NoError(t, err)

		var leida struct {
			Pregunta string `yaml:"pregunta"`
		}
		require.NoError(t, yaml.Unmarshal(contenido, &leida), "%s de %s es YAML", eval, carpeta)
		require.NotEmpty(t, leida.Pregunta, "%s de %s tiene pregunta", eval, carpeta)

		return leida.Pregunta
	}

	require.FailNow(t, "la eval no está", "%s no está ni en %s ni en %s", eval, evalsDelRepositorio, EvalsRetiradas)

	return ""
}

// Los recuentos de los casos etiquetados de la copia del repositorio (research
// M2 de H24; FR-024): cuántos son, cuántos de cada etiqueta y de cada
// procedencia, cuántos informes nombran, cuántos son de la eval que H7.2 retiró
// y cuántos derivados quitan su texto del de un boe articulos.
const (
	casosDeLaCopia         = 259
	defectosDeLaCopia      = 212
	correctosDeLaCopia     = 47
	deLaBitacoraEnLaCopia  = 5
	derivadosDeLaCopia     = 140
	deLaLecturaEnLaCopia   = 114
	informesDeLaCopia      = 6
	casosDeLaEvalRetirada  = 2
	derivadosDeUnArticulos = 2
)

// Lo que el control de derivaciones lee por su cuenta de los casos y de las
// órdenes de sus textos: las tres procedencias de un caso, el applet y los dos
// verbos con los que una orden lee bloques y las dos herramientas con las que
// los lee una llamada.
const (
	procedenciaDeLaBitacora = "bitacora"
	procedenciaDeUnDerivado = "derivado"
	procedenciaDeLaLectura  = "lectura"

	appletQueLeeBloques   = "boe"
	verboQueLeeUnBloque   = "articulo"
	verboQueLeeVarios     = "articulos"
	herramientaDeUnBloque = "boe_articulo"
	herramientaDeVarios   = "boe_articulos"
)

// formaDeUnInformeVersionado es la de la ruta, desde la raíz del repositorio,
// del informe del job de evals de boe-legislacion que versiona el cierre de un
// hito.
var formaDeUnInformeVersionado = regexp.MustCompile(`^specs/[0-9]{3}-[a-z0-9-]+/gates/evals/boe-legislacion\.json$`)

// reconstructorDelRepositorio es el reconstructor de los casos etiquetados de
// boe-legislacion con lo del repositorio. Es uno por proceso de test y lo
// comparten los tests que lo usan: recuerda lo que ya ha reconstruido de cada
// sesión, así que la reconstrucción de los casos se hace una sola vez.
var reconstructorDelRepositorio = sync.OnceValue(func() *reconstructor {
	return nuevoReconstructor(evalsDelRepositorio)
})

// TestGrabacionesDerivadas es el control de umbral de FR-109 y SC-009 de H24
// (contracts/medida-del-juez.md §6 y §8; FR-024), con los casos etiquetados de
// la copia del repositorio, que son los 259 de research M2:
//
//   - cada caso nombra un informe versionado y una sesión que está en él, y su
//     respuesta es, byte a byte, la de esa sesión leída aparte;
//   - la pregunta de cada caso es la de su eval, la de las evals de la skill o,
//     en los dos casos de la eval que H7.2 retiró, la de las retiradas;
//   - cada uno de los 140 derivados se diferencia de su sesión solo en el texto
//     quitado: sus textos son los de la sesión, en su orden, menos los del
//     bloque quitado; en el de un boe articulos, los mismos valores JSON menos
//     esos elementos; y se quita al menos uno.
//
// Nada está escrito a mano: los textos salen de repetir en proceso, sin red, sin
// modelo y sin el binario instalado, las invocaciones de cada sesión. Los
// informes versionados solo se leen.
func TestGrabacionesDerivadas(t *testing.T) {
	t.Parallel()

	leidos, err := leerCasosEtiquetados(juezDe(t, evalsDelRepositorio).Casos)
	require.NoError(t, err)

	casos := leidos.Casos
	exigirLosRecuentosDeLaCopia(t, leidos)

	resueltos, err := reconstructorDelRepositorio().resolver(casos)
	require.NoError(t, err)
	require.Len(t, resueltos, len(casos))

	for posicion, resuelto := range resueltos {
		resuelto.Pregunta, resuelto.Respuesta, resuelto.Textos = "", "", nil
		require.Equal(t, casos[posicion], resuelto, "el caso %d resuelto es el mismo caso", posicion+1)
	}

	sesiones := sesionesLeidasAparte(t, casos)

	t.Run("respuestas", func(t *testing.T) {
		t.Parallel()

		for _, caso := range resueltos {
			sesion, esta := sesiones[caso.Informe][caso.Sesion]
			if !assert.True(t, esta, "%s: su informe tiene su sesión", caso.nombre()) {
				continue
			}

			assert.NotEmpty(t, caso.Respuesta, "%s: su respuesta no está vacía", caso.nombre())
			assert.Equal(t, sesion.Respuesta, caso.Respuesta, "%s: su respuesta es la de su sesión, byte a byte", caso.nombre())
		}
	})

	t.Run("preguntas", func(t *testing.T) {
		t.Parallel()

		deLaRetirada := 0

		for _, caso := range resueltos {
			eval := sesiones[caso.Informe][caso.Sesion].Eval
			assert.Equal(t, preguntaLeidaAparte(t, eval), caso.Pregunta, "%s: su pregunta es la de %s", caso.nombre(), eval)

			if eval == evalRetiradaDelRepositorio {
				deLaRetirada++
			}
		}

		assert.Equal(t, casosDeLaEvalRetirada, deLaRetirada, "los casos de la eval que H7.2 retiró (research M2 de H24)")
	})

	t.Run("derivados", func(t *testing.T) {
		t.Parallel()

		exigirLosDerivados(t, resueltos)
	})
}

// exigirLosRecuentosDeLaCopia exige que los casos leídos sean los de research M2
// de H24: de la clase que decide, 259, con 212 defectos y 47 correctos, y 5 de
// la bitácora, 140 derivados y 114 de la lectura; que nombren seis informes,
// todos con la forma de uno versionado; y que los derivados sean, exactamente,
// los que llevan el bloque que se quita. Son la premisa del control: con otros
// casos, lo que comprueba sería otra cosa.
func exigirLosRecuentosDeLaCopia(t *testing.T, leidos CasosEtiquetados) {
	t.Helper()

	require.Equal(t, claseQueDecideEnElRepositorio, leidos.Clase)
	require.Len(t, leidos.Casos, casosDeLaCopia)

	porEtiqueta, porProcedencia, informes := map[string]int{}, map[string]int{}, map[string]bool{}

	for _, caso := range leidos.Casos {
		porEtiqueta[caso.Etiqueta]++
		porProcedencia[caso.Procedencia]++
		informes[caso.Informe] = true

		require.Regexp(t, formaDeUnInformeVersionado, caso.Informe, "%s nombra un informe versionado", caso.nombre())
		require.Equal(t, caso.Procedencia == procedenciaDeUnDerivado, caso.Quitado != nil,
			"%s: un derivado, y solo un derivado, lleva el bloque que se quita", caso.nombre())
	}

	require.Equal(t, map[string]int{etiquetaDefecto: defectosDeLaCopia, etiquetaCorrecto: correctosDeLaCopia}, porEtiqueta)
	require.Equal(t, map[string]int{
		procedenciaDeLaBitacora: deLaBitacoraEnLaCopia,
		procedenciaDeUnDerivado: derivadosDeLaCopia,
		procedenciaDeLaLectura:  deLaLecturaEnLaCopia,
	}, porProcedencia)
	require.Len(t, informes, informesDeLaCopia)
}

// sesionLeidaAparte es lo que el control lee por su cuenta de cada sesión de un
// informe versionado: su nombre, el fichero de su eval y su respuesta.
type sesionLeidaAparte struct {
	Sesion    string `json:"sesion"`
	Eval      string `json:"eval"`
	Respuesta string `json:"respuesta"`
}

// sesionesLeidasAparte son las sesiones de cada informe que nombran los casos,
// por la ruta del informe y por su nombre, leídas del fichero versionado sin la
// reconstrucción. Los informes solo se leen.
func sesionesLeidasAparte(t *testing.T, casos []CasoEtiquetado) map[string]map[string]sesionLeidaAparte {
	t.Helper()

	sesiones := map[string]map[string]sesionLeidaAparte{}

	for _, caso := range casos {
		if _, leido := sesiones[caso.Informe]; leido {
			continue
		}

		var informe struct {
			Evals []sesionLeidaAparte `json:"evals"`
		}

		ruta := filepath.Join(raizDelRepositorio, filepath.FromSlash(caso.Informe))
		require.NoError(t, json.Unmarshal(contenidoDelFichero(t, ruta), &informe), "%s es un informe del job de evals", ruta)

		sesiones[caso.Informe] = map[string]sesionLeidaAparte{}
		for _, sesion := range informe.Evals {
			sesiones[caso.Informe][sesion.Sesion] = sesion
		}
	}

	return sesiones
}

// exigirLosDerivados exige de cada derivado de los casos resueltos que solo se
// diferencie de su sesión en el texto quitado (exigirSoloElTextoQuitado). Los
// textos de la sesión son los del mismo caso sin el bloque que se quita, que el
// reconstructor del repositorio ya recuerda.
func exigirLosDerivados(t *testing.T, resueltos []CasoEtiquetado) {
	t.Helper()

	var derivados, sinQuitar []CasoEtiquetado

	for _, caso := range resueltos {
		if caso.Quitado == nil {
			continue
		}

		derivados = append(derivados, caso)

		caso.Quitado, caso.Textos = nil, nil
		sinQuitar = append(sinQuitar, caso)
	}

	require.Len(t, derivados, derivadosDeLaCopia)

	deSuSesion, err := reconstructorDelRepositorio().resolver(sinQuitar)
	require.NoError(t, err)
	require.Len(t, deSuSesion, len(derivados))

	deUnArticulos := 0

	for posicion, derivado := range derivados {
		if exigirSoloElTextoQuitado(t, derivado, deSuSesion[posicion].Textos) {
			deUnArticulos++
		}
	}

	assert.Equal(t, derivadosDeUnArticulos, deUnArticulos,
		"los derivados que quitan su texto del de un boe articulos (research M2 de H24)")
}

// exigirSoloElTextoQuitado exige que los textos del derivado sean los de su
// sesión, en su orden, menos los del bloque quitado, y que se quite al menos
// uno. Lo decide por su cuenta, con lo que cada texto dice (lecturaDelBloque):
// el de un boe articulo del bloque quitado no está en el derivado; el de un boe
// articulos que lo pidió está sin los elementos de ese bloque, o no está si no
// le queda ninguno; y cualquier otro está tal cual. Devuelve si algo de lo
// quitado es del texto de un boe articulos.
func exigirSoloElTextoQuitado(t *testing.T, derivado CasoEtiquetado, deLaSesion []Texto) bool {
	t.Helper()

	quitado := *derivado.Quitado
	quedan := derivado.Textos
	quitados, deUnArticulos := 0, false

	for _, texto := range deLaSesion {
		verbo, esSuLectura := lecturaDelBloque(t, texto, quitado)

		switch {
		case !esSuLectura:
			require.NotEmpty(t, quedan, "%s: le falta el texto de «%s», que no es del bloque quitado",
				derivado.nombre(), texto.Orden)
			require.Equal(t, texto, quedan[0], "%s: el texto de «%s», que no es del bloque quitado, está tal cual",
				derivado.nombre(), texto.Orden)

			quedan = quedan[1:]
		case verbo == verboQueLeeUnBloque:
			quitados++
		default:
			deUnArticulos = true

			bloques := bloquesDelSobre(t, texto.Salida)
			if !slices.ContainsFunc(bloques, func(bloque string) bool { return bloque != quitado.Bloque }) {
				quitados += len(bloques)

				continue
			}

			require.NotEmpty(t, quedan, "%s: le falta el texto de «%s», al que le quedan bloques", derivado.nombre(), texto.Orden)

			quitados += exigirElSobreSinSusElementos(t, texto, quedan[0], quitado.Bloque)
			quedan = quedan[1:]
		}
	}

	assert.Empty(t, quedan, "%s: no tiene ningún texto que no sea de su sesión", derivado.nombre())
	assert.Positive(t, quitados, "%s: se quita al menos un texto", derivado.nombre())

	return deUnArticulos
}

// lecturaDelBloque dice si el texto es el de una lectura de ese bloque que
// terminó con 0, y con qué verbo. Lo lee de la orden, que en una llamada nombra
// la herramienta y no el applet y el verbo, y del ok del sobre de su salida, que
// es verdadero si y solo si la invocación terminó con 0.
func lecturaDelBloque(t *testing.T, texto Texto, bloque BloqueQuitado) (verbo string, esSuLectura bool) {
	t.Helper()

	palabras := strings.Fields(texto.Orden)
	if len(palabras) == 0 {
		return "", false
	}

	switch palabras[0] {
	case herramientaDeUnBloque:
		palabras = slices.Concat([]string{appletQueLeeBloques, verboQueLeeUnBloque}, palabras[1:])
	case herramientaDeVarios:
		palabras = slices.Concat([]string{appletQueLeeBloques, verboQueLeeVarios}, palabras[1:])
	}

	if len(palabras) < 4 || palabras[0] != appletQueLeeBloques || palabras[2] != bloque.Norma ||
		!slices.Contains(palabras[3:], bloque.Bloque) {
		return "", false
	}

	if palabras[1] != verboQueLeeUnBloque && palabras[1] != verboQueLeeVarios {
		return "", false
	}

	return palabras[1], leerSobre(t, texto.Salida).Ok
}
