package boe

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cache"
	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/httpx"
)

// TestErrorDeBoeMensajes fija las formas de mensaje del contrato
// errores-y-codigos §2 y la clase y el código de salida que el kernel da a cada
// constructor, también cuando el error llega envuelto con %w (FR-100, D10).
//
// Las formas se fijan por igualdad literal: el mensaje es el motivo; si hay
// dirección, la dirección entre paréntesis detrás; y si la causa es un error con
// clase —el de httpx o el de la caché—, su texto detrás de dos puntos, porque el
// contrato obliga a conservarlo. La lista «nombra» es lo que la fila del
// contrato exige que el mensaje diga, y se comprueba aparte para que reescribir
// una frase no pueda hacer desaparecer en silencio un dato que hace falta para
// entender el fallo sin leer el registro. Y en ninguna forma entra el detalle
// técnico de una causa sin clase, que es del registro y no del sobre.
//
// Las direcciones son las de la API que el adaptador pedirá (data-model.md §8 y
// contrato verbos-y-salidas) y el instante uno fijo con nanosegundos: ninguna
// prueba de este fichero abre ninguna conexión ni depende del reloj.
func TestErrorDeBoeMensajes(t *testing.T) {
	t.Parallel()

	const (
		api                             = "https://www.boe.es/datosabiertos/api/legislacion-consolidada"
		direccionDelBloque              = api + "/id/BOE-A-2015-10565/texto/bloque/a21"
		direccionDeUnBloqueInexistente  = api + "/id/BOE-A-2015-10565/texto/bloque/a9999"
		direccionDeLosMetadatos         = api + "/id/BOE-A-2015-10565/metadatos"
		direccionDeUnaNormaInexistente  = api + "/id/BOE-A-2099-99999/metadatos"
		claveDeLosMetadatos             = "boe.legislacion-consolidada|1|metadatos|" + direccionDeLosMetadatos
		sinDeclarar                     = "el adaptador del BOE ha fallado sin declarar el motivo"
		motivoDeLaEntradaIlegible       = "la entrada de la caché con la clave %q no se puede leer"
		motivoDelBloqueSinVigencia      = "el bloque a21 se obtuvo, pero no se pudo comprobar su vigencia: %w"
		motivoDelBloqueEnSuPosicion     = "el bloque 2 de la lista, a9999: %w"
		motivoDeLosMetadatosIlegibles   = "el cuerpo de los metadatos no es JSON legible"
		motivoDelBloqueQueNoEsta        = "la norma BOE-A-2015-10565 no tiene el bloque a9999"
		motivoDeLaAusenciaEnSoloLectura = "con --offline no hay entrada vigente de metadatos con la clave %q"
	)

	instante := time.Date(2026, time.September, 13, 10, 30, 0, 123456789, time.UTC)
	causaTecnica := errors.New("XML syntax error on line 3: unexpected EOF")

	// Los dos errores de adaptador son reales y se obtienen sin red ni disco: una
	// hora nula para httpx y un directorio vacío para la caché, que las dos
	// rechazan al construir con «argumentos».
	_, deHTTPX := httpx.New(httpx.ConHora(nil))
	require.Error(t, deHTTPX)
	require.Equal(t, schema.ClaseArgumentos, cli.Clasificar(deHTTPX))

	_, deLaCache := cache.New(t.Context(), cache.ConDirectorio(""))
	require.Error(t, deLaCache)
	require.Equal(t, schema.ClaseArgumentos, cli.Clasificar(deLaCache))

	formas := []struct {
		nombre  string
		fallo   error
		mensaje string
		nombra  []string
		clase   schema.Clase
	}{
		{
			nombre: "filas 2 y 3: la norma, con el valor recibido y la forma esperada",
			fallo:  ValidarNorma("BOE-A-2015"),
			mensaje: `la norma "BOE-A-2015" no tiene la forma BOE-A-<año>-<número>, ` +
				`con cuatro dígitos en el año y de uno a nueve en el número`,
			nombra: []string{`"BOE-A-2015"`, "BOE-A-<año>-<número>"},
			clase:  schema.ClaseArgumentos,
		},
		{
			nombre: "filas 2 y 3: el bloque, con el valor recibido y la forma esperada",
			fallo:  ValidarBloque("a21/x"),
			mensaje: `el bloque "a21/x" no tiene la forma de un id de bloque: ` +
				`de 1 a 64 caracteres, el primero letra o dígito ASCII y los demás letras, ` +
				`dígitos, guiones o puntos, como a21, da3, preambulo, a1-30 o a85bis.`,
			nombra: []string{`"a21/x"`, "el primero letra o dígito ASCII"},
			clase:  schema.ClaseArgumentos,
		},
		{
			nombre:  "fila 4: sin dirección, el mensaje es el motivo",
			fallo:   errorDeArgumentos("", time.Time{}, nil, "la búsqueda no tiene ninguna palabra"),
			mensaje: "la búsqueda no tiene ninguna palabra",
			nombra:  []string{"ninguna palabra"},
			clase:   schema.ClaseArgumentos,
		},
		{
			nombre: "fila 5: --offline, el verbo y la clave, con la dirección del recurso que falta",
			fallo: errorDeFuenteNoDisponible(direccionDeLosMetadatos, time.Time{}, nil,
				fmt.Sprintf(motivoDeLaAusenciaEnSoloLectura, claveDeLosMetadatos)),
			mensaje: fmt.Sprintf(motivoDeLaAusenciaEnSoloLectura, claveDeLosMetadatos) +
				" (" + direccionDeLosMetadatos + ")",
			nombra: []string{"--offline", "metadatos", claveDeLosMetadatos},
			clase:  schema.ClaseFuenteNoDisponible,
		},
		{
			nombre:  "fila 6: la norma y el bloque, con la dirección del bloque",
			fallo:   errorDeNoEncontrado(direccionDeUnBloqueInexistente, instante, nil, motivoDelBloqueQueNoEsta),
			mensaje: motivoDelBloqueQueNoEsta + " (" + direccionDeUnBloqueInexistente + ")",
			nombra:  []string{"BOE-A-2015-10565", "a9999"},
			clase:   schema.ClaseNoEncontrado,
		},
		{
			nombre: "filas 7 y 8: la norma y el recurso, con la dirección del recurso",
			fallo: errorDeNoEncontrado(direccionDeUnaNormaInexistente, instante, nil,
				"la norma BOE-A-2099-99999 no tiene metadatos"),
			mensaje: "la norma BOE-A-2099-99999 no tiene metadatos (" + direccionDeUnaNormaInexistente + ")",
			nombra:  []string{"BOE-A-2099-99999", "metadatos"},
			clase:   schema.ClaseNoEncontrado,
		},
		{
			nombre: "filas 9 y 10: el estado HTTP y la dirección",
			fallo: errorDeFuenteNoDisponible(direccionDeLosMetadatos, instante, nil,
				"la fuente ha respondido con el estado 403"),
			mensaje: "la fuente ha respondido con el estado 403 (" + direccionDeLosMetadatos + ")",
			nombra:  []string{"403", direccionDeLosMetadatos},
			clase:   schema.ClaseFuenteNoDisponible,
		},
		{
			nombre: "filas 16 y 17: qué no se pudo interpretar y la dirección, sin el detalle técnico",
			fallo: errorDeFuenteNoDisponible(direccionDelBloque, instante, causaTecnica,
				"el cuerpo del bloque no es XML legible"),
			mensaje: "el cuerpo del bloque no es XML legible (" + direccionDelBloque + ")",
			nombra:  []string{"cuerpo del bloque", direccionDelBloque},
			clase:   schema.ClaseFuenteNoDisponible,
		},
		{
			nombre: "fila 18: el bloque obtenido y que la vigencia no se pudo comprobar, envuelto con %w",
			fallo: fmt.Errorf(motivoDelBloqueSinVigencia, errorDeFuenteNoDisponible(direccionDeLosMetadatos,
				instante, causaTecnica, motivoDeLosMetadatosIlegibles)),
			mensaje: "el bloque a21 se obtuvo, pero no se pudo comprobar su vigencia: " +
				motivoDeLosMetadatosIlegibles + " (" + direccionDeLosMetadatos + ")",
			nombra: []string{"a21", "vigencia", direccionDeLosMetadatos},
			clase:  schema.ClaseFuenteNoDisponible,
		},
		{
			nombre: "fila 19: la posición y el id del bloque que falló, envuelto con %w",
			fallo: fmt.Errorf(motivoDelBloqueEnSuPosicion, errorDeNoEncontrado(direccionDeUnBloqueInexistente,
				instante, nil, motivoDelBloqueQueNoEsta)),
			mensaje: "el bloque 2 de la lista, a9999: " + motivoDelBloqueQueNoEsta +
				" (" + direccionDeUnBloqueInexistente + ")",
			nombra: []string{"bloque 2", "a9999"},
			clase:  schema.ClaseNoEncontrado,
		},
		{
			nombre: "fila 20: la clave de la entrada que no se puede leer, sin el detalle técnico",
			fallo: errorInesperado(direccionDeLosMetadatos, time.Time{}, causaTecnica,
				fmt.Sprintf(motivoDeLaEntradaIlegible, claveDeLosMetadatos)),
			mensaje: fmt.Sprintf(motivoDeLaEntradaIlegible, claveDeLosMetadatos) +
				" (" + direccionDeLosMetadatos + ")",
			nombra: []string{claveDeLosMetadatos},
			clase:  schema.ClaseInesperado,
		},
		{
			nombre: "httpx conserva su texto, detrás del contexto del verbo",
			fallo: errorDeArgumentos(direccionDelBloque, time.Time{}, deHTTPX,
				"no se ha podido pedir el bloque"),
			mensaje: "no se ha podido pedir el bloque (" + direccionDelBloque + "): " + deHTTPX.Error(),
			nombra:  []string{deHTTPX.Error()},
			clase:   schema.ClaseArgumentos,
		},
		{
			nombre: "la caché conserva su texto, detrás del contexto del verbo",
			fallo: errorDeArgumentos(direccionDeLosMetadatos, time.Time{}, deLaCache,
				"no se ha podido abrir la caché"),
			mensaje: "no se ha podido abrir la caché (" + direccionDeLosMetadatos + "): " + deLaCache.Error(),
			nombra:  []string{deLaCache.Error()},
			clase:   schema.ClaseArgumentos,
		},
		{
			nombre: "la causa con clase conserva también el contexto que la envuelve",
			fallo: errorDeArgumentos(direccionDeLosMetadatos, time.Time{},
				fmt.Errorf("al cerrar la caché: %w", deLaCache), "no se ha podido terminar la consulta"),
			mensaje: "no se ha podido terminar la consulta (" + direccionDeLosMetadatos + "): " +
				"al cerrar la caché: " + deLaCache.Error(),
			nombra: []string{"al cerrar la caché", deLaCache.Error()},
			clase:  schema.ClaseArgumentos,
		},
		{
			nombre:  "un *Error nulo no deja el mensaje vacío",
			fallo:   (*Error)(nil),
			mensaje: sinDeclarar,
			clase:   schema.ClaseInesperado,
		},
		{
			nombre:  "un Error construido a cero desde fuera, tampoco",
			fallo:   &Error{},
			mensaje: sinDeclarar,
			clase:   schema.ClaseInesperado,
		},
		{
			nombre:  "sin motivo pero con dirección, la dirección sigue en el mensaje",
			fallo:   &Error{URL: direccionDelBloque, Causa: causaTecnica},
			mensaje: sinDeclarar + " (" + direccionDelBloque + ")",
			nombra:  []string{direccionDelBloque},
			clase:   schema.ClaseInesperado,
		},
	}

	for _, forma := range formas {
		t.Run(forma.nombre, func(t *testing.T) {
			t.Parallel()

			mensaje := forma.fallo.Error()

			assert.Equal(t, forma.mensaje, mensaje)
			for _, dato := range forma.nombra {
				assert.Contains(t, mensaje, dato)
			}
			assert.NotContains(t, mensaje, causaTecnica.Error())
			assert.Equal(t, forma.clase, cli.Clasificar(forma.fallo))
		})
	}

	t.Run("el instante no entra en el mensaje", func(t *testing.T) {
		t.Parallel()

		sinInstante := errorDeFuenteNoDisponible(direccionDelBloque, time.Time{}, nil, motivoDeLosMetadatosIlegibles)
		conInstante := errorDeFuenteNoDisponible(direccionDelBloque, instante, nil, motivoDeLosMetadatosIlegibles)

		assert.Equal(t, sinInstante.Error(), conInstante.Error())
		assert.NotContains(t, conInstante.Error(), "2026")
	})

	t.Run("la clase y el código de cada constructor, también envuelto con %w", func(t *testing.T) {
		t.Parallel()

		constructores := []struct {
			nombre    string
			construir func(direccion string, instante time.Time, causa error, motivo string) *Error
			clase     schema.Clase
			codigo    int
		}{
			{nombre: "argumentos", construir: errorDeArgumentos, clase: schema.ClaseArgumentos, codigo: 2},
			{nombre: "no encontrado", construir: errorDeNoEncontrado, clase: schema.ClaseNoEncontrado, codigo: 3},
			{
				nombre: "fuente no disponible", construir: errorDeFuenteNoDisponible,
				clase:  schema.ClaseFuenteNoDisponible,
				codigo: 4,
			},
			{nombre: "límite o TOS", construir: errorDeLimiteOTos, clase: schema.ClaseLimiteOTos, codigo: 5},
			{nombre: "inesperado", construir: errorInesperado, clase: schema.ClaseInesperado, codigo: 1},
		}

		// Los cinco constructores cubren las cinco clases que una fuente pública
		// puede producir, cada una una vez, y ninguno la de identidad humana ni la
		// de conflicto: ningún verbo de boe termina con el código 6 (FR-100) ni
		// con el 7, porque boe no tiene estado local que le impida actuar (ADR
		// 0023).
		clasesDeLaFuente := slices.DeleteFunc(schema.Clases(), func(clase schema.Clase) bool {
			return clase == schema.ClaseIdentidadHumana || clase == schema.ClaseConflicto
		})
		declaradas := make([]schema.Clase, 0, len(constructores))
		for _, constructor := range constructores {
			declaradas = append(declaradas, constructor.clase)
		}
		assert.ElementsMatch(t, clasesDeLaFuente, declaradas)

		for _, constructor := range constructores {
			t.Run(constructor.nombre, func(t *testing.T) {
				t.Parallel()

				fallo := constructor.construir(direccionDelBloque, instante, causaTecnica, motivoDeLosMetadatosIlegibles)
				unaCapa := fmt.Errorf("articulo BOE-A-2015-10565 a21: %w", fallo)
				dosCapas := fmt.Errorf("el applet no ha podido terminar: %w", unaCapa)

				assert.Equal(t, constructor.clase, fallo.Clase())
				for _, err := range []error{fallo, unaCapa, dosCapas} {
					assert.Equal(t, constructor.clase, cli.Clasificar(err))
					assert.Equal(t, constructor.codigo, cli.CodigoSalida(err))
				}

				// El error sigue alcanzable con sus datos bajo las dos capas, que es
				// de donde el applet toma la dirección y el instante del sobre de
				// fallo (FR-101, FR-096), y la causa sigue alcanzable a través de él.
				var recuperado *Error
				require.ErrorAs(t, dosCapas, &recuperado)
				assert.Same(t, fallo, recuperado)
				assert.Equal(t, direccionDelBloque, recuperado.URL)
				assert.Equal(t, instante, recuperado.Instante)
				assert.Same(t, causaTecnica, recuperado.Unwrap())
				require.ErrorIs(t, dosCapas, causaTecnica)
			})
		}
	})

	t.Run("sin causa, Unwrap no devuelve nada", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, errorDeNoEncontrado(direccionDelBloque, instante, nil, motivoDelBloqueQueNoEsta).Unwrap())
		require.NoError(t, (*Error)(nil).Unwrap())
	})
}

// TestClasesDeErrorDeBoe fija que ninguna ruta de fallo de la fuente queda sin
// clase (contrato errores-y-codigos §1 y §3; FR-090, FR-100; SC-008). Cada fila
// 2-4, 6-10 y 16-22 de la tabla cerrada es una subprueba, y cada situación de la
// fila se provoca con el código ya existente: los validadores, un Pedidor de
// prueba, las grabaciones y los sintéticos, una entrada de caché ilegible, una
// AperturaDeCache de prueba y Nueva sin dependencias. Sobre el error real que
// entrega la fuente, la prueba exige:
//
//   - que provocar la situación no entre en pánico;
//   - la clase que le da cli.Clasificar y el código de cli.CodigoSalida, que
//     nunca es el 6 de la identidad humana;
//   - un fragmento del mensaje que solo da esa situación, para que el caso no
//     pase por otro fallo de la misma clase;
//   - y el sobre de fallo de la fila. Si no se ha consultado nada, va sin
//     procedencia, y lo firma y lo fecha el kernel. Si no, lleva la fuente del
//     BOE, la dirección del recurso y el instante de la petición que falló; sin
//     petición no hay instante, y la fecha la pone el montaje.
//
// Las demás filas las fijan las pruebas que nombra el contrato §3.
func TestClasesDeErrorDeBoe(t *testing.T) {
	t.Parallel()

	const (
		formaDeLaNorma = "no tiene la forma BOE-A-<año>-<número>"
		formaDelBloque = "no tiene la forma de un id de bloque"
		sinPalabras    = "la búsqueda no tiene ninguna palabra"
		sinVigencia    = "el bloque a21 se obtuvo, pero no se pudo comprobar su vigencia: "
	)

	grabaciones := reproduce(carpetaDeLasGrabaciones)
	dataVacio := respondeConElCuerpo(`{"data": []}`)
	articulo21 := ConsultaArticulo{Norma: normaVigente, Bloque: bloqueDelArticulo21}
	busqueda := ConsultaBuscar{Texto: []string{"procedimiento", "administrativo", "común"}}

	filas := []filaDeLaTabla{
		{
			numero:    2,
			situacion: "norma-fuera-de-su-gramatica",
			casos: []casoDeFallo{
				{
					nombre:   "metadatos",
					provocar: rechazada(ConsultaMetadatos{Norma: "BOE-A-2015"}),
					clase:    schema.ClaseArgumentos,
					codigo:   2,
					nombra:   formaDeLaNorma,
				},
				{
					nombre:   "indice",
					provocar: rechazada(ConsultaIndice{Norma: "BOE-A-15-10565"}),
					clase:    schema.ClaseArgumentos,
					codigo:   2,
					nombra:   formaDeLaNorma,
				},
				{
					nombre:   "analisis",
					provocar: rechazada(ConsultaAnalisis{Norma: "BOE-A-2015-1234567890"}),
					clase:    schema.ClaseArgumentos,
					codigo:   2,
					nombra:   formaDeLaNorma,
				},
				{
					nombre:   "articulo",
					provocar: rechazada(ConsultaArticulo{Norma: "BOE-A-2015-1056a", Bloque: bloqueDelArticulo21}),
					clase:    schema.ClaseArgumentos,
					codigo:   2,
					nombra:   formaDeLaNorma,
				},
				{
					nombre:   "articulos",
					provocar: rechazada(ConsultaArticulos{Norma: "BOE-B-2015-10565", Bloques: []string{bloqueDelArticulo21}}),
					clase:    schema.ClaseArgumentos,
					codigo:   2,
					nombra:   formaDeLaNorma,
				},
			},
		},
		{
			numero:    3,
			situacion: "bloque-fuera-de-su-gramatica",
			casos: []casoDeFallo{
				{
					nombre:   "articulo",
					provocar: rechazada(ConsultaArticulo{Norma: normaVigente, Bloque: "a21/x"}),
					clase:    schema.ClaseArgumentos,
					codigo:   2,
					nombra:   formaDelBloque,
				},
				{
					nombre: "el-ultimo-de-articulos",
					provocar: rechazada(ConsultaArticulos{
						Norma:   normaVigente,
						Bloques: []string{bloqueDelArticulo21, bloqueDelArticulo22, "../a23"},
					}),
					clase:  schema.ClaseArgumentos,
					codigo: 2,
					nombra: formaDelBloque,
				},
			},
		},
		{
			numero:    4,
			situacion: "busqueda-sin-ninguna-palabra",
			casos: []casoDeFallo{
				{
					nombre:   "vacia",
					provocar: rechazada(ConsultaBuscar{}),
					clase:    schema.ClaseArgumentos,
					codigo:   2,
					nombra:   sinPalabras,
				},
				{
					nombre:   "solo-espacio-en-blanco",
					provocar: rechazada(ConsultaBuscar{Texto: []string{" ", "\t"}}),
					clase:    schema.ClaseArgumentos,
					codigo:   2,
					nombra:   sinPalabras,
				},
			},
		},
		{
			numero:    6,
			situacion: "404-de-un-bloque",
			casos: []casoDeFallo{
				{
					nombre:    "articulo",
					provocar:  pedida(grabaciones, ConsultaArticulo{Norma: normaVigente, Bloque: bloqueInexistente}),
					clase:     schema.ClaseNoEncontrado,
					codigo:    3,
					direccion: direccionDelBloqueInexistente,
					nombra:    "la norma BOE-A-2015-10565 no tiene el bloque a9999",
				},
			},
		},
		{
			numero:    7,
			situacion: "404-del-indice-los-metadatos-o-el-analisis",
			casos: []casoDeFallo{
				{
					nombre:    "indice",
					provocar:  pedida(grabaciones, ConsultaIndice{Norma: normaInexistente}),
					clase:     schema.ClaseNoEncontrado,
					codigo:    3,
					direccion: indiceInexistente,
					nombra:    "la norma BOE-A-2099-99999 no tiene índice",
				},
				{
					nombre:    "metadatos",
					provocar:  pedida(grabaciones, ConsultaMetadatos{Norma: normaInexistente}),
					clase:     schema.ClaseNoEncontrado,
					codigo:    3,
					direccion: metadatosInexistente,
					nombra:    "la norma BOE-A-2099-99999 no tiene metadatos",
				},
				{
					nombre:    "analisis",
					provocar:  pedida(grabaciones, ConsultaAnalisis{Norma: normaInexistente}),
					clase:     schema.ClaseNoEncontrado,
					codigo:    3,
					direccion: analisisInexistente,
					nombra:    "la norma BOE-A-2099-99999 no tiene análisis",
				},
			},
		},
		{
			numero:    8,
			situacion: "data-vacio-del-indice-los-metadatos-o-el-analisis",
			casos: []casoDeFallo{
				{
					nombre:    "indice",
					provocar:  pedida(dataVacio, ConsultaIndice{Norma: normaVigente}),
					clase:     schema.ClaseNoEncontrado,
					codigo:    3,
					direccion: indiceVigente,
					nombra:    "la norma BOE-A-2015-10565 no tiene índice",
				},
				{
					nombre:    "metadatos",
					provocar:  pedida(dataVacio, ConsultaMetadatos{Norma: normaVigente}),
					clase:     schema.ClaseNoEncontrado,
					codigo:    3,
					direccion: metadatosVigente,
					nombra:    "la norma BOE-A-2015-10565 no tiene metadatos",
				},
				{
					nombre:    "analisis",
					provocar:  pedida(dataVacio, ConsultaAnalisis{Norma: normaVigente}),
					clase:     schema.ClaseNoEncontrado,
					codigo:    3,
					direccion: analisisVigente,
					nombra:    "la norma BOE-A-2015-10565 no tiene análisis",
				},
			},
		},
		{
			numero:    9,
			situacion: "404-de-buscar",
			casos: []casoDeFallo{
				{
					nombre:    "buscar",
					provocar:  pedida(respondeConElEstado(404), busqueda),
					clase:     schema.ClaseFuenteNoDisponible,
					codigo:    4,
					direccion: busquedaConResultados,
					nombra:    "la fuente ha respondido con el estado 404 a la petición de la búsqueda",
				},
			},
		},
		{
			numero:    10,
			situacion: "otro-estado-que-no-es-2xx",
			casos: []casoDeFallo{
				{
					nombre:    "400-de-los-metadatos",
					provocar:  pedida(respondeConElEstado(400), ConsultaMetadatos{Norma: normaVigente}),
					clase:     schema.ClaseFuenteNoDisponible,
					codigo:    4,
					direccion: metadatosVigente,
					nombra:    "con el estado 400 a la petición de los metadatos",
				},
				{
					nombre:    "401-del-indice",
					provocar:  pedida(respondeConElEstado(401), ConsultaIndice{Norma: normaVigente}),
					clase:     schema.ClaseFuenteNoDisponible,
					codigo:    4,
					direccion: indiceVigente,
					nombra:    "con el estado 401 a la petición del índice",
				},
				{
					nombre:    "403-del-analisis",
					provocar:  pedida(respondeConElEstado(403), ConsultaAnalisis{Norma: normaVigente}),
					clase:     schema.ClaseFuenteNoDisponible,
					codigo:    4,
					direccion: analisisVigente,
					nombra:    "con el estado 403 a la petición del análisis",
				},
				{
					nombre:    "405-del-bloque",
					provocar:  pedida(respondeConElEstado(405), articulo21),
					clase:     schema.ClaseFuenteNoDisponible,
					codigo:    4,
					direccion: direccionDelArticulo21,
					nombra:    "con el estado 405 a la petición del bloque a21",
				},
				{
					nombre:    "406-de-la-busqueda",
					provocar:  pedida(respondeConElEstado(406), busqueda),
					clase:     schema.ClaseFuenteNoDisponible,
					codigo:    4,
					direccion: busquedaConResultados,
					nombra:    "con el estado 406 a la petición de la búsqueda",
				},
				{
					nombre:    "410-del-indice-de-una-norma-inexistente",
					provocar:  pedida(respondeConElEstado(410), ConsultaIndice{Norma: normaInexistente}),
					clase:     schema.ClaseFuenteNoDisponible,
					codigo:    4,
					direccion: indiceInexistente,
					nombra:    "con el estado 410 a la petición del índice de la norma BOE-A-2099-99999",
				},
			},
		},
		{
			numero:    16,
			situacion: "bloque-que-no-se-puede-leer",
			casos: []casoDeFallo{
				{
					nombre: "cuerpo-que-no-es-utf-8",
					provocar: pedida(respondeConElCuerpo(respuestaConBloque("<bloque titulo=\"T\"><p>Obligaci\xf3n.</p></bloque>")),
						articulo21),
					clase:     schema.ClaseFuenteNoDisponible,
					codigo:    4,
					direccion: direccionDelArticulo21,
					nombra:    "el cuerpo no es UTF-8 válido",
				},
				{
					nombre:    "xml-mal-formado",
					provocar:  pedida(reproduce(sinteticoDelBloqueIlegible), articulo21),
					clase:     schema.ClaseFuenteNoDisponible,
					codigo:    4,
					direccion: direccionDelArticulo21,
					nombra:    "el cuerpo no es XML legible",
				},
				{
					nombre: "varias-raices",
					provocar: pedida(respondeConElCuerpo("<response><data><bloque titulo=\"T\"><p>Texto.</p></bloque></data></response>\n"+
						"<response/>"), articulo21),
					clase:     schema.ClaseFuenteNoDisponible,
					codigo:    4,
					direccion: direccionDelArticulo21,
					nombra:    "el XML tiene más de un elemento raíz",
				},
				{
					nombre:    "ningun-bloque-bajo-la-raiz",
					provocar:  pedida(reproduce(sinteticoDelBloqueSinElemento), articulo21),
					clase:     schema.ClaseFuenteNoDisponible,
					codigo:    4,
					direccion: direccionDelArticulo21,
					nombra:    "el XML no tiene ningún elemento bloque por debajo de la raíz",
				},
			},
		},
		{
			numero:    17,
			situacion: "json-que-no-se-puede-leer",
			casos: []casoDeFallo{
				{
					nombre:    "json-ilegible",
					provocar:  pedida(reproduce(sinteticoDeLosMetadatosIlegibles), ConsultaMetadatos{Norma: normaVigente}),
					clase:     schema.ClaseFuenteNoDisponible,
					codigo:    4,
					direccion: metadatosVigente,
					nombra:    "el cuerpo no es JSON legible",
				},
				{
					nombre:    "raiz-que-no-es-objeto",
					provocar:  pedida(respondeConElCuerpo(`[{"id": "a21", "titulo": "Artículo 21"}]`), ConsultaIndice{Norma: normaVigente}),
					clase:     schema.ClaseFuenteNoDisponible,
					codigo:    4,
					direccion: indiceVigente,
					nombra:    "la raíz del JSON no es un objeto, sino una lista",
				},
				{
					nombre:    "elemento-del-analisis-que-no-es-objeto",
					provocar:  pedida(respondeConElCuerpo(`{"data": ["BOE-A-2015-10565"]}`), ConsultaAnalisis{Norma: normaVigente}),
					clase:     schema.ClaseFuenteNoDisponible,
					codigo:    4,
					direccion: analisisVigente,
					nombra:    "el primer elemento de data no es un objeto, sino una cadena",
				},
				{
					nombre: "numero-donde-se-espera-texto",
					provocar: pedida(respondeConElCuerpo(`{"data": [{"identificador": "BOE-A-2015-10565", "titulo": 39}]}`),
						busqueda),
					clase:     schema.ClaseFuenteNoDisponible,
					codigo:    4,
					direccion: busquedaConResultados,
					nombra:    `el campo "titulo" no es texto, sino un número`,
				},
			},
		},
		{
			numero:    18,
			situacion: "articulo-con-el-bloque-obtenido-y-los-metadatos-fallidos",
			casos: []casoDeFallo{
				{
					nombre:    "data-vacio-de-los-metadatos",
					provocar:  pedida(salvoEn(metadatosVigente, dataVacio), articulo21),
					clase:     schema.ClaseNoEncontrado,
					codigo:    3,
					direccion: metadatosVigente,
					nombra:    sinVigencia + "la norma BOE-A-2015-10565 no tiene metadatos",
				},
				{
					nombre:    "metadatos-caidos",
					provocar:  pedida(reproduce(sinteticoDeLosMetadatosCaidos), articulo21),
					clase:     schema.ClaseFuenteNoDisponible,
					codigo:    4,
					direccion: metadatosVigente,
					nombra:    sinVigencia + "ha fallado la petición de los metadatos",
				},
				{
					nombre:    "429-de-los-metadatos",
					provocar:  pedida(salvoEn(metadatosVigente, reproduce(sinteticoDelLimite)), articulo21),
					clase:     schema.ClaseLimiteOTos,
					codigo:    5,
					direccion: metadatosVigente,
					nombra:    sinVigencia + "ha fallado la petición de los metadatos",
				},
			},
		},
		{
			numero:    19,
			situacion: "articulos-con-un-bloque-o-sus-metadatos-fallidos",
			casos: []casoDeFallo{
				{
					nombre: "404-del-segundo-bloque",
					provocar: pedida(grabaciones, ConsultaArticulos{
						Norma:   normaVigente,
						Bloques: []string{bloqueDelArticulo21, bloqueInexistente, bloqueDelArticulo23},
					}),
					clase:     schema.ClaseNoEncontrado,
					codigo:    3,
					direccion: direccionDelBloqueInexistente,
					nombra:    "no se ha podido resolver el bloque a9999, en la posición 2 de 3: la norma BOE-A-2015-10565",
				},
				{
					nombre: "403-del-segundo-bloque",
					provocar: pedida(salvoEn(direccionDelArticulo22, respondeConElEstado(403)), ConsultaArticulos{
						Norma:   normaVigente,
						Bloques: []string{bloqueDelArticulo21, bloqueDelArticulo22, bloqueDelArticulo23},
					}),
					clase:     schema.ClaseFuenteNoDisponible,
					codigo:    4,
					direccion: direccionDelArticulo22,
					nombra:    "no se ha podido resolver el bloque a22, en la posición 2 de 3: la fuente ha respondido con el estado 403",
				},
				{
					nombre: "429-de-los-metadatos-del-primer-bloque",
					provocar: pedida(salvoEn(metadatosVigente, reproduce(sinteticoDelLimite)), ConsultaArticulos{
						Norma:   normaVigente,
						Bloques: []string{bloqueDelArticulo21, bloqueDelArticulo22},
					}),
					clase:     schema.ClaseLimiteOTos,
					codigo:    5,
					direccion: metadatosVigente,
					nombra:    "no se ha podido resolver el bloque a21, en la posición 1 de 2: " + sinVigencia,
				},
			},
		},
		{
			numero:    20,
			situacion: "entrada-de-cache-que-no-se-puede-leer",
			casos: []casoDeFallo{
				{
					nombre:    "metadatos",
					provocar:  conLaEntradaIlegible,
					clase:     schema.ClaseInesperado,
					codigo:    1,
					direccion: metadatosVigente,
					nombra:    `la entrada de la caché con la clave "boe.legislacion-consolidada|1|metadatos|` + metadatosVigente + `" no se puede leer`,
				},
			},
		},
		{
			numero:    21,
			situacion: "fallo-de-la-cache",
			casos:     casosDeLaFila21(),
		},
		{
			numero:    22,
			situacion: "dependencia-ausente-o-consulta-de-otro-tipo",
			casos: []casoDeFallo{
				{
					nombre:   "nueva-sin-dependencias",
					provocar: sinDependencias,
					clase:    schema.ClaseInesperado,
					codigo:   1,
					nombra:   "la fuente del BOE no se puede componer: faltan las dependencias ConCliente y ConCache",
				},
				{
					nombre:   "fuente-sin-componer",
					provocar: sinComponer,
					clase:    schema.ClaseInesperado,
					codigo:   1,
					nombra:   "la fuente del BOE no se puede componer: la fuente no se ha construido con Nueva",
				},
				{
					nombre:   "consulta-de-otro-tipo",
					provocar: rechazada(consultaAjena{verbo: verboMetadatos}),
					clase:    schema.ClaseInesperado,
					codigo:   1,
					nombra:   "la fuente del BOE no resuelve consultas de tipo boe.consultaAjena",
				},
			},
		},
	}

	// Una subprueba por cada fila que el contrato §3 asigna a esta prueba, y
	// ninguna de otra.
	numeros := make([]int, 0, len(filas))
	for _, fila := range filas {
		numeros = append(numeros, fila.numero)
	}

	require.Equal(t, []int{2, 3, 4, 6, 7, 8, 9, 10, 16, 17, 18, 19, 20, 21, 22}, numeros)

	for _, fila := range filas {
		t.Run(fmt.Sprintf("fila-%02d-%s", fila.numero, fila.situacion), func(t *testing.T) {
			t.Parallel()

			require.NotEmpty(t, fila.casos)

			for _, caso := range fila.casos {
				t.Run(caso.nombre, func(t *testing.T) {
					t.Parallel()

					caso.comprueba(t)
				})
			}
		})
	}
}

// filaDeLaTabla es una fila del contrato errores-y-codigos §1 con las situaciones
// que la provocan.
type filaDeLaTabla struct {
	numero    int
	situacion string
	casos     []casoDeFallo
}

// casoDeFallo es una situación de una fila, provocada con el código ya
// existente, y lo que la fila fija de su fallo.
type casoDeFallo struct {
	nombre string
	// provocar provoca la situación y devuelve lo que entregó la fuente. Exige
	// además lo que solo esa situación fija.
	provocar func(t *testing.T) falloProvocado
	// clase y codigo son los de la fila.
	clase  schema.Clase
	codigo int
	// direccion es la url del sobre de fallo. Va vacía cuando el sobre lo firma
	// el kernel, porque no se ha consultado nada.
	direccion string
	// nombra es un fragmento del mensaje que solo da esta situación, para que el
	// caso no pase por otro fallo de la misma clase.
	nombra string
}

// falloProvocado es lo que entrega la fuente al provocar una situación.
type falloProvocado struct {
	resultado schema.Resultado
	err       error
	// instante es el de emisión de la petición que falló, el del reloj de la
	// prueba. Va a cero cuando no hubo ninguna, y entonces la fecha del sobre la
	// pone el montaje del kernel.
	instante time.Time
}

// comprueba provoca la situación del caso y exige, sobre el error que entrega la
// fuente: que provocarla no entre en pánico; la clase y el código de la fila,
// nunca el de identidad humana; un mensaje que nombra la situación; y el sobre de
// la fila. Sin dirección, el sobre va sin procedencia. Con ella, lleva la fuente
// del BOE, la dirección y el instante de la petición que falló, o ninguno.
func (caso casoDeFallo) comprueba(t *testing.T) {
	t.Helper()

	var provocado falloProvocado

	require.NotPanics(t, func() { provocado = caso.provocar(t) })

	err := provocado.err
	require.Error(t, err)
	assert.Equal(t, caso.clase, cli.Clasificar(err))
	assert.Equal(t, caso.codigo, cli.CodigoSalida(err))
	assert.NotEqual(t, codigoDeIdentidadHumana, cli.CodigoSalida(err))
	assert.Contains(t, err.Error(), caso.nombra)

	var fallo *Error
	require.ErrorAs(t, err, &fallo)
	assert.Equal(t, caso.direccion, fallo.URL)
	assert.Equal(t, provocado.instante, fallo.Instante)

	if caso.direccion == "" {
		assert.Zero(t, provocado.resultado, "el sobre de fallo lo firma y lo fecha el kernel")

		return
	}

	compruebaDireccionDeLaFuente(t, caso.direccion)
	assert.Equal(t, schema.Resultado{Procedencia: schema.Procedencia{
		Fuente:        NombreDeLaFuente,
		URL:           caso.direccion,
		FechaConsulta: provocado.instante,
	}}, provocado.resultado)
}

// rechazada es la situación de la consulta que la fuente rechaza sin abrir la
// caché ni construir el cliente.
func rechazada(consulta core.Consulta) func(*testing.T) falloProvocado {
	return func(t *testing.T) falloProvocado {
		t.Helper()

		dependencias := &dependenciasDePrueba{cacheAbierta: &cacheDePrueba{}}
		fuente := fuenteDePrueba(t, dependencias)

		resultado, err := fuente.Fetch(t.Context(), schema.Contexto{}, consulta)

		dependencias.compruebaSinUso(t)

		return falloProvocado{resultado: resultado, err: err}
	}
}

// pedida es la situación de la consulta que la fuente resuelve pidiendo al
// Pedidor de prueba que construye pedidor, con la caché real vacía en una carpeta
// temporal. Algo se pide, y el instante del fallo es el del reloj del banco, que
// no se mueve mientras se resuelve.
func pedida(pedidor func(*testing.T, *relojDePrueba) *pedidorDePrueba, consulta core.Consulta,
) func(*testing.T) falloProvocado {
	return func(t *testing.T) falloProvocado {
		t.Helper()

		banco := nuevoBanco(t, pedidor)

		resultado, err := banco.resuelve(t, schema.Contexto{}, consulta)

		assert.NotEmpty(t, banco.pedidor.peticiones, "la situación se provoca pidiendo")

		return falloProvocado{resultado: resultado, err: err, instante: banco.reloj.ahora()}
	}
}

// salvoEn es el Pedidor de prueba que reproduce las grabaciones salvo la petición
// de la dirección, a la que responde el Pedidor que construye otro con el mismo
// reloj: lo demás se obtiene de verdad y solo falla lo que la situación necesita.
func salvoEn(direccion string, otro func(*testing.T, *relojDePrueba) *pedidorDePrueba,
) func(*testing.T, *relojDePrueba) *pedidorDePrueba {
	return func(t *testing.T, reloj *relojDePrueba) *pedidorDePrueba {
		t.Helper()

		grabadas := clienteDeReproduccion(t, carpetaDeLasGrabaciones, reloj.ahora)
		excepcion := otro(t, reloj)

		return &pedidorDePrueba{
			responde: func(ctx context.Context, ec schema.Contexto, peticion httpx.Peticion) (httpx.Respuesta, error) {
				if peticion.URL == direccion {
					return excepcion.Pedir(ctx, ec, peticion)
				}

				return grabadas.Pedir(ctx, ec, peticion)
			},
		}
	}
}

// conLaEntradaIlegible es la situación de los metadatos de la Ley 39/2015 con una
// entrada vigente bajo su clave que no tiene la forma de la consulta guardada: no
// se construye ningún cliente ni se pide nada.
func conLaEntradaIlegible(t *testing.T) falloProvocado {
	t.Helper()

	const clave = "boe.legislacion-consolidada|1|metadatos|" + metadatosVigente

	banco := nuevoBanco(t, reproduce(t.TempDir()))

	escritora, err := cache.New(t.Context(), cache.ConDirectorio(banco.carpetaDeLaCache), cache.ConReloj(banco.reloj.ahora))
	require.NoError(t, err)
	require.NoError(t, escritora.Put(t.Context(), clave, []byte(`{"fecha_consulta":`), vigenciaDeLosMetadatos))
	require.NoError(t, escritora.Close())

	resultado, err := banco.resuelve(t, schema.Contexto{}, ConsultaMetadatos{Norma: normaVigente})

	banco.compruebaSinPedirNada(t)

	return falloProvocado{resultado: resultado, err: err}
}

// sinDependencias es la situación de Nueva sin ConCliente ni ConCache, que no
// compone ninguna fuente.
func sinDependencias(t *testing.T) falloProvocado {
	t.Helper()

	fuente, err := Nueva()

	assert.Nil(t, fuente)

	return falloProvocado{err: err}
}

// sinComponer es la situación de la consulta a una fuente construida a cero, sin
// Nueva, a la que le faltan las dos dependencias.
func sinComponer(t *testing.T) falloProvocado {
	t.Helper()

	resultado, err := (&Fuente{}).Fetch(t.Context(), schema.Contexto{}, ConsultaMetadatos{Norma: normaVigente})

	return falloProvocado{resultado: resultado, err: err}
}

// casosDeLaFila21 son las situaciones de la fila 21 con cada una de las tres
// clases que produce cache.Error. La apertura falla, sin pedir nada. La lectura
// falla fuera de solo lectura, sin pedir nada. La escritura falla después de
// obtener la respuesta, con una petición y ninguna entrada escrita. Y el cierre
// falla: tras una invocación correcta, con la clase del cierre; y tras una
// fallida, con la clase del fallo de la invocación y los dos errores unidos. En
// este último caso el cierre falla con otra clase, para que se vea cuál
// prevalece.
func casosDeLaFila21() []casoDeFallo {
	clases := []struct {
		clase  schema.Clase
		codigo int
	}{
		{clase: schema.ClaseArgumentos, codigo: 2},
		{clase: schema.ClaseFuenteNoDisponible, codigo: 4},
		{clase: schema.ClaseInesperado, codigo: 1},
	}

	var casos []casoDeFallo

	for indice, deLaCache := range clases {
		falla := func(operacion string) error {
			return &falloDeLaCacheDePrueba{operacion: operacion, clase: deLaCache.clase}
		}
		cierreDeOtraClase := &falloDeLaCacheDePrueba{operacion: "cerrar", clase: clases[(indice+1)%len(clases)].clase}

		situaciones := []struct {
			nombre string
			caso   casoDeLaCache
			motivo string
		}{
			{
				nombre: "la-apertura-falla",
				caso:   casoDeLaCache{errDeLaApertura: falla("abrir")},
				motivo: "no se ha podido abrir la caché",
			},
			{
				nombre: "get-falla-fuera-de-solo-lectura",
				caso:   casoDeLaCache{errDeLaLectura: falla("leer")},
				motivo: "no se ha podido leer la caché",
			},
			{
				nombre: "put-falla-tras-obtener-la-respuesta",
				caso:   casoDeLaCache{errDeLaEscritura: falla("escribir"), peticiones: 1},
				motivo: "no se ha podido escribir en la caché",
			},
			{
				nombre: "close-falla-tras-una-invocacion-correcta",
				caso:   casoDeLaCache{errDelCierre: falla("cerrar"), peticiones: 1, escritas: 1},
				motivo: "no se ha podido cerrar la caché",
			},
			{
				nombre: "close-falla-tras-una-invocacion-fallida",
				caso:   casoDeLaCache{errDeLaLectura: falla("leer"), errDelCierre: cierreDeOtraClase},
				motivo: "no se ha podido leer la caché",
			},
		}

		for _, situacion := range situaciones {
			casos = append(casos, casoDeFallo{
				nombre:    situacion.nombre + "-" + string(deLaCache.clase),
				provocar:  situacion.caso.provocar,
				clase:     deLaCache.clase,
				codigo:    deLaCache.codigo,
				direccion: metadatosVigente,
				nombra:    situacion.motivo,
			})
		}
	}

	return casos
}

// casoDeLaCache es una situación de la fila 21: metadatos BOE-A-2015-10565 sobre
// las grabaciones, con una AperturaDeCache de prueba cuya apertura, lectura,
// escritura o cierre fallan con los errores que se le dan.
type casoDeLaCache struct {
	errDeLaApertura  error
	errDeLaLectura   error
	errDeLaEscritura error
	errDelCierre     error
	// peticiones son las que la invocación entrega al cliente, y escritas, las
	// entradas que deja en la caché.
	peticiones int
	escritas   int
}

// provocar resuelve la consulta con la caché del caso, sin --offline ni
// --dry-run. Exige además que la caché se abra una sola vez y fuera de solo
// lectura, y que se cierre si se abrió; las peticiones y las entradas escritas
// del caso; y que cada error de la caché siga alcanzable con errors.Is.
func (caso casoDeLaCache) provocar(t *testing.T) falloProvocado {
	t.Helper()

	reloj := &relojDePrueba{inicio: time.Date(2026, time.September, 13, 10, 30, 0, 123456789, time.UTC)}
	pedidor := reproduce(carpetaDeLasGrabaciones)(t, reloj)
	abierta := &cacheQueFalla{
		errDeLaLectura:   caso.errDeLaLectura,
		errDeLaEscritura: caso.errDeLaEscritura,
		errDelCierre:     caso.errDelCierre,
		escritas:         map[string][]byte{},
	}

	var aperturas []bool

	fuente, err := Nueva(
		ConCliente(func() (Pedidor, error) { return pedidor, nil }),
		ConCache(func(_ context.Context, soloLectura bool) (CacheAbierta, error) {
			aperturas = append(aperturas, soloLectura)
			if caso.errDeLaApertura != nil {
				return nil, caso.errDeLaApertura
			}

			return abierta, nil
		}),
	)
	require.NoError(t, err)

	resultado, err := fuente.Fetch(t.Context(), schema.Contexto{}, ConsultaMetadatos{Norma: normaVigente})

	for _, deLaCache := range []error{caso.errDeLaApertura, caso.errDeLaLectura, caso.errDeLaEscritura, caso.errDelCierre} {
		if deLaCache != nil {
			require.ErrorIs(t, err, deLaCache)
		}
	}

	cierres := 1
	if caso.errDeLaApertura != nil {
		cierres = 0
	}

	assert.Equal(t, []bool{false}, aperturas, "la caché se abre una sola vez, y fuera de solo lectura")
	assert.Equal(t, cierres, abierta.cierres)
	assert.Len(t, pedidor.peticiones, caso.peticiones)
	assert.Len(t, abierta.escritas, caso.escritas)

	return falloProvocado{resultado: resultado, err: err}
}

// cacheQueFalla es la CacheAbierta de la fila 21: guarda en memoria lo que se
// escribe, cuenta los cierres y falla al leer, al escribir o al cerrar con el
// error que se le dé. No es segura para usarla desde varias goroutines a la vez:
// cada caso construye la suya.
type cacheQueFalla struct {
	errDeLaLectura   error
	errDeLaEscritura error
	errDelCierre     error
	escritas         map[string][]byte
	cierres          int
}

// Get da lo escrito bajo la clave, o falla con el error de la lectura.
func (c *cacheQueFalla) Get(_ context.Context, clave string) ([]byte, bool, error) {
	if c.errDeLaLectura != nil {
		return nil, false, c.errDeLaLectura
	}

	contenido, presente := c.escritas[clave]

	return contenido, presente, nil
}

// Put escribe el contenido bajo la clave, o falla con el error de la escritura
// sin escribir nada.
func (c *cacheQueFalla) Put(_ context.Context, clave string, contenido []byte, _ time.Duration) error {
	if c.errDeLaEscritura != nil {
		return c.errDeLaEscritura
	}

	c.escritas[clave] = contenido

	return nil
}

// Close cuenta el cierre y devuelve el error del cierre.
func (c *cacheQueFalla) Close() error {
	c.cierres++

	return c.errDelCierre
}

// La caché de prueba es una CacheAbierta de verdad, y que lo siga siendo no
// depende de que alguien lo recuerde.
var _ CacheAbierta = (*cacheQueFalla)(nil)

// falloDeLaCacheDePrueba es el error de la caché de prueba. Declara la clase que
// se le da, como cache.Error, y se usa por puntero: errors.Is lo alcanza por
// identidad, y dos fallos de la misma clase no se confunden.
type falloDeLaCacheDePrueba struct {
	operacion string
	clase     schema.Clase
}

// Error nombra la operación y la clase.
func (e *falloDeLaCacheDePrueba) Error() string {
	return "la caché de prueba ha fallado al " + e.operacion + " con la clase " + string(e.clase)
}

// Clase es la que se le dio.
func (e *falloDeLaCacheDePrueba) Clase() schema.Clase {
	return e.clase
}

var _ schema.ConClase = (*falloDeLaCacheDePrueba)(nil)
