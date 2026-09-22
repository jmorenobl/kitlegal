---
name: legal-core
description: >-
  Punto de partida de las preguntas de derecho público español que dependen de un municipio o de un ayuntamiento:
  identifica su territorio —municipio, código INE, provincia, comunidad autónoma, régimen común o foral, DIR3 del
  ayuntamiento y boletines oficiales que le corresponden— con el binario kitlegal, y razona con la jerarquía normativa
  (qué nivel regula qué, de la Unión Europea al municipio, y dónde publica cada nivel) y con los identificadores BOE de
  las leyes vertebrales (Constitución, LPAC, LRJSP, LRBRL, TRLRHL, LCSP, LGT…). Úsala cuando se pregunte qué
  comunidad, provincia o boletines corresponden a un ayuntamiento, dónde se publican las normas de un municipio, si un
  municipio es de régimen foral, qué tipo de norma prevalece sobre otra o qué ley vertebral rige una materia. Para el
  texto de un artículo se apoya en boe-legislacion.
metadata:
  kitlegal-applets: territorio
  kitlegal-referencias: leyes_vertebrales jerarquia_normativa
---

# Territorio, jerarquía normativa y leyes vertebrales

Esta skill es el punto de partida de las preguntas de derecho público español que dependen de dónde se plantean. Antes
de razonar sobre ninguna norma identifica el territorio —municipio, provincia, comunidad autónoma, régimen, DIR3 del
ayuntamiento y boletines oficiales— con `scripts/territorio`, y después razona con dos referencias:
`references/leyes_vertebrales.md`, con el identificador `BOE-A-…` de cada ley vertebral, y
`references/jerarquia_normativa.md`, con qué nivel regula qué, dónde publica cada nivel y las reglas de
interpretación. No da el texto de ningún artículo: eso lo hace `boe-legislacion`. La skill identifica y orienta: no
tramita nada y no sustituye el asesoramiento de un profesional.

## Protocolo

Sigue los pasos en orden: del 1 al 4 siempre, y el 5 y el 6 cuando la pregunta va más allá del territorio. Escribe
todas las órdenes con `--json`: el sobre trae entonces `fuente`, `url`, `fecha_consulta` y `hash` junto a `data`.

### 1. Identificar el territorio

Antes de razonar sobre ninguna norma, averigua de qué municipio se habla.

- Toma el municipio de la conversación: su nombre o su código INE, tal como los haya dado la persona.
- Si la conversación no lo dice, **pregúntalo** y espera la respuesta. No lo supongas, no lo deduzcas de otros datos
  ni sigas con un municipio de ejemplo. Si solo se nombra una provincia o una comunidad, pregunta también por el
  municipio: `scripts/territorio` resuelve municipios.

### 2. Resolverlo con `scripts/territorio`

```bash
scripts/territorio resolver <nombre o código INE> --json
```

- Pasa el nombre o el código tal como los dio la persona; no conviertas de memoria un nombre en un código ni al revés.
- Ningún dato de territorio —comunidad, provincia, régimen, DIR3, boletines— se da por sabido ni se escribe de
  memoria, aunque parezca evidente: todos salen de `data` en esta conversación. Ningún municipio ni ninguna comunidad
  tiene un trato propio: lo que cambia de un territorio a otro lo dice `data`.
- Con el código 0, `data` trae siempre ocho claves —`municipio`, `codigo_ine`, `provincia`, `comunidad`, `dir3`,
  `regimen`, `boletines` y `cobertura`—, y cada dato lleva su `source`.
- Si `scripts/territorio` no resuelve a un binario, di que falta instalar kitlegal y no suplas los datos.

### 3. Leer `cobertura` y trasladarla a la respuesta

`cobertura` declara siempre sus tres aspectos, con un vocabulario cerrado:

| Aspecto | Valores | Qué dice |
|---|---|---|
| `boletin_autonomico` | `configurado`, `no-configurado` | si kitlegal tiene declarado el boletín de la comunidad |
| `boletin_provincial` | `configurado`, `no-configurado` | si kitlegal tiene declarado el de la provincia |
| `dir3` | `verificado`, `no-verificado` | si el DIR3 del ayuntamiento está en la correspondencia verificada |

- Responde con lo que dice `data`: el municipio y su código INE, la provincia, la comunidad, el régimen, el DIR3 y cada
  boletín de `boletines` con su nivel, su código y su nombre, y con su `motivo` si lo trae.
- Escribe la cobertura con sus tres aspectos, cada uno en su forma fija: la clave y el valor tal como los da `data`,
  separados por dos puntos y en la misma línea (más en «Cómo se presenta el territorio»).
- Lo que conste `no-configurado` o `no-verificado`, dilo explícitamente con sus palabras: kitlegal no tiene declarado
  ese boletín para ese territorio, o no ha verificado ese DIR3. Eso **no** significa que no exista (regla 5).
- **No nombres ningún boletín que el applet no haya devuelto**: ni su nombre, ni su sigla, ni su dirección, aunque
  creas conocerlo. La clase de boletín de cada nivel que da `references/jerarquia_normativa.md` dice dónde publica
  cada nivel en general, no cuál es el boletín de esa comunidad o de esa provincia: no la uses para suplir un nivel
  `no-configurado`.
- Con `dir3` en `no-verificado`, `dir3.codigo` va vacío: di que el DIR3 no está verificado y no lo compongas a partir
  del código INE.
- Si `regimen.valor` es `foral`, dilo: el municipio es de una comunidad de régimen foral, y lo preguntado puede regirse
  por normas forales propias (regla 4).
- Di de cuándo son los datos: `fecha_consulta` es la fecha de los ficheros de los que sale la respuesta.

### 4. Ambigüedad y ausencia

- **Código 2 con candidatos**: el nombre es el de más de un municipio, y `data.mensaje` los da todos, cada uno como
  `<código INE> <nombre> (<provincia>)`. Ofrécelos todos, pregunta a cuál se refiere la persona y vuelve al paso 2 con
  el código INE del que elija. No elijas por ella.
- **Código 2 sin candidatos**: la consulta no forma un código INE válido o su dígito de control no es el oficial, y el
  mensaje dice qué falla. Díselo a la persona y pide el nombre o el código; no corrijas tú el código.
- **Código 3**: ese nombre o ese código no están en la relación de municipios que lleva kitlegal. Dilo así —«no está
  en la relación»—, sin concluir que el municipio no exista: puede estar escrito de otra forma o haber cambiado de
  nombre. Pide otra forma del nombre o su código INE.
- Con cualquier otro código, di que no se pudo resolver el territorio y no suplas sus datos.

### 5. Razonar con las referencias

- `references/leyes_vertebrales.md` da cada ley vertebral con su abreviatura, su identificador `BOE-A-…`, su rango y
  sus materias. Nombra cada norma con el identificador de esa tabla.
- `references/jerarquia_normativa.md` da los niveles —Unión Europea, Estado, comunidad autónoma, provincia y
  municipio—, la clase de boletín que publica las normas de cada uno, sus tipos de norma de mayor a menor rango y las
  reglas de interpretación. Úsala para decir en qué nivel se regula lo preguntado, qué tipo de norma es cada una y cuál
  prevalece, aplicando las reglas tal como las enuncia.
- Une las referencias con el territorio resuelto: el nivel autonómico es el de la `comunidad` de `data` y el provincial,
  el de su `provincia`; los boletines concretos son solo los de `boletines`.

### 6. Delegar el texto en `boe-legislacion`

- Citar una norma por su identificador basta con `references/leyes_vertebrales.md`.
- **Afirmar lo que dice un artículo exige consultarlo con la skill `boe-legislacion` en esta misma conversación**, y
  citarlo como ella cita. No lo cites de memoria ni desde las referencias, que no llevan el texto de ninguna norma. Si
  no se puede consultar, di que no lo has consultado y no suplas el texto.
- El identificador de una norma que no está en `references/leyes_vertebrales.md` también se resuelve con
  `boe-legislacion`; no lo escribas de memoria.
- La delegación va en un solo sentido, de esta skill a `boe-legislacion`: esta skill no lleva las órdenes de `boe` ni
  su enlace.

## Cómo se presenta el territorio

Cada dato sale de `data`, con el nombre y el código tal como los da:

```text
Municipio: <nombre>, código INE <codigo> (dígito de control <digito_de_control>)
Provincia: <nombre de la provincia>
Comunidad o ciudad autónoma: <nombre de la comunidad>, régimen <común o foral, según regimen.valor>
DIR3 del ayuntamiento: <código, o «no verificado»>
Boletines: <código> — <nombre> (<nivel>), uno por cada entrada de `boletines`
boletin_autonomico: <valor>
boletin_provincial: <valor>
dir3: <valor>
```

- La cobertura va con la clave y el valor exactos de `data`, por ejemplo `boletin_provincial: no-configurado`: decirlo
  solo con otras palabras no la traslada. Detrás puedes explicar qué significa.
- Un boletín se nombra por su `codigo` y su `nombre` exactos. Si el mismo boletín cubre dos niveles, di los dos y su
  `motivo`.

## Comandos

Invoca el binario por el enlace `scripts/territorio` de esta skill. Responde con la relación de municipios y la
configuración territorial que lleva dentro el binario, sin consultar ninguna fuente. Códigos de salida: 0 correcto, 2
argumentos inválidos —también un nombre que es el de más de un municipio—, 3 no está en la relación.

<!-- inicio de la tabla de comandos: generada desde --describe con make skills-sync, no editar -->

### `scripts/territorio`

| Orden | Qué hace | Qué devuelve en `data` |
|---|---|---|
| `scripts/territorio resolver <consulta>` | Devuelve el territorio de un municipio, por su nombre o por su código INE, con la cobertura de lo que está configurado y verificado. | objeto con `municipio`, `codigo_ine`, `provincia`, `comunidad`, `dir3`, `regimen`, `boletines`, `cobertura` |

Todas devuelven el sobre `ok`, `fuente`, `url`, `fecha_consulta`, `hash`, `data`; con `ok` falso, `data` lleva `clase` y `mensaje`.

Banderas comunes: `--json`, `--timeout <valor>`, `--offline`, `--dry-run`, `--describe`, `--no-graph`, `--asunto <valor>`, `--verbose`.

<!-- fin de la tabla de comandos -->

## Reglas

1. **Nunca inventar contenido legal ni citar de memoria.** Ni el texto de una norma, ni su identificador, ni ningún
   dato de territorio: lo que no salga en esta conversación de `scripts/territorio`, de las referencias o de
   `boe-legislacion`, no se afirma.
2. **Cada afirmación sobre una norma, con su identificador.** El identificador `BOE-A-…` sale de
   `references/leyes_vertebrales.md` o de `boe-legislacion`; el texto de un artículo, solo de `boe-legislacion`.
3. **Distinguir ley de reglamento.** Di el rango de cada norma que nombres —el de la referencia— y recuerda que un
   reglamento nunca puede contradecir la ley.
4. **Señalar la variación autonómica.** Di lo que una comunidad autónoma puede haber regulado de otro modo, y más aún
   si su régimen es foral; las ordenanzas y demás normas locales no están en las referencias.
5. **No concluir «no existe»** a partir de un resultado sin cobertura completa: ni de un municipio que no está en la
   relación, ni de un boletín `no-configurado`, ni de un DIR3 `no-verificado`, ni de una norma que falta en una
   referencia.
6. **Ninguna acción con identidad.** No presentes, notifiques, firmes ni tramites nada en nombre de nadie, ni lo
   simules. Si la pregunta lo pide, di que es una acción que hace la persona.
