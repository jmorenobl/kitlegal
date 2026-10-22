# Los hitos del roadmap

Cada hito es un fichero, y el directorio en el que está dice su estado. Las fases, la Definition of Done, la
arquitectura, los controles y el ritual por hito están en `docs/ROADMAP.md`; aquí está lo que se hace, uno a uno.

| Directorio | Qué hay | Nombre del fichero |
|---|---|---|
| `done/` | Los hitos terminados, con el texto con que se lanzaron y sus enmiendas fechadas | `H9-…md` |
| `in-progress/` | El que se implementa ahora: como mucho uno | `H29-…md` |
| `todo/` | La cola, en orden: el prefijo es la posición | `010-H29-…md` |
| `backlog/` | Los candidatos, sin número ni especificación | libre |

`make roadmap` enseña el estado: lo hecho, lo que está en curso y la cola en orden.

## Un fichero de hito

```markdown
---
id: H30
fase: "1"
depende_de: [H29]
---

# H30 · Título del hito

- **Objetivo**: …
- **Entrega**: …
- **Alcance**: …
- **Controles**: …
- **Aceptación**: …
```

`id` es el del nombre del fichero y no cambia nunca: lo citan los `specs/`, los ADR, el CHANGELOG, los commits y la
serie de cierres. `fase` es la de `docs/ROADMAP.md` §4. `depende_de` es opcional y nombra hitos con fichero. La entrada
—objetivo, entrega, alcance, controles y aceptación— es lo que el workflow lee como entrada del hito, y la tiene que
tener completa antes de lanzarlo. Un fichero del backlog es texto libre; si lleva cabecera, es un hito con número que
espera, como los de `transparencia-y-fiscalizacion/`.

## Del backlog a `done/`

1. **Al backlog** va cualquier candidato, en una línea si hace falta.
2. **A `todo/`** pasa cuando se especifica: recibe el siguiente número libre y una posición en la cola. Reordenar
   es renombrar el prefijo; los huecos van de diez en diez para no tocar los demás ficheros. `make roadmap-sync` añade
   el hito a la lista que admite el workflow, y `make ci` falla si la lista no es la del directorio.
3. **A `in-progress/`** pasa en `main` cuando se decide lanzarlo: es la decisión de la persona sobre qué viene ahora (ADR
   0018), y `scripts/hito.sh H<n>` solo lanza el hito que está ahí.
4. **A `done/`** lo pasa el workflow en el primer commit de la rama del hito (`scripts/workflow/hito-a-done.sh`): al
   fusionar la propuesta de cambio, `main` lo tiene hecho y `in-progress/` queda libre. Como el movimiento va en la rama
   desde el principio, lo ven todas las revisiones y todas las mediciones (ADR 0030). Un hito que no se lanza con el
   workflow, como H29, se mueve a mano en su propuesta de cambio.

`make ci` comprueba la forma de cada fichero, que ningún identificador tiene dos ficheros, que `in-progress/` tiene como
mucho uno, que dos hitos de la cola no comparten posición y que cada dependencia existe y está hecha, en curso o antes
en la cola (`internal/workflow/roadmap_test.go`).
