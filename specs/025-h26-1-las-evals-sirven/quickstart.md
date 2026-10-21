# Quickstart: validar H26.1

Guía para comprobar la entrega sobre la cabeza del hito. Todas las órdenes se dan desde la raíz del repositorio, con
`bash` o `zsh`. Contratos: [formato de eval](./contracts/formato-de-eval.md), [caché de la sesión](./contracts/cache-de-la-sesion.md),
[juicio e informe](./contracts/juicio-e-informe.md) y [skill legal-core](./contracts/skill-legal-core.md).

Ningún escenario abre una sesión con modelo ni pide nada a ninguna fuente. La única red es la de las herramientas de Go
en `make ci` (`vuln`), en el 8. El 10 lo ejecuta el workflow en el cierre: aquí solo se lee lo que deja.

**Lo que dejan en el árbol.** Los escenarios 0 a 6 y el 9 no dejan nada: solo escriben en la caché de compilación de
Go. El 7 crea un worktree en `$TMPDIR/h261-copia`, con su registro en `.git/worktrees/`, y lo retira al acabar: no toca
el árbol de trabajo, el índice ni el historial. El 8 deja `coverage.out` y `coverage-integration.out`, que git ignora
(`.gitignore`: `/coverage.*`), y su registro en `$TMPDIR`.

**Antes de implementar nada**, en `6bf2da3`, valen el 0, el 5 (con 16 261 s), el 9 y, del 7.b, `TestTuberiaCerrada` en
`FAIL` con el mutante; los tests nuevos no existen (`go test` avisa de que no hay ninguno que ejecutar) y
`TestEvalsDelRepositorio` pasa sin el subtest `lo-que-lleva-a-leer`. Sobre la cabeza del hito, cada escenario da lo que
su «Esperado» dice.

**Con `rtk` delante**, la salida de `go test`, `git diff` y `diff` llega resumida («`ok`» es `Go test: N passed`):
`rtk proxy <orden>` da la original.

## 0. Antes de empezar

```bash
go version
git status --short
git worktree list
```

Esperado: Go 1.27; la lista de lo que haya sin commitear y un solo worktree, que el escenario 7 tiene que dejar igual.

## 1. La caché sirve lo que la skill lleva a leer (US1; FR-001 a FR-006, FR-010, FR-011, FR-013)

```bash
go test -count=1 -v -run '^TestEvalsDelRepositorio$' ./internal/evals/
go test -count=1 -v -run '^(TestCacheDeLasSesiones|TestLoQueLlevaALeerSinGrabar|TestBaseDeLaReconstruccion|TestConsultasNecesarias)$' ./internal/evals/
```

Esperado:

1. `ok`. El subtest `lo-que-lleva-a-leer` en `PASS`: desde la caché preparada de una sesión de una eval de cada skill
   (`boe-legislacion`, `legal-core` y `jurisprudencia`), cada bloque de `data/plazos.yaml` y de `data/cobertura.yaml`,
   sin fecha y con `--fecha 9999-12-31`, y el índice y los metadatos de cada norma de los tres ficheros de `data/`
   salen con 0 con `--offline`; hoy, 17 bloques y 19 normas. `grabado` y `conjunto-legal-core` en `PASS`, con las
   `lecturas` de la 14, la 15 y la 16 grabadas y las doce evals de `legal-core` bien formadas.
2. `ok`. `TestCacheDeLasSesiones`: desde la caché de la 11, `boe articulo BOE-A-2015-10565 a124 --offline` sale con 4
   sin fecha, con `--fecha 2026-09-09` y con `--fecha 9999-12-31`, y desde la de la 05, con 0 en las tres; desde la de
   la 07, `daprimera` sale con 0 en esas tres fechas, que ninguna eval de `legal-core` nombra, y desde la de la 12, con
   4 en las tres; desde la de cualquier eval de `boe-legislacion`, `a26` y
   `a27` de `BOE-A-2015-10566`, `a18` de `BOE-A-2006-20764` y `a15` de `BOE-A-2004-4214`, con 0 sin fecha y con
   `9999-12-31`. `TestLoQueLlevaALeerSinGrabar`: con una copia de `data/` con un precepto sin grabar, una falta que
   empieza por `data/plazos.yaml: lo que la skill lleva a leer` y nombra la orden. `TestBaseDeLaReconstruccion`: la base
   de la medida de `boe-legislacion` sirve `daprimera`, que no nombra ninguna de sus evals. `TestConsultasNecesarias`:
   cada `lecturas` da su bloque sin fecha y con `9999-12-31`, con el punto `lectura de la eval`.

## 2. `lecturas` y `sin_servir` en el formato (US2, US3; FR-008, FR-009, FR-012, FR-014)

```bash
go test -count=1 -v -run '^(TestFormatoDeLecturas|TestLeerEval|TestEsquemaDeEval)$' ./internal/evals/
```

Esperado: `ok`. `TestFormatoDeLecturas` lee una eval con `lecturas` y `sin_servir` (un bloque y una consulta de norma) y
rechaza, cada una con el nombre del fichero y la entrada: una lectura sin `motivo`, con el motivo vacío y con solo
blancos; una lectura cuyo bloque ya es un comando de bloque, con y sin `fecha`; una cuyo bloque es una cita; un bloque
de `sin_servir` que es una cita; un `verbo` que no es `indice` ni `metadatos`; y una entrada con `bloque` y `verbo`.

## 3. El juicio (US2, US3, US5; FR-007, FR-015 a FR-018, FR-023 a FR-028)

```bash
go test -count=1 -v -run '^(TestJuzgarLasLecturasSinServir|TestJuzgarElMotivoDeLasFallidas|TestLineaDeLaSalidaNoEscrita|TestLeerLasLlamadas|TestTextosDeLaSesion|TestJuzgar|TestJuzgarLasLlamadas|TestJuzgarPlazos)$' ./internal/evals/
```

Esperado: `ok`.

- `TestJuzgarLasLecturasSinServir`: la sesión de la 11 que resuelve Leganés, pide `plazos calcular reposicion`, lee
  `a124` con `--fecha 2026-09-09`, sale con 5 y da `⚠ PLAZO: 2026-10-13` pasa, con la lectura en `lecturas_sin_servir`
  y no en `fuera_de_lo_grabado` ni en `otras_fallidas`; la que la lee sin fecha o en otra fecha, y la que no la lee,
  llevan el motivo del comando ausente; la de la 12 con `⚠ NO CUBIERTO:` y `BOE-A-2015-10565` pasa, y la que ejecuta
  `plazos calcular` no; la de la 16 que no lee `a26` ni `a27` tiene el mismo juicio que sin `lecturas`.
- `TestJuzgarElMotivoDeLasFallidas`: la llamada a `boe_indice` sin sobre sale con `codigo: null` y `motivo: "sin_sobre"`
  en `invocaciones` y en `otras_fallidas`, también si consulta una lectura de `sin_servir`; la orden con `| python3`
  cuya salida trae la línea de la escritura fallida, con `codigo: 1`, `motivo: "salida_no_escrita"` y `bash`, en
  `otras_fallidas`, o en `lecturas_sin_servir` si consulta una lectura de `sin_servir`; el 1 con otra salida de error,
  sin enlace o con dos usos que lo cumplen, como hoy; y en todos, `pasa` y `motivos` iguales a los de antes.
- `TestLineaDeLaSalidaNoEscrita`: la línea que deja `app.Main` en la salida de error con una salida estándar que falla
  es la que el arnés reconoce.
- Los demás, con lo que esperaban, salvo los casos nuevos de `TestLeerLasLlamadas` (`SinSobre`) y de
  `TestTextosDeLaSesion` (`OrdenesDeBash`).

## 4. El informe (US2, US4; FR-015, FR-019, FR-021, FR-022, FR-033)

```bash
go test -count=1 -v -run '^(TestUmbralesDeLasLecturas|TestUmbralesDeLasTresSkills|TestUmbralesDeLegalCore|TestUmbralesDeBoeLegislacion|TestUmbralesDeJurisprudencia|TestUmbralesDelInforme|TestInformeEnDosModos|TestInforme)$' ./internal/evals/
```

Esperado: `ok`.

- `TestUmbralesDeLasLecturas`: con tres sesiones medidas de `claude-sonnet-5-5` en el modo orden, dos lecturas con 5
  fuera de la caché en una y una de `sin_servir` con 5 en otra, `fuera_de_lo_grabado:claude-sonnet-5-5:orden` con
  medida 2, total 3, `decide: false` y el veredicto igual; un modelo y un modo sin ninguna, con medida 0; la serie de la
  11 con `sesiones_con_lectura_sin_servir` 3, también 0 si ninguna la intenta, y `null` en la de otra eval;
  `lectura_sin_intentar:claude-sonnet-5-5:orden` con medida 0 cuando las tres la intentan, también con la orden de
  salida no escrita, y con medida 1, su motivo con la sesión y el veredicto `fallo` cuando una no la intenta o solo la
  intenta una llamada sin sobre; el del modelo informativo, con `decide: false`; ausente en una skill sin eval así.
- `TestUmbralesDeLasTresSkills`: `legal-core`, 18 umbrales y 12 que deciden; `boe-legislacion` y `jurisprudencia`, 19 y
  11, sin ningún `lectura_sin_intentar`; con el plan del job, en su orden (contracts/juicio-e-informe.md §4).

## 5. El tope del trabajo de `legal-core` (FR-020, SC-009)

```bash
go test -count=1 -v -run '^TestDefinicionDelJob$' ./internal/evals/
grep -n 'peorCasoDeLegalCore = ' internal/evals/definicion_test.go
grep -n 'timeout-minutes: 352' .github/workflows/evals.yml
```

Esperado: `ok`; `peorCasoDeLegalCore = "19525 s = 485 s + (⌈67 / 2⌉ + ⌈66 / 2⌉ + ⌈6 / 2⌉)…"` (los caracteres `⌈` y `⌉`
van escritos `\xe2\x8c\x88` y `\xe2\x8c\x89`); y el tope, 352, sin cambiar: 19 525 s caben en 21 120 s.

## 6. `legal-core` v1.2.3 (US6; FR-029 a FR-031)

```bash
wc -l skills/legal-core/SKILL.md
make skills-check
grep -n 'legal-core. v1.2.3' CHANGELOG.md
```

Esperado: 299 líneas o menos; `ok`; y la entrada de *Unreleased*. Los pasos 7 y 10 dicen lo de
contracts/skill-legal-core.md §1 (lo lee la revisión final).

## 7. Un precepto sin grabar y los mutantes, en una copia fuera del árbol (FR-003, FR-025, SC-007, SC-008)

```bash
copia="$TMPDIR/h261-copia"
git worktree add --detach "$copia" HEAD

# a. Un precepto de data/ sin grabar: TestEvalsDelRepositorio sale en rojo y nombra la lectura (FR-003).
f="$copia/data/plazos.yaml"
sed -i.orig 's/bloque: a128,/bloque: a129,/' "$f"
git diff --no-index --quiet "$f.orig" "$f" && echo "LA SUSTITUCIÓN NO SE HA APLICADO"
go -C "$copia" test -count=1 -run '^TestEvalsDelRepositorio$' ./internal/evals/ 2>&1 | grep -E 'a129|FAIL' | head -5
mv "$f.orig" "$f"

# b. El mensaje de la escritura fallida cambiado en el binario (FR-025).
f="$copia/internal/render/render.go"
sed -i.orig 's/render: no se pudo escribir en %s/render: no se ha podido escribir en %s/' "$f"
git diff --no-index --quiet "$f.orig" "$f" && echo "EL MUTANTE NO SE HA APLICADO"
go -C "$copia" test -count=1 -run '^TestTuberiaCerrada$' ./internal/app/ 2>&1 | tail -3
go -C "$copia" test -count=1 -run '^TestLineaDeLaSalidaNoEscrita$' ./internal/evals/ 2>&1 | tail -1
mv "$f.orig" "$f"

git worktree remove --force "$copia"
git worktree list
git status --short
```

Esperado:

- a. Ninguna línea «NO SE HA APLICADO»; la salida, con `FAIL` y la falta
  `data/plazos.yaml: lo que la skill lleva a leer boe articulo BOE-A-1998-16718 a129: …`.
- b. `TestTuberiaCerrada` en `FAIL` (el binario ya no dice «no se pudo escribir en la salida estándar»), y
  `TestLineaDeLaSalidaNoEscrita` en `ok`: el arnés no copia el mensaje, lo toma del binario.
- Al final, `git worktree list` y `git status --short` como en el escenario 0.

Los otros tres mutantes —sin la derivación de `data/` (FR-002), con el `sin_servir` de cualquier eval aplicado a todas
las de la skill (FR-013) y con `lectura_sin_intentar` sin decidir (FR-033)— los deja la tarea que cierra cada pieza en
[`mutantes.md`](./mutantes.md), cada uno con su sustitución sobre `$copia`, su comprobación de que se aplicó y el test
que tiene que salir en rojo. Se ejecutan igual, entre el `git worktree add` y el `git worktree remove` de arriba,
restaurando el fichero tras cada uno.

## 8. `make ci` (SC-001)

```bash
make ci > "$TMPDIR/h261-ci.log" 2>&1; echo "código: $?"
tail -1 "$TMPDIR/h261-ci.log"
```

Esperado: `código: 0` y `ci: todos los controles en verde`.

## 9. Lo que no cambia (FR-028; D8 y D15 de research)

```bash
git diff --name-only main -- cmd internal/app internal/cli internal/render internal/source internal/core internal/httpx internal/cache internal/graph internal/mcp .github testdata evidencias
```

Esperado: sin salida. El binario, sus códigos de salida, su presentador y el job no cambian; ni las grabaciones de H4 y
de H5 ni ninguna evidencia.

## 10. El cierre (lo mide el workflow; SC-011 a SC-020)

Tras el cierre, el workflow deja el informe de cada skill en `specs/025-h26-1-las-evals-sirven/gates/evals/<skill>.json`.

```bash
d=specs/025-h26-1-las-evals-sirven/gates/evals
jq -r '.veredicto' "$d/legal-core.json" "$d/boe-legislacion.json" "$d/jurisprudencia.json"
jq -c '.umbrales[] | select(.decide) | {nombre, medida, total, cumple}' "$d/legal-core.json"
jq -c '.tasas[] | select(.eval | test("^1[12]-")) | {eval, modelo, modo, pasan, sesiones, sesiones_con_lectura_sin_servir}' "$d/legal-core.json"
jq -c '.umbrales[] | select(.nombre | startswith("fuera_de_lo_grabado")) | {nombre, medida, total, decide}' "$d/legal-core.json" "$d/boe-legislacion.json"
jq -c '[.evals[].otras_fallidas[] | select(.codigo == 1)] | group_by(.motivo) | map({motivo: .[0].motivo, n: length})' "$d/boe-legislacion.json"
jq -r '.red | length' "$d/legal-core.json"
```

Esperado: `aprobado` en los tres; en `legal-core`, los doce que deciden con `cumple: true`; las series de la 11 y la 12
del modelo que decide, con `pasan` al menos 2 de 3 y la 11 con `sesiones_con_lectura_sin_servir` 3 en los dos modos;
los `fuera_de_lo_grabado`, con `decide: false`; ninguna llamada sin sobre entre los `1` de `boe-legislacion`, y los de
salida no escrita con su motivo (lo que salga confirma o descarta la causa del ADR 0043, SC-017); y `red` vacía.
