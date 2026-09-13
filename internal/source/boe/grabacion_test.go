//go:build grabacion

package boe

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/httpx"
)

// Lo que fija la grabación de las respuestas reales de la fuente (contrato
// esquemas-fixtures-y-controles §3; research.md D12).
//
// El manifiesto, sus recursos y la construcción de sus peticiones no se
// declaran aquí: son los de casos_test.go, que se compila siempre. Así se pide
// exactamente la petición que después reproduce TestGrabacionesCompletas, y no
// dos construcciones que pueden divergir (redelimitación de T009, intento 1).
const (
	// raizDeGrabacion es la carpeta de datos de prueba del paquete: la grabación
	// escribe bajo ella <fuente>/<nombre>.json.
	raizDeGrabacion = "testdata"
	// valorQueGraba es el único valor de la variable de grabación que la
	// enciende en internal/httpx.
	valorQueGraba = "1"
)

// TestGrabarFixtures graba contra la fuente real, con internal/httpx, cada
// recurso del manifiesto, con la dirección de direcciones.go o busqueda.go, el
// formato de la fuente y el ritmo de terminos.go (FR-113). No lee ninguna
// respuesta: lo grabado lo revisa una persona.
//
// Toca la red. Solo lo ejecuta scripts/grabar-fixtures.sh, que pone la etiqueta
// grabacion y la variable de grabación, a mano y en la pausa de la tarea [datos]
// del manifiesto (contrato §3.2): sin la variable falla antes de pedir nada.
func TestGrabarFixtures(t *testing.T) {
	t.Parallel()

	if valor := os.Getenv(httpx.VariableGrabacion); valor != valorQueGraba {
		t.Fatalf("TestGrabarFixtures pide a la fuente real y graba lo que responde: solo se ejecuta con %s=%s, "+
			"desde scripts/grabar-fixtures.sh (la variable vale %q)", httpx.VariableGrabacion, valorQueGraba, valor)
	}

	peticiones, err := peticionesDeGrabacion(ficheroDelManifiesto)
	require.NoError(t, err)

	cliente, err := httpx.New(
		httpx.ConFuente(NombreDeLaFuente),
		httpx.ConRaizDeGrabacion(raizDeGrabacion),
		httpx.ConIntervalo(IntervaloEntrePeticiones),
	)
	require.NoError(t, err)

	for _, peticion := range peticiones {
		_, err := cliente.Pedir(t.Context(), schema.Contexto{}, peticion)
		require.NoErrorf(t, err, "la grabación queda incompleta en %s %s", peticion.Metodo, peticion.URL)

		t.Logf("grabado: %s %s", peticion.Metodo, peticion.URL)
	}
}

// peticionesDeGrabacion lee el manifiesto con leerManifiesto y construye todas
// sus peticiones antes de que se pida ninguna, de modo que un manifiesto que no
// se puede grabar entero no graba nada.
func peticionesDeGrabacion(ruta string) ([]httpx.Peticion, error) {
	manifiesto, err := leerManifiesto(ruta)
	if err != nil {
		return nil, err
	}

	peticiones := make([]httpx.Peticion, 0, len(manifiesto.Recursos))

	for indice, recurso := range manifiesto.Recursos {
		peticion, err := recurso.peticion()
		if err != nil {
			return nil, fmt.Errorf("%s, recurso %d (%s): %w", ruta, indice+1, recurso.Recurso, err)
		}

		peticiones = append(peticiones, peticion)
	}

	return peticiones, nil
}
