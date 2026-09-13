# Research: H4 · Applet `boe`: puerto de `boe.py`

**Modo**: desatendido. Cada decisión se tomó con el «Criterio de decisión autónoma» de
`.specify/memory/constitution.md` (siempre la mejor solución; lo no especificado no se implementa; elegir
con criterio y dejar rastro) y registra la alternativa rechazada y por qué. Ninguna afecta a alcance,
frontera humana, privacidad, reglas de anomalías ni a una decisión cerrada de `CLAUDE.md`. Dos cosas que
sí tocan términos de uso y responsabilidad legal —los términos de uso y el `robots.txt` del BOE, y la fila
de `docs/SOURCES.md`— **no se deciden aquí**: se verifican por una persona antes de grabar (FR-123,
constitución capa 3) y, si prohíben el acceso, el hito se detiene ([D15](#d15--la-fuente-docssourcesmd-ritmo-términos-robotstxt-y-verificación-nocturna)).

**Verificación.** Toda afirmación sobre una herramienta, una dependencia o la biblioteca estándar se ha
comprobado en local contra su código o su documentación, y cada decisión cita dónde. `refs/boe.py` no se
ejecuta: su comportamiento se lee en su fuente, y el de las piezas de CPython en las que se apoya se lee
en la fuente de CPython 3.9 que trae instalada esta máquina (Command Line Tools), **sin ejecutar Python**.
Lo que no se ha podido verificar sin red o sin ejecutar Python está en
[D20 · Supuestos no verificados](#d20--supuestos-no-verificados), declarado como supuesto y no como hecho.

| Qué | Dónde se comprobó |
|---|---|
| Toolchain | `go.mod`: `go 1.27.0` + `toolchain go1.27.1`; `go env GOROOT` → `/Users/jorge/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.1.darwin-arm64`; `go env GOMODCACHE` → `/Users/jorge/go/pkg/mod` |
| Dependencias del hito | Ninguna nueva. `go.mod` ya requiere las ocho de la constitución §V que el hito usa: `alecthomas/kong` v1.16.1, `invopop/jsonschema` v0.14.0, `rogpeppe/go-internal` v1.16.0, `santhosh-tekuri/jsonschema/v6` v6.0.3, `stretchr/testify` v1.12.1, `temoto/robotstxt` v1.1.2, `golang.org/x/time` v0.16.0, `modernc.org/sqlite` v1.58.0 |
| `encoding/xml` (GOROOT `src/encoding/xml/xml.go`) | 147 («The parser assumes that its input is encoded in UTF-8»); 170 (`Strict`); 188-193 y 644-656 (una declaración `encoding` distinta de UTF-8 sin `CharsetReader` es error: «encoding %q declared but Decoder.CharsetReader is nil»); 860-868 (el valor de un atributo entrecomillado pasa por `text`, **sin** normalizar tabuladores ni saltos a espacio); 1130-1136 (`\r\n` y `\r` sueltos se escriben como `\n` en el texto) |
| Biblioteca estándar (`go doc`) | `time.Parse` (un desplazamiento que corresponde a `Local` se lee en `Local`; si no, en una zona fija con ese desplazamiento: la ida y vuelta RFC 3339 no cambia el texto); `net/url.URL.String` (la consulta sale de `RawQuery` tal cual); `mime.ParseMediaType`; `os.CopyFS` (crea el destino y no sobrescribe); `go help testflag` 77-90 (`-fuzz`, `-fuzztime`); `go help test` 23 (el directorio `testdata` lo ignora el go command); `encoding/json.Marshal` (las cadenas pasan por `HTMLEscape`, «which replaces "<", ">", "&", U+2028, and U+2029»; «The map keys are sorted»); `encoding/json.Encoder.SetEscapeHTML` (`false` desactiva ese escape); `encoding/json.Decoder.UseNumber` (un número se lee como `Number`, no como `float64`); `encoding/json.Decoder.DisallowUnknownFields`; `encoding/xml.Decoder.Token` («guarantees that the StartElement and EndElement tokens it returns are properly nested and matched»: no promete nada sobre cuántos elementos raíz hay); `regexp/syntax` («Perl character classes (all ASCII-only): \d digits (== [0-9])»); `unicode.IsSpace` (propiedad `White_Space`; en Latin-1, `\t \n \v \f \r`, espacio, U+0085 y U+00A0: no incluye U+001C–U+001F); `unicode.IsDigit` («a decimal digit»); `strings.ToLower` («all Unicode letters»); `unicode/utf8.Valid`; `context.WithValue` (datos de ámbito de petición, con una clave de un tipo propio y no predefinido) |
| Biblioteca estándar (fuentes del GOROOT, `src/`) | `encoding/json/encode.go` 998-1064 (`appendString`: escapa `"`, `\`, los bytes < 0x20 —`\b \f \n \r \t` con su escape corto—, `<`, `>` y `&` con el escape HTML, y U+2028/U+2029; cualquier otra runa, ASCII o no, se copia tal cual, 1060-1062) y 797-803 (miembros de un objeto unidos por `,` y `:` sin espacios); `net/url/url.go` 79-82 (`shouldEscape` consulta la tabla que genera `gen_encoding_table.go`), 186-194 (`QueryEscape` y `PathEscape`) y 196-227 (`escape`: en `encodeQueryComponent` el espacio sale como `+`, 200 y 225-226); `net/url/gen_encoding_table.go` 175-211 (de los reservados `$ & + , / : ; = ? @`, un segmento de ruta solo escapa `/ ; , ?`, 190-193, y un componente de consulta los escapa todos, 202-204); `encoding/xml/xml.go` 275-290 (`Token`: al final de la entrada solo es error que quede un elemento abierto, 286-287), 440-500 (la pila guarda elementos abiertos, espacios de nombres y marcas de fin; nada cuenta elementos raíz, y el fichero no contiene ninguna comprobación de raíz) y 699-715 (`<![CDATA[…]]>` se entrega como `CharData`) |
| CPython 3.9 (lectura, sin ejecutar), `/Library/Developer/CommandLineTools/Library/Frameworks/Python3.framework/Versions/3.9/lib/python3.9/` | `xml/etree/ElementTree.py` 406-423 (`itertext`: texto propio, texto de cada hijo y su `tail`, en orden) y 979-983 (`_serialize_text`, que usa `tostring(..., method="text")`: `itertext` **más el `tail` del propio elemento**); `xml/etree/ElementPath.py` 182-209 (`.//tag` recorre descendientes y **excluye el propio elemento**, `if e is not elem`); `json/__init__.py` 183 (`dumps` con `ensure_ascii=True` por omisión); `json/encoder.py` 31 y 48-66 (con `ensure_ascii`, todo lo no ASCII y los controles salen como `\uXXXX` en **minúsculas**, con pares sustitutos fuera del plano básico) y 102-103 (separadores `', '` y `': '`); `urllib/parse.py` 790 (`_ALWAYS_SAFE` = letras ASCII, dígitos y `_.-~`), 815 (`'%{:02X}'`: hexadecimal en **mayúsculas**) y 890 (`quote_from_bytes(bs, safe='/')`); `xml/etree/ElementTree.py` 1334-1347 (`XML`, que es `fromstring` en 1373: `XMLParser(target=TreeBuilder())` y `feed`), 1413-1415 (`TreeBuilder` con `insert_comments=False` e `insert_pis=False` por omisión), 1439-1449 (`_flush`: el texto acumulado va al `text` o al `tail` del último elemento), 1486-1501 (`comment` y `pi` llaman a `_handle_single`), 1503-1511 (`_handle_single`: sin inserción ni añade el nodo ni vacía el texto acumulado, así que el texto de los dos lados queda unido), 1557-1559 (el `XMLParser` de Python conecta los manejadores de comentarios e instrucciones de proceso de expat a `target.comment` y `target.pi`) y 2085-2086 (`from _elementtree import *`: si existe, el acelerador C sustituye a esas clases; su fuente no está en local, S5 (e)) |
| `santhosh-tekuri/jsonschema/v6` v6.0.3 | `compiler.go` 47-53 (`AssertFormat`), 121 (`AddResource`), 178 (`Compile(loc)`); `roots.go` 104-124 (un subesquema con `$id` es un **recurso propio** con su base) y 166-188 (los subesquemas se recorren por las ubicaciones del borrador, `$defs` incluida); `root.go` 34 (`resolveFragmentIn`: un fragmento `#/…` se resuelve contra el recurso que lo contiene) |
| `invopop/jsonschema` v0.14.0 | `reflect.go` 772-801 (`stringKeywords`: cada `enum=<valor>` de la etiqueta se añade a `Enum`; `enum=` sin valor añade la cadena vacía) |
| `alecthomas/kong` v1.16.1 | `build.go` 338-339 («args are required by default»: un posicional es obligatorio salvo `optional`) |
| `rogpeppe/go-internal` v1.16.0 (`testscript`) | `testscript.go` 145-150 (`Setup` con `WorkDir` ya creado), 157-159 (`Cmds`), 495 (`HOME=/no-home` en el entorno del guion), 1211 (`TestScript.Exec`), 1239 (`Fatalf`); `doc.go` 231-232 (`symlink`); `cmd.go` 29-50 (órdenes integradas, `mkdir` y `rm` incluidas) |
| `golangci-lint` y analizadores (versiones de `tools/golangci-lint/go.mod`: `golangci-lint` v2.13.2 en 91, `misspell` v0.8.0 en 93) | **Comportamiento de `misspell`**: citado a su código en D18 (`pkg/golinters/misspell/misspell.go` de `golangci-lint`; `replace.go`, `notwords.go`, `case.go` y `stringreplacer.go` de `misspell`). **Barrido del diccionario**, repetible: a spec, plan, research, data-model, los cinco contratos y quickstart se les quitan tildes y diéresis (`ñ` → `n`), se pasan a minúsculas y se parten con la clase de palabra de `misspell` (`[a-zA-Z0-9']+`, `replace.go` 19); cada palabra distinta se busca como primera columna de una pareja de `words.go` (28096 parejas) y de `words_us.go` (1618, las que añade `locale: US`, `misspell.go` 53-54). Se hizo con un programa Go de un solo fichero fuera del repositorio, sin ejecutar el linter; da un superconjunto de lo que el linter vería en un `.go`, porque no blanquea URL, rutas ni nombres de sitio (`notwords.go` 103-106) ni descarta las palabras con mayúsculas mezcladas (`replace.go` 232-236). Resultado: 60 palabras marcadas, todas de `words.go`, en cuatro grupos. **(1) Van a `ignore-rules`** (D18): **`administrativo`** (527, «administration»), **`capitulo`** (21070, «capitol»), **`dependencias`** (3330, «dependencies»), **`disposicion`** (6998, «disposition»), **`materias`** (22765, «materials») y **`regulares`** (19592, «regulars»). **(2) Ya están en `ignore-rules`**: `argumentos` (10720), `controles` (17507), `descripcion` (6856), `inventario` (13685), `legislacion` (8117), `reproduccion` (4986) y `versiones` (20323). **(3) Formas sin tilde de palabras que se escriben con ella**: `adaptacion` (10405), `compilacion` (6333), `composicion` (6364), `configuracion` (1328), `conjuncion` (11560), `constitucion` (3124), `construccion` (3133), `convencion` (11739), `correccion` (11795), `corrupcion` (11815), `declaracion` (6735), `distribucion` (3497), `documentacion` (1540), `emision` (25592), `evaluacion` (12580), `historicas` (13109), `identificacion` (710), `implementacion` (717), `informacion` (7830), `justificacion` (1887), `omision` (26355), `orientacion` (8560), `parametros` (14506), `presentacion` (4791), `produccion` (15012), `prohibicion` (9050), `proposito` (19432), `proteccion` (15135), `regeneracion` (4951), `transcripcion` (2398) y `verificacion` (5524). **(4) Palabras españolas correctas que el diccionario marca**: `calcular` (21032), `clientes` (21218), `comando` (25278), `comandos` (21247), `contradice` (11710), `decisiones` (11920), `defectos` (21506), `directorios` (6948), `directos` (21633), `distribuye` (12294), `inaccesibles` (3868), `momento` (26239), `posicional` (14805), `producto` (23463), `recorre` (26678) y `resolverse` (15468). **No** están: `consolidacion`, `publicacion`, `derogacion`, `analisis`, `relacion`, `preambulo`, `seccion`, `metadatos`, `articulo`, `indice`, `vigencia`, `bloque`, ni los verbos que D18 propone en lugar de los sustantivos del grupo (3) (búsqueda directa en `words.go` y `words_us.go`). **`gosec`** v2.28.0: G304 → `NewReadFile` y G306 → `NewWritePerms` (`rules/rulelist.go` 92 y 94); `filepath.Clean` sanea la ruta para G304 (`rules/readfile.go` 157 y 240); G306 admite como máximo `0o600` por omisión (`rules/fileperms.go` 80-81); `os.Getenv` es fuente de propagación en `analyzers/commandinjection.go` 34 (G702) y `analyzers/pathtraversal.go` 35 (G703). El resto (`depguard` por prefijo, `run.build-tags` como lista, G204) quedó verificado en H3 contra el código de esas herramientas (`specs/004-h3-internal-cache-sqlite/research.md`, tabla, filas de `gosec`, `golangci-lint` y `depguard`) |
| `gh` 2.100.0 | `gh issue create --help` (`--title`, `--body`; de `--label` solo dice «-l, --label name  Add labels by name», nada de una etiqueta inexistente: S12); `gh issue comment --help` (`<número>`, `--body`); `gh issue list --help` (`--search`, `--state`, `--json`, `--jq`); `gh help formatting` (`--jq` con sintaxis de jq) |
| Kernel y adaptadores de H1-H3 | `internal/app/main.go` 104 (`Emitir` recibe el `Resultado` también en fallo), 301-320 (plazo de `--timeout` y ensayo), 386-397 (`conPlazoAgotado`); `internal/app/registro.go` 151-163 (`RegistroDeProduccion` vacío); `internal/app/ejemplo/kitlegal-e2e/main.go` 51-58 (el registro inválido revienta con `panic`); `internal/app/e2e_test.go` 113-171 (binario de e2e y guiones); `internal/cli/sobre.go` 176-190 (`montar` fecha con el reloj del montador) y 198-204 (`procedenciaConocida`); `internal/cli/errors.go` 90-112 (`Clasificar`); `internal/cli/describe.go` 272-343 (reflexión con nombres `paquete.Tipo`); `internal/core/schema/sobre.go` 47-54 (`Procedencia{Fuente, URL}`) y 77-91 (`Resultado`); `internal/httpx/peticion.go` 11-16 (`Peticion` con dos campos) y 22-49 (`Respuesta`, `Descripcion`); `cliente.go` 243-252 (orden de la cadena), 278-298 (cadena de `Replay`), 412-426 (`Pedir` y ensayo), 516-550 (`emitir`), 615-638 (`entregar`: todo 4xx salvo el 429 se **entrega** a quien llama); `robots.go` 126 y 184-190 (la petición del `robots.txt` usa el mismo contexto); `reintentos.go` 82-98 (cada intento es un clon con el mismo contexto); `grabar.go` 171-179 (las cabeceras de la petición se graban); `reproducir.go` 106-110 (la reproducción empareja por método y dirección, sin cabeceras); `nombre.go` 47-57 (nombre del fichero de grabación, recorte a 120); `internal/cache/entradas.go` 77-128 (`Get` en solo lectura devuelve la ausencia como error de clase 4) y 148-193 (`Put`); `internal/cache/errores.go` (`Error`, clases 1, 2, 4); `internal/arch_test.go` 105-122 (`modulosDelBinario`), 146-204 (los dos tests que H4 retira); `docs/PENDIENTES.md` («Antes de la primera tarea `[datos]` de H4» y «En H4») |
| Módulos que enlazaría el binario | `go list -deps -f "{{if .Module}}{{.Module.Path}}{{end}}" ./internal/cache ./internal/httpx ./cmd/kitlegal` (hoy): además de los seis de `modulosDelBinario`, `golang.org/x/time`, `github.com/temoto/robotstxt`, `modernc.org/sqlite`, `modernc.org/libc`, `modernc.org/mathutil`, `modernc.org/memory`, `golang.org/x/sys`, `github.com/dustin/go-humanize`, `github.com/google/uuid`, `github.com/mattn/go-isatty`, `github.com/ncruces/go-strftime`, `github.com/remyoudompheng/bigfft`. Es orientación: la lista que vale es la que mida `TestDependenciasDelBinario` al enlazar (S10) |
| Skills `golang-*` | `.agents/skills/golang-{how-to,testing,cli,project-layout}/SKILL.md` ([D19](#d19--aplicación-de-las-skills-golang--instaladas)) |

---

## D1 · Reparto: puerto en `core`, adaptador en `internal/source/boe`, applet en `internal/app`

**Decisión.**

- `internal/core/source.go` declara el puerto `core.Source` y sus dos tipos auxiliares (D2).
- `internal/source/boe` es el **adaptador**: implementa `core.Source`, construye las direcciones, pide por
  `internal/httpx`, lee XML y JSON, deriva los avisos, lee y escribe la caché por el puerto `core.Cache` y
  declara los tipos de `data`. No conoce la línea de órdenes: no lleva etiquetas de Kong, no importa
  `internal/app`, `internal/cli` ni `internal/cache` (recibe la caché ya abierta por una función).
- `internal/app/boe.go` es el **applet** (Command del roadmap §2): nombre, descripción, los seis verbos con
  sus argumentos etiquetados, y la composición por invocación de la fuente con sus dependencias
  (`DependenciasDeBoe`). `RegistroDeProduccion` lo registra con las dependencias de red.
- El binario de e2e (`internal/app/ejemplo/kitlegal-e2e`) registra, además de los ejemplos, el mismo applet
  con un cliente de reproducción ([D13](#d13--tests-e2e-y-verificación-reproducción-en-el-binario-de-e2e-etiquetas-fuentes-y-grabacion)).

**Por qué.** `app.Verbo` es un tipo de `internal/app`: el paquete que declare los verbos tiene que
importarlo, y `RegistroDeProduccion` vive en `internal/app` (ADR 0005). Si el catálogo viviera en
`internal/source/boe`, `app → source/boe → app` sería un ciclo. Poner la gramática en `internal/app` deja el
adaptador limpio de CLI —es el Adapter de roadmap §2, y la futura herramienta MCP lo reutilizará sin Kong— y
es exactamente lo que el roadmap asigna a `internal/app` («registro de applets, Command + Registry»).

**Alternativas rechazadas.** *Catálogo en `internal/source/boe`*: ciclo de importación. *Subpaquete
`internal/app/boe`*: mismo ciclo con `RegistroDeProduccion`, salvo mover el registro de producción fuera de
`internal/app`, contra ADR 0005. *Etiquetas de Kong en los tipos del adaptador*: filtra la gramática al
adaptador y la ata a la línea de órdenes.

**Verificado en.** `internal/app/applet.go` 22-73; `internal/app/registro.go` 151-163;
`internal/app/ejemplo/ejemplo.go` 44-54 (el patrón del registro fuera de `app` para el binario de e2e).

---

## D2 · El puerto `core.Source`: `Name`, `Fetch(ctx, ec, consulta)`, `TTL(consulta)`, `Terms`

**Decisión.**

```go
// internal/core/source.go — dominio puro: importa context, time y internal/core/schema.
type Source interface {
    Name() string
    Fetch(ctx context.Context, ec schema.Contexto, consulta Consulta) (schema.Resultado, error)
    TTL(consulta Consulta) time.Duration
    Terms() Terminos
}
type Consulta interface{ Verbo() string }
type Terminos struct {
    URL       string    // términos de uso de la fuente
    Revisados time.Time // fecha (00:00 UTC) en que una persona los revisó
}
```

- `Fetch` recibe el contexto de ejecución del kernel como segundo parámetro, igual que `httpx.Pedir`: de él
  salen `--offline` y `--dry-run`, que la fuente honra (FR-092, FR-094). Un fallo devuelve **también** un
  `Resultado` con la procedencia de la petición que falló (FR-101, D3).
- `TTL` recibe la consulta porque la vigencia es **por verbo** (FR-091); `boe` responde 300 s para `buscar` y
  `metadatos` y 7 días para `indice`, `articulo`, `articulos` y `analisis`.
- `Terms` declara la URL de los términos y la fecha de revisión (FR-121), que viven en `terminos.go`; un test las ata a la fila de
  `docs/SOURCES.md` ([D15](#d15--la-fuente-docssourcesmd-ritmo-términos-robotstxt-y-verificación-nocturna)).
- Las consultas de `boe` son tipos del adaptador (`ConsultaBuscar`, `ConsultaIndice`, `ConsultaArticulo`,
  `ConsultaArticulos`, `ConsultaMetadatos`, `ConsultaAnalisis`); un tipo de consulta ajeno es un fallo
  inesperado (defecto del applet, nunca de quien invoca).
- El registrador de eventos **no** viaja por el puerto (el dominio no importa `log/slog`, R1): se da a la
  fuente al construirla, una por invocación (`boe.ConRegistrador`).
- **ADR 0015** registra la firma concreta, porque `CLAUDE.md` escribe `Source{Name, Fetch(ctx,req), TTL,
  Terms}` sin fijar parámetros y el spec permite ADR si el plan concreta el puerto (*Fuera de alcance*,
  «ADR nuevo»).

**Alternativas rechazadas.** *`Fetch(ctx, req)` con el contexto de ejecución dentro de `req`*: mezcla lo
que se pide con cómo se invoca y rompe la simetría con `httpx.Pedir`. *`TTL()` sin argumentos*: una sola
vigencia por fuente contradice FR-091. *Consulta como `any` o `map[string]string`*: pierde el tipo y obliga
a validar formas en tiempo de ejecución. *Registrador en `Fetch`*: `internal/core` no puede importar
`log/slog` (depguard lista `core`, `compruebaDominioPuro`).

**Verificado en.** `internal/core/doc.go` («Los puertos que todavía no existen —Source…— nacerán aquí»);
`.golangci.yml` lista `core` (deniega `log/slog`); `internal/httpx/cliente.go` 412.

---

## D3 · La fecha de consulta la declara quien consulta: `schema.Procedencia.FechaConsulta`

**Decisión.**

- `schema.Procedencia` gana un tercer campo, `FechaConsulta time.Time`. **Valor cero = el kernel fecha con
  su reloj al montar**, que es exactamente lo que hace hoy: retrocompatible con los applets de ejemplo, con
  el adaptador de prueba de H3 y con todo sobre de fallo anterior a una petición (FR-096: código 2 y
  `--offline` sin entrada). `Procedencia.Validar` no cambia (la fecha no es obligatoria).
- `cli.Montador.montar` usa `proc.FechaConsulta` cuando no es cero y `m.instante()` cuando lo es, **en éxito y
  en fallo** por el mismo camino (ADR 0006: «se calculan igual»). `procedenciaConocida` sigue sustituyendo
  una procedencia inválida por la del kernel, que no trae fecha.
- `boe` rellena la fecha con el **instante de emisión** que le da `internal/httpx` (D4) y, cuando el `data`
  sale de la caché, con el instante guardado en la entrada (D5). Si `data` se sostiene sobre varias
  consultas, lleva el **más antiguo** (FR-096): `articulo` = mín(bloque, metadatos); `articulos` = mín de sus
  bloques; en un fallo, el de la petición que falló.
- La huella no cambia: sigue calculándose sobre `data`, que no lleva la fecha (ADR 0006).
- **ADR 0015** lo registra (el spec lo exige si FR-096 amplía `Procedencia` o `Resultado`).

**Alternativas rechazadas.** *Campo en `Resultado`* (`Resultado.FechaConsulta`): la fecha califica la
procedencia —«de dónde y cuándo»— y en H7 la `Procedencia` es el `source` de cada nodo del grafo (ADR 0014),
que necesitará el instante para `fuente-caducada`; separarlas obligaría a casar dos valores. *Fecha dentro
de `data`*: duplica un campo del sobre y hace depender la huella del instante (ADR 0006). *Reloj del kernel
siempre*: hace pasar por recién consultado lo servido de la caché (spec, Assumptions). *Que el applet monte el
sobre*: prohibido por ADR 0005.

**Verificado en.** `internal/cli/sobre.go` 176-190 y 198-204; `internal/core/schema/sobre.go` 47-71;
`internal/app/main.go` 104 (`Emitir` recibe el `Resultado` también en fallo, así que la procedencia de un
fallo ya llega al montador); `go doc time.Parse` (ida y vuelta del texto RFC 3339 sin cambios).

---

## D4 · `internal/httpx` gana tres cosas: `Peticion.Acepta`, `Instante` y `ConHora`

**Decisión.**

1. **`Peticion.Acepta string`**: el tipo de contenido que se pide en la cabecera `Accept`. Vacío = sin
   cabecera (comportamiento de H2 intacto). Se valida antes de abrir nada con `mime.ParseMediaType` y
   rechazando cualquier carácter de control: inválido → «argumentos» (2), sin petición. `emitir` la pone en la
   petición del recurso y en cada salto de redirección; **no** la lleva la petición del `robots.txt`. La
   grabación la guarda porque guarda las cabeceras de la petición; la reproducción sigue emparejando solo por
   método y dirección (contrato de grabación de H2 §4 intacto).
2. **`Respuesta.Instante` y `Error.Instante` (`time.Time`)**: el instante de emisión de la petición a la
   fuente —el del **último intento** si hubo reintentos, y el del último salto si hubo redirecciones—; si no
   llegó a emitirse (denegada por `robots.txt`, sin turno en el ritmo, plazo agotado antes de salir,
   argumentos inválidos), el instante en que `Pedir` la abandonó. En ensayo (`Respuesta.Ensayo`) va a cero:
   no hubo consulta.
   - Mecanismo: un escalón nuevo, `conMarcaDeEmision`, justo **encima del transporte** (de red y de
     reproducción) y debajo de la grabación, anota `hora()` en una marca que `Pedir` pone en el contexto con
     una clave sin exportar. Cada salto la reinicia; cada intento la sobrescribe. El decorador de `robots.txt`
     pide su fichero con un contexto que **oculta** la marca, para que su obtención no pase por emisión de la
     petición. Al volver, `Pedir` copia la marca en `Respuesta.Instante` o en `Error.Instante`, o pone `hora()`
     si está vacía.
3. **`ConHora(func() time.Time) Opcion`**: de dónde sale ese instante; por omisión `time.Now`; nula →
   «argumentos». Vale para `New` y para `Replay` (tiene sentido en los dos).

**Por qué.** FR-003 exige pedir XML para el bloque y JSON para lo demás, «el formato que pide `boe.py`», que lo
hace por la cabecera `Accept` (`refs/boe.py` 77-79 y 405-407); H2 dejó `Peticion` sin cabeceras. FR-096 y
SC-008 definen el instante de una consulta como el de emisión del último intento o el de abandono, que solo
se conoce **dentro** de la cadena: los reintentos viven por debajo de `emitir`. Un valor en el contexto es
la forma idiomática de llevar un dato de una petición a través de decoradores `http.RoundTripper` que no se
pueden cambiar de firma (`golang-context`: valores de ámbito de petición; `go doc context.WithValue`: «request-scoped
data», con una clave de un tipo propio y no predefinido).

**Alternativas rechazadas.** *Medir el instante en el adaptador al volver de `Pedir`*: es el de recepción, no
el de emisión, y con reintentos y ritmo difiere en segundos (contra SC-008). *Una cabecera sintética en la
respuesta*: ensucia el contrato de `Respuesta.Cabeceras` y la grabación. *Cambiar la firma de `RoundTrip`*:
imposible (interfaz de la biblioteca). *Un campo `Cabeceras` general en `Peticion`*: abre la puerta a
sobrescribir la identificación, que H2 prohíbe (FR-008 de H2); `Acepta` es la única cabecera que una fuente
necesita elegir. *Emparejar la reproducción también por `Accept`*: cambia el contrato de H2 sin necesidad
—`boe` nunca pide la misma dirección en dos formatos— y rompería las grabaciones existentes.

**Verificado en.** `internal/httpx/peticion.go` 5-16 («No lleva cabeceras…»); `cliente.go` 243-252,
278-298, 432-451, 516-550; `robots.go` 126, 159 y 184-190; `reintentos.go` 82-98; `grabar.go` 171-179;
`reproducir.go` 106-110; `go doc mime.ParseMediaType`; `go doc context.WithValue`.

---

## D5 · Caché: claves, forma de la entrada, vigencias, metadatos compartidos y modos

**Decisión.**

- **Clave** = `boe.legislacion-consolidada|1|<verbo>|<dirección de la API del recurso>`. El `1` es la versión
  del formato de la entrada: si cambia la forma de `data`, sube, y las entradas viejas dejan de encontrarse
  (nunca se leen con otra forma). La dirección lleva todos los argumentos que cambian la respuesta (FR-090):
  en `buscar`, la consulta construida y el límite; en `articulo`, norma y bloque.
  - `articulo` usa la dirección del bloque; `metadatos` (el verbo y la lectura que hacen `articulo` y
    `articulos`) usa la de los metadatos; `articulos` **no tiene entrada propia**: lee y escribe la de
    `articulo` de cada bloque (FR-020).
- **Contenido** (JSON, sin campos desconocidos al leer: `Decoder.DisallowUnknownFields`): `{"fecha_consulta": "<RFC 3339>", "url": "<dirección
  del sobre>", "datos": <data del verbo>}`. `fecha_consulta` es la que llevó o llevaría el sobre de la
  consulta que la escribe (FR-096); servirla da el mismo texto byte a byte (`go doc time.Parse`). Una entrada
  de la clave correcta que no se puede leer es un fallo «inesperado» (1) que nombra la clave: con la versión en
  la clave solo puede ser corrupción o un defecto, y no se tapa.
- **Vigencias** (`Fuente.TTL`, FR-091): `buscar` y `metadatos` 300 s; `indice`, `articulo` y `analisis`
  604 800 s. La entrada de metadatos que escriben `articulo` y `articulos` usa la de `metadatos`.
- **Metadatos compartidos** (FR-090, FR-020): `articulo` lee la entrada de `metadatos` antes de pedirlos y, si
  los pide y los obtiene, la escribe. `articulos` los resuelve **una vez** por invocación (en memoria tras la
  primera lectura o petición) y solo si falta en caché algún bloque.
- **Qué se escribe y cuándo** (FR-093, FR-021): solo lo que la fuente respondió y se interpretó. `buscar` sin
  resultados se escribe (FR-032). En `articulos`, cada bloque se escribe en cuanto se resuelve; el que falla,
  no.
- **Modos** (FR-092, FR-094): la caché se abre en **solo lectura** con `--offline` **o** con `--dry-run`, y en
  normal en otro caso. En solo lectura, `cache.Get` devuelve la ausencia como error de clase
  «fuente no disponible» (H3 FR-016). Con `--offline` ese error se propaga (código 4). Con `--dry-run` sin
  `--offline`, un error de esa clase **con el contexto vivo** se trata como ausencia y la petición se describe
  en el ensayo; con el contexto terminado se propaga (el kernel lo convierte en 4 igualmente).
- **Apertura perezosa del cliente HTTP**: la fuente recibe una función que lo construye y solo la llama si hay
  que pedir algo. Lo servido de la caché no construye ningún cliente (y por eso el e2e puede quitar la
  reproducción y seguir respondiendo desde la caché, D13).

**Por qué.** La dirección de la API ya es la identidad exacta de lo pedido; construir la clave con ella
garantiza «dos consultas distintas nunca comparten entrada» sin inventar una segunda codificación de
argumentos. Abrir en solo lectura bajo `--dry-run` es lo único que garantiza que un ensayo no cree
`~/.cache/kitlegal/cache.db` (la apertura normal crea y migra, H3 D4); distinguir la ausencia por la clase y
el contexto sigue la tabla cerrada de H3 (en solo lectura, la clase 4 solo sale por ausencia o por contexto
terminado).

**Alternativas rechazadas.** *Clave con el texto crudo como `boe.py`* (`md5(cmd:args)`, líneas 49-66): dos
textos que producen la misma consulta no compartirían entrada, y el MD5 no aporta nada a una clave opaca.
*Guardar el cuerpo crudo y reinterpretarlo al servir*: repite el trabajo en cada lectura y, si el lector
cambia, lo guardado devolvería otra cosa con la misma fecha. *Formato con versión dentro del contenido*: una
entrada vieja se encontraría y habría que decidir qué hacer con ella; con la versión en la clave simplemente
no está. *Ensayo sin mirar la caché*: describiría peticiones que no se emitirían. *Ensayo con la caché en modo
normal*: crea la base (efecto) en una invocación que promete no hacer nada. *Tratar como ausencia una entrada
ilegible*: silenciaría una corrupción durante siete días.

**Verificado en.** `go doc encoding/json.Decoder.DisallowUnknownFields`; `internal/cache/entradas.go` 64-68 y 77-128; `internal/cache/cliente.go` 62-72
(`SoloLectura` no crea nada); `internal/core/cache.go`; `refs/boe.py` 37-66, 186, 282-283, 302, 411-416,
433, 440-441.

---

## D6 · Lectura del bloque XML: el mismo recorrido que `ElementTree`

**Decisión.** `leerBloque(cuerpo []byte)` con `encoding/xml` en modo estricto:

1. El cuerpo tiene que ser UTF-8 válido; si no, ilegible (4). `boe.py` decodifica siempre como UTF-8 antes de
   analizar (línea 82) y su fallo acaba en el análisis de emergencia que no se porta (FR-014).
2. Se ignora la declaración `encoding` (un `CharsetReader` que devuelve la entrada tal cual), porque `boe.py`
   analiza la cadena ya decodificada (S5).
3. El documento tiene que tener **un único elemento raíz** y nada más que espacio en blanco, comentarios o
   instrucciones de proceso fuera de él; si no, ilegible (4), como hace expat (S5 (f)). `encoding/xml` no lo exige
   por sí mismo al leer por fichas: `Decoder.Token` solo garantiza que las fichas de inicio y fin estén anidadas y
   casadas (`go doc encoding/xml.Decoder.Token`), al final de la entrada solo es error que quede un elemento abierto
   (`xml.go` 286-287) y su pila no cuenta elementos raíz (`xml.go` 440-500), así que un segundo elemento, o texto,
   tras cerrar el primero llegan como fichas sin error. Por eso lo comprueba la lectura.
4. El bloque es el **primer descendiente** llamado `bloque` (sin espacio de nombres) **que no sea la raíz**
   (`ElementPath.py` 182-209). Si no hay ninguno: ilegible (4) (FR-014).
5. `titulo` y `tipo` son sus atributos (vacíos si faltan; líneas 106-107). El valor de un atributo se
   normaliza como manda XML 1.0 §3.3.3 para atributos CDATA —tabulador, salto y retorno → espacio—, que
   `encoding/xml` no hace (`xml.go` 860-868) y expat sí (S5).
6. Las versiones son los **hijos directos** `version` (línea 109). Sin versiones: texto de todo el bloque,
   fecha «original», vigencia y modificadora vacías (líneas 110-112). Con versiones: la **última**; fecha =
   `fecha_publicacion` si el atributo **existe** (aunque venga vacío, como `dict.get`) y «original» si no;
   `fecha_vigencia` e `id_norma`, vacíos si faltan (líneas 114-118).
7. El texto de un elemento es `itertext()` **más su propio `tail`** (`ElementTree.py` 406-423 y 979-983): la
   concatenación de todo `CharData` del elemento y sus descendientes, en orden, más el `CharData` que sigue a
   su cierre hasta la siguiente etiqueta. Los comentarios y las instrucciones de proceso no aportan texto ni cortan
   el que los rodea: el `TreeBuilder` de Python puro se crea con `insert_comments=False` e `insert_pis=False`
   (`ElementTree.py` 1413-1415), y sus `comment` y `pi` (1486-1501) acaban en `_handle_single`, que sin inserción ni
   añade el nodo ni vacía el texto acumulado (1503-1511; `_flush`, 1439-1449), de modo que el texto de los dos lados
   se une en el mismo `text` o `tail`. `ET.fromstring` usa el acelerador C `_elementtree` cuando existe (2085-2086),
   cuya fuente no está en local: que se comporte igual es el supuesto S5 (e). `CDATA` aporta su contenido:
   `encoding/xml` lo entrega como `CharData` (`xml.go` 699-715).
8. Normalización de líneas (líneas 132-136): partir por `\n`, recortar cada línea con el conjunto de espacio
   en blanco de Python (`unicode.IsSpace`, que no incluye U+001C–U+001F según `go doc unicode.IsSpace`, **más**
   U+001C–U+001F, S5 (a)), descartar las vacías y unir con `\n`.
   `\r\n` ya llega como `\n` (`xml.go` 1130-1136, igual que expat).

**Por qué.** El diff de aceptación (FR-116) compara título, tipo, fechas, modificadora y texto con lo que
`xml_bloque_to_text` produciría; cada paso anterior es una propiedad observable de `ElementTree` que, si no se
porta, cambia el texto en casos reales (el `tail` y la exclusión de la raíz) o en bordes (atributos con saltos,
declaración de codificación).

**Alternativas rechazadas.** *`xml.Unmarshal` sobre structs*: pierde el orden de texto mezclado con hijos y el
`tail`. *Expresiones regulares*: son el `_fallback_parse` que FR-014 prohíbe. *Aceptar la raíz `bloque`*:
`find(".//bloque")` no la encuentra; `boe.py` iría a la rama «sin bloque». *Respetar la declaración de
codificación*: `boe.py` no lo hace.

**Verificado en.** CPython `xml/etree/ElementTree.py` 406-423, 979-983, 1334-1347, 1413-1415, 1439-1449, 1486-1511 y
2085-2086, y `xml/etree/ElementPath.py` 182-209; GOROOT `src/encoding/xml/xml.go` 147, 188-193, 275-290, 440-500,
644-656, 699-715, 860-868 y 1130-1136; `go doc encoding/xml.Decoder.Token`; `go doc unicode.IsSpace`; `refs/boe.py`
95-136.

---

## D7 · Lectura de las respuestas JSON

**Decisión.** `encoding/json` con `UseNumber` sobre un valor genérico, y lecturas explícitas:

- **Envoltorio**: la raíz tiene que ser un objeto; si no, o si no es JSON, ilegible (4) (`boe.py` cae en
  `AttributeError`/`JSONDecodeError`, líneas 274-280, 357-360, 454-457, 513-516). `data` ausente o `null`
  equivale a `[]` (`get("data", [])`). «Vacío» es lista vacía, objeto vacío, cadena vacía o `null`, como la
  veracidad de Python (`if not data`).
- **Objeto suelto → lista de uno** (FR-070) en: resultados de `buscar`; `data` de `indice` y su `bloque`
  anidado; `materias`; `referencias.anteriores` y `posteriores`; y el contenido de los envoltorios `materia`,
  `anterior` y `posterior`.
- **Primer elemento o el objeto** (líneas 195, 463 y 522) en metadatos y análisis.
- **Envoltorios portados tal cual**: materia `{materia: …}` o directa (531); notas `{nota: …}` o directas
  (540); referencias `{anterior|posterior: […]}` o directas (550-554, 570-574); índice anidado en `data[0].bloque`
  o plano (365-369).
- **Elementos que no son objeto** en listas de resultados, bloques o referencias: se saltan, como `boe.py`
  (líneas 287, 386, 556, 576). Una materia que no es objeto aporta su texto (línea 535).
- **Campos de texto**: cadena → su valor; ausente o `null` → vacío (FR-016). `rango`, `estado_consolidacion` y
  `relacion`: objeto → su `texto` (y `codigo` en el estado de `metadatos`); cadena → la cadena (el `str(x)` de
  las líneas 466, 469, 559, 579); en `buscar`, un estado o rango que no es objeto → vacío (líneas 289 y 291,
  que dan «?»). **Cualquier otro tipo** donde se espera texto (número, booleano, lista, objeto) → ilegible (4)
  nombrando el campo.
- **Avisos** (FR-012): se leen del objeto de metadatos con las comparaciones exactas de `boe.py`:
  `estado_consolidacion.codigo == "4"` (cadena), `estatus_derogacion == "S"`, `vigencia_agotada == "S"`; en ese
  orden y con los textos literales de las líneas 202-211.

**Por qué.** Son las lecturas de `boe.py` donde el spec no declara desviación (FR-060, FR-070), más la regla
mínima que exige tipar `data`: `boe.py` imprime el `repr` de Python de un valor inesperado (p. ej.
`{'codigo': 1}` como título), cosa que no puede ir en un campo de texto de un esquema y que la constitución
§II impide presentar como contenido de la fuente. La verificación nocturna (FR-115) avisa si la fuente cambia
de forma.

**Alternativas rechazadas.** *Convertir números y booleanos a su literal*: el literal JSON (`true`) no es lo
que imprime `boe.py` (`True`), así que tampoco sería «el mismo comportamiento», y enmascara un cambio de forma
de la fuente. *Omitir en silencio el campo raro*: presentaría la respuesta como completa sin serlo (el mismo
motivo de FR-070). *Structs con `json:"…"` y `Unmarshal` directo*: no admite a la vez objeto y lista ni los
envoltorios opcionales sin tipos intermedios frágiles.

**Verificado en.** `refs/boe.py` 181-225, 272-303, 355-396, 452-502, 511-585; `go doc encoding/json.Decoder.UseNumber`
(un número se lee como `json.Number`, no como `float64`).

---

## D8 · `buscar`: la misma consulta y la misma dirección, byte a byte

**Decisión.**

- Texto = argumentos unidos por un espacio (línea 784). Si contiene ` AND `, ` OR `, ` NOT `, `titulo:`,
  `materia:` o `"`, se envía tal cual (258-262). Si no, palabras = partir por espacio en blanco de Python
  (`unicode.IsSpace` más U+001C–U+001F, como `str.split()`, S5); cero palabras → «argumentos» (2) sin petición
  (FR-030); una o más → `titulo:<p1> AND titulo:<p2>…` (FR-030: con una sola palabra, `titulo:<palabra>`
  recortada).
- Cuerpo de la consulta = exactamente `json.dumps({"query": {"query_string": {"query": q}}})`:
  `{"query": {"query_string": {"query": "<q escapada>"}}}` con separadores `': '` y `', '`, `"` → `\"`, `\` →
  `\\`, `\n \r \t \b \f` → sus escapes cortos, el resto de controles y todo lo no ASCII → `\uxxxx` en minúsculas
  (pares sustitutos fuera del plano básico).
- Dirección = `https://www.boe.es/datosabiertos/api/legislacion-consolidada?limit=10&query=` + `quote(cuerpo)`:
  UTF-8 y todo byte fuera de `A-Za-z0-9_.-~/` como `%XX` en mayúsculas.
- Se codifica a mano (`busqueda.go`), con una tabla de test contra salidas escritas a partir de las reglas de
  las fuentes de CPython citadas.

**Por qué.** «Los mismos recursos pedidos con los mismos parámetros» (resumen del spec). Ninguna función de la
biblioteca estándar reproduce la dirección de `boe.py`: `encoding/json` escapa `<`, `>` y `&` (`go doc
encoding/json.Marshal`: las cadenas pasan por `HTMLEscape`), deja sin escapar lo no ASCII (GOROOT
`src/encoding/json/encode.go` 998-1064: `appendString` solo escapa comillas, barra invertida, controles, `<>&` y
U+2028/U+2029, y copia el resto de runas tal cual, 1060-1062) y no pone espacios tras `:` ni `,` (`encode.go`
797-803); `url.QueryEscape` codifica el espacio como `+` (`src/net/url/url.go` 186-188 y 196-227, líneas 200 y
225-226); `url.PathEscape` deja `:`, `@`, `&`, `=`, `+` y `$` sin codificar (`url.go` 192-194;
`src/net/url/gen_encoding_table.go` 179 y 190-193: en un segmento de ruta, de los reservados solo se escapan `/`, `;`,
`,` y `?`).

**Alternativas rechazadas.** *`json.Marshal` + `url.QueryEscape`*: otra dirección, otra grabación y, para la
fuente, otra consulta. *Dejar el texto sin recortar con una palabra*: FR-030 lo prohíbe.

**Verificado en.** `refs/boe.py` 252-271, 783-784; CPython `json/encoder.py` 31, 48-66, 102-103,
`json/__init__.py` 183, `urllib/parse.py` 790, 815, 890; GOROOT `src/encoding/json/encode.go` 797-803 y 998-1064,
`src/net/url/url.go` 79-82 y 186-227, `src/net/url/gen_encoding_table.go` 175-211; `go doc encoding/json.Marshal`;
`go doc net/url.URL.String` (la consulta se conserva tal cual al emitir).

---

## D9 · Identificadores de entrada: gramáticas, tipo inferido y fuzz

**Decisión.**

- **Norma** (FR-081): `^BOE-A-[0-9]{4}-[0-9]{1,9}$`, sensible a mayúsculas. Otra forma → 2 antes de nada.
- **Bloque** (FR-080): `^[A-Za-z0-9]{1,64}$`. Letras y dígitos ASCII, de 1 a 64. Todo lo demás —vacío, `/`,
  `?`, `#`, `%`, espacios, controles, no ASCII, más de 64— → 2 antes de nada. **Atada a los índices grabados**:
  `TestGramaticaCubreLosIndicesGrabados` recorre los `id` de los tres índices grabados y exige que la
  gramática los acepte todos; si una grabación trae un id fuera de ella, el test falla y la gramática se
  revisa **dentro** de FR-080 (si hiciera falta un carácter que FR-080 no admite, es un conflicto con el
  spec y se escala).
- **Tipo inferido** (`TipoDesdeID`, FR-040), porte literal de `_tipo_from_id` (155-176) en el mismo orden:
  minúsculas Unicode (`strings.ToLower`, «all Unicode letters»), índices por runa, «dígito» = `unicode.IsDigit` («a
  decimal digit») para `^a\d` y para el
  segundo carácter de `s`; devuelve `articulo`, `titulo`, `capitulo`, `seccion`, `preambulo`,
  `disposicion_adicional`, `disposicion_transitoria`, `disposicion_derogatoria`, `disposicion_final` o vacío.
  Se aplica a los `id` que entrega la fuente en `indice`, que no pasan por la gramática.
- **Fuzz** (FR-082): `FuzzIDDeBloque` con semillas `a21`, `da3`, `dt1`: nunca entra en pánico; si
  `ValidarBloque` acepta, `url.PathEscape(id) == id`, no contiene `/?#%`, mide ≤ 64 y la dirección del bloque
  construida y analizada con `url.Parse` termina exactamente en `/<id>`; `TipoDesdeID` devuelve lo mismo dos
  veces. En `make test` corren las semillas; el fuzz ligero se lanza a mano (quickstart).

**Por qué.** FR-080 deja la gramática exacta al plan «contra los índices grabados»; como el plan se escribe
antes de grabar, la atadura es un test sobre las grabaciones y no una afirmación. 64 cubre con holgura los
ids del BOE (`a21`, `da3`, `preambulo`, `a108bis`) y acota «longitud desmedida».

**Alternativas rechazadas.** *Admitir `-` y `_`*: FR-080 dice letras y dígitos. *Solo minúsculas*: rechazaría
un id legítimo si la fuente usara mayúsculas, sin ganar seguridad. *`regexp` con `\d`*: en Go `\d` es solo
ASCII (`go doc regexp/syntax`: «Perl character classes (all ASCII-only): \d digits (== [0-9])») y `_tipo_from_id`
usa `\d` e `isdigit()` de Python, que son Unicode (S5 (b)).

**Verificado en.** `refs/boe.py` 155-176, 186, 355, 406, 452, 511; `go help testflag` 77-90; `go doc regexp/syntax`;
`go doc unicode.IsDigit`; `go doc strings.ToLower`.

---

## D10 · Clasificación de fallos y sobre de fallo

**Decisión.** Tabla cerrada en [contracts/errores-y-codigos.md](./contracts/errores-y-codigos.md). En resumen:

- Estado HTTP que `httpx` entrega: 2xx → se interpreta; **404** → «no encontrado» (3) en índice, bloque,
  metadatos y análisis, y «fuente no disponible» (4) en `buscar`; cualquier otro (4xx salvo 429) → 4 con el
  estado en el mensaje. Los fallos que ya trae `httpx` (429 y `robots.txt` → 5; 5xx, plazo y transporte → 4)
  pasan con su clase.
- `data` vacío en índice, metadatos o análisis → 3 (FR-041, FR-051, FR-061). XML sin bloque o ilegible, JSON
  ilegible o con tipos inesperados → 4 (FR-014, D7).
- `boe.Error{URL, Instante, Causa}` implementa `schema.ConClase` (como `httpx.Error` y `cache.Error`); los
  errores de `httpx` y de la caché se envuelven con `%w` y conservan su clase (`cli.Clasificar` usa `errors.As`).
- El `Resultado` de un fallo lleva `Procedencia{Fuente: "boe.legislacion-consolidada", URL: <la petición que
  falló>, FechaConsulta: <su instante>}` (FR-101, FR-096). En `--offline` sin entrada: URL del recurso que
  falta y fecha cero. En argumentos inválidos: `Resultado{}` (el kernel firma `kitlegal.cli` / `kitlegal:cli`).
- La URL de un fallo es siempre la **de la API que `boe` pidió**, nunca la del `robots.txt` que `httpx` nombra
  en su mensaje cuando el permiso no se obtiene.

**Por qué.** `httpx` entrega el 404 a quien llama porque solo la fuente sabe qué significa (H2 FR-032). Que un
404 de búsqueda sea 4 y no 3 sale de FR-032: una búsqueda nunca es «no encontrado».

**Alternativas rechazadas.** *403 → 5*: ni FR-100 ni la fuente lo declaran límite o términos; se trataría como
tal un bloqueo que puede ser pasajero. *Leer el `status.code` del envoltorio además del HTTP*: sin grabación que
lo muestre sería código sin caso; se decide sobre el estado HTTP y se verifica con la grabación de lo
inexistente (S2).

**Verificado en.** `internal/httpx/cliente.go` 615-638; `internal/cli/errors.go` 90-112;
`internal/cache/errores.go`.

---

## D11 · Tipos de `data`, claves JSON y `schemas/`

**Decisión.**

- Tipos en `internal/source/boe/datos.go` ([data-model.md](./data-model.md) §2): `ResultadoDeBusqueda`,
  `Indice`, `EntradaDeIndice`, `Articulo`, `Aviso`, `Metadatos`, `EstadoDeConsolidacion`, `Analisis`, `Materia`,
  `Referencias`, `ReferenciaAnterior`, `ReferenciaPosterior`. Claves en español y **todas obligatorias** (sin
  `omitempty`): un campo ausente en la fuente va vacío (FR-016), nunca falta. Listas siempre `[]`, nunca `null`.
- Enumerados en la etiqueta `jsonschema`: `Aviso.codigo` con exactamente los tres valores (FR-012) y
  `EntradaDeIndice.tipo` con los nueve tipos más la cadena vacía; `hash_texto` con `pattern=^sha256:[0-9a-f]{64}$`.
  Un test compara cada enumerado con la lista de constantes del paquete.
- **`schemas/norma.json`** (`buscar`, `indice`, `metadatos`, `analisis`) y **`schemas/bloque.json`**
  (`articulo`, `articulos`): un documento del borrador 2020-12 con `$id`
  `https://ventanillalegal.es/schemas/<fichero>` y un `$defs` por verbo cuyo valor es **el documento que emite
  `kitlegal boe <verbo> --describe` con una sola clave añadida**, `$id`
  `https://ventanillalegal.es/schemas/<fichero>/<verbo>`. Así cada parte es un recurso embebido y sus
  `#/$defs/…` internas resuelven contra ella (`roots.go` 104-124, `root.go` 34). Serialización: claves
  ordenadas (`json.Marshal` ordena las claves de los mapas), dos espacios de sangría, sin escape HTML
  (`Encoder.SetEscapeHTML(false)`) y salto final.
- **Validación en test** (FR-111): se lee el fichero, se compila `…/<fichero>#/$defs/<verbo>/properties/salida`
  con `AssertFormat` y se valida el sobre real; nunca contra lo que emite `--describe` en ese momento.
- **`make schema-check`** = `go test -count=1 -run '^TestEsquemasPublicados$' ./internal/app/`. El test pide
  `--describe` de los seis verbos al registro de producción, arma los dos ficheros en memoria y los compara:
  por verbo (falla con «schemas/<fichero>: la parte de «<verbo>» no coincide con `kitlegal boe <verbo>
  --describe`») y el fichero entero (falla nombrando el fichero si difiere fuera de las partes). Con
  `-args -actualizar-esquemas` los escribe; `make ci` nunca pasa esa bandera. `TestEsquemasCubrenTodosLosVerbos`
  exige que cada verbo del registro de producción esté en exactamente un fichero.

**Por qué.** Q1 fija dos ficheros por entidad y verbo separado; «generados desde `--describe`» se cumple
literalmente si la parte **es** lo emitido. Un recurso embebido con `$id` es la forma estándar de juntar
documentos independientes sin reescribir sus `$ref`.

**Alternativas rechazadas.** *Reescribir los `$ref` a `#/$defs/<verbo>/$defs/…`*: la parte ya no es lo
emitido. *Un fichero por verbo*: contradice Q1. *Generar con un script y `jq`*: segunda herramienta, no
garantizada en todos los puestos, y otra serialización que vigilar. *Un binario generador en `cmd/`*: un
segundo `main` con su excepción de lint para escribir dos ficheros que un test ya sabe construir.

**Verificado en.** `go doc encoding/json.Marshal`; `go doc encoding/json.Encoder.SetEscapeHTML`;
`internal/cli/describe.go` 104-135 y 328-343; `internal/app/esquema_test.go` 90-111 (patrón
de compilación con `AssertFormat`); `invopop/jsonschema` `reflect.go` 772-801; `santhosh-tekuri/jsonschema`
`roots.go`, `root.go`, `compiler.go` (tabla).

---

## D12 · Fixtures: dónde viven, cómo se graban, sintéticos, referencias y golden

**Decisión.**

- **Dónde** (resuelve `docs/PENDIENTES.md`, «Antes de la primera tarea `[datos]` de H4»): junto al paquete.
  - Grabaciones reales: `internal/source/boe/testdata/boe.legislacion-consolidada/` (raíz de grabación
    `internal/source/boe/testdata`, fuente `boe.legislacion-consolidada`: la ruta es literalmente
    `testdata/<fuente>/` del paquete de la fuente, como piden constitución §III y `CLAUDE.md`).
  - Sintéticos derivados a mano de una grabación real: `internal/source/boe/testdata/sintetico/<escenario>/boe.legislacion-consolidada/`.
  - Referencias del diff de aceptación: `internal/source/boe/testdata/referencias/<norma>-<bloque>.json`.
  - Golden: `internal/source/boe/testdata/golden/<caso>.json`.
- **Grabación** (FR-113; constitución: «`KITLEGAL_RECORD=1` solo en un job separado y revisado»): la hace **una
  persona**, nunca el ejecutor, en la pausa humana de la tarea `[datos]` que escribe el **manifiesto**
  `internal/source/boe/testdata/grabaciones.json` (la lista cerrada de recursos de
  [contracts/esquemas-fixtures-y-controles.md](./contracts/esquemas-fixtures-y-controles.md) §3). Esa tarea es la que
  provoca la pausa: `clasificar_datos` solo pausa cuando una tarea `[datos]` añade ficheros bajo
  `internal/source/*`, y una tarea que solo pidiera grabar no cambiaría ninguno. En la pausa, la persona revisa el
  manifiesto y, por este orden: deja la fila definitiva de `docs/SOURCES.md` con la fecha real de la revisión y
  alinea con ella `IntervaloEntrePeticiones` y `terminosDeUso` en `terminos.go` ([D15](#d15--la-fuente-docssourcesmd-ritmo-términos-robotstxt-y-verificación-nocturna));
  comprueba `TestFuenteCoincideConSources`; graba con `scripts/grabar-fixtures.sh`, que así usa el intervalo revisado;
  revisa lo grabado (S1, S2, S4, S7); escribe y revisa las cinco referencias (abajo); lo confirma todo en la rama y
  aprueba. La tarea siguiente parte de ese commit. El arnés, que una tarea de código anterior deja
  escrito, es `internal/source/boe/grabacion_test.go` con `//go:build grabacion` (`TestGrabarFixtures`): exige
  `KITLEGAL_RECORD=1` (si no, `t.Fatal`), lee el manifiesto y pide cada recurso con `httpx.New(ConFuente,
  ConRaizDeGrabacion("testdata"), ConIntervalo(boe.IntervaloEntrePeticiones))`, construyendo la dirección y el
  `Accept` con `direcciones.go` y `busqueda.go`. No usa la lógica de lectura de `boe`, que todavía no existe.
  `TestGrabacionesCompletas`, en la tarea de código que sigue a la pausa, exige una grabación por entrada del
  manifiesto, y `TestReferenciasCompletas`, las cinco referencias con la forma del contrato.
- **Sintéticos** (lista cerrada, §4 del mismo contrato): los que la fuente no da a voluntad —ilegible, sin bloque,
  503, 429, metadatos caídos, los tres avisos— copiados de una grabación real y alterados en el mínimo, en
  directorios separados para no pisar la grabación de la misma dirección.
- **Casos de lectura** que no necesitan el kernel (sin versiones, última sin fecha, `tail`, CRLF, objeto suelto,
  índice anidado o plano, envoltorios) se prueban con **tablas en `_test.go`**, sin fixture.
- **Referencias** (FR-116, Q2): cinco ficheros que **una persona** deriva a mano de las grabaciones con la lógica de
  `refs/boe.py`, con la línea que justifica cada campo, y revisa **en la pausa de la tarea `[datos]` que graba**, tras
  grabar y antes de confirmar. Es la letra de FR-116 y Q2 («revisado por una persona en la tarea `[datos]` que graba
  los fixtures») y lo único posible en esa tarea, cuya parte automática —el manifiesto— es anterior a que exista
  ninguna grabación. Quedan confirmadas **antes** de todo código de lectura y de `articulo`; el ejecutor nunca las
  crea, las completa ni las ajusta. `TestReferenciasCompletas` comprueba su forma; `TestArticuloCoincideConBoePy`, su
  contenido.
- **Golden** (FR-112) y **esquemas** (FR-110): los produce el código con una bandera de test y los revisa una
  persona. Como una tarea `[datos]` solo toca `testdata/` y `schemas/`, la secuencia es de tres pasos:
  1. tarea de código: el comparador (`TestGolden`, `TestEsquemasPublicados`) compara **cada fichero que exista**
     y la bandera de regeneración escribe los de la lista cerrada; con cero ficheros no compara nada;
  2. tarea `[datos]`: se ejecuta la regeneración y una persona revisa;
  3. tarea de código: `TestGoldenCubreTodosLosCasos` y `TestEsquemasCubrenTodosLosVerbos` exigen que existan
     **todos** (lista cerrada, seis de seis verbos) y `make schema-check` pasa a apoyarse en ellos.
- **Artículos del diff** (SC-001, S7): `BOE-A-2015-10565 a21` (LPAC; obligatorio), `BOE-A-2015-10565 a1` (LPAC;
  candidato «sin modificaciones»), `BOE-A-1985-5392 a22` (LRBRL; candidato «varias versiones»),
  `BOE-A-2017-12902 a118` (LCSP; candidato «varias versiones»), `BOE-A-2017-12902 da3` (LCSP; disposición). Si
  al grabar no hay al menos uno con una sola `version` y otro con dos o más, se sustituye un candidato por
  `BOE-A-2015-10565 a5` o `BOE-A-1985-5392 a1` (suplentes declarados) en la misma tarea `[datos]`.

**Por qué.** Junto al paquete: los tests abren `testdata/…` sin `../../..`, el go command ignora `testdata`, el
paso `clasificar_datos` del workflow pausa igual (ficheros nuevos bajo `internal/source/*`) y es la recomendación de
`docs/PENDIENTES.md`. Grabar es un acto humano por constitución (capa 3) y por el prompt del workflow («nunca
se graban fixtures dentro del workflow»).

**Alternativas rechazadas.** *`testdata/boe/` en la raíz*: tests con `../../../testdata`, y los de `internal/app`
con otra profundidad; nada gana. *Grabar desde el binario con una variable para la raíz*: una variable de
entorno nueva en el binario distribuido para algo que solo hace una persona al preparar fixtures. *Un job de
Actions `workflow_dispatch`*: exige la plataforma en mitad del hito (las tareas `[plataforma]` van al final).
*Golden escritos a mano*: el índice de la LPAC tiene cientos de bloques; lo independiente del código ya lo son
las referencias. *Generar golden y esquemas en la misma tarea que el test*: una tarea `[datos]` no toca código. *Referencias que
deriva el ejecutor en una tarea `[datos]` posterior a la grabación y revisa una persona en su pausa*: se aparta de FR-116
y Q2, que fijan la revisión en la tarea que graba, y pone la misma mano en la referencia y en el código que se compara
con ella, de modo que una misma lectura equivocada de `boe.py` (el `tail`, la exclusión de la raíz) podría aparecer en
los dos y dejar el diff vacío. *Referencias escritas antes de grabar*: no hay fixtures de los que derivarlas. *Dejar las
grabaciones sin confirmar para que una tarea `[datos]` posterior las añada con las referencias*: dependería de ficheros
sin confirmar entre tareas.

**Verificado en.** `docs/PENDIENTES.md`; `.specify/workflows/hito/workflow.yml` 882-900 (`clasificar_datos`:
pausa por ficheros nuevos bajo `internal/source/*` o `schemas/`), 902-910 (`gate_humano_datos`: aprobar o rechazar,
y rechazar aborta) y 430-438 (reglas de `tasks`; 432: el e2e en la primera tarea que pueda dejarlo en verde); `docs/WORKFLOW.md`
57 y 71; `internal/httpx/cliente.go` 111-134 (`ConRaizDeGrabacion`), `grabar.go` 332-387.

---

## D13 · Tests, e2e y verificación: reproducción en el binario de e2e, etiquetas `fuentes` y `grabacion`

**Decisión.**

- **Tests del adaptador** (`internal/source/boe`, paquete `boe_test` salvo los que usan símbolos sin exportar:
  lecturas internas y `terminos_test.go`): contra
  `httpx.Replay` (nunca `httptest`, R2) y una caché real en `t.TempDir()` con `cache.ConReloj` y
  `httpx.ConHora` del mismo reloj controlado. Las peticiones se cuentan envolviendo el cliente de reproducción
  en un `Pedidor` que cuenta; la «reproducción estricta» es `Replay` sobre un directorio vacío.
- **Tests del applet en proceso** (`internal/app`): `app.Main` sobre un registro con `AppletBoe` y dependencias de
  reproducción; validan sobres contra `schemas/`, códigos 2-5 y fechas (SC-006, SC-008, SC-013).
- **e2e** (FR-114): los guiones `internal/app/testdata/script/boe-*.txtar` sobre el binario de e2e, que registra
  `boe` con `httpx.Replay` sobre el directorio **relativo** `reproduccion/boe.legislacion-consolidada` (el
  directorio de trabajo del guion). `Setup` copia ahí las grabaciones con `os.CopyFS` y fija
  `KITLEGAL_CACHE_DIR=$WORK/cache`. Una ruta literal evita leer una variable de entorno en el `main` del e2e
  (`gosec` v2.28.0 declara `os.Getenv` fuente de propagación en `analyzers/commandinjection.go` 34, G702, y
  `analyzers/pathtraversal.go` 35, G703).
- **Consecuencia en un guion de H1**: `argumentos.txtar` exige `applets disponibles: contar, echo` (líneas 24 y 30) y el
  binario de e2e pasa a registrar `boe`. Como una tarea `[datos]` no toca código y cada tarea deja `make ci` en verde, el
  guion se relaja a `(boe, )?contar, echo` en una tarea `[datos]` anterior al registro y se ajusta a `boe, contar, echo`
  con los guiones nuevos, en la tarea `[datos]` que sigue **inmediatamente** a la de registro: es la primera que puede
  dejar en verde el e2e de la entrega (los guiones solo necesitan `boe` en el binario de e2e, `Setup` y `cronometra`),
  así que va antes de golden, esquemas y contratos, que no le hacen falta; las dos modifican material existente y pausan. Los demás guiones de H1 siguen casando:
  `ayuda.txtar` busca cada applet en su línea y el relleno de la columna no cambia (`contar` sigue siendo el nombre más
  largo), y `multicall.txtar` invoca `echo` por su enlace.
- **< 200 ms** (FR-117, SC-002): guion `boe-cache-rapida.txtar` con una orden propia `cronometra <máximo>
  <programa> <args…>` (`Params.Cmds`, `TestScript.Exec`), que falla si la invocación dura el máximo o más. El
  guion siembra la entrada, **vacía** la reproducción (`rm` + `mkdir`, reproducción estricta) y cronometra diez
  invocaciones: cualquier petición terminaría en código 1 nombrándola. El binario se construye sin `-race`
  (`go install` en `TestMain`).
- **Verificación nocturna** (FR-115): `scripts/verify-sources.sh` ejecuta `go test -tags=fuentes -count=1 -run
  '^TestVerificarFuentes$' ./internal/app/`, que invoca en proceso `boe articulo BOE-A-2015-10565 a21 --json`
  con las dependencias de red y la caché en `t.TempDir()`, y comprueba código 0, sobre válido contra la parte
  `articulo` de `schemas/bloque.json` y `texto` no vacío. La comprobación es una función (`verificarArticulo`)
  que devuelve error con «caso «boe articulo»: …»; `TestVerificacionDeFuentesDetectaCambios` (sin etiqueta, en
  `make test`) la ejercita contra la grabación buena (pasa) y contra el sintético ilegible (falla nombrando el
  caso): el control negativo de SC-009 sin red.
- **Grabación**: etiqueta `grabacion` (D12).
- **Lint**: `.golangci.yml` `run.build-tags: [integration, fuentes, grabacion]`, para que los dos ficheros
  etiquetados no queden fuera del análisis. Ni `make test` ni `make test-integration` los compilan.

**Por qué.** El e2e tiene que ejercer el binario compilado sin red (FR-114); cambiar el binario distribuido para
que acepte una reproducción por variable permitiría servir texto «del BOE» desde un directorio cualquiera con
`fuente: boe.legislacion-consolidada`, que es falsificar una cita. El binario de e2e ya es la raíz de
composición de pruebas (ADR 0010).

**Alternativas rechazadas.** *Variable `KITLEGAL_REPLAY` en el binario distribuido*: lo anterior. *Segundo
binario de e2e*: otra excepción de lint y otra construcción por ejecución. *Cronometrar desde Go con
`exec.Command`*: `gosec` G204 sobre la ruta variable del binario y un `//nolint` que el plan no admite.
*`verify-sources.sh` que valide con un CLI de esquemas*: dependencia nueva fuera de §V.

**Verificado en.** `testscript.go` 145-159, 495, 1211; `internal/app/e2e_test.go` 59-171; `go doc os.CopyFS`;
`.golangci.yml` (`run.build-tags`, H3 D11).

---

## D14 · Controles de arquitectura y superficie del binario

**Decisión.**

- Se **retiran** `TestElBinarioNoEnlazaHTTPX` y `TestElBinarioNoEnlazaCache` (FR-124, `docs/PENDIENTES.md`).
- `modulosDelBinario` se amplía con los módulos que `go list -deps` muestre **al enlazar** `boe`, justificados
  uno a uno en el comentario de la lista y en la propuesta de cambio: `golang.org/x/time` y
  `github.com/temoto/robotstxt` (§V, por `internal/httpx`); `modernc.org/sqlite` (§V, por `internal/cache`);
  y los que arrastra el controlador, medidos hoy como `modernc.org/libc`, `modernc.org/mathutil`,
  `modernc.org/memory`, `golang.org/x/sys`, `github.com/dustin/go-humanize`, `github.com/google/uuid`,
  `github.com/mattn/go-isatty`, `github.com/ncruces/go-strftime` y `github.com/remyoudompheng/bigfft` (S10).
- **Espacio reservado** (FR-002, SC-011, ADR 0006 «prohibición que nace en H4»): `TestLasFuentesNoFirmanComoKitlegal`
  en `internal/arch_test.go` lista con `go list -f '{{.Dir}}' ./internal/source/...` los paquetes de fuentes
  (exige al menos uno: no pasa en vacío), analiza con `go/parser` sus ficheros `.go` de producción y falla ante
  cualquier literal de cadena que empiece por `kitlegal.` o `kitlegal:`, nombrando fichero y línea. Refuerzo en
  ejecución: los tests de `boe` comprueban `fuente` y `url` `https://www.boe.es/…` de los seis verbos.
- R1-R5 **sin cambios de configuración**: `internal/source/boe` no importa `net/http` (depguard `red`, también
  en tests), ni SQLite (lista `sql`), ni llama a `os.Exit` ni nombra `os.Stdout`/`os.Stderr` (forbidigo). R2
  conserva `duenoObligatorio`; R3 sigue sin él (faltan `store` y `graph`).

**Por qué.** Las dos pruebas retiradas eran ciertas solo mientras ningún applet usara red y caché. El control
del espacio reservado no puede ser de `forbidigo` (vigila identificadores, no literales) ni de depguard; un
recorrido del AST de los paquetes de fuentes es mecánico y nombra dónde.

**Alternativas rechazadas.** *Copiar la lista de módulos de `go.mod`*: declararía módulos que el binario no
enlaza (H3 sonda 6). *Comprobar el espacio reservado solo en ejecución*: un adaptador sin tests lo esquivaría.

**Verificado en.** `internal/arch_test.go` 105-228; medida `go list -deps` de la tabla; `.golangci.yml` 169-200.

---

## D15 · La fuente: `docs/SOURCES.md`, ritmo, términos, `robots.txt` y verificación nocturna

**Decisión.**

- `docs/SOURCES.md` nace con una tabla y una fila `boe.legislacion-consolidada`: applet `boe`, URL base de la
  API, licencia de reutilización, términos de uso (URL entre `<…>`), `robots.txt` (resultado de la revisión),
  **ritmo** (literal de duración de Go entre comillas invertidas), formato (XML/JSON, sin autenticación) y
  **revisado** (`AAAA-MM-DD`). La tarea de código del arnés de grabación la escribe como **propuesta**: ritmo `1s` (el
  mismo que la omisión conservadora de `httpx`, S6), la URL de términos que propone, licencia y `robots.txt` por revisar
  y «Revisado» `pendiente`, porque una fecha escrita por el ejecutor afirmaría una revisión que no ha ocurrido.
- **Procedimiento en la pausa** de la tarea `[datos]` del manifiesto, **antes de grabar**, a cargo de una persona:
  1. revisa los términos de uso y el `robots.txt`; si prohíben el acceso automatizado, rechaza la pausa y el hito se
     detiene (FR-123, criterio 4);
  2. deja la fila definitiva, con la fecha real de la revisión en «Revisado»;
  3. alinea con ella `IntervaloEntrePeticiones` y `terminosDeUso` en `internal/source/boe/terminos.go`;
  4. comprueba `go test -count=1 -run '^TestFuenteCoincideConSources$' ./internal/source/boe/`;
  5. graba (D12), de modo que `TestGrabarFixtures`, que usa `boe.IntervaloEntrePeticiones`, pide con el ritmo revisado;
  6. lo confirma en la rama con las grabaciones.

  Que ninguna tarea posterior cambie la fila ni `terminos.go` depende de una regla sobre **el texto entero de las líneas
  de tarea**, no de que ninguna tarea pretenda declarar esos ficheros: el extractor de rutas toma como declarada toda
  ruta completa que aparezca en cualquier parte de la línea (`workflow.yml` 583 y 606), también la de un fichero que la
  tarea solo lee, compara o dice no tocar y la de un paquete dentro de una orden; el guardián admite esa ruta y todo lo
  que cuelga de ella (695-698), y el extractor deja en su directorio una ruta con llaves o con comodín (606 y 696). Por
  eso, desde la tarea que sigue a la pausa, ninguna línea contiene en ninguna parte —prosa u orden— la ruta completa de
  `docs/SOURCES.md` o `internal/source/boe/terminos.go`, ni de un directorio que los contenga (`docs/`, `internal/`,
  `internal/source/`, `internal/source/boe/`): se nombran sin directorio (`SOURCES.md`, `terminos.go`) o por el test que
  los lee (`TestFuenteCoincideConSources`), que el guardián no casa con la ruta completa (698 compara la cadena entera),
  y los ficheros del paquete que la tarea cambia van por su ruta completa, sin llaves ni comodines (plan, obligación 3).
  Con esa regla el guardián rechaza cualquier cambio en ellos, lleve o no la tarea `[datos]` (705); lo mismo vale para el
  manifiesto, las grabaciones y las referencias que la persona confirma en la misma pausa (D12), cuyas rutas, las de lo
  que cuelga de ellas y las de sus directorios tampoco aparecen en ninguna línea posterior. La línea de la propia tarea
  del manifiesto lleva, de todo ello, solo la ruta del manifiesto. Sin la regla, `docs/SOURCES.md` y `terminos.go`, fuera
  de `testdata/`, no los frenaría la etiqueta `[datos]` (704) ni los pausaría `clasificar_datos` (882-900): solo la
  pausa final del workflow (`rutas_sensibles`, 1134), que salta porque el diff del hito toca `docs/SOURCES.md` y lo
  muestra entero como fichero nuevo, sin distinguir lo que fijó la persona de un cambio posterior.
- `boe.IntervaloEntrePeticiones` y `terminosDeUso` (URL y día de revisión, que devuelve `Fuente.Terms()`) viven solos en
  `terminos.go`, para que la persona los fije sin tocar otro código y ninguna tarea posterior tenga que cambiar ese
  fichero ni llevar su ruta en la línea. `TestFuenteCoincideConSources` (`terminos_test.go`, interno) lee `docs/SOURCES.md`, localiza la fila por su
  primera celda y exige igualdad del ritmo y de la URL de términos, y que «Revisado» sea una fecha igual a
  `terminosDeUso.Revisados` (FR-122: «tomarse de su fila»; FR-121). Hasta la pausa admite **exactamente** la pareja
  «Revisado» `pendiente` con `Revisados` cero; la tarea de código que sigue a la pausa retira esa admisión, de modo que
  desde entonces una fila sin revisar deja `make ci` en rojo.
- **Nocturno** (FR-115): trabajo nuevo `fuentes` en `.github/workflows/nightly.yml`, separado del de `make ci`,
  con `permissions: contents: read, issues: write`, que ejecuta `make verify-sources` y, **si falla**, busca una
  incidencia abierta con el título «verify-sources: boe articulo» (`gh issue list --state open --search`) y
  comenta en ella, o la crea (`gh issue create --title --body`). `make verify-sources` es un objetivo nuevo del
  `Makefile` que llama a `scripts/verify-sources.sh` con el `GOTOOLCHAIN` exportado; **no** entra en `make ci`.

**Por qué.** El `Makefile` es la única superficie de invocación (`CLAUDE.md`): el nocturno llama a `make`, como el
trabajo de CI. Un test que ata constantes a la tabla es la única forma mecánica de que el ritmo «salga de» un
documento que el binario no lleva dentro.

**Alternativas rechazadas.** *Leer `docs/SOURCES.md` en tiempo de ejecución*: el binario no lo distribuye.
*Dejar el ritmo por omisión de `httpx`*: `docs/PENDIENTES.md` («En H4») lo prohíbe. *Fecha «Revisado» propuesta por el
ejecutor*: afirma una revisión que no ha ocurrido. *Que una tarea de código posterior a la pausa alinee las constantes
con la fila*: la grabación, anterior, pediría con un intervalo sin revisar, y el ejecutor tocaría lo que fija una persona.
*Atar la fila a las constantes solo tras la pausa*: la persona no podría comprobar antes de grabar que casan. *Constantes
en `fuente.go`*: las tareas que completan la fuente tendrían que declarar el fichero que fija la persona y el guardián
dejaría de protegerlo. *Etiquetas de incidencia*: FR-115 no las pide y el título basta para encontrar la incidencia con
`gh issue list --search`; una etiqueta sería un objeto más que mantener en el repositorio remoto. Qué hace `gh issue
create --label` con una etiqueta inexistente no se ha comprobado (S12) y no sostiene esta decisión.

**Verificado en.** `.specify/workflows/hito/workflow.yml` 583 (`texto` es la línea entera de la tarea sin la casilla) y
606 (extractor de rutas sobre ese texto entero). Comprobado en local con la misma tubería (`tr` de comillas invertidas y
`,;()` a espacios, `grep -oE` con la expresión de 606, `grep -vE '^https?:'`, `sort -u`) sobre líneas de ejemplo: de
«… en internal/source/boe/terminos_test.go (lee docs/SOURCES.md; no toca internal/source/boe/terminos.go) y
TestGrabacionesCompletas sobre internal/source/boe/testdata/grabaciones.json y internal/source/boe/testdata/referencias/»
saca las cinco rutas, incluidas las cuatro que la tarea solo lee o dice no tocar; de la orden
`go test -count=1 -run '^TestGolden$' ./internal/source/boe/ -args -actualizar-golden` saca `internal/source/boe/`, y
de la de esquemas con `./internal/app/`, `internal/app/`; `internal/source/boe/testdata/boe.legislacion-consolidada/`
sale como `internal/source/boe/testdata/boe.legislacion-consolidada`; con nombres sin directorio salen `SOURCES.md`,
`terminos.go`, `grabaciones.json` y `referencias/`, y una ruta parcial sale tal cual (`boe/terminos.go`); de
`internal/httpx/{peticion,errores}.go` saca `internal/httpx/`. 695-698 (una ruta declarada admite ella misma y todo lo
que cuelga de ella; 696 recorta el comodín, de modo que `internal/source/boe/*_test.go` queda en `internal/source/boe`)
y 698-699 (comparan la cadena entera; comprobado en local con los mismos `case`: `terminos.go` y `boe/terminos.go` no
casan con `internal/source/boe/terminos.go` ni, por la regla de `x_test.go`, con `internal/source/boe/terminos_test.go`,
e `internal/source/boe` sí casa); 704-705 (con `[datos]` se admite tocar `testdata/`; lo que queda fuera de las rutas
se rechaza con o sin `[datos]`); 882-900 (`clasificar_datos` solo mira `testdata/` y `schemas/`); 1134
(`rutas_sensibles` pausa si el diff del hito toca `docs/SOURCES.md`); `docs/WORKFLOW.md` 71 («Declarar un directorio
(`internal/cli/`) permite todo lo que cuelga de él») y 201 («El guardián de diff extrae rutas de la línea de la tarea»); `gh issue create --help` (`--title`, `--body`; `--label` sin más que «Add labels by name»),
`gh issue list --help`, `gh issue comment --help`; `gh help formatting`
(`--jq` admite la sintaxis de jq sin `jq` instalado, de ahí `.[0].number // empty`);
`.github/workflows/nightly.yml` y `ci.yml`; `Makefile` 14-20 (`GOTOOLCHAIN`).

---

## D16 · Registro de producción sin pánico: `app.Arrancar`

**Decisión.** `RegistroDeProduccion() (*Registro, error)` registra `boe` con `DependenciasDeRed()`.
`app.Arrancar(argv, construir func() (*Registro, error), stdout, stderr, version, commit, fecha) int` construye
el registro y, si falla, emite el fallo por el mismo `cli.Montador` (sobre `kitlegal.cli` de clase
«inesperado» si se pidió `--json`, mensaje en la salida de error, código 1); si no, delega en `Main`.
`cmd/kitlegal/main.go` y el `main` del binario de e2e llaman a `Arrancar` (este último deja de usar `panic`).

**Por qué.** Registrar el primer applet hace falible el registro de producción; un registro inválido es un
defecto de programación que `TestRegistroDeProduccion` impide, pero la constitución §IV no admite `panic` y la
traducción de errores a código vive en un único punto.

**Alternativas rechazadas.** *`panic` como el e2e de H1*: evitable sin coste. *Cambiar la firma de `Main`*:
toca todos sus tests y el adaptador de prueba de H3. *Ignorar el error*: silenciarlo.

**Verificado en.** `internal/app/main.go` 78-105; `cmd/kitlegal/main.go`; `ejemplo/kitlegal-e2e/main.go` 51-58.

---

## D17 · Documentación, ADR y registro de cambios

**Decisión.** ADR `docs/ADR/0015-puerto-source-y-fecha-de-consulta.md` (D2, D3). `CHANGELOG.md` › *Unreleased* ›
*Añadido*: applet `boe`, `schemas/`, `make schema-check` activo, `make verify-sources`; › *Cambiado*: `fecha_consulta`
de lo servido desde caché, `Arrancar`. `README.md` y `CONTRIBUTING.md`: fila de `make schema-check` (ya no «los
aportan H4 y H10», ahora activo) y `make verify-sources`. `docs/PENDIENTES.md`: se borran las dos entradas de H4
resueltas. `docs/SOURCES.md`: D15.

**Por qué.** Definition of Done de `docs/ROADMAP.md` §1, puntos 6 (comportamiento visible → `CHANGELOG.md`), 7
(decisión de arquitectura → ADR) y 8 (fuente externa → fila de `docs/SOURCES.md` y caso en `scripts/verify-sources.sh`).
Y el hito cambia el `Makefile` en lo que la documentación del repositorio describe hoy: `README.md` 130 dice de
`make schema-check` que los esquemas «los aportan H4 y H10», `CONTRIBUTING.md` 161 que «Anuncia que no hay `schemas/`
todavía y termina con éxito», y ninguno de los dos lista `make verify-sources`; si no se alinean en la misma rama, la
documentación dice lo contrario que el `Makefile`.

---

## D18 · Nombres que `misspell` y `gosec` aceptan

**Decisión.**

- **Qué revisa `misspell`** en el `golangci-lint` del repositorio (v2.13.2 con `misspell` v0.8.0,
  `tools/golangci-lint/go.mod` 91 y 93), comprobado en su código:
  - Solo ficheros Go (`pkg/golinters/misspell/misspell.go` 31-32 y 77-80), pero **enteros**: `.golangci.yml` no fija
    `mode` (203-235), así que se aplica `Replace` a todo el contenido —comentarios, cadenas, etiquetas de campo e
    identificadores— (`misspell.go` 89-102); solo `mode: restricted` aplicaría `ReplaceGo`, que se limita a los
    comentarios (`replace.go` 85-112). Guiones `.txtar`, JSON, YAML y Markdown no pasan por él.
  - Antes de buscar blanquea URL, correos, nombres de sitio, rutas y escapes con barra invertida (`notwords.go` 103-106,
    llamado en `replace.go` 220). Un nombre de sitio es `([[:alnum:]-]+\.)+[[:alpha:]]{2,63}` y solo se blanquea si no
    lleva mayúsculas (`notwords.go` 17 y 76-84): `boe.legislacion` en minúsculas no se revisa.
  - Una palabra es `[a-zA-Z0-9']+` (`replace.go` 19 y 222): la corta cualquier otro carácter —espacio, comillas dobles o
    invertidas, `/`, `-`, `_`, `.`, una letra con tilde—, pero **no** el apóstrofo, que forma parte de la palabra. Así,
    `fecha_disposicion` y `disposicion-adicional` contienen la palabra `disposicion`, e `información` no contiene
    `informacion` (la `ó` la corta en `informaci`).
  - Solo informa si la palabra **entera** es la primera columna de una pareja: la sustitución tiene que ser la de
    `corrected[strings.ToLower(word)]` (`replace.go` 226-238; si no, 254 la ignora). Compara sin distinguir mayúsculas
    (`stringreplacer.go` 186, 210 y 234): marca la palabra en minúsculas, en mayúsculas o con solo la inicial mayúscula
    (`case.go` 19-44) y descarta la de mayúsculas mezcladas (`replace.go` 232-236, `CaseUnknown`), de modo que marca
    `Dependencias` al principio de un comentario y no `DependenciasDeBoe`.
  - `ignore-rules` retira la pareja sin distinguir mayúsculas (`misspell.go` 66-68; `replace.go` 49-64).
- **Lista cerrada**, comprobada contra `words.go` y `words_us.go` con el barrido de la tabla del principio:
  - **A `misspell.ignore-rules`**, cada una con su comentario como las de H1-H3, en la tarea que las introduce, que
    **declara `.golangci.yml`**:

    | Palabra | `words.go` | Por qué no se puede evitar | Tarea |
    |---|---|---|---|
    | `administrativo` | 527 | «procedimiento administrativo común», consulta de la lista cerrada (manifiesto, golden `buscar-procedimiento-administrativo-comun`, `TestConsultaDeBusqueda`) | paso 3 |
    | `capitulo` | 21070 | valor que devuelve `TipoDesdeID` y enumerado de `EntradaDeIndice.tipo` | paso 3 |
    | `disposicion` | 6998 | `disposicion_adicional`… de `TipoDesdeID`; clave `fecha_disposicion` | paso 3 |
    | `materias` | 22765 | clave de `data` de `analisis`; verbo `materias` de la entrada 32 de FR-120 en `doc.go` | paso 3 |
    | `regulares` | 19592 | «expresiones regulares», entrada 7 de FR-120 en `doc.go` | paso 3 |
    | `dependencias` | 3330 | contrato `DependenciasDeBoe`/`DependenciasDeRed`, que comentarios y mensajes de `Nueva` ante dependencias ausentes nombran | paso 9 (`fuente.go`) |

  - **Ya cubiertas** por `ignore-rules`: `argumentos` (clase «argumentos»), `legislacion` (`legislacion-consolidada` sin
    el `boe.` delante, que en minúsculas se blanquea como nombre de sitio),
    `reproduccion` (directorio de reproducción del binario de e2e), `versiones`, `controles`, `descripcion`,
    `inventario`.
  - **Nunca sueltas sin tilde** (se escriben con ella) las 31 del grupo (3) de la tabla del principio: `adaptacion`,
    `compilacion`, `composicion`, `configuracion`, `conjuncion`, `constitucion`, `construccion`, `convencion`,
    `correccion`, `corrupcion`, `declaracion`, `distribucion`, `documentacion`, `emision`, `evaluacion`, `historicas`,
    `identificacion`, `implementacion`, `informacion`, `justificacion`, `omision`, `orientacion`, `parametros`,
    `presentacion`, `produccion`, `prohibicion`, `proposito`, `proteccion`, `regeneracion`, `transcripcion` y
    `verificacion`. En comentarios, cadenas y mensajes se escriben con tilde, que corta la palabra antes de que case.
    Donde no se escriben tildes —nombres de subtest y de caso, claves, nombres de fichero, `snake_case` o `kebab-case`—
    se reformulan con el verbo, que el diccionario no marca (`adaptar`, `compilar`, `componer`, `configurar`,
    `construir`, `corregir`, `declarar`, `distribuir`, `documentar`, `emitir`, `evaluar`, `identificar`, `implementar`,
    `informar`, `justificar`, `omitir`, `orientar`, `presentar`, `producir`, `prohibir`, `proteger`, `regenerar`,
    `transcribir`, `verificar`; búsqueda directa en `words.go` y `words_us.go`), o con otra palabra comprobada igual:
    «defecto de composición» → `defecto-al-componer`, «bandera de regeneración» → `bandera-para-regenerar`, «valor por
    omisión» → `valor-si-se-omite`, «la identificación de `httpx`» → `httpx-se-identifica`.
  - **Palabras correctas que el diccionario marca**, que se sustituyen: `decisiones` → «decisión» o «lo decidido»;
    `directos` → «de primer nivel» («hijos de primer nivel»); `directorios` → «carpeta» o el nombre concreto;
    `recorre` → «visita» o «recorrido»; `clientes` → singular; `comando`, `comandos` → «orden», «órdenes»; `defectos`
    → singular; `posicional` → «de posición»; `producto` → no se usa; `momento` → «instante»; `contradice` → «choca
    con»; `distribuye` → «publica»; `resolverse` → «resolver»; `calcular` → «obtener»; `inaccesibles` → «no
    accesibles».
  - Una palabra marcada que no esté en la lista se reescribe; si es contenido fijado por el contrato, la tarea se
    detiene y se redelimita declarando `.golangci.yml`.
- `gosec`: lecturas de `testdata` por `filepath.Clean`; escrituras de golden y esquemas con `0o600`, en un
  auxiliar distinto del que lee (G703); ningún `exec` nuevo fuera de `testscript` y del `ejecutaGo` existente.

**Verificado en.** `tools/golangci-lint/go.mod` 91 y 93 (versiones); `golangci-lint` v2.13.2
`pkg/golinters/misspell/misspell.go` 31-32 (recorre `pass.Files`), 50-54 (`locale: US` añade `DictAmerican`), 66-68
(`ignore-rules` → `RemoveRule`), 77-80 (solo ficheros Go) y 89-102 (`restricted` → `ReplaceGo`; sin `mode` → `Replace`
sobre el fichero entero); `misspell` v0.8.0 `replace.go` 19 (clase de palabra), 49-64 (`RemoveRule`), 85-112
(`ReplaceGo`: solo comentarios), 143-168 (`Replace`) y 218-258 (`recheckLine`: blanqueo, palabra entera, mayúsculas
mezcladas descartadas), `notwords.go` 17, 76-84 y 103-106, `case.go` 19-44, `stringreplacer.go` 186, 210 y 234,
`words.go` y `words_us.go` (barrido y líneas de la tabla del principio); `.golangci.yml` 203-235 (`locale: US`, sin
`mode`, `ignore-rules` actuales); `gosec` v2.28.0 `analyzers/pathtraversal.go` 35 (G703), `rules/rulelist.go` 92 y 94,
`rules/readfile.go` 157 y 240 (G304: `filepath.Clean` como saneado) y `rules/fileperms.go` 80-81 (G306: `0o600` por
omisión).

---

## D19 · Aplicación de las skills `golang-*` instaladas

`golang-how-to` enruta a `golang-project-layout`, `golang-cli`, `golang-testing` (y a `golang-error-handling`,
`golang-design-patterns`, `golang-context` por sus tablas). Aplicado:

- **Layout** (`golang-project-layout`): lógica en `internal/`, `main` mínimos, `_test.go` junto al código y
  `testdata/` para fixtures junto al paquete (D12); sin `pkg/`.
- **CLI** (`golang-cli`): stdout solo para el resultado y stderr para diagnóstico (kernel), códigos de salida por
  una única traducción (D10, D16), salida legible por máquina con `--json` y validación de argumentos antes de
  actuar (D9). No se adopta Cobra/Viper (decisión cerrada: Kong).
- **Testing** (`golang-testing`): tablas con subtests nombrados, `t.Parallel()` salvo `t.Setenv`, un fichero de
  test por fichero de código y en su orden, fuzz con `f.Add` de semillas, etiquetas de compilación para lo que
  toca la red, golden con bandera de regeneración y comparación byte a byte, `-race` en CI.
- **Errores y diseño**: errores tipados con clase y `%w` (`golang-error-handling`); opciones funcionales en
  `boe.Nueva` y `httpx.ConHora`; interfaces pequeñas definidas por quien las consume (`Pedidor`, `CacheAbierta`);
  valor de ámbito de petición en el contexto con clave sin exportar (`golang-context`, D4).

**Verificado en.** `.agents/skills/golang-how-to/SKILL.md`, `.agents/skills/golang-project-layout/SKILL.md`,
`.agents/skills/golang-cli/SKILL.md` y `.agents/skills/golang-testing/SKILL.md`, leídos como ficheros.

---

## D20 · Supuestos no verificados

| # | Supuesto | Por qué no se verifica ahora | Dónde se comprueba |
|---|---|---|---|
| S1 | La API elige el formato por `Accept`: XML para `…/texto/bloque/<id>` y JSON para búsqueda, índice, metadatos y análisis | Requiere red | `Content-Type` de cada grabación en la tarea `[datos]`; los tests de lectura fallan si no |
| S2 | Norma o bloque inexistente → **HTTP 404** | Requiere red | Grabaciones de `BOE-A-2099-99999` y `a9999`; `TestArticulo/bloque-inexistente` y `TestMetadatos/inexistente` fallan si no. Si la fuente respondiera 200 sin bloque, FR-014 (4) y US1-5 (3) chocarían: se escala |
| S3 | Forma JSON `{status, data}` y formas de `materias`, `notas` y `referencias` (directas o con envoltorio) | Requiere red | Grabaciones; data-model §3 cubre las dos formas de cada una |
| S4 | XML con raíz distinta de `bloque`, sin espacio de nombres, en UTF-8 | Requiere red | Grabaciones; `TestArticuloCoincideConBoePy` |
| S5 | (a) `str.split()`/`str.strip()` de Python tratan U+001C–U+001F como espacio además de `White_Space`; (b) `isdigit()`/`\d` de Python y `unicode.IsDigit` difieren solo en dígitos no decimales (fuera del alfabeto de ids); (c) `XMLParser.feed(str)` de `_elementtree` ignora la codificación declarada; (d) expat normaliza el valor de los atributos (XML 1.0 §3.3.3); (e) el acelerador C `_elementtree`, que `ET.fromstring` usa cuando existe (`ElementTree.py` 2085-2086), trata comentarios e instrucciones de proceso como el `TreeBuilder` de Python puro: no los inserta y deja unido el texto de sus dos lados (1413-1415, 1486-1511); (f) expat rechaza un documento con más de un elemento raíz o con texto que no sea espacio en blanco fuera de él (XML 1.0 §2.1, `document ::= prolog element Misc*`) | Exigen ejecutar Python o leer la fuente C, que no está en local | Documentado en `doc.go`; los tests fijan el comportamiento portado; si una grabación trae comentarios, las referencias que escribe una persona (FR-116) lo contrastan |
| S12 | `gh issue create --label` con una etiqueta que no existe en el repositorio falla | La ayuda local no lo dice y comprobarlo exige crear una incidencia en el repositorio remoto | No se comprueba: el trabajo nocturno no usa etiquetas y ninguna decisión se apoya en ello (D15) |
| S6 | Los términos de uso y el `robots.txt` del BOE admiten el acceso automatizado a la API; un ritmo de `1s` es aceptable; URL de términos | Requiere red y es responsabilidad legal (capa 3) | Revisión humana en la pausa previa a grabar; `docs/SOURCES.md`; si lo prohíben, parada (FR-123) |
| S7 | Entre los cinco artículos hay uno con varias versiones y otro sin modificaciones | Requiere red | Tarea `[datos]` de grabación, con suplentes (D12) |
| S8 | `permissions: issues: write` y `GH_TOKEN: ${{ github.token }}` bastan para `gh issue list/comment/create` en Actions | Plataforma remota | Primera ejecución nocturna tras fusionar |
| S9 | Diez invocaciones con caché caliente bajan de 200 ms en el ejecutor de CI con la suite en paralelo | Depende de la máquina | `boe-cache-rapida.txtar` en `make ci` de la PR |
| S10 | Los módulos enlazados tras H4 son los medidos hoy para `internal/cache` + `internal/httpx` | El enlace no existe todavía | `TestDependenciasDelBinario` al implementar |
| S11 | `boe-fiscal` ejecutaba `boe.py` con CPython 3 con la semántica de las fuentes 3.9 locales | La versión original no consta | Referencias revisadas por una persona (FR-116) |
