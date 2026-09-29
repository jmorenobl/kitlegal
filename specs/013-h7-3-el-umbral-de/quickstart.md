# Quickstart: validar H7.3

Escenarios para comprobar la entrega una vez implementada, desde la raíz del repositorio y en la rama del hito. Los §1
a §5 y el §7 no usan red ni modelo (el §8, `make ci`, sí red para `govulncheck`, como siempre); el §6 abre sesiones de
Claude Code con la suscripción de quien lo ejecuta y lo ejecuta la persona al leer el informe final, en un Mac sin
strace, fuera de `make ci` y del run (FR-068, FR-070); el §7 lee el informe del job de cierre, que deja el workflow.

Efectos: ninguno en el índice ni en el historial de git. En el árbol de trabajo, solo el §8 escribe `coverage.out` y
`coverage-integration.out`, que git ignora (`/*.out`, `/coverage.*`). Los §4, §5 y §6 escriben en directorios
temporales que borran ellos mismos (los de `make evals-sondeo`, su guion; los de las copias, la última orden del
escenario); `go` escribe en su caché de compilación, fuera del árbol. Formatos:
[contracts/lista-de-expresiones.md](./contracts/lista-de-expresiones.md),
[contracts/skill-boe-legislacion.md](./contracts/skill-boe-legislacion.md),
[contracts/informe-del-job.md](./contracts/informe-del-job.md),
[contracts/ejecucion-del-job.md](./contracts/ejecucion-del-job.md) y [contracts/sondeo.md](./contracts/sondeo.md).

## 1. La lista, la prosa de la skill y el calibrado, sin modelo (FR-010 a FR-024, FR-091, FR-095; SC-003 a SC-005)

```bash
go test -count=1 -v -run '^TestEvalsDelRepositorio$' ./internal/evals/ | grep -E -- '--- (PASS|FAIL)'
go test -count=1 -v -run '^(TestProsaDeLaSkill|TestExtraerExpresionesProhibidas|TestJuzgarLasExpresionesProhibidas)$' ./internal/evals/ | grep -E -- '--- (PASS|FAIL)'
make skills-check
wc -l < skills/boe-legislacion/SKILL.md
```

Esperado: `--- PASS` en `TestEvalsDelRepositorio` y en todas sus subpruebas, entre ellas `expresiones-calibradas` (35 de
93 en H7.1 y 10 de 93 en H7.2, con el reparto de contracts/lista-de-expresiones.md §4), `expresiones-en-los-bloques`,
`expresiones-de-la-skill` y `prosa-de-la-skill`; `--- PASS` en los otros tres tests; ningún `--- FAIL`; `make
skills-check` en verde; y menos de 300 líneas.

## 2. El informe: umbrales, sesiones sin medir, reintentos y duración, sin modelo (FR-001 a FR-008, FR-033, FR-040 a FR-043, FR-050, FR-051; SC-006, SC-007)

```bash
go test -count=1 -v -run '^(TestUmbralesDelInforme|TestInformeMarkdownDeLosUmbrales|TestInformeConSesionesSinMedir|TestLeerSesionConReintentos|TestClasificarElLimite)$' ./internal/evals/ | grep -E -- '--- (PASS|FAIL)'
```

Esperado: `--- PASS` en los cinco y en sus casos (3 de 51 → `fallo`; 2 de 51 → sin cambio; Haiku 4.5 incumplido → sin
cambio; `[]` sin lista; 901 s → `fallo`; 900 s → sin cambio; (a), (b), (c), sin abrir y 429 recuperado); ningún
`--- FAIL`.

## 3. El repartidor, con los sustitutos de `claude` y `strace` (FR-030 a FR-032, FR-036, FR-037, FR-044, FR-094; SC-008)

```bash
go test -count=1 -v -run '^(TestEjecutarSesiones|TestTopeDeLaSesion)' ./internal/evals/ | grep -E -- '--- (PASS|FAIL)'
```

Esperado: `--- PASS` en `TestEjecutarSesionesEnParalelo` (≤ 4 a la vez con 4, 1 con 1, directorios distintos, el mismo
informe), `TestEjecutarSesionesTrasElLimiteDeUso`, `TestEjecutarSesionesConElContextoCancelado` y `TestTopeDeLaSesion`
(124 y 137); ningún `--- FAIL`. Funciona igual en macOS: ni strace, ni sudo, ni `timeout`.

## 4. La definición del job y lo que la hace fallar (FR-030, FR-034, FR-035, FR-094; SC-008)

```bash
go test -count=1 -run '^TestDefinicionDelJob$' ./internal/evals/
copia=$(mktemp -d "${TMPDIR:-/tmp}/kitlegal-h73-definicion.XXXXXX")
git archive HEAD | tar -x -C "$copia"
sed -i.orig 's/cancel-in-progress: false/cancel-in-progress: true/' "$copia/.github/workflows/evals.yml"
(cd "$copia" && go test -count=1 -run '^TestDefinicionDelJob$' ./internal/evals/); echo "código: $?"
rm -rf "$copia"
```

Esperado: la primera orden, `ok`; en la copia, `FAIL` con un mensaje que nombra `cancel-in-progress` y lo esperado
(`false`), y `código: 1`. La copia sale de `git archive` y se borra: ni el árbol ni el índice cambian.

## 5. El sondeo sin modelo (FR-060 a FR-068, FR-096; SC-009)

```bash
go test -count=1 -v -run '^(TestJuicioDelSondeo|TestSalidaDelSondeo|TestComprobarElSondeo|TestSondear|TestGuionDelSondeo)$' ./internal/evals/ | grep -E -- '--- (PASS|FAIL)'
go test -count=1 -tags integration -run '^TestPrepararElArbolDelSondeo$' ./internal/evals/
make help | grep evals-sondeo
env -u CLAUDE_CODE_OAUTH_TOKEN make evals-sondeo SKILL=boe-legislacion EVALS=03 MODELO=claude-sonnet-5 REPETICIONES=1; echo "código: $?"
```

Esperado: `--- PASS` en los cinco; `ok` en el de integración; la línea de `evals-sondeo` en la ayuda; y en la última,
por la salida de error, `falta la credencial: CLAUDE_CODE_OAUTH_TOKEN…`, `código: 2` (el de `make`) y ninguna sesión
abierta: la comprobación de la credencial va antes de construir nada. Su directorio temporal ya no existe al terminar.

## 6. El sondeo reproduce lo que mide el job (SC-002; FR-070) — con modelo, lo ejecuta la persona

Requisitos: un Mac sin strace, Go, Claude Code (`claude` en el `PATH`) y el token de la suscripción en el entorno
(`claude setup-token`). Consume 30 sesiones de Sonnet 5 de esa suscripción.

```bash
export CLAUDE_CODE_OAUTH_TOKEN='<el token que da claude setup-token>'
claude --version
copia=$(mktemp -d "${TMPDIR:-/tmp}/kitlegal-h73-v012.XXXXXX")
git archive HEAD | tar -x -C "$copia"
git show main:skills/boe-legislacion/SKILL.md > "$copia/skills/boe-legislacion/SKILL.md"
inicio=$(date +%s)
make -C "$copia" evals-sondeo SKILL=boe-legislacion EVALS=03,06,13,14,15 MODELO=claude-sonnet-5 REPETICIONES=3
echo "v0.1.2: $(( $(date +%s) - inicio )) s"
rm -rf "$copia"
inicio=$(date +%s)
make evals-sondeo SKILL=boe-legislacion EVALS=03,06,13,14,15 MODELO=claude-sonnet-5 REPETICIONES=3
echo "v0.1.3: $(( $(date +%s) - inicio )) s"
```

La copia es la rama del hito con la `SKILL.md` de `main` (v0.1.2): el binario que construye el sondeo en ella la lleva
empotrada y la instala. Esperado, en la línea «Respuestas con alguna expresión prohibida…» de cada salida (que empieza
por «Esto es un sondeo, no un veredicto»): **al menos 5 de 15** con v0.1.2 (el job de cierre de H7.2 dio 9 de 15 en
estas evals) y **como mucho 2 de 15** con v0.1.3. Si alguna sesión sale sin medir por límite de uso, el denominador baja
y el sondeo se repite cuando se reponga el límite. Se anota:

| Sondeo | `claude --version` | Respuestas con alguna expresión prohibida | Duración |
|---|---|---|---|
| v0.1.2 (copia) | | | |
| v0.1.3 (rama) | | | |

## 7. El job de cierre (SC-001; FR-098)

Tras el cierre del workflow, que deja los informes en `gates/evals/`:

```bash
jq '{veredicto, red, duracion_de_las_sesiones, sesiones_sin_medir, umbrales: [.umbrales[] | {nombre, medida, total, cumple, decide}]}' specs/013-h7-3-el-umbral-de/gates/evals/boe-legislacion.json
jq '.umbrales' specs/013-h7-3-el-umbral-de/gates/evals/legal-core.json
```

Esperado en `boe-legislacion`: `veredicto` `aprobado`, `red` `[]`, `sesiones_sin_medir` `[]`,
`expresiones_prohibidas:claude-sonnet-5` con `medida` ≤ 2, `total` 51, `cumple: true`, `decide: true`; el de
`claude-haiku-4-5-20251001` con `total` 30 y `decide: false`; `duracion_de_las_sesiones` con `medida` ≤ 900, `cumple:
true` y `decide: true`. En `legal-core`, `[]`. En el informe final, los dos umbrales del plan como «comprobado por su
control».

## 8. `make ci` (FR-090; SC-010)

```bash
make ci
```

Esperado: `ci: todos los controles en verde`, con `skills-check` y `schema-check` sin drift.
