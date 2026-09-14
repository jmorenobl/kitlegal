package evals

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jmorenobl/kitlegal/internal/app"
	"github.com/jmorenobl/kitlegal/internal/cache"
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
)

// textoDeLaPruebaDeRed es el texto literal que la sesión de prueba de red añade
// a la pregunta de su eval: dos invocaciones de un bloque que no está en ninguna
// grabación, sin --offline y con él (contrato job-de-evals §6; SC-012). Lo pone
// el job, nunca SKILL.md (FR-077).
const textoDeLaPruebaDeRed = "Antes de responder, ejecuta también exactamente estas dos órdenes y di qué " +
	"devolvieron: `~/.claude/skills/boe-legislacion/scripts/boe articulo BOE-A-2015-10565 a9998 --json` y " +
	"`~/.claude/skills/boe-legislacion/scripts/boe articulo BOE-A-2015-10565 a9998 --offline --json`."

// programaDeLasConsultas es el nombre de programa con el que se invoca cada
// consulta: el del binario, que despacha por el primer argumento.
const programaDeLasConsultas = "kitlegal"

// sinDatosDeConstruccion son la versión, el commit y la fecha que reciben las
// invocaciones en proceso: solo los muestra el verbo reservado version, que
// ninguna consulta necesaria usa.
const sinDatosDeConstruccion = ""

// Falta es una consulta necesaria que no terminó en 0 al preparar la caché o al
// comprobarla sin red: lo grabado, o la caché preparada, no la sirve (FR-074,
// FR-075). Su eval y su punto son los Origenes de la consulta.
type Falta struct {
	// Consulta es la que no terminó en 0, con cada eval y punto de los que sale.
	Consulta Consulta

	// Codigo es el código de salida con el que terminó.
	Codigo int

	// Mensaje es lo que escribió en la salida de error, sin el salto de línea
	// final: nombra lo que faltaba, la petición sin grabación o la entrada que la
	// caché no tiene.
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
// la consulta, qué es lo que falta.
func queFaltaDesde(punto Punto, consulta Consulta) string {
	switch punto {
	case PuntoComandoEsperado:
		return "el comando esperado " + ordenDe(consulta)
	case PuntoNormaDeLaEval:
		return fmt.Sprintf("la norma %s sin %s", strings.Join(consulta.Argumentos, " "), consulta.Verbo)
	case PuntoCitaEsperada:
		return "la cita esperada " + strings.Join(consulta.Argumentos, " ") + " sin su bloque"
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

	return ejecutarConsultas(registro, consultas, "--json"), nil
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

	return ejecutarConsultas(registro, consultas, "--offline", "--json"), nil
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

	// Fichero es el nombre, dentro de Evals, de la eval con la que se juzga la
	// sesión.
	Fichero string

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
//  2. si Fichero no es ninguna de las evals, termina con un error que lo nombra,
//     sin preparar ni escribir nada;
//  3. prepara cache/ con Preparar y las consultas necesarias de todas las evals,
//     no solo las de Fichero, para que ninguna eval dependa del orden; sus faltas
//     o su error se devuelven tal cual, sin escribir nada más;
//  4. escribe eval.txt con Fichero y pregunta.txt con la pregunta de esa eval o,
//     con PruebaDeRed, con la pregunta, una línea en blanco y el texto de la
//     prueba de red, cada uno con un salto de línea final.
func PrepararSesion(s SesionAPreparar) ([]Falta, error) {
	conjunto, err := LeerConjunto(s.Evals)
	if err != nil {
		return nil, err
	}

	if len(conjunto.MalFormados) > 0 {
		motivos := make([]error, 0, len(conjunto.MalFormados))
		for _, malFormado := range conjunto.MalFormados {
			motivos = append(motivos, malFormado.Error)
		}

		return nil, fmt.Errorf("el directorio de evals %s tiene ficheros mal formados:\n%w", s.Evals, errors.Join(motivos...))
	}

	posicion := slices.IndexFunc(conjunto.Evals, func(eval Eval) bool { return eval.Fichero == s.Fichero })
	if posicion < 0 {
		return nil, fmt.Errorf("la eval %s no es ninguna de las evals bien formadas de %s", s.Fichero, s.Evals)
	}

	faltas, err := Preparar(filepath.Join(s.Directorio, directorioDeLaCache), s.Grabaciones,
		ConsultasNecesarias(conjunto.Evals))
	if len(faltas) > 0 || err != nil {
		return faltas, err
	}

	pregunta := conjunto.Evals[posicion].Pregunta
	if s.PruebaDeRed {
		pregunta += "\n\n" + textoDeLaPruebaDeRed
	}

	if err := escribirFichero(filepath.Join(s.Directorio, ficheroDeLaEval), []byte(s.Fichero+"\n")); err != nil {
		return nil, fmt.Errorf("la sesión %s: %w", s.Directorio, err)
	}

	if err := escribirFichero(filepath.Join(s.Directorio, ficheroDeLaPregunta), []byte(pregunta+"\n")); err != nil {
		return nil, fmt.Errorf("la sesión %s: %w", s.Directorio, err)
	}

	return nil, nil
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

// ejecutarConsultas invoca cada consulta, en su orden, con app.Main, las
// banderas detrás de sus argumentos y búferes nuevos, y devuelve como Falta cada
// una que termina con un código distinto de 0, con copias de sus argumentos y
// sus orígenes.
func ejecutarConsultas(registro *app.Registro, consultas []Consulta, banderas ...string) []Falta {
	var faltas []Falta

	for _, consulta := range consultas {
		argv := slices.Concat([]string{programaDeLasConsultas, consulta.Applet, consulta.Verbo}, consulta.Argumentos, banderas)

		var salida, errores bytes.Buffer

		codigo := app.Main(argv, registro, &salida, &errores,
			sinDatosDeConstruccion, sinDatosDeConstruccion, sinDatosDeConstruccion)
		if codigo != 0 {
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
