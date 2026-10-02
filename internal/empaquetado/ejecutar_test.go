package empaquetado_test

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal"
	"github.com/jmorenobl/kitlegal/internal/app"
	"github.com/jmorenobl/kitlegal/internal/empaquetado"
)

// lineaDeUso es la línea con la que el paso responde a una invocación que no
// tiene la forma de ninguna de sus dos órdenes (contracts/paso.md §1).
const lineaDeUso = "empaquetar: uso: empaquetar piezas -version … -macos … -windows … -icono … -salida … | " +
	"empaquetar catalogo -version … -sha256 … -salida …\n"

// prefijoDeLaLinea es con lo que empieza la línea de todo fallo del paso.
const prefijoDeLaLinea = "empaquetar: "

// herramientasDeHoy son las diez herramientas que el servidor MCP anuncia con
// el registro de producción de hoy, escritas a mano (FR-014; SC-004). El test
// pide que estén, no que sean las únicas: una herramienta nueva del registro
// llega a `tools` sin tocar el paso ni este test (US5).
var herramientasDeHoy = []string{
	"boe_buscar", "boe_indice", "boe_articulo", "boe_articulos", "boe_metadatos", "boe_analisis",
	"graph_show", "graph_stats", "graph_check",
	"territorio_resolver",
}

// skillsDeHoy son los SKILL.md de las dos skills empotradas hoy (FR-020), con
// su ruta en el plugin. Como herramientasDeHoy, tienen que estar.
var skillsDeHoy = []string{"skills/boe-legislacion/SKILL.md", "skills/legal-core/SKILL.md"}

// TestEjecutar es el control en `make ci` de la línea de órdenes del paso
// (contracts/paso.md §1 y §7): `piezas`, con la composición de producción y
// binarios de prueba, termina con 0 y sin nada en la salida de error, con
// `tools` igual a lo que app.HerramientasAnunciadas da del registro de
// producción y con el árbol `skills` del plugin igual a lo empotrado, sin un
// fichero de más ni de menos; `catalogo` deja el documento en su fichero; y
// una invocación que no vale o a la que le falta una entrada termina con 1 y
// una línea `empaquetar: …`. FR-001, FR-005, FR-014, FR-020, FR-070.
func TestEjecutar(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		comprobar func(t *testing.T)
	}{
		{nombre: "piezas, con la composición de producción", comprobar: comprobarLasPiezasDeProduccion},
		{nombre: "catalogo deja el documento en su fichero", comprobar: comprobarElCatalogoEscrito},
		{nombre: "con la salida de error rota, el código sigue siendo 1", comprobar: comprobarLaSalidaDeErrorRota},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			caso.comprobar(t)
		})
	}

	for _, caso := range fallosDeLaOrden() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			argumentos, linea := caso.preparar(t)

			var errores bytes.Buffer

			codigo := empaquetado.Ejecutar(argumentos, &errores)

			assert.Equal(t, 1, codigo, "un fallo del paso termina con 1 (research.md D14)")

			escrito := errores.String()

			require.True(t, strings.HasPrefix(linea, prefijoDeLaLinea),
				"la línea que el caso espera empieza por `empaquetar: `: %q", linea)
			assert.Equal(t, []string{linea}, slices.Collect(strings.Lines(escrito)),
				"el fallo es una sola línea en la salida de error, con su salto, que nombra lo que falta o lo que falló")
		})
	}
}

// ordenDePiezas es una invocación de `piezas` que vale: los dos binarios de
// prueba, el icono del árbol y una carpeta de salida que existe.
func ordenDePiezas(t *testing.T) []string {
	t.Helper()

	return []string{
		"piezas",
		"-version", versionDePrueba,
		"-macos", escribirFichero(t, "kitlegal", binarioDeMacOS),
		"-windows", escribirFichero(t, "kitlegal.exe", binarioDeWindows),
		"-icono", iconoDelArbol,
		"-salida", t.TempDir(),
	}
}

// ordenDeCatalogo es una invocación de `catalogo` que vale: una versión, una
// huella con su forma y un fichero de una carpeta que existe.
func ordenDeCatalogo(t *testing.T) []string {
	t.Helper()

	return []string{
		"catalogo",
		"-version", "0.4.0",
		"-sha256", huellaDePrueba,
		"-salida", filepath.Join(t.TempDir(), "marketplace.json"),
	}
}

// valorDe es el valor que la orden da a esa bandera, que tiene que estar.
func valorDe(t *testing.T, orden []string, bandera string) string {
	t.Helper()

	posicion := slices.Index(orden, bandera)
	require.GreaterOrEqual(t, posicion, 0, "la orden de prueba no lleva %s", bandera)

	return orden[posicion+1]
}

// con devuelve la orden con ese valor en esa bandera, que tiene que estar.
func con(t *testing.T, orden []string, bandera, valor string) []string {
	t.Helper()

	posicion := slices.Index(orden, bandera)
	require.GreaterOrEqual(t, posicion, 0, "la orden de prueba no lleva %s", bandera)

	cambiada := slices.Clone(orden)
	cambiada[posicion+1] = valor

	return cambiada
}

// sin devuelve la orden sin esa bandera ni su valor.
func sin(t *testing.T, orden []string, bandera string) []string {
	t.Helper()

	posicion := slices.Index(orden, bandera)
	require.GreaterOrEqual(t, posicion, 0, "la orden de prueba no lleva %s", bandera)

	return slices.Delete(slices.Clone(orden), posicion, posicion+2)
}

// comprobarLasPiezasDeProduccion ejecuta `piezas` como la ejecuta goreleaser,
// con el registro de producción y lo empotrado, y comprueba las dos piezas con
// las mismas comprobaciones de TestPiezas: lo que cambia es de dónde salen las
// herramientas y las skills esperadas.
func comprobarLasPiezasDeProduccion(t *testing.T) {
	t.Helper()

	orden := ordenDePiezas(t)

	var errores bytes.Buffer

	codigo := empaquetado.Ejecutar(orden, &errores)

	require.Equal(t, 0, codigo, "el paso termina con 0 cuando escribe lo pedido: %s", errores.String())
	assert.Empty(t, errores.String(), "si termina bien no escribe nada en la salida de error")

	registro, err := app.RegistroDeProduccion("")
	require.NoError(t, err)

	anunciadas := app.HerramientasAnunciadas(registro)

	nombres := make([]string, 0, len(anunciadas))
	for _, anunciada := range anunciadas {
		nombres = append(nombres, anunciada.Nombre)
	}

	require.Subset(t, nombres, herramientasDeHoy,
		"el registro de producción anuncia al menos las diez herramientas de hoy (FR-014)")

	empotrado := kitlegal.Skills()
	ficheros := ficherosDe(t, empotrado)

	require.Subset(t, ficheros, skillsDeHoy, "lo empotrado lleva al menos las dos skills de hoy (FR-020)")

	piezas := empaquetado.PiezasAEscribir{
		Version:      valorDe(t, orden, "-version"),
		MacOS:        valorDe(t, orden, "-macos"),
		Windows:      valorDe(t, orden, "-windows"),
		Icono:        valorDe(t, orden, "-icono"),
		Salida:       valorDe(t, orden, "-salida"),
		Descripcion:  empaquetado.Descripcion,
		Herramientas: anunciadas,
		Skills:       empotrado,
	}

	comprobarLaSalida(t, piezas.Salida)
	// `tools` es lo que da app.HerramientasAnunciadas del registro de producción,
	// en su orden, y la descripción corta, la del repositorio.
	comprobarLaExtension(t, piezas)
	// El árbol `skills` del plugin es lo empotrado: ni un fichero de más ni uno
	// de menos, y cada uno byte a byte.
	comprobarElPlugin(t, piezas, ficheros)
}

// ficherosDe son las rutas de los ficheros de la carpeta skills del árbol, en
// el orden en que fs.WalkDir los da.
func ficherosDe(t *testing.T, arbol fs.FS) []string {
	t.Helper()

	var ficheros []string

	require.NoError(t, fs.WalkDir(arbol, "skills", func(ruta string, entrada fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !entrada.IsDir() {
			ficheros = append(ficheros, ruta)
		}

		return nil
	}))

	return ficheros
}

// comprobarElCatalogoEscrito ejecuta `catalogo` y comprueba que deja en su
// fichero, y solo en él, el documento de esa versión y esa huella, que
// TestCatalogo lee campo a campo.
func comprobarElCatalogoEscrito(t *testing.T) {
	t.Helper()

	orden := ordenDeCatalogo(t)

	var errores bytes.Buffer

	codigo := empaquetado.Ejecutar(orden, &errores)

	require.Equal(t, 0, codigo, "el paso termina con 0 cuando escribe lo pedido: %s", errores.String())
	assert.Empty(t, errores.String(), "si termina bien no escribe nada en la salida de error")

	salida := valorDe(t, orden, "-salida")

	documento, err := empaquetado.DocumentoDelCatalogo(valorDe(t, orden, "-version"), valorDe(t, orden, "-sha256"))
	require.NoError(t, err)

	assert.Equal(t, string(documento), string(leerFichero(t, salida)),
		"el fichero de salida lleva el catálogo de esa versión y esa huella (FR-030)")

	escritos, err := os.ReadDir(filepath.Dir(salida))
	require.NoError(t, err)
	assert.Len(t, escritos, 1, "`catalogo` no deja nada más en la carpeta de su fichero")
}

// escritorRoto es una salida de error en la que no se puede escribir.
type escritorRoto struct{}

// Write falla siempre.
func (escritorRoto) Write([]byte) (int, error) {
	return 0, errors.New("la salida de error está rota")
}

// comprobarLaSalidaDeErrorRota comprueba que el código de un fallo no depende
// de que su línea llegue a alguien.
func comprobarLaSalidaDeErrorRota(t *testing.T) {
	t.Helper()

	assert.Equal(t, 1, empaquetado.Ejecutar(nil, escritorRoto{}))
}

// falloDeLaOrden es una invocación del paso que termina con 1.
type falloDeLaOrden struct {
	nombre string
	// preparar da los argumentos de la invocación y la línea que el paso
	// escribe en la salida de error, con su salto.
	preparar func(t *testing.T) (argumentos []string, linea string)
}

// fallosDeLaOrden son las invocaciones de contracts/paso.md §1 que terminan con
// 1, cada una con su línea: sin orden, con una desconocida, sin una bandera —o
// con ella vacía—, sin un binario, con una huella sin su forma y con una salida
// que no se puede escribir. Una invocación que no tiene la forma de la orden
// —una bandera que el paso no tiene, una sin su valor, un argumento de más, la
// ayuda— responde con la línea de uso.
func fallosDeLaOrden() []falloDeLaOrden {
	uso := func(argumentos func(t *testing.T) []string) func(t *testing.T) ([]string, string) {
		return func(t *testing.T) ([]string, string) {
			t.Helper()

			return argumentos(t), lineaDeUso
		}
	}

	fallos := []falloDeLaOrden{
		{nombre: "sin orden", preparar: uso(func(*testing.T) []string { return nil })},
		{nombre: "con una orden desconocida", preparar: uso(func(*testing.T) []string { return []string{"publicar"} })},
		{
			nombre: "piezas con una bandera que el paso no tiene",
			preparar: uso(func(t *testing.T) []string {
				t.Helper()

				return append(ordenDePiezas(t), "-linux", "kitlegal")
			}),
		},
		{
			nombre: "piezas con un argumento de más",
			preparar: uso(func(t *testing.T) []string {
				t.Helper()

				return append(ordenDePiezas(t), "sobra")
			}),
		},
		{
			nombre: "piezas con una bandera sin su valor al final",
			preparar: uso(func(t *testing.T) []string {
				t.Helper()

				return append(sin(t, ordenDePiezas(t), "-salida"), "-salida")
			}),
		},
		{
			nombre: "catalogo pidiendo la ayuda",
			preparar: uso(func(t *testing.T) []string {
				t.Helper()

				return append(ordenDeCatalogo(t), "-h")
			}),
		},
		{
			nombre: "piezas con una bandera vacía",
			preparar: func(t *testing.T) ([]string, string) {
				t.Helper()

				return con(t, ordenDePiezas(t), "-version", ""), "empaquetar: falta -version\n"
			},
		},
		{
			nombre: "piezas sin un binario",
			preparar: func(t *testing.T) ([]string, string) {
				t.Helper()

				noExiste := filepath.Join(t.TempDir(), "no-existe.exe")

				return con(t, ordenDePiezas(t), "-windows", noExiste),
					"empaquetar: falta el binario de Windows: open " + noExiste + ": " + syscall.ENOENT.Error() + "\n"
			},
		},
		{
			nombre: "catalogo con una huella sin su forma",
			preparar: func(t *testing.T) ([]string, string) {
				t.Helper()

				return con(t, ordenDeCatalogo(t), "-sha256", "ABC"),
					"empaquetar: «ABC» no es una huella SHA-256: 64 dígitos hexadecimales en minúsculas\n"
			},
		},
		{
			nombre: "catalogo con una salida en una carpeta que no existe",
			preparar: func(t *testing.T) ([]string, string) {
				t.Helper()

				salida := filepath.Join(t.TempDir(), "no-existe", "marketplace.json")

				return con(t, ordenDeCatalogo(t), "-salida", salida),
					"empaquetar: no se puede escribir " + salida + ": " + syscall.ENOENT.Error() + "\n"
			},
		},
	}

	for _, bandera := range []string{"-version", "-macos", "-windows", "-icono", "-salida"} {
		fallos = append(fallos, falloDeLaOrden{
			nombre: "piezas sin " + bandera,
			preparar: func(t *testing.T) ([]string, string) {
				t.Helper()

				return sin(t, ordenDePiezas(t), bandera), "empaquetar: falta " + bandera + "\n"
			},
		})
	}

	for _, bandera := range []string{"-version", "-sha256", "-salida"} {
		fallos = append(fallos, falloDeLaOrden{
			nombre: "catalogo sin " + bandera,
			preparar: func(t *testing.T) ([]string, string) {
				t.Helper()

				return sin(t, ordenDeCatalogo(t), bandera), "empaquetar: falta " + bandera + "\n"
			},
		})
	}

	return fallos
}
