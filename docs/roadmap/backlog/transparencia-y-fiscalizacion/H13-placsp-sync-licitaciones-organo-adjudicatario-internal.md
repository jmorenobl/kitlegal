---
id: H13
fase: "transparencia"
---

# H13 · `placsp sync | licitaciones | organo | adjudicatario` + `internal/store` + skill `contratacion-publica` v0

- Sync incremental de Atom + CODICE XML a `store`; el órgano se identifica por su DIR3, obtenido con `territorio` a partir del municipio. Skill `contratacion-publica` v0: qué ha licitado y adjudicado el ayuntamiento, a quién y por qué importe, incluidos los contratos menores. Si el órgano publica en una plataforma autonómica cuyos datos no llegan a PLACSP, la salida lo declara en la cobertura. Emite `Organo`, `Licitacion`, `Contrato`, `Entidad` y `lb:convoca`, `lb:licita`, `lb:adjudica`, `lb:modifica_contrato` con las props que las anomalías (backlog) necesitarán (importes, procedimiento, número de licitadores, fechas). Sin anomalías.
- Es el hito más pesado del roadmap. Si el plan lo desborda, se parte en dos sin renumerar: sync + `store` + `licitaciones` con la skill v0 primero; `organo`, `adjudicatario` y el resto después.
- Controles: fuzz del parser CODICE; integración de sync incremental (segunda ejecución no duplica, ni en `store` ni en el grafo); e2e con replay para Leganés y para un municipio de otra comunidad; evals.
