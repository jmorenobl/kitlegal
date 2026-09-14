package boe

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Las tres frases de los avisos, copiadas carácter a carácter de _check_vigencia
// (refs/boe.py 202-211) y partidas donde las parte refs/boe.py, no desde
// avisos.go: si el código cambiara una sola letra, estas tablas lo verían
// (FR-012, FR-050).
const (
	fraseEsperadaDeConsolidacionNoFinalizada = "⚠ TEXTO POSIBLEMENTE DESACTUALIZADO: la consolidación de esta norma " +
		"no está finalizada. Puede haber modificaciones recientes aún no integradas."
	fraseEsperadaDeNormaDerogada   = "⚠ NORMA DEROGADA: esta norma ha sido derogada."
	fraseEsperadaDeVigenciaAgotada = "⚠ VIGENCIA AGOTADA: esta norma ya no está en vigor."
)

// TestAvisosDe fija los avisos de vigencia que se derivan de los metadatos de una
// norma (FR-012, FR-050, data-model.md §2.2): las tres condiciones de
// _check_vigencia (refs/boe.py 199-211), comparadas como cadenas exactas y en su
// orden —consolidación no finalizada, derogada, vigencia agotada—, cada una con
// su código y su frase literal; ninguna, la lista vacía y nunca nula.
//
// Los metadatos se leen como los entrega la API a la lectura, con los números
// como json.Number: así «código numérico» es el valor que de verdad llegaría.
func TestAvisosDe(t *testing.T) {
	t.Parallel()

	consolidacion := Aviso{Codigo: "consolidacion-no-finalizada", Texto: fraseEsperadaDeConsolidacionNoFinalizada}
	derogada := Aviso{Codigo: "derogada", Texto: fraseEsperadaDeNormaDerogada}
	agotada := Aviso{Codigo: "vigencia-agotada", Texto: fraseEsperadaDeVigenciaAgotada}

	t.Run("las ocho combinaciones, en el orden de _check_vigencia", func(t *testing.T) {
		t.Parallel()

		casos := []struct {
			nombre    string
			metadatos string
			avisos    []Aviso
		}{
			{nombre: "ninguna condición", metadatos: metadatosDePrueba("3", "N", "N"), avisos: []Aviso{}},
			{
				nombre:    "consolidación no finalizada",
				metadatos: metadatosDePrueba("4", "N", "N"),
				avisos:    []Aviso{consolidacion},
			},
			{nombre: "derogada", metadatos: metadatosDePrueba("3", "S", "N"), avisos: []Aviso{derogada}},
			{nombre: "vigencia agotada", metadatos: metadatosDePrueba("3", "N", "S"), avisos: []Aviso{agotada}},
			{
				nombre:    "consolidación no finalizada y derogada",
				metadatos: metadatosDePrueba("4", "S", "N"),
				avisos:    []Aviso{consolidacion, derogada},
			},
			{
				nombre:    "consolidación no finalizada y vigencia agotada",
				metadatos: metadatosDePrueba("4", "N", "S"),
				avisos:    []Aviso{consolidacion, agotada},
			},
			{
				nombre:    "derogada y vigencia agotada, como el a42 de BOE-A-1992-26318",
				metadatos: metadatosDePrueba("3", "S", "S"),
				avisos:    []Aviso{derogada, agotada},
			},
			{
				nombre:    "las tres",
				metadatos: metadatosDePrueba("4", "S", "S"),
				avisos:    []Aviso{consolidacion, derogada, agotada},
			},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				avisos := avisosDe(objetoJSON(t, caso.metadatos))

				require.NotNil(t, avisos, "sin avisos, la lista vacía y no nula")
				assert.Equal(t, caso.avisos, avisos)
			})
		}
	})

	t.Run("lo que no es exactamente la cadena de la condición no avisa", func(t *testing.T) {
		t.Parallel()

		casos := []struct {
			nombre    string
			metadatos string
		}{
			{nombre: "código numérico", metadatos: `{"estado_consolidacion": {"codigo": 4, "texto": "En proceso"}}`},
			{nombre: "código numérico con decimales", metadatos: `{"estado_consolidacion": {"codigo": 4.0}}`},
			{nombre: "código con un cero delante", metadatos: `{"estado_consolidacion": {"codigo": "04"}}`},
			{nombre: "código con espacio", metadatos: `{"estado_consolidacion": {"codigo": " 4"}}`},
			{nombre: "estado como cadena, no objeto", metadatos: `{"estado_consolidacion": "4"}`},
			{nombre: "estado como lista, no objeto", metadatos: `{"estado_consolidacion": [{"codigo": "4"}]}`},
			{nombre: "estado nulo", metadatos: `{"estado_consolidacion": null}`},
			{nombre: "estado sin código", metadatos: `{"estado_consolidacion": {"texto": "4"}}`},
			{nombre: "código nulo", metadatos: `{"estado_consolidacion": {"codigo": null}}`},
			{nombre: "derogación en minúscula", metadatos: `{"estatus_derogacion": "s"}`},
			{nombre: "derogación con espacio detrás", metadatos: `{"estatus_derogacion": "S "}`},
			{nombre: "derogación como Sí", metadatos: `{"estatus_derogacion": "Sí"}`},
			{nombre: "derogación booleana", metadatos: `{"estatus_derogacion": true}`},
			{nombre: "derogación como objeto", metadatos: `{"estatus_derogacion": {"codigo": "S"}}`},
			{nombre: "vigencia en minúscula", metadatos: `{"vigencia_agotada": "s"}`},
			{nombre: "vigencia numérica", metadatos: `{"vigencia_agotada": 1}`},
			{nombre: "vigencia como lista", metadatos: `{"vigencia_agotada": ["S"]}`},
			{
				nombre:    "claves en otra caja",
				metadatos: `{"Estado_Consolidacion": {"codigo": "4"}, "ESTATUS_DEROGACION": "S", "Vigencia_Agotada": "S"}`,
			},
			{nombre: "sin ninguna de las tres claves", metadatos: `{"titulo": "Ley 39/2015"}`},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				avisos := avisosDe(objetoJSON(t, caso.metadatos))

				assert.Empty(t, avisos)
				assert.JSONEq(t, `[]`, comoJSON(t, avisos), "sin avisos, la lista vacía y no nula")
			})
		}
	})
}

// TestCodigosDeAviso fija el enumerado de Aviso.Codigo (FR-012, Q5): exactamente
// consolidacion-no-finalizada, derogada y vigencia-agotada, en el orden de sus
// condiciones; con las tres condiciones, los avisos llevan esos códigos en ese
// orden y una frase distinta cada uno, de modo que a cada código le corresponde
// una sola frase (FR-050).
func TestCodigosDeAviso(t *testing.T) {
	t.Parallel()

	t.Run("exactamente los tres códigos, en el orden de sus condiciones", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, []string{"consolidacion-no-finalizada", "derogada", "vigencia-agotada"}, CodigosDeAviso())
	})

	t.Run("con las tres condiciones, un aviso por código en su orden y una frase por código", func(t *testing.T) {
		t.Parallel()

		avisos := avisosDe(objetoJSON(t, metadatosDePrueba("4", "S", "S")))

		codigos := make([]string, 0, len(avisos))
		frases := map[string]string{}

		for _, aviso := range avisos {
			codigos = append(codigos, aviso.Codigo)
			frases[aviso.Texto] = aviso.Codigo
		}

		assert.Equal(t, CodigosDeAviso(), codigos)
		assert.Len(t, frases, len(CodigosDeAviso()), "dos códigos comparten frase: %v", frases)
	})

	t.Run("cada llamada devuelve su propia lista", func(t *testing.T) {
		t.Parallel()

		codigos := CodigosDeAviso()
		codigos[0] = "otro"

		assert.Equal(t, "consolidacion-no-finalizada", CodigosDeAviso()[0])
	})
}

// metadatosDePrueba es el objeto de metadatos de una norma con los tres campos
// que miran las condiciones y los que no deben influir, en la forma que da la
// API (estado_consolidacion como objeto con codigo y texto).
func metadatosDePrueba(codigoDeConsolidacion, estatusDeDerogacion, vigenciaAgotada string) string {
	return fmt.Sprintf(`{
		"identificador": "BOE-A-2015-10565",
		"titulo": "Ley 39/2015, de 1 de octubre, del Procedimiento Administrativo Común",
		"rango": {"codigo": "1300", "texto": "Ley"},
		"estatus_derogacion": %q,
		"vigencia_agotada": %q,
		"estado_consolidacion": {"codigo": %q, "texto": "Texto de prueba"},
		"url_eli": "https://www.boe.es/eli/es/l/2015/10/01/39"
	}`, estatusDeDerogacion, vigenciaAgotada, codigoDeConsolidacion)
}
