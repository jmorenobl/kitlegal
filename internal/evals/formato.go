package evals

import (
	"fmt"
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
	// modificadas, territorio, informativa ni reproduce: decide siempre y se mide
	// una vez por modelo, fuera de los dos modos (data-model §7 y
	// contracts/evals-en-dos-modos.md §1 de H21; FR-046).
	SinBinarioNiServidor bool `yaml:"sin_binario_ni_servidor"`

	// Sentencias es lo que la respuesta tiene que llevar, y lo que no, de las
	// sentencias de las que habla, cada cosa por su forma; vacío si la eval no lo
	// declara. Solo lo admite una eval que activa la skill y que no es sin
	// binario ni servidor, y con él puede no llevar comandos, citas ni territorio
	// (contracts/evals-jurisprudencia.md §1 de H23; FR-052).
	Sentencias SentenciasEsperadas `yaml:"sentencias"`
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
// siete formas excluyentes de data-model §6.1 de H5 y de H6, del contrato
// evals-y-skill §1 de H7 y de contracts/evals-jurisprudencia.md §1 de H23, que
// decide formaDelComando: bloque (Applet, Norma y Bloque, sin Verbo), consulta
// de norma (Applet, Verbo indice, metadatos o analisis, y Norma), búsqueda
// (Applet, Verbo buscar y Terminos), territorio (Applet, Verbo resolver y
// Municipio), comprobación (Applet y Verbo check, y desde H7.1 Norma opcional;
// contrato evals-y-skill §1 de H7.1) o, desde H23, preparar (Applet y Verbo
// preparar, con ROJ y ConTexto opcionales) y cotejar (Applet y Verbo cotejar,
// con ROJ opcional). Lo que su forma no lleva queda vacío.
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

// formaDeComando es una de las siete formas de un comando esperado.
type formaDeComando int

// Las siete formas de un comando esperado (data-model §6.1 de H5 y de H6;
// contrato evals-y-skill §1 de H7; contracts/evals-jurisprudencia.md §1 de
// H23).
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
)

// formaDelComando es la forma del comando esperado, que decide su verbo, el que
// distingue las siete de data-model §6.1, del contrato evals-y-skill §1 de H7 y
// de contracts/evals-jurisprudencia.md §1 de H23: sin verbo, la forma bloque;
// buscar, la búsqueda; resolver, el comando de territorio; check, la
// comprobación; preparar y cotejar, los dos de cita; y cualquier otro, la
// consulta de norma, porque el esquema de eval solo admite en esa forma los
// verbos de su enumerado.
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
// dos líneas, nunca la última que gana— y lo valida contra eval.yaml.json. Todo
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

	eval.Fichero = nombre

	return eval, nil
}
