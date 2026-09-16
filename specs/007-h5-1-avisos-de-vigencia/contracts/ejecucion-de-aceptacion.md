# Contrato: la ejecución de aceptación

FR-070 a FR-073; SC-007. Decisión: research.md D17. Supuestos de plataforma: research §S (S1 a S4). Evidencia:
data-model §10. Las órdenes literales están en [quickstart.md](../quickstart.md) §11; aquí, qué hace cada paso y cuándo
vale. Lo ejecuta la tarea `[plataforma]`, la última del hito, que es la única que usa la plataforma remota.

## 1. Antes de publicar

- La rama del hito está limpia fuera de `specs/007-h5-1-avisos-de-vigencia/` y `make ci` está en verde sobre su cabeza.
- `specs/007-h5-1-avisos-de-vigencia/gates/pr-h5.1.md` existe, con la plantilla del ritual (objetivo, alcance, controles
  añadidos, decisiones, pendientes) y «Dependencias: ninguna nueva» (constitución §V). Es también el cuerpo que usaría el
  paso `publicar_rama` del workflow (`gates/pr-<hito en minúsculas>.md`).
- Prerrequisitos de la plataforma, que da de alta una persona y la tarea solo comprueba: sesión de `gh`, el secreto
  `CLAUDE_CODE_OAUTH_TOKEN` y la etiqueta `evals`. Si falta alguno, la tarea se detiene y lo anota.

## 2. Publicar

`git push -u origin 007-h5-1-avisos-de-vigencia` y, si no existe ya una propuesta de cambio de la rama, `gh pr create`
hacia `main` con el título `feat(H5.1): Avisos de vigencia en las evals` y el cuerpo de `gates/pr-h5.1.md`. Nunca se
fusiona, ni se empuja a `main`, ni se fuerza, ni se borra nada (ADR 0007).

## 3. Identificar la ejecución

- **Apertura** (el caso normal): la ejecución de aceptación es la **primera** ejecución de la rama con evento
  `pull_request` y `workflowName` `evals`, ordenadas por `createdAt`. La crea la apertura de la propuesta de cambio,
  porque la propuesta toca rutas que el job mide (research V22; supuesto S1).
- **Repetición** (§6): la primera ejecución de `evals` de la rama creada en el instante del **último** evento `labeled`
  de la etiqueta `evals` o después.

La orden se repite hasta que la ejecución aparezca, y la espera (`gh run watch --exit-status`) hasta que termine.

## 4. Leer el informe

Con la ejecución terminada, se descarga su registro (`gh run view --log`) a la carpeta temporal del quickstart, se exige
que estén las cuatro marcas de `scripts/evals.sh` y se extraen `informe.json` e `informe.md` de entre sus marcas
(research V23, V24). Si falta una marca, el job se detuvo antes del informe: la tarea registra `gh run view --log-failed`,
se detiene y lo anota; el arreglo, si es del repositorio, va en una tarea nueva antes de esta.

## 5. Comprobar la aceptación

Un único programa de `jq` sobre `informe.json` (research V26) imprime el resumen y termina con `true` solo si:

| Condición | Requisito |
|---|---|
| `veredicto` es `aprobado` | FR-072 |
| `red` es `[]` | FR-073 |
| hay exactamente una tasa de `18-lrjpac-norma-derogada.yaml` con el modelo que decide y sin pregunta ampliada, `planificada` y sin `decide`, con `sesiones` 3 | FR-071, SC-007 (la tasa sobre 3 sesiones; ADR 0016) |
| hay exactamente 3 sesiones de esa eval con ese modelo, y en cada una `avisos_encontrados` + `avisos_ausentes`, ordenados, son `derogada` y `vigencia-agotada` | FR-071 (la eval declara sus dos avisos y cada sesión publica su reparto) |

Y una orden aparte comprueba que el `commit` del informe es el `headSha` de la ejecución (supuesto S2) y que
`git diff --name-only <headSha> HEAD` solo lista ficheros bajo `specs/007-h5-1-avisos-de-vigencia/` (FR-070). Cuántas de
las 3 sesiones pasan, o cuántas trasladaron cada aviso, **no** decide la aceptación: se publica, porque la eval es
informativa.

Si el programa termina con `false` o la segunda orden falla, la aceptación no vale: la tarea registra la salida, se
detiene y lo anota. No se relanza la ejecución para buscar otro resultado.

## 6. Repetir

Si después de la ejecución de aceptación entra en la rama un commit que cambia cualquier fichero fuera de
`specs/007-h5-1-avisos-de-vigencia/` —por ejemplo, una corrección de la revisión final—, la aceptación ya no cubre la
cabeza y se repite: se quita la etiqueta `evals` si está puesta, se pone, y se identifica, espera, lee y comprueba la
ejecución nueva por el último evento `labeled` (§3), con las mismas condiciones de §5. Quien cambia algo fuera del
directorio del hito tras la aceptación deja anotado en `gates/pr-h5.1.md`, en «Pendientes», que la aceptación se repite.

## 7. Evidencia

`specs/007-h5-1-avisos-de-vigencia/gates/evals-aceptacion.md` (data-model §10): la ruta de identificación usada, el
enlace, el `databaseId`, el `headSha`, la conclusión, la duración del paso «Ejecutar las evals» (supuesto S4), la salida
entera del programa de `jq` y la salida entera de la comprobación del commit. Una repetición añade su sección con los
mismos datos y dice cuál es la vigente. La tarea no modifica ningún fichero fuera de ese directorio.
