# Contrato: la skill `skills/boe-legislacion/`

FR-001 a FR-015, FR-077, SC-004. Estructura de `SKILL.md`, formato de cita y reglas. El texto definitivo lo escribe la
tarea de la skill; este contrato fija lo que ese texto **tiene que contener** y lo que no puede contener.

## 1. Ficheros

| Ruta | Qué es | Quién la escribe |
|---|---|---|
| `skills/boe-legislacion/SKILL.md` | frontmatter, protocolo, formato de cita, tabla de comandos (región generada) y reglas | persona o ejecutor; la región, `make skills-sync` |
| `skills/boe-legislacion/references/normas.md` | tabla de normas | solo `make skills-sync`, desde `data/normas.yaml` |
| `skills/boe-legislacion/scripts/boe` | enlace simbólico a `../../../bin/instalado/kitlegal` | solo `make skills-sync` |

Nada más en el directorio (data-model §1). Ningún fichero invoca Python ni ningún intérprete (FR-013, SC-007).

## 2. `SKILL.md`

### 2.1 Frontmatter

```yaml
---
name: boe-legislacion
description: >-
  Consulta y cita normativa consolidada del Boletín Oficial del Estado (BOE) de cualquier materia: procedimiento
  administrativo, contratación pública, régimen local, tributos y haciendas locales, régimen jurídico del sector
  público, transparencia, relaciones laborales… Úsala cuando se pregunte qué dice un artículo, una ley o un real
  decreto; cuando se nombre una norma por su número y año (Ley 39/2015, Real Decreto Legislativo 2/2004), por su
  abreviatura (LPAC, LCSP, LRBRL, LGT, TRLRHL, LRJSP) o por su identificador BOE-A-…; o cuando haya que citar el texto
  vigente de una norma estatal o autonómica consolidada en el BOE. Lee el índice y los artículos con el binario
  kitlegal y responde citando identificador y bloque.
metadata:
  kitlegal-applets: boe
  kitlegal-referencias: normas
---
```

- Es la redacción de partida; la tarea puede ajustarla sin salirse de FR-001 (consulta y cita de normativa consolidada
  de cualquier materia, `BOE-A-…`, número y año, abreviaturas, sin restringirse a una materia) ni de las reglas de
  data-model §1.1 (≤ 1024 caracteres, sin `<` ni `>`).
- Toda norma nombrada por número y año o por abreviatura en `SKILL.md` está en `data/normas.yaml` (FR-020).

### 2.2 Orden de las secciones (FR-002)

1. `# Consultar y citar legislación consolidada del BOE` y un párrafo de alcance.
2. `## Protocolo` (§2.3).
3. `## Cómo se cita` (§3).
4. `## Comandos`: la región generada (contrato de sincronización §3).
5. `## Reglas` (§2.4).

Menos de 300 líneas en total (FR-002, FR-041).

### 2.3 Protocolo: pasos obligatorios

Cada paso es un encabezado o elemento numerado explícito, de modo que SC-004 se cuenta sin interpretar:

| Paso | Contenido obligatorio | FR |
|---|---|---|
| 1. **Identificar la norma** | antes de consultar nada, leer `references/normas.md` y localizar la norma por nombre, número y año o abreviatura | FR-004 |
| 2. **Resolver `BOE-A-…`** | si está en la referencia, tomar de ahí el identificador; si no, `scripts/boe buscar` y elegir por título y rango, diciendo cuál y, si hay varias plausibles, cuáles y por qué; igual para una norma autonómica consolidada en el BOE | FR-005 |
| 3. **Leer índice y bloques con `scripts/boe`** | `scripts/boe indice` si no se conoce el id de bloque; el id se copia de la entrada del índice cuyo `titulo` es el artículo («Artículo 118» de la Ley 9/2017 → `a1-30`) y nunca se compone del número, porque los ids de muchas normas no son `a<número>`; cada bloque con `scripts/boe articulo` o `scripts/boe articulos`: de uno en uno con `articulo`, y `articulos` solo cuando hacen falta varios bloques a la vez y todos salen del índice; si una orden con varios bloques termina con código 4 o 5, cada bloque por separado con `articulo` antes de dar ninguno por no consultado, porque el fallo de un bloque no impide leer los demás; seguir las remisiones necesarias leyendo el bloque remitido; `scripts/boe metadatos` y `scripts/boe analisis` cuando la pregunta dependa de la vigencia o de las modificaciones; nunca un id de bloque que no salga del índice o de la pregunta; ante un código 3, volver al índice (research D23, V64) | FR-007 |
| 4. **Evaluar si falta contexto** | remisiones, vigencia, modificaciones | FR-003 |
| 5. **Responder citando** | cada afirmación sobre el contenido con su cita (§3), del texto devuelto en la sesión; el rango de cada norma cuando se citan normas de rango distinto y que la ley prevalece sobre el reglamento que la desarrolla (**distinguir ley y reglamento**); **señalar variación autonómica** cuando lo preguntado puede variar por normativa autonómica (competencias compartidas o cedidas, desarrollo autonómico, régimen foral) y cuándo corresponde a normas locales que no están en la fuente | FR-008, FR-009, FR-012 |

Las órdenes del protocolo se escriben con `--json`, que es lo que da `fuente`, `url`, `fecha_consulta` y `hash` en la
salida. Los ejemplos del protocolo no dependen de que la materia sea fiscal y no nombran `buscar-materia`, `materias` ni
`sumario` (FR-003).

### 2.4 Reglas: explícitas y numeradas

| Regla | FR |
|---|---|
| No concluir que una norma o una regulación no existe por una búsqueda vacía o por su ausencia en `references/normas.md`: decir «no encontrada con esta búsqueda» y proponer reformular | FR-006 |
| Nunca inventar contenido legal: si `scripts/boe` falla (código 3, 4 o 5, o sin caché con `--offline`) o no está disponible, decir qué no se pudo consultar y no suplir el texto con conocimiento propio; si la orden que falló pedía varios bloques, decirlo solo después de haber pedido cada bloque por separado con `scripts/boe articulo` (paso 3; research D23) | FR-010 |
| Trasladar los avisos de vigencia del binario (derogada, vigencia agotada, consolidación no finalizada); no presentar como vigente el texto de una norma derogada; recordar que los textos consolidados del BOE tienen carácter informativo y no son asesoramiento | FR-011 |
| Nunca presentar, notificar, firmar ni tramitar nada en nombre de nadie, ni simularlo; si la pregunta lo pide, decir que es una acción que hace la persona y citar, si procede, la norma aplicable | FR-015 |
| Ningún caso especial para un municipio o una comunidad concretos; si se pregunta por un municipio, responder con la normativa estatal o autonómica consolidada y señalar que ordenanzas y normas locales no están en esta fuente | FR-012 |

### 2.5 Lo que `SKILL.md` y `references/` no contienen (FR-077, SC-012)

- Ninguna instrucción de usar siempre `--offline`, de fijar un directorio de caché o `KITLEGAL_CACHE_DIR`.
- Ninguna referencia a evals, al job, a GitHub Actions ni a modelos concretos.
- No cuentan como tales la tabla generada (que declara `--offline` porque `--describe` lo declara) ni la regla de FR-010
  sobre el fallo sin caché con `--offline`.

Lo vigila `TestSkillsDelRepositorio/sin-instrucciones-de-evals`: ninguna línea de `SKILL.md` fuera de la región generada
ni de `references/*.md` contiene, sin distinguir mayúsculas y como palabra completa (delimitada por lo que no es letra ni cifra Unicode, de
modo que «evalúa» no cuenta), `KITLEGAL_CACHE_DIR`, `eval`, `evals`, `job`,
`GitHub Actions`, `claude-[a-z]+-[0-9]`, `siempre[^.\n]*--offline`; la regla de FR-010 se redacta sin la palabra
«siempre» junto a `--offline`.

## 3. Formato de cita (FR-008)

```text
art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]
```

- La forma legible de la norma y del bloque va antes; la parte mecánica es exactamente
  `[<identificador>, bloque <id de bloque>]`, con el identificador y el id tal como los da la fuente.
- Expresión con la que se extrae: `\[(BOE-A-[0-9]{4}-[0-9]{1,9}), bloque ([A-Za-z0-9][A-Za-z0-9.-]{0,63})\]`. El
  corchete de cierre delimita el id, que puede terminar en punto (`a85bis.`, `ids.go`).
- Dentro de los corchetes no va nada más que el identificador y el id: ni «art. 21», ni «artículo 21», ni el nombre, el
  número o el rango de la norma, que van delante; tampoco el identificador sin el id.
- Una cita por bloque. Un bloque remitido se cita por separado.
- `SKILL.md` muestra este formato con la Ley 39/2015 (que está en `data/normas.yaml`), dice que no se admite otra
  forma para la parte entre corchetes y muestra `[Ley 39/2015, BOE-A-2015-10565, bloque a21]` como forma que no vale,
  con la que sí (research D23).

## 4. `references/normas.md`

Generado (contrato de normas y referencias §3). `SKILL.md` lo nombra en el paso 1 del protocolo.

## 5. `scripts/boe`

Enlace simbólico versionado a `../../../bin/instalado/kitlegal` (research.md D5). Tras `make install` resuelve al
binario instalado (contrato de instalación §3). Invocado como `scripts/boe <verbo> …` ejecuta el applet `boe`.
