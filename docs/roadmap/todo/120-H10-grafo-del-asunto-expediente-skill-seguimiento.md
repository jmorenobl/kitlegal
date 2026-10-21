---
id: H10
fase: "2"
depende_de: [H37]
---

# H10 · Grafo del asunto + `expediente` + skill `seguimiento-expedientes`

- `.kitlegal/case.db` y el `config.yaml` del asunto, con los municipios que mira el art. 30.6 LPAC —el del interesado y el de la sede del órgano—, no el de quien usa el kit (ADR 0027); nodo `Consulta`, `lb:en_asunto`, `--asunto`; `ATTACH` de `world.db` (ADR 0014, pieza G1). `expediente add | listar | vencimientos`: solicitudes, recursos y alegaciones presentados ante cualquier administración y actos notificados, con la administración y el órgano como datos del asunto —el DIR3 por `territorio` si es un ayuntamiento; si no, lo que dice la notificación (art. 40.2 LPAC)— y con los silencios calculados por `plazos` según la norma aplicable (la estatal por defecto; la autonómica cuando el territorio la configure). Un expediente de un procedimiento que la ley regula aparte se registra, pero `vencimientos` declara su plazo no cubierto en `data`, con la norma que lo rige, en lugar de aplicar el del común. `graph check` gana `plazo-vencido` y `plazo-sin-base`. La skill responde «¿qué tengo pendiente y qué vence?» y avisa del silencio que habilita reclamación.
- Controles: e2e de un flujo completo (registrar una solicitud, avanzar el reloj, `vencimientos` avisa; simular nueva versión en fixture y `check` marca warning); e2e de un asunto ante un ayuntamiento, otro ante un órgano sin DIR3 en `territorio` y un expediente de un procedimiento que la ley regula aparte (Definition of Done §1.13); test de que `case.db` nunca contiene texto ni props del mundo, solo ids; ningún comando saca el asunto de `.kitlegal/` (la exportación es backlog y pedirá confirmación); evals con una pregunta de un despacho sobre el asunto de un cliente y otra de un ciudadano sobre el suyo.
