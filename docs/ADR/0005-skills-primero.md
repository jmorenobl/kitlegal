# 0005 · Skills primero: el producto son las skills y el binario es su herramienta

- **Estado**: aceptada
- **Fecha**: 2026-09-11
- **Hito**: transversal (reordena el roadmap desde H7)

## Contexto y problema

`kitlegal` nació como «un binario Go multicall + skills agénticas», y los documentos lo contaban en ese
orden. El roadmap lo reflejaba: la *Definition of Done* era íntegramente de Go, de los 24 controles de
calidad solo uno medía skills, la fase del grafo no entregaba ninguna y el ritual colocaba la skill al
final («core → adaptador → applet → skill»). Un hito podía cerrarse sin que ninguna skill supiera
resolver nada nuevo.

Eso invierte el propósito. Quien usa `kitlegal` es un agente que razona con una skill; el binario
existe para que ese razonamiento se apoye en datos verificables y cálculos deterministas. Si la
planificación parte del binario, se construyen herramientas que ninguna skill necesita y las skills
llegan tarde y sin medir.

## Opciones consideradas

1. **Mantener el enfoque y añadir entregables de skills a los hitos existentes.** Barato, pero deja la
   prioridad donde estaba: los hitos siguen definidos por lo que hace el binario.
2. **Separar skills y binario en dos repos o dos roadmaps.** Rompe la decisión cerrada de un solo repo
   y duplica la fuente de verdad (`data/*.yaml`).
3. **Skills primero en el mismo repo**: cada hito se define por lo que una skill pasa a poder resolver,
   se mide con evals y el binario crece solo con las herramientas que esas skills necesitan.

## Decisión

Se adopta la **opción 3**, como principio VIII de la constitución:

- El producto son las skills (`SKILL.md`, `references/` generadas, tabla de comandos). El binario Go
  es la capa de herramientas deterministas: solo va a Go lo que exige determinismo o verificabilidad
  (acceso a fuentes, parseo, caché, fechas y plazos, identificadores, hashes, grafo).
- Todo hito entrega o mejora una skill medible con evals, o protege las que ya existen (fundación,
  contratos, release). Una herramienta que ninguna skill usa no se construye.
- Un hito se planifica de fuera adentro (skill y evals → herramientas) y se implementa de dentro afuera.
- La *Definition of Done* y los controles incorporan las skills: evals escritas antes que el código,
  comprobación mecánica de `SKILL.md` y de `references/` en `make ci`.

No se reabre ninguna decisión cerrada: «skills sin código», el multicall y el sobre de salida siguen
igual. Cambia la prioridad y cómo se mide un hito. La fase 0 (H0-H6) no cambia: ya terminaba en la
skill `boe-fiscal` funcionando con el binario.

## Consecuencias

**A favor**

- Cada hito deja algo que el agente sabe hacer, y se puede comprobar con evals.
- YAGNI tiene un criterio objetivo: sin skill que la use, una herramienta no entra.
- El orden del roadmap sigue al valor (primero lo que se usa en el municipio, ver ADR 0006), no a las
  capas técnicas.

**En contra, y asumido**

- Las evals con modelo cuestan dinero y no son deterministas; por eso las citas se comparan por
  identificador y el job con modelo es semanal o manual, no parte de `make ci`.
- Los applets anteriores al grafo (`territorio`, `placsp`, `bdns`) incorporan `Emit` en H17, cuando el
  grafo existe, en lugar de emitir desde el primer día.
- La numeración de hitos cambia a partir de H7. Los artefactos que citen números antiguos deben
  actualizarse con la tabla siguiente.

## Correspondencia con la numeración anterior

| Antes | Ahora | Nota |
|---|---|---|
| H0-H6 | H0-H6 | Sin cambios (H5 añade el formato de evals y la comprobación mecánica de skills) |
| H7 `ids` | H7, H8, H12… | Absorbido: cada tipo de id entra con el hito que lo usa |
| — | H7 `territorio` + `legal-core` v0 | Nuevo (ADR 0006); recoge la verificación de leyes vertebrales de la antigua H10 |
| H8 `cita` | H8 | |
| H9 `plazos` | H9 | Festivos locales por territorio |
| H10 `boe sumario/vigilar/eli` | H24 | |
| H11 `--describe` completo | H10 | |
| H12 grafo | H17 | Incorpora `Emit` a los applets existentes |
| H13 `boe analisis` | H25 | |
| H14 grafo del asunto | H18 | Con `expediente` (antes en H26) |
| H15 spike de fuentes | H11 | Añade BOCM y TEU; BORME se verifica en H19 |
| H16 `placsp` | H12 | Sin grafo ni anomalías |
| H17 `bdns` | H13 | |
| — | H14 `boletin` + `edictos` | Nuevo |
| — | H15 ordenanzas vía boletín | Nuevo |
| H25 `escrito` | H16 (mínimo) + H26 (grafo y `check`) | |
| H18 `borme` | H19 | |
| — | H20 `presupuesto` | Nuevo |
| H19 anomalías | H21 | Umbrales relativos a municipios comparables |
| H20 cruces + pack | H22 | |
| H21 `vigilar run` | H23 | |
| H26 exportaciones | H27 | `expediente` pasa a H18 |
| H22 MCP | H28 | |
| H23 `skills install` | H29 | |
| H24 `pkg/legalkit` | H30 | |
