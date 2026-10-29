package instalacion

import (
	"maps"
	"slices"
	"strings"
)

// Retiradas son los nombres de skill que el binario ya no empotra y, por cada
// uno, las skills empotradas que heredan sus preguntas (FR-025; ADR 0048). La
// lista la lleva la composición, que sabe qué se empotra, y llega a
// Planificar como argumento: el dominio no conoce ningún nombre de skill.
type Retiradas map[string][]string

// Herederas son las que heredan las preguntas del nombre retirado, como las
// nombran el aviso y la salida legible: «a», «a y b», «a, b y c»; nada si el
// nombre no es retirado.
func (r Retiradas) Herederas(nombre string) string {
	herederas := r[nombre]
	if len(herederas) < 2 {
		return strings.Join(herederas, "")
	}

	ultima := len(herederas) - 1

	return strings.Join(herederas[:ultima], ", ") + " y " + herederas[ultima]
}

// avisoDeRetirada es el aviso de la skill de nombre retirado de ruta, que no
// se retira, con las herederas que la sustituyen
// (contracts/instalacion-retiradas.md §2). Su verbo va en dos literales porque
// misspell, con su diccionario inglés, lee su raíz como una errata de
// «install».
func avisoDeRetirada(ruta, herederas string) string {
	return ruta + " no es lo que insta" + "ló kitlegal: su protocolo puede activarse junto al de " + herederas +
		", que la sustituye; retirarla es cosa tuya."
}

// retiradaExaminada es lo que se sabe de una skill que el manifiesto declara
// con un nombre retirado: su entrada del manifiesto, las herederas que la
// sustituyen y lo que hay en su directorio y en su ruta de cada host del
// ámbito.
type retiradaExaminada struct {
	nombre    string
	declarada SkillDeclarada
	herederas string
	neutro    parteRetirada
	hosts     map[string]parteRetirada
}

// parteRetirada es lo examinado de una parte de una skill de nombre retirado,
// su directorio o su entrada en un host: si hay algo en su ruta, si es lo que
// declara el manifiesto y, si es un directorio que lo es, todo lo que hay en
// él, que es lo que se retira.
type parteRetirada struct {
	existe    bool
	intacta   bool
	contenido examinado
}

// presente dice si queda algo de la skill en el disco: su directorio o una
// entrada en la ruta de alguno de los hosts.
func (r retiradaExaminada) presente() bool {
	existe := r.neutro.existe
	for _, host := range r.hosts {
		existe = existe || host.existe
	}

	return existe
}

// intacta dice si la skill es lo que declara el manifiesto: su directorio y
// cada una de sus entradas de host, también la que no declara y no tiene que
// existir (FR-026).
func (r retiradaExaminada) intacta() bool {
	intacta := r.neutro.intacta
	for _, host := range r.hosts {
		intacta = intacta && host.intacta
	}

	return intacta
}

// examinarLasRetiradas examina, en orden de nombre, cada skill que el
// manifiesto del ámbito declara con un nombre de retiradas, sin mirar qué pide
// la invocación (FR-026), y guarda lo examinado. No anota ningún conflicto: lo
// que no es lo que declara el manifiesto solo hace que no se retire (FR-027).
// Un fallo del Disco se devuelve tal cual.
func (c *comprobacion) examinarLasRetiradas(retiradas Retiradas) error {
	for _, nombre := range slices.Sorted(maps.Keys(c.manifiesto.Skills)) {
		if _, retirado := retiradas[nombre]; !retirado {
			continue
		}

		examinada := retiradaExaminada{
			nombre:    nombre,
			declarada: c.manifiesto.Skills[nombre],
			herederas: retiradas.Herederas(nombre),
			hosts:     map[string]parteRetirada{},
		}

		var err error

		examinada.neutro, err = c.examinarDirectorioRetirado(c.ambito.RutaDeSkill(nombre),
			relativas(examinada.declarada.Ficheros, nombre+"/"))
		if err != nil {
			return err
		}

		for _, host := range c.ambito.Hosts() {
			examinada.hosts[host], err = c.examinarEntradaRetirada(host, nombre, entradaDeclarada(examinada.declarada, host))
			if err != nil {
				return err
			}
		}

		c.retiradas = append(c.retiradas, examinada)
	}

	return nil
}

// examinarDirectorioRetirado examina base, el directorio de una skill de
// nombre retirado o su copia de host, sin seguir ningún enlace: está intacto si
// es un directorio real con exactamente los ficheros declarados, relativos a
// él, cada uno un fichero regular con su huella, y cada directorio intermedio
// de ellos. Es lo que se exige a una copia de host que se retira entera, en un
// examen aparte: lo que allí sería un conflicto, aquí solo dice que no es lo
// que dejó kitlegal; y lo declarado que falta tampoco lo es.
func (c *comprobacion) examinarDirectorioRetirado(base string, declarados map[string]string) (parteRetirada, error) {
	entrada, err := c.disco.Examinar(base)
	if err != nil {
		return parteRetirada{}, err
	}

	switch entrada.Tipo {
	case EntradaAusente:
		return parteRetirada{}, nil
	case EntradaFichero, EntradaEnlace, EntradaOtra:
		return parteRetirada{existe: true}, nil
	case EntradaDirectorio:
	}

	examen := &comprobacion{disco: c.disco, enConflicto: map[string]ClaseDeConflicto{}}

	contenido, err := examen.comprobarCopiaQueSeRetira(base, declarados)
	if err != nil {
		return parteRetirada{}, err
	}

	intacta := len(examen.enConflicto) == 0
	for rel := range declarados {
		_, existe := contenido[rel]
		intacta = intacta && existe
	}

	return parteRetirada{existe: true, intacta: intacta, contenido: contenido}, nil
}

// examinarEntradaRetirada examina la entrada de la skill de nombre retirado en
// el host, con declarada su entrada declarada, si la tiene: está intacta si es
// el enlace de FR-021, la copia con sus ficheros y sus huellas, o nada si no
// declara ninguna. En un host cuyo directorio de skills no es un directorio
// real no puede haber nada sin seguir un enlace, y no se examina.
func (c *comprobacion) examinarEntradaRetirada(host, nombre string, declarada *EntradaDeHost) (parteRetirada, error) {
	if !c.conEntradas[host] {
		return parteRetirada{intacta: declarada == nil}, nil
	}

	ruta := c.ambito.RutaDeHost(host, nombre)

	if declarada != nil && declarada.Modo == ModoCopia {
		return c.examinarDirectorioRetirado(ruta, relativas(declarada.Ficheros, declarada.Ruta+"/"))
	}

	entrada, err := c.disco.Examinar(ruta)
	if err != nil {
		return parteRetirada{}, err
	}

	existe := entrada.Tipo != EntradaAusente
	if declarada == nil {
		return parteRetirada{existe: existe, intacta: !existe}, nil
	}

	enlace := entrada.Tipo == EntradaEnlace && entrada.Destino == destinoDeHost(host, nombre)

	return parteRetirada{existe: existe, intacta: enlace}, nil
}

// planificarRetirada añade al plan lo de la skill de nombre retirado
// (contracts/instalacion-retiradas.md §2): si no queda nada de ella, su
// entrada sale del manifiesto y no se nombra; si está intacta, se retira
// entera —sus ficheros, después cada directorio suyo de abajo arriba, después
// cada entrada de host que declara— y su entrada sale del manifiesto; y si
// no, no se toca nada, su entrada se conserva y sale con su aviso.
func (p *planificacion) planificarRetirada(r retiradaExaminada) {
	ruta := p.c.ambito.RutaDeSkill(r.nombre)

	switch {
	case !r.presente():
		p.quitadas = append(p.quitadas, r.nombre)
	case r.intacta():
		p.plan.Retirar = append(p.plan.Retirar, retiradaDeCopia(ruta, r.neutro.contenido)...)

		enlaces := []Enlace{}

		for _, host := range p.c.ambito.Hosts() {
			entrada, declarada := r.declarada.Hosts[host]
			if !declarada {
				continue
			}

			enElHost := p.c.ambito.RutaDeHost(host, r.nombre)

			if entrada.Modo == ModoCopia {
				p.plan.Retirar = append(p.plan.Retirar, retiradaDeCopia(enElHost, r.hosts[host].contenido)...)
			} else {
				p.plan.Retirar = append(p.plan.Retirar, enElHost)
			}

			enlaces = append(enlaces, Enlace{Host: host, Ruta: enElHost, Modo: entrada.Modo})
		}

		p.quitadas = append(p.quitadas, r.nombre)
		p.plan.Skills = append(p.plan.Skills,
			SkillInstalada{Nombre: r.nombre, Ruta: ruta, Estado: EstadoRetirada, Enlaces: enlaces})
	default:
		p.plan.Skills = append(p.plan.Skills, SkillInstalada{
			Nombre:  r.nombre,
			Ruta:    ruta,
			Estado:  EstadoSinRetirar,
			Enlaces: []Enlace{},
			Aviso:   avisoDeRetirada(ruta, r.herederas),
		})
	}
}
