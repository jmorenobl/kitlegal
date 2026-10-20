package plazos

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// plazosDelContrato es data/plazos.yaml tal como lo fija contracts/datos-de-plazos.md
// §2, con el día desde el que rige cada precepto y los preceptos de cada
// cómputo del ADR 0042. Es una copia a propósito: el test del dominio no abre ficheros (R1), y el
// del repositorio lo comprueba contra lo grabado del BOE otro test
// (TestPreceptosDeLosPlazos).
const plazosDelContrato = `tipos:
  - codigo: alzada
    nombre: recurso de alzada
    computo: lpac
    duracion: {cantidad: 1, unidad: meses}
    preceptos:
      - {norma: BOE-A-2015-10565, bloque: a122, apartado: "122.1", fija: duracion, vigente_desde: "2016-10-02"}
  - codigo: reposicion
    nombre: recurso potestativo de reposición
    computo: lpac
    duracion: {cantidad: 1, unidad: meses}
    preceptos:
      - {norma: BOE-A-2015-10565, bloque: a124, apartado: "124.1", fija: duracion, vigente_desde: "2016-10-02"}
  - codigo: contencioso
    nombre: recurso contencioso-administrativo
    computo: contencioso
    duracion: {cantidad: 2, unidad: meses}
    preceptos:
      - {norma: BOE-A-1998-16718, bloque: a46, apartado: "46.1", fija: duracion, vigente_desde: "1998-12-14"}
      - {norma: BOE-A-1998-16718, bloque: a128, apartado: "128.2", fija: agosto, vigente_desde: "1998-12-14"}
  - codigo: acceso-resolver
    nombre: plazo para resolver una solicitud de acceso a la información pública
    computo: lpac
    duracion: {cantidad: 1, unidad: meses}
    preceptos:
      - {norma: BOE-A-2013-12887, bloque: a20, apartado: "20.1", fija: duracion, vigente_desde: "2014-12-10"}
  - codigo: acceso-reclamar
    nombre: reclamación contra la resolución de una solicitud de acceso
    computo: lpac
    duracion: {cantidad: 1, unidad: meses}
    preceptos:
      - {norma: BOE-A-2013-12887, bloque: a24, apartado: "24.2", fija: duracion, vigente_desde: "2014-12-10"}
  - codigo: silencio
    nombre: día en que se produce el silencio administrativo
    computo: lpac
    preceptos:
      - {norma: BOE-A-2015-10565, bloque: a21, apartado: "21.3", fija: inicio, vigente_desde: "2016-10-02"}
      - {norma: BOE-A-2015-10565, bloque: a24, apartado: "24.4", fija: silencio, vigente_desde: "2016-10-02"}
  - codigo: generico
    nombre: plazo genérico
    computo: lpac
    unidades:
      dias-habiles: {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.2", fija: computo, vigente_desde: "2016-10-02"}
      dias-naturales: {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.2", fija: computo, vigente_desde: "2016-10-02"}
      meses: {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.4", fija: computo, vigente_desde: "2016-10-02"}
computos:
  - codigo: lpac
    preceptos:
      - {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.2", fija: habiles, vigente_desde: "2016-10-02"}
      - {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.3", fija: inicio, vigente_desde: "2016-10-02"}
      - {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.4", fija: meses, vigente_desde: "2016-10-02"}
      - {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.5", fija: prorroga, vigente_desde: "2016-10-02"}
      - {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.6", fija: municipios, vigente_desde: "2016-10-02"}
  - codigo: contencioso
    preceptos:
      - {norma: BOE-A-1985-12666, bloque: acientoochentaydos, apartado: "182.1", fija: inhabiles, vigente_desde: "2022-12-23"}
      - {norma: BOE-A-1985-12666, bloque: acientoochentaytres, apartado: "183", fija: aviso, vigente_desde: "2022-12-23"}
      - {norma: BOE-A-1985-12666, bloque: acientoochentaycinco, apartado: "185.1", fija: meses, vigente_desde: "1985-07-03"}
      - {norma: BOE-A-1985-12666, bloque: acientoochentaycinco, apartado: "185.2", fija: prorroga, vigente_desde: "1985-07-03"}
      - {norma: BOE-A-1889-4763, bloque: art5, apartado: "5.1", fija: meses, vigente_desde: "1974-07-29"}
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
				{Norma: "BOE-A-1998-16718", Bloque: "a46", Apartado: "46.1", Fija: FijaDuracion, VigenteDesde: "1998-12-14"},
				{Norma: "BOE-A-1998-16718", Bloque: "a128", Apartado: "128.2", Fija: "agosto", VigenteDesde: "1998-12-14"},
			},
			PreceptosDelComputo: []Precepto{
				{Norma: "BOE-A-1985-12666", Bloque: "acientoochentaydos", Apartado: "182.1", Fija: "inhabiles", VigenteDesde: "2022-12-23"},
				{Norma: "BOE-A-1985-12666", Bloque: "acientoochentaytres", Apartado: "183", Fija: "aviso", VigenteDesde: "2022-12-23"},
				{Norma: "BOE-A-1985-12666", Bloque: "acientoochentaycinco", Apartado: "185.1", Fija: "meses", VigenteDesde: "1985-07-03"},
				{Norma: "BOE-A-1985-12666", Bloque: "acientoochentaycinco", Apartado: "185.2", Fija: "prorroga", VigenteDesde: "1985-07-03"},
				{Norma: "BOE-A-1889-4763", Bloque: "art5", Apartado: "5.1", Fija: "meses", VigenteDesde: "1974-07-29"},
			},
		}, tipos[2])

		for _, tipo := range tipos {
			if tipo.Computo == ComputoLPAC {
				assert.Len(t, tipo.PreceptosDelComputo, 5, "el tipo %q lleva los cinco apartados del art. 30 LPAC",
					tipo.Codigo)
			}
		}

		assert.Nil(t, tipos[5].Duracion, "el silencio no fija su duración")
		assert.Nil(t, tipos[5].Unidades)
		const desdeLaLPAC = "2016-10-02"

		assert.Equal(t, map[Unidad]Precepto{
			UnidadDiasHabiles:   {Norma: "BOE-A-2015-10565", Bloque: "a30", Apartado: "30.2", Fija: "computo", VigenteDesde: desdeLaLPAC},
			UnidadDiasNaturales: {Norma: "BOE-A-2015-10565", Bloque: "a30", Apartado: "30.2", Fija: "computo", VigenteDesde: desdeLaLPAC},
			UnidadMeses:         {Norma: "BOE-A-2015-10565", Bloque: "a30", Apartado: "30.4", Fija: "computo", VigenteDesde: desdeLaLPAC},
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
      - {norma: BOE-A-2013-12887, bloque: a24, apartado: "24.2", fija: duracion, vigente_desde: "2014-12-10"}
`

	const unidadesDeGenerico = `    unidades:
      dias-habiles: {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.2", fija: computo, vigente_desde: "2016-10-02"}
      dias-naturales: {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.2", fija: computo, vigente_desde: "2016-10-02"}
      meses: {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.4", fija: computo, vigente_desde: "2016-10-02"}
`

	const computoContencioso = "  - codigo: contencioso\n    preceptos:\n"

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
				a: "    unidades:\n      meses: {norma: BOE-A-2015-10565, bloque: a30, apartado: \"30.4\", fija: computo, " +
					"vigente_desde: \"2016-10-02\"}\n" +
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
				de: "      meses: {norma: BOE-A-2015-10565, bloque: a30, apartado: \"30.4\", fija: computo, " +
					"vigente_desde: \"2016-10-02\"}\n",
			},
			motivo: `el tipo "generico" no lleva la unidad "meses"`,
		},
		{
			nombre: "generico con una unidad desconocida",
			cambio: &cambioDeTexto{
				de: "      meses: {norma",
				a: "      semanas: {norma: BOE-A-2015-10565, bloque: a30, apartado: \"30.4\", fija: computo, " +
					"vigente_desde: \"2016-10-02\"}\n      meses: {norma",
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
				de: "    preceptos:\n      - {norma: BOE-A-2015-10565, bloque: a21, apartado: \"21.3\", fija: inicio, " +
					"vigente_desde: \"2016-10-02\"}\n      - {norma: BOE-A-2015-10565, bloque: a24, apartado: \"24.4\", " +
					"fija: silencio, vigente_desde: \"2016-10-02\"}\n",
				a: "    preceptos: []\n",
			},
			motivo: `el tipo "silencio" no nombra ningún precepto`,
		},
		{
			nombre: "un precepto sin vigente_desde",
			cambio: &cambioDeTexto{de: `fija: duracion, vigente_desde: "2016-10-02"}`, a: "fija: duracion}"},
			motivo: `el precepto 1 del tipo "alzada" no tiene vigente_desde`,
		},
		{
			nombre: "un vigente_desde que no es un día",
			cambio: &cambioDeTexto{
				de: `apartado: "128.2", fija: agosto, vigente_desde: "1998-12-14"`,
				a:  `apartado: "128.2", fija: agosto, vigente_desde: "1998-02-30"`,
			},
			motivo: `el precepto 2 del tipo "contencioso" tiene un vigente_desde, "1998-02-30", que no es una fecha AAAA-MM-DD`,
		},
		{
			nombre: "un vigente_desde sin sus ceros",
			cambio: &cambioDeTexto{
				de: `apartado: "183", fija: aviso, vigente_desde: "2022-12-23"`,
				a:  `apartado: "183", fija: aviso, vigente_desde: "2022-12-3"`,
			},
			motivo: `el precepto 2 del cómputo "contencioso" tiene un vigente_desde, "2022-12-3", que no es una fecha AAAA-MM-DD`,
		},
		{
			nombre: "sin la sección de los cómputos",
			cambio: &cambioDeTexto{de: "computos:\n", a: "otros:\n"},
			motivo: "no se puede leer",
		},
		{
			nombre: "falta un cómputo",
			cambio: &cambioDeTexto{de: computoContencioso, a: "  - codigo: civil\n    preceptos:\n"},
			motivo: `falta el cómputo "contencioso" en computos`,
		},
		{
			nombre: "un cómputo desconocido",
			cambio: &cambioDeTexto{de: computoContencioso, a: "  - codigo: civil\n    preceptos:\n"},
			motivo: `el cómputo "civil" de computos no es lpac ni contencioso`,
		},
		{
			nombre: "un cómputo repetido",
			cambio: &cambioDeTexto{de: computoContencioso, a: "  - codigo: lpac\n    preceptos:\n"},
			motivo: `el cómputo "lpac" está repetido en computos`,
		},
		{
			nombre: "un cómputo sin preceptos",
			cambio: &cambioDeTexto{de: computoContencioso, a: "  - codigo: contencioso\n    preceptos: []\n  - codigo: otro\n    preceptos:\n"},
			motivo: `el cómputo "contencioso" no nombra ningún precepto`,
		},
		{
			nombre: "un precepto del cómputo sin apartado",
			cambio: &cambioDeTexto{de: `apartado: "5.1",`, a: `apartado: "",`},
			motivo: `el precepto 5 del cómputo "contencioso" no tiene apartado`,
		},
	}
}
