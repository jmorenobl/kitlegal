# T037 · intento 1 de 3: `ci` en rojo en el ejecutor linux por un test que solo mide la plataforma local

> **Intento 2 de 3 (2026-09-13): en verde.** T038 quedó commiteada en `d45f0ea`; el intento 2 la publicó en avance
> rápido, encontró la #24 y no creó otra, y `ci` (run `34782609287`) y los cuatro estados de Codecov salieron en verde
> con medida real (`project` 94,73 %, `internal/core` 90,36 %, `internal/cli` 98,09 %, `patch` 97,49 % del diff sobre
> un objetivo de 92,95 %). La lectura completa está en [`evidencia-plataforma.md`](./evidencia-plataforma.md). Lo que
> sigue es el análisis del intento 1, sin cambios.

**Estado tras el intento 1**: T037 **sin marcar**. Rama publicada en `40efb06` y propuesta de cambio
[#24](https://github.com/jmorenobl/kitlegal/pull/24) abierta, pero `ci` en rojo (run `34781187264`) y ningún estado de
Codecov emitido. Base del intento: `40efb06` (`feat(H4): T036`). La evidencia completa, con las salidas literales, está en
[`evidencia-plataforma.md`](./evidencia-plataforma.md).

## Qué pasó

- `git push -u origin h4-applet-boe-puerto`: `guardia-push` en verde y `* [new branch]`; `origin` tiene exactamente
  `40efb06`.
- `gh pr view` no encontró propuesta; `gh pr create … --body-file …/gates/pr-h4.md` abrió la #24.
- `gh pr checks --watch` terminó con `ci	fail	2m13s` (salida 1). El paso «Ejecutar los controles» (`make ci`) se detuvo
  en `make test` (`Makefile:67`): `formato` y `lint` en verde (`0 issues.`), todos los paquetes en `ok` salvo
  `internal`, con `--- FAIL: TestDependenciasDelBinario`, «extra elements in list A: `github.com/mattn/go-isatty`,
  `github.com/ncruces/go-strftime`» y una lista medida de 16 módulos frente a los 18 declarados.
- **Codecov no emitió ningún estado, y la causa no es la configuración.** «Publicar el perfil de cobertura» salió
  `skipped`: un paso de GitHub Actions no se ejecuta si ha fallado uno anterior, y sin perfil subido Codecov no tiene
  nada que evaluar. No es un verde vacío ni falta nada en `codecov.yml` o `ci.yml`: los estados llegan con la primera
  ejecución en la que `make ci` pase.

## Causa

`TestDependenciasDelBinario` (T026) ejecuta `go list -deps` sobre `./cmd/kitlegal` con el entorno del proceso, así que
mide **solo la plataforma del ordenador que ejecuta el test**. Y los módulos que enlaza el binario dependen de la
plataforma, porque `modernc.org/libc` elige sus ficheros por sistema:

| Plataforma (con `CGO_ENABLED=0`) | Módulos de terceros | Faltan respecto a la lista declarada |
|---|---|---|
| darwin/amd64, darwin/arm64 | 18 | ninguno |
| linux/amd64, linux/arm64 | 16 | `github.com/mattn/go-isatty`, `github.com/ncruces/go-strftime` |
| windows/amd64, windows/arm64 | 17 | `github.com/google/uuid` |

En `modernc.org/libc@v1.75.6`, `go-isatty` lo importa `libc.go` (`//go:build !linux || mips64le`); `go-strftime`,
`libc_unix.go` (`unix && !(linux && (amd64 || arm64 || …))`) y `libc_windows.go`; `uuid`, los ficheros de darwin, de
linux (`libc_musl.go`) y del resto de unix, pero ninguno de windows.

- **La lista declarada es correcta**: `modulosDelBinario` ya es la unión de las seis plataformas de distribución, y
  `go.mod` y `go.sum` no cambian. El defecto está en el test: su veredicto depende de dónde corre.
- **Por qué no salió antes.** Todos los `make ci` del hito, incluido el escenario 16 de T036, se ejecutaron en
  darwin/arm64, donde el binario enlaza los 18. La integración continua es la primera ejecución en otra plataforma.
- **Reproducido en local** con un binario de test nativo cuyo `go list` hijo hereda la plataforma de linux:
  `go test -count=1 -exec 'env GOOS=linux GOARCH=amd64 CGO_ENABLED=0' -run '^TestDependenciasDelBinario$' ./internal/`
  sobre un clon de `40efb06` falla con la misma lista A, la misma lista B y el mismo mensaje que el run.

## Por qué no se arregla en este intento, y cómo se arregla

- Las rutas congeladas de T037 son `gates/pr-h4.md` y `gates/evidencia-plataforma.md`. `internal/arch_test.go` queda
  fuera y el guardián de diff lo rechazaría.
- **Redelimitar T037 tampoco serviría.** El commit de cada tarea lo hace el workflow **después** del intento
  (`commit_tarea`), y el push de T037 ocurre **durante** el intento. Un intento que arreglase el test empujaría `HEAD`
  sin el arreglo, porque el ejecutor desatendido no puede commitear. Además, `[plataforma]` no mezcla otro trabajo.
- **Arreglo: una tarea nueva, T038, sin la etiqueta de plataforma, colocada en `tasks.md` justo antes de T037.**
  `siguiente_tarea` elige la primera línea sin marcar, así que T038 va ahora, con sus tres intentos: test primero,
  `make ci` en verde y commit del workflow. Después, T037 (intento 2) publica ese commit y lee `ci` y Codecov sobre el
  árbol arreglado. Comprobado con la extracción de rutas del workflow sobre las dos líneas: T038 declara solo
  `internal/arch_test.go` y no lleva ninguna etiqueta; T037 conserva sus dos rutas.
- **No se rebaja nada**: la lista de módulos no cambia, no se retira ninguna comprobación, no se salta ningún test y
  no se añade ninguna supresión.

## Arreglo comprobado sobre un clon desechable

Clon de `40efb06` en `/tmp/kitlegal-t037/copia`, fuera del repositorio. Diseño:

- `plataformasDeDistribucion`: darwin, linux y windows sobre amd64 y arm64, las del release sin cgo (ADR 0002;
  `docs/ROADMAP.md`, entrega).
- `TestDependenciasDelBinario` mide el cierre de cada una con `GOOS`, `GOARCH` y `CGO_ENABLED=0` en el entorno de
  `go list`, exige que cada una enlace algo y compara la unión con `modulosDelBinario`. Falla un módulo sin declarar en
  cualquier plataforma y uno declarado que no enlace ninguna, y el mensaje da las plataformas de cada módulo.
- `ejecutaGoCon` añade el entorno; `ejecutaGo` pasa a ser su caso sin entorno. La línea `//nolint:gosec` que ya existía
  no se mueve ni cambia: `git diff -U0 | grep nolint` no da ninguna línea.

Resultados:

| Comprobación | Resultado |
|---|---|
| Rojo antes del arreglo, con `-exec 'env GOOS=linux GOARCH=amd64 CGO_ENABLED=0'` | `FAIL`: los mismos dos módulos que el run `34781187264` |
| `TestArquitectura`, `TestElBinarioNoEnlazaLosEjemplos`, `TestDependenciasDelBinario`, `TestLasFuentesNoFirmanComoKitlegal`, nativo | cuatro `--- PASS` |
| `TestDependenciasDelBinario` con `-exec` de linux/amd64 y de windows/arm64 | `--- PASS` en los dos |
| `golangci-lint run ./internal/` | primero dos hallazgos de `misspell` en comentarios nuevos (`distribuye`, `prevalecen`), reescritos («se entrega», «ganan»); después, `0 issues.` |
| Sonda: quitar `"github.com/mattn/go-isatty"` de la lista | `FAIL`, «extra elements in list B: github.com/mattn/go-isatty», y en el mensaje `github.com/mattn/go-isatty:[darwin/amd64 darwin/arm64 windows/amd64 windows/arm64]` |
| Sonda: declarar `"example.com/ficticio"` | `FAIL`, «extra elements in list A: example.com/ficticio» |
| `make -C /tmp/kitlegal-t037/copia ci` | código 0, `ci: todos los controles en verde` |

Cada sonda se hizo en su propia copia (`rsync -a --exclude .git` del clon arreglado).

El diff comprobado (`git -C /tmp/kitlegal-t037/copia diff`), para que T038 no dependa de que el clon siga existiendo:

```diff
--- a/internal/arch_test.go
+++ b/internal/arch_test.go
@@ -117,7 +117,8 @@ func TestArquitectura(t *testing.T) {
 // gates/pr-h1.md; los que entran en H4, cuando el applet boe enlaza
 // internal/httpx e internal/cache, en gates/pr-h4.md (FR-060, FR-124;
 // research.md D14 de H4). Lo que importa cada uno lo mide `go list -deps` sobre
-// el binario.
+// el binario de cada una de plataformasDeDistribucion; el que no llega a todas
+// lo dice en su línea.
 //
 // Es una lista escrita a mano a propósito. Cuando un hito, o una actualización
 // de módulos, enlace uno nuevo, este test falla y obliga a hacer lo que la
@@ -133,13 +134,13 @@ var modulosDelBinario = []string{
 	// H4: lo importa modernc.org/libc, el entorno de C traducido a Go sobre el
 	// que corre el controlador de SQLite de internal/cache.
 	"github.com/dustin/go-humanize",
-	// H4: lo importa modernc.org/libc.
+	// H4: lo importa modernc.org/libc en darwin y linux; no llega a windows.
 	"github.com/google/uuid",
 	// §V, H1: el esquema de entrada y salida de --describe, en internal/cli.
 	"github.com/invopop/jsonschema",
-	// H4: lo importa modernc.org/libc.
+	// H4: lo importa modernc.org/libc en darwin y windows; no llega a linux.
 	"github.com/mattn/go-isatty",
-	// H4: lo importa modernc.org/libc.
+	// H4: lo importa modernc.org/libc en darwin y windows; no llega a linux.
 	"github.com/ncruces/go-strftime",
 	// H1: lo importa github.com/invopop/jsonschema para las propiedades en orden.
 	"github.com/pb33f/ordered-map/v2",
@@ -187,28 +188,58 @@ func TestElBinarioNoEnlazaLosEjemplos(t *testing.T) {
 	}
 }
 
-// TestDependenciasDelBinario comprueba que el binario distribuido no enlaza
-// ningún módulo de terceros fuera de los declarados (FR-060, constitución §V).
-// Mira el cierre transitivo real de `go list -deps` sobre el punto de entrada,
-// que es lo mismo que acaba en `go version -m` del ejecutable: ni los módulos
-// que solo usan los tests ni los de las herramientas cuentan aquí.
+// plataformasDeDistribucion son las plataformas para las que se entrega el
+// binario, sin cgo: darwin, linux y windows sobre amd64 y arm64 (ADR 0002;
+// docs/ROADMAP.md, entrega). Los módulos que enlaza dependen de la plataforma,
+// porque modernc.org/libc elige sus ficheros por sistema: en linux no llegan
+// github.com/mattn/go-isatty ni github.com/ncruces/go-strftime, y en windows no
+// llega github.com/google/uuid. Por eso se mide en todas y no solo en la del
+// ordenador que ejecuta el test, que haría depender el veredicto de dónde corre.
+var plataformasDeDistribucion = []struct{ sistema, arquitectura string }{
+	{"darwin", "amd64"},
+	{"darwin", "arm64"},
+	{"linux", "amd64"},
+	{"linux", "arm64"},
+	{"windows", "amd64"},
+	{"windows", "arm64"},
+}
+
+// TestDependenciasDelBinario comprueba que el binario distribuido no enlaza, en
+// ninguna de sus plataformas, ningún módulo de terceros fuera de los declarados
+// (FR-060, constitución §V), y que no se declara ninguno que no enlace en
+// ninguna. Mira el cierre transitivo real de `go list -deps` sobre el punto de
+// entrada con GOOS, GOARCH y CGO_ENABLED=0 de cada plataforma, que es lo mismo
+// que acaba en `go version -m` de su ejecutable: ni los módulos que solo usan
+// los tests ni los de las herramientas cuentan aquí.
 func TestDependenciasDelBinario(t *testing.T) {
 	t.Parallel()
 
 	modulo := rutaDelModulo(t)
-	enlazados := map[string]bool{}
+	// Cada módulo con las plataformas cuyo binario lo enlaza.
+	enlazados := map[string][]string{}
+
+	for _, plataforma := range plataformasDeDistribucion {
+		nombre := plataforma.sistema + "/" + plataforma.arquitectura
+		entorno := []string{"GOOS=" + plataforma.sistema, "GOARCH=" + plataforma.arquitectura, "CGO_ENABLED=0"}
+		deEsta := map[string]bool{}
 
-	for linea := range strings.SplitSeq(ejecutaGo(t, "list", "-deps", "-f", plantillaDeModulos, "./cmd/kitlegal"), "\n") {
-		if linea = strings.TrimSpace(linea); linea != "" && linea != modulo {
-			enlazados[linea] = true
+		for linea := range strings.SplitSeq(ejecutaGoCon(t, entorno, "list", "-deps", "-f", plantillaDeModulos, "./cmd/kitlegal"), "\n") {
+			if linea = strings.TrimSpace(linea); linea != "" && linea != modulo {
+				deEsta[linea] = true
+			}
 		}
-	}
 
-	require.NotEmpty(t, enlazados, "el binario enlaza al menos el analizador de la línea de órdenes")
+		require.NotEmpty(t, deEsta, "en %s el binario enlaza al menos el analizador de la línea de órdenes", nombre)
+
+		for enlazado := range deEsta {
+			enlazados[enlazado] = append(enlazados[enlazado], nombre)
+		}
+	}
 
 	assert.ElementsMatch(t, modulosDelBinario, slices.Sorted(maps.Keys(enlazados)),
-		"el binario distribuido enlaza un módulo que no está declarado y justificado "+
-			"(FR-060, constitución §V): justifícalo en plan.md y en la propuesta de cambio antes de añadirlo")
+		"el binario distribuido enlaza en alguna plataforma un módulo que no está declarado y justificado, o se "+
+			"declara uno que no enlaza en ninguna (FR-060, constitución §V): justifícalo en plan.md y en la propuesta "+
+			"de cambio antes de añadirlo; plataformas que enlazan cada módulo: %v", enlazados)
 }
 
 // prefijosReservados son los del espacio de nombres con el que firma lo que no
@@ -619,19 +650,30 @@ func rutaDelModulo(t *testing.T) string {
 	return modulo
 }
 
-// ejecutaGo ejecuta el go command en la raíz del módulo y devuelve su salida
-// estándar. No toca la red: `go list` solo consulta el módulo y la caché.
+// ejecutaGo ejecuta el go command en la raíz del módulo, con el entorno del
+// proceso, y devuelve su salida estándar. No toca la red: `go list` solo
+// consulta el módulo y la caché.
 func ejecutaGo(t *testing.T, argumentos ...string) string {
 	t.Helper()
 
-	// El ejecutable es constante y los argumentos no vienen de fuera: son
-	// literales de este fichero más las rutas de paquete que sale de enumerar
-	// internal/app/testdata en el propio árbol. No hay entrada de usuario, red ni
-	// variable de entorno en la orden, de modo que G204 no tiene aquí nada que
-	// prevenir.
+	return ejecutaGoCon(t, nil, argumentos...)
+}
+
+// ejecutaGoCon es ejecutaGo con variables de entorno añadidas a las del
+// proceso, que ganan a las heredadas: con GOOS y GOARCH, `go list`
+// describe el cierre de otra plataforma sin compilar nada para ella.
+func ejecutaGoCon(t *testing.T, entorno []string, argumentos ...string) string {
+	t.Helper()
+
+	// El ejecutable es constante y ni los argumentos ni las variables añadidas
+	// vienen de fuera: son literales de este fichero, las rutas de paquete que
+	// sale de enumerar internal/app/testdata en el propio árbol y las plataformas
+	// de plataformasDeDistribucion. No hay entrada de usuario ni red en la orden,
+	// de modo que G204 no tiene aquí nada que prevenir.
 	//nolint:gosec // los argumentos son literales de este fichero y rutas del propio árbol; no hay entrada externa.
 	orden := exec.CommandContext(t.Context(), "go", argumentos...)
 	orden.Dir = raizDelModulo
+	orden.Env = append(os.Environ(), entorno...)
 
 	salida, err := orden.Output()
 	if err != nil {
```

## Qué cambia en este intento, todo dentro del directorio del feature

- `tasks.md`: la línea de T038, justo antes de T037, y sus referencias en las tablas (rebanadas completas, Definition of
  Done 1, FR-124, dependencias y orden, ficheros de H0-H3 tocados). La línea de T037 no cambia.
- `gates/pr-h4.md`, que es el cuerpo de la #24: la sección «Dependencias» explica que los módulos enlazados dependen de
  la plataforma, la tabla dice a qué plataformas llegan `uuid`, `go-isatty` y `go-strftime`, y «Pendientes» recoge el
  rojo, T038 y S9. La propuesta se actualizó con ese fichero (`gh pr edit --body-file`).
- `gates/evidencia-plataforma.md` (nueva) y esta nota.

## Qué hace el intento 2, cuando T038 esté commiteada

1. `git push -u origin h4-applet-boe-puerto`: avance rápido sobre `40efb06`, sin forzar.
2. `gh pr view` encuentra la #24, así que no se crea otra.
3. `gh pr checks --watch` y `gh run list`. Además, leer los títulos de los check-runs de Codecov
   (`gh api repos/jmorenobl/kitlegal/commits/<sha>/check-runs`): `codecov/project` ≥ 70 %,
   `codecov/project/internal/core` ≥ 85 %, `codecov/project/internal/cli` ≥ 90 % y `codecov/patch` con objetivo
   `auto`. Hay que distinguir un verde que mide de uno vacío.
4. Completar `evidencia-plataforma.md` y «Pendientes» de `pr-h4.md` con el resultado, actualizar el cuerpo de la
   propuesta, ejecutar `make ci` y marcar solo si todo está en verde.
