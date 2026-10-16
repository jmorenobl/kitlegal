# Contrato: `jurisprudencia` v0.1.1, el cotejo antes de nada

Qué tiene que decir `skills/jurisprudencia/SKILL.md` al acabar el hito, dónde, y qué no cambia (FR-020 a FR-024). La
redacción la escribe el run; este contrato fija su contenido y sus límites. La salida que la skill consume, la de
`cita cotejar`, no cambia (§4).

## 1. Lo que cambia

Solo dos sitios del «Protocolo»:

### 1.1 Paso 1, «Qué sentencias hacen falta, y cuáles están delante»

Su última viñeta (hoy, «La que no está delante va por el paso 2; la que sí, por el paso 3…», `SKILL.md:48`) dice que
con un documento en la conversación **lo primero** es el paso 3, antes de decir nada de él. Las del Tribunal
Constitucional siguen yendo por el paso 7.

### 1.2 Paso 3, «Con el documento delante: cotejar su ficha»

Al principio del paso, antes de cómo se pasa la ficha (hoy, `SKILL.md:83-89`), tres instrucciones:

1. **El cotejo es lo primero** que se hace con un documento que está en la conversación —texto pegado o PDF leído—:
   su ficha va a `cita_cotejar`, o a su orden, antes de escribir ninguna frase sobre él (FR-020).
2. **Nada del documento antes del cotejo**: ni que es la sentencia pedida, ni que coincide con ella, ni de qué trata
   (FR-021).
3. **Si es o no la pedida lo dice `es_la_pedida`**, lo que devuelve la operación, y nunca una comparación hecha a ojo
   entre el documento y la referencia: falla, por ejemplo, cuando el número de resolución del documento es el ROJ
   pedido —1088/2023 es el número de resolución del documento del ROJ STS 3144/2023, y el ROJ STS 1088/2023 es el de
   otra sentencia— (FR-022).

Lo que el paso 3 ya dice sigue igual, también en su sitio dentro del paso (FR-020): la ficha y nada más, desde la línea
de `Roj:` hasta la de `Tipo de Resolución:`; la referencia, si se pidió una, en la forma en que se pidió; la orden con
la ficha por la entrada estándar, o en `--documento`; el PDF, que se lee y del que se pasa la ficha; lo que devuelve;
qué hacer si el documento no es el pedido, con las palabras del hallazgo; `correspondencia`; y el código 2.

Si para decir lo de §1.2 hace falta mover una frase del paso 3 —por ejemplo, la que hoy dice que devuelve
`es_la_pedida`—, se mueve, sin cambiar lo que dice.

## 2. Lo que no cambia

| Parte de `SKILL.md` | Hoy | Cómo se comprueba |
|---|---|---|
| El frontmatter (`name`, `description`, `metadata`) | `SKILL.md:1-14` | quickstart §5: idéntico al de `main` (FR-023) |
| El título y la entrada | `:16-25` | quickstart §5 |
| La introducción del «Protocolo» (las dos formas de pedir las operaciones) | `:27-38` | quickstart §5 |
| El paso 2 | `:50-81` | quickstart §5 (FR-024) |
| Los pasos 4 a 7 | `:114-156` | quickstart §5 (FR-024) |
| «Comandos», con su tabla generada, y «Reglas» | `:158-197` | `make skills-check` (la tabla); quickstart §5 (el resto) (FR-023, FR-024) |

Ninguna frase de los pasos 2 y 4 a 7 ni de las reglas hace hablar del documento antes del cotejo: el paso 6 manda a
«los pasos 3 a 5», en su orden, y el paso 5 dice qué se dice de una sentencia cuya ficha ya se cotejó. No hay que
cambiar ninguna (FR-024).

## 3. La forma

- Menos de 300 líneas: 197 en `main`; con §1, 207. **Control**: `TestSkillsDelRepositorio` en `make skills-check`
  falla con 300 (FR-023).
- La tabla de comandos es la que regenera `make skills-sync`, sin cambios, y ninguna herramienta nueva: solo
  `cita_preparar` y `cita_cotejar`. **Control**: `TestSkillsDelRepositorio` y `TestOrdenesDeLasSkillsEmpotradas`, en
  `make skills-check` (FR-023).
- Ningún bloque de código nuevo con una orden que no esté ya en la skill: si §1.2 enseña una, es la de cotejar que ya
  está (`SKILL.md:91-96`).
- Sin campo de versión: la v0.1.1 la registra `CHANGELOG.md` (FR-024, FR-040).

## 4. La salida que consume la skill: `cita cotejar`, sin cambios

| | |
|---|---|
| **Quién la pide** | La skill, en el paso 3, una vez por documento delante y antes de decir nada de él: con un documento por pregunta, una invocación; con *n*, *n* |
| **Forma** | El sobre `{ok, fuente, url, fecha_consulta, hash, data}`, y en `data`: `ficha`, `correspondencia`, `pedida`, `es_la_pedida` (solo con una referencia pedida) y `hallazgos` |
| **Ejemplo** | `data` del caso de la 06, el golden `internal/app/testdata/cita/cotejar-roj-cruzado.json` (673 bytes): `{"ficha":{"roj":"STS 3144/2023","ecli":"ECLI:ES:TS:2023:3144","organo":"Tribunal Supremo. Sala de lo Civil","fecha":"2023-07-04","recurso":"4703/2019","resolucion":"1088/2023","ponente":"PEDRO JOSE VELA TORRES","tipo":"Sentencia"},"correspondencia":"se-corresponden","pedida":{"forma":"roj","valor":"STS 1088/2023"},"es_la_pedida":false,"hallazgos":[{"clase":"documento-distinto","difiere":[{"dato":"roj","pedido":"STS 1088/2023","documento":"STS 3144/2023"}],"cruce":"numero-de-resolucion","explicacion":"El documento no es el pedido: se pidió el ROJ STS 1088/2023 y el del documento es STS 3144/2023. 1088/2023 es el número de resolución del documento, no su ROJ."}]}` |
| **Tamaño** | El sobre entero de ese golden, 937 bytes (research V20); el spec midió 949 con `go run`, con su fecha y su huella. Lo que se pasa es la ficha, no el texto: el tamaño no depende del documento. Con *n* documentos, *n* × ~1 KB. No crece con el uso del kit, sino con los documentos de la pregunta |
| **Señal y cuándo se apaga** | `es_la_pedida: false` y su hallazgo: valen para ese documento en esa pregunta. Con un documento nuevo, un cotejo nuevo; nada se acumula de una pregunta a otra |

## 5. Cómo se comprueba

- La forma (§3), en `make ci` (`make skills-check`).
- Que lo de §1.2 está y que nada de §2 cambia: la revisión final, por lectura del diff (SC-007, FR-024), con la ayuda
  de quickstart §5.
- Su efecto: el job de cierre, con los dos `documento_sin_cotejar` a 0 de 12 (SC-001); y, fuera del run, el sondeo de
  la persona (SC-010).
