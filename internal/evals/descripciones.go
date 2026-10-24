package evals

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"slices"
	"strings"
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
