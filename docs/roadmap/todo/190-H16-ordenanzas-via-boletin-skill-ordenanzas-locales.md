---
id: H16
fase: "2"
depende_de: [H15]
---

# H16 · Ordenanzas vía boletín + skill `ordenanzas-locales`

- Las ordenanzas se publican íntegras en el boletín provincial o, en comunidades uniprovinciales, en el autonómico (art. 70.2 LRBRL; ordenanzas fiscales, art. 17.4 TRLRHL). `ordenanzas listar | ver --municipio` busca con `boletin` los anuncios de aprobación definitiva y de modificación del territorio, sin crawler por ayuntamiento. El texto sale del formato que publique el boletín; si solo hay PDF, la dependencia de extracción se justifica en el plan. La skill presenta el texto aprobado y las modificaciones que encuentre, citando cada anuncio, sin afirmar que es un texto consolidado.
- Controles: fixtures; e2e con la matriz territorial; evals («¿qué dice la ordenanza de terrazas de mi municipio?»). El crawler de sedes municipales queda para el backlog de fuentes.
- [enmienda 2026-11-21, revisión del plan para la web] Leídos con `kitlegal boe articulo` el 2026-11-21: el art. 70.2 LRBRL (BOE-A-1985-5392, a70) dice que las ordenanzas se publicarán en el «Boletín Oficial» de la Provincia, y remite los presupuestos y las ordenanzas fiscales a la ley de haciendas locales; el art. 17.4 TRLRHL (BOE-A-2004-4214, a17) publica el texto íntegro de las fiscales y de sus modificaciones en el boletín oficial de la provincia o, en su caso, de la comunidad autónoma uniprovincial. Ninguno de los dos dice que, en una comunidad uniprovincial, las demás ordenanzas se publiquen en el boletín autonómico: el spec busca el precepto que lo funde antes de fijarlo, o lo declara como supuesto de la validación en la Comunidad de Madrid.
