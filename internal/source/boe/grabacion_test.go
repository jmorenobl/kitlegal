//go:build grabacion

package boe

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/httpx"
)

// Lo que fija la grabación de las respuestas reales de la fuente (contrato
// esquemas-fixtures-y-controles §3; research.md D12).
const (
	// raizDeGrabacion es la carpeta de datos de prueba del paquete: la grabación
	// escribe bajo ella <fuente>/<nombre>.json.
	raizDeGrabacion = "testdata"
	// rutaDelManifiesto es la lista cerrada de recursos que se graban.
	rutaDelManifiesto = "testdata/grabaciones.json"
	// valorQueGraba es el único valor de la variable de grabación que la
	// enciende en internal/httpx.
	valorQueGraba = "1"
	// aceptaXML y aceptaJSON son los formatos en que la fuente pide cada recurso:
	// XML el bloque y JSON todo lo demás (contrato verbos-y-salidas §1 a §6;
	// refs/boe.py 406-407). La reproducción no empareja por Accept, así que lo que
	// decide es que la grabación reciba el cuerpo en el formato que leerá la
	// fuente.
	aceptaXML  = "application/xml"
	aceptaJSON = "application/json"
)

// Los recursos que admite el manifiesto.
const (
	recursoBusqueda  = "busqueda"
	recursoIndice    = "indice"
	recursoMetadatos = "metadatos"
	recursoAnalisis  = "analisis"
	recursoBloque    = "bloque"
)

// manifiestoDeGrabacion es testdata/grabaciones.json: la fuente y sus recursos,
// sin dirección ni Accept, que construye el código.
type manifiestoDeGrabacion struct {
	Fuente   string               `json:"fuente"`
	Recursos []recursoDeGrabacion `json:"recursos"`
}

// recursoDeGrabacion es una entrada del manifiesto: el recurso, sus argumentos y
// para qué se graba.
type recursoDeGrabacion struct {
	Recurso string `json:"recurso"`
	Texto   string `json:"texto"`
	Norma   string `json:"norma"`
	Bloque  string `json:"bloque"`
	Para    string `json:"para"`
}

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

	peticiones, err := peticionesDelManifiesto(rutaDelManifiesto)
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

// peticionesDelManifiesto lee el manifiesto sin admitir claves desconocidas y
// construye todas sus peticiones antes de que se pida ninguna, de modo que un
// manifiesto que no se puede grabar entero no graba nada.
func peticionesDelManifiesto(ruta string) ([]httpx.Peticion, error) {
	contenido, err := os.ReadFile(filepath.Clean(ruta))
	if err != nil {
		return nil, fmt.Errorf("el manifiesto de grabación no se puede leer: %w", err)
	}

	decodificador := json.NewDecoder(bytes.NewReader(contenido))
	decodificador.DisallowUnknownFields()

	var manifiesto manifiestoDeGrabacion
	if err := decodificador.Decode(&manifiesto); err != nil {
		return nil, fmt.Errorf("%s no es un manifiesto de grabación: %w", ruta, err)
	}

	if _, err := decodificador.Token(); !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("%s lleva algo detrás del manifiesto de grabación", ruta)
	}

	if manifiesto.Fuente != NombreDeLaFuente {
		return nil, fmt.Errorf("%s es de la fuente %q y este paquete es %q", ruta, manifiesto.Fuente, NombreDeLaFuente)
	}

	if len(manifiesto.Recursos) == 0 {
		return nil, fmt.Errorf("%s no lista ningún recurso", ruta)
	}

	peticiones := make([]httpx.Peticion, 0, len(manifiesto.Recursos))

	for indice, recurso := range manifiesto.Recursos {
		peticion, err := peticionDeGrabacion(recurso)
		if err != nil {
			return nil, fmt.Errorf("%s, recurso %d (%s): %w", ruta, indice+1, recurso.Recurso, err)
		}

		peticiones = append(peticiones, peticion)
	}

	return peticiones, nil
}

// peticionDeGrabacion es la petición GET de un recurso del manifiesto: exige los
// argumentos de su clase y ningún otro, y los valida como la fuente antes de
// construir su dirección.
func peticionDeGrabacion(recurso recursoDeGrabacion) (httpx.Peticion, error) {
	switch recurso.Recurso {
	case recursoBusqueda:
		if recurso.Norma != "" || recurso.Bloque != "" {
			return httpx.Peticion{}, errors.New("una búsqueda solo lleva texto")
		}

		direccion, err := direccionDeBusqueda(recurso.Texto)
		if err != nil {
			return httpx.Peticion{}, err
		}

		return httpx.Peticion{Metodo: "GET", URL: direccion, Acepta: aceptaJSON}, nil

	case recursoIndice, recursoMetadatos, recursoAnalisis:
		if recurso.Texto != "" || recurso.Bloque != "" {
			return httpx.Peticion{}, errors.New("un índice, unos metadatos o un análisis solo llevan norma")
		}

		if err := ValidarNorma(recurso.Norma); err != nil {
			return httpx.Peticion{}, err
		}

		direccion := map[string]func(string) string{
			recursoIndice:    direccionDelIndice,
			recursoMetadatos: direccionDeLosMetadatos,
			recursoAnalisis:  direccionDelAnalisis,
		}[recurso.Recurso](recurso.Norma)

		return httpx.Peticion{Metodo: "GET", URL: direccion, Acepta: aceptaJSON}, nil

	case recursoBloque:
		if recurso.Texto != "" {
			return httpx.Peticion{}, errors.New("un bloque solo lleva norma y bloque")
		}

		if err := errors.Join(ValidarNorma(recurso.Norma), ValidarBloque(recurso.Bloque)); err != nil {
			return httpx.Peticion{}, err
		}

		return httpx.Peticion{Metodo: "GET", URL: direccionDelBloque(recurso.Norma, recurso.Bloque), Acepta: aceptaXML}, nil

	default:
		return httpx.Peticion{}, fmt.Errorf("recurso desconocido %q: el manifiesto admite %s, %s, %s, %s y %s",
			recurso.Recurso, recursoBusqueda, recursoIndice, recursoMetadatos, recursoAnalisis, recursoBloque)
	}
}
