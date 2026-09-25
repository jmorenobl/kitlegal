# T022 — tarea mal delimitada: con `legal-core` en `skills/`, el guion de reinstalación exige una sola skill y T022 no puede tocar `testdata/` (intento 1)

**Estado**: sin marcar (`[ ]`), **redelimitada**. En el árbol no queda ningún fichero de T022: solo cambian `tasks.md` y
esta nota, los dos en el directorio del hito (además de `tarea-actual.json` y `tareas-intentos.json`, que escribe el
workflow). Si quedara algo, el guardián de la tarea siguiente —T028, que solo declara el guion de reinstalación— lo
rechazaría por estar fuera de sus rutas. T022 entera está implementada y verificada en un clon desechable (sección
«Verificación del arreglo»), y su `SKILL.md` verificado va al final de esta nota, así que el intento 2 puede partir de
ahí.

## Causa

`make ci` ejecuta `test-integration`, y con él `TestInstalacion` (`internal/skills/instalacion_test.go`, etiqueta
`integration`), que corre `make install` sobre una copia mínima del árbol con el `skills/` entero. Su guion
`internal/skills/testdata/script/instalar-de-nuevo.txtar` termina así:

```text
exec ls -A $HOME/.claude/skills
stdout '\Aboe-legislacion\n\z'
```

`scripts/instalar-skills.sh` enlaza en el directorio personal **cada** directorio de `skills/`. En cuanto
`skills/legal-core/` existe, la instalación la enlaza también y el guion queda en rojo. Con T022 entera en un clon y el
guion tal como está en la rama:

```text
--- FAIL: TestInstalacion/instalar-de-nuevo (1.75s)
    > exec ls -A $HOME/.claude/skills
    [stdout]
    boe-legislacion
    legal-core
    > stdout '\Aboe-legislacion\n\z'
    FAIL: testdata/script/instalar-de-nuevo.txtar:32: no match for `\Aboe-legislacion\n\z` found in stdout
```

Es el único control que cambia de resultado: los demás guiones de la instalación comprueban `boe-legislacion` sin
excluir otras skills, y el resto de `make ci` queda en verde (sección siguiente). Cumplir T022 exige, por tanto,
cambiar un guion existente bajo el `testdata/` de un paquete, y T022 no lleva la etiqueta de datos: el guardián rechaza
cualquier cambio bajo `testdata/`. Quitar la comprobación o relajarla a «contiene `boe-legislacion`» sería un atajo que
la batería prohíbe y vaciaría el control de FR-053 de H5 («no aparece nada más»). Ni el plan ni el juez de tareas lo
detectaron: es el mismo patrón que T015 y T017, un test que da por fijo algo que el hito hace crecer.

## Decisión

- **Tarea nueva T028, de datos, colocada en `tasks.md` justo antes de T022** (el bucle toma la primera línea `- [ ]`
  en el orden del fichero, como T027 antes de T007). Solo cambia la última comprobación del guion y su comentario de
  cabecera: la lista de skills sale del árbol copiado, recorrido como lo recorre `instalar-skills.sh`, se exige que
  traiga `boe-legislacion` —para que la comparación no pase en vacío— y se compara byte a byte con el listado del
  directorio personal. Es **divisible**: queda en verde con una skill, antes de T022, y con dos, después; por eso va
  aparte y no como excepción dentro de T022, que no puede mezclar material de test protegido con código. Al modificar
  material existente, **provoca pausa humana** (`clasificar_datos` lo marca `(existente)`).
- **No se escribe la lista nueva literal** (`boe-legislacion` y `legal-core`): quedaría en rojo antes de T022, así que
  no sería divisible, y volvería a romperse con cada skill que llegue del backlog (la primera vertical, `boe-fiscal`).
  Lo que el guion afirma es «una entrada por skill y ninguna más», no una lista.
- **La lista se toma recorriendo los directorios**, como la instalación (`"$WORK"/repo/skills/*/`), y no con
  `ls -A` sobre `skills/`: un fichero suelto en `skills/`, que la instalación no enlaza, daría un falso rojo.
- La línea de T022 **no cambia**: sus rutas son las mismas. En `tasks.md` cambian las cuentas (doce tareas de datos,
  diez pausas), las excepciones de rebanadas verticales, la dependencia **T028 → T022**, la trazabilidad (FR-060 y
  FR-085), la tabla de obligaciones, los ficheros existentes que se tocan, la entrega de US5 y el criterio e de la
  rúbrica.

## Verificación del arreglo

Dos clones desechables de la base de este intento (`bc46eac`), con `git clone` a `/tmp`, fuera del repositorio y sin
versionar:

- **`/tmp/kitlegal-t022`, T022 entera**:
  1. `internal/app/skills_test.go` con `legal-core` en `skillsExigidas`, primero: `TestSkillsDelRepositorio/skills` en
     rojo con `[]string{"boe-legislacion"} does not contain "legal-core"`.
  2. `skills/legal-core/SKILL.md` con la región de la tabla vacía y `make -C /tmp/kitlegal-t022 skills-sync`, que
     escribe la región con el verbo `resolver`, `references/leyes_vertebrales.md` (cabecera
     `<!-- generado desde data/normas.yaml, no editar -->`, las quince vertebrales), `references/jerarquia_normativa.md`
     (cabecera `<!-- generado desde data/jerarquia.yaml, no editar -->`, los cinco niveles y las cuatro reglas) y el
     enlace `scripts/territorio -> ../../../bin/instalado/kitlegal`. Ninguna de las dos referencias lleva texto de
     norma.
  3. `TestSkillsDelRepositorio`, `TestTablaDeComandosCoincideConLaGramatica` y `TestRegenerarYComparar` en verde; los
     casos negativos parametrizados de T021 se ejercen ya sobre `legal-core` (37 subtests: deriva de cada referencia,
     de sus datos y de `--describe`; 299 y 300 líneas; los trece de frontmatter y los dos de cada referencia sin sus
     datos; los cuatro de enlaces; los tres de la región; regenerar dos veces; normas nombradas; sin instrucciones de
     evals), todos en PASS.
  4. `SKILL.md` regenerado: 170 líneas. Un `grep -i` de `eval`, `job`, `modelo`, `cach`, `plazo`, `recurso`,
     `competenc`, `claude`, `offline`, `github` y de los municipios, la comunidad y el boletín de las evals solo da
     «prevalece» (subcadena, no palabra) y `--offline` dentro de la región generada, que la comprobación excluye.
  5. `TestInstalacion` con el guion de la rama: el rojo de la sección «Causa». Con el guion de T028: los cinco guiones
     en PASS.
  6. `make -C /tmp/kitlegal-t022 ci` → **`ci: todos los controles en verde`**, dos veces (la segunda tras los últimos
     retoques de redacción de `SKILL.md`). `git status` del clon: exactamente las cinco rutas de T022 y el guion.
- **`/tmp/kitlegal-t028`, solo T028**: `TestInstalacion` con sus cinco guiones en PASS, con `boe-legislacion` como única
  skill. Sonda negativa: un `mkdir $HOME/.claude/skills/sobrante` antes del listado deja `instalar-de-nuevo` en rojo en
  el `cmp`, con el diff `-sobrante`.

## Para T028: el guion verificado

Las dos piezas que cambian en `internal/skills/testdata/script/instalar-de-nuevo.txtar`. El comentario de cabecera,
líneas 3 y 4:

```text
# se duplica ni se rompe, y en el directorio personal de skills no aparece nada
# más que una entrada por cada skill del árbol (US3 escenario 2, FR-053,
# SC-006).
```

Y el final del guion, tras `! exists $WORK/repo/skills/boe-legislacion/boe-legislacion`, en lugar del
`exec ls -A` con su `stdout` literal:

```text
# Y en el directorio personal de skills hay una entrada por cada skill del
# árbol, con su nombre, y ninguna más. La lista sale del árbol, recorrido como lo
# recorre la instalación, y no del guion: así no cambia al añadirse una skill, y
# que traiga boe-legislacion impide que la comparación pase en vacío.
exec sh -c 'for skill in "$WORK"/repo/skills/*/; do basename "$skill"; done'
stdout '^boe-legislacion$'
cp stdout skills-del-arbol.txt

exec ls -A $HOME/.claude/skills
cmp stdout skills-del-arbol.txt
```

Dentro de las comillas simples no expande testscript sino `sh`, que recibe `WORK` en su entorno, como el
`exec sh -c 'cd "$WORK/repo/skills/boe-legislacion" && pwd -P'` de `instalar.txtar`. El `stdout` de testscript casa
en modo multilínea, así que `^boe-legislacion$` casa con una línea de la lista.

## Para el intento 2 de T022

En el orden rojo → verde:

1. `internal/app/skills_test.go`: `var skillsExigidas = []string{"boe-legislacion", "legal-core"}`, con el comentario
   «boe-legislacion la trajo H5 y legal-core, H6». `go test -count=1 -run '^TestSkillsDelRepositorio$/^skills$'
   ./internal/app/` queda en rojo.
2. `skills/legal-core/SKILL.md` con el texto de abajo, que lleva la región de la tabla **vacía**.
3. `make skills-sync`: escribe la región, las dos referencias y el enlace. Nada de esto se escribe a mano.
4. Los tres tests de la tarea y `make ci`, en primer plano.

El `SKILL.md` verificado, antes de `make skills-sync`:

````markdown
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
````

Lo que hace que el texto cumpla la línea de la tarea sin pasar por ningún control en vacío:

- **Forma fija de lo que se mide.** La plantilla de «Cómo se presenta el territorio» pone en la respuesta la comunidad
  y la provincia por su nombre, cada boletín por su `codigo` —exacto y como palabra— y cada aspecto de la cobertura en
  la forma `<aspecto>: <valor>`, que es lo que `ExtraerTerritorio` busca (contrato de evals §2).
- **Genericidad.** Ningún municipio, comunidad ni boletín concreto aparece en el texto: los ejemplos son la forma de
  los datos (`<código INE> <nombre> (<provincia>)`, `boletin_provincial: no-configurado`).
- **Lo que no dice.** Ni evals, ni el job, ni modelos, ni la caché, ni plazos, recursos o competencia (FR-068); ninguna
  norma por su número y año, de modo que `normas-nombradas` no depende de esta skill.
- **Seis reglas, las del contrato §1.4**, en su orden; la genericidad y la delegación van en el protocolo, no como
  reglas añadidas.
