# T014 · resuelta el 2026-09-11 — esta nota queda solo como historia

**Estado actual de la tarea**: `[X]`. El registro válido es
[`gates/verificacion-pr.md`](./verificacion-pr.md); este fichero **ya no describe el estado del hito**.

## Qué decía y por qué dejó de valer

Los intentos 1, 2 y 3 (2026-09-10) dejaron T014 sin marcar por un prerrequisito humano: el repositorio no
tenía remoto en la plataforma, así que no había push, ni propuesta de cambio, ni ejecución de `ci.yml` que
medir. Marcar `[X]` habría afirmado que SC-005 y SC-006 estaban validados sin estarlo.

El 2026-09-11 el prerrequisito quedó resuelto —`origin` creado (`jmorenobl/kitlegal`) y `CODECOV_TOKEN`
dado de alta— y el escenario 11 se ejecutó entero: PR sucia bloqueada por `forbidigo: 1`, PR limpia en
verde con los cuatro controles, 1 min 00 s en caliente y 2 min 48 s en frío, ambas propuestas cerradas sin
integrar y las dos ramas borradas en local y en el remoto. Los números y las pruebas están en
`gates/verificacion-pr.md`.

## Lo que estos tres intentos sí aportaron, y sigue vigente

**Intento 1** — detectó, de paso, que el fixture que el guion del escenario 11 dictaba **no producía el
hallazgo de `forbidigo` que su propio apartado «Esperado» anunciaba**: el procesador `uniq-by-line` deja un
solo hallazgo por línea y el `exported` de `revive` desplazaba al de `forbidigo`. No pudo arreglarlo:
`quickstart.md` no figuraba entre sus rutas declaradas.

**Intento 2** — declaró `quickstart.md` y **corrigió el guion** (comentarios de paquete y de función en el
fixture, exigencia explícita de `forbidigo: 1`, y la advertencia de que no basta con que el lint falle:
hay que leer qué regla falló). Sin esa corrección, la ejecución del 2026-09-11 habría dado SC-005 por bueno
dejando FR-014 sin comprobar. El log de la plataforma confirma la corrección: `1 issues: * forbidigo: 1`.

**Intento 3** — reverificó el bloqueo sin inventar resultados ni rebajar `fail_ci_if_error`, que es lo que
mantuvo la casilla de SC-006 limpia hasta poder medirla de verdad.
