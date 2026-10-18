package plazos

import (
	"cmp"
	"errors"
	"fmt"
	"maps"
	"slices"
	"time"
)

// Las rutas de data/festivos/, relativas a la raíz del repositorio, las que
// nombran los defectos de carga.
const (
	rutaDeLosNacionales   = "data/festivos/nacionales-y-autonomicos.yaml"
	carpetaDeLosLocales   = "data/festivos/locales/"
	extensionDeLosLocales = ".yaml"
)

// rutaDeLocales es la ruta del fichero de festivos locales de una comunidad.
func rutaDeLocales(comunidad string) string {
	return carpetaDeLosLocales + comunidad + extensionDeLosLocales
}

// Ambito es el de un festivo.
type Ambito string

// Los tres ámbitos de un festivo (data-model §1). El fichero nacional lleva
// los dos primeros: nacional («*» o «**» en la resolución) y autonómico («***»).
const (
	AmbitoNacional   Ambito = "nacional"
	AmbitoAutonomico Ambito = "autonomico"
	AmbitoLocal      Ambito = "local"
)

// Ambitos es el vocabulario cerrado de ámbitos de un festivo, del más amplio
// al más estrecho. Devuelve una lista nueva en cada llamada.
func Ambitos() []Ambito {
	return []Ambito{AmbitoNacional, AmbitoAutonomico, AmbitoLocal}
}

// AmbitoDeCobertura es el de un tramo de la cobertura.
type AmbitoDeCobertura string

// Los dos ámbitos de la cobertura: los festivos de las resoluciones del BOE y
// los locales de un municipio.
const (
	CoberturaNacionalYAutonomica AmbitoDeCobertura = "nacional-y-autonomico"
	CoberturaLocal               AmbitoDeCobertura = "local"
)

// AmbitosDeCobertura es el vocabulario cerrado de ámbitos de la cobertura.
// Devuelve una lista nueva en cada llamada.
func AmbitosDeCobertura() []AmbitoDeCobertura {
	return []AmbitoDeCobertura{CoberturaNacionalYAutonomica, CoberturaLocal}
}

// Falta es el motivo por el que no hay festivos (FR-022).
type Falta string

// Los motivos de falta. Los tres de un municipio y el de una línea los escribe
// la derivación en el fichero local; los dos primeros no salen de un fichero:
// un año sin datos y una comunidad sin fichero local.
const (
	FaltaSinCargar          Falta = "sin-cargar"
	FaltaSinFestivosLocales Falta = "sin-festivos-locales"
	FaltaNoComunicado       Falta = "no-comunicado"
	FaltaLineaSinLeer       Falta = "linea-sin-leer"
	FaltaSinLinea           Falta = "sin-linea"
	FaltaSinMunicipio       Falta = "sin-municipio"
)

// FaltasDeCobertura es el vocabulario cerrado de las faltas de un tramo de la
// cobertura. Devuelve una lista nueva en cada llamada.
func FaltasDeCobertura() []Falta {
	return []Falta{FaltaSinCargar, FaltaSinFestivosLocales, FaltaNoComunicado, FaltaLineaSinLeer, FaltaSinLinea}
}

// MotivosDeFalta es el vocabulario cerrado de los motivos de falta del fichero
// local: los tres de un municipio y FaltaSinMunicipio, el de una línea que no
// casa con un único municipio. Devuelve una lista nueva en cada llamada.
func MotivosDeFalta() []Falta {
	return []Falta{FaltaNoComunicado, FaltaLineaSinLeer, FaltaSinLinea, FaltaSinMunicipio}
}

// Municipio es un municipio ya resuelto por territorio, como lo recibe el
// dominio (data-model §1, petición): el calendario usa su código INE y el de
// su comunidad.
type Municipio struct {
	// CodigoINE son sus cinco cifras.
	CodigoINE string
	// Nombre es el oficial.
	Nombre string
	// Comunidad es el código de su comunidad, el de data/territorio/comunidades/.
	Comunidad string
}

// Festivo es un festivo que hace inhábil un día para unos municipios.
type Festivo struct {
	Ambito Ambito
	// Municipio es el código INE del municipio, solo en uno local.
	Municipio string
	// Source es la disposición que lo fija.
	Source string
}

// FestivoDeUnaParte es un festivo que la resolución limita a una parte de la
// comunidad: no hace inhábil el día (FR-011).
type FestivoDeUnaParte struct {
	Ambito Ambito
	// Parte es la que nombra la resolución.
	Parte string
	// Source es la resolución.
	Source string
}

// Tramo es un tramo de años de la cobertura de un ámbito: un año cargado, con
// la disposición hasta la que llegan sus datos o la falta de festivos del
// municipio, o un hueco de años seguidos sin datos, con su falta.
type Tramo struct {
	Ambito AmbitoDeCobertura
	// Municipio es el código INE, solo en uno local.
	Municipio   string
	Desde       int
	Hasta       int
	Disposicion string
	Falta       Falta
}

// Alcance es hasta dónde llegan los festivos cargados de un ámbito: el primer y
// el último año cargados y la última disposición del último.
type Alcance struct {
	Desde       int
	Hasta       int
	Disposicion string
}

// fecha es un día del calendario, la clave de los festivos.
type fecha struct {
	anio int
	mes  time.Month
	dia  int
}

// fechaDe es el día de un instante, en su zona.
func fechaDe(instante time.Time) fecha {
	anio, mes, dia := instante.Date()

	return fecha{anio: anio, mes: mes, dia: dia}
}

// Calendario son los festivos de data/festivos/, ya leídos. Solo lo construye
// CargarCalendario y no cambia después: puede usarse desde varias goroutines a
// la vez.
type Calendario struct {
	// anios lleva de cada año cargado a su resolución.
	anios map[int]string
	// nacionales lleva de cada día a sus filas de la resolución.
	nacionales map[fecha][]filaNacional
	// locales lleva del código de cada comunidad con fichero a sus festivos.
	locales map[string]*calendarioLocal
}

// filaNacional es una fila del fichero nacional.
type filaNacional struct {
	ambito      Ambito
	comunidades []string
	parte       string
	source      string
}

// deAlguna dice si la fila es de la comunidad de alguno de los municipios.
func (f filaNacional) deAlguna(municipios []Municipio) bool {
	return slices.ContainsFunc(municipios, func(m Municipio) bool { return slices.Contains(f.comunidades, m.Comunidad) })
}

// calendarioLocal son los festivos locales de una comunidad.
type calendarioLocal struct {
	// disposiciones lleva de cada año cargado a la última de sus disposiciones.
	disposiciones map[int]string
	// festivos lleva de cada día al source de cada municipio con fiesta.
	festivos map[fecha]map[string]string
	// faltas lleva de cada año al motivo de cada municipio sin festivos.
	faltas map[int]map[string]Falta
}

// CargarCalendario lee los dos ficheros de data/festivos/: el nacional y, por
// el código de su comunidad, los locales (contracts/datos-de-plazos.md §1.1 y
// §1.2). Un fichero que no se puede leer —vacío, con una clave que no declara,
// una fecha que no es AAAA-MM-DD o un ámbito o un motivo de falta fuera de su
// vocabulario— no da calendario: el error dice todos, cada uno con su fichero,
// y no lleva clase de usuario. No hay más reglas de carga: lo que llevan los
// ficheros lo garantiza la derivación (research D25).
func CargarCalendario(nacionales []byte, locales map[string][]byte) (*Calendario, error) {
	c := &carga{calendario: &Calendario{
		anios:      map[int]string{},
		nacionales: map[fecha][]filaNacional{},
		locales:    make(map[string]*calendarioLocal, len(locales)),
	}}

	if fichero, err := leerNacionales(nacionales); err != nil {
		c.defectos = append(c.defectos, err)
	} else {
		c.nacional(fichero)
	}

	for _, comunidad := range slices.Sorted(maps.Keys(locales)) {
		if fichero, err := leerLocales(comunidad, locales[comunidad]); err != nil {
			c.defectos = append(c.defectos, err)
		} else {
			c.local(comunidad, fichero)
		}
	}

	if err := errors.Join(c.defectos...); err != nil {
		return nil, err
	}

	return c.calendario, nil
}

// leerNacionales lee data/festivos/nacionales-y-autonomicos.yaml.
func leerNacionales(contenido []byte) (FestivosNacionales, error) {
	var fichero FestivosNacionales
	err := decodificar(rutaDeLosNacionales, contenido, &fichero)

	return fichero, err
}

// leerLocales lee data/festivos/locales/<comunidad>.yaml.
func leerLocales(comunidad string, contenido []byte) (FestivosLocales, error) {
	var fichero FestivosLocales
	err := decodificar(rutaDeLocales(comunidad), contenido, &fichero)

	return fichero, err
}

// carga construye el calendario con los ficheros ya decodificados y acumula
// los datos que no se pueden leer.
type carga struct {
	calendario *Calendario
	defectos   []error
}

// defecto anota un dato de un fichero que no se puede leer.
func (c *carga) defecto(fichero, formato string, argumentos ...any) {
	c.defectos = append(c.defectos, &defectoDeDatos{fichero: fichero, motivo: fmt.Sprintf(formato, argumentos...)})
}

// fecha lee una fecha AAAA-MM-DD de un fichero.
func (c *carga) fecha(fichero, texto string) (fecha, bool) {
	dia, err := time.Parse(time.DateOnly, texto)
	if err != nil {
		c.defecto(fichero, "la fecha %q no es una fecha AAAA-MM-DD", texto)

		return fecha{}, false
	}

	return fechaDe(dia), true
}

// nacional carga el fichero nacional.
func (c *carga) nacional(fichero FestivosNacionales) {
	for _, anio := range fichero.Anios {
		c.calendario.anios[anio.Anio] = anio.Disposicion
	}

	for _, festivo := range fichero.Festivos {
		if festivo.Ambito != AmbitoNacional && festivo.Ambito != AmbitoAutonomico {
			c.defecto(rutaDeLosNacionales, "el ámbito %q no es %s ni %s", festivo.Ambito, AmbitoNacional, AmbitoAutonomico)
		}

		dia, leida := c.fecha(rutaDeLosNacionales, festivo.Fecha)
		if !leida {
			continue
		}

		c.calendario.nacionales[dia] = append(c.calendario.nacionales[dia], filaNacional{
			ambito:      festivo.Ambito,
			comunidades: slices.Clone(festivo.Comunidades),
			parte:       festivo.Parte,
			source:      festivo.Source,
		})
	}
}

// local carga el fichero local de una comunidad.
func (c *carga) local(comunidad string, fichero FestivosLocales) {
	ruta := rutaDeLocales(comunidad)
	local := &calendarioLocal{
		disposiciones: make(map[int]string, len(fichero.Anios)),
		festivos:      map[fecha]map[string]string{},
		faltas:        map[int]map[string]Falta{},
	}

	for _, anio := range fichero.Anios {
		var ultima string
		if len(anio.Disposiciones) > 0 {
			ultima = anio.Disposiciones[len(anio.Disposiciones)-1]
		}

		local.disposiciones[anio.Anio] = ultima
	}

	for _, festivo := range fichero.Festivos {
		if dia, leida := c.fecha(ruta, festivo.Fecha); leida {
			if local.festivos[dia] == nil {
				local.festivos[dia] = map[string]string{}
			}

			local.festivos[dia][festivo.Municipio] = festivo.Source
		}
	}

	for _, falta := range fichero.Faltas {
		if !slices.Contains(MotivosDeFalta(), falta.Motivo) {
			c.defecto(ruta, "el motivo de falta %q no es %s", falta.Motivo, enumeracion(MotivosDeFalta()))
		}

		if falta.Municipio == "" {
			continue
		}

		if local.faltas[falta.Anio] == nil {
			local.faltas[falta.Anio] = map[string]Falta{}
		}

		local.faltas[falta.Anio][falta.Municipio] = falta.Motivo
	}

	c.calendario.locales[comunidad] = local
}

// Inhabil dice si un día es inhábil para unos municipios: si es sábado o
// domingo, festivo nacional o autonómico de toda la comunidad de alguno de
// ellos, o festivo local de alguno (arts. 30.2 y 30.6 LPAC; FR-003). Un festivo
// de una parte de la comunidad no lo hace inhábil (FR-011), y un año sin
// festivos cargados solo tiene sus sábados y sus domingos.
func (c *Calendario) Inhabil(dia time.Time, municipios []Municipio) bool {
	if semana := dia.Weekday(); semana == time.Saturday || semana == time.Sunday {
		return true
	}

	clave := fechaDe(dia)

	for _, fila := range c.nacionales[clave] {
		if fila.parte == "" && fila.deAlguna(municipios) {
			return true
		}
	}

	return slices.ContainsFunc(municipios, func(municipio Municipio) bool {
		_, festivo := c.festivoLocal(clave, municipio)

		return festivo
	})
}

// festivoLocal da el source del festivo local de un municipio en un día, si lo
// tiene.
func (c *Calendario) festivoLocal(clave fecha, municipio Municipio) (string, bool) {
	local := c.locales[municipio.Comunidad]
	if local == nil {
		return "", false
	}

	source, festivo := local.festivos[clave][municipio.CodigoINE]

	return source, festivo
}

// FestivosDe son los festivos que hacen inhábil un día para unos municipios,
// cada uno una vez, con su ámbito, su municipio si es local y su source: del
// más amplio al más estrecho y, los locales, por municipio. Un sábado o un
// domingo no son festivos.
func (c *Calendario) FestivosDe(dia time.Time, municipios []Municipio) []Festivo {
	clave := fechaDe(dia)

	var festivos []Festivo

	for _, fila := range c.nacionales[clave] {
		if fila.parte == "" && fila.deAlguna(municipios) {
			festivos = append(festivos, Festivo{Ambito: fila.ambito, Source: fila.source})
		}
	}

	for _, municipio := range municipios {
		if source, festivo := c.festivoLocal(clave, municipio); festivo {
			festivos = append(festivos, Festivo{Ambito: AmbitoLocal, Municipio: municipio.CodigoINE, Source: source})
		}
	}

	ambitos := Ambitos()

	slices.SortFunc(festivos, func(a, b Festivo) int {
		return cmp.Or(
			cmp.Compare(slices.Index(ambitos, a.Ambito), slices.Index(ambitos, b.Ambito)),
			cmp.Compare(a.Municipio, b.Municipio),
			cmp.Compare(a.Source, b.Source),
		)
	})

	return slices.Compact(festivos)
}

// PartesDe son los festivos de un día que la resolución limita a una parte de
// una comunidad, por parte (FR-011, FR-025).
func (c *Calendario) PartesDe(dia time.Time, comunidad string) []FestivoDeUnaParte {
	var partes []FestivoDeUnaParte

	for _, fila := range c.nacionales[fechaDe(dia)] {
		if fila.parte != "" && slices.Contains(fila.comunidades, comunidad) {
			partes = append(partes, FestivoDeUnaParte{Ambito: fila.ambito, Parte: fila.parte, Source: fila.source})
		}
	}

	slices.SortFunc(partes, func(a, b FestivoDeUnaParte) int {
		return cmp.Or(cmp.Compare(a.Parte, b.Parte), cmp.Compare(a.Ambito, b.Ambito), cmp.Compare(a.Source, b.Source))
	})

	return slices.Compact(partes)
}

// Cobertura es lo que cubren los datos de los años de desde a hasta, ambos
// incluidos, para unos municipios: primero el ámbito nacional y autonómico y
// después el local de cada municipio, una vez cada uno y en el orden dado. Cada
// ámbito va por tramos: uno por año cargado, con la disposición hasta la que
// llegan sus datos o, en uno local, la falta de festivos del municipio; y uno
// por hueco de años seguidos sin datos, con FaltaSinCargar o, en una comunidad
// sin fichero local, FaltaSinFestivosLocales (FR-026; research D5). Con desde
// mayor que hasta no hay ningún tramo.
func (c *Calendario) Cobertura(desde, hasta int, municipios []Municipio) []Tramo {
	if desde > hasta {
		return nil
	}

	tramos := tramosDe(Tramo{Ambito: CoberturaNacionalYAutonomica}, desde, hasta, c.anioNacional)
	vistos := make(map[string]bool, len(municipios))

	for _, municipio := range municipios {
		if vistos[municipio.CodigoINE] {
			continue
		}

		vistos[municipio.CodigoINE] = true

		tramos = append(tramos, tramosDe(Tramo{Ambito: CoberturaLocal, Municipio: municipio.CodigoINE}, desde, hasta,
			func(anio int) (string, Falta, bool) { return c.anioLocal(municipio, anio) })...)
	}

	return tramos
}

// anioNacional es la cobertura nacional y autonómica de un año: su resolución
// si está cargado.
func (c *Calendario) anioNacional(anio int) (disposicion string, falta Falta, cargado bool) {
	disposicion, cargado = c.anios[anio]
	if !cargado {
		return "", FaltaSinCargar, false
	}

	return disposicion, "", true
}

// anioLocal es la cobertura local de un municipio en un año: la última
// disposición del año si está cargado y el municipio no tiene falta, o su falta.
func (c *Calendario) anioLocal(municipio Municipio, anio int) (disposicion string, falta Falta, cargado bool) {
	local := c.locales[municipio.Comunidad]
	if local == nil {
		return "", FaltaSinFestivosLocales, false
	}

	disposicion, cargado = local.disposiciones[anio]
	if !cargado {
		return "", FaltaSinCargar, false
	}

	if falta, tiene := local.faltas[anio][municipio.CodigoINE]; tiene {
		return "", falta, true
	}

	return disposicion, "", true
}

// tramosDe parte los años de desde a hasta en los tramos de un ámbito: cada
// año cargado es un tramo, y los años seguidos sin cargar con la misma falta,
// otro.
func tramosDe(base Tramo, desde, hasta int, deUnAnio func(int) (string, Falta, bool)) []Tramo {
	var tramos []Tramo

	hueco := false

	for anio := desde; anio <= hasta; anio++ {
		disposicion, falta, cargado := deUnAnio(anio)

		if !cargado && hueco && tramos[len(tramos)-1].Falta == falta {
			tramos[len(tramos)-1].Hasta = anio

			continue
		}

		tramo := base
		tramo.Desde, tramo.Hasta, tramo.Disposicion, tramo.Falta = anio, anio, disposicion, falta
		tramos = append(tramos, tramo)
		hueco = !cargado
	}

	return tramos
}

// AlcanceNacional es hasta dónde llegan los festivos nacionales y
// autonómicos cargados; sin ningún año cargado, nada.
func (c *Calendario) AlcanceNacional() (Alcance, bool) {
	return alcanceDe(c.anios)
}

// AlcanceLocal es hasta dónde llegan los festivos locales cargados de una
// comunidad; sin fichero local o sin ningún año cargado, nada.
func (c *Calendario) AlcanceLocal(comunidad string) (Alcance, bool) {
	local := c.locales[comunidad]
	if local == nil {
		return Alcance{}, false
	}

	return alcanceDe(local.disposiciones)
}

// alcanceDe es el alcance de los años cargados de un ámbito, cada uno con la
// disposición hasta la que llegan sus datos.
func alcanceDe(disposiciones map[int]string) (Alcance, bool) {
	if len(disposiciones) == 0 {
		return Alcance{}, false
	}

	anios := slices.Sorted(maps.Keys(disposiciones))
	hasta := anios[len(anios)-1]

	return Alcance{Desde: anios[0], Hasta: hasta, Disposicion: disposiciones[hasta]}, true
}
