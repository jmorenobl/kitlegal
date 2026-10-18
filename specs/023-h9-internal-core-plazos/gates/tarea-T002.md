# T002, intento 2 (2026-10-18): la tarea está completa y `make ci` en verde en un árbol limpio; en rojo en el árbol de trabajo por el mismo residuo del arnés que paró T001

## Para el intento siguiente, antes de leer nada más

```
ls -d internal/source/boe/testdata/boe.legislacion-consolidada/.claude testdata/evals/boe.legislacion-consolidada/.claude internal/app/testdata/cita/.claude internal/app/testdata/script/.claude internal/skills/testdata/script/.claude
```

- **Si alguno existe**: no repitas ninguna verificación (abajo está todo, medido dos veces). Deja la tarea `[ ]`, añade a
  esta nota la fecha y el número del intento y termina. La sesión no puede retirarlos: `rmdir` sobre una ruta `.claude/`
  pide aprobación («sensitive file»; probado en este intento con los treinta directorios vacíos de `internal/`,
  `testdata/` y `evidencias/`) y en modo desatendido nadie la da; borrarlos con otra orden (`rm -r`, `find -delete`,
  `git clean`, un programa en Go) sería sortear ese rechazo, y no corresponde.
- **Si no existe ninguno**: `make ci` en primer plano y, en verde, marca T002 `[X]`. No toques ningún fichero de la
  tarea: cumplen la tarea tal como están.
- **Nunca `cd` a un subdirectorio en una orden de Bash**: usa rutas absolutas. Es lo que crea el residuo (abajo).

## Lo que está hecho y no hay que tocar

Los nueve ficheros de la tarea, y nada más (`git status`: solo ellos y el directorio del feature):

- `internal/source/boe/fiestas/`: `doc.go`; `testdata/grabaciones.json` (`"fuente": "boe.fiestas-laborales"`,
  `disposiciones` `{id, anio, url}` de `BOE-A-2024-21316` 2025, `BOE-A-2025-21667` 2026 y `BOE-A-2026-21254` 2027, con
  `https://www.boe.es/diario_boe/txt.php?id=<id>`); `grabacion_test.go` (`//go:build grabacion`,
  `TestGrabarFiestasLaborales`: `httpx.New(ConFuente("boe.fiestas-laborales"), ConRaizDeGrabacion($KITLEGAL_EVIDENCIAS),
  ConIntervalo(boe.IntervaloEntrePeticiones))`, solo GET, estado 200 exigido; exige `KITLEGAL_RECORD=1` y
  `KITLEGAL_EVIDENCIAS` no vacía antes de pedir nada).
- `internal/source/bocm/`: `doc.go`; `testdata/grabaciones.json` (`"fuente": "bocm.fiestas-locales"`, las siete
  disposiciones con su año y `https://www.bocm.es/boletin/CM_Orden_BOCM/AAAA/MM/DD/<id>.xml`, cada dirección igual,
  letra por letra, a la que lleva dentro el XML de `evidencias/adr-0039/bocm/<id>.xml`); `grabacion_test.go`
  (`TestGrabarFiestasLocales`, igual, a `10 * time.Second`; la dirección se deriva de la fecha del identificador y el
  manifiesto tiene que coincidir con ella).
- `internal/evals/`: `testdata/grabaciones.json`, el segundo manifiesto, con la forma del de H5 y las siete normas
  (LPAC `a24 a30 a112 a122 a124 daprimera`; LJCA `a46 a128`; LOPJ `acientoochentaydos acientoochentaytres
  acientoochentaycinco`; LTAIBG `a24`; LGT `a97`; Código Civil `art5`; Tráfico `a1` con la búsqueda `texto refundido
  tráfico circulación vehículos motor seguridad vial` y el prefijo `Real Decreto Legislativo 6/2015,`, sin identificador);
  `grabaciones.go` (constante `GrabacionesDeH9 = "testdata/" + boe.NombreDeLaFuente`; `UnionDeGrabaciones` no cambia:
  es T005); `grabacion_test.go` (`TestGrabarEvals` lee los dos manifiestos y construye los dos clientes —uno por raíz,
  un solo `Ritmo`— antes de pedir nada, siembra cada consulta desde H4 y H5 y escribe lo que falta en la raíz de cada
  manifiesto); `grabaciones_test.go` (`TestManifiestoDeGrabaciones` valida también el segundo, subtest
  `repositorio-h9`, al menos siete entradas; `rutaDelManifiestoJuntoA(conjunto)`).
- `gates/supuestos.md`: dos líneas `T002` ([alcance] lector del manifiesto de festivos en el test etiquetado y estado
  200; [interno] palabras de la búsqueda de Tráfico).

Verificado en este intento:

- Cada id de bloque, salvo el `a1` de Tráfico (research S8), está en el índice grabado de su norma: `grep -o` de
  `\"<id>\"` sobre `…_texto_indice.json` de `BOE-A-2015-10565` (H4), `BOE-A-1998-16718`, `BOE-A-1985-12666`,
  `BOE-A-2013-12887`, `BOE-A-2003-23186` y `BOE-A-1889-4763` (H5): uno por id; ningún `a182`, `a183`, `a185` ni `a5`.
- Rojo → verde: `TestManifiestoDeGrabaciones/repositorio-h9` en rojo sin el manifiesto, en verde con él; con
  `"fuente": ""` un momento en el segundo manifiesto, en rojo (`grabaciones_test.go:251`), y se restauró.
- Un arnés temporal etiquetado en cada paquete nuevo (retirado antes de `make ci`): `peticionesDeGrabacion` da 3 y 7
  peticiones iguales a las direcciones del manifiesto; rechaza fichero ausente, dirección distinta de la del
  identificador, año 0, identificador sin la forma del BOE y seis identificadores del BOCM mal formados.
- `go vet -tags grabacion` de los tres paquetes: compila. `TestGrabarFiestasLaborales` y `TestGrabarFiestasLocales` sin
  `KITLEGAL_RECORD`: fallan en su guarda antes de pedir nada, como tiene que ser. Ninguna orden tocó la red.
- `make fmt-check` y `make lint`: `0 issues.` (cuatro palabras que `misspell` marcaba se cambiaron en comentarios).
- **`make ci` en un árbol limpio de `HEAD` (8209889) con los nueve ficheros aplicados** (`git worktree add --detach`
  bajo `$TMPDIR`, parches con `git diff HEAD` y `git diff --no-index /dev/null <nuevo>`, `git apply`, `make -C … ci`):
  **`ci: todos los controles en verde`** (`0 issues.`, `test` e `integration` en verde, `no leaks found`, `go mod
  tidy -diff` limpio). El árbol se retiró con `git worktree remove --force` y `git worktree prune`.
- **`make ci` en el árbol de trabajo: rojo**, solo en `test`, por la única causa de la nota de T001: `read .claude: is
  a directory` (`internal/app`, `grafo_test.go:1842`) y `el conjunto de grabaciones
  ../../internal/source/boe/testdata/boe.legislacion-consolidada no se puede copiar: .claude no es un fichero regular`
  (`internal/evals`, decenas de tests que preparan una caché). Nada de esta tarea toca esos tests ni ese directorio.
- `scripts/workflow/grabar-datos.sh comprobar` no se pudo ejecutar: pide aprobación en esta sesión. Lo que comprueba
  está a mano: `fuente` en los tres manifiestos, las tres con fila revisada en `docs/SOURCES.md` (líneas 42, 73 y 74 de
  `main`, research V1) y un `TestGrabar*` con `//go:build grabacion` en cada paquete.

## La causa

La de `gates/tarea-T001.md`: directorios vacíos `<directorio>/.claude/.cc-writes/` que deja el arnés de Claude Code,
que git ignora por `~/.config/git/ignore` y que los tests que listan o copian un directorio de `testdata/` encuentran y,
con razón, rechazan. Ahora hay cuarenta en el árbol (`find internal testdata evidencias data specs -name .claude`):
los de T001 y, nuevos, `evidencias/adr-0039/bocm/.claude` y `evidencias/adr-0039/.claude`, creados en este intento
por dos órdenes de Bash con `cd` a ese subdirectorio. Eso confirma el candidato de la nota de T001: **el directorio de
trabajo de una orden de Bash recibe el residuo**; una orden con ruta absoluta y sin `cd` no lo deja. Ningún test lee
`evidencias/adr-0039/` (`grep -rln adr-0039` en `internal/`, `scripts/` y `cmd/` solo da un comentario de
`internal/source/bocm/grabacion_test.go`), así que esos dos no rompen `make ci`; los que lo rompen son los de T001.

## Lo que hace falta, y solo lo puede hacer una persona

Lo mismo que pide la nota de T001, que sigue sin hacerse:

1. **Parar el run** antes del intento 3 de T002 (TERM a `scripts/hito.sh`): caería en el mismo muro y la tarea entraría
   en cuarentena con el trabajo completo, la segunda seguida.
2. **Retirar los residuos** sin tocar nada versionado (`-empty` protege lo que tiene contenido, como `.claude/` de la
   raíz):
   ```
   find . -path ./.git -prune -o -type d -name .cc-writes -path '*/.claude/*' -empty -print -delete
   find . -path ./.git -prune -o -type d -name .claude -empty -print -delete
   ```
3. **Reanudar** con `scripts/hito.sh --resume <run_id>`: el intento siguiente de T002 ve el árbol limpio, ejecuta
   `make ci` y marca `[X]`; T001 está en cuarentena por la misma causa y su parche cumple la tarea.
4. **Arreglo de raíz, fuera de la rama del hito**: que el workflow retire esos directorios vacíos antes de medir
   (`scripts/workflow/verificar.sh` antes de `make ci`, o `scripts/hito.sh` antes de cada paso), y que los pasos con
   modelo no hagan `cd` a subdirectorios.

## Intento 3 (2026-10-18T07:02Z): el mismo muro; la tarea sigue completa y pasa a cuarentena

Los cinco directorios de la orden `ls -d` de arriba existen; en total siguen los 40 residuos (`find internal testdata
evidencias data specs -name .claude -type d`), sin ningún fichero regular dentro, y esta sesión no añadió ninguno
(todas sus órdenes corrieron desde la raíz). No se tocó ningún fichero de la tarea ni `tasks.md`.

Medido en este intento, de primera mano:

- `go vet -tags grabacion ./internal/source/boe/fiestas ./internal/source/bocm ./internal/evals`: compila.
- `TestManifiestoDeGrabaciones`: en verde con los dos manifiestos (15 subtests, `repositorio-h9` incluido); con
  `"fuente": ""` un momento en `internal/evals/testdata/grabaciones.json`, en rojo en `repositorio-h9`
  (`grabaciones_test.go:251`); restaurado, de nuevo en verde y el fichero igual que antes.
- `make ci` en primer plano: `fmt-check` y `lint` en verde (make llegó a `test`); **`test` en rojo solo por el residuo
  `internal/source/boe/testdata/boe.legislacion-consolidada/.claude`**, en tres paquetes: `internal/app`
  (`TestLaSalidaDeBoeNoCambiaConElGrafo`: `read .claude: is a directory`), `internal/evals` (`TestInterrumpirLaSesion` y
  los demás que preparan una caché: `el conjunto de grabaciones … no se puede copiar: .claude no es un fichero
  regular`) e `internal/skills` (`TestEsquemaDeNormas/rangos-grabados`, mismo directorio). Los demás paquetes, en
  verde. Ninguno de esos tests lee nada de esta tarea.
- Los residuos no se retiraron, por lo dicho en el intento 2: borrar una ruta `.claude/` pide una aprobación que en
  este modo nadie da, y hacerlo con otra orden sería sortear ese rechazo.

Lo que hace el bucle a partir de aquí (`scripts/workflow/tarea.sh`): la verificación de este intento sale en rojo y
`fallo` añade a esta nota la cola de `gates/ci.log`; al elegir la tarea siguiente, el contador de T002 pasa a 4, mayor
que `max_intentos`, y `poner_en_cuarentena` la aparca: `aparcar_desde` añade con `git add -N` los ficheros nuevos de
fuera del directorio del feature, así que **los nueve ficheros de la tarea quedan enteros en
`gates/cuarentena/T002.patch`** (el de T001 está vacío porque su trabajo vive en `aceptacion/`, que se conserva), el
árbol vuelve a `8209889` y `consecutivas` pasa a 2. T003 caerá por la misma causa, ajena a cualquier tarea:
`consecutivas` 3, DAG bloqueado, y el run se para con `gates/dossier.md`.

Para recuperarlo, en una sesión con persona:

1. Parar el run si no se ha parado solo.
2. Retirar los residuos con las dos órdenes `find` del punto 2 de arriba (no tocan nada versionado).
3. `git apply specs/023-h9-internal-core-plazos/gates/cuarentena/T002.patch` y comprobar con `git status` que aparecen
   los nueve ficheros de la tarea y nada más.
4. En `gates/cuarentena.json`, borrar la entrada `T002` (y la `T001`, cuyo trabajo sigue en `aceptacion/`); en
   `gates/tareas-intentos.json`, bajar `T002` a 2 o quitarla: `siguiente` aparca sin más una tarea cuyo contador
   supere 3.
5. `scripts/hito.sh --resume <run_id>`. El intento que vea el árbol limpio solo tiene que correr `make ci` y marcar
   `[X]`; tras su commit, `grabar_datos` graba los tres manifiestos.

## Reanudación (2026-10-18, supervisor)

Los residuos `.claude/.cc-writes` se retiraron y ya no vuelven: el workflow 2.9.1 (#161, cherry-pick `f266c12` en esta rama) los retira al abrir cada sesión, tras cada herramienta y antes de `make ci`. La tarea salió de la cuarentena con su contador a cero. Su trabajo de los intentos anteriores está en `gates/cuarentena/T002.patch` (vacío en T001: sus guiones siguen en `aceptacion/`); devuélvelo al árbol con `git apply` y verifica con `make ci` en primer plano.
