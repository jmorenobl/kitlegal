# T003 · intento 1: redelimitada

## Por qué no queda en verde en este intento

T003 estaba mal delimitada: su cambio obligado de `errorInutilizable` (el mensaje pasa a
`grafo: "<ruta>" no es una base de datos utilizable: <causa>`, sin «; no se modifica» ni «: acceso denegado»)
toca un constructor que comparten la lectura y la escritura, y la creación en su sitio con
`os.OpenFile(…, O_RDWR|O_CREATE, 0o600)` sigue un enlace simbólico sin destino y crea el destino. Con el código de
T003, tres ficheros de test que la tarea no declaraba quedan en rojo:

- `internal/graph/lectura_test.go` y `internal/graph/migraciones_test.go`, que esperan el mensaje de H7.
  `go test -count=1 ./internal/graph/` sobre el árbol de este intento:

  ```
  --- FAIL: TestLeerSinPermisoDeLectura (0.00s)
  --- FAIL: TestLeerReabreInmutable (0.00s)
      --- FAIL: TestLeerReabreInmutable/con_un_-wal_y_sin_su_-shm:_inutilizable (0.01s)
  --- FAIL: TestMigrar (0.00s)
      --- FAIL: TestMigrar/una_tabla_schema_version_ajena_no_es_una_base_utilizable (0.00s)
  --- FAIL: TestLecturaDeFilasDanadas (0.00s)
      --- FAIL: TestLecturaDeFilasDanadas/datos_que_no_son_JSON,_en_la_ficha (0.01s)
      --- FAIL: TestLecturaDeFilasDanadas/una_vigencia_que_no_cabe_en_una_duración (0.01s)
      --- FAIL: TestLecturaDeFilasDanadas/una_vigencia_negativa (0.01s)
      --- FAIL: TestLecturaDeFilasDanadas/datos_que_no_son_JSON,_en_la_instantánea (0.01s)
  --- FAIL: TestLeerEstados (0.00s)
      --- FAIL: TestLeerEstados/una_base_dañada (0.00s)
      --- FAIL: TestLeerEstados/lo_que_no_es_una_base,_sin_permiso_de_escritura (0.00s)
      --- FAIL: TestLeerEstados/lo_que_no_es_una_base_de_datos (0.00s)
  FAIL	github.com/jmorenobl/kitlegal/internal/graph
  ```

- `internal/graph/integracion_enlace_test.go`: su subtest «sin destino» afirma que la entrega falla sin crear el
  destino del enlace (falso con la creación en su sitio) y usa `compruebaNoSePuedeEscribir`, que sale de
  `integracion_test.go` con `TestIntegracionSinPermisoDeEscritura`. El paquete de integración no compila:

  ```
  vet: internal/graph/integracion_enlace_test.go:61:3: undefined: compruebaNoSePuedeEscribir
  ```

La línea de T003 en tasks.md declara ahora esas tres rutas y lo que cambia en cada una. No se tocó ninguna en el
árbol del repositorio.

## Lo que ya está en el árbol (lo hereda el intento 2)

Todo lo de las rutas declaradas en el intento 1, con `go test ./internal/app/ ./internal/evals/ ./internal/core/...`
en verde sobre este árbol:

- `internal/graph/publicar.go` y `publicar_test.go` retirados (`git rm`).
- `almacen.go`: `Almacen` sin `enlazar` ni `enlazador`; `Apply` consolida, mira el contexto, ubica y llama a
  `aplicarEnSuSitio`, sin `os.Stat` ni caso de directorio ni de fichero ausente.
- `aplicar.go`: `crearEnSuSitio` (pasos 2 y 3: `os.MkdirAll(dir, 0o700)` → `errorDeDirectorioNoEscribible`;
  `os.OpenFile(world.db, O_RDWR|O_CREATE, 0o600)` y cerrar → `errorDeEntradaSalida`, «no se pudo escribir»);
  `abrirParaEscribir` y `prepararLaBase` sin el lote, sin `comprobarEscritura` ni `ValidarContraGrafoVacio`; fuera
  `errFilaDanada` (un fallo al fusionar con lo guardado lo clasifica `falloAlEscribir` como `errorDeEntradaSalida`)
  y el método `historiaGuardada.vigencia`, que solo lo envolvía; `permisosDeDirectorio` y `permisosDeFichero` pasan
  aquí desde `publicar.go`.
- `errores.go`: fuera `errorDeFicheroNoEscribible` y `errorDePublicacion`; `errorInutilizable` con el mensaje
  nuevo (sin causa, que solo pasan las guardas de una versión negativa, sin «: »).
- `doc.go`: la creación en su sitio y la regla genérica sin promesa sobre los bytes.
- Tests: `TestApplyCreaEnSuSitio` (lo que afirmaba `TestPublicar` del resultado); `TestApplyEnSuSitio` con 0 bytes y
  «en WAL y sin tablas»; `TestApplyNoModifica` con «no es una base» (sin la aserción de que no cambia) y el esquema
  posterior; fuera `TestApplySinPermisoDeEscritura`, `TestApplySobreFilasDanadas`, `TestPasoAWALDeUnaBaseDeFuera` con
  sus cotas y listas de bytes, y las filas retiradas de `TestErrores`; en `integracion_test.go`, lo que dice la
  línea de la tarea; `internal/app/grafo_test.go` y `internal/evals/preparar_test.go`, lo que dice la línea.
- Supuestos de T003 en gates/supuestos.md (tres líneas `[interno]`).

## Lo que falta: el diff de las tres rutas nuevas, comprobado en una copia

Aplicado sobre una copia desechable del árbol de este intento (`rsync -a` a `/tmp/kitlegal-t003-copia`), con
`go test -race` de `internal/graph`, con y sin `-tags=integration`, en verde, y `make ci` entero (resultado abajo):

```diff
--- a/internal/graph/integracion_enlace_test.go
+++ b/internal/graph/integracion_enlace_test.go
@@ import (
-	"io/fs"
 	"os"
@@
 // TestIntegracionEnlace fija world.db que es un enlace simbólico por la API
-// (FR-004, FR-033; research.md V47): a un destino de otro directorio con un -wal
+// (FR-004; research.md V47): a un destino de otro directorio con un -wal
 // huérfano junto a él, los tres verbos leen lo confirmado en el WAL, el destino
 // y su -wal quedan con los mismos bytes —su -shm existe al terminar: la
-// desviación declarada de §3— y junto al enlace no aparece nada; sin destino,
-// los verbos leen el grafo vacío y la entrega falla con «no se puede escribir»
-// sin crear el destino, ningún world.db-nuevo-* ni nada en el directorio.
+// desviación declarada de §3— y junto al enlace no aparece nada.
@@
 		assert.Equal(t, junto, estadoDelArbol(t, directorio), "junto al enlace no aparece ningún auxiliar")
 	})
-
-	t.Run("sin destino", func(t *testing.T) {
-		… (el subtest entero)
-	})
 }

--- a/internal/graph/lectura_test.go
+++ b/internal/graph/lectura_test.go
@@ type casoDeLectura struct {
-	// fallo es el mensaje esperado para la ruta de world.db; nil si no falla.
-	fallo func(ruta string) string
+	// fallo es el mensaje esperado para la ruta de world.db y la causa del
+	// fallo; nil si no falla.
+	fallo func(ruta string, causa error) string
@@ func TestLeerEstados(t *testing.T) {
-	noUtilizable := func(ruta string) string {
-		return "grafo: " + strconv.Quote(ruta) + " no es una base de datos utilizable; no se modifica"
+	noUtilizable := func(ruta string, causa error) string {
+		return "grafo: " + strconv.Quote(ruta) + " no es una base de datos utilizable: " + causa.Error()
 	}
-	posterior := func(ruta string) string {
+	posterior := func(ruta string, _ error) string {
@@ (fila «un directorio»)
-			fallo: func(ruta string) string {
+			fallo: func(ruta string, _ error) string {
@@
 // TestLeerSinPermisoDeLectura fija world.db que el proceso no puede leer
-// (FR-010; contracts/almacen-world-db.md §6, fila 2): «inesperado» con la
-// variante del acceso denegado, que es lo que quien lo lee puede arreglar, y
-// sin crear, cambiar ni retirar nada. Como el fichero no se deja leer, lo que se
-// compara es la lista del directorio y el tamaño, los permisos y la fecha de
-// cada entrada.
+// (FR-010; contracts/almacen-world-db.md §6): «inesperado» con la causa, el
+// acceso denegado, y sin crear, cambiar ni retirar nada. Como el fichero no se
+// deja leer, lo que se compara es la lista del directorio y el tamaño, los
+// permisos y la fecha de cada entrada.
@@
 	lectura, err := Leer(t.Context(), ConDirectorio(directorio))
 	assert.Nil(t, lectura)
-	compruebaFallo(t, err, schema.ClaseInesperado, ruta,
-		"grafo: "+strconv.Quote(ruta)+" no es una base de datos utilizable: acceso denegado; no se modifica")
+	require.ErrorIs(t, err, fs.ErrPermission)
+	compruebaFallo(t, err, schema.ClaseInesperado, ruta, mensajeInutilizable(t, ruta, err))
 	assert.Equal(t, antes, entradasDe(t, directorio))
 }
 
+// mensajeInutilizable es el mensaje de la regla genérica con la ruta de
+// world.db y la causa del fallo (contracts/almacen-world-db.md §6; H7.1
+// FR-070).
+func mensajeInutilizable(t *testing.T, ruta string, err error) string {
+	t.Helper()
+
+	var fallo *Error
+
+	require.ErrorAs(t, err, &fallo)
+
+	return "grafo: " + strconv.Quote(ruta) + " no es una base de datos utilizable: " + fallo.Causa.Error()
+}
+
@@ func compruebaEstado(t *testing.T, caso casoDeLectura) {
 	if caso.fallo != nil {
+		var fallo *Error
+
 		assert.Nil(t, lectura)
-		compruebaFallo(t, err, schema.ClaseInesperado, ruta, caso.fallo(ruta))
+		require.ErrorAs(t, err, &fallo)
+		compruebaFallo(t, err, schema.ClaseInesperado, ruta, caso.fallo(ruta, fallo.Causa))
 	} else {
@@ func TestLeerReabreInmutable(t *testing.T) { (subtest «con un -wal y sin su -shm: inutilizable»)
-		compruebaFallo(t, err, schema.ClaseInesperado, ruta,
-			"grafo: "+strconv.Quote(ruta)+" no es una base de datos utilizable; no se modifica")
+		compruebaFallo(t, err, schema.ClaseInesperado, ruta, mensajeInutilizable(t, ruta, err))
@@ func TestLecturaDeFilasDanadas(t *testing.T) {
-			compruebaFallo(t, caso.verbo(t.Context(), lectura), schema.ClaseInesperado, ruta,
-				"grafo: "+strconv.Quote(ruta)+" no es una base de datos utilizable; no se modifica")
+			err = caso.verbo(t.Context(), lectura)
+			compruebaFallo(t, err, schema.ClaseInesperado, ruta, mensajeInutilizable(t, ruta, err))

--- a/internal/graph/migraciones_test.go
+++ b/internal/graph/migraciones_test.go
@@ func TestMigrar(t *testing.T) { (subtest «una tabla schema_version ajena no es una base utilizable»)
-		assert.Equal(t, fmt.Sprintf("grafo: %q no es una base de datos utilizable; no se modifica", ruta), fallo.Error())
+		assert.Equal(t, fmt.Sprintf("grafo: %q no es una base de datos utilizable: %s", ruta, fallo.Causa), fallo.Error())
```

## Verificación

- Árbol del repositorio al terminar el intento 1, `make ci` en primer plano: código 2, en `lint`:

  ```
  internal/graph/integracion_enlace_test.go:61:3: undefined: compruebaNoSePuedeEscribir (typecheck)
  1 issues:
  * typecheck: 1
  make: *** [lint] Error 1
  ```

- Copia con el diff de arriba aplicado, `make -C /tmp/kitlegal-t003-copia ci`: `0 issues.` en `lint`, todos los
  paquetes `ok` en `test` y `test-integration`, `test-tiempos`, `vuln`, `schema-check`, `skills-check`,
  `goreleaser check`, `no leaks found`, `all modules verified` y `ci: todos los controles en verde`, código 0.

## Qué hace el intento 2

1. Aplicar ese diff a las tres rutas (nada más cambia: lo declarado ya está en el árbol).
2. `make ci` en primer plano y, en verde, marcar T003 `[X]` en el mismo turno.

# T003 · intento 2: en verde

Aplicado el diff de arriba a `internal/graph/lectura_test.go`, `internal/graph/migraciones_test.go` e
`internal/graph/integracion_enlace_test.go`, sin ningún otro cambio (lo declarado ya estaba en el árbol del intento 1).
Antes de aplicarlo se comprobó que la única llamada a `errorInutilizable` con causa nula (`abrir.go`, una versión
negativa del esquema) no la alcanza ningún test de lectura por `Leer`, y que `TestMigrar` la comprueba con `Contains`:
`mensajeInutilizable` y `caso.fallo(ruta, fallo.Causa)` nunca leen una causa nula.

`make ci` en primer plano sobre el árbol del repositorio: `0 issues.` en `lint`, todos los paquetes `ok` en `test` y
`test-integration`, `test-tiempos`, `vuln` («No vulnerabilities found»), `schema-check`, `skills-check`, `goreleaser
check`, `no leaks found`, `all modules verified`, `go mod tidy -diff` sin salida y `ci: todos los controles en verde`,
código 0. `internal/graph/publicar.go` y `publicar_test.go` no existen.
