package cendoj

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Las referencias de la sentencia que docs/JURISPRUDENCIA.md §3 resuelve por sus
// tres vías, y la del Tribunal Constitucional de su §1.
const (
	ecliConocido       = "ECLI:ES:TS:2023:3144"
	rojConocido        = "STS 3144/2023"
	resolucionConocida = "1088/2023"
	fechaConocida      = "2023-07-04"
	ecliConstitucional = "ECLI:ES:TC:2024:79"
)

// Lo que dice cada rechazo de NuevaReferencia que no viene de internal/core/ids:
// el fragmento de su mensaje que nombra lo que falla.
const (
	fragmentoFalta      = "falta la referencia"
	fragmentoUnaSola    = "tiene que ser una sola"
	fragmentoSinFecha   = "un ECLI no lleva fecha"
	fragmentoResolucion = "la forma es <número>/<año>"
	fragmentoNecesita   = "necesita su fecha"
	fragmentoForma      = "la forma es AAAA-MM-DD"
	fragmentoNoExiste   = "no es un día que existe"
)

// TestNuevaReferencia fija la validación de la referencia de cita resolver, en
// su orden (contrato cita-resolver §1; FR-002 a FR-005): exactamente una forma;
// un ECLI con su forma y sin fecha; un ROJ con su forma y, si lleva fecha, la
// suya; y un número de resolución con su forma y con su fecha. Lo primero que
// falla es un error de la clase «argumentos» que dice qué falla, y nada se
// recorta ni se pasa a mayúsculas. Lleva los siete rechazos de quickstart.md §2
// y los casos límite del spec.
func TestNuevaReferencia(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre     string
		ecli       string
		roj        string
		resolucion string
		fecha      string
		// esperada es la referencia de una entrada que vale.
		esperada Referencia
		// mensaje es un fragmento del error de una entrada que no vale; vacío,
		// la entrada vale.
		mensaje string
	}{
		// Las tres formas, y el ROJ con su fecha.
		{
			nombre:   "un ECLI",
			ecli:     ecliConocido,
			esperada: Referencia{forma: formaECLI, valor: ecliConocido},
		},
		{
			nombre:   "un ECLI del Tribunal Constitucional",
			ecli:     ecliConstitucional,
			esperada: Referencia{forma: formaECLI, valor: ecliConstitucional, constitucional: true},
		},
		{
			nombre:   "un ROJ",
			roj:      rojConocido,
			esperada: Referencia{forma: formaROJ, valor: rojConocido},
		},
		{
			nombre:   "un ROJ con su fecha",
			roj:      rojConocido,
			fecha:    fechaConocida,
			esperada: Referencia{forma: formaROJ, valor: rojConocido, fecha: fechaConocida},
		},
		{
			nombre:     "un número de resolución con su fecha",
			resolucion: resolucionConocida,
			fecha:      fechaConocida,
			esperada:   Referencia{forma: formaResolucion, valor: resolucionConocida, fecha: fechaConocida},
		},
		{
			nombre:     "un número de resolución con ceros por delante",
			resolucion: "0079/2024",
			fecha:      "2024-02-29",
			esperada:   Referencia{forma: formaResolucion, valor: "0079/2024", fecha: "2024-02-29"},
		},
		{
			// Que es del Tribunal Constitucional solo lo dice su ECLI (spec, casos
			// límite): por su ROJ es una referencia como cualquier otra.
			nombre:   "una sentencia del Tribunal Constitucional por su ROJ",
			roj:      "STC 79/2024",
			esperada: Referencia{forma: formaROJ, valor: "STC 79/2024"},
		},

		// Los siete rechazos de quickstart.md §2.
		{nombre: "quickstart: un ECLI de cuatro partes", ecli: "ECLI:ES:TS:2023", mensaje: `el ECLI "ECLI:ES:TS:2023"`},
		{nombre: "quickstart: un ECLI de otro país", ecli: "ECLI:FR:CC:2023:1", mensaje: "no es español"},
		{nombre: "quickstart: un ECLI en minúsculas", ecli: "ecli:es:ts:2023:3144", mensaje: `el ECLI "ecli:es:ts:2023:3144"`},
		{nombre: "quickstart: un número de resolución sin fecha", resolucion: resolucionConocida, mensaje: fragmentoNecesita},
		{nombre: "quickstart: un ECLI con fecha", ecli: ecliConocido, fecha: fechaConocida, mensaje: fragmentoSinFecha},
		{nombre: "quickstart: un ECLI y un ROJ", ecli: ecliConocido, roj: rojConocido, mensaje: fragmentoUnaSola},
		{nombre: "quickstart: ninguna forma", mensaje: fragmentoFalta},

		// Exactamente una forma, antes que nada.
		{nombre: "solo la fecha", fecha: fechaConocida, mensaje: fragmentoFalta},
		{nombre: "un ROJ y un número de resolución", roj: rojConocido, resolucion: resolucionConocida, mensaje: "un ROJ y un número de resolución"},
		{nombre: "un ECLI y un número de resolución", ecli: ecliConocido, resolucion: resolucionConocida, fecha: fechaConocida, mensaje: "un ECLI y un número de resolución"},
		{
			nombre:     "las tres formas",
			ecli:       ecliConocido,
			roj:        rojConocido,
			resolucion: resolucionConocida,
			mensaje:    "un ECLI, un ROJ y un número de resolución",
		},
		{nombre: "dos formas mal escritas", ecli: "ecli", roj: "roj", mensaje: fragmentoUnaSola},

		// El ECLI: su forma y, después, sin fecha.
		{nombre: "un ECLI con un espacio delante", ecli: " " + ecliConocido, mensaje: `el ECLI " ECLI:ES:TS:2023:3144"`},
		{nombre: "un ECLI con un prefijo", ecli: "ECLI: " + ecliConocido, mensaje: "el ECLI"},
		{nombre: "un ECLI mal formado con fecha", ecli: "ECLI:ES:TS:2023", fecha: fechaConocida, mensaje: `el ECLI "ECLI:ES:TS:2023"`},
		{nombre: "un ECLI del Tribunal Constitucional con fecha", ecli: ecliConstitucional, fecha: "2024-06-20", mensaje: fragmentoSinFecha},
		{nombre: "un ECLI con una fecha mal escrita", ecli: ecliConocido, fecha: "ayer", mensaje: fragmentoSinFecha},

		// El ROJ: su forma y, después, la de su fecha.
		{nombre: "un ROJ con el prefijo ROJ:", roj: "ROJ: " + rojConocido, mensaje: `el ROJ "ROJ: STS 3144/2023"`},
		{nombre: "un ROJ en minúsculas", roj: "sts 3144/2023", mensaje: `el ROJ "sts 3144/2023"`},
		{nombre: "un ROJ mal formado con una fecha mal escrita", roj: "STS 3144", fecha: "ayer", mensaje: `el ROJ "STS 3144"`},
		{nombre: "un ROJ con una fecha de otra forma", roj: rojConocido, fecha: "04/07/2023", mensaje: fragmentoForma},
		{nombre: "un ROJ con una fecha sin ceros", roj: rojConocido, fecha: "2023-7-4", mensaje: fragmentoForma},
		{nombre: "un ROJ con un día que no existe", roj: rojConocido, fecha: "2023-02-30", mensaje: fragmentoNoExiste},
		{nombre: "un ROJ con un mes que no existe", roj: rojConocido, fecha: "2023-13-01", mensaje: fragmentoNoExiste},

		// El número de resolución: su forma, después su fecha y, después, la
		// forma de la fecha.
		{nombre: "un número de resolución con guion", resolucion: "1088-2023", fecha: fechaConocida, mensaje: fragmentoResolucion},
		{nombre: "un número de resolución con el año de dos cifras", resolucion: "1088/23", fecha: fechaConocida, mensaje: fragmentoResolucion},
		{nombre: "un número de resolución con el año de cinco cifras", resolucion: "1088/20234", fecha: fechaConocida, mensaje: fragmentoResolucion},
		{nombre: "un número de resolución sin número", resolucion: "/2023", fecha: fechaConocida, mensaje: fragmentoResolucion},
		{nombre: "un número de resolución con letras", resolucion: "STS 1088/2023", fecha: fechaConocida, mensaje: fragmentoResolucion},
		{nombre: "un número de resolución con un espacio detrás", resolucion: "1088/2023 ", fecha: fechaConocida, mensaje: fragmentoResolucion},
		{nombre: "un número de resolución con un salto de línea detrás", resolucion: "1088/2023\n", fecha: fechaConocida, mensaje: fragmentoResolucion},
		{nombre: "un número de resolución con cifras de otra escritura", resolucion: "１０８８/2023", fecha: fechaConocida, mensaje: fragmentoResolucion},
		{nombre: "un número de resolución mal formado y sin fecha", resolucion: "1088", mensaje: fragmentoResolucion},
		{nombre: "un número de resolución con una fecha de otra forma", resolucion: resolucionConocida, fecha: "4 de julio de 2023", mensaje: fragmentoForma},
		{nombre: "un número de resolución con un espacio tras la fecha", resolucion: resolucionConocida, fecha: fechaConocida + " ", mensaje: fragmentoForma},
		{nombre: "un número de resolución con un 29 de febrero que no existe", resolucion: resolucionConocida, fecha: "2023-02-29", mensaje: fragmentoNoExiste},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			referencia, err := NuevaReferencia(caso.ecli, caso.roj, caso.resolucion, caso.fecha)

			if caso.mensaje == "" {
				require.NoError(t, err)
				assert.Equal(t, caso.esperada, referencia)

				return
			}

			compruebaLaClase(t, err, schema.ClaseArgumentos)
			require.ErrorContains(t, err, caso.mensaje)
			assert.Zero(t, referencia, "una referencia que no vale no se entrega")
		})
	}
}

// TestDelTribunalConstitucional fija qué referencia es del Tribunal
// Constitucional: el ECLI cuyo órgano es TC, y ninguna otra (FR-015). Una
// sentencia suya pedida por su ROJ o por su número no se reconoce como tal
// (spec, casos límite), y la referencia a cero no es de ningún órgano.
func TestDelTribunalConstitucional(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre     string
		ecli       string
		roj        string
		resolucion string
		fecha      string
		esperado   bool
	}{
		{nombre: "un ECLI del Tribunal Constitucional", ecli: ecliConstitucional, esperado: true},
		{nombre: "un auto del Tribunal Constitucional", ecli: "ECLI:ES:TC:2024:79A", esperado: true},
		{nombre: "un ECLI del Tribunal Supremo", ecli: ecliConocido},
		{nombre: "un órgano que empieza por TC", ecli: "ECLI:ES:TCT:1988:1"},
		{nombre: "un órgano que termina en TC", ecli: "ECLI:ES:ATC:2024:79"},
		{nombre: "por su ROJ", roj: "STC 79/2024"},
		{nombre: "por su número y su fecha", resolucion: "79/2024", fecha: "2024-06-20"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			referencia, err := NuevaReferencia(caso.ecli, caso.roj, caso.resolucion, caso.fecha)
			require.NoError(t, err)

			assert.Equal(t, caso.esperado, referencia.DelTribunalConstitucional())
		})
	}

	t.Run("la referencia a cero", func(t *testing.T) {
		t.Parallel()

		assert.False(t, Referencia{}.DelTribunalConstitucional())
	})
}
