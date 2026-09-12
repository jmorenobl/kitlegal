package cache

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	// entradasConfirmadas es cuántas entradas confirma el escritor mientras el
	// otro lee. Pocas y de varias páginas cada una: bastan para que cada lectura
	// coincida con una escritura en curso, y juntas no llenan el registro de
	// escritura hasta el punto de control automático de SQLite (mil páginas), que
	// es lo que permite comprobar al final que lo leído seguía en el registro.
	entradasConfirmadas = 32

	// bytesPorPagina es el tamaño de página por omisión de SQLite. Un contenido
	// que ocupa varias es el que podría leerse a medias si una escritura no fuera
	// atómica, y por eso los de estas pruebas ocupan varias.
	bytesPorPagina = 4096
)

// TestDosClientesEnElMismoProceso fija el escenario 4 de US5, FR-032 y SC-007 en
// el mismo proceso: uno escribe y otro lee a la vez sobre la misma base, cada
// cual con su propia conexión —que es como se comporta cada invocación—, y el
// lector encuentra toda entrada confirmada, entera y sin fallar por bloqueo.
//
// La pareja se mide dos veces, una por cada modo de apertura del lector, porque
// el de solo lectura puede fallar de una forma que el normal no tiene: un lector
// que abriera la base como inmutable no vería el registro de escritura y
// declararía ausentes —código 4— entradas confirmadas (FR-015, US3 escenario 6).
// Las dos subpruebas van en un solo nivel: el escenario 12 del quickstart cuenta
// sus líneas.
//
// El escritor avisa por un canal de cada entrada en cuanto Put vuelve, y espera
// a que el lector recoja el aviso antes de escribir la siguiente: así cada
// lectura coincide con una escritura en curso y no con un escritor que ya
// terminó. Al final el lector las vuelve a leer todas con el escritor todavía
// abierto, y cache.db conserva la huella de antes de empezar: lo leído estaba en
// el registro de escritura y no en el fichero principal, que es lo que hace que
// la subprueba de solo lectura mida algo.
func TestDosClientesEnElMismoProceso(t *testing.T) {
	t.Parallel()

	for _, modo := range modosDeApertura {
		t.Run("normal-"+modo.nombre, func(t *testing.T) {
			t.Parallel()

			directorio := t.TempDir()
			ruta := filepath.Join(directorio, ficheroDeLaBase)

			// El escritor se construye primero porque es quien crea la base y la
			// migra: un lector de solo lectura construido antes se quedaría sin
			// base para toda su vida (FR-015).
			escritor := clienteAbierto(t, directorio)
			lector := clienteAbierto(t, directorio, modo.opciones...)

			antes := huella(t, ruta)

			avisos := make(chan int)
			resultado := make(chan error, 1)

			var grupo sync.WaitGroup

			grupo.Go(func() { resultado <- escribeYAvisa(t.Context(), escritor, avisos) })

			// Si una comprobación falla a mitad, el contexto de la prueba se
			// cancela antes de las limpiezas y el escritor deja de esperar. Esta
			// espera se registra después de los cierres, así que corre antes que
			// ellos y ningún cliente se cierra con una escritura en curso.
			t.Cleanup(grupo.Wait)

			leidas := 0

			for indice := range avisos {
				compruebaConfirmada(t, lector, indice)

				leidas++
			}

			require.NoError(t, <-resultado, "el escritor no falla por bloqueo mientras el otro lee (FR-032)")
			require.Equal(t, entradasConfirmadas, leidas, "el lector recibió el aviso de cada entrada confirmada")

			for indice := range entradasConfirmadas {
				compruebaConfirmada(t, lector, indice)
			}

			require.Equal(t, antes, huella(t, ruta),
				"sin punto de control: lo que el lector encontró estaba en el registro de escritura")
		})
	}
}

// TestMismaClaveDosEscritores fija FR-007 y el caso límite de la misma clave
// escrita a la vez desde dos conexiones: la entrada nunca queda mezclada entre
// las dos escrituras, y la última escritura completa es la que queda.
//
// Cada escritor guarda siempre lo mismo —un contenido de varias páginas, de
// longitud y de byte distintos— con su propio reloj, de modo que el contenido y
// el instante de expiración de la fila dicen, cada uno por su lado, de qué
// escritura salen. Mientras los dos escriben, un tercer cliente lee sin parar y
// cada lectura tiene que ser exactamente una de las dos; al terminar, la fila es
// una sola y su expiración es la del mismo escritor que su contenido. Una
// sustitución en dos sentencias —el contenido primero, la vigencia después—
// puede dejar la fila con el contenido de uno y la vigencia del otro, y eso es lo
// que mira la segunda comprobación.
//
// Quién gana mientras los dos escriben a la vez no se puede prever, y por eso la
// otra mitad se mide en orden: en los dos órdenes posibles lo que queda es lo del
// último, y lo lee también el cliente cuya escritura fue sustituida.
func TestMismaClaveDosEscritores(t *testing.T) {
	t.Parallel()

	const rondas = 32

	directorio := t.TempDir()
	ruta := filepath.Join(directorio, ficheroDeLaBase)

	escritores := []escritorDeLaClave{
		{contenido: bytes.Repeat([]byte("A"), 3*bytesPorPagina), instante: instanteDePrueba},
		{contenido: bytes.Repeat([]byte("B"), 5*bytesPorPagina), instante: instanteDePrueba.Add(time.Minute)},
	}

	for indice := range escritores {
		escritores[indice].cliente = clienteAbierto(t, directorio,
			ConReloj(relojEn(escritores[indice].instante).Ahora))
	}

	lector := clienteAbierto(t, directorio, ConReloj(relojEn(instanteDePrueba).Ahora))

	// La clave está guardada antes de empezar, para que ninguna lectura pueda
	// encontrarla ausente y toda lectura tenga que ser una de las dos.
	require.NoError(t, escritores[0].cliente.Put(t.Context(), claveDePrueba, escritores[0].contenido, vigenciaDePrueba))

	resultados := make(chan error, len(escritores))
	terminados := make(chan struct{})

	var grupo sync.WaitGroup

	for _, escritor := range escritores {
		grupo.Go(func() { resultados <- escritor.escribe(t.Context(), rondas) })
	}

	go func() {
		grupo.Wait()
		close(terminados)
	}()

	// Registrada después de los cierres, corre antes que ellos: ningún cliente
	// se cierra con una escritura en curso, tampoco si una lectura falla a mitad.
	t.Cleanup(func() { <-terminados })

	for escribiendo := true; escribiendo; {
		select {
		case <-terminados:
			escribiendo = false
		default:
		}

		deQuienEs(t, lector, escritores)
	}

	for range escritores {
		require.NoError(t, <-resultados, "dos escrituras de la misma clave se turnan: ninguna falla por bloqueo (FR-032)")
	}

	compruebaFilaDe(t, ruta, escritores[deQuienEs(t, lector, escritores)])

	for _, orden := range [][2]int{{0, 1}, {1, 0}} {
		sustituido, ultimo := escritores[orden[0]], escritores[orden[1]]

		require.NoError(t, sustituido.cliente.Put(t.Context(), claveDePrueba, sustituido.contenido, vigenciaDePrueba))
		require.NoError(t, ultimo.cliente.Put(t.Context(), claveDePrueba, ultimo.contenido, vigenciaDePrueba))

		require.Equal(t, orden[1], deQuienEs(t, sustituido.cliente, escritores),
			"la última escritura completa gana, y la ve también quien escribió antes (FR-007)")
		compruebaFilaDe(t, ruta, ultimo)
	}
}

// escritorDeLaClave es cada uno de los escritores de TestMismaClaveDosEscritores:
// su cliente, lo que guarda siempre y el instante fijo de su reloj, del que sale
// la expiración que deja en la fila.
type escritorDeLaClave struct {
	cliente   *Cliente
	contenido []byte
	instante  time.Time
}

// escribe guarda su contenido bajo la clave de prueba tantas veces como rondas.
func (e escritorDeLaClave) escribe(ctx context.Context, rondas int) error {
	for ronda := range rondas {
		if err := e.cliente.Put(ctx, claveDePrueba, e.contenido, vigenciaDePrueba); err != nil {
			return fmt.Errorf("ronda %d: %w", ronda, err)
		}
	}

	return nil
}

// escribeYAvisa guarda las entradas numeradas una a una y avisa de cada una en
// cuanto está confirmada, que es cuando Put vuelve sin error: el aviso sale
// después de confirmar y nunca antes. Cierra el canal al terminar, también si
// falla, para que el lector no espere para siempre, y deja de esperar al lector
// si el contexto termina.
func escribeYAvisa(ctx context.Context, escritor *Cliente, avisos chan<- int) error {
	defer close(avisos)

	for indice := range entradasConfirmadas {
		if err := escritor.Put(ctx, claveNumerada(indice), cuerpoNumerado(indice), vigenciaDePrueba); err != nil {
			return fmt.Errorf("escribir la entrada %d: %w", indice, err)
		}

		select {
		case avisos <- indice:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

// compruebaConfirmada lee una entrada que el escritor ya confirmó y exige
// encontrarla entera: sin fallo —ni por bloqueo ni por una ausencia falsa, que
// en solo lectura sería el código 4— y byte a byte como se guardó.
func compruebaConfirmada(t *testing.T, lector *Cliente, indice int) {
	t.Helper()

	contenido, presente, err := lector.Get(t.Context(), claveNumerada(indice))
	require.NoError(t, err,
		"una entrada confirmada no falla por bloqueo ni se declara ausente (FR-032, SC-007)")
	require.True(t, presente, "una entrada confirmada está presente")

	if esperado := cuerpoNumerado(indice); !bytes.Equal(esperado, contenido) {
		t.Fatalf("la entrada %d llega a medias o cambiada: %d bytes de %d (FR-032)",
			indice, len(contenido), len(esperado))
	}
}

// deQuienEs lee la clave de prueba y devuelve qué escritor guardó lo leído.
// Falla si no es exactamente lo de ninguno, que sería una mezcla.
func deQuienEs(t *testing.T, lector *Cliente, escritores []escritorDeLaClave) int {
	t.Helper()

	contenido, presente, err := lector.Get(t.Context(), claveDePrueba)
	require.NoError(t, err, "leer mientras otros escriben la misma clave no falla por bloqueo (FR-032)")
	require.True(t, presente, "la clave estaba guardada desde antes de empezar: no puede faltar")

	for indice, escritor := range escritores {
		if bytes.Equal(escritor.contenido, contenido) {
			return indice
		}
	}

	t.Fatalf("lo leído (%d bytes) no es lo que guardó ninguno de los escritores: la entrada quedó mezclada (FR-007)",
		len(contenido))

	return -1
}

// compruebaFilaDe mira la fila de la clave de prueba por debajo del cliente y
// exige que sea una sola y que su expiración salga del mismo escritor que su
// contenido: contenido y vigencia se sustituyen a la vez (FR-007).
func compruebaFilaDe(t *testing.T, ruta string, escritor escritorDeLaClave) {
	t.Helper()

	filas, expiraEn := entradaGuardada(t, ruta, claveDePrueba)
	require.Equal(t, int64(1), filas, "la clave es una sola fila, la escriba quien la escriba")
	require.Equal(t, escritor.instante.Add(vigenciaDePrueba).UnixNano(), expiraEn,
		"la vigencia de la fila es la de la misma escritura que su contenido: nunca una mezcla (FR-007)")
}

// claveNumerada y cuerpoNumerado son la clave y el contenido de la entrada que
// ocupa una posición. Cada contenido repite su propio número a lo largo de
// varias páginas, de modo que ninguno responde por otro y uno leído a medias no
// coincide con el suyo.
func claveNumerada(indice int) string {
	return fmt.Sprintf("%s/%d", claveDePrueba, indice)
}

func cuerpoNumerado(indice int) []byte {
	return bytes.Repeat(fmt.Appendf(nil, "<entrada %04d>", indice), bytesPorPagina/4)
}
