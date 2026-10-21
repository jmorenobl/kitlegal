---
id: H32
fase: "1"
depende_de: [H31]
---

# H32 · Arquitectura de skills: una entrada genérica, skills por capacidad y las materias como referencias

- **Objetivo**: que una pregunta de cualquier materia entre por una skill que sabe qué capacidad la resuelve y qué cubre la base, y que una materia nueva sea un paquete de ficheros, no una skill ni código.
- **Entrega**: una skill de entrada con el método jurídico genérico (qué tipo de pregunta es, qué capacidad la resuelve, qué cubre la base y, si no lo cubre, cómo lo busca el agente y cómo lo marca); skills por capacidad —normativa, jurisprudencia, plazos, territorio y destinatarios, documentos—, con los nombres que fije el plan; cada materia como un fichero de referencia en cada skill de capacidad (la organización por dominio de la guía de autoría); el **molde de vertical** —qué trae un paquete de materia: normas, referencias, cobertura y evals—, probado con una vertical de ensayo en `testdata/` que entra sin tocar Go ni ningún `SKILL.md`; lo administrativo de hoy como primer paquete; `legal-core` deja de presentarse como «derecho público».
- **Riesgos y dependencias**: activar de más o de menos, que miden las evals de activación; las instalaciones existentes (manifiesto, plugin) con skills que cambian de nombre, que el plan resuelve; depende de H31.
