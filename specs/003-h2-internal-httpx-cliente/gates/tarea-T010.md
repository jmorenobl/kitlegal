# T010 — intento 1: sin marcar; `make ci` en rojo por `misspell` sobre dos palabras que fija el contrato

**Estado**: el código y el test de la tarea están escritos y en verde (21 subpruebas); `make ci` **no** lo
está, y el paso de reparación no puede ponerlo en verde sin tocar un fichero que la tarea no declaraba.
Es la contingencia S2 (`misspell` sobre español nuevo) en su variante prevista pero no asignada: el término
lo fija un contrato y no se puede reescribir. Salida aplicada: anotar aquí, **redelimitar la tarea** en
`tasks.md` con la misma acotación que ya lleva T003, dejar rastro en `research.md` S2 y dejar la tarea
como `[ ]`. Lo que falta para cerrarla es una edición de dos entradas en `.golangci.yml`, descrita abajo
al detalle y **comprobada** sobre una copia.

## Lo que sí quedó hecho (y verde)

- `internal/httpx/nombre.go`: derivación del nombre de fichero a partir del método y de la dirección
  completa, con las cinco reglas del contrato de grabación §2 (host en minúsculas, puerto solo si es
  explícito, `_q_` para la consulta, saneado a `[A-Za-z0-9._-]` con colapso y recorte de guiones bajos, y
  recorte a 100 caracteres más `-` y ocho dígitos de `sha256("<MÉTODO> <url>")` pasados los 120). Nada
  exportado: la deriva el propio paquete (FR-039, D12).
- `internal/httpx/nombre_test.go`: `TestNombreDeGrabacion` con las dos tablas —las siete filas del
  contrato §2, el par que colisiona (`/a,b` y `/a_b`) incluido, y los diez pares petición → fichero de
  plan.md §«Fixtures»—, más la comprobación de portabilidad (`<>:"/\|?*`, espacios y tope de 120) en
  cada fila. Solo `fuente.prueba`, `otra.prueba` y `127.0.0.1` (obligación 11: la orden de los
  prerrequisitos del quickstart sobre este fichero imprime «solo direcciones locales»).
- Rojo antes que verde: el test se escribió primero y falló por `undefined: nombreDeGrabacion`. El
  resumen `d11bc93b` de la fila larga del contrato §2 coincide con el que produce el código.
- Todo lo demás de `make ci` en verde: `fmt-check`, `make test` (`-race -shuffle=on`; `internal/httpx`
  al 97,9 % de cobertura), `vuln`, `schema-check`, `secrets`, `mod-verify` y `mod-tidy-check`. No se
  tocó `testdata/`, ni `schemas/`, ni la red, ni `KITLEGAL_RECORD`.

## Lo que deja `make ci` en rojo

```
internal/httpx/nombre_test.go:110:56: `legislacion` is a misspelling of `legislation` (misspell)
internal/httpx/nombre_test.go:118:66: `resolucion` is a misspelling of `resolution` (misspell)
```

Son la **segunda columna** de las filas 4 y 5 de la tabla del contrato de grabación §2
(`contracts/formato-de-grabacion.md`), que `TestNombreDeGrabacion` «copia tal cual» por mandato del propio
contrato. Las direcciones de la primera columna contienen las mismas palabras y **no** se marcan: `misspell`
descarta las URL antes de buscar; el nombre derivado ya no es una URL. Los otros dos avisos del primer
`make lint` (`componentes`, `reproduccion`) eran prosa propia y se reescribieron en el sitio.

## Por qué la reparación no lo cierra en este paso

Se descartaron, por orden, todas las salidas que estaban al alcance del paso:

- **Reescribir el término** (primer remedio de research S2): imposible sin cambiar el ejemplo del
  contrato, que además reproduce a propósito la forma real de la dirección del BOE (`legislacion-
  consolidada`) y determina dónde corta la regla 5 (157 caracteres, corte en el 100, resumen `d11bc93b`).
  Cambiar el contrato para contentar a un diccionario inglés deja el artefacto peor.
- **`//nolint:misspell`** con explicación: lo admite la constitución en la letra, pero `tasks.md` §Notas lo
  excluye para todo el hito («la única supresión admitida… es una entrada en `ignore-rules`»). Es una
  supresión de hallazgo; la entrada en `ignore-rules` es configuración de un diccionario (research D9, S2).
- **Partir la cadena literal** o sacar la tabla a un fichero: engaña al linter sin arreglar nada.
- **Editar `.golangci.yml` desde aquí**: el guardián de diff de la reparación lo rechazaría (rutas de
  `gates/tarea-actual.json`: `nombre.go`, `nombre_test.go`, `plan.md`), y en `--resume` el motor
  reejecutaría ese mismo guardián por índice con las mismas rutas, dejando el run sin salida limpia.
- **Ampliar a mano las rutas de `gates/tarea-actual.json`** para que el guardián lo dejara pasar: el
  agente se ampliaría sus propias rutas en el mismo paso en que las usa. Eso es relajar el gate.

## Decisión tomada (autorización permanente de Jorge: arreglar la raíz, no esperar)

1. **`tasks.md`, línea de T010**: declara ahora `.golangci.yml`, «acotado a **dos palabras nuevas bajo
   `misspell.ignore-rules`** —`legislacion` y `resolucion`…—, ninguna regla, ningún linter, ninguna
   exclusión y nunca una supresión», con la misma redacción que T003. El extractor de rutas del workflow
   la lee bien (comprobado con su propio filtro: `.golangci.yml`, `internal/httpx/nombre.go`,
   `internal/httpx/nombre_test.go`, `plan.md`). Es la opción 1 que recomendaba la primera versión de esta
   nota: la contingencia pertenece a la tarea que la dispara, que es la regla que el plan ya aplicó con
   T003 (obligación 3).
2. **`tasks.md` §Notas**: la nota de «nada de atajos» dice ya que la supresión la añaden T003 y T010, por
   qué el supuesto del plan (el español nuevo se agotaba en los tipos) se quedó corto, y qué hace una tarea
   futura en el mismo caso: anotar, redelimitar con la misma acotación y dejar sin marcar.
3. **`research.md` S2**: párrafo «Resultado (implementación)» con lo anterior y el aviso para H4:
   `legislacion-consolidada` es el nombre real del servicio del BOE y volverá a aparecer en nombres de
   fixtures citados desde tests; la entrada es durable.

No se ha tocado `plan.md` (la obligación 3 ya habla de «la tarea que lo comprueba», sin nombrar a T003), ni
`.golangci.yml`, ni ningún fichero fuera del directorio del feature.

## Qué falta, exactamente, y cómo reanudar

`--resume` reejecuta el paso fallido **por índice** (`docs/WORKFLOW.md`, «Cada run congela el workflow»);
el paso fallido será `verificar_reparacion`, que vuelve a lanzar `make ci`. Por tanto **el árbol tiene que
estar en verde antes de reanudar**, y el guardián de este intento ya pasó: la legitimidad del cambio en
`.golangci.yml` la da la línea redelimitada de T010, no el guardián. Pasos:

1. Añadir estas dos entradas a `misspell.ignore-rules` de `.golangci.yml`, a continuación de
   `- inventario` (línea 203), con el mismo sangrado y el mismo tipo de comentario que las cuatro que ya
   están. **Ninguna otra línea** (escenario 12 del quickstart: el diff frente a `main` admite esas
   palabras y nada más):

   ```yaml
           # «Legislacion consolidada», el servicio del BOE cuyo nombre imita la ruta
           # larga de la tabla del contrato de grabación §2 de H2 que
           # TestNombreDeGrabacion copia tal cual; misspell lo lee como «legislation».
           - legislacion
           # «Resolucion de adjudicacion», parte de la otra ruta larga de la misma
           # tabla, la que se recorta; misspell lo lee como «resolution».
           - resolucion
   ```

   **Comprobado** en este paso sobre una copia de `.golangci.yml` con exactamente ese bloque, sin tocar
   el repositorio: `golangci-lint run --config <copia> --default none --enable misspell ./internal/httpx/...`
   → `0 issues`; con la configuración del repositorio, los mismos 2 hallazgos.
2. `make ci` en verde.
3. Marcar T010 `[X]` en `tasks.md` (la línea ya redelimitada).
4. `scripts/hito.sh --resume bfa8c3ac`. `verificar_reparacion` pasa y `commit_tarea` hace `git add -A`
   y commitea todo —`nombre.go`, `nombre_test.go`, `.golangci.yml`, `tasks.md`, `research.md` y este
   fichero— como `feat(H2): T010`, coherente con las rutas que la tarea declara ahora.

## Estado del árbol al terminar este paso

`internal/httpx/nombre.go` y `internal/httpx/nombre_test.go` sin cambios respecto al intento 1 (correctos y
completos); `tasks.md` (línea de T010 y §Notas), `research.md` (S2) y este fichero modificados, todos
dentro del directorio del feature. T010 sigue `[ ]`. `make ci` sigue en rojo únicamente por las dos líneas
de `misspell` de arriba.

## Posdata (cierre, tras reanudar)

Los pasos 1 a 3 de «Qué falta» se aplicaron tal cual y `make ci` quedó en verde. El paso 4 no fue como se
esperaba: `--resume` **no** reejecuta el paso anidado por índice, sino el paso de nivel superior que lo
contiene (`engine.py`: «resume will re-run the parent step and its nested body»), así que el bucle arrancó
una iteración nueva, `siguiente_tarea` eligió T011 y el guardián de T011 encontró `nombre.go`,
`nombre_test.go` y `.golangci.yml` fuera de sus rutas. Salida: T010 se commiteó a mano con exactamente
sus ficheros (`a27311a`, mismo mensaje que `commit_tarea`), y la causa se arregló en el workflow 1.6.1:
`siguiente_tarea` detecta una tarea marcada `[X]` con trabajo sin commitear y la cierra —guardián con su
base original, verificación y commit— antes de elegir otra (`docs/WORKFLOW.md`, «Limitaciones
conocidas»).
