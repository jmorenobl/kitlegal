# Contrato: la skill de laboratorio, sus variantes y sus evals

FR-010 a FR-014, FR-020 a FR-029; SC-003, SC-004, SC-005. Data-model §2 a §5; research D4 a D6.

## 1. Dónde vive

```text
laboratorio/
├── variantes.yaml
├── evals/
│   ├── 01-constitucion-articulo-140.yaml
│   ├── 02-lgt-prescripcion.yaml
│   ├── 03-lrjsp-principio-de-legalidad.yaml
│   └── 04-et-vacaciones.yaml
└── variantes/
    ├── control/skills/articulos-del-boe/SKILL.md
    ├── instrucciones/{instrucciones.txt,skills/articulos-del-boe/SKILL.md}
    ├── entrada/skills/{articulos-del-boe,consulta-legal}/SKILL.md
    ├── referencia-secundaria/skills/articulos-del-boe/{SKILL.md,references/cierre.md}
    ├── parametro/{parametro.txt,skills/articulos-del-boe/SKILL.md}
    ├── referencia-paso/skills/articulos-del-boe/{SKILL.md,references/cierre.md}
    ├── referencia-en-el-punto/skills/articulos-del-boe/{SKILL.md,references/cierre.md}
    └── descripcion-sin-herramientas/skills/articulos-del-boe/SKILL.md
```

Fuera de `skills/` (no se empotra, research V10), de `evals/` (el job no lo lee) y de lo común y lo que mide el job
(ningún cambio suyo elige una skill, research V9). Nada de `laboratorio/` lo genera `skills-sync`: se escribe a mano.

## 2. `variantes.yaml`

```yaml
# Variantes de la skill de laboratorio de H39: la de control y siete que difieren de ella en las partes que declaran
# (specs/031-h39-c-omo-se/data-model.md §2 y §3). Lo comprueba TestVariantesDelLaboratorio.
skill: articulos-del-boe
entrada: consulta-legal
herramienta: boe_articulo
parametro: bloque
referencia: cierre.md
variantes:
  - {nombre: control, modos: [orden, herramienta], cambia: []}
  - {nombre: instrucciones, modos: [herramienta], cambia: [instrucciones]}
  - {nombre: entrada, modos: [orden, herramienta], cambia: [entrada]}
  - {nombre: referencia-secundaria, modos: [orden, herramienta], cambia: [cuerpo, referencia]}
  - {nombre: parametro, modos: [herramienta], cambia: [cuerpo, parametro]}
  - {nombre: referencia-paso, modos: [orden, herramienta], cambia: [cuerpo, referencia]}
  - {nombre: referencia-en-el-punto, modos: [orden, herramienta], cambia: [cuerpo, referencia]}
  - {nombre: descripcion-sin-herramientas, modos: [orden, herramienta], cambia: [descripcion]}
```

## 3. La regla testigo

- **Párrafo de la regla** (`<REGLA>`), el mismo en todos los sitios donde va:

  > La respuesta termina con una última línea, sola y sin ningún formato —ni negrita, ni cursiva, ni comillas, ni
  > código—, que dice exactamente: Sello de lectura: KL-7Q4

- **Guion** (`reglaTestigo(respuesta string) bool`, `internal/evals`): cumple si la última línea no vacía de la
  respuesta, sin espacios ni `\r` a los lados, es `Sello de lectura: KL-7Q4`.
- **Respuestas fijas** (`internal/evals/testdata/laboratorio/respuestas/`), con el veredicto en el nombre:
  `cumple.txt` (el sello al final), `cumple-con-lineas-en-blanco.txt` (y líneas vacías detrás), `no-cumple-sin-sello.txt`,
  `no-cumple-sello-no-al-final.txt`, `no-cumple-sello-en-negrita.txt` (`**Sello de lectura: KL-7Q4**`) y
  `no-cumple-sello-con-texto.txt` (`Sello de lectura: KL-7Q4.`). Y la respuesta del producto: la de la sesión
  `09-constitucion-articulo-140-claude-sonnet-5-5-01` de `specs/030-h32-arquitectura-de-skills/gates/evals/normativa.json`,
  leída del informe, que no cumple (research V20).
- **`TestReglaTestigo`**: cada respuesta fija con su veredicto; y la línea del sello no aparece en ningún fichero de
  `skills/` ni en `mcp.Instrucciones` (FR-013, FR-014, SC-003).

## 4. La skill de control (`articulos-del-boe`)

Forma de las del producto (FR-011): frontmatter con `name` y `description`, sin `metadata` (no declara applets, así
que no lleva tabla generada ni nada que `skills-sync` regenere), cuerpo por debajo de 11 000 caracteres y descripción
por debajo de 900, en tercera persona.

- **`description`** de control:

  > Lee y cita con kitlegal un artículo de una norma consolidada del Boletín Oficial del Estado (BOE), de cualquier
  > materia: busca su bloque en el índice de la norma y lee su texto vigente. Se usa siempre que la pregunta pida qué
  > dice un artículo de una ley, de un real decreto o de la Constitución, también cuando el agente crea saberlo. Se
  > activa antes de boe_indice o boe_articulo, o de kitlegal boe indice o kitlegal boe articulo: qué pedir y cómo
  > responder lo dice la skill.

- **Cuerpo** de control, en este orden: un título; una frase de qué hace («nada de lo que diga la respuesta sale de
  memoria: sale del texto que devuelve la herramienta»); «Cómo se lee un artículo», con los pasos numerados —la norma
  por su identificador (con los cuatro de las evals: Constitución Española, BOE-A-1978-31229; Ley 58/2003, General
  Tributaria, BOE-A-2003-23186; Ley 40/2015, de Régimen Jurídico del Sector Público, BOE-A-2015-10566; Estatuto de los
  Trabajadores, BOE-A-2015-11430), el bloque en el índice (`kitlegal boe indice <norma>` o `boe_indice`; el bloque es
  el campo `id` de la entrada del artículo) y la lectura (`kitlegal boe articulo <norma> <bloque>` o `boe_articulo`)—;
  y «Cómo se responde», con la frase de la cita («cita la norma y el bloque así: [BOE-A-1978-31229, bloque a140].») y
  detrás, en párrafo propio, `<REGLA>`.

## 5. Lo que cambia cada variante

Todo lo que no se nombra aquí es el de control, byte a byte.

| Variante | Cambio |
|---|---|
| `instrucciones` | `instrucciones.txt`: el texto de `mcp.Instrucciones`, un espacio, «Antes de la primera herramienta de kitlegal, activa la skill articulos-del-boe.» y un salto de línea |
| `entrada` | `skills/consulta-legal/SKILL.md`: `description` «Punto de entrada de cualquier pregunta legal sobre una norma española: dice qué skill la resuelve y la activa. Se usa al empezar cualquier pregunta legal, antes que las demás skills, también cuando el agente cree saber la respuesta.»; cuerpo: un título y «Para saber qué dice un artículo de una norma del BOE, activa la skill articulos-del-boe antes de ninguna herramienta: dice qué pedir y cómo responder.» |
| `referencia-secundaria` | Sin `<REGLA>` en el cuerpo; la frase de la cita termina en «[BOE-A-1978-31229, bloque a140] (la forma del final de la respuesta está en [references/cierre.md](references/cierre.md)).» |
| `parametro` | Sin `<REGLA>` en el cuerpo, sin nada en su lugar; `parametro.txt`: la ayuda de `bloque`, un espacio, `<REGLA>` y un salto de línea |
| `referencia-paso` | Sin `<REGLA>` en el cuerpo; primer paso de «Cómo se lee un artículo», delante de los demás, que se renumeran: «Antes de la primera llamada a una herramienta, lee [references/cierre.md](references/cierre.md): dice cómo termina la respuesta.» |
| `referencia-en-el-punto` | `<REGLA>` sustituido en su sitio por «Antes de escribir la respuesta, lee [references/cierre.md](references/cierre.md): dice cómo termina.» |
| `descripcion-sin-herramientas` | `description` sin la frase «Se activa antes de boe_indice o boe_articulo, o de kitlegal boe indice o kitlegal boe articulo:», con «Qué pedir y cómo responder lo dice la skill.» como última frase |

`references/cierre.md`, el mismo en las tres de referencia: «# Cómo termina la respuesta», una línea vacía y `<REGLA>`.
El texto de `instrucciones.txt` y de `parametro.txt` es el fichero sin su salto de línea final.

## 6. Las evals (`laboratorio/evals/`)

Las cuatro, con la pregunta, `activa: true`, `comandos` y `citas` de las evals 09, 04, 07 y 10 de `normativa` y un
comentario de cabecera que lo dice (data-model §4). Ejemplo, `01-constitucion-articulo-140.yaml`:

```yaml
# Eval de laboratorio de H39: la pregunta y la lectura de la eval 09 de normativa, servidas por testdata/evals/.
pregunta: "¿Qué dice el artículo 140 de la Constitución?"
activa: true
comandos:
  - applet: boe
    norma: BOE-A-1978-31229
    bloque: a140
citas:
  - norma: BOE-A-1978-31229
    bloque: a140
```

## 7. Controles

- `TestVariantesDelLaboratorio` (`internal/evals`): la declaración (§2) con su lector estricto; cada variante frente a
  la de control, parte a parte (data-model §3), con la regla testigo donde la pone cada una; y la forma del producto de
  cada variante, con `skills.Regenerar` y `skills.Comparar` sobre `laboratorio/variantes/<nombre>`, como
  `TestSkillsDelRepositorio` sobre la raíz, sin defectos ni deriva: las skills del laboratorio no declaran applets, así
  que no hay tabla que generar (research V11) (FR-011, FR-020 a FR-028, SC-004).
- `TestEvalsDelLaboratorio` (`internal/evals`): data-model §4 (FR-012).
- `TestReglaTestigo` (`internal/evals`): §3 (FR-013, FR-014, SC-003).
- `TestLaboratorioFueraDelProducto` (`internal/evals`): ningún fichero de `kitlegal.Skills()` tiene la ruta de una skill
  del laboratorio ni los bytes de un fichero de `laboratorio/`; `app.Arrancar` con `skills install --dir <vacío>` no deja
  ninguno; la definición del job (`.github/workflows/evals.yml`) no ejecuta ninguna de las dos skills, no hay
  `evals/<skill>` de ninguna, y ninguna ruta de `laboratorio/` mide (`rutaQueMide`) (FR-010, SC-005).
