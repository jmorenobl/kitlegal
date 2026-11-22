# Data Model: H36 · El motor de cómputo descrito por datos

Las entidades son las de «Key Entities» del spec, con la forma de §C.1. Aquí, cómo las lleva `internal/core/plazos`
(research D1) y qué regla de §C.2 valida cada campo. Los nombres de lo que no se exporta son orientativos: lo que fija
el contrato son los campos, las reglas y lo que se exporta.

## 1. Cómputo (`computos[]` de `data/plazos.yaml`)

Tipo sin exportar (`computoDescrito`), que `LeerTipos` decodifica con `KnownFields(true)` y copia en cada tipo que lo
nombra.

| Campo | YAML | Tipo Go | Regla |
|---|---|---|---|
| Codigo | `codigo` | `Computo` | único entre los cómputos (de hoy); gramática de §C.8 (regla 8) |
| Municipios | `municipios` | `[]string` | uno o dos textos no vacíos, en el orden en que se dan (regla 9) |
| Unidades | `unidades` | `[]Unidad` | no vacía, sin repetidos, cada una de `Unidades()` (regla 9) |
| Periodos | `periodos` | `[]periodo` | opcional; ver §2 (reglas 4 a 7 y 9) |
| Preceptos | `preceptos` | `[]Precepto` | uno o más, cada uno con sus cinco datos y `vigente_desde` `AAAA-MM-DD` (de hoy) |
| posicion | — | `int` | su índice en `computos`, que anota `LeerTipos` para el orden de `CodigosDeAviso` (research D3) |

Un cómputo puede no tener tipos (el civil). Sus municipios fijan cuántos pide cada tipo suyo y el mensaje de §C.6; sus
unidades, qué unidad admite un tipo sin duración (FR-007) y cuál puede fijar uno con duración (regla 2).

## 2. Periodo (`periodos[]`)

| Campo | YAML | Tipo Go | Regla |
|---|---|---|---|
| Desde, Hasta | `desde`, `hasta` | `string` `MM-DD` | un día que tienen todos los años, no el `02-29`; `hasta` anterior a `desde` cruza el año (regla 9) |
| Efecto | `efecto` | `string` | `inhabil`, `no-corre` o `aviso` (regla 9) |
| Precepto | `precepto` | `{Norma, Apartado string}` | los de un precepto del cómputo o, si el cómputo tiene tipos, de un precepto de cada uno (regla 4) |
| Aviso | `aviso` | `*avisoDelPeriodo` | solo y siempre con `efecto: aviso` (regla 9); ver §3 |

Además: un `no-corre` es un mes entero de 30 o 31 días, del día 1 al último, en un cómputo de solo `meses` (regla 5);
dos periodos de un cómputo no comparten ningún día, contados sobre los 365 días `MM-DD` del año (regla 6). El precepto
del periodo no añade nada a `data.preceptos` ni a la comprobación del ADR 0042: ya está entre los del cómputo o del
tipo, que son los que se comprueban.

Efectos en el cómputo (§C.3): `inhabil` hace inhábil cada día del periodo; `no-corre` mueve la partida y el
vencimiento (research D2); `aviso` no cambia el cómputo y da su aviso (§3).

## 3. Aviso de periodo (`aviso`)

| Campo | YAML | Regla |
|---|---|---|
| Codigo | `codigo` | gramática de §C.8 (regla 8); único en el fichero y distinto de los tres del motor (regla 7) |
| Etiqueta | `etiqueta` | no vacía, de una línea, sin `:`; única en el fichero y distinta de las tres del motor (reglas 7 y 9) |
| Texto | `texto` | no vacío y de una línea (regla 9) |

En la salida, `{codigo: <codigo>, texto: "⚠ <etiqueta>: <texto>"}`, una vez por plazo cuyo tramo —del día siguiente
al de inicio a la fecha que se da— pasa por algún día del periodo (FR-008).

## 4. Tipo (`tipos[]`)

Tipo exportado `Tipo` (research D1).

| Campo | YAML | Tipo Go | Regla |
|---|---|---|---|
| Codigo | `codigo` | `string` | único; gramática de §C.8 (regla 8) |
| Nombre | `nombre` | `string` | no vacío (de hoy) |
| Computo | `computo` | `Computo` | el código de un cómputo de `computos` (regla 1) |
| Duracion | `duracion` | `*Duracion` | opcional: `cantidad` ≥ 1 y una unidad que admite su cómputo (regla 2), con un precepto `fija: duracion` (de hoy) |
| Da | `da` | `string` | `vencimiento` o `dia-siguiente`, obligatoria (regla 10) |
| Preceptos | `preceptos` | `[]Precepto` | uno o más, salvo con `unidades` (de hoy) |
| Unidades | `unidades` | `map[Unidad]Precepto` | opcional y solo sin `duracion`: exactamente las unidades de su cómputo, cada una con su precepto (regla 3) |
| PreceptosDelComputo | — | `[]Precepto` | los de su cómputo, que copia `LeerTipos` (de hoy) |
| computo | — | `computoDescrito` | su cómputo entero, que copia `LeerTipos` (research D1) |

Un tipo con `Duracion` la fija: rechaza `--duracion` y `--unidad` (§C.6, de hoy). Uno sin ella la recibe con una unidad
del vocabulario que su cómputo admita (FR-007).

## 5. Aviso del motor

Los tres de H9, fijos en `internal/core/plazos`: `faltan-festivos-nacionales-y-autonomicos`, `faltan-festivos-locales`
y `festivo-de-una-parte`, con sus constantes, sus etiquetas y sus textos de hoy. No son datos ni códigos de
`data/plazos.yaml`.

## 6. Ficheros de prueba de `internal/core/plazos/testdata/`

| Fichero | Qué es | Lo usan |
|---|---|---|
| `computo-de-ensayo.yaml` | un `data/plazos.yaml` completo con el cómputo `ensayo` y los tipos `ensayo-fijo` y `ensayo-libre` de §C.10 (contracts/datos-de-plazos.md §3) | `TestComputoDeEnsayo` (empotrado), `TestNingunCodigoDePlazosEnGo` (sus códigos) |
| `tipos-de-prueba-civil.yaml` | solo `tipos:` con `prueba-civil-dias` y `prueba-civil-meses`, sin duración, `da: vencimiento`, del cómputo `civil` (contracts/datos-de-plazos.md §4) | `TestComputoCivil` (`internal/app`, con `data/plazos.yaml`) y `TestPropiedadesDelComputo` (con la copia del contrato) |

## 7. El conjunto de grabaciones de H36

`evidencias/h36/boe.legislacion-consolidada/`: lo que grabe `grabar_datos` de `boe articulo BOE-A-2000-323 a130` y
`a133`, y `robots.txt`; `evidencias/h36/manifiesto.json`, su huella. Es la raíz `GrabacionesDeH36` de
`UnionDeGrabaciones`, la cuarta, detrás de H4, H5 y H9 (contracts/grabacion.md).

## Transiciones de estado

Ninguna: los datos se leen una vez por applet y no cambian; el cálculo no guarda nada.
