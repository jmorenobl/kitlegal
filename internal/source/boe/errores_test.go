package boe

import (
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cache"
	"github.com/jmorenobl/kitlegal/internal/cli"
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
				`de 1 a 64 caracteres, todos letras o dígitos ASCII, como a21, da3 o preambulo`,
			nombra: []string{`"a21/x"`, "de 1 a 64 caracteres, todos letras o dígitos ASCII"},
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
				clase: schema.ClaseFuenteNoDisponible, codigo: 4,
			},
			{nombre: "límite o TOS", construir: errorDeLimiteOTos, clase: schema.ClaseLimiteOTos, codigo: 5},
			{nombre: "inesperado", construir: errorInesperado, clase: schema.ClaseInesperado, codigo: 1},
		}

		// Los cinco constructores cubren las cinco clases que una fuente pública
		// puede producir, cada una una vez, y ninguno la de identidad humana: ningún
		// verbo de boe termina con el código 6 (FR-100).
		clasesDeLaFuente := slices.DeleteFunc(schema.Clases(), func(clase schema.Clase) bool {
			return clase == schema.ClaseIdentidadHumana
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
