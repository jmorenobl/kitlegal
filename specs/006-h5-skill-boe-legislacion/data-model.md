# Modelo de datos: H5 · Skill `boe-legislacion` + andamiaje de skills

Entidades del spec (*Key Entities*) bajadas a campos, reglas de validación y ciclo de vida. Los contratos de
[contracts/](./contracts/) fijan formatos y órdenes; este documento fija **qué es cada cosa** y qué la hace válida.
Las decisiones y sus alternativas están en [research.md](./research.md).

Gramáticas de identificador que se reutilizan tal cual de H4 (`internal/source/boe/ids.go`, `ValidarNorma` y
`ValidarBloque`), sin copiarlas a mano en ningún otro sitio que no las compruebe contra aquellas:

| Nombre | Expresión | Origen |
|---|---|---|
| `NORMA` | `^BOE-A-[0-9]{4}-[0-9]{1,9}$` | `ids.go`, `prefijoDeNorma`, `digitosDelAnio`, `maximoDeDigitosDelNumero` |
| `BLOQUE` | `^[A-Za-z0-9][A-Za-z0-9.-]{0,63}$` | `ids.go`, `maximoDeCaracteresDelBloque`, `esBloqueValido` |
| `NOMBRE_DE_SKILL` | `^[a-z0-9]+(-[a-z0-9]+)*$`, 1-64 caracteres | validador del estándar Agent Skills (research.md D3) |

Un test (`TestGramaticasCoincidenConBoe`, `internal/evals`) comprueba que los patrones de `NORMA` y `BLOQUE` escritos en
`schemas/normas.yaml.json` y `schemas/eval.yaml.json` aceptan y rechazan exactamente lo mismo que `boe.ValidarNorma` y
`boe.ValidarBloque` sobre una tabla de casos límite (vacío, 3 y 5 dígitos de año, 10 dígitos de número, minúsculas,
`a85bis.`, `.a1`, 64 y 65 caracteres).

**Lectura de documentos YAML**, igual para `data/*.yaml`, las evals y el frontmatter de `SKILL.md`
(`internal/skills/esquemas.go`; research.md D8):

1. se analiza con `go.yaml.in/yaml/v3` a un `yaml.Node`, que conserva la línea de cada nodo;
2. se recorre el nodo y **toda clave repetida en cualquier mapa es un defecto** que nombra la ruta del mapa, la clave y
   sus dos líneas (p. ej. `normas: BOE-A-2015-10565 repetido en las líneas 12 y 20`). Analizar a `yaml.Node` no basta:
   esa llamada no comprueba repetidos, y al convertir el nodo la última clave ganaría en silencio (research.md V43);
3. se convierte a `any` con `(*yaml.Node).Decode`, que también rechaza la clave repetida (V43), y se normaliza a tipos
   JSON codificándolo en JSON y leyéndolo con `jsonschema.UnmarshalJSON` (números como `json.Number`, V44); un mapa con
   claves que no son texto no se puede codificar y es un defecto que nombra su ruta;
4. si el documento tiene esquema, se valida contra él, y cada error del esquema se presenta con la ruta dentro del
   documento y la línea del nodo que la ocupa.

Ningún error de estos pasos se descarta: el primero que aparece es el defecto del documento.

---

## 1. Skill

Directorio `skills/<nombre>/`. Todo directorio directamente bajo `skills/` es una skill; no hay otra forma de
declararla.

| Campo | Fuente | Regla |
|---|---|---|
| `nombre` | nombre del directorio | `NOMBRE_DE_SKILL` |
| `SKILL.md` | fichero | existe; frontmatter (§1.1); **menos de 300 líneas** (§1.2); exactamente una región de tabla de comandos si declara applets (§2) |
| `references/` | directorio | solo ficheros generados (§4.2); ni uno de más ni uno de menos |
| `scripts/` | directorio | solo enlaces simbólicos generados (§3); ni uno de más ni uno de menos |

### 1.1 Frontmatter

Bloque YAML entre la primera línea `---` y la siguiente línea `---`. Se lee con el lector común («Lectura de documentos
YAML», arriba): una clave repetida, también dentro de `metadata`, es un defecto que nombra la clave y sus dos líneas.

| Clave | Obligatoria | Regla (research.md D3) |
|---|---|---|
| `name` | sí | cadena; `NOMBRE_DE_SKILL`; igual al nombre del directorio |
| `description` | sí | cadena; no vacía tras recortar espacios; ≤ 1024 caracteres (runas); sin `<` ni `>` |
| `license` | no | cadena |
| `allowed-tools` | no | cadena o lista de cadenas |
| `compatibility` | no | cadena de ≤ 500 caracteres |
| `metadata` | no | mapa de cadena a cadena; aquí vive la declaración de kitlegal (§1.3) |

Cualquier otra clave de primer nivel es un defecto que nombra la clave.

### 1.2 Líneas

Número de líneas de `SKILL.md` regenerado, con la tabla de comandos que genera la sincronía y no la del árbol: número
de saltos de línea, más uno si el fichero no está vacío y no termina en salto.
Válido si es **≤ 299** (FR-041: «300 líneas o más» falla).

### 1.3 Declaración de kitlegal (en `metadata`; tipo Go `DeclaracionDeKitlegal`)

El tipo no se llama `Declaracion` a secas: suelta y sin tilde, `misspell` la lee como «declaration» (research.md V42);
en camelCase no la marca.

Lo que la sincronización genera para la skill sale de aquí y de `data/` (FR-035), nunca de una lista escrita para una
skill concreta.

| Clave de `metadata` | Valor | Regla |
|---|---|---|
| `kitlegal-applets` | nombres de applet separados por un espacio, en el orden en que la tabla los presenta | cada uno registrado en `app.RegistroDeProduccion()`; sin repetidos; si falta, la skill no tiene tabla ni `scripts/` |
| `kitlegal-referencias` | nombres separados por un espacio | cada nombre `n` tiene `data/n.yaml` y un generador conocido; en H5 el único es `normas`; si falta, `references/` debe estar vacío o no existir |

`boe-legislacion` declara `kitlegal-applets: boe` y `kitlegal-referencias: normas`.

---

## 2. Tabla de comandos (región generada de `SKILL.md`)

Delimitada por dos líneas exactas, en este orden y una sola vez cada una:

```text
<!-- inicio de la tabla de comandos: generada desde --describe con make skills-sync, no editar -->
<!-- fin de la tabla de comandos -->
```

Todo lo que hay entre ellas se sustituye al regenerar; el resto de `SKILL.md` queda byte a byte igual (FR-032).

### 2.1 Descripción de verbo (derivada de `--describe`)

Se obtiene del documento que emite `--describe` para cada verbo de cada applet declarado, en el orden del registro.

| Campo | De dónde sale en el documento |
|---|---|
| `applet`, `verbo` | `title` (`"<applet> <verbo>"`) |
| `hace` | `description` |
| `argumentos` | `properties.entrada.properties` en su orden, **menos** las banderas globales; cada uno con `obligatorio` (está en `entrada.required`) y `varios` (`type: array`) |
| `banderas` | las propiedades de la entrada de un verbo sin argumentos (`cli.Describir` con `Argumentos` nulo), en su orden, cada una con `con_valor` (tipo distinto de `boolean`) |
| `devuelve` | `properties.salida.then.properties.data`: la definición de `$defs` a la que apunta `$ref` (objeto) o `items.$ref` (lista), con sus `properties` en orden; sin `$ref`, «sin forma declarada» |
| `sobre` | `properties.salida.required` y, para el fallo, `required` de la definición de `else.properties.data.$ref` |

### 2.2 Sintaxis de una orden

`scripts/<applet> <verbo>` seguido de cada argumento en su orden: `<nombre>` si es obligatorio, `<nombre>...` si además
es de varios valores, `[--nombre]` si es opcional. `--describe` declara obligatoriedad y orden pero no si un argumento
va por posición o como bandera; la regla «obligatorio ⇒ por posición» la vigila
`TestTablaDeComandosCoincideConLaGramatica` invocando cada sintaxis generada contra la gramática real (research.md D6).

---

## 3. Enlace de `scripts/`

| Campo | Regla |
|---|---|
| ruta | `skills/<skill>/scripts/<applet>`, uno por cada applet de `kitlegal-applets` |
| tipo | enlace simbólico (un fichero regular con ese nombre es un defecto) |
| destino literal | `../../../bin/instalado/kitlegal` (research.md D5) |
| sobrante | cualquier entrada de `scripts/` que no sea uno de los anteriores |

El destino no resuelve en un clon recién hecho: resuelve tras `make install`, que crea `bin/instalado/kitlegal` (§11).

---

## 4. Norma y referencia generada

### 4.1 Norma (`data/normas.yaml`)

Documento con una sola clave, `normas`: un mapa **indexado por identificador** (research.md D8).

| Campo | Obligatorio | Regla |
|---|---|---|
| clave del mapa | sí | `NORMA`; única: el lector común rechaza la clave repetida antes de validar, nombrando la norma y sus dos líneas («Lectura de documentos YAML», arriba), porque analizar a `yaml.Node` no la comprueba (research.md V43) |
| `titulo` | sí | cadena no vacía; el título oficial tal como lo devuelve la búsqueda grabada (FR-024) |
| `rango` | sí | uno de los valores del vocabulario grabado de la fuente (`enum` de `schemas/normas.yaml.json`, §2 de su contrato) |
| `abreviatura` | no | cadena no vacía |
| `materias` | sí | lista de ≥ 1 cadenas no vacías y sin repetidos |
| cualquier otro (`vertical` incluido) | — | defecto que nombra la norma y el campo |

Mínimo: toda norma de una cita esperada de `evals/boe-legislacion/` y toda norma que `SKILL.md` nombra por número y
año, entre ellas LPAC, LCSP, LRBRL, LGT y TRLRHL (FR-020).

### 4.2 Referencia generada (`references/<n>.md`)

| Campo | Regla |
|---|---|
| primera línea | `<!-- generado desde data/<n>.yaml, no editar -->` (FR-031) |
| contenido de `normas.md` | una fila por norma: título, abreviatura, identificador, rango, materias; ordenadas por año y después por número del identificador, numéricamente (FR-030) |
| determinismo | mismo `data/normas.yaml` ⇒ mismos bytes (FR-034) |

---

## 5. Deriva

Resultado de comparar lo regenerado en memoria con el árbol (FR-042). Cada deriva nombra la skill y la ruta.

| Clase | Cuándo |
|---|---|
| `contenido-distinto` | un fichero generado existe con otros bytes (`references/*.md`, o `SKILL.md` con otra región) |
| `fichero-ausente` | falta un fichero generado |
| `fichero-sobrante` | sobra un fichero en `references/` |
| `enlace-ausente` | falta `scripts/<applet>` |
| `enlace-sobrante` | sobra una entrada en `scripts/` |
| `enlace-con-otro-destino` | `scripts/<applet>` apunta a otro sitio o no es un enlace |

Junto a las derivas, los defectos de la skill: frontmatter (§1.1, §1.3), líneas (§1.2), región ausente o duplicada
(§2), norma inválida (§4.1).

---

## 6. Eval

Fichero `evals/<skill>/<nn>-<descripción>.yaml`, leído con el lector común («Lectura de documentos YAML», al principio:
una clave repetida es un defecto que nombra el fichero, la clave y sus dos líneas) y validado contra
`schemas/eval.yaml.json`.

| Campo | Obligatorio | Regla |
|---|---|---|
| `pregunta` | sí | cadena no vacía |
| `activa` | sí | booleano: si la skill debe activarse |
| `reproduce` | no | nombre de skill (`NOMBRE_DE_SKILL`); declara que la eval reproduce una consulta de esa skill (FR-064) |
| `comandos` | si `activa` | lista de ≥ 1 comandos esperados (§6.1); prohibida si `activa: false` (FR-061) |
| `citas` | si `activa` | lista de ≥ 1 citas esperadas (§6.2); prohibida si `activa: false` |

`Fichero` (tipo Go `Eval`) no es una clave del YAML: es el nombre con el que se leyó el fichero, lo pone `LeerEval`, y
con él nombran la eval las consultas necesarias (§7.1), `Juzgar` y el informe (§10.2). Un fichero del directorio de
evals que no se puede leer como eval es un **fichero mal formado** (`FicheroMalFormado`: `Fichero` y `Error`);
`LeerConjunto` devuelve a la vez las evals bien formadas y los ficheros mal formados, y reserva su `error` para el
directorio que no se puede listar (contrato de evals §1).

### 6.1 Comando esperado (tres formas, excluyentes)

| Forma | Campos | Lo satisface (FR-072) |
|---|---|---|
| **bloque** | `applet`, `norma` (`NORMA`), `bloque` (`BLOQUE`); **sin** `verbo` | una invocación con consulta (§9) y código 0 del mismo applet, verbo `articulo` o `articulos`, esa norma y ese bloque entre los pedidos |
| **consulta de norma** | `applet`, `verbo` ∈ {`indice`, `metadatos`, `analisis`}, `norma` | una invocación con consulta (§9) y código 0 del mismo applet, verbo y norma |
| **búsqueda** | `applet`, `verbo: buscar`, `terminos` (≥ 1 cadenas no vacías) | una invocación con consulta (§9) y código 0 del mismo applet y verbo cuyos argumentos, en minúsculas, contienen cada término como palabra |

Una invocación sin consulta (§9) —la que pide la ayuda, `--help` o `-h`, o lleva `--describe` o `--dry-run` con valor
verdadero— no satisface ninguna forma aunque termine con 0, porque no leyó nada (FR-072, «consultó lo mismo»; FR-008;
research.md D12), y no va a `fuera_de_lo_grabado` ni a `otras_fallidas` (§10.2;
`TestJuzgar/describe-y-dry-run-no-satisfacen`, `/ayuda-no-satisface`). Con `--describe=false` o `--dry-run=false` sí
hay consulta (`TestJuzgar/describe-y-dry-run-falsos-consultan`).

### 6.2 Cita esperada

| Campo | Regla |
|---|---|
| `norma` | `NORMA` |
| `bloque` | `BLOQUE` |

Se compara con las citas extraídas de la respuesta por su forma fija (los corchetes que terminan en
`<norma>, bloque <bloque>]`, con la forma legible delante del identificador dentro de ellos o sin ella; contrato de la
skill §3, research D23): igualdad exacta de la pareja.

### 6.3 Conjunto de evals de `boe-legislacion` (FR-062 a FR-064)

Las reglas no son de la lectura: `LeerConjunto` solo lee y separa las evals bien formadas de los ficheros mal formados.
Las aplica `ComprobarConjuntoDeBoeLegislacion` (contrato de evals §2) sobre las bien formadas y lo que necesita de
`data/normas.yaml`, con un defecto por incumplimiento que lleva el nombre de la regla de esta tabla; la de revisión no es
suya.

| Regla | Comprobación |
|---|---|
| tamaño | entre 10 y 20 ficheros |
| positivas | exactamente 10 con `activa: true` |
| no activación | ≥ 1 con `activa: false` |
| materias distintas | cada positiva cita al menos una norma que no es cita esperada de ninguna otra positiva |
| normas del hito | para cada abreviatura LPAC, LCSP, LRBRL, LGT y TRLRHL, la norma con esa `abreviatura` en `data/normas.yaml` es cita esperada de alguna positiva |
| art. 21 | una eval con `pregunta` exactamente «¿qué dice el art. 21 de la Ley 39/2015?» y cita `BOE-A-2015-10565` + `a21` |
| fiscal | otra positiva cita una norma cuyas `materias` incluyen `tributos` |
| `boe-fiscal` | ≥ 1 con `reproduce: boe-fiscal` |
| normas conocidas | toda norma de una cita o de un comando esperado está en `data/normas.yaml` (FR-020) |
| sin municipio | ninguna eval declara comportamiento de un municipio (FR-066; revisión, no test) |

---

## 7. Consulta necesaria y grabaciones

### 7.1 Consulta necesaria de una eval (FR-074)

Invocación `kitlegal <applet> <verbo> <args…>` que la caché preparada tiene que poder servir. Conjunto sin repetidos:

1. por cada comando esperado: bloque → `articulo <norma> <bloque>`; consulta de norma → `<verbo> <norma>`; búsqueda →
   `buscar <terminos…>`;
2. por cada norma de la eval (de sus citas y de sus comandos con norma): `indice <norma>` y `metadatos <norma>`;
3. por cada cita esperada: `articulo <norma> <bloque>`.

Cada consulta conserva de qué eval y de qué punto sale, para nombrarlo en un fallo.

### 7.2 Manifiesto de grabación de H5 (`testdata/evals/grabaciones.json`)

En `testdata/` de la raíz y no en `internal/evals/testdata/`: un fichero nuevo solo hace pausar al workflow bajo
`testdata/` de la raíz, `internal/source/` o `schemas/` (research.md V36, D11), y en esa pausa graba la persona.

Lo lee un único lector, `LeerManifiesto` (tipos Go `Manifiesto` y `EntradaDelManifiesto`; contrato de evals §3.1), que
rechaza el documento con un miembro desconocido, una clave repetida o algo detrás de su valor (research.md V55).

| Campo | Regla |
|---|---|
| `fuente` | `boe.legislacion-consolidada` |
| `normas[]` | al menos una entrada; toda norma de `data/normas.yaml` tiene la suya: la entrada que la resuelve por su prefijo (lo comprueba `TestIdentificadoresDeLasNormas`, contrato de normas y referencias §6) |
| `normas[].busqueda` | texto no vacío con el que se busca la norma (se graba la búsqueda) |
| `normas[].titulo_empieza_por` | prefijo no vacío que identifica **un único** resultado de esa búsqueda; de él sale el identificador |
| `normas[].bloques` | ids de bloque que se graban de esa norma (los de sus citas y comandos esperados), cada uno `BLOQUE` y sin repetidos; puede faltar o estar vacía |
| `normas[].para` | texto no vacío: por qué está (eval y requisito) |

**Norma repetida.** El manifiesto no tiene identificadores: una entrada resuelve su norma como el único resultado de su
búsqueda cuyo título empieza por su prefijo. Dos entradas cuyos `titulo_empieza_por` son iguales, o uno empieza por el
otro, podrían resolver el mismo título, y son la misma norma repetida: un defecto que nombra las dos entradas (contrato de
evals §3.1, regla 5).

### 7.3 Conjuntos de grabaciones

| Conjunto | Directorio | Dueño |
|---|---|---|
| H4 | `internal/source/boe/testdata/boe.legislacion-consolidada/` | H4; H5 solo lo lee |
| H5 | `testdata/evals/boe.legislacion-consolidada/` | H5; lo llena una persona en la pausa de la tarea del manifiesto (§7.2) |

Unión para reproducir: copia del conjunto H4 y encima el H5 en un directorio temporal. Los dos conjuntos no comparten
nombre de fichero salvo `GET_https_www.boe.es_robots.txt.json`, que la reproducción no usa (`httpx.Replay`, research.md
V21). Lo garantiza el arnés de grabación, que sirve desde H4 toda consulta que H4 ya grabó antes de pedir nada a la red,
también la búsqueda `procedimiento administrativo común` de la primera entrada, que es la misma que grabó H4 y cuyo
fichero tiene el mismo nombre (contrato de evals §3.2, paso 1).

---

## 8. Caché preparada

| Aspecto | Valor |
|---|---|
| Dónde | un directorio temporal nuevo por uso: por test en `make ci`, **por sesión** en el job |
| Cómo se llena | cada consulta necesaria de **todas** las evals del directorio, ejecutada en proceso con `boe` sobre `httpx.Replay` de la unión de grabaciones y la caché en ese directorio |
| Vigencias | las de la fuente de H4: 300 s para `buscar` y `metadatos`, 604 800 s para `indice`, `articulo` y `articulos` (`internal/source/boe/fuente.go`, `vigenciaCorta`, `vigenciaLarga`) |
| Consecuencia | pasada la vigencia, `--offline` la trata como ausente; por eso se prepara justo antes de cada sesión y la sesión tiene un tope de 240 s (research.md D13, D14) |
| Versionado | nunca |

Ciclo de vida en el job: `creada vacía → llena (preparación sin red) → leída por la sesión → descartada`.

---

## 9. Invocación registrada y conexión

Sale de la traza del sistema de la sesión (contrato del job §4), no de lo que el agente escribe. `strace -ff` deja un
fichero `t.<n>` por hilo, con `n` su identificador (research.md V53). `LeerTrazas` recibe si el tope cortó la sesión
(`cortada`, §10.1), que decide qué admite de lo que deja el corte (regla 6), y devuelve las invocaciones ordenadas por el
número de su proceso.

| Campo | Regla |
|---|---|
| `argv` | el de la última `execve` con resultado 0 del proceso (si `bash` se reemplaza por el applet en el mismo proceso, cuenta la del applet) |
| `applet` | `base(argv[0])` si es un applet registrado; si es `kitlegal`, `argv[1]` si es un applet registrado; si no, no es una invocación de applet y se ignora |
| `verbo`, `argumentos` | los tokens que quedan, quitando las banderas globales (con su valor en `--timeout` y `--asunto`, sea `--x v` o `--x=v`; research.md D12) y la ayuda (`--help`, con o sin valor, y `-h`) |
| `consulta` | la del verbo y sus argumentos, tal como la ejecuta el binario: las banderas globales y la ayuda se analizan con la gramática de `cli.Globales` y la ayuda integrada de Kong, como en el binario. **Ninguna** si pide la ayuda —`--help`, también con un valor falso, porque Kong la imprime en cuanto aparece la bandera, o `-h`—, con la que el binario imprime la ayuda y termina con 0 sin leer nada; y ninguna si `--describe` o `--dry-run` valen verdadero —sin valor o con un valor verdadero, como `--describe=true`—, que no consultan. Con `--describe=false` o `--dry-run=false` hay consulta, y también si esas banderas no se pueden analizar (`--describe=quizá`): el binario termina entonces con un error de argumentos, y la invocación es una consulta que falló |
| `codigo` | el de la línea final del fichero del hilo principal del proceso: `+++ exited with N +++` da N; `+++ killed by … +++`, distinto de 0. En una sesión cortada, si ese fichero no tiene línea final (regla 6), la invocación queda **sin código**: no es 0 ni ningún otro número, no satisface ningún comando esperado (§6.1), no va a `fuera_de_lo_grabado` ni a `otras_fallidas` (§10.2) y el informe la presenta con `codigo` `null` |
| `conexiones[]` | las llamadas `connect` atribuidas a la invocación (abajo), cada una con familia, dirección y puerto (o ruta en `AF_UNIX`) y resultado —o sin resultado: la que interrumpió el corte (regla 6) o la que el fin del proceso dejó sin terminar (regla 5), con `?`, `? <unavailable>` o, sin cerrar, nada como texto del resultado—, en orden de fichero, por su número, y de línea |

**Atribución de hilos y conexiones** (la aplica `LeerTrazas`; contrato del job §4):

1. **Origen.** Todo fichero, salvo el raíz —el del proceso que arrancó `strace`, el único sin la línea que lo crea que
   tiene alguna llamada—, lo crea una línea `clone`, `clone3`, `fork` o `vfork` de otro fichero cuyo resultado es su
   número. La única otra excepción es el **fichero huérfano**: un fichero sin esa línea y sin ninguna llamada —solo su
   línea final, o vacío en una sesión cortada— se admite si alguna `clone`, `clone3`, `fork` o `vfork` de la traza es
   una llamada que el fin del proceso dejó sin terminar (regla 5, formas A, B o C): es el hilo que esa llamada pudo
   crear y que el núcleo mató con el proceso antes de ninguna llamada trazada, cuya creación `strace` no vio porque la
   llamada no tiene resultado (research.md V65: 8 de 1000 trazas, siempre con una `clone` de forma A o B). No
   pertenece a ningún proceso, no es una invocación y no tiene conexiones que atribuir. Si ningún fichero sin la línea
   que lo crea tiene llamadas —ninguno, en un `traza/` vacío—, si hay más de uno con llamadas, o si uno sin esa línea y
   sin llamadas no tiene ninguna creación sin terminar que lo explique, la traza es ilegible, con un error que nombra el
   directorio en el primer caso y los ficheros en los otros dos: una línea de creación que no se reconociera no puede
   dejar hilos sin atribuir en silencio, ni una traza vacía dejar `red` vacío. Un directorio que no existe o no se
   puede listar es un error que lo nombra (`TestLeerTrazasSinFicheros`, contrato del job §9).
2. **Proceso.** Un hilo creado por una llamada cuyas banderas incluyen `CLONE_THREAD` —en cualquier orden y con
   cualquier otro argumento— pertenece al proceso del hilo que lo creó, y la atribución es **transitiva**: vale para
   `clone` y para `clone3`, y para el hilo creado por otro hilo que no es el principal. El runtime de Go crea sus hilos con
   `clone`, no con `clone3`, y a menudo desde un hilo que no es el principal (research.md V51 y V53). Un hilo creado sin
   `CLONE_THREAD` (también con `fork` o `vfork`), o el fichero raíz, es el hilo principal de un proceso nuevo.
3. **Invocación.** Un proceso cuya última `execve` con resultado 0 es de un applet es una invocación. Sus `conexiones`
   son las líneas `connect` de su hilo principal posteriores a esa `execve` y las de los hilos del proceso creados,
   directa o transitivamente, desde una línea posterior a ella.
4. **Lo no atribuido se ignora.** Un `connect` de un fichero que no pertenece a ninguna invocación de applet —el de
   `claude` a la API del modelo, que es pública, o el de `bash`— no se atribuye ni se clasifica: si contara, cada
   ejecución daría una llegada a la red.
5. **Líneas.** Cada línea es `execve`, `clone`, `clone3`, `fork`, `vfork` o `connect` con su resultado (`0`, `N` o
   `-1 ERRNO (descripción)`), una línea de señal (`--- SIGNOMBRE {…} ---`, que se admite y no cuenta) o una línea final
   (`+++ exited with N +++`, con N de 0 a 255, el estado de salida del proceso, o `+++ killed by … +++`), y cada
   fichero termina en su línea final. En una llamada, el resultado va tras `= `, y entre el paréntesis de cierre y el
   `=` hay el espacio y, en las llamadas más cortas que la columna de alineación de `strace` (`-a 40`, su valor por
   defecto), el relleno de espacios hasta ella, con el `=` en la columna 41: `vfork()` seguida de 33 espacios y `= N`
   es la línea con la que Claude Code 2.1.270 de x86_64 crea los procesos de sus órdenes en el runner (research.md V63;
   supuesto S4). Ahí solo caben espacios, uno o más, y el resultado sigue siendo obligatorio, salvo en la llamada que
   **el fin del proceso deja sin terminar**, que se admite en cualquier sesión, cortada o no, porque no la produce el
   corte sino que el proceso salga con `exit_group` desde otro hilo mientras uno está dentro de una llamada: el
   binario de Go sale en cuanto escribe su respuesta, mientras su runtime crea un hilo o mientras otro hilo espera en
   un `connect` (research.md V65; en el runner, `t.14465` de la sesión 09 de la prueba de red del intento 6 de T030,
   terminada con código 0). `strace` la escribe de cuatro formas, los estados `unfinished` y `unavailable` de su
   opción `-e status=`: **(A)** `<llamada>(<argumentos>) = ?`, el hilo que llegó a su parada de fin dentro de la
   llamada, con ` <unfinished ...>` delante del paréntesis de cierre solo si al decodificador de la llamada le quedaba
   algo por escribir a la salida —la `clone` de amd64, que escribe `tls=0x…` a la salida, la lleva; la de arm64 y
   todo `connect`, que escribe su dirección a la entrada, salen sin ella—; **(B)** `<llamada>(<argumentos>) = ?
   <unavailable>`, la que llegó a su parada de salida sin que `strace` pudiera leer el resultado del hilo, que ya
   moría; **(C)** `<llamada>(<argumentos> <unfinished ...>`, la línea de entrada que la línea final del hilo deja sin
   cerrar, sin paréntesis de cierre ni resultado; y **(D)** `???( <unfinished ...>`, sin argumentos, la de un hilo
   recién creado que murió en la parada de entrada de una llamada que `strace` no llegó a identificar y que no se
   ejecutó. Y ninguna otra: la marca con cualquier otro resultado (`= N`, `= -1 ERRNO (…)`, `= ? ERRNO (…)` o
   `= ? <unavailable>`) o dentro de los argumentos, `???(` con argumentos o con resultado y la forma sin cerrar de una
   llamada que no es del filtro siguen siendo ilegibles. Es una llamada sin resultado, a la que en su fichero solo
   pueden seguir líneas de señal y la línea final, como la de la regla 6: no crea ningún hilo (el fichero del que pudo
   crear lo admite la regla 1), una `execve` así no es una `execve` con resultado 0, un `connect` así se clasifica solo
   por su dirección (tabla de abajo: `red` fuera del bucle local, porque nada muestra que no llegara, FR-076), su
   `resultado` es el texto tras `= ` (`?` o `? <unavailable>`; vacío en la forma C) y `???` no es ninguna llamada
   del filtro ni tiene conexión que atribuir. El relleno de alineación cabe igual delante de `= ?`. La traza se toma
   sin `-e signal=none`, que suprimiría `+++ killed by … +++` de todo proceso que muere por una señal (research.md
   V54). Cualquier otra línea, una cortada sin su resultado que no sea ninguna de esas cuatro formas o un fichero sin
   línea final hacen la traza ilegible, con un error que nombra el fichero, el número de línea y su texto; también un
   `connect` atribuido a una invocación con una familia distinta de `AF_INET`, `AF_INET6` y `AF_UNIX`. Así, un formato
   de la traza distinto del comprobado (research.md S4) hace que la sesión no pase y lo dice, en lugar de dejar `red`
   vacío. Los ficheros se leen en orden de número, cada uno línea a línea hasta su final, y el primer defecto de una
   línea o del final de un fichero es el error; el de origen (regla 1) se comprueba después, sobre todos. La única
   otra excepción es la de la regla 6, y solo en una sesión cortada: sin corte, `strace` escribe la línea final de
   cada fichero, también tras una llamada sin terminar (research.md V53, V54 y V65).
6. **Sesión cortada por el tope.** Con `cortada` (`codigo-de-la-sesion` 124 o 137, §10.1), `LeerTrazas` admite lo que
   puede dejar el corte y nada más. Al agotarse el tope, la señal llega a los procesos de la sesión: una llamada
   bloqueante que interrumpe queda con el resultado `? ERRNO (descripción)`, como
   `? ERESTARTSYS (To be restarted if SA_RESTART is set)`, seguida de sus líneas de señal y de `+++ killed by SIGTERM +++`;
   y si `timeout` tiene que matar `strace` con `KILL` (137), los ficheros de los procesos que seguían vivos se quedan sin
   línea final, y vacío el de un hilo que no había hecho ninguna llamada trazada (research.md V54 (1) a (3)). En esa
   sesión, por tanto: (a) un fichero puede no terminar en su línea final, o estar vacío; (b) una llamada de la regla 5
   puede tener el resultado `? ERRNO (descripción)` si en su fichero solo la siguen líneas de señal y, como mucho, la
   línea final: es una llamada sin resultado, así que una `execve` sin resultado no es una `execve` con resultado 0, una
   `clone`, `clone3`, `fork` o `vfork` sin resultado no crea ningún hilo, y un `connect` sin resultado se clasifica solo
   por su dirección (tabla de abajo). Ese resultado seguido de otra llamada, o en una sesión sin corte, sigue siendo un
   defecto de la regla 5: `? ERRNO (descripción)` lo deja solo la señal del tope, y es distinto de la llamada que el
   fin del proceso deja sin terminar (`?`, `? <unavailable>`, sin cerrar o `???(`), que la regla 5 admite en cualquier
   sesión y no necesita `cortada`. Las reglas 1 a 4, y la 5 en todo lo demás, se aplican igual; la
   invocación cuyo fichero de hilo principal no tiene línea final queda sin código (campo `codigo`). Así una sesión
   cortada lleva el motivo del tope y no `sesión ilegible` (§10.1), y sus conexiones y sus invocaciones se informan
   (§10.2).

| Clase de conexión | Cuándo |
|---|---|
| `local` | `AF_UNIX`, o dirección de bucle local (`127.0.0.0/8`, `::1`) |
| `bloqueada` | a otra dirección, con resultado de error distinto de `EINPROGRESS` |
| `red` | a otra dirección, con resultado `0` o `EINPROGRESS` (se toma como llegada a la red), o sin resultado porque el corte interrumpió la llamada (regla 6) o el fin del proceso la dejó sin terminar (regla 5): nada muestra que no llegara, y FR-076 no deja sin detectar una llegada a la red |

`destino` de una conexión, como lo presenta el informe (§10.2): `<dirección>:<puerto>` en `AF_INET` (`127.0.0.1:9`),
`[<dirección>]:<puerto>` en `AF_INET6` (`[2001:db8::7]:443`) y `unix:<ruta>` en `AF_UNIX`
(`unix:/var/run/nscd/socket`).

---

## 10. Sesión, resultado de eval e informe

### 10.1 Sesión

| Campo | De dónde |
|---|---|
| `modelo` | mensaje `system`/`init` del transcript `stream-json`, campo `model` |
| `version_de_claude_code` | el mismo mensaje, `claude_code_version` |
| `activada` | algún bloque `tool_use` con `name: "Skill"` e `input.skill` igual al nombre de la skill |
| `respuesta` | mensaje `result` con `subtype: "success"` e `is_error: false`, campo `result`; en otro caso, vacía |
| `codigo` | el entero de `codigo-de-la-sesion`, que el guion escribe siempre (contrato del job §3.2); si falta o no es un entero, la sesión es ilegible, nunca 0 |
| `fin` | texto fijo del último mensaje del transcript: si es `result`, `result <subtype>`, seguido de ` con is_error` si `is_error` es verdadero (`result success`, `result error_max_turns`, `result success con is_error`); si es otro mensaje, su `type` (`system`, `assistant`, `user`); si el transcript no tiene mensajes, `sin mensajes`. No es el motivo de una sesión sin terminar (abajo), que se escribe aparte |
| `terminada` | `codigo` es 0 **y** el último mensaje es `result` con `subtype: "success"` e `is_error: false` |
| `cortada` | `codigo` es 124 o 137: el tope cortó la sesión. `EscribirInforme` se lo pasa a `LeerTrazas`, que admite en la traza lo que deja el corte (§9, regla 6) |
| `salida_de_error` | contenido de `sesion.err`, que el guion escribe siempre (vacío si no hubo salida de error); si falta, la sesión es ilegible |
| `invocaciones[]` | §9 |

Motivo de una sesión sin terminar, en este orden: `tope de 240 s agotado (código 124)`; `terminada por señal tras el
tope (código 137)`; `código N`; `sin mensaje result`; `result con subtype <subtype>`; `result con is_error`. 124 y 137
son los códigos con los que `timeout` termina al agotar el tope y al tener que enviar `KILL`, y si la sesión termina
antes, `strace` devuelve el de `claude` (research.md V52 y V53; en el runner, S10). Esos dos códigos eligen el texto del
motivo y si la traza se lee como la de una sesión cortada (`cortada`), nunca si la sesión pasa: con cualquier código
distinto de 0 la sesión no terminó. Una sesión cortada lleva su motivo, y no `sesión ilegible`, por lo que el corte deja
en su traza (§9, regla 6; research.md V54): sus invocaciones, sus conexiones, sus llegadas a la red y sus invocaciones
fuera de lo grabado se informan como las de cualquier otra sesión (§10.2). Si su traza tiene otro defecto, es ilegible
como la de cualquier sesión.

### 10.2 Resultado de eval (FR-071, FR-072)

| Campo | Regla |
|---|---|
| `sesion` | nombre del directorio de la sesión (contrato del job §3.2); lo pone `EscribirInforme` |
| `eval` | nombre del fichero de eval con el que se juzga la sesión, leído de su `eval.txt` sin el salto de línea final, o vacío si `eval.txt` falta; dos sesiones pueden compartirlo (la de la prueba de red se juzga con la eval 01, contrato del job §6). La sesión solo se juzga si nombra una eval bien formada del directorio (contrato del job §3.3, paso 2) |
| `activa`, `activada` | lo esperado y lo observado; la activación coincide si son iguales |
| `codigo_de_la_sesion`, `sesion_terminada` | `codigo` y `terminada` de la sesión (§10.1) |
| `respuesta` | la `respuesta` de la sesión (§10.1), vacía si no la hay o si la sesión no se pudo leer. La sección de la sesión en `informe.md` lleva la pregunta de su `pregunta.txt` y esta respuesta (contrato del job §5): de ahí sale la aceptación (FR-080, SC-001, SC-002) |
| `fin_de_la_sesion` | el `fin` de la sesión (§10.1), con su texto fijo; vacío si la sesión no se pudo leer, como la `respuesta` |
| `comandos_ejecutados`, `comandos_ausentes` | reparto de los comandos esperados según §6.1 |
| `citas_encontradas`, `citas_ausentes` | reparto de las citas esperadas según §6.2 |
| `invocaciones[]` | todas las invocaciones de applet de la sesión (§9), en orden de número de proceso, cada una con `orden` (el applet seguido de los argumentos que le siguen en `argv`, separados por un espacio), `codigo` (`null` si quedó sin código, §9) y `conexiones[]`: una entrada por pareja distinta de `destino` y `clase` (§9), en el orden en que aparece por primera vez, o ninguna si la invocación no conectó. Así el informe muestra qué conexiones hizo cada invocación y de qué clase, también las `local` y `bloqueada`, que no cambian nada pero son la evidencia de que la traza se leyó (research.md S4) |
| `fuera_de_lo_grabado[]` | invocaciones con consulta y código 4 o 5, cada una con `orden` y `codigo` |
| `otras_fallidas[]` | invocaciones con consulta y otro código distinto de 0, cada una con `orden` y `codigo`; no las que quedaron sin código, que no terminaron y solo están en `invocaciones` |
| `llegadas_a_la_red[]` | una entrada por cada invocación y `destino` de sus conexiones de clase `red`, con `orden` y `destino` |
| `motivos[]` | por qué no pasa, uno por causa: la sesión ilegible —`sesión ilegible: <fichero>: <error>`, uno por cada fichero que falta o no se puede leer, en el orden `eval.txt`, `pregunta.txt`, sesión y traza, y el de un `eval.txt` que no nombra ninguna eval bien formada (contrato del job §3.3, paso 2)— o sin terminar (con su motivo de §10.1), la activación que no coincide, cada comando ausente y cada cita ausente; vacío si pasa |
| `pasa` | la sesión **terminó**, `activa == activada` y ningún comando ni cita ausente. Sin la primera condición, una eval de no activación cuya sesión murió sin activar nada pasaría en vacío. **No** lo cambian `fuera_de_lo_grabado`, `otras_fallidas` ni `llegadas_a_la_red` (FR-076) |

FR-072 dice cuándo una eval **solo** puede pasar; exigir además que la sesión terminara no afloja ninguna de sus
condiciones: sin una sesión terminada no hay activación ni respuesta observadas que comparar.

### 10.3 Informe y veredicto global

| Campo | Regla |
|---|---|
| `skill` | `InformeAEscribir.Skill`, tal cual |
| `modelo` | `InformeAEscribir.Modelo`, tal cual: el id fijado en el job (FR-070, FR-071, SC-003); no se toma de las sesiones |
| `modelos_de_sesion[]` | el `modelo` (§10.1) de cada sesión que `LeerSesion` lee sin error, sin repetir y en el orden en que aparece cada uno por primera vez, recorriendo las sesiones por nombre; vacía si ninguna sesión se pudo leer. No se toma de `InformeAEscribir.Modelo` |
| `versiones_de_claude_code[]` | con la misma regla, la `version_de_claude_code` (§10.1) de esas sesiones |
| `commit` | `InformeAEscribir.Commit`, tal cual (FR-071, SC-003) |
| `sin_python` | el contenido íntegro del fichero `InformeAEscribir.SinPython`, igual byte a byte, con su salto de línea final y sin recortar ni reformatear nada: la constancia de cómo se comprobó que no había Python (FR-081; contrato del job §3.1). Si no se puede leer, `EscribirInforme` devuelve un error que nombra la ruta y no escribe el informe (contrato del job §3.3), como cuando no puede leer `Evals` o `Sesiones` o escribir en `Destino`; lo fija `TestEscribirInformeSinSusEntradas` (contrato del job §9) |
| cabecera de `informe.md` | las cuatro líneas literales y la sección `## Comprobación sin Python` del contrato del job §5, con los valores de `modelo`, `modelos_de_sesion`, `versiones_de_claude_code`, `commit` y `sin_python` de `informe.json` |
| `ficheros_mal_formados[]` | los `MalFormados` que devuelve `LeerConjunto` sobre las evals con las que se juzga (§6; contrato de evals §1), cada uno con `fichero` y `error`; en `informe.md`, una línea por fichero con su error, o «ninguno» |
| `evals[]` | §10.2, uno por sesión: con la prueba de red, dos resultados comparten la eval 01 y se distinguen por `sesion` |
| `motivos[]` | en este orden: los `motivos` de cada eval que no pasa, cada uno precedido del nombre de su sesión; `<fichero>: sin ninguna sesión` por cada eval bien formada que ningún `eval.txt` nombra, en orden de fichero; `ninguna eval bien formada que juzgar` si `LeerConjunto` no da ninguna; `<fichero>: mal formado: <error>` por cada fichero mal formado, en orden de fichero y con el mismo `error` de `ficheros_mal_formados`; y `<sesión>: petición llegada a la red: <orden> → <destino>` por cada entrada de `red`, en su orden |
| `fuera_de_lo_grabado` | todas las de los resultados, cada una con `sesion` y `eval` de su resultado, `orden` y `codigo` (en `informe.md`, tabla sesión · eval · orden · código), o «ninguna» |
| `red` | todas las `llegadas_a_la_red`, cada una con `sesion` y `eval` de su resultado, `orden` y `destino` (en `informe.md`, tabla sesión · eval · orden · destino), o «ninguna petición llegó a la red de una fuente» |
| `veredicto` | `fallo` si hay ficheros mal formados, si alguna eval no pasa (también por una sesión sin terminar o ilegible), si alguna eval bien formada no tiene ninguna sesión que la juzgue (ningún `eval.txt` la nombra), si no hay ninguna eval bien formada o si alguna petición llegó a la red; `aprobado` en otro caso. Así no aprueba un informe que no evaluó todas las evals: sin la sesión de una eval, con el directorio de sesiones vacío o sin ninguna eval que juzgar |

---

## 11. Instalación

### 11.1 Estado de `~/.claude/skills/<nombre>` antes de instalar

| Estado | Qué hace `make install` |
|---|---|
| no existe | lo crea como enlace a `<raíz física del repositorio>/skills/<nombre>` |
| enlace cuyo destino literal es ese directorio | lo deja igual |
| cualquier otra cosa (directorio, fichero, enlace a otro sitio, enlace roto) | conflicto: termina con código distinto de 0 nombrándolo, y no crea ni cambia nada |

### 11.2 Enlace del binario instalado

| Campo | Regla |
|---|---|
| ruta | `bin/instalado/kitlegal` del repositorio (ignorado por `.gitignore`, `/bin/`) |
| destino | la ruta absoluta que da `go list -f '{{.Target}}' ./cmd/kitlegal`, donde `go install` dejó el binario |
| idempotencia | se rehace igual en cada instalación |
