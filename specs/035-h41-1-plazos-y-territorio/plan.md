# Implementation Plan: H41.1 · `plazos-y-territorio` manda leer sus referencias sin condición

**Branch**: `035-h41-1-plazos-y-territorio` | **Date**: 2026-11-10 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/035-h41-1-plazos-y-territorio/spec.md`

## Summary

`skills/plazos-y-territorio` pasa a ser `laboratorio/plazos-y-territorio/plazos-y-territorio-sin-condicion` con el
movimiento M1 y nada más (ADR 0053, punto 1): la orden de leer `municipios.md` abre el «Protocolo», la de
`calculo_de_plazos.md` cierra el paso 7, sus dos viñetas salen de «Referencias», y la frase del 4 o 5 del paso 10 pasa
al final de `calculo_de_plazos.md`. El cuerpo queda en 10 882 caracteres (10 952 con la vertical de ensayo). Dos tests
de `make ci` cambian para protegerlo: `TestSkillsDelProductoDesdeControl` calcula la skill desde esa variante con M1
(un campo `desde` en la declaración), y `TestRemisionesDelRepositorio` deja de excluirla y admite, por igualdad con una
tabla, el único fallo de su regla, la viñeta de `jerarquia_normativa.md` de v0.9.0. El guion de H41 pasa a afirmar la
forma aplicada (FR-013), y `TestArnesDeLasFamilias` deja de exigir que la sesión de `plazos-y-territorio-producto` sea
también su control, la misma igualdad de H41 que sustituye FR-005 (research D12). Un guion nuevo y congelado describe lo
que instala el binario, y `docs/SKILLS.md` §1 y `CHANGELOG.md` lo dicen. El laboratorio, el arnés, las evals y las otras
skills no cambian un byte. La aceptación es el sondeo de la persona, fuera del run. Todo verificado sobre una copia en
el directorio temporal (research V1 a V14).

## Technical Context

**Language/Version**: Go 1.27 (`go 1.27.0`, `toolchain go1.27.3` en `go.mod`; `go version`: go1.27.3 darwin/arm64).

**Primary Dependencies**: las fijadas, ninguna nueva. En los tests, `stretchr/testify` v1.12.1 y `rogpeppe/go-internal`
v1.16.0 (testscript), como hoy.

**Storage**: ficheros versionados: `skills/plazos-y-territorio/` (dos ficheros cambian), tres ficheros de test, un guion
de `internal/app/testdata/script/` que cambia y uno nuevo, `docs/SKILLS.md` y `CHANGELOG.md`.

**Testing**: `go test -race` en `make ci`; un guion testscript congelado nuevo y el de H41 con su bloque cambiado; tests
de tabla con `testify`. Ningún test con red ni con modelo. Skills aplicadas (research D14): `golang-how-to`, con
`golang-testing`, `golang-stretchr-testify`, `golang-project-layout` (ningún paquete ni fichero de Go nuevo),
`golang-naming`, `golang-lint` (ningún `//nolint`; `golangci-lint` limpio sobre la copia, V8) y `golang-cli` (ningún
argumento ni bandera).

**Target Platform**: macOS y Linux, donde corren `make ci`, `kitlegal skills install` y el sondeo de la persona.

**Project Type**: skill sin código del producto y tests del repositorio; el binario solo cambia en la skill que empotra.

**Performance Goals**: ninguno.

**Constraints**: cuerpo ≤ 11 000 caracteres, también con la línea de la vertical de ensayo; cero bytes cambiados en
`laboratorio/`, `normativa`, `jurisprudencia`, `consulta-juridica`, `evals/` y el arnés; el run no graba, no escribe en
`evidencias/` y no abre sesiones con modelo.

**Scale/Scope**: 1 skill, 2 de sus ficheros; 1 cambio y 1 movimiento declarados; 3 remisiones en la lista del plan; 1
fallo admitido; 2 guiones; 2 documentos.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principio o regla | Cómo lo cumple este hito |
|---|---|
| I · Fuentes públicas y frontera humana (INNEGOCIABLE) | No toca ninguna fuente, `internal/httpx` ni petición. El sondeo lo lanza la persona con su credencial, fuera del run, con el arnés de H41 sin cambiar |
| II · Nada sin cita ni fuente | Ningún sobre cambia. La forma fija `⚠ PLAZO:` y la regla del 4 o 5 siguen enteras, ahora en `calculo_de_plazos.md`, que la skill manda leer antes del paso 8 en toda pregunta que llega a él; la cobertura y lo no cubierto, como en v0.9.0 |
| III · Tests primero y offline | Primero el guion `aceptacion/skills-instaladas.txtar`, que hoy falla por una aserción (research V6). Los tests que cambian, offline contra los ficheros del repositorio; ninguna grabación. Ningún código de producción cambia, así que la cobertura de `internal/core` y la global no se mueven; ninguna salida de applet nueva que validar contra `schemas/` |
| IV · Arquitectura hexagonal con reglas ejecutables | Solo cambian tres ficheros `_test.go` de `internal/evals` e `internal/skills`, y datos de la skill; ningún paquete, import, error ni código de salida nuevo; `TestArquitectura` y `depguard`/`forbidigo` no cambian |
| V · Simplicidad y dependencias fijadas (YAGNI) | Ninguna dependencia. Cada mecanismo se traza a un FR («Trazabilidad»); sale `fueraDeLasRemisiones`, que quedaría vacío; el campo `desde` evita duplicar la declaración de la variante (research D1) |
| VI · Un binario, convenciones de agente | Un binario; ninguna orden, bandera ni variable nueva. `kitlegal skills install` instala la skill nueva desde lo empotrado, como hoy |
| VII · Grafo y privacidad | No toca el grafo; ningún dato de una persona |
| VIII · Skills primero; el binario es la herramienta | Mejora una skill del producto, medible con sus evals (FR-011) y con el sondeo de la persona (SC-003); el binario no cambia más que lo que empotra |
| IX · Genericidad territorial, validación local | Ninguna orden ni texto nombra un municipio: «el municipio», «los municipios que cuentan». Las evals 02 (municipio no cubierto) y 05 (Leganés) del sondeo siguen siendo la matriz cubierto/no cubierto |
| X · Genericidad por materia, validación administrativa | Nada por materia en el cuerpo: las órdenes son de capacidad (territorio, plazos); la referencia del paquete sigue en la región que genera `make skills-sync` desde `data/` |
| R1 · `internal/core/**` no importa el kernel, adaptadores ni entrada y salida | `internal/core` no cambia |
| R2 · solo `internal/httpx` importa `net/http` | Nada nuevo importa `net/http` |
| R3 · solo `internal/{cache,store,graph}` importan SQLite y `database/sql` | Nada nuevo los importa |
| R4 · solo `internal/cli` y `cmd/` llaman a `os.Exit` (`forbidigo`) | Ningún `os.Exit` nuevo |
| R5 · solo `internal/render` escribe en stdout; logs con `log/slog` a stderr (`forbidigo`) | Nada nuevo escribe en stdout; los tests informan con `testing.T` |
| R6 · `internal/graph` no importa `internal/source` ni `internal/render` | `internal/graph` no cambia |
| R7 · solo el paquete del servidor importa el SDK de MCP | Nada nuevo lo importa |
| Garantía sin red (H19) del paquete raíz y del applet `skills` | `skills.go` y el applet no cambian; solo los bytes empotrados de la skill |

Re-evaluado tras la fase 1: sin violaciones; Complexity Tracking, vacío.

## Project Structure

### Documentation (this feature)

```text
specs/035-h41-1-plazos-y-territorio/
├── plan.md              # este fichero
├── research.md          # fase 0: V1-V14 (lo verificado) y D1-D14 (las decisiones)
├── data-model.md        # fase 1
├── quickstart.md        # fase 1
├── contracts/
│   ├── skill.md         # la skill entregada, sus medidas y su uso
│   ├── controles.md     # los tests que cambian y las mutaciones de SC-001
│   └── aceptacion.md    # el guion congelado y el bloque de FR-013, byte a byte
├── aceptacion/          # lo escribe la primera tarea (skills-instaladas.txtar)
└── tasks.md             # fase 2 (/speckit-tasks)
```

### Source Code (repository root)

```text
skills/plazos-y-territorio/
├── SKILL.md                          # CAMBIA: forma sin condición y M1 (contracts/skill.md)
└── references/calculo_de_plazos.md   # CAMBIA: la frase de M1 al final
internal/evals/laboratorio_familias_test.go   # CAMBIA: TestSkillsDelProductoDesdeControl (contracts/controles.md §1)
internal/evals/laboratorio_sondeo_test.go     # CAMBIA: origenesDeLaVariante de TestArnesDeLasFamilias (contracts/controles.md §3)
internal/skills/remisiones_test.go            # CAMBIA: TestRemisionesDelRepositorio (contracts/controles.md §2)
internal/app/testdata/script/
├── h41-skills-instaladas.txtar       # CAMBIA ([datos]): FR-013 (contracts/aceptacion.md §2)
└── h41-1-skills-instaladas.txtar     # NUEVO: lo activa el workflow desde aceptacion/ (contracts/aceptacion.md §1)
docs/SKILLS.md                        # CAMBIA: §1 «Cómo se remite» (research D8)
CHANGELOG.md                          # CAMBIA: Unreleased (research D9)
```

Sin cambios: `laboratorio/`, `skills/{normativa,jurisprudencia,consulta-juridica}`, `evals/`, `data/`, `schemas/`, el
arnés (`scripts/evals-*.sh`, `internal/evals/laboratorio*.go` que no son `_test.go`, `cmd/laboratorio`), `Makefile`,
`.golangci.yml`, `CONTRIBUTING.md` (research D10).

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`: el producto en `skills/`, sus controles junto a los de
H41 en `internal/evals` e `internal/skills`, los guiones e2e en `internal/app/testdata/script/`. Ningún paquete ni
fichero de Go nuevo.

## Diseño

- **La skill** (contracts/skill.md; research D2, D3): `-sin-condicion` con dos líneas del paso 10 cambiadas por una y
  la frase, en una línea, detrás de una línea en blanco al final de `calculo_de_plazos.md`. Medido: 10 882 caracteres
  (V2).
- **`TestSkillsDelProductoDesdeControl`** (contracts/controles.md §1; research D1, D5): `declaracionDelProducto.desde`
  elige la base; la de `plazos-y-territorio` es `-sin-condicion`, con un cambio y un movimiento; sale el subtest «hasta
  H41.1»; el mensaje nombra la base.
- **`TestRemisionesDelRepositorio`** (contracts/controles.md §2; research D4): sin exclusiones; `fallosAdmitidosDelProducto`
  y `assert.Equal` por skill; la lista del plan de `plazos-y-territorio`, tres frases.
- **`TestArnesDeLasFamilias`** (contracts/controles.md §3; research D12, V13): `origenesDeLaVariante` deja de añadir
  `plazos-y-territorio-control` a los orígenes de `plazos-y-territorio-producto`; la sesión se sigue comparando, byte a
  byte, con `skills/plazos-y-territorio`.
- **Guiones** (contracts/aceptacion.md; research D6, D7): el nuevo, sobre lo instalado; el de H41, su bloque de
  `plazos-y-territorio` y dos tramos de su cabecera.
- **Documentación** (research D8, D9): la línea de FR-009 y las dos frases «hasta H41.1» de `docs/SKILLS.md` §1; la
  entrada de `CHANGELOG.md`.

**Aceptación e2e:** un guion testscript que escribe la primera tarea en
`specs/035-h41-1-plazos-y-territorio/aceptacion/skills-instaladas.txtar` y queda congelado (contracts/aceptacion.md §1),
activado como `internal/app/testdata/script/h41-1-skills-instaladas.txtar` y ejecutado por `TestEntregaDelHito`: lo que
pone `kitlegal skills install --dir` lleva la orden de `municipios.md` en un párrafo propio al principio del
«Protocolo» y la de `calculo_de_plazos.md` al final del paso 7, una vez cada enlace, la lista de «Referencias» sin sus
dos viñetas y terminada en la de `jerarquia_normativa.md`, «pregúntalo y espera» en el paso 1, el paso 10 terminado en
``⚠ REDACCIÓN A FECHA:`.`` y la frase del 4 o 5 al final de `calculo_de_plazos.md`. Cubre las historias 1 y 2: FR-001,
FR-002, FR-003; SC-001 (el guion en verde). La historia 3 es de `make ci` y no tiene guion propio: la cubren los tests
de contracts/controles.md, sus mutaciones (research V7) y el guion de H41 con FR-013. La 4 la hace la persona fuera del
run (SC-003), y la 5 es documentación sin salida del binario (FR-009, FR-010).

## Orden de implementación

De dentro afuera, con `make ci` en verde tras cada tarea:

1. `[aceptacion]`: el guion de contracts/aceptacion.md §1, que falla hoy por una aserción (research V6).
2. `[datos]`, en una sola tarea porque, con la skill cambiada, cada una de las otras partes que falte deja `make ci` en
   rojo (research D7, V5, V13): los dos ficheros de la skill, la declaración y el campo `desde` de
   `TestSkillsDelProductoDesdeControl` (sin el subtest «hasta H41.1»), `origenesDeLaVariante` de
   `TestArnesDeLasFamilias` y el bloque de FR-013 de `h41-skills-instaladas.txtar`.
3. `TestRemisionesDelRepositorio` (necesita la skill ya cambiada: con la de hoy, la lista del plan no se cumple).
4. `docs/SKILLS.md` y `CHANGELOG.md`.

## Trazabilidad

| Mecanismo | FR/SC |
|---|---|
| Tramos de §2.1 de H41 en el `SKILL.md` | FR-001, FR-002 |
| M1: el cambio del paso 10 y la frase al final de `calculo_de_plazos.md` | FR-003, FR-004 |
| Campo `desde` y declaración de `plazos-y-territorio`; salida del subtest «hasta H41.1»; mensaje con la base | FR-005, SC-001, SC-002 |
| `fallosAdmitidosDelProducto`, `assert.Equal` por skill y salida de `fueraDeLasRemisiones` | FR-007, SC-001 |
| Lista del plan de `plazos-y-territorio` en `remisionesDelPlan` | FR-007, SC-001 |
| `origenesDeLaVariante` sin el segundo origen de `plazos-y-territorio-producto` | FR-005 (la igualdad con el control deja de ser cierta), FR-008 (el arnés no cambia), SC-001 |
| `aceptacion/skills-instaladas.txtar` | FR-001, FR-002, FR-003, SC-001 (III; ADR 0018) |
| Bloque y cabecera de `h41-skills-instaladas.txtar` | FR-013, SC-001 |
| Línea de FR-009 y las dos frases «hasta H41.1» de `docs/SKILLS.md` | FR-009 |
| Entrada de `CHANGELOG.md` | FR-010 |

## Uso, de fuera adentro

- **La skill** (contracts/skill.md §6): el agente lee el cuerpo una vez por pregunta (12 113 bytes; 10 882 caracteres de
  cuerpo), `municipios.md` una vez por pregunta a la que llega (3 944 bytes) y `calculo_de_plazos.md` una vez por
  pregunta que llega al paso 7 (2 869). Una de territorio, unos 16 KB; una de plazos, con la referencia del paquete,
  unos 29 KB. Fijo con meses de uso: cientos de normas y miles de bloques consultados viven en la caché y el grafo, no en
  la skill; el cuerpo solo crece con los paquetes de `data/`, unos 70 caracteres cada uno, bajo el tope. Las
  invocaciones de `kitlegal` por pregunta, las de v0.9.0. Ninguna señal en el producto.
- **Los defectos de los controles** (contracts/controles.md §5): los lee quien cambia la skill o el laboratorio, una
  vez por `make ci` o `make skills-check`, una línea por fichero o remisión que difiere, con su ruta; dejan de darse
  cuando lo cambiado vuelve a ser lo declarado. No dependen del uso.
- **La tabla del sondeo** (US4): la lee la persona una vez, antes de fusionar: cuatro filas y el detalle de 160
  sesiones. En la de H41 (`evidencias/sondeos-20261109/tabla-plazos-y-territorio.md`: 54 478 bytes, 10 filas y 400
  sesiones), una fila mide 316 bytes y una línea del detalle, unos 128: la de H41.1, unos 4 × 316 + 160 × 128 ≈ 22 KB,
  estimado y no medido. No se repite en el producto ni crece con el uso.

## Gates: controles mecánicos que añade o toca

- **Objetivos de `make`**: ninguno nuevo ni cambiado. `make skills-check` ya corre `TestRemisionesDelRepositorio`, que
  ahora mira también `plazos-y-territorio`.
- **Tests que cambian**: `TestSkillsDelProductoDesdeControl` (`internal/evals/laboratorio_familias_test.go`),
  `TestRemisionesDelRepositorio` (`internal/skills/remisiones_test.go`) y, en `TestArnesDeLasFamilias`, solo
  `origenesDeLaVariante` (`internal/evals/laboratorio_sondeo_test.go`), con contracts/controles.md §1 a §3.
- **Sin cambiar ningún valor esperado** (research V5, V13): `TestVariantesDeControl`,
  `TestVariantesDePlazosYTerritorio`, `TestSkillsDelRepositorio`, `TestVerticalDeEnsayo`, `TestRecetasDelSondeo`,
  `TestLaboratorioFueraDelProducto`, `TestFrases`, `TestArnesDeNormativa`, `TestArnesDelLaboratorio`,
  `TestVariantesDeNormativa` y los de la tabla.
- **Guiones e2e**: el nuevo y congelado `h41-1-skills-instaladas.txtar`; el de H41, `h41-skills-instaladas.txtar`,
  cambiado por una tarea `[datos]` (fichero existente de `testdata/`: el informe final lo lleva a la persona, capa 3).
- **Fixtures**: ninguno nuevo; ninguna grabación.
- **Configuración de verificación**: `.golangci.yml`, `Makefile` y `.github/` no cambian.

## Datos externos

Ninguno (research, «Fuentes externas»): no hay manifiesto `grabaciones.json`, test `TestGrabar*` ni tarea que grabe; el
run no escribe en `evidencias/` ni en `data/`. La tabla del sondeo la deja la persona en `evidencias/`, fuera del run.

## Controles de umbral

| Requisito | Umbral | Control | Dónde |
|---|---|---|---|
| FR-004, SC-002 | 0 skills de `skills/` con un cuerpo de más de 11 000 caracteres | El defecto `<skill>: SKILL.md: cuerpo de N caracteres (máximo 11000)` | ci:internal/app/skills_test.go:TestSkillsDelRepositorio |
| FR-004, SC-002 | 0 cuerpos de más de 11 000 caracteres con la línea de la vertical de ensayo de H32 (`plazos-y-territorio`: 10 952) | La sincronía de la copia con la vertical, sin defectos; mutante de 49 caracteres en research V7 | ci:internal/skills/vertical_de_ensayo_test.go:TestVerticalDeEnsayo |
| FR-004, SC-002, FR-005 | el cuerpo de `plazos-y-territorio` mide 10 882: 0 caracteres de diferencia con `-sin-condicion` con M1 | El cálculo desde la variante, byte a byte | ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl |
| FR-001, FR-002, FR-003, FR-005, SC-001 | 0 bytes de diferencia fuera de lo generado con `-sin-condicion` + M1; 0 ficheros de más o de menos; la frase de M1 0 veces en el cuerpo y al final de la referencia; 0 bytes de diferencia en el frontmatter con control | «desde control», «movimientos» y «frontmatter» (contracts/controles.md §1.2) | ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl |
| FR-006, SC-005 | 0 bytes cambiados en los cinco ficheros de `plazos-y-territorio-control` | Las huellas SHA-256 por ruta | ci:internal/evals/laboratorio_familias_test.go:TestVariantesDeControl |
| FR-006, SC-005 | 0 bytes fuera de lo declarado en las tres formas de H41 | Cada forma, su declaración sobre control | ci:internal/evals/laboratorio_familias_test.go:TestVariantesDePlazosYTerritorio |
| FR-007, SC-001 | los fallos de la regla son exactamente los admitidos: 1 en `plazos-y-territorio`, 0 en las demás; 0 frases distintas de la lista del plan | `assert.Equal` con `fallosAdmitidosDelProducto`; la lista, frase a frase | ci:internal/skills/remisiones_test.go:TestRemisionesDelRepositorio |
| FR-012, SC-005 | 0 bytes fuera de lo declarado en H41 en `normativa` y `jurisprudencia` | Sus declaraciones, sin cambiar; la rama con `desde` vacío, que el hito reescribe, en rojo con un byte de `con_herramientas.md` y otro de `documento_aportado.md` (contracts/controles.md §5) | ci:internal/evals/laboratorio_familias_test.go:TestSkillsDelProductoDesdeControl |
| FR-008 | 0 cambios en las recetas de `make evals-laboratorio` y `make laboratorio-tabla` | Las recetas, tal cual | ci:internal/evals/laboratorio_sondeo_test.go:TestRecetasDelSondeo |
| FR-008 | 0 bytes de diferencia entre la skill de cada sesión de `plazos-y-territorio-producto` y `skills/plazos-y-territorio`; las sesiones que abre cada invocación, las de hoy | El arnés con el sustituto de `claude` | ci:internal/evals/laboratorio_sondeo_test.go:TestArnesDeLasFamilias |
| FR-001, FR-002, FR-003, SC-001 | 100 % de los `grep` del guion de H41.1 sobre lo instalado | El guion congelado `h41-1-skills-instaladas.txtar` | ci:internal/app/e2e_test.go:TestEntregaDelHito |
| FR-013, SC-001 | 100 % de los `grep` del guion de H41, con el bloque de `plazos-y-territorio` de la forma aplicada | `h41-skills-instaladas.txtar`; cada uno de sus seis `grep` nuevos (líneas 54 a 59), en rojo con una mutación de la skill y el guion intacto (contracts/aceptacion.md §2) | ci:internal/app/e2e_test.go:TestEntregaDelHito |
| FR-011, SC-004 | Las evals de `plazos-y-territorio` en verde en el cierre, en los dos modos | El job de evals del cierre | evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:orden, evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:herramienta |
| FR-011, SC-004 | Las evals de `normativa` en verde en el cierre, en los dos modos | El job de evals del cierre | evals:normativa:sin_activar:claude-sonnet-5-5:orden, evals:normativa:sin_activar:claude-sonnet-5-5:herramienta, evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:orden, evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:herramienta, evals:normativa:afirma_lo_no_leido:claude-sonnet-5-5, evals:normativa:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar, evals:normativa:medida_del_juez:afirma_lo_no_leido:correctos_marcados, evals:normativa:duracion_de_las_sesiones:orden, evals:normativa:duracion_de_las_sesiones:herramienta, evals:normativa:duracion_del_juez:orden, evals:normativa:duracion_del_juez:herramienta |
| FR-011, SC-004 | Las evals de `jurisprudencia` en verde en el cierre, en los dos modos | El job de evals del cierre | evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden, evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta, evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden, evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta, evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden, evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta, evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5, evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar, evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:correctos_marcados, evals:jurisprudencia:duracion_del_juez:orden, evals:jurisprudencia:duracion_del_juez:herramienta |

FR-011 y SC-004 usan los umbrales del job tal como están (`.github/workflows/evals.yml` y las carpetas de evals), que el
hito no toca: las filas nombran los que publican con `decide: true` los informes del cierre de H41 (research V11), y el
veredicto del job sale en rojo con cualquiera de ellos. `consulta-juridica` no publica ninguno con `decide: true` y su
skill no cambia; su veredicto lo deciden sus series (ADR 0016), configuración que el hito no toca. En SC-005, los cero
bytes de `consulta-juridica`, `evals/` y el arnés no tienen control en `make ci` ni el spec lo nombra: los guarda el
guardián de diff del workflow (`scripts/workflow/guardian-diff.sh`, las rutas que declara cada tarea, y ninguna tarea
declara esas), configuración que el hito no toca. SC-003 lo mide la persona con el sondeo (capa 3), fuera de `make ci` y
del cierre.

## Cierre y pendientes

- **SC-003**: el run no lanza el sondeo; la propuesta de cambio queda pendiente de su tabla y su lectura, y no se fusiona
  sin ellas; si no cumple, lo medido va a un ADR que dice qué medir después (quickstart.md, escenario 8).
- **FR-011**: si la medición del cierre da un rojo, la reparación corrige la skill dentro de lo que declara
  contracts/skill.md; ningún umbral se rebaja.
- **Para la persona, en el informe final**: el cambio de `internal/app/testdata/script/h41-skills-instaladas.txtar`
  (FR-013; capa 3); el de `origenesDeLaVariante`, que FR-008 no nombra y sin el que `make ci` sale en rojo (research
  D12; `gates/supuestos.md`); `CONTRIBUTING.md` líneas 1813-1814 y 1828-1830, que siguen describiendo `skills/plazos-y-territorio`
  como su control (research D10); y el ejemplo de remisión de `docs/SKILLS.md` §1 (línea 90), que el spec deja fuera.

## Complexity Tracking

Ninguna desviación: ni dependencias fuera de la lista, ni violaciones de la constitución que justificar.
