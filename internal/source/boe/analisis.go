package boe

import (
	"context"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Las claves de cada materia y de cada referencia con las que se compone
// Analisis (refs/boe.py 527-581; data-model.md §2.6). El texto de relacion,
// cuando es un objeto, se lee con subclaveDelTexto; y la materia o la nota que no
// es un objeto se nombra en el mensaje de su fallo con su envoltorio,
// envoltorioDeMateria o envoltorioDeNota.
const (
	// claveDelCodigoDeLaMateria y claveDelTextoDeLaMateria son las de la materia
	// que es un objeto (refs/boe.py 533).
	claveDelCodigoDeLaMateria = "codigo"
	claveDelTextoDeLaMateria  = "texto"
	// claveDeLaRelacion y claveDeLaNormaReferida son las de toda referencia, y
	// claveDelTextoDeLaReferencia, la que solo se lee en las anteriores
	// (refs/boe.py 557-562 y 577-581).
	claveDeLaRelacion           = "relacion"
	claveDeLaNormaReferida      = "id_norma"
	claveDelTextoDeLaReferencia = "texto"
)

// analisis resuelve analisis <norma> (contrato verbos-y-salidas §6) con
// resolverRecursoDeLaNorma: la norma se valida antes de abrir nada y el análisis
// se resuelve con analisisDeLaNorma; el resultado lleva la dirección del
// análisis.
func (f *Fuente) analisis(ctx context.Context, ec schema.Contexto, consulta ConsultaAnalisis) (schema.Resultado, error) {
	return resolverRecursoDeLaNorma(ctx, f, ec, consulta.Norma, direccionDelAnalisis, analisisDeLaNorma)
}

// analisisDeLaNorma es el análisis de una norma ya validada, resuelto con
// consultar sobre su entrada y con la vigencia de analisis, siete días (FR-091):
// servido de la entrada vigente, o pedido en JSON, leído con leerAnalisis y
// escrito. Su 404 y su data vacío son «no encontrado» (FR-061).
func analisisDeLaNorma(ctx context.Context, en *invocacion, norma string) (consultaResuelta[Analisis], error) {
	return consultar(ctx, en, claveDelAnalisis(norma), pedidoDelAnalisis(norma), vigenciaLarga,
		func(datos any) (Analisis, error) { return leerAnalisis(norma, datos) })
}

// leerAnalisis compone el data de analisis de la norma con el data de su
// respuesta, ya descartado el vacío (data-model.md §2.6; refs/boe.py 522-581): el
// objeto es el primer elemento de data, o data si llega suelto (J5); sus
// materias, sus notas y sus referencias anteriores y posteriores, abiertas de sus
// envoltorios con materiasDe, notasDe, referenciasAnterioresDe y
// referenciasPosterioresDe —en lista o sueltas (J4, J7; FR-070)—, se leen en el
// orden de la fuente con leerMateria, leerNota, leerReferenciaAnterior y
// leerReferenciaPosterior; y una lista sin elementos va vacía, nunca nula. Lo que
// no se puede leer es el fallo de la lectura, que nombra el primer campo que no
// es texto, en ese orden.
func leerAnalisis(norma string, datos any) (Analisis, error) {
	objeto, err := primerElemento(datos)
	if err != nil {
		return Analisis{}, err
	}

	materias, err := leerCadaUno(materiasDe(objeto), leerMateria)
	if err != nil {
		return Analisis{}, err
	}

	notas, err := leerCadaUno(notasDe(objeto), leerNota)
	if err != nil {
		return Analisis{}, err
	}

	anteriores, err := leerCadaUno(referenciasAnterioresDe(objeto), leerReferenciaAnterior)
	if err != nil {
		return Analisis{}, err
	}

	posteriores, err := leerCadaUno(referenciasPosterioresDe(objeto), leerReferenciaPosterior)
	if err != nil {
		return Analisis{}, err
	}

	return Analisis{
		Norma:       norma,
		Materias:    materias,
		Notas:       notas,
		Referencias: Referencias{Anteriores: anteriores, Posteriores: posteriores},
	}, nil
}

// leerCadaUno lee con leer cada elemento, en su orden, y se detiene en el primer
// fallo. La lista que devuelve nunca es nula: sin elementos, va vacía.
func leerCadaUno[E, T any](elementos []E, leer func(E) (T, error)) ([]T, error) {
	leidos := make([]T, 0, len(elementos))

	for _, elemento := range elementos {
		leido, err := leer(elemento)
		if err != nil {
			return nil, err
		}

		leidos = append(leidos, leido)
	}

	return leidos, nil
}

// leerMateria lee una materia ya abierta de su envoltorio (refs/boe.py 531-535):
// la que es un objeto, con su código y su texto; y la que no lo es aporta solo su
// texto, leído como el valor de un campo que se llama como su envoltorio. Los dos
// siguen las reglas de J9, sin el marcador ? de refs/boe.py 533 (FR-016), de
// modo que una materia nula es la de texto vacío, y lo que no es texto es el
// fallo de la lectura que nombra el campo.
func leerMateria(materia any) (Materia, error) {
	objeto, esObjeto := materia.(map[string]any)
	if !esObjeto {
		texto, err := textoDelValor(materia, envoltorioDeMateria)
		if err != nil {
			return Materia{}, err
		}

		return Materia{Texto: texto}, nil
	}

	campos := lectorDeCampos{objeto: objeto}
	leida := Materia{
		Codigo: campos.texto(claveDelCodigoDeLaMateria),
		Texto:  campos.texto(claveDelTextoDeLaMateria),
	}

	if campos.err != nil {
		return Materia{}, campos.err
	}

	return leida, nil
}

// leerNota lee una nota ya abierta de su envoltorio (refs/boe.py 538-541) como el
// valor de un campo que se llama como su envoltorio, con las reglas de J9.
func leerNota(nota any) (string, error) {
	return textoDelValor(nota, envoltorioDeNota)
}

// leerReferenciaAnterior lee una referencia a una norma que esta modifica o
// deroga (refs/boe.py 557-564): su relación —el texto del objeto o la cadena—, la
// norma referida y su texto completo, sin el recorte a 200 caracteres de
// refs/boe.py 564 (FR-060; entrada 29 del porte anotado en doc.go), cada uno con
// las reglas de J9.
func leerReferenciaAnterior(referencia map[string]any) (ReferenciaAnterior, error) {
	campos := lectorDeCampos{objeto: referencia}
	leida := ReferenciaAnterior{
		Relacion: campos.textoDelObjetoOCadena(claveDeLaRelacion, subclaveDelTexto),
		Norma:    campos.texto(claveDeLaNormaReferida),
		Texto:    campos.texto(claveDelTextoDeLaReferencia),
	}

	if campos.err != nil {
		return ReferenciaAnterior{}, campos.err
	}

	return leida, nil
}

// leerReferenciaPosterior lee una referencia a una norma que modifica esta
// (refs/boe.py 577-581): su relación y la norma referida, con las reglas de J9.
// Su texto no se lee, porque refs/boe.py no lo presenta (data-model.md §2.6).
func leerReferenciaPosterior(referencia map[string]any) (ReferenciaPosterior, error) {
	campos := lectorDeCampos{objeto: referencia}
	leida := ReferenciaPosterior{
		Relacion: campos.textoDelObjetoOCadena(claveDeLaRelacion, subclaveDelTexto),
		Norma:    campos.texto(claveDeLaNormaReferida),
	}

	if campos.err != nil {
		return ReferenciaPosterior{}, campos.err
	}

	return leida, nil
}
