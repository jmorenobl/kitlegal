package empaquetado

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/jmorenobl/kitlegal/internal/app"
)

// Los dos ficheros que la orden piezas deja en la carpeta de salida. No llevan
// la versión en el nombre, para que «la última release» sea una dirección fija
// (H22 FR-002).
const (
	nombreDeLaExtension = "kitlegal.mcpb"
	nombreDelPlugin     = "kitlegal-plugin.zip"
)

// Lo que el manifiesto, plugin.json y el catálogo dicen igual y no es un texto
// de la ficha: el nombre de las piezas, la web del proyecto y su licencia
// (data-model §3, §4 y §6 de H22).
const (
	nombreDeLasPiezas   = "kitlegal"
	paginaDelProyecto   = "https://kitlegal.es"
	licenciaDelProyecto = "EUPL-1.2"
)

// Las cuatro entradas de la extensión, que son todas las que lleva (H22
// FR-010; data-model §2).
const (
	rutaDelManifiesto        = "manifest.json"
	rutaDelIcono             = "icon.png"
	rutaDelServidor          = "server/kitlegal"
	rutaDelServidorDeWindows = "server/kitlegal.exe"
)

// Lo fijo del manifiesto (H22 FR-012; data-model §3).
const (
	// versionDelManifiesto es la del manifiesto de MCP Bundle con la que la
	// extensión se declara (H22 FR-017).
	versionDelManifiesto = "0.3"
	// tipoDelServidor dice que el servidor es un ejecutable, sin intérprete.
	tipoDelServidor = "binary"
	// carpetaDeLaExtension es lo que la app sustituye por la carpeta en la que
	// extrae la extensión: la orden del servidor cuelga de ella.
	carpetaDeLaExtension = "${__dirname}/"
	// Las dos plataformas de la extensión, como las nombra el manifiesto.
	plataformaMacOS   = "darwin"
	plataformaWindows = "win32"
	// La orden y el verbo con los que el binario sirve por la entrada y la
	// salida estándar: `kitlegal mcp serve` (H21).
	appletDelServidor = "mcp"
	verboDelServidor  = "serve"
)

// Lo fijo del plugin: dónde va su ficha y de qué carpeta del árbol que se le da
// cuelgan las skills, que es también su carpeta en el plugin (H22 FR-020;
// data-model §4).
const (
	rutaDeLaFichaDelPlugin = ".claude-plugin/plugin.json"
	carpetaDeSkills        = "skills"
	// rutaDelServidorDelPlugin es la de la extensión dentro del plugin, entera y
	// con su nombre: `mcpServers` la nombra con una ruta relativa, que es la
	// forma con la que la app de escritorio de Claude sincroniza el catálogo; con
	// la dirección de la release da un error (ADR 0035, «Prueba con un solo
	// plugin»).
	rutaDelServidorDelPlugin = "servers/" + nombreDeLaExtension
)

// Los dos modos de las entradas de los zips y el de los ficheros que el paso
// escribe (research.md D8 y D16 de H22).
const (
	modoDeEjecutable fs.FileMode = 0o755
	modoDeFichero    fs.FileMode = 0o644
)

// ladoDelIcono es lo que mide, en píxeles, cada lado del icono de la extensión
// (H22 FR-016).
const ladoDelIcono = 512

// fechaDeLasEntradas es la fecha de modificación de toda entrada de los dos
// zips: la más antigua que un zip sabe decir, y la misma en cada ejecución, que
// es lo que hace que no dependa de cuándo se empaqueta (research.md D8 de H22).
var fechaDeLasEntradas = time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC)

// piezasAEscribir es lo que el paso necesita para escribir la extensión y el
// plugin. Lo compone Ejecutar con lo que recibe de la línea de órdenes, la
// descripción corta del repositorio, el registro de producción y lo empotrado;
// los tests lo componen con unas herramientas y unas skills dadas, que es lo
// que hace comprobable que una herramienta o una skill nueva llega a las piezas
// sin tocar el paso (H22 US5).
type piezasAEscribir struct {
	// Version es la del manifiesto y la de plugin.json, tal cual llega: quien
	// llama la da sin `v` (H22 FR-013).
	Version string
	// MacOS y Windows son las rutas de los dos binarios, el universal de macOS y
	// el de Windows amd64. El paso los copia sin mirarlos: que sean lo que su
	// nombre dice lo comprueba el snapshot (H22 FR-011).
	MacOS, Windows string
	// Icono es la ruta del icono, un PNG de ladoDelIcono píxeles de lado.
	Icono string
	// Salida es la carpeta en la que se escriben las dos piezas. Tiene que
	// existir.
	Salida string
	// Descripcion es la descripción corta del manifiesto y de plugin.json, que
	// el paso comprueba antes de escribir nada. Es el único texto de la ficha
	// con un umbral, y por eso el único que llega como dato: la de producción
	// es siempre Descripcion.
	Descripcion string
	// Herramientas son las de `tools` del manifiesto, en su orden.
	Herramientas []app.HerramientaAnunciada
	// Skills es el árbol con las skills, con rutas skills/<skill>/…
	Skills fs.FS
}

// escribirPiezas escribe en la carpeta de salida la extensión de escritorio y
// el plugin (contracts/paso.md §2 y §3 de H22). Compone las dos en memoria
// antes de escribir ninguna: si falta una entrada, no deja nada.
func escribirPiezas(piezas piezasAEscribir) error {
	if err := comprobarDescripcion(piezas.Descripcion); err != nil {
		return err
	}

	extension, err := extensionDe(piezas)
	if err != nil {
		return err
	}

	plugin, err := pluginDe(piezas, extension)
	if err != nil {
		return err
	}

	if err := escribirEn(piezas.Salida, nombreDeLaExtension, extension); err != nil {
		return err
	}

	return escribirEn(piezas.Salida, nombreDelPlugin, plugin)
}

// extensionDe compone kitlegal.mcpb: un zip con el manifiesto, el icono y los
// dos binarios, en ese orden y con ninguna entrada más (H22 FR-010; data-model
// §2). El icono va byte a byte, tras comprobar que es el PNG que la ficha pide,
// y los dos binarios, sin mirarlos y marcados como ejecutables.
func extensionDe(piezas piezasAEscribir) ([]byte, error) {
	macos, err := os.ReadFile(piezas.MacOS)
	if err != nil {
		return nil, fmt.Errorf("falta el binario de macOS: %w", err)
	}

	windows, err := os.ReadFile(piezas.Windows)
	if err != nil {
		return nil, fmt.Errorf("falta el binario de Windows: %w", err)
	}

	icono, err := os.ReadFile(piezas.Icono)
	if err != nil {
		return nil, fmt.Errorf("falta el icono: %w", err)
	}

	if err := comprobarIcono(piezas.Icono, icono); err != nil {
		return nil, err
	}

	return comprimir([]ficheroDeZip{
		{ruta: rutaDelManifiesto, modo: modoDeFichero, escribir: documento(manifiestoDe(piezas))},
		{ruta: rutaDelIcono, modo: modoDeFichero, escribir: copia(icono)},
		{ruta: rutaDelServidor, modo: modoDeEjecutable, escribir: copia(macos)},
		{ruta: rutaDelServidorDeWindows, modo: modoDeEjecutable, escribir: copia(windows)},
	})
}

// comprobarIcono rechaza un icono que no es un PNG de ladoDelIcono píxeles de
// lado, diciendo lo que es: lo que el lector de PNG no pudo leer, o sus medidas
// (H22 FR-016, FR-066; research.md D20). Solo lee la cabecera: el icono se
// copia tal cual.
func comprobarIcono(ruta string, icono []byte) error {
	medidas, err := png.DecodeConfig(bytes.NewReader(icono))
	if err != nil {
		return fmt.Errorf("el icono %s no es un PNG de %d × %d px: %w", ruta, ladoDelIcono, ladoDelIcono, err)
	}

	if medidas.Width != ladoDelIcono || medidas.Height != ladoDelIcono {
		return fmt.Errorf("el icono %s no es un PNG de %d × %d px: mide %d × %d px",
			ruta, ladoDelIcono, ladoDelIcono, medidas.Width, medidas.Height)
	}

	return nil
}

// manifiesto es manifest.json, de la versión 0.3 del manifiesto de MCP Bundle:
// los campos de data-model §3 de H22, en ese orden, y ninguno más. No lleva
// `user_config`: no hay nada que configurar (H22 FR-012, FR-017).
type manifiesto struct {
	VersionDelManifiesto string         `json:"manifest_version"`
	Nombre               string         `json:"name"`
	NombreVisible        string         `json:"display_name"`
	Version              string         `json:"version"`
	Descripcion          string         `json:"description"`
	DescripcionLarga     string         `json:"long_description"`
	Autoria              autoria        `json:"author"`
	Pagina               string         `json:"homepage"`
	Licencia             string         `json:"license"`
	Icono                string         `json:"icon"`
	Servidor             servidor       `json:"server"`
	Herramientas         []herramienta  `json:"tools"`
	Compatibilidad       compatibilidad `json:"compatibility"`
}

// autoria es quien firma una pieza o el catálogo: solo su nombre.
type autoria struct {
	Nombre string `json:"name"`
}

// servidor es `server` del manifiesto: qué clase de servidor lleva la
// extensión, cuál es su ejecutable y cómo lo arranca la app.
type servidor struct {
	Tipo     string               `json:"type"`
	Entrada  string               `json:"entry_point"`
	Arranque configuracionDeLaApp `json:"mcp_config"`
}

// configuracionDeLaApp es `mcp_config`: la orden y los argumentos con los que la
// app arranca el servidor, y lo que cambia en cada plataforma.
type configuracionDeLaApp struct {
	Orden         string           `json:"command"`
	Argumentos    []string         `json:"args"`
	PorPlataforma ordenesDistintas `json:"platform_overrides"`
}

// ordenesDistintas es `platform_overrides`: solo Windows cambia la orden, por
// la de su ejecutable.
type ordenesDistintas struct {
	Windows ordenDePlataforma `json:"win32"`
}

// ordenDePlataforma es lo que una plataforma cambia de `mcp_config`.
type ordenDePlataforma struct {
	Orden string `json:"command"`
}

// herramienta es un elemento de `tools`: el nombre y la descripción con los que
// el servidor anuncia una herramienta.
type herramienta struct {
	Nombre      string `json:"name"`
	Descripcion string `json:"description"`
}

// compatibilidad es `compatibility`: las plataformas para las que la extensión
// lleva binario.
type compatibilidad struct {
	Plataformas []string `json:"platforms"`
}

// manifiestoDe es el manifiesto de esas piezas: lo fijo, los textos, la versión
// tal cual llega y una herramienta por cada una de las que se le dan, en su
// orden (H22 FR-012 a FR-015).
func manifiestoDe(piezas piezasAEscribir) manifiesto {
	herramientas := make([]herramienta, 0, len(piezas.Herramientas))
	for _, anunciada := range piezas.Herramientas {
		herramientas = append(herramientas, herramienta{Nombre: anunciada.Nombre, Descripcion: anunciada.Descripcion})
	}

	return manifiesto{
		VersionDelManifiesto: versionDelManifiesto,
		Nombre:               nombreDeLasPiezas,
		NombreVisible:        NombreVisible,
		Version:              piezas.Version,
		Descripcion:          piezas.Descripcion,
		DescripcionLarga:     DescripcionLarga,
		Autoria:              autoria{Nombre: Autoria},
		Pagina:               paginaDelProyecto,
		Licencia:             licenciaDelProyecto,
		Icono:                rutaDelIcono,
		Servidor: servidor{
			Tipo:    tipoDelServidor,
			Entrada: rutaDelServidor,
			Arranque: configuracionDeLaApp{
				Orden:      carpetaDeLaExtension + rutaDelServidor,
				Argumentos: []string{appletDelServidor, verboDelServidor},
				PorPlataforma: ordenesDistintas{
					Windows: ordenDePlataforma{Orden: carpetaDeLaExtension + rutaDelServidorDeWindows},
				},
			},
		},
		Herramientas:   herramientas,
		Compatibilidad: compatibilidad{Plataformas: []string{plataformaMacOS, plataformaWindows}},
	}
}

// fichaDelPlugin es plugin.json: los campos de data-model §4 de H22, en ese
// orden, y `mcpServers`, con la ruta de la extensión dentro del plugin. H22 lo
// dejó sin servidor, porque las herramientas de un plugin solo llegaban a las
// conversaciones con una carpeta elegida; desde que llegan a todas, el plugin
// lleva las skills y el servidor, y se instala una sola pieza (ADR 0035,
// «Prueba con un solo plugin»).
type fichaDelPlugin struct {
	Nombre      string  `json:"name"`
	Version     string  `json:"version"`
	Descripcion string  `json:"description"`
	Autoria     autoria `json:"author"`
	Pagina      string  `json:"homepage"`
	Licencia    string  `json:"license"`
	Servidores  string  `json:"mcpServers"`
}

// pluginDe compone kitlegal-plugin.zip: plugin.json, la extensión que se le da,
// byte a byte, y, detrás, cada fichero de la carpeta de skills del árbol, con su
// ruta y en el orden en que fs.WalkDir lo da, y nada más: ni entradas de
// directorio, ni `bin/`, ni `.mcp.json` (H22 FR-020; data-model §4). Todos van
// con el modo de un fichero, sea cual sea el que el árbol diga de ellos: los
// binarios van dentro de la extensión, con el suyo.
func pluginDe(piezas piezasAEscribir, extension []byte) ([]byte, error) {
	ficheros := []ficheroDeZip{
		{
			ruta: rutaDeLaFichaDelPlugin,
			modo: modoDeFichero,
			escribir: documento(fichaDelPlugin{
				Nombre:      nombreDeLasPiezas,
				Version:     piezas.Version,
				Descripcion: piezas.Descripcion,
				Autoria:     autoria{Nombre: Autoria},
				Pagina:      paginaDelProyecto,
				Licencia:    licenciaDelProyecto,
				Servidores:  fuenteRelativa + rutaDelServidorDelPlugin,
			}),
		},
		{ruta: rutaDelServidorDelPlugin, modo: modoDeFichero, escribir: copia(extension)},
	}

	err := fs.WalkDir(piezas.Skills, carpetaDeSkills, func(ruta string, entrada fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entrada.IsDir() {
			return nil
		}

		contenido, err := fs.ReadFile(piezas.Skills, ruta)
		if err != nil {
			return err
		}

		ficheros = append(ficheros, ficheroDeZip{ruta: ruta, modo: modoDeFichero, escribir: copia(contenido)})

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("las skills no se pueden leer: %w", err)
	}

	return comprimir(ficheros)
}

// ficheroDeZip es una entrada de un zip: su ruta, su modo y cómo se escribe su
// contenido en ella.
type ficheroDeZip struct {
	ruta     string
	modo     fs.FileMode
	escribir func(entrada io.Writer) error
}

// copia es el contenido de una entrada que lleva esos bytes tal cual.
func copia(contenido []byte) func(io.Writer) error {
	return func(entrada io.Writer) error {
		_, err := entrada.Write(contenido)

		return err
	}
}

// documento es el contenido de una entrada que lleva ese valor como documento
// JSON, con la forma de escribirDocumento.
func documento(valor any) func(io.Writer) error {
	return func(entrada io.Writer) error {
		return escribirDocumento(entrada, valor)
	}
}

// escribirDocumento escribe un valor como los tres documentos JSON del paso: los
// campos en el orden de su tipo, sangría de dos espacios, `<`, `>` y `&` sin
// escapar —nadie los lee como HTML, y una descripción los lleva tal cual— y
// salto de línea final (contracts/paso.md §5 de H22).
func escribirDocumento(destino io.Writer, valor any) error {
	codificador := json.NewEncoder(destino)
	codificador.SetEscapeHTML(false)
	codificador.SetIndent("", "  ")

	return codificador.Encode(valor)
}

// comprimir compone un zip reproducible con esos ficheros, en su orden: sin
// entradas de directorio, con Deflate de archive/zip, la fecha fija y el modo
// de cada uno, y sin comentario. Dos llamadas con los mismos ficheros dan los
// mismos bytes (H22 FR-004; research.md D8).
//
// El zip se compone en memoria, así que una entrada solo falla si no cabe en el
// formato —una ruta de más de 65 535 bytes— o si su contenido no se puede
// escribir, y el error la nombra.
func comprimir(ficheros []ficheroDeZip) ([]byte, error) {
	var comprimido bytes.Buffer

	escritor := zip.NewWriter(&comprimido)

	for _, fichero := range ficheros {
		if err := anadir(escritor, fichero); err != nil {
			return nil, fmt.Errorf("la entrada %s no cabe en el zip: %w", fichero.ruta, err)
		}
	}

	if err := escritor.Close(); err != nil {
		return nil, fmt.Errorf("el zip no se puede cerrar: %w", err)
	}

	return comprimido.Bytes(), nil
}

// anadir escribe una entrada en el zip: su cabecera, con lo que hace que el
// zip no dependa de cuándo ni dónde se compone, y su contenido.
func anadir(escritor *zip.Writer, fichero ficheroDeZip) error {
	cabecera := &zip.FileHeader{Name: fichero.ruta, Method: zip.Deflate, Modified: fechaDeLasEntradas}
	cabecera.SetMode(fichero.modo)

	entrada, err := escritor.CreateHeader(cabecera)
	if err != nil {
		return err
	}

	return fichero.escribir(entrada)
}

// escribirEn escribe un fichero con ese nombre y ese contenido en la carpeta,
// que tiene que existir, con el modo de un fichero y sin salir de ella: abre la
// carpeta como raíz y lo escribe por su nombre (research.md D16 de H22). Si no
// se puede, por la carpeta o por el fichero, el error nombra la ruta del
// fichero y la causa.
func escribirEn(carpeta, nombre string, contenido []byte) (err error) {
	raiz, err := os.OpenRoot(carpeta)
	if err != nil {
		return falloDeEscritura(filepath.Join(carpeta, nombre), err)
	}

	defer func() { err = errors.Join(err, raiz.Close()) }()

	if err := raiz.WriteFile(nombre, contenido, modoDeFichero); err != nil {
		return falloDeEscritura(filepath.Join(carpeta, nombre), err)
	}

	return nil
}

// falloDeEscritura es el error de un fichero que no se puede escribir: «no se
// puede escribir <ruta>: <causa>» (contracts/paso.md §1 de H22). La causa es la
// del sistema, sin la operación ni la ruta con que la envuelve el paquete os,
// para no nombrar la ruta dos veces.
func falloDeEscritura(ruta string, err error) error {
	var enRuta *fs.PathError
	if errors.As(err, &enRuta) {
		err = enRuta.Err
	}

	return fmt.Errorf("no se puede escribir %s: %w", ruta, err)
}
