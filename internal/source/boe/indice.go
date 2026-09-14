package boe

import (
	"context"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Las claves de cada bloque del índice con las que se compone EntradaDeIndice
// (refs/boe.py 387-388; data-model.md §2.4).
const (
	claveDelIDDelBloque     = "id"
	claveDelTituloDelBloque = "titulo"
)

// indice resuelve indice <norma> (contrato verbos-y-salidas §2) con
// resolverRecursoDeLaNorma: la norma se valida antes de abrir nada y el índice
// se resuelve con indiceDeLaNorma; el resultado lleva la dirección del índice.
func (f *Fuente) indice(ctx context.Context, ec schema.Contexto, consulta ConsultaIndice) (schema.Resultado, error) {
	return resolverRecursoDeLaNorma(ctx, f, ec, consulta.Norma, direccionDelIndice, indiceDeLaNorma)
}

// indiceDeLaNorma es el índice de una norma ya validada, resuelto con consultar
// sobre su entrada y con la vigencia de indice, siete días (FR-091): servido de
// la entrada vigente, o pedido en JSON, leído con leerIndice y escrito. Su 404 y
// su data vacío son «no encontrado» (FR-041).
func indiceDeLaNorma(ctx context.Context, en *invocacion, norma string) (consultaResuelta[Indice], error) {
	return consultar(ctx, en, claveDelIndice(norma), pedidoDelIndice(norma), vigenciaLarga,
		func(datos any) (Indice, error) { return leerIndice(norma, datos) })
}

// leerIndice compone el data de indice de la norma con el data de su respuesta,
// ya descartado el vacío (data-model.md §2.4; refs/boe.py 365-392): la dirección
// pública de la norma y sus bloques en el orden de la fuente, los de
// bloquesDelIndice —anidados en el primer elemento de data o planos, en lista o
// sueltos (J4, J6; FR-040, FR-070), sin los que no son objeto (J8)—. Cada bloque
// lleva su id y su título leídos con las reglas de J9, sin el marcador ? de
// refs/boe.py 387 (FR-016), y el tipo que TipoDesdeID infiere del id, vacío si
// ninguna regla casa (entrada 24 del porte anotado en doc.go). Lo que no se puede
// leer es el fallo de la lectura, que nombra el primer campo que no es texto.
func leerIndice(norma string, datos any) (Indice, error) {
	bloques := bloquesDelIndice(datos)
	entradas := make([]EntradaDeIndice, 0, len(bloques))

	for _, bloque := range bloques {
		campos := lectorDeCampos{objeto: bloque}
		id := campos.texto(claveDelIDDelBloque)
		titulo := campos.texto(claveDelTituloDelBloque)

		if campos.err != nil {
			return Indice{}, campos.err
		}

		entradas = append(entradas, EntradaDeIndice{ID: id, Titulo: titulo, Tipo: TipoDesdeID(id)})
	}

	return Indice{Norma: norma, URL: direccionPublicaDeLaNorma(norma), Bloques: entradas}, nil
}
