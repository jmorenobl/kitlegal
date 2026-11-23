# Contrato: la suite de aceptación de H45

Dos guiones testscript en `specs/042-h45-el-molde-admite/aceptacion/`, que escribe la tarea `[aceptacion]` y quedan
congelados; `aceptacion.sh activar H45` los copia a `internal/app/testdata/script/` como `h45-paquete-de-mas.txtar` y
`h45-skills-instaladas.txtar`, y desde ahí los ejecuta `TestEntregaDelHito` en `make ci`. Los dos fallan hoy por una
aserción (`rojo-primero`).

## 0. Lo que la tarea `[aceptacion]` añade al arnés (research D4)

En `internal/app/e2e_test.go`, la preparación del arnés lee una vez `go env -json GOCACHE GOMODCACHE` y el `Setup` de
`ejecutarGuiones` da esas dos variables a cada guion, como `prepararInstalacion` (`internal/skills/instalacion_test.go:
161-171`). Ninguna orden nueva del arnés ni ninguna otra variable.

## 1. `paquete-de-mas.txtar` (US4 y US2.1-US2.2 sobre datos reales · FR-001, FR-002, FR-005, FR-006, FR-016 · SC-004)

Lo que hace, en este orden:

1. `env` de `HOME=$WORK/casa` (con `mkdir casa`), `GOPATH=$WORK/go`, `GOPROXY=off`, `GOENV=off` y
   `GOFLAGS=-mod=readonly`.
2. `exec go run paquete-de-mas.go $KITLEGAL_SKILLS/.. repo`. El programa, de solo biblioteca estándar y dentro del
   archivo del guion:
   - copia la raíz del módulo bajo prueba en `repo`, sin las carpetas de primer nivel `.git`, `web`, `dist`, `bin`,
     `.codegraph`, `specs`, `.specify`, `.agents` y `.claude` y sin enlaces simbólicos, e imprime `copia: hecha`;
   - instala el paquete de data-model §6: añade las entradas de `paquete/normas.yaml`, `paquete/plazos.yaml` y
     `paquete/cobertura.yaml` (del archivo) al final del bloque de su clave de primer nivel en `repo/data/` —`normas`,
     `tipos`, `cubierto` y `regulados_aparte`—, copia `paquete/prueba.md` a
     `repo/skills/plazos-y-territorio/references/prueba.md` y el primer fichero de `repo/evals/plazos-y-territorio/`, en
     orden de nombre, a `99-<su nombre>`;
   - en `repo`, `go test -count=1 -run '^TestSkillsDelRepositorio$' ./internal/app/ -args -regenerar-skills`, e imprime
     `sincronía: <código>`;
   - en `repo`, `go test -count=1 -run '^(TestLeerCobertura|TestCoberturaDelRepositorio|TestRegionDeLoCubierto|TestRenderizarRecursosYPlazos|TestRegionDeLaCobertura|TestRegiones|TestVerticalDeEnsayo|TestDefinicionDelJob|TestPlanEnDosModos|TestUmbralesDePlazosYTerritorio|TestUmbralesDeNormativa|TestUmbralesDeLasSkills|TestInformeEnDosModos|TestMatrizDeLasPropuestas|TestSkillsDelProductoDesdeControl|TestDescripcionesDelProducto|TestDescripcionesCortas)$' ./internal/skills/ ./internal/evals/`,
     e imprime `tests enteros: <código>`;
   - en `repo`, `go test -count=1 -run '^(TestEvalsDelRepositorio|TestCasosDePlazos|TestEntregaDelHito)$/^(precepto-de-la-cobertura|ordenes-para-powershell|conjunto-plazos-y-territorio|19-eval-c-tributos|h32-paquetes|h33-cobertura-instalada|h36-plazos-de-hoy|h41-4-skills-instaladas)$' ./internal/evals/ ./internal/app/`,
     e imprime `subtests: <código>`;
   - si una orden sale con otro código que 0, escribe el principio de su salida, hasta 1200 bytes y con los que quedan
     fuera, en la salida de errores (testscript la enseña al fallar) y sigue; sale con 0 salvo que no pueda copiar o
     instalar. Entera, la salida de las tres órdenes pasa hoy de 400 kB, y con ella `rojo-primero` toma el guion por
     ilegible (gates/supuestos.md, T001).
3. Afirma, con `stdout`: `^copia: hecha$`, `^sincronía: 0$`, `^tests enteros: 0$` y `^subtests: 0$`.
4. Afirma, con `grep` sobre la copia sincronizada, que `prueba.md` lleva la fila del tipo
   `prueba-de-un-paquete-mas`, la de «procedimientos del paquete de prueba de un paquete más» y la de «procedimientos
   sancionadores en materia de tráfico y seguridad vial», y con `! grep`, que `administrativo.md` no nombra
   `BOE-A-2015-11722` ni el tipo.

Ficheros del archivo: `paquete-de-mas.go`, `paquete/normas.yaml`, `paquete/plazos.yaml`, `paquete/cobertura.yaml` (con
las claves `cubierto:` y `regulados_aparte:`) y `paquete/prueba.md` (los dos motivos y las dos regiones vacías, como
`ensayo.md` de la vertical).

Rojo hoy: con el código de `main`, la sincronía de la copia da defectos (el tipo del paquete entra también en
`administrativo.md`, y la fila de tráfico se queda en ella), `sincronía: 1`, y la aserción cae. Con el molde y sin los
tests derivados, `sincronía: 0` y caen `tests enteros` o `subtests` (P3). El mutante de SC-004, «la lista literal de
hoy devuelta a uno de los tests», pone `tests enteros` o `subtests` en 1 (mutantes.md).

Coste: dos órdenes `go test` en una copia nueva por ejecución, medidas en P3 en 7,4 s y 3,5 s de reloj; la copia, unos
150 MB (research V10). Un `go test` anidado no hereda `-race`.

## 2. `skills-instaladas.txtar` (US5.1, US9.1 · FR-009, FR-023 · SC-007)

1. `exec kitlegal skills install --dir destino --json` (como `h41-4-skills-instaladas.txtar:19`).
2. `! grep 'En PowerShell 5\.1:' destino/normativa/SKILL.md`.
3. `grep` de las cuatro líneas de §C.2 al final de `destino/normativa/references/con_herramientas.md`, tras una línea en
   blanco y ancladas con `\z` (entre comillas simples, que en testscript no expanden `$`; el `$` de `$LASTEXITCODE`,
   escapado para la expresión).
4. `grep` de la `description` de FR-023 en el frontmatter instalado de `plazos-y-territorio`, anclada entre `\A---\nname:
   plazos-y-territorio\ndescription: >-\n  ` y `\nmetadata:\n`, con `( |\n  )` entre palabras (la forma de
   `h41-4-skills-instaladas.txtar:34`, porque el YAML pliega la línea).

Rojo hoy: el paso 2 (el cuerpo lleva la frase) y el 4 (la descripción nombra la LPAC).

## 3. Lo que no es de la suite

US3, US5.2-US5.4, US6, US7 y US8 no tienen comportamiento visible del binario: los cubren `TestRegiones/defectos`,
`TestControlesSobreLasReferenciasAMano`, `TestEvalsDelRepositorio/ordenes-para-powershell`,
`TestSkillsDelProductoDesdeControl`, `TestGrabacionesSinSolape`, `TestManifiestoDeGrabaciones` y
`TestDefinicionDelJob`. US1.3 y FR-026 (`kitlegal plazos` igual) los cubren los guiones de H9 y H36 y
`TestCasosDePlazos`, que no cambian salvo `error-tipo` (contracts/tests-sin-listas.md). La activación ante las preguntas de
hoy (US9.3) es de las evals del cierre. Ninguna eval cambia (FR-025).
