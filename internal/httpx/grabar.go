package httpx

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"unicode/utf8"
)

// VariableGrabacion es la variable de entorno que activa la grabación, y la
// única forma de activarla: no hay bandera de la línea de órdenes ni opción del
// constructor que la encienda, porque grabar es algo que se hace a propósito al
// preparar un fixture y nunca durante una consulta normal (FR-041, D12).
const VariableGrabacion = "KITLEGAL_RECORD"

// valorQueActivaLaGrabacion es el único valor que la enciende. Ausente o vacía
// la deja apagada; cualquier otro valor es un error de argumentos, porque una
// variable mal escrita no puede grabar ni dejar de grabar en silencio (D12).
const valorQueActivaLaGrabacion = "1"

// formatoDeGrabacion es la versión del formato del fichero, que todo fichero
// declara y la reproducción comprueba: un fichero de otro formato se rechaza
// entero, en vez de leerse a medias (contrato de grabación §1).
const formatoDeGrabacion = 1

// permisoDelDirectorioDeGrabacion es el del árbol <raíz>/<fuente> que la
// construcción crea. La umask del proceso lo restringe todavía más; lo que este
// valor fija es que no sea legible para todo el mundo.
const permisoDelDirectorioDeGrabacion = 0o750

// patronDelTemporal nombra el fichero a medio escribir. Va en el mismo
// directorio que el de destino —os.Rename no cruza sistemas de ficheros— y con
// extensión distinta de la de una grabación, para que uno que sobreviva a un
// corte no lo tome nadie por buena.
const patronDelTemporal = "grabacion-*.json.parcial"

// indentacionDeLaGrabacion son los dos espacios por nivel del contrato §1: lo
// bastante para que un diff de una revisión humana se lea, sin que el fichero
// crezca por la sangría.
const indentacionDeLaGrabacion = "  "

// grabacion es el fichero entero, y el orden de sus campos es el orden en que se
// escriben: «the encoding of each struct field [becomes] a member object …
// [in] the order they are declared» (go doc encoding/json.Marshal). Que el orden
// sea el de la declaración, y no el de un mapa, es lo que hace estable el
// formato sin ninguna regla adicional (FR-040, contrato §1).
type grabacion struct {
	Formato   int              `json:"formato"`
	GrabadoEn string           `json:"grabado_en"`
	Peticion  peticionGrabada  `json:"peticion"`
	Respuesta respuestaGrabada `json:"respuesta"`
}

// peticionGrabada es la petición tal como se emitió, y es además la clave de
// emparejamiento: el método, la dirección completa y el cuerpo son lo que la
// grabación compara para detectar una colisión y lo que la reproducción compara
// para aceptar un fichero (FR-039, contrato §3 y §4; FR-034 de H23).
//
// El cuerpo es el de un envío de formulario, ya codificado, y solo lo lleva la
// petición que envía campos: en las demás va vacío y omitempty deja el fichero
// sin la clave, byte a byte como antes de que hubiera cuerpos. No hace falta
// distinguir «no está» de «está y va vacío», porque un envío lleva al menos un
// campo y su cuerpo nunca es la cadena vacía (contrato httpx-formulario §2 y §5
// de H23).
type peticionGrabada struct {
	Metodo    string    `json:"metodo"`
	URL       string    `json:"url"`
	Cuerpo    string    `json:"cuerpo,omitempty"`
	Cabeceras Cabeceras `json:"cabeceras"`
}

// esLaMisma es la regla de emparejamiento, la misma al grabar y al reproducir:
// el método, la dirección y el cuerpo tienen que coincidir **exactamente**. Las
// cabeceras no participan, porque la identificación lleva la versión del binario
// y la haría cambiar de una versión a otra (FR-039, FR-047, contrato §4), y la
// cookie de una consulta es otra en cada sesión.
func (p peticionGrabada) esLaMisma(otra peticionGrabada) bool {
	return p.Metodo == otra.Metodo && p.URL == otra.URL && p.Cuerpo == otra.Cuerpo
}

// descripcion nombra la petición en un mensaje con lo que la empareja: el método
// y la dirección y, si envía un formulario, su cuerpo, que es lo único que
// distingue a dos envíos a la misma dirección.
func (p peticionGrabada) descripcion() string {
	if p.Cuerpo == sinCuerpo {
		return p.Metodo + " " + p.URL
	}

	return p.Metodo + " " + p.URL + " con el cuerpo " + p.Cuerpo
}

// respuestaGrabada es lo que la fuente respondió. El cuerpo va en una de las dos
// claves y nunca en las dos: como texto si es UTF-8 válido —para que el diff de
// una revisión humana se lea— y como base64 si no —un PDF—. Son punteros porque
// lo que distingue «no está» de «está y va vacío» es justamente eso: un cuerpo
// vacío se graba como "cuerpo": "" y omitempty no lo omite, porque el puntero no
// es nulo (contrato §1).
type respuestaGrabada struct {
	Estado       int       `json:"estado"`
	Cabeceras    Cabeceras `json:"cabeceras"`
	Cuerpo       *string   `json:"cuerpo,omitempty"`
	CuerpoBase64 *string   `json:"cuerpo_base64,omitempty"`
}

// decoradorDeGrabacion es el escalón que va justo encima de la marca de emisión
// y, por tanto, del transporte: graba lo que la fuente respondió y entrega la
// respuesta intacta a quien la pidió (FR-036, FR-038, D3; contrato
// httpx-acepta-e-instante §4 de H4).
//
// Va ahí y no más arriba porque lo que se graba tiene que ser la petición tal
// como salió —ya identificada, ya con su turno esperado y ya en el intento que
// de verdad se emitió— y la respuesta tal como llegó, sin que ningún escalón de
// encima la haya interpretado todavía. La marca, que solo anota la hora, no
// cambia nada de lo que se graba.
type decoradorDeGrabacion struct {
	// directorio es <raíz>/<fuente>, ya validado y creado en la construcción:
	// aquí no se deduce ninguna ruta ni se comprueba ninguna, porque lo que
	// llega a este punto es una escritura que ya se decidió que se podía hacer
	// (FR-042, FR-064).
	directorio string
	// siguiente es la marca de emisión y, bajo ella, el transporte, el escalón
	// que de verdad abre la conexión.
	siguiente http.RoundTripper
}

// conGrabacion envuelve el escalón que recibe con el decorador. Solo se compone
// cuando la grabación está activa: sin la variable, la cadena no lleva este
// escalón y no hay ruta alguna por la que se escriba un fichero (FR-041).
func conGrabacion(siguiente http.RoundTripper, directorio string) http.RoundTripper {
	return &decoradorDeGrabacion{directorio: directorio, siguiente: siguiente}
}

// RoundTrip emite la petición, graba lo que la fuente respondió y devuelve la
// respuesta con el cuerpo otra vez entero y sin leer. Leerlo es inevitable
// —grabarlo exige tenerlo— y por eso se repone con uno equivalente: quien llama
// lo recibe íntegro y una sola vez, exactamente como si no se estuviera grabando
// (FR-038, D12).
//
// El fallo al leer el cuerpo no se declara aquí: es la fuente la que no ha
// sabido entregarlo, así que sube sin clase y lo clasifica quien clasifica la
// cadena, igual que cualquier otro tropiezo del transporte. El fallo al grabar sí
// se declara, y con la clase del mecanismo (FR-042).
func (g *decoradorDeGrabacion) RoundTrip(peticion *http.Request) (*http.Response, error) {
	respuesta, err := g.siguiente.RoundTrip(peticion)
	if err != nil {
		return nil, err
	}

	cuerpo, err := leerYCerrar(respuesta.Body)
	if err != nil {
		return nil, err
	}

	if err := g.escribir(peticion, respuesta, cuerpo); err != nil {
		return nil, err
	}

	respuesta.Body = io.NopCloser(bytes.NewReader(cuerpo))

	return respuesta, nil
}

// leerYCerrar deja el cuerpo cerrado y devuelve lo que traía. El cierre se hace
// pase lo que pase con la lectura —«the caller should close resp.Body when done
// reading from it» (go doc net/http.RoundTripper)—, y el error que gana es el de
// la lectura, que es el que dice qué falló; el del cierre solo cuenta cuando la
// lectura fue bien. Vale igual para el cuerpo de una respuesta que para la copia
// del de una petición.
func leerYCerrar(cuerpo io.ReadCloser) ([]byte, error) {
	contenido, errDeLectura := io.ReadAll(cuerpo)

	errAlCerrar := cuerpo.Close()

	if errDeLectura != nil {
		return nil, errDeLectura
	}

	if errAlCerrar != nil {
		return nil, errAlCerrar
	}

	return contenido, nil
}

// escribir deja la grabación en <directorio>/<nombre>.json, con el nombre que la
// petición misma determina (FR-039). Antes comprueba que el fichero que pueda
// haber ahí sea de esta misma petición, porque el nombre no es único por
// construcción; después escribe en firme.
func (g *decoradorDeGrabacion) escribir(peticion *http.Request, respuesta *http.Response, cuerpo []byte) error {
	implicada := Peticion{Metodo: peticion.Method, URL: peticion.URL.String()}

	emitida, err := peticionEmitida(peticion, implicada)
	if err != nil {
		return err
	}

	ruta := filepath.Join(g.directorio, nombreDeGrabacion(peticion.Method, peticion.URL, emitida.Cuerpo))

	if err := comprobarQueEsLaMisma(ruta, implicada, emitida); err != nil {
		return err
	}

	contenido := &grabacion{
		Formato:   formatoDeGrabacion,
		GrabadoEn: time.Now().UTC().Format(time.RFC3339),
		Peticion:  emitida,
		Respuesta: respuestaGrabadaDe(respuesta, cuerpo),
	}

	if err := escribirEnFirme(ruta, contenido); err != nil {
		return errorInesperado(implicada, err, "la grabación no se ha podido escribir en "+ruta)
	}

	return nil
}

// peticionEmitida traduce la petición que baja por la cadena al objeto del
// contrato §1, que es también lo que se empareja: su método, su dirección, el
// cuerpo con el que sale y sus cabeceras. La usan la grabación, para escribirla,
// y la reproducción, para buscarla, de modo que las dos leen el cuerpo de la
// misma manera (research D6 de H23).
func peticionEmitida(peticion *http.Request, implicada Peticion) (peticionGrabada, error) {
	cuerpo, err := cuerpoEnviado(peticion, implicada)
	if err != nil {
		return peticionGrabada{}, err
	}

	return peticionGrabada{
		Metodo:    implicada.Metodo,
		URL:       implicada.URL,
		Cuerpo:    cuerpo,
		Cabeceras: Cabeceras(peticion.Header),
	}, nil
}

// cuerpoEnviado devuelve el cuerpo con el que sale la petición —los campos de un
// formulario, ya codificados— o sinCuerpo si no envía nada. No toca el que se
// envía: lee una copia, la que da GetBody, que es como la biblioteca deja leer un
// cuerpo otra vez (go doc net/http.Request.GetBody), y el transporte recibe el
// suyo entero y sin empezar.
//
// El único cuerpo que sale del paquete lo pone ponerFormulario, que declara
// GetBody. Un cuerpo sin GetBody no se podría leer sin consumirlo, y grabarlo o
// emparejarlo como si la petición no enviara nada sería hacerlo mal en silencio:
// es un fallo del mecanismo, como el de la copia que no se deja abrir o leer.
func cuerpoEnviado(peticion *http.Request, implicada Peticion) (string, error) {
	if peticion.GetBody == nil {
		if peticion.Body == nil || peticion.Body == http.NoBody {
			return sinCuerpo, nil
		}

		return sinCuerpo, errorInesperado(implicada, nil,
			"el cuerpo de la petición no se puede leer sin consumir el que se envía: no declara GetBody")
	}

	copia, err := peticion.GetBody()
	if err != nil {
		return sinCuerpo, errorInesperado(implicada, err, "el cuerpo de la petición no se ha podido leer")
	}

	contenido, err := leerYCerrar(copia)
	if err != nil {
		return sinCuerpo, errorInesperado(implicada, err, "el cuerpo de la petición no se ha podido leer")
	}

	return string(contenido), nil
}

// respuestaGrabadaDe traduce la respuesta al objeto del contrato §1, con el
// cuerpo en la clave que le toca: texto si es UTF-8 válido y base64 si no.
// «Estándar» es la codificación con relleno de RFC 4648 §4, que es la que
// base64.StdEncoding produce y la que cualquier herramienta descodifica sin
// preguntar.
func respuestaGrabadaDe(respuesta *http.Response, cuerpo []byte) respuestaGrabada {
	grabada := respuestaGrabada{
		Estado:    respuesta.StatusCode,
		Cabeceras: Cabeceras(respuesta.Header),
	}

	if utf8.Valid(cuerpo) {
		comoTexto := string(cuerpo)
		grabada.Cuerpo = &comoTexto

		return grabada
	}

	comoBase64 := base64.StdEncoding.EncodeToString(cuerpo)
	grabada.CuerpoBase64 = &comoBase64

	return grabada
}

// comprobarQueEsLaMisma es la detección de colisión por contenido de FR-039. El
// nombre se deriva de la petición, pero dos peticiones distintas pueden
// sanearse al mismo —«/a,b» y «/a_b», «/A» y «/a» en un sistema de ficheros
// insensible a mayúsculas, o dos envíos cuyos cuerpos solo difieren en lo que el
// saneado iguala—, así que lo que decide es lo que el fichero guarda dentro: la
// misma petición se regraba y otra distinta es un fallo del mecanismo que nombra
// las dos y el fichero, sin escribir nada (contrato §3).
//
// Un fichero que está pero no se deja leer o no se deja interpretar tampoco
// autoriza a escribir encima: no se sabe qué guarda, y sobrescribirlo sería
// perder una grabación ajena en silencio.
func comprobarQueEsLaMisma(ruta string, implicada Peticion, emitida peticionGrabada) error {
	// filepath.Clean es lo que el control de rutas reconoce como saneado antes
	// de abrir un fichero (gosec G304); la ruta la compone este paquete a partir
	// del directorio validado en la construcción y de un nombre que solo lleva
	// caracteres de [A-Za-z0-9._-], pero el control no lo distingue y aquí no se
	// suprime ninguna comprobación.
	limpia := filepath.Clean(ruta)

	contenido, err := os.ReadFile(limpia)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return errorInesperado(implicada, err, "la grabación que ya había en "+ruta+" no se ha podido leer")
	}

	var existente grabacion
	if err := json.Unmarshal(contenido, &existente); err != nil {
		return errorInesperado(implicada, err,
			"la grabación que ya había en "+ruta+" no se ha podido interpretar")
	}

	if existente.Peticion.esLaMisma(emitida) {
		return nil
	}

	return errorInesperado(implicada, nil,
		"la grabación "+ruta+" guarda "+existente.Peticion.descripcion()+
			" y aquí se iba a grabar "+emitida.descripcion())
}

// escribirEnFirme escribe la grabación en un temporal del mismo directorio y la
// pone en su sitio con un rename, que en un mismo sistema de ficheros sustituye
// el destino de una vez. Así un corte a media escritura no deja nunca una
// grabación truncada que la reproducción tomaría por buena (D12).
//
// El temporal se retira en cuanto algo falla después de crearlo —volcar el
// contenido o ponerlo en su sitio—, y por el mismo camino en los dos casos: lo
// que no llegó a ser una grabación no se queda en el directorio
// (TestGrabarRetiraElTemporal).
func escribirEnFirme(ruta string, contenido *grabacion) error {
	// El destino pasa por filepath.Clean por lo mismo que la lectura de
	// comprobarQueEsLaMisma: es lo que el control de rutas reconoce como saneo
	// antes de tocar un fichero, y aquí no se suprime ninguna comprobación.
	destino := filepath.Clean(ruta)

	temporal, err := os.CreateTemp(filepath.Dir(destino), patronDelTemporal)
	if err != nil {
		return err
	}

	if err := volcarYCerrar(temporal, contenido); err != nil {
		retirar(temporal.Name())

		return err
	}

	if err := os.Rename(filepath.Clean(temporal.Name()), destino); err != nil {
		retirar(temporal.Name())

		return err
	}

	return nil
}

// volcarYCerrar serializa la grabación en el temporal y lo cierra. El cierre se
// hace pase lo que pase con la escritura, y el error que gana es el de la
// escritura, que es el que dice qué falló; el del cierre solo cuenta cuando la
// escritura fue bien.
func volcarYCerrar(temporal *os.File, contenido *grabacion) error {
	errAlEscribir := serializar(temporal, contenido)

	errAlCerrar := temporal.Close()

	if errAlEscribir != nil {
		return errAlEscribir
	}

	return errAlCerrar
}

// serializar escribe el objeto del contrato §1 con la forma que fija FR-040: dos
// espacios de sangría y **sin** escape de HTML, porque sin SetEscapeHTML(false)
// «<», «>» y «&» saldrían como <, > y & (go doc
// encoding/json.Encoder.SetEscapeHTML) y el cuerpo XML o HTML de una fuente no
// habría quien lo leyera en un diff. Las cabeceras son mapas, y sus claves las
// ordena la biblioteca —«the map keys are sorted» (go doc encoding/json.Marshal)—,
// de modo que su orden también es estable.
func serializar(destino io.Writer, contenido *grabacion) error {
	codificador := json.NewEncoder(destino)
	codificador.SetIndent("", indentacionDeLaGrabacion)
	codificador.SetEscapeHTML(false)

	return codificador.Encode(contenido)
}

// retirar borra el temporal que quedó a medias, ya cerrado. El resultado no se
// comprueba, y no es un error silenciado: lo que se va a devolver es el fallo
// que trajo hasta aquí, que es el que explica qué pasó, y sustituirlo por el de
// no haber podido borrar un fichero que ya no se va a usar cambiaría un motivo
// cierto por uno accesorio.
func retirar(nombre string) {
	aBorrar := filepath.Clean(nombre)

	_ = os.Remove(aBorrar)
}

// directorioDeGrabacion decide si este cliente graba y dónde, con la tabla de
// reglas de construcción del contrato §3. Devuelve la cadena vacía cuando la
// grabación está apagada, que es lo que deja la cadena sin el escalón.
//
// Todo lo que puede estar mal se comprueba aquí, lo antes posible (FR-042): a
// partir de este punto, escribir una grabación solo puede fallar por un tropiezo
// sobrevenido de entrada y salida, que es de otra clase.
func directorioDeGrabacion(valor string, declarada bool, config configuracionDelCliente) (string, error) {
	if !declarada || valor == "" {
		return "", nil
	}

	if valor != valorQueActivaLaGrabacion {
		return "", errorDeArgumentos(Peticion{}, nil,
			"la variable de entorno "+VariableGrabacion+" solo admite el valor "+
				valorQueActivaLaGrabacion+": "+valor)
	}

	if config.fuente == "" {
		return "", errorDeArgumentos(Peticion{}, nil,
			"la grabación exige declarar la fuente (ConFuente)")
	}

	if config.raizDeGrabacion == "" {
		return "", errorDeArgumentos(Peticion{}, nil,
			"la grabación exige declarar la raíz de grabación (ConRaizDeGrabacion)")
	}

	return crearDirectorioDeGrabacion(config.raizDeGrabacion, config.fuente)
}

// crearDirectorioDeGrabacion comprueba la raíz que se declaró y crea bajo ella
// el directorio de la fuente. La raíz tiene que existir y ser un directorio: se
// declara para escribir dentro de algo que ya está, no para inventar un árbol
// entero a partir de una ruta mal escrita. El de la fuente sí se crea, porque es
// el que la propia grabación organiza (contrato §3).
func crearDirectorioDeGrabacion(raiz, fuente string) (string, error) {
	deLaRaiz, err := os.Stat(raiz)
	if err != nil {
		return "", errorDeArgumentos(Peticion{}, err,
			"la raíz de grabación no existe o no se puede consultar: "+raiz)
	}

	if !deLaRaiz.IsDir() {
		return "", errorDeArgumentos(Peticion{}, nil,
			"la raíz de grabación no es un directorio: "+raiz)
	}

	directorio := filepath.Join(raiz, fuente)
	if err := os.MkdirAll(directorio, permisoDelDirectorioDeGrabacion); err != nil {
		return "", errorDeArgumentos(Peticion{}, err,
			"el directorio de grabación no se puede crear: "+directorio)
	}

	return directorio, nil
}
