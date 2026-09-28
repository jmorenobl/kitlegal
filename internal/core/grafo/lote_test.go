package grafo_test

import (
	"crypto/sha256"
	"encoding/hex"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core"
	"github.com/jmorenobl/kitlegal/internal/core/grafo"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Los ids del lote de ejemplo: una norma, uno de sus bloques, un municipio y
// su ayuntamiento, cuyo DIR3 tiene la forma de un NIF de persona jurídica.
const (
	idNorma     = "eli/es/l/2015/10/01/39"
	idBloque    = "eli/es/l/2015/10/01/39#a21"
	idMunicipio = "ine:28074"
	idOrgano    = "L01280745"
	cuerpoA21   = "Artículo 21. Obligación de resolver."
)

// Los motivos que ValidarLote da sin necesitar la base (FR-024, FR-025;
// contracts/almacen-world-db.md §5).
const (
	motivoSinID      = "no lleva id"
	motivoSinTipo    = "no lleva tipo"
	motivoSinOrigen  = "no lleva origen"
	motivoSinRel     = "no lleva relación"
	motivoSinDestino = "no lleva destino"
	motivoHuella     = "su huella no es la de su cuerpo: «sha256:» y los 64 hexadecimales en minúscula de su SHA-256"
)

// huellaDe es la huella de un texto como la exige FR-024: «sha256:» y los 64
// hexadecimales en minúscula de la SHA-256 de los bytes del cuerpo, tal cual.
func huellaDe(cuerpo string) string {
	suma := sha256.Sum256([]byte(cuerpo))

	return "sha256:" + hex.EncodeToString(suma[:])
}

// loteDeEjemplo es un lote que ValidarLote y ValidarContraGrafoVacio aceptan,
// con cada clase de operación. Cada llamada construye uno nuevo, de modo que
// un caso puede cambiarlo sin tocar el de los demás.
func loteDeEjemplo() core.Lote {
	return core.Lote{
		Fuente:        "boe",
		URL:           "https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565/texto/bloque/a21",
		FechaConsulta: "2026-09-28T12:00:00.123456789+02:00",
		Vigencia:      24 * time.Hour,
		Operaciones: []schema.Operacion{
			schema.Nodo{ID: idNorma, Tipo: grafo.TipoNorma, Datos: map[string]any{grafo.DatoIdentificador: "BOE-A-2015-10565"}},
			schema.Nodo{ID: idBloque, Tipo: grafo.TipoBloque, Datos: map[string]any{grafo.DatoBloque: "a21"}},
			schema.Arista{Origen: idNorma, Relacion: grafo.RelacionTieneParte, Destino: idBloque},
			schema.Texto{Huella: huellaDe(cuerpoA21), Cuerpo: cuerpoA21},
			schema.Nodo{ID: idMunicipio, Tipo: grafo.TipoMunicipio, Datos: map[string]any{grafo.DatoCodigoINE: "28074"}},
			schema.Nodo{ID: idOrgano, Tipo: grafo.TipoOrgano, Datos: map[string]any{grafo.DatoDIR3: idOrgano}},
			schema.Arista{Origen: idOrgano, Relacion: grafo.RelacionPerteneceA, Destino: idMunicipio},
		},
	}
}

// conOperacion mete una operación en el lote de ejemplo entre las válidas,
// para que se vea que una sola rechaza el lote entero.
func conOperacion(operacion schema.Operacion) func(*core.Lote) {
	return func(lote *core.Lote) {
		lote.Operaciones = slices.Insert(lote.Operaciones, 3, operacion)
	}
}

// TestValidarLote fija lo que un lote tiene que cumplir para entrar en el
// grafo del mundo sin mirar lo guardado (FR-024, FR-025;
// contracts/almacen-world-db.md §5; data-model §4.2; research.md D14): cada
// motivo rechaza el lote entero —la operación que lo incumple va entre otras
// válidas— con un Rechazo de clase «inesperado» que nombra la operación y el
// motivo; y, contra un grafo vacío, cada extremo de arista tiene que ser un
// nodo del lote.
func TestValidarLote(t *testing.T) {
	t.Parallel()

	t.Run("aceptados", probarLotesAceptados)
	t.Run("rechazos de la procedencia", probarRechazosDeLaProcedencia)
	t.Run("rechazos de una operacion", probarRechazosDeUnaOperacion)
	t.Run("contra el grafo vacio", probarContraGrafoVacio)
}

// probarLotesAceptados fija lo que entra: el lote de ejemplo y sus variantes
// válidas, que ValidarContraGrafoVacio también acepta porque cada extremo de
// sus aristas es un nodo del lote.
func probarLotesAceptados(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre  string
		cambiar func(*core.Lote)
	}{
		{"el lote de ejemplo", func(*core.Lote) {}},
		{"sin operaciones", func(lote *core.Lote) { lote.Operaciones = nil }},
		{"sin vigencia declarada", func(lote *core.Lote) { lote.Vigencia = 0 }},
		{"una fecha en UTC sin fraccion", func(lote *core.Lote) { lote.FechaConsulta = "2026-09-28T10:00:00Z" }},
		{"una url de otro esquema", func(lote *core.Lote) { lote.URL = "urn:ine:28074" }},
		{
			"el mismo id dos veces con el mismo tipo",
			conOperacion(schema.Nodo{ID: idNorma, Tipo: grafo.TipoNorma, Datos: map[string]any{"otro": "dato"}}),
		},
		{
			"una arista antes que sus extremos",
			func(lote *core.Lote) { slices.Reverse(lote.Operaciones) },
		},
		{
			"una Persona sin documento",
			conOperacion(schema.Nodo{ID: "persona:ana", Tipo: grafo.TipoPersona, Datos: map[string]any{"nombre": "Ana"}}),
		},
		{
			"un documento fuera de una Persona, que no se examina",
			conOperacion(schema.Nodo{ID: "12345678Z", Tipo: grafo.TipoNorma, Datos: map[string]any{"12345678Z": "12345678Z"}}),
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			lote := loteDeEjemplo()
			caso.cambiar(&lote)

			require.NoError(t, grafo.ValidarLote(lote))
			require.NoError(t, grafo.ValidarContraGrafoVacio(lote))
		})
	}
}

// probarRechazosDeLaProcedencia fija los rechazos del lote mismo: sin fuente,
// sin url, con una url que no es un URI absoluto (el criterio de
// schema.Procedencia.Validar), sin fecha, con una fecha que no es RFC 3339 y
// con una vigencia negativa. El Rechazo no nombra ninguna operación.
func probarRechazosDeLaProcedencia(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre  string
		cambiar func(*core.Lote)
		mensaje string
	}{
		{"sin fuente", func(lote *core.Lote) { lote.Fuente = "" }, "el lote: no lleva fuente"},
		{"sin url", func(lote *core.Lote) { lote.URL = "" }, "el lote: no lleva url"},
		{
			"una url relativa",
			func(lote *core.Lote) { lote.URL = "/buscar/act.php?id=BOE-A-2015-10565" },
			`el lote: la url "/buscar/act.php?id=BOE-A-2015-10565" no es un URI absoluto`,
		},
		{
			"una url sin esquema delante de los dos puntos",
			func(lote *core.Lote) { lote.URL = "://www.boe.es" },
			`el lote: la url "://www.boe.es" no es un URI absoluto`,
		},
		{
			"una url con un caracter de control",
			func(lote *core.Lote) { lote.URL = "https://www.boe.es/\n" },
			`el lote: la url "https://www.boe.es/\n" no es un URI absoluto`,
		},
		{"sin fecha", func(lote *core.Lote) { lote.FechaConsulta = "" }, "el lote: no lleva fecha de consulta"},
		{
			"una fecha sin hora",
			func(lote *core.Lote) { lote.FechaConsulta = "2026-09-28" },
			`el lote: la fecha de consulta "2026-09-28" no es RFC 3339`,
		},
		{
			"una fecha sin desplazamiento",
			func(lote *core.Lote) { lote.FechaConsulta = "2026-09-28T12:00:00" },
			`el lote: la fecha de consulta "2026-09-28T12:00:00" no es RFC 3339`,
		},
		{
			"una fecha con un espacio en vez de la T",
			func(lote *core.Lote) { lote.FechaConsulta = "2026-09-28 12:00:00Z" },
			`el lote: la fecha de consulta "2026-09-28 12:00:00Z" no es RFC 3339`,
		},
		{
			"una fecha que no existe",
			func(lote *core.Lote) { lote.FechaConsulta = "2026-02-30T12:00:00Z" },
			`el lote: la fecha de consulta "2026-02-30T12:00:00Z" no es RFC 3339`,
		},
		{
			"una vigencia negativa",
			func(lote *core.Lote) { lote.Vigencia = -time.Nanosecond },
			"el lote: la vigencia -1ns es negativa",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			lote := loteDeEjemplo()
			caso.cambiar(&lote)

			exigirRechazoDelLote(t, grafo.ValidarLote(lote), nil, caso.mensaje)
		})
	}
}

// probarRechazosDeUnaOperacion fija los rechazos de una operación, que va
// entre las válidas del lote de ejemplo: un nodo sin id o sin tipo, una arista
// sin origen, relación o destino, un texto cuya huella no es la de su cuerpo,
// un id con dos tipos en el lote y una Persona con un documento de identidad
// (los demás casos de FR-025, en TestPersonaSinDocumento). Una operación nula
// o que no es un valor Nodo, Arista o Texto tampoco entra.
func probarRechazosDeUnaOperacion(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		operacion schema.Operacion
		mensaje   string
	}{
		{"un nodo sin id", schema.Nodo{Tipo: grafo.TipoNorma}, `el nodo "": ` + motivoSinID},
		{"un nodo sin tipo", schema.Nodo{ID: idMunicipio}, `el nodo "ine:28074": ` + motivoSinTipo},
		{"una Persona sin id", schema.Nodo{Tipo: grafo.TipoPersona}, `el nodo de tipo "Persona": ` + motivoSinID},
		{
			"una arista sin origen",
			schema.Arista{Relacion: grafo.RelacionTieneParte, Destino: idBloque},
			`la arista de "" a "eli/es/l/2015/10/01/39#a21" por "eli:has_part": ` + motivoSinOrigen,
		},
		{
			"una arista sin relacion",
			schema.Arista{Origen: idNorma, Destino: idBloque},
			`la arista de "eli/es/l/2015/10/01/39" a "eli/es/l/2015/10/01/39#a21" por "": ` + motivoSinRel,
		},
		{
			"una arista sin destino",
			schema.Arista{Origen: idNorma, Relacion: grafo.RelacionTieneParte},
			`la arista de "eli/es/l/2015/10/01/39" a "" por "eli:has_part": ` + motivoSinDestino,
		},
		{
			"un id con dos tipos en el lote",
			schema.Nodo{ID: idNorma, Tipo: grafo.TipoMunicipio},
			`el nodo "eli/es/l/2015/10/01/39": el lote le da dos tipos, "Norma" y "Municipio"`,
		},
		{
			"una Persona con un documento en su id",
			schema.Nodo{ID: "12345678Z", Tipo: grafo.TipoPersona},
			`el nodo de tipo "Persona": ` + motivoIDConDocumento,
		},
		{"una operacion nula", nil, "el lote: trae una operación nula"},
		{
			"una operacion que no es un valor",
			&schema.Nodo{ID: idMunicipio, Tipo: grafo.TipoMunicipio},
			"la operación de tipo *schema.Nodo: solo entran los valores schema.Nodo, schema.Arista y schema.Texto",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			lote := loteDeEjemplo()
			conOperacion(caso.operacion)(&lote)

			exigirRechazoDelLote(t, grafo.ValidarLote(lote), caso.operacion, caso.mensaje)
		})
	}

	t.Run("huellas que no son la del cuerpo", probarHuellasQueNoSonLaDelCuerpo)
	t.Run("un id con dos tipos, uno de Persona", probarPersonaConDosTipos)
	t.Run("una Persona antes que la arista que la nombra", probarPersonaAntesQueSusAristas)
}

// probarPersonaAntesQueSusAristas fija que los nodos se validan antes que las
// aristas: una Persona con un documento en su id rechaza el lote por serlo
// aunque antes vaya una arista incompleta que la nombra por su id, y el
// mensaje no repite el documento.
func probarPersonaAntesQueSusAristas(t *testing.T) {
	t.Parallel()

	persona := schema.Nodo{ID: "12345678Z", Tipo: grafo.TipoPersona}
	lote := loteDeEjemplo()
	conOperacion(persona)(&lote)
	conOperacion(schema.Arista{Origen: persona.ID, Destino: idBloque})(&lote)

	err := grafo.ValidarLote(lote)
	exigirRechazoDelLote(t, err, persona, `el nodo de tipo "Persona": `+motivoIDConDocumento)
	assert.NotContains(t, err.Error(), persona.ID, "el mensaje nunca repite el documento")
}

// probarHuellasQueNoSonLaDelCuerpo fija que la huella de un texto es
// exactamente «sha256:» y los 64 hexadecimales en minúscula de la SHA-256 de
// los bytes de su cuerpo, tal cual: cualquier otra forma, o la de otro cuerpo,
// rechaza el lote. El mensaje nombra el texto por su huella y nunca repite el
// cuerpo.
func probarHuellasQueNoSonLaDelCuerpo(t *testing.T) {
	t.Parallel()

	buena := huellaDe(cuerpoA21)
	hexadecimal := strings.TrimPrefix(buena, "sha256:")

	casos := []struct {
		nombre string
		huella string
	}{
		{"la de otro cuerpo", huellaDe("Artículo 22. Suspensión del plazo máximo para resolver.")},
		{"la del cuerpo con un salto de linea detras", huellaDe(cuerpoA21 + "\n")},
		{"vacia", ""},
		{"sin prefijo", hexadecimal},
		{"con el prefijo en mayusculas", "SHA256:" + hexadecimal},
		{"con el hexadecimal en mayusculas", "sha256:" + strings.ToUpper(hexadecimal)},
		{"con un hexadecimal de menos", buena[:len(buena)-1]},
		{"con un hexadecimal de mas", buena + "0"},
		{"con un espacio detras", buena + " "},
		{"con otro algoritmo", "sha512:" + hexadecimal},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			texto := schema.Texto{Huella: caso.huella, Cuerpo: cuerpoA21}
			lote := loteDeEjemplo()
			conOperacion(texto)(&lote)

			err := grafo.ValidarLote(lote)
			exigirRechazoDelLote(t, err, texto, `el texto "`+caso.huella+`": `+motivoHuella)
			assert.NotContains(t, err.Error(), cuerpoA21, "el mensaje nunca repite el cuerpo")
		})
	}
}

// probarPersonaConDosTipos fija que, cuando uno de los dos tipos de un id es
// Persona, el Rechazo nombra la Persona —por su tipo y no por su id— llegue
// antes o después que el otro nodo: el mensaje nunca repite el id de una
// Persona.
func probarPersonaConDosTipos(t *testing.T) {
	t.Parallel()

	persona := schema.Nodo{ID: "persona:ana", Tipo: grafo.TipoPersona, Datos: map[string]any{"nombre": "Ana"}}
	norma := schema.Nodo{ID: persona.ID, Tipo: grafo.TipoNorma}

	casos := []struct {
		nombre  string
		primero schema.Nodo
		segundo schema.Nodo
		mensaje string
	}{
		{"la Persona primero", persona, norma, `el nodo de tipo "Persona": el lote le da dos tipos, "Persona" y "Norma"`},
		{"la Persona despues", norma, persona, `el nodo de tipo "Persona": el lote le da dos tipos, "Norma" y "Persona"`},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			lote := loteDeEjemplo()
			conOperacion(caso.segundo)(&lote)
			conOperacion(caso.primero)(&lote)

			err := grafo.ValidarLote(lote)
			exigirRechazoDelLote(t, err, persona, caso.mensaje)
			assert.NotContains(t, err.Error(), persona.ID, "el mensaje nunca repite el id de una Persona")
		})
	}
}

// probarContraGrafoVacio fija que, contra un grafo vacío —world.db ausente o
// sin esquema—, cada extremo de una arista tiene que ser un nodo del lote,
// esté antes o después que ella; el primero que falta rechaza el lote entero.
// ValidarLote no lo mira: lo que no está en el lote puede estar en el grafo.
func probarContraGrafoVacio(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre  string
		arista  schema.Arista
		mensaje string
	}{
		{
			"sin su origen",
			schema.Arista{Origen: "eli/es/l/2015/10/02/40", Relacion: grafo.RelacionTieneParte, Destino: idBloque},
			`la arista de "eli/es/l/2015/10/02/40" a "eli/es/l/2015/10/01/39#a21" por "eli:has_part": ` +
				"su origen no es un nodo del lote y el grafo está vacío",
		},
		{
			"sin su destino",
			schema.Arista{Origen: idOrgano, Relacion: grafo.RelacionPerteneceA, Destino: "ine:28079"},
			`la arista de "L01280745" a "ine:28079" por "lb:pertenece_a": ` +
				"su destino no es un nodo del lote y el grafo está vacío",
		},
		{
			"sin ninguno de los dos, que nombra el origen",
			schema.Arista{Origen: "eli/es/l/2015/10/02/40", Relacion: grafo.RelacionTieneParte, Destino: "otro"},
			`la arista de "eli/es/l/2015/10/02/40" a "otro" por "eli:has_part": ` +
				"su origen no es un nodo del lote y el grafo está vacío",
		},
		{
			"con el id de un texto como extremo, que no es un nodo",
			schema.Arista{Origen: idNorma, Relacion: grafo.RelacionTieneParte, Destino: huellaDe(cuerpoA21)},
			`la arista de "eli/es/l/2015/10/01/39" a "` + huellaDe(cuerpoA21) + `" por "eli:has_part": ` +
				"su destino no es un nodo del lote y el grafo está vacío",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			lote := loteDeEjemplo()
			conOperacion(caso.arista)(&lote)

			require.NoError(t, grafo.ValidarLote(lote), "lo que no está en el lote puede estar en el grafo")
			exigirRechazoDelLote(t, grafo.ValidarContraGrafoVacio(lote), caso.arista, caso.mensaje)
		})
	}
}

// exigirRechazoDelLote comprueba que err es el Rechazo de la operación
// rechazada —nula si es el lote mismo—, de clase «inesperado» y con el mensaje
// que nombra la operación y el motivo.
func exigirRechazoDelLote(t *testing.T, err error, rechazada schema.Operacion, mensaje string) {
	t.Helper()

	require.EqualError(t, err, mensaje)

	var rechazo *grafo.Rechazo
	require.ErrorAs(t, err, &rechazo)
	assert.Equal(t, rechazada, rechazo.Operacion, "la operación rechazada, tal como la trae el lote")

	var conClase schema.ConClase
	require.ErrorAs(t, err, &conClase)
	assert.Equal(t, schema.ClaseInesperado, conClase.Clase())
}
