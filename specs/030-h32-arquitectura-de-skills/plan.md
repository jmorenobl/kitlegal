# Implementation Plan: H32 · Arquitectura de skills: una entrada genérica, skills por capacidad y las materias como referencias

**Branch**: `030-h32-arquitectura-de-skills` | **Date**: 2026-10-29 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/030-h32-arquitectura-de-skills/spec.md`

## Summary

Las tres skills de hoy pasan a cuatro: una de entrada, **`consulta-juridica`**, que clasifica la pregunta, dice qué
capacidad la resuelve y qué cubre la base y declara en prosa lo que no cubre; y tres de capacidad, **`normativa`** (de
`boe-legislacion`: `boe`, `graph`), **`jurisprudencia`** (la misma: `cita`) y **`plazos-y-territorio`** (de
`legal-core`: `plazos`, `territorio`), cada una con su protocolo, sus invariantes y sus formas fijas, sin remitir a
otra. Lo administrativo de `legal-core` sale a un **paquete de materia**: diez normas marcadas con
`vertical: administrativo` en `data/normas.yaml` y una referencia, `plazos-y-territorio/references/administrativo.md`,
escrita a mano con la prosa del procedimiento común y las tres tablas que hoy se generan desde `data/`, ahora generadas
entre marcas dentro de ella; la cobertura da la línea `⚠ NO CUBIERTO:` entera de cada fila. `make skills-sync`
enlaza cada referencia de un paquete desde su skill entre marcas, y una vertical de ensayo en `testdata/` prueba que
una materia entra sin Go ni ediciones de un `SKILL.md`. Un bloque se cita igual en todas las skills,
`<forma legible> [<identificador>, bloque <id>]`, con dos controles nuevos en `make ci` (la forma única y las formas
compartidas) y la medida de `cobertura_sin_declarar` exigiendo la cita del precepto. `normativa` recupera en el paso
de la comprobación que un bloque sin `version-obsoleta` no da nada a la respuesta, y su paso 5 deja de permitir la
materia pegada a una remisión. `kitlegal skills install` retira `boe-legislacion` y `legal-core` de un ámbito cuando
sus huellas son las del manifiesto, y avisa si no. Las evals pasan con `git mv` a la skill que hereda su pregunta, la
entrada trae dos (la póliza, con la Ley 50/1980 grabada en `grabar_datos`, y una de una sola capacidad), y la matriz
del job pasa a cuatro trabajos en once sesiones a la vez, con el tope de hoy.

## Technical Context

**Language/Version**: Go 1.27 (`go 1.27.0`, `toolchain go1.27.3`); Markdown (las skills, la guía, el ADR, el
CHANGELOG); YAML (las evals, `data/`).

**Primary Dependencies**: ninguna nueva. `alecthomas/kong` (el applet `skills`), `gopkg.in/yaml.v3` y
`santhosh-tekuri/jsonschema` (los datos y las evals), `invopop/jsonschema` (`schemas/instalacion.json` desde
`--describe`), `rogpeppe/go-internal` (`testscript`) y `stretchr/testify`, ya en `go.mod`.

**Storage**: ficheros del repositorio (`skills/`, `data/`, `evals/`, `schemas/`) y, en el equipo de quien instala, el
manifiesto `kitlegal.json` de cada ámbito (H19), que ahora también se encoge.

**Testing**: `make ci`, sin red ni modelo: unitarios en `internal/skills`, `internal/core/instalacion`, `internal/app` e
`internal/evals`; la suite de aceptación (`testscript`, `TestEntregaDelHito`); y una grabación nueva de
`boe.legislacion-consolidada` que hace el paso `grabar_datos`.

**Target Platform**: macOS y Linux para `make ci`; `ubuntu-24.04` para el job de evals.

**Project Type**: skills agénticas sobre el CLI multicall (el binario cambia en `skills install`, en `internal/skills`
y en `internal/evals`).

**Performance Goals**: cada cuerpo ≤ 11 000 caracteres con lo generado, y con sitio para la línea de un paquete más
([contracts/textos.md](contracts/textos.md)); el peor caso de cada trabajo de evals ≤ su tope (352 min).

**Constraints**: el run no usa red ni modelo salvo `grabar_datos` (ADR 0032); ningún umbral cambia; ningún applet de
capacidad ni `internal/source` cambian; `scripts/workflow/`, `.specify/` e `internal/workflow/` no se tocan.

**Scale/Scope**: cuatro skills; 45 evals movidas y 2 nuevas; un paquete; diez normas marcadas y una nueva; dos nombres
retirados; seis guiones de aceptación; un ADR.

## Constitution Check

*GATE: pasado antes de la fase 0 y re-evaluado tras el diseño (fase 1): sin violaciones.*

### Principios

| Principio | Cómo lo cumple H32 |
|---|---|
| I · Fuentes públicas y frontera humana | La única fuente que se graba, `boe.legislacion-consolidada`, es pública y tiene fila en `docs/SOURCES.md` de `main` con «Revisado» 2026-09-13; la graba `grabar_datos` con `TestGrabarEvals`, que pide por `internal/httpx` (GET, `robots.txt`, User-Agent, ritmo `1s`). Ningún código HTTP cambia. La instalación solo escribe y borra en los ámbitos de quien la ejecuta, y solo lo que su manifiesto dice que instaló kitlegal; lo que no, lo deja y avisa: retirarlo «es cosa tuya». Cada skill de capacidad conserva «no actuar en nombre de nadie» |
| II · Nada sin cita ni fuente | Ningún sobre cambia. Un bloque se cita en una sola forma en todas las skills, con su control (FR-010); la línea `⚠ NO CUBIERTO:` sale de la referencia del paquete con su precepto citado y la norma que lo rige, y su medida la exige (FR-012, FR-014). Cada skill de capacidad lleva las invariantes de `CLAUDE.md`, con su frase ([contracts/reparto.md](contracts/reparto.md) §4). La entrada declara en prosa lo que la base no cubre, sin plazos, órganos ni vías de memoria (FR-001, FR-002), y lo que la ley regula aparte sigue declarado no cubierto con la norma que lo rige |
| III · Tests primero y offline | La primera tarea escribe los seis guiones de aceptación, que fallan hoy por una aserción y quedan congelados ([contracts/aceptacion.md](contracts/aceptacion.md)). Cada control nuevo trae su mutante en el mismo test. La grabación de la Ley 50/1980 la hace `grabar_datos` tras la tarea `[datos]` del manifiesto; los tests la leen sin red. Las salidas de `install` se validan contra `schemas/instalacion.json`, regenerado desde `--describe`. Cobertura: los paquetes de código que cambian (`internal/skills`, `internal/core/instalacion`, `internal/evals`) llevan sus tests; `internal/core/**` sigue ≥ 85 % |
| IV · Arquitectura hexagonal con reglas ejecutables | `internal/core/instalacion` recibe la lista de retiradas como argumento (`instalacion.Retiradas`) y lee y borra por sus puertos de hoy, `Disco` y `Escritor` (`internal/core/instalacion/puertos.go`), que implementa `internal/disco` y que inyecta `internal/app`: no importa adaptadores ni lleva nombres de skill. `install` sale con 0 tras retirar o avisar; con un nombre retirado pedido, 2 (`argumentos`, como hoy); un fallo de disco, el de hoy (7 o 1). Ningún `panic` |
| V · Simplicidad y dependencias fijadas (YAGNI) | Ninguna dependencia. Ningún paquete nuevo; los tipos y funciones nuevos se trazan a su FR («Trazabilidad»). El molde no añade ningún fichero de declaración: un paquete es su `vertical`, sus referencias, sus filas y sus evals (research D3). Lo que llega manipulado de fuera es la regla genérica (`inesperado`, 1), sin casos |
| VI · Un binario, convenciones de agente | Un solo binario; `skills install` conserva sus banderas (`-g`, `--dir`, `--host`, `--dry-run`, `--json`); `--describe` da el esquema de `data` con los dos estados y el `aviso` nuevos; la tabla de comandos de cada skill sigue generada desde `--describe` |
| VII · Grafo y privacidad | Nada del grafo cambia. La comprobación `kitlegal graph check <norma> <bloque>` sigue detrás de cada lectura de `normativa`; el hito solo quita de la respuesta lo que no da |
| VIII · Skills primero; el binario es la herramienta | El hito entrega una skill nueva y reparte las tres de hoy por capacidad, medidas por sus evals con los umbrales de hoy (SC-010 a SC-015) y por el sondeo de una persona (SC-016). Lo que va a Go es lo que exige determinismo: la sincronía de las regiones, los dos controles, la medida de la cobertura y la retirada de lo instalado. Las referencias siguen la regla de `references/` plano (ADR 0046) y viajan en el binario |
| IX · Genericidad territorial, validación local | Ningún municipio ni comunidad nuevos; `plazos-y-territorio` conserva los pasos y las reglas del territorio de `legal-core` y su referencia de municipios; las normas locales y la variación autonómica, en cada skill de capacidad |
| X · Genericidad por materia, validación administrativa | Es el objetivo del hito: ningún cuerpo de capacidad nombra una norma con `vertical` (FR-021); lo administrativo es un paquete de datos y una referencia; una materia nueva entra sin Go ni ediciones de un `SKILL.md`, probado con la vertical de ensayo (FR-024); la línea `⚠ NO CUBIERTO:` de cada fila sale de la forma del cuerpo, no de una plantilla de materia en Go (research D6). Lo que sigue administrativo por construcción —los tipos de plazo, la cobertura en negativo— es lo que el ADR 0044 deja a H33 y H36 |

### Reglas de dependencia (constitución IV; `docs/ROADMAP.md` §2; `internal/arch_test.go`)

| Regla | Cómo la cumple |
|---|---|
| R1 · `internal/core/**` no importa el kernel, los adaptadores ni entrada y salida | `internal/core/instalacion` gana `Retiradas` y la planificación de la retirada sobre sus puertos `Disco` y `Escritor`; el adaptador `internal/disco` lo inyecta `internal/app`, como hoy |
| R2 · Solo `internal/httpx` importa `net/http` | Ninguna importación nueva; la grabación usa `TestGrabarEvals`, que pide por `internal/httpx` |
| R3 · Solo `internal/{cache,store,graph}` importan SQLite y `database/sql` | Ninguna importación nueva; `TestCacheDeLasSesiones` prepara la caché con `PrepararSesion`, como hoy |
| R4 · Solo `internal/cli` y `cmd/` llaman a `os.Exit` | Ningún `os.Exit`: `install` devuelve su `Resultado` |
| R5 · Solo `internal/render` escribe en stdout; logs con `log/slog` a stderr | La salida legible de la retirada va en `Resultado.Legible` (ADR 0026) y las líneas de `--dry-run` en el `Ensayo`, que escribe el kernel; ningún `fmt.Print*` |
| R6 · `internal/graph` no importa `internal/source` ni `internal/render` | Sin cambios en `internal/graph` |
| R7 · Solo el paquete del servidor importa el SDK de MCP | Ninguna importación del SDK |
| Sin red (`arch_test.go`) | Ningún paquete de producción cambia de importaciones de red |

## Project Structure

### Documentation (this feature)

```text
specs/030-h32-arquitectura-de-skills/
├── plan.md              # este fichero
├── research.md          # decisiones D1–D21, comprobaciones V1–V16, supuestos S1–S5
├── data-model.md        # entidades: skills, norma con vertical, paquete, regiones, retirada, evals
├── quickstart.md        # escenarios de comprobación, sin efectos en el árbol
├── contracts/
│   ├── reparto.md       # las skills, sus nombres, lo heredado, invariantes, formas, tabla por eval
│   ├── paquetes.md      # el molde: vertical, referencias, regiones, enlaces, vertical de ensayo
│   ├── forma-de-cita.md # FR-010, FR-013 y la medida de FR-014
│   ├── instalacion-retiradas.md
│   ├── evals.md         # carpetas, evals nuevas, reglas, elección, matriz y peor caso, umbrales, caché
│   ├── aceptacion.md    # los seis guiones y la búsqueda de FR-032
│   ├── textos.md        # pasajes de c991808 y de la entrega, con sus medidas
│   └── borradores/      # los cuatro SKILL.md de la entrega y administrativo.md, medidos
├── aceptacion/          # la suite (T001)
└── tasks.md             # /speckit-tasks
```

### Source Code (repository root)

```text
skills/
├── consulta-juridica/SKILL.md                     # nueva
├── normativa/                                     # git mv de boe-legislacion/; SKILL.md y como_se_cita.md cambian
├── jurisprudencia/SKILL.md                        # cambia (J1–J3)
└── plazos-y-territorio/                           # git mv de legal-core/
    ├── SKILL.md                                   # cambia (P1–P17)
    └── references/
        ├── administrativo.md                      # nueva: la referencia del paquete
        ├── calculo_de_plazos.md                   # sin lo del procedimiento común
        ├── jerarquia_normativa.md, municipios.md  # sin cambios
        └── (salen cobertura_por_materia.md, recursos_y_plazos.md, leyes_vertebrales.md)
data/normas.yaml                                   # vertical en diez normas; la Ley 50/1980
schemas/normas.yaml.json                           # vertical
schemas/eval.yaml.json                             # se_activan; no_cubierto.precepto
schemas/instalacion.json                           # estados retirada y sin retirar; aviso
evals/
├── consulta-juridica/01-…, 02-…                   # nuevas
├── normativa/                                     # git mv de boe-legislacion/
├── jurisprudencia/                                # sin cambios
└── plazos-y-territorio/                           # git mv de legal-core/; 01–04, 07, 08, 12 cambian
internal/skills/                                   # regiones en referencias a mano, enlaces de paquetes, vertical, cobertura filtrada
internal/skills/testdata/vertical-de-ensayo/       # nueva
internal/core/instalacion/                         # Retiradas, retirada y aviso
internal/app/                                      # retiradas.go, install, salida legible, skills_test.go (dos controles)
internal/app/testdata/script/                      # h32-*.txtar (activación); salen h31-1-h31-1-*.txtar; h19-*, nombres nuevos
internal/evals/                                    # se_activan, cobertura con cita, reglas, dependencias, tests con nombres nuevos
internal/evals/testdata/grabaciones.json           # la entrada de la Ley 50/1980
.github/workflows/evals.yml                        # matriz y nombres
.github/workflows/release.yml                      # el paso de humo de skills install, con normativa
skills_test.go                                     # skillsDeHoy, con las cuatro skills de la entrega
release_test.go                                    # comprobacionesDelHumo: la orden de skills-install, con normativa
README.md, web/src/pages/instalar/index.astro      # nombres nuevos (el texto alternativo de la captura, no)
docs/SKILLS.md, docs/ADR/0048-…, CHANGELOG.md
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`: las skills en `skills/`, sus evals en `evals/`, la
verdad en `data/`, el dominio de la instalación en `internal/core/instalacion` y su composición en `internal/app`, la
sincronía en `internal/skills`, el arnés de evals en `internal/evals`. Ningún paquete nuevo.

## Aceptación e2e

**Aceptación e2e:** seis guiones `testscript` que la primera tarea escribe en
`specs/030-h32-arquitectura-de-skills/aceptacion/` y que se activan como `internal/app/testdata/script/h32-*.txtar`
([contracts/aceptacion.md](contracts/aceptacion.md)):

| Guion | Historia | Cubre |
|---|---|---|
| `entrada.txtar` | US1 | FR-001, FR-004, FR-005, FR-007 · SC-017 |
| `capacidades.txtar` | US2 | FR-006 · SC-017 |
| `normativa.txtar` | US3 | FR-015, FR-016, FR-017, FR-018, FR-021, FR-040 · SC-017 |
| `paquetes.txtar` | US4 | FR-019, FR-021, FR-023 · SC-017 |
| `retiradas.txtar` | US5 | FR-026 a FR-029, FR-031 · SC-006 |
| `nombres-retirados.txtar` | US7 | FR-032 · SC-018 |

US6 no tiene guion: su aceptación son `TestDefinicionDelJob`, `TestEvalsDelRepositorio` y los tests de los umbrales en
`make ci`, y el job de cierre (SC-008 a SC-015). US8, el sondeo, lo mide una persona (SC-016). Los tres guiones
`h31-1-h31-1-*.txtar` se retiran en la misma entrega (FR-040).

## Controles mecánicos que este hito añade o toca

| Control | Qué | Dónde |
|---|---|---|
| `forma-de-cita` (nuevo) | ninguna cita de un bloque fuera de la forma única, con dos mutantes | `TestSkillsDelRepositorio`, `internal/app/skills_test.go` (en `make skills-check`) |
| `formas-compartidas` (nuevo) | las líneas de una misma etiqueta `⚠`, iguales en cada `SKILL.md`, con su mutante | ídem |
| regiones en referencias a mano (nuevo en la sincronía) | la deriva de `cobertura_por_materia`, `recursos_y_plazos` y `leyes_vertebrales` dentro de `administrativo.md`; la línea de cada fila igual a la forma del cuerpo sustituida, con el mutante de la forma cambiada sin regenerar | `internal/skills/sincronia.go`, `referencias.go`; `TestSkillsDelRepositorio` y `internal/skills/referencias_test.go` |
| enlaces de paquetes (nuevo en la sincronía) | las líneas entre `<!-- paquetes … -->` y las de la entrada; la referencia de un paquete sin marcas donde enlazarla | ídem |
| `TestVerticalDeEnsayo` (nuevo) | la vertical de `testdata/` entra sin `.go` ni cambios fuera de las marcas, con su mutante | `internal/skills/vertical_de_ensayo_test.go`; fixtures en `internal/skills/testdata/vertical-de-ensayo/` |
| `TestSkillsRetiradas` (nuevo) | la lista de nombres retirados frente a lo empotrado y a `c991808` | `internal/app/retiradas_test.go` |
| `TestSkillsDelRepositorio`, `normas-nombradas`, presupuestos, frontmatter (tocados) | `skillsExigidas` con los cuatro nombres; las mismas reglas sobre cuatro skills | `internal/app/skills_test.go` |
| controles de contenido de `TestEvalsDelRepositorio` (tocados) | los de `boe-legislacion`, sobre `normativa`; los de `legal-core`, sobre `plazos-y-territorio` (SC-001) | `internal/evals/conjunto_test.go` |
| reglas de las evals (tocadas y nuevas) | `ReglasDeNormativa`, `ReglasDePlazosYTerritorio`, `ReglasDeConsultaJuridica`; cada skill de capacidad en un `no_se_activan` de otra carpeta; los nombres de `se_activan` y `no_se_activan`, de skills que existen; `no_cubierto.precepto`, de una fila | `internal/evals/conjunto.go`, `conjunto_test.go` |
| `TestJuzgar` (tocado) | `se_activan`: la sesión que no activa una de sus skills no pasa, con el motivo que la nombra, y la que las activa todas pasa, junto a los casos de `no_se_activan` | `internal/evals/juzgar_test.go` |
| `TestJuzgarPlazos` (tocado) | la medida de `cobertura_sin_declarar` con la cita del precepto en la misma línea | `internal/evals/plazos_test.go` |
| `TestDefinicionDelJob` (tocado) | la matriz de cuatro skills, once sesiones, el peor caso recalculado | `internal/evals/definicion_test.go` |
| `TestEleccionDeLosMutantes` (tocado) | una referencia de un paquete y `data/cobertura.yaml` eligen las evals de la skill que la enlaza | `internal/evals/eleccion_test.go` |
| umbrales del informe (tocados) | `TestUmbralesDeLasSkills` con cuatro skills y los valores de hoy: `consulta-juridica`, de sus respuestas, solo con `fuera_de_lo_grabado` por modelo y modo, que no decide, sin `sin_activar` ni `lectura_sin_intentar` ([contracts/evals.md](contracts/evals.md) §7); `TestLimiteDeLaTasa` con `normativa` | `internal/evals/informe_test.go`, `tasa_test.go` |
| `TestCacheDeLasSesiones` (tocado) | los bloques de la eval de la póliza; las carpetas nuevas | `internal/evals/preparar_test.go` |
| `TestIdentificadoresDeLasNormas` (tocado) | la Ley 50/1980 contra el manifiesto de H9 | `internal/evals/grabaciones_test.go` |
| `make schema-check` (tocado) | `schemas/instalacion.json` regenerado | `esquemas_test.go` |
| `TestSkillsEmpotradas` (tocado) | `skillsDeHoy` pasa a `consulta-juridica`, `jurisprudencia`, `normativa` y `plazos-y-territorio`: con los nombres de `c991808`, `require.Subset` sale en rojo en cuanto el árbol los pierde (FR-005, FR-007) | `skills_test.go` de la raíz |
| `TestConfiguracionDeLaRelease`, subtest `humo` (tocado) | la orden `if [ ! -f .agents/skills/normativa/SKILL.md ]; then` de la comprobación `skills-install` de `comprobacionesDelHumo`, copia de la del paso de `.github/workflows/release.yml` (su nombre, el `if` y el `echo`), que cambia a la vez por FR-032: `normativa` hereda de `boe-legislacion` lo que el paso comprobaba | `release_test.go` de la raíz |
| `TestEntregaDelHito` (tocado) | los seis `h32-*`; los `h19-*` y `skills-*` con los nombres nuevos; salen los `h31-1-*` | `internal/app/e2e_test.go` |
| fixtures | nuevos: la grabación de la Ley 50/1980 (`internal/evals/testdata/boe.legislacion-consolidada/`), la vertical de ensayo; modificados: los guiones de `internal/app/testdata/script/` y `internal/skills/testdata/script/` que nombran las skills de hoy | tareas `[datos]` |

## Controles de umbral

| Requisito | Umbral | Control | Dónde |
|---|---|---|---|
| SC-001 | `make ci` termina con 0, con todos los controles de §1 de `docs/SKILLS.md` sobre las skills nuevas | todos los de esta tabla que viven en `make ci` y los de siempre | `ci:Makefile:ci` |
| FR-009, SC-002 | cada cuerpo ≤ 11 000 caracteres con lo generado; con 11 001, rojo | gate del cuerpo de H31 (`maximoDeCaracteresDelCuerpo`), sobre las cuatro skills | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio` |
| FR-010, SC-003 | 0 citas de un bloque fuera de la forma única en los `SKILL.md` y las referencias | subtest `forma-de-cita`, con sus dos mutantes | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio` |
| FR-013, SC-004 | 0 etiquetas `⚠` con líneas distintas entre dos `SKILL.md` | subtest `formas-compartidas`, con su mutante | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio` |
| FR-012, FR-019 | 0 diferencias entre cada región generada y lo que genera la sincronía; 0 líneas `⚠ NO CUBIERTO:` distintas de la forma del cuerpo sustituida; 0 filas de cobertura sin región | la deriva de `make skills-check` y la prueba de la región, con el mutante de la forma cambiada | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, `ci:internal/skills/referencias_test.go:TestRegionDeLaCobertura` |
| FR-023 | 0 diferencias entre las marcas de los paquetes y lo generado; 0 referencias de un paquete sin enlazar | la deriva y el defecto de la sincronía | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio` |
| FR-024, SC-005 | 0 defectos y 0 derivas con la vertical de ensayo; 0 `.go`; 0 cambios en un `SKILL.md` fuera de las marcas; 0 cambios en `skills/` y `data/` del repositorio | `TestVerticalDeEnsayo`, con su mutante | `ci:internal/skills/vertical_de_ensayo_test.go:TestVerticalDeEnsayo` |
| FR-004, FR-005, FR-006, FR-007, FR-015, FR-016, FR-017, FR-019, FR-021, SC-017 | 0 incumplimientos de cada criterio | los guiones `h32-entrada`, `h32-capacidades`, `h32-normativa` y `h32-paquetes` | `ci:internal/app/e2e_test.go:TestEntregaDelHito` |
| FR-032, SC-018 | 0 líneas de la búsqueda de FR-032 | el guion `h32-nombres-retirados` | `ci:internal/app/e2e_test.go:TestEntregaDelHito` |
| FR-025 | 0 nombres de `c991808` no empotrados fuera de la lista; 0 nombres de la lista empotrados | `TestSkillsRetiradas` | `ci:internal/app/retiradas_test.go:TestSkillsRetiradas` |
| FR-026, FR-027, FR-028, FR-029, FR-031, SC-006 | sale con 0 y retira, avisa o lo dice en `--dry-run` sin tocar nada; con el nombre pedido, 2 | el guion `h32-retiradas` | `ci:internal/app/e2e_test.go:TestEntregaDelHito` |
| FR-030 | lo de H19, sin cambios | `h19-skills-no-empotrada` | `ci:internal/app/e2e_test.go:TestEntregaDelHito` |
| SC-007 | descripción ≤ 900 caracteres; con 901, rojo | `maximoDeLaDescripcion` en la validación del frontmatter | `ci:internal/app/skills_test.go:TestSkillsDelRepositorio` |
| SC-007 | 0 descripciones marcadas en primera o segunda persona, decide | el juez de las descripciones de cada trabajo | `evals:consulta-juridica:habla_en_primera_o_segunda_persona`, `evals:normativa:habla_en_primera_o_segunda_persona`, `evals:jurisprudencia:habla_en_primera_o_segunda_persona`, `evals:plazos-y-territorio:habla_en_primera_o_segunda_persona` |
| FR-033, FR-037 | 0 evals que no se cargan; 0 skills de capacidad fuera de todo `no_se_activan` ajeno; 0 nombres de `se_activan` o `no_se_activan` que no son skills | `TestEvalsDelRepositorio` | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio` |
| FR-014 | 0 respuestas contadas como declaradas sin la norma y la cita del precepto en la misma línea | `TestJuzgarPlazos` | `ci:internal/evals/plazos_test.go:TestJuzgarPlazos` |
| FR-014, SC-011 | `cobertura_sin_declarar` 0 en cada modo, decide | el umbral del informe | `evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:orden`, `evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:herramienta` |
| FR-038, SC-010 | `afirma_lo_no_leido:claude-sonnet-5-5` ≤ 0.028 en `normativa` y ≤ 0.017 en `jurisprudencia`, deciden; los valores de `clases.yaml`, los de hoy | el umbral en tasa del informe y los tests de los valores | `evals:normativa:afirma_lo_no_leido:claude-sonnet-5-5`, `evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5`, `ci:internal/evals/tasa_test.go:TestLimiteDeLaTasa`, `ci:internal/evals/informe_test.go:TestUmbralesDeLasSkills` |
| FR-038 | 0 cambios que una skill carga sin elegir sus evals, también una referencia de un paquete | `TestEleccionDeLosMutantes` | `ci:internal/evals/eleccion_test.go:TestEleccionDeLosMutantes` |
| FR-035, SC-012 | 0 bloques de las `lecturas` de las evals 14, 15 y 17, y de la eval de la póliza, sin servir sin fecha y con `--fecha 9999-12-31`; la preparación da la falta y no abre la sesión | `TestCacheDeLasSesiones` y `TestLecturaDeLaEvalSinGrabar` | `ci:internal/evals/preparar_test.go:TestCacheDeLasSesiones`, `ci:internal/evals/preparar_test.go:TestLecturaDeLaEvalSinGrabar` |
| FR-039, SC-008 | ≤ 11 sesiones a la vez; tope ≥ peor caso de cada trabajo; ninguna skill de la entrega fuera y ningún nombre retirado dentro | `TestDefinicionDelJob` | `ci:internal/evals/definicion_test.go:TestDefinicionDelJob` |
| FR-035, FR-036, SC-015 | cada eval que decide pasa por la regla de su serie (2 de 3; ADR 0016), y una que no pasa pone el veredicto de su skill en fallo y el job en rojo: en `consulta-juridica`, que no publica `sin_activar` ([contracts/evals.md](contracts/evals.md) §7), es lo único que decide sobre sus sesiones, y una sesión que no activa la skill o una de su `se_activan` no pasa; en `normativa`, `jurisprudencia` y `plazos-y-territorio`, además, 0 sesiones sin activar su skill en cada modo, decide | el motivo de `se_activan` en el juicio de la sesión (`anotarLasQueSeActivan`); el veredicto del informe por la regla de cada serie (`veredictoDelInforme`, sin cambios); `sin_activar` en las tres skills que lo publican | `ci:internal/evals/juzgar_test.go:TestJuzgar`, `ci:internal/evals/informe_test.go:TestInforme`, `evals:normativa:sin_activar:claude-sonnet-5-5:orden`, `evals:normativa:sin_activar:claude-sonnet-5-5:herramienta`, `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden`, `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta`, `evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:orden`, `evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:herramienta` |
| SC-013 | los demás umbrales con `decide: true` de cada skill, con los valores de hoy | los de `umbralesDelInforme`, sin cambios | `evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:orden`, `evals:normativa:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`, `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden`, `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden`, `evals:jurisprudencia:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`, `evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:orden`, `evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:orden`, `evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:orden`, y los mismos en `herramienta`; `ci:internal/evals/informe_test.go:TestUmbralesDeLasSkills` |

`cuenta_su_proceso` (SC-014) y `fuera_de_lo_grabado` se publican sin decidir, como hoy: no son umbrales que decidan. El
sondeo de SC-016 lo mide una persona fuera del run.

## Uso, de fuera adentro

| Salida | Quién la pide y cuántas veces por pregunta | Tamaño con meses de uso | Cuándo deja de darse |
|---|---|---|---|
| El cuerpo de `consulta-juridica` | el agente, entero, una vez por pregunta legal que la activa | 3 660 caracteres, más unos 45 a 65 por paquete en su lista (la línea `- lo administrativo: \`plazos-y-territorio\`.`); no crece con lo consultado: cientos de normas y miles de bloques no añaden ninguna línea | es estático; cambia cuando una persona añade un paquete |
| El cuerpo de cada skill de capacidad | el agente, entero, una vez por pregunta que pide su capacidad | 10 858 (`normativa`), 10 842 (`jurisprudencia`), 10 896 (`plazos-y-territorio`, con la línea de su paquete); el gate los acota a 11 000 | estático |
| La línea del enlace de un paquete | parte del cuerpo; la lee el agente con él | unos 70 a 95 caracteres con su salto por paquete y skill (la de lo administrativo, 94, en `plazos-y-territorio`; la de la vertical de ensayo, 70) | crece con los paquetes que se escriben, no con lo consultado; el que no cabe pone `make ci` en rojo |
| `administrativo.md` | el agente, una vez por pregunta de plazos o de recursos o que nombra una norma del paquete, cuando el enlace se lo pide | unos 10 000 caracteres: la cobertura, 3 202 (9 filas de unos 340, con su línea); los tipos de plazo, unos 1 400; las quince vertebrales, unos 2 600; la prosa, unos 2 800. Crece con las filas que una persona escribe en `data/`, nunca con lo consultado | se deja de leer en una pregunta que no es de la materia |
| La línea `⚠ NO CUBIERTO:` | la respuesta, copiada de la referencia, una vez por pregunta cuyo procedimiento está en una fila | una línea de unos 215 a 340 caracteres (la de los tributos, unos 305; medidas en el borrador) | deja de darse para un procedimiento cuando su fila sale de la cobertura (H33 o su materia); no se repite: solo va en la pregunta que cae en su fila |
| La comprobación de `normativa` | `kitlegal graph check` una vez por bloque leído, detrás de su lectura; como mucho 50 hallazgos | sin `version-obsoleta`, nada en la respuesta; con ella, una línea `⚠ REDACCIÓN MODIFICADA:` por bloque | la línea deja de darse al leer el bloque de nuevo; el hito no añade ninguna señal: quita frases |
| Lo que `install` dice de un nombre retirado | quien instala, una vez por instalación | una entrada en `data` y una línea legible por skill de nombre retirado que el manifiesto declara: como mucho dos, sea cual sea lo consultado (ejemplo con bytes en [contracts/instalacion-retiradas.md](contracts/instalacion-retiradas.md) §3) | la retirada, una vez (la entrada sale del manifiesto); el aviso, en cada instalación hasta que la persona quita los ficheros, y la siguiente lo calla |
| `docs/SKILLS.md`, el ADR 0048, `CHANGELOG.md` | quien escribe una skill o un paquete, o actualiza kitlegal, una vez | unas secciones | cuando otro ADR cambie el reparto |

## Datos externos

| Fuente | Fila en `docs/SOURCES.md` de `main` | Manifiesto | Test de grabación | Lo que deja |
|---|---|---|---|---|
| `boe.legislacion-consolidada` | sí, «Revisado» 2026-09-13 | `internal/evals/testdata/grabaciones.json` (`"fuente": "boe.legislacion-consolidada"`), una entrada nueva: la búsqueda «contrato de seguro», `titulo_empieza_por` «Ley 50/1980,», bloques `a4` y `a22` | `TestGrabarEvals`, `//go:build grabacion`, `internal/evals/grabacion_test.go`, por `internal/httpx` | las respuestas en `internal/evals/testdata/boe.legislacion-consolidada/`; el material de origen en `$KITLEGAL_EVIDENCIAS` (`evidencias/H32/`) |

La entrada del manifiesto la escribe una tarea `[datos]`; la graba el paso `grabar_datos`, sin modelo; la tarea
siguiente lleva a `data/normas.yaml` el identificador y el título que da la búsqueda grabada, y
`TestIdentificadoresDeLasNormas` los compara con lo grabado (research D13). Ninguna otra fuente: las marcas `vertical`
son una clasificación del plan, sin fuente (research D5), y la vertical de ensayo son datos de un test, no normas.

## Trazabilidad

| Mecanismo | Lo exige |
|---|---|
| la skill `consulta-juridica` y sus dos evals | FR-001 a FR-004, FR-035, FR-036 |
| los nombres y el reparto; `jurisprudencia` conservada | FR-005, FR-007, FR-042 |
| los cuerpos reescritos (textos N, J, P) | FR-006, FR-009, FR-015 a FR-018, FR-021 |
| `vertical` en `data/normas.yaml` y su esquema | FR-021, FR-022 |
| regiones generadas dentro de una referencia a mano | FR-019, FR-022 (una referencia por skill con datos y prosa) |
| la región `cobertura_por_materia` filtrada por el paquete y con la línea de la fila | FR-012, FR-022 («sus filas») |
| la forma legible del precepto con la abreviatura | FR-010 (la forma única lleva forma legible), FR-019 |
| las marcas de los paquetes en las skills de capacidad y en la de entrada | FR-023, FR-001 (3), FR-024 |
| el defecto «fila de cobertura sin región» | FR-012 («por cada fila») |
| la vertical de ensayo y `TestVerticalDeEnsayo` | FR-024 |
| `forma-de-cita` y `formas-compartidas` | FR-010, FR-013 |
| `no_cubierto.precepto` y la medida con la cita | FR-014 |
| `se_activan` | FR-036 (US1, escenario 2) |
| la Ley 50/1980 en `data/normas.yaml`, su grabación y el manifiesto de H9 en `TestIdentificadoresDeLasNormas` | FR-035 (la sesión la lee sin red) |
| `ReglasDeConsultaJuridica`; la regla de los `no_se_activan` ajenos; la de los nombres que existen | FR-037, FR-033 |
| `dependencias.go` sin tabla en la skill de entrada; regiones como datos de la skill | FR-038, FR-001 |
| la matriz de cuatro trabajos con 1, 3, 5 y 2 | FR-039 |
| `instalacion.Retiradas`, la retirada, el aviso, los estados y `aviso` en `data` | FR-025 a FR-029 |
| `TestSkillsRetiradas`; `h32-retiradas` | FR-025, FR-031 |
| los nombres de hoy fuera de `internal/`, `README.md`, `release.yml`, `evals.yml`, la página `/instalar/` | FR-032 |
| `skillsDeHoy` (`skills_test.go`) y la orden del humo (`release_test.go`), en la raíz | FR-005, FR-007 (los nombres de `c991808` salen del árbol), FR-032 (el paso de `release.yml`), SC-001 |
| la retirada de los guiones de H31.1 | FR-040 |
| `docs/SKILLS.md`, el ADR 0048, `CHANGELOG.md` | FR-041, FR-042, FR-043 |

## Lo que dicen la guía, el ADR y el CHANGELOG (FR-041 a FR-043)

- **`docs/SKILLS.md`**: §1 dice el reparto (las cuatro skills, la capacidad de cada una, la de entrada), las dos clases
  de referencia más la de un paquete (a mano, con regiones generadas), el molde de [contracts/paquetes.md](contracts/paquetes.md),
  lo medido (los tres defectos del Contexto del spec, con sus cifras y sus fuentes: `brazo-P80`, las ejecuciones
  40700995075 y 40682099061, `e1bfb77`) y la tabla de controles con los nombres nuevos y `forma-de-cita` y
  `formas-compartidas`; §5, los defectos nuevos de `skills-check`. Las prácticas de las guías se citan con los dos
  enlaces que ya registra (agentskills.io, «Best practices for skill creators»; Anthropic, «Skill authoring best
  practices»): ninguna URL nueva.
- **`docs/ADR/0048-las-skills-se-reparten-por-capacidad-con-una-de-entrada.md`**: research D20.
- **`CHANGELOG.md`**, *Unreleased*: la skill de entrada, las de capacidad y sus nombres (con la tabla de los retirados),
  el paquete de lo administrativo, la forma única de cita, los dos arreglos de `normativa` y la retirada al instalar.

## Pendientes que el informe final da a una persona

El sondeo de SC-016; las capturas de `/instalar/` que enseñan una skill de hoy; `CLAUDE.md`, `CONTRIBUTING.md` y los
ADR anteriores, que nombran las skills de hoy; la grabación nueva de la Ley 50/1980 y su evidencia, y los ficheros
modificados de `testdata/` y `schemas/` (constitución, capa 3).

## Complexity Tracking

Ninguna violación que justificar: ninguna dependencia nueva ni paquete nuevo.
