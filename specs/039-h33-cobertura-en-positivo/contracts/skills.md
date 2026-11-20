# Contrato: lo que cambia en las skills, sus referencias y la documentación

FR-004, FR-005, FR-007 a FR-010, FR-020, FR-022; SC-003, SC-005. Los textos son los del spec; aquí van con sus cortes
de línea, que son los de `h33-prototipo` (`27413343`), comprobados con `git diff` y con las medidas de research V1
(4481 y 10 686 caracteres). Nada más de esos ficheros cambia: ni el frontmatter ni ninguna `description` (ADR 0055 y
0057), ni `skills/normativa` ni `skills/jurisprudencia`, ni `calculo_de_plazos.md`.

## 1. `skills/consulta-juridica/SKILL.md` (FR-007, FR-008)

**§3.** Las cuatro viñetas de las capacidades (de `- **Normativa**:` a la línea que termina en «con su línea.» de los
plazos y la del territorio) pasan a ser la región de lo cubierto, que `make skills-sync` llena (contracts/
datos-y-regiones.md §3). Tras `### 3. Qué cubre la base` y su línea en blanco:

````text
<!-- lo que cubre la base: lo genera make skills-sync -->
- **Normativa**: leer y citar cualquier norma consolidada del BOE, estatal o autonómica, de cualquier materia y en cualquier fecha.
- **Jurisprudencia**: preparar la consulta exacta de una sentencia, que la persona busca con su navegador, y cotejar el documento que trae. No consulta ningún buscador ni dice si una sentencia existe.
- **Plazos**: solo los del procedimiento administrativo común, los del acceso a la información pública y la interposición del recurso contencioso-administrativo, con los festivos de los municipios que cuentan. Lo que la ley regula aparte de ese procedimiento lo declara `plazos-y-territorio` con su línea.
- **Territorio**: la provincia, la comunidad, el régimen, el DIR3 del ayuntamiento y los boletines de un municipio.
<!-- fin de lo que cubre la base -->
````

Detrás, sin cambios, la viñeta «Paquetes de materia» con su región y la de «Sin herramienta, no cubiertos».

**§4.** El título, el párrafo y el bloque, en lugar del título y el párrafo de hoy; detrás, el párrafo «Por qué», sin
cambios:

````text
### 4. Lo que no cubre, con su línea

Lo que la pregunta pide y no resuelve ninguna capacidad de la lista del 3 —el plazo de una materia que no es la del
procedimiento administrativo común, a quién reclamar, qué vía seguir, un escrito— lleva su línea, con el motivo «fuera
de lo que kitlegal cubre». Lo que no pide no lleva ninguna, tampoco el cómputo de una fecha que no ha pedido. Esta línea
no es para lo que la ley regula aparte del procedimiento común: eso lo declara, con su propia línea, la skill de la
capacidad que lo pide, y tras ella no leas su norma. Lo que leas con la skill de normativa para completarla va detrás,
citado, con su marca: el texto citado es el de la norma; lo que no está verificado es que se aplique a este asunto. No
des de memoria ningún plazo, órgano ni vía, ni apliques los de otra materia o los del procedimiento común, ni pidas
`kitlegal plazos calcular` para un asunto de otra materia: ese plazo no es el suyo.

```text
⚠ NO CUBIERTO: <asunto>: <motivo>. kitlegal no da sus plazos, órganos, vías ni escritos.
⚠ NO VERIFICADO: <lo encontrado>: <forma legible> [<identificador>, bloque <id>]; kitlegal no ha comprobado que se aplique a este asunto.
```
````

El cuerpo pasa de 3810 a 4481 caracteres. Lo fija `TestConsultaJuridicaDesdeLaBase` con su huella nueva (contracts/
controles.md §4).

## 2. `skills/plazos-y-territorio/SKILL.md` (FR-009)

Tres cambios de líneas enteras, que son la declaración del subtest de `plazos-y-territorio` de
`TestSkillsDelProductoDesdeControl` (contracts/controles.md §3):

**Paso 7.** Antes:

```text
   su fila, entera, **ni `plazos calcular`** ni ninguna fecha (regla 9). Si no dice la materia, **no la preguntes ni
```

Después:

```text
   su fila, entera, **ni `plazos calcular`** ni ninguna fecha (regla 9); si es de otra materia, igual, con el motivo de
   lo no declarado. Si no dice la materia, **no la preguntes ni
```

**«Cómo se presenta un plazo».** Antes:

````text
Lo no cubierto (paso 7) es la línea de su fila en la referencia del paquete, que es esta forma con sus datos: la
primera si la entrada trae la norma que lo rige; si no, la segunda:

```text
⚠ NO CUBIERTO: <procedimiento> se rige por <norma> (<identificador de la norma>), no por el procedimiento común: kitlegal no calcula sus plazos ni sus recursos (<precepto> [<identificador>, bloque <id>]).
⚠ NO CUBIERTO: <procedimiento> lo regula aparte la ley, no el procedimiento común: kitlegal no calcula sus plazos ni sus recursos (<precepto> [<identificador>, bloque <id>]).
```
````

Después:

````text
Lo no cubierto (paso 7 y regla 9) va con esta forma: en un asunto de la cobertura del paquete, la línea de su
fila, entera; en otra materia, con el motivo «fuera de lo que kitlegal cubre»:

```text
⚠ NO CUBIERTO: <asunto>: <motivo>. kitlegal no da sus plazos, órganos, vías ni escritos.
```
````

**Regla 9.** Antes:

```text
   plazo, un órgano o una vía de otra materia no están cubiertos: dilo, sin darlos de memoria.
```

Después:

```text
   plazo, un órgano o una vía de otra materia tampoco: su línea, sin darlos de memoria ni pedir `plazos calcular`.
```

El cuerpo baja de 10 882 a 10 686 caracteres; con la línea de la vertical de ensayo, 10 756. La plantilla es la misma
línea, byte a byte, que la de `consulta-juridica`, y no lleva la marca.

## 3. `skills/plazos-y-territorio/references/administrativo.md` (FR-005)

En «Lo que la ley regula aparte», entre la última viñeta y la marca de inicio de la región, se insertan la frase, el
bloque y una línea en blanco (antes, la línea en blanco y la marca; después, lo de abajo). Lo demás escrito a mano no
cambia; la región la regenera `make skills-sync` con la plantilla (contracts/datos-y-regiones.md §5).

````text
- Si la conversación no dice la materia, no cae en ninguna fila. Al responder, di que la fecha no vale si el asunto es
  de los de esta tabla, nombrándolos solo con ella.

Los motivos de la línea; el primero es el de la fila con norma:

```text
se rige por <norma> (<identificador de la norma>), no por el procedimiento común (<precepto> [<identificador>, bloque <id>])
lo regula aparte la ley, no el procedimiento común (<precepto> [<identificador>, bloque <id>])
```

<!-- inicio de cobertura_por_materia: generada desde data/cobertura.yaml y data/normas.yaml con make skills-sync, no editar -->
````

## 4. Documentación (FR-022)

Sin control en `make ci` (FR-022: se verifica en la revisión y con el quickstart, escenario 9):

| Fichero | Cambio |
|---|---|
| `docs/SKILLS.md` | la plantilla única y la marca; los motivos en la referencia del paquete; la región de lo cubierto y sus marcas; en la tabla de defectos de la región, «la plantilla» donde dice «las dos formas» y la fila de los motivos (contracts/datos-y-regiones.md §5); la fila de `consulta-juridica` de la tabla de las skills sin «en prosa» ni «sin… líneas `⚠`» |
| `CONTRIBUTING.md` | la fila de `no_cubierto` con `fuera_de_lo_declarado`; la fila nueva de `tambien_la_resuelven`; la de `cobertura_sin_declarar` con la descripción de FR-014; los umbrales de `consulta-juridica`, con `sin_activar` y `cobertura_sin_declarar` |
| `.github/workflows/evals.yml` | solo el comentario del tope: «15 013 s en consulta-juridica, 250,2 minutos, con sus cuatro evals» donde dice «8485 s …, 141,4 minutos, con sus dos evals» |
| `docs/HISTORIAL.md` | la viñeta del hito |
| `docs/USO.md` | la entrada de la bitácora de FR-022, con fecha del día de la tarea |
| `CHANGELOG.md` | en «[Unreleased]», «Cambiado»: la plantilla única, lo cubierto en datos y las evals de la frontera |
| `web/src/pages/como-se-evalua/index.astro` | la frase de la línea 225, que anuncia H33, pasa a decir lo hecho |

Ningún ADR: el 0058 ya está.
