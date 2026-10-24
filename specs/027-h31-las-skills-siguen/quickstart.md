# Quickstart: validar H31

Escenarios para comprobar la entrega en una copia de trabajo de la rama del hito, desde la raíz del repositorio, sin
red ni modelo. Ninguno escribe en el árbol, el índice ni el historial de git, salvo el 3, que en un árbol sincronizado
no escribe nada y lo comprueba. Los nombres de los tests son los de plan.md («Controles mecánicos»). Lo que se mide en
el job de evals (SC-010) y la medida del juez con la etiqueta `evals-medir-juez` (SC-011) no son de este quickstart: los
miden el cierre del workflow y una persona.

Requisitos: Go del `go.mod` (1.27.x), `make`, las herramientas que comprueba `make check-tools`.

## 1. Las medidas de cada skill

```sh
LC_ALL=en_US.UTF-8 awk 'f>=2{print} /^---/{f++}' skills/boe-legislacion/SKILL.md | LC_ALL=en_US.UTF-8 wc -lm
LC_ALL=en_US.UTF-8 awk 'f>=2{print} /^---/{f++}' skills/legal-core/SKILL.md | LC_ALL=en_US.UTF-8 wc -lm
LC_ALL=en_US.UTF-8 awk 'f>=2{print} /^---/{f++}' skills/jurisprudencia/SKILL.md | LC_ALL=en_US.UTF-8 wc -lm
```

Esperado: en cada una, menos de 500 líneas y como mucho 11 000 caracteres (en `main`: 283 y 24 439, 283 y 19 225, 193 y
13 637). Solo lee. El gate que decide es el del escenario 2; esta orden cuenta igual que él mientras el `SKILL.md`
termine en salto de línea.

## 2. Los gates de `skills-check` sobre las tres skills y sus mutantes (SC-001, SC-002, SC-009)

```sh
make skills-check
go test -count=1 -run '^TestSkillsDelRepositorio$' ./internal/app/
go test -count=1 ./internal/skills/
```

Esperado: las tres salen con 0. `TestSkillsDelRepositorio` lleva los subtests `frontmatter`, `presupuestos`,
`referencias`, `deriva`, `sincronia-con-referencias-a-mano`, `normas-nombradas` y `sin-instrucciones-de-evals`, y ya no
`trescientas-lineas`; cada mutante de SC-002 trabaja sobre una copia en `t.TempDir()`. Solo lee el árbol.

## 3. `make skills-sync` no toca las referencias a mano (SC-003)

```sh
make skills-sync
git status --porcelain -- skills/
```

Esperado: `make skills-sync` sale con 0 y `git status` no da ninguna línea: en un árbol sincronizado no escribe nada, y
no retira ninguna referencia a mano (`skills/*/references/*.md` sin la cabecera «generado desde»). El caso con una
referencia generada alterada lo cubre el subtest `sincronia-con-referencias-a-mano` del escenario 2, sobre una copia.

## 4. Los controles de contenido leen la skill entera (SC-005)

```sh
go test -count=1 -run '^(TestEvalsDelRepositorio|TestOrdenesDeEjemplo|TestControlesSobreLasReferenciasAMano)$' ./internal/evals/
```

Esperado: sale con 0. `TestEvalsDelRepositorio` lleva, además de los de hoy, `lineas-de-aviso-de-las-referencias` y
`ordenes-de-ejemplo`, que ejecuta con `--offline` sobre la caché de una sesión preparada en `t.TempDir()` cada orden
concreta —`kitlegal boe`, verbo y al menos un argumento, sin marcador—, cuatro en `boe-legislacion`; la mención del
verbo sin argumentos, como la de la regla 6, no se ejecuta (`TestOrdenesDeEjemplo`).

## 5. Las dependencias de las evals cuentan la referencia a mano (SC-004, SC-010)

```sh
go test -count=1 -run '^(TestConjuntoDeCadaEval|TestEleccionDeLosMutantes)$' ./internal/evals/
```

Esperado: sale con 0: la referencia a mano en la parte (b), sin (c) ni error; el mutante que solo la cambia elige las
evals de su skill y de las que la nombran; el de una descripción cambiada elige las 45.

## 6. El juez de las descripciones, sin abrir ninguna sesión (SC-006, SC-007, SC-008, SC-014, SC-015)

```sh
go test -count=1 -run '^(TestJuezDeLasDescripciones|TestMedidaVersionada|TestEjecucionSinMedir|TestDecisionDelJob|TestBusquedaDeMedidasDeMain|TestInformeConLaDescripcion|TestDefinicionDelJob|TestGuionDeLaMedida)$' ./internal/evals/
```

Esperado: sale con 0, con votos simulados y el `claude` sustituto; ninguna orden abre una sesión con modelo (los puntos
de entrada que la abren llevan `//go:build evals` y no se compilan aquí).

## 7. Lo empotrado lleva todas las referencias (FR-015)

```sh
go test -count=1 -run '^TestSkillsEmpotradas$' .
```

Esperado: sale con 0.

## 8. Todo junto (SC-001)

```sh
make ci
```

Esperado: sale con 0. Escribe solo en los directorios de caché de Go y de las herramientas y en temporales, como
siempre.
