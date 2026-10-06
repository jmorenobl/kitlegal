package httpx

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"sync/atomic"
	"testing"
	"testing/iotest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Las tablas de este fichero que construyen un cliente no declaran t.Parallel(),
// y no es un descuido: todas usan t.Setenv, que «cannot be used in parallel
// tests» (go doc testing.T.Setenv). No hay carrera con el resto del paquete
// porque las tablas paralelas —las de los demás ficheros y la única de este que
// no toca el entorno, TestGrabarRetiraElTemporal— no arrancan hasta que las
// secuenciales han terminado, de modo que ninguna construye su cliente con la
// variable puesta aquí (D12).

// fuenteDePrueba es el nombre lógico de la fuente de estas tablas: el que nombra
// el directorio que cuelga de la raíz de grabación.
const fuenteDePrueba = "prueba"

// variableActiva es el único valor que enciende la grabación. Se escribe aquí
// tal cual, y no con la constante del paquete, porque es parte de lo que la
// tabla comprueba: si el valor que activa cambiara, estas tablas tienen que
// dejar de pasar (FR-041).
const variableActiva = "1"

// TestGrabarEscribeElFichero es el control de SC-008 y de la forma entera del
// contrato de grabación §1: con la variable puesta y la raíz declarada, la
// petición queda en <raíz>/<fuente>/<nombre>.json con las cuatro claves en su
// orden, indentada con dos espacios, sin escapar «<» ni «>», con las cabeceras
// canónicas y ordenadas y con el cuerpo como texto (FR-036, FR-037, FR-040).
//
// Se compara el fichero **entero** contra el literal del contrato, neutralizando
// solo lo que declara cuándo se grabó, porque cualquier comprobación campo a
// campo dejaría fuera precisamente lo que hace falta que sea estable: el orden y
// la forma.
func TestGrabarEscribeElFichero(t *testing.T) {
	t.Setenv(VariableGrabacion, variableActiva)

	raiz := t.TempDir()
	servidor, _ := servidorIdentificado(t, redireccionesDePrueba())
	direccion := servidor.URL + "/norma?id=BOE-A-2015-10565"

	respuesta, err := clienteQueGraba(t, raiz).Pedir(t.Context(), schema.Contexto{},
		Peticion{Metodo: http.MethodGet, URL: direccion})
	require.NoError(t, err, "grabar no cambia el resultado de la petición")
	require.Equal(t, http.StatusOK, respuesta.Estado)

	esperado := `{
  "formato": 1,
  "grabado_en": "` + fechaNeutralizada + `",
  "peticion": {
    "metodo": "GET",
    "url": "` + direccion + `",
    "cabeceras": {
      "User-Agent": [
        "` + AgenteDeUsuario() + `"
      ]
    }
  },
  "respuesta": {
    "estado": 200,
    "cabeceras": {
      "Content-Length": [
        "` + strconv.Itoa(len(contenidoDePrueba)) + `"
      ],
      "Content-Type": [
        "application/xml; charset=utf-8"
      ],
      "Date": [
        "` + fechaNeutralizada + `"
      ]
    },
    "cuerpo": "` + contenidoDePrueba + `"
  }
}
`

	ruta := rutaGrabada(raiz, nombreDelServidor(t, servidor, "GET", "_norma_q_id_BOE-A-2015-10565"))
	assert.Equal(t, esperado, neutralizado(t, ruta),
		"el fichero es el objeto del contrato §1, con sus claves en orden y sin escape de HTML")

	grabada := grabacionLeida(t, ruta)
	assert.Equal(t, 1, grabada.Formato)

	fecha, err := time.Parse(time.RFC3339, grabada.GrabadoEn)
	require.NoError(t, err, "grabado_en es RFC 3339")
	assert.Equal(t, time.UTC, fecha.Location(), "y va en UTC")
	assert.Equal(t, fecha.Truncate(time.Second), fecha, "con segundos enteros")

	// El robots.txt del sitio lo pidió el propio cliente, no quien llamó, y se
	// graba igual: por eso nadie declara el nombre del fichero (FR-039). Su
	// respuesta va sin cuerpo, que el contrato §1 manda grabar como cuerpo vacío
	// y nunca como base64.
	delRobots := grabacionLeida(t, rutaGrabada(raiz, nombreDelServidor(t, servidor, "GET", "_robots.txt")))
	require.NotNil(t, delRobots.Respuesta.Cuerpo, "un cuerpo vacío sigue siendo texto")
	assert.Empty(t, *delRobots.Respuesta.Cuerpo)
	assert.Nil(t, delRobots.Respuesta.CuerpoBase64, "«cuerpo» y «cuerpo_base64» son excluyentes")
}

// TestGrabarNoAlteraLaRespuesta es FR-038: el decorador lee el cuerpo para
// grabarlo, pero quien llamó recibe exactamente lo mismo que recibiría sin
// grabar, íntegro y una sola vez. Se compara contra la respuesta del mismo
// servidor sin la variable puesta, que es la única forma de comprobar «lo
// mismo» y no «lo que el test creía».
func TestGrabarNoAlteraLaRespuesta(t *testing.T) {
	t.Setenv(VariableGrabacion, "")

	servidor, _ := servidorIdentificado(t, redireccionesDePrueba())
	peticion := Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/norma"}

	sinGrabar, err := clienteDePrueba(t, ConIntervalo(time.Millisecond)).
		Pedir(t.Context(), schema.Contexto{}, peticion)
	require.NoError(t, err)

	t.Setenv(VariableGrabacion, variableActiva)

	grabando, err := clienteQueGraba(t, t.TempDir()).Pedir(t.Context(), schema.Contexto{}, peticion)
	require.NoError(t, err)

	assert.Equal(t, sinGrabar.Cuerpo, grabando.Cuerpo, "el cuerpo se entrega íntegro (FR-038)")
	assert.Equal(t, contenidoDePrueba, string(grabando.Cuerpo))
	assert.Equal(t, sinGrabar.Estado, grabando.Estado)
	assert.Equal(t, sinGrabar.URL, grabando.URL)
	assert.Equal(t, sinGrabar.Cabeceras.Get("Content-Type"), grabando.Cabeceras.Get("Content-Type"))
}

// TestGrabarFormatoEstable es el escenario 3 de US4 y FR-040: dos grabaciones de
// la misma petición y la misma respuesta son idénticas byte a byte una vez
// neutralizado lo que declara cuándo se grabó —la fecha que el fichero anota y
// la que el servidor regenera—, y nada más. Las dos van contra el mismo
// servidor, porque el puerto que httptest elige entra en la dirección y en el
// nombre del fichero.
func TestGrabarFormatoEstable(t *testing.T) {
	t.Setenv(VariableGrabacion, variableActiva)

	servidor, _ := servidorIdentificado(t, redireccionesDePrueba())
	peticion := Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/norma"}
	nombre := nombreDelServidor(t, servidor, "GET", "_norma")

	primera, segunda := t.TempDir(), t.TempDir()

	for _, raiz := range []string{primera, segunda} {
		_, err := clienteQueGraba(t, raiz).Pedir(t.Context(), schema.Contexto{}, peticion)
		require.NoError(t, err)
	}

	assert.Equal(t, neutralizado(t, rutaGrabada(primera, nombre)), neutralizado(t, rutaGrabada(segunda, nombre)),
		"ni el orden de las claves ni el de las cabeceras pueden variar entre dos ejecuciones (FR-040)")
}

// TestGrabarSinVariableNoEscribe es la otra mitad de SC-008: sin la variable no
// se escribe nada, ni siquiera con la fuente y la raíz declaradas. Es lo que
// hace que make ci no deje una sola grabación en el árbol de trabajo (FR-041).
func TestGrabarSinVariableNoEscribe(t *testing.T) {
	t.Setenv(VariableGrabacion, "")

	casos := []struct {
		nombre  string
		ausente bool
	}{
		{nombre: "la variable vacía", ausente: false},
		{nombre: "la variable ausente", ausente: true},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			// t.Setenv es lo que devuelve el entorno a su estado al terminar,
			// también cuando lo que la subprueba necesita es que la variable no
			// esté: se declara vacía y se retira, y la restauración sigue siendo
			// la suya.
			t.Setenv(VariableGrabacion, "")

			if caso.ausente {
				require.NoError(t, os.Unsetenv(VariableGrabacion))
			}

			raiz := t.TempDir()
			servidor, _ := servidorIdentificado(t, redireccionesDePrueba())

			respuesta, err := clienteQueGraba(t, raiz).Pedir(t.Context(), schema.Contexto{},
				Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/norma"})
			require.NoError(t, err, "sin grabar, el cliente es el de siempre")
			assert.Equal(t, contenidoDePrueba, string(respuesta.Cuerpo))

			assert.Empty(t, entradasDe(t, raiz),
				"con la variable apagada no se crea ni el directorio de la fuente")
		})
	}
}

// TestGrabarConfiguracionIncompleta es el escenario 6 de US4: con la variable
// puesta, un cliente al que le falta la fuente o la raíz no se construye, el
// fallo es de la clase «argumentos» y nombra la opción que falta, y no se
// escribe nada en ninguna parte. Que falle en vez de deducir la raíz del
// directorio de trabajo o de la del módulo es justamente FR-064.
func TestGrabarConfiguracionIncompleta(t *testing.T) {
	t.Setenv(VariableGrabacion, variableActiva)

	raiz := t.TempDir()

	casos := []struct {
		nombre   string
		opciones []Opcion
		mencion  string
	}{
		{
			nombre:   "sin la fuente",
			opciones: []Opcion{ConRaizDeGrabacion(raiz)},
			mencion:  "ConFuente",
		},
		{
			nombre:   "sin la raíz",
			opciones: []Opcion{ConFuente(fuenteDePrueba)},
			mencion:  "ConRaizDeGrabacion",
		},
		{
			nombre:   "sin ninguna de las dos",
			opciones: nil,
			mencion:  "ConFuente",
		},
		{
			nombre:   "con la raíz vacía, que no es declararla",
			opciones: []Opcion{ConFuente(fuenteDePrueba), ConRaizDeGrabacion("")},
			mencion:  "ConRaizDeGrabacion",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Setenv(VariableGrabacion, variableActiva)

			cliente, err := New(caso.opciones...)

			assert.Nil(t, cliente, "la grabación mal configurada no construye ningún cliente")

			fallo := falloDe(t, err)
			assert.Equal(t, schema.ClaseArgumentos, fallo.Clase(),
				"lo que falta es una opción, que es algo que quien llama puede corregir (FR-039, FR-064)")
			assert.Contains(t, fallo.Error(), caso.mencion, "el mensaje nombra la opción que falta")
			assert.Empty(t, entradasDe(t, raiz), "y no se escribe nada en ninguna parte")
		})
	}
}

// TestGrabarRaizInvalida es FR-042 por su mitad temprana: una raíz que no existe,
// que no es un directorio o bajo la que no se puede crear el directorio de la
// fuente es un valor de opción inválido —clase «argumentos», comprobado en la
// construcción y nombrando la ruta—, y no un fallo de la fuente ni del mecanismo.
func TestGrabarRaizInvalida(t *testing.T) {
	t.Setenv(VariableGrabacion, variableActiva)

	inexistente := filepath.Join(t.TempDir(), "ausente")

	noEsDirectorio := filepath.Join(t.TempDir(), "raiz.txt")
	require.NoError(t, os.WriteFile(noEsDirectorio, []byte("no soy un directorio"), 0o600))

	conLaFuenteOcupada := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(conLaFuenteOcupada, fuenteDePrueba),
		[]byte("tampoco"), 0o600))

	casos := []struct {
		nombre string
		raiz   string
	}{
		{nombre: "la raíz no existe", raiz: inexistente},
		{nombre: "la raíz no es un directorio", raiz: noEsDirectorio},
		{nombre: "el directorio de la fuente no se puede crear", raiz: conLaFuenteOcupada},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Setenv(VariableGrabacion, variableActiva)

			cliente, err := New(ConFuente(fuenteDePrueba), ConRaizDeGrabacion(caso.raiz))

			assert.Nil(t, cliente, "un cliente que no podría grabar no se construye")

			fallo := falloDe(t, err)
			assert.Equal(t, schema.ClaseArgumentos, fallo.Clase(),
				"una raíz inválida es un valor de opción que quien llama puede corregir (FR-042)")
			assert.Contains(t, fallo.Error(), caso.raiz, "el mensaje nombra la ruta")
		})
	}
}

// TestGrabarFalloSobrevenido es FR-042 por su otra mitad, la tardía: lo que
// falla después de que la construcción validara y creara el directorio no es un
// valor de opción sino un tropiezo del mecanismo, y sale con la clase
// «inesperado» nombrando el fichero que no se pudo escribir, sin entregar
// respuesta y sin dejar nada escrito. El tropiezo es el directorio que
// desaparece entre una petición y la siguiente: es el que se puede provocar en
// cualquier máquina sin depender de permisos, que a quien ejecute la suite como
// administrador no se le negarían.
func TestGrabarFalloSobrevenido(t *testing.T) {
	t.Setenv(VariableGrabacion, variableActiva)

	raiz := t.TempDir()
	servidor, _ := servidorIdentificado(t, redireccionesDePrueba())
	cliente := clienteQueGraba(t, raiz)

	// La primera petición graba bien y deja decidido el robots.txt del sitio,
	// de modo que lo que falle después sea la grabación del recurso y no la
	// obtención del permiso.
	primera, err := cliente.Pedir(t.Context(), schema.Contexto{},
		Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/norma"})
	require.NoError(t, err, "la grabación de la primera petición sí se puede escribir")
	require.Equal(t, http.StatusOK, primera.Estado)

	require.NoError(t, os.RemoveAll(filepath.Join(raiz, fuenteDePrueba)),
		"el directorio que la construcción validó y creó desaparece")

	direccion := servidor.URL + "/norma?otra=1"

	respuesta, err := cliente.Pedir(t.Context(), schema.Contexto{}, Peticion{Metodo: http.MethodGet, URL: direccion})

	assert.Equal(t, Respuesta{}, respuesta, "una grabación que no se puede escribir no entrega ninguna respuesta")

	fallo := falloDe(t, err)
	assert.Equal(t, schema.ClaseInesperado, fallo.Clase(),
		"un tropiezo sobrevenido del mecanismo es «inesperado», ni de la fuente ni de los argumentos (FR-042)")
	assert.Contains(t, fallo.Error(), rutaGrabada(raiz, nombreDelServidor(t, servidor, "GET", "_norma_q_otra_1")),
		"el mensaje nombra el fichero que no se pudo escribir")
	assert.Equal(t, Peticion{Metodo: http.MethodGet, URL: direccion}, fallo.Peticion,
		"y la petición implicada es la que se iba a grabar")
	assert.Empty(t, entradasDe(t, raiz), "no queda nada escrito: ni grabación ni temporal (contrato §1)")
}

// TestGrabarRetiraElTemporal es la promesa de escritura del contrato §1 —«nunca
// queda una grabación a medias»— por la mitad que ninguna tabla contra un
// servidor puede provocar: la escritura que falla **después** de crear el
// temporal. Se prueba sobre la propia escritura en firme, con un destino que un
// rename de fichero no puede sustituir —un directorio—, y lo que se exige es que
// el temporal desaparezca y que el directorio quede como estaba. El otro fallo
// posible tras crearlo, el de volcar el contenido, retira el temporal por el
// mismo camino.
func TestGrabarRetiraElTemporal(t *testing.T) {
	t.Parallel()

	directorio := t.TempDir()
	destino := filepath.Join(directorio, "GET_http_fuente.prueba_norma.json")
	require.NoError(t, os.Mkdir(destino, permisoDelDirectorioDeGrabacion),
		"el destino lo ocupa un directorio, que el rename del temporal no puede sustituir")

	err := escribirEnFirme(destino, &grabacion{Formato: formatoDeGrabacion, GrabadoEn: "2026-09-12T00:00:00Z"})
	require.Error(t, err, "la escritura en firme falla al poner el temporal en su sitio")

	entradas := entradasDe(t, directorio)
	require.Len(t, entradas, 1, "en el directorio queda solo lo que ya estaba: ni temporal ni grabación")
	assert.Equal(t, filepath.Base(destino), entradas[0].Name())
	assert.True(t, entradas[0].IsDir(), "el destino sigue siendo el directorio que ya estaba")

	parciales, err := filepath.Glob(filepath.Join(directorio, patronDelTemporal))
	require.NoError(t, err)
	assert.Empty(t, parciales, "el temporal de una escritura fallida se retira (contrato §1)")
}

// TestGrabarColision es FR-039 por su mitad de escritura: el nombre no es único
// por construcción, así que antes de escribir se compara la petición que el
// fichero guarda. Otra petición es un fallo del mecanismo —clase «inesperado»—
// que nombra las dos y el fichero y no toca nada; la misma petición sustituye.
func TestGrabarColision(t *testing.T) {
	t.Setenv(VariableGrabacion, variableActiva)

	t.Run("otra petición no se pisa en silencio", func(t *testing.T) {
		t.Setenv(VariableGrabacion, variableActiva)

		raiz := t.TempDir()
		servidor, _ := servidorIdentificado(t, servidorQueNumera())

		// «/a,b» y «/a_b» se sanean al mismo nombre: es la colisión que la tabla
		// del contrato §2 declara, y la única que se puede provocar a mano.
		ocupada := servidor.URL + "/a,b"
		ruta := rutaGrabada(raiz, nombreDelServidor(t, servidor, "GET", "_a_b"))
		require.NoError(t, os.MkdirAll(filepath.Dir(ruta), 0o750))

		yaEscrita := grabacionAMano(ocupada)
		require.NoError(t, os.WriteFile(ruta, []byte(yaEscrita), 0o600))

		respuesta, err := clienteQueGraba(t, raiz).Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/a_b"})

		assert.Equal(t, Respuesta{}, respuesta, "una colisión no entrega ninguna respuesta")

		fallo := falloDe(t, err)
		assert.Equal(t, schema.ClaseInesperado, fallo.Clase(),
			"que dos peticiones compartan fichero es un fallo del mecanismo, no de la fuente (FR-063)")
		assert.Contains(t, fallo.Error(), ocupada, "el mensaje nombra la petición ya grabada")
		assert.Contains(t, fallo.Error(), servidor.URL+"/a_b", "y la que se iba a grabar")
		assert.Contains(t, fallo.Error(), ruta, "y el fichero de las dos")

		assert.Equal(t, yaEscrita, string(contenidoDe(t, ruta)), "la grabación ajena se queda como estaba")
	})

	t.Run("la misma petición sustituye la grabación anterior", func(t *testing.T) {
		t.Setenv(VariableGrabacion, variableActiva)

		raiz := t.TempDir()
		servidor, _ := servidorIdentificado(t, servidorQueNumera())
		peticion := Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/norma"}

		primera, err := clienteQueGraba(t, raiz).Pedir(t.Context(), schema.Contexto{}, peticion)
		require.NoError(t, err)

		segunda, err := clienteQueGraba(t, raiz).Pedir(t.Context(), schema.Contexto{}, peticion)
		require.NoError(t, err, "regrabar la misma petición no es una colisión")
		require.NotEqual(t, primera.Cuerpo, segunda.Cuerpo, "el servidor responde algo distinto cada vez")

		grabada := grabacionLeida(t, rutaGrabada(raiz, nombreDelServidor(t, servidor, "GET", "_norma")))
		require.NotNil(t, grabada.Respuesta.Cuerpo)
		assert.Equal(t, string(segunda.Cuerpo), *grabada.Respuesta.Cuerpo,
			"lo que queda es la última grabación")
	})
}

// TestGrabarValorDeVariableInvalido es la última fila de la tabla de
// construcción: la variable solo admite «1». Cualquier otro valor es un error de
// argumentos que nombra el valor, y no una grabación silenciosa ni un silencio
// que deje de grabar (D12).
func TestGrabarValorDeVariableInvalido(t *testing.T) {
	t.Setenv(VariableGrabacion, variableActiva)

	raiz := t.TempDir()

	for _, valor := range []string{"0", "true", "si", "yes", " 1"} {
		t.Run("la variable vale "+strconv.Quote(valor), func(t *testing.T) {
			t.Setenv(VariableGrabacion, valor)

			cliente, err := New(ConFuente(fuenteDePrueba), ConRaizDeGrabacion(raiz))

			assert.Nil(t, cliente)

			fallo := falloDe(t, err)
			assert.Equal(t, schema.ClaseArgumentos, fallo.Clase())
			assert.Contains(t, fallo.Error(), VariableGrabacion, "el mensaje nombra la variable")
			assert.Contains(t, fallo.Error(), valor, "y el valor que trae")
			assert.Empty(t, entradasDe(t, raiz), "un valor inválido no escribe nada")
		})
	}
}

// TestGrabarCuerpoBinario es la última regla del contrato §1: un cuerpo que no
// es UTF-8 válido —un PDF— va en «cuerpo_base64» y nunca en «cuerpo», y quien
// llamó sigue recibiendo los bytes exactos.
func TestGrabarCuerpoBinario(t *testing.T) {
	t.Setenv(VariableGrabacion, variableActiva)

	binario := []byte{0x25, 0x50, 0x44, 0x46, 0x2d, 0xff, 0xfe, 0x00, 0x0a}

	raiz := t.TempDir()
	servidor, _ := servidorIdentificado(t, func(escritor http.ResponseWriter, _ *http.Request) {
		escritor.Header().Set("Content-Type", "application/pdf")
		_, _ = escritor.Write(binario)
	})

	respuesta, err := clienteQueGraba(t, raiz).Pedir(t.Context(), schema.Contexto{},
		Peticion{Metodo: http.MethodGet, URL: servidor.URL + "/documento.pdf"})
	require.NoError(t, err)
	assert.Equal(t, binario, respuesta.Cuerpo, "quien llama recibe los bytes tal cual (FR-038)")

	ruta := rutaGrabada(raiz, nombreDelServidor(t, servidor, "GET", "_documento.pdf"))
	grabada := grabacionLeida(t, ruta)

	require.NotNil(t, grabada.Respuesta.CuerpoBase64, "lo que no es UTF-8 válido va en base64")
	assert.Equal(t, base64.StdEncoding.EncodeToString(binario), *grabada.Respuesta.CuerpoBase64)
	assert.Nil(t, grabada.Respuesta.Cuerpo, "«cuerpo» y «cuerpo_base64» son excluyentes")
	assert.NotContains(t, string(contenidoDe(t, ruta)), `"cuerpo":`,
		"la clave del texto ni siquiera aparece")
}

// TestGrabacionDeFormularios es el control de FR-034 y la mitad de FR-112 que
// toca a la grabación (contrato httpx-formulario §5 de H23): dos envíos a la
// misma dirección con campos distintos son dos grabaciones. Lo que las separa
// es el cuerpo, que entra en el nombre del fichero, se guarda dentro y se
// compara —exacto, con el método y la dirección— al grabar y al reproducir.
//
// Todo ocurre en un temporal. Las grabaciones sin cuerpo que el paquete ya
// versiona solo se leen: lo que se vuelve a grabar de ellas va a otro
// directorio.
func TestGrabacionDeFormularios(t *testing.T) {
	t.Setenv(VariableGrabacion, variableActiva)

	t.Run("dos envíos dejan dos ficheros y cada uno reproduce el suyo", func(t *testing.T) {
		t.Setenv(VariableGrabacion, variableActiva)

		probarDosEnviosDosGrabaciones(t)
	})

	t.Run("otro cuerpo con el mismo nombre ni se graba encima ni se reproduce", func(t *testing.T) {
		t.Setenv(VariableGrabacion, variableActiva)

		probarEnviosQueCompartenNombre(t)
	})

	t.Run("una grabación sin cuerpo se reproduce y se vuelve a grabar igual", func(t *testing.T) {
		// Con la variable apagada: aquí no graba ningún cliente, sino el escalón
		// de grabación montado a mano, que no la lee.
		t.Setenv(VariableGrabacion, "")

		probarGrabacionSinCuerpo(t)
	})
}

// probarDosEnviosDosGrabaciones graba dos envíos contra un sitio de prueba y los
// reproduce después sin él. El sitio responde a cada uno con el cuerpo que
// recibió, de modo que la respuesta dice dos cosas: que el cuerpo llegó entero
// —leerlo para grabarlo no consumió el que se enviaba— y de cuál de los dos
// envíos es cada grabación.
func probarDosEnviosDosGrabaciones(t *testing.T) {
	t.Helper()

	raiz := t.TempDir()
	servidor, contador := servidorIdentificado(t, formularioQueRepite())
	formulario := servidor.URL + rutaDelFormulario
	envios := enviosDePrueba()

	consulta := consultaDePrueba(t, clienteQueGraba(t, raiz, ConFormulario(formulario)))

	for _, envio := range envios {
		respuesta, err := consulta.Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodPost, URL: formulario, Campos: envio.campos})
		require.NoError(t, err, "grabar no cambia el resultado del envío")
		assert.Equal(t, loRecibido+envio.cuerpo, string(respuesta.Cuerpo),
			"el sitio recibe el cuerpo entero: leerlo para grabarlo no consume el que se envía")
	}

	directorio := filepath.Join(raiz, fuenteDePrueba)

	require.ElementsMatch(t, []string{
		nombreDelServidor(t, servidor, "GET", "_robots.txt"),
		nombreDelServidor(t, servidor, "POST", envios[0].resto),
		nombreDelServidor(t, servidor, "POST", envios[1].resto),
	}, nombresDe(t, directorio),
		"dos envíos a la misma dirección con campos distintos son dos grabaciones (FR-034)")

	esperado := `{
  "formato": 1,
  "grabado_en": "` + fechaNeutralizada + `",
  "peticion": {
    "metodo": "POST",
    "url": "` + formulario + `",
    "cuerpo": "` + envios[0].cuerpo + `",
    "cabeceras": {
      "Content-Type": [
        "application/x-www-form-urlencoded"
      ],
      "User-Agent": [
        "` + AgenteDeUsuario() + `"
      ],
      "X-Requested-With": [
        "XMLHttpRequest"
      ]
    }
  },
  "respuesta": {
    "estado": 200,
    "cabeceras": {
      "Content-Length": [
        "` + strconv.Itoa(len(loRecibido+envios[0].cuerpo)) + `"
      ],
      "Content-Type": [
        "text/plain; charset=utf-8"
      ],
      "Date": [
        "` + fechaNeutralizada + `"
      ]
    },
    "cuerpo": "` + loRecibido + envios[0].cuerpo + `"
  }
}
`

	assert.Equal(t, esperado, neutralizado(t, rutaGrabada(raiz, nombreDelServidor(t, servidor, "POST", envios[0].resto))),
		"la petición gana «cuerpo» entre «url» y «cabeceras», y «formato» sigue en 1 (contrato §5)")

	// Grabar y reproducir a la vez es un error de argumentos: lo grabado se
	// reproduce con la variable apagada, que es como lo hace quien lo usa.
	t.Setenv(VariableGrabacion, "")

	emitidas := contador.total.Load()
	reproduccion := consultaDePrueba(t, clienteDeReproduccion(t, directorio, ConFormulario(formulario)))

	// En el orden contrario al de la grabación: lo que empareja un envío con su
	// fichero es su cuerpo, no el turno en que se pidió.
	for _, envio := range slices.Backward(envios) {
		respuesta, err := reproduccion.Pedir(t.Context(), schema.Contexto{},
			Peticion{Metodo: http.MethodPost, URL: formulario, Campos: envio.campos})
		require.NoError(t, err, "un envío grabado se reproduce")
		assert.Equal(t, http.StatusOK, respuesta.Estado)
		assert.Equal(t, loRecibido+envio.cuerpo, string(respuesta.Cuerpo),
			"cada envío reproduce su grabación, y no la del otro (FR-034)")
	}

	respuesta, err := reproduccion.Pedir(t.Context(), schema.Contexto{}, Peticion{
		Metodo: http.MethodPost, URL: formulario,
		Campos: map[string]string{"action": "query", "ECLI": "ECLI:ES:TS:2023:1"},
	})

	fallo := falloDe(t, err)
	assert.Equal(t, schema.ClaseInesperado, fallo.Clase(),
		"un envío cuyos campos no están grabados falla como una petición sin grabación (contrato §5)")
	assert.Equal(t, Respuesta{}, respuesta, "y no reproduce la de ningún otro envío")
	assert.Contains(t, fallo.Error(),
		filepath.Join(directorio, nombreDelServidor(t, servidor, "POST", "_buscar_c_ECLI_ECLI_3AES_3ATS_3A2023_3A1_action_query")),
		"el mensaje nombra el fichero que falta, que es el de su cuerpo")

	assert.Equal(t, emitidas, contador.total.Load(), "reproducir no pide nada al sitio")
}

// probarEnviosQueCompartenNombre es la colisión que el nombre no puede evitar:
// dos cuerpos distintos que se sanean a los mismos caracteres —la coma
// codificada de uno, «%2C», y el «_2C» literal del otro—. La resuelve el
// contenido, igual que entre dos direcciones: el segundo no se graba encima del
// primero ni se reproduce con su respuesta, y el fallo nombra los dos cuerpos y
// el fichero. Es lo que deja de pasar si el emparejamiento no mira el cuerpo.
func probarEnviosQueCompartenNombre(t *testing.T) {
	t.Helper()

	const (
		cuerpoGrabado  = "q=a%2Cb"
		cuerpoParecido = "q=a_2Cb"
	)

	raiz := t.TempDir()
	servidor, _ := servidorIdentificado(t, formularioQueRepite())
	formulario := servidor.URL + rutaDelFormulario

	grabado := Peticion{Metodo: http.MethodPost, URL: formulario, Campos: map[string]string{"q": "a,b"}}
	parecido := Peticion{Metodo: http.MethodPost, URL: formulario, Campos: map[string]string{"q": "a_2Cb"}}

	consulta := consultaDePrueba(t, clienteQueGraba(t, raiz, ConFormulario(formulario)))

	_, err := consulta.Pedir(t.Context(), schema.Contexto{}, grabado)
	require.NoError(t, err, "el primer envío se graba")

	ruta := rutaGrabada(raiz, nombreDelServidor(t, servidor, "POST", "_buscar_c_q_a_2Cb"))
	yaEscrita := string(contenidoDe(t, ruta))

	respuesta, err := consulta.Pedir(t.Context(), schema.Contexto{}, parecido)

	assert.Equal(t, Respuesta{}, respuesta, "una colisión no entrega ninguna respuesta")

	fallo := falloDe(t, err)
	assert.Equal(t, schema.ClaseInesperado, fallo.Clase(),
		"que dos envíos compartan fichero es un fallo del mecanismo, no de la fuente")
	assert.Contains(t, fallo.Error(), cuerpoGrabado, "el mensaje nombra el cuerpo ya grabado")
	assert.Contains(t, fallo.Error(), cuerpoParecido, "y el que se iba a grabar")
	assert.Contains(t, fallo.Error(), ruta, "y el fichero de los dos")
	assert.Equal(t, yaEscrita, string(contenidoDe(t, ruta)), "la grabación del otro envío se queda como estaba")

	t.Setenv(VariableGrabacion, "")

	reproduccion := consultaDePrueba(t, clienteDeReproduccion(t, filepath.Dir(ruta), ConFormulario(formulario)))

	propia, err := reproduccion.Pedir(t.Context(), schema.Contexto{}, grabado)
	require.NoError(t, err, "el envío grabado se reproduce")
	assert.Equal(t, loRecibido+cuerpoGrabado, string(propia.Cuerpo))

	ajena, err := reproduccion.Pedir(t.Context(), schema.Contexto{}, parecido)

	assert.Equal(t, Respuesta{}, ajena, "la grabación de otro envío no se sirve nunca")

	fallo = falloDe(t, err)
	assert.Equal(t, schema.ClaseInesperado, fallo.Clase())
	assert.Contains(t, fallo.Error(), cuerpoParecido, "el mensaje nombra el cuerpo que se buscaba")
	assert.Contains(t, fallo.Error(), cuerpoGrabado, "y el que encontró en su lugar")
	assert.Contains(t, fallo.Error(), ruta, "y el fichero de los dos")
}

// probarGrabacionSinCuerpo es la otra mitad de FR-034: una petición sin campos
// se nombra, se graba y se reproduce byte a byte como antes de que hubiera
// cuerpos. Cada grabación que el paquete versiona se reproduce con la petición
// que guarda y se vuelve a grabar, en un temporal, con el mismo nombre y el
// mismo contenido; lo único que cambia es la fecha en que se graba.
//
// La grabación se monta aquí sobre la reproducción, y no sobre un servidor,
// porque es la única forma de volver a grabar sin red una dirección que no es
// de esta máquina. La petición lleva las cabeceras con las que se grabó: no
// entran en el emparejamiento, pero sí en el fichero.
func probarGrabacionSinCuerpo(t *testing.T) {
	t.Helper()

	origen := grabacionesDePrueba(t)
	destino := t.TempDir()
	cadena := conGrabacion(nuevoTransporteDeReproduccion(origen), destino)

	versionadas := nombresDe(t, origen)
	require.NotEmpty(t, versionadas, "el paquete versiona grabaciones sin cuerpo")

	for _, nombre := range versionadas {
		original := contenidoDe(t, filepath.Join(origen, nombre))
		grabada := grabacionLeida(t, filepath.Join(origen, nombre))
		require.Nil(t, grabada.Peticion.Cuerpo, "%s es de una petición sin campos", nombre)

		peticion, err := nuevaPeticionIdentificada(t.Context(), grabada.Peticion.Metodo, grabada.Peticion.URL)
		require.NoError(t, err)

		peticion.Header = grabada.Peticion.Cabeceras

		respuesta, err := cadena.RoundTrip(peticion)
		require.NoError(t, err, "%s se reproduce como hasta ahora", nombre)
		require.NoError(t, respuesta.Body.Close())
		assert.Equal(t, grabada.Respuesta.Estado, respuesta.StatusCode)

		assert.Equal(t, sinFechaDeGrabacion(original), sinFechaDeGrabacion(contenidoDe(t, filepath.Join(destino, nombre))),
			"%s se vuelve a grabar con su nombre y byte a byte, sin la clave «cuerpo» en la petición", nombre)
	}

	assert.Equal(t, versionadas, nombresDe(t, destino), "una grabación por cada una de las versionadas, y ninguna más")
}

// TestCuerpoQueNoSePuedeLeer cubre lo que ningún envío del paquete puede
// provocar, porque el único cuerpo que sale de él lo pone ponerFormulario: una
// petición cuyo cuerpo no declara cómo leerlo otra vez, o cuya copia no se deja
// abrir o leer. Ni se graba ni se empareja como si no enviara nada, que sería
// hacerlo mal en silencio: es un fallo del mecanismo —clase «inesperado»—, igual
// al grabar que al reproducir, y no deja ningún fichero.
//
// Los dos escalones se prueban sueltos, y el de grabación sobre uno de mentira,
// porque lo que se estropea es la petición que les llega y eso no lo puede hacer
// nadie desde fuera del paquete.
func TestCuerpoQueNoSePuedeLeer(t *testing.T) {
	t.Parallel()

	errDeLaCopia := errors.New("la copia del cuerpo no está")

	casos := []struct {
		nombre   string
		estropea func(*http.Request)
		causa    error
	}{
		{
			nombre:   "el cuerpo no declara cómo leerlo otra vez",
			estropea: func(peticion *http.Request) { peticion.GetBody = nil },
		},
		{
			nombre: "la copia del cuerpo no se deja abrir",
			estropea: func(peticion *http.Request) {
				peticion.GetBody = func() (io.ReadCloser, error) { return nil, errDeLaCopia }
			},
			causa: errDeLaCopia,
		},
		{
			nombre: "la copia del cuerpo no se deja leer",
			estropea: func(peticion *http.Request) {
				peticion.GetBody = func() (io.ReadCloser, error) {
					return io.NopCloser(iotest.ErrReader(errDeLaCopia)), nil
				}
			},
			causa: errDeLaCopia,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			directorio := t.TempDir()

			escalones := []struct {
				paso    string
				escalon http.RoundTripper
			}{
				{"al grabar", conGrabacion(&escalonQueFalla{estado: http.StatusOK, aciertaEn: 1}, directorio)},
				{"al reproducir", nuevoTransporteDeReproduccion(directorio)},
			}

			for _, fila := range escalones {
				peticion, err := nuevaPeticionIdentificada(t.Context(), http.MethodPost, direccionDelEscalon)
				require.NoError(t, err)

				ponerFormulario(peticion, camposDePrueba())
				caso.estropea(peticion)

				fallo := falloDelEscalon(t, fila.escalon, peticion)

				assert.Equal(t, schema.ClaseInesperado, fallo.Clase(),
					"%s, un cuerpo que no se puede leer es un fallo del mecanismo", fila.paso)
				assert.Contains(t, fallo.Error(), "el cuerpo de la petición", "%s, el mensaje dice qué no se pudo leer", fila.paso)
				assert.Equal(t, Peticion{Metodo: http.MethodPost, URL: direccionDelEscalon}, fallo.Peticion,
					"%s, la petición implicada es la que bajaba por la cadena", fila.paso)

				if caso.causa != nil {
					require.ErrorIs(t, fallo, caso.causa, "%s, el fallo conserva su causa", fila.paso)
				}
			}

			assert.Empty(t, entradasDe(t, directorio), "y no queda ninguna grabación escrita")
		})
	}
}

// falloDelEscalon baja la petición por un escalón suelto que tiene que
// rechazarla, y devuelve su fallo. Un escalón que falla no entrega respuesta;
// si la entregara, se cierra aquí antes de decirlo, como manda la biblioteca.
func falloDelEscalon(t *testing.T, escalon http.RoundTripper, peticion *http.Request) *Error {
	t.Helper()

	respuesta, err := escalon.RoundTrip(peticion)
	if respuesta != nil {
		require.NoError(t, respuesta.Body.Close())
	}

	require.Nil(t, respuesta, "un escalón que falla no entrega ninguna respuesta")

	return falloDe(t, err)
}

// envioDePrueba es un envío de TestGrabacionDeFormularios: sus campos, el cuerpo
// con el que tienen que salir y lo que ese cuerpo añade al nombre del fichero,
// los dos escritos a mano y no derivados con el código que se prueba.
type envioDePrueba struct {
	campos map[string]string
	cuerpo string
	resto  string
}

// enviosDePrueba son dos envíos a la misma dirección que solo se distinguen por
// el valor de un campo. Sus cuerpos son cortos a propósito: el nombre del
// fichero no llega al tope con ningún puerto del servidor de prueba, y se puede
// escribir entero.
func enviosDePrueba() []envioDePrueba {
	return []envioDePrueba{
		{
			campos: map[string]string{"action": "query", "ECLI": "ECLI:ES:TS:2023:3144"},
			cuerpo: "ECLI=ECLI%3AES%3ATS%3A2023%3A3144&action=query",
			resto:  "_buscar_c_ECLI_ECLI_3AES_3ATS_3A2023_3A3144_action_query",
		},
		{
			campos: map[string]string{"action": "query", "ECLI": "ECLI:ES:TS:2023:999999"},
			cuerpo: "ECLI=ECLI%3AES%3ATS%3A2023%3A999999&action=query",
			resto:  "_buscar_c_ECLI_ECLI_3AES_3ATS_3A2023_3A999999_action_query",
		},
	}
}

// loRecibido es lo que el formulario de prueba antepone al cuerpo que repite.
const loRecibido = "recibido: "

// formularioQueRepite responde al formulario con el cuerpo que recibió, tal
// cual, y a nada más.
func formularioQueRepite() http.HandlerFunc {
	return func(escritor http.ResponseWriter, peticion *http.Request) {
		if peticion.URL.Path != rutaDelFormulario {
			http.NotFound(escritor, peticion)

			return
		}

		cuerpo, err := io.ReadAll(peticion.Body)
		if err != nil {
			http.Error(escritor, err.Error(), http.StatusBadRequest)

			return
		}

		escritor.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = io.WriteString(escritor, loRecibido+string(cuerpo))
	}
}

// clienteQueGraba construye el cliente que graba bajo la raíz que se le indica,
// con el ritmo acelerado que usan todas las tablas del paquete: el intervalo por
// omisión separaría un segundo la petición del robots.txt de la del recurso y
// haría de cada tabla una espera.
func clienteQueGraba(t *testing.T, raiz string, opciones ...Opcion) *Cliente {
	t.Helper()

	return clienteDePrueba(t, append([]Opcion{
		ConFuente(fuenteDePrueba), ConRaizDeGrabacion(raiz), ConIntervalo(time.Millisecond),
	}, opciones...)...)
}

// rutaGrabada es donde queda la grabación de una petición bajo una raíz: el
// árbol <raíz>/<fuente>/<nombre>.json que FR-036 declara.
func rutaGrabada(raiz, nombre string) string {
	return filepath.Join(raiz, fuenteDePrueba, nombre)
}

// nombreDelServidor escribe a mano el nombre de fichero que le corresponde a una
// ruta del servidor local, con la forma del contrato §2 y sin derivarlo con el
// código que se prueba. Lo único que no se escribe a mano es el puerto, que
// httptest elige en cada ejecución.
func nombreDelServidor(t *testing.T, servidor *httptest.Server, metodo, resto string) string {
	t.Helper()

	direccion, err := url.Parse(servidor.URL)
	require.NoError(t, err)
	require.Equal(t, "127.0.0.1", direccion.Hostname(),
		"las tablas del hito solo conocen direcciones locales")

	return metodo + "_http_127.0.0.1-" + direccion.Port() + resto + ".json"
}

// grabacionEnDisco es la forma del fichero tal como la declara el contrato §1,
// con sus claves escritas aquí literalmente: leerla con los tipos del código que
// se prueba haría que un cambio de nombre de clave pasara desapercibido.
type grabacionEnDisco struct {
	Formato   int    `json:"formato"`
	GrabadoEn string `json:"grabado_en"`
	Peticion  struct {
		Metodo    string              `json:"metodo"`
		URL       string              `json:"url"`
		Cuerpo    *string             `json:"cuerpo"`
		Cabeceras map[string][]string `json:"cabeceras"`
	} `json:"peticion"`
	Respuesta struct {
		Estado       int                 `json:"estado"`
		Cabeceras    map[string][]string `json:"cabeceras"`
		Cuerpo       *string             `json:"cuerpo"`
		CuerpoBase64 *string             `json:"cuerpo_base64"`
	} `json:"respuesta"`
}

// grabacionLeida interpreta el fichero que dejó la grabación.
func grabacionLeida(t *testing.T, ruta string) grabacionEnDisco {
	t.Helper()

	var leida grabacionEnDisco
	require.NoError(t, json.Unmarshal(contenidoDe(t, ruta), &leida))

	return leida
}

// contenidoDe lee un fichero que el propio test ha escrito o ha hecho escribir.
// La ruta pasa por filepath.Clean porque es lo que el control de rutas reconoce
// como saneado antes de abrir un fichero (gosec G304) y este proyecto no admite
// ninguna supresión.
func contenidoDe(t *testing.T, ruta string) []byte {
	t.Helper()

	limpia := filepath.Clean(ruta)

	contenido, err := os.ReadFile(limpia)
	require.NoError(t, err, "la grabación tiene que existir en %s", ruta)

	return contenido
}

// entradasDe es lo que hay dentro de un directorio, que es como estas tablas
// comprueban que no se escribió nada.
func entradasDe(t *testing.T, directorio string) []os.DirEntry {
	t.Helper()

	entradas, err := os.ReadDir(directorio)
	require.NoError(t, err)

	return entradas
}

// nombresDe son los nombres de lo que hay dentro de un directorio, en el orden
// en que lo da el sistema, que es el alfabético (go doc os.ReadDir).
func nombresDe(t *testing.T, directorio string) []string {
	t.Helper()

	entradas := entradasDe(t, directorio)
	nombres := make([]string, 0, len(entradas))

	for _, entrada := range entradas {
		nombres = append(nombres, entrada.Name())
	}

	return nombres
}

// fechaNeutralizada es lo que ocupa el sitio de los dos campos que declaran
// cuándo se grabó, y que FR-040 deja fuera de la comparación.
const fechaNeutralizada = "«cuando sea»"

// Los dos únicos campos que pueden variar entre dos grabaciones de la misma
// petición y la misma respuesta: la fecha que el fichero anota y la que el
// servidor regenera en cada ejecución. Ningún otro se neutraliza, que es lo que
// hace que la comparación siga siendo la de FR-040 y no una más laxa.
var (
	fechaDelFichero  = regexp.MustCompile(`"grabado_en": "[^"]*"`)
	fechaDelServidor = regexp.MustCompile(`("Date": \[\n\s*)"[^"]*"`)
)

// neutralizado devuelve el fichero con esos dos campos sustituidos, dejando
// intacto todo lo demás —el orden de las claves, la indentación y las comillas
// incluidas—, que es lo que se compara byte a byte.
func neutralizado(t *testing.T, ruta string) string {
	t.Helper()

	sinFecha := sinFechaDeGrabacion(contenidoDe(t, ruta))

	return fechaDelServidor.ReplaceAllString(sinFecha, `${1}"`+fechaNeutralizada+`"`)
}

// sinFechaDeGrabacion neutraliza solo la fecha que el fichero anota. Es todo lo
// que cambia al volver a grabar desde una grabación, donde hasta la fecha del
// servidor es la que ya estaba guardada.
func sinFechaDeGrabacion(contenido []byte) string {
	return string(fechaDelFichero.ReplaceAll(contenido, []byte(`"grabado_en": "`+fechaNeutralizada+`"`)))
}

// grabacionAMano es el fichero que un intento anterior habría dejado para otra
// petición, escrito con la forma del contrato §1 para que la comprobación de
// colisión lo interprete igual que interpretaría uno suyo.
func grabacionAMano(direccion string) string {
	return `{
  "formato": 1,
  "grabado_en": "2026-09-12T00:00:00Z",
  "peticion": {
    "metodo": "GET",
    "url": "` + direccion + `",
    "cabeceras": {
      "User-Agent": [
        "` + AgenteDeUsuario() + `"
      ]
    }
  },
  "respuesta": {
    "estado": 200,
    "cabeceras": {
      "Content-Type": [
        "text/plain; charset=utf-8"
      ]
    },
    "cuerpo": "la que ya estaba"
  }
}
`
}

// servidorQueNumera responde a toda ruta con un cuerpo distinto en cada
// petición, que es lo que deja ver si una regrabación sustituyó la anterior o la
// dejó como estaba.
func servidorQueNumera() http.HandlerFunc {
	var atendidas atomic.Int64

	return func(escritor http.ResponseWriter, _ *http.Request) {
		escritor.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = io.WriteString(escritor, "respuesta "+strconv.FormatInt(atendidas.Add(1), 10))
	}
}
