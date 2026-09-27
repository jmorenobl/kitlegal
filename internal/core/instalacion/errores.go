package instalacion

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// errorDeInvocacion es el rechazo de una invocación antes de tocar el disco
// (contracts/applet-skills.md §2): uno de argumentos, que el kernel traduce a
// código 2, o -g sin HOME, un conflicto con el entorno que la persona resuelve
// y que traduce a 7 (data-model §9; ADR 0023). No se exporta:
// quien lo recibe lo reconoce por su clase, con errors.As a schema.ConClase,
// que es como lo reconoce el kernel.
type errorDeInvocacion struct {
	// clase es schema.ClaseArgumentos o, solo para -g sin HOME,
	// schema.ClaseConflicto.
	clase schema.Clase
	// mensaje contiene la frase de su fila de la tabla y dice por qué.
	mensaje string
}

// Un rechazo de la invocación declara su clase él mismo.
var _ schema.ConClase = (*errorDeInvocacion)(nil)

// Error devuelve el mensaje, que llega literal al sobre de fallo y a la salida
// de error.
func (e *errorDeInvocacion) Error() string {
	return e.mensaje
}

// Clase es la del rechazo.
func (e *errorDeInvocacion) Clase() schema.Clase {
	return e.clase
}

// argumentosInvalidos es un rechazo de clase «argumentos», el de las filas 1
// a 4.
func argumentosInvalidos(mensaje string) error {
	return &errorDeInvocacion{clase: schema.ClaseArgumentos, mensaje: mensaje}
}

// globalConDir es la fila 1: -g y --dir a la vez, que serían dos ámbitos
// (FR-013).
func globalConDir() error {
	return argumentosInvalidos("-g y --dir se excluyen: cada invocación actúa en un único ámbito, " +
		"el global o el de --dir")
}

// hostConDir es la fila 2: --host junto a --dir, cuyo ámbito no tiene hosts
// (FR-013).
func hostConDir() error {
	return argumentosInvalidos("--host no se combina con --dir: el ámbito de --dir es solo ese directorio, " +
		"sin ningún host")
}

// hostNoAdmitido es la fila 3: --host con un valor distinto de claude
// (FR-020), que nombra con %q.
func hostNoAdmitido(host string) error {
	return argumentosInvalidos(fmt.Sprintf("el host %q no se admite: el único host admitido es %s", host, hostClaude))
}

// skillDesconocida es la fila 4: nombre, con %q, no es el de ninguna skill
// empotrada; disponibles son las que sí, en orden de nombre (FR-010).
func skillDesconocida(nombre string, disponibles []string) error {
	return argumentosInvalidos(fmt.Sprintf("%q no es ninguna skill de este binario; skills disponibles: %s",
		nombre, strings.Join(disponibles, ", ")))
}

// homeSinDefinir es la fila 5: -g con HOME sin definir o vacío (FR-012). No
// es un error de argumentos —la orden está bien escrita y es el entorno el que
// no la permite— ni un defecto del programa: es un conflicto con el entorno que
// la persona resuelve, así que declara la clase «conflicto» (data-model §9; ADR
// 0023).
func homeSinDefinir() error {
	return &errorDeInvocacion{
		clase: schema.ClaseConflicto,
		mensaje: "HOME no está definido o está vacío: el ámbito global de -g está en " +
			"$HOME/" + directorioNeutro,
	}
}

// ManifiestoIlegible es un kitlegal.json que existe y no se puede usar: no es
// un fichero regular, no se puede leer o no respeta la forma de
// contracts/manifiesto.md (FR-035). Quien no puede leer el manifiesto no sabe
// qué es suyo, así que install lo nombra como conflicto y no toca nada, list y
// doctor salen nombrándolo y el aviso no tiene ningún efecto.
//
// No declara una clase a propósito: nunca llega así al kernel. El conflicto
// de install y el ámbito ilegible de list y doctor son los errores que la
// declaran (data-model §9). Se exporta para reconocerlo con errors.As.
type ManifiestoIlegible struct {
	// causa dice qué lo hace ilegible; puede envolver el error de
	// encoding/json/v2 que rechazó el documento.
	causa error
}

// Error dice que el manifiesto es ilegible y por qué, si se sabe.
func (e *ManifiestoIlegible) Error() string {
	if e.causa == nil {
		return "manifiesto ilegible"
	}

	return "manifiesto ilegible: " + e.causa.Error()
}

// Unwrap devuelve lo que lo hace ilegible.
func (e *ManifiestoIlegible) Unwrap() error {
	return e.causa
}

// ClaseDeConflicto es la de una entrada que install no puede crear, cambiar
// ni retirar, con el nombre literal con el que se nombra (FR-041;
// contracts/applet-skills.md §5).
type ClaseDeConflicto string

// Las nueve clases de FR-041, en el orden de su lista.
const (
	// ConflictoCarpetaAjena (a) es un directorio real con el nombre de una
	// skill donde el manifiesto no declara un directorio.
	ConflictoCarpetaAjena ClaseDeConflicto = "carpeta ajena"
	// ConflictoFichero (b) es un fichero regular, una tubería, un socket o un
	// dispositivo donde va el directorio de una skill o su entrada de host.
	ConflictoFichero ClaseDeConflicto = "fichero"
	// ConflictoEnlaceAOtroSitio (c) es un enlace simbólico que resuelve donde
	// no va el enlace de FR-021.
	ConflictoEnlaceAOtroSitio ClaseDeConflicto = "enlace a otro sitio"
	// ConflictoEnlaceRoto (d) es un enlace simbólico que cuelga o está en un
	// ciclo donde no va el enlace de FR-021.
	ConflictoEnlaceRoto ClaseDeConflicto = "enlace roto"
	// ConflictoFicheroEditado (e) es un fichero declarado cuya huella ya no
	// coincide o que ya no es un fichero regular.
	ConflictoFicheroEditado ClaseDeConflicto = "fichero editado"
	// ConflictoFicheroAjeno (f) es una entrada no declarada donde install
	// escribiría, o dentro de una copia de host que retiraría.
	ConflictoFicheroAjeno ClaseDeConflicto = "fichero ajeno"
	// ConflictoRutaQueNoEsDirectorio (g) es una ruta que tiene que ser un
	// directorio real o no existir, y es otra cosa.
	ConflictoRutaQueNoEsDirectorio ClaseDeConflicto = "ruta que no es directorio"
	// ConflictoManifiestoIlegible (h) es el manifiesto ilegible de FR-035.
	ConflictoManifiestoIlegible ClaseDeConflicto = "manifiesto ilegible"
	// ConflictoManifiestoConEntradasDeHost (i) es, con --dir, un manifiesto
	// que declara alguna entrada de host (FR-013).
	ConflictoManifiestoConEntradasDeHost ClaseDeConflicto = "manifiesto con entradas de host"
)

// Conflicto es una entrada que install no puede crear, cambiar ni retirar:
// su clase y su ruta, como se alcanza desde el directorio de trabajo.
type Conflicto struct {
	// Clase es la única de la entrada.
	Clase ClaseDeConflicto

	// Ruta es la de la entrada, la que se presenta.
	Ruta string
}

// prefijoDeInstall encabeza el mensaje de todo fallo de install que nombra
// entradas: el de un conflicto y el de escritura (contracts/applet-skills.md §5).
const prefijoDeInstall = "skills install: "

// cabeceraDeInstall es la primera línea del mensaje con que install nombra
// cada conflicto (contracts/applet-skills.md §5). Su última palabra va en dos
// literales porque misspell, con su diccionario inglés, marca la palabra
// española entera como una errata de «conflicts».
const cabeceraDeInstall = prefijoDeInstall + "nada se ha creado ni cambiado; conflict" + "os:"

// ErrorDeConflictos es el rechazo de install cuando alguna entrada que iba a
// crear, cambiar o retirar no es suya: los nombra todos, cada entrada con una
// sola clase y en orden de ruta byte a byte, antes de escribir nada (FR-040 a
// FR-042; data-model §9). Como cada entrada tiene una sola clase, nunca hay
// dos con la misma ruta y el orden de las clases no llega a desempatar.
//
// Declara la clase «conflicto», que el kernel traduce a código 7 (ADR 0023), y
// su mensaje es el del sobre de fallo y la salida de error: una cabecera y una
// línea «<clase>: <ruta>» por conflicto (research.md D12). Se exporta para
// reconocerlo con errors.As; su valor cero no nombra ninguno.
type ErrorDeConflictos struct {
	// lista tiene cada conflicto, en orden de ruta.
	lista []Conflicto
}

// El rechazo por conflicto declara su clase él mismo.
var _ schema.ConClase = (*ErrorDeConflictos)(nil)

// nuevoErrorDeConflictos es el error que nombra cada conflicto de porRuta, la
// clase de cada entrada por su ruta, en orden de ruta.
func nuevoErrorDeConflictos(porRuta map[string]ClaseDeConflicto) *ErrorDeConflictos {
	lista := make([]Conflicto, 0, len(porRuta))
	for _, ruta := range slices.Sorted(maps.Keys(porRuta)) {
		lista = append(lista, Conflicto{Clase: porRuta[ruta], Ruta: ruta})
	}

	return &ErrorDeConflictos{lista: lista}
}

// Error es la cabecera seguida de una línea por conflicto, sin salto de línea
// final.
func (e *ErrorDeConflictos) Error() string {
	var mensaje strings.Builder

	mensaje.WriteString(cabeceraDeInstall)

	for _, conflicto := range e.lista {
		mensaje.WriteString("\n" + string(conflicto.Clase) + ": " + conflicto.Ruta)
	}

	return mensaje.String()
}

// Clase es «conflicto»: un conflicto sale con código 7.
func (e *ErrorDeConflictos) Clase() schema.Clase {
	return schema.ClaseConflicto
}

// Lista es cada conflicto, en orden de ruta, en una copia.
func (e *ErrorDeConflictos) Lista() []Conflicto {
	return slices.Clone(e.lista)
}

// ClaseDeHallazgo es la de algo que doctor encuentra fuera de su sitio en una
// skill declarada y empotrada, con el nombre literal con el que se nombra
// (FR-065; contracts/applet-skills.md §5).
type ClaseDeHallazgo string

// Las cinco clases de FR-065, en el orden de su número, que desempata dos
// hallazgos con la misma ruta (FR-066).
const (
	// HallazgoFicheroEditado (1) es un fichero declarado cuya huella ya no
	// coincide, que ya no es un fichero regular o que falta.
	HallazgoFicheroEditado ClaseDeHallazgo = "fichero editado"
	// HallazgoEnlaceColgando (2) es el enlace de host de FR-021 declarado que
	// no resuelve, o un enlace que no resuelve donde va el directorio de una
	// skill o un directorio intermedio.
	HallazgoEnlaceColgando ClaseDeHallazgo = "enlace colgando"
	// HallazgoEnlaceAOtroSitio (3) es una entrada de host que ya no es lo que
	// se declaró, o que falta, o cualquier otra cosa que no es un directorio
	// real donde va el directorio de una skill o un directorio intermedio.
	HallazgoEnlaceAOtroSitio ClaseDeHallazgo = "enlace a otro sitio"
	// HallazgoCopia (4) es una copia de host declarada donde ya se puede
	// crear el enlace.
	HallazgoCopia ClaseDeHallazgo = "copia"
	// HallazgoVersionDistinta (5) es la versión del manifiesto, o la de una
	// skill, distinta de la del binario (FR-077).
	HallazgoVersionDistinta ClaseDeHallazgo = "versión distinta"
)

// numero es el de la clase en FR-065, de 1 a 5, y 0 si no es ninguna de ellas.
func (c ClaseDeHallazgo) numero() int {
	return slices.Index([]ClaseDeHallazgo{
		HallazgoFicheroEditado, HallazgoEnlaceColgando, HallazgoEnlaceAOtroSitio, HallazgoCopia, HallazgoVersionDistinta,
	}, c) + 1
}

// Los verbos que nombran su ámbito ilegible (contracts/applet-skills.md §5).
const (
	verboList   = "list"
	verboDoctor = "doctor"
)

// AmbitoIlegible es el de list y doctor cuando el ámbito no se puede leer: una
// de sus guardas no es un directorio real (FR-027), el manifiesto es ilegible
// (FR-035) o, con --dir, declara alguna entrada de host (FR-013). Nombra solo
// esa entrada, con su clase, y ninguna skill ni ningún hallazgo (FR-061,
// FR-067; data-model §9).
//
// Declara la clase «conflicto», que el kernel traduce a código 7 (ADR 0023), y
// su mensaje es una sola línea, «skills <verbo>: <clase>: <ruta>»
// (contracts/applet-skills.md §5). Se exporta para reconocerlo con errors.As.
type AmbitoIlegible struct {
	// verbo es list o doctor.
	verbo string
	// motivo es la entrada que no se puede leer, con su clase.
	motivo Conflicto
}

// El ámbito ilegible declara su clase él mismo.
var _ schema.ConClase = (*AmbitoIlegible)(nil)

// ambitoIlegible es el error del verbo que nombra la entrada de ruta con esa
// clase.
func ambitoIlegible(verbo string, clase ClaseDeConflicto, ruta string) *AmbitoIlegible {
	return &AmbitoIlegible{verbo: verbo, motivo: Conflicto{Clase: clase, Ruta: ruta}}
}

// Error es la línea que nombra la entrada.
func (e *AmbitoIlegible) Error() string {
	return "skills " + e.verbo + ": " + string(e.motivo.Clase) + ": " + e.motivo.Ruta
}

// Clase es «conflicto»: un ámbito ilegible sale con código 7.
func (e *AmbitoIlegible) Clase() schema.Clase {
	return schema.ClaseConflicto
}

// Motivo es la entrada que no se puede leer, con su clase: ruta que no es
// directorio, manifiesto ilegible o manifiesto con entradas de host.
func (e *AmbitoIlegible) Motivo() Conflicto {
	return e.motivo
}

// falloAlAplicar es el de una operación del Escritor que falla tras la
// comprobación, con el que Aplicar se para: «skills install: <operación>
// <ruta>: <error del sistema>», con la ruta como se alcanza desde el
// directorio de trabajo, envolviendo el error del Escritor (FR-044;
// contracts/applet-skills.md §5). No declara clase: es un fallo de entrada y
// salida, lo no previsto, que el kernel traduce a código 1 (data-model §9).
func falloAlAplicar(operacion, ruta string, err error) error {
	return fmt.Errorf("%s%s %s: %w", prefijoDeInstall, operacion, ruta, err)
}
