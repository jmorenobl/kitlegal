<!-- Propuesta de cambio de H5.1. La escribió la tarea de cierre (T013, intento 1) con la medida hecha sobre 85a2cf4 (`feat(H5.1): T012`) el 2026-09-16; la tarea de plataforma (T014) registra la ejecución de aceptación en gates/evals-aceptacion.md. -->

## Objetivo

«Que las evals midan lo que hoy no comprueba nadie: que la skill **traslada a su respuesta** los avisos de vigencia que
emite el binario» (`docs/ROADMAP.md` §4, H5.1). Entra desde la bitácora `docs/USO.md`, entrada del 2026-09-16
«Ninguna eval comprueba qué hace la skill ante una norma derogada» (ADR 0013, §6), y va antes de H6 para que el formato
de eval deje de crecer justo cuando un hito empieza a apoyarse en él.

H5.1 no entrega skill nueva: **protege `boe-legislacion`** (constitución, principio VIII). El mecanismo lo fija el hito:
el aviso se compara **por su forma fija —`⚠`, la etiqueta del aviso tal como la da el binario y dos puntos—, como la
cita por su identificador**, sin leer la redacción libre de la respuesta y sin ningún modelo. Cuatro piezas:

1. **Forma fija en la skill.** `SKILL.md` (de 181 a 199 líneas) enseña a trasladar cada aviso del sobre con su forma
   fija: `⚠ NORMA DEROGADA:`, `⚠ VIGENCIA AGOTADA:` y `⚠ TEXTO POSIBLEMENTE DESACTUALIZADO:`, con la frase del binario
   o una explicación detrás (FR-001 a FR-005, SC-001). La etiqueta de cada código la exporta `internal/source/boe`, su
   única fuente de verdad, sin cambiar un byte de lo que emite el applet (FR-010 a FR-012).
2. **`avisos` en el formato común de eval**: los códigos de aviso que la respuesta tiene que llevar, solo en una eval
   que activa la skill y solo con los códigos del binario (FR-020 a FR-023, SC-002).
3. **Juicio e informe.** `Juzgar` reparte los avisos esperados en encontrados y ausentes, un aviso ausente impide pasar
   y el informe publica los dos repartos y sus motivos (FR-030 a FR-043, SC-003, SC-004). Otra redacción o la negación
   dejan el aviso ausente.
4. **Una eval sobre una norma derogada**: `18-lrjpac-norma-derogada.yaml`, el artículo 42 de la Ley 30/1992
   (`BOE-A-1992-26318`), informativa (ADR 0016), que exige los avisos `derogada` y `vigencia-agotada` (FR-050 a FR-057,
   SC-006). Se confirmó antes del primer cambio de `SKILL.md` (SC-009).

## Alcance

Frente a `main` (`2d2efb8`), en la cabeza medida `85a2cf4`: fuera de `specs/`, **24 ficheros, 1 323 líneas añadidas y
65 retiradas**; en `specs/007-h5-1-avisos-de-vigencia/`, los artefactos del hito, que siguen cambiando con lo que
registran el cierre y la plataforma. Por árboles:

- **`internal/source/boe`** (2): `avisos.go` compone cada frase con su etiqueta y exporta `EtiquetasDeAviso()`, un mapa
  nuevo en cada llamada de código a etiqueta; `avisos_test.go` gana `TestEtiquetasDeAviso`. Ningún otro fichero del
  paquete, ningún golden ni esquema de salida cambia.
- **`internal/evals`** (13): nuevo `avisos.go` (`ExtraerAvisos`, `ComprobarFormasDeAviso`, `ComprobarCodigosDeAviso`) y
  su test; `formato.go` (`Eval.Avisos`), `juzgar.go` (`AvisosEncontrados`, `AvisosAusentes`, el motivo
  `aviso ausente: <código>` y `Pasa`) e `informe.go` (dos celdas en la tabla de sesiones), con sus tests;
  `conjunto_test.go` (subtests `avisos-del-esquema` y `avisos-de-la-skill` de `TestEvalsDelRepositorio`);
  `grabaciones_test.go` (premisa de `otraNormaDeLaBusqueda`) y `grabacion_test.go` (etiqueta `grabacion`: el arnés
  siembra desde la unión de las grabaciones de H4 y de H5).
- **`skills/boe-legislacion/`** (2): `SKILL.md` con exactamente los tres cambios del contrato de la forma fija §3 (el
  último punto del paso 5, el párrafo y la lista de formas de «Cómo se cita» y la regla 3), región generada intacta; y
  `references/normas.md`, regenerado con `make skills-sync`.
- **Datos y evals** (2): `data/normas.yaml` gana `BOE-A-1992-26318` (título y rango de la búsqueda grabada,
  `abreviatura: LRJPAC`, sin ninguna marca de derogación); `evals/boe-legislacion/18-lrjpac-norma-derogada.yaml`, nueva.
- **Material protegido** (3), en tareas `[datos]` con pausa: `schemas/eval.yaml.json` (propiedad `avisos`, definición
  `codigo-de-aviso` y la tercera alternativa de la rama `else`; T003); `testdata/evals/grabaciones.json` (la entrada de
  la Ley 30/1992; T008); y el índice grabado
  `testdata/evals/boe.legislacion-consolidada/GET_…_BOE-A-1992-26318_texto_indice.json`, que grabó una persona en la
  pausa de T008 y confirmó en su propio commit (`cc1362d`).
- **Documentación** (3): `CHANGELOG.md` (bloque *De H5.1* en «Añadido» de *Unreleased*), `README.md` y
  `CONTRIBUTING.md` (fila `avisos` de la tabla del formato común de eval y, en `CONTRIBUTING.md`, la frase de cuándo
  pasa una eval).
- **`.golangci.yml`** (1): solo la palabra `variantes` en `misspell.ignore-rules`, con su motivo (el subtest
  `aviso-con-variantes-toleradas`, cuyo nombre fija el plan).

**Sin cambios**, como exige el spec: `internal/core` (`git diff --quiet main...HEAD -- internal/core`, código 0),
`internal/app`, `internal/cli`, `internal/httpx`, `internal/cache`, `cmd/`, `scripts/`, `.github/`, el `Makefile`,
`codecov.yml`, `go.mod`, `go.sum`, `docs/SOURCES.md`, `docs/USO.md`, los esquemas de salida `schemas/norma.json` y
`schemas/bloque.json`, los golden y los guiones `testscript`.

**Fuera de alcance** (spec, *Fuera de alcance*): promover la eval 18 a decisoria; marcar la derogación en
`data/normas.yaml` o en `references/`; extender los avisos a otras skills o fuentes; juzgar la redacción libre (listas
de expresiones, atribución a normas, similitud o modelo), incluido el caso de una respuesta con la forma fija que además
dice lo contrario, limitación declarada y visible en el informe, que publica la respuesta; una eval para
`consolidacion-no-finalizada`; un recuento por aviso en el informe; cambios en el job de evals; ADR nuevo. Puntos 7, 8,
11 y 12 de la Definition of Done: no aplican; el 6, solo `CHANGELOG.md`, porque el comportamiento visible del binario
no cambia.

## Dependencias (constitución §V)

**Dependencias: ninguna nueva.** `go.mod` y `go.sum` no aparecen en el diff frente a `main`. `internal/evals` ya
importaba `santhosh-tekuri/jsonschema/v6`; lo demás es biblioteca estándar (`regexp`, `sync`, `errors`, `slices`,
`strings`, `os`). Sin cambio de versión en ninguna herramienta de control, ni en el job de evals (Claude Code, `strace`,
`ubuntu-24.04`, ADR 0016).

## Controles añadidos

Los 17 controles de la tabla del plan («Controles mecánicos que este hito añade o toca») están en el árbol: 16 dentro de
`make ci` y el último, la aceptación con modelo, fuera de él por diseño. Lo que pasa a ser mecánico, con el escenario
del quickstart que lo demuestra:

- **Etiquetas con una sola fuente de verdad** (`TestEtiquetasDeAviso`: `exactamente-tres`, `frases-con-su-forma`,
  `cada-llamada-su-mapa`), con la salida intacta: `TestAvisosDe`, `TestCodigosDeAviso`, los golden, `schema-check` y
  `make test-e2e` sin cambios y en verde (escenario 3).
- **Forma fija con su tolerancia exacta** (`TestExtraerAvisos`, 30 subtests del contrato de la forma fija §4): tolera
  el selector de presentación del emoji, el énfasis con `*` y `_`, los blancos horizontales de Unicode y las mayúsculas;
  no tolera una palabra de más, de menos, distinta o pegada, otros dos puntos, un salto dentro, el tachado, otro emoji,
  dos selectores, otra redacción ni la negación (escenario 5).
- **`avisos` en el formato** (`TestLeerEval`, casos `avisos`, `aviso-desconocido`, `no-activa-con-avisos`,
  `avisos-vacio`, `citas-vacio`, `aviso-repetido`, `cita-repetida`, `informativa-con-avisos`): un código desconocido o
  `avisos` en una eval de no activación es un fichero mal formado (escenario 6).
- **Esquema ↔ `CodigosDeAviso()`** (FR-013; `TestComprobarCodigosDeAviso` con `exacto`, `falta-un-codigo`,
  `sobra-un-codigo`, `sin-enumerado`, `sin-avisos` y `referencia-circular`, y `TestEvalsDelRepositorio/avisos-del-esquema`
  dentro de `make skills-check`): un código de menos falla con «el esquema de eval no enumera el código de aviso
  derogada en avisos» y uno de más con «el esquema de eval enumera en avisos el código otro-aviso, que no es un código
  de aviso del binario» (escenario 4).
- **`SKILL.md` ↔ forma fija de cada código** (FR-014; `TestComprobarFormasDeAviso` y
  `TestEvalsDelRepositorio/avisos-de-la-skill`): sin la etiqueta, sin la marca o sin los dos puntos, falla con «falta la
  forma fija del aviso derogada: ⚠ NORMA DEROGADA:» (escenario 4).
- **Sin copias de etiquetas en `internal/evals`** (FR-011; `TestEtiquetasSoloDesdeBoe`): ningún fichero de código del
  paquete escribe una etiqueta, tampoco en un comentario (escenario 9).
- **Juicio de avisos** (`TestJuzgar`, diez casos nuevos: `aviso-con-su-forma-fija`, `aviso-con-variantes-toleradas`,
  `aviso-ausente`, `aviso-con-otra-redaccion`, `aviso-negado`, `forma-fija-y-lo-contrario`, `aviso-no-esperado`,
  `avisos-en-el-orden-de-la-eval`, `aviso-repetido`, `cita-y-aviso-ausentes`); las evals sin `avisos` se juzgan igual que
  en H5, con las dos listas nulas (escenario 5).
- **Informe con avisos** (`TestInformeConAvisos`: `uno-encontrado-y-otro-ausente`, `aviso-detras-de-la-cita`, sobre una
  copia en un temporal del caso `aprobado`; `TestInforme/aprobado` con `[]` en las dos claves de cada sesión): un rojo por
  un aviso ausente se distingue de uno por una cita ausente (escenario 7).
- **La eval 18 en el conjunto y servida sin red** (`TestEvalsDelRepositorio/formato`, `/conjunto`, `/normas-conocidas`,
  `/grabado`; `TestIdentificadoresDeLasNormas` con la premisa corregida y sus cuatro negativos): sin el índice grabado,
  falla con «18-lrjpac-norma-derogada.yaml: la norma BOE-A-1992-26318 sin indice», primero con código 1 al preparar y
  después con código 4 con `--offline` (escenario 8).
- **Lint del arnés con etiqueta**: `run.build-tags` sigue incluyendo `grabacion`, así que el cambio de
  `grabacion_test.go` pasa por `make lint`.
- Los de H0-H5 (R1-R5, formato, `-race`, `govulncheck`, `gosec`, `gitleaks`, `go mod verify`, `tidy -diff`,
  `schema-check`, `skills-check`, `TestDependenciasDelBinario`, `TestArquitectura`) siguen sin exclusiones nuevas.
- **Aceptación con modelo** (fuera de `make ci`): la primera ejecución `evals` de la rama con evento `pull_request`, la
  que dispara la apertura de esta propuesta de cambio, comprobada con el programa de `jq` de quickstart §11.5; la
  registra T014 en `gates/evals-aceptacion.md`.

## Evidencia

Medida por T013 (intento 1) el **2026-09-16, entre las 22:38 y las 22:43 (hora de Madrid), sobre `85a2cf4`**
(`feat(H5.1): T012`, la cabeza de la rama), ejecutando en una sesión desatendida y en primer plano `make ci` y después
`quickstart.md` desde los prerrequisitos hasta la limpieza (escenarios 1 a 10), con cada orden tal cual y en orden y
las formas `rtk proxy` de su tabla. Ningún resultado distinto del esperado.

| Escenario | Resultado |
|---|---|
| `make ci` | **`código 0`** en 45 s (22:38:39-22:39:24): `0 issues.`; los once paquetes con tests en `ok` en los dos perfiles (`-race -shuffle=on`, ninguno de la caché de `go test`; `-race -tags=integration`, con `internal/app` e `internal/skills` ejecutados y los demás de la caché); `govulncheck` «No vulnerabilities found.» y «Your code is affected by 0 vulnerabilities.» (ver *Pendientes*); `schema-check` y `skills-check` en `ok`; `gitleaks` «no leaks found»; `all modules verified` en la raíz y en `tools/gitleaks`, `tools/golangci-lint`, `tools/govulncheck` y `tools/lefthook`; `go mod tidy -diff` sin salida; `ci: todos los controles en verde` |
| Prerrequisitos | `go version go1.27.1 darwin/arm64`; rama `007-h5-1-avisos-de-vigencia`; solo `fin del estado`; clon creado sin mensajes y solo `fin del estado del clon` |
| 1 · controles | `make skills-check` en `ok` para `internal/app`, `internal/skills` e `internal/evals`; `make schema-check` en `ok`; `go test -count=1` de `internal/source/boe` e `internal/evals` en `ok` |
| 2 · `SKILL.md` | **199 líneas** (181 en `main`); `⚠ NORMA DEROGADA:` en las líneas 147 y 152, `⚠ VIGENCIA AGOTADA:` en la 148 y `⚠ TEXTO POSIBLEMENTE DESACTUALIZADO:` en la 149, entre `## Cómo se cita` (121) y `## Comandos` (158); `### 5. Responder citando` en la 102 y `## Reglas` en la 182; el diff, tres bloques `@@ -114 +114,3 @@`, `@@ -140,0 +143,15 @@` y `@@ -174,3 +191,4 @@`. Comprobado además byte a byte: los dos textos quitados del contrato de la forma fija §3 están en `SKILL.md` de `main` y no en la cabeza, y los tres añadidos (3, 15 y 4 líneas) en la cabeza y no en `main` |
| 3 · salida del binario | Primer diff vacío (solo `fin del diff`); en `internal/source/boe`, solo `avisos.go` (30 líneas añadidas y 7 retiradas) y `avisos_test.go` (55 añadidas); `make test-e2e` en `ok` |
| 4 · comprobaciones mecánicas | Cada `diff --stat` del esquema, `1 file changed, 1 insertion(+), 1 deletion(-)`, y cada uno de `SKILL.md`, `1 file changed, 2 insertions(+), 2 deletions(-)`; las cinco pruebas en `FAIL` con `código 1` y su mensaje exacto (arriba, *Controles añadidos*); el clon, limpio al final |
| 5 · forma fija y juicio | Todo `PASS`, ningún `FAIL`: `TestExtraerAvisos` con sus 30 subtests, exactamente los del contrato de la forma fija §4; `TestJuzgar` con 32, los diez nuevos entre ellos |
| 6 · formato | Todo `PASS`: `TestLeerEval` con 22 subtests, los ocho nuevos entre ellos; `TestComprobarCodigosDeAviso` con los cinco que nombra la guía y además `referencia-circular` (T002: un `$ref` cíclico sin enumerado falla nombrando los tres códigos), que no contradice lo esperado |
| 7 · informe | Todo `PASS`: `TestInformeConAvisos/uno-encontrado-y-otro-ausente`, `/aviso-detras-de-la-cita` y `TestInforme` con 20, `aprobado` entre ellos |
| 8 · eval, norma y grabación | 18 ficheros de eval, el último `18-lrjpac-norma-derogada.yaml`, idéntico byte a byte al bloque del contrato de la eval y la grabación §6; `BOE-A-1992-26318` en `data/normas.yaml:69` y `references/normas.md:9`; `data/normas.yaml:0` y `skills/boe-legislacion/references/normas.md:0` seguidos de `fin del recuento`; el diff de material de prueba, exactamente `A testdata/evals/boe.legislacion-consolidada/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-1992-26318_texto_indice.json` y `M testdata/evals/grabaciones.json`; «la eval (0a03a116…) precede al primer cambio de SKILL.md (394e1388…)»; todo `PASS` en los cuatro tests, con `formato`, `conjunto`, `normas-conocidas`, `grabado`, `avisos-del-esquema`, `avisos-de-la-skill` y los cuatro negativos de `TestIdentificadoresDeLasNormas`; sin el índice, `FAIL` con «18-lrjpac-norma-derogada.yaml: la norma BOE-A-1992-26318 sin indice» y `código 1`; el clon, limpio |
| 9 · etiquetas solo del binario | Solo `fin de la búsqueda`; todo `PASS`, con `exactamente-tres`, `frases-con-su-forma` y `cada-llamada-su-mapa` en `TestEtiquetasDeAviso`, y `TestAvisosDe`, `TestCodigosDeAviso` y `TestEtiquetasSoloDesdeBoe` |
| 10 · documentación | La fila `avisos` en `README.md:210` y `CONTRIBUTING.md:227`; la frase de cuándo pasa una eval en `CONTRIBUTING.md:237-239`; `*De H5.1 — los avisos de vigencia en las evals:*` en `CHANGELOG.md:250`, dentro de «Añadido» de *Unreleased*, con la forma fija en la skill, el campo `avisos`, el reparto en el informe y la eval nueva |
| Limpieza | La carpeta desaparece y el estado del árbol da solo `fin del estado` |

**La skill y el conjunto de evals**, contados sobre `85a2cf4`: **199 líneas de `SKILL.md`** (`wc -l`; 181 en `main`,
por debajo de 300); **18 evals con 10 que deciden** en `evals/boe-legislacion/`: las 10 positivas que deciden, 6
informativas (13 a 18) y 2 de no activación; `TestEvalsDelRepositorio/conjunto` en verde. Con la eval 18,
el job pasa de 87 a 90 sesiones (research V21).

**Cobertura**, con `go tool cover -func` sobre el `coverage.out` y el `coverage-integration.out` que dejó ese
`make ci` sobre `85a2cf4` (y, por árbol, la suma de sentencias del perfil contando cada bloque una vez):

| Umbral | Exigido | Perfil unitario (`make test`) | Perfil de integración |
|---|---|---|---|
| Global (`codecov/project`) | ≥ 70 % | **96,9 %** (6003/6194) | **97,4 %** (6032/6194) |
| `internal/core/**` | ≥ 85 % | **90,1 %** (73/81) | 90,1 % |
| `internal/cli/**` | ≥ 90 % | **98,6 %** (348/353) | 98,6 % |
| `internal/evals` | — | 99,0 % (1753/1770) | 99,0 % |
| `internal/source/boe` | — | 99,4 % (876/881) | 99,4 % |

`internal/core` no cambia respecto de `main` (diff vacío) y conserva sus 81 sentencias. **Todo lo que añade H5.1 está
cubierto**: `internal/evals/avisos.go` 67/67, `internal/evals/juzgar.go` 139/139 e `internal/source/boe/avisos.go`
20/20; en `formato.go` (17/18) e `informe.go` (524/528) los bloques sin cubrir (`formato.go:119` e `informe.go:266`,
`270`, `298` y `699`) son de H5 y quedan fuera de las líneas del diff. **Ningún umbral se rebaja**: `codecov.yml` no
aparece en el diff.

**Sin ninguna supresión nueva**: `0` líneas `//nolint` y `0` `t.Skip` añadidas en ficheros `.go` frente a `main`.

**Sin red en las tareas**: ninguna ejecutó `KITLEGAL_RECORD`, `scripts/grabar-evals.sh`, `make evals` ni
`make verify-sources`. El esquema y el manifiesto los revisó una persona en las pausas de T003 y T008, y el índice lo
grabó una persona en la de T008.

## Decisiones

- **La forma fija, no la redacción** (*Decisión del mecanismo* del hito, 2026-09-16): el aviso se compara por su forma
  fija, como la cita por su identificador. Descartados las listas de expresiones con reglas de atribución a normas (el
  primer run del hito: no escala) y cualquier juicio semántico (capa 1 de la constitución; la similitud no separa «ha
  sido derogada» de «no ha sido derogada»).
- **Etiquetas en el binario, en un mapa** (research D1): cada frase se compone con su etiqueta, así que no pueden
  divergir; `EtiquetasDeAviso()` se enumera, para que un test diga «exactamente tres», y es nuevo en cada llamada, como
  `CodigosDeAviso()`.
- **La gramática de la forma** (D2): una expresión por código construida desde las etiquetas exportadas, con
  `regexp.QuoteMeta` en cada palabra y dentro de una línea; tolerante solo con lo que no cambia qué aviso es. La
  limitación declarada —la forma fija seguida de lo contrario pasa— queda visible en la respuesta publicada.
- **El reparto como el de las citas** (D3): orden de la eval y repeticiones incluidas; el motivo, detrás de los de las
  citas y antes del del modelo; lo no esperado no se mira.
- **El esquema admite `avisos` con el trato de `citas`** (D4): `minItems: 1`, sin `uniqueItems`, prohibido en la rama
  `else` y enumerado en `$defs` en el orden de `CodigosDeAviso()`.
- **Las comprobaciones leen lo que de verdad se aplica** (D5, D7): FR-013 sobre el esquema compilado, siguiendo `$ref`
  hasta un `Enum`; FR-014 con la misma función que el juicio y la forma completa (Clarifications, Q2).
- **Dentro de `TestEvalsDelRepositorio`** (D6): los dos subtests entran en `make skills-check` sin tocar el `Makefile`.
- **El informe publica los repartos junto a las citas** (D9) y su test trabaja sobre una copia en un temporal (D10),
  sin material nuevo bajo `testdata/`.
- **Una sola petición nueva al BOE** (D11): el arnés siembra desde la unión de las grabaciones de H4 y de H5, así que
  solo pide el índice de `BOE-A-1992-26318` (y el `robots.txt`, que la persona restaura), en lugar de treinta peticiones
  ya grabadas.
- **Premisa de `TestIdentificadoresDeLasNormas` corregida antes del manifiesto** (D12): la norma de
  `norma-en-la-busqueda-de-otra-entrada` es la que ninguna entrada resuelve, que es lo que el subtest enuncia.
- **La norma sin marca de derogación** (D13): el aviso lo emite el binario al consultar, no la tabla.
- **La eval 18** (D14): la pregunta nombra la norma y el artículo y no dice nada de su vigencia; nace informativa para
  no tocar la regla de 10 positivas que deciden.
- **El ejemplo de `SKILL.md` es la frase del binario** (D15), byte a byte y sin explicación añadida, para que la skill
  no enseñe ninguna afirmación que el sobre no traiga.
- **Aceptación por `jq` sobre la ejecución de apertura** (D17): la primera ejecución `evals` de la rama con evento
  `pull_request`, sin juicio de ningún modelo.
- **`variantes` en `misspell.ignore-rules`** (D19): el nombre del subtest `aviso-con-variantes-toleradas` lo fija el
  plan; es el mecanismo del repositorio para palabras españolas, no una exclusión de lint.

## Pendientes

- **La ejecución de aceptación** (T014): publicar la rama y abrir esta propuesta de cambio, esperar la ejecución de
  apertura del job `evals`, leer su informe y comprobar con quickstart §11.5 el veredicto `aprobado`, `red` vacío, la
  tasa de la eval 18 con el modelo que decide sobre 3 sesiones y el reparto de sus dos avisos en cada sesión; la
  evidencia va a `gates/evals-aceptacion.md`. Con ella se comprueban los supuestos de plataforma S1 a S4 (research §S):
  la ejecución de apertura y su job, el `headSha` igual al `commit` del informe, el formato del registro y que las 90
  sesiones caben en el tope del job.
- **Regla de repetición de la aceptación** (contrato de la ejecución de aceptación §6): si después de la ejecución de
  aceptación entra en la rama un commit que cambia cualquier fichero fuera de `specs/007-h5-1-avisos-de-vigencia/` —por
  ejemplo, una corrección de la revisión final—, la aceptación ya no cubre la cabeza y **se repite**: se quita la
  etiqueta `evals` si está puesta, se pone, y se identifica, espera, lee y comprueba la ejecución nueva por el último
  evento `labeled` (quickstart §11.6), con las mismas condiciones; su evidencia se añade como sección nueva y vigente de
  `gates/evals-aceptacion.md`. Quien cambie algo fuera del directorio del hito tras la aceptación lo deja anotado aquí.
- **`govulncheck` y un módulo requerido**: el mismo `make ci` informa, sin fallar, de una vulnerabilidad en un módulo
  que el código no llama, `GO-2026-5970` («Infinite loop on invalid input in golang.org/x/text»; encontrada en
  `golang.org/x/text@v0.14.0`, corregida en v0.39.0), según `govulncheck -show verbose` sobre `85a2cf4`. H5.1 no toca
  `go.mod` ni `go.sum`, así que viene de `main`; subir la versión queda fuera de este hito.
- **Integración continua y Codecov**: leer los estados de la propuesta de cambio antes de fusionar; la rama principal no
  impide fusionar en rojo. Fusionar es humano (ADR 0007).
- **Backlog, fuera de este hito**: promover la eval 18 a decisoria con los datos de varias ejecuciones (enmienda de
  FR-062); una eval para `consolidacion-no-finalizada`; extender los avisos a otras skills o fuentes.
