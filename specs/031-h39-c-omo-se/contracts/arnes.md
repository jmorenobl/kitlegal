# Contrato: el arnés del sondeo del laboratorio

FR-029 a FR-033; SC-007. Research D7, D8; data-model §6.

## 1. Cómo se lanza

Lo lanza una persona, con la credencial del sondeo (ADR 0032); ni `make ci`, ni un gancho, ni el job de evals, ni un
paso del workflow (FR-032):

```sh
make evals-laboratorio VARIANTES=<v>[,<v>…] MODOS=<m>[,<m>] SESIONES=<n> MODELO=<id> CLAUDE_CODE=<versión> \
  CONCURRENCIA=<c> DIRECTORIO=<ruta>
```

Cadena: `Makefile` → `scripts/evals-sondeo-llavero.sh laboratorio <los siete>` (toma la credencial como hoy y ejecuta
el guion que nombra su primer argumento, de una lista cerrada: `sondeo` → `scripts/evals-sondeo.sh`, `laboratorio` →
`scripts/evals-laboratorio.sh`; `make evals-sondeo` pasa a llamarlo con `sondeo` delante) →
`scripts/evals-laboratorio.sh` (crea su temporal y lo borra al salir; `DIRECTORIO` relativo, desde la raíz del
repositorio) → `go test -tags evals -count=1 -timeout 0 -run '^TestSondeoDelLaboratorio$' ./internal/evals/ -args …` →
`sondearElLaboratorio`. Como `scripts/evals-sondeo.sh`: un error de uso deja su mensaje en `uso.txt`, el guion lo
imprime en la salida de error y sale con 1; otro error imprime el registro de `go test` y sale con 1; si no, imprime
`salida.txt` y sale con 0.

## 2. Argumentos y lo que se comprueba antes de abrir ninguna sesión

Todos los errores de argumentos van juntos, uno por línea, cada uno con el nombre del argumento de `make`:

| Argumento | Vale si | Mensaje si no |
|---|---|---|
| `VARIANTES` | nombres de `variantes.yaml`, separados por comas, sin repetir | `VARIANTES: «<v>» no es ninguna variante del laboratorio` / `… se repite` |
| `MODOS` | `orden`, `herramienta` o los dos, separados por comas, sin repetir | `MODOS: «<m>» no es orden ni herramienta` |
| cada par variante × modo | el modo está en los `modos` de la variante (FR-020 a FR-027) | `VARIANTES y MODOS: <v> no se mide en modo <m>` |
| `SESIONES` | entero mayor o igual que 4 y múltiplo de 4 | `SESIONES: «<n>» no es un múltiplo de 4 mayor que 0` |
| `MODELO` | `claude-<familia>-<versión>`: `^claude-[a-z]+-[0-9]+(-[0-9]+)*$` | `MODELO: «<id>» no es un identificador completo de modelo (claude-<familia>-<versión>, como claude-sonnet-5-5)` |
| `CLAUDE_CODE` | no vacía | `CLAUDE_CODE: está vacía` |
| `CONCURRENCIA` | entero mayor o igual que 1 | `CONCURRENCIA: «<c>» no es un entero mayor o igual que 1` |
| `DIRECTORIO` | no vacío y no existe | `DIRECTORIO: está vacío` / `DIRECTORIO: <ruta> ya existe` |

Detrás, en este orden y solo si los argumentos valen: la credencial (`CLAUDE_CODE_OAUTH_TOKEN` no vacía, el mensaje de
hoy, `errSinSuscripcion`); y `claude --version` con el `PATH` de quien lo lanza, cuyo primer campo tiene que ser
`CLAUDE_CODE` (research S1): si no, `CLAUDE_CODE: la versión de Claude Code instalada es <v> y se pidió <pedida>`, o
`CLAUDE_CODE: claude --version no se pudo ejecutar: <error>`. Todos son errores de uso: ninguna sesión se abre, no se
construye nada y no se crea `DIRECTORIO` (FR-030, FR-031, SC-007).

## 3. Lo que prepara y abre

1. En el temporal: `kitlegal` del árbol (`go install -trimpath ./cmd/kitlegal`) en `bin/`; `laboratorio`
   (`go install -trimpath ./cmd/laboratorio`) en `laboratorio-bin/`, fuera del `PATH` de las sesiones; `home/` vacío;
   y en `skills/<variante>/` una copia de `laboratorio/variantes/<variante>/skills/`, que es el directorio de skills del
   reparto: cada sesión tiene solo la skill de su variante y, en `entrada`, la de entrada (FR-030).
2. `DIRECTORIO/`, con los directorios que le falten delante, y `DIRECTORIO/sondeo.json`:

   ```json
   {"modelo":"claude-sonnet-5-5","claude_code":"2.1.317"}
   ```

3. Por cada variante pedida, en el orden de `variantes.yaml`, y por cada modo pedido, `orden` antes que `herramienta`:
   `ejecutarSesiones` con `Sesiones` = `DIRECTORIO/<variante>/<modo>/`, `Skills` = la copia de la variante, `Evals` =
   `laboratorio/evals`, el guion de sesión de hoy (`scripts/evals-sesion.sh`), el tope de hoy, la `CONCURRENCIA` y un
   plan de `SESIONES/4` sesiones de cada eval, en el orden de sus ficheros, con el `MODELO` y el modo, de nombre
   `<eval sin .yaml>-<nn>` (FR-030, FR-033). Las sesiones se sirven de la grabación de hoy con los proxies que rechazan
   toda petición (ADR 0043).
4. En modo herramienta, `servidor.json` arranca `kitlegal mcp serve` (la orden de hoy) salvo en `instrucciones` y
   `parametro`, donde `Servidor` es `<laboratorio-bin>/laboratorio servidor -kitlegal <bin>/kitlegal` seguido de
   `-instrucciones <texto de instrucciones.txt>` o de `-herramienta boe_articulo -parametro bloque -descripcion <texto de
   parametro.txt>`. `SesionesAEjecutar.Servidor` vacío deja `servidor.json` byte a byte como hoy.
5. Tras un reparto que vuelve con sesiones sin abrir por un límite de uso de la cuenta, no abre más: las de los repartos
   que faltan quedan sin abrir. SIGINT y SIGTERM cierran las abiertas como hoy y no se abre ninguna más.

## 4. Salida

`salida.txt`, una línea por variante y modo pedidos, en el orden de §3.3, y la ruta:

```text
Sondeo del laboratorio en /Users/…/kitlegal-sondeos/h39-a, con claude-sonnet-5-5 y Claude Code 2.1.317.
- control, orden: 40 abiertas.
- control, herramienta: 40 abiertas.
- referencia-paso, herramienta: 25 abiertas, 15 sin abrir tras el límite de uso.
La tabla: make laboratorio-tabla DIRECTORIO=/Users/…/kitlegal-sondeos/h39-a
```

La lee la persona al terminar cada lanzamiento: como mucho 14 líneas de unos 60 bytes y dos más. No crece con el uso.
Las cifras son de la tabla (contracts/tabla.md), no de esta salida.

## 5. `laboratorio servidor` (FR-021, FR-024, FR-029)

`laboratorio servidor -kitlegal <ruta> [-instrucciones <texto>] [-herramienta <h> -parametro <p> -descripcion <texto>]`:
arranca `<ruta> mcp serve` con el entorno del proceso; copia su entrada estándar a la del hijo sin tocarla; escribe la
salida de error del hijo en la suya; y de la salida estándar del hijo, línea a línea (un mensaje JSON por línea,
research V13), reescribe solo:

- la respuesta a `initialize` (un objeto con `result.instructions`): con `-instrucciones`, ese texto;
- la respuesta a `tools/list` (un objeto con `result.tools`): con `-herramienta`, en la herramienta de ese `name`, la
  `description` de `inputSchema.properties.<p>`, que pasa a ser `-descripcion`;

y deja las demás líneas byte a byte. Termina con el código del hijo cuando el hijo termina. Unos argumentos que no
tienen esa forma, o una herramienta o un parámetro que la lista no trae, son un defecto: lo dice en la salida de error y
sale con 1. No importa el SDK de MCP (R7). `kitlegal` no gana ninguna bandera, variable ni orden (FR-029).
`TestServidorDelLaboratorio` lo comprueba con las líneas que escribe el servidor de producción en proceso
(`mcptest.Sesion.Lineas`): con cada opción cambia solo su campo, y sin ninguna, nada.

## 6. Controles

- `TestArnesDelLaboratorio` (`internal/evals`, en `make ci`, sin modelo): `sondearElLaboratorio` con la preparación del
  árbol sustituida y un sustituto de `claude` en el `PATH` que contesta a `--version` y anota cada sesión que abre. Con
  la versión distinta, con `MODELO=sonnet`, con `VARIANTES=instrucciones MODOS=orden` y con `SESIONES=6`, 0 sesiones y
  el mensaje de §2; con una petición válida (dos variantes, una con un modo y otra con dos, `SESIONES=8`), dos sesiones
  de cada eval en cada variante y modo pedidos, y ninguna más (FR-030, FR-031, SC-007).
- `TestRecetasDelSondeo` (`internal/evals`, en `make ci`, sin modelo): ejecuta desde la raíz del repositorio, sin el
  `MAKEFLAGS` del `make` que lo lanza, `make -n evals-sondeo …` y `make -n evals-laboratorio …` con valores fijos, y
  exige en cada salida una sola línea del guion del llavero, igual carácter a carácter a la cadena de §1:
  `scripts/evals-sondeo-llavero.sh sondeo` y sus cinco argumentos, y `scripts/evals-sondeo-llavero.sh laboratorio` y sus
  siete, entre comillas dobles y en su orden. `-n` solo imprime: ningún guion, credencial ni sesión (FR-030, FR-033).
  Es el control de las dos recetas: la política de las sesiones del run veda `make … evals-sondeo` como orden.
- `TestServidorDelLaboratorio` (§5) y el test de hoy de `servidor.json` con `Servidor` vacío.
- `TestSondeoDelLaboratorio` (`-tags evals`) no corre en `make ci`: es la entrada de la persona.
