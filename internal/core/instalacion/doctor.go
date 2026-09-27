package instalacion

import (
	"fmt"
	"maps"
	"path"
	"slices"
)

// Diagnosticar compara el disco con el manifiesto del ámbito y con el binario
// de version, sin cambiar nada en el disco, y da cada hallazgo con la orden que
// lo arregla (FR-065 a FR-069; data-model §6; contracts/applet-skills.md
// §4.3, §5 y §6). Solo mira las skills que el manifiesto declara y el binario
// empotra: las demás ni se examinan ni se nombran (FR-036). Por cada una, en
// orden de nombre:
//
//  1. fichero editado: un fichero declarado, del directorio neutro o de una
//     copia de host que sigue siendo un directorio real, con otra huella, que
//     ya no es un fichero regular —se retira con rm, o con rm -r si es un
//     directorio real— o que falta;
//  2. enlace colgando: su directorio o un directorio intermedio que es un
//     enlace que no resuelve, que se retira con rm; o su enlace de host de
//     FR-021, declarado, que no resuelve;
//  3. enlace a otro sitio: su directorio o un directorio intermedio que es
//     cualquier otra cosa que no es un directorio real, que se retira con rm;
//     o una entrada suya de host, que ya no es lo que se declaró —se retira
//     con rm, o con rm -r si es un directorio real— o que falta, también
//     porque algún directorio de la raíz al de skills del host ya no es un
//     directorio real, que no se lee a través de él;
//  4. copia: una copia suya de host declarada, que sigue siendo un directorio
//     real, donde el Enlazador dice que ya se puede crear el enlace: el mismo
//     que usa install, preguntado una sola vez por el directorio de skills de
//     cada host (FR-069);
//  5. versión distinta: la del manifiesto, distinta de version según FR-077,
//     con un único hallazgo cuya orden reinstala todas; y si no, la de cada
//     skill que difiere.
//
// Por debajo de una entrada que no es un directorio real no hay ningún
// hallazgo: lo es ella; y lo que no está declarado no es un hallazgo (FR-047,
// FR-066). Devuelve el Diagnostico con los hallazgos en el orden de FR-066, o
// con la lista vacía si no hay ninguno o no hay manifiesto: encontrar algo es el
// resultado de una verificación que ha funcionado, no un fallo (ADR 0023).
//
// Antes lee el ámbito como list: con una guarda que no es un directorio real,
// un manifiesto ilegible o un manifiesto con entradas de un host que el ámbito
// no tiene, devuelve un *AmbitoIlegible que lo nombra y ningún hallazgo. Un fallo del
// Disco o del Enlazador se devuelve tal cual, y una version que el manifiesto
// no admitiría —la que declararía el install de cada orden— es un error antes
// de examinar nada.
func Diagnosticar(
	disco Disco, enlazador Enlazador, ambito Ambito, empotradas []SkillEmpotrada, version string,
) (Diagnostico, error) {
	if err := comprobarVersion(version); err != nil {
		return Diagnostico{}, fmt.Errorf("la versión del binario no se puede declarar en el manifiesto: %w", err)
	}

	manifiesto, hay, err := leerElAmbito(disco, ambito, verboDoctor)
	if err != nil {
		return Diagnostico{}, err
	}

	diagnostico := Diagnostico{
		Directorio:        ambito.Neutro(),
		Manifiesto:        hay,
		VersionDelBinario: version,
		Hallazgos:         []Hallazgo{},
	}
	if !hay {
		return diagnostico, nil
	}

	versionDelManifiesto := manifiesto.Version
	diagnostico.Version = &versionDelManifiesto

	r := &revision{
		disco:       disco,
		enlazador:   enlazador,
		ambito:      ambito,
		manifiesto:  manifiesto,
		hostsReales: map[string]bool{},
		sondas:      map[string]bool{},
	}

	nombres := declaradasYEmpotradas(manifiesto, empotradas)
	for _, nombre := range nombres {
		if err := r.revisarSkill(nombre); err != nil {
			return Diagnostico{}, err
		}
	}

	r.revisarVersiones(nombres, version)

	diagnostico.Hallazgos = enSuOrden(r.pendientes)

	return diagnostico, nil
}

// declaradasYEmpotradas son los nombres de las skills que el manifiesto
// declara y el binario empotra, en orden de nombre.
func declaradasYEmpotradas(manifiesto Manifiesto, empotradas []SkillEmpotrada) []string {
	nombres := []string{}

	for _, nombre := range slices.Sorted(maps.Keys(manifiesto.Skills)) {
		if empotra(empotradas, nombre) {
			nombres = append(nombres, nombre)
		}
	}

	return nombres
}

// revision es el estado de una invocación de doctor: lo que examina, dónde,
// lo que ya sabe de cada host y la respuesta de su sonda, y cada hallazgo que
// lleva encontrado.
type revision struct {
	disco      Disco
	enlazador  Enlazador
	ambito     Ambito
	manifiesto Manifiesto

	// hostsReales dice, por cada host ya examinado, si cada directorio de la
	// raíz al de skills del host es un directorio real.
	hostsReales map[string]bool
	// sondas es la respuesta de Disponible por el directorio de skills de cada
	// host al que ya se ha preguntado.
	sondas map[string]bool

	pendientes []pendiente
}

// anotar añade el hallazgo de esa clase en ruta, con la orden que retira la
// entrada como diga quitar y reinstala las skills nombres.
func (r *revision) anotar(clase ClaseDeHallazgo, ruta string, quitar retirada, nombres ...string) {
	r.pendientes = append(r.pendientes, pendiente{
		hallazgo: Hallazgo{
			Clase: clase,
			Ruta:  ruta,
			Orden: ordenQueArregla(r.ambito, quitar, ruta, nombres, r.hostsDeclarados(nombres)),
		},
		retira: quitar != sinRetirar,
	})
}

// hostsDeclarados son los hosts, en su orden, en los que alguna de las skills
// nombres tiene una entrada declarada.
func (r *revision) hostsDeclarados(nombres []string) []string {
	var hosts []string

	for _, host := range nombresDeHosts() {
		if slices.ContainsFunc(nombres, func(nombre string) bool {
			_, hay := r.manifiesto.Skills[nombre].Hosts[host]

			return hay
		}) {
			hosts = append(hosts, host)
		}
	}

	return hosts
}

// anotarNoDirectorio añade el hallazgo de la entrada de ruta, que tiene que
// ser un directorio real y no lo es: enlace colgando si es un enlace que no
// resuelve y enlace a otro sitio si es cualquier otra cosa; en los dos casos,
// con rm, que no la sigue (FR-065 (2) y (3)).
func (r *revision) anotarNoDirectorio(nombre, ruta string, entrada Entrada) {
	clase := HallazgoEnlaceAOtroSitio
	if entrada.Tipo == EntradaEnlace && !entrada.Resuelve {
		clase = HallazgoEnlaceColgando
	}

	r.anotar(clase, ruta, conRm, nombre)
}

// revisarSkill revisa la skill nombre en el directorio neutro y, en el orden
// de los hosts, en cada uno en el que tiene una entrada declarada.
func (r *revision) revisarSkill(nombre string) error {
	declarada := r.manifiesto.Skills[nombre]
	base := r.ambito.RutaDeSkill(nombre)
	declarados := relativas(declarada.Ficheros, nombre+"/")

	entrada, err := r.disco.Examinar(base)
	if err != nil {
		return err
	}

	switch entrada.Tipo {
	case EntradaDirectorio:
		err = r.revisarFicheros(nombre, base, declarados)
	case EntradaAusente:
		for _, rel := range slices.Sorted(maps.Keys(declarados)) {
			r.anotar(HallazgoFicheroEditado, path.Join(base, rel), sinRetirar, nombre)
		}
	case EntradaEnlace, EntradaFichero, EntradaOtra:
		r.anotarNoDirectorio(nombre, base, entrada)
	}

	if err != nil {
		return err
	}

	for _, host := range nombresDeHosts() {
		entrada, hay := declarada.Hosts[host]
		if !hay {
			continue
		}

		if err := r.revisarHost(nombre, host, entrada); err != nil {
			return err
		}
	}

	return nil
}

// revisarFicheros revisa los ficheros declarados de base, el directorio real
// de la skill nombre o de su copia de host, relativos a él con su huella:
// primero cada directorio intermedio, que tiene que ser real o no existir, y
// después cada fichero cuyo directorio es real o falta. Nada se examina por
// debajo de un intermedio que no es un directorio real: el hallazgo es él.
func (r *revision) revisarFicheros(nombre, base string, declarados map[string]string) error {
	rutas := slices.Sorted(maps.Keys(declarados))
	estados := map[string]estadoDeDirectorio{".": directorioReal}

	for _, rel := range rutas {
		if _, err := r.estadoDelIntermedio(nombre, base, path.Dir(rel), estados); err != nil {
			return err
		}
	}

	for _, rel := range rutas {
		if estado, esIntermedio := estados[rel]; esIntermedio && estado == directorioRoto {
			continue
		}

		ruta := path.Join(base, rel)

		switch estados[path.Dir(rel)] {
		case directorioReal:
			if err := r.revisarFichero(nombre, ruta, declarados[rel]); err != nil {
				return err
			}
		case directorioAusente:
			r.anotar(HallazgoFicheroEditado, ruta, sinRetirar, nombre)
		case directorioRoto:
			// Lo que cuelga de un intermedio que no es un directorio real no
			// se examina: el hallazgo es él (FR-065 (1)).
		}
	}

	return nil
}

// estadoDelIntermedio es el del directorio dir, relativo a base, que se
// examina solo si el de encima es real: bajo uno que falta, falta, y bajo uno
// que no es un directorio real, no se examina. El que existe y no es un
// directorio real es un hallazgo. Cada uno se examina una vez y se guarda en
// estados.
func (r *revision) estadoDelIntermedio(
	nombre, base, dir string, estados map[string]estadoDeDirectorio,
) (estadoDeDirectorio, error) {
	if estado, examinado := estados[dir]; examinado {
		return estado, nil
	}

	estado, err := r.estadoDelIntermedio(nombre, base, path.Dir(dir), estados)
	if err != nil {
		return estado, err
	}

	if estado == directorioReal {
		ruta := path.Join(base, dir)

		entrada, err := r.disco.Examinar(ruta)
		if err != nil {
			return directorioRoto, err
		}

		switch entrada.Tipo {
		case EntradaDirectorio:
		case EntradaAusente:
			estado = directorioAusente
		case EntradaEnlace, EntradaFichero, EntradaOtra:
			estado = directorioRoto
			r.anotarNoDirectorio(nombre, ruta, entrada)
		}
	}

	estados[dir] = estado

	return estado, nil
}

// revisarFichero revisa el fichero declarado de ruta, cuyo directorio es real:
// tiene que ser un fichero regular con la huella declarada. Solo se abre un
// fichero regular.
func (r *revision) revisarFichero(nombre, ruta, huellaDeclarada string) error {
	entrada, err := r.disco.Examinar(ruta)
	if err != nil {
		return err
	}

	switch entrada.Tipo {
	case EntradaAusente:
		r.anotar(HallazgoFicheroEditado, ruta, sinRetirar, nombre)
	case EntradaDirectorio:
		r.anotar(HallazgoFicheroEditado, ruta, conRmR, nombre)
	case EntradaEnlace, EntradaOtra:
		r.anotar(HallazgoFicheroEditado, ruta, conRm, nombre)
	case EntradaFichero:
		huella, err := r.disco.Huella(ruta)
		if err != nil {
			return err
		}

		if huella != huellaDeclarada {
			r.anotar(HallazgoFicheroEditado, ruta, conRm, nombre)
		}
	}

	return nil
}

// revisarHost revisa la entrada declarada de la skill nombre en el host: si
// algún directorio de la raíz al de skills del host no es un directorio real,
// no se lee a través de él y la entrada cuenta como que falta.
func (r *revision) revisarHost(nombre, host string, declarada EntradaDeHost) error {
	ruta := r.ambito.RutaDeHost(host, nombre)

	hayHost, err := r.skillsDelHostReal(host)
	if err != nil {
		return err
	}

	if !hayHost {
		r.anotar(HallazgoEnlaceAOtroSitio, ruta, sinRetirar, nombre)

		return nil
	}

	entrada, err := r.disco.Examinar(ruta)
	if err != nil {
		return err
	}

	switch entrada.Tipo {
	case EntradaAusente:
		r.anotar(HallazgoEnlaceAOtroSitio, ruta, sinRetirar, nombre)
	case EntradaEnlace:
		r.revisarEnlaceDeHost(nombre, host, ruta, entrada, declarada.Modo)
	case EntradaDirectorio:
		if declarada.Modo == ModoEnlace {
			r.anotar(HallazgoEnlaceAOtroSitio, ruta, conRmR, nombre)

			return nil
		}

		return r.revisarCopia(nombre, host, ruta, declarada)
	case EntradaFichero, EntradaOtra:
		r.anotar(HallazgoEnlaceAOtroSitio, ruta, conRm, nombre)
	}

	return nil
}

// revisarEnlaceDeHost revisa la entrada del host de ruta, que es un enlace:
// solo es lo declarado si se declaró enlace y su destino literal es el de
// FR-021 para ese host; si no, está en otro sitio, y se retira. Siéndolo,
// cuelga si no resuelve, y su orden solo reinstala.
func (r *revision) revisarEnlaceDeHost(nombre, host, ruta string, entrada Entrada, modo Modo) {
	switch {
	case entrada.Destino != destinoDeHost(host, nombre) || modo != ModoEnlace:
		r.anotar(HallazgoEnlaceAOtroSitio, ruta, conRm, nombre)
	case !entrada.Resuelve:
		r.anotar(HallazgoEnlaceColgando, ruta, sinRetirar, nombre)
	}
}

// revisarCopia revisa la copia declarada de ruta en el host, que sigue siendo
// un directorio real: sus ficheros declarados, como los del directorio neutro,
// y si el Enlazador ya puede crear el enlace en el directorio de skills del
// host (FR-069).
func (r *revision) revisarCopia(nombre, host, ruta string, declarada EntradaDeHost) error {
	if err := r.revisarFicheros(nombre, ruta, relativas(declarada.Ficheros, declarada.Ruta+"/")); err != nil {
		return err
	}

	disponible, preguntado := r.sondas[host]
	if !preguntado {
		var err error

		disponible, err = r.enlazador.Disponible(r.ambito.SkillsDelHost(host))
		if err != nil {
			return err
		}

		r.sondas[host] = disponible
	}

	if disponible {
		r.anotar(HallazgoCopia, ruta, sinRetirar, nombre)
	}

	return nil
}

// skillsDelHostReal dice si cada directorio de la raíz al de skills del host
// es un directorio real, examinados de arriba abajo, sin seguir enlaces, una
// sola vez y cada uno solo si el de encima lo es.
func (r *revision) skillsDelHostReal(host string) (bool, error) {
	if esReal, examinado := r.hostsReales[host]; examinado {
		return esReal, nil
	}

	hayHost := true

	for _, dir := range r.ambito.cadenaDelHost(host) {
		entrada, err := r.disco.Examinar(dir)
		if err != nil {
			return false, err
		}

		if entrada.Tipo != EntradaDirectorio {
			hayHost = false

			break
		}
	}

	r.hostsReales[host] = hayHost

	return hayHost, nil
}

// revisarVersiones compara con version la del manifiesto y, si coincide, la de
// cada skill de nombres, las declaradas y empotradas, según FR-077. Si la del
// manifiesto difiere, hay un único hallazgo, cuya orden las reinstala todas; y
// si no, uno por cada skill que difiere.
func (r *revision) revisarVersiones(nombres []string, version string) {
	if !MismaVersion(r.manifiesto.Version, version) {
		r.anotar(HallazgoVersionDistinta, r.ambito.RutaDelManifiesto(), sinRetirar, nombres...)

		return
	}

	for _, nombre := range nombres {
		if !MismaVersion(r.manifiesto.Skills[nombre].Version, version) {
			r.anotar(HallazgoVersionDistinta, r.ambito.RutaDeSkill(nombre), sinRetirar, nombre)
		}
	}
}
