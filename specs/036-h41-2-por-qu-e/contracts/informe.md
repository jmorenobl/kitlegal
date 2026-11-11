# Contrato: las skills activadas en el informe del job

FR-001 a FR-006; SC-001, SC-002. Research D1 a D6; data-model §1 y §2. Lo que no se nombra aquí es el informe de hoy, sin
cambiar un byte.

## 1. Qué se cuenta, por sesión

La lista de una sesión son sus bloques `tool_use` de `Skill`, en el orden del transcript, cada uno con su `input.skill`
tal cual y si va **antes** de la primera herramienta de kitlegal: en un turno anterior al de la primera, o en cualquiera si
la sesión no tiene ninguna; en el mismo turno, no. Turno, el de la tabla de H39: lo empieza el primer `assistant` y cada
`assistant` que llega tras un `user` con algún `tool_result` y lleva otro `message.id`. Primera herramienta de kitlegal:

| Modo de la sesión | Lo es | No lo es |
|---|---|---|
| `herramienta` | `mcp__kitlegal__boe_articulo`, `boe_articulo`: un uso cuyo nombre, o lo que sigue a su último `__`, es una herramienta del registro | `ToolSearch`, `Read`, `Skill` |
| `orden` | `Bash` `kitlegal boe articulo BOE-A-2003-23186 a66 --json` (la consulta de `evals/normativa/04-lgt-prescripcion.yaml`): una orden simple que invoca `kitlegal` con un verbo que da herramienta y consulta | `Bash` `which kitlegal`, `kitlegal --help`, `kitlegal skills list` |
| sin binario ni servidor (`""`) | nada | todo: toda activación es «antes» |

Es la misma regla con la que cuenta la tabla del laboratorio: las dos usan `activacionesDeLosUsos` (data-model §1).

## 2. `informe.json`

Cada elemento de `evals` lleva, detrás de `activada`, `activaciones`. Ejemplo, una sesión del modo herramienta que activa
`consulta-juridica` antes de `boe_articulo` y `normativa` después:

```json
      "activa": true,
      "activada": true,
      "activaciones": [
        {
          "skill": "consulta-juridica",
          "antes": true
        },
        {
          "skill": "normativa",
          "antes": false
        }
      ],
      "por_materia": false,
```

| Sesión | `activaciones` |
|---|---|
| leída, sin ningún `tool_use` de `Skill` | `[]` |
| leída, con alguno | la lista, una entrada por bloque |
| ilegible por su eval, su traza o su modo, pero con el transcript leído | la lista, como `activada` |
| sin leer (un fichero de la sesión que falta o que no se lee), o cuya lista no se pudo leer | sin la clave |

`activada`, cada tasa, cada umbral, cada motivo y el veredicto se calculan como hoy: una sesión con `activada` falso cuenta
en `sin_activar` aunque su lista tenga otra skill (FR-003, FR-004). Un error al leer la lista es un motivo más,
`sesión ilegible: <error>`, que solo da un transcript que Claude Code no escribe (research D5).

## 3. `informe.md`

En la sección de cada sesión, detrás de «Modelo de la sesión:», un párrafo:

```text
## Sesión herramienta-las-dos

Eval: 22-lcsp-contrato-menor-en-2019.yaml

Modelo pedido: claude-sonnet-5-5

Modelo de la sesión: claude-sonnet-5-5

Skills activadas: consulta-juridica (antes), normativa (después)

Pregunta:
```

`Skills activadas: ninguna` con la lista vacía y `Skills activadas: sin leer` sin la lista. La tabla «Sesiones», con su
columna «Activada», no cambia.

## 4. Uso, de fuera adentro

Ninguna skill ni el binario la piden: la lee la persona en el informe de una medición —también en el del cierre que sale
en rojo—, sobre todo la de las sesiones que cuentan en `sin_activar`, una o ninguna por medición en la serie de partida.
Por sesión, 109 bytes en `informe.json` con una activación, 195 con dos, 26 con ninguna y 0 sin leer (research V2), y un
párrafo de 35 a 65 bytes en `informe.md` (research V16). Una medición de `normativa`, con sus 24 evals, tres
repeticiones, dos modelos y dos modos, son como mucho unas 288 sesiones: unos 31 KB más en `informe.json` y unos 11 KB en
`informe.md` con una activación cada una. No crece con el uso del producto —cientos de normas y miles de bloques
consultados no la cambian—, solo con las sesiones del plan del job. No es una señal que persista: cada medición escribe
su informe, y el de la cabeza siguiente da otra lista.

## 5. Controles

- **`TestInformeConLasActivaciones`** (`internal/evals/informe_test.go`; FR-001 a FR-004, FR-006; SC-001, SC-002):
  escribe con `EscribirInforme` el informe de las sesiones fijas de `internal/evals/testdata/sesiones/activaciones/`, con
  la skill `normativa`, el modelo que decide `claude-sonnet-5-5` y los dos modos. Su `evals/` tiene una eval con
  `activa: true` y `redaccion_fechada`, como `evals/normativa/22-lcsp-contrato-menor-en-2019.yaml`, con la que el informe
  publica sin juez `sin_activar:claude-sonnet-5-5:<modo>` (research V10). Sus `sesiones/`, cinco por modo, con la forma
  de las de `testdata/sesiones/informe/` (modo orden) y `testdata/sesiones/leer-llamadas/` (modo herramienta, con
  `servidor.json`) y un `message.id` por respuesta; el número entre paréntesis es el turno:

  | Sesión (`orden-…` y `herramienta-…`) | Lo que hace | `activaciones` | `activada` |
  |---|---|---|---|
  | `antes` | `Skill` `normativa` (1); la consulta (2); respuesta | `normativa` antes | sí |
  | `ninguna` | la consulta (1); respuesta | `[]` | no |
  | `otra` | `Skill` `consulta-juridica` (1); la consulta (2); respuesta | `consulta-juridica` antes | no |
  | `las-dos` | `Skill` `consulta-juridica` (1); la consulta (2); `Skill` `normativa` (3); respuesta | `consulta-juridica` antes, `normativa` después | sí |
  | `mismo-turno` | `Skill` `normativa` y la consulta en la misma respuesta (1); respuesta | `normativa` después | sí |

  La consulta es `mcp__kitlegal__boe_articulo` en modo herramienta y `Bash` `kitlegal boe articulo … --json` en modo
  orden. El test exige, sesión a sesión, la lista de `informe.json` y el párrafo de `informe.md`, y falla con la primera
  que difiere. Y exige los valores de hoy: `activada` de cada sesión, `sin_activar:claude-sonnet-5-5:orden` y `:herramienta`
  con 2 de 5, las tasas, los motivos y el veredicto; estas aserciones se escriben y pasan con el informe de hoy antes de
  tocar `informe.go` (plan, «Diseño»).
- **`TestInforme`**, caso `sesion-ilegible` (FR-005): su resultado no lleva `activaciones` y su sección dice
  `Skills activadas: sin leer`. Los demás casos, sin cambiar ningún valor esperado.
- **`TestInformeDeLaSuiteCompletaNoCambia`** (SC-002): quita de lo escrito, además de lo de H30, las claves
  `activaciones` de `informe.json` y los párrafos `Skills activadas:` de `informe.md`, y lo compara byte a byte con
  `informeJSONDeLaSuiteCompleta` e `informeMDDeLaSuiteCompleta`, que no cambian; sus casos de lo que se quita fijan que
  se quita eso y nada más (research D6).
- **`TestActivacionesDeLosUsos`** (`internal/evals/activaciones_test.go`, NUEVO; FR-001): con usos escritos en el test,
  en los dos modos y en el de una eval sin binario ni servidor: una activación antes, una en el mismo turno, una sin
  ninguna herramienta (antes), una skill activada dos veces (dos entradas), una orden `kitlegal` de `Bash` en una sesión
  sin binario ni servidor (antes) y ninguna activación (`[]`, no nil).
- **Sin cambiar ningún valor esperado**: los demás tests del informe, de `LeerSesion` y de la serie de cierres, y
  `TestJuzgarSinLaLista`, que lee informes versionados de H7.1 a H7.3 (FR-005; research V12).
