package evals

import (
	"fmt"
	"path"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/jmorenobl/kitlegal/internal/core/territorio"
)

// Los dos códigos de salida estables con los que una invocación con consulta dice
// que no tuvo lo que pedía: 4, fuente no disponible, el de --offline sin la
// consulta en la caché, y 5, límite o términos de uso, el de la petición que el
// proxy del job rechaza (CLAUDE.md, «Exit codes estables»; contrato job-de-evals
// §6). Una invocación que termina con uno de ellos está fuera de lo grabado
// (data-model §10.2).
const (
	codigoFuenteNoDisponible = 4
	codigoLimiteOTos         = 5
)

// verboArticulos es el otro verbo, junto a articulo, con el que se lee un bloque
// esperado: pide varios bloques de una norma (data-model §6.1).
const verboArticulos = "articulos"

// Principio de los motivos por los que una eval no pasa (data-model §10.2). El de
// la sesión sin terminar lo fija el contrato job-de-evals §5; los de un comando,
// una cita, un aviso o un elemento del territorio ausentes van seguidos de su
// texto, el mismo con el que los presenta el informe, que en un aviso es su código
// (contrato de formato, juicio e informe §4 de H5.1; contrato de evals §2 de H6).
const (
	motivoDeSesionSinTerminar = "la sesión no terminó: "
	motivoDeComandoAusente    = "comando ausente: "
	motivoDeCitaAusente       = "cita ausente: "
	motivoDeAvisoAusente      = "aviso ausente: "
	motivoDeTerritorioAusente = "territorio ausente: "
	motivoDeOtroModelo        = "la sesión no declara el modelo que se le pidió: "
)

// ResultadoDeEval es el juicio de una sesión con su eval (data-model §10.2): lo
// esperado y lo observado, el reparto de los comandos, las citas, los avisos y el
// territorio esperados, lo que hicieron las invocaciones de la sesión, por qué no
// pasa y si pasa. Sus claves JSON son las de cada eval de informe.json (contrato
// job-de-evals §5).
type ResultadoDeEval struct {
	// Sesion es el nombre del directorio de la sesión. Juzgar no lo conoce: lo
	// pone EscribirInforme.
	Sesion string `json:"sesion"`

	// Eval es el nombre del fichero de la eval con la que se juzga la sesión.
	Eval string `json:"eval"`

	// Modelo es el id del modelo con el que se pidió abrir la sesión, el de su
	// modelo.txt, y ModeloDeLaSesion, el que la propia sesión declara en su
	// transcript. Los dos los pone EscribirInforme: Juzgar no los conoce.
	Modelo           string `json:"modelo"`
	ModeloDeLaSesion string `json:"modelo_de_la_sesion"`

	// Decide dice si la sesión pertenece a una serie que decide el veredicto; lo
	// pone EscribirInforme al repartir las sesiones en series (data-model §10.4).
	Decide bool `json:"decide"`

	// Activa dice si la eval espera que la skill se active, y Activada, si la
	// sesión la activó: la activación coincide si son iguales.
	Activa   bool `json:"activa"`
	Activada bool `json:"activada"`

	// ComandosEjecutados y ComandosAusentes reparten los comandos esperados, en
	// el orden de la eval, entre los que satisface alguna invocación de la sesión
	// y los que no (data-model §6.1), cada uno con su texto: bloque <applet>
	// <norma> <bloque>, <applet> <verbo> <norma>, <applet> buscar <términos…> o
	// <applet> resolver <municipio>.
	ComandosEjecutados []string `json:"comandos_ejecutados"`
	ComandosAusentes   []string `json:"comandos_ausentes"`

	// CitasEncontradas y CitasAusentes reparten las citas esperadas, en el orden
	// de la eval, entre las que están en la respuesta y las que no (data-model
	// §6.2), cada una con su texto: <norma> <bloque>.
	CitasEncontradas []string `json:"citas_encontradas"`
	CitasAusentes    []string `json:"citas_ausentes"`

	// AvisosEncontrados y AvisosAusentes reparten los avisos esperados, en el orden de la eval y con sus
	// repeticiones, entre los que la respuesta lleva con su forma fija y los que no (ExtraerAvisos), cada uno con su
	// código.
	AvisosEncontrados []string `json:"avisos_encontrados"`
	AvisosAusentes    []string `json:"avisos_ausentes"`

	// TerritorioEncontrado y TerritorioAusente reparten los elementos del
	// territorio esperado, en el orden de la eval —comunidad, provincia, cada
	// boletín y cada aspecto de cobertura— y con sus repeticiones, entre los que la
	// respuesta declara con su forma fija y los que no (ExtraerTerritorio), cada
	// uno con su texto: comunidad: <nombre>, provincia: <nombre>, boletín:
	// <código> o <aspecto>: <valor>.
	TerritorioEncontrado []string `json:"territorio_encontrado"`
	TerritorioAusente    []string `json:"territorio_ausente"`

	// Invocaciones son todas las invocaciones de applet de la sesión, en su
	// orden.
	Invocaciones []InvocacionInformada `json:"invocaciones"`

	// FueraDeLoGrabado son las invocaciones con consulta y código 4 o 5, en su
	// orden.
	FueraDeLoGrabado []InvocacionFallida `json:"fuera_de_lo_grabado"`

	// OtrasFallidas son las invocaciones con consulta y otro código distinto de
	// 0, en su orden; no las que quedaron sin código porque el tope cortó la
	// sesión antes de que acabaran.
	OtrasFallidas []InvocacionFallida `json:"otras_fallidas"`

	// LlegadasALaRed tienen una entrada por cada invocación y destino de sus
	// conexiones de clase red, en su orden.
	LlegadasALaRed []LlegadaALaRed `json:"llegadas_a_la_red"`

	// Respuesta es la de la sesión: vacía si no la hay o si la sesión no se pudo
	// leer.
	Respuesta string `json:"respuesta"`

	// CodigoDeLaSesion es el código de la sesión. Es nil solo si la sesión no se
	// pudo leer, porque el código de una sesión nunca es 0 por omisión.
	CodigoDeLaSesion *int `json:"codigo_de_la_sesion"`

	// FinDeLaSesion es el fin de la sesión con su texto fijo (data-model §10.1):
	// vacío si la sesión no se pudo leer.
	FinDeLaSesion string `json:"fin_de_la_sesion"`

	// SesionTerminada dice si la sesión terminó.
	SesionTerminada bool `json:"sesion_terminada"`

	// Motivos son las causas por las que la eval no pasa, una por causa y en este
	// orden: la sesión ilegible, que pone EscribirInforme, o sin terminar; la
	// activación que no coincide; cada comando ausente; cada cita ausente; cada
	// aviso ausente; cada elemento del territorio ausente; y el modelo que la
	// sesión declara sin ser el pedido, que pone EscribirInforme. Vacío si pasa.
	Motivos []string `json:"motivos"`

	// Pasa dice si la sesión terminó, la activación coincide y no falta ningún
	// comando, ninguna cita, ningún aviso ni ningún elemento del territorio
	// esperados. No lo cambian FueraDeLoGrabado, OtrasFallidas ni LlegadasALaRed
	// (FR-076), ni la forma fija de un aviso que la eval no espera.
	Pasa bool `json:"pasa"`
}

// InvocacionInformada es una invocación de applet de la sesión como la presenta
// el informe (data-model §10.2).
type InvocacionInformada struct {
	// Orden es el applet seguido de los argumentos que le siguen en argv,
	// separados por un espacio.
	Orden string `json:"orden"`

	// Codigo es el de la invocación, o nil en la que quedó sin código porque el
	// tope cortó la sesión (data-model §9).
	Codigo *int `json:"codigo"`

	// Conexiones tiene una entrada por pareja distinta de destino y clase de las
	// conexiones de la invocación, en el orden en que aparece por primera vez;
	// vacía si la invocación no conectó.
	Conexiones []ConexionInformada `json:"conexiones"`
}

// ConexionInformada es una pareja de destino y clase de las conexiones de una
// invocación (data-model §9).
type ConexionInformada struct {
	// Destino es el de Conexion.Destino.
	Destino string `json:"destino"`

	// Clase es local, bloqueada o red.
	Clase ClaseDeConexion `json:"clase"`
}

// InvocacionFallida es una invocación con consulta que terminó con un código
// distinto de 0, con su orden y su código.
type InvocacionFallida struct {
	Orden  string `json:"orden"`
	Codigo int    `json:"codigo"`
}

// LlegadaALaRed es una invocación con una conexión de clase red, con su orden y
// el destino de esa conexión.
type LlegadaALaRed struct {
	Orden   string `json:"orden"`
	Destino string `json:"destino"`
}

// Juzgar compara una sesión con su eval sin ningún modelo (data-model §10.2;
// contrato evals-y-grabaciones §6; FR-072): si la sesión terminó, si la
// activación de la skill coincide con la esperada, qué comandos esperados
// satisfacen sus invocaciones (data-model §6.1), qué citas esperadas están en su
// respuesta (data-model §6.2) y qué avisos esperados lleva su respuesta con su
// forma fija, la marca, la etiqueta y los dos puntos que reconoce ExtraerAvisos
// (FR-030 a FR-033 de H5.1). Informa además de todas sus invocaciones, de las
// que quedaron fuera de lo grabado, de las otras fallidas y de las que llegaron a
// la red, sin que nada de eso cambie si la eval pasa (FR-076).
//
// Una sesión sin terminar no pasa aunque todo lo demás coincida: sin ella, una
// eval de no activación cuya sesión murió sin activar nada pasaría en vacío. Una
// eval sin avisos deja vacíos los encontrados y los ausentes, y su juicio es el de
// antes de H5.1 (FR-034).
//
// Desde H6, reparte además los elementos del territorio esperado entre los que la
// respuesta declara con su forma fija (ExtraerTerritorio) y los ausentes, y un
// ausente impide pasar; una eval sin territorio esperado deja vacíos los dos y su
// juicio es el de antes (contrato de evals §2 de H6; FR-084).
func Juzgar(eval Eval, sesion Sesion, skill string) ResultadoDeEval {
	codigo := sesion.Codigo

	resultado := ResultadoDeEval{
		Eval:             eval.Fichero,
		Activa:           eval.Activa,
		Activada:         sesion.Activada(skill),
		Respuesta:        sesion.Respuesta,
		CodigoDeLaSesion: &codigo,
		FinDeLaSesion:    sesion.Fin,
		SesionTerminada:  sesion.Terminada,
	}

	if !sesion.Terminada {
		resultado.Motivos = append(resultado.Motivos, motivoDeSesionSinTerminar+sesion.MotivoSinTerminar)
	}

	if resultado.Activa != resultado.Activada {
		resultado.Motivos = append(resultado.Motivos, motivoDeActivacion(skill, eval.Activa))
	}

	resultado.repartirComandos(eval.Comandos, sesion.Invocaciones)
	resultado.repartirCitas(eval.Citas, ExtraerCitas(sesion.Respuesta))
	resultado.repartirAvisos(eval.Avisos, ExtraerAvisos(sesion.Respuesta))
	resultado.repartirTerritorio(eval.Territorio, ExtraerTerritorio(sesion.Respuesta, eval.Territorio))

	for _, invocacion := range sesion.Invocaciones {
		resultado.informar(invocacion)
	}

	resultado.Pasa = sesion.Terminada && resultado.Activa == resultado.Activada &&
		len(resultado.ComandosAusentes) == 0 && len(resultado.CitasAusentes) == 0 &&
		len(resultado.AvisosAusentes) == 0 && len(resultado.TerritorioAusente) == 0

	return resultado
}

// exigirElModeloPedido deja de pasar, con su motivo, la sesión que declara un
// modelo que no es el que se le pidió: si no, el informe publicaría como medida
// de un modelo lo que hizo otro (contrato job-de-evals §4; ADR 0016). El motivo va
// detrás de los de Juzgar, porque no es un defecto de la skill sino de la
// ejecución. El id declarado puede llevar detrás la fecha de la versión, porque el
// proveedor resuelve el alias que se pidió, así que basta con que empiece por el
// pedido; una sesión que no llegó a declarar ninguno no tiene modelo que comparar
// y ya no terminó.
func (r *ResultadoDeEval) exigirElModeloPedido() {
	if r.ModeloDeLaSesion == "" || strings.HasPrefix(r.ModeloDeLaSesion, r.Modelo) {
		return
	}

	r.Motivos = append(r.Motivos, motivoDeOtroModelo+r.ModeloDeLaSesion+", y se pidió "+r.Modelo)
	r.Pasa = false
}

// motivoDeActivacion es el motivo de una activación que no coincide con la que la
// eval espera.
func motivoDeActivacion(skill string, activa bool) string {
	if activa {
		return fmt.Sprintf("la activación no coincide: se esperaba que la skill %s se activara y no se activó", skill)
	}

	return fmt.Sprintf("la activación no coincide: se esperaba que la skill %s no se activara y se activó", skill)
}

// repartirComandos reparte los comandos esperados entre ejecutados y ausentes,
// con un motivo por cada ausente.
func (r *ResultadoDeEval) repartirComandos(comandos []ComandoEsperado, invocaciones []Invocacion) {
	for _, comando := range comandos {
		texto := textoDelComando(comando)

		if slices.ContainsFunc(invocaciones, func(invocacion Invocacion) bool { return satisface(invocacion, comando) }) {
			r.ComandosEjecutados = append(r.ComandosEjecutados, texto)

			continue
		}

		r.ComandosAusentes = append(r.ComandosAusentes, texto)
		r.Motivos = append(r.Motivos, motivoDeComandoAusente+texto)
	}
}

// repartirCitas reparte las citas esperadas entre encontradas y ausentes según
// las citas de la respuesta, con un motivo por cada ausente.
func (r *ResultadoDeEval) repartirCitas(esperadas []CitaEsperada, citas []Cita) {
	for _, esperada := range esperadas {
		texto := esperada.Norma + " " + esperada.Bloque

		if slices.Contains(citas, Cita(esperada)) {
			r.CitasEncontradas = append(r.CitasEncontradas, texto)

			continue
		}

		r.CitasAusentes = append(r.CitasAusentes, texto)
		r.Motivos = append(r.Motivos, motivoDeCitaAusente+texto)
	}
}

// repartirAvisos reparte los avisos esperados entre encontrados y ausentes según
// los avisos cuya forma fija lleva la respuesta, con un motivo por cada ausente.
func (r *ResultadoDeEval) repartirAvisos(esperados, avisos []string) {
	for _, esperado := range esperados {
		if slices.Contains(avisos, esperado) {
			r.AvisosEncontrados = append(r.AvisosEncontrados, esperado)

			continue
		}

		r.AvisosAusentes = append(r.AvisosAusentes, esperado)
		r.Motivos = append(r.Motivos, motivoDeAvisoAusente+esperado)
	}
}

// repartirTerritorio reparte los elementos del territorio esperado entre
// encontrados y ausentes según lo que de él declara la respuesta, con un motivo
// por cada ausente.
func (r *ResultadoDeEval) repartirTerritorio(esperado, declarado TerritorioEsperado) {
	declarados := declarado.elementos()

	for _, elemento := range esperado.elementos() {
		if slices.Contains(declarados, elemento) {
			r.TerritorioEncontrado = append(r.TerritorioEncontrado, elemento)

			continue
		}

		r.TerritorioAusente = append(r.TerritorioAusente, elemento)
		r.Motivos = append(r.Motivos, motivoDeTerritorioAusente+elemento)
	}
}

// informar añade la invocación a las del resultado con sus parejas de destino y
// clase, una llegada a la red por cada destino de clase red y, si consultó y
// terminó con un código distinto de 0, la lleva a fuera de lo grabado con 4 o 5
// y a las otras fallidas con cualquier otro. La que no consultó (la ayuda,
// --describe o --dry-run) o quedó sin código no va a ninguna de las dos.
func (r *ResultadoDeEval) informar(invocacion Invocacion) {
	informada := InvocacionInformada{Orden: ordenDeLaInvocacion(invocacion), Codigo: copiaDelCodigo(invocacion.Codigo)}

	for _, conexion := range invocacion.Conexiones {
		pareja := ConexionInformada{Destino: conexion.Destino(), Clase: conexion.Clase}
		if slices.Contains(informada.Conexiones, pareja) {
			continue
		}

		informada.Conexiones = append(informada.Conexiones, pareja)

		if pareja.Clase == ConexionRed {
			r.LlegadasALaRed = append(r.LlegadasALaRed, LlegadaALaRed{Orden: informada.Orden, Destino: pareja.Destino})
		}
	}

	r.Invocaciones = append(r.Invocaciones, informada)

	if !invocacion.Consulta || invocacion.Codigo == nil || *invocacion.Codigo == 0 {
		return
	}

	fallida := InvocacionFallida{Orden: informada.Orden, Codigo: *invocacion.Codigo}

	if fallida.Codigo == codigoFuenteNoDisponible || fallida.Codigo == codigoLimiteOTos {
		r.FueraDeLoGrabado = append(r.FueraDeLoGrabado, fallida)

		return
	}

	r.OtrasFallidas = append(r.OtrasFallidas, fallida)
}

// satisface dice si la invocación satisface el comando esperado (data-model
// §6.1 de H5 y de H6): tiene que consultar, terminar con código 0 y ser del mismo
// applet; y, según la forma del comando, en la forma bloque, leer ese bloque de
// esa norma; en la consulta de norma, ser el mismo verbo con esa norma; en la
// búsqueda, ser buscar con cada término como palabra de sus argumentos; y en el
// comando de territorio, ser resolver con el municipio como argumento.
func satisface(invocacion Invocacion, comando ComandoEsperado) bool {
	if !consultoConExito(invocacion) || invocacion.Applet != comando.Applet {
		return false
	}

	var satisfecho bool

	switch formaDelComando(comando) {
	case formaBloque:
		satisfecho = leeElBloque(invocacion, comando.Norma, comando.Bloque)
	case formaConsultaDeNorma:
		satisfecho = invocacion.Verbo == comando.Verbo && esDeLaNorma(invocacion, comando.Norma)
	case formaBusqueda:
		satisfecho = invocacion.Verbo == verboBuscar && contieneLosTerminos(invocacion.Argumentos, comando.Terminos)
	case formaTerritorio:
		satisfecho = invocacion.Verbo == verboResolver && resuelveElMunicipio(invocacion.Argumentos, comando.Municipio)
	}

	return satisfecho
}

// consultoConExito dice si la invocación consultó y terminó con código 0: ni
// --describe ni --dry-run, que no leen nada, ni la que quedó sin código.
func consultoConExito(invocacion Invocacion) bool {
	return invocacion.Consulta && invocacion.Codigo != nil && *invocacion.Codigo == 0
}

// leeElBloque dice si la invocación es articulo o articulos de la norma con el
// bloque entre los que pide.
func leeElBloque(invocacion Invocacion, norma, bloque string) bool {
	return (invocacion.Verbo == verboArticulo || invocacion.Verbo == verboArticulos) &&
		esDeLaNorma(invocacion, norma) && slices.Contains(invocacion.Argumentos[1:], bloque)
}

// esDeLaNorma dice si la norma es el primer argumento de la invocación. Los
// verbos de una norma y los de sus bloques la reciben siempre en primer lugar y no
// tienen banderas propias (internal/app/boe.go), y InterpretarInvocacion ya quitó
// las globales.
func esDeLaNorma(invocacion Invocacion, norma string) bool {
	return len(invocacion.Argumentos) > 0 && invocacion.Argumentos[0] == norma
}

// resuelveElMunicipio dice si el único argumento de la invocación es el
// municipio, plegados los dos con territorio.Plegar, el mismo pliegue con el que
// el applet compara los nombres: resolver recibe un solo argumento y no tiene
// banderas propias (contrato del applet territorio §1), e InterpretarInvocacion ya
// quitó las globales. Resolverlo por su código INE no es resolver el municipio que
// la eval escribe.
func resuelveElMunicipio(argumentos []string, municipio string) bool {
	return len(argumentos) == 1 && territorio.Plegar(argumentos[0]) == territorio.Plegar(municipio)
}

// contieneLosTerminos dice si los argumentos, unidos por un espacio, contienen
// cada término como palabra, comparados los dos en minúsculas: así cuenta un
// término dentro de un argumento con espacios, como en buscar "bases del régimen
// local", y el que se escribe con otras mayúsculas.
func contieneLosTerminos(argumentos, terminos []string) bool {
	texto := strings.ToLower(strings.Join(argumentos, " "))

	for _, termino := range terminos {
		if !contieneComoPalabra(texto, strings.ToLower(termino)) {
			return false
		}
	}

	return true
}

// contieneComoPalabra dice si el texto contiene la palabra delimitada a los dos
// lados por el principio o el final del texto o por un carácter que no es letra
// ni cifra: «común» no está en «comúnmente» ni en «intercomún», y «7/1985» sí
// está en «ley 7/1985».
func contieneComoPalabra(texto, palabra string) bool {
	for desde := 0; desde <= len(texto)-len(palabra); {
		posicion := strings.Index(texto[desde:], palabra)
		if posicion < 0 {
			return false
		}

		inicio := desde + posicion
		anterior, _ := utf8.DecodeLastRuneInString(texto[:inicio])
		siguiente, _ := utf8.DecodeRuneInString(texto[inicio+len(palabra):])

		if !esDePalabra(anterior) && !esDePalabra(siguiente) {
			return true
		}

		desde = inicio + 1
	}

	return false
}

// esDePalabra dice si el carácter es letra o cifra. El que da el principio o el
// final del texto, utf8.RuneError, no lo es.
func esDePalabra(caracter rune) bool {
	return unicode.IsLetter(caracter) || unicode.IsNumber(caracter)
}

// ordenDeLaInvocacion es la orden de una invocación como la presenta el informe (data-model
// §10.2): el applet seguido de los argumentos que le siguen en argv, separados
// por un espacio. El applet es el nombre de invocación, como scripts/boe, o, si
// no lo es, el primer argumento de kitlegal (data-model §9).
func ordenDeLaInvocacion(invocacion Invocacion) string {
	argv := invocacion.Argv

	tokensDelApplet := 1
	if len(argv) > 0 && path.Base(argv[0]) != invocacion.Applet {
		tokensDelApplet = 2
	}

	return strings.Join(slices.Concat([]string{invocacion.Applet}, argv[min(tokensDelApplet, len(argv)):]), " ")
}

// textoDelComando es el comando esperado con el texto con el que lo presentan el
// informe y los motivos (contrato job-de-evals §5), según su forma: bloque
// <applet> <norma> <bloque> en la forma bloque, que satisfacen dos verbos;
// <applet> <verbo> <norma> en la consulta de norma; <applet> buscar <términos…> en
// la búsqueda; y <applet> resolver <municipio> en el comando de territorio.
func textoDelComando(comando ComandoEsperado) string {
	var partes []string

	switch formaDelComando(comando) {
	case formaBloque:
		partes = []string{"bloque", comando.Applet, comando.Norma, comando.Bloque}
	case formaConsultaDeNorma:
		partes = []string{comando.Applet, comando.Verbo, comando.Norma}
	case formaBusqueda:
		partes = slices.Concat([]string{comando.Applet, comando.Verbo}, comando.Terminos)
	case formaTerritorio:
		partes = []string{comando.Applet, comando.Verbo, comando.Municipio}
	}

	return strings.Join(partes, " ")
}

// copiaDelCodigo es una copia del código de una invocación, para que el resultado
// no comparta memoria con la sesión: nil si la invocación quedó sin código.
func copiaDelCodigo(codigo *int) *int {
	if codigo == nil {
		return nil
	}

	copia := *codigo

	return &copia
}
