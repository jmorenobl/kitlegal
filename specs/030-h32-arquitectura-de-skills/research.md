# Research · H32 · Arquitectura de skills

Decisiones del plan (D), comprobaciones hechas en esta sesión contra el repositorio o la herramienta (V) y supuestos que
no se han podido comprobar sin red ni modelo (S). Cada decisión, con la alternativa rechazada y su porqué
(constitución, «Criterio de decisión autónoma»).

## Decisiones

### D1 · Los nombres: `consulta-juridica`, `normativa`, `jurisprudencia`, `plazos-y-territorio` (FR-007, FR-042)

- **Decisión**: cada skill de capacidad se llama por su capacidad, en español, en minúsculas y con guiones, como
  sintagma nominal: `normativa`, `jurisprudencia`, `plazos-y-territorio` (dos capacidades, D2). La de entrada, por lo
  que hace: `consulta-juridica`. `jurisprudencia` conserva su nombre: ya es el de su capacidad.
- **Por qué**: FR-007 pide nombres de capacidad, no de fuente —H34 añade EUR-Lex a `normativa` sin renombrarla—, con el
  mismo patrón, en español y como sustantivo; ninguno lleva `boe`, `cendoj`, `eur`, `anthropic` ni `claude`. Conservar
  `jurisprudencia` evita un nombre retirado más, con sus evals y su juez movidos, sin ganar nada.
- **Rechazadas**: un prefijo común (`kitlegal-normativa`…), que repite la marca y obliga a renombrar `jurisprudencia`;
  `derecho` para la entrada, demasiado genérico para activarse con precisión («Design coherent units»); `plazos` y
  `territorio` por separado (D2).

### D2 · Plazos y territorio en una sola skill (FR-005)

- **Decisión**: `plazos-y-territorio` reúne `kitlegal plazos` y `kitlegal territorio`.
- **Por qué**: todo plazo con festivos locales resuelve antes sus municipios con `territorio` (paso 8; evals 05, 06, 09
  a 11). Separadas, cada pregunta de plazos cargaría dos skills con el mismo paso de resolver el municipio: el defecto
  de una skill demasiado estrecha de la guía de agentskills.io. Las de solo territorio (evals 01 y 02) ya se activan
  con precisión con la skill de hoy, que las junta.
- **Rechazada**: dos skills, `plazos` y `territorio`: duplica los pasos 1 a 4 y añade un trabajo a la matriz del job, que
  ya va a once sesiones a la vez (D15).

### D3 · El molde: el paquete es su `vertical`, sus referencias, sus filas y sus evals, sin fichero de declaración (FR-022)

- **Decisión**: el nombre de un paquete es el valor de `vertical` de sus normas en `data/normas.yaml`; su referencia en
  una skill es `skills/<skill>/references/<nombre>.md`; sus filas de cobertura, las de `data/cobertura.yaml` cuyo
  precepto es de una norma suya; sus evals, las de la carpeta de la skill que las hereda ([contracts/paquetes.md](contracts/paquetes.md)).
- **Por qué**: es la lista del ADR 0044, punto 5 —«normas con `vertical: <nombre>`, un fichero de referencia por skill
  de capacidad, sus filas de cobertura y sus evals, y nada más»—, sin una pieza más que la persona tenga que escribir.
  La pertenencia de una fila sale del precepto, que ya está en los datos.
- **Rechazadas**: un `data/paquetes.yaml` que declare cada paquete y sus referencias (una pieza que el ADR no lista); un
  campo `vertical` en cada fila de `data/cobertura.yaml` (dato repetido: el precepto ya dice su norma); los packs
  (`packs/`), fuera de alcance.

### D4 · La referencia de un paquete: escrita a mano, con las tablas generadas entre marcas (FR-012, FR-019, FR-022)

- **Decisión**: la referencia de lo administrativo en `plazos-y-territorio` es un fichero a mano con la prosa del
  procedimiento común y tres regiones generadas (`cobertura_por_materia`, `recursos_y_plazos`, `leyes_vertebrales`),
  cada una entre sus marcas y con su «generada desde data/…» ([contracts/paquetes.md](contracts/paquetes.md) §3).
- **Por qué**: el paquete lleva una referencia por skill (FR-022; ADR 0044, punto 5) y en ella van a la vez lo que sale
  de `data/` (FR-019: «generadas desde los mismos datos») y lo que solo pide el procedimiento común, que es protocolo
  escrito a mano (el acto presunto, el silencio, la reclamación de acceso). Es el mecanismo que ya usa el `SKILL.md`
  con la tabla de comandos (`SustituirRegion`, `internal/skills/comandos.go`), llevado a una referencia.
- **Rechazadas**: una referencia generada entera desde tres ficheros de `data/`, que no puede llevar la prosa; tres
  referencias por skill (las de hoy), que no son «una referencia por skill»; un generador por materia en Go, que es un
  caso especial de materia en el código (principio X).

### D5 · Qué normas llevan `vertical: administrativo` (FR-021, FR-022)

- **Decisión**: las diez de lo administrativo —el procedimiento común y el acceso a la información ante cualquier
  administración, y lo que regula su actuación y su control—: LPAC, LRJSP, LRJPAC, LTAIBG, LRBRL, LCSP, LGS, LGP, TRLRHL
  y LJCA. No la llevan la CE, el CC, la LEC ni la LOPJ (la base: lo constitucional, lo común y lo procesal) ni la LGT,
  la LIRPF, el ET ni la LOPDGDD (otras materias, sin paquete). La Ley 50/1980 entra sin `vertical` (D13).
- **Por qué**: es la materia de validación del principio X, y la regla se lee en los `materias` de cada norma
  (V12). La LPAC tiene que llevarla: es la norma de los preceptos de las nueve filas de cobertura (D3). La LJCA, también:
  es la del recurso cuyo plazo cuenta `plazos`, y FR-021 la saca entonces de los cuerpos.
- **Rechazadas**: las quince vertebrales, que marcarían la LOPJ —su etiqueta «⚠ DÍAS DEL ART. 183 LOPJ:» la escribe el
  binario y va en el cuerpo de `plazos-y-territorio`— y la CE, la del ejemplo de `normativa` (D8); solo la LPAC, que
  dejaría como base normas que solo trae esta materia.
- No es un dato externo: es una clasificación que fija el plan, sin fuente que grabar, como `vertebral`. Los
  identificadores y los títulos de esas normas ya están en `data/` y los comprueba `TestIdentificadoresDeLasNormas`.

### D6 · La línea `⚠ NO CUBIERTO:` de cada fila sale de la forma del cuerpo (FR-011, FR-012)

- **Decisión**: el generador de la región `cobertura_por_materia` lee las dos formas `⚠ NO CUBIERTO:` del bloque `text`
  del `SKILL.md` de la skill que tiene la referencia y sustituye sus marcadores con los datos de la fila
  ([contracts/paquetes.md](contracts/paquetes.md) §3).
- **Por qué**: FR-012 pide «la forma del cuerpo que le corresponde con cada marcador sustituido», y su mutante —la forma
  del cuerpo cambiada sin regenerar— lo pone en rojo la deriva de `make skills-check`. Escrita en Go, la forma estaría
  dos veces y sería un caso de materia en el código.
- **Rechazada**: la plantilla de la línea en `internal/skills`.

### D7 · La forma legible del precepto, con la abreviatura de su norma (FR-010, FR-019)

- **Decisión**: `art. <apartado>` o el apartado, un espacio y la abreviatura de `data/normas.yaml` (o el título hasta
  la primera coma), y la cita: `DA 1ª.2.a LPAC [BOE-A-2015-10565, bloque daprimera]`.
- **Por qué**: la forma única es `<forma legible> [<identificador>, bloque <id>]`; «DA 1ª.2.a» sola no dice de qué
  norma es. Sin artículo delante, la forma vale en las dos líneas sin concordancias de género.
- **Rechazada**: el apartado solo, como hoy dentro del paréntesis.

### D8 · Los ejemplos de `normativa`, con la Constitución (FR-021)

- **Decisión**: `BOE-A-1978-31229`, bloque `a140`, en las órdenes del paso 2 y en «Cómo se cita».
- **Por qué**: FR-021 saca la Ley 39/2015 de los cuerpos; un ejemplo concreto se ejecuta en `ordenes-de-ejemplo` contra
  la caché de las sesiones, que sirve el `a140` (eval 09, V9), y no lleva marcadores que el agente pueda copiar tal
  cual. La CE no la marca ningún paquete (D5).
- **Rechazadas**: marcadores (`<identificador> <id>`) en las órdenes: el control no las ejecuta y la guía de autoría
  pide ejemplos concretos; el Estatuto de los Trabajadores, que es de otra materia.

### D9 · El sitio en los cuerpos (FR-009, FR-024)

- **Decisión**: cada cuerpo de capacidad deja sitio para la línea de un paquete más: la de la vertical de ensayo (70
  caracteres con su salto), que el test de FR-024 añade en una copia. Lo que sale o se acorta, con su texto de `c991808`, está en
  [contracts/textos.md](contracts/textos.md); las cifras, medidas en esta sesión (V4).
- **Por qué**: el test de la vertical de ensayo corre `skills-check` sobre la copia con su línea en cada skill de
  capacidad: un cuerpo a menos de 70 caracteres del límite lo pondría en rojo. Lo que sale de `normativa` es la viñeta
  «Nada de otra conversación», que dice ahora el *gotcha* del paso 2 con su porqué, y frases que el texto nuevo del paso
  5 ya cubre; ninguna forma fija, ni la tabla de comandos, ni lo que toda respuesta necesita.
- **Rechazada**: llevar a una referencia la forma para PowerShell del paso 2: `ordenes-para-powershell` la exige junto a
  su orden de Bash, en el mismo fichero.

### D10 · La skill de entrada no declara applets ni referencias (FR-001, FR-003, FR-038)

- **Decisión**: `consulta-juridica` no tiene `kitlegal-applets` ni tabla de comandos; nombra las skills de capacidad
  como código en línea, y la elección de las evals le da sus paquetes y sus datos por el cierre de las skills que nombra
  (`calcularDependencias`). `dependencias.go` deja de dar un error con una skill sin tabla cuando no declara applets.
- **Por qué**: FR-005 pone cada applet en una sola skill de capacidad; una skill sin herramienta no es de capacidad.
  `make skills-sync` ya admite un `SKILL.md` sin `kitlegal-applets` y sin marcas (`internal/skills/sincronia.go`).
- **Rechazada**: dar a la entrada un applet de capacidad para que tenga tabla: rompería FR-005.

### D11 · La retirada: la lista en la composición, el dominio genérico, dos estados y un aviso en `data` (FR-025 a FR-030)

- **Decisión**: `skillsRetiradas` en `internal/app/retiradas.go`; `instalacion.Planificar` la recibe como argumento;
  `SkillInstalada` gana los estados `retirada` y `sin retirar` y el campo opcional `aviso`
  ([contracts/instalacion-retiradas.md](contracts/instalacion-retiradas.md)).
- **Por qué**: el dominio (`internal/core/instalacion`) no lleva hoy ningún nombre de skill y sigue sin llevarlo; la
  composición ya sabe qué skills se empotran. `data` sigue siendo la lista de hoy, con dos valores más, y la consume
  igual quien la lee: una entrada por skill.
- **Rechazadas**: `data` como objeto con listas aparte (rompe la forma de hoy, un array); un verbo nuevo
  (`kitlegal skills retirar`), que el spec no pide; condicionar la retirada a que el ámbito tenga las herederas
  (fuera de alcance por la clarificación 3).

### D12 · Las huellas se comparan con el manifiesto, no con lo empotrado (FR-026, FR-027)

- **Decisión**: una skill de nombre retirado se retira si su directorio tiene exactamente los ficheros que le declara el
  manifiesto, con su huella (`HuellaDe`, SHA-256 de los bytes), y cada entrada de host es la declarada.
- **Por qué**: el binario ya no empotra esa skill: el manifiesto es lo único que dice qué instaló kitlegal (H19).
- **Rechazada**: compararla con la skill de `c991808`, que el binario no lleva.

### D13 · La eval de la póliza y la Ley 50/1980 en `data/normas.yaml` (FR-035)

- **Decisión**: la pregunta habla de la prórroga de la póliza tras el fallecimiento, y la eval exige la cita del
  art. 22 (`a22`, la prórroga) y sirve el 4 (`a4`) como lectura. La Ley 50/1980 entra en `data/normas.yaml`
  (`materias: [seguros]`, sin `vertical`), con su identificador y su título de la búsqueda grabada, y
  `TestIdentificadoresDeLasNormas` resuelve las normas de la tabla con el manifiesto de H5 y, la que no está en él, con
  el de H9.
- **Por qué**: la caché de las sesiones no sirve `boe buscar` salvo con los términos exactos de un comando de la eval
  (V9), así que, fuera de la tabla, la sesión no encontraría el identificador sin red. En la tabla, `normativa` lo toma
  en el paso 1 y la caché sirve su índice y sus metadatos. El manifiesto que dispara `grabar_datos` es el de H9
  (`internal/evals/testdata/grabaciones.json`, V11); el de H5 no se graba en el paso.
- **Rechazadas**: un comando de búsqueda en la eval (solo serviría unos términos); la pregunta con el número del
  artículo, que ya no es la de `docs/USO.md`; la eval como `informativa`, que no decidiría y dejaría SC-015 sin
  control.

### D14 · `se_activan` en el formato de las evals (FR-036)

- **Decisión**: un campo nuevo, lista de nombres de skill, que la sesión tiene que activar además de la de su carpeta;
  se juzga como `no_se_activan`, al revés.
- **Por qué**: US1, escenario 2: «se activan la skill de entrada y la de esa capacidad». Hoy el job solo comprueba la
  skill de la carpeta (`juzgarEnModo`, V10); los comandos de la eval no prueban que la otra se activó.
- **Rechazada**: dar la activación por buena con los comandos esperados.

### D15 · La matriz del job: la entrada a una sesión y `jurisprudencia` a tres (FR-039, SC-008)

- **Decisión**: `consulta-juridica` 1, `jurisprudencia` 3, `normativa` 5, `plazos-y-territorio` 2: once. El tope no
  cambia: 352 minutos cubren el mayor peor caso, 20 725 s ([contracts/evals.md](contracts/evals.md) §6).
- **Por qué**: hoy son 5 + 4 + 2 = 11 y la entrada necesita una. `jurisprudencia` a tres pasa de 13 717 s a 17 637 s,
  dentro del tope; `normativa` a cuatro pasaría a 24 549 s, por encima; `plazos-y-territorio` a una, a más de 38 000 s.
- **Rechazadas**: subir el tope (el límite de un trabajo de un runner de GitHub no se ha podido comprobar sin red, S5, y
  no hace falta); un trabajo aparte para la entrada (más de once sesiones a la vez).

### D16 · `cobertura_sin_declarar`: el precepto, en la eval (FR-014)

- **Decisión**: `no_cubierto.precepto: {norma, bloque}` en las evals 07, 08 y 12, comprobado contra `data/cobertura.yaml`
  en `TestEvalsDelRepositorio`; la medida exige en la misma línea la norma y esa cita.
- **Por qué**: el informe no lee `data/` al juzgar (`InformeAEscribir` no tiene datos), y la norma sola no dice la fila
  de la eval 12, cuya norma esperada es la propia LPAC.
- **Rechazada**: buscar la fila en `data/cobertura.yaml` al juzgar.

### D17 · Los dos controles nuevos, en `TestSkillsDelRepositorio` (FR-010, FR-013)

- **Decisión**: los subtests `forma-de-cita` y `formas-compartidas`, sobre todas las skills, con sus mutantes en
  copias, como `normas-nombradas` ([contracts/forma-de-cita.md](contracts/forma-de-cita.md)).
- **Por qué**: valen para todas las skills, también la de entrada y las que no tienen evals; los controles de
  `conjunto_test.go` se aplican a las skills con evals y nombran sus skills. `make skills-check` ya corre ese test.
- **Rechazada**: un control de contenido en `internal/evals`.

### D18 · El test de la vertical de ensayo, en `internal/skills` (FR-024)

- **Decisión**: `TestVerticalDeEnsayo`, paquete `skills_test` de `internal/skills`, con la vertical en
  `internal/skills/testdata/vertical-de-ensayo/`.
- **Por qué**: prueba la sincronía, que vive en `internal/skills`; para cargar la eval importa `internal/evals`, como ya
  hace `internal/skills/normas_test.go` (V12). Las dos normas de ensayo tienen identificadores que ninguna búsqueda del
  BOE da: son datos de un test, no normas de la base.
- **Rechazada**: ponerlo en `internal/app`, que no es donde vive lo que prueba.

### D19 · Los nombres de hoy en los tests y los comentarios (FR-032)

- **Decisión**: donde un test o un comentario nombra una skill para operar con ella, pasa al nombre de la entrega
  (`ReglasDeBoeLegislacion` → `ReglasDeNormativa`, `skillBoeLegislacion`, las rutas de `evals/` y `skills/`, los nombres
  de ejemplo de `internal/core/instalacion`). Donde un test compara con una sesión grabada, conserva el nombre que la
  grabación lleva, como dato. Lo mismo vale para los dos tests de la raíz que `make ci` corre y que nombran una skill
  de hoy (`git grep -n -E 'boe-legislacion|legal-core' -- '*.go' ':!internal/'`, en esta sesión): `skillsDeHoy`
  (`skills_test.go`, línea 36), que `TestSkillsEmpotradas` exige con `require.Subset`, pasa a las cuatro skills de la
  entrega; y la orden de la comprobación `skills-install` de `comprobacionesDelHumo` (`release_test.go`, línea 1261),
  copia de la del paso de `.github/workflows/release.yml` que FR-032 cambia, pasa a
  `.agents/skills/normativa/SKILL.md`, a la vez que el paso (su nombre, el `if` y el `echo`).
- **Por qué**: FR-032 busca en `internal/` y en `.github/workflows/`; lo grabado no se reescribe (constitución) y está
  fuera de la búsqueda. Los dos tests de la raíz quedan fuera de esa búsqueda, pero el renombrado los pone en rojo en
  `make ci` (SC-001): el primero, porque el árbol pierde los nombres de `c991808`; el segundo, porque fija entera la
  orden del paso. El humo comprueba `normativa` porque hereda de `boe-legislacion` lo que el paso comprobaba.
- **Rechazada**: dejar `skillsDeHoy` con solo las dos herederas; la lista dice las skills que el repositorio tiene, y
  con las cuatro la comparación entre lo empotrado y el árbol tampoco pasa en vacío.

### D20 · El ADR del hito: 0048 (FR-042)

`docs/ADR/0048-las-skills-se-reparten-por-capacidad-con-una-de-entrada.md`, MADR corto: contexto (los tres defectos y
el ADR 0044, punto 7), decisión (los cuatro nombres, el reparto, los nombres retirados y sus herederas, el molde) y la
tabla de nombres de `c991808` a la entrega. El último ADR del repositorio es el 0047.

### D21 · La serie y la reutilización con un nombre nuevo (FR-038)

Ninguna decisión: es la consecuencia del cambio de árbol (supuesto del spec). Lo que ocurre, en
[contracts/evals.md](contracts/evals.md) §5. `scripts/workflow/` no cambia (fuera de alcance).

### D22 · La entrada sin grabar entre el commit de T002 y la grabación (FR-035)

- **Decisión**: T002 añade la entrada «contrato de seguro» al manifiesto de H9 y, en
  `internal/evals/grabaciones_test.go`, una excepción temporal en `verificarLasNormasDeLaCobertura`: esa entrada, y
  solo ella, no cuenta como defecto cuando su búsqueda da la falta de la reproducción (no está grabada); con cualquier
  otro error, o cualquier otra entrada, el defecto es el de hoy. T003, la primera tarea con lo grabado, quita la
  excepción, y la función queda como en `c991808`.
- **Por qué**: `verificar.sh tarea` corre `make ci` antes del commit y `grabar_datos` graba después; el test, sin
  etiqueta de compilación, resuelve contra lo grabado cada entrada del manifiesto de H9 (`TestPreceptosDeLosPlazos` →
  `verificarLasNormasDeLaCobertura` → `resolverEntrada` → `reproducirBusqueda`), y ni la búsqueda ni
  `BOE-A-1980-22501` están grabadas. Sin la excepción, T002 no puede quedar en verde, no hay commit y no se graba nada.
  H26.1 y H31.1 añadieron entradas o bloques a ese manifiesto sin ella porque sus búsquedas ya estaban grabadas.
- **Rechazadas**: rebajar la verificación para siempre (dejaría pasar un manifiesto que no está grabado); un manifiesto
  aparte (`grabar_datos` solo graba `<paquete>/testdata/grabaciones.json`, y el de H5 no lo es: V11); cambiar el orden
  del workflow (`scripts/workflow/` está fuera de alcance); una búsqueda ya grabada que diera la Ley 50/1980 (ninguna
  la da).

## Comprobaciones hechas en esta sesión

| # | Qué | Dónde se comprobó |
|---|---|---|
| V1 | `git grep` sale con 1 cuando no hay coincidencias | `rtk bypass git -C <repo> grep -c 'zz-cadena-que-no-existe-qq' -- README.md` → «Exit code 1»; git 2.56.0 (`git --version`) |
| V2 | `testscript` tiene `exists`, `grep`, `stdout`, `cmp`, el prefijo `!` y `exec`, con `PATH` real y `HOME=/no-home` | `go doc github.com/rogpeppe/go-internal/testscript`, desde la caché de módulos |
| V3 | `git diff -M` detecta renombrados | `git diff -h` (git 2.56.0) |
| V4 | Las medidas de los cuerpos: 10 995, 10 944 y 10 859 en `c991808`; 10 858, 10 842, 10 896 y 3 660 en los borradores | `awk 'f>=2{print} /^---/{f++}' <fichero> \| wc -lm`, la receta de `docs/SKILLS.md` §6, sobre `skills/*/SKILL.md` y `contracts/borradores/*.SKILL.md` |
| V5 | El peor caso de hoy, 20 117 s, 13 717 s y 20 725 s, y el tope de 352 minutos | comentarios de `.github/workflows/evals.yml` (líneas 257 a 258 y 276) y la fórmula de `internal/evals/definicion.go`; las cifras nuevas de [contracts/evals.md](contracts/evals.md) §6 son esa fórmula aplicada a mano, y las fija `TestDefinicionDelJob` |
| V6 | El juez de las citas acepta `[<identificador>, bloque <id>]` con la forma legible delante; la línea del plazo, con la fecha tras la etiqueta | `formaDeCita` (`internal/evals/citas.go`) y `formaDelPlazo` (`internal/evals/plazos.go`) |
| V7 | `⚠ PLAZO SIN CALCULAR:` se juzga por los identificadores de su línea | `ExtraerPlazoSinCalcular` e `identificadoresDeLasLineas` (`internal/evals/plazos.go`) |
| V8 | El control `lineas-de-aviso-de-las-referencias` no mira una fila de tabla | `inicioDeLineaDeAviso` (`internal/evals/conjunto_test.go`): exige `⚠` al principio, tras sangría o marcador de lista |
| V9 | La caché de las sesiones sirve `buscar` solo con los términos de un comando de la eval, y el índice y los metadatos de cada norma de `data/normas.yaml` | `ConsultasNecesarias` y `anotarLosDatos` (`internal/evals/consultas.go`) |
| V10 | La sesión solo se juzga activada por la skill de su trabajo, y `no_se_activan` añade un motivo por cada nombre activado | `juzgarEnModo` y `anotarLasQueNoSeActivan` (`internal/evals/juzgar.go`), leídos por un subagente de esta sesión |
| V11 | `TestGrabarEvals` graba los manifiestos de H5 y de H9 y solo escribe en la raíz de H9; las entradas de H9 con sus bloques, `titulo_empieza_por` y `para` | `internal/evals/grabacion_test.go` (líneas 28 a 60) e `internal/evals/testdata/grabaciones.json` |
| V12 | Los `materias` de las 18 normas de `data/normas.yaml`; `internal/skills/normas_test.go` es `package skills_test` e importa `internal/evals` | `data/normas.yaml`; cabecera de `internal/skills/normas_test.go` |
| V13 | Las skills se empotran con `//go:embed skills/*/SKILL.md skills/*/references/*`: una referencia de un paquete en `references/` viaja en el binario | `skills.go`, línea 24 |
| V14 | Los borradores cumplen FR-010 (ninguna cita de un bloque fuera de la forma), FR-013 (las tres líneas compartidas, iguales) y FR-021 (ninguna norma del paquete en los cuerpos; solo en las descripciones) | `grep -nE` sobre `contracts/borradores/` en esta sesión |
| V15 | Fuera de `internal/`, solo dos `.go` nombran una skill de hoy: `skills_test.go:36` (`skillsDeHoy`, que `TestSkillsEmpotradas` exige con `require.Subset`) y `release_test.go:1261` (la orden de `skills-install` en `comprobacionesDelHumo`, que `probarHumo` compara con el paso de `release.yml`, líneas 189 a 198); los dos pasan hoy y corren en `make test` (`./...`) | `git grep -n -E 'boe-legislacion\|legal-core' -- '*.go' ':!internal/'` y `go test -count=1 -run '^(TestSkillsEmpotradas\|TestConfiguracionDeLaRelease)$' .`, en verde, en la sesión del corrector del plan |
| V16 | Una skill sin juez, sentencias ni evals de plazo, de cobertura, de vigencia o con una redacción fechada no tiene umbrales de sus respuestas, tampoco `sin_activar`; `lectura_sin_intentar` pide un comando de bloque en `sin_servir`; el de la descripción va con la descripción en el informe | `medirLasRespuestas` (`internal/evals/informe.go`, líneas 1584 a 1595), `internal/evals/doc.go` (líneas 219 y 220, 232 a 238) y el comentario de `Nombre` en `internal/evals/umbrales.go` (líneas 249 a 255) |

## Supuestos no verificados

| # | Supuesto | Qué pasa si no se cumple |
|---|---|---|
| S1 | La búsqueda «contrato de seguro» del BOE da la Ley 50/1980 con un título que empieza por «Ley 50/1980,», y su índice tiene los bloques `a4` y `a22` | `grabar_datos` falla nombrando la entrada: causa mayor del run (dependencia externa inaccesible, supuesto del spec); una persona corrige la entrada del manifiesto y reanuda |
| S2 | La sesión que activa `consulta-juridica` activa después la skill de la capacidad que la pregunta pide | lo mide el job de cierre: la sesión que no activa la skill de su `se_activan` no pasa, y la serie de la eval decide por su regla (SC-015); `consulta-juridica` no publica `sin_activar` ([contracts/evals.md](contracts/evals.md) §7) |
| S3 | La sesión de la póliza cita el art. 22, el de la prórroga | lo mide el job de cierre (SC-015) |
| S4 | `git diff --find-renames` detecta como renombrados los ficheros de `evals/` con una o dos líneas cambiadas (umbral por omisión del 50 % según la documentación de git, no consultada sin red) | `git diff --find-renames=90%` lo dice igual; lo coteja la revisión final (FR-033) |
| S5 | El límite de seis horas de un trabajo de un runner de GitHub | no se usa: el tope no cambia (D15) |
