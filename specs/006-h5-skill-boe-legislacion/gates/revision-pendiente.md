# Revisión final de H5: pendiente de una persona

Todos los motivos de las rondas 1, 2 y 3 se han corregido en la rama. Quedan dos acciones en la plataforma que el
corrector no hace (no empuja ni publica), en este orden, y la segunda es la última del hito antes de la fusión humana.

## 1. Sincronizar el cuerpo de la propuesta de cambio #27 con `gates/pr-h5.md`

Ronda 1, motivo [f] del juez A sobre D8; ronda 2, motivo [f] del juez B sobre las cifras; y ronda 3, que cambia el
control de `make install` y las cifras al corregir el motivo [b] del juez B. El fichero ya está corregido.
En la ronda 1 cambiaron la decisión D8, que dice qué hace cada lector con los alias y con la clave de fusión `<<`, los
controles de la comparación mecánica, de la lectura de trazas, de `SKILL.md` < 300 líneas y de `make install`, y las
correcciones de la revisión en «Decisiones» y «Pendientes». En la ronda 2 cambian «Alcance», con las cifras sobre la
cabeza y lo que llegó después de `536359c`, las cifras de «Controles añadidos» (`TestLeerTrazas` 27 y el test nuevo
`TestLeerTrazasAyudaSeguidaDeConsulta`), las sesiones sintéticas de «Evidencia», las correcciones de la ronda 2 y
«Pendientes». En la ronda 3 cambian «Alcance», con sus cifras y el test nuevo `TestFicherosDelBinario`, el control de
`make install` en «Controles añadidos», las correcciones de la ronda 3 y «Pendientes». El paso `publicar_rama` del
workflow solo crea la propuesta si no existe. Tras empujar la rama:
`gh pr edit 27 --body-file specs/006-h5-skill-boe-legislacion/gates/pr-h5.md`.

## 2. Repetir T031, la ejecución de cierre y la aceptación, sobre la cabeza empujada

Ronda 2, motivo [f] del juez B: FR-082, SC-003 y el punto 10 de la Definition of Done.

- **Por qué.** La ejecución de cierre registrada (35002104338, `gates/evals-cierre.md`) corrió sobre `5c6c552`.
  Después, `ede21ba` (ronda 1) y las correcciones de las rondas 2 y 3 cambian doce ficheros fuera del directorio del
  hito: `Makefile`, `scripts/instalar-skills.sh`, `internal/evals/trazas.go`, `internal/evals/juzgar.go`,
  `internal/evals/conjunto_test.go`, `internal/evals/juzgar_test.go`, `internal/evals/trazas_test.go`,
  `internal/skills/instalacion_test.go`, `internal/skills/sincronia_test.go`, `README.md`, `CONTRIBUTING.md` y
  `CHANGELOG.md`. Entre ellos está el código con el que el job juzga las sesiones y la receta de `make install`. FR-082
  dice que cualquier cambio posterior fuera del directorio del hito obliga a repetirla, así que **hasta la repetición
  FR-082 no se cumple**: la evidencia de «las 10 evals pasan» no cubre la cabeza. `gates/evals-cierre.md` y
  `gates/aceptacion.md` lo dicen en su cabecera.
- **Cómo.** Se ejecuta quickstart §12.3 tal cual sobre #27:
  1. Quitar la etiqueta `evals` (el intento 1 la dejó puesta) y volver a ponerla.
  2. Identificar la ejecución por el último evento `labeled` y `workflowName`, esperarla e imprimir el informe entre
     sus marcas.
  3. Exigir veredicto `aprobado`, con 10 de 10 positivas y 2 de 2 de no activación en verde, las doce sesiones
     terminadas, modelo `claude-haiku-4-5-20251001`, «ninguna petición llegó a la red de una fuente» y `sin_python` con
     sus tres líneas.
  4. La última orden debe dar la lista de ficheros cambiados entre el commit evaluado y la cabeza vacía, o solo con
     ficheros bajo `specs/006-h5-skill-boe-legislacion/`, y la línea `todos bajo specs/006-h5-skill-boe-legislacion/`.
  5. Registrar en `gates/evals-cierre.md` el enlace, el nuevo `headSha`, el modelo, el veredicto, el informe y la
     salida entera de la última orden, y en `gates/aceptacion.md` las sesiones `01-lpac-articulo-21` y
     `06-irpf-rendimientos-del-trabajo` y el `sin_python`. Las dos notas de cabecera se sustituyen entonces por la
     referencia a la ejecución nueva.
- **Cuándo.** Es la última acción de plataforma: después de que la revisión final quede en verde, de empujar la rama y
  de sincronizar el cuerpo de #27. Cualquier corrección posterior fuera del directorio del hito la invalidaría otra
  vez. Si la plataforma descubre un defecto, el arreglo va en una tarea nueva antes de T031 (plan, obligación 1), y la
  repetición vuelve a ser lo último.
