package app

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/plazos"
)

// Lo que la copia de data/plazos.yaml de TestDescribeDePlazosSinDatos añade a
// lpac: un tipo y un periodo de aviso con su código, su etiqueta y su texto
// (contracts/applet-calcular.md §3; research D7 de H36). Ninguno de los dos
// códigos es de data/plazos.yaml, así que una ayuda o un enumerado de calcular
// que salieran de los datos los nombrarían.
const (
	// codigoDelTipoAnadido es el código del tipo de más.
	codigoDelTipoAnadido = "ensayo-del-esquema"

	// tipoAnadido es el tipo de más, de lpac, con la sangría de los de
	// data/plazos.yaml: entra detrás del último, delante de la lista de los
	// cómputos.
	tipoAnadido = "  - codigo: " + codigoDelTipoAnadido + `
    nombre: plazo de ensayo del esquema de calcular
    computo: lpac
    duracion: {cantidad: 10, unidad: dias-habiles}
    da: vencimiento
    preceptos:
      - {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.2", fija: duracion, vigente_desde: "2016-10-02"}
`

	// periodoAnadido es el periodo de aviso de más, el único de lpac, que
	// nombra uno de los preceptos del cómputo: entra detrás de sus unidades.
	periodoAnadido = `    periodos:
      - desde: "03-01"
        hasta: "03-02"
        efecto: aviso
        precepto: {norma: BOE-A-2015-10565, apartado: "30.2"}
        aviso:
          codigo: ensayo-del-esquema-aviso
          etiqueta: ENSAYO DEL ESQUEMA
          texto: "el plazo pasa por los días de un periodo de ensayo del esquema de calcular."
`
)

// juegoDelDescribeDePlazos son unas fuentes del applet plazos con las que
// --describe de calcular tiene que dar su golden.
type juegoDelDescribeDePlazos struct {
	nombre string
	// fuentes las compone y exige la premisa del juego: lo que hace de él lo
	// que dice su nombre.
	fuentes func(t *testing.T) FicherosDePlazos
}

// TestDescribeDePlazosSinDatos fija que el esquema de calcular no depende de
// data/plazos.yaml (H36 FR-012; US4.1; SC-005; contracts/applet-calcular.md
// §3): con los ficheros del binario, con una copia de data/plazos.yaml que
// lleva un tipo y un aviso de periodo más, y sin los tipos, que no cargan,
// --describe de calcular termina en 0 y da, byte a byte, el golden
// describe.json, que este test solo lee.
func TestDescribeDePlazosSinDatos(t *testing.T) {
	t.Parallel()

	golden, err := os.ReadFile(filepath.Clean(filepath.Join(carpetaDeLosGoldenDePlazos, goldenDelDescribeDePlazos)))
	require.NoError(t, err, "el golden lo escribe --describe, con -actualizar-golden-de-plazos")

	for _, juego := range juegosDelDescribeDePlazos() {
		t.Run(juego.nombre, func(t *testing.T) {
			t.Parallel()

			res := invocar(t, registroDePlazos(t, juego.fuentes(t)),
				"kitlegal", "plazos", "calcular", "x", "x", "x", "--describe")

			require.Equal(t, 0, res.codigo, res.errores)
			assert.Equal(t, string(golden), res.salida, "--describe de calcular, byte a byte el golden")
		})
	}
}

// juegosDelDescribeDePlazos son los tres juegos de fuentes de
// contracts/applet-calcular.md §3.
func juegosDelDescribeDePlazos() []juegoDelDescribeDePlazos {
	return []juegoDelDescribeDePlazos{
		{nombre: "los-del-binario", fuentes: fuentesDePlazos},
		{nombre: "con-un-tipo-y-un-aviso-mas", fuentes: fuentesConUnTipoYUnAvisoMas},
		{nombre: "sin-los-tipos", fuentes: fuentesSinLosTipos},
	}
}

// fuentesConUnTipoYUnAvisoMas son las del binario con data/plazos.yaml más el
// tipo y el periodo de aviso de lpac de arriba, añadidos por sustitución de
// texto (research D7 de H36). Exige que LeerTipos lea la copia sin defecto —las
// reglas de §C.2 rechazan un periodo de aviso mal formado— y que dé un tipo más
// que el fichero del binario, el añadido.
func fuentesConUnTipoYUnAvisoMas(t *testing.T) FicherosDePlazos {
	t.Helper()

	fuentes := fuentesDePlazos(t)

	reales, err := plazos.LeerTipos(fuentes.Tipos)
	require.NoError(t, err, "data/plazos.yaml carga")

	copia := sustituida(t, string(fuentes.Tipos), "\ncomputos:\n", "\n"+tipoAnadido+"computos:\n")
	copia = sustituida(t, copia, "    unidades: [meses, dias-habiles, dias-naturales]\n",
		"    unidades: [meses, dias-habiles, dias-naturales]\n"+periodoAnadido)

	ampliados, err := plazos.LeerTipos([]byte(copia))
	require.NoError(t, err, "la copia con un tipo y un aviso de periodo más carga")
	require.Len(t, ampliados, len(reales)+1, "la copia da un tipo más que data/plazos.yaml")
	require.True(t, slices.ContainsFunc(ampliados, func(tipo plazos.Tipo) bool {
		return tipo.Codigo == codigoDelTipoAnadido
	}), "el tipo de más es el añadido")

	fuentes.Tipos = []byte(copia)

	return fuentes
}

// fuentesSinLosTipos son las del binario con Tipos vacío, que no carga.
func fuentesSinLosTipos(t *testing.T) FicherosDePlazos {
	t.Helper()

	fuentes := fuentesDePlazos(t)
	fuentes.Tipos = nil

	_, err := plazos.LeerTipos(fuentes.Tipos)
	require.Error(t, err, "sin los tipos no hay tipos que leer")

	return fuentes
}

// sustituida es el texto con viejo, que tiene que aparecer en él exactamente una
// vez, sustituido por nuevo: una copia que no cambiara nada, o que cambiara algo
// más, no sería la que dice el test.
func sustituida(t *testing.T, texto, viejo, nuevo string) string {
	t.Helper()

	require.Equal(t, 1, strings.Count(texto, viejo), "%q aparece una sola vez en el texto que se cambia", viejo)

	return strings.Replace(texto, viejo, nuevo, 1)
}
