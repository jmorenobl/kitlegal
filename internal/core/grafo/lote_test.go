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

// consolidadoDeEjemplo es lo que Consolidar da del lote de ejemplo: un
// registro por id, por terna y por huella, ordenados por su clave comparando
// bytes, cada uno con la procedencia del lote como primera y como última
// observación, su vigencia y, en un nodo, sus datos canónicos.
func consolidadoDeEjemplo() grafo.Consolidado {
	lote := loteDeEjemplo()
	procedencia := grafo.Procedencia{Fuente: lote.Fuente, URL: lote.URL, FechaConsulta: lote.FechaConsulta}

	nodo := func(id, tipo, datos string) grafo.RegistroDeNodo {
		return grafo.RegistroDeNodo{
			ID: id, Tipo: tipo, Datos: datos,
			PrimeraObservacion: procedencia, UltimaObservacion: procedencia, Vigencia: lote.Vigencia,
		}
	}
	arista := func(origen, relacion, destino string) grafo.RegistroDeArista {
		return grafo.RegistroDeArista{
			Origen: origen, Relacion: relacion, Destino: destino,
			PrimeraObservacion: procedencia, UltimaObservacion: procedencia, Vigencia: lote.Vigencia,
		}
	}

	return grafo.Consolidado{
		Nodos: []grafo.RegistroDeNodo{
			nodo(idOrgano, grafo.TipoOrgano, `{"dir3":"L01280745"}`),
			nodo(idNorma, grafo.TipoNorma, `{"identificador":"BOE-A-2015-10565"}`),
			nodo(idBloque, grafo.TipoBloque, `{"bloque":"a21"}`),
			nodo(idMunicipio, grafo.TipoMunicipio, `{"codigo_ine":"28074"}`),
		},
		Aristas: []grafo.RegistroDeArista{
			arista(idOrgano, grafo.RelacionPerteneceA, idMunicipio),
			arista(idNorma, grafo.RelacionTieneParte, idBloque),
		},
		Textos: []grafo.RegistroDeTexto{{Huella: huellaDe(cuerpoA21), Cuerpo: cuerpoA21, Procedencia: procedencia}},
	}
}

// TestConsolidar fija la consolidación de un lote (data-model §4.2; FR-023):
// dos operaciones con la misma clave —id, terna o huella— comparten la
// observación del lote y quedan en un solo registro; dos nodos con el mismo id
// y datos distintos se quedan con los datos canónicos menores, comparando
// bytes, en los dos órdenes de llegada; el resultado no depende del orden de
// las operaciones ni de que se repitan; y un lote que no se puede consolidar
// se rechaza entero.
func TestConsolidar(t *testing.T) {
	t.Parallel()

	t.Run("consolidados", probarLotesConsolidados)
	t.Run("sin operaciones", probarLoteSinOperaciones)
	t.Run("rechazos", probarLotesQueNoSeConsolidan)
}

// probarLotesConsolidados fija lo que Consolidar da del lote de ejemplo y de
// sus variantes: en otro orden, con operaciones repetidas y con un nodo
// repetido con datos distintos antes o después del primero.
func probarLotesConsolidados(t *testing.T) {
	t.Parallel()

	conMasDatos := schema.Nodo{
		ID: idNorma, Tipo: grafo.TipoNorma, Datos: map[string]any{grafo.DatoIdentificador: "BOE-A-2015-10565", "otro": "x"},
	}
	conOtrosDatos := schema.Nodo{
		ID: idNorma, Tipo: grafo.TipoNorma, Datos: map[string]any{grafo.DatoIdentificador: "BOE-A-2015-10566"},
	}
	alPrincipio := func(operacion schema.Operacion) func(*core.Lote) {
		return func(lote *core.Lote) { lote.Operaciones = slices.Insert(lote.Operaciones, 0, operacion) }
	}
	otroTexto := schema.Texto{Huella: huellaDe("Articulo 22."), Cuerpo: "Articulo 22."}
	conOtroTexto := func(consolidado *grafo.Consolidado) {
		registro := consolidado.Textos[0]
		registro.Huella, registro.Cuerpo = otroTexto.Huella, otroTexto.Cuerpo
		consolidado.Textos = append(consolidado.Textos, registro)
		slices.SortFunc(consolidado.Textos, func(a, b grafo.RegistroDeTexto) int {
			return strings.Compare(a.Huella, b.Huella)
		})
	}
	igual := func(*grafo.Consolidado) {}
	// La coma va antes que la llave: los datos con la clave «otro» son los
	// menores.
	losDeMasDatos := func(consolidado *grafo.Consolidado) {
		consolidado.Nodos[1].Datos = `{"identificador":"BOE-A-2015-10565","otro":"x"}`
	}

	casos := []struct {
		nombre   string
		cambiar  func(*core.Lote)
		esperado func(*grafo.Consolidado)
	}{
		{"el lote de ejemplo", func(*core.Lote) {}, igual},
		{"en orden inverso", func(lote *core.Lote) { slices.Reverse(lote.Operaciones) }, igual},
		{"cada operacion dos veces", func(lote *core.Lote) {
			lote.Operaciones = slices.Concat(lote.Operaciones, lote.Operaciones)
		}, igual},
		{"cada operacion dos veces, la segunda en orden inverso", func(lote *core.Lote) {
			inversas := slices.Clone(lote.Operaciones)
			slices.Reverse(inversas)
			lote.Operaciones = slices.Concat(lote.Operaciones, inversas)
		}, igual},
		{"un nodo repetido con datos menores, despues del primero", conOperacion(conMasDatos), losDeMasDatos},
		{"un nodo repetido con datos menores, antes del primero", alPrincipio(conMasDatos), losDeMasDatos},
		{"un nodo repetido con datos mayores, despues del primero", conOperacion(conOtrosDatos), igual},
		{"un nodo repetido con datos mayores, antes del primero", alPrincipio(conOtrosDatos), igual},
		{"dos textos, por su huella", conOperacion(otroTexto), conOtroTexto},
		{"dos textos en el otro orden", alPrincipio(otroTexto), conOtroTexto},
		{"sin vigencia declarada", func(lote *core.Lote) { lote.Vigencia = 0 }, func(consolidado *grafo.Consolidado) {
			for i := range consolidado.Nodos {
				consolidado.Nodos[i].Vigencia = 0
			}

			for i := range consolidado.Aristas {
				consolidado.Aristas[i].Vigencia = 0
			}
		}},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			lote := loteDeEjemplo()
			caso.cambiar(&lote)
			esperado := consolidadoDeEjemplo()
			caso.esperado(&esperado)

			consolidado, err := grafo.Consolidar(lote)
			require.NoError(t, err)
			assert.Equal(t, esperado, consolidado)
		})
	}
}

// probarLoteSinOperaciones fija que un lote válido sin operaciones no da
// ningún registro.
func probarLoteSinOperaciones(t *testing.T) {
	t.Parallel()

	lote := loteDeEjemplo()
	lote.Operaciones = nil

	consolidado, err := grafo.Consolidar(lote)
	require.NoError(t, err)
	assert.Empty(t, consolidado.Nodos)
	assert.Empty(t, consolidado.Aristas)
	assert.Empty(t, consolidado.Textos)
}

// probarLotesQueNoSeConsolidan fija que Consolidar no consolida un lote que
// ValidarLote rechaza, con el mismo Rechazo, ni uno con un nodo cuyos datos no
// tienen forma JSON canónica, que no se podrían guardar ni comparar: un
// Rechazo que nombra el nodo y dice por qué.
func probarLotesQueNoSeConsolidan(t *testing.T) {
	t.Parallel()

	sinFuente := loteDeEjemplo()
	sinFuente.Fuente = ""

	consolidado, err := grafo.Consolidar(sinFuente)
	exigirRechazoDelLote(t, err, nil, "el lote: no lleva fuente")
	assert.Zero(t, consolidado)

	sinJSON := schema.Nodo{ID: "eli/es/l/2015/10/02/40", Tipo: grafo.TipoNorma, Datos: map[string]any{"rango": complex(1, 2)}}
	lote := loteDeEjemplo()
	conOperacion(sinJSON)(&lote)
	require.NoError(t, grafo.ValidarLote(lote), "ValidarLote no examina los datos de lo que no es una Persona")

	consolidado, err = grafo.Consolidar(lote)
	require.ErrorContains(t, err, `el nodo "eli/es/l/2015/10/02/40": los datos no tienen forma JSON`)
	assert.Zero(t, consolidado)

	var rechazo *grafo.Rechazo
	require.ErrorAs(t, err, &rechazo)
	assert.Equal(t, sinJSON, rechazo.Operacion, "la operación rechazada, tal como la trae el lote")
}
