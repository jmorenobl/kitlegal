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

// motivoHuella es el motivo con el que ValidarLote rechaza un texto cuya huella
// no es la de su cuerpo (H7.1 FR-075; contracts/almacen-world-db.md §5).
const motivoHuella = "su huella no es la de su cuerpo: «sha256:» y los 64 hexadecimales en minúscula de su SHA-256"

// huellaDe es la huella de un texto como la exige FR-024: «sha256:» y los 64
// hexadecimales en minúscula de la SHA-256 de los bytes del cuerpo, tal cual.
func huellaDe(cuerpo string) string {
	suma := sha256.Sum256([]byte(cuerpo))

	return "sha256:" + hex.EncodeToString(suma[:])
}

// loteDeEjemplo es un lote que ValidarLote acepta, con cada clase de
// operación. Cada llamada construye uno nuevo, de modo que un caso puede
// cambiarlo sin tocar el de los demás.
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
// grafo del mundo sin mirar lo guardado (H7.1 FR-074, FR-075;
// contracts/almacen-world-db.md §5; data-model §4.2; research.md D20): cada
// motivo rechaza el lote entero —la operación que lo incumple va entre otras
// válidas— con un Rechazo de clase «inesperado» que nombra la operación y el
// motivo.
func TestValidarLote(t *testing.T) {
	t.Parallel()

	t.Run("aceptados", probarLotesAceptados)
	t.Run("rechazos de la procedencia", probarRechazosDeLaProcedencia)
	t.Run("rechazos de una operacion", probarRechazosDeUnaOperacion)
}

// probarLotesAceptados fija lo que entra: el lote de ejemplo y sus variantes
// válidas.
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
			"el mismo id dos veces con el mismo tipo y los mismos datos",
			conOperacion(schema.Nodo{ID: idNorma, Tipo: grafo.TipoNorma, Datos: map[string]any{grafo.DatoIdentificador: "BOE-A-2015-10565"}}),
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
		})
	}
}

// probarRechazosDeLaProcedencia fija los rechazos del lote mismo: sin fuente,
// sin url y sin fecha de consulta. El Rechazo no nombra ninguna operación.
func probarRechazosDeLaProcedencia(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre  string
		cambiar func(*core.Lote)
		mensaje string
	}{
		{"sin fuente", func(lote *core.Lote) { lote.Fuente = "" }, "el lote: no lleva fuente"},
		{"sin url", func(lote *core.Lote) { lote.URL = "" }, "el lote: no lleva url"},
		{"sin fecha", func(lote *core.Lote) { lote.FechaConsulta = "" }, "el lote: no lleva fecha de consulta"},
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
// entre las válidas del lote de ejemplo: un id con dos tipos en el lote, un
// texto cuya huella no es la de su cuerpo —que el mensaje nombra por su huella,
// sin repetir el cuerpo— y una Persona con un documento de identidad (los demás
// casos, en TestPersonaSinDocumento).
func probarRechazosDeUnaOperacion(t *testing.T) {
	t.Parallel()

	laDeOtroCuerpo := huellaDe("Art\xc3\xadculo 22. Suspensi\xc3\xb3n del plazo m\xc3\xa1ximo para resolver.")

	casos := []struct {
		nombre    string
		operacion schema.Operacion
		mensaje   string
	}{
		{
			"un id con dos tipos en el lote",
			schema.Nodo{ID: idNorma, Tipo: grafo.TipoMunicipio},
			`el nodo "eli/es/l/2015/10/01/39": el lote le da dos tipos, "Norma" y "Municipio"`,
		},
		{
			"un texto con la huella de otro cuerpo",
			schema.Texto{Huella: laDeOtroCuerpo, Cuerpo: cuerpoA21},
			`el texto "` + laDeOtroCuerpo + `": ` + motivoHuella,
		},
		{
			"una Persona con un documento en su id",
			schema.Nodo{ID: "12345678Z", Tipo: grafo.TipoPersona},
			`el nodo de tipo "Persona": ` + motivoIDConDocumento,
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

	t.Run("un id con dos tipos, uno de Persona", probarPersonaConDosTipos)
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

// TestConsolidar fija la consolidación de un lote (data-model §4.2; H7.1
// FR-076): dos operaciones con la misma clave —id, terna o huella— comparten
// la observación del lote y quedan en un solo registro; el resultado no
// depende del orden de las operaciones ni de que se repitan; y un lote que no
// se puede consolidar se rechaza entero.
func TestConsolidar(t *testing.T) {
	t.Parallel()

	t.Run("consolidados", probarLotesConsolidados)
	t.Run("sin operaciones", probarLoteSinOperaciones)
	t.Run("rechazos", probarLotesQueNoSeConsolidan)
}

// probarLotesConsolidados fija lo que Consolidar da del lote de ejemplo y de
// sus variantes: en otro orden y con operaciones repetidas.
func probarLotesConsolidados(t *testing.T) {
	t.Parallel()

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
// ValidarLote rechaza: da el mismo Rechazo y nada más.
func probarLotesQueNoSeConsolidan(t *testing.T) {
	t.Parallel()

	sinFuente := loteDeEjemplo()
	sinFuente.Fuente = ""

	consolidado, err := grafo.Consolidar(sinFuente)
	exigirRechazoDelLote(t, err, nil, "el lote: no lleva fuente")
	assert.Zero(t, consolidado)
}
