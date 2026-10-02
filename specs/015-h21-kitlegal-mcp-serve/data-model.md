# Data model · H21 · `kitlegal mcp serve`

Entidades del spec («Key Entities») con su forma en el código. Los contratos dan los bytes y los ejemplos; aquí, los
campos, las reglas y de dónde sale cada uno. Nada de esto se guarda en disco salvo donde se dice.

## 1. Herramienta (`internal/mcp.Herramienta`)

Una operación del servidor. La compone `internal/app` desde el registro, una por verbo; `internal/mcp` la anuncia.

| Campo | Tipo | De dónde sale | Regla |
|---|---|---|---|
| `Nombre` | `string` | `<applet>.Nombre() + "_" + <verbo>.Nombre` | único; FR-003 |
| `Descripcion` | `string` | `Verbo.Descripcion` | la de la ayuda y de `--describe` |
| `Entrada` | `json.RawMessage` | `cli.EsquemasDeHerramienta` | la `entrada` de `--describe` sin las ocho banderas globales; `type: object` |
| `Salida` | `json.RawMessage` | `cli.EsquemasDeHerramienta` | la `salida` de `--describe`: el sobre, con `data` condicionado a `ok`, y sus `$defs` |
| `Llamar` | `func(json.RawMessage) Resultado` | cierre de `internal/app` | una invocación del kernel (§2) |

**Qué verbos dan herramienta** (FR-002): todos los de cada applet de `Registro.Nombres()`, salvo los de los applets
`skills` y `mcp`. Los verbos reservados (`version`) no son de ningún applet. Todas se anuncian con `readOnlyHint`
verdadero (FR-005).

## 2. Llamada y resultado (`internal/mcp.Resultado`)

| Campo | Tipo | Regla |
|---|---|---|
| `Sobre` | `[]byte` | los bytes que la orden del verbo escribe en la salida estándar con `--json`, sin el salto de línea final: un sobre de éxito o de fallo |
| `Fallo` | `bool` | verdadero si y solo si el código de la invocación no es 0, que es cuando el sobre lleva `ok` falso (FR-011) |

Estados de una llamada: recibida → argumentos convertidos (o rechazados: `argumentos`) → verbo ejecutado con su plazo →
sobre montado → **entrega al grafo, si terminó bien y sin `--no-graph`** → resultado devuelto. Lo que una llamada deja
es lo que deja la orden (caché y grafo del mundo). En memoria, entre llamadas, solo queda el turno de cada sitio
(`internal/httpx.Ritmo`, uno por proceso, que crea `DependenciasDeRed`): el cliente HTTP es de cada llamada, y con él lo
que obtiene de `robots.txt` (research D8).

Correspondencia de la clase del sobre de fallo con el código de la orden (ADR 0023): `argumentos` 2, `no-encontrado` 3,
`fuente-no-disponible` 4, `limite-o-tos` 5, `identidad-humana` 6, `conflicto` 7, `inesperado` 1.

## 3. Servicio (`internal/mcp.Servicio`)

Lo que `Servir` necesita: `Entrada io.Reader`, `Salida io.Writer`, `Version string` (la del binario, para `serverInfo`),
`Registrador *slog.Logger` y `Herramientas []Herramienta`. Vive lo que vive el proceso. Termina bien cuando la entrada
llega a su fin.

## 4. `instructions` (`internal/mcp.Instrucciones`)

Constante de texto: cinco frases, 485 bytes, tope de 512 (FR-006). Texto en contracts/servidor-mcp.md §5.

## 5. Dependencias del applet (`internal/app.DependenciasDeMCP`)

| Campo | Tipo | Producción | Tests |
|---|---|---|---|
| `Entrada` | `io.Reader` | `os.Stdin` | el extremo de lectura de una tubería |
| `Version` | `string` | la del binario | `""` |

## 6. Modo de una sesión de eval (`internal/evals.Modo`)

| Valor | Texto | Qué tiene la sesión | Cómo se reconoce al leerla |
|---|---|---|---|
| `ModoOrden` | `orden` | `kitlegal` en el `PATH`, ningún servidor | no tiene `servidor.json` y su eval no es sin binario ni servidor |
| `ModoHerramienta` | `herramienta` | el servidor declarado, `kitlegal` fuera del `PATH` | tiene `servidor.json` |
| (ninguno) | `""` | ni binario ni servidor | su eval declara `sin_binario_ni_servidor` |

Lo llevan `SesionPlanificada`, `SerieDeSesiones`, la clave de serie del informe, `ResultadoDeEval`, `TasaDelInforme` y
`RecuentoDeExpresiones`. `PlanDeEvals.Modos` son los modos que el plan abre: los dos en el job; solo `orden`, que es su
valor por omisión, en el sondeo.

## 7. Eval sin binario ni servidor (`Eval.SinBinarioNiServidor`)

Clave `sin_binario_ni_servidor: true` del formato de eval. Reglas del esquema: solo con `activa: true`; sin `comandos`,
`prohibidos`, `grafo_previo`, `citas`, `avisos`, `hallazgos`, `redacciones_modificadas`, `territorio`, `informativa` ni
`reproduce`. Decide siempre. Se mide una vez por modelo, fuera de los dos modos.

## 8. Llamada leída del transcript (`internal/evals.Llamada`)

| Campo | Tipo | De dónde |
|---|---|---|
| `Herramienta` | `string` | el `name` del bloque `tool_use`, sin lo que precede a su último `__` |
| `Argumentos` | `json.RawMessage` | su `input` |
| `ConResultado` | `bool` | si el transcript trae su `tool_result` |
| `Error` | `bool` | `is_error` verdadero, o contenido que es un sobre con `ok` falso |
| `Clase` | `schema.Clase` | `data.clase` de ese sobre; vacía si no se puede leer |

Solo se leen las de nombres que son de una herramienta del registro de producción. `Sesion.Llamadas` las lleva en el
orden del transcript.

## 9. Línea `⚠ SIN CONSULTA AL BOE:`

Forma fija de la respuesta: la marca `⚠`, la etiqueta `SIN CONSULTA AL BOE`, dos puntos, la causa y, en la misma línea,
`https://kitlegal.es/instalar/`. La reconoce `ExtraerSinConsulta` con la misma tolerancia que las etiquetas de los avisos
(blancos y énfasis de Markdown alrededor). No se guarda.

## 10. Umbral por modo

Un elemento de `umbrales` (contrato del ADR 0029, sin cambios en sus campos) medido sobre las sesiones de un solo modo.
Nombres: `expresiones_prohibidas:<modelo>:<modo>`, `sin_activar:<modelo>:<modo>`, `redaccion_no_leida:<modelo>:<modo>`
y `duracion_de_las_sesiones:<modo>`, con `<modo>` `orden` u `herramienta`. Patrón `^[a-z0-9_.:-]+$`.

## 11. Ficheros

| Fichero | Quién lo escribe | Vida |
|---|---|---|
| `<sesión>/servidor.json` | el repartidor del job, en las sesiones del modo herramienta | la del directorio de la sesión |
| `schemas/servidor.json` | `TestEsquemasPublicados -actualizar-esquemas` | versionado |
| `evals/boe-legislacion/21-sin-binario-ni-servidor.yaml`, `evals/legal-core/04-sin-binario-ni-servidor.yaml` | una tarea | versionados |

El servidor no crea ningún fichero propio: la caché y el grafo del mundo son los de las órdenes, en el mismo sitio.
