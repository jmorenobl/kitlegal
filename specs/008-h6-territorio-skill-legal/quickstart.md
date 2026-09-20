# Quickstart: validación de H6

Guía **ejecutable** para comprobar que H6 entrega lo que dice. Cada escenario se ejecuta tal cual, desde la raíz del
repositorio, sobre la rama `008-h6-territorio-skill-legal` **con el hito ya implementado y confirmado**. Los nombres
de test son los del inventario de [plan.md](./plan.md) («Inventario de tests»); los formatos, los de
[contracts/](./contracts/). Los criterios de aceptación que valida cada escenario van en su título.

**Sin efectos colaterales.** Las secciones 1 a 13 no crean ni modifican ningún fichero del árbol de trabajo, no tocan
el índice ni el historial de git y no escriben en la caché de `kitlegal` de la cuenta. Lo que se escribe o se rompe a
propósito vive en `/tmp/kitlegal-quickstart-h6/`: los prerrequisitos crean esa carpeta vacía, cada defecto se provoca
sobre **su propio clon desechable** de la rama (`git clone` solo lee el repositorio) y la limpieza la borra. Lo único
que puede aparecer en el árbol son `bin/kitlegal` (`make build`) y `coverage.out` y `coverage-integration.out`
(`make ci`), ignorados desde H0 (`.gitignore`).

**Sin red**: ninguna de las secciones **1 a 13** necesita red **salvo dos cosas ajenas al hito**, las mismas que en
H4: la primera ejecución con la caché de módulos fría y `make vuln`, que corre dentro de `make ci` (§1) y consulta la
base de vulnerabilidades. El applet no abre conexiones en ningún camino (FR-043) y todos los tests —dominio, datos
congelados, esquemas, skill y evals— corren offline. La **sección 14 es la de plataforma**: la ejecuta la tarea
`[plataforma]`, necesita la propuesta de cambio abierta y el secreto del modelo, y no se ejecuta en local.

**El directorio del hito queda fuera de las comprobaciones del árbol**: el workflow `hito` reescribe ahí sus ficheros
de estado antes de cada tarea, así que `git status` descarta `specs/008-h6-territorio-skill-legal/` en cualquier
estado. Una ejecución retomada empieza siempre por los prerrequisitos.

**Cada orden se ejecuta tal cual en el modo desatendido**, con las formas que la sesión
`claude -p --permission-mode acceptEdits` acepta sin pedir aprobación (`.claude/settings.json`) y que el gancho de
`rtk` no reescribe:

| Necesidad | Forma de la guía | Forma que no se usa |
|---|---|---|
| Leer o filtrar la salida de `make`, `go test`, `git status`, `git diff`, `git log` o del binario | `rtk proxy` delante de la orden **y de cada etapa de la tubería** | la orden sin `rtk proxy`: el gancho la resume o la reescribe, y una línea vacía de más falsea cualquier recuento |
| Crear, borrar o editar en `/tmp/kitlegal-quickstart-h6/` | `rtk proxy mkdir`, `rtk proxy rm -r`, `rtk proxy chmod`, `rtk proxy perl -0pi` | `rm -rf` (denegado), `mkdir` y `rm` sin `rtk proxy`, `mktemp`, `printf … >> fichero` |
| Leer el código de salida, encadenar órdenes, capturar en una variable o redirigir | `rtk proxy sh -c '…; echo "código: $?"'`, en una sola línea | `…; echo "código: $?"` fuera de `sh -c`, `variable=$(…)` suelto, `cd` |
| Probar sobre un clon | `rtk proxy make -C /tmp/kitlegal-quickstart-h6/<clon> …`, `rtk proxy git -C …` | `cd` al clon |
| Leer un enlace o contar líneas | `rtk proxy readlink`, `rtk proxy wc -l` | `ls -l`, `ls -t` y `wc` sin `rtk proxy` |

`go`, `git`, `make` y `echo` con texto literal están permitidos tal cual —también `git clone` con destino en la
carpeta temporal—, pero `make` y `go test` van con `rtk proxy` delante para que su salida llegue entera.

## Prerrequisitos

```bash
go version
git rev-parse --abbrev-ref HEAD
rtk proxy make check-tools
rtk proxy git status --porcelain | rtk proxy grep -vE '^.. specs/008-h6-territorio-skill-legal/' ; echo "fin del estado"
rtk proxy mkdir -p /tmp/kitlegal-quickstart-h6
rtk proxy chmod -R u+w /tmp/kitlegal-quickstart-h6
rtk proxy rm -r /tmp/kitlegal-quickstart-h6
rtk proxy mkdir /tmp/kitlegal-quickstart-h6
rtk proxy make build
rtk proxy sh -c 'true; echo "código: $?"'
```

Espera: `go version go1.27.1 …`; la rama `008-h6-territorio-skill-legal`; `make check-tools` en verde; el estado
muestra solo `fin del estado`; la carpeta temporal queda recreada y vacía; `make build` deja `bin/kitlegal`; y la
sonda imprime `código: 0`, que es la forma con la que el resto de la guía lee códigos de salida.

## 1. Todo el hito en verde (SC-012, Definition of Done §1.1)

```bash
rtk proxy make ci
```

Espera: `ci: todos los controles en verde`. Diez prerrequisitos, entre ellos `schema-check` (que ahora cubre
`schemas/municipio.json`) y `skills-check` (que ahora cubre `data/territorio/`, `data/jerarquia.yaml` y las evals de
las dos skills). Ninguno de los controles que H6 añade pide nada por red; la única petición de toda la orden es la de
`make vuln` a la base de vulnerabilidades, que ya estaba en `make ci` desde H0.

## 2. El municipio cubierto (US1, FR-001 a FR-008, SC-001)

```bash
rtk proxy ./bin/kitlegal territorio resolver Leganés --json | rtk proxy jq .
rtk proxy ./bin/kitlegal territorio resolver Leganés --json | rtk proxy jq -r '.data | keys_unsorted | join(",")'
rtk proxy ./bin/kitlegal territorio resolver Leganés --json | rtk proxy jq -r '.data.cobertura'
rtk proxy ./bin/kitlegal territorio resolver Leganés --json | rtk proxy jq -r '.data.boletines[] | "\(.nivel) \(.codigo)"'
rtk proxy ./bin/kitlegal territorio resolver Leganés --json | rtk proxy jq -r '.data.dir3.codigo, .data.regimen.valor'
rtk proxy sh -c './bin/kitlegal territorio resolver Leganés --json > /dev/null; echo "código: $?"'
```

Espera, en el sobre: `ok` verdadero; `fuente` `kitlegal.territorio` y `url` `kitlegal:applet/territorio`; `hash` con
prefijo `sha256:`; `data` con **exactamente ocho claves**. Después: las ocho claves de la entrega; `cobertura` con sus
tres aspectos, los dos boletines `configurado` y el DIR3 `verificado`; un boletín `estatal`, uno `autonomico` y uno
`provincial`, los dos últimos con el mismo código, y el provincial con su `motivo`; el DIR3 no vacío y el régimen
`comun`; y `código: 0`.

## 3. Resolver por nombre, por código y con `--offline` da lo mismo (US1 escenarios 3 y 4, SC-001, D6)

```bash
rtk proxy sh -c 'set -e; t=/tmp/kitlegal-quickstart-h6; c=$(./bin/kitlegal territorio resolver Leganés --json | jq -r .data.codigo_ine.codigo); d=$(./bin/kitlegal territorio resolver Leganés --json | jq -r .data.codigo_ine.digito_de_control); ./bin/kitlegal territorio resolver "Leganés" --json > $t/nombre.json; ./bin/kitlegal territorio resolver "$c" --json > $t/codigo.json; ./bin/kitlegal territorio resolver "$c$d" --json > $t/digito.json; ./bin/kitlegal territorio resolver leganes --offline --json > $t/offline.json; cmp $t/nombre.json $t/codigo.json; cmp $t/nombre.json $t/digito.json; cmp $t/nombre.json $t/offline.json; echo "las cuatro salidas son idénticas byte a byte"'
```

Espera: la línea `las cuatro salidas son idénticas byte a byte`, y ningún mensaje de `cmp`. **Byte a byte lo mismo**,
incluida `fecha_consulta`, que es la del fichero congelado más antiguo y no la del reloj; y `leganes`, sin tilde y en
minúsculas, resuelve al mismo municipio (FR-015). Los cuatro ficheros se escriben en la carpeta temporal y los borra
la limpieza.

## 4. El municipio no cubierto no inventa nada (US2, FR-020 a FR-022, SC-002)

```bash
rtk proxy ./bin/kitlegal territorio resolver Tordesillas --json | rtk proxy jq -r '.data.cobertura'
rtk proxy ./bin/kitlegal territorio resolver Tordesillas --json | rtk proxy jq -r '.data.boletines[].nivel'
rtk proxy ./bin/kitlegal territorio resolver Tordesillas --json | rtk proxy jq -r '.data.provincia.nombre, .data.comunidad.nombre, .data.regimen.valor'
rtk proxy ./bin/kitlegal territorio resolver Tordesillas --json | rtk proxy grep -ciE 'bocyl|boletín oficial de castilla|boletín oficial de la provincia'
rtk proxy sh -c './bin/kitlegal territorio resolver Tordesillas --json > /dev/null; echo "código: $?"'
```

Espera: `cobertura` con el boletín autonómico y el provincial `no-configurado`; **solo** el nivel `estatal` en
`boletines`; provincia, comunidad y régimen completos; `0` coincidencias —el nombre o el código de un boletín no
configurado no aparece en ninguna parte de la salida—; y `código: 0`. La cuarta orden imprime el recuento `0`; su
propio código de salida es 1, porque `grep -c` no encontró nada, y lo que se comprueba es el número que imprime.

## 5. El régimen foral (US3, FR-055, SC-003)

El municipio no se escribe a mano: sale del propio registro congelado, que es quien fija los códigos (research.md S5).

```bash
rtk proxy sh -c 'set -e; f=$(grep -l "^regimen: foral$" data/territorio/comunidades/*.yaml | head -1); cc=$(sed -n "s/^codigo: \"\(.*\)\"$/\1/p" "$f"); m=$(grep -m1 "comunidad: \"$cc\"" data/territorio/municipios.yaml | sed "s/.*nombre: \"\([^\"]*\)\".*/\1/"); echo "comunidad foral: $cc · municipio: $m"; ./bin/kitlegal territorio resolver "$m" --json | jq -r ".data.regimen.valor, .data.cobertura.boletin_autonomico"'
```

Espera: la línea con la comunidad y el municipio elegidos, y después `foral` y `no-configurado`. El régimen consta
aunque su comunidad no tenga boletines configurados. La orden lee la `comunidad` de la fila del municipio, que es una
de las cinco columnas que FR-040 exige (data-model §3.1).

## 6. Ambigüedad, inexistencia y código mal formado (US4, FR-010 a FR-014, SC-004)

```bash
rtk proxy sh -c './bin/kitlegal territorio resolver "Villanueva" --json; echo "código: $?"'
rtk proxy ./bin/kitlegal territorio resolver "Villanueva" --json | rtk proxy jq -r '.data.clase, .data.mensaje'
rtk proxy sh -c './bin/kitlegal territorio resolver "Municipio Que No Existe" --json; echo "código: $?"'
rtk proxy sh -c './bin/kitlegal territorio resolver 99999 --json; echo "código: $?"'
rtk proxy sh -c './bin/kitlegal territorio resolver 2807 --json; echo "código: $?"'
rtk proxy sh -c 'set -e; c=$(./bin/kitlegal territorio resolver Leganés --json | jq -r .data.codigo_ine.codigo); d=$(./bin/kitlegal territorio resolver Leganés --json | jq -r .data.codigo_ine.digito_de_control); otro=$(( (d + 1) % 10 )); set +e; ./bin/kitlegal territorio resolver "$c$otro" --json; echo "código: $?"'
```

Espera, en este orden: **2** con clase `argumentos` y un mensaje que lista todos los candidatos con su código INE y su
provincia, ordenados por código; **3**; **3** (código bien formado que no está en la relación); **2** (cuatro cifras
no son un código); **2** (dígito de control distinto del oficial —se toma el de Leganés y se le suma uno— y el mensaje
dice cuál se esperaba). Ninguna invocación devuelve 4, 5 ni 6.

Si el nombre `Villanueva` resultara no ser ambiguo en la relación congelada, el caso ambiguo se toma del propio
registro:

```bash
rtk proxy sh -c 'sed -n "s/.*nombre: \"\([^\"]*\)\".*/\1/p" data/territorio/municipios.yaml | sort | uniq -d | head -3'
```

## 7. La matriz territorial y el tiempo, en el e2e (FR-090, FR-091, control 17)

```bash
rtk proxy go test -count=1 -run '^TestEntregaDelHito$/^territorio-matriz$' ./internal/app/
```

Espera: en verde. El guion ejerce los cuatro casos de la matriz —cubierto, no cubierto, foral y ambiguo— con sus
códigos de salida y el contenido del sobre, comprueba que en el caso no cubierto no aparece ningún boletín no
configurado, repite la invocación desde otro directorio de trabajo (FR-056) y mide el tiempo de una invocación con
`cronometra`.

## 8. El contrato publicado y `--describe` (FR-007, FR-092, SC-012)

```bash
rtk proxy make schema-check
rtk proxy ./bin/kitlegal territorio resolver Leganés --describe | rtk proxy jq -r '.title, (.properties.entrada.required | join(","))'
rtk proxy jq -r '.title, ([.["$defs"] | keys[]] | join(","))' schemas/municipio.json
rtk proxy sh -c './bin/kitlegal territorio resolver --describe; echo "código: $?"'
```

Espera: `schema-check` en verde; `--describe` con el título `territorio resolver` y `consulta` entre lo exigido; el
fichero publicado con el título `territorio · municipio` y la parte `resolver`. La última orden **no** describe nada:
el posicional es obligatorio y el análisis de la invocación corre antes de la decisión de describir, así que imprime
`argumentos inválidos: expected "<consulta>"` y `código: 2` (research.md V40, contrato del applet §1). Por eso el
verbo se describe siempre **con su argumento**, como los guiones del e2e de `boe`.

Para ver el control fallar, sobre un clon desechable:

```bash
git clone --quiet --branch 008-h6-territorio-skill-legal . /tmp/kitlegal-quickstart-h6/clon-esquema
rtk proxy perl -0pi -e 's/territorio · municipio/territorio · otra cosa/' /tmp/kitlegal-quickstart-h6/clon-esquema/schemas/municipio.json
rtk proxy sh -c 'make -C /tmp/kitlegal-quickstart-h6/clon-esquema schema-check; echo "código: $?"'
```

Espera: falla nombrando el fichero `municipio.json`, y `código: 2`. El árbol de trabajo no se toca.

## 9. Los ficheros congelados y su validación (US7, FR-044, FR-048, SC-007, SC-008)

```bash
rtk proxy go test -count=1 -run '^TestTerritorioDelRepositorio$' ./internal/skills/ -v
rtk proxy head -3 data/territorio/municipios.yaml
rtk proxy grep -c 'nombre:' data/territorio/municipios.yaml
rtk proxy ls data/territorio/comunidades/ | rtk proxy wc -l
rtk proxy grep -L 'boletines:' data/territorio/comunidades/*.yaml | rtk proxy wc -l
rtk proxy cat specs/008-h6-territorio-skill-legal/gates/verificacion-dir3.md
```

Espera: los subtests `esquema`, `integridad`, `fuentes`, `solo-madrid-configurada`, `regimen-de-todas`, `gramaticas`,
`pliegue-cubre-el-corpus`, `nombres-alcanzables` y `ningun-nombre-es-solo-cifras` en verde; la cabecera con `fecha` y
`source`; el número de municipios de la relación; 19 ficheros de comunidad; 18 sin `boletines` —solo la Comunidad de
Madrid los trae (FR-051)—; y, en el registro de verificación, una fila por municipio de la muestra con el código
derivado, el real y de dónde salió, incluidos el municipio fusionado o renombrado, el foral y el que tiene entidades
locales menores, con su conclusión.

Los tres últimos son los controles sobre el **corpus congelado real** y por eso viven aquí y no en
`internal/core/territorio`, que no puede leer ficheros ni en sus tests (research.md D27, V7).

Que un municipio sin DIR3 verificado se resuelve igual, sin código y declarado:

```bash
rtk proxy sh -c 'c=$(sed -n "s/.*\"\([0-9]\{5\}\)\": {.*/\1/p" data/territorio/municipios.yaml | while read -r x; do grep -q "\"$x\":" data/territorio/dir3.yaml || { echo "$x"; break; }; done); if [ -n "$c" ]; then ./bin/kitlegal territorio resolver "$c" --json | jq -r ".data.dir3.codigo, .data.dir3.source, .data.cobertura.dir3"; else echo "todos los municipios tienen DIR3 verificado"; fi'
```

Espera: dos líneas vacías y `no-verificado`, o el mensaje de que no hay ningún municipio sin verificar.

Que la integridad entre ficheros falla cuando debe —la fila de un municipio cuya `comunidad` no es la que declara su
provincia (data-model §2.1, punto 6)—, sobre un clon desechable:

```bash
git clone --quiet --branch 008-h6-territorio-skill-legal . /tmp/kitlegal-quickstart-h6/clon-datos
rtk proxy perl -0pi -e 's/comunidad: "\d\d"/comunidad: "99"/' /tmp/kitlegal-quickstart-h6/clon-datos/data/territorio/municipios.yaml
rtk proxy sh -c 'make -C /tmp/kitlegal-quickstart-h6/clon-datos skills-check; echo "código: $?"'
```

Espera: `TestTerritorioDelRepositorio/integridad` falla nombrando el municipio de la primera fila del fichero, la
comunidad que dice su fila (`99`) y la que declara su provincia, y `código: 2`. La sustitución cambia **solo la
primera** aparición, porque `perl -0pi` sin `/g` sustituye una vez sobre todo el fichero.

Y que el control del pliegue falla cuando aparece una runa que su tabla no cubre, sobre otro clon desechable:

```bash
git clone --quiet --branch 008-h6-territorio-skill-legal . /tmp/kitlegal-quickstart-h6/clon-pliegue
rtk proxy perl -0pi -e 's/nombre: "/nombre: "Ø/' /tmp/kitlegal-quickstart-h6/clon-pliegue/data/territorio/municipios.yaml
rtk proxy sh -c 'make -C /tmp/kitlegal-quickstart-h6/clon-pliegue skills-check; echo "código: $?"'
```

Espera: `TestTerritorioDelRepositorio/pliegue-cubre-el-corpus` falla nombrando el municipio de la primera fila y la
runa `Ø`, y `código: 2`. Es la demostración del control 5: una runa nueva en la relación del INE hace fallar `make ci`
en lugar de pasar en silencio (research.md D10).

## 10. Identificadores y fuzz (FR-030 a FR-035, SC-006)

```bash
rtk proxy go test -count=1 ./internal/core/ids/
rtk proxy ls internal/core/ids/testdata/fuzz/FuzzCodigoINE internal/core/ids/testdata/fuzz/FuzzCodigoDIR3
rtk proxy go test -run '^$' -fuzz '^FuzzCodigoINE$' -fuzztime 30s ./internal/core/ids/
rtk proxy go test -run '^$' -fuzz '^FuzzCodigoDIR3$' -fuzztime 30s ./internal/core/ids/
```

Espera: en verde. Las dos primeras órdenes son las que corren en `make ci` (el corpus versionado se ejecuta como
corpus semilla); las dos últimas son la campaña de fuzz, que se lanza a mano y **escribe** en el directorio de caché
de Go, no en el repositorio. Los dos objetivos viven en el fichero de test de su analizador —`ine_test.go` y
`dir3_test.go`— y su corpus, en el directorio con el nombre del objetivo. Si el fuzz encontrara un caso, `go` lo
escribiría en `internal/core/ids/testdata/fuzz/<objetivo>/`: entonces el árbol deja de estar limpio y ese fichero es
un hallazgo que hay que arreglar en el código, nunca borrar.

## 11. La skill y lo que se genera desde `data/` (US5, US6, FR-060 a FR-067, SC-009, SC-010)

```bash
rtk proxy head -14 skills/legal-core/SKILL.md
rtk proxy wc -l skills/legal-core/SKILL.md
rtk proxy head -2 skills/legal-core/references/leyes_vertebrales.md
rtk proxy head -2 skills/legal-core/references/jerarquia_normativa.md
rtk proxy readlink skills/legal-core/scripts/territorio
rtk proxy grep -c 'vertebral: true' data/normas.yaml
rtk proxy git diff --stat main -- skills/boe-legislacion
rtk proxy go test -count=1 -run '^(TestSkillsDelRepositorio|TestNormasDelRepositorio|TestIdentificadoresDeLasNormas)$' ./internal/app/ ./internal/skills/ ./internal/evals/
```

Espera: frontmatter con `kitlegal-applets: territorio` y las dos referencias; menos de 300 líneas; la cabecera
`<!-- generado desde data/normas.yaml, no editar -->` y `<!-- generado desde data/jerarquia.yaml, no editar -->`;
`../../../bin/instalado/kitlegal` como destino del enlace; quince normas marcadas; **un solo fichero cambiado en
`skills/boe-legislacion`, `references/normas.md`**, y solo por regeneración al entrar las siete normas (FR-074,
obligación 6: `SKILL.md` y todo lo demás de esa skill, intactos); y los tres tests en verde, incluido el que ata cada
identificador `BOE-A-…` a su búsqueda grabada.

Que la generación es idempotente y que la deriva se detecta, sobre un clon desechable:

```bash
git clone --quiet --branch 008-h6-territorio-skill-legal . /tmp/kitlegal-quickstart-h6/clon-skill
rtk proxy make -C /tmp/kitlegal-quickstart-h6/clon-skill skills-sync
rtk proxy git -C /tmp/kitlegal-quickstart-h6/clon-skill status --porcelain ; echo "fin del estado del clon"
rtk proxy perl -0pi -e 's/\z/<!-- editado a mano -->\n/' /tmp/kitlegal-quickstart-h6/clon-skill/skills/legal-core/references/jerarquia_normativa.md
rtk proxy sh -c 'make -C /tmp/kitlegal-quickstart-h6/clon-skill skills-check; echo "código: $?"'
```

Espera: `make skills-sync` deja el clon igual —su estado muestra solo `fin del estado del clon`—; tras la edición a
mano, `make skills-check` falla nombrando `legal-core` y el fichero, con `código: 2`.

## 12. Las evals y la compatibilidad del formato (FR-080 a FR-084, SC-011)

```bash
rtk proxy ls evals/legal-core/
rtk proxy go test -count=1 -run '^TestEvalsDelRepositorio$' ./internal/evals/ -v
rtk proxy git log --oneline --reverse -- evals/legal-core skills/legal-core/SKILL.md | rtk proxy head -5
rtk proxy git diff --stat main -- evals/boe-legislacion ; echo "fin del diff"
```

Espera: las tres evals (cubierto, no cubierto y no activación); los subtests `formato`, `conjunto`,
`conjunto-legal-core`, `normas-conocidas`, `grabado` y `cobertura-del-esquema` en verde; el primer commit de
`evals/legal-core` **anterior** al de `skills/legal-core/SKILL.md` (FR-083); y, en la última orden, **solo**
`fin del diff`, porque ninguna eval de `boe-legislacion` se ha modificado (FR-074, SC-015). La sonda es positiva a
propósito: se lee la ausencia de líneas de diff antes del rótulo, no un recuento, que el gancho de `rtk` podría
alterar con una línea vacía.

## 13. Documentación y fuentes (FR-049, FR-098, FR-099, SC-014)

```bash
rtk proxy grep -n 'territorio' CHANGELOG.md | rtk proxy head -5
rtk proxy grep -n 'legal-core' README.md CONTRIBUTING.md | rtk proxy head -5
rtk proxy grep -n 'mpt.rel' docs/SOURCES.md
rtk proxy sh -c 'grep -c territorio scripts/verify-sources.sh; echo "código de grep: $?"'
```

Espera: `CHANGELOG.md` (*Unreleased*) con el applet, la skill y los datos nuevos; `README.md` y `CONTRIBUTING.md` con
la skill y el directorio `data/territorio/` donde enumeran skills, applets o datos; la fila del REL con la fecha del
volcado; y **`0`** con `código de grep: 1`: `scripts/verify-sources.sh` no gana ningún caso para las fuentes
congeladas (ADR 0017).

## 14. `[plataforma]` Ejecución de aceptación (SC-013, SC-015)

Requiere la propuesta de cambio abierta y el secreto del modelo; no se ejecuta en local.

```bash
rtk proxy sh -c 'gh run list --workflow evals.yml --branch 008-h6-territorio-skill-legal --limit 3 --json databaseId,workflowName,status,conclusion,headSha'
rtk proxy sh -c 'set -e; id=$(gh run list --workflow evals.yml --branch 008-h6-territorio-skill-legal --limit 1 --json databaseId --jq ".[0].databaseId"); test -n "$id"; gh run view "$id" --log | sed -n "/^# Informe de evals/,/^## Sesiones/p"'
```

Espera: **una sola ejecución** con sus dos trabajos de la matriz en verde; en el informe de `legal-core`, las tres
evals con su tasa (al menos 2 de 3 sesiones), el commit evaluado y el identificador del modelo, y `red` vacío; en el
de `boe-legislacion`, su conjunto pasando con la regla de H5 y H5.1. La respuesta de la eval del municipio no cubierto
dice explícitamente qué no está configurado y no nombra ningún boletín que el applet no haya devuelto (SC-013). El
informe se guarda como evidencia en `specs/008-h6-territorio-skill-legal/gates/evals-cierre.md`.

## Limpieza

```bash
rtk proxy chmod -R u+w /tmp/kitlegal-quickstart-h6
rtk proxy rm -r /tmp/kitlegal-quickstart-h6
rtk proxy git status --porcelain | rtk proxy grep -vE '^.. specs/008-h6-territorio-skill-legal/' ; echo "fin del estado"
rtk proxy rm -r bin
```

Espera: la carpeta temporal desaparece con sus tres clones y los cuatro ficheros de la sección 3; el estado del árbol
muestra solo `fin del estado`. Los ficheros que estos escenarios escriben fuera de `bin/` y de la carpeta temporal son
`coverage.out` y `coverage-integration.out`, que genera `make ci` y que git ignora. La última orden es opcional:
`bin/kitlegal` lo regenera `make build`.
