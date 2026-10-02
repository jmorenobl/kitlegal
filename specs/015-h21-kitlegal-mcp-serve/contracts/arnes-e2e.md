# Contrato · el arnés e2e de H21 y los guiones de aceptación

Lo que el arnés de `internal/app/e2e_test.go` gana para que un guion `testscript` hable con `kitlegal mcp serve` por un
cliente MCP, y los guiones que describen la entrega. Amplía, sin cambiar nada de ellos, los contratos del arnés de H1,
H4, H19 y H7. Las referencias V y D son de [research.md](../research.md).

## 1. Por qué una orden nueva

Un servidor MCP por la entrada estándar no se puede ejercer con `stdin fichero` + `exec`: si todas las peticiones llegan
de golpe y la entrada se cierra, el servidor no escribe ninguna respuesta (V12). Hace falta un cliente que espere cada
respuesta antes de enviar la siguiente y que cierre la entrada al final.

## 2. El binario de e2e

`internal/app/ejemplo/kitlegal-e2e` registra además `app.AppletMCP(app.DependenciasDeMCPDelSistema(version))`, como el
binario distribuido. Lo demás no cambia: boe sobre la reproducción de `reproduccion/` del directorio de trabajo, el reloj
de `KITLEGAL_T0_BIN`, `KITLEGAL_T1_BIN` y `KITLEGAL_T8_BIN`, las versiones de `KITLEGAL_V1_BIN` y `KITLEGAL_V2_BIN`. Su
registro lleva los applets de ejemplo (`echo`, `contar`), así que su servidor anuncia trece herramientas: las diez de
producción, `contar_letras`, `contar_palabras` y `echo_repetir`. Es el «registro con un applet más» de FR-004.

## 3. La orden `mcp`

```text
mcp [-anterior] [-dir <directorio>] <llamadas> <programa> [<argumento>…]
```

Arranca `<programa> <argumento>…` con el entorno del guion y con `<directorio>` como directorio de trabajo (`$WORK` si no
se da; relativo a `$WORK` si no es absoluto), le conecta un cliente por su entrada y su salida estándar, hace en orden
las operaciones de `<llamadas>` esperando cada respuesta, cierra la entrada y espera a que el programa termine.

- **Cliente**: sin `-anterior`, el del SDK del protocolo, que negocia la versión vigente (2026-07-28). Con `-anterior`,
  uno de la especificación 2025-11-25, escrito con la biblioteca estándar: `initialize` con
  `"protocolVersion":"2025-11-25"`, `notifications/initialized`, y después `tools/list` y `tools/call` (V15).
- **`<llamadas>`**: un fichero de `$WORK` con una operación por línea; las líneas vacías no cuentan. Cada operación tiene
  su número, desde 1:
  - `herramientas`: lista las herramientas;
  - `<herramienta> <objeto JSON>`: llama a la herramienta con esos argumentos;
  - `<herramienta>`: la llama sin `arguments`. Con `-anterior` no se envían; el cliente del SDK no puede no enviarlos
    y manda `{}`.

`<programa>` se da por su ruta absoluta: la orden no lo busca en ningún `PATH`, y otra cosa es un error de uso. La
conversación entera tiene un tope de un minuto; al agotarse, la orden interrumpe el programa y el guion falla. La orden
no retira lo que dejó otra ejecución suya en el mismo guion.

**Lo que escribe en su salida estándar** (la que el guion lee con `stdout` o `cmp stdout`), una línea por cosa:

```text
protocolo <versión negociada>
capacidades <nombres de las capacidades anunciadas, ordenados y separados por un espacio>
<n> herramientas <cuántas>
<n> <herramienta> resultado
<n> <herramienta> error <clase>
<n> <herramienta> protocolo <mensaje del error del protocolo>
salida <código con el que terminó el programa>
```

**Lo que deja en `$WORK`**:

| Fichero | Contenido |
|---|---|
| `mcp-<n>.json` | de una llamada con resultado o con error de herramienta: el texto de su bloque de texto y un salto de línea. Es comparable con `cmp` con la salida estándar de la orden con `--json` |
| `mcp-<n>.txt` | de `herramientas`: una línea por herramienta, en el orden del servidor: `<nombre> lectura=<true\|false>` |
| `mcp.err` | la salida de error del programa, entera |
| `mcp.out` | solo con `-anterior`: cada línea que el programa escribió en su salida estándar |
| `mcp.instrucciones` | las `instructions` que dio el servidor y un salto de línea |

**Lo que comprueba por su cuenta**, y hace fallar el guion:

- que `structuredContent` de cada resultado es el mismo documento JSON que su bloque de texto, y que hay un solo bloque;
- que `isError` es verdadero si y solo si el sobre lleva `ok` falso;
- con `-anterior`, que cada línea de la salida estándar del programa es un mensaje JSON-RPC 2.0 (un objeto con
  `"jsonrpc":"2.0"`) y que a cada petición con `id` le llega su respuesta;
- que el programa termina tras cerrarle la entrada.

**Si el programa termina sin completar el saludo** (lo que hacen `--asunto` y `--dry-run`): escribe `sin respuesta` y
`salida <código>`, deja `mcp.err` (y `mcp.out` con `-anterior`), y la orden falla; el guion la escribe con `!` delante.
En ese caso la entrada del programa sigue abierta hasta que termina: es lo que distingue «termina sin esperar a que se
cierre la entrada» (FR-022) de un cierre.

**Dónde vive**: la orden se registra en `Cmds` de `ejecutarGuiones`; los dos clientes, en `internal/mcp/mcptest`
(`Abrir(ctx, lee, escribe, anterior) (*Sesion, error)`, con `Protocolo`, `Instrucciones`, `Capacidades`, `Herramientas`,
`Llamar`, `Lineas` y `Cerrar`), que es donde `depguard` deja importar el SDK. `internal/app` no lo importa fuera de sus
tests. Los mismos clientes sirven a `TestHerramientasDelServidor` y `TestLlamadasSimultaneas`, en proceso.

## 4. Los guiones de aceptación

Los escribe la primera tarea en `specs/015-h21-kitlegal-mcp-serve/aceptacion/` y quedan congelados; el workflow los
activa en `internal/app/testdata/script/` con el prefijo `h21-` tras el bucle de tareas. Ninguno necesita una grabación
nueva: usan las de H4 que el arnés copia en `reproduccion/` y las derivadas de H7 y H7.1 de `derivadas/` (D23).

**Rojo primero, por una aserción.** `scripts/workflow/aceptacion.sh rojo-primero` da por inválido un guion cuya salida
lleva `unknown command`, `unexpected command` o `usage: `: es lo que daría hoy la orden `mcp` del arnés, que todavía no
existe, o un `exec` que falla. Por eso cada guion empieza, antes de cualquier otra orden, por esta precondición, que hoy
se ejecuta bien y falla en su aserción (la ayuda de hoy lista `boe`, `graph`, `skills` y `territorio`, y empieza por
`uso: `, en minúsculas y sin `usage: `):

```text
exec kitlegal --help
stdout '^\s+mcp\s'
```

| Guion | Historia | Qué comprueba | FR / SC |
|---|---|---|---|
| `mcp-llamadas.txtar` | US1 (1, 5, 6), US6 | Con `KITLEGAL_T0_BIN`: para cada una de las diez herramientas de producción, la orden con `--json` y después la llamada, en la misma caché, y `cmp mcp-<n>.json` con la salida de la orden; `territorio_resolver`, con un municipio cubierto (Leganés) y con uno no cubierto (Tordesillas), como pide la matriz territorial (constitución IX). Con una lectura previa de otra redacción (derivadas, como `h7-grafo-version-obsoleta`): `boe_articulo` y después `graph_check` con esa norma y ese bloque dan `resultado`, y el sobre lleva `version-obsoleta`. Tras `boe_articulo`, `graph_show` del id del bloque da `resultado`; con `mcp serve --no-graph` y el grafo vacío, `error no-encontrado` | FR-010, FR-012, FR-013, FR-020; SC-004 |
| `mcp-errores.txtar` | US1 (2, 3, 4) | En un mismo proceso: un bloque que no existe, `error no-encontrado`; sin `bloque`, con una propiedad de más, con `offline: true` y con un número donde va una cadena, `error argumentos`; `skills_install`, `mcp_serve` y `version`, `protocolo`; y después una llamada que termina bien. Con `mcp serve --offline` y la caché vacía, `boe_articulo` da `error fuente-no-disponible` y `territorio_resolver`, `resultado`. Cada `mcp-<n>.json` de error lleva `"ok":false` y su `clase` | FR-011, FR-015, FR-020; SC-004 |
| `mcp-herramientas.txtar` | US6 (1, 2), US4 (6) | `herramientas` contra el binario de e2e: `cmp` de `mcp-1.txt` con las trece líneas esperadas, todas `lectura=true`; `capacidades tools` y nada más; las `instructions`, con sus cinco frases | FR-001 a FR-005, FR-006, FR-008; SC-003 |
| `mcp-protocolo.txtar` | US4 (1, 2) | El mismo binario con el cliente vigente (`protocolo 2026-07-28`) y con `-anterior` (`protocolo 2025-11-25`): la misma lista de herramientas y el mismo `mcp-<n>.json` de una misma llamada. Con `KITLEGAL_V2_BIN`, las skills instaladas por `KITLEGAL_V1_BIN` y `mcp serve --verbose`, con `-anterior` y dos llamadas: la orden no falla (cada línea de la salida estándar es del protocolo) y el aviso de versión está en `mcp.err` exactamente una vez | FR-007, FR-023; SC-005, SC-006 |
| `mcp-proceso.txtar` | US4 (3, 4, 5, 7) | `salida 0` al cerrar la entrada. `! mcp … mcp serve --asunto x`: `sin respuesta`, `salida 2`, el mensaje de `argumentos` en `mcp.err` y `mcp.out` vacío. `! mcp … mcp serve --dry-run`: `sin respuesta`, `salida 0`, `mcp.out` vacío y la descripción del kernel en `mcp.err`. El binario copiado a `con espacios/kitlegal` y `-dir /`: `territorio_resolver` da el sobre de la orden, y `world.db` aparece en `$WORK/cache` | FR-021, FR-022, FR-024, FR-025; SC-008 |

US2, US3, US5 y US7 no tienen guion: no son comportamiento del binario (las dos `SKILL.md`, el juicio y el informe del
job, la documentación). Sus controles, en plan.md, «Aceptación e2e».

## 5. Tests del propio arnés

Las comprobaciones de §3 las hacen los clientes de `internal/mcp/mcptest`, y ahí se prueban: `TestSesion`
(`internal/mcp/mcptest/sesion_test.go`), contra servidores montados en proceso con el SDK sobre tuberías, fija lo que
dan `Protocolo`, `Capacidades`, `Herramientas` y `Llamar` con los dos clientes, y el error con un `structuredContent` que
no es el del texto, con `isError` contrario a `ok`, con una línea que no es del protocolo (cliente anterior) y con la
conexión que se cierra antes del saludo. La forma de la salida y de los ficheros de la orden la fijan los propios guiones,
que la comparan con `cmp`. Lo que un guion no puede ejercer —el uso de la orden, qué deja pasar `!`, el programa que no
termina tras cerrarle la entrada o que ni saluda ni termina— lo fija `TestOrdenMCP` (`internal/app/e2e_test.go`), con
sustitutos del servidor. `TestVariablesDeLosGuiones` no cambia: la orden no añade variables.
