# Medida de las skills

Lo que hace que el verde y el rojo de las evals se puedan explicar.

- Guardar las sesiones del job de evals —transcript, traza, salida de error y código de cada una— como artefacto del run, con una retención corta, para que la causa de un fallo se lea en lugar de inferirse. En H26.1, la causa más probable de los `1` de `boe indice` se tuvo que reproducir en local porque el job no guarda las órdenes de las sesiones (`docs/USO.md`, 2026-10-20). Condición de entrada: comprobar que lo guardado no lleva credenciales, rutas de secretos ni nada que no deba publicarse en un repositorio público (ADR 0032), con un escaneo en el propio job antes de subirlo, y medir su tamaño por run.
