# Revisión final de H5: acciones de plataforma, hechas

Todos los motivos de las rondas de la revisión final se corrigieron en la rama. Las dos acciones de plataforma que el
corrector no hace (no empuja ni publica) están hechas, en este orden, sobre la cabeza
`a73574e5d84b94c6752829cb61f2cff5ee14c850` (`docs(H5): veredictos de la revisión final`). Solo queda lo humano: la
revisión y la fusión (ADR 0007).

## 1. Sincronizar el cuerpo de la propuesta de cambio #27 con `gates/pr-h5.md`

Hecho el 2026-09-15, con la rama empujada y antes de volver a poner la etiqueta `evals`:
`gh pr edit 27 --body-file specs/006-h5-skill-boe-legislacion/gates/pr-h5.md`. Lo que T031 registra después solo
cambia ficheros del directorio del hito, y el cuerpo se vuelve a sincronizar tras empujarlo.

## 2. Repetir T031, la ejecución de cierre y la aceptación, sobre la cabeza empujada

Hecho: intento 2 de T031, ejecución [35023013878](https://github.com/jmorenobl/kitlegal/actions/runs/35023013878),
2026-09-15, sobre `a73574e` (ronda 2 de la revisión final, motivo [f] del juez B: FR-082, SC-003 y el punto 10 de la
Definition of Done).

- **Cómo.** Quickstart §12.3 tal cual sobre #27: se quitó la etiqueta `evals`, que el intento 1 había dejado puesta
  (evento `unlabeled` a las 21:00:00Z), y se volvió a poner (`labeled` a las 21:00:06Z); la ejecución se creó a las
  21:00:10Z y es la única posterior a la etiqueta.
- **Resultado.** Veredicto `aprobado`, sin motivos; 10 de 10 positivas y 2 de 2 de no activación; las doce sesiones
  terminadas con código 0 y `result success`; modelo `claude-haiku-4-5-20251001` en el job y en las sesiones;
  «ninguna petición llegó a la red de una fuente»; ninguna invocación fuera de lo grabado; `sin_python` con sus tres
  líneas; y la última orden con código 0, la lista de ficheros cambiados entre el commit evaluado y la cabeza vacía y
  `todos bajo specs/006-h5-skill-boe-legislacion/`.
- **Dónde.** `gates/evals-cierre.md` (enlace, `headSha`, modelo, veredicto, informe entero en el anexo B y salida
  entera de la última orden) y `gates/aceptacion.md` (sesiones `01-lpac-articulo-21` y
  `06-irpf-rendimientos-del-trabajo` y `sin_python`). `ci` y los cuatro estados de Codecov están en verde sobre la
  misma cabeza.
- **Después.** Solo cambian ficheros del directorio del hito, que FR-082 admite sin invalidar la ejecución. Cualquier
  cambio fuera de él obligaría a repetirla.
