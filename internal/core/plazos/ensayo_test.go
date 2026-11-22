package plazos

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// computoDeEnsayo es el fichero del cómputo de ensayo (spec §C.10), un cómputo
// inventado que no es de ninguna materia: entra con la forma de §C.1 sin tocar
// Go, y el motor lo cuenta con lo que dicen sus datos (FR-022).
//
//go:embed testdata/computo-de-ensayo.yaml
var computoDeEnsayo []byte

// textoDelAvisoDeEnsayo es el «aviso del periodo» de spec §C.10.
const textoDelAvisoDeEnsayo = "⚠ PERIODO DE ENSAYO: el plazo pasa por el periodo de ensayo, del 19 al 23 de octubre; " +
	"sus días se han contado."

// casoDeEnsayo es una fila de la tabla de spec §C.10: la petición y, si da
// plazo, su fecha y los códigos de sus avisos en su orden o, si no, el mensaje
// de su error, de clase argumentos, que el kernel da con el código 2.
type casoDeEnsayo struct {
	nombre   string
	peticion Peticion
	fecha    string
	avisos   []string
	error    string
}

// peticionDeEnsayo es una petición de un tipo del ensayo; sin municipios, la
// del lugar del ensayo, 81001 «Lugar Uno».
func peticionDeEnsayo(tipo, inicio, duracion string, unidad Unidad, municipios ...MunicipioResuelto) Peticion {
	if len(municipios) == 0 {
		municipios = los(municipiosDeLaRejilla[0])
	}

	return Peticion{Tipo: tipo, Inicio: inicio, Duracion: duracion, Unidad: string(unidad), Municipios: municipios}
}

// casosDeEnsayo son las filas E1 a E13 de spec §C.10, cada una sobre un eje del
// cómputo que dan los datos: un periodo inhabil en días hábiles y en días
// naturales, con su primer día y su último; dia-siguiente, sin prórroga; un
// periodo aviso, por su primer día, por el último, antes y después; el orden de
// los avisos de FR-008; las unidades del cómputo (FR-007) y sus municipios
// (FR-006).
func casosDeEnsayo() []casoDeEnsayo {
	lugarDos, lugarTres := municipiosDeLaRejilla[1], municipiosDeLaRejilla[2]

	return []casoDeEnsayo{
		{
			nombre:   "E1 inhabil en días hábiles",
			peticion: peticionDeEnsayo("ensayo-libre", "2026-04-02", "3", UnidadDiasHabiles),
			fecha:    "2026-04-14",
		},
		{
			nombre:   "E2 primer día del periodo inhabil",
			peticion: peticionDeEnsayo("ensayo-libre", "2026-04-03", "1", UnidadDiasHabiles),
			fecha:    "2026-04-13",
		},
		{
			nombre:   "E3 último día del periodo inhabil",
			peticion: peticionDeEnsayo("ensayo-libre", "2026-04-09", "1", UnidadDiasHabiles),
			fecha:    "2026-04-13",
		},
		{
			nombre:   "E4 inhabil en días naturales",
			peticion: peticionDeEnsayo("ensayo-libre", "2026-03-27", "10", UnidadDiasNaturales),
			fecha:    "2026-04-13",
		},
		{
			nombre:   "E5 dia-siguiente",
			peticion: peticionDeEnsayo("ensayo-fijo", "2026-06-01", "", ""),
			fecha:    "2026-06-09",
		},
		{
			nombre:   "E6 dia-siguiente sin prórroga",
			peticion: peticionDeEnsayo("ensayo-fijo", "2026-06-05", "", ""),
			fecha:    "2026-06-13",
		},
		{
			nombre:   "E7 aviso por su primer día",
			peticion: peticionDeEnsayo("ensayo-libre", "2026-10-16", "1", UnidadDiasHabiles),
			fecha:    "2026-10-19",
			avisos:   []string{"periodo-de-ensayo"},
		},
		{
			nombre:   "E8 aviso por su último día",
			peticion: peticionDeEnsayo("ensayo-libre", "2026-10-22", "1", UnidadDiasHabiles),
			fecha:    "2026-10-23",
			avisos:   []string{"periodo-de-ensayo"},
		},
		{
			nombre:   "E9 tramo después del periodo aviso",
			peticion: peticionDeEnsayo("ensayo-libre", "2026-10-23", "1", UnidadDiasHabiles),
			fecha:    "2026-10-26",
		},
		{
			nombre:   "E10 tramo antes del periodo aviso",
			peticion: peticionDeEnsayo("ensayo-libre", "2026-10-15", "1", UnidadDiasHabiles),
			fecha:    "2026-10-16",
		},
		{
			nombre:   "E11 orden de los avisos",
			peticion: peticionDeEnsayo("ensayo-libre", "2026-06-23", "120", UnidadDiasNaturales, lugarTres),
			fecha:    "2026-10-21",
			avisos:   []string{AvisoFaltanLocales, "periodo-de-ensayo", AvisoFestivoDeUnaParte},
		},
		{
			nombre:   "E12 unidades del cómputo",
			peticion: peticionDeEnsayo("ensayo-libre", "2026-06-01", "1", UnidadMeses),
			error: `el tipo "ensayo-libre" no admite la unidad "meses": su cómputo, "ensayo", admite dias-habiles o ` +
				`dias-naturales`,
		},
		{
			nombre: "E13 municipios del cómputo",
			peticion: peticionDeEnsayo("ensayo-libre", "2026-06-01", "1", UnidadDiasHabiles,
				municipiosDeLaRejilla[0], lugarDos),
			error: `el tipo "ensayo-libre" se cuenta con un municipio, el del lugar del ensayo, y se han dado 2`,
		},
	}
}

// TestComputoDeEnsayo lee el cómputo de ensayo con LeerTipos y calcula sobre la
// rejilla de las propiedades cada fila de spec §C.10 (FR-004 a FR-008, FR-022;
// SC-003): ningún código suyo está en Go, así que cada fecha, cada aviso y cada
// error sale de lo que dicen sus datos. Un plazo da su fecha, ningún festivo
// saltado y los avisos en su orden, el del periodo con su texto de los datos;
// un error, el mismo con Calcular y con ComprobarPeticion, de clase argumentos
// y con su mensaje.
func TestComputoDeEnsayo(t *testing.T) {
	t.Parallel()

	tipos, err := LeerTipos(computoDeEnsayo)
	require.NoError(t, err, "el cómputo de ensayo entra sin defecto")
	require.Len(t, tipos, 2)

	calendario := nuevaRejilla(t).calendario

	for _, caso := range casosDeEnsayo() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			plazo, err := Calcular(calendario, tipos, caso.peticion)

			if caso.error != "" {
				require.Error(t, err)
				assert.Equal(t, Plazo{}, plazo)
				comprobarClaseArgumentos(t, err)
				assert.Equal(t, caso.error, err.Error())

				comprobado := ComprobarPeticion(tipos, caso.peticion)
				require.Error(t, comprobado)
				assert.Equal(t, caso.error, comprobado.Error())

				return
			}

			require.NoError(t, err)
			assert.Equal(t, caso.fecha, plazo.Fecha)
			assert.Equal(t, caso.peticion.Tipo, plazo.Tipo)
			assert.Empty(t, plazo.Festivos, "ningún caso de ensayo salta festivos")
			assert.Equal(t, append([]string{}, caso.avisos...), codigosDe(plazo.Avisos))

			for _, aviso := range plazo.Avisos {
				if aviso.Codigo == "periodo-de-ensayo" {
					assert.Equal(t, textoDelAvisoDeEnsayo, aviso.Texto)
				}
			}
		})
	}
}
