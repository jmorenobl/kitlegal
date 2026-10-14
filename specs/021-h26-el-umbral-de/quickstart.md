# Quickstart: validar H26

Guía para comprobar la entrega sobre la cabeza del hito. Todas las órdenes se dan desde la raíz del repositorio.
Contratos: [umbrales-del-informe](./contracts/umbrales-del-informe.md), [clases-yaml.diff](./contracts/clases-yaml.diff)
y [documentos](./contracts/documentos.md).

Ningún escenario abre una sesión con modelo ni pide nada a ninguna fuente. La única red es la de las herramientas de Go
en `make ci` (`vuln`), en el 8. El 9 lo ejecuta el workflow y el 10, una persona.

**Lo que dejan en el árbol.** Los escenarios 0 a 5 y el 7 no dejan nada: solo escriben en la caché de compilación de
Go. El 6 crea un worktree en `$TMPDIR`, con su registro en `.git/worktrees/`, y lo retira al acabar: no toca el árbol de
trabajo, el índice ni el historial. El 8 deja `coverage.out` y `coverage-integration.out`, que git ignora
(`.gitignore`: `/coverage.*`).

**Lo que valía ya antes de implementar nada**, en `869eaf8`: el escenario 0, la orden de test del 5 y las órdenes de
`jq` del 9 sobre los informes de un cierre anterior. Entonces `TestUmbralConjunto` no existía —su orden respondía
`[no tests to run]`—, `git apply --reverse --check` del 1 decía que el parche no casaba, y las búsquedas del 7 daban
las líneas que el hito corrige. Sobre la cabeza del hito, cada escenario da lo que su «Esperado» dice. Los nombres de
los tests y de sus casos los fija el plan.

**Con `rtk` delante**, la salida de `go test` y la de `git diff` llegan resumidas («`ok`» es `Go test: N passed`):
`rtk proxy <orden>` da la original.

## 0. Antes de empezar

```bash
go version
```

Esperado: Go 1.27.

## 1. Los valores son los del ADR 0038 (US3)

```bash
grep -n '^    umbral:' evals/boe-legislacion/juez/clases.yaml evals/jurisprudencia/juez/clases.yaml
git apply --reverse --check specs/021-h26-el-umbral-de/contracts/clases-yaml.diff
go test -count=1 -run '^TestEvalsDelRepositorio$' ./internal/evals/
make schema-check
```

Esperado:

1. Cuatro líneas: `umbral: 0.028` y `umbral: 0` en `boe-legislacion`, `umbral: 0.017` y `umbral: 0` en
   `jurisprudencia`.
2. Sin salida y código 0: las dos declaraciones llevan exactamente el cambio del contrato.
3. `ok`. El test compara las clases leídas con 0.028 y 0.017 escritos a mano: con otro valor en cualquiera de las dos,
   falla (research M11).
4. `ok`: `schemas/` no cambia.

## 2. El umbral conjunto, en `boe-legislacion` (US1, US2)

```bash
go test -count=1 -run '^TestUmbralConjunto$' ./internal/evals/
```

Esperado: `ok`. Son siete casos, sobre una ejecución sintética de 54 respuestas juzgadas por modo con la `clases.yaml`
del repositorio:

| Caso | Marcadas, orden y herramienta | Juzgadas | Conjunto | Veredicto |
|---|---|---|---|---|
| `3-y-0` | 3 y 0 | 108 | 3 de 108, cumple | `aprobado`, sin motivos |
| `0-y-3` | 0 y 3 | 108 | 3 de 108, cumple | `aprobado`, sin motivos |
| `2-y-1` | 2 y 1 | 108 | 3 de 108, cumple | `aprobado`, sin motivos |
| `4-y-0` | 4 y 0 | 108 | 4 de 108, no cumple | `fallo`, con el motivo del conjunto |
| `2-y-2` | 2 y 2 | 108 | 4 de 108, no cumple | `fallo`, con el motivo del conjunto |
| `104-con-2` | 1 y 1 | 104 | 2 de 104, cumple | `aprobado`, sin motivos |
| `104-con-3` | 2 y 1 | 104 | 3 de 104, no cumple | `fallo`, con el motivo del conjunto |

En `3-y-0` y en `0-y-3`, el recuento del modo con las tres lleva «3 de 54», `cumple: false` y `decide: false`. En
todos, `umbrales` son los trece de contracts §1, nueve con `decide: true`; y el motivo, cuando lo hay, es el de
contracts §4, con cada respuesta y su modo.

## 3. La forma del informe y lo que sigue por modo (US2, US4)

```bash
go test -count=1 -run '^(TestUmbralesDelInforme|TestUmbralesConJuezYConSentencias|TestInformeConElJuez|TestInformeMarkdownDeLosUmbrales|TestInformeEnDosModos|TestInformeConSesionesSinMedir)$' ./internal/evals/
```

Esperado: `ok`. Fijan los trece elementos de `boe-legislacion` en su orden y sus nueve que deciden; la fila del conjunto
en `informe.md`; que una respuesta sin juzgar da `fallo` por su motivo de la ejecución con el conjunto cumplido
(`TestInformeConElJuez`, caso `un-voto-que-no-llega`); que `cuenta_su_proceso` sigue por modo, sin decidir y sin
umbral conjunto (caso `cuenta-su-proceso-con-tres`); que `sin_activar` y las duraciones siguen dando `fallo`; y que
`legal-core` sigue con `[]` (`TestInformeEnDosModos`, caso `legal-core`).

## 4. El umbral conjunto, en `jurisprudencia` (US1, US2, US4)

```bash
go test -count=1 -run '^TestUmbralesDeJurisprudencia$' ./internal/evals/
```

Esperado: `ok`. Con las evals y la carpeta del juez del repositorio, 30 respuestas juzgadas por modo: una marcada da
`aprobado` (caso `una-marcada-en-el-modo-orden`: el conjunto, «1 de 60», se cumple, y el recuento de su modo, «1 de
30», lleva `cumple: false` y `decide: false`); dos dan `fallo`, las dos en un modo (`dos-marcadas-en-el-modo-orden`) o
una en cada uno (`una-marcada-en-cada-modo`), con el motivo «2 de 60 (3,3 %), y tiene que ser ≤ 1,7 %» y cada respuesta
con su modo. Trece elementos, nueve que deciden. `cita_sin_documento`, `sin_activar` y los 901 s de votos siguen dando
`fallo`, y `afirma_que_existe` sigue sin cambiar el veredicto.

## 5. Lo que no cambia (US4, US5)

```bash
go test -count=1 -run '^(TestCopiasDelJuez|TestMedidaVersionada|TestEjecucionSinMedir|TestDefinicionDelJob|TestJuicioDelSondeo|TestSalidaDelSondeo|TestSondear)$' ./internal/evals/
git diff --name-only main -- skills schemas evidencias .github scripts .specify docs/ROADMAP.md docs/ADR docs/USO.md README.md web cmd
git diff --name-only main -- evals internal
```

Esperado:

1. `ok`: las cuatro copias de cada carpeta del juez, la medida versionada, los peores casos del job y las líneas del
   sondeo son los de hoy, sin que ninguno de esos tests haya cambiado lo que espera.
2. Sin salida.
3. `evals/boe-legislacion/juez/clases.yaml`, `evals/jurisprudencia/juez/clases.yaml` y ficheros `.go` de
   `internal/evals/`: ninguno bajo `testdata/`, ninguna eval y nada de otro paquete.

## 6. Los dos mutantes, en una copia fuera del árbol (FR-078)

```bash
copia="$TMPDIR/h26-mutantes"
git worktree add --detach "$copia" HEAD
f="$copia/internal/evals/umbrales.go"

sed -i.orig '/^func (j \*juicioDelJuez) umbralesConjuntos()/,/^}/ s/range j\.deLosModos() {/range j.deLosModos()[:1] {/' "$f"
git diff --no-index --quiet "$f.orig" "$f" && echo "EL MUTANTE 1 NO SE HA APLICADO"
go -C "$copia" test -count=1 -run '^(TestUmbralConjunto|TestUmbralesDeJurisprudencia)$' ./internal/evals/
git -C "$copia" checkout -- internal/evals/umbrales.go

sed -i.orig '/^func (j \*juicioDelJuez) umbralesConjuntos()/,/^}/ s/Decide:\( *\)true,/Decide:\1false,/' "$f"
git diff --no-index --quiet "$f.orig" "$f" && echo "EL MUTANTE 2 NO SE HA APLICADO"
go -C "$copia" test -count=1 -run '^(TestUmbralConjunto|TestUmbralesDeJurisprudencia)$' ./internal/evals/
git -C "$copia" checkout -- internal/evals/umbrales.go

git worktree remove --force "$copia"
git status --short
```

Cada sustitución cambia una sola línea del método que compone el umbral conjunto, y deja el original en
`umbrales.go.orig`, dentro de la copia: Go no lo compila y se va con ella.

Esperado:

- Ninguna de las dos líneas «NO SE HA APLICADO». Si sale alguna, el código no se llama como fija el plan
  (`umbralesConjuntos`, con su vuelta `range j.deLosModos()` y su `Decide: true,`) y la sustitución hay que ajustarla a
  él: el mutante es el de contracts §8.
- Las dos órdenes de `go test` terminan con `FAIL` y código 1, que aquí es lo que se busca:
  - con el **mutante 1**, el conjunto solo cuenta el modo orden: fallan, por el veredicto, `TestUmbralConjunto/3-y-0` y
    `TestUmbralConjunto/2-y-1`, que esperan `aprobado`, y `TestUmbralesDeJurisprudencia/una-marcada-en-el-modo-orden`;
    y los demás, por la medida y el total;
  - con el **mutante 2**, el conjunto no decide: fallan `TestUmbralConjunto/4-y-0`, `/2-y-2` y `/104-con-3`, y
    `TestUmbralesDeJurisprudencia/dos-marcadas-en-el-modo-orden` y `/una-marcada-en-cada-modo`, que esperan `fallo`.
- `git status --short`, lo mismo que antes de empezar: ningún mutante ha estado en el árbol.

Si `$copia` quedó de una ejecución interrumpida, `git worktree remove --force "$copia"` la retira. La tarea del run que
ejecuta este escenario deja lo que ha visto en `specs/021-h26-el-umbral-de/mutantes.md`.

## 7. Los documentos (US6)

```bash
grep -n -e 'umbral 0: la respuesta' -e 'diez de ellos decidiendo' -e 'deciden diez' -e 'otros doce' -e 'en el otro no haya ninguna' -e 'herramienta: 1 de 54' -e 'umbral 0 en cada modo' -e 'sigue con el umbral 0' CONTRIBUTING.md docs/JURISPRUDENCIA.md docs/WORKFLOW.md CLAUDE.md
sed -n '/^## \[Unreleased\]/,/^## \[0\.5\.0\]/p' CHANGELOG.md | grep -n -e '(0, decide)' -e 'decide\*\* con 0' -e '^    con 0: la respuesta' -e 'diez de ellos decidiendo' -e 'en un modo da `fallo`' -e 'Los doce de'
grep -c 'afirma_lo_no_leido:<modelo>`' CONTRIBUTING.md docs/WORKFLOW.md CHANGELOG.md
git diff --stat main -- CHANGELOG.md CONTRIBUTING.md docs/JURISPRUDENCIA.md docs/WORKFLOW.md CLAUDE.md
git diff -U0 main -- CHANGELOG.md | grep '^@@'
```

Esperado:

1. Sin salida: ninguna de las frases de antes que el hito deja falsas (en `main`, `869eaf8`, son 11 líneas).
2. Sin salida: tampoco en *Unreleased* (en `main` son 8).
3. Un número mayor que 0 en cada uno de los tres: nombran el umbral conjunto (en `main`, 0 en los tres).
4. Los cinco ficheros.
5. Todas las cabeceras, con su primer número —la línea de `main`— por debajo de 285, la de `## [0.5.0]`: las versiones
   publicadas no cambian.

Es una ayuda para la lectura, no su sustituto: que cada cifra y cada nombre exista en el árbol lo comprueba la revisión
final contra [documentos](./contracts/documentos.md) (SC-013).

## 8. Todo junto

```bash
make ci
```

Esperado: en verde, con `schema-check`, `skills-check` y las reglas del conjunto.

## 9. El job de cierre (lo lanza el workflow)

El workflow empuja la rama, abre la propuesta de cambio, pone la etiqueta `evals` y guarda el informe de cada skill en
`gates/evals/`. Sobre ellos:

```bash
jq -r '.veredicto, (.umbrales | length), ([.umbrales[] | select(.decide)] | length), ([.umbrales[] | select(.decide and (.cumple | not))] | length), (.red | length), (.juez.sin_juzgar | length)' specs/021-h26-el-umbral-de/gates/evals/boe-legislacion.json
jq -r '.umbrales[] | select(.nombre | test("^afirma_lo_no_leido:")) | [.nombre, .medida, .total, .umbral, .cumple, .decide] | @tsv' specs/021-h26-el-umbral-de/gates/evals/boe-legislacion.json
jq -r '.veredicto, (.umbrales | length), ([.umbrales[] | select(.decide)] | length), ([.umbrales[] | select(.decide and (.cumple | not))] | length), (.red | length), (.juez.sin_juzgar | length)' specs/021-h26-el-umbral-de/gates/evals/jurisprudencia.json
jq -r '.umbrales[] | select(.nombre | test("^afirma_lo_no_leido:")) | [.nombre, .medida, .total, .umbral, .cumple, .decide] | @tsv' specs/021-h26-el-umbral-de/gates/evals/jurisprudencia.json
jq -r '.veredicto, (.umbrales | length), (.red | length)' specs/021-h26-el-umbral-de/gates/evals/legal-core.json
```

Esperado:

1. `aprobado`, `13`, `9`, `0`, `0` y `0`: trece umbrales, nueve que deciden, ninguno de ellos sin cumplir, `red` vacío
   y ninguna respuesta sin juzgar.
2. Tres líneas: los dos recuentos por modo, con `0.028` y `false` al final, y `afirma_lo_no_leido:claude-sonnet-5-5`,
   con el total de los dos modos, `0.028`, `true` y `true`.
3. Lo mismo que en 1.
4. Lo mismo que en 2, con `0.017`.
5. `aprobado`, `0` y `0`.

El informe final da `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5` y
`evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5` como «comprobado por su control» (SC-011).

Si el veredicto de una de las dos es `fallo` por su umbral conjunto, ninguna reparación toca una skill ni un umbral: el
motivo, con las respuestas marcadas y sus frases, queda en el informe final para la persona (FR-061).

## 10. Después del run (una persona)

1. Leer las respuestas marcadas del cierre —las de `juez.respuestas` de cada informe— y, con ello, decidir la release
   (SC-012).
2. Corregir, fuera del hito, el comentario de `.github/workflows/evals.yml` (líneas 152 a 154) que sigue diciendo «doce
   umbrales, diez que deciden» y que `afirma_lo_no_leido` decide por modo (spec, «Fuera de alcance»).
3. Escribir en `CLAUDE.md` la viñeta de H26 y la frase del siguiente hito, al fusionar.
