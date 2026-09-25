package territorio

import (
	"fmt"
	"runtime"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// filasDelTamanoReal son las filas de la relación congelada, una por municipio,
// y también las de la correspondencia, que los trae todos verificados
// (research.md S3).
const filasDelTamanoReal = 8132

// Lo que Cargar puede asignar, como mucho, con las fuentes del tamaño real: un
// tercio de lo que asignaba la carga que leía las filas con el lector de YAML,
// medido con estas mismas fuentes antes de cambiarla —779 065 asignaciones y
// 37 455 908 bytes, la menor de tres medidas— (research.md S3).
const (
	maximoDeAsignaciones = 779_065 / 3
	maximoDeBytes        = 37_455_908 / 3
)

// TestCosteDeLaCarga fija lo que cuesta cargar un territorio del tamaño real,
// que el binario paga en cada invocación: el tiempo de `territorio resolver`
// es casi todo carga, y lo que lo lleva al borde de su máximo es lo que asigna
// y la recogida de basura que eso arrastra (plan, Performance Goals; FR-056).
//
// Mide asignaciones y bytes, no tiempo, para que el resultado no dependa de la
// máquina ni de su carga: es la media de varias cargas, tras una primera que
// no cuenta, como testing.AllocsPerRun.
//
//nolint:paralleltest // Mide lo que asigna todo el proceso: con otra prueba a la vez, contaría también lo suyo.
func TestCosteDeLaCarga(t *testing.T) {
	fuentes := fuentesDelTamanoReal(t)

	var err error

	asignaciones, bytes := costeDe(func() { _, err = Cargar(fuentes) })
	require.NoError(t, err, "las fuentes del tamaño real tienen que cargar sin ningún defecto")

	t.Logf("cargar el tamaño real asigna %d veces y %d bytes", asignaciones, bytes)
	assert.LessOrEqual(t, asignaciones, uint64(maximoDeAsignaciones), "asignaciones de una carga")
	assert.LessOrEqual(t, bytes, uint64(maximoDeBytes), "bytes asignados por una carga")
}

// costeDe devuelve lo que asigna una llamada a f, en asignaciones y en bytes:
// la media de varias, tras una primera que no cuenta. Como
// testing.AllocsPerRun, corre con un solo procesador lógico.
func costeDe(f func()) (asignaciones, bytes uint64) {
	const veces = 4

	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(1))

	f()

	var antes, despues runtime.MemStats

	runtime.ReadMemStats(&antes)

	for range veces {
		f()
	}

	runtime.ReadMemStats(&despues)

	return (despues.Mallocs - antes.Mallocs) / veces, (despues.TotalAlloc - antes.TotalAlloc) / veces
}

// fuentesDelTamanoReal devuelve un territorio sintético completo y coherente
// con tantos municipios como la relación congelada, todos con su DIR3, y la
// relación y la correspondencia escritas como los ficheros congelados (FR-041,
// research.md D4). Es inventado como el de ficherosSinteticos: los nombres no
// son de ningún municipio y el dígito de control no es el oficial de ningún
// código, así que nada de esto puede tomarse por un dato real.
//
// Las provincias son las 52 y se reparten entre 19 comunidades; los nombres
// mezclan, como la relación, los bilingües, los de artículo pospuesto y los
// que llevan diacríticos o apóstrofo.
func fuentesDelTamanoReal(t *testing.T) Fuentes {
	t.Helper()

	const provincias, comunidades = 52, 19

	sinteticos := ficherosSinteticos()

	ficheros := Ficheros{
		Municipios: FicheroDeMunicipios{
			Fecha: fechaDeLaRelacion, Source: fuenteDeLaRelacion, Municipios: map[string]FilaDeMunicipio{},
		},
		DIR3: FicheroDeDIR3{
			Fecha: fechaDeLaCorrespondencia, Source: fuenteDeLaCorrespondencia, Correspondencia: map[string]string{},
		},
		Estado:      sinteticos.Estado,
		Comunidades: map[string]FicheroDeComunidad{},
	}

	for provincia := 1; provincia <= provincias; provincia++ {
		comunidad := fmt.Sprintf("%02d", (provincia-1)%comunidades+1)

		fichero, esta := ficheros.Comunidades[comunidad]
		if !esta {
			fichero = FicheroDeComunidad{
				Fecha: fechaDeLasComunidades, Source: fuenteDeLosNombres, Codigo: comunidad,
				Nombre: "Comunidad Sintética " + comunidad, Regimen: regimenComun, Provincias: map[string]string{},
			}
			ficheros.Comunidades[comunidad] = fichero
		}

		fichero.Provincias[fmt.Sprintf("%02d", provincia)] = fmt.Sprintf("Provincia Sintética %02d", provincia)
	}

	for indice := range filasDelTamanoReal {
		provincia := indice%provincias + 1
		codigo := fmt.Sprintf("%02d%03d", provincia, indice/provincias+1)
		digito := strconv.Itoa(indice % 10)

		ficheros.Municipios.Municipios[codigo] = FilaDeMunicipio{
			DC:        digito,
			Nombre:    nombreSintetico(indice),
			Provincia: codigo[:2],
			Comunidad: fmt.Sprintf("%02d", (provincia-1)%comunidades+1),
		}
		ficheros.DIR3.Correspondencia[codigo] = "L01" + codigo + digito
	}

	return fuentesEnFilas(t, ficheros)
}

// nombreSintetico es el nombre inventado del municipio de un índice.
func nombreSintetico(indice int) string {
	switch {
	case indice%11 == 0:
		return fmt.Sprintf("Villa de Prueba %d/Proba Herria %d", indice, indice)
	case indice%13 == 0:
		return fmt.Sprintf("Rozas de Prueba %d, Las", indice)
	case indice%17 == 0:
		return fmt.Sprintf("Sant Martí de l'Assaig %d", indice)
	default:
		return fmt.Sprintf("Peñón de Prueba %d", indice)
	}
}
