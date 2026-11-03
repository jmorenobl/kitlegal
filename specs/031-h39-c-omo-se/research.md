# Research: H39 · Cómo se organiza una skill para que se active y se lea en los dos modos

Decisiones del plan (D) y lo comprobado en local en esta sesión (V), con dónde. Lo que no se ha podido comprobar sin
red ni sesión con modelo va como supuesto no verificado (S). Nada de lo que sigue se ha medido con un modelo.

## Verificaciones (V)

| Id | Qué | Dónde se comprobó |
|---|---|---|
| V1 | Kong toma la ayuda de la etiqueta `help` del campo (`t.Help = t.Get("help")`) y la interpola al construir el analizador; solo cambian `${var}`, `${var=defecto}` y `$$`; un `$` suelto, como en `$HOME/.agents/skills`, queda igual. Ninguna de las 28 ayudas lleva `${` ni `$$`: la etiqueta y lo que imprime `--help` coinciden carácter a carácter | `~/go/pkg/mod/github.com/alecthomas/kong@v1.16.1/tag.go:277`, `kong.go:229` y `:285`, `interpolate.go:8`; `grep -h 'help:"' internal/app/{boe,cita,grafo,instalacion,plazos,territorio}.go \| grep -c '\${\|\$\$'` → 0 |
| V2 | Las 28 ayudas suman 2373 y la mayor mide 146, las cifras del spec; 12 de los 28 campos no llevan `name:` y Kong los nombra por el campo | mismo `grep`, con `awk '{n+=length($0)…}'` → `28 2373 146`; `grep -vc 'name:"'` → 12 |
| V3 | La entrada de `--describe` es un objeto con una propiedad por campo, cuyo esquema sale de reflejar **el tipo** del campo (`g.reflejar(campo.Type)`): las etiquetas del campo no llegan al reflector. En `--describe` se reflejan los campos del verbo y detrás las ocho banderas globales en una sola llamada (`g.entrada(append(campos, camposVisibles(tipoDeGlobales)...))`), seguida de `anotarBanderas(entrada, campos)`; en la herramienta MCP, solo los del verbo | `internal/cli/describe.go:122-123`, `:188-205`, `:307-309`; `internal/cli/herramienta.go:32-68` |
| V4 | El documento se escribe con `SetEscapeHTML(false)`: `BOE-A-<año>-<número>` sale con `<` y `>` literales | `internal/cli/describe.go:541-553` |
| V5 | `kitlegal boe articulo --describe` sin sus dos argumentos sale con 2 (`expected "<norma> <bloque>"`): el análisis va antes que la descripción. Con `BOE-A-2015-10565 a21` emite 4989 bytes, y `norma`, `bloque` y `fecha` llevan solo `"type": "string"` | `go run ./cmd/kitlegal boe articulo --describe` y `… BOE-A-2015-10565 a21 --describe`, ejecutadas en esta sesión; `internal/app/herramientas_test.go:287-291` lo documenta |
| V6 | El `inputSchema` de cada herramienta es la entrada de `--describe` sin las banderas globales, y lo vigila `TestHerramientasDelServidor` (`compruebaLaHerramienta`) | `internal/app/herramientas_test.go:138`, `:246-285` |
| V7 | La tabla de comandos lee del documento solo título, ayuda del verbo, nombres, obligatoriedad, tipos y `x-banderas`; la `description` de una propiedad se decodifica (`esquemaDescrito.Ayuda`) y no se usa al pintar. Lo que lo garantiza es la comprobación de deriva de `TestSkillsDelRepositorio` | `internal/skills/comandos.go:84-134`, `:395-401`, `:531-566`; el único uso de `.Ayuda` es el del verbo, `:283` (`grep -n '\.Ayuda'`); `internal/app/skills_test.go:88-118` |
| V8 | Fijan bytes de `--describe` con argumentos de ayuda: `schemas/` (siete de los ocho ficheros que salen de `--describe`, `ficherosDeEsquemas`: todos menos `servidor.json`), `internal/app/testdata/plazos/describe.json` (bandera `-actualizar-golden-de-plazos`) y la sección `main-articulos.describe` de `internal/app/testdata/script/h20-fecha-main.txtar`; `schemas/` se regenera con `-actualizar-esquemas` | `git grep -l '"entrada": {'`; `internal/app/plazos_test.go:29`; `internal/app/esquemas_test.go:28`, `:62-71` |
| V9 | `internal/cli/` e `internal/evals/` y `cmd/` son de lo común del job de evals: un cambio en ellos elige todas las skills (FR-053). `laboratorio/` no está en lo común ni en lo que mide: ninguna ruta suya elige ninguna skill | `internal/evals/dependencias.go:36-56`, `:234-260` |
| V10 | El binario empotra `skills/*/SKILL.md` y `skills/*/references/*` y nada más; `skills install --dir <ruta>` instala en esa ruta, «sin hosts» | `skills.go:24`; ayuda de `Dir` en `internal/app/instalacion.go` |
| V11 | `skills.Regenerar(raiz, descripciones)` lista `raiz/skills/` y, sin `data/normas.yaml`, no lee ningún paquete ni falla: sirve para una raíz que solo tiene `skills/` | `internal/skills/sincronia.go:206-247`; `internal/skills/paquetes.go:158-166`; `internal/skills/skill.go:57-100` |
| V12 | Una sesión ve como skills solo las entradas del directorio `Skills` del reparto, enlazadas en su `claude/skills/`, con `CLAUDE_CONFIG_DIR` propio y `--setting-sources user`; `servidor.json` se escribe con la orden `Binario mcp serve`, fija; en el modo herramienta el directorio de `Binario` sale del `PATH` | `internal/evals/sesiones.go:421-473`, `:475`, `:569-620`; `scripts/evals-sesion.sh` |
| V13 | El SDK escribe y lee un mensaje JSON por línea (stdio delimitado por saltos de línea); `instructions` es un campo del resultado de `initialize` | `~/go/pkg/mod/github.com/modelcontextprotocol/go-sdk@v1.8.0/mcp/transport.go:450-806`, `mcp/protocol.go:1089` |
| V14 | R7: solo `internal/mcp` importa el SDK de MCP; `mcptest` da las líneas en bruto que escribe el servidor (`Sesion.Lineas`) | `internal/arch_test.go:33-34`; `internal/mcp/mcptest/sesion.go:209-212` |
| V15 | `cmd/` está exento de R4 (`os.Exit`) y R5 (`os.Stdout`, `os.Stderr`); `cmd/empaquetar` es un programa del repositorio que no se distribuye ni se instala | `.golangci.yml:461-468`; `cmd/empaquetar/main.go`; `.goreleaser.yaml` (`main: ./cmd/kitlegal`) |
| V16 | La política de las sesiones de un run veda `make evals`, `make evals-sondeo`, `scripts/evals[a-z-]*.sh` y `TestSondeo` seguido de algo que no es letra: veda `scripts/evals-laboratorio.sh`, y no `make evals-laboratorio` ni `TestSondeoDelLaboratorio`. El guion del llavero no abre nada en un paso del workflow (`KITLEGAL_PASO_DE_WORKFLOW`) y la garantía es el sandbox (ADR 0032) | `scripts/workflow/politica-paso.sh:51-55`; `scripts/evals-sondeo-llavero.sh:21-26` |
| V17 | El transcript es stream-json: `system/init` con `model` y `claude_code_version`; cada `assistant` con `message.id` y `message.content`; los resultados llegan en `user` con bloques `tool_result` y su `tool_use_id`. `LeerSesion` da modelo, versión, respuesta, si terminó y su motivo | `internal/evals/testdata/sesiones/leer-sesion/activada/sesion.jsonl`; `internal/evals/sesion.go:243-331`, `:580-681` |
| V18 | `InterpretarInvocacion(argv)` dice si un argv es de un applet y si consulta (`Consulta` falso con ayuda, `--describe` o `--dry-run`); `NombresDeHerramientas` da las herramientas del registro sin `skills` ni `mcp` | `internal/evals/trazas.go:232-332`; `internal/app/herramientas.go:34-89` |
| V19 | Hay grabadas, para las evals 09, 04, 07 y 10 de `normativa`, metadatos, índice y bloque de BOE-A-1978-31229 a140, BOE-A-2003-23186 a66, BOE-A-2015-10566 a25 y BOE-A-2015-11430 a38 | `ls testdata/evals/boe.legislacion-consolidada/`; `evals/normativa/{04,07,09,10}-*.yaml` |
| V20 | Respuesta real de una skill del producto a la pregunta «¿Qué dice el artículo 140 de la Constitución?»: la sesión `09-constitucion-articulo-140-claude-sonnet-5-5-01` del informe versionado de H32, activada y terminada, de 1390 caracteres | `jq` sobre `specs/030-h32-arquitectura-de-skills/gates/evals/normativa.json` |
| V21 | Wilson al 95 % (z = 1,96): 0/40 → 0,0–8,8; 20/40 → 35,2–64,8; 40/40 → 91,2–100,0; 31/40 → 62,5–87,7; 38/40 → 83,5–98,6; 3/40 → 2,6–19,9 | cálculo con `awk` en esta sesión |
| V22 | `claude --version` escribe `<versión> (Claude Code)` | registrado en el repositorio: `specs/006-h5-skill-boe-legislacion/research.md` V1 y `specs/013-h7-3-el-umbral-de/research.md` V1 (2.1.270). No se ha ejecutado en esta sesión: ver S1 |

## Supuestos no verificados (S)

- **S1.** La versión instalada de Claude Code sigue escribiendo `<versión> (Claude Code)` con `--version` (V22 es de
  2.1.270; esta sesión no puede ejecutar `claude`). El arnés toma el primer campo y lo compara entero con el pedido;
  si el formato cambiara, el arnés diría que la versión no coincide y no abriría ninguna sesión: falla del lado seguro.
- **S2.** Claude Code escribe cada bloque de una misma respuesta del modelo como un mensaje `assistant` aparte con el
  mismo `message.id`. No hay en el repositorio ninguna transcripción real (las 50 `sesion.jsonl` son sintéticas). La
  regla del turno de D9 no depende de ello: separa por el resultado de una herramienta y solo usa el `message.id` para
  no partir una misma respuesta.
- **S3.** Con la orden del servidor cambiada en `servidor.json` (D7), Claude Code arranca el programa que se le dé
  igual que `kitlegal mcp serve`: es una orden con argumentos, como la de hoy.

## Decisiones (D)

### D1. Dónde nace la `description` de un parámetro (FR-001 a FR-003)

**Decisión.** En `internal/cli`, una función `describirArgumentos(entrada, campos)` que, tras reflejar la entrada, pone
en la propiedad de cada campo **del verbo** con etiqueta `help` no vacía su valor tal cual; la llaman `esquema()` (el
documento de `--describe`, detrás de `anotarBanderas`) y `EsquemasDeHerramienta` (la herramienta MCP). Las banderas
globales no se tocan (Fuera de alcance), y nada más del esquema cambia.
**Por qué.** Es el mismo patrón que `anotarBanderas` (V3), un solo sitio para los dos esquemas, y deja fuera las
globales sin distinguirlas después. La ayuda es la etiqueta: coincide con lo que imprime `--help` (V1).
**Alternativas.** (a) El reflector de invopop con `jsonschema_description` o `FieldNameTag`: no ve las etiquetas del
campo, porque se refleja su tipo (V3). (b) Leer la ayuda ya interpolada del modelo de Kong: obliga a construir el
analizador para describir, y hoy da lo mismo (V1). (c) Ponerla también a las globales: fuera de alcance.

### D2. El control de SC-001 compara con lo que muestra `--help`

**Decisión.** `TestAyudasDeLosArgumentos` (`internal/app`) recorre los verbos de `RegistroDeProduccion`; de cada uno
toma la ayuda de cada argumento y bandera propios del modelo que Kong construye con la gramática del applet (la que
imprime `--help`), y la compara con la `description` de su propiedad en el documento de `--describe` (orden en proceso,
con un valor por cada argumento obligatorio, V5) y en el `inputSchema` de la herramienta MCP (servidor en proceso con
`mcptest`). Falla con el primer argumento que no cumple, si una propiedad sin ayuda gana `description` o si una global
la lleva, y si encuentra menos de las 28 ayudas de hoy (no pasa en vacío).
**Por qué.** El oráculo no es el código que se prueba (V1). `TestHerramientasDelServidor` ya exige que los dos esquemas
sean el mismo (V6); el test nuevo mide el umbral en las dos formas que promete FR-002.
**Alternativa.** Leer las etiquetas por reflexión en el test: repetiría la regla de nombres de `internal/cli`.

### D3. Ficheros que fijan bytes de `--describe` y se regeneran (FR-004)

**Decisión.** Una tarea `[datos]` regenera `schemas/` con `-actualizar-esquemas`, `internal/app/testdata/plazos/describe.json`
con `-actualizar-golden-de-plazos` y la sección `main-articulos.describe` de `h20-fecha-main.txtar` con la salida del
binario (V8), y actualiza los tests de `internal/cli` que fijan el esquema de un struct con ayuda. `skills/` no se toca:
la deriva de `TestSkillsDelRepositorio` es el control (V7).
**Por qué.** Son los únicos ficheros que fijan esos bytes (V8); lo demás de `make ci` los compara con lo emitido.

### D4. Dónde vive el laboratorio (FR-010)

**Decisión.** Un directorio de primer nivel `laboratorio/` con la declaración de variantes, las cuatro evals y una
copia completa de la skill (y, en `entrada`, de las dos) por variante (contracts/laboratorio.md §1). El código que lo
lee, el arnés y el guion, en `internal/evals` (ficheros `laboratorio*.go`), y el programa de sus dos órdenes, en
`cmd/laboratorio`.
**Por qué.** Fuera de `skills/` no se empotra ni se instala (V10); fuera de `evals/` el job no lo lee; y `laboratorio/`
no es de lo común ni de lo que mide, así que un cambio solo del laboratorio no elige ninguna skill del producto (V9).
El código va a `internal/evals` porque reutiliza su reparto de sesiones, su preparación y su lector sin exportarlos.
**Alternativas.** (a) `internal/evals/laboratorio/` o `internal/evals/testdata/`: son de lo común, y cada cambio del
laboratorio mediría todas las skills del producto. (b) `testdata/laboratorio/`: `testdata/` es de grabaciones y está
protegido. (c) Un subpaquete de Go con el código: obliga a exportar el repartidor y la preparación de las sesiones.

### D5. Variantes como copias completas, comparadas por partes (FR-020 a FR-028)

**Decisión.** Cada variante es un directorio con lo que se instala tal cual (`skills/<nombre>/…`) y, en `instrucciones`
y `parametro`, un `.txt` con el texto de la sesión. La declaración dice, por variante, sus modos y las partes que cambia
de un vocabulario cerrado (`descripcion`, `cuerpo`, `referencia`, `entrada`, `instrucciones`, `parametro`; data-model §2).
`TestVariantesDelLaboratorio` compara cada parte con la de control: igual byte a byte si no está declarada, distinta si
lo está, y ningún fichero fuera de las partes.
**Por qué.** Lo instalado es lo comparado; el test es FR-028 literal.
**Alternativa.** Guardar solo diferencias y componer al instalar: añade una composición que también habría que probar.

### D6. La regla testigo (FR-013, FR-014)

**Decisión.** La respuesta termina con una última línea, sola y sin formato, que dice exactamente `Sello de lectura:
KL-7Q4`. El guion: la última línea no vacía de la respuesta, sin espacios a los lados, es esa cadena. Sus formas de no
cumplir: no aparece; aparece y no es la última; es la última con formato u otro texto (contracts/laboratorio.md §3). La
respuesta del producto de V20 no la cumple.
**Por qué.** Es inventada (un código que el modelo no puede saber), tiene forma y un guion la decide sin modelo (ADR
0037). No depende de la eval ni de la herramienta, así que es la misma en las ocho variantes.
**Alternativa.** Un sello que dependa del bloque leído: obliga al guion a saber qué bloque pide cada eval sin ganar nada.

### D7. El servidor de las variantes `instrucciones` y `parametro` (FR-021, FR-024, FR-029)

**Decisión.** `laboratorio servidor` (`cmd/laboratorio`), un filtro de líneas: arranca `kitlegal mcp serve` como
proceso hijo con su entorno, le pasa la entrada sin tocarla y, de su salida, reescribe solo dos mensajes —en la
respuesta a `initialize`, `result.instructions`; en la de `tools/list`, la `description` del parámetro de la
herramienta declarada— y deja pasar los demás byte a byte. `SesionesAEjecutar` gana el campo `Servidor` (la orden de
`servidor.json`; vacío, la de hoy), y el arnés lo da solo en esas dos variantes.
**Por qué.** El binario que se distribuye no gana ninguna bandera, variable ni orden (FR-029): el cambio vive fuera, en
un programa del repositorio como `cmd/empaquetar` (V15). Un mensaje por línea es el transporte del SDK (V13), y el
filtro no importa el SDK, que R7 reserva a `internal/mcp` (V14). El `kitlegal` de la sesión sigue saliendo del `PATH`
del modo herramienta, porque `Binario` no cambia (V12).
**Alternativas.** (a) Una bandera o variable oculta en `kitlegal`: la prohíbe FR-029. (b) Construir otro `kitlegal`
con `go build -overlay` y el texto sustituido en el fuente: reescribir Go fuente es frágil y cambiaría `--help`. (c)
Volver a servir las herramientas con el SDK: lo prohíbe R7 y es más código.

### D8. El arnés del sondeo (FR-030 a FR-033)

**Decisión.** `make evals-laboratorio` → `scripts/evals-sondeo-llavero.sh laboratorio …` (el guion del llavero gana un
primer argumento de una lista cerrada, `sondeo` o `laboratorio`, que elige el guion que ejecuta con la credencial) →
`scripts/evals-laboratorio.sh` → `TestSondeoDelLaboratorio` (`-tags evals`) → `sondearElLaboratorio`. Comprueba, antes
de abrir nada y con todos los errores juntos, los argumentos (contracts/arnes.md §2), la credencial y `claude
--version`; prepara en un temporal el binario, el programa del laboratorio y una copia de las skills de cada variante;
y, por cada variante y modo pedidos, en el orden de la declaración, reparte con `ejecutarSesiones` las n/4 sesiones de
cada eval en `<DIRECTORIO>/<variante>/<modo>/<eval>-<nn>/`. `DIRECTORIO` no puede existir; el arnés lo crea con su
`sondeo.json` (modelo y versión). Tras un límite de uso de la cuenta no abre más.
**Por qué.** Reutiliza la toma de la credencial sin copiarla, el reparto, la preparación de la caché y la grabación de
hoy (ADR 0043). La ruta de cada sesión dice su variante, su modo y su eval (FR-033), y `DIRECTORIO` nuevo evita que un
segundo lanzamiento pise sesiones.
**Alternativas.** (a) Un guion del llavero aparte: duplicaría la parte que toca la credencial. (b) Un solo reparto con
todas las sesiones: el directorio de skills y el servidor son del reparto, no de la sesión. (c) `CONCURRENCIA` con
valor por defecto: no hay ninguno que no sea el de otra skill; es obligatoria.

### D9. El guion que cuenta (FR-040 a FR-043)

**Decisión.** `laboratorio tabla <DIRECTORIO>` (`make laboratorio-tabla DIRECTORIO=…`): lee `DIRECTORIO` como un
sondeo si tiene `sondeo.json`, o cada subdirectorio que lo tenga (así se juntan los lanzamientos de FR-050); de cada
sesión, `LeerSesion` (terminada, respuesta) y un lector propio de turnos del transcript (D10); agrupa en filas por
variante, modo, modelo y versión; escribe en la salida estándar la tabla y el detalle en Markdown (contracts/tabla.md).
Una transcripción que no sabe leer es un defecto `inesperado`: sale con 1 sin tabla.
**Por qué.** Sin red ni modelo (FR-040); el modelo y la versión van en la fila y no hay que decidir qué hacer si dos
lanzamientos difieren.

### D10. Qué es un turno y qué cuenta (FR-041)

**Decisión.** Turno nuevo: el primer mensaje `assistant` y cada `assistant` que llega después de un `user` con algún
`tool_result` y con un `message.id` distinto del `assistant` anterior. Activar: `tool_use` de `Skill` con
`input.skill` igual al nombre. Primera herramienta en modo herramienta: el primer `tool_use` cuyo nombre, o lo que sigue
a su último `__`, está en `NombresDeHerramientas`. En modo orden: el primer `tool_use` de `Bash` cuya orden, partida
en órdenes simples por `|`, `||`, `&&`, `;`, `&` y saltos de línea y en palabras con las comillas y la barra de sh,
tiene una orden simple cuya primera palabra que no es una asignación es `kitlegal` (o una ruta que termina en
`/kitlegal`) y cuyo argv, para `InterpretarInvocacion`, es de un applet, consulta y su `<applet>_<verbo>` está en
`NombresDeHerramientas` (V18). Abrir la referencia: `Read` cuyo `file_path` termina en
`<skill>/references/<fichero>` o `Bash` cuya orden contiene `references/<fichero>`.
**Por qué.** Es la definición del spec; el `message.id` guarda contra partir una respuesta si un resultado llega
mientras se escriben sus bloques (S2). El lector de órdenes cubre lo que el spec enumera (orden, tubería, secuencia).
**Alternativas.** (a) Solo el `message.id`: depende de S2. (b) Solo el resultado de una herramienta: partiría una
respuesta en el caso que guarda (a). (c) Contar `kitlegal` dentro de `$( )` o de una subshell: el spec no lo enumera.

### D11. Formato de las cifras

**Decisión.** «`<k> de <n> · <p> % [<lo>–<hi> %]`», porcentajes con un decimal y coma decimal, extremos de Wilson al
95 % recortados a [0, 100] antes de redondear; la mediana del turno, entera o con «,5», y los extremos entre paréntesis.
**Por qué.** Es la forma de los ejemplos del spec (SC-006, V21).

### D12. Aceptación e2e

**Decisión.** Un guion testscript, `aceptacion/describe.txtar` (activado como `h39-describe.txtar`), para la historia 1:
lo que `--describe` da en `boe articulo`, `plazos calcular` y `skills install`. FR-002 lo comprueban los tests de Go
(D2), porque la orden `mcp` del arnés e2e escribe de cada herramienta solo su nombre y si es de solo lectura
(`internal/app/e2e_test.go:2362-2381`), y la primera tarea solo escribe guiones. Las historias 2 y 3 no tienen
comportamiento visible del binario `kitlegal`: su aceptación son los tests de `make ci` de SC-003 a SC-007; la 4 la hace
la persona antes de fusionar (FR-050 a FR-052).

### D13. Lo que el cierre no arregla (FR-053)

**Decisión.** Se anota en `gates/supuestos.md` que una reparación del cierre no toca `skills/` para los rojos de partida
y que el informe da sus cifras; y que la propuesta queda pendiente del sondeo y del ADR (FR-052), para que el informe
final lo diga (`scripts/workflow/informe.sh`, sección «Supuestos y pendientes»).

## Fuentes externas

Ninguna. El hito no graba nada (FR-012): las sesiones del sondeo se sirven de `testdata/evals/` (V19), y el guion lee
las transcripciones que deja el arnés. No hay manifiesto `grabaciones.json` ni test `TestGrabar*`, ni fila nueva en
`docs/SOURCES.md`.
