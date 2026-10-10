package evals

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
	"unicode"

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

// claseQueDecideEnElRepositorio es la clase que decide del juez de cada skill de
// la tabla de las copias, boe-legislacion y jurisprudencia, que es la de su
// medida (FR-032 de H24; FR-002 y FR-022 de H25).
const claseQueDecideEnElRepositorio = "afirma_lo_no_leido"

// Los casos de la medida versionada del juez de jurisprudencia, escritos a mano:
// de sus 249, los etiquetados como defecto y los etiquetados como correctos, que
// son los totales de sus dos umbrales (contracts/juez-de-jurisprudencia.md §5 de
// H25; FR-022 de H25).
const (
	defectosDeJurisprudencia  = 125
	correctosDeJurisprudencia = 124
)

// TestLeerMedidaDelJuez fija la lectura de la medida del juez de su contenido
// (contracts/medida-del-juez.md §1 de H24; FR-040): se leen diez claves, cada
// una a su campo, y las demás no se miran; y la que no tiene alguna de las diez,
// o la tiene con null, o no es un documento JSON con esa forma, no se lee, con
// un error que dice qué claves faltan o por qué no es ese documento. Una clave
// repetida es un error, nunca la última que gana.
func TestLeerMedidaDelJuez(t *testing.T) {
	t.Parallel()

	t.Run("entera", func(t *testing.T) {
		t.Parallel()

		medida, err := leerMedidaDelJuez([]byte(medidaSintetica))
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

			_, err := leerMedidaDelJuez(caso.contenido)
			require.EqualError(t, err, faltanClavesDeLaMedida+strings.Join(caso.faltan, ", "))
		})
	}

	for _, caso := range medidasQueNoSonJSON(t) {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			_, err := leerMedidaDelJuez(caso.contenido)
			require.Error(t, err)
			assert.True(t, strings.HasPrefix(err.Error(), medidaQueNoEsJSON), err.Error())
		})
	}
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
// a la vez, una línea por lo que falla y en el orden del contrato; y la medida
// sin una clave y la que no es JSON, que no corresponden. Las copias del
// repositorio solo se leen (FR-045).
//
// Desde H25 (contracts/juez-de-jurisprudencia.md §7 y §9 de H25; research V13 de
// H25; FR-031, FR-104; SC-004), lo ve fallar con la carpeta del juez de cada
// fila de la tabla de las copias, y no solo con la de la primera: la de
// jurisprudencia da, con cada una de las seis, su línea.
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
	require.Len(t, mutaciones, 6, "las seis mutaciones de SC-005 de H24 y de SC-004 de H25")

	for _, pareja := range copiasDelJuez {
		t.Run(pareja.skill(), func(t *testing.T) {
			t.Parallel()

			for _, mutacion := range slices.Concat(mutaciones, otrasMedidasSinCorresponder(modelo, version)) {
				t.Run(mutacion.nombre, func(t *testing.T) {
					t.Parallel()

					probarMutacionDeLaMedida(t, pareja, modelo, version, mutacion)
				})
			}
		})
	}

	t.Run("clases-que-deciden", func(t *testing.T) {
		t.Parallel()

		probarLasClasesQueDeciden(t, modelo, version)
	})
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

// probarMutacionDeLaMedida copia la carpeta del juez de la pareja, exige que la
// copia sin tocar no dé ninguna línea con el modelo y la versión que fija la
// definición del job, hace los cambios del caso, lee el juez de la copia y exige
// que la comprobación, con el modelo y la versión del caso, dé exactamente sus
// líneas.
func probarMutacionDeLaMedida(
	t *testing.T, pareja parejaDeCarpetas, modelo, version string, mutacion mutacionDeLaMedida,
) {
	t.Helper()

	evals := copiarLaCarpetaDelJuezDe(t, pareja)
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

// copiarLaCarpetaDelJuez es copiarLaCarpetaDelJuezDe con la primera pareja de la
// tabla de las copias, la de boe-legislacion.
func copiarLaCarpetaDelJuez(t *testing.T) string {
	t.Helper()

	return copiarLaCarpetaDelJuezDe(t, copiasDelJuez[0])
}

// copiarLaCarpetaDelJuezDe copia en un directorio temporal del test, dentro de
// su carpeta juez, todos los ficheros de la carpeta del juez de esa pareja de la
// tabla de las copias, y devuelve ese directorio, que es el de las evals de la
// copia. La carpeta del repositorio solo se lee (FR-045 de H24).
func copiarLaCarpetaDelJuezDe(t *testing.T, pareja parejaDeCarpetas) string {
	t.Helper()

	origen := filepath.Join(raizDelRepositorio, filepath.FromSlash(pareja.copias))

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

// skill es la skill de la pareja: la carpeta de evals que tiene la de su juez.
func (p parejaDeCarpetas) skill() string {
	return path.Base(path.Dir(p.copias))
}

// copiasDelJuez es la tabla de TestCopiasDelJuez: la carpeta del juez de cada
// skill que lo tiene, con la de la evidencia que la validó
// (contracts/medida-del-juez.md §3 de H24; FR-023). Una skill con juez que no
// esté aquí hace fallar el test. Desde H25 lleva la fila de jurisprudencia
// (contracts/juez-de-jurisprudencia.md §1 de H25; FR-001, FR-102).
var copiasDelJuez = []parejaDeCarpetas{
	{copias: "evals/boe-legislacion/juez", origen: "evidencias/adr-0037"},
	{copias: "evals/jurisprudencia/juez", origen: "evidencias/adr-0037-jurisprudencia"},
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
// Sobre una copia de las carpetas de la tabla en t.TempDir(), lo ve fallar: con
// un byte cambiado en cada una de las cuatro copias, y sin cada una de ellas, el
// único defecto es el de ese fichero, cuatro de cuatro; y la carpeta del juez de
// una skill que no está en la tabla da el suyo.
//
// Desde H25 (contracts/juez-de-jurisprudencia.md §1 y §9 de H25; FR-001, FR-102;
// SC-002), la tabla lleva la fila de jurisprudencia, y lo ve fallar con las
// cuatro copias de cada fila: el defecto nombra el fichero de la carpeta del
// juez de esa skill y el de su evidencia.
func TestCopiasDelJuez(t *testing.T) {
	t.Parallel()

	t.Run("del-repositorio", func(t *testing.T) {
		t.Parallel()

		defectos := defectosDeLasCopias(t, raizDelRepositorio, copiasDelJuez)
		assert.Empty(t, defectos, "copias del juez que no son las de su original:\n%s", strings.Join(defectos, "\n"))
	})

	for _, pareja := range copiasDelJuez {
		t.Run(pareja.skill(), func(t *testing.T) {
			t.Parallel()

			probarLasCopiasDe(t, pareja)
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

// probarLasCopiasDe ve fallar la comparación con cada una de las cuatro copias
// de la carpeta del juez de esa pareja: con un byte cambiado y sin ella.
func probarLasCopiasDe(t *testing.T, pareja parejaDeCarpetas) {
	t.Helper()

	for _, fichero := range ficherosCopiadosDelJuez {
		t.Run("cambiada-"+fichero, func(t *testing.T) {
			t.Parallel()

			probarCopiaCambiada(t, pareja, fichero)
		})

		t.Run("sin-"+fichero, func(t *testing.T) {
			t.Parallel()

			probarCopiaQueFalta(t, pareja, fichero)
		})
	}
}

// probarCopiaCambiada cambia un byte, el último, de la copia de ese fichero de
// la pareja en una copia de las carpetas de la tabla, que antes no tenía ningún
// defecto, y exige el defecto que la nombra y ningún otro.
func probarCopiaCambiada(t *testing.T, pareja parejaDeCarpetas, fichero string) {
	t.Helper()

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

// probarCopiaQueFalta quita la copia de ese fichero de la pareja de una copia de
// las carpetas de la tabla y exige el defecto que la nombra y ningún otro.
func probarCopiaQueFalta(t *testing.T, pareja parejaDeCarpetas, fichero string) {
	t.Helper()

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

// casosDerivadosPorTexto son los casos etiquetados de una skill cuyos derivados
// quitan una parte del texto pegado en la pregunta
// (contracts/medida-y-casos.md §1 de H25): el caso leído y su derivado, que
// lleva quitado con texto y, además, regla, que dice con qué cuenta se etiquetó
// y no se lee.
const casosDerivadosPorTexto = `clase: una_clase
casos:
  - informe: informes/uno.json
    sesion: una-sesion-01
    grupo: medida
    etiqueta: correcto
    procedencia: lectura
  - informe: informes/uno.json
    sesion: una-sesion-01
    quitado: {texto: fallo}
    grupo: medida
    etiqueta: defecto
    procedencia: derivado
    regla: cuenta-el-fallo
`

// TestLeerCasosEtiquetados fija la lectura de los casos etiquetados del juez
// (contracts/medida-del-juez.md §4 de H24; data-model §4; FR-024): la clase y
// cada caso con su informe, su sesión, su grupo, su etiqueta, su procedencia y,
// según el caso, el bloque que se quita o la frase, sin nada resuelto. Se leen
// con el lector común de YAML —una clave repetida es un error con sus dos
// líneas— y sin esquema publicado; lo único que la lectura exige de un caso es
// que su etiqueta sea una de las dos, porque con otra no se podría contar.
//
// El derivado de una skill como jurisprudencia quita una parte del texto pegado
// en la pregunta: se lee con su texto, y su nombre en un error lo lleva; la
// clave regla del fichero no se lee (contracts/medida-y-casos.md §1 y
// data-model §3 de H25; research D3).
func TestLeerCasosEtiquetados(t *testing.T) {
	t.Parallel()

	t.Run("leidos", func(t *testing.T) {
		t.Parallel()

		leidos, err := leerCasosEtiquetados([]byte(casosSinteticos))
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
					Quitado:  &Quitado{Norma: "BOE-A-2015-10565", Bloque: "a21"},
					Etiqueta: "defecto", Procedencia: "derivado",
				},
				{
					Informe: "informes/dos.json", Sesion: "otra-sesion-02", Grupo: "medida",
					Etiqueta: "correcto", Procedencia: "bitacora",
				},
			},
		}, leidos)
	})

	t.Run("derivado-por-texto", func(t *testing.T) {
		t.Parallel()

		leidos, err := leerCasosEtiquetados([]byte(casosDerivadosPorTexto))
		require.NoError(t, err)

		assert.Equal(t, CasosEtiquetados{
			Clase: "una_clase",
			Casos: []CasoEtiquetado{
				{
					Informe: "informes/uno.json", Sesion: "una-sesion-01", Grupo: "medida",
					Etiqueta: "correcto", Procedencia: "lectura",
				},
				{
					Informe: "informes/uno.json", Sesion: "una-sesion-01", Grupo: "medida",
					Quitado:  &Quitado{Texto: "fallo"},
					Etiqueta: "defecto", Procedencia: "derivado",
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
		{
			nombre:    "derivado-por-texto-sin-etiqueta",
			contenido: strings.Replace(casosDerivadosPorTexto, "    etiqueta: defecto\n", "", 1),
			dice:      `el caso 2 (informes/uno.json una-sesion-01 sin fallo) tiene la etiqueta ""`,
		},
	}

	for _, ilegible := range ilegibles {
		t.Run(ilegible.nombre, func(t *testing.T) {
			t.Parallel()

			leidos, err := leerCasosEtiquetados([]byte(ilegible.contenido))

			require.ErrorContains(t, err, ilegible.dice)
			assert.Zero(t, leidos, "sin los casos enteros no hay ninguno")
		})
	}
}

// TestArgumentosDeLaInvocacion fija con qué argumentos repite la reconstrucción
// cada invocación de una sesión (contracts/medida-del-juez.md §5, paso 3, de
// H24): las palabras de su orden; en una llamada a una herramienta,
// <applet>_<verbo> pasa a <applet> <verbo> y gana --json, que es lo que hace
// que su texto sea el sobre; y todas ganan --offline, también la que ya lo
// llevaba.
//
// La orden del applet cita, que empieza por «cita » o por «cita_», tiene su
// regla (contracts/medida-y-casos.md §4 de H25; research D5; FR-041): se parte
// en cada blanco seguido de -- y una letra minúscula. Lo de delante da el
// applet, el verbo y la referencia que va sin bandera; cada trozo de detrás,
// una bandera con su nombre, hasta el primer = o blanco, y su valor, que puede
// llevar espacios y va sin blancos en los extremos, salvo el de --documento,
// que va tal cual, con sus saltos de línea. --json va una sola vez, al final,
// lo llevara o no, y no gana --offline: el applet no pide nada a la red.
func TestArgumentosDeLaInvocacion(t *testing.T) {
	t.Parallel()

	// Las dos primeras líneas de la ficha de un documento del CENDOJ, como las
	// lleva la orden de una llamada del informe.
	const dosLineasDeLaFicha = "Roj: STS 3144/2023 - ECLI:ES:TS:2023:3144\nId Cendoj: 28079110012023101073"

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
		{
			nombre:     "cita-con-un-valor-con-espacios",
			invocacion: invocacionDelInforme{Orden: "cita cotejar --roj STS 1088/2023 --json"},
			argumentos: []string{"cita", "cotejar", "--roj", "STS 1088/2023", "--json"},
		},
		{
			nombre: "llamada-de-cita-con-el-documento",
			invocacion: invocacionDelInforme{
				Orden: "cita_cotejar --roj=STS 1088/2023 --documento=" + dosLineasDeLaFicha, Llamada: true,
			},
			argumentos: []string{"cita", "cotejar", "--roj", "STS 1088/2023", "--documento", dosLineasDeLaFicha, "--json"},
		},
		{
			nombre:     "cita-con-la-referencia-sin-bandera",
			invocacion: invocacionDelInforme{Orden: "cita preparar ECLI:ES:TS:2023:3144 --json"},
			argumentos: []string{"cita", "preparar", "ECLI:ES:TS:2023:3144", "--json"},
		},
		{
			nombre:     "llamada-de-cita-con-la-referencia-sin-bandera",
			invocacion: invocacionDelInforme{Orden: "cita_preparar ECLI:ES:TS:2023:3144", Llamada: true},
			argumentos: []string{"cita", "preparar", "ECLI:ES:TS:2023:3144", "--json"},
		},
		{
			nombre:     "cita-con-dos-banderas",
			invocacion: invocacionDelInforme{Orden: "cita preparar --resolucion 1088/2023 --fecha 2023-07-04 --json"},
			argumentos: []string{"cita", "preparar", "--resolucion", "1088/2023", "--fecha", "2023-07-04", "--json"},
		},
		{
			nombre:     "cita-con-json-delante-y-detras",
			invocacion: invocacionDelInforme{Orden: "cita preparar --json --texto  cláusula suelo \t--json"},
			argumentos: []string{"cita", "preparar", "--texto", "cláusula suelo", "--json"},
		},
		{
			nombre:     "llamada-de-cita-sin-banderas",
			invocacion: invocacionDelInforme{Orden: "cita_cotejar", Llamada: true},
			argumentos: []string{"cita", "cotejar", "--json"},
		},
		{
			nombre: "cita-con-el-documento-tal-cual",
			invocacion: invocacionDelInforme{
				Orden: "cita_cotejar --documento= " + dosLineasDeLaFicha + "\n\n --roj= STS 1088/2023 ", Llamada: true,
			},
			argumentos: []string{
				"cita", "cotejar", "--documento", " " + dosLineasDeLaFicha + "\n\n", "--roj", "STS 1088/2023", "--json",
			},
		},
		{
			nombre:     "cita-con-un-valor-vacio",
			invocacion: invocacionDelInforme{Orden: "cita_cotejar --documento=", Llamada: true},
			argumentos: []string{"cita", "cotejar", "--documento", "", "--json"},
		},
		{
			nombre:     "cita-con-una-bandera-sin-valor",
			invocacion: invocacionDelInforme{Orden: "cita preparar --no-graph --roj STS 1088/2023"},
			argumentos: []string{"cita", "preparar", "--no-graph", "--roj", "STS 1088/2023", "--json"},
		},
		{
			nombre:     "cita-con-guiones-que-no-son-de-una-bandera",
			invocacion: invocacionDelInforme{Orden: "cita preparar --texto cláusula --Suelo -- nula --json"},
			argumentos: []string{"cita", "preparar", "--texto", "cláusula --Suelo -- nula", "--json"},
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

	quitado := Quitado{Norma: "BOE-A-2015-10565", Bloque: "a21"}

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

// entradaConElTexto es lo que escribe quien pega un texto en las preguntas
// sintéticas de los tests de la reconstrucción: lo que va delante de la línea
// en blanco y del texto pegado.
const entradaConElTexto = "Cítame esta sentencia. Este es el texto del documento que he descargado del buscador del CENDOJ:"

// textosDelFragmento son el fragmento del repositorio y lo que sale de él,
// calculado por líneas y sin las funciones del paquete: su ficha, que es lo
// anterior a su primera línea en blanco; lo anterior a la línea en blanco que
// precede a la línea «F A L L O»; y el fragmento sin la línea del apartado 2.º
// del fallo ni la línea en blanco que la sigue. Cada uno termina con un salto
// de línea.
type textosDelFragmento struct {
	entero        string
	ficha         string
	hastaElFallo  string
	sinElApartado string
}

// leerTextosDelFragmento lee el fragmento del repositorio, que solo se lee, y
// saca de él sus textos.
func leerTextosDelFragmento(t *testing.T) textosDelFragmento {
	t.Helper()

	entero := string(contenidoDelFichero(t, fragmentoDelRepositorio))
	lineas := strings.Split(entero, "\n")

	enBlanco := slices.Index(lineas, "")
	delFallo := slices.Index(lineas, "F A L L O")
	delApartado := slices.IndexFunc(lineas, func(linea string) bool { return strings.HasPrefix(linea, "2.º-") })

	require.Positive(t, enBlanco, "%s tiene una ficha y, detrás, una línea en blanco", fragmentoDelRepositorio)
	require.Greater(t, delFallo, enBlanco+1, "%s tiene la línea del fallo detrás de su ficha", fragmentoDelRepositorio)
	require.Empty(t, lineas[delFallo-1], "delante de la línea del fallo hay una en blanco")
	require.NotEmpty(t, lineas[delFallo-2], "y delante de ella, una que no lo está")
	require.Greater(t, delApartado, delFallo, "%s tiene el apartado 2.º en su fallo", fragmentoDelRepositorio)
	require.Empty(t, lineas[delApartado+1], "detrás del apartado 2.º hay una línea en blanco")

	return textosDelFragmento{
		entero:        entero,
		ficha:         strings.Join(lineas[:enBlanco], "\n") + "\n",
		hastaElFallo:  strings.Join(lineas[:delFallo-1], "\n") + "\n",
		sinElApartado: strings.Join(slices.Concat(lineas[:delApartado], lineas[delApartado+2:]), "\n"),
	}
}

// TestTextoQuitado fija el texto pegado de una pregunta y lo que un derivado
// quita de él (contracts/medida-y-casos.md §3 y data-model §5 de H25; FR-043),
// sobre el fragmento del repositorio y sobre su ficha.
//
// La pregunta con una línea en blanco seguida de «Roj:» se parte por su primera
// línea en blanco: delante, la entrada, y detrás, el texto pegado; otra no lleva
// ninguno. Con documento no queda nada; con fallo, lo anterior a la línea
// «F A L L O», sin sus saltos de línea finales y con uno; y con apartado-2, el
// texto sin el párrafo que empieza por «2.º-» ni la línea en blanco que lo
// sigue. La pregunta del derivado es la entrada y, si queda algo, una línea en
// blanco y lo que queda. Sin quitar nada, la pregunta va tal cual y queda el
// texto pegado entero.
//
// No hay recorte, y es un error, en la pregunta sin texto pegado, con fallo o
// apartado-2 sobre un texto sin la línea «F A L L O», con apartado-2 sobre uno
// sin ese párrafo y con una parte que no es ninguna de las tres.
func TestTextoQuitado(t *testing.T) {
	t.Parallel()

	const (
		sinTextoPegado = "¿existe la STS 1088/2023, de 4 de julio?\n\nLa necesito para un recurso."
		noHayTexto     = "la pregunta no lleva ningún texto pegado"
		noHayFallo     = "el texto pegado no tiene la línea «F A L L O»"
		noHayApartado  = "el texto pegado no tiene ningún párrafo que empiece por «2.º-» con una línea en blanco detrás"
	)

	fragmento := leerTextosDelFragmento(t)
	conElFragmento := entradaConElTexto + "\n\n" + fragmento.entero
	conLaFicha := entradaConElTexto + "\n\n" + fragmento.ficha

	pegados := []struct {
		nombre   string
		pregunta string
		entrada  string
		pegado   string
	}{
		{nombre: "el-fragmento", pregunta: conElFragmento, entrada: entradaConElTexto, pegado: fragmento.entero},
		{nombre: "la-ficha", pregunta: conLaFicha, entrada: entradaConElTexto, pegado: fragmento.ficha},
		{nombre: "ninguno", pregunta: sinTextoPegado, entrada: sinTextoPegado},
		{
			nombre:   "roj-sin-linea-en-blanco-delante",
			pregunta: "¿Es esta?\nRoj: STS 3144/2023\n",
			entrada:  "¿Es esta?\nRoj: STS 3144/2023\n",
		},
		{
			nombre:   "por-su-primera-linea-en-blanco",
			pregunta: "Hola.\n\nTraigo esto:\n\nRoj: STS 3144/2023\n",
			entrada:  "Hola.",
			pegado:   "Traigo esto:\n\nRoj: STS 3144/2023\n",
		},
	}

	for _, caso := range pegados {
		t.Run("texto-pegado-"+caso.nombre, func(t *testing.T) {
			t.Parallel()

			entrada, pegado := textoPegado(caso.pregunta)

			assert.Equal(t, caso.entrada, entrada)
			assert.Equal(t, caso.pegado, pegado)
		})
	}

	recortes := []struct {
		nombre   string
		pregunta string
		quitado  string
		derivada string
		queda    string
	}{
		{nombre: "el-documento", pregunta: conElFragmento, quitado: "documento", derivada: entradaConElTexto},
		{
			nombre:   "el-fallo",
			pregunta: conElFragmento, quitado: "fallo",
			derivada: entradaConElTexto + "\n\n" + fragmento.hastaElFallo, queda: fragmento.hastaElFallo,
		},
		{
			nombre:   "el-apartado-2",
			pregunta: conElFragmento, quitado: "apartado-2",
			derivada: entradaConElTexto + "\n\n" + fragmento.sinElApartado, queda: fragmento.sinElApartado,
		},
		{nombre: "el-documento-de-la-ficha", pregunta: conLaFicha, quitado: "documento", derivada: entradaConElTexto},
		{
			nombre:   "el-fallo-sin-linea-en-blanco-delante",
			pregunta: entradaConElTexto + "\n\n" + fragmento.ficha + "F A L L O\n\nSe estima.\n", quitado: "fallo",
			derivada: entradaConElTexto + "\n\n" + fragmento.ficha, queda: fragmento.ficha,
		},
		{nombre: "nada", pregunta: conElFragmento, derivada: conElFragmento, queda: fragmento.entero},
		{nombre: "nada-de-la-ficha", pregunta: conLaFicha, derivada: conLaFicha, queda: fragmento.ficha},
		{nombre: "nada-sin-texto-pegado", pregunta: sinTextoPegado, derivada: sinTextoPegado},
	}

	for _, caso := range recortes {
		t.Run("sin-"+caso.nombre, func(t *testing.T) {
			t.Parallel()

			derivada, queda, err := sinElTexto(caso.pregunta, caso.quitado)

			require.NoError(t, err)
			assert.Equal(t, caso.derivada, derivada, "la pregunta del derivado")
			assert.Equal(t, caso.queda, queda, "lo que queda del texto pegado")
		})
	}

	sinRecorte := []struct {
		nombre   string
		pregunta string
		quitado  string
		dice     string
	}{
		{nombre: "documento-sin-texto-pegado", pregunta: sinTextoPegado, quitado: "documento", dice: noHayTexto},
		{nombre: "fallo-sin-texto-pegado", pregunta: sinTextoPegado, quitado: "fallo", dice: noHayTexto},
		{nombre: "apartado-2-sin-texto-pegado", pregunta: sinTextoPegado, quitado: "apartado-2", dice: noHayTexto},
		{nombre: "fallo-sin-la-linea-del-fallo", pregunta: conLaFicha, quitado: "fallo", dice: noHayFallo},
		{nombre: "apartado-2-sin-la-linea-del-fallo", pregunta: conLaFicha, quitado: "apartado-2", dice: noHayFallo},
		{
			nombre:   "fallo-en-mitad-de-una-linea",
			pregunta: conLaFicha + "\nVisto el F A L L O de la instancia.\n\n2.º- Se confirma.\n\nFin.\n",
			quitado:  "fallo", dice: noHayFallo,
		},
		{
			nombre:   "apartado-2-sin-su-parrafo",
			pregunta: entradaConElTexto + "\n\n" + fragmento.sinElApartado, quitado: "apartado-2", dice: noHayApartado,
		},
		{
			nombre:   "apartado-2-sin-linea-en-blanco-detras",
			pregunta: conLaFicha + "\nF A L L O\n\n2.º- Se confirma.\n", quitado: "apartado-2", dice: noHayApartado,
		},
		{
			nombre:   "otra-parte",
			pregunta: conElFragmento, quitado: "antecedentes",
			dice: `quitado.texto es "antecedentes" y tiene que ser documento, fallo o apartado-2`,
		},
	}

	for _, caso := range sinRecorte {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			derivada, queda, err := sinElTexto(caso.pregunta, caso.quitado)

			require.ErrorContains(t, err, caso.dice)
			assert.Empty(t, derivada, "sin recorte no hay pregunta")
			assert.Empty(t, queda)
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
		caso.Quitado = &Quitado{Norma: "BOE-A-2015-10565", Bloque: quitado}
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
//
// Los casos de una skill como jurisprudencia —las órdenes del applet cita, el
// informe de un sondeo y los derivados que quitan una parte del texto pegado—
// los fija probarLaResolucionDeJurisprudencia, sobre otro mundo sintético.
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

	t.Run("jurisprudencia", probarLaResolucionDeJurisprudencia)
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

// El mundo sintético de jurisprudencia de TestResolverCasos
// (contracts/medida-y-casos.md §2 a §7 de H25): el informe de un job cuyas
// sesiones invocaron el applet cita, en sus dos modos; el informe de un sondeo,
// con sus preguntas junto a él y el fragmento que nombran, que el test copia
// bajo su raíz; y cinco carpetas con el informe de otro sondeo al que le falta
// algo.
const (
	informeDeCita       = "informes/jurisprudencia.json"
	informeDeUnSondeo   = "sondeos/sondeo-con-skill.json"
	preguntasDeUnSondeo = "sondeos/preguntas.json"
	fragmentoDeUnSondeo = "textos/fragmento.txt"

	carpetaSinPreguntas         = "sin-preguntas"
	carpetaDePreguntasIlegibles = "preguntas-ilegibles"
	carpetaSinFragmento         = "sin-fragmento"
	carpetaSinFicha             = "sin-ficha"
	carpetaSinSesiones          = "sin-sesiones"

	otroSondeo          = "sondeo.json"
	susPreguntas        = "preguntas.json"
	fragmentoQueNoEsta  = "textos/no-esta.txt"
	fragmentoDeUnaLinea = "textos/sin-linea-en-blanco.txt"
)

// Las sesiones del informe del job del mundo de jurisprudencia.
const (
	// sesionDelCotejoPorOrden es la de una sesión del modo orden de la eval con
	// el fragmento pegado: cotejó lo pegado, que leyó por la entrada estándar,
	// y preparó la consulta de un ROJ, cuyo valor lleva un espacio.
	sesionDelCotejoPorOrden = "04-documento-pegado-un-modelo-01"

	// sesionDelCotejoPorLlamada es la de una sesión del modo herramienta de esa
	// eval: el proceso del servidor y una llamada que cotejó la ficha, que lleva
	// en --documento con sus saltos de línea.
	sesionDelCotejoPorLlamada = "04-documento-pegado-herramienta-un-modelo-01"

	// sesionSinTextoPegado es la de una sesión de la eval sin texto pegado:
	// preparó una consulta y pidió cotejar sin tener nada que cotejar.
	sesionSinTextoPegado = "01-existe-con-numero-y-fecha-un-modelo-01"

	// sesionConOtroCodigo y sesionConOtroCodigoPorLlamada son las de dos
	// sesiones con una orden que el informe da por terminada con 0 y que,
	// repetida, termina con 2: su ROJ no tiene forma de ROJ.
	sesionConOtroCodigo           = "04-documento-pegado-un-modelo-02"
	sesionConOtroCodigoPorLlamada = "04-documento-pegado-herramienta-un-modelo-02"

	// sesionSinCodigo es la de una sesión con una orden que quedó sin código.
	sesionSinCodigo = "04-documento-pegado-un-modelo-03"
)

// Las órdenes de las sesiones del informe del job del mundo de jurisprudencia,
// como las publica un informe.
const (
	cotejarLoPegado      = "cita cotejar --json"
	prepararElROJ        = "cita preparar --roj STS 1088/2023 --json"
	prepararPorNumero    = "cita preparar --resolucion 1088/2023 --fecha 2023-07-04 --json"
	cotejarUnROJSinForma = "cita cotejar --roj 1088 --json"

	// llamarACotejar y llamarACotejarSinForma van seguidas del documento.
	llamarACotejar         = "cita_cotejar --roj=STS 3144/2023 --documento="
	llamarACotejarSinForma = "cita_cotejar --roj=1088 --documento="

	// codigoDeCotejarSinTexto es el código con el que terminó, en su sesión, la
	// orden cotejar que no recibió ningún texto: el de un error de argumentos.
	codigoDeCotejarSinTexto = 2
)

// Las sesiones del informe del sondeo del mundo de jurisprudencia, cada una con
// el id de su pregunta delante, y las de los otros sondeos.
const (
	sesionDelSondeoConEval         = "01-existe-con-numero-y-fecha-con-skill-01"
	sesionDelSondeoSinInvocaciones = "07-resumen-de-una-conocida-sin-skill-01"
	sesionDelSondeoSinSuPregunta   = "08-no-esta-entre-las-preguntas-con-skill-01"
	sesionDelSondeoConElFragmento  = "09-doctrina-con-el-fallo-delante-con-skill-01"
	sesionDelSondeoConLaFicha      = "11-de-que-trata-con-la-ficha-sola-con-skill-01"
	sesionDelSondeoConOtraEval     = "13-con-una-eval-que-no-esta-con-skill-01"
	sesionDelSondeoSinNada         = "14-sin-eval-ni-pregunta-con-skill-01"

	sesionDeOtroSondeoConElFragmento = "con-el-fragmento-01"
	sesionDeOtroSondeoConLaFicha     = "con-la-ficha-01"
	sesionDeOtroSondeoSinMarcas      = "sin-marcas-01"
)

// Las órdenes de Bash de las sesiones del sondeo y lo que devolvieron, que no
// es ningún sobre: la reconstrucción no las repite ni las lee, las copia.
const (
	prepararEnElSondeo = "kitlegal cita preparar --resolucion 1088/2023 --fecha 2023-07-04 --json"
	buscarEnElSondeo   = `kitlegal cita preparar --texto "vencimiento anticipado" --json`

	// cotejarEnElSondeo va seguida del documento y de finDelDocumento.
	cotejarEnElSondeo = "kitlegal cita cotejar --json <<'DOCUMENTO'\n"
	finDelDocumento   = "DOCUMENTO"

	salidaDePreparar = "lo que devolvió cita preparar en el sondeo\n"
	salidaDeBuscar   = "lo que devolvió la búsqueda por texto,\n  con dos líneas y sin salto final"
	salidaDeCotejar  = "lo que devolvió cita cotejar en el sondeo\n"
)

// Las preguntas del mundo de jurisprudencia: la de la eval sin texto pegado, lo
// que va delante del texto pegado en las dos plantillas del sondeo y la
// plantilla que no lleva ninguno.
const (
	preguntaSinTextoPegado  = "¿existe la STS 1088/2023, de 4 de julio?"
	entradaConElFragmento   = "¿Qué doctrina fija esta sentencia? Este es el texto del documento que he descargado:"
	entradaConLaFicha       = "¿De qué trata esta sentencia? Esta es la ficha del documento que he descargado:"
	plantillaSinTextoPegado = "resúmeme la STS 241/2013, de 9 de mayo"
)

// contenidoDeLaEvalSinTexto es el de la eval de hoy del mundo de jurisprudencia
// cuya pregunta no lleva ningún texto pegado.
const contenidoDeLaEvalSinTexto = `pregunta: "` + preguntaSinTextoPegado + `"
activa: true
comandos:
  - applet: cita
    verbo: preparar
sentencias:
  no_comprobada: true
  direcciones:
    - https://www.poderjudicial.es/search/indexAN.jsp
  casillas:
    - nombre: Nº Resolución
      valor: 1088/2023
    - nombre: Fecha resolución
      valor: 04/07/2023
  ninguna_cita: true
`

// contenidoDeLasPreguntas es el de las preguntas del sondeo del mundo de
// jurisprudencia (data-model §4 de H25), con claves que la reconstrucción no
// lee: una con eval, una sin texto pegado, una con el fragmento, una con su
// ficha, una con una eval que no es de las de hoy y una sin eval ni pregunta.
const contenidoDeLasPreguntas = `{
  "nota": "Las preguntas de un sondeo sintético.",
  "commit": "0000000000000000000000000000000000000000",
  "fragmento": "` + fragmentoDeUnSondeo + `",
  "preguntas": [
    {"id": "01-existe-con-numero-y-fecha", "eval": "` + jurisprudenciaExiste + `"},
    {"id": "07-resumen-de-una-conocida", "pregunta": "` + plantillaSinTextoPegado + `"},
    {"id": "09-doctrina-con-el-fallo-delante", "pregunta": "` + entradaConElFragmento + `\n\n{fragmento}"},
    {"id": "11-de-que-trata-con-la-ficha-sola", "pregunta": "` + entradaConLaFicha + `\n\n{ficha}"},
    {"id": "13-con-una-eval-que-no-esta", "eval": "13-no-esta.yaml"},
    {"id": "14-sin-eval-ni-pregunta"}
  ]
}
`

// preguntasDeOtroSondeo es el contenido de las preguntas de los otros sondeos
// del mundo de jurisprudencia, con la ruta de su fragmento por poner: una con
// el fragmento, una con su ficha y una sin ninguno de los dos. Y
// contenidoDelFragmentoDeUnaLinea, el de un fragmento sin línea en blanco.
const (
	preguntasDeOtroSondeo = `{"fragmento": "%s", "preguntas": [
  {"id": "con-el-fragmento", "pregunta": "Resúmeme esto:\n\n{fragmento}"},
  {"id": "con-la-ficha", "pregunta": "¿De qué trata esto?\n\n{ficha}"},
  {"id": "sin-marcas", "pregunta": "` + plantillaSinTextoPegado + `"}
]}
`
	contenidoDelFragmentoDeUnaLinea = "Roj: STS 3144/2023 - ECLI:ES:TS:2023:3144\nTipo de Resolución: Sentencia\n"
)

// contenidoDeLaEvalConElTexto es el de la eval de hoy del mundo de
// jurisprudencia cuya pregunta lleva pegado ese texto, detrás de
// entradaConElTexto y de una línea en blanco: un escalar de bloque, con cada
// línea que no está en blanco sangrada.
func contenidoDeLaEvalConElTexto(texto string) string {
	var pregunta strings.Builder

	for linea := range strings.Lines(entradaConElTexto + "\n\n" + texto) {
		if linea != "\n" {
			pregunta.WriteString("  ")
		}

		pregunta.WriteString(linea)
	}

	return "pregunta: |\n" + pregunta.String() + `activa: true
comandos:
  - applet: cita
    verbo: cotejar
sentencias:
  citas:
    - ecli: ECLI:ES:TS:2023:3144
      roj: STS 3144/2023
`
}

// objetoJSON es un objeto de un informe sintético que el test escribe con el
// codificador, y no a mano, porque lleva textos con saltos de línea.
type objetoJSON = map[string]any

// documentoJSON escribe el valor como un documento JSON.
func documentoJSON(t *testing.T, valor any) string {
	t.Helper()

	escrito, err := json.Marshal(valor)
	require.NoError(t, err)

	return string(escrito)
}

// respuestaDeLaSesion es la respuesta de esa sesión en los informes del mundo
// de jurisprudencia, con blancos que no se pueden perder.
func respuestaDeLaSesion(sesion string) string {
	return "La respuesta de " + sesion + ".\n\n  Con dos espacios delante y un salto detrás.\n"
}

// contenidoDelInformeDeCita es el del informe del job del mundo de
// jurisprudencia, con la forma de un informe versionado del job de evals y con
// claves que la reconstrucción no lee.
func contenidoDelInformeDeCita(t *testing.T, fragmento textosDelFragmento) string {
	t.Helper()

	invocacion := func(orden string, codigo any, llamada bool) objetoJSON {
		return objetoJSON{"orden": orden, "codigo": codigo, "conexiones": []string{}, "llamada": llamada}
	}

	sesion := func(nombre, eval string, invocaciones ...objetoJSON) objetoJSON {
		return objetoJSON{
			"sesion": nombre, "eval": eval, "pasa": true,
			"respuesta": respuestaDeLaSesion(nombre), "invocaciones": invocaciones,
		}
	}

	return documentoJSON(t, objetoJSON{
		"commit": "0000000000000000000000000000000000000000",
		"skill":  "jurisprudencia",
		"evals": []objetoJSON{
			sesion(sesionDelCotejoPorOrden, jurisprudenciaDocumento,
				invocacion(cotejarLoPegado, 0, false), invocacion(prepararElROJ, 0, false)),
			sesion(sesionDelCotejoPorLlamada, jurisprudenciaDocumento,
				invocacion(ordenDelProcesoDelServidor, nil, false), invocacion(llamarACotejar+fragmento.ficha, 0, true)),
			sesion(sesionSinTextoPegado, jurisprudenciaExiste,
				invocacion(prepararPorNumero, 0, false), invocacion(cotejarLoPegado, codigoDeCotejarSinTexto, false)),
			sesion(sesionConOtroCodigo, jurisprudenciaDocumento, invocacion(cotejarUnROJSinForma, 0, false)),
			sesion(sesionConOtroCodigoPorLlamada, jurisprudenciaDocumento,
				invocacion(llamarACotejarSinForma+fragmento.ficha, 0, true)),
			sesion(sesionSinCodigo, jurisprudenciaDocumento, invocacion(prepararElROJ, nil, false)),
		},
	})
}

// contenidoDelInformeDelSondeo es el del informe del sondeo del mundo de
// jurisprudencia, con la forma del informe de un sondeo de la validación del
// juez y con claves que la reconstrucción no lee. Sus invocaciones llevan lo
// que da texto —la orden de Bash que empieza por «kitlegal »— y lo que no: la
// de otra herramienta, la orden de Bash que no empieza así y la que es de otra
// herramienta aunque su entrada lleve una orden, que ni siquiera es un texto.
func contenidoDelInformeDelSondeo(t *testing.T, fragmento textosDelFragmento) string {
	t.Helper()

	deLaHerramienta := func(herramienta string, entrada objetoJSON, salida string) objetoJSON {
		return objetoJSON{"herramienta": herramienta, "entrada": entrada, "salida": salida, "error": false}
	}

	bash := func(orden, salida string) objetoJSON {
		return deLaHerramienta("Bash", objetoJSON{"command": orden, "description": "Una orden."}, salida)
	}

	sesion := func(nombre, pregunta string, invocaciones ...objetoJSON) objetoJSON {
		return objetoJSON{
			"sesion": nombre, "pregunta": pregunta, "modelo": "un-modelo", "activada": true,
			"respuesta": respuestaDeLaSesion(nombre), "invocaciones": invocaciones,
		}
	}

	laSkill := deLaHerramienta("Skill", objetoJSON{"skill": "jurisprudencia"}, "Launching skill: jurisprudencia")

	return documentoJSON(t, objetoJSON{
		"sondeo":       "con-skill",
		"modelo":       "un-modelo",
		"preguntas":    7,
		"repeticiones": 1,
		"sesiones": []objetoJSON{
			sesion(sesionDelSondeoConEval, "01-existe-con-numero-y-fecha",
				laSkill,
				bash(prepararEnElSondeo, salidaDePreparar),
				bash("cd /tmp && "+prepararEnElSondeo, "no empieza por kitlegal"),
				bash("kitlegal", "tampoco: es kitlegal sin nada detrás"),
				deLaHerramienta("Read", objetoJSON{"command": prepararEnElSondeo}, "no es de Bash"),
				deLaHerramienta("Otra", objetoJSON{"command": []string{"kitlegal ", "cita"}}, "ni es un texto"),
				bash(cotejarEnElSondeo+fragmento.ficha+finDelDocumento, "cotejar sin texto pegado en la pregunta")),
			sesion(sesionDelSondeoSinInvocaciones, "07-resumen-de-una-conocida"),
			sesion(sesionDelSondeoSinSuPregunta, "08-no-esta-entre-las-preguntas"),
			sesion(sesionDelSondeoConElFragmento, "09-doctrina-con-el-fallo-delante",
				laSkill,
				bash(cotejarEnElSondeo+fragmento.entero+finDelDocumento, salidaDeCotejar),
				bash(buscarEnElSondeo, salidaDeBuscar)),
			sesion(sesionDelSondeoConLaFicha, "11-de-que-trata-con-la-ficha-sola",
				bash(cotejarEnElSondeo+fragmento.ficha+finDelDocumento, salidaDeCotejar)),
			sesion(sesionDelSondeoConOtraEval, "13-con-una-eval-que-no-esta"),
			sesion(sesionDelSondeoSinNada, "14-sin-eval-ni-pregunta"),
		},
	})
}

// contenidoDeOtroSondeo es el del informe de los otros sondeos del mundo de
// jurisprudencia: tres sesiones sin invocaciones, una por cada pregunta de
// preguntasDeOtroSondeo.
func contenidoDeOtroSondeo(t *testing.T) string {
	t.Helper()

	sesion := func(nombre, pregunta string) objetoJSON {
		return objetoJSON{
			"sesion": nombre, "pregunta": pregunta, "respuesta": respuestaDeLaSesion(nombre), "invocaciones": []objetoJSON{},
		}
	}

	return documentoJSON(t, objetoJSON{
		"sondeo": "otro",
		"sesiones": []objetoJSON{
			sesion(sesionDeOtroSondeoConElFragmento, "con-el-fragmento"),
			sesion(sesionDeOtroSondeoConLaFicha, "con-la-ficha"),
			sesion(sesionDeOtroSondeoSinMarcas, "sin-marcas"),
		},
	})
}

// mundoDeJurisprudencia es el mundo sintético de jurisprudencia de
// TestResolverCasos: los textos del fragmento, con los que el test dice qué
// espera, y el reconstructor de ese mundo.
type mundoDeJurisprudencia struct {
	fragmento textosDelFragmento
	sintetico *reconstructor
}

// nuevoMundoDeJurisprudencia escribe el mundo de jurisprudencia bajo una raíz
// temporal del test: sus informes, las preguntas de cada sondeo junto a su
// informe, los fragmentos que nombran y sus dos evals de hoy.
func nuevoMundoDeJurisprudencia(t *testing.T) mundoDeJurisprudencia {
	t.Helper()

	fragmento := leerTextosDelFragmento(t)
	deOtroSondeo := contenidoDeOtroSondeo(t)

	ficheros := map[string]string{
		informeDeCita:       contenidoDelInformeDeCita(t, fragmento),
		informeDeUnSondeo:   contenidoDelInformeDelSondeo(t, fragmento),
		preguntasDeUnSondeo: contenidoDeLasPreguntas,
		fragmentoDeUnSondeo: fragmento.entero,
		fragmentoDeUnaLinea: contenidoDelFragmentoDeUnaLinea,

		path.Join(carpetaSinPreguntas, otroSondeo):           deOtroSondeo,
		path.Join(carpetaDePreguntasIlegibles, otroSondeo):   deOtroSondeo,
		path.Join(carpetaDePreguntasIlegibles, susPreguntas): "{",
		path.Join(carpetaSinFragmento, otroSondeo):           deOtroSondeo,
		path.Join(carpetaSinFragmento, susPreguntas):         fmt.Sprintf(preguntasDeOtroSondeo, fragmentoQueNoEsta),
		path.Join(carpetaSinFicha, otroSondeo):               deOtroSondeo,
		path.Join(carpetaSinFicha, susPreguntas):             fmt.Sprintf(preguntasDeOtroSondeo, fragmentoDeUnaLinea),
		path.Join(carpetaSinSesiones, otroSondeo):            `{"sondeo": "roto", "sesiones": "ninguna"}`,
	}

	evals := []entradaDeConjunto{
		{nombre: jurisprudenciaExiste, contenido: contenidoDeLaEvalSinTexto},
		{nombre: jurisprudenciaDocumento, contenido: contenidoDeLaEvalConElTexto(fragmento.entero)},
	}

	return mundoDeJurisprudencia{fragmento: fragmento, sintetico: reconstructorSintetico(t, ficheros, evals)}
}

// casoDeJurisprudencia es un caso etiquetado de esa sesión de ese informe del
// mundo de jurisprudencia, sin resolver: con texto, el derivado que quita esa
// parte del texto pegado en su pregunta.
func casoDeJurisprudencia(informe, sesion, texto string) CasoEtiquetado {
	caso := CasoEtiquetado{
		Informe: informe, Sesion: sesion, Grupo: "medida",
		Etiqueta: etiquetaCorrecto, Procedencia: procedenciaDeLaLectura, Frase: "una frase",
	}

	if texto != "" {
		caso.Quitado = &Quitado{Texto: texto}
		caso.Etiqueta, caso.Procedencia, caso.Frase = etiquetaDefecto, procedenciaDeUnDerivado, ""
	}

	return caso
}

// resolver resuelve esos casos del mundo, que se pueden resolver, y los
// devuelve en su orden: cada uno es el mismo caso con su pregunta, su respuesta
// —la de su sesión en su informe, tal cual— y sus textos.
func (m mundoDeJurisprudencia) resolver(t *testing.T, casos ...CasoEtiquetado) []CasoEtiquetado {
	t.Helper()

	resueltos, err := m.sintetico.resolver(casos)
	require.NoError(t, err)
	require.Len(t, resueltos, len(casos))

	for posicion, resuelto := range resueltos {
		assert.Equal(t, respuestaDeLaSesion(casos[posicion].Sesion), resuelto.Respuesta,
			"la respuesta del caso %d va tal cual, con sus blancos", posicion+1)

		resuelto.Pregunta, resuelto.Respuesta, resuelto.Textos = "", "", nil
		assert.Equal(t, casos[posicion], resuelto, "el caso %d resuelto es el mismo caso", posicion+1)
	}

	return resueltos
}

// probarLaResolucionDeJurisprudencia fija, dentro de TestResolverCasos, la
// resolución de los casos de una skill como jurisprudencia sobre su mundo
// sintético (contracts/medida-y-casos.md §2 a §7 y data-model §3 a §6 de H25;
// FR-041 a FR-044): las órdenes del applet cita de un informe del job, que se
// repiten en proceso; el informe de un sondeo, cuyas órdenes no se repiten; los
// derivados que quitan una parte del texto pegado en la pregunta, de los dos
// tipos de informe; lo reconstruido, que se recuerda por lo que se quita; y los
// casos que no se pueden resolver, cada uno con un error que lo nombra.
func probarLaResolucionDeJurisprudencia(t *testing.T) {
	t.Parallel()

	mundo := nuevoMundoDeJurisprudencia(t)

	partes := []struct {
		nombre string
		probar func(t *testing.T)
	}{
		{nombre: "ordenes-de-cita", probar: mundo.probarLasOrdenesDeCita},
		{nombre: "derivados-del-job", probar: mundo.probarLosDerivadosDelJob},
		{nombre: "informe-de-un-sondeo", probar: mundo.probarElInformeDeUnSondeo},
		{nombre: "derivados-del-sondeo", probar: mundo.probarLosDerivadosDelSondeo},
		{nombre: "otros-sondeos", probar: mundo.probarOtrosSondeos},
		{nombre: "lo-reconstruido-se-recuerda", probar: mundo.probarQueSeRecuerda},
	}

	for _, parte := range partes {
		t.Run(parte.nombre, func(t *testing.T) {
			t.Parallel()

			parte.probar(t)
		})
	}

	for _, caso := range casosDeJurisprudenciaSinResolver() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			resueltos, err := mundo.sintetico.resolver([]CasoEtiquetado{
				casoDeJurisprudencia(informeDeUnSondeo, sesionDelSondeoSinInvocaciones, ""), caso.caso,
			})

			for _, fragmento := range caso.dice {
				require.ErrorContains(t, err, fragmento)
			}

			if caso.es != nil {
				require.ErrorIs(t, err, caso.es)
			}

			assert.NotContains(t, err.Error(), "\n", "el error de un caso que no se resuelve es una línea")
			assert.Nil(t, resueltos, "con un caso sin resolver no hay ninguno")
		})
	}
}

// probarLasOrdenesDeCita fija los textos de las sesiones de un informe del job
// que invocaron el applet cita (contracts/medida-y-casos.md §4 de H25; FR-041):
// cada orden se repite en proceso y da su orden, la del informe tal cual, y su
// sobre. La orden cotejar del modo orden recibe por la entrada estándar el texto
// pegado en la pregunta, y la del modo herramienta, el documento que lleva; el
// valor de una bandera llega entero, con sus espacios; el proceso del servidor
// no da texto; y sin texto pegado en la pregunta, tampoco la orden cotejar, que
// ni se compara con el código de su informe.
func (m mundoDeJurisprudencia) probarLasOrdenesDeCita(t *testing.T) {
	t.Helper()

	resueltos := m.resolver(t,
		casoDeJurisprudencia(informeDeCita, sesionDelCotejoPorOrden, ""),
		casoDeJurisprudencia(informeDeCita, sesionDelCotejoPorLlamada, ""),
		casoDeJurisprudencia(informeDeCita, sesionSinTextoPegado, ""),
	)
	porOrden, porLlamada, sinTextoPegado := resueltos[0], resueltos[1], resueltos[2]

	conElFragmento := entradaConElTexto + "\n\n" + m.fragmento.entero
	assert.Equal(t, conElFragmento, porOrden.Pregunta, "la pregunta de su eval, con el fragmento byte a byte")
	assert.Equal(t, conElFragmento, porLlamada.Pregunta)
	assert.Equal(t, preguntaSinTextoPegado, sinTextoPegado.Pregunta)

	require.Equal(t, []string{cotejarLoPegado, prepararElROJ}, ordenesDe(porOrden.Textos))

	cotejo := exigirElCotejo(t, porOrden.Textos[0], cotejarLoPegado, m.fragmento.entero)
	assert.Equal(t, rojDelFragmento, cotejo.Ficha.ROJ)
	assert.Nil(t, cotejo.Pedida, "la orden no pidió ninguna referencia")

	consulta := exigirElSobreDeCita(t, porOrden.Textos[1])
	assert.Equal(t, referenciaLeida{Forma: "roj", Valor: rojQueNoEsElSuyo},
		datosDelSobre[consultaLeida](t, porOrden.Textos[1].Salida).Referencia, "el valor de --roj llega con su espacio")
	assert.Equal(t, direccionDelCendoj, consulta.URL)

	llamada := llamarACotejar + m.fragmento.ficha
	require.Equal(t, []string{llamada}, ordenesDe(porLlamada.Textos), "mcp serve no da ningún texto")

	cotejo = exigirElCotejo(t, porLlamada.Textos[0], llamada, m.fragmento.ficha)
	require.NotNil(t, cotejo.Pedida)
	assert.Equal(t, referenciaLeida{Forma: "roj", Valor: rojDelFragmento}, *cotejo.Pedida)
	require.NotNil(t, cotejo.EsLaPedida)
	assert.True(t, *cotejo.EsLaPedida)

	require.Equal(t, []string{prepararPorNumero}, ordenesDe(sinTextoPegado.Textos),
		"sin texto pegado, la orden cotejar no da texto")
	exigirElSobreDeCita(t, sinTextoPegado.Textos[0])
}

// probarLosDerivadosDelJob fija los derivados por texto de una sesión de un
// informe del job (contracts/medida-y-casos.md §3 y §4 de H25; FR-043): la
// pregunta de cada uno es la de su sesión sin la parte quitada. Sin el
// documento, ninguna orden cotejar da texto, lleve o no --documento, y las
// demás siguen; sin el fallo y sin su apartado 2.º, la orden cotejar que leyó
// por la entrada estándar se repite con lo que queda, y la que lleva su
// documento, con él.
func (m mundoDeJurisprudencia) probarLosDerivadosDelJob(t *testing.T) {
	t.Helper()

	resueltos := m.resolver(t,
		casoDeJurisprudencia(informeDeCita, sesionDelCotejoPorOrden, "documento"),
		casoDeJurisprudencia(informeDeCita, sesionDelCotejoPorOrden, "fallo"),
		casoDeJurisprudencia(informeDeCita, sesionDelCotejoPorOrden, "apartado-2"),
		casoDeJurisprudencia(informeDeCita, sesionDelCotejoPorLlamada, "documento"),
		casoDeJurisprudencia(informeDeCita, sesionDelCotejoPorLlamada, "fallo"),
	)
	sinElDocumento, sinElFallo, sinElApartado := resueltos[0], resueltos[1], resueltos[2]
	porLlamadaSinElDocumento, porLlamadaSinElFallo := resueltos[3], resueltos[4]

	assert.Equal(t, entradaConElTexto, sinElDocumento.Pregunta, "sin el texto pegado ni la línea en blanco")
	require.Equal(t, []string{prepararElROJ}, ordenesDe(sinElDocumento.Textos), "ninguna orden cotejar")
	exigirElSobreDeCita(t, sinElDocumento.Textos[0])

	assert.Equal(t, entradaConElTexto, porLlamadaSinElDocumento.Pregunta)
	assert.Empty(t, porLlamadaSinElDocumento.Textos, "tampoco la orden cotejar que lleva su documento")

	quedan := []struct {
		derivado CasoEtiquetado
		queda    string
	}{
		{derivado: sinElFallo, queda: m.fragmento.hastaElFallo},
		{derivado: sinElApartado, queda: m.fragmento.sinElApartado},
	}

	for _, caso := range quedan {
		assert.Equal(t, entradaConElTexto+"\n\n"+caso.queda, caso.derivado.Pregunta,
			"la pregunta de %s", caso.derivado.Quitado.Texto)
		require.Equal(t, []string{cotejarLoPegado, prepararElROJ}, ordenesDe(caso.derivado.Textos))
		exigirElCotejo(t, caso.derivado.Textos[0], cotejarLoPegado, caso.queda)
	}

	assert.Equal(t, entradaConElTexto+"\n\n"+m.fragmento.hastaElFallo, porLlamadaSinElFallo.Pregunta)
	require.Len(t, porLlamadaSinElFallo.Textos, 1)
	exigirElCotejo(t, porLlamadaSinElFallo.Textos[0], llamarACotejar+m.fragmento.ficha, m.fragmento.ficha)
}

// probarElInformeDeUnSondeo fija la resolución de los casos del informe de un
// sondeo (contracts/medida-y-casos.md §2 y §5 de H25; FR-042): la pregunta es
// la de sus preguntas —la de la eval de hoy que nombran, o su plantilla con el
// fragmento, byte a byte, o con su ficha—, y los textos, sin repetir ninguna
// orden, los de cada orden de Bash que empieza por «kitlegal », con su salida
// tal cual y en su orden. Una sesión sin invocaciones no tiene textos, y sin
// texto pegado en la pregunta, la orden cita cotejar no da ninguno.
func (m mundoDeJurisprudencia) probarElInformeDeUnSondeo(t *testing.T) {
	t.Helper()

	resueltos := m.resolver(t,
		casoDeJurisprudencia(informeDeUnSondeo, sesionDelSondeoConEval, ""),
		casoDeJurisprudencia(informeDeUnSondeo, sesionDelSondeoConElFragmento, ""),
		casoDeJurisprudencia(informeDeUnSondeo, sesionDelSondeoConLaFicha, ""),
		casoDeJurisprudencia(informeDeUnSondeo, sesionDelSondeoSinInvocaciones, ""),
	)
	conEval, conElFragmento, conLaFicha, sinInvocaciones := resueltos[0], resueltos[1], resueltos[2], resueltos[3]

	assert.Equal(t, preguntaSinTextoPegado, conEval.Pregunta, "la de la eval de hoy que nombra su pregunta")
	assert.Equal(t, []Texto{{Orden: prepararEnElSondeo, Salida: salidaDePreparar}}, conEval.Textos,
		"solo la orden de Bash que empieza por «kitlegal », y sin la de cotejar: su pregunta no lleva texto pegado")

	assert.Equal(t, entradaConElFragmento+"\n\n"+m.fragmento.entero, conElFragmento.Pregunta)
	assert.Equal(t, []Texto{
		{Orden: cotejarEnElSondeo + m.fragmento.entero + finDelDocumento, Salida: salidaDeCotejar},
		{Orden: buscarEnElSondeo, Salida: salidaDeBuscar},
	}, conElFragmento.Textos)

	assert.Equal(t, entradaConLaFicha+"\n\n"+m.fragmento.ficha, conLaFicha.Pregunta)
	assert.Equal(t, []Texto{
		{Orden: cotejarEnElSondeo + m.fragmento.ficha + finDelDocumento, Salida: salidaDeCotejar},
	}, conLaFicha.Textos)

	assert.Equal(t, plantillaSinTextoPegado, sinInvocaciones.Pregunta)
	assert.Empty(t, sinInvocaciones.Textos, "una sesión sin invocaciones no tiene textos")
}

// probarLosDerivadosDelSondeo fija los derivados por texto de una sesión del
// informe de un sondeo (contracts/medida-y-casos.md §3 y §5 de H25): la
// pregunta de cada uno es la de su sesión sin la parte quitada; sin el
// documento, sus textos son los de su sesión menos los de cita cotejar, y con
// otra parte quitada, los de su sesión, como están en su informe.
func (m mundoDeJurisprudencia) probarLosDerivadosDelSondeo(t *testing.T) {
	t.Helper()

	resueltos := m.resolver(t,
		casoDeJurisprudencia(informeDeUnSondeo, sesionDelSondeoConElFragmento, ""),
		casoDeJurisprudencia(informeDeUnSondeo, sesionDelSondeoConElFragmento, "documento"),
		casoDeJurisprudencia(informeDeUnSondeo, sesionDelSondeoConElFragmento, "fallo"),
		casoDeJurisprudencia(informeDeUnSondeo, sesionDelSondeoConElFragmento, "apartado-2"),
		casoDeJurisprudencia(informeDeUnSondeo, sesionDelSondeoConLaFicha, "documento"),
	)
	deLaSesion, sinElDocumento, sinElFallo, sinElApartado := resueltos[0], resueltos[1], resueltos[2], resueltos[3]
	laFichaSinElDocumento := resueltos[4]

	assert.Equal(t, entradaConElFragmento, sinElDocumento.Pregunta)
	assert.Equal(t, []Texto{{Orden: buscarEnElSondeo, Salida: salidaDeBuscar}}, sinElDocumento.Textos)

	assert.Equal(t, entradaConElFragmento+"\n\n"+m.fragmento.hastaElFallo, sinElFallo.Pregunta)
	assert.Equal(t, deLaSesion.Textos, sinElFallo.Textos)

	assert.Equal(t, entradaConElFragmento+"\n\n"+m.fragmento.sinElApartado, sinElApartado.Pregunta)
	assert.Equal(t, deLaSesion.Textos, sinElApartado.Textos)

	assert.Equal(t, entradaConLaFicha, laFichaSinElDocumento.Pregunta)
	assert.Empty(t, laFichaSinElDocumento.Textos)
}

// probarOtrosSondeos fija que una pregunta solo necesita del fragmento lo que
// su plantilla nombra (contracts/medida-y-casos.md §2 de H25): la que no lleva
// ninguna de las dos marcas se compone aunque el fragmento no esté, y la que
// lleva el fragmento, aunque no tenga la línea en blanco que cierra su ficha.
func (m mundoDeJurisprudencia) probarOtrosSondeos(t *testing.T) {
	t.Helper()

	resueltos := m.resolver(t,
		casoDeJurisprudencia(path.Join(carpetaSinFragmento, otroSondeo), sesionDeOtroSondeoSinMarcas, ""),
		casoDeJurisprudencia(path.Join(carpetaSinFicha, otroSondeo), sesionDeOtroSondeoConElFragmento, ""),
	)
	sinMarcas, conElFragmento := resueltos[0], resueltos[1]

	assert.Equal(t, plantillaSinTextoPegado, sinMarcas.Pregunta)
	assert.Equal(t, "Resúmeme esto:\n\n"+contenidoDelFragmentoDeUnaLinea, conElFragmento.Pregunta)
}

// probarQueSeRecuerda fija que lo reconstruido de una sesión se recuerda por su
// informe, su sesión y el texto quitado (research D10 de H25): el sobre de una
// orden repetida lleva el instante de su invocación, así que si la sesión se
// reconstruyera otra vez no sería el mismo; y el derivado que quita otra parte
// no toma lo de su sesión, porque su orden cotejar recibe otro texto.
func (m mundoDeJurisprudencia) probarQueSeRecuerda(t *testing.T) {
	t.Helper()

	casos := []CasoEtiquetado{
		casoDeJurisprudencia(informeDeCita, sesionDelCotejoPorOrden, ""),
		casoDeJurisprudencia(informeDeCita, sesionDelCotejoPorOrden, "fallo"),
		casoDeJurisprudencia(informeDeCita, sesionDelCotejoPorOrden, "apartado-2"),
		casoDeJurisprudencia(informeDeUnSondeo, sesionDelSondeoConElFragmento, "fallo"),
	}

	resueltos := m.resolver(t, casos...)

	assert.Equal(t, resueltos, m.resolver(t, casos...))

	assert.NotEqual(t, resueltos[0].Textos[0].Salida, resueltos[1].Textos[0].Salida,
		"el cotejo del derivado no es el de su sesión: recibe lo que queda")
	assert.NotEqual(t, resueltos[1].Textos[0].Salida, resueltos[2].Textos[0].Salida)
}

// casosDeJurisprudenciaSinResolver son los casos del mundo de jurisprudencia
// que la reconstrucción no puede resolver (contracts/medida-y-casos.md §7 de
// H25; FR-044): la orden repetida que termina con otro código que el de su
// informe, o que no tiene ninguno en él; el recorte que no se puede hacer; y la
// pregunta de un sondeo que no está. Su error empieza por el caso, con su
// informe, su sesión y, en un derivado, lo quitado, y es una sola línea: de una
// orden con saltos de línea lleva la primera.
func casosDeJurisprudenciaSinResolver() []casoSinResolver {
	const (
		delJob     = "el caso " + informeDeCita + " "
		delSondeo  = "el caso " + informeDeUnSondeo + " "
		otroCodigo = "» termina con 2 y el informe dice 0"
	)

	conLasDosFormas := casoDeJurisprudencia(informeDeCita, sesionDelCotejoPorOrden, "fallo")
	conLasDosFormas.Quitado.Norma, conLasDosFormas.Quitado.Bloque = "BOE-A-2015-10565", "a21"

	deOtroSondeo := func(carpeta, sesion string) CasoEtiquetado {
		return casoDeJurisprudencia(path.Join(carpeta, otroSondeo), sesion, "")
	}

	return []casoSinResolver{
		{
			nombre: "orden-con-otro-codigo",
			caso:   casoDeJurisprudencia(informeDeCita, sesionConOtroCodigo, ""),
			dice:   []string{delJob + sesionConOtroCodigo + ": la orden «" + cotejarUnROJSinForma + otroCodigo},
		},
		{
			nombre: "orden-con-otro-codigo-en-un-derivado",
			caso:   casoDeJurisprudencia(informeDeCita, sesionConOtroCodigo, "fallo"),
			dice:   []string{delJob + sesionConOtroCodigo + " sin fallo: la orden «" + cotejarUnROJSinForma + otroCodigo},
		},
		{
			nombre: "orden-de-varias-lineas-con-otro-codigo",
			caso:   casoDeJurisprudencia(informeDeCita, sesionConOtroCodigoPorLlamada, ""),
			dice: []string{
				delJob + sesionConOtroCodigoPorLlamada + ": la orden «" + llamarACotejarSinForma +
					"Roj: STS 3144/2023 - ECLI:ES:TS:2023:3144…" + otroCodigo,
			},
		},
		{
			nombre: "orden-sin-codigo",
			caso:   casoDeJurisprudencia(informeDeCita, sesionSinCodigo, ""),
			dice: []string{
				delJob + sesionSinCodigo + ": la orden «" + prepararElROJ + "» termina con 0 y el informe no le da ningún código",
			},
		},
		{
			nombre: "derivado-sin-texto-pegado",
			caso:   casoDeJurisprudencia(informeDeCita, sesionSinTextoPegado, "documento"),
			dice:   []string{delJob + sesionSinTextoPegado + " sin documento: la pregunta no lleva ningún texto pegado"},
		},
		{
			nombre: "derivado-de-otra-parte",
			caso:   casoDeJurisprudencia(informeDeCita, sesionDelCotejoPorOrden, "antecedentes"),
			dice:   []string{delJob + sesionDelCotejoPorOrden + ` sin antecedentes: quitado.texto es "antecedentes"`},
		},
		{
			nombre: "derivado-con-las-dos-formas",
			caso:   conLasDosFormas,
			dice: []string{
				delJob + sesionDelCotejoPorOrden + " sin fallo: quitado lleva un bloque (BOE-A-2015-10565 a21) y un texto (fallo)",
			},
		},
		{
			nombre: "sesion-que-no-esta-en-el-sondeo",
			caso:   casoDeJurisprudencia(informeDeUnSondeo, "una-sesion-que-no-esta", ""),
			dice: []string{
				delSondeo + "una-sesion-que-no-esta: el informe " + informeDeUnSondeo + " no tiene la sesión una-sesion-que-no-esta",
			},
		},
		{
			nombre: "pregunta-que-no-esta",
			caso:   casoDeJurisprudencia(informeDeUnSondeo, sesionDelSondeoSinSuPregunta, ""),
			dice: []string{
				delSondeo + sesionDelSondeoSinSuPregunta + ": las preguntas " + preguntasDeUnSondeo +
					" no tienen la pregunta 08-no-esta-entre-las-preguntas",
			},
		},
		{
			nombre: "pregunta-de-una-eval-que-no-es-de-hoy",
			caso:   casoDeJurisprudencia(informeDeUnSondeo, sesionDelSondeoConOtraEval, ""),
			dice: []string{
				delSondeo + sesionDelSondeoConOtraEval +
					": la eval 13-no-esta.yaml de la pregunta 13-con-una-eval-que-no-esta no es de las de hoy de ",
			},
		},
		{
			nombre: "pregunta-sin-eval-ni-plantilla",
			caso:   casoDeJurisprudencia(informeDeUnSondeo, sesionDelSondeoSinNada, ""),
			dice: []string{
				delSondeo + sesionDelSondeoSinNada + ": la pregunta 14-sin-eval-ni-pregunta de " + preguntasDeUnSondeo +
					" no lleva eval ni pregunta",
			},
		},
		{
			nombre: "derivado-del-sondeo-sin-texto-pegado",
			caso:   casoDeJurisprudencia(informeDeUnSondeo, sesionDelSondeoSinInvocaciones, "fallo"),
			dice:   []string{delSondeo + sesionDelSondeoSinInvocaciones + " sin fallo: la pregunta no lleva ningún texto pegado"},
		},
		{
			nombre: "derivado-del-sondeo-sin-la-linea-del-fallo",
			caso:   casoDeJurisprudencia(informeDeUnSondeo, sesionDelSondeoConLaFicha, "apartado-2"),
			dice: []string{
				delSondeo + sesionDelSondeoConLaFicha + " sin apartado-2: el texto pegado no tiene la línea «F A L L O»",
			},
		},
		{
			nombre: "sondeo-sin-preguntas",
			caso:   deOtroSondeo(carpetaSinPreguntas, sesionDeOtroSondeoSinMarcas),
			dice:   []string{"las preguntas " + path.Join(carpetaSinPreguntas, susPreguntas) + " no se pueden leer: "},
			es:     fs.ErrNotExist,
		},
		{
			nombre: "sondeo-con-preguntas-ilegibles",
			caso:   deOtroSondeo(carpetaDePreguntasIlegibles, sesionDeOtroSondeoSinMarcas),
			dice: []string{
				"las preguntas " + path.Join(carpetaDePreguntasIlegibles, susPreguntas) + " no son las de un sondeo: ",
			},
		},
		{
			nombre: "sondeo-sin-fragmento",
			caso:   deOtroSondeo(carpetaSinFragmento, sesionDeOtroSondeoConElFragmento),
			dice: []string{
				"el fragmento " + fragmentoQueNoEsta + " de " + path.Join(carpetaSinFragmento, susPreguntas) +
					" no se puede leer: ",
			},
			es: fs.ErrNotExist,
		},
		{
			nombre: "sondeo-sin-fragmento-para-la-ficha",
			caso:   deOtroSondeo(carpetaSinFragmento, sesionDeOtroSondeoConLaFicha),
			dice:   []string{"el fragmento " + fragmentoQueNoEsta + " de "},
			es:     fs.ErrNotExist,
		},
		{
			nombre: "sondeo-con-un-fragmento-sin-ficha",
			caso:   deOtroSondeo(carpetaSinFicha, sesionDeOtroSondeoConLaFicha),
			dice: []string{
				"el fragmento " + fragmentoDeUnaLinea + " de " + path.Join(carpetaSinFicha, susPreguntas) +
					" no tiene ninguna línea en blanco: no hay ficha que poner en la pregunta con-la-ficha",
			},
		},
		{
			nombre: "sondeo-sin-sesiones",
			caso:   deOtroSondeo(carpetaSinSesiones, sesionDeOtroSondeoSinMarcas),
			dice: []string{
				"el informe " + path.Join(carpetaSinSesiones, otroSondeo) + " no es un informe de un sondeo: ",
			},
		},
	}
}

// Lo que los tests de la reconstrucción leen del sobre de una salida, por su
// cuenta: sus tres primeras claves y su data, que es, según la orden, un
// bloque, una lista de bloques, una comprobación del grafo, un fallo, la
// consulta que prepara cita o su cotejo.
type (
	sobreLeido struct {
		Ok     bool           `json:"ok"`
		Fuente string         `json:"fuente"`
		URL    string         `json:"url"`
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

	referenciaLeida struct {
		Forma string `json:"forma"`
		Valor string `json:"valor"`
	}

	consultaLeida struct {
		Referencia referenciaLeida `json:"referencia"`
	}

	cotejoLeido struct {
		Ficha struct {
			ROJ string `json:"roj"`
		} `json:"ficha"`
		Pedida     *referenciaLeida `json:"pedida"`
		EsLaPedida *bool            `json:"es_la_pedida"`
	}
)

// Las fuentes y la clase de error con las que los tests de la reconstrucción
// reconocen cada sobre, y lo que va delante de la huella del texto recibido en
// la url del sobre de cita cotejar.
const (
	fuenteDelBoeEnElSobre   = "boe.legislacion-consolidada"
	fuenteDelGrafoEnElSobre = "kitlegal.graph"
	fuenteDeCitaEnElSobre   = "kitlegal.cita"
	claseFuenteNoDisponible = "fuente-no-disponible"

	documentoEnLaURLDelSobre = "kitlegal:documento/sha256:"
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

// exigirElSobreDeCita exige que la salida del texto sea el sobre de una orden
// del applet cita que terminó con 0, como lo escribe el binario, y lo devuelve.
func exigirElSobreDeCita(t *testing.T, texto Texto) sobreLeido {
	t.Helper()

	assert.Equal(t, clavesDelSobre, clavesEnSuOrden(t, texto.Salida), "la salida de «%s» es su sobre", texto.Orden)
	assert.True(t, strings.HasSuffix(texto.Salida, "}\n"), "la salida de «%s» va como la escribe el binario", texto.Orden)

	sobre := leerSobre(t, texto.Salida)
	assert.True(t, sobre.Ok, "«%s» termina con 0", texto.Orden)
	assert.Equal(t, fuenteDeCitaEnElSobre, sobre.Fuente)

	return sobre
}

// exigirElCotejo exige que el texto sea el de esa orden cotejar, tal cual, con
// el sobre de cotejar ese documento: su url lleva la huella SHA-256 del texto
// que la orden recibió, byte a byte, que es como se ve cuál fue. Devuelve su
// cotejo.
func exigirElCotejo(t *testing.T, texto Texto, orden, documento string) cotejoLeido {
	t.Helper()

	assert.Equal(t, orden, texto.Orden, "la orden va tal cual")

	huella := sha256.Sum256([]byte(documento))
	assert.Equal(t, documentoEnLaURLDelSobre+hex.EncodeToString(huella[:]), exigirElSobreDeCita(t, texto).URL,
		"«%s» recibe ese documento y ningún otro", strings.SplitN(orden, "\n", 2)[0])

	return datosDelSobre[cotejoLeido](t, texto.Salida)
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
// una consulta de esas evals, el informe que no es un informe, la eval retirada
// que no se puede leer o que no tiene pregunta, y el grafo previo que no se
// puede copiar o cuyo comando no termina con 0. Ninguno resuelve nada, y cada
// error nombra el caso y lo que falla.
func TestResolverCasosSinPoderPreparar(t *testing.T) {
	t.Parallel()

	const (
		informeRoto   = "informes/roto.json"
		grafoDeLaLCSP = "lcsp-a1-30-redaccion-original"
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

	leidos, err := leerCasosEtiquetados([]byte(juezDe(t, evalsDelRepositorio).Casos))
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
func lecturaDelBloque(t *testing.T, texto Texto, bloque Quitado) (verbo string, esSuLectura bool) {
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

// Lo que TestEjecucionDeLaMedida da a medirAlJuez de quien lanza la medida, que
// no sale de la carpeta del juez: la skill; el commit; cuántos casos se votan a
// la vez, que son los cuatro de contracts/medida-del-juez.md §7 de H24; y la
// fecha, que la medida lleva sin su hora.
const (
	skillDeLaMedicion        = "boe-legislacion"
	commitDeLaMedicion       = "0123456789abcdef0123456789abcdef01234567"
	concurrenciaDeLaMedicion = 4
	fechaDeLaMedidaDada      = "2026-10-06"
)

// votosDeLaCopia son los votos que pide la ejecución de la medida con los casos
// de la copia del repositorio cuando la medida se cumple (research M2 de H24;
// FR-051): tres por cada uno de los 212 defectos y uno por cada uno de los 47
// correctos.
const votosDeLaCopia = 683

// formaDeLaMedidaDada es el texto de la medida que da medirAlJuez, carácter a
// carácter, el de contracts/medida-del-juez.md §7 de H24 (FR-052): sus claves en
// ese orden, con dos espacios de sangría y su salto final. Sus huecos son, por
// orden, la skill, la clase, la fecha, el modelo del juez, la versión de Claude
// Code de sus votos, la huella de la rúbrica, la de los casos, los defectos y
// los que no quedan marcados, los correctos y los que quedan marcados, y el
// commit.
const formaDeLaMedidaDada = `{
  "skill": "%s",
  "clase": "%s",
  "fecha": "%s",
  "modelo_del_juez": "%s",
  "version_de_claude_code": "%s",
  "rubrica": {
    "fichero": "rubrica.md",
    "sha256": "%s"
  },
  "casos": {
    "fichero": "casos.yaml",
    "sha256": "%s"
  },
  "defectos": {
    "casos": %d,
    "sin_marcar": %d
  },
  "correctos": {
    "casos": %d,
    "marcados": %d
  },
  "origen": "ejecución de la medida del juez del job de evals sobre %s"
}
`

// bytesDelEjemploDeLaMedida son los de la medida del ejemplo de
// contracts/medida-del-juez.md §7 de H24, con su salto final.
const bytesDelEjemploDeLaMedida = 656

// Cómo nombra la ejecución de la medida lo que el juez deja de un caso que no
// da lo que dice su etiqueta, y con qué empieza el motivo del que queda sin
// juzgar.
const (
	quedaMarcado   = "marcado"
	quedaSinMarcar = "sin marcar"
	quedaSinJuzgar = "sin juzgar: "
)

// instanteDeLaMedicion es cuándo lanza la medida TestEjecucionDeLaMedida: el
// último segundo del día de fechaDeLaMedidaDada.
func instanteDeLaMedicion() time.Time {
	return time.Date(2026, time.October, 6, 23, 59, 59, 0, time.UTC)
}

// TestEjecucionDeLaMedida es el control de umbral de FR-106 y SC-006 de H24
// (contracts/medida-del-juez.md §7 y §8; research D18 de H24; FR-050 a FR-054),
// sobre el texto que medirAlJuez devuelve, que es el que su punto de entrada
// escribe y su guion imprime. Vota un votante que responde según la etiqueta de
// cada caso —sí tres veces, cada una con una frase de su respuesta, al que es un
// defecto, y no al correcto— y cuenta sus llamadas, sin leer los votos de la
// evidencia: cuando una persona versione otra medida, pueden no estar.
//
//   - Los 259 casos bien, sobre una copia de la carpeta del juez con la rúbrica
//     cambiada, de modo que la medida versionada no corresponde: vota los 259,
//     pide 683 votos y da la medida con las huellas de la copia, el modelo y la
//     versión recibidos, y 0 de 212 y 0 de 47. Y lo mismo sin la medida
//     versionada, que no lee para nada, y con las clases del juez en otro orden.
//   - Un defecto sin marcar, y un correcto marcado: la medida, con su recuento,
//     y un error con el caso y sus frases.
//   - Un voto que no llega: un error con el caso y su motivo, y ninguna medida.
//   - Un voto nulo, que se repite: de él cuenta su repetición, y la frase que
//     no está en la respuesta no es una de las del caso.
//   - Un modelo que no es texto: un error tras los 683 votos, y ninguna medida.
//   - Lo que impide votar —sin juez, sin votante, sin ningún caso a la vez, con
//     unos casos que no se leen, que son de otra clase o que no se resuelven—:
//     un error, ningún voto y ninguna medida.
//
// Ningún caso depende de que la medida versionada corresponda. No es paralelo,
// ni lo son sus casos: cada uno pone como directorio temporal del proceso uno
// suyo (t.Setenv), que al terminar sigue vacío. medirAlJuez no recibe a quien
// abre las sesiones de evals ni un directorio de sesiones, y no deja ninguna,
// ni nada más, en el temporal.
func TestEjecucionDeLaMedida(t *testing.T) {
	for _, caso := range casosDeLaEjecucionDeLaMedida() {
		t.Run(caso.nombre, func(t *testing.T) {
			temporal := t.TempDir()
			t.Setenv("TMPDIR", temporal)

			caso.probar(t)

			entradas, err := os.ReadDir(temporal)
			require.NoError(t, err)
			assert.Empty(t, entradas, "la ejecución de la medida no deja nada en el directorio temporal: ninguna sesión de evals")
		})
	}
}

// casoDeLaEjecucion es un caso de TestEjecucionDeLaMedida.
type casoDeLaEjecucion struct {
	nombre string
	probar func(t *testing.T)
}

// casosDeLaEjecucionDeLaMedida son los de TestEjecucionDeLaMedida, todos en su
// primer nivel: las ejecuciones con los 259 casos bien, las que tienen algún
// caso al que el votante responde otra cosa que lo de su etiqueta y las que no
// llegan a votar.
func casosDeLaEjecucionDeLaMedida() []casoDeLaEjecucion {
	casos := []casoDeLaEjecucion{
		{nombre: "los-259-bien-con-una-medida-que-no-corresponde", probar: probarLos259Bien},
		{nombre: "los-259-bien-sin-la-medida-versionada", probar: probarLos259SinLaMedidaVersionada},
		{nombre: "los-259-bien-con-las-clases-en-otro-orden", probar: probarLos259ConLasClasesEnOtroOrden},
		{nombre: "los-259-bien-con-un-modelo-que-no-es-texto", probar: probarLaMedidaQueNoSePuedeEscribir},
	}

	for _, conCambios := range ejecucionesConCasosCambiados() {
		casos = append(casos, casoDeLaEjecucion{nombre: conCambios.nombre, probar: conCambios.probar})
	}

	for _, sinVotar := range medicionesQueNoVotan() {
		casos = append(casos, casoDeLaEjecucion{nombre: sinVotar.nombre, probar: sinVotar.probar})
	}

	return casos
}

// probarLos259Bien ejecuta la medida de la copia cuya medida versionada no
// corresponde con el votante que responde a cada caso según su etiqueta, y
// exige la medida de lo que hay, con 0 de 212 y 0 de 47, sin error y con 683
// votos, como mucho cuatro casos a la vez. Y que esa medida sea la que una
// persona puede versionar: puesta en el lugar de la versionada, corresponde y
// se cumple.
func probarLos259Bien(t *testing.T) {
	t.Helper()

	ejemplo := fmt.Sprintf(formaDeLaMedidaDada, "boe-legislacion", "afirma_lo_no_leido", "2026-10-06", "claude-opus-5-5",
		"2.1.289", "5f1e2115b107d942ccdb4dd9b55f8f9606b1b56b52a2cc3e5a8fdef9692d06ee",
		"4827894aefc4133f99d0af93672585f70c9d3f2b8803af36eb2cec4f1811401a", 212, 0, 47, 0,
		"0123456789abcdef0123456789abcdef01234567")
	require.Len(t, ejemplo, bytesDelEjemploDeLaMedida, "premisa: la forma es la del ejemplo del contrato")

	copia := nuevaCopiaAMedir(t)
	votante := nuevoVotanteDeLaMedida(t, copia.casos, nil)

	texto, err := copia.medir(votante, concurrenciaDeLaMedicion)

	require.NoError(t, err)
	assert.Equal(t, copia.medidaDada(t, 0, 0), texto)
	votante.exigirLosVotos(t, votosDeLaCopia, concurrenciaDeLaMedicion)

	copia.juez.Medida = texto
	assert.Empty(t, comprobarLaMedida(copia.juez, copia.modelo, copia.version),
		"la medida dada, puesta en el lugar de la versionada, corresponde a lo que hay y se cumple")
}

// probarLos259SinLaMedidaVersionada deja al juez de la copia sin el contenido
// de su medida versionada, y exige lo mismo que con ella: la ejecución de la
// medida no la lee para decidir nada (FR-043, FR-051).
func probarLos259SinLaMedidaVersionada(t *testing.T) {
	t.Helper()

	copia := nuevaCopiaAMedir(t)
	copia.juez.Medida = ""

	lineas := comprobarLaMedida(copia.juez, copia.modelo, copia.version)
	require.Len(t, lineas, 1, "premisa: la medida versionada de la copia no se puede leer")
	require.True(t, strings.HasPrefix(lineas[0], medidaSinCorresponder), lineas[0])

	votante := nuevoVotanteDeLaMedida(t, copia.casos, nil)

	texto, err := copia.medir(votante, concurrenciaDeLaMedicion)

	require.NoError(t, err)
	assert.Equal(t, copia.medidaDada(t, 0, 0), texto)
	votante.exigirLosVotos(t, votosDeLaCopia, concurrenciaDeLaMedicion)
}

// probarLos259ConLasClasesEnOtroOrden ejecuta la medida de la copia con las
// clases del juez en el orden contrario al de su declaración, la que solo se
// publica delante, y exige lo mismo: lo que el juez deja de cada caso se lee de
// la clase de los casos, esté donde esté entre las del juez.
func probarLos259ConLasClasesEnOtroOrden(t *testing.T) {
	t.Helper()

	copia := nuevaCopiaAMedir(t)

	slices.Reverse(copia.juez.Clases)
	require.NotEqual(t, claseQueDecideEnElRepositorio, copia.juez.Clases[0].Nombre,
		"premisa: la clase de los casos no es la primera del juez")

	votante := nuevoVotanteDeLaMedida(t, copia.casos, nil)

	texto, err := copia.medir(votante, concurrenciaDeLaMedicion)

	require.NoError(t, err)
	assert.Equal(t, copia.medidaDada(t, 0, 0), texto)
	votante.exigirLosVotos(t, votosDeLaCopia, concurrenciaDeLaMedicion)
}

// probarLaMedidaQueNoSePuedeEscribir ejecuta la medida de la copia con un
// modelo del juez que no es texto en UTF-8, y exige que, con sus 683 votos
// dados, la ejecución termine con un error y sin ninguna medida: la que no se
// puede escribir tal cual no se da.
func probarLaMedidaQueNoSePuedeEscribir(t *testing.T) {
	t.Helper()

	copia := nuevaCopiaAMedir(t)
	copia.modelo = "claude-juez-\xff"

	votante := nuevoVotanteDeLaMedida(t, copia.casos, nil)

	texto, err := copia.medir(votante, concurrenciaDeLaMedicion)

	require.ErrorContains(t, err, "la medida no se puede escribir: ")
	assert.Empty(t, texto)
	votante.exigirLosVotos(t, votosDeLaCopia, concurrenciaDeLaMedicion)
}

// copiaAMedir es la copia de la carpeta del juez del repositorio sobre la que
// TestEjecucionDeLaMedida ejecuta la medida: con la rúbrica y los casos
// cambiados y con otro modelo y otra versión, de modo que su medida
// versionada, que es la del repositorio, no corresponde en ninguna de sus
// cuatro claves.
type copiaAMedir struct {
	// carpeta es la del juez de la copia.
	carpeta string

	// juez es el leído de ella, con sus cambios.
	juez *Juez

	// modelo y version son los que se dan a la ejecución de la medida.
	modelo  string
	version string

	// casos son los de la copia, en su orden, resueltos por el test.
	casos []CasoEtiquetado
}

// nuevaCopiaAMedir copia la carpeta del juez, le añade una línea a su rúbrica y
// un comentario a sus casos, que siguen siendo los 259, lee el juez de ella y
// resuelve sus casos con el reconstructor del repositorio. Exige como premisa
// que la comprobación de la medida versionada dé sus cuatro líneas: con una
// medida que correspondiera, el test no distinguiría la ejecución que vota de
// la que termina sin votar (FR-106).
func nuevaCopiaAMedir(t *testing.T) copiaAMedir {
	t.Helper()

	evals := copiarLaCarpetaDelJuez(t)
	carpeta := filepath.Join(evals, carpetaDelJuez)

	anadirA(ficheroDeRubricaDelJuez, "\nUna línea más, que la medida versionada no midió.\n")(t, carpeta)
	anadirA(ficheroDeCasosDelJuez, "# Un comentario más, que no cambia ningún caso.\n")(t, carpeta)

	juez := juezDe(t, evals)

	versionada, err := leerMedidaDelJuez([]byte(juez.Medida))
	require.NoError(t, err)

	copia := copiaAMedir{
		carpeta: carpeta,
		juez:    juez,
		modelo:  otroFijado(versionada.ModeloDelJuez),
		version: otroFijado(versionada.VersionDeClaudeCode),
	}

	require.Equal(t, []string{
		lineaDeLaRubrica,
		lineaDeLosCasos,
		lineaDelModelo(versionada.ModeloDelJuez, copia.modelo),
		lineaDeLaVersion(versionada.VersionDeClaudeCode, copia.version),
	}, comprobarLaMedida(juez, copia.modelo, copia.version),
		"premisa: la medida versionada de la copia no corresponde en ninguna de sus cuatro claves")

	leidos, err := leerCasosEtiquetados([]byte(juez.Casos))
	require.NoError(t, err)
	exigirLosRecuentosDeLaCopia(t, leidos)

	copia.casos, err = reconstructorDelRepositorio().resolver(leidos.Casos)
	require.NoError(t, err)

	return copia
}

// anadirA da el cambio que añade ese texto al final del fichero de ese nombre
// de la copia de la carpeta del juez.
func anadirA(fichero, texto string) cambioDeLaCarpetaDelJuez {
	return func(t *testing.T, carpeta string) {
		t.Helper()

		ruta := filepath.Join(carpeta, fichero)
		contenido := append(contenidoDelFichero(t, ruta), texto...)
		require.NoError(t, os.WriteFile(ruta, contenido, 0o600))
	}
}

// medicionDe es la medición de TestEjecucionDeLaMedida de ese juez con ese
// votante: lo demás es lo de quien lanza la medida, con un modelo y una versión
// que no son los de ninguna medida versionada.
func medicionDe(juez *Juez, votar Votante) MedicionDelJuez {
	return MedicionDelJuez{
		Skill:          skillDeLaMedicion,
		Juez:           juez,
		Votar:          votar,
		ModeloDelJuez:  "claude-juez-1-2",
		VersionDelJuez: "1.2.3",
		Concurrencia:   concurrenciaDeLaMedicion,
		Commit:         commitDeLaMedicion,
		Fecha:          instanteDeLaMedicion(),
	}
}

// medir ejecuta la medida del juez de la copia con ese votante, con el modelo y
// la versión de la copia y con tantos casos a la vez, sobre el reconstructor
// del repositorio, que ya recuerda sus casos.
func (c copiaAMedir) medir(votante *votanteDeLaMedida, aLaVez int) (string, error) {
	medicion := medicionDe(c.juez, votante.votar)
	medicion.ModeloDelJuez, medicion.VersionDelJuez, medicion.Concurrencia = c.modelo, c.version, aLaVez

	return medirAlJuez(reconstructorDelRepositorio(), medicion)
}

// medidaDada es el texto de la medida que la ejecución tiene que dar de la
// copia con esos dos recuentos: la forma del contrato con lo que hay, que son
// las huellas de la rúbrica y de los casos de la copia, calculadas aparte, y el
// modelo y la versión dados, y con lo de quien la lanza.
func (c copiaAMedir) medidaDada(t *testing.T, sinMarcar, marcados int) string {
	t.Helper()

	return fmt.Sprintf(formaDeLaMedidaDada, skillDeLaMedicion, claseQueDecideEnElRepositorio, fechaDeLaMedidaDada,
		c.modelo, c.version,
		huellaLeidaAparte(t, filepath.Join(c.carpeta, ficheroDeRubricaDelJuez)),
		huellaLeidaAparte(t, filepath.Join(c.carpeta, ficheroDeCasosDelJuez)),
		defectosDeLaCopia, sinMarcar, correctosDeLaCopia, marcados, commitDeLaMedicion)
}

// huellaLeidaAparte es la huella SHA-256 del fichero de esa ruta, en
// hexadecimal, calculada sin la del paquete.
func huellaLeidaAparte(t *testing.T, ruta string) string {
	t.Helper()

	suma := sha256.Sum256(contenidoDelFichero(t, ruta))

	return hex.EncodeToString(suma[:])
}

// votoDeLaMedida es lo que el votante de TestEjecucionDeLaMedida hace con un
// voto de un caso.
type votoDeLaMedida int

const (
	// diceQueNo es el voto que dice no en la clase de los casos.
	diceQueNo votoDeLaMedida = iota + 1

	// diceQueSi es el que dice sí, con una frase de la respuesta del caso.
	diceQueSi

	// citaLoQueNoEsta es el que dice sí con una frase que no está en la respuesta
	// del caso: un voto nulo, que se repite una vez y cuyo sí no cuenta.
	citaLoQueNoEsta

	// nuloPorLaOtraClase es el que dice sí en la clase de los casos, con la
	// primera frase de la respuesta, y también en la que solo se publica, con una
	// que no está: un voto nulo, que se repite una vez, y del que cuenta su
	// repetición aunque su frase de la clase de los casos sí esté.
	nuloPorLaOtraClase

	// noLlega es el que agota su tope y no llega a darse.
	noLlega
)

// fraseDeNingunCaso es la frase que cita el voto nulo del votante de la medida:
// no está en la respuesta de ningún caso.
const fraseDeNingunCaso = "Esta frase no la dice la respuesta de ningún caso etiquetado."

// votosDeLaEtiqueta son los votos con los que el votante responde a un caso
// según su etiqueta, que son los que dejan al juez bien medido: tres síes al
// defecto, que queda marcado, y un no al correcto, que no.
func votosDeLaEtiqueta(etiqueta string) []votoDeLaMedida {
	if etiqueta == etiquetaDefecto {
		return []votoDeLaMedida{diceQueSi, diceQueSi, diceQueSi}
	}

	return []votoDeLaMedida{diceQueNo}
}

// frasesDeLaRespuesta son las frases de una respuesta que citan los votos que
// dicen sí de ella: sus líneas con alguna letra, sin los blancos de sus
// extremos. El enésimo voto que dice sí con su frase en la respuesta cita la
// enésima, y vuelve a la primera si no hay tantas.
func frasesDeLaRespuesta(respuesta string) []string {
	var frases []string

	for linea := range strings.Lines(respuesta) {
		if strings.ContainsFunc(linea, unicode.IsLetter) {
			frases = append(frases, strings.TrimSpace(linea))
		}
	}

	return frases
}

// nombreDelCaso nombra un caso como lo hace la ejecución de la medida: su
// informe y su sesión y, en un derivado, el bloque que se quita.
func nombreDelCaso(caso CasoEtiquetado) string {
	nombre := caso.Informe + " " + caso.Sesion
	if caso.Quitado != nil {
		nombre += " sin " + caso.Quitado.Norma + " " + caso.Quitado.Bloque
	}

	return nombre
}

// lineaDelCasoMedido es la línea con la que la ejecución de la medida nombra un
// caso que no da lo que dice su etiqueta (contracts/medida-del-juez.md §7 de
// H24): «<informe> <sesión> [sin <norma> <bloque>]: etiquetado <etiqueta> y
// <marcado | sin marcar>» y, si alguno de sus votos dijo sí, «: «<frase>» ·
// …», con la frase de cada uno.
func lineaDelCasoMedido(caso CasoEtiquetado, queda string, frases ...string) string {
	linea := nombreDelCaso(caso) + ": etiquetado " + caso.Etiqueta + " y " + queda
	if len(frases) == 0 {
		return linea
	}

	citadas := make([]string, 0, len(frases))
	for _, frase := range frases {
		citadas = append(citadas, "«"+frase+"»")
	}

	return linea + ": " + strings.Join(citadas, " · ")
}

// votosDeUnCaso son los votos que el votante de la medida da de un caso.
type votosDeUnCaso struct {
	// nombre es el del caso.
	nombre string

	// votos son las grabaciones de sus votos, en su orden.
	votos []grabacion

	// pedidos son los que ya se le han pedido.
	pedidos int
}

// votanteDeLaMedida es el votante de TestEjecucionDeLaMedida. Reconoce cada
// caso por el mensaje de sus votos, que es el de su pregunta, su respuesta y
// sus textos reconstruidos, y le responde con sus votos, por orden: los de su
// etiqueta o, si el test los cambia, otros. Cuenta los votos que se le piden,
// los de cada caso y cuántos tiene en curso a la vez. Un voto con otro mensaje,
// o uno de más de un caso, hace fallar el test y no llega a darse. medirAlJuez
// lo llama desde varias gorrutinas.
type votanteDeLaMedida struct {
	t *testing.T

	// casos son los votos de cada caso, por el mensaje con el que se piden. El
	// mapa no cambia; lo que cambia de cada caso lo protege mutex.
	casos map[string]*votosDeUnCaso

	// mutex protege lo que sigue y los pedidos de cada caso.
	mutex sync.Mutex

	pedidos      int
	aLaVez       int
	maximoALaVez int
}

// nuevoVotanteDeLaMedida da el votante de esos casos resueltos: a cada uno le
// responde según su etiqueta, y a los de las posiciones de otros, con esos
// votos. Exige como premisa que no haya dos casos con el mismo mensaje, que no
// podría distinguir, y que la respuesta de cada uno tenga alguna frase que
// citar.
func nuevoVotanteDeLaMedida(t *testing.T, casos []CasoEtiquetado, otros map[int][]votoDeLaMedida) *votanteDeLaMedida {
	t.Helper()

	votante := &votanteDeLaMedida{t: t, casos: make(map[string]*votosDeUnCaso, len(casos))}

	for posicion, caso := range casos {
		votos, cambiados := otros[posicion]
		if !cambiados {
			votos = votosDeLaEtiqueta(caso.Etiqueta)
		}

		mensaje := mensajeDelVoto(caso.Pregunta, caso.Respuesta, caso.Textos)
		require.NotContains(t, votante.casos, mensaje, "premisa: el mensaje de %s no es el de otro caso", nombreDelCaso(caso))

		votante.casos[mensaje] = &votosDeUnCaso{nombre: nombreDelCaso(caso), votos: grabacionesDelCaso(t, caso, votos)}
	}

	return votante
}

// grabacionesDelCaso son las grabaciones de esos votos de un caso: la del que
// dice sí cita la frase de su respuesta que le toca entre los que la citan; la
// del nulo, una que no está en ella; la del que dice no, ninguna; y la del que
// no llega es el error del tope. Ninguno dice sí de la clase que solo se
// publica.
func grabacionesDelCaso(t *testing.T, caso CasoEtiquetado, votos []votoDeLaMedida) []grabacion {
	t.Helper()

	frases := frasesDeLaRespuesta(caso.Respuesta)
	require.NotEmpty(t, frases, "premisa: la respuesta de %s tiene alguna frase que citar", nombreDelCaso(caso))

	grabaciones := make([]grabacion, 0, len(votos))
	citadas := 0

	for _, voto := range votos {
		switch voto {
		case diceQueSi:
			grabaciones = append(grabaciones, votoDeLasDosClases(t, afirmaQueSi(frases[citadas%len(frases)]), cuentaQueNo()))
			citadas++
		case citaLoQueNoEsta:
			require.NotContains(t, caso.Respuesta, fraseDeNingunCaso, "premisa: %s no dice la frase del voto nulo", nombreDelCaso(caso))

			grabaciones = append(grabaciones, votoDeLasDosClases(t, afirmaQueSi(fraseDeNingunCaso), cuentaQueNo()))
		case nuloPorLaOtraClase:
			require.NotContains(t, caso.Respuesta, fraseDeNingunCaso, "premisa: %s no dice la frase del voto nulo", nombreDelCaso(caso))

			grabaciones = append(grabaciones, votoDeLasDosClases(t, afirmaQueSi(frases[0]), cuentaQueSi(fraseDeNingunCaso)))
		case diceQueNo:
			grabaciones = append(grabaciones, votoDeLasDosClases(t, afirmaQueNo(), cuentaQueNo()))
		case noLlega:
			grabaciones = append(grabaciones, grabacion{err: errTopeDelVoto})
		}
	}

	return grabaciones
}

func (v *votanteDeLaMedida) votar(mensaje string) ([]byte, error) {
	v.mutex.Lock()

	v.pedidos++
	v.aLaVez++
	v.maximoALaVez = max(v.maximoALaVez, v.aLaVez)

	caso, suyo := v.casos[mensaje]
	numero := 0

	if suyo {
		caso.pedidos++
		numero = caso.pedidos
	}

	v.mutex.Unlock()

	// Cede el paso con el voto en curso, para que los que se piden a la vez
	// coincidan y se cuenten juntos.
	runtime.Gosched()

	v.mutex.Lock()
	v.aLaVez--
	v.mutex.Unlock()

	switch {
	case !suyo:
		v.t.Errorf("se pide un voto con un mensaje de %d bytes que no es el de ningún caso resuelto", len(mensaje))

		return nil, errors.New("voto de ningún caso")
	case numero > len(caso.votos):
		v.t.Errorf("de %s se pide el voto %d y solo tiene %d", caso.nombre, numero, len(caso.votos))

		return nil, errors.New("voto de más")
	}

	return []byte(caso.votos[numero-1].salida), caso.votos[numero-1].err
}

// exigirLosVotos exige que al votante se le hayan pedido esos votos en total,
// de cada caso exactamente los suyos —tres del defecto que se marca y uno del
// correcto que no, si el test no los cambia— y nunca más de tantos a la vez,
// que es uno por caso que se vota.
func (v *votanteDeLaMedida) exigirLosVotos(t *testing.T, votos, aLaVez int) {
	t.Helper()

	v.mutex.Lock()
	defer v.mutex.Unlock()

	assert.Equal(t, votos, v.pedidos, "los votos pedidos")
	assert.Positive(t, v.maximoALaVez)
	assert.LessOrEqual(t, v.maximoALaVez, aLaVez, "los votos en curso a la vez, uno por caso que se vota")

	var conOtrosVotos []string

	for _, caso := range v.casos {
		if caso.pedidos != len(caso.votos) {
			conOtrosVotos = append(conOtrosVotos, fmt.Sprintf("%s: %d de %d", caso.nombre, caso.pedidos, len(caso.votos)))
		}
	}

	slices.Sort(conOtrosVotos)
	assert.Empty(t, conOtrosVotos, "los casos a los que no se les piden sus votos, ni uno más ni uno menos")
}

// casoCambiado es un caso de la copia al que el votante de una ejecución de
// TestEjecucionDeLaMedida responde otra cosa que lo de su etiqueta.
type casoCambiado struct {
	// que dice qué caso es, para la premisa de que lo hay.
	que string

	// es dice si un caso es el que se busca, con las frases que sus votos pueden
	// citar: se cambia el primero que lo es.
	es func(caso CasoEtiquetado, frases []string) bool

	// votos son los que el votante da de él.
	votos []votoDeLaMedida

	// linea da la línea con la que el error de la ejecución lo nombra, con esas
	// frases; nil si no lo nombra.
	linea func(caso CasoEtiquetado, frases []string) string
}

// ejecucionConCasosCambiados es un caso de TestEjecucionDeLaMedida en el que el
// votante responde a algún caso otra cosa que lo de su etiqueta.
type ejecucionConCasosCambiados struct {
	nombre string

	// cambiados son esos casos.
	cambiados []casoCambiado

	// aLaVez son los casos que se votan a la vez, y votos, los que se piden.
	aLaVez int
	votos  int

	// sinMarcar y marcados son los recuentos de la medida que da; con
	// sinMedida, no da ninguna.
	sinMarcar int
	marcados  int
	sinMedida bool
}

// Los casos que cambian las ejecuciones de TestEjecucionDeLaMedida: un defecto
// derivado y un correcto con frases bastantes para que cada voto que dice sí
// cite una distinta, y un defecto que no es un derivado.

func esUnDefectoDerivado(caso CasoEtiquetado, frases []string) bool {
	return caso.Etiqueta == etiquetaDefecto && caso.Quitado != nil && len(frases) >= votosParaMarcar
}

func esUnDefectoSinDerivar(caso CasoEtiquetado, _ []string) bool {
	return caso.Etiqueta == etiquetaDefecto && caso.Quitado == nil
}

func esUnCorrecto(caso CasoEtiquetado, frases []string) bool {
	return caso.Etiqueta == etiquetaCorrecto && len(frases) >= votosParaMarcar
}

// ejecucionesConCasosCambiados son las de TestEjecucionDeLaMedida (FR-052,
// FR-053, FR-106; SC-006):
//
//   - un defecto sin marcar, con los votos sí, sí y no: la medida con 1 de 212
//     y un error con el caso, que es un derivado, y sus dos frases; de uno en
//     uno, con los mismos 683 votos;
//   - un correcto marcado, con tres síes: la medida con 1 de 47 y un error con
//     el caso y sus tres frases; dos votos más;
//   - los dos a la vez, con otro defecto que no tiene ningún sí: una línea por
//     caso, en el orden de los casos, y la de ese sin ninguna frase;
//   - un correcto marcado tras un voto nulo por la clase que solo se publica,
//     que se repite: sus tres frases son las de los votos que cuentan, sin la
//     del nulo, que también está en la respuesta; tres votos más;
//   - un defecto sin marcar por un voto nulo que vuelve a serlo: su sí no
//     cuenta, y su única frase es la del voto que dijo sí con ella;
//   - y un voto que no llega, con un correcto marcado en otra parte: un error
//     con el caso sin juzgar y su motivo, y ninguna medida.
func ejecucionesConCasosCambiados() []ejecucionConCasosCambiados {
	sinMarcarConDosFrases := casoCambiado{
		que: "un defecto derivado con tres frases", es: esUnDefectoDerivado,
		votos: []votoDeLaMedida{diceQueSi, diceQueSi, diceQueNo},
		linea: func(caso CasoEtiquetado, frases []string) string {
			return lineaDelCasoMedido(caso, quedaSinMarcar, frases[0], frases[1])
		},
	}
	marcado := casoCambiado{
		que: "un correcto con tres frases", es: esUnCorrecto,
		votos: []votoDeLaMedida{diceQueSi, diceQueSi, diceQueSi},
		linea: func(caso CasoEtiquetado, frases []string) string {
			return lineaDelCasoMedido(caso, quedaMarcado, frases[0], frases[1], frases[2])
		},
	}
	sinNingunSi := casoCambiado{
		que: "un defecto que no es un derivado", es: esUnDefectoSinDerivar,
		votos: []votoDeLaMedida{diceQueNo},
		linea: func(caso CasoEtiquetado, _ []string) string { return lineaDelCasoMedido(caso, quedaSinMarcar) },
	}
	sinJuzgar := casoCambiado{
		que: "un defecto que no es un derivado", es: esUnDefectoSinDerivar,
		votos: []votoDeLaMedida{diceQueSi, noLlega},
		linea: func(caso CasoEtiquetado, _ []string) string {
			return nombreDelCaso(caso) + ": " + quedaSinJuzgar + motivoDelTopeDelVotoDos
		},
	}
	marcadoSinNombrar := marcado
	marcadoSinNombrar.linea = nil

	marcadoTrasUnNulo := marcado
	marcadoTrasUnNulo.votos = []votoDeLaMedida{nuloPorLaOtraClase, diceQueSi, diceQueSi, diceQueSi}

	sinMarcarPorUnNulo := casoCambiado{
		que: "un defecto derivado con tres frases", es: esUnDefectoDerivado,
		votos: []votoDeLaMedida{diceQueSi, citaLoQueNoEsta, citaLoQueNoEsta},
		linea: func(caso CasoEtiquetado, frases []string) string {
			return lineaDelCasoMedido(caso, quedaSinMarcar, frases[0])
		},
	}

	return []ejecucionConCasosCambiados{
		{
			nombre:    "un-defecto-sin-marcar",
			cambiados: []casoCambiado{sinMarcarConDosFrases},
			aLaVez:    1, votos: votosDeLaCopia,
			sinMarcar: 1,
		},
		{
			nombre:    "un-correcto-marcado",
			cambiados: []casoCambiado{marcado},
			aLaVez:    concurrenciaDeLaMedicion, votos: votosDeLaCopia + 2,
			marcados: 1,
		},
		{
			nombre:    "varios-casos-sin-lo-que-dice-su-etiqueta",
			cambiados: []casoCambiado{marcado, sinNingunSi, sinMarcarConDosFrases},
			aLaVez:    concurrenciaDeLaMedicion, votos: votosDeLaCopia + 2 - 2,
			sinMarcar: 2, marcados: 1,
		},
		{
			nombre:    "un-correcto-marcado-tras-un-voto-nulo",
			cambiados: []casoCambiado{marcadoTrasUnNulo},
			aLaVez:    concurrenciaDeLaMedicion, votos: votosDeLaCopia + 3,
			marcados: 1,
		},
		{
			nombre:    "un-defecto-sin-marcar-por-un-voto-nulo",
			cambiados: []casoCambiado{sinMarcarPorUnNulo},
			aLaVez:    concurrenciaDeLaMedicion, votos: votosDeLaCopia,
			sinMarcar: 1,
		},
		{
			nombre:    "un-voto-que-no-llega",
			cambiados: []casoCambiado{marcadoSinNombrar, sinJuzgar},
			aLaVez:    concurrenciaDeLaMedicion, votos: votosDeLaCopia + 2 - 1,
			sinMedida: true,
		},
	}
}

// probar ejecuta la medida de la copia con el votante que responde a los casos
// cambiados con sus votos, y exige el error con las líneas de los que nombra,
// una por caso y en el orden de los casos; la medida con sus dos recuentos, o
// ninguna; y los votos pedidos.
func (e ejecucionConCasosCambiados) probar(t *testing.T) {
	t.Helper()

	copia := nuevaCopiaAMedir(t)

	otros := make(map[int][]votoDeLaMedida, len(e.cambiados))
	lineas := make(map[int]string, len(e.cambiados))

	for _, cambiado := range e.cambiados {
		posicion := slices.IndexFunc(copia.casos, func(caso CasoEtiquetado) bool {
			return cambiado.es(caso, frasesDeLaRespuesta(caso.Respuesta))
		})
		require.GreaterOrEqual(t, posicion, 0, "premisa: entre los casos de la copia hay %s", cambiado.que)
		require.NotContains(t, otros, posicion, "premisa: cada caso cambiado es uno distinto")

		otros[posicion] = cambiado.votos

		if cambiado.linea != nil {
			lineas[posicion] = cambiado.linea(copia.casos[posicion], frasesDeLaRespuesta(copia.casos[posicion].Respuesta))
		}
	}

	enSuOrden := make([]string, 0, len(lineas))
	for _, posicion := range slices.Sorted(maps.Keys(lineas)) {
		enSuOrden = append(enSuOrden, lineas[posicion])
	}

	votante := nuevoVotanteDeLaMedida(t, copia.casos, otros)

	texto, err := copia.medir(votante, e.aLaVez)

	require.EqualError(t, err, strings.Join(enSuOrden, "\n"))
	votante.exigirLosVotos(t, e.votos, e.aLaVez)

	if e.sinMedida {
		assert.Empty(t, texto, "con un caso sin juzgar no hay medida")

		return
	}

	assert.Equal(t, copia.medidaDada(t, e.sinMarcar, e.marcados), texto)
}

// medicionQueNoVota es un caso de TestEjecucionDeLaMedida en el que la ejecución
// termina con un error antes de pedir ningún voto, y sin dar ninguna medida.
type medicionQueNoVota struct {
	nombre string

	// preparar da la medición del caso, con ese votante.
	preparar func(t *testing.T, votar Votante) MedicionDelJuez

	// dice es lo que su error contiene.
	dice string
}

// medicionesQueNoVotan son las de TestEjecucionDeLaMedida: sin juez, sin
// votante, sin ningún caso que votar a la vez, con un esquema con el que no se
// puede validar ningún voto y con unos casos que no se pueden votar —los que no
// tienen su forma, los de una clase que no es del juez, los de una que solo se
// publica y los que no se pueden resolver—.
func medicionesQueNoVotan() []medicionQueNoVota {
	const (
		casosConOtraEtiqueta = "los casos etiquetados del juez (juez/casos.yaml): el caso 3 (informes/dos.json otra-sesion-02) " +
			`tiene la etiqueta "dudoso"`
		casosDeOtraClase      = "los casos son de la clase una_clase, que no es una clase del juez que decide"
		casosDeLaQueSePublica = "los casos son de la clase " + claseCuentaSuProceso + ", que no es una clase del juez que decide"
	)

	return []medicionQueNoVota{
		{
			nombre: "sin-juez",
			preparar: func(_ *testing.T, votar Votante) MedicionDelJuez {
				return medicionDe(nil, votar)
			},
			dice: "la skill " + skillDeLaMedicion + " no tiene juez",
		},
		{
			nombre: "sin-votante",
			preparar: func(t *testing.T, _ Votante) MedicionDelJuez {
				t.Helper()

				return medicionDe(juezDe(t, copiarLaCarpetaDelJuez(t)), nil)
			},
			dice: "el juez de la skill " + skillDeLaMedicion + " no tiene votante",
		},
		{
			nombre: "sin-ningun-caso-a-la-vez",
			preparar: func(t *testing.T, votar Votante) MedicionDelJuez {
				t.Helper()

				medicion := medicionDe(juezDe(t, copiarLaCarpetaDelJuez(t)), votar)
				medicion.Concurrencia = 0

				return medicion
			},
			dice: "los casos que se votan a la vez son 0 y tienen que ser al menos 1",
		},
		{
			nombre: "con-un-esquema-que-no-compila",
			preparar: func(t *testing.T, votar Votante) MedicionDelJuez {
				t.Helper()

				juez := juezDe(t, copiarLaCarpetaDelJuez(t))
				juez.Esquema = "{"

				return medicionDe(juez, votar)
			},
			dice: "el esquema de la respuesta del juez no sirve para validar sus votos",
		},
		{
			nombre:   "casos-con-otra-etiqueta",
			preparar: medicionConEstosCasos(strings.Replace(casosSinteticos, "etiqueta: correcto", "etiqueta: dudoso", 1)),
			dice:     casosConOtraEtiqueta,
		},
		{
			nombre:   "casos-de-una-clase-que-no-es-del-juez",
			preparar: medicionConEstosCasos(casosSinteticos),
			dice:     casosDeOtraClase,
		},
		{
			nombre:   "casos-de-una-clase-que-solo-se-publica",
			preparar: medicionConEstosCasos(strings.Replace(casosSinteticos, "una_clase", claseCuentaSuProceso, 1)),
			dice:     casosDeLaQueSePublica,
		},
		{
			nombre:   "un-caso-que-no-se-resuelve",
			preparar: medicionConEstosCasos(strings.Replace(casosSinteticos, "una_clase", claseQueDecideEnElRepositorio, 1)),
			dice:     "el caso informes/uno.json una-sesion-01: el informe informes/uno.json no se puede leer",
		},
	}
}

// medicionConEstosCasos da la medición del juez de una copia de la carpeta del
// juez cuyos casos son ese contenido.
func medicionConEstosCasos(casos string) func(t *testing.T, votar Votante) MedicionDelJuez {
	return func(t *testing.T, votar Votante) MedicionDelJuez {
		t.Helper()

		evals := copiarLaCarpetaDelJuez(t)
		escribirEnLaCarpeta(ficheroDeCasosDelJuez, casos)(t, filepath.Join(evals, carpetaDelJuez))

		return medicionDe(juezDe(t, evals), votar)
	}
}

// probar ejecuta la medición del caso con un votante que cuenta sus llamadas, y
// exige su error, ninguna medida y ningún voto.
func (m medicionQueNoVota) probar(t *testing.T) {
	t.Helper()

	var pedidos atomic.Int64

	medicion := m.preparar(t, func(string) ([]byte, error) {
		pedidos.Add(1)

		return nil, errors.New("un voto que no se tenía que pedir")
	})

	texto, err := medirAlJuez(reconstructorDelRepositorio(), medicion)

	require.ErrorContains(t, err, m.dice)
	assert.Empty(t, texto, "sin votar no hay medida")
	assert.Zero(t, pedidos.Load(), "no se pide ningún voto")
}

// guionDeLaMedida es scripts/evals-medir-juez.sh, relativo al directorio de este
// paquete, que es donde go test ejecuta los tests.
const guionDeLaMedida = "../../scripts/evals-medir-juez.sh"

// Lo que el guion de la medida imprime alrededor de medida.json, cada marca en
// su línea (contracts/medida-del-juez.md §7 de H24), y el nombre de ese fichero
// en su temporal.
const (
	inicioDeLaMedidaImpresa   = "--- inicio de medida.json ---\n"
	finDeLaMedidaImpresa      = "--- fin de medida.json ---\n"
	ficheroDeLaMedidaDelGuion = "medida.json"
)

// Las variables obligatorias del guion de la medida que no tienen nombre en el
// paquete (contracts/job-de-evals.md §3 de H24): cuántos casos se votan a la
// vez y la ruta del claude del juez.
const (
	variableDeLaConcurrencia = "CONCURRENCIA_DE_EVALS"
	variableDelClaudeDelJuez = "CLAUDE_DEL_JUEZ"
)

// obligatoriasDelGuionDeLaMedida son las seis variables sin las que el guion de
// la medida no ejecuta nada, en el orden en el que las mira.
var obligatoriasDelGuionDeLaMedida = []string{
	variableDelModeloDelJuez, variableDeLaVersionDelJuez, variableDeLaConcurrencia, variableDelCommitEvaluado,
	variableDelClaudeDelJuez, variableDeLaSuscripcion,
}

// Lo que TestGuionDeLaMedida da al guion en las variables del modelo del juez,
// de la versión de Claude Code de sus votos y de los casos a la vez: en las
// demás, el commit de la medición, un claude que no hace nada y la suscripción
// de la base.
const (
	modeloDelGuion       = "claude-juez-1-2"
	versionDelGuion      = "1.2.3"
	concurrenciaDelGuion = "4"
)

// Lo que el guion de la medida dice en su salida de error cuando no ejecuta
// nada: su uso, sin la skill, y el principio de la línea de la variable que
// falta, que sigue con su nombre.
const (
	usoDelGuionDeLaMedida    = "evals-medir-juez: uso: scripts/evals-medir-juez.sh <skill>\n"
	faltaEnElGuionDeLaMedida = "evals-medir-juez: falta "
)

// nombreDelTemporalDeLaMedida es el del directorio que el guion de la medida
// crea en TMPDIR: la plantilla de mktemp, kitlegal-medida-del-juez.XXXXXX, con
// sus seis caracteres sustituidos.
var nombreDelTemporalDeLaMedida = regexp.MustCompile(`^kitlegal-medida-del-juez\.[A-Za-z0-9]{6}$`)

// casoDelGuionDeLaMedida es un caso de TestGuionDeLaMedida: la orden que ejecuta
// el guion de la ruta dada, escrita entera con constantes (gosec G204); lo que
// cambia en las seis variables obligatorias, que el test da todas; la medida
// que el sustituto de go escribe, ninguna si está vacía, y el código con el que
// sale; y lo que se espera del guion: su código, sus dos salidas y si no
// ejecuta go.
type casoDelGuionDeLaMedida struct {
	nombre     string
	orden      func(ctx context.Context, guion string) *exec.Cmd
	cambiar    func(t *testing.T, variables map[string]string)
	medida     string
	codigoDeGo int

	codigo  int
	salida  string
	deError string
	sinGo   bool
}

// TestGuionDeLaMedida fija scripts/evals-medir-juez.sh
// (contracts/medida-del-juez.md §7 y §8 y contracts/job-de-evals.md §3 de H24;
// FR-050, FR-052 a FR-054) con el sustituto de go de la medida delante en el
// PATH y un TMPDIR vacío del test, como TestGuionDelSondeo.
//
//   - Sin la skill —sin argumentos, con ella vacía, que es lo que da make sin
//     SKILL, o con un argumento más—, su uso; y sin una de sus seis variables
//     obligatorias, con ella vacía o con un CLAUDE_DEL_JUEZ que no es un
//     fichero ejecutable, la línea que dice cuál falta. En todos sale con 1,
//     con la salida estándar vacía y sin ejecutar go.
//   - Con todo, crea en TMPDIR su temporal con la plantilla
//     kitlegal-medida-del-juez.XXXXXX y ejecuta, en la raíz del repositorio y
//     con el tmp/ de ese temporal como TMPDIR, la orden go test del punto de
//     entrada de la medida, sin límite de tiempo, con sus banderas tras -args y
//     -salida con el medida.json de ese temporal, donde aún no hay más que
//     tmp/. Las dos salidas de go test son las del guion, que no las guarda.
//   - Si go test escribe la medida, el guion la imprime detrás, entera, entre
//     sus dos marcas, y sale con el código de go test: 0, u otro si además
//     falla, que es el defecto sin marcar o el correcto marcado (FR-052).
//   - Si go test falla sin escribirla, que es el caso sin juzgar, el guion no
//     imprime ninguna marca y sale con el código de go test (FR-053).
//
// En todos los casos el TMPDIR del test queda vacío: la medida y lo que go test
// deja en su TMPDIR, que el sustituto no borra, están en el temporal, y el
// guion lo borra. No queda nada escrito fuera de él (FR-054).
func TestGuionDeLaMedida(t *testing.T) {
	t.Parallel()

	guion, err := filepath.Abs(guionDeLaMedida)
	require.NoError(t, err)

	raiz, err := filepath.EvalSymlinks(filepath.Dir(filepath.Dir(guion)))
	require.NoError(t, err)

	for _, caso := range casosDelGuionDeLaMedida() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			comun, temporal := t.TempDir(), t.TempDir()

			variables := variablesDelGuionDeLaMedida(t)
			if caso.cambiar != nil {
				caso.cambiar(t, variables)
			}

			orden := caso.orden(t.Context(), guion)
			prepararElGuionDeLaMedida(t, orden, caso, variables, comun, temporal)

			codigo, salida, deError := codigoYSalidas(t, orden)

			assert.Equalf(t, caso.codigo, codigo, "el código del guion, con esta salida de error:\n%s", deError)
			assert.Equal(t, caso.salida, salida, "la salida estándar del guion")
			assert.Equal(t, caso.deError, deError, "la salida de error del guion")

			entradas, err := os.ReadDir(temporal)
			require.NoError(t, err)
			assert.Empty(t, entradas, "el guion borra su temporal y el TMPDIR queda vacío")

			anotaciones := filepath.Join(comun, sustitutoGo)
			if caso.sinGo {
				assert.NoDirExists(t, anotaciones, "sin la skill o sin una variable obligatoria, el guion no ejecuta go")

				return
			}

			exigirLaOrdenDeLaMedida(t, leerElGoAnotado(t, anotaciones), variables[variableDelClaudeDelJuez], temporal, raiz)
		})
	}
}

// casosDelGuionDeLaMedida son los casos de TestGuionDeLaMedida: los tres en los
// que el guion ejecuta go test —escribe la medida y sale con 0, falla sin
// escribirla y la escribe y falla, con dos códigos que no son el 1 de los demás
// casos—, los tres sin la skill y los de cada variable obligatoria.
func casosDelGuionDeLaMedida() []casoDelGuionDeLaMedida {
	cumplida, sinMarcar := medidaDelGuion(0, 0), medidaDelGuion(1, 0)

	casos := []casoDelGuionDeLaMedida{
		{
			nombre:  "go-test-escribe-la-medida-y-sale-con-0",
			orden:   guionDeLaMedidaConLaSkill,
			medida:  cumplida,
			codigo:  0,
			salida:  registroDeGoEnSuSalida + "\n" + inicioDeLaMedidaImpresa + cumplida + finDeLaMedidaImpresa,
			deError: registroDeGoEnLaDeError + "\n",
		},
		{
			nombre:     "go-test-falla-sin-escribir-la-medida",
			orden:      guionDeLaMedidaConLaSkill,
			codigoDeGo: 2,
			codigo:     2,
			salida:     registroDeGoEnSuSalida + "\n",
			deError:    registroDeGoEnLaDeError + "\n",
		},
		{
			nombre:     "go-test-escribe-la-medida-y-falla",
			orden:      guionDeLaMedidaConLaSkill,
			medida:     sinMarcar,
			codigoDeGo: 3,
			codigo:     3,
			salida:     registroDeGoEnSuSalida + "\n" + inicioDeLaMedidaImpresa + sinMarcar + finDeLaMedidaImpresa,
			deError:    registroDeGoEnLaDeError + "\n",
		},
		{
			nombre: "sin-argumentos",
			orden: func(ctx context.Context, guion string) *exec.Cmd {
				return exec.CommandContext(ctx, guion)
			},
			codigo:  1,
			deError: usoDelGuionDeLaMedida,
			sinGo:   true,
		},
		{
			nombre: "con-la-skill-vacia",
			orden: func(ctx context.Context, guion string) *exec.Cmd {
				return exec.CommandContext(ctx, guion, "")
			},
			codigo:  1,
			deError: usoDelGuionDeLaMedida,
			sinGo:   true,
		},
		{
			nombre: "con-un-argumento-mas",
			orden: func(ctx context.Context, guion string) *exec.Cmd {
				return exec.CommandContext(ctx, guion, skillDeLaMedicion, otraSkillQueSondea)
			},
			codigo:  1,
			deError: usoDelGuionDeLaMedida,
			sinGo:   true,
		},
	}

	for _, variable := range obligatoriasDelGuionDeLaMedida {
		casos = append(casos,
			casoSinLaVariable("sin-"+variable, variable, func(_ *testing.T, variables map[string]string) {
				delete(variables, variable)
			}),
			casoSinLaVariable("con-"+variable+"-vacia", variable, func(_ *testing.T, variables map[string]string) {
				variables[variable] = ""
			}))
	}

	return append(casos, casosDelClaudeQueNoSirve()...)
}

// casosDelClaudeQueNoSirve son los casos de TestGuionDeLaMedida con un
// CLAUDE_DEL_JUEZ que tiene valor y no es la ruta de un fichero ejecutable: una
// en la que no hay nada, la de un fichero sin permiso de ejecución y la de un
// directorio. Lo que falta es el ejecutable, y la línea es la misma.
func casosDelClaudeQueNoSirve() []casoDelGuionDeLaMedida {
	return []casoDelGuionDeLaMedida{
		casoSinLaVariable("con-un-claude-del-juez-que-no-esta", variableDelClaudeDelJuez,
			func(t *testing.T, variables map[string]string) {
				t.Helper()

				variables[variableDelClaudeDelJuez] = filepath.Join(t.TempDir(), sustitutoClaude)
			}),
		casoSinLaVariable("con-un-claude-del-juez-que-no-es-ejecutable", variableDelClaudeDelJuez,
			func(t *testing.T, variables map[string]string) {
				t.Helper()

				ruta := filepath.Join(t.TempDir(), sustitutoClaude)
				require.NoError(t, os.WriteFile(ruta, []byte(kitlegalQueNoHaceNada), 0o600))

				variables[variableDelClaudeDelJuez] = ruta
			}),
		casoSinLaVariable("con-un-claude-del-juez-que-es-un-directorio", variableDelClaudeDelJuez,
			func(t *testing.T, variables map[string]string) {
				t.Helper()

				variables[variableDelClaudeDelJuez] = t.TempDir()
			}),
	}
}

// casoSinLaVariable es el caso de TestGuionDeLaMedida de ese nombre en el que,
// con la skill y con ese cambio en sus variables, el guion dice que falta esa
// variable y sale con 1 sin ejecutar go.
func casoSinLaVariable(nombre, variable string, cambiar func(t *testing.T, variables map[string]string),
) casoDelGuionDeLaMedida {
	return casoDelGuionDeLaMedida{
		nombre:  nombre,
		orden:   guionDeLaMedidaConLaSkill,
		cambiar: cambiar,
		codigo:  1,
		deError: faltaEnElGuionDeLaMedida + variable + "\n",
		sinGo:   true,
	}
}

// guionDeLaMedidaConLaSkill es la orden que ejecuta el guion de la medida de
// esa ruta con la skill de la medición, su único argumento.
func guionDeLaMedidaConLaSkill(ctx context.Context, guion string) *exec.Cmd {
	return exec.CommandContext(ctx, guion, skillDeLaMedicion)
}

// medidaDelGuion es la medida que el sustituto de go escribe en
// TestGuionDeLaMedida: la forma de la que da medirAlJuez, con lo que el test da
// al guion y con esos dos recuentos.
func medidaDelGuion(sinMarcar, marcados int) string {
	return fmt.Sprintf(formaDeLaMedidaDada, skillDeLaMedicion, claseQueDecideEnElRepositorio, fechaDeLaMedidaDada,
		modeloDelGuion, versionDelGuion, "huella-de-la-rubrica", "huella-de-los-casos",
		defectosDeLaCopia, sinMarcar, correctosDeLaCopia, marcados, commitDeLaMedicion)
}

// variablesDelGuionDeLaMedida son las seis variables obligatorias del guion de
// la medida, por su nombre, con lo que TestGuionDeLaMedida les da: el claude
// del juez es un fichero ejecutable de un directorio temporal del test, que el
// guion no llega a ejecutar.
func variablesDelGuionDeLaMedida(t *testing.T) map[string]string {
	t.Helper()

	dir := t.TempDir()

	raiz, err := os.OpenRoot(dir)
	require.NoError(t, err)

	defer func() { require.NoError(t, raiz.Close()) }()

	require.NoError(t, escribirEjecutable(raiz, sustitutoClaude, kitlegalQueNoHaceNada))

	return map[string]string{
		variableDelModeloDelJuez:   modeloDelGuion,
		variableDeLaVersionDelJuez: versionDelGuion,
		variableDeLaConcurrencia:   concurrenciaDelGuion,
		variableDelCommitEvaluado:  commitDeLaMedicion,
		variableDelClaudeDelJuez:   filepath.Join(dir, sustitutoClaude),
		variableDeLaSuscripcion:    valorDeLaSuscripcion,
	}
}

// prepararElGuionDeLaMedida da a la orden del guion de la medida su entorno: el
// del proceso sin sus credenciales de Claude Code ni ninguna de las variables
// obligatorias del guion, con el sustituto de go de la medida delante en el
// PATH, el TMPDIR y el directorio común dados, el código de go del caso, el
// fichero con la medida que el sustituto escribe, si el caso tiene alguna, y
// las variables obligatorias dadas.
func prepararElGuionDeLaMedida(t *testing.T, orden *exec.Cmd, caso casoDelGuionDeLaMedida,
	variables map[string]string, comun, temporal string,
) {
	t.Helper()

	entorno := []string{
		"TMPDIR=" + temporal,
		variableDelComun + "=" + comun,
		variableDeCodigo + "=" + strconv.Itoa(caso.codigoDeGo),
		variableDelPATH + "=" + escribirElGoDeLaMedida(t) + string(os.PathListSeparator) + os.Getenv(variableDelPATH),
	}

	if caso.medida != "" {
		fichero := filepath.Join(t.TempDir(), ficheroDeLaMedidaDelGuion)
		require.NoError(t, os.WriteFile(fichero, []byte(caso.medida), 0o600))

		entorno = append(entorno, variableDeLaMedida+"="+fichero)
	}

	for _, variable := range obligatoriasDelGuionDeLaMedida {
		if valor, esta := variables[variable]; esta {
			entorno = append(entorno, variable+"="+valor)
		}
	}

	orden.Env = sobreLaBase(os.Environ(), entorno,
		slices.Concat(obligatoriasDelGuionDeLaMedida, accesosDeClaudeCode, []string{variableDeLaMedida})...)
}

// exigirLaOrdenDeLaMedida exige que el sustituto de go se haya ejecutado en la
// raíz del repositorio con la orden del punto de entrada de la medida
// (ordenDeLaMedida), con ese claude del juez y con -salida en un temporal del
// TMPDIR dado, con el nombre de la plantilla, que solo tenía su tmp/, y con ese
// tmp/ como TMPDIR.
func exigirLaOrdenDeLaMedida(t *testing.T, anotado goAnotado, claude, temporal, raiz string) {
	t.Helper()

	require.NotEmpty(t, anotado.argumentos)
	temporalDelGuion := filepath.Dir(anotado.argumentos[len(anotado.argumentos)-1])

	fisico, err := filepath.EvalSymlinks(temporal)
	require.NoError(t, err)

	assert.Equal(t, fisico, filepath.Dir(temporalDelGuion), "el temporal del guion está en TMPDIR")
	assert.Regexp(t, nombreDelTemporalDeLaMedida, filepath.Base(temporalDelGuion))
	assert.Equal(t, goAnotado{
		argumentos: ordenDeLaMedida(claude, temporalDelGuion),
		directorio: raiz,
		tmpdir:     filepath.Join(temporalDelGuion, "tmp"),
		temporal:   "tmp",
	}, anotado, "lo que el sustituto de go anota de su ejecución")
}

// ordenDeLaMedida son los argumentos de go de la orden del punto de entrada de
// la medida (contracts/job-de-evals.md §3 de H24), con la skill del guion, lo de
// sus variables y, en su temporal, el fichero de la medida.
func ordenDeLaMedida(claude, temporal string) []string {
	return []string{
		"test", "-tags", "evals", "-count=1", "-timeout", "0", "-run", "^TestMedidaDelJuez$", "./internal/evals/", "-args",
		"-skill", skillDeLaMedicion, "-modelo-del-juez", modeloDelGuion, "-version-del-juez", versionDelGuion,
		"-claude-del-juez", claude, "-concurrencia", concurrenciaDelGuion, "-commit", commitDeLaMedicion,
		"-salida", filepath.Join(temporal, ficheroDeLaMedidaDelGuion),
	}
}
