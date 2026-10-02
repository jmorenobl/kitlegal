package empaquetado_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/empaquetado"
)

// El esquema JSON oficial de la versión `0.3` del manifiesto de MCP Bundle,
// versionado en testdata/mcpb/ de la raíz tal como lo publica su fuente
// (testdata/mcpb/README.md; ADR 0035). El control del manifiesto de H22 lo
// pedía y el run no pudo traerlo: lo que escribe el paso se valida aquí contra
// él, en make ci, y lo que lleva el snapshot, en TestSnapshot.
const (
	rutaDelEsquemaOficial = "../../testdata/mcpb/mcpb-manifest-v0.3.schema.json"
	// huellaDelEsquemaOficial es el sha256 del fichero de la fuente: el esquema
	// no se edita a mano, se sustituye entero con su huella.
	huellaDelEsquemaOficial = "3a0ac9d845711a1b9b17dfa5a52f8b60628239d6a86a9db417206a9efc78592d"
	urlDelEsquemaOficial    = "https://kitlegal.es/testdata/mcpb-manifest-v0.3.schema.json"
)

// esquemaOficial compila el esquema versionado.
func esquemaOficial(t *testing.T) *jsonschema.Schema {
	t.Helper()

	leido, err := os.ReadFile(filepath.FromSlash(rutaDelEsquemaOficial))
	require.NoError(t, err)

	documento, err := jsonschema.UnmarshalJSON(bytes.NewReader(leido))
	require.NoError(t, err)

	compilador := jsonschema.NewCompiler()
	require.NoError(t, compilador.AddResource(urlDelEsquemaOficial, documento))

	esquema, err := compilador.Compile(urlDelEsquemaOficial)
	require.NoError(t, err)

	return esquema
}

// validarConElEsquemaOficial devuelve el error de validar un manifiesto contra
// el esquema oficial de la versión `0.3`.
func validarConElEsquemaOficial(t *testing.T, manifiesto []byte) error {
	t.Helper()

	documento, err := jsonschema.UnmarshalJSON(bytes.NewReader(manifiesto))
	require.NoError(t, err)

	return esquemaOficial(t).Validate(documento)
}

// TestEsquemaOficial comprueba que el esquema versionado es el de su fuente,
// byte a byte, y que rechaza lo que tiene que rechazar: un manifiesto del paso
// sin `server`, con un tipo de servidor que no existe y con un campo que el
// esquema no nombra. Sin esto, un esquema vacío daría por bueno cualquier
// manifiesto.
func TestEsquemaOficial(t *testing.T) {
	t.Parallel()

	t.Run("es el de su fuente, byte a byte", func(t *testing.T) {
		t.Parallel()

		leido, err := os.ReadFile(filepath.FromSlash(rutaDelEsquemaOficial))
		require.NoError(t, err)

		huella := sha256.Sum256(leido)

		assert.Equal(t, huellaDelEsquemaOficial, hex.EncodeToString(huella[:]),
			"%s no es el fichero de su fuente: se sustituye entero, con su huella aquí y en testdata/mcpb/README.md",
			rutaDelEsquemaOficial)
	})

	piezas := piezasDePrueba(t)
	require.NoError(t, empaquetado.EscribirPiezas(piezas))

	escrito := string(contenidoDe(t, leerZip(t, filepath.Join(piezas.Salida, nombreDeLaExtension)), rutaDelManifiesto))
	require.NoError(t, validarConElEsquemaOficial(t, []byte(escrito)), "el manifiesto del paso, sin tocar")

	rechazados := []struct {
		nombre string
		viejo  string
		nuevo  string
	}{
		{nombre: "sin `server`", viejo: `"server":`, nuevo: `"servidor":`},
		{nombre: "con un tipo de servidor que no existe", viejo: `"type": "binary"`, nuevo: `"type": "go"`},
		{nombre: "con un campo que el esquema no nombra", viejo: `"manifest_version":`, nuevo: `"de_mas": true, "manifest_version":`},
	}

	for _, caso := range rechazados {
		t.Run("rechaza un manifiesto "+caso.nombre, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, 1, strings.Count(escrito, caso.viejo), "premisa: el manifiesto lleva %s una vez", caso.viejo)

			cambiado := strings.Replace(escrito, caso.viejo, caso.nuevo, 1)

			assert.Error(t, validarConElEsquemaOficial(t, []byte(cambiado)))
		})
	}
}
