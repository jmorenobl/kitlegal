package disco

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// nombreDeSonda es la forma del nombre de la sonda: el prefijo y dieciséis
// hexadecimales aleatorios (research.md D9).
var nombreDeSonda = regexp.MustCompile(`^\.kitlegal-sonda-[0-9a-f]{16}$`)

// errRetirarSimulado es el fallo con el que el test sustituye la retirada de
// la sonda.
var errRetirarSimulado = errors.New("retirada simulada que falla")

// casoDelEnlazador es un caso de TestEnlazadorDelSistema: con qué TMPDIR se
// ejecuta y qué comprueba sobre un directorio temporal recién creado.
type casoDelEnlazador struct {
	nombre string
	// tmpdir es lo que vale TMPDIR durante el caso, preparado dentro de
	// temporal, un directorio distinto del que se sondea.
	tmpdir func(t *testing.T, temporal string) string
	probar func(t *testing.T, directorio, temporal string)
}

// TestEnlazadorDelSistema fija research.md D9 para el Enlazador del sistema:
// Disponible responde con una sonda **en el directorio que se le pasa**, no en
// TMPDIR, y lo deja con las mismas entradas y los mismos bytes; un nombre de
// sonda que ya existe se reintenta con otro sin tocar lo que había, hasta
// ocho; y una sonda que no se puede retirar es un error que la nombra
// (contracts/applet-skills.md §5). Enlazar crea el enlace con su destino
// literal, y el Escritor enlaza con el Enlazador que recibe (FR-024).
//
// Cada caso fija TMPDIR con t.Setenv —casi todos en un fichero, donde no se
// puede crear nada, para que cualquier uso de TMPDIR falle—, así que la tabla
// entera va en secuencia: t.Setenv no admite t.Parallel. Los dos temporales
// del caso se crean antes, porque t.TempDir usa TMPDIR.
func TestEnlazadorDelSistema(t *testing.T) {
	casos := []casoDelEnlazador{
		{
			nombre: "con TMPDIR en un fichero, sondea en el directorio y lo deja igual",
			tmpdir: tmpdirEnUnFichero, probar: probarQueLoDejaIgual,
		},
		{
			nombre: "la sonda es un enlace con destino literal en el propio directorio",
			tmpdir: tmpdirEnUnFichero, probar: probarDondeSondea,
		},
		{
			nombre: "sin permiso de escritura no hay enlaces, aunque TMPDIR los admita",
			tmpdir: tmpdirQueAdmiteEnlaces, probar: probarSinPermisoDeEscritura,
		},
		{
			nombre: "en un directorio que no existe no hay enlaces, y no se crea",
			tmpdir: tmpdirEnUnFichero, probar: probarDirectorioQueNoExiste,
		},
		{
			nombre: "un nombre de sonda que ya existe se reintenta con otro sin tocarlo",
			tmpdir: tmpdirEnUnFichero, probar: probarNombreQueYaExiste,
		},
		{
			nombre: "sin un nombre libre en ocho intentos es un error",
			tmpdir: tmpdirEnUnFichero, probar: probarSinNombreLibre,
		},
		{
			nombre: "una sonda que no se puede retirar es un error que la nombra",
			tmpdir: tmpdirEnUnFichero, probar: probarRetiradaQueFalla,
		},
		{
			nombre: "Enlazar crea el enlace con su destino literal, aunque cuelgue",
			tmpdir: tmpdirEnUnFichero, probar: probarEnlazar,
		},
		{
			nombre: "Enlazar donde ya hay algo no toca nada",
			tmpdir: tmpdirEnUnFichero, probar: probarEnlazarDondeHayAlgo,
		},
		{
			nombre: "el Escritor enlaza con el Enlazador que recibe",
			tmpdir: tmpdirEnUnFichero, probar: probarEscritorEnlaza,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			directorio := t.TempDir()
			temporal := t.TempDir()

			t.Setenv("TMPDIR", caso.tmpdir(t, temporal))

			caso.probar(t, directorio, temporal)
		})
	}
}

// tmpdirEnUnFichero deja en temporal un fichero y lo devuelve como TMPDIR:
// ahí no se puede crear nada, así que lo que use TMPDIR falla.
func tmpdirEnUnFichero(t *testing.T, temporal string) string {
	t.Helper()

	fichero := filepath.Join(temporal, "no-es-un-directorio")
	escribirDePrueba(t, fichero, "")

	return fichero
}

// tmpdirQueAdmiteEnlaces devuelve temporal como TMPDIR, tras comprobar que en
// él se pueden crear enlaces.
func tmpdirQueAdmiteEnlaces(t *testing.T, temporal string) string {
	t.Helper()

	enlace := filepath.Join(temporal, "comprobacion")
	enlazarDePrueba(t, "destino", enlace)
	require.NoError(t, os.Remove(enlace))

	return temporal
}

// probarQueLoDejaIgual: el Enlazador del sistema, su valor cero, da verdadero
// en un directorio escribible con TMPDIR en un fichero, y lo deja con las
// mismas entradas y los mismos bytes (FR-048, FR-068).
func probarQueLoDejaIgual(t *testing.T, directorio, _ string) {
	t.Helper()

	conEnlaceA("fichero")(t, directorio)
	antes := arbolDe(t, directorio)

	disponible, err := Enlazador{}.Disponible(directorio)
	require.NoError(t, err)

	assert.True(t, disponible)
	assert.Equal(t, antes, arbolDe(t, directorio))
}

// probarDondeSondea: la sonda es un enlace de nombre .kitlegal-sonda-<16
// hexadecimales> con el destino literal kitlegal-sonda, creado en el
// directorio que se pasa; se observa en la retirada, que el test sustituye por
// una que la mira antes de retirarla.
func probarDondeSondea(t *testing.T, directorio, _ string) {
	t.Helper()

	var sondas []string

	enlazador := Enlazador{retirar: func(sonda string) error {
		sondas = append(sondas, sonda)

		destino, err := os.Readlink(sonda)
		require.NoError(t, err)
		assert.Equal(t, "kitlegal-sonda", destino)

		return os.Remove(sonda)
	}}

	disponible, err := enlazador.Disponible(directorio)
	require.NoError(t, err)

	assert.True(t, disponible)
	require.Len(t, sondas, 1)
	assert.Equal(t, directorio, filepath.Dir(sondas[0]))
	assert.Regexp(t, nombreDeSonda, filepath.Base(sondas[0]))
	assert.Equal(t, map[string]string{".": "directorio"}, arbolDe(t, directorio))
}

// probarSinPermisoDeEscritura: en un directorio donde no se puede crear nada,
// Disponible es falso aunque en TMPDIR sí se puedan crear enlaces, y ni el
// directorio ni TMPDIR cambian.
func probarSinPermisoDeEscritura(t *testing.T, directorio, temporal string) {
	t.Helper()

	conFichero(t, directorio)
	cambiarPermisos(t, directorio, 0o500)
	antes, antesEnTemporal := arbolDe(t, directorio), arbolDe(t, temporal)

	disponible, err := Enlazador{}.Disponible(directorio)
	require.NoError(t, err)

	assert.False(t, disponible)
	assert.Equal(t, antes, arbolDe(t, directorio))
	assert.Equal(t, antesEnTemporal, arbolDe(t, temporal), "TMPDIR no se toca")
}

// probarDirectorioQueNoExiste: la sonda no crea ningún directorio; donde no
// hay directorio no hay enlaces.
func probarDirectorioQueNoExiste(t *testing.T, directorio, _ string) {
	t.Helper()

	disponible, err := Enlazador{}.Disponible(filepath.Join(directorio, "falta"))
	require.NoError(t, err)

	assert.False(t, disponible)
	assert.Equal(t, map[string]string{".": "directorio"}, arbolDe(t, directorio))
}

// probarNombreQueYaExiste: si el primer nombre de sonda ya existe, se prueba
// otro, y lo que había con el primero se queda como estaba.
func probarNombreQueYaExiste(t *testing.T, directorio, _ string) {
	t.Helper()

	nombres := []string{"0123456789abcdef", "fedcba9876543210"}
	escribirDePrueba(t, filepath.Join(directorio, ".kitlegal-sonda-"+nombres[0]), "ajeno")
	antes := arbolDe(t, directorio)

	pedidos := 0
	enlazador := Enlazador{aleatorio: func() string {
		pedidos++

		return nombres[pedidos-1]
	}}

	disponible, err := enlazador.Disponible(directorio)
	require.NoError(t, err)

	assert.True(t, disponible)
	assert.Equal(t, 2, pedidos, "un nombre nuevo tras el que ya existía")
	assert.Equal(t, antes, arbolDe(t, directorio))
}

// probarSinNombreLibre: si los ocho nombres de sonda que se prueban ya
// existen, Disponible no decide y es un error que nombra el directorio.
func probarSinNombreLibre(t *testing.T, directorio, _ string) {
	t.Helper()

	escribirDePrueba(t, filepath.Join(directorio, ".kitlegal-sonda-0123456789abcdef"), "ajeno")
	antes := arbolDe(t, directorio)

	pedidos := 0
	enlazador := Enlazador{aleatorio: func() string {
		pedidos++

		return "0123456789abcdef"
	}}

	disponible, err := enlazador.Disponible(directorio)

	require.Error(t, err)
	assert.False(t, disponible)
	assert.Equal(t, "crear la sonda en "+directorio+": los 8 nombres probados ya existían", err.Error())
	assert.Equal(t, 8, pedidos)
	assert.Equal(t, antes, arbolDe(t, directorio))
}

// probarRetiradaQueFalla: si la sonda no se puede retirar, el disco ya no
// queda como estaba, y Disponible es un error «retirar la sonda <ruta>: <error
// del sistema>» (contracts/applet-skills.md §5).
func probarRetiradaQueFalla(t *testing.T, directorio, _ string) {
	t.Helper()

	var sonda string

	enlazador := Enlazador{retirar: func(ruta string) error {
		sonda = ruta

		return &fs.PathError{Op: "remove", Path: ruta, Err: errRetirarSimulado}
	}}

	disponible, err := enlazador.Disponible(directorio)

	require.ErrorIs(t, err, errRetirarSimulado)
	assert.False(t, disponible)
	assert.Equal(t, "retirar la sonda "+sonda+": "+errRetirarSimulado.Error(), err.Error())
	assert.True(t, strings.HasPrefix(sonda, directorio+string(filepath.Separator)))
	assert.Equal(t, "enlace -> kitlegal-sonda", arbolDe(t, directorio)[filepath.Base(sonda)],
		"la sonda sigue ahí: por eso es un error")
}

// probarEnlazar: Enlazar crea el enlace con el destino literal, sin limpiarlo
// ni comprobar que resuelve (FR-021).
func probarEnlazar(t *testing.T, directorio, _ string) {
	t.Helper()

	ruta := filepath.Join(directorio, "boe-legislacion")

	require.NoError(t, Enlazador{}.Enlazar("../../.agents/skills/boe-legislacion", ruta))

	assert.Equal(t, map[string]string{
		".": "directorio", "boe-legislacion": "enlace -> ../../.agents/skills/boe-legislacion",
	}, arbolDe(t, directorio))
}

// probarEnlazarDondeHayAlgo: Enlazar no sustituye lo que ya hay en la ruta.
func probarEnlazarDondeHayAlgo(t *testing.T, directorio, _ string) {
	t.Helper()

	ruta := conFichero(t, directorio)
	antes := arbolDe(t, directorio)

	err := Enlazador{}.Enlazar("../../.agents/skills/boe-legislacion", ruta)

	require.ErrorIs(t, err, fs.ErrExist)
	assert.Equal(t, antes, arbolDe(t, directorio))
}

// probarEscritorEnlaza: el Escritor no crea enlaces por su cuenta, sino con el
// Enlazador de la invocación, que en test y en el binario de e2e puede ser uno
// que falla (FR-024; research.md D9).
func probarEscritorEnlaza(t *testing.T, directorio, _ string) {
	t.Helper()

	ruta := filepath.Join(directorio, "boe-legislacion")

	err := NuevoEscritor(enlazadorQueFalla{}).Enlazar("../../.agents/skills/boe-legislacion", ruta)

	require.ErrorIs(t, err, errSinEnlaces)
	assert.Equal(t, map[string]string{".": "directorio"}, arbolDe(t, directorio))

	require.NoError(t, NuevoEscritor(Enlazador{}).Enlazar("../../.agents/skills/boe-legislacion", ruta))
	assert.Equal(t, "enlace -> ../../.agents/skills/boe-legislacion", arbolDe(t, directorio)["boe-legislacion"])
}

// errSinEnlaces es el fallo de enlazadorQueFalla.
var errSinEnlaces = errors.New("este sistema de ficheros no admite enlaces simbólicos")

// enlazadorQueFalla es un Enlazador que no puede crear ningún enlace.
type enlazadorQueFalla struct{}

// Disponible dice que no se pueden crear enlaces.
func (enlazadorQueFalla) Disponible(string) (bool, error) { return false, nil }

// Enlazar falla sin crear nada.
func (enlazadorQueFalla) Enlazar(string, string) error { return errSinEnlaces }
