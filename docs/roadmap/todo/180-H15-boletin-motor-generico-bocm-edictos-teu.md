---
id: H15
fase: "2"
---

# H15 · `boletin` (motor genérico, BOCM) + `edictos` (TEU) + skill `bop-y-edictos`

- `boletin sumario | buscar` con un motor genérico configurado por YAML (`data/boletines/bocm.yaml` es la primera y única configuración) y `edictos buscar` sobre el Tablón Edictal Único, que es nacional. El filtro por municipio usa `territorio`. Añadir otro boletín es un YAML más, y un adaptador solo si el motor no basta (backlog · territorios). Emite `Publicacion` y `lb:publica`.
- Controles: fixtures de BOCM y TEU; e2e con Leganés (BOCM + TEU) y Tordesillas (TEU, más la cobertura que declara no configurados los boletines autonómico y provincial); `verify-sources`; evals.
- [enmienda 2026-10-21] En la fase 2, por delante de H12: su primera tarea verifica, sin código de producto, el BOCM (sumario, buscador, formatos, condiciones) y el Tablón Edictal Único, con sus filas de `docs/SOURCES.md` y sus fixtures.
