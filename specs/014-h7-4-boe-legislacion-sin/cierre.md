# H7.4 · Cierre de la Definition of Done

Cierre de T013 (FR-047, FR-072, FR-090, FR-091, FR-092; SC-011) sobre `95be315`, que es T012, el último commit antes de
T013. La rama está al día con `main`: `git merge-base HEAD main` da la cabeza de `main`, `df4e1a9`, así que
`git diff --name-status main` da exactamente lo que cambia el hito. Las cuatro partes son las de la tarea: los esquemas,
evals y datos de prueba que toca el hito, la cobertura, el quickstart y lo que el hito no cambia.

## 1. Esquemas, evals, lista y datos de prueba tocados en el hito

Esta es la lista para la capa 3 del informe final. Recoge cada fichero de `schemas/`, de `evals/` y de cualquier
`testdata/` que el hito crea o modifica, con la tarea que lo tocó y el motivo. Sale de `git diff --name-status main --
schemas evals testdata '*/testdata/*'`, y la tarea de cada fila, de `git log --format=%s main..HEAD -- <fichero>`. Son
doce ficheros: seis `M` y seis `A`. No hay ninguna `D` ni ninguna `R`.

### Esquemas (`schemas/`)

| Estado | Fichero | Tarea | Motivo |
|---|---|---|---|
| M | `schemas/expresiones-prohibidas.yaml.json` | T001 `[datos]` | La lista gana dos claves (FR-030, FR-031, FR-035; contracts/lista-de-expresiones.md §2). `redaccion_no_leida` es la familia de la clase B, con el mismo `$ref` a `#/$defs/expresiones` que las otras tres. `formas_fijas` es una lista de al menos una cadena de una línea con algún carácter fuera de los marcadores. Las dos entran en `required` y `additionalProperties: false` se queda. El diff son 7 líneas añadidas y 2 quitadas. Lo fijan los casos de `formato_test.go`: la lista del repositorio valida, y no validan una lista sin `redaccion_no_leida`, otra sin `formas_fijas`, otra con `formas_fijas: []` ni otra con una forma de dos líneas. |
| M | `schemas/eval.yaml.json` | T002 `[datos]` | La eval gana dos claves (FR-003, FR-004, FR-053; contracts/evals-y-juicio.md §1). `no_se_activan` es una lista no vacía de nombres de skill sin repetir y vale en toda eval. `redacciones_modificadas` es una lista de `redaccion-modificada` (`norma`, `bloque`, `fecha_vigencia`, `fecha_vigencia_reciente`), entra en el `anyOf` de las positivas, como `hallazgos`, y una eval de no activación no la admite. También entran las definiciones `redaccion-modificada` y `fecha`, esta con un mes y un día posibles. El diff son 22 líneas añadidas y 2 quitadas. Los dos esquemas no salen de `--describe`, así que `schema-check` no los compara. |

### Evals (`evals/`)

| Estado | Fichero | Tarea | Motivo |
|---|---|---|---|
| A | `evals/boe-legislacion/20-lcsp-dos-bloques-redaccion-cambiada.yaml` | T009 `[datos]` | La eval 20, informativa y que activa la skill, igual carácter a carácter a contracts/evals-y-juicio.md §3 (FR-050 a FR-052, FR-054). Pregunta por el art. 118 (`a1-30`) y la disposición adicional tercera (`da-3`) de la LCSP con un grafo previo, `lcsp-a1-30-y-da-3-redaccion-original`, que vio sus redacciones de 20180309. Exige leer los dos bloques, `graph check` con la norma, no pedir `graph show`, las dos citas, `hallazgos: [version-obsoleta]` y dos `redacciones_modificadas`: `a1-30` de 20180309 a 20200206 y `da-3` de 20180309 a 20230101. Son 48 líneas. `boe-legislacion` pasa a 20 evals: 10 positivas que deciden, 2 de no activación y 8 informativas. |
| M | `evals/boe-legislacion/expresiones-prohibidas.yaml` | T001 y T010 | La lista por clases (FR-030 a FR-033). T001 añade el comentario de cabecera de contracts §1 y `redaccion_no_leida` con su primera expresión (`ya no exige`). También añade `formas_fijas` con las dos formas definitivas: la línea `⚠ REDACCIÓN MODIFICADA: <cita>: …` con sus dos `<fecha>` y la frase de la regla 7 sin su punto final. T010 añade a `anuncio` las 25 formas de la clase A detrás de sus 14 y a `redaccion_no_leida` las otras nueve. Queda así: `maquinaria` 22, `otra_conversacion` 16, `anuncio` 39 y `redaccion_no_leida` 10, **87 expresiones** frente a las 52 de `main`, y 2 formas fijas. El diff son 46 líneas añadidas y 4 quitadas, las del comentario. Está calibrada en `expresiones-calibradas` sobre los tres informes versionados: marca 36, 11 y 9 respuestas de 93 (§3). |
| M | `evals/legal-core/01-territorio-municipio-cubierto.yaml` | T002 `[datos]` | Gana `no_se_activan: [boe-legislacion]` detrás de `activa`, y nada más (FR-004). Su pregunta, sus comandos y su juicio no cambian. |
| M | `evals/legal-core/02-territorio-municipio-no-cubierto.yaml` | T002 `[datos]` | Lo mismo que la 01. |
| M | `evals/legal-core/03-no-activa-receta-de-cocina.yaml` | T002 `[datos]` | Lo mismo que la 01. |

`legal-core` se queda con sus 3 evals. Las evals 01 a 19 de `boe-legislacion` no cambian.

### Datos de prueba (`testdata/`)

| Estado | Fichero | Tarea | Motivo |
|---|---|---|---|
| A | `testdata/evals/grafo-previo/lcsp-a1-30-y-da-3-redaccion-original/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2017-12902_texto_bloque_a1-30.json` | T009 `[datos]` | Derivada del grafo previo de la eval 20 para `a1-30` (FR-051, FR-096; contracts/evals-y-juicio.md §4). La escribió `TestGrabacionesDerivadas -actualizar-derivadas` a partir de la grabación de H4, quitándole solo la redacción de 20200206. Son 52 líneas y es byte a byte la de la eval 19: `cmp` con `lcsp-a1-30-redaccion-original/…a1-30.json` no da ninguna diferencia. |
| A | `testdata/evals/grafo-previo/lcsp-a1-30-y-da-3-redaccion-original/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2017-12902_texto_bloque_da-3.json` | T009 `[datos]` | La derivada de `da-3`: su grabación de H4 sin la redacción de 20230101. Son 52 líneas. `TestGrabacionesDerivadas` sin la bandera la rehace byte a byte (§3, subpruebas `lcsp-a1-30-y-da-3-redaccion-original/a1-30` y `…/da-3`). |
| A | `internal/evals/testdata/sesiones/leer-sesion/respuesta-antes-de-una-tarea-en-segundo-plano/sesion.jsonl` | T005 `[datos]` | Transcript sintético del caso nuevo de `TestLeerSesion` (FR-060, FR-098; contracts/evals-y-juicio.md §6). En este orden: `system/init`, la activación de `boe-legislacion`, un `Bash` en segundo plano y su resultado, la respuesta, un `result` con la respuesta con cita, el aviso `task_notification`, la réplica y un segundo `result` con la réplica. Son 9 líneas y 3 273 bytes. |
| A | `…/respuesta-antes-de-una-tarea-en-segundo-plano/codigo-de-la-sesion` | T005 `[datos]` | `0`. |
| A | `…/respuesta-antes-de-una-tarea-en-segundo-plano/sesion.err` | T005 `[datos]` | Vacío (0 bytes). |

### Ninguna grabación ni ninguna fuente cambia

- **Ninguna grabación cambia.** `git diff --name-status main -- internal/source docs/SOURCES.md evidencias
  '*grabaciones.json' '*grabacion_test.go' internal/core` no imprime nada. Las respuestas grabadas de `a1-30` y `da-3`
  de `BOE-A-2017-12902` están desde H4 en `internal/source/boe/testdata/boe.legislacion-consolidada/`, y ninguna de las
  dos cambia. Fuera de las tablas de arriba, `git diff --name-status main -- testdata '*/testdata/*'` no nombra ningún
  otro fichero. No hay ningún manifiesto `grabaciones.json` nuevo ni ningún test `//go:build grabacion` nuevo: los dos
  que hay, `internal/evals/grabacion_test.go` e `internal/source/boe/grabacion_test.go`, no cambian. El paso
  `grabar_datos` no tenía nada que grabar (FR-055; plan.md, «Datos externos»; research D21).
- **Ninguna fuente cambia**: `docs/SOURCES.md` es igual que en `main` (§4).
- **Los tres informes del calibrado** de H7.1, H7.2 y H7.3 solo se leen: sus carpetas no cambian (§4, FR-090).

### Lo que no está en la lista

- **`skills/boe-legislacion/SKILL.md`** (M, T010) está fuera de esos árboles. Es la skill v0.1.4, con **295 líneas**
  (270 en `main`; 69 añadidas y 44 quitadas). Su prosa la comprueban `prosa-de-la-skill` y `ordenes-para-powershell`,
  y su tamaño y su `description`, `TestSkillsDelRepositorio` en `skills-check` (§3).
- **`internal/app/grafo_test.go`** (M, T009): `derivadasDelGrafoPrevio()` gana las dos entradas de las derivadas de
  arriba. Es el único fichero de `cmd/` o de `internal/` fuera del paquete de evals que cambia (§4).
- **`.github/workflows/evals.yml`** (M, T008 y T012), **`scripts/evals-sondeo.sh`** (M, T011), **`CHANGELOG.md`** y
  **`CONTRIBUTING.md`** (M, T012), y el paquete `internal/evals`, que no son datos.
- **Sondas y mutantes momentáneos** (T003 a T011): se crearon y se retiraron dentro de su tarea. `git log --name-only
  main..HEAD` nombra 80 ficheros distintos, y ninguno lleva `zz-`, `.orig`, `mutante` ni `sonda` (0 coincidencias).

## 2. Cobertura (Definition of Done §1.9)

La cobertura se mide sobre los dos perfiles que deja el `make ci` de esta tarea, en verde (`ci: todos los controles en
verde`): `coverage.out` (`make test`) y `coverage-integration.out` (`make test-integration`). El global sale de
`go tool cover -func` sobre cada perfil. El de cada árbol sale de `go tool cover -func` sobre el perfil filtrado a sus
líneas, con su cabecera `mode:`, escrito en un temporal fuera del repositorio y borrado después. Los recuentos de
sentencias salen de los bloques del perfil, contando cada bloque una vez. «Unión» son los dos perfiles fundidos bloque a
bloque, como hace Codecov con los dos que publica CI.

| Árbol | `coverage.out` | `coverage-integration.out` | Unión | Umbral (`codecov.yml`) |
|---|---|---|---|---|
| Global | **96,8 %** (11 539 de 11 916 sentencias) | **97,4 %** (11 601 de 11 916) | 97,4 % | ≥ 70 % |
| `internal/core/**` (dominio) | **98,6 %** (2 545 de 2 581) | **98,6 %** (2 545 de 2 581) | 98,6 % | ≥ 85 % |
| `internal/cli/**` | 98,7 % (372 de 377) | 98,7 % | 98,7 % | ≥ 90 % |
| `internal/evals` | 97,1 % (3 280 de 3 376) | 98,1 % (3 311 de 3 376) | 98,1 % | — |

Los tres umbrales de proyecto se cumplen. No hace falta ningún test, así que T013 no toca ningún fichero de test. El
hito no toca el dominio: `git diff --name-status main -- internal/core` no imprime nada, y `internal/core` da las mismas
2 545 de 2 581 sentencias que en el cierre de H7.3. En `coverage.out`, `-func` sobre el paquete de evals da 97,1 %, y el
recuento de bloques, 97,16 %: el recuento incluye las funciones anónimas de nivel de paquete, que `-func` no cuenta.

**Los ficheros nuevos del paquete de evals** dan lo mismo en los dos perfiles:

| Fichero | Sentencias | Funciones (`go tool cover -func`) |
|---|---|---|
| `internal/evals/redacciones.go` | 31 de 31 (100 %) | `ExtraerRedaccionesModificadas`, `fechasDeVigencia` y `texto`, al 100 % |
| `internal/evals/tanda.go` | 124 de 133 (93,2 %) | `esperarLaDecision` 92,9 %, `consultarLasEjecuciones` 75,0 %, `leerLasEjecucionesDelCommit` 85,7 %, `leerLaTandaDeLaEjecucion` 91,7 % y `escribirLaDecision` 81,8 %. Las otras diez (`cuentaPara`, `decidirLaTanda`, `listar`, `verLosTrabajos`, `ejecutar`, `String`, `registroDeLaConsulta`, `motivoDeLaDecision`, `nombrarLasEjecuciones` y `dormir`), al 100 % |

Las nueve sentencias de `tanda.go` sin cubrir son todas ramas de error de la regla genérica: un error que nombra la
consulta, la ejecución o la ruta, y que `TestTandaDelCommit` convierte en un fallo del paso. Son estas:

- la espera entre consultas que termina con el contexto (175-177);
- el error de `listar` y el de la lista que no se puede leer, devueltos por `consultarLasEjecuciones` (192-193 y
  197-198);
- el error de `verLosTrabajos` y el del JSON de los trabajos que no se puede leer (207-208 y 212-213);
- los dos `json.Unmarshal` que fallan (232-233 y 272-273);
- la escritura y el cierre de `GITHUB_OUTPUT` que fallan (427-429 y 432-433).

Una respuesta de `gh` que no se puede leer sigue la regla que ya hay, sin caso propio (tasks.md, «Proporcionalidad»). El
error de la consulta en sí sí tiene caso: `segundo-disparo/consulta-que-falla`.

**El diff.** H7.4 no pide ninguna cifra del diff, pero Codecov la mide en la propuesta de cambio con el estado `patch`,
bloqueante y con `target: auto`, que es la cobertura de la base. He hecho una estimación local con el criterio de los
cierres de H7.1 a H7.3: cada línea añadida de los `.go` de producto de `git diff -U0 main` que cae en un bloque de la
unión de los dos perfiles cuenta, y cuenta como cubierta si alguno de esos bloques lo está. Da **92,5 %: 335 de 362
líneas**. La base es la cobertura de `main`: 97,4 % de la unión en el cierre de H7.3, y desde entonces `main` solo
cambia 30 líneas de producto de `internal/evals`. La estimación queda por debajo de la base, así que es probable que
`codecov/patch` salga por debajo de su objetivo. Las 27 líneas que ningún perfil ejecuta están todas en `internal/evals`:

| Fichero | Líneas | Qué son |
|---|---|---|
| `internal/evals/tanda.go` | 175-177, 192-193, 197-198, 207-208, 212-213, 232-233, 272-273, 427-429, 432-433 (20) | Las ramas de error de arriba. |
| `internal/evals/avisos.go` | 72-73, 84 (3) | `formaFija` y `patronDeEtiqueta` con una etiqueta sin palabras. T003 las reescribió al compartir el patrón con las formas fijas. Ninguna etiqueta de aviso ni de hallazgo está vacía, y `cabezaDeFormaFija` descarta una forma cuya etiqueta lo esté antes de compilarla. |
| `internal/evals/prohibidas.go` | 281-282 (2) | El cierre del grupo opcional de la cita cuando una forma fija termina en `<cita>`. Ninguna de las dos formas de la lista termina así. |
| `internal/evals/sondeo.go` | 66-67 (2) | `(*errorDeUso).Unwrap`. `TestSondeo` y los tests reconocen el error de uso con `errors.As`, que lo encuentra sin desenvolverlo, y nadie pregunta por su causa con `errors.Is`. |

Queda como supuesto `[alcance]` de T013 en `gates/supuestos.md`. La tarea solo pide tests si la cobertura global o la
del dominio quedan por debajo de su umbral, y no quedan. Si `codecov/patch` sale rojo, lo decide quien lea el informe,
como en H7.1 a H7.3. Ningún umbral se toca.

## 3. Quickstart (§1 a §7 y las dos primeras órdenes de §8)

Se ejecutaron sobre `95be315`, desde la raíz del repositorio. §1 es el `make ci` de esta tarea. Las órdenes de §2 a §8
no se reescribieron: las extrajo del propio `quickstart.md` un programa fuera del repositorio, que ejecutó cada una con
`rtk proxy bash -c <orden>`, sin el resumen del proxy de la sesión, y contó los `--- PASS` y los `--- FAIL` de su salida.
A las de `go test` sin `-v` las ejecutó además con `-v`, para contar sus tests y subpruebas y que ninguna pasara en
vacío. Antes y después, `git status --porcelain` daba lo mismo: solo los dos ficheros de `gates/` que el workflow
modifica antes de la tarea. Tampoco quedaba ningún `kitlegal-sondeo.*` en el `TMPDIR`, ni antes ni después (0 y 0).
**Todos los escenarios dan lo esperado.**

| Escenario | Resultado |
|---|---|
| §1. `make ci` | En primer plano: `ci: todos los controles en verde`. `0 issues.` de `golangci-lint`, `ok` en todos los paquetes con `-race` y con la etiqueta `integration`, `ok` en `TestMedidasDeTiempo` y `TestCosteDelGrafo`, `No vulnerabilities found.`, `ok` en `TestEsquemasPublicados` (`schema-check`, sin drift) y en los siete tests de `skills-check` (`ok` en `internal/app`, `internal/skills` e `internal/evals`, sin drift), `no leaks found`, `all modules verified` y `go mod tidy -diff` sin salida. Es el `make ci` del que salen los perfiles del §2 y también la verificación de esta tarea. |
| §2. La lista, su calibrado y las formas de la skill | Primera orden: `ok`, y con `-v`, **17 `--- PASS`, ningún `--- FAIL`**. Son `TestEvalsDelRepositorio` y sus 16 subpruebas: `expresiones-calibradas`, `expresiones-en-los-bloques`, `expresiones-de-la-skill`, `prosa-de-la-skill`, `ordenes-para-powershell`, `grafo-previo`, `grabado`, `formato`, `conjunto`, `conjunto-legal-core`, `normas-conocidas`, `cobertura-del-esquema`, `avisos-del-esquema`, `avisos-de-la-skill`, `hallazgos-del-esquema` y `hallazgos-de-la-skill`. Segunda orden (`-v`): **5 `--- PASS`, ningún `--- FAIL`**: el test y las cuatro subpruebas que nombra. El reparto del calibrado lo comprueba la propia subprueba contra la tabla de contracts/lista-de-expresiones.md §4 (36, 11 y 9 de 93), eval por eval y columna por columna. |
| §3. El juicio de las dos clases y de la eval 20 | `ok`, y con `-v`, **146 `--- PASS`, ningún `--- FAIL`**. `TestJuzgarLasClasesDeLaRespuesta`, 28: el test y 27 frases. Cada frase de la bitácora va marcada en su clase, y quedan sin marcar la respuesta compuesta (`-_skill`), `-_sin-avisos` y `-_fr-022`, mientras que se marcan `-_la-linea-fuera-de-su-forma` y `-_las-dos-fuera-de-su-forma`. `TestExtraerExpresionesProhibidas`, 39: el test y 38, entre ellas `redaccion-modificada-con-cita`, `-con-enfasis`, `-con-otra-cita` y `-sin-cita`, `la-que-se-consulto-antes-fuera-de-la-forma`, `las-palabras-de-las-dos-formas-fuera-de-ellas`, `cuatro-familias`, `clase-b` y `sin-formas-fijas`. `TestJuzgar`, 64: el test y 63, entre ellas `no-se-activan` y `no-se-activan-detras-del-motivo-de-legal-core` (la sesión de `legal-core` que activa `boe-legislacion` no pasa), y los casos de la eval 20: `redacciones-modificadas-una-por-bloque`, `-leidas-en-dos-ordenes`, `redaccion-modificada-una-sola`, `-sin-cita`, `-con-la-cita-de-un-bloque-y-las-fechas-del-otro`, `-con-las-fechas-en-otro-orden`, `-en-una-sola-linea`, `-ausentes-detras-del-hallazgo` y `sin-redacciones-modificadas-el-juicio-de-antes`. El patrón no lleva `$`, así que ejecuta además `TestJuzgarLasExpresionesProhibidas` (14) y `TestJuzgarLasEvalsSinHallazgos` (1), que también pasan. |
| §4. `boe-legislacion` v0.1.4 | `make skills-check`, código 0, con `ok` en `internal/app`, `internal/skills` e `internal/evals`. `wc -l`: **295**, por debajo de 300. La tercera orden da `ok`, y con `-v`, **32 `--- PASS`, ningún `--- FAIL`**: `TestEvalsDelRepositorio` con sus 16 subpruebas, entre ellas `ordenes-para-powershell` (las dos órdenes de lectura y comprobación con su forma para PowerShell), y `TestOrdenesParaPowerShell` con 14. Esas 14 son `las-dos-ordenes-con-sus-dos-formas`, `sin-la-de-powershell`, `con-ampersands-en-la-de-powershell`, `sin-el-if`, `con-otra-condicion`, `con-otra-norma-en-la-de-powershell`, `con-otros-bloques-en-la-de-powershell`, `con-otros-bloques-en-la-comprobacion-de-bash`, `con-otros-bloques-en-la-comprobacion-de-powershell`, `sin-la-de-articulos`, `sin-la-ultima-de-powershell`, `sin-argumentos-en-la-comprobacion`, `en-codigo-en-linea` y `sin-ninguna-orden`. |
| §5. Los umbrales y los recuentos | **27 `--- PASS`, ningún `--- FAIL`**. `TestUmbralesDelInforme`, 16: el test y 15, con los casos de contracts/informe-del-job.md §6. Son `sin-activar-una`, `sin-activar-ninguna`, `redaccion-no-leida-una`, `redaccion-no-leida-ninguna`, `tres-de-54`, `dos-de-54`, `tres-de-54-y-seis-sin-terminar` y `una-sin-activar-con-expresion`, más los de antes: `900-s-con-objetivo-900`, `901-s-con-objetivo-900`, `dos-de-30-de-haiku-4-5`, `sin-lista-ni-objetivo`, `objetivo-negativo`, `todas-las-de-sonnet-5-5-sin-medir` y `los-tres-motivos-en-su-orden`. `TestInformeMarkdownDeLosUmbrales`, 5: `las-filas-del-contrato`, `los-cinco-sin-cumplir`, `con-redacciones-modificadas` y `sin-umbrales`. `TestInformeConSesionesSinMedir`, 6: `mensaje-del-limite-de-uso`, `reintentos-agotados`, `cortada-durante-reintentos`, `sin-abrir-tras-el-limite-de-uso` y `reintentos-de-los-que-se-recupera`. |
| §6. La respuesta a la pregunta | **2 `--- PASS`, ningún `--- FAIL`**: `TestLeerSesion` y `TestLeerSesion/respuesta-antes-de-una-tarea-en-segundo-plano`. |
| §7. La tanda única y la eval 20 en el job | Primera orden (`-v`): **53 `--- PASS`, ningún `--- FAIL`**. `TestDefinicionDelJob` tiene cinco subpruebas de primer nivel: `del-repositorio`, `sinteticas`, `errores`, `segundo-disparo` y `estado-de-la-tanda`. `segundo-disparo` lleva `anterior-que-mide-corre`, `anterior-que-mide-espera`, `anterior-sin-decidir-que-mide`, `anterior-sin-decidir-que-no-mide`, `anterior-que-no-mide`, `anterior-terminada`, `posterior-que-mide`, `sola`, `espera-agotada` y `consulta-que-falla`. `estado-de-la-tanda` lleva `marca-en-success`, `marca-saltada`, `tanda-en-curso`, `sin-tanda` y `tanda-saltada`. Segunda orden: `ok`, y con `-v`, **8 `--- PASS`, ningún `--- FAIL`**: `TestGrabacionesDerivadas` con 7, entre ellas `lcsp-a1-30-y-da-3-redaccion-original/a1-30` y `lcsp-a1-30-y-da-3-redaccion-original/da-3`. |
| §8, primera orden (sin la credencial) | La salida estándar es solo `código: 2`. La salida de error, exactamente dos líneas: `falta la credencial: CLAUDE_CODE_OAUTH_TOKEN, el token de la suscripción que da claude setup-token, no está en el entorno o está vacía` y la de `make`, `make: *** [evals-sondeo] Error 1`. No sale nada de `go test` ni de testify (ni `---`, ni `FAIL`, ni `Error Trace`). Tardó 1,5 s y no abrió ninguna sesión. |
| §8, segunda orden (argumentos que no valen) | La salida estándar es solo `código: 2`. La salida de error, exactamente tres líneas: `MODELO: está vacío`, `REPETICIONES: «0» no es un entero mayor o igual que 1`, en ese orden, y la de `make`. Tardó 0,8 s y no abrió ninguna sesión. La credencial tampoco estaba en el entorno, pero los argumentos se comprueban antes (`comprobarElSondeo`), así que su línea no sale. |
| §8, tercera orden (con modelo) | **No se ejecuta.** Abre nueve sesiones con la suscripción de quien la lanza, así que queda fuera del run. La lanza una persona al leer el informe final. Lo que nombra existe: el objetivo `evals-sondeo` del `Makefile`, las evals 04, 19 y 20 de `boe-legislacion` y `claude-sonnet-5-5`, el `MODELO_DE_EVALS` de `.github/workflows/evals.yml`. |
| §9. El job de cierre | **No se ejecuta.** Lee `gates/evals/*.json`, que deja el cierre del workflow después del run (FR-102). |

## 4. Lo que el hito no cambia, y los umbrales

Todo sale de `git diff --name-status main` sobre `95be315`.

- **No hay ningún ADR nuevo** (FR-090): `git diff --name-status main -- docs` no imprime nada, y `docs/ADR/` tiene los
  mismos 31 ficheros en `HEAD` y en `main`. El hito aplica sin reabrirlos el ADR 0016 (regla por serie, tres
  repeticiones, modelo informativo), el contrato de `umbrales` del ADR 0029 y el modelo que decide del ADR 0031.
- **Los artefactos de H7.1, H7.2 y H7.3 no cambian** (FR-090): `git diff --quiet main -- specs/011-h7-1-graph-check-acotado
  specs/012-h7-2-la-consulta-repetida specs/013-h7-3-el-umbral-de` sale con 0, sobre 172 ficheros versionados. Dentro
  están los tres informes que lee `expresiones-calibradas`, `gates/evals/boe-legislacion.json` de cada uno, que tampoco
  cambian.
- **Los guiones del workflow no cambian** (FR-072): `git diff --name-status main -- scripts/workflow .specify
  scripts/hito.sh scripts/paso.sh` no imprime nada, sobre 59 ficheros versionados. La tanda única se consigue en el job
  (`.github/workflows/evals.yml`) y en `internal/evals`.
- **El binario no cambia en su código.** `git diff --name-status main -- cmd internal ':!internal/evals'` solo nombra
  `M internal/app/grafo_test.go`, el test de derivadas de la aplicación, sobre 460 ficheros versionados. `go list -deps
  ./cmd/kitlegal` nombra 17 paquetes del módulo, y ninguno es `internal/evals` (`grep -c` da `0`). Lo único del producto
  que cambia es lo que el binario empotra: `skills/boe-legislacion/SKILL.md`, la skill v0.1.4, que es la entrega del
  hito (plan.md, «Structure Decision»). Sus `references/` no cambian.
- **La tabla de fuentes no cambia** (ninguna fuente nueva): `git diff --name-status main -- docs/SOURCES.md evidencias
  data web README.md` no imprime nada.
- **`skills/boe-legislacion/SKILL.md` tiene 295 líneas**, menos de 300 (FR-026; SC-011). También la comprueba
  `TestSkillsDelRepositorio` en `skills-check`, que falla con 300 o más o con la `description` por encima de 1024
  caracteres.
- **Ningún umbral de FR-040 a FR-043 se ha rebajado, retirado ni dejado sin decidir** (FR-047):
  - `umbralDeExpresionesProhibidas = 0.05` (`internal/evals/umbrales.go:16`) es el mismo valor que en `main`
    (`umbrales.go:13` allí). A su lado, T006 añade `umbralDeRespuestasSinActivar = 0` y
    `umbralDeRespuestasConRedaccionNoLeida = 0`. `git log -G 'umbralDe[A-Za-z]+ += ' main..HEAD` solo nombra T006, y
    ningún commit posterior los cambia.
  - En `umbralesDelInforme`, `expresiones_prohibidas:<modelo>` decide en el modelo que decide
    (`delModelo.modelo == e.ModeloQueDecide`). Detrás de él, y solo en ese modelo, van `sin_activar:<modelo>` y
    `redaccion_no_leida:<modelo>`, con `decide` `true`. El de la duración sigue con `Decide: true`.
  - El total son las respuestas del modelo que decide en las evals que activan la skill, sin sacar las informativas.
    Solo quedan fuera la prueba de red, las sesiones sin medir y, desde T006, las que no terminaron (FR-045). Por eso los
    casos son «3 de 54» y «1 de 54».
  - La definición del job no mueve nada que decida: `git log -G 'objetivo_de_duracion|MODELO_DE_EVALS:|REPETICIONES_DE_EVALS:|UMBRAL_DE_EVALS:'
    main..HEAD -- .github/workflows/evals.yml` no nombra ningún commit. Siguen `objetivo_de_duracion: 900` en
    `boe-legislacion`, `MODELO_DE_EVALS: claude-sonnet-5-5`, `REPETICIONES_DE_EVALS: 3` y `UMBRAL_DE_EVALS: 2` (la regla
    por serie, FR-062). Fuera de los comentarios, el diff de `evals.yml` solo añade el trabajo `tanda`, `needs: [tanda]` y
    el `if` de `evals`, y sube el tope de 120 a 122 minutos (FR-054).
  - `TestUmbralesDelInforme` los ve fallar: `sin-activar-una`, `redaccion-no-leida-una`, `tres-de-54`,
    `tres-de-54-y-seis-sin-terminar` y `901-s-con-objetivo-900` dan `fallo` (§3).
- **La lista no ha perdido ninguna expresión** (FR-047). Comparadas familia por familia, `main` tiene `maquinaria` 22,
  `otra_conversacion` 16 y `anuncio` 14 (52), y `HEAD` tiene 22, 16, 39 y `redaccion_no_leida` 10 (87), más 2 formas
  fijas. Las 52 de `main` siguen en su familia y en su posición: 0 que falten o cambien de sitio.
- **plan.md tiene las cuatro filas del cierre de FR-092** en «Controles de umbral»:
  - `FR-040, SC-001` (≤ 5 %, ≤ 2 de 54) en `evals:boe-legislacion:expresiones_prohibidas:claude-sonnet-5-5`;
  - `FR-041, SC-001` (0) en `evals:boe-legislacion:sin_activar:claude-sonnet-5-5`;
  - `FR-042, SC-001` (0) en `evals:boe-legislacion:redaccion_no_leida:claude-sonnet-5-5`;
  - `FR-043, SC-001` (≤ 900 s) en `evals:boe-legislacion:duracion_de_las_sesiones`.

  El umbral del modelo informativo (FR-044) no tiene fila, y el texto de la sección lo dice. Las demás filas son los
  umbrales que se miden en `make ci`, cada una con su test (`ci:…`).
- **SC-011**: `make ci` en verde, con `schema-check` y `skills-check` sin drift y con las reglas del conjunto; la skill
  con 295 líneas; y la entrada de `boe-legislacion` v0.1.4 en `CHANGELOG.md`, bajo *Unreleased* (T012).
