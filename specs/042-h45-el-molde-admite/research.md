# Research: H45 · El molde admite un paquete con sus plazos y su norma, sin código

Todas las decisiones se tomaron sin persona, con el «Criterio de decisión autónoma» de la constitución. Cada afirmación
sobre una herramienta externa lleva dónde se comprobó (V*); lo que no se pudo ejecutar en esta sesión se dice como
supuesto no verificado. Los prototipos (P*) se ejecutaron en esta sesión, en copias del módulo fuera del repositorio
(`/tmp/claude/h45-*`, el temporal del sandbox), y no dejaron nada en el árbol de trabajo.

## Prototipos (FR-016: «la lista … más los que el plan encuentre con un prototipo entero, incluido `make test-integration`»)

Base de los tres: `main` en `aa96ccce` (la rama del hito sin código), copiado sin `.git`, `web`, `dist`, `bin` ni
`.codegraph` con un copiador de solo biblioteca estándar; el molde, parcheado en la copia de forma rápida (FR-001,
FR-002, FR-005, FR-006 en `internal/skills/regiones.go`); sincronizado con el mecanismo de `make skills-sync`
(`go test -run '^TestSkillsDelRepositorio$' ./internal/app/ -args -regenerar-skills`). Línea base: la misma copia sin
paquete ni parche, `go test -count=1 ./internal/evals/ ./internal/app/ ./internal/skills/`: caen solo 4 tests del entorno
(la copia no es un repositorio git: `TestHuellaDeLasEvals`, `TestFilaConReutilizadas`,
`TestCalculoDeLaDecisionDelCheckout` y `TestEntregaDelHito/h32-nombres-retirados`), que se restan de todo lo que sigue.

- **P1 · la vertical de ensayo** con un tipo (`BOE-A-1900-1`, `a2`), una fila con su precepto en la LPAC y norma de
  ensayo, la región `recursos_y_plazos` en su `ensayo.md` y una eval más de `plazos-y-territorio` (copia de la `05`).
  Con el molde parcheado, `administrativo.md` sale idéntica y el tipo y las dos filas van a `ensayo.md`. Caen 190
  subtests: la mayoría de `internal/evals` porque la preparación de una sesión exige que lo que la skill lleva a leer
  esté grabado («lo grabado no sirve estas consultas: data/plazos.yaml: … boe articulo BOE-A-1900-1 a2») —legítimo:
  un paquete real trae sus grabaciones—, y entre ellos `TestEvalsDelRepositorio/conjunto-plazos-y-territorio`, que cae
  por su cuenta literal y además por lo grabado. Ningún test de la lista §C.5 cae por lo grabado salvo ese.
- **P2 · un paquete `fiscal` sobre lo ya grabado** (como el primer prototipo del hito): la LGT con `vertical: fiscal`,
  un tipo de la LGT (`a66`), una fila con la LGT por norma, una fila de lo cubierto, `fiscal.md` en las tres skills de
  capacidad y una eval más. Sin ruido de grabaciones. Caen, frente a la base: los de la lista del hito y además
  `TestRegiones` (todos sus subtests: `arbolDeRegiones` copia el `data/plazos.yaml` real junto a normas sintéticas,
  `internal/skills/regiones_test.go:160-164`), `TestEvalsDelRepositorio/conjunto-plazos-y-territorio` (fija «las doce
  evals», `internal/evals/conjunto_test.go:2670`), `TestVerticalDeEnsayo` (por el sitio de `normativa`: «cuerpo de 11066
  caracteres (máximo 11000)» con la línea `fiscal` y la de ensayo) y los subtests de
  `TestDefinicionDelJob/tope-con-las-evals-del-repositorio` (líneas de minutos fijas). `TestPreceptosDeLosPlazos` cae
  porque el `a66` no dice «un mes»: comprueba el dato contra lo grabado, no es una lista. **Con `-tags=integration`**
  (`go test -count=1 -tags=integration ./...`, 26 paquetes en verde y los demás con los mismos fallos) no cae ningún
  subtest que no caiga sin la etiqueta.
- **P3 · el paquete de la copia de SC-004** (D3): solo añade —la norma de tráfico `BOE-A-2015-11722`, que tiene
  metadatos, índice y bloque `a1` grabados y no está en `data/normas.yaml`, con `vertical: prueba`; un tipo suyo; una
  fila con ella por norma; su fila de lo cubierto; `prueba.md` solo en `plazos-y-territorio`, con las dos regiones; y una
  eval más—. Tras sincronizar, `prueba.md` lleva la fila de tráfico de hoy (que deja `administrativo.md`, FR-005), la
  nueva y el tipo. Corrida la lista §C.5 más lo encontrado en P1 y P2, con los filtros de D3, caen exactamente los tests
  de listas literales, ninguno por lo grabado ni por el entorno: `TestLeerCobertura/repositorio`,
  `TestCoberturaDelRepositorio`, `TestRegionDeLoCubierto/repositorio` y `/sin-data-cobertura`,
  `TestRenderizarRecursosYPlazos/desde-data`, `TestRegionDeLaCobertura`, `TestRegiones`, `TestVerticalDeEnsayo` (aquí por
  «la fila de la vertical va detrás de la del territorio», `vertical_de_ensayo_test.go:153`), `TestDefinicionDelJob`,
  `TestPlanEnDosModos/del-repositorio`, `TestEvalsDelRepositorio/conjunto-plazos-y-territorio`,
  `TestUmbralesDePlazosYTerritorio`, `TestUmbralesDeLasSkills/plazos-y-territorio`, `TestInformeEnDosModos/plazos-y-territorio`,
  `TestMatrizDeLasPropuestas`, `TestSkillsDelProductoDesdeControl/plazos-y-territorio` y los guiones `h32-paquetes`,
  `h33-cobertura-instalada` y `h36-plazos-de-hoy`. Pasan `precepto-de-la-cobertura`, `ordenes-para-powershell`,
  `TestCasosDePlazos/19-eval-c-tributos`, `h41-4-skills-instaladas` y `TestDescripciones*` (que caen con el cambio de la
  descripción, no con un paquete). Coste medido en una copia nueva (`/tmp/claude/h45-c1`, con `-trimpath`): 7,4 s de reloj
  (46,8 s de CPU) la parte de `internal/skills` e `internal/evals`, y 3,5 s (33,2 s de CPU) la de `internal/app`.

**Lista completa de FR-016** (contracts/tests-sin-listas.md): la del hito y del segundo prototipo, más
`TestRegiones` (P2), `TestEvalsDelRepositorio/conjunto-plazos-y-territorio` (P1, P2), el orden de lo cubierto de
`TestVerticalDeEnsayo` (P3) y `TestDefinicionDelJob/tope-con-las-evals-del-repositorio` (P2; ya dentro de
`TestDefinicionDelJob`, que FR-022 deriva). `make test-integration` no añade ninguno.

## Decisiones

### D1 · El paquete de un tipo y sus dos defectos (FR-001 a FR-004, §C.1)

Decisión: una función sin exportar de `internal/skills/regiones.go`, `paquetesDelTipo(tipo plazos.Tipo, normas
map[string]Norma) []string`, da los `vertical` distintos, en orden alfabético, de las normas de `tipo.Preceptos` y de
los valores de `tipo.Unidades`; una norma que no está en `data/normas.yaml` o sin `vertical` no cuenta, y
`tipo.PreceptosDelComputo` (que `LeerTipos` copia en cada tipo, `internal/core/plazos/tipos.go:89-93`) no se mira.
`filasDeLosRecursosYPlazos` usa por fin su `contextoDeLaRegion`: primero recorre todos los tipos y junta, en el orden de
`data/plazos.yaml`, el defecto de cada tipo sin paquete o de más de uno con los textos de §C.1; si hay alguno, devuelve
`&defectoDeLosDatos{fichero: "plazos.yaml", causa: errors.Join(…)}` —la vía de los demás defectos de `data/plazos.yaml`
(`regiones.go:540-543`), que `anotarLosDatos` pone, con `data/plazos.yaml: ` delante y una vez por skill, en los
defectos de los datos, de modo que la referencia no se regenera y `skills.Escribir` no escribe nada («no se escribe nada
porque hay defectos», visto en P2)—; si no, da las filas de los tipos cuyo único paquete es el de la región, con
`filasDeLosTipos` sin cambiar. Visto en P1: «plazos-y-territorio: data/plazos.yaml: el tipo ensayo no es de ningún
paquete: ninguna norma de sus preceptos lleva vertical en data/normas.yaml», el texto de §C.1.

Alternativas: un campo `paquete:` en `data/plazos.yaml` (fuera de alcance, ADR 0062 opción 2); dar el defecto una vez
en la sincronía, fuera de la región (§C.1 lo quiere «como los demás defectos de `data/plazos.yaml`», en cada skill con
la región); contar los preceptos del cómputo (FR-001 lo prohíbe: el cómputo `lpac` es de lo administrativo y haría de
dos paquetes a todo tipo de otra materia, US3.3). `RenderizarRecursosYPlazos` (exportada, solo la usan tests) no
cambia: pinta los tipos que recibe.

### D2 · El paquete de una fila (FR-005 a FR-007)

Decisión: `paqueteDeLaFila(entrada ReguladoAparte, normas map[string]Norma) string`, sin exportar, en `regiones.go`:
vacío si la norma de `entrada.Precepto.Norma` no está en `data/normas.yaml`, como hoy, que la fila da el defecto de
`filasSinRegion` (corregido en la revisión final, `gates/supuestos.md`); si no, el `vertical` de
`entrada.Norma.Identificador` si la fila lleva `norma` y esa norma está en `data/normas.yaml` con `vertical`; y si no, el
de la norma del precepto (vacío si no lo lleva). La usan
`filasDeLaCobertura` (`regiones.go:510-513`, que hoy compara `norma.Vertical` del precepto) y `filasSinRegion`
(`regiones.go:839-842`), con lo que FR-006 tiene la misma regla por construcción. La línea no cambia (FR-007):
`region.linea.linea(entrada, norma)` sigue recibiendo la norma del precepto, como hoy. Visto en P1 y P3: con los datos de
hoy, ninguna fila cambia de referencia (`administrativo.md` idéntica en P1); con `BOE-A-2015-11722` con `vertical`, su fila
pasa a `prueba.md` con la misma línea.

Alternativa rechazada: dos funciones, una por sitio (lo que hoy permite que las dos reglas diverjan; FR-006 las quiere
iguales).

### D3 · La comprobación de SC-004: un guion congelado que corre la lista sobre una copia del módulo con un paquete más

Decisión: el guion de aceptación `paquete-de-mas.txtar` (activado como `h45-paquete-de-mas.txtar`) lleva en su archivo
un programa de Go de solo biblioteca estándar, `paquete-de-mas.go`, y los ficheros del paquete. El guion fija el entorno
de Go (`HOME`, `GOPATH`, `GOPROXY=off`, `GOENV=off`, `GOFLAGS=-mod=readonly`; D4) y ejecuta `exec go run paquete-de-mas.go
$KITLEGAL_SKILLS/.. repo`, que: (1) copia el módulo bajo prueba (la raíz de `$KITLEGAL_SKILLS/..`) en `repo`, sin `.git`,
`web`, `dist`, `bin`, `.codegraph`, `specs`, `.specify`, `.agents` ni `.claude` y sin enlaces simbólicos (el único de
primer nivel fuera de ellas es `AGENTS.md`); (2) instala el paquete añadiendo entradas al final del bloque de su clave
—la lógica de `anadirEntradas` de `TestVerticalDeEnsayo`, `vertical_de_ensayo_test.go:223-240`— y copiando sus
ficheros; (3) sincroniza con el mecanismo de `make skills-sync`; (4) ejecuta la lista §C.5 en la copia con dos órdenes
`go test -count=1` (los tests enteros de `internal/skills` e `internal/evals` que no tienen filtro de subtest, y con
filtro `^(TestEvalsDelRepositorio|TestCasosDePlazos|TestEntregaDelHito)$/^(…)$` los subtests de la lista); y (5)
imprime una línea de estado por paso («sincronía: 0», «lista: 0», el código de cada orden) y la salida de la orden que
falle en la de errores. Sale siempre con 0 salvo que no pueda copiar o instalar; el guion afirma las líneas con
`stdout`, y además, tras la sincronía, que `repo/skills/plazos-y-territorio/references/prueba.md` lleva el tipo y las
dos filas del paquete y `administrativo.md` ninguna (US2.1 y US2.2 sobre datos reales). Así el rojo de hoy es una
aserción (`aceptacion.sh` toma por inválido un guion que cae con «unexpected command», V3) y su causa es el molde: con
el código de `main`, la sincronía de la copia da defectos.

El paquete (P3) solo añade, para no tocar ninguna entrada de `data/` que otro hito vaya a cambiar: la norma de tráfico
`BOE-A-2015-11722` (grabada: metadatos, índice y `a1` en `testdata/evals/boe.legislacion-consolidada/`; no está en
`data/normas.yaml`) con `vertical: prueba`; un tipo `prueba-de-un-paquete-mas` con su precepto en el `a1`; una fila de lo
regulado aparte con su precepto en la DA 1ª.2.c de la LPAC y la norma de tráfico; una fila de lo cubierto de
`materia: prueba`; `skills/plazos-y-territorio/references/prueba.md` con las dos regiones; y una eval más, la copia del
primer fichero de `evals/plazos-y-territorio/` con el prefijo `99-`. Con él la fila de tráfico de hoy cambia de paquete
por su norma, lo que ejercita los tests que fijan las filas de `administrativo.md`. La fila de lo cubierto no la nombra
US4.1, pero sin ella tres tests de la lista que fijan lo cubierto (`TestLeerCobertura/repositorio`,
`TestRegionDeLoCubierto/repositorio` y la línea 35 de `h33-cobertura-instalada.txtar`, que solo caen con una fila más de
`cubierto`; P3) pasarían en vacío en la copia; es la que trae un paquete que cubre sus plazos, como la vertical de
ensayo. La referencia va solo en
`plazos-y-territorio`: la región la piden el tipo y la fila (FR-006), y con una en `normativa` o `jurisprudencia`
`TestVerticalDeEnsayo` no tendría el sitio de FR-020 dentro de la copia (ya habría un paquete más), lo que no es una
lista.

Alternativas rechazadas: (a) la vertical de ensayo como paquete de la copia: sus normas no tienen grabaciones y
`conjunto-plazos-y-territorio` cae por lo grabado (P1), y `TestVerticalDeEnsayo` no podría instalarla otra vez en la copia
(`os.CopyFS` no sobrescribe, V5); (b) dar `vertical` a la LGT, como el prototipo del hito: edita una entrada que H42
cambiará; (c) copiar con `tar` y añadir con `awk` desde el guion: testscript no tiene copia recursiva y el `--exclude`
de BSD tar no se ha comprobado; (d) una orden nueva del arnés (`modulo`): antes de implementarla el guion cae por
«unknown command» y `rojo-primero` lo da por inválido (V3); (e) un test de Go fuera de la suite: SC-004 la quiere en la
suite de aceptación, congelada, con la lista dentro; (f) correr los tests de la lista sobre otra raíz con una variable:
los guiones usan el binario con `data/plazos.yaml` y las skills empotrados (`skills.go:24`, `data/datos.go:72`), así que
hace falta un binario construido desde la copia, que es lo que da `go test` en ella (el arnés construye los binarios
desde el árbol en que corre, `internal/app/e2e_test.go:225-275`). Un hito futuro que añada `BOE-A-2015-11722` a
`data/normas.yaml` duplicaría la clave en la copia: es un dato de un emisor que todavía no existe, bajo el umbral de
materialidad; ese hito ajustaría el guion en una tarea `[datos]`, como H45 hace con `h32-paquetes`.

### D4 · El entorno de Go de los guiones de `internal/app`

Decisión: el `Setup` de `ejecutarGuiones` (`internal/app/e2e_test.go:600-616`) añade a cada guion `GOCACHE` y
`GOMODCACHE`, leídos una vez en la preparación del arnés con `go env -json GOCACHE GOMODCACHE`, como hace
`internal/skills/instalacion_test.go:161-171` con sus guiones. Lo demás lo fija el guion con `env`. Lo escribe la tarea
`[aceptacion]` con los guiones, para que `rojo-primero` los ejecute con un `go` que funciona y el rojo sea el del molde.
Ningún otro guion cambia: `kitlegal` no lee esas variables.

Por qué: testscript solo da a un guion `WORK`, `PATH`, `GOTRACEBACK`, `HOME=/no-home`, `TMPDIR` y `devnull` además de lo
de `Setup` (V2), y con `HOME=/no-home` `go` no tiene caché de compilación ni de módulos (V4). Alternativa: que el guion
use un `GOCACHE` vacío en `$WORK`: sin `GOMODCACHE` poblado y con `GOPROXY=off` no hay módulos.

### D5 · Las grabaciones de los hitos (FR-015, §C.4)

Decisión: `UnionDeGrabaciones() ([]string, error)` lee `../../evidencias` (relativa al directorio del paquete, la
convención de las constantes de hoy, `internal/evals/grabaciones.go:13-17`) con una función sin exportar
`unionDeGrabacionesDe(raiz string) ([]string, error)`, y devuelve `GrabacionesDeH4`, `GrabacionesDeH5`,
`GrabacionesDeH9` y, detrás, `raiz/evidencias/<carpeta>/boe.legislacion-consolidada` de cada carpeta cuyo nombre es el de
un hito —`h` y su número, y para un subhito un guion y el suyo, comprobado con `strings.CutPrefix`, `strings.Cut` y
`strconv.Atoi`, sin `regexp`— que tiene `manifiesto.json` y esa carpeta, ordenadas por número de hito y después de
subhito. `GrabacionesDeH36` desaparece. El error de leer `evidencias/` se devuelve envuelto; los llamadores de
producción (`medida.go:1158`, `medida.go:1244`, `sesiones.go:441`, `dependencias.go:819`) ya devuelven error y lo
propagan. `rutaQueMide` (`dependencias.go:234`, «sin leer el disco», bool) deja de enumerar raíces: un predicado puro
`esDeUnaRaizDeGrabaciones(ruta)` dice si la ruta empieza por una raíz fija o por `evidencias/<carpeta de un
hito>/boe.legislacion-consolidada/`, con la misma función del nombre. Las raíces siguen siendo las del arnés que corre,
como hoy las constantes (que no dependen del árbol `c.arbol` que mira `dependencias.go`).

Alternativas: recibir un `fs.FS` (cambia también lo que mide `dependencias.go` sobre otro árbol, lo que no pide el
spec); `sync.OnceValues` (estado global sin necesidad); seguir sin error (habría que callar el de `os.ReadDir`, que la
constitución prohíbe). Un `evidencias/` truncado o ilegible lo cubre la regla genérica: el error sube y quien prepara la
sesión falla con él, como con cualquier otro error de lectura, sin casos propios.

### D6 · El control de las órdenes para PowerShell (FR-011, FR-012, §C.3)

Decisión: en `controlDeLasOrdenes` (`internal/evals/conjunto_test.go:4736`), las órdenes del `SKILL.md` se comprueban con
una regla más: una orden de Bash del `SKILL.md` que no sigue su forma para PowerShell no da defecto si alguna
referencia a mano de la skill (`skills.ReferenciasAMano`, `internal/skills/referencias.go:174`, la que ya usa
`leerLaSkillEntera`) lleva una orden PowerShell con el mismo verbo, la misma norma y los mismos bloques
(`esSuFormaParaPowerShell`, `:4337`, sin cambiar), en cualquier sitio. Las órdenes de una referencia siguen con
`defectosDeLasOrdenes` (`:4316`) sobre su propio fichero. Los mensajes no cambian. Los mutantes de
`TestControlesSobreLasReferenciasAMano` (`:5028`, `mutantesDeLasOrdenes` `:5224`) ganan los casos 1 a 4 de §C.3; los
casos 5 y 6 ya son `orden-con-su-forma-en-otra-referencia` y `sin-las-ordenes-de-articulos`, que se quedan.

Alternativa rechazada: exigir que la referencia esté enlazada desde el paso de la orden o se lea siempre (Fuera de
alcance, Clarifications).

### D7 · El laboratorio de H41 (FR-010, FR-014)

Decisión: `compararLosFicheros` (`internal/evals/laboratorio_familias_test.go:972`) deja de dar «sobra» para un
`skills/<skill>/references/<v>.md` del producto que la base no tiene cuando `<v>` es un `vertical` de
`data/normas.yaml`, leído con `skills.LeerNormas` de `tablaDeNormasDelRepositorio` (`grabaciones_test.go:331`); una
referencia que la base sí tiene se sigue comparando. La declaración de `normativa` (`:439-523`) gana un `cambio` con
`antes` las líneas 38 a 41 del cuerpo (con su sangría) y `despues` vacío, y un `movimiento` a `con_herramientas.md` con
`enLaReferencia` igual a `"\n"` + las cuatro líneas sin sangría + `"\n"` (§C.2): un movimiento no quita nada del
`SKILL.md` por sí solo (`conLaDeclaracion`, `:916-947`), y el subtest `movimientos` (`:714`) exige que el texto movido
no siga en el cuerpo (US5.4).

### D8 · La descripción (FR-023)

Decisión: el texto de FR-023 en el frontmatter de `skills/plazos-y-territorio/SKILL.md` y en
`laboratorio/producto-descripciones-cortas/plazos-y-territorio.txt`, y en `descripcionCortaDePlazosYTerritorio`
(`laboratorio_descripciones_test.go:38-42`). Los subtests con sumas literales derivan la suma esperada de las cuatro
descripciones leídas al empezar el test: «una de 451» y «una quinta skill sin texto» suman lo que añaden (25 y 20) a la
suma de hoy; «una suma de 1701» añade a `plazos-y-territorio` `1701 − suma` letras, para que la suma sea 1701 como hoy;
«una que empieza por <ejemplo>» espera el defecto de la suma solo si la suma de hoy más 10 pasa de 1700 (con la nueva,
1702, sí). Los textos que anclan esos subtests («tiene un municipio.») siguen al final de la nueva descripción.

### D9 · El trabajo de evals (FR-021, FR-022)

Decisión: `concurrencia: 3` para `plazos-y-territorio` en `.github/workflows/evals.yml` (`:254`) con su comentario
(`:232-239`, la cuenta con tres y las doce sesiones), `ajustesDelTrabajo` (`definicion_test.go:254-259`) y
`sesionesALaVezComoMucho = 12` (`:162`). `probarElPeorCasoDelTrabajo` (`:3194`) y `probarElPlanDelRepositorio`
(`plan_test.go:295`) calculan lo esperado desde `LeerConjunto` (series por modo de cada eval, con el modelo que decide y
los informativos, por repeticiones, y la sesión de la prueba de red) en lugar de `[]int{67, 66, 6}`, `20725 *
time.Second` y los mapas literales; el peor caso lo sigue dando `DefinicionDelJob.peorCaso` (`definicion.go:752`), que ya
deriva de `evals/`. Una copia de `evals/` en `t.TempDir()` (`os.CopyFS`, como `preparar_test.go:990`) con cuatro copias de
más de una eval de modo de `plazos-y-territorio` pasa por `peorCaso` y tiene que caber en los 352 min. Los mutantes de
`tope-con-las-evals-del-repositorio` (`:1702-1732`) dejan de fijar los minutos: cada caso calcula sus líneas con el peor
caso de la definición que construye. Con la fórmula de `peorCaso` (tandas de ⌈sesiones del grupo / concurrencia⌉ por
272 s más 1 685 s fijos): hoy, a tres, 47 tandas, 14 469 s; con cuatro evals más (91, 90 y 6 sesiones), 63 tandas,
18 821 s (313,7 min) a tres y 94 tandas, 27 253 s (454,2 min) a dos. Son cuentas, no medidas: las mide el test.

### D10 · Los tests sin listas (FR-016)

Decisión: cada test de la lista completa deriva lo esperado de `data/` y de `evals/` si comprueba el repositorio, o
usa datos de `testdata/` si comprueba una regla, con el detalle por test en contracts/tests-sin-listas.md. Los guiones
de `internal/app` cambian en una tarea `[datos]` (FR-017). `TestRegiones` pasa a un `data/plazos.yaml` de prueba en
`internal/skills/testdata/regiones/plazos.yaml` (comprueba reglas, no el repositorio). Se mantiene
`TestNormasDelRepositorio/paquete-de-lo-administrativo` (fuera de la lista por el spec).

### D11 · La vertical de ensayo crecida (FR-019, FR-020)

Decisión: `internal/skills/testdata/vertical-de-ensayo/` gana `data/plazos.yaml` (un tipo `ensayo` con su precepto en
`BOE-A-1900-1`, `a2`, cómputo `lpac`), una fila en `data/cobertura.yaml` con su precepto en la DA 1ª.2.c de la LPAC y su
`norma` `BOE-A-1900-2`, y la región `recursos_y_plazos` en `skills/plazos-y-territorio/references/ensayo.md`.
`instalarLaVertical` añade también las entradas de `tipos`. `TestVerticalDeEnsayo` exige: el tipo `ensayo` en la región
de `ensayo.md` y en ninguna otra referencia de la copia; la fila con norma en `ensayo.md` y en ninguna línea de
`administrativo.md`; que la fila de lo cubierto de la vertical vaya detrás de la última de la región de hoy (en lugar de
«detrás de la del territorio»); y el sitio de FR-020 con la comprobación de producción: tras sincronizar, una línea de
69 caracteres y su salto añadida al final del cuerpo de cada skill que tiene la referencia de la vertical
(`jurisprudencia`, `normativa`, `plazos-y-territorio`) no da ningún defecto de `skills.Regenerar` («cuerpo de N
caracteres (máximo 11000)», `sincronia.go:621-626`), sin exportar nada. Un subtest con la `ensayo.md` de
`plazos-y-territorio` sin la región de la cobertura exige el defecto de FR-006 para la fila con norma. Medido en P2: con
dos líneas de paquete, el cuerpo de `normativa` sin el movimiento da 11 066; con él, 176 caracteres menos.

### D12 · Los subtests de los defectos (US3, SC-003)

Decisión: tres subtests nuevos en `TestRegiones/defectos` (`internal/skills/regiones_test.go`), sobre el
`plazos.yaml` de prueba de D10 con un tipo cambiado: preceptos de una norma sin `vertical` (defecto de §C.1, sin
paquete); un precepto de lo administrativo y otro de una norma de prueba con otro `vertical` (defecto de más de uno, con
los dos en orden alfabético); y un tipo de lo administrativo con un cómputo que cita una norma de otro paquete (cero
defectos).

### D13 · `mutantes.md` (FR-027)

Decisión: lo escribe la última tarea del hito, una entrada por control de «Criterios del hito (literales)» con la orden
exacta, lo que se espera y lo visto. Cada mutante se aplica en una copia del módulo en el directorio temporal, como en
P1-P3 (`git worktree` no se puede en el sandbox), nunca en el árbol de trabajo, el índice ni el historial; la comparación
«contra `main`» se hace con `git show main:<ruta>` o `git diff main -- <rutas>`, que solo leen.

### D14 · Skills de Go aplicadas

`golang-how-to` → `golang-testing` (tablas de casos, `t.Parallel`, `t.TempDir`, `os.CopyFS` para copias, subtests con
nombre), `golang-error-handling` (`errors.Join` para los defectos de una región, `%w` al subir el error de
`evidencias/`, ningún error callado), `golang-naming` (funciones sin exportar en español, como el paquete),
`golang-code-style` y `golang-lint` (la configuración de `.golangci.yml` sin cambios; ningún `//nolint` nuevo),
`golang-project-layout` (ningún paquete nuevo: todo vive en `internal/skills`, `internal/evals` e `internal/app`).

## Verificaciones (dónde se comprobó cada afirmación sobre una herramienta)

- **V1 · Toolchain**: `go version` → `go1.27.4 linux/amd64`; `go env` → `GOCACHE=/home/jorge/.cache/go-build`,
  `GOMODCACHE=/home/jorge/go/pkg/mod`, `GOPROXY=https://proxy.golang.org,direct`, `GOTOOLCHAIN=auto` (ejecutado).
- **V2 · Entorno de un guion**: `github.com/rogpeppe/go-internal@v1.16.0/testscript/testscript.go:492-497`: `WORK`,
  `PATH`, `GOTRACEBACK=system`, `HOME=/no-home`, `TMPDIR`, `devnull`; lo demás, de `Setup`.
- **V3 · Prefijos y fallos de testscript**: `testscript.go:749-757` solo reconoce `!` (no hay `?` en v1.16.0); un `exec`
  que falla da «unexpected command failure» (`cmd.go:274`), y `scripts/workflow/aceptacion.sh:55` cuenta como inválido el
  guion cuya salida casa `unknown command|cannot parse|unexpected command|usage: `.
- **V4 · Cachés de Go**: `go help cache` (línea 3: la caché vive «in the standard user cache directory», y `GOCACHE` la
  sustituye) y `go help environment` (`GOCACHE`, `GOMODCACHE`, `GOPATH`); con `HOME=/no-home` la caché por defecto queda
  bajo un directorio que no existe. No pude ejecutar `go` con `HOME=/no-home` en esta sesión (el permiso del sandbox
  rechazó `env -i`): **supuesto no verificado** que el guion, con el entorno de D4, corre `go` como los de
  `internal/skills/testdata/script/` con el de `prepararInstalacion` (`instalacion_test.go:161-171`), que `make ci`
  ejecuta en `test-integration`.
- **V5 · `os.CopyFS`**: `go doc os.CopyFS`: no sobrescribe un fichero que existe (`fs.ErrExist`).
- **V6 · `go run` de un fichero fuera de un módulo**: ejecutado en esta sesión (`go run /tmp/claude/h45-tools/…/main.go`,
  sin `go.mod` en su árbol) con solo biblioteca estándar.
- **V7 · `go -C`**: tiene que ser la primera bandera (`go test -C …` da «-C flag must be first flag on command line»,
  visto); el programa del guion usa `cmd.Dir` en lugar de `-C`.
- **V8 · Los binarios del arnés**: `internal/app/e2e_test.go:225-244` los construye con `go install` en el árbol en que
  corre, el de v0.1.0 con `ldflags` y sin red.
- **V9 · Lo grabado de `BOE-A-2015-11722`**: `ls testdata/evals/boe.legislacion-consolidada/` lleva sus `metadatos`,
  `texto_indice` y `texto_bloque_a1`; `grep` de `data/normas.yaml` no la tiene (ejecutado).
- **V10 · Tamaño del árbol copiado**: `du -sh` sin las carpetas excluidas, 49 MB en el barrido global; `specs` (78 MB) y
  `.specify` (35 MB) no entran, excluidas también en D3.
- **V11 · El directorio de trabajo de un guion**: `testscript.go:355-359` lo borra al acabar salvo con `TestWork` o
  `-testwork`, que el arnés de `internal/app` no pone; la copia de D3 no queda en el disco.
- **V12 · `go env -json`**: `go help env` («usage: go env [-json] [-changed] [-u] [-w] [var ...]»; línea 10, «The -json
  flag prints the environment in JSON format»), la forma que usa D4 para leer `GOCACHE` y `GOMODCACHE`.

## Datos externos

Ninguno. El hito no graba nada ni deriva ningún fichero de `data/`: no hay paso `grabar_datos` con contenido, ni
manifiesto `grabaciones.json` nuevo ni test `TestGrabar*`. La copia de SC-004 solo lee grabaciones que ya están en el
repositorio (V9). `docs/SOURCES.md` no cambia.

## Fuera de alcance confirmado

Lo de la sección «Fuera de alcance» del spec, sin añadir nada. En particular: el defecto de un paquete con tipos y sin
referencia con la región; un control que rechace que una descripción nombre una materia; que el control de PowerShell
mire cuándo se lee una referencia; cambiar el texto de FR-023 en el run.
