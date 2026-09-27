package app

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
)

// Lo que install, list y doctor cuentan a una persona cuando no se pide
// --json: el texto que va en Resultado.Legible y que el kernel escribe en la
// salida estándar en lugar de la tabla mínima (docs/ADR/0026). Quien lo lee
// acaba de instalar kitlegal y muchas veces no es técnico, así que no hay ni
// una palabra del sobre —fuente, url, huella—, cada host lleva el nombre de su
// marca, HOME se abrevia a ~ y no hay colores ni secuencias de escape: el
// texto es determinista y se puede comparar byte a byte. Se compone a partir
// de los mismos valores que van en data y en el mismo instante, así que no
// puede decir otra cosa que el sobre.

// Los nombres con los que se presenta cada host: el de la marca, y no la
// clave con la que se invoca y que va en el sobre (ADR 0025). Un host que
// faltara aquí saldría por su clave; TestNombresDeMarcaDeLosHosts exige que
// no falte ninguno.
var nombresDeMarca = map[string]string{
	"claude":      "Claude Code",
	"antigravity": "Antigravity",
}

// preguntaDeEjemplo es con la que se invita a probar el agente tras instalar:
// una pregunta a la que boe-legislacion responde con el texto vigente.
const preguntaDeEjemplo = "«¿Qué dice el artículo 21 de la Ley 39/2015?»"

// Las sangrías de la lista: dos espacios para cada skill y cuatro para cada
// una de sus entradas de host.
const (
	sangriaDeSkill = "  "
	sangriaDeHost  = "    "
)

// hogar es lo que hace falta para abreviar una ruta: el valor de HOME con el
// que se validó la invocación. Vacío, no se abrevia nada.
type hogar string

// abreviar sustituye HOME al principio de ruta por ~, como la escribiría quien
// la lee: la propia raíz queda en ~ y lo que cuelga de ella en ~/…. Una ruta
// relativa, o que no cuelga de HOME, o un HOME vacío o que es la raíz del
// sistema, la dejan como está.
func (h hogar) abreviar(ruta string) string {
	home := strings.TrimSuffix(string(h), "/")
	if home == "" {
		return ruta
	}

	if ruta == home {
		return "~"
	}

	if resto, cuelga := strings.CutPrefix(ruta, home+"/"); cuelga {
		return "~/" + resto
	}

	return ruta
}

// nombreDeMarca es el nombre con el que se presenta el host, o su clave si
// no tiene uno.
func nombreDeMarca(host string) string {
	if nombre, hay := nombresDeMarca[host]; hay {
		return nombre
	}

	return host
}

// quienLeeElNeutro es, entre paréntesis y con un espacio delante, qué agentes
// leen el directorio neutro del ámbito sin necesitar un enlace: Codex y
// Antigravity en un proyecto, solo Codex en la cuenta (ADR 0025); nada con
// --dir, que no es el directorio de ningún agente.
func quienLeeElNeutro(ambito instalacion.Ambito) string {
	switch ambito.Clase() {
	case instalacion.AmbitoLocal:
		return " (Codex y Antigravity las leen de ahí)"
	case instalacion.AmbitoGlobal:
		return " (Codex las lee de ahí)"
	case instalacion.AmbitoDir:
	}

	return ""
}

// ordenDeInstall es «kitlegal skills install» con las banderas del ámbito y,
// si se dan, las de los hosts, tal como se copia y se pega.
func ordenDeInstall(ambito instalacion.Ambito, hosts ...string) string {
	partes := []string{"kitlegal", "skills", "install"}

	if banderas := ambito.Banderas(); banderas != "" {
		partes = append(partes, banderas)
	}

	for _, host := range hosts {
		partes = append(partes, "--host", host)
	}

	return strings.Join(partes, " ")
}

// ordenDeDoctor es «kitlegal skills doctor» con las banderas del ámbito.
func ordenDeDoctor(ambito instalacion.Ambito) string {
	orden := "kitlegal skills doctor"
	if banderas := ambito.Banderas(); banderas != "" {
		orden += " " + banderas
	}

	return orden
}

// entradaDeHost es una entrada de host tal como se presenta: el nombre de la
// marca, la ruta abreviada y, si es una copia, por qué no es un enlace.
type entradaDeHost struct {
	marca string
	ruta  string
	copia bool
}

// lineaDeSkill es una skill de la lista: su nombre, lo que se dice de ella en
// la misma línea —su estado en install, su versión en list— y sus entradas.
type lineaDeSkill struct {
	nombre   string
	detalle  string
	entradas []entradaDeHost
}

// entradasDe convierte los enlaces de una skill en sus entradas presentables.
func entradasDe(enlaces []instalacion.Enlace, h hogar) []entradaDeHost {
	entradas := make([]entradaDeHost, 0, len(enlaces))
	for _, enlace := range enlaces {
		entradas = append(entradas, entradaDeHost{
			marca: nombreDeMarca(enlace.Host),
			ruta:  h.abreviar(enlace.Ruta),
			copia: enlace.Modo == instalacion.ModoCopia,
		})
	}

	return entradas
}

// lista escribe las skills, una por línea con su detalle alineado en columna,
// y debajo de cada una sus entradas de host, con el nombre de la marca también
// en columna. Una copia lo dice y dice por qué: el sistema no dejó crear el
// enlace en esa carpeta, así que la entrada es una copia de la skill, que
// install vuelve a actualizar en cada ejecución.
func lista(b *strings.Builder, skills []lineaDeSkill) {
	anchoDeNombre, anchoDeMarca := 0, 0

	for _, skill := range skills {
		anchoDeNombre = max(anchoDeNombre, utf8.RuneCountInString(skill.nombre))
		for _, entrada := range skill.entradas {
			anchoDeMarca = max(anchoDeMarca, utf8.RuneCountInString(entrada.marca))
		}
	}

	for _, skill := range skills {
		fmt.Fprintf(b, "%s%-*s  %s\n", sangriaDeSkill, anchoDeNombre, skill.nombre, skill.detalle)

		for _, entrada := range skill.entradas {
			fmt.Fprintf(b, "%s%-*s  %s", sangriaDeHost, anchoDeMarca, entrada.marca, entrada.ruta)

			if entrada.copia {
				b.WriteString(" (copia, porque en esa carpeta no se pueden crear enlaces)")
			}

			b.WriteString("\n")
		}
	}
}

// hostsSinEntrada son, en el orden de los hosts, los del ámbito en los que
// ninguna de las skills —cada una con su lista de enlaces— tiene una entrada:
// los agentes que no verán las skills salvo que se les enlace con --host.
func hostsSinEntrada(ambito instalacion.Ambito, enlacesPorSkill [][]instalacion.Enlace) []string {
	var sinEntrada []string

	for _, host := range ambito.Hosts() {
		conEntrada := false

		for _, enlaces := range enlacesPorSkill {
			for _, enlace := range enlaces {
				conEntrada = conEntrada || enlace.Host == host
			}
		}

		if !conEntrada {
			sinEntrada = append(sinEntrada, host)
		}
	}

	return sinEntrada
}

// avisosDeHostsSinEntrada escribe, por cada host del ámbito sin ninguna
// entrada, la línea que dice que ese agente no verá las skills y la orden que
// lo enlaza.
func avisosDeHostsSinEntrada(b *strings.Builder, ambito instalacion.Ambito, hosts []string) {
	for _, host := range hosts {
		fmt.Fprintf(b, "%s no las verá: si lo usas, ejecuta «%s».\n",
			nombreDeMarca(host), ordenDeInstall(ambito, host))
	}
}

// legibleDeInstall es lo que install cuenta: dónde han quedado las skills y
// quién las lee de ahí, cada skill con su estado y sus entradas de host, los
// agentes que no las verán y, para terminar, qué hacer ahora: probar el
// agente, o nada, si ya estaban al día.
func legibleDeInstall(
	skills []instalacion.SkillInstalada, ambito instalacion.Ambito, version string, h hogar,
) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Skills de kitlegal %s en %s%s:\n\n",
		version, h.abreviar(ambito.Neutro()), quienLeeElNeutro(ambito))

	lineas := make([]lineaDeSkill, 0, len(skills))
	enlacesPorSkill := make([][]instalacion.Enlace, 0, len(skills))
	sinCambios := true

	for _, skill := range skills {
		lineas = append(lineas, lineaDeSkill{
			nombre:   skill.Nombre,
			detalle:  string(skill.Estado),
			entradas: entradasDe(skill.Enlaces, h),
		})
		enlacesPorSkill = append(enlacesPorSkill, skill.Enlaces)
		sinCambios = sinCambios && skill.Estado == instalacion.EstadoSinCambios
	}

	lista(&b, lineas)
	b.WriteString("\n")
	avisosDeHostsSinEntrada(&b, ambito, hostsSinEntrada(ambito, enlacesPorSkill))

	if sinCambios {
		b.WriteString("Ya estaban instaladas y al día: no se ha cambiado nada.\n")
	} else {
		b.WriteString("Abre tu agente y pregúntale, por ejemplo: " + preguntaDeEjemplo + "\n")
	}

	return b.String()
}

// sinNadaInstalado es lo que list y doctor cuentan de un ámbito sin manifiesto:
// que no hay skills y cómo instalarlas.
func sinNadaInstalado(ambito instalacion.Ambito, h hogar) string {
	return fmt.Sprintf("No hay skills de kitlegal en %s. Para instalarlas: %s\n",
		h.abreviar(ambito.Neutro()), ordenDeInstall(ambito))
}

// legibleDeList es lo que list cuenta: dónde están las skills y quién las lee
// de ahí, cada una con su versión y sus entradas de host, la que este binario
// ya no lleva señalada, y los agentes que no las verán.
func legibleDeList(listado instalacion.Listado, ambito instalacion.Ambito, h hogar) string {
	if !listado.Manifiesto {
		return sinNadaInstalado(ambito, h)
	}

	var b strings.Builder

	if len(listado.Skills) == 0 {
		fmt.Fprintf(&b, "kitlegal %s dejó preparado %s, pero no hay ninguna skill instalada. Para instalarlas: %s\n",
			*listado.Version, h.abreviar(ambito.Neutro()), ordenDeInstall(ambito))

		return b.String()
	}

	fmt.Fprintf(&b, "Skills de kitlegal en %s%s:\n\n", h.abreviar(ambito.Neutro()), quienLeeElNeutro(ambito))

	lineas := make([]lineaDeSkill, 0, len(listado.Skills))
	enlacesPorSkill := make([][]instalacion.Enlace, 0, len(listado.Skills))

	for _, skill := range listado.Skills {
		detalle := skill.Version
		if !skill.Empotrada {
			detalle += " (esta versión de kitlegal ya no la lleva)"
		}

		lineas = append(lineas, lineaDeSkill{
			nombre:   skill.Nombre,
			detalle:  detalle,
			entradas: entradasDe(skill.Enlaces, h),
		})
		enlacesPorSkill = append(enlacesPorSkill, skill.Enlaces)
	}

	lista(&b, lineas)

	if sinEntrada := hostsSinEntrada(ambito, enlacesPorSkill); len(sinEntrada) > 0 {
		b.WriteString("\n")
		avisosDeHostsSinEntrada(&b, ambito, sinEntrada)
	}

	return b.String()
}

// legibleDeDoctor es lo que doctor cuenta: que todo está en orden, o cada
// hallazgo explicado en una frase con la orden que lo arregla debajo, lista
// para copiar, y qué hacer con ellas.
func legibleDeDoctor(diagnostico instalacion.Diagnostico, ambito instalacion.Ambito, h hogar) string {
	if !diagnostico.Manifiesto {
		return sinNadaInstalado(ambito, h)
	}

	neutro := h.abreviar(ambito.Neutro())

	if len(diagnostico.Hallazgos) == 0 {
		return fmt.Sprintf("Todo en orden: las skills de %s son de kitlegal %s y están como las dejó.\n",
			neutro, diagnostico.VersionDelBinario)
	}

	var b strings.Builder

	fmt.Fprintf(&b, "%s en las skills de %s:\n\n", cuentaDeHallazgos(len(diagnostico.Hallazgos)), neutro)

	for i, hallazgo := range diagnostico.Hallazgos {
		fmt.Fprintf(&b, "%s%d. %s\n", sangriaDeSkill, i+1, explicacionDelHallazgo(hallazgo, diagnostico, ambito, h))
		fmt.Fprintf(&b, "%s   %s\n", sangriaDeSkill, hallazgo.Orden)
	}

	fmt.Fprintf(&b, "\n%s y vuelve a comprobarlo con «%s».\n", ejecutaLasOrdenes(len(diagnostico.Hallazgos)), ordenDeDoctor(ambito))

	return b.String()
}

// cuentaDeHallazgos es la cabecera de la lista de hallazgos, en singular o en
// plural.
func cuentaDeHallazgos(n int) string {
	if n == 1 {
		return "Hay 1 cosa que arreglar"
	}

	return fmt.Sprintf("Hay %d cosas que arreglar", n)
}

// ejecutaLasOrdenes es el principio de la despedida: en singular, o en plural
// y en el orden de la lista, que es el que deja doctor en 0 (FR-066).
func ejecutaLasOrdenes(n int) string {
	if n == 1 {
		return "Ejecuta la orden"
	}

	return "Ejecuta las órdenes en ese orden"
}

// explicacionDelHallazgo es la frase que explica un hallazgo a quien no sabe
// qué es un manifiesto ni un enlace simbólico: qué hay en esa ruta y por qué
// no es lo que kitlegal dejó (FR-065; contracts/applet-skills.md §4.3).
func explicacionDelHallazgo(
	hallazgo instalacion.Hallazgo, diagnostico instalacion.Diagnostico, ambito instalacion.Ambito, h hogar,
) string {
	ruta := h.abreviar(hallazgo.Ruta)

	switch hallazgo.Clase {
	case instalacion.HallazgoFicheroEditado:
		return fmt.Sprintf("%s ya no es el fichero que dejó kitlegal: se ha editado, se ha sustituido o falta.", ruta)
	case instalacion.HallazgoEnlaceColgando:
		return fmt.Sprintf("%s es un enlace que apunta a algo que ya no existe.", ruta)
	case instalacion.HallazgoEnlaceAOtroSitio:
		return fmt.Sprintf("%s no es el enlace que creó kitlegal: apunta a otro sitio, es otra cosa o falta.", ruta)
	case instalacion.HallazgoCopia:
		return fmt.Sprintf("%s es una copia, y en esa carpeta ya se pueden crear enlaces: puede pasar a ser uno.", ruta)
	case instalacion.HallazgoVersionDistinta:
		if hallazgo.Ruta == ambito.RutaDelManifiesto() && diagnostico.Version != nil {
			return fmt.Sprintf("Las skills son de kitlegal %s y este kitlegal es %s.",
				*diagnostico.Version, diagnostico.VersionDelBinario)
		}

		return fmt.Sprintf("La skill de %s es de otra versión de kitlegal, y este kitlegal es %s.",
			ruta, diagnostico.VersionDelBinario)
	}

	return fmt.Sprintf("%s: %s.", ruta, hallazgo.Clase)
}
