package cendoj

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/httpx"
)

// Los datos de prueba de la fuente que ninguna tarea escribe a mano, vistos
// desde este paquete (contrato fuente-cendoj-y-grabacion §4 y §5; research
// D27): el manifiesto dice qué consultas se graban, y el paso que graba deja lo
// que la fuente responde en la carpeta de las grabaciones.
const (
	// ficheroDelManifiesto es la lista cerrada de las consultas que se graban.
	ficheroDelManifiesto = "testdata/grabaciones.json"
	// carpetaDeLasGrabaciones es la de las grabaciones reales de la fuente,
	// <raíz de grabación>/<fuente>: el destino de grabarConsultas cuando graba
	// el paso que graba, y la que reproduce httpx.Replay.
	carpetaDeLasGrabaciones = "testdata/" + NombreDeLaFuente
	// consultasDelManifiesto son las del contrato §4: las siete respuestas que
	// reproducen los tests y las evals, y ninguna más (FR-092).
	consultasDelManifiesto = 7
)

// Los permisos con los que se copia lo grabado: los de internal/httpx, que deja
// cada grabación solo para su dueño y su carpeta sin acceso para los demás.
const (
	permisoDeLaCarpetaDeGrabaciones = 0o750
	permisoDeUnaGrabacion           = 0o600
)

// manifiestoDeLasGrabaciones es grabaciones.json: la fuente y la lista cerrada
// de sus consultas, sin direcciones ni campos, que construye el código
// (contrato §4; data-model §8).
type manifiestoDeLasGrabaciones struct {
	Fuente    string                  `json:"fuente"`
	Consultas []consultaDelManifiesto `json:"consultas"`
}

// consultaDelManifiesto es una entrada del manifiesto: una referencia como la
// de cita resolver —un ECLI, o un ROJ, o un número de resolución con su fecha—
// y para qué se graba.
type consultaDelManifiesto struct {
	ECLI       string `json:"ecli"`
	ROJ        string `json:"roj"`
	Resolucion string `json:"resolucion"`
	Fecha      string `json:"fecha"`
	Para       string `json:"para"`
}

// Los nombres con los que internal/httpx graba y busca las peticiones de las
// grabaciones sintéticas de este fichero, que son los del contrato §4 (research
// M1): se escriben aquí, y no se calculan, porque la regla que los da es de
// internal/httpx, y la reproducción solo sirve una grabación que lleva el suyo.
const (
	grabacionDeLaPagina = "GET_https_www.poderjudicial.es_search_indexAN.jsp.json"
	grabacionDelRobots  = "GET_https_www.poderjudicial.es_robots.txt.json"

	prefijoDeLosEnvios        = "POST_https_www.poderjudicial.es_search_search.action_c_"
	grabacionDelECLIConocido  = prefijoDeLosEnvios + "ECLI_ECLI_3AES_3ATS_3A2023_3A3144_action_quer-2740d948.json"
	grabacionDelECLIInventado = prefijoDeLosEnvios + "ECLI_ECLI_3AES_3ATS_3A2023_3A999999_action_qu-27a990b2.json"
)

// ecliInventado es el de la consulta sin resultados de docs/JURISPRUDENCIA.md
// §3: un ECLI con su forma que no es el de ninguna resolución.
const ecliInventado = "ECLI:ES:TS:2023:999999"

// Lo que responde la fuente en las grabaciones sintéticas que no es la
// respuesta de un envío: la página del buscador, de la que solo cuenta el
// estado, y su robots.txt.
const (
	paginaDelBuscadorDePrueba = `<html><body><form id="frmBusquedajurisprudencia"></form></body></html>`
	robotsDePrueba            = "User-agent: *\nCrawl-delay: 5\n"
)

// respuestaDePrueba es lo que responde una petición en una grabación
// sintética: su estado y su cuerpo.
type respuestaDePrueba struct {
	estado int
	cuerpo string
}

// envioDePrueba es una consulta de las que graban los tests de este fichero y el
// nombre de la grabación de su envío.
type envioDePrueba struct {
	consulta consultaDelManifiesto
	fichero  string
}

// enviosDePrueba son las dos consultas de TestGrabacionRechazaLoNoReconocido,
// en su orden: el ECLI de la sentencia conocida y el inventado.
func enviosDePrueba() []envioDePrueba {
	return []envioDePrueba{
		{consulta: consultaDelManifiesto{ECLI: ecliConocido, Para: "una lista"}, fichero: grabacionDelECLIConocido},
		{
			consulta: consultaDelManifiesto{ECLI: ecliInventado, Para: "sin resultados"},
			fichero:  grabacionDelECLIInventado,
		},
	}
}

// formatoDeLasGrabaciones es la versión del formato de grabación de
// internal/httpx, la única que su reproducción sirve.
const formatoDeLasGrabaciones = 1

// grabacionSintetica es una grabación escrita por un test con el formato de
// internal/httpx (contrato httpx-formulario §5), que es el que lee su
// reproducción: la petición con la que se empareja —su método, su dirección y
// el cuerpo de su envío— y la respuesta que sirve.
type grabacionSintetica struct {
	Formato   int                `json:"formato"`
	GrabadoEn string             `json:"grabado_en"`
	Peticion  peticionSintetica  `json:"peticion"`
	Respuesta respuestaSintetica `json:"respuesta"`
}

// peticionSintetica es la petición de una grabación sintética.
type peticionSintetica struct {
	Metodo    string              `json:"metodo"`
	URL       string              `json:"url"`
	Cuerpo    string              `json:"cuerpo,omitempty"`
	Cabeceras map[string][]string `json:"cabeceras"`
}

// respuestaSintetica es la respuesta de una grabación sintética, con el cuerpo
// como texto.
type respuestaSintetica struct {
	Estado    int                 `json:"estado"`
	Cabeceras map[string][]string `json:"cabeceras"`
	Cuerpo    string              `json:"cuerpo"`
}

// TestGrabacionRechazaLoNoReconocido es el control de umbral de FR-093: ejecuta
// grabarConsultas, que es lo que hace el test de grabación con la fuente real,
// con un cliente en reproducción sobre grabaciones sintéticas escritas aquí
// (FR-094; contrato fuente-cendoj-y-grabacion §5). Con la página en 200 y todos
// los envíos reconocidos —una lista y una consulta sin resultados—, el destino
// recibe lo grabado, byte a byte y con el robots.txt, que la reproducción no
// pide. Con un envío cuya respuesta es una página que no se reconoce —que es
// como llega un CAPTCHA—, con la página del buscador en 403 y con una petición
// que falla, devuelve un error que nombra lo que falla y el destino no llega ni
// a crearse.
//
// El envío que no se reconoce es el segundo: que el primero ya esté reconocido
// no deja nada copiado. Y con la página en 403 los dos envíos están grabados y
// se reconocerían: si se enviaran, el destino los recibiría.
func TestGrabacionRechazaLoNoReconocido(t *testing.T) {
	t.Parallel()

	var (
		paginaCorrecta = respuestaDePrueba{estado: estadoCorrecto, cuerpo: paginaDelBuscadorDePrueba}
		conLista       = respuestaDePrueba{estado: estadoCorrecto, cuerpo: paginaConLista}
		sinResultados  = respuestaDePrueba{estado: estadoCorrecto, cuerpo: paginaSinResultados}
		sinMarcas      = respuestaDePrueba{estado: estadoCorrecto, cuerpo: paginaSinMarcas}
		noDisponible   = respuestaDePrueba{estado: 503}
	)

	casos := []struct {
		nombre string
		// pagina es lo que responde la página del buscador, y envios, lo que
		// responde el envío de cada consulta de enviosDePrueba, en su orden.
		pagina respuestaDePrueba
		envios []respuestaDePrueba
		// mensajes son fragmentos del error esperado; sin ninguno, la grabación
		// vale y el destino la recibe.
		mensajes []string
	}{
		{
			nombre: "todas reconocidas: una lista y una sin resultados",
			pagina: paginaCorrecta,
			envios: []respuestaDePrueba{conLista, sinResultados},
		},
		{
			nombre:   "un envío recibe una página que no se reconoce",
			pagina:   paginaCorrecta,
			envios:   []respuestaDePrueba{conLista, sinMarcas},
			mensajes: []string{"la consulta 2 (ecli " + ecliInventado + ")", "no se reconoce", "estado 200"},
		},
		{
			nombre:   "un envío recibe un 403",
			pagina:   paginaCorrecta,
			envios:   []respuestaDePrueba{{estado: 403, cuerpo: paginaConLista}, sinResultados},
			mensajes: []string{"la consulta 1 (ecli " + ecliConocido + ")", "no se reconoce", "estado 403"},
		},
		{
			nombre:   "la página del buscador responde 403",
			pagina:   respuestaDePrueba{estado: 403},
			envios:   []respuestaDePrueba{conLista, sinResultados},
			mensajes: []string{"la página del buscador (" + paginaDelBuscador + ")", "estado 403"},
		},
		{
			nombre:   "la página del buscador no se puede pedir",
			pagina:   noDisponible,
			envios:   []respuestaDePrueba{conLista, sinResultados},
			mensajes: []string{"la página del buscador (" + paginaDelBuscador + ")", "no se ha podido pedir"},
		},
		{
			nombre:   "un envío no se puede enviar",
			pagina:   paginaCorrecta,
			envios:   []respuestaDePrueba{conLista, noDisponible},
			mensajes: []string{"la consulta 2 (ecli " + ecliInventado + ")", "no se ha podido enviar"},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			raiz := t.TempDir()
			destino := filepath.Join(t.TempDir(), NombreDeLaFuente)
			grabadas := escribeGrabacionesDePrueba(t, raiz, caso.pagina, caso.envios)

			cliente, err := httpx.Replay(filepath.Join(raiz, NombreDeLaFuente), OpcionesDeReproduccion()...)
			require.NoError(t, err)

			err = grabarConsultas(t.Context(), cliente, consultasDePrueba(), raiz, destino)

			if len(caso.mensajes) == 0 {
				require.NoError(t, err)
				assert.Equal(t, grabadas, ficherosDe(t, destino), "el destino recibe lo grabado, y nada más")

				return
			}

			require.Error(t, err)

			for _, mensaje := range caso.mensajes {
				require.ErrorContains(t, err, mensaje)
			}

			assert.NoDirExists(t, destino, "con un fallo no se copia nada: el destino no llega a crearse")
		})
	}
}

// TestManifiestoDeLasGrabaciones fija la lectura estricta del manifiesto
// (contrato fuente-cendoj-y-grabacion §4; data-model §8). El primer subtest lee
// el del repositorio y exige sus siete consultas, con el envío que cada una da,
// para que ni el test de grabación ni los que reproducen pasen en vacío
// (FR-092). El resto demuestra, sobre manifiestos escritos aquí, lo que la
// lectura no admite: una clave que no conoce, otra fuente, ninguna consulta, una
// entrada cuya referencia no vale para cita resolver, una entrada con roj y
// fecha —esa consulta no envía la fecha— y una entrada que no dice para qué se
// graba.
func TestManifiestoDeLasGrabaciones(t *testing.T) {
	t.Parallel()

	t.Run("el del repositorio lleva sus siete consultas", func(t *testing.T) {
		t.Parallel()

		comprobarManifiestoDelRepositorio(t)
	})

	const (
		porECLI       = `{"ecli": "ECLI:ES:TS:2023:3144", "para": "una lista"}`
		porROJ        = `{"roj": "STS 3144/2023", "para": "la misma, por su ROJ"}`
		porResolucion = `{"resolucion": "1088/2023", "fecha": "2023-07-04", "para": "la misma, por su número"}`
	)

	casos := []struct {
		nombre string
		// contenido es el del manifiesto; vacío, el fichero no se escribe.
		contenido string
		// mensaje es un fragmento del error esperado; vacío, el manifiesto vale.
		mensaje string
	}{
		{nombre: "una consulta de cada forma", contenido: manifiestoCon(porECLI, porROJ, porResolucion)},
		{nombre: "sin fichero", mensaje: "no se puede leer"},
		{
			nombre:    "una clave desconocida",
			contenido: `{"fuente": "cendoj.jurisprudencia", "nota": "", "consultas": [` + porECLI + `]}`,
			mensaje:   `no tiene la forma esperada: json: unknown field "nota"`,
		},
		{
			nombre:    "una clave desconocida en una consulta",
			contenido: manifiestoCon(`{"ecli": "ECLI:ES:TS:2023:3144", "organo": "TS", "para": "una lista"}`),
			mensaje:   `no tiene la forma esperada: json: unknown field "organo"`,
		},
		{
			nombre:    "algo detrás del manifiesto",
			contenido: manifiestoCon(porECLI) + "{}",
			mensaje:   "lleva algo detrás de su valor",
		},
		{
			nombre:    "de otra fuente",
			contenido: `{"fuente": "boe.legislacion-consolidada", "consultas": [` + porECLI + `]}`,
			mensaje:   `es de la fuente "boe.legislacion-consolidada" y este paquete es "cendoj.jurisprudencia"`,
		},
		{nombre: "sin consultas", contenido: manifiestoCon(), mensaje: "no lista ninguna consulta"},
		{
			nombre:    "una consulta sin referencia",
			contenido: manifiestoCon(porECLI, `{"para": "nada"}`),
			mensaje:   "consulta 2: " + fragmentoFalta,
		},
		{
			nombre:    "una consulta con dos formas",
			contenido: manifiestoCon(`{"ecli": "ECLI:ES:TS:2023:3144", "roj": "STS 3144/2023", "para": "dos"}`),
			mensaje:   "consulta 1: la referencia lleva un ECLI y un ROJ, y " + fragmentoUnaSola,
		},
		{
			nombre:    "un ECLI con fecha",
			contenido: manifiestoCon(`{"ecli": "ECLI:ES:TS:2023:3144", "fecha": "2023-07-04", "para": "con fecha"}`),
			mensaje:   "consulta 1: " + fragmentoSinFecha,
		},
		{
			nombre:    "un ROJ con fecha",
			contenido: manifiestoCon(porECLI, porROJ, `{"roj": "STS 9999/2023", "fecha": "2023-07-04", "para": "con fecha"}`),
			mensaje:   "consulta 3: una entrada con roj no lleva fecha",
		},
		{
			nombre:    "un número de resolución sin fecha",
			contenido: manifiestoCon(`{"resolucion": "1088/2023", "para": "sin fecha"}`),
			mensaje:   "consulta 1: el número de resolución 1088/2023 " + fragmentoNecesita,
		},
		{
			nombre:    "una consulta que no dice para qué se graba",
			contenido: manifiestoCon(porECLI, `{"roj": "STS 3144/2023", "para": " "}`),
			mensaje:   "consulta 2 (roj STS 3144/2023): no dice para qué se graba",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			ruta := filepath.Join(t.TempDir(), "grabaciones.json")
			if caso.contenido != "" {
				escribeDatoDePrueba(t, ruta, []byte(caso.contenido))
			}

			manifiesto, err := leerManifiesto(ruta)
			if caso.mensaje != "" {
				require.ErrorContains(t, err, caso.mensaje)
				require.ErrorContains(t, err, ruta, "el error dice qué manifiesto es")

				return
			}

			require.NoError(t, err)
			assert.Equal(t, NombreDeLaFuente, manifiesto.Fuente)
			assert.Len(t, manifiesto.Consultas, 3)
		})
	}
}

// comprobarManifiestoDelRepositorio lee el manifiesto del repositorio y exige
// sus siete consultas, en su orden y con el envío que da cada una: los cinco
// campos fijos y los de su referencia, que son la columna «Petición» de la tabla
// del contrato §4. Una entrada con roj no lleva fecha, y su envío tampoco.
func comprobarManifiestoDelRepositorio(t *testing.T) {
	t.Helper()

	propios := []map[string]string{
		{campoECLI: ecliConocido},
		{campoROJ: rojConocido},
		{campoNumeroDeResolucion: resolucionConocida, campoFechaDesde: "04/07/2023", campoFechaHasta: "04/07/2023"},
		{campoECLI: ecliInventado},
		{campoNumeroDeResolucion: "9999/2023", campoFechaDesde: "01/01/2023", campoFechaHasta: "01/01/2023"},
		{campoROJ: "STS 9999/2023"},
		{campoNumeroDeResolucion: "3144/2023", campoFechaDesde: "01/01/2023", campoFechaHasta: "01/01/2023"},
	}

	manifiesto, err := leerManifiesto(ficheroDelManifiesto)
	require.NoError(t, err)
	require.Len(t, manifiesto.Consultas, consultasDelManifiesto,
		"el manifiesto lleva las siete consultas del contrato, y ninguna más")
	require.Len(t, propios, consultasDelManifiesto)

	for indice, consulta := range manifiesto.Consultas {
		peticion, err := consulta.peticion()
		require.NoErrorf(t, err, "consulta %d (%s)", indice+1, consulta.nombre())

		campos := camposFijos()
		maps.Copy(campos, propios[indice])

		assert.Equalf(t, httpx.Peticion{Metodo: metodoPOST, URL: formularioDeConsulta, Campos: campos}, peticion,
			"consulta %d (%s)", indice+1, consulta.nombre())
	}
}

// leerManifiesto lee el manifiesto de las grabaciones sin admitir ninguna clave
// que no conozca ni nada detrás de su valor, y exige que sea de esta fuente, que
// liste alguna consulta y que cada una lleve una referencia que vale para cita
// resolver —la que construye NuevaReferencia—, sin fecha si es un roj, y diga
// para qué se graba.
func leerManifiesto(ruta string) (manifiestoDeLasGrabaciones, error) {
	var manifiesto manifiestoDeLasGrabaciones
	if err := leerDatoDePrueba(ruta, &manifiesto); err != nil {
		return manifiestoDeLasGrabaciones{}, err
	}

	if manifiesto.Fuente != NombreDeLaFuente {
		return manifiestoDeLasGrabaciones{}, fmt.Errorf("%s es de la fuente %q y este paquete es %q",
			ruta, manifiesto.Fuente, NombreDeLaFuente)
	}

	if len(manifiesto.Consultas) == 0 {
		return manifiestoDeLasGrabaciones{}, fmt.Errorf("%s no lista ninguna consulta", ruta)
	}

	for indice, consulta := range manifiesto.Consultas {
		if _, err := consulta.referencia(); err != nil {
			return manifiestoDeLasGrabaciones{}, fmt.Errorf("%s, consulta %d: %w", ruta, indice+1, err)
		}

		if strings.TrimSpace(consulta.Para) == "" {
			return manifiestoDeLasGrabaciones{}, fmt.Errorf("%s, consulta %d (%s): no dice para qué se graba (para)",
				ruta, indice+1, consulta.nombre())
		}
	}

	return manifiesto, nil
}

// nombre es el de la consulta en un mensaje: las claves de su referencia que el
// manifiesto da, cada una con su valor.
func (consulta consultaDelManifiesto) nombre() string {
	claves := []struct{ clave, valor string }{
		{"ecli", consulta.ECLI},
		{"roj", consulta.ROJ},
		{"resolucion", consulta.Resolucion},
		{"fecha", consulta.Fecha},
	}

	partes := make([]string, 0, len(claves))

	for _, dada := range claves {
		if dada.valor != "" {
			partes = append(partes, dada.clave+" "+dada.valor)
		}
	}

	return strings.Join(partes, ", ")
}

// referencia es la de la consulta, validada como la valida cita resolver. Una
// entrada con roj no lleva fecha: la consulta por ROJ no la envía, de modo que
// la misma consulta con fecha y sin ella sería la misma grabación (contrato §4).
func (consulta consultaDelManifiesto) referencia() (Referencia, error) {
	referencia, err := NuevaReferencia(consulta.ECLI, consulta.ROJ, consulta.Resolucion, consulta.Fecha)
	if err != nil {
		return Referencia{}, err
	}

	if referencia.forma == formaROJ && referencia.fecha != "" {
		return Referencia{}, errors.New("una entrada con roj no lleva fecha: la consulta por ROJ no la envía")
	}

	return referencia, nil
}

// peticion es el envío del formulario de la consulta, el mismo con el que se
// graba y con el que se reproduce: a la dirección del formulario, con los campos
// que camposDe da para su referencia.
func (consulta consultaDelManifiesto) peticion() (httpx.Peticion, error) {
	referencia, err := consulta.referencia()
	if err != nil {
		return httpx.Peticion{}, err
	}

	return httpx.Peticion{Metodo: metodoPOST, URL: formularioDeConsulta, Campos: camposDe(referencia)}, nil
}

// peticionDeLaPagina es la primera petición de una sesión con el buscador: el
// GET de su página, que da la cookie con la que se envía el formulario.
func peticionDeLaPagina() httpx.Peticion {
	return httpx.Peticion{Metodo: metodoGET, URL: paginaDelBuscador}
}

// grabarConsultas es lo que hace el test de grabación con el cliente que graba
// (contrato fuente-cendoj-y-grabacion §5, pasos 2 a 4): abre una sola consulta
// de internal/httpx, pide la página del buscador y, con ella, envía las
// consultas en su orden y clasifica cada respuesta. Solo si la página responde
// 200 y todas las respuestas se reconocen copia a destino lo que el cliente ha
// grabado bajo la raíz de grabación, en la carpeta de la fuente. Con la página
// en otro estado, con un envío «no reconocido» o con una petición que falla
// devuelve un error que nombra lo que falla, y no copia nada: lo grabado se
// queda en la raíz, que es de quien llama (FR-092).
//
// El cliente es el que decide de dónde sale cada respuesta: el de
// TestGrabarConsultas la pide a la fuente y la graba; el de
// TestGrabacionRechazaLoNoReconocido la reproduce de la misma carpeta, sin red.
func grabarConsultas(
	ctx context.Context, cliente *httpx.Cliente, consultas []consultaDelManifiesto, raiz, destino string,
) error {
	// Todos los envíos se construyen antes de pedir nada: una consulta que no
	// se puede enviar no deja la sesión a medias.
	envios := make([]httpx.Peticion, 0, len(consultas))

	for indice, consulta := range consultas {
		envio, err := consulta.peticion()
		if err != nil {
			return fmt.Errorf("la consulta %d (%s) no se puede enviar: %w", indice+1, consulta.nombre(), err)
		}

		envios = append(envios, envio)
	}

	sesion, err := cliente.Consulta()
	if err != nil {
		return fmt.Errorf("la consulta de internal/httpx no se ha podido abrir: %w", err)
	}

	pagina, err := sesion.Pedir(ctx, schema.Contexto{}, peticionDeLaPagina())
	if err != nil {
		return fmt.Errorf("la página del buscador (%s) no se ha podido pedir: %w", paginaDelBuscador, err)
	}

	if pagina.Estado != estadoCorrecto {
		return fmt.Errorf("la página del buscador (%s) responde con el estado %d y no con el %d: "+
			"no se envía ninguna consulta y no se copia nada a %s",
			paginaDelBuscador, pagina.Estado, estadoCorrecto, destino)
	}

	for indice, envio := range envios {
		respuesta, err := sesion.Pedir(ctx, schema.Contexto{}, envio)
		if err != nil {
			return fmt.Errorf("la consulta %d (%s) no se ha podido enviar: %w",
				indice+1, consultas[indice].nombre(), err)
		}

		if clasificar(respuesta) == respuestaNoReconocida {
			return fmt.Errorf("la consulta %d (%s) recibe una respuesta que no se reconoce, con el estado %d: "+
				"no es una lista de resultados ni «%s», y no se copia nada a %s",
				indice+1, consultas[indice].nombre(), respuesta.Estado, marcaSinResultados, destino)
		}
	}

	return copiarGrabaciones(filepath.Join(raiz, NombreDeLaFuente), destino)
}

// copiarGrabaciones copia cada fichero de la carpeta de lo grabado a la de
// destino, que crea si no está: la página, el robots.txt y los envíos. Lee y
// escribe a través de un os.Root por carpeta, de modo que nada sale de ellas.
func copiarGrabaciones(origen, destino string) (err error) {
	grabado, err := os.OpenRoot(origen)
	if err != nil {
		return fmt.Errorf("lo grabado en %s no se puede abrir: %w", origen, err)
	}

	defer func() { err = errors.Join(err, grabado.Close()) }()

	entradas, err := fs.ReadDir(grabado.FS(), ".")
	if err != nil {
		return fmt.Errorf("lo grabado en %s no se puede listar: %w", origen, err)
	}

	if err := os.MkdirAll(filepath.Clean(destino), permisoDeLaCarpetaDeGrabaciones); err != nil {
		return fmt.Errorf("la carpeta de destino %s no se puede crear: %w", destino, err)
	}

	copiado, err := os.OpenRoot(destino)
	if err != nil {
		return fmt.Errorf("la carpeta de destino %s no se puede abrir: %w", destino, err)
	}

	defer func() { err = errors.Join(err, copiado.Close()) }()

	for _, entrada := range entradas {
		contenido, err := grabado.ReadFile(entrada.Name())
		if err != nil {
			return fmt.Errorf("la grabación %s de %s no se puede leer: %w", entrada.Name(), origen, err)
		}

		if err := copiado.WriteFile(entrada.Name(), contenido, permisoDeUnaGrabacion); err != nil {
			return fmt.Errorf("la grabación %s no se puede copiar a %s: %w", entrada.Name(), destino, err)
		}
	}

	return nil
}

// leerDatoDePrueba lee en destino un fichero JSON de datos de prueba: la ruta
// pasa por filepath.Clean antes de abrirse (gosec G304), y no se admiten claves
// que destino no declare ni nada detrás del valor.
func leerDatoDePrueba(ruta string, destino any) error {
	contenido, err := os.ReadFile(filepath.Clean(ruta))
	if err != nil {
		return fmt.Errorf("%s no se puede leer: %w", ruta, err)
	}

	decodificador := json.NewDecoder(bytes.NewReader(contenido))
	decodificador.DisallowUnknownFields()

	if err := decodificador.Decode(destino); err != nil {
		return fmt.Errorf("%s no tiene la forma esperada: %w", ruta, err)
	}

	if _, err := decodificador.Token(); !errors.Is(err, io.EOF) {
		return fmt.Errorf("%s lleva algo detrás de su valor", ruta)
	}

	return nil
}

// consultasDePrueba son las consultas de enviosDePrueba, en su orden.
func consultasDePrueba() []consultaDelManifiesto {
	envios := enviosDePrueba()
	consultas := make([]consultaDelManifiesto, 0, len(envios))

	for _, envio := range envios {
		consultas = append(consultas, envio.consulta)
	}

	return consultas
}

// manifiestoCon es un manifiesto de esta fuente con esas consultas, cada una
// escrita como en el fichero.
func manifiestoCon(consultas ...string) string {
	return `{"fuente": "` + NombreDeLaFuente + `", "consultas": [` + strings.Join(consultas, ", ") + `]}`
}

// escribeGrabacionesDePrueba deja en <raíz>/<fuente>, que es donde graba un
// cliente con esa raíz de grabación, las grabaciones sintéticas de una sesión
// con el buscador: la del robots.txt, que la reproducción no pide; la de la
// página; y la del envío de cada consulta de enviosDePrueba, con lo que
// responde. Devuelve el contenido de cada fichero por su nombre.
func escribeGrabacionesDePrueba(
	t *testing.T, raiz string, pagina respuestaDePrueba, envios []respuestaDePrueba,
) map[string]string {
	t.Helper()

	consultas := enviosDePrueba()
	require.Len(t, envios, len(consultas), "una respuesta por consulta de prueba")

	grabaciones := map[string]grabacionSintetica{
		grabacionDelRobots: grabacionDePrueba(
			httpx.Peticion{Metodo: metodoGET, URL: "https://www.poderjudicial.es/robots.txt"},
			respuestaDePrueba{estado: estadoCorrecto, cuerpo: robotsDePrueba}),
		grabacionDeLaPagina: grabacionDePrueba(peticionDeLaPagina(), pagina),
	}

	for indice, envio := range consultas {
		peticion, err := envio.consulta.peticion()
		require.NoError(t, err)

		grabaciones[envio.fichero] = grabacionDePrueba(peticion, envios[indice])
	}

	escritas := make(map[string]string, len(grabaciones))

	for fichero, grabacion := range grabaciones {
		contenido, err := json.MarshalIndent(grabacion, "", "  ")
		require.NoError(t, err)

		escribeDatoDePrueba(t, filepath.Join(raiz, NombreDeLaFuente, fichero), contenido)

		escritas[fichero] = string(contenido)
	}

	return escritas
}

// grabacionDePrueba es la grabación sintética de una petición con lo que
// responde: el cuerpo de un envío va codificado como lo escribe internal/httpx,
// que es con lo que la reproducción lo empareja.
func grabacionDePrueba(peticion httpx.Peticion, respuesta respuestaDePrueba) grabacionSintetica {
	var cuerpo string
	if len(peticion.Campos) > 0 {
		cuerpo = cuerpoCodificado(peticion.Campos)
	}

	return grabacionSintetica{
		Formato:   formatoDeLasGrabaciones,
		GrabadoEn: "2026-10-07T00:00:00Z",
		Peticion: peticionSintetica{
			Metodo:    peticion.Metodo,
			URL:       peticion.URL,
			Cuerpo:    cuerpo,
			Cabeceras: map[string][]string{},
		},
		Respuesta: respuestaSintetica{
			Estado:    respuesta.estado,
			Cabeceras: map[string][]string{},
			Cuerpo:    respuesta.cuerpo,
		},
	}
}

// ficherosDe devuelve el contenido de cada fichero de una carpeta por su nombre.
func ficherosDe(t *testing.T, carpeta string) map[string]string {
	t.Helper()

	arbol := os.DirFS(carpeta)

	entradas, err := fs.ReadDir(arbol, ".")
	require.NoError(t, err)

	ficheros := make(map[string]string, len(entradas))

	for _, entrada := range entradas {
		contenido, err := fs.ReadFile(arbol, entrada.Name())
		require.NoError(t, err)

		ficheros[entrada.Name()] = string(contenido)
	}

	return ficheros
}

// escribeDatoDePrueba crea un fichero de una carpeta temporal, y su carpeta,
// solo para su dueño (gosec G301 y G306), desde un auxiliar distinto del que lee
// (G703).
func escribeDatoDePrueba(t *testing.T, ruta string, contenido []byte) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(ruta), 0o750))
	require.NoError(t, os.WriteFile(ruta, contenido, 0o600))
}
