# Revisión pendiente: commits posteriores a los veredictos de la ronda 2

Los veredictos de `gates/revision-a.json` y `gates/revision-b.json` (ronda 2, commit `6ab3add`) se dieron antes de
estos commits, que ningún juez ha juzgado. Júzgalos con la rúbrica completa, como la corrección de esta ronda: lo que
«CONVERGENCIA» dice del último commit «fix(H7.3): motivos de la revisión final» vale también para cada uno de ellos, y
un motivo sobre lo que introducen cuenta aunque no estuviera en el veredicto anterior.

- `eb6b4c8` «fix(H7.3): cierre en la plataforma». Lo hizo el paso `reparar_cierre`, sin juez, después de una primera
  medición del job de evals en rojo sobre `6ab3add` (SC-001: `expresiones_prohibidas:claude-sonnet-5`, 5 de 51).
  Cambia el protocolo de `skills/boe-legislacion/SKILL.md`: la comprobación de la redacción
  (`kitlegal graph check <norma> <bloques> --json`) pasa de ir una vez por norma citada, después de la última lectura
  y antes de la respuesta, a ir en la misma orden que cada lectura, detrás de ella
  (`kitlegal boe articulo … --json && kitlegal graph check … --json`). También cambia `spec.md` («Clarifications»,
  «Cierre 2026-09-30»), `plan.md`, `research.md` («Cierre (2026-09-30)…»), `contracts/skill-boe-legislacion.md`,
  `CHANGELOG.md` y `gates/supuestos.md` (líneas de `reparar_cierre`). La segunda medición del job, sobre `eb6b4c8`,
  está en `gates/evals/boe-legislacion.json`.
- `0222c39` «test(H7.3): sustitutos ejecutables escritos con syscall.ForkLock». Arreglo de un test inestable de
  `internal/evals` que dejó `ci` en rojo sobre `e3bae8d` (ejecución 36678988453 de GitHub Actions); el detalle, en la
  última línea de `gates/supuestos.md`.

Para el cambio de protocolo, las referencias son:

- `docs/ROADMAP.md`, sección H7.3, «La comprobación, en su sitio» (línea 339): «Si el research encuentra que la causa
  está en que `graph check` es la última orden antes de responder, el plan decide con esa evidencia si cambia su sitio
  en el protocolo y lo registra en «Decisiones», manteniendo lo que H7.1 decidió: una vez por norma citada, después de
  leer y antes de responder.»
- La decisión de H7.1: `specs/011-h7-1-graph-check-acotado/spec.md`, FR-040 (línea 234) y la fila de `graph check` de
  «Uso, de fuera adentro» (línea 331).

Juzga si el cambio cumple la sección H7.3 y la decisión de H7.1 con los criterios de la rúbrica. Si una desviación de
lo que pedían es la mejor solución, di si está registrada como decisión, con su evidencia, en `research.md` y en
`gates/supuestos.md` con su etiqueta de impacto; si no lo es, di qué cambiar y dónde.
