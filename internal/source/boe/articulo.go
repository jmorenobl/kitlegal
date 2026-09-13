package boe

import (
	"context"
	"fmt"
	"time"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// motivoDelBloqueSinVigencia es el del fallo de los metadatos de la norma cuando
// el bloque ya se obtuvo: nombra el bloque y dice que su vigencia no se pudo
// comprobar, delante del fallo de los metadatos, que conserva su clase, su
// dirección y su instante (contrato errores-y-codigos §2, fila 18; FR-013).
const motivoDelBloqueSinVigencia = "el bloque %s se obtuvo, pero no se pudo comprobar su vigencia: %w"

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
		resuelto, err := articuloDelBloque(ctx, en, consulta.Norma, consulta.Bloque)
		if err != nil {
			return resultadoDelError(err), err
		}

		return resultadoDeLaConsulta(direccion, resuelto), nil
	})
}

// articuloDelBloque es el artículo de un bloque de una norma, los dos ya
// validados, resuelto con la caché de la invocación en el orden de data-model.md
// §7.1 (FR-013):
//
//  1. la entrada vigente del artículo se sirve con su fecha, sin pedir nada ni
//     mirar los metadatos, que ya se comprobaron al escribirla; sin ella, con
//     --offline, «fuente no disponible» con la dirección del bloque (FR-090,
//     FR-092, FR-096, FR-101);
//  2. se pide el bloque con pedirElBloque, y su fallo termina sin pedir los
//     metadatos, una petición cuyo resultado no se entregaría (FR-013, FR-014);
//  3. los metadatos de la norma se resuelven con metadatosDeLaNorma, que los
//     sirve de su entrada vigente o los pide y los escribe (FR-090); si fallan
//     con el bloque obtenido, el artículo falla con su clase, su dirección y su
//     instante, y el texto no se emite: nunca un artículo con la vigencia sin
//     comprobar (FR-013; contrato errores-y-codigos, fila 18);
//  4. bajo --dry-run, las líneas de las peticiones que se habrían emitido —la del
//     bloque y, si su entrada no estaba vigente, la de los metadatos—, sin leer
//     ni escribir nada (FR-094);
//  5. y el artículo se compone con componerArticulo y se escribe en su entrada
//     con la vigencia de articulo y la más antigua de las fechas de consulta del
//     bloque y de los metadatos, que es la suya (FR-091, FR-096).
//
// Nada de lo que falla se escribe (FR-093).
func articuloDelBloque(ctx context.Context, en *invocacion, norma, bloque string) (consultaResuelta[Articulo], error) {
	clave := claveDelArticulo(norma, bloque)
	if guardado, resuelto, err := resolverSinPedir(ctx, en, clave); resuelto {
		return guardado, err
	}

	delBloque, err := pedirElBloque(ctx, en, norma, bloque)
	if err != nil {
		return consultaResuelta[Articulo]{}, err
	}

	metadatos, err := metadatosDeLaNorma(ctx, en, norma)

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

	articulo := componerArticulo(norma, bloque, delBloque.leido, metadatos.datos)
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
