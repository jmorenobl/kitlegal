# Contrato: la skill `jurisprudencia` v0

La redacción es [SKILL-jurisprudencia.prototipo.md](./SKILL-jurisprudencia.prototipo.md): el borrador del
prototipo, con su tabla de comandos generada (research V15, V19; M2). La tarea que entrega la skill lo copia a
`skills/jurisprudencia/SKILL.md` y regenera la tabla con `make skills-sync`; si la tabla regenerada difiere de la del
borrador, manda la regenerada.

## 1. Lo que `SKILL.md` lleva, y su requisito

| Parte del borrador | Qué dice | Requisito |
|---|---|---|
| Frontmatter | `name`, `description` con las preguntas que la activan —una sentencia nombrada, pedida, citada o resumida; si existe; jurisprudencia por materia; un documento traído— y `kitlegal-applets: cita`; sin `kitlegal-referencias` | FR-040 |
| Introducción | Que kitlegal no consulta el buscador, por qué, y qué hace en su lugar | FR-040, FR-041 |
| Protocolo, cabecera | Las dos formas de pedir cada operación, herramienta y orden; la herramienta, solo con los argumentos que se usan, porque uno vacío es un error (contracts/applet-cita.md §1); la orden, siempre con `--json` | FR-040, FR-049; contracts/evals-jurisprudencia.md §3 |
| Paso 1 | Qué es tener una sentencia delante: su documento, con su ficha; qué no lo es | FR-041 |
| Paso 2 | La línea `⚠ SENTENCIA NO COMPROBADA:` con su forma, una por referencia; cómo se prepara cada forma de referencia; sin fecha no se prepara; la consulta sale de lo devuelto; ni «existe» ni «no existe» | FR-042 |
| Paso 3 | Cotejar la ficha y nada más —sus líneas, de la de `Roj:` a la de `Tipo de Resolución:`—, en `documento` o por la entrada estándar; el documento que no es el pedido; `no-se-corresponden`; el texto sin ficha, y el cotejo que se repite si la ficha no se pasó entera | FR-020, FR-043, FR-048, FR-049 |
| Paso 4 | La forma fija de la cita, con los datos de la ficha; que sale del documento aportado; ningún ECLI que no venga de una operación o de la persona | FR-044 |
| Paso 5 | Con la ficha sola, la cita y sus datos; con el texto, lo que el texto dice; nunca un resumen de lo que no está | FR-045 |
| Paso 6 | La pregunta por materia: `cita preparar --texto`, la dirección, y ninguna sentencia de memoria | FR-046 |
| Paso 7 | El Tribunal Constitucional: no cubierto, su buscador, sin línea y sin `cita preparar` | FR-047 |
| Comandos | Los códigos, y la tabla generada con las dos operaciones como orden y como herramienta | FR-040; ADR 0035 |
| Reglas 1 a 6 | Las invariantes de toda skill; el error de argumentos; sin herramienta ni binario, ninguna cita | FR-049; spec, Edge Cases |

Lo que la skill dice con `no-se-corresponden` y que no resuma lo que no tiene delante no tienen control en este hito
(FR-065): los pasos 3 y 5 son reglas de la skill, y el informe final no las da por comprobadas.

## 2. Las formas fijas

```text
⚠ SENTENCIA NO COMPROBADA: <la referencia, como se dio>
STS 1088/2023, de 4 de julio [ECLI:ES:TS:2023:3144, ROJ: STS 3144/2023]
```

La primera, al principio de una línea. La segunda tiene delante del corchete las siglas del ROJ de la ficha, su
número de resolución y su fecha; lo que se compara sin modelo es el corchete (FR-051).

## 3. Uso, de fuera adentro

Lo que pide la skill por pregunta, lo que el modelo escribe para pedirlo, lo que recibe y lo que la respuesta
lleva. Los bytes recibidos son los medidos en research M1; los de la ficha, los de M3; lo que el modelo escribe
alrededor de ella es un cálculo.

| Pregunta | Operaciones | Lo que el modelo escribe al pedirlas | Lo que recibe | Lo que lleva la respuesta |
|---|---|---|---|---|
| Una sentencia que no está delante | 1 `cita preparar` por sentencia | La referencia: menos de 80 bytes, con la orden entera | 470 a 572 bytes por sentencia | Una línea de 60 a 150 bytes y la consulta: la dirección y de una a tres casillas |
| Cinco sentencias que no están delante | 5 | Cinco referencias | Unos 2,9 kB | Cinco líneas y cinco consultas |
| Un documento traído, sin haber pedido ninguno | 1 `cita cotejar` | La ficha: 316 bytes en la del fragmento, y menos de 60 de orden o de llamada alrededor | 559 bytes | La cita, de unos 70 caracteres |
| Un documento traído, que es el pedido | 1 `cita cotejar` | La ficha y la referencia: unos 400 bytes | 628 a 652 bytes | La cita |
| Un documento traído, que no es el pedido | 1 `cita cotejar` y 1 `cita preparar` | La ficha y, dos veces, la referencia: menos de 500 bytes | 869 a 1 005 bytes, y 470 a 572 | Qué difiere, la línea de la pedida y su consulta |
| Una pregunta por materia | 1 `cita preparar` con texto | Los términos de la búsqueda | 389 bytes con «cláusula suelo»; el texto va una vez tal cual y dos codificado | Una dirección |
| Una sentencia del Tribunal Constitucional | Ninguna | Nada | Nada | La declaración y una dirección |
| Un número sin fecha | Ninguna | Nada | Nada | La línea y la petición de la fecha |

Con la herramienta, el sobre va dos veces en el mensaje (H21): el doble de los bytes recibidos. `cita cotejar` no
devuelve el texto que se le pasa.

**Lo que el modelo escribe por cotejo no crece con el documento.** Con la herramienta, `documento` lo escribe el
modelo; con la orden, también, en la entrada de la orden. El paso 3 manda pasar la ficha y nada más: once líneas y
316 bytes en la del fragmento, las mismas etiquetas en cualquier documento del CENDOJ (research S5), mida lo que
mida la sentencia. Con el texto entero serían 2 353 bytes ya en el fragmento, que es un encabezamiento y un fallo; el
documento completo de esa sentencia es un PDF de 175 955 bytes cuyo texto no está en el repositorio y no se ha
medido. La ficha da el mismo cotejo que el texto entero (research V41).

**Ninguna salida crece con el uso.** Tras meses de uso diario, la consulta de la sentencia siguiente y el cotejo del
documento siguiente tienen esos mismos bytes, y la ficha que se pasa, los suyos: el applet no guarda nada.

**Cuándo deja de darse cada señal.**

- La línea `⚠ SENTENCIA NO COMPROBADA:`: cuando la persona trae el documento y se coteja como el pedido. No vuelve
  en otra respuesta salvo que se necesite de nuevo esa sentencia sin su documento.
- El hallazgo `documento-distinto`: en el cotejo del documento pedido. No se guarda en ningún sitio.
- `no-se-corresponden`: cuando la persona trae de nuevo el documento y su ficha se corresponde.
- La declaración de cobertura del Tribunal Constitucional: responde a su consulta y no es una señal que se repita;
  deja de darse con el candidato `tc` del backlog.
- `no-se-deduce` no llega a la respuesta.

`SKILL.md` lo lee el modelo una vez por conversación en que se activa: 195 líneas, 13,6 kB (research M2). Las dos
herramientas, una vez por conexión.

## 4. Lo que comprueba `make ci`

- `make skills-check` (`TestSkillsDelRepositorio`): frontmatter válido, menos de 300 líneas, tabla idéntica a la
  generada, sin `scripts/` (FR-040; SC-009). Su arnés deja de suponer referencias en toda skill (research D15).
- `TestOrdenesDeLasSkillsEmpotradas`: las órdenes que nombra la skill son verbos del registro.
- `TestEvalsDelRepositorio/linea-sin-consulta`: sigue exigiendo su línea a `boe-legislacion` y a `legal-core`, y no a
  `jurisprudencia` (research D16).
- Que `skills/boe-legislacion/` y `skills/legal-core/` no cambian lo comprueba el diff del hito (FR-049).

La activación y las formas las mide el job de cierre (SC-001; research S2).
