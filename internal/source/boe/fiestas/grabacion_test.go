//go:build grabacion

package fiestas

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/httpx"
	"github.com/jmorenobl/kitlegal/internal/source/boe"
)

// Lo que fija la grabación de las resoluciones de fiestas laborales
// (contracts/datos-de-plazos.md §4.1; research D15 de H9; fila
// boe.fiestas-laborales de docs/SOURCES.md).
const (
	// nombreDeLaFuente es el nombre lógico de la fuente: el de su fila de
	// docs/SOURCES.md, el que declara el manifiesto y el del directorio de sus
	// grabaciones bajo la raíz de evidencias.
	nombreDeLaFuente = "boe.fiestas-laborales"

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

	// metodoDeLaGrabacion es el único método con el que se pide: solo GET.
	metodoDeLaGrabacion = "GET"

	// prefijoDeLaDireccion es la dirección del documento del diario en el BOE,
	// a la que sigue el identificador de la resolución; /diario_boe/txt.php
	// queda abierto en el robots.txt de la fuente (su fila de docs/SOURCES.md).
	prefijoDeLaDireccion = "https://www.boe.es/diario_boe/txt.php?id="

	// estadoQueSeGraba es el único estado HTTP que vale como resolución
	// grabada: otro no es el documento y deja la grabación incompleta.
	estadoQueSeGraba = 200
)

// manifiestoDeGrabacion es testdata/grabaciones.json: la fuente y las
// disposiciones que se graban (data-model §2 de H9).
type manifiestoDeGrabacion struct {
	Fuente        string        `json:"fuente"`
	Disposiciones []disposicion `json:"disposiciones"`
}

// disposicion es una resolución que se graba: su identificador en el BOE, el
// año cuyas fiestas declara y la dirección de su documento.
type disposicion struct {
	ID   string `json:"id"`
	Anio int    `json:"anio"`
	URL  string `json:"url"`
}

// TestGrabarFiestasLaborales graba contra la fuente real, con internal/httpx,
// cada resolución del manifiesto: solo GET, al ritmo de
// boe.legislacion-consolidada, que es el de su fila de docs/SOURCES.md, y bajo
// <raíz de evidencias>/boe.fiestas-laborales/ (FR-020, FR-027 de H9). Exige que
// cada respuesta sea el documento (estado 200): una grabación de otra cosa no
// es la resolución y no vale como material de origen. No lee el anexo: lo
// grabado lo lee la derivación de data/festivos/ y lo revisa una persona.
//
// Toca la red. Solo lo ejecuta el paso grabar_datos del workflow
// (scripts/workflow/grabar-datos.sh), sin modelo, tras el commit de la tarea
// del manifiesto, con la etiqueta grabacion, la variable de grabación y la raíz
// de evidencias; sin cualquiera de las dos variables, o con un manifiesto que
// no se puede grabar entero, falla antes de pedir nada.
func TestGrabarFiestasLaborales(t *testing.T) {
	t.Parallel()

	if valor := os.Getenv(httpx.VariableGrabacion); valor != valorQueGraba {
		t.Fatalf("TestGrabarFiestasLaborales pide a la fuente real y graba lo que responde: solo se ejecuta con %s=%s, "+
			"desde scripts/workflow/grabar-datos.sh (la variable vale %q)", httpx.VariableGrabacion, valorQueGraba, valor)
	}

	raiz := os.Getenv(variableDeEvidencias)
	if raiz == "" {
		t.Fatalf("TestGrabarFiestasLaborales graba en la raíz de evidencias que da %s (evidencias/<hito>/), "+
			"que pone scripts/workflow/grabar-datos.sh y que está vacía", variableDeEvidencias)
	}

	peticiones, err := peticionesDeGrabacion(ficheroDelManifiesto)
	require.NoError(t, err)

	cliente, err := httpx.New(
		httpx.ConFuente(nombreDeLaFuente),
		httpx.ConRaizDeGrabacion(raiz),
		httpx.ConIntervalo(boe.IntervaloEntrePeticiones),
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
// con la forma de los del BOE, un año y la dirección del documento de su
// identificador.
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

// comprobar exige el identificador con la forma de los del BOE, el año y la
// dirección del documento del identificador: la misma para todas, que es la que
// la fila de docs/SOURCES.md deja abierta y la que lee la derivación.
func (d disposicion) comprobar() error {
	if err := boe.ValidarNorma(d.ID); err != nil {
		return err
	}

	if d.Anio <= 0 {
		return errors.New("sin año")
	}

	if esperada := prefijoDeLaDireccion + d.ID; d.URL != esperada {
		return fmt.Errorf("la dirección es %q y tiene que ser %q", d.URL, esperada)
	}

	return nil
}
