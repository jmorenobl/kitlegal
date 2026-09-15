# Revisión final de H5: pendiente de una persona

Ronda 1 (motivos de `gates/revision-a-r1.json` y `gates/revision-b-r1.json`). Todos los motivos se han corregido en la
rama; queda una acción en la plataforma que el corrector no hace.

- **Sincronizar el cuerpo de la propuesta de cambio #27 con `gates/pr-h5.md`** (motivo [f] del juez A sobre D8).
  El fichero ya está corregido: la decisión D8 dice qué hace cada lector con los alias y con la clave de fusión `<<`, y
  cambian también los controles de la comparación mecánica, de la lectura de trazas, de `SKILL.md` < 300 líneas y de
  `make install`, las correcciones de la revisión en «Decisiones» y «Pendientes». El paso `publicar_rama` del workflow
  solo crea la propuesta si no existe, y el corrector no empuja ni publica. Tras empujar la rama:
  `gh pr edit 27 --body-file specs/006-h5-skill-boe-legislacion/gates/pr-h5.md`.
