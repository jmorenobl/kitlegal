# T003, intentos 1, 2 y 3 (2026-10-18): la tarea está completa y `make ci` en verde en un árbol limpio; en rojo en el árbol de trabajo por el residuo del arnés que paró T001 y T002. El intento 3 la lleva a cuarentena: la tercera seguida, DAG bloqueado (receta para la persona al final)

## Intento 2 (2026-10-18): el mismo muro, medido de nuevo

- La orden `ls -d` de abajo encontró los cinco directorios. En el árbol hay 40 directorios `.claude/.cc-writes`
  vacíos fuera de la raíz (`find . -path ./.git -prune -o -type d -name .claude -print`), los mismos que en T002 más
  los que dejó el intento 1 en `specs/023-h9-internal-core-plazos/`, que ningún test lee.
- `go vet ./internal/core/plazos/` y `go test -race -count=1 -cover ./internal/core/plazos/`: en verde, 39 subtests.
  Las importaciones del paquete y de sus tests son `internal/core/schema`, `go.yaml.in/yaml/v3`, `testify` y la
  biblioteca estándar; ni `os` ni `io`. Ningún fichero de la tarea se tocó.
- `make ci` no se repitió: la causa es la misma, está medida abajo y los tests que caen no leen nada de esta tarea.
- Los residuos no se retiraron, por lo mismo que en el intento 1 y en `gates/tarea-T002.md` (intento 2): borrar una
  ruta `.claude/` pide una aprobación que en modo desatendido nadie da, y hacerlo con otra orden sería sortear ese
  rechazo. Es la única opción que deja la tarea en verde y no es decidible por la sesión: queda para una persona.
- Esta nota cambia en este intento para que el bucle no gaste una sesión de `reparar` (la tarea queda «redelimitada»
  sin cambiar su línea). **El intento 3 pondrá T003 en cuarentena y, con T001 y T002, serán tres seguidas: el run se
  parará con el DAG bloqueado.** Para evitarlo hay que parar el run antes del intento 3 y hacer lo que dice la
  sección final; si no, el dossier del DAG bloqueado lleva a estas mismas notas.

## Para el intento siguiente, antes de leer nada más

```
ls -d internal/source/boe/testdata/boe.legislacion-consolidada/.claude testdata/evals/boe.legislacion-consolidada/.claude internal/app/testdata/cita/.claude internal/app/testdata/script/.claude internal/skills/testdata/script/.claude
```

- **Si alguno existe**: no repitas ninguna verificación (abajo está todo, medido en este intento). Deja la tarea `[ ]`,
  añade a esta nota la fecha y el número del intento y termina. Esta sesión no puede retirarlos: borrar una ruta
  `.claude/` pide una aprobación que en modo desatendido nadie da (`gates/tarea-T002.md`, intento 2), y borrarlos con
  otra orden sería sortear ese rechazo.
- **Si no existe ninguno**: `make ci` en primer plano desde la raíz y, en verde, marca T003 `[X]`. No toques ningún
  fichero de la tarea: cumplen la tarea tal como están.
- **Nunca `cd` a un subdirectorio en una orden de Bash**: rutas absolutas desde la raíz. Es lo que deja el residuo.

## Lo que está hecho y no hay que tocar

Paquete nuevo `internal/core/plazos`, siete ficheros y nada más fuera del directorio del feature (`git status`):

- `doc.go`: el comentario del paquete.
- `tipos.go`: `Tipo`, `Duracion`, `Precepto`, los vocabularios `CodigosDeTipo`, `Computos` y `Unidades`, y
  `LeerTipos(bytes)` con `KnownFields`: los siete códigos y ninguno más, cada uno una vez; `duracion` en los cinco que
  la fijan y en ningún otro; `unidades` solo en `generico`, con las tres; `fija: duracion` en un precepto de cada tipo
  con duración; y lo que dan por hecho data-model §1 y la salida (supuesto anotado en `supuestos.md`). Los defectos van
  juntos, cada uno con su tipo, en un error sin clase de usuario (`defectoDeDatos`: el kernel lo da como `inesperado`).
- `calendario.go`: `Calendario`, inmutable tras cargar; `CargarCalendario(nacionales, locales)` lee las dos formas de
  contracts §1.1 y §1.2, sin reglas de carga propias (D25): lo que no se lee —vacío, YAML roto, clave no declarada,
  fecha que no es `AAAA-MM-DD`, ámbito o motivo fuera de su vocabulario— es un error que nombra el fichero. Consultas:
  `Inhabil`, `FestivosDe`, `PartesDe`, `Cobertura` (por ámbito y tramos: un tramo por año cargado, con `disposicion` o
  `falta`; uno por hueco de años seguidos con la misma falta), y `AlcanceNacional`/`AlcanceLocal` (supuesto anotado:
  los textos de los avisos de T004 lo necesitan y T004 no declara `calendario.go`). Vocabularios: `Ambitos`,
  `AmbitosDeCobertura`, `FaltasDeCobertura` y `MotivosDeFalta` (con `sin-municipio`, el de una línea). `Municipio` es
  `{CodigoINE, Nombre, Comunidad}`, el municipio ya resuelto que recibirá la `Peticion` de T004.
- `festivos.go`: `FestivosNacionales`, `FestivosLocales` y sus filas, `Cabecera`, y `EscribirFestivosNacionales` y
  `EscribirFestivosLocales`: línea de fichero generado, claves en el orden del contrato, una fila por línea en flujo y
  las listas ordenadas (orden elegido, anotado en `supuestos.md`); no cambian lo que reciben.
- Tests: `TestLeerTipos` (el fichero de contracts §2 como constante y 23 variantes, una por regla),
  `TestCargarCalendario` (dos comunidades sintéticas, una con fichero local, festivos con y sin `parte`, una falta de
  cada motivo, dos años; cada consulta, en subtests paralelos sobre el mismo calendario) y `TestEscribirFestivos`
  (forma exacta, orden de filas y claves, ida y vuelta byte a byte, lo escrito se carga). Ni `os` ni `io`.

## Lo medido en este intento, de primera mano

- `go test -race ./internal/core/plazos/`: en verde; 36 subtests; cobertura 98,5 %.
- Rojo antes del código: los tests no compilaban (`undefined: Municipio`, `Calendario`…).
- Verificación de la tarea: con la comprobación de los siete códigos quitada (el caso «no es uno de los siete» y el
  bucle «falta el tipo»), `TestLeerTipos` en rojo en `falta_un_código` y `un_código_de_más`; restaurada, en verde.
- `golangci-lint run ./internal/core/plazos/...`: `0 issues.` (tras pasar a `assert.NotErrorAs` y aplicar
  `golangci-lint fmt`).
- **`make ci` en un árbol limpio** (`git worktree add --detach` de `8209889` bajo `$TMPDIR`, con solo los siete ficheros
  copiados, ningún `.claude/`): **`ci: todos los controles en verde`** (`0 issues.`, `test` e `integration` en verde,
  `no leaks found`). El árbol se retiró con `git worktree remove --force` y `git worktree prune`.
- **`make ci` en el árbol de trabajo: rojo**, solo en `test` (`fmt-check` y `lint` en verde), en `internal/app`
  (`TestNingunVerboDelGrafoDevuelveTexto`, `TestLaSalidaDeBoeNoCambiaConElGrafo`: `read .claude: is a directory`,
  `grafo_test.go:1842`), `internal/evals` (`TestReconstruccionDeJurisprudencia`, `TestEjecucionDeLaMedida` y los demás
  que preparan una caché: `el conjunto de grabaciones ../../internal/source/boe/testdata/boe.legislacion-consolidada no
  se puede copiar: .claude no es un fichero regular`) e `internal/skills`. Ninguno lee nada de esta tarea;
  `internal/core/plazos`, en verde también ahí.
- Esta sesión no añadió residuos donde los leen los tests: el único `.cc-writes` más nuevo que
  `gates/cuarentena/T002.patch` es `specs/023-h9-internal-core-plazos/.claude/.cc-writes`, que ningún test lee.

## La causa y lo que hace falta

La de `gates/tarea-T001.md` y `gates/tarea-T002.md`: directorios vacíos `<directorio>/.claude/.cc-writes/` que deja el
arnés de Claude Code en el directorio de trabajo de una orden de Bash con `cd`, que git ignora y que los tests que
listan o copian un directorio de `testdata/` encuentran y, con razón, rechazan. Es ajena a esta tarea y la sesión no la
puede retirar sin una aprobación que en este modo nadie da.

Solo lo puede hacer una persona, lo mismo que piden esas dos notas: parar el run antes de que T003 entre en cuarentena
(sería la tercera seguida y el run se pararía con el DAG bloqueado), retirar los residuos vacíos sin tocar nada
versionado (las dos órdenes `find … -empty -print -delete` de `gates/tarea-T002.md`), reanudar con
`scripts/hito.sh --resume <run_id>`, y el arreglo de raíz fuera de la rama del hito (que el workflow retire esos
directorios antes de medir y que los pasos con modelo no hagan `cd` a subdirectorios).

## Intento 3 (2026-10-18T07:45Z): el mismo muro; la tarea sigue completa y pasa a cuarentena

Los cinco directorios de la orden `ls -d` de arriba existen; siguen los 40 residuos (`find . -path ./.git -prune -o
-type d -name .claude -print`, sin contar `./.claude`), todos con un solo `.cc-writes` vacío dentro y ningún fichero
regular. Esta sesión no añadió ninguno: todas sus órdenes corrieron desde la raíz. No se tocó ningún fichero de la
tarea ni `tasks.md`; `git status` sigue igual que al empezar, más esta nota.

Medido en este intento, de primera mano:

- `go vet ./internal/core/plazos/` y `go test -race -count=1 ./internal/core/plazos/`: en verde, 39 subtests.
- `make ci` en primer plano, desde la raíz: `fmt-check` y `lint` en verde (make llegó a `test`, el tercer objetivo de
  `ci:`); **`test` en rojo solo por el residuo `internal/source/boe/testdata/boe.legislacion-consolidada/.claude`**:
  `internal/app` (`TestNingunVerboDelGrafoDevuelveTexto` y `TestLaSalidaDeBoeNoCambiaConElGrafo`, `grafo_test.go:1842`,
  repetidos aparte con `-run`: los dos en rojo), `internal/evals` (`TestInterrumpirLaSesion` y los demás que preparan
  una caché: `el conjunto de grabaciones … no se puede copiar: .claude no es un fichero regular`) e `internal/skills`
  (`TestEsquemaDeNormas/rangos-grabados`: `read …/.claude: is a directory`). Los demás paquetes, en verde; ninguno de
  esos tests lee nada de esta tarea. Los objetivos posteriores a `test` no llegaron a correr.
- Los residuos no se retiraron: `rmdir` sobre una ruta `.claude/` pide una aprobación («sensitive file», probado en el
  intento 2 de T002) que en este modo nadie da, y hacerlo con otra orden sería sortear ese rechazo.

Lo que hace el bucle a partir de aquí (`scripts/workflow/tarea.sh`): esta nota cambia en el intento, así que
`redelimitada` la da por redelimitada y no gasta una sesión de `reparar`; al elegir la tarea siguiente, el contador de
T003 pasa a 4, mayor que `max_intentos`, y `poner_en_cuarentena` la aparca: `aparcar_desde` guarda los siete ficheros de
`internal/core/plazos/` en `gates/cuarentena/T003.patch` y los retira del árbol (como pasó con los nueve de T002, que
están en `T002.patch`; el de T001 está vacío porque sus guiones viven en `aceptacion/`, dentro del directorio del
feature, que la cuarentena conserva). Con `consecutivas` en 3, `bloqueado` pasa a `true`, se escribe `gates/dossier.md`
y el run se para (causa mayor del ADR 0018: DAG bloqueado).

## Receta para la persona (todo en la raíz del repositorio)

El trabajo de las tres tareas está completo y verificado en un árbol limpio; lo único que falta es retirar los residuos
y devolver el trabajo al árbol. Dos detalles del bucle que, si se olvidan, lo mandan todo otra vez a cuarentena:
`tarea.sh siguiente` incrementa el contador de `gates/tareas-intentos.json` **antes** de compararlo con `max_intentos`
(con T003 en 3 pasaría a 4 y volvería a cuarentena sin ejecutar nada), y una tarea que siga en `gates/cuarentena.json`
no se reintenta. Y el guardián de diff de una tarea rechaza cualquier fichero fuera de sus rutas: si el trabajo de T002 y
T003 vuelve al árbol sin commitear, el reintento de T001 lo vería como fuera de alcance y una sesión de `reparar` lo
desharía. De ahí el orden:

```
# 1. Retirar los residuos sin tocar nada versionado (-empty protege .claude/ de la raíz, que tiene contenido)
find . -path ./.git -prune -o -type d -name .cc-writes -path '*/.claude/*' -empty -print -delete
find . -path ./.git -prune -o -type d -name .claude -empty -print -delete

# 2. Devolver el trabajo de T002 y T003 al árbol, verificarlo y commitearlo (T001.patch está vacío: sus guiones ya están en aceptacion/)
git apply specs/023-h9-internal-core-plazos/gates/cuarentena/T002.patch specs/023-h9-internal-core-plazos/gates/cuarentena/T003.patch
make ci                     # tiene que acabar en «ci: todos los controles en verde»; cada parche lo estuvo por separado en un árbol limpio
git add -A && git commit -m "feat(H9): T002 y T003, desde cuarentena" -m "Trabajo completo de las dos tareas, apartado por los residuos .claude/.cc-writes del arnés (gates/tarea-T002.md, gates/tarea-T003.md)."

# 3. Sacar las tres tareas de cuarentena y poner sus contadores a cero
jq '.tareas = {} | .consecutivas = 0 | .bloqueado = false' specs/023-h9-internal-core-plazos/gates/cuarentena.json > "$TMPDIR/c.json" && mv "$TMPDIR/c.json" specs/023-h9-internal-core-plazos/gates/cuarentena.json
jq 'del(.T001, .T002, .T003)' specs/023-h9-internal-core-plazos/gates/tareas-intentos.json > "$TMPDIR/i.json" && mv "$TMPDIR/i.json" specs/023-h9-internal-core-plazos/gates/tareas-intentos.json

# 4. Reanudar
scripts/hito.sh --resume <run_id>
```

Al reanudar, el bucle toma T001 (sus guiones están en `aceptacion/`: la sesión comprueba que fallan por una aserción y la
marca `[X]`; el workflow los congela y commitea), y después T002 y T003, que la sesión encuentra hechas y en verde y se
limita a marcar (el diff desde su base es solo `tasks.md`). Si se prefiere no commitear a mano, la alternativa es
reanudar solo con los pasos 1, 3 y 4: el bucle rehace T002 y T003 desde cero, con estas notas como guía, y cuesta dos
sesiones más.

El arreglo de raíz sigue siendo fuera de la rama del hito: que `scripts/workflow/verificar.sh` (o `scripts/hito.sh`
antes de cada paso) retire los `.claude/.cc-writes` vacíos antes de medir, y que los prompts de los pasos con modelo
digan que ninguna orden de Bash haga `cd` a un subdirectorio.

## Reanudación (2026-10-18, supervisor)

Los residuos `.claude/.cc-writes` se retiraron y ya no vuelven: el workflow 2.9.1 (#161, cherry-pick `f266c12` en esta rama) los retira al abrir cada sesión, tras cada herramienta y antes de `make ci`. La tarea salió de la cuarentena con su contador a cero. Su trabajo de los intentos anteriores está en `gates/cuarentena/T003.patch` (vacío en T001: sus guiones siguen en `aceptacion/`); devuélvelo al árbol con `git apply` y verifica con `make ci` en primer plano.
