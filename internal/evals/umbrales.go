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
// contracts/evals-en-dos-modos.md §5.1 de H21; data-model §10 de H21). Cada uno
// se mide sobre las sesiones de un solo modo: el de las respuestas de un modelo
// se nombra con lo que mide seguido del modelo y del modo, y el de la duración,
// con el modo; su descripción los nombra, primero el modelo y después el modo; y
// la única comparación es «<=», que informe.md escribe «≤».
//
// Desde H24, lo que mide un umbral de las respuestas puede ser una clase del
// juez, y entonces se nombra con ella; los dos de la medida del juez no son de
// ningún modo, y se nombran con su prefijo, la clase que decide y lo que
// cuentan; y los votos del juez tienen su duración, aparte de la de las
// sesiones.
const (
	prefijoDelUmbralSinActivar          = "sin_activar:"
	prefijoDelUmbralDeLaMedidaDelJuez   = "medida_del_juez:"
	prefijoDelUmbralDeLaDuracion        = "duracion_de_las_sesiones:"
	prefijoDelUmbralDeLaDuracionDelJuez = "duracion_del_juez:"
	separadorDelModoDelUmbral           = ":"

	sufijoDeLosDefectosSinMarcar = ":defectos_sin_marcar"
	sufijoDeLosCorrectosMarcados = ":correctos_marcados"

	descripcionDelUmbralSinActivar = "Respuestas de %s en el modo %s sin la skill activada, sobre sus respuestas " +
		"medidas en las evals que la activan"
	descripcionDeLaClaseQueDecide = "Respuestas de %s en el modo %s que el juez marca en %s con sus tres votos, " +
		"sobre sus respuestas juzgadas en las evals que activan la skill"
	descripcionDeLaClaseQueSePublica = "Respuestas de %s en el modo %s con sí en %s en el primer voto del juez, " +
		"sobre sus respuestas juzgadas en las evals que activan la skill"
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
// que no se cumple; y, en el de una clase del juez, cómo nombra a cada respuesta
// marcada —su sesión y, entre comillas, las frases de los votos que la marcan—
// y qué separa una de otra.
const (
	soloSePublica      = "no: solo se publica"
	motivoDeUnUmbral   = "umbral %s: %s, y tiene que ser %s"
	motivoDeLaDuracion = "%s: %s s, y tiene que ser %s s"

	separadorDeLasMarcadas = "; "
	separadorDeLasFrases   = " · "
	comillaQueAbre         = "«"
	comillaQueCierra       = "»"
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
	// (contracts/informe-del-job.md §2 de H24).
	Nombre string `json:"nombre"`

	// Descripcion dice en una línea qué se mide, en qué modo y sobre qué
	// respuestas.
	Descripcion string `json:"descripcion"`

	// Medida es lo medido: en las sesiones de su modo, las respuestas sin la
	// skill activada, las que el juez marca en una clase o los segundos de su
	// tanda o de sus votos; y, en la medida versionada del juez, los casos que
	// no dan lo que dice su etiqueta.
	Medida float64 `json:"medida"`

	// Total, solo en los de las respuestas y en los de la medida del juez, es
	// aquello sobre lo que se mide: las respuestas medidas del modelo en ese
	// modo en las evals que activan la skill, o los casos con esa etiqueta. Con
	// él, lo que se compara es la proporción Medida/Total, 0 si Total es 0. Sin
	// él, la propia Medida.
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
// solo existen si la skill tiene juez (research D13 de H24), y, detrás, los de
// la duración de las sesiones si el job da un objetivo mayor que 0, sin total y
// decidiendo. Nil, [] en informe.json, si no hay ninguno.
//
// Desde H21 (contracts/evals-en-dos-modos.md §5.1 de H21; data-model §10;
// research.md D19; FR-043 a FR-045), cada umbral es de un modo del plan y lo
// nombra: los de las respuestas van por modo, los del modo orden delante, y
// detrás de ellos, el de la duración de cada modo, en su orden, que mide los
// segundos de la tanda de ese modo frente al mismo objetivo. Ninguno suma las
// medidas de dos modos ni cuenta las sesiones o la tanda de las evals sin
// binario ni servidor.
//
// Desde H24 (contracts/informe-del-job.md §2 y §6 de H24; FR-030 a FR-034),
// ninguno mide las expresiones de la lista —salen
// expresiones_prohibidas:<modelo>:<modo>, de todos los modelos, y
// redaccion_no_leida:<modelo>:<modo>— y los de una skill con juez son, en este
// orden:
//
//   - por cada modo, el de las respuestas sin la skill activada y el de cada
//     clase del juez, sobre las mismas respuestas (umbralesDeLasRespuestas);
//   - los dos de la medida versionada del juez de cada clase que decide
//     (umbralesDeLaMedida);
//   - los de la duración de las sesiones, como hasta ahora;
//   - y, por cada modo, el de la duración de sus votos (umbralesDeLaDuracion).
//
// Con el plan del job, los dos modos, y el juez y el objetivo de
// boe-legislacion son doce, y diez deciden: todos menos el de la clase que solo
// se publica en cada modo.
func umbralesDelInforme(e InformeAEscribir, delJuez *juicioDelJuez) []Umbral {
	var umbrales []Umbral

	if delJuez != nil {
		umbrales = append(umbrales, delJuez.umbralesDeLasRespuestas()...)
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

// umbralesDeLasRespuestas son, por cada modo del plan y en su orden, los
// umbrales de las respuestas del modelo que decide en ese modo
// (contracts/informe-del-job.md §2 de H24; FR-030, FR-031, FR-034): el de las
// que no activaron la skill, que decide con 0, y, por cada clase del juez en el
// orden de clases.yaml, el de las que el juez deja marcadas en ella —con sus
// tres votos, si la clase decide, y con el primero, si solo se publica—, con el
// umbral y el decide de la clase. Todos tienen como total las respuestas del
// modo, las medidas del modelo en las evals que activan la skill: la respuesta
// sin juzgar sigue en él y no está en ninguna medida. Ninguno cuenta las
// respuestas de las evals sin binario ni servidor (FR-013).
func (j *juicioDelJuez) umbralesDeLasRespuestas() []Umbral {
	var umbrales []Umbral

	for _, grupo := range j.deLosModos() {
		umbrales = append(umbrales, umbralDeSinActivar(j.modelo, grupo))

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
				Decide:      clase.Decide,
			}))
		}
	}

	return umbrales
}

// nombreDelUmbralDeClase es el nombre del umbral de las respuestas de un modelo
// en un modo que el juez deja marcadas en una clase: la clase seguida del
// modelo y del modo.
func nombreDelUmbralDeClase(clase, modelo string, modo Modo) string {
	return clase + separadorDelModoDelUmbral + modelo + separadorDelModoDelUmbral + string(modo)
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
// informe-del-job §2.1 de H7.3; FR-003). El nombre lleva el modo del umbral, y
// con él lo nombra el motivo (contracts/evals-en-dos-modos.md §5.2 de H21;
// FR-044).
//
// El de una clase del juez nombra además, detrás de dos puntos, cada respuesta
// que el juez dejó marcada en ella, por su sesión y con sus tres frases
// (marcadasDelUmbral; contracts/informe-del-job.md §4 de H24; FR-035). delJuez
// es el juicio del juez de la ejecución, nil si la skill no lo tiene.
func motivosDeLosUmbrales(umbrales []Umbral, delJuez *juicioDelJuez) []string {
	var motivos []string

	for _, umbral := range umbrales {
		if umbral.esDeLaDuracion() || !umbral.Decide || umbral.Cumple {
			continue
		}

		motivo := fmt.Sprintf(motivoDeUnUmbral, umbral.Nombre, umbral.medidaEscrita(), umbral.condicionEscrita())
		if marcadas := delJuez.marcadasDelUmbral(umbral.Nombre); marcadas != "" {
			motivo += ": " + marcadas
		}

		motivos = append(motivos, motivo)
	}

	return motivos
}

// marcadasDelUmbral nombra las respuestas que el juez dejó marcadas en la clase
// y el modo del umbral con ese nombre, como las lleva su motivo
// (contracts/informe-del-job.md §4 de H24): de cada una, «<sesión>: «<frase 1>»
// · «<frase 2>» · «<frase 3>»», con las frases de los votos que la marcan, y
// unas de otras separadas por «; », en orden de sesión. Vacío si el umbral no
// es el de una clase del juez en un modo, si ninguna respuesta quedó marcada o
// si la skill no tiene juez.
func (j *juicioDelJuez) marcadasDelUmbral(nombre string) string {
	if j == nil {
		return ""
	}

	for _, grupo := range j.deLosModos() {
		for posicion, clase := range j.clases {
			if nombreDelUmbralDeClase(clase.Nombre, j.modelo, grupo.modo) != nombre {
				continue
			}

			marcadas := grupo.marcadasEn(posicion)
			nombradas := make([]string, 0, len(marcadas))

			for _, respuesta := range marcadas {
				frases := respuesta.juicio.Clases[posicion].frasesQueCuentan()
				for deLaFrase, frase := range frases {
					frases[deLaFrase] = comillaQueAbre + frase + comillaQueCierra
				}

				nombradas = append(nombradas, respuesta.sesion+": "+strings.Join(frases, separadorDeLasFrases))
			}

			return strings.Join(nombradas, separadorDeLasMarcadas)
		}
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
