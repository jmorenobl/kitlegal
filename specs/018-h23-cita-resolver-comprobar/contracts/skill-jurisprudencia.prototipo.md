---
name: jurisprudencia
description: >-
  Comprueba que una sentencia existe antes de citarla, y la cita. Úsala cuando la conversación nombre, cite o pida una
  sentencia o un auto de un tribunal español —por su número y su fecha («STS 1088/2023, de 4 de julio»), por su ROJ
  (STS 3144/2023) o por su ECLI (ECLI:ES:TS:2023:3144)—; cuando se pegue o se adjunte el texto de una sentencia; cuando
  se pregunte qué dice la jurisprudencia o qué han resuelto los tribunales sobre una materia; o cuando la respuesta
  vaya a apoyarse en una sentencia, la aporte la persona o la recuerdes tú. Resuelve cada referencia con el binario
  kitlegal en el buscador del CENDOJ, cita solo la que se resuelve, con su ECLI y su ROJ, y dice cuál no ha podido
  comprobar. No busca sentencias por materia ni lee su texto: prepara la búsqueda para que la haga la persona.
metadata:
  kitlegal-applets: cita
---

# Jurisprudencia: comprobar una sentencia antes de citarla

Esta skill hace una cosa: comprobar que una sentencia existe antes de que la respuesta la cite. Una sentencia que se
recuerda puede no existir, o existir con otro número o con otra fecha, y citarla así hace daño a quien se fía. Cada
referencia se resuelve con `kitlegal cita`, que la busca en el buscador del CENDOJ —el Centro de Documentación Judicial
del Consejo General del Poder Judicial— y devuelve, si existe, sus datos y la dirección oficial de su documento.

Lo que no hace: no busca sentencias por materia, no descarga ni lee el texto de ninguna y no dice qué resolvió una
sentencia cuyo texto no tiene delante. La búsqueda la hace la persona, con la consulta que le preparas. La skill
comprueba y cita: no tramita nada y no sustituye el asesoramiento de un profesional.

## Protocolo

Cada referencia se resuelve de una de dos formas, que devuelven el mismo sobre, con `fuente`, `url`, `fecha_consulta`
y `hash` junto a `data`:

- **Con la herramienta** `cita_resolver`, si entre las tuyas hay una con ese nombre, solo o detrás del prefijo que le
  ponga tu agente, como `mcp__kitlegal__cita_resolver`. Si la tienes, úsala siempre.
- **Con la orden** `kitlegal cita resolver … --json`, si no la tienes. Ejecútala tal como está escrita y lee su salida
  entera, sin filtrarla ni recortarla.

Donde un paso dice que la orden termina con un código, con la herramienta el resultado viene marcado como error y su
`data.clase` dice cuál («Comandos»).

### 1. Qué hay que resolver

- **Toda sentencia que la respuesta vaya a citar**, la haya dado la persona o la propongas tú. Antes de citarla.
- También la que conoces bien: lo que recuerdas de una sentencia no es una comprobación.
- **De una en una**: espera el resultado de cada consulta antes de lanzar la siguiente, aunque haya varias sentencias
  que comprobar. El buscador se consulta despacio, y varias consultas a la vez se quedan sin responder.

### 2. Con qué se resuelve cada referencia

| Lo que tienes | Consulta |
|---|---|
| Su ECLI (`ECLI:ES:TS:2023:3144`) | por el ECLI |
| Su ROJ, dado como tal («ROJ: STS 3144/2023») | por el ROJ; si viene con su fecha, con ella |
| Su ECLI y su ROJ | por el ECLI |
| Una cita con número y fecha («STS 1088/2023, de 4 de julio») | por el número de resolución con su fecha; y, solo si eso termina en «no encontrado», por el ROJ con esa misma fecha |

```bash
kitlegal cita resolver ECLI:ES:TS:2023:3144 --json
kitlegal cita resolver --roj "STS 3144/2023" --json
kitlegal cita resolver --resolucion 1088/2023 --fecha 2023-07-04 --json
kitlegal cita resolver --roj "STS 1088/2023" --fecha 2023-07-04 --json
```

Con la herramienta, los mismos datos van en `ecli`, `roj`, `resolucion` y `fecha`.

- **El número de una cita no es su ROJ.** «STS 1088/2023» es la sentencia número 1088 de 2023 de una sala; el ROJ es
  otro número, el del repertorio del CENDOJ. Por eso la cita se prueba primero como número de resolución.
- **El segundo intento lleva siempre la fecha**: las siglas y el número de la cita como ROJ, con la fecha de la cita.
  Los ROJ son correlativos, y casi cualquier número es el ROJ de alguna sentencia de otro asunto: sin la fecha, darías
  por comprobada una sentencia que no es la que te han citado.
- **Una cita con número y sin fecha no se consulta.** No la pruebes como ROJ. Di que sin la fecha no se puede comprobar
  («Lo que no se ha podido comprobar») y pide la fecha.
- La fecha se escribe `AAAA-MM-DD`. Si la cita da el día y el mes y no el año, el año es el del número.
- Pasa el ECLI, el ROJ y el número tal como te los han dado. No conviertas uno en otro ni corrijas un número.
- Si el primer intento termina con un fallo que no es «no encontrado», no hay segundo intento.

### 3. Leer el resultado

- **Código 0**, o un resultado de la herramienta que no es un error: `data.resoluciones` trae las resoluciones
  encontradas, cada una con `ecli`, `roj`, `organo`, `fecha`, `numero_resolucion`, `numero_recurso`, `ponente` y `url`.
  - **Una sola**: está comprobada. Cítala («Cómo se cita»).
  - **Más de una**: la referencia no identifica una sola sentencia. No cites ninguna. Di que con ese número y esa
    fecha hay más de una resolución («Lo que no se ha podido comprobar»), da de cada una su órgano, su ROJ y su ECLI,
    sin los corchetes de la cita, y pide el ECLI o el ROJ de la que busca la persona. No elijas por ella, tampoco por
    el órgano.
  - Si `pagina_completa` es verdadero, dilo: puede haber más resoluciones que las que ves.
  - Si `cobertura` es `tribunal-constitucional-no-cubierto`, no hay ninguna resolución: ve al paso 6.
- **Código 3 (`no-encontrado`)**: el buscador no tiene ninguna resolución con esa referencia, o ninguna con esa fecha.
  Si era el primer intento de una cita con número y fecha, haz el segundo. Si no, la sentencia no está comprobada.
- **Código 4 (`fuente-no-disponible`)**: el buscador no responde. No está comprobada. No insistas.
- **Código 5 (`limite-o-tos`)**: el buscador ha bloqueado la consulta o ha respondido algo que kitlegal no reconoce.
  No está comprobada. No insistas, no pruebes otra forma de la referencia y no busques otra manera de entrar.
- **Código 2 (`argumentos`)**: la referencia no tiene la forma de un ECLI español, de un ROJ o de un número con su
  fecha; el mensaje dice qué falla. Díselo a la persona con sus palabras y pídele la referencia.

Una sentencia que no queda comprobada no se cita: se dice, con su línea.

### 4. Lo que no se ha leído no se resume

- Sin el texto delante, de una sentencia comprobada solo se dan su cita y los datos de `data`: órgano, fecha, número de
  resolución, número de recurso, ponente y dirección del documento.
- No digas qué resolvió, qué doctrina sienta, a quién dio la razón ni de qué trata. Tampoco de memoria, tampoco «en
  términos generales». Di que no has leído su texto y dónde puede leerlo la persona: en la dirección de `url`.
- **Si la persona pega o adjunta el texto**, léelo, resuelve la sentencia —el texto del CENDOJ trae en su cabecera el
  ROJ y el ECLI— y cítala. Entonces sí puedes decir lo que ese texto dice, y solo eso.
- Si ese texto no se puede resolver, puedes leerlo, pero no escribas su cita: dilo con su línea.

### 5. Una pregunta por materia

«Qué dice la jurisprudencia sobre…» no se responde con sentencias de memoria, y kitlegal no busca por materia. Prepara
la búsqueda para que la haga la persona:

- el texto que escribiría en el buscador: los términos de la materia, pocos y precisos, como los escribiría una
  sentencia;
- el orden jurisdiccional de la materia y el órgano cuya doctrina interesa;
- y las fechas entre las que buscar, si la pregunta las acota.

Dale la dirección del buscador, `https://www.poderjudicial.es/search/indexAN.jsp`, y dile que vuelva con el ECLI o el
ROJ de las sentencias que elija, o con su texto: lo que traiga se resuelve por los pasos 1 a 4. Esa respuesta no cita
ninguna sentencia.

### 6. Una sentencia del Tribunal Constitucional

El Tribunal Constitucional no está en el buscador del CENDOJ, y kitlegal no puede comprobar sus sentencias. Se
reconocen por la referencia: «STC», «ATC», «Tribunal Constitucional» o un ECLI que empieza por `ECLI:ES:TC:`.

- Di que no está cubierta y que no la has comprobado.
- Da la dirección de su buscador, `https://hj.tribunalconstitucional.es/`, para que la persona la consulte.
- No la cites, y no escribas su ECLI si no lo ha escrito la persona.

## Cómo se cita

Una sentencia comprobada se cita así, con todo en la misma línea:

```text
STS 1088/2023, de 4 de julio [ECLI:ES:TS:2023:3144, ROJ: STS 3144/2023] — https://www.poderjudicial.es/search/AN/openDocument/…
```

- Delante de los corchetes, las siglas del ROJ, el `numero_resolucion` y la `fecha`, escrita con el día y el mes.
- Dentro de los corchetes, el `ecli`, una coma, `ROJ:` y el `roj`, exactamente como los da `data`. Los corchetes se
  abren y se cierran en la misma línea y terminan en el ROJ.
- Detrás, la `url` de `data`, entera.
- **Todos los datos son los de `data`**, también si la persona escribió otros: si citó el ROJ donde iba el número, o
  una fecha distinta en una consulta que no la llevaba, la cita lleva los de `data` y la respuesta dice la diferencia.

## Lo que no se ha podido comprobar

Cada referencia que no queda comprobada lleva su línea, que empieza así y sigue con el motivo:

```text
⚠ SENTENCIA NO COMPROBADA: STS 9999/2023, de 1 de enero. No se encuentra ninguna resolución con ese número y esa fecha, ni con ese ROJ y esa fecha.
```

- Una línea por referencia, con la referencia tal como se dio.
- El motivo, con palabras de quien pregunta y sin el código: no se encuentra; el buscador no responde; el buscador ha
  bloqueado la consulta; no se ha podido consultar; falta la fecha; hay más de una resolución con ese número y esa
  fecha.
- Si no tienes ni la herramienta ni el binario —el shell no encuentra `kitlegal`, o no puedes ejecutar órdenes—, no has
  podido consultar: cada sentencia lleva su línea con ese motivo, y ninguna se cita.
- Una sentencia que propones tú y no se resuelve no se cita. Si la nombras, es solo en su línea.

## Comandos

`kitlegal` se invoca desde el `PATH`, y la herramienta, por el nombre de su fila. Consulta el buscador del CENDOJ: una
consulta por referencia, despacio, y lo comprobado se sirve de la caché durante treinta días. Códigos de salida, y entre
paréntesis la `data.clase` del error de la herramienta: 0 correcto, 2 (`argumentos`) la referencia no tiene su forma,
3 (`no-encontrado`) no se encuentra, 4 (`fuente-no-disponible`) el buscador no responde, 5 (`limite-o-tos`) el buscador
ha bloqueado la consulta o su respuesta no se reconoce.

<!-- inicio de la tabla de comandos: generada desde --describe con make skills-sync, no editar -->

### `kitlegal cita`

| Orden | Herramienta | Qué hace | Qué devuelve en `data` |
|---|---|---|---|
| `kitlegal cita resolver [<ecli>] [--roj=<roj>] [--resolucion=<resolucion>] [--fecha=<fecha>]` | `cita_resolver` | Comprueba en el buscador del CENDOJ que existe una resolución judicial identificada por su ECLI, su ROJ o su número con su fecha, y devuelve sus datos y la dirección de su documento. | objeto con `resoluciones`, `pagina_completa`, `cobertura` |

La orden y la herramienta de cada fila devuelven el mismo sobre: `ok`, `fuente`, `url`, `fecha_consulta`, `hash`, `data`; con `ok` falso, `data` lleva `clase` y `mensaje`.

Banderas comunes: `--json`, `--timeout <valor>`, `--offline`, `--dry-run`, `--describe`, `--no-graph`, `--asunto <valor>`, `--verbose`.

<!-- fin de la tabla de comandos -->

## Reglas

1. **Ninguna sentencia sin resolver.** No cites una sentencia que no hayas resuelto en esta conversación, ni des por
   comprobada una que no lo esté.
2. **Ningún ECLI de memoria.** No escribas ningún ECLI que no salga de `data` en esta conversación o que no haya escrito
   la persona. Tampoco uno que deduzcas de un ROJ o del número de una sentencia.
3. **Nada de lo que no se ha leído.** No resumas ni caracterices una sentencia cuyo texto no tienes delante, y no
   inventes contenido legal ni referencias.
4. **Si la respuesta habla de normas**, distingue ley de reglamento, señala lo que una comunidad autónoma puede haber
   regulado de otro modo y no apliques el procedimiento común a lo que la ley regula aparte. No cites de memoria el
   texto de ningún artículo.
5. **No concluir «no existe».** Que el buscador no dé una sentencia no prueba que no exista: puede no estar publicada,
   o estar con otro número o con otra fecha. Di que no la has podido comprobar.
6. **Un bloqueo se dice, no se sortea.** Si el buscador bloquea o no responde, no insistas ni busques otro camino.
7. **Ninguna acción con identidad.** No presentes, notifiques, firmes ni tramites nada en nombre de nadie, ni lo
   simules.
