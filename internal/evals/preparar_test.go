package evals

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cache"
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

// preguntaDeLaPruebaDeRed es el pregunta.txt entero de la sesión de prueba de
// red: la pregunta de la eval 01, una línea en blanco, el texto literal del
// contrato job-de-evals §6 y un salto de línea. Se escribe aquí entero, y no con
// el texto del paquete, para que un cambio en él no pase en silencio.
const preguntaDeLaPruebaDeRed = "¿qué dice el art. 21 de la Ley 39/2015?\n" +
	"\n" +
	"Antes de responder, ejecuta también exactamente estas dos órdenes y di qué devolvieron: " +
	"`~/.claude/skills/boe-legislacion/scripts/boe articulo BOE-A-2015-10565 a9998 --json` y " +
	"`~/.claude/skills/boe-legislacion/scripts/boe articulo BOE-A-2015-10565 a9998 --offline --json`.\n"

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
			} else {
				assert.Equal(t, caso.pregunta, contenidoDeLaSesion(t, sesion, "pregunta.txt"))
				assert.Equal(t, caso.fichero+"\n", contenidoDeLaSesion(t, sesion, "eval.txt"))
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
