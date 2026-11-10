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
	// directorio de este paquete. Solo se lee de ella.
	raizDelRepositorio = "../.."

	// carpetaDelLaboratorio y sufijoDelControl dan el directorio de la skill de la
	// variante de control de una skill del producto, relativo a la raíz:
	// laboratorio/<skill>/<skill>-control/skills/<skill> (contracts/remisiones.md
	// §1 de H41).
	carpetaDelLaboratorio = "laboratorio"
	sufijoDelControl      = "-control"

	// destinoDeUnaReferencia abre el destino de un enlace del cuerpo a una
	// referencia: references/<f> o references/<f>#….
	destinoDeUnaReferencia = carpetaDeLasReferencias + "/"
)

// Los motivos por los que falla la regla de las remisiones, como los nombran sus
// fallos (contracts/remisiones.md §7.1 de H41; FR-001 a FR-003).
const (
	remisionSinElVerbo = "un enlace suyo no va en una frase que lo mande leer («lee» antes del enlace)"
	remisionConLa      = "la frase de un enlace suyo lleva «%s»"
	referenciaSinOrden = "ningún enlace del cuerpo la manda leer"
)

// expresionesFueraDeLasRemisiones son las que la frase de un enlace a una
// referencia no lleva: la condición que pone el agente (FR-002) y la lectura
// antes de escribir la respuesta o de dar la consulta (FR-003). Se buscan en
// minúscula en la frase, que ya tiene los blancos juntados.
var expresionesFueraDeLasRemisiones = []string{
	"si la tienes",
	"antes de escribir la respuesta",
	"antes de dar la consulta",
}

// verboDeLeer es la palabra «lee» o «Lee», entera: ni «leer» ni «léelo».
var verboDeLeer = regexp.MustCompile(`(?:^|[^\p{L}\p{N}])[Ll]ee(?:[^\p{L}\p{N}]|$)`)

// fallosAdmitidosDelProducto son, por skill de skills/, los que la regla da con
// su SKILL.md y el test admite; una skill que no está aquí no admite ninguno.
var fallosAdmitidosDelProducto = map[string][]string{
	// La viñeta de jerarquia_normativa.md queda como en v0.9.0 (FR-002 de H41.1;
	// ADR 0053, punto 1 y «Lo que cuesta»): ninguna eval llega a su condición.
	"plazos-y-territorio": {
		"plazos-y-territorio: references/jerarquia_normativa.md: " + remisionSinElVerbo,
	},
}

// remisionesDelPlan son, por skill, las frases del cuerpo con un enlace a una
// referencia, fuera de la región de los paquetes, en su orden y como las da
// Frases (contracts/remisiones.md §5 de H41; data-model §7; FR-001; la de
// plazos-y-territorio, contracts/controles.md §2.2 de H41.1).
var remisionesDelPlan = map[string][]string{
	"jurisprudencia": {
		"Antes de pedir una u otra, lee [references/consulta_preparada.md](references/consulta_preparada.md): la " +
			"orden de cada forma, cómo se dan las casillas y qué hacer con el equivalente que devuelva la operación.",
		"Con cada documento, antes del cotejo, lee [references/documento_aportado.md](references/documento_aportado.md): " +
			"qué se pasa del documento y cómo, y qué se dice si no es la pedida, si su ROJ y su ECLI no se corresponden " +
			"o si no lleva ficha.",
		"Ante «qué dice la jurisprudencia sobre…», lee «La búsqueda por materia» en " +
			"[references/consulta_preparada.md](references/consulta_preparada.md), pide `cita_preparar` con los " +
			"términos de la pregunta en `texto`, o su orden, y da la dirección que devuelve.",
	},
	"normativa": {
		"Antes de la primera orden o llamada, lee [con_herramientas](references/con_herramientas.md): cómo se pide " +
			"cada una con herramienta.",
		"Lee [normas](references/normas.md) y toma el `BOE-A-…` de su tabla; si no está, lee " +
			"[buscar_la_norma](references/buscar_la_norma.md) y búscala como dice.",
		"- Con una fecha en la pregunta, lee [redaccion_en_una_fecha](references/redaccion_en_una_fecha.md) antes de " +
			"leer el bloque: cómo se lee.",
		"Por cada entrada `version-obsoleta`, o si se pregunta qué cambió, lee " +
			"[redaccion_modificada](references/redaccion_modificada.md): qué se dice.",
		"- Con avisos o un id que termina en punto, lee [como_se_cita](references/como_se_cita.md): cómo se citan.",
		"- Para varios bloques a la vez, lee antes [varios_bloques](references/varios_bloques.md): su orden.",
		"Si una orden no termina con `0`, lee [lecturas_que_fallan](references/lecturas_que_fallan.md) antes de " +
			"seguir: qué hacer.",
	},
	"plazos-y-territorio": {
		"- [jerarquia_normativa.md](references/jerarquia_normativa.md): para el nivel que regula algo, con el " +
			"autonómico de la `comunidad` de `data`, y la norma que prevalece.",
		"Antes de la primera orden o llamada, lee [municipios.md](references/municipios.md): qué hacer si falta el " +
			"municipio o no sale con 0, qué trae `data`, la plantilla del territorio y qué dice su cobertura.",
		"Antes del 8, lee [calculo_de_plazos.md](references/calculo_de_plazos.md): los municipios que cuentan, el " +
			"plazo sin calcular y la lectura en la fecha de inicio.",
	},
}

// fallosDeControl son los que la regla da con el SKILL.md de cada <skill>-control,
// que lleva las remisiones de v0.9.0 (contracts/remisiones.md §7.3 de H41): así
// se ve fallar por encima del umbral.
var fallosDeControl = map[string][]string{
	"jurisprudencia": {
		"jurisprudencia: references/consulta_preparada.md: la frase de un enlace suyo lleva «antes de dar la consulta»",
	},
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
}

// TestRemisionesDelRepositorio es el control de FR-001 a FR-003 de H41 y de FR-007
// de H41.1 (SC-001; contracts/remisiones.md §7 de H41 y contracts/controles.md §2
// de H41.1; research D5 a D7): por cada skill de skills/ con references/, en orden
// de nombre, (1) la regla de las remisiones —cada enlace del cuerpo de su SKILL.md
// a una referencia va en una frase de Frases con «lee» o «Lee» antes del enlace y
// sin ninguna de las expresiones que no van en una remisión, y cada referencia
// tiene alguno—, sobre todo el cuerpo, da exactamente los fallos que la skill
// admite en fallosAdmitidosDelProducto; y (2) las frases con un enlace a una
// referencia, fuera de la región de los paquetes, son, enteras y en orden, las de
// remisionesDelPlan. El subtest «control» ve fallar (1) con las remisiones de
// normativa-control y jurisprudencia-control.
func TestRemisionesDelRepositorio(t *testing.T) {
	t.Parallel()

	repositorio := os.DirFS(raizDelRepositorio)

	conReferencias, err := skillsConReferencias(repositorio)
	require.NoError(t, err)
	require.NotEmpty(t, conReferencias, "hay skills con %s/ en %s", carpetaDeLasReferencias, carpetaDeLasSkills)

	for _, skill := range conReferencias {
		cuerpo, referencias, err := cuerpoYReferencias(repositorio, path.Join(carpetaDeLasSkills, skill))
		require.NoError(t, err)

		segmentos := Frases(cuerpo)

		assert.Equal(t, fallosAdmitidosDelProducto[skill], fallosDeLaRegla(skill, segmentos, referencias),
			"%s: la regla da exactamente los fallos que admite", skill)

		comprobarLaListaDelPlan(t, skill, segmentos)
	}

	t.Run("control", func(t *testing.T) {
		t.Parallel()

		obtenidos := make(map[string][]string, len(fallosDeControl))

		for _, skill := range slices.Sorted(maps.Keys(fallosDeControl)) {
			directorio := path.Join(carpetaDelLaboratorio, skill, skill+sufijoDelControl, carpetaDeLasSkills, skill)

			cuerpo, referencias, err := cuerpoYReferencias(repositorio, directorio)
			require.NoError(t, err)

			obtenidos[skill] = fallosDeLaRegla(skill, Frases(cuerpo), referencias)
		}

		assert.Equal(t, fallosDeControl, obtenidos,
			"la regla da exactamente estos fallos con las remisiones de cada <skill>-control")
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

// cuerpoYReferencias son el cuerpo del SKILL.md de la skill del directorio dado y
// los nombres de los ficheros de su references/, en orden de nombre.
func cuerpoYReferencias(repositorio fs.FS, directorio string) (string, []string, error) {
	contenido, err := fs.ReadFile(repositorio, path.Join(directorio, ficheroDeLaSkill))
	if err != nil {
		return "", nil, err
	}

	cuerpo, conCuerpo := cuerpoDeLaSkill(contenido)
	if !conCuerpo {
		return "", nil, fmt.Errorf("%s: su %s no cierra el frontmatter", directorio, ficheroDeLaSkill)
	}

	entradas, err := fs.ReadDir(repositorio, path.Join(directorio, carpetaDeLasReferencias))
	if err != nil {
		return "", nil, err
	}

	var referencias []string

	for _, entrada := range entradas {
		if !entrada.IsDir() {
			referencias = append(referencias, entrada.Name())
		}
	}

	return string(cuerpo), referencias, nil
}

// enlaceAUnaReferencia es un enlace de un segmento a una referencia: el nombre
// del fichero, sin lo que sigue a #, y el byte del segmento en que empieza.
type enlaceAUnaReferencia struct {
	referencia string
	inicio     int
}

// enlacesAReferencias son los enlaces del segmento cuyo destino es references/<f>
// o references/<f>#…, en su orden.
func enlacesAReferencias(segmento string) []enlaceAUnaReferencia {
	var enlaces []enlaceAUnaReferencia

	for _, enlace := range enlaceDeMarkdown.FindAllStringSubmatchIndex(segmento, -1) {
		destino, aUnaReferencia := strings.CutPrefix(segmento[enlace[2]:enlace[3]], destinoDeUnaReferencia)
		nombre, _, _ := strings.Cut(destino, "#")

		if aUnaReferencia && nombre != "" {
			enlaces = append(enlaces, enlaceAUnaReferencia{referencia: nombre, inicio: enlace[0]})
		}
	}

	return enlaces
}

// incumplidasDeUnaReferencia son las reglas que incumplen los enlaces del cuerpo
// a una referencia, y si tiene alguno.
type incumplidasDeUnaReferencia struct {
	enlazada    bool
	sinElVerbo  bool
	expresiones map[string]bool
}

// fallosDeLaRegla son los de la regla de las remisiones sobre los segmentos del
// cuerpo de una skill con las referencias dadas (contracts/remisiones.md §7.1 de
// H41): uno por referencia y motivo, «<skill>: references/<f>: <motivo>», en el
// orden de los nombres de las referencias —las de references/ y las que enlaza el
// cuerpo— y, en cada una, en el de los motivos.
func fallosDeLaRegla(skill string, segmentos, referencias []string) []string {
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

	for _, segmento := range segmentos {
		enMinuscula := strings.ToLower(segmento)

		for _, enlace := range enlacesAReferencias(segmento) {
			referencia := deLaReferencia(enlace.referencia)
			referencia.enlazada = true
			referencia.sinElVerbo = referencia.sinElVerbo || !verboDeLeer.MatchString(segmento[:enlace.inicio])

			for _, expresion := range expresionesFueraDeLasRemisiones {
				referencia.expresiones[expresion] = referencia.expresiones[expresion] ||
					strings.Contains(enMinuscula, expresion)
			}
		}
	}

	var fallos []string

	for _, nombre := range slices.Sorted(maps.Keys(incumplidas)) {
		fallos = append(fallos, fallosDeUnaReferencia(skill, nombre, incumplidas[nombre])...)
	}

	return fallos
}

// fallosDeUnaReferencia son los de una referencia, en el orden de los motivos: el
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

// comprobarLaListaDelPlan exige que las remisiones de los segmentos del cuerpo de
// la skill, fuera de la región de los paquetes, sean, enteras y en orden, las de
// remisionesDelPlan (contracts/remisiones.md §7.2 de H41).
func comprobarLaListaDelPlan(t *testing.T, skill string, segmentos []string) {
	t.Helper()

	delPlan, conLista := remisionesDelPlan[skill]
	if !conLista {
		t.Errorf("%s: no tiene lista de remisiones del plan", skill)

		return
	}

	remisiones := remisionesFueraDeLosPaquetes(segmentos)

	if numero := primeraDistinta(delPlan, remisiones); numero > 0 {
		assert.Equal(t, delPlan, remisiones, "%s: la remisión %d no es la del plan", skill, numero)
	}
}

// remisionesFueraDeLosPaquetes son los segmentos con un enlace a una referencia
// que no están entre las marcas de la región de los paquetes, en su orden.
func remisionesFueraDeLosPaquetes(segmentos []string) []string {
	var (
		remisiones    []string
		enLosPaquetes bool
	)

	for _, segmento := range segmentos {
		switch {
		case segmento == regionDeLosPaquetes.inicio:
			enLosPaquetes = true
		case segmento == regionDeLosPaquetes.fin:
			enLosPaquetes = false
		case !enLosPaquetes && len(enlacesAReferencias(segmento)) > 0:
			remisiones = append(remisiones, segmento)
		}
	}

	return remisiones
}

// primeraDistinta es el número, desde 1, de la primera posición en que las dos
// listas difieren, contando como distinta la que solo tiene una de ellas, o 0 si
// son iguales.
func primeraDistinta(unas, otras []string) int {
	for posicion := range max(len(unas), len(otras)) {
		if posicion >= len(unas) || posicion >= len(otras) || unas[posicion] != otras[posicion] {
			return posicion + 1
		}
	}

	return 0
}
