# Contrato · el applet `mcp` y su servidor

Lo que ve quien declara `kitlegal mcp serve` en su agente, y lo que ve el agente por el protocolo. Requisitos: FR-001 a
FR-026 y FR-070 a FR-076. Las referencias V y D son de [research.md](../research.md).

## 1. La orden

`kitlegal mcp serve [banderas globales]`. El applet `mcp` tiene un solo verbo, `serve`, sin argumentos propios ni verbo
por omisión.

| Invocación | Qué hace | Salida estándar | Salida de error | Código |
|---|---|---|---|---|
| `mcp serve` | Atiende el protocolo MCP por la entrada y la salida estándar hasta que la entrada se cierra | solo mensajes del protocolo, uno por línea | el aviso de versión distinta, una vez, si lo hay; el registro de eventos según su nivel; el mensaje de cada llamada que falla | 0 al cerrarse la entrada (FR-024) |
| `mcp serve --offline`, `--no-graph`, `--timeout <d>` | Lo mismo; la bandera vale para todas las llamadas (§3) | | | 0 |
| `mcp serve --verbose` | Lo mismo, con el registro de eventos en detalle (§4) | solo protocolo | más líneas | 0 |
| `mcp serve --json` | Lo mismo: no cambia nada del protocolo | | | 0 |
| `mcp serve --asunto <valor>` | No atiende ningún mensaje | vacía (con `--json`, el sobre de fallo del kernel) | el evento de la invocación, como en cualquier orden que falla, y `argumentos inválidos: mcp serve no admite --asunto: el servidor no expone nada del asunto` | 2 |
| `mcp serve --dry-run` | No atiende ningún mensaje ni espera a que se cierre la entrada | vacía | `--dry-run: no se ha ejecutado nada; se habría ejecutado el applet "mcp", el verbo "serve", con los argumentos ["serve" "--dry-run"]` | 0 |
| `mcp serve --describe` | El esquema del verbo, como en cualquier otro | el documento | — | 0 |
| `mcp`, `mcp --help`, `mcp serve --help` | La ayuda, como en cualquier applet | la ayuda | — | 0 (2 sin verbo y sin `--help`) |

Un error que no es el cierre de la entrada —una línea que no es un mensaje del protocolo, una salida estándar rota— termina
el servidor como cualquier fallo del kernel: clase `inesperado`, código 1 (regla genérica; V13).

## 2. Las herramientas

Una por cada verbo de cada applet del registro, salvo los de `skills` y `mcp` (FR-002, FR-004). Con el registro de
producción, diez: `boe_analisis`, `boe_articulo`, `boe_articulos`, `boe_buscar`, `boe_indice`, `boe_metadatos`,
`graph_check`, `graph_show`, `graph_stats` y `territorio_resolver`. El orden en `tools/list` es el del SDK (por nombre).

Cada una (FR-003, FR-005):

| Campo de `tools/list` | Valor |
|---|---|
| `name` | `<applet>_<verbo>` |
| `description` | la descripción del verbo |
| `inputSchema` | la `entrada` de `--describe` del verbo, sin las ocho banderas globales, con los `$defs` que referencie |
| `outputSchema` | la `salida` de `--describe` del verbo, con los `$defs` que referencie |
| `annotations` | `{"idempotentHint":false,"readOnlyHint":true}` (lo que el SDK escribe con `ReadOnlyHint: true`; V15) |

Ejemplo, `boe_articulo` (`inputSchema`, 144 B; `outputSchema`, 1 708 B; V25):

```json
{"properties":{"norma":{"type":"string"},"bloque":{"type":"string"}},"additionalProperties":false,"type":"object","required":["norma","bloque"]}
```

```json
{"$defs":{"DatosError":{…},"boe.Articulo":{…},"boe.Aviso":{…}},"if":{"properties":{"ok":{"const":true}}},"then":{"properties":{"data":{"$ref":"#/$defs/boe.Articulo"}}},"else":{"properties":{"data":{"$ref":"#/$defs/DatosError"}}},"properties":{"ok":{"type":"boolean"},"fuente":{"type":"string","minLength":1},"url":{"type":"string","minLength":1,"format":"uri"},"fecha_consulta":{"type":"string","format":"date-time"},"hash":{"type":"string","pattern":"^sha256:[0-9a-f]{64}$"},"data":true},"additionalProperties":false,"type":"object","required":["ok","fuente","url","fecha_consulta","hash","data"]}
```

`$defs` va delante: es el orden en que el generador de `--describe` escribe las claves.

Lo que el esquema de entrada de cada herramienta de hoy declara (V25):

| Herramienta | Propiedades | Obligatorias |
|---|---|---|
| `boe_buscar` | `texto`: lista de cadenas | `texto` |
| `boe_indice`, `boe_metadatos`, `boe_analisis` | `norma`: cadena | `norma` |
| `boe_articulo` | `norma`, `bloque`: cadenas | las dos |
| `boe_articulos` | `norma`: cadena; `bloques`: lista de cadenas | las dos |
| `territorio_resolver` | `consulta`: cadena | `consulta` |
| `graph_show` | `id`: cadena | `id` |
| `graph_check` | `norma`: cadena; `bloques`: lista de cadenas | ninguna |
| `graph_stats` | ninguna | — |

Todos llevan `additionalProperties: false`: una propiedad de más, también el nombre de una bandera global, es un argumento
inválido (FR-020).

## 3. Una llamada

`tools/call` con `name` y `arguments`. El resultado (FR-010 a FR-013):

- `content`: un bloque `{"type":"text","text":"<sobre>"}`, donde `<sobre>` son los bytes que la orden del verbo escribe en
  la salida estándar con `--json` para esa entrada y ese estado, sin el salto de línea final;
- `structuredContent`: ese mismo documento JSON;
- `isError`: ausente si el sobre lleva `ok` verdadero; `true` si lleva `ok` falso.

El sobre es el de la orden salvo `fecha_consulta`, que es la de cada invocación cuando la pone el reloj.

**Ejemplo que termina bien**, `territorio_resolver` con `{"consulta":"Leganés"}` (sobre de 1 356 B; el resultado, en el
cable, unos 3 KB):

```json
{"content":[{"type":"text","text":"{\"ok\":true,\"fuente\":\"kitlegal.territorio\",\"url\":\"kitlegal:applet/territorio\",\"fecha_consulta\":\"2026-02-04T00:00:00Z\",\"hash\":\"sha256:3f71…66a2\",\"data\":{\"municipio\":{\"nombre\":\"Leganés\",…}}}"}],"structuredContent":{"ok":true,"fuente":"kitlegal.territorio","url":"kitlegal:applet/territorio","fecha_consulta":"2026-02-04T00:00:00Z","hash":"sha256:3f71…66a2","data":{"municipio":{"nombre":"Leganés",…}}}}
```

**Ejemplo que falla**, `boe_articulo` con `{"norma":"BOE-A-2015-10565"}` (277 B de sobre):

```json
{"content":[{"type":"text","text":"{\"ok\":false,\"fuente\":\"kitlegal.cli\",\"url\":\"kitlegal:cli\",\"fecha_consulta\":\"…\",\"hash\":\"sha256:…\",\"data\":{\"clase\":\"argumentos\",\"mensaje\":\"argumentos inválidos: expected \\\"<bloque>\\\"\"}}"}],"structuredContent":{"ok":false,"fuente":"kitlegal.cli","url":"kitlegal:cli","fecha_consulta":"…","hash":"sha256:…","data":{"clase":"argumentos","mensaje":"argumentos inválidos: expected \"<bloque>\""}},"isError":true}
```

El mensaje de un argumento obligatorio que falta es el de la orden (lo da el analizador). Los que comprueba la
conversión, antes de analizar (D5):

| Caso | Mensaje (`data.mensaje`) | Clase |
|---|---|---|
| los argumentos no son un objeto | `argumentos inválidos: los argumentos de <herramienta> no son un objeto JSON` | `argumentos` |
| sobra una propiedad | `argumentos inválidos: <herramienta> no tiene el argumento "<nombre>"` | `argumentos` |
| un valor no es del tipo declarado | `argumentos inválidos: el argumento "<nombre>" de <herramienta> tiene que ser <una cadena \| una lista de cadenas \| un booleano \| un número entero \| …>` | `argumentos` |
| un posicional sin el que le precede | `argumentos inválidos: el argumento "<nombre>" de <herramienta> no se puede dar sin "<el anterior>"` | `argumentos` |

Qué hace cada llamada, en orden (D3):

1. Convierte `arguments` en la línea de órdenes del verbo (`cli.LineaDeLlamada`): si no vale, el sobre de fallo de la
   clase `argumentos`, sin ejecutar nada.
2. Ejecuta el verbo como el kernel ejecuta la orden `kitlegal <applet> <verbo> --json --timeout=<plazo> [--offline]
   [--no-graph] -- <posicionales…>`: mismo análisis, mismo plazo por llamada, misma caché y, si llega a la red, un
   cliente de `internal/httpx` propio de la llamada, como el de la orden: pide su `robots.txt` y espera turno en el
   ritmo por sitio, que es de todo el servidor (D8).
3. Monta el sobre —de éxito o de fallo— con `cli.Montador`, en un búfer.
4. Si terminó bien, entrega al grafo del mundo lo que el verbo observó, antes de devolver. Si la entrega falla, el sobre
   no cambia y la línea `kitlegal: lo observado no ha llegado al grafo del mundo: …` va a la salida de error (H7).
5. Registra un evento (§4) y devuelve el resultado.

Otros casos:

| Caso | Resultado |
|---|---|
| Hallazgos en `graph_check` | sobre con `ok` verdadero y los hallazgos en `data`; no es un error (FR-012) |
| El plazo de `--timeout` vence | sobre de fallo con `clase` `fuente-no-disponible` y `isError`. El plazo es de esa llamada: el servidor no tiene ninguno y atiende las siguientes, también pasado ese tiempo desde su arranque (FR-020) |
| `--offline` con la caché vacía | `fuente-no-disponible`; `territorio_resolver` y las de `graph` responden igual que sin él |
| Una herramienta que el servidor no anuncia (`skills_install`, `mcp_serve`, `version`) | error del protocolo `-32602`, `unknown tool "<nombre>"`, sin sobre (V17) |
| Varias llamadas a la vez | cada una, su resultado (V18); las que piden al BOE esperan turno en el mismo ritmo por sitio: una petición por intervalo entre todas, el `robots.txt` de cada una incluido (D8) |
| El `robots.txt` del BOE no se puede obtener (sin conexión, 5xx, 429) | el sobre de su orden: `limite-o-tos`, o `fuente-no-disponible` si vence el plazo. No pasa de esa llamada: la siguiente lo vuelve a pedir y, si la fuente responde, lee (D8) |
| Una llamada que falla, de la clase que sea | el servidor sigue (FR-015) |
| La entrada se cierra con llamadas en curso | terminan, sus resultados no se escriben, y el servidor sale con 0 (V11) |

## 4. La salida de error

| Qué | Cuándo | Forma |
|---|---|---|
| Aviso de versión distinta (H19) | una vez por arranque, antes de servir, si las skills instaladas son de otra versión | la línea de H19, sin cambios |
| Evento de una llamada | uno por llamada: a nivel `info` si termina bien (no se ve sin `--verbose` ni `KITLEGAL_LOG`), `warn` si falla | `time=… level=WARN msg=invocación applet=boe verbo=articulo duracion=… clase=no-encontrado` (con `argumentos=…` solo en `debug`), unos 150 B |
| Mensaje de una llamada que falla | uno por llamada que falla | el mismo `mensaje` del sobre, como en la orden |
| Entrega al grafo que falla | uno por entrega que falla | la línea de H7 |
| Líneas del SDK | sus errores, siempre; lo informativo, solo con `--verbose` o `KITLEGAL_LOG=info` o `debug` (V19) | `slog`, en inglés |
| Evento de `mcp serve` | uno, al terminar | el de cualquier invocación |

Nada de esto se guarda en ningún fichero del kit.

## 5. Las `instructions`

Texto fijo, 485 bytes (FR-006; D10):

```text
Herramientas de kitlegal: consultan fuentes legales públicas españolas y devuelven un sobre con el texto vigente de una norma consolidada del BOE o con el territorio de un municipio. No afirmes ningún contenido legal que no venga del texto devuelto en esta conversación. Cada afirmación lleva su cita, con la norma y el bloque: art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]. Traslada los avisos del sobre (data.avisos). El protocolo completo son las skills de kitlegal.
```

Sus cinco frases son, en orden, los cinco contenidos de FR-006. `serverInfo` es `{"name":"kitlegal","version":"<versión
del binario>"}`.

## 6. Lo que el servidor anuncia y lo que no

`capabilities` es exactamente `{"tools":{}}` (V8): sin `listChanged`, `logging`, `prompts`, `resources` ni
`completions`. Versiones del protocolo: las del SDK, sin estrechar (FR-007; V6). Transporte: la entrada y la salida
estándar; el paquete no abre ningún puerto ni usa `net/http` (FR-001).

## 7. Código: paquetes y firmas

```go
// internal/mcp — el único paquete que importa github.com/modelcontextprotocol/go-sdk
const Instrucciones = "…"
type Herramienta struct {
	Nombre, Descripcion string
	Entrada, Salida     json.RawMessage
	Llamar              func(argumentos json.RawMessage) Resultado
}
type Resultado struct {
	Sobre []byte
	Fallo bool
}
type Servicio struct {
	Entrada      io.Reader
	Salida       io.Writer
	Version      string
	Registrador  *slog.Logger
	Herramientas []Herramienta
}
func Servir(ctx context.Context, s Servicio) error

// internal/cli — junto a describe.go
func EsquemasDeHerramienta(def Verbo) (entrada, salida []byte, err error)
func LineaDeLlamada(def Verbo, argumentos []byte) ([]string, error)

// internal/app
type DependenciasDeMCP struct {
	Entrada io.Reader
	Version string
}
func DependenciasDeMCPDelSistema(version string) DependenciasDeMCP
func AppletMCP(dependencias DependenciasDeMCP) Applet

// internal/httpx — junto a ConIntervalo
type Ritmo struct{ /* sin campos exportados */ }
func NuevoRitmo(intervalo time.Duration) *Ritmo
func ConRitmo(ritmo *Ritmo) Opcion
```

Sin exportar, en `internal/app`: la interfaz `servidor` que `ejecutarVerbo` reconoce (D4), `herramientasDe(registro, …)`
y `emitir`, el final que `Main` y cada llamada comparten. `RegistroDeProduccion` y el registro del binario de e2e
registran `AppletMCP`.

`Ritmo` son los turnos por sitio, con su intervalo, de los clientes que lo reciben con `ConRitmo`; lo demás de cada
cliente —también lo que recuerda de `robots.txt`— sigue siendo suyo. `DependenciasDeRed` crea uno con el intervalo del
BOE y construye con él un cliente por invocación, como hoy: en la orden hay uno solo y nada cambia. `ConRitmo` va en
lugar de `ConIntervalo` (dadas las dos, vale la última); un `Ritmo` nulo o de intervalo no positivo es un error de la
clase `argumentos` al construir el cliente, y `Replay` la rechaza como rechaza `ConIntervalo` (D8).

## 8. Uso, de fuera adentro

| Salida | Quién la pide y cuántas veces | Tamaño | Cuándo deja de darse |
|---|---|---|---|
| `tools/list` | el agente, una vez por conexión | diez herramientas: 1 190 B de esquemas de entrada + 19 289 B de salida + 971 B de descripciones + unos 150 B de nombre, anotaciones y claves por herramienta ≈ 23 KB. Al modelo le llegan el nombre, la descripción y el esquema de entrada: unos 2,3 KB. Crece con los verbos (≈ 2,3 KB por verbo), no con el uso | no da señales |
| Resultado de una llamada | la skill, las veces que hoy la orden (§1 de [skills.md](./skills.md)) | el sobre, dos veces en el mensaje. Art. 21 de la Ley 39/2015: 4 433 B de sobre, ≈ 9 KB en el cable. Índice de la LCSP: 34 720 B, ≈ 70 KB. Lo acotan la norma y el bloque pedidos | cada llamada es independiente |
| `graph_check` | la skill, una vez por lectura de bloques, con su norma y sus bloques | ≈ 325 B sin nada; ≈ 1 000 B por bloque cambiado (H7.1) | la señal de un bloque se apaga con su siguiente lectura (H7.1) |
| Error de herramienta | la skill, cuando una llamada falla | sobre de fallo: de 277 B (falta un argumento) a 478 B (un bloque que no existe, con su dirección) en los medidos | en la primera llamada que se hace cuando su causa ha cesado: ninguna hereda el fallo de otra, tampoco el de `robots.txt`, que cada llamada vuelve a pedir (D8) |
| `instructions` | el agente, una vez por conexión | 485 B | no da señales |
| Aviso de versión, en la salida de error | quien mira el registro de su agente, una vez por arranque | una línea | al reinstalar las skills |
| Eventos, en la salida de error | quien depura; sin `--verbose`, solo las llamadas que fallan | ≈ 150 B por línea | no es una señal para la skill |

Con cientos de normas y miles de bloques consultados, ninguna de estas salidas cambia de tamaño: el servidor no guarda
nada propio y cada sobre es el de su orden.

## 9. Tests

| Test | Dónde | Qué fija |
|---|---|---|
| Guiones `h21-mcp-*` | `TestEntregaDelHito`, `internal/app/e2e_test.go` | [arnes-e2e.md §4](./arnes-e2e.md) |
| `TestHerramientasDelServidor` | `internal/app/herramientas_test.go` | conformidad (FR-070): con el registro de producción y con uno que lleva además los applets de ejemplo, el conjunto anunciado es el de los verbos menos los excluidos; cada nombre, descripción y par de esquemas es el de `--describe` de su verbo, el de entrada sin las ocho banderas; cada `$ref` de un esquema resuelve en él; todas de solo lectura; `capabilities` es `{"tools":{}}`; y el tipo JSON que `LineaDeLlamada` exige de cada argumento es el que declara su esquema |
| `TestLlamadasSimultaneas` | `internal/app/mcp_test.go` | FR-074: N llamadas a la vez a la misma herramienta y a otras, sobre la reproducción; cada una, el sobre de su entrada; `-race` |
| `TestServirSinEntrada`, `TestServirEnEnsayo` | `internal/app/mcp_test.go` | la entrada que se cierra con una llamada en curso da 0; `--dry-run` con la entrada abierta vuelve sin leerla |
| `TestPlazoDeCadaLlamada` | `internal/app/mcp_test.go` | FR-020: con `mcp serve --timeout=<plazo corto>` y un registro del test con un verbo que espera a que termine su contexto y otro que vuelve en el acto, la llamada al primero devuelve el sobre de `fuente-no-disponible` con `isError` (§3, «Otros casos»); la llamada al segundo, hecha después y con ese plazo ya vencido desde el arranque, recibe su resultado; y el servidor termina con 0 al cerrarse la entrada |
| `TestLineaDeLlamada`, `TestEsquemasDeHerramienta` | `internal/cli/herramienta_test.go` | los casos de §3 y el orden de los posicionales |
| `TestInstrucciones` | `internal/mcp/instrucciones_test.go` | las cinco frases y los 512 bytes (FR-077) |
| `TestServir` | `internal/mcp/servir_test.go` | lo que se anuncia, el resultado de una llamada y su `isError`, y el final con la entrada cerrada |
| `TestRitmoCompartido` | `internal/httpx/ritmo_test.go` | con el servidor local de los tests del paquete y dos clientes que reciben el mismo `Ritmo`: sus cuatro llegadas —el `robots.txt` y el recurso de cada uno— no se adelantan a su turno, medido como en `TestRitmoSeparaPeticionesDelMismoSitio` (FR-014); y con un `robots.txt` que responde 429 la primera vez que se pide y sus reglas después, el primer cliente recibe `limite-o-tos` y no lo vuelve a pedir, y el segundo lo pide y lee el recurso (FR-010) |
| `TestDependenciasDeRed` | `internal/app/boe_test.go` | la composición: dos invocaciones reciben dos clientes distintos, y los dos esperan turno en el mismo `Ritmo`. Con un `Ritmo` de una hora, un plazo en el contexto y un sitio local —un `net.Listener` del test que responde 404 a todo y cuenta sus conexiones—, el primer cliente gasta el turno en su `robots.txt`, el segundo falla sin abrir ninguna conexión y uno de otras dependencias sí la abre; nada espera (FR-010, FR-014) |
| `TestOpcionesInvalidas`, `TestReplayRechazaOpciones`, `TestErrorMensajesEnEspanol` | `internal/httpx/` | las filas de `ConRitmo` en las tablas de opciones del cliente (§7) |
| `TestArquitectura` (R7), `TestDependenciasDelBinario` | `internal/arch_test.go` | el SDK, solo desde `internal/mcp`; los siete módulos nuevos, declarados |
