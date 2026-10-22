# Quickstart: validar H30 en local

Escenarios para comprobar la entrega una vez implementada, desde la raíz del repositorio, en la rama del hito. Ninguno
abre sesiones con modelo, toca la red ni la plataforma, ni deja nada en el árbol de trabajo, el índice o el historial:
los tests escriben solo en sus `t.TempDir()` y en la caché de Go, fuera del repositorio. Contratos en
[contracts/](contracts/); entidades en [data-model.md](data-model.md).

**Requisitos.** Go 1.27.3 (el `toolchain` de `go.mod`), git, y la caché de módulos con las dependencias del módulo
(la deja cualquier `make build` o `make ci` anterior): `go list` y los clones de los tests no piden nada a la red.

## 1. Cada eval tiene su conjunto (FR-001, FR-005, SC-002)

```sh
go test -count=1 -run '^TestConjuntoDeCadaEval$' -v ./internal/evals/
```

Esperado: `PASS`. El registro da, por eval, cuántas entradas y módulos tiene su conjunto (45 líneas hoy). El subtest del
mutante (un `SKILL.md` en memoria con `kitlegal inexistente` en su tabla) comprueba que el error nombra la eval.

## 2. La elección de los mutantes y el golden de la matriz (FR-002 a FR-008, FR-019, SC-003, SC-004)

```sh
go test -count=1 -run '^(TestEleccionDeLosMutantes|TestMatrizDeLasPropuestas)$' -v ./internal/evals/
```

Esperado: `PASS`. Los subtests de `TestEleccionDeLosMutantes`: `referencia-de-legal-core` (las 12 de `legal-core`),
`cuerpo-de-boe-legislacion` (las 45: 23 + 12 + 10), `cuerpo-de-jurisprudencia` (sus 10), `internal-cli` (45),
`sqlite-en-go-mod` (45), `descripcion-de-legal-core` (45), `paquete-sin-asignar` (45, con la ruta), `solo-docs` (0),
`una-eval` (`evals/jurisprudencia/03-por-materia.yaml`, solo esa) y `referencia-renombrada` (las 12 de `legal-core`).

## 3. La huella (FR-009, SC-008)

```sh
go test -count=1 -run '^TestHuellaDeLasEvals$' -v ./internal/evals/
```

Esperado: `PASS`. Un subtest por clase de ruta de FR-001 a FR-004; `go-mod-de-lo-comun`, `replace-en-go-mod` y
`go-sum-de-otro-modulo` cambian las 45; `docs` no cambia ninguna; `dos-clones` clona la cabeza dos veces en directorios temporales (`git clone` local, sin red) y
compara las 45 huellas (el subtest se llama `dos-clones` y no `dos-directorios`: `misspell` lee «directorios» como
«directors» y la tarea no puede tocar `.golangci.yml`).

## 4. La reutilización y la decisión del job (FR-010 a FR-014, FR-016, FR-017, SC-006, SC-007)

```sh
go test -count=1 -run '^(TestBusquedaDeMedidasDeMain|TestDecisionDelJob|TestOrdenesDeLaPlataforma|TestInformeDelRegistro)$' -v ./internal/evals/
```

Esperado: `PASS`. En `TestDecisionDelJob`, `solo-el-cuerpo-de-jurisprudencia` planifica 0 sesiones de
`boe-legislacion` y de `legal-core`, y `etiqueta-evals-sin-cambios` planifica 0 sesiones y deja las 45 reutilizadas con
su huella; `sin-medida-de-main` planifica solo las sesiones de la eval sin resultado, con `sin_resultado_de_main`; y
`ficheros-sin-leer` (el `gh` de los ficheros de la propuesta falla) mide todas las evals de la skill con `plataforma` y
deja la orden y su error en `origen.error_de_la_plataforma`. En `TestBusquedaDeMedidasDeMain`, con un instante fijo,
una eval que la medida más reciente reutilizó de una ejecución de hace 90 días se mide con `sin_resultado_de_main`, y
una orden de `gh` que falla a mitad de la búsqueda la para: lo que queda se mide y la orden y su error quedan en el
resultado.

## 5. El informe con lo reutilizado (FR-015, FR-022, FR-026, SC-012)

```sh
go test -count=1 -run '^(TestInformeConReutilizadas|TestInformeDeLaSuiteCompletaNoCambia)$' -v ./internal/evals/
```

Esperado: `PASS`. El segundo compara, byte a byte y sin el bloque `origen`, el informe de la suite completa de unos datos
de prueba con el que daba el código antes del hito.

## 6. La serie de cierres no ve lo reutilizado (FR-023, SC-009)

```sh
go test -count=1 -run '^TestFilaConReutilizadas$' -v ./internal/evals/
git diff --quiet main -- scripts/workflow/serie.sh && echo "serie.sh es el de main"
```

Esperado: `PASS` y `serie.sh es el de main`.

## 7. El límite de la tasa (FR-024, SC-010, SC-013)

```sh
go test -count=1 -run '^TestLimiteDeLaTasa$' -v ./internal/evals/
```

Esperado: `PASS`, y en el registro la tasa de referencia de cada skill con juez: `boe-legislacion` 120 respuestas,
límite 3, tasa de referencia 0,0085; `jurisprudencia` 60, límite 1, 0,0036. El subtest `umbral-estrechado` da límite 2
con 120 respuestas y `umbral: 0.017`.

## 8. La definición del job y la tanda (FR-016 a FR-021, SC-011)

```sh
go test -count=1 -run '^(TestDefinicionDelJob|TestTandaConLaSuiteCompleta)$' -v ./internal/evals/
```

Esperado: `PASS`. El subtest del tope recalcula el peor caso con los 15 minutos de la decisión (19 877 s en
`boe-legislacion`, dentro de 352 minutos).

## 9. La release (FR-025, SC-005)

```sh
go test -count=1 -run '^TestConfiguracionDeLaRelease$' -v .
go test -count=1 -run '^TestComprobarLaSuiteCompleta$' -v ./internal/evals/
```

Esperado: `PASS` en los dos; el subtest `suite-completa` del primero y los casos del segundo (sin ejecución, otro
commit, en rojo, informe que falta, eval reutilizada) dan cada uno su motivo.

## 10. Lo que el hito no toca (FR-027, SC-015) y la documentación (SC-013, SC-016)

```sh
git diff --name-only main...HEAD -- scripts/workflow .specify internal/workflow
ls docs/ADR/0045-*.md
grep -n "evals-completa" CHANGELOG.md docs/WORKFLOW.md CONTRIBUTING.md .github/workflows/evals.yml
```

Esperado: la primera orden no imprime nada; la segunda, el ADR; la tercera, al menos una línea por fichero.

## 11. Todo junto (SC-001)

```sh
make ci
```

Esperado: termina con 0.

## En la plataforma (no local)

La medición de cierre la hace el workflow tras la revisión (SC-014): pone la etiqueta `evals`, y como el hito toca lo
común, el `origen` de cada informe da las 45 evals medidas con motivo `comun`. La primera reutilización real se ve en la
primera propuesta posterior a la fusión.
