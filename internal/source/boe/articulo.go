package boe

import (
	"context"
	"fmt"
	"time"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Los mensajes de los fallos de articulo y articulos que no son los de un
// recurso (contrato errores-y-codigos §2).
const (
	// motivoDelBloqueSinVigencia es el del fallo de los metadatos de la norma
	// cuando el bloque ya se obtuvo: nombra el bloque y dice que su vigencia no se
	// pudo comprobar, delante del fallo de los metadatos, que conserva su clase,
	// su dirección y su instante (fila 18; FR-013).
	motivoDelBloqueSinVigencia = "el bloque %s se obtuvo, pero no se pudo comprobar su vigencia: %w"
	// motivoDelBloqueQueFalla es el del primer bloque de articulos que falla:
	// nombra su id y su posición entre los bloques pedidos, delante del fallo del
	// bloque o de sus metadatos, que conserva su clase, su dirección y su instante
	// (fila 19; FR-021).
	motivoDelBloqueQueFalla = "no se ha podido resolver el bloque %s, en la posición %d de %d: %w"
	// motivoSinBloques es el de articulos sin ningún id de bloque, que nombra la
	// norma y la forma de lo que falta (FR-020).
	motivoSinBloques = "articulos necesita al menos un id de bloque de la norma %s, como a21 o da3"
)

// articulo resuelve articulo <norma> <bloque> (contrato verbos-y-salidas §3):
// valida la norma y después el bloque antes de abrir nada —fuera de su gramática,
// «argumentos» sin procedencia, que firma y fecha el kernel (contrato
// errores-y-codigos, filas 2 y 3)— y, dentro de invocar, resuelve el artículo con
// la caché de la invocación. El resultado lleva la dirección del bloque en éxito,
// en ensayo y en los fallos del bloque, y la de los metadatos en los suyos
// (FR-002, FR-101); en éxito, la fecha de consulta del artículo (FR-096).
func (f *Fuente) articulo(ctx context.Context, ec schema.Contexto, consulta ConsultaArticulo) (schema.Resultado, error) {
	if err := ValidarNorma(consulta.Norma); err != nil {
		return schema.Resultado{}, err
	}

	if err := ValidarBloque(consulta.Bloque); err != nil {
		return schema.Resultado{}, err
	}

	direccion := direccionDelBloque(consulta.Norma, consulta.Bloque)

	return f.invocar(ctx, ec, direccion, func(ctx context.Context, en *invocacion) (schema.Resultado, error) {
		resuelto, err := articuloDelBloque(ctx, en, &normaDeLaInvocacion{identificador: consulta.Norma}, consulta.Bloque)
		if err != nil {
			return resultadoDelError(err), err
		}

		return resultadoDeLaConsulta(direccion, resuelto), nil
	})
}

// articulos resuelve articulos <norma> <bloque> [<bloque>…] (contrato
// verbos-y-salidas §4): valida la norma, que haya al menos un bloque y cada uno
// de ellos antes de abrir nada —fuera de su gramática, «argumentos» sin
// procedencia, que firma y fecha el kernel, también cuando el inválido es el
// último (contrato errores-y-codigos, filas 2 y 3; data-model.md §5)— y, dentro
// de invocar, resuelve los bloques con articulosDeLosBloques. El resultado lleva
// en éxito y en ensayo la dirección de la norma, en la que se apoyan todos sus
// bloques, y en fallo la de la petición que falló (FR-002, FR-101); en éxito, la
// más antigua de las fechas de consulta de sus elementos (FR-096).
func (f *Fuente) articulos(ctx context.Context, ec schema.Contexto, consulta ConsultaArticulos) (schema.Resultado, error) {
	if err := ValidarNorma(consulta.Norma); err != nil {
		return schema.Resultado{}, err
	}

	if len(consulta.Bloques) == 0 {
		return schema.Resultado{}, errorDeArgumentos("", time.Time{}, nil, fmt.Sprintf(motivoSinBloques, consulta.Norma))
	}

	for _, bloque := range consulta.Bloques {
		if err := ValidarBloque(bloque); err != nil {
			return schema.Resultado{}, err
		}
	}

	direccion := direccionDeLaNorma(consulta.Norma)

	return f.invocar(ctx, ec, direccion, func(ctx context.Context, en *invocacion) (schema.Resultado, error) {
		resueltos, err := articulosDeLosBloques(ctx, en, consulta.Norma, consulta.Bloques)
		if err != nil {
			return resultadoDelError(err), err
		}

		return resultadoDeLaConsulta(direccion, resueltos), nil
	})
}

// articulosDeLosBloques son los artículos de los bloques de una norma, todos ya
// validados, resueltos con la caché de la invocación en el orden de data-model.md
// §7.2:
//
//  1. cada id distinto se resuelve una sola vez, en el orden de su primera
//     aparición y en secuencia, con articuloDelBloque: su entrada vigente se
//     sirve sin pedir nada y, si no la hay, se pide el bloque y su entrada se
//     escribe en cuanto se resuelve (FR-020, FR-021);
//  2. los metadatos de la norma, que solo necesitan los bloques que se piden, los
//     comparten todos los bloques de la invocación con normaDeLaInvocacion, que
//     los resuelve como mucho una vez, también bajo --dry-run (FR-020);
//  3. el primer bloque que falla detiene la invocación sin pedir los siguientes,
//     con su fallo —su clase, la dirección de la petición que falló y su
//     instante— detrás de su id y de su posición entre los pedidos, y los
//     bloques resueltos antes quedan escritos (FR-021, FR-093, FR-101; contrato
//     errores-y-codigos, fila 19);
//  4. bajo --dry-run, si algo se habría pedido, las líneas de esas peticiones en
//     su orden y sin repetir ninguna, sin datos (FR-094);
//  5. y si no, los artículos en el orden pedido, cada repetición en su posición,
//     con la más antigua de sus fechas de consulta, guardadas o de esta
//     invocación (FR-020, FR-096).
func articulosDeLosBloques(ctx context.Context, en *invocacion, norma string, bloques []string,
) (consultaResuelta[[]Articulo], error) {
	deLaNorma := &normaDeLaInvocacion{identificador: norma}
	resueltos := make(map[string]consultaResuelta[Articulo], len(bloques))

	var ensayo []string

	for posicion, bloque := range bloques {
		if _, repetido := resueltos[bloque]; repetido {
			continue
		}

		resuelto, err := articuloDelBloque(ctx, en, deLaNorma, bloque)
		if err != nil {
			return consultaResuelta[[]Articulo]{}, fmt.Errorf(motivoDelBloqueQueFalla, bloque, posicion+1, len(bloques), err)
		}

		resueltos[bloque] = resuelto
		ensayo = append(ensayo, resuelto.ensayo...)
	}

	if len(ensayo) > 0 {
		return consultaResuelta[[]Articulo]{ensayo: ensayo}, nil
	}

	articulos := make([]Articulo, 0, len(bloques))
	fechaConsulta := resueltos[bloques[0]].fechaConsulta

	for _, bloque := range bloques {
		articulos = append(articulos, resueltos[bloque].datos)
		fechaConsulta = laMasAntigua(fechaConsulta, resueltos[bloque].fechaConsulta)
	}

	return consultaResuelta[[]Articulo]{datos: articulos, fechaConsulta: fechaConsulta}, nil
}

// normaDeLaInvocacion es la norma de una invocación, ya validada, con los
// metadatos que comparten todos sus bloques: la primera vez que un bloque los
// necesita se resuelven con metadatosDeLaNorma, y después se dan de memoria, sin
// volver a leer su entrada ni a pedirlos, de modo que una invocación los pide
// como mucho una vez aunque su entrada caduque mientras tanto o, bajo --dry-run,
// no llegue a escribirse (FR-020; research.md D5). Un fallo no se recuerda: el
// primero termina la invocación.
type normaDeLaInvocacion struct {
	// identificador es el de la norma, BOE-A-<año>-<número>.
	identificador string
	// metadatos son los resueltos la primera vez, sin la línea de ensayo, y
	// resueltos dice si ya lo están.
	metadatos consultaResuelta[Metadatos]
	resueltos bool
}

// resolverMetadatos da los metadatos de la norma. La primera vez son los de
// metadatosDeLaNorma, con la línea de su petición si bajo --dry-run se habrían
// pedido; las siguientes, los mismos datos y la misma fecha de consulta sin
// ninguna línea, porque esa petición ya está descrita (FR-020, FR-094).
func (n *normaDeLaInvocacion) resolverMetadatos(ctx context.Context, en *invocacion) (consultaResuelta[Metadatos], error) {
	if n.resueltos {
		return n.metadatos, nil
	}

	resuelta, err := metadatosDeLaNorma(ctx, en, n.identificador)
	if err != nil {
		return consultaResuelta[Metadatos]{}, err
	}

	n.metadatos = consultaResuelta[Metadatos]{datos: resuelta.datos, fechaConsulta: resuelta.fechaConsulta}
	n.resueltos = true

	return resuelta, nil
}

// articuloDelBloque es el artículo de un bloque de la norma de la invocación, los
// dos ya validados, resuelto con la caché de la invocación en el orden de
// data-model.md §7.1 (FR-013):
//
//  1. la entrada vigente del artículo se sirve con su fecha, sin pedir nada ni
//     mirar los metadatos, que ya se comprobaron al escribirla; sin ella, con
//     --offline, «fuente no disponible» con la dirección del bloque (FR-090,
//     FR-092, FR-096, FR-101);
//  2. se pide el bloque con pedirElBloque, y su fallo termina sin pedir los
//     metadatos, una petición cuyo resultado no se entregaría (FR-013, FR-014);
//  3. los metadatos se resuelven con la norma de la invocación, que la primera
//     vez los sirve de su entrada vigente o los pide y los escribe (FR-090) y
//     después los da de memoria (FR-020); si fallan con el bloque obtenido, el
//     artículo falla con su clase, su dirección y su instante, y el texto no se
//     emite: nunca un artículo con la vigencia sin comprobar (FR-013; contrato
//     errores-y-codigos, fila 18);
//  4. bajo --dry-run, las líneas de las peticiones que se habrían emitido —la del
//     bloque y, si los metadatos no estaban en su entrada ni descritos ya en la
//     invocación, la suya—, sin leer ni escribir nada (FR-094);
//  5. y el artículo se compone con componerArticulo y se escribe en su entrada
//     con la vigencia de articulo y la más antigua de las fechas de consulta del
//     bloque y de los metadatos, que es la suya (FR-091, FR-096).
//
// Nada de lo que falla se escribe (FR-093).
func articuloDelBloque(ctx context.Context, en *invocacion, norma *normaDeLaInvocacion, bloque string,
) (consultaResuelta[Articulo], error) {
	clave := claveDelArticulo(norma.identificador, bloque)
	if guardado, resuelto, err := resolverSinPedir(ctx, en, clave); resuelto {
		return guardado, err
	}

	delBloque, err := pedirElBloque(ctx, en, norma.identificador, bloque)
	if err != nil {
		return consultaResuelta[Articulo]{}, err
	}

	metadatos, err := norma.resolverMetadatos(ctx, en)

	switch {
	case err != nil && delBloque.ensayo != "":
		// Bajo --dry-run el bloque no se obtuvo y el fallo, que solo puede ser de
		// la caché, no tiene ninguna vigencia sin comprobar que nombrar.
		return consultaResuelta[Articulo]{}, err
	case err != nil:
		return consultaResuelta[Articulo]{}, fmt.Errorf(motivoDelBloqueSinVigencia, bloque, err)
	case delBloque.ensayo != "":
		return consultaResuelta[Articulo]{ensayo: append([]string{delBloque.ensayo}, metadatos.ensayo...)}, nil
	}

	articulo := componerArticulo(norma.identificador, bloque, delBloque.leido, metadatos.datos)
	fechaConsulta := laMasAntigua(delBloque.instante, metadatos.fechaConsulta)

	if err := guardar(ctx, en, clave, fechaConsulta, articulo, vigenciaLarga); err != nil {
		return consultaResuelta[Articulo]{}, err
	}

	return consultaResuelta[Articulo]{datos: articulo, fechaConsulta: fechaConsulta}, nil
}

// bloqueObtenido es lo que da pedirElBloque cuando no falla: lo leído del XML y
// el instante de emisión de su petición, que es su fecha de consulta (FR-096); o,
// bajo --dry-run, la línea de la petición que se habría emitido, sin nada leído
// ni instante (ADR 0011).
type bloqueObtenido struct {
	leido    bloqueLeido
	instante time.Time
	ensayo   string
}

// pedirElBloque pide en XML el bloque de la norma con el Pedidor de la invocación
// y lo lee con leerBloque (data-model.md §3.2). El fallo de la petición es el de
// pedir, que da «no encontrado» al 404 del bloque (contrato errores-y-codigos,
// fila 6), y el cuerpo que no se puede leer —no UTF-8, XML mal formado, varias
// raíces o ningún bloque bajo la raíz— es el fallo respuestaIlegible del bloque,
// con su dirección, su instante y lo que no se pudo interpretar, nunca el cuerpo
// ni un recorte suyo (fila 16; FR-014).
func pedirElBloque(ctx context.Context, en *invocacion, norma, bloque string) (bloqueObtenido, error) {
	recurso := pedidoDelBloque(norma, bloque)

	respuesta, err := en.pedirRecurso(ctx, recurso)

	switch {
	case err != nil:
		return bloqueObtenido{}, err
	case respuesta.ensayo != "":
		return bloqueObtenido{ensayo: respuesta.ensayo}, nil
	}

	leido, err := leerBloque(respuesta.cuerpo)
	if err != nil {
		return bloqueObtenido{}, recurso.respuestaIlegible(respuesta.instante, err)
	}

	return bloqueObtenido{leido: leido, instante: respuesta.instante}, nil
}

// componerArticulo compone el data de articulo (data-model.md §2.1; FR-010,
// FR-015): la norma y el bloque pedidos; el título, el tipo, las dos fechas, la
// norma modificadora y el texto, los que leyó leerBloque; la huella del texto;
// los avisos y el ELI, los de los metadatos de la norma (FR-012); y la dirección
// pública del bloque (refs/boe.py 430).
func componerArticulo(norma, bloque string, leido bloqueLeido, metadatos Metadatos) Articulo {
	return Articulo{
		Norma:             norma,
		Bloque:            bloque,
		Titulo:            leido.titulo,
		Tipo:              leido.tipo,
		FechaVersion:      leido.fechaVersion,
		FechaVigencia:     leido.fechaVigencia,
		NormaModificadora: leido.normaModificadora,
		Texto:             leido.texto,
		HashTexto:         huellaDelTexto(leido.texto),
		Avisos:            metadatos.Avisos,
		URL:               direccionPublicaDelBloque(norma, bloque),
		URLELI:            metadatos.URLELI,
	}
}

// laMasAntigua es la más antigua de dos fechas de consulta: la que lleva lo que
// se apoya en las dos consultas, para que la cita nunca aparente más frescura que
// su parte más vieja (FR-096).
func laMasAntigua(una, otra time.Time) time.Time {
	if otra.Before(una) {
		return otra
	}

	return una
}
