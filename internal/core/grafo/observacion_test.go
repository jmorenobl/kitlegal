package grafo_test

import (
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/grafo"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Las fuentes y las direcciones de las observaciones de ejemplo. fuenteCorta
// es menor que fuenteBOE comparando bytes, porque es su prefijo, y la url del
// bloque a21 es menor que la del a22.
const (
	fuenteBOE   = "boe.legislacion-consolidada"
	fuenteCorta = "boe"
	urlA21      = "https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565/texto/bloque/a21"
	urlA22      = "https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565/texto/bloque/a22"
)

// Las fechas de consulta de ejemplo, cada una como la escribiría un sobre.
const (
	domingo = "2026-09-27T12:00:00Z"
	lunes   = "2026-09-28T12:00:00Z"
	// lunesEnMadrid es el mismo instante que lunes con un texto mayor.
	lunesEnMadrid = "2026-09-28T14:00:00+02:00"
	// lunesConFraccion es el mismo instante que lunes con un texto menor: el
	// punto va antes que la Z.
	lunesConFraccion = "2026-09-28T12:00:00.000Z"
	// lunesAntesEnMadrid es una hora antes que lunes, con un texto mayor.
	lunesAntesEnMadrid = "2026-09-28T13:00:00+02:00"
	// lunesYMedio es medio segundo después que lunes, con un texto menor.
	lunesYMedio = "2026-09-28T12:00:00.5Z"
	martes      = "2026-09-29T12:00:00Z"
	miercoles   = "2026-09-30T12:00:00Z"
	jueves      = "2026-10-01T12:00:00Z"
	// fechaImposible no es RFC 3339.
	fechaImposible = "28/09/2026"
	// mensajeFechaImposible es como empieza el error de una fecha guardada o
	// llegada que no es RFC 3339.
	mensajeFechaImposible = `la fecha de consulta "28/09/2026" no es RFC 3339`
)

// Las vigencias de ejemplo; sin vigencia es cero, que la fuente no la declara.
const (
	unaHora   = time.Hour
	unaSemana = 7 * 24 * time.Hour
)

// Datos canónicos de ejemplo: datosMenores es menor que datosMayores
// comparando bytes.
const (
	datosMenores = `{"bloque":"a21"}`
	datosMayores = `{"bloque":"a22"}`
)

// vista es una observación de ejemplo: la procedencia del lote que la trae, la
// vigencia que declara y, si es de un nodo, sus datos canónicos.
type vista struct {
	fecha    string
	url      string
	fuente   string
	vigencia time.Duration
	datos    string
}

// ver es la observación de url en fecha, de la fuente del BOE, sin vigencia y
// con los datos menores.
func ver(fecha, url string) vista {
	return vista{fecha: fecha, url: url, fuente: fuenteBOE, datos: datosMenores}
}

// de es la misma observación de otra fuente.
func (v vista) de(fuente string) vista {
	v.fuente = fuente

	return v
}

// conVigencia es la misma observación con otra vigencia.
func (v vista) conVigencia(vigencia time.Duration) vista {
	v.vigencia = vigencia

	return v
}

// conDatos es la misma observación con otros datos.
func (v vista) conDatos(datos string) vista {
	v.datos = datos

	return v
}

// procedencia es la de la observación, como la guarda el grafo.
func (v vista) procedencia() grafo.Procedencia {
	return grafo.Procedencia{Fuente: v.fuente, URL: v.url, FechaConsulta: v.fecha}
}

// historia es lo que el grafo recuerda de un nodo, una arista o un texto
// observados una o más veces: la primera y la última observación. De un texto
// solo se guarda la procedencia de la primera, la más antigua.
type historia struct {
	primera vista
	ultima  vista
}

// sola es la historia de lo que se ha observado una sola vez: su observación
// es la primera y la última.
func sola(v vista) historia {
	return historia{primera: v, ultima: v}
}

// nodo es el registro del bloque a21 con esta historia.
func (h historia) nodo() grafo.RegistroDeNodo {
	return grafo.RegistroDeNodo{
		ID:                 idBloque,
		Tipo:               grafo.TipoBloque,
		Datos:              h.ultima.datos,
		PrimeraObservacion: h.primera.procedencia(),
		UltimaObservacion:  h.ultima.procedencia(),
		Vigencia:           h.ultima.vigencia,
	}
}

// arista es el registro de la arista de la norma a su bloque a21 con esta
// historia; los datos no son de una arista.
func (h historia) arista() grafo.RegistroDeArista {
	return grafo.RegistroDeArista{
		Origen:             idNorma,
		Relacion:           grafo.RelacionTieneParte,
		Destino:            idBloque,
		PrimeraObservacion: h.primera.procedencia(),
		UltimaObservacion:  h.ultima.procedencia(),
		Vigencia:           h.ultima.vigencia,
	}
}

// texto es el registro del texto del artículo 21 con la primera observación
// de esta historia como procedencia.
func (h historia) texto() grafo.RegistroDeTexto {
	return grafo.RegistroDeTexto{Huella: huellaDe(cuerpoA21), Cuerpo: cuerpoA21, Procedencia: h.primera.procedencia()}
}

// casoDeFusion es lo que saben dos registros de lo mismo y lo que tiene que
// saber su fusión, llegue antes el que llegue.
type casoDeFusion struct {
	nombre   string
	una      historia
	otra     historia
	esperada historia
}

// casosDeFusion son los de FR-023 que valen igual para nodos, aristas y
// textos: el instante decide la primera y la última, y a igual instante, el
// desempate por url, fuente, texto de la fecha y vigencia, cada criterio antes
// que el siguiente. La que gana lleva los datos mayores, para que se vea que en
// un nodo los datos no deciden antes que ningún otro criterio. Un texto guarda
// la procedencia de la primera.
func casosDeFusion() []casoDeFusion {
	return []casoDeFusion{
		{
			"fuera de orden, la primera es la mas antigua y la ultima la mas reciente",
			sola(ver(martes, urlA22).conDatos(datosMayores)), sola(ver(lunes, urlA21)),
			historia{ver(lunes, urlA21), ver(martes, urlA22).conDatos(datosMayores)},
		},
		{
			"a igual instante, la url menor",
			sola(ver(lunes, urlA21).conDatos(datosMayores)), sola(ver(lunes, urlA22)),
			sola(ver(lunes, urlA21).conDatos(datosMayores)),
		},
		{
			"a igual url, la fuente menor",
			sola(ver(lunes, urlA21).de(fuenteCorta).conDatos(datosMayores)), sola(ver(lunes, urlA21)),
			sola(ver(lunes, urlA21).de(fuenteCorta).conDatos(datosMayores)),
		},
		{
			"a igual fuente, el texto menor de la fecha",
			sola(ver(lunes, urlA21).conDatos(datosMayores)), sola(ver(lunesEnMadrid, urlA21)),
			sola(ver(lunes, urlA21).conDatos(datosMayores)),
		},
		{
			"el mismo instante con fraccion de segundo, por su texto",
			sola(ver(lunesConFraccion, urlA21).conDatos(datosMayores)), sola(ver(lunes, urlA21)),
			sola(ver(lunesConFraccion, urlA21).conDatos(datosMayores)),
		},
		{
			"a igual fecha, sin vigencia antes que con ella",
			sola(ver(lunes, urlA21).conDatos(datosMayores)), sola(ver(lunes, urlA21).conVigencia(unaSemana)),
			sola(ver(lunes, urlA21).conDatos(datosMayores)),
		},
		{
			"entre dos vigencias, la menor",
			sola(ver(lunes, urlA21).conVigencia(unaHora).conDatos(datosMayores)),
			sola(ver(lunes, urlA21).conVigencia(unaSemana)),
			sola(ver(lunes, urlA21).conVigencia(unaHora).conDatos(datosMayores)),
		},
		{
			"el instante decide antes que la url",
			sola(ver(lunes, urlA22)), sola(ver(martes, urlA21).conDatos(datosMayores)),
			historia{ver(lunes, urlA22), ver(martes, urlA21).conDatos(datosMayores)},
		},
		{
			"el instante decide antes que el texto de la fecha, con fraccion de segundo",
			sola(ver(lunesYMedio, urlA22)), sola(ver(lunes, urlA21).conDatos(datosMayores)),
			historia{ver(lunes, urlA21).conDatos(datosMayores), ver(lunesYMedio, urlA22)},
		},
		{
			"el instante decide antes que el texto de la fecha, con otro desplazamiento",
			sola(ver(lunesAntesEnMadrid, urlA22)), sola(ver(lunes, urlA21).conDatos(datosMayores)),
			historia{ver(lunesAntesEnMadrid, urlA22), ver(lunes, urlA21).conDatos(datosMayores)},
		},
		{
			"la url decide antes que la fuente",
			sola(ver(lunes, urlA21).conDatos(datosMayores)), sola(ver(lunes, urlA22).de(fuenteCorta)),
			sola(ver(lunes, urlA21).conDatos(datosMayores)),
		},
		{
			"la fuente decide antes que el texto de la fecha",
			sola(ver(lunesEnMadrid, urlA21).de(fuenteCorta).conDatos(datosMayores)), sola(ver(lunes, urlA21)),
			sola(ver(lunesEnMadrid, urlA21).de(fuenteCorta).conDatos(datosMayores)),
		},
		{
			"el texto de la fecha decide antes que la vigencia",
			sola(ver(lunes, urlA21).conVigencia(unaSemana).conDatos(datosMayores)), sola(ver(lunesEnMadrid, urlA21)),
			sola(ver(lunes, urlA21).conVigencia(unaSemana).conDatos(datosMayores)),
		},
		{
			"una observacion identica",
			sola(ver(lunes, urlA21).conVigencia(unaSemana)), sola(ver(lunes, urlA21).conVigencia(unaSemana)),
			sola(ver(lunes, urlA21).conVigencia(unaSemana)),
		},
		{
			"una posterior a la ultima no mueve la primera",
			historia{ver(lunes, urlA21), ver(martes, urlA21)},
			sola(ver(miercoles, urlA22).conVigencia(unaHora)),
			historia{ver(lunes, urlA21), ver(miercoles, urlA22).conVigencia(unaHora)},
		},
		{
			"una anterior a la primera no mueve la ultima",
			historia{ver(lunes, urlA21), ver(miercoles, urlA21).conVigencia(unaHora)},
			sola(ver(domingo, urlA22)),
			historia{ver(domingo, urlA22), ver(miercoles, urlA21).conVigencia(unaHora)},
		},
		{
			"una entre la primera y la ultima no cambia nada",
			historia{ver(lunes, urlA21), ver(miercoles, urlA22)},
			sola(ver(martes, urlA21)),
			historia{ver(lunes, urlA21), ver(miercoles, urlA22)},
		},
		{
			"a igual instante que la primera, la url menor la sustituye",
			historia{ver(lunes, urlA22), ver(martes, urlA22)},
			sola(ver(lunes, urlA21)),
			historia{ver(lunes, urlA21), ver(martes, urlA22)},
		},
		{
			"dos historias",
			historia{ver(lunes, urlA21), ver(miercoles, urlA21)},
			historia{ver(domingo, urlA22), ver(jueves, urlA22)},
			historia{ver(domingo, urlA22), ver(jueves, urlA22)},
		},
	}
}

// casosDeFusionDeDatos son los de FR-023 que solo tiene un nodo: sus datos
// identificativos, que son los de su última observación, enteros y sin
// mezclar, y el último criterio del desempate.
func casosDeFusionDeDatos() []casoDeFusion {
	datosDeLunes := `{"bloque":"a21","fecha_vigencia":"20161002"}`
	datosDeMartes := `{"hash_texto":"sha256:00"}`

	return []casoDeFusion{
		{
			"a igual vigencia, los datos canonicos menores",
			sola(ver(lunes, urlA21).conVigencia(unaHora).conDatos(datosMayores)),
			sola(ver(lunes, urlA21).conVigencia(unaHora).conDatos(datosMenores)),
			sola(ver(lunes, urlA21).conVigencia(unaHora).conDatos(datosMenores)),
		},
		{
			"la vigencia decide antes que los datos",
			sola(ver(lunes, urlA21).conDatos(datosMayores)), sola(ver(lunes, urlA21).conVigencia(unaHora)),
			sola(ver(lunes, urlA21).conDatos(datosMayores)),
		},
		{
			"los datos de la ultima, enteros y sin mezclar",
			sola(ver(lunes, urlA21).conDatos(datosDeLunes)), sola(ver(martes, urlA21).conDatos(datosDeMartes)),
			historia{ver(lunes, urlA21).conDatos(datosDeLunes), ver(martes, urlA21).conDatos(datosDeMartes)},
		},
	}
}

// historiaImposible es un registro con una fecha que no es RFC 3339, en su
// primera o en su última observación, guardado o llegando.
type historiaImposible struct {
	nombre   string
	guardada historia
	llegada  historia
	enUltima bool
}

// historiasImposibles son las cuatro posiciones de una fecha que no es RFC
// 3339: un defecto de lo guardado o de quien llama, nunca un Rechazo del lote.
// Un texto solo tiene la primera.
func historiasImposibles() []historiaImposible {
	valida := ver(lunes, urlA21)
	imposible := ver(fechaImposible, urlA21)

	return []historiaImposible{
		{"una fecha imposible en la primera guardada", historia{imposible, valida}, sola(valida), false},
		{"una fecha imposible en la ultima guardada", historia{valida, imposible}, sola(valida), true},
		{"una fecha imposible en la primera que llega", sola(valida), historia{imposible, valida}, false},
		{"una fecha imposible en la ultima que llega", sola(valida), historia{valida, imposible}, true},
	}
}

// TestFusionarNodo fija la fusión de lo guardado de un nodo con lo que llega
// de él (FR-022, FR-023; data-model §4.1; research.md D13): la primera
// observación nunca avanza, la última nunca retrocede y lleva enteros los datos
// y la vigencia de la observación que la sostiene, con el desempate de FR-023,
// y el resultado es el mismo en los dos órdenes de llegada y no cambia al
// repetir una observación. Un tipo distinto rechaza el lote.
func TestFusionarNodo(t *testing.T) {
	t.Parallel()

	for _, caso := range slices.Concat(casosDeFusion(), casosDeFusionDeDatos()) {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			exigirFusion(t, grafo.FusionarNodo, caso.una.nodo(), caso.otra.nodo(), caso.esperada.nodo())
		})
	}

	t.Run("otro tipo", probarNodoDeOtroTipo)
	t.Run("defectos", probarNodosQueNoSeFusionan)
}

// probarNodoDeOtroTipo fija que un nodo que llega con otro tipo que el
// guardado rechaza el lote entero (FR-024) con un Rechazo que nombra el nodo
// que llega, por su id y su tipo, salvo si uno de los dos es una Persona, que
// se nombra por su tipo: el mensaje nunca repite el id de una Persona.
func probarNodoDeOtroTipo(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre   string
		id       string
		guardado string
		llegado  string
		nombrado string
		mensaje  string
	}{
		{
			"una Norma que llega como Bloque", idNorma, grafo.TipoNorma, grafo.TipoBloque, grafo.TipoBloque,
			`el nodo "eli/es/l/2015/10/01/39": el grafo ya lo tiene con el tipo "Norma" y el lote le da el tipo "Bloque"`,
		},
		{
			"una Persona guardada", "persona:ana", grafo.TipoPersona, grafo.TipoNorma, grafo.TipoPersona,
			`el nodo de tipo "Persona": el grafo ya lo tiene con el tipo "Persona" y el lote le da el tipo "Norma"`,
		},
		{
			"una Persona que llega", "persona:ana", grafo.TipoNorma, grafo.TipoPersona, grafo.TipoPersona,
			`el nodo de tipo "Persona": el grafo ya lo tiene con el tipo "Norma" y el lote le da el tipo "Persona"`,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			guardado := sola(ver(lunes, urlA21)).nodo()
			guardado.ID, guardado.Tipo = caso.id, caso.guardado
			llegado := sola(ver(martes, urlA21)).nodo()
			llegado.ID, llegado.Tipo = caso.id, caso.llegado

			fusionado, err := grafo.FusionarNodo(guardado, llegado)
			exigirRechazoDelLote(t, err, schema.Nodo{ID: caso.id, Tipo: caso.nombrado}, caso.mensaje)
			assert.Zero(t, fusionado)

			if caso.nombrado == grafo.TipoPersona {
				assert.NotContains(t, err.Error(), caso.id, "el mensaje nunca repite el id de una Persona")
			}
		})
	}
}

// probarNodosQueNoSeFusionan fija los defectos de quien llama o de lo
// guardado, que no son un Rechazo del lote: dos nodos distintos, sin nombrar
// sus ids, y una fecha que no es RFC 3339 en cualquiera de las cuatro
// posiciones.
func probarNodosQueNoSeFusionan(t *testing.T) {
	t.Parallel()

	guardado := sola(ver(lunes, urlA21)).nodo()
	otro := guardado
	otro.ID = idNorma
	exigirDefecto(t, grafo.FusionarNodo, guardado, otro, "no se pueden fusionar las observaciones de dos nodos distintos")

	_, err := grafo.FusionarNodo(guardado, otro)
	require.Error(t, err)
	assert.NotContains(t, err.Error(), guardado.ID, "no nombra los ids, que pueden ser de una Persona")
	assert.NotContains(t, err.Error(), otro.ID, "no nombra los ids, que pueden ser de una Persona")

	for _, caso := range historiasImposibles() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			exigirDefecto(t, grafo.FusionarNodo, caso.guardada.nodo(), caso.llegada.nodo(), mensajeFechaImposible)
		})
	}
}

// TestFusionarArista fija la fusión de lo guardado de una arista con lo que
// llega de ella (FR-022, FR-023; data-model §4.1; research.md D13): la primera
// observación nunca avanza, la última nunca retrocede y lleva la vigencia de la
// observación que la sostiene, con el desempate de FR-023, y el resultado es el
// mismo en los dos órdenes de llegada y no cambia al repetir una observación.
func TestFusionarArista(t *testing.T) {
	t.Parallel()

	for _, caso := range casosDeFusion() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			exigirFusion(t, grafo.FusionarArista, caso.una.arista(), caso.otra.arista(), caso.esperada.arista())
		})
	}

	t.Run("defectos", probarAristasQueNoSeFusionan)
}

// probarAristasQueNoSeFusionan fija los defectos de quien llama o de lo
// guardado, que no son un Rechazo del lote: dos aristas que difieren en su
// origen, su relación o su destino, y una fecha que no es RFC 3339 en
// cualquiera de las cuatro posiciones.
func probarAristasQueNoSeFusionan(t *testing.T) {
	t.Parallel()

	guardada := sola(ver(lunes, urlA21)).arista()
	distintas := []func(*grafo.RegistroDeArista){
		func(arista *grafo.RegistroDeArista) { arista.Origen = idMunicipio },
		func(arista *grafo.RegistroDeArista) { arista.Relacion = grafo.RelacionTieneVersion },
		func(arista *grafo.RegistroDeArista) { arista.Destino = idMunicipio },
	}

	for _, cambiar := range distintas {
		llegada := guardada
		cambiar(&llegada)
		exigirDefecto(t, grafo.FusionarArista, guardada, llegada,
			"no se pueden fusionar las observaciones de dos aristas distintas")
	}

	for _, caso := range historiasImposibles() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			exigirDefecto(t, grafo.FusionarArista, caso.guardada.arista(), caso.llegada.arista(), mensajeFechaImposible)
		})
	}
}

// TestFusionarTexto fija la fusión de lo guardado de un texto con lo que llega
// de él (FR-022, FR-023; data-model §4.1; research.md D13): un solo texto por
// huella, con el mismo cuerpo y la procedencia de su observación más antigua,
// que solo cambia si la que llega es estrictamente anterior o del mismo
// instante y gana el desempate por url, fuente y texto de la fecha; el mismo
// resultado en los dos órdenes de llegada. Una huella guardada con otro cuerpo
// rechaza el lote.
func TestFusionarTexto(t *testing.T) {
	t.Parallel()

	for _, caso := range casosDeFusion() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			exigirFusion(t, grafo.FusionarTexto, caso.una.texto(), caso.otra.texto(), caso.esperada.texto())
		})
	}

	t.Run("otro cuerpo con la misma huella", probarTextoConOtroCuerpo)
	t.Run("defectos", probarTextosQueNoSeFusionan)
}

// probarTextoConOtroCuerpo fija que una huella ya guardada con otro cuerpo
// rechaza el lote entero (FR-024) con un Rechazo que nombra el texto que llega
// por su huella y no repite ninguno de los dos cuerpos.
func probarTextoConOtroCuerpo(t *testing.T) {
	t.Parallel()

	guardado := sola(ver(lunes, urlA21)).texto()
	llegado := sola(ver(martes, urlA21)).texto()
	llegado.Cuerpo = "Articulo 21. Otro cuerpo."

	fusionado, err := grafo.FusionarTexto(guardado, llegado)
	exigirRechazoDelLote(t, err, schema.Texto{Huella: llegado.Huella, Cuerpo: llegado.Cuerpo},
		`el texto "`+llegado.Huella+`": el grafo ya guarda otro cuerpo con esa huella`)
	assert.Zero(t, fusionado)
	assert.NotContains(t, err.Error(), guardado.Cuerpo, "nunca repite un cuerpo")
	assert.NotContains(t, err.Error(), llegado.Cuerpo, "nunca repite un cuerpo")
}

// probarTextosQueNoSeFusionan fija los defectos de quien llama o de lo
// guardado, que no son un Rechazo del lote: dos huellas distintas y una fecha
// que no es RFC 3339 en la procedencia guardada o en la que llega.
func probarTextosQueNoSeFusionan(t *testing.T) {
	t.Parallel()

	guardado := sola(ver(lunes, urlA21)).texto()
	otro := guardado
	otro.Huella = huellaDe("otro cuerpo")
	otro.Cuerpo = "otro cuerpo"
	exigirDefecto(t, grafo.FusionarTexto, guardado, otro,
		"no se pueden fusionar las observaciones de dos textos distintos")

	for _, caso := range historiasImposibles() {
		if caso.enUltima {
			continue
		}

		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			exigirDefecto(t, grafo.FusionarTexto, caso.guardada.texto(), caso.llegada.texto(), mensajeFechaImposible)
		})
	}
}

// exigirFusion comprueba que fusionar da lo esperado en los dos órdenes de
// llegada —la una guardada y la otra llegando, y al revés— y que lo esperado
// ya no cambia al fusionarlo con cualquiera de las dos ni consigo mismo: una
// observación ya vista no cambia nada (FR-022, FR-023).
func exigirFusion[R comparable](t *testing.T, fusionar func(guardado, llegado R) (R, error), una, otra, esperado R) {
	t.Helper()

	ordenes := []struct {
		nombre   string
		guardado R
		llegado  R
	}{
		{"la una guardada y la otra llegando", una, otra},
		{"la otra guardada y la una llegando", otra, una},
		{"la fusion guardada y la una llegando", esperado, una},
		{"la fusion guardada y la otra llegando", esperado, otra},
		{"la fusion guardada y ella misma llegando", esperado, esperado},
	}

	for _, orden := range ordenes {
		fusionado, err := fusionar(orden.guardado, orden.llegado)
		require.NoError(t, err, orden.nombre)
		assert.Equal(t, esperado, fusionado, orden.nombre)
	}
}

// exigirDefecto comprueba que fusionar falla con un error que contiene el
// texto dado y que no es un Rechazo del lote, sino un defecto de quien llama o
// de lo guardado, y lo devuelve.
func exigirDefecto[R any](t *testing.T, fusionar func(guardado, llegado R) (R, error), guardado, llegado R, texto string) {
	t.Helper()

	fusionado, err := fusionar(guardado, llegado)
	require.ErrorContains(t, err, texto)
	assert.Zero(t, fusionado)

	var rechazo *grafo.Rechazo
	assert.NotErrorAs(t, err, &rechazo, "no es un Rechazo del lote")
}
