# Research: H41 · Las skills del producto mandan leer sus referencias

Verificaciones hechas en esta sesión (V), con dónde se comprobaron; supuestos que no se pueden comprobar sin red ni
modelo (S); y decisiones del plan (D), con la alternativa rechazada y por qué. «La #245» es lo que conserva la rama
`033-h41-las-skills-del` (`git show 033-h41-las-skills-del:<ruta>`), leída en esta sesión.

## Verificaciones

- **V1 · `skills/` es la de `v0.9.0` y `normativa-control` es `skills/normativa`.** `git diff --quiet v0.9.0 HEAD --
  skills/` sale con 0; `diff -r laboratorio/normativa/normativa-control/skills/normativa skills/normativa` no da nada;
  `git ls-tree -r --name-only v0.9.0 -- skills/normativa skills/plazos-y-territorio skills/jurisprudencia` da 17
  ficheros (9, 5 y 3). `laboratorio/normativa/` no existe en `v0.9.0` (la creó H40 después): que `normativa-control`
  sea la de `v0.9.0` se sigue de las dos primeras órdenes, no de un `git diff` de su carpeta. Las tres
  `<skill>-control` son, pues, `skills/<skill>` del árbol de partida (FR-020).
- **V2 · Cómo cuenta `make ci` el cuerpo.** `internal/skills/sincronia.go:584-597`: el cuerpo es lo que sigue al
  frontmatter; con más de 499 líneas o de 11 000 runas (`utf8.RuneCount`) da `<skill>: SKILL.md: cuerpo de N caracteres
  (máximo 11000)`. Lo ejecuta `TestSkillsDelRepositorio` (`internal/app/skills_test.go:88`), en `make skills-check` y
  en `make ci`. Los cuerpos del hito tienen 156 (`normativa`) y 159 (`jurisprudencia`) líneas.
- **V3 · Cuerpos medidos** con `awk 'f; NR>1 && /^---$/ && !f {f=1}' <SKILL.md> | LC_ALL=en_US.UTF-8 wc -m`, sobre el
  árbol y sobre copias en el directorio temporal con los cambios de contracts/remisiones.md y contracts/variantes.md:
  - hoy: `normativa` 10 929, `jurisprudencia` 10 842, `plazos-y-territorio` 10 930, `normativa-remisiones-en-orden`
    11 084;
  - tras el hito: `normativa` **10 926** (el `SKILL.md` de la #245 con el cambio de D1), `jurisprudencia` **10 842**,
    `plazos-y-territorio` 10 930 (no cambia);
  - las variantes de `plazos-y-territorio`: `-sin-condicion` **10 991**, `-condicion-por-orden` **11 245**,
    `-lista-con-verbo` **10 938**; con lo que movería H41.1 (D13), 10 882, 10 891 y 10 829.
  - La línea de la vertical de ensayo de H32 (`internal/skills/vertical_de_ensayo_test.go:38`, `lineaDelEnsayo`) mide
    70: con ella, 10 996, 10 912 y 11 000 en las tres skills del producto, dentro del tope (FR-008).
- **V4 · Ninguna suite congelada fija lo que cambia.** `grep` de «si se pregunta qué cambió», «tras leer»,
  «repásalas», «Antes de dar la consulta», «Nada de otra conversación» y «Si una lectura falla, di qué bloques» en
  `internal/app/testdata/script/` y en el código de `internal/skills` e `internal/evals`: solo un comentario sin
  relación (`h7-1-grafo-lecturas.txtar:4`) y la declaración de H40 sobre `normativa-control`, que no cambia. Con los
  mismos cambios de `normativa` y `jurisprudencia` salvo el de D1, el `make ci` de la #245 salió en verde
  (`specs/033-h41-las-skills-del/gates/informe-final.md` §1, en esa rama): también `h32-normativa.txtar`, que fija las
  cinco formas fijas en bloques `text`, la frase del paso 2 sobre otras conversaciones y los pasos 3 y 6.
- **V5 · Tamaños de las referencias** (`wc -c`): `normativa` 909 a 3510 bytes (`con_herramientas.md` 2442,
  `normas.md` 3510); `redaccion_modificada.md` 2359 → 2520 y `lecturas_que_fallan.md` 2776 → 2834 con lo movido (los
  de la rama 033); `jurisprudencia`: `consulta_preparada.md` 2038, `documento_aportado.md` 2805; `plazos-y-territorio`:
  `administrativo.md` 9930, `calculo_de_plazos.md` 2760, `jerarquia_normativa.md` 1543, `municipios.md` 3944.
- **V6 · A qué casos llegan las evals del sondeo de `plazos-y-territorio`.** `evals/plazos-y-territorio/02-…yaml`:
  Tordesillas, con `territorio` en 0 y la cobertura `no-configurado`; `05-…`, `10-…` y `11-…`: Leganés, con
  `territorio` en 0 y `plazos calcular`; la 05 lee el art. 124 de la LPAC el 2026-09-09, con fecha de vigencia
  20161002 (la de hoy); la 10, según su cabecera, termina `plazos calcular` con 3 (`sin_calcular`); la 11 trae
  `sin_servir` para la lectura. `grep -rn redaccion_posterior evals/plazos-y-territorio/ testdata/evals`: nada.
  Ninguna deja el municipio sin decir (también lo dice el ADR 0052).
- **V7 · testscript** (`github.com/rogpeppe/go-internal` v1.16.0, de `go.mod`, en la caché de módulos): `grep` compila
  su patrón con `(?m)` delante (`testscript/cmd.go:690`), así que `\s` cruza líneas; `! exists <f>…` falla si existe
  alguno (`cmd.go:280-300`); el subtest de un guion es su nombre sin `.txtar` (`testscript.go:326`).
  `go test -count=1 -run '^TestEntregaDelHito$/^h(39|40)-' ./internal/app/`, ejecutado en esta sesión: verde.
- **V8 · Lo que instala `kitlegal skills install --dir destino --json`**: `destino/<skill>/SKILL.md` por cada skill
  empotrada (`internal/app/testdata/script/h19-skills-ambito-dir.txtar:30-41`). Los guiones reciben `KITLEGAL_SKILLS`
  y `KITLEGAL_LABORATORIO_BIN` (`internal/app/e2e_test.go:159-175`), no la ruta de `laboratorio/`.
- **V9 · Biblioteca estándar** (`go doc`, Go 1.27.3): `os.CopyFS` no sobrescribe y sigue los enlaces de `dir`;
  `os.ReadDir` devuelve las entradas «sorted by filename»; en `encoding/json`, `omitempty` omite un slice vacío igual en
  v1 y v2, y `Unmarshal` ignora las claves sin campo.
- **V10 · `internal/skills` no va en el binario.** Fuera de los tests, solo lo importan ficheros de `internal/evals`
  (`grep -rl` de su ruta de importación), y `.goreleaser.yaml:30` construye solo `./cmd/kitlegal`.
- **V11 · Los umbrales del job que deciden.** `jq '.umbrales[] | select(.decide == true) | .nombre'` sobre
  `specs/032-h40-por-qu-e/gates/evals/*.json`: los nombres de «Controles de umbral» del plan; `consulta-juridica` no
  publica ninguno con `decide: true`.
- **V12 · Lo que ya funciona hoy en quickstart.md.** En esta sesión: `make skills-check`, verde; los guiones de H39 y
  H40, verdes; `TestArnesDeNormativa`, `TestArnesDelLaboratorio`, `TestRecetasDelSondeo`, `TestTablaDeNormativa`,
  `TestTablaDelLaboratorio`, `TestVariantesDeNormativa` y `TestLaboratorioFueraDelProducto`, verdes;
  `go run ./cmd/laboratorio tabla internal/evals/testdata/laboratorio/normativa/sondeo`, igual byte a byte (`cmp`) a su
  `tabla.md`. `git status --porcelain` quedó como estaba: el directorio del feature y el renombrado del
  fichero del hito, que ya estaba en el índice antes de esta sesión.
- **V13 · misspell lee «producto» como «production».** En el repositorio: la rama 033 añade `producto` a las palabras
  que ignora misspell en `.golangci.yml`, con ese motivo (`git diff main...033-h41-las-skills-del -- .golangci.yml`).
  No se ha ejecutado `golangci-lint` en esta sesión sobre ese código, que aún no existe en esta rama.

## Supuestos no verificados

- **S1.** Que con las órdenes del hito `normativa-producto` abra `con_herramientas.md` y `normas.md` como
  `normativa-remisiones-en-orden` en modo herramienta, y que ni `normativa` ni `jurisprudencia` queden por debajo de su
  `control`: es la aceptación de la persona (FR-052), con sesiones con modelo que esta sesión no abre.
- **S2.** Que las evals de las skills del producto sigan en verde en el cierre (FR-051): lo mide el job del cierre.
- **S3.** Qué forma de `plazos-y-territorio` se sigue mejor: es lo que mide el sondeo para el ADR de FR-053.

## Decisiones

### D1 · Las remisiones de `normativa` (FR-001 a FR-004): las de la #245, con un cambio

**Decisión.** Las de la #245 (su contrato, §2), con el texto exacto en contracts/remisiones.md §2 y su lista en §5. Un
solo cambio: en la orden de `redaccion_modificada.md`, «Por cada entrada `version-obsoleta`, **y** si se pregunta qué
cambió, lee …» pasa a «…, **o** si se pregunta qué cambió, lee …».

**Por qué.** Son dos casos que se dan por separado y llevan a la misma referencia: FR-002 (con el ADR 0052, punto 2)
pide unirlos con «o», no con «y». Mismo número de caracteres; ninguna eval del sondeo llega a esa referencia (D15), así
que no cambia lo que mide la aceptación. Las demás condiciones ya cumplen FR-002: «con una fecha en la pregunta»,
«con avisos o un id que termina en punto», «para varios bloques a la vez» y «si una orden no termina con `0`» las
trae el caso o lo que devuelve una llamada; las de `con_herramientas.md` y `normas.md` van sin condición.
**Rechazadas.** Dejar la «y» de la #245 (incumple FR-002); partirla en dos órdenes (dos enlaces a la misma referencia
en la misma viñeta, más caracteres y ninguna ganancia: FR-002 admite «o»); quitar «si se pregunta qué cambió»
(cambiaría una regla, que FR-007 no deja).

### D2 · El sitio en `normativa` (FR-006, FR-008)

**Decisión.** Los dos movimientos de la #245, palabra por palabra (contracts/remisiones.md §3): con ellos, el cuerpo
queda en 10 926 (10 996 con la vertical de ensayo; V3). No hace falta otro.

### D3 · Las remisiones de `jurisprudencia` (FR-005): las de la #245, sin cambios

**Decisión.** La de `consulta_preparada.md` pasa del final del paso 2 al párrafo que nombra `cita_preparar`, en la frase
que sigue a la que lo nombra, antes de la tabla de lo que se pide; la de `documento_aportado.md`, del final al principio del paso 3, antes de la viñeta que pide
el cotejo; la del paso 6 no cambia (contracts/remisiones.md §4). Cuerpo, 10 842. Ninguna junta casos con «y».

### D4 · `skills/plazos-y-territorio` no cambia, y su control lo compara todo (FR-010)

**Decisión.** Un subtest de `TestSkillsDelProductoDesdeControl` compara `skills/plazos-y-territorio` con
`plazos-y-territorio-control` ruta a ruta y byte a byte, **con lo generado** (`administrativo.md` y la región de los
paquetes), y nombra el fichero que difiere, falta o sobra. El test de las remisiones la deja fuera con una tabla
`skill → hito que la incluye` (`"plazos-y-territorio": "H41.1"`).
**Rechazada.** Dejar fuera lo generado, como en `normativa` y `jurisprudencia` (D8): FR-010 pide cada fichero idéntico.
Consecuencia (Complexity Tracking): un paquete que añadiera una línea a la región de `plazos-y-territorio` antes de
H41.1 pondría el subtest en rojo; este hito no añade ninguno (fuera de alcance).

### D5 · Una sola definición de «frase», en `internal/skills` (FR-001, FR-023)

**Decisión.** `skills.Frases(cuerpo string) []string` (`internal/skills/frases.go`), exportada, con su test de tabla
`TestFrases`: la usan `TestRemisionesDelRepositorio` (FR-001 a FR-003) y el control de que una variante de
`plazos-y-territorio` solo toca las órdenes de sus dos referencias (FR-023). Contrato: contracts/remisiones.md §6.
**Rechazadas.** Dos implementaciones, una en cada paquete de test (dos definiciones de la frase de FR-001 que pueden
divergir); un defecto nuevo de la sincronía (`sincronia.go`), que cambiaría las listas exactas de defectos de los
fixtures sintéticos de `internal/skills/sincronia_test.go` y de `internal/app/skills_test.go`, que enlazan referencias
sin «lee» (research D5 de la #245): la regla es del contenido de las skills del repositorio, no de la forma de toda
skill. `internal/skills` no va en el binario (V10): exportarla no añade nada a `kitlegal`.

### D6 · Qué es una frase (FR-001)

**Decisión.** La de FR-001, con las precisiones que necesita un lector: cada línea de un bloque de código es un segmento
aparte (un bloque corta la frase y sus líneas se comparan tal cual); una línea de encabezado (`#`) o que es entera un
comentario HTML (`<!-- … -->`) es un segmento por sí misma; la marca de una viñeta o de un paso abre la frase y va en
ella. «Lee» es la palabra entera (ni «leer» ni «léelo»); las tres expresiones prohibidas se buscan con los blancos
juntados y sin distinguir mayúsculas (la de `jurisprudencia-control` abre su frase: «Antes de dar la consulta, lee…»).
**Rechazadas.** Que una frase cruce un bloque de código (daría por buena una orden escrita antes del bloque para un
enlace escrito después); dejar las líneas del bloque fuera de los segmentos (FR-023 no vería un cambio dentro de un
bloque); que un encabezado o un comentario se junte con la línea siguiente (la región de los paquetes de
`plazos-y-territorio` se pegaría a la viñeta que la presenta).

### D7 · El control de FR-001 a FR-003: `TestRemisionesDelRepositorio`, con la lista del plan

**Decisión.** En `internal/skills/remisiones_test.go`, por cada skill de `skills/` con `references/` que no esté en la
tabla de D4: (1) cada frase con un enlace a `references/<f>` lleva «lee»/«Lee» antes del enlace y ninguna de las tres
expresiones, y cada referencia tiene alguna; (2) las frases con enlace, fuera de las regiones que genera `make
skills-sync`, son, en orden, las de la lista del plan (contracts/remisiones.md §5). Una skill con referencias sin
lista ni exclusión es un fallo. El subtest «control» aplica (1) al `SKILL.md` de `normativa-control` y
`jurisprudencia-control` y exige exactamente sus fallos (§7). Entra en la expresión de `make skills-check`.
**Por qué fuera lo generado en (2).** Una materia entra como datos sin tocar el código (principio X): la línea que un
paquete nuevo genera en la región de `normativa` no puede pedir un cambio de la lista. (1) sí la recorre, y la línea
generada ya dice «lee».
**Rechazada.** Comparar la lista con lo generado dentro.

### D8 · El producto se calcula desde control (FR-006, FR-007, FR-009)

**Decisión.** La de la #245 (research D7 y contrato §7): `TestSkillsDelProductoDesdeControl` aplica al control de
`normativa` y de `jurisprudencia` sus cambios declarados y sus movimientos y exige el resultado, byte a byte, sin lo que
genera `make skills-sync`; comprueba que cada texto movido está en su referencia y no en el cuerpo, y que el
frontmatter es el de control. Que lo generado sea lo de `make skills-sync` lo sigue comprobando
`TestSkillsDelRepositorio`. **Rechazada:** comparar también lo generado (un paquete nuevo pondría el test en rojo).

### D9 · `<skill>-control`, por huellas (FR-020)

**Decisión.** `TestVariantesDeControl` con la huella SHA-256 de cada uno de los 17 ficheros, por ruta, escrita al copiar
las dos nuevas (V1). **Rechazadas:** comparar con `skills/<skill>` (deja de valer al cambiar el producto; FR-020) y con
`git show v0.9.0:…` (el `make ci` del cierre corre sobre un clon superficial: research D8 de la #245).

### D10 · Las familias del producto, en Go (FR-020, FR-022, FR-030)

**Decisión.** Como en la #245 (research D9): `familiaDelLaboratorio` gana `producto`, y `familiasDelProducto` declara
`normativa`, `plazos-y-territorio` y `jurisprudencia`; `internal/evals/laboratorio_normativa.go` pasa a
`laboratorio_familias.go`. Cambia respecto a la #245: la de `plazos-y-territorio` tiene además sus tres variantes, en los
dos modos (data-model §1). **Rechazada:** una declaración en YAML (otro lector y otro esquema para datos que fija el
hito).

### D11 · Las tres variantes de `plazos-y-territorio` (FR-023)

**Decisión.** Nombres `plazos-y-territorio-sin-condicion`, `plazos-y-territorio-condicion-por-orden` y
`plazos-y-territorio-lista-con-verbo`; texto exacto en contracts/variantes.md §2. Dónde va cada orden:

- **Sin condición.** `municipios.md`, en un párrafo propio al principio del «Protocolo», antes de la primera orden o
  llamada, como `con_herramientas.md` en `normativa`. `calculo_de_plazos.md`, al final del paso 7, «Antes del 8»: la
  referencia es el detalle de los pasos 8 a 10, y el 8 ya sitúa los municipios que cuentan, su primer apartado.
  Rechazadas: delante de la llamada del paso 9 (el paso 8 ya habría usado la referencia; FR-003); al principio del paso
  8 (habría que reescribir su frase, que FR-023 no deja tocar).
- **Una condición del caso por orden.** `municipios.md`, al final del paso 2, detrás de `territorio resolver`, cuyo
  código trae la segunda mitad de su condición; la primera («si falta el municipio») ya se sabe en el paso 1. Las de
  `calculo_de_plazos.md`: «si alguno de los municipios cuenta», al final del paso 8, donde se sitúan; el plazo sin
  calcular, detrás de la llamada del paso 9 que lo trae («Con ese 3»); 4 o 5 y `redaccion_posterior`, detrás de la
  lectura del paso 10. Rechazadas: la de `municipios.md` en el paso 4 (cambiaría su frase, que no enlaza nada y no es
  una orden de esa referencia); antes del paso 2 (la condición «no sale con `0`» aún no se sabe; FR-003).
- **La lista con el verbo.** «Lee » delante de los dos enlaces de la lista, en su sitio; la línea de
  `calculo_de_plazos.md` se reparte en dos líneas de otra forma para no pasar de 120 columnas (mismos caracteres).

Lo que trae cada orden es el de su caso (FR-023: «cada una con lo que trae»): en la primera, lo que trae la referencia
entera, porque se lee siempre; en la segunda, lo del caso de cada orden. «Al presentar un territorio» y «para explicar
su cobertura» solo quedan en la tercera, que conserva las condiciones de control por definición.

### D12 · El control de FR-023: la declaración y las frases

**Decisión.** `TestVariantesDePlazosYTerritorio` (`internal/evals/laboratorio_familias_test.go`): (1) cada variante es,
ruta a ruta y byte a byte, la que da su declaración —sustituciones del `SKILL.md`, el mecanismo de H40
(`declaracionDeLaVariante`)— sobre `plazos-y-territorio-control`, y el «después» de cada sustitución es el texto del
plan; (2) quitadas de los dos cuerpos las frases (`skills.Frases`) que enlazan `references/municipios.md` o
`references/calculo_de_plazos.md`, el de la variante y el de control tienen los mismos segmentos en el mismo orden;
(3) toda sustitución es del `SKILL.md`. Así sale en rojo una variante que difiere de lo declarado (1) y una declaración
que toca otra cosa que esas órdenes (2 y 3). Los ayudantes de H40 (`comprobarLaVarianteDeclarada`, `calcularLaVariante`)
pasan a recibir la familia; `TestVariantesDeNormativa` no cambia ningún valor esperado (FR-021).
**Rechazadas.** Solo (1): no distingue una declaración que toca otra cosa. Un diff por líneas: una orden insertada al
final de una línea que ya tenía otra frase haría distinta esa línea entera.

### D13 · Los cuerpos de las variantes y lo que movería H41.1 (FR-024)

**Decisión.** Las tres pasan de 10 930 (V3). Lo que movería H41.1, con la regla del ADR 0052, punto 4 —solo a una
referencia a cuya condición llega una eval del sondeo en esa variante (D15) y solo una regla a cuyo caso llega una
eval—, medido en esta sesión sobre copias (contracts/variantes.md §3):

| Variante | Cuerpo | Movimientos de H41.1 | Tras ellos |
|---|---|---|---|
| `-sin-condicion` | 10 991 | M1 | 10 882 |
| `-condicion-por-orden` | 11 245 | M1, M2, M3 | 10 891 |
| `-lista-con-verbo` | 10 938 | M1 | 10 829 |

- **M1**: la frase del paso 10 «Si sale con 4 o 5, **`⚠ PLAZO:` se mantiene**: di aparte que no has podido leerlo, sin
  afirmar qué dice.» (105 caracteres), al final de `calculo_de_plazos.md`: llega la eval 11 (`sin_servir`).
- **M2** (solo la segunda): la regla del paso 9 «**Con el 3 y un `data.mensaje` que empieza por «el plazo de tipo», no
  hay fecha**: …, y sin el 10.», al final de `calculo_de_plazos.md`, con la orden «Con ese 3, lee …» reescrita con la
  condición entera: llega la eval 10.
- **M3** (solo la segunda): del paso 8, «: **el del interesado y el de la sede del órgano**, o en el
  contencioso-administrativo **el de la sede del juzgado**», que `calculo_de_plazos.md` ya dice en «Los municipios que
  cuentan»: la referencia se manda leer en las evals 05, 10 y 11 del sondeo, por «si alguno de los municipios cuenta»,
  que llegan a la regla en su rama administrativa; a la del contencioso llega la eval 09 del job
  (`09-plazo-contencioso-agosto.yaml`, la sede del juzgado), que el sondeo no usa. Si H41.1 lee el punto 4 como «una eval
  del sondeo» también para la regla, M3 no vale y le faltan unos 80 caracteres: lo dirá su plan.

**Rechazadas.** Mover una forma fija (el control 7 de H31 no deja que una línea de una referencia a mano empiece por
`⚠` sin ser una de `SKILL.md`); mover texto a `municipios.md` en la segunda variante (ninguna eval llega a su
condición: D15) o la línea de `REDACCIÓN DE HOY` (ninguna eval trae `redaccion_posterior`: V6); acortar lo que trae
cada orden para que quepa (mezclaría dos cambios en la medida, como dice el hito). El run no aplica ninguno.

### D14 · El arnés (FR-022, FR-030 a FR-033)

**Decisión.** El de la #245 (research D10 y D11): `<skill>-producto` se copia de `skills/<skill>` con `os.CopyFS` en
el temporal del sondeo (V9) y `sondeo.json` gana `referencias_del_producto`, campo de `sondeoLeido` con `omitempty`, que
los guiones de H39 y H40 no leen (V9: claves sin campo, ignoradas); el error de una mezcla nombra cada familia pedida en
el orden de las familias, y la de H39 con su texto de H40. Cambia respecto a la #245: la familia de
`plazos-y-territorio` acepta sus cinco variantes, y el test del arnés comprueba las cuatro invocaciones de FR-033.
**Rechazadas.** Un fichero aparte para la lista (otro lector y otro defecto); una copia o una huella de la skill (fuera
de alcance).

### D15 · Las evals del sondeo y las referencias a cuya condición llegan (FR-030, FR-052, FR-053)

Las cuatro de cada familia son las de la research D12 de la #245. Con las remisiones de este hito:

| Familia y variante | Llegan | No llega ninguna eval |
|---|---|---|
| `normativa` (control, producto, `remisiones-en-orden`) | `con_herramientas.md` y `normas.md` (04, 05, 15, 22); `redaccion_en_una_fecha.md` (22) | `buscar_la_norma.md`, `como_se_cita.md`, `lecturas_que_fallan.md`, `redaccion_modificada.md`, `varios_bloques.md` |
| `jurisprudencia` (control, producto) | `consulta_preparada.md` (01, 03, 06); `documento_aportado.md` (04, 06) | — |
| `plazos-y-territorio-control`, `-producto`, `-lista-con-verbo` | `administrativo.md` (05, 10, 11); `municipios.md` (02, por «al presentar un territorio»; en 05 y 11 también la abrió control en la #245, según el ADR 0052, porque el agente juzga que sitúa un territorio); `calculo_de_plazos.md` (05, 10, 11) | `jerarquia_normativa.md` |
| `plazos-y-territorio-sin-condicion` | `administrativo.md` (05, 10, 11); `municipios.md` (las cuatro); `calculo_de_plazos.md` (05, 10, 11, por llegar al 8) | `jerarquia_normativa.md` |
| `plazos-y-territorio-condicion-por-orden` | `administrativo.md` (05, 10, 11); `calculo_de_plazos.md` (05, 10, 11 por los municipios que cuentan; 10 por el 3; 11 por 4 o 5) | `jerarquia_normativa.md`; `municipios.md`: las cuatro resuelven su municipio con 0 y ninguna lo deja sin decir (V6) |

En `normativa` y `jurisprudencia` la lista coincide con la de la #245, que es la que compara FR-052. En
`plazos-y-territorio`, FR-053 compara cada variante con su control con esa misma lista (la de control); la tabla da la
celda de cada variante para cada referencia, y el desglose por eval del detalle deja ver que en
`-condicion-por-orden` ninguna eval llega a la condición de `municipios.md`: lo que esa celda mida no es una orden que
no se siga. Lo lee la persona al escribir el ADR (FR-053); la tabla no lo calcula (fuera de alcance).

### D16 · La lectura con `cd` (FR-041)

**Decisión.** La de la #245 (research D13): en las familias del producto, una orden simple cuya primera palabra que no
es una asignación es `cd` y cuya siguiente, sin una barra final, es `<skill>/references` o termina en
`/<skill>/references`, y detrás, en otra orden simple de la misma orden, una palabra que es exactamente `<f>` o `./<f>`.
La partición en órdenes simples y palabras es la de hoy (`ordenesSimples`, `internal/evals/laboratorio_tabla.go:1369`).
La familia de H39 no la usa. **Rechazada:** comparar el final del texto sin tramos de ruta (otra regla que la de `Read`).

### D17 · Las referencias de la tabla (FR-040)

**Decisión.** La de la #245 (research D14): la unión, en orden de nombre, de las de la familia por cada directorio de
variante versionada y de la lista de `sondeo.json` por cada directorio `<skill>-producto`; celda vacía donde la variante
no tiene la referencia. La lista la escribe el arnés siempre que el sondeo pide `<skill>-producto` (FR-032): la tabla
no la comprueba, y lo que no sabe leer es la regla genérica del spec (casos límite: defecto `inesperado`, código 1, sin
más casos).

### D18 · Aceptación e2e

**Decisión.** Dos guiones congelados (contracts/aceptacion.md): `skills-instaladas.txtar` (historias 1, 2 y 3: lo que
pone `kitlegal skills install`) y `laboratorio-tabla.txtar` (historia 5). La historia 3 va con la 1 y la 2 porque un
guion solo suyo —`plazos-y-territorio` sin cambios, ninguna variante instalada— pasaría hoy: el workflow exige que cada
guion falle antes de implementar. La historia 4 abre sesiones con modelo y credencial: su aceptación en `make ci` es
`TestArnesDeLasFamilias` con el sustituto de `claude`, como en H40 y la #245. La 6 la hace la persona.

### D19 · Lo que el hito documenta (FR-050)

`CHANGELOG.md` (*Unreleased*): el cambio de `normativa` y `jurisprudencia`. `docs/SKILLS.md` §1, tras «Cómo se remite»:
que una orden cuya condición la trae lo que devuelve una llamada va en el paso de esa llamada, detrás de ella; que el
sitio de las órdenes se hace llevando a la referencia que el caso manda leer el texto que solo ese caso necesita
(con el límite del ADR 0052, punto 4, que ya está aceptado); y que `make ci` y `make skills-check` comprueban la frase
de cada remisión. El punto 2 del ADR 0052, no (FR-050). Lo que el hito deja de hacer cierto se corrige en el mismo
cambio: `normativa-control` «es `skills/normativa`» y «las dos familias» en `CONTRIBUTING.md` («El sondeo del
laboratorio» y la tabla de objetivos), la ayuda de `make evals-laboratorio` y `make laboratorio-tabla`, la cabecera de
`scripts/evals-laboratorio.sh` y el comentario de paquete de `cmd/laboratorio`.

### D20 · El código de la #245 es el punto de partida

Las tareas parten del código de la rama `033-h41-las-skills-del` (local y en `origin`) para lo que vuelve tal cual
—`laboratorio_familias.go`, el arnés, la tabla, la lectura con `cd`, `remisiones_test.go`, `.golangci.yml`—, con los
cambios de este plan: el de D1, `plazos-y-territorio` fuera de las remisiones (D4), la lista (D7), `skills.Frases`
(D5), las tres variantes (D11, D12) y sus fixtures. Lo que no vuelve: los cambios de `skills/plazos-y-territorio` y sus
dos referencias; los guiones de la #245, que se reescriben (contracts/aceptacion.md); y, en
`referenciasDeLasVariantes` de `laboratorio_tabla.go`, la comprobación de un `sondeo.json` sin
`referencias_del_producto` en un sondeo con `<skill>-producto` y su mensaje («… no tiene la lista
referencias_del_producto de la variante …»): solo una edición a mano llega a ese estado, porque el arnés siempre
escribe la lista (FR-032), y no cambia nada que vea una persona o una skill (D17).

## Fuentes externas

Ninguna. El hito no graba nada ni toca ninguna fuente (FR-031): no hay manifiesto `grabaciones.json`, test
`TestGrabar*` ni tarea de grabación; las sesiones del sondeo se sirven de las grabaciones que ya usa el job (ADR 0043), y
nada se escribe en `evidencias/` ni en `data/` desde el run.

## Dependencias

Ninguna nueva: biblioteca estándar (`crypto/sha256`, `encoding/json`, `os`, `regexp`, `strings`, `unicode`) y lo que
ya usan `internal/skills` e `internal/evals`; `stretchr/testify` y `rogpeppe/go-internal` en los tests (constitución V).
