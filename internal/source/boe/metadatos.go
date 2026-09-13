package boe

import (
	"context"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Las claves del objeto de metadatos con las que se compone Metadatos, salvo las
// que miran los avisos —estatus_derogacion, vigencia_agotada y
// estado_consolidacion con su codigo—, que son las de avisos.go (refs/boe.py
// 465-482; data-model.md §2.5).
const (
	claveDelTitulo              = "titulo"
	claveDelRango               = "rango"
	claveDelNumeroOficial       = "numero_oficial"
	claveDeLaFechaDeDisposicion = "fecha_disposicion"
	claveDeLaFechaDePublicacion = "fecha_publicacion"
	claveDeLaFechaDeVigencia    = "fecha_vigencia"
	claveDeLaURLELI             = "url_eli"
	// subclaveDelTexto es la del texto de rango y de estado_consolidacion cuando
	// son objetos (refs/boe.py 466 y 469).
	subclaveDelTexto = "texto"
)

// metadatos resuelve metadatos <norma> (contrato verbos-y-salidas §5) con
// resolverRecursoDeLaNorma: la norma se valida antes de abrir nada y los
// metadatos se resuelven con metadatosDeLaNorma; el resultado lleva la
// dirección de los metadatos.
func (f *Fuente) metadatos(ctx context.Context, ec schema.Contexto, consulta ConsultaMetadatos) (schema.Resultado, error) {
	return resolverRecursoDeLaNorma(ctx, f, ec, consulta.Norma, direccionDeLosMetadatos, metadatosDeLaNorma)
}

// metadatosDeLaNorma son los metadatos de una norma ya validada, resueltos con
// consultar sobre su entrada —la misma para metadatos, articulo y articulos
// (FR-090)— y con la vigencia de metadatos, 300 s (FR-091): servidos de la
// entrada vigente, o pedidos en JSON, leídos con leerMetadatos y escritos. Su 404
// y su data vacío son «no encontrado» (FR-051).
func metadatosDeLaNorma(ctx context.Context, en *invocacion, norma string) (consultaResuelta[Metadatos], error) {
	return consultar(ctx, en, claveDeLosMetadatos(norma), pedidoDeLosMetadatos(norma), vigenciaCorta,
		func(datos any) (Metadatos, error) { return leerMetadatos(norma, datos) })
}

// leerMetadatos compone el data de metadatos de la norma con el data de su
// respuesta, ya descartado el vacío (data-model.md §2.5; refs/boe.py 463-482):
// el objeto es el primer elemento de data, o data si llega suelto (J5); cada
// campo de texto se lee con las reglas de J9, sin el marcador ? de refs/boe.py
// (FR-016); rango y el texto de estado_consolidacion admiten un objeto o una
// cadena, y su código, solo un objeto; y los avisos son los de avisosDe, con las
// frases de _check_vigencia y no las de cmd_metadatos (FR-050; entrada 17 del
// porte anotado en doc.go). Lo que no se puede leer es el fallo de la lectura,
// que nombra el primer campo que no es texto.
func leerMetadatos(norma string, datos any) (Metadatos, error) {
	objeto, err := primerElemento(datos)
	if err != nil {
		return Metadatos{}, err
	}

	campos := lectorDeCampos{objeto: objeto}
	metadatos := Metadatos{
		Norma:             norma,
		Titulo:            campos.texto(claveDelTitulo),
		Rango:             campos.textoDelObjetoOCadena(claveDelRango, subclaveDelTexto),
		NumeroOficial:     campos.texto(claveDelNumeroOficial),
		FechaDisposicion:  campos.texto(claveDeLaFechaDeDisposicion),
		FechaPublicacion:  campos.texto(claveDeLaFechaDePublicacion),
		FechaVigencia:     campos.texto(claveDeLaFechaDeVigencia),
		EstatusDerogacion: campos.texto(claveDeEstatusDeDerogacion),
		VigenciaAgotada:   campos.texto(claveDeVigenciaAgotada),
		EstadoConsolidacion: EstadoDeConsolidacion{
			Codigo: campos.textoDelObjeto(claveDeEstadoDeConsolidacion, claveDeCodigoDeConsolidacion),
			Texto:  campos.textoDelObjetoOCadena(claveDeEstadoDeConsolidacion, subclaveDelTexto),
		},
		URLELI: campos.texto(claveDeLaURLELI),
		Avisos: avisosDe(objeto),
	}

	if campos.err != nil {
		return Metadatos{}, campos.err
	}

	return metadatos, nil
}

// lectorDeCampos lee como texto, con las reglas de J9, los campos de un objeto
// de la respuesta y se queda con el primer fallo, para que quien compone un data
// de muchos campos los lea seguidos, en el orden en que los compone, y mire el
// fallo una sola vez al final. Tras un fallo, cada lectura da la cadena vacía.
type lectorDeCampos struct {
	objeto map[string]any
	err    error
}

// texto es textoDe del campo.
func (l *lectorDeCampos) texto(campo string) string {
	return l.anota(textoDe(l.objeto, campo))
}

// textoDelObjeto es textoDelObjetoDe del subcampo del campo.
func (l *lectorDeCampos) textoDelObjeto(campo, subcampo string) string {
	return l.anota(textoDelObjetoDe(l.objeto, campo, subcampo))
}

// textoDelObjetoOCadena es textoDelObjetoOCadenaDe del subcampo del campo.
func (l *lectorDeCampos) textoDelObjetoOCadena(campo, subcampo string) string {
	return l.anota(textoDelObjetoOCadenaDe(l.objeto, campo, subcampo))
}

// anota devuelve el texto leído, o la cadena vacía si esta lectura u otra
// anterior falló, y se queda con el primer fallo.
func (l *lectorDeCampos) anota(texto string, err error) string {
	switch {
	case l.err != nil:
		return ""
	case err != nil:
		l.err = err

		return ""
	default:
		return texto
	}
}
