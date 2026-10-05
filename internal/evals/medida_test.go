package evals

import (
	"encoding/json/v2"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
