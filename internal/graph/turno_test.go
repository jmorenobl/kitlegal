package graph

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// TestEntregaSinTurnoDentroDelPlazo fija que la entrega que espera su turno de
// escritura en el proceso y no lo recibe antes de que termine su contexto es el
// plazo agotado, «fuente-no-disponible» con la causa del contexto, y que no
// toca el disco: el directorio de world.db sigue sin existir.
func TestEntregaSinTurnoDentroDelPlazo(t *testing.T) {
	t.Parallel()

	directorio := filepath.Join(t.TempDir(), "no-existe")
	ruta := filepath.Join(directorio, ficheroDelGrafo)

	soltar, err := tomarTurno(t.Context(), ruta)
	require.NoError(t, err)
	t.Cleanup(soltar)

	ctx, cancelar := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancelar()

	err = Nuevo(ConDirectorio(directorio)).Apply(ctx, loteDelTerritorio())

	var fallo *Error

	require.ErrorAs(t, err, &fallo)
	assert.Equal(t, schema.ClaseFuenteNoDisponible, fallo.Clase())
	assert.Equal(t, operacionEscribir, fallo.Operacion)
	assert.Equal(t, ruta, fallo.Ruta)
	require.ErrorIs(t, err, context.DeadlineExceeded)

	_, err = os.Lstat(directorio)
	assert.ErrorIs(t, err, os.ErrNotExist, "sin turno, la entrega no crea nada")
}

// TestElTurnoNoCuentaEnLaEsperaPropia fija que lo que una entrega espera a otra
// del mismo proceso no consume su espera propia de 5 s, que es la de otra
// invocación (FR-014): con el turno tomado más que esa espera, la entrega no
// termina mientras no lo recibe y, al recibirlo, sale bien.
func TestElTurnoNoCuentaEnLaEsperaPropia(t *testing.T) {
	t.Parallel()

	directorio := t.TempDir()

	soltar, err := tomarTurno(t.Context(), filepath.Join(directorio, ficheroDelGrafo))
	require.NoError(t, err)

	terminada := make(chan error, 1)

	go func() { terminada <- Nuevo(ConDirectorio(directorio)).Apply(t.Context(), loteDelTerritorio()) }()

	select {
	case err := <-terminada:
		soltar()
		require.Failf(t, "la entrega termina sin turno", "con %v", err)
	case <-time.After(esperaPropia + tramoDeEspera):
	}

	soltar()

	require.NoError(t, <-terminada, "con el turno, la entrega sale bien aunque haya esperado más que la espera propia")
}

// TestEntregasSimultaneasDelProceso fija que muchas entregas del mismo proceso
// a la vez contra el mismo world.db, como las llamadas simultáneas del servidor
// MCP, salen todas bien, cada una con su almacén.
func TestEntregasSimultaneasDelProceso(t *testing.T) {
	t.Parallel()

	const entregas = 32

	directorio := t.TempDir()
	salida := make(chan struct{})
	fallos := make([]error, entregas)

	var grupo sync.WaitGroup

	for i := range entregas {
		grupo.Go(func() {
			<-salida

			fallos[i] = Nuevo(ConDirectorio(directorio)).Apply(t.Context(), loteDelTerritorio())
		})
	}

	close(salida)
	grupo.Wait()

	for i, err := range fallos {
		require.NoError(t, err, "la entrega %d", i)
	}
}
