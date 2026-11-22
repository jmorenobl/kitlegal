package plazos

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// plazosDelContrato es data/plazos.yaml tal como lo fija
// specs/041-h36-el-motor-de/contracts/datos-de-plazos.md §1, sin su cabecera y
// con el cómputo civil, que no tiene tipos (FR-017): cada tipo con lo que da,
// cada cómputo con sus municipios, sus unidades y sus periodos, y cada precepto
// con el día desde el que rige (ADR 0042). Es una copia a propósito: el test
// del dominio no abre ficheros (R1), y el del repositorio lo comprueba contra
// lo grabado del BOE otro test (TestPreceptosDeLosPlazos).
const plazosDelContrato = `tipos:
  - codigo: alzada
    nombre: recurso de alzada
    computo: lpac
    duracion: {cantidad: 1, unidad: meses}
    da: vencimiento
    preceptos:
      - {norma: BOE-A-2015-10565, bloque: a122, apartado: "122.1", fija: duracion, vigente_desde: "2016-10-02"}
  - codigo: reposicion
    nombre: recurso potestativo de reposición
    computo: lpac
    duracion: {cantidad: 1, unidad: meses}
    da: vencimiento
    preceptos:
      - {norma: BOE-A-2015-10565, bloque: a124, apartado: "124.1", fija: duracion, vigente_desde: "2016-10-02"}
  - codigo: contencioso
    nombre: recurso contencioso-administrativo
    computo: contencioso
    duracion: {cantidad: 2, unidad: meses}
    da: vencimiento
    preceptos:
      - {norma: BOE-A-1998-16718, bloque: a46, apartado: "46.1", fija: duracion, vigente_desde: "1998-12-14"}
      - {norma: BOE-A-1998-16718, bloque: a128, apartado: "128.2", fija: agosto, vigente_desde: "1998-12-14"}
  - codigo: acceso-resolver
    nombre: plazo para resolver una solicitud de acceso a la información pública
    computo: lpac
    duracion: {cantidad: 1, unidad: meses}
    da: vencimiento
    preceptos:
      - {norma: BOE-A-2013-12887, bloque: a20, apartado: "20.1", fija: duracion, vigente_desde: "2014-12-10"}
  - codigo: acceso-reclamar
    nombre: reclamación contra la resolución de una solicitud de acceso
    computo: lpac
    duracion: {cantidad: 1, unidad: meses}
    da: vencimiento
    preceptos:
      - {norma: BOE-A-2013-12887, bloque: a24, apartado: "24.2", fija: duracion, vigente_desde: "2014-12-10"}
  - codigo: silencio
    nombre: día en que se produce el silencio administrativo
    computo: lpac
    da: dia-siguiente
    preceptos:
      - {norma: BOE-A-2015-10565, bloque: a21, apartado: "21.3", fija: inicio, vigente_desde: "2016-10-02"}
      - {norma: BOE-A-2015-10565, bloque: a24, apartado: "24.4", fija: silencio, vigente_desde: "2016-10-02"}
  - codigo: generico
    nombre: plazo genérico
    computo: lpac
    da: vencimiento
    unidades:
      dias-habiles: {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.2", fija: computo, vigente_desde: "2016-10-02"}
      dias-naturales: {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.2", fija: computo, vigente_desde: "2016-10-02"}
      meses: {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.4", fija: computo, vigente_desde: "2016-10-02"}
computos:
  - codigo: lpac
    municipios: [el del interesado, el de la sede del órgano]
    unidades: [meses, dias-habiles, dias-naturales]
    preceptos:
      - {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.2", fija: habiles, vigente_desde: "2016-10-02"}
      - {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.3", fija: inicio, vigente_desde: "2016-10-02"}
      - {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.4", fija: meses, vigente_desde: "2016-10-02"}
      - {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.5", fija: prorroga, vigente_desde: "2016-10-02"}
      - {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.6", fija: municipios, vigente_desde: "2016-10-02"}
  - codigo: contencioso
    municipios: [el de la sede del juzgado]
    unidades: [meses]
    periodos:
      - {desde: "08-01", hasta: "08-31", efecto: no-corre, precepto: {norma: BOE-A-1998-16718, apartado: "128.2"}}
      - desde: "12-24"
        hasta: "01-06"
        efecto: aviso
        precepto: {norma: BOE-A-1985-12666, apartado: "183"}
        aviso:
          codigo: dias-del-art-183-lopj
          etiqueta: DÍAS DEL ART. 183 LOPJ
          texto: "el plazo pasa por días del 24 de diciembre al 6 de enero, que el art. 183 LOPJ declara inhábiles para las actuaciones judiciales; se han contado, salvo los inhábiles del art. 182.1 LOPJ, porque el art. 128.2 LJCA no los excluye del plazo para interponer el recurso. La fecha es la más temprana posible."
    preceptos:
      - {norma: BOE-A-1985-12666, bloque: acientoochentaydos, apartado: "182.1", fija: inhabiles, vigente_desde: "2022-12-23"}
      - {norma: BOE-A-1985-12666, bloque: acientoochentaytres, apartado: "183", fija: aviso, vigente_desde: "2022-12-23"}
      - {norma: BOE-A-1985-12666, bloque: acientoochentaycinco, apartado: "185.1", fija: meses, vigente_desde: "1985-07-03"}
      - {norma: BOE-A-1985-12666, bloque: acientoochentaycinco, apartado: "185.2", fija: prorroga, vigente_desde: "1985-07-03"}
      - {norma: BOE-A-1889-4763, bloque: art5, apartado: "5.1", fija: meses, vigente_desde: "1974-07-29"}
  - codigo: civil
    municipios: [el de la sede del tribunal]
    unidades: [meses, dias-habiles]
    periodos:
      - {desde: "08-01", hasta: "08-31", efecto: inhabil, precepto: {norma: BOE-A-2000-323, apartado: "130.2"}}
      - {desde: "12-24", hasta: "01-06", efecto: inhabil, precepto: {norma: BOE-A-2000-323, apartado: "130.2"}}
    preceptos:
      - {norma: BOE-A-2000-323, bloque: a130, apartado: "130.2", fija: inhabiles, vigente_desde: "2022-12-23"}
      - {norma: BOE-A-2000-323, bloque: a133, apartado: "133.1", fija: expiracion, vigente_desde: "2001-01-08"}
      - {norma: BOE-A-2000-323, bloque: a133, apartado: "133.2", fija: habiles, vigente_desde: "2010-05-04"}
      - {norma: BOE-A-2000-323, bloque: a133, apartado: "133.3", fija: meses, vigente_desde: "2001-01-08"}
      - {norma: BOE-A-2000-323, bloque: a133, apartado: "133.4", fija: prorroga, vigente_desde: "2010-05-04"}
`

// textoDelAvisoDelContrato es el texto del aviso del periodo del art. 183 LOPJ
// en la copia del contrato.
const textoDelAvisoDelContrato = "el plazo pasa por días del 24 de diciembre al 6 de enero, que el art. 183 LOPJ " +
	"declara inhábiles para las actuaciones judiciales; se han contado, salvo los inhábiles del art. 182.1 LOPJ, " +
	"porque el art. 128.2 LJCA no los excluye del plazo para interponer el recurso. La fecha es la más temprana " +
	"posible."

// TestLeerTipos lee los siete tipos del contrato, cada uno con su cómputo, y
// admite el civil, que no tiene tipos; y rechaza el fichero en cada una de las
// diez reglas de spec §C.2 y en cada una de las de H9 que no dependen de un
// código, con un error que nombra el fichero, el tipo o el cómputo y lo que
// falla, y que no lleva clase de usuario: el kernel lo da como inesperado
// (FR-013; SC-005; US5).
func TestLeerTipos(t *testing.T) {
	t.Parallel()

	t.Run("la copia del contrato", probarLaCopiaDelContrato)

	t.Run("los vocabularios", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, []Unidad{"meses", "dias-habiles", "dias-naturales"}, Unidades())

		unidades := Unidades()
		unidades[0] = "otra"
		assert.Equal(t, UnidadMeses, Unidades()[0], "cada llamada da una lista nueva")
	})

	t.Run("las reglas 1 y 10 a la vez", func(t *testing.T) {
		t.Parallel()

		contenido := (&cambioDeTexto{de: "computo: lpac", a: "computo: inexistente"}).aplicar(t, plazosDelContrato)
		contenido = (&cambioDeTexto{de: "    da: dia-siguiente\n"}).aplicar(t, contenido)

		tipos, err := LeerTipos([]byte(contenido))
		require.Error(t, err)
		assert.Nil(t, tipos)

		var unidos interface{ Unwrap() []error }
		require.ErrorAs(t, err, &unidos, "los defectos van unidos con errors.Join")

		defectos := unidos.Unwrap()
		require.Len(t, defectos, 2, "los dos defectos y ninguno más: %v", err)
		assert.Contains(t, defectos[0].Error(), `tipo "alzada"`, "en el orden del fichero")
		assert.Contains(t, defectos[0].Error(), `cómputo "inexistente"`)
		assert.Contains(t, defectos[1].Error(), `tipo "silencio"`)
		assert.Contains(t, defectos[1].Error(), "lleva da")
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

			defectos := []error{err}

			var unidos interface{ Unwrap() []error }
			if errors.As(err, &unidos) {
				defectos = unidos.Unwrap()
			}

			for _, defecto := range defectos {
				assert.True(t, strings.HasPrefix(defecto.Error(), "data/plazos.yaml: "), "el defecto nombra el fichero: %q",
					defecto)
			}

			for _, motivo := range caso.motivos {
				assert.Contains(t, err.Error(), motivo)
			}

			var conClase schema.ConClase
			assert.NotErrorAs(t, err, &conClase, "un fichero empotrado que no se lee es un defecto inesperado (código 1)")
		})
	}
}

// probarLaCopiaDelContrato fija que la copia del contrato da los siete tipos,
// en el orden del fichero y cada uno con lo que da y su cómputo copiado, con su
// posición en el fichero; y que el civil, que ningún tipo nombra, no es un
// defecto (US5.1).
func probarLaCopiaDelContrato(t *testing.T) {
	t.Parallel()

	tipos, err := LeerTipos([]byte(plazosDelContrato))
	require.NoError(t, err, "un cómputo sin tipos, como el civil, no es un defecto")

	codigos := make([]string, 0, len(tipos))
	das := make([]string, 0, len(tipos))

	for _, tipo := range tipos {
		codigos = append(codigos, tipo.Codigo)
		das = append(das, tipo.Da)
	}

	assert.Equal(t, []string{
		"alzada", "reposicion", "contencioso", "acceso-resolver", "acceso-reclamar", "silencio", "generico",
	}, codigos, "los siete, en el orden del fichero")
	assert.Equal(t, []string{
		"vencimiento", "vencimiento", "vencimiento", "vencimiento", "vencimiento", "dia-siguiente", "vencimiento",
	}, das)

	lpac, contencioso := computosDelContrato()

	for _, tipo := range tipos {
		assert.NotEqual(t, Computo("civil"), tipo.Computo, "ningún tipo cuenta con el civil")

		switch tipo.Computo {
		case "lpac":
			assert.Equal(t, lpac, tipo.computo, "el tipo %q lleva su cómputo copiado", tipo.Codigo)
		case "contencioso":
			assert.Equal(t, contencioso, tipo.computo, "el tipo %q lleva su cómputo copiado", tipo.Codigo)
		default:
			t.Errorf("el tipo %q cuenta con %q, que no es un cómputo de sus tipos", tipo.Codigo, tipo.Computo)
		}

		assert.Equal(t, tipo.computo.Preceptos, tipo.PreceptosDelComputo,
			"el tipo %q lleva los preceptos de su cómputo, como hoy", tipo.Codigo)
	}

	assert.Equal(t, Tipo{
		Codigo:   "contencioso",
		Nombre:   "recurso contencioso-administrativo",
		Computo:  "contencioso",
		Duracion: &Duracion{Cantidad: 2, Unidad: UnidadMeses},
		Da:       "vencimiento",
		Preceptos: []Precepto{
			{Norma: "BOE-A-1998-16718", Bloque: "a46", Apartado: "46.1", Fija: FijaDuracion, VigenteDesde: "1998-12-14"},
			{Norma: "BOE-A-1998-16718", Bloque: "a128", Apartado: "128.2", Fija: "agosto", VigenteDesde: "1998-12-14"},
		},
		PreceptosDelComputo: contencioso.Preceptos,
		computo:             contencioso,
	}, tipos[2])

	assert.Nil(t, tipos[5].Duracion, "el silencio no fija su duración")
	assert.Nil(t, tipos[5].Unidades)

	const desdeLaLPAC = "2016-10-02"

	assert.Equal(t, map[Unidad]Precepto{
		UnidadDiasHabiles:   {Norma: "BOE-A-2015-10565", Bloque: "a30", Apartado: "30.2", Fija: "computo", VigenteDesde: desdeLaLPAC},
		UnidadDiasNaturales: {Norma: "BOE-A-2015-10565", Bloque: "a30", Apartado: "30.2", Fija: "computo", VigenteDesde: desdeLaLPAC},
		UnidadMeses:         {Norma: "BOE-A-2015-10565", Bloque: "a30", Apartado: "30.4", Fija: "computo", VigenteDesde: desdeLaLPAC},
	}, tipos[6].Unidades)
	assert.Empty(t, tipos[6].Preceptos)
	assert.Nil(t, tipos[6].Duracion)
}

// computosDelContrato son los cómputos lpac y contencioso de la copia del
// contrato, como los copia LeerTipos en cada tipo suyo.
func computosDelContrato() (lpac, contencioso computoDescrito) {
	const desdeLaLPAC = "2016-10-02"

	lpac = computoDescrito{
		Codigo:     "lpac",
		Municipios: []string{"el del interesado", "el de la sede del órgano"},
		Unidades:   []Unidad{UnidadMeses, UnidadDiasHabiles, UnidadDiasNaturales},
		Preceptos: []Precepto{
			{Norma: "BOE-A-2015-10565", Bloque: "a30", Apartado: "30.2", Fija: "habiles", VigenteDesde: desdeLaLPAC},
			{Norma: "BOE-A-2015-10565", Bloque: "a30", Apartado: "30.3", Fija: "inicio", VigenteDesde: desdeLaLPAC},
			{Norma: "BOE-A-2015-10565", Bloque: "a30", Apartado: "30.4", Fija: "meses", VigenteDesde: desdeLaLPAC},
			{Norma: "BOE-A-2015-10565", Bloque: "a30", Apartado: "30.5", Fija: "prorroga", VigenteDesde: desdeLaLPAC},
			{Norma: "BOE-A-2015-10565", Bloque: "a30", Apartado: "30.6", Fija: "municipios", VigenteDesde: desdeLaLPAC},
		},
		posicion: 0,
	}

	contencioso = computoDescrito{
		Codigo:     "contencioso",
		Municipios: []string{"el de la sede del juzgado"},
		Unidades:   []Unidad{UnidadMeses},
		Periodos: []periodo{
			{
				Desde: "08-01", Hasta: "08-31", Efecto: "no-corre",
				Precepto: &referenciaDeUnPrecepto{Norma: "BOE-A-1998-16718", Apartado: "128.2"},
			},
			{
				Desde: "12-24", Hasta: "01-06", Efecto: "aviso",
				Precepto: &referenciaDeUnPrecepto{Norma: "BOE-A-1985-12666", Apartado: "183"},
				Aviso: &avisoDeUnPeriodo{
					Codigo: "dias-del-art-183-lopj", Etiqueta: "DÍAS DEL ART. 183 LOPJ", Texto: textoDelAvisoDelContrato,
				},
			},
		},
		Preceptos: []Precepto{
			{Norma: "BOE-A-1985-12666", Bloque: "acientoochentaydos", Apartado: "182.1", Fija: "inhabiles", VigenteDesde: "2022-12-23"},
			{Norma: "BOE-A-1985-12666", Bloque: "acientoochentaytres", Apartado: "183", Fija: "aviso", VigenteDesde: "2022-12-23"},
			{Norma: "BOE-A-1985-12666", Bloque: "acientoochentaycinco", Apartado: "185.1", Fija: "meses", VigenteDesde: "1985-07-03"},
			{Norma: "BOE-A-1985-12666", Bloque: "acientoochentaycinco", Apartado: "185.2", Fija: "prorroga", VigenteDesde: "1985-07-03"},
			{Norma: "BOE-A-1889-4763", Bloque: "art5", Apartado: "5.1", Fija: "meses", VigenteDesde: "1974-07-29"},
		},
		posicion: 1,
	}

	return lpac, contencioso
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
// cambio, o un contenido entero, y lo que debe nombrar el error: el tipo o el
// cómputo y lo que falla.
type varianteDeTipos struct {
	nombre    string
	cambio    *cambioDeTexto
	contenido string
	motivos   []string
}

// variantesQueSeRechazan son las variantes del fichero del contrato: las de
// cada regla de H9 que no depende de un código y las de cada una de las diez de
// spec §C.2, cada una con la copia rota de contracts/datos-de-plazos.md §2 y,
// si la regla dice más de una cosa, una variante por cada una.
func variantesQueSeRechazan() []varianteDeTipos {
	return append(variantesDeHoy(), variantesDeLasReglas()...)
}

// variantesDeHoy son las de las reglas de H9 que no dependen de un código.
func variantesDeHoy() []varianteDeTipos {
	const preceptosDeLaLPAC = "    preceptos:\n" +
		`      - {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.2", fija: habiles, vigente_desde: "2016-10-02"}` + "\n" +
		`      - {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.3", fija: inicio, vigente_desde: "2016-10-02"}` + "\n" +
		`      - {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.4", fija: meses, vigente_desde: "2016-10-02"}` + "\n" +
		`      - {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.5", fija: prorroga, vigente_desde: "2016-10-02"}` + "\n" +
		`      - {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.6", fija: municipios, vigente_desde: "2016-10-02"}` + "\n"

	return []varianteDeTipos{
		{nombre: "vacío", contenido: " \n", motivos: []string{"está vacío"}},
		{nombre: "no es YAML", contenido: "tipos: [", motivos: []string{"no se puede leer"}},
		{nombre: "tipos no es una lista", contenido: "tipos: 7\n", motivos: []string{"no se puede leer"}},
		{
			nombre:  "clave que un tipo no declara",
			cambio:  &cambioDeTexto{de: "    computo: contencioso\n", a: "    computo: contencioso\n    plazo: 2\n"},
			motivos: []string{"no se puede leer"},
		},
		{
			nombre: "clave que un cómputo no declara",
			cambio: &cambioDeTexto{
				de: "    municipios: [el de la sede del juzgado]\n",
				a:  "    municipios: [el de la sede del juzgado]\n    agosto: no-corre\n",
			},
			motivos: []string{"no se puede leer"},
		},
		{
			nombre: "clave que un periodo no declara",
			cambio: &cambioDeTexto{
				de: `{desde: "08-01", hasta: "08-31", efecto: no-corre,`,
				a:  `{desde: "08-01", hasta: "08-31", dias: 31, efecto: no-corre,`,
			},
			motivos: []string{"no se puede leer"},
		},
		{
			nombre:  "un código repetido",
			cambio:  &cambioDeTexto{de: "codigo: reposicion", a: "codigo: alzada"},
			motivos: []string{`el tipo "alzada" está repetido`},
		},
		{
			nombre:  "sin código",
			cambio:  &cambioDeTexto{de: "codigo: reposicion", a: `codigo: ""`},
			motivos: []string{"el tipo 2 no tiene código"},
		},
		{
			nombre:  "sin nombre",
			cambio:  &cambioDeTexto{de: "nombre: recurso de alzada", a: `nombre: ""`},
			motivos: []string{`el tipo "alzada" no tiene nombre`},
		},
		{
			nombre:  "cantidad cero",
			cambio:  &cambioDeTexto{de: "cantidad: 2", a: "cantidad: 0"},
			motivos: []string{`la duración del tipo "contencioso" no es un entero de 1 o más`},
		},
		{
			nombre:  "unidad de la duración desconocida",
			cambio:  &cambioDeTexto{de: "cantidad: 2, unidad: meses", a: "cantidad: 2, unidad: semanas"},
			motivos: []string{`la unidad "semanas" del tipo "contencioso" no es meses, dias-habiles ni dias-naturales`},
		},
		{
			nombre:  "sin fija: duracion",
			cambio:  &cambioDeTexto{de: `apartado: "122.1", fija: duracion`, a: `apartado: "122.1", fija: agosto`},
			motivos: []string{`ningún precepto del tipo "alzada" fija su duración`},
		},
		{
			nombre:  "un precepto sin bloque",
			cambio:  &cambioDeTexto{de: "bloque: a124,", a: `bloque: "",`},
			motivos: []string{`el precepto 1 del tipo "reposicion" no tiene bloque`},
		},
		{
			nombre: "un precepto de una unidad sin norma",
			cambio: &cambioDeTexto{
				de: "dias-naturales: {norma: BOE-A-2015-10565,",
				a:  `dias-naturales: {norma: "",`,
			},
			motivos: []string{`el precepto de la unidad "dias-naturales" del tipo "generico" no tiene norma`},
		},
		{
			nombre: "un tipo sin precepto",
			cambio: &cambioDeTexto{
				de: "    preceptos:\n      - {norma: BOE-A-2015-10565, bloque: a21, apartado: \"21.3\", fija: inicio, " +
					"vigente_desde: \"2016-10-02\"}\n      - {norma: BOE-A-2015-10565, bloque: a24, apartado: \"24.4\", " +
					"fija: silencio, vigente_desde: \"2016-10-02\"}\n",
				a: "    preceptos: []\n",
			},
			motivos: []string{`el tipo "silencio" no nombra ningún precepto`},
		},
		{
			nombre:  "un precepto sin vigente_desde",
			cambio:  &cambioDeTexto{de: `fija: duracion, vigente_desde: "2016-10-02"}`, a: "fija: duracion}"},
			motivos: []string{`el precepto 1 del tipo "alzada" no tiene vigente_desde`},
		},
		{
			nombre: "un vigente_desde que no es un día",
			cambio: &cambioDeTexto{
				de: `apartado: "128.2", fija: agosto, vigente_desde: "1998-12-14"`,
				a:  `apartado: "128.2", fija: agosto, vigente_desde: "1998-02-30"`,
			},
			motivos: []string{
				`el precepto 2 del tipo "contencioso" tiene un vigente_desde, "1998-02-30", que no es una fecha AAAA-MM-DD`,
			},
		},
		{
			nombre: "un vigente_desde sin sus ceros",
			cambio: &cambioDeTexto{
				de: `apartado: "183", fija: aviso, vigente_desde: "2022-12-23"`,
				a:  `apartado: "183", fija: aviso, vigente_desde: "2022-12-3"`,
			},
			motivos: []string{
				`el precepto 2 del cómputo "contencioso" tiene un vigente_desde, "2022-12-3", que no es una fecha AAAA-MM-DD`,
			},
		},
		{
			nombre:  "sin la sección de los cómputos",
			cambio:  &cambioDeTexto{de: "computos:\n", a: "otros:\n"},
			motivos: []string{"no se puede leer"},
		},
		{
			nombre:  "un cómputo repetido",
			cambio:  &cambioDeTexto{de: "  - codigo: civil\n", a: "  - codigo: lpac\n"},
			motivos: []string{`el cómputo "lpac" está repetido en computos`},
		},
		{
			nombre:  "un cómputo sin código",
			cambio:  &cambioDeTexto{de: "  - codigo: civil\n", a: "  - codigo: \"\"\n"},
			motivos: []string{"el cómputo 3 no tiene código"},
		},
		{
			nombre:  "un cómputo sin preceptos",
			cambio:  &cambioDeTexto{de: preceptosDeLaLPAC, a: "    preceptos: []\n"},
			motivos: []string{`el cómputo "lpac" no nombra ningún precepto`},
		},
		{
			nombre:  "un precepto del cómputo sin apartado",
			cambio:  &cambioDeTexto{de: `apartado: "5.1",`, a: `apartado: "",`},
			motivos: []string{`el precepto 5 del cómputo "contencioso" no tiene apartado`},
		},
	}
}

// variantesDeLasReglas son las de las diez reglas de spec §C.2. La primera de
// cada regla es la copia rota de la tabla de contracts/datos-de-plazos.md §2.
func variantesDeLasReglas() []varianteDeTipos {
	const lineaDelTexto = `          texto: "el plazo pasa por días del 24 de diciembre al 6 de enero, que el art. 183 LOPJ ` +
		`declara inhábiles para las actuaciones judiciales; se han contado, salvo los inhábiles del art. 182.1 LOPJ, ` +
		`porque el art. 128.2 LJCA no los excluye del plazo para interponer el recurso. La fecha es la más temprana ` +
		`posible."` + "\n"

	const segundoDelCivil = `{desde: "12-24", hasta: "01-06", efecto: inhabil, precepto: {norma: BOE-A-2000-323, ` +
		`apartado: "130.2"}}`

	return []varianteDeTipos{
		{
			nombre:  "regla 1 · el cómputo del tipo existe",
			cambio:  &cambioDeTexto{de: "computo: lpac", a: "computo: inexistente"},
			motivos: []string{`tipo "alzada"`, `cómputo "inexistente"`},
		},
		{
			nombre:  "regla 2 · la unidad de la duración la admite su cómputo",
			cambio:  &cambioDeTexto{de: "cantidad: 2, unidad: meses", a: "cantidad: 2, unidad: dias-habiles"},
			motivos: []string{`tipo "contencioso"`, `unidad "dias-habiles"`, `cómputo, "contencioso"`},
		},
		{
			nombre: "regla 3 · las unidades del tipo son las de su cómputo",
			cambio: &cambioDeTexto{
				de: `      meses: {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.4", fija: computo, ` +
					`vigente_desde: "2016-10-02"}` + "\n",
			},
			motivos: []string{`tipo "generico"`, `unidad "meses"`, `cómputo, "lpac"`},
		},
		{
			nombre: "regla 3 · una unidad que su cómputo no admite",
			cambio: &cambioDeTexto{
				de: "      meses: {norma",
				a: `      semanas: {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.4", fija: computo, ` +
					`vigente_desde: "2016-10-02"}` + "\n      meses: {norma",
			},
			motivos: []string{`tipo "generico"`, `unidad "semanas"`, `cómputo, "lpac"`},
		},
		{
			nombre: "regla 3 · las unidades vacías",
			cambio: &cambioDeTexto{
				de: "    unidades:\n" +
					`      dias-habiles: {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.2", fija: computo, ` +
					`vigente_desde: "2016-10-02"}` + "\n" +
					`      dias-naturales: {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.2", fija: computo, ` +
					`vigente_desde: "2016-10-02"}` + "\n" +
					`      meses: {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.4", fija: computo, ` +
					`vigente_desde: "2016-10-02"}` + "\n",
				a: "    unidades: {}\n",
			},
			motivos: []string{`tipo "generico"`, `unidad "meses"`, `unidad "dias-habiles"`, `unidad "dias-naturales"`},
		},
		{
			nombre: "regla 3 · unidades y duración",
			cambio: &cambioDeTexto{
				de: "    preceptos:\n      - {norma: BOE-A-2015-10565, bloque: a122,",
				a: "    unidades:\n" +
					`      meses: {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.4", fija: computo, ` +
					`vigente_desde: "2016-10-02"}` + "\n" +
					"    preceptos:\n      - {norma: BOE-A-2015-10565, bloque: a122,",
			},
			motivos: []string{`tipo "alzada"`, "unidades y duración"},
		},
		{
			nombre: "regla 4 · el precepto de un periodo es del cómputo o de cada tipo suyo",
			cambio: &cambioDeTexto{
				de: `precepto: {norma: BOE-A-1998-16718, apartado: "128.2"}`,
				a:  `precepto: {norma: BOE-A-1998-16718, apartado: "128.3"}`,
			},
			motivos: []string{`periodo 1 del cómputo "contencioso"`, "BOE-A-1998-16718 128.3"},
		},
		{
			nombre: "regla 4 · el de un solo tipo de un cómputo con varios",
			cambio: &cambioDeTexto{
				de: "    unidades: [meses, dias-habiles, dias-naturales]\n",
				a: "    unidades: [meses, dias-habiles, dias-naturales]\n    periodos:\n" +
					`      - {desde: "03-01", hasta: "03-02", efecto: inhabil, ` +
					`precepto: {norma: BOE-A-2015-10565, apartado: "122.1"}}` + "\n",
			},
			motivos: []string{`periodo 1 del cómputo "lpac"`, "BOE-A-2015-10565 122.1"},
		},
		{
			nombre: "regla 4 · en un cómputo sin tipos",
			cambio: &cambioDeTexto{
				de: `efecto: inhabil, precepto: {norma: BOE-A-2000-323, apartado: "130.2"}}`,
				a:  `efecto: inhabil, precepto: {norma: BOE-A-2000-323, apartado: "130.3"}}`,
			},
			motivos: []string{`periodo 1 del cómputo "civil"`, "BOE-A-2000-323 130.3"},
		},
		{
			nombre: "regla 5 · un no-corre es un mes entero",
			cambio: &cambioDeTexto{
				de: `{desde: "08-01", hasta: "08-31", efecto: no-corre`,
				a:  `{desde: "08-01", hasta: "08-30", efecto: no-corre`,
			},
			motivos: []string{`periodo 1 del cómputo "contencioso"`, "mes entero"},
		},
		{
			nombre: "regla 5 · desde el día 1",
			cambio: &cambioDeTexto{
				de: `{desde: "08-01", hasta: "08-31", efecto: no-corre`,
				a:  `{desde: "08-02", hasta: "08-31", efecto: no-corre`,
			},
			motivos: []string{`periodo 1 del cómputo "contencioso"`, "mes entero"},
		},
		{
			nombre: "regla 5 · de 30 o 31 días",
			cambio: &cambioDeTexto{
				de: `{desde: "08-01", hasta: "08-31", efecto: no-corre`,
				a:  `{desde: "02-01", hasta: "02-28", efecto: no-corre`,
			},
			motivos: []string{`periodo 1 del cómputo "contencioso"`, "mes entero"},
		},
		{
			nombre: "regla 5 · en un cómputo de solo meses",
			cambio: &cambioDeTexto{
				de: `hasta: "08-31", efecto: inhabil`,
				a:  `hasta: "08-31", efecto: no-corre`,
			},
			motivos: []string{`periodo 1 del cómputo "civil"`, "otras unidades que meses"},
		},
		{
			nombre: "regla 6 · dos periodos no comparten ningún día",
			cambio: &cambioDeTexto{
				de: `{desde: "12-24", hasta: "01-06", efecto: inhabil`,
				a:  `{desde: "08-31", hasta: "01-06", efecto: inhabil`,
			},
			motivos: []string{`periodos 1 y 2 del cómputo "civil"`, "08-31"},
		},
		{
			nombre: "regla 6 · al cruzar el año",
			cambio: &cambioDeTexto{
				de: `{desde: "08-01", hasta: "08-31", efecto: inhabil`,
				a:  `{desde: "01-01", hasta: "01-31", efecto: inhabil`,
			},
			motivos: []string{`periodos 1 y 2 del cómputo "civil"`, "01-01"},
		},
		{
			nombre:  "regla 7 · el aviso de un periodo no es del motor",
			cambio:  &cambioDeTexto{de: "codigo: dias-del-art-183-lopj", a: "codigo: festivo-de-una-parte"},
			motivos: []string{`aviso "festivo-de-una-parte"`, `periodo 2 del cómputo "contencioso"`, "del motor"},
		},
		{
			nombre:  "regla 7 · la etiqueta de un aviso del motor",
			cambio:  &cambioDeTexto{de: "etiqueta: DÍAS DEL ART. 183 LOPJ", a: "etiqueta: FESTIVO DE UNA PARTE"},
			motivos: []string{`aviso "dias-del-art-183-lopj"`, `periodo 2 del cómputo "contencioso"`, "etiqueta"},
		},
		{
			nombre: "regla 7 · dos avisos con el mismo código",
			cambio: &cambioDeTexto{
				de: segundoDelCivil,
				a: `{desde: "12-24", hasta: "01-06", efecto: aviso, precepto: {norma: BOE-A-2000-323, ` +
					`apartado: "130.2"}, aviso: {codigo: dias-del-art-183-lopj, etiqueta: OTRA, texto: otro}}`,
			},
			motivos: []string{`aviso "dias-del-art-183-lopj"`, `periodo 2 del cómputo "civil"`, "código"},
		},
		{
			nombre: "regla 7 · dos avisos con la misma etiqueta",
			cambio: &cambioDeTexto{
				de: segundoDelCivil,
				a: `{desde: "12-24", hasta: "01-06", efecto: aviso, precepto: {norma: BOE-A-2000-323, ` +
					`apartado: "130.2"}, aviso: {codigo: otro-aviso, etiqueta: DÍAS DEL ART. 183 LOPJ, texto: otro}}`,
			},
			motivos: []string{`aviso "otro-aviso"`, `periodo 2 del cómputo "civil"`, "etiqueta"},
		},
		{
			nombre:  "regla 8 · la gramática de un código de tipo",
			cambio:  &cambioDeTexto{de: "codigo: alzada", a: "codigo: Alzada"},
			motivos: []string{`"Alzada"`, "tipo 1"},
		},
		{
			nombre:  "regla 8 · la de un código de cómputo",
			cambio:  &cambioDeTexto{de: "  - codigo: civil\n", a: "  - codigo: civil-\n"},
			motivos: []string{`"civil-"`, "cómputo 3"},
		},
		{
			nombre:  "regla 8 · la de un código de aviso",
			cambio:  &cambioDeTexto{de: "codigo: dias-del-art-183-lopj", a: "codigo: dias--del-art-183-lopj"},
			motivos: []string{`"dias--del-art-183-lopj"`, `periodo 2 del cómputo "contencioso"`},
		},
		{
			nombre: "regla 9 · uno o dos municipios",
			cambio: &cambioDeTexto{
				de: "municipios: [el del interesado, el de la sede del órgano]",
				a:  "municipios: [el del interesado, el de la sede del órgano, el del representante]",
			},
			motivos: []string{`cómputo "lpac"`, "municipios"},
		},
		{
			nombre:  "regla 9 · sin municipios",
			cambio:  &cambioDeTexto{de: "    municipios: [el de la sede del juzgado]\n"},
			motivos: []string{`cómputo "contencioso"`, "municipios"},
		},
		{
			nombre: "regla 9 · un municipio vacío",
			cambio: &cambioDeTexto{
				de: "municipios: [el del interesado, el de la sede del órgano]",
				a:  `municipios: ["", el de la sede del órgano]`,
			},
			motivos: []string{`cómputo "lpac"`, "municipio vacío"},
		},
		{
			nombre:  "regla 9 · sin unidades",
			cambio:  &cambioDeTexto{de: "    unidades: [meses, dias-habiles]\n", a: "    unidades: []\n"},
			motivos: []string{`cómputo "civil"`, "unidades"},
		},
		{
			nombre:  "regla 9 · una unidad repetida",
			cambio:  &cambioDeTexto{de: "    unidades: [meses, dias-habiles]\n", a: "    unidades: [meses, meses]\n"},
			motivos: []string{`cómputo "civil"`, `unidad "meses"`},
		},
		{
			nombre:  "regla 9 · una unidad que no es del motor",
			cambio:  &cambioDeTexto{de: "    unidades: [meses, dias-habiles]\n", a: "    unidades: [meses, semanas]\n"},
			motivos: []string{`cómputo "civil"`, `unidad "semanas"`},
		},
		{
			nombre: "regla 9 · un periodo sin desde",
			cambio: &cambioDeTexto{
				de: `{desde: "08-01", hasta: "08-31", efecto: no-corre`,
				a:  `{hasta: "08-31", efecto: no-corre`,
			},
			motivos: []string{`periodo 1 del cómputo "contencioso"`, "desde"},
		},
		{
			nombre:  "regla 9 · un periodo sin efecto",
			cambio:  &cambioDeTexto{de: `hasta: "08-31", efecto: inhabil, `, a: `hasta: "08-31", `},
			motivos: []string{`periodo 1 del cómputo "civil"`, "efecto"},
		},
		{
			nombre: "regla 9 · un periodo sin precepto",
			cambio: &cambioDeTexto{
				de: `efecto: inhabil, precepto: {norma: BOE-A-2000-323, apartado: "130.2"}}`,
				a:  `efecto: inhabil}`,
			},
			motivos: []string{`periodo 1 del cómputo "civil"`, "precepto"},
		},
		{
			nombre: "regla 9 · el precepto de un periodo sin apartado",
			cambio: &cambioDeTexto{
				de: `efecto: inhabil, precepto: {norma: BOE-A-2000-323, apartado: "130.2"}}`,
				a:  `efecto: inhabil, precepto: {norma: BOE-A-2000-323}}`,
			},
			motivos: []string{`periodo 1 del cómputo "civil"`, "apartado"},
		},
		{
			nombre: "regla 9 · un día que no tienen todos los años",
			cambio: &cambioDeTexto{
				de: `{desde: "08-01", hasta: "08-31", efecto: inhabil`,
				a:  `{desde: "02-29", hasta: "08-31", efecto: inhabil`,
			},
			motivos: []string{`periodo 1 del cómputo "civil"`, `"02-29"`},
		},
		{
			nombre: "regla 9 · un día sin sus ceros",
			cambio: &cambioDeTexto{
				de: `{desde: "08-01", hasta: "08-31", efecto: inhabil`,
				a:  `{desde: "08-01", hasta: "8-31", efecto: inhabil`,
			},
			motivos: []string{`periodo 1 del cómputo "civil"`, `"8-31"`},
		},
		{
			nombre:  "regla 9 · un efecto que no es de los tres",
			cambio:  &cambioDeTexto{de: `hasta: "08-31", efecto: inhabil`, a: `hasta: "08-31", efecto: festivo`},
			motivos: []string{`periodo 1 del cómputo "civil"`, `"festivo"`},
		},
		{
			nombre:  "regla 9 · un periodo de aviso sin su aviso",
			cambio:  &cambioDeTexto{de: `hasta: "08-31", efecto: inhabil`, a: `hasta: "08-31", efecto: aviso`},
			motivos: []string{`periodo 1 del cómputo "civil"`, "no lleva su aviso"},
		},
		{
			nombre:  "regla 9 · un aviso en un periodo que no es de aviso",
			cambio:  &cambioDeTexto{de: "        efecto: aviso\n", a: "        efecto: inhabil\n"},
			motivos: []string{`periodo 2 del cómputo "contencioso"`, "lleva un aviso"},
		},
		{
			nombre:  "regla 9 · un aviso sin código",
			cambio:  &cambioDeTexto{de: "codigo: dias-del-art-183-lopj", a: `codigo: ""`},
			motivos: []string{`periodo 2 del cómputo "contencioso"`, "código"},
		},
		{
			nombre:  "regla 9 · un aviso sin etiqueta",
			cambio:  &cambioDeTexto{de: "etiqueta: DÍAS DEL ART. 183 LOPJ", a: `etiqueta: ""`},
			motivos: []string{`periodo 2 del cómputo "contencioso"`, "etiqueta"},
		},
		{
			nombre:  "regla 9 · un aviso sin texto",
			cambio:  &cambioDeTexto{de: lineaDelTexto, a: `          texto: ""` + "\n"},
			motivos: []string{`periodo 2 del cómputo "contencioso"`, "texto"},
		},
		{
			nombre:  "regla 9 · una etiqueta con dos puntos",
			cambio:  &cambioDeTexto{de: "etiqueta: DÍAS DEL ART. 183 LOPJ", a: `etiqueta: "DÍAS: ART. 183 LOPJ"`},
			motivos: []string{`periodo 2 del cómputo "contencioso"`, "dos puntos"},
		},
		{
			nombre:  "regla 9 · una etiqueta de dos líneas",
			cambio:  &cambioDeTexto{de: "etiqueta: DÍAS DEL ART. 183 LOPJ", a: `etiqueta: "DÍAS DEL ART.\n183 LOPJ"`},
			motivos: []string{`periodo 2 del cómputo "contencioso"`, "salto de línea", "etiqueta"},
		},
		{
			nombre:  "regla 9 · un texto de dos líneas",
			cambio:  &cambioDeTexto{de: `texto: "el plazo pasa por`, a: `texto: "el plazo\npasa por`},
			motivos: []string{`periodo 2 del cómputo "contencioso"`, "salto de línea", "texto"},
		},
		{
			nombre:  "regla 10 · da",
			cambio:  &cambioDeTexto{de: "    duracion: {cantidad: 1, unidad: meses}\n    da: vencimiento\n", a: "    duracion: {cantidad: 1, unidad: meses}\n"},
			motivos: []string{`tipo "alzada"`, "lleva da"},
		},
		{
			nombre:  "regla 10 · uno de sus dos valores",
			cambio:  &cambioDeTexto{de: "da: vencimiento", a: "da: siempre"},
			motivos: []string{`tipo "alzada"`, `"siempre"`},
		},
	}
}
