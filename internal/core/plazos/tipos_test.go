package plazos

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// plazosDelContrato es data/plazos.yaml tal como lo fija contracts/datos-de-plazos.md
// §2. Es una copia a propósito: el test del dominio no abre ficheros (R1), y el
// del repositorio lo comprueba contra lo grabado del BOE otro test
// (TestPreceptosDeLosPlazos).
const plazosDelContrato = `tipos:
  - codigo: alzada
    nombre: recurso de alzada
    computo: lpac
    duracion: {cantidad: 1, unidad: meses}
    preceptos:
      - {norma: BOE-A-2015-10565, bloque: a122, apartado: "122.1", fija: duracion}
  - codigo: reposicion
    nombre: recurso potestativo de reposición
    computo: lpac
    duracion: {cantidad: 1, unidad: meses}
    preceptos:
      - {norma: BOE-A-2015-10565, bloque: a124, apartado: "124.1", fija: duracion}
  - codigo: contencioso
    nombre: recurso contencioso-administrativo
    computo: contencioso
    duracion: {cantidad: 2, unidad: meses}
    preceptos:
      - {norma: BOE-A-1998-16718, bloque: a46, apartado: "46.1", fija: duracion}
      - {norma: BOE-A-1998-16718, bloque: a128, apartado: "128.2", fija: agosto}
  - codigo: acceso-resolver
    nombre: plazo para resolver una solicitud de acceso a la información pública
    computo: lpac
    duracion: {cantidad: 1, unidad: meses}
    preceptos:
      - {norma: BOE-A-2013-12887, bloque: a20, apartado: "20.1", fija: duracion}
  - codigo: acceso-reclamar
    nombre: reclamación contra la resolución de una solicitud de acceso
    computo: lpac
    duracion: {cantidad: 1, unidad: meses}
    preceptos:
      - {norma: BOE-A-2013-12887, bloque: a24, apartado: "24.2", fija: duracion}
  - codigo: silencio
    nombre: día en que se produce el silencio administrativo
    computo: lpac
    preceptos:
      - {norma: BOE-A-2015-10565, bloque: a21, apartado: "21.3", fija: inicio}
      - {norma: BOE-A-2015-10565, bloque: a24, apartado: "24.4", fija: silencio}
  - codigo: generico
    nombre: plazo genérico
    computo: lpac
    unidades:
      dias-habiles: {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.2", fija: computo}
      dias-naturales: {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.2", fija: computo}
      meses: {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.4", fija: computo}
`

// TestLeerTipos lee los siete tipos del contrato y rechaza el fichero en cada
// una de las reglas de LeerTipos, con un error que nombra el fichero y lo que
// falla y que no lleva clase de usuario: el kernel lo da como inesperado
// (FR-008, FR-030).
func TestLeerTipos(t *testing.T) {
	t.Parallel()

	t.Run("los siete del contrato", func(t *testing.T) {
		t.Parallel()

		tipos, err := LeerTipos([]byte(plazosDelContrato))
		require.NoError(t, err)

		codigos := make([]string, 0, len(tipos))
		for _, tipo := range tipos {
			codigos = append(codigos, tipo.Codigo)
		}

		assert.Equal(t, CodigosDeTipo(), codigos, "los siete, en el orden del fichero, que es el del contrato")

		assert.Equal(t, Tipo{
			Codigo:   "contencioso",
			Nombre:   "recurso contencioso-administrativo",
			Computo:  ComputoContencioso,
			Duracion: &Duracion{Cantidad: 2, Unidad: UnidadMeses},
			Preceptos: []Precepto{
				{Norma: "BOE-A-1998-16718", Bloque: "a46", Apartado: "46.1", Fija: FijaDuracion},
				{Norma: "BOE-A-1998-16718", Bloque: "a128", Apartado: "128.2", Fija: "agosto"},
			},
		}, tipos[2])

		assert.Nil(t, tipos[5].Duracion, "el silencio no fija su duración")
		assert.Nil(t, tipos[5].Unidades)
		assert.Equal(t, map[Unidad]Precepto{
			UnidadDiasHabiles:   {Norma: "BOE-A-2015-10565", Bloque: "a30", Apartado: "30.2", Fija: "computo"},
			UnidadDiasNaturales: {Norma: "BOE-A-2015-10565", Bloque: "a30", Apartado: "30.2", Fija: "computo"},
			UnidadMeses:         {Norma: "BOE-A-2015-10565", Bloque: "a30", Apartado: "30.4", Fija: "computo"},
		}, tipos[6].Unidades)
		assert.Empty(t, tipos[6].Preceptos)
	})

	t.Run("los vocabularios", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, []string{
			"alzada", "reposicion", "contencioso", "acceso-resolver", "acceso-reclamar", "silencio", "generico",
		}, CodigosDeTipo())
		assert.Equal(t, []Computo{"lpac", "contencioso"}, Computos())
		assert.Equal(t, []Unidad{"meses", "dias-habiles", "dias-naturales"}, Unidades())

		codigos := CodigosDeTipo()
		codigos[0] = "otro"
		assert.Equal(t, "alzada", CodigosDeTipo()[0], "cada llamada da una lista nueva")
	})

	for _, caso := range variantesQueSeRechazan() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			contenido := caso.contenido
			if caso.cambio != nil {
				contenido = caso.cambio.aplicar(t, plazosDelContrato)
			}

			tipos, err := LeerTipos([]byte(contenido))
			require.Error(t, err)
			assert.Nil(t, tipos)
			assert.Contains(t, err.Error(), "data/plazos.yaml: ")
			assert.Contains(t, err.Error(), caso.motivo)

			var conClase schema.ConClase
			assert.NotErrorAs(t, err, &conClase, "un fichero empotrado que no se lee es un defecto inesperado (código 1)")
		})
	}
}

// cambioDeTexto sustituye la primera aparición de un fragmento del fichero.
type cambioDeTexto struct {
	de, a string
}

// aplicar devuelve el fichero con el cambio, y falla si el fragmento no está:
// una variante que no cambia nada pasaría sin probar su regla.
func (c *cambioDeTexto) aplicar(t *testing.T, fichero string) string {
	t.Helper()

	require.Contains(t, fichero, c.de, "la variante no cambia nada")

	return strings.Replace(fichero, c.de, c.a, 1)
}

// varianteDeTipos es un fichero que LeerTipos rechaza: el del contrato con un
// cambio, o un contenido entero, y el motivo que debe nombrar el error.
type varianteDeTipos struct {
	nombre    string
	cambio    *cambioDeTexto
	contenido string
	motivo    string
}

// variantesQueSeRechazan son las variantes del fichero del contrato, una por
// cada regla de LeerTipos.
func variantesQueSeRechazan() []varianteDeTipos {
	const bloqueDeReclamar = `  - codigo: acceso-reclamar
    nombre: reclamación contra la resolución de una solicitud de acceso
    computo: lpac
    duracion: {cantidad: 1, unidad: meses}
    preceptos:
      - {norma: BOE-A-2013-12887, bloque: a24, apartado: "24.2", fija: duracion}
`

	const unidadesDeGenerico = `    unidades:
      dias-habiles: {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.2", fija: computo}
      dias-naturales: {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.2", fija: computo}
      meses: {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.4", fija: computo}
`

	return []varianteDeTipos{
		{nombre: "vacío", contenido: " \n", motivo: "está vacío"},
		{nombre: "no es YAML", contenido: "tipos: [", motivo: "no se puede leer"},
		{nombre: "tipos no es una lista", contenido: "tipos: 7\n", motivo: "no se puede leer"},
		{
			nombre: "clave que el fichero no declara",
			cambio: &cambioDeTexto{de: "    computo: contencioso\n", a: "    computo: contencioso\n    plazo: 2\n"},
			motivo: "no se puede leer",
		},
		{
			nombre: "falta un código",
			cambio: &cambioDeTexto{de: bloqueDeReclamar},
			motivo: `falta el tipo "acceso-reclamar"`,
		},
		{
			nombre: "un código de más",
			cambio: &cambioDeTexto{de: bloqueDeReclamar, a: bloqueDeReclamar + strings.ReplaceAll(
				bloqueDeReclamar, "codigo: acceso-reclamar", "codigo: revision")},
			motivo: `el tipo "revision" no es uno de los siete`,
		},
		{
			nombre: "un código repetido",
			cambio: &cambioDeTexto{de: "codigo: reposicion", a: "codigo: alzada"},
			motivo: `el tipo "alzada" está repetido`,
		},
		{
			nombre: "sin código",
			cambio: &cambioDeTexto{de: "codigo: reposicion", a: `codigo: ""`},
			motivo: "el tipo 2 no tiene código",
		},
		{
			nombre: "sin nombre",
			cambio: &cambioDeTexto{de: "nombre: recurso de alzada", a: `nombre: ""`},
			motivo: `el tipo "alzada" no tiene nombre`,
		},
		{
			nombre: "cómputo desconocido",
			cambio: &cambioDeTexto{de: "computo: contencioso", a: "computo: civil"},
			motivo: `el cómputo "civil" del tipo "contencioso" no es lpac ni contencioso`,
		},
		{
			nombre: "un tipo que fija su duración sin ella",
			cambio: &cambioDeTexto{de: "    computo: lpac\n    duracion: {cantidad: 1, unidad: meses}\n", a: "    computo: lpac\n"},
			motivo: `el tipo "alzada" fija su duración y no la lleva`,
		},
		{
			nombre: "duración en un tipo que no la fija",
			cambio: &cambioDeTexto{
				de: "    nombre: día en que se produce el silencio administrativo\n    computo: lpac\n",
				a:  "    nombre: día en que se produce el silencio administrativo\n    computo: lpac\n    duracion: {cantidad: 3, unidad: meses}\n",
			},
			motivo: `el tipo "silencio" no fija su duración y lleva una`,
		},
		{
			nombre: "cantidad cero",
			cambio: &cambioDeTexto{de: "cantidad: 2", a: "cantidad: 0"},
			motivo: `la duración del tipo "contencioso" no es un entero de 1 o más`,
		},
		{
			nombre: "unidad de la duración desconocida",
			cambio: &cambioDeTexto{de: "cantidad: 2, unidad: meses", a: "cantidad: 2, unidad: semanas"},
			motivo: `la unidad "semanas" del tipo "contencioso" no es meses, dias-habiles ni dias-naturales`,
		},
		{
			nombre: "sin fija: duracion",
			cambio: &cambioDeTexto{de: `apartado: "122.1", fija: duracion`, a: `apartado: "122.1", fija: agosto`},
			motivo: `ningún precepto del tipo "alzada" fija su duración`,
		},
		{
			nombre: "unidades fuera de generico",
			cambio: &cambioDeTexto{
				de: "    preceptos:\n      - {norma: BOE-A-2015-10565, bloque: a122,",
				a: "    unidades:\n      meses: {norma: BOE-A-2015-10565, bloque: a30, apartado: \"30.4\", fija: computo}\n" +
					"    preceptos:\n      - {norma: BOE-A-2015-10565, bloque: a122,",
			},
			motivo: `el tipo "alzada" lleva unidades, que solo lleva generico`,
		},
		{
			nombre: "generico sin unidades",
			cambio: &cambioDeTexto{de: unidadesDeGenerico},
			motivo: `el tipo "generico" no lleva unidades`,
		},
		{
			nombre: "generico con las unidades vacías",
			cambio: &cambioDeTexto{de: unidadesDeGenerico, a: "    unidades: {}\n"},
			motivo: `el tipo "generico" no lleva la unidad "dias-habiles"`,
		},
		{
			nombre: "generico sin una de las tres unidades",
			cambio: &cambioDeTexto{
				de: "      meses: {norma: BOE-A-2015-10565, bloque: a30, apartado: \"30.4\", fija: computo}\n",
			},
			motivo: `el tipo "generico" no lleva la unidad "meses"`,
		},
		{
			nombre: "generico con una unidad desconocida",
			cambio: &cambioDeTexto{
				de: "      meses: {norma",
				a:  "      semanas: {norma: BOE-A-2015-10565, bloque: a30, apartado: \"30.4\", fija: computo}\n      meses: {norma",
			},
			motivo: `el tipo "generico" lleva la unidad "semanas", que no es meses, dias-habiles ni dias-naturales`,
		},
		{
			nombre: "un precepto sin bloque",
			cambio: &cambioDeTexto{de: "bloque: a124,", a: `bloque: "",`},
			motivo: `el precepto 1 del tipo "reposicion" no tiene bloque`,
		},
		{
			nombre: "un precepto de una unidad sin norma",
			cambio: &cambioDeTexto{
				de: "dias-naturales: {norma: BOE-A-2015-10565,",
				a:  `dias-naturales: {norma: "",`,
			},
			motivo: `el precepto de la unidad "dias-naturales" del tipo "generico" no tiene norma`,
		},
		{
			nombre: "un tipo sin precepto",
			cambio: &cambioDeTexto{
				de: "    preceptos:\n      - {norma: BOE-A-2015-10565, bloque: a21, apartado: \"21.3\", fija: inicio}\n" +
					"      - {norma: BOE-A-2015-10565, bloque: a24, apartado: \"24.4\", fija: silencio}\n",
				a: "    preceptos: []\n",
			},
			motivo: `el tipo "silencio" no nombra ningún precepto`,
		},
	}
}
