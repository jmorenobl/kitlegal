package evals

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jmorenobl/kitlegal/internal/app"
	"github.com/jmorenobl/kitlegal/internal/cache"
	"github.com/jmorenobl/kitlegal/internal/graph"
	"github.com/jmorenobl/kitlegal/internal/httpx"
)

// Lo que PrepararSesion encuentra y deja en el directorio de una sesión
// (contrato job-de-evals §3.2).
const (
	// directorioDeLaCache es el de la caché de la sesión, que ya existe vacío.
	directorioDeLaCache = "cache"

	// ficheroDeLaEval lleva el nombre de la eval con la que se juzga la sesión.
	ficheroDeLaEval = "eval.txt"

	// ficheroDeLaPregunta lleva lo que se pregunta en la sesión.
	ficheroDeLaPregunta = "pregunta.txt"

	// ficheroDelModelo lleva el id del modelo con el que se abre la sesión.
	ficheroDelModelo = "modelo.txt"
)

// textoDeLaPruebaDeRed es el texto literal que la sesión de prueba de red añade
// a la pregunta de su eval: dos invocaciones de un bloque que no está en ninguna
// grabación, sin --offline y con él (contrato job-de-evals §6; SC-012), con
// kitlegal desde el PATH, como lo invocan las skills (H19,
// contracts/skills-e-invocacion.md §6; FR-127). Lo pone el job, nunca SKILL.md
// (FR-077).
const textoDeLaPruebaDeRed = "Antes de responder, ejecuta también exactamente estas dos órdenes y di qué " +
	"devolvieron: `kitlegal boe articulo BOE-A-2015-10565 a9998 --json` y " +
	"`kitlegal boe articulo BOE-A-2015-10565 a9998 --offline --json`."

// programaDeLasConsultas es el nombre de programa con el que se invoca cada
// consulta: el del binario, que despacha por el primer argumento.
const programaDeLasConsultas = "kitlegal"

// sinDatosDeConstruccion son la versión, el commit y la fecha que reciben las
// invocaciones en proceso: solo los muestra el verbo reservado version, que
// ninguna consulta necesaria usa.
const sinDatosDeConstruccion = ""

// PuntoGrafoPrevio no es de los tres de data-model §7.1: es el de un comando del
// grafo previo de una eval, que no es una consulta que la caché preparada tenga
// que servir, sino una de las que llenan el grafo de la sesión antes de
// prepararla (contrato evals-y-skill §3 de H7).
const PuntoGrafoPrevio Punto = "grafo previo"

// Falta es una consulta necesaria que no terminó en 0 al preparar la caché o al
// comprobarla sin red: lo grabado, o la caché preparada, no la sirve (FR-074,
// FR-075). También lo es un comando del grafo previo que no terminó en 0 o que
// escribió algo en la salida de error (contrato evals-y-skill §3 de H7). Su eval
// y su punto son los Origenes de la consulta.
type Falta struct {
	// Consulta es la que no terminó en 0, o el comando del grafo previo que
	// escribió en la salida de error, con cada eval y punto de los que sale.
	Consulta Consulta

	// Codigo es el código de salida con el que terminó.
	Codigo int

	// Mensaje es lo que escribió en la salida de error, sin el salto de línea
	// final: nombra lo que faltaba, la petición sin grabación o la entrada que la
	// caché no tiene, o la entrega al grafo que falló.
	Mensaje string
}

// String nombra la falta desde cada origen de su consulta, una línea por
// origen: la eval y el comando esperado; la eval, la norma y lo que le falta, el
// índice o los metadatos; o la eval y la cita esperada sin su bloque (FR-075).
// Cada línea termina con la invocación, su código y el mensaje. Una consulta sin
// orígenes da solo la invocación, su código y el mensaje.
func (f Falta) String() string {
	causa := fmt.Sprintf("«%s» terminó con código %d", ordenDe(f.Consulta), f.Codigo)
	if f.Mensaje != "" {
		causa += ": " + f.Mensaje
	}

	if len(f.Consulta.Origenes) == 0 {
		return causa
	}

	lineas := make([]string, 0, len(f.Consulta.Origenes))
	for _, origen := range f.Consulta.Origenes {
		lineas = append(lineas, origen.Eval+": "+queFaltaDesde(origen.Punto, f.Consulta)+": "+causa)
	}

	return strings.Join(lineas, "\n")
}

// queFaltaDesde dice, en los términos del punto de data-model §7.1 del que sale
// la consulta, o del grafo previo, qué es lo que falta.
func queFaltaDesde(punto Punto, consulta Consulta) string {
	switch punto {
	case PuntoComandoEsperado:
		return "el comando esperado " + ordenDe(consulta)
	case PuntoNormaDeLaEval:
		return fmt.Sprintf("la norma %s sin %s", strings.Join(consulta.Argumentos, " "), consulta.Verbo)
	case PuntoCitaEsperada:
		return "la cita esperada " + strings.Join(consulta.Argumentos, " ") + " sin su bloque"
	case PuntoGrafoPrevio:
		return "el comando del grafo previo " + ordenDe(consulta)
	default:
		return fmt.Sprintf("la consulta %s del punto %q", ordenDe(consulta), punto)
	}
}

// ordenDe es la invocación de una consulta sin el nombre del programa, como la
// nombran los fallos: boe articulo BOE-A-2015-10565 a21.
func ordenDe(consulta Consulta) string {
	return strings.Join(slices.Concat([]string{consulta.Applet, consulta.Verbo}, consulta.Argumentos), " ")
}

// Preparar llena la caché de dirCache con las consultas, en proceso y sin red
// (contrato evals-y-grabaciones §5.1; research.md D14):
//
//  1. copia en un directorio temporal los conjuntos de grabaciones en el orden
//     de grabaciones, cada uno encima del anterior;
//  2. registra, en el valor cero de app.Registro, el applet boe sobre
//     httpx.Replay de ese temporal, que no abre ninguna conexión, y la caché en
//     dirCache;
//  3. ejecuta cada consulta con app.Main y --json, con búferes nuevos por
//     consulta;
//  4. cada consulta que termina con un código distinto de 0 es una Falta.
//
// El error queda para lo que impide preparar, nombrando el directorio o el
// applet: el temporal que no se crea o no se retira, un conjunto que no se copia
// o el applet que Registrar rechaza. Nunca se convierte en una lista de faltas
// vacía. No modifica ni la fuente ni el kernel: compone lo que exportan.
//
// No recibe un contexto: cada consulta es una invocación entera del binario por
// app.Main, la raíz de composición, que abre el suyo con el plazo de --timeout y
// no admite el de quien llama. Un contexto que no llegara a las invocaciones
// sería una cancelación a medias (research.md D14 y V59).
func Preparar(dirCache string, grabaciones []string, consultas []Consulta) (_ []Falta, err error) {
	reproduccion, err := os.MkdirTemp("", "kitlegal-evals-grabaciones-")
	if err != nil {
		return nil, fmt.Errorf("preparar la caché %s: el directorio temporal de las grabaciones no se puede crear: %w",
			dirCache, err)
	}

	defer func() { err = errors.Join(err, retirarTemporal(reproduccion)) }()

	for _, conjunto := range grabaciones {
		if err := copiarGrabaciones(conjunto, reproduccion); err != nil {
			return nil, fmt.Errorf("preparar la caché %s: %w", dirCache, err)
		}
	}

	registro, err := registroDeBoe(reproduccion, cache.ConDirectorio(dirCache))
	if err != nil {
		return nil, fmt.Errorf("preparar la caché %s: %w", dirCache, err)
	}

	return ejecutarConsultas(registro, consultas, conOtroCodigo, "--json"), nil
}

// ComprobarSinRed ejecuta cada consulta con --offline sobre la caché de
// dirCache (contrato evals-y-grabaciones §5.2; FR-075): el applet boe se
// registra como en Preparar, pero sobre httpx.Replay de un directorio temporal
// vacío, para que una petición que se escapara fallara en lugar de salir a la
// red, y con la caché en dirCache más las opciones, que solo usa un test para
// adelantar su reloj. Cada consulta que termina con un código distinto de 0 es
// una Falta, que nombra la eval y el comando, o la eval, la norma y lo que falta.
//
// El error, como en Preparar, queda para lo que impide comprobar: el directorio
// vacío que no se crea o no se retira, o el registro que no se monta. Tampoco
// recibe un contexto, por lo mismo que Preparar.
func ComprobarSinRed(dirCache string, consultas []Consulta, opciones ...cache.Opcion) (_ []Falta, err error) {
	vacio, err := os.MkdirTemp("", "kitlegal-evals-sin-grabaciones-")
	if err != nil {
		return nil, fmt.Errorf("comprobar sin red la caché %s: el directorio vacío de reproducción no se puede crear: %w",
			dirCache, err)
	}

	defer func() { err = errors.Join(err, retirarTemporal(vacio)) }()

	registro, err := registroDeBoe(vacio, slices.Concat([]cache.Opcion{cache.ConDirectorio(dirCache)}, opciones)...)
	if err != nil {
		return nil, fmt.Errorf("comprobar sin red la caché %s: %w", dirCache, err)
	}

	return ejecutarConsultas(registro, consultas, conOtroCodigo, "--offline", "--json"), nil
}

// SesionAPreparar es lo que PrepararSesion necesita para preparar el directorio
// de una sesión del job de evals (contrato job-de-evals §3.2).
type SesionAPreparar struct {
	// Evals es el directorio de las evals de la skill: se preparan las consultas
	// de todas.
	Evals string

	// Grabaciones son los conjuntos de grabaciones, en el orden en que los copia
	// Preparar.
	Grabaciones []string

	// GrafosPrevios es el directorio en el que el grafo previo de la eval de la
	// sesión es el subdirectorio que nombra su grafo_previo. Vacío, el de la
	// constante GrafosPrevios, el del repositorio, que es el del job; los tests
	// dan uno temporal.
	GrafosPrevios string

	// Fichero es el nombre, dentro de Evals, de la eval con la que se juzga la
	// sesión.
	Fichero string

	// Modelo es el id del modelo con el que se abre la sesión, que queda escrito
	// en modelo.txt: de él sale la serie a la que la sesión pertenece en el
	// informe (data-model §10.4).
	Modelo string

	// Directorio es el de la sesión; su cache/ ya existe y está vacío.
	Directorio string

	// PruebaDeRed dice si la pregunta lleva además el texto de la prueba de red.
	PruebaDeRed bool
}

// PrepararSesion prepara el directorio de una sesión del job de evals (contrato
// job-de-evals §3.2; research.md D14):
//
//  1. lee las evals con LeerConjunto; su error, o cualquier fichero mal
//     formado, que nombra con su error, termina sin preparar ni escribir nada;
//  2. si Fichero no es ninguna de las evals, o si Modelo no tiene la forma de un
//     id de modelo, termina con un error que lo nombra, sin preparar ni escribir
//     nada;
//  3. si la eval de Fichero lleva grafo_previo, y solo la suya, antes de llenar
//     cache/ prepara con él el grafo del mundo de cache/ (prepararGrafoPrevio);
//     sus faltas o su error se devuelven tal cual, sin preparar la caché ni
//     escribir nada más;
//  4. prepara cache/ con Preparar y las consultas necesarias de todas las evals,
//     no solo las de Fichero, para que ninguna eval dependa del orden; sus faltas
//     o su error se devuelven tal cual, sin escribir nada más. Su registro no
//     entrega al grafo: el de la sesión queda con lo del grafo previo;
//  5. escribe eval.txt con Fichero, modelo.txt con Modelo y pregunta.txt con la
//     pregunta de esa eval o, con PruebaDeRed, con la pregunta, una línea en
//     blanco y el texto de la prueba de red, cada uno con un salto de línea final.
func PrepararSesion(s SesionAPreparar) ([]Falta, error) {
	if !formaDelModelo.MatchString(s.Modelo) {
		return nil, fmt.Errorf("el modelo %q de la sesión %s no tiene la forma de un id de modelo", s.Modelo, s.Directorio)
	}

	evals, err := leerEvalsBienFormadas(s.Evals)
	if err != nil {
		return nil, err
	}

	posicion := slices.IndexFunc(evals, func(eval Eval) bool { return eval.Fichero == s.Fichero })
	if posicion < 0 {
		return nil, fmt.Errorf("la eval %s no es ninguna de las evals bien formadas de %s", s.Fichero, s.Evals)
	}

	eval := evals[posicion]

	faltas, err := prepararLaCache(s, evals, eval)
	if len(faltas) > 0 || err != nil {
		return faltas, err
	}

	pregunta := eval.Pregunta
	if s.PruebaDeRed {
		pregunta += "\n\n" + textoDeLaPruebaDeRed
	}

	escritos := []struct{ fichero, contenido string }{
		{ficheroDeLaEval, s.Fichero},
		{ficheroDelModelo, s.Modelo},
		{ficheroDeLaPregunta, pregunta},
	}
	for _, escrito := range escritos {
		if err := escribirFichero(filepath.Join(s.Directorio, escrito.fichero), []byte(escrito.contenido+"\n")); err != nil {
			return nil, fmt.Errorf("la sesión %s: %w", s.Directorio, err)
		}
	}

	return nil, nil
}

// leerEvalsBienFormadas lee las evals del directorio con LeerConjunto y las
// devuelve, en orden de nombre, solo si ningún fichero suyo está mal formado:
// el error de LeerConjunto va tal cual, y con algún fichero mal formado el error
// nombra el directorio y cada uno con su motivo. Es la lectura del paso 1 de
// PrepararSesion y la de la reconstrucción de los casos etiquetados del juez
// (contracts/medida-del-juez.md §5 de H24): ninguna de las dos prepara nada con
// un conjunto a medias.
func leerEvalsBienFormadas(dir string) ([]Eval, error) {
	conjunto, err := LeerConjunto(dir)
	if err != nil {
		return nil, err
	}

	if len(conjunto.MalFormados) > 0 {
		motivos := make([]error, 0, len(conjunto.MalFormados))
		for _, malFormado := range conjunto.MalFormados {
			motivos = append(motivos, malFormado.Error)
		}

		return nil, fmt.Errorf("el directorio de evals %s tiene ficheros mal formados:\n%w", dir, errors.Join(motivos...))
	}

	return conjunto.Evals, nil
}

// prepararLaCache hace los pasos 3 y 4 de PrepararSesion en cache/ de la
// sesión: el grafo previo de la eval, si lo lleva, y después, si no dio faltas
// ni error, la caché con las consultas necesarias de todas las evals.
func prepararLaCache(s SesionAPreparar, evals []Eval, eval Eval) ([]Falta, error) {
	dirCache := filepath.Join(s.Directorio, directorioDeLaCache)

	if eval.GrafoPrevio.Grabaciones != "" {
		faltas, err := prepararGrafoPrevio(dirCache, s.Grabaciones, cmp.Or(s.GrafosPrevios, GrafosPrevios), eval)
		if len(faltas) > 0 || err != nil {
			return faltas, err
		}
	}

	return Preparar(dirCache, s.Grabaciones, ConsultasNecesarias(evals))
}

// prepararGrafoPrevio llena el grafo del mundo de dirCache, el directorio de la
// caché de la sesión, con el grafo previo de la eval, en proceso y sin red
// (contrato evals-y-skill §3 de H7; research D26; FR-085):
//
//  1. copia en un directorio temporal los conjuntos de grabaciones, en su orden,
//     y encima el del grafo previo, el subdirectorio de grafosPrevios que nombra;
//  2. registra, como Preparar, el applet boe sobre httpx.Replay de ese
//     temporal, pero con la caché en otro temporal que se descarta, y entrega al
//     grafo del mundo de dirCache;
//  3. ejecuta cada comando del grafo previo, en su orden, como boe articulo
//     <norma> <bloque> con app.Main y --json;
//  4. cada comando que termina con un código distinto de 0, o que escribe
//     cualquier cosa en la salida de error —donde el kernel avisa, con el código
//     0, de una entrega que falló—, es una Falta con el punto PuntoGrafoPrevio.
//
// Nada de lo que lee llega a la caché de la sesión, que después prepara
// Preparar con lo grabado: el grafo tiene la versión derivada y la caché, la
// grabada. El error, como en Preparar, queda para lo que impide preparar,
// nombrando el grafo previo y su eval: un temporal que no se crea o no se
// retira, un conjunto que no se copia o el applet que Registrar rechaza.
func prepararGrafoPrevio(dirCache string, grabaciones []string, grafosPrevios string, eval Eval) (_ []Falta, err error) {
	previo := eval.GrafoPrevio
	contexto := fmt.Sprintf("preparar el grafo previo %s de la eval %s", previo.Grabaciones, eval.Fichero)

	reproduccion, err := os.MkdirTemp("", "kitlegal-evals-grafo-previo-grabaciones-")
	if err != nil {
		return nil, fmt.Errorf("%s: el directorio temporal de las grabaciones no se puede crear: %w", contexto, err)
	}

	defer func() { err = errors.Join(err, retirarTemporal(reproduccion)) }()

	cacheQueSeDescarta, err := os.MkdirTemp("", "kitlegal-evals-grafo-previo-cache-")
	if err != nil {
		return nil, fmt.Errorf("%s: el directorio temporal de la caché no se puede crear: %w", contexto, err)
	}

	defer func() { err = errors.Join(err, retirarTemporal(cacheQueSeDescarta)) }()

	for _, conjunto := range slices.Concat(grabaciones, []string{filepath.Join(grafosPrevios, previo.Grabaciones)}) {
		if err := copiarGrabaciones(conjunto, reproduccion); err != nil {
			return nil, fmt.Errorf("%s: %w", contexto, err)
		}
	}

	registro, err := registroDeBoe(reproduccion, cache.ConDirectorio(cacheQueSeDescarta))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", contexto, err)
	}

	registro.EntregarAlGrafo(graph.Nuevo(graph.ConDirectorio(dirCache)))

	origen := Origen{Eval: eval.Fichero, Punto: PuntoGrafoPrevio}
	consultas := make([]Consulta, 0, len(previo.Comandos))

	for _, comando := range previo.Comandos {
		consultas = append(consultas, Consulta{
			Applet:     comando.Applet,
			Verbo:      verboArticulo,
			Argumentos: []string{comando.Norma, comando.Bloque},
			Origenes:   []Origen{origen},
		})
	}

	return ejecutarConsultas(registro, consultas, conOtroCodigoOSalidaDeError, "--json"), nil
}

// registroDeBoe monta el registro de las consultas en proceso: el valor cero de
// app.Registro con el applet boe, cuyo cliente es httpx.Replay de grabaciones y
// cuya caché se abre con las opciones (contrato evals-y-grabaciones §5.1).
func registroDeBoe(grabaciones string, opciones ...cache.Opcion) (*app.Registro, error) {
	applet := app.AppletBoe(app.DependenciasDeBoe{
		Cliente: func(*slog.Logger) (*httpx.Cliente, error) { return httpx.Replay(grabaciones) },
		Cache:   opciones,
	})

	var registro app.Registro
	if err := registro.Registrar(applet); err != nil {
		return nil, fmt.Errorf("el applet %s no se puede registrar: %w", applet.Nombre(), err)
	}

	return &registro, nil
}

// criterioDeFalta dice si una invocación que terminó con el código y escribió
// errores en la salida de error es una Falta.
type criterioDeFalta func(codigo int, errores string) bool

// conOtroCodigo es el criterio de Preparar y ComprobarSinRed: la consulta que
// termina con un código distinto de 0 (FR-074, FR-075).
func conOtroCodigo(codigo int, _ string) bool {
	return codigo != 0
}

// conOtroCodigoOSalidaDeError es el de los comandos del grafo previo: además,
// el que escribe cualquier cosa en la salida de error, como la línea con la que
// el kernel avisa de una entrega al grafo que falló y que termina con 0
// (contrato evals-y-skill §3 de H7).
func conOtroCodigoOSalidaDeError(codigo int, errores string) bool {
	return codigo != 0 || errores != ""
}

// ejecutarConsultas invoca cada consulta, en su orden, con app.Main, las
// banderas detrás de sus argumentos y búferes nuevos, y devuelve como Falta cada
// una que lo es según el criterio, con copias de sus argumentos y sus orígenes.
func ejecutarConsultas(registro *app.Registro, consultas []Consulta, esFalta criterioDeFalta, banderas ...string) []Falta {
	var faltas []Falta

	for _, consulta := range consultas {
		argv := slices.Concat([]string{programaDeLasConsultas, consulta.Applet, consulta.Verbo}, consulta.Argumentos, banderas)

		var salida, errores bytes.Buffer

		codigo := app.Main(argv, registro, &salida, &errores,
			sinDatosDeConstruccion, sinDatosDeConstruccion, sinDatosDeConstruccion)
		if esFalta(codigo, errores.String()) {
			faltas = append(faltas, Falta{
				Consulta: Consulta{
					Applet:     consulta.Applet,
					Verbo:      consulta.Verbo,
					Argumentos: slices.Clone(consulta.Argumentos),
					Origenes:   slices.Clone(consulta.Origenes),
				},
				Codigo:  codigo,
				Mensaje: strings.TrimSuffix(errores.String(), "\n"),
			})
		}
	}

	return faltas
}

// copiarGrabaciones copia en destino cada fichero del conjunto de grabaciones
// de origen, encima del que tuviera el mismo nombre. Una entrada que no es un
// fichero regular no se salta: el conjunto no se puede copiar.
func copiarGrabaciones(origen, destino string) error {
	entradas, err := os.ReadDir(origen)
	if err != nil {
		return fmt.Errorf("el conjunto de grabaciones %s no se puede copiar: %w", origen, err)
	}

	for _, entrada := range entradas {
		if !entrada.Type().IsRegular() {
			return fmt.Errorf("el conjunto de grabaciones %s no se puede copiar: %s no es un fichero regular",
				origen, entrada.Name())
		}

		contenido, err := leerFichero(filepath.Join(origen, entrada.Name()))
		if err != nil {
			return fmt.Errorf("el conjunto de grabaciones %s no se puede copiar: %w", origen, err)
		}

		if err := escribirFichero(filepath.Join(destino, entrada.Name()), contenido); err != nil {
			return fmt.Errorf("el conjunto de grabaciones %s no se puede copiar: %w", origen, err)
		}
	}

	return nil
}

// leerFichero lee un fichero entero. filepath.Clean es lo que el control de
// rutas reconoce como saneado antes de abrirlo (gosec G304): la ruta la compone
// este paquete con un directorio que recibe y el nombre de una de sus entradas.
func leerFichero(ruta string) ([]byte, error) {
	return os.ReadFile(filepath.Clean(ruta))
}

// escribirFichero escribe un fichero entero con permisos 0o600, creándolo o
// reemplazando el que hubiera. Es un auxiliar distinto del que lee, para que la
// escritura no reciba en la misma función nada derivado de lo leído (gosec
// G703).
func escribirFichero(ruta string, contenido []byte) error {
	return os.WriteFile(filepath.Clean(ruta), contenido, 0o600)
}

// retirarTemporal borra un directorio temporal con todo lo que contiene.
func retirarTemporal(dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("el directorio temporal %s no se puede retirar: %w", dir, err)
	}

	return nil
}
