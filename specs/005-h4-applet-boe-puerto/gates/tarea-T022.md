# T022 · `analisis`

## Intento 1 (2026-09-13)

**Sin marcar: redelimitada** (decidible por el ejecutor). El verbo está hecho, con su test, y `make ci` queda en verde;
pero la solución completa toca dos ficheros que la línea no declaraba (ver «Por qué no se marca»).

### Lo que deja escrito (sin commitear: el workflow no commitea una tarea redelimitada)

- `internal/source/boe/analisis_test.go` (nuevo), escrito antes que el código y visto en rojo por la razón esperada
  (`la fuente del BOE no resuelve consultas de tipo boe.ConsultaAnalisis`). `TestAnalisis`, 27 subtests:
  - `lpac`: la grabación de la Ley 39/2015, comparada con el `data` compuesto sin el código de lectura (estructura fija
    con la forma de la API) y fijada a mano (1 materia, 2 notas —una con su doble espacio—, 11 anteriores y 18
    posteriores en varias posiciones); servida sin pedir hasta el último nanosegundo de los siete días y vuelta a pedir al
    cumplirse, con la fecha de la nueva petición.
  - `texto-completo`: un texto de referencia de más de 200 caracteres y con letras de varios octetos llega completo, al
    pedirlo y al servirlo con `--offline` (FR-060).
  - `en-listas` y `objetos-sueltos` dan el mismo `data` (US5, escenario 2; FR-070); `sin-envoltorios`; `vacios` (listas
    vacías y nunca nulas, también en JSON).
  - Fallos, y en cada uno nada escrito: `inexistente` (404 → 3), `estado-403`, `ilegible`,
    `primer-elemento-que-no-es-objeto`, un campo que no es texto en cada lectura (`codigo`, `materia`, `nota`,
    `relacion.texto`, `texto`, `id_norma`) y `data` vacío en cinco formas (→ 3, FR-061).
  - `ensayo` (la línea de la petición, sin crear la caché, y la consulta siguiente pide de verdad),
    `offline-sin-entrada`, `offline-y-ensayo-sin-entrada` y `norma-invalida` en los tres modos.
- `internal/source/boe/analisis.go` (nuevo): `analisis`; `analisisDeLaNorma` (`consultar` con la clave, el pedido,
  604 800 s y `leerAnalisis`); `leerAnalisis` (J5; los envoltorios de `lectura.go`; orden materias → notas → anteriores →
  posteriores; el fallo nombra el primer campo que no es texto); `leerCadaUno`; `leerMateria`; `leerNota`;
  `leerReferenciaAnterior` (texto completo) y `leerReferenciaPosterior`.
- `internal/source/boe/fuente.go`: el caso `ConsultaAnalisis` en `Fetch`, su comentario sin «un verbo que la fuente
  todavía no resuelve», y `resolverRecursoDeLaNorma`, que por ahora solo usa `analisis`.

### Decisiones de lectura (data-model §2.6 y §3.1)

- Una materia que no es objeto aporta solo su texto con las reglas de J9: una cadena es su valor y una materia nula, la
  de texto vacío (`refs/boe.py` 535 también presenta ese elemento). Lo que no es texto ni objeto es ilegible, y el
  mensaje lo nombra con el nombre de su envoltorio: `materia` o `nota`.
- De las referencias posteriores no se lee `texto`: `data-model.md` no lo lleva y `refs/boe.py` 577-581 no lo presenta,
  así que un `texto` raro en una posterior no hace fallar la consulta.

### Por qué no se marca

- **El hallazgo.** Con `analisis` escrito como sus hermanos, `dupl` (umbral 100) marca el verbo entero: primero contra
  `indice.go` 25-41 (validar la norma, `invocar`, `consultar` con su lector y el resultado); reescrito con la forma de
  `metadatos.go`, con un `analisisDeLaNorma` aparte, contra `metadatos.go` 33-48. No es una casualidad de tokens:
  `metadatos`, `indice` y `analisis` son el mismo esqueleto de verbo —validar la norma antes de abrir nada, `invocar`,
  resolver con la caché de la invocación y componer el resultado con la dirección del recurso— y `dupl` lo marca en
  cuanto existe el tercero.
- **Lo que no vale.** Retorcer `analisis` para que el detector no lo vea (un lector currificado, otro orden) deja el
  esqueleto triplicado y un verbo escrito distinto de sus hermanos sin motivo. Y cerrar con el árbol tal como queda
  —`resolverRecursoDeLaNorma` usado solo por `analisis`, en verde porque la copia del esqueleto en el helper no llega al
  umbral frente a `metadatos` e `indice`— deja dos formas de hacer lo mismo en el paquete y un comentario que nombra
  tres verbos cuando lo usa uno: un arreglo diferido que ninguna tarea posterior recoge (T023 y T024 no añaden
  producto).
- **La solución.** `resolverRecursoDeLaNorma[T]` en `fuente.go` recoge el esqueleto, y `metadatos`, `indice` y
  `analisis` quedan en una línea con su dirección y su función de resolución: `metadatosDeLaNorma`, que ya existe;
  `indiceDeLaNorma`, nueva, con el `consultar` que hoy está dentro de `indice`; y `analisisDeLaNorma`. No cambia ningún
  comportamiento ni ningún test de `metadatos` ni de `indice`. Necesita `metadatos.go` e `indice.go`, que la línea no
  declaraba y que el guardián de este intento rechazaría.

### Redelimitación

- La línea de T022 declara ahora también `internal/source/boe/metadatos.go` e `internal/source/boe/indice.go` y describe
  el helper; la tabla de la rúbrica (criterio d) lo recoge.
- **Rutas extraídas** con la expresión del workflow (`workflow.yml` 615): `gates/tarea-T022.md`,
  `internal/source/boe/analisis.go`, `internal/source/boe/fuente.go`, `internal/source/boe/indice.go` e
  `internal/source/boe/metadatos.go`.
- **Cómo se comprobó**, sobre una copia desechable del árbol (`rsync` a `/tmp`, borrada al terminar) con `metadatos.go`,
  `indice.go`, `analisis.go` y `fuente.go` ya migrados: `golangci-lint run ./internal/source/boe/...` con 0 hallazgos y
  `go test -race -count=1 ./internal/source/boe/` en verde.

### Verificación de este intento

- `TestAnalisis`: 27 de 27 en verde con `-race`.
- `make -k ci` en primer plano sobre el árbol que queda: «ci: todos los controles en verde» (`golangci-lint run ./...`
  con 0 hallazgos, tests con `-race` e integración, `govulncheck`, `gitleaks`, módulos y `tidy -diff`).
- Por el camino, `misspell` marcó `directos` (nombre de subtest, ahora `sin-envoltorios`) y `constitucional` (en un
  texto de prueba, sustituido por uno compuesto con tramos de las referencias grabadas). `dupl` marcó además dos
  subtests de `analisis_test.go` calcados de `indice_test.go` (`lpac` y `ensayo`). Se reescribieron sobre lo propio del
  verbo: la frontera de los siete días en `lpac` y, en `ensayo`, que la consulta siguiente pide de verdad. La matriz de
  modos de la caché queda para `TestCacheDeLosSeisVerbos`, `TestOfflineDeLosSeisVerbos` y `TestEnsayoDeLosSeisVerbos`
  (T023).

### Para el intento siguiente

- Aplicar en `metadatos.go` e `indice.go` lo descrito en «La solución»: los dos verbos en una línea sobre
  `resolverRecursoDeLaNorma`, e `indiceDeLaNorma` con el `consultar` actual, su vigencia y `leerIndice`, cada uno con su
  comentario.
- `golangci-lint` del paquete sin hallazgos; `TestMetadatos`, `TestIndice` y `TestAnalisis`, sin cambios, en verde; y
  `make ci` en primer plano. Marcar `[X]` en el mismo turno.

## Intento 2 (2026-09-13)

**Marcada `[X]`.** Sobre lo que dejó el intento 1 (sin cambios en `analisis.go`, `analisis_test.go` ni `fuente.go`), aplica
en `metadatos.go` e `indice.go` lo descrito en «La solución»:

- `metadatos` queda en una línea sobre `resolverRecursoDeLaNorma` con `direccionDeLosMetadatos` y el `metadatosDeLaNorma`
  que ya tenía; su comentario deja de repetir lo que ahora cuenta el helper.
- `indice` queda en una línea sobre `resolverRecursoDeLaNorma` con `direccionDelIndice` e `indiceDeLaNorma`, nueva, con el
  `consultar` que antes estaba dentro de `indice` (`claveDelIndice`, `pedidoDelIndice`, `vigenciaLarga` y `leerIndice`),
  con la misma forma que `metadatosDeLaNorma` y `analisisDeLaNorma`.
- Ningún test cambia: `metadatos_test.go` e `indice_test.go` quedan intactos y son la red de seguridad del cambio
  estructural.

### Verificación de este intento

- `golangci-lint run ./internal/source/boe/...` con el binario fijado por el repo: 0 hallazgos (`dupl` ya no marca
  ningún verbo).
- `go test -race -count=1 -run '^(TestMetadatos|TestIndice|TestAnalisis)$'`: los tres en verde; `TestAnalisis`, 27 de 27.
- `make ci` en primer plano: «ci: todos los controles en verde» (lint, tests con `-race` en los dos perfiles,
  `govulncheck`, `gitleaks`, módulos y `tidy -diff`).
- Ficheros tocados en los dos intentos, todos dentro de las rutas declaradas: `internal/source/boe/analisis.go`,
  `internal/source/boe/analisis_test.go`, `internal/source/boe/fuente.go`, `internal/source/boe/indice.go`,
  `internal/source/boe/metadatos.go` y el directorio del feature.
