package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// casoDeArticulo es el nombre del único caso de la verificación contra la fuente
// real, con el que empieza todo error de verificarArticulo (FR-115; contrato
// esquemas-fixtures-y-controles §7).
const casoDeArticulo = "caso «boe articulo»"

// invocacionDeArticulo es la invocación que comprueba ese caso: la que hace una
// skill para citar un artículo, por el enlace boe del multicall. Cada llamada da
// una lista nueva.
func invocacionDeArticulo() []string {
	return []string{"boe", "articulo", normaDeBoe, "a21", "--json"}
}

// verificarArticulo es la verificación contra la fuente real (FR-115, SC-009;
// contrato esquemas-fixtures-y-controles §7; research.md D13): invoca
// «boe articulo BOE-A-2015-10565 a21 --json» por Main sobre el registro y
// comprueba la invocación con comprobarArticulo. Todo error suyo empieza por
// «caso «boe articulo»:», que es lo que lee en la incidencia quien mantiene el
// proyecto.
//
// Devuelve un error y no hace fallar ningún test, para que la misma comprobación
// la ejerzan TestVerificarFuentes contra la red y
// TestVerificacionDeFuentesDetectaCambios sin ella. No recibe contexto porque
// Main no lo recibe: el plazo de la invocación es el de --timeout.
func verificarArticulo(registro *Registro, esquema *jsonschema.Schema) error {
	var salida, errores bytes.Buffer

	codigo := Main(invocacionDeArticulo(), registro, &salida, &errores,
		versionDePrueba, commitDePrueba, fechaDePrueba)

	res := invocacionDePrueba{codigo: codigo, salida: salida.String(), errores: errores.String()}

	if err := comprobarArticulo(res, esquema); err != nil {
		return fmt.Errorf("%s: %w", casoDeArticulo, err)
	}

	return nil
}

// comprobarArticulo exige a una invocación de articulo lo que FR-115 pide a la
// fuente real: código 0, un sobre válido contra la parte articulo de
// schemas/bloque.json y un data.texto con algo más que blancos. Comprueba forma y
// no contenido: el texto de un artículo cambia legítimamente con cada reforma,
// así que contra la red ningún fixture sirve de referencia.
func comprobarArticulo(res invocacionDePrueba, esquema *jsonschema.Schema) error {
	if res.codigo != 0 {
		return fmt.Errorf("termina con código %d y no con 0; salida estándar: %s; salida de error: %s",
			res.codigo, strings.TrimSpace(res.salida), strings.TrimSpace(res.errores))
	}

	documento, err := jsonschema.UnmarshalJSON(strings.NewReader(res.salida))
	if err != nil {
		return fmt.Errorf("la salida estándar no es un único documento JSON: %w", err)
	}

	err = esquema.Validate(documento)
	if err != nil {
		return fmt.Errorf("el sobre no valida contra la parte articulo de schemas/bloque.json: %w", err)
	}

	// Validado el sobre con ok, data es un artículo y su texto, un texto.
	var sobre struct {
		Data struct {
			Texto string `json:"texto"`
		} `json:"data"`
	}

	err = json.Unmarshal([]byte(res.salida), &sobre)
	if err != nil {
		return fmt.Errorf("el data del sobre no es el de un artículo: %w", err)
	}

	if strings.TrimSpace(sobre.Data.Texto) == "" {
		return errors.New("data.texto está vacío")
	}

	return nil
}

// TestVerificacionDeFuentesDetectaCambios es el control negativo de SC-009 sin
// red (research.md D13): la comprobación que TestVerificarFuentes hace cada noche
// contra la fuente real pasa sobre las grabaciones y falla, nombrando el caso,
// sobre el sintético bloque-ilegible, una respuesta que ya no se interpreta.
//
// El resto demuestra sobre la invocación real, alterada como la dejaría una
// fuente que ha cambiado, que ninguna de las tres exigencias pasa en vacío.
func TestVerificacionDeFuentesDetectaCambios(t *testing.T) {
	t.Parallel()

	esquema := esquemaDeArticulo(t)

	t.Run("grabaciones", func(t *testing.T) {
		t.Parallel()

		banco := nuevoBancoDeBoe(t, grabacionesDeBoe)

		require.NoError(t, verificarArticulo(registroDeBoe(t, banco.dependencias), esquema))
	})

	t.Run("bloque-ilegible", func(t *testing.T) {
		t.Parallel()

		banco := nuevoBancoDeBoe(t, sintetico("bloque-ilegible"))
		err := verificarArticulo(registroDeBoe(t, banco.dependencias), esquema)

		require.ErrorContains(t, err, "termina con código 4 y no con 0")
		assert.True(t, strings.HasPrefix(err.Error(), casoDeArticulo+": "), "el error nombra el caso: %v", err)
	})

	grabada := nuevoBancoDeBoe(t, grabacionesDeBoe).invocar(t, invocacionDeArticulo()...)
	require.NoError(t, comprobarArticulo(grabada.invocacionDePrueba, esquema),
		"la invocación que se altera es una que pasa la comprobación")

	const (
		noValida   = "el sobre no valida contra la parte articulo de schemas/bloque.json"
		textoVacio = "data.texto está vacío"
	)

	casos := []struct {
		nombre     string
		invocacion invocacionDePrueba
		fallo      string
	}{
		{
			nombre:     "codigo-distinto-de-0",
			invocacion: invocacionDePrueba{codigo: 1, salida: grabada.salida, errores: grabada.errores},
			fallo:      "termina con código 1 y no con 0",
		},
		{
			nombre:     "no-es-json",
			invocacion: invocacionDePrueba{salida: "no es JSON"},
			fallo:      "la salida estándar no es un único documento JSON",
		},
		{
			nombre: "sin-texto",
			invocacion: invocacionDePrueba{salida: sobreAlterado(t, grabada.salida, func(data map[string]any) {
				delete(data, "texto")
			})},
			fallo: noValida,
		},
		{
			nombre: "texto-vacio",
			invocacion: invocacionDePrueba{salida: sobreAlterado(t, grabada.salida, func(data map[string]any) {
				data["texto"] = ""
			})},
			fallo: textoVacio,
		},
		{
			nombre: "texto-en-blanco",
			invocacion: invocacionDePrueba{salida: sobreAlterado(t, grabada.salida, func(data map[string]any) {
				data["texto"] = " \n\t "
			})},
			fallo: textoVacio,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			require.ErrorContains(t, comprobarArticulo(caso.invocacion, esquema), caso.fallo)
		})
	}
}

// esquemaDeArticulo es la salida de articulo compilada desde schemas/bloque.json,
// el fichero publicado, igual que la valida TestSalidaDeBoeContraSchemas.
func esquemaDeArticulo(t *testing.T) *jsonschema.Schema {
	t.Helper()

	publicado, id := ficheroPublicadoDelVerbo(t, "articulo")

	return salidaPublicada(t, publicado, id, "articulo")
}

// registroDeBoe es un registro con el applet boe sobre esas dependencias y
// ningún otro.
func registroDeBoe(t *testing.T, dependencias DependenciasDeBoe) *Registro {
	t.Helper()

	var registro Registro

	require.NoError(t, registro.Registrar(AppletBoe(dependencias)))

	return &registro
}

// sobreAlterado es la salida con el objeto de data cambiado por cambiar.
func sobreAlterado(t *testing.T, salida string, cambiar func(data map[string]any)) string {
	t.Helper()

	sobre := sobreValidable(t, salida)
	cambiar(objetoDeData(t, sobre))

	alterado, err := json.Marshal(sobre)
	require.NoError(t, err)

	return string(alterado)
}
