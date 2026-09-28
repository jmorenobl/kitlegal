package grafo_test

import (
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

// deVersiones es el grafo del bloque a21, observado el lunes y sin vigencia,
// con estas versiones, cada una unida a él por eli:has_version.
func deVersiones(versiones ...versionDeEjemplo) grafo.Instantanea {
	instantanea := grafo.Instantanea{Nodos: []grafo.NodoDeInstantanea{bloqueDeLaLPAC(idBloque, "a21", lunes)}}

	for _, version := range versiones {
		instantanea.Nodos = append(instantanea.Nodos, version.nodo())
		instantanea.Aristas = append(instantanea.Aristas, arista(idBloque, grafo.RelacionTieneVersion, version.id()))
	}

	return instantanea
}

// deNodos es la instantánea de estos nodos, sin aristas.
func deNodos(nodos ...grafo.NodoDeInstantanea) grafo.Instantanea {
	return grafo.Instantanea{Nodos: nodos}
}

// conMas es la misma instantánea con más nodos y más aristas.
func conMas(instantanea grafo.Instantanea, nodos []grafo.NodoDeInstantanea, aristas ...schema.Arista) grafo.Instantanea {
	return grafo.Instantanea{
		Nodos:   slices.Concat(instantanea.Nodos, nodos),
		Aristas: slices.Concat(instantanea.Aristas, aristas),
	}
}

// invertida es la misma instantánea con sus nodos y sus aristas en el orden
// inverso.
func invertida(instantanea grafo.Instantanea) grafo.Instantanea {
	nodos := slices.Clone(instantanea.Nodos)
	aristas := slices.Clone(instantanea.Aristas)

	slices.Reverse(nodos)
	slices.Reverse(aristas)

	return grafo.Instantanea{Nodos: nodos, Aristas: aristas}
}

// obsoleta es el hallazgo, sin la explicación, de la versión superada con la
// procedencia y la fecha de vigencia de la más reciente (FR-063).
func obsoleta(superada, reciente versionDeEjemplo) grafo.Hallazgo {
	return grafo.Hallazgo{
		Clase:                 grafo.ClaseVersionObsoleta,
		ID:                    superada.id(),
		Procedencia:           reciente.nodo().UltimaObservacion,
		FechaVigencia:         superada.fechaVigencia,
		FechaVigenciaReciente: reciente.fechaVigencia,
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

// TestComprobar fija las dos reglas de `graph check` y su orden (FR-060 a
// FR-064, FR-066, FR-067; data-model §6; research.md D15, D34): una
// version-obsoleta por cada versión con otra de su bloque de fecha de vigencia
// estrictamente posterior, entre las fechas de vigencia válidas, con la
// procedencia y la fecha de la más reciente; una fuente-caducada por cada nodo
// cuya consulta ha caducado estrictamente antes del instante de la
// comprobación; las dos sobre el mismo nodo; una lista nunca nula, ordenada por
// clase y por id, que no depende del orden de la instantánea ni cambia al
// repetir la comprobación; y un error, sin hallazgos, con una instantánea que
// ninguna lectura de world.db da.
func TestComprobar(t *testing.T) {
	t.Parallel()

	for _, caso := range slices.Concat(casosDeVersiones(), casosDeFechasDeVigencia(), casosDeCaducidad()) {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			exigirComprobacion(t, caso)
		})
	}

	t.Run("instantaneas que ninguna lectura da", probarInstantaneasImposibles)
}

// casosDeVersiones son los de version-obsoleta con fechas de vigencia válidas.
func casosDeVersiones() []casoDeComprobacion {
	v2015 := versionDeEjemplo{fechaVigencia: "20150101", letra: "a", consulta: lunes}
	v2016 := versionDeEjemplo{fechaVigencia: "20160101", letra: "b", consulta: martes}
	v2017 := versionDeEjemplo{fechaVigencia: "20170101", letra: "c", consulta: miercoles}
	// v2016Antes y v2016Despues tienen la misma fecha de vigencia y el mismo
	// instante de consulta, con textos distintos: el de v2016Despues es mayor y
	// su id también.
	v2016Antes := versionDeEjemplo{fechaVigencia: "20160101", letra: "b", consulta: lunes}
	v2016Despues := versionDeEjemplo{fechaVigencia: "20160101", letra: "c", consulta: lunesEnMadrid}
	// Y v2016EnMadrid y v2016Utc, al revés: el id menor lleva el texto mayor.
	v2016EnMadrid := versionDeEjemplo{fechaVigencia: "20160101", letra: "b", consulta: lunesEnMadrid}
	v2016Utc := versionDeEjemplo{fechaVigencia: "20160101", letra: "c", consulta: lunes}
	// v2016Martes es de la misma fecha de vigencia con una consulta posterior
	// y el id mayor.
	v2016Martes := versionDeEjemplo{fechaVigencia: "20160101", letra: "c", consulta: martes}
	// v2015Martes es la más antigua observada la última.
	v2015Martes := versionDeEjemplo{fechaVigencia: "20150101", letra: "a", consulta: martes}
	v2016Lunes := versionDeEjemplo{fechaVigencia: "20160101", letra: "b", consulta: lunes}
	ninguno := []grafo.Hallazgo{}

	return []casoDeComprobacion{
		{"un grafo vacio", grafo.Instantanea{}, elMartes, ninguno},
		{"una sola version", deVersiones(v2015), elMartes, ninguno},
		{
			"una vez por cada version superada, en orden de id",
			deVersiones(v2017, v2015, v2016), elMartes,
			[]grafo.Hallazgo{obsoleta(v2015, v2017), obsoleta(v2016, v2017)},
		},
		{
			"empate en la fecha mas alta: un hallazgo, sobre la anterior",
			deVersiones(v2015, v2016Antes, v2016Despues), elMartes,
			[]grafo.Hallazgo{obsoleta(v2015, v2016Antes)},
		},
		{"misma fecha y distinta huella", deVersiones(v2016Antes, v2016Martes), elMartes, ninguno},
		{
			"la mas reciente, a igual fecha, por la ultima observacion posterior",
			deVersiones(v2015, v2016Antes, v2016Martes), elMartes,
			[]grafo.Hallazgo{obsoleta(v2015, v2016Martes)},
		},
		{
			"la mas reciente, a igual instante, por el id menor y no por el texto",
			deVersiones(v2016Utc, v2015, v2016EnMadrid), elMartes,
			[]grafo.Hallazgo{obsoleta(v2015, v2016EnMadrid)},
		},
		{
			"la fecha de vigencia manda sobre la observacion",
			deVersiones(v2015Martes, v2016Lunes), elMartes,
			[]grafo.Hallazgo{obsoleta(v2015Martes, v2016Lunes)},
		},
		{
			"una version de dos bloques, una vez, por el de id menor",
			conMas(deVersiones(v2015, v2016), []grafo.NodoDeInstantanea{bloqueDeLaLPAC(idBloqueA22, "a22", lunes), v2017.nodo()},
				arista(idBloqueA22, grafo.RelacionTieneVersion, v2015.id()),
				arista(idBloqueA22, grafo.RelacionTieneVersion, v2017.id())),
			elMartes,
			[]grafo.Hallazgo{obsoleta(v2015, v2016)},
		},
	}
}

// casosDeFechasDeVigencia son los de las versiones que no se comparan: una
// fecha de vigencia que no es válida no recibe ni provoca un hallazgo, y solo
// se comparan las versiones de un mismo Bloque unidas a él por eli:has_version
// (FR-063, FR-064; research.md D34). En cada uno, la única comparación posible
// es la de 2015 con 2017.
func casosDeFechasDeVigencia() []casoDeComprobacion {
	v2015 := versionDeEjemplo{fechaVigencia: "20150101", letra: "a", consulta: lunes}
	v2017 := versionDeEjemplo{fechaVigencia: "20170101", letra: "c", consulta: lunes}
	unica := []grafo.Hallazgo{obsoleta(v2015, v2017)}
	ninguno := []grafo.Hallazgo{}

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

	casos := make([]casoDeComprobacion, 0, len(noValidas)+7)
	for _, noValida := range noValidas {
		otra := versionDeEjemplo{fechaVigencia: noValida.fecha, letra: "z", consulta: lunes}
		casos = append(casos, casoDeComprobacion{
			"una fecha de vigencia " + noValida.nombre, deVersiones(v2015, otra, v2017), elMartes, unica,
		})
	}

	sinFecha := observado(idBloque+"@sin-fecha", grafo.TipoBloqueVersion, map[string]any{}, lunes)
	numerica := observado(idBloque+"@numerica", grafo.TipoBloqueVersion,
		map[string]any{grafo.DatoFechaVigencia: float64(20180101)}, lunes)
	w2016 := observado(idBloqueA22+"@20160101", grafo.TipoBloqueVersion,
		map[string]any{grafo.DatoFechaVigencia: "20160101"}, lunes)
	noVersion := observado(idMunicipio, grafo.TipoMunicipio, map[string]any{grafo.DatoFechaVigencia: "20160101"}, lunes)
	v2016 := versionDeEjemplo{fechaVigencia: "20160101", letra: "b", consulta: lunes}
	solo2015 := deVersiones(v2015)

	return append(casos,
		casoDeComprobacion{
			"una version sin fecha de vigencia",
			conMas(deVersiones(v2015, v2017), []grafo.NodoDeInstantanea{sinFecha},
				arista(idBloque, grafo.RelacionTieneVersion, sinFecha.ID)),
			elMartes, unica,
		},
		casoDeComprobacion{
			"una fecha de vigencia que no es un texto",
			conMas(deVersiones(v2015, v2017), []grafo.NodoDeInstantanea{numerica},
				arista(idBloque, grafo.RelacionTieneVersion, numerica.ID)),
			elMartes, unica,
		},
		casoDeComprobacion{
			"versiones de dos bloques",
			conMas(solo2015, []grafo.NodoDeInstantanea{bloqueDeLaLPAC(idBloqueA22, "a22", lunes), w2016},
				arista(idBloqueA22, grafo.RelacionTieneVersion, w2016.ID)),
			elMartes, ninguno,
		},
		casoDeComprobacion{
			"una arista que no sale de un Bloque",
			conMas(solo2015, []grafo.NodoDeInstantanea{normaLPAC(lunes), v2016.nodo()},
				arista(idNorma, grafo.RelacionTieneVersion, v2016.id())),
			elMartes, ninguno,
		},
		casoDeComprobacion{
			"una arista que no llega a una BloqueVersion",
			conMas(solo2015, []grafo.NodoDeInstantanea{noVersion},
				arista(idBloque, grafo.RelacionTieneVersion, idMunicipio)),
			elMartes, ninguno,
		},
		casoDeComprobacion{
			"una arista de otra relacion",
			conMas(solo2015, []grafo.NodoDeInstantanea{v2016.nodo()},
				arista(idBloque, grafo.RelacionTieneParte, v2016.id())),
			elMartes, ninguno,
		},
		casoDeComprobacion{
			"una arista hacia un nodo que no esta",
			conMas(solo2015, nil, arista(idBloque, grafo.RelacionTieneVersion, v2016.id())),
			elMartes, ninguno,
		},
	)
}

// casosDeCaducidad son los de fuente-caducada, y los de las dos clases sobre
// el mismo nodo (FR-066).
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
	v2016 := versionDeEjemplo{fechaVigencia: "20160101", letra: "b", consulta: lunes}
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
		{
			"las dos clases sobre el mismo nodo, por clase y por id",
			deVersiones(v2016.conVigencia(unaSemana), v2015), dentroDeUnMes,
			[]grafo.Hallazgo{
				caducada(v2015.nodo()), caducada(v2016.conVigencia(unaSemana).nodo()), obsoleta(v2015, v2016),
			},
		},
	}
}

// exigirComprobacion exige que la comprobación del caso dé sus hallazgos, cada
// uno con su explicación, en una lista que no es nula, y la misma lista al
// repetirla y con la instantánea en el orden inverso (FR-060, FR-062).
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

// probarInstantaneasImposibles fija que una instantánea con lo que ninguna
// entrega guarda —una fecha de consulta que no es RFC 3339, una vigencia
// negativa o con fracción de segundo— da un error y ningún hallazgo: es un
// defecto de lo guardado, no un grafo que comprobar.
func probarInstantaneasImposibles(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre  string
		nodo    grafo.NodoDeInstantanea
		mensaje string
	}{
		{"una fecha de consulta que no es RFC 3339", normaLPAC(fechaImposible), mensajeFechaImposible},
		{
			"una vigencia negativa", conVigenciaDe(normaLPAC(lunes), -unaHora),
			"la vigencia guardada -1h0m0s no es un n\xc3\xbamero entero de segundos positivo",
		},
		{
			"una vigencia con fraccion de segundo", conVigenciaDe(normaLPAC(lunes), 1500*time.Millisecond),
			"la vigencia guardada 1.5s no es un n\xc3\xbamero entero de segundos positivo",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			hallazgos, err := grafo.Comprobar(deNodos(caso.nodo), elMartes)
			require.ErrorContains(t, err, caso.mensaje)
			assert.Nil(t, hallazgos)
		})
	}
}
