package graph

import (
	"context"
	"database/sql"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cache"
	"github.com/jmorenobl/kitlegal/internal/core"
	"github.com/jmorenobl/kitlegal/internal/core/grafo"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// fechaSiguiente es un instante posterior a fechaDelBloque: la observación
// reciente de las pruebas que llegan fuera de orden.
const fechaSiguiente = "2026-09-30T08:00:00Z"

// loteDelBOE es el lote de una consulta al BOE con la url y la fecha dadas, la
// vigencia de boe articulo y las operaciones.
func loteDelBOE(url, fecha string, operaciones ...schema.Operacion) core.Lote {
	return core.Lote{Fuente: fuenteDelBOE, URL: url, FechaConsulta: fecha, Vigencia: semana, Operaciones: operaciones}
}

// laNorma es el nodo de la norma con los datos dados.
func laNorma(datos map[string]any) schema.Nodo {
	return schema.Nodo{ID: idNorma, Tipo: grafo.TipoNorma, Datos: datos}
}

// operacionesDelBloque son las de una consulta de boe articulo a21 que ve la
// redacción de 2016.
func operacionesDelBloque() []schema.Operacion {
	return unaLectura("a21", "20161002", cuerpo2016).operaciones()
}

// lecturaDePrueba es una consulta de boe articulo sobre un bloque de la LPAC
// que ve una redacción: la lectura de ese bloque que deja su entrega (H7.1
// data-model §2).
type lecturaDePrueba struct {
	// enLaNorma es el id del bloque dentro de la norma, como a21.
	enLaNorma string
	// vigencia es la fecha de vigencia de la redacción vista.
	vigencia string
	// cuerpo es su texto.
	cuerpo string
}

// unaLectura es la del bloque que ve la redacción con la fecha de vigencia y el
// cuerpo dados.
func unaLectura(enLaNorma, vigencia, cuerpo string) lecturaDePrueba {
	return lecturaDePrueba{enLaNorma: enLaNorma, vigencia: vigencia, cuerpo: cuerpo}
}

// bloque es el id del Bloque.
func (l lecturaDePrueba) bloque() string {
	return idNorma + "#" + l.enLaNorma
}

// version es el id de la BloqueVersion que ve.
func (l lecturaDePrueba) version() string {
	return l.bloque() + "@" + l.vigencia + ":" + huella(l.cuerpo)
}

// url es la de la consulta del bloque.
func (l lecturaDePrueba) url() string {
	return urlDeLaNorma + "/texto/bloque/" + l.enLaNorma
}

// operaciones son las de la consulta: la norma, el bloque y la redacción vista,
// las dos aristas que los unen y el texto de la redacción, en un orden que no
// es el de sus claves.
func (l lecturaDePrueba) operaciones() []schema.Operacion {
	return []schema.Operacion{
		schema.Arista{Origen: l.bloque(), Relacion: grafo.RelacionTieneVersion, Destino: l.version()},
		schema.Texto{Huella: huella(l.cuerpo), Cuerpo: l.cuerpo},
		schema.Nodo{ID: l.version(), Tipo: grafo.TipoBloqueVersion, Datos: map[string]any{
			grafo.DatoFechaVigencia: l.vigencia, grafo.DatoHashTexto: huella(l.cuerpo),
		}},
		schema.Arista{Origen: idNorma, Relacion: grafo.RelacionTieneParte, Destino: l.bloque()},
		schema.Nodo{ID: l.bloque(), Tipo: grafo.TipoBloque, Datos: map[string]any{grafo.DatoBloque: l.enLaNorma}},
		laNorma(map[string]any{grafo.DatoIdentificador: "BOE-A-2015-10565"}),
	}
}

// entregada es el lote de la consulta con la fecha dada.
func (l lecturaDePrueba) entregada(fecha string) core.Lote {
	return loteDelBOE(l.url(), fecha, l.operaciones()...)
}

// registrosDe es lo que un lote deja en un grafo vacío: cada registro
// consolidado, con la observación del lote como primera y como última.
func registrosDe(t *testing.T, lote core.Lote) grafoDePrueba {
	t.Helper()

	consolidado, err := grafo.Consolidar(lote)
	require.NoError(t, err)

	return grafoDePrueba{nodos: consolidado.Nodos, aristas: consolidado.Aristas, textos: consolidado.Textos}
}

// primeraYUltima es lo que dejan dos observaciones de las mismas claves en
// instantes distintos, lleguen en el orden que lleguen: de cada nodo y cada
// arista, la primera de la antigua y la última de la reciente, con sus datos y
// su vigencia; de cada texto, la procedencia de la antigua (FR-023). Las dos
// traen las mismas claves, así que sus registros van en el mismo orden.
func primeraYUltima(antigua, reciente grafoDePrueba) grafoDePrueba {
	fusion := grafoDePrueba{textos: antigua.textos}

	for i, nodo := range reciente.nodos {
		nodo.PrimeraObservacion = antigua.nodos[i].PrimeraObservacion
		fusion.nodos = append(fusion.nodos, nodo)
	}

	for i, arista := range reciente.aristas {
		arista.PrimeraObservacion = antigua.aristas[i].PrimeraObservacion
		fusion.aristas = append(fusion.aristas, arista)
	}

	return fusion
}

// grafoGuardado lee todas las filas de world.db, cada tabla ordenada por su
// clave comparando bytes, con una conexión que no deja ningún auxiliar ni
// cambia ningún fichero (research.md V9).
func grafoGuardado(t *testing.T, ruta string) grafoDePrueba {
	t.Helper()

	base := abrirBaseDePrueba(t, ruta, "mode=rw&"+pragmaSoloConsultas)

	var guardado grafoDePrueba

	guardado.nodos = filasGuardadas(t, base, `SELECT id, type, props, `+historiaDeLaPrueba+` FROM nodes ORDER BY id`,
		func(filas *sql.Rows) grafo.RegistroDeNodo {
			var (
				nodo     grafo.RegistroDeNodo
				historia historiaLeida
			)

			require.NoError(t, filas.Scan(append([]any{&nodo.ID, &nodo.Tipo, &nodo.Datos}, historia.destinos()...)...))
			nodo.PrimeraObservacion, nodo.UltimaObservacion, nodo.Vigencia = historia.primera, historia.ultima,
				historia.vigencia()

			return nodo
		})

	guardado.aristas = filasGuardadas(t, base,
		`SELECT src, rel, dst, `+historiaDeLaPrueba+` FROM edges ORDER BY src, rel, dst`,
		func(filas *sql.Rows) grafo.RegistroDeArista {
			var (
				arista   grafo.RegistroDeArista
				historia historiaLeida
			)

			require.NoError(t, filas.Scan(append([]any{&arista.Origen, &arista.Relacion, &arista.Destino},
				historia.destinos()...)...))
			arista.PrimeraObservacion, arista.UltimaObservacion, arista.Vigencia = historia.primera, historia.ultima,
				historia.vigencia()

			return arista
		})

	guardado.textos = filasGuardadas(t, base, `SELECT hash, body, fetched_at, source, url FROM texts ORDER BY hash`,
		func(filas *sql.Rows) grafo.RegistroDeTexto {
			var texto grafo.RegistroDeTexto

			require.NoError(t, filas.Scan(&texto.Huella, &texto.Cuerpo, &texto.Procedencia.FechaConsulta,
				&texto.Procedencia.Fuente, &texto.Procedencia.URL))

			return texto
		})

	require.NoError(t, base.Close())

	return guardado
}

// historiaDeLaPrueba son los campos de la historia de un nodo o una arista en
// el orden en que los lee historiaLeida.
const historiaDeLaPrueba = `first_seen, first_source, first_url, last_seen, source, url, ttl`

// historiaLeida es la historia de un nodo o una arista leída de su fila: la
// primera y la última observación y la vigencia en segundos.
type historiaLeida struct {
	primera, ultima grafo.Procedencia
	ttl             sql.NullInt64
}

// destinos son dónde se leen los campos de historiaDeLaPrueba.
func (h *historiaLeida) destinos() []any {
	return []any{
		&h.primera.FechaConsulta, &h.primera.Fuente, &h.primera.URL,
		&h.ultima.FechaConsulta, &h.ultima.Fuente, &h.ultima.URL, &h.ttl,
	}
}

// vigencia es la de la columna ttl, en segundos; NULL es cero.
func (h historiaLeida) vigencia() time.Duration {
	return vigenciaGuardada(h.ttl)
}

// filasGuardadas lee cada fila de la consulta con leer; nil si no hay ninguna,
// como los registros de un lote consolidado sin esa clase de operación.
func filasGuardadas[T any](t *testing.T, base *sql.DB, consulta string, leer func(*sql.Rows) T) []T {
	t.Helper()

	filas, err := base.QueryContext(t.Context(), consulta)
	require.NoError(t, err)

	defer func() { require.NoError(t, filas.Close()) }()

	var lista []T

	for filas.Next() {
		lista = append(lista, leer(filas))
	}

	require.NoError(t, filas.Err())

	return lista
}

// lecturasGuardadas son las filas de lecturas de world.db, por bloque comparando
// bytes, leídas con una conexión que no deja ningún auxiliar ni cambia ningún
// fichero; nil si no hay ninguna.
func lecturasGuardadas(t *testing.T, ruta string) []grafo.LecturasDeBloque {
	t.Helper()

	base := abrirBaseDePrueba(t, ruta, "mode=rw&"+pragmaSoloConsultas)
	lecturas := filasGuardadas(t, base, `SELECT bloque, ultima, anterior FROM lecturas ORDER BY bloque`,
		func(filas *sql.Rows) grafo.LecturasDeBloque {
			var fila grafo.LecturasDeBloque

			require.NoError(t, filas.Scan(&fila.Bloque, &fila.Ultima, &fila.Anterior))

			return fila
		})
	require.NoError(t, base.Close())

	return lecturas
}

// compruebaFalloDeEntrega compara el fallo de una entrega con su clase, su ruta
// y su mensaje exacto.
func compruebaFalloDeEntrega(t *testing.T, err error, clase schema.Clase, ruta, mensaje string) {
	t.Helper()

	var fallo *Error

	require.ErrorAs(t, err, &fallo)
	assert.Equal(t, clase, fallo.Clase())
	assert.Equal(t, operacionEscribir, fallo.Operacion)
	assert.Equal(t, ruta, fallo.Ruta)
	assert.Equal(t, mensaje, err.Error())
}

// TestNuevo fija que construir el almacén no resuelve la ruta ni crea nada: una
// opción que no da ningún directorio falla al entregar, con «argumentos», y las
// opciones son las de cuando se construye (contracts/almacen-world-db.md
// §1 y §2; FR-001, FR-011).
func TestNuevo(t *testing.T) {
	t.Parallel()

	t.Run("no resuelve ni crea nada al construirse", func(t *testing.T) {
		t.Parallel()

		raiz := t.TempDir()
		antes := huellasDelArbol(t, raiz)

		assert.NotNil(t, Nuevo(ConDirectorio(filepath.Join(raiz, "no-existe"))))
		assert.NotNil(t, Nuevo(ConDirectorio("")))
		assert.Equal(t, antes, huellasDelArbol(t, raiz))
	})

	t.Run("la opción sin directorio falla al entregar", func(t *testing.T) {
		t.Parallel()

		err := Nuevo(ConDirectorio("")).Apply(t.Context(), loteDelBOE(urlDeLaNorma, fechaDelBloque, laNorma(nil)))

		var fallo *Error

		require.ErrorAs(t, err, &fallo)
		assert.Equal(t, schema.ClaseArgumentos, fallo.Clase())
		assert.Equal(t, operacionEscribir, fallo.Operacion)
		assert.True(t, strings.HasPrefix(err.Error(), "grafo: no se puede ubicar world.db: "), err.Error())
		require.ErrorIs(t, err, errDirectorioVacio)
	})

	t.Run("las opciones son las de la construcción", func(t *testing.T) {
		t.Parallel()

		raiz := t.TempDir()
		opciones := []Opcion{ConDirectorio(filepath.Join(raiz, "declarado"))}
		almacen := Nuevo(opciones...)
		opciones[0] = ConDirectorio(filepath.Join(raiz, "cambiado"))

		require.NoError(t, almacen.Apply(t.Context(), loteDelBOE(urlDeLaNorma, fechaDelBloque, laNorma(nil))))
		assert.FileExists(t, filepath.Join(raiz, "declarado", "world.db"))
		assert.NoDirExists(t, filepath.Join(raiz, "cambiado"))
	})
}

// TestApplySinOpcion fija que sin ConDirectorio world.db vive donde dice la
// regla de la caché, resuelta al entregar (FR-001, FR-011), y que el almacén
// cero y el nulo entregan igual que el de Nuevo. No es paralela: cambia el
// entorno.
func TestApplySinOpcion(t *testing.T) {
	raiz := t.TempDir()
	directorio := filepath.Join(raiz, "cache")
	lote := loteDelBOE(urlDelBloque, fechaDelBloque, operacionesDelBloque()...)

	t.Setenv(cache.VariableDirectorio, directorio)

	var nulo *Almacen

	for _, almacen := range []*Almacen{Nuevo(), {}, nulo} {
		require.NoError(t, almacen.Apply(t.Context(), lote))
		assert.Equal(t, registrosDe(t, lote), grafoGuardado(t, filepath.Join(directorio, "world.db")))
	}

	t.Setenv(cache.VariableDirectorio, "")

	var fallo *Error

	require.ErrorAs(t, Nuevo().Apply(t.Context(), lote), &fallo)
	assert.Equal(t, schema.ClaseArgumentos, fallo.Clase())
}

// TestApplyIdempotente fija FR-022 y SC-001: aplicar el mismo lote 2 y 10 veces
// deja un nodo por id, una arista por terna y un texto por huella, con la misma
// primera y última observación que tras la primera vez; y, como una observación
// idéntica no cambia nada, no vuelve a escribir: ningún fichero del directorio
// cambia ni aparece.
func TestApplyIdempotente(t *testing.T) {
	t.Parallel()

	for _, veces := range []int{2, 10} {
		t.Run(strconv.Itoa(veces)+" veces", func(t *testing.T) {
			t.Parallel()

			directorio := t.TempDir()
			ruta := filepath.Join(directorio, "world.db")
			almacen := Nuevo(ConDirectorio(directorio))
			lote := loteDelBOE(urlDelBloque, fechaDelBloque, operacionesDelBloque()...)

			require.NoError(t, almacen.Apply(t.Context(), lote))

			tras1 := huellasDelArbol(t, directorio)
			assert.Equal(t, registrosDe(t, lote), grafoGuardado(t, ruta))

			for range veces - 1 {
				require.NoError(t, almacen.Apply(t.Context(), lote))
				assert.Equal(t, tras1, huellasDelArbol(t, directorio), "la entrega repetida no escribe nada")
			}

			assert.Equal(t, registrosDe(t, lote), grafoGuardado(t, ruta))

			lectura, err := Leer(t.Context(), ConDirectorio(directorio))
			require.NoError(t, err)

			recuento, err := lectura.Recuento(t.Context())
			require.NoError(t, err)
			require.NoError(t, lectura.Close())

			assert.Equal(t, 3, recuento.Nodos)
			assert.Equal(t, 2, recuento.Aristas)
			assert.Equal(t, 1, recuento.Textos)
		})
	}
}

// TestApplyLecturas fija las filas de lecturas que deja la entrega (H7.1
// FR-020, FR-021, FR-026; data-model §1 y §4; contracts/almacen-world-db.md
// §4, pasos 5 a 8): la primera lectura de un bloque nuevo guarda (v, v); cada
// lectura siguiente pasa la última a la anterior —(B, A) y, al leer B otra
// vez, (B, B)— y una entrega idéntica a la anterior ya no escribe nada; boe
// articulos deja una fila por bloque; la primera lectura de un bloque que el
// grafo de H7 ya observó parte de su redacción vista sin lecturas, calculada
// sobre lo guardado antes del lote —(v, R)—; y un lote rechazado no deja
// ninguna fila.
func TestApplyLecturas(t *testing.T) {
	t.Parallel()

	a := unaLectura("a21", "20161002", cuerpo2016)
	b := unaLectura("a21", "20250101", cuerpo2025)

	t.Run("una lectura tras otra", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		ruta := filepath.Join(directorio, "world.db")
		almacen := Nuevo(ConDirectorio(directorio))

		require.NoError(t, almacen.Apply(t.Context(), a.entregada(fechaDelBloque)))
		assert.Equal(t, []grafo.LecturasDeBloque{filaDeLecturas(a, a)}, lecturasGuardadas(t, ruta),
			"la primera, de un bloque nuevo: (A, A)")

		require.NoError(t, almacen.Apply(t.Context(), b.entregada(fechaSiguiente)))
		assert.Equal(t, []grafo.LecturasDeBloque{filaDeLecturas(b, a)}, lecturasGuardadas(t, ruta),
			"otra que ve una redacción nueva: (B, A)")

		require.NoError(t, almacen.Apply(t.Context(), b.entregada(fechaSiguiente)))
		assert.Equal(t, []grafo.LecturasDeBloque{filaDeLecturas(b, b)}, lecturasGuardadas(t, ruta),
			"otra que ve B, la sirva la caché con su misma consulta: (B, B)")

		antes := huellasDelArbol(t, directorio)

		require.NoError(t, almacen.Apply(t.Context(), b.entregada(fechaSiguiente)))
		assert.Equal(t, antes, huellasDelArbol(t, directorio), "con (B, B), leer B otra vez no escribe nada")
	})

	t.Run("boe articulos deja una fila por bloque", func(t *testing.T) {
		t.Parallel()

		a22 := unaLectura("a22", "20161002", cuerpo2025)
		directorio := t.TempDir()

		require.NoError(t, Nuevo(ConDirectorio(directorio)).Apply(t.Context(),
			loteDelBOE(urlDeLaNorma, fechaDelBloque, slices.Concat(a.operaciones(), a22.operaciones())...)))
		assert.Equal(t, []grafo.LecturasDeBloque{filaDeLecturas(a, a), filaDeLecturas(a22, a22)},
			lecturasGuardadas(t, filepath.Join(directorio, "world.db")))
	})

	t.Run("la primera sobre un bloque que el grafo de H7 ya observó", func(t *testing.T) {
		t.Parallel()

		m := laMuestra(t)
		require.Equal(t, a.version(), m.v2016.ID, "premisa: la lectura ve la redacción de 2016 de la muestra")
		require.Less(t, m.v2016.UltimaObservacion.FechaConsulta, m.v2025.UltimaObservacion.FechaConsulta,
			"premisa: antes del lote, la redacción observada la última es la de 2025")

		directorio := conMuestraDeH7(t, t.TempDir())
		ruta := filepath.Join(directorio, "world.db")

		require.NoError(t, Nuevo(ConDirectorio(directorio)).Apply(t.Context(), a.entregada(fechaSiguiente)))
		assert.Equal(t, []grafo.LecturasDeBloque{{Bloque: idBloque, Ultima: a.version(), Anterior: m.v2025.ID}},
			lecturasGuardadas(t, ruta), "(A, R): R es la de 2025, no la que el lote vuelve a observar")

		base := abrirBaseDePrueba(t, ruta, "mode=rw&"+pragmaSoloConsultas)
		assert.Equal(t, []int64{1, 2}, versionesDe(t, base), "la entrega migra la base de H7 a la 2")
		require.NoError(t, base.Close())
	})

	t.Run("un lote rechazado no deja ninguna fila", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()
		ruta := filepath.Join(directorio, "world.db")
		almacen := Nuevo(ConDirectorio(directorio))
		require.NoError(t, almacen.Apply(t.Context(), a.entregada(fechaDelBloque)))

		antes := huellasDelArbol(t, directorio)

		// La lectura de B con la norma de otro tipo que el guardado: la fusión
		// la rechaza dentro de la transacción, con la fila de B ya calculada.
		operaciones := b.operaciones()
		operaciones[len(operaciones)-1] = schema.Nodo{ID: idNorma, Tipo: grafo.TipoMunicipio}

		err := almacen.Apply(t.Context(), loteDelBOE(b.url(), fechaSiguiente, operaciones...))

		var rechazo *grafo.Rechazo

		require.ErrorAs(t, err, &rechazo)
		assert.Equal(t, []grafo.LecturasDeBloque{filaDeLecturas(a, a)}, lecturasGuardadas(t, ruta))
		assert.Equal(t, antes, huellasDelArbol(t, directorio), "el grafo queda como estaba, byte a byte")
	})
}

// filaDeLecturas es la fila del bloque de las dos lecturas: la última redacción
// vista y la anterior.
func filaDeLecturas(ultima, anterior lecturaDePrueba) grafo.LecturasDeBloque {
	return grafo.LecturasDeBloque{Bloque: ultima.bloque(), Ultima: ultima.version(), Anterior: anterior.version()}
}

// aplicarEnOrden entrega los lotes, uno tras otro, en un directorio nuevo y
// devuelve lo que queda guardado.
func aplicarEnOrden(t *testing.T, lotes ...core.Lote) grafoDePrueba {
	t.Helper()

	directorio := t.TempDir()
	almacen := Nuevo(ConDirectorio(directorio))

	for _, lote := range lotes {
		require.NoError(t, almacen.Apply(t.Context(), lote))
	}

	return grafoGuardado(t, filepath.Join(directorio, "world.db"))
}

// TestApplyFueraDeOrden fija FR-023 sobre world.db: dos observaciones de las
// mismas claves en instantes distintos dejan lo mismo lleguen en el orden que
// lleguen —de cada nodo y arista, la primera de la antigua y la última de la
// reciente; de cada texto, la procedencia de la antigua—, y una reobservación
// con otros datos los sustituye enteros, sin mezclar claves, con la vigencia
// que declara (FR-065).
func TestApplyFueraDeOrden(t *testing.T) {
	t.Parallel()

	t.Run("mismas claves en dos instantes", func(t *testing.T) {
		t.Parallel()

		antigua := loteDelBOE(urlDelBloque, fechaDelBloque, operacionesDelBloque()...)
		reciente := loteDelBOE(urlDelBloque, fechaSiguiente, operacionesDelBloque()...)
		esperado := primeraYUltima(registrosDe(t, antigua), registrosDe(t, reciente))

		assert.Equal(t, esperado, aplicarEnOrden(t, antigua, reciente), "en orden")
		assert.Equal(t, esperado, aplicarEnOrden(t, reciente, antigua), "fuera de orden")
	})

	t.Run("una reobservación con otros datos y sin vigencia", func(t *testing.T) {
		t.Parallel()

		antigua := loteDelBOE(urlDeLaNorma, fechaDelBloque,
			laNorma(map[string]any{grafo.DatoIdentificador: "BOE-A-2015-10565", "anterior": true}))
		reciente := loteDelBOE(urlDeLaNorma, fechaSiguiente, laNorma(map[string]any{grafo.DatoIdentificador: "BOE-A-2015-10565"}))
		reciente.Vigencia = 0

		esperado := primeraYUltima(registrosDe(t, antigua), registrosDe(t, reciente))
		require.JSONEq(t, `{"identificador":"BOE-A-2015-10565"}`, esperado.nodos[0].Datos)
		require.Zero(t, esperado.nodos[0].Vigencia)

		assert.Equal(t, esperado, aplicarEnOrden(t, antigua, reciente), "en orden")
		assert.Equal(t, esperado, aplicarEnOrden(t, reciente, antigua), "fuera de orden")
	})
}

// TestApplyEmpates fija el desempate de H7.1 FR-076 sobre world.db, en los dos
// órdenes de llegada: con el mismo instante, gana la url menor, aunque la fecha
// esté escrita con otro desplazamiento; lo que queda es la observación
// ganadora entera, como si hubiera llegado sola.
func TestApplyEmpates(t *testing.T) {
	t.Parallel()

	gana := loteDelBOE(urlDeLaNorma, fechaConDesplazamiento, operacionesDelBloque()...)
	pierde := loteDelBOE(urlDelBloque, fechaDelBloque, operacionesDelBloque()...)
	esperado := registrosDe(t, gana)

	assert.Equal(t, esperado, aplicarEnOrden(t, gana, pierde), "la que gana llega primero")
	assert.Equal(t, esperado, aplicarEnOrden(t, pierde, gana), "la que gana llega después")
}

// grafoConElBloque deja en un directorio nuevo el grafo de una consulta de boe
// articulo a21 y devuelve el directorio y la ruta de world.db.
func grafoConElBloque(t *testing.T) (string, string) {
	t.Helper()

	directorio := t.TempDir()
	require.NoError(t, Nuevo(ConDirectorio(directorio)).Apply(t.Context(),
		loteDelBOE(urlDelBloque, fechaDelBloque, operacionesDelBloque()...)))

	return directorio, filepath.Join(directorio, "world.db")
}

// TestApplyRechazaContraLoGuardado fija lo que solo se sabe dentro de la
// transacción (FR-024; H7.1 FR-075; contracts/almacen-world-db.md §4, paso 6,
// y §5): un id con otro tipo que el guardado y una huella guardada con otro
// cuerpo rechazan el lote entero, también lo que en él venía bien, con
// «inesperado» y un mensaje que nombra world.db y el motivo. world.db queda
// con los mismos bytes y sin auxiliares (V11; §4.1).
func TestApplyRechazaContraLoGuardado(t *testing.T) {
	t.Parallel()

	huellaDe2025 := huella(cuerpo2025)
	nuevo := schema.Nodo{ID: idMunicipio, Tipo: grafo.TipoMunicipio, Datos: map[string]any{grafo.DatoCodigoINE: "28074"}}

	casos := []struct {
		nombre string
		// preparar cambia lo guardado antes de entregar; nil si no hace falta.
		preparar func(t *testing.T, ruta string)
		lote     core.Lote
		motivo   string
	}{
		{
			nombre: "un id con otro tipo que el guardado",
			lote:   loteDelBOE(urlDeLaNorma, fechaSiguiente, nuevo, schema.Nodo{ID: idNorma, Tipo: grafo.TipoMunicipio}),
			motivo: `el nodo "` + idNorma + `": el grafo ya lo tiene con el tipo "Norma" y el lote le da el tipo "Municipio"`,
		},
		{
			nombre: "una huella guardada con otro cuerpo",
			preparar: func(t *testing.T, ruta string) {
				t.Helper()

				alterar(t, ruta, `INSERT INTO texts VALUES ('`+huellaDe2025+`', 'otro cuerpo', '`+fechaDelBloque+
					`', '`+fuenteDelBOE+`', '`+urlDelBloque+`')`)
			},
			lote:   loteDelBOE(urlDelBloque, fechaSiguiente, nuevo, schema.Texto{Huella: huellaDe2025, Cuerpo: cuerpo2025}),
			motivo: `el texto "` + huellaDe2025 + `": el grafo ya guarda otro cuerpo con esa huella`,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			directorio, ruta := grafoConElBloque(t)

			if caso.preparar != nil {
				caso.preparar(t, ruta)
			}

			antes := huellasDelArbol(t, directorio)

			err := Nuevo(ConDirectorio(directorio)).Apply(t.Context(), caso.lote)
			compruebaFalloDeEntrega(t, err, schema.ClaseInesperado, ruta,
				"grafo: el lote no entra en "+strconv.Quote(ruta)+": "+caso.motivo)

			var rechazo *grafo.Rechazo

			require.ErrorAs(t, err, &rechazo)
			assert.Equal(t, antes, huellasDelArbol(t, directorio), "el grafo queda como estaba, byte a byte")
		})
	}
}

// TestApplyRechazaElLote fija lo que se rechaza sin tocar el disco (FR-024,
// FR-025; SC-010; H7.1 FR-074, FR-075; contracts/almacen-world-db.md §4, paso
// 1, y §5): con world.db
// ausente no se crea nada, y con world.db presente no cambia ni aparece nada. El
// mensaje nombra world.db, que todavía no tiene ruta, y el motivo del dominio,
// que no repite el id de una Persona ni el cuerpo de un texto.
func TestApplyRechazaElLote(t *testing.T) {
	t.Parallel()

	persona := func(id string, datos map[string]any) schema.Nodo {
		return schema.Nodo{ID: id, Tipo: grafo.TipoPersona, Datos: datos}
	}

	sinFuente := loteDelBOE(urlDeLaNorma, fechaDelBloque, laNorma(nil))
	sinFuente.Fuente = ""

	casos := map[string]core.Lote{
		"sin fuente": sinFuente,
		"un id con dos tipos en el lote": loteDelBOE(urlDeLaNorma, fechaDelBloque, laNorma(nil),
			schema.Nodo{ID: idNorma, Tipo: grafo.TipoBloque}),
		"un texto con la huella de otro cuerpo": loteDelBOE(urlDelBloque, fechaDelBloque,
			schema.Texto{Huella: huella(cuerpo2016), Cuerpo: cuerpo2025}),
		"una Persona con un DNI en su id": loteDelBOE(urlDeLaNorma, fechaDelBloque, persona("12345678Z", nil)),
		"una Persona con un NIE en un valor": loteDelBOE(urlDeLaNorma, fechaDelBloque,
			persona("ana-garcia-lopez", map[string]any{"documento": "X-1234567-L"})),
		"una Persona con un NIF dentro de una lista anidada": loteDelBOE(urlDeLaNorma, fechaDelBloque,
			persona("ana-garcia-lopez", map[string]any{"contacto": map[string]any{"documentos": []any{"B-12.345.678"}}})),
		"una Persona con un DNI como clave": loteDelBOE(urlDeLaNorma, fechaDelBloque,
			persona("ana-garcia-lopez", map[string]any{"12 345 678 z": "nombre"})),
	}

	for nombre, lote := range casos {
		t.Run(nombre, func(t *testing.T) {
			t.Parallel()

			for _, preparar := range []func(*testing.T, string) string{cacheVacia, conMuestra} {
				raiz := t.TempDir()
				directorio := preparar(t, raiz)
				antes := huellasDelArbol(t, raiz)

				err := Nuevo(ConDirectorio(directorio)).Apply(t.Context(), lote)

				var rechazo *grafo.Rechazo

				require.ErrorAs(t, err, &rechazo)
				compruebaFalloDeEntrega(t, err, schema.ClaseInesperado, "", "grafo: el lote no entra en world.db: "+rechazo.Error())
				assert.NotContains(t, err.Error(), "12345678Z")
				assert.NotContains(t, err.Error(), cuerpo2025)
				assert.Equal(t, antes, huellasDelArbol(t, raiz), "no se toca el disco")
			}
		})
	}

	t.Run("una Persona sin documento entra", func(t *testing.T) {
		t.Parallel()

		lote := loteDelBOE(urlDeLaNorma, fechaDelBloque,
			persona("ana-garcia-lopez", map[string]any{"nombre": "Ana Garc\xc3\xada L\xc3\xb3pez", "nacida": "1990-01-01"}))

		assert.Equal(t, registrosDe(t, lote), aplicarEnOrden(t, lote))
	})
}

// TestApplyCreaEnSuSitio fija la entrega sobre world.db ausente
// (contracts/almacen-world-db.md §4, pasos 2 y 3; H7.1 FR-071; H7 FR-001,
// FR-003, FR-013): crea cada directorio que falta, en 0700, y world.db en su
// sitio, en 0600, en WAL, con el esquema y el lote; en el directorio no queda
// nada más, ni un temporal ni un auxiliar.
func TestApplyCreaEnSuSitio(t *testing.T) {
	t.Parallel()

	raiz := t.TempDir()
	directorio := filepath.Join(raiz, "a", "b")
	ruta := filepath.Join(directorio, "world.db")
	lote := loteDelBOE(urlDelBloque, fechaDelBloque, operacionesDelBloque()...)

	require.NoError(t, Nuevo(ConDirectorio(directorio)).Apply(t.Context(), lote))

	for _, cada := range []struct {
		ruta     string
		permisos fs.FileMode
	}{
		{ruta: filepath.Join(raiz, "a"), permisos: fs.ModeDir | 0o700},
		{ruta: directorio, permisos: fs.ModeDir | 0o700},
		{ruta: ruta, permisos: 0o600},
	} {
		info, err := os.Stat(cada.ruta)
		require.NoError(t, err)
		assert.Equal(t, cada.permisos, info.Mode(), "los permisos de %s", cada.ruta)
	}

	assert.Equal(t, []string{"world.db"}, nombresEn(t, directorio), "ni un temporal ni un auxiliar")
	assert.Equal(t, []byte{2, 2}, leerFichero(t, ruta)[18:20], "world.db nace en WAL")
	assert.Equal(t, registrosDe(t, lote), grafoGuardado(t, ruta))

	base := abrirBaseDePrueba(t, ruta, "mode=rw&"+pragmaSoloConsultas)
	compruebaEsquema(t, base)
	assert.Equal(t, int64(2), versionDe(t, base))
	require.NoError(t, base.Close())
}

// TestApplyEnSuSitio fija la entrega sobre un world.db sin esquema, el que deja
// una creación interrumpida (contracts/almacen-world-db.md §4; H7.1 FR-071,
// FR-077; H7 FR-004, FR-013): de 0 bytes, o ya en WAL y sin ninguna tabla, la
// entrega siguiente lo pone en WAL si no lo está, crea el esquema y aplica el
// lote, sin dejar ningún auxiliar.
func TestApplyEnSuSitio(t *testing.T) {
	t.Parallel()

	lote := loteDelBOE(urlDelBloque, fechaDelBloque, operacionesDelBloque()...)

	for nombre, preparar := range map[string]func(*testing.T, string) string{
		"de 0 bytes":          conBase(nil),
		"en WAL y sin tablas": enWALSinTablas,
	} {
		t.Run(nombre, func(t *testing.T) {
			t.Parallel()

			directorio := preparar(t, t.TempDir())
			ruta := filepath.Join(directorio, "world.db")

			require.NoError(t, Nuevo(ConDirectorio(directorio)).Apply(t.Context(), lote))

			assert.Equal(t, registrosDe(t, lote), grafoGuardado(t, ruta))
			assert.Equal(t, []byte{2, 2}, leerFichero(t, ruta)[18:20], "world.db queda en WAL")

			base := abrirBaseDePrueba(t, ruta, "mode=rw&"+pragmaSoloConsultas)
			assert.Equal(t, int64(2), versionDe(t, base))
			require.NoError(t, base.Close())

			assert.Equal(t, []string{"world.db"}, nombresEn(t, directorio), "no queda ningún auxiliar")
		})
	}
}

// enWALSinTablas prepara world.db como lo deja una creación que se interrumpe
// después de ponerlo en WAL (contracts/almacen-world-db.md §4, pasos 3 y 4): el
// fichero que crea la entrega, en WAL y sin ninguna tabla.
func enWALSinTablas(t *testing.T, raiz string) string {
	t.Helper()

	directorio := conBase(nil)(t, raiz)
	ruta := filepath.Join(directorio, "world.db")

	base, err := abrirConexion(operacionEscribir, ruta, cadenaDeEscritura(ruta))
	require.NoError(t, err)
	require.NoError(t, fijarWAL(t.Context(), base, ruta))
	require.NoError(t, base.Close())
	require.Equal(t, []byte{2, 2}, leerFichero(t, ruta)[18:20], "premisa: world.db está en WAL")

	return directorio
}

// TestApplyNoModifica fija las entregas que fallan sobre un world.db que el
// binario no puede usar (contracts/almacen-world-db.md §4, paso 4, y §6): lo que
// no es una base de datos sigue la regla genérica —«inesperado», con su ruta y
// la causa en el mensaje, y nada se promete sobre sus bytes (H7.1 FR-070)—, y un
// esquema posterior no se modifica (H7 FR-012): nada cambia ni aparece.
func TestApplyNoModifica(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre   string
		preparar func(*testing.T, string) string
		mensaje  func(ruta string, fallo *Error) string
		// intacto dice que la entrega no cambia ni crea ningún fichero.
		intacto bool
	}{
		{
			nombre:   "lo que no es una base de datos",
			preparar: conBase(contenido),
			mensaje: func(ruta string, fallo *Error) string {
				return "grafo: " + strconv.Quote(ruta) + " no es una base de datos utilizable: " + fallo.Causa.Error()
			},
		},
		{
			nombre:   "un esquema posterior",
			preparar: conSentencias(versionPosterior),
			mensaje: func(ruta string, _ *Error) string {
				return "grafo: " + strconv.Quote(ruta) +
					" tiene el esquema en la versi\xc3\xb3n 3 y este binario conoce la 2: no se modifica"
			},
			intacto: true,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			raiz := t.TempDir()
			directorio := caso.preparar(t, raiz)
			ruta := filepath.Join(directorio, "world.db")
			antes := huellasDelArbol(t, raiz)

			err := Nuevo(ConDirectorio(directorio)).Apply(t.Context(),
				loteDelBOE(urlDelBloque, fechaDelBloque, operacionesDelBloque()...))

			var fallo *Error

			require.ErrorAs(t, err, &fallo)
			compruebaFalloDeEntrega(t, err, schema.ClaseInesperado, ruta, caso.mensaje(ruta, fallo))

			if caso.intacto {
				assert.Equal(t, antes, huellasDelArbol(t, raiz), "nada cambia ni aparece")
			}
		})
	}
}

// TestApplyConElPlazoAgotado fija que la entrega atiende el contexto (FR-014;
// contracts/almacen-world-db.md §4, pasos 1 y 5, y §6): terminado antes de
// entregar, no toca nada; mientras otra invocación retiene world.db, sale con
// «fuente-no-disponible» sin agotar la espera propia y el grafo queda como
// estaba.
func TestApplyConElPlazoAgotado(t *testing.T) {
	t.Parallel()

	t.Run("el contexto terminado antes de entregar", func(t *testing.T) {
		t.Parallel()

		raiz := t.TempDir()
		antes := huellasDelArbol(t, raiz)

		terminado, cancela := context.WithCancel(t.Context())
		cancela()

		err := Nuevo(ConDirectorio(filepath.Join(raiz, "cache"))).Apply(terminado,
			loteDelBOE(urlDeLaNorma, fechaDelBloque, laNorma(nil)))
		compruebaFalloDeEntrega(t, err, schema.ClaseFuenteNoDisponible, "",
			"grafo: el plazo termin\xc3\xb3 antes de escribir world.db")
		require.ErrorIs(t, err, context.Canceled)
		assert.Equal(t, antes, huellasDelArbol(t, raiz))
	})

	t.Run("otra invocación retiene world.db", func(t *testing.T) {
		t.Parallel()

		directorio, ruta := grafoConElBloque(t)
		antes := grafoGuardado(t, ruta)
		suelta := retenerElBloqueo(t, ruta)

		conPlazo, cancela := context.WithTimeout(t.Context(), 300*time.Millisecond)
		defer cancela()

		inicio := time.Now()
		err := Nuevo(ConDirectorio(directorio)).Apply(conPlazo,
			loteDelBOE(urlDelBloque, fechaSiguiente, operacionesDelBloque()...))
		tardo := time.Since(inicio)

		compruebaFalloDeEntrega(t, err, schema.ClaseFuenteNoDisponible, ruta,
			"grafo: el plazo termin\xc3\xb3 antes de escribir "+strconv.Quote(ruta))
		assert.Less(t, tardo, esperaPropia, "el plazo llega antes que la espera propia")

		suelta()
		assert.Equal(t, antes, grafoGuardado(t, ruta))
	})
}

// nombresEn son los nombres de las entradas del directorio, ordenados.
func nombresEn(t *testing.T, directorio string) []string {
	t.Helper()

	entradas, err := os.ReadDir(directorio)
	require.NoError(t, err)

	nombres := make([]string, 0, len(entradas))

	for _, entrada := range entradas {
		nombres = append(nombres, entrada.Name())
	}

	return nombres
}
