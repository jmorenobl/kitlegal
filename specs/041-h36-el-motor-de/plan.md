# Implementation Plan: H36 · El motor de cómputo descrito por datos (ADR 0061)

**Branch**: `041-h36-el-motor-de` | **Date**: 2026-11-22 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/041-h36-el-motor-de/spec.md`

## Summary

`internal/core/plazos` deja de decidir por el código de un tipo o de un cómputo. `data/plazos.yaml` describe cada
cómputo con sus tres ejes —los municipios que cuentan con su texto, las unidades que admite y sus periodos del año
(`inhabil`, `no-corre` o `aviso`, cada uno con su precepto)— y cada tipo con su cómputo, su duración o sus unidades, sus
preceptos y la fecha que da (`da`). Un solo algoritmo cuenta cualquier cómputo con lo que el ADR 0061 deja fijo en el
motor (research D2); `LeerTipos` conserva su firma, cambia sus reglas por las diez de §C.2 y copia en cada tipo su
cómputo (D1); `CodigosDeAviso` y `EtiquetaDeAviso` reciben los tipos leídos, e `internal/evals` los toma del
`data/plazos.yaml` del binario (D3, D4). La ayuda y la salida de `calcular` dejan de nombrar tipos y avisos (D5). Entra
el cómputo civil de la LEC, sin ningún tipo que lo pida, comprobado contra los arts. 130 y 133 grabados en
`evidencias/h36/` (D9, D14). Con los datos de hoy, el `data`, el `hash`, los avisos y los errores de `calcular` son los de
`main`, byte a byte, y lo comprueban sobres capturados con el binario de `main`. Ninguna dependencia, paquete, orden,
bandera ni código de salida nuevo.

## Technical Context

**Language/Version**: Go 1.27 (`go 1.27.0`, `toolchain go1.27.4` en `go.mod`; research V1).

**Primary Dependencies**: las fijadas, ninguna nueva: `alecthomas/kong` v1.17.0 (ayuda de los argumentos),
`invopop/jsonschema` v0.14.0 (esquema de `--describe`, con la etiqueta `pattern`; V3), `go.yaml.in/yaml/v3` v3.0.6 (el
lector de `data/plazos.yaml`, el mismo de hoy; V4). En los tests, `stretchr/testify` v1.12.1, `rogpeppe/go-internal`
v1.16.0 (testscript; V6) y la biblioteca estándar `go/parser`, `go/ast` y `go/token` (V5).

**Storage**: ficheros versionados: `data/plazos.yaml` (empotrado como `data.Plazos`), dos ficheros de prueba en
`internal/core/plazos/testdata/` y las grabaciones de `evidencias/h36/`. Nada en la caché ni en el grafo: `plazos` no
los lee.

**Testing**: `go test -race` en `make ci`: tests de tabla del dominio sobre el calendario sintético, la rejilla de las
propiedades y el cómputo de ensayo; tests de `internal/app` sobre los datos reales y los tres juegos de `--describe`; un
test con `go/ast` sobre el Go de cuatro paquetes; `TestPreceptosDeLosPlazos` contra lo grabado; dos guiones testscript
de aceptación congelados. Sin red ni modelo. Skills de Go en research D15.

**Target Platform**: Linux y macOS, donde corre `make ci`.

**Project Type**: CLI multicall; el cambio es del dominio `internal/core/plazos`, su applet y el arnés de evals.

**Performance Goals**: ninguno del spec. La rejilla de las propiedades crece con la del civil (tres años de inicios por
40 duraciones, por municipio y unidad), del mismo orden que la de H9.

**Constraints**: con los datos de hoy, la salida estándar y el código de cada orden de `calcular`, byte a byte, los de
`main` salvo `fecha_consulta` (FR-011); `--help` y `--describe` independientes de los datos (FR-012); ningún código de
los datos en el Go que no es de test (FR-010); solo los tests de H9 que nombra FR-017, más las tres desviaciones que
obliga FR-016 y la entrada de `probarLosDatosQueNoSeLeen` («Desviaciones declaradas»; `gates/supuestos.md`).

**Scale/Scope**: 3 cómputos (uno sin tipos) y 7 tipos en `data/plazos.yaml`; 1 campo exportado nuevo (`Tipo.Da`), 2
firmas que cambian, 10 constantes, 2 funciones y 1 tipo exportados que desaparecen; 1 error de argumentos nuevo; 6
ficheros de producción que cambian (`tipos.go`, `computo.go`, `calcular.go`, `avisos.go`, `errores.go`, `salida.go`), el comentario
de paquete de `doc.go`, `internal/app/plazos.go`, `internal/evals/plazos.go`, `internal/evals/grabaciones.go` y `data/plazos.yaml`; 2 ficheros
de prueba y 4 ficheros de test nuevos; 2 guiones de aceptación.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principio o regla | Cómo lo cumple este hito |
|---|---|
| I · Fuentes públicas y frontera humana (INNEGOCIABLE) | La única petición es la grabación de dos bloques del BOE consolidado, fuente pública con fila revisada en `docs/SOURCES.md` de `main` (2026-09-13), en el paso `grabar_datos` y por `internal/httpx` (GET, `robots.txt`, ritmo; research V7, V17). Ninguna identidad, ningún POST, ningún CAPTCHA |
| II · Nada sin cita ni fuente | El sobre no cambia (FR-011). Cada periodo nombra el precepto que lo funda, que ya está entre los del cómputo o del tipo y sale en `data.preceptos` como hoy (§C.1); el texto del aviso del art. 183 es el de H9, ahora dato. El civil no lo pide ningún tipo, así que `calcular` no declara cubierto nada nuevo (US2.5); la cobertura en positivo de H33 no cambia |
| III · Tests primero y offline | Primero los dos guiones de `aceptacion/`, que fallan hoy por una aserción (contracts/aceptacion.md); unitarios offline sobre el calendario sintético, la rejilla y los datos empotrados; la verificación del civil, contra lo grabado. Toda salida de `calcular` se sigue validando contra `schemas/plazo.json` (`TestAppletPlazos`). El código nuevo de `internal/core/plazos` lo cubren `TestLeerTipos` (cada regla), `TestComputoDeEnsayo` (cada eje), `TestCalcular` y las propiedades, para el ≥ 85 % de `internal/core/**` |
| IV · Arquitectura hexagonal con reglas ejecutables | El dominio recibe `data/plazos.yaml` como bytes; `internal/app` compone. El error nuevo de FR-007 es un `errorDeArgumentos` (clase `argumentos`, código 2); los defectos de `LeerTipos` siguen sin clase de usuario y el kernel los da como `inesperado`, código 1 (FR-013). Ningún código de salida nuevo, ningún `panic` (research D4) |
| V · Simplicidad y dependencias fijadas (YAGNI) | Ninguna dependencia ni paquete nuevo; la superficie exportada se encoge (desaparecen 10 constantes, `CodigosDeTipo`, `Computos` y `PreceptosDeUnComputo`) y gana un campo; el cómputo descrito no se exporta (D1). Cada mecanismo, con su requisito («Trazabilidad»). Rechazados: un lenguaje de reglas, un esquema JSON de `data/plazos.yaml`, un lector de cómputos exportado y la combinación con `yaml.Node` (D7, D11) |
| VI · Un binario, convenciones de agente | Ninguna orden, bandera ni argumento nuevo. `--describe` deja de depender de los datos (FR-012): su esquema, del que salen la herramienta MCP `plazos_calcular` y la tabla de comandos, es el mismo con cualquier `data/plazos.yaml`; la descripción del verbo, la de la tabla de `SKILL.md`, no cambia |
| VII · Grafo y privacidad | `plazos` no toca el grafo ni la caché; ningún dato de nadie |
| VIII · Skills primero; el binario es la herramienta | Protege `plazos-y-territorio`: su paso 9 recibe lo mismo (FR-011) y ninguna `SKILL.md`, referencia ni eval cambia (FR-020); prepara la base para que el paquete de una materia (H9.1, H42) entre como datos. El cierre mide las evals de todas las skills (FR-025) |
| IX · Genericidad territorial, validación local | Cuántos municipios cuentan y cómo se nombran es dato del cómputo, no de un municipio; ningún caso por municipio. Se prueba con municipios cubiertos (Madrid `28079` en C1 a C10, Leganés en los guiones) y no cubiertos (2027 sin locales de Madrid en C2 y C4; `82001` sin fichero local en E11) |
| X · Genericidad por materia, validación administrativa | Es el objeto del hito: el motor no nombra ningún tipo, cómputo ni aviso de periodo, y lo vigila `TestNingunCodigoDePlazosEnGo` (FR-010); la abstracción se valida con dos implementaciones reales más un ensayo: `lpac` y `contencioso` como datos, el civil de la LEC (§C.9) y el cómputo de ensayo, que entra sin tocar Go (§C.10). Lo administrativo de hoy da lo mismo (US1) |
| R1 · `internal/core/**` no importa el kernel, adaptadores ni entrada y salida | El código y los tests del dominio solo usan la biblioteca estándar sin `os` ni `io`, `go.yaml.in/yaml/v3` y `internal/core/schema`, como hoy; los ficheros de prueba entran con `//go:embed` (research D8), y lo que lee ficheros (`os`, `go/parser`) vive en tests de `internal/app` (D6, D8) |
| R2 · solo `internal/httpx` importa `net/http` | Nada nuevo lo importa; la grabación usa `httpx.ConRaizDeGrabacion` |
| R3 · solo `internal/{cache,store,graph}` importan SQLite y `database/sql` | Nada nuevo los importa |
| R4 · solo `internal/cli` y `cmd/` llaman a `os.Exit` (`forbidigo`) | Ningún `os.Exit` nuevo |
| R5 · solo `internal/render` escribe en stdout; logs con `log/slog` a stderr (`forbidigo`) | Nada nuevo escribe en stdout; la salida de `calcular` sigue saliendo por el kernel y `render` |
| R6 · `internal/graph` no importa `internal/source` ni `internal/render` | `internal/graph` no cambia |
| R7 · solo el paquete del servidor importa el SDK de MCP | Nada nuevo lo importa; la herramienta `plazos_calcular` sale de `--describe`, como hoy |
| Garantía sin red de `TestArquitectura` (`paquetesSinRed`, `ficherosSinRed`, `internal/arch_test.go:700-721`) | `internal/core/plazos` y `internal/app/plazos.go` siguen en las listas y no ganan ninguna importación que llegue a `net`, `net/http` ni `internal/httpx`; lo nuevo que importa `data` es `internal/evals/plazos.go`, que no está en ellas |

Re-evaluado tras la fase 1: sin violaciones; Complexity Tracking, vacío. Tres desviaciones declaradas en «Decisiones»:
FR-016 obliga a tocar `TestGrabacionesSinSolape`, la raíz del cliente de `TestGrabarEvals` y la exigencia por conjunto
de `rangosDeLasBusquedasGrabadas` (`TestEsquemaDeNormas/rangos-grabados`), que FR-017 no nombra (research D10).

## Project Structure

### Documentation (this feature)

```text
specs/041-h36-el-motor-de/
├── spec.md
├── plan.md                         # este fichero
├── research.md                     # D1-D15, V1-V24, S2-S6, datos externos
├── data-model.md                   # cómputo, periodo, aviso de periodo, tipo, ficheros de prueba, grabaciones
├── quickstart.md                   # escenarios de comprobación
├── contracts/
│   ├── datos-de-plazos.md          # data/plazos.yaml completo, reglas de LeerTipos y sus subtests, ficheros de prueba
│   ├── api-del-dominio.md          # lo exportado que cambia, el cómputo, los avisos, los errores, internal/evals, el uso
│   ├── applet-calcular.md          # ayuda y esquema de calcular, goldens, TestDescribeDePlazosSinDatos, test de FR-010
│   ├── aceptacion.md               # los dos guiones congelados y la captura con el binario de main
│   └── grabacion.md                # la entrada del manifiesto, evidencias/h36/ y la raíz nueva de la unión
├── aceptacion/                     # la escribe la tarea [aceptacion]: plazos-de-hoy.txtar, describe.txtar
├── mutantes.md                     # registro del run: el mutante de cada control y lo que puso en rojo
└── tasks.md                        # /speckit-tasks
```

### Source Code (repository root)

```text
data/plazos.yaml                                   # forma de §C.1 y el cómputo civil
internal/core/plazos/
├── tipos.go                                       # Tipo.Da, el cómputo descrito, LeerTipos con las reglas de §C.2
├── computo.go                                     # un algoritmo con los ejes como datos
├── calcular.go                                    # municipios del cómputo; la unidad que el cómputo admite
├── avisos.go                                      # avisos de periodo; CodigosDeAviso(tipos), EtiquetaDeAviso(tipos, …)
├── errores.go                                     # mensaje de municipios compuesto; error de FR-007
├── salida.go                                      # pattern en Plazo.Tipo y Aviso.Codigo
├── testdata/computo-de-ensayo.yaml                # nuevo
├── testdata/tipos-de-prueba-civil.yaml            # nuevo
├── ensayo_test.go                                 # nuevo: TestComputoDeEnsayo
└── calcular_test.go, tipos_test.go, propiedades_test.go, avisos_test.go, salida_test.go   # FR-017
internal/app/
├── plazos.go                                      # la ayuda de §C.7
├── plazos_civil_test.go                           # nuevo: TestComputoCivil
├── plazos_describe_test.go                        # nuevo: TestDescribeDePlazosSinDatos
├── plazos_codigos_test.go                         # nuevo: TestNingunCodigoDePlazosEnGo
├── plazos_test.go                                 # FR-017: los enumerados y la ayuda
├── testdata/plazos/describe.json                  # regenerado
└── testdata/script/h39-describe.txtar             # líneas 39-43 (y su copia en specs/031-h39-c-omo-se/aceptacion/)
internal/evals/
├── plazos.go                                      # los tipos del binario para los avisos
├── grabaciones.go                                 # GrabacionesDeH36 en la unión
├── grabacion_test.go                              # //go:build grabacion: graba en $KITLEGAL_EVIDENCIAS
├── testdata/grabaciones.json                      # la entrada de la Ley 1/2000
├── grabaciones_test.go                            # FR-015 y la lista de la unión
└── formato_test.go                                # FR-017: el test de ComprobarCodigosDeAvisoDelPlazo
internal/skills/referencias_test.go                # FR-017: las constantes
schemas/plazo.json                                 # regenerado: la parte de plazos calcular
evidencias/h36/                                    # lo escribe solo grabar_datos
CHANGELOG.md                                       # FR-019
```

**Structure Decision**: la de `docs/ROADMAP.md` §2 y `CLAUDE.md`, sin paquetes nuevos: el dominio en
`internal/core/plazos`, la composición y los tests que leen ficheros en `internal/app`, el arnés de evals en
`internal/evals`.

## Diseño

1. **Los datos** (contracts/datos-de-plazos.md §1; FR-001 a FR-003, FR-014). `data/plazos.yaml` pasa a la forma de §C.1:
   `da` en cada tipo y los tres ejes en cada cómputo, con los preceptos de hoy intactos. Se fija desde el spec y lo
   comprueba lo grabado (research D14).
2. **El lector** (contracts/datos-de-plazos.md §2; FR-013). `LeerTipos` decodifica con `KnownFields(true)`, acumula los
   defectos de las diez reglas y los de hoy que no dependen de un código, y copia en cada tipo su cómputo y la posición
   de este en el fichero (D1, D3). Con las dos reglas de la duración que dependen de un código se va `fijaSuDuracion`
   (`tipos.go:38`), sus únicos usos, que si no quedaría sin uso y `unused` pondría `make lint` en rojo.
3. **El motor** (contracts/api-del-dominio.md §2 a §4; FR-004 a FR-008). `computar` cuenta por la unidad, con los periodos
   del cómputo (D2); `inhabil` suma los periodos `inhabil`; `da` decide la fecha; los avisos de periodo van en su sitio;
   el número de municipios y su mensaje salen del cómputo; la unidad que el cómputo no admite es el error nuevo, en el
   orden de D13.
4. **El esquema** (contracts/applet-calcular.md §1 y §2; FR-012). La ayuda de §C.7 y el `pattern` de §C.8 en las
   etiquetas; se regeneran `describe.json` y `schemas/plazo.json`, y cambian las líneas 39 a 43 de `h39-describe.txtar` y
   de su copia (D12).
5. **Las evals** (contracts/api-del-dominio.md §5; FR-009). `internal/evals/plazos.go` lee los tipos de `data.Plazos` una
   vez (D4).
6. **La grabación y el civil** (contracts/grabacion.md; FR-014 a FR-016). La entrada de la Ley 1/2000 y el destino de la
   grabación; tras `grabar_datos`, la raíz de H36 en la unión, el civil en `data/plazos.yaml` y `TestPreceptosDeLosPlazos`
   también sobre los cómputos sin tipos, sin cambiar lo que la skill lleva a leer (D9 a D11).
7. **Los tests** (FR-017, FR-021 a FR-024): los que cambian, solo los de FR-017, más los tres de D10 («Desviaciones
   declaradas») y la entrada de `probarLosDatosQueNoSeLeen` de H32 (`gates/supuestos.md`); los nuevos, los de
   «Trazabilidad».
8. **La documentación** (FR-019): `CHANGELOG.md`, en *Unreleased* > «Cambiado», las tres frases de FR-019.

**Orden y piezas indivisibles** (para `tasks.md`): la tarea `[aceptacion]` va antes de cualquier cambio de producto,
porque captura con el binario del árbol, que entonces es el de `main` (contracts/aceptacion.md §1). La tarea `[datos]`
de la grabación (§1 de contracts/grabacion.md) va antes de la que añade la raíz de H36 a la unión y el civil (§3), que
no puede ir antes de que `grabar_datos` grabe (research D9, V8). `data/plazos.yaml` en la forma nueva y el `LeerTipos`
que la lee van juntos, porque el lector de hoy rechaza las claves nuevas y el nuevo las exige. El esquema y sus goldens
(`describe.json`, `schemas/plazo.json`, `h39-describe.txtar` y su copia, los enumerados de `plazos_test.go` y
`salida_test.go`) van en una sola tarea `[datos]`. `TestNingunCodigoDePlazosEnGo` va cuando ya no queda ninguna
constante ni la ayuda de hoy, y existe el fichero de ensayo. `TestComputoCivil` y las propiedades del civil van con o
después del civil en `data/plazos.yaml`. `mutantes.md`, al final.

## Uso, de fuera adentro

Quién lee cada salida, cuánto mide y cuándo se apaga; el detalle, en contracts/api-del-dominio.md §6.

- **El `data` de `plazos calcular`** lo pide `plazos-y-territorio` en su paso 9 (`SKILL.md:59`), una vez por plazo de la
  pregunta —casi siempre una; dos con la reposición y el contencioso—, y lo usa entero. No cambia (FR-011): de 670 a
  853 bytes en los goldens de un plazo de un mes (research V19), unos 90 por festivo saltado y unos 100 por tramo de
  cobertura. `plazos` no lee la caché ni el grafo: con meses de uso diario, cientos de normas y miles de bloques
  consultados, la llamada siguiente mide lo mismo que la primera. Son de 1 a 2 KB por pregunta, sin crecer con lo
  acumulado.
- **Un aviso de periodo** añade una línea al `data`: el del art. 183 mide 335 bytes de texto con su prefijo, unos 380 con
  su objeto (V19), y es el único con los datos de hoy. Sale como mucho una vez por plazo y solo si el tramo pasa por el
  periodo; se apaga en la llamada cuyo tramo no lo toca. La skill lo copia tal cual.
- **El error de FR-007** llega a la skill en `data.mensaje`: 115 bytes en C7. Con los datos de hoy no se da; con un
  paquete que lo haga posible, una vez por llamada con una unidad que el cómputo no admite, y se apaga cuando la llamada
  pide una que admite.
- **El esquema de `plazos_calcular`** lo recibe el agente una vez por conexión con el servidor MCP, y quien pide
  `--describe` una vez por orden. Hoy, `describe.json` mide 8317 bytes; las ayudas pasan de 567 a 603 bytes (lo contó
  el spec) y las propiedades `tipo` (121 bytes compactas) y `codigo` (143) cambian sus enumerados por un `pattern` de 24
  caracteres: unos 8,2 KB, que fija el golden. Ya no crece con cada tipo ni cada aviso que entre.
- **`CodigosDeAviso` y `EtiquetaDeAviso`** los lee `internal/evals` una vez por proceso al juzgar las respuestas: los
  cuatro códigos de hoy, uno más por aviso de periodo que traiga un paquete.
- **El cómputo civil** no lo consume nadie desde el binario hasta H9.1 (US2.5): solo `TestComputoCivil` y las
  propiedades.

## Trazabilidad

| Mecanismo | Lo exige |
|---|---|
| `data/plazos.yaml` en la forma de §C.1, con `lpac` y `contencioso` como datos | FR-001, FR-002, FR-003 |
| El cómputo civil en `data/plazos.yaml`, sin tipos | FR-014, FR-015 |
| `Tipo.Da` y la regla 10 | FR-002, FR-005 |
| El cómputo descrito, copiado en cada tipo (municipios, unidades, periodos, preceptos) | FR-004, FR-006, FR-007, FR-008 |
| La posición del cómputo en el fichero | FR-009 (el orden de `CodigosDeAviso`) |
| `computar` con un algoritmo y los ejes como datos; `inhabil` con los periodos `inhabil`; `pasaPor` con un periodo | FR-004, FR-005; §C.3 |
| Los avisos de periodo y su orden | FR-008 |
| El mensaje de municipios compuesto con los textos del cómputo | FR-006 |
| El error de la unidad que el cómputo no admite, y su lugar en el orden | FR-007 |
| `CodigosDeAviso(tipos)`, `EtiquetaDeAviso(tipos, …)` | FR-009 |
| La lectura única de `data.Plazos` en `internal/evals/plazos.go` | FR-009 |
| La retirada de `Tipo…`, `Computo…`, `CodigosDeTipo`, `Computos`, `AvisoDiasDelArt183` y `PreceptosDeUnComputo` | FR-010; research D1 |
| La ayuda de §C.7 y el `pattern` de §C.8 en las etiquetas | FR-012 |
| `describe.json`, `schemas/plazo.json`, `h39-describe.txtar` y su copia, regenerados | FR-012, FR-017 |
| Las diez reglas de `LeerTipos` y sus defectos | FR-013 |
| La entrada de la Ley 1/2000 en el manifiesto y el destino `$KITLEGAL_EVIDENCIAS` de `TestGrabarEvals` | FR-016; research D10 |
| `GrabacionesDeH36` en `UnionDeGrabaciones` y su lugar en `TestGrabacionesSinSolape` | FR-016; research D10 |
| La búsqueda grabada exigida a la unión, no a cada conjunto, en `rangosDeLasBusquedasGrabadas` | FR-016; research D10 |
| La retirada de `fijaSuDuracion` con las reglas de la duración que dependen de un código | FR-010, FR-013; contracts/datos-de-plazos.md §2 |
| Las lecturas de los cómputos sin tipos en `verificarPreceptos` y las palabras de `a130` y `a133` | FR-015 (con FR-014: `lecturasDeLosTipos` y el oráculo no cambian; research D11) |
| `testdata/computo-de-ensayo.yaml` y `TestComputoDeEnsayo` | FR-022; US3.1 a US3.5 |
| `testdata/tipos-de-prueba-civil.yaml` y la combinación con su base | FR-023, FR-024 |
| `TestComputoCivil` | FR-023; US2.5 (FR-014), US5.1 (FR-013) |
| Las propiedades del civil en `TestPropiedadesDelComputo` | FR-024 |
| `TestNingunCodigoDePlazosEnGo` | FR-010; US3.6 |
| `TestDescribeDePlazosSinDatos` | FR-012; US4.1 |
| Los subtests de `TestLeerTipos` (uno por regla, dos defectos, la copia del contrato) | FR-013; US5.1 a US5.4 |
| Los cambios de `calcular_test.go`, `tipos_test.go`, `propiedades_test.go`, `avisos_test.go`, `salida_test.go`, `plazos_test.go`, `referencias_test.go` y `formato_test.go` | FR-017 |
| `aceptacion/plazos-de-hoy.txtar` y sus doce sobres capturados con `main` | FR-011, FR-021; US1 |
| `aceptacion/describe.txtar` | FR-012; US4.1 (datos de hoy), US4.2, US4.3 |
| `CHANGELOG.md` | FR-019 |
| `mutantes.md` | spec, «Controles»: cada control con la comprobación que lo pone en rojo |

## Controles mecánicos que el hito añade o toca

Según «Gates» de la constitución, capa 1:

- **`make ci`**, sin objetivos nuevos: `make test` ejecuta los tests nuevos (`TestComputoDeEnsayo`, `TestComputoCivil`,
  `TestDescribeDePlazosSinDatos`, `TestNingunCodigoDePlazosEnGo`) y los que cambian, y `TestEntregaDelHito` con los dos
  guiones de H36 activados; `make lint` pasa con la configuración de hoy y una sola entrada más, `periodos` en
  `ignore-rules` de `misspell` (T004: la clave de spec §C.1, que `misspell` lee como «periods»; gates/supuestos.md, T013),
  sin tocar ninguna regla (`depguard` R1 vigila también los tests del dominio, que no usan `os`); ningún `//nolint` ni
  `t.Skip`.
- **`make schema-check`** (`TestEsquemasPublicados`): `schemas/plazo.json` igual a lo que emite `--describe`, con el
  `pattern` en lugar de los dos `enum`. Los tests de contrato de la salida contra `schemas/plazo.json` no cambian.
- **`make skills-check`** y `TestRegiones`: las referencias generadas desde `data/plazos.yaml` salen iguales (FR-020).
- **`TestPreceptosDeLosPlazos`**: corrección de las citas contra la respuesta grabada del BOE, ahora también para los
  cómputos sin tipos (FR-015).
- **Fixtures**: nuevos, `internal/core/plazos/testdata/computo-de-ensayo.yaml` y `tipos-de-prueba-civil.yaml`; las
  grabaciones de `evidencias/h36/` (solo `grabar_datos`); regenerados, `internal/app/testdata/plazos/describe.json` y
  `schemas/plazo.json`; cambiados, `internal/app/testdata/script/h39-describe.txtar` y su copia congelada; todos en
  tareas `[datos]`. `caso-a.json`, `caso-b.json`, `caso-e.json` y los `h9-plazos-*.txtar` no cambian.
- **Test de arquitectura**: `TestArquitectura` sin cambios.
- **Exit codes**: ninguno nuevo; el 2 del error de FR-007 lo fijan C7 y E12, y cada error de §C.6 de hoy, su sobre
  capturado.
- **Los contratos de H9**: ningún fichero de `specs/023-h9-internal-core-plazos/` cambia (FR-018); lo que cambia de
  ellos lo reescribe spec §C y lo detallan los contratos de este plan.

**Mutantes** (`mutantes.md`, registro del run que escribe la última tarea; spec, «Controles»). Cada uno se aplica a un
árbol de prueba, se ve el control en rojo y se deshace; el registro dice qué test cayó:

| Control | Mutante | En rojo |
|---|---|---|
| El `data` de hoy (SC-001) | una palabra del texto del art. 183 en `data/plazos.yaml`; el silencio con `da: vencimiento`; el texto del municipio de `contencioso` | `h36-plazos-de-hoy`: `caso-15.json`, `caso-11.json` (y `TestCasosDePlazos`), `error-municipios-contencioso.json` |
| El esquema sin datos (SC-005) | la ayuda y los enumerados de hoy en las etiquetas | `TestDescribeDePlazosSinDatos`, `TestEsquemasPublicados`, `h36-describe`, `TestNingunCodigoDePlazosEnGo` |
| Ningún código en Go (SC-003) | `if tipo.Codigo == "silencio"` en `internal/core/plazos`; la ayuda de hoy | `TestNingunCodigoDePlazosEnGo` |
| Las propiedades (SC-004) | el motor cuenta como hábiles los días de un periodo `inhabil` | `TestPropiedadesDelComputo` (civil) |
| El civil (SC-002) | agosto `aviso` (con un aviso completo, que pide la regla 9); agosto `no-corre` (que la regla 5 solo admite en un cómputo de solo `meses`: o el cómputo pasa a solo `meses` o `LeerTipos` lo rechaza, y en los dos casos el test cae); sin el periodo del 24 de diciembre; agosto desde el `08-02`; el calendario sin los locales de Madrid | `TestComputoCivil`: C1, C3, C2, C9, C10 |
| El ensayo (SC-003) | un motor que solo reconoce agosto; que ignora `dia-siguiente` fuera del silencio; que no comprueba las unidades del cómputo; que da los avisos de periodo con un texto fijo | `TestComputoDeEnsayo`: E1, E5, E12, E7 |
| `LeerTipos` (SC-005) | quitar cada una de las diez reglas | su subtest de `TestLeerTipos` |
| Los preceptos del civil (SC-006) | `vigente_desde` del 133.1 con el 2010-05-04; el del 130.2 con el 2015-10-07 | `TestPreceptosDeLosPlazos` |
| La búsqueda grabada de la unión (desviación de D10) | `rutaDeLaBusqueda` de `internal/skills/normas_test.go` con otra ruta, que deja la unión sin búsquedas | `TestEsquemaDeNormas/rangos-grabados` |
| El aviso de más de la copia (SC-005) | la sustitución de `TestDescribeDePlazosSinDatos` sin su periodo `aviso` | `TestDescribeDePlazosSinDatos` (su `require`) |

**Aceptación e2e:** dos guiones testscript que escribe la primera tarea en `specs/041-h36-el-motor-de/aceptacion/` y
quedan congelados (contracts/aceptacion.md): `plazos-de-hoy.txtar` (US1 · FR-005, FR-006, FR-008, FR-011, FR-012,
FR-021 · SC-001), con los sobres de los casos 11, 15 y 16 y de los nueve errores de §C.6 capturados con el binario de
`main`, y `describe.txtar` (US4.1 con los datos de hoy, US4.2, US4.3 · FR-012 · SC-005). US2, US3, US5 y US6 no tienen
comportamiento visible desde el binario —ningún tipo pide el civil (US2.5), el ensayo es un fichero de prueba y
`LeerTipos` y la verificación de los preceptos son internos—: los cubren `TestComputoCivil`, `TestComputoDeEnsayo`,
`TestNingunCodigoDePlazosEnGo`, `TestLeerTipos` y `TestPreceptosDeLosPlazos`.

**Datos externos:** una fuente, `boe.legislacion-consolidada`, con fila revisada en `docs/SOURCES.md` de `main`
(2026-09-13). Manifiesto `internal/evals/testdata/grabaciones.json` (`"fuente": "boe.legislacion-consolidada"`), con la
entrada de la Ley 1/2000 y sus bloques `a130` y `a133`; test `//go:build grabacion` `TestGrabarEvals`
(`internal/evals/grabacion_test.go`), que pide con `internal/httpx`. Lo grabado es material de origen de la
verificación de `data/plazos.yaml` y queda en `$KITLEGAL_EVIDENCIAS` = `evidencias/h36/boe.legislacion-consolidada/`, con
su `evidencias/h36/manifiesto.json` (contracts/grabacion.md). Ningún fichero de `data/` se deriva de ello:
`data/plazos.yaml` lo fija el spec y lo comprueba lo grabado (research D14).

**Autonomía:** ninguna tarea pide nada a una persona ni publica la rama, abre la propuesta de cambio o mide CI y evals
remotas; las grabaciones las hace el paso `grabar_datos`, y el cierre, el workflow tras la revisión final.

## Controles de umbral

| Requisito | Umbral | Control | Dónde |
|---|---|---|---|
| SC-001, FR-011, FR-021 | 100 % de los casos de fecha 1 a 18 y 21 con la fecha y los avisos de `main`; 12 de 12 sobres capturados iguales salvo `fecha_consulta`, con su código; `caso-a.json`, `caso-b.json`, `caso-e.json` y los cinco `h9-plazos-*.txtar` sin cambiar | `h36-plazos-de-hoy` (`cmp stdout` de cada sobre y `test $c -eq N`), `TestCasosDePlazos`, el subtest `los-golden-de-los-casos` de `TestAppletPlazos`, los guiones de H9, `TestCalcular` y `TestComprobarPeticion`, en rojo con la primera diferencia | `ci:internal/app/e2e_test.go:TestEntregaDelHito`, `ci:internal/app/plazos_casos_test.go:TestCasosDePlazos`, `ci:internal/app/plazos_test.go:TestAppletPlazos`, `ci:internal/core/plazos/calcular_test.go:TestCalcular`, `ci:internal/core/plazos/calcular_test.go:TestComprobarPeticion` |
| SC-002, FR-023 | 10 de 10 filas de §C.9 (fecha, festivos saltados, avisos, o código, clase y mensaje) | `TestComputoCivil`, una fila por subtest | `ci:internal/app/plazos_civil_test.go:TestComputoCivil` |
| SC-003, FR-022 | 13 de 13 filas de §C.10 | `TestComputoDeEnsayo`, una fila por subtest | `ci:internal/core/plazos/ensayo_test.go:TestComputoDeEnsayo` |
| SC-003, FR-010 | 0 literales enteros y 0 etiquetas con un código de `data/plazos.yaml` o del fichero de ensayo | `TestNingunCodigoDePlazosEnGo`, un defecto por aparición | `ci:internal/app/plazos_codigos_test.go:TestNingunCodigoDePlazosEnGo` |
| SC-004, FR-024 | 0 contraejemplos en la rejilla de H9 y en la del civil | `TestPropiedadesDelComputo` | `ci:internal/core/plazos/propiedades_test.go:TestPropiedadesDelComputo` |
| SC-005, FR-012 | 3 de 3 salidas de `--describe` iguales a `describe.json`, con código 0; `schemas/plazo.json` igual a lo emitido; 0 `enum` y 1 `pattern` en `tipo` y en `avisos[].codigo` | `TestDescribeDePlazosSinDatos`, `TestEsquemasPublicados`, el subtest `los-enumerados-publicados` de `TestAppletPlazos` y `h36-describe` | `ci:internal/app/plazos_describe_test.go:TestDescribeDePlazosSinDatos`, `ci:internal/app/esquemas_test.go:TestEsquemasPublicados`, `ci:internal/app/plazos_test.go:TestAppletPlazos`, `ci:internal/app/e2e_test.go:TestEntregaDelHito` |
| SC-005, FR-013 | 10 de 10 reglas de §C.2 detectadas en su copia rota, y 2 de 2 defectos en la copia con dos | los subtests de `TestLeerTipos` | `ci:internal/core/plazos/tipos_test.go:TestLeerTipos` |
| SC-006, FR-015, FR-016 | 5 de 5 preceptos del civil grabados, con sus palabras clave y el `vigente_desde` de sus redacciones grabadas | `TestPreceptosDeLosPlazos` | `ci:internal/evals/grabaciones_test.go:TestPreceptosDeLosPlazos` |
| FR-009 | el esquema de eval enumera exactamente los códigos de `CodigosDeAviso` de los tipos del binario: 4 de 4, en su orden | `ComprobarCodigosDeAvisoDelPlazo` en el subtest `enumerado-de-los-avisos` | `ci:internal/evals/formato_test.go:TestFormatoDePlazos` |
| SC-007, FR-020 | 0 diferencias entre las regiones de `skills/` y lo que se genera desde `data/` | `TestRegiones` | `ci:internal/skills/regiones_test.go:TestRegiones` |
| SC-009, FR-017 | `make ci` con 0 fallos | `make ci` | `ci:Makefile:ci` |
| SC-008, FR-025 | ninguna clase de `plazos-y-territorio` por debajo de su umbral, en los dos modos | el job de evals del cierre, sus umbrales con `decide: true` | evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:sin_activar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:orden, evals:plazos-y-territorio:plazo_sin_fecha:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:cobertura_sin_declarar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:vigencia_sin_declarar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:redaccion_sin_fechar:claude-sonnet-5-5:herramienta, evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:orden, evals:plazos-y-territorio:lectura_sin_intentar:claude-sonnet-5-5:herramienta |
| SC-008, FR-025 | ninguna clase de `normativa` por debajo de su umbral | el job de evals del cierre | evals:normativa:sin_activar:claude-sonnet-5-5:orden, evals:normativa:sin_activar:claude-sonnet-5-5:herramienta, evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:orden, evals:normativa:redaccion_sin_fechar:claude-sonnet-5-5:herramienta, evals:normativa:afirma_lo_no_leido:claude-sonnet-5-5 |
| SC-008, FR-025 | ninguna clase de `consulta-juridica` por debajo de su umbral | el job de evals del cierre | evals:consulta-juridica:sin_activar:claude-sonnet-5-5:orden, evals:consulta-juridica:sin_activar:claude-sonnet-5-5:herramienta, evals:consulta-juridica:cobertura_sin_declarar:claude-sonnet-5-5:orden, evals:consulta-juridica:cobertura_sin_declarar:claude-sonnet-5-5:herramienta |
| SC-008, FR-025 | ninguna clase de `jurisprudencia` por debajo de su umbral | el job de evals del cierre | evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden, evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta, evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden, evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta, evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:orden, evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:herramienta, evals:jurisprudencia:afirma_lo_no_leido:claude-sonnet-5-5 |

Los nombres de los umbrales de las evals son los que publicó el cierre de H33 (research S6); el hito no cambia ningún
umbral, clase ni eval. Cada control de las filas de `ci:` tiene su mutante en `mutantes.md` (spec, «Controles»).

## Complexity Tracking

Vacío: ninguna violación que justificar.

## Decisiones

Todas en research.md, con la alternativa descartada: D1 (la API y el cómputo copiado en cada tipo), D2 (un algoritmo),
D3 (el orden de los avisos), D4 (las evals leen los tipos del binario), D5 (el esquema), D6 (el test de FR-010), D7 (los
tipos de prueba por texto), D8 (dónde va cada test), D9 (la grabación en `evidencias/h36/`), D10 (los tests que FR-016
obliga a tocar), D11 (los cómputos sin tipos en `TestPreceptosDeLosPlazos`), D12 (las copias de H39 sin su huella), D13
(el error nuevo y su orden), D14 (`data/plazos.yaml` desde el spec) y D15 (skills de Go).

**Desviaciones declaradas** (research D10), las tres por FR-016 y fuera de la lista cerrada de FR-017:

- `TestGrabacionesSinSolape` (`internal/evals/grabaciones_test.go`) fija hoy que la unión son H4, H5 y H9; FR-016 obliga
  a añadir la raíz de H36, y su lista esperada gana ese elemento y nada más, aunque FR-017 solo nombra, en ese fichero,
  lo de `TestPreceptosDeLosPlazos`. Sin ese cambio `make ci` sale en rojo.
- `TestGrabarEvals` (`internal/evals/grabacion_test.go`, `//go:build grabacion`), que FR-017 no nombra, construye hoy el
  cliente de cada raíz con la de su conjunto (`httpx.ConRaizDeGrabacion(raiz)`, línea 94); FR-016 pide lo grabado bajo
  `evidencias/h36/`, así que cambia solo esa raíz, que es `$KITLEGAL_EVIDENCIAS` cuando la variable no está vacía, y
  la de su conjunto si lo está, como hoy. Sin ese cambio la grabación no queda donde la pide FR-016.
- `rangosDeLasBusquedasGrabadas` (`internal/skills/normas_test.go:632-652`, de H5), que recorre `UnionDeGrabaciones()`
  en `TestEsquemaDeNormas/rangos-grabados` (`normas_test.go:531-536`), exige hoy a cada conjunto alguna búsqueda grabada
  (`require.NotEmpty`, línea 642). El conjunto de H36 solo guarda `robots.txt` y los bloques `a130` y `a133`: la
  búsqueda «enjuiciamiento civil» ya la sirve el de H5 (contracts/grabacion.md §2; V16), y grabarla otra vez la rechaza
  `TestGrabacionesSinSolape`, que solo deja compartir `robots.txt`. Cambia solo ese `require`, que pasa a la unión (entre
  todos los conjuntos hay alguna búsqueda) con su comentario; lo comparado, el enum de rango con los rangos de todas las
  búsquedas de la unión, no cambia. Sin ese cambio `make ci` sale en rojo en cuanto la raíz de H36 entra en la unión.
