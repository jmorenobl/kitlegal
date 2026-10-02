package empaquetado_test

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/app"
	"github.com/jmorenobl/kitlegal/internal/empaquetado"
)

// Lo que el contrato fija de las dos piezas, escrito aquí como lo escribe
// contracts/paso.md y no leído del paquete: es contra lo que se compara.
const (
	// nombreDeLaExtension y nombreDelPlugin son los dos ficheros que deja la
	// orden piezas (FR-002).
	nombreDeLaExtension = "kitlegal.mcpb"
	nombreDelPlugin     = "kitlegal-plugin.zip"
	// rutaDelManifiesto y rutaDeLaFichaDelPlugin son las entradas de los dos
	// documentos JSON.
	rutaDelManifiesto      = "manifest.json"
	rutaDeLaFichaDelPlugin = ".claude-plugin/plugin.json"
	// iconoDelArbol es el icono versionado, visto desde este paquete (FR-016).
	iconoDelArbol = "../../mcp/icon.png"
	// versionDePrueba es la que los tests dan al paso, con la forma de la de
	// un snapshot.
	versionDePrueba = "0.0.0-prueba"
	// ladoDelIcono es el umbral de FR-016, en píxeles.
	ladoDelIcono = 512
)

// fechaDeLasEntradas es la fecha de modificación de toda entrada de los dos
// zips (research.md D8).
var fechaDeLasEntradas = time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC)

// Los dos binarios de prueba: unos bytes cualesquiera, con un nulo, un salto
// de línea de cada clase y bytes que no son UTF-8, que es lo que estropearía
// quien los mirase como texto. El paso los copia sin mirarlos (FR-011).
var (
	binarioDeMacOS   = []byte("binario de macOS de prueba\x00\xca\xfe\xba\xbe\r\n\xff")
	binarioDeWindows = []byte("binario de Windows de prueba\x00MZ\r\n\xfe\xff")
)

// manifiestoLeido es el manifiesto como lo fija data-model §3: sus campos, en
// su orden y ninguno más. Es la transcripción del contrato con la que los tests
// leen de forma estricta lo que el paso escribe y con la que comparan su forma;
// no es el tipo del paquete.
type manifiestoLeido struct {
	VersionDelManifiesto string              `json:"manifest_version"`
	Nombre               string              `json:"name"`
	NombreVisible        string              `json:"display_name"`
	Version              string              `json:"version"`
	Descripcion          string              `json:"description"`
	DescripcionLarga     string              `json:"long_description"`
	Autoria              autoriaLeida        `json:"author"`
	Pagina               string              `json:"homepage"`
	Licencia             string              `json:"license"`
	Icono                string              `json:"icon"`
	Servidor             servidorLeido       `json:"server"`
	Herramientas         []herramientaLeida  `json:"tools"`
	Compatibilidad       compatibilidadLeida `json:"compatibility"`
}

// autoriaLeida es `author` del manifiesto, de plugin.json y de la entrada del
// catálogo, y `owner` del catálogo.
type autoriaLeida struct {
	Nombre string `json:"name"`
}

// servidorLeido es `server` del manifiesto.
type servidorLeido struct {
	Tipo     string             `json:"type"`
	Entrada  string             `json:"entry_point"`
	Arranque configuracionLeida `json:"mcp_config"`
}

// configuracionLeida es `mcp_config`: la orden con la que la app arranca el
// servidor (FR-012, FR-064).
type configuracionLeida struct {
	Orden         string            `json:"command"`
	Argumentos    []string          `json:"args"`
	PorPlataforma plataformasLeidas `json:"platform_overrides"`
}

// plataformasLeidas es `platform_overrides`: solo Windows cambia la orden, y la
// lectura estricta rechaza cualquier otra plataforma.
type plataformasLeidas struct {
	Win32 ordenLeida `json:"win32"`
}

// ordenLeida es lo que una plataforma cambia de `mcp_config`.
type ordenLeida struct {
	Orden string `json:"command"`
}

// herramientaLeida es un elemento de `tools`.
type herramientaLeida struct {
	Nombre      string `json:"name"`
	Descripcion string `json:"description"`
}

// compatibilidadLeida es `compatibility`.
type compatibilidadLeida struct {
	Plataformas []string `json:"platforms"`
}

// fichaDelPluginLeida es plugin.json como lo fija data-model §4: sin
// `mcpServers`, que la lectura estricta rechazaría (FR-021, FR-022).
type fichaDelPluginLeida struct {
	Nombre      string       `json:"name"`
	Version     string       `json:"version"`
	Descripcion string       `json:"description"`
	Autoria     autoriaLeida `json:"author"`
	Pagina      string       `json:"homepage"`
	Licencia    string       `json:"license"`
}

// manifiestoEsperado es el manifiesto de data-model §3 para esa versión, esa
// descripción corta y esas herramientas: lo fijo, escrito aquí; los textos, de
// su único sitio (FR-015).
func manifiestoEsperado(version, descripcion string, herramientas []app.HerramientaAnunciada) manifiestoLeido {
	return manifiestoLeido{
		VersionDelManifiesto: "0.3",
		Nombre:               "kitlegal",
		NombreVisible:        empaquetado.NombreVisible,
		Version:              version,
		Descripcion:          descripcion,
		DescripcionLarga:     empaquetado.DescripcionLarga,
		Autoria:              autoriaLeida{Nombre: empaquetado.Autoria},
		Pagina:               "https://kitlegal.es",
		Licencia:             "EUPL-1.2",
		Icono:                "icon.png",
		Servidor: servidorLeido{
			Tipo:    "binary",
			Entrada: "server/kitlegal",
			Arranque: configuracionLeida{
				Orden:         "${__dirname}/server/kitlegal",
				Argumentos:    []string{"mcp", "serve"},
				PorPlataforma: plataformasLeidas{Win32: ordenLeida{Orden: "${__dirname}/server/kitlegal.exe"}},
			},
		},
		Herramientas:   herramientasLeidas(herramientas),
		Compatibilidad: compatibilidadLeida{Plataformas: []string{"darwin", "win32"}},
	}
}

// fichaDelPluginEsperada es el plugin.json de data-model §4 para esa versión y
// esa descripción corta.
func fichaDelPluginEsperada(version, descripcion string) fichaDelPluginLeida {
	return fichaDelPluginLeida{
		Nombre:      "kitlegal",
		Version:     version,
		Descripcion: descripcion,
		Autoria:     autoriaLeida{Nombre: empaquetado.Autoria},
		Pagina:      "https://kitlegal.es",
		Licencia:    "EUPL-1.2",
	}
}

// herramientasLeidas son las herramientas anunciadas como van en `tools`: el
// nombre y la descripción de cada una, en su orden.
func herramientasLeidas(anunciadas []app.HerramientaAnunciada) []herramientaLeida {
	leidas := make([]herramientaLeida, 0, len(anunciadas))
	for _, anunciada := range anunciadas {
		leidas = append(leidas, herramientaLeida{Nombre: anunciada.Nombre, Descripcion: anunciada.Descripcion})
	}

	return leidas
}

// leerEstricto lee un documento JSON en destino sin admitir un campo que su
// tipo no declare ni nada detrás del documento.
func leerEstricto(t *testing.T, documento []byte, destino any) {
	t.Helper()

	lector := json.NewDecoder(bytes.NewReader(documento))
	lector.DisallowUnknownFields()

	require.NoError(t, lector.Decode(destino), "el documento no es el del contrato:\n%s", documento)

	_, err := lector.Token()
	require.ErrorIs(t, err, io.EOF, "detrás del documento no hay nada")
}

// conLaFormaDelContrato escribe un valor como contracts/paso.md §5 pide los
// tres documentos JSON: los campos en el orden de su tipo, sangría de dos
// espacios, `<`, `>` y `&` sin escapar y salto de línea final. Con un tipo que
// transcribe el contrato, lo que devuelve es el documento esperado, byte a
// byte.
func conLaFormaDelContrato(t *testing.T, valor any) string {
	t.Helper()

	var escrito bytes.Buffer

	codificador := json.NewEncoder(&escrito)
	codificador.SetEscapeHTML(false)
	codificador.SetIndent("", "  ")

	require.NoError(t, codificador.Encode(valor))

	return escrito.String()
}

// entradaLeida es una entrada de un zip como la ve quien lo abre.
type entradaLeida struct {
	Nombre    string
	Modo      fs.FileMode
	Contenido []byte
}

// formaDeEntrada es lo que una entrada dice de sí sin su contenido: su ruta y
// su modo.
type formaDeEntrada struct {
	Nombre string
	Modo   fs.FileMode
}

// formasDe son las rutas y los modos de las entradas, en su orden.
func formasDe(entradas []entradaLeida) []formaDeEntrada {
	formas := make([]formaDeEntrada, 0, len(entradas))
	for _, entrada := range entradas {
		formas = append(formas, formaDeEntrada{Nombre: entrada.Nombre, Modo: entrada.Modo})
	}

	return formas
}

// leerZip abre el zip de esa ruta y devuelve sus entradas en el orden en que
// están escritas, tras comprobar lo que research.md D8 fija de todas: Deflate,
// la fecha fija y ningún comentario, tampoco el del zip.
func leerZip(t *testing.T, ruta string) []entradaLeida {
	t.Helper()

	datos := leerFichero(t, ruta)

	lector, err := zip.NewReader(bytes.NewReader(datos), int64(len(datos)))
	require.NoError(t, err, "%s no es un zip", ruta)

	assert.Empty(t, lector.Comment, "%s no lleva comentario", ruta)

	entradas := make([]entradaLeida, 0, len(lector.File))

	for _, fichero := range lector.File {
		assert.Equal(t, zip.Deflate, fichero.Method, "%s va comprimida con Deflate", fichero.Name)
		assert.True(t, fichero.Modified.Equal(fechaDeLasEntradas),
			"%s lleva la fecha fija, 1980-01-01T00:00:00Z, y no %s", fichero.Name, fichero.Modified)
		assert.Empty(t, fichero.Comment, "%s no lleva comentario", fichero.Name)

		abierto, err := fichero.Open()
		require.NoError(t, err)

		contenido, err := io.ReadAll(abierto)
		require.NoError(t, err)
		require.NoError(t, abierto.Close())

		entradas = append(entradas, entradaLeida{Nombre: fichero.Name, Modo: fichero.Mode(), Contenido: contenido})
	}

	return entradas
}

// contenidoDe es el contenido de la entrada con ese nombre, que tiene que
// estar.
func contenidoDe(t *testing.T, entradas []entradaLeida, nombre string) []byte {
	t.Helper()

	for _, entrada := range entradas {
		if entrada.Nombre == nombre {
			return entrada.Contenido
		}
	}

	require.FailNow(t, "falta la entrada", "el zip no lleva %s", nombre)

	return nil
}

// leerFichero lee un fichero que tiene que estar.
func leerFichero(t *testing.T, ruta string) []byte {
	t.Helper()

	datos, err := os.ReadFile(filepath.Clean(ruta))
	require.NoError(t, err)

	return datos
}

// escribirFichero deja esos bytes en un fichero con ese nombre de una carpeta
// temporal nueva y devuelve su ruta.
func escribirFichero(t *testing.T, nombre string, datos []byte) string {
	t.Helper()

	ruta := filepath.Join(t.TempDir(), nombre)
	require.NoError(t, os.WriteFile(ruta, datos, 0o600))

	return ruta
}

// escribirPNG deja en una carpeta temporal un PNG de esas medidas y devuelve
// su ruta.
func escribirPNG(t *testing.T, ancho, alto int) string {
	t.Helper()

	var codificado bytes.Buffer

	require.NoError(t, png.Encode(&codificado, image.NewRGBA(image.Rect(0, 0, ancho, alto))))

	return escribirFichero(t, "icono.png", codificado.Bytes())
}

// dosHerramientas son las herramientas que los tests dan al paso cuando no
// miden las de producción. La descripción de la primera lleva los tres
// caracteres que un JSON para HTML escaparía.
func dosHerramientas() []app.HerramientaAnunciada {
	return []app.HerramientaAnunciada{
		{Nombre: "alfa_leer", Descripcion: "Lee <alfa> & lo cita."},
		{Nombre: "beta_contar", Descripcion: "Cuenta lo de beta."},
	}
}

// dosSkills es el árbol de skills que los tests dan al paso cuando no miden lo
// empotrado: dos skills, escritas en un orden que no es el del recorrido, con
// un fichero bajo un subdirectorio de references/ —que fs.WalkDir da antes que
// su vecino `a-b.md`, al revés que el orden de las rutas— y uno marcado como
// ejecutable, que en el plugin va con el modo de los demás (FR-020).
func dosSkills() fstest.MapFS {
	return fstest.MapFS{
		"skills/zeta/SKILL.md":            {Data: []byte("# zeta\n")},
		"skills/alfa/references/a-b.md":   {Data: []byte("a-b\x00\xff")},
		"skills/alfa/references/a/b.md":   {Data: []byte("a/b\r\n"), Mode: 0o755},
		"skills/alfa/SKILL.md":            {Data: []byte("# alfa\n")},
		"skills/alfa/references/.oculto":  {Data: []byte("oculto")},
		"skills/alfa/references/vacio.md": {},
	}
}

// ficherosDeDosSkills son los ficheros de dosSkills en el orden en que
// fs.WalkDir los da, escrito a mano.
func ficherosDeDosSkills() []string {
	return []string{
		"skills/alfa/SKILL.md",
		"skills/alfa/references/.oculto",
		"skills/alfa/references/a/b.md",
		"skills/alfa/references/a-b.md",
		"skills/alfa/references/vacio.md",
		"skills/zeta/SKILL.md",
	}
}

// piezasDePrueba es lo que los tests dan al paso: los dos binarios de prueba en
// una carpeta temporal, el icono del árbol, una carpeta de salida temporal que
// existe, la descripción corta del repositorio, dos herramientas y dos skills.
// Cada test cambia lo que mide.
func piezasDePrueba(t *testing.T) empaquetado.PiezasAEscribir {
	t.Helper()

	return empaquetado.PiezasAEscribir{
		Version:      versionDePrueba,
		MacOS:        escribirFichero(t, "kitlegal", binarioDeMacOS),
		Windows:      escribirFichero(t, "kitlegal.exe", binarioDeWindows),
		Icono:        iconoDelArbol,
		Salida:       t.TempDir(),
		Descripcion:  empaquetado.Descripcion,
		Herramientas: dosHerramientas(),
		Skills:       dosSkills(),
	}
}

// TestPiezas es el control en `make ci` de lo que la orden piezas escribe con
// unas herramientas y unas skills dadas (contracts/paso.md §2, §3 y §7): la
// extensión con exactamente sus cuatro entradas, en orden, con los bytes del
// icono y de los dos binarios y sus modos; el manifiesto y plugin.json, leídos
// de forma estricta, con cada campo de data-model §3 y §4 y ninguno más; y el
// plugin con cada fichero del árbol, byte a byte, y nada más. Con una
// herramienta y una skill más en la entrada, aparecen sin tocar el paso (US5).
// FR-002, FR-010 a FR-017, FR-020 a FR-022, FR-070; SC-004, SC-005, SC-007.
func TestPiezas(t *testing.T) {
	t.Parallel()

	unaHerramientaMas := append(dosHerramientas(),
		app.HerramientaAnunciada{Nombre: "gamma_ver", Descripcion: "Enseña lo de gamma."})

	unaSkillMas := dosSkills()
	unaSkillMas["skills/mu/SKILL.md"] = &fstest.MapFile{Data: []byte("# mu\n")}
	unaSkillMas["skills/mu/references/normas.md"] = &fstest.MapFile{Data: []byte("normas de mu\n")}

	casos := []struct {
		nombre       string
		herramientas []app.HerramientaAnunciada
		skills       fstest.MapFS
		// ficheros son los del árbol, en el orden de fs.WalkDir, escritos a mano.
		ficheros []string
	}{
		{
			nombre:       "con dos herramientas y dos skills",
			herramientas: dosHerramientas(),
			skills:       dosSkills(),
			ficheros:     ficherosDeDosSkills(),
		},
		{
			nombre:       "con una herramienta y una skill más, aparecen",
			herramientas: unaHerramientaMas,
			skills:       unaSkillMas,
			ficheros: []string{
				"skills/alfa/SKILL.md",
				"skills/alfa/references/.oculto",
				"skills/alfa/references/a/b.md",
				"skills/alfa/references/a-b.md",
				"skills/alfa/references/vacio.md",
				"skills/mu/SKILL.md",
				"skills/mu/references/normas.md",
				"skills/zeta/SKILL.md",
			},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			piezas := piezasDePrueba(t)
			piezas.Herramientas = caso.herramientas
			piezas.Skills = caso.skills

			require.NoError(t, empaquetado.EscribirPiezas(piezas))

			comprobarLaSalida(t, piezas.Salida)
			comprobarLaExtension(t, piezas)
			comprobarElPlugin(t, piezas, caso.ficheros)
		})
	}
}

// comprobarLaSalida comprueba que la carpeta de salida lleva los dos ficheros
// del paso y ninguno más (FR-002).
func comprobarLaSalida(t *testing.T, salida string) {
	t.Helper()

	escritos, err := os.ReadDir(salida)
	require.NoError(t, err)

	nombres := make([]string, 0, len(escritos))
	for _, escrito := range escritos {
		nombres = append(nombres, escrito.Name())
	}

	assert.Equal(t, []string{nombreDelPlugin, nombreDeLaExtension}, nombres,
		"el paso deja en la carpeta de salida sus dos ficheros y nada más")
}

// comprobarLaExtension comprueba el `.mcpb` que el paso dejó con esas piezas:
// sus cuatro entradas y sus modos, los bytes de lo que copia y el manifiesto.
func comprobarLaExtension(t *testing.T, piezas empaquetado.PiezasAEscribir) {
	t.Helper()

	extension := leerZip(t, filepath.Join(piezas.Salida, nombreDeLaExtension))

	assert.Equal(t, []formaDeEntrada{
		{Nombre: rutaDelManifiesto, Modo: 0o644},
		{Nombre: "icon.png", Modo: 0o644},
		{Nombre: "server/kitlegal", Modo: 0o755},
		{Nombre: "server/kitlegal.exe", Modo: 0o755},
	}, formasDe(extension),
		"la extensión lleva exactamente cuatro entradas, en este orden, con los dos binarios como ejecutables (FR-010)")

	assert.Equal(t, leerFichero(t, piezas.Icono), contenidoDe(t, extension, "icon.png"),
		"icon.png es el icono que se le da, byte a byte (FR-016)")
	assert.Equal(t, binarioDeMacOS, contenidoDe(t, extension, "server/kitlegal"),
		"server/kitlegal es el binario de macOS que se le da, sin cambiar un byte (FR-011)")
	assert.Equal(t, binarioDeWindows, contenidoDe(t, extension, "server/kitlegal.exe"),
		"server/kitlegal.exe es el binario de Windows que se le da, sin cambiar un byte (FR-011)")

	escrito := contenidoDe(t, extension, rutaDelManifiesto)
	esperado := manifiestoEsperado(piezas.Version, piezas.Descripcion, piezas.Herramientas)

	var manifiesto manifiestoLeido

	leerEstricto(t, escrito, &manifiesto)

	assert.Equal(t, esperado, manifiesto,
		"el manifiesto lleva cada campo de data-model §3 con su valor, sin `user_config` ni ninguno más, con la versión "+
			"tal cual y `tools` de las herramientas que se le dan, en su orden (FR-012 a FR-015, FR-017)")
	require.NoError(t, validarConElEsquemaOficial(t, escrito),
		"el manifiesto cumple el esquema oficial de la versión `0.3` del manifiesto de MCP Bundle (FR-017)")
	assert.Equal(t, conLaFormaDelContrato(t, esperado), string(escrito),
		"el manifiesto va con sus campos en el orden del contrato, sangría de dos espacios, sin escapar `<`, `>` ni "+
			"`&` y con salto de línea final (contracts/paso.md §5)")
	assert.Contains(t, string(escrito), piezas.Herramientas[0].Descripcion,
		"la descripción de una herramienta va tal cual, con su `<`, su `>` y su `&`")
}

// comprobarElPlugin comprueba el plugin que el paso dejó con esas piezas:
// plugin.json y cada fichero del árbol, en el orden de fs.WalkDir, byte a byte
// y con el modo de un fichero, y nada más.
func comprobarElPlugin(t *testing.T, piezas empaquetado.PiezasAEscribir, ficheros []string) {
	t.Helper()

	plugin := leerZip(t, filepath.Join(piezas.Salida, nombreDelPlugin))

	formas := []formaDeEntrada{{Nombre: rutaDeLaFichaDelPlugin, Modo: 0o644}}
	for _, fichero := range ficheros {
		formas = append(formas, formaDeEntrada{Nombre: fichero, Modo: 0o644})
	}

	require.Equal(t, formas, formasDe(plugin),
		"el plugin lleva plugin.json y, detrás, cada fichero del árbol con su ruta bajo skills, en el orden de "+
			"fs.WalkDir, todos con modo 0644 y sin entradas de directorio, y nada más (FR-020, FR-022)")

	for _, fichero := range ficheros {
		delArbol, err := fs.ReadFile(piezas.Skills, fichero)
		require.NoError(t, err)

		assert.Equal(t, delArbol, contenidoDe(t, plugin, fichero), "%s va byte a byte (FR-020)", fichero)
	}

	escrita := contenidoDe(t, plugin, rutaDeLaFichaDelPlugin)
	esperada := fichaDelPluginEsperada(piezas.Version, piezas.Descripcion)

	var ficha fichaDelPluginLeida

	leerEstricto(t, escrita, &ficha)

	assert.Equal(t, esperada, ficha,
		"plugin.json lleva cada campo de data-model §4 con su valor, sin `mcpServers` ni ninguno más, con la versión "+
			"del manifiesto (FR-021, FR-022)")
	assert.Equal(t, conLaFormaDelContrato(t, esperada), string(escrita),
		"plugin.json va con la forma de contracts/paso.md §5")
}

// TestPiezasReproducibles es el control en `make ci` de FR-004 y FR-067: dos
// ejecuciones del paso sobre los mismos binarios, en dos carpetas, dan los
// mismos dos ficheros, byte a byte (SC-010). Entre las dos se cambia la fecha
// de los binarios en el disco, que es lo que cambia de una construcción a
// otra sin que cambie un byte de ellos: lo que el paso escribe no depende de
// ella. Que las fechas de las entradas no son las de cada ejecución lo fija
// TestPiezas, con la fecha fija.
func TestPiezasReproducibles(t *testing.T) {
	t.Parallel()

	primeras := piezasDePrueba(t)

	require.NoError(t, empaquetado.EscribirPiezas(primeras))

	segundas := primeras
	segundas.Salida = t.TempDir()

	otraFecha := time.Date(2001, time.February, 3, 4, 5, 6, 0, time.UTC)
	require.NoError(t, os.Chtimes(segundas.MacOS, otraFecha, otraFecha))
	require.NoError(t, os.Chtimes(segundas.Windows, otraFecha, otraFecha))

	require.NoError(t, empaquetado.EscribirPiezas(segundas))

	for _, nombre := range []string{nombreDeLaExtension, nombreDelPlugin} {
		primero := leerFichero(t, filepath.Join(primeras.Salida, nombre))
		segundo := leerFichero(t, filepath.Join(segundas.Salida, nombre))

		require.NotEmpty(t, primero, "%s no está vacío", nombre)
		assert.True(t, bytes.Equal(primero, segundo),
			"las dos ejecuciones dan el mismo %s byte a byte: 0 bytes de diferencia (FR-004)", nombre)
	}
}

// TestPiezasSinEntrada es el control en `make ci` de FR-005 sobre la orden
// piezas: sin el binario de macOS, sin el de Windows, sin el icono y con una
// salida que no se puede escribir —la carpeta no existe, o donde va uno de los
// dos ficheros hay una carpeta—, el paso falla con un error que nombra lo que
// falta o lo que falló, con su ruta una sola vez, y la causa del sistema
// (contracts/paso.md §1). El último caso es el árbol de skills sin su carpeta,
// que lo empotrado no da nunca: tampoco entonces sale un plugin sin skills.
func TestPiezasSinEntrada(t *testing.T) {
	t.Parallel()

	noExiste := func(t *testing.T) string {
		t.Helper()

		return filepath.Join(t.TempDir(), "no-existe")
	}

	casos := []struct {
		nombre string
		// quitar deja las piezas sin la entrada del caso y devuelve lo que el
		// error dice delante de la causa, con la ruta que tiene que nombrar.
		quitar func(t *testing.T, piezas *empaquetado.PiezasAEscribir) string
		// causa es la del sistema, que el error conserva y con cuyo texto
		// termina.
		causa error
	}{
		{
			nombre: "sin el binario de macOS",
			quitar: func(t *testing.T, piezas *empaquetado.PiezasAEscribir) string {
				t.Helper()

				piezas.MacOS = noExiste(t)

				return "falta el binario de macOS: open " + piezas.MacOS + ": "
			},
			causa: syscall.ENOENT,
		},
		{
			nombre: "sin el binario de Windows",
			quitar: func(t *testing.T, piezas *empaquetado.PiezasAEscribir) string {
				t.Helper()

				piezas.Windows = noExiste(t)

				return "falta el binario de Windows: open " + piezas.Windows + ": "
			},
			causa: syscall.ENOENT,
		},
		{
			nombre: "sin el icono",
			quitar: func(t *testing.T, piezas *empaquetado.PiezasAEscribir) string {
				t.Helper()

				piezas.Icono = noExiste(t)

				return "falta el icono: open " + piezas.Icono + ": "
			},
			causa: syscall.ENOENT,
		},
		{
			nombre: "con una carpeta de salida que no existe",
			quitar: func(t *testing.T, piezas *empaquetado.PiezasAEscribir) string {
				t.Helper()

				piezas.Salida = noExiste(t)

				return "no se puede escribir " + filepath.Join(piezas.Salida, nombreDeLaExtension) + ": "
			},
			causa: syscall.ENOENT,
		},
		{
			nombre: "con una carpeta donde va el plugin",
			quitar: func(t *testing.T, piezas *empaquetado.PiezasAEscribir) string {
				t.Helper()

				ocupado := filepath.Join(piezas.Salida, nombreDelPlugin)
				require.NoError(t, os.Mkdir(ocupado, 0o700))

				return "no se puede escribir " + ocupado + ": "
			},
			causa: syscall.EISDIR,
		},
		{
			nombre: "con un árbol de skills sin su carpeta",
			quitar: func(t *testing.T, piezas *empaquetado.PiezasAEscribir) string {
				t.Helper()

				piezas.Skills = fstest.MapFS{}

				return "las skills no se pueden leer: open skills: "
			},
			causa: fs.ErrNotExist,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			piezas := piezasDePrueba(t)
			delante := caso.quitar(t, &piezas)

			err := empaquetado.EscribirPiezas(piezas)

			require.ErrorIs(t, err, caso.causa, "el error conserva la causa del sistema")
			assert.Equal(t, delante+caso.causa.Error(), err.Error(),
				"el error nombra lo que falta o lo que falló, con su ruta una sola vez, y su causa (FR-005)")
		})
	}

	// Una ruta de más de 65 535 bytes no cabe en un zip. Lo empotrado tampoco la
	// da nunca: lo que se fija es que el paso falla nombrándola, en lugar de
	// dejar un zip que nadie puede abrir.
	t.Run("con una skill cuya ruta no cabe en un zip", func(t *testing.T) {
		t.Parallel()

		larga := "skills/" + strings.Repeat("a", 1<<16) + "/SKILL.md"

		piezas := piezasDePrueba(t)
		piezas.Skills = fstest.MapFS{
			larga:                  {Data: []byte("# larga\n")},
			"skills/zeta/SKILL.md": {Data: []byte("# zeta\n")},
		}

		err := empaquetado.EscribirPiezas(piezas)

		require.Error(t, err)
		assert.True(t, strings.HasPrefix(err.Error(), "la entrada "+larga+" no cabe en el zip: "),
			"el error nombra la entrada que no cabe: %.80s…", err.Error())

		escritos, err := os.ReadDir(piezas.Salida)
		require.NoError(t, err)
		assert.Empty(t, escritos, "el paso compone las dos piezas antes de escribir ninguna")
	})
}

// TestIcono es el control en `make ci` del umbral del icono (FR-016, FR-066;
// SC-009): el icono versionado del árbol es un PNG de 512 × 512 px y el paso lo
// acepta; uno de otro tamaño, en cualquiera de sus dos medidas, y unos bytes
// que no son un PNG, creados aquí, lo hacen fallar con la línea del contrato
// (contracts/paso.md §1; research.md D20).
func TestIcono(t *testing.T) {
	t.Parallel()

	t.Run("el del árbol es un PNG de 512 × 512 px y el paso lo acepta", func(t *testing.T) {
		t.Parallel()

		medidas, err := png.DecodeConfig(bytes.NewReader(leerFichero(t, iconoDelArbol)))
		require.NoError(t, err, "mcp/icon.png es un PNG")

		assert.Equal(t, ladoDelIcono, medidas.Width, "mcp/icon.png mide 512 px de ancho")
		assert.Equal(t, ladoDelIcono, medidas.Height, "mcp/icon.png mide 512 px de alto")

		require.NoError(t, empaquetado.EscribirPiezas(piezasDePrueba(t)))
	})

	otrasMedidas := []struct {
		nombre      string
		ancho, alto int
		loQueEs     string
	}{
		{nombre: "uno de 256 × 256 px falla", ancho: 256, alto: 256, loQueEs: "mide 256 × 256 px"},
		{nombre: "uno de 512 px de ancho y 256 de alto falla", ancho: 512, alto: 256, loQueEs: "mide 512 × 256 px"},
		{nombre: "uno de 256 px de ancho y 512 de alto falla", ancho: 256, alto: 512, loQueEs: "mide 256 × 512 px"},
	}

	for _, caso := range otrasMedidas {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			piezas := piezasDePrueba(t)
			piezas.Icono = escribirPNG(t, caso.ancho, caso.alto)

			require.EqualError(t, empaquetado.EscribirPiezas(piezas),
				"el icono "+piezas.Icono+" no es un PNG de 512 × 512 px: "+caso.loQueEs)
		})
	}

	t.Run("unos bytes que no son un PNG fallan", func(t *testing.T) {
		t.Parallel()

		piezas := piezasDePrueba(t)
		piezas.Icono = escribirFichero(t, "icono.png", []byte("esto no es un PNG, aunque su nombre lo diga"))

		err := empaquetado.EscribirPiezas(piezas)

		var formato png.FormatError

		require.ErrorAs(t, err, &formato, "lo que es lo dice quien lo intentó leer como PNG")
		assert.Equal(t, "el icono "+piezas.Icono+" no es un PNG de 512 × 512 px: "+formato.Error(), err.Error())
	})
}
