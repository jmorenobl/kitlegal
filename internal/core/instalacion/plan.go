package instalacion

import (
	"errors"
	"fmt"
	"maps"
	"path"
	"slices"
	"strings"
)

// errNoEsFicheroRegular es lo que hace ilegible un kitlegal.json que existe y
// no es un fichero regular (contracts/manifiesto.md §3, regla 1).
var errNoEsFicheroRegular = errors.New("no es un fichero regular")

// ComprobarConflictos comprueba, antes de escribir nada, todas las entradas
// que install va a crear, cambiar o retirar para las skills del pedido, en el
// directorio neutro y en el host claude si se enlaza, y las nombra todas si
// alguna no es suya (FR-040 a FR-043, FR-047; data-model §4):
//
//  1. el ámbito: cada guarda es un directorio real o no existe, el manifiesto
//     es legible y, con --dir, no declara entradas de host; con cualquiera de
//     estos no se sabe qué es de quién y no se examina ninguna skill. Si se
//     enlaza en el host, .claude y .claude/skills tienen que ser, cada uno, un
//     directorio real o no existir; sin --host, un .claude que no es un
//     directorio real cuenta como ausente y no se enlaza (FR-022, FR-023,
//     FR-026, FR-027);
//  2. el directorio neutro de cada skill pedida y, si está declarada, cada
//     directorio intermedio y cada fichero suyo, empotrado o declarado (§4.2);
//  3. su entrada de host, si se enlaza y .claude/skills es un directorio real:
//     el enlace de FR-021 se adopta aunque cuelgue, y una copia declarada se
//     examina entera si va a pasar a enlace, o como el directorio neutro si se
//     mantiene, según diga el Enlazador por .claude/skills, una sola vez
//     (§4.3; data-model §3).
//
// Las skills declaradas que no se piden o que el binario no empotra no se
// examinan (§4.4). Nada se examina por debajo de una entrada que no es un
// directorio real, lo que no es un fichero regular no se abre, y cada entrada
// cae como mucho en una clase: la primera en ese orden (FR-028, FR-041).
//
// Con algún conflicto devuelve un *ErrorDeConflictos; sin ninguno, nil. Un
// fallo del Disco o del Enlazador no es un conflicto: se devuelve tal cual,
// como una skill del pedido que no está entre las empotradas, que se rechaza
// antes de examinar nada (FR-036).
func ComprobarConflictos(disco Disco, enlazador Enlazador, pedido Pedido, empotradas []SkillEmpotrada) error {
	pedidas, err := skillsDelPedido(pedido.Skills, empotradas)
	if err != nil {
		return err
	}

	c := &comprobacion{
		disco:       disco,
		enlazador:   enlazador,
		ambito:      pedido.Ambito,
		enConflicto: map[string]ClaseDeConflicto{},
		sondas:      map[string]bool{},
	}

	manifiesto, neutroReal, err := c.comprobarElNeutro()
	if err != nil {
		return err
	}

	ambitoIlegible := len(c.enConflicto) > 0

	hostReal, err := c.comprobarElHost(pedido.HostClaude)
	if err != nil {
		return err
	}

	if !ambitoIlegible {
		err = c.comprobarLasSkills(pedidas, manifiesto, neutroReal, hostReal)
		if err != nil {
			return err
		}
	}

	if len(c.enConflicto) > 0 {
		return nuevoErrorDeConflictos(c.enConflicto)
	}

	return nil
}

// skillsDelPedido son las skills empotradas de cada nombre, en el mismo orden,
// o un error con el primero que no es de ninguna.
func skillsDelPedido(nombres []string, empotradas []SkillEmpotrada) ([]SkillEmpotrada, error) {
	pedidas := make([]SkillEmpotrada, 0, len(nombres))

	for _, nombre := range nombres {
		i := slices.IndexFunc(empotradas, func(skill SkillEmpotrada) bool { return skill.Nombre == nombre })
		if i < 0 {
			return nil, fmt.Errorf("la skill %q no la lleva este binario: install no la pide nunca", nombre)
		}

		pedidas = append(pedidas, empotradas[i])
	}

	return pedidas, nil
}

// comprobacion es el estado de una comprobación: lo que examina, dónde, cada
// conflicto que lleva encontrado y lo que ya respondió el Enlazador.
type comprobacion struct {
	disco     Disco
	enlazador Enlazador
	ambito    Ambito
	// enConflicto es la clase de cada entrada en conflicto, por su ruta.
	enConflicto map[string]ClaseDeConflicto
	// sondas es la respuesta de Disponible por directorio, que se pregunta
	// una sola vez en la invocación.
	sondas map[string]bool
}

// anotar deja la entrada de ruta en conflicto con clase, salvo que ya lo
// esté: cada entrada cae como mucho en una clase, la primera en el orden de
// comprobación (FR-041).
func (c *comprobacion) anotar(ruta string, clase ClaseDeConflicto) {
	if _, anotada := c.enConflicto[ruta]; !anotada {
		c.enConflicto[ruta] = clase
	}
}

// estadoDeDirectorio es lo que se sabe de una ruta que tiene que ser un
// directorio real o no existir.
type estadoDeDirectorio int

// Los tres estados de una ruta que tiene que ser un directorio.
const (
	// directorioReal es un directorio real: lo de debajo se examina.
	directorioReal estadoDeDirectorio = iota
	// directorioAusente no existe, así que no existe nada por debajo.
	directorioAusente
	// directorioRoto existe y no es un directorio real: es el conflicto, y lo
	// de debajo no se examina.
	directorioRoto
)

// examinarDirectorio examina ruta, que tiene que ser un directorio real o no
// existir, y si es otra cosa la anota como ruta que no es directorio.
func (c *comprobacion) examinarDirectorio(ruta string) (estadoDeDirectorio, error) {
	entrada, err := c.disco.Examinar(ruta)
	if err != nil {
		return directorioRoto, err
	}

	switch entrada.Tipo {
	case EntradaDirectorio:
		return directorioReal, nil
	case EntradaAusente:
		return directorioAusente, nil
	default:
		c.anotar(ruta, ConflictoRutaQueNoEsDirectorio)

		return directorioRoto, nil
	}
}

// comprobarElNeutro aplica las tres primeras filas de data-model §4.1 —las
// guardas del ámbito, el manifiesto ilegible y, con --dir, el manifiesto con
// entradas de host— y devuelve el manifiesto, vacío si no hay o no se puede
// leer, y si el directorio neutro es un directorio real.
func (c *comprobacion) comprobarElNeutro() (Manifiesto, bool, error) {
	for _, guarda := range c.ambito.Guardas() {
		estado, err := c.examinarDirectorio(guarda)
		if err != nil || estado != directorioReal {
			return Manifiesto{}, false, err
		}
	}

	ruta := c.ambito.RutaDelManifiesto()

	manifiesto, err := leerManifiestoDe(c.disco, ruta)

	var ilegible *ManifiestoIlegible
	if errors.As(err, &ilegible) {
		c.anotar(ruta, ConflictoManifiestoIlegible)

		return Manifiesto{}, true, nil
	}

	if err != nil {
		return Manifiesto{}, true, err
	}

	if !c.ambito.ConHosts() && conEntradasDeHost(manifiesto) {
		c.anotar(ruta, ConflictoManifiestoConEntradasDeHost)
	}

	return manifiesto, true, nil
}

// leerManifiestoDe lee el manifiesto de ruta sin seguir un enlace (FR-028):
// sin él, el Manifiesto vacío; si existe y no es un fichero regular, no se
// puede leer o no respeta la forma de contracts/manifiesto.md, un
// *ManifiestoIlegible (FR-035); y si no se puede examinar, el error del
// Disco.
func leerManifiestoDe(disco Disco, ruta string) (Manifiesto, error) {
	entrada, err := disco.Examinar(ruta)
	if err != nil {
		return Manifiesto{}, err
	}

	if entrada.Tipo == EntradaAusente {
		return Manifiesto{}, nil
	}

	if entrada.Tipo != EntradaFichero {
		return Manifiesto{}, &ManifiestoIlegible{causa: errNoEsFicheroRegular}
	}

	contenido, err := disco.Leer(ruta)
	if err != nil {
		return Manifiesto{}, &ManifiestoIlegible{causa: err}
	}

	return LeerManifiesto(contenido)
}

// conEntradasDeHost dice si el manifiesto declara alguna entrada de host.
func conEntradasDeHost(manifiesto Manifiesto) bool {
	for _, skill := range manifiesto.Skills {
		if skill.Claude != nil {
			return true
		}
	}

	return false
}

// comprobarElHost aplica las dos últimas filas de data-model §4.1 y dice si
// hay entradas de host que examinar: solo si se enlaza en el host (§4.3) y
// .claude/skills es un directorio real. Si se enlaza y falta .claude o
// .claude/skills, no existe ninguna entrada de host; si alguno de los dos
// existe y no es un directorio real, es el conflicto y nada por debajo se
// examina.
func (c *comprobacion) comprobarElHost(conHostClaude bool) (bool, error) {
	if !c.ambito.ConHosts() {
		return false, nil
	}

	claude := c.ambito.DirectorioDelHost()

	entrada, err := c.disco.Examinar(claude)
	if err != nil {
		return false, err
	}

	if entrada.Tipo != EntradaDirectorio {
		if conHostClaude && entrada.Tipo != EntradaAusente {
			c.anotar(claude, ConflictoRutaQueNoEsDirectorio)
		}

		return false, nil
	}

	estado, err := c.examinarDirectorio(c.ambito.SkillsDelHost())

	return estado == directorioReal, err
}

// comprobarLasSkills examina cada skill pedida en el directorio neutro, si es
// un directorio real, y en el host, si hay entradas de host que examinar.
func (c *comprobacion) comprobarLasSkills(
	pedidas []SkillEmpotrada, manifiesto Manifiesto, neutroReal, hostReal bool,
) error {
	for _, skill := range pedidas {
		declarada, esDeclarada := manifiesto.Skills[skill.Nombre]

		if neutroReal {
			if err := c.comprobarEnElNeutro(skill, declarada, esDeclarada); err != nil {
				return err
			}
		}

		if hostReal {
			if err := c.comprobarEnElHost(skill, declarada.Claude); err != nil {
				return err
			}
		}
	}

	return nil
}

// comprobarEnElNeutro aplica a la skill la primera tabla de data-model §4.2 y,
// si su directorio es real y está declarada, la segunda a sus ficheros.
func (c *comprobacion) comprobarEnElNeutro(skill SkillEmpotrada, declarada SkillDeclarada, esDeclarada bool) error {
	ruta := c.ambito.RutaDeSkill(skill.Nombre)

	entrada, err := c.disco.Examinar(ruta)
	if err != nil {
		return err
	}

	switch entrada.Tipo {
	case EntradaAusente:
	case EntradaDirectorio:
		if esDeclarada {
			return c.comprobarFicheros(ruta, relativas(declarada.Ficheros, skill.Nombre+"/"), skill.Ficheros)
		}

		c.anotar(ruta, ConflictoCarpetaAjena)
	case EntradaEnlace:
		c.anotar(ruta, claseDelEnlace(entrada))
	case EntradaFichero, EntradaOtra:
		c.anotar(ruta, ConflictoFichero)
	}

	return nil
}

// comprobarEnElHost aplica a la entrada de host de la skill la tabla de
// data-model §4.3, con host su entrada declarada, si la tiene.
func (c *comprobacion) comprobarEnElHost(skill SkillEmpotrada, host *EntradaDeHost) error {
	ruta := c.ambito.RutaDeHost(skill.Nombre)

	entrada, err := c.disco.Examinar(ruta)
	if err != nil {
		return err
	}

	switch entrada.Tipo {
	case EntradaAusente:
	case EntradaEnlace:
		if entrada.Destino != destinoDeHost(skill.Nombre) {
			c.anotar(ruta, claseDelEnlace(entrada))
		}
	case EntradaDirectorio:
		return c.comprobarDirectorioDeHost(ruta, skill, host)
	case EntradaFichero, EntradaOtra:
		c.anotar(ruta, ConflictoFichero)
	}

	return nil
}

// destinoDeHost es el destino literal del enlace de host de la skill nombre,
// ../../.agents/skills/<nombre> (FR-021).
func destinoDeHost(nombre string) string {
	return path.Join("../..", directorioNeutro, nombre)
}

// claseDelEnlace es la de un enlace en conflicto: a otro sitio si resuelve y
// roto si cuelga o está en un ciclo, nunca las dos (FR-041 (c) y (d)).
func claseDelEnlace(entrada Entrada) ClaseDeConflicto {
	if entrada.Resuelve {
		return ConflictoEnlaceAOtroSitio
	}

	return ConflictoEnlaceRoto
}

// comprobarDirectorioDeHost aplica las filas de data-model §4.3 de una entrada
// de host que es un directorio real: sin declarar o declarada enlace, es una
// carpeta ajena; declarada copia, pasa a enlace si el Enlazador está
// disponible en .claude/skills, y entonces todo lo que contiene tiene que
// estar declarado e intacto, o se mantiene y se comprueba como el directorio
// neutro.
func (c *comprobacion) comprobarDirectorioDeHost(ruta string, skill SkillEmpotrada, host *EntradaDeHost) error {
	if host == nil || host.Modo != ModoCopia {
		c.anotar(ruta, ConflictoCarpetaAjena)

		return nil
	}

	disponible, err := c.disponible(c.ambito.SkillsDelHost())
	if err != nil {
		return err
	}

	declarados := relativas(host.Ficheros, host.Ruta+"/")

	if disponible {
		return c.comprobarCopiaQueSeRetira(ruta, declarados)
	}

	return c.comprobarFicheros(ruta, declarados, skill.Ficheros)
}

// disponible es la respuesta del Enlazador por directorio, que se le pregunta
// una sola vez en la invocación (data-model §3).
func (c *comprobacion) disponible(directorio string) (bool, error) {
	if respuesta, preguntado := c.sondas[directorio]; preguntado {
		return respuesta, nil
	}

	respuesta, err := c.enlazador.Disponible(directorio)
	if err != nil {
		return false, err
	}

	c.sondas[directorio] = respuesta

	return respuesta, nil
}

// relativas son los ficheros declarados con la ruta relativa al directorio de
// la skill o de su copia, sin prefijo, que el manifiesto garantiza.
func relativas(ficheros map[string]string, prefijo string) map[string]string {
	sinPrefijo := make(map[string]string, len(ficheros))
	for ruta, huella := range ficheros {
		sinPrefijo[strings.TrimPrefix(ruta, prefijo)] = huella
	}

	return sinPrefijo
}

// comprobarFicheros aplica la segunda tabla de data-model §4.2 a los ficheros
// de base, el directorio real de una skill declarada o de su copia de host
// que se mantiene: declarados, relativos a base con su huella, y empotrados.
// Primero, cada directorio intermedio de todos ellos tiene que ser real o no
// existir, y lo que cuelga de uno que no lo es no se examina; después, un
// fichero declarado tiene que ser un fichero regular con su huella, o faltar,
// y uno empotrado que no está declarado tiene que faltar. Lo que no está
// declarado ni empotrado no se examina (FR-047).
func (c *comprobacion) comprobarFicheros(base string, declarados map[string]string, empotrados []FicheroEmpotrado) error {
	rutas := slices.Collect(maps.Keys(declarados))
	for _, fichero := range empotrados {
		rutas = append(rutas, fichero.Ruta)
	}

	slices.Sort(rutas)
	rutas = slices.Compact(rutas)

	estados := map[string]estadoDeDirectorio{".": directorioReal}

	for _, rel := range rutas {
		if _, err := c.estadoDelDirectorio(base, path.Dir(rel), estados); err != nil {
			return err
		}
	}

	for _, rel := range rutas {
		if estados[path.Dir(rel)] != directorioReal {
			continue
		}

		ruta := path.Join(base, rel)

		entrada, err := c.disco.Examinar(ruta)
		if err != nil {
			return err
		}

		huella, declarado := declarados[rel]
		if err := c.clasificarFichero(ruta, entrada, huella, declarado); err != nil {
			return err
		}
	}

	return nil
}

// estadoDelDirectorio es el del directorio dir, relativo a base, que se
// examina solo si el de encima es real: bajo uno que falta, falta, y bajo uno
// que no es un directorio real, no se examina. Cada uno se examina una vez y
// se guarda en estados.
func (c *comprobacion) estadoDelDirectorio(
	base, dir string, estados map[string]estadoDeDirectorio,
) (estadoDeDirectorio, error) {
	if estado, examinado := estados[dir]; examinado {
		return estado, nil
	}

	estado, err := c.estadoDelDirectorio(base, path.Dir(dir), estados)
	if err != nil {
		return estado, err
	}

	if estado == directorioReal {
		estado, err = c.examinarDirectorio(path.Join(base, dir))
		if err != nil {
			return estado, err
		}
	}

	estados[dir] = estado

	return estado, nil
}

// clasificarFichero anota la entrada de ruta si es un conflicto: una que no
// está declarada, de cualquier tipo, es un fichero ajeno; una declarada que no
// es un fichero regular, o cuya huella no es la declarada, un fichero
// editado; lo que falta no es conflicto. Solo se abre un fichero regular.
func (c *comprobacion) clasificarFichero(ruta string, entrada Entrada, huellaDeclarada string, declarado bool) error {
	switch {
	case entrada.Tipo == EntradaAusente:
	case !declarado:
		c.anotar(ruta, ConflictoFicheroAjeno)
	case entrada.Tipo != EntradaFichero:
		c.anotar(ruta, ConflictoFicheroEditado)
	default:
		huella, err := c.disco.Huella(ruta)
		if err != nil {
			return err
		}

		if huella != huellaDeclarada {
			c.anotar(ruta, ConflictoFicheroEditado)
		}
	}

	return nil
}

// copiaDeHost es una copia de host declarada que va a pasar a enlace: su
// directorio, sus ficheros declarados relativos a él con su huella y cada
// directorio intermedio de esos ficheros.
type copiaDeHost struct {
	base        string
	declarados  map[string]string
	intermedios map[string]bool
}

// comprobarCopiaQueSeRetira exige que todo lo que contiene la copia de base,
// que se va a retirar entera para crear el enlace, esté declarado e intacto
// (FR-046): una entrada no declarada es un fichero ajeno, un fichero editado
// lo es, y un intermedio que no es un directorio real, una ruta que no es
// directorio; lo declarado que falta no es conflicto.
func (c *comprobacion) comprobarCopiaQueSeRetira(base string, declarados map[string]string) error {
	copia := copiaDeHost{base: base, declarados: declarados, intermedios: map[string]bool{}}

	for rel := range declarados {
		for dir := path.Dir(rel); dir != "."; dir = path.Dir(dir) {
			copia.intermedios[dir] = true
		}
	}

	return c.recorrerCopia(copia, ".")
}

// recorrerCopia examina cada entrada del directorio dir de la copia, relativo
// a su base, y entra en cada una que es un intermedio y un directorio real.
func (c *comprobacion) recorrerCopia(copia copiaDeHost, dir string) error {
	nombres, err := c.disco.Nombres(path.Join(copia.base, dir))
	if err != nil {
		return err
	}

	for _, nombre := range slices.Sorted(slices.Values(nombres)) {
		rel := path.Join(dir, nombre)
		ruta := path.Join(copia.base, rel)

		entrada, err := c.disco.Examinar(ruta)
		if err != nil {
			return err
		}

		switch {
		case !copia.intermedios[rel]:
			huella, declarado := copia.declarados[rel]
			err = c.clasificarFichero(ruta, entrada, huella, declarado)
		case entrada.Tipo == EntradaDirectorio:
			err = c.recorrerCopia(copia, rel)
		case entrada.Tipo != EntradaAusente:
			c.anotar(ruta, ConflictoRutaQueNoEsDirectorio)
		}

		if err != nil {
			return err
		}
	}

	return nil
}
