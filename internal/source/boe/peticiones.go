package boe

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/httpx"
)

// Pedidor es lo único que la fuente necesita del cliente HTTP: pedir, con el
// contexto de cancelación, cuyo plazo es el de --timeout (FR-095), y con el
// contexto de ejecución del kernel, que es el que lleva --dry-run hasta quien
// tiene el efecto (contrato puerto-y-applet §3.1, ADR 0011). Lo declara la
// fuente, que es quien lo consume, para que sus tests lo sustituyan sin red; en
// producción es *httpx.Cliente, la única puerta a la red (FR-003).
type Pedidor interface {
	Pedir(ctx context.Context, ec schema.Contexto, p httpx.Peticion) (httpx.Respuesta, error)
}

// El cliente de httpx es el Pedidor de producción, y que lo siga siendo no
// depende de que alguien lo recuerde.
var _ Pedidor = (*httpx.Cliente)(nil)

// Cómo pide la fuente: siempre con GET y en el formato que pide refs/boe.py para
// cada recurso, XML para el texto de un bloque y JSON para todo lo demás, en la
// cabecera Accept (líneas 77 y 405-407; FR-003, research.md D4). Van escritos y
// no con net/http, que solo importa internal/httpx (R2).
const (
	metodoDeLasPeticiones = "GET"
	formatoXML            = "application/xml"
	formatoJSON           = "application/json"
)

// Los estados que la fuente distingue en lo que httpx le entrega, que es todo
// salvo el 3xx, el 429 y el 5xx: esos ya llegan como error con su clase (H2
// FR-032; research.md D10).
const (
	// primerEstadoCorrecto y primerEstadoTrasLosCorrectos acotan los 2xx, los
	// únicos cuyo cuerpo se interpreta.
	primerEstadoCorrecto         = 200
	primerEstadoTrasLosCorrectos = 300
	// estadoNoEncontrado es el 404, el único con el que la fuente dice que lo
	// pedido no existe (S2).
	estadoNoEncontrado = 404
)

// Los motivos de un fallo al pedir, que nombran lo pedido detrás de «la
// petición»; la dirección la añade el propio *Error (contrato errores-y-codigos
// §2, filas 9 a 15).
const (
	motivoDelEstado       = "la fuente ha respondido con el estado %d a la petición %s"
	motivoDelFalloAlPedir = "ha fallado la petición %s"
)

// pedido es un recurso de la API tal como la fuente lo pide: su dirección, el
// formato en que lo quiere y lo que hace falta para que un fallo al pedirlo diga
// qué se pidió sin leer el registro (contrato errores-y-codigos §2). Solo lo
// construyen las funciones de este fichero, una por recurso, con la norma y el
// bloque ya validados.
type pedido struct {
	// direccion es la dirección de la API del recurso: la que se pide y la que
	// lleva cualquier fallo al pedirlo, nunca la del robots.txt (FR-101).
	direccion string
	// formato es el tipo de contenido que se pide en la cabecera Accept.
	formato string
	// deQue nombra lo pedido detrás de «la petición»: «del bloque a21 de la
	// norma BOE-A-2015-10565».
	deQue string
	// inexistente es el motivo con el que el 404 de un recurso que puede no
	// existir es «no encontrado» —«la norma BOE-A-2015-10565 no tiene el bloque
	// a9999»— y va vacío en la búsqueda, cuyo 404 no dice que nada no exista:
	// una búsqueda nunca es «no encontrado» (FR-032; contrato errores-y-codigos,
	// filas 6, 7 y 9).
	inexistente string
}

// pedidoDelBloque es el texto de un bloque de la norma, en XML (contrato
// verbos-y-salidas §3).
func pedidoDelBloque(norma, bloque string) pedido {
	return pedido{
		direccion:   direccionDelBloque(norma, bloque),
		formato:     formatoXML,
		deQue:       "del bloque " + bloque + " de la norma " + norma,
		inexistente: "la norma " + norma + " no tiene el bloque " + bloque,
	}
}

// pedidoDelIndice es el índice de bloques de la norma, en JSON (contrato
// verbos-y-salidas §2).
func pedidoDelIndice(norma string) pedido {
	return pedidoDeLaNorma(direccionDelIndice(norma), norma, "del índice", "índice")
}

// pedidoDeLosMetadatos son los metadatos de la norma, en JSON (contrato
// verbos-y-salidas §3 y §5).
func pedidoDeLosMetadatos(norma string) pedido {
	return pedidoDeLaNorma(direccionDeLosMetadatos(norma), norma, "de los metadatos", "metadatos")
}

// pedidoDelAnalisis es el análisis de la norma, en JSON (contrato
// verbos-y-salidas §6).
func pedidoDelAnalisis(norma string) pedido {
	return pedidoDeLaNorma(direccionDelAnalisis(norma), norma, "del análisis", "análisis")
}

// pedidoDeLaNorma es lo que comparten el índice, los metadatos y el análisis: se
// piden en JSON y su 404 es «no encontrado», que nombra la norma y el recurso
// (contrato errores-y-codigos §2, fila 7).
func pedidoDeLaNorma(direccion, norma, deQue, recurso string) pedido {
	return pedido{
		direccion:   direccion,
		formato:     formatoJSON,
		deQue:       deQue + " de la norma " + norma,
		inexistente: "la norma " + norma + " no tiene " + recurso,
	}
}

// pedidoDeLaBusqueda es la búsqueda, en JSON, con la dirección que construye
// direccionDeBusqueda (contrato verbos-y-salidas §1).
func pedidoDeLaBusqueda(direccion string) pedido {
	return pedido{
		direccion: direccion,
		formato:   formatoJSON,
		deQue:     "de la búsqueda",
	}
}

// obtenido es lo que pedir entrega cuando la petición no falla.
type obtenido struct {
	// cuerpo es el de la respuesta 2xx, que interpreta quien pidió.
	cuerpo []byte
	// instante es el de emisión que declara la respuesta, Respuesta.Instante: la
	// fecha de consulta de lo que se lea del cuerpo (FR-096).
	instante time.Time
	// ensayo es, bajo --dry-run, la línea «GET <dirección>» de la petición que no
	// se emitió, la que quien pidió copia en schema.Resultado.Ensayo (ADR 0011),
	// y va vacía cuando se emitió. En ensayo no hay cuerpo ni instante.
	ensayo string
}

// pedir pide el recurso al Pedidor y clasifica lo que entrega (contrato
// errores-y-codigos, filas 6, 7 y 9 a 15; research.md D10):
//
//   - en ensayo, la descripción de la petición que el Pedidor no emitió, sin
//     mirar su estado, que no tiene (ADR 0011);
//   - un 2xx, su cuerpo y su instante;
//   - el 404 de lo que puede no existir, «no encontrado» (filas 6 y 7), y
//     cualquier otro estado —el 404 de una búsqueda también—, «fuente no
//     disponible» con el estado en el mensaje (filas 9 y 10): httpx entrega a
//     quien llama todo 4xx salvo el 429 porque solo la fuente sabe qué
//     significa, y ni el 403 ni ningún otro declaran un límite o los términos de
//     uso;
//   - y el error del Pedidor, con la clase, la dirección y el instante de
//     falloAlPedir (filas 11 a 15).
//
// Un fallo de estado lleva el instante de emisión de la respuesta que lo trae,
// porque el sobre de fallo se fecha como el de éxito (FR-096). pedir no
// interpreta el cuerpo: eso y sus fallos (filas 8, 16 y 17) son de quien pidió,
// con la dirección del pedido y el instante de lo obtenido.
func pedir(ctx context.Context, pedidor Pedidor, ec schema.Contexto, recurso pedido) (obtenido, error) {
	peticion := httpx.Peticion{Metodo: metodoDeLasPeticiones, URL: recurso.direccion, Acepta: recurso.formato}

	respuesta, err := pedidor.Pedir(ctx, ec, peticion)
	if err != nil {
		return obtenido{}, recurso.falloAlPedir(err)
	}

	switch {
	case respuesta.Ensayo:
		return obtenido{ensayo: respuesta.Descripcion()}, nil

	case respuesta.Estado >= primerEstadoCorrecto && respuesta.Estado < primerEstadoTrasLosCorrectos:
		return obtenido{cuerpo: respuesta.Cuerpo, instante: respuesta.Instante}, nil

	case respuesta.Estado == estadoNoEncontrado && recurso.inexistente != "":
		return obtenido{}, errorDeNoEncontrado(recurso.direccion, respuesta.Instante, nil, recurso.inexistente)

	default:
		return obtenido{}, errorDeFuenteNoDisponible(recurso.direccion, respuesta.Instante, nil,
			fmt.Sprintf(motivoDelEstado, respuesta.Estado, recurso.deQue))
	}
}

// falloAlPedir es el error del Pedidor visto desde la fuente: con la clase de
// claseDeLaCausa, la dirección pedida —nunca la del robots.txt que nombra httpx
// cuando no obtiene el permiso (FR-101)— y el instante que declara httpx, el de
// emisión del último intento o el del abandono (FR-095, FR-096). El error queda
// como causa: si declara su clase, como el de httpx, su texto sigue en el
// mensaje detrás de lo pedido (contrato errores-y-codigos §2), y el applet lo
// alcanza con errors.As. Uno que no es de httpx no declara instante, y el sobre
// lo fecha el montaje del kernel.
func (p pedido) falloAlPedir(causa error) *Error {
	var instante time.Time

	var deHTTPX *httpx.Error
	if errors.As(causa, &deHTTPX) {
		instante = deHTTPX.Instante
	}

	return nuevoError(claseDeLaCausa(causa), p.direccion, instante, causa,
		fmt.Sprintf(motivoDelFalloAlPedir, p.deQue))
}

// claseDeLaCausa es la clase con la que la fuente entrega el error de su
// Pedidor: la que el error declara si es una de las cuatro que dicen qué pasó
// con la petición —argumentos, no encontrado, fuente no disponible, y límite o
// términos de uso—, e «inesperado» en cualquier otro caso: sin clase, la
// inesperada misma, una fuera del vocabulario del dominio, que el *Error no
// podría declarar, o la de identidad humana, que ninguna fuente pública produce
// (FR-100). Así ningún fallo al pedir queda sin clase ni termina con el
// código 6.
func claseDeLaCausa(causa error) schema.Clase {
	var conClase schema.ConClase
	if !errors.As(causa, &conClase) {
		return schema.ClaseInesperado
	}

	switch clase := conClase.Clase(); clase {
	case schema.ClaseArgumentos, schema.ClaseNoEncontrado, schema.ClaseFuenteNoDisponible, schema.ClaseLimiteOTos:
		return clase
	default:
		return schema.ClaseInesperado
	}
}
