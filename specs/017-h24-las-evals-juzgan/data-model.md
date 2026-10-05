# Data model: H24

Tipos y ficheros que el hito añade o cambia, todos de `internal/evals` y de las carpetas que lee. El detalle de cada
salida, en `contracts/`.

## 1. El juez de una skill

`Juez`, lo que `LeerConjunto` deja en `Conjunto.Juez` cuando la carpeta de evals tiene `juez/`; `nil` si no la tiene.

| Campo | Qué es | Regla |
|---|---|---|
| `Clases` | Las de `clases.yaml`, en su orden | Al menos una; nombres sin repetir; las propiedades de `esquema.json` son exactamente sus nombres |
| `Rubrica` | El contenido de `rubrica.md`, entero | Va tal cual como instrucciones del juez |
| `Esquema` | El contenido de `esquema.json` | Con él se valida cada voto |
| `Casos` | El contenido de `casos.yaml`, entero | Su huella la compara `comprobarLaMedida`; como casos solo los leen la ejecución de la medida, el control de derivaciones y el peor caso del trabajo `medida` |
| `Medida` | El contenido de `medida.json`, entero | De él la leen `comprobarLaMedida`, el informe y el sondeo |

Los cinco ficheros se leen una vez, con el conjunto: quien usa el juez no vuelve a la carpeta.

`ClaseDelJuez`: `Nombre` (`^[a-z0-9_]+$`), `Decide`, `Umbral` (0 a 1). Con `Decide`, la clase tiene su umbral del
informe con `decide: true` y sus dos umbrales de la medida; sin él, se publica.

## 2. El texto de una herramienta

`Texto`: `Orden` y `Salida`. `Sesion.Textos` los lleva en el orden del transcript
([contracts/juez-y-voto.md](./contracts/juez-y-voto.md) §2). En un caso etiquetado, los da la reconstrucción.

## 3. El voto

| Tipo | Campos | Notas |
|---|---|---|
| `Votante` | `func(mensaje string) ([]byte, error)` | Devuelve la salida estándar de la sesión del juez; el error es el del proceso del voto: con el del tope, el voto no llega a darse; con cualquier otro, la salida que dejó se lee igual, y su código va en el motivo si no es JSON |
| Voto de una clase (`VotoDeClase`) | `Voto` (1 a 3), `Nulo`, `Motivo`, `Respuesta` (`si` o `no`), `Frase`, `Precepto` (si la clase lo tiene), `FraseEnLaRespuesta` | Sus claves JSON son las de [contracts/informe-del-job.md](./contracts/informe-del-job.md) §3 |
| Juicio de una respuesta (`JuicioDeRespuesta`) | Por clase (`JuicioDeClase`), su nombre, sus votos y `Marcada`; y, si quedó sin juzgar, su motivo. La sesión la pone el informe | Lo da la regla |

Estados de una respuesta juzgada, por clase que decide:

```text
sin votar ── voto 1 ──► no ──► sin marcar (fin)
                 └────► sí ── voto 2 ──► no ──► sin marcar (fin)
                                  └────► sí ── voto 3 ──► no ──► sin marcar, con dos frases (fin)
                                                   └────► sí ──► marcada (fin)
cualquier voto que no llega a darse ──► sin juzgar (fin)
un voto nulo ──► se repite una vez; cuenta el repetido
```

## 4. La medida y los casos

`MedidaDelJuez`: `Clase`, `Fecha`, `ModeloDelJuez`, `VersionDeClaudeCode`, la huella de la rúbrica y la de los casos,
`Defectos` (`Casos`, `SinMarcar`) y `Correctos` (`Casos`, `Marcados`). **Corresponde** si las cuatro claves son las de
lo que hay; **se cumple** si los dos recuentos son 0 ([contracts/medida-del-juez.md](./contracts/medida-del-juez.md)
§2).

`CasoEtiquetado`: `Informe`, `Sesion`, `Quitado` (`Norma`, `Bloque`; solo en un derivado), `Grupo`, `Etiqueta`
(`defecto` o `correcto`), `Procedencia`, `Frase`. Resuelto, lleva además la pregunta, la respuesta y los textos.

| Relación | Regla |
|---|---|
| Caso → informe | `Informe` es uno de los versionados y tiene la sesión `Sesion` |
| Caso → eval | La que nombra esa sesión: en `evals/<skill>/` o en `testdata/evals/retiradas/` |
| Derivado → sesión | Sus textos son los de la sesión menos los del bloque `Quitado`, y se quita al menos uno |
| Medida → caso | Un defecto cuenta en `SinMarcar` si no queda marcado; un correcto, en `Marcados` si queda marcado; un caso sin juzgar impide dar la medida |

## 5. El informe

| Tipo | Cambio |
|---|---|
| `InformeAEscribir` | Gana `Votar`, `ModeloDelJuez`, `VersionDelJuez`, `ConcurrenciaDelJuez`, `Ahora` e `InstrumentoSinMedir` (las líneas de `comprobarLaMedida`; con alguna, el informe es el de la medida que no corresponde) |
| `Informe` | Gana `Juez` (`juez`); pierde `ExpresionesProhibidasPorModelo` |
| `ResultadoDeEval` | Pierde `ExpresionesProhibidas`; su `Pasa` y sus `Motivos` ya no dependen de la lista |
| `Umbral` | Sin cambios de forma. Nombres nuevos: `afirma_lo_no_leido:…`, `cuenta_su_proceso:…`, `medida_del_juez:…` y `duracion_del_juez:…`; salen `expresiones_prohibidas:…` y `redaccion_no_leida:…` |
| `Eval` | Pierde `Prohibidas`: la lista se queda en `Conjunto.Prohibidas`, para la prosa |
| `ExpresionesProhibidas` | Gana `SalidaDeLasHerramientas` (`salida_de_las_herramientas`), que no está entre las expresiones que se buscan en una respuesta |

## 6. El job

`EjecucionDelJob` (`ejecucion.go`) es lo que `ejecutarElJob` necesita: lo de hoy de `TestEjecucionDelJob` —la skill,
el plan, el repartidor como función, los directorios— más el votante, el modelo y la versión del juez.

`DefinicionDelJob` gana `ModeloDelJuez`, `VersionDelJuez`, el trabajo `medida` (su nombre, su condición, su `env`, sus
skills con su concurrencia, su tope y si tiene `needs`), las entradas del despacho y el `run` del paso de instalación.

## 7. Ficheros

| Ruta | Estado | Tarea |
|---|---|---|
| `evals/boe-legislacion/juez/{clases.yaml,rubrica.md,esquema.json,casos.yaml,medida.json}` | Nuevos: la declaración y cuatro copias | `clases.yaml`, con su esquema, `[datos]` |
| `schemas/juez-clases.yaml.json` | Nuevo | `[datos]` |
| `schemas/expresiones-prohibidas.yaml.json`, `evals/boe-legislacion/expresiones-prohibidas.yaml` | Ganan `salida_de_las_herramientas` | `[datos]` |
| `testdata/evals/retiradas/…` (dos ficheros) | Restaurados de `c4819d1^` | `[datos]` |
| `skills/boe-legislacion/SKILL.md` | v0.1.7 | — |
| `scripts/evals-voto.sh`, `scripts/evals-medir-juez.sh` | Nuevos | — |
| `evidencias/adr-0037/` | No se toca | — |
