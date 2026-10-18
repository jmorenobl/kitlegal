package bocm

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/jmorenobl/kitlegal/internal/core/plazos"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/core/territorio"
	"github.com/jmorenobl/kitlegal/internal/httpx"
)

// NombreDeLaFuente es el nombre lógico de la fuente: el de su fila de
// docs/SOURCES.md, el del manifiesto de grabación, el del directorio de lo
// grabado y el de la clave fuente de data/festivos/locales/13.yaml.
const NombreDeLaFuente = "bocm.fiestas-locales"

// ComunidadDelBoletin es el código de la comunidad cuyo boletín es el BOCM, la
// Comunidad de Madrid: el nombre de su fichero en data/territorio/comunidades/,
// que lo declara como su boletín, y el del fichero de sus festivos locales en
// data/festivos/locales/. El cruce de las líneas se limita a sus municipios
// (FR-022).
const ComunidadDelBoletin = "13"

// Lo que fija la lectura de lo grabado y el orden de las disposiciones.
const (
	// metodoDeLaLectura es el único con el que se pide una disposición: GET.
	metodoDeLaLectura = "GET"
	// estadoDelDocumento es el único estado HTTP de una disposición grabada: lo
	// que no lo trae no es el documento.
	estadoDelDocumento = 200
	// prefijoDelID es el principio de todo identificador, BOCM-AAAAMMDD-N, al
	// que sigue la fecha de publicación.
	prefijoDelID = "BOCM-"
	// formatoDeLaFechaDePublicacion es la fecha de publicación tal como va en
	// el identificador: AAAAMMDD.
	formatoDeLaFechaDePublicacion = "20060102"
)

// Manifiesto es testdata/grabaciones.json: la fuente y las disposiciones que se
// graban y de las que se deriva el fichero (data-model §2 de H9).
type Manifiesto struct {
	Fuente        string        `json:"fuente"`
	Disposiciones []Disposicion `json:"disposiciones"`
}

// Disposicion es una resolución del manifiesto: su identificador en el BOCM,
// el año cuyas fiestas declara o corrige y la dirección de su XML.
type Disposicion struct {
	ID   string `json:"id"`
	Anio int    `json:"anio"`
	URL  string `json:"url"`
}

// publicada es una disposición leída con su fecha de publicación, la de su
// identificador, que es la que la ordena.
type publicada struct {
	resolucion  Resolucion
	publicacion time.Time
}

// anioDerivado es lo que las disposiciones de un año dan al fichero.
type anioDerivado struct {
	// disposiciones son los identificadores en orden de publicación.
	disposiciones []string
	festivos      []plazos.FestivoLocal
	faltas        []plazos.FaltaLocal
}

// entrada es lo que la última disposición que nombra a un municipio dice de
// él: sus fechas, o su falta, y la disposición.
type entrada struct {
	fechas []string
	falta  plazos.Falta
	source string
}

// Derivar deriva data/festivos/locales/13.yaml de las disposiciones del
// manifiesto, que pide al cliente —uno de httpx.Replay sobre lo grabado por
// grabar_datos— y lee con LeerDisposicion, y del territorio, con cuyos
// municipios de la comunidad del boletín cruza cada línea (FR-020 a FR-024;
// research D13; contracts/datos-de-plazos.md §1.2 y §1.3). Las disposiciones de
// cada año se aplican en orden de publicación, el de la fecha de su
// identificador, y cada línea de una posterior sustituye entera la del
// municipio, con su source (FR-021); al final, el municipio sin línea queda
// sin-linea con la última del año, y la línea que no casa con un único
// municipio, sin-municipio con la suya. Lo devuelve para que
// plazos.EscribirFestivosLocales lo escriba en su forma canónica. Recibe la
// lista para que un test componga un año con una parte de sus disposiciones
// (research D30).
//
// Falla, en lugar de dar un fichero incompleto o atribuido a quien no es, con
// un manifiesto de otra fuente, con una disposición dos veces o con un
// identificador sin forma, con una disposición que no se puede reproducir o
// leer, cuyo documento es otro o que es de otro año que el del manifiesto, con
// un territorio sin la comunidad del boletín y con una fecha que no existe en
// el año.
func Derivar(ctx context.Context, cliente *httpx.Cliente, manifiesto Manifiesto, registro *territorio.Registro) (
	plazos.FestivosLocales, error,
) {
	if manifiesto.Fuente != NombreDeLaFuente {
		return plazos.FestivosLocales{}, fmt.Errorf("el manifiesto es de la fuente %q y la derivación, de %q",
			manifiesto.Fuente, NombreDeLaFuente)
	}

	municipios := registro.MunicipiosDe(ComunidadDelBoletin)
	if len(municipios) == 0 {
		return plazos.FestivosLocales{}, fmt.Errorf("el territorio no tiene ningún municipio de la comunidad %s",
			ComunidadDelBoletin)
	}

	porAnio, err := leerGrabadas(ctx, cliente, manifiesto.Disposiciones)
	if err != nil {
		return plazos.FestivosLocales{}, err
	}

	c := nuevoCruce(municipios)
	fichero := plazos.FestivosLocales{Fuente: NombreDeLaFuente, Comunidad: ComunidadDelBoletin}

	for _, anio := range slices.Sorted(maps.Keys(porAnio)) {
		derivado, err := derivarAnio(anio, porAnio[anio], c)
		if err != nil {
			return plazos.FestivosLocales{}, fmt.Errorf("el año %d: %w", anio, err)
		}

		fichero.Anios = append(fichero.Anios, plazos.AnioLocal{Anio: anio, Disposiciones: derivado.disposiciones})
		fichero.Festivos = append(fichero.Festivos, derivado.festivos...)
		fichero.Faltas = append(fichero.Faltas, derivado.faltas...)
	}

	return fichero, nil
}

// leerGrabadas lee las disposiciones del manifiesto y las agrupa por el año
// que les da, cada una con su fecha de publicación.
func leerGrabadas(ctx context.Context, cliente *httpx.Cliente, disposiciones []Disposicion) (map[int][]publicada, error) {
	porAnio := map[int][]publicada{}
	vistas := make(map[string]bool, len(disposiciones))

	for _, disposicion := range disposiciones {
		if vistas[disposicion.ID] {
			return nil, fmt.Errorf("el manifiesto da la disposición %s dos veces", disposicion.ID)
		}

		vistas[disposicion.ID] = true

		publicacion, err := fechaDePublicacion(disposicion.ID)
		if err != nil {
			return nil, err
		}

		resolucion, err := leerGrabada(ctx, cliente, disposicion)
		if err != nil {
			return nil, err
		}

		porAnio[disposicion.Anio] = append(porAnio[disposicion.Anio], publicada{resolucion: resolucion, publicacion: publicacion})
	}

	return porAnio, nil
}

// leerGrabada pide una disposición al cliente y la lee, y exige que sea la del
// manifiesto y del año que le da.
func leerGrabada(ctx context.Context, cliente *httpx.Cliente, disposicion Disposicion) (Resolucion, error) {
	respuesta, err := cliente.Pedir(ctx, schema.Contexto{}, httpx.Peticion{Metodo: metodoDeLaLectura, URL: disposicion.URL})
	if err != nil {
		return Resolucion{}, fmt.Errorf("la disposición %s no se puede pedir a lo grabado: %w", disposicion.ID, err)
	}

	if respuesta.Estado != estadoDelDocumento {
		return Resolucion{}, fmt.Errorf("la disposición %s grabada responde con %d y no con el documento",
			disposicion.ID, respuesta.Estado)
	}

	resolucion, err := LeerDisposicion(respuesta.Cuerpo)
	if err != nil {
		return Resolucion{}, fmt.Errorf("la disposición %s no se lee: %w", disposicion.ID, err)
	}

	if resolucion.ID != disposicion.ID {
		return Resolucion{}, fmt.Errorf("lo grabado para %s es la disposición %s", disposicion.ID, resolucion.ID)
	}

	if resolucion.Anio != disposicion.Anio {
		return Resolucion{}, fmt.Errorf("la disposición %s es del año %d y el manifiesto la da para %d",
			disposicion.ID, resolucion.Anio, disposicion.Anio)
	}

	return resolucion, nil
}

// fechaDePublicacion es la fecha de publicación que lleva el identificador
// BOCM-AAAAMMDD-N: un identificador sin esa forma, o con una fecha que no es un
// día del calendario, es un error que lo nombra.
func fechaDePublicacion(id string) (time.Time, error) {
	resto, conPrefijo := strings.CutPrefix(id, prefijoDelID)
	fecha, numero, conNumero := strings.Cut(resto, "-")

	if !conPrefijo || !conNumero || len(fecha) != len(formatoDeLaFechaDePublicacion) || !sonCifras(fecha) ||
		!sonCifras(numero) {
		return time.Time{}, fmt.Errorf("el identificador %q no tiene la forma %sAAAAMMDD-N", id, prefijoDelID)
	}

	publicacion, err := time.Parse(formatoDeLaFechaDePublicacion, fecha)
	if err != nil {
		return time.Time{}, fmt.Errorf("la fecha de publicación del identificador %q no es un día del calendario: %w", id, err)
	}

	return publicacion, nil
}

// sonCifras dice si el texto no está vacío y es todo cifras ASCII.
func sonCifras(texto string) bool {
	return texto != "" && !strings.ContainsFunc(texto, func(r rune) bool { return r < '0' || r > '9' })
}

// derivarAnio aplica las disposiciones de un año en orden de publicación: cada
// línea de una posterior sustituye entera la del municipio, con su source
// (FR-021; data-model §2, transiciones); la que no casa con un único municipio
// es una falta sin-municipio de su disposición; y, al final, el municipio sin
// línea en ninguna es una falta sin-linea de la última. Con la misma fecha de
// publicación manda el orden en que llegan.
func derivarAnio(anio int, disposiciones []publicada, c *cruce) (anioDerivado, error) {
	if len(disposiciones) == 0 {
		return anioDerivado{}, errors.New("sin ninguna disposición")
	}

	ordenadas := slices.Clone(disposiciones)
	slices.SortStableFunc(ordenadas, func(a, b publicada) int { return a.publicacion.Compare(b.publicacion) })

	entradas := map[string]entrada{}

	var derivado anioDerivado

	for _, disposicion := range ordenadas {
		id := disposicion.resolucion.ID
		derivado.disposiciones = append(derivado.disposiciones, id)

		for _, linea := range disposicion.resolucion.Lineas {
			codigo, casa := c.municipioDe(linea.Nombre)
			if !casa {
				derivado.faltas = append(derivado.faltas, plazos.FaltaLocal{
					Anio: anio, Linea: linea.Nombre, Motivo: plazos.FaltaSinMunicipio, Source: id,
				})

				continue
			}

			fechas, err := fechasDelAnio(anio, linea.Fechas)
			if err != nil {
				return anioDerivado{}, fmt.Errorf("la disposición %s, la línea de «%s»: %w", id, linea.Nombre, err)
			}

			entradas[codigo] = entrada{fechas: fechas, falta: linea.Falta, source: id}
		}
	}

	ultima := derivado.disposiciones[len(derivado.disposiciones)-1]

	for _, municipio := range c.municipios {
		codigo := municipio.Codigo.String()

		switch e, conLinea := entradas[codigo]; {
		case !conLinea:
			derivado.faltas = append(derivado.faltas, plazos.FaltaLocal{
				Anio: anio, Municipio: codigo, Motivo: plazos.FaltaSinLinea, Source: ultima,
			})
		case e.falta != "":
			derivado.faltas = append(derivado.faltas, plazos.FaltaLocal{
				Anio: anio, Municipio: codigo, Motivo: e.falta, Source: e.source,
			})
		default:
			for _, fecha := range e.fechas {
				derivado.festivos = append(derivado.festivos, plazos.FestivoLocal{Fecha: fecha, Municipio: codigo, Source: e.source})
			}
		}
	}

	return derivado, nil
}

// fechasDelAnio pone el año a las fechas de una línea, como AAAA-MM-DD,
// ordenadas y sin repetir. Un día que no existe en ese año —el 29 de febrero
// de uno que no es bisiesto— es un error.
func fechasDelAnio(anio int, fechas []Fecha) ([]string, error) {
	dias := make([]string, 0, len(fechas))

	for _, fecha := range fechas {
		dia := time.Date(anio, fecha.Mes, fecha.Dia, 0, 0, 0, 0, time.UTC)
		if dia.Day() != fecha.Dia {
			return nil, fmt.Errorf("el %d de %s no es un día de %d", fecha.Dia, nombresDeLosMeses[fecha.Mes-1], anio)
		}

		dias = append(dias, dia.Format(time.DateOnly))
	}

	slices.Sort(dias)

	return slices.Compact(dias), nil
}
