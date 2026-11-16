# Contrato: el arnés con la variante compuesta y el presupuesto del listado

FR-010 a FR-018; SC-002 a SC-005. Research D5 a D9 y D13; data-model §2 a §5. Lo que no se nombra aquí es el arnés de
H41.2 tal cual (`specs/036-h41-2-por-qu-e/contracts/arnes.md`).

## 1. Cómo se lanza

La misma cadena (`Makefile` → `scripts/evals-sondeo-llavero.sh laboratorio` → `scripts/evals-laboratorio.sh` →
`go test -tags evals … -run '^TestSondeoDelLaboratorio$' ./internal/evals/` → `sondearElLaboratorio`), con un argumento
más detrás de los nueve de hoy y un valor más de `INSTALACION`:

```sh
make evals-laboratorio VARIANTES=<v>[,<v>…] MODOS=<m>[,<m>] SESIONES=<n> MODELO=<id> CLAUDE_CODE=<versión> \
  CONCURRENCIA=<c> DIRECTORIO=<ruta> [INSTALACION=producto|producto-descripciones-cortas] [EVAL=<nn>] \
  [PRESUPUESTO=<caracteres>]
```

La receta pasa `"$(PRESUPUESTO)"` detrás de `"$(EVAL)"`; `scripts/evals-laboratorio.sh` exige diez argumentos y lo da a
`go test` como `-presupuesto`, delante de `-temporal`; su uso, con nueve, nombra `<presupuesto>` detrás de `<eval>`.
`scripts/evals-sondeo-llavero.sh` pasa sus argumentos con `"$@"` y solo cambia su comentario. Lo lanza una persona con la
credencial del sondeo; nunca `make ci`, un gancho, el job ni un paso del workflow (FR-017).

El sondeo de H41.3 (FR-030) son cuatro invocaciones y una tabla, con las sesiones fuera del repositorio; el `DIRECTORIO`
de cada una no puede existir antes, y `claude` en el `PATH` tiene que ser la 2.1.284, que el arnés comprueba antes de
abrir nada:

```sh
make evals-laboratorio VARIANTES=normativa-producto MODOS=herramienta SESIONES=160 MODELO=claude-sonnet-5-5 \
  CLAUDE_CODE=2.1.284 CONCURRENCIA=5 DIRECTORIO=../kitlegal-sondeos/h41-3/sola EVAL=04
make evals-laboratorio VARIANTES=normativa-producto MODOS=herramienta SESIONES=160 MODELO=claude-sonnet-5-5 \
  CLAUDE_CODE=2.1.284 CONCURRENCIA=5 DIRECTORIO=../kitlegal-sondeos/h41-3/control EVAL=04 INSTALACION=producto
make evals-laboratorio VARIANTES=normativa-producto MODOS=herramienta SESIONES=160 MODELO=claude-sonnet-5-5 \
  CLAUDE_CODE=2.1.284 CONCURRENCIA=5 DIRECTORIO=../kitlegal-sondeos/h41-3/diagnostico EVAL=04 INSTALACION=producto \
  PRESUPUESTO=1000000
make evals-laboratorio VARIANTES=normativa-producto MODOS=herramienta SESIONES=160 MODELO=claude-sonnet-5-5 \
  CLAUDE_CODE=2.1.284 CONCURRENCIA=5 DIRECTORIO=../kitlegal-sondeos/h41-3/remedio EVAL=04 \
  INSTALACION=producto-descripciones-cortas
make laboratorio-tabla DIRECTORIO=../kitlegal-sondeos/h41-3
```

4 × 160 = 640 sesiones, `04-lgt-prescripcion-01` a `-160` en cada fila.

## 2. Argumentos

Los errores van juntos, uno por línea, cada uno con el nombre del argumento de `make`; con cualquiera no se abre ninguna
sesión, no se construye nada y no se crea `DIRECTORIO` (FR-016). Los de hoy no cambian ni de texto ni de orden; el de
`PRESUPUESTO` va el último, detrás del de `EVAL`; los de la variante, detrás de los demás de `INSTALACION`. Vacíos,
`INSTALACION` y `PRESUPUESTO` son «sin la opción» y no dan error.

| Argumento | Vale si | Mensaje si no |
|---|---|---|
| `INSTALACION` | vacío, `producto` o `producto-descripciones-cortas` | `INSTALACION: «<v>» no es producto`, el de hoy |
| `INSTALACION=producto` o `=producto-descripciones-cortas` | las variantes pedidas no son de la familia de H39 | `INSTALACION: las variantes del laboratorio de H39 no son de ninguna familia del producto`, el de hoy |
| `INSTALACION=producto-descripciones-cortas` | las variantes pedidas que valen son solo `normativa-producto` | `INSTALACION: producto-descripciones-cortas solo vale con VARIANTES=normativa-producto` |
| `INSTALACION=producto-descripciones-cortas` | la variante no tiene ningún defecto de contracts/variante.md §3 | `INSTALACION: <defecto>`, uno por defecto |
| `PRESUPUESTO` | vacío, o un entero mayor que 0 escrito solo con cifras, sin signo ni ceros delante, que cabe en un `int` | `PRESUPUESTO: «<v>» no es un entero mayor que 0` |
| `PRESUPUESTO` con valor que vale | las variantes pedidas no son de la familia de H39 | `PRESUPUESTO: las variantes del laboratorio de H39 no son de ninguna familia del producto` |

El de H39 de `INSTALACION` y el de `PRESUPUESTO` solo se dan cuando las variantes que valen son de una sola familia, la
de H39, como hoy el de `INSTALACION=producto`; sin familia, el de `VARIANTES` o el de la mezcla ya lo dicen. El de «solo
vale con» no depende de la familia: sale con cualquier variante que vale y no es `normativa-producto`, también en una
mezcla, salvo con la familia de H39, que da solo el suyo; de los tres primeros de `INSTALACION` sale uno como mucho, y los
de la variante se suman. `PRESUPUESTO` da una sola línea: la de la forma si no la tiene, y si no, la de H39. La variante solo se lee con
`INSTALACION=producto-descripciones-cortas`. Dos ejemplos en la salida de error. El de `VARIANTES=normativa-control
MODOS=herramienta SESIONES=4 MODELO=claude-sonnet-5-5 CLAUDE_CODE=2.1.284 CONCURRENCIA=1 DIRECTORIO=<uno que no existe>
INSTALACION=producto-descripciones-cortas PRESUPUESTO=0100`:

```text
INSTALACION: producto-descripciones-cortas solo vale con VARIANTES=normativa-producto
PRESUPUESTO: «0100» no es un entero mayor que 0
```

Y el de `VARIANTES=control MODOS=orden SESIONES=4 MODELO=claude-sonnet-5-5 CLAUDE_CODE=2.1.284 CONCURRENCIA=1
DIRECTORIO=<uno que no existe> INSTALACION=producto-descripciones-cortas PRESUPUESTO=1000000`, con la de H39:

```text
INSTALACION: las variantes del laboratorio de H39 no son de ninguna familia del producto
PRESUPUESTO: las variantes del laboratorio de H39 no son de ninguna familia del producto
```

Una petición sin las dos opciones que no vale da los mensajes de hoy, byte a byte (FR-015).

## 3. La instalación compuesta

Con `INSTALACION=producto-descripciones-cortas`, `preparar` hace lo de `INSTALACION=producto` —en
`skills/normativa-producto/` del temporal, la copia de `skills/normativa` del árbol y, a su lado, la de cada otra skill de
`skills/`, con `os.CopyFS`— y después, en cada `SKILL.md` de esa copia, sustituye las líneas de la clave `description`
del frontmatter por la `description` de la variante para esa skill (research D5). Cada sesión enlaza las cuatro, como
hoy. Lo que no es la `description` queda byte a byte como en `skills/` del árbol que lanza el sondeo, fichero a fichero;
la `description` leída con `skills.LeerFrontmatter` es el texto de la variante, exacto; y cada `SKILL.md` compuesto pasa
`skills.ValidarFrontmatter` como el de `skills/` (FR-011). Ejemplo, el principio del `SKILL.md` compuesto de
`jurisprudencia`, con la `description` en una línea, entrecomillada porque el texto lleva `: `:

```text
---
name: jurisprudencia
description: 'Sentencias y jurisprudencia de los tribunales españoles, sin citar ninguna de memoria. Se usa siempre que la pregunta nombre, pida o cite una sentencia (STS, ECLI, ROJ), pregunte si existe o qué dice, pida jurisprudencia sobre una materia o traiga una resolución, aunque el agente crea conocerla. kitlegal no consulta el CENDOJ: prepara la búsqueda y coteja el documento. Se activa antes de cita_preparar o cita_cotejar.'
metadata:
  kitlegal-applets: cita
---
```

Qué comillas elige el emisor no forma parte del contrato: lo que se fija es lo que se lee. La copia no se guarda en
`DIRECTORIO`, como hoy la de `skills/`.

## 4. El entorno de cada sesión

Con `PRESUPUESTO=N`, el entorno de cada sesión es el de hoy con `SLASH_COMMAND_TOOL_CHAR_BUDGET=N` y nada más distinto: el
valor tal como se pidió, detrás de las variables de la lista cerrada, `PATH` y `HOME`, y debajo de lo de cada sesión. Sin
`PRESUPUESTO`, ninguna sesión la tiene, aunque esté en el entorno de quien lanza el arnés: la lista cerrada de hoy no la
deja pasar (research D8). Vale con cualquier `INSTALACION` válida de una familia del producto, también vacía.

## 5. `sondeo.json`

Las claves de hoy y, detrás de `skills_instaladas`, `instalacion` con `INSTALACION=producto-descripciones-cortas` y
`presupuesto` con `PRESUPUESTO`. Sin las dos opciones, el de hoy byte a byte (FR-012, FR-015). Los de las cuatro filas de
FR-030, con `R` por `["buscar_la_norma.md","como_se_cita.md","con_herramientas.md","lecturas_que_fallan.md","normas.md","redaccion_en_una_fecha.md","redaccion_modificada.md","varios_bloques.md"]`
y `P` por `["consulta-juridica","jurisprudencia","normativa","plazos-y-territorio"]`:

```text
sola        {"modelo":"claude-sonnet-5-5","claude_code":"2.1.284","referencias_del_producto":R,"skills_instaladas":["normativa"]}
control     {"modelo":"claude-sonnet-5-5","claude_code":"2.1.284","referencias_del_producto":R,"skills_instaladas":P}
diagnóstico {"modelo":"claude-sonnet-5-5","claude_code":"2.1.284","referencias_del_producto":R,"skills_instaladas":P,"presupuesto":1000000}
remedio     {"modelo":"claude-sonnet-5-5","claude_code":"2.1.284","referencias_del_producto":R,"skills_instaladas":P,"instalacion":"producto-descripciones-cortas"}
```

La constancia nueva son 22 bytes (`,"presupuesto":1000000`) y 46 (`,"instalacion":"producto-descripciones-cortas"`), una vez
por sondeo; la lee el guion de la tabla una vez por sondeo (contracts/tabla.md). La salida del arnés no cambia con las
opciones.

## 6. Controles (FR-018)

Sin red ni modelo, con la construcción sustituida y el sustituto de `claude` de `TestArnesDelLaboratorio`, que además
deja por cada sesión una anotación `entorno.XXXXXX` con la ruta de la sesión y la salida de `env` (research D13). Cada
uno falla con la primera sesión cuyas skills, descripción, bytes o entorno difieren de lo esperado, o con una petición
inválida que abre alguna sesión.

**`TestArnesConLasDescripcionesCortas`** (`internal/evals/laboratorio_sondeo_test.go`):

- errores de uso, con su mensaje, ninguna sesión, nada construido, el temporal vacío y `DIRECTORIO` sin crear: con
  `normativa-control`; con `jurisprudencia-producto`; con `control`, de H39; con `normativa-producto,normativa-control`;
  y, sobre una raíz de prueba con `laboratorio/variantes.yaml` del repositorio, `skills/` del repositorio y la variante
  sin `plazos-y-territorio.txt`, el de la variante:
  `INSTALACION: laboratorio/producto-descripciones-cortas: falta la descripción de plazos-y-territorio`;
- la petición del remedio, `VARIANTES=normativa-producto MODOS=herramienta SESIONES=2 EVAL=04
  INSTALACION=producto-descripciones-cortas`: dos sesiones; cada una enlaza exactamente las cuatro skills de `skills/`, en
  orden de nombre y sin repetir, a la copia de su variante en el temporal; la `description` de cada `SKILL.md` enlazado,
  leída, es el texto de su fichero de la variante; lo demás de ese `SKILL.md` es el de `skills/`, con la clasificación de
  las líneas de `descripcionYCuerpo`; cada otro fichero, ruta a ruta y byte a byte, el de `skills/`;
  `skills.ValidarFrontmatter` no da ningún defecto; su servidor, el de la familia; `sondeo.json`, el de remedio de §5; y
  ningún `SKILL.md` en `DIRECTORIO`.

**`TestArnesConElPresupuesto`** (mismo fichero), con `SLASH_COMMAND_TOOL_CHAR_BUDGET=500` en el entorno de quien lanza en
todos los casos:

- errores de uso, igual que arriba: `PRESUPUESTO=0`, `0100`, `+1000000` y `99999999999999999999`, cada uno con
  `PRESUPUESTO: «<v>» no es un entero mayor que 0`; `PRESUPUESTO=1000000` con `control`, de H39; y los dos ejemplos de §2
  con sus líneas en su orden;
- con `PRESUPUESTO=1000000`: la petición del diagnóstico (`INSTALACION=producto`), la del remedio con presupuesto y
  `jurisprudencia-control` sin `INSTALACION`; cada sesión tiene una sola línea `SLASH_COMMAND_TOOL_CHAR_BUDGET=1000000`;
  `sondeo.json` lleva `"presupuesto":1000000` detrás de lo de §5;
- sin `PRESUPUESTO`, la petición del control: ninguna sesión tiene la variable; el entorno que vuelca cada sesión es,
  nombre a nombre y valor a valor, el que componen para su modo las funciones de hoy, que el hito no cambia —en modo
  herramienta, el de `disponerElModo`: `sinElDirectorioEnElPATH(entornoDeLaSesion(entornoDelSondeo(<entorno del test>,
  <temporal>), <sesión>, false), <temporal>/bin)`, con `<sesión>` el directorio que le da el arnés,
  `DIRECTORIO/normativa-producto/herramienta/<nombre>` con ruta absoluta—, sin ninguna de más ni de menos salvo las que
  añaden al arrancar `bash`, el guion de la sesión, y `sh`, el sustituto: `PWD`, `SHLVL` y `_` (research V17); como
  conjunto, sin exigir el orden, que los intérpretes no tienen por qué conservar; `sondeo.json` es el de control de §5; y los nombres de las
  variables de cada sesión son los de la misma sesión del diagnóstico menos `SLASH_COMMAND_TOOL_CHAR_BUDGET`. Este caso
  se escribe y pasa con el arnés de hoy antes de añadir la opción (plan.md, «Diseño»): lo que fija «hoy» lo dan esas
  funciones, no el código nuevo, y el caso sale en rojo si el cambio añade, quita o cambia una variable sin
  `PRESUPUESTO`.

**Sin las dos opciones** (FR-015, SC-004): `TestArnesDelLaboratorio`, `TestArnesDeNormativa`, `TestArnesDeLasFamilias`,
`TestArnesConLaInstalacionDelProducto` y `TestArnesConUnaEval`, con sus valores esperados de hoy sin cambiar, para las
skills, las sesiones, las invocaciones, `sondeo.json`, la salida y los mensajes; el entorno, que ninguno de ellos lee, lo
fija el caso sin `PRESUPUESTO` de `TestArnesConElPresupuesto`.

**Las recetas y el guion**: `TestRecetasDelSondeo` espera en `make -n evals-laboratorio … PRESUPUESTO=1000000` la orden
del llavero con `"1000000"` detrás de `"04"`; `TestGuionDelLaboratorio` pasa diez argumentos a
`scripts/evals-laboratorio.sh`, exige `-presupuesto <décimo>` delante de `-temporal` en la orden de `go test`, y, con nueve,
el uso con `<presupuesto>` y ningún `go`.
