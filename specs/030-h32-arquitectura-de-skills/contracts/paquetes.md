# Contrato · El molde de un paquete de materia (FR-012, FR-019, FR-021 a FR-024)

## 1. Qué es un paquete y dónde vive

Un paquete de materia es, y no es más que (ADR 0044, punto 5):

| Pieza | Dónde | Quién la escribe |
|---|---|---|
| su nombre | el valor `vertical: <nombre>` de sus normas en `data/normas.yaml` | la persona que añade la materia |
| sus normas | las entradas de `data/normas.yaml` con `vertical: <nombre>` | ídem; el identificador y el título, de lo grabado (`TestIdentificadoresDeLasNormas`) |
| una referencia por skill de capacidad que la necesite | `skills/<skill>/references/<nombre>.md`, en `references/` plano (ADR 0046) | a mano, con las tablas que salen de `data/` generadas entre marcas (§3) |
| sus filas de cobertura | las de `data/cobertura.yaml` cuyo `precepto.norma` es una norma del paquete | la persona |
| sus evals | `evals/<skill>/NN-*.yaml` de la skill que las hereda | la persona |

No hay fichero de declaración del paquete: el conjunto de paquetes es el de los valores de `vertical` de
`data/normas.yaml`, y la referencia de un paquete en una skill es el `references/<nombre>.md` cuyo nombre es uno de
ellos. `vertical` tiene la forma de un nombre de fichero, `^[a-z][a-z0-9-]*$`, y es un adjetivo («lo <nombre>»:
administrativo, fiscal, laboral), porque así lo lee la condición del enlace (§2). `schemas/normas.yaml.json` admite el
campo, opcional.

**El paquete de lo administrativo**: `vertical: administrativo` en las diez normas de la materia de validación
(principio X) —LPAC, LRJSP, LRJPAC, LTAIBG, LRBRL, LCSP, LGS, LGP, TRLRHL y LJCA—; ni la CE, ni el CC, la LEC y la LOPJ
(la base: lo constitucional, lo común y lo procesal), ni la LGT, la LIRPF, el ET ni la LOPDGDD (otras materias)
llevan ninguna (research D5). Su única referencia es `skills/plazos-y-territorio/references/administrativo.md`
(`borradores/administrativo.md`); `normativa` y `jurisprudencia` no necesitan ninguna. Sus filas de cobertura son las
nueve de hoy, todas con su precepto en la LPAC.

## 2. El enlace desde cada `SKILL.md` de capacidad (FR-023)

Cada `SKILL.md` de capacidad lleva, una vez, estas dos marcas, cada una en su línea, en el paso que lee las
referencias de los paquetes (`normativa`: paso 1; `jurisprudencia`: tras la introducción del protocolo;
`plazos-y-territorio`: «Referencias»):

```text
<!-- paquetes: lo genera make skills-sync -->
<!-- fin de los paquetes -->
```

Entre ellas, `make skills-sync` escribe una línea por cada paquete que tiene referencia en esa skill, en el orden de
los nombres:

```text
- Si la pregunta es de lo <nombre>, lee [<nombre>](references/<nombre>.md).
```

Ejemplo, con sus bytes, el de `plazos-y-territorio` en la entrega (93 caracteres y su salto; la de la vertical de
ensayo, 69 y su salto):

```text
- Si la pregunta es de lo administrativo, lee [administrativo](references/administrativo.md).
```

Un paquete no edita ningún `SKILL.md`: ni su cuerpo fuera de las marcas, ni su frontmatter. `make skills-check` sale en
rojo si el texto entre las marcas no es el que genera la sincronía (deriva `contenido-distinto`), y si una skill tiene
la referencia de un paquete y no tiene las marcas (defecto `referencia de paquete sin enlazar: <nombre>.md`).

## 3. Las tablas generadas dentro de una referencia escrita a mano (FR-012, FR-019)

Una referencia a mano puede llevar una región generada por cada generador de `internal/skills` que sale de `data/`:

```text
<!-- inicio de <generador>: generada desde data/<fichero>[ y data/<fichero>] con make skills-sync, no editar -->
…
<!-- fin de <generador> -->
```

`make skills-sync` reescribe lo que hay entre las dos marcas; `make skills-check` da la deriva si difiere. La marca de
inicio nombra cada fichero de `data/` del que sale la tabla: es la «cabecera generado desde» de la región (FR-019), y
la elección de las evals la lee para saber qué datos carga la skill (FR-038; [evals.md](evals.md) §5). Los tres
generadores que admiten región en la entrega:

| Generador | Datos | Filas | Columnas |
|---|---|---|---|
| `cobertura_por_materia` | `data/cobertura.yaml` y `data/normas.yaml` | las filas cuyo `precepto.norma` lleva el `vertical` del paquete del fichero | `Procedimiento`, `Línea que da la respuesta` |
| `recursos_y_plazos` | `data/plazos.yaml` y `data/normas.yaml` | todos los tipos, como hoy | las de `c991808`; `Precepto` en la forma única |
| `leyes_vertebrales` | `data/normas.yaml` | las de `vertebral: true`, como hoy | las de `c991808` |

**La forma legible de un precepto** de las dos primeras: `art. <apartado>` si el apartado empieza por cifra, el apartado
si no, un espacio y la abreviatura de su norma en `data/normas.yaml` (o, sin abreviatura, el título hasta la primera
coma), y detrás `[<identificador>, bloque <id>]`: `art. 124.1 LPAC [BOE-A-2015-10565, bloque a124]`,
`DA 1ª.2.a LPAC [BOE-A-2015-10565, bloque daprimera]`.

**La línea de cada fila de la cobertura** (FR-012) es una de las dos formas `⚠ NO CUBIERTO:` del bloque `text` del
`SKILL.md` de la skill que tiene la referencia —la que lleva `<norma>` si la fila trae `norma`; la otra si no— con cada
marcador sustituido:

| Marcador | Valor |
|---|---|
| `<procedimiento>` | `procedimiento` de la fila |
| `<norma>` | `norma.nombre` |
| `<identificador de la norma>` | `norma.identificador` |
| `<precepto>` | la forma legible de `precepto` (arriba) |
| `<identificador>` | `precepto.norma` |
| `<id>` | `precepto.bloque` |

Ejemplo, con sus bytes, la fila de la aplicación de los tributos (US2, escenario 3):

```text
| aplicación de los tributos en materia tributaria y aduanera y su revisión en vía administrativa | ⚠ NO CUBIERTO: aplicación de los tributos en materia tributaria y aduanera y su revisión en vía administrativa se rige por Ley 58/2003, General Tributaria (BOE-A-2003-23186), no por el procedimiento común: kitlegal no calcula sus plazos ni sus recursos (DA 1ª.2.a LPAC [BOE-A-2015-10565, bloque daprimera]). |
```

Defectos de `make skills-check` que añade el molde, cada uno con su fichero:

- una región con un generador que no admite región, o cuyas marcas no van en pareja;
- una región `cobertura_por_materia` en una skill cuyo `SKILL.md` no tiene las dos formas `⚠ NO CUBIERTO:`, o con un
  marcador que la tabla de arriba no conoce;
- una fila de `data/cobertura.yaml` que ninguna región `cobertura_por_materia` genera (FR-012: «por cada fila»).

El estado que crea algo que no es el binario ni la persona que escribe el paquete —marcas editadas a mano en otro
orden, un fichero de `references/` que no es un `.md`— es el defecto genérico de hoy: lo da `skills-check` como
cualquier otro, sin casos nuevos.

## 4. La lista de paquetes de la skill de entrada (FR-001, punto 3)

`consulta-juridica` lleva otras dos marcas, en «3. Qué cubre la base»:

```text
<!-- paquetes de la base: lo genera make skills-sync -->
<!-- fin de los paquetes de la base -->
```

Entre ellas, una línea por paquete, en el orden de los nombres, con las skills de capacidad que tienen su referencia,
en el orden de sus nombres:

```text
- lo administrativo: `plazos-y-territorio`.
```

## 5. La vertical de ensayo (FR-024)

Vive en `internal/skills/testdata/vertical-de-ensayo/`, con la misma forma que en el repositorio:

```text
internal/skills/testdata/vertical-de-ensayo/
├── data/normas.yaml          # dos normas con vertical: ensayo, que se añaden a las de data/normas.yaml
├── data/cobertura.yaml       # una fila, con su precepto en la primera de las dos, que se añade a regulados_aparte
├── skills/normativa/references/ensayo.md            # a mano
├── skills/jurisprudencia/references/ensayo.md       # a mano
├── skills/plazos-y-territorio/references/ensayo.md  # a mano, con la región cobertura_por_materia vacía
└── evals/normativa/24-ensayo-norma-del-paquete.yaml # una eval que cita la primera norma
```

Las dos normas son de una materia que no existe, con identificadores que ninguna búsqueda del BOE devuelve
(`BOE-A-1900-1` y `BOE-A-1900-2`, `Ley 1/1900` y `Ley 2/1900`, materia «ensayo»): son datos de un test, no normas que
la skill vaya a leer. `TestVerticalDeEnsayo` (`internal/skills/vertical_de_ensayo_test.go`, paquete `skills_test`):

1. copia `skills/`, `data/` y `evals/normativa/` del repositorio a `t.TempDir()`, y anota la huella de cada fichero
   del repositorio;
2. instala la vertical: añade sus entradas a `data/normas.yaml` y `data/cobertura.yaml` de la copia y copia sus
   ficheros de `skills/` y `evals/`;
3. corre la sincronía (`skills.Regenerar` y `skills.Escribir`) sobre la copia y después la comprobación
   (`skills.Regenerar` y `skills.Comparar`): ningún defecto y ninguna deriva;
4. carga la eval con `evals.LeerEval`;
5. compara la copia con el repositorio: ningún `.go`, y cada `SKILL.md` igual fuera de las marcas de §2 y §4;
6. comprueba que las huellas de `skills/` y `data/` del repositorio no han cambiado.

Su mutante, en el mismo test: el molde instalado con un `SKILL.md` editado fuera de las marcas (una línea con el enlace
a `ensayo.md` escrita a mano en el cuerpo de `normativa`, como haría un paquete sin sincronía) pone en rojo el paso 5.

## 6. Ninguna norma de un paquete en un cuerpo de capacidad (FR-021)

La lista sale de `data/normas.yaml` de la rama: por cada norma con `vertical`, su identificador, su título hasta la
primera coma («Ley 39/2015», «Real Decreto Legislativo 2/2004») y su abreviatura. Se le suma la lista mínima del spec.
Se busca en el cuerpo de los `SKILL.md` de capacidad, desde la línea que sigue al `---` que cierra el frontmatter: la
descripción no cuenta. Las abreviaturas, como palabra entera. Los ejemplos de `normativa` pasan a la Constitución
(`BOE-A-1978-31229`, bloque `a140`, que ningún paquete marca y que la caché de las sesiones sirve: eval 09).
