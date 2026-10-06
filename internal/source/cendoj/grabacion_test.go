//go:build grabacion

package cendoj

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/httpx"
)

// valorQueGraba es el único valor de la variable de grabación que la enciende
// en internal/httpx.
const valorQueGraba = "1"

// TestGrabarConsultas graba contra la fuente real, con internal/httpx, la
// página del buscador y las consultas del manifiesto, con las opciones de red de
// la fuente y al ritmo de su fila: nueve peticiones con el robots.txt, separadas
// cinco segundos (contrato fuente-cendoj-y-grabacion §5; research D27; FR-092).
//
// El manifiesto, sus peticiones y lo que se hace con cada respuesta no se
// declaran aquí: son los de casos_test.go, que se compila siempre, de modo que
// lo que graba este test es lo que TestGrabacionRechazaLoNoReconocido comprueba
// sin red. Graba en un directorio temporal del test, y solo si la página
// responde 200 y cada respuesta es una lista de resultados o «No se ha
// encontrado ningún resultado» copia lo grabado a la carpeta de las grabaciones
// del paquete. Si alguna no se reconoce —un bloqueo, un CAPTCHA, una página que
// ha cambiado—, falla nombrando la consulta, no insiste y no deja nada: lo
// grabado se va con el temporal. No deja material de origen en ningún otro
// sitio.
//
// Toca la red. Solo lo ejecuta el paso grabar_datos del workflow, sin modelo,
// que pone la etiqueta grabacion y la variable de grabación: sin la variable
// falla antes de pedir nada.
func TestGrabarConsultas(t *testing.T) {
	t.Parallel()

	if valor := os.Getenv(httpx.VariableGrabacion); valor != valorQueGraba {
		t.Fatalf("TestGrabarConsultas pide a la fuente real y graba lo que responde: solo se ejecuta con %s=%s, "+
			"desde el paso grabar_datos del workflow (la variable vale %q)",
			httpx.VariableGrabacion, valorQueGraba, valor)
	}

	manifiesto, err := leerManifiesto(ficheroDelManifiesto)
	require.NoError(t, err)

	raiz := t.TempDir()

	cliente, err := httpx.New(append(OpcionesDeRed(),
		httpx.ConIntervalo(IntervaloEntrePeticiones),
		httpx.ConRaizDeGrabacion(raiz),
	)...)
	require.NoError(t, err)

	require.NoError(t, grabarConsultas(t.Context(), cliente, manifiesto.Consultas, raiz, carpetaDeLasGrabaciones))

	t.Logf("grabado en %s: la página del buscador y %d consultas", carpetaDeLasGrabaciones, len(manifiesto.Consultas))
}
