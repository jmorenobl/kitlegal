package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/cita"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// La procedencia con la que firma cita. Es un applet calculado: no consulta
// ninguna fuente, tampoco el buscador del CENDOJ, así que firma en el espacio
// de nombres reservado, que es lo que dice que no ha habido consulta (H23
// FR-004; contracts/applet-cita.md §2 de H23; docs/ADR/0036).
const (
	fuenteDeCita = "kitlegal.cita"
	// urlDeCita es la del sobre que no tiene nada que abrir, y la de todo fallo.
	urlDeCita = "kitlegal:applet/cita"
	// prefijoDelDocumento va delante de la huella SHA-256 del texto recibido,
	// en hexadecimal, en la url del sobre de cotejar: lo que ese sobre
	// identifica es el documento que aporta la persona.
	prefijoDelDocumento = "kitlegal:documento/sha256:"
)

// Los dos verbos de cita, como se escriben en la invocación.
const (
	verboPreparar = "preparar"
	verboCotejar  = "cotejar"
)

// AppletCita es el applet de la cita de una sentencia, con dos verbos que hay
// que nombrar: preparar, que dice a una persona qué hacer para encontrarla en
// el buscador del CENDOJ, y cotejar, que dice si el documento que esa persona
// trae es el pedido (H23 FR-001; contracts/applet-cita.md §1 de H23).
//
// No tiene dependencias: no pide nada a la red, no abre la caché ni el grafo
// del mundo y no guarda nada, así que cada invocación depende solo de sus
// argumentos y de su texto (H23 FR-002, FR-003). La entrada estándar tampoco lo
// es: se la da el kernel al verbo que la lee, en cada orden.
func AppletCita() Applet {
	return appletCita{}
}

// appletCita declara lo que declara un applet y nada más: su nombre, su línea
// de ayuda y sus dos verbos.
type appletCita struct{}

func (appletCita) Nombre() string { return "cita" }

func (appletCita) Descripcion() string {
	return "Prepara la consulta de una sentencia en el buscador del CENDOJ y coteja el documento que trae la" +
		" persona. No consulta nada."
}

// Verbos son preparar y cotejar, en su orden y sin ninguno por omisión, con
// sus argumentos y el valor cero del tipo de su data (contracts/applet-cita.md
// §1, §3 y §4 de H23).
func (appletCita) Verbos() []Verbo {
	return []Verbo{
		{
			Nombre: verboPreparar,
			Descripcion: "Dice qué tiene que hacer la persona para encontrar una sentencia en el buscador del CENDOJ:" +
				" la dirección y cada casilla con su valor, o la dirección de una búsqueda por texto. No consulta nada.",
			Argumentos: func() Argumentos { return &argumentosDePreparar{} },
			Salida:     cita.Consulta{},
		},
		{
			Nombre: verboCotejar,
			Descripcion: "Lee la ficha del documento del CENDOJ que trae la persona y dice si es la sentencia que se" +
				" pidió. No consulta nada.",
			Argumentos: func() Argumentos { return &argumentosDeCotejar{} },
			Salida:     cita.Cotejo{},
		},
	}
}

// referenciaDeCita son los cuatro argumentos con los que los dos verbos reciben
// una referencia: el ECLI, por su posición, y el ROJ y el número de resolución
// con su fecha, como banderas (contracts/applet-cita.md §1 de H23).
//
// Son punteros porque un argumento escrito con valor vacío también está dado:
// el analizador los deja en nil cuando no se escriben, y solo entonces. Cuál
// tiene su forma y qué combinaciones valen no lo decide la gramática sino el
// dominio, con sus mensajes (H23 FR-006; research.md D3 de H23).
type referenciaDeCita struct {
	ECLI       *string `arg:"" optional:"" name:"ecli" help:"ECLI de la sentencia, español: ECLI:ES:<órgano>:<año>:<número>."`
	ROJ        *string `name:"roj" placeholder:"<ROJ>" help:"ROJ de la sentencia, en lugar del ECLI: <siglas> <número>/<año>."`
	Resolucion *string `placeholder:"<n>/<año>" help:"Número de resolución de la sentencia, en lugar del ECLI; va con --fecha."`
	Fecha      *string `placeholder:"AAAA-MM-DD" help:"Fecha de la resolución cuyo número da --resolucion."`
}

// referencia es la que dan los cuatro argumentos y si dan alguna, o el error de
// argumentos del dominio que la rechaza.
func (r referenciaDeCita) referencia() (cita.Referencia, bool, error) {
	return cita.NuevaReferencia(r.ECLI, r.ROJ, r.Resolucion, r.Fecha)
}

// argumentosDePreparar son los de preparar: una referencia o el texto de una
// búsqueda por materia, nunca los dos (contracts/applet-cita.md §1 de H23).
type argumentosDePreparar struct {
	referenciaDeCita

	Texto *string `placeholder:"<texto>" help:"Texto de una búsqueda por materia, en lugar de una referencia."`
}

// Ejecutar prepara la consulta de la referencia o la de la búsqueda por texto.
// No abre ninguna conexión ni toca la caché, así que --offline no cambia nada,
// y no describe ninguna operación con --dry-run, porque no tiene ninguna capa
// con efectos.
//
// La referencia se comprueba antes que el texto. Con una bien dada, el texto
// sobra, también vacío; sin ninguna, hace falta el texto; y el vacío o solo de
// blancos lo rechaza el dominio. Todo rechazo es de clase «argumentos» (H23
// FR-006, FR-015; contracts/applet-cita.md §6 de H23).
//
// La url del sobre es la dirección que abre la persona, o la del applet si no
// hay nada que abrir. La fecha no la declara: la pone el reloj del kernel
// (contracts/applet-cita.md §2 de H23).
func (a *argumentosDePreparar) Ejecutar(
	_ context.Context, _ schema.Contexto, _ *slog.Logger,
) (schema.Resultado, error) {
	referencia, hay, err := a.referencia()
	if err != nil {
		return falloDeCita(err)
	}

	consulta, err := a.consulta(referencia, hay)
	if err != nil {
		return falloDeCita(err)
	}

	url := consulta.Direccion
	if url == "" {
		url = urlDeCita
	}

	return schema.Resultado{Procedencia: schema.Procedencia{Fuente: fuenteDeCita, URL: url}, Datos: consulta}, nil
}

// consulta es la de la referencia, si se dio una, o la del texto de búsqueda.
func (a *argumentosDePreparar) consulta(referencia cita.Referencia, hay bool) (cita.Consulta, error) {
	switch {
	case hay && a.Texto != nil:
		return cita.Consulta{}, fmt.Errorf("%w: una referencia y --texto no se pueden dar a la vez: se prepara la"+
			" consulta de una sentencia, por su referencia, o una búsqueda por texto", cli.ErrArgumentos)
	case hay:
		return cita.Preparar(referencia), nil
	case a.Texto != nil:
		return cita.PrepararTexto(*a.Texto)
	}

	return cita.Consulta{}, fmt.Errorf("%w: cita preparar necesita una referencia —un ECLI, --roj, o --resolucion"+
		" con --fecha— o el texto de una búsqueda, con --texto, y no se ha dado ninguno", cli.ErrArgumentos)
}

// argumentosDeCotejar son los de cotejar: el texto del documento y,
// opcionalmente, la referencia que se había pedido (contracts/applet-cita.md §1
// de H23).
//
// Documento es un literal para que el texto dado por la bandera y el dado por
// la entrada estándar sean los mismos bytes, también los que no son UTF-8, y
// den la misma huella; y un puntero porque escrito con valor vacío también está
// dado: es el texto, y la entrada no se lee (H23 FR-020; research.md D12 de
// H23). La entrada va en un campo no exportado, que la gramática no ve.
type argumentosDeCotejar struct {
	referenciaDeCita

	Documento *cli.Literal `placeholder:"<texto>" help:"Texto del documento, con su ficha; sin él, se lee de la entrada estándar."`

	// entrada es la de la invocación, que da el kernel antes de Ejecutar: la
	// entrada estándar en una orden, y nil en una llamada de herramienta.
	entrada io.Reader
}

// leerDe recibe la entrada de la invocación. De ella solo se lee al ejecutar, y
// solo si no se escribió --documento.
func (a *argumentosDeCotejar) leerDe(entrada io.Reader) {
	a.entrada = entrada
}

// Ejecutar lee la ficha del texto recibido y la coteja con la referencia que se
// pidió, si se dio una. Como preparar, no abre ninguna conexión, no toca la
// caché y no tiene nada que describir con --dry-run.
//
// La referencia se comprueba antes que el texto. Sin texto —un --documento
// vacío o, sin él, una entrada vacía o ninguna— y con un texto sin ficha
// reconocible, el rechazo es de clase «argumentos». Que el documento no sea el
// pedido no es un fallo: es un hallazgo en data (H23 FR-006, FR-025, FR-026;
// contracts/applet-cita.md §6 de H23).
//
// La url del sobre lleva la huella del texto recibido, byte a byte: el mismo
// texto da siempre la misma, y uno distinto, otra. La fecha no la declara: la
// pone el reloj del kernel (contracts/applet-cita.md §2 de H23).
func (a *argumentosDeCotejar) Ejecutar(
	_ context.Context, _ schema.Contexto, _ *slog.Logger,
) (schema.Resultado, error) {
	referencia, hay, err := a.referencia()
	if err != nil {
		return falloDeCita(err)
	}

	texto, err := a.texto()
	if err != nil {
		return schema.Resultado{}, err
	}

	if texto == "" {
		return falloDeCita(fmt.Errorf("%w: cita cotejar no ha recibido el texto del documento: va en --documento o,"+
			" en la orden y sin él, por la entrada estándar", cli.ErrArgumentos))
	}

	ficha, err := cita.LeerFicha(texto)
	if err != nil {
		return falloDeCita(err)
	}

	var pedida *cita.Referencia
	if hay {
		pedida = &referencia
	}

	huella := sha256.Sum256([]byte(texto))

	return schema.Resultado{
		Procedencia: schema.Procedencia{Fuente: fuenteDeCita, URL: prefijoDelDocumento + hex.EncodeToString(huella[:])},
		Datos:       cita.Cotejar(ficha, pedida),
	}, nil
}

// texto es el del documento: el de --documento si se escribió, también vacío, y
// solo sin él el de la entrada, leída hasta su final. Sin ninguno de los dos
// —una llamada de herramienta no tiene entrada— no hay texto (H23 FR-020).
//
// Una entrada que no se puede leer no es algo que quien invoca pueda corregir
// con otros argumentos: sale como inesperado, sin procedencia, que firma el
// kernel.
func (a *argumentosDeCotejar) texto() (string, error) {
	if a.Documento != nil {
		return string(*a.Documento), nil
	}

	if a.entrada == nil {
		return "", nil
	}

	leido, err := io.ReadAll(a.entrada)
	if err != nil {
		return "", fmt.Errorf("cita: la entrada estándar no se puede leer: %w", err)
	}

	return string(leido), nil
}

// falloDeCita es el rechazo de unos argumentos, con la firma del applet: un
// fallo no tiene nada que abrir ni documento que identificar.
func falloDeCita(err error) (schema.Resultado, error) {
	return schema.Resultado{Procedencia: schema.Procedencia{Fuente: fuenteDeCita, URL: urlDeCita}}, err
}

// Las comprobaciones en tiempo de compilación del contrato del applet.
var (
	_ Applet     = appletCita{}
	_ Argumentos = (*argumentosDePreparar)(nil)
	_ Argumentos = (*argumentosDeCotejar)(nil)
	_ lector     = (*argumentosDeCotejar)(nil)
)
