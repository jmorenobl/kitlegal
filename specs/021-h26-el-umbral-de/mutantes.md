# Registro de los mutantes: H26

Lo escribe la tarea T003 (FR-078, SC-008; [data-model](./data-model.md) §6). Los dos mutantes son los de
[umbrales-del-informe](./contracts/umbrales-del-informe.md) §8, y el escenario, el 6 del [quickstart](./quickstart.md).
Es un registro del run: nada lo lee en `make ci`.

**Resultado: 4 de 4.** Cada uno de los dos mutantes deja en rojo `TestUmbralConjunto` (`boe-legislacion`) y
`TestUmbralesDeJurisprudencia` (`jurisprudencia`), con los casos que fija el contrato fallando por el veredicto.

| Mutante | `TestUmbralConjunto` | `TestUmbralesDeJurisprudencia` | `go test` |
|---|---|---|---|
| 1, un solo modo | rojo, 7 casos; por el veredicto, `3-y-0` y `2-y-1` | rojo, 10 casos; por el veredicto, `una-marcada-en-el-modo-orden` | código 1 |
| 2, sin decidir | rojo, 7 casos; por el veredicto, `4-y-0`, `2-y-2` y `104-con-3` | rojo, 10 casos; por el veredicto, `dos-marcadas-en-el-modo-orden` y `una-marcada-en-cada-modo` | código 1 |

## Dónde y sobre qué

- Cabeza: `8ae24050d733516ba366a5a9c7014cf6d40ee1f4` (`feat(H26): T002`). Go 1.27.2, `darwin/arm64`.
- Copia: un worktree de la cabeza en el directorio temporal, `copia="$TMPDIR/h26-mutantes"`, con
  `git worktree add --detach "$copia" HEAD`; y `f="$copia/internal/evals/umbrales.go"`.
- Sin mutar, la orden de test da en la copia `ok  	github.com/jmorenobl/kitlegal/internal/evals	2.582s`.
- La salida de `go test` y la de `git` se han leído con `rtk proxy` delante, sin resumir.
- El escenario se ejecutó dos veces, las dos en su orden y cada una con su worktree. En la primera, la orden de test
  iba seguida de otra y su código de salida no se recogió; la segunda, con la orden sola, lo recoge: 1 con cada
  mutante. Los casos en rojo son los mismos en las dos, y los que fallan por el veredicto, también; cambia el orden en
  que salen, porque los casos corren en paralelo. Las líneas copiadas aquí son de la segunda.

## Cómo se aplicó cada sustitución

`sed -i.orig`, que es como la escribe el quickstart, **no se ha ejecutado**: en la sesión de un paso el arnés pide para
esa orden una aprobación que nadie puede dar, y no se ha buscado otra forma de pasarla. La misma sustitución, de una
sola línea, se hizo con la herramienta de edición de la sesión sobre el `umbrales.go` de la copia, después de dejar el
original a su lado con `cp "$f" "$f.orig"`. La guarda, la orden de test y la restauración son las del quickstart, tal
cual.

Las dos expresiones de `sed` del quickstart casan con el código de la cabeza. Ejecutadas sin `-i` sobre el
`umbrales.go` del árbol, con la salida en un fichero del directorio temporal, cambian esa misma línea y ninguna otra, y
dan el mismo contenido que probaron los tests: `git diff --no-index` enseña el mismo blob de llegada, `925e77c` con la
primera y `fe957b7` con la segunda, desde `d52a8aa`. Quien repita el escenario en una sesión que admita `sed -i` no
tiene que ajustar nada.

## Mutante 1: el conjunto suma solo el primer modo del plan

La línea 349 de `internal/evals/umbrales.go`, dentro de `umbralesConjuntos`. Lo que enseña
`git diff --no-index -U0 "$f.orig" "$f"`:

```diff
@@ -349 +349 @@ func (j *juicioDelJuez) umbralesConjuntos() []Umbral {
-		for _, grupo := range j.deLosModos() {
+		for _, grupo := range j.deLosModos()[:1] {
```

Las órdenes:

```bash
cp "$f" "$f.orig"
# con la herramienta de edición, en la línea 349: `range j.deLosModos() {` → `range j.deLosModos()[:1] {`
git diff --no-index --quiet "$f.orig" "$f" && echo "EL MUTANTE 1 NO SE HA APLICADO"
go -C "$copia" test -count=1 -run '^(TestUmbralConjunto|TestUmbralesDeJurisprudencia)$' ./internal/evals/
git -C "$copia" checkout -- internal/evals/umbrales.go
```

La guarda no escribe nada: el mutante está aplicado. `go test` termina con código 1 y estas líneas:

```text
--- FAIL: TestUmbralConjunto (0.00s)
    --- FAIL: TestUmbralConjunto/0-y-3 (1.19s)
    --- FAIL: TestUmbralConjunto/2-y-1 (1.21s)
    --- FAIL: TestUmbralConjunto/104-con-3 (1.21s)
    --- FAIL: TestUmbralConjunto/104-con-2 (1.23s)
    --- FAIL: TestUmbralConjunto/3-y-0 (1.24s)
    --- FAIL: TestUmbralConjunto/4-y-0 (1.27s)
    --- FAIL: TestUmbralConjunto/2-y-2 (1.29s)
--- FAIL: TestUmbralesDeJurisprudencia (0.00s)
    --- FAIL: TestUmbralesDeJurisprudencia/ninguna-cuenta (0.62s)
    --- FAIL: TestUmbralesDeJurisprudencia/901-s-del-juez-en-el-modo-orden (0.62s)
    --- FAIL: TestUmbralesDeJurisprudencia/una-sin-activar (0.78s)
    --- FAIL: TestUmbralesDeJurisprudencia/dos-sin-documento-en-el-modo-herramienta (0.78s)
    --- FAIL: TestUmbralesDeJurisprudencia/una-sin-documento-en-el-modo-orden (0.79s)
    --- FAIL: TestUmbralesDeJurisprudencia/una-marcada-en-las-dos-clases (0.77s)
    --- FAIL: TestUmbralesDeJurisprudencia/una-marcada-en-cada-modo (0.77s)
    --- FAIL: TestUmbralesDeJurisprudencia/tres-con-si-en-afirma-que-existe (0.78s)
    --- FAIL: TestUmbralesDeJurisprudencia/una-marcada-en-el-modo-orden (0.73s)
    --- FAIL: TestUmbralesDeJurisprudencia/dos-marcadas-en-el-modo-orden (0.75s)
FAIL
FAIL	github.com/jmorenobl/kitlegal/internal/evals	2.395s
FAIL
```

Tres fallan por el veredicto, con `expected: "aprobado"` y `actual  : "fallo"`: son los que el contrato nombra. En los
tres, el conjunto pasa además de `Cumple: (bool) true` a `false`, y el informe lleva un motivo donde el test espera
«Motivos: ninguno».

| Caso | Lo que el test espera del conjunto | Lo que da el mutante | Motivo que escribe el mutante |
|---|---|---|---|
| `TestUmbralConjunto/3-y-0` | 3 de 108, cumple | 3 de 54, no cumple | «3 de 54 (5,6 %), y tiene que ser ≤ 2,8 %» |
| `TestUmbralConjunto/2-y-1` | 3 de 108, cumple | 2 de 54, no cumple | «2 de 54 (3,7 %), y tiene que ser ≤ 2,8 %» |
| `TestUmbralesDeJurisprudencia/una-marcada-en-el-modo-orden` | 1 de 60, cumple | 1 de 30, no cumple | «1 de 30 (3,3 %), y tiene que ser ≤ 1,7 %» |

Los otros catorce fallan por la medida y el total del conjunto, con el veredicto que el test espera. En los cinco de
ellos que esperan `fallo` —`4-y-0`, `2-y-2`, `104-con-3`, `dos-marcadas-en-el-modo-orden` y `una-marcada-en-cada-modo`—
falta además la línea del motivo, que el mutante escribe con otras cifras:

| Caso | Esperado | Con el mutante |
|---|---|---|
| `TestUmbralConjunto/0-y-3` | 3 de 108 | 0 de 54 |
| `TestUmbralConjunto/4-y-0` | total 108 | total 54 |
| `TestUmbralConjunto/2-y-2` | 4 de 108 | 2 de 54 |
| `TestUmbralConjunto/104-con-2` | 2 de 104 | 1 de 52 |
| `TestUmbralConjunto/104-con-3` | 3 de 104 | 2 de 52 |
| `TestUmbralesDeJurisprudencia/una-marcada-en-las-dos-clases` | 1 de 60 | 0 de 30 |
| `TestUmbralesDeJurisprudencia/una-marcada-en-cada-modo` | 2 de 60 | 1 de 30 |
| `TestUmbralesDeJurisprudencia/dos-marcadas-en-el-modo-orden` | total 60 | total 30 |
| `ninguna-cuenta`, `901-s-del-juez-en-el-modo-orden`, `una-sin-activar`, `una-sin-documento-en-el-modo-orden`, `dos-sin-documento-en-el-modo-herramienta` y `tres-con-si-en-afirma-que-existe`, de `TestUmbralesDeJurisprudencia` | total 60 | total 30 |

Después de la restauración, `git diff --no-index --quiet "$f.orig" "$f"` sale con 0: el `umbrales.go` de la copia vuelve
a ser el de la cabeza antes del mutante 2.

## Mutante 2: el `decide` del conjunto, en falso

La línea 360 de `internal/evals/umbrales.go`, dentro de `umbralesConjuntos`. Lo que enseña
`git diff --no-index -U0 "$f.orig" "$f"`:

```diff
@@ -360 +360 @@ func (j *juicioDelJuez) umbralesConjuntos() []Umbral {
-			Decide:      true,
+			Decide:      false,
```

Las órdenes:

```bash
cp "$f" "$f.orig"
# con la herramienta de edición, en la línea 360: `Decide:      true,` → `Decide:      false,`
git diff --no-index --quiet "$f.orig" "$f" && echo "EL MUTANTE 2 NO SE HA APLICADO"
go -C "$copia" test -count=1 -run '^(TestUmbralConjunto|TestUmbralesDeJurisprudencia)$' ./internal/evals/
git -C "$copia" checkout -- internal/evals/umbrales.go
```

La guarda no escribe nada: el mutante está aplicado. `go test` termina con código 1 y estas líneas:

```text
--- FAIL: TestUmbralConjunto (0.00s)
    --- FAIL: TestUmbralConjunto/0-y-3 (1.28s)
    --- FAIL: TestUmbralConjunto/104-con-2 (1.29s)
    --- FAIL: TestUmbralConjunto/104-con-3 (1.30s)
    --- FAIL: TestUmbralConjunto/4-y-0 (1.31s)
    --- FAIL: TestUmbralConjunto/2-y-1 (1.33s)
    --- FAIL: TestUmbralConjunto/3-y-0 (1.35s)
    --- FAIL: TestUmbralConjunto/2-y-2 (1.36s)
--- FAIL: TestUmbralesDeJurisprudencia (0.00s)
    --- FAIL: TestUmbralesDeJurisprudencia/ninguna-cuenta (0.65s)
    --- FAIL: TestUmbralesDeJurisprudencia/una-sin-activar (0.67s)
    --- FAIL: TestUmbralesDeJurisprudencia/dos-sin-documento-en-el-modo-herramienta (0.68s)
    --- FAIL: TestUmbralesDeJurisprudencia/una-sin-documento-en-el-modo-orden (0.68s)
    --- FAIL: TestUmbralesDeJurisprudencia/901-s-del-juez-en-el-modo-orden (0.68s)
    --- FAIL: TestUmbralesDeJurisprudencia/una-marcada-en-las-dos-clases (0.68s)
    --- FAIL: TestUmbralesDeJurisprudencia/una-marcada-en-el-modo-orden (0.66s)
    --- FAIL: TestUmbralesDeJurisprudencia/dos-marcadas-en-el-modo-orden (0.68s)
    --- FAIL: TestUmbralesDeJurisprudencia/tres-con-si-en-afirma-que-existe (0.70s)
    --- FAIL: TestUmbralesDeJurisprudencia/una-marcada-en-cada-modo (0.70s)
FAIL
FAIL	github.com/jmorenobl/kitlegal/internal/evals	2.355s
FAIL
```

Cinco fallan por el veredicto, con `expected: "fallo"` y `actual  : "aprobado"`: son los que el contrato nombra. El
informe no lleva la línea del motivo que el test espera, porque un umbral que no decide no da ninguno.

| Caso | Motivo que el test espera y no está |
|---|---|
| `TestUmbralConjunto/4-y-0` | «4 de 108 (3,7 %), y tiene que ser ≤ 2,8 %» |
| `TestUmbralConjunto/2-y-2` | «4 de 108 (3,7 %), y tiene que ser ≤ 2,8 %» |
| `TestUmbralConjunto/104-con-3` | «3 de 104 (2,9 %), y tiene que ser ≤ 2,8 %» |
| `TestUmbralesDeJurisprudencia/dos-marcadas-en-el-modo-orden` | «2 de 60 (3,3 %), y tiene que ser ≤ 1,7 %» |
| `TestUmbralesDeJurisprudencia/una-marcada-en-cada-modo` | «2 de 60 (3,3 %), y tiene que ser ≤ 1,7 %» |

Los diecisiete fallan además por el elemento del conjunto, que pasa de `Decide: (bool) true` a `false`: los siete de
`TestUmbralConjunto`, con el mensaje «nueve de los trece deciden», y los diez de `TestUmbralesDeJurisprudencia`, con
«deciden todos menos los recuentos de las dos clases del juez de cada modo». Los doce que no fallan por el veredicto
—`3-y-0`, `0-y-3`, `2-y-1` y `104-con-2`, y los ocho restantes de `jurisprudencia`— fallan solo por eso.

Después de la restauración, `git diff --no-index --quiet "$f.orig" "$f"` sale con 0.

## Al terminar

```bash
git worktree remove --force "$copia"
git status --short
```

`$copia` ya no existe, `git worktree list` no la nombra, y `git status --short` da lo mismo que antes de empezar, todo
dentro del directorio del feature:

```text
 M specs/021-h26-el-umbral-de/gates/tarea-actual.json
 M specs/021-h26-el-umbral-de/gates/tareas-intentos.json
```

Ningún mutante ha estado en el árbol: `git diff --stat HEAD -- internal/ evals/` no escribe nada.
