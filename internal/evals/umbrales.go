package evals

import (
	"fmt"
	"strconv"
	"strings"
)

// La proporción de las respuestas medidas del modelo que decide que admite el
// umbral de las que no activaron la skill: ninguna (FR-041 de H7.4). Va en el
// paquete, y no en la definición del job, que solo da el objetivo de duración
// (research D5 de H7.3; FR-002, FR-004 de H7.3).
const umbralDeRespuestasSinActivar = 0

// La proporción de esas mismas respuestas que admite el umbral de las que
// llevan una cita de sentencia sin documento cotejado o un ECLI sin origen:
// ninguna (contracts/evals-jurisprudencia.md §4 de H23; FR-060 de H23).
const umbralDeRespuestasSinDocumento = 0

// La proporción de las respuestas medidas del modelo que decide en las evals
// que piden el cotejo que admite el umbral de las de una sesión que no cotejó
// ningún documento: ninguna (contracts/umbral-documento-sin-cotejar.md §2 de
// H23.1; FR-001, FR-004 de H23.1).
const umbralDeRespuestasSinCotejar = 0

// La proporción de las respuestas medidas del modelo que decide en las evals de
// plazo que admite el umbral de las que no llevan la línea del plazo con la
// fecha esperada, y la de las de las evals de cobertura que admite el de las que
// no llevan la declaración de no cubierto con la norma esperada o llevan una
// línea del plazo: ninguna (contracts/evals-legal-core.md §3 de H9; FR-053,
// FR-054 de H9).
const (
	umbralDeRespuestasSinFecha    = 0
	umbralDeRespuestasSinDeclarar = 0
)

// La proporción de las respuestas medidas del modelo que decide en las evals
// que esperan una redacción fechada que admite el umbral de las de una sesión
// que no leyó el bloque de la eval en su fecha o que no llevan las líneas que la
// eval espera: ninguna (contracts/evals-redaccion-fechada.md §3 de H20; FR-054,
// FR-055 de H20).
const umbralDeRedaccionesSinFechar = 0

// Lo que admiten los umbrales del juez que no salen de su declaración de clases
// (contracts/informe-del-job.md §2 de H24): la proporción de los casos
// etiquetados de su medida versionada que no dan lo que dice su etiqueta,
// ninguno (FR-032), y los segundos de los votos de las respuestas de un modo
// (FR-033). El de cada clase lo da clases.yaml.
const (
	umbralDeLaMedidaDelJuez = 0
	segundosDelJuezPorModo  = 900
)

// Lo que lleva cada umbral del informe (contrato informe-del-job §1 de H7.3;
// contracts/informe-del-job.md §2 de H7.4 y de H24; data-model §1;
// contracts/evals-en-dos-modos.md §5.1 de H21; data-model §10 de H21). El de
// las respuestas de un modelo en un modo se nombra con lo que mide seguido del
// modelo y del modo, y el de la duración, con el modo; su descripción los
// nombra, primero el modelo y después el modo; y la única comparación es «<=»,
// que informe.md escribe «≤».
//
// Desde H24, lo que mide un umbral de las respuestas puede ser una clase del
// juez, y entonces se nombra con ella; los dos de la medida del juez no son de
// ningún modo, y se nombran con su prefijo, la clase que decide y lo que
// cuentan; y los votos del juez tienen su duración, aparte de la de las
// sesiones.
//
// Desde H23, el de las respuestas con una cita de sentencia sin documento
// cotejado o con un ECLI sin origen se nombra con su prefijo, como el de las
// que no activaron la skill (contracts/evals-jurisprudencia.md §4 de H23).
//
// Desde H26 (contracts/umbrales-del-informe.md §2 de H26), no todos se miden
// sobre las sesiones de un solo modo: el umbral conjunto de una clase del juez
// que decide junta las respuestas de todos los modos, se nombra con la clase y
// el modelo, sin modo, y su descripción lo dice.
//
// Desde H23.1, el de las respuestas de las evals que piden el cotejo cuya
// sesión no cotejó ningún documento se nombra con su prefijo, como el de
// cita_sin_documento, y su descripción dice que su total son solo esas
// respuestas (contracts/umbral-documento-sin-cotejar.md §2 de H23.1; research D6
// de H23.1).
//
// Desde H9, los de las respuestas de las evals de plazo sin la línea del plazo
// con la fecha esperada y de las de cobertura sin la declaración de no cubierto
// con la norma esperada o con una línea del plazo se nombran con su prefijo,
// como el de documento_sin_cotejar, y su descripción dice que su total son solo
// las respuestas de esas evals (contracts/evals-legal-core.md §3 de H9; research
// D20 de H9).
//
// Desde H20, el de las respuestas de las evals que esperan una redacción fechada
// cuya sesión no leyó el bloque de la eval en su fecha o que no llevan las
// líneas que la eval espera se nombra con su prefijo, como el de plazo_sin_fecha,
// y su descripción dice que su total son solo las respuestas de esas evals
// (contracts/evals-redaccion-fechada.md §3 de H20; research D14 de H20).
const (
	prefijoDelUmbralSinActivar             = "sin_activar:"
	prefijoDelUmbralDeCitaSinDocumento     = "cita_sin_documento:"
	prefijoDelUmbralDeDocumentoSinCotejar  = "documento_sin_cotejar:"
	prefijoDelUmbralDePlazoSinFecha        = "plazo_sin_fecha:"
	prefijoDelUmbralDeCoberturaSinDeclarar = "cobertura_sin_declarar:"
	prefijoDelUmbralDeRedaccionSinFechar   = "redaccion_sin_fechar:"
	prefijoDelUmbralDeLaMedidaDelJuez      = "medida_del_juez:"
	prefijoDelUmbralDeLaDuracion           = "duracion_de_las_sesiones:"
	prefijoDelUmbralDeLaDuracionDelJuez    = "duracion_del_juez:"
	separadorDelModoDelUmbral              = ":"

	sufijoDeLosDefectosSinMarcar = ":defectos_sin_marcar"
	sufijoDeLosCorrectosMarcados = ":correctos_marcados"

	descripcionDelUmbralSinActivar = "Respuestas de %s en el modo %s sin la skill activada, sobre sus respuestas " +
		"medidas en las evals que la activan"
	descripcionDelUmbralDeCitaSinDocumento = "Respuestas de %s en el modo %s con una cita de sentencia sin " +
		"documento cotejado o con un ECLI que no viene de una operación ni de la pregunta, sobre sus respuestas " +
		"medidas en las evals que activan la skill"
	descripcionDelUmbralDeDocumentoSinCotejar = "Respuestas de %s en el modo %s cuya sesión no ejecutó ningún cita " +
		"cotejar que consultara y terminara bien, sobre sus respuestas medidas en las evals que activan la skill y " +
		"piden cita cotejar"
	descripcionDelUmbralDePlazoSinFecha = "Respuestas de %s en el modo %s sin la línea ⚠ PLAZO: con la fecha " +
		"esperada, sobre sus respuestas medidas en las evals que esperan un plazo"
	descripcionDelUmbralDeCoberturaSinDeclarar = "Respuestas de %s en el modo %s sin la línea ⚠ NO CUBIERTO: con " +
		"la norma esperada, o con una línea ⚠ PLAZO:, sobre sus respuestas medidas en las evals que esperan una " +
		"declaración de no cubierto"
	descripcionDelUmbralDeRedaccionSinFechar = "Respuestas de %s en el modo %s cuya sesión no leyó con boe articulo " +
		"el bloque de la eval en su fecha, o sin las líneas ⚠ REDACCIÓN A FECHA:, ⚠ REDACCIÓN DE HOY: y ⚠ SIN " +
		"REDACCIÓN VIGENTE: que la eval espera, sobre sus respuestas medidas en las evals que esperan una redacción " +
		"fechada"
	descripcionDeLaClaseQueDecide = "Respuestas de %s en el modo %s que el juez marca en %s con sus tres votos, " +
		"sobre sus respuestas juzgadas en las evals que activan la skill"
	descripcionDeLaClaseQueSePublica = "Respuestas de %s en el modo %s con sí en %s en el primer voto del juez, " +
		"sobre sus respuestas juzgadas en las evals que activan la skill"
	descripcionDelUmbralConjunto = "Respuestas de %s, con todos los modos juntos, que el juez marca en %s con sus " +
		"tres votos, sobre sus respuestas juzgadas en las evals que activan la skill"
	descripcionDeLosDefectosSinMarcar = "Casos etiquetados como defecto que el juez no marca en %s, sobre los casos " +
		"etiquetados como defecto de su medida versionada"
	descripcionDeLosCorrectosMarcados = "Casos etiquetados como correctos que el juez marca en %s, sobre los casos " +
		"etiquetados como correctos de su medida versionada"
	descripcionDelUmbralDeLaDuracion = "Segundos de la tanda del modo %s, desde que se prepara su primera sesión " +
		"hasta que termina la última"
	descripcionDelUmbralDeLaDuracionDelJuez = "Segundos de los votos del juez sobre las respuestas del modo %s, " +
		"desde que empieza el primero hasta que termina el último"

	comparacionMenorOIgual = "<="
	signoMenorOIgual       = "≤ "
)

// Textos fijos de los umbrales en informe.md y en los motivos (contrato
// informe-del-job §2 y §4 de H7.3; contracts/informe-del-job.md §4 de H24): la
// celda de uno que no decide; lo que sigue a la medida y a la condición de uno
// que no se cumple; y, en el del conjunto de una clase del juez, cómo nombra a
// cada respuesta marcada —su sesión, con su modo entre paréntesis, y, entre
// comillas, las frases de los votos que la marcan— y qué separa una de otra
// (contracts/umbrales-del-informe.md §4 de H26; research D4 de H26). El de
// cita_sin_documento nombra igual a cada respuesta que cuenta —su sesión y cada
// ECLI que la hace contar, con su condición entre paréntesis—, con los mismos
// separadores (contracts/evals-jurisprudencia.md §4 de H23; research D21 de
// H23). Y el de documento_sin_cotejar nombra cada respuesta que cuenta por su
// sesión y nada más, separadas igual (contracts/umbral-documento-sin-cotejar.md
// §3 de H23.1; research D7 de H23.1), como, desde H9, los de plazo_sin_fecha y
// cobertura_sin_declarar (contracts/evals-legal-core.md §3 de H9) y, desde H20,
// el de redaccion_sin_fechar (contracts/evals-redaccion-fechada.md §3 de H20).
const (
	soloSePublica      = "no: solo se publica"
	motivoDeUnUmbral   = "umbral %s: %s, y tiene que ser %s"
	motivoDeLaDuracion = "%s: %s s, y tiene que ser %s s"

	separadorDeLasMarcadas = "; "
	separadorDeLasFrases   = " · "
	comillaQueAbre         = "«"
	comillaQueCierra       = "»"

	sesionConSuModo    = "%s (modo %s)"
	ecliConSuCondicion = "%s (%s)"
)

// Umbral es un umbral del informe del job de evals, con los campos, las claves y
// los invariantes del contrato de umbrales del ADR 0029 (data-model §1 de H7.3).
// Sus claves JSON, en el orden de sus campos, son las de cada elemento de
// umbrales en informe.json.
type Umbral struct {
	// Nombre es lo que cita el plan en evals:<skill>:<nombre>, único en el
	// informe: sin_activar:<modelo>:<modo> o duracion_de_las_sesiones:<modo>, con
	// <modo> orden u herramienta (data-model §10 de H21); y, con juez,
	// <clase>:<modelo>:<modo>, medida_del_juez:<clase>:defectos_sin_marcar,
	// medida_del_juez:<clase>:correctos_marcados y duracion_del_juez:<modo>
	// (contracts/informe-del-job.md §2 de H24) y, de cada clase que decide, su
	// umbral conjunto, <clase>:<modelo>, sin modo
	// (contracts/umbrales-del-informe.md §2 de H26); con alguna eval que
	// declara sentencias, cita_sin_documento:<modelo>:<modo>
	// (contracts/evals-jurisprudencia.md §4 de H23); con alguna eval que pide
	// el cotejo, documento_sin_cotejar:<modelo>:<modo>
	// (contracts/umbral-documento-sin-cotejar.md §2 de H23.1); y, con alguna
	// eval de plazo, plazo_sin_fecha:<modelo>:<modo>, y con alguna de
	// cobertura, cobertura_sin_declarar:<modelo>:<modo>
	// (contracts/evals-legal-core.md §3 de H9).
	Nombre string `json:"nombre"`

	// Descripcion dice en una línea qué se mide, en qué modo —o con todos los
	// modos juntos— y sobre qué respuestas.
	Descripcion string `json:"descripcion"`

	// Medida es lo medido: en las sesiones de su modo, las respuestas sin la
	// skill activada, las que el juez marca en una clase, las que llevan una
	// cita sin documento cotejado o un ECLI sin origen, las de una sesión que no
	// cotejó ningún documento, las de una eval de plazo sin la línea del plazo
	// con la fecha esperada, las de una eval de cobertura sin la declaración de
	// no cubierto con la norma esperada o con una línea del plazo, o los
	// segundos de su tanda o de sus votos; en el
	// umbral conjunto de una clase, las que el juez marca en ella en las
	// sesiones de todos los modos; y, en la medida versionada del juez, los
	// casos que no dan lo que dice su etiqueta.
	Medida float64 `json:"medida"`

	// Total, solo en los de las respuestas y en los de la medida del juez, es
	// aquello sobre lo que se mide: las respuestas medidas del modelo en ese
	// modo en las evals que activan la skill —en el umbral conjunto de una
	// clase, las de todos los modos; en documento_sin_cotejar, solo las de las
	// evals que piden el cotejo; en plazo_sin_fecha, solo las de las evals de
	// plazo; y en cobertura_sin_declarar, solo las de las de cobertura—, o los
	// casos con esa etiqueta. Con él, lo que
	// se compara es la proporción Medida/Total, 0 si Total es 0. Sin él, la
	// propia Medida.
	Total *int `json:"total,omitzero"`

	// Comparacion es siempre «<=» en este hito, y Umbral, el valor con el que se
	// compara, en la unidad de lo que se compara: una proporción, o segundos.
	Comparacion string  `json:"comparacion"`
	Umbral      float64 `json:"umbral"`

	// Cumple es el resultado de la comparación, en coma flotante de doble
	// precisión y sin redondeos.
	Cumple bool `json:"cumple"`

	// Decide dice si incumplirlo pone el veredicto en fallo; si no, solo se
	// publica.
	Decide bool `json:"decide"`
}

// umbralesDelInforme son los umbrales del informe, sin ningún caso por skill
// (research D5 de H7.3 y D14 de H7.4; FR-002, FR-004 a FR-006, FR-051 de H7.3;
// FR-040 a FR-048 de H7.4): los de las respuestas del modelo que decide, que
// solo existen si la skill tiene juez (research D13 de H24) o alguna eval que
// declara sentencias (research D20 de H23), y, detrás, los de la duración de
// las sesiones si el job da un objetivo mayor que 0, sin total y decidiendo.
// Nil, [] en informe.json, si no hay ninguno. respuestas es lo que se mide de
// esas respuestas, nil si la skill no tiene umbrales de ellas.
//
// Desde H21 (contracts/evals-en-dos-modos.md §5.1 de H21; data-model §10;
// research.md D19; FR-043 a FR-045), los de las respuestas y los de la duración
// son de un modo del plan y lo nombran: los de las respuestas van por modo, los
// del modo orden delante, y detrás de ellos, el de la duración de cada modo, en
// su orden, que mide los segundos de la tanda de ese modo frente al mismo
// objetivo. Ninguno cuenta las sesiones o la tanda de las evals sin binario ni
// servidor, y ninguno de un modo suma las medidas de otro: las de dos modos
// solo las junta, desde H26, el umbral conjunto de una clase del juez.
//
// Desde H24 (contracts/informe-del-job.md §2 y §6 de H24; FR-030 a FR-034),
// ninguno mide las expresiones de la lista —salen
// expresiones_prohibidas:<modelo>:<modo>, de todos los modelos, y
// redaccion_no_leida:<modelo>:<modo>— y los de una skill con juez son, en este
// orden:
//
//   - por cada modo, el de las respuestas sin la skill activada y el de cada
//     clase del juez, sobre las mismas respuestas (respuestasMedidas.umbrales);
//   - desde H26, el umbral conjunto de cada clase que decide
//     (umbralesConjuntos);
//   - los dos de la medida versionada del juez de cada clase que decide
//     (umbralesDeLaMedida);
//   - los de la duración de las sesiones, como hasta ahora;
//   - y, por cada modo, el de la duración de sus votos (umbralesDeLaDuracion).
//
// Desde H23 (contracts/evals-jurisprudencia.md §4 de H23; research D20 de H23;
// FR-060 a FR-063 de H23), los de las respuestas de cada modo son tres cosas,
// cada una por su razón: el de las que no activaron la skill, con juez o con
// sentencias; el de cada clase, con juez; y el de las que llevan una cita sin
// documento cotejado o un ECLI sin origen, con sentencias. Una skill sin juez
// con alguna eval que declara sentencias lleva, de cada modo, sin_activar y
// cita_sin_documento, y los dos deciden. Los de una skill sin juez ni
// sentencias siguen siendo ninguno.
//
// Desde H25 (contracts/juez-de-jurisprudencia.md §5 de H25; FR-020 a FR-026 de
// H25), jurisprudencia tiene juez y sigue sin objetivo: lleva, de cada modo,
// sin_activar, afirma_lo_no_leido, afirma_que_existe y cita_sin_documento; los
// dos de la medida versionada; y duracion_del_juez de cada modo, sin
// duracion_de_las_sesiones.
//
// Desde H26 (contracts/umbrales-del-informe.md §1 a §3 de H26; data-model §3
// de H26; research D1 y D2 de H26; FR-001, FR-002, FR-005, FR-010 a FR-012 de
// H26), lo que decide de una clase del juez es su umbral conjunto, que junta
// las respuestas de todos los modos y va detrás de los umbrales de las
// respuestas de todos ellos y delante de los dos de la medida; el de la clase
// en un modo es un recuento, que se publica sin decidir, también el de la que
// decide. Con el plan del job, los dos modos, los umbrales de boe-legislacion,
// con su objetivo, son trece, de los que deciden nueve: de cada modo,
// sin_activar y los recuentos de afirma_lo_no_leido y de cuenta_su_proceso; el
// conjunto de afirma_lo_no_leido; los dos de la medida; y
// duracion_de_las_sesiones y duracion_del_juez de cada modo.
//
// Desde H23.1 (contracts/umbral-documento-sin-cotejar.md §1 de H23.1; data-model
// §4 de H23.1; research D4 y D6 de H23.1; FR-001, FR-006 de H23.1), los de las
// respuestas de cada modo de una skill que ya los tiene llevan, detrás de
// cita_sin_documento, documento_sin_cotejar si alguna de sus evals pide el
// cotejo, que decide con 0. Con el plan del job, los de jurisprudencia, sin
// objetivo, son quince, de los que deciden once: de cada modo, sin_activar, los
// recuentos de afirma_lo_no_leido y de afirma_que_existe, cita_sin_documento y
// documento_sin_cotejar; el conjunto de afirma_lo_no_leido; los dos de la
// medida; y duracion_del_juez de cada modo. Ninguna eval de boe-legislacion
// pide el cotejo, y sus trece no cambian. En las dos, no deciden los cuatro
// recuentos: los de las dos clases del juez en los dos modos.
//
// Desde H9 (contracts/evals-legal-core.md §3 de H9; data-model §3 de H9;
// research D20 de H9; FR-053 a FR-056 de H9), una skill tiene umbrales de sus
// respuestas también si alguna de sus evals es de plazo o de cobertura, y los
// de cada modo llevan, detrás de documento_sin_cotejar, plazo_sin_fecha si
// alguna es de plazo y cobertura_sin_declarar si alguna es de cobertura, los
// dos decidiendo con 0. Con el plan del job, los de legal-core, sin juez ni
// objetivo, son seis, y deciden los seis: de cada modo, sin_activar,
// plazo_sin_fecha y cobertura_sin_declarar. Ninguna eval de boe-legislacion ni
// de jurisprudencia es de plazo ni de cobertura, y sus trece y sus quince no
// cambian.
//
// Desde H20 (contracts/evals-redaccion-fechada.md §3 de H20; data-model §7.4 de
// H20; research D14 de H20; FR-054 a FR-056 de H20), una skill tiene umbrales
// de sus respuestas también si alguna de sus evals espera una redacción
// fechada, y los de cada modo llevan, detrás de todos los demás de ese modo,
// redaccion_sin_fechar, que decide con 0. Con el plan del job, los de
// boe-legislacion, la única con evals con fecha, con su objetivo, son quince,
// de los que deciden once: los trece de antes, con redaccion_sin_fechar de cada
// modo detrás de su cuenta_su_proceso. Los de jurisprudencia y legal-core no
// cambian.
func umbralesDelInforme(e InformeAEscribir, respuestas *respuestasMedidas) []Umbral {
	delJuez := respuestas.juez()
	umbrales := respuestas.umbrales()
	umbrales = append(umbrales, delJuez.umbralesConjuntos()...)

	if delJuez != nil {
		umbrales = append(umbrales, delJuez.umbralesDeLaMedida()...)
	}

	if e.ObjetivoDeDuracion > 0 {
		for _, modo := range e.plan(nil).modos() {
			umbrales = append(umbrales, compararUmbral(Umbral{
				Nombre:      prefijoDelUmbralDeLaDuracion + string(modo),
				Descripcion: fmt.Sprintf(descripcionDelUmbralDeLaDuracion, modo),
				Medida:      float64(e.DuracionDeLosModos[modo]),
				Umbral:      float64(e.ObjetivoDeDuracion),
				Decide:      true,
			}))
		}
	}

	if delJuez != nil {
		umbrales = append(umbrales, delJuez.umbralesDeLaDuracion()...)
	}

	return umbrales
}

// umbrales son, por cada modo del plan y en su orden, los umbrales de las
// respuestas del modelo que decide en ese modo (contracts/informe-del-job.md §2
// de H24; FR-030, FR-031, FR-034; contracts/evals-jurisprudencia.md §4 de H23;
// FR-060, FR-062 de H23), en este orden: el de las que no activaron la skill,
// que decide con 0; con juez, el recuento de cada una de sus clases, que no
// decide (umbralesDeLasClases); con alguna eval que declara sentencias, el de
// las que llevan una cita sin documento cotejado o un ECLI sin origen, que
// decide con 0; y, desde H23.1, con alguna eval que pide el cotejo, el de las
// de esas evals cuya sesión no cotejó ningún documento, que decide con 0
// (contracts/umbral-documento-sin-cotejar.md §1 y §2 de H23.1; FR-001, FR-006
// de H23.1); y, desde H9, con alguna eval de plazo, el de las de esas evals sin
// la línea del plazo con la fecha esperada, y con alguna de cobertura, el de las
// de esas evals sin la declaración de no cubierto con la norma esperada o con
// una línea del plazo, los dos decidiendo con 0 (contracts/evals-legal-core.md
// §3 de H9; FR-053, FR-054 de H9); y, desde H20, con alguna eval que espera una
// redacción fechada, el de las de esas evals cuya sesión no leyó el bloque de la
// eval en su fecha o que no llevan las líneas que la eval espera, que decide con
// 0 (contracts/evals-redaccion-fechada.md §3 de H20; FR-054, FR-055 de H20).
// Todos tienen como total las respuestas del modo, las medidas del modelo en
// las evals que activan la skill, menos documento_sin_cotejar, plazo_sin_fecha,
// cobertura_sin_declarar y redaccion_sin_fechar, que tienen solo las de las
// evals que piden el cotejo, las de plazo, las de cobertura y las que esperan
// una redacción fechada. Ninguno cuenta las respuestas de las evals sin binario
// ni servidor (FR-013 de H24). Ninguno, si la skill no tiene umbrales de sus
// respuestas.
func (r *respuestasMedidas) umbrales() []Umbral {
	var umbrales []Umbral

	for _, grupo := range r.deLosModos() {
		umbrales = append(umbrales, umbralDeSinActivar(r.modelo, grupo))
		umbrales = append(umbrales, r.delJuez.umbralesDeLasClases(grupo)...)

		if r.conSentencias {
			umbrales = append(umbrales, umbralDeCitaSinDocumento(r.modelo, grupo))
		}

		if r.conCotejo {
			umbrales = append(umbrales, umbralDeDocumentoSinCotejar(r.modelo, grupo))
		}

		if r.conPlazo {
			umbrales = append(umbrales, umbralDePlazoSinFecha(r.modelo, grupo))
		}

		if r.conCobertura {
			umbrales = append(umbrales, umbralDeCoberturaSinDeclarar(r.modelo, grupo))
		}

		if r.conRedaccionFechada {
			umbrales = append(umbrales, umbralDeRedaccionSinFechar(r.modelo, grupo))
		}
	}

	return umbrales
}

// umbralesDeLasClases son, por cada clase del juez en el orden de clases.yaml,
// el recuento de las respuestas del grupo de un modo que el juez deja marcadas
// en ella —con sus tres votos, si la clase decide, y con el primero, si solo se
// publica—, con el umbral de la clase (contracts/informe-del-job.md §2 de H24;
// FR-030, FR-031). Su total son las respuestas del grupo: la respuesta sin
// juzgar sigue en él y no está en ninguna medida. Ninguno si la skill no tiene
// juez.
//
// Desde H26 (contracts/umbrales-del-informe.md §3 de H26; research D2 de H26;
// FR-010, FR-011 de H26) ninguno decide: el de una clase que solo se publica,
// como hasta ahora, y el de una que decide, porque lo que decide de ella es su
// umbral conjunto (umbralesConjuntos). Su cumple sigue siendo la comparación
// del contrato, con el mismo umbral que el conjunto, y solo informa.
func (j *juicioDelJuez) umbralesDeLasClases(grupo grupoJuzgado) []Umbral {
	if j == nil {
		return nil
	}

	umbrales := make([]Umbral, 0, len(j.clases))

	for posicion, clase := range j.clases {
		total := len(grupo.respuestas)

		descripcion := descripcionDeLaClaseQueSePublica
		if clase.Decide {
			descripcion = descripcionDeLaClaseQueDecide
		}

		umbrales = append(umbrales, compararUmbral(Umbral{
			Nombre:      nombreDelUmbralDeClase(clase.Nombre, j.modelo, grupo.modo),
			Descripcion: fmt.Sprintf(descripcion, j.modelo, grupo.modo, clase.Nombre),
			Medida:      float64(len(grupo.marcadasEn(posicion))),
			Total:       &total,
			Umbral:      clase.Umbral,
			Decide:      false,
		}))
	}

	return umbrales
}

// umbralesConjuntos son, por cada clase del juez que decide y en el orden de
// clases.yaml, su umbral conjunto, ya comparado
// (contracts/umbrales-del-informe.md §2 de H26; data-model §3 de H26; research
// D1 de H26; FR-001 a FR-003 de H26): las respuestas de todos los modos del
// plan que el juez deja marcadas en ella con sus tres votos, sobre las
// respuestas de todos ellos, con el umbral de la clase, que así es una tasa
// sobre los modos juntos, y decidiendo. Su medida y su total son la suma de los
// de los recuentos por modo de la clase (umbralesDeLasClases), sobre los mismos
// grupos, y con un plan de un solo modo, los de ese modo: no cuenta las
// respuestas de las evals sin binario ni servidor ni las de un modelo
// informativo, y la respuesta sin juzgar sigue en su total y fuera de su
// medida. Una clase que solo se publica no tiene ninguno. Ninguno si la skill
// no tiene juez.
func (j *juicioDelJuez) umbralesConjuntos() []Umbral {
	if j == nil {
		return nil
	}

	var umbrales []Umbral

	for posicion, clase := range j.clases {
		if !clase.Decide {
			continue
		}

		marcadas, total := 0, 0

		for _, grupo := range j.deLosModos() {
			marcadas += len(grupo.marcadasEn(posicion))
			total += len(grupo.respuestas)
		}

		umbrales = append(umbrales, compararUmbral(Umbral{
			Nombre:      nombreDelUmbralConjunto(clase.Nombre, j.modelo),
			Descripcion: fmt.Sprintf(descripcionDelUmbralConjunto, j.modelo, clase.Nombre),
			Medida:      float64(marcadas),
			Total:       &total,
			Umbral:      clase.Umbral,
			Decide:      true,
		}))
	}

	return umbrales
}

// nombreDelUmbralConjunto es el nombre del umbral conjunto de una clase del
// juez que decide, el de las respuestas de un modelo en todos los modos que el
// juez deja marcadas en ella: la clase seguida del modelo, sin modo
// (contracts/umbrales-del-informe.md §2 de H26; research D3 de H26).
func nombreDelUmbralConjunto(clase, modelo string) string {
	return clase + separadorDelModoDelUmbral + modelo
}

// nombreDelUmbralDeClase es el nombre del recuento de las respuestas de un
// modelo en un modo que el juez deja marcadas en una clase: el del umbral
// conjunto de la clase (nombreDelUmbralConjunto) seguido del modo, que es la
// clase seguida del modelo y del modo.
func nombreDelUmbralDeClase(clase, modelo string, modo Modo) string {
	return nombreDelUmbralConjunto(clase, modelo) + separadorDelModoDelUmbral + string(modo)
}

// umbralDeSinActivar es el umbral de las respuestas sin la skill activada de un
// modelo en un modo, sobre sus respuestas medidas, ya comparado: se nombra con
// su prefijo seguido del modelo y del modo, su descripción los nombra, y decide.
func umbralDeSinActivar(modelo string, delModo grupoJuzgado) Umbral {
	total := len(delModo.respuestas)

	return compararUmbral(Umbral{
		Nombre:      prefijoDelUmbralSinActivar + modelo + separadorDelModoDelUmbral + string(delModo.modo),
		Descripcion: fmt.Sprintf(descripcionDelUmbralSinActivar, modelo, delModo.modo),
		Medida:      float64(delModo.sinActivar()),
		Total:       &total,
		Umbral:      umbralDeRespuestasSinActivar,
		Decide:      true,
	})
}

// umbralDeCitaSinDocumento es el umbral de las respuestas de un modelo en un
// modo que llevan una cita de sentencia sin documento cotejado o un ECLI sin
// origen, sobre sus respuestas medidas, ya comparado
// (contracts/evals-jurisprudencia.md §4 de H23; data-model §7 de H23; FR-060 de
// H23): se nombra con su prefijo seguido del modelo y del modo, su descripción
// los nombra, y decide con 0. Es un hecho de la sesión, que se comprueba sin
// modelo (ADR 0037).
func umbralDeCitaSinDocumento(modelo string, delModo grupoJuzgado) Umbral {
	total := len(delModo.respuestas)

	return compararUmbral(Umbral{
		Nombre:      nombreDelUmbralDeCitaSinDocumento(modelo, delModo.modo),
		Descripcion: fmt.Sprintf(descripcionDelUmbralDeCitaSinDocumento, modelo, delModo.modo),
		Medida:      float64(len(delModo.sinDocumento())),
		Total:       &total,
		Umbral:      umbralDeRespuestasSinDocumento,
		Decide:      true,
	})
}

// nombreDelUmbralDeCitaSinDocumento es el nombre del umbral de las respuestas
// de un modelo en un modo con una cita sin documento cotejado o un ECLI sin
// origen: su prefijo seguido del modelo y del modo.
func nombreDelUmbralDeCitaSinDocumento(modelo string, modo Modo) string {
	return prefijoDelUmbralDeCitaSinDocumento + modelo + separadorDelModoDelUmbral + string(modo)
}

// umbralDeDocumentoSinCotejar es el umbral de las respuestas de un modelo en un
// modo de las evals que piden el cotejo cuya sesión no cotejó ningún documento,
// sobre esas respuestas, ya comparado (contracts/umbral-documento-sin-cotejar.md
// §2 de H23.1; data-model §4 de H23.1; FR-001, FR-004 de H23.1): se nombra con
// su prefijo seguido del modelo y del modo, su descripción los nombra, y decide
// con 0. Su medida, las de sinCotejar, y su total, las de queLePidenElCotejo:
// cada sesión cuenta una vez como mucho, y una de una eval que no pide el
// cotejo no está en ninguno. Es un hecho de la sesión, que se comprueba sin
// modelo (ADR 0037).
func umbralDeDocumentoSinCotejar(modelo string, delModo grupoJuzgado) Umbral {
	total := len(delModo.queLePidenElCotejo())

	return compararUmbral(Umbral{
		Nombre:      nombreDelUmbralDeDocumentoSinCotejar(modelo, delModo.modo),
		Descripcion: fmt.Sprintf(descripcionDelUmbralDeDocumentoSinCotejar, modelo, delModo.modo),
		Medida:      float64(len(delModo.sinCotejar())),
		Total:       &total,
		Umbral:      umbralDeRespuestasSinCotejar,
		Decide:      true,
	})
}

// nombreDelUmbralDeDocumentoSinCotejar es el nombre del umbral de las
// respuestas de un modelo en un modo de las evals que piden el cotejo cuya
// sesión no cotejó ningún documento: su prefijo seguido del modelo y del modo.
func nombreDelUmbralDeDocumentoSinCotejar(modelo string, modo Modo) string {
	return prefijoDelUmbralDeDocumentoSinCotejar + modelo + separadorDelModoDelUmbral + string(modo)
}

// umbralDePlazoSinFecha es el umbral de las respuestas de un modelo en un modo de
// las evals de plazo sin la línea del plazo con la fecha esperada, sobre esas
// respuestas, ya comparado (contracts/evals-legal-core.md §3 de H9; data-model
// §3 de H9; FR-053 de H9): se nombra con su prefijo seguido del modelo y del
// modo, su descripción los nombra, y decide con 0. Su medida, las de sinLaFecha,
// y su total, las de deLasEvalsDePlazo: cada sesión cuenta una vez como mucho,
// sus avisos no cuentan (FR-055 de H9), y una de una eval que no es de plazo no
// está en ninguno. Es un hecho de la respuesta, que se comprueba sin modelo
// (ADR 0037).
func umbralDePlazoSinFecha(modelo string, delModo grupoJuzgado) Umbral {
	dePlazo := len(delModo.deLasEvalsDePlazo())

	return compararUmbral(Umbral{
		Nombre:      nombreDelUmbralDePlazoSinFecha(modelo, delModo.modo),
		Descripcion: fmt.Sprintf(descripcionDelUmbralDePlazoSinFecha, modelo, delModo.modo),
		Medida:      float64(len(delModo.sinLaFecha())),
		Total:       &dePlazo,
		Umbral:      umbralDeRespuestasSinFecha,
		Decide:      true,
	})
}

// nombreDelUmbralDePlazoSinFecha es el nombre del umbral de las respuestas de un
// modelo en un modo de las evals de plazo sin la línea del plazo con la fecha
// esperada: su prefijo seguido del modelo y del modo.
func nombreDelUmbralDePlazoSinFecha(modelo string, modo Modo) string {
	return prefijoDelUmbralDePlazoSinFecha + modelo + separadorDelModoDelUmbral + string(modo)
}

// umbralDeCoberturaSinDeclarar es el umbral de las respuestas de un modelo en un
// modo de las evals de cobertura sin la declaración de no cubierto con la norma
// esperada o con una línea del plazo, sobre esas respuestas, ya comparado
// (contracts/evals-legal-core.md §3 de H9; data-model §3 de H9; FR-054 de H9):
// se nombra con su prefijo seguido del modelo y del modo, su descripción los
// nombra, y decide con 0. Su medida, las de sinLaDeclaracion, y su total, las de
// deLasEvalsDeCobertura: cada sesión cuenta una vez como mucho, aunque le falte
// la declaración y le sobre la línea, y una de una eval que no es de cobertura
// no está en ninguno. Es un hecho de la respuesta, que se comprueba sin modelo
// (ADR 0037).
func umbralDeCoberturaSinDeclarar(modelo string, delModo grupoJuzgado) Umbral {
	deCobertura := len(delModo.deLasEvalsDeCobertura())

	return compararUmbral(Umbral{
		Nombre:      nombreDelUmbralDeCoberturaSinDeclarar(modelo, delModo.modo),
		Descripcion: fmt.Sprintf(descripcionDelUmbralDeCoberturaSinDeclarar, modelo, delModo.modo),
		Medida:      float64(len(delModo.sinLaDeclaracion())),
		Total:       &deCobertura,
		Umbral:      umbralDeRespuestasSinDeclarar,
		Decide:      true,
	})
}

// nombreDelUmbralDeCoberturaSinDeclarar es el nombre del umbral de las
// respuestas de un modelo en un modo de las evals de cobertura sin la
// declaración de no cubierto con la norma esperada o con una línea del plazo:
// su prefijo seguido del modelo y del modo.
func nombreDelUmbralDeCoberturaSinDeclarar(modelo string, modo Modo) string {
	return prefijoDelUmbralDeCoberturaSinDeclarar + modelo + separadorDelModoDelUmbral + string(modo)
}

// umbralDeRedaccionSinFechar es el umbral de las respuestas de un modelo en un
// modo de las evals que esperan una redacción fechada cuya sesión no leyó el
// bloque de la eval en su fecha o que no llevan las líneas que la eval espera,
// sobre esas respuestas, ya comparado (contracts/evals-redaccion-fechada.md §3
// de H20; data-model §7.4 de H20; FR-054, FR-055 de H20): se nombra con su
// prefijo seguido del modelo y del modo, su descripción los nombra, y decide con
// 0. Su medida, las de sinFechar, y su total, las de deLasEvalsConFecha: cada
// sesión cuenta una vez como mucho, aunque no lea en su fecha y le falten las
// líneas, y una de una eval que no espera una redacción fechada no está en
// ninguno. Es un hecho de la sesión y de la respuesta, que se comprueba sin
// modelo (ADR 0037).
func umbralDeRedaccionSinFechar(modelo string, delModo grupoJuzgado) Umbral {
	conFecha := len(delModo.deLasEvalsConFecha())

	return compararUmbral(Umbral{
		Nombre:      nombreDelUmbralDeRedaccionSinFechar(modelo, delModo.modo),
		Descripcion: fmt.Sprintf(descripcionDelUmbralDeRedaccionSinFechar, modelo, delModo.modo),
		Medida:      float64(len(delModo.sinFechar())),
		Total:       &conFecha,
		Umbral:      umbralDeRedaccionesSinFechar,
		Decide:      true,
	})
}

// nombreDelUmbralDeRedaccionSinFechar es el nombre del umbral de las respuestas
// de un modelo en un modo de las evals que esperan una redacción fechada cuya
// sesión no leyó el bloque en su fecha o que no llevan las líneas esperadas: su
// prefijo seguido del modelo y del modo.
func nombreDelUmbralDeRedaccionSinFechar(modelo string, modo Modo) string {
	return prefijoDelUmbralDeRedaccionSinFechar + modelo + separadorDelModoDelUmbral + string(modo)
}

// umbralesDeLaMedida son, por cada clase del juez que decide, los dos umbrales
// de su medida versionada (contracts/informe-del-job.md §2 de H24; FR-032,
// FR-044): los casos etiquetados como defecto que no quedaron marcados, sobre
// los etiquetados como defecto, y los etiquetados como correctos que sí, sobre
// los etiquetados como correctos. Los dos deciden con 0. Los recuentos y los
// totales son los de la medida leída: ningún caso se vota aquí.
func (j *juicioDelJuez) umbralesDeLaMedida() []Umbral {
	var umbrales []Umbral

	for _, clase := range j.clases {
		if !clase.Decide {
			continue
		}

		defectos, correctos := j.medida.Defectos.Casos, j.medida.Correctos.Casos
		deLaClase := prefijoDelUmbralDeLaMedidaDelJuez + clase.Nombre

		umbrales = append(umbrales,
			compararUmbral(Umbral{
				Nombre:      deLaClase + sufijoDeLosDefectosSinMarcar,
				Descripcion: fmt.Sprintf(descripcionDeLosDefectosSinMarcar, clase.Nombre),
				Medida:      float64(j.medida.Defectos.SinMarcar),
				Total:       &defectos,
				Umbral:      umbralDeLaMedidaDelJuez,
				Decide:      true,
			}),
			compararUmbral(Umbral{
				Nombre:      deLaClase + sufijoDeLosCorrectosMarcados,
				Descripcion: fmt.Sprintf(descripcionDeLosCorrectosMarcados, clase.Nombre),
				Medida:      float64(j.medida.Correctos.Marcados),
				Total:       &correctos,
				Umbral:      umbralDeLaMedidaDelJuez,
				Decide:      true,
			}))
	}

	return umbrales
}

// umbralesDeLaDuracion son, por cada modo del plan y en su orden, el umbral de
// los segundos de los votos de sus respuestas, del principio del primero al
// final del último (contracts/informe-del-job.md §2 de H24; FR-033): sin total,
// con 900 s y decidiendo. Los votos de las respuestas de las evals sin binario
// ni servidor no entran en ninguno (FR-013).
func (j *juicioDelJuez) umbralesDeLaDuracion() []Umbral {
	var umbrales []Umbral

	for _, grupo := range j.deLosModos() {
		umbrales = append(umbrales, compararUmbral(Umbral{
			Nombre:      prefijoDelUmbralDeLaDuracionDelJuez + string(grupo.modo),
			Descripcion: fmt.Sprintf(descripcionDelUmbralDeLaDuracionDelJuez, grupo.modo),
			Medida:      float64(grupo.segundos),
			Umbral:      segundosDelJuezPorModo,
			Decide:      true,
		}))
	}

	return umbrales
}

// esDeLaDuracion dice si el umbral es el de la duración de las sesiones de un
// modo o el de la de los votos del juez sobre sus respuestas: su motivo no es
// de la skill sino de la ejecución.
func (u Umbral) esDeLaDuracion() bool {
	return strings.HasPrefix(u.Nombre, prefijoDelUmbralDeLaDuracion) ||
		strings.HasPrefix(u.Nombre, prefijoDelUmbralDeLaDuracionDelJuez)
}

// compararUmbral pone al umbral su comparación, «<=», y su resultado: su valor
// comparado con Umbral en float64, sin redondear nada (contrato de umbrales del
// ADR 0029).
func compararUmbral(umbral Umbral) Umbral {
	umbral.Comparacion = comparacionMenorOIgual
	umbral.Cumple = umbral.valor() <= umbral.Umbral

	return umbral
}

// valor es lo que el umbral compara: la proporción Medida/Total —0 si Total es
// 0— si lleva total y, si no, la propia Medida.
func (u Umbral) valor() float64 {
	switch {
	case u.Total == nil:
		return u.Medida
	case *u.Total == 0:
		return 0
	default:
		return u.Medida / float64(*u.Total)
	}
}

// motivosDeLosUmbrales son los motivos de la raíz de los umbrales que deciden y no
// se cumplen, en su orden, sin los de la duración, que son de la ejecución
// (motivosDeLaDuracionDeLasSesiones): «umbral <nombre>: <medida>, y tiene que ser
// <condición>», con la medida y la condición como en informe.md (contrato
// informe-del-job §2.1 de H7.3; FR-003). El nombre lleva el modo del umbral, si
// es de uno, y con él lo nombra el motivo (contracts/evals-en-dos-modos.md §5.2
// de H21; FR-044).
//
// El del umbral conjunto de una clase del juez, cuyo nombre no lleva modo,
// nombra además, detrás de dos puntos, cada respuesta que el juez dejó marcada
// en ella en cualquiera de los modos, por su sesión, con su modo entre
// paréntesis y con sus tres frases (marcadasDelUmbral;
// contracts/informe-del-job.md §4 de H24; FR-035 de H24;
// contracts/umbrales-del-informe.md §4 de H26; FR-004 de H26). El recuento de
// la clase en un modo no decide, y no da ninguno. El de cita_sin_documento
// nombra cada respuesta que cuenta, por su sesión y con sus ECLI
// (sinDocumentoDelUmbral; contracts/evals-jurisprudencia.md §4 de H23; research
// D21 de H23). Y el de documento_sin_cotejar, cada respuesta que cuenta, por su
// sesión y sin nada más (nombradasPorSuSesion;
// contracts/umbral-documento-sin-cotejar.md §3 de H23.1; FR-003 de H23.1), como,
// desde H9, los de plazo_sin_fecha y cobertura_sin_declarar
// (contracts/evals-legal-core.md §3 de H9; FR-055 de H9) y, desde H20, el de
// redaccion_sin_fechar (contracts/evals-redaccion-fechada.md §3 de H20; FR-055
// de H20). respuestas es lo que
// se mide de las respuestas de la ejecución, nil si la skill no tiene umbrales
// de ellas.
func motivosDeLosUmbrales(umbrales []Umbral, respuestas *respuestasMedidas) []string {
	var motivos []string

	for _, umbral := range umbrales {
		if umbral.esDeLaDuracion() || !umbral.Decide || umbral.Cumple {
			continue
		}

		motivo := fmt.Sprintf(motivoDeUnUmbral, umbral.Nombre, umbral.medidaEscrita(), umbral.condicionEscrita())
		if nombradas := respuestas.nombradasEn(umbral.Nombre); nombradas != "" {
			motivo += ": " + nombradas
		}

		motivos = append(motivos, motivo)
	}

	return motivos
}

// nombradasEn nombra las respuestas que cuentan en el umbral con ese nombre,
// como las lleva su motivo: las que el juez dejó marcadas en todos los modos,
// si es el conjunto de una de sus clases (marcadasDelUmbral); las que llevan
// una cita sin documento cotejado o un ECLI sin origen, si es el de
// cita_sin_documento de un modo (sinDocumentoDelUmbral); o, si es el de
// documento_sin_cotejar, el de plazo_sin_fecha, el de cobertura_sin_declarar o
// el de redaccion_sin_fechar de un modo, las que cuentan en él
// (nombradasPorSuSesion). Vacío con cualquier otro umbral, y si la skill no
// tiene umbrales de sus respuestas.
func (r *respuestasMedidas) nombradasEn(nombre string) string {
	if marcadas := r.juez().marcadasDelUmbral(nombre); marcadas != "" {
		return marcadas
	}

	if sinDocumento := r.sinDocumentoDelUmbral(nombre); sinDocumento != "" {
		return sinDocumento
	}

	return r.nombradasPorSuSesion(nombre)
}

// nombradasPorSuSesion nombra las respuestas que cuentan en el umbral con ese
// nombre, como las lleva su motivo, si es uno de los de un modo que las nombra
// por su sesión y nada más: documento_sin_cotejar, las de una sesión que no
// cotejó ningún documento (contracts/umbral-documento-sin-cotejar.md §3 de
// H23.1; research D7 de H23.1; FR-003 de H23.1); y, desde H9, plazo_sin_fecha,
// las de una eval de plazo sin la línea del plazo con la fecha esperada, y
// cobertura_sin_declarar, las de una eval de cobertura sin la declaración de no
// cubierto con la norma esperada o con una línea del plazo
// (contracts/evals-legal-core.md §3 de H9; FR-055 de H9); y, desde H20,
// redaccion_sin_fechar, las de una eval que espera una redacción fechada cuya
// sesión no leyó el bloque en su fecha o que no llevan las líneas esperadas
// (contracts/evals-redaccion-fechada.md §3 de H20; FR-055 de H20). Cada una por
// su sesión, sin el modo, que ya va en el nombre del umbral, separadas por «; »
// y en orden de sesión. Vacío si el umbral no es ninguno de ellos o si ninguna
// respuesta cuenta.
func (r *respuestasMedidas) nombradasPorSuSesion(nombre string) string {
	for _, grupo := range r.deLosModos() {
		var cuentan []respuestaDelJuez

		switch nombre {
		case nombreDelUmbralDeDocumentoSinCotejar(r.modelo, grupo.modo):
			cuentan = grupo.sinCotejar()
		case nombreDelUmbralDePlazoSinFecha(r.modelo, grupo.modo):
			cuentan = grupo.sinLaFecha()
		case nombreDelUmbralDeCoberturaSinDeclarar(r.modelo, grupo.modo):
			cuentan = grupo.sinLaDeclaracion()
		case nombreDelUmbralDeRedaccionSinFechar(r.modelo, grupo.modo):
			cuentan = grupo.sinFechar()
		default:
			continue
		}

		nombradas := make([]string, 0, len(cuentan))

		for _, respuesta := range cuentan {
			nombradas = append(nombradas, respuesta.sesion)
		}

		return strings.Join(nombradas, separadorDeLasMarcadas)
	}

	return ""
}

// sinDocumentoDelUmbral nombra las respuestas que cuentan en el umbral
// cita_sin_documento con ese nombre, como las lleva su motivo
// (contracts/evals-jurisprudencia.md §4 de H23; research D21 de H23): de cada
// una, «<sesión>: <ECLI> (<condición>)», con cada ECLI que la hace contar y
// «cita sin documento cotejado» o «sin origen» entre paréntesis, separados por
// « · », y unas de otras separadas por «; », en orden de sesión. Vacío si el
// umbral no es el de cita_sin_documento de un modo o si ninguna respuesta
// cuenta.
func (r *respuestasMedidas) sinDocumentoDelUmbral(nombre string) string {
	for _, grupo := range r.deLosModos() {
		if nombreDelUmbralDeCitaSinDocumento(r.modelo, grupo.modo) != nombre {
			continue
		}

		cuentan := grupo.sinDocumento()
		nombradas := make([]string, 0, len(cuentan))

		for _, respuesta := range cuentan {
			eclis := make([]string, 0, len(respuesta.sinDocumento))
			for _, ecli := range respuesta.sinDocumento {
				eclis = append(eclis, fmt.Sprintf(ecliConSuCondicion, ecli.ECLI, ecli.Condicion))
			}

			nombradas = append(nombradas, respuesta.sesion+": "+strings.Join(eclis, separadorDeLasFrases))
		}

		return strings.Join(nombradas, separadorDeLasMarcadas)
	}

	return ""
}

// marcadasDelUmbral nombra las respuestas que el juez dejó marcadas en la clase
// del umbral conjunto con ese nombre, como las lleva su motivo
// (contracts/informe-del-job.md §4 de H24; contracts/umbrales-del-informe.md §4
// de H26; research D4 de H26; FR-004 de H26): las de todos los modos, las del
// modo orden delante y, dentro de un modo, en orden de sesión; de cada una,
// «<sesión> (modo <modo>): «<frase 1>» · «<frase 2>» · «<frase 3>»», con las
// frases de los votos que la marcan, y unas de otras separadas por «; ». El
// modo va escrito porque el nombre de la sesión solo lo deja ver en el modo
// herramienta. Vacío si el umbral no es el conjunto de una clase del juez
// —tampoco con el recuento de una clase en un modo, que no decide ni da
// motivo—, si ninguna respuesta quedó marcada o si la skill no tiene juez.
func (j *juicioDelJuez) marcadasDelUmbral(nombre string) string {
	if j == nil {
		return ""
	}

	for posicion, clase := range j.clases {
		if nombreDelUmbralConjunto(clase.Nombre, j.modelo) != nombre {
			continue
		}

		var nombradas []string

		for _, grupo := range j.deLosModos() {
			for _, respuesta := range grupo.marcadasEn(posicion) {
				frases := respuesta.juicio.Clases[posicion].frasesQueCuentan()
				for deLaFrase, frase := range frases {
					frases[deLaFrase] = comillaQueAbre + frase + comillaQueCierra
				}

				nombradas = append(nombradas, fmt.Sprintf(sesionConSuModo, respuesta.sesion, grupo.modo)+": "+
					strings.Join(frases, separadorDeLasFrases))
			}
		}

		return strings.Join(nombradas, separadorDeLasMarcadas)
	}

	return ""
}

// motivosDeLaDuracionDeLasSesiones son los motivos de la raíz de los umbrales de
// la duración que deciden y no se cumplen, en su orden, uno por modo, con el
// prefijo de la ejecución: no son de la skill (contrato informe-del-job §2.3 de
// H7.3; contracts/evals-en-dos-modos.md §5.2 de H21; FR-051). Ninguno si no los
// hay. Desde H24 van con ellos, detrás, los de la duración de los votos del juez
// de cada modo (contracts/informe-del-job.md §4 de H24; FR-033).
func motivosDeLaDuracionDeLasSesiones(umbrales []Umbral) []string {
	var motivos []string

	for _, umbral := range umbrales {
		if umbral.esDeLaDuracion() && umbral.Decide && !umbral.Cumple {
			motivos = append(motivos, motivoDeLaEjecucion+fmt.Sprintf(motivoDeLaDuracion, umbral.Nombre,
				umbral.medidaEscrita(), umbral.condicionEscrita()))
		}
	}

	return motivos
}

// medidaEscrita es la medida del umbral como la escriben informe.md y los
// motivos: con total, «<medida> de <total> (<p> %)», con el porcentaje con un
// decimal y coma, 0 si total es 0; sin él, la medida sola.
func (u Umbral) medidaEscrita() string {
	medida := strconv.FormatFloat(u.Medida, 'f', -1, 64)
	if u.Total == nil {
		return medida
	}

	return medida + " de " + strconv.Itoa(*u.Total) + " (" + porcentajeConComa(u.valor()) + " %)"
}

// condicionEscrita es la condición del umbral como la escriben informe.md y los
// motivos: «≤ <u> %», con un decimal y coma, si compara una proporción, y
// «≤ <u>» si no.
func (u Umbral) condicionEscrita() string {
	if u.Total == nil {
		return signoMenorOIgual + strconv.FormatFloat(u.Umbral, 'f', -1, 64)
	}

	return signoMenorOIgual + porcentajeConComa(u.Umbral) + " %"
}

// porcentajeConComa es la proporción como porcentaje con un decimal y coma
// decimal, como el informe final: 3/51 es «5,9».
func porcentajeConComa(cociente float64) string {
	return strings.Replace(strconv.FormatFloat(cociente*100, 'f', 1, 64), ".", ",", 1)
}

// filasDeUmbrales son las filas de la tabla de los umbrales de informe.md
// (contrato informe-del-job §4 de H7.3): el nombre, como código, como en las
// filas del contrato; la medida, la condición, si se cumple y si hace fallar el
// veredicto, «no: solo se publica» en los que no deciden.
func filasDeUmbrales(umbrales []Umbral) [][]string {
	filas := make([][]string, 0, len(umbrales))

	for _, umbral := range umbrales {
		decide := siONo(umbral.Decide)
		if !umbral.Decide {
			decide = soloSePublica
		}

		filas = append(filas, []string{
			"`" + umbral.Nombre + "`", umbral.medidaEscrita(), umbral.condicionEscrita(), siONo(umbral.Cumple), decide,
		})
	}

	return filas
}
