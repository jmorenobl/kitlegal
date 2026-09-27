package instalacion

import (
	"errors"
	"maps"
	"slices"
)

// Listar da lo que kitlegal tiene instalado en el ámbito según su manifiesto,
// sin examinar nada más ni cambiar nada en el disco (FR-060 a FR-062;
// contracts/applet-skills.md §4.2): el directorio neutro, si hay manifiesto,
// su versión y, por cada skill que declara, en orden de nombre, su ruta, la
// versión del binario de su instalación, si el binario la empotra —también se
// dan las que no (FR-036)— y sus entradas de host con su modo. Sin
// manifiesto, la versión es nula y la lista está vacía.
//
// Antes lee el ámbito como lo lee doctor: con una guarda que no es un
// directorio real, un manifiesto ilegible o un manifiesto con entradas de un
// host que el ámbito no tiene, devuelve un *AmbitoIlegible que lo nombra y
// ninguna skill.
// Un fallo del Disco se devuelve tal cual.
func Listar(disco Disco, ambito Ambito, empotradas []SkillEmpotrada) (Listado, error) {
	manifiesto, hay, err := leerElAmbito(disco, ambito, verboList)
	if err != nil {
		return Listado{}, err
	}

	listado := Listado{Directorio: ambito.Neutro(), Manifiesto: hay, Skills: []SkillListada{}}
	if !hay {
		return listado, nil
	}

	version := manifiesto.Version
	listado.Version = &version

	for _, nombre := range slices.Sorted(maps.Keys(manifiesto.Skills)) {
		declarada := manifiesto.Skills[nombre]

		listado.Skills = append(listado.Skills, SkillListada{
			Nombre:    nombre,
			Ruta:      ambito.RutaDeSkill(nombre),
			Version:   declarada.Version,
			Empotrada: empotra(empotradas, nombre),
			Enlaces:   enlacesDeclarados(ambito, nombre, declarada.Hosts),
		})
	}

	return listado, nil
}

// empotra dice si nombre es el de alguna de las skills empotradas.
func empotra(empotradas []SkillEmpotrada, nombre string) bool {
	return slices.ContainsFunc(empotradas, func(skill SkillEmpotrada) bool { return skill.Nombre == nombre })
}

// enlacesDeclarados son las entradas de host de la skill nombre que declara el
// manifiesto, como se presentan: cada una con su host y su modo, en el orden de
// los hosts; ninguna si no declara ninguna.
func enlacesDeclarados(ambito Ambito, nombre string, entradas map[string]EntradaDeHost) []Enlace {
	enlaces := []Enlace{}

	for _, host := range nombresDeHosts() {
		if entrada, hay := entradas[host]; hay {
			enlaces = append(enlaces, Enlace{Host: host, Ruta: ambito.RutaDeHost(host, nombre), Modo: entrada.Modo})
		}
	}

	return enlaces
}

// leerElAmbito lee el manifiesto del ámbito para list y doctor, con las tres
// primeras filas de data-model §4.1, las mismas que comprueba install: cada
// guarda tiene que ser un directorio real o no existir, sin seguir enlaces
// (FR-027), y el manifiesto, legible (FR-035) y sin entradas de un host que el
// ámbito no tiene (FR-013; ADR 0025). Si no, devuelve el *AmbitoIlegible del verbo que nombra la
// primera entrada que no lo es, sin examinar nada por debajo de ella.
//
// Si falta una guarda o el manifiesto, no hay manifiesto: devuelve falso, sin
// error. Un fallo del Disco al examinar se devuelve tal cual; no poder leer el
// manifiesto lo hace ilegible, como para install.
func leerElAmbito(disco Disco, ambito Ambito, verbo string) (Manifiesto, bool, error) {
	for _, guarda := range ambito.Guardas() {
		entrada, err := disco.Examinar(guarda)
		if err != nil {
			return Manifiesto{}, false, err
		}

		switch entrada.Tipo {
		case EntradaDirectorio:
		case EntradaAusente:
			return Manifiesto{}, false, nil
		default:
			return Manifiesto{}, false, ambitoIlegible(verbo, ConflictoRutaQueNoEsDirectorio, guarda)
		}
	}

	ruta := ambito.RutaDelManifiesto()

	manifiesto, err := leerManifiestoDe(disco, ruta)

	var ilegible *ManifiestoIlegible

	switch {
	case errors.As(err, &ilegible):
		return Manifiesto{}, false, ambitoIlegible(verbo, ConflictoManifiestoIlegible, ruta)
	case err != nil:
		return Manifiesto{}, false, err
	case manifiesto.Version == "":
		// Un manifiesto leído lleva siempre versión, que LeerManifiesto no
		// admite vacía: sin ella, no había kitlegal.json.
		return Manifiesto{}, false, nil
	case conEntradasDeHostAjenas(manifiesto, ambito):
		return Manifiesto{}, false, ambitoIlegible(verbo, ConflictoManifiestoConEntradasDeHost, ruta)
	}

	return manifiesto, true, nil
}
