//go:build grabacion

package bocm

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/httpx"
)

// Lo que fija la grabación de las disposiciones de fiestas locales
// (contracts/datos-de-plazos.md §4.1; research D15 de H9; fila
// bocm.fiestas-locales de docs/SOURCES.md).
const (
	// nombreDeLaFuente es el nombre lógico de la fuente: el de su fila de
	// docs/SOURCES.md, el que declara el manifiesto y el del directorio de sus
	// grabaciones bajo la raíz de evidencias.
	nombreDeLaFuente = "bocm.fiestas-locales"

	// ficheroDelManifiesto es el manifiesto de grabación del paquete, el único
	// que scripts/workflow/grabar-datos.sh graba (research V2 de H9).
	ficheroDelManifiesto = "testdata/grabaciones.json"

	// variableDeEvidencias es la variable de entorno con la raíz de grabación,
	// evidencias/<hito>/, que pone grabar-datos.sh: lo grabado es material de
	// origen de data/festivos/, no una respuesta que reproduzca un test, y por
	// eso no va a testdata/ (docs/WORKFLOW.md, «Datos externos»).
	variableDeEvidencias = "KITLEGAL_EVIDENCIAS"

	// valorQueGraba es el único valor de la variable de grabación que la
	// enciende en internal/httpx.
	valorQueGraba = "1"

	// intervaloEntrePeticiones es la separación mínima entre dos peticiones al
	// BOCM: el Crawl-delay: 10 de su robots.txt, el ritmo de su fila de
	// docs/SOURCES.md.
	intervaloEntrePeticiones = 10 * time.Second

	// metodoDeLaGrabacion es el único método con el que se pide: solo GET.
	metodoDeLaGrabacion = "GET"

	// prefijoDelIdentificador es el principio de todo identificador de una
	// disposición, BOCM-AAAAMMDD-N, al que sigue la fecha de su publicación.
	prefijoDelIdentificador = "BOCM-"

	// formatoDeLaFechaDelIdentificador es la fecha de publicación tal como va
	// en el identificador: AAAAMMDD.
	formatoDeLaFechaDelIdentificador = "20060102"

	// baseDeLasDisposiciones es el principio de la dirección del XML de toda
	// disposición, a la que siguen la fecha de publicación como AAAA/MM/DD, el
	// identificador y .xml; /boletin/ queda abierto en el robots.txt de la
	// fuente (su fila de docs/SOURCES.md; evidencias/adr-0039/LEEME.md).
	baseDeLasDisposiciones = "https://www.bocm.es/boletin/CM_Orden_BOCM/"

	// formatoDeLaFechaDeLaDireccion es la fecha de publicación tal como va en
	// la dirección: AAAA/MM/DD.
	formatoDeLaFechaDeLaDireccion = "2006/01/02"

	// extensionDeLaDisposicion es la del documento que se pide: el XML.
	extensionDeLaDisposicion = ".xml"

	// estadoQueSeGraba es el único estado HTTP que vale como disposición
	// grabada: otro no es el documento y deja la grabación incompleta.
	estadoQueSeGraba = 200
)

// manifiestoDeGrabacion es testdata/grabaciones.json: la fuente y las
// disposiciones que se graban (data-model §2 de H9).
type manifiestoDeGrabacion struct {
	Fuente        string        `json:"fuente"`
	Disposiciones []disposicion `json:"disposiciones"`
}

// disposicion es una resolución que se graba: su identificador en el BOCM, el
// año cuyas fiestas declara o corrige y la dirección de su XML.
type disposicion struct {
	ID   string `json:"id"`
	Anio int    `json:"anio"`
	URL  string `json:"url"`
}

// TestGrabarFiestasLocales graba contra la fuente real, con internal/httpx,
// cada disposición del manifiesto: solo GET, a 10 s entre peticiones, que es
// el ritmo de su fila de docs/SOURCES.md, y bajo
// <raíz de evidencias>/bocm.fiestas-locales/ (FR-020, FR-027 de H9). Exige que
// cada respuesta sea el documento (estado 200): una grabación de otra cosa no
// es la disposición y no vale como material de origen. No lee el texto: lo
// grabado lo lee la derivación de data/festivos/ y lo revisa una persona.
//
// Toca la red. Solo lo ejecuta el paso grabar_datos del workflow
// (scripts/workflow/grabar-datos.sh), sin modelo, tras el commit de la tarea
// del manifiesto, con la etiqueta grabacion, la variable de grabación y la raíz
// de evidencias; sin cualquiera de las dos variables, o con un manifiesto que
// no se puede grabar entero, falla antes de pedir nada.
func TestGrabarFiestasLocales(t *testing.T) {
	t.Parallel()

	if valor := os.Getenv(httpx.VariableGrabacion); valor != valorQueGraba {
		t.Fatalf("TestGrabarFiestasLocales pide a la fuente real y graba lo que responde: solo se ejecuta con %s=%s, "+
			"desde scripts/workflow/grabar-datos.sh (la variable vale %q)", httpx.VariableGrabacion, valorQueGraba, valor)
	}

	raiz := os.Getenv(variableDeEvidencias)
	if raiz == "" {
		t.Fatalf("TestGrabarFiestasLocales graba en la raíz de evidencias que da %s (evidencias/<hito>/), "+
			"que pone scripts/workflow/grabar-datos.sh y que está vacía", variableDeEvidencias)
	}

	peticiones, err := peticionesDeGrabacion(ficheroDelManifiesto)
	require.NoError(t, err)

	cliente, err := httpx.New(
		httpx.ConFuente(nombreDeLaFuente),
		httpx.ConRaizDeGrabacion(raiz),
		httpx.ConIntervalo(intervaloEntrePeticiones),
	)
	require.NoError(t, err)

	for _, peticion := range peticiones {
		respuesta, err := cliente.Pedir(t.Context(), schema.Contexto{}, peticion)
		require.NoErrorf(t, err, "la grabación queda incompleta en %s %s", peticion.Metodo, peticion.URL)
		require.Equalf(t, estadoQueSeGraba, respuesta.Estado,
			"%s %s no ha respondido con el documento: la grabación queda incompleta", peticion.Metodo, peticion.URL)

		t.Logf("grabado: %s %s", peticion.Metodo, peticion.URL)
	}
}

// peticionesDeGrabacion lee el manifiesto con leerManifiesto y construye todas
// sus peticiones antes de que se pida ninguna, de modo que un manifiesto que no
// se puede grabar entero no graba nada: cada disposición lleva un identificador
// BOCM-AAAAMMDD-N, un año y la dirección del XML que su identificador determina.
func peticionesDeGrabacion(ruta string) ([]httpx.Peticion, error) {
	manifiesto, err := leerManifiesto(ruta)
	if err != nil {
		return nil, err
	}

	peticiones := make([]httpx.Peticion, 0, len(manifiesto.Disposiciones))

	for indice, entrada := range manifiesto.Disposiciones {
		if err := entrada.comprobar(); err != nil {
			return nil, fmt.Errorf("%s, disposición %d (%q): %w", ruta, indice+1, entrada.ID, err)
		}

		peticiones = append(peticiones, httpx.Peticion{Metodo: metodoDeLaGrabacion, URL: entrada.URL})
	}

	return peticiones, nil
}

// leerManifiesto lee el manifiesto como un único documento JSON con la forma de
// manifiestoDeGrabacion, sin miembros que no sean campos suyos ni claves
// repetidas (encoding/json/v2 con RejectUnknownMembers), de la fuente de este
// paquete y con al menos una disposición.
func leerManifiesto(ruta string) (manifiestoDeGrabacion, error) {
	contenido, err := os.ReadFile(filepath.Clean(ruta))
	if err != nil {
		return manifiestoDeGrabacion{}, fmt.Errorf("el manifiesto de grabación no se puede leer: %w", err)
	}

	var manifiesto manifiestoDeGrabacion
	if err := json.Unmarshal(contenido, &manifiesto, json.RejectUnknownMembers(true)); err != nil {
		return manifiestoDeGrabacion{}, fmt.Errorf("%s no es un manifiesto de grabación válido: %w", ruta, err)
	}

	if manifiesto.Fuente != nombreDeLaFuente {
		return manifiestoDeGrabacion{}, fmt.Errorf("%s es de la fuente %q y este paquete es %q",
			ruta, manifiesto.Fuente, nombreDeLaFuente)
	}

	if len(manifiesto.Disposiciones) == 0 {
		return manifiestoDeGrabacion{}, fmt.Errorf("%s no lista ninguna disposición", ruta)
	}

	return manifiesto, nil
}

// comprobar exige el año y la dirección del XML que determina el identificador,
// que es la que lee la derivación: la fecha de publicación del identificador es
// la de la dirección, comprobada contra las direcciones de las siete
// disposiciones del cotejo del ADR 0039 (evidencias/adr-0039/LEEME.md).
func (d disposicion) comprobar() error {
	esperada, err := direccionDeLaDisposicion(d.ID)
	if err != nil {
		return err
	}

	if d.Anio <= 0 {
		return errors.New("sin año")
	}

	if d.URL != esperada {
		return fmt.Errorf("la dirección es %q y tiene que ser %q", d.URL, esperada)
	}

	return nil
}

// direccionDeLaDisposicion es la dirección del XML de una disposición, que se
// compone con la fecha de publicación de su identificador BOCM-AAAAMMDD-N: un
// identificador sin esa forma, sin número o cuya fecha no es un día del
// calendario es un error que lo nombra.
func direccionDeLaDisposicion(id string) (string, error) {
	resto, conPrefijo := strings.CutPrefix(id, prefijoDelIdentificador)
	fecha, numero, conNumero := strings.Cut(resto, "-")

	if !conPrefijo || !conNumero || !sonDigitos(fecha) || !sonDigitos(numero) ||
		len(fecha) != len(formatoDeLaFechaDelIdentificador) {
		return "", fmt.Errorf("el identificador %q no tiene la forma %sAAAAMMDD-N", id, prefijoDelIdentificador)
	}

	publicacion, err := time.Parse(formatoDeLaFechaDelIdentificador, fecha)
	if err != nil {
		return "", fmt.Errorf("la fecha del identificador %q no es un día del calendario: %w", id, err)
	}

	return baseDeLasDisposiciones + publicacion.Format(formatoDeLaFechaDeLaDireccion) + "/" + id +
		extensionDeLaDisposicion, nil
}

// sonDigitos dice si el texto no está vacío y es todo dígitos ASCII.
func sonDigitos(texto string) bool {
	if texto == "" {
		return false
	}

	for _, caracter := range texto {
		if caracter < '0' || caracter > '9' {
			return false
		}
	}

	return true
}
