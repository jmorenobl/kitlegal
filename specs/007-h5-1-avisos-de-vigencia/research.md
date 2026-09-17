# Research: H5.1 · Avisos de vigencia en las evals

Fase 0 del plan. Resuelve lo que el spec deja al plan —la tolerancia de la forma fija al detalle (FR-032), la redacción
literal de la pregunta (FR-051), cómo se exportan las etiquetas, dónde viven las comprobaciones mecánicas y el texto de
los motivos (checklist de requisitos, «No implementation details»)— y lo que solo aparece al planificar: el orden que
deja `make ci` en verde tarea a tarea con las tareas `[datos]` separadas, y una grabación que pida a la fuente solo lo
que falta. Cada decisión lleva la alternativa rechazada y por qué (constitución, «Criterio de decisión autónoma»,
punto 3). No queda ningún `NEEDS CLARIFICATION`.

**Cómo se verificó.** Toda afirmación sobre el comportamiento de una herramienta o dependencia se comprobó en local
contra su código, su documentación (`go doc`, `--help`) o un experimento, y la tabla dice dónde. Los experimentos se
hicieron el 2026-09-16 sobre un **clon desechable** del repositorio en `/tmp/kitlegal-plan-h51` (commit `2d2efb8`, la
base de la rama), con prototipos de usar y tirar de lo que el plan decide, y en un segundo clon desechable de la misma
base, `/tmp/kitlegal-plan-h51-r2`, sin prototipos, donde se aplicó el texto literal de `SKILL.md` que fija el contrato
de la forma fija §3 (V33); los clones no forman parte del repositorio y ningún experimento tocó el árbol de trabajo, el
índice ni el historial. Lo que depende de la plataforma remota o de la red de la fuente no se puede verificar en local
y figura en §S como **supuesto no verificado**, con quién lo comprueba y cuándo.

## Tabla de verificación

| # | Afirmación | Dónde se comprobó |
|---|---|---|
| V1 | Con `avisos: {type: array, minItems: 1, items: {$ref: "#/$defs/codigo-de-aviso"}}` y el enumerado en `$defs`, `LeerEval` rechaza un código desconocido con `01-lpac-articulo-21.yaml: avisos/0, línea 11: value must be one of 'consolidacion-no-finalizada', 'derogada', 'vigencia-agotada'` | Experimento en el clon: `LeerEval` (`internal/evals/formato.go:111-125`) con el esquema de D4, `santhosh-tekuri/jsonschema/v6 v6.0.3` (`go.mod:11`) |
| V2 | `avisos: []` da `avisos, línea 10: minItems: got 0, want 1`, igual que `citas: []` (`citas, línea 7: minItems: got 0, want 1`); un código repetido se acepta, igual que una cita repetida, porque ninguno de los dos lleva `uniqueItems` | Mismo experimento; `schemas/eval.yaml.json:23-26` (`citas` sin `uniqueItems`) |
| V3 | `avisos` en una eval con `activa: false` da `línea 1: 'not' failed` si la rama `else` añade `{required: [avisos]}` al `anyOf`; con `activa: true` e `informativa: true` se acepta | Mismo experimento; `schemas/eval.yaml.json:28-30` |
| V4 | En el esquema compilado, los códigos de `avisos` están en `Properties["avisos"].Items2020.Ref.Enum.Values` (`items` de 2020-12 va a `Items2020`, no a `Items`, y `$ref` a `Ref`) | `go doc github.com/santhosh-tekuri/jsonschema/v6.Schema` y `…v6.Enum`; experimento en el clon (`Items2020: true`, `Items: <nil>`, `Ref` con `Enum.Values` = los tres códigos) |
| V5 | Un esquema con `avisos` y un `Eval` sin ese campo no rompen nada: `(*yaml.Node).Decode` ignora la clave que el tipo no tiene | `internal/skills/esquemas.go:166-169`; experimento: `go test ./internal/evals/ ./internal/app/ ./internal/skills/` en verde en el clon con solo el esquema cambiado |
| V6 | `(?i)` compara sin distinguir mayúsculas con el plegado de Unicode (también `ſ` con `s`); `\s` es solo `[\t\n\f\r ]`, así que los blancos de Unicode se escriben `[\t\p{Zs}]`, que incluye U+00A0 y U+202F; `regexp.QuoteMeta` deja literal una palabra | `go doc regexp/syntax` (líneas «i case-insensitive», «\s whitespace (== [\t\n\f\r ])», «\p{Foo} Unicode character class»); experimento con 29 textos de D2 en el clon |
| V7 | La gramática de D2 encuentra la forma con U+FE0F, U+FE0E, `**…**`, `⚠ **…**:`, `_…_`, `__…__`, espacios, tabuladores, U+00A0/U+202F, minúsculas, sin blanco tras `⚠`, tras `>` o tras un salto de línea; y **no** la encuentra con una palabra de más, de menos o pegada (`PARCIALMENTE`, `NORMADEROGADA`, `DEROGADAS`, `NO NORMA`), sin `⚠`, sin dos puntos, con los dos puntos de ancho completo `：`, con dos selectores seguidos, con un salto de línea dentro, con `~~` dentro, con `🚨` o con otra redacción («Esta ley fue derogada.», «La Ley 30/1992 sigue en vigor.») | Experimento en el clon (salida registrada en la sesión de planificación; los casos son los de `TestExtraerAvisos`, contrato de la forma fija §4) |
| V8 | `encoding/json/v2` codifica un slice nulo como `[]` salvo con `FormatNilSliceAsNull`, así que `avisos_encontrados` y `avisos_ausentes` salen como listas vacías sin inicializarlos | `go doc encoding/json/v2 Marshal` («By default, a nil slice is encoded as an empty JSON array») y `go doc encoding/json/v2 FormatNilSliceAsNull`; experimento (`{"l":[]}`); `internal/evals/informe.go:296` |
| V9 | Exportar las etiquetas y componer las frases con ellas no cambia ningún byte de salida: `TestAvisosDe` compara con las frases literales y los golden de `articulo` y `metadatos` de `BOE-A-1992-26318` pasan | Prototipo en el clon de `internal/source/boe/avisos.go`; `go test ./internal/source/boe/` en verde; `internal/source/boe/avisos_test.go:11-20` |
| V10 | El prototipo de `EtiquetasDeAviso`, `ExtraerAvisos` y las dos comprobaciones pasa el linter fijado sin ninguna exclusión (solo un aviso de alineación de `gofumpt` del borrador) | `go tool -modfile=tools/golangci-lint/go.mod golangci-lint run ./internal/evals/... ./internal/source/boe/...` en el clon |
| V11 | La grabación sobrescribe el fichero de una petición que ya estaba grabada («la misma petición se regraba») | `internal/httpx/grabar.go:216-256` (`comprobarQueEsLaMisma`) |
| V12 | `TestGrabarEvals` siembra cada consulta solo desde las grabaciones de H4: con el manifiesto entero, pediría a la fuente y volvería a grabar lo que grabó H5 para el manifiesto, que son treinta peticiones —las nueve búsquedas de las entradas 2 a 10 y las veintiuna consultas de norma de las entradas 4 a 10 (metadatos, índice y bloque de siete normas)—, más el `robots.txt` que `httpx` pide en cualquier caso antes de la primera petición real. La búsqueda de la entrada 1 (`procedimiento administrativo común`) y los metadatos, el índice y el bloque de la LPAC, la LCSP y la LRBRL están en las grabaciones de H4, así que no se piden; de la entrada nueva, H4 ya tiene la búsqueda, los metadatos y el bloque `a42` (V18) | `internal/evals/grabacion_test.go:189-193` (`Preparar(g.dirCache, []string{GrabacionesDeH4}, …)`) y `:96-113`; `ls testdata/evals/boe.legislacion-consolidada/` (21 ficheros de norma de 7 normas, 10 de búsqueda y `robots.txt`) y `ls internal/source/boe/testdata/boe.legislacion-consolidada/` (la LPAC, la LCSP y la LRBRL con metadatos, índice y bloque; 2 búsquedas); consulta de cada búsqueda grabada leída de `.peticion.url`: las 10 de H5 son las de las entradas 2 a 10 y `texto refundido haciendas locales`, ajena al manifiesto; las 2 de H4, `procedimiento administrativo común` y `zzqxkwvjh` |
| V13 | Con el cliente de `httpx.New`, `robots.txt` pasa por el escalón de grabación (va por encima de él en la cadena), y `httpx.Replay` no consulta `robots.txt` («una grabación suya en el directorio queda sin usar») | `internal/httpx/cliente.go:262-289` y `:300-312` |
| V14 | Solo con la entrada nueva del manifiesto, `TestIdentificadoresDeLasNormas/norma-en-la-busqueda-de-otra-entrada` falla («An error is expected but got nil»): elige como «otra» el primer resultado de la búsqueda de la LPAC que no está en la tabla, que es `BOE-A-1992-26318`, y la entrada nueva lo resuelve. Con la premisa de D12, la misma prueba pasa con y sin la entrada, y con la norma ya en la tabla | Experimento en el clon; `internal/evals/grabaciones_test.go:319-321`, `:603-624` |
| V15 | Con la entrada del manifiesto, la norma en `data/normas.yaml`, `references/normas.md` regenerado y la eval 18, `TestEvalsDelRepositorio` solo falla en `grabado` y solo por el índice: `18-lrjpac-norma-derogada.yaml: la norma BOE-A-1992-26318 sin indice: «boe indice BOE-A-1992-26318» terminó con código 1 … falta el fichero …_BOE-A-1992-26318_texto_indice.json`; `formato`, `conjunto` y `normas-conocidas` pasan con 18 evals y 10 positivas que deciden, y `TestIdentificadoresDeLasNormas`, `TestNormasDelRepositorio` y `TestSkillsDelRepositorio` pasan | Experimento en el clon |
| V16 | La búsqueda grabada `procedimiento administrativo común` da `BOE-A-1992-26318` como primer resultado, con título «Ley 30/1992, de 26 de noviembre, de Régimen Jurídico de las Administraciones Públicas y del Procedimiento Administrativo Común.» y rango `Ley`; ningún otro título empieza por `Ley 30/1992,` | `internal/source/boe/testdata/golden/buscar-procedimiento-administrativo-comun.json` (`jq '.data'`) |
| V17 | Los metadatos y el bloque `a42` de `BOE-A-1992-26318` emiten los avisos `derogada` («⚠ NORMA DEROGADA: esta norma ha sido derogada.») y `vigencia-agotada` («⚠ VIGENCIA AGOTADA: esta norma ya no está en vigor.»); el texto del bloque empieza por «Artículo 42. Obligación de resolver.» | `internal/source/boe/testdata/golden/metadatos-BOE-A-1992-26318.json` y `articulo-BOE-A-1992-26318-a42.json` |
| V18 | H4 grabó los metadatos y el bloque `a42` de `BOE-A-1992-26318`, no su índice; H5 no grabó nada de esa norma | `ls internal/source/boe/testdata/boe.legislacion-consolidada/` y `ls testdata/evals/boe.legislacion-consolidada/` |
| V19 | `TestInforme` no compara filas de la tabla «Sesiones» de `informe.md`, así que añadir dos columnas no rompe sus casos; y exige que cada directorio de `testdata/sesiones/informe/` sea un caso de su tabla | `internal/evals/informe_test.go:154-160` y búsqueda de `Sesiones` en el fichero |
| V20 | La respuesta de la sesión `01-lpac-articulo-21` del caso `aprobado` aparece dos veces en su `sesion.jsonl` (el mensaje del asistente y el `result`), igual que su cita; `os.CopyFS` copia un árbol a un directorio nuevo y ya se usa en el paquete | `grep -c` y `grep -o … \| wc -l` sobre `internal/evals/testdata/sesiones/informe/aprobado/sesiones/01-lpac-articulo-21/sesion.jsonl`; `go doc os.CopyFS`; `internal/evals/preparar_test.go:516` |
| V21 | Las evals informativas solo las abre el modelo que decide y cada sesión se llama `<eval sin .yaml>-<modelo>-<nn>`: con 18 evals (5+1 informativas) y 3 repeticiones, el plan abre 54 sesiones de `claude-sonnet-5` y 36 de `claude-haiku-4-5-20251001`, 90 en total | `internal/evals/plan.go:117-176`; `.github/workflows/evals.yml:76-81` |
| V22 | El job arranca en `opened`, `reopened` y `labeled`; su job `cambios` marca `coincide=si` si la propuesta toca `skills/`, `evals/`, `data/`, `internal/source/boe/`, `internal/evals/` o `schemas/eval.yaml.json`, y evalúa `github.event.pull_request.head.sha` | `.github/workflows/evals.yml:18-19`, `:50-58`, `:85` |
| V23 | `scripts/evals.sh` imprime `informe.md` e `informe.json` enteros, cada uno entre `--- inicio de <f> ---` y `--- fin de <f> ---` | `scripts/evals.sh:198-205` (el bucle, `:201-205`) |
| V24 | `gh run view --log` da cada línea como `<job>\t<paso>\t<instante> <contenido>`, con las líneas vacías como `<instante> ` (espacio final); la tubería `sed -n '/inicio/,/fin/p' \| sed '1d;$d' \| cut -f3- \| sed -E 's/^[^ ]+ //'` devuelve `informe.json` legible por `jq` | Registro real de H5 guardado en `specs/006-h5-skill-boe-legislacion/gates/evals-cierre.md:409-416` y `:960-1400` (`cat -et`); experimento: la tubería sobre esas líneas da `{"veredicto":"aprobado","commit":"5c6c552…","red":[],"evals":12}` |
| V25 | `gh` 2.100.0 tiene `run list --branch --event --commit --limit --json` con los campos `databaseId`, `workflowName`, `headSha`, `status`, `conclusion`, `createdAt`, `url`, `event`; `run view --log --json`; `run watch --exit-status --interval`; `pr view --json number,state,baseRefName,headRefName,headRefOid,url,labels`; `pr create --base --head --title --body-file`; `pr edit --add-label --remove-label`; `api --paginate --jq` | `gh --version`, `gh run list --json` (lista de campos), `gh run list --help`, `gh run view --help`, `gh run watch --help`, `gh pr view --json`, `gh pr create --help`, `gh pr edit --help`, `gh api --help` |
| V26 | El programa de `jq` de la aceptación (quickstart §11.5) imprime el resumen y termina con `true` sobre un informe con la forma de ADR 0016 | `jq --version` (`jq-1.7.1-apple`); experimento sobre un `informe.json` sintético en el clon |
| V27 | `make skills-check` ejecuta `TestSkillsDelRepositorio`, `TestNormasDelRepositorio`, `TestEvalsDelRepositorio` y `TestIdentificadoresDeLasNormas`, así que un subtest nuevo de `TestEvalsDelRepositorio` entra en él y en `make ci` sin tocar el `Makefile` | `Makefile:103-105` y `:162` (`ci`) |
| V28 | `make skills-sync` regenera `references/` con `TestSkillsDelRepositorio -regenerar-skills` | `scripts/skills-sync.sh`; experimento en el clon |
| V29 | `SKILL.md` no puede nombrar una norma por rango, número y año sin que su título esté en `data/normas.yaml`, ni decir `eval`, `evals`, `job`, `GitHub Actions` ni un id de modelo fuera de la región generada | `internal/app/skills_test.go:569-570` y `:640-642` |
| V30 | Las mutaciones del quickstart §4 funcionan con `perl -0pi` sobre UTF-8 sin `use utf8` (la marca se compara byte a byte) y `go -C <clon> test -run '^TestEvalsDelRepositorio$/^avisos-de-la-skill$'` falla nombrando el código; `go` admite `-C dir` | Experimento en el clon con prototipos de los subtests (las de `SKILL.md`, repetidas sobre el texto literal del contrato de la forma fija §3 en V33); `go help build` («-C dir») |
| V31 | El workflow pausa tras una tarea `[datos]` que modifica un fichero existente bajo `testdata/` o `schemas/`, y rechaza en `tasks.md` una tarea que nombra `testdata/` o `schemas/` sin `[datos]` | `.specify/workflows/hito/workflow.yml:937-954` (`clasificar_datos`) y `:470-471` (`precheck_tasks`) |
| V32 | Las formas de orden del quickstart están en la lista de permitidos de las sesiones sin aprobación (`rtk:*`, `git:*`, `go:*`, `make:*`, `jq:*`, `gh run list/view/watch:*`, `gh pr view/create/edit:*`) | `.claude/settings.json` (`permissions.allow`) |
| V33 | El texto literal del contrato de la forma fija §3, aplicado al `SKILL.md` de la base (181 líneas), añade 18 líneas —2 en el paso 5, 15 en «Cómo se cita» y 1 en la regla 3— y deja 199; `git diff --unified=0` da tres bloques, `@@ -114 +114,3 @@` (paso 5, entre `### 5.` en la 102 y `## Cómo se cita` en la 121), `@@ -140,0 +143,15 @@` (hasta `## Comandos` en la 158) y `@@ -174,3 +191,4 @@` (la regla 3 de `## Reglas`, en la 182); la línea añadida más larga tiene 119 caracteres, y las únicas de más de 120 del fichero son las de la región generada (163 a 180), igual que en la base; `make skills-check` pasa, con `TestSkillsDelRepositorio/trescientas-lineas`, `/normas-nombradas` y `/sin-instrucciones-de-evals` en verde, así que el texto no nombra ninguna norma ni dice ninguna palabra que prohíba FR-077 de H5 (V29); las búsquedas del quickstart §2 dan `⚠ NORMA DEROGADA:` en las líneas 147 (la lista) y 152 (el ejemplo), `⚠ VIGENCIA AGOTADA:` en la 148 y `⚠ TEXTO POSIBLEMENTE DESACTUALIZADO:` en la 149; con los prototipos, `avisos-de-la-skill` pasa sobre ese texto y cada una de las tres mutaciones de `SKILL.md` del quickstart §4 da `1 file changed, 2 insertions(+), 2 deletions(-)` y falla con `falta la forma fija del aviso derogada: ⚠ NORMA DEROGADA:` | Experimento en `/tmp/kitlegal-plan-h51-r2` (commit `2d2efb8`): los tres cambios del §3 aplicados tal cual, `wc -l`, `git diff --numstat` (`22 4`) y `--unified=0`, longitud en caracteres con `perl -CSD`, `make skills-check` y `go test -v -run '^TestSkillsDelRepositorio$' ./internal/app/`; en `/tmp/kitlegal-plan-h51`, el mismo `SKILL.md` confirmado encima de los prototipos y las órdenes del quickstart §4 |

## D1 · Etiquetas de aviso exportadas desde `internal/source/boe`

**Decisión.** `internal/source/boe/avisos.go` declara una constante por etiqueta (`TEXTO POSIBLEMENTE DESACTUALIZADO`,
`NORMA DEROGADA`, `VIGENCIA AGOTADA`), compone cada frase con su etiqueta (`"⚠ " + etiqueta + ": …"`, con el resto de la
frase igual que hoy) y exporta `func EtiquetasDeAviso() map[string]string`, que devuelve en cada llamada un mapa nuevo
de código a etiqueta con exactamente los tres códigos de `CodigosDeAviso()`.

**Por qué.** Es la única fuente de verdad de FR-010 y FR-011 también dentro del binario: la etiqueta no está copiada en
la frase, así que no puede divergir de lo que el applet emite. Un mapa se recorre y se compara entero en un test
(«exactamente 3 etiquetas», SC-005), y un mapa nuevo por llamada sigue la convención de `CodigosDeAviso()` (lista nueva
por llamada). Las frases siguen siendo constantes y byte a byte las mismas (V9): `TestAvisosDe` las compara con sus
literales y los golden de salida no cambian (FR-012).

**Alternativas rechazadas.**

- *Constantes exportadas por etiqueta* (`EtiquetaDeNormaDerogada`…): `internal/evals` necesitaría escribir a mano la
  correspondencia código → constante, que es otra copia que puede divergir de `CodigosDeAviso()`.
- *`func EtiquetaDeAviso(codigo string) (string, bool)`*: no se puede enumerar, así que ni un test del binario ni la
  comprobación de FR-014 podrían decir «exactamente estas tres»; y obliga a quien llama a tratar un `false` imposible.
- *Una lista de estructuras `{Codigo, Etiqueta}` de la que salga `CodigosDeAviso()`*: cambiaría cómo se construye un
  enumerado existente, más de lo que el alcance permite («tocar `internal/source/boe` más allá de exportar las
  etiquetas»).
- *Dejar las frases con la etiqueta escrita dentro y comprobar en un test que empiezan por ella*: la etiqueta quedaría
  escrita dos veces en el mismo fichero; el test detectaría la divergencia, pero la composición la hace imposible, que
  es mejor (constitución, criterio 1). El test de prefijo se conserva igualmente como red (`TestEtiquetasDeAviso`).

## D2 · La forma fija y su tolerancia, al detalle (FR-032)

**Decisión.** Un aviso está en un texto si el texto contiene, en una misma línea, esta forma para su etiqueta:

```text
forma     = marca [variante] sep palabra₁ { entre palabraᵢ } sep ":"
marca     = U+26A0 (⚠)
variante  = U+FE0F | U+FE0E            (a lo sumo una)
blanco    = U+0009 | carácter de la categoría Zs de Unicode (espacio, U+00A0, U+202F, …)
sep       = { blanco | "*" | "_" }      (cero o más)
entre     = sep blanco sep              (al menos un blanco)
palabraᵢ  = la i-ésima palabra de la etiqueta exportada (strings.Fields), exacta salvo mayúsculas y minúsculas
```

Se implementa con una expresión regular por código, compilada una sola vez desde `boe.EtiquetasDeAviso()`: cada palabra
con `regexp.QuoteMeta` dentro de `(?i:…)`, y las piezas `\x{26A0}`, `[\x{FE0E}\x{FE0F}]?`, `[\t\p{Zs}*_]*` y
`[\t\p{Zs}]` para lo demás. Lo que va detrás de los dos puntos no se lee.

| Tolerado (no cambia qué aviso es) | Ejemplo, encontrado | No tolerado (cambia qué aviso es, o no es la forma) | Ejemplo, ausente |
|---|---|---|---|
| Selector de presentación del emoji | `⚠️ NORMA DEROGADA:` | Palabra de más, de menos, distinta o pegada | `⚠ NORMA PARCIALMENTE DEROGADA:`, `⚠ NORMADEROGADA:`, `⚠ NORMA DEROGADAS:` |
| Énfasis de Markdown con `*` o `_` entre las partes | `**⚠ NORMA DEROGADA:**`, `⚠ **NORMA DEROGADA**:`, `_⚠ NORMA DEROGADA:_` | Sin la marca, sin los dos puntos o con otros dos puntos | `NORMA DEROGADA:`, `⚠ NORMA DEROGADA`, `⚠ NORMA DEROGADA：` |
| Espacios: ninguno o varios alrededor de las partes, uno o varios entre palabras, tabuladores y blancos de Unicode | `⚠NORMA DEROGADA:`, `⚠  NORMA  DEROGADA :`, `⚠ NORMA DEROGADA:` con U+00A0 | Un salto de línea dentro de la forma | `⚠ NORMA` + salto + `DEROGADA:` |
| Mayúsculas y minúsculas | `⚠ norma derogada:`, `⚠ Norma Derogada:` | Tachado (`~~`), código en línea partiendo la forma, otro emoji, dos selectores | `⚠ NORMA ~~DEROGADA~~:`, `🚨 NORMA DEROGADA:` |
| Lo que rodea a la forma | `> ⚠ NORMA DEROGADA:`, la forma dentro de un código en línea entero | Otra redacción o la negación | «Esta ley fue derogada.», «La Ley 30/1992 sigue en vigor.» |

Todas estas filas están comprobadas (V6, V7) y son los casos de `TestExtraerAvisos`.

**Por qué.** Es la lectura literal de FR-032 y de la *Decisión del mecanismo*: tolerancia «solo a lo que no cambia qué
aviso es» y exactitud en la etiqueta, igual que la cita es exacta en `<identificador>, bloque <id>]` y tolerante con lo
que la rodea. Cada tolerancia tiene un motivo concreto: el selector U+FE0F lo añaden muchos teclados y modelos al
escribir ⚠ (U+FE0E es la otra variante de presentación del mismo carácter); el énfasis es la forma natural de hacer
visible un aviso en Markdown; los blancos de Unicode los introducen los editores; las mayúsculas no cambian la etiqueta.
El tachado no se tolera porque en Markdown dice lo contrario. Los saltos de línea no se toleran porque la forma es una
marca visible en una línea, como la cita, y un salto entre sus partes ya no se lee como la forma. El plegado de
mayúsculas de Unicode (`(?i)`) también iguala `ſ` con `s` (V6); no cambia qué aviso es y no se restringe.

**Alternativas rechazadas.**

- *Normalizar el texto (quitar `*` y `_`, colapsar blancos, pasar a mayúsculas) y buscar la subcadena
  `⚠ ETIQUETA:`*: quitar `_` y `*` de todo el texto juntaría palabras que el texto separa (`NORMA_DEROGADA` pasaría a
  `NORMADEROGADA` o, según el orden de las normalizaciones, a `NORMA DEROGADA`), y cada normalización añade un caso
  límite que la expresión describe de una vez y con la gramática a la vista.
- *Tolerar saltos de línea o `\s` entero*: `⚠` al final de un párrafo y una etiqueta al principio del siguiente
  pasarían por la forma.
- *Comparar con `strings.EqualFold` recorriendo el texto*: más código para la misma gramática y sin la tolerancia al
  énfasis y a los blancos.
- *Cualquier lectura de la redacción libre* (listas de expresiones, atribución a normas, similitud, modelo): descartada
  por el hito (*Decisión del mecanismo*) y por la constitución (capa 1).

`regexp.MustCompile` no puede fallar con estas piezas y palabras pasadas por `QuoteMeta`, y se usa igual que
`formaDeCita` (`internal/evals/citas.go:13`), en un paquete de herramientas de desarrollo que el binario no enlaza: no
hay `panic` en ninguna ruta de usuario (constitución §IV). Una etiqueta sin palabras no da expresión: ese código nunca
se encuentra, así que falta en el juicio y en la comprobación de FR-014, que fallan nombrándolo; nunca pasa en vacío.

## D3 · `ExtraerAvisos` y el reparto en `Juzgar`

**Decisión.** `func ExtraerAvisos(texto string) []string` (`internal/evals/avisos.go`) devuelve los códigos de
`boe.CodigosDeAviso()` cuya forma fija lleva el texto, en el orden de `CodigosDeAviso()`, o `nil` si ninguno. `Juzgar`
llama a `repartirAvisos(eval.Avisos, ExtraerAvisos(sesion.Respuesta))` justo después de `repartirCitas`: cada aviso
esperado, en el orden de la eval y con sus repeticiones, va a `AvisosEncontrados` si está entre los extraídos y a
`AvisosAusentes` con el motivo `aviso ausente: <código>` si no. `Pasa` exige además `len(AvisosAusentes) == 0`.

**Por qué.** Es la misma forma que `ExtraerCitas` + `repartirCitas` (`internal/evals/citas.go:33`,
`internal/evals/juzgar.go:253-266`), que es lo que pide «como hace con las citas» (FR-030): el motivo va detrás de los
de las citas (FR-041) y, como el del modelo lo añade `exigirElModeloPedido` después de `Juzgar`, queda antes que él.
Una eval sin `avisos` no entra en el bucle y deja las dos listas nulas, que el informe escribe como `[]` (V8): el
resultado es el de antes de H5.1 (FR-034) y ningún caso existente de `TestJuzgar` cambia. Lo no esperado no se mira: la
forma de un aviso que la eval no espera no cambia nada (spec, *Edge Cases*).

**Alternativas rechazadas.** *Devolver los avisos en orden de aparición*: el orden útil es el de la eval, que es el
que se publica, y el de `CodigosDeAviso()` hace la función determinista sin depender de dónde escriba la respuesta cada
forma. *Quitar las repeticiones de `avisos`*: las citas repetidas se reparten repetidas y FR-022 pide el mismo
criterio.

## D4 · `avisos` en `schemas/eval.yaml.json` y en `Eval`

**Decisión.** El esquema gana, en `properties`, `"avisos": {"type": "array", "minItems": 1, "items": {"$ref":
"#/$defs/codigo-de-aviso"}}`; en `$defs`, `"codigo-de-aviso": {"enum": ["consolidacion-no-finalizada", "derogada",
"vigencia-agotada"]}` en una sola línea y en el orden de `CodigosDeAviso()`; y la rama `else` del `if` de `activa`
añade `{"required": ["avisos"]}` a su `anyOf`. `Eval` gana `Avisos []string \`yaml:"avisos"\``, detrás de `Citas`.

**Por qué.** `minItems: 1` y la ausencia de `uniqueItems` dan exactamente el trato de `citas` a la lista vacía y a los
repetidos (V2, FR-022); la rama `else` es la que ya prohíbe `comandos` y `citas` en una eval de no activación (V3); y el
enumerado en `$defs` lo leen la validación y la comprobación de FR-013 por el mismo camino (V4). Una sola línea con el
orden de `CodigosDeAviso()` hace legible el diff de la pausa y reproducibles las mutaciones del quickstart.

**Alternativas rechazadas.** *`uniqueItems: true`*: trataría los repetidos distinto que `citas`, contra FR-022.
*Validar el enumerado en Go en lugar del esquema*: el formato común es el esquema publicado (H5, contrato de evals §1) y
la comprobación de FR-013 compara precisamente el esquema. *Generar el esquema desde el código*: `schemas/eval.yaml.json`
está escrito a mano desde H5 y generarlo es un cambio de mecanismo que el hito no pide.

## D5 · Comprobación de FR-013: esquema ↔ `CodigosDeAviso()`

**Decisión.** `func ComprobarCodigosDeAviso(esquema *jsonschema.Schema) error` (`internal/evals/avisos.go`) lee los
códigos del esquema **compilado** —`Properties["avisos"]`, su `Items2020` y, mientras haya `Ref`, el esquema al que
apunta, hasta un `Enum`— y devuelve `errors.Join` de un error por código de `CodigosDeAviso()` que no enumera
(`el esquema de eval no enumera el código de aviso <código> en avisos`), en el orden de `CodigosDeAviso()`, y otro por
código que enumera sin ser de `CodigosDeAviso()` (`el esquema de eval enumera en avisos el código <código>, que no es un
código de aviso del binario`), en el orden del esquema; `nil` si coinciden como conjuntos. Un esquema sin `avisos`, sin
`items` o sin enumerado no enumera ninguno, así que falla nombrando los tres. Se aplica al esquema publicado en
`TestEvalsDelRepositorio/avisos-del-esquema` y a esquemas sintéticos en `TestComprobarCodigosDeAviso`.

**Por qué.** El esquema compilado es lo que valida `LeerEval`: si alguien mueve el enumerado o lo parte en otro `$ref`,
la comprobación sigue mirando lo que de verdad se aplica. Comparar conjuntos es «coinciden exactamente» (FR-013) sin
depender del orden, que no cambia qué admite el esquema. Probado con un código de menos y uno de más en el clon (V30).

**Alternativas rechazadas.** *Leer el JSON en crudo por un puntero fijo* (`/$defs/codigo-de-aviso/enum`): pasaría en
vacío si el enumerado se mueve y el esquema deja de usarlo. *Un recorrido de todos los `enum` del documento, como
`TestEnumeradosDeLosDatos`*: el esquema de eval tiene otro enumerado (`verbo`), y distinguirlos exige de todas formas
saber cuál es el de `avisos`.

## D6 · Dónde viven las comprobaciones y cómo entran en `make ci`

**Decisión.** Las dos comprobaciones (D5 y D7) son funciones de `internal/evals/avisos.go` con tests propios sobre
entradas sintéticas, y se aplican al repositorio en dos subtests nuevos de `TestEvalsDelRepositorio`:
`avisos-del-esquema` y `avisos-de-la-skill`. `make skills-check` y `make ci` los ejecutan sin tocar el `Makefile` (V27).

**Por qué.** La comprobación de FR-014 tiene que usar «la misma función que usa `Juzgar`» (Clarifications, Q2), que
vive en `internal/evals`; y la de FR-013 compara el formato de eval, que también es suyo. `TestEvalsDelRepositorio` es
el test que ya comprueba «las evals del repositorio» dentro de `make skills-check`, cuya orden la documentación ya
describe como «comprueba skills, datos y evals sin red, sin modelo».

**Alternativas rechazadas.** *Un test de nivel superior nuevo*: obligaría a cambiar la expresión de `skills-check` en
el `Makefile` y, con ella, la documentación de la orden, sin ganar nada. *Ponerlas en `internal/app/skills_test.go`*:
ese test comprueba lo generado de las skills, y tendría que importar `internal/evals` para usar la misma función.

## D7 · Comprobación de FR-014: `SKILL.md` ↔ forma fija de cada código

**Decisión.** `func ComprobarFormasDeAviso(texto string) error` devuelve `errors.Join` de un error por código de
`boe.CodigosDeAviso()` que `ExtraerAvisos(texto)` no encuentra, en ese orden:
`falta la forma fija del aviso <código>: ⚠ <etiqueta>:` (la etiqueta sale de `boe.EtiquetasDeAviso()`). El subtest
`avisos-de-la-skill` lee `../../skills/boe-legislacion/SKILL.md` y exige `nil`; `TestComprobarFormasDeAviso` fija los
casos de SC-005 (falta la etiqueta, la marca o los dos puntos) y el completo.

**Por qué.** Es literalmente FR-014 con la Clarification Q2: en cualquier parte del fichero, con la misma función y
tolerancias que el juicio, exacta en la etiqueta, y fallando nombrando el código. Escribir la forma esperada en el
mensaje dice qué falta sin abrir el código.

**Alternativa rechazada.** *Buscar la etiqueta suelta*: la Clarification Q2 exige la forma completa, y la etiqueta
suelta dentro de una descripción genérica no enseña la forma.

## D8 · Sin copias de las etiquetas en el código de `internal/evals`

**Decisión.** `TestEtiquetasSoloDesdeBoe` lee cada fichero `*.go` del paquete que no termina en `_test.go` y exige que
ninguno contenga ninguna de las etiquetas de `boe.EtiquetasDeAviso()`. Los comentarios de `avisos.go` y `juzgar.go`
hablan de «la marca, la etiqueta y los dos puntos» y de los códigos, nunca de una etiqueta concreta.

**Por qué.** Hace mecánico US4-5 y SC-005 («no contienen ningún literal de las 3 etiquetas»), y cubre todo el código del
paquete, no solo `juzgar.go` y `avisos.go`, de modo que una copia no se pueda esconder moviéndola de fichero. Los tests
sí escriben las formas literales: fijan lo esperado con independencia del código que lo produce.

**Alternativa rechazada.** *Revisarlo a mano en la revisión final*: lo que un script comprueba no se delega a un juez
(constitución, «Gates»).

## D9 · El informe publica los avisos

**Decisión.** `ResultadoDeEval` gana `AvisosEncontrados []string \`json:"avisos_encontrados"\`` y
`AvisosAusentes []string \`json:"avisos_ausentes"\``, detrás de `CitasAusentes`. En `informe.md`, la tabla «Sesiones»
gana las columnas `Avisos encontrados` y `Avisos ausentes`, detrás de `Citas ausentes`, con los códigos separados por
«, » o `ninguno`. La sección de cada sesión no cambia: sigue publicando la respuesta. Los motivos de la sesión llegan a
la raíz del informe como los demás (`<sesión>: aviso ausente: <código>`).

**Por qué.** FR-040 a FR-042: las claves y el orden de la eval en el resultado estructurado, las listas vacías que da
`encoding/json/v2` (V8), y «junto a sus citas» en la forma legible, que es la tabla donde están las citas ausentes. La
respuesta, donde queda visible la limitación declarada (una forma fija seguida de lo contrario), ya se publica en la
sección de la sesión. `ninguno` concuerda con «aviso» y es el texto que el informe ya usa (`ningunoEnElInforme`).

**Alternativas rechazadas.** *Líneas `Avisos encontrados:` y `Avisos ausentes:` en la sección de cada sesión*: las
citas no están ahí, así que no estarían «junto a sus citas». *Solo la columna de ausentes, como con las citas*: FR-042
pide encontrados y ausentes. *Un recuento por código o por eval*: fuera de alcance (spec, Clarifications Q1).

## D10 · El test del informe con avisos, sin material nuevo bajo `testdata/`

**Decisión.** `TestInformeConAvisos` copia con `os.CopyFS` el caso `testdata/sesiones/informe/aprobado` a un
`t.TempDir()`, y en la copia añade a `evals/01-lpac-articulo-21.yaml` `avisos` con `derogada` y `vigencia-agotada`, y
en `sesiones/01-lpac-articulo-21/sesion.jsonl` antepone a la respuesta `⚠ NORMA DEROGADA: esta norma ha sido derogada. `
(exigiendo exactamente dos sustituciones, V20); en un segundo subtest, además, quita la cita de la respuesta. Después
llama a `EscribirInforme` y comprueba el resultado, `informe.json` en crudo e `informe.md`. `TestInforme/aprobado`
exige además `[]` en las dos claves de cada sesión.

**Por qué.** Un caso nuevo bajo `internal/evals/testdata/sesiones/informe/` sería material de test en territorio de
fixtures: obligaría a una tarea `[datos]` y, como `TestInforme` exige que cada directorio sea un caso de su tabla (V19),
dejaría `make ci` en rojo entre la tarea que lo crea y la que añade el caso. Copiar y modificar en un temporal reutiliza
una ejecución sintética ya revisada y pone a la vista, en el propio test, lo único que cambia. Es el mismo criterio que
se aplicó en la revisión final de H5 (`TestLeerTrazasHiloCreadoAntesDeLaEjecucion`, constantes escritas a un temporal).

**Alternativa rechazada.** *Escribir la ejecución entera desde constantes*: repetiría un transcript y una traza que el
caso `aprobado` ya tiene, sin ganar nada.

## D11 · La grabación del índice de `BOE-A-1992-26318`

**Decisión.**

1. El manifiesto `testdata/evals/grabaciones.json` gana una entrada: búsqueda `procedimiento administrativo común` (la
   misma de H4), `titulo_empieza_por` `Ley 30/1992,`, `bloques` `["a42"]` y su `para`.
2. `TestGrabarEvals` siembra cada consulta desde **la unión** de las grabaciones de H4 y de H5
   (`UnionDeGrabaciones()`), no solo desde las de H4: así solo llega a la fuente lo que no está grabado en ningún
   conjunto. Con la entrada nueva, eso es una única consulta, el índice de `BOE-A-1992-26318` (V15, V16, V18), más el
   `robots.txt` que `httpx` pide antes de la primera petición real.
3. En la pausa de la tarea `[datos]` del manifiesto, una persona ejecuta `scripts/grabar-evals.sh`, **restaura** la
   grabación de `robots.txt` que la ejecución reescribe (V11, V13) y confirma solo el índice nuevo (procedimiento en el
   contrato de la eval y la grabación §4).

**Por qué.** FR-054 y FR-055: reutilizar lo grabado y añadir solo el índice, grabado por una persona en una pausa. Con
el arnés tal como está (V12), la ejecución pediría a la fuente las nueve búsquedas y las veintiuna consultas de norma
—metadatos, índice y bloque de las siete normas— que ya grabó H5 (la búsqueda de la LPAC, y los metadatos, el índice y
el bloque de la LPAC, la LCSP y la LRBRL, están en las grabaciones de H4) y reescribiría sus ficheros: treinta
peticiones innecesarias al BOE, más el `robots.txt` que se pide en cualquier caso (constitución §I: sin carga que la
fuente no necesita), y un diff de esos treinta ficheros que la persona tendría que revisar y deshacer, cuando grabar
«respuestas del BOE distintas de las que necesita la eval nueva» está fuera de alcance. La grabación de `robots.txt` no
la usa `httpx.Replay` (V13), así que restaurarla deja el diff en lo que la eval necesita. El cambio del arnés es de una
línea en un fichero que solo compila la etiqueta `grabacion` y que el lint cubre (`.golangci.yml`, `run.build-tags`).

**Alternativas rechazadas.**

- *Que la persona deje en el manifiesto solo la entrada nueva mientras graba y lo restaure después*: edita a mano un
  fichero versionado en mitad de la pausa, no es reproducible y un olvido deja el manifiesto roto en la rama.
- *Grabar el manifiesto entero y deshacer lo que no sea el índice*: las treinta peticiones y el riesgo de conservar por
  error una respuesta que la fuente cambió, que rompería las evals de otras normas.
- *Una variable o una bandera para grabar solo algunas entradas*: más superficie en un arnés que la unión resuelve sin
  ninguna opción nueva.

## D12 · La premisa de `TestIdentificadoresDeLasNormas/norma-en-la-busqueda-de-otra-entrada`

**Decisión.** `otraNormaDeLaBusqueda` elige el primer resultado de la búsqueda grabada de la entrada que **ni está en la
tabla de normas ni lo resuelve ninguna entrada del manifiesto** (su título no empieza por el `titulo_empieza_por` de
ninguna), y recibe para ello el manifiesto entero y la posición de la entrada.

**Por qué.** El subtest describe «una norma que da la búsqueda de otra entrada sin ser la que esa entrada resuelve», que
tiene que fallar porque **ninguna** entrada la resuelve. Hoy elige `BOE-A-1992-26318` solo porque es el primer resultado
fuera de la tabla; con la entrada nueva del manifiesto, esa norma pasa a tener entrada y el subtest falla por su premisa,
no por el código que prueba (V14). La premisa corregida es la que el subtest ya enuncia: mientras el manifiesto no
tenga la entrada nueva sigue eligiendo `BOE-A-1992-26318`, y con ella elige `BOE-A-2018-12131`. Sin la corrección, el manifiesto y `data/normas.yaml` tendrían que cambiar en la misma tarea, y
una tarea `[datos]` solo toca `testdata/` y `schemas/` (V31; juez de tareas, criterio e).

**Alternativas rechazadas.** *Una tarea `[datos]` que toque también `data/normas.yaml` y `references/`*: la prohíbe la
separación de tareas del workflow. *Añadir la norma a `data/normas.yaml` antes que la entrada del manifiesto*:
`TestIdentificadoresDeLasNormas` falla con «ninguna entrada del manifiesto lo resuelve». *Buscar la norma nueva con otra
búsqueda*: cualquier entrada que la resuelva la saca de «otra» igual.

## D13 · La norma en `data/normas.yaml`

**Decisión.** `BOE-A-1992-26318` entra al final de la tabla con el título y el rango de la búsqueda grabada (V16),
`abreviatura: LRJPAC` y `materias` `régimen jurídico de las administraciones públicas` y `procedimiento administrativo`.
Ni la tabla ni `references/normas.md` marcan la derogación (FR-056). `make skills-sync` regenera la referencia.

**Por qué.** Título y rango vienen de la respuesta grabada, como exige `TestIdentificadoresDeLasNormas`. Las materias
son las dos del propio título, sin inventar nada, y `procedimiento administrativo` es la materia que ya usa la LPAC,
que es la norma que la sustituyó. La abreviatura es la usual de la ley y sigue la convención de las diez entradas, todas
con la suya; el paso 1 de la skill localiza normas por abreviatura. Ninguna regla del conjunto de evals mira las
materias de una eval informativa (data-model §6.3 de H5, enmienda de ADR 0016).

**Alternativa rechazada.** *Sin abreviatura*: el esquema la admite, pero rompería la convención de la tabla y dejaría
sin resolver por abreviatura una ley que se cita así.

## D14 · La eval de la norma derogada (FR-050, FR-051)

**Decisión.** `evals/boe-legislacion/18-lrjpac-norma-derogada.yaml`, con la pregunta literal
**«¿Qué dice el artículo 42 de la Ley 30/1992 sobre la obligación de resolver?»**, `activa: true`, `informativa: true`,
el comando esperado bloque `boe` `BOE-A-1992-26318` `a42`, la cita esperada `BOE-A-1992-26318` `a42` y `avisos`
`derogada` y `vigencia-agotada`, en ese orden.

**Por qué.** Como las positivas desde T043 de H5, la pregunta nombra la norma y el artículo para medir el protocolo y no
la memoria del modelo (spec, *Assumptions*); la materia es la del propio artículo grabado (V17). La pregunta **no**
menciona la vigencia ni la derogación: lo que se mide es que la skill traslade los avisos que le da el binario, y una
pregunta que los anunciara mediría la pregunta. El comando y la cita son los de toda positiva (el bloque), y el bloque
`a42` ya emite los dos avisos (V17), así que no hace falta esperar `metadatos`. El número 18 es el siguiente libre y el
nombre sigue la forma de los ficheros de eval.

**Alternativas rechazadas.** *«¿Sigue vigente el artículo 42 de la Ley 30/1992?»*: sugiere la respuesta. *Esperar
también `metadatos`*: no lo necesita el aviso y exigiría un comando que el protocolo solo pide cuando la pregunta
depende de la vigencia. *Una norma derogada distinta*: fuera de alcance.

## D15 · El texto de `SKILL.md`

**Decisión.** Tres cambios, todos sobre cómo se traslada un aviso (FR-001 a FR-005): la regla 3 dice que cada aviso se
traslada con su forma fija; el último punto del paso 5 describe la misma forma y remite a «Cómo se cita»; y «Cómo se
cita» gana, detrás de las reglas de la cita, un párrafo con la forma, las tres formas completas (`⚠ NORMA DEROGADA:`,
`⚠ VIGENCIA AGOTADA:`, `⚠ TEXTO POSIBLEMENTE DESACTUALIZADO:`) con su código, un ejemplo —la frase literal que el
binario emite para `derogada`, `⚠ NORMA DEROGADA: esta norma ha sido derogada.` (V17)— y la regla de escribir la
etiqueta entera en la misma línea. El texto literal está en el contrato de la forma fija §3.

**Por qué.** Es donde FR-001 a FR-003 lo ponen. El ejemplo es, byte a byte, la frase que el binario emite para
`derogada` (data-model §1) y no nombra ninguna norma: todo lo que enseña sale del sobre (constitución §II), no enseña la
respuesta de ninguna eval (una eval que viera su propio caso en la skill mediría la skill copiada, no el protocolo) y no
activa la comprobación de normas nombradas (V29). Que detrás de la forma también puede ir una explicación lo dicen el
párrafo, el paso 5 y la regla 3 (FR-001), sin que la skill dicte ninguna. Verificado sobre el texto literal del §3, no
sobre un prototipo (V33): añade 18 líneas a las 181 de la base —2 en el paso 5, 15 en «Cómo se cita» y 1 en la regla 3—
y deja 199, por debajo de 300; ninguna línea añadida pasa de 120 caracteres (la más larga tiene 119); no usa ninguna
palabra que prohíba FR-077 de H5 ni nombra ninguna norma (V29, con `make skills-check` en verde); y lleva la forma fija
de los tres códigos, que `avisos-de-la-skill` encuentra con los prototipos.

**Alternativas rechazadas.**

- *Un ejemplo con el artículo 42 de la Ley 30/1992*: es la pregunta de la eval 18.
- *Un ejemplo con la frase del binario seguida de una explicación* (`⚠ NORMA DEROGADA: esta norma ha sido derogada; lo
  citado no es derecho vigente.`): la explicación sería una afirmación escrita por la skill que el sobre no trae, y un
  ejemplo es un modelo que la respuesta puede reproducir tal cual; la skill propondría añadir a cada aviso una
  consecuencia fija en lugar de la que se desprenda de lo consultado (constitución §II: «nunca se inventa contenido
  legal»). La posibilidad de explicar ya la dan el párrafo, el paso 5 y la regla 3.

## D16 · Orden de implementación y tareas `[datos]`

**Decisión.** El orden de plan.md («Orden de implementación»): etiquetas en el binario → forma fija y comprobaciones →
`[datos]` esquema → formato → juicio → informe → arnés y premisa de identificadores → `[datos]` manifiesto (pausa con la
grabación) → norma y referencia → eval 18 → `SKILL.md` con su subtest → documentación → cierre → `[plataforma]`.

**Por qué.** Deja `make ci` en verde tras cada paso, comprobado en el clon en los puntos delicados: el esquema antes que
el campo (V5), el manifiesto antes que la norma con la premisa corregida (V14), la norma antes que la eval (regla
«normas conocidas») y la grabación antes que la eval (V15). La eval precede al primer commit que cambia `SKILL.md`
(FR-052, Definition of Done §1.10): es la medida de la skill y se escribe antes que la regla que mide. No puede ir
antes que el formato, porque sin `avisos` en el esquema es un fichero mal formado; el formato y el juicio son el
instrumento, no el comportamiento medido. Las dos tareas `[datos]` solo tocan `schemas/` y `testdata/` (V31).

**Alternativa rechazada.** *La eval antes que el formato*: `make ci` en rojo, con un fichero mal formado.

## D17 · La ejecución de aceptación en la plataforma (FR-070 a FR-073)

**Decisión.** La tarea `[plataforma]`, la última, empuja la rama y abre la propuesta de cambio con el cuerpo de
`gates/pr-h5.1.md`; abrirla dispara el job, porque la propuesta toca sus rutas (V22). La ejecución de aceptación es **la
primera ejecución del flujo `evals` de la rama con evento `pull_request`**, que se espera, se lee entre las marcas del
registro (V23, V24) y se comprueba con un programa de `jq` (V26): veredicto `aprobado`, `red` vacío, la serie de
`18-lrjpac-norma-derogada.yaml` con el modelo que decide planificada, sin decidir y con 3 sesiones, y cada una de esas 3
sesiones con `avisos_encontrados` + `avisos_ausentes` igual a `derogada` y `vigencia-agotada`; y que el `commit` del
informe es el `headSha` de la ejecución y que entre ese commit y la cabeza solo cambian ficheros bajo
`specs/007-h5-1-avisos-de-vigencia/`. Todo se registra en `gates/evals-aceptacion.md`. Si un commit posterior cambia
algo fuera de ese directorio (por ejemplo, una corrección de la revisión final), la aceptación se repite poniendo la
etiqueta `evals` e identificando la ejecución por el último evento `labeled`, como el cierre de H5 (quickstart §11.6).

**Por qué.** FR-070 nombra la ejecución «del job de evals que abre la propuesta de cambio». Sin `synchronize` en el
disparador, la única ejecución de apertura es la primera de la rama, y seleccionarla por orden de creación hace la orden
repetible tal cual si la sesión se retoma. La lectura de `informe.json` con `jq` convierte la aceptación en una
comprobación mecánica en lugar de una lectura (constitución, «Gates»). La repetición por etiqueta es el mecanismo que ya
existe (ADR 0016) y que el cierre de H5 validó.

**Alternativas rechazadas.** *«La última ejecución de evals de la rama»*: una etiqueta ajena crea una ejecución con el
job saltado y la última podría ser esa. *Leer solo `informe.md`*: no permite una comprobación mecánica. *Repetir siempre
la ejecución tras la revisión final*: FR-070 solo lo exige si cambia algo fuera del directorio del hito.

## D18 · Documentación

**Decisión.** `CHANGELOG.md` (*Unreleased*, «Añadido», un bloque *De H5.1*) registra la forma fija de los avisos en la
skill, el campo `avisos` del formato, el reparto de avisos en el informe y la eval nueva; `README.md` y
`CONTRIBUTING.md` añaden la fila `avisos` a su tabla del formato común de eval (opcional, solo con `activa: true`,
códigos del binario) y, en `CONTRIBUTING.md`, «y lleva la forma fija de cada aviso de `avisos`» a la frase que dice
cuándo pasa una eval. Ningún ADR, ninguna fila de `docs/SOURCES.md`, ningún cambio en `docs/USO.md`.

**Por qué.** FR-061 y Definition of Done §1.6; los demás puntos están fuera de alcance en el spec. Otras cifras de esas
páginas que no tratan de `avisos` no se tocan: no las pide el spec.

## D19 · Aplicación de las skills `golang-*`

`golang-how-to` orquesta; se aplicaron:

- **golang-project-layout**: ningún paquete nuevo; `avisos.go` en `internal/evals` junto a `citas.go`, un fichero de
  test por fichero de código. Nada en `internal/core` ni `pkg/`.
- **golang-naming**: `EtiquetasDeAviso` sigue a `CodigosDeAviso`; `ExtraerAvisos` a `ExtraerCitas`; `Comprobar…` a
  `ComprobarConjuntoDeBoeLegislacion`; identificadores en español según la convención del repositorio.
- **golang-error-handling**: las comprobaciones devuelven `error` con `errors.Join` de un error por código, sin
  silenciar ninguno; ningún `panic` en rutas de usuario.
- **golang-testing** y **golang-stretchr-testify**: tablas con subtests con nombre, `t.Parallel()`, `require` para las
  precondiciones y `assert` para las comparaciones; material nuevo en `t.TempDir()`; mutaciones del quickstart para
  demostrar que las comprobaciones no pasan en vacío.
- **golang-lint**: ningún `//nolint`, ninguna exclusión nueva en `.golangci.yml` (solo la palabra `variantes` en
  `misspell.ignore-rules`, que exige el nombre del subtest `aviso-con-variantes-toleradas`); prototipo en verde con el linter
  fijado (V10).
- **golang-safety**: el mapa de `EtiquetasDeAviso` y la lista de `ExtraerAvisos` son nuevos en cada llamada, sin
  aliasing con el estado del paquete.
- **golang-cli**: el binario no cambia (ni banderas, ni verbos, ni códigos de salida).

## S · Supuestos no verificados

Dependen de la plataforma o de la red de la fuente y no se pueden comprobar en local. Los comprueba y registra quien
ejecuta la tarea indicada; si uno no se cumple, la tarea se detiene y lo anota.

| # | Supuesto | Lo comprueba |
|---|---|---|
| S1 | Abrir la propuesta de cambio crea una ejecución con `workflowName` `evals` y evento `pull_request` cuyo job `evals` se ejecuta (`cambios` da `coincide=si`), y es la primera ejecución de `evals` de la rama | Tarea `[plataforma]`, quickstart §11.3 |
| S2 | El `headSha` de esa ejecución es la cabeza de la rama al abrir la propuesta y coincide con el `commit` de su informe | Tarea `[plataforma]`, quickstart §11.5 |
| S3 | El formato de `gh run view --log` sigue siendo el de V24 | Tarea `[plataforma]`, quickstart §11.4 (la orden falla si no) |
| S4 | Las 90 sesiones caben en `timeout-minutes: 120` (ADR 0016: por debajo de ~76 s por sesión con la preparación medida en H5) | Tarea `[plataforma]`: tiempo del paso «Ejecutar las evals» en la evidencia |
| S5 | La API del BOE sigue sirviendo el índice de `BOE-A-1992-26318` y su `robots.txt` lo permite | La persona, en la pausa del manifiesto (contrato de la eval y la grabación §4) |
