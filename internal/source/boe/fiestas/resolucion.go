package fiestas

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/jmorenobl/kitlegal/internal/core/plazos"
	"github.com/jmorenobl/kitlegal/internal/core/territorio"
)

// Resolucion es el anexo de una resolución de fiestas laborales leído tal como
// está grabado (research D14; S1 de H9, corregido con lo grabado en
// gates/supuestos.md): una tabla con una fila por fecha y una columna por
// comunidad o ciudad autónoma, cuyas celdas llevan el código de la leyenda que
// sigue a la tabla —«*» y «**», fiestas nacionales; «***», fiestas de la
// comunidad autónoma—, y unas notas aclaratorias, de las que solo se leen las
// que dan fiestas de una parte de la comunidad que las llama (FR-025). La fila
// o la nota que no se lee sin interpretarla queda fuera, y se dice por qué
// (ADR 0039, punto 3).
type Resolucion struct {
	// Anio es el del título del anexo, «Año <aaaa>».
	Anio int
	// Cabecera son las comunidades de la tabla, una por columna y en su orden.
	Cabecera []Columna
	// Filas son las filas leídas, una por fecha y en el orden de la tabla.
	Filas []Fila
	// Partes son las fiestas que las notas leídas limitan a una parte de la
	// comunidad que las llama, en el orden de las notas.
	Partes []FiestaDeUnaParte
	// FilasSinLeer son las filas que quedan fuera.
	FilasSinLeer []FilaSinLeer
	// NotasSinLeer son las notas que quedan fuera.
	NotasSinLeer []NotaSinLeer
}

// Columna es una comunidad de la cabecera de la tabla.
type Columna struct {
	// Nombre es el que escribe la tabla, sin las llamadas a las notas.
	Nombre string
	// Notas son los números de las notas que llama, como el «(1)» de
	// «Canarias (1)», en su orden.
	Notas []int
}

// Fila es una fila del anexo que se lee: una fecha y el ámbito con que la marca
// cada comunidad.
type Fila struct {
	// Fecha es AAAA-MM-DD: el día de la primera celda, en el mes de la última
	// fila de mes y en el año del anexo.
	Fecha string
	// Texto es el de su primera celda, como «12 Fiesta Nacional de España.».
	Texto string
	// Ambitos son, uno por comunidad de la cabecera y en su orden, el ámbito de
	// su código, o vacío si su celda no lleva ninguno.
	Ambitos []plazos.Ambito
}

// FiestaDeUnaParte es una fiesta que una nota limita a una parte de la
// comunidad que la llama (FR-025).
type FiestaDeUnaParte struct {
	// Fecha es AAAA-MM-DD, en el año del anexo.
	Fecha string
	// Parte es la que nombra la nota, tal como la escribe.
	Parte string
	// Columna es la posición en la cabecera de la comunidad que llama a la nota.
	Columna int
	// Nota es su número.
	Nota int
}

// FilaSinLeer es una fila que queda fuera: su primera celda y el motivo.
type FilaSinLeer struct {
	Texto  string
	Motivo string
}

// NotaSinLeer es una nota que queda fuera entera: su número y el motivo.
type NotaSinLeer struct {
	Numero int
	Motivo string
}

// Lo que la lectura reconoce del documento del diario y de su anexo.
const (
	// principioDelTexto abre el texto de la disposición en la página del
	// diario; lo que lo precede —la cabecera, los guiones— no se lee.
	principioDelTexto = `<div id="textoxslt"`
	// claseDelTitulo es la del título del anexo, «Año <aaaa>».
	claseDelTitulo = "anexo_tit"
	// principioDeLaLeyenda y principioDeLasNotas son los párrafos que, tras la
	// tabla, abren la leyenda de los códigos y las notas aclaratorias.
	principioDeLaLeyenda = "Códigos de las fiestas:"
	principioDeLasNotas  = "Notas aclaratorias:"
	// metodoDeLaLectura es el único con el que se pide una resolución: GET.
	metodoDeLaLectura = "GET"
)

// Las formas fijas de lo que se lee. Una celda, una fila o una nota que no
// tiene la suya no se lee: no se aproxima.
var (
	tituloDelAnexo = regexp.MustCompile(`^Año ([0-9]{4})$`)
	// llamadaANota es una llamada al final de un nombre de la cabecera.
	llamadaANota = regexp.MustCompile(`\s*\(([0-9]{1,3})\)$`)
	// codigoDeLaLeyenda es un párrafo de la leyenda: «<significado> (<código>).».
	codigoDeLaLeyenda = regexp.MustCompile(`^(.+) \((\*+)\)\.$`)
	// principioDeNota es el del primer párrafo de una nota: «<número>. ».
	principioDeNota = regexp.MustCompile(`^([0-9]{1,3})\.\s+`)
	// diaDeLaFila es el principio de la primera celda de una fila de fecha:
	// el día y el nombre de la fiesta.
	diaDeLaFila = regexp.MustCompile(`^([0-9]{1,2})\s+\S`)
	// citaDeLaNota es lo que una nota transcribe de la disposición de la
	// comunidad, entre comillas angulares.
	citaDeLaNota = regexp.MustCompile(`«([^«»]*)»`)
	// fiestaDeUnaParte es una entrada de la relación de una cita: «en <parte>:
	// el <día> de <mes>», con lo que la describe detrás de una coma.
	fiestaDeUnaParte = regexp.MustCompile(`^en (.+?): el ([0-9]{1,2}) de (\pL+)(?:,\s.+)?\.?$`)
)

// leyendaConocida es la leyenda de los códigos de las tres resoluciones
// grabadas, con el ámbito de cada código (FR-020). Una resolución cuya leyenda
// no es exactamente esta no se lee: un código con otro significado daría a sus
// fiestas el ámbito equivocado.
var leyendaConocida = []struct {
	codigo, significado string
	ambito              plazos.Ambito
}{
	{"*", "Fiesta Nacional no sustituible", plazos.AmbitoNacional},
	{"**", "Fiesta Nacional respecto de la que no se ha ejercido la facultad de sustitución", plazos.AmbitoNacional},
	{"***", "Fiesta de Comunidad Autónoma", plazos.AmbitoAutonomico},
}

// nombresDeLosMeses son los meses plegados, de enero a diciembre.
var nombresDeLosMeses = []string{
	"enero", "febrero", "marzo", "abril", "mayo", "junio",
	"julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre",
}

// LeerResolucion lee el anexo de una resolución de fiestas laborales del
// documento del diario tal como lo graba grabar_datos
// (https://www.boe.es/diario_boe/txt.php?id=<id>): el año del título del anexo;
// la cabecera de la tabla, con las llamadas a las notas; cada fila, con el
// ámbito que la leyenda da al código de cada celda; y, de cada nota cuya cita
// es una relación de fiestas de una parte, esas fiestas. Lo que no tiene la
// forma de una resolución —sin texto, sin título del anexo, sin una tabla con
// cabecera y cuerpo, sin la leyenda conocida, con otra cosa tras la tabla— es
// un error: no hay nada que leer. La fila o la nota que no se lee sin
// interpretarla queda fuera, con su motivo (ADR 0039, punto 3).
func LeerResolucion(contenido []byte) (Resolucion, error) {
	texto, err := leerElTexto(contenido)
	if err != nil {
		return Resolucion{}, err
	}

	anio, err := anioDelAnexo(texto)
	if err != nil {
		return Resolucion{}, err
	}

	tabla, tras, err := tablaDelAnexo(texto)
	if err != nil {
		return Resolucion{}, err
	}

	cabecera, cuerpo, err := partesDeLaTabla(tabla)
	if err != nil {
		return Resolucion{}, err
	}

	ambitos, notas, err := leyendaYNotas(tras)
	if err != nil {
		return Resolucion{}, err
	}

	resolucion := Resolucion{Anio: anio, Cabecera: cabecera}
	resolucion.leerFilas(cuerpo, ambitos)
	resolucion.leerNotas(notas)

	return resolucion, nil
}

// nodo es un elemento del texto de la disposición, con su clase, o un trozo de
// su texto.
type nodo struct {
	// elemento es el nombre del elemento, vacío en un trozo de texto.
	elemento string
	clase    string
	// texto es el del trozo; vacío en un elemento.
	texto string
	hijos []*nodo
}

// leerElTexto lee el texto de la disposición, el elemento que abre
// principioDelTexto, como XHTML tolerante con las entidades de HTML: es lo que
// el BOE genera de la disposición, y lo único de la página que se lee.
func leerElTexto(contenido []byte) (*nodo, error) {
	inicio := bytes.Index(contenido, []byte(principioDelTexto))
	if inicio < 0 {
		return nil, fmt.Errorf("el documento no tiene el texto de una disposición (%s>)", principioDelTexto)
	}

	lector := xml.NewDecoder(bytes.NewReader(contenido[inicio:]))
	lector.Strict = false
	lector.AutoClose = xml.HTMLAutoClose
	lector.Entity = xml.HTMLEntity

	raiz := &nodo{}
	abiertos := []*nodo{raiz}

	for {
		simbolo, err := lector.Token()
		if err != nil {
			return nil, fmt.Errorf("el texto de la disposición no se lee entero: %w", err)
		}

		abierto := abiertos[len(abiertos)-1]

		switch leido := simbolo.(type) {
		case xml.StartElement:
			hijo := &nodo{elemento: leido.Name.Local, clase: claseDe(leido)}
			abierto.hijos = append(abierto.hijos, hijo)
			abiertos = append(abiertos, hijo)
		case xml.EndElement:
			abiertos = abiertos[:len(abiertos)-1]
			if len(abiertos) == 1 {
				return raiz.hijos[0], nil
			}
		case xml.CharData:
			abierto.hijos = append(abierto.hijos, &nodo{texto: string(leido)})
		}
	}
}

// claseDe es el atributo class de un elemento, o vacío.
func claseDe(elemento xml.StartElement) string {
	for _, atributo := range elemento.Attr {
		if atributo.Name.Local == "class" {
			return atributo.Value
		}
	}

	return ""
}

// elementos son los hijos de un nodo que son elementos, en su orden; con
// nombres, solo los que se llaman así.
func (n *nodo) elementos(nombres ...string) []*nodo {
	var elementos []*nodo

	for _, hijo := range n.hijos {
		if hijo.elemento != "" && (len(nombres) == 0 || slices.Contains(nombres, hijo.elemento)) {
			elementos = append(elementos, hijo)
		}
	}

	return elementos
}

// plano es el texto de un nodo y de todo lo que contiene, con los blancos
// —también el espacio duro— reducidos a un espacio entre palabras.
func (n *nodo) plano() string {
	var texto strings.Builder

	n.escribir(&texto)

	return strings.Join(strings.Fields(texto.String()), " ")
}

// escribir añade el texto de un nodo y el de lo que contiene.
func (n *nodo) escribir(texto *strings.Builder) {
	texto.WriteString(n.texto)

	for _, hijo := range n.hijos {
		hijo.escribir(texto)
	}
}

// anioDelAnexo es el año del único título del anexo del texto.
func anioDelAnexo(texto *nodo) (int, error) {
	var titulos []string

	for _, elemento := range texto.elementos("h5") {
		if elemento.clase == claseDelTitulo {
			titulos = append(titulos, elemento.plano())
		}
	}

	if len(titulos) != 1 {
		return 0, fmt.Errorf("el texto no tiene un título del anexo, «Año <aaaa>», sino %d", len(titulos))
	}

	anio := tituloDelAnexo.FindStringSubmatch(titulos[0])
	if anio == nil {
		return 0, fmt.Errorf("el título del anexo es «%s» y no «Año <aaaa>»", titulos[0])
	}

	return numeroDe(anio[1]), nil
}

// tablaDelAnexo es la única tabla del texto y el texto de cada párrafo que la
// sigue, que tienen que ser todos párrafos.
func tablaDelAnexo(texto *nodo) (*nodo, []string, error) {
	elementos := texto.elementos()

	var tablas []int

	for indice, elemento := range elementos {
		if elemento.elemento == "table" {
			tablas = append(tablas, indice)
		}
	}

	if len(tablas) != 1 {
		return nil, nil, fmt.Errorf("el texto no tiene una tabla del anexo, sino %d", len(tablas))
	}

	tras := make([]string, 0, len(elementos)-tablas[0]-1)

	for _, elemento := range elementos[tablas[0]+1:] {
		if elemento.elemento != "p" {
			return nil, nil, fmt.Errorf("tras la tabla del anexo viene un <%s>, y solo se leen párrafos", elemento.elemento)
		}

		tras = append(tras, elemento.plano())
	}

	return elementos[tablas[0]], tras, nil
}

// partesDeLaTabla son la cabecera de la tabla —las comunidades de la última
// fila de su <thead>— y las filas de su <tbody>.
func partesDeLaTabla(tabla *nodo) ([]Columna, []*nodo, error) {
	encabezados, cuerpos := tabla.elementos("thead"), tabla.elementos("tbody")
	if len(encabezados) != 1 || len(cuerpos) != 1 {
		return nil, nil, fmt.Errorf("la tabla del anexo no tiene una cabecera (<thead>) y un cuerpo (<tbody>), sino %d y %d",
			len(encabezados), len(cuerpos))
	}

	var cabecera []Columna

	if filas := encabezados[0].elementos("tr"); len(filas) > 0 {
		for _, celda := range filas[len(filas)-1].elementos("th", "td") {
			cabecera = append(cabecera, separarLlamadas(celda.plano()))
		}
	}

	if len(cabecera) == 0 {
		return nil, nil, errors.New("la cabecera de la tabla del anexo no nombra ninguna comunidad")
	}

	return cabecera, cuerpos[0].elementos("tr"), nil
}

// separarLlamadas separa el nombre de una comunidad de la cabecera de las
// llamadas a notas que lo siguen.
func separarLlamadas(texto string) Columna {
	var notas []int

	for {
		llamada := llamadaANota.FindStringSubmatchIndex(texto)
		if llamada == nil {
			return Columna{Nombre: texto, Notas: notas}
		}

		notas = append([]int{numeroDe(texto[llamada[2]:llamada[3]])}, notas...)
		texto = texto[:llamada[0]]
	}
}

// nota es una nota aclaratoria: su número y su texto, con los párrafos que la
// siguen hasta la siguiente.
type nota struct {
	numero int
	texto  string
}

// leyendaYNotas lee los párrafos que siguen a la tabla: la leyenda de los
// códigos, que tiene que ser la conocida, y, si las hay, las notas aclaratorias.
func leyendaYNotas(parrafos []string) (map[string]plazos.Ambito, []nota, error) {
	if len(parrafos) == 0 || parrafos[0] != principioDeLaLeyenda {
		return nil, nil, fmt.Errorf("tras la tabla del anexo no viene la leyenda de los códigos («%s»)", principioDeLaLeyenda)
	}

	leida := map[string]string{}
	siguiente := 1

	for ; siguiente < len(parrafos); siguiente++ {
		codigo := codigoDeLaLeyenda.FindStringSubmatch(parrafos[siguiente])
		if codigo == nil {
			break
		}

		if _, repetido := leida[codigo[2]]; repetido {
			return nil, nil, fmt.Errorf("la leyenda da el código «%s» dos veces", codigo[2])
		}

		leida[codigo[2]] = codigo[1]
	}

	ambitos, err := ambitosDeLaLeyenda(leida)
	if err != nil {
		return nil, nil, err
	}

	if siguiente == len(parrafos) {
		return ambitos, nil, nil
	}

	if parrafos[siguiente] != principioDeLasNotas {
		return nil, nil, fmt.Errorf("tras la leyenda de los códigos viene «%s», y no las notas («%s»)",
			parrafos[siguiente], principioDeLasNotas)
	}

	notas, err := notasDe(parrafos[siguiente+1:])

	return ambitos, notas, err
}

// ambitosDeLaLeyenda da el ámbito de cada código de una leyenda que es
// exactamente la conocida.
func ambitosDeLaLeyenda(leida map[string]string) (map[string]plazos.Ambito, error) {
	if len(leida) != len(leyendaConocida) {
		return nil, fmt.Errorf("la leyenda de los códigos da %d y la conocida, %d", len(leida), len(leyendaConocida))
	}

	ambitos := make(map[string]plazos.Ambito, len(leyendaConocida))

	for _, conocido := range leyendaConocida {
		if significado := leida[conocido.codigo]; significado != conocido.significado {
			return nil, fmt.Errorf("la leyenda da al código «%s» el significado «%s», y el conocido es «%s»",
				conocido.codigo, significado, conocido.significado)
		}

		ambitos[conocido.codigo] = conocido.ambito
	}

	return ambitos, nil
}

// notasDe junta los párrafos de las notas: el que empieza por «<número>. »
// abre una, y los demás siguen la anterior.
func notasDe(parrafos []string) ([]nota, error) {
	var notas []nota

	for _, parrafo := range parrafos {
		if numero := principioDeNota.FindStringSubmatch(parrafo); numero != nil {
			abierta := nota{numero: numeroDe(numero[1]), texto: parrafo}
			if slices.ContainsFunc(notas, func(n nota) bool { return n.numero == abierta.numero }) {
				return nil, fmt.Errorf("la nota %d está dos veces", abierta.numero)
			}

			notas = append(notas, abierta)

			continue
		}

		if len(notas) == 0 {
			return nil, fmt.Errorf("la primera nota no empieza por su número: «%s»", parrafo)
		}

		notas[len(notas)-1].texto += " " + parrafo
	}

	return notas, nil
}

// leerFilas lee las filas del cuerpo de la tabla. Una fila de mes da el mes de
// las de fecha que la siguen; una de fecha, el ámbito de cada comunidad.
func (r *Resolucion) leerFilas(cuerpo []*nodo, ambitos map[string]plazos.Ambito) {
	var mes time.Month

	for _, fila := range cuerpo {
		var celdas []string
		for _, celda := range fila.elementos("td", "th") {
			celdas = append(celdas, celda.plano())
		}

		if len(celdas) == 0 {
			r.dejarFuera("", "no tiene celdas")

			continue
		}

		if delMes, esMes := mesDe(celdas[0]); esMes {
			mes = delMes
			if slices.ContainsFunc(celdas[1:], func(celda string) bool { return celda != "" }) {
				r.dejarFuera(celdas[0], "es la fila de un mes y lleva marcas")
			}

			continue
		}

		leida, motivo := r.filaDeFecha(celdas, mes, ambitos)
		if motivo != "" {
			r.dejarFuera(celdas[0], motivo)

			continue
		}

		r.Filas = append(r.Filas, leida)
	}
}

// filaDeFecha lee una fila de fecha del mes, o da el motivo por el que no se
// lee.
func (r *Resolucion) filaDeFecha(celdas []string, mes time.Month, ambitos map[string]plazos.Ambito) (Fila, string) {
	dia := diaDeLaFila.FindStringSubmatch(celdas[0])
	if dia == nil {
		return Fila{}, "no empieza por el día ni es la fila de un mes"
	}

	if mes == 0 {
		return Fila{}, "no tiene antes la fila de su mes"
	}

	fecha, existe := fechaDe(r.Anio, mes, numeroDe(dia[1]))
	if !existe {
		return Fila{}, fmt.Sprintf("el día %s de %s de %d no existe", dia[1], nombresDeLosMeses[mes-1], r.Anio)
	}

	if len(celdas)-1 != len(r.Cabecera) {
		return Fila{}, fmt.Sprintf("tiene %d celdas de comunidad y la cabecera nombra %d comunidades",
			len(celdas)-1, len(r.Cabecera))
	}

	leida := Fila{Fecha: fecha, Texto: celdas[0], Ambitos: make([]plazos.Ambito, len(r.Cabecera))}
	marcada := false

	for indice, celda := range celdas[1:] {
		if celda == "" {
			continue
		}

		ambito, conocido := ambitos[celda]
		if !conocido {
			return Fila{}, fmt.Sprintf("la celda de «%s» dice «%s», que no es un código de la leyenda",
				r.Cabecera[indice].Nombre, celda)
		}

		leida.Ambitos[indice] = ambito
		marcada = true
	}

	if !marcada {
		return Fila{}, "ninguna comunidad la marca"
	}

	return leida, ""
}

// dejarFuera anota una fila que no se lee.
func (r *Resolucion) dejarFuera(texto, motivo string) {
	r.FilasSinLeer = append(r.FilasSinLeer, FilaSinLeer{Texto: texto, Motivo: motivo})
}

// leerNotas lee de cada nota las fiestas de una parte, o la deja fuera entera.
func (r *Resolucion) leerNotas(notas []nota) {
	for _, leida := range notas {
		partes, motivo := r.partesDeLaNota(leida)
		if motivo != "" {
			r.NotasSinLeer = append(r.NotasSinLeer, NotaSinLeer{Numero: leida.numero, Motivo: motivo})

			continue
		}

		r.Partes = append(r.Partes, partes...)
	}
}

// partesDeLaNota son las fiestas de una parte de una nota a la que llama una
// sola comunidad y cuyas citas son todas relaciones de fiestas de una parte, o
// el motivo por el que la nota no se lee.
func (r *Resolucion) partesDeLaNota(leida nota) ([]FiestaDeUnaParte, string) {
	var llaman []int

	for indice, columna := range r.Cabecera {
		if slices.Contains(columna.Notas, leida.numero) {
			llaman = append(llaman, indice)
		}
	}

	switch len(llaman) {
	case 0:
		return nil, "ninguna comunidad de la cabecera la llama"
	case 1:
	default:
		return nil, fmt.Sprintf("la llaman %d comunidades de la cabecera, y es de una", len(llaman))
	}

	citas := citaDeLaNota.FindAllStringSubmatch(leida.texto, -1)
	if len(citas) == 0 || strings.Count(leida.texto, "«") != len(citas) || strings.Count(leida.texto, "»") != len(citas) {
		return nil, "no transcribe la disposición de la comunidad entre comillas que se cierran"
	}

	var partes []FiestaDeUnaParte

	for _, cita := range citas {
		deLaCita, motivo := r.relacionDePartes(cita[1], llaman[0], leida.numero)
		if motivo != "" {
			return nil, motivo
		}

		partes = append(partes, deLaCita...)
	}

	return partes, ""
}

// relacionDePartes lee una cita que, tras sus primeros dos puntos, es una
// relación de fiestas de una parte separadas por punto y coma, cada una
// «en <parte>: el <día> de <mes>», o da el motivo por el que no se lee.
func (r *Resolucion) relacionDePartes(cita string, columna, numero int) ([]FiestaDeUnaParte, string) {
	_, relacion, conRelacion := strings.Cut(cita, ": ")
	if !conRelacion {
		return nil, fmt.Sprintf("«%s» no es una relación de fiestas de una parte", cita)
	}

	var partes []FiestaDeUnaParte

	for entrada := range strings.SplitSeq(relacion, "; ") {
		fiesta := fiestaDeUnaParte.FindStringSubmatch(entrada)
		if fiesta == nil {
			return nil, fmt.Sprintf("«%s» no es una fiesta de una parte, «en <parte>: el <día> de <mes>»", entrada)
		}

		mes, esMes := mesDe(fiesta[3])
		fecha, existe := fechaDe(r.Anio, mes, numeroDe(fiesta[2]))

		if !esMes || !existe {
			return nil, fmt.Sprintf("«%s» no da un día de %d", entrada, r.Anio)
		}

		partes = append(partes, FiestaDeUnaParte{Fecha: fecha, Parte: fiesta[1], Columna: columna, Nota: numero})
	}

	return partes, ""
}

// mesDe es el mes que nombra un texto, plegado, si es el nombre de un mes.
func mesDe(texto string) (time.Month, bool) {
	indice := slices.Index(nombresDeLosMeses, territorio.Plegar(texto))

	return time.Month(indice + 1), indice >= 0
}

// fechaDe es el día de un mes de un año como AAAA-MM-DD, si existe.
func fechaDe(anio int, mes time.Month, dia int) (string, bool) {
	fecha := time.Date(anio, mes, dia, 0, 0, 0, 0, time.UTC)
	if fecha.Month() != mes || fecha.Day() != dia {
		return "", false
	}

	return fecha.Format(time.DateOnly), true
}

// numeroDe es el valor de unas cifras ASCII, las que casan las expresiones de
// este fichero, que son como mucho cuatro.
func numeroDe(cifras string) int {
	numero := 0
	for _, cifra := range cifras {
		numero = numero*10 + int(cifra-'0')
	}

	return numero
}
