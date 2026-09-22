# Implementation Plan: H6 · `territorio` + skill `legal-core` v0

**Branch**: `008-h6-territorio-skill-legal` | **Date**: 2026-09-20 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/008-h6-territorio-skill-legal/spec.md`

**Modo**: desatendido. Las decisiones técnicas se tomaron con el «Criterio de decisión autónoma» de
`.specify/memory/constitution.md` y están en [research.md](./research.md) (D1-D29) con su alternativa rechazada y su
motivo. Toda afirmación sobre una herramienta o dependencia externa remite a la tabla de verificación de
[research.md](./research.md) (V1-V47), con la orden o el `fichero:línea` de la comprobación; lo que no se puede
comprobar sin red, sin la plataforma o sin el fichero que genera la tarea `[datos]` está declarado como supuesto
(S1-S9) y **no se afirma como hecho** en ningún punto de este plan.

## Summary

H6 entrega la pieza que hace genéricas a las demás skills: el applet `territorio` con el verbo `resolver`, que a
partir de un nombre de municipio o de un código INE devuelve municipio, código INE, provincia, comunidad autónoma,
DIR3 del ayuntamiento, régimen, boletines aplicables y `cobertura`, **sin pedir nada por red** (ADR 0017); y la skill
madre `legal-core` v0, cuyo protocolo empieza por identificar el territorio y que razona con dos referencias
generadas desde `data/`. Con el applet entran los dos primeros identificadores del proyecto, código INE y DIR3, en
`internal/core/ids`.

Decisiones que sostienen el diseño:

1. **Los datos congelados viven en `data/territorio/` y viajan dentro del binario** (D2, D5): cuatro ficheros YAML con
   una línea por fila, embebidos por un paquete `data` mínimo, porque `//go:embed` no sube de directorio (V2). Un
   fichero ausente es un error de compilación; uno inválido lo caza `make ci`. No hay modo de fallo por datos
   ausentes, de ahí que el applet nunca devuelva 4, 5 ni 6.
2. **El dominio recibe bytes, no un sistema de ficheros** (D3): `io/fs` está denegado a `internal/core` y la
   prohibición es por componente de ruta (V5, V6, V7). `territorio.Cargar(Fuentes)` es puro y comprobable con datos
   sintéticos.
3. **La fecha del sobre es la del fichero congelado más antiguo que sostiene la respuesta** (D6). Es la regla de
   FR-096 de H4 aplicada a un dato congelado (ADR 0015) y, de paso, lo que hace que la salida sea byte a byte la
   misma en dos ejecuciones y con `--offline`, como pide SC-001.
4. **El dígito de control es un dato oficial, no un algoritmo** (D9): viaja en la relación del INE, así que la
   comprobación es correcta por construcción y no hace falta ninguna hipótesis.
5. **La coincidencia por nombre usa un pliegue propio verificado contra el corpus** (D10), sin promover
   `golang.org/x/text` a dependencia directa: tres controles mecánicos exigen que el pliegue cubra toda runa del
   corpus, que no deje ningún municipio inalcanzable y que ningún nombre plegado sea solo cifras. Por leer los
   ficheros congelados reales, esos tres **viven en `internal/skills`** y no en el dominio, que tiene denegada la
   entrada y salida también en sus `_test.go` (D27, V7).
6. **`cobertura` tiene tres claves y un vocabulario cerrado sin ningún valor que signifique «no existe»** (D11, D12):
   es lo que hace mecánicamente cierto que ninguna skill pueda concluir «no hay boletín».
7. **El formato común de eval se extiende, no se duplica** (D21, D22): cuarta variante de comando, esperado de
   territorio comparado por forma fija, regla del esperado verificable, y reglas de conjunto parametrizadas por skill
   en lugar de copiadas.
8. **Registrar el applet y publicar su esquema van juntos** (D16): en cuanto `territorio resolver` está en el
   registro, el control de cobertura de esquemas exige su parte publicada, y el esquema no existe antes que el applet
   (V8).

Artefactos de diseño: [data-model.md](./data-model.md) y [contracts/](./contracts/)
([applet-territorio](./contracts/applet-territorio.md),
[identificadores-ine-y-dir3](./contracts/identificadores-ine-y-dir3.md),
[datos-de-territorio](./contracts/datos-de-territorio.md),
[skill-legal-core](./contracts/skill-legal-core.md),
[evals-de-legal-core](./contracts/evals-de-legal-core.md)); validación en [quickstart.md](./quickstart.md).

## Technical Context

**Language/Version**: Go 1.27 sin cambios (`go 1.27.0` + `toolchain go1.27.1`, que el `Makefile` lee y exporta; V1).
YAML y JSON Schema 2020-12 para datos y esquemas; Markdown para la skill; Bash para los guiones ya existentes.

**Primary Dependencies**: las de `go.mod`, sin ninguna nueva. `go.yaml.in/yaml/v3` (§V, directa desde H5) pasa a estar
**enlazada en el binario**, porque el dominio analiza los ficheros congelados: eso obliga a añadirla a
`modulosDelBinario` con su motivo (V11; *Complexity Tracking*). `santhosh-tekuri/jsonschema/v6` e `invopop/jsonschema`
se siguen usando solo desde tests y desde `--describe`, como hasta ahora. Biblioteca estándar nueva en el árbol:
`embed` (paquete `data`), `strings`, `sort`/`slices`, `unicode`.

**Herramientas de control**: las de H0-H5, sin cambio de versión y **sin ninguna exclusión nueva de lint**.
`.golangci.yml` no se toca: de las 124 palabras del vocabulario de H6, `misspell` solo marca `aspectos`,
`configuracion` y `autonomos`, y ninguna hace falta suelta en Go (V4, D24).

**Storage**: ninguno nuevo. `territorio` no usa la caché de H3 ni abre ninguna base de datos. Los datos congelados
viajan dentro del binario (`//go:embed`).

**Testing**: `make test` (unitarios y e2e con `-race`), `make test-integration` (sin cambios), `make schema-check`
(gana `schemas/municipio.json`), `make skills-check` (gana la validación de `data/territorio/` y las reglas del
conjunto de `legal-core`). Todo **sin red**: el applet no la usa, los tests del dominio usan datos sintéticos y los de
repositorio, los ficheros congelados. La red de la fuente solo sigue en `scripts/grabar-evals.sh` (una persona, en la
pausa) y en `make verify-sources`, que **no gana ningún caso** (FR-049, V33).

**Target Platform**: sin cambios (`CGO_ENABLED=0`, `-trimpath`); desarrollo en darwin/arm64, `make ci` en
`ubuntu-latest`, job de evals en `ubuntu-24.04`.

**Project Type**: CLI multicall + skills del estándar Agent Skills + datos congelados versionados.

**Performance Goals**: `territorio resolver` por debajo de 200 ms, incluido el análisis de los ficheros embebidos; lo
mide el e2e con la orden `cronometra` (V35). Estimación del coste de análisis, no medida todavía: S3.

**Constraints**: ninguna petición de red en ejecución ni en grabación para INE, REL o DIR3 (ADR 0017); ningún caso
especial para un municipio en código, `data/` ni skills (principio IX); `SKILL.md` < 300 líneas; generación
determinista e idempotente; sin cambios en `internal/httpx`, `internal/cache`, `internal/source/boe` (salvo nada) ni
en el protocolo, las referencias y las evals de `boe-legislacion` más allá de regenerar `references/normas.md`.

**Scale/Scope**: 1 applet con 1 verbo · 2 paquetes de dominio nuevos · 1 paquete `data` de un solo fichero
(`data/datos.go`) · 4 clases de fichero congelado —uno de ~8.100 filas, uno de correspondencia, uno de estado y uno
por comunidad—, que son **22 ficheros** contando las 19 comunidades · **6 esquemas nuevos** (los cuatro de
`data/territorio/`, el de la jerarquía y el contrato publicado del applet) y 2 modificados · 1 skill (4 ficheros) ·
3 evals · 2 objetivos de fuzz con su corpus · 1 guion e2e nuevo y 2 tocados · documentación.

## Constitution Check

*GATE: debe pasar antes de la fase 0 y volver a evaluarse tras la fase 1.*

### Principios

| # | Principio | Cómo lo cumple H6 | Veredicto |
|---|---|---|---|
| **I** | Fuentes públicas y frontera humana | H6 **no añade ninguna fuente de red**: `territorio` no abre conexiones en ningún camino (FR-043) y no hay base nueva en `internal/httpx` ni fila de fuente consultable en ejecución. INE, REL y DIR3 entran congelados por tarea `[datos]` con pausa (ADR 0017), descargados por una persona **fuera del repositorio**; el ejecutor desatendido no descarga nada ni usa `KITLEGAL_RECORD`. Ningún intento de sortear WAF, CAPTCHA, certificado ni Red SARA (ADR 0017, decisión 4). `SKILL.md` de `legal-core` prohíbe toda acción con identidad (FR-063) | ✅ Cumple |
| **II** | Nada sin cita ni fuente | Todo sale en el sobre de seis claves con la procedencia del espacio reservado (ADR 0006, V36) y con la fecha del fichero congelado más antiguo que lo sostiene (D6), de modo que la cita nunca aparenta más frescura que su parte más vieja. **Cada dato de `data` lleva su `source`** (FR-005), y un test exige que sea una fila de `docs/SOURCES.md` o un fichero de `data/territorio/`. Lo no configurado se declara en `cobertura` y nunca se inventa un boletín ni un DIR3 (FR-021 a FR-023). La skill no cita texto de norma: lo delega en `boe-legislacion` (FR-069) | ✅ Cumple |
| **III** | Tests primero y offline | Cada tarea escribe su test antes del código. Las evals de `legal-core` —el test de aceptación de la skill— se commitean **antes** que `SKILL.md` (FR-083), y el e2e de la matriz territorial entra en cuanto hay applet que ejercer. Todo test corre offline: el applet no tiene red, el dominio se prueba con fuentes sintéticas y el repositorio, con sus ficheros congelados. Toda salida se valida contra `schemas/municipio.json` en test (FR-092). Umbrales intactos: `internal/core/**` ≥ 85 % —que con dos paquetes nuevos **mide de verdad** (V19)— y global ≥ 70 % | ✅ Cumple |
| **IV** | Arquitectura hexagonal con reglas ejecutables | Dominio puro en `internal/core/{ids,territorio}`, que reciben bytes y no un sistema de ficheros justamente para no romper R1 (D3, V5-V7); composición en `internal/app`; presentación intacta. Errores tipados con `schema.ConClase` que el kernel traduce a 2 y 3, sin que el dominio importe el kernel (V18). Ningún `panic` en rutas de usuario: el fuzz lo comprueba. Reglas de dependencia, abajo | ✅ Cumple |
| **V** | Simplicidad y dependencias fijadas | **Ninguna dependencia nueva.** `golang.org/x/text` se rechaza expresamente y se resuelve con un pliegue propio verificado contra el corpus (D10). Sin DI, ORM ni generador de CLI; `internal/cli` no se extrae ni se refactoriza (se revisa al cerrar H9). Lo único que cambia en el perfil del binario es que `go.yaml.in/yaml/v3` —dependencia ya fijada en §V— pasa a enlazarse: va a *Complexity Tracking* | ✅ Cumple con justificación |
| **VI** | Un binario, convenciones de agente | Un applet más del mismo ejecutable, invocable por multicall y por enlace (`scripts/territorio`); hereda las ocho banderas globales sin declarar ninguna; `--describe` emite su esquema y de ahí sale la tabla de comandos de la skill (FR-007, FR-064) | ✅ Cumple |
| **VII** | Grafo y privacidad | Sin grafo: `territorio` **no emite operaciones** en este hito, que llegan en H7 (ADR 0014; spec, *Fuera de alcance*). Los ids naturales que H7 necesitará —código INE y DIR3— quedan tipados en `internal/core/ids`. Ningún dato personal: municipios, provincias y boletines son datos públicos de registros nacionales; no hay `Persona` ni NIF | ✅ Cumple |
| **VIII** | Skills primero; el binario es la herramienta | El hito entrega la skill madre `legal-core`, medida con evals, y el applet existe para ella. A Go va solo lo que exige determinismo: identificadores, resolución sobre datos congelados y cobertura; el razonamiento y el protocolo viven en `SKILL.md` y en las dos referencias generadas. `SKILL.md` < 300 líneas, `references/` generadas desde `data/`, `scripts/` como enlaces | ✅ Cumple |
| **IX** | Genericidad territorial, validación local | Ningún municipio ni comunidad como caso especial en código, en `data/` ni en la skill (FR-024): los municipios concretos solo aparecen en fixtures, e2e y evals. Los datos salen de registros nacionales; lo que varía por territorio se configura por comunidad, y **añadir un territorio es rellenar su fichero** (FR-053). Fuera del territorio configurado la salida declara su cobertura en lugar de fallar o vaciarse (FR-054), y ninguna skill puede concluir «no existe» porque el vocabulario de `cobertura` no tiene ese valor (FR-022). Matriz territorial completa en e2e: cubierto, no cubierto, foral y ambiguo (FR-090) | ✅ Cumple |

### Reglas de dependencia (`docs/ROADMAP.md` §2, constitución §IV)

| Regla | Situación en H6 | Cómo se hace cumplir | Veredicto |
|---|---|---|---|
| `internal/core/**` no importa `internal/{source,httpx,cache,store,graph,render,cli,app}` ni entrada y salida | Los dos paquetes nuevos son dominio puro: reciben bytes, devuelven valores y errores con `schema.ConClase`. **No importan `io/fs`**, que el dominio tiene denegado por colgar de `io` (V6, V7), ni el paquete `data` (D3). **La regla alcanza a sus tests**: todo test de `internal/core/**` es sintético y ninguno lee un fichero; lo que mira los ficheros congelados reales vive en `internal/skills` (D27) | `depguard` lista `core` (alcanza también a los `_test.go`, `run.tests: true`) + `TestArquitectura` R1, que recorre automáticamente todo paquete nuevo bajo `internal/core` (V5) | ✅ Cumple |
| Solo `internal/httpx` importa `net/http` | H6 no toca la red: ni el applet, ni el dominio, ni el paquete `data` la importan | `depguard` lista `red` + `TestArquitectura` R2 | ✅ Cumple |
| Solo `internal/{cache,store,graph}` importan SQLite y `database/sql` | H6 no abre ninguna base de datos | `depguard` lista `sql` + `TestArquitectura` R3 | ✅ Cumple |
| Solo `internal/cli` y `cmd/` llaman a `os.Exit` | Ningún `package main` nuevo; ni el dominio, ni el applet, ni el paquete `data` terminan el proceso | `forbidigo` `^os\.Exit$`, sin excepciones nuevas | ✅ Cumple |
| Solo `internal/render` escribe en stdout; logs con `slog` a stderr | El applet devuelve un `Resultado`; no escribe. El dominio no registra eventos y el paquete `data` solo embebe y devuelve bytes | `forbidigo` `^fmt\.Print…$`, `^os\.Stdout$`, `^os\.Stderr$`, sin excepciones nuevas | ✅ Cumple |
| `internal/graph` no importa `internal/source/*` ni `internal/render` | `internal/graph` no existe (H7) | — (sin objeto) | ✅ Cumple |
| Los applets de ejemplo no se enlazan en el binario distribuido (ADR 0010) | El binario de e2e registra `territorio` igual que registra `boe`; el distribuido no enlaza `ejemplo` | `depguard` lista `ejemplo` + `TestElBinarioNoEnlazaLosEjemplos` | ✅ Cumple |
| Ningún adaptador firma en el espacio reservado `kitlegal.` / `kitlegal:` (ADR 0006) | `territorio` **sí** firma ahí, y puede: es un applet calculado y su código vive en `internal/app`, no bajo `internal/source/` (V36) | `TestLasFuentesNoFirmanComoKitlegal`, que solo vigila `internal/source/**` | ✅ Cumple |
| El binario no enlaza módulos no justificados (FR-124 de H4) | Pasa a enlazar `go.yaml.in/yaml/v3`, dependencia de §V ya directa, porque el dominio analiza los ficheros congelados | `TestDependenciasDelBinario` con su lista escrita a mano, que gana esa línea **con su motivo** (V11); *Complexity Tracking* | ✅ Cumple con justificación |

### Gates (constitución, «Gates»)

- **Capa 1 (mecánica)**, lo que H6 añade o toca: validación de cada fichero de `data/territorio/` contra su esquema y
  de la integridad entre ficheros; salida del applet contra `schemas/municipio.json` y `make schema-check` sin
  deriva; exit codes con casos que fuerzan 2 y 3 y comprobación de que 4, 5 y 6 no ocurren; **matriz territorial en
  e2e** (cubierto, no cubierto, foral, ambiguo), que la constitución exige a toda herramienta con dimensión
  territorial; fuzz de los dos analizadores con corpus versionado; frontmatter, 300 líneas y deriva de `references/`
  y de la tabla de comandos de la skill nueva; formato y reglas del conjunto de `evals/legal-core/`; identificadores
  de las normas nuevas contra su búsqueda grabada. Guardián de diff con `[datos]` para `testdata/` y `schemas/`.
- **Capa 2 (jueces)**: que el protocolo de `SKILL.md` empiece por el territorio y delegue el texto lo juzgan los dos
  jueces de la revisión final con su rúbrica; que la `description` active con lo que debe y no con lo que no debe lo
  miden las evals del job (activación positiva y de no activación).
- **Capa 3 (humano)**, pausas previstas: los ficheros de `data/territorio/` —22, en cuatro clases— con sus cuatro
  esquemas nuevos (una pausa por tarea `[datos]`, disparada porque cada una trae su esquema: V9, D17);
  `schemas/jerarquia.yaml.json`;
  `schemas/normas.yaml.json` y `schemas/eval.yaml.json`, **ficheros existentes** (FR-085) —el de normas, en **dos**
  pausas: el campo `vertebral` en el paso 13 y el `enum` de `rango` en el paso 15, con las grabaciones que lo fijan
  (D29); el de eval, con la expectativa de test que su forma arrastra (D28)—; `schemas/municipio.json`
  publicado, en la misma pausa que los dos guiones e2e **existentes** que el paso 10 modifica (`argumentos.txtar` y
  `ayuda.txtar`: casan `(^|/)testdata/` y existen en la base, así que también los marca `clasificar_datos`); el
  manifiesto y las grabaciones de las búsquedas de las siete normas nuevas, bajo el `testdata/` de la
  raíz, donde graba la persona; la fila de `mpt.rel` de `docs/SOURCES.md` (FR-049); y la fusión. El corpus de fuzz y
  el guion e2e **nuevo** del paso 12, ficheros que no existen en la base, **no pausan** y los revisa la revisión final
  (FR-086, V9).

**Reglas del modo desatendido**: el ejecutor arregla el código, nunca el test ni el fixture; no usa `KITLEGAL_RECORD`,
`scripts/grabar-evals.sh`, `make evals` ni `make verify-sources`; no escribe ningún identificador `BOE-A-…`, título,
código INE ni DIR3 que no haya leído de una respuesta grabada o de un fichero congelado; ninguna tarea `[datos]` toca
código, con **dos** excepciones razonadas, las dos en *Complexity Tracking*: la tarea que publica el esquema del
applet (paso 10, D16) y la que cambia el esquema de eval (paso 17, D28). En las dos, lo que se toca del código es
exactamente lo que la forma de un control existente hace inseparable del fichero protegido, y en las dos la pausa
humana revisa las dos cosas juntas.

**Veredicto del gate: PASA.** Las dos piezas marcadas «con justificación» están en *Complexity Tracking* y ninguna
afecta a alcance, frontera humana, privacidad, términos de uso ni a una decisión cerrada.

### Re-evaluación tras la fase 1 (diseño)

- **Paquete `data` en la raíz de `data/`** (§V, estructura): es la única ubicación desde la que `//go:embed` alcanza
  `data/territorio/` sin duplicar los ficheros ni moverlos fuera de donde el spec los fija (V2, D2). No añade
  dependencias y no lo importa el dominio.
- **`go.yaml.in/yaml/v3` en el binario** (§V): dependencia ya fijada; la alternativa (`v4`, que el binario ya enlaza)
  solo tiene versiones candidatas y H5 la rechazó por eso (D15).
- **Registrar el applet y publicar su esquema en la misma tarea** (capa 3): lo impone la forma de los controles
  existentes (V8); la pausa humana sigue existiendo y revisa el contrato publicado (D16).
- **Parametrizar `internal/skills`, `internal/evals` y los casos negativos de `internal/app/skills_test.go`** (§V,
  YAGNI): no es capacidad nueva, es quitar el cableado a una sola skill que impide cumplir FR-067 y FR-084 y que
  dejaría los controles de la skill nueva pasando en vacío (D20, D22, D26).
- **Los controles sobre el corpus congelado viven en `internal/skills`** (§IV, R1): un test del dominio no puede leer
  un fichero, porque la denegación de `os` e `io` alcanza a los `_test.go` (V7). La consecuencia sobre la superficie
  del dominio —`Plegar` y los tipos de los ficheros, exportados— no añade capacidad y evita un segundo pliegue en el
  árbol (D27).
- **La tarea del esquema de eval toca una expectativa de test** (capa 3): el `then` nuevo cambia el mensaje que un
  caso existente compara literalmente, y el esquema se compila del fichero real (V28, V44). Va en *Complexity
  Tracking* como segunda excepción razonada del tipo D16 (D28).
- **La tarea de las grabaciones trae también el `enum` de `rango` del esquema de normas** (capa 3): el `enum` está
  atado por igualdad al conjunto de rangos grabados, así que los dos órdenes posibles dejan `make ci` en rojo (V47).
  Va en *Complexity Tracking* (D29). A diferencia de D16 y D28 **no toca código**: las dos piezas son material
  `[datos]` y caen dentro de la misma pausa.

**Veredicto tras el diseño: PASA**, sin ninguna violación no justificada.

## Project Structure

### Documentation (this feature)

```text
specs/008-h6-territorio-skill-legal/
├── plan.md              # Este fichero
├── research.md          # Fase 0: V1-V47 (verificaciones), D1-D29 (decisiones), S1-S9 (supuestos)
├── data-model.md        # Fase 1: ids, territorio, ficheros congelados, skill, formato de eval
├── quickstart.md        # Fase 1: guía de validación ejecutable
├── contracts/
│   ├── applet-territorio.md              # invocación, sobre, data, códigos, esquema publicado
│   ├── identificadores-ine-y-dir3.md     # internal/core/ids, gramáticas, fuzz
│   ├── datos-de-territorio.md            # data/territorio/, esquemas, tarea [datos], verificación
│   ├── skill-legal-core.md               # SKILL.md, referencias generadas, generación
│   └── evals-de-legal-core.md            # formato común ampliado, reglas del conjunto, job
├── spec.md
├── checklists/
├── gates/               # veredictos y evidencias (incluye verificacion-dir3.md)
└── tasks.md             # Fase 2 (/speckit-tasks; no lo crea este comando)
```

### Source Code (repository root)

Estructura de `docs/ROADMAP.md` §2 y `CLAUDE.md`. **NUEVO** = lo crea H6; ← = cambia.

```text
skills/legal-core/                          **NUEVO** el producto
├── SKILL.md                                protocolo que empieza por el territorio + región generada     (contrato de la skill)
├── references/leyes_vertebrales.md         generado desde data/normas.yaml (vertebral: true)             (D20)
├── references/jerarquia_normativa.md       generado desde data/jerarquia.yaml                            (D20)
└── scripts/territorio -> ../../../bin/instalado/kitlegal   generado
skills/boe-legislacion/references/normas.md ← **solo por regeneración**: las siete normas nuevas de
                                            data/normas.yaml (FR-074). Nada más de esa skill se toca        (obligación 6)
data/
├── datos.go                                **NUEVO** paquete data: solo //go:embed                       (D2, D15)
├── normas.yaml                             ← siete normas nuevas y la marca vertebral                    (FR-070)
├── jerarquia.yaml                          **NUEVO** [datos] niveles, boletín por nivel y reglas         (FR-067)
└── territorio/                             **NUEVO** [datos] ficheros congelados                          (D4, D5)
    ├── municipios.yaml                     relación del INE: código, dígito, nombre, provincia, comunidad
    ├── dir3.yaml                           correspondencia verificada INE→DIR3
    ├── estado.yaml                         el boletín estatal
    └── comunidades/<código>.yaml           19 ficheros; solo el de la Comunidad de Madrid con boletines
schemas/
├── territorio-municipios.yaml.json         **NUEVO** [datos]
├── territorio-dir3.yaml.json               **NUEVO** [datos]
├── territorio-estado.yaml.json             **NUEVO** [datos]
├── territorio-comunidad.yaml.json          **NUEVO** [datos]
├── jerarquia.yaml.json                     **NUEVO** [datos]
├── municipio.json                          **NUEVO** [datos] contrato del applet, desde --describe       (D16, D25)
├── normas.yaml.json                        ← [datos] campo vertebral (paso 13) y enum de rango con las
│                                           grabaciones (paso 15)                                        (FR-085, D29)
└── eval.yaml.json                          ← [datos] variante de territorio y esperado verificable       (FR-085)
evals/legal-core/                           **NUEVO** cubierto, no cubierto y no activación                (FR-080)
internal/
├── core/ids/                               **NUEVO** doc.go, ine.go, dir3.go, errores.go + tests + fuzz  (D8)
│   └── testdata/fuzz/Fuzz{CodigoINE,CodigoDIR3}/   **NUEVO** [datos] corpus versionado                    (FR-035)
├── core/territorio/                        **NUEVO** doc.go, fuentes.go, registro.go, nombres.go,
│                                           resolver.go, salida.go, errores.go + tests **sintéticos**:
│                                           ningún test del dominio lee un fichero                         (D3, D10, D11, D27)
├── app/territorio.go                       **NUEVO** el applet y su composición                           (contrato del applet)
├── app/territorio_test.go                  **NUEVO** sobre, data, códigos, esquema publicado
├── app/registro.go                         ← registra AppletTerritorio                                    (D16)
├── app/registro_test.go                    ← la lista de applets del registro
├── app/esquemas_test.go                    ← tabla de esquemas parametrizada por applet                   (V8, D16)
├── app/skills_test.go                      ← casos negativos parametrizados por skill                     (V25, D26)
├── app/testdata/script/territorio-matriz.txtar  **NUEVO** [datos] matriz territorial                      (FR-090)
├── app/testdata/script/argumentos.txtar    ← [datos] la lista de applets de la ayuda
├── app/testdata/script/ayuda.txtar         ← [datos] la fila del applet nuevo
├── app/ejemplo/kitlegal-e2e/main.go        ← registra el applet en el binario de e2e
├── skills/referencias.go                   ← cabecera, título y columnas por referencia                   (D20)
├── skills/sincronia.go                     ← tabla de generadores en vez del switch por nombre            (D20)
├── skills/frontmatter.go                   ← la referencia declara su fichero de datos                    (D20)
├── skills/normas.go                        ← campo Vertebral                                              (FR-067)
├── skills/jerarquia.go                     **NUEVO** lector de data/jerarquia.yaml
├── skills/territorio.go                    **NUEVO** validación de data/territorio/ contra su esquema; la
│                                           integridad la delega en territorio.Cargar, sin repetir tipos  (FR-044)
│                                           su test lleva además los tres controles del corpus real       (D27)
├── evals/formato.go                        ← campo Municipio, campo Territorio, formaDelComando           (D21)
├── evals/juzgar.go                         ← satisface, textoDelComando, reparto y condición de Pasa      (D21)
├── evals/territorio.go                     **NUEVO** extracción por forma fija                            (D21)
├── evals/consultas.go                      ← un comando de territorio no genera consulta                  (D21)
├── evals/conjunto.go                       ← ComprobarConjunto parametrizado + ReglasDeLegalCore          (D22)
├── evals/informe.go                        ← dos columnas más                                             (D21)
└── arch_test.go                            ← go.yaml.in/yaml/v3 en modulosDelBinario                      (V11)
cmd/kitlegal/main_test.go                   ← la lista de applets del binario
Makefile                                    ← skills-check: dos tests más en su expresión -run
testdata/evals/                             ← [datos] manifiesto y grabaciones de las siete normas nuevas  (FR-073)
.github/workflows/evals.yml                 ← matriz de skills y filtro de rutas                           (D23)
docs/SOURCES.md                             ← [datos] fecha del volcado del REL                            (FR-049)
README.md, CONTRIBUTING.md, CHANGELOG.md    ←                                                              (FR-098, FR-099)
```

El árbol nombra los ficheros de código y los de test que **cambian de forma** (parametrizaciones, listas fijas); el
fichero de test que acompaña a cada fichero de código nuevo está en el «Inventario de tests», con la **tarea en la que
entra**, y `tasks.md` declara sus rutas desde los dos sitios: ninguna tarea toca un fichero que no aparezca en uno de
ellos, y ningún fichero de los dos se queda sin tarea que lo cree o lo cambie.

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`. Dominio en `internal/core/{ids,territorio}`,
composición en `internal/app`, datos en `data/`, esquemas en `schemas/`, skill en `skills/`, evals en `evals/`. El
único elemento que el árbol no tenía es un paquete Go dentro de `data/`, que existe solo porque `//go:embed` no sube
de directorio (D2, V2) y que no contiene lógica. No se crean `packs/`, `mcp/`, `plugin/`, `pkg/`, `internal/graph`
ni `internal/store`.

## Controles mecánicos que este hito añade o toca

Según la sección «Gates» de la constitución. «Demostración» dice qué falla si el control se retira o se viola.

| # | Control | Test / forma | Orden | ¿En `make ci`? | Demostración |
|---|---|---|---|---|---|
| 1 | Ficheros de `data/territorio/` contra su esquema | `TestTerritorioDelRepositorio/esquema` | `skills-check`, `test` | Sí | Una clave de más, un código de cuatro cifras o una clave repetida → falla nombrando fichero y línea (quickstart §9) |
| 2 | Integridad entre ficheros congelados, incluida la coherencia de la `comunidad` de cada municipio con la que declara su provincia | `TestTerritorioDelRepositorio/integridad` | `skills-check`, `test` | Sí | Un DIR3 de un municipio inexistente, o incoherente con su código, → falla nombrándolo; una fila cuya `comunidad` no es la de su provincia → falla nombrando el municipio y las dos comunidades (data-model §2.1, punto 6; quickstart §9) |
| 3 | Procedencia de cada dato | `TestTerritorioDelRepositorio/fuentes` | `skills-check`, `test` | Sí | Un `source` que no es fila de `docs/SOURCES.md` ni fichero de `data/territorio/` → falla |
| 4 | Madrid configurada (sin prohibir otras); las 19 con régimen | `TestTerritorioDelRepositorio/madrid-configurada`, `/regimen-de-todas` | `skills-check`, `test` | Sí | Añadir boletines a otra comunidad, o quitar el régimen de una, → falla (SC-005, FR-055) |
| 5 | El pliegue de nombres cubre el corpus, no deja municipios inalcanzables y no deja ningún nombre que sea solo cifras | `TestTerritorioDelRepositorio/pliegue-cubre-el-corpus`, `/nombres-alcanzables`, `/ningun-nombre-es-solo-cifras`, en `internal/skills` (D27) | `skills-check`, `test` | Sí | Una runa fuera de la tabla del pliegue → falla nombrando el municipio; un pliegue que fusionara dos municipios dejando uno inalcanzable, sin declararlo como ambigüedad, → falla; un nombre que plegado fuera solo cifras → falla, porque la resolución lo leería como código (D10, data-model §2.6) |
| 6 | Gramáticas de INE y DIR3 | `TestAnalizarCodigoINE`, `TestAnalizarDIR3`, `TestComprobarDigito` | `test` | Sí | Aceptar `28074 ` con espacio, `2807`, provincia `00`, provincia `>52` (`99999`) o `X01280748` → falla (SC-006) |
| 7 | Ida y vuelta e idempotencia de los identificadores | `TestIdaYVuelta` | `test` | Sí | Un `String()` que pierda un cero por delante → falla (FR-032) |
| 8 | Fuzz de los dos analizadores con corpus versionado | `FuzzCodigoINE`, `FuzzCodigoDIR3` + `internal/core/ids/testdata/fuzz/` | `test` | Sí (semillas) | Un `panic` o una entrada aceptada sin ida y vuelta estable → falla; el corpus versionado se ejecuta sin `-fuzz` (V3) |
| 9 | Gramáticas atadas a los esquemas de datos | `TestTerritorioDelRepositorio/gramaticas` | `skills-check`, `test` | Sí | Un `pattern` del esquema que acepte lo que el analizador rechaza → falla (V38) |
| 10 | Forma del sobre y de `data` del applet | `TestResolverDevuelveElTerritorio` | `test` | Sí | Una clave de más o de menos en `data`, o una `cobertura` con dos claves, → falla (FR-006, FR-020) |
| 11 | Salida real contra el esquema publicado | `TestSalidaDeTerritorioContraSchemas` | `test` | Sí | Un `data` con una clave que el esquema no declara → falla (punto 4 de la Definition of Done) |
| 12 | `--describe` sin deriva y cobertura de esquemas | `TestEsquemasPublicados`, `TestEsquemasCubrenTodosLosVerbos` | `schema-check`, `test` | Sí | Cambiar la ayuda del verbo sin regenerar, o registrar un verbo sin publicar su parte, → falla (V8, SC-012) |
| 13 | Códigos de salida del applet | `TestCodigosDeTerritorio` y `territorio-matriz.txtar` | `test` | Sí | Un nombre ambiguo que devolviera 0, o un municipio inexistente que devolviera 2, → falla; cualquier 4, 5 o 6 → falla (SC-004, FR-016) |
| 14 | Candidatos del nombre ambiguo | `TestCodigosDeTerritorio/ambiguo` | `test` | Sí | Un mensaje con un candidato de menos, o sin provincia, o sin orden por código, → falla (FR-014) |
| 15 | Matriz territorial en e2e | `territorio-matriz.txtar` | `test-e2e` dentro de `test` | Sí | Que en la salida del municipio no cubierto aparezca el código o el nombre de un boletín no configurado → falla (FR-090, FR-091, SC-002) |
| 16 | Determinismo: nombre, código y `--offline` dan lo mismo | `territorio-matriz.txtar` (`cmp`) | `test` | Sí | Fechar el sobre con el reloj → las tres salidas dejan de ser iguales y falla (SC-001, D6) |
| 17 | Tiempo de respuesta | `cronometra` en `territorio-matriz.txtar` | `test` | Sí | Un análisis que se dispare por invocación → falla (S3) |
| 18 | Skill nueva: frontmatter, 300 líneas, deriva de referencias, tabla y enlaces | `TestSkillsDelRepositorio`, con sus casos negativos parametrizados por skill en `internal/app/skills_test.go` (V25, D26) | `skills-check`, `test` | Sí | Editar `references/jerarquia_normativa.md` a mano → falla nombrando el fichero (quickstart §11) |
| 19 | Referencias generadas desde `data/` e idempotencia | `TestRegenerarYComparar`, `make skills-sync` dos veces | `skills-check`, `test` | Sí | Marcar una norma `vertebral` sin regenerar → falla (FR-067, SC-009) |
| 20 | `data/normas.yaml` y `data/jerarquia.yaml` contra su esquema | `TestNormasDelRepositorio`, `TestJerarquiaDelRepositorio` | `skills-check`, `test` | Sí | Un `vertebral: "sí"` o un nivel fuera del enumerado → falla (FR-044) |
| 21 | Identificadores de las normas nuevas contra su búsqueda grabada | `TestIdentificadoresDeLasNormas` | `skills-check`, `test` | Sí | Un `BOE-A-…` escrito de memoria → falla nombrando la norma (FR-071, FR-072, SC-010) |
| 22 | Formato de eval ampliado y compatible | `TestLeerEval`, `TestEsquemaDeEval`, `TestEvalsDelRepositorio/formato` | `skills-check`, `test` | Sí | Una eval de `boe-legislacion` que dejara de validar → falla; una eval activa sin `citas` ni `territorio` → falla (FR-084) |
| 23 | Juicio del comando y del esperado de territorio | `TestFormaDelComando`, `TestJuzgar/territorio-*` | `test` | Sí | Dar por satisfecho `territorio resolver` con otra invocación, o dar por pasada una eval con un esperado ausente, → falla |
| 24 | Reglas del conjunto de `legal-core` | `TestConjuntoDeEvals`, `TestEvalsDelRepositorio/conjunto-legal-core` | `skills-check`, `test` | Sí | Quitar la eval de no activación o la de municipio no cubierto → falla (FR-081, FR-082) |
| 25 | Vocabulario de cobertura del esquema de eval y del applet | `TestEvalsDelRepositorio/cobertura-del-esquema` | `skills-check`, `test` | Sí | Un aspecto nuevo en el applet sin su valor en el esquema → falla nombrándolo |
| 26 | Reglas de arquitectura y perfil del binario | `TestArquitectura` R1-R3, `TestDependenciasDelBinario`, `depguard`, `forbidigo` | `lint`, `test` | Sí | Importar `io/fs` desde el dominio, o enlazar un módulo sin justificarlo en la lista, → falla (V5-V7, V11) |
| 27 | Que nada de esto toca la red | `TestArquitectura` R2 y la ausencia de caso nuevo en `scripts/verify-sources.sh` | `lint`, `test` | Sí | Una petición a INE, REL o DIR3 → falla en lint y en el test de arquitectura (FR-043, ADR 0017) |
| 28 | Evals con modelo | `.github/workflows/evals.yml` (matriz) → `make evals SKILL=…` | `evals` | **No** (job) | Una respuesta que no traslada la cobertura → la eval no pasa (SC-015) |

### Objetivos del `Makefile`

Ningún objetivo nuevo. **Una sola receta cambia**, la de `skills-check`:

- `skills-check` (←): su expresión `-run` gana `TestTerritorioDelRepositorio` y `TestJerarquiaDelRepositorio`; el
  paquete `./internal/skills/` ya está en su lista, así que no cambia nada más. Como la receta cambia, la fila de
  `make skills-check` de las tablas de controles de `README.md` y `CONTRIBUTING.md` se alinea en la misma rama con lo
  que el objetivo pasa a cubrir (obligación 8).
- `schema-check`: sin cambio de receta; gana `schemas/municipio.json` por la vía del registro de producción.
- `test` y `test-e2e`: sin cambio de receta; ganan el guion `territorio-matriz.txtar` y los tests de los paquetes
  nuevos.
- `evals`: sin cambio de receta; el job la invoca una vez por skill de la matriz.
- `ci`: los mismos diez prerrequisitos, en el mismo orden.

### Fixtures y datos protegidos

| Material | Tarea | Pausa |
|---|---|---|
| `schemas/territorio-municipios.yaml.json` + `data/territorio/municipios.yaml` | 1 `[datos]` | sí (esquema nuevo) |
| `schemas/territorio-comunidad.yaml.json` + `schemas/territorio-estado.yaml.json` + `data/territorio/comunidades/*` + `data/territorio/estado.yaml` | 2 `[datos]` | sí |
| `schemas/territorio-dir3.yaml.json` + `data/territorio/dir3.yaml` + `docs/SOURCES.md` + registro de verificación | 3 `[datos]` | sí |
| `schemas/jerarquia.yaml.json` + `data/jerarquia.yaml` | 14 `[datos]` | sí |
| `schemas/normas.yaml.json` (existente: campo `vertebral`) | 13 `[datos]` | sí |
| `schemas/eval.yaml.json` (existente: variante y esperado) + la expectativa de `internal/evals/formato_test.go` que su forma impone (D28) | 17 `[datos]` | sí |
| Manifiesto y grabaciones de las siete normas nuevas (`testdata/evals/`) + el `enum` de `rango` de `schemas/normas.yaml.json`, que esas grabaciones fijan (D29) | 15 `[datos]`; graba una persona | sí |
| `schemas/municipio.json` + registro del applet | 10 `[datos]` | sí (D16) |
| Guion e2e **nuevo** `territorio-matriz.txtar` (`internal/app/testdata/script/`) | 12 `[datos]` | no (fichero nuevo fuera del territorio de fixtures, V9) |
| Guiones e2e **existentes** `argumentos.txtar` y `ayuda.txtar` (`internal/app/testdata/script/`) | 10 `[datos]` | sí: casan `(^\|/)testdata/` y existen en la base, así que `clasificar_datos` los marca «(existente)» (V9); el paso 10 pausa además por `schemas/municipio.json` |
| Corpus de fuzz (`internal/core/ids/testdata/fuzz/`) | 5 `[datos]` | no (ficheros nuevos, V9) |

### Tests de contrato

| Contrato | Tests |
|---|---|
| [applet-territorio](./contracts/applet-territorio.md) | `TestResolverDevuelveElTerritorio`, `TestCodigosDeTerritorio`, `TestSalidaDeTerritorioContraSchemas`, `TestEsquemasPublicados`, `TestEsquemasCubrenTodosLosVerbos`, `territorio-matriz.txtar` |
| [identificadores-ine-y-dir3](./contracts/identificadores-ine-y-dir3.md) | `TestAnalizarCodigoINE`, `TestAnalizarDIR3`, `TestComprobarDigito`, `TestDIR3DeAyuntamiento`, `TestIdaYVuelta`, `TestClaseDeLosErrores`, `TestSuperficieDeIds`, `FuzzCodigoINE`, `FuzzCodigoDIR3`, `TestTerritorioDelRepositorio/gramaticas` |
| [datos-de-territorio](./contracts/datos-de-territorio.md) | `TestLeerTerritorio`, `TestTerritorioDelRepositorio` (`esquema`, `integridad`, `fuentes`, `madrid-configurada`, `regimen-de-todas`, `gramaticas`, `pliegue-cubre-el-corpus`, `nombres-alcanzables`, `ningun-nombre-es-solo-cifras`), `TestCargar` |
| [skill-legal-core](./contracts/skill-legal-core.md) | `TestSkillsDelRepositorio` (casos negativos por skill), `TestTablaDeComandosCoincideConLaGramatica`, `TestRegenerarYComparar`, `TestRenderizarLeyesVertebrales`, `TestRenderizarJerarquia`, `TestJerarquiaDelRepositorio`, `TestNormasDelRepositorio`, `TestIdentificadoresDeLasNormas` |
| [evals-de-legal-core](./contracts/evals-de-legal-core.md) | `TestLeerEval`, `TestEsquemaDeEval`, `TestFormaDelComando`, `TestJuzgar`, `TestConjuntoDeEvals`, `TestEvalsDelRepositorio`, `TestInforme` |

## Inventario de tests

Nombres fijados aquí para que `tasks.md` y `quickstart.md` los usen tal cual. Un fichero de test por fichero de
código, en el orden del fichero (`golang-testing`); tablas con subtests con nombre; `t.Parallel()` salvo donde se use
`t.Setenv`. La columna **Tarea** dice en qué paso del «Orden de implementación» entra cada fichero, para que
`tasks.md` pueda declarar su ruta y ninguna tarea lo toque antes de tiempo. Excepciones declaradas: los `doc.go` no
tienen test; `internal/app/territorio_test.go` cubre el applet y su esquema; y `data/datos.go` no tiene test propio. Sus directivas
`//go:embed` no necesitan uno —un patrón que no casa es error de compilación, y que el contenido valga lo comprueba
`TestTerritorioDelRepositorio` sobre los mismos ficheros, que lee por ruta relativa sin importar el paquete `data`
(contracts/datos-de-territorio.md §3: lo importan `internal/app` y el binario de e2e, nadie más)—; la única lógica del fichero, `Comunidades() (map[string][]byte, error)` —clave
por código y camino de error—, la ejercen `TestRegistroDeProduccion` (`internal/app/registro_test.go:233-237`, que
declara el paso 10) y el e2e.

**Ningún test del dominio lee un fichero** (D27): `internal/core/{ids,territorio}` se prueban enteros con datos
sintéticos, porque `depguard` deniega `os` e `io` a `**/internal/core/**` también en los `_test.go` (V7) y
`//go:embed` no sube de directorio (V2). Lo que hay que comprobar **sobre los ficheros congelados reales** —las tres
propiedades del pliegue sobre el corpus— son subtests de `TestTerritorioDelRepositorio`, en `internal/skills`, que es
donde el repositorio ya lee sus ficheros de `data/` por ruta relativa y ya importa el dominio (V42).

La regla alcanza también al **fuzz**: cada objetivo vive en el fichero de test de su analizador —`FuzzCodigoINE` en
`ine_test.go`, `FuzzCodigoDIR3` en `dir3_test.go`—, no hay ningún `ids.go` ni `ids_test.go` en el paquete, y el corpus
versionado se indexa por el nombre del objetivo, no por el del fichero (D8, contrato de identificadores §6). Y las
firmas que la tabla nombra son las del contrato: `ComprobarDigito` es **método** de `CodigoINE` (D8, data-model §1.1).

| Fichero | Tarea | Tests |
|---|---|---|
| `internal/core/ids/ine_test.go` | 4 | `TestAnalizarCodigoINE` (cinco cifras, seis cifras, ceros por delante, provincia fuera de rango, municipio `000`, vacío, no cifras, longitud), `TestComprobarDigito`, `TestIdaYVuelta/ine`, `FuzzCodigoINE` |
| `internal/core/ids/dir3_test.go` | 4 | `TestAnalizarDIR3` (mayúscula y minúscula, longitud, letra ajena), `TestDIR3DeAyuntamiento`, `TestIdaYVuelta/dir3`, `FuzzCodigoDIR3` |
| `internal/core/ids/errores_test.go` | 4 | `TestClaseDeLosErrores` (todo error del paquete declara `schema.ClaseArgumentos` y nombra la entrada), `TestSuperficieDeIds` |
| `internal/core/territorio/fuentes_test.go` | 6 | `TestCargar`: `completo`, `municipio-sin-provincia`, `provincia-en-dos-comunidades`, `comunidad-de-municipio-discrepante` (la de la fila no es la que declara su provincia, data-model §2.1 punto 6), `dir3-de-municipio-inexistente`, `dir3-incoherente`, `regimen-desconocido`, `codigo-de-fichero-distinto`, `yaml-ilegible`, `clave-repetida`. Todo con fuentes sintéticas |
| `internal/core/territorio/registro_test.go` | 6 | `TestRegistro`: consulta por código, consulta por forma plegada, candidatos de una forma que lleva a varios y su orden por código, y que el índice alcanza a todos los municipios de las fuentes sintéticas |
| `internal/core/territorio/nombres_test.go` | 6 | `TestPlegar` (mayúsculas, diacríticos, espacios), `TestFormasDelNombre` (artículo pospuesto, nombre bilingüe, las dos cosas), `TestPlegarEsIdempotente` |
| `internal/core/territorio/resolver_test.go` | 6 | `TestResolver`: `por-nombre`, `por-codigo`, `por-codigo-con-digito`, `digito-incorrecto`, `codigo-inexistente` (**provincia dentro de `01`-`52`** y municipio `001`-`999`, ausente de las fuentes sintéticas: es el único caso de código que da `ClaseNoEncontrado`), `nombre-inexistente`, `nombre-ambiguo` (candidatos y su orden), `entrada-de-solo-cifras-invalida` (provincia `00` o `>52`, municipio `000`, longitud imposible: `ClaseArgumentos`, **nunca** no encontrado), `entrada-vacia`, `nombre-con-mayusculas-y-sin-tildes`, `forma-alternativa` |
| `internal/core/territorio/salida_test.go` | 6 | `TestTerritorioResuelto`: las ocho claves, el `source` de cada dato, `boletines` del cubierto y del no cubierto, `cobertura` completa, la invariante del DIR3 no verificado, `TestFechaMasAntigua` |
| `internal/core/territorio/errores_test.go` | 6 | `TestClaseDeLosErroresDeTerritorio`: todo error del paquete declara `schema.ClaseArgumentos` (entrada mal formada, dígito distinto del oficial, nombre ambiguo) o `schema.ClaseNoEncontrado` (código o nombre que no está en la relación), ninguna otra clase, y cada mensaje nombra la entrada; los errores de carga no llevan clase de usuario (contrato del applet §7) |
| `internal/arch_test.go` (←) | 9 | `TestDependenciasDelBinario` con `go.yaml.in/yaml/v3` en la lista **y su motivo**. Va en el paso 9 y no más tarde: `go list -deps ./cmd/kitlegal` (`internal/arch_test.go:214-243`) recorre el paquete `internal/app` entero, así que el módulo queda enlazado en cuanto existe `internal/app/territorio.go`, esté o no registrado el applet |
| `internal/app/territorio_test.go` | 9 y 11 | Paso 9: `TestResolverDevuelveElTerritorio`, `TestCodigosDeTerritorio` (`ambiguo`, `no-encontrado`, `codigo-mal-formado`, `digito-incorrecto`, `sin-argumento`, y que ninguna invocación da 4, 5 ni 6). Paso 11, cuando el esquema ya está publicado: `TestSalidaDeTerritorioContraSchemas`, `TestSalidaSinBoletinesNoConfigurados` |
| `internal/app/esquemas_test.go` (←) | 8 y 10 | Paso 8: la tabla parametrizada por applet, sin fila nueva. Paso 10: la fila de `municipio.json` (entidad `municipio`, verbo `resolver`; D25) |
| `internal/app/skills_test.go` (←) | 20 y 21 | Paso 20: los existentes —`TestSkillsDelRepositorio` y `TestTablaDeComandosCoincideConLaGramatica`—, con los **casos negativos parametrizados por skill** en lugar de cableados a `boe-legislacion` (V25, D26), de modo que `/skills`, `/normas-nombradas` y `/sin-instrucciones-de-evals` los ejerzan también sobre `legal-core`; la constante `skillDelHito` pasa a la lista `skillsExigidas` (un elemento) y `/skills` exige con `require.Subset`. Paso 21: `legal-core` entra en esa lista, antes de crear sus ficheros (D26) |
| `internal/skills/territorio_test.go` | 7 | `TestLeerTerritorio` (sintéticos: válido, clave desconocida, patrón, clave repetida), `TestTerritorioDelRepositorio` (`esquema`, `integridad` —que delega en `territorio.Cargar`—, `fuentes`, `madrid-configurada`, `regimen-de-todas`, `gramaticas` —los `pattern` de los tres esquemas contra los analizadores de `internal/core/ids`, V38—, y los tres del corpus congelado: `pliegue-cubre-el-corpus`, `nombres-alcanzables` y `ningun-nombre-es-solo-cifras`, con `territorio.Plegar`, `Cargar` y `Resolver` sobre los ficheros reales; D27) |
| `internal/skills/jerarquia_test.go` | 16 | `TestLeerJerarquia`, `TestJerarquiaDelRepositorio`, `TestEsquemaDeJerarquia` |
| `internal/skills/normas_test.go` (←) | 15 (sin cambio) y 16 | Paso 15: **no se modifica**, pero es el control que hace indivisible esa tarea: `TestEsquemaDeNormas/rangos-grabados` exige que el `enum` de `rango` sea el conjunto exacto de rangos de las búsquedas grabadas (V47, D29). Paso 16: los existentes más `vertebral` (válido, tipo incorrecto) y `TestNormasDelRepositorio/vertebrales` (las quince de la tabla y ninguna más) |
| `internal/skills/referencias_test.go` (←) | 16 | `TestRenderizarNormas` (sin cambio), `TestRenderizarLeyesVertebrales` (filtra, cabecera y columnas), `TestRenderizarJerarquia` |
| `internal/skills/sincronia_test.go` (←) | 16 | `TestRegenerarYComparar` con las tres referencias y con una referencia declarada sin generador |
| `internal/skills/frontmatter_test.go` (←) | 16 | `TestLeerFrontmatter` y `TestValidarFrontmatter` con la referencia que declara su fichero de datos: una referencia conocida cuyo fichero no se llama como ella (`leyes_vertebrales` desde `data/normas.yaml`) y una declarada sin generador (D20, V22) |
| `internal/evals/formato_test.go` (←) | 17 y 18 | Paso 17, con el esquema y en su misma tarea `[datos]` (D28): el caso `positiva-sin-citas` pasa a `positiva-sin-esperado-verificable` con el mensaje que el `then` nuevo produce (V44). Paso 18: `TestLeerEval` con la variante de territorio y el esperado; `TestFormaDelComando`. `TestEsquemaDeEval` y `TestGramaticasCoincidenConBoe`, sin cambio |
| `internal/evals/territorio_test.go` | 18 | `TestExtraerTerritorio` (comunidad, provincia, boletín, aspecto de cobertura; tolerancias y lo que no cuenta) |
| `internal/evals/juzgar_test.go` (←) | 18 | `TestJuzgar/territorio-satisface`, `/territorio-otro-municipio-no-satisface`, `/territorio-ausente`, `/territorio-y-citas` |
| `internal/evals/consultas_test.go` (←) | 18 | `TestConsultasNecesarias` con un comando de territorio: no genera ninguna consulta que grabar, y la regla de lo grabado sigue aplicándose a los comandos del BOE (D21, FR-043) |
| `internal/evals/conjunto_test.go` (←) | 18 y 19 | Paso 18: `TestConjuntoDeEvals` con los dos juegos de reglas, sobre evals sintéticas, y `TestEvalsDelRepositorio/cobertura-del-esquema`. Paso 19: `TestEvalsDelRepositorio/conjunto-legal-core`, **en la misma tarea que las tres evals** —antes queda en rojo y después pasaría en vacío (obligación 12)—, así que esa tarea declara también esta ruta |
| `internal/evals/informe_test.go` (←) | 18 | Las dos columnas nuevas en `TestInforme/aprobado` |
| `internal/app/testdata/script/territorio-matriz.txtar` | 12 | Leganés (cubierto), Tordesillas (no cubierto, sin ningún boletín no configurado en la salida), un municipio foral, un nombre ambiguo, un código inexistente, la igualdad byte a byte entre nombre, código y `--offline`, la misma salida ejecutando el binario **desde otro directorio de trabajo** (FR-056), `--describe` **con su argumento** —sin él la invocación termina en 2 (V40)— y `cronometra` |

## Orden de implementación

`datos y esquemas → identificadores → dominio → datos embebidos → applet → esquema publicado y activación → e2e →
normas y referencias → formato de evals → evals → skill → job → documentación → plataforma` (D1). Las evals van
**antes** que `SKILL.md` (FR-083), y la parametrización de los casos negativos, justo antes de la skill (D26). Cada
tarea escribe su test antes del código y deja `make ci` en verde al terminar.

1. **`[datos]` Municipios**: `schemas/territorio-municipios.yaml.json` + `data/territorio/municipios.yaml`, generados
   por una persona en la pausa desde la relación del INE (contrato de datos §4). Nada de código.
2. **`[datos]` Comunidades y estado**: sus dos esquemas y los veinte ficheros; solo la Comunidad de Madrid con
   `boletines`, con el motivo de la equivalencia en su propio fichero.
3. **`[datos]` Correspondencia DIR3**: `schemas/territorio-dir3.yaml.json` + `data/territorio/dir3.yaml`, la fila de
   `mpt.rel` de `docs/SOURCES.md` con la fecha del volcado y el registro de verificación de la muestra
   (`gates/verificacion-dir3.md`). Es la pausa donde se decide si la derivación es regla o tabla (D19).
4. **Identificadores**: `internal/core/ids` con sus tests.
5. **`[datos]` Corpus de fuzz**: `internal/core/ids/testdata/fuzz/…` (sin pausa; tarea propia porque toca `testdata/`).
6. **Dominio del territorio**: `internal/core/territorio` (carga e integridad, registro e índices, pliegue de
   nombres, resolución, salida y errores) con sus tests sintéticos, uno por fichero.
7. **Datos embebidos y validación de repositorio**: `data/datos.go`, `internal/skills/territorio.go` y
   `internal/skills/territorio_test.go` con `TestLeerTerritorio` y `TestTerritorioDelRepositorio` —incluidos los tres
   subtests del corpus congelado, que viven aquí y no en el dominio (D27)—. Como `skills-check` ejecuta sus tests por
   nombre (V45), la tarea declara también `Makefile`: su expresión `-run` gana `TestTerritorioDelRepositorio`.
8. **Parametrización de la tabla de esquemas**: `internal/app/esquemas_test.go` deja de cablear `boe`, sin añadir
   ninguna fila todavía (V8).
9. **El applet**: `internal/app/territorio.go` con `TestResolverDevuelveElTerritorio` y `TestCodigosDeTerritorio`
   sobre un registro local del test, **sin registrarlo** en producción: así `make ci` sigue verde sin esquema
   publicado (V8). En la **misma** tarea, `internal/arch_test.go` gana `go.yaml.in/yaml/v3` en `modulosDelBinario` con
   su motivo: el cierre de `go list -deps ./cmd/kitlegal` recorre el paquete `internal/app` entero, así que el módulo
   queda enlazado por la mera existencia de este fichero, antes de registrar nada (V11).
10. **`[datos]` Esquema publicado y activación** (la tarea indivisible de D16): `schemas/municipio.json` generado con
    `-actualizar-esquemas`, la fila de la tabla en `internal/app/esquemas_test.go`, el registro en
    `RegistroDeProduccion` y en el binario de e2e, y los cuatro sitios que fijan la lista de applets
    (`internal/app/registro_test.go`, `cmd/kitlegal/main_test.go`, `internal/app/testdata/script/argumentos.txtar` y
    `ayuda.txtar`). Nada más.
11. **Salida contra el contrato publicado**: `TestSalidaDeTerritorioContraSchemas` y
    `TestSalidaSinBoletinesNoConfigurados`, que ya pueden leer el fichero publicado.
12. **`[datos]` Matriz territorial en e2e**: `internal/app/testdata/script/territorio-matriz.txtar`, con los cuatro
    casos, la igualdad byte a byte, la invocación desde otro directorio de trabajo (FR-056), `--describe` con su
    argumento (V40) y `cronometra`.
13. **`[datos]` Esquema de normas, campo `vertebral`**: `schemas/normas.yaml.json`, **solo la propiedad nueva**. No
    toca código y deja `make ci` en verde: la propiedad es opcional, ninguna norma la trae todavía y el lector
    decodifica sin `KnownFields`, de modo que el campo Go puede llegar después (V46). El `enum` de `rango` del mismo
    fichero **no se toca aquí**: `TestEsquemaDeNormas/rangos-grabados` lo compara con `assert.Equal` contra el
    conjunto exacto de rangos de **todas** las búsquedas grabadas (V47), así que ampliarlo antes de que existan las
    grabaciones que introducen esos rangos deja `make ci` en rojo. Crece en el paso 15, con ellas (D29).
14. **`[datos]` Jerarquía**: `schemas/jerarquia.yaml.json` + `data/jerarquia.yaml`. Tampoco toca código: hasta el paso
    16 no hay lector que lea ese fichero ni control que lo valide, así que no hay nada que pueda quedar en rojo.
15. **`[datos]` Normas vertebrales**: manifiesto y grabaciones de las siete búsquedas, que graba una persona en la
    pausa (FR-073), **y en la misma tarea** el `enum` de `rango` de `schemas/normas.yaml.json`, que gana los valores
    que esas búsquedas introducen. Es indivisible por la misma razón que D16 y D28 —un control ata las dos piezas—,
    aunque a diferencia de ellas **no toca código**: el `enum` tiene que ser el conjunto exacto de rangos grabados
    (V47), así que grabaciones y `enum` son un solo cambio y separarlos deja `make ci` en rojo entre dos tareas, en
    cualquiera de los dos órdenes (D29). Los valores se copian **de las respuestas
    grabadas**, nunca escritos de memoria, igual que el identificador y el título (S7): al menos entra «Ley Orgánica»,
    que hoy falta y que traen la LOPJ 6/1985 y la LOPDGDD 3/2018, y la pausa comprueba si el Código Civil —cuya norma
    de cabecera es de 1889— introduce alguno más. La pausa revisa las dos cosas juntas.
16. **Normas y generación**: `data/normas.yaml` con las siete normas y las quince marcas `vertebral`; el lector
    `internal/skills/jerarquia.go` con `internal/skills/jerarquia_test.go`; la tabla de generadores de
    `internal/skills` —`referencias.go`, `sincronia.go`, `frontmatter.go`, `normas.go` y sus tests— y **la única
    referencia que hay que regenerar aquí**: `skills/boe-legislacion/references/normas.md`, que cambia **solo por
    regeneración** al entrar las siete normas (FR-074). Las dos referencias de `legal-core` no existen todavía: se
    generan en el paso 21, con la skill que las declara. La tarea declara también `Makefile`, cuya expresión `-run` de
    `skills-check` gana `TestJerarquiaDelRepositorio` (V45). Este paso **no es indivisible** y se parte en dos
    rebanadas verdes, en este orden: primero el lector de la jerarquía con su control y el `Makefile` —que deja
    `make ci` en verde por sí solo—, y después las normas nuevas, las marcas `vertebral` y la tabla de generadores,
    que sí son un solo cambio porque `TestRegenerarYComparar` y `TestNormasDelRepositorio/vertebrales` los atan a
    `data/normas.yaml` y a la referencia regenerada.
17. **`[datos]` Esquema de eval**: `schemas/eval.yaml.json` con la variante de territorio y la regla del esperado
    verificable, **y en la misma tarea** la expectativa de `internal/evals/formato_test.go` que el `then` nuevo
    cambia —y `internal/evals/formato.go` si la composición del mensaje lo exige—. Es la segunda excepción razonada
    del tipo D16: el esquema y ese mensaje son un solo cambio, y separarlos deja `make ci` en rojo entre dos tareas
    (D28, V44). La pausa revisa las dos cosas juntas.
18. **Formato, juicio e informe**: `internal/evals` (campo, forma del comando, extracción, reparto, condición de
    `Pasa`, consultas necesarias, columnas, reglas parametrizadas del conjunto), con sus tests, incluidos
    `TestConjuntoDeEvals` sobre evals sintéticas y `TestEvalsDelRepositorio/cobertura-del-esquema`. El subtest
    `conjunto-legal-core` **no entra aquí**: exige tres evals que todavía no existen.
19. **Las evals**: `evals/legal-core/` con las tres, **antes** que `SKILL.md` (FR-083), y en la misma tarea el subtest
    `TestEvalsDelRepositorio/conjunto-legal-core` que las exige, primero el subtest y después los ficheros. La tarea
    declara por tanto las rutas de las tres evals y la de `internal/evals/conjunto_test.go`.
20. **Parametrización de los casos negativos de skills**: `internal/app/skills_test.go` deja de cablear
    `boe-legislacion`, sin añadir ninguna skill todavía —el recorrido pasa por una sola y `make ci` sigue verde—
    (V25, D26). La constante `skillDelHito` pasa a ser la lista `skillsExigidas`, que aquí tiene un solo elemento, y
    la exigencia de `/skills` pasa de `require.Contains` a `require.Subset` sobre ella. No es la última tarea que toca
    ese fichero: el paso 21 vuelve a él para añadir `legal-core` a esa lista.
21. **La skill**: primero la exigencia y después los ficheros. La tarea declara `internal/app/skills_test.go` y añade
    `legal-core` a `skillsExigidas` —el control de que la skill existe queda en rojo—, y a continuación
    `skills/legal-core/SKILL.md`, sus **dos** referencias —`leyes_vertebrales.md` y `jerarquia_normativa.md`, que
    nacen aquí— y su enlace, regeneradas con `make skills-sync`, que lo devuelven a verde. Los casos negativos
    parametrizados en el paso 20 la alcanzan sin tocar nada más. Que `make skills-sync` **ejecute** ese fichero
    (`scripts/skills-sync.sh:7`) no lo modifica, así que nada impide tocarlo aquí.
22. **El job**: `.github/workflows/evals.yml` con la matriz y `internal/core/*` en el filtro de rutas.
23. **Documentación**: `CHANGELOG.md`, `README.md`, `CONTRIBUTING.md` y la fila de `make skills-check`.
24. **`[plataforma]`**: empujar la rama, abrir la propuesta de cambio y recoger la ejecución de evals de cierre.

## Complexity Tracking

> Piezas que el spec no enumera y el diseño necesita, y todo lo que se aparta de la lista de la constitución §V o de
> una práctica establecida del repositorio.

| Divergencia | Por qué es necesaria | Alternativa más simple, y por qué se rechaza |
|---|---|---|
| **Un paquete Go dentro de `data/`** (`data/datos.go`) | `//go:embed` interpreta sus patrones relativos al directorio del paquete y no admite `..` ni enlaces simbólicos (V2), y FR-041, FR-042 y FR-056 exigen a la vez que los ficheros estén en `data/territorio/` y viajen dentro del binario | *Mover los datos a `internal/core/territorio/data/`*, como `internal/cache/migraciones`: contradice FR-041 y `docs/ROADMAP.md` §2. *Un paquete en la raíz del módulo*: ocupa el nombre de importación público que la constitución reserva a `pkg/legalkit`. *Un fichero Go generado con los municipios como literales*: contradice la clarificación Q4 del spec y añade 8.000 líneas generadas con su propio control de deriva |
| **`go.yaml.in/yaml/v3` pasa a enlazarse en el binario** (§V, FR-124 de H4) | El dominio analiza los ficheros congelados, que son YAML porque `docs/ROADMAP.md` §2 lo fija para `data/territorio/` | *`go.yaml.in/yaml/v4`*, que el binario ya enlaza (V11): solo tiene versiones candidatas y H5 la rechazó por eso. *JSON en lugar de YAML*: se analizaría con la biblioteca estándar, pero rompe la homogeneidad de `data/` que el roadmap fija. *Analizar en `internal/app`*: mueve el análisis a la raíz de composición y deja al dominio sin validar lo que recibe |
| **Pliegue de nombres propio en lugar de `golang.org/x/text/unicode/norm`** (§V) | El corpus es cerrado y versionado, así que lo que el pliegue cubre se puede probar sobre las filas reales; `x/text` sería una dependencia directa nueva fuera de §V y seguiría necesitando los mismos dos tests | *Promover `x/text` a directa*: dependencia fuera de la lista para algo que el corpus permite cerrar mejor. *No plegar*: `leganes` no encontraría `Leganés` (FR-015) |
| **La tarea `[datos]` que publica `schemas/municipio.json` registra también el applet** (capa 3, criterio «no mezclan otro trabajo») | En cuanto `territorio resolver` está en el registro, `TestEsquemasCubrenTodosLosVerbos` exige su parte publicada, y el esquema se genera **desde** el applet ya registrado: separarlos deja `make ci` en rojo entre dos tareas (V8, D16) | *Dos tareas con un rojo en medio*: rompe la regla de rebanadas verdes. *Escribir el esquema a mano antes del applet*: tiene que coincidir byte a byte con `--describe`, y corregirlo desde una tarea de código tocaría `schemas/`, que el guardián rechaza (V10). *Relajar el control de cobertura*: arreglar el control en lugar del código |
| **La tarea `[datos]` del esquema de eval trae la expectativa de test que su forma impone** (capa 3, criterio «no mezclan otro trabajo») | El `then` nuevo (`required: [comandos]` + `anyOf`) cambia el defecto que ve el lector: el nodo `anyOf` tiene causas, así que `incumplimientosDe` desciende a las hojas de sus dos ramas y el caso `positiva-sin-citas` de `TestLeerEval`, que compara con `EqualError`, deja de ver «`missing property 'citas'`» (V44). El esquema se compila del fichero real (V28): esquema y expectativa son un solo cambio | *Dos tareas, esquema y código*: deja `make ci` en rojo en medio, contra la regla de rebanadas verdes; y la segunda tocaría `schemas/` desde una tarea de código, que el guardián rechaza (V10). *Reordenar 17 y 18*: el rojo está dentro del paso 17 y los tests del 18 necesitan el esquema ya cambiado. *Relajar el caso a `ErrorContains`*: arreglar el control en lugar del contrato (D28) |
| **La tarea `[datos]` de las grabaciones trae el `enum` de `rango` de `schemas/normas.yaml.json`** (capa 3, criterio «no mezclan otro trabajo») | `TestEsquemaDeNormas/rangos-grabados` compara ese `enum` con `assert.Equal` contra el conjunto exacto de `rango.texto` de todos los resultados de todas las búsquedas grabadas (V47): ampliarlo antes deja el `enum` con un valor que ninguna grabación produce, y grabar antes deja una grabación con un rango que el `enum` no admite. Los dos órdenes son rojos | *Dos tareas, grabaciones y `enum`*: rojo en medio en cualquiera de los dos órdenes, contra la regla de rebanadas verdes. *Dejar el `enum` en el paso 13*: es el orden que el juez rechazó; el paso 13 corre antes de que existan las grabaciones. *Relajar el control a «el `enum` contiene los rangos grabados»*: arreglar el control en lugar del contrato, y perdería la detección de valores muertos que hoy da la igualdad (D29) |
| **Las tareas `[datos]` de datos congelados traen su esquema** | La pausa humana que FR-045 exige **solo se dispara** si el diff toca `testdata/` o `schemas/` (V9); un fichero nuevo bajo `data/` no pausa | *Dejar los datos en una tarea sin pausa*: incumple FR-045. *Cambiar `clasificar_datos`*: tocar el proceso desde dentro del hito, rechazo por alcance |
| **Parametrizar la tabla de esquemas por applet** (`internal/app/esquemas_test.go`) | Nueve puntos cablean `boe` (V8) y ningún applet nuevo puede publicar su contrato sin tocarlos | *Un segundo comparador para `territorio`*: dos caminos para el mismo control |
| **Parametrizar por skill los casos negativos de `TestSkillsDelRepositorio`** (`internal/app/skills_test.go`) | Los controles recorren todas las skills, pero sus casos negativos están cableados a `skillDelHito = "boe-legislacion"` (32 usos, V25): sin parametrizar, el control 18 y el contrato de la skill §5 prometerían una demostración que `legal-core` no recibe, y sus controles pasarían en vacío (obligación 12) | *Dejarlo cableado y retirar el control del plan y del contrato*: la skill nueva se queda sin la demostración de que sus controles fallan cuando deben. *Copiar los casos para `legal-core`*: segunda copia que diverge en H8, lo mismo que D22 rechaza (D26) |
| **Tabla de generadores de referencias** (`internal/skills`) | El nombre de la referencia es hoy también el del fichero de datos (V22), y `leyes_vertebrales` sale de `data/normas.yaml`: sin romper el acoplamiento, la skill no puede declararla | *`data/leyes_vertebrales.yaml` aparte*: duplica las normas y rompe la fuente única, contra la clarificación Q2 |
| **`ComprobarConjunto` parametrizado por reglas** (`internal/evals`) | Las reglas están cableadas al nombre de una skill (V27) y `evals/legal-core/` quedaría sin más control que el formato (V26) | *Copiar la función para `legal-core`*: tercera copia en H8, con las reglas comunes divergiendo |
| **Matriz de skills en el job de evals** | SC-015 exige que en **la misma ejecución** pasen las de `legal-core` y sigan pasando las de `boe-legislacion` | *Recorrer las dos skills dentro de `scripts/evals.sh`*: un solo informe con dos skills mezcladas y el doble de tiempo. *Dos ejecuciones*: no son la misma |
| **`internal/skills` gana dos lectores (`territorio.go`, `jerarquia.go`) y los tres controles del corpus real** | FR-044 pide validar los ficheros nuevos contra su esquema en `make ci`, y ahí es donde vive el lector común (V37); es además el único paquete desde el que se pueden leer esos ficheros, porque el dominio tiene denegada la entrada y salida también en sus `_test.go` (V7, D27) | *Validarlos desde el dominio*: metería `santhosh-tekuri/jsonschema` en el binario, y un test del dominio no puede ni abrirlos. *No validarlos*: incumple FR-044. *Los controles del corpus en `internal/app`*: separa lo que se afirma del mismo corpus en dos paquetes y los deja fuera de `make skills-check` |

## Obligaciones que este plan traslada a `tasks.md`

1. **Orden y pausas** del apartado «Orden de implementación». Las **diez** tareas `[datos]` —pasos 1, 2, 3, 5, 10,
   12, 13, 14, 15 y 17, **una por tarea**: la tabla «Fixtures y datos protegidos» tiene once filas porque el paso 10
   ocupa dos, la del esquema publicado con el registro y la de los dos guiones e2e existentes— no tocan código, con **dos** excepciones
   razonadas y solo esas: el paso 10, que registra el applet porque el esquema no se puede generar antes (D16), y el
   paso 17, que trae la expectativa de `internal/evals/formato_test.go` que el `then` nuevo cambia (D28). Ninguna otra
   tarea `[datos]` declara una ruta de código. La `[plataforma]` va la última y no mezcla otro trabajo.
2. **El ejecutor nunca descarga, graba ni escribe lo que fija la persona**: ninguna tarea ejecuta `KITLEGAL_RECORD`,
   `scripts/grabar-evals.sh`, `make evals`, `make verify-sources` ni ninguna descarga; los identificadores `BOE-A-…`,
   los títulos, los códigos INE, los dígitos de control y los DIR3 se copian solo de una respuesta grabada o de un
   fichero congelado ya revisado. Si falta algo, la tarea se detiene y lo anota en `gates/tarea-Tnnn.md`.
3. **Rutas declaradas** (guardián de diff). El extractor toma como declarada **toda ruta que aparezca en cualquier
   parte de la línea**, así que:
   1. cada fichero que la tarea cambia va por su ruta completa, sin llaves ni comodines, incluidos los ficheros de
      test;
   2. lo que solo se lee, se compara o se ejecuta se nombra sin directorio (`municipios.yaml`, `grabaciones.json`) o
      por su test o su objetivo de `make` (`TestTerritorioDelRepositorio`, `make skills-check`), nunca con la ruta de
      un paquete ni de un directorio de datos;
   3. desde la tarea siguiente a cada `[datos]`, ninguna línea puede dejar extraer `data/`, `data/territorio/`,
      `data/territorio/comunidades/`, `schemas/`, `testdata/`, `testdata/evals/`, `internal/app/testdata/`,
      `internal/core/ids/testdata/` ni ningún directorio que contenga material ya protegido;
   4. se comprueba con la tubería de extracción del propio `workflow.yml`, aplicada tal cual a la línea entera, antes
      de dar la tarea por buena.
4. **`misspell`** (V4): en Go no se escriben sueltas `aspectos`, `configuracion` ni `autonomos`; se usan `aspecto` en
   singular, `configuración` y `autónomas` con tilde. No se añade ninguna entrada a `ignore-rules` salvo que la
   implementación demuestre que necesita una de las tres literal, y entonces con su comentario de motivo.
5. **Sin atajos**: ningún `//nolint`, ningún test saltado, ningún TODO diferido, ningún error silenciado y ninguna
   exclusión de lint nueva. Los dos analizadores de `ids` viven en ficheros distintos y comparten la comprobación de
   cifras, para no disparar `dupl` (umbral 100 fichas).
6. **Sin cambios fuera del alcance**: ninguna tarea declara ficheros de `internal/httpx`, `internal/cache`,
   `internal/source/boe`, `internal/cli` ni `internal/core/schema`; `skills/boe-legislacion/SKILL.md` no se toca y
   `skills/boe-legislacion/references/normas.md` solo cambia por regeneración, en el paso 16, que es la única tarea
   que declara esa ruta; ninguna eval de `boe-legislacion` se modifica (FR-074, SC-015).
7. **Los dos paquetes de dominio nacen con su `doc.go`** (lo exige `revive`) y con cobertura suficiente para no bajar
   del 85 % de `internal/core/**`, que en esta propuesta de cambio mide de verdad (V19). Esa cobertura la dan **sus
   propios tests sintéticos**: `make test` mide por paquete y sin `-coverpkg` (V43), así que lo que los subtests de
   `internal/skills` ejercen del dominio no cuenta para el umbral.
8. **Documentación alineada en la misma rama** (FR-098, FR-099): `CHANGELOG.md` (*Unreleased*) registra el applet, la
   skill y los datos nuevos; `README.md` y `CONTRIBUTING.md` actualizan donde enumeran applets, skills, directorios de
   `data/` y pasos de trabajo. Si alguna tarea cambia lo que cubre un objetivo de `make`, la fila correspondiente de
   sus tablas se actualiza en la misma rama.
9. **Evidencias** fechadas por commit en `gates/`: `verificacion-dir3.md` (FR-047), `pr-h6.md` (salida del quickstart,
   supuestos pendientes) y `evals-cierre.md` con el informe de la ejecución de aceptación (SC-013, SC-015). SC-015
   añade una condición sobre el cierre: **el commit del informe es de la rama del hito y la cabeza que se fusiona
   solo puede diferir de él en ficheros bajo `specs/008-h6-territorio-skill-legal/`**, así que después de la
   ejecución de evals ninguna tarea toca código, datos, esquemas ni documentación; si hiciera falta, se repite la
   ejecución.
10. **Umbrales**: global ≥ 70 % e `internal/core/**` ≥ 85 %; nunca se rebaja un umbral ni se excluye un fichero.
11. **`rtk` y las formas de la sesión desatendida**: toda orden cuya salida se filtre, se cuente o se compare se
    ejecuta con `rtk proxy`, **cada etapa de la tubería incluida** —el gancho reescribe y resume `go test`, `make`,
    `git status` y `git diff`, y una línea de más falsea cualquier recuento—. Las órdenes del quickstart están
    escritas en la forma exacta en que la sesión `claude -p --permission-mode acceptEdits` las ejecuta sin pedir
    aprobación (`.claude/settings.json`), como en los quickstarts de H4, H5 y H5.1: nada de `rm -rf` (denegado),
    `mktemp`, `cd`, `printf … >> fichero` ni `test` sueltos; carpeta temporal de nombre fijo, `make -C` y `git -C`
    para el clon, `rtk proxy sh -c '…'` para códigos de salida y capturas, y `rtk proxy perl -0pi` para provocar un
    defecto sobre el clon.
12. **Sin pasar en vacío**: `TestTerritorioDelRepositorio` exige que existan los cuatro ficheros y las 19 comunidades;
    `TestEvalsDelRepositorio/conjunto-legal-core` exige las tres evals; `TestSkillsDelRepositorio/skills` exige que
    `skills/legal-core` esté entre las skills listadas, porque el paso 21 lo añade a `skillsExigidas` antes de crear
    sus ficheros (D26); `TestNormasDelRepositorio/vertebrales` exige las quince marcas.

## Comprobación contra la rúbrica del juez (`juez_plan`, criterios a-j)

| Criterio | Dónde se cumple |
|---|---|
| a. constitution_check | Un ítem por principio (I-IX) y por regla de dependencia (las cinco de §IV, la del grafo, la de los ejemplos del ADR 0010, la del espacio reservado del ADR 0006 y la de módulos del binario), con cómo lo cumple H6, qué lo vigila y veredicto; gates por capa con las pausas previstas; reglas del modo desatendido; re-evaluación tras el diseño. Los dos «cumple con justificación» remiten a *Complexity Tracking* |
| b. dependencias | **Ninguna dependencia nueva**; `golang.org/x/text` se rechaza expresamente (D10). Lo único que cambia del perfil del binario —`go.yaml.in/yaml/v3` enlazada— está en *Complexity Tracking* y en `modulosDelBinario` con su motivo |
| c. reglas_dependencia | El dominio recibe bytes precisamente para no importar `io/fs`, denegado por colgar de `io` (V5-V7); `net/http` y SQLite no aparecen; ningún `package main` nuevo; el applet no escribe (devuelve `Resultado`); el paquete `data` solo declara variables embebidas; `TestArquitectura` cubre los paquetes nuevos sin tocar nada. La regla alcanza a los `_test.go` (`run.tests: true`) y el inventario la respeta: **ningún test de `internal/core/**` lee un fichero**, todos son sintéticos, y los tres controles sobre el corpus congelado son subtests de `TestTerritorioDelRepositorio` en `internal/skills`, que es donde el repositorio ya lee `data/` por ruta relativa (D27, V42) |
| d. errores_exit_codes | Errores tipados con `schema.ConClase` → `argumentos` (2) y `no-encontrado` (3); el mensaje llega literal al sobre (V17); tabla de códigos por caso de entrada (data-model §2.6); comprobación de que 4, 5 y 6 no ocurren nunca (FR-016) y de que no hay `panic` (fuzz) |
| e. tests_primero | Inventario con nombres fijos, subtests y **la tarea en la que entra cada fichero**, contratos con su columna «qué lo vigila», 28 controles mecánicos con demostración, orden que pone las evals antes de `SKILL.md` y el e2e en cuanto hay applet, y la lista de fixtures con su tarea y su pausa. Todo fichero que un control toca está declarado en el árbol, en el inventario y en el orden —`internal/app/skills_test.go`, `internal/arch_test.go`, `internal/skills/jerarquia.go` y `skills/boe-legislacion/references/normas.md` incluidos—, de modo que su tarea puede declarar la ruta y ninguno se queda sin tarea; cada rebanada deja `make ci` en verde, y las tres tareas en que dos piezas son inseparables —porque un control las ata— están razonadas (D16, D28, D29). Todo fichero que **dos** tareas modifican lo dice en su columna «Tarea» y en las dos tareas: `internal/app/esquemas_test.go` (8 y 10), `internal/app/territorio_test.go` (9 y 11), `internal/evals/formato_test.go` (17 y 18), `internal/evals/conjunto_test.go` (18 y 19), `internal/app/skills_test.go` (20 la parametrización y 21 la exigencia de `legal-core`, primero en rojo y después los ficheros que la cierran) y `schemas/normas.yaml.json` (13 el campo, 15 el `enum`, D29). `internal/skills/normas_test.go` figura además en el paso 15 **sin modificarse**: es el control que hace indivisible esa tarea. Las firmas y las ubicaciones que el inventario fija son las mismas en los tres artefactos (`ComprobarDigito` como método, los dos objetivos de fuzz en el test de su analizador, los tres controles del corpus en `internal/skills`), y la redundancia `comunidad`→provincia tiene su control (control 2) |
| f. alcance | Solo lo que el spec pide; lo que el spec deja explícitamente fuera (grafo, festivos, `.kitlegal/config.yaml`, otros verbos, otros territorios, `data/boletines/`, competencia, `legal-core` v1) no aparece en ninguna pieza; lo que el diseño añade sin que el spec lo enumere está en *Complexity Tracking* |
| g. sin_atajos | Obligación 5: ningún `nolint`, test saltado, TODO ni error silenciado; ninguna exclusión de lint nueva; ningún control relajado —en particular, la alternativa de relajar el control de cobertura de esquemas se rechaza por escrito (D16)— |
| h. mejor_alternativa | D1-D29, cada una con su alternativa rechazada y su motivo; en particular D2 (dónde vive el embebido), D4 (formato de los ficheros), D6 (fecha del sobre), D9 (dígito como dato y no como algoritmo), D10 (pliegue propio frente a `x/text`), D15 (v3 frente a v4), D16 (orden de la publicación del esquema), D20, D21, D22, D23, D25 (nombre del fichero publicado), D26 (parametrizar frente a copiar), D27 (dónde viven los controles sobre el corpus real), D28 (qué hace indivisible la tarea del esquema de eval) y D29 (en qué tarea crece el `enum` de `rango`) |
| i. afirmaciones_verificadas | Tabla V1-V47 con la orden o el `fichero:línea` de cada comprobación: dónde lee `internal/skills` los ficheros del repositorio y qué puede importar (V42), que la cobertura se mide por paquete (V43), qué mensaje cambia el `then` nuevo del esquema de eval (V44), que `skills-check` ejecuta sus tests por nombre (V45), que el lector decodifica sin `KnownFields` (V46), que el `enum` de `rango` está atado por igualdad al conjunto de rangos grabados y hoy no tiene «Ley Orgánica» (V47), el recuento de `misspell.ignore-rules` (V39, 20 entradas), el orden entre el análisis de la invocación y la decisión de describir, con su sonda (V40), las dos formas del quickstart que leen la ausencia de salida, ejecutadas tal cual (V41), y además `go doc embed`, `go doc testing.F`, `go doc` de yaml, dos sondas reales de `misspell` con el `golangci-lint` del repositorio, el recorrido de `internal/arch_test.go`, las listas de `depguard` y `forbidigo`, el paso `clasificar_datos` y el `precheck_tasks` del workflow, los controles de esquemas, la lista de módulos del binario, `codecov.yml`, el flujo de evals y el andamiaje de skills y evals. Lo que depende de la red, de la plataforma o del fichero que genera la tarea `[datos]` está en S1-S9 como supuesto, no como hecho |
| j. quickstart_ejecutable | Escenarios con órdenes, rutas, flags y datos reales; los defectos se provocan solo sobre clones desechables bajo `/tmp/kitlegal-quickstart-h6/`, uno por defecto provocado, nunca sobre el árbol; cada escenario dice qué escribe y lo deja como estaba; el que depende de datos que aún no existen lo declara y da la forma exacta con la que se ejecutará. Las órdenes están escritas en la forma en que se ejecutan: `rtk proxy` en cada etapa de lo que se filtra o se compara, `--describe` **con su argumento** (V40) y la tabla de formas aceptadas por la sesión desatendida, como en H4, H5 y H5.1; las secciones 1 a 13 son locales y sin red, la 14 es de plataforma |
