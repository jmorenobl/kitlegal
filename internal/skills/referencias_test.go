package skills_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/plazos"
	"github.com/jmorenobl/kitlegal/internal/skills"
)

// comienzoDeLaReferencia es lo que abre toda referencia de normas, hasta la fila
// de separación de su tabla (contrato normas-y-referencias §5).
const comienzoDeLaReferencia = "<!-- generado desde data/normas.yaml, no editar -->\n" +
	"\n" +
	"# Normas de referencia\n" +
	"\n" +
	"| Norma | Abreviatura | Identificador | Rango | Materias |\n" +
	"|---|---|---|---|---|\n"

// referenciaDeLasNormasDePrueba son los bytes exactos de la referencia de
// normasDePrueba: por año y número, los dos por su valor; la barra escrita \|;
// cada salto de línea, \n, \r\n o \r, como un espacio; y la abreviatura ausente
// como una celda vacía.
const referenciaDeLasNormasDePrueba = comienzoDeLaReferencia +
	"| Ley 7/1985, de 2 de abril, reguladora de las Bases del Régimen Local. | LRBRL | `BOE-A-1985-5392` | Ley | régimen local |\n" +
	"| Norma de prueba con un salto de línea. | NP\\|9999 | `BOE-A-2015-9999` | Real Decreto | primera materia, segunda \\| materia |\n" +
	"| Norma de prueba con saltos de línea de Windows y de retorno de carro. | NP | `BOE-A-2015-10565` | Real Decreto-ley | procedimiento administrativo, tributos |\n" +
	"| Norma de prueba con una barra \\| en el título. |  | `BOE-A-2015-10566` | Ley | régimen jurídico del sector público |\n"

// normasDePrueba son las normas de TestRenderizarNormas en un orden que no es el
// de la referencia: 9999 va antes que 10565 por su valor, aunque como texto irían
// al revés, y 10565 antes que 10566.
func normasDePrueba() []skills.Norma {
	return []skills.Norma{
		{
			Identificador: "BOE-A-2015-10566",
			Titulo:        "Norma de prueba con una barra | en el título.",
			Rango:         "Ley",
			Materias:      []string{"régimen jurídico del sector público"},
		},
		{
			Identificador: "BOE-A-2015-9999",
			Titulo:        "Norma de prueba con un salto\nde línea.",
			Rango:         "Real Decreto",
			Abreviatura:   "NP|9999",
			Materias:      []string{"primera materia", "segunda | materia"},
		},
		{
			Identificador: "BOE-A-1985-5392",
			Titulo:        "Ley 7/1985, de 2 de abril, reguladora de las Bases del Régimen Local.",
			Rango:         "Ley",
			Abreviatura:   "LRBRL",
			Materias:      []string{"régimen local"},
		},
		{
			Identificador: "BOE-A-2015-10565",
			Titulo:        "Norma de prueba con saltos\r\nde línea de Windows\ry de retorno de carro.",
			Rango:         "Real Decreto-ley",
			Abreviatura:   "NP",
			Materias:      []string{"procedimiento\nadministrativo", "tributos"},
		},
	}
}

// TestRenderizarNormas fija RenderizarNormas (contrato normas-y-referencias §5;
// data-model §4.2; FR-030, FR-031, FR-034): los bytes exactos de la referencia,
// con la cabecera que nombra data/normas.yaml, las filas por año y número del
// identificador por su valor y los tres casos de escape de una celda —la barra,
// el salto de línea y la abreviatura ausente—; y la misma salida en dos llamadas
// y con las normas en cualquier orden.
func TestRenderizarNormas(t *testing.T) {
	t.Parallel()

	t.Run("bytes-exactos", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, referenciaDeLasNormasDePrueba, string(skills.RenderizarNormas(normasDePrueba())))
	})

	t.Run("dos-llamadas-dan-lo-mismo", func(t *testing.T) {
		t.Parallel()

		normas := normasDePrueba()

		primera := skills.RenderizarNormas(normas)
		segunda := skills.RenderizarNormas(normas)

		assert.Equal(t, primera, segunda)
		assert.Equal(t, normasDePrueba(), normas, "las normas que recibe no se reordenan")
	})

	t.Run("orden-sin-depender-de-la-entrada", func(t *testing.T) {
		t.Parallel()

		normas := normasDePrueba()
		slices.Reverse(normas)

		assert.Equal(t, referenciaDeLasNormasDePrueba, string(skills.RenderizarNormas(normas)))
	})

	t.Run("ceros-a-la-izquierda", func(t *testing.T) {
		t.Parallel()

		// 010565 vale lo mismo que 10565 y va antes que 10566, aunque como número
		// de cifras sería mayor; a igual valor decide el identificador tal cual.
		normas := []skills.Norma{
			{Identificador: "BOE-A-2015-10566", Titulo: "Mayor.", Rango: "Ley", Materias: []string{"a"}},
			{Identificador: "BOE-A-2015-010565", Titulo: "Igual, con un cero.", Rango: "Ley", Materias: []string{"a"}},
			{Identificador: "BOE-A-2015-10565", Titulo: "Igual, sin ceros.", Rango: "Ley", Materias: []string{"a"}},
		}
		esperada := comienzoDeLaReferencia +
			"| Igual, con un cero. |  | `BOE-A-2015-010565` | Ley | a |\n" +
			"| Igual, sin ceros. |  | `BOE-A-2015-10565` | Ley | a |\n" +
			"| Mayor. |  | `BOE-A-2015-10566` | Ley | a |\n"

		assert.Equal(t, esperada, string(skills.RenderizarNormas(normas)))

		slices.Reverse(normas)
		assert.Equal(t, esperada, string(skills.RenderizarNormas(normas)))
	})
}

// comienzoDeLasLeyesVertebrales es lo que abre toda referencia de leyes
// vertebrales, hasta la fila de separación de su tabla (contrato de la skill
// legal-core §2): la cabecera nombra data/normas.yaml, del que sale aunque no se
// llame como ella, y los encabezados son los de la referencia de normas.
const comienzoDeLasLeyesVertebrales = "<!-- generado desde data/normas.yaml, no editar -->\n" +
	"\n" +
	"# Leyes vertebrales\n" +
	"\n" +
	"| Norma | Abreviatura | Identificador | Rango | Materias |\n" +
	"|---|---|---|---|---|\n"

// normasDePruebaConDosVertebrales son normasDePrueba con la marca vertebral en la
// de la barra en el título, BOE-A-2015-10566, y en la LRBRL, BOE-A-1985-5392.
func normasDePruebaConDosVertebrales() []skills.Norma {
	normas := normasDePrueba()
	normas[0].Vertebral = true
	normas[2].Vertebral = true

	return normas
}

// TestRenderizarLeyesVertebrales fija RenderizarLeyesVertebrales (contrato de la
// skill legal-core §2 y §4; research.md D20; FR-065, FR-067): solo las normas
// marcadas vertebral, con la cabecera que nombra data/normas.yaml, su propio
// título y los encabezados, el orden y los escapes de la referencia de normas; sin
// ninguna marcada, la tabla vacía; y las normas que recibe no cambian.
func TestRenderizarLeyesVertebrales(t *testing.T) {
	t.Parallel()

	t.Run("solo-las-marcadas", func(t *testing.T) {
		t.Parallel()

		esperada := comienzoDeLasLeyesVertebrales +
			"| Ley 7/1985, de 2 de abril, reguladora de las Bases del Régimen Local. | LRBRL | `BOE-A-1985-5392` | Ley | régimen local |\n" +
			"| Norma de prueba con una barra \\| en el título. |  | `BOE-A-2015-10566` | Ley | régimen jurídico del sector público |\n"

		assert.Equal(t, esperada, string(skills.RenderizarLeyesVertebrales(normasDePruebaConDosVertebrales())))
	})

	t.Run("todas-marcadas-dan-las-filas-de-las-normas", func(t *testing.T) {
		t.Parallel()

		normas := normasDePrueba()
		for indice := range normas {
			normas[indice].Vertebral = true
		}

		filas := strings.TrimPrefix(referenciaDeLasNormasDePrueba, comienzoDeLaReferencia)
		assert.Equal(t, comienzoDeLasLeyesVertebrales+filas, string(skills.RenderizarLeyesVertebrales(normas)))
	})

	t.Run("ninguna-marcada", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, comienzoDeLasLeyesVertebrales, string(skills.RenderizarLeyesVertebrales(normasDePrueba())))
	})

	t.Run("no-cambia-las-normas-que-recibe", func(t *testing.T) {
		t.Parallel()

		normas := normasDePruebaConDosVertebrales()
		skills.RenderizarLeyesVertebrales(normas)

		assert.Equal(t, normasDePruebaConDosVertebrales(), normas)
	})
}

// referenciaDeLaJerarquiaConEscapes son los bytes exactos de la referencia de
// jerarquiaConEscapes: una fila por nivel en el orden en que llegan, con los tipos
// de norma separados por «, », la barra escrita \| y cada salto de línea como un
// espacio; y detrás la lista de las reglas, cada una con su código y su
// enunciado en una sola línea.
const referenciaDeLaJerarquiaConEscapes = "<!-- generado desde data/jerarquia.yaml, no editar -->\n" +
	"\n" +
	"# Jerarquía normativa\n" +
	"\n" +
	"| Nivel | Boletín | Tipos de norma, de mayor a menor rango |\n" +
	"|---|---|---|\n" +
	"| Unión Europea | Diario de prueba \\| con una barra | Reglamento, Directiva |\n" +
	"| Estado | Boletín de prueba en dos líneas | Ley \\| con una barra |\n" +
	"\n" +
	"## Reglas de interpretación\n" +
	"\n" +
	"- `ley-posterior`: La posterior deroga a la anterior, con saltos de Windows y de retorno de carro.\n" +
	"- `ley-especial`: La especial | prevalece sobre la general.\n"

// jerarquiaConEscapes es la jerarquía de TestRenderizarJerarquia: dos niveles y dos
// reglas, con una barra y saltos de línea de las tres formas en sus textos.
func jerarquiaConEscapes() skills.Jerarquia {
	return skills.Jerarquia{
		Niveles: []skills.NivelNormativo{
			{
				Codigo:  "ue",
				Nombre:  "Unión Europea",
				Boletin: "Diario de prueba | con una barra",
				Normas:  []string{"Reglamento", "Directiva"},
			},
			{Codigo: "estado", Nombre: "Estado", Boletin: "Boletín de prueba\nen dos líneas", Normas: []string{"Ley | con una barra"}},
		},
		Reglas: []skills.ReglaDeInterpretacion{
			{
				Codigo:    "ley-posterior",
				Enunciado: "La posterior deroga a la anterior,\r\ncon saltos de Windows\ry de retorno de carro.",
			},
			{Codigo: "ley-especial", Enunciado: "La especial | prevalece\nsobre la general."},
		},
	}
}

// TestRenderizarJerarquia fija RenderizarJerarquia (contrato de la skill
// legal-core §2 y §4; research.md D20; FR-066, FR-067): los bytes exactos de la
// referencia, con la cabecera que nombra data/jerarquia.yaml, la tabla de niveles
// con su boletín y sus tipos de norma, y la lista de reglas; los niveles y las
// reglas van en el orden en que llegan, que es el que el esquema impone al
// documento; y la jerarquía que recibe no cambia.
func TestRenderizarJerarquia(t *testing.T) {
	t.Parallel()

	t.Run("bytes-exactos", func(t *testing.T) {
		t.Parallel()

		jerarquia := jerarquiaConEscapes()

		assert.Equal(t, referenciaDeLaJerarquiaConEscapes, string(skills.RenderizarJerarquia(jerarquia)))
		assert.Equal(t, jerarquiaConEscapes(), jerarquia, "la jerarquía que recibe no cambia")
	})

	t.Run("en-el-orden-en-que-llegan", func(t *testing.T) {
		t.Parallel()

		jerarquia := jerarquiaConEscapes()
		slices.Reverse(jerarquia.Niveles)
		slices.Reverse(jerarquia.Reglas)

		referencia := string(skills.RenderizarJerarquia(jerarquia))

		assert.Less(t, strings.Index(referencia, "| Estado |"), strings.Index(referencia, "| Unión Europea |"))
		assert.Less(t, strings.Index(referencia, "`ley-especial`"), strings.Index(referencia, "`ley-posterior`"))
	})
}

// datosDelRepositorio es data/, relativo al directorio de este paquete, que es
// donde go test ejecuta sus tests.
const datosDelRepositorio = "../../data"

// leerDatosDelRepositorio es el contenido del YAML de datos del repositorio de
// nombre, leído dentro de data/.
func leerDatosDelRepositorio(t *testing.T, nombre string) []byte {
	t.Helper()

	contenido, err := fs.ReadFile(os.DirFS(datosDelRepositorio), nombre)
	require.NoError(t, err)

	return contenido
}

// referenciaDeLosTiposDePrueba son los bytes exactos de la referencia de
// tiposDePrueba: una fila por tipo en el orden en que llegan y, en el genérico,
// una por unidad en el orden de plazos.Unidades; la duración con el nombre de su
// unidad en singular o en plural; los preceptos separados por «; », con «art.»
// delante del apartado de un artículo; y la barra escrita \| y cada salto de
// línea como un espacio.
const referenciaDeLosTiposDePrueba = "<!-- generado desde data/plazos.yaml, no editar -->\n" +
	"\n" +
	"# Recursos y plazos\n" +
	"\n" +
	"| Tipo | Código | Duración | Cómputo | Precepto |\n" +
	"|---|---|---|---|---|\n" +
	"| recurso de alzada | `alzada` | 1 mes | `lpac` | art. 122.1 (BOE-A-2015-10565, bloque a122) |\n" +
	"| recurso \\| con una barra en dos líneas | `contencioso` | 2 meses | `contencioso` | " +
	"art. 46.1 (BOE-A-1998-16718, bloque a46); art. 128.2 (BOE-A-1998-16718, bloque a128) |\n" +
	"| de un día hábil | `un-dia-habil` | 1 día hábil | `lpac` | art. 30.2 (BOE-A-2015-10565, bloque a30) |\n" +
	"| de quince días hábiles | `quince-dias-habiles` | 15 días hábiles | `lpac` | art. 30.2 (BOE-A-2015-10565, bloque a30) |\n" +
	"| de un día natural | `un-dia-natural` | 1 día natural | `lpac` | art. 30.2 (BOE-A-2015-10565, bloque a30) |\n" +
	"| de veinte días naturales | `veinte-dias-naturales` | 20 días naturales | `lpac` | " +
	"DA 3ª (BOE-A-2015-10565, bloque datercera) |\n" +
	"| día en que se produce el silencio administrativo | `silencio` | la que recibe, con su unidad | `lpac` | " +
	"art. 21.3 (BOE-A-2015-10565, bloque a21); art. 24.4 (BOE-A-2015-10565, bloque a24) |\n" +
	"| plazo genérico | `generico` | la que recibe, en `meses` | `lpac` | art. 30.4 (BOE-A-2015-10565, bloque a30) |\n" +
	"| plazo genérico | `generico` | la que recibe, en `dias-habiles` | `lpac` | art. 30.2 (BOE-A-2015-10565, bloque a30) |\n" +
	"| plazo genérico | `generico` | la que recibe, en `dias-naturales` | `lpac` | art. 30.2 (BOE-A-2015-10565, bloque a30) |\n"

// tiposDePrueba son los tipos de TestRenderizarRecursosYPlazos: los de
// data/plazos.yaml que fijan su duración en meses, uno con una barra y un salto de
// línea en el nombre y dos preceptos; cuatro más con cada unidad en días en
// singular y en plural, uno con el apartado de una disposición; el silencio, sin
// duración; y el genérico con sus tres unidades.
func tiposDePrueba() []plazos.Tipo {
	articulo302 := plazos.Precepto{Norma: "BOE-A-2015-10565", Bloque: "a30", Apartado: "30.2", Fija: "computo"}
	fijo := func(codigo, nombre string, cantidad int, unidad plazos.Unidad, precepto plazos.Precepto) plazos.Tipo {
		return plazos.Tipo{
			Codigo:    codigo,
			Nombre:    nombre,
			Computo:   plazos.ComputoLPAC,
			Duracion:  &plazos.Duracion{Cantidad: cantidad, Unidad: unidad},
			Preceptos: []plazos.Precepto{precepto},
		}
	}

	return []plazos.Tipo{
		fijo("alzada", "recurso de alzada", 1, plazos.UnidadMeses,
			plazos.Precepto{Norma: "BOE-A-2015-10565", Bloque: "a122", Apartado: "122.1", Fija: plazos.FijaDuracion}),
		{
			Codigo:   "contencioso",
			Nombre:   "recurso | con una barra\nen dos líneas",
			Computo:  plazos.ComputoContencioso,
			Duracion: &plazos.Duracion{Cantidad: 2, Unidad: plazos.UnidadMeses},
			Preceptos: []plazos.Precepto{
				{Norma: "BOE-A-1998-16718", Bloque: "a46", Apartado: "46.1", Fija: plazos.FijaDuracion},
				{Norma: "BOE-A-1998-16718", Bloque: "a128", Apartado: "128.2", Fija: "agosto"},
			},
		},
		fijo("un-dia-habil", "de un día hábil", 1, plazos.UnidadDiasHabiles, articulo302),
		fijo("quince-dias-habiles", "de quince días hábiles", 15, plazos.UnidadDiasHabiles, articulo302),
		fijo("un-dia-natural", "de un día natural", 1, plazos.UnidadDiasNaturales, articulo302),
		fijo("veinte-dias-naturales", "de veinte días naturales", 20, plazos.UnidadDiasNaturales,
			plazos.Precepto{Norma: "BOE-A-2015-10565", Bloque: "datercera", Apartado: "DA 3ª", Fija: plazos.FijaDuracion}),
		{
			Codigo:  "silencio",
			Nombre:  "día en que se produce el silencio administrativo",
			Computo: plazos.ComputoLPAC,
			Preceptos: []plazos.Precepto{
				{Norma: "BOE-A-2015-10565", Bloque: "a21", Apartado: "21.3", Fija: "inicio"},
				{Norma: "BOE-A-2015-10565", Bloque: "a24", Apartado: "24.4", Fija: "silencio"},
			},
		},
		{
			Codigo:  plazos.TipoGenerico,
			Nombre:  "plazo genérico",
			Computo: plazos.ComputoLPAC,
			Unidades: map[plazos.Unidad]plazos.Precepto{
				plazos.UnidadDiasNaturales: articulo302,
				plazos.UnidadDiasHabiles:   articulo302,
				plazos.UnidadMeses:         {Norma: "BOE-A-2015-10565", Bloque: "a30", Apartado: "30.4", Fija: "computo"},
			},
		},
	}
}

// TestRenderizarRecursosYPlazos fija RenderizarRecursosYPlazos (contrato de la
// skill legal-core §4 de H9; FR-032): los bytes exactos de la referencia, con la
// cabecera que nombra data/plazos.yaml; cada unidad de plazos.Unidades con su
// nombre; los tipos que recibe no cambian; y references/recursos_y_plazos.md de
// una skill que la declara sale de data/plazos.yaml con plazos.LeerTipos —una
// fila por tipo y por unidad del genérico—, o el defecto del fichero que no se
// lee.
func TestRenderizarRecursosYPlazos(t *testing.T) {
	t.Parallel()

	t.Run("bytes-exactos", func(t *testing.T) {
		t.Parallel()

		tipos := tiposDePrueba()

		assert.Equal(t, referenciaDeLosTiposDePrueba, string(skills.RenderizarRecursosYPlazos(tipos)))
		assert.Equal(t, tiposDePrueba(), tipos, "los tipos que recibe no cambian")
	})

	t.Run("cada-unidad-con-su-nombre", func(t *testing.T) {
		t.Parallel()

		// Sin nombre, la unidad iría tal cual en singular y en plural: el singular
		// sería su código y los dos nombres, el mismo.
		for _, unidad := range plazos.Unidades() {
			singular := strings.TrimPrefix(celdaDeLaDuracion(t, 1, unidad), "1 ")
			plural := strings.TrimPrefix(celdaDeLaDuracion(t, 2, unidad), "2 ")

			assert.NotEqual(t, string(unidad), singular, "la unidad %q se nombra con palabras", unidad)
			assert.NotEqual(t, singular, plural, "la unidad %q tiene singular y plural", unidad)
		}
	})

	t.Run("unidad-sin-nombre-tal-cual", func(t *testing.T) {
		t.Parallel()

		tipo := tiposDePrueba()[0]
		tipo.Duracion = &plazos.Duracion{Cantidad: 3, Unidad: "semanas"}

		assert.Contains(t, string(skills.RenderizarRecursosYPlazos([]plazos.Tipo{tipo})), "| 3 semanas |")
	})

	t.Run("desde-data", func(t *testing.T) {
		t.Parallel()

		contenido := leerDatosDelRepositorio(t, "plazos.yaml")

		tipos, err := plazos.LeerTipos(contenido)
		require.NoError(t, err)

		referencia := referenciaRegenerada(t, "plazos.yaml", string(contenido), "recursos_y_plazos")

		assert.Equal(t, string(skills.RenderizarRecursosYPlazos(tipos)), referencia)
		assert.Equal(t, 6+9, strings.Count(referencia, "\n"),
			"la cabecera, el título y la tabla, con nueve filas: seis tipos y las tres unidades del genérico")
	})

	t.Run("defecto-de-los-datos", func(t *testing.T) {
		t.Parallel()

		defectos := defectosDeLaRegeneracion(t, "plazos.yaml", "tipos: []\n", "recursos_y_plazos")

		require.NotEmpty(t, defectos)
		assert.True(t, strings.HasPrefix(defectos[0], "gamma: data/plazos.yaml: "), "el defecto nombra el fichero: %q", defectos[0])
	})
}

// celdaDeLaDuracion es la celda de la duración de la única fila de la referencia
// de un tipo que fija la duración dada.
func celdaDeLaDuracion(t *testing.T, cantidad int, unidad plazos.Unidad) string {
	t.Helper()

	tipo := tiposDePrueba()[0]
	tipo.Duracion = &plazos.Duracion{Cantidad: cantidad, Unidad: unidad}

	lineas := strings.Split(strings.TrimSuffix(string(skills.RenderizarRecursosYPlazos([]plazos.Tipo{tipo})), "\n"), "\n")
	celdas := strings.Split(lineas[len(lineas)-1], " | ")
	require.Len(t, celdas, 5)

	return celdas[2]
}

// coberturaDePruebaParaLaReferencia es la cobertura de TestRenderizarCobertura:
// una entrada con norma y otra sin ella, las dos con el apartado de una
// disposición, y una tercera con el de un artículo, con una barra y un salto de
// línea en sus textos.
func coberturaDePruebaParaLaReferencia() skills.Cobertura {
	return skills.Cobertura{ReguladosAparte: []skills.ReguladoAparte{
		{
			Procedimiento: "aplicación de los tributos",
			Precepto:      skills.PreceptoDeLaCobertura{Norma: "BOE-A-2015-10565", Bloque: "daprimera", Apartado: "DA 1ª.2.a"},
			Norma: &skills.NormaQueLoRige{
				Identificador: "BOE-A-2003-23186",
				Nombre:        "Ley 58/2003, General | Tributaria",
				Aplicable:     skills.PreceptoAplicable{Bloque: "a97", Apartado: "97"},
			},
		},
		{
			Procedimiento: "procedimiento con una barra |\r\ny un salto de línea",
			Precepto:      skills.PreceptoDeLaCobertura{Norma: "BOE-A-2015-10565", Bloque: "daprimera", Apartado: "DA 1ª.2.b"},
		},
		{
			Procedimiento: "reclamaciones económico-administrativas",
			Precepto:      skills.PreceptoDeLaCobertura{Norma: "BOE-A-2015-10565", Bloque: "a112", Apartado: "112.4"},
		},
	}}
}

// referenciaDeLaCoberturaDePrueba son los bytes exactos de la referencia de
// coberturaDePruebaParaLaReferencia: una fila por entrada en el orden en que
// llegan, «art.» solo delante del apartado de un artículo, la norma que lo rige
// con su nombre y su identificador o «—», y la barra escrita \| y cada salto de
// línea como un espacio.
const referenciaDeLaCoberturaDePrueba = "<!-- generado desde data/cobertura.yaml, no editar -->\n" +
	"\n" +
	"# Cobertura por materia\n" +
	"\n" +
	"| Procedimiento | Lo regula aparte | Norma que lo rige |\n" +
	"|---|---|---|\n" +
	"| aplicación de los tributos | DA 1ª.2.a (BOE-A-2015-10565, bloque daprimera) | " +
	"Ley 58/2003, General \\| Tributaria (BOE-A-2003-23186) |\n" +
	"| procedimiento con una barra \\| y un salto de línea | DA 1ª.2.b (BOE-A-2015-10565, bloque daprimera) | — |\n" +
	"| reclamaciones económico-administrativas | art. 112.4 (BOE-A-2015-10565, bloque a112) | — |\n"

// TestRenderizarCobertura fija RenderizarCobertura (contrato de la skill
// legal-core §4 de H9; FR-031, FR-032): los bytes exactos de la referencia, con
// la cabecera que nombra data/cobertura.yaml; la cobertura que recibe no cambia; y
// references/cobertura_por_materia.md de una skill que la declara sale de
// data/cobertura.yaml con skills.LeerCobertura —una fila por entrada—, o los
// defectos del fichero que no valida.
func TestRenderizarCobertura(t *testing.T) {
	t.Parallel()

	t.Run("bytes-exactos", func(t *testing.T) {
		t.Parallel()

		cobertura := coberturaDePruebaParaLaReferencia()

		assert.Equal(t, referenciaDeLaCoberturaDePrueba, string(skills.RenderizarCobertura(cobertura)))
		assert.Equal(t, coberturaDePruebaParaLaReferencia(), cobertura, "la cobertura que recibe no cambia")
	})

	t.Run("desde-data", func(t *testing.T) {
		t.Parallel()

		contenido := leerDatosDelRepositorio(t, "cobertura.yaml")

		cobertura, err := skills.LeerCobertura(contenido)
		require.NoError(t, err)

		referencia := referenciaRegenerada(t, "cobertura.yaml", string(contenido), "cobertura_por_materia")

		assert.Equal(t, string(skills.RenderizarCobertura(cobertura)), referencia)
		assert.Equal(t, 6+len(cobertura.ReguladosAparte), strings.Count(referencia, "\n"),
			"la cabecera, el título y la tabla, con una fila por entrada")
	})

	t.Run("defecto-de-los-datos", func(t *testing.T) {
		t.Parallel()

		defectos := defectosDeLaRegeneracion(t, "cobertura.yaml", "regulados_aparte: []\n", "cobertura_por_materia")

		require.NotEmpty(t, defectos)
		assert.True(t, strings.HasPrefix(defectos[0], "gamma: data/cobertura.yaml: "), "el defecto nombra el fichero: %q", defectos[0])
	})
}

// regenerarGamma es lo que skills.Regenerar construye para un repositorio
// temporal con data/<datos> y la skill gamma, que declara la referencia dada y
// la enlaza (FR-007 de H31).
func regenerarGamma(t *testing.T, datos, contenido, referencia string) skills.Regenerado {
	t.Helper()

	raiz := t.TempDir()
	escribirFicheroDePrueba(t, filepath.Join(raiz, "data", datos), contenido)
	escribirFicheroDePrueba(t, rutaDeSkill(raiz, "gamma", "SKILL.md"), "---\n"+
		"name: gamma\n"+
		"description: Skill de prueba con una referencia.\n"+
		"metadata:\n"+
		"  kitlegal-referencias: "+referencia+"\n"+
		"---\n"+
		"# Gamma\n"+
		"\n"+
		"- ["+referencia+"](references/"+referencia+".md)\n")

	regenerado, err := skills.Regenerar(raiz, nil)
	require.NoError(t, err)

	return regenerado
}

// referenciaRegenerada es el contenido de references/<referencia>.md de gamma,
// regenerada sin defectos desde data/<datos>.
func referenciaRegenerada(t *testing.T, datos, contenido, referencia string) string {
	t.Helper()

	regenerado := regenerarGamma(t, datos, contenido, referencia)
	require.Empty(t, presentarDefectos(regenerado.Defectos()))
	require.Len(t, regenerado.Skills, 1)
	require.Len(t, regenerado.Skills[0].Referencias, 1)
	require.Equal(t, referencia+".md", regenerado.Skills[0].Referencias[0].Fichero)

	return string(regenerado.Skills[0].Referencias[0].Contenido)
}

// defectosDeLaRegeneracion son los defectos de gamma regenerada desde
// data/<datos>, que no se lee.
func defectosDeLaRegeneracion(t *testing.T, datos, contenido, referencia string) []string {
	t.Helper()

	return presentarDefectos(regenerarGamma(t, datos, contenido, referencia).Defectos())
}

// cabeceraDeUnaGenerada es la sonda independiente de DatosDeLaReferencia: la
// primera línea de una referencia generada, que nombra el YAML de datos de
// data/ del que sale (contrato normas-y-referencias §5).
var cabeceraDeUnaGenerada = regexp.MustCompile(`\A<!-- generado desde (data/[^,\n]+\.yaml), no editar -->(\n|\z)`)

// TestDatosDeLaReferencia fija DatosDeLaReferencia (research D11 y data-model §3
// de H30; FR-001(c)): cada referencia generada de skills/*/references/ da el
// YAML de data/ que nombra su cabecera, que existe, y las demás no dan ninguno;
// y un texto sin la cabecera en su primera línea, o con una cabecera que no
// nombra ningún fichero de data/, no da ninguno.
func TestDatosDeLaReferencia(t *testing.T) {
	t.Parallel()

	t.Run("las-referencias-del-repositorio", comprobarDatosDeLasReferenciasDelRepositorio)
	t.Run("textos", comprobarDatosDeTextos)
}

// comprobarDatosDeLasReferenciasDelRepositorio comprueba cada referencia de
// skills/*/references/ contra la sonda de su cabecera.
func comprobarDatosDeLasReferenciasDelRepositorio(t *testing.T) {
	t.Parallel()

	raiz := os.DirFS(skillsDelRepositorio)

	rutas, err := fs.Glob(raiz, "*/references/*")
	require.NoError(t, err)
	require.NotEmpty(t, rutas)

	for _, ruta := range rutas {
		t.Run(ruta, func(t *testing.T) {
			t.Parallel()

			contenido, err := fs.ReadFile(raiz, ruta)
			require.NoError(t, err)

			datos, generada := skills.DatosDeLaReferencia(contenido)

			cabecera := cabeceraDeUnaGenerada.FindSubmatch(contenido)
			if cabecera == nil {
				assert.False(t, generada)
				assert.Empty(t, datos)

				return
			}

			require.True(t, generada)
			assert.Equal(t, string(cabecera[1]), datos)

			fichero, err := os.Stat(filepath.Join("..", "..", filepath.FromSlash(datos)))
			require.NoError(t, err)
			assert.True(t, fichero.Mode().IsRegular())
		})
	}
}

// comprobarDatosDeTextos comprueba textos con la cabecera, sin ella, con ella
// fuera de la primera línea y con una cabecera que no nombra ningún fichero de
// data/.
func comprobarDatosDeTextos(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre, contenido, datos string
		generada                 bool
	}{
		{
			nombre:    "con-cabecera",
			contenido: "<!-- generado desde data/normas.yaml, no editar -->\n\n# Normas de referencia\n",
			datos:     "data/normas.yaml",
			generada:  true,
		},
		{
			nombre:    "solo-la-cabecera-sin-salto-final",
			contenido: "<!-- generado desde data/plazos.yaml, no editar -->",
			datos:     "data/plazos.yaml",
			generada:  true,
		},
		{
			nombre:    "sin-cabecera",
			contenido: "# Normas de referencia\n\n| Norma |\n",
		},
		{
			nombre:    "vacio",
			contenido: "",
		},
		{
			nombre:    "la-cabecera-en-la-segunda-linea",
			contenido: "\n<!-- generado desde data/normas.yaml, no editar -->\n",
		},
		{
			nombre:    "la-cabecera-con-retorno-de-carro",
			contenido: "<!-- generado desde data/normas.yaml, no editar -->\r\n",
		},
		{
			nombre:    "la-cabecera-fuera-de-data",
			contenido: "<!-- generado desde normas.yaml, no editar -->\n",
		},
		{
			nombre:    "la-cabecera-sin-fichero",
			contenido: "<!-- generado desde data/, no editar -->\n",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			datos, generada := skills.DatosDeLaReferencia([]byte(caso.contenido))
			assert.Equal(t, caso.generada, generada)
			assert.Equal(t, caso.datos, datos)
		})
	}
}

// Los dos contenidos de las entradas de references/ de TestReferenciasAMano: uno
// que abre con la cabecera de una referencia generada y otro que no.
const (
	contenidoConCabecera = "<!-- generado desde data/normas.yaml, no editar -->\n\n# Normas de referencia\n"
	contenidoSinCabecera = "# Cómo se cita\n\nEl protocolo de detalle, escrito a mano.\n"
)

// TestReferenciasAMano fija la clasificación de FR-012 (contracts/
// gates-de-skills-check.md §3; research D13; data-model §3) sobre la raíz de un
// repositorio en un fstest.MapFS, con las skills de la sincronía: alfa declara
// normas, leyes_vertebrales y jerarquia_normativa, y beta no declara ninguna. De
// las cinco clases de entradas de references/ —lo que no es un fichero regular
// .md (un subdirectorio, un enlace, otra extensión), la referencia declarada, con
// o sin su cabecera, la declarada y ausente, la no declarada que abre con la
// cabecera y la no declarada que no abre con ella—, solo la última es una
// referencia a mano, también con el nombre de una fila de la tabla de generadores
// en una skill que no la declara; van con sus bytes y en orden de nombre. Y la
// primera, la entrada que no es un fichero regular .md, declarada o no, es la
// que da el defecto de FR-006 de H31 (entradasQueNoSonFicherosMd), por su
// nombre y en orden de nombre.
func TestReferenciasAMano(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string

		// skill es la del árbol cuyas referencias a mano se piden.
		skill string

		// entradas son las de references/ de la skill, por su ruta dentro de él.
		entradas fstest.MapFS

		// aMano son los nombres de las referencias a mano, en su orden.
		aMano []string

		// noMd son los nombres de las entradas que no son un fichero regular .md,
		// en su orden.
		noMd []string
	}{
		{
			nombre:   "subdirectorio",
			skill:    "alfa",
			entradas: fstest.MapFS{"anexos.md/nota.md": sinCabecera()},
			noMd:     []string{"anexos.md"},
		},
		{
			nombre: "enlace",
			skill:  "alfa",
			entradas: fstest.MapFS{
				"enlace.md":       {Data: []byte("como_se_cita.md"), Mode: fs.ModeSymlink},
				"como_se_cita.md": sinCabecera(),
			},
			aMano: []string{"como_se_cita.md"},
			noMd:  []string{"enlace.md"},
		},
		{
			nombre:   "otra-extension",
			skill:    "alfa",
			entradas: fstest.MapFS{"notas.txt": sinCabecera()},
			noMd:     []string{"notas.txt"},
		},
		{
			nombre:   "sin-extension",
			skill:    "beta",
			entradas: fstest.MapFS{"LEEME": sinCabecera()},
			noMd:     []string{"LEEME"},
		},
		{
			nombre:   "declarada-que-es-un-enlace",
			skill:    "alfa",
			entradas: fstest.MapFS{"normas.md": {Data: []byte("../../../data/normas.yaml"), Mode: fs.ModeSymlink}},
			noMd:     []string{"normas.md"},
		},
		{nombre: "declarada-con-cabecera", skill: "alfa", entradas: fstest.MapFS{"normas.md": conCabecera()}},
		{nombre: "declarada-sin-cabecera", skill: "alfa", entradas: fstest.MapFS{"normas.md": sinCabecera()}},
		{nombre: "declaradas-y-ausentes", skill: "alfa", entradas: fstest.MapFS{}},
		{nombre: "no-declarada-con-cabecera", skill: "alfa", entradas: fstest.MapFS{"antigua.md": conCabecera()}},
		{
			nombre:   "no-declarada-sin-cabecera",
			skill:    "alfa",
			entradas: fstest.MapFS{"como_se_cita.md": sinCabecera()},
			aMano:    []string{"como_se_cita.md"},
		},
		{
			nombre:   "con-el-nombre-de-un-generador-sin-declararla",
			skill:    "beta",
			entradas: fstest.MapFS{"normas.md": sinCabecera(), "antigua.md": conCabecera()},
			aMano:    []string{"normas.md"},
		},
		{
			nombre: "todas-las-clases-juntas",
			skill:  "alfa",
			entradas: fstest.MapFS{
				"anexos.md/nota.md":      sinCabecera(),
				"enlace.md":              {Data: []byte("ambito.md"), Mode: fs.ModeSymlink},
				"notas.txt":              sinCabecera(),
				"normas.md":              sinCabecera(),
				"jerarquia_normativa.md": conCabecera(),
				"antigua.md":             conCabecera(),
				"como_se_cita.md":        sinCabecera(),
				"ambito.md":              {Data: []byte("# Ámbito\n\nOtra referencia escrita a mano.\n")},
			},
			aMano: []string{"ambito.md", "como_se_cita.md"},
			noMd:  []string{"anexos.md", "enlace.md", "notas.txt"},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			arbol := arbolDeLasSkills()
			references := "skills/" + caso.skill + "/references"
			arbol[references] = &fstest.MapFile{Mode: fs.ModeDir | 0o750}

			for ruta, entrada := range caso.entradas {
				arbol[references+"/"+ruta] = entrada
			}

			var esperadas []skills.Referencia
			for _, nombre := range caso.aMano {
				esperadas = append(esperadas, skills.Referencia{Fichero: nombre, Contenido: caso.entradas[nombre].Data})
			}

			aMano, err := skills.ReferenciasAMano(arbol, caso.skill)
			require.NoError(t, err)
			assert.Equal(t, esperadas, aMano)

			noMd, err := skills.EntradasQueNoSonFicherosMd(arbol, caso.skill)
			require.NoError(t, err)
			assert.Equal(t, caso.noMd, noMd)
		})
	}

	t.Run("sin-references", func(t *testing.T) {
		t.Parallel()

		aMano, err := skills.ReferenciasAMano(arbolDeLasSkills(), "alfa")
		require.NoError(t, err)
		assert.Empty(t, aMano)

		noMd, err := skills.EntradasQueNoSonFicherosMd(arbolDeLasSkills(), "alfa")
		require.NoError(t, err)
		assert.Empty(t, noMd)
	})

	t.Run("entradas-de-references-que-no-es-un-directorio", func(t *testing.T) {
		t.Parallel()

		arbol := arbolDeLasSkills()
		arbol["skills/alfa/references"] = sinCabecera()

		noMd, err := skills.EntradasQueNoSonFicherosMd(arbol, "alfa")
		require.Error(t, err)
		assert.Nil(t, noMd)
		assert.True(t, strings.HasPrefix(err.Error(), "skills/alfa/references no se puede listar: "), "error: %v", err)
	})

	probarReferenciasAManoSinPoderLeer(t)
}

// probarReferenciasAManoSinPoderLeer fija los errores de ReferenciasAMano, cada
// uno con lo que no se puede leer: references/ que no es un directorio, la skill
// sin SKILL.md o con un frontmatter que no se lee, y una referencia que no se
// puede abrir.
func probarReferenciasAManoSinPoderLeer(t *testing.T) {
	t.Helper()

	casos := []struct {
		nombre string

		// alterar deja el árbol, que tiene ya references/como_se_cita.md en alfa,
		// como lo pide el caso, y devuelve el que se lee.
		alterar func(arbol fstest.MapFS) fs.FS

		prefijo string
		causa   error
	}{
		{
			nombre: "references-que-no-es-un-directorio",
			alterar: func(arbol fstest.MapFS) fs.FS {
				delete(arbol, "skills/alfa/references/como_se_cita.md")
				arbol["skills/alfa/references"] = sinCabecera()

				return arbol
			},
			prefijo: "skills/alfa/references no se puede listar: ",
		},
		{
			nombre: "sin-skill-md",
			alterar: func(arbol fstest.MapFS) fs.FS {
				delete(arbol, "skills/alfa/SKILL.md")

				return arbol
			},
			prefijo: "skills/alfa/SKILL.md no se puede leer: ",
			causa:   fs.ErrNotExist,
		},
		{
			nombre: "frontmatter-que-no-se-lee",
			alterar: func(arbol fstest.MapFS) fs.FS {
				arbol["skills/alfa/SKILL.md"] = &fstest.MapFile{Data: []byte("# Alfa sin frontmatter\n")}

				return arbol
			},
			prefijo: "skills/alfa/SKILL.md: ",
		},
		{
			nombre: "referencia-que-no-se-puede-abrir",
			alterar: func(arbol fstest.MapFS) fs.FS {
				return arbolQueNoAbre{FS: arbol, ruta: "skills/alfa/references/como_se_cita.md"}
			},
			prefijo: "skills/alfa/references/como_se_cita.md no se puede leer: ",
			causa:   fs.ErrPermission,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			arbol := arbolDeLasSkills()
			arbol["skills/alfa/references/como_se_cita.md"] = sinCabecera()

			aMano, err := skills.ReferenciasAMano(caso.alterar(arbol), "alfa")
			require.Error(t, err)
			assert.Nil(t, aMano)
			assert.True(t, strings.HasPrefix(err.Error(), caso.prefijo), "error: %v", err)

			if caso.causa != nil {
				require.ErrorIs(t, err, caso.causa)
			}
		})
	}
}

// arbolDeLasSkills es la raíz de un repositorio con el SKILL.md de las dos skills
// de la sincronía, sin ninguna references/, en un árbol nuevo en cada llamada.
func arbolDeLasSkills() fstest.MapFS {
	return fstest.MapFS{
		"skills/alfa/SKILL.md": {Data: []byte(skillMdDeAlfaSinSincronizar)},
		"skills/beta/SKILL.md": {Data: []byte(skillMdDeBetaSinNadaQueGenerar)},
	}
}

// conCabecera y sinCabecera son un fichero regular nuevo con cada uno de los dos
// contenidos.
func conCabecera() *fstest.MapFile { return &fstest.MapFile{Data: []byte(contenidoConCabecera)} }

func sinCabecera() *fstest.MapFile { return &fstest.MapFile{Data: []byte(contenidoSinCabecera)} }

// arbolQueNoAbre es un árbol que no abre la ruta dada. Solo tiene Open, así que
// fs.ReadDir y fs.ReadFile pasan por él.
type arbolQueNoAbre struct {
	fs.FS

	ruta string
}

// Open da un error de permiso para la ruta dada y abre las demás.
func (a arbolQueNoAbre) Open(nombre string) (fs.File, error) {
	if nombre == a.ruta {
		return nil, &fs.PathError{Op: "open", Path: nombre, Err: fs.ErrPermission}
	}

	return a.FS.Open(nombre)
}
