# H19 · Esquemas y datos de prueba tocados en el hito

Lista para el informe final y su revisión humana posterior (constitución, capa 3; FR-145): cada fichero bajo
`schemas/` o bajo cualquier `testdata/` que el hito crea o modifica, con la tarea que lo tocó y su motivo. Sale de
`git diff --name-status main` sobre `1ae8471` (T025, el último commit antes de esta tarea), filtrado a esos dos
árboles, más los guiones que añade la activación de la suite (`95efed2`), que ya están en el diff de la rama:
`git diff --name-status main...HEAD` sobre `95efed2` da exactamente los ficheros de las tres tablas. Ningún fichero de
esos árboles se retira.

## Esquemas (`schemas/`)

| Estado | Fichero | Tarea | Motivo |
|---|---|---|---|
| A | `schemas/instalacion.json` | T014 `[datos]` | Contrato publicado de la salida de `install`, `list` y `doctor` del applet `skills` (FR-053; un `$defs` por verbo). Generado con `TestEsquemasPublicados -actualizar-esquemas`, no escrito a mano: `make schema-check` lo compara con `--describe` y `TestEsquemasCubrenTodosLosVerbos` lo exige en cuanto el verbo está registrado. T015 valida contra él la salida real, también sin manifiesto. |

Los esquemas que ya estaban en `main` no cambian.

## Datos de prueba (`testdata/`)

| Estado | Fichero | Tarea | Motivo |
|---|---|---|---|
| M | `internal/app/testdata/script/argumentos.txtar` | T014 `[datos]` | La lista literal de applets de las líneas 24 y 30 pasa a `boe, contar, echo, skills, territorio`: registrar `skills` la cambia (research V29). Solo esas dos líneas; ninguna aserción nueva ni retirada. |
| M | `internal/skills/testdata/script/instalar.txtar` | T019 `[datos]` | Reescrito según contracts/skills-e-invocacion.md §5: `make install` hace `go install` del binario y, con ese binario, `skills install -g --host claude`; deja las dos skills y su manifiesto en `~/.agents/skills/` y un enlace relativo por skill en `~/.claude/skills/`, con el destino literal `../../.agents/skills/<skill>`, y no crea `bin/instalado/` (US2 escenario 3, FR-125, FR-126, SC-018). |
| M | `internal/skills/testdata/script/instalar-de-nuevo.txtar` | T019 `[datos]` | La segunda `make install` sale con 0, dice `sin cambios` para las dos skills, deja el manifiesto byte a byte igual y los dos enlaces con su destino literal (FR-125, FR-126, SC-018). |
| M | `internal/skills/testdata/script/instalar-con-conflicto.txtar` | T019 `[datos]` | El enlace absoluto que dejaba el `make install` anterior en `~/.claude/skills/boe-legislacion` es un conflicto «enlace a otro sitio»: código distinto de 0, sin tocar `~/.agents` ni `~/.claude` (FR-125, FR-126; caso límite «Enlaces antiguos de `make install`» del spec). |
| M | `internal/skills/testdata/script/instalar-sin-gobin.txtar` | T019 `[datos]` | Sin `GOBIN`, el binario queda en el `bin` de `GOPATH` y las skills se instalan igual, con el binario por la ruta que da `go list -f '{{.Target}}'` (FR-125, FR-126). |

Los cuatro guiones `instalar*.txtar` comprobaban la instalación por enlaces de `scripts/instalar-skills.sh` y
`bin/instalado`, que ADR 0019 retira; cambian en la misma tarea que la receta `install` del `Makefile` porque, por
separado, `make ci` queda en rojo (*Complexity Tracking* de plan.md).

## Guiones que añade la activación de la suite

Tras la última tarea, el paso `activar_aceptacion` del workflow (`scripts/workflow/aceptacion.sh activar H19`) copia
los 18 guiones congelados de `specs/009-h19-instalar-sin-clonar/aceptacion/` a
`internal/app/testdata/script/h19-<nombre>.txtar` y los añade a la huella de `gates/aceptacion-congelada.json`. Los
escribió T001 `[aceptacion]` desde el spec y ninguna tarea posterior los toca; entran como ficheros nuevos, con los
mismos bytes que el congelado, y desde ahí los ejecuta `TestEntregaDelHito` en `make ci`. Lo que cubre cada uno es su
línea «Cubre» de cabecera.

| Estado | Fichero | Cubre |
|---|---|---|
| A | `internal/app/testdata/script/h19-skills-install-local.txtar` | US1.1, US1.4, US1.6 · FR-001, FR-003, FR-004, FR-010, FR-011, FR-014, FR-015, FR-016, FR-030 a FR-032, FR-050, FR-051, FR-122 · SC-001, SC-006, SC-021 |
| A | `internal/app/testdata/script/h19-skills-install-hosts.txtar` | US1.2, US1.3, US1.5 · FR-020 a FR-023, FR-025 · SC-002, SC-003 |
| A | `internal/app/testdata/script/h19-skills-ambito-global.txtar` | US6.1, US6.4 · FR-012 · SC-004 |
| A | `internal/app/testdata/script/h19-skills-ambito-dir.txtar` | US6.2, US6.3, US6.5 · FR-013, FR-052 · SC-005 |
| A | `internal/app/testdata/script/h19-skills-conflictos-entradas.txtar` | FR-040, FR-041 (a)-(d), FR-042, FR-043, FR-052 · SC-008, SC-009 |
| A | `internal/app/testdata/script/h19-skills-conflictos-rutas.txtar` | FR-022, FR-023, FR-026, FR-027, FR-035, FR-041 (g)-(h), FR-061, FR-067 · SC-009 |
| A | `internal/app/testdata/script/h19-skills-conflictos-dentro.txtar` | FR-028, FR-041 (e)-(g), FR-047 · SC-009 |
| A | `internal/app/testdata/script/h19-skills-dry-run.txtar` | FR-048 · SC-010 |
| A | `internal/app/testdata/script/h19-skills-idempotencia.txtar` | US3.4 · FR-033, FR-034, FR-045, FR-046 · SC-007 |
| A | `internal/app/testdata/script/h19-skills-aviso.txtar` | US3.1, US3.2, US3.5, US3.6, US3.7, US3.9 · FR-070, FR-071, FR-072, FR-075, FR-077 · SC-013, SC-022 |
| A | `internal/app/testdata/script/h19-skills-aviso-sin-aviso.txtar` | US3.3, US3.7, US3.8, US3.9 · FR-070, FR-072, FR-073, FR-074, FR-076 · SC-013 |
| A | `internal/app/testdata/script/h19-skills-list-doctor.txtar` | US4.1, US4.5 · FR-060, FR-061, FR-062, FR-067, FR-068 · SC-011 |
| A | `internal/app/testdata/script/h19-skills-doctor-hallazgos.txtar` | US4.2, US4.3, US4.4, US4.6, US4.7, US4.10, US4.11 · FR-065, FR-066, FR-077 · SC-011 |
| A | `internal/app/testdata/script/h19-skills-doctor-copia.txtar` | US4.3, US4.7, US4.8 · FR-024, FR-046, FR-069 · SC-012 |
| A | `internal/app/testdata/script/h19-skills-no-empotrada.txtar` | US4.9, US3.8 · FR-010, FR-036 · SC-011, SC-013 |
| A | `internal/app/testdata/script/h19-skills-invocacion.txtar` | US2.1, US2.2 · FR-081, FR-082, FR-084 · SC-014 |
| A | `internal/app/testdata/script/h19-instalador-correcto.txtar` | US5.4, US5.6 · FR-100 a FR-107 · SC-017 |
| A | `internal/app/testdata/script/h19-instalador-rechazos.txtar` | US5.5, US5.6 · FR-100 a FR-104, FR-108 · SC-017 |

## Lo que no está en la lista

- **Corpus de fuzz**: `FuzzLeerManifiesto` (T003) lleva su corpus en `f.Add` dentro del test; no hay ficheros nuevos en
  ningún `testdata/fuzz/`.
- **Material de los tests nuevos**: el dominio se prueba sobre el disco en memoria de
  `internal/core/instalacion/discoEnMemoria_test.go`, y el adaptador, el arnés e2e y el origen de release sintético de
  T017 construyen lo suyo en `t.TempDir()`; nada de eso es un fichero bajo `testdata/`.
- **Copias momentáneas**: las `zz-rojo-*` de la verificación de T001 y las `zz-instalador-*` de T022, T023 y T024
  (obligación 6 de plan.md; research D28) se crearon y se retiraron dentro de su tarea; ningún commit de la rama las
  contiene.
- **`evals/` y `data/`** no cambian: ninguna eval se toca (FR-086) y lo empotrado sale de `skills/`, que ya genera
  `make skills-sync` desde `data/`.

## Puntos de la Definition of Done que no aplican

- **ADR nuevo**: no aplica. ADR 0019 (`docs/ADR/0019-distribucion-binario-con-skills-empotradas.md`) ya decide la
  distribución por gestor de paquetes, las skills empotradas en el binario y `kitlegal skills install` local por
  defecto con los hosts por enlace; el hito lo implementa sin reabrirlo y no añade nada a `docs/ADR/`.
- **Fila de `docs/SOURCES.md`**: no aplica. El hito no toca ninguna fuente: `skills` es un applet calculado (procedencia
  `kitlegal.skills`, `kitlegal:applet/skills`) que no abre red. La subprueba sin red de `TestArquitectura` (T018; la
  amplió la revisión final) fija que ni el dominio de la instalación, ni el adaptador del disco, ni lo empotrado, ni
  los ficheros de `internal/app` que componen el applet y el aviso (`instalacion.go`, `empotradas.go`, `aviso.go`)
  importan la red o algo del módulo que la alcance; lo que esos ficheros llamen de otros ficheros del paquete no lo ve
  un análisis de importaciones, y lo miden los guiones e2e con los proxies cerrados. Ni `scripts/install.sh` ni la
  release son fuentes de datos. `docs/SOURCES.md` no cambia.
- **Grabaciones**: no aplica. No hay manifiesto `grabaciones.json`, ni test `//go:build grabacion`, ni paso
  `grabar_datos`, ni material en `evidencias/`. La única red que usó una tarea fue la del proxy de módulos de Go para
  `tools/goreleaser/go.sum` (T023; research S10), que no es una fuente de datos.

## Cobertura (Definition of Done, punto 9; FR-141, SC-015)

Medida sobre el código de `1ae8471` (T026 no toca código) con el perfil unitario `coverage.out` que deja `make test`,
dentro de `make ci`, y `go tool cover -func`; cada subtotal, con la misma orden sobre ese perfil filtrado a las líneas
del árbol, con su cabecera `mode:`.

| Árbol | Cobertura | Umbral |
|---|---|---|
| `internal/core/instalacion` | 99,9 % | ≥ 85 % |
| `internal/core/**` | 99,5 % | ≥ 85 % |
| Global | 97,7 % | ≥ 70 % |

Las cifras valen para `1ae8471`: si la activación de la suite o la revisión final cambian código, se vuelven a medir
sobre la cabeza final. La revisión final cambió código (`internal/core/instalacion/ambito.go`, `scripts/install.sh` y
tests); medidas de nuevo con el perfil de su `make ci`, sobre `95efed2` con sus cambios, dan las mismas tres cifras:
`internal/core/instalacion` 99,9 % (1194 de 1195 sentencias), `internal/core/**` 99,5 % y global 97,7 %.
