package fiestas

import (
	"cmp"
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/jmorenobl/kitlegal/internal/core/plazos"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/core/territorio"
	"github.com/jmorenobl/kitlegal/internal/httpx"
)

// NombreDeLaFuente es el nombre lógico de la fuente: el de su fila de
// docs/SOURCES.md, el del manifiesto de grabación, el del directorio de lo
// grabado y el de la clave fuente de data/festivos/nacionales-y-autonomicos.yaml.
const NombreDeLaFuente = "boe.fiestas-laborales"

// estadoDelDocumento es el único estado HTTP de una resolución grabada: lo que
// no lo trae no es el documento.
const estadoDelDocumento = 200

// Manifiesto es testdata/grabaciones.json: la fuente y las resoluciones que se
// graban y de las que se deriva el fichero (data-model §2 de H9).
type Manifiesto struct {
	Fuente        string        `json:"fuente"`
	Disposiciones []Disposicion `json:"disposiciones"`
}

// Disposicion es una resolución: su identificador en el BOE, el año cuyas
// fiestas declara y la dirección de su documento en el diario.
type Disposicion struct {
	ID   string `json:"id"`
	Anio int    `json:"anio"`
	URL  string `json:"url"`
}

// Derivar deriva data/festivos/nacionales-y-autonomicos.yaml de las
// resoluciones del manifiesto, que pide al cliente —uno de httpx.Replay sobre
// lo grabado por grabar_datos— y lee con LeerResolucion, y del territorio, con
// cuyas comunidades cruza la cabecera de cada una (FR-020, FR-025; research
// D14, S2; contracts/datos-de-plazos.md §1.1 y §1.3): una fila por fecha,
// ámbito y parte, con las comunidades que la resolución marca así y con source
// la resolución de su año, que plazos.EscribirFestivosNacionales escribe en su
// forma canónica.
//
// Falla, en lugar de dar un fichero incompleto, con un manifiesto de otra
// fuente o con un año dos veces, con una resolución que no se puede reproducir
// o leer o cuyo anexo es de otro año, con un nombre de la cabecera que no casa
// con una sola comunidad o que casa con la de otro, y con una comunidad sin
// festivos de toda ella en algún año: una columna que no se lee no queda sin
// rastro.
func Derivar(ctx context.Context, cliente *httpx.Cliente, manifiesto Manifiesto, registro *territorio.Registro) (
	plazos.FestivosNacionales, error,
) {
	if manifiesto.Fuente != NombreDeLaFuente {
		return plazos.FestivosNacionales{}, fmt.Errorf("el manifiesto es de la fuente %q y la derivación, de %q",
			manifiesto.Fuente, NombreDeLaFuente)
	}

	comunidades := registro.Comunidades()
	fichero := plazos.FestivosNacionales{Fuente: NombreDeLaFuente}

	for _, disposicion := range manifiesto.Disposiciones {
		if slices.ContainsFunc(fichero.Anios, func(a plazos.AnioNacional) bool { return a.Anio == disposicion.Anio }) {
			return plazos.FestivosNacionales{}, fmt.Errorf("el manifiesto da el año %d dos veces", disposicion.Anio)
		}

		resolucion, err := leerGrabada(ctx, cliente, disposicion)
		if err != nil {
			return plazos.FestivosNacionales{}, err
		}

		festivos, err := festivosDe(resolucion, disposicion.ID, comunidades)
		if err != nil {
			return plazos.FestivosNacionales{}, fmt.Errorf("la resolución %s: %w", disposicion.ID, err)
		}

		fichero.Anios = append(fichero.Anios, plazos.AnioNacional{Anio: disposicion.Anio, Disposicion: disposicion.ID})
		fichero.Festivos = append(fichero.Festivos, festivos...)
	}

	return fichero, nil
}

// leerGrabada pide una resolución al cliente y la lee, y exige que su anexo sea
// del año que le da el manifiesto.
func leerGrabada(ctx context.Context, cliente *httpx.Cliente, disposicion Disposicion) (Resolucion, error) {
	respuesta, err := cliente.Pedir(ctx, schema.Contexto{}, httpx.Peticion{Metodo: metodoDeLaLectura, URL: disposicion.URL})
	if err != nil {
		return Resolucion{}, fmt.Errorf("la resolución %s no se puede pedir a lo grabado: %w", disposicion.ID, err)
	}

	if respuesta.Estado != estadoDelDocumento {
		return Resolucion{}, fmt.Errorf("la resolución %s grabada responde con %d y no con el documento",
			disposicion.ID, respuesta.Estado)
	}

	resolucion, err := LeerResolucion(respuesta.Cuerpo)
	if err != nil {
		return Resolucion{}, fmt.Errorf("la resolución %s no se lee: %w", disposicion.ID, err)
	}

	if resolucion.Anio != disposicion.Anio {
		return Resolucion{}, fmt.Errorf("el anexo de la resolución %s es de %d y el manifiesto la da para %d",
			disposicion.ID, resolucion.Anio, disposicion.Anio)
	}

	return resolucion, nil
}

// claveDeFila es lo que hace distinta una fila del fichero.
type claveDeFila struct {
	fecha  string
	ambito plazos.Ambito
	parte  string
}

// festivosDe son las filas que da una resolución ya leída, con source el de su
// año: por cada fecha, ámbito y parte, las comunidades de la cabecera que la
// marcan así, y las fiestas de una parte, autonómicas, de la comunidad que
// llama a su nota. Falla si un nombre de la cabecera no da una comunidad sola y
// distinta de las demás, o si una comunidad no tiene ningún festivo de toda
// ella.
func festivosDe(resolucion Resolucion, source string, comunidades []territorio.Comunidad) (
	[]plazos.FestivoNacional, error,
) {
	codigos, err := codigosDeLaCabecera(resolucion.Cabecera, comunidades)
	if err != nil {
		return nil, err
	}

	filas := map[claveDeFila][]string{}

	for _, fila := range resolucion.Filas {
		for indice, ambito := range fila.Ambitos {
			if ambito != "" {
				clave := claveDeFila{fecha: fila.Fecha, ambito: ambito}
				filas[clave] = append(filas[clave], codigos[indice])
			}
		}
	}

	for _, parte := range resolucion.Partes {
		clave := claveDeFila{fecha: parte.Fecha, ambito: plazos.AmbitoAutonomico, parte: parte.Parte}
		filas[clave] = append(filas[clave], codigos[parte.Columna])
	}

	claves := slices.SortedFunc(maps.Keys(filas), func(a, b claveDeFila) int {
		return cmp.Or(cmp.Compare(a.fecha, b.fecha), cmp.Compare(a.ambito, b.ambito), cmp.Compare(a.parte, b.parte))
	})

	festivos := make([]plazos.FestivoNacional, 0, len(claves))
	for _, clave := range claves {
		festivos = append(festivos, plazos.FestivoNacional{
			Fecha:       clave.fecha,
			Ambito:      clave.ambito,
			Comunidades: slices.Compact(slices.Sorted(slices.Values(filas[clave]))),
			Parte:       clave.parte,
			Source:      source,
		})
	}

	for _, comunidad := range comunidades {
		if !slices.ContainsFunc(festivos, func(festivo plazos.FestivoNacional) bool {
			return festivo.Parte == "" && slices.Contains(festivo.Comunidades, comunidad.Codigo)
		}) {
			return nil, fmt.Errorf("no da ningún festivo de toda %s (%s)", comunidad.Nombre, comunidad.Codigo)
		}
	}

	return festivos, nil
}

// codigosDeLaCabecera da, en el orden de la cabecera, el código de la comunidad
// de cada nombre, y falla si dos nombres casan con la misma.
func codigosDeLaCabecera(cabecera []Columna, comunidades []territorio.Comunidad) ([]string, error) {
	codigos := make([]string, 0, len(cabecera))

	for indice, columna := range cabecera {
		comunidad, err := comunidadDeLaColumna(columna.Nombre, comunidades)
		if err != nil {
			return nil, err
		}

		if anterior := slices.Index(codigos, comunidad.Codigo); anterior >= 0 {
			return nil, fmt.Errorf("«%s» y «%s», de la cabecera, casan las dos con %s (%s)",
				cabecera[anterior].Nombre, cabecera[indice].Nombre, comunidad.Nombre, comunidad.Codigo)
		}

		codigos = append(codigos, comunidad.Codigo)
	}

	return codigos, nil
}

// reglasDelCruce son, en su orden, las reglas con las que un nombre de la
// cabecera casa con el de una comunidad de data/territorio/, sin ninguna lista
// de alias (research D14, S2; supuesto de T007 en gates/supuestos.md): cada una
// se aplica solo si la anterior no da ninguna comunidad.
var reglasDelCruce = []func(columna, comunidad string) bool{
	mismaFormaCompacta,
	mismasPalabras,
	conTituloDelante,
}

// comunidadDeLaColumna es la única comunidad con la que casa un nombre de la
// cabecera: la de la primera regla del cruce que da alguna. Si esa da más de
// una, o ninguna regla da ninguna, el nombre no se atribuye y es un error que
// lo nombra.
func comunidadDeLaColumna(columna string, comunidades []territorio.Comunidad) (territorio.Comunidad, error) {
	for _, regla := range reglasDelCruce {
		var casan []territorio.Comunidad

		for _, comunidad := range comunidades {
			if regla(columna, comunidad.Nombre) {
				casan = append(casan, comunidad)
			}
		}

		switch len(casan) {
		case 0:
			continue
		case 1:
			return casan[0], nil
		default:
			nombres := make([]string, 0, len(casan))
			for _, comunidad := range casan {
				nombres = append(nombres, fmt.Sprintf("%s (%s)", comunidad.Nombre, comunidad.Codigo))
			}

			return territorio.Comunidad{}, fmt.Errorf("«%s», de la cabecera, casa con más de una comunidad: %s",
				columna, strings.Join(nombres, ", "))
		}
	}

	return territorio.Comunidad{}, fmt.Errorf("«%s», de la cabecera, no casa con ninguna de las %d comunidades de "+
		"data/territorio/", columna, len(comunidades))
}

// mismaFormaCompacta es la regla de D13: las dos formas compactas, iguales
// («Castilla-La Mancha» y «Castilla - La Mancha», «Región Murcia» y «Región de
// Murcia»).
func mismaFormaCompacta(columna, comunidad string) bool {
	compacta := territorio.FormaCompacta(columna)

	return compacta != "" && compacta == territorio.FormaCompacta(comunidad)
}

// mismasPalabras dice si los dos nombres, sin las palabras que quita la forma
// compacta, tienen las mismas palabras en el mismo orden, donde una palabra
// abreviada con un punto casa con la que empieza por ella («Com. Madrid» y
// «Comunidad de Madrid», «C. Foral Navarra» y «Comunidad Foral de Navarra»).
func mismasPalabras(columna, comunidad string) bool {
	deLaColumna, deLaComunidad := significativas(palabrasDe(columna)), significativas(palabrasDe(comunidad))

	return len(deLaColumna) > 0 && slices.EqualFunc(deLaColumna, deLaComunidad, casan)
}

// conTituloDelante dice si uno de los dos nombres es el otro con algo delante
// unido por «de» o «del», como D13 admite el nombre seguido de «de» o «del» y
// algo detrás: «Ciudad de Ceuta» y «Ceuta», «Asturias» y «Principado de
// Asturias».
func conTituloDelante(columna, comunidad string) bool {
	largo, corto := palabrasDe(columna), palabrasDe(comunidad)
	if len(largo) < len(corto) {
		largo, corto = corto, largo
	}

	delante := len(largo) - len(corto)
	if len(corto) == 0 || delante < 2 {
		return false
	}

	if union := largo[delante-1].plegada; union != "de" && union != "del" {
		return false
	}

	return slices.EqualFunc(largo[delante:], corto, casan)
}

// palabra es una palabra plegada de un nombre y si el nombre la escribe
// abreviada, con un punto detrás.
type palabra struct {
	plegada   string
	abreviada bool
}

// palabrasDe son las palabras plegadas de un nombre, en su orden.
func palabrasDe(nombre string) []palabra {
	var palabras []palabra

	for trozo := range strings.FieldsSeq(nombre) {
		plegadas := strings.Fields(territorio.Plegar(trozo))
		for indice, plegada := range plegadas {
			palabras = append(palabras, palabra{
				plegada:   plegada,
				abreviada: indice == len(plegadas)-1 && strings.HasSuffix(trozo, "."),
			})
		}
	}

	return palabras
}

// significativas son las palabras que no quita la forma compacta.
func significativas(palabras []palabra) []palabra {
	return slices.DeleteFunc(slices.Clone(palabras), func(p palabra) bool { return territorio.FormaCompacta(p.plegada) == "" })
}

// casan dice si dos palabras son la misma, o una abreviada es el principio de
// la otra.
func casan(a, b palabra) bool {
	return a.plegada == b.plegada ||
		a.abreviada && strings.HasPrefix(b.plegada, a.plegada) ||
		b.abreviada && strings.HasPrefix(a.plegada, b.plegada)
}
