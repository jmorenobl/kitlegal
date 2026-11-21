---
id: H16
fase: "2"
depende_de: [H15]
---

# H16 · Ordenanzas vía boletín + skill `ordenanzas-locales`

- Las ordenanzas se publican íntegras en el boletín provincial o, en comunidades uniprovinciales, en el autonómico (art. 70.2 LRBRL; ordenanzas fiscales, art. 17.4 TRLRHL). `ordenanzas listar | ver --municipio` busca con `boletin` los anuncios de aprobación definitiva y de modificación del territorio, sin crawler por ayuntamiento. El texto sale del formato que publique el boletín; si solo hay PDF, la dependencia de extracción se justifica en el plan. La skill presenta el texto aprobado y las modificaciones que encuentre, citando cada anuncio, sin afirmar que es un texto consolidado.
- Controles: fixtures; e2e con la matriz territorial; evals («¿qué dice la ordenanza de terrazas de mi municipio?»). El crawler de sedes municipales queda para el backlog de fuentes.
