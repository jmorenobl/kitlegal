package skills

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// raizDelRepositorio es la del árbol de quien ejecuta los tests, relativa al
	// directorio de este paquete. Solo se lee de ella (research D5 de H41).
	raizDelRepositorio = "../.."

	// controlesDelLaboratorio son los SKILL.md de las variantes de control del
	// laboratorio, relativos a la raíz: laboratorio/<skill>/<skill>-control/
	// skills/<skill>/SKILL.md (contracts/remisiones.md §1 de H41).
	controlesDelLaboratorio = "laboratorio/*/*-control/skills/*/" + ficheroDeLaSkill
	sufijoDelControl        = "-control"

	// destinoDeUnaReferencia abre el destino de un enlace del cuerpo a una
	// referencia: references/<f> o references/<f>#….
	destinoDeUnaReferencia = carpetaDeLasReferencias + "/"
)

// Las reglas de una remisión, como las nombran sus fallos
// (contracts/remisiones.md §6.3 de H41; FR-001 a FR-003).
const (
	remisionSinElVerbo = "un enlace suyo no va en una frase que lo mande leer («lee» antes del enlace)"
	remisionConLa      = "la frase de un enlace suyo lleva «%s»"
	referenciaSinOrden = "ningún enlace del cuerpo la manda leer"
)

// expresionesFueraDeLasRemisiones son las que la frase de un enlace a una
// referencia no lleva: la condición que pone el agente (FR-002) y la lectura
// antes de escribir o dar la respuesta (FR-003). Se buscan en minúscula, con los
// blancos de la frase juntados.
var expresionesFueraDeLasRemisiones = []string{
	"si la tienes",
	"antes de escribir la respuesta",
	"antes de dar la consulta",
}

// verboDeLeer es la palabra «lee» o «Lee», entera: ni «leer» ni «léelo».
var verboDeLeer = regexp.MustCompile(`(?:^|[^\p{L}\p{N}])[Ll]ee(?:[^\p{L}\p{N}]|$)`)

// fallosDeControl son los que la regla da con el SKILL.md de cada <skill>-control,
// que son las remisiones de v0.9.0 (contracts/remisiones.md §6.4 de H41): así se
// ve fallar por encima del umbral.
var fallosDeControl = map[string][]string{
	"normativa": {
		"normativa: references/buscar_la_norma.md: " + remisionSinElVerbo,
		"normativa: references/como_se_cita.md: " + remisionSinElVerbo,
		"normativa: references/con_herramientas.md: " + remisionSinElVerbo,
		"normativa: references/con_herramientas.md: la frase de un enlace suyo lleva «si la tienes»",
		"normativa: references/lecturas_que_fallan.md: " + remisionSinElVerbo,
		"normativa: references/normas.md: " + remisionSinElVerbo,
		"normativa: references/redaccion_en_una_fecha.md: " + remisionSinElVerbo,
		"normativa: references/varios_bloques.md: " + remisionSinElVerbo,
	},
	"plazos-y-territorio": {
		"plazos-y-territorio: references/calculo_de_plazos.md: " + remisionSinElVerbo,
		"plazos-y-territorio: references/jerarquia_normativa.md: " + remisionSinElVerbo,
		"plazos-y-territorio: references/municipios.md: " + remisionSinElVerbo,
	},
	"jurisprudencia": {
		"jurisprudencia: references/consulta_preparada.md: la frase de un enlace suyo lleva «antes de dar la consulta»",
	},
}

// TestRemisionesDelRepositorio es el control de FR-001 a FR-003 de H41 (SC-001;
// contracts/remisiones.md §6; research D5 y D6): en el cuerpo del SKILL.md de cada
// skill de skills/ que tiene references/, cada enlace a una referencia va en una
// frase con «lee» o «Lee» antes del enlace y sin ninguna de las expresiones que
// no van en una remisión, y cada referencia tiene algún enlace. El subtest
// «control» ve fallar la regla con las remisiones de cada <skill>-control, y el
// de «frases», con lo que un cuerpo tiene: un bloque de código y una expresión
// partida en dos líneas.
func TestRemisionesDelRepositorio(t *testing.T) {
	t.Parallel()

	repositorio := os.DirFS(raizDelRepositorio)

	conReferencias, err := skillsConReferencias(repositorio)
	require.NoError(t, err)
	require.NotEmpty(t, conReferencias, "hay skills con references/ en %s", carpetaDeLasSkills)

	var fallos []string

	for _, skill := range conReferencias {
		deLaSkill, err := fallosDeLaSkill(repositorio, path.Join(carpetaDeLasSkills, skill), skill)
		require.NoError(t, err)

		fallos = append(fallos, deLaSkill...)
	}

	assert.Empty(t, strings.Join(fallos, "\n"), "cada referencia de las skills de %s tiene en el cuerpo de su %s "+
		"una orden de leerla, y cada enlace suyo va en una", carpetaDeLasSkills, ficheroDeLaSkill)

	t.Run("control", func(t *testing.T) {
		t.Parallel()

		controles, err := fs.Glob(repositorio, controlesDelLaboratorio)
		require.NoError(t, err)

		obtenidos := map[string][]string{}

		for _, control := range controles {
			directorio := path.Dir(control)

			skill, esControl := skillDelControl(directorio)
			if !esControl {
				continue
			}

			obtenidos[skill], err = fallosDeLaSkill(repositorio, directorio, skill)
			require.NoError(t, err)
		}

		assert.Equal(t, fallosDeControl, obtenidos,
			"la regla da exactamente estos fallos con las remisiones de cada <skill>-control")
	})

	t.Run("frases", func(t *testing.T) {
		t.Parallel()

		casos := []struct {
			nombre string
			cuerpo string
			fallos []string
		}{
			{
				nombre: "una orden de leer antes de un bloque de código no vale para un enlace escrito después",
				cuerpo: "# Alfa\n\nAntes de la primera orden, lee\n\n```text\nkitlegal alfa --json\n```\n" +
					"[uno](references/uno.md): lo que trae.\n",
				fallos: []string{"alfa: references/uno.md: " + remisionSinElVerbo},
			},
			{
				nombre: "«si la» y «tienes» en dos líneas de la frase de un enlace",
				cuerpo: "# Alfa\n\nCon la herramienta, si la\ntienes, lee [uno](references/uno.md): lo que trae.\n",
				fallos: []string{"alfa: references/uno.md: la frase de un enlace suyo lleva «si la tienes»"},
			},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, caso.fallos, fallosDeLasRemisiones("alfa", []byte(caso.cuerpo), []string{"uno.md"}))
			})
		}
	})
}

// skillsConReferencias son los nombres de las skills de skills/ que tienen
// references/, en orden de nombre.
func skillsConReferencias(repositorio fs.FS) ([]string, error) {
	entradas, err := fs.ReadDir(repositorio, carpetaDeLasSkills)
	if err != nil {
		return nil, err
	}

	var nombres []string

	for _, entrada := range entradas {
		if !entrada.IsDir() {
			continue
		}

		referencias, err := fs.Stat(repositorio, path.Join(carpetaDeLasSkills, entrada.Name(), carpetaDeLasReferencias))

		switch {
		case errors.Is(err, fs.ErrNotExist):
			continue
		case err != nil:
			return nil, err
		case referencias.IsDir():
			nombres = append(nombres, entrada.Name())
		}
	}

	return nombres, nil
}

// skillDelControl es la skill de un directorio laboratorio/<skill>/<skill>-control/
// skills/<skill>, y si lo es.
func skillDelControl(directorio string) (string, bool) {
	partes := strings.Split(directorio, "/")
	if len(partes) != 5 || partes[3] != carpetaDeLasSkills {
		return "", false
	}

	skill := partes[1]

	return skill, partes[2] == skill+sufijoDelControl && partes[4] == skill
}

// fallosDeLaSkill son los de la regla sobre la skill del directorio dado: el
// cuerpo de su SKILL.md y los ficheros de su references/.
func fallosDeLaSkill(repositorio fs.FS, directorio, skill string) ([]string, error) {
	contenido, err := fs.ReadFile(repositorio, path.Join(directorio, ficheroDeLaSkill))
	if err != nil {
		return nil, err
	}

	cuerpo, conCuerpo := cuerpoDeLaSkill(contenido)
	if !conCuerpo {
		return nil, fmt.Errorf("%s: su %s no cierra el frontmatter", directorio, ficheroDeLaSkill)
	}

	entradas, err := fs.ReadDir(repositorio, path.Join(directorio, carpetaDeLasReferencias))
	if err != nil {
		return nil, err
	}

	var referencias []string

	for _, entrada := range entradas {
		if !entrada.IsDir() {
			referencias = append(referencias, entrada.Name())
		}
	}

	return fallosDeLasRemisiones(skill, cuerpo, referencias), nil
}

// incumplidasDeUnaReferencia son las reglas que incumplen los enlaces del cuerpo
// a una referencia, y si tiene alguno.
type incumplidasDeUnaReferencia struct {
	enlazada    bool
	sinElVerbo  bool
	expresiones map[string]bool
}

// fallosDeLasRemisiones son los de la regla de las remisiones sobre un cuerpo
// con las referencias dadas (contracts/remisiones.md §6.3 de H41): uno por
// referencia y regla, «<skill>: references/<f>: <regla>», en el orden de los
// nombres de las referencias y, en cada una, en el de las reglas.
func fallosDeLasRemisiones(skill string, cuerpo []byte, referencias []string) []string {
	incumplidas := map[string]*incumplidasDeUnaReferencia{}

	deLaReferencia := func(nombre string) *incumplidasDeUnaReferencia {
		if incumplidas[nombre] == nil {
			incumplidas[nombre] = &incumplidasDeUnaReferencia{expresiones: map[string]bool{}}
		}

		return incumplidas[nombre]
	}

	for _, nombre := range referencias {
		deLaReferencia(nombre)
	}

	lineas := lineasDelCuerpo(cuerpo)

	for numero, linea := range lineas {
		for _, enlace := range enlaceDeMarkdown.FindAllStringSubmatchIndex(linea, -1) {
			destino, aUnaReferencia := strings.CutPrefix(linea[enlace[2]:enlace[3]], destinoDeUnaReferencia)
			nombre, _, _ := strings.Cut(destino, "#")

			if !aUnaReferencia || nombre == "" {
				continue
			}

			antes, frase := fraseDelEnlace(lineas, numero, enlace[0], enlace[1])

			referencia := deLaReferencia(nombre)
			referencia.enlazada = true
			referencia.sinElVerbo = referencia.sinElVerbo || !verboDeLeer.MatchString(antes)

			juntada := strings.ToLower(strings.Join(strings.Fields(frase), " "))
			for _, expresion := range expresionesFueraDeLasRemisiones {
				referencia.expresiones[expresion] = referencia.expresiones[expresion] ||
					strings.Contains(juntada, expresion)
			}
		}
	}

	var fallos []string

	for _, nombre := range slices.Sorted(maps.Keys(incumplidas)) {
		fallos = append(fallos, fallosDeUnaReferencia(skill, nombre, incumplidas[nombre])...)
	}

	return fallos
}

// fallosDeUnaReferencia son los de una referencia, en el orden de las reglas: el
// verbo, cada expresión en el de su lista, y la referencia sin ningún enlace.
func fallosDeUnaReferencia(skill, nombre string, incumplidas *incumplidasDeUnaReferencia) []string {
	prefijo := skill + ": " + destinoDeUnaReferencia + nombre + ": "

	if !incumplidas.enlazada {
		return []string{prefijo + referenciaSinOrden}
	}

	var fallos []string

	if incumplidas.sinElVerbo {
		fallos = append(fallos, prefijo+remisionSinElVerbo)
	}

	for _, expresion := range expresionesFueraDeLasRemisiones {
		if incumplidas.expresiones[expresion] {
			fallos = append(fallos, prefijo+fmt.Sprintf(remisionConLa, expresion))
		}
	}

	return fallos
}

// lineasDelCuerpo son las líneas del cuerpo, sin su salto, con las de sus bloques
// de código, vallas incluidas, en blanco (contracts/remisiones.md §6.1 de H41;
// research D6): una línea está fuera de los bloques si lineasFueraDeBloques,
// leído el cuerpo hasta ella, da una línea más que leído hasta la anterior.
func lineasDelCuerpo(cuerpo []byte) []string {
	var (
		lineas []string
		fuera  int
		leido  int
	)

	for linea := range strings.Lines(string(cuerpo)) {
		leido += len(linea)

		if hasta := len(lineasFueraDeBloques(cuerpo[:leido])); hasta > fuera {
			fuera = hasta
			lineas = append(lineas, strings.TrimSuffix(linea, "\n"))

			continue
		}

		lineas = append(lineas, "")
	}

	return lineas
}

// fraseDelEnlace es la frase del enlace que ocupa en la línea numero de inicio a
// fin (contracts/remisiones.md §6.2 de H41), en dos tramos con sus líneas unidas
// por un salto: de su principio al enlace, y entera, con el enlace.
func fraseDelEnlace(lineas []string, numero, inicio, fin int) (string, string) {
	lineaDelPrincipio, principio := principioDeLaFrase(lineas, numero, inicio)
	lineaDelFinal, final := finalDeLaFrase(lineas, numero, fin)

	return tramoDeLineas(lineas, lineaDelPrincipio, principio, numero, inicio),
		tramoDeLineas(lineas, lineaDelPrincipio, principio, lineaDelFinal, final)
}

// principioDeLaFrase es la línea y la columna en que empieza la frase de lo que
// hay en la columna dada de la línea numero: tras el último límite anterior —un
// punto seguido de un espacio o de un salto de línea, una línea en blanco, o el
// principio de una línea que, tras su sangría, empieza por «- », «* », un número
// y «. », o «#»—, o el principio del cuerpo.
func principioDeLaFrase(lineas []string, numero, columna int) (int, int) {
	tramo := lineas[numero][:columna]

	for {
		if punto := strings.LastIndex(tramo, ". "); punto >= 0 {
			return numero, punto + len(". ")
		}

		if abreUnElemento(lineas[numero]) || numero == 0 {
			return numero, 0
		}

		if anterior := lineas[numero-1]; strings.TrimSpace(anterior) == "" || strings.HasSuffix(anterior, ".") {
			return numero, 0
		}

		numero--
		tramo = lineas[numero]
	}
}

// finalDeLaFrase es la línea y la columna, exclusiva, en que acaba la frase de lo
// que hay hasta la columna dada de la línea numero: en el primer punto posterior
// seguido de un espacio o de un salto de línea, que queda dentro, antes de la
// primera línea en blanco, o en el final del cuerpo.
func finalDeLaFrase(lineas []string, numero, columna int) (int, int) {
	tramo, desplazamiento := lineas[numero][columna:], columna

	for {
		if punto := strings.Index(tramo, ". "); punto >= 0 {
			return numero, desplazamiento + punto + len(".")
		}

		if strings.HasSuffix(tramo, ".") || numero+1 == len(lineas) || strings.TrimSpace(lineas[numero+1]) == "" {
			return numero, len(lineas[numero])
		}

		numero++
		tramo, desplazamiento = lineas[numero], 0
	}
}

// abreUnElemento dice si la línea, tras su sangría, empieza por la marca de una
// viñeta o de un paso: «- », «* », un número y «. » (elementoDeLista), o «#».
func abreUnElemento(linea string) bool {
	return elementoDeLista.MatchString(linea) || strings.HasPrefix(strings.TrimLeft(linea, " \t"), "#")
}

// tramoDeLineas es el texto de las líneas desde la columna desde de la línea
// primera hasta la columna, exclusiva, hasta de la línea ultima, con sus líneas
// unidas por un salto.
func tramoDeLineas(lineas []string, primera, desde, ultima, hasta int) string {
	if primera == ultima {
		return lineas[primera][desde:hasta]
	}

	tramos := slices.Concat([]string{lineas[primera][desde:]}, lineas[primera+1:ultima],
		[]string{lineas[ultima][:hasta]})

	return strings.Join(tramos, "\n")
}
