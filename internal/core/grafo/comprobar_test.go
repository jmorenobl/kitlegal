package grafo_test

import (
	"maps"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/grafo"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Los instantes de las comprobaciones de ejemplo, el reloj de la invocación de
// `graph check` (FR-067).
var (
	// elLunes es el instante de lunes.
	elLunes = time.Date(2026, time.September, 28, 12, 0, 0, 0, time.UTC)
	// elMartes es el instante de martes.
	elMartes = time.Date(2026, time.September, 29, 12, 0, 0, 0, time.UTC)
	// alCaducarLaHora es el instante en que caduca una consulta de lunes con
	// una vigencia de una hora: todavía no ha caducado.
	alCaducarLaHora = elLunes.Add(unaHora)
	// dentroDeUnMes es un instante en el que ya ha caducado cualquier consulta
	// de ejemplo con una vigencia de una semana.
	dentroDeUnMes = time.Date(2026, time.October, 28, 12, 0, 0, 0, time.UTC)
)

// Los datos de la norma de ejemplo y el id de otro bloque suyo.
const (
	identificadorLPAC = "BOE-A-2015-10565"
	idBloqueA22       = idNorma + "#a22"
)

// huellaDeLetra es una huella de ejemplo: «sha256:» y 64 veces la letra, que
// fija el orden de los ids que la llevan.
func huellaDeLetra(letra string) string {
	return "sha256:" + strings.Repeat(letra, 64)
}

// observado es un nodo cuya última observación es la de la url del bloque a21
// en fecha, de la fuente del BOE, sin vigencia.
func observado(id, tipo string, datos map[string]any, fecha string) grafo.NodoDeInstantanea {
	return grafo.NodoDeInstantanea{
		ID: id, Tipo: tipo, Datos: datos,
		UltimaObservacion: grafo.Procedencia{Fuente: fuenteBOE, URL: urlA21, FechaConsulta: fecha},
	}
}

// conVigenciaDe es el mismo nodo con la vigencia que declaró su última
// observación.
func conVigenciaDe(nodo grafo.NodoDeInstantanea, vigencia time.Duration) grafo.NodoDeInstantanea {
	nodo.Vigencia = vigencia

	return nodo
}

// normaLPAC es la norma de ejemplo, con su identificador, observada en fecha.
func normaLPAC(fecha string) grafo.NodoDeInstantanea {
	return observado(idNorma, grafo.TipoNorma, map[string]any{grafo.DatoIdentificador: identificadorLPAC}, fecha)
}

// bloqueDeLaLPAC es el bloque de la norma de ejemplo con ese id y esa parte,
// observado en fecha.
func bloqueDeLaLPAC(id, parte, fecha string) grafo.NodoDeInstantanea {
	return observado(id, grafo.TipoBloque, map[string]any{grafo.DatoBloque: parte}, fecha)
}

// arista es la de esa relación de origen a destino.
func arista(origen, relacion, destino string) schema.Arista {
	return schema.Arista{Origen: origen, Relacion: relacion, Destino: destino}
}

// versionDeEjemplo es una versión del bloque a21: su fecha de vigencia, la
// letra de su huella, la fecha de consulta de su última observación y la
// vigencia que declaró.
type versionDeEjemplo struct {
	fechaVigencia string
	letra         string
	consulta      string
	vigencia      time.Duration
}

// id es el suyo, con la forma que emite boe (contracts/emision.md §1).
func (v versionDeEjemplo) id() string {
	return idBloque + "@" + v.fechaVigencia + ":" + huellaDeLetra(v.letra)
}

// nodo es el suyo, con su fecha de vigencia y su huella como datos.
func (v versionDeEjemplo) nodo() grafo.NodoDeInstantanea {
	return conVigenciaDe(observado(v.id(), grafo.TipoBloqueVersion, map[string]any{
		grafo.DatoFechaVigencia: v.fechaVigencia, grafo.DatoHashTexto: huellaDeLetra(v.letra),
	}, v.consulta), v.vigencia)
}

// conVigencia es la misma versión con otra vigencia.
func (v versionDeEjemplo) conVigencia(vigencia time.Duration) versionDeEjemplo {
	v.vigencia = vigencia

	return v
}

// lecturaDeEjemplo es una lectura de un bloque de la norma de ejemplo: la
// redacción que vio y la fecha de consulta del sobre que la trajo, que es la de
// su última observación si ninguna lectura posterior la vuelve a ver.
type lecturaDeEjemplo struct {
	redaccion redaccionDeEjemplo
	fecha     string
}

// nodo es el de la redacción vista, observada por última vez en la fecha de la
// lectura y sin vigencia.
func (l lecturaDeEjemplo) nodo() grafo.NodoDeInstantanea {
	return l.redaccion.vistaEl(l.fecha)
}

// trasLeer es la instantánea que dejan las lecturas, en su orden, como la lee
// graph check de world.db (data-model §3 y §4): la norma de ejemplo; cada
// bloque leído, con la arista eli:has_part que llega a él y su fila de
// lecturas; y cada redacción vista, con la arista eli:has_version desde su
// bloque y la fecha de su última lectura como última observación. La primera
// lectura de un bloque deja su fila en (v, v) y cada una de las siguientes la
// pasa por Leida. Nada declara vigencia.
func trasLeer(lecturas ...lecturaDeEjemplo) grafo.Instantanea {
	filas := map[string]grafo.LecturasDeBloque{}
	partes := map[string]string{}
	vistas := map[string]lecturaDeEjemplo{}

	for _, lectura := range lecturas {
		bloque, vista := lectura.redaccion.bloque(), lectura.redaccion.id()

		// Un bloque sin leer parte de una fila cuya última es la que ve su
		// primera lectura, así que Leida la deja en (v, v).
		fila, leido := filas[bloque]
		if !leido {
			fila = grafo.LecturasDeBloque{Bloque: bloque, Ultima: vista}
		}

		filas[bloque] = fila.Leida(vista)
		partes[bloque] = lectura.redaccion.parte
		vistas[vista] = lectura
	}

	instantanea := grafo.Instantanea{Nodos: []grafo.NodoDeInstantanea{normaLPAC(lunes)}}

	for _, bloque := range slices.Sorted(maps.Keys(filas)) {
		instantanea.Nodos = append(instantanea.Nodos, bloqueDeLaLPAC(bloque, partes[bloque], lunes))
		instantanea.Aristas = append(instantanea.Aristas, arista(idNorma, grafo.RelacionTieneParte, bloque))
		instantanea.Lecturas = append(instantanea.Lecturas, filas[bloque])
	}

	for _, id := range slices.Sorted(maps.Keys(vistas)) {
		instantanea.Nodos = append(instantanea.Nodos, vistas[id].nodo())
		instantanea.Aristas = append(instantanea.Aristas,
			arista(vistas[id].redaccion.bloque(), grafo.RelacionTieneVersion, id))
	}

	return instantanea
}

// sinFilas es la misma instantánea sin ninguna fila de lecturas: la de un
// world.db de H7, en el que cada bloque cuenta con una sola lectura (FR-026).
func sinFilas(instantanea grafo.Instantanea) grafo.Instantanea {
	instantanea.Lecturas = nil

	return instantanea
}

// conVigenciaEnTodos es la misma instantánea con esa vigencia declarada en la
// última observación de cada nodo.
func conVigenciaEnTodos(instantanea grafo.Instantanea, vigencia time.Duration) grafo.Instantanea {
	nodos := make([]grafo.NodoDeInstantanea, 0, len(instantanea.Nodos))
	for _, nodo := range instantanea.Nodos {
		nodos = append(nodos, conVigenciaDe(nodo, vigencia))
	}

	instantanea.Nodos = nodos

	return instantanea
}

// deNodos es la instantánea de estos nodos, sin aristas.
func deNodos(nodos ...grafo.NodoDeInstantanea) grafo.Instantanea {
	return grafo.Instantanea{Nodos: nodos}
}

// conMas es la misma instantánea con más nodos y más aristas, y con sus filas
// de lecturas.
func conMas(instantanea grafo.Instantanea, nodos []grafo.NodoDeInstantanea, aristas ...schema.Arista) grafo.Instantanea {
	return grafo.Instantanea{
		Nodos:    slices.Concat(instantanea.Nodos, nodos),
		Aristas:  slices.Concat(instantanea.Aristas, aristas),
		Lecturas: slices.Clone(instantanea.Lecturas),
	}
}

// invertida es la misma instantánea con sus nodos, sus aristas y sus filas de
// lecturas en el orden inverso.
func invertida(instantanea grafo.Instantanea) grafo.Instantanea {
	nodos := slices.Clone(instantanea.Nodos)
	aristas := slices.Clone(instantanea.Aristas)
	lecturas := slices.Clone(instantanea.Lecturas)

	slices.Reverse(nodos)
	slices.Reverse(aristas)
	slices.Reverse(lecturas)

	return grafo.Instantanea{Nodos: nodos, Aristas: aristas, Lecturas: lecturas}
}

// obsoleta es el hallazgo, sin la explicación, de la redacción que vio la
// lectura anterior de su bloque, superada por la que vio la última: con la
// fecha de vigencia de las dos y la procedencia de la última observación de la
// reciente (FR-023; data-model §5).
func obsoleta(superada, reciente lecturaDeEjemplo) grafo.Hallazgo {
	return grafo.Hallazgo{
		Clase:                 grafo.ClaseVersionObsoleta,
		ID:                    superada.redaccion.id(),
		Procedencia:           reciente.nodo().UltimaObservacion,
		FechaVigencia:         superada.redaccion.fechaVigencia,
		FechaVigenciaReciente: reciente.redaccion.fechaVigencia,
	}
}

// caducada es el hallazgo, sin la explicación, del nodo cuya última consulta
// ha caducado, con la vigencia que declaró en segundos (FR-066).
func caducada(nodo grafo.NodoDeInstantanea) grafo.Hallazgo {
	return grafo.Hallazgo{
		Clase:            grafo.ClaseFuenteCaducada,
		ID:               nodo.ID,
		Procedencia:      nodo.UltimaObservacion,
		VigenciaSegundos: int64(nodo.Vigencia / time.Second),
	}
}

// sinExplicacion son los mismos hallazgos sin la explicación, que fija
// TestExplicaciones; una lista nula sigue siendo nula.
func sinExplicacion(hallazgos []grafo.Hallazgo) []grafo.Hallazgo {
	sin := slices.Clone(hallazgos)
	for i := range sin {
		sin[i].Explicacion = ""
	}

	return sin
}

// casoDeComprobacion es una instantánea, el instante de la comprobación y los
// hallazgos que da, sin la explicación.
type casoDeComprobacion struct {
	nombre      string
	instantanea grafo.Instantanea
	ahora       time.Time
	esperados   []grafo.Hallazgo
}

// TestComprobar fija las dos reglas de `graph check` y su orden (FR-011,
// FR-023 a FR-026; H7 FR 064, FR 066, FR 067; data-model §4 y §5; research.md
// D5): una version-obsoleta sobre la redacción que vio la lectura anterior de
// un bloque cuando la que vio la última tiene una fecha de vigencia válida y
// estrictamente posterior, con la procedencia y la fecha de la reciente, y
// ninguna sin fila de lecturas; una fuente-caducada por cada nodo cuya
// consulta ha caducado estrictamente antes del instante de la comprobación; y
// una lista nunca nula, con todas las version-obsoleta antes que todas las
// fuente-caducada y por id dentro de cada clase, que no depende del orden de
// la instantánea ni cambia al repetir la comprobación.
func TestComprobar(t *testing.T) {
	t.Parallel()

	for _, caso := range slices.Concat(casosDeLecturas(), casosDeFechasDeVigencia(), casosDeCaducidad(), casosDeOrden()) {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			exigirComprobacion(t, caso)
		})
	}
}

// Las lecturas del bloque a21 de la secuencia de FR-025 (data-model §4): A,
// la grabación de H4, que sirve la fuente el lunes; B, de fecha de vigencia
// posterior, que sirve la fuente el martes y después la caché, con la misma
// fecha de consulta; y C, posterior a las dos, que sirve la fuente el
// miércoles.
var (
	leidaA = lecturaDeEjemplo{redaccion: redaccionA, fecha: lunes}
	leidaB = lecturaDeEjemplo{redaccion: redaccionB, fecha: martes}
	leidaC = lecturaDeEjemplo{redaccion: redaccionC, fecha: miercoles}
)

// casosDeLecturas son los de version-obsoleta decidida por las lecturas: la
// secuencia de FR-025 paso a paso —el paso 3, con --no-graph, no lee y deja la
// instantánea del 2—; un bloque sin fila, que cuenta con una sola lectura
// aunque tenga dos redacciones (FR-026); dos lecturas con la misma fecha de
// vigencia y otra huella; y una última lectura de fecha de vigencia anterior
// (FR-023). Ninguna consulta declara vigencia, así que no hay fuente-caducada.
func casosDeLecturas() []casoDeComprobacion {
	mismaFecha := lecturaDeEjemplo{
		redaccion: redaccionDeEjemplo{parte: "a21", fechaVigencia: redaccionA.fechaVigencia, cuerpo: cuerpoA21 + " Otra huella."},
		fecha:     martes,
	}
	ninguno := []grafo.Hallazgo{}

	return []casoDeComprobacion{
		{"un grafo vacio", grafo.Instantanea{}, dentroDeUnMes, ninguno},
		{"1: la fuente sirve A, una sola lectura", trasLeer(leidaA), dentroDeUnMes, ninguno},
		{
			"2: caducada la cache, la fuente sirve B: sobre A", trasLeer(leidaA, leidaB), dentroDeUnMes,
			[]grafo.Hallazgo{obsoleta(leidaA, leidaB)},
		},
		{"4: la cache sirve B otra vez", trasLeer(leidaA, leidaB, leidaB), dentroDeUnMes, ninguno},
		{
			"5: caducada la cache, la fuente sirve C: sobre B", trasLeer(leidaA, leidaB, leidaB, leidaC), dentroDeUnMes,
			[]grafo.Hallazgo{obsoleta(leidaB, leidaC)},
		},
		{"un bloque sin fila con dos redacciones", sinFilas(trasLeer(leidaA, leidaB)), dentroDeUnMes, ninguno},
		{"la misma fecha de vigencia y otra huella", trasLeer(leidaA, mismaFecha), dentroDeUnMes, ninguno},
		{"una ultima lectura de fecha de vigencia anterior", trasLeer(leidaB, leidaA), dentroDeUnMes, ninguno},
	}
}

// casosDeFechasDeVigencia son los de una fecha de vigencia que no es válida
// (H7 FR 064; research.md D34 de H7): ocho cifras ASCII que nombran un día que
// existe. Una redacción con otra fecha no provoca un hallazgo si la vio la
// última lectura ni lo recibe si la vio la anterior, aunque la otra fecha sea
// válida.
func casosDeFechasDeVigencia() []casoDeComprobacion {
	noValidas := []struct{ nombre, fecha string }{
		{"vacia", ""},
		{"de otra longitud, corta", "2018010"},
		{"de otra longitud, larga", "201801011"},
		{"con guiones", "2018-01-01"},
		{"con signo", "+0180101"},
		{"de un dia que no existe", "20180230"},
		{"de un mes que no existe", "20181301"},
		// Las cifras de 20180101 en anchura completa (U+FF10-U+FF19).
		{"con cifras de anchura completa", "\xef\xbc\x92\xef\xbc\x90\xef\xbc\x91\xef\xbc\x98\xef\xbc\x90\xef\xbc\x91\xef\xbc\x90\xef\xbc\x91"},
		// Las cifras de 20180101 arábigo-índicas (U+0660-U+0669).
		{"con cifras arabigo-indicas", "\xd9\xa2\xd9\xa0\xd9\xa1\xd9\xa8\xd9\xa0\xd9\xa1\xd9\xa0\xd9\xa1"},
	}
	ninguno := []grafo.Hallazgo{}

	casos := make([]casoDeComprobacion, 0, 2*len(noValidas))

	for _, noValida := range noValidas {
		otra := lecturaDeEjemplo{
			redaccion: redaccionDeEjemplo{parte: "a21", fechaVigencia: noValida.fecha, cuerpo: cuerpoA21 + " Otra fecha."},
			fecha:     martes,
		}

		casos = append(casos,
			casoDeComprobacion{"la ultima, con una fecha " + noValida.nombre, trasLeer(leidaA, otra), dentroDeUnMes, ninguno},
			casoDeComprobacion{"la anterior, con una fecha " + noValida.nombre, trasLeer(otra, leidaB), dentroDeUnMes, ninguno},
		)
	}

	return casos
}

// casosDeCaducidad son los de fuente-caducada (FR-066).
func casosDeCaducidad() []casoDeComprobacion {
	deUnaHora := conVigenciaDe(normaLPAC(lunes), unaHora)
	// enMadrid es una consulta de una hora antes que lunes, escrita en otro
	// desplazamiento: caduca en el instante de lunes.
	enMadrid := conVigenciaDe(normaLPAC(lunesAntesEnMadrid), unaHora)
	norma := conVigenciaDe(normaLPAC(lunes), unaSemana)
	bloque := conVigenciaDe(bloqueDeLaLPAC(idBloque, "a21", lunes), unaSemana)
	v2015 := versionDeEjemplo{fechaVigencia: "20150101", letra: "a", consulta: lunes, vigencia: unaSemana}
	municipio := conVigenciaDe(observado(idMunicipio, grafo.TipoMunicipio, map[string]any{}, lunes), unaSemana)
	organo := observado(idOrgano, grafo.TipoOrgano, map[string]any{}, lunes)
	ninguno := []grafo.Hallazgo{}

	return []casoDeComprobacion{
		{
			"caducada un instante despues del limite", deNodos(deUnaHora), alCaducarLaHora.Add(time.Nanosecond),
			[]grafo.Hallazgo{caducada(deUnaHora)},
		},
		{"en el limite exacto", deNodos(deUnaHora), alCaducarLaHora, ninguno},
		{"en el limite exacto, en otro desplazamiento", deNodos(enMadrid), elLunes, ninguno},
		{
			"caducada en otro desplazamiento", deNodos(enMadrid), elLunes.Add(time.Nanosecond),
			[]grafo.Hallazgo{caducada(enMadrid)},
		},
		{
			"una consulta posterior a la comprobacion",
			deNodos(conVigenciaDe(normaLPAC(martes), unaHora)), elLunes, ninguno,
		},
		{"sin vigencia", deNodos(normaLPAC(lunes)), dentroDeUnMes, ninguno},
		{
			"cualquier tipo de nodo, en orden de id",
			deNodos(municipio, organo, v2015.nodo(), bloque, norma), dentroDeUnMes,
			[]grafo.Hallazgo{caducada(norma), caducada(bloque), caducada(v2015.nodo()), caducada(municipio)},
		},
	}
}

// casosDeOrden son los del orden de la lista (FR-011): todas las
// version-obsoleta antes que todas las fuente-caducada, aunque el id de una
// fuente-caducada sea menor, y por id comparando bytes dentro de cada clase,
// cualquiera que sea el orden de las lecturas y de las filas.
func casosDeOrden() []casoDeComprobacion {
	leidaA22 := lecturaDeEjemplo{redaccion: redaccionA22, fecha: lunes}
	leidaA22Posterior := lecturaDeEjemplo{
		redaccion: redaccionDeEjemplo{parte: "a22", fechaVigencia: "20250101", cuerpo: redaccionA22.cuerpo + " Version posterior."},
		fecha:     martes,
	}
	caducadas := conVigenciaEnTodos(trasLeer(leidaA, leidaB), unaSemana)

	esperados := []grafo.Hallazgo{obsoleta(leidaA, leidaB)}
	for _, nodo := range caducadas.Nodos {
		esperados = append(esperados, caducada(nodo))
	}

	slices.SortFunc(esperados[1:], func(a, b grafo.Hallazgo) int { return strings.Compare(a.ID, b.ID) })

	return []casoDeComprobacion{
		{
			"dos bloques, por id",
			trasLeer(leidaA22, leidaA, leidaA22Posterior, leidaB), dentroDeUnMes,
			[]grafo.Hallazgo{obsoleta(leidaA, leidaB), obsoleta(leidaA22, leidaA22Posterior)},
		},
		{"las version-obsoleta antes que las fuente-caducada", caducadas, dentroDeUnMes, esperados},
	}
}

// exigirComprobacion exige que la comprobación del caso dé sus hallazgos, cada
// uno con su explicación, en una lista que no es nula, y la misma lista al
// repetirla y con la instantánea en el orden inverso (FR-011; H7 FR 060).
func exigirComprobacion(t *testing.T, caso casoDeComprobacion) {
	t.Helper()

	hallazgos, err := grafo.Comprobar(caso.instantanea, caso.ahora)
	require.NoError(t, err)
	require.NotNil(t, hallazgos)

	for _, hallazgo := range hallazgos {
		assert.NotEmpty(t, hallazgo.Explicacion, hallazgo.ID)
	}

	assert.Equal(t, caso.esperados, sinExplicacion(hallazgos))

	otraVez, err := grafo.Comprobar(caso.instantanea, caso.ahora)
	require.NoError(t, err)
	assert.Equal(t, hallazgos, otraVez, "otra vez")

	alReves, err := grafo.Comprobar(invertida(caso.instantanea), caso.ahora)
	require.NoError(t, err)
	assert.Equal(t, hallazgos, alReves, "con la instantanea al reves")
}
