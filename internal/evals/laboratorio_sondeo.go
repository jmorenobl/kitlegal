package evals

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// Los argumentos de make evals-laboratorio que no tiene el sondeo de una skill,
// con el nombre con que los nombran sus errores (contracts/arnes.md §2 de H39 y
// de H41.2). MODELO y CONCURRENCIA son los del sondeo.
const (
	argumentoVariantes   = "VARIANTES"
	argumentoModos       = "MODOS"
	argumentoSesiones    = "SESIONES"
	argumentoClaudeCode  = "CLAUDE_CODE"
	argumentoDirectorio  = "DIRECTORIO"
	argumentoInstalacion = "INSTALACION"
)

// instalacionDelProducto es el único valor de INSTALACION, el que instala en
// cada sesión las cuatro skills del producto (contracts/arnes.md §2 y §3.1 de
// H41.2; research D7 de H41.2). Vacía, la instalación de hoy.
const instalacionDelProducto = "producto"

// separadorDeLaLista separa los elementos de VARIANTES y de MODOS.
const separadorDeLaLista = ","

// evalsDelLaboratorio es entre cuántas evals se reparten por igual las
// SESIONES de cada variante y modo, las cuatro del laboratorio (FR-012 y FR-030
// de H39): SESIONES tiene que ser múltiplo suyo.
const evalsDelLaboratorio = 4

// formaDelModeloCompleto es la de un identificador completo de modelo,
// claude-<familia>-<versión>, y no la de un alias como sonnet (FR-031 de H39;
// contracts/arnes.md §2 de H39).
var formaDelModeloCompleto = regexp.MustCompile(`^claude-[a-z]+-[0-9]+(-[0-9]+)*$`)

// Lo que el arnés prepara en su temporal, además del bin/ y el home/ del sondeo
// (contracts/arnes.md §3.1 de H39): el programa del laboratorio, en una carpeta
// fuera del PATH de las sesiones, y la copia de las skills de cada variante.
const (
	carpetaDelProgramaDelLaboratorio = "laboratorio-bin"
	paqueteDelLaboratorio            = "./cmd/laboratorio"
	programaDelLaboratorio           = "laboratorio"
	carpetaDeLasSkillsDelSondeo      = "skills"
)

// ordenDelServidorDelLaboratorio es la orden de laboratorio con la que
// servidor.json arranca el filtro en las variantes que cambian lo que anuncia el
// servidor (contracts/arnes.md §3.4 y §5 de H39).
const ordenDelServidorDelLaboratorio = "servidor"

// Lo que el arnés pregunta a Claude Code antes de abrir ninguna sesión: su
// versión, que escribe como <versión> (Claude Code) (research V22 y S1 de H39).
const (
	programaDeClaudeCode = "claude"
	banderaDeLaVersion   = "--version"
)

// ArgumentosDelLaboratorio son los ocho argumentos de make evals-laboratorio,
// tal como llegan, sin interpretar (contracts/arnes.md §1 y §2 de H39 y de
// H41.2; data-model §3 de H41.2): los siete de H39 e INSTALACION, detrás.
type ArgumentosDelLaboratorio struct {
	Variantes    string
	Modos        string
	Sesiones     string
	Modelo       string
	ClaudeCode   string
	Concurrencia string
	Directorio   string
	Instalacion  string
}

// SondeoDelLaboratorio es lo que sondearElLaboratorio necesita para abrir las
// sesiones de un sondeo del laboratorio (contracts/arnes.md §3 de H39; research
// D8 de H39).
type SondeoDelLaboratorio struct {
	// Argumentos son los de make, sin comprobar.
	Argumentos ArgumentosDelLaboratorio

	// Entorno es el de quien lanza el arnés: de él salen la credencial, el PATH
	// con el que se pregunta la versión de Claude Code, el entorno del go install
	// y lo poco que ven las sesiones, como en el sondeo de una skill.
	Entorno []string

	// Raiz es la del repositorio: la de la declaración del laboratorio, sus
	// variantes y sus evals, y la del go install de kitlegal y laboratorio.
	Raiz string

	// Temporal es el directorio del arnés, que crea y borra su guion: lo que
	// prepara va dentro; las sesiones, en DIRECTORIO.
	Temporal string

	// Guion es la ruta absoluta de scripts/evals-sesion.sh, que abre cada sesión.
	Guion string

	// ConstruirLosProgramas deja en el temporal, que recibe con su ruta
	// absoluta, kitlegal en bin/ y laboratorio en laboratorio-bin/, construidos
	// desde la raíz con el entorno de quien lanza el arnés como base:
	// construirLosProgramasDelLaboratorio en el punto de entrada, y en los
	// tests, un kitlegal y un laboratorio que no hacen nada, sin construir.
	ConstruirLosProgramas func(raiz, temporal string, base []string) error
}

// parDelLaboratorio es una variante pedida en uno de sus modos: lo que abre un
// reparto.
type parDelLaboratorio struct {
	variante varianteDelLaboratorio
	modo     Modo
}

// laboratorioComprobado es un sondeo del laboratorio cuyos argumentos,
// credencial y versión de Claude Code valen, con sus argumentos ya
// interpretados.
type laboratorioComprobado struct {
	// declarada es la declaración del laboratorio de H39; familia, la de las
	// variantes pedidas, la de H39 o una del producto (FR-010 de H40; FR-030);
	// y pares, los pedidos, en el orden de la declaración de su familia y,
	// dentro de cada variante, orden antes que herramienta.
	declarada declaracionDelLaboratorio
	familia   familiaDelLaboratorio
	pares     []parDelLaboratorio

	// porEval son las sesiones de cada eval en cada par: SESIONES/4.
	porEval int

	// conElProducto es INSTALACION=producto: cada sesión ve, además de la
	// skill de su variante, las demás de skills/ del árbol (FR-010 de H41.2).
	conElProducto bool

	modelo       string
	claudeCode   string
	concurrencia int
	directorio   string
}

// resumenDelPar es lo que la salida dice de un par: cuántas sesiones abrió y
// cuántas quedaron sin abrir tras el límite de uso.
type resumenDelPar struct {
	par      parDelLaboratorio
	abiertas int
	sinAbrir int
}

// sondearElLaboratorio abre las sesiones de un sondeo del laboratorio y devuelve
// su salida (contracts/arnes.md §2, §3 y §4 de H39, de H40 y de H41; data-model
// §6 de H39, §1 de H40 y §1, §3 y §4 de H41; research D8 de H39, D7 y D8 de H40
// y D14 de H41; FR-030, FR-031, FR-033 de H39; FR-008 y FR-010 de H40; FR-022,
// FR-030 y FR-032 de H41):
//
//  1. comprueba, antes de construir nada, los argumentos, la credencial y la
//     versión de Claude Code (comprobarElLaboratorio), y lee las evals del
//     directorio de evals de la familia pedida y, de ellas, las que planifica
//     su sondeo (evalsDelPlan);
//  2. prepara en el temporal kitlegal, laboratorio, home/ y la copia de las
//     skills de cada variante pedida, desde el directorio de su familia o,
//     la del producto, desde skills/<skill> del árbol, y, con
//     INSTALACION=producto, a su lado, la de cada otra skill de skills/ del
//     árbol (FR-010 de H41.2);
//  3. crea DIRECTORIO y su sondeo.json, con el modelo, la versión, si se
//     pidió la variante del producto, las referencias de su copia y, en una
//     familia del producto, las skills instaladas (skillsInstaladasDelSondeo);
//  4. por cada par, en su orden, abre con el repartidor, en
//     DIRECTORIO/<variante>/<modo>/, SESIONES/4 sesiones de cada eval del
//     sondeo de la familia, en su orden y de nombre <eval>-<nn>, con el modelo
//     pedido y el modo del par: preparadas con el directorio de evals de la
//     familia, sin traza, con la copia de las skills de la variante, el entorno
//     del sondeo de una skill (entornoDelSondeo) y el tope de 240 s con su
//     margen de 10 s; en el modo herramienta, con el servidor de la variante
//     (servidorDelPar). Tras un reparto que vuelve con sesiones sin abrir por el
//     límite de uso, no abre ninguna más;
//  5. compone la salida: una línea por par, con las abiertas y las que quedaron
//     sin abrir.
//
// Vuelve sin error aunque la cuenta no deje abrir todas las sesiones: las
// cifras son de la tabla, no de esta salida. El error es el de la comprobación
// —solo el de los argumentos, la credencial y la versión es un errorDeUso, y con
// él no construye nada, no abre ninguna sesión y no crea DIRECTORIO—, el de la
// declaración o las evals del laboratorio que no se pueden leer, el de una eval
// del sondeo de la familia que no está en su directorio de evals, también antes
// de construir nada, el de la preparación, el de DIRECTORIO o el del repartidor,
// también tras la interrupción, con interrupcion cerrado, que cierra las
// abiertas como en el sondeo de una skill.
func sondearElLaboratorio(interrupcion <-chan struct{}, s SondeoDelLaboratorio) (string, error) {
	comprobado, err := comprobarElLaboratorio(s)
	if err != nil {
		return "", err
	}

	temporal, err := filepath.Abs(s.Temporal)
	if err != nil {
		return "", fmt.Errorf("el temporal del sondeo del laboratorio %s no tiene ruta absoluta: %w", s.Temporal, err)
	}

	directorioDeLasEvals := filepath.Join(s.Raiz, filepath.FromSlash(comprobado.familia.evals))

	leidas, err := leerEvalsBienFormadas(directorioDeLasEvals)
	if err != nil {
		return "", fmt.Errorf("las evals del laboratorio no se pueden leer: %w", err)
	}

	evals, err := comprobado.familia.evalsDelPlan(leidas)
	if err != nil {
		return "", err
	}

	servidores := make([][]string, 0, len(comprobado.pares))

	for _, par := range comprobado.pares {
		servidor, err := servidorDelPar(s.Raiz, temporal, comprobado.declarada, par)
		if err != nil {
			return "", err
		}

		servidores = append(servidores, servidor)
	}

	referencias, err := s.preparar(temporal, comprobado)
	if err != nil {
		return "", err
	}

	instaladas, err := skillsInstaladasDelSondeo(temporal, comprobado.familia, comprobado.pares)
	if err != nil {
		return "", err
	}

	if err := crearElSondeoDelLaboratorio(comprobado.directorio, comprobado.modelo, comprobado.claudeCode,
		referencias, instaladas); err != nil {
		return "", err
	}

	resumenes := make([]resumenDelPar, 0, len(comprobado.pares))
	trasElLimite := false

	for posicion, par := range comprobado.pares {
		plan := planDelPar(evals, comprobado.porEval, comprobado.modelo, par.modo)

		if trasElLimite {
			resumenes = append(resumenes, resumenDelPar{par: par, sinAbrir: len(plan)})

			continue
		}

		ejecucion, err := ejecutarSesiones(interrupcion, SesionesAEjecutar{
			Plan:          plan,
			Concurrencia:  comprobado.concurrencia,
			Evals:         directorioDeLasEvals,
			Sesiones:      filepath.Join(comprobado.directorio, par.variante.Nombre, string(par.modo)),
			Skills:        filepath.Join(temporal, carpetaDeLasSkillsDelSondeo, par.variante.Nombre),
			Guion:         s.Guion,
			Binario:       filepath.Join(temporal, carpetaDeBinarios, programaDeLasConsultas),
			Servidor:      servidores[posicion],
			Entorno:       entornoDelSondeo(s.Entorno, temporal),
			Traza:         false,
			Tope:          topeDeUnaSesion,
			MargenDelTope: margenDelTopeDeUnaSesion,
		})
		if err != nil {
			return "", err
		}

		resumenes = append(resumenes, resumenDelPar{
			par: par, abiertas: len(ejecucion.Abiertas), sinAbrir: len(ejecucion.SinAbrir),
		})
		trasElLimite = len(ejecucion.SinAbrir) > 0
	}

	return salidaDelLaboratorio(comprobado, resumenes), nil
}

// comprobarElLaboratorio comprueba el sondeo del laboratorio antes de construir
// nada, en este orden (contracts/arnes.md §2 de H39; FR-030, FR-031): lee la
// declaración del laboratorio; comprueba sus argumentos, con todos sus errores
// juntos; y, solo si valen, la credencial y, solo si está, que el primer campo
// de claude --version, con el PATH de quien lo lanza, sea CLAUDE_CODE. Los
// errores de los argumentos, el de la credencial y el de la versión van
// envueltos en errorDeUso, con su mismo texto; el de una declaración que no se
// puede leer, no.
func comprobarElLaboratorio(s SondeoDelLaboratorio) (laboratorioComprobado, error) {
	declarada, err := leerLaDeclaracionDelRepositorio(s.Raiz)
	if err != nil {
		return laboratorioComprobado{}, err
	}

	comprobado, err := s.Argumentos.comprobar(declarada)
	if err != nil {
		return laboratorioComprobado{}, &errorDeUso{causa: err}
	}

	if valorEnElEntorno(s.Entorno, variableDeLaSuscripcion) == "" {
		return laboratorioComprobado{}, &errorDeUso{causa: errSinSuscripcion}
	}

	if err := comprobarLaVersionDeClaudeCode(s.Entorno, comprobado.claudeCode); err != nil {
		return laboratorioComprobado{}, &errorDeUso{causa: err}
	}

	return comprobado, nil
}

// variantesDeLaFamiliaDeH39 es cómo nombra el error de una mezcla las variantes
// de la familia de H39, la única que no es del producto (contracts/arnes.md §2
// de H40 y de H41; research D14).
const variantesDeLaFamiliaDeH39 = "variantes del laboratorio de H39"

// comprobar comprueba los argumentos del sondeo del laboratorio con las cuatro
// familias, la de la declaración de H39 y las tres del producto
// (familiasDelProducto), en ese orden (contracts/arnes.md §2 de H39, de H40 y de
// H41; research D8 de H40 y D14): VARIANTES, nombres de una variante de alguna
// de las cuatro, separados por comas y sin repetir; los que valen, todos de una
// sola familia; MODOS, orden, herramienta o los dos, separados por comas y sin
// repetir; cada par de una variante y un modo que valen, un modo de los de la
// variante, en el orden de las familias; SESIONES, un entero mayor que 0 y
// múltiplo de 4; MODELO, un identificador completo; CLAUDE_CODE, no vacía;
// CONCURRENCIA, un entero mayor o igual que 1; DIRECTORIO, no vacío y que no
// existe; e INSTALACION, vacía o producto, y producto solo con variantes de una
// familia del producto (errorDeLaInstalacion; contracts/arnes.md §2 de H41.2).
// Los errores de todos van juntos, uno por línea y en ese orden, para que quien
// lanza la orden los corrija de una vez.
func (a ArgumentosDelLaboratorio) comprobar(declarada declaracionDelLaboratorio) (laboratorioComprobado, error) {
	familias := slices.Concat([]familiaDelLaboratorio{declarada.familia()}, familiasDelProducto)

	variantes, errDeLasVariantes := listaDelArgumento(argumentoVariantes, a.Variantes,
		func(nombre string) bool {
			return slices.ContainsFunc(familias, func(familia familiaDelLaboratorio) bool { return familia.tiene(nombre) })
		},
		"no es ninguna variante del laboratorio")

	familia, errDeLaMezcla := familiaDeLasVariantes(familias, variantes)

	modos, errDeLosModos := listaDelArgumento(argumentoModos, a.Modos,
		func(modo Modo) bool { return modo == ModoOrden || modo == ModoHerramienta }, "no es orden ni herramienta")

	var (
		pares         []parDelLaboratorio
		errDeLosPares []error
	)

	for _, deLaFamilia := range familias {
		suyos, noValen := paresPedidos(deLaFamilia.variantes, variantes, modos)
		pares = append(pares, suyos...)
		errDeLosPares = append(errDeLosPares, noValen...)
	}

	porEval, errDeLasSesiones := sesionesDeCadaEval(a.Sesiones)
	concurrencia, errDeLaConcurrencia := enteroDelArgumento(argumentoConcurrencia, a.Concurrencia)

	errDeLaVersion := error(nil)
	if a.ClaudeCode == "" {
		errDeLaVersion = fmt.Errorf("%s: está vacía", argumentoClaudeCode)
	}

	err := errors.Join(slices.Concat(errDeLasVariantes, []error{errDeLaMezcla}, errDeLosModos, errDeLosPares,
		[]error{
			errDeLasSesiones, comprobarElModeloCompleto(a.Modelo), errDeLaVersion, errDeLaConcurrencia,
			comprobarElDirectorioDelSondeo(a.Directorio), errorDeLaInstalacion(a.Instalacion, familia),
		})...)
	if err != nil {
		return laboratorioComprobado{}, err
	}

	return laboratorioComprobado{
		declarada: declarada, familia: familia, pares: pares, porEval: porEval,
		conElProducto: a.Instalacion == instalacionDelProducto, modelo: a.Modelo, claudeCode: a.ClaudeCode,
		concurrencia: concurrencia, directorio: a.Directorio,
	}, nil
}

// errorDeLaInstalacion es el error de INSTALACION con la familia de las
// variantes pedidas que valen (contracts/arnes.md §2 de H41.2; research D7 de
// H41.2): ninguno vacía; «no es producto» con cualquier otro valor que no sea
// producto, haya familia o no; y con producto, el de la familia de H39, la única
// que no es del producto, porque la instalación del producto es la de las skills
// de skills/ junto a la de una familia suya. Sin familia —la vacía, sin
// variantes, porque ninguna vale o porque las que valen se mezclan—, producto
// no da error: ya lo dan el de VARIANTES o el de la mezcla.
func errorDeLaInstalacion(instalacion string, familia familiaDelLaboratorio) error {
	switch {
	case instalacion == "":
		return nil
	case instalacion != instalacionDelProducto:
		return fmt.Errorf("%s: «%s» no es %s", argumentoInstalacion, instalacion, instalacionDelProducto)
	case len(familia.variantes) > 0 && familia.producto == "":
		return fmt.Errorf("%s: las %s no son de ninguna familia del producto", argumentoInstalacion,
			variantesDeLaFamiliaDeH39)
	default:
		return nil
	}
}

// familiaDeLasVariantes es la familia de las variantes pedidas que valen, y
// errorDeLaMezcla si son de más de una. Sin ninguna, es la familia vacía, sin
// error: el de VARIANTES ya lo da su lista.
func familiaDeLasVariantes(familias []familiaDelLaboratorio, variantes []string) (familiaDelLaboratorio, error) {
	var pedidas []familiaDelLaboratorio

	for _, familia := range familias {
		if slices.ContainsFunc(variantes, familia.tiene) {
			pedidas = append(pedidas, familia)
		}
	}

	switch len(pedidas) {
	case 0:
		return familiaDelLaboratorio{}, nil
	case 1:
		return pedidas[0], nil
	default:
		return familiaDelLaboratorio{}, errorDeLaMezcla(pedidas)
	}
}

// errorDeLaMezcla es el error de unas VARIANTES que valen y son de más de una
// familia, las pedidas, en el orden de las familias (FR-030; contracts/arnes.md
// §2; research D14): «se mezclan» y las variantes de cada una, las de una
// familia del producto con su skill y las de la de H39 con
// variantesDeLaFamiliaDeH39, unidas con comas y la última con «y». Con la de H39
// y la de normativa es el de H40, byte a byte.
func errorDeLaMezcla(pedidas []familiaDelLaboratorio) error {
	nombradas := make([]string, 0, len(pedidas))

	for _, familia := range pedidas {
		nombradas = append(nombradas, familia.variantesEnLaMezcla())
	}

	return fmt.Errorf("%s: se mezclan %s", argumentoVariantes, unidasConY(nombradas))
}

// variantesEnLaMezcla es cómo nombra el error de una mezcla las variantes de la
// familia: «variantes de <skill>» las de una familia del producto, la que tiene
// variante del producto, y variantesDeLaFamiliaDeH39 las de la de H39, la única
// que no la tiene.
func (f familiaDelLaboratorio) variantesEnLaMezcla() string {
	if f.producto == "" {
		return variantesDeLaFamiliaDeH39
	}

	return "variantes de " + f.skill
}

// unidasConY une las partes con comas y la última con «y»: «a», «a y b», «a, b
// y c».
func unidasConY(partes []string) string {
	if len(partes) < 2 {
		return strings.Join(partes, "")
	}

	ultima := len(partes) - 1

	return strings.Join(partes[:ultima], ", ") + " y " + partes[ultima]
}

// listaDelArgumento son los elementos de la lista del argumento, separados por
// comas, que valen y no se repiten, en su orden, y un error por cada uno que no
// vale —«<argumento>: «<elemento>» <queNoEs>»— y por cada uno que se repite, una
// sola vez por elemento —«<argumento>: «<elemento>» se repite»—, en el orden en
// que aparecen. Un valor vacío es la lista de un elemento vacío, que no vale.
func listaDelArgumento[T ~string](argumento, valor string, vale func(T) bool, queNoEs string) ([]T, []error) {
	elementos := strings.Split(valor, separadorDeLaLista)

	var (
		validos []T
		noValen []error
	)

	for posicion, elemento := range elementos {
		anteriores := elementos[:posicion]

		switch {
		case slices.Contains(anteriores, elemento):
			if vecesEn(anteriores, elemento) == 1 {
				noValen = append(noValen, fmt.Errorf("%s: «%s» se repite", argumento, elemento))
			}
		case !vale(T(elemento)):
			noValen = append(noValen, fmt.Errorf("%s: «%s» %s", argumento, elemento, queNoEs))
		default:
			validos = append(validos, T(elemento))
		}
	}

	return validos, noValen
}

// vecesEn es cuántas veces está el elemento en la lista.
func vecesEn(lista []string, elemento string) int {
	veces := 0

	for _, otro := range lista {
		if otro == elemento {
			veces++
		}
	}

	return veces
}

// paresPedidos son los pares de las variantes declaradas que se piden, con los
// modos pedidos que valen, en el orden de la declaración y, dentro de cada
// variante, orden antes que herramienta (contracts/arnes.md §3.3 de H39), y un
// error por cada par cuyo modo no es uno de los de su variante (FR-020 a FR-027
// de H39; FR-010 de H40), en ese mismo orden.
func paresPedidos(declaradas []varianteDelLaboratorio, variantes []string, modos []Modo) (
	[]parDelLaboratorio, []error,
) {
	var (
		pares   []parDelLaboratorio
		noValen []error
	)

	for _, variante := range declaradas {
		if !slices.Contains(variantes, variante.Nombre) {
			continue
		}

		for _, modo := range []Modo{ModoOrden, ModoHerramienta} {
			switch {
			case !slices.Contains(modos, modo):
			case !slices.Contains(variante.Modos, modo):
				noValen = append(noValen, fmt.Errorf("%s y %s: %s no se mide en modo %s", argumentoVariantes,
					argumentoModos, variante.Nombre, modo))
			default:
				pares = append(pares, parDelLaboratorio{variante: variante, modo: modo})
			}
		}
	}

	return pares, noValen
}

// sesionesDeCadaEval son las sesiones de cada eval en cada par, SESIONES/4, si
// SESIONES es un entero mayor que 0, escrito sin nada más, que cabe en un int y
// es múltiplo de 4.
func sesionesDeCadaEval(valor string) (int, error) {
	sesiones, err := strconv.Atoi(valor)
	if err != nil || !formaDeUnEnteroPositivo.MatchString(valor) || sesiones%evalsDelLaboratorio != 0 {
		return 0, fmt.Errorf("%s: «%s» no es un múltiplo de %d mayor que 0", argumentoSesiones, valor,
			evalsDelLaboratorio)
	}

	return sesiones / evalsDelLaboratorio, nil
}

// comprobarElModeloCompleto exige un MODELO con la forma de un identificador
// completo de modelo (FR-031 de H39).
func comprobarElModeloCompleto(modelo string) error {
	if !formaDelModeloCompleto.MatchString(modelo) {
		return fmt.Errorf("%s: «%s» no es un identificador completo de modelo (claude-<familia>-<versión>, "+
			"como claude-sonnet-5-5)", argumentoModelo, modelo)
	}

	return nil
}

// comprobarElDirectorioDelSondeo exige un DIRECTORIO no vacío y que no exista:
// el arnés lo crea, y uno que ya existe mezclaría las sesiones de dos
// lanzamientos (research D8 de H39). Uno que no se puede mirar tampoco vale.
func comprobarElDirectorioDelSondeo(directorio string) error {
	if directorio == "" {
		return fmt.Errorf("%s: está vacío", argumentoDirectorio)
	}

	_, err := os.Lstat(directorio)

	switch {
	case err == nil:
		return fmt.Errorf("%s: %s ya existe", argumentoDirectorio, directorio)
	case errors.Is(err, fs.ErrNotExist):
		return nil
	default:
		return fmt.Errorf("%s: %s no se puede mirar: %w", argumentoDirectorio, directorio, err)
	}
}

// comprobarLaVersionDeClaudeCode exige que el primer campo de lo que escribe
// claude --version, con el entorno y el PATH de quien lanza el arnés, sea la
// versión pedida (contracts/arnes.md §2 de H39; research S1 de H39; FR-031).
func comprobarLaVersionDeClaudeCode(entorno []string, pedida string) error {
	instalada, err := versionDeClaudeCode(entorno)
	if err != nil {
		return fmt.Errorf("%s: %s %s no se pudo ejecutar: %w", argumentoClaudeCode, programaDeClaudeCode,
			banderaDeLaVersion, err)
	}

	if instalada != pedida {
		return fmt.Errorf("%s: la versión de Claude Code instalada es %s y se pidió %s", argumentoClaudeCode,
			instalada, pedida)
	}

	return nil
}

// versionDeClaudeCode es el primer campo de lo que escribe claude --version en
// su salida estándar: el claude que resuelve el PATH del entorno, con ese
// entorno. El error es el de un claude que no está en el PATH, el de uno que no
// termina con 0, con lo que escribió en su salida de error, y el de uno que no
// escribe nada.
func versionDeClaudeCode(entorno []string) (string, error) {
	claude, err := buscarEnLaRuta(valorEnElEntorno(entorno, variableDelPATH), programaDeClaudeCode)
	if err != nil {
		return "", err
	}

	orden := ordenDeLaVersion(claude)
	orden.Env = entorno

	salida, err := orden.Output()
	if err != nil {
		var terminado *exec.ExitError
		if errors.As(err, &terminado) && len(bytes.TrimSpace(terminado.Stderr)) > 0 {
			return "", fmt.Errorf("%w: %s", err, bytes.TrimSpace(terminado.Stderr))
		}

		return "", err
	}

	campos := strings.Fields(string(salida))
	if len(campos) == 0 {
		return "", errors.New("no escribió ninguna versión")
	}

	return campos[0], nil
}

// ordenDeLaVersion es claude --version con el claude de la ruta dada, el único
// argumento que no es constante (gosec G204).
func ordenDeLaVersion(claude string) *exec.Cmd {
	return exec.CommandContext(context.Background(), claude, banderaDeLaVersion)
}

// buscarEnLaRuta es la ruta del programa en la primera carpeta de la lista de
// rutas, como la de un PATH, que lo tiene como fichero ejecutable, como lo
// resuelve una shell. Una carpeta relativa, también la vacía, que es la de
// trabajo, no se mira: lo que se ejecuta no depende de dónde se lanza.
func buscarEnLaRuta(ruta, programa string) (string, error) {
	for _, carpeta := range filepath.SplitList(ruta) {
		if !filepath.IsAbs(carpeta) {
			continue
		}

		candidato := filepath.Join(carpeta, programa)

		estado, err := os.Stat(candidato)
		if err == nil && estado.Mode().IsRegular() && estado.Mode().Perm()&0o111 != 0 {
			return candidato, nil
		}
	}

	return "", fmt.Errorf("%s no está en ninguna carpeta del %s", programa, variableDelPATH)
}

// servidorDelPar es la orden del servidor de las sesiones del par
// (contracts/arnes.md §3.4 de H39): vacía, la de hoy, salvo en el modo
// herramienta de una variante que cambia las instrucciones o el parámetro, en la
// que es laboratorio servidor, del laboratorio-bin/ del temporal, delante del
// kitlegal de su bin/, con -instrucciones y el texto de su instrucciones.txt, o
// con -herramienta y -parametro de la declaración y -descripcion y el texto de
// su parametro.txt. El texto es el fichero sin su salto de línea final
// (contracts/laboratorio.md §4 de H39). El error es el de un fichero que no se
// puede leer.
func servidorDelPar(raiz, temporal string, declarada declaracionDelLaboratorio, par parDelLaboratorio) (
	[]string, error,
) {
	if par.modo != ModoHerramienta {
		return nil, nil
	}

	directorio := directorioDeLaVariante(raiz, par.variante.Nombre)

	var opciones []string

	if slices.Contains(par.variante.Cambia, parteInstrucciones) {
		texto, err := textoDeLaVariante(directorio, ficheroDeLasInstrucciones)
		if err != nil {
			return nil, err
		}

		opciones = append(opciones, "-"+banderaDeInstrucciones, texto)
	}

	if slices.Contains(par.variante.Cambia, parteParametro) {
		texto, err := textoDeLaVariante(directorio, ficheroDelParametro)
		if err != nil {
			return nil, err
		}

		opciones = append(opciones, "-"+banderaDeLaHerramienta, declarada.Herramienta, "-"+banderaDelParametro,
			declarada.Parametro, "-"+banderaDeLaDescripcion, texto)
	}

	if len(opciones) == 0 {
		return nil, nil
	}

	return append([]string{
		filepath.Join(temporal, carpetaDelProgramaDelLaboratorio, programaDelLaboratorio), ordenDelServidorDelLaboratorio,
		"-" + banderaDelKitlegal, filepath.Join(temporal, carpetaDeBinarios, programaDeLasConsultas),
	}, opciones...), nil
}

// textoDeLaVariante es el texto del fichero de la variante de ese directorio: el
// fichero sin su salto de línea final.
func textoDeLaVariante(directorio, fichero string) (string, error) {
	contenido, err := leerFichero(filepath.Join(directorio, fichero))
	if err != nil {
		return "", fmt.Errorf("el texto de la variante %s no se puede leer: %w", filepath.Base(directorio), err)
	}

	return strings.TrimSuffix(string(contenido), "\n"), nil
}

// preparar deja en el temporal, que recibe con su ruta absoluta, lo que
// necesitan las sesiones de los pares (contracts/arnes.md §3.1 de H39 y de H40
// y §3 de H41): kitlegal y laboratorio, con ConstruirLosProgramas; home/, vacío,
// el HOME de las sesiones; y en skills/<variante>/, una vez por variante y antes
// de abrir ninguna sesión, la copia de las skills de cada variante de los pares
// de la familia del sondeo comprobado, con las demás del producto si se pidió
// INSTALACION=producto, el directorio de skills de sus repartos
// (copiarLasSkillsDeLaVariante; contracts/arnes.md §3.1 de H41.2). Devuelve las
// referencias de la copia de la variante del producto, si alguno de los pares es
// suyo, y si no, ninguna. El error nombra lo que no se pudo hacer.
func (s SondeoDelLaboratorio) preparar(temporal string, comprobado laboratorioComprobado) ([]string, error) {
	if err := s.ConstruirLosProgramas(s.Raiz, temporal, s.Entorno); err != nil {
		return nil, err
	}

	personal := filepath.Join(temporal, carpetaPersonal)
	if err := os.Mkdir(personal, permisosDeLaSesion); err != nil {
		return nil, fmt.Errorf("el HOME del sondeo del laboratorio %s no se puede crear: %w", personal, err)
	}

	var (
		copiadas    []string
		referencias []string
	)

	for _, par := range comprobado.pares {
		if slices.Contains(copiadas, par.variante.Nombre) {
			continue
		}

		deLaCopia, err := s.copiarLasSkillsDeLaVariante(temporal, comprobado.familia, par.variante.Nombre,
			comprobado.conElProducto)
		if err != nil {
			return nil, err
		}

		copiadas = append(copiadas, par.variante.Nombre)
		referencias = append(referencias, deLaCopia...)
	}

	return referencias, nil
}

// copiarLasSkillsDeLaVariante deja en skills/<variante>/ del temporal la copia
// de las skills de la variante de ese nombre de la familia (contracts/arnes.md
// §3.1; data-model §3): la de su directorio en la familia, si es versionada; y
// si es la del producto, la de skills/<skill> del árbol, en
// skills/<variante>/<skill>/ (FR-022). Con conElProducto, INSTALACION=producto,
// deja además a su lado la de cada otra skill de skills/ del árbol
// (copiarLasDemasSkillsDelArbol; FR-010 de H41.2), así que en la del producto la
// copia es skills/ entera. De la del producto devuelve los nombres de los
// ficheros de references/ de la copia de su skill, en orden de nombre (FR-032);
// de una versionada, ninguno.
func (s SondeoDelLaboratorio) copiarLasSkillsDeLaVariante(temporal string, familia familiaDelLaboratorio,
	nombre string, conElProducto bool,
) ([]string, error) {
	destino := filepath.Join(temporal, carpetaDeLasSkillsDelSondeo, nombre)

	referencias, err := s.copiarLaSkillDeLaVariante(destino, familia, nombre)
	if err != nil {
		return nil, err
	}

	if !conElProducto {
		return referencias, nil
	}

	if err := s.copiarLasDemasSkillsDelArbol(destino, familia.skill, nombre); err != nil {
		return nil, err
	}

	return referencias, nil
}

// copiarLaSkillDeLaVariante deja en destino, skills/<variante>/ del temporal, la
// copia de la skill de la familia de la variante de ese nombre: la de su
// directorio en la familia, si es versionada, y si es la del producto, la de
// skills/<skill> del árbol, en destino/<skill>/, de la que devuelve sus
// referencias (referenciasDeLaCopia).
func (s SondeoDelLaboratorio) copiarLaSkillDeLaVariante(destino string, familia familiaDelLaboratorio,
	nombre string,
) ([]string, error) {
	if nombre != familia.producto {
		origen := filepath.Join(s.Raiz, filepath.FromSlash(familia.directorioDeLaVariante(nombre)),
			carpetaDeLasSkillsDeLaVariante)

		return nil, copiarLasSkills(nombre, origen, destino)
	}

	copia := filepath.Join(destino, familia.skill)
	if err := copiarLasSkills(nombre, filepath.Join(s.Raiz, directorioDeLasSkills, familia.skill), copia); err != nil {
		return nil, err
	}

	return referenciasDeLaCopia(copia)
}

// copiarLasDemasSkillsDelArbol deja en destino, skills/<variante>/ del temporal
// de la variante de ese nombre, la copia de cada directorio de skills/ del árbol
// que no es el de la skill de su familia, en destino/<otra>/, con os.CopyFS
// (copiarLasSkills), como las deja make install: byte a byte (contracts/arnes.md
// §3.1 de H41.2; research D10 de H41.2). La de la familia es la de la variante,
// que ya está en destino, así que ningún nombre se repite.
func (s SondeoDelLaboratorio) copiarLasDemasSkillsDelArbol(destino, skill, variante string) error {
	arbol := filepath.Join(s.Raiz, directorioDeLasSkills)

	entradas, err := os.ReadDir(arbol)
	if err != nil {
		return fmt.Errorf("las skills del árbol %s no se pueden leer: %w", arbol, err)
	}

	for _, entrada := range entradas {
		if !entrada.IsDir() || entrada.Name() == skill {
			continue
		}

		if err := copiarLasSkills(variante, filepath.Join(arbol, entrada.Name()),
			filepath.Join(destino, entrada.Name())); err != nil {
			return err
		}
	}

	return nil
}

// copiarLasSkills copia con os.CopyFS el directorio de origen en el de destino,
// que no existe, para la variante de ese nombre (research V9).
func copiarLasSkills(variante, origen, destino string) error {
	if err := os.CopyFS(destino, os.DirFS(origen)); err != nil {
		return fmt.Errorf("las skills de la variante %s no se pueden copiar en %s: %w", variante, destino, err)
	}

	return nil
}

// referenciasDeLaCopia son los nombres de los ficheros de references/ de la
// copia de una skill, en orden de nombre, el de os.ReadDir (research V9). Unas
// referencias que no se pueden leer son un error.
func referenciasDeLaCopia(copia string) ([]string, error) {
	directorio := filepath.Join(copia, directorioDeLasReferencias)

	entradas, err := os.ReadDir(directorio)
	if err != nil {
		return nil, fmt.Errorf("las referencias de la copia %s no se pueden leer: %w", directorio, err)
	}

	referencias := make([]string, 0, len(entradas))

	for _, entrada := range entradas {
		if entrada.Type().IsRegular() {
			referencias = append(referencias, entrada.Name())
		}
	}

	return referencias, nil
}

// skillsInstaladasDelSondeo son las skills que instala un sondeo de una familia
// del producto, la constancia de su sondeo.json (contracts/arnes.md §3.3 de
// H41.2; data-model §3 y research D11 de H41.2; FR-012 de H41.2): los nombres
// de las entradas de skills/<variante>/ del temporal, ya preparado, en el orden
// de nombre en que las da os.ReadDir, las que enlaza cada sesión de la variante
// (enlazarLasSkills). Las variantes de un sondeo instalan las mismas skills, así
// que basta la de su primer par. Un sondeo de H39 no la anota, y su sondeo.json
// es el de hoy: ninguna.
func skillsInstaladasDelSondeo(temporal string, familia familiaDelLaboratorio,
	pares []parDelLaboratorio,
) ([]string, error) {
	if familia.producto == "" || len(pares) == 0 {
		return nil, nil
	}

	directorio := filepath.Join(temporal, carpetaDeLasSkillsDelSondeo, pares[0].variante.Nombre)

	entradas, err := os.ReadDir(directorio)
	if err != nil {
		return nil, fmt.Errorf("las skills instaladas %s no se pueden leer: %w", directorio, err)
	}

	instaladas := make([]string, 0, len(entradas))

	for _, entrada := range entradas {
		instaladas = append(instaladas, entrada.Name())
	}

	return instaladas, nil
}

// construirLosProgramasDelLaboratorio deja en el temporal del arnés, que recibe
// con su ruta absoluta, kitlegal y laboratorio del árbol de trabajo de la raíz,
// no los que tenga instalados quien lo lanza (contracts/arnes.md §3.1 de H39):
// go install -trimpath de kitlegal en bin/ y de laboratorio en laboratorio-bin/,
// con el entorno de la base, su GOBIN y CGO_ENABLED=0, como el sondeo de una
// skill. El error nombra lo que no se pudo construir, con la orden y lo que
// escribió.
func construirLosProgramasDelLaboratorio(raiz, temporal string, base []string) error {
	installs := []struct {
		orden   *exec.Cmd
		carpeta string
	}{
		{ordenDeConstruirElBinario(raiz), carpetaDeBinarios},
		{ordenDeConstruirElLaboratorio(raiz), carpetaDelProgramaDelLaboratorio},
	}

	for _, install := range installs {
		install.orden.Env = sobreLaBase(base, []string{
			"GOBIN=" + filepath.Join(temporal, install.carpeta), "CGO_ENABLED=0",
		})

		if err := ejecutarLaOrdenDelArbol(install.orden); err != nil {
			return fmt.Errorf("kitlegal y laboratorio no se pudieron construir para el sondeo del laboratorio: %w", err)
		}
	}

	return nil
}

// ordenDeConstruirElLaboratorio es go install -trimpath del programa del
// laboratorio en la raíz dada, toda con constantes (gosec G204).
func ordenDeConstruirElLaboratorio(raiz string) *exec.Cmd {
	orden := exec.CommandContext(context.Background(), "go", "install", "-trimpath", paqueteDelLaboratorio)
	orden.Dir = raiz

	return orden
}

// crearElSondeoDelLaboratorio crea DIRECTORIO, con las carpetas que le falten
// delante, y su sondeo.json, con el modelo y la versión de Claude Code de todas
// sus sesiones y, si las hay, las referencias de la copia de la variante del
// producto y, si las hay, las skills instaladas de un sondeo de una familia del
// producto (contracts/arnes.md §3.2 de H39, §3.4 de H41 y §3.3 de H41.2;
// data-model §6 de H39, §4 de H41 y §3 de H41.2; FR-032 de H41; FR-012 de
// H41.2): sin ninguna de las dos, el de H39 byte a byte. Solo los nombres: ni la
// copia ni su huella se guardan en DIRECTORIO. DIRECTORIO se crea él solo, así
// que uno que ya existe es un error: ningún sondeo escribe donde escribe otro.
func crearElSondeoDelLaboratorio(directorio, modelo, version string, referencias, instaladas []string) error {
	limpio := filepath.Clean(directorio)

	if err := os.MkdirAll(filepath.Dir(limpio), permisosDeLaSesion); err != nil {
		return fmt.Errorf("el directorio del sondeo del laboratorio %s no se puede crear: %w", directorio, err)
	}

	if err := os.Mkdir(limpio, permisosDeLaSesion); err != nil {
		return fmt.Errorf("el directorio del sondeo del laboratorio %s no se puede crear: %w", directorio, err)
	}

	// sondeoLeido es lo que lee de sondeo.json el guion que cuenta el sondeo, y
	// sus claves van en el orden de contracts/arnes.md §3.2 de H39, con
	// referencias_del_producto detrás, solo si hay referencias (data-model §4), y
	// skills_instaladas detrás de todo, solo si hay skills instaladas
	// (data-model §3 de H41.2).
	contenido, err := json.Marshal(sondeoLeido{
		Modelo: modelo, ClaudeCode: version, ReferenciasDelProducto: referencias, SkillsInstaladas: instaladas,
	})
	if err != nil {
		return fmt.Errorf("%s no se puede codificar: %w", ficheroDelSondeo, err)
	}

	return escribirFichero(filepath.Join(limpio, ficheroDelSondeo), append(contenido, '\n'))
}

// planDelPar es el plan de un reparto (contracts/arnes.md §3.3 de H39 y §3.2 de
// H40): porEval sesiones de cada eval dada, en su orden, con el modelo y el modo
// dados, de nombre <eval sin .yaml>-<nn>, con nn desde 01 y dos cifras
// (data-model §6 de H39).
func planDelPar(evals []Eval, porEval int, modelo string, modo Modo) []SesionPlanificada {
	plan := make([]SesionPlanificada, 0, len(evals)*porEval)

	for _, eval := range evals {
		for numero := 1; numero <= porEval; numero++ {
			plan = append(plan, SesionPlanificada{
				Nombre:  fmt.Sprintf("%s-%02d", strings.TrimSuffix(eval.Fichero, extensionDeUnaEval), numero),
				Fichero: eval.Fichero,
				Modelo:  modelo,
				Modo:    modo,
			})
		}
	}

	return plan
}

// salidaDelLaboratorio es la salida del sondeo del laboratorio
// (contracts/arnes.md §4 de H39): la ruta, el modelo y la versión; una línea por
// par, en su orden, con las sesiones abiertas y, si quedaron, las que no se
// abrieron tras el límite de uso; y la orden que da la tabla.
func salidaDelLaboratorio(comprobado laboratorioComprobado, resumenes []resumenDelPar) string {
	lineas := []string{fmt.Sprintf("Sondeo del laboratorio en %s, con %s y Claude Code %s.", comprobado.directorio,
		comprobado.modelo, comprobado.claudeCode)}

	for _, resumen := range resumenes {
		linea := fmt.Sprintf("- %s, %s: %d abiertas", resumen.par.variante.Nombre, resumen.par.modo, resumen.abiertas)
		if resumen.sinAbrir > 0 {
			linea += fmt.Sprintf(", %d %s", resumen.sinAbrir, descripcionSinAbrir)
		}

		lineas = append(lineas, linea+".")
	}

	lineas = append(lineas, "La tabla: make laboratorio-tabla "+argumentoDirectorio+"="+comprobado.directorio)

	return strings.Join(lineas, "\n") + "\n"
}
