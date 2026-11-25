# Mutantes de H46

Registro de T006 (SC-003, SC-004, SC-005, SC-006; FR-001, FR-002, FR-003, FR-005, FR-006, FR-014, FR-015, FR-016;
plan.md, «Mutantes» y «Controles de umbral»; quickstart Q3, Q4, Q5, Q8 y Q9; research D10 y V14). Una entrada por fila
de la tabla «Mutantes» del plan, con las 15 entradas de FR-014 y las 2 frases de FR-015 por separado, y otra por cada
orden de Q3, Q4, Q5 y Q8. Cada una dice la orden, lo esperado y lo visto, tal cual salió. Todo lo que sigue se ejecutó en
la sesión de T006 y se vio terminar; nada se da por visto de una sesión anterior.

## El clon

- Orden, en la raíz del repositorio: `git clone -q --no-hardlinks . "$TMPDIR/h46-mut"`, con `$TMPDIR` =
  `/tmp/claude/claude-1000` (lo que imprimió `os.Getenv("TMPDIR")` en la sesión), así que el clon es
  `/tmp/claude/claude-1000/h46-mut`.
- Visto: `HEAD` del clon en `7dad83dd feat(H46): T005`; `origin/main` en `78bb6f76 docs(roadmap): lanza H46, …`; `git
  status --short` sin salida.
- Cómo se corrió: con los permisos de esta sesión, `bash`, los bucles y las variables en la orden piden una aprobación
  que nadie da, así que cada orden de Q9 se corrió suelta, con `git -C <clon>` y `go -C <clon>` en vez de `cd` (la misma
  orden en el mismo directorio) y sin bucles: cada entrada de FR-014, escrita entera. Tras cada mutante, `git checkout -q
  -- <fichero>` en el clon y `git status --short` sin salida, salvo donde se dice. Abajo, `<clon>` es
  `/tmp/claude/claude-1000/h46-mut`.

## La línea de base, en el clon sin mutante

- `go -C <clon> test -count=1 -v -run '^TestSkillsDelRepositorio$/^presupuestos$' ./internal/app/` → 19 casos en `---
  PASS`: los cuatro de `consulta-juridica` (`cuatrocientas-noventa-y-nueve-lineas`, `quinientas-lineas`,
  `once-mil-caracteres`, `once-mil-y-un-caracteres`) y los cinco de cada una de `normativa`, `plazos-y-territorio` y
  `jurisprudencia` (`cuatrocientas-noventa-y-nueve-lineas`, `quinientas-lineas`, `nueve-mil-seiscientos-sin-los-paquetes`,
  `nueve-mil-seiscientos-uno-sin-los-paquetes`, `once-mil-y-un-caracteres`); `ok …/internal/app 1.699s`.
- `go -C <clon> test -count=1 -v -run '^(TestRenderizarTabla|TestRegenerarYComparar|TestCuerpoDeLaSkill)$'
  ./internal/skills/` → los tres en `--- PASS`; `ok …/internal/skills 0.048s`.
- `go -C <clon> test -count=1 -run '^TestSkillsDelProductoDesdeControl$' ./internal/evals/` → `ok …/internal/evals
  0.010s`.
- `go -C <clon> test -count=1 -run '^TestRemisionesDelRepositorio$' ./internal/skills/` → `ok …/internal/skills
  0.008s`.
- `go -C <clon> test -count=1 -v -run '^TestEntregaDelHito$/^(h41-skills-instaladas|h41-1-skills-instaladas)$'
  ./internal/app/` → los dos en `--- PASS`; `ok …/internal/app 0.734s`.
- Los dos guiones de la suite, copiados como en M28, sobre los `skills/` del hito: en M28.

## Los mutantes de `sincronia.go` (FR-005, FR-006; SC-003)

En todos, la orden que mide es la de Q9: `go -C <clon> test -count=1 -run '^TestSkillsDelRepositorio$/^presupuestos$'
./internal/app/`, con `grep -c -- '--- FAIL: .*presupuestos/[a-z-]*/'` sobre su salida, y los casos en rojo, con `grep
-o`.

### M1 · `maximoSinLosPaquetes = 9601` (FR-005, SC-003)

- Orden de Q9, tal cual: `sed -i 's/maximoSinLosPaquetes = 9600/maximoSinLosPaquetes = 9601/'
  internal/skills/sincronia.go`.
- Visto con ella: `git diff -U0` sin salida y `0` casos en rojo, `ok …/internal/app 0.739s`. El `sed` no cambia nada: en
  `sincronia.go:36` la constante va alineada por gofmt con las otras dos del bloque, `maximoSinLosPaquetes        =
  9600`, y el patrón de Q9 lleva un solo espacio. El mutante no llega a existir con esa orden.
- Orden que lo aplica: `sed -i 's/maximoSinLosPaquetes\( *\)= 9600/maximoSinLosPaquetes\1= 9601/'
  internal/skills/sincronia.go` (el `git diff -U0`: `-	maximoSinLosPaquetes        = 9600`, `+	maximoSinLosPaquetes
  = 9601`).
- Esperado: 6 casos de `presupuestos` en rojo (los de 9601 y de 11 001 de las tres skills con marcas).
- Visto: `6`: `normativa`, `plazos-y-territorio` y `jurisprudencia`, cada una con
  `nueve-mil-seiscientos-uno-sin-los-paquetes` y `once-mil-y-un-caracteres`; `FAIL …/internal/app 1.270s`.

### M2 · `maximoSinLosPaquetes = 9599` (FR-005, SC-003)

- Orden de Q9, tal cual: `sed -i 's/maximoSinLosPaquetes = 9600/maximoSinLosPaquetes = 9599/'
  internal/skills/sincronia.go`.
- Visto con ella: `git diff --numstat` sin ninguna línea y `0` casos en rojo, por lo mismo que en M1.
- Orden que lo aplica: `sed -i 's/maximoSinLosPaquetes\( *\)= 9600/maximoSinLosPaquetes\1= 9599/'
  internal/skills/sincronia.go`.
- Esperado: 9 casos (los 6 de M1 y, además, los de 9600).
- Visto: `9`: en cada una de las tres skills con marcas, `nueve-mil-seiscientos-sin-los-paquetes`,
  `nueve-mil-seiscientos-uno-sin-los-paquetes` y `once-mil-y-un-caracteres`; `FAIL …/internal/app 1.228s`.

### M3 · La regla que no se aplica: el `sincronia.go` de `main` con los tests del hito (FR-005, SC-003)

- Orden: `git -C <clon> show origin/main:internal/skills/sincronia.go > <clon>/internal/skills/sincronia.go` (`git diff
  --numstat`: `6 32 internal/skills/sincronia.go`).
- Esperado: 6 casos (los de 9601 y de 11 001 de las tres skills con marcas).
- Visto: `6`: `nueve-mil-seiscientos-uno-sin-los-paquetes` y `once-mil-y-un-caracteres` de `jurisprudencia`,
  `plazos-y-territorio` y `normativa`; `FAIL …/internal/app 1.271s`.

### M4 · Sin restar las líneas de los paquetes (FR-005)

- Orden: `sed -i -e 's/region, conLasMarcas, err := regionDeLosPaquetes.buscar(cuerpo)/_, conLasMarcas, err :=
  regionDeLosPaquetes.buscar(cuerpo)/' -e 's/sinLosPaquetes := utf8.RuneCount(cuerpo) -
  utf8.RuneCount(cuerpo\[region.desde:region.hasta\])/sinLosPaquetes := utf8.RuneCount(cuerpo)/'
  internal/skills/sincronia.go` (la primera, para que `region` sin uso no deje el paquete sin compilar).
- Esperado: 3 casos de `presupuestos/plazos-y-territorio`, la única skill con algo entre sus marcas (94 caracteres; Q3).
- Visto: `3`: `plazos-y-territorio/nueve-mil-seiscientos-sin-los-paquetes`,
  `plazos-y-territorio/nueve-mil-seiscientos-uno-sin-los-paquetes` y `plazos-y-territorio/once-mil-y-un-caracteres`;
  `FAIL …/internal/app 1.225s`.

### M5 · El defecto de FR-005 delante del de 11 000 (FR-005)

- Orden (el de 11 000 se guarda aparte y se añade detrás, en las dos salidas de `defectosDelCuerpo`):

  ```sh
  sed -i -e '/^func defectosDelCuerpo/,/^}/{/if caracteres := utf8.RuneCount(cuerpo); caracteres > maximoDeCaracteresDelCuerpo {/{i var deLosOnceMil []*DefectoDeSkill
  n;s/defectos = append(defectos,/deLosOnceMil = append(deLosOnceMil,/};s/return defectos$/return append(defectos, deLosOnceMil...)/}' internal/skills/sincronia.go
  ```

  (`git diff -U0`: `+var deLosOnceMil []*DefectoDeSkill`; `defectos = append(defectos, &DefectoDeSkill{` →
  `deLosOnceMil = append(deLosOnceMil, &DefectoDeSkill{`; los dos `return defectos` → `return append(defectos,
  deLosOnceMil...)`).
- Esperado: los `once-mil-y-un-caracteres` de las skills con marcas.
- Visto: `3`: `normativa/once-mil-y-un-caracteres`, `plazos-y-territorio/once-mil-y-un-caracteres` y
  `jurisprudencia/once-mil-y-un-caracteres`; `FAIL …/internal/app 1.245s`.

### M6 · La regla también sin las marcas (FR-006)

- Orden: `sed -i -e 's/region, conLasMarcas, err := regionDeLosPaquetes.buscar(cuerpo)/region, _, err :=
  regionDeLosPaquetes.buscar(cuerpo)/' -e 's/if err != nil || !conLasMarcas {/if err != nil {/'
  internal/skills/sincronia.go` (sin marcas, `buscar` da `entreMarcas{}` y no se resta nada).
- Esperado: `presupuestos/consulta-juridica`: `once-mil-caracteres` y `once-mil-y-un-caracteres`.
- Visto: `2`: `consulta-juridica/once-mil-y-un-caracteres` y `consulta-juridica/once-mil-caracteres`; `FAIL
  …/internal/app 1.196s`.

## Los mutantes de `comandos.go` (FR-001, FR-002, FR-003)

### M7 · La tabla de cuatro columnas: el `comandos.go` de `main` (FR-001, FR-003)

- Orden: `git -C <clon> show origin/main:internal/skills/comandos.go > <clon>/internal/skills/comandos.go` (`git diff
  --numstat`: `75 14 internal/skills/comandos.go`); `go -C <clon> test -count=1 -v -run
  '^(TestRenderizarTabla|TestRegenerarYComparar|TestCuerpoDeLaSkill)$' ./internal/skills/ | grep -E -- '^--- FAIL'`.
- Esperado: `TestRenderizarTabla`, `TestRegenerarYComparar` y `TestCuerpoDeLaSkill` en `--- FAIL`.
- Visto: `--- FAIL: TestRenderizarTabla`, `--- FAIL: TestCuerpoDeLaSkill` y `--- FAIL: TestRegenerarYComparar`; por
  dentro, `TestRenderizarTabla/una-sola-clave-de-fallo`, `secciones-en-el-orden-declarado`, `banderas-propias`,
  `sintaxis-y-escapes`, `opcionales-de-posicion` y `filas-del-contrato`,
  `TestCuerpoDeLaSkill/quinientas-lineas-al-regenerar` y `TestRegenerarYComparar/escribir-sincroniza`; `FAIL
  …/internal/skills 0.049s`.

### M8 · `RenderizarTabla` sin la llamada a `mismoSobreYBanderas` (FR-002)

- Orden: `sed -i -e '/^func RenderizarTabla/,/^}/{/comun := secciones\[0\]\[0\]/{N;d};/if err :=
  mismoSobreYBanderas(comun, verbo); err != nil {/{N;N;N;d}}' internal/skills/comandos.go` (salen el `if` con su
  `return nil, err` y la línea de `comun`, que solo usaba esa llamada); `go -C <clon> test -count=1 -v -run
  '^TestRenderizarTabla$' ./internal/skills/`.
- Esperado: `TestRenderizarTabla`, casos `banderas-distintas` y `sobre-distinto`.
- Visto: `--- FAIL: TestRenderizarTabla/banderas-distintas` y `--- FAIL: TestRenderizarTabla/sobre-distinto`; los otros
  8 casos en `--- PASS`; `FAIL …/internal/skills 0.008s`.

## Las 15 entradas de FR-014, de una en una (SC-004)

Cada entrada se deshace en su fichero de skill con un test de un solo uso, `TestDeshacerH46`, que se copió en el clon
como `internal/evals/deshacer_h46_test.go` (con `//go:build mutanteh46`, así que ninguna otra orden lo compila) y se
borró del clon al terminar; nunca estuvo en el árbol de trabajo. Lee la entrada de `declaracionesDelProducto` y, en el
fichero de la skill, cambia su `despues`, que tiene que estar una vez, por su `antes`; con un `despues` vacío (entradas 3,
4 y 5), pone el `antes` delante de su ancla, que también tiene que estar una vez: la línea que lo sigue en `main`. Las
entradas, en el orden de `declaracionesDelProducto`: las nuevas de H46 son las últimas de `cambios` y de
`enLasReferencias` de cada skill.

| N.º | Skill y fichero | Dónde en `declaracionesDelProducto` | Ancla |
|---|---|---|---|
| 1 | `normativa/SKILL.md` | `cambios`, la última | — |
| 2 | `jurisprudencia/SKILL.md` | `cambios`, la 5.ª desde el final | — |
| 3 | `jurisprudencia/SKILL.md` | `cambios`, la 4.ª desde el final | `- La respuesta da la consulta con lo que devuelve la operación` |
| 4 | `jurisprudencia/SKILL.md` | `cambios`, la 3.ª desde el final | `- **Si el documento no es el pedido, no lo cites como si fuera el pedido.**` |
| 5 | `jurisprudencia/SKILL.md` | `cambios`, la 2.ª desde el final | `\n### 5. Qué se dice de una sentencia\n` |
| 6 | `jurisprudencia/SKILL.md` | `cambios`, la última | — |
| 7 | `jurisprudencia/references/consulta_preparada.md` | `enLasReferencias`, la 4.ª desde el final | — |
| 8 | `jurisprudencia/references/consulta_preparada.md` | `enLasReferencias`, la 3.ª desde el final | — |
| 9 | `jurisprudencia/references/documento_aportado.md` | `enLasReferencias`, la 2.ª desde el final | — |
| 10 | `jurisprudencia/references/documento_aportado.md` | `enLasReferencias`, la última | — |
| 11 | `plazos-y-territorio/SKILL.md` | `cambios`, la 3.ª desde el final | — |
| 12 | `plazos-y-territorio/SKILL.md` | `cambios`, la 2.ª desde el final | — |
| 13 | `plazos-y-territorio/SKILL.md` | `cambios`, la última | — |
| 14 | `plazos-y-territorio/references/municipios.md` | `enLasReferencias`, la 2.ª desde el final | — |
| 15 | `plazos-y-territorio/references/municipios.md` | `enLasReferencias`, la última | — |

**Control del mecanismo**: con las 15 deshechas a la vez, de la 15 a la 1 (`go -C <clon> test -count=1 -tags
mutanteh46 -v -run '^TestDeshacerH46$' ./internal/evals/ -args -h46-entradas=15,14,13,12,11,10,9,8,7,6,5,4,3,2,1`, `ok`
y las 15 líneas `entrada N deshecha en …`), `git -C <clon> diff --numstat origin/main -- skills/` da solo los tres
`SKILL.md` (`4 8` en `jurisprudencia`, `11 15` en `normativa`, `6 10` en `plazos-y-territorio`) y ninguna referencia, y
sus líneas distintas son todas filas de la tabla de comandos, líneas en blanco y, en cada uno, la del sobre y la de las
banderas comunes: la diferencia de T002 y nada más. Cada entrada deshecha es, pues, el texto de `main`.

**Orden de cada mutante** (N, del 1 al 15, escrito en cada orden; sin bucle): `go -C <clon> test -count=1 -tags
mutanteh46 -run '^TestDeshacerH46$' ./internal/evals/ -args -h46-entradas=N` (`ok` en las 15); `git -C <clon> diff
--numstat -- skills/`; `go -C <clon> test -count=1 -v -run '^TestSkillsDelProductoDesdeControl$' ./internal/evals/ |
grep -E -- '--- FAIL|^(ok|FAIL)'`; `git -C <clon> checkout -q -- skills/`.

**Esperado**, en cada una: `TestSkillsDelProductoDesdeControl` en rojo.

### M9 · Entrada 1: `normativa`, «o con su herramienta»

- Deshecha: el `antes` que empieza «una con herramienta. Sigue los pasos en orden, cada operación con `kitlegal …
  --json` desde el `PATH` o, si la»; `2 2 skills/normativa/SKILL.md`.
- Visto: `--- FAIL: TestSkillsDelProductoDesdeControl/normativa/desde_control`, y también
  `referencias-de-un-paquete/de-un-vertical-que-la-base-no-tiene` y
  `referencias-de-un-paquete/de-una-materia-que-no-es-un-vertical`; `FAIL …/internal/evals 0.011s`.

### M10 · Entrada 2: `jurisprudencia`, la orden de leer `consulta_preparada.md`

- Deshecha: «[references/consulta_preparada.md](references/consulta_preparada.md): la orden de cada forma, cómo se dan
  las casillas»…; `2 3 skills/jurisprudencia/SKILL.md`.
- Visto: `--- FAIL: TestSkillsDelProductoDesdeControl/jurisprudencia/desde_control`; `FAIL …/internal/evals 0.011s`.

### M11 · Entrada 3: `jurisprudencia`, la tabla «Cómo se dio | Qué se pide» y su viñeta, de vuelta al cuerpo

- Deshecha: «| Cómo se dio | Qué se pide |»…, delante de su ancla; `9 0 skills/jurisprudencia/SKILL.md`.
- Visto: `--- FAIL: TestSkillsDelProductoDesdeControl/jurisprudencia/desde_control`; `FAIL …/internal/evals 0.010s`.

### M12 · Entrada 4: `jurisprudencia`, «Si se había pedido una sentencia concreta», de vuelta al cuerpo

- Deshecha: «- Si se había pedido una sentencia concreta, da además la referencia en la forma en que se pidió (`ecli`,
  `roj`, o»…, delante de su ancla; `2 0 skills/jurisprudencia/SKILL.md`.
- Visto: `--- FAIL: TestSkillsDelProductoDesdeControl/jurisprudencia/desde_control`; `FAIL …/internal/evals 0.011s`.

### M13 · Entrada 5: `jurisprudencia`, «Las sentencias que el documento menciona», de vuelta al cuerpo

- Deshecha: «- Las sentencias que el documento menciona no se citan con esta forma ni se les pone ECLI: se nombran como
  las»…, delante de su ancla; `2 0 skills/jurisprudencia/SKILL.md`.
- Visto: `--- FAIL: TestSkillsDelProductoDesdeControl/jurisprudencia/desde_control`; `FAIL …/internal/evals 0.010s`.

### M14 · Entrada 6: `jurisprudencia`, los códigos de «Comandos» con «también cuando el documento no es el pedido»

- Deshecha: «entre paréntesis la `data.clase` del resultado que la herramienta marca como error: 0 correcto —también
  cuando el»…; `4 4 skills/jurisprudencia/SKILL.md`.
- Visto: `--- FAIL: TestSkillsDelProductoDesdeControl/jurisprudencia/desde_control`; `FAIL …/internal/evals 0.010s`.

### M15 · Entrada 7: `consulta_preparada.md`, su primer párrafo

- Deshecha: «para una pregunta por materia (paso 6). La línea que va encima de la consulta, lo que se pide en cada caso y
  lo que se»…; `2 2 skills/jurisprudencia/references/consulta_preparada.md`.
- Visto: `--- FAIL: TestSkillsDelProductoDesdeControl/jurisprudencia/desde_control`; `FAIL …/internal/evals 0.011s`.

### M16 · Entrada 8: `consulta_preparada.md`, sin «Lo que se pide»

- Deshecha: el `despues` que empieza «sentencias que le interesen y las adjunte, y sigue con lo que traiga (pasos 3 a
  5).» vuelve a esa sola línea; `0 14 skills/jurisprudencia/references/consulta_preparada.md`.
- Visto: `--- FAIL: TestSkillsDelProductoDesdeControl/jurisprudencia/desde_control`; `FAIL …/internal/evals 0.011s`.

### M17 · Entrada 9: `documento_aportado.md`, sin «sale con 0 también cuando el documento no es el pedido»

- Deshecha: vuelve «si no lo es, un hallazgo que dice qué difiere (`hallazgos`).»; `1 2
  skills/jurisprudencia/references/documento_aportado.md`.
- Visto: `--- FAIL: TestSkillsDelProductoDesdeControl/jurisprudencia/desde_control`; `FAIL …/internal/evals 0.011s`.

### M18 · Entrada 10: `documento_aportado.md`, sin «La referencia pedida y las sentencias que menciona»

- Deshecha: vuelve «documento completo. Sobre un texto sin ficha puedes responder, sin atribuirlo a una sentencia
  identificada.» sin la sección detrás; `0 7 skills/jurisprudencia/references/documento_aportado.md`.
- Visto: `--- FAIL: TestSkillsDelProductoDesdeControl/jurisprudencia/desde_control`; `FAIL …/internal/evals 0.011s`.

### M19 · Entrada 11: `plazos-y-territorio`, la orden de leer `municipios.md`

- Deshecha: «Antes de la primera orden o llamada, lee [municipios.md](references/municipios.md): qué hacer si falta el
  municipio o»…; `2 3 skills/plazos-y-territorio/SKILL.md`.
- Visto: `--- FAIL: TestSkillsDelProductoDesdeControl/plazos-y-territorio/desde_control`, y también
  `referencias-de-un-paquete/un-byte-en-una-de-un-vertical-que-la-base-tiene`; `FAIL …/internal/evals 0.010s`.

### M20 · Entrada 12: `plazos-y-territorio`, «Cada dato como lo da `data`…» en el cuerpo

- Deshecha: «Cada dato como lo da `data`; un boletín, por su `codigo` y su `nombre` exactos; la cobertura, con la clave
  y el valor»…; `2 1 skills/plazos-y-territorio/SKILL.md`.
- Visto: `--- FAIL: TestSkillsDelProductoDesdeControl/plazos-y-territorio/desde_control` y
  `referencias-de-un-paquete/un-byte-en-una-de-un-vertical-que-la-base-tiene`; `FAIL …/internal/evals 0.010s`.

### M21 · Entrada 13: `plazos-y-territorio`, la herramienta u orden de vuelta en «Comandos»

- Deshecha: «Con la herramienta de su fila si entre las tuyas hay una con ese nombre, sola o tras el prefijo de tu
  agente, como»…; `5 1 skills/plazos-y-territorio/SKILL.md`.
- Visto: `--- FAIL: TestSkillsDelProductoDesdeControl/plazos-y-territorio/desde_control` y
  `referencias-de-un-paquete/un-byte-en-una-de-un-vertical-que-la-base-tiene`; `FAIL …/internal/evals 0.011s`.

### M22 · Entrada 14: `municipios.md`, su primer párrafo

- Deshecha: «El detalle del territorio de un municipio (pasos 1 a 4 de `SKILL.md`): qué hacer si falta el municipio, qué
  trae»…; `3 3 skills/plazos-y-territorio/references/municipios.md`.
- Visto: `--- FAIL: TestSkillsDelProductoDesdeControl/plazos-y-territorio/desde_control` y
  `referencias-de-un-paquete/un-byte-en-una-de-un-vertical-que-la-base-tiene`; `FAIL …/internal/evals 0.010s`.

### M23 · Entrada 15: `municipios.md`, sin «Cómo se presenta cada dato» ni «Herramienta u orden»

- Deshecha: vuelve «- Con cualquier otro código o `clase`, di que no se pudo resolver y no suplas sus datos.» sin las
  dos secciones detrás; `0 14 skills/plazos-y-territorio/references/municipios.md`.
- Visto: `--- FAIL: TestSkillsDelProductoDesdeControl/plazos-y-territorio/desde_control` y
  `referencias-de-un-paquete/un-byte-en-una-de-un-vertical-que-la-base-tiene`; `FAIL …/internal/evals 0.011s`.

Las 15, en rojo; al terminar, `rm <clon>/internal/evals/deshacer_h46_test.go` y `git -C <clon> status --short` sin
salida.

## Las 2 frases de FR-015, de una en una (SC-005)

- Orden: `git -C <clon> diff origin/main HEAD -- internal/skills/remisiones_test.go > $TMPDIR/h46-rem.patch` (2 hunks);
  `awk '/^@@/{h++} h<2'` da el del primero (`h46-rem-1.patch`) y `awk '/^@@/{h++} h!=1'`, el del segundo
  (`h46-rem-2.patch`), un `@@` cada uno; `git -C <clon> apply -R <patch>`; `go -C <clon> test -count=1 -v -run
  '^TestRemisionesDelRepositorio$' ./internal/skills/`; `git -C <clon> checkout -q --
  internal/skills/remisiones_test.go`.
- Esperado: `TestRemisionesDelRepositorio` en rojo con cada una.

### M24 · La frase de `main` de `jurisprudencia`

- Mutante (`git diff -U0`): la frase de §3 («…(references/consulta_preparada.md): qué se pide según cómo se dio la
  referencia, la orden de cada forma, …») vuelve a la de `main` («…(references/consulta_preparada.md): la orden de cada
  forma, cómo se dan las casillas y qué hacer con el equivalente que devuelva la operación.»).
- Visto: `--- FAIL: TestRemisionesDelRepositorio`; `FAIL`.

### M25 · La frase de `main` de `plazos-y-territorio`

- Mutante (`git diff -U0`): la frase de §3 («…(references/municipios.md): cómo se pide cada operación, qué hacer si falta
  el municipio …») vuelve a la de `main` («…(references/municipios.md): qué hacer si falta el municipio o no sale con 0,
  qué trae `data`, la plantilla del territorio y qué dice su cobertura.»).
- Visto: `--- FAIL: TestRemisionesDelRepositorio`; `FAIL`.

## Los guiones de H41 y H41.1 con la línea de `main` (FR-016)

### M26 · La orden de `main` en `h41-1-skills-instaladas.txtar`

- Orden: `git -C <clon> show origin/main:internal/app/testdata/script/h41-1-skills-instaladas.txtar >
  <clon>/internal/app/testdata/script/h41-1-skills-instaladas.txtar` (`1 1`, la única línea que T003 cambió); `go -C
  <clon> test -count=1 -v -run '^TestEntregaDelHito$/^h41-1-skills-instaladas$' ./internal/app/`.
- Esperado: `h41-1-skills-instaladas` en `--- FAIL`.
- Visto: `--- FAIL: TestEntregaDelHito/h41-1-skills-instaladas`, por `FAIL:
  testdata/script/h41-1-skills-instaladas.txtar:21: no match for "^## Protocolo\\n\\nAntes de la primera orden o
  llamada, lee \\[municipios\\.md\\]…`; `FAIL …/internal/app 0.683s`.

### M27 · La línea de `main` en `h41-skills-instaladas.txtar`

- Orden: `git -C <clon> show origin/main:internal/app/testdata/script/h41-skills-instaladas.txtar >
  <clon>/internal/app/testdata/script/h41-skills-instaladas.txtar` (`1 2`: la línea de `main` en lugar de las dos de
  §4); `go -C <clon> test -count=1 -v -run '^TestEntregaDelHito$/^h41-skills-instaladas$' ./internal/app/`.
- Esperado: `h41-skills-instaladas` en `--- FAIL`.
- Visto: `--- FAIL: TestEntregaDelHito/h41-skills-instaladas`, por `FAIL: testdata/script/h41-skills-instaladas.txtar:47:
  no match for `(?s)Antes de pedir una u otra, lee\s+\[references/consulta_preparada\.md\]…\| Cómo se…`; `FAIL
  …/internal/app 0.686s`.

## La suite sobre los `skills/` de `main` (SC-006)

### M28 · Los dos guiones `h46-`, sobre los ficheros de skill de `main`

- Orden: como el workflow aún no los ha activado, se copian en el clon como los guiones de `TestEntregaDelHito`: `cp
  <clon>/specs/044-h46-los-cuerpos-de/aceptacion/tabla-de-tres-columnas.txtar
  <clon>/internal/app/testdata/script/h46-tabla-de-tres-columnas.txtar` y lo mismo con `textos-en-su-referencia.txtar` →
  `h46-textos-en-su-referencia.txtar`. **Control**, sobre los `skills/` del hito: `go -C <clon> test -count=1 -v -run
  '^TestEntregaDelHito$/^h46-' ./internal/app/`. **Mutante**: `git -C <clon> checkout -q origin/main -- skills/` (seis
  ficheros: `jurisprudencia/SKILL.md` `28 12`, `consulta_preparada.md` `2 16`, `documento_aportado.md` `1 9`,
  `normativa/SKILL.md` `17 13`, `plazos-y-territorio/SKILL.md` `19 11`, `municipios.md` `3 17`) y la misma orden;
  después, `git -C <clon> checkout -q HEAD -- skills/` y `rm` de las dos copias.
- Esperado: los dos en `--- FAIL`.
- Visto, control: `--- PASS: TestEntregaDelHito/h46-tabla-de-tres-columnas` y `--- PASS:
  TestEntregaDelHito/h46-textos-en-su-referencia`; `ok …/internal/app 0.715s`.
- Visto, mutante: `--- FAIL: TestEntregaDelHito/h46-tabla-de-tres-columnas`, por `FAIL:
  testdata/script/h46-tabla-de-tres-columnas.txtar:20: no match for "^\\| Orden \\| Herramienta \\| Qué hace \\|…`, y
  `--- FAIL: TestEntregaDelHito/h46-textos-en-su-referencia`, por `FAIL:
  testdata/script/h46-textos-en-su-referencia.txtar:25: no match for "`kitlegal`\\s+se\\s+invoca\\s+desde\\s+el…`;
  `FAIL …/internal/app 0.944s`. Son las dos líneas del «rojo primero» de T001 (research P1).

## El árbol de trabajo (Q3, Q4, Q5 y Q8)

En la raíz del repositorio, en la rama `044-h46-los-cuerpos-de` con T001-T005 commiteados (`HEAD` en `7dad83dd`, T005).
Solo leen el árbol.

### Q3 · Los cuerpos, medidos como el gate (SC-002; FR-007; las cifras de FR-017)

- Orden: la de Q3, sin el bucle (una skill por orden), con `LANG=en_US.UTF-8` y `LC_CTYPE="en_US.UTF-8"`.
- Esperado: `normativa 9498 9498`, `plazos-y-territorio 9577 9483` y `jurisprudencia 9473 9473`, las cifras de T004.
- Visto: `9498` y `9498`; `9577` y `9483`; `9473` y `9473`.

### Q4 · `skills/` es el del prototipo (US3.7; FR-010)

- Orden: `git diff --quiet f89093dd -- skills/ && printf 'igual\n'`.
- Esperado y visto: `igual`.

### Q5 · La copia congelada de H41.1 es igual a su guion (US4.4; FR-016)

- Orden: `cmp specs/035-h41-1-plazos-y-territorio/aceptacion/skills-instaladas.txtar
  internal/app/testdata/script/h41-1-skills-instaladas.txtar && printf 'igual\n'`.
- Esperado y visto: `igual`.

### Q8 · Qué tests y guiones cambian, y qué Go de producción (SC-007, SC-008; FR-019)

- `git diff --name-only main HEAD -- '*_test.go' '*.txtar'` → esperado, en este punto, once rutas: las trece de Q8 menos
  `internal/app/testdata/script/h46-tabla-de-tres-columnas.txtar` y `h46-textos-en-su-referencia.txtar`, que añade la
  activación del workflow. Visto: `internal/app/skills_test.go`,
  `internal/app/testdata/script/h41-1-skills-instaladas.txtar`, `internal/app/testdata/script/h41-skills-instaladas.txtar`,
  `internal/evals/laboratorio_familias_test.go`, `internal/skills/comandos_test.go`, `internal/skills/remisiones_test.go`,
  `internal/skills/sincronia_test.go`, `internal/skills/vertical_de_ensayo_test.go`,
  `specs/035-h41-1-plazos-y-territorio/aceptacion/skills-instaladas.txtar`,
  `specs/044-h46-los-cuerpos-de/aceptacion/tabla-de-tres-columnas.txtar` y
  `specs/044-h46-los-cuerpos-de/aceptacion/textos-en-su-referencia.txtar`: las once.
- `git diff --name-only main HEAD -- '*.go' ':!*_test.go'` → esperado `internal/skills/comandos.go` e
  `internal/skills/sincronia.go`; visto, esas dos.
- `git diff --name-only main HEAD -- schemas/ data/ evals/ laboratorio/ .github/ cmd/` → esperado, nada: ninguna eval,
  esquema ni dato cambia (SC-007, SC-008); visto, sin salida.

## El clon al terminar

- Orden de Q9: `rm -rf "$TMPDIR/h46-mut"`.
- Visto: `Permission to use Bash with command rm -rf /tmp/claude/claude-1000/h46-mut has been denied`. La deniega
  `"Bash(rm -rf:*)"` de la lista `deny` de `.claude/settings.json`, y no se sorteó con otra orden: el clon queda en
  `/tmp/claude/claude-1000/h46-mut`, fuera del repositorio, con `git -C <clon> status --short` sin salida (sin mutante
  ni fichero de más: el test de un solo uso y las dos copias de los guiones se borraron con `rm`, uno a uno). Lo borra
  quien tenga ese permiso.

## El árbol al terminar

- `git status --short --untracked-files=no` antes de escribir este registro: solo
  `specs/044-h46-los-cuerpos-de/gates/modelos.jsonl`, `gates/tarea-actual.json` y `gates/tareas-intentos.json`, que el
  workflow ya había cambiado antes de empezar T006; ningún mutante tocó el árbol de trabajo. Después, T006 escribe este
  registro y una línea `[interno]` en `gates/supuestos.md` (la orden de Q9 de los topes, en M1 y M2).
- `make ci`, con este registro escrito, en primer plano: `golangci-lint` con `0 issues.`, 63 líneas `ok` de `go test`
  y ninguna con `FAIL`, `govulncheck` con `No vulnerabilities found.`, el `skills-check` con `ok` en `./internal/app/`,
  `./internal/skills/` y `./internal/evals/`, `goreleaser check` con `1 configuration file(s) validated`, `gitleaks`
  con `no leaks found`, `go mod verify` con `all modules verified` en el módulo y en los cinco de `tools/`, `go mod tidy
  -diff` sin diferencias y `ci: todos los controles en verde`.
