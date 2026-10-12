# Evidencia del ADR 0038: cuatro sondeos sobre las preguntas por materia de `boe-legislacion`

Material de los cuatro sondeos que motivan el ADR 0038, hechos el 2026-10-11 y el 2026-10-12 en sesiones interactivas
a petición de Jorge, fuera de cualquier run (ADR 0032). Como todo `evidencias/`, un run del workflow no la escribe.

No es evidencia de una fuente: ninguno de sus ficheros sale de una fila de `docs/SOURCES.md`, y no la escribió el
paso `grabar_datos`. `manifiesto.json` da la huella y el tamaño de cada fichero.

## Qué se midió

Cuántas respuestas marca el juez en `afirma_lo_no_leido`, con sus tres votos, en las evals que preguntan por materia
sin nombrar el artículo. Cada sondeo es `make evals-sondeo` con 15 repeticiones por modo, la concurrencia del job y
el juez del job (`claude-opus-5-5`); sesiones y votos, con Claude Code 2.1.289. Ninguna sesión quedó sin medir y
ninguna respuesta sin juzgar.

| Carpeta | Fecha | Modelo | Skill | Evals | Marcadas |
|---|---|---|---|---|---|
| `sondeo-20261011-sonnet/` | 2026-10-11 | `claude-sonnet-5-5` | la de `main` (v0.1.7) | 15 | 4 de 30 |
| `sondeo-20261011-sonnet-con-rubricas/` | 2026-10-11 | `claude-sonnet-5-5` | v0.1.7 con la rúbrica de cada artículo en `boe indice` | 15 | 6 de 30 |
| `sondeo-20261012-opus/` | 2026-10-12 | `claude-opus-5-5` | la de `main` (v0.1.7) | 15 y 16 | 3 de 60 |
| `sondeo-20261012-sonnet-con-repaso/` | 2026-10-12 | `claude-sonnet-5-5` | v0.1.7 con un paso 6, «Repasar antes de responder» | 15 y 16 | 2 de 60 |

Por eval, el del 2026-10-12 con Opus da 1 de 30 en la 15 y 2 de 30 en la 16, y el del repaso, 2 de 30 y 0 de 30.

## Qué hay en cada carpeta

- **`salida.txt`**: lo que imprimió el sondeo, con su recuento.
- **Los del 2026-10-11**: `respuestas/`, una por sesión. El de las rúbricas lleva además `votos/`, los votos del
  juez de cada respuesta. Ese día el sondeo no guardaba los votos: los del sondeo con rúbricas se pidieron de nuevo
  con `parches/votos-del-sondeo-20261011.py` y coinciden en el recuento, 6 de 30; los del sondeo sin rúbricas no se
  recuperaron, y de él solo queda el recuento de `salida.txt`.
- **Los del 2026-10-12**: `votos.json`, con la respuesta de cada sesión y el juicio del juez, voto a voto, como lo
  lleva el informe del job.
- **`parches/`**: lo que se aplicó sobre `main` para cada sondeo. Son prototipos, no parte del producto.
  - `arnes-del-sondeo.diff`: el sondeo abre también el modo herramienta, conserva su directorio temporal y, desde
    el 2026-10-12, escribe `votos.json`.
  - `indice-con-rubricas.diff`: `boe indice` da la rúbrica de cada artículo, con las de dos normas grabadas, y la
    skill lo dice.
  - `paso-de-repaso.diff`: el paso 6 de `SKILL.md`.

## Lo que no está

- **Los textos que devolvieron las herramientas** en cada sesión, que son los que el juez tuvo delante. Están en
  los transcripts, que llevan rutas del equipo donde se ejecutó y no se versionan.
- **El razonamiento de las sesiones.** El transcript dice si hubo un bloque de razonamiento, pero no guarda su
  texto: en el sondeo del repaso lo hay justo antes de la respuesta en 28 de 60 sesiones, y en el del 2026-10-11
  sin rúbricas, en 7 de 30. No se sabe si en ese bloque se hizo el repaso.

## Salvedades

- **En el sondeo con Opus, el juez es el mismo modelo que el juzgado**, que el ADR 0037 descarta para el job. Las
  tres respuestas marcadas se leyeron contra los textos devueltos y las marcas son coherentes con la rúbrica; las 57
  sin marcar no se leyeron.
- **En el modo orden, las sesiones filtran el índice con `python3`**, que el runner del job no tiene: ese modo no
  reproduce el job.
- **Con 30 o 60 respuestas por condición no se puede ordenar una condición frente a otra.** Lo que los cuatro dicen
  juntos es que ninguna se acerca a cero.
