package evals

import (
	"bytes"
	"cmp"
	"encoding/json/v2"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/app"
	"github.com/jmorenobl/kitlegal/internal/cache"
	"github.com/jmorenobl/kitlegal/internal/core/grafo"
	"github.com/jmorenobl/kitlegal/internal/graph"
)

// Las grabaciones de H4 de la Ley 39/2015 que retiran los casos de
// TestPrepararYComprobar, por su nombre dentro del conjunto.
const (
	grabacionDeLosMetadatos = "GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_" +
		"BOE-A-2015-10565_metadatos.json"
	grabacionDelIndice = "GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_" +
		"BOE-A-2015-10565_texto_indice.json"
	grabacionDelBloqueA21 = "GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_" +
		"BOE-A-2015-10565_texto_bloque_a21.json"
)

// Contenidos de las evals sintéticas de TestPrepararDirectorioDeSesion que no
// tiene TestLeerConjunto (contrato job-de-evals §9): la del art. 22, la del
// art. 9998, que no está en ninguna grabación, y la del art. 21 sin la clave
// pregunta.
const (
	contenidoDelArticulo22 = "pregunta: \"¿qué dice el art. 22 de la Ley 39/2015?\"\n" +
		"activa: true\n" +
		"comandos:\n" +
		"  - applet: boe\n" +
		"    norma: BOE-A-2015-10565\n" +
		"    bloque: a22\n" +
		"citas:\n" +
		"  - norma: BOE-A-2015-10565\n" +
		"    bloque: a22\n"
	contenidoDelArticulo9998 = "pregunta: \"¿qué dice el art. 9998 de la Ley 39/2015?\"\n" +
		"activa: true\n" +
		"comandos:\n" +
		"  - applet: boe\n" +
		"    norma: BOE-A-2015-10565\n" +
		"    bloque: a9998\n" +
		"citas:\n" +
		"  - norma: BOE-A-2015-10565\n" +
		"    bloque: a9998\n"
	contenidoDelArticulo21SinPregunta = "activa: true\n" + comandoDelArticulo21 + citaDelArticulo21
)

// modeloDeLaSesion es el id de modelo que los tests de PrepararSesion piden para
// la sesión cuando el caso no pide otro.
const modeloDeLaSesion = "claude-haiku-4-5-20251001"

// preguntaDeLaPruebaDeRed es el pregunta.txt entero de la sesión de prueba de
// red: la pregunta de la eval 01, una línea en blanco, el texto literal del
// contrato job-de-evals §6, con kitlegal invocado desde el PATH (H19,
// contracts/skills-e-invocacion.md §6; FR-127), y un salto de línea. Se escribe
// aquí entero, y no con el texto del paquete, para que un cambio en él no pase
// en silencio.
const preguntaDeLaPruebaDeRed = "¿qué dice el art. 21 de la Ley 39/2015?\n" +
	"\n" +
	"Antes de responder, ejecuta también exactamente estas dos órdenes y di qué devolvieron: " +
	"`kitlegal boe articulo BOE-A-2015-10565 a9998 --json` y " +
	"`kitlegal boe articulo BOE-A-2015-10565 a9998 --offline --json`.\n"

// TestPrepararYComprobar fija Preparar y ComprobarSinRed (contrato
// evals-y-grabaciones §5; FR-074, FR-075, SC-010) sobre una copia de las
// grabaciones de H4 y la eval sintética del art. 21 de la LPAC: con todo lo
// grabado, ni faltas ni error; sin la grabación que necesita una consulta, la
// preparación y, por su cuenta, la comprobación sin red dan una falta que nombra
// la eval y el comando, la norma y lo que falta, o la cita; y con el reloj de la
// caché adelantado más allá de la vigencia de los metadatos, solo la
// comprobación la da.
func TestPrepararYComprobar(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string

		// retirada es la grabación que falta en la copia; vacía, ninguna.
		retirada string

		// adelanto es lo que se adelanta el reloj de la caché al comprobar.
		adelanto time.Duration

		// fragmentos es lo que tiene que nombrar una misma falta; sin ninguno,
		// no hay faltas.
		fragmentos []string
	}{
		{nombre: "completo"},
		{
			nombre:     "sin-metadatos-de-un-bloque-esperado",
			retirada:   grabacionDeLosMetadatos,
			fragmentos: []string{nombreDeEval, "comando esperado", "boe articulo BOE-A-2015-10565 a21"},
		},
		{
			// La eval 01 no espera indice como comando: lo que falta es de su norma.
			nombre:     "sin-indice-de-una-norma",
			retirada:   grabacionDelIndice,
			fragmentos: []string{nombreDeEval, "norma BOE-A-2015-10565", "indice"},
		},
		{
			nombre:     "sin-bloque-de-una-cita",
			retirada:   grabacionDelBloqueA21,
			fragmentos: []string{nombreDeEval, "cita esperada BOE-A-2015-10565 a21"},
		},
		{
			// La vigencia de los metadatos es de 300 s (data-model §8).
			nombre:     "consulta-caducada",
			adelanto:   301 * time.Second,
			fragmentos: []string{nombreDeEval, "metadatos"},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			conjunto, err := LeerConjunto(crearConjunto(t, []entradaDeConjunto{
				{nombre: nombreDeEval, contenido: contenidoDelArticulo21},
			}))
			require.NoError(t, err)
			require.Empty(t, conjunto.MalFormados)

			consultas := ConsultasNecesarias(conjunto.Evals)
			dirCache := t.TempDir()

			preparadas, err := Preparar(dirCache, []string{copiaDeLasGrabacionesDeH4(t, caso.retirada)}, consultas)
			require.NoError(t, err, "una grabación que falta es una falta, nunca el error")

			var opciones []cache.Opcion
			if caso.adelanto > 0 {
				opciones = append(opciones, cache.ConReloj(func() time.Time { return time.Now().Add(caso.adelanto) }))
			}

			comprobadas, err := ComprobarSinRed(dirCache, consultas, opciones...)
			require.NoError(t, err, "lo que falta en la caché es una falta, nunca el error")

			if len(caso.fragmentos) == 0 {
				assert.Empty(t, preparadas, "todo lo necesario está grabado")
				assert.Empty(t, comprobadas, "todo lo necesario se sirve sin red")

				return
			}

			if caso.retirada == "" {
				assert.Empty(t, preparadas, "al preparar, lo grabado sigue vigente")
			} else {
				exigeFaltaQueNombra(t, preparadas, caso.fragmentos)
			}

			exigeFaltaQueNombra(t, comprobadas, caso.fragmentos)
		})
	}
}

// TestPrepararDirectorioDeSesion fija PrepararSesion (contrato job-de-evals
// §3.2 y §9; FR-074, SC-012) sobre las grabaciones de H4 y evals sintéticas de
// la LPAC: la caché de la sesión se llena con las consultas de todas las evals y
// la sesión recibe su eval y su pregunta, con el texto de la prueba de red si se
// pide; una eval que no está, un fichero mal formado o una falta de lo grabado
// no dejan ni pregunta.txt ni eval.txt, y los dos primeros tampoco preparan nada.
func TestPrepararDirectorioDeSesion(t *testing.T) {
	t.Parallel()

	const (
		deArticulo22   = "02-lpac-articulo-22.yaml"
		sinPregunta    = "03-sin-pregunta.yaml"
		deArticulo9998 = "03-lpac-articulo-9998.yaml"
		inexistente    = "03-inexistente.yaml"
	)

	articulo21 := entradaDeConjunto{nombre: nombreDeEval, contenido: contenidoDelArticulo21}
	articulo22 := entradaDeConjunto{nombre: deArticulo22, contenido: contenidoDelArticulo22}

	// El error con el que LeerEval rechaza la eval sin pregunta es el que
	// PrepararSesion tiene que nombrar.
	_, errSinPregunta := LeerEval(sinPregunta, []byte(contenidoDelArticulo21SinPregunta))
	require.Error(t, errSinPregunta)

	casos := []struct {
		nombre      string
		entradas    []entradaDeConjunto
		fichero     string
		pruebaDeRed bool

		// modelo es el que se pide para la sesión; vacío, el de modeloDeLaSesion.
		modelo string

		// pregunta es el contenido de pregunta.txt; vacía, ni pregunta.txt ni
		// eval.txt.
		pregunta string

		// errores son los fragmentos del error; sin ninguno, sin error.
		errores []string

		// falta son los fragmentos que tiene que nombrar una misma falta; sin
		// ninguno, sin faltas.
		falta []string

		// sinPreparar exige que cache/ siga vacío.
		sinPreparar bool

		// comprobar exige que ComprobarSinRed sirva de la caché de la sesión las
		// consultas de todas las evals.
		comprobar bool
	}{
		{
			nombre:    "eval-normal",
			entradas:  []entradaDeConjunto{articulo21, articulo22},
			fichero:   deArticulo22,
			pregunta:  "¿qué dice el art. 22 de la Ley 39/2015?\n",
			comprobar: true,
		},
		{
			nombre:      "prueba-de-red",
			entradas:    []entradaDeConjunto{articulo21, articulo22},
			fichero:     nombreDeEval,
			pruebaDeRed: true,
			pregunta:    preguntaDeLaPruebaDeRed,
		},
		{
			nombre:      "eval-inexistente",
			entradas:    []entradaDeConjunto{articulo21, articulo22},
			fichero:     inexistente,
			errores:     []string{inexistente},
			sinPreparar: true,
		},
		{
			nombre: "eval-mal-formada",
			entradas: []entradaDeConjunto{
				articulo21, articulo22, {nombre: sinPregunta, contenido: contenidoDelArticulo21SinPregunta},
			},
			fichero:     deArticulo22,
			errores:     []string{sinPregunta, errSinPregunta.Error()},
			sinPreparar: true,
		},
		{
			nombre: "con-faltas",
			entradas: []entradaDeConjunto{
				articulo21, articulo22, {nombre: deArticulo9998, contenido: contenidoDelArticulo9998},
			},
			fichero: nombreDeEval,
			falta:   []string{deArticulo9998, "boe articulo BOE-A-2015-10565 a9998"},
		},
		{
			// Sin la forma de un id de modelo, modelo.txt dejaría la sesión en una
			// serie que el plan no pide (data-model §10.4).
			nombre:      "modelo-con-otra-forma",
			entradas:    []entradaDeConjunto{articulo21, articulo22},
			fichero:     nombreDeEval,
			modelo:      "Claude Haiku 4.5",
			errores:     []string{"Claude Haiku 4.5", "id de modelo"},
			sinPreparar: true,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			evals := crearConjunto(t, caso.entradas)
			sesion := t.TempDir()
			dirCache := filepath.Join(sesion, "cache")
			require.NoError(t, os.Mkdir(dirCache, 0o750))

			faltas, err := PrepararSesion(SesionAPreparar{
				Evals:       evals,
				Grabaciones: []string{GrabacionesDeH4},
				Fichero:     caso.fichero,
				Modelo:      cmp.Or(caso.modelo, modeloDeLaSesion),
				Directorio:  sesion,
				PruebaDeRed: caso.pruebaDeRed,
			})

			if len(caso.errores) == 0 {
				require.NoError(t, err)
			} else {
				require.Error(t, err)

				for _, fragmento := range caso.errores {
					require.ErrorContains(t, err, fragmento)
				}
			}

			if len(caso.falta) == 0 {
				assert.Empty(t, faltas)
			} else {
				exigeFaltaQueNombra(t, faltas, caso.falta)
			}

			if caso.pregunta == "" {
				assert.NoFileExists(t, filepath.Join(sesion, "pregunta.txt"))
				assert.NoFileExists(t, filepath.Join(sesion, "eval.txt"))
				assert.NoFileExists(t, filepath.Join(sesion, "modelo.txt"))
			} else {
				assert.Equal(t, caso.pregunta, contenidoDeLaSesion(t, sesion, "pregunta.txt"))
				assert.Equal(t, caso.fichero+"\n", contenidoDeLaSesion(t, sesion, "eval.txt"))
				assert.Equal(t, modeloDeLaSesion+"\n", contenidoDeLaSesion(t, sesion, "modelo.txt"))
			}

			if caso.sinPreparar {
				preparado, err := os.ReadDir(dirCache)
				require.NoError(t, err)
				assert.Empty(t, preparado, "sin preparar nada")
			}

			if caso.comprobar {
				exigeCacheDeTodasLasEvals(t, evals, dirCache)
			}
		})
	}
}

// Los conjuntos del directorio temporal que hace de GrafosPrevios en
// TestPrepararGrafoPrevio, por su nombre en grafo_previo.
const (
	// grafoPrevioAnterior es el único que está: la derivada sintética del art. 21
	// de la LPAC con la fecha de vigencia anterior.
	grafoPrevioAnterior = "version-anterior"

	// grafoPrevioQueNoEsta es el que ninguna preparación puede copiar.
	grafoPrevioQueNoEsta = "sin-grabaciones"
)

// Las fechas de vigencia del art. 21 de la LPAC en TestPrepararGrafoPrevio: la
// de su grabación de H4 y la anterior que le pone la derivada sintética, que no
// cambia nada más.
const (
	vigenciaGrabadaDelArticulo21  = "20161002"
	vigenciaAnteriorDelArticulo21 = "20151002"
)

// TestPrepararGrafoPrevio fija la preparación del grafo previo de la sesión
// (contrato evals-y-skill §3 de H7; research D26; FR-085) sobre un directorio
// temporal que hace de GrafosPrevios, con una derivada sintética del art. 21 de
// la LPAC escrita por el test, y las grabaciones de H4: si la eval de la sesión
// lleva grafo_previo, el grafo del mundo de cache/ tiene solo la versión
// anterior y la caché sirve la grabada; sin él, no hay grafo; solo se prepara el
// de la eval de la sesión, aunque las demás lleven uno que no se podría
// preparar; y un comando del grafo previo que no termina en 0 o que escribe en
// la salida de error —la línea de una entrega fallida— es una falta, que, como
// el grafo previo que no se puede copiar, que es el error, deja la caché sin
// preparar y la sesión sin pregunta.
func TestPrepararGrafoPrevio(t *testing.T) {
	t.Parallel()

	const (
		sinGrafoPrevio     = "02-lpac-articulo-22.yaml"
		conGrafoQueNoEsta  = "03-grafo-previo-que-no-esta.yaml"
		conBloqueSinGrabar = "04-grafo-previo-sin-grabar.yaml"
	)

	grafosPrevios := grafosPreviosConLaAnterior(t)
	evals := crearConjunto(t, []entradaDeConjunto{
		{nombre: nombreDeEval, contenido: contenidoDelArticulo21 + grafoPrevioDe(grafoPrevioAnterior, "a21")},
		{nombre: sinGrafoPrevio, contenido: contenidoDelArticulo22},
		{nombre: conGrafoQueNoEsta, contenido: contenidoDelArticulo21 + grafoPrevioDe(grafoPrevioQueNoEsta, "a21")},
		{nombre: conBloqueSinGrabar, contenido: contenidoDelArticulo21 + grafoPrevioDe(grafoPrevioAnterior, "a9998")},
	})

	casos := []struct {
		nombre  string
		fichero string

		// mundoOcupado crea en cache/, antes de preparar, un directorio con el
		// nombre de world.db: la entrega falla y el kernel lo avisa en la salida de
		// error, con el código 0.
		mundoOcupado bool

		// conGrafo exige que el grafo de cache/ tenga solo la versión anterior del
		// art. 21 y que la caché sirva sin red la grabada; sin él, no hay grafo.
		conGrafo bool

		// errores son los fragmentos del error; sin ninguno, sin error.
		errores []string

		// falta son los fragmentos que tiene que nombrar una misma falta; sin
		// ninguno, sin faltas.
		falta []string
	}{
		{nombre: "grafo-con-la-version-anterior", fichero: nombreDeEval, conGrafo: true},
		{nombre: "sin-grafo-previo", fichero: sinGrafoPrevio},
		{
			nombre:  "grafo-previo-que-no-esta",
			fichero: conGrafoQueNoEsta,
			errores: []string{conGrafoQueNoEsta, filepath.Join(grafosPrevios, grafoPrevioQueNoEsta)},
		},
		{
			nombre:  "comando-que-no-termina-en-0",
			fichero: conBloqueSinGrabar,
			// La reproducción no tiene la grabación: el fallo es inesperado.
			falta: []string{
				conBloqueSinGrabar, "el comando del grafo previo boe articulo BOE-A-2015-10565 a9998",
				"terminó con código 1", "no hay grabación",
			},
		},
		{
			nombre:       "salida-de-error",
			fichero:      nombreDeEval,
			mundoOcupado: true,
			falta: []string{
				nombreDeEval, "el comando del grafo previo boe articulo BOE-A-2015-10565 a21", "terminó con código 0",
				"lo observado no ha llegado al grafo del mundo",
			},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			sesion := t.TempDir()
			dirCache := filepath.Join(sesion, "cache")
			require.NoError(t, os.Mkdir(dirCache, 0o750))

			if caso.mundoOcupado {
				require.NoError(t, os.Mkdir(filepath.Join(dirCache, "world.db"), 0o750))
			}

			faltas, err := PrepararSesion(SesionAPreparar{
				Evals:         evals,
				Grabaciones:   []string{GrabacionesDeH4},
				GrafosPrevios: grafosPrevios,
				Fichero:       caso.fichero,
				Modelo:        modeloDeLaSesion,
				Directorio:    sesion,
			})

			if len(caso.errores) == 0 {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, syscall.ENOENT)

				for _, fragmento := range caso.errores {
					require.ErrorContains(t, err, fragmento)
				}
			}

			if len(caso.falta) == 0 {
				assert.Empty(t, faltas)
			} else {
				exigeFaltaQueNombra(t, faltas, caso.falta)
			}

			if len(caso.errores) > 0 || len(caso.falta) > 0 {
				assert.NoFileExists(t, filepath.Join(dirCache, "cache.db"), "sin preparar la caché")
				assert.NoFileExists(t, filepath.Join(sesion, "pregunta.txt"))

				if !caso.mundoOcupado {
					assert.NoFileExists(t, filepath.Join(dirCache, "world.db"), "sin nada en el grafo")
				}

				return
			}

			exigeCacheDeTodasLasEvals(t, evals, dirCache)
			assert.FileExists(t, filepath.Join(sesion, "pregunta.txt"))

			if !caso.conGrafo {
				assert.NoFileExists(t, filepath.Join(dirCache, "world.db"), "sin grafo previo no hay grafo")

				return
			}

			versiones := versionesDelGrafo(t, dirCache)
			require.Len(t, versiones, 1, "el grafo de la sesión tiene una sola versión")
			assert.Equal(t, vigenciaAnteriorDelArticulo21, versiones[0].Datos[grafo.DatoFechaVigencia],
				"la del grafo previo, no la grabada")
			assert.Contains(t, versiones[0].ID, "#a21@"+vigenciaAnteriorDelArticulo21+":")

			assert.Equal(t, vigenciaGrabadaDelArticulo21, vigenciaServidaSinRed(t, dirCache),
				"la caché de la sesión sirve la grabada, no la del grafo previo")
		})
	}
}

// grafoPrevioDe es el trozo grafo_previo de una eval sintética, con sus líneas
// completas: el conjunto de grabaciones derivadas y un comando de bloque de la
// LPAC.
func grafoPrevioDe(grabaciones, bloque string) string {
	return "grafo_previo:\n" +
		"  grabaciones: " + grabaciones + "\n" +
		"  comandos:\n" +
		"    - applet: boe\n" +
		"      norma: BOE-A-2015-10565\n" +
		"      bloque: " + bloque + "\n"
}

// grafosPreviosConLaAnterior crea un directorio temporal que hace de
// GrafosPrevios con un solo conjunto, grafoPrevioAnterior, que tiene la derivada
// sintética del art. 21 de la LPAC: su grabación de H4 con la fecha de vigencia
// anterior y nada más cambiado. Devuelve el directorio.
func grafosPreviosConLaAnterior(t *testing.T) string {
	t.Helper()

	grabada := contenidoDelFichero(t, filepath.Join(GrabacionesDeH4, grabacionDelBloqueA21))
	vigenciaGrabada := []byte(`fecha_vigencia=\"` + vigenciaGrabadaDelArticulo21 + `\"`)
	require.Equal(t, 1, bytes.Count(grabada, vigenciaGrabada),
		"la grabación de H4 del art. 21 lleva una sola versión, con su fecha de vigencia")

	derivada := bytes.Replace(grabada, vigenciaGrabada,
		[]byte(`fecha_vigencia=\"`+vigenciaAnteriorDelArticulo21+`\"`), 1)

	grafosPrevios := t.TempDir()
	conjunto := filepath.Join(grafosPrevios, grafoPrevioAnterior)
	require.NoError(t, os.Mkdir(conjunto, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(conjunto, grabacionDelBloqueA21), derivada, 0o600))

	return grafosPrevios
}

// versionesDelGrafo son las BloqueVersion del grafo del mundo de dirCache, en el
// orden de su instantánea.
func versionesDelGrafo(t *testing.T, dirCache string) []grafo.NodoDeInstantanea {
	t.Helper()

	lectura, err := graph.Leer(t.Context(), graph.ConDirectorio(dirCache))
	require.NoError(t, err)

	instantanea, err := lectura.Instantanea(t.Context())
	require.NoError(t, errors.Join(err, lectura.Close()))

	var versiones []grafo.NodoDeInstantanea

	for _, nodo := range instantanea.Nodos {
		if nodo.Tipo == grafo.TipoBloqueVersion {
			versiones = append(versiones, nodo)
		}
	}

	return versiones
}

// vigenciaServidaSinRed es la fecha de vigencia del art. 21 de la LPAC que sirve
// sin red la caché de dirCache, leída con boe articulo --offline sobre una
// reproducción vacía y sin entrega al grafo.
func vigenciaServidaSinRed(t *testing.T, dirCache string) string {
	t.Helper()

	registro, err := registroDeBoe(t.TempDir(), cache.ConDirectorio(dirCache))
	require.NoError(t, err)

	var salida, errores bytes.Buffer

	codigo := app.Main([]string{"kitlegal", "boe", "articulo", "BOE-A-2015-10565", "a21", "--offline", "--json"},
		registro, &salida, &errores, sinDatosDeConstruccion, sinDatosDeConstruccion, sinDatosDeConstruccion)
	require.Zero(t, codigo, "la caché sirve el art. 21 sin red: %s", errores.String())

	var sobre struct {
		Data struct {
			FechaVigencia string `json:"fecha_vigencia"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(salida.Bytes(), &sobre))

	return sobre.Data.FechaVigencia
}

// TestFaltaSinOrigenesODeOtroPunto fija los dos textos de Falta.String que no
// dan las consultas de ConsultasNecesarias: el de una consulta sin orígenes, que
// es solo la invocación, su código y el mensaje, y el de un origen cuyo punto no
// es ninguno de los tres de data-model §7.1, que nombra la consulta y el punto.
func TestFaltaSinOrigenesODeOtroPunto(t *testing.T) {
	t.Parallel()

	const causa = "«boe articulo BOE-A-2015-10565 a9998» terminó con código 4: " +
		"la petición GET https://www.boe.es/… no está grabada"

	casos := []struct {
		nombre   string
		origenes []Origen
		texto    string
	}{
		{nombre: "sin-origenes", texto: causa},
		{
			nombre:   "de-otro-punto",
			origenes: []Origen{{Eval: nombreDeEval, Punto: "comando de la respuesta"}},
			texto: nombreDeEval + `: la consulta boe articulo BOE-A-2015-10565 a9998 del punto "comando de la respuesta": ` +
				causa,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			falta := Falta{
				Consulta: Consulta{
					Applet:     "boe",
					Verbo:      "articulo",
					Argumentos: []string{"BOE-A-2015-10565", "a9998"},
					Origenes:   caso.origenes,
				},
				Codigo:  4,
				Mensaje: "la petición GET https://www.boe.es/… no está grabada",
			}

			assert.Equal(t, caso.texto, falta.String())
		})
	}
}

// TestPrepararConConjuntosQueNoSeCopian fija el error de Preparar con un conjunto
// de grabaciones que no se puede copiar por la estructura de un directorio
// temporal del test (contrato evals-y-grabaciones §5.1): un fichero donde se
// espera el directorio del conjunto y una carpeta donde se espera un fichero de
// grabación. Es el error, que nombra la caché y el conjunto, y nunca una lista de
// faltas.
func TestPrepararConConjuntosQueNoSeCopian(t *testing.T) {
	t.Parallel()

	t.Run("conjunto-que-es-un-fichero", func(t *testing.T) {
		t.Parallel()

		conjunto := filepath.Join(t.TempDir(), "boe")
		require.NoError(t, os.WriteFile(conjunto, nil, 0o600))

		faltas, err := exigeErrorAlPreparar(t, conjunto)

		require.ErrorIs(t, err, syscall.ENOTDIR)
		assert.Nil(t, faltas)
	})

	t.Run("grabacion-que-es-una-carpeta", func(t *testing.T) {
		t.Parallel()

		conjunto := t.TempDir()
		require.NoError(t, os.Mkdir(filepath.Join(conjunto, "respuestas"), 0o750))

		faltas, err := exigeErrorAlPreparar(t, conjunto)

		require.ErrorContains(t, err, "respuestas no es un fichero regular")
		assert.Nil(t, faltas)
	})
}

// TestPrepararSesionSinPoderLeerOEscribir fija los errores de PrepararSesion que
// provoca la estructura de un directorio temporal del test (contrato job-de-evals
// §3.2): un fichero donde se espera el directorio de evals, que termina sin
// preparar nada; y una carpeta donde va eval.txt o pregunta.txt, que termina,
// después de preparar la caché, con el error que nombra la sesión y el fichero, y
// sin faltas.
func TestPrepararSesionSinPoderLeerOEscribir(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string

		// evalsEnUnFichero pasa como directorio de evals el fichero de la eval.
		evalsEnUnFichero bool

		// carpeta es el fichero de la sesión que se crea antes como carpeta.
		carpeta string
	}{
		{nombre: "evals-que-son-un-fichero", evalsEnUnFichero: true},
		{nombre: "eval-txt-que-es-una-carpeta", carpeta: "eval.txt"},
		{nombre: "pregunta-txt-que-es-una-carpeta", carpeta: "pregunta.txt"},
		{nombre: "modelo-txt-que-es-una-carpeta", carpeta: "modelo.txt"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			evals := crearConjunto(t, []entradaDeConjunto{{nombre: nombreDeEval, contenido: contenidoDelArticulo21}})
			if caso.evalsEnUnFichero {
				evals = filepath.Join(evals, nombreDeEval)
			}

			sesion := t.TempDir()
			dirCache := filepath.Join(sesion, "cache")
			require.NoError(t, os.Mkdir(dirCache, 0o750))

			if caso.carpeta != "" {
				require.NoError(t, os.Mkdir(filepath.Join(sesion, caso.carpeta), 0o750))
			}

			faltas, err := PrepararSesion(SesionAPreparar{
				Evals:       evals,
				Grabaciones: []string{GrabacionesDeH4},
				Fichero:     nombreDeEval,
				Modelo:      modeloDeLaSesion,
				Directorio:  sesion,
			})

			require.Error(t, err)
			assert.Nil(t, faltas)

			if caso.evalsEnUnFichero {
				require.ErrorContains(t, err, "el directorio de evals "+evals+" no se puede listar")

				preparado, err := os.ReadDir(dirCache)
				require.NoError(t, err)
				assert.Empty(t, preparado, "sin preparar nada")

				return
			}

			require.ErrorContains(t, err, "la sesión "+sesion+": ")
			require.ErrorContains(t, err, filepath.Join(sesion, caso.carpeta))
		})
	}
}

// TestPrepararYComprobarSinDirectorioTemporal fija el error de Preparar y de
// ComprobarSinRed cuando no pueden crear su directorio temporal, con TMPDIR en un
// fichero de un directorio temporal del test (contrato evals-y-grabaciones §5): el
// error nombra la caché y el directorio que no se crea, y no hay faltas. Desde H7,
// lo mismo al preparar el grafo previo, cuyo error nombra el grafo previo y su
// eval (contrato evals-y-skill §3 de H7). No es paralelo, porque t.Setenv cambia
// el entorno de todo el proceso.
func TestPrepararYComprobarSinDirectorioTemporal(t *testing.T) {
	dirCache := t.TempDir()

	noEsUnDirectorio := filepath.Join(t.TempDir(), "tmp")
	require.NoError(t, os.WriteFile(noEsUnDirectorio, nil, 0o600))

	t.Setenv("TMPDIR", noEsUnDirectorio)

	preparadas, err := Preparar(dirCache, nil, nil)

	require.ErrorIs(t, err, syscall.ENOTDIR)
	require.ErrorContains(t, err,
		"preparar la caché "+dirCache+": el directorio temporal de las grabaciones no se puede crear")
	assert.Nil(t, preparadas)

	comprobadas, err := ComprobarSinRed(dirCache, nil)

	require.ErrorIs(t, err, syscall.ENOTDIR)
	require.ErrorContains(t, err,
		"comprobar sin red la caché "+dirCache+": el directorio vacío de reproducción no se puede crear")
	assert.Nil(t, comprobadas)

	previo := Eval{Fichero: nombreDeEval, GrafoPrevio: GrafoPrevio{Grabaciones: grafoPrevioAnterior}}
	delGrafoPrevio, err := prepararGrafoPrevio(dirCache, nil, GrafosPrevios, previo)

	require.ErrorIs(t, err, syscall.ENOTDIR)
	require.ErrorContains(t, err, "preparar el grafo previo "+grafoPrevioAnterior+" de la eval "+nombreDeEval+
		": el directorio temporal de las grabaciones no se puede crear")
	assert.Nil(t, delGrafoPrevio)
}

// exigeErrorAlPreparar prepara una caché nueva con el conjunto de grabaciones y
// sin consultas, exige el error que nombra la caché y el conjunto que no se puede
// copiar, y devuelve lo que Preparar devolvió.
func exigeErrorAlPreparar(t *testing.T, conjunto string) ([]Falta, error) {
	t.Helper()

	dirCache := t.TempDir()

	faltas, err := Preparar(dirCache, []string{conjunto}, nil)

	require.ErrorContains(t, err,
		"preparar la caché "+dirCache+": el conjunto de grabaciones "+conjunto+" no se puede copiar")

	return faltas, err
}

// copiaDeLasGrabacionesDeH4 copia en un directorio temporal del test las
// grabaciones de H4 sin la retirada, si la hay, y devuelve su ruta. Retirar una
// grabación que no está en el conjunto es un fallo del test.
func copiaDeLasGrabacionesDeH4(t *testing.T, retirada string) string {
	t.Helper()

	copia := t.TempDir()
	require.NoError(t, os.CopyFS(copia, os.DirFS(GrabacionesDeH4)))

	if retirada != "" {
		require.NoError(t, os.Remove(filepath.Join(copia, retirada)))
	}

	return copia
}

// exigeFaltaQueNombra exige que alguna de las faltas nombre a la vez todos los
// fragmentos, y si ninguna los nombra da todas las faltas.
func exigeFaltaQueNombra(t *testing.T, faltas []Falta, fragmentos []string) {
	t.Helper()

	textos := make([]string, 0, len(faltas))
	for _, falta := range faltas {
		textos = append(textos, falta.String())
	}

	nombra := slices.ContainsFunc(textos, func(texto string) bool {
		return !slices.ContainsFunc(fragmentos, func(fragmento string) bool {
			return !strings.Contains(texto, fragmento)
		})
	})

	assert.True(t, nombra, "ninguna falta nombra a la vez %q; faltas:\n%s", fragmentos, strings.Join(textos, "\n"))
}

// contenidoDeLaSesion es el contenido de un fichero del directorio de la
// sesión.
func contenidoDeLaSesion(t *testing.T, sesion, fichero string) string {
	t.Helper()

	// filepath.Clean es lo que el control de rutas reconoce como saneado (gosec
	// G304): la sesión es un t.TempDir() y el fichero, una constante del test.
	contenido, err := os.ReadFile(filepath.Clean(filepath.Join(sesion, fichero)))
	require.NoError(t, err)

	return string(contenido)
}

// exigeCacheDeTodasLasEvals exige que la caché de la sesión sirva sin red las
// consultas necesarias de todas las evals del directorio, entre ellas el bloque
// del art. 21, que solo necesita la eval con la que no se juzga la sesión.
func exigeCacheDeTodasLasEvals(t *testing.T, evals, dirCache string) {
	t.Helper()

	conjunto, err := LeerConjunto(evals)
	require.NoError(t, err)

	consultas := ConsultasNecesarias(conjunto.Evals)
	require.True(t, slices.ContainsFunc(consultas, func(consulta Consulta) bool {
		return consulta.Applet == "boe" && consulta.Verbo == "articulo" &&
			slices.Equal(consulta.Argumentos, []string{"BOE-A-2015-10565", "a21"})
	}), "las consultas de todas las evals incluyen el bloque del art. 21")

	comprobadas, err := ComprobarSinRed(dirCache, consultas)
	require.NoError(t, err)
	assert.Empty(t, comprobadas, "la caché de la sesión sirve las consultas de todas las evals")
}
