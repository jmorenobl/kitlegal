package evals

import (
	"cmp"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// Lo que EscribirInforme lee del directorio de una sesión además de lo que lee
// LeerSesion, y lo que deja en su destino (contrato job-de-evals §3.2, §3.3 y
// §5).
const (
	// directorioDeLaTraza es el traza/ de la sesión, con un fichero por hilo.
	directorioDeLaTraza = "traza"

	ficheroDelInformeMD   = "informe.md"
	ficheroDelInformeJSON = "informe.json"
)

// Veredicto es el veredicto global del informe de una ejecución del job de evals
// (data-model §10.3).
type Veredicto string

const (
	// VeredictoAprobado es el de una ejecución sin ficheros de eval mal formados,
	// con cada eval bien formada juzgada por alguna sesión, todas las evals
	// pasando y ninguna petición llegada a la red.
	VeredictoAprobado Veredicto = "aprobado"

	// VeredictoFallo es el de cualquier otra ejecución.
	VeredictoFallo Veredicto = "fallo"
)

// Motivos de texto fijo del informe (data-model §10.2 y §10.3; contrato
// job-de-evals §3.3). Los de una sesión ilegible van seguidos del fichero y su
// error; los de la raíz, detrás del fichero o de la sesión de los que hablan.
const (
	motivoDeSesionIlegible    = "sesión ilegible: "
	motivoSinNingunaSesion    = ": sin ninguna sesión"
	motivoSinEvalsQueJuzgar   = "ninguna eval bien formada que juzgar"
	motivoDeFicheroMalFormado = ": mal formado: "
	motivoDeLlegadaALaRed     = ": petición llegada a la red: "
)

// Textos fijos de informe.md (contrato job-de-evals §5).
const (
	ningunoEnElInforme  = "ninguno"
	ningunaEnElInforme  = "ninguna"
	sinLlegadasALaRed   = "ninguna petición llegó a la red de una fuente"
	sinCodigoPorElCorte = "sin código (sesión cortada)"
	sinConexiones       = "sin conexiones"
	sinLeer             = "sin leer"
	vaciaEnElInforme    = "vacía"
)

// Encabezados de las tablas de informe.md (contrato job-de-evals §5).
var (
	encabezadosDeFueraDeLoGrabado = []string{"Sesión", "Eval", "Orden", "Código"}
	encabezadosDeRed              = []string{"Sesión", "Eval", "Orden", "Destino"}
	encabezadosDeSesiones         = []string{
		"Sesión", "Eval", "Activa", "Activada", "Sesión terminada", "Comandos ausentes", "Citas ausentes", "Resultado",
	}
	encabezadosDeInvocaciones = []string{"Orden", "Código", "Conexiones"}
)

// enUnaLinea deja un texto en su línea de informe.md: cada salto de línea —\r\n,
// \r o \n, los tres finales de línea de CommonMark— se sustituye por un espacio.
// \r\n va antes que \r para que cuente como un solo salto.
var enUnaLinea = strings.NewReplacer("\r\n", " ", "\r", " ", "\n", " ")

// escapeDeCelda deja el texto de una celda en su celda y en su línea: además de
// los saltos de línea, la barra, que la cerraría, se escribe \|.
var escapeDeCelda = strings.NewReplacer("\r\n", " ", "\r", " ", "\n", " ", "|", `\|`)

// InformeAEscribir es lo que EscribirInforme necesita para escribir el informe de
// una ejecución del job de evals (contrato job-de-evals §3.3).
type InformeAEscribir struct {
	// Skill es la skill evaluada: la activación que se busca y el campo skill del
	// informe.
	Skill string

	// Evals es el directorio de las evals con las que se juzga: el job lo deriva
	// de la skill; TestInforme pasa el sintético.
	Evals string

	// Sesiones es el directorio con un subdirectorio por sesión.
	Sesiones string

	// Destino es el directorio en el que escribe informe.md e informe.json.
	Destino string

	// Modelo es el modelo fijado en el job.
	Modelo string

	// Commit es el commit evaluado.
	Commit string

	// SinPython es la ruta de sin-python.txt: su contenido entero, byte a byte,
	// es el sin_python del informe.
	SinPython string
}

// Informe es el informe de una ejecución del job de evals (data-model §10.3;
// contrato job-de-evals §5). Sus claves JSON, en el orden de sus campos, son las
// de informe.json.
type Informe struct {
	// Skill, Modelo y Commit son los recibidos, tal cual: el modelo es el fijado
	// en el job, no el de las sesiones.
	Skill  string `json:"skill"`
	Modelo string `json:"modelo"`

	// ModelosDeSesion y VersionesDeClaudeCode son el modelo y la versión de
	// Claude Code que declara cada sesión que LeerSesion leyó, sin repetir y en
	// el orden en que aparece cada uno por primera vez, recorriendo las sesiones
	// por nombre.
	ModelosDeSesion       []string `json:"modelos_de_sesion"`
	VersionesDeClaudeCode []string `json:"versiones_de_claude_code"`

	Commit string `json:"commit"`

	// SinPython es el contenido de sin-python.txt, igual byte a byte: la
	// constancia de cómo se comprobó que no había Python (FR-081).
	SinPython string `json:"sin_python"`

	// FicherosMalFormados son los que LeerConjunto no leyó como eval, en orden de
	// fichero.
	FicherosMalFormados []FicheroMalFormadoDelInforme `json:"ficheros_mal_formados"`

	Veredicto Veredicto `json:"veredicto"`

	// Motivos son las causas del veredicto fallo, en el orden de data-model
	// §10.3; vacío con el veredicto aprobado.
	Motivos []string `json:"motivos"`

	// FueraDeLoGrabado son las invocaciones fuera de lo grabado de todas las
	// sesiones, y Red, sus llegadas a la red, cada una con su sesión y su eval.
	FueraDeLoGrabado []FueraDeLoGrabadoDelInforme `json:"fuera_de_lo_grabado"`
	Red              []RedDelInforme              `json:"red"`

	// Evals son los resultados, uno por sesión y en orden de sesión.
	Evals []ResultadoDeEval `json:"evals"`
}

// FicheroMalFormadoDelInforme es un fichero del directorio de evals que no es una
// eval, con su error.
type FicheroMalFormadoDelInforme struct {
	Fichero string `json:"fichero"`
	Error   string `json:"error"`
}

// FueraDeLoGrabadoDelInforme es una invocación fuera de lo grabado con la sesión
// y la eval de su resultado.
type FueraDeLoGrabadoDelInforme struct {
	Sesion string `json:"sesion"`
	Eval   string `json:"eval"`
	Orden  string `json:"orden"`
	Codigo int    `json:"codigo"`
}

// RedDelInforme es una llegada a la red con la sesión y la eval de su resultado.
type RedDelInforme struct {
	Sesion  string `json:"sesion"`
	Eval    string `json:"eval"`
	Orden   string `json:"orden"`
	Destino string `json:"destino"`
}

// EscribirInforme juzga una ejecución del job de evals y escribe su informe.md y
// su informe.json en e.Destino, con permisos 0o600 y un salto de línea final,
// antes de devolver el Informe (contrato job-de-evals §3.3 y §5; data-model
// §10.3; FR-071, FR-073, FR-076):
//
//  1. lee sin-python.txt y las evals con LeerConjunto: los ficheros mal formados
//     van al informe y las sesiones se juzgan con las bien formadas, sin las
//     reglas del conjunto, que aplica antes el guion;
//  2. por cada entrada de e.Sesiones, en orden de nombre, lee siempre eval.txt,
//     pregunta.txt y la sesión con LeerSesion y, si LeerSesion la leyó, su traza
//     con LeerTrazas y el corte de la sesión. Solo si todo se leyó y eval.txt
//     nombra una eval bien formada, la juzga con Juzgar; si no, la sesión no
//     pasa, con un motivo «sesión ilegible: <fichero>: <error>» por cada fichero
//     que falta o no se puede leer, o por el eval.txt que no nombra ninguna eval,
//     en el orden eval.txt, pregunta.txt, sesión y traza. Una entrada que no es
//     un directorio no se salta: sus ficheros no se pueden leer. Lo que sí se
//     leyó de una sesión sin juzgar —la eval que nombra, lo observado de la
//     sesión y lo que hicieron sus invocaciones— se informa igual, para que
//     ninguna llegada a la red quede sin detectar (FR-076);
//  3. los motivos de la raíz van en el orden de data-model §10.3: los de cada
//     eval que no pasa, precedidos de su sesión; cada eval bien formada que
//     ningún eval.txt nombra; ninguna eval bien formada que juzgar; cada fichero
//     mal formado; y cada petición llegada a la red. El veredicto es fallo por
//     cualquiera de esas causas y aprobado sin ninguna.
//
// El error es solo para lo que impide escribir el informe —sin-python.txt, las
// evals o las sesiones que no se pueden leer, o un destino en el que no se puede
// escribir— y nombra el fichero o el directorio. Todo se lee antes de escribir
// nada y, con cualquiera de esos errores, en el destino no queda ni informe.md ni
// informe.json, de modo que un sin_python vacío o a medias no llega nunca a un
// informe (FR-081).
func EscribirInforme(e InformeAEscribir) (Informe, error) {
	sinPython, err := leerFichero(e.SinPython)
	if err != nil {
		return Informe{}, fmt.Errorf("el informe no se puede escribir: la comprobación sin Python %s no se puede leer: %w",
			e.SinPython, err)
	}

	conjunto, err := LeerConjunto(e.Evals)
	if err != nil {
		return Informe{}, fmt.Errorf("el informe no se puede escribir: %w", err)
	}

	sesiones, err := juzgarSesiones(e, conjunto.Evals)
	if err != nil {
		return Informe{}, fmt.Errorf("el informe no se puede escribir: %w", err)
	}

	informe := componerInforme(e, string(sinPython), conjunto, sesiones)

	// Un argv de la traza puede traer cualquier octeto: el que no es UTF-8 válido
	// se escribe como el carácter de sustitución, en lugar de dejar la ejecución
	// sin informe.
	codificado, err := json.Marshal(informe, jsontext.WithIndent("  "), jsontext.AllowInvalidUTF8(true))
	if err != nil {
		return Informe{}, fmt.Errorf("el informe no se puede escribir: %s no se puede codificar: %w", ficheroDelInformeJSON, err)
	}

	if err := escribirElInforme(e.Destino, renderizarInforme(informe, sesiones), append(codificado, '\n')); err != nil {
		return Informe{}, err
	}

	return informe, nil
}

// sesionJuzgada es lo que el informe lleva de una sesión: su resultado y, para su
// sección de informe.md, lo leído de ella.
type sesionJuzgada struct {
	resultado ResultadoDeEval

	// ilegible dice si la sesión quedó sin juzgar porque algo no se pudo leer.
	ilegible bool

	// sesion es lo que leyó LeerSesion, con las invocaciones de su traza si se
	// leyó; leida dice si LeerSesion la leyó, y trazaLeida, si LeerTrazas leyó su
	// traza.
	sesion     Sesion
	leida      bool
	trazaLeida bool

	// pregunta es el contenido de pregunta.txt, y preguntaLeida dice si se leyó.
	pregunta      string
	preguntaLeida bool
}

// juzgarSesiones juzga cada entrada del directorio de sesiones, en orden de
// nombre. El error queda para el directorio que no se puede listar.
func juzgarSesiones(e InformeAEscribir, evals []Eval) ([]sesionJuzgada, error) {
	entradas, err := os.ReadDir(e.Sesiones)
	if err != nil {
		return nil, fmt.Errorf("el directorio de sesiones %s no se puede listar: %w", e.Sesiones, err)
	}

	sesiones := make([]sesionJuzgada, 0, len(entradas))
	for _, entrada := range entradas {
		sesiones = append(sesiones, juzgarSesion(e, evals, entrada.Name()))
	}

	return sesiones, nil
}

// juzgarSesion lee la sesión del subdirectorio nombre y la juzga con la eval que
// nombra su eval.txt, o la deja sin pasar con un motivo por cada fichero que no
// se pudo leer (paso 2 de EscribirInforme).
func juzgarSesion(e InformeAEscribir, evals []Eval, nombre string) sesionJuzgada {
	dir := filepath.Join(e.Sesiones, nombre)

	var (
		juzgada sesionJuzgada
		motivos []string
	)

	nombreDeEval, eval, errDeEval := leerEvalDeLaSesion(dir, e.Evals, evals)
	if errDeEval != nil {
		motivos = append(motivos, motivoDeSesionIlegible+errDeEval.Error())
	}

	pregunta, err := leerFicheroDeSesion(dir, ficheroDeLaPregunta)
	if err != nil {
		motivos = append(motivos, motivoDeSesionIlegible+err.Error())
	} else {
		juzgada.pregunta, juzgada.preguntaLeida = string(pregunta), true
	}

	juzgada.sesion, err = LeerSesion(dir)
	if err != nil {
		motivos = append(motivos, motivoDeSesionIlegible+err.Error())
	} else {
		juzgada.leida = true

		// Sin el código de la sesión no se sabría si el tope la cortó: la traza
		// solo se lee de una sesión que LeerSesion leyó.
		juzgada.sesion.Invocaciones, err = LeerTrazas(filepath.Join(dir, directorioDeLaTraza), juzgada.sesion.Cortada)
		if err != nil {
			motivos = append(motivos, motivoDeSesionIlegible+directorioDeLaTraza+": "+err.Error())
		} else {
			juzgada.trazaLeida = true
		}
	}

	if len(motivos) == 0 {
		juzgada.resultado = Juzgar(eval, juzgada.sesion, e.Skill)
		juzgada.resultado.Sesion = nombre

		return juzgada
	}

	juzgada.ilegible = true
	juzgada.resultado = ResultadoDeEval{Sesion: nombre, Eval: nombreDeEval, Motivos: motivos}

	if errDeEval == nil {
		juzgada.resultado.Activa = eval.Activa
	}

	if juzgada.leida {
		codigo := juzgada.sesion.Codigo
		juzgada.resultado.Activada = juzgada.sesion.Activada(e.Skill)
		juzgada.resultado.Respuesta = juzgada.sesion.Respuesta
		juzgada.resultado.CodigoDeLaSesion = &codigo
		juzgada.resultado.FinDeLaSesion = juzgada.sesion.Fin
		juzgada.resultado.SesionTerminada = juzgada.sesion.Terminada
	}

	for _, invocacion := range juzgada.sesion.Invocaciones {
		juzgada.resultado.informar(invocacion)
	}

	return juzgada
}

// leerEvalDeLaSesion lee de eval.txt el nombre de la eval con la que se juzga la
// sesión, sin su salto de línea final, y la busca entre las evals bien formadas
// del directorio de evals. El nombre es vacío si eval.txt no se leyó, y el error,
// que empieza por eval.txt como el de cualquier fichero de la sesión, nombra
// también el que no es ninguna eval bien formada.
func leerEvalDeLaSesion(dir, directorioDeEvals string, evals []Eval) (string, Eval, error) {
	contenido, err := leerFicheroDeSesion(dir, ficheroDeLaEval)
	if err != nil {
		return "", Eval{}, err
	}

	nombre := strings.TrimSuffix(string(contenido), "\n")

	posicion := slices.IndexFunc(evals, func(eval Eval) bool { return eval.Fichero == nombre })
	if posicion < 0 {
		return nombre, Eval{}, fmt.Errorf("%s: %s nombra %q, que no es ninguna eval bien formada de %s",
			ficheroDeLaEval, filepath.Join(dir, ficheroDeLaEval), nombre, directorioDeEvals)
	}

	return nombre, evals[posicion], nil
}

// componerInforme reúne en el informe lo recibido, lo leído y los resultados de
// las sesiones, y decide sus motivos y su veredicto.
func componerInforme(e InformeAEscribir, sinPython string, conjunto Conjunto, sesiones []sesionJuzgada) Informe {
	informe := Informe{Skill: e.Skill, Modelo: e.Modelo, Commit: e.Commit, SinPython: sinPython}

	for _, malFormado := range conjunto.MalFormados {
		informe.FicherosMalFormados = append(informe.FicherosMalFormados,
			FicheroMalFormadoDelInforme{Fichero: malFormado.Fichero, Error: malFormado.Error.Error()})
	}

	for _, juzgada := range sesiones {
		resultado := juzgada.resultado
		informe.Evals = append(informe.Evals, resultado)

		if juzgada.leida {
			informe.ModelosDeSesion = agregarSinRepetir(informe.ModelosDeSesion, juzgada.sesion.Modelo)
			informe.VersionesDeClaudeCode = agregarSinRepetir(informe.VersionesDeClaudeCode,
				juzgada.sesion.VersionDeClaudeCode)
		}

		for _, fuera := range resultado.FueraDeLoGrabado {
			informe.FueraDeLoGrabado = append(informe.FueraDeLoGrabado, FueraDeLoGrabadoDelInforme{
				Sesion: resultado.Sesion, Eval: resultado.Eval, Orden: fuera.Orden, Codigo: fuera.Codigo,
			})
		}

		for _, llegada := range resultado.LlegadasALaRed {
			informe.Red = append(informe.Red, RedDelInforme{
				Sesion: resultado.Sesion, Eval: resultado.Eval, Orden: llegada.Orden, Destino: llegada.Destino,
			})
		}
	}

	informe.Motivos = motivosDelInforme(informe, conjunto.Evals)
	informe.Veredicto = veredictoDelInforme(informe, conjunto.Evals)

	return informe
}

// agregarSinRepetir añade el valor a la lista si no está ya. Un valor vacío no se
// añade: es el de una sesión cuyo transcript no llegó a declararlo.
func agregarSinRepetir(lista []string, valor string) []string {
	if valor == "" || slices.Contains(lista, valor) {
		return lista
	}

	return append(lista, valor)
}

// motivosDelInforme son los motivos de la raíz del informe en el orden de
// data-model §10.3.
func motivosDelInforme(informe Informe, evals []Eval) []string {
	var motivos []string

	for _, resultado := range informe.Evals {
		if resultado.Pasa {
			continue
		}

		for _, motivo := range resultado.Motivos {
			motivos = append(motivos, resultado.Sesion+": "+motivo)
		}
	}

	for _, eval := range evals {
		if !nombradaPorAlgunaSesion(informe.Evals, eval.Fichero) {
			motivos = append(motivos, eval.Fichero+motivoSinNingunaSesion)
		}
	}

	if len(evals) == 0 {
		motivos = append(motivos, motivoSinEvalsQueJuzgar)
	}

	for _, malFormado := range informe.FicherosMalFormados {
		motivos = append(motivos, malFormado.Fichero+motivoDeFicheroMalFormado+malFormado.Error)
	}

	for _, llegada := range informe.Red {
		motivos = append(motivos, llegada.Sesion+motivoDeLlegadaALaRed+llegada.Orden+" → "+llegada.Destino)
	}

	return motivos
}

// veredictoDelInforme es fallo si hay ficheros mal formados, si alguna eval no
// pasa —también por una sesión sin terminar o ilegible—, si alguna eval bien
// formada no tiene ninguna sesión que la juzgue, si no hay ninguna eval bien
// formada o si alguna petición llegó a la red; aprobado en otro caso. Así no
// aprueba un informe que no evaluó todas las evals (data-model §10.3).
func veredictoDelInforme(informe Informe, evals []Eval) Veredicto {
	algunaNoPasa := slices.ContainsFunc(informe.Evals, func(resultado ResultadoDeEval) bool { return !resultado.Pasa })
	algunaSinSesion := slices.ContainsFunc(evals, func(eval Eval) bool {
		return !nombradaPorAlgunaSesion(informe.Evals, eval.Fichero)
	})

	if len(informe.FicherosMalFormados) > 0 || algunaNoPasa || algunaSinSesion || len(evals) == 0 || len(informe.Red) > 0 {
		return VeredictoFallo
	}

	return VeredictoAprobado
}

// nombradaPorAlgunaSesion dice si el eval.txt de alguna sesión nombra la eval,
// aunque la sesión no se pudiera leer entera.
func nombradaPorAlgunaSesion(resultados []ResultadoDeEval, fichero string) bool {
	return slices.ContainsFunc(resultados, func(resultado ResultadoDeEval) bool { return resultado.Eval == fichero })
}

// escribirElInforme escribe informe.md y después informe.json en el destino. Si
// el segundo no se puede escribir, retira el primero: ningún informe queda a
// medias. El error nombra el destino y lleva el del fichero.
func escribirElInforme(destino string, md, codificado []byte) error {
	rutaDelMD := filepath.Join(destino, ficheroDelInformeMD)

	if err := escribirFichero(rutaDelMD, md); err != nil {
		return fmt.Errorf("el informe no se puede escribir en %s: %w", destino, err)
	}

	if err := escribirFichero(filepath.Join(destino, ficheroDelInformeJSON), codificado); err != nil {
		return errors.Join(fmt.Errorf("el informe no se puede escribir en %s: %w", destino, err), retirarFichero(rutaDelMD))
	}

	return nil
}

// retirarFichero borra un fichero ya escrito.
func retirarFichero(ruta string) error {
	if err := os.Remove(ruta); err != nil {
		return fmt.Errorf("%s, ya escrito, no se puede retirar: %w", ruta, err)
	}

	return nil
}

// renderizarInforme da informe.md (contrato job-de-evals §5): el título; el
// veredicto y sus motivos; la cabecera con sus cuatro líneas; la comprobación sin
// Python en un bloque; los ficheros mal formados, las invocaciones fuera de lo
// grabado y las peticiones llegadas a la red; la tabla de las sesiones; y una
// sección por sesión.
func renderizarInforme(informe Informe, sesiones []sesionJuzgada) []byte {
	var md documento

	md.parrafo("# Informe de evals de " + informe.Skill)

	md.parrafo("## Veredicto")
	md.parrafo("Veredicto: " + string(informe.Veredicto))
	md.listaConEtiqueta("Motivos", informe.Motivos)

	md.parrafo("## Cabecera")
	md.parrafo("Modelo del job: " + informe.Modelo)
	md.parrafo("Modelos de las sesiones: " + unidosOVacio(informe.ModelosDeSesion, ningunoEnElInforme))
	md.parrafo("Versiones de Claude Code: " + unidosOVacio(informe.VersionesDeClaudeCode, ningunaEnElInforme))
	md.parrafo("Commit: " + informe.Commit)

	md.parrafo("## Comprobación sin Python")
	md.bloqueDeTexto(informe.SinPython)

	md.parrafo("## Ficheros mal formados")

	if len(informe.FicherosMalFormados) == 0 {
		md.parrafo(ningunoEnElInforme)
	} else {
		lineas := make([]string, 0, len(informe.FicherosMalFormados))
		for _, malFormado := range informe.FicherosMalFormados {
			lineas = append(lineas, malFormado.Fichero+": "+malFormado.Error)
		}

		md.lista(lineas)
	}

	md.parrafo("## Invocaciones fuera de lo grabado")
	md.tablaOVacia(encabezadosDeFueraDeLoGrabado, filasDeFueraDeLoGrabado(informe.FueraDeLoGrabado), ningunaEnElInforme)

	md.parrafo("## Peticiones llegadas a la red")
	md.tablaOVacia(encabezadosDeRed, filasDeRed(informe.Red), sinLlegadasALaRed)

	md.parrafo("## Sesiones")
	md.tablaOVacia(encabezadosDeSesiones, filasDeSesiones(informe.Evals), ningunaEnElInforme)

	for _, juzgada := range sesiones {
		md.sesion(juzgada)
	}

	return md.unido()
}

// sesion añade la sección de una sesión: su eval, la pregunta, las invocaciones
// con su código y sus conexiones, la respuesta y, si la sesión no terminó o no se
// pudo leer, sus motivos y su salida de error.
func (d *documento) sesion(juzgada sesionJuzgada) {
	resultado := juzgada.resultado

	d.parrafo("## Sesión " + resultado.Sesion)
	d.parrafo("Eval: " + cmp.Or(resultado.Eval, ningunaEnElInforme))
	d.textoLeido("Pregunta", juzgada.pregunta, juzgada.preguntaLeida)

	switch {
	case !juzgada.trazaLeida:
		d.parrafo("Invocaciones: " + sinLeer)
	case len(resultado.Invocaciones) == 0:
		d.parrafo("Invocaciones: " + ningunaEnElInforme)
	default:
		d.parrafo("Invocaciones:")
		d.tablaOVacia(encabezadosDeInvocaciones, filasDeInvocaciones(resultado.Invocaciones), ningunaEnElInforme)
	}

	d.textoLeido("Respuesta", resultado.Respuesta, juzgada.leida)

	if resultado.SesionTerminada && !juzgada.ilegible {
		return
	}

	var motivos []string

	for _, motivo := range resultado.Motivos {
		if strings.HasPrefix(motivo, motivoDeSesionIlegible) || strings.HasPrefix(motivo, motivoDeSesionSinTerminar) {
			motivos = append(motivos, motivo)
		}
	}

	d.listaConEtiqueta("Motivos de la sesión", motivos)
	d.textoLeido("Salida de error", juzgada.sesion.SalidaDeError, juzgada.leida)
}

// filasDeFueraDeLoGrabado son las filas de la tabla de invocaciones fuera de lo
// grabado.
func filasDeFueraDeLoGrabado(fuera []FueraDeLoGrabadoDelInforme) [][]string {
	filas := make([][]string, 0, len(fuera))
	for _, invocacion := range fuera {
		filas = append(filas, []string{invocacion.Sesion, invocacion.Eval, invocacion.Orden, strconv.Itoa(invocacion.Codigo)})
	}

	return filas
}

// filasDeRed son las filas de la tabla de peticiones llegadas a la red.
func filasDeRed(red []RedDelInforme) [][]string {
	filas := make([][]string, 0, len(red))
	for _, llegada := range red {
		filas = append(filas, []string{llegada.Sesion, llegada.Eval, llegada.Orden, llegada.Destino})
	}

	return filas
}

// filasDeSesiones son las filas de la tabla de las sesiones: sesión, eval, activa,
// activada, sesión terminada con su código, comandos ausentes, citas ausentes y
// resultado.
func filasDeSesiones(resultados []ResultadoDeEval) [][]string {
	filas := make([][]string, 0, len(resultados))

	for _, resultado := range resultados {
		codigo := "sin código"
		if resultado.CodigoDeLaSesion != nil {
			codigo = "código " + strconv.Itoa(*resultado.CodigoDeLaSesion)
		}

		pasa := "no pasa"
		if resultado.Pasa {
			pasa = "pasa"
		}

		filas = append(filas, []string{
			resultado.Sesion,
			resultado.Eval,
			siONo(resultado.Activa),
			siONo(resultado.Activada),
			siONo(resultado.SesionTerminada) + " (" + codigo + ")",
			unidosOVacio(resultado.ComandosAusentes, ningunoEnElInforme),
			unidosOVacio(resultado.CitasAusentes, ningunaEnElInforme),
			pasa,
		})
	}

	return filas
}

// filasDeInvocaciones son las filas de la tabla de invocaciones de una sesión:
// la orden, su código o «sin código (sesión cortada)» y sus conexiones, cada una
// con su destino y su clase, o «sin conexiones».
func filasDeInvocaciones(invocaciones []InvocacionInformada) [][]string {
	filas := make([][]string, 0, len(invocaciones))

	for _, invocacion := range invocaciones {
		codigo := sinCodigoPorElCorte
		if invocacion.Codigo != nil {
			codigo = strconv.Itoa(*invocacion.Codigo)
		}

		conexiones := make([]string, 0, len(invocacion.Conexiones))
		for _, conexion := range invocacion.Conexiones {
			conexiones = append(conexiones, conexion.Destino+" ("+string(conexion.Clase)+")")
		}

		filas = append(filas, []string{invocacion.Orden, codigo, unidosOVacio(conexiones, sinConexiones)})
	}

	return filas
}

// siONo es «sí» o «no».
func siONo(valor bool) string {
	if valor {
		return "sí"
	}

	return "no"
}

// unidosOVacio son los elementos separados por «, », o el texto de la lista
// vacía.
func unidosOVacio(elementos []string, vacia string) string {
	return cmp.Or(strings.Join(elementos, ", "), vacia)
}

// documento son los bloques de informe.md —párrafo o título, lista, tabla o
// bloque de texto—, cada uno terminado en un salto de línea. Unidos, cada dos
// quedan separados por una línea en blanco.
type documento []string

// parrafo añade una línea de texto, que puede ser un título, con sus saltos de
// línea como espacios.
func (d *documento) parrafo(texto string) {
	*d = append(*d, enUnaLinea.Replace(texto)+"\n")
}

// lista añade una lista con un elemento por línea.
func (d *documento) lista(elementos []string) {
	var lista strings.Builder
	for _, elemento := range elementos {
		lista.WriteString("- " + enUnaLinea.Replace(elemento) + "\n")
	}

	*d = append(*d, lista.String())
}

// listaConEtiqueta añade «<etiqueta>: ninguno» o, con elementos, la etiqueta y
// una lista con uno por línea.
func (d *documento) listaConEtiqueta(etiqueta string, elementos []string) {
	if len(elementos) == 0 {
		d.parrafo(etiqueta + ": " + ningunoEnElInforme)

		return
	}

	d.parrafo(etiqueta + ":")
	d.lista(elementos)
}

// tablaOVacia añade una tabla con sus encabezados y una fila por elemento, con cada
// celda escapada, o el texto de la tabla vacía.
func (d *documento) tablaOVacia(encabezados []string, filas [][]string, vacia string) {
	if len(filas) == 0 {
		d.parrafo(vacia)

		return
	}

	var tabla strings.Builder

	for _, fila := range slices.Concat([][]string{encabezados, slices.Repeat([]string{"---"}, len(encabezados))}, filas) {
		celdas := make([]string, 0, len(fila))
		for _, celda := range fila {
			celdas = append(celdas, escapeDeCelda.Replace(celda))
		}

		tabla.WriteString("| " + strings.Join(celdas, " | ") + " |\n")
	}

	*d = append(*d, tabla.String())
}

// bloqueDeTexto añade el contenido tal cual entre una línea ```text y una línea
// ```. Si el contenido tiene una racha de tres o más comillas invertidas, la
// valla tiene una más que la más larga, para que ninguna línea del contenido
// cierre el bloque; y si no termina en un salto de línea, se le añade uno antes
// de la valla de cierre.
func (d *documento) bloqueDeTexto(contenido string) {
	valla := strings.Repeat("`", max(len("```"), rachaDeComillasMasLarga(contenido)+1))

	if !strings.HasSuffix(contenido, "\n") {
		contenido += "\n"
	}

	*d = append(*d, valla+"text\n"+contenido+valla+"\n")
}

// textoLeido añade «<etiqueta>: sin leer» si el texto no se leyó, «<etiqueta>:
// vacía» si está vacío y, si no, la etiqueta y el texto en un bloque.
func (d *documento) textoLeido(etiqueta, texto string, leido bool) {
	switch {
	case !leido:
		d.parrafo(etiqueta + ": " + sinLeer)
	case texto == "":
		d.parrafo(etiqueta + ": " + vaciaEnElInforme)
	default:
		d.parrafo(etiqueta + ":")
		d.bloqueDeTexto(texto)
	}
}

// unido es informe.md entero: los bloques separados por una línea en blanco, con
// el salto de línea final del último.
func (d documento) unido() []byte {
	return []byte(strings.Join(d, "\n"))
}

// rachaDeComillasMasLarga es la longitud de la racha más larga de comillas
// invertidas seguidas del texto.
func rachaDeComillasMasLarga(texto string) int {
	mayor, racha := 0, 0

	for _, caracter := range []byte(texto) {
		if caracter != '`' {
			racha = 0

			continue
		}

		racha++
		mayor = max(mayor, racha)
	}

	return mayor
}
