# Informe del hito H23 · skill `jurisprudencia`: ninguna sentencia citada sin el documento que trae la persona + `cita preparar` y `cita cotejar` (adelantado; ADR 0036)

Generado por el workflow `hito` el 2026-10-07T18:14:03Z, sobre `bf4befd` de `019-h23-skill-jurisprudencia-ninguna`.
Lo escribe scripts/workflow/informe.sh sin modelo, desde los artefactos de `specs/019-h23-skill-jurisprudencia-ninguna/`. Fusionar (squash-merge) es una decisión humana:
si algo de lo que sigue no es lo que se quería, se corrige la sección del hito en docs/ROADMAP.md y se relanza.

## 1. Estado

- **make ci local**: verde.
- **CI y evals remotos** sobre `bf4befd` (medición 1): verde; es el producto de la cabeza (lo posterior solo toca `gates/`).
- **Evals por skill**: boe-legislacion aprobado, jurisprudencia aprobado, legal-core aprobado (tasas en la sección 3).
- **Umbrales del job**: boe-legislacion: 12 umbrales, **2 sin cumplir o solo publicados** (`cuenta_su_proceso:claude-sonnet-5-5:orden`, `cuenta_su_proceso:claude-sonnet-5-5:herramienta`); jurisprudencia: 4 umbrales, todos cumplen y hacen fallar el job; legal-core: sin umbrales (sección 3).
- **Revisión final**: juez A aprobado, juez B aprobado, 2 rondas.
- **Cambios que ningún juez vio**: ninguno.
- **Tareas**: 12 hechas, 0 en cuarentena, 0 pendientes sin cuarentena.
- **Diff**: 19 commits; 194 files changed, 18028 insertions(+), 576 deletions(-).

## 2. Supuestos y pendientes

Decisiones que el run tomó sin preguntar, ordenadas por impacto: cada paso que escribe una etiqueta la suya (ADR 0028).

### Cambian el comportamiento visible, el alcance o una skill

**Comportamiento (salida, códigos, ficheros, argumentos)**

- corrector_plan: el plan daba por no dado un argumento escrito con valor vacío —`cita preparar ECLI:ES:TS:2023:3144 --texto ""` y `… --roj ""` terminaban con 0, y `cita cotejar --documento ""` leía la entrada estándar—, con una premisa sobre el analizador que era falsa → un argumento escrito con valor vacío está dado y es el error de su requisito, código 2 y clase `argumentos`, como orden y como llamada de herramienta: una referencia vacía no tiene su forma y cuenta como una forma dada (FR-006), … (entera en `specs/019-h23-skill-jurisprudencia-ninguna/gates/supuestos.md` o `clarify-respuestas.json`)
- corrector_plan: `data` de `cita preparar` llevaba `cobertura` en toda salida, con `cubierto` también para `--roj "STC 79/2024"`, para un número con su fecha y para un texto, de los que FR-014 no declara nada → lectura conservadora: `cobertura` solo está cuando declara algo fuera de cobertura —el ECLI de órgano `TC`, con `cendoj` `no-cubierto` y su `motivo`—, y `cubierto` deja de ser un valor; en las demás salidas la clave no está, porque el binario no sabe si el CENDOJ tiene lo que se busca. Nin… (entera en `specs/019-h23-skill-jurisprudencia-ninguna/gates/supuestos.md` o `clarify-respuestas.json`)
- plan: una llamada de herramienta no puede leer la entrada estándar del servidor, que es el protocolo, y FR-030 no deja tocar el applet `mcp` → la entrada estándar la entrega el kernel al verbo que la lee, solo en una orden; `cita_cotejar` sin `documento` es un error de argumentos. El kernel gana `Registro.LeerDe` y `Despacho.Entrada` (research D12).
- plan: la tabla de comandos escribía como argumentos de posición las banderas propias de un verbo, y los de `cita` son los primeros con banderas en la tabla de una skill → `--describe` gana en `entrada` la anotación `x-banderas`, solo en los verbos que tienen banderas propias; `schemas/instalacion.json` la gana en sus tres partes, porque los verbos de `skills` ya las tenían. No va en el esquema de entrada de las herramientas (research D14).
- T002: FR-012 dice que un ROJ de siglas `STS` da su ECLI, pero el número de un ROJ no tiene cota (FR-005) y el de un ECLI admite 25 caracteres: de `STS <26 cifras o más>/<año>` saldría un ECLI que el propio reconocimiento rechaza → de ese ROJ no se deduce nada: `ROJ.ECLI` da el valor cero y falso, y `cita preparar --roj` no llevará equivalente. Es la lectura que menos añade y la que pide FR-012 al vetar un identificador «compuesto a medias». Lo fijan `TestEquivalencia` («número que no cabe en un … (entera en `specs/019-h23-skill-jurisprudencia-ninguna/gates/supuestos.md` o `clarify-respuestas.json`)
- T002: FR-006 no dice qué distingue «un ECLI de otro país» de uno mal formado → es de otro país el que lleva en su segunda parte dos letras mayúsculas ASCII distintas de `ES` —también `EU`, la del Tribunal de Justicia de la Unión—, y su mensaje dice que no es español; con cualquier otra cosa en esa parte (minúsculas, otra longitud, vacía) el mensaje la contrasta con la forma. La clase y el código son los mismos en los dos casos; es como lo hace el material de lectura.
- T003: FR-006 y contracts/applet-cita.md §6 listan los errores de una referencia y no dicen cuál gana cuando hay más de uno → el orden es: más de una forma dada; una fecha sin número; la forma dada mal escrita; un número sin su fecha; y la fecha mal escrita. Así `--resolucion 1088` sin fecha nombra el número, que no tiene su forma, y no la fecha que falta, y un ECLI junto a `--fecha ""` es «una fecha sin su número», como pide research D3. La clase y el código son los mismos en todos; solo cambia … (entera en `specs/019-h23-skill-jurisprudencia-ninguna/gates/supuestos.md` o `clarify-respuestas.json`)
- T003: FR-021 dice que la ficha es reconocible con sus ocho datos y los cuatro que se comparan con su forma, y no qué dato nombra el error si fallan varios, ni qué pasa con una línea `Roj:` sin el separador ` - ` → el error nombra el primero que falta o que no tiene su forma en el orden de la ficha (ROJ, ECLI, órgano, fecha, recurso, resolución, ponente, tipo); una línea `Roj:` sin separador es una ficha sin su ECLI, y el mensaje añade lo que la línea lleva y su forma. Un dato cuya primera línea … (entera en `specs/019-h23-skill-jurisprudencia-ninguna/gates/supuestos.md` o `clarify-respuestas.json`)
- T003: el prototipo del plan admitía blancos entre la etiqueta y sus dos puntos (`Roj : …`), y contracts/applet-cita.md §5 pide «la etiqueta entera delante de los dos puntos» y ninguna regla más que las de FR-021 → la etiqueta va seguida de sus dos puntos, sin nada en medio; `Roj : …` no abre una ficha. Es la lectura que menos admite: con ella un documento así da «falta la ficha» y la skill no cita, que es el lado del que tiene que fallar. Los blancos de FR-015 («solo de blancos») y de FR-021 («a… (entera en `specs/019-h23-skill-jurisprudencia-ninguna/gates/supuestos.md` o `clarify-respuestas.json`)
- T003: FR-025 da el cruce `numero-del-roj` «si se pidió un número de resolución y es el `<número>/<año>` del ROJ del documento», sin decir qué pasa cuando ese número es también el número de resolución del documento y solo difiere la fecha → no hay cruce: solo lo hay cuando el número pedido difiere del del documento, porque el hallazgo diría que ese número «no es su número de resolución» y lo es. Lo mismo en el otro sentido: pedido por su ROJ, es el pedido aunque el número de ese ROJ sea también e… (entera en `specs/019-h23-skill-jurisprudencia-ninguna/gates/supuestos.md` o `clarify-respuestas.json`)
- T006: research D9 dice que todo error de argumentos del applet lo devuelve el dominio, y el de T003 —cuyas rutas esta tarea no declara— no decide tres filas de contracts/applet-cita.md §6, porque lo que hace falta para decidirlas solo lo tiene el applet: `preparar` sin referencia ni `--texto` (`NuevaReferencia` devuelve «no hay» y deja la decisión a quien llama), una referencia junto a `--texto`, y `cotejar` sin ningún texto (las dos vías de FR-020 son del applet) → las tres las da `cita.go` con… (entera en `specs/019-h23-skill-jurisprudencia-ninguna/gates/supuestos.md` o `clarify-respuestas.json`)
- T006: contracts §6 da una sola fila para «ningún texto» y §7 pide que la llamada devuelva el sobre de su orden → el mensaje es uno, el mismo por la orden y por la herramienta: nombra `--documento` y la entrada estándar de la orden, y no hay otro para la llamada. «Ningún texto» es el de cero bytes; uno solo de blancos es un texto sin ficha, con el error del dominio.
- clarify Q1 (criterio d, conservadora): Opción A. El equivalente se deduce solo en la pareja que el repositorio documenta con su ejemplo: un ECLI de órgano `TS` cuyo número final es solo de cifras y un ROJ de siglas `STS`, `ECLI:ES:TS:<año>:<n>` ↔ `STS <n>/<año>`, con el número y el año trasladados y comparados carácter a carácter, sin normalizar. `cita preparar`: de un ECLI así da su ROJ, y de un ROJ de siglas `STS`, su ECLI; con cualquier otro órgano, con otras siglas o con un número final que … (entera en `specs/019-h23-skill-jurisprudencia-ninguna/gates/supuestos.md` o `clarify-respuestas.json`)

**Alcance (lo que queda fuera o dentro del hito)**

- plan: que la respuesta no resuma ni caracterice una sentencia cuyo texto no tenía delante (FR-045) no lo decide ningún control de este hito → queda como regla de la skill (paso 5 de `SKILL.md`); lo lee una persona en el cierre (SC-002) y lo decide el juez de H25. El informe final no puede darlo por comprobado (FR-065).
- plan: lo que la respuesta hace cuando `cita cotejar` dice que el ROJ y el ECLI de la ficha no se corresponden (FR-048) no tiene eval ni control → queda como regla de la skill (paso 3): no cita el documento y pide traerlo de nuevo; `cita_sin_documento` da por leído el ECLI de toda ficha reconocible. Lo lee una persona en el cierre (FR-065).
- T011: la tarea pide que el README no prometa ninguna release, y no dice si declara que lo que hoy se instala no lleva la skill → lo declara como hecho y sin fecha, en «¿Y las sentencias?»: «La skill `jurisprudencia` y sus dos operaciones están en el repositorio; la v0.5.0 no las lleva». Por qué: el README manda instalar la versión publicada, y «Hay tres skills» se leería como que ya vienen en ella; comprobado en esta sesión con `git ls-tree -d v0.5.0 skills/`, que da `boe-legislacion` y `legal-c… (entera en `specs/019-h23-skill-jurisprudencia-ninguna/gates/supuestos.md` o `clarify-respuestas.json`)
- T011: la skill da para el Tribunal Constitucional una dirección que no se ha podido comprobar sin red (research S1) → el README no la escribe: dice que el agente remite al buscador del propio tribunal. La dirección queda solo en el paso 7 de `SKILL.md` y en la eval 05, donde ya estaba anotada como supuesto.
- T011: el README decía que las condiciones del CENDOJ «prohíben expresamente consultarlo de forma masiva o automatizada», y el ADR 0036 dice que lo segundo no está en el aviso legal → «¿Y las sentencias?» da como razón la del ADR —el buscador responde con un CAPTCHA al programa que se identifica, y no se sortea— y, del aviso, solo lo que dice su texto, citado en `docs/JURISPRUDENCIA.md` §2: uso particular y prohibición de la descarga masiva.

**Skill (lo que pide, dice o comprueba una skill)**

- corrector_spec: FR-042 mandaba la línea `⚠ SENTENCIA NO COMPROBADA:` y la consulta de `cita preparar` para toda sentencia que no está en la conversación, y FR-047, declarar no cubierta la del Tribunal Constitucional, sin decir cuál rige → lectura conservadora: la sentencia del Tribunal Constitucional se declara no cubierta, con la dirección de su buscador y sin cita, no lleva la línea ni una consulta del CENDOJ, y la skill no pide `cita preparar` para ella, tampoco con su ECLI (FR-047, con la ex… (entera en `specs/019-h23-skill-jurisprudencia-ninguna/gates/supuestos.md` o `clarify-respuestas.json`)
- corrector_spec: FR-061 no decía dónde acaba un ECLI dentro de un texto, y una respuesta correcta que cierra una frase en un ECLI («…ECLI:ES:TS:2023:3144.») podía contar en `cita_sin_documento` → el número de un ECLI llega hasta el primer carácter que no es letra ASCII, cifra o punto, y los puntos en que termine no son suyos; la misma regla en la respuesta, en la pregunta y en la salida de una operación, con su caso en FR-086. No rebaja el umbral ni cambia lo que cuenta como salida de una operaci… (entera en `specs/019-h23-skill-jurisprudencia-ninguna/gates/supuestos.md` o `clarify-respuestas.json`)
- corrector_plan: el borrador de `SKILL.md` mandaba pasar a `cita cotejar` el texto entero del documento, que con la herramienta escribe el modelo y crece con la sentencia y no con la pregunta → decisión de uso: el paso 3 manda pasar la ficha y nada más, de la línea `Roj:` a la de `Tipo de Resolución:` —316 bytes en la del fragmento, frente a los 2 353 del fragmento—, como dicen la entrada del hito y Edge Cases del spec; si el error nombra un dato que sí está en el documento, la skill repite el co… (entera en `specs/019-h23-skill-jurisprudencia-ninguna/gates/supuestos.md` o `clarify-respuestas.json`)
- plan: la dirección del buscador del Tribunal Constitucional que da la skill y espera la eval 05, `https://hj.tribunalconstitucional.es/`, no está en el repositorio y no se ha podido comprobar sin red → supuesto no verificado (research S1): la persona la comprueba al leer el informe final; cambiarla son el paso 7 de `SKILL.md` y la eval 05.
- plan: `jurisprudencia` no tiene `references/` ni lleva la línea `⚠ SIN CONSULTA AL BOE:` → el arnés de `TestSkillsDelRepositorio` deja de suponer referencias en toda skill, y `linea-sin-consulta` comprueba las dos skills que llevan esa regla y no todas las de `skills/` (research D15, D16).
- plan: para `cita_sin_documento`, la salida de una operación es cada línea que es un sobre de kitlegal en la salida de una orden o de una llamada; una orden sin `--json` no da ninguno, y lo que escriba no cuenta (FR-061: lo que la sesión lee de otro sitio no es la salida de una operación) → la skill pide las órdenes siempre con `--json`. El fallo va del lado que cuenta de más, nunca de menos (research D19).
- T007: FR-061 dice que un ECLI «se compara sin distinguir mayúsculas de minúsculas» y no dice si se reconoce el escrito en minúsculas, `ecli:es:ts:2023:3144` → se reconoce entero sin distinguirlas, también la palabra `ECLI`: si no, el mismo ECLI en minúsculas no sería un ECLI y la comparación sin distinguirlas no tendría a qué aplicarse. Consecuencia en el juicio de T007: `[ecli:es:ts:2023:3144, ROJ: STS 3144/2023]` es una cita —cuenta donde `ninguna_cita`— y no es la cita esperada, que se compar… (entera en `specs/019-h23-skill-jurisprudencia-ninguna/gates/supuestos.md` o `clarify-respuestas.json`)
- T007: contracts §2 da para `direccion_de_busqueda` «la `direccion` que devolvió en la sesión un `cita preparar` con texto» y un motivo con una sola dirección, y no dice qué pasa si la sesión hizo varias búsquedas → vale la de cualquiera de ellas, que es lo que dice «un»; y si la respuesta no lleva ninguna, hay un solo motivo que las nombra todas, sin repetir y separadas por « o »: `falta la dirección de búsqueda <d1> o <d2>`. Con una sola búsqueda, que es lo que pide el protocolo, el motivo es e… (entera en `specs/019-h23-skill-jurisprudencia-ninguna/gates/supuestos.md` o `clarify-respuestas.json`)
- T007: contracts §2 dice que el `--roj` de la invocación «vale eso» y no qué vale una orden que lo escribe dos veces → el último, que es lo que vale para el binario (Kong v1.16.1 no rechaza una bandera repetida y aplica cada una en su orden; leído en esta sesión en `context.go`). Lo fija el caso `roj-escrito-dos-veces` de `TestJuzgarSentencias`.
- T007: contracts §5 describe los seis ficheros como «ninguno informativo» y su párrafo de reglas no lo enumera → `ReglasDeJurisprudencia` son las ocho que ese párrafo enumera —tamaño, activación, sentencias y las cinco clases, cada una con su número exacto de evals— y ninguna mira `informativa`: una regla que el contrato no da no se añade. Las seis evals del repositorio no la llevan.
- T008: contracts/evals-jurisprudencia.md §4 da el motivo de `cita_sin_documento` con una sesión y un ECLI, y dice que las sesiones van separadas por `; `, pero no qué separa dos ECLI de una misma sesión → ` · `, lo que separa las frases de una respuesta en el motivo de una clase del juez, que es con lo que research D21 lo compara: `<sesión>: <ECLI> (sin origen) · <ECLI> (sin origen); <sesión>: <ECLI> (cita sin documento cotejado)`. Lo fija el caso `dos-en-el-modo-herramienta` de `TestUmbralesDeJu… (entera en `specs/019-h23-skill-jurisprudencia-ninguna/gates/supuestos.md` o `clarify-respuestas.json`)
- T008: FR-060 no dice si el ECLI que cumple las dos condiciones —está en una cita sin documento cotejado y además suelto y sin origen— se nombra una vez o dos, ni en qué orden van los de una respuesta → una vez por condición, primero los de las citas y detrás los sueltos, cada grupo en el orden en que aparecen. La medida no cambia: la respuesta cuenta una vez. Lo fija el caso `las-dos-condiciones-sin-repetir` de `TestCitaSinDocumento`.
- T008: FR-060 deja fuera de la segunda condición el ECLI escrito en una línea que empieza por `⚠` y no dice lo mismo de la primera → la cita cuenta en cualquier línea, también en una de aviso, como ya la ve `ninguna_cita` (T007): la excepción de la línea es solo para el ECLI suelto, que es lo que el texto dice. Lo fija el caso `cita-en-una-linea-de-aviso`.
- clarify Q2 (criterio c): Opción A. Cuando `cita cotejar` dice que el documento no es el pedido, la sentencia pedida sigue sin estar en la conversación, y el paso 2 del protocolo rige sin excepción: la respuesta dice que el documento no es el pedido, con lo que difiere (FR-043), y lleva la línea `⚠ SENTENCIA NO COMPROBADA:` con la referencia pedida, como se dio, y la consulta que `cita preparar` da para esa referencia, en la forma en que se pidió (FR-042). La eval (f) añade a lo ya fijado en FR-0… (entera en `specs/019-h23-skill-jurisprudencia-ninguna/gates/supuestos.md` o `clarify-respuestas.json`)
- clarify Q3 (criterio d, conservadora): Opción A. No lo cita, ni con la forma fija ni de otra manera. La respuesta dice que el ROJ y el ECLI de la ficha no se corresponden, da los dos tal como están en la ficha y pide a la persona que vuelva a descargar el documento del buscador y lo traiga. No lleva por esto la línea `⚠ SENTENCIA NO COMPROBADA:`: el hito la define para la sentencia que no está en la conversación, con la referencia como se dio, y aquí hay un documento y puede no haberse dado ning… (entera en `specs/019-h23-skill-jurisprudencia-ninguna/gates/supuestos.md` o `clarify-respuestas.json`)
- clarify Q4 (criterio c): Opción A. Cita el documento con la forma fija, con los datos de la ficha, y no dice nada de la correspondencia: lo mismo que hace cuando se corresponden. «No se deduce» no es un hallazgo ni un fallo (FR-023): dice hasta dónde llega la regla del binario, no algo del documento, y la persona no puede hacer nada con ello. La respuesta ya dice que la cita sale del documento aportado (FR-044). FR-048 queda así: la skill lee la correspondencia en cada cotejo; con «se correspond… (entera en `specs/019-h23-skill-jurisprudencia-ninguna/gates/supuestos.md` o `clarify-respuestas.json`)

### Del propio run


- Observaciones de los jueces, por debajo del umbral y sin corregir: 16 en `spec-r2.json`, 13 en `plan-r2.json`, 12 en `tasks-r2.json`, 7 en `revision-a-r2.json`, 4 en `revision-b-r2.json`.

### Internos (29)

Decisiones que no cambian nada observable (técnica, estructura, tests), por autor: T001 (3), T003 (2), T004 (4), T005 (1), T006 (2), T007 (1), T008 (1), T009 (3), T010 (2), T011 (3), barrido (1), corrector_revision (1), corrector_tasks (1), plan (2), tasks (2). Enteras en `specs/019-h23-skill-jurisprudencia-ninguna/gates/supuestos.md`.

## 3. Evals sobre la cabeza

**boe-legislacion**: aprobado sobre `bf4befd`; decide `claude-sonnet-5-5` con 2 de 3; 21 evals, 0 nuevas, 8 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-lpac-articulo-21.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `02-lcsp-contrato-menor.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `03-lrbrl-atribuciones-del-pleno.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `04-lgt-prescripcion.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |
| `05-trlrhl-impuestos-municipales.yaml` | 3/3 | 3/3 | 3/3 | 1/3 |  |
| `06-irpf-rendimientos-del-trabajo.yaml` | 3/3 | 2/3 | 3/3 | 1/3 |  |
| `07-lrjsp-principio-de-legalidad.yaml` | 3/3 | 3/3 | 3/3 | 2/3 |  |
| `08-ltaibg-plazo-de-resolucion.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `09-constitucion-articulo-140.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `10-et-vacaciones.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `11-no-activa-programacion.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `12-no-activa-acuerdo-entre-amigos.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `13-lrbrl-atribuciones-por-materia.yaml` | 3/3 | — | 2/3 | — | informativa |
| `14-trlrhl-impuestos-por-materia.yaml` | 3/3 | — | 2/3 | — | informativa |
| `15-irpf-rendimientos-por-materia.yaml` | 3/3 | — | 3/3 | — | informativa |
| `16-lrjsp-legalidad-por-materia.yaml` | 3/3 | — | 3/3 | — | informativa |
| `17-ltaibg-plazo-por-materia.yaml` | 3/3 | — | 3/3 | — | informativa |
| `18-lrjpac-norma-derogada.yaml` | 3/3 | — | 2/3 | — | informativa |
| `19-lcsp-contrato-menor-redaccion-cambiada.yaml` | 3/3 | — | 3/3 | — | informativa |
| `20-lcsp-dos-bloques-redaccion-cambiada.yaml` | 3/3 | — | 2/3 | — | informativa |
| `21-sin-binario-ni-servidor.yaml` | 3/3 | 3/3 | — | — |  |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 54 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:orden` | 0 de 54 (0,0 %) | ≤ 0,0 % | sí | sí |
| `cuenta_su_proceso:claude-sonnet-5-5:orden` | 1 de 54 (1,9 %) | ≤ 0,0 % | ✗ **NO** | no: solo se publica, no es un control |
| `sin_activar:claude-sonnet-5-5:herramienta` | 0 de 50 (0,0 %) | ≤ 0,0 % | sí | sí |
| `afirma_lo_no_leido:claude-sonnet-5-5:herramienta` | 0 de 50 (0,0 %) | ≤ 0,0 % | sí | sí |
| `cuenta_su_proceso:claude-sonnet-5-5:herramienta` | 1 de 50 (2,0 %) | ≤ 0,0 % | ✗ **NO** | no: solo se publica, no es un control |
| `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | 0 de 212 (0,0 %) | ≤ 0,0 % | sí | sí |
| `medida_del_juez:afirma_lo_no_leido:correctos_marcados` | 0 de 47 (0,0 %) | ≤ 0,0 % | sí | sí |
| `duracion_de_las_sesiones:orden` | 452 | ≤ 900 | sí | sí |
| `duracion_de_las_sesiones:herramienta` | 542 | ≤ 900 | sí | sí |
| `duracion_del_juez:orden` | 98 | ≤ 900 | sí | sí |
| `duracion_del_juez:herramienta` | 99 | ≤ 900 | sí | sí |

**jurisprudencia**: aprobado sobre `bf4befd`; decide `claude-sonnet-5-5` con 2 de 3; 6 evals, 6 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-existe-con-numero-y-fecha.yaml` | 3/3 | 3/3 | 3/3 | 2/3 | **nueva** |
| `02-resumen-sin-documento.yaml` | 3/3 | 2/3 | 3/3 | 3/3 | **nueva** |
| `03-por-materia.yaml` | 3/3 | 2/3 | 3/3 | 3/3 | **nueva** |
| `04-documento-pegado.yaml` | 3/3 | 3/3 | 2/3 | 3/3 | **nueva** |
| `05-tribunal-constitucional.yaml` | 3/3 | 3/3 | 3/3 | 3/3 | **nueva** |
| `06-documento-que-no-es-el-pedido.yaml` | 3/3 | 0/3 | 3/3 | 0/3 | **nueva** |

Umbrales que publica el job (ADR 0029):

| Umbral | Medida | Condición | Cumple | Hace fallar el job |
|---|---|---|---|---|
| `sin_activar:claude-sonnet-5-5:orden` | 0 de 18 (0,0 %) | ≤ 0,0 % | sí | sí |
| `cita_sin_documento:claude-sonnet-5-5:orden` | 0 de 18 (0,0 %) | ≤ 0,0 % | sí | sí |
| `sin_activar:claude-sonnet-5-5:herramienta` | 0 de 17 (0,0 %) | ≤ 0,0 % | sí | sí |
| `cita_sin_documento:claude-sonnet-5-5:herramienta` | 0 de 17 (0,0 %) | ≤ 0,0 % | sí | sí |

**legal-core**: aprobado sobre `bf4befd`; decide `claude-sonnet-5-5` con 2 de 3; 4 evals, 0 nuevas, 0 informativas.

| Eval | claude-sonnet-5-5 | claude-haiku-4-5-20251001 | claude-sonnet-5-5 (herramienta) | claude-haiku-4-5-20251001 (herramienta) | Marca |
|---|---|---|---|---|---|
| `01-territorio-municipio-cubierto.yaml` | 3/3 | 1/3 | 3/3 | 0/3 |  |
| `02-territorio-municipio-no-cubierto.yaml` | 3/3 | 0/3 | 3/3 | 0/3 |  |
| `03-no-activa-receta-de-cocina.yaml` | 3/3 | 3/3 | 3/3 | 3/3 |  |
| `04-sin-binario-ni-servidor.yaml` | 3/3 | 0/3 | — | — |  |

Umbrales: el job no publica ninguno para esta skill.

Cada celda: sesiones que pasan de las abiertas; ✗, una serie que decide y no llega al umbral. Una eval informativa publica su tasa sin decidir el veredicto (ADR 0016). La regla por serie no hace cumplir un umbral agregado sobre todas las respuestas: eso solo lo hace un umbral del job que lo hace fallar (ADR 0029).

## 4. Revisión que la constitución reserva a la persona

- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/argumentos.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h19-skills-ambito-dir.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h19-skills-ambito-global.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h19-skills-aviso-sin-aviso.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h19-skills-aviso.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h19-skills-conflictos-dentro.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h19-skills-conflictos-entradas.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h19-skills-conflictos-rutas.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h19-skills-doctor-copia.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h19-skills-doctor-hallazgos.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h19-skills-dry-run.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h19-skills-idempotencia.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h19-skills-install-hosts.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h19-skills-install-local.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h19-skills-list-doctor.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h19-skills-no-empotrada.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h21-mcp-herramientas.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h21-mcp-proceso.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/h21-mcp-protocolo.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/skills-host-antigravity.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/app/testdata/script/skills-salida-legible.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/skills/testdata/script/instalar-sin-gobin.txtar`.
- Fixture o esquema EXISTENTE modificado: `internal/skills/testdata/script/instalar.txtar`.
- Fixture o esquema EXISTENTE modificado: `schemas/eval.yaml.json`.
- Fixture o esquema EXISTENTE modificado: `schemas/instalacion.json`.
- Fixtures nuevos (45), por directorio:
  - `internal/app/testdata/cita/`: `cotejar-ecli.json`, `cotejar-otra-fecha.json`, `cotejar-resolucion-cruzada.json`, `cotejar-resolucion.json`, `cotejar-roj-cruzado.json`, `cotejar-roj.json`, `cotejar-sin-referencia.json`, `preparar-ecli.json`, `preparar-resolucion.json`, `preparar-roj.json`, `preparar-texto.json`
  - `internal/core/cita/testdata/fuzz/FuzzLeerFicha/`: `blancos-en-los-extremos`, `cadena-vacia`, `dos-fichas`, `fecha-imposible`, `ficha-sola`, `finales-crlf`, `fragmento`, `otro-organo`, `sin-ficha`, `sin-ponente`, `texto-delante`
  - `internal/core/ids/testdata/fuzz/FuzzECLI/`: `blanco-delante`, `blanco-detras`, `cadena-vacia`, `mal-formado`, `minusculas`, `numero-con-punto`, `otro-organo`, `otro-pais`, `sentencia-conocida`, `termina-en-letra`, `tribunal-constitucional`
  - `internal/core/ids/testdata/fuzz/FuzzROJ/`: `anio-de-dos-cifras`, `blanco-detras`, `cadena-vacia`, `dos-espacios`, `minusculas`, `numero-con-letras`, `numero-de-26-cifras`, `otras-siglas`, `prefijo-roj`, `sentencia-conocida`, `siglas-de-dos-palabras`, `sin-anio`
- Esquemas nuevos: `schemas/cita.json`.
- Suite de aceptación activada y congelada (4 guiones, escritos desde el spec en la primera tarea): `h23-cita-cotejar.txtar`, `h23-cita-herramientas.txtar`, `h23-cita-preparar.txtar`, `h23-cita-sin-efectos.txtar`.

## 5. Tareas en cuarentena

Ninguna.

## 6. Trazabilidad

**Estado**: «tareas hechas» solo dice que las tareas que citan el requisito están marcadas; nada del run lo ha medido. «comprobado por su control» exige su fila en «Controles de umbral» de `plan.md` y que el control esté: un test o una comprobación de `make ci` que existe en la cabeza, con `make ci` en verde, o un umbral que el job de evals publica, cumple y hace fallar el job (ADR 0029). «UMBRAL NO CUMPLIDO» y «CONTROL SIN VERIFICAR» son lo que hay que mirar.

| Requisito | Tareas | Aceptación | Estado | Control de umbral |
|---|---|---|---|---|
| FR-001 | T001(hecha) T006(hecha) | cita-cotejar.txtar cita-herramientas.txtar cita-preparar.txtar cita-sin-efectos.txtar  | tareas hechas | — |
| FR-002 | T001(hecha) T003(hecha) T006(hecha) | — | comprobado por su control | `ci:internal/arch_test.go:TestArquitectura`, en make ci (verde) |
| FR-003 | T001(hecha) T006(hecha) | cita-sin-efectos.txtar  | comprobado por su control | `ci:internal/app/testdata/script/h23-cita-sin-efectos.txtar`, en make ci (verde) |
| FR-004 | T001(hecha) T006(hecha) | cita-cotejar.txtar cita-preparar.txtar  | tareas hechas | — |
| FR-005 | T001(hecha) T002(hecha) T003(hecha) | cita-preparar.txtar  | tareas hechas | — |
| FR-006 | T001(hecha) T002(hecha) T003(hecha) T006(hecha) | cita-preparar.txtar  | tareas hechas | — |
| FR-010 | T001(hecha) T003(hecha) T006(hecha) | cita-preparar.txtar  | tareas hechas | — |
| FR-011 | T001(hecha) T003(hecha) T006(hecha) | cita-preparar.txtar  | tareas hechas | — |
| FR-012 | T001(hecha) T002(hecha) T003(hecha) T006(hecha) | cita-preparar.txtar  | tareas hechas | — |
| FR-013 | T001(hecha) T003(hecha) T006(hecha) | cita-preparar.txtar  | tareas hechas | — |
| FR-014 | T001(hecha) T002(hecha) T003(hecha) T006(hecha) | cita-preparar.txtar  | tareas hechas | — |
| FR-015 | T001(hecha) T003(hecha) T006(hecha) | cita-preparar.txtar  | tareas hechas | — |
| FR-020 | T001(hecha) T004(hecha) T006(hecha) | cita-cotejar.txtar cita-herramientas.txtar  | tareas hechas | — |
| FR-021 | T001(hecha) T003(hecha) | cita-cotejar.txtar  | tareas hechas | — |
| FR-022 | T001(hecha) T003(hecha) T006(hecha) | cita-cotejar.txtar  | tareas hechas | — |
| FR-023 | T001(hecha) T002(hecha) T003(hecha) | cita-cotejar.txtar  | tareas hechas | — |
| FR-024 | T001(hecha) T003(hecha) T006(hecha) | cita-cotejar.txtar  | tareas hechas | — |
| FR-025 | T001(hecha) T002(hecha) T003(hecha) T006(hecha) | cita-cotejar.txtar cita-herramientas.txtar  | tareas hechas | — |
| FR-026 | T001(hecha) T003(hecha) T004(hecha) T006(hecha) | cita-cotejar.txtar cita-herramientas.txtar  | tareas hechas | — |
| FR-030 | T001(hecha) T004(hecha) T005(hecha) T006(hecha) | cita-herramientas.txtar  | comprobado por su control | `ci:internal/app/herramientas_test.go:TestHerramientasDelServidor`, en make ci (verde) |
| FR-031 | T006(hecha) T009(hecha) T012(hecha) | — | tareas hechas | — |
| FR-040 | T005(hecha) T009(hecha) | — | comprobado por su control | `ci:Makefile:skills-check`, en make ci (verde); `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde) |
| FR-041 | T009(hecha) | — | tareas hechas | — |
| FR-042 | T009(hecha) | — | tareas hechas | — |
| FR-043 | T009(hecha) | — | tareas hechas | — |
| FR-044 | T009(hecha) | — | tareas hechas | — |
| FR-045 | T009(hecha) | — | tareas hechas | — |
| FR-046 | T009(hecha) | — | tareas hechas | — |
| FR-047 | T009(hecha) | — | tareas hechas | — |
| FR-048 | T009(hecha) | — | tareas hechas | — |
| FR-049 | T005(hecha) T009(hecha) T012(hecha) | — | tareas hechas | — |
| FR-050 | T007(hecha) T010(hecha) | — | tareas hechas | — |
| FR-051 | T007(hecha) | — | tareas hechas | — |
| FR-052 | T007(hecha) | — | tareas hechas | — |
| FR-053 | T007(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestPreguntasConElFragmento`, en make ci (verde) |
| FR-054 | T007(hecha) T010(hecha) | — | tareas hechas | — |
| FR-055 | T007(hecha) T008(hecha) T012(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia`, en make ci (verde); `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, en make ci (verde) |
| FR-056 | T007(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestEvalsDelRepositorio`, en make ci (verde) |
| FR-060 | T008(hecha) T012(hecha) | — | comprobado por su control | `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden`: 0 de 18 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta`: 0 de 17 (0,0 %), ≤ 0,0 % |
| FR-061 | T007(hecha) T008(hecha) | — | comprobado por su control | `ci:internal/evals/sentencias_test.go:TestCitaSinDocumento`, en make ci (verde); `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia`, en make ci (verde) |
| FR-062 | T008(hecha) T012(hecha) | — | comprobado por su control | `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden`: 0 de 18 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 17 (0,0 %), ≤ 0,0 % |
| FR-063 | T008(hecha) T010(hecha) | — | comprobado por su control | `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia`, en make ci (verde); `ci:internal/evals/definicion_test.go:TestDefinicionDelJob`, en make ci (verde) |
| FR-064 | T008(hecha) T012(hecha) | — | tareas hechas | — |
| FR-065 | T012(hecha) | — | tareas hechas | — |
| FR-066 | — | — | SIN TAREA | — |
| FR-070 | T011(hecha) | — | tareas hechas | — |
| FR-080 | T006(hecha) | cita-preparar.txtar  | comprobado por su control | `ci:internal/app/cita_test.go:TestCitaPreparar`, en make ci (verde); `ci:internal/app/cita_test.go:TestCitaCotejar`, en make ci (verde) |
| FR-081 | T003(hecha) T006(hecha) | cita-cotejar.txtar  | comprobado por su control | `ci:internal/app/cita_test.go:TestCitaPreparar`, en make ci (verde); `ci:internal/app/cita_test.go:TestCitaCotejar`, en make ci (verde) |
| FR-082 | T001(hecha) T006(hecha) | cita-cotejar.txtar cita-preparar.txtar  | comprobado por su control | `ci:internal/app/testdata/script/h23-cita-preparar.txtar`, en make ci (verde); `ci:internal/app/testdata/script/h23-cita-cotejar.txtar`, en make ci (verde) |
| FR-083 | T002(hecha) T003(hecha) | — | tareas hechas | — |
| FR-084 | T006(hecha) | — | comprobado por su control | `ci:internal/arch_test.go:TestArquitectura`, en make ci (verde) |
| FR-085 | T001(hecha) T006(hecha) | cita-sin-efectos.txtar  | comprobado por su control | `ci:internal/app/testdata/script/h23-cita-sin-efectos.txtar`, en make ci (verde) |
| FR-086 | T008(hecha) | — | comprobado por su control | `ci:internal/evals/sentencias_test.go:TestCitaSinDocumento`, en make ci (verde); `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia`, en make ci (verde) |
| FR-087 | T007(hecha) | — | tareas hechas | — |
| FR-088 | T005(hecha) T009(hecha) T011(hecha) T012(hecha) | — | tareas hechas | — |
| SC-001 | T010(hecha) | — | comprobado por su control | `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:orden`: 0 de 18 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:cita_sin_documento:claude-sonnet-5-5:herramienta`: 0 de 17 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:orden`: 0 de 18 (0,0 %), ≤ 0,0 %; `evals:jurisprudencia:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 17 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5:orden`: 0 de 54 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:afirma_lo_no_leido:claude-sonnet-5-5:herramienta`: 0 de 50 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:orden`: 0 de 54 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:sin_activar:claude-sonnet-5-5:herramienta`: 0 de 50 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar`: 0 de 212 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:medida_del_juez:afirma_lo_no_leido:correctos_marcados`: 0 de 47 (0,0 %), ≤ 0,0 %; `evals:boe-legislacion:duracion_de_las_sesiones:orden`: 452, ≤ 900; `evals:boe-legislacion:duracion_de_las_sesiones:herramienta`: 542, ≤ 900; `evals:boe-legislacion:duracion_del_juez:orden`: 98, ≤ 900; `evals:boe-legislacion:duracion_del_juez:herramienta`: 99, ≤ 900 |
| SC-002 | — | — | SIN TAREA | — |
| SC-003 | T001(hecha) T006(hecha) | cita-cotejar.txtar cita-preparar.txtar  | comprobado por su control | `ci:internal/app/testdata/script/h23-cita-preparar.txtar`, en make ci (verde); `ci:internal/app/testdata/script/h23-cita-cotejar.txtar`, en make ci (verde) |
| SC-004 | T006(hecha) | — | comprobado por su control | `ci:internal/app/cita_test.go:TestCitaPreparar`, en make ci (verde); `ci:internal/app/cita_test.go:TestCitaCotejar`, en make ci (verde) |
| SC-005 | T006(hecha) | — | comprobado por su control | `ci:internal/arch_test.go:TestArquitectura`, en make ci (verde) |
| SC-006 | T001(hecha) T006(hecha) | cita-sin-efectos.txtar  | comprobado por su control | `ci:internal/app/testdata/script/h23-cita-sin-efectos.txtar`, en make ci (verde) |
| SC-007 | T008(hecha) | — | comprobado por su control | `ci:internal/evals/sentencias_test.go:TestCitaSinDocumento`, en make ci (verde); `ci:internal/evals/informe_test.go:TestUmbralesDeJurisprudencia`, en make ci (verde) |
| SC-008 | T007(hecha) | — | comprobado por su control | `ci:internal/evals/conjunto_test.go:TestPreguntasConElFragmento`, en make ci (verde) |
| SC-009 | T006(hecha) T009(hecha) T012(hecha) | — | comprobado por su control | `ci:Makefile:skills-check`, en make ci (verde); `ci:internal/app/skills_test.go:TestSkillsDelRepositorio`, en make ci (verde); `ci:internal/app/herramientas_test.go:TestHerramientasDelServidor`, en make ci (verde) |
| SC-010 | T011(hecha) | — | tareas hechas | — |

## 7. Cambios posteriores a la revisión final

Cada commit posterior a lo que juzgó la primera ronda de la revisión final (`8d0533d`), con la ronda que lo vio y su veredicto, y lo que toca fuera de `gates/` (ADR 0030):

- `be7e8ec` fix(H23): motivos de la revisión final: lo vio la ronda 2 (juez A: aprobado; juez B: aprobado). Toca `internal/core/cita/ficha_test.go`.
- `29e9b7d` docs(H23): veredictos de la revisión final: solo registros de `gates/`.
- `bf4befd` docs(H23): registros del run: solo registros de `gates/`.

### Cambios que ningún juez vio

Ninguno.

## 8. Cómo comprobarlo y consumo

- Escenarios manuales: `specs/019-h23-skill-jurisprudencia-ninguna/quickstart.md`. Suite de aceptación congelada: `specs/019-h23-skill-jurisprudencia-ninguna/aceptacion/` (activada en `internal/app/testdata/script/`).
- Run `d25953aa`: 10 h 31 min de reloj. Coste por paso y por rol: `scripts/coste-run.sh d25953aa`.
