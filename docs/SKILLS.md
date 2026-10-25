# Guía de autoría de las skills

Para quien escribe o revisa una skill de `skills/`: qué va en el cuerpo de `SKILL.md` y qué en una referencia, cómo se
escribe la descripción, qué presupuestos hay y por qué, y la regla de `references/`. La decide el ADR 0046 y la nombra
el principio VIII de la constitución. `CONTRIBUTING.md` («Skills y evals») dice cómo se regenera lo generado y qué
comprueba cada objetivo del `Makefile`; las reglas legales de cada skill —qué cita, qué avisa, qué declara no
cubierto— no son de esta guía, sino de su `SKILL.md`, y las deciden sus evals.

Lo que tiene forma lo comprueba `make ci` sin red ni modelo: `make skills-check` y los controles de contenido. Lo que
es significado, la tercera persona de la descripción, lo juzga un modelo en el job de evals. Cada defecto de
`skills-check` es una línea `<skill>: <fichero>: <regla>`, con la medida y el límite cuando los hay, y
`make skills-sync` da los mismos y no escribe nada mientras quede alguno.

- [Las dos fuentes](#las-dos-fuentes)
- [1. Qué va en el cuerpo y qué en una referencia](#1-qué-va-en-el-cuerpo-y-qué-en-una-referencia)
- [2. La descripción](#2-la-descripción)
- [3. Los presupuestos y su porqué](#3-los-presupuestos-y-su-porqué)
- [4. La regla de `references/`](#4-la-regla-de-references)
- [5. Los defectos de `skills-check`](#5-los-defectos-de-skills-check)
- [6. Antes de proponer el cambio](#6-antes-de-proponer-el-cambio)

## Las dos fuentes

Consultadas el **2026-10-23**. Lo que esta guía cita de ellas es lo que registra el ADR 0046 («Contexto y problema» y
«Decisión»), con esa fecha; no se han vuelto a consultar para escribirla.

- **La especificación Agent Skills** (agentskills.io, «Specification»). Carga cada skill en tres niveles: el nombre y
  la descripción de todas desde el principio; el cuerpo de `SKILL.md` entero cuando se activa —«< 5000 tokens
  recommended», en menos de 500 líneas—; y los ficheros de `references/` solo cuando hacen falta, a un nivel de
  `SKILL.md`. La descripción admite como mucho 1024 caracteres.
- **La guía de autoría de Anthropic** («Skill authoring best practices»). Añade la descripción en tercera persona
  —«The description is injected into the system prompt, and inconsistent point-of-view can cause discovery
  problems»—, el índice en las referencias de más de 100 líneas, las referencias enlazadas directamente desde
  `SKILL.md`, y que ni el nombre ni la descripción lleven etiquetas XML y que el nombre no contenga `anthropic` ni
  `claude`.

kitlegal es más estricto que los ejemplos de la guía de Anthropic, que escriben «Use when…»: en español «Úsala» es un
imperativo de tú, y aquí la descripción va en tercera persona (ADR 0046, «Consecuencias»). La validación de la
especificación es la de `internal/skills`, en Go y sin red; `skills-ref`, la de la especificación, es Python y
`make ci` no la usa (ADR 0046, punto 3).

## 1. Qué va en el cuerpo y qué en una referencia

El agente tiene siempre el nombre y la descripción de todas las skills; el cuerpo de `SKILL.md`, entero, cuando
activa una; y una referencia, solo si la lee. Lo que va a una referencia el modelo puede no leerlo. De ahí el reparto
(ADR 0046, punto 1).

**En el cuerpo, lo que toda respuesta de la skill necesita:**

- el protocolo, con sus pasos en orden;
- las reglas invariantes;
- las formas fijas que las evals comprueban sin modelo en la respuesta, las líneas `⚠` (`⚠ SIN CONSULTA AL BOE:`,
  `⚠ SENTENCIA NO COMPROBADA:`, `⚠ NO CUBIERTO:`, las de aviso, las de la redacción en una fecha, las del plazo…),
  cada una tal cual, en un bloque `text`;
- la tabla de comandos, que `make skills-sync` genera entre sus marcas desde `--describe` del binario y no se edita;
- y, por cada referencia, cuándo leerla, con su enlace (§4).

**En una referencia, el detalle que solo pide una clase de pregunta**: cómo se pide cada orden con herramientas, cómo
se busca una norma que no está en la tabla, qué se hace con cada código de salida, lo que la ley regula aparte en un
plazo, qué se dice de un documento que no es el pedido.

Cuándo leer cada referencia lo dice el cuerpo de una de dos maneras. Con un apartado «Referencias», una línea por
referencia, como `legal-core`:

```markdown
- [municipios.md](references/municipios.md): al presentar un territorio, si falta el municipio o no sale con 0, y
  para explicar su cobertura.
```

o en el paso del protocolo que la necesita, con el enlace en la frase, como `boe-legislacion` y `jurisprudencia`:
«si no está, búscala por su título con [buscar_la_norma](references/buscar_la_norma.md)».

Una referencia es de una de dos clases, y la clase la decide la regla de §4:

- **A mano**: el protocolo de detalle de su skill, escrito sin la cabecera «generado desde». No repite lo que sale de
  `data/`, que tiene su referencia generada. Las de hoy abren con un título `# ` y un párrafo que dice lo que llevan y
  desde qué paso se leen; ningún gate lo exige.
- **Generada**: lo que sale de `data/*.yaml` —la tabla de normas, la jerarquía normativa, la cobertura por materia…—,
  con su cabecera `<!-- generado desde data/<fichero>, no editar -->`, declarada en `kitlegal-referencias` del
  frontmatter y escrita por `make skills-sync`. `data/` es su única fuente de verdad: se cambia el YAML, no la
  referencia.

El cuerpo y las referencias a mano le hablan al agente, en imperativo («toma», «búscala»); la tercera persona es de la
descripción (§2), que es lo único que juzga el juez de las descripciones.

**Mover un texto a una referencia a mano no lo saca de ningún control** (ADR 0046, punto 5). Los controles de
contenido de `make ci` leen el `SKILL.md` y cada referencia a mano por separado, y cada defecto nombra su fichero; las
referencias generadas solo las lee `sin-instrucciones-de-evals`:

| Control (subtest) | Skills | Lo que exige |
|---|---|---|
| `prosa-de-la-skill` | `boe-legislacion`, la que tiene `evals/<skill>/expresiones-prohibidas.yaml` | Ninguna expresión de la lista fuera del código y ninguna fecha AAAAMMDD con cifras, en ningún fichero |
| `ordenes-para-powershell` | `boe-legislacion` | Cada orden de lectura o de comprobación lleva detrás, en el mismo fichero, su forma para PowerShell |
| `linea-sin-consulta` | `boe-legislacion`, `legal-core` | La línea `⚠ SIN CONSULTA AL BOE:`, sola en un bloque `text` del `SKILL.md` |
| `formas-de-la-redaccion-fechada`, `avisos-de-la-skill`, `hallazgos-de-la-skill` | `boe-legislacion` | Cada forma, en el `SKILL.md` |
| `lineas-de-aviso-de-las-referencias` | `boe-legislacion`, `legal-core` | Cada línea `⚠` de una referencia a mano es, sin los blancos de los extremos, una línea del `SKILL.md` |
| `ordenes-de-ejemplo` | Las que tienen evals | Cada orden concreta de `kitlegal boe` —un verbo y al menos un argumento, sin marcador `<…>`, `...` ni `…`— sale con 0 con `--offline` sobre la caché de una sesión de su skill (ADR 0043) |
| `normas-nombradas` | Todas | Cada norma nombrada por su rango, número y año («Ley 39/2015», «Real Decreto Legislativo 2/2004») empieza un título de `data/normas.yaml` |
| `sin-instrucciones-de-evals` | Todas | Ninguna instrucción de las evals en el `SKILL.md`, sin la tabla de comandos, ni en ningún fichero de `references/` |

Así, una forma fija llevada solo a una referencia pone en rojo su control, y una variante mal escrita de ella en una
referencia también, con el defecto `<skill>: references/<fichero>.md:<n>: la línea ⚠ no es ninguna de SKILL.md: «…»`:
la referencia no puede enseñarle al agente una forma que la eval no reconoce. Y un ejemplo que la caché de las
sesiones no sirve sin red —hoy, `buscar` y `analisis`; sí `indice`, `metadatos` y `articulo`, con y sin `--fecha`, de
las normas de `data/normas.yaml`— se escribe con marcadores, o no se escribe.

## 2. La descripción

La descripción es lo que el agente lee, junto con las de las demás skills, para decidir cuál activa: va al prompt de
sistema, y un punto de vista distinto entre skills estorba esa elección (la guía de Anthropic). Dice **qué hace la
skill y cuándo se usa, en tercera persona**: no le habla a quien la lee ni habla como alguien. Por la misma razón, una
propuesta que cambia una descripción hace que el job de evals las mida todas (ADR 0045, motivo `descripcion`).

| En tercera persona | En primera o segunda |
|---|---|
| «Se usa cuando se pregunta…», «Se activa antes de llamar a…», «Conviene activarla» | «Úsala cuando…», «Actívala antes de…», «Úsela» |
| «aunque el agente crea conocer la respuesta», «cuando la persona trae el texto» | «también cuando creas conocer la respuesta», «tu pregunta», «te preguntan», «si os piden» |
| «Consulta y cita normativa…», «Calcula el último día del plazo», con la skill de sujeto | «Te ayudo a…», «puedo calcular», «nuestras herramientas» |
| «¿hasta cuándo puedo recurrir?», una pregunta entre comillas | |

Dos de los casos correctos son **los que una lista de palabras no distingue**, y por eso la tercera persona la juzga un
modelo y no una lista, que «marcaría lo correcto o dejaría pasar lo que no lo es» (ADR 0046; ADR 0037):

- **La pregunta citada.** La primera o la segunda persona dentro de una cita de lo que alguien dice o pregunta, entre
  comillas —«¿hasta cuándo puedo recurrir?», «¿puedes decirme qué dice el artículo 21?»—, es correcta.
- **El verbo con la skill de sujeto.** En español «Consulta», «Lee» o «Calcula» es a la vez la tercera persona que dice
  lo que hace la skill y el imperativo de tú. En «Lee el índice y los artículos con kitlegal y responde citando
  identificador y bloque» el sujeto es la skill, y es correcto; solo es segunda persona si la misma frase lleva otra
  marca de segunda persona, como un pronombre o un posesivo.

Tampoco cuentan la primera o la segunda persona dentro de un nombre propio, un identificador, un título o un texto
escrito como código entre comillas invertidas. La forma de la descripción la comprueba `skills-check`: como mucho 900
caracteres (§3) y ninguna etiqueta XML (§5).

### El juez de las descripciones

La tercera persona no la comprueba `make ci`: la juzga en el job de evals el juez con modelo de las evals (ADR 0037),
con la clase `habla_en_primera_o_segunda_persona` y la rúbrica, el esquema, los casos y la medida de
`evidencias/adr-0046/` (ADR 0046, punto 4), cuya huella `make ci` compara con la de su manifiesto. Su pregunta cerrada
es la de `rubrica.md`: «¿Hay en la descripción, fuera de las citas, alguna forma de primera o de segunda persona?».

- **Cuándo juzga.** En una ejecución sobre una propuesta de cambio que cambia, respecto de su base, la descripción del
  `SKILL.md` de alguna skill: entonces el trabajo `evals (<skill>)` de cada skill juzga la descripción de la suya, haya
  cambiado o no; sin ninguna descripción cambiada, o fuera de una propuesta, ningún trabajo lleva juicio, umbral ni
  apartado en el informe. Si una medida de `main` ya tiene el juicio de la misma descripción con la misma huella —la
  de la descripción, los cuatro ficheros del juez, su modelo y su versión—, lo reutiliza, salvo en la suite completa,
  y lo reutilizado no cuenta en el umbral. Nunca en `make ci` ni en un paso de un run (ADR 0032): sus tests votan con votos simulados o con un `claude`
  sustituto.
- **Cómo vota.** Con `claude-opus-5-5` y Claude Code 2.1.289, los fijados para el juez de las evals; cada voto recibe
  la descripción entre `<descripcion skill="…">` y `</descripcion>`. La marcan tres «si» unánimes, pedidos por orden,
  cada uno con una frase que se comprueba sin modelo en la descripción; como mucho seis votos.
- **Qué decide.** El umbral `habla_en_primera_o_segunda_persona` es 0 y decide: una descripción es un texto fijo, no
  una muestra. Marcada, el job falla, y el apartado «La descripción» de `informe.md` da la frase y el motivo; sin
  juzgar, el job falla con su propio motivo, «el juez de las descripciones dejó sin juzgar la descripción de
  <skill>: …».
- **Medido.** Antes de abrir ninguna sesión, el trabajo comprueba sin modelo que la medida versionada corresponde a la
  rúbrica, a los casos, al modelo y a la versión y que se cumple: 63 de 63 defectos marcados y 0 de 36 correctos,
  entre estos los dos casos de arriba (`evidencias/adr-0046/LEEME.md`). Si cambia alguna de las cuatro, una persona
  repite la medida con la etiqueta `evals-medir-juez` (`make evals-medir-juez SKILL=descripciones`).

El juez no valora si la descripción es buena ni si hace que se active la skill: eso lo deciden las evals de
activación. Con una descripción marcada, se reescribe en tercera persona la frase que cita el informe y se repite el
job con la etiqueta `evals`.

## 3. Los presupuestos y su porqué

| Qué | Límite | Por qué |
|---|---|---|
| Líneas del cuerpo | Menos de 500 (499 como mucho) | Las de la especificación |
| Caracteres del cuerpo | Como mucho 11 000 | Con 2,26 caracteres por token, unos 4870 tokens, por debajo de los 5000 de la especificación |
| Caracteres de la descripción | Como mucho 900 | El 88 % de los 1024 de la especificación, que la de `boe-legislacion` rozaba con 1023 |

**El cuerpo** es el `SKILL.md` desde el byte que sigue a la línea `---` que cierra el frontmatter hasta el final, con
la tabla de comandos como la deja `make skills-sync`: la tabla generada cuenta. Un `SKILL.md` sin frontmatter cerrado
no tiene cuerpo que medir, y su defecto es el del frontmatter.

**Los caracteres se cuentan como runas, no como bytes**: «á», «—» y «⚠» cuentan uno. Un cuerpo de 11 000 caracteres
con más de 11 000 bytes está en el límite, y uno de 11 001 caracteres, fuera. Las líneas se cuentan como las contaba
el límite anterior: los saltos de línea, más uno si el fichero no termina en uno.

**Por qué en caracteres.** Lo que carga el contexto son los tokens, pero no hay tokenizador sin red, y la suscripción
no da la API de recuento (ADR 0046, «Opciones consideradas»). La razón es una medida: el 2026-10-23, con
`claude-sonnet-5-5`, la diferencia de tokens de entrada de una sesión con el cuerpo y sin él dio 2,26 caracteres por
token en el español de las tres skills. Es una aproximación, porque la razón depende del texto: **se vuelve a medir, y
la cifra se revisa con una propuesta de cambio, cuando cambia el modelo que decide** (ADR 0031).

Las 500 líneas de la especificación ya se cumplían antes de H31, con el límite anterior, de menos de 300 líneas del
fichero entero, que se cumplía con líneas de 120 columnas: lo que carga el contexto son los tokens, y el presupuesto
que decide es el de caracteres. Al cerrar H31:

| Skill | Cuerpo antes de H31 | Cuerpo | Descripción |
|---|---|---|---|
| `boe-legislacion` | 283 líneas, 24 439 caracteres | 152 líneas, 10 993 caracteres | 878 caracteres |
| `legal-core` | 283 líneas, 19 225 caracteres | 168 líneas, 10 944 caracteres | 888 caracteres |
| `jurisprudencia` | 193 líneas, 13 637 caracteres | 159 líneas, 10 859 caracteres | 842 caracteres |

Sin margen en el cuerpo, una regla nueva cabe sacando a una referencia el detalle que solo pide una clase de pregunta
(§1), nunca una forma fija ni una regla que toda respuesta necesita. Una materia nueva (ADR 0044) puede entrar como
referencia sin agrandar el cuerpo.

## 4. La regla de `references/`

**`references/` es plano**: solo ficheros regulares `.md`. Un subdirectorio, un enlace simbólico o un fichero de otra
extensión es un defecto (`references/notas.txt: no es un fichero .md`, `references/anexos: no es un fichero .md`), y
`make skills-sync` no lo retira ni escribe nada. Un subdirectorio sin ningún fichero que se pueda empotrar (vacío, o
con solo ficheros que empiezan por punto, como `.DS_Store`) no llega a ese defecto: el `//go:embed` de `skills.go` no
compila, y el error de Go lo nombra (`cannot embed directory skills/<skill>/references/<dir>: contains no embeddable
files`). Solo pasa en local, porque git no guarda directorios vacíos; se arregla borrando el directorio.

**Cada fichero se clasifica primero por la declaración** (FR-012 de H31), en este orden:

| El fichero | Es | `make skills-check` | `make skills-sync` |
|---|---|---|---|
| `.md` declarado en `kitlegal-referencias` | La referencia generada, con o sin cabecera | Deriva si no es lo que genera su fila | La rehace |
| Declarado y ausente | — | Deriva | La escribe |
| `.md` no declarado que abre con `<!-- generado desde data/<fichero>, no editar -->` | Sobrante | Deriva | Lo retira |
| `.md` no declarado que no abre con la cabecera | Una referencia a mano | Solo los gates de esta sección | No la toca, byte a byte |

De ahí que una referencia a mano no abra con la cabecera ni lleve el nombre de una declarada, y que una generada nueva
se declare en `kitlegal-referencias` con su generador y se escriba con `make skills-sync`.

**Cada referencia, a mano o generada, está enlazada desde `SKILL.md`** con un enlace de Markdown, fuera de los bloques
de código, cuyo destino es su ruta relativa tal cual: `[como_se_cita](references/como_se_cita.md)`, también con
`#apartado`. No cuentan el mismo texto en código en línea, el enlace dentro de un bloque ni `./references/…`.

**Ninguna referencia enlaza otra de su skill**: la especificación las quiere a un nivel de `SKILL.md`. Cuenta el
destino de un enlace de la referencia, sin `#…` y resuelto desde `references/`, que es otra referencia de la skill:
`varios_bloques.md`, `./varios_bloques.md` y `../references/varios_bloques.md` cuentan; un `#apartado`, el enlace a la
propia referencia y un destino que no es ninguna referencia, no. Si una referencia necesita otra, el cuerpo dice
cuándo leer cada una.

**La de más de 100 líneas, a mano o generada, abre con un índice**: una lista —líneas que empiezan, tras su sangría,
por `- `, `* ` o `<n>. `— antes de su primer apartado `## `, en la que cada apartado `## ` tiene, en su orden, un
elemento que contiene su texto, sin `[`, `](…)`, `*`, `_` ni `` ` ``. Los elementos de más, como los de los apartados
`### `, no estorban. Tiene que haber algún `## `: una de más de 100 líneas sin ninguno también es un defecto. Con 100
líneas o menos no se exige nada. Por ejemplo:

```markdown
# Cómo se cita

- [Lo que hace cita](#lo-que-hace-cita)
- [Los avisos](#los-avisos)

## Lo que hace cita
…
## Los avisos
…
```

Los generadores de las referencias generadas no escriben hoy ningún índice, y ninguna pasa de 100 líneas: el cambio de
`data/` que lleve una por encima adapta su generador en el mismo cambio, porque sin el índice `skills-check` la da como
defecto.

**Un bloque de código** va de una línea que abre una valla —tras hasta tres espacios, tres o más `` ` `` o `~`— a la
que la cierra con la misma marca, repetida al menos tantas veces y sin nada detrás salvo blancos, o al final del
fichero. Los enlaces, los apartados y la lista del índice se leen fuera de los bloques.

**Lo que viaja con la skill.** Todo fichero de `references/` va dentro del binario, y `kitlegal skills install` lo
instala con su skill: el agente de quien usa el kit tiene cada referencia donde el cuerpo la enlaza
(`TestSkillsEmpotradas`). Y una referencia a mano cuenta en el conjunto de dependencias de cada eval de su skill: una
propuesta que solo la cambia elige las evals de esa skill y las de las skills que la nombran, y una skill nombrada solo
en una referencia a mano entra en el conjunto con su `SKILL.md` y sus `references/` (ADR 0045). Las generadas no se
leen para nombres: entran por su fichero de `data/`.

## 5. Los defectos de `skills-check`

Los da `TestSkillsDelRepositorio` (`internal/app/skills_test.go`) con lo que regenera `skills.Regenerar`, y cada uno
lleva su mutante en rojo y su límite en verde. Con la skill de ejemplo, `boe-legislacion`:

| Regla | Defecto |
|---|---|
| Las del frontmatter de la especificación (nombre de como mucho 64 caracteres, en `a-z`, `0-9` y `-`, sin guion al principio, al final ni dos seguidos e igual al directorio; descripción presente y no vacía; compatibilidad de como mucho 500; solo las claves `name`, `description`, `license`, `allowed-tools`, `compatibility` y `metadata`) y las de kitlegal (applets registrados, referencias declaradas con su generador) | `SKILL.md: ` y la regla, p. ej. `boe-legislacion: SKILL.md: name vacío` |
| `name` con `anthropic` o `claude`, como subcadena | `boe-legislacion: SKILL.md: name "claude-legal" contiene «claude»` |
| Etiqueta XML en `name` | `boe-legislacion: SKILL.md: name con la etiqueta XML «<x>»` |
| Etiqueta XML en `description` | `boe-legislacion: SKILL.md: description con la etiqueta XML «<ejemplo>»` |
| `description` de más de 900 caracteres | `boe-legislacion: SKILL.md: description de 901 caracteres (máximo 900)` |
| Cuerpo de 500 líneas o más | `boe-legislacion: SKILL.md: cuerpo de 500 líneas (máximo 499)` |
| Cuerpo de más de 11 000 caracteres | `boe-legislacion: SKILL.md: cuerpo de 11001 caracteres (máximo 11000)` |
| Entrada de `references/` que no es un fichero `.md` | `boe-legislacion: references/notas.txt: no es un fichero .md` |
| Referencia sin enlace desde `SKILL.md` | `boe-legislacion: references/como_se_cita.md: SKILL.md no la enlaza con [](references/como_se_cita.md)` |
| Referencia que enlaza otra | `boe-legislacion: references/como_se_cita.md: enlaza otra referencia, references/varios_bloques.md` |
| Más de 100 líneas, sin índice | `boe-legislacion: references/como_se_cita.md: 101 líneas sin un índice que nombre sus apartados ## antes del primero (más de 100)` |
| Más de 100 líneas, sin ningún `## ` | `boe-legislacion: references/como_se_cita.md: 101 líneas sin ningún apartado ## (más de 100)` |
| Un `scripts/` en la skill (ADR 0019) | `boe-legislacion: una skill no lleva scripts/ (ADR 0019)` |

**Una etiqueta XML** es un texto de la forma `<nombre…>`, `</nombre>` o `<nombre/>`, con `nombre` empezando por una
letra (`</?\p{L}[^<>]*>`): lo son `<ejemplo>`, `</b>`, `<br/>` y `<descripcion skill="x">`; no lo son «< 5000», «a<b»,
«a < b > c», `<>` ni `<1>`. El defecto cita la primera. Sustituyen a las reglas anteriores el límite de 900 al de 1024,
que ya no puede darse, y la etiqueta XML a la de «description con < o >», que daba como defecto «< 5000».

En los límites, en verde: 499 líneas, 11 000 caracteres de cuerpo, 900 de descripción y 100 líneas de referencia sin
índice o sin ningún `## `.

## 6. Antes de proponer el cambio

```sh
make skills-sync                                                      # regenera la tabla y las referencias generadas
make skills-check                                                     # los gates de esta guía, sin escribir nada
go test -count=1 -run '^TestSkillsDelRepositorio$' ./internal/app/   # los mismos, con sus mutantes
go test -count=1 -run '^(TestEvalsDelRepositorio|TestControlesSobreLasReferenciasAMano)$' ./internal/evals/
```

Las dos últimas ejecutan los controles de contenido de §1 con sus mutantes: `normas-nombradas` y
`sin-instrucciones-de-evals` son subtests de `TestSkillsDelRepositorio`, y los demás, de `TestEvalsDelRepositorio`,
con los mutantes de `TestControlesSobreLasReferenciasAMano`. Para ver las medidas de un cuerpo, que cuentan
como el gate mientras el `SKILL.md` termine en salto de línea:

```sh
LC_ALL=en_US.UTF-8 awk 'f>=2{print} /^---/{f++}' skills/<skill>/SKILL.md | LC_ALL=en_US.UTF-8 wc -lm
```

Todo va dentro de `make ci`. El cambio de una skill sube su versión en `CHANGELOG.md` con lo que cambia, y lo que la
skill responde lo deciden sus evals en el job (`CONTRIBUTING.md`, «Job de evals»), que con una descripción cambiada
también la juzga (§2).
