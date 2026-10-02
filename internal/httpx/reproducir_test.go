package httpx

import (
	"context"
	"encoding/base64"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// sitioGrabado es el host ficticio de las diez grabaciones que estas tablas
// reproducen: material de test escrito a mano, nunca una fuente real ni nada
// obtenido de la red (FR-044, control 18 del plan).
const sitioGrabado = "http://fuente.prueba"

// Las direcciones grabadas, una por fichero del directorio (plan.md §«Fixtures»).
// Se escriben aquí y no se derivan de nada: son lo que cada tabla pide.
const (
	normaGrabada     = sitioGrabado + "/norma?id=BOE-A-2015-10565"
	documentoGrabado = sitioGrabado + "/documento.pdf"
	antiguaGrabada   = sitioGrabado + "/antigua"
	movidaGrabada    = sitioGrabado + "/movida"
	perdidaSinGrabar = sitioGrabado + "/desaparecida"
	bucleGrabado     = sitioGrabado + "/bucle"
	caidaGrabada     = sitioGrabado + "/caida"
	limitadaGrabada  = sitioGrabado + "/limitada"
	colisionGrabada  = sitioGrabado + "/a,b"
	colisionBuscada  = sitioGrabado + "/a_b"
)

// Los nombres de fichero implicados, escritos a mano con la forma del contrato
// de grabación §2 y no derivados con el código que se prueba, que es lo que hace
// que un fichero mal nombrado se vea aquí y no pase por bueno.
const (
	ficheroDeLaNorma    = "GET_http_fuente.prueba_norma_q_id_BOE-A-2015-10565.json"
	ficheroDeLaCabeza   = "HEAD_http_fuente.prueba_norma_q_id_BOE-A-2015-10565.json"
	ficheroDelDocumento = "GET_http_fuente.prueba_documento.pdf.json"
	ficheroDeLasDos     = "GET_http_fuente.prueba_a_b.json"
	ficheroDelRobots    = "GET_http_fuente.prueba_robots.txt.json"
)

// permisoDeLaCopia es el de los ficheros que una tabla copia a su directorio
// temporal: solo los lee ella misma.
const permisoDeLaCopia = 0o600

// TestReplaySirveLasGrabaciones es US3 escenario 1: cada respuesta es la que el
// fichero guarda —estado, cabeceras y cuerpo— y ninguna abre una conexión,
// porque en la cadena de un cliente de reproducción no hay transporte que
// pudiera abrirla (FR-045, FR-046).
//
// Las tres formas del cuerpo que el contrato §1 admite tienen su subprueba: el
// texto, el cuerpo vacío de un HEAD —que además comprueba que el método forma
// parte de la clave de emparejamiento— y el binario en base64. Lo que se espera
// se lee del propio fichero con el tipo de grabar_test.go, que declara las
// claves del contrato literalmente: la grabación es la entrada del test, así que
// es ella y no el código que se prueba quien dice qué tiene que salir.
func TestReplaySirveLasGrabaciones(t *testing.T) {
	t.Parallel()

	t.Run("texto", func(t *testing.T) {
		t.Parallel()

		directorio := grabacionesDePrueba(t)
		peticion := Peticion{Metodo: http.MethodGet, URL: normaGrabada}

		respuesta, err := clienteDeReproduccion(t, directorio).Pedir(t.Context(), schema.Contexto{}, peticion)
		require.NoError(t, err)

		grabada := grabacionLeida(t, filepath.Join(directorio, ficheroDeLaNorma))
		require.NotNil(t, grabada.Respuesta.Cuerpo, "el cuerpo de texto va en «cuerpo» (contrato §1)")

		assert.Equal(t, http.StatusOK, respuesta.Estado)
		assert.Equal(t, "application/xml; charset=utf-8", respuesta.Cabeceras.Get("content-type"),
			"las cabeceras son las grabadas, con su nombre canónico")
		assert.Equal(t, "Sat, 12 Sep 2026 00:00:00 GMT", respuesta.Cabeceras.Get("Date"),
			"la fecha es la del fichero y no la de hoy: la reproducción no consulta el reloj (FR-048)")
		assert.Equal(t, *grabada.Respuesta.Cuerpo, string(respuesta.Cuerpo), "y el cuerpo es el grabado, entero")
		assert.Contains(t, string(respuesta.Cuerpo), "&amp;",
			"tal cual, sin interpretar ni reescribir lo que el fichero guarda (contrato §1)")
		assert.Equal(t, strconv.Itoa(len(respuesta.Cuerpo)), respuesta.Cabeceras.Get("Content-Length"))
		assert.Equal(t, peticion, respuesta.Peticion, "la petición devuelta es la que se pidió")
		assert.Equal(t, normaGrabada, respuesta.URL, "y la dirección final, la que entregó el contenido")
		assert.False(t, respuesta.Ensayo, "esto no es un ensayo: la respuesta viene de una grabación")
	})

	t.Run("cabeza", func(t *testing.T) {
		t.Parallel()

		directorio := grabacionesDePrueba(t)

		respuesta, err := clienteDeReproduccion(t, directorio).Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodHead, URL: normaGrabada})
		require.NoError(t, err)

		grabada := grabacionLeida(t, filepath.Join(directorio, ficheroDeLaCabeza))
		require.Equal(t, http.MethodHead, grabada.Peticion.Metodo,
			"el fichero que sirve a un HEAD es el suyo: el método forma parte de la clave (contrato §4)")
		require.NotNil(t, grabada.Respuesta.Cuerpo, "un cuerpo vacío se graba como «cuerpo»: \"\" y nunca en base64")

		assert.Equal(t, http.StatusOK, respuesta.Estado)
		assert.Equal(t, "application/xml; charset=utf-8", respuesta.Cabeceras.Get("content-type"))
		assert.Empty(t, respuesta.Cuerpo, "un HEAD grabado sin cuerpo se sirve sin cuerpo")
	})

	t.Run("binario", func(t *testing.T) {
		t.Parallel()

		directorio := grabacionesDePrueba(t)

		respuesta, err := clienteDeReproduccion(t, directorio).Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: documentoGrabado})
		require.NoError(t, err)

		grabada := grabacionLeida(t, filepath.Join(directorio, ficheroDelDocumento))
		require.NotNil(t, grabada.Respuesta.CuerpoBase64, "lo que no es UTF-8 válido va en «cuerpo_base64»")
		require.Nil(t, grabada.Respuesta.Cuerpo, "y las dos claves son excluyentes")

		esperado, err := base64.StdEncoding.DecodeString(*grabada.Respuesta.CuerpoBase64)
		require.NoError(t, err, "la grabación declara base64 estándar")

		assert.Equal(t, http.StatusOK, respuesta.Estado)
		assert.Equal(t, "application/pdf", respuesta.Cabeceras.Get("content-type"))
		assert.Equal(t, esperado, respuesta.Cuerpo, "quien llama recibe los bytes exactos, ya descodificados")
		assert.False(t, utf8.Valid(respuesta.Cuerpo), "que es justo por lo que se grabaron en base64")
		assert.Equal(t, strconv.Itoa(len(respuesta.Cuerpo)), respuesta.Cabeceras.Get("Content-Length"))
	})
}

// TestReplayEsDeterminista es FR-048 y US3 escenario 2: la misma petición da el
// mismo resultado siempre, con el mismo cliente y con otro, sin que intervengan
// ni el reloj ni el orden. `make test` lo ejecuta además con `-shuffle=on` y el
// quickstart con `-count=2`, que es lo que comprueba que tampoco queda estado de
// una ejecución para la siguiente.
//
// Lo único que la grabación no guarda es el instante de emisión, que sale de la
// hora del cliente al servirla (contrato httpx-acepta-e-instante §2 de H4): uno
// y otro cliente declaran la misma hora fija, y así la comparación de la
// respuesta entera exige que todo lo demás —lo que sí está grabado— salga
// idéntico.
func TestReplayEsDeterminista(t *testing.T) {
	t.Parallel()

	directorio := grabacionesDePrueba(t)
	peticion := Peticion{Metodo: http.MethodGet, URL: normaGrabada}
	horaFija := ConHora(func() time.Time { return instanteDePrueba })

	cliente := clienteDeReproduccion(t, directorio, horaFija)

	primera, err := cliente.Pedir(t.Context(), schema.Contexto{}, peticion)
	require.NoError(t, err)

	repetida, err := cliente.Pedir(t.Context(), schema.Contexto{}, peticion)
	require.NoError(t, err)
	assert.Equal(t, primera, repetida, "el mismo cliente devuelve dos veces exactamente lo mismo")

	deOtroCliente, err := clienteDeReproduccion(t, directorio, horaFija).Pedir(t.Context(), schema.Contexto{}, peticion)
	require.NoError(t, err)
	assert.Equal(t, primera, deOtroCliente, "y otro cliente sobre el mismo directorio, también")

	caida := Peticion{Metodo: http.MethodGet, URL: caidaGrabada}

	_, primerFallo := clienteDeReproduccion(t, directorio).Pedir(t.Context(), schema.Contexto{}, caida)
	require.Error(t, primerFallo)

	_, otroFallo := clienteDeReproduccion(t, directorio).Pedir(t.Context(), schema.Contexto{}, caida)
	require.Error(t, otroFallo)

	assert.Equal(t, primerFallo.Error(), otroFallo.Error(),
		"y lo que falla falla igual las dos veces, con el mismo mensaje")
}

// TestReplayPeticionSinGrabacion es US3 escenarios 3 y 5: lo que no se puede
// reproducir falla de forma ruidosa, con la clase «inesperado» que ningún fallo
// de fuente usa —para que un fixture ausente o cambiado no haga pasar por bueno
// un test que esperaba un 3, un 4 o un 5 (FR-063)—, nombrando la petición que
// faltaba y el fichero, y sin devolver nunca una respuesta vacía ni caer a la
// red (FR-047).
//
// La tabla cubre las cuatro filas del contrato §4 que no son la colisión: el
// fichero que no está y los tres que están pero no se pueden reproducir —el
// ilegible, el de otro formato y el que no declara un cuerpo que servir—, que
// para quien pide son lo mismo que no tener grabación.
func TestReplayPeticionSinGrabacion(t *testing.T) {
	t.Parallel()

	t.Run("el fichero no está", func(t *testing.T) {
		t.Parallel()

		directorio := grabacionesDePrueba(t)
		cliente, espia := clienteEspiado(t, directorio)

		respuesta, err := cliente.Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: sitioGrabado + "/inexistente"})

		fallo := falloDe(t, err)
		assert.Equal(t, schema.ClaseInesperado, fallo.Clase(),
			"una grabación ausente es un fallo del mecanismo, no de ninguna fuente (FR-063)")
		assert.Equal(t, Respuesta{}, respuesta, "y no se devuelve nunca una respuesta vacía como si valiera")
		assert.Contains(t, fallo.Error(), "GET "+sitioGrabado+"/inexistente",
			"el mensaje nombra la petición que faltaba, con su método y su dirección completa (FR-047)")
		assert.Contains(t, fallo.Error(), "GET_http_fuente.prueba_inexistente.json",
			"y el fichero que se buscó en su lugar (contrato §4)")
		assert.Equal(t, 1, espia.busquedas, "se buscó una vez y no se volvió a intentar")
	})

	ilegibles := []struct {
		nombre    string
		contenido string
	}{
		{
			nombre:    "el fichero no es un objeto JSON",
			contenido: "esto no es una grabación\n",
		},
		{
			nombre:    "el fichero declara otro formato",
			contenido: grabacionEscritaAMano(2, `{"estado": 200, "cabeceras": {}, "cuerpo": ""}`),
		},
		{
			nombre:    "el fichero declara las dos claves del cuerpo",
			contenido: grabacionEscritaAMano(1, `{"estado": 200, "cabeceras": {}, "cuerpo": "", "cuerpo_base64": ""}`),
		},
		{
			nombre:    "el fichero no declara ningún cuerpo",
			contenido: grabacionEscritaAMano(1, `{"estado": 200, "cabeceras": {}}`),
		},
	}

	for _, caso := range ilegibles {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			directorio := t.TempDir()
			ruta := filepath.Join(directorio, ficheroEscritoAMano)
			require.NoError(t, os.WriteFile(ruta, []byte(caso.contenido), permisoDeLaCopia))

			respuesta, err := clienteDeReproduccion(t, directorio).Pedir(t.Context(), schema.Contexto{},
				Peticion{Metodo: http.MethodGet, URL: direccionEscritaAMano})

			fallo := falloDe(t, err)
			assert.Equal(t, schema.ClaseInesperado, fallo.Clase(),
				"una grabación que no se puede reproducir falla como la que falta (contrato §4)")
			assert.Equal(t, Respuesta{}, respuesta, "y no se sirve a medias")
			assert.Contains(t, fallo.Error(), ruta, "el mensaje nombra el fichero")
		})
	}
}

// TestReplayGrabacionDeOtraPeticion es US3 escenario 4: el nombre no es único
// por construcción —«/a,b» y «/a_b» se sanean al mismo—, así que lo que decide
// es la petición que el fichero guarda dentro. La grabación equivocada no se
// sirve nunca, y el fallo nombra las dos peticiones y el fichero (FR-047,
// contrato §4).
func TestReplayGrabacionDeOtraPeticion(t *testing.T) {
	t.Parallel()

	directorio := grabacionesDePrueba(t)

	grabada := grabacionLeida(t, filepath.Join(directorio, ficheroDeLasDos))
	require.Equal(t, colisionGrabada, grabada.Peticion.URL,
		"el fichero de la colisión guarda la dirección con la coma (plan.md §«Fixtures», fichero 9)")

	respuesta, err := clienteDeReproduccion(t, directorio).Pedir(t.Context(), schema.Contexto{},
		Peticion{Metodo: http.MethodGet, URL: colisionBuscada})

	fallo := falloDe(t, err)
	assert.Equal(t, schema.ClaseInesperado, fallo.Clase())
	assert.Equal(t, Respuesta{}, respuesta, "la grabación de la otra petición no se sirve nunca")
	assert.Contains(t, fallo.Error(), "GET "+colisionBuscada, "el mensaje nombra la petición que se buscaba")
	assert.Contains(t, fallo.Error(), "GET "+colisionGrabada, "y la que encontró en su lugar")
	assert.Contains(t, fallo.Error(), ficheroDeLasDos, "y el fichero de las dos")
}

// TestReplayGarantiasVigentes es US3 escenario 6: la lista de FR-049 es cerrada
// y las cinco que declara vigentes siguen exigiéndose. Aquí van las tres que no
// tienen tabla propia —el contexto, la identificación y el método—; las
// redirecciones grabadas y la clasificación por el estado grabado las comprueban
// TestReplayRedirecciones y TestReplayEstadoDeErrorGrabado.
func TestReplayGarantiasVigentes(t *testing.T) {
	t.Parallel()

	t.Run("contexto", func(t *testing.T) {
		t.Parallel()

		cliente, espia := clienteEspiado(t, grabacionesDePrueba(t))

		ctx, cancelar := context.WithCancel(t.Context())
		cancelar()

		respuesta, err := cliente.Pedir(ctx, schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: normaGrabada})

		fallo := falloDe(t, err)
		assert.Equal(t, schema.ClaseFuenteNoDisponible, fallo.Clase(),
			"el contexto cancelado corta la operación igual que contra una fuente real (FR-049 a)")
		require.ErrorIs(t, err, context.Canceled, "con su causa intacta")
		assert.Equal(t, Respuesta{}, respuesta)
		assert.Zero(t, espia.busquedas, "y corta antes de buscar nada en el directorio")
	})

	t.Run("identifica la petición", func(t *testing.T) {
		t.Parallel()

		cliente, espia := clienteEspiado(t, grabacionesDePrueba(t))

		// Bajo la identificación va la marca de emisión, que no es una garantía
		// sino una medida —no abre nada ni espera nada—, y bajo ella la
		// reproducción y nada más (contrato httpx-acepta-e-instante §4 de H4).
		escalonDeLaMarca, esLaMarca := espia.siguiente.(*decoradorDeMarcaDeEmision)
		require.True(t, esLaMarca,
			"bajo la identificación está la marca de emisión, que anota la hora al servir cada grabación")

		_, esDeReproduccion := escalonDeLaMarca.siguiente.(*transporteDeReproduccion)
		assert.True(t, esDeReproduccion,
			"y bajo la marca está la reproducción y nada más: ni robots.txt, ni ritmo, ni "+
				"reintentos, ni transporte alguno que pudiera abrir una conexión (FR-046, FR-049)")

		_, err := cliente.Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: normaGrabada})
		require.NoError(t, err)

		assert.Equal(t, AgenteDeUsuario(), espia.agente,
			"la petición reproducida lleva la identificación del proyecto (FR-049 b)")
	})

	t.Run("método", func(t *testing.T) {
		t.Parallel()

		cliente, espia := clienteEspiado(t, grabacionesDePrueba(t))

		respuesta, err := cliente.Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodPost, URL: normaGrabada})

		fallo := falloDe(t, err)
		assert.Equal(t, schema.ClaseArgumentos, fallo.Clase(),
			"un método que no es GET ni HEAD se rechaza también en reproducción (FR-049 c)")
		assert.Equal(t, Respuesta{}, respuesta)
		assert.Zero(t, espia.busquedas, "sin llegar a buscar ninguna grabación")
	})
}

// TestReplaySinRobots es US3 escenario 7 y la mitad de SC-007: la reproducción
// no consulta el robots.txt, ni siquiera cuando el directorio trae su grabación,
// y no espera por ningún ritmo (FR-049 i y ii).
//
// La grabación del robots.txt del directorio lo deniega todo, así que si la
// cadena lo consultara ninguna de estas peticiones se serviría: quedarse sin
// usar no es un error, y que las peticiones salgan es la prueba.
func TestReplaySinRobots(t *testing.T) {
	t.Parallel()

	t.Run("con-grabacion", func(t *testing.T) {
		t.Parallel()

		directorio := grabacionesDePrueba(t)
		ruta := filepath.Join(directorio, ficheroDelRobots)
		require.FileExists(t, ruta, "el directorio trae la grabación del robots.txt (plan.md §«Fixtures», fichero 10)")

		grabado := grabacionLeida(t, ruta)
		require.NotNil(t, grabado.Respuesta.Cuerpo)
		require.Contains(t, *grabado.Respuesta.Cuerpo, "Disallow: /", "y ese robots.txt lo deniega todo")

		exigeQueSirvaTodo(t, directorio)
	})

	t.Run("sin-grabacion", func(t *testing.T) {
		t.Parallel()

		directorio := copiadoSinElRobots(t, grabacionesDePrueba(t))

		require.NoFileExists(t, filepath.Join(directorio, ficheroDelRobots),
			"un directorio preparado a mano no tiene por qué traer ninguna grabación del robots.txt")

		exigeQueSirvaTodo(t, directorio)
	})
}

// TestReplayRedirecciones es US3 escenario 8 y FR-049 d: la cadena de saltos se
// resuelve enteramente dentro del directorio —cada salto es una búsqueda más—,
// quien llama recibe la respuesta final como en red, un destino sin grabación es
// el fallo de FR-047 nombrándolo, y una cadena que vuelve sobre sí misma es «la
// fuente no sabe entregar el recurso», igual que contra una fuente real.
func TestReplayRedirecciones(t *testing.T) {
	t.Parallel()

	t.Run("seguida", func(t *testing.T) {
		t.Parallel()

		directorio := grabacionesDePrueba(t)
		cliente, espia := clienteEspiado(t, directorio)

		respuesta, err := cliente.Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: antiguaGrabada})
		require.NoError(t, err)

		grabada := grabacionLeida(t, filepath.Join(directorio, ficheroDeLaNorma))
		require.NotNil(t, grabada.Respuesta.Cuerpo)

		assert.Equal(t, http.StatusOK, respuesta.Estado, "lo que se entrega es la respuesta final, no el 302")
		assert.Equal(t, normaGrabada, respuesta.URL, "la dirección final resuelve la Location relativa (FR-011)")
		assert.Equal(t, antiguaGrabada, respuesta.Peticion.URL, "y la petición devuelta sigue siendo la pedida")
		assert.Equal(t, *grabada.Respuesta.Cuerpo, string(respuesta.Cuerpo))
		assert.Equal(t, 2, espia.busquedas, "cada salto es una búsqueda más en el directorio (FR-049 d)")
	})

	t.Run("destino ausente", func(t *testing.T) {
		t.Parallel()

		respuesta, err := clienteDeReproduccion(t, grabacionesDePrueba(t)).Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: movidaGrabada})

		fallo := falloDe(t, err)
		assert.Equal(t, schema.ClaseInesperado, fallo.Clase(),
			"si falta la grabación del destino, el fallo es el de la petición ausente (FR-047)")
		assert.Equal(t, Respuesta{}, respuesta)
		assert.Contains(t, fallo.Error(), "GET "+perdidaSinGrabar, "nombrando el salto que no estaba grabado")
	})

	t.Run("bucle", func(t *testing.T) {
		t.Parallel()

		respuesta, err := clienteDeReproduccion(t, grabacionesDePrueba(t)).Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: bucleGrabado})

		fallo := falloDe(t, err)
		assert.Equal(t, schema.ClaseFuenteNoDisponible, fallo.Clase(),
			"una cadena que vuelve sobre sí misma se corta con la clase de siempre (FR-049 d)")
		assert.Equal(t, Respuesta{}, respuesta)
		assert.Contains(t, fallo.Error(), bucleGrabado, "nombrando la dirección que se repite")
	})
}

// TestReplayEstadoDeErrorGrabado es FR-049 e y la otra mitad de SC-007: un
// fixture de 5xx o de 429 produce en el test la misma clase que produciría la
// fuente, y lo produce en la primera y única búsqueda, sin reintentos que
// repitan la lectura y sin esperas de ningún limitador.
func TestReplayEstadoDeErrorGrabado(t *testing.T) {
	t.Parallel()

	t.Run("el 5xx grabado no se reintenta", func(t *testing.T) {
		t.Parallel()

		cliente, espia := clienteEspiado(t, grabacionesDePrueba(t))
		peticion := Peticion{Metodo: http.MethodGet, URL: caidaGrabada}

		comienzo := time.Now()

		for range 2 {
			respuesta, err := cliente.Pedir(t.Context(), schema.Contexto{}, peticion)

			fallo := falloDe(t, err)
			assert.Equal(t, schema.ClaseFuenteNoDisponible, fallo.Clase(),
				"un 5xx grabado se clasifica como lo haría la política de reintentos ya agotada (FR-029)")
			assert.Equal(t, http.StatusServiceUnavailable, fallo.Estado, "con el estado que la fuente dio")
			assert.Zero(t, fallo.Espera, "y sin espera, que es del 429 y no de esto")
			assert.Equal(t, Respuesta{}, respuesta)
		}

		assert.Equal(t, 2, espia.busquedas,
			"una sola búsqueda por petición: un 5xx grabado no repite la lectura (FR-049 iii)")
		assert.Less(t, time.Since(comienzo), intervaloPorOmision,
			"y dos peticiones seguidas al mismo sitio no esperan turno: no hay ritmo (FR-049 ii)")
	})

	t.Run("el 429 grabado trae lo que la fuente dijo que hay que esperar", func(t *testing.T) {
		t.Parallel()

		cliente, espia := clienteEspiado(t, grabacionesDePrueba(t))

		respuesta, err := cliente.Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: limitadaGrabada})

		fallo := falloDe(t, err)
		assert.Equal(t, schema.ClaseLimiteOTos, fallo.Clase(), "el 429 grabado es límite o términos de uso (FR-030)")
		assert.Equal(t, http.StatusTooManyRequests, fallo.Estado)
		assert.True(t, fallo.EsperaConocida, "la grabación trae Retry-After")
		assert.Equal(t, 120*time.Second, fallo.Espera, "que llega a quien llama como dato y no como texto (D20)")
		assert.Equal(t, Respuesta{}, respuesta)
		assert.Equal(t, 1, espia.busquedas, "un 429 no se reintenta nunca")
	})
}

// TestReplayRechazaOpciones es la tabla de construcción de Replay del contrato
// §3: el directorio tiene que existir y ser un directorio, la grabación y la
// reproducción no pueden estar activas a la vez (FR-043), y las opciones que
// solo tienen sentido contra una fuente real se rechazan en vez de aceptarse y
// no hacer nada. Todas son de la clase «argumentos», que quien llama corrige.
//
// Sin t.Parallel() y con cada subprueba declarando el valor de la variable de
// entorno del que depende —la que la enciende y las que exigen que esté
// apagada—: t.Setenv «cannot be used in parallel tests» (go doc testing.T.Setenv)
// y deja además la tabla a salvo de lo que traiga el entorno de quien la
// ejecute, igual que las tablas de la grabación.
func TestReplayRechazaOpciones(t *testing.T) {
	t.Setenv(VariableGrabacion, "")

	directorio := grabacionesDePrueba(t)

	casos := []struct {
		nombre     string
		directorio string
		opciones   []Opcion
		mencion    string
	}{
		{
			nombre:     "el directorio de reproducción no puede ir vacío",
			directorio: "",
			mencion:    "Replay",
		},
		{
			nombre:     "ni nombrar algo que no existe",
			directorio: filepath.Join(directorio, "no-esta"),
			mencion:    "no-esta",
		},
		{
			nombre:     "ni un fichero, que no es un directorio",
			directorio: filepath.Join(directorio, ficheroDelRobots),
			mencion:    ficheroDelRobots,
		},
		{
			nombre:     "la raíz de grabación no tiene sentido en reproducción",
			directorio: directorio,
			opciones:   []Opcion{ConRaizDeGrabacion(directorio)},
			mencion:    "ConRaizDeGrabacion",
		},
		{
			nombre:     "ni el ritmo, que la reproducción no espera nunca",
			directorio: directorio,
			opciones:   []Opcion{ConIntervalo(time.Second)},
			mencion:    "ConIntervalo",
		},
		{
			nombre:     "ni el ritmo común a más de un cliente, que es el mismo escalón",
			directorio: directorio,
			opciones:   []Opcion{ConRitmo(NuevoRitmo(time.Second))},
			mencion:    "ConRitmo",
		},
		{
			nombre:     "ni los intentos, que no se repite ninguna búsqueda",
			directorio: directorio,
			opciones:   []Opcion{ConIntentos(2)},
			mencion:    "ConIntentos",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Setenv(VariableGrabacion, "")

			cliente, err := Replay(caso.directorio, caso.opciones...)

			assert.Nil(t, cliente, "una construcción inválida no produce cliente a medio construir")

			fallo := falloDe(t, err)
			assert.Equal(t, schema.ClaseArgumentos, fallo.Clase(), "lo que quien llama puede corregir (FR-063)")
			assert.Empty(t, fallo.Peticion.URL, "un fallo de construcción no tiene petición implicada (FR-034)")
			assert.Contains(t, fallo.Error(), caso.mencion, "el mensaje nombra lo que falla (FR-034)")
		})
	}

	t.Run("la grabación y la reproducción no pueden estar activas a la vez", func(t *testing.T) {
		t.Setenv(VariableGrabacion, variableActiva)

		cliente, err := Replay(directorio)

		assert.Nil(t, cliente)

		fallo := falloDe(t, err)
		assert.Equal(t, schema.ClaseArgumentos, fallo.Clase(),
			"la combinación sin sentido se rechaza explícitamente y no se resuelve por azar (FR-043)")
		assert.Contains(t, fallo.Error(), VariableGrabacion, "el mensaje nombra la variable")
		assert.Contains(t, fallo.Error(), directorio, "y el directorio de reproducción")
	})

	t.Run("y las opciones que sí valen se aceptan", func(t *testing.T) {
		t.Setenv(VariableGrabacion, "")

		registrador := slog.New(slog.DiscardHandler)

		cliente, err := Replay(directorio, ConFuente(fuenteDePrueba), ConRegistrador(registrador))
		require.NoError(t, err)

		assert.Equal(t, fuenteDePrueba, cliente.fuente)
		assert.Same(t, registrador, cliente.registrador)
		assert.Nil(t, cliente.sitios,
			"un cliente de reproducción no lleva registro de sitios: no hay ritmo ni robots.txt que guardar")
	})
}

// grabacionesDePrueba es el directorio de las diez grabaciones escritas a mano
// de T012, el único que este hito versiona (plan.md §«Fixtures»).
//
// El tramo intermedio de la ruta no se escribe: se localiza bajo testdata/, del
// que el hito declara **un solo** directorio de grabaciones y ninguno más, de
// modo que la búsqueda comprueba de paso esa lista cerrada. Escribirlo sería
// además un nombre que el diccionario de misspell —solo inglés— lee como una
// palabra mal escrita, y este hito no añade ninguna supresión (SC-010).
func grabacionesDePrueba(t *testing.T) string {
	t.Helper()

	encontrados, err := filepath.Glob(filepath.Join("testdata", "*", fuenteDePrueba))
	require.NoError(t, err)
	require.Len(t, encontrados, 1, "el paquete versiona un solo directorio de grabaciones de la fuente de prueba")

	return encontrados[0]
}

// clienteDeReproduccion construye el cliente de reproducción de una tabla. Un
// fallo aquí es de la construcción, no de lo que el test comprueba.
func clienteDeReproduccion(t *testing.T, directorio string, opciones ...Opcion) *Cliente {
	t.Helper()

	cliente, err := Replay(directorio, opciones...)
	require.NoError(t, err, "el cliente de la tabla debe construirse sin error")

	return cliente
}

// clienteEspiado es el mismo cliente con el espía interpuesto entre los dos
// escalones de su cadena: lo que el test mide es la cadena real —la que monta
// Replay— y no una imitación suya.
func clienteEspiado(t *testing.T, directorio string) (*Cliente, *espiaDeLaCadena) {
	t.Helper()

	cliente := clienteDeReproduccion(t, directorio)

	escalonDeIdentificacion, esIdentificador := cliente.cliente.Transport.(*decoradorDeIdentificacion)
	require.True(t, esIdentificador,
		"la cadena de reproducción lleva la identificación arriba del todo (FR-049 b)")

	espia := &espiaDeLaCadena{siguiente: escalonDeIdentificacion.siguiente}
	escalonDeIdentificacion.siguiente = espia

	return cliente, espia
}

// espiaDeLaCadena anota lo que de verdad baja hasta el directorio: cuántas
// búsquedas se hacen —una por petición, porque no hay reintentos que la
// repitan— y con qué identificación llegan. No responde por su cuenta: delega en
// el escalón que envuelve.
//
// No necesita exclusión: el cliente atiende cada petición en la goroutine de
// quien la pide, y ninguna tabla de este fichero usa el espía desde varias.
type espiaDeLaCadena struct {
	siguiente http.RoundTripper
	busquedas int
	agente    string
}

// RoundTrip anota la petición y la entrega al escalón siguiente.
func (e *espiaDeLaCadena) RoundTrip(peticion *http.Request) (*http.Response, error) {
	e.busquedas++
	e.agente = peticion.Header.Get("User-Agent")

	return e.siguiente.RoundTrip(peticion)
}

// exigeQueSirvaTodo pide las grabaciones que tiene que servir cualquier
// directorio de estas tablas y comprueba que ninguna falla y que las tres
// seguidas no esperan turno, que es lo que distingue «no consulta el robots.txt
// ni el limitador» de «lo consulta y da la casualidad de que permite».
func exigeQueSirvaTodo(t *testing.T, directorio string) {
	t.Helper()

	cliente := clienteDeReproduccion(t, directorio)
	comienzo := time.Now()

	for _, direccion := range []string{normaGrabada, documentoGrabado, antiguaGrabada} {
		respuesta, err := cliente.Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: direccion})

		require.NoError(t, err, "la grabación de %s se sirve sin consultar ningún robots.txt (FR-049 i)", direccion)
		assert.Equal(t, http.StatusOK, respuesta.Estado)
		assert.NotEmpty(t, respuesta.Cuerpo)
	}

	assert.Less(t, time.Since(comienzo), intervaloPorOmision,
		"y las tres del mismo sitio se sirven sin esperar ningún turno (FR-049 ii, US3 escenario 7)")
}

// copiadoSinElRobots deja en un directorio temporal las grabaciones del
// directorio que recibe menos la del robots.txt: el directorio preparado a mano
// de US3 escenario 7, que no tiene por qué traerla.
func copiadoSinElRobots(t *testing.T, origen string) string {
	t.Helper()

	destino := t.TempDir()

	for _, entrada := range entradasDe(t, origen) {
		if entrada.Name() == ficheroDelRobots {
			continue
		}

		require.NoError(t, os.WriteFile(filepath.Join(destino, entrada.Name()),
			contenidoDe(t, filepath.Join(origen, entrada.Name())), permisoDeLaCopia))
	}

	return destino
}

// La petición y el fichero de la tabla de grabaciones que no se pueden
// reproducir: una dirección que no está en el directorio de T012 y el nombre que
// le corresponde según el contrato §2, escrito a mano.
const (
	direccionEscritaAMano = sitioGrabado + "/rara"
	ficheroEscritoAMano   = "GET_http_fuente.prueba_rara.json"
)

// grabacionEscritaAMano es un fichero con la forma del contrato §1 y con el
// formato y la respuesta que cada fila necesita, para provocar los fallos de la
// tabla de emparejamiento sin tocar ninguna de las diez grabaciones del
// repositorio.
func grabacionEscritaAMano(formato int, respuesta string) string {
	return `{
  "formato": ` + strconv.Itoa(formato) + `,
  "grabado_en": "2026-09-12T00:00:00Z",
  "peticion": {
    "metodo": "GET",
    "url": "` + direccionEscritaAMano + `",
    "cabeceras": {}
  },
  "respuesta": ` + respuesta + `
}
`
}
