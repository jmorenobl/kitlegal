package evals

import (
	"encoding/json"
	"regexp"
	"slices"
	"strings"
)

// Las piezas de lo que se reconoce de una sentencia en un texto
// (contracts/evals-jurisprudencia.md §3 de H23; FR-005 y FR-061 de H23).
const (
	// patronDeECLI es el de un ECLI del país que sea: ECLI y, separados por dos
	// puntos, el país y el órgano, de letras ASCII o cifras, el año, de cuatro
	// cifras, y el número, de letras ASCII, cifras o puntos, que llega hasta el
	// primer carácter que no lo es y no termina en punto: los puntos en que
	// termine son la puntuación que cierra la frase. No distingue mayúsculas.
	patronDeECLI = `(?i:ECLI:[A-Z0-9]+:[A-Z0-9]+:[0-9]{4}:[A-Z0-9.]*[A-Z0-9])`

	// patronDeROJ es el de un ROJ: las siglas, en una o más palabras de letras
	// mayúsculas ASCII separadas por un solo espacio, el número y el año de
	// cuatro cifras.
	patronDeROJ = `[A-Z]+(?: [A-Z]+)* [0-9]+/[0-9]{4}`

	// etiquetaNoComprobada es la de la forma fija de la línea con la que una
	// respuesta dice que no ha comprobado una sentencia.
	etiquetaNoComprobada = "SENTENCIA NO COMPROBADA"

	// fuenteDeLaCita es la fuente con la que firma sus sobres el applet cita.
	fuenteDeLaCita = "kitlegal.cita"
)

// Las expresiones con las que se reconoce en un texto lo que el juicio de
// sentencias compara por su forma, sin ningún modelo.
var (
	// formaDeCitaDeSentencia casa con la cita de una sentencia: un corchete,
	// abierto y cerrado en la misma línea, con el ECLI, una coma y un espacio,
	// ROJ:, un espacio y el ROJ, y nada más. Otra escritura no es una cita.
	formaDeCitaDeSentencia = regexp.MustCompile(`\[(` + patronDeECLI + `), ROJ: (` + patronDeROJ + `)\]`)

	// formaDeECLI casa con cada ECLI de un texto.
	formaDeECLI = regexp.MustCompile(patronDeECLI)

	// lineaNoComprobadaDeSentencia casa con el principio de cada línea que
	// empieza por la forma fija de etiquetaNoComprobada, con la tolerancia de
	// las etiquetas de los avisos (patronDeEtiqueta): blancos y énfasis de
	// Markdown delante de la marca y alrededor de las partes.
	lineaNoComprobadaDeSentencia = regexp.MustCompile(`(?m)^` + separadorDeAviso +
		patronDeEtiqueta(etiquetaNoComprobada))

	// lineaDeAviso casa con cada línea que, tras esos mismos blancos y énfasis,
	// empieza por la marca de los avisos, entera.
	lineaDeAviso = regexp.MustCompile(`(?m)^` + separadorDeAviso + marcaDeAviso + `[^\n]*`)
)

// CitaDeSentencia es la cita de una sentencia en la respuesta de una sesión: el
// ECLI y el ROJ de su corchete, como están escritos. Se compara con una
// CitaDeSentenciaEsperada por igualdad exacta de la pareja
// (contracts/evals-jurisprudencia.md §2 de H23; FR-051).
type CitaDeSentencia struct {
	// ECLI es el ECLI citado, del país que sea.
	ECLI string

	// ROJ es el ROJ citado.
	ROJ string
}

// texto es la cita como la nombra un motivo: su corchete.
func (c CitaDeSentencia) texto() string {
	return "[" + c.ECLI + ", ROJ: " + c.ROJ + "]"
}

// ExtraerCitasDeSentencia devuelve las citas de sentencia del texto, cada una
// la de un corchete con la forma [<ECLI>, ROJ: <ROJ>], en el orden en que
// aparecen y con sus repeticiones, o nil si no hay ninguna. Lo que el texto
// escribe delante del corchete no se lee (contracts/evals-jurisprudencia.md §3
// de H23; FR-051).
func ExtraerCitasDeSentencia(texto string) []CitaDeSentencia {
	var citas []CitaDeSentencia

	for _, partes := range formaDeCitaDeSentencia.FindAllStringSubmatch(texto, -1) {
		citas = append(citas, CitaDeSentencia{ECLI: partes[1], ROJ: partes[2]})
	}

	return citas
}

// ExtraerNoComprobada dice si el texto tiene una línea que empieza por la forma
// fija ⚠ SENTENCIA NO COMPROBADA: —la marca, la etiqueta y los dos puntos, con
// la tolerancia de las etiquetas de los avisos—. Lo que sigue a los dos puntos
// no se lee (contracts/evals-jurisprudencia.md §3 de H23; FR-051).
func ExtraerNoComprobada(texto string) bool {
	return lineaNoComprobadaDeSentencia.MatchString(texto)
}

// ExtraerLineasDeAviso devuelve las líneas del texto que empiezan por la marca
// de los avisos, con la tolerancia de sus etiquetas, enteras y en su orden, o
// nil si no hay ninguna (contracts/evals-jurisprudencia.md §3 de H23; FR-061).
func ExtraerLineasDeAviso(texto string) []string {
	return lineaDeAviso.FindAllString(texto, -1)
}

// ExtraerECLI devuelve los ECLI del texto, del país que sean, como están
// escritos, en el orden en que aparecen y con sus repeticiones, o nil si no hay
// ninguno. Cada uno llega hasta el primer carácter que no es suyo, sin los
// puntos en que termine: en «su ECLI es ECLI:ES:TS:2023:3144.» es
// ECLI:ES:TS:2023:3144. La regla es la misma en una respuesta, en una pregunta
// y en la salida de una operación (contracts/evals-jurisprudencia.md §3 de H23;
// FR-061).
func ExtraerECLI(texto string) []string {
	return formaDeECLI.FindAllString(texto, -1)
}

// MismoECLI dice si dos ECLI son el mismo: se comparan sin distinguir
// mayúsculas de minúsculas (FR-061 de H23).
func MismoECLI(uno, otro string) bool {
	return strings.EqualFold(uno, otro)
}

// SobreDeSesion es un sobre de kitlegal en la salida de una operación de una
// sesión (contracts/evals-jurisprudencia.md §3 de H23).
type SobreDeSesion struct {
	// Linea es la línea de la salida que es el sobre, sin los blancos de
	// alrededor.
	Linea string

	// OK y Fuente son su ok y su fuente.
	OK     bool
	Fuente string

	// Data es su data, tal como está en la línea.
	Data json.RawMessage
}

// ExtraerSobres devuelve los sobres de la sesión, en el orden de sus
// operaciones y, dentro de cada una, en el de sus líneas, o nil si no hay
// ninguno: cada línea de la salida de una orden de Bash que nombra kitlegal, o
// de una llamada a una herramienta del registro, que es un objeto JSON con las
// seis claves del sobre, terminara la operación como terminara. Es lo que la
// sesión guarda de cada una en Textos. Lo que la orden escribe en otras líneas
// —lo que una orden encadenada lee de otro sitio— no es la salida de una
// operación de kitlegal, y una orden sin --json no da ningún sobre
// (contracts/evals-jurisprudencia.md §3 de H23; research D19 de H23; FR-061).
func ExtraerSobres(sesion Sesion) []SobreDeSesion {
	var sobres []SobreDeSesion

	for _, texto := range sesion.Textos {
		for linea := range strings.SplitSeq(texto.Salida, "\n") {
			if sobre, esSobre := sobreDeLaLinea(strings.TrimSpace(linea)); esSobre {
				sobres = append(sobres, sobre)
			}
		}
	}

	return sobres
}

// sobreDeLaLinea lee la línea como un sobre, y dice si lo es: un objeto JSON con las
// seis claves del sobre, con ok booleano y fuente de texto.
func sobreDeLaLinea(linea string) (SobreDeSesion, bool) {
	var campos struct {
		OK            *bool           `json:"ok"`
		Fuente        *string         `json:"fuente"`
		URL           json.RawMessage `json:"url"`
		FechaConsulta json.RawMessage `json:"fecha_consulta"`
		Hash          json.RawMessage `json:"hash"`
		Data          json.RawMessage `json:"data"`
	}

	// La línea que no empieza por una llave no es un objeto: no hace falta
	// analizarla para saberlo, y casi ninguna lo es.
	if !strings.HasPrefix(linea, "{") || json.Unmarshal([]byte(linea), &campos) != nil {
		return SobreDeSesion{}, false
	}

	if campos.OK == nil || campos.Fuente == nil || campos.URL == nil || campos.FechaConsulta == nil ||
		campos.Hash == nil || campos.Data == nil {
		return SobreDeSesion{}, false
	}

	return SobreDeSesion{Linea: linea, OK: *campos.OK, Fuente: *campos.Fuente, Data: campos.Data}, true
}

// sobresDeCitaConExito son los sobres de la sesión con ok verdadero que firma
// el applet cita: los de sus dos verbos cuando terminan bien.
func sobresDeCitaConExito(sesion Sesion) []SobreDeSesion {
	return slices.DeleteFunc(ExtraerSobres(sesion), func(sobre SobreDeSesion) bool {
		return !sobre.OK || sobre.Fuente != fuenteDeLaCita
	})
}

// ExtraerECLICotejados devuelve los ECLI que leyó un cita cotejar de la sesión
// en la ficha de un documento, en su orden y sin repetir, o nil si no leyó
// ninguno: data.ficha.ecli de cada sobre con ok verdadero que firma el applet
// cita. El ECLI que la salida repite porque se le dio como referencia pedida,
// el de data.pedida, no lo ha leído en ningún documento
// (contracts/evals-jurisprudencia.md §3 de H23; FR-061).
func ExtraerECLICotejados(sesion Sesion) []string {
	var cotejados []string

	for _, sobre := range sobresDeCitaConExito(sesion) {
		var data struct {
			Ficha struct {
				ECLI string `json:"ecli"`
			} `json:"ficha"`
		}

		if json.Unmarshal(sobre.Data, &data) != nil || data.Ficha.ECLI == "" {
			continue
		}

		if !slices.ContainsFunc(cotejados, func(otro string) bool { return MismoECLI(otro, data.Ficha.ECLI) }) {
			cotejados = append(cotejados, data.Ficha.ECLI)
		}
	}

	return cotejados
}

// ExtraerDireccionesDeBusqueda devuelve las direcciones de búsqueda por texto
// que devolvió un cita preparar de la sesión, en su orden y sin repetir, o nil
// si no devolvió ninguna: data.direccion de cada sobre con ok verdadero que
// firma el applet cita y que lleva data.texto. La dirección de la consulta de
// una referencia no es la de una búsqueda
// (contracts/evals-jurisprudencia.md §3 de H23).
func ExtraerDireccionesDeBusqueda(sesion Sesion) []string {
	var direcciones []string

	for _, sobre := range sobresDeCitaConExito(sesion) {
		var data struct {
			Texto     *string `json:"texto"`
			Direccion string  `json:"direccion"`
		}

		if json.Unmarshal(sobre.Data, &data) != nil || data.Texto == nil || data.Direccion == "" {
			continue
		}

		if !slices.Contains(direcciones, data.Direccion) {
			direcciones = append(direcciones, data.Direccion)
		}
	}

	return direcciones
}

// Principio de los motivos por los que una eval no pasa por lo que su
// respuesta lleva, o no lleva, de las sentencias de las que habla, y los dos
// de texto fijo (contracts/evals-jurisprudencia.md §2 de H23). Los de una
// cita van seguidos de su corchete; el de una dirección, de la dirección; y
// los de una casilla los componen motivoDeCasillaAusente y
// motivoDeValorAusente.
const (
	motivoDeCitaDeSentenciaAusente = "falta la cita de sentencia "
	motivoDeCitaDeSentencia        = "la respuesta cita una sentencia: "
	motivoDeCitaConElROJ           = "la respuesta cita con el ROJ "
	motivoDeLineaNoComprobada      = "falta la línea ⚠ " + etiquetaNoComprobada + ":"
	motivoDeDireccionAusente       = "falta la dirección "
	motivoDeCasillaAusente         = "falta la casilla "
	motivoDeValorAusente           = "falta el valor "
	motivoSinDireccionDeBusqueda   = "ninguna orden devolvió una dirección de búsqueda"
	motivoDeDireccionDeBusqueda    = "falta la dirección de búsqueda "
)

// juzgarLasSentencias juzga lo que la eval espera de la respuesta sobre las
// sentencias de las que habla (contracts/evals-jurisprudencia.md §2 de H23;
// FR-051): anota un motivo por cada cosa que falta o que sobra, en el orden de
// motivosDeLasSentencias, y dice si la respuesta lo cumple todo. El resultado
// no gana claves: lo que falta va en sus motivos. Una eval que no declara
// sentencias no juzga nada y cumple.
func (r *ResultadoDeEval) juzgarLasSentencias(esperadas SentenciasEsperadas, sesion Sesion) bool {
	motivos := motivosDeLasSentencias(esperadas, sesion)
	r.Motivos = append(r.Motivos, motivos...)

	return len(motivos) == 0
}

// motivosDeLasSentencias son los motivos por los que la respuesta de la sesión
// no cumple lo que la eval espera de las sentencias, en el orden de la tabla
// del contrato: cada cita esperada que falta; cada cita, si no puede haber
// ninguna; cada cita con un ROJ que no se puede citar; la línea que falta; cada
// dirección que falta; cada casilla sin su nombre o sin su valor; y la
// dirección de búsqueda que no devolvió ninguna orden o que falta. Nada de
// ello usa un modelo ni cambia de un modo a otro.
func motivosDeLasSentencias(esperadas SentenciasEsperadas, sesion Sesion) []string {
	respuesta := sesion.Respuesta
	motivos := motivosDeLasCitasDeSentencia(esperadas, ExtraerCitasDeSentencia(respuesta))

	if esperadas.NoComprobada && !ExtraerNoComprobada(respuesta) {
		motivos = append(motivos, motivoDeLineaNoComprobada)
	}

	for _, direccion := range esperadas.Direcciones {
		if !strings.Contains(respuesta, direccion) {
			motivos = append(motivos, motivoDeDireccionAusente+direccion)
		}
	}

	motivos = append(motivos, motivosDeLasCasillas(esperadas.Casillas, respuesta)...)

	if esperadas.DireccionDeBusqueda {
		motivos = append(motivos, motivosDeLaDireccionDeBusqueda(ExtraerDireccionesDeBusqueda(sesion), respuesta)...)
	}

	return motivos
}

// motivosDeLasCitasDeSentencia son los de las tres claves que miran las citas
// de la respuesta: el de cada cita esperada que no está entre ellas, con el
// ECLI y el ROJ iguales carácter a carácter; si no puede haber ninguna, uno por
// cita, con sus repeticiones; y, por cada ROJ que no se puede citar, uno por
// cita que lo lleva.
func motivosDeLasCitasDeSentencia(esperadas SentenciasEsperadas, citas []CitaDeSentencia) []string {
	var motivos []string

	for _, esperada := range esperadas.Citas {
		if !slices.Contains(citas, CitaDeSentencia(esperada)) {
			motivos = append(motivos, motivoDeCitaDeSentenciaAusente+CitaDeSentencia(esperada).texto())
		}
	}

	if esperadas.NingunaCita {
		for _, cita := range citas {
			motivos = append(motivos, motivoDeCitaDeSentencia+cita.texto())
		}
	}

	for _, roj := range esperadas.SinCitaDelROJ {
		for _, cita := range citas {
			if cita.ROJ == roj {
				motivos = append(motivos, motivoDeCitaConElROJ+roj+": "+cita.texto())
			}
		}
	}

	return motivos
}

// motivosDeLasCasillas son los de cada casilla cuyo nombre o cuyo valor no
// está en la respuesta tal cual: el del nombre y, detrás, el del valor.
func motivosDeLasCasillas(casillas []CasillaEsperada, respuesta string) []string {
	var motivos []string

	for _, casilla := range casillas {
		if !strings.Contains(respuesta, casilla.Nombre) {
			motivos = append(motivos, motivoDeCasillaAusente+casilla.Nombre)
		}

		if !strings.Contains(respuesta, casilla.Valor) {
			motivos = append(motivos, motivoDeValorAusente+casilla.Valor+" de la casilla "+casilla.Nombre)
		}
	}

	return motivos
}

// motivosDeLaDireccionDeBusqueda es el motivo, si lo hay, de la dirección de
// búsqueda: el de la sesión en la que ningún cita preparar con texto devolvió
// una, o el de la respuesta que no lleva tal cual ninguna de las que devolvió,
// que las nombra todas, separadas por « o »: con varias búsquedas en una
// sesión, vale la dirección de cualquiera.
func motivosDeLaDireccionDeBusqueda(devueltas []string, respuesta string) []string {
	if len(devueltas) == 0 {
		return []string{motivoSinDireccionDeBusqueda}
	}

	if slices.ContainsFunc(devueltas, func(direccion string) bool { return strings.Contains(respuesta, direccion) }) {
		return nil
	}

	return []string{motivoDeDireccionDeBusqueda + strings.Join(devueltas, " o ")}
}

// pideLaCita dice si la invocación, que ya consultó con éxito y es del applet
// del comando, satisface un comando de preparar o de cotejar
// (contracts/evals-jurisprudencia.md §2 de H23): es de su verbo; si el comando
// lleva ROJ, su --roj vale eso; y si es con texto, lleva --texto con un valor
// no vacío. Vale igual para la orden y para la llamada a una herramienta, cuyas
// banderas llegan como las de la orden equivalente.
func pideLaCita(invocacion Invocacion, comando ComandoEsperado) bool {
	return invocacion.Verbo == comando.Verbo &&
		(comando.ROJ == "" || llevaElROJ(invocacion.Argumentos, comando.ROJ)) &&
		(!comando.ConTexto || llevaTexto(invocacion.Argumentos))
}

// llevaElROJ dice si los argumentos de la invocación dan ese ROJ con --roj.
func llevaElROJ(argumentos []string, roj string) bool {
	dado, conROJ := valorDeLaBandera(argumentos, banderaDelROJ)

	return conROJ && dado == roj
}

// llevaTexto dice si los argumentos de la invocación dan con --texto un valor
// no vacío: el de una búsqueda por texto, sea cual sea.
func llevaTexto(argumentos []string) bool {
	texto, _ := valorDeLaBandera(argumentos, banderaDelTexto)

	return texto != ""
}

// Las dos banderas del applet cita que mira el juicio, sin sus guiones.
const (
	banderaDelROJ   = "roj"
	banderaDelTexto = "texto"
)

// prefijoDeBandera es el de una bandera en los argumentos de una invocación.
const prefijoDeBandera = "--"

// valorDeLaBandera es lo que vale en los argumentos de una invocación del
// applet cita la bandera con ese nombre, y si se ha dado, se escriba
// --<nombre> <valor>, en dos argumentos, o --<nombre>=<valor>, en uno; si se
// escribe más de una vez vale la última, como para el binario. Todas las
// banderas propias de los verbos de cita llevan valor, e InterpretarInvocacion
// ya quitó las globales: el argumento que sigue a una bandera sin su valor
// pegado es su valor, no otra bandera ni un argumento de posición. Lo que va
// detrás de un «--» suelto son argumentos de posición.
func valorDeLaBandera(argumentos []string, nombre string) (valor string, dada bool) {
	for posicion := 0; posicion < len(argumentos) && argumentos[posicion] != terminadorDeLaLlamada; posicion++ {
		escrita, pegado, conIgual := strings.Cut(argumentos[posicion], "=")
		if !strings.HasPrefix(escrita, prefijoDeBandera) {
			continue
		}

		suValor := pegado
		if !conIgual {
			if posicion+1 == len(argumentos) {
				break
			}

			posicion++
			suValor = argumentos[posicion]
		}

		if escrita == prefijoDeBandera+nombre {
			valor, dada = suValor, true
		}
	}

	return valor, dada
}

// textoDelComandoDeCita es el comando de preparar o de cotejar con el texto
// con el que lo presentan el informe y los motivos: <applet> <verbo>, con
// --roj <ROJ> detrás si lo lleva y, detrás, --texto si es con texto.
func textoDelComandoDeCita(comando ComandoEsperado) string {
	partes := []string{comando.Applet, comando.Verbo}

	if comando.ROJ != "" {
		partes = append(partes, prefijoDeBandera+banderaDelROJ, comando.ROJ)
	}

	if comando.ConTexto {
		partes = append(partes, prefijoDeBandera+banderaDelTexto)
	}

	return strings.Join(partes, " ")
}
