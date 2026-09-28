# Contrato: `boe-legislacion` v0.1.1, la eval 19 y el formato común de eval

FR-040 a FR-055, FR-093, FR-094, FR-097. Modelo: [../data-model.md](../data-model.md) §7-§8. Decisiones:
[../research.md](../research.md) D12-D16.

## 1. Formato común de eval (`schemas/eval.yaml.json`, `internal/evals/formato.go`)

Cambios del esquema, en una tarea `[datos]` (FR-082):

- `$defs.clase-de-hallazgo`: `{"enum": ["version-obsoleta"]}`.
- Propiedad `hallazgos`: `{"type": "array", "minItems": 1, "items": {"$ref": "#/$defs/clase-de-hallazgo"}}`, y en el
  `else` (eval que no activa la skill) entra en el `anyOf` de lo prohibido, como `avisos`.
- `$defs.comando-comprobacion` gana `norma` opcional (`{"$ref": "#/$defs/norma"}`); `required` sigue siendo `applet` y
  `verbo`.
- `avisos` y su enumerado no cambian (FR-054).

En Go: `Eval.Hallazgos []string` (`yaml:"hallazgos"`); `ComandoEsperado.Norma` ya existe y ahora también lo lleva una
comprobación. Las demás evals se leen y se juzgan igual. Un valor de `hallazgos` fuera del enumerado es un fichero mal
formado (el esquema lo rechaza), como un aviso desconocido en H5.1.

## 2. Forma fija y juicio (`internal/evals`)

- Forma fija de una etiqueta: `formaFija(etiqueta)` —extraída del constructor de `formasDeAviso`
  (`avisos.go:48-64`) y usada por avisos y hallazgos—: `⚠` con su selector de presentación opcional (U+FE0E/U+FE0F), un
  separador de blancos Zs, tabuladores, `*` o `_`, cada palabra de la etiqueta sin distinguir mayúsculas (V14) con al
  menos un blanco entre dos, otro separador y `:`, en la misma línea (FR-051).
- `ExtraerHallazgos(texto) []string`: las clases de `grafo.EtiquetasDeHallazgo()` cuya forma lleva el texto, en el
  orden de las clases, sin repetir.
- `ComprobarFormasDeHallazgo(texto) error`: un error por clase etiquetada cuya forma falta, que nombra la clase:
  `falta la forma fija del hallazgo version-obsoleta: ⚠ REDACCIÓN MODIFICADA:` (FR-045).
- `ComprobarClasesDeHallazgo(esquema) error`: el enumerado de `hallazgos` del esquema compilado es exactamente el
  conjunto de clases etiquetadas (como `ComprobarCodigosDeAviso`).
- `Juzgar`: `repartirHallazgos(eval.Hallazgos, ExtraerHallazgos(respuesta))` → `HallazgosEncontrados` /
  `HallazgosAusentes`, con el motivo `forma de hallazgo ausente: <clase>` por cada ausente; la sesión pasa solo sin
  ausentes. `satisface` de una comprobación con `Norma` exige además `esDeLaNorma(invocacion, comando.Norma)`; su texto
  es `graph check <norma>` (sin norma, `graph check`, como en H7).
- Tests de `Juzgar` (FR-094, SC-008): respuesta con `⚠ REDACCIÓN MODIFICADA: …` → encontrada y pasa; sin ella →
  ausente y no pasa; con «la redacción ha cambiado» sin la forma → ausente y no pasa; con `**⚠️ Redacción modificada**:`
  → encontrada (tolerancias). Y una comprobación con otra norma no satisface la de la eval.

## 3. Comprobaciones mecánicas en `make ci` (`skills-check`, `TestEvalsDelRepositorio`)

- Subtest `hallazgos-de-la-skill`: `ComprobarFormasDeHallazgo(SKILL.md de boe-legislacion)` sin error (FR-045, FR-093).
- Subtest `hallazgos-del-esquema`: `ComprobarClasesDeHallazgo(esquema de eval)` sin error.
- `TestEtiquetasDeHallazgo` (`internal/evals`): ninguna etiqueta de hallazgo coincide con una de
  `boe.EtiquetasDeAviso()` (FR-045), y `version-obsoleta` es la única clase etiquetada; `fuente-caducada`, no.
- Unit tests de `ComprobarFormasDeHallazgo`: pasa sobre un texto con la forma y falla nombrando `version-obsoleta`
  sobre uno sin ella (SC-008).

## 4. `skills/boe-legislacion/SKILL.md` (v0.1.1)

Solo cambia lo que piden FR-040 a FR-046 (FR-047); la forma de la cita, la de los avisos de vigencia y las reglas 1-6
no cambian; sigue por debajo de 300 líneas y sin nombrar evals, el job ni modelos.

- **Paso 2**: se retira la comprobación de la memoria antes de leer (el bloque «Resuelto el `BOE-A-…`… antes de leer
  ninguno de sus bloques, comprueba la memoria de consultas» y su `kitlegal graph check --json`).
- **Paso 3**: se añade «Lee cada bloque una sola vez por pregunta: una segunda lectura del mismo bloque apagaría lo que
  la memoria de consultas tiene que decirte (más en «Memoria de consultas»).» (FR-041).
- **Paso 5**, primer punto, sustituido por: «Cuando ya no quede nada por leer, y antes de redactar la respuesta,
  comprueba la memoria de consultas una vez por cada norma cuyos bloques vas a citar, con esa norma y los bloques de
  ella que has leído, y traslada lo que encuentre como dice «Memoria de consultas»:» con el ejemplo
  `kitlegal graph check BOE-A-2015-10565 a21 --json` (FR-040). Nunca sin argumentos ni antes de leer.
- **«Memoria de consultas»**, reescrita:
  - `kitlegal` recuerda en local los bloques que ha leído con `kitlegal boe articulo` o `articulos` y qué redacción vio
    cada lectura. `kitlegal graph check <norma> <bloques>... --json` devuelve en `data.hallazgos` los de esa norma y
    esos bloques.
  - `version-obsoleta`: la redacción del bloque ha cambiado desde la lectura anterior. Trasládalo con su forma fija:
    `⚠`, la etiqueta `REDACCIÓN MODIFICADA` y dos puntos, en la misma línea, con las dos fechas de vigencia del
    hallazgo detrás —la de la redacción superada (`fecha_vigencia`) y la de la que acabas de leer
    (`fecha_vigencia_reciente`)—. Ejemplo:
    `⚠ REDACCIÓN MODIFICADA: la redacción con fecha de vigencia 20161002, la que se consultó antes, ha sido sustituida por la de 20250101, que es la que se cita.`
    Decirlo con otras palabras no lo traslada (FR-042).
  - `fuente-caducada` no se traslada: la respuesta cita el texto que acabas de leer, que la caché no sirve pasada su
    vigencia (FR-043).
  - Con código 0 y sin `version-obsoleta`, no digas nada de la memoria de consultas (FR-044); con otro código, la regla
    7 (FR-046).
  - Se retira «Un hallazgo no es un aviso de vigencia: no lleva la forma fija de los avisos.»; se dice en su lugar que
    la etiqueta de `version-obsoleta` no es la de ningún aviso.
- **Comandos**: la tabla generada (`make skills-sync`) da `kitlegal graph check [<norma> [<bloques>...]]`, la
  descripción de contracts/applet-graph.md §1 y `objeto con norma, bloques, version-obsoleta, fuente-caducada,
  omitidos, hallazgos`. La frase de los códigos «(por ejemplo, un `world.db` que no se puede leer en `kitlegal
  graph`)» se queda.
- **Regla 7**: se queda (H7 FR 082).

Invocaciones de `kitlegal` por pregunta con v0.1.1: las de `boe` que ya hacía y **una** de `graph check` por norma
citada (una en la mayoría; H7 pedía dos, sin argumentos, ≈ 6,1 MB con la medida).

## 5. Eval 19 (`evals/boe-legislacion/19-lpac-articulo-21-redaccion-cambiada.yaml`)

```yaml
# Eval informativa (ADR 0016): el grafo de la sesión ya registró una lectura del art. 21 LPAC con una redacción
# anterior (derivada de la grabación, fecha de vigencia 20151002) y la caché sirve la grabada (20161002). La sesión
# tiene que leer el bloque con kitlegal boe articulo, comprobar la memoria con kitlegal graph check y la norma, no
# pedir graph show, citar el bloque y trasladar el cambio de redacción con la forma fija ⚠ REDACCIÓN MODIFICADA:.
pregunta: "Ya te pregunté hace tiempo por el artículo 21 de la Ley 39/2015. ¿Qué dice ahora?"
activa: true
informativa: true
grafo_previo:
  grabaciones: lpac-a21-version-anterior
  comandos:
    - applet: boe
      norma: BOE-A-2015-10565
      bloque: a21
comandos:
  - applet: boe
    norma: BOE-A-2015-10565
    bloque: a21
  - applet: graph
    verbo: check
    norma: BOE-A-2015-10565
prohibidos:
  - applet: graph
    verbo: show
citas:
  - norma: BOE-A-2015-10565
    bloque: a21
hallazgos:
  - version-obsoleta
```

Sigue informativa; las reglas del conjunto no cambian (FR-055). Estado previo sin red: research D15.

## 6. Informe del job de evals (`internal/evals/informe.go`)

- `TasaDelInforme` gana `formas` (`[]string`, nunca `null`): la forma fija literal de cada clase de `hallazgos` de la
  eval, `⚠ ` + etiqueta + `:`; en la 19, `["⚠ REDACCIÓN MODIFICADA:"]`.
- `informe.md`: la tabla de tasas gana la columna «Formas exigidas» y la de sesiones, «Hallazgos encontrados» y
  «Hallazgos ausentes».
- `ResultadoDeEval` gana `hallazgos_encontrados` y `hallazgos_ausentes` (listas, nunca `null`).
- Lo demás del informe no cambia; `scripts/workflow/informe.sh` tampoco.

## 7. `CHANGELOG.md` (*Unreleased*, FR-048, FR-097)

H7 no ha salido en ninguna release, así que *Unreleased* describe el comportamiento final: se corrigen las entradas de
H7 que este hito deja falsas —la creación de `world.db` «en un temporal», el rechazo de «una arista sin sus
extremos» «sin tocar nada», los códigos de `graph` («es un directorio, tiene una transacción interrumpida…, que el
verbo… no modifica»), «sin permiso de escritura» en la entrega fallida, `graph check` («una lista de hallazgos»),
`boe-legislacion` v0.1 y el formato de eval— y se añaden: el ámbito y la cota de `check` y su `data`; el apagado de
`version-obsoleta`; `fuente-caducada` solo sobre lo vigente; la salida legible de `graph`; la regla genérica y lo
retirado; y `boe-legislacion` v0.1.1 (una comprobación por norma citada y la forma `⚠ REDACCIÓN MODIFICADA:`).
