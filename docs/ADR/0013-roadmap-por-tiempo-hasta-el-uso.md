# 0013 · Roadmap por tiempo hasta el uso: actuar antes que consultar, backlog sin número y bitácora de uso

- **Estado**: aceptada
- **Fecha**: 2026-09-13
- **Hito**: transversal (renumera desde H6; el detalle numerado termina en la fase 3)

> Nota (2026-09-27): el ADR 0027 cambia el encuadre de las fases 2 y 3 sin tocar su orden. La 2 pasa a ser
> «actuar: llevar un asunto ante cualquier administración», genérica en la materia salvo lo que la ley regula
> aparte, que se declara no cubierto; la 3, «consultar lo que hacen las administraciones, empezando por el
> municipio». Entra H20 (la redacción a una fecha) entre H7 y H8 con el siguiente número libre.

## Contexto y problema

El roadmap tenía 31 hitos en ocho fases, todos numerados y detallados. Su primera mitad ya cumplía el
criterio del proyecto —cada hito deja algo usable y la primera skill llega en H5—, pero la ordenación de
la segunda no seguía al valor sino a las capas:

- La fase 3 (H17–H23) mezclaba tres cosas distintas: infraestructura (`internal/graph`), fuentes de
  consulta (BORME, presupuestos) y fiscalización (anomalías, cruces, `vigilar run`). Las fases 4 y 5
  esperaban a la 3 entera, de modo que la vigilancia, que no hace falta para usar el kit, bloqueaba los
  escritos fundamentados, el sumario del BOE y la distribución.
- `escrito generar` (H16) solo depende de `territorio`, `cita` y `plazos` (H7–H9), y estaba detrás de
  cinco hitos de fuentes municipales de los que no depende. El bucle completo —pedir algo al
  ayuntamiento, seguir el plazo, avisar del silencio— llegaba en el hito 19.
- `expediente` (H18) iba atado al grafo del asunto, que a su vez esperaba al grafo del mundo (H17).
- La distribución (H28–H30) no depende de nada salvo de que existan applets y `--describe`, y estaba al
  final por convención, no por dependencia.
- H10 (`--describe` completo, `schemas/` generado, `render` markdown) no era un hito: la generación de
  esquemas existe desde H1, `make schema-check` es gate desde H0 y la validación de salida es Definition
  of Done desde H4; el markdown solo lo pide el hito de escritos. Como hito propio no entregaba ni
  mejoraba ninguna skill ni protegía nada que no estuviera ya protegido, que es lo que el principio VIII
  exige a un hito para existir.
- El release firmado (H6) es para terceros; desde H5 `make install` deja el binario y las skills en el
  sitio para quien desarrolla, que es el único usuario hasta que haya alguien más.
- Detallar 31 hitos con número es lo contrario de planificar «solo la fase en curso al detalle» (§0 del
  propio roadmap): crea una obligación falsa y nada en el ritual hace que el uso real cambie el plan.

## Opciones consideradas

1. **Dejar el orden** y aceptar que la fiscalización llegue antes que los escritos y la distribución.
   Contradice el criterio de uso temprano y el propio §0.
2. **Mover solo la fase 3 al final** sin tocar lo demás. Arregla el tapón pero deja los escritos detrás de
   las fuentes, el `expediente` atado al grafo tardío y 31 hitos numerados.
3. **Reordenar por tiempo hasta el uso**, numerar solo hasta donde se ve (fases 0–3) y convertir el resto
   en backlog sin número, con una bitácora de uso que decida qué entra.

## Decisión

Se adopta la **opción 3**. El grafo se trata en el ADR 0014, que es lo que hace posible el orden de la
fase 1 y 2; este ADR fija lo demás:

- **Fases numeradas: 0 a 3.** Fundación (H0–H5), terreno y memoria (H6–H9), actuar (H10–H11) y
  consultar el municipio (H12–H19). Cada una termina en algo que se usa; la 2 es el primer bucle completo
  (consultar → actuar → seguir) y llega en el hito 11, no en el 19.
- **Actuar antes que consultar las fuentes municipales.** `expediente` (H10) y `escrito generar` (H11)
  van inmediatamente después de `plazos`, porque no dependen de PLACSP, BDNS ni boletines. Su primera
  versión fundamenta solo con normativa estatal (LTAIBG, LPAC), que es justo lo que necesitan la solicitud
  de acceso a información y el recurso de reposición; los escritos que se apoyen en ordenanzas esperan a
  la fase 3.
- **BORME y presupuestos son consulta, no fiscalización**, y van con las demás fuentes municipales
  (H17, H18). La fiscalización (anomalías, cruces, `vigilar run`) pasa entera al backlog.
- **El release cierra la fase 3** (H19, `v0.1.0`). Hasta entonces `make install`. Se adelanta en el
  momento en que alguien distinto de quien desarrolla tenga que instalarlo.
- **H10 se disuelve** en lo que ya era Definition of Done (esquemas y drift desde H4, tabla de comandos
  desde H5) y en H11 (`render` markdown). La revisión de `internal/cli` con cinco applets se hace al
  cerrar H9, como nota, no como hito.
- **Backlog sin número.** Todo lo posterior a la fase 3 se agrupa por tema (distribución, verticales,
  territorios, fuentes, profundidad del BOE y escritos, fiscalizar y vigilar) sin orden entre grupos. Un
  candidato recibe número cuando entra en la fase en curso. Ninguna dependencia obliga a un orden entre
  grupos: la distribución solo necesita applets; las verticales solo `boe-legislacion`; la fiscalización
  necesita las fuentes de la fase 3 y el grafo, y por eso es la única que no puede entrar antes.
- **Bitácora de uso, `docs/USO.md`.** Quien usa el kit apunta qué pidió, qué falló y qué faltó. Cada tres
  o cuatro hitos, antes de abrir el siguiente, se relee: el siguiente hito sale de la bitácora o del
  backlog, y ante discrepancia gana la bitácora y el roadmap se actualiza. Es el único mecanismo que hace
  adaptativo un plan incremental; sin él, «solo la fase en curso al detalle» es una frase.

No se reabre ninguna decisión cerrada: skills primero (ADR 0008), genericidad territorial (ADR 0009),
base antes que vertical (ADR 0012), contrato de applet (ADR 0005, enmendado solo en lo que dice el
ADR 0014), frontera humana, sobre de salida y exit codes siguen igual. Cambia el orden y hasta dónde se
numera.

## Consecuencias

**A favor**

- El bucle completo de valor llega ocho hitos antes. A la fecha de este ADR hay cuatro hitos cerrados: son
  siete hitos hasta ese bucle en lugar de quince.
- Lo que no hace falta para usar el kit (fiscalización, vigilancia, exportaciones) deja de bloquear lo
  que sí (escritos, distribución, profundidad del BOE).
- El detalle numerado coincide con lo que se ve; lo lejano se decide con datos de uso y no de antemano.
- La distribución y las verticales quedan como hojas que se pueden meter en cualquier momento, que es lo
  que son.

**En contra, y asumido**

- Tercera renumeración (ADR 0008, 0012 y esta). Los artefactos que citen números anteriores se leen con
  la tabla siguiente; `specs/NNN-hN-*/` de los hitos cerrados conservan su numeración de origen, que
  coincide (H0–H5 no cambian).
- H13 (PLACSP) sigue siendo el hito más pesado del roadmap y puede desbordar los tres días: si el plan lo
  muestra, se parte en dos sin renumerar (sync + `store` + `licitaciones` con la skill v0; el resto
  después).
- Sin release hasta H19, cualquier persona ajena que quiera probarlo compila. Si eso ocurre antes, el
  release se adelanta y punto.
- La bitácora depende de que se escriba. Si nadie la rellena, el backlog se ejecuta en el orden en que
  está, que es el mejor que sabemos hoy.

## Correspondencia con la numeración anterior

| Antes (ADR 0008/0012) | Ahora | Nota |
|---|---|---|
| H0–H5 | H0–H5 | Sin cambios |
| H6 release `v0.1.0` | H19 | Cierra la fase 3; se adelanta si alguien más tiene que instalar |
| H7 `territorio` + `legal-core` v0 | H6 | |
| H8 `cita` + `cita-verificada` | H8 | Nace emitiendo al grafo (ADR 0014) |
| H9 `plazos` + `legal-core` v1 | H9 | Absorbe la nota de revisión de `internal/cli` |
| H10 `--describe`, `schemas/`, markdown | disuelto | Esquemas y drift: DoD desde H4; tabla de comandos: DoD desde H5; markdown: H11 |
| H11 spike de fuentes | H12 | |
| H12 `placsp` + `store` + `contratacion-publica` v0 | H13 | Nace emitiendo; sin anomalías |
| H13 `bdns` + `subvenciones` v0 | H14 | |
| H14 `boletin` + `edictos` + `bop-y-edictos` | H15 | |
| H15 ordenanzas | H16 | |
| H16 `escrito generar` + `redaccion-escritos` | H11 | Más `render` markdown; emite `Escrito` al asunto |
| H17 `internal/graph` + `Emit` | H7 | Solo la pieza G0 (ADR 0014) |
| H18 grafo del asunto + `expediente` | H10 | Pieza G1; sin las reglas que necesitan más fuentes |
| H19 `borme` + `entidades-y-registros` | H17 | |
| H20 `presupuesto` + `presupuestos-y-cuentas` | H18 | |
| H21 anomalías de contratación | backlog · fiscalizar y vigilar | |
| H22 cruces + pack `fiscalizador` | backlog · fiscalizar y vigilar | El mecanismo de packs va con la distribución |
| H23 `vigilar run` + `anomalies mark` | backlog · fiscalizar y vigilar | |
| H24 `boe sumario\|vigilar\|eli` | backlog · profundidad | |
| H25 `boe analisis` → aristas ELI | backlog · profundidad | Con `eli:cites` y `graph history` |
| H26 escritos fundamentados en el grafo | backlog · profundidad | `Afirmacion`, `fundamenta`, `check` bloqueante, DOCX |
| H27 exportaciones del grafo | backlog · fiscalizar y vigilar | |
| H28 MCP | backlog · distribución | |
| H29 `skills install` + `plugin/` | backlog · distribución | Con el mecanismo de packs |
| H30 `pkg/legalkit` + `v1.0.0` | backlog · distribución | |
| Fase 6 territorios | backlog · territorios | Mismo molde |
| Fase 7 fuentes | backlog · fuentes | Mismo molde |
| Fase 8 verticales | backlog · verticales | Mismo molde (ADR 0012) |
