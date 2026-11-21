---
id: H43
fase: "1"
depende_de: [H42, H37]
---

# H43 · La aceptación de la fase 1: evals de varias materias sobre la base (ADR 0059)

- **Objetivo**: comprobar con evals que la base cumple lo que promete la fase 1 en cinco o seis materias. En cada una, la
  norma se lee y se cita, ningún plazo sale de memoria, la cobertura se declara y lo que falta se busca y se marca como
  no verificado (`docs/ROADMAP.md` §4, «Aceptación de la fase»).
- **Entrega**:
  - las evals de las materias que falten hasta cinco o seis, sobre las que ya tiene `evals/consulta-juridica/`
    (seguros, con el caso de la póliza, y laboral) y la tributaria de H42; por ejemplo, civil, consumo y familia, con
    el spec eligiendo cuáles;
  - cada pregunta grabada con lo que la skill lleva a leer (ADR 0043).
  El hito no cambia ninguna skill ni ningún motor. Una eval que no pasa es un defecto de la base y se corrige en su
  capacidad, sin caso especial para la materia (principio X).
- **Alcance**: solo evals y sus grabaciones; fuera, cualquier paquete de materia nuevo.
- **Controles**: los umbrales sin modelo de lo no cubierto y el juez de las clases medidas (ADR 0037), con los topes
  de siempre.
- **Aceptación**: las evals de las cinco o seis materias pasan en `main` con el modelo de `MODELO_DE_EVALS`.
