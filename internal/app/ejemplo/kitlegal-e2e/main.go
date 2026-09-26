// Command kitlegal-e2e es el binario contra el que se ejecuta el test de
// extremo a extremo: el **kernel real** —el mismo internal/app que enlaza el
// binario que se publica— con los applets de ejemplo y los applets boe, skills y
// territorio. Lo único que cambia entre este binario y el distribuido es la
// composición: qué applets se registran, de dónde responde boe, que aquí es la
// reproducción de sus grabaciones y nunca la red, y, si la construcción lo
// elige, un creador de enlaces de skills que siempre falla (FR-009, FR-024,
// FR-114, contracts/registro-y-describe.md §3; contrato puerto-y-applet §5 de
// H4; contracts/arnes-e2e.md §2 de H19).
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
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/jmorenobl/kitlegal/internal/app"
	"github.com/jmorenobl/kitlegal/internal/app/ejemplo"
	"github.com/jmorenobl/kitlegal/internal/httpx"
	"github.com/jmorenobl/kitlegal/internal/source/boe"
)

// Datos de construcción. El binario de e2e de desarrollo se compila sin
// -ldflags, así que los valores que se ven son estos; el arnés construye además
// este mismo paquete con -X sobre version (v0.1.0 y v0.2.0) y sobre enlazador
// (contracts/arnes-e2e.md §2). El contrato de `version` que ejerce el binario
// distribuido lo comprueban los tests de cmd/kitlegal (D16).
var (
	version = "dev"
	commit  = "none"
	fecha   = "unknown"

	// enlazador elige el creador de enlaces del applet skills: sin valor, el del
	// sistema, como el binario distribuido; enlazadorQueFalla, uno que siempre
	// falla, para ejercer el recurso de copia (FR-024, SC-012). Se fija al
	// construir, con -X, y nunca por el entorno: este binario no lee del entorno
	// nada que decida su comportamiento. Es una variable de cadena, y no una
	// constante, porque -X solo fija variables de cadena (research.md D24).
	enlazador string
)

// enlazadorQueFalla es el valor de enlazador que elige el creador de enlaces
// que siempre falla.
const enlazadorQueFalla = "falla"

// errEnlazadorDesconocido es un valor de enlazador que no elige ningún creador
// de enlaces: un defecto de quien construyó el binario, que se nombra en lugar
// de caer en silencio al creador del sistema.
var errEnlazadorDesconocido = errors.New("kitlegal-e2e: el binario se construyó con un creador de enlaces desconocido")

// errSinEnlaces es el fallo de enlazadorFallido al crear un enlace.
var errSinEnlaces = errors.New("kitlegal-e2e: este binario no crea enlaces simbólicos")

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
// boe sobre la reproducción, skills con las mismas dependencias del sistema que
// el binario distribuido —la versión de este binario, lo empotrado y el creador
// de enlaces de internal/disco, salvo que la construcción eligiera el que
// falla— y territorio con los mismos ficheros embebidos, que no dependen del
// entorno (contrato del applet territorio §7). Construirlo no pide nada ni abre
// nada. Un registro que no se construye es un defecto de quien escribió un
// applet o esta composición, y app.Arrancar lo convierte en el fallo inesperado
// antes de atender ninguna invocación: nunca en un código de salida de usuario
// ni en un pánico (FR-008; research.md D16 de H4). Recibe la versión del binario
// como el registro de producción, y como allí la lleva a skills y compone con
// las mismas dependencias el aviso de versión, que registra (research.md D4 y
// D5 de H19).
func registroDeE2E(version string) (*app.Registro, error) {
	skills, err := dependenciasDeSkills(version, enlazador)
	if err != nil {
		return nil, err
	}

	registro, err := ejemplo.Registro()
	if err != nil {
		return nil, err
	}

	fuentes, err := app.FuentesEmbebidas()
	if err != nil {
		return nil, err
	}

	applets := []app.Applet{
		app.AppletBoe(dependenciasDeReproduccion()),
		app.AppletSkills(skills),
		app.AppletTerritorio(fuentes),
	}

	for _, applet := range applets {
		if err := registro.Registrar(applet); err != nil {
			return nil, err
		}
	}

	registro.Avisar(app.AvisoDeVersion(skills))

	return registro, nil
}

// dependenciasDeSkills son las del applet skills con la versión del binario y
// el creador de enlaces que elige la construcción: sin elección, las del
// sistema de app.DependenciasDeSkillsDelSistema tal cual; con
// enlazadorQueFalla, las mismas con el Enlazador sustituido por enlazadorFallido
// (FR-024; contracts/arnes-e2e.md §2). El applet las recibe hechas y no sabe de
// dónde sale su Enlazador.
func dependenciasDeSkills(version, eleccion string) (app.DependenciasDeSkills, error) {
	dependencias := app.DependenciasDeSkillsDelSistema(version)

	switch eleccion {
	case "":
		return dependencias, nil
	case enlazadorQueFalla:
		dependencias.Enlazador = enlazadorFallido{}

		return dependencias, nil
	default:
		return app.DependenciasDeSkills{}, fmt.Errorf("%w: %q", errEnlazadorDesconocido, eleccion)
	}
}

// enlazadorFallido es el creador de enlaces que siempre falla: ningún
// directorio admite enlaces y crear uno es siempre un error, sin tocar el disco
// (contracts/arnes-e2e.md §2). Con él, install deja como copia cada entrada de
// host, como en un sistema de ficheros sin enlaces (FR-024, SC-012). Es un tipo
// de este package main, y no de internal/disco, porque solo existe para
// probarse: el binario distribuido no lo enlaza (ADR 0010).
type enlazadorFallido struct{}

// Disponible dice que en ningún directorio se puede crear un enlace, sin
// examinarlo.
func (enlazadorFallido) Disponible(string) (bool, error) {
	return false, nil
}

// Enlazar falla siempre, sin crear nada, nombrando el enlace que no crea.
func (enlazadorFallido) Enlazar(destino, ruta string) error {
	return fmt.Errorf("enlazar %s -> %s: %w", ruta, destino, errSinEnlaces)
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
