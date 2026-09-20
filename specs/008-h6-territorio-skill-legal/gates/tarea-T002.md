# T002 — los dos esquemas y el boletín estatal los cierra el ejecutor; los 19 ficheros de comunidad los escribe la persona en la pausa (intento 1)

**Estado**: marcada `[X]` con `make ci` en verde. Están escritos y verificados `schemas/territorio-comunidad.yaml.json`,
`schemas/territorio-estado.yaml.json` y `data/territorio/estado.yaml`. `data/territorio/comunidades/` **no existe
todavía**: sus 19 ficheros los escribe una persona en la pausa humana que este mismo commit dispara, con el manifiesto
de abajo («Lo que hace la persona en la pausa»), que ya trae derivado de `municipios.yaml` todo lo que la tarea permite
derivar.

## Por qué el reparto es este

La tarea reparte su material en dos mitades, y lo dice en su propia línea:

- **Del fichero congelado**: «el código de cada comunidad y las provincias que declara se copian de las columnas
  `comunidad` y `provincia` de `municipios.yaml`, que T001 dejó congelado». Eso lo hace el ejecutor y está abajo, en la
  tabla del manifiesto: 19 códigos, sus 52 provincias y el recuento de municipios de cada una.
- **De la persona**: «el nombre y el régimen de cada comunidad, y el código, el nombre, la dirección y el motivo del
  BOCM, los aporta y los revisa una persona en la pausa». Ninguno de esos datos está en el repositorio —`municipios.yaml`
  solo trae códigos y nombres de municipio— y escribirlos sería escribirlos de memoria, que la batería de tasks.md
  prohíbe. Descargarlos tampoco: esta tarea «no descarga nada» (FR-043, ADR 0017).

Un fichero de comunidad sin esos datos no es un fichero a medias: no valida contra su propio esquema (`nombre` no vacío,
`regimen` obligatorio). Por eso no se escribe ninguno: lo que entra es el esquema que los 19 tendrán que cumplir.

**Por qué la tarea se marca igualmente** (misma mecánica que en T001, `gates/tarea-T001.md`): la pausa `gate_humano_datos`
solo se dispara cuando `clasificar_datos` ve un fichero nuevo bajo `schemas/` en el diff **ya commiteado**, y
`estado_cierre` solo commitea una tarea marcada `[X]`. Dejarla `[ ]` a la espera de los 19 ficheros haría imposible la
pausa en la que se escriben. La cláusula «si ese material no está disponible, la tarea se detiene sin marcarse» es la
guarda que impide que el ejecutor invente nombres, regímenes y direcciones; no una condición que cierre la única puerta
por la que el material entra.

## Lo que entra en este commit

### `schemas/territorio-estado.yaml.json`

| Clave | Forma |
|---|---|
| raíz | `additionalProperties: false`; obligatorias `fecha`, `source`, `boletin` |
| `fecha` | texto con `pattern` `^[0-9]{4}-[0-9]{2}-[0-9]{2}$` |
| `source` | texto no vacío |
| `boletin` | objeto, `additionalProperties: false`; obligatorias `codigo` y `nombre` (no vacíos) y `url` (`format: uri`, no vacía) |

Sin `motivo`: el boletín estatal no cubre dos niveles, y data-model §3.3 no se lo da.

### `schemas/territorio-comunidad.yaml.json`

| Clave | Forma |
|---|---|
| raíz | `additionalProperties: false`; obligatorias `fecha`, `source`, `codigo`, `nombre`, `regimen`, `provincias` |
| `fecha` | texto con `pattern` `^[0-9]{4}-[0-9]{2}-[0-9]{2}$` |
| `source` | texto no vacío |
| `codigo` | texto con `pattern` `^[0-9]{2}$` |
| `nombre` | texto no vacío |
| `regimen` | `enum` `comun` \| `foral`, **obligatorio en las 19** (FR-055) |
| `provincias` | objeto con `minProperties: 1`, `propertyNames.pattern` `^[0-9]{2}$` y nombre no vacío |
| `boletines` | **opcional** (ausente = territorio no configurado); `additionalProperties: false`, `minProperties: 1`, con `autonomico` y `provincial` |
| cada boletín | `additionalProperties: false`; obligatorias `codigo`, `nombre`, `url`; `motivo` opcional y no vacío si está |

Tres decisiones de forma, que no estaban escritas en los artefactos y aquí quedan dichas:

1. **`autonomico` y `provincial` son opcionales dentro de `boletines`, con `minProperties: 1`.** La cobertura los mide
   por separado (`boletin_autonomico` y `boletin_provincial`, cada uno `configurado` \| `no-configurado`, data-model
   §2.4), así que una comunidad puede traer solo el autonómico —el caso de toda comunidad con varias provincias, donde
   un único `provincial` no significaría nada—. Exigir los dos rompería FR-053 en la siguiente comunidad que se añada.
   Un `boletines: {}` sí se rechaza: para eso está la ausencia de la clave.
2. **`motivo` es opcional y, si está, no vacío.** El ejemplo de data-model §3.4 lo trae solo en `provincial`, y §2.3 lo
   define como «vacío salvo cuando el mismo boletín cubre dos niveles». Un `motivo: ""` explícito es ruido y se rechaza.
3. **`codigo` del boletín es texto no vacío, sin `pattern`.** El vocabulario de códigos de boletín no lo fija ningún
   artefacto del hito; inventar una gramática ahora podría rechazar un código legítimo al añadir un territorio (FR-053).

### `data/territorio/estado.yaml`

```yaml
fecha: "2026-09-20"
source: data/territorio/estado.yaml
boletin:
  codigo: "BOE"
  nombre: "Boletín Oficial del Estado"
  url: "https://www.boe.es/"
```

Nada de esto es memoria: la forma y los tres valores son los de data-model §3.3, `source` es la ruta del propio fichero
(el vocabulario de `source` de data-model §2.5 para lo que fija la configuración, y lo que el subtest `fuentes` de T007
comprobará contra el árbol), la dirección es la del sitio que ya declara la fila `boe.legislacion-consolidada` de
`docs/SOURCES.md` y `fecha` es la de su redacción, como dice el contrato de datos §1 para este fichero.

## Verificación (este intento)

Ningún control mira todavía estos esquemas —eso llega en T007—, así que se verificaron con un **arnés temporal**
(`internal/skills/territorio_config_arnes_temporal_test.go`, borrado antes de `make ci` y no commiteado; patrón de T001)
sobre el lector común, `skills.CompilarEsquema` + `skills.ValidarDocumentoYAML`. **39 subtests, todos en verde, ninguno
saltado**, primero en rojo por no existir los esquemas:

- **Acepta**: el ejemplo de data-model §3.3; `data/territorio/estado.yaml` tal como queda en el repositorio; el ejemplo
  de data-model §3.4 con los dos boletines; una comunidad de tres provincias, `foral` y sin `boletines`; y una con solo
  el autonómico configurado.
- **Rechaza**: 18 documentos en estado y 35 en comunidad, cada uno con su defecto —fecha sin comillas y sin forma; raíz
  sin cada obligatoria, con clave de más y con clave repetida; `source` vacío; `codigo` de una y de tres cifras, con
  letras y sin comillas; `nombre` vacío; `regimen` ausente y fuera del enumerado; `provincias` ausente, vacío, como
  lista, con clave de tres cifras, con letras, sin comillas y con nombre vacío; `boletines` vacío, como lista y con un
  nivel desconocido; boletín sin `codigo`, sin `nombre` o sin `url`, con cada uno vacío, con `motivo` vacío, con clave
  de más y con una `url` que no es una dirección—. La tabla se comprueba dos veces: leída en el tipo del fichero y
  leída en `map[string]any`, para que el rechazo sea del esquema y del lector, nunca de la lectura final en el tipo.
- **Mutantes** (32: 11 en estado y 21 en comunidad), cada uno una copia del esquema con **una** restricción quitada, en
  memoria: para cada copia se exige el **conjunto exacto** de casos de la tabla que pasa a aceptar. Cada restricción
  vigila lo suyo y nada más.

Dos hallazgos del ejercicio de mutantes, ambos anotados en el arnés antes de borrarlo:

- **El `pattern` de `fecha` es también lo que caza el timestamp.** Sin comillas, el lector normaliza `2026-09-20` a
  `"2026-09-20T00:00:00Z"`, que sigue siendo texto: lo que lo rechaza es la forma, no el tipo. Confirma el aviso de
  `gates/tarea-T001.md` y vale para los 19 ficheros de comunidad.
- **`minLength: 1` en `url` no sostiene ningún caso por sí solo**: `format: uri` ya rechaza la cadena vacía. Se mantiene
  porque es la forma que ya tienen las `url` de `schemas/norma.json`, no porque haga falta.

`make ci` en primer plano tras borrar el arnés: **en verde**.

## Lo que hace la persona en la pausa

Escribir los **19 ficheros** de `data/territorio/comunidades/`, uno por comunidad autónoma y ciudad autónoma, con esta
forma (data-model §3.4) y **todo entrecomillado** como arriba:

```yaml
fecha: "AAAA-MM-DD"          # la de su redacción
source: ine.municipios
codigo: "CC"                 # el mismo que da nombre al fichero: CC.yaml
nombre: "…"                  # el oficial, del INE
regimen: comun               # o foral
provincias:
  "PP": "…"                  # los códigos, ya derivados abajo; los nombres, del INE
```

Solo `13.yaml` lleva además `boletines`, con el BOCM **a la vez** como autonómico y como provincial y el motivo de la
equivalencia escrito en el propio fichero (FR-052):

```yaml
boletines:
  autonomico: {codigo: "…", nombre: "…", url: "…"}
  provincial: {codigo: "…", nombre: "…", url: "…", motivo: "Comunidad uniprovincial: …"}
```

Ninguna otra comunidad trae `boletines`, **tampoco las demás uniprovinciales**: la equivalencia no se deduce, se
declara (data-model §2.3; FR-051).

### Manifiesto: lo que ya está derivado de `municipios.yaml`

Códigos y provincias tomados de las columnas `comunidad` y `provincia` del fichero congelado por T001, contando sus
8.132 filas. **Las 52 provincias aparecen una sola vez, `01` a `52` sin huecos**, que es la integridad que `Cargar`
exigirá en T006 (data-model §2.1, punto 1). El recuento de municipios está para que la persona pueda reconocer cada
comunidad sin más ayuda que el fichero.

| Fichero | `provincias` (códigos) | municipios |
|---|---|---|
| `01.yaml` | `"04"`, `"11"`, `"14"`, `"18"`, `"21"`, `"23"`, `"29"`, `"41"` | 785 |
| `02.yaml` | `"22"`, `"44"`, `"50"` | 731 |
| `03.yaml` | `"33"` | 78 |
| `04.yaml` | `"07"` | 67 |
| `05.yaml` | `"35"`, `"38"` | 88 |
| `06.yaml` | `"39"` | 102 |
| `07.yaml` | `"05"`, `"09"`, `"24"`, `"34"`, `"37"`, `"40"`, `"42"`, `"47"`, `"49"` | 2248 |
| `08.yaml` | `"02"`, `"13"`, `"16"`, `"19"`, `"45"` | 919 |
| `09.yaml` | `"08"`, `"17"`, `"25"`, `"43"` | 947 |
| `10.yaml` | `"03"`, `"12"`, `"46"` | 542 |
| `11.yaml` | `"06"`, `"10"` | 388 |
| `12.yaml` | `"15"`, `"27"`, `"32"`, `"36"` | 313 |
| `13.yaml` | `"28"` | 179 |
| `14.yaml` | `"30"` | 45 |
| `15.yaml` | `"31"` | 272 |
| `16.yaml` | `"01"`, `"20"`, `"48"` | 252 |
| `17.yaml` | `"26"` | 174 |
| `18.yaml` | `"51"` | 1 |
| `19.yaml` | `"52"` | 1 |

Cinco de esos códigos ya están confirmados contra la hoja del INE en la pausa de T001 (`gates/tarea-T001.md`, «S1 y S5
quedan resueltos»): **13** es la Comunidad de Madrid —la única que lleva `boletines` en este hito—, **15** Navarra,
**16** País Vasco, **18** Ceuta y **19** Melilla. Los catorce restantes los nombra la persona desde la misma hoja.

### Comprobaciones antes de aprobar la pausa

1. Los 19 ficheros existen, cada uno con el `codigo` que da nombre a su fichero.
2. Cada uno declara exactamente las provincias de su fila de la tabla, con su nombre.
3. Los 19 traen `regimen`, y solo `13.yaml` trae `boletines`.
4. Cada fichero valida contra `schemas/territorio-comunidad.yaml.json` (con el lector común; en T007 lo hará `make ci`).
5. Confirmarlo en la rama del hito (`git add data/territorio/comunidades && git commit`) y aprobar la pausa. T003
   arranca desde ese commit.

## Notas

- No se ejecutó ninguna descarga, ninguna grabación ni ninguna orden de red en esta tarea.
- El arnés temporal comprobó los mutantes **en memoria**, sin escribir ninguna copia fuera del repositorio: no quedan
  ficheros sueltos de esta tarea en ningún sitio.
- Sigue en pie la mejora de proceso anotada en T001, que afecta igual a T003 y T015: la línea de una tarea `[datos]`
  cuya entrega completa una persona debería decir la secuencia explícita en lugar de la cláusula «se detiene sin
  marcarse».
