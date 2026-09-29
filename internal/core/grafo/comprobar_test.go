package grafo_test

import (
	"maps"
	"slices"
	"strconv"
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
	// trasCaducarElMiercoles es un instante después de que caduque una
	// consulta del miércoles con una vigencia de una semana —y con ella las del
	// lunes y el martes— y antes de que caduque una del jueves.
	trasCaducarElMiercoles = time.Date(2026, time.October, 7, 12, 0, 0, 1, time.UTC)
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
// graph check de world.db (data-model §3 y §4): la norma de ejemplo, observada
// por última vez en la fecha de la última lectura, que la renueva la de
// cualquiera de sus bloques (H7.1 FR-031); cada bloque leído, con la arista
// eli:has_part que llega a él, su fila de lecturas y la fecha de su última
// lectura como última observación; y cada redacción vista, con la arista
// eli:has_version desde su bloque y la fecha de su última lectura como última
// observación. La primera lectura de un bloque deja su fila en (v, v) y cada
// una de las siguientes la pasa por Leida. Nada declara vigencia.
func trasLeer(lecturas ...lecturaDeEjemplo) grafo.Instantanea {
	filas := map[string]grafo.LecturasDeBloque{}
	ultimas := map[string]lecturaDeEjemplo{}
	vistas := map[string]lecturaDeEjemplo{}
	normaLeida := lunes

	for _, lectura := range lecturas {
		bloque, vista := lectura.redaccion.bloque(), lectura.redaccion.id()

		// Un bloque sin leer parte de una fila cuya última es la que ve su
		// primera lectura, así que Leida la deja en (v, v).
		fila, leido := filas[bloque]
		if !leido {
			fila = grafo.LecturasDeBloque{Bloque: bloque, Ultima: vista}
		}

		filas[bloque] = fila.Leida(vista)
		ultimas[bloque] = lectura
		vistas[vista] = lectura
		normaLeida = lectura.fecha
	}

	instantanea := grafo.Instantanea{Nodos: []grafo.NodoDeInstantanea{normaLPAC(normaLeida)}}

	for _, bloque := range slices.Sorted(maps.Keys(filas)) {
		instantanea.Nodos = append(instantanea.Nodos,
			bloqueDeLaLPAC(bloque, ultimas[bloque].redaccion.parte, ultimas[bloque].fecha))
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
// FR-023 a FR-026, FR-030 a FR-032; SC-004; H7 FR 064, FR 066, FR 067;
// data-model §4 y §5; research.md D5): una version-obsoleta sobre la redacción
// que vio la lectura anterior de un bloque cuando la que vio la última tiene
// una fecha de vigencia válida y estrictamente posterior, con la procedencia y
// la fecha de la reciente, y ninguna sin fila de lecturas; una fuente-caducada
// por cada Norma, cada Bloque y la redacción vista de cada bloque cuya consulta
// ha caducado estrictamente antes del instante de la comprobación, y ninguna en
// otra BloqueVersion ni en otro tipo de nodo; y una lista nunca nula, con todas
// las version-obsoleta antes que todas las fuente-caducada y por id dentro de
// cada clase, con el total de cada clase (FR-012), que no depende del orden de
// la instantánea ni cambia al repetir la comprobación. La cota la fija
// TestComprobarConLaCota.
func TestComprobar(t *testing.T) {
	t.Parallel()

	casos := slices.Concat(casosDeLecturas(), casosDeFechasDeVigencia(), casosDeCaducidad(), casosDeLoVigente(),
		casosDeOrden())

	for _, caso := range casos {
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

// casosDeCaducidad son los de cuándo caduca una consulta (FR-066).
func casosDeCaducidad() []casoDeComprobacion {
	deUnaHora := conVigenciaDe(normaLPAC(lunes), unaHora)
	// enMadrid es una consulta de una hora antes que lunes, escrita en otro
	// desplazamiento: caduca en el instante de lunes.
	enMadrid := conVigenciaDe(normaLPAC(lunesAntesEnMadrid), unaHora)
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
	}
}

// casosDeLoVigente son los de los nodos que dan fuente-caducada (H7.1 FR-030 a
// FR-032; SC-004; data-model §5): la Norma, el Bloque y la redacción vista de
// cada bloque —la que vio su última lectura o, sin fila, la de
// RedaccionVistaSinLecturas—, y nunca otra BloqueVersion ni otro tipo de nodo,
// aunque su consulta también haya caducado. Cada lectura declara la semana de
// vigencia de boe, que renueva la siguiente lectura del bloque.
func casosDeLoVigente() []casoDeComprobacion {
	// La secuencia de FR-025 hasta el paso 5 y con otra lectura, el jueves,
	// que ve C otra vez.
	paso5 := conVigenciaEnTodos(trasLeer(leidaA, leidaB, leidaB, leidaC), unaSemana)
	leidaCElJueves := lecturaDeEjemplo{redaccion: redaccionC, fecha: jueves}
	renovada := conVigenciaEnTodos(trasLeer(leidaA, leidaB, leidaB, leidaC, leidaCElJueves), unaSemana)

	semanal := func(nodo grafo.NodoDeInstantanea) grafo.NodoDeInstantanea { return conVigenciaDe(nodo, unaSemana) }
	v2015 := versionDeEjemplo{fechaVigencia: "20150101", letra: "a", consulta: lunes, vigencia: unaSemana}
	municipio := semanal(observado(idMunicipio, grafo.TipoMunicipio, map[string]any{}, lunes))
	organo := semanal(observado(idOrgano, grafo.TipoOrgano, map[string]any{}, lunes))
	ninguno := []grafo.Hallazgo{}

	return []casoDeComprobacion{
		{
			"SC-004: tras el paso 5, pasada la vigencia de la lectura de C, ninguna sobre A ni B", paso5,
			trasCaducarElMiercoles,
			[]grafo.Hallazgo{
				obsoleta(leidaB, leidaC),
				caducada(semanal(normaLPAC(miercoles))),
				caducada(semanal(bloqueDeLaLPAC(idBloque, "a21", miercoles))),
				caducada(semanal(leidaC.nodo())),
			},
		},
		{"FR-031: otra lectura que ve C las renueva", renovada, trasCaducarElMiercoles, ninguno},
		{
			"un bloque sin fila, solo su redaccion vista", sinFilas(conVigenciaEnTodos(trasLeer(leidaA, leidaB), unaSemana)),
			dentroDeUnMes,
			[]grafo.Hallazgo{
				caducada(semanal(normaLPAC(martes))),
				caducada(semanal(bloqueDeLaLPAC(idBloque, "a21", martes))),
				caducada(semanal(leidaB.nodo())),
			},
		},
		{
			"la Norma, el Bloque y su redaccion vista, en orden de id",
			conMas(deNodos(municipio, organo, v2015.nodo(), semanal(bloqueDeLaLPAC(idBloque, "a21", lunes)),
				semanal(normaLPAC(lunes))), nil, arista(idBloque, grafo.RelacionTieneVersion, v2015.id())),
			dentroDeUnMes,
			[]grafo.Hallazgo{
				caducada(semanal(normaLPAC(lunes))),
				caducada(semanal(bloqueDeLaLPAC(idBloque, "a21", lunes))),
				caducada(v2015.nodo()),
			},
		},
		{"una BloqueVersion que no es la redaccion vista de ningun bloque", deNodos(v2015.nodo()), dentroDeUnMes, ninguno},
		{"un Municipio y un Organo con la vigencia pasada", deNodos(municipio, organo), dentroDeUnMes, ninguno},
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
	esperados := []grafo.Hallazgo{
		obsoleta(leidaA, leidaB),
		caducada(conVigenciaDe(normaLPAC(martes), unaSemana)),
		caducada(conVigenciaDe(bloqueDeLaLPAC(idBloque, "a21", martes), unaSemana)),
		caducada(conVigenciaDe(leidaB.nodo(), unaSemana)),
	}

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
// uno con su explicación, en una lista que no es nula, con el total de cada
// clase y ninguno omitido, y la misma comprobación al repetirla y con la
// instantánea en el orden inverso (FR-011, FR-012; H7 FR 060).
func exigirComprobacion(t *testing.T, caso casoDeComprobacion) {
	t.Helper()

	comprobacion, err := grafo.Comprobar(caso.instantanea, grafo.Ambito{}, caso.ahora)
	require.NoError(t, err)
	require.NotNil(t, comprobacion.Hallazgos)

	for _, hallazgo := range comprobacion.Hallazgos {
		assert.NotEmpty(t, hallazgo.Explicacion, hallazgo.ID)
	}

	assert.Equal(t, caso.esperados, sinExplicacion(comprobacion.Hallazgos))
	assert.Equal(t, contarPorClase(caso.esperados, grafo.ClaseVersionObsoleta), comprobacion.VersionObsoleta)
	assert.Equal(t, contarPorClase(caso.esperados, grafo.ClaseFuenteCaducada), comprobacion.FuenteCaducada)
	assert.Zero(t, comprobacion.Omitidos)

	otraVez, err := grafo.Comprobar(caso.instantanea, grafo.Ambito{}, caso.ahora)
	require.NoError(t, err)
	assert.Equal(t, comprobacion, otraVez, "otra vez")

	alReves, err := grafo.Comprobar(invertida(caso.instantanea), grafo.Ambito{}, caso.ahora)
	require.NoError(t, err)
	assert.Equal(t, comprobacion, alReves, "con la instantanea al reves")
}

// contarPorClase es cuántos de los hallazgos son de la clase.
func contarPorClase(hallazgos []grafo.Hallazgo, clase grafo.ClaseDeHallazgo) int {
	cuantos := 0

	for _, hallazgo := range hallazgos {
		if hallazgo.Clase == clase {
			cuantos++
		}
	}

	return cuantos
}

// TestComprobarConLaCota fija la cota de `graph check` (FR-010 a FR-012;
// data-model §5 y §6; research.md D10): Comprobar calcula todos los hallazgos,
// los ordena —todas las version-obsoleta antes que todas las fuente-caducada y
// por id comparando bytes dentro de cada clase— y lista los
// grafo.MaximoDeHallazgos primeros; los totales de cada clase cuentan también
// los que no se listan, y omitidos es su suma menos los listados.
func TestComprobarConLaCota(t *testing.T) {
	t.Parallel()

	require.Equal(t, 50, grafo.MaximoDeHallazgos, "la cota de FR-010")

	casos := []struct {
		nombre               string
		obsoletas, caducadas int
	}{
		{"sin hallazgos", 0, 0},
		{"exactamente 50: ninguno omitido", 0, grafo.MaximoDeHallazgos},
		{"51 fuente-caducada: se omite la ultima", 0, grafo.MaximoDeHallazgos + 1},
		{"51 version-obsoleta: se omite la ultima", grafo.MaximoDeHallazgos + 1, 0},
		{"la cota corta dentro de las fuente-caducada", 3, 60},
		{"la cota corta dentro de las version-obsoleta y no lista ninguna fuente-caducada", 55, 5},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			obsoletas, instantanea := bloquesQueCambiaron(caso.obsoletas)
			caducadas := normasCaducadas(caso.caducadas)
			instantanea = conMas(instantanea, caducadas)

			var ids []string

			for _, id := range slices.Sorted(slices.Values(obsoletas)) {
				ids = append(ids, string(grafo.ClaseVersionObsoleta)+" "+id)
			}

			for _, nodo := range slices.SortedFunc(slices.Values(caducadas), porID) {
				ids = append(ids, string(grafo.ClaseFuenteCaducada)+" "+nodo.ID)
			}

			listados := min(len(ids), grafo.MaximoDeHallazgos)

			comprobacion, err := grafo.Comprobar(instantanea, grafo.Ambito{}, dentroDeUnMes)
			require.NoError(t, err)

			assert.Equal(t, caso.obsoletas, comprobacion.VersionObsoleta, "el total cuenta los omitidos")
			assert.Equal(t, caso.caducadas, comprobacion.FuenteCaducada, "el total cuenta los omitidos")
			assert.Equal(t, caso.obsoletas+caso.caducadas-listados, comprobacion.Omitidos)
			assert.Equal(t, ids[:listados], clasesEIDs(comprobacion.Hallazgos), "los primeros, en orden")
		})
	}
}

// bloquesQueCambiaron son n bloques de la norma de ejemplo, cada uno leído el
// lunes con una redacción y el martes con otra de fecha de vigencia posterior,
// sin vigencia declarada: los ids de las n redacciones superadas, cada una de
// las cuales da una version-obsoleta, y la instantánea que dejan esas lecturas.
func bloquesQueCambiaron(n int) ([]string, grafo.Instantanea) {
	superadas := make([]string, 0, n)
	lecturas := make([]lecturaDeEjemplo, 0, 2*n)

	for i := range n {
		parte := "a" + strconv.Itoa(100+i)
		superada := redaccionDeEjemplo{parte: parte, fechaVigencia: "20161002", cuerpo: "Articulo " + parte + "."}
		reciente := redaccionDeEjemplo{parte: parte, fechaVigencia: "20250101", cuerpo: "Articulo " + parte + " cambiado."}

		superadas = append(superadas, superada.id())
		lecturas = append(lecturas,
			lecturaDeEjemplo{redaccion: superada, fecha: lunes}, lecturaDeEjemplo{redaccion: reciente, fecha: martes})
	}

	return superadas, trasLeer(lecturas...)
}

// normasCaducadas son n normas, distintas de la de ejemplo, consultadas el
// lunes con una hora de vigencia: cada una da una fuente-caducada dentro de un
// mes.
func normasCaducadas(n int) []grafo.NodoDeInstantanea {
	normas := make([]grafo.NodoDeInstantanea, 0, n)

	for i := range n {
		numero := strconv.Itoa(20000 + i)
		normas = append(normas, conVigenciaDe(observado("eli/es/l/2015/10/02/"+numero, grafo.TipoNorma,
			map[string]any{grafo.DatoIdentificador: "BOE-A-2015-" + numero}, lunes), unaHora))
	}

	return normas
}

// porID ordena los nodos por su id, comparando bytes.
func porID(a, b grafo.NodoDeInstantanea) int {
	return strings.Compare(a.ID, b.ID)
}

// clasesEIDs son la clase y el id de cada hallazgo, en su orden; nil si no hay
// ninguno.
func clasesEIDs(hallazgos []grafo.Hallazgo) []string {
	var ids []string

	for _, hallazgo := range hallazgos {
		ids = append(ids, string(hallazgo.Clase)+" "+hallazgo.ID)
	}

	return ids
}

// TestComprobarCopiaElAmbito fija que Comprobar copia en su resultado la norma y
// los bloques del ámbito, tal como se pidieron y en su orden, y que aplica las
// reglas a todo lo que recibe: lo que entra en el ámbito lo decide la lectura
// acotada, no Comprobar (FR-002, FR-012; data-model §6). Los bloques del
// resultado son una copia de los del ámbito.
func TestComprobarCopiaElAmbito(t *testing.T) {
	t.Parallel()

	instantanea := conVigenciaEnTodos(trasLeer(leidaA, leidaB), unaSemana)

	todo, err := grafo.Comprobar(instantanea, grafo.Ambito{}, dentroDeUnMes)
	require.NoError(t, err)
	require.NotEmpty(t, todo.Hallazgos, "premisa: la instantanea da hallazgos")
	assert.Empty(t, todo.Norma, "sin norma, todo lo consultado")
	assert.Empty(t, todo.Bloques)

	for _, ambito := range []grafo.Ambito{
		{Norma: identificadorLPAC},
		{Norma: identificadorLPAC, Bloques: []string{"a21"}},
		{Norma: identificadorLPAC, Bloques: []string{"a99", "a21"}},
		{Norma: "BOE-A-2099-99999", Bloques: []string{"a1"}},
	} {
		pedidos := slices.Clone(ambito.Bloques)

		comprobacion, err := grafo.Comprobar(instantanea, ambito, dentroDeUnMes)
		require.NoError(t, err)

		assert.Equal(t, ambito.Norma, comprobacion.Norma)
		assert.Equal(t, pedidos, comprobacion.Bloques, "los bloques pedidos, en su orden")
		assert.Equal(t, todo.Hallazgos, comprobacion.Hallazgos, "las reglas, sobre todo lo que recibe")
		assert.Equal(t, [3]int{todo.VersionObsoleta, todo.FuenteCaducada, todo.Omitidos},
			[3]int{comprobacion.VersionObsoleta, comprobacion.FuenteCaducada, comprobacion.Omitidos})

		if len(ambito.Bloques) > 0 {
			ambito.Bloques[0] = "cambiado"
			assert.Equal(t, pedidos, comprobacion.Bloques, "los bloques del resultado son una copia de los del ambito")
		}
	}
}
