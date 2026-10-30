package boe

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Los motivos del bloque que no está en el índice guardado de su norma (ADR
// 0049): nombran la norma y el bloque y, si el id tiene la forma del número de un
// artículo y el índice da ese artículo con otro id, el id o los ids que tiene.
const (
	motivoFueraDelIndiceConArticulo = "la norma %s no tiene el bloque %s en su índice: el «%s» es el bloque %s"
	motivoFueraDelIndice            = "la norma %s no tiene el bloque %s en su índice: el id de un bloque es el " +
		"campo id de kitlegal boe indice %s, no el número del artículo"
)

// idDeArticulo es la forma del id compuesto con el número de un artículo: a, el
// número y, si lo hay, el sufijo en minúsculas (a21, a118, a85bis).
var idDeArticulo = regexp.MustCompile(`^a([0-9]+)([a-z]*)$`)

// fueraDelIndice es el fallo de la lectura de un bloque cuando el índice vigente
// de su norma, guardado en la caché, no tiene ese id (ADR 0049): «no encontrado»,
// con la dirección del bloque, el instante del fallo y el motivo que nombra el
// bloque del artículo si el índice lo da con otro id. No pide nada: sin el índice
// guardado, o con el id en él, el fallo es el de siempre. Solo cambia los que
// dicen que la fuente no dio el bloque —«no encontrado», «fuente no disponible»,
// «límite o TOS»— y nunca con el contexto terminado: un id que el índice no tiene
// no existe, lo diga o no la petición, y el fallo de los argumentos o el
// inesperado no son de la fuente.
func fueraDelIndice(ctx context.Context, en *invocacion, norma, bloque string, fallo error) error {
	if ctx.Err() != nil || !esFalloDeLaFuente(fallo) {
		return fallo
	}

	guardado, presente, err := leerGuardada(ctx, en, claveDelIndice(norma))
	if err != nil || !presente || tieneElBloque(guardado.Datos, bloque) {
		return fallo
	}

	var instante time.Time

	var deLaFuente *Error
	if errors.As(fallo, &deLaFuente) {
		instante = deLaFuente.Instante
	}

	return errorDeNoEncontrado(direccionDelBloque(norma, bloque), instante, nil,
		motivoDelBloqueFueraDelIndice(norma, bloque, guardado.Datos))
}

// esFalloDeLaFuente dice si el fallo es de una de las clases con las que la
// fuente no da lo pedido.
func esFalloDeLaFuente(fallo error) bool {
	var conClase schema.ConClase
	if !errors.As(fallo, &conClase) {
		return false
	}

	switch conClase.Clase() {
	case schema.ClaseNoEncontrado, schema.ClaseFuenteNoDisponible, schema.ClaseLimiteOTos:
		return true
	default:
		return false
	}
}

// tieneElBloque dice si el índice tiene una entrada con ese id.
func tieneElBloque(indice Indice, bloque string) bool {
	for _, entrada := range indice.Bloques {
		if entrada.ID == bloque {
			return true
		}
	}

	return false
}

// motivoDelBloqueFueraDelIndice es el motivo del fallo: con el título del
// artículo y los ids de las entradas del índice que lo llevan, si el id tiene la
// forma del número de un artículo y alguna lo lleva; y si no, el que remite al
// campo id del índice.
func motivoDelBloqueFueraDelIndice(norma, bloque string, indice Indice) string {
	titulo, ids := bloquesDelArticulo(indice, bloque)
	if len(ids) == 0 {
		return fmt.Sprintf(motivoFueraDelIndice, norma, bloque, norma)
	}

	return fmt.Sprintf(motivoFueraDelIndiceConArticulo, norma, bloque, titulo, strings.Join(ids, ", "))
}

// bloquesDelArticulo son, si el id tiene la forma del número de un artículo, el
// título de ese artículo —«Artículo» y el número, con el sufijo detrás si lo
// hay— y los ids de las entradas del índice cuyo título es él, en el orden del
// índice. El título casa si empieza por el del artículo y lo que sigue no
// continúa ni el número ni el sufijo: «Artículo 118.» y «Artículo 118. Expediente
// de contratación» casan con a118; «Artículo 1180» y «Artículo 118 bis», no.
func bloquesDelArticulo(indice Indice, bloque string) (string, []string) {
	partes := idDeArticulo.FindStringSubmatch(bloque)
	if partes == nil {
		return "", nil
	}

	titulo := "Artículo " + partes[1]
	if partes[2] != "" {
		titulo += " " + partes[2]
	}

	var ids []string

	for _, entrada := range indice.Bloques {
		if esElTituloDelArticulo(strings.TrimSpace(entrada.Titulo), titulo) {
			ids = append(ids, entrada.ID)
		}
	}

	return titulo, ids
}

// esElTituloDelArticulo dice si el título de una entrada es el del artículo:
// empieza por él y lo que sigue, si algo sigue, no es una letra ni un dígito.
func esElTituloDelArticulo(deLaEntrada, delArticulo string) bool {
	resto, empieza := strings.CutPrefix(deLaEntrada, delArticulo)

	switch {
	case !empieza:
		return false
	case resto == "":
		return true
	case resto[0] == ' ':
		return !empiezaPorLetraODigito(strings.TrimLeft(resto, " "))
	default:
		return !esLetraODigitoASCII(resto[0])
	}
}

// empiezaPorLetraODigito dice si el texto empieza por una letra o un dígito
// ASCII.
func empiezaPorLetraODigito(texto string) bool {
	return texto != "" && esLetraODigitoASCII(texto[0])
}
