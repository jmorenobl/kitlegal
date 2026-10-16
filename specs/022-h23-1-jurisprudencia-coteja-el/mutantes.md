# Registro de los mutantes: H23.1

Lo escribe la tarea T002 (FR-064, SC-006; research D12, M8). Los dos mutantes son los de
[umbral-documento-sin-cotejar](./contracts/umbral-documento-sin-cotejar.md) §6, y el escenario, el 4 del
[quickstart](./quickstart.md). Es un registro del run: nada lo lee en `make ci`.

**Resultado: 2 de 2.** Cada mutante deja en rojo `TestUmbralesDeJurisprudencia`, con los casos que fija el contrato.
La revisión final añadió un tercero, que T002 no tenía, y los dos casos que lo matan (abajo, «La revisión final»): con
ellos, **3 de 3**.

| Mutante | Casos en rojo | Por el veredicto (`fallo` → `aprobado`) | `go test` |
|---|---|---|---|
| 1, «ninguna sesión con cotejo» | 3: `una-sin-activar-y-sin-cotejar`, `una-sin-cotejar-en-el-modo-orden`, `una-sin-cotejar-en-el-modo-herramienta` | `una-sin-cotejar-en-el-modo-orden` y `una-sin-cotejar-en-el-modo-herramienta` | código 1 |
| 2, `Decide` en falso | 14, los catorce de la tabla | `una-sin-cotejar-en-el-modo-orden` y `una-sin-cotejar-en-el-modo-herramienta` | código 1 |

El test tiene un subtest más, `el-elemento-del-contrato`, que no está en la tabla de casos
(`casosDeLosUmbralesDeJurisprudencia`, en `internal/evals/informe_test.go`) y pasa con los dos mutantes.

## Dónde y sobre qué

- Cabeza: `33876081c05f1cb418a98ebf088930f682827916` (`feat(H23.1): T001`). Go 1.27.2, `darwin/arm64`.
- Copia: un worktree de la cabeza en el directorio temporal, `copia="$TMPDIR/h231-mutantes"`, con
  `git worktree add --detach "$copia" HEAD`.
- Sin mutar, la orden de test se ejecutó en el repositorio, no en la copia, con `internal/` igual que la cabeza
  (`git diff --quiet HEAD -- internal/`): los quince subtests en `PASS` y
  `ok  	github.com/jmorenobl/kitlegal/internal/evals	1.541s`.
- Las sustituciones son las `sed -i.orig` del quickstart, tal cual, y las dos guardas (`git diff --no-index --quiet`)
  callaron: ninguna «NO SE HA APLICADO». Cada sustitución cambia una sola línea, comprobado con `diff` entre el
  `.orig` y el fichero.
- Todas las órdenes corrieron con `rtk proxy sh -c '…'`, en primer plano y en el orden del quickstart. La salida de
  `go test` se guardó en un fichero del directorio temporal (`$TMPDIR/t002-m1.log`, `$TMPDIR/t002-m2.log`) para leer
  los mensajes de cada aserción, y se filtró con el `grep` del quickstart; el código de salida es el de `go test`
  solo. Las líneas copiadas abajo son de esos ficheros, sin resumir; los casos corren en paralelo y salen en el orden
  en que terminan.
- Tras cada mutante, `git -C "$copia" checkout -- <fichero>` restauró el original: `git -C "$copia" status --short`
  solo enseñaba los `.orig`, que se fueron con la copia.

## Mutante 1: «ninguna sesión con cotejo» en lugar de «cada sesión»

`internal/evals/informe.go`, línea 707, en `sinCotejar`:

```diff
< 		if !respuesta.cotejo {
---
> 		if !slices.ContainsFunc(g.queLePidenElCotejo(), func(otra respuestaDelJuez) bool { return otra.cotejo }) {
```

Órdenes:

```bash
f="$copia/internal/evals/informe.go"
sed -i.orig '/^func (g grupoJuzgado) sinCotejar()/,/^}/ s/if !respuesta\.cotejo {/if !slices.ContainsFunc(g.queLePidenElCotejo(), func(otra respuestaDelJuez) bool { return otra.cotejo }) {/' "$f"
git diff --no-index --quiet "$f.orig" "$f" && echo "EL MUTANTE 1 NO SE HA APLICADO"
go -C "$copia" test -count=1 -v -run '^TestUmbralesDeJurisprudencia$' ./internal/evals/ > "$TMPDIR/t002-m1.log" 2>&1; echo "go-test-exit=$?"
grep -e "--- FAIL" -e "^ok" -e "^FAIL" "$TMPDIR/t002-m1.log"
git -C "$copia" checkout -- internal/evals/informe.go
```

Salida:

```text
go-test-exit=1
--- FAIL: TestUmbralesDeJurisprudencia (0.00s)
    --- FAIL: TestUmbralesDeJurisprudencia/una-sin-activar-y-sin-cotejar (0.66s)
    --- FAIL: TestUmbralesDeJurisprudencia/una-sin-cotejar-en-el-modo-orden (0.70s)
    --- FAIL: TestUmbralesDeJurisprudencia/una-sin-cotejar-en-el-modo-herramienta (0.50s)
FAIL
FAIL	github.com/jmorenobl/kitlegal/internal/evals	1.664s
FAIL
```

Los tres casos que pide el contrato, y ninguno más. En los tres, el elemento `documento_sin_cotejar` de su modo
espera 1 de 12 y recibe 0 (`exigirUmbrales`, `umbrales_test.go:1404`; espacios comprimidos):

```text
 - Medida: (float64) 1,
 + Medida: (float64) 0,
 Total: (*int)(12),
 - Cumple: (bool) false,
 + Cumple: (bool) true,
```

Y el motivo de la raíz falta (`exigirMotivosDeLaRaiz`, `informe_test.go:3865`):

```text
una-sin-activar-y-sin-cotejar:          falta la línea "- umbral documento_sin_cotejar:claude-sonnet-5-5:orden: 1 de 12 (8,3 %), y tiene que ser ≤ 0,0 %: 06-documento-que-no-es-el-pedido-claude-sonnet-5-5-02"
una-sin-cotejar-en-el-modo-orden:       falta la línea "- umbral documento_sin_cotejar:claude-sonnet-5-5:orden: 1 de 12 (8,3 %), y tiene que ser ≤ 0,0 %: 06-documento-que-no-es-el-pedido-claude-sonnet-5-5-01"
una-sin-cotejar-en-el-modo-herramienta: falta la línea "- umbral documento_sin_cotejar:claude-sonnet-5-5:herramienta: 1 de 12 (8,3 %), y tiene que ser ≤ 0,0 %: 06-documento-que-no-es-el-pedido-herramienta-claude-sonnet-5-5-01"
```

En `una-sin-cotejar-en-el-modo-orden` y `una-sin-cotejar-en-el-modo-herramienta` cae también el veredicto
(`informe_test.go:5537`): `expected: "fallo"`, `actual  : "aprobado"`. En `una-sin-activar-y-sin-cotejar` el veredicto
sigue en `fallo` por `sin_activar`, y el caso cae por el elemento y por el motivo.

## Mutante 2: `Decide` en falso

`internal/evals/umbrales.go`, línea 481, en `umbralDeDocumentoSinCotejar`:

```diff
< 		Decide:      true,
---
> 		Decide:      false,
```

Órdenes:

```bash
f="$copia/internal/evals/umbrales.go"
sed -i.orig '/^func umbralDeDocumentoSinCotejar(/,/^}/ s/Decide:\( *\)true,/Decide:\1false,/' "$f"
git diff --no-index --quiet "$f.orig" "$f" && echo "EL MUTANTE 2 NO SE HA APLICADO"
go -C "$copia" test -count=1 -v -run '^TestUmbralesDeJurisprudencia$' ./internal/evals/ > "$TMPDIR/t002-m2.log" 2>&1; echo "go-test-exit=$?"
grep -e "--- FAIL" -e "^ok" -e "^FAIL" "$TMPDIR/t002-m2.log"
git -C "$copia" checkout -- internal/evals/umbrales.go
```

Salida:

```text
go-test-exit=1
--- FAIL: TestUmbralesDeJurisprudencia (0.00s)
    --- FAIL: TestUmbralesDeJurisprudencia/dos-sin-documento-en-el-modo-herramienta (0.69s)
    --- FAIL: TestUmbralesDeJurisprudencia/901-s-del-juez-en-el-modo-orden (0.71s)
    --- FAIL: TestUmbralesDeJurisprudencia/una-sin-activar-y-sin-cotejar (0.72s)
    --- FAIL: TestUmbralesDeJurisprudencia/una-sin-documento-en-el-modo-orden (0.73s)
    --- FAIL: TestUmbralesDeJurisprudencia/una-sin-medir-y-sin-cotejar (0.73s)
    --- FAIL: TestUmbralesDeJurisprudencia/tres-con-si-en-afirma-que-existe (0.74s)
    --- FAIL: TestUmbralesDeJurisprudencia/ninguna-cuenta (0.74s)
    --- FAIL: TestUmbralesDeJurisprudencia/una-sin-cotejar-en-el-modo-orden (0.74s)
    --- FAIL: TestUmbralesDeJurisprudencia/una-sin-cotejar-en-el-modo-herramienta (0.53s)
    --- FAIL: TestUmbralesDeJurisprudencia/una-marcada-en-las-dos-clases (0.51s)
    --- FAIL: TestUmbralesDeJurisprudencia/una-marcada-en-cada-modo (0.52s)
    --- FAIL: TestUmbralesDeJurisprudencia/dos-marcadas-en-el-modo-orden (0.53s)
    --- FAIL: TestUmbralesDeJurisprudencia/una-marcada-en-el-modo-orden (0.52s)
    --- FAIL: TestUmbralesDeJurisprudencia/una-sin-activar (0.55s)
FAIL
FAIL	github.com/jmorenobl/kitlegal/internal/evals	1.748s
FAIL
```

Los catorce son los catorce `nombre:` de `casosDeLosUmbralesDeJurisprudencia`. Lo que ve cada uno, contado en la
salida:

- Los dos elementos, `documento_sin_cotejar` de `orden` y de `herramienta`, ya no son los esperados: 28 pares
  `- Decide: (bool) true` / `+ Decide: (bool) false`, dos por caso.
- Deciden nueve y no once: 14 veces `should have 11 item(s), but has 9`, con el mensaje «deciden todos menos los
  recuentos de las dos clases del juez de cada modo» (`informe_test.go:5532`).
- `una-sin-cotejar-en-el-modo-orden` y `una-sin-cotejar-en-el-modo-herramienta` dan `aprobado` donde se espera
  `fallo` (`informe_test.go:5537`), y les falta el motivo de la raíz `umbral documento_sin_cotejar:…`; a
  `una-sin-activar-y-sin-cotejar` le falta también el suyo.

## Al terminar

```bash
git worktree remove --force "$copia"   # remove-exit=0
git worktree list                       # solo el repositorio
git status --short
```

`git status --short`, igual antes de crear la copia y después de retirarla: las tres líneas
` M specs/022-h23-1-jurisprudencia-coteja-el/gates/{modelos.jsonl,tarea-actual.json,tareas-intentos.json}`, que son
del workflow. Ningún mutante ha estado en el árbol.

## La revisión final: el mutante 3, y los dos de antes con dieciséis casos

Los dos jueces de la revisión final encontraron un mutante que sobrevivía a todo `go test ./internal/evals/`: contar
como cotejo cualquier `cita cotejar`, terminara como terminara. Ningún caso tenía una sesión cuyo único `cita cotejar`
no terminase bien, que FR-002 y los «Edge Cases» del spec cuentan como sesión sin cotejo. El corrector añadió dos casos
a `casosDeLosUmbralesDeJurisprudencia`, `una-con-el-cotejo-fallido-en-el-modo-orden` y
`una-con-el-cotejo-fallido-en-el-modo-herramienta`: la sesión de la 06 del modelo que decide conserva su `cita
cotejar`, que devuelve el sobre de fallo de una ficha que no se reconoce (clase `argumentos`) y, en el modo orden,
termina con 2 (`escribirConElCotejoFallido`). Cada uno espera `documento_sin_cotejar` de su modo con 1 de 12,
`cumple: false`, veredicto `fallo` y el motivo con el nombre de la sesión.

**Dónde.** Sobre la cabeza `daa6867` con los cambios del corrector sin commitear, Go 1.27.2, `darwin/arm64`. Esta vez
no hubo copia: el sandbox de la sesión rechazó la copia de seguridad del fichero con `cp` a `$TMPDIR`, y cada
mutante se aplicó en el árbol con una sola edición de una línea y se deshizo con la edición inversa. Tras los tres,
`git diff --stat -- internal/evals/informe.go internal/evals/umbrales.go` no da nada: el código de producción es el de
la cabeza. La orden de test es la del quickstart, `go test -count=1 -v -run '^TestUmbralesDeJurisprudencia$'
./internal/evals/`, con su salida filtrada por `--- FAIL`, `^ok` y `^FAIL`.

**Sin mutar**: `ok`, y los diecisiete subtests en `PASS`, los dieciséis de la tabla y `el-elemento-del-contrato`.

### Mutante 3: cualquier `cita cotejar` cuenta como cotejo

`internal/evals/informe.go`, línea 1435, en `cotejaUnDocumento`:

```diff
< 		return satisface(juzgada.Invocacion, comandoDelCotejo)
---
> 		return juzgada.Invocacion.Applet == appletDeLaCita && juzgada.Invocacion.Verbo == verboCotejar
```

```text
--- FAIL: TestUmbralesDeJurisprudencia (0.00s)
    --- FAIL: TestUmbralesDeJurisprudencia/una-con-el-cotejo-fallido-en-el-modo-herramienta (0.68s)
    --- FAIL: TestUmbralesDeJurisprudencia/una-con-el-cotejo-fallido-en-el-modo-orden (0.74s)
FAIL
FAIL	github.com/jmorenobl/kitlegal/internal/evals	1.876s
FAIL
```

Los dos casos nuevos, y ninguno más: sin ellos, el mutante pasaba. Tras pasar la preparación de las sesiones del caso
a `armarLasSesionesDelCaso` (por `gocyclo`), los tres mutantes se repitieron sobre el código final, con el mismo
resultado: el 3, los mismos dos en rojo (1.999s); el 1, los mismos cinco; y el 2, dieciséis subtests en rojo. Sin
mutar, diecisiete en `PASS`.

### Mutante 1, con los dieciséis casos

```text
--- FAIL: TestUmbralesDeJurisprudencia (0.00s)
    --- FAIL: TestUmbralesDeJurisprudencia/una-con-el-cotejo-fallido-en-el-modo-herramienta (0.77s)
    --- FAIL: TestUmbralesDeJurisprudencia/una-sin-cotejar-en-el-modo-herramienta (0.70s)
    --- FAIL: TestUmbralesDeJurisprudencia/una-sin-cotejar-en-el-modo-orden (0.72s)
    --- FAIL: TestUmbralesDeJurisprudencia/una-sin-activar-y-sin-cotejar (0.74s)
    --- FAIL: TestUmbralesDeJurisprudencia/una-con-el-cotejo-fallido-en-el-modo-orden (0.76s)
FAIL
FAIL	github.com/jmorenobl/kitlegal/internal/evals	1.937s
FAIL
```

Cinco: los tres de T002 y los dos nuevos, en los que once sesiones del modo cotejan y una no.

### Mutante 2, con los dieciséis casos

`go test` da `FAIL` con los dieciséis `nombre:` de la tabla en rojo, del primero al último, y
`el-elemento-del-contrato` en verde (1.871s). En una segunda ejecución, `actual  : "aprobado"` sale en cuatro:
`una-sin-cotejar-en-el-modo-orden`, `una-sin-cotejar-en-el-modo-herramienta` y los dos del cotejo fallido, los casos
cuyo único motivo es el de `documento_sin_cotejar`.
