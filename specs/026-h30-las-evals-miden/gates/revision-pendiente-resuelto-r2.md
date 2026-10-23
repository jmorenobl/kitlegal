# Pendiente de la revisión final de H30, ronda 1

Lo que el corrector de la revisión no puede arreglar, porque solo se arregla en una ruta que el workflow deshace.

## Las cifras de `legal-core` en el comentario del tope de `.github/workflows/evals.yml` (juez B, [f])

El motivo pide corregir, solo en el texto del comentario del tope del trabajo `evals` (`.github/workflows/evals.yml`,
líneas 262-263), las cifras del peor caso de `legal-core`: dice «hoy de 6 469 s, 107,8 minutos, a dos a la vez (22
tandas de 272 s), y con las cinco evals de H9, de 14 629 s, 243,8 minutos (52 tandas)», detrás de «Las cifras que
siguen son las de sus sesiones y su juez, sin la decisión» (línea 253). Sin la decisión son 19 525 s, 325,4 minutos
(70 tandas de 272 s), y con ella 20 425 s, como dicen las líneas 250-253 del mismo comentario, `peorCasoDeLegalCore` de
`TestDefinicionDelJob` y `CONTRIBUTING.md`, «Job de evals».

No se arregla aquí: `.github/` es configuración de CI, que el corrector no modifica, y `scripts/workflow/guardian-diff.sh`
la deshace en modo global (línea 78). Es solo texto de un comentario: no cambia la definición del trabajo ni ningún
tope, y `TestDefinicionDelJob` recalcula el peor caso con las evals del repositorio. Lo arregla una persona, en la
propuesta del hito o en otra, cambiando esa frase por: «El de legal-core, que no tiene juez, es hoy de 19 525 s,
325,4 minutos, a dos a la vez (70 tandas de 272 s).» El barrido ya lo dejó anotado en `gates/supuestos.md`.
