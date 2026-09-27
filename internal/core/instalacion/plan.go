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

// Plan es lo que install va a hacer, sin ningún conflicto, en las cuatro fases
// del orden de aplicación de research.md D7 (data-model §5): retirar, enlazar,
// escribir el manifiesto final y escribir los ficheros. Tras un fallo en
// cualquiera de sus operaciones, volver a planificar sobre lo que quedó no da
// ningún conflicto y completa la instalación (FR-044). Sin aplicar, es lo que
// describe --dry-run (FR-048). Con nada que cambiar, todas sus listas están
// vacías y el manifiesto no se escribe (FR-033, FR-045).
//
// Ningún fichero se escribe con permiso de ejecución (FR-015): fuera de un
// directorio o un enlace, el plan solo crea ficheros regulares con
// EscribirFichero, que no lo da, y no tiene forma de pedir otro modo.
type Plan struct {
	// Skills son las pedidas, en orden de nombre, tal como quedan si cada
	// entrada de host se crea en el modo previsto: la salida de la orden y la
	// de --dry-run (FR-048, FR-051). Una skill declarada que el binario no
	// empotra no está nunca (FR-036). Es una lista, aunque esté vacía.
	Skills []SkillInstalada

	// Retirar es la fase 1: cada fichero declarado e intacto que se reescribe
	// o que ya no se empotra, en el directorio neutro o en una copia de host
	// que se mantiene, y cada copia de host que pasa a enlace, entera —sus
	// ficheros, después cada directorio suyo de abajo arriba y el de la copia
	// el último—, en ese orden (FR-046).
	Retirar []string

	// Enlazar es la fase 2.
	Enlazar Enlazado

	// Manifiesto es la fase 3.
	Manifiesto ManifiestoFinal

	// Escribir es la fase 4: cada directorio que falta, después del de
	// encima, y cada fichero nuevo o que se reescribe, del directorio neutro y
	// de las copias de host que se crean o se mantienen (FR-014, FR-046).
	Escribir Escrituras
}

// Enlazado es la fase 2 del plan: lo que falta hasta .claude/skills y cada
// entrada de host que se crea como enlace (FR-021, FR-023).
type Enlazado struct {
	// DirectoriosQueFaltan son los que faltan hasta .claude/skills incluido,
	// de arriba abajo, si hay que crear alguna entrada de host, enlace o copia
	// (FR-014, FR-023).
	DirectoriosQueFaltan []string

	// Enlaces son las entradas de host que se crean como enlace, en orden de
	// skill.
	Enlaces []EnlaceNuevo
}

// EnlaceNuevo es una entrada de host que se crea como el enlace de FR-021, con
// el recurso de copia que la sustituye si el Enlazador no lo puede crear
// (FR-024).
type EnlaceNuevo struct {
	// Skill es el nombre de su skill.
	Skill string

	// Ruta es la de la entrada, <raíz>/.claude/skills/<skill>.
	Ruta string

	// Destino es el literal del enlace, ../../.agents/skills/<skill>.
	Destino string

	// Copia es lo que se escribe en la fase 4 si el enlace no se puede crear:
	// el directorio de la entrada, con los que le falten dentro, y cada
	// fichero empotrado. Entonces el manifiesto final declara la entrada en
	// copia (ManifiestoFinal.Contenido).
	Copia Escrituras
}

// ManifiestoFinal es la fase 3 del plan: lo que falta hasta el directorio
// neutro, que hay que crear antes porque el manifiesto vive en él, y el
// manifiesto que declara el estado final (FR-014, FR-030).
type ManifiestoFinal struct {
	// DirectoriosQueFaltan son los que faltan hasta el directorio neutro
	// incluido, de arriba abajo, si el manifiesto cambia.
	DirectoriosQueFaltan []string

	// Ruta es la del manifiesto, kitlegal.json en el directorio neutro.
	Ruta string

	// actual es el manifiesto que hay, vacío si no hay ninguno.
	actual Manifiesto
	// version es la del binario, que declara cada skill pedida.
	version string
	// declaradas son las entradas finales de las skills pedidas, cada entrada
	// de host en el modo previsto.
	declaradas map[string]SkillDeclarada
	// copias son, por skill con un enlace que crear, la entrada de host que
	// se declara si el enlace no se puede crear.
	copias map[string]*EntradaDeHost
}

// Contenido son los bytes canónicos del manifiesto final, con la entrada de
// host de cada skill de enCopia —las de un enlace que no se pudo crear en la
// fase 2— en copia; un nombre que no es el de ningún enlace del plan no cambia
// nada. Si ninguna entrada de una skill pedida cambia, es nil: el manifiesto
// no se escribe y queda byte a byte igual (FR-033, FR-045). Las demás
// entradas se conservan tal como están, también las de las skills que el
// binario no empotra (FR-034, FR-036), y la versión de nivel superior pasa a
// ser la del binario (contracts/manifiesto.md §2).
func (m ManifiestoFinal) Contenido(enCopia []string) ([]byte, error) {
	skills := maps.Clone(m.actual.Skills)
	if skills == nil {
		skills = map[string]SkillDeclarada{}
	}

	cambia := false

	for nombre, declarada := range m.declaradas {
		if copia, hay := m.copias[nombre]; hay && slices.Contains(enCopia, nombre) {
			declarada.Claude = copia
		}

		anterior, estaba := skills[nombre]
		cambia = cambia || !estaba || !mismaDeclaracion(anterior, declarada)
		skills[nombre] = declarada
	}

	if !cambia {
		return nil, nil
	}

	return Manifiesto{Version: m.version, Skills: skills}.Bytes()
}

// Escrituras son lo que se crea en una fase: cada directorio que falta, en
// orden y cada uno después del de encima, y después cada fichero.
type Escrituras struct {
	// DirectoriosQueFaltan se crean en este orden, antes que los ficheros.
	DirectoriosQueFaltan []string

	// Ficheros se escriben en este orden.
	Ficheros []Escritura
}

// Escritura es un fichero regular que se escribe con EscribirFichero: su ruta
// y sus bytes, sin ningún modo (FR-015).
type Escritura struct {
	// Ruta es la del fichero.
	Ruta string

	// Contenido son sus bytes, los empotrados.
	Contenido []byte
}

// Planificar decide lo que install hace con las skills del pedido: primero
// comprueba cada conflicto, como ComprobarConflictos, sin escribir nada y con
// el mismo resultado si hay alguno; y, sin ninguno, devuelve el Plan, en el
// que cada skill pedida declara version, la del binario (FR-031). No cambia
// nada en el disco: es todo lo que hace --dry-run (FR-048).
//
// Por cada skill pedida (data-model §4.2 y §4.3; FR-046): en el directorio
// neutro, se escribe lo empotrado que falta o cuya huella declarada no es la
// suya, retirando antes lo intacto que se reescribe, y se retira lo declarado
// e intacto que ya no se empotra; en el host, si se enlaza en esta ejecución,
// el enlace de FR-021 que ya está se adopta, la entrada que falta se crea como
// enlace o como copia, una copia declarada pasa a enlace o se mantiene y se
// actualiza, y, si no se enlaza, la entrada declarada se quita del
// manifiesto sin tocar el disco. Sale «instalada» si el manifiesto no la
// declaraba, «actualizada» si cambia algo en el disco o en su entrada del
// manifiesto, y «sin cambios» si no. Lo que falta hasta el directorio neutro
// y hasta .claude/skills se crea de arriba abajo, desde lo primero que existe,
// que se usa tal cual (FR-014, FR-027).
//
// Si una entrada de host se tiene que crear, el modo previsto sale de
// preguntar a Disponible por el directorio de la sonda: el directorio real
// existente más próximo a .claude/skills dentro del ámbito, donde vivirá la
// entrada, una vez por directorio; si no hay ninguno, porque la raíz del
// ámbito no existe, no se pregunta y se prevé enlace (data-model §3; research.md
// D9). Un fallo del Disco o del Enlazador se devuelve tal cual. Una version
// que el manifiesto no admitiría es un error antes de examinar nada.
func Planificar(
	disco Disco, enlazador Enlazador, pedido Pedido, empotradas []SkillEmpotrada, version string,
) (Plan, error) {
	if err := comprobarVersion(version); err != nil {
		return Plan{}, fmt.Errorf("la versión del binario no se puede declarar en el manifiesto: %w", err)
	}

	c, err := comprobar(disco, enlazador, pedido, empotradas)
	if err != nil {
		return Plan{}, err
	}

	return c.planificar(version)
}

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
	_, err := comprobar(disco, enlazador, pedido, empotradas)

	return err
}

// comprobar hace la comprobación de ComprobarConflictos y, sin ningún
// conflicto, devuelve lo que examinó, de lo que sale el plan.
func comprobar(disco Disco, enlazador Enlazador, pedido Pedido, empotradas []SkillEmpotrada) (*comprobacion, error) {
	pedidas, err := skillsDelPedido(pedido.Skills, empotradas)
	if err != nil {
		return nil, err
	}

	c := &comprobacion{
		disco:       disco,
		enlazador:   enlazador,
		ambito:      pedido.Ambito,
		enConflicto: map[string]ClaseDeConflicto{},
		sondas:      map[string]bool{},
		encima:      map[string]Entrada{},
	}

	neutroReal, err := c.comprobarElNeutro()
	if err != nil {
		return nil, err
	}

	ambitoIlegible := len(c.enConflicto) > 0

	hostReal, err := c.comprobarElHost(pedido.HostClaude)
	if err != nil {
		return nil, err
	}

	if !ambitoIlegible {
		err = c.comprobarLasSkills(pedidas, neutroReal, hostReal)
		if err != nil {
			return nil, err
		}
	}

	if len(c.enConflicto) > 0 {
		return nil, nuevoErrorDeConflictos(c.enConflicto)
	}

	return c, nil
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
// conflicto que lleva encontrado, lo que ya respondió el Enlazador y lo que
// examinó, que es de lo que sale el plan si no hay ningún conflicto.
type comprobacion struct {
	disco     Disco
	enlazador Enlazador
	ambito    Ambito
	// enConflicto es la clase de cada entrada en conflicto, por su ruta.
	enConflicto map[string]ClaseDeConflicto
	// sondas es la respuesta de Disponible por directorio, que se pregunta
	// una sola vez en la invocación.
	sondas map[string]bool

	// manifiesto es el del ámbito, vacío si no hay ninguno.
	manifiesto Manifiesto
	// neutroQueFalta son las guardas del ámbito que no existen, desde la
	// primera que falta y de arriba abajo.
	neutroQueFalta []string
	// host es lo que se sabe del host claude.
	host hostExaminado
	// skills es lo examinado de cada skill pedida, en orden de nombre.
	skills []skillExaminada
	// encima es lo examinado por encima de la raíz del ámbito o de la ruta de
	// --dir, por ruta, cada una una sola vez.
	encima map[string]Entrada
}

// hostExaminado es lo que se sabe del host claude: si se enlaza en esta
// ejecución y, si se enlaza, si .claude y .claude/skills son, cada uno, un
// directorio real o faltan.
type hostExaminado struct {
	seEnlaza bool
	claude   estadoDeDirectorio
	skills   estadoDeDirectorio
}

// skillExaminada es lo que se sabe de una skill pedida sin ningún conflicto:
// lo que lleva el binario, lo que declara el manifiesto y lo que hay en el
// disco.
type skillExaminada struct {
	empotrada   SkillEmpotrada
	declarada   SkillDeclarada
	esDeclarada bool
	// neutro es lo que hay en su directorio del directorio neutro; nada si
	// falta.
	neutro examinado
	// host es su entrada de host, si se enlaza y .claude/skills existe.
	host hostDeSkill
}

// hostDeSkill es lo que se sabe de la entrada de host de una skill: su tipo
// y, si es una copia declarada, si pasa a enlace, sus ficheros declarados,
// relativos a ella, y lo que hay en ella.
type hostDeSkill struct {
	tipo       TipoDeEntrada
	aEnlace    bool
	declarados map[string]string
	copia      examinado
}

// examinado es lo que existe en el directorio de una skill o de su copia de
// host, visto sin seguir enlaces: el tipo de cada entrada examinada, por su
// ruta relativa a él, "." incluido. Sin ningún conflicto, cada directorio es
// real y cada fichero declarado está intacto.
type examinado map[string]TipoDeEntrada

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
// entradas de host—, guarda el manifiesto, si lo hay y se puede leer, y las
// guardas que faltan, y dice si el directorio neutro es un directorio real.
func (c *comprobacion) comprobarElNeutro() (bool, error) {
	guardas := c.ambito.Guardas()

	for i, guarda := range guardas {
		estado, err := c.examinarDirectorio(guarda)
		if err != nil {
			return false, err
		}

		if estado == directorioAusente {
			c.neutroQueFalta = guardas[i:]
		}

		if estado != directorioReal {
			return false, nil
		}
	}

	ruta := c.ambito.RutaDelManifiesto()

	manifiesto, err := leerManifiestoDe(c.disco, ruta)

	var ilegible *ManifiestoIlegible
	if errors.As(err, &ilegible) {
		c.anotar(ruta, ConflictoManifiestoIlegible)

		return true, nil
	}

	if err != nil {
		return true, err
	}

	if !c.ambito.ConHosts() && conEntradasDeHost(manifiesto) {
		c.anotar(ruta, ConflictoManifiestoConEntradasDeHost)
	}

	c.manifiesto = manifiesto

	return true, nil
}

// leerManifiestoDe lee el manifiesto de ruta sin seguir un enlace (FR-028):
// sin él, el Manifiesto vacío; si existe y no es un fichero regular o no
// respeta la forma de contracts/manifiesto.md, un *ManifiestoIlegible (FR-035),
// que es un conflicto: lo reconoce y lo nombra la propia orden. Si no se puede
// examinar ni leer, el error del Disco tal cual: un error del sistema que la
// orden no interpreta es un defecto del entorno, no un conflicto (ADR 0023).
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
		return Manifiesto{}, err
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

// comprobarElHost aplica las dos últimas filas de data-model §4.1, guarda si
// se enlaza en el host en esta ejecución (§4.3) y dice si hay entradas de
// host que examinar: solo si se enlaza y .claude/skills es un directorio
// real. Si se enlaza y falta .claude o .claude/skills, no existe ninguna
// entrada de host; si alguno de los dos existe y no es un directorio real, es
// el conflicto y nada por debajo se examina.
func (c *comprobacion) comprobarElHost(conHostClaude bool) (bool, error) {
	if !c.ambito.ConHosts() {
		return false, nil
	}

	claude := c.ambito.DirectorioDelHost()

	entrada, err := c.disco.Examinar(claude)
	if err != nil {
		return false, err
	}

	switch entrada.Tipo {
	case EntradaDirectorio:
	case EntradaAusente:
		c.host = hostExaminado{seEnlaza: conHostClaude, claude: directorioAusente, skills: directorioAusente}

		return false, nil
	default:
		if conHostClaude {
			c.anotar(claude, ConflictoRutaQueNoEsDirectorio)
		}

		return false, nil
	}

	estado, err := c.examinarDirectorio(c.ambito.SkillsDelHost())
	c.host = hostExaminado{seEnlaza: true, claude: directorioReal, skills: estado}

	return estado == directorioReal, err
}

// comprobarLasSkills examina cada skill pedida en el directorio neutro, si es
// un directorio real, y en el host, si hay entradas de host que examinar, y
// guarda lo examinado.
func (c *comprobacion) comprobarLasSkills(pedidas []SkillEmpotrada, neutroReal, hostReal bool) error {
	for _, skill := range pedidas {
		declarada, esDeclarada := c.manifiesto.Skills[skill.Nombre]
		examinada := skillExaminada{empotrada: skill, declarada: declarada, esDeclarada: esDeclarada}

		var err error

		if neutroReal {
			examinada.neutro, err = c.comprobarEnElNeutro(skill, declarada, esDeclarada)
			if err != nil {
				return err
			}
		}

		if hostReal {
			examinada.host, err = c.comprobarEnElHost(skill, declarada.Claude)
			if err != nil {
				return err
			}
		}

		c.skills = append(c.skills, examinada)
	}

	return nil
}

// comprobarEnElNeutro aplica a la skill la primera tabla de data-model §4.2 y,
// si su directorio es real y está declarada, la segunda a sus ficheros, y
// devuelve lo que hay en él; nada si falta.
func (c *comprobacion) comprobarEnElNeutro(
	skill SkillEmpotrada, declarada SkillDeclarada, esDeclarada bool,
) (examinado, error) {
	ruta := c.ambito.RutaDeSkill(skill.Nombre)

	entrada, err := c.disco.Examinar(ruta)
	if err != nil {
		return nil, err
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

	return nil, nil
}

// comprobarEnElHost aplica a la entrada de host de la skill la tabla de
// data-model §4.3, con host su entrada declarada, si la tiene, y devuelve lo
// que se sabe de ella.
func (c *comprobacion) comprobarEnElHost(skill SkillEmpotrada, host *EntradaDeHost) (hostDeSkill, error) {
	ruta := c.ambito.RutaDeHost(skill.Nombre)

	entrada, err := c.disco.Examinar(ruta)
	if err != nil {
		return hostDeSkill{}, err
	}

	examinada := hostDeSkill{tipo: entrada.Tipo}

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

	return examinada, nil
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
func (c *comprobacion) comprobarDirectorioDeHost(ruta string, skill SkillEmpotrada, host *EntradaDeHost) (hostDeSkill, error) {
	examinada := hostDeSkill{tipo: EntradaDirectorio}

	if host == nil || host.Modo != ModoCopia {
		c.anotar(ruta, ConflictoCarpetaAjena)

		return examinada, nil
	}

	disponible, err := c.disponible(c.ambito.SkillsDelHost())
	if err != nil {
		return examinada, err
	}

	examinada.aEnlace = disponible
	examinada.declarados = relativas(host.Ficheros, host.Ruta+"/")

	if disponible {
		examinada.copia, err = c.comprobarCopiaQueSeRetira(ruta, examinada.declarados)
	} else {
		examinada.copia, err = c.comprobarFicheros(ruta, examinada.declarados, skill.Ficheros)
	}

	return examinada, err
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
// declarado ni empotrado no se examina (FR-047). Devuelve lo que existe de
// todo ello.
func (c *comprobacion) comprobarFicheros(
	base string, declarados map[string]string, empotrados []FicheroEmpotrado,
) (examinado, error) {
	rutas := slices.Collect(maps.Keys(declarados))
	for _, fichero := range empotrados {
		rutas = append(rutas, fichero.Ruta)
	}

	slices.Sort(rutas)
	rutas = slices.Compact(rutas)

	estados := map[string]estadoDeDirectorio{".": directorioReal}

	for _, rel := range rutas {
		if _, err := c.estadoDelDirectorio(base, path.Dir(rel), estados); err != nil {
			return nil, err
		}
	}

	existe := examinado{}

	for dir, estado := range estados {
		if estado == directorioReal {
			existe[dir] = EntradaDirectorio
		}
	}

	for _, rel := range rutas {
		if estados[path.Dir(rel)] != directorioReal {
			continue
		}

		ruta := path.Join(base, rel)

		entrada, err := c.disco.Examinar(ruta)
		if err != nil {
			return nil, err
		}

		huella, declarado := declarados[rel]
		if err := c.clasificarFichero(ruta, entrada, huella, declarado); err != nil {
			return nil, err
		}

		if entrada.Tipo != EntradaAusente {
			existe[rel] = entrada.Tipo
		}
	}

	return existe, nil
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
// directorio, sus ficheros declarados relativos a él con su huella, cada
// directorio intermedio de esos ficheros y lo que existe en ella.
type copiaDeHost struct {
	base        string
	declarados  map[string]string
	intermedios map[string]bool
	existe      examinado
}

// comprobarCopiaQueSeRetira exige que todo lo que contiene la copia de base,
// que se va a retirar entera para crear el enlace, esté declarado e intacto
// (FR-046): una entrada no declarada es un fichero ajeno, un fichero editado
// lo es, y un intermedio que no es un directorio real, una ruta que no es
// directorio; lo declarado que falta no es conflicto. Devuelve todo lo que
// hay en ella.
func (c *comprobacion) comprobarCopiaQueSeRetira(base string, declarados map[string]string) (examinado, error) {
	copia := copiaDeHost{
		base:        base,
		declarados:  declarados,
		intermedios: map[string]bool{},
		existe:      examinado{".": EntradaDirectorio},
	}

	for rel := range declarados {
		for dir := path.Dir(rel); dir != "."; dir = path.Dir(dir) {
			copia.intermedios[dir] = true
		}
	}

	if err := c.recorrerCopia(copia, "."); err != nil {
		return nil, err
	}

	return copia.existe, nil
}

// recorrerCopia examina cada entrada del directorio dir de la copia, relativo
// a su base, la anota si existe y entra en cada una que es un intermedio y un
// directorio real.
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

		if entrada.Tipo != EntradaAusente {
			copia.existe[rel] = entrada.Tipo
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

// planificacion es el plan que se construye, skill a skill, con lo que
// examinó una comprobación sin ningún conflicto.
type planificacion struct {
	c       *comprobacion
	version string
	plan    Plan
	// declaradas son las entradas finales de las skills pedidas.
	declaradas map[string]SkillDeclarada
	// copias son las entradas en copia de las de un enlace que crear.
	copias map[string]*EntradaDeHost
	// hostQueCrear dice si se crea alguna entrada de host, enlace o copia.
	hostQueCrear bool
	// creados es cada directorio de encima del ámbito que el plan ya crea.
	creados map[string]bool
}

// planificar construye el plan con lo examinado, con version como la de cada
// skill pedida.
func (c *comprobacion) planificar(version string) (Plan, error) {
	p := &planificacion{
		c:          c,
		version:    version,
		plan:       Plan{Skills: make([]SkillInstalada, 0, len(c.skills))},
		declaradas: map[string]SkillDeclarada{},
		copias:     map[string]*EntradaDeHost{},
		creados:    map[string]bool{},
	}

	for _, skill := range c.skills {
		if err := p.planificarSkill(skill); err != nil {
			return Plan{}, err
		}
	}

	if err := p.planificarLoQueFalta(); err != nil {
		return Plan{}, err
	}

	return p.plan, nil
}

// planificarSkill añade al plan lo de la skill: su directorio en el neutro,
// su entrada de host, su entrada final del manifiesto y su salida.
func (p *planificacion) planificarSkill(skill skillExaminada) error {
	nombre := skill.empotrada.Nombre
	neutro := p.c.ambito.RutaDeSkill(nombre)

	cambiaElNeutro := p.actualizar(neutro, skill.neutro, relativas(skill.declarada.Ficheros, nombre+"/"),
		skill.empotrada.Ficheros)

	host, cambiaElHost, err := p.planificarHost(skill)
	if err != nil {
		return err
	}

	nueva := SkillDeclarada{
		Version:  p.version,
		Ficheros: ficherosDeclarados(skill.empotrada.Ficheros, nombre+"/"),
		Claude:   host,
	}
	p.declaradas[nombre] = nueva

	estado := EstadoInstalada

	if skill.esDeclarada {
		estado = EstadoSinCambios
		if cambiaElNeutro || cambiaElHost || !mismaDeclaracion(skill.declarada, nueva) {
			estado = EstadoActualizada
		}
	}

	enlaces := []Enlace{}
	if host != nil {
		enlaces = append(enlaces, Enlace{Host: hostClaude, Ruta: p.c.ambito.RutaDeHost(nombre), Modo: host.Modo})
	}

	p.plan.Skills = append(p.plan.Skills, SkillInstalada{Nombre: nombre, Ruta: neutro, Estado: estado, Enlaces: enlaces})

	return nil
}

// planificarHost añade al plan lo de la entrada de host de la skill
// (data-model §4.3) y devuelve la que queda declarada, si queda alguna, y si
// cambia algo en el disco: sin enlazar en esta ejecución, ninguna, y la
// declarada se quita del manifiesto sin tocar el disco (FR-046); el enlace de
// FR-021 que ya está se adopta o se queda (FR-041); una copia declarada pasa
// a enlace, retirándola entera, o se mantiene y se actualiza como el
// directorio neutro; y la que falta se crea en el modo previsto.
func (p *planificacion) planificarHost(skill skillExaminada) (*EntradaDeHost, bool, error) {
	if !p.c.host.seEnlaza {
		return nil, false, nil
	}

	ruta := p.c.ambito.RutaDeHost(skill.empotrada.Nombre)

	switch {
	case skill.host.tipo == EntradaEnlace:
		return entradaEnlazada(skill.empotrada), false, nil
	case skill.host.tipo == EntradaDirectorio && !skill.host.aEnlace:
		cambia := p.actualizar(ruta, skill.host.copia, skill.host.declarados, skill.empotrada.Ficheros)

		return entradaEnCopia(skill.empotrada), cambia, nil
	case skill.host.tipo == EntradaDirectorio:
		p.plan.Retirar = append(p.plan.Retirar, retiradaDeCopia(ruta, skill.host.copia)...)
		p.enlazar(skill.empotrada, ruta)

		return entradaEnlazada(skill.empotrada), true, nil
	}

	p.hostQueCrear = true

	modo, err := p.c.modoPrevisto()
	if err != nil {
		return nil, false, err
	}

	if modo == ModoCopia {
		p.actualizar(ruta, nil, nil, skill.empotrada.Ficheros)

		return entradaEnCopia(skill.empotrada), true, nil
	}

	p.enlazar(skill.empotrada, ruta)

	return entradaEnlazada(skill.empotrada), true, nil
}

// actualizar añade al plan lo que se retira y se escribe en base, el
// directorio de una skill o de su copia de host, y dice si hay algo.
func (p *planificacion) actualizar(
	base string, existe examinado, declarados map[string]string, empotrados []FicheroEmpotrado,
) bool {
	retirar := aRetirar(base, existe, declarados, empotrados)
	escribir := aEscribir(base, existe, declarados, empotrados)

	p.plan.Retirar = append(p.plan.Retirar, retirar...)
	p.plan.Escribir.DirectoriosQueFaltan = append(p.plan.Escribir.DirectoriosQueFaltan,
		escribir.DirectoriosQueFaltan...)
	p.plan.Escribir.Ficheros = append(p.plan.Escribir.Ficheros, escribir.Ficheros...)

	return len(retirar) > 0 || len(escribir.Ficheros) > 0
}

// enlazar añade al plan el enlace de FR-021 en la entrada de host ruta de la
// skill, con su recurso de copia (FR-024).
func (p *planificacion) enlazar(skill SkillEmpotrada, ruta string) {
	p.plan.Enlazar.Enlaces = append(p.plan.Enlazar.Enlaces, EnlaceNuevo{
		Skill:   skill.Nombre,
		Ruta:    ruta,
		Destino: destinoDeHost(skill.Nombre),
		Copia:   aEscribir(ruta, nil, nil, skill.Ficheros),
	})
	p.copias[skill.Nombre] = entradaEnCopia(skill)
}

// aRetirar son los ficheros de base que se retiran en la fase 1, en orden de
// ruta: cada declarado que existe, intacto, y que ya no se empotra o cuya
// huella declarada no es la empotrada, porque se va a reescribir (FR-046).
func aRetirar(base string, existe examinado, declarados map[string]string, empotrados []FicheroEmpotrado) []string {
	huellas := make(map[string]string, len(empotrados))
	for _, fichero := range empotrados {
		huellas[fichero.Ruta] = fichero.Huella
	}

	var retirar []string

	for _, rel := range slices.Sorted(maps.Keys(declarados)) {
		if _, hay := existe[rel]; hay && huellas[rel] != declarados[rel] {
			retirar = append(retirar, path.Join(base, rel))
		}
	}

	return retirar
}

// aEscribir es lo que se escribe en base en la fase 4: cada fichero empotrado
// que falta o cuya huella declarada no es la suya, con cada directorio que le
// falte, de arriba abajo. Lo que no está declarado ni se empotra no se toca
// (FR-047).
func aEscribir(base string, existe examinado, declarados map[string]string, empotrados []FicheroEmpotrado) Escrituras {
	var escribir Escrituras

	creados := map[string]bool{}

	for _, fichero := range empotrados {
		if _, hay := existe[fichero.Ruta]; hay && declarados[fichero.Ruta] == fichero.Huella {
			continue
		}

		for _, dir := range cadenaDe(path.Dir(fichero.Ruta)) {
			if _, hay := existe[dir]; !hay && !creados[dir] {
				creados[dir] = true
				escribir.DirectoriosQueFaltan = append(escribir.DirectoriosQueFaltan, path.Join(base, dir))
			}
		}

		escribir.Ficheros = append(escribir.Ficheros, Escritura{Ruta: path.Join(base, fichero.Ruta), Contenido: fichero.Contenido})
	}

	return escribir
}

// cadenaDe es dir, relativo, y cada directorio de encima hasta ".", de arriba
// abajo.
func cadenaDe(dir string) []string {
	cadena := []string{dir}
	for dir != "." {
		dir = path.Dir(dir)
		cadena = append(cadena, dir)
	}

	slices.Reverse(cadena)

	return cadena
}

// retiradaDeCopia es el orden en que se retira entera la copia de base, con
// todo lo que existe en ella: cada fichero, después cada directorio, de abajo
// arriba —uno va detrás de todo lo que cuelga de él, que es mayor en orden de
// bytes—, y el de la copia el último (FR-046).
func retiradaDeCopia(base string, existe examinado) []string {
	var ficheros, dirs []string

	for rel, tipo := range existe {
		switch {
		case rel == ".":
		case tipo == EntradaDirectorio:
			dirs = append(dirs, rel)
		default:
			ficheros = append(ficheros, rel)
		}
	}

	slices.Sort(ficheros)
	slices.Sort(dirs)
	slices.Reverse(dirs)

	rutas := make([]string, 0, len(ficheros)+len(dirs)+1)
	for _, rel := range slices.Concat(ficheros, dirs) {
		rutas = append(rutas, path.Join(base, rel))
	}

	return append(rutas, base)
}

// ficherosDeclarados son los empotrados como los declara el manifiesto: cada
// ruta con prefijo delante, con su huella.
func ficherosDeclarados(empotrados []FicheroEmpotrado, prefijo string) map[string]string {
	ficheros := make(map[string]string, len(empotrados))
	for _, fichero := range empotrados {
		ficheros[prefijo+fichero.Ruta] = fichero.Huella
	}

	return ficheros
}

// rutaDeclaradaDeHost es la de la entrada de host de la skill nombre como la
// declara el manifiesto, relativa a la raíz del ámbito (FR-031).
func rutaDeclaradaDeHost(nombre string) string {
	return path.Join(directorioDeSkillsDelHostClaude, nombre)
}

// entradaEnlazada es la entrada de host de la skill declarada en enlace.
func entradaEnlazada(skill SkillEmpotrada) *EntradaDeHost {
	return &EntradaDeHost{Ruta: rutaDeclaradaDeHost(skill.Nombre), Modo: ModoEnlace}
}

// entradaEnCopia es la entrada de host de la skill declarada en copia, con la
// huella de cada fichero copiado (FR-024).
func entradaEnCopia(skill SkillEmpotrada) *EntradaDeHost {
	ruta := rutaDeclaradaDeHost(skill.Nombre)

	return &EntradaDeHost{Ruta: ruta, Modo: ModoCopia, Ficheros: ficherosDeclarados(skill.Ficheros, ruta+"/")}
}

// mismaDeclaracion dice si dos entradas de skill del manifiesto declaran lo
// mismo: su versión, sus ficheros con su huella y su entrada de host.
func mismaDeclaracion(a, b SkillDeclarada) bool {
	return a.Version == b.Version && maps.Equal(a.Ficheros, b.Ficheros) && mismaEntradaDeHost(a.Claude, b.Claude)
}

// mismaEntradaDeHost dice si dos entradas de host declaran lo mismo, o si
// faltan las dos.
func mismaEntradaDeHost(a, b *EntradaDeHost) bool {
	if a == nil || b == nil {
		return a == b
	}

	return a.Ruta == b.Ruta && a.Modo == b.Modo && maps.Equal(a.Ficheros, b.Ficheros)
}

// planificarLoQueFalta añade al plan lo que falta hasta .claude/skills, si se
// crea alguna entrada de host, y el manifiesto final con lo que falta hasta el
// directorio neutro, si cambia.
func (p *planificacion) planificarLoQueFalta() error {
	if p.hostQueCrear {
		faltan, err := p.queFaltaHastaElHost()
		if err != nil {
			return err
		}

		p.plan.Enlazar.DirectoriosQueFaltan = faltan
	}

	p.plan.Manifiesto = ManifiestoFinal{
		Ruta:       p.c.ambito.RutaDelManifiesto(),
		actual:     p.c.manifiesto,
		version:    p.version,
		declaradas: p.declaradas,
		copias:     p.copias,
	}

	contenido, err := p.plan.Manifiesto.Contenido(nil)
	if err != nil || contenido == nil {
		return err
	}

	p.plan.Manifiesto.DirectoriosQueFaltan, err = p.queFaltaHastaElNeutro()

	return err
}

// queFaltaHastaElHost es lo que falta hasta .claude/skills incluido.
func (p *planificacion) queFaltaHastaElHost() ([]string, error) {
	switch {
	case p.c.host.claude == directorioAusente:
		faltan, err := p.queFaltaHasta(p.c.ambito.DirectorioDelHost())

		return append(faltan, p.nuevos(p.c.ambito.SkillsDelHost())...), err
	case p.c.host.skills == directorioAusente:
		return p.nuevos(p.c.ambito.SkillsDelHost()), nil
	}

	return nil, nil
}

// queFaltaHastaElNeutro es lo que falta hasta el directorio neutro incluido:
// lo que falta por encima de la primera guarda que falta, ella y las demás.
func (p *planificacion) queFaltaHastaElNeutro() ([]string, error) {
	if len(p.c.neutroQueFalta) == 0 {
		return nil, nil
	}

	faltan, err := p.queFaltaHasta(p.c.neutroQueFalta[0])
	if err != nil {
		return nil, err
	}

	return append(faltan, p.nuevos(p.c.neutroQueFalta[1:]...)...), nil
}

// queFaltaHasta son dir, que no existe, y cada directorio de encima que
// tampoco existe, de arriba abajo, hasta el primero en el que hay algo, que se
// usa tal cual (FR-014, FR-027), sin los que el plan ya crea. El directorio de
// trabajo y la raíz del sistema existen siempre y no se examinan.
func (p *planificacion) queFaltaHasta(dir string) ([]string, error) {
	faltan := []string{dir}

	for encima := path.Dir(dir); encima != "." && encima != "/" && !p.creados[encima]; encima = path.Dir(encima) {
		entrada, err := p.c.examinarEncima(encima)
		if err != nil {
			return nil, err
		}

		if entrada.Tipo != EntradaAusente {
			break
		}

		faltan = append(faltan, encima)
	}

	slices.Reverse(faltan)

	return p.nuevos(faltan...), nil
}

// nuevos son las rutas que el plan todavía no crea, que desde ahora crea.
func (p *planificacion) nuevos(rutas ...string) []string {
	var nuevos []string

	for _, ruta := range rutas {
		if !p.creados[ruta] {
			p.creados[ruta] = true
			nuevos = append(nuevos, ruta)
		}
	}

	return nuevos
}

// examinarEncima es la entrada de ruta, por encima de la raíz del ámbito, que
// se examina una sola vez.
func (c *comprobacion) examinarEncima(ruta string) (Entrada, error) {
	if entrada, examinada := c.encima[ruta]; examinada {
		return entrada, nil
	}

	entrada, err := c.disco.Examinar(ruta)
	if err != nil {
		return Entrada{}, err
	}

	c.encima[ruta] = entrada

	return entrada, nil
}

// modoPrevisto es el modo en que se crea una entrada de host que falta: el que
// diga el Enlazador por el directorio de la sonda, enlace si se puede y copia
// si no (FR-024), o enlace si no hay ningún directorio donde sondear (research.md
// D9).
func (c *comprobacion) modoPrevisto() (Modo, error) {
	sonda, hay, err := c.directorioDeLaSonda()
	if err != nil {
		return "", err
	}

	if !hay {
		return ModoEnlace, nil
	}

	disponible, err := c.disponible(sonda)
	if err != nil {
		return "", err
	}

	if disponible {
		return ModoEnlace, nil
	}

	return ModoCopia, nil
}

// directorioDeLaSonda es el directorio real existente más próximo a
// .claude/skills sin salir del ámbito, donde vivirá la entrada de host, y si
// hay alguno: .claude/skills, .claude o la raíz. La raíz se usa tal cual y
// existe si es un directorio o un enlace que resuelve (FR-027); en local es el
// directorio de trabajo, que existe siempre y no se examina (data-model §3).
func (c *comprobacion) directorioDeLaSonda() (string, bool, error) {
	switch {
	case c.host.skills == directorioReal:
		return c.ambito.SkillsDelHost(), true, nil
	case c.host.claude == directorioReal:
		return c.ambito.DirectorioDelHost(), true, nil
	}

	raiz := path.Clean(c.ambito.Raiz())
	if raiz == "." {
		return raiz, true, nil
	}

	entrada, err := c.examinarEncima(raiz)
	if err != nil {
		return "", false, err
	}

	return raiz, entrada.Tipo == EntradaDirectorio || (entrada.Tipo == EntradaEnlace && entrada.Resuelve), nil
}
