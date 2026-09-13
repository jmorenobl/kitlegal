package httpx

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// TestNewSinOpciones comprueba que el cliente nace con sus garantías puestas:
// sin declarar nada se obtiene uno que no impone plazo propio y que descarta los eventos en
// vez de escribirlos por su cuenta, que es lo que impide que un paquete sin
// registrador acabe hablando por slog.Default() (FR-035, D15).
func TestNewSinOpciones(t *testing.T) {
	t.Parallel()

	cliente, err := New()
	require.NoError(t, err)
	require.NotNil(t, cliente)

	assert.Empty(t, cliente.fuente, "sin ConFuente el cliente no nombra ninguna fuente")

	require.NotNil(t, cliente.registrador, "el registrador nunca es nulo (D15)")
	assert.False(t, cliente.registrador.Enabled(t.Context(), slog.LevelError),
		"el registrador por omisión descarta todo lo que se le entregue")

	require.NotNil(t, cliente.cliente)
	assert.Zero(t, cliente.cliente.Timeout, "el plazo es el del contexto (FR-004)")

	require.NotNil(t, cliente.sitios, "el cliente tiene un solo registro de sitios (data-model.md §4, D14)")
	assert.Equal(t, time.Second, intervaloPorOmision,
		"el ritmo por omisión es conservador: una petición por segundo y sitio (FR-020, D8)")
	assert.Equal(t, intervaloPorOmision, cliente.sitios.intervalo,
		"sin ConIntervalo los sitios de este cliente nacen con el ritmo por omisión")
	assert.Equal(t, 3, intentosPorOmision,
		"sin ConIntentos son tres: una petición y dos reintentos (FR-024, D9)")
}

// TestOpcionesInvalidas fija la regla de validez de cada opción y dónde se
// entrega el fallo: en la construcción, no en la opción, que es el patrón que
// permite que el error llegue una sola vez y con su clase (golang-design-patterns).
// Un fallo de configuración no tiene petición implicada, así que el mensaje
// nombra en su lugar la opción que falla (FR-034).
func TestOpcionesInvalidas(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre  string
		opcion  Opcion
		mencion string
	}{
		{
			nombre:  "la fuente no puede ir vacía",
			opcion:  ConFuente(""),
			mencion: "ConFuente",
		},
		{
			nombre:  "la fuente no puede llevar separadores de ruta",
			opcion:  ConFuente("boe/estatal"),
			mencion: "ConFuente",
		},
		{
			nombre:  "la fuente no puede llevar separadores de Windows",
			opcion:  ConFuente(`boe\estatal`),
			mencion: "ConFuente",
		},
		{
			nombre:  "la fuente no puede subir de directorio",
			opcion:  ConFuente(".."),
			mencion: "ConFuente",
		},
		{
			nombre:  "el intervalo entre peticiones no puede ser nulo",
			opcion:  ConIntervalo(0),
			mencion: "ConIntervalo",
		},
		{
			nombre:  "el intervalo entre peticiones no puede ser negativo",
			opcion:  ConIntervalo(-time.Second),
			mencion: "ConIntervalo",
		},
		{
			nombre:  "no se puede pedir menos de un intento",
			opcion:  ConIntentos(0),
			mencion: "ConIntentos",
		},
		{
			nombre:  "ni un número negativo de intentos",
			opcion:  ConIntentos(-1),
			mencion: "ConIntentos",
		},
		{
			nombre:  "el registrador no puede ser nulo",
			opcion:  ConRegistrador(nil),
			mencion: "ConRegistrador",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			cliente, err := New(caso.opcion)

			assert.Nil(t, cliente, "una opción inválida no produce cliente a medio construir")

			fallo := falloDe(t, err)
			assert.Equal(t, schema.ClaseArgumentos, fallo.Clase(),
				"una configuración mal formada la puede corregir quien llama (FR-063)")
			assert.Empty(t, fallo.Peticion.URL, "un fallo de configuración no tiene petición implicada (FR-034)")
			assert.Contains(t, fallo.Error(), caso.mencion, "el mensaje nombra la opción que falla (FR-034)")
		})
	}

	t.Run("una opción válida se acepta y la última repetida gana", func(t *testing.T) {
		t.Parallel()

		registrador := slog.New(slog.DiscardHandler)

		cliente, err := New(ConFuente("placsp"), ConFuente("boe"), ConRegistrador(registrador),
			ConIntervalo(time.Minute), ConIntervalo(2*time.Second))
		require.NoError(t, err)

		assert.Equal(t, "boe", cliente.fuente)
		assert.Same(t, registrador, cliente.registrador)
		assert.Equal(t, 2*time.Second, cliente.sitios.intervalo,
			"el intervalo declarado es el que llevan los sitios de ese cliente (FR-020)")
	})
}

// TestPedirRechazaMetodo comprueba la restricción de la constitución §I: no
// existe en el módulo ninguna llamada HTTP con método distinto de GET o HEAD.
// El rechazo es de argumentos y llega **sin abrir ninguna conexión**, que es lo
// que el contador del servidor demuestra (FR-010, tabla §3 fila 9).
func TestPedirRechazaMetodo(t *testing.T) {
	t.Parallel()

	metodos := []struct {
		nombre string
		metodo string
	}{
		{nombre: "POST no se emite", metodo: http.MethodPost},
		{nombre: "PUT no se emite", metodo: http.MethodPut},
		{nombre: "DELETE no se emite", metodo: http.MethodDelete},
		{nombre: "el método va en mayúsculas y solo en mayúsculas", metodo: "get"},
		{nombre: "un método vacío no es GET por omisión", metodo: ""},
	}

	servidor, contador := servidorIdentificado(t, redireccionesDePrueba())
	cliente := clienteDePrueba(t)

	for _, caso := range metodos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			exigeArgumentosSinConexion(t, cliente, contador, schema.Contexto{},
				Peticion{Metodo: caso.metodo, URL: servidor.URL + "/norma"})
		})
	}
}

// TestPedirRechazaDireccion fija la otra comprobación previa: la dirección tiene
// que ser absoluta, de esquema http o https y nombrar un sitio. Lo que no lo es
// no es «fuente no disponible» —no hay ninguna fuente a la que culpar— sino algo
// que quien llama puede corregir (research D19, tabla §3 fila 10).
func TestPedirRechazaDireccion(t *testing.T) {
	t.Parallel()

	direcciones := []struct {
		nombre    string
		direccion string
	}{
		{nombre: "la dirección vacía", direccion: ""},
		{nombre: "una ruta sin sitio no es absoluta", direccion: "/norma?id=1"},
		{nombre: "un esquema que no es de red", direccion: "ftp://fuente.prueba/norma"},
		{nombre: "un esquema de fichero tampoco", direccion: "file:///etc/hosts"},
		{nombre: "un esquema http sin sitio", direccion: "http:///norma"},
		{nombre: "una dirección que no se puede interpretar", direccion: "://fuente.prueba"},
	}

	_, contador := servidorIdentificado(t, redireccionesDePrueba())
	cliente := clienteDePrueba(t)

	for _, caso := range direcciones {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			exigeArgumentosSinConexion(t, cliente, contador, schema.Contexto{},
				Peticion{Metodo: http.MethodGet, URL: caso.direccion})
		})
	}
}

// TestPedirConAcepta fija la única cabecera que elige quien pide: el formato del
// recurso, que boe.py pide por Accept —XML para el texto de un bloque y JSON para
// lo demás (FR-003 de H4)—. Va en la petición y en cada salto de su cadena de
// redirecciones, se graba con las demás cabeceras de la petición y no la lleva
// nunca la del robots.txt, que el paquete pide por su cuenta. Vacía, la petición
// sale como en H2; mal formada, es de argumentos y no abre nada, tampoco en
// ensayo. Y la reproducción sigue emparejando solo por método y dirección
// (contrato httpx-acepta-e-instante §1, research D4 de H4).
//
// Ninguna subprueba declara t.Parallel(), y no es un descuido: todas usan
// t.Setenv, que «cannot be used in parallel tests» (go doc testing.T.Setenv). La
// de la grabación la enciende; las demás la declaran apagada, porque New y Replay
// leen la variable y lo que cada una comprueba no puede depender del entorno de
// quien ejecuta los tests —Replay, además, rechaza cualquier valor no vacío
// (FR-043)—. Es la regla de las tablas de grabar_test.go y reproducir_test.go, y
// no hay carrera con el resto del paquete porque las tablas paralelas no arrancan
// hasta que las secuenciales han terminado.
func TestPedirConAcepta(t *testing.T) {
	const formato = "application/xml"

	t.Run("la cabecera va en la petición y en cada salto, y no en la del robots.txt", func(t *testing.T) {
		t.Setenv(VariableGrabacion, "")

		servidor, formatos := servidorQueAnotaFormatos(t)

		respuesta, err := clienteDePrueba(t, ConIntervalo(time.Millisecond)).Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/antigua", Acepta: formato})
		require.NoError(t, err)
		require.Equal(t, servidor.URL+"/norma", respuesta.URL, "la cadena de la tabla tiene un salto")

		assert.Equal(t, []string{formato}, formatos.de(t, "/antigua"), "la petición pedida lleva el formato tal cual")
		assert.Equal(t, []string{formato}, formatos.de(t, "/norma"),
			"y el salto también: cada uno es una petición completa que no hereda nada del anterior (D10)")
		assert.Nil(t, formatos.de(t, rutaDelRobots),
			"la del robots.txt no: la fabrica el paquete y no es ninguno de los recursos pedidos")
	})

	t.Run("la grabación la guarda en cada salto, y no en la del robots.txt", func(t *testing.T) {
		t.Setenv(VariableGrabacion, variableActiva)

		raiz := t.TempDir()
		servidor, _ := servidorIdentificado(t, redireccionesDePrueba())

		_, err := clienteQueGraba(t, raiz).Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/antigua", Acepta: formato})
		require.NoError(t, err, "grabar no cambia el resultado de la petición")

		for _, salto := range []string{"_antigua", "_norma"} {
			grabada := grabacionLeida(t, rutaGrabada(raiz, nombreDelServidor(t, servidor, "GET", salto)))
			assert.Equal(t, []string{formato}, grabada.Peticion.Cabeceras["Accept"],
				"la grabación guarda las cabeceras de la petición tal como salió, el formato incluido (FR-037)")
		}

		delRobots := grabacionLeida(t, rutaGrabada(raiz, nombreDelServidor(t, servidor, "GET", "_robots.txt")))
		assert.Contains(t, delRobots.Peticion.Cabeceras, "User-Agent", "la del robots.txt se graba con sus cabeceras")
		assert.NotContains(t, delRobots.Peticion.Cabeceras, "Accept", "y entre ellas no está el formato del recurso")
	})

	t.Run("sin formato la petición no lleva Accept", func(t *testing.T) {
		t.Setenv(VariableGrabacion, "")

		servidor, formatos := servidorQueAnotaFormatos(t)

		_, err := clienteDePrueba(t, ConIntervalo(time.Millisecond)).Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/antigua"})
		require.NoError(t, err)

		assert.Nil(t, formatos.de(t, "/antigua"), "vacío, la petición sale como en H2")
		assert.Nil(t, formatos.de(t, "/norma"), "y sus saltos también")
	})

	formatosMalFormados := []struct {
		nombre string
		acepta string
	}{
		{nombre: "un tipo sin subtipo", acepta: "application/"},
		{nombre: "una lista de tipos", acepta: formato + ", application/json"},
		{nombre: "solo espacios", acepta: "   "},
		// Los tres siguientes los acepta mime.ParseMediaType, que recorta los
		// espacios que rodean el tipo —un «\r\n» final incluido—, admite
		// tabuladores entre parámetros y cualquier byte salvo el retorno de carro y
		// el salto de línea dentro de un valor entrecomillado: son los que
		// demuestran que la comprobación de los caracteres de control no sobra.
		{nombre: "un fin de línea tras el tipo", acepta: formato + "\r\n"},
		{nombre: "un tabulador entre parámetros", acepta: formato + ";\tcharset=utf-8"},
		{nombre: "un carácter de control entrecomillado", acepta: formato + `; charset="utf` + "\x01" + `-8"`},
	}

	for _, caso := range formatosMalFormados {
		t.Run("un formato mal formado es de argumentos y no abre nada: "+caso.nombre, func(t *testing.T) {
			t.Setenv(VariableGrabacion, "")

			servidor, contador := servidorIdentificado(t, redireccionesDePrueba())
			cliente := clienteDePrueba(t)

			for _, ejecucion := range []schema.Contexto{{}, {DryRun: true}} {
				respuesta, err := cliente.Pedir(t.Context(), ejecucion,
					Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/norma", Acepta: caso.acepta})

				exigeClaseYCodigo(t, schema.ClaseArgumentos, 2, err)
				assert.Equal(t, Respuesta{}, respuesta, "un rechazo no entrega ninguna respuesta, tampoco la del ensayo")
				require.ErrorContains(t, err, "Accept", "el mensaje nombra la cabecera que no se puede pedir (FR-034)")
			}

			assert.Zero(t, contador.total.Load(), "un formato rechazado no abre ninguna conexión, tampoco en ensayo")
			assert.Zero(t, contador.robots.Load(), "ni siquiera la del robots.txt del sitio")
		})
	}

	t.Run("un formato válido no cambia la descripción del ensayo", func(t *testing.T) {
		t.Setenv(VariableGrabacion, "")

		pedida := Peticion{Metodo: peticionDeEnsayo.Metodo, URL: peticionDeEnsayo.URL, Acepta: formato}

		respuesta, err := clienteDePrueba(t).Pedir(t.Context(), schema.Contexto{DryRun: true}, pedida)
		require.NoError(t, err)

		assert.Equal(t, pedida, respuesta.Peticion, "la petición devuelta es la que se pidió, con su formato")
		assert.Equal(t, "GET http://fuente.prueba/norma?id=BOE-A-2015-10565", respuesta.Descripcion(),
			"la descripción sigue siendo «<Metodo> <URL>»: el formato no la cambia")
	})

	t.Run("la reproducción empareja solo por método y dirección", func(t *testing.T) {
		t.Setenv(VariableGrabacion, "")

		directorio := grabacionesDePrueba(t)
		grabada := grabacionLeida(t, filepath.Join(directorio, ficheroDeLaNorma))
		require.NotContains(t, grabada.Peticion.Cabeceras, "Accept", "la grabación de la tabla no declara ningún formato")
		require.NotNil(t, grabada.Respuesta.Cuerpo)

		cliente := clienteDeReproduccion(t, directorio)

		for _, acepta := range []string{"", formato, "application/json"} {
			respuesta, err := cliente.Pedir(t.Context(), schema.Contexto{},
				Peticion{Metodo: http.MethodGet, URL: normaGrabada, Acepta: acepta})
			require.NoError(t, err,
				"el formato %q no participa en el emparejamiento (contrato de grabación de H2 §4)", acepta)
			assert.Equal(t, *grabada.Respuesta.Cuerpo, string(respuesta.Cuerpo), "y se sirve la misma grabación")
		}
	})
}

// TestPedirRespetaElPlazo comprueba SC-002: el plazo de la operación es el del
// contexto y de nadie más, y su vencimiento corta en ese instante —no más tarde—
// con la clase «fuente no disponible» y con la causa del contexto intacta, de
// modo que quien llama pueda reconocerla con errors.Is (FR-004, FR-005, FR-029).
func TestPedirRespetaElPlazo(t *testing.T) {
	t.Parallel()

	const plazo = 50 * time.Millisecond

	servidor, _ := servidorIdentificado(t, func(_ http.ResponseWriter, peticion *http.Request) {
		// El servidor no responde nunca por su cuenta: quien termina la
		// operación es el contexto de quien la pidió.
		<-peticion.Context().Done()
	})

	ctx, cancelar := context.WithTimeout(t.Context(), plazo)
	defer cancelar()

	comienzo := time.Now()

	// El ritmo se aparta del camino: con el intervalo por omisión, la petición
	// del robots.txt del sitio ocuparía el turno recién nacido y la del recurso
	// se quedaría sin él dentro del plazo, de modo que lo que cortaría la
	// operación sería el limitador (FR-022) y no el plazo que este test mide.
	_, err := clienteDePrueba(t, ConIntervalo(time.Millisecond)).Pedir(ctx, schema.Contexto{},
		Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/lenta"})

	fallo := falloDe(t, err)
	assert.Equal(t, schema.ClaseFuenteNoDisponible, fallo.Clase(),
		"el vencimiento del plazo es «fuente no disponible» (FR-029, tabla §3 fila 2)")
	require.ErrorIs(t, err, context.DeadlineExceeded, "la causa del contexto llega intacta hasta quien llama")
	assert.Less(t, time.Since(comienzo), time.Second, "la operación termina con el contexto, no más tarde (FR-005)")
}

// TestPedirSigueRedirecciones comprueba FR-011: lo que recibe quien llama es la
// respuesta final de la cadena, nunca el 3xx intermedio; la dirección final es
// la que de verdad entregó el contenido y la petición devuelta sigue siendo la
// que se pidió. Cada salto es una petición completa, con su identificación y su
// método conservado (D10).
func TestPedirSigueRedirecciones(t *testing.T) {
	t.Parallel()

	t.Run("una Location relativa se resuelve sobre la dirección del salto", func(t *testing.T) {
		t.Parallel()

		servidor, contador := servidorIdentificado(t, redireccionesDePrueba())

		respuesta, err := clienteDePrueba(t).Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/antigua"})
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, respuesta.Estado)
		assert.Equal(t, servidor.URL+"/norma", respuesta.URL, "la dirección final es la que citará el sobre (FR-011)")
		assert.Equal(t, servidor.URL+"/antigua", respuesta.Peticion.URL, "la petición devuelta es la que se pidió")
		assert.Equal(t, contenidoDePrueba, string(respuesta.Cuerpo))
		assert.Equal(t, "application/xml; charset=utf-8", respuesta.Cabeceras.Get("content-type"))
		assert.Equal(t, int64(2), contador.total.Load())
	})

	t.Run("el salto puede llevar a otro sitio, que no hereda nada del origen", func(t *testing.T) {
		t.Parallel()

		destino, contadorDelDestino := servidorIdentificado(t, redireccionesDePrueba())
		origen, _ := servidorIdentificado(t, func(escritor http.ResponseWriter, peticion *http.Request) {
			http.Redirect(escritor, peticion, destino.URL+"/norma", http.StatusMovedPermanently)
		})

		respuesta, err := clienteDePrueba(t).Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: origen.URL + "/antigua"})
		require.NoError(t, err)

		assert.Equal(t, destino.URL+"/norma", respuesta.URL)
		assert.Equal(t, int64(1), contadorDelDestino.total.Load())
		assert.Zero(t, contadorDelDestino.sinIdentificar.Load(),
			"el sitio de destino recibe la identificación igual que el de origen (FR-009)")
	})

	t.Run("el método se conserva en el salto", func(t *testing.T) {
		t.Parallel()

		var cabezas atomic.Int64

		servidor, contador := servidorIdentificado(t, func(escritor http.ResponseWriter, peticion *http.Request) {
			if peticion.Method == http.MethodHead {
				cabezas.Add(1)
			}

			redireccionesDePrueba()(escritor, peticion)
		})

		respuesta, err := clienteDePrueba(t).Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodHead, URL: servidor.URL + "/antigua"})
		require.NoError(t, err)

		assert.Equal(t, int64(2), contador.total.Load())
		assert.Equal(t, int64(2), cabezas.Load(), "HEAD sigue HEAD en todos los saltos (D10)")
		assert.Empty(t, respuesta.Cuerpo, "una petición HEAD no trae cuerpo")
	})

	t.Run("la respuesta final se entrega aunque no sea un éxito", func(t *testing.T) {
		t.Parallel()

		servidor, _ := servidorIdentificado(t, redireccionesDePrueba())

		respuesta, err := clienteDePrueba(t).Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/movida"})
		require.NoError(t, err, "«no encontrado» no es un fallo del cliente: lo entrega y lo interpreta la fuente (FR-032)")

		assert.Equal(t, http.StatusNotFound, respuesta.Estado)
		assert.Equal(t, servidor.URL+"/perdida", respuesta.URL)
	})
}

// TestPedirCortaCadenasDeRedirecciones comprueba lo que impide que una cadena
// termine en una espera indefinida: el bucle se corta en cuanto una dirección se
// repite, el tope acota la cadena que no se repite, y un 3xx que no dice a dónde
// ir —o que no es de los cinco que se siguen— no se entrega. Los cuatro son «la
// fuente no sabe entregar el recurso» (FR-011, FR-032, tabla §3 fila 3).
func TestPedirCortaCadenasDeRedirecciones(t *testing.T) {
	t.Parallel()

	t.Run("una cadena que vuelve sobre sí misma se corta en el primer repetido", func(t *testing.T) {
		t.Parallel()

		servidor, contador := servidorIdentificado(t, func(escritor http.ResponseWriter, peticion *http.Request) {
			http.Redirect(escritor, peticion, "/bucle", http.StatusFound)
		})

		exigeFuenteNoDisponible(t, servidor.URL+"/bucle", "/bucle")

		assert.Equal(t, int64(1), contador.total.Load(),
			"el bucle se detecta antes de emitir la petición repetida, sin agotar el tope")
	})

	t.Run("una cadena sin fin se corta en el tope de saltos", func(t *testing.T) {
		t.Parallel()

		var visitas atomic.Int64

		servidor, contador := servidorIdentificado(t, func(escritor http.ResponseWriter, peticion *http.Request) {
			http.Redirect(escritor, peticion, "/salto/"+strconv.FormatInt(visitas.Add(1), 10), http.StatusFound)
		})

		exigeFuenteNoDisponible(t, servidor.URL+"/salto/0")

		assert.Equal(t, int64(topeDeRedirecciones+1), contador.total.Load(),
			"la petición inicial más los diez saltos del tope, y ni una más (D10)")
	})

	t.Run("una redirección sin Location no se puede seguir", func(t *testing.T) {
		t.Parallel()

		servidor, contador := servidorIdentificado(t, func(escritor http.ResponseWriter, _ *http.Request) {
			escritor.WriteHeader(http.StatusFound)
		})

		exigeFuenteNoDisponible(t, servidor.URL+"/incompleta")

		assert.Equal(t, int64(1), contador.total.Load())
	})

	t.Run("un 3xx que no es de los cinco que se siguen tampoco se entrega", func(t *testing.T) {
		t.Parallel()

		servidor, contador := servidorIdentificado(t, func(escritor http.ResponseWriter, _ *http.Request) {
			escritor.Header().Set("Location", "/norma")
			escritor.WriteHeader(http.StatusMultipleChoices)
		})

		exigeFuenteNoDisponible(t, servidor.URL+"/eleccion")

		assert.Equal(t, int64(1), contador.total.Load())
	})
}

// TestPedirDesdeVariasGoroutines es el control literal de FR-057 y de SC-011: el
// mismo cliente, usado a la vez desde varias goroutines contra el mismo sitio,
// no tiene ninguna carrera de datos —el test se ejecuta con el detector
// activado— y sigue pidiendo el robots.txt de ese sitio **una sola vez**, porque
// la exclusión del sitio se mantiene tomada mientras se obtiene: las demás
// esperan a la primera en vez de repetirla (FR-021, SC-003, D7).
func TestPedirDesdeVariasGoroutines(t *testing.T) {
	t.Parallel()

	const goroutines = 8

	servidor, contador := servidorIdentificado(t, redireccionesDePrueba())
	cliente := clienteDePrueba(t, ConIntervalo(time.Millisecond))

	var simultaneas sync.WaitGroup

	simultaneas.Add(goroutines)

	for consulta := range goroutines {
		go func() {
			defer simultaneas.Done()

			respuesta, err := cliente.Pedir(t.Context(), schema.Contexto{},
				Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/norma?n=" + strconv.Itoa(consulta)})

			assert.NoError(t, err, "cada goroutine recibe su respuesta")
			assert.Equal(t, http.StatusOK, respuesta.Estado)
		}()
	}

	simultaneas.Wait()

	assert.Equal(t, int64(goroutines), contador.total.Load(), "ninguna petición se pierde ni se duplica")
	assert.Equal(t, int64(1), contador.robots.Load(),
		"y el robots.txt del sitio se pide una sola vez aunque se pida desde varias goroutines a la vez (SC-003)")
	assert.Zero(t, contador.sinIdentificar.Load())
}

// clienteDePrueba construye el cliente contra la red que usan estas tablas. Un
// fallo aquí es de la construcción, no de lo que el test comprueba.
func clienteDePrueba(t *testing.T, opciones ...Opcion) *Cliente {
	t.Helper()

	cliente, err := New(opciones...)
	require.NoError(t, err, "el cliente de la tabla debe construirse sin error")

	return cliente
}

// falloDe recupera el *Error del paquete, que es el único que produce y el que
// lleva la clase con la que el kernel decide el código de salida (FR-063).
func falloDe(t *testing.T, err error) *Error {
	t.Helper()

	var fallo *Error
	require.ErrorAs(t, err, &fallo, "toda ruta de fallo del paquete produce un *httpx.Error")

	return fallo
}

// exigeArgumentosSinConexion comprueba las dos mitades de un rechazo previo: que
// la clase es «argumentos» y que el servidor no ha recibido nada, que es lo que
// distingue «rechazada antes de abrir nada» de «emitida y luego descartada»
// (FR-010, FR-050, research D19).
func exigeArgumentosSinConexion(
	t *testing.T, cliente *Cliente, contador *contadorDeIdentificacion,
	ejecucion schema.Contexto, p Peticion,
) {
	t.Helper()

	respuesta, err := cliente.Pedir(t.Context(), ejecucion, p)

	fallo := falloDe(t, err)
	assert.Equal(t, schema.ClaseArgumentos, fallo.Clase(), "lo que quien llama puede corregir es «argumentos»")
	assert.Equal(t, Respuesta{}, respuesta, "un rechazo no entrega ninguna respuesta")
	assert.Zero(t, contador.total.Load(), "una petición rechazada no abre ninguna conexión")
}

// exigeFuenteNoDisponible pide la dirección con el cliente por omisión y exige
// que el fallo sea de la clase «fuente no disponible» y que su mensaje nombre lo
// que se le indique.
func exigeFuenteNoDisponible(t *testing.T, direccion string, menciones ...string) {
	t.Helper()

	respuesta, err := clienteDePrueba(t).Pedir(t.Context(), schema.Contexto{},
		Peticion{Metodo: http.MethodGet, URL: direccion})

	fallo := falloDe(t, err)
	assert.Equal(t, schema.ClaseFuenteNoDisponible, fallo.Clase(),
		"una cadena que no se puede seguir es «la fuente no sabe entregar el recurso» (FR-011)")
	assert.Equal(t, Respuesta{}, respuesta, "un 3xx no se entrega nunca a quien llama (FR-032)")

	for _, mencion := range menciones {
		assert.Contains(t, fallo.Error(), mencion, "el mensaje nombra lo que ha fallado (FR-034)")
	}
}

// formatosPedidos anota, por ruta, los valores de la cabecera Accept con que
// llegó la petición a esa ruta. Guarda nil cuando llegó sin ella, que es lo que
// distingue «pedida sin formato» de «pedida con un formato vacío»; que la ruta no
// llegara a pedirse es un fallo de la tabla, y no una cabecera ausente.
type formatosPedidos struct {
	mu      sync.Mutex
	porRuta map[string][]string
}

// servidorQueAnotaFormatos levanta el servidor de las redirecciones de prueba,
// con un robots.txt sin reglas, y anota el formato de toda petición que recibe,
// la del robots.txt incluida.
func servidorQueAnotaFormatos(t *testing.T) (*httptest.Server, *formatosPedidos) {
	t.Helper()

	formatos := &formatosPedidos{porRuta: make(map[string][]string)}
	servidor, _ := servidorConRobots(t, formatos.anota(robotsQueDice("")), formatos.anota(redireccionesDePrueba()))

	return servidor, formatos
}

// anota envuelve un manejador para guardar el formato de cada petición antes de
// atenderla.
func (f *formatosPedidos) anota(manejador http.HandlerFunc) http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		f.mu.Lock()
		f.porRuta[peticion.URL.Path] = slices.Clone(peticion.Header.Values("Accept"))
		f.mu.Unlock()

		manejador(escritor, peticion)
	}
}

// de devuelve los valores de Accept con que llegó la petición a la ruta, y exige
// que llegara.
func (f *formatosPedidos) de(t *testing.T, ruta string) []string {
	t.Helper()

	f.mu.Lock()
	defer f.mu.Unlock()

	valores, pedida := f.porRuta[ruta]
	require.True(t, pedida, "el servidor no ha recibido ninguna petición a %s", ruta)

	return valores
}
