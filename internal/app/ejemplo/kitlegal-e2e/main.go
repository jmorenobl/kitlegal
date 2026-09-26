// Command kitlegal-e2e es el binario contra el que se ejecuta el test de
// extremo a extremo: el **kernel real** —el mismo internal/app que enlaza el
// binario que se publica— con los applets de ejemplo y los applets boe y
// territorio. Lo único que cambia entre este binario y el distribuido es la
// composición: qué applets se registran y de dónde responde boe, que aquí es la
// reproducción de sus grabaciones y nunca la red (FR-009, FR-114,
// contracts/registro-y-describe.md §3; contrato puerto-y-applet §5 de H4).
//
// Es la segunda —y última— raíz de composición del proyecto, y por eso es uno de
// los dos únicos sitios del árbol donde se nombran os.Exit, os.Stdout y
// os.Stderr: no hay forma de que un `package main` propague un código de salida
// sin lo primero —retornar de main sale siempre con 0— ni de que inyecte los
// descriptores sin nombrarlos. La excepción del lint se acota a este directorio
// y no alcanza al paquete padre de applets de ejemplo (research.md D18).
//
// Es un `package main` bajo internal/app/ejemplo y no bajo cmd/: no es un
// binario que se distribuya, y `go install ./cmd/...` o goreleaser no lo
// alcanzan. Que el binario distribuido no lo enlace lo vigilan depguard y el
// test de arquitectura (docs/ADR/0010-applets-de-ejemplo-fuera-de-testdata.md).
package main

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/jmorenobl/kitlegal/internal/app"
	"github.com/jmorenobl/kitlegal/internal/app/ejemplo"
	"github.com/jmorenobl/kitlegal/internal/httpx"
	"github.com/jmorenobl/kitlegal/internal/source/boe"
)

// Datos de construcción. El e2e compila este paquete sin -ldflags, así que los
// valores que se ven son estos; el contrato de `version` que ejerce el binario
// distribuido lo comprueban los tests de cmd/kitlegal (D16).
var (
	version = "dev"
	commit  = "none"
	fecha   = "unknown"
)

// directorioDeReproduccion es la carpeta de la que boe sirve sus grabaciones,
// con una subcarpeta por fuente, relativa al directorio desde el que se invoca
// el binario: el test de e2e deja allí, en el directorio de trabajo de cada
// guion, la copia de las grabaciones. Es una ruta escrita aquí y no una variable
// de entorno a propósito: este binario no lee del entorno nada que decida de
// dónde responde (research.md D13 de H4).
const directorioDeReproduccion = "reproduccion"

// main es, línea por línea, el mismo que el del binario distribuido salvo el
// registro que inyecta: app.Arrancar construye el registro, atiende la
// invocación y devuelve el código de salida sin terminar nunca el proceso, y
// acabar con él es lo único que este punto de entrada hace
// (contracts/reglas-de-arquitectura.md §2, R4).
func main() {
	os.Exit(app.Arrancar(
		os.Args, registroDeE2E, os.Stdout, os.Stderr, version, commit, fecha,
	))
}

// registroDeE2E construye el registro de este binario: los applets de ejemplo,
// boe sobre la reproducción y territorio con los mismos ficheros embebidos que el
// binario distribuido, que no dependen del entorno (contrato del applet
// territorio §7). Construirlo no pide nada ni abre nada. Un registro que no se
// construye es un defecto de quien escribió un applet o esta composición, y
// app.Arrancar lo convierte en el fallo inesperado antes de atender ninguna
// invocación: nunca en un código de salida de usuario ni en un pánico (FR-008;
// research.md D16 de H4). Recibe la versión del binario como el registro de
// producción (research.md D4 de H19), y como allí, ninguno de los applets que
// registra hoy la necesita, así que el parámetro va en blanco.
func registroDeE2E(_ string) (*app.Registro, error) {
	registro, err := ejemplo.Registro()
	if err != nil {
		return nil, err
	}

	fuentes, err := app.FuentesEmbebidas()
	if err != nil {
		return nil, err
	}

	for _, applet := range []app.Applet{app.AppletBoe(dependenciasDeReproduccion()), app.AppletTerritorio(fuentes)} {
		if err := registro.Registrar(applet); err != nil {
			return nil, err
		}
	}

	return registro, nil
}

// dependenciasDeReproduccion son las de boe en este binario: el cliente de
// reproducción sobre la carpeta de la fuente, que no abre ninguna conexión, con
// el nombre de la fuente y el registrador que el kernel entrega al applet; y la
// caché de siempre, la de KITLEGAL_CACHE_DIR o la de la cuenta, que el e2e fija
// en el directorio de trabajo de cada guion. No lleva intervalo entre
// peticiones: la reproducción no tiene sitio al que esperar y lo rechaza
// (contrato puerto-y-applet §5 de H4).
func dependenciasDeReproduccion() app.DependenciasDeBoe {
	return app.DependenciasDeBoe{
		Cliente: func(registrador *slog.Logger) (*httpx.Cliente, error) {
			return httpx.Replay(
				filepath.Join(directorioDeReproduccion, boe.NombreDeLaFuente),
				httpx.ConFuente(boe.NombreDeLaFuente),
				httpx.ConRegistrador(registrador),
			)
		},
	}
}
