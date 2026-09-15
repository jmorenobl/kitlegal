# Aceptación de H5 (FR-080, FR-081, SC-001, SC-002)

**Nota de la ronda 2 de la revisión final (2026-09-15):** esta es la aceptación de la ejecución 35002104338 sobre
`5c6c552`, que dejó de cubrir la cabeza desde `ede21ba`, entre otros cambios con el `Makefile` de `make install`
(`gates/evals-cierre.md`). La repetición de T031 sobre la cabeza que se empuje tras la revisión final la registra de
nuevo desde su propia ejecución (contrato del job §7). Lo que sigue es lo del intento 1, tal cual.

De la ejecución de cierre de T031 (intento 1, 2026-09-15; `gates/evals-cierre.md`): ejecución
[35002104338](https://github.com/jmorenobl/kitlegal/actions/runs/35002104338) del job `evals` sobre la propuesta de
cambio [#27](https://github.com/jmorenobl/kitlegal/pull/27), commit evaluado
`5c6c552d20419d9ab01769533e01ff181297e3b2`, modelo `claude-haiku-4-5-20251001`, Claude Code `2.1.270`, veredicto
`aprobado`. Las dos sesiones son sesiones de Claude Code en `-p` en el runner, con la skill y el binario instalados como
los deja `make install` (paso «Instalar kitlegal y las skills como las deja make install») y después de retirar Python
(paso «Retirar Python del runner» y comprobación 3 del guion, §3); las invocaciones son las que `strace` registró en
cada sesión y el informe atribuyó (contrato del job §4).

Lo que sigue sale de `informe.json` tal como lo imprimió la quinta orden de quickstart §12.3 (anexo A de
`gates/evals-cierre.md`): las líneas entre `--- inicio de informe.json ---` y `--- fin de informe.json ---`, sin el
prefijo de tarea, paso e instante que pone `gh run view --log`, leídas con `jq`. Las tablas, las invocaciones, las
respuestas y el `sin_python` se generaron del JSON, sin transcribirlos a mano; las preguntas, que `informe.json` no
lleva, son las de la sección de cada sesión en `informe.md`, en el mismo anexo.

## 1. Norma no fiscal: `01-lpac-articulo-21` (FR-080, SC-001)

Pregunta: «¿qué dice el art. 21 de la Ley 39/2015?»

| Clave de `informe.json` | Valor |
|---|---|
| `sesion` | `01-lpac-articulo-21` |
| `eval` | `01-lpac-articulo-21.yaml` |
| `activa` | `true` |
| `activada` | `true` |
| `codigo_de_la_sesion` | `0` |
| `fin_de_la_sesion` | `"result success"` |
| `sesion_terminada` | `true` |
| `comandos_ejecutados` | `["bloque boe BOE-A-2015-10565 a21"]` |
| `comandos_ausentes` | `[]` |
| `citas_encontradas` | `["BOE-A-2015-10565 a21"]` |
| `citas_ausentes` | `[]` |
| `fuera_de_lo_grabado` | `[]` |
| `otras_fallidas` | `[]` |
| `llegadas_a_la_red` | `[]` |
| `motivos` | `[]` |
| `pasa` | `true` |

Invocaciones del binario (`invocaciones`, en el orden de la traza):

| Orden | Código | Conexiones |
|---|---|---|
| `boe indice BOE-A-2015-10565 --json` | 0 | sin conexiones |
| `boe indice BOE-A-2015-10565 --json` | 0 | sin conexiones |
| `boe articulo BOE-A-2015-10565 a21 --json` | 0 | sin conexiones |

Respuesta (`respuesta`, tal cual):

`````text
El art. 21 de la Ley 39/2015 regula la **obligación de resolver** en los procedimientos administrativos. Sus aspectos principales son:

**1. Obligación de resolución expresa (apartado 1)**
La Administración está obligada a dictar resolución expresa y notificarla en todos los procedimientos. Se exceptúan los supuestos de terminación por pacto o convenio, y los procedimientos que se someten solo a declaración responsable o comunicación. [BOE-A-2015-10565, bloque a21]

**2. Plazos máximos (apartado 2)**
La resolución debe notificarse dentro del plazo establecido por la norma reguladora del procedimiento, que no podrá exceder de **seis meses**, salvo que una norma con rango de ley establezca otro plazo mayor o así lo prevea el Derecho de la Unión Europea. [BOE-A-2015-10565, bloque a21]

**3. Plazo supletorio (apartado 3)**
Si la norma no fija plazo máximo, este será de **tres meses**. El cómputo comienza:
- En procedimientos iniciados de oficio: desde la fecha del acuerdo de iniciación.
- En procedimientos iniciados a solicitud: desde la entrada de la solicitud en el registro electrónico de la Administración competente. [BOE-A-2015-10565, bloque a21]

**4. Información a interesados (apartado 4)**
La Administración debe publicar los plazos máximos de duración de sus procedimientos e informar a los interesados sobre ellos en los diez días siguientes a la recepción de la solicitud. [BOE-A-2015-10565, bloque a21]

**5. Recursos adicionales (apartado 5)**
Ante congestión de solicitudes, se pueden habilitar medios personales y materiales adicionales para cumplir los plazos. [BOE-A-2015-10565, bloque a21]

**6. Responsabilidad (apartado 6)**
El incumplimiento de la obligación de dictar resolución expresa en plazo genera responsabilidad disciplinaria. [BOE-A-2015-10565, bloque a21]
`````

Lo que muestra: la respuesta expone el contenido del artículo 21 apartado por apartado (obligación de resolver
expresamente, plazo máximo, plazo supletorio y su cómputo, información a los interesados, medios adicionales y
responsabilidad), y cada apartado lleva la cita `[BOE-A-2015-10565, bloque a21]`, que la extracción cuenta como
`BOE-A-2015-10565 a21`. Los comandos que leyeron la norma quedan registrados: `boe indice BOE-A-2015-10565 --json` (dos
veces) y `boe articulo BOE-A-2015-10565 a21 --json`, los tres con código 0 y sin ninguna conexión, porque se leyeron de
lo grabado.

## 2. Norma fiscal: `06-irpf-rendimientos-del-trabajo` (FR-080, SC-002)

Pregunta: «¿Qué rendimientos se consideran rendimientos íntegros del trabajo según el artículo 17 de la ley del IRPF?»

| Clave de `informe.json` | Valor |
|---|---|
| `sesion` | `06-irpf-rendimientos-del-trabajo` |
| `eval` | `06-irpf-rendimientos-del-trabajo.yaml` |
| `activa` | `true` |
| `activada` | `true` |
| `codigo_de_la_sesion` | `0` |
| `fin_de_la_sesion` | `"result success"` |
| `sesion_terminada` | `true` |
| `comandos_ejecutados` | `["bloque boe BOE-A-2006-20764 a17"]` |
| `comandos_ausentes` | `[]` |
| `citas_encontradas` | `["BOE-A-2006-20764 a17"]` |
| `citas_ausentes` | `[]` |
| `fuera_de_lo_grabado` | `[]` |
| `otras_fallidas` | `[]` |
| `llegadas_a_la_red` | `[]` |
| `motivos` | `[]` |
| `pasa` | `true` |

Invocaciones del binario (`invocaciones`, en el orden de la traza):

| Orden | Código | Conexiones |
|---|---|---|
| `boe indice BOE-A-2006-20764 --json` | 0 | sin conexiones |
| `boe articulo BOE-A-2006-20764 a17 --json` | 0 | sin conexiones |

Respuesta (`respuesta`, tal cual):

`````text
Según el art. 17 de la Ley 35/2006 [BOE-A-2006-20764, bloque a17], **se consideran rendimientos íntegros del trabajo** las siguientes categorías:

## 1. Definición general (apartado 1)

Todas las **contraprestaciones o utilidades, dinerarias o en especie**, que deriven directa o indirectamente del trabajo personal o de la relación laboral o estatutaria, siempre que no tengan carácter de rendimientos de actividades económicas. En particular:

- Sueldos y salarios
- Prestaciones por desempleo
- Remuneraciones por gastos de representación
- Dietas y asignaciones para gastos de viaje (con límites reglamentarios)
- Aportaciones a planes de pensiones satisfechas por promotores y empresas
- Contribuciones de empresarios para hacer frente a compromisos por pensiones

## 2. Otros rendimientos del trabajo (apartado 2)

El artículo también incluye como rendimientos del trabajo:

**a) Prestaciones sociales:**
- Pensiones y haberes pasivos de Seguridad Social, clases pasivas y prestaciones por incapacidad, jubilación, accidente, enfermedad, viudedad
- Prestaciones de planes de pensiones y fondos de pensiones
- Prestaciones de mutualidades de previsión social
- Prestaciones por seguros de dependencia

**b) Retribuciones por cargo público:** cantidades abonadas a diputados europeos, parlamentarios, concejales y similares (excluido gastos de viaje y desplazamiento)

**c) Otros rendimientos:** cursos y conferencias, obras literarias/artísticas/científicas cedidas, retribuciones de administradores, pensiones compensatorias, becas, retribuciones en actividades humanitarias, etc.

## 3. Criterio de calificación (apartado 3)

Cuando los rendimientos del trabajo (particularmente los de cursos/conferencias y obras cedidas) supongan la **ordenación por cuenta propia de medios de producción y recursos humanos** con finalidad de intervenir en la producción o distribución de bienes o servicios, se calificarán como **rendimientos de actividades económicas**, no de trabajo.
`````

Lo que muestra: la respuesta cita la norma, el artículo y su bloque desde la primera línea
(`Según el art. 17 de la Ley 35/2006 [BOE-A-2006-20764, bloque a17]`), que la extracción cuenta como
`BOE-A-2006-20764 a17`, y expone los tres apartados del artículo (definición general, otros rendimientos del trabajo y
criterio de calificación como actividad económica). Los comandos registrados son `boe indice BOE-A-2006-20764 --json` y
`boe articulo BOE-A-2006-20764 a17 --json`, los dos con código 0 y sin ninguna conexión.

## 3. Cómo se comprobó que no había Python (FR-081)

`sin_python` de `informe.json`, que es el contenido de `sin-python.txt`:

```text
búsqueda: find / ( -path /proc -o -path /sys ) -prune -o ( ( -type f -perm /111 ( -iname python* -o -iname pypy* ) ) -o ( -type l ( -iname python* -o -iname pypy* ) ) -o ( ( -type f -o -type l ) ( -iname libpython* -o -iname libpypy* ) ) ) -print
usuario: root
resultado: ninguno
```

Es la comprobación 3 de `scripts/evals.sh` (contrato del job §3.1; research D17 y V56), que corre antes de la primera
sesión y después del paso «Retirar Python del runner»: la misma búsqueda que ese paso, con el mismo arreglo de
argumentos, ejecutada como root (`usuario: root`) en todo el sistema de ficheros salvo `/proc` y `/sys`, de ficheros
ejecutables y enlaces llamados `python*` o `pypy*` y de bibliotecas `libpython*` o `libpypy*`, sin ningún resultado
(`resultado: ninguno`). Si hubiera encontrado alguna ruta, o no hubiera podido buscar como root, o `find` hubiera
fallado, el guion habría terminado con 1 antes de la primera sesión y la ejecución no tendría informe ni aceptación que
registrar. El texto coincide byte a byte con el que el contrato del job §3.1 fija para una búsqueda sin Python
(comparado con `cmp` contra ese bloque). Las dos sesiones de arriba corrieron, por tanto, sin ningún intérprete de
Python accesible en el runner.
