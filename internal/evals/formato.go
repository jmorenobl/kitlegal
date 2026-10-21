package evals

import (
	"fmt"
	"slices"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/jmorenobl/kitlegal/internal/skills"
)

// rutaDelEsquemaDeEval es el esquema publicado del formato común de eval,
// relativo al directorio de este paquete, que es donde go test ejecuta los
// tests desde los que se usa (research.md D2 y V46).
const rutaDelEsquemaDeEval = "../../schemas/eval.yaml.json"

// Eval es un fichero de evals/<skill>/ leído y validado contra el formato común
// (data-model §6; contrato evals-y-grabaciones §1).
type Eval struct {
	// Fichero es el nombre con el que se leyó el fichero. No es una clave del
	// YAML: lo pone LeerEval, y con él se nombra la eval.
	Fichero string `yaml:"-"`

	// Pregunta es lo que se pregunta en la sesión.
	Pregunta string `yaml:"pregunta"`

	// Activa dice si la skill debe activarse con la pregunta.
	Activa bool `yaml:"activa"`

	// NoSeActivan son los nombres de las skills que la sesión no puede activar,
	// en el orden del fichero y sin repetir; vacío si la eval no los declara. Los
	// admite cualquier eval, también una de no activación: las de legal-core
	// declaran boe-legislacion, que sus sesiones también ven instalada
	// (contracts/evals-y-juicio.md §1 y §5 de H7.4; FR-003, FR-004).
	NoSeActivan []string `yaml:"no_se_activan"`

	// Reproduce es el nombre de la skill cuya consulta reproduce la eval, o
	// vacío si no reproduce ninguna (FR-064).
	Reproduce string `yaml:"reproduce"`

	// Informativa dice que la eval se ejecuta y se publica, pero no decide el
	// veredicto: mide algo que la skill todavía no puede hacer con las
	// herramientas que hay (ADR 0016), o algo que todavía no hay datos para
	// exigir (H5.1). Solo la ejecuta el modelo que decide.
	Informativa bool `yaml:"informativa"`

	// PorMateria dice que la pregunta no identifica lo que pregunta —ni el precepto
	// ni la sentencia— y lo busca por su materia: la clase de pregunta con la que
	// la serie de cierres reparte las respuestas marcadas por el juez (ADR 0038,
	// punto 6). Falso por omisión, que es el caso normal: la pregunta nombra el
	// precepto o la sentencia. Solo lo admite una eval que activa la skill y que
	// no es sin binario ni servidor. No es lo mismo que Informativa: una eval
	// informativa puede nombrar el precepto.
	PorMateria bool `yaml:"por_materia"`

	// GrafoPrevio es lo que el grafo de la sesión tiene que haber registrado
	// antes de que empiece: las grabaciones con las que se prepara y los comandos
	// de bloque que lo llenan. Vacío si la eval no lo lleva; solo lo admite una
	// eval que activa la skill (contrato evals-y-skill §1 y §3 de H7; FR-085).
	GrafoPrevio GrafoPrevio `yaml:"grafo_previo"`

	// Comandos son los comandos esperados de una eval que activa la skill, en el
	// orden del fichero; vacío en una de no activación (FR-061).
	Comandos []ComandoEsperado `yaml:"comandos"`

	// Prohibidos son los comandos que ninguna invocación de la sesión puede
	// ejecutar, en el orden del fichero; vacío si la eval no los tiene. Solo los
	// admite una eval que activa la skill (contrato evals-y-skill §1 y §2 de H7;
	// FR-085).
	Prohibidos []ComandoProhibido `yaml:"prohibidos"`

	// Citas son las citas esperadas de una eval que activa la skill, en el orden
	// del fichero; vacío en una de no activación (FR-061).
	Citas []CitaEsperada `yaml:"citas"`

	// Avisos son los códigos de aviso de vigencia cuya forma fija tiene que llevar la respuesta, en el orden del
	// fichero; vacío si la eval no los espera. Solo los admite una eval que activa la skill, y sus valores son los de
	// boe.CodigosDeAviso (FR-020 a FR-022 de H5.1).
	Avisos []string `yaml:"avisos"`

	// Hallazgos son las clases de hallazgo de graph check cuya forma fija tiene que llevar la respuesta, en el orden
	// del fichero; vacío si la eval no los espera. Solo los admite una eval que activa la skill, como los avisos, y sus
	// valores son las clases de grafo.EtiquetasDeHallazgo (contrato evals-y-skill §1 de H7.1; FR-054).
	Hallazgos []string `yaml:"hallazgos"`

	// RedaccionesModificadas son las redacciones modificadas que la respuesta
	// tiene que trasladar, cada una en su línea ⚠ REDACCIÓN MODIFICADA:, en el
	// orden del fichero; vacío si la eval no las espera. Solo las admite una eval
	// que activa la skill, como los hallazgos (contracts/evals-y-juicio.md §1 de
	// H7.4; FR-053).
	RedaccionesModificadas []RedaccionEsperada `yaml:"redacciones_modificadas"`

	// Territorio es lo que la respuesta tiene que declarar del territorio del
	// municipio, cada elemento por su forma fija (ExtraerTerritorio); vacío si la
	// eval no lo espera. Solo lo admite una eval que activa la skill, y con él
	// puede no llevar citas (contrato de evals §1.2 y §1.3 de H6).
	Territorio TerritorioEsperado `yaml:"territorio"`

	// SinBinarioNiServidor dice que la sesión de la eval tiene la skill y nada
	// más: ni kitlegal en el PATH ni el servidor declarado, así que la respuesta
	// no puede consultar nada. Solo lo admite una eval que activa la skill, sin
	// comandos, prohibidos, grafo previo, citas, avisos, hallazgos, redacciones
	// modificadas, territorio, informativa ni reproduce, ni, desde H9, plazo ni
	// no_cubierto: decide siempre y se mide
	// una vez por modelo, fuera de los dos modos (data-model §7 y
	// contracts/evals-en-dos-modos.md §1 de H21; FR-046).
	SinBinarioNiServidor bool `yaml:"sin_binario_ni_servidor"`

	// Sentencias es lo que la respuesta tiene que llevar, y lo que no, de las
	// sentencias de las que habla, cada cosa por su forma; vacío si la eval no lo
	// declara. Solo lo admite una eval que activa la skill y que no es sin
	// binario ni servidor, y con él puede no llevar comandos, citas ni territorio
	// (contracts/evals-jurisprudencia.md §1 de H23; FR-052).
	Sentencias SentenciasEsperadas `yaml:"sentencias"`

	// Plazo es el plazo que la respuesta tiene que dar en su forma fija, con los
	// avisos de plazos que tiene que trasladar; vacío si la eval no lo espera. La
	// eval que lo lleva es una eval de plazo. Solo lo admite una eval que activa la
	// skill y que no es sin binario ni servidor, nunca con NoCubierto, y con él
	// puede no llevar citas (contracts/evals-legal-core.md §1 de H9; FR-051).
	Plazo PlazoEsperado `yaml:"plazo"`

	// NoCubierto es la declaración de no cubierto que la respuesta tiene que
	// llevar en su forma fija; vacía si la eval no la espera. La eval que la lleva
	// es una eval de cobertura. La admite la misma eval que Plazo, nunca con él, y
	// con ella puede no llevar comandos ni citas (contracts/evals-legal-core.md §1
	// de H9; FR-051).
	NoCubierto NoCubiertoEsperado `yaml:"no_cubierto"`

	// PlazoSinCalcular es la declaración de que plazos no calcula el plazo, porque
	// alguno de sus preceptos no regía así el día de inicio, que la respuesta
	// tiene que llevar en su forma fija; vacía si la eval no la espera. La eval
	// que la lleva es una eval de vigencia. La admite la misma eval que Plazo,
	// nunca con él ni con NoCubierto, y con ella puede no llevar citas (ADR 0042).
	PlazoSinCalcular PlazoSinCalcularEsperado `yaml:"plazo_sin_calcular"`

	// RedaccionFechada es lo que la respuesta tiene que decir, en las líneas de
	// ExtraerRedaccionesFechadas, de la redacción de un bloque en una fecha; nil
	// si la eval no lo espera. Solo lo admite una eval que activa la skill y que
	// no es sin binario ni servidor (contracts/evals-redaccion-fechada.md §1.2 de
	// H20; data-model §7.2; FR-052).
	RedaccionFechada *RedaccionFechadaEsperada `yaml:"redaccion_fechada"`

	// Lecturas son los bloques que la caché de las sesiones de la skill sirve
	// porque la eval los provoca, cada uno con su motivo, en el orden del
	// fichero; vacío si la eval no los declara. No son comandos ni citas, no dan
	// motivo ni cuentan en ningún umbral, y ninguno puede ser un comando de bloque
	// ni una cita de la eval (ADR 0043, clase (b); contracts/formato-de-eval.md §1
	// y §2 de H26.1; FR-006 a FR-008).
	Lecturas []LecturaDeLaEval `yaml:"lecturas"`

	// SinServir son las lecturas que la caché de las sesiones de esta eval no
	// sirve con ninguna fecha, en el orden del fichero; vacío si la eval no las
	// declara. Un bloque suyo puede ser un comando de bloque de la eval, que es la
	// lectura que la eval quiere hacer fallar, pero no una de sus citas
	// (contracts/formato-de-eval.md §1 y §2 de H26.1; FR-012 a FR-014, FR-016).
	SinServir []LecturaSinServir `yaml:"sin_servir"`
}

// esDePlazo dice si la eval es una eval de plazo: si espera un plazo, que su
// esquema no admite sin fecha.
func (e Eval) esDePlazo() bool {
	return e.Plazo.Fecha != ""
}

// esDeCobertura dice si la eval es una eval de cobertura: si espera una
// declaración de no cubierto, que su esquema no admite sin norma.
func (e Eval) esDeCobertura() bool {
	return e.NoCubierto.Norma != ""
}

// esDeVigencia dice si la eval es una eval de vigencia: si espera la
// declaración de un plazo sin calcular, que su esquema no admite sin norma
// (ADR 0042).
func (e Eval) esDeVigencia() bool {
	return e.PlazoSinCalcular.Norma != ""
}

// PlazoEsperado es lo que una eval de plazo espera de la respuesta
// (contracts/evals-legal-core.md §1 de H9; data-model §3 de H9).
type PlazoEsperado struct {
	// Fecha es el último día del plazo, AAAA-MM-DD, calculado a mano: la que
	// tiene que llevar la línea ⚠ PLAZO: de la respuesta.
	Fecha string `yaml:"fecha"`

	// Avisos son los códigos de plazos.CodigosDeAviso que da plazos con la
	// entrada de la eval, en el orden del fichero y sin repetir: la respuesta
	// tiene que llevar la forma fija de cada uno. Vacío si no da ninguno.
	Avisos []string `yaml:"avisos"`

	// Razonamiento es el cálculo a mano de la fecha, independiente del código.
	// Ningún juicio lo lee.
	Razonamiento string `yaml:"razonamiento"`
}

// NoCubiertoEsperado es lo que una eval de cobertura espera de la respuesta
// (contracts/evals-legal-core.md §1 de H9; data-model §3 de H9).
type NoCubiertoEsperado struct {
	// Norma es el identificador de la norma que rige el procedimiento, BOE-A-…,
	// que tiene que llevar la línea ⚠ NO CUBIERTO: de la respuesta.
	Norma string `yaml:"norma"`
}

// PlazoSinCalcularEsperado es lo que una eval de vigencia espera de la
// respuesta (ADR 0042).
type PlazoSinCalcularEsperado struct {
	// Norma es el identificador BOE-A-… de una norma de los preceptos que no
	// regían el día de inicio, que tiene que llevar la línea
	// ⚠ PLAZO SIN CALCULAR: de la respuesta.
	Norma string `yaml:"norma"`
}

// SentenciasEsperadas es lo que una eval espera de la respuesta sobre las
// sentencias de las que habla (contracts/evals-jurisprudencia.md §1 y §2 de
// H23): cada clave que la eval escribe es una cosa que el juicio compara por su
// forma, sin ningún modelo, y lo que no escribe queda vacío. Una eval que la
// declara escribe al menos una, y nunca Citas y NingunaCita a la vez.
type SentenciasEsperadas struct {
	// Citas son las citas de sentencia que la respuesta tiene que llevar, cada
	// una con su corchete, en el orden del fichero.
	Citas []CitaDeSentenciaEsperada `yaml:"citas"`

	// NingunaCita dice que la respuesta no puede llevar ninguna cita de
	// sentencia, del ECLI que sea.
	NingunaCita bool `yaml:"ninguna_cita"`

	// SinCitaDelROJ son los ROJ con los que ninguna cita de la respuesta puede
	// ir, en el orden del fichero.
	SinCitaDelROJ []string `yaml:"sin_cita_del_roj"`

	// NoComprobada dice que la respuesta tiene que llevar la línea
	// ⚠ SENTENCIA NO COMPROBADA:.
	NoComprobada bool `yaml:"no_comprobada"`

	// Direcciones son las direcciones que la respuesta tiene que llevar, tal
	// cual, en el orden del fichero.
	Direcciones []string `yaml:"direcciones"`

	// Casillas son las casillas del buscador cuyo nombre y cuyo valor tiene que
	// llevar la respuesta, tal cual, en el orden del fichero.
	Casillas []CasillaEsperada `yaml:"casillas"`

	// DireccionDeBusqueda dice que la respuesta tiene que llevar la dirección
	// que devolvió en la sesión un cita preparar con texto.
	DireccionDeBusqueda bool `yaml:"direccion_de_busqueda"`
}

// declaradas dice si la eval declara sentencias: si escribe alguna de sus
// claves, que es lo que su esquema exige a la que escribe la clave sentencias.
func (s SentenciasEsperadas) declaradas() bool {
	return len(s.Citas) > 0 || s.NingunaCita || len(s.SinCitaDelROJ) > 0 || s.NoComprobada ||
		len(s.Direcciones) > 0 || len(s.Casillas) > 0 || s.DireccionDeBusqueda
}

// CitaDeSentenciaEsperada es una cita de sentencia que la respuesta tiene que
// contener: la pareja exacta de ECLI y ROJ de su corchete
// (contracts/evals-jurisprudencia.md §2 de H23).
type CitaDeSentenciaEsperada struct {
	// ECLI es el ECLI de la sentencia citada, español.
	ECLI string `yaml:"ecli"`

	// ROJ es el ROJ de la sentencia citada.
	ROJ string `yaml:"roj"`
}

// CasillaEsperada es una casilla del buscador del CENDOJ que la respuesta tiene
// que dar: su nombre y su valor, como los da data de cita preparar.
type CasillaEsperada struct {
	// Nombre es el nombre de la casilla en el buscador, como «Nº ROJ».
	Nombre string `yaml:"nombre"`

	// Valor es lo que hay que escribir en ella.
	Valor string `yaml:"valor"`
}

// ComandoEsperado es un comando que la sesión tiene que ejecutar, en una de las
// ocho formas excluyentes de data-model §6.1 de H5 y de H6, del contrato
// evals-y-skill §1 de H7, de contracts/evals-jurisprudencia.md §1 de H23 y de
// contracts/evals-legal-core.md §1 de H9, que decide formaDelComando: bloque
// (Applet, Norma y Bloque, sin Verbo, y desde H20 Fecha opcional;
// contracts/evals-redaccion-fechada.md §1.1 de H20), consulta de norma (Applet,
// Verbo indice, metadatos o analisis, y Norma), búsqueda (Applet, Verbo buscar y
// Terminos), territorio (Applet, Verbo resolver y Municipio), comprobación
// (Applet y Verbo check, y desde H7.1 Norma opcional; contrato evals-y-skill §1
// de H7.1), desde H23, preparar (Applet y Verbo preparar, con ROJ y ConTexto
// opcionales) y cotejar (Applet y Verbo cotejar, con ROJ opcional) o, desde H9,
// calcular (Applet, Verbo calcular y Tipo). Lo que su forma no lleva queda
// vacío.
type ComandoEsperado struct {
	// Applet es el applet que se invoca, como boe, territorio o graph.
	Applet string `yaml:"applet"`

	// Verbo es el verbo de una consulta de norma, de una búsqueda, de un comando
	// de territorio o de una comprobación; vacío en la forma bloque.
	Verbo string `yaml:"verbo"`

	// Norma es el identificador de la norma de la forma bloque o de una
	// consulta de norma, o el de la norma que consulta una comprobación; vacío en
	// una comprobación de todo lo consultado.
	Norma string `yaml:"norma"`

	// Bloque es el id del bloque de la forma bloque.
	Bloque string `yaml:"bloque"`

	// Terminos son los términos que tienen que aparecer, como palabras, en los
	// argumentos de una búsqueda.
	Terminos []string `yaml:"terminos"`

	// Municipio es el municipio que resuelve un comando de territorio, tal como
	// lo escribe la eval.
	Municipio string `yaml:"municipio"`

	// ROJ es el ROJ que un comando de preparar o de cotejar tiene que recibir
	// con --roj; vacío si la eval no lo exige.
	ROJ string `yaml:"roj"`

	// ConTexto dice que un comando de preparar tiene que recibir --texto con
	// algún valor: es una búsqueda por texto, del texto que sea.
	ConTexto bool `yaml:"con_texto"`

	// Tipo es el tipo de plazo que un comando calcular tiene que pedir, como
	// reposicion o contencioso.
	Tipo string `yaml:"tipo"`

	// SinCalcular dice que el plazo del comando calcular no se calcula, porque
	// alguno de sus preceptos no regía así el día de inicio: la invocación
	// termina con 3 (ADR 0042).
	SinCalcular bool `yaml:"sin_calcular"`

	// Fecha es el día, AAAA-MM-DD, con el que la forma bloque tiene que leer su
	// bloque, con --fecha o con fecha en la herramienta; vacía en la lectura de
	// hoy y en las demás formas (contracts/evals-redaccion-fechada.md §1.1 de
	// H20; FR-051).
	Fecha string `yaml:"fecha"`
}

// verboResolver es el verbo de un comando de territorio, el único del applet
// territorio (contrato de evals §1.1 de H6).
const verboResolver = "resolver"

// verboCheck es el verbo de un comando de comprobación, el de graph check
// (contrato evals-y-skill §1 de H7).
const verboCheck = "check"

// Los verbos de los dos comandos de cita, los del applet cita
// (contracts/evals-jurisprudencia.md §1 de H23).
const (
	verboPreparar = "preparar"
	verboCotejar  = "cotejar"
)

// verboCalcular es el verbo de un comando calcular, el de plazos calcular
// (contracts/evals-legal-core.md §1 de H9).
const verboCalcular = "calcular"

// ComandoProhibido es un comando que ninguna invocación de la sesión puede
// ejecutar: un verbo de un applet, con cualquier argumento (contrato
// evals-y-skill §1 y §2 de H7).
type ComandoProhibido struct {
	// Applet es el applet del comando, como graph.
	Applet string `yaml:"applet"`

	// Verbo es el verbo del comando, como show.
	Verbo string `yaml:"verbo"`
}

// GrafoPrevio es el estado del grafo con el que empieza la sesión de una eval: el
// que dejan los comandos de bloque ejecutados sobre las grabaciones derivadas
// nombradas antes de preparar la caché de la sesión (contrato evals-y-skill §1 y
// §3 de H7; research D26).
type GrafoPrevio struct {
	// Grabaciones es el nombre del conjunto de grabaciones derivadas con las que
	// se prepara el grafo.
	Grabaciones string `yaml:"grabaciones"`

	// Comandos son los comandos de bloque que llenan el grafo, en el orden del
	// fichero; solo tienen la forma bloque.
	Comandos []ComandoEsperado `yaml:"comandos"`
}

// formaDeComando es una de las ocho formas de un comando esperado.
type formaDeComando int

// Las ocho formas de un comando esperado (data-model §6.1 de H5 y de H6;
// contrato evals-y-skill §1 de H7; contracts/evals-jurisprudencia.md §1 de
// H23; contracts/evals-legal-core.md §1 de H9).
const (
	// formaConsultaDeNorma es la de una consulta de norma: el verbo con la norma.
	formaConsultaDeNorma formaDeComando = iota

	// formaBloque es la de la lectura de un bloque de una norma, sin verbo.
	formaBloque

	// formaBusqueda es la de una búsqueda: buscar con sus términos.
	formaBusqueda

	// formaTerritorio es la de un comando de territorio: resolver con el
	// municipio.
	formaTerritorio

	// formaComprobacion es la de un comando de comprobación: check, con la
	// norma que consulta o sin ella.
	formaComprobacion

	// formaPreparar es la de un comando que prepara la consulta de una
	// sentencia: preparar, con el ROJ que recibe, con un texto o sin nada más.
	formaPreparar

	// formaCotejar es la de un comando que coteja el documento de una
	// sentencia: cotejar, con el ROJ que recibe o sin él.
	formaCotejar

	// formaCalcular es la de un comando que calcula un plazo: calcular, con el
	// tipo que pide.
	formaCalcular
)

// formaDelComando es la forma del comando esperado, que decide su verbo, el que
// distingue las ocho de data-model §6.1, del contrato evals-y-skill §1 de H7,
// de contracts/evals-jurisprudencia.md §1 de H23 y de
// contracts/evals-legal-core.md §1 de H9: sin verbo, la forma bloque; buscar,
// la búsqueda; resolver, el comando de territorio; check, la comprobación;
// preparar y cotejar, los dos de cita; calcular, el de un plazo; y cualquier
// otro, la consulta de norma, porque el esquema de eval solo admite en esa
// forma los verbos de su enumerado.
// Es el único sitio que decide la variante: la consumen el juicio, el texto del
// comando y las consultas necesarias (research D21 de H6).
func formaDelComando(comando ComandoEsperado) formaDeComando {
	switch comando.Verbo {
	case "":
		return formaBloque
	case verboBuscar:
		return formaBusqueda
	case verboResolver:
		return formaTerritorio
	case verboCheck:
		return formaComprobacion
	case verboPreparar:
		return formaPreparar
	case verboCotejar:
		return formaCotejar
	case verboCalcular:
		return formaCalcular
	default:
		return formaConsultaDeNorma
	}
}

// CitaEsperada es una cita que la respuesta tiene que contener: la pareja exacta
// de norma y bloque (data-model §6.2).
type CitaEsperada struct {
	// Norma es el identificador de la norma citada.
	Norma string `yaml:"norma"`

	// Bloque es el id del bloque citado.
	Bloque string `yaml:"bloque"`
}

// RedaccionEsperada es una redacción modificada que la respuesta tiene que
// trasladar: el bloque de la norma cuya redacción cambió entre dos lecturas, con
// la fecha de vigencia de cada una, con los nombres de los campos del hallazgo
// version-obsoleta de graph check (data-model §2 de H7.4; research D9). Su texto,
// en los motivos y en el informe, es <norma> <bloque> <fecha_vigencia>
// <fecha_vigencia_reciente>.
type RedaccionEsperada struct {
	// Norma es el identificador de la norma, BOE-A-….
	Norma string `yaml:"norma"`

	// Bloque es el id del bloque cuya redacción cambió.
	Bloque string `yaml:"bloque"`

	// FechaVigencia es la fecha de vigencia de la redacción superada, la de la
	// lectura anterior, en la forma AAAAMMDD.
	FechaVigencia string `yaml:"fecha_vigencia"`

	// FechaVigenciaReciente es la fecha de vigencia de la redacción leída, la
	// vigente, en la forma AAAAMMDD.
	FechaVigenciaReciente string `yaml:"fecha_vigencia_reciente"`
}

// RedaccionFechadaEsperada es lo que una eval con fecha espera de la respuesta
// sobre la redacción de un bloque en un día (contracts/evals-redaccion-fechada.md
// §1.2 de H20; data-model §7.2; research D11; FR-052): la norma, el bloque y el
// día pedido, y o bien la fecha de vigencia de la redacción vigente ese día y, si
// hoy la redacción es otra, la de hoy, o bien, si ese día el bloque no tenía
// redacción vigente, la fecha de vigencia de la primera. El esquema no admite
// las dos.
type RedaccionFechadaEsperada struct {
	// Norma es el identificador de la norma, BOE-A-….
	Norma string `yaml:"norma"`

	// Bloque es el id del bloque.
	Bloque string `yaml:"bloque"`

	// Fecha es el día pedido, AAAA-MM-DD.
	Fecha string `yaml:"fecha"`

	// FechaVigencia es la fecha de vigencia, AAAAMMDD, de la redacción vigente
	// el día pedido, la que se cita; vacía si ese día no había ninguna.
	FechaVigencia string `yaml:"fecha_vigencia"`

	// Posterior es la redacción de hoy cuando no es la citada; nil si la citada
	// es la de hoy o si ese día no había ninguna.
	Posterior *RedaccionPosteriorEsperada `yaml:"posterior"`

	// VigenteDesde es la fecha de vigencia, AAAAMMDD, de la primera redacción
	// del bloque cuando el día pedido no tenía ninguna; vacía en otro caso.
	VigenteDesde string `yaml:"vigente_desde"`
}

// RedaccionPosteriorEsperada es la redacción de hoy de un bloque cuya redacción
// en el día pedido es otra: su fecha de vigencia y la norma que la dio, como las
// da el binario en redaccion_posterior (data-model §4 y §7.2 de H20).
type RedaccionPosteriorEsperada struct {
	// FechaVigencia es la fecha de vigencia de la redacción de hoy, AAAAMMDD.
	FechaVigencia string `yaml:"fecha_vigencia"`

	// NormaModificadora es el identificador BOE-A-… de la norma que la dio.
	NormaModificadora string `yaml:"norma_modificadora"`
}

// LecturaDeLaEval es un bloque de una norma consolidada del BOE que la caché de
// las sesiones de la skill sirve porque la eval lo provoca, y por qué: no es un
// comando ni una cita de la eval, ni da motivo, ni cuenta en ningún umbral (ADR
// 0043, clase (b); data-model §1 de H26.1; FR-007).
type LecturaDeLaEval struct {
	// Norma es el identificador de la norma, BOE-A-….
	Norma string `yaml:"norma"`

	// Bloque es el id del bloque.
	Bloque string `yaml:"bloque"`

	// Motivo dice por qué la eval provoca la lectura; el esquema no lo admite
	// vacío ni solo de blancos.
	Motivo string `yaml:"motivo"`
}

// LecturaSinServir es lo que la caché de las sesiones de una eval no sirve con
// ninguna fecha: un bloque de una norma, o su índice o sus metadatos
// (data-model §1 de H26.1; FR-013). Lleva Bloque o Verbo, nunca los dos: el
// esquema no admite otra cosa.
type LecturaSinServir struct {
	// Norma es el identificador de la norma, BOE-A-….
	Norma string `yaml:"norma"`

	// Bloque es el id del bloque que no se sirve; vacío en una consulta de
	// norma.
	Bloque string `yaml:"bloque"`

	// Verbo es la consulta de norma que no se sirve, indice o metadatos; vacío
	// en un bloque.
	Verbo string `yaml:"verbo"`
}

// consulta dice si la invocación es del applet boe y consulta la lectura
// (data-model §1 de H26.1; research D3 de H26.1): la de un bloque, articulo o
// articulos con la norma como primer argumento de posición y el bloque entre los
// demás, con --fecha o sin ella, leídos por lecturaDeLosArgumentos; la consulta
// de norma, ese verbo con esa norma (esDeLaNorma).
func (l LecturaSinServir) consulta(invocacion Invocacion) bool {
	if invocacion.Applet != appletDeLasNormas {
		return false
	}

	if l.Bloque == "" {
		return invocacion.Verbo == l.Verbo && esDeLaNorma(invocacion, l.Norma)
	}

	if invocacion.Verbo != verboArticulo && invocacion.Verbo != verboArticulos {
		return false
	}

	lectura := lecturaDeLosArgumentos(invocacion.Argumentos)

	return len(lectura.dePosicion) > 0 && lectura.dePosicion[0] == l.Norma &&
		slices.Contains(lectura.dePosicion[1:], l.Bloque)
}

// comprobarLasLecturas comprueba de Lecturas y de SinServir lo que el esquema no
// puede, porque cruza claves de la eval (contracts/formato-de-eval.md §2 de
// H26.1; FR-008, FR-014): ningún bloque de Lecturas es un comando de bloque de la
// eval, con fecha o sin ella, ni una de sus citas, y ningún bloque de SinServir
// es una de sus citas. Un bloque de SinServir sí puede ser un comando de bloque
// (FR-016). El error, que nombra la clave y la entrada, es el del primer defecto
// en el orden del fichero, Lecturas antes que SinServir; el de una lectura que
// es comando y cita a la vez es el del comando.
func (e Eval) comprobarLasLecturas() error {
	for _, lectura := range e.Lecturas {
		switch {
		case e.esComandoDeBloque(lectura.Norma, lectura.Bloque):
			return fmt.Errorf("lecturas: %s %s: ya es un comando esperado de la eval", lectura.Norma, lectura.Bloque)
		case e.esCitaEsperada(lectura.Norma, lectura.Bloque):
			return fmt.Errorf("lecturas: %s %s: ya es una cita esperada de la eval", lectura.Norma, lectura.Bloque)
		}
	}

	for _, lectura := range e.SinServir {
		if lectura.Bloque != "" && e.esCitaEsperada(lectura.Norma, lectura.Bloque) {
			return fmt.Errorf("sin_servir: %s %s: es una cita esperada de la eval", lectura.Norma, lectura.Bloque)
		}
	}

	return nil
}

// esComandoDeBloque dice si el bloque de la norma es el de un comando de la
// forma bloque de la eval, con fecha o sin ella.
func (e Eval) esComandoDeBloque(norma, bloque string) bool {
	return slices.ContainsFunc(e.Comandos, func(comando ComandoEsperado) bool {
		return formaDelComando(comando) == formaBloque && comando.Norma == norma && comando.Bloque == bloque
	})
}

// esCitaEsperada dice si el bloque de la norma es una de las citas de la eval.
func (e Eval) esCitaEsperada(norma, bloque string) bool {
	return slices.Contains(e.Citas, CitaEsperada{Norma: norma, Bloque: bloque})
}

// sinServir dice si la invocación consulta alguna lectura del sin_servir de la
// eval (data-model §1 de H26.1; FR-013).
func (e Eval) sinServir(invocacion Invocacion) bool {
	return slices.ContainsFunc(e.SinServir, func(lectura LecturaSinServir) bool {
		return lectura.consulta(invocacion)
	})
}

// TerritorioEsperado es lo que la respuesta tiene que declarar del territorio de
// un municipio (contrato de evals §1.2 de H6): cada clave que la eval escribe es
// un elemento que se busca en la respuesta por su forma fija, y lo que no escribe
// queda vacío.
type TerritorioEsperado struct {
	// Comunidad es el nombre de la comunidad o ciudad autónoma.
	Comunidad string `yaml:"comunidad"`

	// Provincia es el nombre de la provincia.
	Provincia string `yaml:"provincia"`

	// Boletines son los códigos de los boletines, en el orden del fichero.
	Boletines []string `yaml:"boletines"`

	// Cobertura tiene cada aspecto de cobertura en la forma <aspecto>: <valor>,
	// del vocabulario del applet territorio, en el orden del fichero.
	Cobertura []string `yaml:"cobertura"`
}

// esquemaDeEval compila una sola vez el esquema publicado del formato común de
// eval, que no cambia mientras se ejecutan los tests.
var esquemaDeEval = sync.OnceValues(func() (*jsonschema.Schema, error) {
	return compilarEsquemaDeEval(rutaDelEsquemaDeEval)
})

// compilarEsquemaDeEval lee el esquema del formato común de eval de la ruta y lo
// compila con compilarEsquemaPublicado.
func compilarEsquemaDeEval(ruta string) (*jsonschema.Schema, error) {
	return compilarEsquemaPublicado(ruta, "del formato de eval")
}

// compilarEsquemaPublicado lee un esquema publicado de la ruta y lo compila con
// skills.CompilarEsquema; de dice de qué documentos es, como «del formato de
// eval», y va en el error. El error nombra la ruta: la del fichero que no se
// puede leer o la del esquema que no compila.
func compilarEsquemaPublicado(ruta, de string) (*jsonschema.Schema, error) {
	contenido, err := leerFichero(ruta)
	if err != nil {
		return nil, fmt.Errorf("no se puede leer el esquema %s: %w", de, err)
	}

	esquema, err := skills.CompilarEsquema(contenido)
	if err != nil {
		return nil, fmt.Errorf("el esquema %s %s: %w", de, ruta, err)
	}

	return esquema, nil
}

// LeerEval lee el contenido de un fichero de eval con el lector común de
// documentos YAML de internal/skills —una clave repetida es un defecto con sus
// dos líneas, nunca la última que gana— y lo valida contra eval.yaml.json y,
// detrás, con comprobarLasLecturas, que cruza lecturas y sin_servir con los
// comandos y las citas de la eval (contracts/formato-de-eval.md §2 de H26.1). Todo
// error empieza por el nombre del fichero y va con una Eval vacía; la Eval leída
// lleva ese nombre en Fichero (US4, escenario 4; contrato evals-y-grabaciones §1).
func LeerEval(nombre string, contenido []byte) (Eval, error) {
	esquema, err := esquemaDeEval()
	if err != nil {
		return Eval{}, fmt.Errorf("%s: %w", nombre, err)
	}

	eval, err := skills.ValidarDocumentoYAML[Eval](contenido, esquema)
	if err != nil {
		return Eval{}, fmt.Errorf("%s: %w", nombre, err)
	}

	if err := eval.comprobarLasLecturas(); err != nil {
		return Eval{}, fmt.Errorf("%s: %w", nombre, err)
	}

	eval.Fichero = nombre

	return eval, nil
}
