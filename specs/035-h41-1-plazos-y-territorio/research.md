# Research: H41.1 · `plazos-y-territorio` manda leer sus referencias sin condición

Fase 0 del plan. Cada decisión, con lo que se eligió, por qué y lo rechazado (constitución, «Criterio de decisión
autónoma»). Lo verificado lleva su orden y su resultado; lo que no se pudo medir en esta sesión, lo dice.

## Verificaciones de esta sesión

Todas en primer plano, sin red ni modelo. Las que cambian ficheros las hice sobre una copia parcial del repositorio en el
directorio temporal (`go.mod`, `go.sum`, `skills.go`, `internal/`, `cmd/`, `skills/`, `laboratorio/`, `data/`,
`evals/`, `schemas/`, y luego `Makefile`, `.github/workflows/evals.yml`, `.golangci.yml` y enlaces simbólicos a
`testdata/` y `scripts/` del árbol, que los tests solo leen), con el estado de destino
aplicado: `SKILL.md` de `plazos-y-territorio-sin-condicion` con M1 (D2), `calculo_de_plazos.md` con la frase al final
(D3), los tests de D1, D4 y D12 y los dos guiones de D6 y D7. El árbol de trabajo, el índice y el historial no
cambiaron (`git status --porcelain`: solo el directorio del feature).

- **V1 · Estado de partida.** `diff -rq skills/plazos-y-territorio
  laboratorio/plazos-y-territorio/plazos-y-territorio-control/skills/plazos-y-territorio`: sin diferencias. Las cuatro
  referencias de `-sin-condicion` son las de control (`diff -rq …/references`: sin diferencias). `make skills-check` en
  el árbol: verde (`internal/app`, `internal/skills`, `internal/evals`).
- **V2 · Cuerpos.** `awk 'f; NR>1 && /^---$/ && !f {f=1}' <SKILL.md> | LC_ALL=en_US.UTF-8 wc -m` (la medida de V3 de
  H41): `skills/plazos-y-territorio` **10 930**; `-sin-condicion` **10 991**; `-sin-condicion` con M1 **10 882**. El
  `SKILL.md` con M1 tiene 12 113 bytes (12 161 hoy), y su cuerpo, 11 052 (`wc -c`; y `LC_ALL=C wc -m`, que con la
  locale `C` cuenta bytes, da lo mismo). `calculo_de_plazos.md` pasa de 2 760 a **2 869 bytes**.
- **V3 · La frase de M1.** El `⚠` del fichero es U+26A0 sin selector de variante (`od -tx1`: `e2 9a a0`). La frase, con
  los blancos juntados, mide **104** caracteres (`wc -m`); los 105 del spec y del ADR 0053 cuentan también el espacio
  que la precede en el paso 10. Cuadra con V2: 10 991 − 10 882 = 109 = 1 (el espacio) + 104 + 4 (el salto
  «sin⏎····afirmar» frente a un espacio). Nada del diseño depende de la cifra: los controles comparan el texto.
- **V4 · La regla y las remisiones con el estado de destino** (un test temporal en la copia que imprime
  `fallosDeLaRegla` y `remisionesFueraDeLosPaquetes` sobre `skills.Frases`): un solo fallo,
  `plazos-y-territorio: references/jerarquia_normativa.md: un enlace suyo no va en una frase que lo mande leer («lee»
  antes del enlace)`; y tres remisiones fuera de los paquetes, en este orden, las de contracts/controles.md §2.2. La
  de `administrativo.md` está en la región de los paquetes y cumple la regla («lee» antes del enlace). Con
  `plazos-y-territorio-control`, la regla da tres fallos (`calculo_de_plazos.md`, `jerarquia_normativa.md` y
  `municipios.md`).
- **V5 · Los tests del estado de destino, en la copia.** Verdes: `TestSkillsDelProductoDesdeControl`,
  `TestVariantesDeControl`, `TestVariantesDePlazosYTerritorio`, `TestLaboratorioFueraDelProducto`,
  `TestRecetasDelSondeo` (`internal/evals`); `TestRemisionesDelRepositorio`, `TestVerticalDeEnsayo`, `TestFrases`
  (`internal/skills`); `TestSkillsDelRepositorio` y `TestEntregaDelHito/h41-*` con `h41-1-skills-instaladas`
  (`internal/app`). Con la skill de destino y el guion de H41 sin cambiar, `h41-skills-instaladas` sale en rojo en su
  línea 50 (`^- \[municipios\.md\]…`): FR-013 hace falta.
- **V6 · Rojo primero del guion de aceptación.** Con `skills/plazos-y-territorio` restaurada a la de hoy en la copia
  (igual a la del árbol: `diff -rq`), `h41-1-skills-instaladas.txtar` falla en su línea 21 por una aserción: `no match
  for "^## Protocolo\n\nAntes de la primera orden o llamada, lee …"`.
- **V7 · Las mutaciones de SC-001, en la copia y de una en una** (salen en rojo con este mensaje):
  - un byte añadido a `skills/plazos-y-territorio/references/calculo_de_plazos.md` →
    `TestSkillsDelProductoDesdeControl/plazos-y-territorio/desde_control`: `skills/plazos-y-territorio/references/
    calculo_de_plazos.md difiere de lo que dan …`;
  - un byte añadido a `…/plazos-y-territorio-sin-condicion/…/references/municipios.md` →
    `TestVariantesDePlazosYTerritorio` (la ruta de la variante) y `TestSkillsDelProductoDesdeControl` (la del producto);
  - un byte añadido a `…/plazos-y-territorio-control/…/references/municipios.md` → `TestVariantesDeControl` («difiere:
    su huella es …»);
  - un byte añadido al `SKILL.md` de `…/plazos-y-territorio-lista-con-verbo` → `TestVariantesDePlazosYTerritorio`;
  - «sin calcular» → «sin contar» en la remisión de `calculo_de_plazos.md` (`sed` del árbol a la copia) →
    `TestRemisionesDelRepositorio`: `plazos-y-territorio: la remisión 3 no es la del plan`;
  - «Lee » delante del enlace de `jerarquia_normativa.md` → `TestRemisionesDelRepositorio`: `plazos-y-territorio: la
    regla da exactamente los fallos que admite` y `la remisión 1 no es la del plan`;
  - 49 caracteres más en el cuerpo (una línea de 48 «x» y su salto): el cuerpo mide 10 931 y
    `TestSkillsDelRepositorio/skills` sigue en verde, y `TestVerticalDeEnsayo` sale en rojo con `plazos-y-territorio:
    SKILL.md: cuerpo de 11001 caracteres (máximo 11000)`. Confirma que la línea de la vertical suma 70 (10 882 + 70 =
    10 952; FR-004).
- **V8 · Lint.** `go tool -modfile=…/tools/golangci-lint/go.mod golangci-lint run ./internal/evals/ ./internal/skills/`
  sobre la copia, con el `.golangci.yml` del repositorio: `0 issues` (tras quitar una línea en blanco que gofumpt
  rechaza al borrar el subtest; D1).
- **V9 · `grep` de testscript.** `rogpeppe/go-internal@v1.16.0/testscript/cmd.go:662-715` (module cache):
  `scriptMatch` compila `(?m)` + el patrón con `regexp.Compile` y lo busca en el fichero entero; `-count=N` no vale con
  `!`. En RE2, `\n`, `\s` y `\z` (fin del texto) son válidos; con `(?m)`, `^` y `$` son de línea.
- **V10 · El guardián de diff.** `scripts/workflow/guardian-diff.sh:9-10, 75, 94-95`: una tarea solo toca `testdata/`
  con la etiqueta `[datos]`. `scripts/workflow/aceptacion.sh:77-80`: el guion de `aceptacion/` se activa como
  `internal/app/testdata/script/h41-1-<nombre>.txtar` (el hito en minúsculas y el punto por un guion).
  `scripts/workflow/informe.sh:389`: un fichero existente de `testdata/` modificado sale en el informe para la persona.
- **V11 · Los umbrales del cierre.** `jq '.umbrales[] | select(.decide == true) | .nombre'` sobre
  `specs/034-h41-las-skills-del/gates/evals/*.json` (el cierre de H41, los cuatro con `veredicto` `aprobado`): los de
  «Controles de umbral» del plan; `consulta-juridica` no publica ninguno con `decide: true`.
- **V12 · `make ci`.** Su objetivo `vuln` (`Makefile:118`) corre `govulncheck`, cuya base por defecto es
  `https://vuln.go.dev` (`govulncheck -h`): pide red. `test` y `test-integration` dejan `coverage.out` y
  `coverage-integration.out` en la raíz (`Makefile:87` y `:91`), que `.gitignore` ignora (`/*.out`, `/coverage.*`).
  **No he ejecutado `make ci` entero** en esta sesión, ni en el árbol ni en la copia: la copia es parcial (sin
  `evidencias/`, `docs/`, `specs/` ni `.git`), y en ella los tests que leen esos directorios fallan por eso
  (`open ../../evidencias/adr-0036/…`, `open ../../docs/SOURCES.md`), no por el cambio (V14).
- **V13 · El arnés también daba por hecho que el producto es su control.** Con el estado de destino y
  `testdata/` y `scripts/` del árbol enlazados en la copia (solo se leen), `TestArnesDeLasFamilias` salió en rojo en sus
  sesiones de `plazos-y-territorio-producto`: «la skill de la sesión … es, ruta a ruta y byte a byte, la de
  ../../laboratorio/…». Lo exige `origenesDeLaVariante` (`internal/evals/laboratorio_sondeo_test.go:1257-1275`): la skill
  de una sesión de `<skill>-producto` tiene que ser la de `skills/` y, en `plazos-y-territorio-producto`, también la de
  `plazos-y-territorio-control`, porque en H41 eran iguales (FR-022 de H41). Sin ese segundo origen (D12), en verde
  `TestArnesDeLasFamilias`, `TestArnesDeNormativa`, `TestArnesDelLaboratorio`, `TestTablaDeLasFamilias`,
  `TestLecturaConCd`, `TestTablaDeNormativa`, `TestTablaDelLaboratorio`, `TestVariantesDeNormativa`,
  `TestRecetasDelSondeo` y `TestLaboratorioFueraDelProducto`; `golangci-lint` sobre `internal/evals`, `0 issues`.
- **V14 · Nada más depende del cambio en los tres paquetes.** `go test -count=1 ./internal/evals/ ./internal/skills/
  ./internal/app/` sobre la copia, con el estado de destino (y D12) y con el de hoy restaurado: las mismas 210 líneas
  `--- FAIL:` en los dos (`diff` de las dos listas: idénticas), así que ninguna la causa el cambio; las que miré fallan
  por lo que la copia no lleva (`evidencias/`, `docs/`, `specs/`). Tras las dos medidas, `git status --porcelain` del
  árbol: solo el directorio del feature. Fuera de esos tres
  paquetes, ningún test cita `municipios.md`, `calculo_de_plazos.md`, la frase de M1 ni `plazos-y-territorio-control`
  (`grep` sobre `*_test.go` de la raíz, `cmd/` e `internal/`, y `web/src`): solo `laboratorio_familias_test.go` y
  `laboratorio_sondeo_test.go`.

## Decisiones

### D1 · `TestSkillsDelProductoDesdeControl` calcula `plazos-y-territorio` desde `-sin-condicion` (FR-005)

**Decisión.** `declaracionDelProducto` gana un campo, `desde`: el nombre de la variante versionada de la familia sobre
la que se aplican los cambios y los movimientos; vacío, la de control. `plazos-y-territorio` entra en
`declaracionesDelProducto` con `desde: "plazos-y-territorio-sin-condicion"`, un cambio (M1, D2) y un movimiento (D3).
`comprobarElProductoDesdeControl` lee esa variante como base de «desde control» y de «movimientos»; el frontmatter
sigue comparándose con el de control. El subtest de H41 que comparaba la skill con su control «hasta H41.1»
desaparece, y los comentarios de `declaracionesDelProducto` y del test dejan de decir que la skill es su control. El
mensaje del fichero que difiere nombra la base: «lo que dan su control y su declaración» para `normativa` y
`jurisprudencia`, como hoy, y «lo que dan plazos-y-territorio-sin-condicion y su declaración» para esta (D5).

**Por qué.** FR-005 pide «`-sin-condicion` con M1 y nada más»: lo que la variante cambia sobre control ya lo comprueba
`TestVariantesDePlazosYTerritorio` (FR-006), y declararlo otra vez sería duplicar sus tres sustituciones. Con la base
explícita, la declaración de esta skill es solo M1, que es lo que el hito añade. `normativa` y `jurisprudencia` no
cambian una línea de su declaración (FR-012). Verificado en V5 y V7.

**Rechazadas.** (a) Declarar `plazos-y-territorio` desde control con las sustituciones de §2.1 de H41 más M1: duplica
la declaración de la variante, y un cambio en una tendría que repetirse en la otra. (b) Un subtest propio para esta
skill con su propia lectura y comparación: repite `comprobarElProductoDesdeControl`. (c) Un campo obligatorio en las
tres declaraciones: cambia las de `normativa` y `jurisprudencia` sin necesidad.

### D2 · M1 en el cuerpo: dos líneas enteras por una (FR-003, FR-004)

**Decisión.** El cambio de la declaración sustituye, en el `SKILL.md` de `-sin-condicion`, las líneas enteras
``    `⚠ REDACCIÓN A FECHA:`. Si sale con 4 o 5, **`⚠ PLAZO:` se mantiene**: di aparte que no has podido leerlo, sin``
y `    afirmar qué dice.` por ``    `⚠ REDACCIÓN A FECHA:`.``. El paso 10 no se vuelve a partir en líneas.

**Por qué.** Es lo único que M1 quita, y con esta forma el cuerpo mide los 10 882 de FR-004 (V2). `conLaDeclaracion`
exige que el «antes» esté una vez en el fichero, y estas dos líneas lo están.

**Rechazada.** Recomponer el párrafo del paso 10 con su ancho: cambiaría más líneas que las de M1 («y ningún otro»,
FR-003) y la cifra de caracteres.

### D3 · M1 en la referencia: una línea en blanco y la frase en una línea (FR-003)

**Decisión.** `calculo_de_plazos.md` es la de v0.9.0 entera y, detrás, `\n` + la frase + `\n`: su última línea es la
frase, de 104 caracteres, detrás de una línea en blanco. En la declaración, `texto` y el cuerpo de `enLaReferencia`
son la misma constante, `siSaleConCuatroOCinco`.

**Por qué.** FR-003 pide la frase «entera, palabra por palabra, detrás de una línea en blanco»; en una línea, el texto
de la referencia es el mismo byte a byte que el del contrato y el guion lo busca tal cual (`…\n\z`). Las referencias a
mano no tienen ancho fijo (la línea 34 de `calculo_de_plazos.md` tiene más de 120 columnas) y ninguna comprobación de
`TestSkillsDelRepositorio` la rechaza (V5).

**Rechazada.** Partir la frase en dos líneas: no aporta nada y obliga al guion a `\s+`.

### D4 · `TestRemisionesDelRepositorio` sin exclusiones y con una tabla de fallos admitidos (FR-007)

**Decisión.** Sale `fueraDeLasRemisiones` (con `plazos-y-territorio` fuera, queda vacía) y su `continue`. Entra
`fallosAdmitidosDelProducto map[string][]string` con una sola entrada, `plazos-y-territorio`, y su único fallo, con un
comentario que dice por qué (la viñeta de `jerarquia_normativa.md` queda como en v0.9.0; FR-002, ADR 0053 punto 1 y «Lo
que cuesta»; ninguna eval llega a su condición). Por cada skill de `skills/` con `references/`, el test compara por
igualdad (`assert.Equal`) `fallosAdmitidosDelProducto[skill]` con lo que da `fallosDeLaRegla`: una skill que no está en
la tabla admite cero (`nil` frente a `nil`). `remisionesDelPlan` gana la lista de `plazos-y-territorio`, las tres frases
de V4 en su orden. `fallosDeControl` y el subtest «control» no cambian.

**Por qué.** Es la opción C de la clarificación del spec: la regla entera sobre todo el cuerpo, la lista del plan con sus
tres frases, y un fallo admitido que deja de valer en cuanto la viñeta cambie (V7: «Lee » delante del enlace pone el test
en rojo). Un mapa vacío de exclusiones sería un mecanismo sin requisito (constitución V). La igualdad sobre la lista
entera sustituye a un `t.Error` por fallo: el mensaje de `testify` muestra la lista obtenida, que sigue nombrando la
skill, la referencia y la regla de cada fallo (`docs/SKILLS.md` §1).

**Rechazadas.** (a) Dejar el mapa vacío: código muerto. (b) Quitar de la regla los fallos admitidos y mirar aparte que
sigan dándose: dos caminos para lo mismo. (c) Añadir `plazos-y-territorio-control` a `fallosDeControl`: el spec no lo
pide y la regla ya se ve fallar con los otros dos controles.

### D5 · El mensaje del fichero que difiere nombra la base (FR-005)

**Decisión.** `compararLosFicheros` recibe, en «desde control», «lo que dan <base> y su declaración», y el mensaje de
«movimientos» que hoy dice «sale del SKILL.md de control» pasa a «sale del SKILL.md de <base>», donde <base> es «su
control» si `desde` está vacío y el nombre de la variante si no. Los demás mensajes, como están.

**Por qué.** FR-005 pide que el rojo dé la ruta del fichero que difiere; quien lo lea tiene que saber contra qué se
calculó. Con el texto de hoy diría «su control» de una skill que sale de otra variante (V7 lo mostró).

**Rechazada.** Dejar el texto: describe mal la comprobación de esta skill.

### D6 · Un guion de aceptación para las historias 1 y 2 (ADR 0018; FR-001 a FR-003)

**Decisión.** La primera tarea escribe `specs/035-h41-1-plazos-y-territorio/aceptacion/skills-instaladas.txtar`
(contracts/aceptacion.md §1), activado como `internal/app/testdata/script/h41-1-skills-instaladas.txtar` (V10). Corre
`kitlegal skills install --dir destino --json` desde las skills empotradas y busca, en lo instalado, las dos órdenes con
sus saltos de línea y su sitio, una vez cada enlace, la lista sin las dos viñetas y terminada en la de
`jerarquia_normativa.md`, la pregunta del paso 1, el paso 10 terminado en ``⚠ REDACCIÓN A FECHA:`.`` y la frase al
final de `calculo_de_plazos.md` (`\z`). Las órdenes y la frase se buscan con sus bytes, porque FR-001 pide sus saltos de
línea; la ausencia de la frase en el cuerpo, con `\s+` entre palabras, porque el escenario US2.1 la busca «con los
blancos juntados».

**Por qué.** Las dos historias se ven en los mismos ficheros instalados: un guion, una instalación. La región de los
paquetes no se busca por su contenido —la genera `make skills-sync` desde `data/`, y un paquete nuevo la haría
crecer—: el guion ancla la viñeta de `jerarquia_normativa.md` a la marca de fin y a «## Protocolo». V5 y V6: verde con
la forma aplicada, rojo por una aserción con la de hoy.

**Sin guion.** La historia 3 son tests de Go de `make ci` (FR-005 a FR-007 y FR-013) y sus mutaciones (V7); la 4 abre
sesiones con modelo y la lanza la persona; la 5 es documentación, sin salida del binario.

**Rechazadas.** (a) Un guion por historia: dos instalaciones para los mismos ficheros. (b) Comparar las referencias
instaladas byte a byte dentro del guion: copiaría en el `txtar` lo que `TestSkillsDelProductoDesdeControl` ya compara.

### D7 · FR-013: el bloque de `plazos-y-territorio` del guion de H41, en la misma tarea que la skill

**Decisión.** El bloque (líneas 48-53) y las frases de la cabecera (11-13 y 17-20) de
`internal/app/testdata/script/h41-skills-instaladas.txtar` pasan a los de contracts/aceptacion.md §2; `normativa`,
`jurisprudencia` y las variantes no instaladas, como están. Lo cambia una tarea `[datos]` (V10), la misma que cambia
`skills/plazos-y-territorio` y la declaración de D1.

**Por qué.** Con la skill nueva y el guion viejo, `make ci` sale en rojo (V5); el guion nuevo, con la skill vieja, no se
cumple, porque su primer `grep` busca una orden que hoy no está (no es una medida de V: la mide la tarea que cambia el
guion, con la skill de hoy repuesta en la copia, en rojo en la línea 54 del fichero con la cabecera nueva, y con una
mutación por cada `grep` del bloque; contracts/aceptacion.md §2); y la declaración de D1 no se cumple con la skill a
medias. Juntos, cada tarea deja `make ci` en verde. Es
un fichero existente de `testdata/`: el informe final lo lleva a la persona (V10; spec, «Assumptions»).

**Rechazada.** Quitar el bloque del guion de H41 y dejarlo todo al de H41.1: FR-013 pide que pase a afirmar la forma
aplicada.

### D8 · `docs/SKILLS.md` §1 (FR-009)

**Decisión.** Dos cambios en «Cómo se remite», ninguno más:

1. La frase de las líneas 141-142, «`plazos-y-territorio` sigue como estaba hasta H41.1, que le aplica la forma de sus
   órdenes que mida el laboratorio (ADR 0052, punto 3).», pasa a dos: «H41.1 le aplica a `plazos-y-territorio` la forma
   que midió el laboratorio (ADR 0053): `municipios.md`, antes de la primera orden o llamada, y `calculo_de_plazos.md`,
   al final del paso 7, antes del 8; con ella, el agente abrió `municipios.md` en 39 y 40 de 40 sesiones, frente a 3 y
   11 con la lista de v0.9.0 (`evidencias/sondeos-20261109/`). Lo medido en `normativa` y en `plazos-y-territorio` es que lo que la skill necesita en las preguntas a las
   que llega se manda leer sin condición, antes de la primera orden o llamada o del paso que lo usa (ADR 0053, punto
   3).» Es la línea de FR-009; «De ahí, tres reglas más:» sigue como está.
2. En la viñeta de `TestRemisionesDelRepositorio` (líneas 158-169), sale «—salvo `plazos-y-territorio`, que el test
   declara fuera hasta H41.1—», y detrás de «Cada una lleva `lee` o `Lee` antes del enlace» entra «—salvo los fallos
   que el test admite por skill (`fallosAdmitidosDelProducto`), hoy uno: la viñeta de `jerarquia_normativa.md` de
   `plazos-y-territorio`, que queda como en v0.9.0 porque ninguna eval llega a su condición—».

El punto 2 del ADR 0052 no entra. El ejemplo de la línea 90 («como `plazos-y-territorio`») no cambia (spec, «Fuera de
alcance»).

**Por qué.** Es lo que pide FR-009, con la cifra del ADR 0053; las dos frases «hasta H41.1» describen el control que
cambia FR-007, y quitarlas sin más haría decir a la guía que toda frase con un enlace lleva «lee». La excepción va con
la regla que la viñeta incumple, la del verbo.

### D9 · `CHANGELOG.md` (FR-010)

**Decisión.** Una entrada en *Unreleased* → «Cambiado», detrás de la de H41: «**`plazos-y-territorio` manda leer
`municipios.md` y `calculo_de_plazos.md` sin condición** (H41.1; ADR 0053). `municipios.md`, antes de la primera orden o
llamada, y `calculo_de_plazos.md`, al final del paso 7, antes de situar el plazo; sus dos viñetas salen de la lista de
«Referencias». Para que el cuerpo quepa en el tope, la frase de una lectura que sale con 4 o 5 («`⚠ PLAZO:` se
mantiene») pasa al final de `calculo_de_plazos.md`. Ni la descripción, ni las órdenes que pide la skill, ni sus otras
referencias cambian. `make ci` comprueba que la skill es la forma que midió el laboratorio con ese movimiento y nada
más, y la frase de cada una de sus remisiones.»

**Por qué.** Es un cambio de lo que hace la skill para quien la instala (*Cambiado*), como la entrada de H41.

### D10 · `CONTRIBUTING.md` no cambia (FR-012; supuesto, impacto `alcance`)

**Decisión.** Las líneas 1813-1814 («`plazos-y-territorio-producto` es `plazos-y-territorio-control` byte a byte») y
1828-1830 («`skills/plazos-y-territorio`, su control byte a byte») de «El sondeo del laboratorio» describen el estado de
H41 y dejan de ser ciertas con el hito. FR-012 dice que, fuera de FR-001 a FR-011 y FR-013, nada más cambia, y ninguno
pide `CONTRIBUTING.md`: lectura conservadora, no se toca, y va a «Pendientes» del informe final para la persona.

**Rechazada.** Corregirlas en este hito: es contenido que el spec no pide.

### D11 · Lo que no cambia, y su control (FR-006, FR-008, FR-012; SC-005)

`laboratorio/plazos-y-territorio/` (cuatro variantes), `normativa`, `jurisprudencia`, `consulta-juridica`, `evals/`, el
arnés (`scripts/evals-laboratorio.sh`, `scripts/evals-sondeo-llavero.sh` y los `internal/evals/laboratorio*.go` que no
son `_test.go`), las recetas de `make evals-laboratorio` y `make laboratorio-tabla`, `cmd/laboratorio` y el generador de
las referencias no cambian un byte. `TestVariantesDeControl`, `TestVariantesDePlazosYTerritorio`, `TestRecetasDelSondeo`,
`TestLaboratorioFueraDelProducto` y los de la tabla siguen en verde sin cambiar ningún valor esperado (V5, V13); de
`TestArnesDeLasFamilias` cambia solo lo de D12. La invocación de US4 es la de
H41 (`specs/034-h41-las-skills-del/contracts/arnes.md` §1) con dos de sus cinco variantes: `SESIONES=40` es múltiplo de
4 (§2 del mismo contrato), y da 2 variantes × 2 modos × 40 = 160 sesiones, diez por eval.

### D12 · `TestArnesDeLasFamilias` deja de exigir que `plazos-y-territorio-producto` sea su control (FR-005, FR-008; supuesto, impacto `alcance`)

**Decisión.** `origenesDeLaVariante` (`internal/evals/laboratorio_sondeo_test.go`) devuelve, para cualquier
`<skill>-producto`, solo `skills/`: sale el segundo origen de `plazos-y-territorio-producto`
(`laboratorio/plazos-y-territorio/plazos-y-territorio-control/skills`), y su comentario deja de decirlo. Lo demás del
test —las sesiones que abre cada invocación, que cada sesión vea solo su skill, enlazada a la copia de su variante y
byte a byte la de `skills/<skill>` en `<skill>-producto`, el servidor y `sondeo.json`— no cambia. Va en la tarea `[datos]`
de D7, con la skill.

**Por qué.** Esa aserción no es del arnés, que copia `skills/<skill>` para `<skill>-producto` y no cambia (FR-008): es
el estado de H41 de la skill, «el producto es su control», el mismo que FR-005 sustituye en
`TestSkillsDelProductoDesdeControl` y FR-013 en el guion de H41. Con la skill del hito, `make ci` sale en rojo con ella
(V13), y FR-001 no se podría entregar. Es la lectura de FR-008 que deja el spec sin contradicción: el arnés y lo que sus
tests comprueban de él no cambian; la igualdad con el control, sí, porque deja de ser cierta. Lo anota
`gates/supuestos.md` para la persona.

**Rechazadas.** (a) Dejar el test como está: `make ci` en rojo. (b) Comparar la sesión con `-sin-condicion` con M1:
repetiría en el arnés lo que ya comprueba `TestSkillsDelProductoDesdeControl`, y la sesión ya se compara con
`skills/plazos-y-territorio`.

### D13 · Las evals del cierre (FR-011, SC-004)

Las evals no cambian. El veredicto del job del cierre (`medir_cierre`) sale en rojo con cualquiera de los umbrales con
`decide: true` de V11, que son los de «Controles de umbral»; ninguno se toca. `consulta-juridica` no publica ninguno con
`decide: true`: su veredicto lo deciden sus series (ADR 0016), configuración que el hito no toca.

### D14 · Skills de Go aplicadas

`golang-how-to` orquesta: `golang-testing` y `golang-stretchr-testify` (la declaración es una tabla; `assert.Equal` por
igualdad entera, `require` para lo que no deja seguir), `golang-project-layout` (ningún paquete ni fichero nuevo de Go:
los cambios van en `internal/evals/laboratorio_familias_test.go` e `internal/skills/remisiones_test.go`),
`golang-naming` (`desde`, `fallosAdmitidosDelProducto`, `siSaleConCuatroOCinco`, en el estilo de los nombres vecinos),
`golang-lint` (ningún `//nolint`; V8) y `golang-cli` (ningún argumento, bandera ni orden nuevos).

## Fuentes externas

Ninguna. El hito no graba nada ni pide datos de una fuente: no hay manifiesto `grabaciones.json`, test `TestGrabar*`,
fila nueva en `docs/SOURCES.md` ni escritura en `evidencias/` o `data/`. La tabla del sondeo la deja la persona en
`evidencias/`, fuera del run (US4).

## Supuestos no verificados

- Que el sondeo de la persona cumpla SC-003: abre sesiones con modelo, que esta sesión no abre.
- Que las evals del cierre salgan en verde (FR-011): las mide el job tras la revisión final.
- `make ci` entero (V12): lo ejecuta el paso `verificar` del run.
