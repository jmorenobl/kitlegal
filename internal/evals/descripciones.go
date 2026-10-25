package evals

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/jmorenobl/kitlegal/internal/skills"
)

// directorioDelJuezDeLasDescripciones es el de la evidencia del ADR 0046, desde
// la raíz del repositorio y con la barra de separador: de él se leen, tal cual y
// sin copia, la rúbrica, el esquema, los casos y la medida del juez de las
// descripciones (contracts/juez-de-las-descripciones.md §1 de H31; research D2
// de H31; FR-025).
const directorioDelJuezDeLasDescripciones = "evidencias/adr-0046"

// manifiestoDeLaEvidencia es el manifiesto del directorio de una evidencia (ADR
// 0018), que registra la huella SHA-256 de cada uno de sus ficheros.
const manifiestoDeLaEvidencia = "manifiesto.json"

// claseDeLasDescripciones es la única clase del juez de las descripciones, que
// decide (FR-025 de H31; research D6 de H31).
const claseDeLasDescripciones = "habla_en_primera_o_segunda_persona"

// peticionDelVotoDeLaDescripcion es la frase final del mensaje del voto de una
// descripción, la de mensaje_de de evidencias/adr-0046/guiones/juez.py
// (contracts/juez-de-las-descripciones.md §4 de H31).
const peticionDelVotoDeLaDescripcion = "Responde a la pregunta de la rúbrica sobre esta descripción."

// Los campos de la huella del juicio de una descripción que no son de un
// fichero del juez (contracts/juez-de-las-descripciones.md §2 de H31): la skill.
// La descripción y el juez entran con campoDescripcion y campoJuez, los de la
// huella de una eval.
const campoSkill = "skill"

// ficheroUsadoPorElJuez es uno de los cuatro ficheros de la evidencia que usa el
// juez de las descripciones: su nombre en el directorio y el campo con el que
// entra en la huella del juicio.
type ficheroUsadoPorElJuez struct {
	fichero string
	campo   string
}

// ficherosUsadosPorElJuez son los cuatro ficheros de la evidencia que usa el
// juez de las descripciones, en el orden del contrato: la rúbrica, el esquema,
// los casos y la medida (contracts/juez-de-las-descripciones.md §1 y §2 de H31).
var ficherosUsadosPorElJuez = []ficheroUsadoPorElJuez{
	{fichero: ficheroDeRubricaDelJuez, campo: "rubrica"},
	{fichero: ficheroDeEsquemaDelJuez, campo: "esquema"},
	{fichero: ficheroDeCasosDelJuez, campo: "casos"},
	{fichero: ficheroDeMedidaDelJuez, campo: "medida"},
}

// contenidosDelJuez son los contenidos de los cuatro ficheros que lleva el juez,
// en el orden de ficherosUsadosPorElJuez.
func contenidosDelJuez(juez *Juez) []string {
	return []string{juez.Rubrica, juez.Esquema, juez.Casos, juez.Medida}
}

// enLaEvidencia es el nombre de un fichero de la evidencia del juez de las
// descripciones desde la raíz del repositorio, con la barra de separador: con
// él lo nombran los errores y las líneas del control de identidad.
func enLaEvidencia(fichero string) string {
	return path.Join(directorioDelJuezDeLasDescripciones, fichero)
}

// leerDeLaEvidencia lee entero el fichero de la evidencia del juez de las
// descripciones de esa raíz del repositorio. El error lo nombra.
func leerDeLaEvidencia(raiz, fichero string) ([]byte, error) {
	contenido, err := leerFichero(filepath.Join(raiz, filepath.FromSlash(enLaEvidencia(fichero))))
	if err != nil {
		return nil, fmt.Errorf("%s no se puede leer: %w", enLaEvidencia(fichero), err)
	}

	return contenido, nil
}

// leerLosFicherosDelJuez lee enteros, de la evidencia de esa raíz, los cuatro
// ficheros que usa el juez de las descripciones, en el orden de
// ficherosUsadosPorElJuez. El error nombra cada uno que no se puede leer.
func leerLosFicherosDelJuez(raiz string) ([][]byte, error) {
	contenidos := make([][]byte, len(ficherosUsadosPorElJuez))

	var errores []error

	for posicion, usado := range ficherosUsadosPorElJuez {
		contenido, err := leerDeLaEvidencia(raiz, usado.fichero)
		if err != nil {
			errores = append(errores, err)

			continue
		}

		contenidos[posicion] = contenido
	}

	if err := errors.Join(errores...); err != nil {
		return nil, err
	}

	return contenidos, nil
}

// juezDeLasDescripciones es el juez de las descripciones de las skills
// (contracts/juez-de-las-descripciones.md §1 de H31; research D2 y D6 de H31;
// FR-025): el de la clase claseDeLasDescripciones, que decide, con el contenido
// entero de la rúbrica, el esquema, los casos y la medida de la evidencia del
// ADR 0046 de esa raíz del repositorio, leídos donde están. Lo votan, como al
// de las respuestas, nuevaVotacion y el votante del guion.
//
// Como el de una skill, no hay juez si alguno de los cuatro no se puede leer o
// si las propiedades del esquema no son exactamente su clase. Que los cuatro
// sean los que registra el manifiesto de la evidencia lo comprueba aparte
// identidadDelJuezDeLasDescripciones, y que la medida corresponda y se cumpla,
// comprobarLaMedida.
func juezDeLasDescripciones(raiz string) (*Juez, error) {
	contenidos, err := leerLosFicherosDelJuez(raiz)
	if err != nil {
		return nil, fmt.Errorf("el juez de las descripciones: %w", err)
	}

	juez := &Juez{
		Clases:  []ClaseDelJuez{{Nombre: claseDeLasDescripciones, Decide: true}},
		Rubrica: string(contenidos[0]),
		Esquema: string(contenidos[1]),
		Casos:   string(contenidos[2]),
		Medida:  string(contenidos[3]),
	}

	propiedades, err := propiedadesDelEsquema([]byte(juez.Esquema))
	if err != nil {
		return nil, fmt.Errorf("el juez de las descripciones: %s: %w", enLaEvidencia(ficheroDeEsquemaDelJuez), err)
	}

	if !slices.Equal(propiedades, []string{claseDeLasDescripciones}) {
		return nil, fmt.Errorf("el juez de las descripciones: %s: sus propiedades (%s) no son exactamente su clase (%s)",
			enLaEvidencia(ficheroDeEsquemaDelJuez), strings.Join(propiedades, ", "), claseDeLasDescripciones)
	}

	return juez, nil
}

// Las líneas del control de identidad del juez de las descripciones: la del
// fichero que el manifiesto de la evidencia no registra y la del que registra
// con otra huella, cada una con el fichero desde la raíz del repositorio
// delante.
const (
	ficheroSinRegistrar = "%s: el manifiesto de la evidencia no lo registra"
	ficheroDeOtraHuella = "%s: su SHA-256 es %s y el manifiesto de la evidencia registra %s"
)

// entradaDelManifiesto es lo que se lee de la entrada de un fichero en el
// manifiesto de una evidencia: su nombre en el directorio y su huella SHA-256,
// en hexadecimal. Las demás claves no se leen.
type entradaDelManifiesto struct {
	Fichero string `json:"fichero"`
	SHA256  string `json:"sha256"`
}

// manifiestoLeido es lo que se lee del manifiesto de una evidencia: la entrada
// de cada fichero.
type manifiestoLeido struct {
	Ficheros []entradaDelManifiesto `json:"ficheros"`
}

// identidadDelJuezDeLasDescripciones es el control de identidad de FR-025 y
// SC-006 de H31 (research D2 de H31): los cuatro ficheros que usa el juez de las
// descripciones, en la evidencia de esa raíz del repositorio, son, byte a byte,
// los que registra el manifiesto de la evidencia. Devuelve, en el orden de
// ficherosUsadosPorElJuez, una línea por fichero que el manifiesto no registra
// y una por cada entrada suya que lo registra con otra huella; ninguna si los
// cuatro son los registrados.
//
// El error es el de una evidencia en la que no hay identidad que comprobar:
// alguno de los cuatro, o el manifiesto, no se puede leer, o el manifiesto no
// es JSON con su forma.
func identidadDelJuezDeLasDescripciones(raiz string) ([]string, error) {
	contenidos, err := leerLosFicherosDelJuez(raiz)
	if err != nil {
		return nil, err
	}

	leido, err := leerDeLaEvidencia(raiz, manifiestoDeLaEvidencia)
	if err != nil {
		return nil, err
	}

	var manifiesto manifiestoLeido
	if err := json.Unmarshal(leido, &manifiesto); err != nil {
		return nil, fmt.Errorf("%s no se puede leer como JSON: %w", enLaEvidencia(manifiestoDeLaEvidencia), err)
	}

	var lineas []string

	for posicion, usado := range ficherosUsadosPorElJuez {
		nombre := enLaEvidencia(usado.fichero)
		huella := huellaSHA256(contenidos[posicion])
		registrado := false

		for _, entrada := range manifiesto.Ficheros {
			if entrada.Fichero != usado.fichero {
				continue
			}

			registrado = true

			if entrada.SHA256 != huella {
				lineas = append(lineas, fmt.Sprintf(ficheroDeOtraHuella, nombre, huella, entrada.SHA256))
			}
		}

		if !registrado {
			lineas = append(lineas, fmt.Sprintf(ficheroSinRegistrar, nombre))
		}
	}

	return lineas, nil
}

// mensajeDeLaDescripcion es el mensaje con el que se pide cada voto de la
// descripción de una skill, el de mensaje_de de
// evidencias/adr-0046/guiones/juez.py carácter a carácter
// (contracts/juez-de-las-descripciones.md §4 de H31; research D5 de H31;
// FR-027): la descripción, la que da LeerFrontmatter, sin los blancos de los
// extremos que quita str.strip() (esBlanco), entre <descripcion
// skill="<skill>"> y </descripcion>, cada marca en su línea; una línea en
// blanco; y la frase final, sin salto de línea detrás.
func mensajeDeLaDescripcion(skill, descripcion string) string {
	return `<descripcion skill="` + skill + `">` + "\n" +
		strings.TrimFunc(descripcion, esBlanco) + "\n" +
		"</descripcion>\n\n" +
		peticionDelVotoDeLaDescripcion
}

// VotoDeLaDescripcion es lo que el juicio de una descripción guarda de cada voto
// que se dio (data-model §4 de H31; contracts/juez-de-las-descripciones.md §6
// de H31): lo que el juez dijo de su única clase, tal como lo dio. Sus claves
// JSON, en el orden de sus campos, son las de un voto del juicio de una
// descripción.
type VotoDeLaDescripcion struct {
	// Respuesta es «si» o «no».
	Respuesta string `json:"respuesta"`

	// Frase es la frase de la descripción que el juez cita como prueba de su
	// sí; vacía con un no.
	Frase string `json:"frase"`

	// Motivo es el motivo que el juez da de su respuesta.
	Motivo string `json:"motivo"`
}

// JuicioDeLaDescripcion es el juicio del juez de las descripciones sobre la
// descripción de una skill, el que da la regla de los votos con su única clase
// (data-model §4 de H31; contracts/juez-de-las-descripciones.md §5 y §6 de H31;
// FR-028, FR-029, FR-032). Sus claves JSON son las del juicio de una
// descripción.
type JuicioDeLaDescripcion struct {
	// Marcada dice si tres votos dicen sí con su frase en la descripción. Con la
	// descripción sin juzgar es falsa.
	Marcada bool `json:"marcada"`

	// Frase es la del primer voto, si la descripción está marcada; vacía si no.
	// Del primer voto cuenta, como en la regla de los votos, la repetición de un
	// nulo o de un cortado.
	Frase string `json:"frase"`

	// Motivo es el del primer voto, el que cuenta, esté o no marcada la
	// descripción; vacío si no llegó a darse.
	Motivo string `json:"motivo"`

	// Votos son todos los votos que se dieron, en su orden: también los nulos,
	// delante de su repetición, y no los que el tope cortó, que no dieron nada.
	Votos []VotoDeLaDescripcion `json:"votos"`

	// SinJuzgar es el motivo del voto que no llegó a darse, con su número;
	// vacío, y sin clave, si la descripción se juzgó.
	SinJuzgar string `json:"sin_juzgar,omitzero"`
}

// juzgarLaDescripcion aplica a la descripción de una skill la regla de los
// votos de la votación (votar), que es la del juez de las descripciones, con
// su única clase, que decide (contracts/juez-de-las-descripciones.md §5 de H31;
// research D6 de H31; FR-028, FR-029, FR-032): pide cada voto, por orden, con
// el mensaje de la descripción, mensajeDeLaDescripcion, y busca su frase en la
// descripción con fraseEsta. Así, la descripción queda marcada con tres «si»
// con su frase; el segundo voto se pide solo tras un primer «si», y el tercero
// solo tras el segundo; un «si» con una frase que no está es nulo y se pide
// otra vez, una sola; el voto que el tope corta se pide otra vez, una sola, y
// cortado otra vez deja la descripción sin juzgar; y la sesión con error, sin
// la forma del esquema o cortada por un límite de uso, sin repetición, la deja
// sin juzgar, con el motivo del voto. Como mucho, votosPorRespuestaComoMucho
// votos.
func (v *votacion) juzgarLaDescripcion(skill, descripcion string) JuicioDeLaDescripcion {
	return juicioDeLaDescripcion(v.votarLaDescripcion(skill, descripcion))
}

// votarLaDescripcion es el juicio de la regla de los votos sobre la descripción
// de una skill, con los votos de cada clase del juez tal como se dieron: el que
// presenta juzgarLaDescripcion en el trabajo de la skill y el que cuenta la
// medida del juez de las descripciones con sus casos
// (medirAlJuezDeLasDescripciones). Pide cada voto con el mensaje de la
// descripción y busca su frase en la descripción.
func (v *votacion) votarLaDescripcion(skill, descripcion string) JuicioDeRespuesta {
	return v.votar(mensajeDeLaDescripcion(skill, descripcion), descripcion)
}

// juicioDeLaDescripcion es el juicio de una descripción que da el de la regla
// de los votos con la única clase del juez de las descripciones: si queda
// marcada, sus votos, sin el número, y el motivo de la descripción sin juzgar;
// y, del primer voto, el último que lleva su número, que es el que cuenta, el
// motivo y, si queda marcada, la frase.
func juicioDeLaDescripcion(juicio JuicioDeRespuesta) JuicioDeLaDescripcion {
	clase := juicio.Clases[0]

	deLaDescripcion := JuicioDeLaDescripcion{
		Marcada:   clase.Marcada,
		Votos:     make([]VotoDeLaDescripcion, 0, len(clase.Votos)),
		SinJuzgar: juicio.SinJuzgar,
	}

	var primero VotoDeClase

	for _, voto := range clase.Votos {
		deLaDescripcion.Votos = append(deLaDescripcion.Votos,
			VotoDeLaDescripcion{Respuesta: voto.Respuesta, Frase: voto.Frase, Motivo: voto.Motivo})

		if voto.Voto == 1 {
			primero = voto
		}
	}

	deLaDescripcion.Motivo = primero.Motivo

	if clase.Marcada {
		deLaDescripcion.Frase = primero.Frase
	}

	return deLaDescripcion
}

// huellaDelJuicio es la huella del juicio de la descripción de una skill
// (contracts/juez-de-las-descripciones.md §2 de H31; data-model §4 de H31;
// research D4 de H31): el SHA-256, en hexadecimal, de un texto con una línea
// por dato, en este orden y cada una terminada en salto de línea, como las de
// la huella de una eval: la skill; el SHA-256 de su descripción, la que da
// LeerFrontmatter; el de cada uno de los cuatro ficheros que lleva el juez; y
// el modelo del juez y la versión de Claude Code de sus votos. Con la misma
// huella, el juicio es de la misma descripción de la misma skill con el mismo
// juez.
func huellaDelJuicio(skill, descripcion string, juez *Juez, modelo, version string) string {
	lineas := []string{lineaDe(campoSkill, skill), lineaDe(campoDescripcion, huellaSHA256([]byte(descripcion)))}

	for posicion, contenido := range contenidosDelJuez(juez) {
		lineas = append(lineas, lineaDe(ficherosUsadosPorElJuez[posicion].campo, huellaSHA256([]byte(contenido))))
	}

	lineas = append(lineas, lineaDe(campoJuez, modelo, version))

	return huellaSHA256([]byte(strings.Join(lineas, "\n") + "\n"))
}

// Lo que el trabajo dice de la descripción de su skill además de su juicio
// (contracts/juez-de-las-descripciones.md §3 y §6 de H31; research D7 y D8 de
// H31): el prefijo de cada línea de la comprobación de la medida del juez de
// las descripciones en el informe del instrumento sin medir, que las distingue
// de las del juez de las respuestas; y el motivo de la raíz de la descripción
// que quedó sin juzgar, con la skill y el motivo del voto, que no es el de
// ningún umbral (FR-032).
const (
	prefijoDelJuezDeLasDescripciones = "juez de las descripciones: "
	motivoDeLaDescripcionSinJuzgar   = "el juez de las descripciones dejó sin juzgar la descripción de %s: %s"
)

// DescripcionInformada es la descripción de la skill del trabajo como la
// presenta el informe, el bloque descripcion de informe.json
// (contracts/juez-de-las-descripciones.md §6 de H31; data-model §5 de H31;
// research D7 de H31; FR-031, FR-032): juzgada en la ejecución, con su juicio y
// sus votos, o reutilizada de una medida de main, con su procedencia y el
// juicio de allí. Lo leen quien revisa la propuesta y la búsqueda de las medidas
// de main de una ejecución posterior. Sus claves JSON, en el orden de sus
// campos, son las del bloque.
type DescripcionInformada struct {
	// Skill es la skill de la descripción; Clase, la única del juez de las
	// descripciones; y Huella, la del juicio que da la decisión
	// (huellaDelJuicio).
	Skill  string `json:"skill"`
	Clase  string `json:"clase"`
	Huella string `json:"huella"`

	// Juzgada dice si se juzgó en esta ejecución. Si no, Procedencia es la de la
	// medida de main en que se juzgó; nil, y ausente, en la juzgada.
	Juzgada     bool                   `json:"juzgada"`
	Procedencia *procedenciaDeLaMedida `json:"procedencia,omitzero"`

	// Marcada, Frase y Motivo son los del juicio, el de esta ejecución o el de
	// la medida de main: si quedó marcada y la frase y el motivo de su primer
	// voto (JuicioDeLaDescripcion).
	Marcada bool   `json:"marcada"`
	Frase   string `json:"frase"`
	Motivo  string `json:"motivo"`

	// Votos son los que se dieron en esta ejecución, en su orden, siempre una
	// lista en la juzgada; nil, y ausentes, en la reutilizada, que no pidió
	// ninguno.
	Votos []VotoDeLaDescripcion `json:"votos,omitzero"`

	// SinJuzgar es el motivo del voto que no llegó a darse, con su número, en la
	// juzgada que quedó sin juzgar; vacío, y ausente, si no.
	SinJuzgar string `json:"sin_juzgar,omitzero"`
}

// motivoSinJuzgar es el motivo de la raíz de la descripción que quedó sin
// juzgar (contracts/juez-de-las-descripciones.md §6 de H31; FR-032): la skill y
// el motivo del voto, sin citar ningún umbral, y ninguno sin descripción o con
// la que tiene juicio.
func (d *DescripcionInformada) motivoSinJuzgar() []string {
	if d == nil || d.SinJuzgar == "" {
		return nil
	}

	return []string{fmt.Sprintf(motivoDeLaDescripcionSinJuzgar, d.Skill, d.SinJuzgar)}
}

// juezDeLasDescripcionesSinVotante es el error de los votos del juez de las
// descripciones que no tienen quien los dé, en el trabajo y en la medida.
const juezDeLasDescripcionesSinVotante = "el juez de las descripciones no tiene votante"

// descripcionDelTrabajo es la descripción de la skill que un trabajo juzga, o
// cuyo juicio reutiliza, con lo que necesita para publicarla
// (contracts/juez-de-las-descripciones.md §3 de H31): el bloque de la
// decisión; el juez de las descripciones, con cuya medida se comprueba antes de
// abrir ninguna sesión que el instrumento está medido; y, con la juzgada, la
// votación preparada y la descripción de su SKILL.md.
type descripcionDelTrabajo struct {
	deLaDecision descripcionDeLaDecision
	juez         *Juez
	votacion     *votacion
	descripcion  string
}

// leerElJuezDeLaDescripcion lee de la raíz del repositorio el juez de las
// descripciones de la descripción del bloque de la decisión. El error es el del
// juez que no se puede leer.
func leerElJuezDeLaDescripcion(raiz string, deLaDecision descripcionDeLaDecision) (*descripcionDelTrabajo, error) {
	juez, err := juezDeLasDescripciones(raiz)
	if err != nil {
		return nil, err
	}

	return &descripcionDelTrabajo{deLaDecision: deLaDecision, juez: juez}, nil
}

// sinMedir son las líneas de la comprobación de la medida del juez de las
// descripciones (comprobarLaMedida) con el modelo y la versión fijados para el
// juez, cada una con su prefijo; ninguna sin descripción que publicar o con la
// medida que corresponde y se cumple (FR-026).
func (d *descripcionDelTrabajo) sinMedir(modelo, version string) []string {
	if d == nil {
		return nil
	}

	lineas := comprobarLaMedida(d.juez, modelo, version)
	for posicion, linea := range lineas {
		lineas[posicion] = prefijoDelJuezDeLasDescripciones + linea
	}

	return lineas
}

// prepararLaVotacion prepara, con la descripción juzgada, lo que sus votos
// necesitan: lee de la raíz la descripción del SKILL.md de su skill, la que da
// LeerFrontmatter, y prepara la votación del juez de las descripciones con el
// votante que da quien lo da. Con la reutilizada, sin descripción o sin
// bloque, no hace nada: su juicio es el del bloque. El error es el del SKILL.md
// que no se puede leer, el de la reutilizada sin juicio, el de quien no está y
// el de la votación que no se puede preparar.
func (d *descripcionDelTrabajo) prepararLaVotacion(raiz string, votante func(juez *Juez) Votante) error {
	switch {
	case d == nil:
		return nil
	case !d.deLaDecision.Juzgada && d.deLaDecision.Juicio == nil:
		return fmt.Errorf("la decisión reutiliza la descripción de %s sin su juicio", d.deLaDecision.Skill)
	case !d.deLaDecision.Juzgada:
		return nil
	case votante == nil:
		return errors.New(juezDeLasDescripcionesSinVotante)
	}

	descripcion, err := descripcionDeLaSkill(path.Join(directorioDeLasSkills, d.deLaDecision.Skill, ficheroDeLaSkill),
		func(ruta string) ([]byte, error) { return leerFichero(filepath.Join(raiz, filepath.FromSlash(ruta))) })
	if err != nil {
		return err
	}

	votacion, err := nuevaVotacion(d.juez, votante(d.juez))
	if err != nil {
		return err
	}

	d.descripcion, d.votacion = descripcion, votacion

	return nil
}

// informada es la descripción del trabajo como la presenta el informe
// (contracts/juez-de-las-descripciones.md §3 y §6 de H31; research D7 de H31):
// la juzgada, con el juicio de sus votos, que pide aquí, después de las
// sesiones de las evals, con la votación preparada (juzgarLaDescripcion); y la
// reutilizada, con el juicio y la procedencia del bloque, sin pedir ningún
// voto. nil sin descripción que publicar.
func (d *descripcionDelTrabajo) informada() *DescripcionInformada {
	if d == nil {
		return nil
	}

	informada := DescripcionInformada{
		Skill: d.deLaDecision.Skill, Clase: claseDeLasDescripciones, Huella: d.deLaDecision.Huella,
		Juzgada: d.deLaDecision.Juzgada,
	}

	if !d.deLaDecision.Juzgada {
		informada.Procedencia = d.deLaDecision.Procedencia
		informada.Marcada = d.deLaDecision.Juicio.Marcada
		informada.Frase = d.deLaDecision.Juicio.Frase
		informada.Motivo = d.deLaDecision.Juicio.Motivo

		return &informada
	}

	juicio := d.votacion.juzgarLaDescripcion(d.deLaDecision.Skill, d.descripcion)
	informada.Marcada, informada.Frase, informada.Motivo = juicio.Marcada, juicio.Frase, juicio.Motivo
	informada.Votos, informada.SinJuzgar = juicio.Votos, juicio.SinJuzgar

	return &informada
}

// entradaDeLasDescripciones es la entrada de la matriz del trabajo medida de la
// definición del job que ejecuta la medida del juez de las descripciones, y el
// argumento con el que la ejecuta scripts/evals-medir-juez.sh
// (contracts/juez-de-las-descripciones.md §7 de H31; research D9 de H31;
// FR-034): las demás entradas son skills con juez.
const entradaDeLasDescripciones = "descripciones"

// casoDeLaDescripcion es un caso etiquetado del juez de las descripciones, de
// casos.yaml de la evidencia del ADR 0046: su id, que lo nombra en las líneas
// de la medida; la skill y la descripción con las que se pide cada voto; y su
// etiqueta, etiquetaDefecto o etiquetaCorrecto. Las demás claves de un caso
// —origen, clase y forma, que dicen cómo se construyó— no se leen.
type casoDeLaDescripcion struct {
	ID          string `yaml:"id"`
	Skill       string `yaml:"skill"`
	Etiqueta    string `yaml:"etiqueta"`
	Descripcion string `yaml:"descripcion"`
}

// casosDeLasDescripciones es lo que se lee de casos.yaml de la evidencia del
// juez de las descripciones: sus casos, en el orden del fichero.
type casosDeLasDescripciones struct {
	Casos []casoDeLaDescripcion `yaml:"casos"`
}

// casosDeLasDescripcionesMalFormados es el error de los casos etiquetados del
// juez de las descripciones que no tienen su forma; sigue con el de
// leerLosCasosDeLasDescripciones.
const casosDeLasDescripcionesMalFormados = "los casos etiquetados del juez de las descripciones (%s): %w"

// leerLosCasosDeLasDescripciones lee los casos etiquetados del juez de las
// descripciones de su contenido, el de casos.yaml de la evidencia, con el lector
// común de documentos YAML de internal/skills —una clave repetida es un defecto,
// nunca la última que gana— y, como leerCasosEtiquetados, sin esquema
// publicado: de un caso solo exige que su etiqueta sea una de las dos, porque
// con otra no se podría contar. El error nombra el fichero de la evidencia.
func leerLosCasosDeLasDescripciones(contenido []byte) ([]casoDeLaDescripcion, error) {
	leidos, err := skills.ValidarDocumentoYAML[casosDeLasDescripciones](contenido, nil)
	if err != nil {
		return nil, fmt.Errorf(casosDeLasDescripcionesMalFormados, enLaEvidencia(ficheroDeCasosDelJuez), err)
	}

	for posicion, caso := range leidos.Casos {
		if caso.Etiqueta != etiquetaDefecto && caso.Etiqueta != etiquetaCorrecto {
			return nil, fmt.Errorf(casosDeLasDescripcionesMalFormados, enLaEvidencia(ficheroDeCasosDelJuez),
				fmt.Errorf("el caso %d (%s) tiene la etiqueta %q, que no es %s ni %s", posicion+1, caso.ID,
					caso.Etiqueta, etiquetaDefecto, etiquetaCorrecto))
		}
	}

	return leidos.Casos, nil
}

// leerLosCasosDeLaEvidencia lee los casos etiquetados del juez de las
// descripciones de la evidencia de esa raíz del repositorio: con ellos cuenta
// el peor caso de la entrada descripciones del trabajo medida (research D9 y
// D10 de H31; SC-015). El error es el del fichero que no se puede leer o el de
// unos casos que no tienen su forma.
func leerLosCasosDeLaEvidencia(raiz string) ([]casoDeLaDescripcion, error) {
	contenido, err := leerDeLaEvidencia(raiz, ficheroDeCasosDelJuez)
	if err != nil {
		return nil, err
	}

	return leerLosCasosDeLasDescripciones(contenido)
}

// origenDeLaMedidaDeLasDescripciones es el origen de la medida que da
// medirAlJuezDeLasDescripciones, que sigue con el commit
// (contracts/juez-de-las-descripciones.md §7 de H31).
const origenDeLaMedidaDeLasDescripciones = "ejecución de la medida del juez de las descripciones del job de evals " +
	"sobre %s"

// medirAlJuezDeLasDescripciones ejecuta la medida del juez de las descripciones
// (contracts/juez-de-las-descripciones.md §7 de H31; research D9 de H31; FR-034,
// SC-011): vota sus casos etiquetados con el código del job y devuelve el texto
// de la medida de lo que hay, lista para que una persona la versione. Es la de
// medirAlJuez con otros casos, que no se reconstruyen de ningún informe: cada
// uno lleva su descripción.
//
//  1. No comprueba la medida versionada ni la lee para decidir nada: vota
//     corresponda o no a lo que hay.
//  2. Lee los casos del juez, los de casos.yaml de la evidencia del ADR 0046
//     (leerLosCasosDeLasDescripciones).
//  3. Vota cada caso con la regla de los votos de la descripción del trabajo de
//     una skill (votarLaDescripcion): el mensaje de su descripción, con su
//     skill, y la votación del juez de las evals con la clase
//     claseDeLasDescripciones, por orden; como mucho Concurrencia casos a la
//     vez, tres votos por cada defecto que se marca y uno por cada correcto que
//     no.
//  4. Devuelve la medida (medidaDada), sin skill, con las huellas de la rúbrica
//     y de los casos del juez, el modelo y la versión recibidos y sus dos
//     recuentos, que son las claves que lee leerMedidaDelJuez, y el origen.
//  5. Si un defecto no queda marcado o un correcto queda marcado, devuelve
//     además, con la medida, un error con una línea por caso, en el orden de
//     los casos: «<id>: etiquetado <etiqueta> y <marcado | sin marcar>» y, si
//     alguno de sus votos dijo sí con su frase, «: «<frase>» · …».
//  6. Si algún caso queda sin juzgar, devuelve solo un error, con una línea
//     «<id>: sin juzgar: <motivo>» por cada uno: una medida con casos sin
//     juzgar no es una medida.
//
// Es un error, sin ningún voto ni ninguna medida, que la medición no tenga juez,
// votante o al menos un caso a la vez, que el esquema del juez no sirva para
// validar sus votos, que el juez no tenga la clase de las descripciones que
// decide o que los casos no tengan su forma. La Skill de la medición no se usa.
//
// No escribe en la salida estándar ni en el repositorio: el texto lo escribe
// quien la llama.
func medirAlJuezDeLasDescripciones(medicion MedicionDelJuez) (string, error) {
	switch {
	case medicion.Juez == nil:
		return "", errors.New("no hay juez de las descripciones que medir")
	case medicion.Votar == nil:
		return "", errors.New(juezDeLasDescripcionesSinVotante)
	case medicion.Concurrencia < 1:
		return "", fmt.Errorf(casosALaVezQueNoSirven, medicion.Concurrencia)
	}

	votacion, err := nuevaVotacion(medicion.Juez, medicion.Votar)
	if err != nil {
		return "", err
	}

	deLaClase := slices.IndexFunc(medicion.Juez.Clases, func(clase ClaseDelJuez) bool {
		return clase.Decide && clase.Nombre == claseDeLasDescripciones
	})
	if deLaClase < 0 {
		return "", fmt.Errorf("el juez no tiene la clase %s, que decide", claseDeLasDescripciones)
	}

	casos, err := leerLosCasosDeLasDescripciones([]byte(medicion.Juez.Casos))
	if err != nil {
		return "", err
	}

	var recuento recuentoDeLaMedida

	for posicion, juicio := range votacion.votarLosCasos(casos, medicion.Concurrencia) {
		recuento.contarElCaso(casos[posicion].ID, casos[posicion].Etiqueta, juicio, deLaClase)
	}

	return medicion.resultadoDeLaMedida(claseDeLasDescripciones,
		fmt.Sprintf(origenDeLaMedidaDeLasDescripciones, medicion.Commit), recuento)
}

// votarLosCasos vota la descripción de cada caso de la medida del juez de las
// descripciones (votarLaDescripcion), como mucho concurrencia casos a la vez, y
// devuelve sus juicios en su mismo orden, como juzgarTodas con las respuestas:
// los votos de un mismo caso van uno detrás de otro, cada caso se vota en su
// gorrutina, un canal con concurrencia huecos es el semáforo y un
// sync.WaitGroup espera a los que están en curso. La concurrencia es al menos
// 1, y el votante se puede llamar desde varias gorrutinas a la vez.
func (v *votacion) votarLosCasos(casos []casoDeLaDescripcion, concurrencia int) []JuicioDeRespuesta {
	juicios := make([]JuicioDeRespuesta, len(casos))
	huecos := make(chan struct{}, concurrencia)

	var enCurso sync.WaitGroup

	for posicion, caso := range casos {
		huecos <- struct{}{}

		enCurso.Go(func() {
			defer func() { <-huecos }()

			// Cada gorrutina escribe solo en la posición de su caso.
			juicios[posicion] = v.votarLaDescripcion(caso.Skill, caso.Descripcion)
		})
	}

	enCurso.Wait()

	return juicios
}
