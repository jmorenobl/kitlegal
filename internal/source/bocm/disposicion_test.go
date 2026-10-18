package bocm

import (
	"encoding/json/v2"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/plazos"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/httpx"
)

// Dónde están, relativos al directorio del paquete, lo que grabó grabar_datos y
// el manifiesto con el que lo grabó (contracts/datos-de-plazos.md §1.3 y §4.1).
const (
	raizDeLoGrabado   = "../../../evidencias/h9/" + NombreDeLaFuente
	rutaDelManifiesto = "testdata/grabaciones.json"
)

// manifiestoDelPaquete es testdata/grabaciones.json, leído como lo lee la
// derivación: sin miembros que no sean campos del manifiesto.
func manifiestoDelPaquete(t *testing.T) Manifiesto {
	t.Helper()

	contenido, err := os.ReadFile(rutaDelManifiesto)
	require.NoError(t, err)

	var manifiesto Manifiesto
	require.NoError(t, json.Unmarshal(contenido, &manifiesto, json.RejectUnknownMembers(true)))

	return manifiesto
}

// clienteDeLoGrabado reproduce, sin red, lo que grabó grabar_datos.
func clienteDeLoGrabado(t *testing.T) *httpx.Cliente {
	t.Helper()

	cliente, err := httpx.Replay(raizDeLoGrabado, httpx.ConFuente(NombreDeLaFuente))
	require.NoError(t, err)

	return cliente
}

// resolucionGrabada es la disposición grabada con ese identificador, leída con
// LeerDisposicion.
func resolucionGrabada(t *testing.T, id string) Resolucion {
	t.Helper()

	manifiesto := manifiestoDelPaquete(t)
	indice := slices.IndexFunc(manifiesto.Disposiciones, func(d Disposicion) bool { return d.ID == id })
	require.GreaterOrEqual(t, indice, 0, "el manifiesto no lista la disposición %s", id)

	respuesta, err := clienteDeLoGrabado(t).Pedir(t.Context(), schema.Contexto{},
		httpx.Peticion{Metodo: metodoDeLaLectura, URL: manifiesto.Disposiciones[indice].URL})
	require.NoError(t, err)

	resolucion, err := LeerDisposicion(respuesta.Cuerpo)
	require.NoError(t, err)

	return resolucion
}

// lineaDe es la línea de la disposición cuyo nombre es ese, tal como lo
// escribe el boletín.
func lineaDe(t *testing.T, resolucion Resolucion, nombre string) Linea {
	t.Helper()

	indice := slices.IndexFunc(resolucion.Lineas, func(l Linea) bool { return l.Nombre == nombre })
	require.GreaterOrEqual(t, indice, 0, "%s no tiene la línea de «%s»", resolucion.ID, nombre)

	return resolucion.Lineas[indice]
}

// fechas es una forma corta de escribir las fechas que se esperan de una línea.
func fechas(diasYMeses ...int) []Fecha {
	var esperadas []Fecha
	for indice := 0; indice+1 < len(diasYMeses); indice += 2 {
		esperadas = append(esperadas, Fecha{Dia: diasYMeses[indice], Mes: time.Month(diasYMeses[indice+1])})
	}

	return esperadas
}

// documentoDePrueba es el XML de una disposición con ese identificador, ese
// título y esas líneas en el texto, en la forma en que las graba el BOCM.
func documentoDePrueba(id, titulo string, lineas ...string) []byte {
	texto := "I. COMUNIDAD DE MADRID\nACUERDA\nDeclarar inhábiles los días siguientes:\n"
	for _, linea := range lineas {
		texto += principioDeLinea + linea + "\n"
	}

	texto += "Madrid, a 1 de enero de 2025.\n"

	return []byte("<documento><metadatos><identificador>" + id + "</identificador><titulo>" + titulo +
		"</titulo></metadatos><texto>" + texto + "</texto></documento>")
}

// TestLeerDisposicion fija lo que LeerDisposicion lee de las siete
// disposiciones grabadas y de la gramática de la línea (contracts/datos-de-plazos.md
// §1.3; research D13, V6, V8; FR-023): el identificador y el año de cada una,
// cuántas líneas tiene y cuántas quedan como «no comunicado» o sin leer; las
// erratas de forma de V8, que se leen; la línea de Valdeolmos-Alalpardo de 2025
// y de 2026, con paréntesis, que no; y lo que no es una disposición.
func TestLeerDisposicion(t *testing.T) {
	t.Parallel()

	t.Run("lo-grabado", func(t *testing.T) {
		t.Parallel()

		casos := []struct {
			id                                   string
			anio, lineas, noComunicadas, sinLeer int
		}{
			{"BOCM-20241213-23", 2025, 182, 6, 1},
			{"BOCM-20241227-32", 2025, 4, 0, 0},
			{"BOCM-20250203-16", 2025, 3, 0, 0},
			{"BOCM-20250217-24", 2025, 1, 0, 0},
			{"BOCM-20251212-34", 2026, 182, 14, 1},
			{"BOCM-20251229-19", 2026, 9, 0, 0},
			{"BOCM-20260108-18", 2026, 6, 0, 0},
		}

		for _, caso := range casos {
			resolucion := resolucionGrabada(t, caso.id)
			assert.Equal(t, caso.id, resolucion.ID)
			assert.Equal(t, caso.anio, resolucion.Anio, caso.id)
			assert.Len(t, resolucion.Lineas, caso.lineas, caso.id)

			noComunicadas, sinLeer := 0, 0

			for _, linea := range resolucion.Lineas {
				assert.NotEmpty(t, linea.Nombre, "%s: una línea sin nombre", caso.id)

				switch linea.Falta {
				case plazos.FaltaNoComunicado:
					noComunicadas++
				case plazos.FaltaLineaSinLeer:
					sinLeer++
				default:
					assert.Empty(t, linea.Falta, "%s, «%s»: la falta no es de una línea", caso.id, linea.Nombre)
				}

				assert.Equal(t, linea.Falta == "", len(linea.Fechas) > 0,
					"%s, «%s»: una línea da festivos o una falta, nunca las dos cosas ni ninguna", caso.id, linea.Nombre)
			}

			assert.Equal(t, caso.noComunicadas, noComunicadas, "%s: líneas «no comunicado»", caso.id)
			assert.Equal(t, caso.sinLeer, sinLeer, "%s: líneas sin leer", caso.id)
		}
	})

	t.Run("erratas-de-forma", func(t *testing.T) {
		t.Parallel()

		de2025, de2026 := resolucionGrabada(t, "BOCM-20241213-23"), resolucionGrabada(t, "BOCM-20251212-34")

		casos := []struct {
			nombre     string
			resolucion Resolucion
			linea      string
			fechas     []Fecha
			falta      plazos.Falta
		}{
			{"mes-sin-de", de2025, "Carabaña", fechas(15, 9, 12, 12), ""},
			{"mes-sin-de-tras-dos-dias", de2025, "Cubas de la Sagra", fechas(3, 2, 15, 9), ""},
			{"punto-suelto-delante", de2025, "Navarredonda y San Mames", fechas(18, 8, 29, 9), ""},
			{"blanco-antes-del-punto", de2025, "Estremera", fechas(13, 10, 14, 10), ""},
			{"dos-dias-de-un-mes", de2026, "Algete", fechas(11, 9, 14, 9), ""},
			{"sin-dos-puntos", de2026, "Nuevo Baztán", fechas(11, 5, 3, 12), ""},
			{"sin-dos-puntos-con-de-en-el-nombre", de2026, "Olmeda de las Fuentes", fechas(6, 4, 24, 8), ""},
			{"no-comunicado-en-minuscula", de2025, "Belmonte de Tajo", nil, plazos.FaltaNoComunicado},
			{"no-comunicado-con-mayuscula", de2026, "Brunete", nil, plazos.FaltaNoComunicado},
		}

		for _, caso := range casos {
			linea := lineaDe(t, caso.resolucion, caso.linea)
			assert.Equal(t, caso.fechas, linea.Fechas, caso.nombre)
			assert.Equal(t, caso.falta, linea.Falta, caso.nombre)
		}
	})

	t.Run("valdeolmos-alalpardo", func(t *testing.T) {
		t.Parallel()

		for _, id := range []string{"BOCM-20241213-23", "BOCM-20251212-34"} {
			linea := lineaDe(t, resolucionGrabada(t, id), "Valdeolmos-Alalpardo")
			assert.Equal(t, plazos.FaltaLineaSinLeer, linea.Falta, id)
			assert.Empty(t, linea.Fechas, id)
			assert.Contains(t, linea.Cuerpo, "(Valdeolmos)", id)
			assert.Contains(t, linea.Cuerpo, "(Alalpardo)", id)
		}
	})

	t.Run("gramatica", func(t *testing.T) {
		t.Parallel()

		casos := []struct {
			nombre, linea, municipio string
			fechas                   []Fecha
			falta                    plazos.Falta
		}{
			{"un-dia-por-mes", "Villa: 1 de mayo y 2 de junio.", "Villa", fechas(1, 5, 2, 6), ""},
			{"dos-dias-de-un-mes", "Villa: 1 y 2 de mayo.", "Villa", fechas(1, 5, 2, 5), ""},
			{"un-solo-grupo", "Villa: 1 de mayo", "Villa", fechas(1, 5), ""},
			{"tres-grupos", "Villa: 1 de enero y 2 de febrero y 3 de marzo.", "Villa", fechas(1, 1, 2, 2, 3, 3), ""},
			{"sin-dos-puntos", "Villa Alta 3 de marzo.", "Villa Alta", fechas(3, 3), ""},
			{"blancos-y-puntos-sueltos", "Villa :  . 3  de marzo  y 4 de  abril . .", "Villa", fechas(3, 3, 4, 4), ""},
			{"mes-en-mayusculas", "Villa: 3 de Marzo Y 4 de ABRIL.", "Villa", fechas(3, 3, 4, 4), ""},
			{"veintinueve-de-febrero", "Villa: 29 de febrero.", "Villa", fechas(29, 2), ""},
			{"no-comunicado", "Villa: No Comunicado Por El Ayuntamiento.", "Villa", nil, plazos.FaltaNoComunicado},
			{"parentesis", "Villa: 1 de mayo (Arriba) y 2 de junio (Abajo).", "Villa", nil, plazos.FaltaLineaSinLeer},
			{"mes-que-no-es-un-mes", "Villa: 1 de maio.", "Villa", nil, plazos.FaltaLineaSinLeer},
			{"dia-que-no-es-un-dia", "Villa: 32 de mayo.", "Villa", nil, plazos.FaltaLineaSinLeer},
			{"treinta-de-febrero", "Villa: 30 de febrero.", "Villa", nil, plazos.FaltaLineaSinLeer},
			{"dia-cero", "Villa: 0 de mayo.", "Villa", nil, plazos.FaltaLineaSinLeer},
			{"tres-dias-con-coma", "Villa: 1, 2 y 3 de mayo.", "Villa", nil, plazos.FaltaLineaSinLeer},
			{"nexo-al-final", "Villa: 1 de mayo y.", "Villa", nil, plazos.FaltaLineaSinLeer},
			{"sin-mes", "Villa: 1 y 2.", "Villa", nil, plazos.FaltaLineaSinLeer},
			{"sin-cuerpo", "Villa:", "Villa", nil, plazos.FaltaLineaSinLeer},
			{
				"sin-dos-puntos-ni-cifra", "Villa no comunicado por el ayuntamiento", "Villa no comunicado por el ayuntamiento",
				nil, plazos.FaltaLineaSinLeer,
			},
			{"otra-cosa-tras-el-mes", "Villa: 1 de mayo festivo.", "Villa", nil, plazos.FaltaLineaSinLeer},
		}

		for _, caso := range casos {
			resolucion, err := LeerDisposicion(documentoDePrueba("BOCM-20250101-1", "Fiestas para el año 2025", caso.linea))
			require.NoError(t, err, caso.nombre)
			require.Len(t, resolucion.Lineas, 1, caso.nombre)

			linea := resolucion.Lineas[0]
			assert.Equal(t, caso.municipio, linea.Nombre, caso.nombre)
			assert.Equal(t, caso.fechas, linea.Fechas, caso.nombre)
			assert.Equal(t, caso.falta, linea.Falta, caso.nombre)
		}
	})

	t.Run("lo-que-no-es-una-disposicion", func(t *testing.T) {
		t.Parallel()

		casos := []struct {
			nombre    string
			documento []byte
			motivo    string
		}{
			{"vacio", nil, "no es una disposición"},
			{"xml-roto", []byte("<documento><texto>"), "no se lee como XML"},
			{"otra-cosa", []byte("<html><body>Fiestas</body></html>"), "no tiene identificador"},
			{"sin-identificador", documentoDePrueba("", "Fiestas para el año 2025", "Villa: 1 de mayo."), "no tiene identificador"},
			{"sin-anio", documentoDePrueba("BOCM-20250101-1", "Fiestas laborales", "Villa: 1 de mayo."), "de qué año"},
			{"anio-sin-cifras", documentoDePrueba("BOCM-20250101-1", "Fiestas para el año dos mil", "Villa: 1."), "de qué año"},
			{"sin-lineas", documentoDePrueba("BOCM-20250101-1", "Fiestas para el año 2025"), "ninguna línea"},
		}

		for _, caso := range casos {
			_, err := LeerDisposicion(caso.documento)
			assert.ErrorContains(t, err, caso.motivo, caso.nombre)
		}
	})
}
