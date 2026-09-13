package boe

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"
	"unicode/utf8"
)

// La consulta guardada (data-model.md §6, research.md D5): lo que la fuente deja
// en la caché por cada consulta que respondió y que se pudo interpretar. La caché
// guarda la clave y el contenido como opacos (internal/cache); su forma es la de
// este fichero:
//
//	clave:     boe.legislacion-consolidada|1|<verbo>|<dirección de la API del recurso>
//	contenido: {"fecha_consulta": "<RFC 3339>", "url": "<dirección>", "datos": <data>}
//
// La dirección de la API ya es la identidad exacta de lo pedido —lleva todo
// argumento que cambia la respuesta—, así que dos consultas distintas nunca
// comparten entrada y dos textos de búsqueda que construyen la misma consulta sí
// (FR-090). La fecha y la url son las del sobre de la consulta que escribe la
// entrada, y servirla da las dos igual, byte a byte (FR-096).
const (
	// versionDeLasEntradas es la versión de la forma del contenido, el 1 de la
	// clave. Si cambia la forma de data o la del contenido, sube, y las entradas
	// escritas con la anterior dejan de encontrarse: ninguna se lee nunca con
	// otra forma que la suya.
	versionDeLasEntradas = "1"
	// separadorDeLaClave separa los cuatro segmentos de la clave. No lo lleva el
	// nombre de la fuente, ni la versión, ni ningún verbo, ni ninguna dirección:
	// las de la norma y el bloque solo admiten letras, dígitos, guiones y puntos
	// (ValidarNorma y ValidarBloque) y la de búsqueda lo escribe como %7C
	// (quoteComoPython). Cada clave se parte en sus segmentos de una sola forma.
	separadorDeLaClave = "|"
	// motivoDeLaEntradaIlegible es el mensaje de la entrada con la clave correcta
	// que no tiene la forma de su contenido (contrato errores-y-codigos §2,
	// fila 20).
	motivoDeLaEntradaIlegible = "la entrada de la caché con la clave %q no se puede leer"
	// motivoDeLaEntradaSinComponer es el mensaje de la entrada que no se puede
	// componer porque lo guardado no volvería igual.
	motivoDeLaEntradaSinComponer = "la entrada de la caché con la clave %q no se puede componer"
)

// Las causas de lo que no se puede leer ni componer. Son detalle técnico: van al
// registro con la causa del *Error, no a su mensaje.
var (
	errContenidoNoUTF8     = errors.New("el contenido no es UTF-8 válido")
	errContenidoDetras     = errors.New("hay contenido detrás del objeto de la entrada")
	errFechaDeConsultaCero = errors.New("la fecha de consulta es el instante cero")
	errFaltaLaFecha        = errors.New("falta fecha_consulta o es nula")
	errFaltaLaURL          = errors.New("falta url o es nula")
	errFaltanLosDatos      = errors.New("faltan datos o son nulos")
)

// datosDeEntrada son los data que se guardan, uno por clase de entrada
// (data-model.md §6): los resultados de buscar, el índice, los metadatos, el
// artículo y el análisis. articulos no tiene entrada propia: lee y escribe la de
// articulo de cada bloque (FR-020).
type datosDeEntrada interface {
	[]ResultadoDeBusqueda | Indice | Metadatos | Articulo | Analisis
}

// claveDeEntrada es la clave de una entrada, con el tipo de sus datos: solo la
// construyen las funciones de este fichero, una por clase de entrada, así que
// ningún verbo puede guardar ni leer bajo una clave los datos de otra.
type claveDeEntrada[T datosDeEntrada] struct {
	// verbo es el que da nombre a la entrada, que en la de los metadatos es
	// metadatos también cuando la leen o la escriben articulo y articulos.
	verbo string
	// direccion es la dirección de la API del recurso, la que se pide, la que
	// lleva el sobre y la que nombra el fallo de la entrada.
	direccion string
}

// claveDeLaBusqueda es la de buscar, con la dirección que construye
// direccionDeBusqueda, que lleva la consulta y el límite.
func claveDeLaBusqueda(direccion string) claveDeEntrada[[]ResultadoDeBusqueda] {
	return claveDeEntrada[[]ResultadoDeBusqueda]{verbo: verboBuscar, direccion: direccion}
}

// claveDelIndice es la de indice de una norma ya validada.
func claveDelIndice(norma string) claveDeEntrada[Indice] {
	return claveDeEntrada[Indice]{verbo: verboIndice, direccion: direccionDelIndice(norma)}
}

// claveDeLosMetadatos es la de los metadatos de una norma ya validada. Solo
// depende de la norma: metadatos, articulo y articulos leen y escriben la misma
// entrada (FR-090).
func claveDeLosMetadatos(norma string) claveDeEntrada[Metadatos] {
	return claveDeEntrada[Metadatos]{verbo: verboMetadatos, direccion: direccionDeLosMetadatos(norma)}
}

// claveDelArticulo es la de un bloque de una norma, los dos ya validados: la
// que usa articulo y la de cada bloque de articulos (FR-020).
func claveDelArticulo(norma, bloque string) claveDeEntrada[Articulo] {
	return claveDeEntrada[Articulo]{verbo: verboArticulo, direccion: direccionDelBloque(norma, bloque)}
}

// claveDelAnalisis es la de analisis de una norma ya validada.
func claveDelAnalisis(norma string) claveDeEntrada[Analisis] {
	return claveDeEntrada[Analisis]{verbo: verboAnalisis, direccion: direccionDelAnalisis(norma)}
}

// String es la clave tal como se entrega a la caché.
func (c claveDeEntrada[T]) String() string {
	return NombreDeLaFuente + separadorDeLaClave + versionDeLasEntradas + separadorDeLaClave +
		c.verbo + separadorDeLaClave + c.direccion
}

// ilegible es el fallo de la entrada con esta clave que no se puede leer:
// «inesperado», código 1, con la dirección del recurso y sin instante, porque no
// hubo petición (contrato errores-y-codigos, fila 20). Con la versión en la
// clave solo puede ser una corrupción o un defecto, y no se tapa: tratarla como
// ausencia la silenciaría durante toda su vigencia (research.md D5).
func (c claveDeEntrada[T]) ilegible(causa error) *Error {
	return errorInesperado(c.direccion, time.Time{}, causa, fmt.Sprintf(motivoDeLaEntradaIlegible, c.String()))
}

// sinComponer es el fallo de la entrada con esta clave que no se puede
// componer. Es un defecto de quien la escribe, no de quien invoca: «inesperado»,
// como la entrada ilegible, con la dirección del recurso y sin instante.
func (c claveDeEntrada[T]) sinComponer(causa error) *Error {
	return errorInesperado(c.direccion, time.Time{}, causa, fmt.Sprintf(motivoDeLaEntradaSinComponer, c.String()))
}

// entrada es el contenido de una consulta guardada: la fecha_consulta y la url
// del sobre de la consulta que la escribe, y su data.
type entrada[T datosDeEntrada] struct {
	// FechaConsulta es el instante de la consulta que sostiene los datos
	// (FR-096), en RFC 3339 con nanosegundos y desplazamiento: la forma en que
	// schema.Sobre escribe fecha_consulta, de modo que la del sobre servido es,
	// byte a byte, la del sobre que la guardó.
	FechaConsulta time.Time `json:"fecha_consulta"`
	// URL es la del sobre, la dirección de la clave.
	URL string `json:"url"`
	// Datos es el data del sobre.
	Datos T `json:"datos"`
}

// entradaLeida es entrada con cada campo por puntero, para que la clave que falta
// o llega nula se distinga de la que trae el valor cero.
type entradaLeida[T datosDeEntrada] struct {
	FechaConsulta *time.Time `json:"fecha_consulta"`
	URL           *string    `json:"url"`
	Datos         *T         `json:"datos"`
}

// contenidoDeEntrada compone el contenido que se guarda bajo la clave: la fecha
// de consulta, la dirección de la clave como url y los datos. Lo que no volvería
// igual al leerlo no se compone: una fecha cero, que el sobre servido no podría
// declarar (schema.Procedencia), o una que RFC 3339 no puede escribir.
func contenidoDeEntrada[T datosDeEntrada](clave claveDeEntrada[T], fechaConsulta time.Time, datos T) ([]byte, error) {
	if fechaConsulta.IsZero() {
		return nil, clave.sinComponer(errFechaDeConsultaCero)
	}

	contenido, err := json.Marshal(entrada[T]{FechaConsulta: fechaConsulta, URL: clave.direccion, Datos: datos})
	if err != nil {
		return nil, clave.sinComponer(err)
	}

	return contenido, nil
}

// leerEntrada lee el contenido guardado bajo la clave. Solo lo da por bueno si
// tiene la forma que escribe contenidoDeEntrada:
//
//   - UTF-8 válido y un único objeto JSON, con solo espacio en blanco detrás;
//   - sin ninguna clave desconocida, ni en él ni dentro de los datos
//     (Decoder.DisallowUnknownFields);
//   - con fecha_consulta, url y datos, ninguna nula;
//   - la fecha en RFC 3339 con desplazamiento y distinta del instante cero;
//   - y la url igual a la dirección de la clave.
//
// Cualquier otra cosa es el fallo ilegible de la clave, con el detalle técnico
// como causa.
func leerEntrada[T datosDeEntrada](clave claveDeEntrada[T], contenido []byte) (entrada[T], error) {
	if !utf8.Valid(contenido) {
		return entrada[T]{}, clave.ilegible(errContenidoNoUTF8)
	}

	decodificador := json.NewDecoder(bytes.NewReader(contenido))
	decodificador.DisallowUnknownFields()

	var leida entradaLeida[T]
	if err := decodificador.Decode(&leida); err != nil {
		return entrada[T]{}, clave.ilegible(err)
	}

	// Decode se detiene al final del primer valor: lo que haya detrás, salvo
	// espacio en blanco, no es de la entrada.
	if _, err := decodificador.Token(); !errors.Is(err, io.EOF) {
		return entrada[T]{}, clave.ilegible(errors.Join(errContenidoDetras, err))
	}

	switch {
	case leida.FechaConsulta == nil:
		return entrada[T]{}, clave.ilegible(errFaltaLaFecha)
	case leida.URL == nil:
		return entrada[T]{}, clave.ilegible(errFaltaLaURL)
	case leida.Datos == nil:
		return entrada[T]{}, clave.ilegible(errFaltanLosDatos)
	case leida.FechaConsulta.IsZero():
		return entrada[T]{}, clave.ilegible(errFechaDeConsultaCero)
	case *leida.URL != clave.direccion:
		return entrada[T]{}, clave.ilegible(fmt.Errorf("la url %q no es la dirección de la clave", *leida.URL))
	}

	return entrada[T]{FechaConsulta: *leida.FechaConsulta, URL: *leida.URL, Datos: *leida.Datos}, nil
}
