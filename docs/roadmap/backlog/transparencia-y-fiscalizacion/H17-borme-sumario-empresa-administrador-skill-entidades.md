---
id: H17
fase: "transparencia"
---

# H17 · `borme sumario | empresa | administrador` + skill `entidades-y-registros`

- Vía API BOE; `Persona` sin NIF, `lb:administra` con `valid_from/to`, `lb:participa`. La skill responde qué sociedades y administradores aparecen en BORME para una entidad o un nombre, con el histórico de nombramientos y ceses.
- Controles: test explícito de que `Persona` rechaza cualquier prop que parezca NIF/DNI (regex) en `graph.Apply` (la regla existe desde H7; aquí se prueba con el primer emisor real); fixtures; evals.
