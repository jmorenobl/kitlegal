package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/mcp"
)

// errSinEntrada es el defecto de composición de unas dependencias sin entrada:
// el servidor no tiene de dónde leer ningún mensaje, y no es algo que quien lo
// arranca pueda corregir. Sale como inesperado.
var errSinEntrada = errors.New("mcp: el applet se compuso sin entrada y no puede atender ningún mensaje")

// DependenciasDeMCP es lo que el applet mcp recibe de la raíz de composición
// (data-model §5 de H21). El binario distribuido da las de
// DependenciasDeMCPDelSistema; las pruebas, el extremo de lectura de una
// tubería.
type DependenciasDeMCP struct {
	// Entrada es por donde llegan los mensajes del cliente: la entrada estándar
	// del proceso. La salida no va aquí: es la del presentador del kernel, que
	// es el único que nombra la salida estándar (research.md D4 de H21).
	Entrada io.Reader
	// Version es la del binario, que el servidor da en `serverInfo`.
	Version string
}

// DependenciasDeMCPDelSistema son las del binario distribuido: la entrada
// estándar del proceso y la versión del binario.
func DependenciasDeMCPDelSistema(version string) DependenciasDeMCP {
	return DependenciasDeMCP{Entrada: os.Stdin, Version: version}
}

// AppletMCP es el applet que sirve las herramientas del binario por el
// protocolo MCP, con un solo verbo, `serve`, sin argumentos propios y sin verbo
// por omisión (H21 FR-001; contracts/servidor-mcp.md §1 de H21). Las
// herramientas no son suyas: las saca del registro con el que el kernel lo
// llama, una por cada verbo de los demás applets (herramientas.go).
//
// Componerlo no lee nada ni abre nada: tampoco la entrada, que solo lee el
// servidor cuando sirve.
func AppletMCP(dependencias DependenciasDeMCP) Applet {
	return appletMCP{dependencias: dependencias}
}

// appletMCP declara lo que declara un applet y nada más: su nombre, su línea de
// ayuda y su catálogo de verbos, que llevan consigo las dependencias.
type appletMCP struct {
	dependencias DependenciasDeMCP
}

func (appletMCP) Nombre() string { return "mcp" }

func (appletMCP) Descripcion() string {
	return "Sirve las herramientas de kitlegal a un agente por el protocolo MCP."
}

// Verbos es el único del contrato. No declara Salida: al servir no emite ningún
// sobre, y --describe deja su data sin restringir (research.md D26 de H21).
func (a appletMCP) Verbos() []Verbo {
	return []Verbo{{
		Nombre: "serve",
		Descripcion: "Atiende el protocolo MCP por la entrada y la salida estándar hasta que la entrada" +
			" se cierra.",
		Argumentos: func() Argumentos { return &argumentosDeServe{dependencias: a.dependencias} },
	}}
}

// argumentosDeServe son los del verbo serve, que no declara ninguno propio: las
// dependencias van en un campo sin exportar, que la gramática no ve.
type argumentosDeServe struct {
	dependencias DependenciasDeMCP
}

var (
	_ Argumentos = (*argumentosDeServe)(nil)
	_ servidor   = (*argumentosDeServe)(nil)
)

// Ejecutar es el camino de `mcp serve --dry-run`, el único por el que el kernel
// llega aquí: valida y no sirve. No atiende ningún mensaje ni lee de la
// entrada, y lo que se ve de la invocación es la descripción del kernel (H21
// FR-022; research.md D4 de H21).
func (a *argumentosDeServe) Ejecutar(
	_ context.Context, ec schema.Contexto, _ *slog.Logger,
) (schema.Resultado, error) {
	return schema.Resultado{}, a.validar(ec)
}

// servir atiende el protocolo MCP por la entrada de las dependencias y por la
// salida estándar del presentador, con las herramientas del registro, hasta
// que la entrada se cierra (H21 FR-001, FR-024). El contexto no tiene plazo: el
// de --timeout es el de cada llamada.
func (a *argumentosDeServe) servir(
	ctx context.Context,
	ec schema.Contexto,
	registrador *slog.Logger,
	p cli.Presentador,
	registro *Registro,
) error {
	if err := a.validar(ec); err != nil {
		return err
	}

	herramientas, err := herramientasDe(registro, ec, registrador, p.Error())
	if err != nil {
		return err
	}

	err = mcp.Servir(ctx, mcp.Servicio{
		Entrada:      a.dependencias.Entrada,
		Salida:       p.Salida(),
		Version:      a.dependencias.Version,
		Registrador:  registrador,
		Herramientas: herramientas,
	})
	if err != nil {
		return fmt.Errorf("mcp: el servidor ha terminado sin que se cierre su entrada: %w", err)
	}

	return nil
}

// validar comprueba lo que `mcp serve` exige antes de atender nada, sirva o no:
// que no se le dé --asunto, porque el servidor no expone nada del asunto —es un
// error de argumentos, con el mensaje de contracts/servidor-mcp.md §1 de H21
// (FR-021)—, y que tenga de dónde leer.
func (a *argumentosDeServe) validar(ec schema.Contexto) error {
	if ec.Asunto != "" {
		return fmt.Errorf("%w: mcp serve no admite --asunto: el servidor no expone nada del asunto",
			cli.ErrArgumentos)
	}

	if a.dependencias.Entrada == nil {
		return errSinEntrada
	}

	return nil
}
