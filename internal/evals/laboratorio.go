package evals

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/jmorenobl/kitlegal/internal/skills"
)

// Lo que forma el laboratorio de H39 en el repositorio, desde su raíz
// (contracts/laboratorio.md §1 de H39; research D4 de H39): un directorio de
// primer nivel, fuera de skills/, que el binario empotra, y de evals/, que lee el
// job de evals, con la declaración de las variantes y un directorio por variante.
const (
	carpetaDelLaboratorio = "laboratorio"
	ficheroDeLasVariantes = "variantes.yaml"
	carpetaDeLasVariantes = "variantes"
)

// Lo que hay en el directorio de una variante (data-model §3 de H39): las skills
// que se instalan tal cual, cada una en su directorio de skills/, con su SKILL.md y
// sus referencias, y los ficheros con el texto que la sesión anuncia en lugar del
// de kitlegal: el de las instrucciones del servidor y el de la descripción del
// parámetro de la herramienta.
const (
	carpetaDeLasSkillsDeLaVariante   = "skills"
	ficheroDeLaSkillDeLaVariante     = "SKILL.md"
	carpetaDeLasReferenciasDeLaSkill = "references"
	ficheroDeLasInstrucciones        = "instrucciones.txt"
	ficheroDelParametro              = "parametro.txt"
)

// parteDeUnaVariante es una de las partes en que se compara una variante del
// laboratorio con la de control (data-model §3 de H39; FR-028 de H39): las que la
// variante declara en cambia difieren y las demás son las de control, byte a
// byte.
type parteDeUnaVariante string

// Las partes de una variante, el vocabulario cerrado de cambia (data-model §3 de
// H39).
const (
	// parteDescripcion son las líneas de la clave description del frontmatter del
	// SKILL.md de la skill de laboratorio: la de la clave y las sangradas que la
	// siguen.
	parteDescripcion parteDeUnaVariante = "descripcion"
	// parteCuerpo es el resto de ese SKILL.md: el frontmatter sin esas líneas y
	// todo lo que sigue a la línea que lo cierra.
	parteCuerpo parteDeUnaVariante = "cuerpo"
	// parteReferencia son los ficheros de references/ de la skill de laboratorio.
	parteReferencia parteDeUnaVariante = "referencia"
	// parteEntrada es el directorio de la skill de entrada, con sus ficheros.
	parteEntrada parteDeUnaVariante = "entrada"
	// parteInstrucciones es el fichero de las instrucciones del servidor.
	parteInstrucciones parteDeUnaVariante = "instrucciones"
	// parteParametro es el fichero de la descripción del parámetro.
	parteParametro parteDeUnaVariante = "parametro"
)

// partesDeLasVariantes son las seis partes de una variante, en el orden de
// data-model §3 de H39.
var partesDeLasVariantes = []parteDeUnaVariante{
	parteDescripcion, parteCuerpo, parteReferencia, parteEntrada, parteInstrucciones, parteParametro,
}

// declaracionDelLaboratorio es la declaración de las variantes del laboratorio,
// laboratorio/variantes.yaml (data-model §2 de H39; contracts/laboratorio.md §2
// de H39).
type declaracionDelLaboratorio struct {
	// Skill es el nombre de la skill de laboratorio, la que tiene cada variante.
	Skill string `yaml:"skill"`
	// Entrada es el nombre de la skill de entrada, la que remite a la de
	// laboratorio en la variante que la cambia.
	Entrada string `yaml:"entrada"`
	// Herramienta es la herramienta de kitlegal que piden las evals del
	// laboratorio, `<applet>_<verbo>`.
	Herramienta string `yaml:"herramienta"`
	// Parametro es el nombre de la propiedad de la entrada de Herramienta cuya
	// descripción cambia la variante que cambia el parámetro.
	Parametro string `yaml:"parametro"`
	// Referencia es el nombre del único fichero de references/ en las variantes
	// que cambian la referencia.
	Referencia string `yaml:"referencia"`
	// Variantes son las variantes, en el orden de la declaración.
	Variantes []varianteDelLaboratorio `yaml:"variantes"`
}

// varianteDelLaboratorio es una variante de la declaración del laboratorio: su
// nombre, que es el de su directorio; los modos en que se mide; y las partes en
// que difiere de la de control.
type varianteDelLaboratorio struct {
	Nombre string               `yaml:"nombre"`
	Modos  []Modo               `yaml:"modos"`
	Cambia []parteDeUnaVariante `yaml:"cambia"`
}

// esquemaDeLaDeclaracionDelLaboratorio es la forma de la declaración del
// laboratorio (data-model §2 de H39): las seis claves, todas obligatorias y
// ninguna más; cada variante, con sus tres claves y ninguna más; los modos, un
// subconjunto no vacío de los dos, y las partes, uno de las seis.
const esquemaDeLaDeclaracionDelLaboratorio = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "additionalProperties": false,
  "required": ["skill", "entrada", "herramienta", "parametro", "referencia", "variantes"],
  "properties": {
    "skill": {"type": "string", "minLength": 1},
    "entrada": {"type": "string", "minLength": 1},
    "herramienta": {"type": "string", "minLength": 1},
    "parametro": {"type": "string", "minLength": 1},
    "referencia": {"type": "string", "minLength": 1},
    "variantes": {
      "type": "array",
      "minItems": 1,
      "items": {
        "type": "object",
        "additionalProperties": false,
        "required": ["nombre", "modos", "cambia"],
        "properties": {
          "nombre": {"type": "string", "minLength": 1},
          "modos": {"type": "array", "minItems": 1, "uniqueItems": true, "items": {"enum": ["orden", "herramienta"]}},
          "cambia": {
            "type": "array",
            "uniqueItems": true,
            "items": {"enum": ["descripcion", "cuerpo", "referencia", "entrada", "instrucciones", "parametro"]}
          }
        }
      }
    }
  }
}`

// leerLaDeclaracionDelLaboratorio lee la declaración del laboratorio de su
// contenido con el lector estricto de documentos YAML de internal/skills: una
// clave repetida o una que la declaración no tiene, en la raíz o en una variante,
// es un defecto, nunca algo que se ignora (data-model §2 de H39).
func leerLaDeclaracionDelLaboratorio(contenido []byte) (declaracionDelLaboratorio, error) {
	esquema, err := skills.CompilarEsquema([]byte(esquemaDeLaDeclaracionDelLaboratorio))
	if err != nil {
		return declaracionDelLaboratorio{}, fmt.Errorf("el esquema de la declaración del laboratorio: %w", err)
	}

	declarada, err := skills.ValidarDocumentoYAML[declaracionDelLaboratorio](contenido, esquema)
	if err != nil {
		return declaracionDelLaboratorio{}, fmt.Errorf("la declaración del laboratorio (%s): %w",
			path.Join(carpetaDelLaboratorio, ficheroDeLasVariantes), err)
	}

	return declarada, nil
}

// leerLaDeclaracionDelRepositorio lee la declaración del laboratorio del
// repositorio de esa raíz: el error es el del fichero que no se puede leer o el de
// una declaración que no tiene su forma.
func leerLaDeclaracionDelRepositorio(raiz string) (declaracionDelLaboratorio, error) {
	// filepath.Clean es lo que el control de rutas reconoce como saneado antes de
	// abrir un fichero (gosec G304); la ruta es la de un fichero fijo del
	// repositorio de esa raíz.
	contenido, err := os.ReadFile(filepath.Clean(filepath.Join(raiz, carpetaDelLaboratorio, ficheroDeLasVariantes)))
	if err != nil {
		return declaracionDelLaboratorio{}, fmt.Errorf("la declaración del laboratorio no se puede leer: %w", err)
	}

	return leerLaDeclaracionDelLaboratorio(contenido)
}

// directorioDeLaVariante es el directorio de la variante de ese nombre en el
// repositorio de esa raíz.
func directorioDeLaVariante(raiz, nombre string) string {
	return filepath.Join(raiz, carpetaDelLaboratorio, carpetaDeLasVariantes, nombre)
}

// familia es la familia del laboratorio de H39 que declara (data-model §1 de
// H40): su skill y sus variantes, con sus modos, en laboratorio/variantes/, y
// los repartos, con todas las evals del laboratorio. Lo demás de H39 —la skill de
// entrada, la herramienta, el parámetro y la referencia— sigue en la
// declaración.
func (d declaracionDelLaboratorio) familia() familiaDelLaboratorio {
	return familiaDelLaboratorio{
		skill:     d.Skill,
		carpeta:   path.Join(carpetaDelLaboratorio, carpetaDeLasVariantes),
		variantes: d.Variantes,
		evals:     path.Join(carpetaDelLaboratorio, carpetaDeLasEvalsDelLaboratorio),
	}
}

// carpetaDeLasEvalsDelLaboratorio es la carpeta de las evals del laboratorio,
// dentro de la suya (contracts/laboratorio.md §1 y §6 de H39).
const carpetaDeLasEvalsDelLaboratorio = "evals"

// directorioDeLasEvalsDelLaboratorio es el directorio de las evals del
// laboratorio en el repositorio de esa raíz (data-model §4 de H39): las mismas
// para todas las variantes, con el formato de las de evals/ y fuera de esa
// carpeta, que es la que lee el job de evals.
func directorioDeLasEvalsDelLaboratorio(raiz string) string {
	return filepath.Join(raiz, carpetaDelLaboratorio, carpetaDeLasEvalsDelLaboratorio)
}

// partesDeUnaVarianteLeidas son las partes de una variante tal como están en su
// directorio (data-model §3 de H39).
type partesDeUnaVarianteLeidas struct {
	// partes son, por parte, sus ficheros: la ruta de cada uno dentro del
	// directorio de la variante, con barras, y sus bytes. Las dos partes del
	// SKILL.md de la skill de laboratorio llevan cada una la ruta de ese fichero
	// y sus bytes de la parte. Una parte sin ningún fichero no está.
	partes map[parteDeUnaVariante]map[string]string
	// fuera son las rutas de los ficheros de la variante que no son de ninguna
	// parte, en orden de ruta.
	fuera []string
}

// leerLasPartesDeLaVariante lee las partes de la variante del directorio con las
// skills de la declaración (data-model §3 de H39): el SKILL.md de la skill de
// laboratorio, partido en su descripción y su cuerpo; cada fichero de su
// references/, en la referencia; cada fichero del directorio de la skill de
// entrada, en la entrada; y los dos ficheros del texto de la sesión, cada uno en
// su parte. Cualquier otro fichero, a cualquier profundidad, está fuera de las
// partes. Un directorio no es un fichero de ninguna parte. El error queda para un
// directorio que no se puede recorrer y un fichero que no se puede leer.
func leerLasPartesDeLaVariante(directorio string, declarada declaracionDelLaboratorio) (
	partesDeUnaVarianteLeidas, error,
) {
	arbol := os.DirFS(directorio)
	leidas := partesDeUnaVarianteLeidas{partes: map[parteDeUnaVariante]map[string]string{}}

	anotar := func(parte parteDeUnaVariante, ruta, contenido string) {
		if leidas.partes[parte] == nil {
			leidas.partes[parte] = map[string]string{}
		}

		leidas.partes[parte][ruta] = contenido
	}

	err := fs.WalkDir(arbol, ".", func(ruta string, entrada fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entrada.IsDir() {
			return nil
		}

		parte, deUnaParte := parteDelFichero(ruta, declarada)
		if !deUnaParte {
			leidas.fuera = append(leidas.fuera, ruta)

			return nil
		}

		contenido, err := fs.ReadFile(arbol, ruta)
		if err != nil {
			return err
		}

		if parte != parteCuerpo {
			anotar(parte, ruta, string(contenido))

			return nil
		}

		descripcion, cuerpo := descripcionYCuerpo(string(contenido))
		if descripcion != "" {
			anotar(parteDescripcion, ruta, descripcion)
		}

		anotar(parteCuerpo, ruta, cuerpo)

		return nil
	})
	if err != nil {
		return partesDeUnaVarianteLeidas{}, fmt.Errorf("la variante de %s no se puede leer: %w", directorio, err)
	}

	return leidas, nil
}

// parteDelFichero es la parte del fichero de esa ruta dentro del directorio de una
// variante, y si es de alguna: el SKILL.md de la skill de laboratorio da
// parteCuerpo, que leerLasPartesDeLaVariante parte en la descripción y el cuerpo.
func parteDelFichero(ruta string, declarada declaracionDelLaboratorio) (parteDeUnaVariante, bool) {
	deLaSkill := path.Join(carpetaDeLasSkillsDeLaVariante, declarada.Skill)
	deLaEntrada := path.Join(carpetaDeLasSkillsDeLaVariante, declarada.Entrada)

	switch {
	case ruta == ficheroDeLasInstrucciones:
		return parteInstrucciones, true
	case ruta == ficheroDelParametro:
		return parteParametro, true
	case ruta == path.Join(deLaSkill, ficheroDeLaSkillDeLaVariante):
		return parteCuerpo, true
	case strings.HasPrefix(ruta, path.Join(deLaSkill, carpetaDeLasReferenciasDeLaSkill)+"/"):
		return parteReferencia, true
	case strings.HasPrefix(ruta, deLaEntrada+"/"):
		return parteEntrada, true
	default:
		return "", false
	}
}

// delimitadorDelFrontmatterDeLaVariante es la línea que abre y cierra el
// frontmatter de un SKILL.md, y claveDeLaDescripcion, el principio de la línea de
// la clave description en él.
const (
	delimitadorDelFrontmatterDeLaVariante = "---"
	claveDeLaDescripcion                  = "description:"
)

// descripcionYCuerpo parte el contenido de un SKILL.md en sus dos partes
// (data-model §3 de H39): la descripción, las líneas de la clave description del
// frontmatter —la que empieza por la clave y las sangradas que la siguen—, y el
// cuerpo, todas las demás, en su orden. El frontmatter son las líneas que siguen a
// la primera, si es el delimitador, hasta la siguiente que lo es. Sin frontmatter,
// o sin la clave en él, la descripción queda vacía y todo es cuerpo: lo que le
// falte a ese SKILL.md lo dice la comprobación de la forma de las skills
// (skills.Regenerar), que no es de este reparto.
func descripcionYCuerpo(contenido string) (string, string) {
	var descripcion, cuerpo strings.Builder

	primera, enElFrontmatter, enLaDescripcion := true, false, false

	for linea := range strings.Lines(contenido) {
		esDelimitador := strings.TrimSuffix(linea, "\n") == delimitadorDelFrontmatterDeLaVariante

		switch {
		case primera:
			primera, enElFrontmatter = false, esDelimitador
		case enElFrontmatter && esDelimitador:
			enElFrontmatter, enLaDescripcion = false, false
		case enElFrontmatter && strings.HasPrefix(linea, claveDeLaDescripcion):
			enLaDescripcion = true
		case enLaDescripcion && !strings.HasPrefix(linea, " ") && !strings.HasPrefix(linea, "\t"):
			enLaDescripcion = false
		}

		if enLaDescripcion {
			descripcion.WriteString(linea)
		} else {
			cuerpo.WriteString(linea)
		}
	}

	return descripcion.String(), cuerpo.String()
}

// lineaDelSello es la línea con la que termina una respuesta que cumple la
// regla testigo de la skill de laboratorio: un código inventado, que un modelo
// no puede saber si no ha leído la regla (FR-013 de H39; contracts/laboratorio.md
// §3 de H39; research D6 de H39). No aparece en ninguna skill, referencia ni
// instrucción de kitlegal.
const lineaDelSello = "Sello de lectura: KL-7Q4"

// parrafoDeLaRegla es el párrafo de la regla testigo, el mismo en todos los
// sitios del laboratorio donde va (contracts/laboratorio.md §3 de H39).
const parrafoDeLaRegla = "La respuesta termina con una última línea, sola y sin ningún formato —ni negrita, " +
	"ni cursiva, ni comillas, ni código—, que dice exactamente: " + lineaDelSello

// bordesDeUnaLinea son lo que el guion de la regla testigo quita a los lados de
// una línea antes de mirarla: los espacios y el retorno de carro de un fin de
// línea CRLF.
const bordesDeUnaLinea = " \r"

// reglaTestigo es el guion de la regla testigo, que decide sin modelo si una
// respuesta la cumple (FR-013 de H39; data-model §5 de H39): la cumple si su
// última línea no vacía, sin espacios ni retorno de carro a los lados, es
// exactamente la línea del sello. No la cumple si el sello no aparece, si
// aparece y no es la última línea no vacía, o si la última lleva el sello con
// formato u otro texto. Una línea que solo tiene espacios o retornos de carro
// cuenta como vacía.
func reglaTestigo(respuesta string) bool {
	lineas := strings.Split(respuesta, "\n")
	for i := len(lineas) - 1; i >= 0; i-- {
		linea := strings.Trim(lineas[i], bordesDeUnaLinea)
		if linea != "" {
			return linea == lineaDelSello
		}
	}

	return false
}
