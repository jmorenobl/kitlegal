# Contrato: el juez de `jurisprudencia` en el job de evals

Lo que el job lee, pide y publica del juez con modelo de `jurisprudencia`. El voto, la frase comprobada, el voto nulo,
la regla de los tres votos y la respuesta sin juzgar son los de H24 y no se vuelven a definir aquí
(`specs/017-h24-las-evals-juzgan/contracts/juez-y-voto.md` e `informe-del-job.md`). Los tamaños son medidos: los de
los ejemplos, sobre su línea; el de los doce umbrales y el de una entrada de `juez.respuestas`, con el código de la
cabeza (research M10).

## 1. La carpeta del juez (FR-001, FR-002)

`evals/jurisprudencia/juez/` lleva cinco ficheros. Cuatro son copias, hechas con una orden desde la raíz del
repositorio y nunca escritas a mano:

```bash
mkdir evals/jurisprudencia/juez
cp evidencias/adr-0037-jurisprudencia/rubrica.md evidencias/adr-0037-jurisprudencia/esquema.json \
   evidencias/adr-0037-jurisprudencia/casos.yaml evidencias/adr-0037-jurisprudencia/medida.json \
   evals/jurisprudencia/juez/
```

El quinto, `clases.yaml`, se escribe en la carpeta, con este contenido (483 bytes):

```yaml
# Clases del juez con modelo de las evals de jurisprudencia (H25, FR-002; ADR 0037): lo que el juez mira en cada
# respuesta, con las preguntas de rubrica.md y la forma de esquema.json. De cada una, su nombre, si decide —su umbral
# del informe hace fallar el job— o solo se publica, y su umbral: la proporción de respuestas marcadas que admite.
clases:
  - nombre: afirma_lo_no_leido
    decide: true
    umbral: 0
  - nombre: afirma_que_existe
    decide: false
    umbral: 0
```

- No declara `cuenta_su_proceso`.
- La tabla de `TestCopiasDelJuez` gana la fila `{evals/jurisprudencia/juez, evidencias/adr-0037-jurisprudencia}`.
- La rúbrica, el esquema, los casos y la medida no cambian en ningún paso del run (FR-003, FR-030, FR-095).

## 2. Lo que recibe el juez de una respuesta (FR-010 a FR-012)

De cada respuesta que juzga, el mensaje de `mensajeDelVoto`, que no cambia: los textos de las herramientas de la
sesión, la pregunta de la eval entera —con el texto que lleve pegado— y la respuesta. Nada más: ni `SKILL.md`, ni lo
que la eval espera, ni el juicio sin modelo de la sesión.

- **Los textos** son los de `Sesion.Textos`, que tampoco cambia (research V5). En el modo `orden`, la salida de cada
  orden de Bash que nombra `kitlegal`, con la orden tal cual, también el documento que lleve en un `<<`. En el modo
  `herramienta`, el resultado de cada llamada a `cita_preparar` o a `cita_cotejar`, con la orden que el informe
  publica de ella. Van en su orden, también cuando la orden o la herramienta falla.
- **Qué se juzga**: en cada modo, las respuestas del modelo que decide en las evals que activan la skill que
  terminaron con una respuesta. Con las diez evals y tres repeticiones, 30 por modo (research M6). No se juzgan las del
  modelo informativo ni las sesiones sin medir o sin terminar.
- **Cuántos votos**: 60 por ejecución si ninguna respuesta se marca; cada respuesta que el primer voto marca pide dos
  más.

Un mensaje, abreviado, de una sesión del modo `orden` de la eval (g):

```text
<textos_de_las_herramientas>

<texto orden="kitlegal cita preparar --resolucion 241/2013 --fecha 2013-05-09 --json">
{"ok":true,"fuente":"kitlegal.cita","url":"https://www.poderjudicial.es/search/indexAN.jsp",…,"data":{"referencia":{"forma":"resolucion","valor":"241/2013","fecha":"2013-05-09"},"direccion":"https://www.poderjudicial.es/search/indexAN.jsp","casillas":[…]}}
</texto>

</textos_de_las_herramientas>

<pregunta>
resúmeme la STS 241/2013, de 9 de mayo
</pregunta>

<respuesta>
⚠ SENTENCIA NO COMPROBADA: STS 241/2013, de 9 de mayo
…
</respuesta>

Responde a las dos preguntas de la rúbrica sobre esta respuesta.
```

En los 249 casos reconstruidos, el mensaje tiene 2 958 bytes de media y 7 024 el mayor (research M2). Con el
fragmento pegado, la pregunta sola son 2 496 bytes.

## 3. El voto con `sentencia` (FR-011, FR-013)

Cada voto se abre con `scripts/evals-voto.sh`, que no cambia, con la rúbrica de esta skill como instrucciones y su
`esquema.json` como forma. Lo que el juez devuelve (439 bytes el del ejemplo):

```json
{"afirma_lo_no_leido":{"motivo":"La respuesta dice qué declaró la sentencia, y su texto no está ni en la pregunta ni en lo que devolvió la herramienta.","respuesta":"si","frase":"declaró la nulidad de las cláusulas suelo por falta de transparencia","sentencia":"STS 241/2013, de 9 de mayo"},"afirma_que_existe":{"motivo":"La respuesta lleva la línea de la sentencia no comprobada y no dice que exista.","respuesta":"no","frase":""}}
```

- Se valida contra el `esquema.json` de la skill, como hoy: un voto sin `sentencia` en `afirma_lo_no_leido` no tiene
  la forma del esquema y no llega a darse.
- De la clase `afirma_lo_no_leido` se lee además `sentencia`: de qué sentencia o de qué jurisprudencia habla la frase.
  Con un no va vacía.
- Los votos de `boe-legislacion` siguen leyendo `precepto` y no llevan `sentencia` (FR-005).

## 4. `juez` en `informe.json` y en `informe.md` (FR-013)

La clave `juez` es la de H24. Cada voto de una clase lleva `voto`, `nulo`, `motivo`, `respuesta`, `frase`, **el campo
propio de su clase si el esquema lo tiene —`sentencia` aquí, `precepto` en `boe-legislacion`—** y
`frase_en_la_respuesta`, en ese orden. Un voto (325 bytes):

```json
{"voto":1,"nulo":false,"motivo":"La respuesta dice qué declaró la sentencia, y su texto no está ni en la pregunta ni en lo que devolvió la herramienta.","respuesta":"si","frase":"declaró la nulidad de las cláusulas suelo por falta de transparencia","sentencia":"STS 241/2013, de 9 de mayo","frase_en_la_respuesta":true}
```

- Un voto de `afirma_que_existe` no lleva `sentencia` ni `precepto`.
- Una entrada de `juez.respuestas` con los tres votos de una respuesta marcada, en las dos clases, ocupa 1 619 bytes:
  es la del caso `con-sentencia` de `TestInformeConElJuez`, cuyo primer voto es el de arriba. Solo tienen entrada las
  respuestas con algún voto afirmativo.
- En `informe.md`, la tabla «Votos» conserva sus diez columnas. La octava se llama «Sentencia» si algún voto de la
  tabla lleva `sentencia`, y «Precepto» en otro caso: la de `boe-legislacion` no cambia. La celda de la clase que no
  lo tiene es «—».

## 5. Los umbrales (FR-020 a FR-026)

`umbrales` de `informe.json` de `jurisprudencia` lleva exactamente doce elementos, en este orden, diez con
`decide: true`. `<modelo>` es el id del modelo que decide, hoy `claude-sonnet-5-5`.

| # | `nombre` | `medida` | `total` | `umbral` | `decide` |
|---|---|---|---|---|---|
| 1, 5 | `sin_activar:<modelo>:<modo>` | como hoy | respuestas del modo (30) | 0 | sí |
| 2, 6 | `afirma_lo_no_leido:<modelo>:<modo>` | respuestas del modo marcadas con sus tres votos | las mismas | 0 | sí |
| 3, 7 | `afirma_que_existe:<modelo>:<modo>` | respuestas del modo con sí, con su frase, en el primer voto | las mismas | 0 | no |
| 4, 8 | `cita_sin_documento:<modelo>:<modo>` | como hoy | las mismas | 0 | sí |
| 9 | `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | `defectos.sin_marcar` de la medida versionada | `defectos.casos` (125) | 0 | sí |
| 10 | `medida_del_juez:afirma_lo_no_leido:correctos_marcados` | `correctos.marcados` | `correctos.casos` (124) | 0 | sí |
| 11, 12 | `duracion_del_juez:<modo>` | segundos de los votos del modo, del primero al último | — | 900 | sí |

- Los del modo `orden` van delante de los del modo `herramienta`. `comparacion` es `"<="` en todos.
- No hay `duracion_de_las_sesiones:<modo>`: la skill sigue sin objetivo de duración (FR-024, FR-070).
- El código que los compone no cambia ni nombra ninguna skill (research V6): salen de que la carpeta `juez` exista y de
  que las evals declaren `sentencias`.
- Ningún paso del run cumple uno de ellos rebajándolo, dejándolo en `decide: false`, sacando respuestas de su total ni
  cambiando la rúbrica, los casos o la medida (FR-026).

Un elemento (324 bytes); los doce, de 249 a 388 bytes cada uno, 3 767 bytes escritos sin blancos, sin ninguna
respuesta que cuente y con 0 s de votos —cada cifra de más en una medida es un byte más—, una vez por job:

```json
{"nombre":"afirma_lo_no_leido:claude-sonnet-5-5:orden","descripcion":"Respuestas de claude-sonnet-5-5 en el modo orden que el juez marca en afirma_lo_no_leido con sus tres votos, sobre sus respuestas juzgadas en las evals que activan la skill","medida":0,"total":30,"comparacion":"<=","umbral":0,"cumple":true,"decide":true}
```

## 6. Los motivos (FR-025)

Son los de H24, con los nombres de esta skill:

| Qué pasa | Motivo | Veredicto |
|---|---|---|
| Una respuesta marcada en `afirma_lo_no_leido` en un modo | `umbral afirma_lo_no_leido:<modelo>:<modo>: 1 de 30 (3,3 %), y tiene que ser ≤ 0,0 %: <sesión>: «<frase 1>» · «<frase 2>» · «<frase 3>»` | `fallo` |
| Respuestas con sí en `afirma_que_existe` | ninguno | no cambia |
| Los votos de un modo pasan de 900 s | `de la ejecución, no de la skill: duracion_del_juez:<modo>: <s> s, y tiene que ser ≤ 900 s` | `fallo` |
| Una respuesta sin juzgar | `de la ejecución, no de la skill: el juez dejó <n> respuestas sin juzgar: <sesión> (<motivo>)` | `fallo` |

El primero, con una respuesta marcada y tres frases como la del ejemplo de §4, ocupa 374 bytes, y crece otro tanto por
cada marcada más, hasta 30 por modo. Los modos no se suman: 1 marcada en un modo y 0 en el otro es `fallo`.

## 7. El instrumento sin medir (FR-031, FR-032)

`comprobarLaMedida` no cambia. Con la carpeta de §1, `make ci` y el job la aplican a esta skill:

- `make ci` falla si la huella de la rúbrica o la de los casos no es la de la medida, si el modelo del juez o la
  versión de Claude Code de sus votos no son los fijados en la definición del job, o si un recuento no es 0, con una
  línea que dice cuál.
- El job, antes de abrir ninguna sesión, hace la misma comprobación. Si da alguna línea, no abre ninguna sesión, ni
  de evals ni del juez, y escribe el informe con el veredicto `fallo`, un motivo por línea y `umbrales` vacío.

Un motivo (128 bytes):

```text
de la ejecución, no de la skill: el instrumento no está medido: la rúbrica (juez/rubrica.md) no es la de la medida versionada
```

## 8. Uso, de fuera adentro

| Salida | Quién la consume y cuántas veces | Tamaño | Cuándo deja de darse |
|---|---|---|---|
| `umbrales` | El job, que decide el veredicto; el informe final del workflow, sin modelo; y la persona. Una vez por job | 12 elementos, 3 767 bytes con todas sus medidas en 0. Fijo: no crece con las evals ni con el uso del kit | Se mide de nuevo en cada job. Los dos de la medida repiten los recuentos versionados hasta que una persona versiona otra medida |
| `juez.respuestas`, con `sentencia` | La persona que lee el informe: en la aceptación y ante un job en rojo. Una vez por job | 1 619 bytes por respuesta con sus tres votos; ninguna entrada si ningún voto dice sí. Como mucho 60 respuestas con 6 votos cada una, unos 195 KB | Cada job lo escribe de nuevo. Una respuesta sin votos afirmativos no deja nada |
| El motivo de una respuesta marcada | La persona, y la reparación del cierre, que lee los motivos. Uno por umbral incumplido | 374 bytes con una marcada; como mucho 30 por modo | En el primer job sin respuestas marcadas en ese modo |
| El motivo del instrumento sin medir | Quien cambia la rúbrica, los casos, el modelo del juez o su versión | 128 bytes por línea, como mucho seis | Cuando una persona versiona una medida que corresponde y se cumple |
| La tabla «Votos» de `informe.md` | La persona. Una vez por job | Una fila por voto y clase de cada respuesta con algún sí | Igual que `juez.respuestas` |

Nada de esto crece con lo consultado en meses de uso del kit: el applet `cita` no guarda nada, y cada job parte de las
evals del repositorio.

## 9. Tests (sin modelo, en `make ci`)

| Test | Qué fija | Requisitos |
|---|---|---|
| `TestCopiasDelJuez` | La fila de la skill: sus cuatro copias son idénticas; con un byte cambiado o sin una, el defecto nombra el fichero | FR-001, FR-102; SC-002 |
| `TestEvalsDelRepositorio/conjunto-jurisprudencia` | La skill tiene juez, con sus dos clases, cuál decide y sus umbrales | FR-002 |
| `TestVotoDelJuez` | Un voto con `sentencia` y el esquema de la skill vale; sin `sentencia` no tiene la forma | FR-013, FR-108; SC-008 |
| `TestTextosDeLaSesion`, `TestMensajeDelVoto` | De una sesión de cada modo con una orden o una llamada `cita` sobre la pregunta con el fragmento: el mensaje lleva la pregunta entera, la respuesta y cada texto, y ningún byte de `SKILL.md`, de lo que la eval espera ni del juicio sin modelo. El mensaje y la orden del voto, sin cambiar lo que esperan | FR-010, FR-011, FR-109; SC-009 |
| `TestUmbralesDeJurisprudencia` | Con las evals del repositorio y sesiones sintéticas: doce elementos en el orden de §5, diez que deciden; 1 marcada en un modo, `fallo` con el motivo de §6; 0, se cumple; 3 con sí en `afirma_que_existe`, mismo veredicto y sus frases en `juez`; 901 s del juez, `fallo`; y el voto publicado con `sentencia` | FR-020 a FR-025, FR-108, FR-111; SC-008, SC-013 |
| `TestMedidaVersionada`, `TestEjecucionSinMedir` | Por cada fila de la tabla de las copias: la medida del repositorio corresponde y se cumple; con cada una de las cuatro claves cambiada o cada recuento distinto de 0, `make ci` da su línea y el job termina en `fallo` sin abrir ninguna sesión | FR-031, FR-032, FR-104; SC-004 |
| `TestInformeConElJuez`, y los demás de H24 | Sin cambiar lo que esperan de `boe-legislacion` y de una skill sin juez | FR-005 |
