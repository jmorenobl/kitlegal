package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// errEntregaRota es la causa de una entrega que falla. Lleva dos saltos de
// línea a propósito: la línea de aviso los convierte en espacios, y es lo que
// hace que una causa de varias líneas quepa en una sola.
var errEntregaRota = errors.New("grafo: world.db no se puede escribir:\nprimera causa\nsegunda causa")

// lineaDeEntregaRota es la línea que el contrato fija para esa causa, escrita a
// mano y no derivada del código que se comprueba (contracts/resultado-y-entrega.md
// §4, research.md D7).
const lineaDeEntregaRota = "kitlegal: lo observado no ha llegado al grafo del mundo: " +
	"grafo: world.db no se puede escribir: primera causa segunda causa"

// consultaDePrueba es la fecha que declara una procedencia que sí la conoce.
// Lleva otro desplazamiento que el reloj del montador y una fracción de segundo
// con ceros a la derecha: el lote tiene que llevar el mismo texto que escribe el
// sobre, y ese texto los quita (research.md V2).
var consultaDePrueba = time.Date(2026, time.September, 4, 8, 45, 30, 250_000_000,
	time.FixedZone("CET", 60*60))

// observadoDePrueba es lo que declara un applet que observa el mundo: una
// operación de cada clase y la vigencia de la consulta.
func observadoDePrueba() schema.Observado {
	norma := "https://www.boe.es/eli/es/l/2015/10/01/39"
	bloque := norma + "/a21"

	return schema.Observado{
		Vigencia: 7 * 24 * time.Hour,
		Operaciones: []schema.Operacion{
			schema.Nodo{ID: norma, Tipo: "Norma", Datos: map[string]any{"identificador": "BOE-A-2015-10565"}},
			schema.Arista{Origen: norma, Relacion: "eli:has_part", Destino: bloque},
			schema.Texto{Huella: schema.PrefijoHuella + strings.Repeat("0", 64), Cuerpo: "texto del bloque a21"},
		},
	}
}

// resultadoQueObserva es un resultado correcto que trae lo observado, con la
// procedencia que se le dé.
func resultadoQueObserva(proc schema.Procedencia) schema.Resultado {
	res := resultadoDePrueba()
	res.Procedencia = proc
	res.Grafo = observadoDePrueba()

	return res
}

// almacenEspia es el doble del grafo del mundo: anota cada lote que recibe, el
// límite y el estado del contexto con que lo recibe y lo que la salida
// estándar ya llevaba cuando llegó, y devuelve el fallo que se le fije.
type almacenEspia struct {
	lotes []core.Lote
	// limites son los instantes límite de cada contexto, y conLimite si lo
	// tenía.
	limites   []time.Time
	conLimite []bool
	// vivos dice si cada contexto seguía sin terminar al entregar.
	vivos []bool
	// presentado es lo que había en la salida estándar observada cuando llegó
	// cada lote.
	presentado []string
	// salida es la salida estándar del presentador que se observa, o nula.
	salida *escritor
	// fallo, si no es nulo, es lo que devuelve Apply.
	fallo error
}

var _ core.GraphStore = (*almacenEspia)(nil)

func (a *almacenEspia) Apply(ctx context.Context, lote core.Lote) error {
	limite, conLimite := ctx.Deadline()

	a.lotes = append(a.lotes, lote)
	a.limites = append(a.limites, limite)
	a.conLimite = append(a.conLimite, conLimite)
	a.vivos = append(a.vivos, ctx.Err() == nil)

	if a.salida != nil {
		a.presentado = append(a.presentado, a.salida.String())
	}

	return a.fallo
}

// presentadorConAvisoRoto presenta con normalidad y falla solo al escribir en la
// salida de error, como un descriptor 2 cerrado con el 1 sano.
type presentadorConAvisoRoto struct {
	presentadorConJSON
}

var _ Presentador = (*presentadorConAvisoRoto)(nil)

func (p *presentadorConAvisoRoto) Aviso(texto string) error {
	p.avisos = append(p.avisos, texto)

	return errEscrituraRota
}

// emitirConGrafo emite con el montador de las tablas, que entrega en almacen, y
// devuelve el código.
func emitirConGrafo(
	ctx context.Context, p Presentador, enJSON bool, almacen core.GraphStore, res schema.Resultado, err error,
) int {
	montador := montadorDePrueba()
	montador.Grafo = almacen

	return montador.Emitir(ctx, p, enJSON, res, err)
}

// casoDeEntrega es un subtest de TestEntregaDelMontador.
type casoDeEntrega struct {
	nombre    string
	comprueba func(t *testing.T)
}

// TestEntregaDelMontador comprueba la entrega que hace el kernel detrás de
// presentar (contracts/resultado-y-entrega.md §3 y §4; research.md D4, D5 y
// D7): el lote lleva la procedencia del sobre presentado —también cuando lo
// fechó el reloj del montador—, en el mismo texto que el sobre escribe; se
// entrega después de presentar, en las tres formas, y con el contexto recibido;
// no se entrega nada sin operaciones, sin almacén o con un fallo; una entrega
// que falla deja una línea de aviso y el mismo código y la misma salida estándar,
// y el error al escribir esa línea no tiene efecto (FR-021, FR-026, FR-030,
// FR-032, FR-033).
func TestEntregaDelMontador(t *testing.T) {
	t.Parallel()

	casos := []casoDeEntrega{
		{"el lote lleva la procedencia del sobre presentado", compruebaLoteConLaProcedenciaDelSobre},
		{"el lote lleva la fecha del reloj cuando la puso el montador", compruebaLoteFechadoPorElReloj},
		{"LoteDe escribe la fecha como el sobre", compruebaLoteDeEscribeLaFechaComoElSobre},
		{"se entrega después de presentar, en las tres formas", compruebaEntregaDespuesDePresentar},
		{"se entrega con el contexto recibido", compruebaEntregaConElContextoRecibido},
		{"sin operaciones no se entrega nada", compruebaSinOperacionesNoEntrega},
		{"sin almacén no se entrega nada", compruebaSinAlmacenNoEntrega},
		{"un fallo no entrega nada", compruebaUnFalloNoEntrega},
		{"una entrega fallida deja una línea de aviso y el mismo código", compruebaEntregaFallida},
		{"el error al escribir el aviso no tiene efecto", compruebaAvisoQueNoSeEscribe},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			caso.comprueba(t)
		})
	}
}

// compruebaLoteConLaProcedenciaDelSobre exige que el lote lleve la fuente, la
// url y la fecha del sobre que salió por la salida estándar —la fecha, en el
// mismo texto que escribe el sobre— y la vigencia y las operaciones del
// Observado, en su orden (FR-021, research.md D5).
func compruebaLoteConLaProcedenciaDelSobre(t *testing.T) {
	t.Helper()

	conFecha := procedenciaDelApplet
	conFecha.FechaConsulta = consultaDePrueba

	doble := &presentadorConJSON{}
	almacen := &almacenEspia{}

	codigo := emitirConGrafo(t.Context(), doble, true, almacen, resultadoQueObserva(conFecha), nil)

	require.Equal(t, 0, codigo)
	documento := unicoDocumento(t, doble.salida.String())

	require.Len(t, almacen.lotes, 1, "una invocación entrega un solo lote")
	lote := almacen.lotes[0]

	assert.Equal(t, documento["fuente"], lote.Fuente)
	assert.Equal(t, documento["url"], lote.URL)
	assert.Equal(t, documento["fecha_consulta"], lote.FechaConsulta,
		"el lote lleva la fecha en el mismo texto que la escribe el sobre")
	assert.Equal(t, "2026-09-04T08:45:30.25+01:00", lote.FechaConsulta,
		"la fecha de la consulta, con su desplazamiento y sin ceros a la derecha")
	assert.Equal(t, core.Lote{
		Fuente:        procedenciaDelApplet.Fuente,
		URL:           procedenciaDelApplet.URL,
		FechaConsulta: "2026-09-04T08:45:30.25+01:00",
		Vigencia:      observadoDePrueba().Vigencia,
		Operaciones:   observadoDePrueba().Operaciones,
	}, lote, "del Observado, la vigencia y las operaciones en su orden")
}

// compruebaLoteFechadoPorElReloj exige que, cuando la procedencia no declara la
// fecha, el lote lleve la que el reloj del montador puso en el sobre: la del
// sobre presentado y no otra lectura del reloj (research.md D4, V18).
func compruebaLoteFechadoPorElReloj(t *testing.T) {
	t.Helper()

	doble := &presentadorConJSON{}
	almacen := &almacenEspia{}

	lecturas := 0
	montador := Montador{
		Ahora: func() time.Time {
			lecturas++

			return instanteDePrueba.Add(time.Duration(lecturas) * time.Hour)
		},
		Grafo: almacen,
	}

	codigo := montador.Emitir(t.Context(), doble, true, resultadoQueObserva(procedenciaDelApplet), nil)

	require.Equal(t, 0, codigo)
	documento := unicoDocumento(t, doble.salida.String())

	require.Len(t, almacen.lotes, 1)
	assert.Equal(t, documento["fecha_consulta"], almacen.lotes[0].FechaConsulta,
		"el lote lleva la fecha del sobre presentado, aunque el reloj diera otra al volver a leerlo")
	assert.Equal(t, "2026-09-11T11:12:00+02:00", almacen.lotes[0].FechaConsulta)
	assert.Equal(t, procedenciaDelApplet.Fuente, almacen.lotes[0].Fuente)
	assert.Equal(t, procedenciaDelApplet.URL, almacen.lotes[0].URL)
}

// compruebaLoteDeEscribeLaFechaComoElSobre exige que LoteDe tome del sobre la
// fuente, la url y la fecha, esta en el texto exacto que el sobre escribe en
// fecha_consulta, y del Observado la vigencia y las operaciones tal cual
// (research.md D5, V2).
func compruebaLoteDeEscribeLaFechaComoElSobre(t *testing.T) {
	t.Helper()

	fechas := []time.Time{
		consultaDePrueba,
		instanteDePrueba,
		time.Date(2026, time.February, 4, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.February, 4, 0, 0, 0, 1, time.FixedZone("", -(3*60*60+30*60))),
	}

	for _, fecha := range fechas {
		sobre := schema.Sobre{
			Ok:            true,
			Fuente:        procedenciaDelApplet.Fuente,
			URL:           procedenciaDelApplet.URL,
			FechaConsulta: fecha,
			Hash:          schema.PrefijoHuella + strings.Repeat("0", 64),
			Data:          map[string]any{},
		}

		escrita, err := sobre.FechaConsulta.MarshalJSON()
		require.NoError(t, err)

		lote := LoteDe(sobre, observadoDePrueba())

		assert.Equal(t, core.Lote{
			Fuente:        procedenciaDelApplet.Fuente,
			URL:           procedenciaDelApplet.URL,
			FechaConsulta: strings.Trim(string(escrita), `"`),
			Vigencia:      observadoDePrueba().Vigencia,
			Operaciones:   observadoDePrueba().Operaciones,
		}, lote, "fecha %s", fecha)
	}
}

// compruebaEntregaDespuesDePresentar exige que la entrega llegue cuando la
// salida estándar ya lleva todo lo que la invocación presenta —el sobre en
// JSON, la tabla mínima o el texto legible— y que no la cambie (FR-026).
func compruebaEntregaDespuesDePresentar(t *testing.T) {
	t.Helper()

	legible := resultadoQueObserva(procedenciaDelApplet)
	legible.Legible = "El bloque a21, contado para una persona."

	formas := []struct {
		nombre string
		enJSON bool
		res    schema.Resultado
	}{
		{"el sobre en JSON", true, resultadoQueObserva(procedenciaDelApplet)},
		{"la tabla mínima", false, resultadoQueObserva(procedenciaDelApplet)},
		{"el texto legible", false, legible},
	}

	for _, forma := range formas {
		doble := &presentadorConJSON{}
		almacen := &almacenEspia{salida: &doble.salida}

		codigo := emitirConGrafo(t.Context(), doble, forma.enJSON, almacen, forma.res, nil)

		require.Equal(t, 0, codigo, forma.nombre)
		require.Len(t, almacen.presentado, 1, forma.nombre)
		assert.NotEmpty(t, almacen.presentado[0], "%s: se entrega cuando ya se ha presentado", forma.nombre)
		assert.Equal(t, doble.salida.String(), almacen.presentado[0],
			"%s: la entrega no añade nada a la salida estándar", forma.nombre)
		assert.Empty(t, doble.avisos, "%s: una entrega correcta no avisa", forma.nombre)
	}
}

// compruebaEntregaConElContextoRecibido exige que Apply reciba el contexto de
// Emitir, con su mismo instante límite: es el que Main crea con el plazo de
// --timeout de la invocación (FR-014).
func compruebaEntregaConElContextoRecibido(t *testing.T) {
	t.Helper()

	limite := time.Now().Add(90 * time.Second)

	ctx, cancelar := context.WithDeadline(t.Context(), limite)
	defer cancelar()

	almacen := &almacenEspia{}

	codigo := emitirConGrafo(ctx, &presentadorConJSON{}, true, almacen,
		resultadoQueObserva(procedenciaDelApplet), nil)

	require.Equal(t, 0, codigo)
	require.Len(t, almacen.limites, 1)
	assert.True(t, almacen.conLimite[0], "el contexto de la entrega tiene límite")
	assert.True(t, almacen.limites[0].Equal(limite), "el límite de la entrega es el del contexto recibido")
	assert.True(t, almacen.vivos[0], "el contexto recibido sigue vivo al entregar")
}

// compruebaSinOperacionesNoEntrega exige que un resultado sin operaciones no
// llegue al almacén, tampoco si declara una vigencia, y que salga igual que con
// ellas (FR-030).
func compruebaSinOperacionesNoEntrega(t *testing.T) {
	t.Helper()

	observados := []schema.Observado{
		{},
		{Vigencia: time.Hour},
		{Vigencia: time.Hour, Operaciones: []schema.Operacion{}},
	}

	for _, observado := range observados {
		res := resultadoDePrueba()
		res.Grafo = observado

		doble := &presentadorConJSON{}
		almacen := &almacenEspia{}

		codigo := emitirConGrafo(t.Context(), doble, true, almacen, res, nil)

		assert.Equal(t, 0, codigo)
		assert.Empty(t, almacen.lotes, "sin operaciones no se entrega nada: %+v", observado)
		assert.Empty(t, doble.avisos)
		assert.Equal(t, salidaSinGrafo(t, res), doble.salida.String())
	}
}

// compruebaSinAlmacenNoEntrega exige que un montador sin almacén —el valor cero,
// el de los tests y el de un registro que no entrega— presente lo mismo y no
// avise de nada (research.md D6).
func compruebaSinAlmacenNoEntrega(t *testing.T) {
	t.Helper()

	res := resultadoQueObserva(procedenciaDelApplet)

	doble := &presentadorConJSON{}
	codigo := emitirConGrafo(t.Context(), doble, true, nil, res, nil)

	assert.Equal(t, 0, codigo)
	assert.Empty(t, doble.avisos, "sin almacén no hay entrega que pueda fallar")
	assert.Equal(t, salidaSinGrafo(t, res), doble.salida.String())
}

// compruebaUnFalloNoEntrega exige que nada llegue al almacén cuando la
// invocación termina con un código distinto de 0: el fallo del applet, la
// procedencia que no sostiene una cita y la salida estándar que no se pudo
// escribir (FR-032).
func compruebaUnFalloNoEntrega(t *testing.T) {
	t.Helper()

	fallos := []struct {
		nombre string
		doble  *presentadorConJSON
		res    schema.Resultado
		err    error
		codigo int
	}{
		{
			nombre: "el applet falla",
			doble:  &presentadorConJSON{},
			res:    resultadoQueObserva(procedenciaDelApplet),
			err:    fmt.Errorf("el bloque a99: %w", ErrNoEncontrado),
			codigo: 3,
		},
		{
			nombre: "la procedencia no sostiene una cita",
			doble:  &presentadorConJSON{},
			res:    resultadoQueObserva(schema.Procedencia{Fuente: procedenciaDelApplet.Fuente}),
			codigo: codigoInesperado,
		},
		{
			nombre: "la salida estándar no se pudo escribir",
			doble:  dobleRoto(),
			res:    resultadoQueObserva(procedenciaDelApplet),
			codigo: codigoInesperado,
		},
	}

	for _, fallo := range fallos {
		almacen := &almacenEspia{}

		codigo := emitirConGrafo(t.Context(), fallo.doble, true, almacen, fallo.res, fallo.err)

		assert.Equal(t, fallo.codigo, codigo, fallo.nombre)
		assert.Empty(t, almacen.lotes, "%s: un fallo no entrega nada", fallo.nombre)
	}
}

// compruebaEntregaFallida exige que una entrega que falla no cambie ni el código
// ni la salida estándar y deje en la salida de error exactamente una línea, la
// del contrato, con los saltos de línea de la causa como espacios (FR-033,
// research.md D7).
func compruebaEntregaFallida(t *testing.T) {
	t.Helper()

	for _, enJSON := range []bool{true, false} {
		res := resultadoQueObserva(procedenciaDelApplet)

		doble := &presentadorConJSON{}
		almacen := &almacenEspia{fallo: errEntregaRota}

		codigo := emitirConGrafo(t.Context(), doble, enJSON, almacen, res, nil)

		assert.Equal(t, 0, codigo, "el sobre ya dijo ok: el código sigue siendo 0 (json=%t)", enJSON)
		require.Len(t, almacen.lotes, 1)
		assert.Equal(t, []string{lineaDeEntregaRota}, doble.avisos, "json=%t", enJSON)
		assert.Equal(t, lineaDeEntregaRota+"\n", doble.errores.String(),
			"exactamente una línea más en la salida de error (json=%t)", enJSON)

		sinFallo := &presentadorConJSON{}
		require.Equal(t, 0, emitirConGrafo(t.Context(), sinFallo, enJSON, &almacenEspia{}, res, nil))
		assert.Equal(t, sinFallo.salida.String(), doble.salida.String(),
			"la salida estándar es la misma que con una entrega correcta (json=%t)", enJSON)
	}
}

// compruebaAvisoQueNoSeEscribe exige que el error al escribir la línea de aviso
// no cambie el código ni la salida estándar y que no se reintente (research.md
// D7).
func compruebaAvisoQueNoSeEscribe(t *testing.T) {
	t.Helper()

	res := resultadoQueObserva(procedenciaDelApplet)

	doble := &presentadorConAvisoRoto{}
	almacen := &almacenEspia{fallo: errEntregaRota}

	codigo := emitirConGrafo(t.Context(), doble, true, almacen, res, nil)

	assert.Equal(t, 0, codigo, "la escritura del aviso no se propaga")
	assert.Equal(t, []string{lineaDeEntregaRota}, doble.avisos, "se intenta una vez y no se reintenta")
	assert.Equal(t, salidaSinGrafo(t, res), doble.salida.String())
}

// salidaSinGrafo es lo que presenta con --json el montador de las tablas, sin
// almacén, para el mismo resultado: la referencia de que el grafo no cambia la
// salida estándar.
func salidaSinGrafo(t *testing.T, res schema.Resultado) string {
	t.Helper()

	doble := &presentadorConJSON{}
	require.Equal(t, 0, montadorDePrueba().Emitir(t.Context(), doble, true, res, nil))

	return doble.salida.String()
}
