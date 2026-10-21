---
id: H33
fase: "1"
depende_de: [H32]
---

# H33 · Cobertura en positivo por capacidad y materia

- **Objetivo**: que la cobertura diga lo que kitlegal cubre, y que todo lo demás quede declarado no cubierto. Hoy dice lo que la LPAC deja fuera, y una pregunta civil no cae en ninguna fila.
- **Entrega**: `data/cobertura.yaml` con lo cubierto por capacidad y materia, cada fila con su norma, y su esquema; lo no declarado no está cubierto; un único `⚠ NO CUBIERTO:` en todas las materias, que dice además con qué herramienta genérica busca el agente lo que falta y cómo lo marca; lo que la ley regula aparte del procedimiento común sigue declarado con su norma; referencias regeneradas.
- **Controles**: evals con una pregunta a cada lado de la frontera en dos materias; el umbral sin modelo `cobertura_sin_declarar` sobre todas las materias.
- **Dependencias**: H32.
