package plazos

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSalidaDePlazos comprueba la forma del data de calcular: las claves de
// contracts/applet-plazos.md §4 en su orden y ninguna más, la omisión solo de
// las opcionales, las listas vacías como [] y los enumerados de las etiquetas
// iguales a los vocabularios del dominio (FR-001; research D5, D29).
func TestSalidaDePlazos(t *testing.T) {
	t.Parallel()

	t.Run("las claves", func(t *testing.T) {
		t.Parallel()

		for tipo, claves := range map[reflect.Type][]string{
			reflect.TypeFor[Plazo](): {
				"fecha", "tipo", "preceptos", "municipios", "festivos", "avisos", "cobertura",
			},
			reflect.TypeFor[PreceptoDelPlazo]():  {"norma", "bloque", "apartado"},
			reflect.TypeFor[MunicipioDelPlazo](): {"codigo_ine", "nombre"},
			reflect.TypeFor[FestivoSaltado]():    {"fecha", "ambito", "municipio,omitempty", "source"},
			reflect.TypeFor[Aviso]():             {"codigo", "texto"},
			reflect.TypeFor[TramoDeCobertura](): {
				"ambito", "municipio,omitempty", "desde", "hasta", "disposicion,omitempty", "falta,omitempty",
			},
		} {
			assert.Equal(t, claves, etiquetasJSON(tipo), tipo.Name())
		}
	})
	t.Run("los enumerados", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, CodigosDeTipo(), enumeradoDelCampo(t, reflect.TypeFor[Plazo](), "Tipo"))
		assert.Equal(t, cadenas(Ambitos()), enumeradoDelCampo(t, reflect.TypeFor[FestivoSaltado](), "Ambito"))
		assert.Equal(t, CodigosDeAviso(), enumeradoDelCampo(t, reflect.TypeFor[Aviso](), "Codigo"))
		assert.Equal(t, cadenas(AmbitosDeCobertura()),
			enumeradoDelCampo(t, reflect.TypeFor[TramoDeCobertura](), "Ambito"))
		assert.Equal(t, cadenas(FaltasDeCobertura()),
			enumeradoDelCampo(t, reflect.TypeFor[TramoDeCobertura](), "Falta"))
	})
	t.Run("las listas vacías y las claves opcionales", func(t *testing.T) {
		t.Parallel()

		plazo, err := Calcular(calendarioDeLosCalculos(t), tiposDelContrato(t), Peticion{
			Tipo: TipoReposicion, Inicio: "2031-09-04", Municipios: los(villaPrimera, villaPrimera),
		})
		require.NoError(t, err)

		serializado, err := json.Marshal(plazo)
		require.NoError(t, err)
		assert.JSONEq(t, `{"fecha":"2031-10-06","tipo":"reposicion",`+
			`"preceptos":[{"norma":"BOE-A-2015-10565","bloque":"a124","apartado":"124.1"}],`+
			`"municipios":[{"codigo_ine":"71001","nombre":"Villa Primera"}],"festivos":[],"avisos":[],`+
			`"cobertura":[{"ambito":"nacional-y-autonomico","desde":2031,"hasta":2031,"disposicion":"PRUEBA-N-2031"},`+
			`{"ambito":"local","municipio":"71001","desde":2031,"hasta":2031,"disposicion":"PRUEBA-L-2"}]}`,
			string(serializado))
	})
}

// etiquetasJSON son las etiquetas json de los campos de un tipo, en su orden.
func etiquetasJSON(tipo reflect.Type) []string {
	etiquetas := make([]string, 0, tipo.NumField())
	for indice := range tipo.NumField() {
		etiquetas = append(etiquetas, tipo.Field(indice).Tag.Get("json"))
	}

	return etiquetas
}

// enumeradoDelCampo son los valores enum de la etiqueta jsonschema de un
// campo, en su orden.
func enumeradoDelCampo(t *testing.T, tipo reflect.Type, nombre string) []string {
	t.Helper()

	campo, existe := tipo.FieldByName(nombre)
	require.True(t, existe, nombre)

	var valores []string

	for parte := range strings.SplitSeq(campo.Tag.Get("jsonschema"), ",") {
		if valor, esEnum := strings.CutPrefix(parte, "enum="); esEnum {
			valores = append(valores, valor)
		}
	}

	return valores
}

// cadenas es un vocabulario como cadenas.
func cadenas[T ~string](valores []T) []string {
	textos := make([]string, 0, len(valores))
	for _, valor := range valores {
		textos = append(textos, string(valor))
	}

	return textos
}
