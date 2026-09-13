# T008 — intento 1: sin marcar; `make ci` en rojo por `misspell` sobre una palabra que fijan dos contratos

**Estado**: el adaptador de prueba y sus dos tests están escritos y en verde (doce subpruebas). `make ci`
**no** lo está: `lint` falla por dos hallazgos de `misspell` sobre `reproduccion`, y cerrarlo exige una
entrada en `.golangci.yml`, que la tarea no declaraba. Es la contingencia S2 en la variante que
`tasks.md` §Notas deja prevista: el término lo fija un contrato y no se puede reescribir. Salida
aplicada: anotar aquí, **redelimitar la línea de T008** para que declare `.golangci.yml` acotado a esa
palabra, dejar rastro en `research.md` S2 y dejar la tarea `[ ]`. Al intento siguiente le queda una
edición de cinco líneas en `.golangci.yml`, descrita abajo al detalle y **comprobada** sobre una copia
desechable.

## Lo que sí quedó hecho (y verde)

- `internal/cache/adaptador_test.go`, en `package cache_test`: los tipos `adaptadorDePrueba` (applet
  `prueba`, con `directorio`, `reloj`, el cliente de reproducción que da el test y un contador
  `peticiones`), `argumentosConsultar`, `argumentosGuardar` y `cuerpoDeLaFuente` (`cuerpo`, `origen`),
  con el patrón del contrato del puerto §8. Lee `schema.Contexto.Offline` y lo pasa como
  `cache.SoloLectura()`, mira la caché antes de pedir, pide por `httpx.Replay` con el literal `"GET"` (sin
  `net/http`) y guarda con una hora de vigencia. Bajo `--dry-run` no guarda la respuesta de ensayo, que no
  trae cuerpo. `guardar` cita el espacio de nombres reservado (`kitlegal.prueba`,
  `kitlegal:applet/prueba/guardar`), porque no consulta ninguna fuente.
- `TestAdaptadorDePruebaConElKernel`: `app.Main` en proceso sobre un registro construido en el test, con
  **nueve** `t.Run` de un solo nivel, cada uno con su propio `t.TempDir()`, y **seis y solo seis** con el
  prefijo `offline-`. Nombres y esperados son los del inventario del plan. La siembra pasa por el propio
  kernel con el reloj fijo en `2026-09-12T00:00:00Z`, y la lectura en el borde usa el último nanosegundo
  vigente o el instante exacto de expiración, sin esperar. «Cero peticiones» se mide dos veces: con el
  contador del adaptador y con la reproducción estricta sobre un directorio vacío. El SHA-256 de
  `cache.db` (o su ausencia) y el listado del directorio, sin `-wal` ni `-shm`, se comparan antes y
  después en los seis `offline-*`.
- `TestAdaptadorConVariableDeEntorno` (sin `t.Parallel`, con `t.Setenv` y `HOME`/`USERPROFILE`
  redirigidos) tiene tres subpruebas: la base aparece bajo `KITLEGAL_CACHE_DIR` y nada bajo `HOME`; la
  variable vacía → 2 nombrándola; la variable que apunta a un fichero → 2 nombrando variable y ruta, con
  el fichero intacto.
- `rtk proxy go test -race -count=1 -v -run '^TestAdaptador' ./internal/cache/` → doce `--- PASS` y `ok`.
- `rtk proxy make -k ci`: `fmt-check`, `test` (`-race -shuffle=on`), `vuln`, `schema-check`, `secrets`,
  `mod-verify` y `mod-tidy-check` en verde. **Solo `lint` en rojo**, con los dos hallazgos de abajo.
- **Rojo comprobado por mutación**, cada una sobre su clon desechable en el scratchpad:
  - adaptador que ignora `--offline` (`if ejecucion.Offline && ejecucion.DryRun`): fallan
    `offline-ausente`, `offline-expirada`, `offline-sin-base`, `offline-directorio-inexistente` y
    `offline-guardar`. `offline-presente` sigue pasando, como corresponde: mide que se sirve, no el modo.
  - adaptador que no sirve de la caché (`if presente && a.URL == ""`): fallan `segunda-consulta-sin-red`
    y `offline-presente`.
  - `sin-cache-la-reproduccion-falla` es el control negativo en el propio árbol: código 1 y el mensaje
    nombra `GET http://fuente.prueba/norma`.
- Direcciones: el test y la grabación solo nombran `http://fuente.prueba/norma`,
  `http://fuente.prueba/otra` y la identificación `https://ventanillalegal.es/bot`.
- No se tocó `testdata/`, ni `schemas/`, ni la red, ni `KITLEGAL_RECORD`, ni ningún fichero fuera de las
  rutas congeladas y del directorio del feature.

## Lo que deja `make ci` en rojo

```
internal/cache/adaptador_test.go:365:22: `reproduccion` is a misspelling of `reproduction` (misspell)
	t.Run("sin-cache-la-reproduccion-falla", func(t *testing.T) {
internal/cache/adaptador_test.go:550:58: `reproduccion` is a misspelling of `reproduction` (misspell)
	cliente, err := httpx.Replay(filepath.Join("testdata", "reproduccion", fuenteDePrueba), httpx.ConFuente(fuenteDePrueba))
```

`misspell` v0.8.0 separa las palabras con `[a-zA-Z0-9']+` (`replace.go:19`), así que el guion del nombre
del subtest no protege la palabra, y `reproduccion` está en su diccionario (`words.go`).

## Por qué no se cierra en este intento

Se descartaron, por orden, todas las salidas al alcance del intento:

- **Reescribir el término**: el nombre `sin-cache-la-reproduccion-falla` lo fijan el inventario de tests
  de `plan.md` («esta lista es la única»), la expresión y el esperado del escenario 2 de `quickstart.md`,
  research D12 y la propia línea de T008. El tramo del directorio es el que el contrato de grabación de H2
  da a la reproducción de cada fuente. Cambiar dos contratos para contentar a un diccionario inglés deja
  los artefactos peor (el mismo criterio que H2 T010).
- **Localizar el directorio con `filepath.Glob` sin escribir el tramo**: quitaría uno de los dos
  hallazgos, pero no el del nombre del subtest, que es obligatorio.
- **Partir la cadena literal**: engaña al linter sin arreglar nada.
- **`//nolint:misspell`**: el hito no admite ninguna supresión (SC-008, `tasks.md` «Ninguna supresión en
  todo el hito»).
- **Editar `.golangci.yml` desde aquí**: no está en las rutas congeladas de `gates/tarea-actual.json`
  (`internal/cache/adaptador_test.go`, más dos fichas sin fichero, `.prueba/norma` y `net/http`), y el
  guardián de diff lo rechazaría. Ampliar a mano esas rutas sería relajar el gate.

## Decisión tomada

1. **`tasks.md`, línea de T008**: declara ahora `.golangci.yml` **acotado a una palabra nueva en la lista
   `ignore-rules` de `misspell`** (`reproduccion`), con su comentario y ninguna otra línea: ninguna regla,
   ningún linter, ninguna exclusión y nunca una supresión. El filtro del propio workflow sobre la línea
   extrae `.golangci.yml`, `internal/cache/adaptador_test.go` y las dos fichas sin fichero que ya
   extraía (`.prueba/norma`, `net/http`).
2. **`tasks.md` §Notas**: la lista de ficheros de H0/H1/H2 que se tocan añade «T008 solo
   `misspell.ignore-rules`, la palabra `reproduccion`», y la nota de la contingencia dice que T008 es ese
   caso.
3. **`research.md` S2**: párrafo «Resultado (implementación, T008 intento 1)», con el aviso de que la
   entrada es durable, porque cada adaptador de fuente volverá a escribir ese directorio en sus tests.

## Qué falta, exactamente (intento siguiente)

1. En `.golangci.yml`, dentro de `misspell.ignore-rules`, entre `- legislacion` y el comentario de
   `resolucion` (orden alfabético), añadir **exactamente** este bloque y ninguna otra línea:

   ```yaml
           # «Reproduccion», el tramo del directorio de grabaciones que fija el
           # contrato de grabación de H2 (testdata/reproduccion/<fuente>) y que nombra
           # el subtest sin-cache-la-reproduccion-falla del inventario de tests de H3;
           # misspell lo lee como «reproduction».
           - reproduccion
   ```

   **Comprobado** en este intento sobre un clon desechable del repositorio (`git clone` en el scratchpad,
   con este `adaptador_test.go` copiado y ese bloque aplicado): `make -C <clon> lint` → `0 issues.`
2. `make ci` en verde.
3. Marcar T008 `[X]` en `tasks.md`.

## Paso `reparar` del intento 1 (verificación en rojo): comprobado de nuevo, sin cambios en código

El paso `reparar` del workflow recibió los dos hallazgos de arriba con las rutas congeladas de
`gates/tarea-actual.json` (`internal/cache/adaptador_test.go` y las dos fichas sin fichero). Volvió a
comprobarlo todo por su cuenta, sin fiarse de este fichero:

- `go test -race -count=1 ./internal/cache/` → `ok`.
- El extractor de rutas del propio workflow (`tr '`,;()' '     ' | grep -oE …` sobre la línea
  redelimitada de T008 en `tasks.md`) devuelve exactamente `.golangci.yml`, `.prueba/norma`,
  `internal/cache/adaptador_test.go` y `net/http`: el intento 2 tendrá `.golangci.yml` entre sus rutas y
  ninguna otra de más.
- Sobre un clon desechable de HEAD (`a2cb61e`) con este `adaptador_test.go` copiado y el bloque de cinco
  líneas de la sección anterior aplicado en `.golangci.yml` (`git diff --stat`: `1 file changed,
  5 insertions(+)`, ningún otro fichero): `make lint` → `0 issues.` y **`make ci` completo → `ci: todos
  los controles en verde`** (exit 0).

No tocó ningún fichero del árbol fuera del directorio del feature, y a propósito:

- `.golangci.yml` no está en las rutas congeladas: `guardian_diff_reparacion` lo rechazaría, y ampliar
  las rutas a mano es relajar el gate.
- Dentro de `adaptador_test.go` no hay arreglo de raíz: el nombre del subtest lo fijan el inventario de
  tests del plan, el escenario 2 del quickstart y D12 (cambiar artefactos aprobados para contentar a un
  diccionario inglés los deja peor), y partir el literal o derivar el nombre de otra cadena engaña al
  linter sin arreglar nada. La entrada en `ignore-rules` es el mecanismo que el proyecto ya usa para el
  mismo caso (`legislacion` y `resolucion`: tramos de ruta que fija el contrato de grabación de H2).
- No es un test en rojo por implementación de una tarea posterior, pero sí una tarea mal delimitada, que
  es la salida que el paso prevé: anotar, dejar `[ ]` y terminar.

Consecuencia en el run: `verificar_reparacion` queda en rojo con los mismos dos hallazgos y `hito.sh` lo
clasifica como parada `deliberada` (docs/WORKFLOW.md, tabla de clasificación de paradas). Reanudar con
`scripts/hito.sh --resume <run_id>`: `siguiente_tarea` relee `tasks.md`, toma T008 como intento 2 con
`.golangci.yml` entre las rutas, y al ejecutor le queda la edición de cinco líneas de la sección
«Qué falta, exactamente», ya comprobada dos veces.

## Estado del árbol al terminar el intento 1 (tras `reparar`)

`internal/cache/adaptador_test.go` nuevo (sin seguimiento), completo y en verde en test.
`specs/004-h3-internal-cache-sqlite/tasks.md` (línea de T008 y §Notas),
`specs/004-h3-internal-cache-sqlite/research.md` (S2) y este fichero, modificados, todos dentro del
directorio del feature. `.golangci.yml` sin tocar. T008 sigue `[ ]`. `make ci` sigue en rojo únicamente
por los dos hallazgos de `misspell` de arriba, hasta el intento 2.

## Intento 2: cerrado en verde

Con `.golangci.yml` ya entre las rutas congeladas de `gates/tarea-actual.json`, el intento hizo
exactamente lo que dejaba previsto la sección «Qué falta, exactamente», y nada más:

1. **Rojo comprobado de nuevo antes de tocar nada**: `make lint` → los mismos dos hallazgos de `misspell`
   sobre `reproduccion` (líneas 365 y 550 de `adaptador_test.go`) y ningún otro.
2. **`.golangci.yml`**: el bloque de cinco líneas de arriba, entre `- legislacion` y el comentario de
   `resolucion`. `git diff --stat`: `1 file changed, 5 insertions(+)`; ninguna regla, ningún linter,
   ninguna exclusión, ninguna supresión.
3. **`make ci` completo → `ci: todos los controles en verde`** (exit 0): `fmt-check`, `lint`
   (`0 issues`), `test` (`-race -shuffle=on`, `internal/cache` en `ok` con 87,6 % de cobertura), `vuln`,
   `schema-check`, `secrets`, `mod-verify` y `mod-tidy-check`.
4. **Sonda positiva** (`go test -race -count=1 -v -run '^TestAdaptador' ./internal/cache/`): doce
   `--- PASS` de subprueba, nueve bajo `TestAdaptadorDePruebaConElKernel` y tres bajo
   `TestAdaptadorConVariableDeEntorno`. Recuento sobre el fichero: nueve `t.Run` de primer nivel en el
   test del kernel, seis y solo seis con el prefijo `offline-`, cero `t.Run` anidados, ninguna
   importación de la biblioteca HTTP.
5. Fuera del directorio del feature solo cambian los dos ficheros declarados: `.golangci.yml` y
   `internal/cache/adaptador_test.go`. Ni `testdata/`, ni `schemas/`, ni red, ni `KITLEGAL_RECORD`.

T008 queda `[X]` en `tasks.md`.
