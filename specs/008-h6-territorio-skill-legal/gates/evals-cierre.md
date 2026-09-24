# Ejecución de aceptación de H6 (quickstart §14, FR-100, FR-101, SC-013, SC-015)

## Revisión final, ronda 6 (2026-09-24): ejecución repetida, la vigente

**La ejecución de evals 36038417662, sobre `ead63d9` (`fix(H6): la guarda de declaraElNombre, con su test y su
motivo`), es la vigente.** La ronda 6 tocó `internal/evals/territorio.go` (un comentario) e
`internal/evals/territorio_test.go` (un caso), fuera de `specs/008-h6-territorio-skill-legal/` y dentro del filtro de
rutas del job, así que la 36019842457 (abajo) dejó de cubrir la cabeza y se repitió por etiqueta sobre `ead63d9`. Todo
lo que se commitea después está bajo `specs/008-h6-territorio-skill-legal/`. Las cuatro anteriores quedan registradas y
no cuentan.

| Dato | Valor |
|---|---|
| Rama empujada | `git push origin 008-h6-territorio-skill-legal`: `6d08752..ead63d9`, `guardia-push` en verde; ni `main`, ni `--force` |
| Disparo | etiqueta `evals` quitada y vuelta a poner en #39 (18:01:59Z) |
| Ejecución de evals vigente | <https://github.com/jmorenobl/kitlegal/actions/runs/36038417662>, evento `pull_request`, creada 18:02:01Z |
| Trabajos | `cambios` `skipped` (por diseño en `labeled`) · `evals (legal-core)` `success` (18:02:04Z–18:14:26Z) · `evals (boe-legislacion)` `success` (18:02:04Z–18:41:51Z) |
| Commit evaluado | `ead63d943079d4bac488ee8889de3e874f372e34`: el `headSha` de la ejecución y el `Commit:` de los dos informes |
| Modelos | decide `claude-sonnet-5`; informativo `claude-haiku-4-5-20251001`; Claude Code `2.1.270`; 3 repeticiones y umbral 2 |
| `legal-core` | `aprobado`, `Motivos: ninguno`. Las tres evals, 3 de 3 con los dos modelos; ninguna invocación fuera de lo grabado |
| `boe-legislacion` | `aprobado`, `Motivos: ninguno`. Las 18 evals, 3 de 3 con `claude-sonnet-5` (de la 01 a la 12 deciden; de la 13 a la 18, «Decide: no») y, de la 01 a la 12, 3 de 3 con el informativo. Nueve invocaciones fuera de lo grabado, todas con código 5 y ninguna llega a la red |
| Red | «ninguna petición llegó a la red de una fuente» en los dos informes |
| Sin Python | en los dos trabajos, `usuario: root` y `resultado: ninguno` |
| `ci` de la cabeza | ejecución [36038400511](https://github.com/jmorenobl/kitlegal/actions/runs/36038400511) sobre `ead63d9`, 18:01:52Z–18:09:25Z, `success` |
| Codecov sobre la cabeza | cuatro check-runs `success` con medida: `codecov/project` `96.76% (target 70.00%)`, `codecov/patch` `98.79% of diff hit (target 96.42%)`, `codecov/project/internal/cli` `98.09% (target 90.00%)`, `codecov/project/internal/core` `98.63% (target 85.00%)` |

### SC-013

**Municipio cubierto (Leganés)**: las seis sesiones pasan por `territorio resolver Leganés --json` (código 0, sin
conexiones) y el juez las da por `pasa`.

**Municipio no cubierto (Tordesillas)**: las seis pasan por `territorio resolver Tordesillas --json` (código 0, sin
conexiones), dan solo el BOE y declaran los dos aspectos `no-configurado` diciendo que no significa que no existan. Las
tres del modelo que decide se refieren a los que faltan solo por su clase y dicen que no pueden nombrarlos
(`sonnet-5-01`: «no debo nombrarlos por mi cuenta»). **Una del informativo sí nombra uno**: `haiku-4-5-20251001-02`,
«Castilla y León publica sus normas en el **Boletín Oficial de Castilla y León (BOCYL)**», con veredicto `pasa`.

**Veredicto: SC-013 se cumple con la lectura que fijó Jorge el 2026-09-24, por serie y umbral** (3 de 3 del modelo que
decide, 2 de 3 del informativo); respuesta a respuesta serían 5 de 6, como en la 36011479943.

### Orden 1 del §14, tal cual

~~~~text
[{"conclusion":"success","databaseId":36038417662,"headSha":"ead63d943079d4bac488ee8889de3e874f372e34","status":"completed","workflowName":"evals"},{"conclusion":"success","databaseId":36019842457,"headSha":"6d0875283816bb3a0eac32456048f997df471578","status":"completed","workflowName":"evals"},{"conclusion":"success","databaseId":36011479943,"headSha":"9a77c6b975130078de8794a3c8df0b45e0140bcd","status":"completed","workflowName":"evals"}]
~~~~

### Orden 2 del §14, tal cual (salida entera)

~~~~text
# Informe de evals de legal-core

## Veredicto

Veredicto: aprobado

Motivos: ninguno

## Cabecera

Modelo que decide: claude-sonnet-5

Modelos informativos: claude-haiku-4-5-20251001

Repeticiones por eval: 3

Umbral: 2

Modelos de las sesiones: claude-haiku-4-5-20251001, claude-sonnet-5

Versiones de Claude Code: 2.1.270

Commit: ead63d943079d4bac488ee8889de3e874f372e34

## Comprobación sin Python

```text
búsqueda: find / ( -path /proc -o -path /sys ) -prune -o ( ( -type f -perm /111 ( -iname python* -o -iname pypy* ) ) -o ( -type l ( -iname python* -o -iname pypy* ) ) -o ( ( -type f -o -type l ) ( -iname libpython* -o -iname libpypy* ) ) ) -print
usuario: root
resultado: ninguno
```

## Ficheros mal formados

ninguno

## Invocaciones fuera de lo grabado

ninguna

## Peticiones llegadas a la red

ninguna petición llegó a la red de una fuente

## Tasas por eval

| Eval | Modelo | Decide | Planificada | Tasa | Resultado |
| --- | --- | --- | --- | --- | --- |
| 01-territorio-municipio-cubierto.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 01-territorio-municipio-cubierto.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 02-territorio-municipio-no-cubierto.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 02-territorio-municipio-no-cubierto.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 03-no-activa-receta-de-cocina.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 03-no-activa-receta-de-cocina.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |

## Sesiones
# Informe de evals de boe-legislacion

## Veredicto

Veredicto: aprobado

Motivos: ninguno

## Cabecera

Modelo que decide: claude-sonnet-5

Modelos informativos: claude-haiku-4-5-20251001

Repeticiones por eval: 3

Umbral: 2

Modelos de las sesiones: claude-haiku-4-5-20251001, claude-sonnet-5

Versiones de Claude Code: 2.1.270

Commit: ead63d943079d4bac488ee8889de3e874f372e34

## Comprobación sin Python

```text
búsqueda: find / ( -path /proc -o -path /sys ) -prune -o ( ( -type f -perm /111 ( -iname python* -o -iname pypy* ) ) -o ( -type l ( -iname python* -o -iname pypy* ) ) -o ( ( -type f -o -type l ) ( -iname libpython* -o -iname libpypy* ) ) ) -print
usuario: root
resultado: ninguno
```

## Ficheros mal formados

ninguno

## Invocaciones fuera de lo grabado

| Sesión | Eval | Orden | Código |
| --- | --- | --- | --- |
| 06-irpf-rendimientos-del-trabajo-claude-haiku-4-5-20251001-03 | 06-irpf-rendimientos-del-trabajo.yaml | boe buscar IRPF rendimientos trabajo --json | 5 |
| 08-ltaibg-plazo-de-resolucion-claude-sonnet-5-01 | 08-ltaibg-plazo-de-resolucion.yaml | boe buscar transparencia acceso a la información pública y buen gobierno --json | 5 |
| 13-lrbrl-atribuciones-por-materia-claude-sonnet-5-02 | 13-lrbrl-atribuciones-por-materia.yaml | boe buscar Reguladora de las Bases del Régimen Local --json | 5 |
| 14-trlrhl-impuestos-por-materia-claude-sonnet-5-03 | 14-trlrhl-impuestos-por-materia.yaml | boe buscar haciendas locales --json | 5 |
| 15-irpf-rendimientos-por-materia-claude-sonnet-5-03 | 15-irpf-rendimientos-por-materia.yaml | boe buscar Impuesto sobre la Renta de las Personas Físicas --json | 5 |
| 16-lrjsp-legalidad-por-materia-claude-sonnet-5-01 | 16-lrjsp-legalidad-por-materia.yaml | boe articulos BOE-A-2015-10566 a25 a26 --json | 5 |
| 16-lrjsp-legalidad-por-materia-claude-sonnet-5-01 | 16-lrjsp-legalidad-por-materia.yaml | boe articulo BOE-A-2015-10566 a26 --json | 5 |
| 17-ltaibg-plazo-por-materia-claude-sonnet-5-01 | 17-ltaibg-plazo-por-materia.yaml | boe buscar transparencia acceso a la información pública buen gobierno --json | 5 |
| 18-lrjpac-norma-derogada-claude-sonnet-5-03 | 18-lrjpac-norma-derogada.yaml | boe buscar régimen jurídico de las administraciones públicas y del procedimiento administrativo común --json | 5 |

## Peticiones llegadas a la red

ninguna petición llegó a la red de una fuente

## Tasas por eval

| Eval | Modelo | Decide | Planificada | Tasa | Resultado |
| --- | --- | --- | --- | --- | --- |
| 01-lpac-articulo-21.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 01-lpac-articulo-21.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 02-lcsp-contrato-menor.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 02-lcsp-contrato-menor.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 03-lrbrl-atribuciones-del-pleno.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 03-lrbrl-atribuciones-del-pleno.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 04-lgt-prescripcion.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 04-lgt-prescripcion.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 05-trlrhl-impuestos-municipales.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 05-trlrhl-impuestos-municipales.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 06-irpf-rendimientos-del-trabajo.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 06-irpf-rendimientos-del-trabajo.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 07-lrjsp-principio-de-legalidad.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 07-lrjsp-principio-de-legalidad.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 08-ltaibg-plazo-de-resolucion.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 08-ltaibg-plazo-de-resolucion.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 09-constitucion-articulo-140.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 09-constitucion-articulo-140.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 10-et-vacaciones.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 10-et-vacaciones.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 11-no-activa-programacion.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 11-no-activa-programacion.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 12-no-activa-acuerdo-entre-amigos.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 12-no-activa-acuerdo-entre-amigos.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 13-lrbrl-atribuciones-por-materia.yaml | claude-sonnet-5 | no | sí | 3 de 3 | llega al umbral |
| 14-trlrhl-impuestos-por-materia.yaml | claude-sonnet-5 | no | sí | 3 de 3 | llega al umbral |
| 15-irpf-rendimientos-por-materia.yaml | claude-sonnet-5 | no | sí | 3 de 3 | llega al umbral |
| 16-lrjsp-legalidad-por-materia.yaml | claude-sonnet-5 | no | sí | 3 de 3 | llega al umbral |
| 17-ltaibg-plazo-por-materia.yaml | claude-sonnet-5 | no | sí | 3 de 3 | llega al umbral |
| 18-lrjpac-norma-derogada.yaml | claude-sonnet-5 | no | sí | 3 de 3 | llega al umbral |

## Sesiones
código: 0
~~~~

## Revisión final, ronda 2 (2026-09-24): ejecución repetida, vigente hasta la ronda 6

**La ejecución de evals 36019842457, sobre `6d08752` (`fix(H6): los motivos de la ronda 2 de la revisión final`), fue
la vigente hasta que la ronda 6 tocó `internal/evals/` (arriba).** La ronda 2 de la revisión tocó ficheros fuera de `specs/008-h6-territorio-skill-legal/` —`README.md`,
`CHANGELOG.md`, `CONTRIBUTING.md`, `docs/SOURCES.md` e `internal/skills/sincronia_test.go`—, así que la 36011479943
(abajo) dejó de cubrir la cabeza y, por la regla de SC-015, se repitió por etiqueta sobre `6d08752`. Todo lo que se
commitea después está bajo `specs/008-h6-territorio-skill-legal/`. Las tres anteriores quedan registradas y no cuentan.

| Dato | Valor |
|---|---|
| Rama empujada | `git push origin 008-h6-territorio-skill-legal`: `9a77c6b..6d08752`, `guardia-push` en verde; ni `main`, ni `--force` |
| Disparo | etiqueta `evals` quitada y vuelta a poner en #39 (15:23:35Z) |
| Ejecución de evals vigente | <https://github.com/jmorenobl/kitlegal/actions/runs/36019842457>, evento `pull_request`, creada 15:23:37Z |
| Trabajos | `cambios` `skipped` (por diseño en `labeled`) · `evals (legal-core)` `success` (15:23:40Z–15:34:47Z) · `evals (boe-legislacion)` `success` (15:23:40Z–16:01:29Z) |
| Commit evaluado | `6d0875283816bb3a0eac32456048f997df471578`: el `headSha` de la ejecución y el `Commit:` de los dos informes |
| Modelos | decide `claude-sonnet-5`; informativo `claude-haiku-4-5-20251001`; Claude Code `2.1.270`; 3 repeticiones y umbral 2 |
| `legal-core` | `aprobado`, `Motivos: ninguno`. Las tres evals, 3 de 3 con los dos modelos; ninguna invocación fuera de lo grabado |
| `boe-legislacion` | `aprobado`, `Motivos: ninguno`. Las 18 evals, 3 de 3 con `claude-sonnet-5` (de la 01 a la 12 deciden; de la 13 a la 18, «Decide: no»); con el informativo, 3 de 3 salvo la 06, 2 de 3. Catorce invocaciones fuera de lo grabado, todas con código 5 y ninguna llega a la red |
| Red | «ninguna petición llegó a la red de una fuente» en los dos informes |
| Sin Python | en los dos trabajos, `usuario: root` y `resultado: ninguno` |
| `ci` de la cabeza | ejecución [36019821690](https://github.com/jmorenobl/kitlegal/actions/runs/36019821690) sobre `6d08752`, 15:23:27Z–15:30:57Z, `success` |
| Codecov sobre la cabeza | cuatro check-runs `success` con medida: `codecov/project` `96.76% (target 70.00%)`, `codecov/patch` `98.79% of diff hit (target 96.42%)`, `codecov/project/internal/cli` `98.09% (target 90.00%)`, `codecov/project/internal/core` `98.63% (target 85.00%)` |

### SC-013

**Municipio cubierto (Leganés)**: las seis sesiones pasan por `territorio resolver Leganés --json` (código 0, sin
conexiones) y el juez las da por `pasa`.

**Municipio no cubierto (Tordesillas)**: las seis pasan por `territorio resolver Tordesillas --json` (código 0, sin
conexiones), dan Castilla y León, Valladolid, el DIR3 `L01471659` y solo el BOE, y declaran los dos aspectos
`no-configurado` diciendo que no significa que no existan. **Ninguna de las seis nombra un boletín que el applet no
devolvió**: todas se refieren a los que faltan por su clase —«el boletín de Castilla y León», «sus boletines
respectivos»—, y las tres del modelo que decide dicen que no pueden dar su nombre (`sonnet-5-01`: «no debo suplirlos
de memoria»). La sesión `haiku-4-5-20251001-01`, del informativo, afirma además que las normas municipales se publican
en el BOE «cuando son de rango estatal», una imprecisión que no nombra ningún boletín.

**Veredicto: SC-013 se cumple con las dos lecturas** (6 de 6 respuesta a respuesta; 3 de 3 y 3 de 3 por serie). La
lectura que vale, por serie y umbral, la fijó Jorge el 2026-09-24 a la vista de la ejecución 36011479943 (abajo).

### Orden 1 del §14, tal cual

~~~~text
[{"conclusion":"success","databaseId":36019842457,"headSha":"6d0875283816bb3a0eac32456048f997df471578","status":"completed","workflowName":"evals"},{"conclusion":"success","databaseId":36011479943,"headSha":"9a77c6b975130078de8794a3c8df0b45e0140bcd","status":"completed","workflowName":"evals"},{"conclusion":"success","databaseId":35722605048,"headSha":"2b811644afa2c388d27945eccdfb6e6e66fc9c32","status":"completed","workflowName":"evals"}]
~~~~

### Orden 2 del §14, tal cual (salida entera)

~~~~text
# Informe de evals de boe-legislacion

## Veredicto

Veredicto: aprobado

Motivos: ninguno

## Cabecera

Modelo que decide: claude-sonnet-5

Modelos informativos: claude-haiku-4-5-20251001

Repeticiones por eval: 3

Umbral: 2

Modelos de las sesiones: claude-haiku-4-5-20251001, claude-sonnet-5

Versiones de Claude Code: 2.1.270

Commit: 6d0875283816bb3a0eac32456048f997df471578

## Comprobación sin Python

```text
búsqueda: find / ( -path /proc -o -path /sys ) -prune -o ( ( -type f -perm /111 ( -iname python* -o -iname pypy* ) ) -o ( -type l ( -iname python* -o -iname pypy* ) ) -o ( ( -type f -o -type l ) ( -iname libpython* -o -iname libpypy* ) ) ) -print
usuario: root
resultado: ninguno
```

## Ficheros mal formados

ninguno

## Invocaciones fuera de lo grabado

| Sesión | Eval | Orden | Código |
| --- | --- | --- | --- |
| 03-lrbrl-atribuciones-del-pleno-claude-sonnet-5-03 | 03-lrbrl-atribuciones-del-pleno.yaml | boe buscar bases del régimen local --json | 5 |
| 05-trlrhl-impuestos-municipales-claude-sonnet-5-02 | 05-trlrhl-impuestos-municipales.yaml | boe buscar texto refundido Ley reguladora Haciendas Locales --json | 5 |
| 06-irpf-rendimientos-del-trabajo-claude-haiku-4-5-20251001-03 | 06-irpf-rendimientos-del-trabajo.yaml | boe buscar IRPF impuesto renta personas físicas --json | 5 |
| 06-irpf-rendimientos-del-trabajo-claude-sonnet-5-02 | 06-irpf-rendimientos-del-trabajo.yaml | boe buscar Impuesto sobre la Renta de las Personas Físicas --json | 5 |
| 08-ltaibg-plazo-de-resolucion-claude-haiku-4-5-20251001-01 | 08-ltaibg-plazo-de-resolucion.yaml | boe buscar Ley 19/2013 Transparencia Acceso Información --json | 5 |
| 08-ltaibg-plazo-de-resolucion-claude-sonnet-5-03 | 08-ltaibg-plazo-de-resolucion.yaml | boe buscar transparencia acceso a la información pública y buen gobierno --json | 5 |
| 14-trlrhl-impuestos-por-materia-claude-sonnet-5-01 | 14-trlrhl-impuestos-por-materia.yaml | boe articulos BOE-A-2004-4214 a59 a60 --json | 5 |
| 14-trlrhl-impuestos-por-materia-claude-sonnet-5-02 | 14-trlrhl-impuestos-por-materia.yaml | boe articulos BOE-A-2004-4214 a59 a60 --json | 5 |
| 14-trlrhl-impuestos-por-materia-claude-sonnet-5-02 | 14-trlrhl-impuestos-por-materia.yaml | boe articulo BOE-A-2004-4214 a60 --json | 5 |
| 14-trlrhl-impuestos-por-materia-claude-sonnet-5-03 | 14-trlrhl-impuestos-por-materia.yaml | boe articulos BOE-A-2004-4214 a59 a60 --json | 5 |
| 16-lrjsp-legalidad-por-materia-claude-sonnet-5-01 | 16-lrjsp-legalidad-por-materia.yaml | boe articulos BOE-A-2015-10566 a25 a26 a27 --json | 5 |
| 16-lrjsp-legalidad-por-materia-claude-sonnet-5-01 | 16-lrjsp-legalidad-por-materia.yaml | boe articulo BOE-A-2015-10566 a27 --json | 5 |
| 17-ltaibg-plazo-por-materia-claude-sonnet-5-01 | 17-ltaibg-plazo-por-materia.yaml | boe buscar transparencia acceso a la información pública y buen gobierno --json | 5 |
| 18-lrjpac-norma-derogada-claude-sonnet-5-01 | 18-lrjpac-norma-derogada.yaml | boe buscar régimen jurídico de las administraciones públicas y del procedimiento administrativo común --json | 5 |

## Peticiones llegadas a la red

ninguna petición llegó a la red de una fuente

## Tasas por eval

| Eval | Modelo | Decide | Planificada | Tasa | Resultado |
| --- | --- | --- | --- | --- | --- |
| 01-lpac-articulo-21.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 01-lpac-articulo-21.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 02-lcsp-contrato-menor.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 02-lcsp-contrato-menor.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 03-lrbrl-atribuciones-del-pleno.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 03-lrbrl-atribuciones-del-pleno.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 04-lgt-prescripcion.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 04-lgt-prescripcion.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 05-trlrhl-impuestos-municipales.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 05-trlrhl-impuestos-municipales.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 06-irpf-rendimientos-del-trabajo.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 06-irpf-rendimientos-del-trabajo.yaml | claude-haiku-4-5-20251001 | no | sí | 2 de 3 | llega al umbral |
| 07-lrjsp-principio-de-legalidad.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 07-lrjsp-principio-de-legalidad.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 08-ltaibg-plazo-de-resolucion.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 08-ltaibg-plazo-de-resolucion.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 09-constitucion-articulo-140.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 09-constitucion-articulo-140.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 10-et-vacaciones.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 10-et-vacaciones.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 11-no-activa-programacion.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 11-no-activa-programacion.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 12-no-activa-acuerdo-entre-amigos.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 12-no-activa-acuerdo-entre-amigos.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 13-lrbrl-atribuciones-por-materia.yaml | claude-sonnet-5 | no | sí | 3 de 3 | llega al umbral |
| 14-trlrhl-impuestos-por-materia.yaml | claude-sonnet-5 | no | sí | 3 de 3 | llega al umbral |
| 15-irpf-rendimientos-por-materia.yaml | claude-sonnet-5 | no | sí | 3 de 3 | llega al umbral |
| 16-lrjsp-legalidad-por-materia.yaml | claude-sonnet-5 | no | sí | 3 de 3 | llega al umbral |
| 17-ltaibg-plazo-por-materia.yaml | claude-sonnet-5 | no | sí | 3 de 3 | llega al umbral |
| 18-lrjpac-norma-derogada.yaml | claude-sonnet-5 | no | sí | 3 de 3 | llega al umbral |

## Sesiones
# Informe de evals de legal-core

## Veredicto

Veredicto: aprobado

Motivos: ninguno

## Cabecera

Modelo que decide: claude-sonnet-5

Modelos informativos: claude-haiku-4-5-20251001

Repeticiones por eval: 3

Umbral: 2

Modelos de las sesiones: claude-haiku-4-5-20251001, claude-sonnet-5

Versiones de Claude Code: 2.1.270

Commit: 6d0875283816bb3a0eac32456048f997df471578

## Comprobación sin Python

```text
búsqueda: find / ( -path /proc -o -path /sys ) -prune -o ( ( -type f -perm /111 ( -iname python* -o -iname pypy* ) ) -o ( -type l ( -iname python* -o -iname pypy* ) ) -o ( ( -type f -o -type l ) ( -iname libpython* -o -iname libpypy* ) ) ) -print
usuario: root
resultado: ninguno
```

## Ficheros mal formados

ninguno

## Invocaciones fuera de lo grabado

ninguna

## Peticiones llegadas a la red

ninguna petición llegó a la red de una fuente

## Tasas por eval

| Eval | Modelo | Decide | Planificada | Tasa | Resultado |
| --- | --- | --- | --- | --- | --- |
| 01-territorio-municipio-cubierto.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 01-territorio-municipio-cubierto.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 02-territorio-municipio-no-cubierto.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 02-territorio-municipio-no-cubierto.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 03-no-activa-receta-de-cocina.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 03-no-activa-receta-de-cocina.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |

## Sesiones
código: 0
~~~~

## Revisión final, ronda 1 (2026-09-24): ejecución repetida, vigente hasta la ronda 2

**La ejecución de evals 36011479943, sobre `9a77c6b` (`fix(H6): la derivación del DIR3, verificada contra DIR3
real`), fue la vigente hasta que la ronda 2 volvió a tocar ficheros fuera de `specs/` (arriba).** La 35722605048 (abajo, «Intento 3») dejó de cubrir la cabeza cuando la revisión final tocó
ficheros fuera de `specs/008-h6-territorio-skill-legal/`: `c481451` (`CHANGELOG.md`, `CONTRIBUTING.md`, `Makefile`,
`data/datos.go` y `data/datos_test.go`, los 19 ficheros de `data/territorio/comunidades/`, el e2e
`territorio-matriz.txtar` y cinco tests de `internal/core/ids` e `internal/skills`) y `9a77c6b` (`docs/SOURCES.md`).
Por la regla de SC-015 se repitió sobre la cabeza resultante, por etiqueta como H5.1: es la medida de otro commit, no
una repetición para buscar otro resultado. Queda registrada aquí y las dos anteriores no cuentan.

| Dato | Valor |
|---|---|
| Rama empujada | `git push origin 008-h6-territorio-skill-legal`: `ee967b6..9a77c6b`, `guardia-push` en verde; ni `main`, ni `--force` |
| Disparo | etiqueta `evals` quitada y vuelta a poner en #39 (14:15:32Z), porque el flujo escucha `labeled` y no `synchronize` |
| Ejecución de evals vigente | <https://github.com/jmorenobl/kitlegal/actions/runs/36011479943>, evento `pull_request`, creada 14:15:34Z |
| Trabajos | `cambios` `skipped` (por diseño en `labeled`) · `evals (legal-core)` `success` (14:15:37Z–14:24:40Z) · `evals (boe-legislacion)` `success` (14:15:37Z–15:02:21Z) |
| Commit evaluado | `9a77c6b975130078de8794a3c8df0b45e0140bcd`: el `headSha` de la ejecución y el `Commit:` de los dos informes |
| Modelos | decide `claude-sonnet-5`; informativo `claude-haiku-4-5-20251001`; Claude Code `2.1.270`; 3 repeticiones y umbral 2 |
| `legal-core` | `aprobado`, `Motivos: ninguno`. Las tres evals, 3 de 3 con los dos modelos; las 18 sesiones `pasa`; ninguna invocación fuera de lo grabado |
| `boe-legislacion` | `aprobado`, `Motivos: ninguno`. Las 18 evals, 3 de 3 con `claude-sonnet-5` —de la 01 a la 12 deciden; de la 13 a la 18 son de materia, «Decide: no»— y, de la 01 a la 12, también 3 de 3 con el informativo. Veinte invocaciones fuera de lo grabado, todas con código 5 y ninguna llega a la red, como en las ejecuciones anteriores |
| Red | «ninguna petición llegó a la red de una fuente» en los dos informes |
| Sin Python | en los dos trabajos, `usuario: root` y `resultado: ninguno` |
| `ci` de la cabeza | ejecución [36011447433](https://github.com/jmorenobl/kitlegal/actions/runs/36011447433) sobre `9a77c6b`, 14:15:19Z–14:22:25Z, `success` |
| Codecov sobre la cabeza | cuatro check-runs `success` con medida: `codecov/project` `96.74% (target 70.00%)`, `codecov/patch` `98.67% of diff hit (target 96.42%)`, `codecov/project/internal/cli` `98.09% (target 90.00%)`, `codecov/project/internal/core` `98.63% (target 85.00%)` |
| La cabeza frente al commit evaluado | lo que se commitee después de `9a77c6b` —este registro, `gates/pr-h6.md` y los veredictos de los jueces— está bajo `specs/008-h6-territorio-skill-legal/` |

### SC-013

**Municipio cubierto (Leganés)**: las seis sesiones pasan por `territorio resolver Leganés --json` (código 0, sin
conexiones) y dan Comunidad de Madrid, Madrid, el BOE y el BOCM en los dos niveles, y el DIR3 `L01280745`.

**Municipio no cubierto (Tordesillas)**: las seis pasan por `territorio resolver Tordesillas --json` (código 0, sin
conexiones), dan Castilla y León, Valladolid, el DIR3 `L01471659` y solo el BOE, y declaran los dos aspectos
`no-configurado` diciendo que no significa que no existan. Las tres del modelo que decide se refieren a los que faltan
solo por su clase —«el boletín oficial de Castilla y León ni el de la provincia de Valladolid»— y las tres añaden que
no pueden nombrarlos con estos datos (`sonnet-5-01`: «no puedo nombrarlos con estos datos»; `sonnet-5-02`: «no puedo
suplirlo de memoria»). **Una sesión del modelo
informativo sí nombra uno**: `haiku-4-5-20251001-02` termina con «el Boletín Oficial de Castilla y León y el de la
provincia de Valladolid publican sus propias normas», con mayúsculas, que es el nombre propio del boletín autonómico,
y el juez la da por `pasa` porque la cláusula no está en el esperado.

**Veredicto: SC-013 se cumple con la lectura por serie y umbral** (3 de 3 del modelo que decide, 2 de 3 del
informativo); **leída respuesta a respuesta, no** (5 de 6). Es la misma indefinición que el intento 1 dejó escrita y
que en el intento 3 no hizo falta resolver porque las dos lecturas coincidían: el spec no fija cuál vale y el juez no ve
esta cláusula. Sigue en `gates/pr-h6.md`, «Pendientes», como pieza aparte después del hito. **Jorge decidió el
2026-09-24, con esto delante, que vale la lectura por serie y umbral**, la misma que el resto de las evals.

### Orden 1 del §14, tal cual

~~~~text
[{"conclusion":"success","databaseId":36011479943,"headSha":"9a77c6b975130078de8794a3c8df0b45e0140bcd","status":"completed","workflowName":"evals"},{"conclusion":"success","databaseId":35722605048,"headSha":"2b811644afa2c388d27945eccdfb6e6e66fc9c32","status":"completed","workflowName":"evals"},{"conclusion":"success","databaseId":35714659803,"headSha":"47f3090e5288d9beccaa1188217f062f975a27d7","status":"completed","workflowName":"evals"}]
~~~~

### Orden 2 del §14, tal cual (salida entera)

~~~~text
# Informe de evals de boe-legislacion

## Veredicto

Veredicto: aprobado

Motivos: ninguno

## Cabecera

Modelo que decide: claude-sonnet-5

Modelos informativos: claude-haiku-4-5-20251001

Repeticiones por eval: 3

Umbral: 2

Modelos de las sesiones: claude-haiku-4-5-20251001, claude-sonnet-5

Versiones de Claude Code: 2.1.270

Commit: 9a77c6b975130078de8794a3c8df0b45e0140bcd

## Comprobación sin Python

```text
búsqueda: find / ( -path /proc -o -path /sys ) -prune -o ( ( -type f -perm /111 ( -iname python* -o -iname pypy* ) ) -o ( -type l ( -iname python* -o -iname pypy* ) ) -o ( ( -type f -o -type l ) ( -iname libpython* -o -iname libpypy* ) ) ) -print
usuario: root
resultado: ninguno
```

## Ficheros mal formados

ninguno

## Invocaciones fuera de lo grabado

| Sesión | Eval | Orden | Código |
| --- | --- | --- | --- |
| 06-irpf-rendimientos-del-trabajo-claude-sonnet-5-02 | 06-irpf-rendimientos-del-trabajo.yaml | boe buscar Impuesto sobre la Renta de las Personas Físicas --json | 5 |
| 08-ltaibg-plazo-de-resolucion-claude-sonnet-5-01 | 08-ltaibg-plazo-de-resolucion.yaml | boe buscar transparencia acceso a la información pública y buen gobierno --json | 5 |
| 13-lrbrl-atribuciones-por-materia-claude-sonnet-5-03 | 13-lrbrl-atribuciones-por-materia.yaml | boe articulo BOE-A-1985-5392 a123 --json | 5 |
| 14-trlrhl-impuestos-por-materia-claude-sonnet-5-01 | 14-trlrhl-impuestos-por-materia.yaml | boe articulos BOE-A-2004-4214 a59 a60 --json | 5 |
| 14-trlrhl-impuestos-por-materia-claude-sonnet-5-01 | 14-trlrhl-impuestos-por-materia.yaml | boe articulo BOE-A-2004-4214 a60 --json | 5 |
| 14-trlrhl-impuestos-por-materia-claude-sonnet-5-01 | 14-trlrhl-impuestos-por-materia.yaml | boe articulo BOE-A-2004-4214 a60 --json | 5 |
| 14-trlrhl-impuestos-por-materia-claude-sonnet-5-03 | 14-trlrhl-impuestos-por-materia.yaml | boe buscar Haciendas Locales --json | 5 |
| 14-trlrhl-impuestos-por-materia-claude-sonnet-5-03 | 14-trlrhl-impuestos-por-materia.yaml | boe articulos BOE-A-2004-4214 a59 a60 --json | 5 |
| 15-irpf-rendimientos-por-materia-claude-sonnet-5-02 | 15-irpf-rendimientos-por-materia.yaml | boe buscar Impuesto sobre la Renta de las Personas Físicas --json | 5 |
| 16-lrjsp-legalidad-por-materia-claude-sonnet-5-01 | 16-lrjsp-legalidad-por-materia.yaml | boe articulos BOE-A-2015-10566 a25 a26 --json | 5 |
| 16-lrjsp-legalidad-por-materia-claude-sonnet-5-01 | 16-lrjsp-legalidad-por-materia.yaml | boe articulo BOE-A-2015-10566 a26 --json | 5 |
| 16-lrjsp-legalidad-por-materia-claude-sonnet-5-01 | 16-lrjsp-legalidad-por-materia.yaml | boe articulo BOE-A-2015-10566 a26 --json | 5 |
| 16-lrjsp-legalidad-por-materia-claude-sonnet-5-02 | 16-lrjsp-legalidad-por-materia.yaml | boe buscar régimen jurídico del sector público --json | 5 |
| 16-lrjsp-legalidad-por-materia-claude-sonnet-5-03 | 16-lrjsp-legalidad-por-materia.yaml | boe articulo BOE-A-2015-10566 a27 --json | 5 |
| 16-lrjsp-legalidad-por-materia-claude-sonnet-5-03 | 16-lrjsp-legalidad-por-materia.yaml | boe articulo BOE-A-2015-10566 a27 --json | 5 |
| 16-lrjsp-legalidad-por-materia-claude-sonnet-5-03 | 16-lrjsp-legalidad-por-materia.yaml | boe articulo BOE-A-2015-10566 a27 --json | 5 |
| 16-lrjsp-legalidad-por-materia-claude-sonnet-5-03 | 16-lrjsp-legalidad-por-materia.yaml | boe articulo BOE-A-2015-10566 a27 --json | 5 |
| 16-lrjsp-legalidad-por-materia-claude-sonnet-5-03 | 16-lrjsp-legalidad-por-materia.yaml | boe articulo BOE-A-2015-10566 a27 --json | 5 |
| 17-ltaibg-plazo-por-materia-claude-sonnet-5-01 | 17-ltaibg-plazo-por-materia.yaml | boe buscar transparencia acceso a la información pública y buen gobierno --json | 5 |
| 18-lrjpac-norma-derogada-claude-sonnet-5-03 | 18-lrjpac-norma-derogada.yaml | boe buscar régimen jurídico de las administraciones públicas y del procedimiento administrativo común --json | 5 |

## Peticiones llegadas a la red

ninguna petición llegó a la red de una fuente

## Tasas por eval

| Eval | Modelo | Decide | Planificada | Tasa | Resultado |
| --- | --- | --- | --- | --- | --- |
| 01-lpac-articulo-21.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 01-lpac-articulo-21.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 02-lcsp-contrato-menor.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 02-lcsp-contrato-menor.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 03-lrbrl-atribuciones-del-pleno.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 03-lrbrl-atribuciones-del-pleno.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 04-lgt-prescripcion.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 04-lgt-prescripcion.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 05-trlrhl-impuestos-municipales.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 05-trlrhl-impuestos-municipales.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 06-irpf-rendimientos-del-trabajo.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 06-irpf-rendimientos-del-trabajo.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 07-lrjsp-principio-de-legalidad.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 07-lrjsp-principio-de-legalidad.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 08-ltaibg-plazo-de-resolucion.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 08-ltaibg-plazo-de-resolucion.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 09-constitucion-articulo-140.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 09-constitucion-articulo-140.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 10-et-vacaciones.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 10-et-vacaciones.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 11-no-activa-programacion.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 11-no-activa-programacion.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 12-no-activa-acuerdo-entre-amigos.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 12-no-activa-acuerdo-entre-amigos.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 13-lrbrl-atribuciones-por-materia.yaml | claude-sonnet-5 | no | sí | 3 de 3 | llega al umbral |
| 14-trlrhl-impuestos-por-materia.yaml | claude-sonnet-5 | no | sí | 3 de 3 | llega al umbral |
| 15-irpf-rendimientos-por-materia.yaml | claude-sonnet-5 | no | sí | 3 de 3 | llega al umbral |
| 16-lrjsp-legalidad-por-materia.yaml | claude-sonnet-5 | no | sí | 3 de 3 | llega al umbral |
| 17-ltaibg-plazo-por-materia.yaml | claude-sonnet-5 | no | sí | 3 de 3 | llega al umbral |
| 18-lrjpac-norma-derogada.yaml | claude-sonnet-5 | no | sí | 3 de 3 | llega al umbral |

## Sesiones
# Informe de evals de legal-core

## Veredicto

Veredicto: aprobado

Motivos: ninguno

## Cabecera

Modelo que decide: claude-sonnet-5

Modelos informativos: claude-haiku-4-5-20251001

Repeticiones por eval: 3

Umbral: 2

Modelos de las sesiones: claude-haiku-4-5-20251001, claude-sonnet-5

Versiones de Claude Code: 2.1.270

Commit: 9a77c6b975130078de8794a3c8df0b45e0140bcd

## Comprobación sin Python

```text
búsqueda: find / ( -path /proc -o -path /sys ) -prune -o ( ( -type f -perm /111 ( -iname python* -o -iname pypy* ) ) -o ( -type l ( -iname python* -o -iname pypy* ) ) -o ( ( -type f -o -type l ) ( -iname libpython* -o -iname libpypy* ) ) ) -print
usuario: root
resultado: ninguno
```

## Ficheros mal formados

ninguno

## Invocaciones fuera de lo grabado

ninguna

## Peticiones llegadas a la red

ninguna petición llegó a la red de una fuente

## Tasas por eval

| Eval | Modelo | Decide | Planificada | Tasa | Resultado |
| --- | --- | --- | --- | --- | --- |
| 01-territorio-municipio-cubierto.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 01-territorio-municipio-cubierto.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 02-territorio-municipio-no-cubierto.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 02-territorio-municipio-no-cubierto.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 03-no-activa-receta-de-cocina.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 03-no-activa-receta-de-cocina.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |

## Sesiones
código: 0
~~~~

## Intento 3 de T026 (2026-09-22): la vigente hasta la revisión final

**La ejecución de evals 35722605048, sobre `2b81164` (`feat(H6): T029`), cierra la aceptación del hito.** Es la
segunda ejecución de evals de la rama y la vigente por la regla de SC-015: la primera, 35714659803 sobre `47f3090`,
dejó de cubrir la cabeza cuando T029 tocó `internal/core/territorio/` (registrada abajo, «Intento 1»; no cuenta). Esta
la abrió la etiqueta `evals` puesta en #39 por el intento 2 (11:38:45Z), porque el flujo no escucha `synchronize`, y
mide el primer commit de la rama con el arreglo de T029: no es una repetición para buscar otro resultado, es la
medida de otro commit. Desde ese commit la cabeza solo ha cambiado bajo `specs/008-h6-territorio-skill-legal/`. Sus
dos trabajos de la matriz salieron `aprobado`, `ci` está en verde sobre el commit evaluado y sobre la cabeza, y los
cuatro estados de Codecov, en verde con medida. No se fusionó nada, no se empujó a `main`, no se forzó nada ni se
borró ninguna rama, etiqueta o release. La ejecución no se relanzó.

| Dato | Valor |
|---|---|
| Propuesta de cambio | [#39](https://github.com/jmorenobl/kitlegal/pull/39), `OPEN`, base `main`, cabeza `008-h6-territorio-skill-legal` (`headRefOid` `ee967b6` tras el push de este intento), título `` feat(H6): `territorio` + skill `legal-core` v0 ``, `MERGEABLE`, etiqueta `evals`. Ya existía (la abrió el intento 1), así que no se creó otra. Cuerpo sincronizado con `gates/pr-h6.md` por `gh pr edit 39 --body-file` en este intento (dos veces: al empezar, porque el intento 2 no lo había hecho y #39 seguía con el cuerpo de T025, y al terminar con la aceptación registrada) |
| Rama empujada | `git push -u origin 008-h6-territorio-skill-legal`: `2b81164..ee967b6`, `guardia-push` en verde. Antes del push la cabeza remota era `2b81164` y la local iba un commit por delante, el del intento 2, solo con ficheros bajo `specs/008-h6-territorio-skill-legal/` |
| Ejecución de evals vigente | <https://github.com/jmorenobl/kitlegal/actions/runs/35722605048> (`databaseId` 35722605048). Evento `pull_request`, actividad `labeled` (etiqueta `evals`, `jmorenobl`, 11:38:45Z); creada 11:38:47Z y terminada 12:19:02Z, `success` |
| Trabajos | `cambios` 106728729745 `skipped` (11:38:48Z; el job lo salta en la actividad `labeled` por diseño) · `evals (boe-legislacion)` 106728728974 `success` (11:38:50Z–12:19:01Z) · `evals (legal-core)` 106728729164 `success` (11:38:50Z–11:47:13Z). Ningún paso fuera de `success` o `skipped` |
| Commit evaluado | `2b811644afa2c388d27945eccdfb6e6e66fc9c32` (`feat(H6): T029`). Coinciden el `headSha` de la ejecución, el `Commit:` de los dos informes, el `commit` de sus dos `informe.json` y la cabeza remota cuando se puso la etiqueta |
| El commit es de la rama | `git branch -r --contains 2b81164…` devuelve solo `origin/008-h6-territorio-skill-legal` |
| Ficheros que difieren entre el commit evaluado y la cabeza | al leer los informes, la cabeza es `ee967b6` y `git diff --name-only 2b81164 HEAD` da cuatro ficheros, todos bajo `specs/008-h6-territorio-skill-legal/`: `gates/pr-h6.md`, `gates/tarea-actual.json`, `gates/tareas-intentos.json` y `quickstart.md`; `git diff --stat 2b81164 HEAD -- . ':(exclude)specs/008-h6-territorio-skill-legal/'` no imprime nada. El commit de este intento añade solo ficheros bajo ese directorio |
| Modelos | decide `claude-sonnet-5`; informativo `claude-haiku-4-5-20251001`; Claude Code `2.1.270`; 3 repeticiones y umbral 2 |
| `legal-core` | `aprobado`, `Motivos: ninguno`. Con el modelo que decide, `01-territorio-municipio-cubierto` 3 de 3, `02-territorio-municipio-no-cubierto` 3 de 3 y `03-no-activa-receta-de-cocina` 3 de 3; con el informativo, 3 de 3, **2 de 3** y 3 de 3 (la sesión `haiku-4-5-20251001-02` de la 02 escribe cada clave de la cobertura entre acentos graves y separada del valor, y el juez no encuentra la forma fija; informativo, no decide, y llega al umbral). Las 18 sesiones terminaron solas con código 0. Ninguna invocación fuera de lo grabado. Ningún fichero mal formado |
| `boe-legislacion` | `aprobado`, `Motivos: ninguno`. Las 18 evals llegan al umbral: de la 01 a la 12 deciden y dan 3 de 3 con `claude-sonnet-5`; de la 13 a la 18 son de materia (H5.1, «Decide: no») y dan 3 de 3. Con el informativo, la 06 da 2 de 3 (`haiku-4-5-20251001-01`: comando ausente `bloque boe BOE-A-2006-20764 a17` y su cita ausente) y el resto 3 de 3. 90 sesiones, 89 `pasa`. Quince invocaciones fuera de lo grabado, todas terminan en 5 y ninguna llega a la red |
| Ficheros de eval de `boe-legislacion` | sin modificar: `git diff --stat origin/main -- evals/boe-legislacion` imprime solo `fin del diff` |
| Red | «ninguna petición llegó a la red de una fuente» en los dos informes; `red: []` en los dos `informe.json`; 0 `llegadas_a_la_red` sumadas en las 18 sesiones de `legal-core` y en las 90 de `boe-legislacion` |
| Sin Python | en los dos trabajos, `usuario: root` y `resultado: ninguno`; la retirada terminó con «búsqueda tras retirar: ninguno» |
| `ci` del commit evaluado | ejecución [35721902798](https://github.com/jmorenobl/kitlegal/actions/runs/35721902798), `pull_request`, 11:31:14Z–11:38:12Z, `success`, sin ningún paso fuera de verde |
| `ci` de la cabeza | ejecución [35724203328](https://github.com/jmorenobl/kitlegal/actions/runs/35724203328) sobre `ee967b6`, `pull_request`, 11:55:47Z–12:03:10Z, `success`, sin ningún paso fuera de verde |
| Codecov sobre la cabeza | cuatro check-runs `success` con medida, no sobre cero ficheros: `codecov/project` `96.47% (target 70.00%)`, `codecov/patch` `96.73% of diff hit (target 96.42%)`, `codecov/project/internal/cli` `98.09% (target 90.00%)`, `codecov/project/internal/core` `98.63% (target 85.00%)` (12:04:21Z–12:04:23Z) |
| Estados de #39 (`gh pr checks 39`) | `ci` `pass` y los cuatro de Codecov `pass`; la ejecución de evals no aparece porque los estados son los de la cabeza `ee967b6` y ella mide `2b81164` |

### Comprobaciones de T026

| Comprobación | Resultado |
|---|---|
| Prerrequisitos: sesión de `gh` | ✓ `Logged in to github.com account jmorenobl (keyring)`, alcance `repo` |
| Prerrequisitos: secreto del token | ✓ `CLAUDE_CODE_OAUTH_TOKEN` (2026-09-14T09:17:03Z) |
| Prerrequisitos: etiqueta del job | ✓ `evals` y `evals-prueba-de-red` existen |
| Prerrequisitos: cuerpo | ✓ `gates/pr-h6.md` presente, con «Dependencias: ninguna nueva» y el motivo del módulo que pasa a enlazarse (constitución §V) |
| Árbol limpio fuera del directorio del hito | ✓ `git status --porcelain` filtrado imprime solo `fin del estado` |
| Rama empujada | ✓ `2b81164..ee967b6` |
| Propuesta de cambio hacia `main` | ✓ #39 ya existía; no se creó otra |
| Ejecución de evals vigente, disparada por la propuesta de cambio | ✓ 35722605048, `pull_request` (`labeled`), sobre `2b81164`, la última de la rama y la primera con el arreglo de T029 |
| Orden 1 del §14 tal cual | ✓ dos ejecuciones, la vigente `success` sobre `2b81164` (§ «Salidas») |
| Orden 2 del §14 tal cual | ✓ imprime los dos informes hasta `## Sesiones` y termina con `código: 0` (§ «Salidas») |
| Las tres evals de `legal-core` en esa misma ejecución, ≥ 2 de 3 cada una | ✓ 3 de 3 las tres con el modelo que decide; el informe lleva el commit evaluado y el id del modelo |
| `boe-legislacion` sigue pasando sin modificar sus evals | ✓ `aprobado` con la regla de H5 y H5.1; diff contra `origin/main` vacío |
| Ninguna petición a la red de una fuente | ✓ en los dos informes y en los dos `informe.json` |
| El commit del informe es de la rama y la cabeza solo difiere en `specs/008-h6-territorio-skill-legal/` | ✓ (tabla de arriba) |
| La pregunta de la aceptación se responde para el municipio cubierto | ✓ las seis sesiones de Leganés pasan por `territorio resolver Leganés --json` (código 0, sin conexiones) y dan comunidad, provincia, BOE y BOCM en sus dos niveles con el motivo, y la cobertura `configurado`/`configurado`/`verificado` |
| … y para el no cubierto, con la cobertura explícita | ✓ las seis sesiones de Tordesillas pasan por `territorio resolver Tordesillas --json` (código 0, sin conexiones), dan Castilla y León, Valladolid y solo el BOE, declaran `boletin_autonomico: no-configurado` y `boletin_provincial: no-configurado` y dicen que eso no significa que no existan |
| … sin nombrar ningún boletín que el applet no haya devuelto (SC-013, quickstart §14) | ✓ en 6 de 6, y 3 de 3 del modelo que decide (§ «SC-013») |
| `ci` sobre el commit evaluado y sobre la cabeza | ✓ 35721902798 y 35724203328, `success` |
| Estados de la propuesta de cambio (`gh pr checks`) | ✓ `ci` y los cuatro de Codecov en `pass`, con medida |
| Nada prohibido | ✓ ni merge, ni push a `main`, ni `--force`, ni borrado de ramas, etiquetas o releases; la ejecución no se relanzó |
| Verificación determinista local (`make ci`, en primer plano, sobre el árbol con este registro) | ✓ código 0 y última línea `ci: todos los controles en verde` |

### SC-013

**Municipio cubierto (Leganés)**: en las seis sesiones, la respuesta pasa por `territorio resolver Leganés --json`
(código 0, sin conexiones) y da Comunidad de Madrid, Madrid, el BOE (estatal) y el BOCM (autonómico y provincial, con
su motivo de comunidad uniprovincial), el DIR3 `L01280745`, el código INE `28074` con su dígito `5`, la cobertura
`configurado`/`configurado`/`verificado` y la fecha de los datos, 2026-02-04.

**Municipio no cubierto (Tordesillas)**: en las seis sesiones, la respuesta pasa por
`territorio resolver Tordesillas --json` (código 0, sin conexiones), da Castilla y León, Valladolid, el DIR3
`L01471659` y solo el BOE, y declara los dos aspectos `no-configurado` diciendo con sus palabras que kitlegal no tiene
declarado ese boletín y que eso no significa que no exista. **Ninguna de las seis nombra un boletín que el applet no
devolvió**: las seis se refieren a los que faltan solo por su clase —«el boletín oficial de Castilla y León», «el
boletín oficial de la provincia de Valladolid», «el boletín de Valladolid»— sin nombre, sigla ni dirección, y tres
dicen además que no pueden dar sus nombres a partir de esta consulta (`sonnet-5-01`: «no puedo darte sus nombres ni
direcciones a partir de esta consulta»; `sonnet-5-03`: «no puedo darte sus nombres o direcciones exactas»). La sesión
`haiku-4-5-20251001-01`, del modelo informativo, añade que «la consulta con `boe-legislacion` te mostraría dónde se
publican», una remisión que la skill no promete, pero no nombra nada. La sesión `haiku-4-5-20251001-02`, la que el juez
da por `no pasa`, también declara los dos aspectos y no nombra nada: lo que le falta es la forma fija de la cobertura.

**Veredicto de este intento: SC-013 se cumple en la ejecución vigente con cualquiera de las dos lecturas.** El
intento 1 dejó sin decidir si la segunda cláusula se cuenta respuesta a respuesta o por serie y umbral como SC-015,
porque en aquella ejecución una sesión de seis nombraba «el BOCyL» y el «BOP de Valladolid» y las dos lecturas daban
resultados distintos. En esta, 6 de 6 y 3 de 3 del modelo que decide: coinciden, y no ha hecho falta elegir. La
lectura sigue sin fijarse en el spec y el juez sigue sin ver esta cláusula (el esperado de territorio de FR-084 declara
lo que tiene que aparecer, no lo que no puede aparecer); las dos cosas quedan en `gates/pr-h6.md`, «Pendientes», como
pieza aparte después del hito, porque cambian FR-084, `schemas/eval.yaml.json` e `internal/evals`.

## Salidas del intento 3

### Orden 1 del §14, tal cual

~~~~text
[{"conclusion":"success","databaseId":35722605048,"headSha":"2b811644afa2c388d27945eccdfb6e6e66fc9c32","status":"completed","workflowName":"evals"},{"conclusion":"success","databaseId":35714659803,"headSha":"47f3090e5288d9beccaa1188217f062f975a27d7","status":"completed","workflowName":"evals"}]
~~~~

### Orden 2 del §14, tal cual (salida entera)

Ejecutada al terminar la ejecución 35722605048, la última de la rama, que es la que la orden elige. Los dos informes
hasta `## Sesiones` y el código, tal como los imprime:

~~~~text
# Informe de evals de boe-legislacion

## Veredicto

Veredicto: aprobado

Motivos: ninguno

## Cabecera

Modelo que decide: claude-sonnet-5

Modelos informativos: claude-haiku-4-5-20251001

Repeticiones por eval: 3

Umbral: 2

Modelos de las sesiones: claude-haiku-4-5-20251001, claude-sonnet-5

Versiones de Claude Code: 2.1.270

Commit: 2b811644afa2c388d27945eccdfb6e6e66fc9c32

## Comprobación sin Python

```text
búsqueda: find / ( -path /proc -o -path /sys ) -prune -o ( ( -type f -perm /111 ( -iname python* -o -iname pypy* ) ) -o ( -type l ( -iname python* -o -iname pypy* ) ) -o ( ( -type f -o -type l ) ( -iname libpython* -o -iname libpypy* ) ) ) -print
usuario: root
resultado: ninguno
```

## Ficheros mal formados

ninguno

## Invocaciones fuera de lo grabado

| Sesión | Eval | Orden | Código |
| --- | --- | --- | --- |
| 03-lrbrl-atribuciones-del-pleno-claude-sonnet-5-03 | 03-lrbrl-atribuciones-del-pleno.yaml | boe buscar Reguladora de las Bases del Régimen Local --json | 5 |
| 05-trlrhl-impuestos-municipales-claude-sonnet-5-01 | 05-trlrhl-impuestos-municipales.yaml | boe buscar texto refundido de la Ley Reguladora de las Haciendas Locales --json | 5 |
| 06-irpf-rendimientos-del-trabajo-claude-haiku-4-5-20251001-01 | 06-irpf-rendimientos-del-trabajo.yaml | boe buscar IRPF impuesto renta personas físicas --json | 5 |
| 06-irpf-rendimientos-del-trabajo-claude-sonnet-5-01 | 06-irpf-rendimientos-del-trabajo.yaml | boe buscar Impuesto sobre la Renta de las Personas Físicas --json | 5 |
| 10-et-vacaciones-claude-sonnet-5-02 | 10-et-vacaciones.yaml | boe buscar Estatuto de los Trabajadores --json | 5 |
| 14-trlrhl-impuestos-por-materia-claude-sonnet-5-02 | 14-trlrhl-impuestos-por-materia.yaml | boe articulos BOE-A-2004-4214 a59 a60 a61 --json | 5 |
| 14-trlrhl-impuestos-por-materia-claude-sonnet-5-03 | 14-trlrhl-impuestos-por-materia.yaml | boe buscar haciendas locales --json | 5 |
| 14-trlrhl-impuestos-por-materia-claude-sonnet-5-03 | 14-trlrhl-impuestos-por-materia.yaml | boe articulos BOE-A-2004-4214 a59 a60 a61 --json | 5 |
| 15-irpf-rendimientos-por-materia-claude-sonnet-5-02 | 15-irpf-rendimientos-por-materia.yaml | boe buscar Impuesto sobre la Renta de las Personas Físicas --json | 5 |
| 16-lrjsp-legalidad-por-materia-claude-sonnet-5-01 | 16-lrjsp-legalidad-por-materia.yaml | boe articulos BOE-A-2015-10566 a25 a26 --json | 5 |
| 16-lrjsp-legalidad-por-materia-claude-sonnet-5-02 | 16-lrjsp-legalidad-por-materia.yaml | boe articulos BOE-A-2015-10566 a25 a26 a27 --json | 5 |
| 16-lrjsp-legalidad-por-materia-claude-sonnet-5-02 | 16-lrjsp-legalidad-por-materia.yaml | boe articulo BOE-A-2015-10566 a26 --json | 5 |
| 16-lrjsp-legalidad-por-materia-claude-sonnet-5-02 | 16-lrjsp-legalidad-por-materia.yaml | boe articulo BOE-A-2015-10566 a26 --json | 5 |
| 16-lrjsp-legalidad-por-materia-claude-sonnet-5-02 | 16-lrjsp-legalidad-por-materia.yaml | boe articulo BOE-A-2015-10566 a26 --json | 5 |
| 18-lrjpac-norma-derogada-claude-sonnet-5-02 | 18-lrjpac-norma-derogada.yaml | boe analisis BOE-A-1992-26318 --json | 5 |

## Peticiones llegadas a la red

ninguna petición llegó a la red de una fuente

## Tasas por eval

| Eval | Modelo | Decide | Planificada | Tasa | Resultado |
| --- | --- | --- | --- | --- | --- |
| 01-lpac-articulo-21.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 01-lpac-articulo-21.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 02-lcsp-contrato-menor.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 02-lcsp-contrato-menor.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 03-lrbrl-atribuciones-del-pleno.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 03-lrbrl-atribuciones-del-pleno.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 04-lgt-prescripcion.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 04-lgt-prescripcion.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 05-trlrhl-impuestos-municipales.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 05-trlrhl-impuestos-municipales.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 06-irpf-rendimientos-del-trabajo.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 06-irpf-rendimientos-del-trabajo.yaml | claude-haiku-4-5-20251001 | no | sí | 2 de 3 | llega al umbral |
| 07-lrjsp-principio-de-legalidad.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 07-lrjsp-principio-de-legalidad.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 08-ltaibg-plazo-de-resolucion.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 08-ltaibg-plazo-de-resolucion.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 09-constitucion-articulo-140.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 09-constitucion-articulo-140.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 10-et-vacaciones.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 10-et-vacaciones.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 11-no-activa-programacion.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 11-no-activa-programacion.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 12-no-activa-acuerdo-entre-amigos.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 12-no-activa-acuerdo-entre-amigos.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 13-lrbrl-atribuciones-por-materia.yaml | claude-sonnet-5 | no | sí | 3 de 3 | llega al umbral |
| 14-trlrhl-impuestos-por-materia.yaml | claude-sonnet-5 | no | sí | 3 de 3 | llega al umbral |
| 15-irpf-rendimientos-por-materia.yaml | claude-sonnet-5 | no | sí | 3 de 3 | llega al umbral |
| 16-lrjsp-legalidad-por-materia.yaml | claude-sonnet-5 | no | sí | 3 de 3 | llega al umbral |
| 17-ltaibg-plazo-por-materia.yaml | claude-sonnet-5 | no | sí | 3 de 3 | llega al umbral |
| 18-lrjpac-norma-derogada.yaml | claude-sonnet-5 | no | sí | 3 de 3 | llega al umbral |

## Sesiones
# Informe de evals de legal-core

## Veredicto

Veredicto: aprobado

Motivos: ninguno

## Cabecera

Modelo que decide: claude-sonnet-5

Modelos informativos: claude-haiku-4-5-20251001

Repeticiones por eval: 3

Umbral: 2

Modelos de las sesiones: claude-haiku-4-5-20251001, claude-sonnet-5

Versiones de Claude Code: 2.1.270

Commit: 2b811644afa2c388d27945eccdfb6e6e66fc9c32

## Comprobación sin Python

```text
búsqueda: find / ( -path /proc -o -path /sys ) -prune -o ( ( -type f -perm /111 ( -iname python* -o -iname pypy* ) ) -o ( -type l ( -iname python* -o -iname pypy* ) ) -o ( ( -type f -o -type l ) ( -iname libpython* -o -iname libpypy* ) ) ) -print
usuario: root
resultado: ninguno
```

## Ficheros mal formados

ninguno

## Invocaciones fuera de lo grabado

ninguna

## Peticiones llegadas a la red

ninguna petición llegó a la red de una fuente

## Tasas por eval

| Eval | Modelo | Decide | Planificada | Tasa | Resultado |
| --- | --- | --- | --- | --- | --- |
| 01-territorio-municipio-cubierto.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 01-territorio-municipio-cubierto.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 02-territorio-municipio-no-cubierto.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 02-territorio-municipio-no-cubierto.yaml | claude-haiku-4-5-20251001 | no | sí | 2 de 3 | llega al umbral |
| 03-no-activa-receta-de-cocina.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 03-no-activa-receta-de-cocina.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |

## Sesiones
código: 0
~~~~

### Resumen de los dos `informe.json`

Del paso «Ejecutar las evals» de cada trabajo, entre `--- inicio de informe.json ---` y `--- fin de informe.json ---`,
los campos de la raíz que no son las sesiones (`jq -c '{skill,commit,veredicto,motivos,modelo_que_decide,modelos_informativos,repeticiones,umbral,red}'`).
En `legal-core`, 18 objetos con `llegadas_a_la_red` y 0 llegadas sumadas; en `boe-legislacion`, 90 y 0.

~~~~text
{"skill":"legal-core","commit":"2b811644afa2c388d27945eccdfb6e6e66fc9c32","veredicto":"aprobado","motivos":[],"modelo_que_decide":"claude-sonnet-5","modelos_informativos":["claude-haiku-4-5-20251001"],"repeticiones":3,"umbral":2,"red":[]}
{"skill":"boe-legislacion","commit":"2b811644afa2c388d27945eccdfb6e6e66fc9c32","veredicto":"aprobado","motivos":[],"modelo_que_decide":"claude-sonnet-5","modelos_informativos":["claude-haiku-4-5-20251001"],"repeticiones":3,"umbral":2,"red":[]}
~~~~

### Informe entero de `legal-core` (`informe.md`, con sus 18 sesiones)

Del paso «Ejecutar las evals» del trabajo `evals (legal-core)` 106728729164, entre `--- inicio de informe.md ---` y
`--- fin de informe.md ---`, sin el prefijo de `gh run view --log`. Descargado dos veces, por la API de trabajos
mientras la ejecución seguía en marcha y con `gh run view --job … --log` al terminar: los dos textos son idénticos
(`cmp`).

~~~~markdown
# Informe de evals de legal-core

## Veredicto

Veredicto: aprobado

Motivos: ninguno

## Cabecera

Modelo que decide: claude-sonnet-5

Modelos informativos: claude-haiku-4-5-20251001

Repeticiones por eval: 3

Umbral: 2

Modelos de las sesiones: claude-haiku-4-5-20251001, claude-sonnet-5

Versiones de Claude Code: 2.1.270

Commit: 2b811644afa2c388d27945eccdfb6e6e66fc9c32

## Comprobación sin Python

```text
búsqueda: find / ( -path /proc -o -path /sys ) -prune -o ( ( -type f -perm /111 ( -iname python* -o -iname pypy* ) ) -o ( -type l ( -iname python* -o -iname pypy* ) ) -o ( ( -type f -o -type l ) ( -iname libpython* -o -iname libpypy* ) ) ) -print
usuario: root
resultado: ninguno
```

## Ficheros mal formados

ninguno

## Invocaciones fuera de lo grabado

ninguna

## Peticiones llegadas a la red

ninguna petición llegó a la red de una fuente

## Tasas por eval

| Eval | Modelo | Decide | Planificada | Tasa | Resultado |
| --- | --- | --- | --- | --- | --- |
| 01-territorio-municipio-cubierto.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 01-territorio-municipio-cubierto.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 02-territorio-municipio-no-cubierto.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 02-territorio-municipio-no-cubierto.yaml | claude-haiku-4-5-20251001 | no | sí | 2 de 3 | llega al umbral |
| 03-no-activa-receta-de-cocina.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 03-no-activa-receta-de-cocina.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |

## Sesiones

| Sesión | Eval | Modelo | Activa | Activada | Sesión terminada | Comandos ausentes | Citas ausentes | Avisos encontrados | Avisos ausentes | Territorio encontrado | Territorio ausente | Resultado |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 01-territorio-municipio-cubierto-claude-haiku-4-5-20251001-01 | 01-territorio-municipio-cubierto.yaml | claude-haiku-4-5-20251001 | sí | sí | sí (código 0) | ninguno | ninguna | ninguno | ninguno | comunidad: Comunidad de Madrid, provincia: Madrid, boletín: BOCM | ninguno | pasa |
| 01-territorio-municipio-cubierto-claude-haiku-4-5-20251001-02 | 01-territorio-municipio-cubierto.yaml | claude-haiku-4-5-20251001 | sí | sí | sí (código 0) | ninguno | ninguna | ninguno | ninguno | comunidad: Comunidad de Madrid, provincia: Madrid, boletín: BOCM | ninguno | pasa |
| 01-territorio-municipio-cubierto-claude-haiku-4-5-20251001-03 | 01-territorio-municipio-cubierto.yaml | claude-haiku-4-5-20251001 | sí | sí | sí (código 0) | ninguno | ninguna | ninguno | ninguno | comunidad: Comunidad de Madrid, provincia: Madrid, boletín: BOCM | ninguno | pasa |
| 01-territorio-municipio-cubierto-claude-sonnet-5-01 | 01-territorio-municipio-cubierto.yaml | claude-sonnet-5 | sí | sí | sí (código 0) | ninguno | ninguna | ninguno | ninguno | comunidad: Comunidad de Madrid, provincia: Madrid, boletín: BOCM | ninguno | pasa |
| 01-territorio-municipio-cubierto-claude-sonnet-5-02 | 01-territorio-municipio-cubierto.yaml | claude-sonnet-5 | sí | sí | sí (código 0) | ninguno | ninguna | ninguno | ninguno | comunidad: Comunidad de Madrid, provincia: Madrid, boletín: BOCM | ninguno | pasa |
| 01-territorio-municipio-cubierto-claude-sonnet-5-03 | 01-territorio-municipio-cubierto.yaml | claude-sonnet-5 | sí | sí | sí (código 0) | ninguno | ninguna | ninguno | ninguno | comunidad: Comunidad de Madrid, provincia: Madrid, boletín: BOCM | ninguno | pasa |
| 02-territorio-municipio-no-cubierto-claude-haiku-4-5-20251001-01 | 02-territorio-municipio-no-cubierto.yaml | claude-haiku-4-5-20251001 | sí | sí | sí (código 0) | ninguno | ninguna | ninguno | ninguno | comunidad: Castilla y León, provincia: Valladolid, boletin_autonomico: no-configurado, boletin_provincial: no-configurado | ninguno | pasa |
| 02-territorio-municipio-no-cubierto-claude-haiku-4-5-20251001-02 | 02-territorio-municipio-no-cubierto.yaml | claude-haiku-4-5-20251001 | sí | sí | sí (código 0) | ninguno | ninguna | ninguno | ninguno | comunidad: Castilla y León, provincia: Valladolid | boletin_autonomico: no-configurado, boletin_provincial: no-configurado | no pasa |
| 02-territorio-municipio-no-cubierto-claude-haiku-4-5-20251001-03 | 02-territorio-municipio-no-cubierto.yaml | claude-haiku-4-5-20251001 | sí | sí | sí (código 0) | ninguno | ninguna | ninguno | ninguno | comunidad: Castilla y León, provincia: Valladolid, boletin_autonomico: no-configurado, boletin_provincial: no-configurado | ninguno | pasa |
| 02-territorio-municipio-no-cubierto-claude-sonnet-5-01 | 02-territorio-municipio-no-cubierto.yaml | claude-sonnet-5 | sí | sí | sí (código 0) | ninguno | ninguna | ninguno | ninguno | comunidad: Castilla y León, provincia: Valladolid, boletin_autonomico: no-configurado, boletin_provincial: no-configurado | ninguno | pasa |
| 02-territorio-municipio-no-cubierto-claude-sonnet-5-02 | 02-territorio-municipio-no-cubierto.yaml | claude-sonnet-5 | sí | sí | sí (código 0) | ninguno | ninguna | ninguno | ninguno | comunidad: Castilla y León, provincia: Valladolid, boletin_autonomico: no-configurado, boletin_provincial: no-configurado | ninguno | pasa |
| 02-territorio-municipio-no-cubierto-claude-sonnet-5-03 | 02-territorio-municipio-no-cubierto.yaml | claude-sonnet-5 | sí | sí | sí (código 0) | ninguno | ninguna | ninguno | ninguno | comunidad: Castilla y León, provincia: Valladolid, boletin_autonomico: no-configurado, boletin_provincial: no-configurado | ninguno | pasa |
| 03-no-activa-receta-de-cocina-claude-haiku-4-5-20251001-01 | 03-no-activa-receta-de-cocina.yaml | claude-haiku-4-5-20251001 | no | no | sí (código 0) | ninguno | ninguna | ninguno | ninguno | ninguno | ninguno | pasa |
| 03-no-activa-receta-de-cocina-claude-haiku-4-5-20251001-02 | 03-no-activa-receta-de-cocina.yaml | claude-haiku-4-5-20251001 | no | no | sí (código 0) | ninguno | ninguna | ninguno | ninguno | ninguno | ninguno | pasa |
| 03-no-activa-receta-de-cocina-claude-haiku-4-5-20251001-03 | 03-no-activa-receta-de-cocina.yaml | claude-haiku-4-5-20251001 | no | no | sí (código 0) | ninguno | ninguna | ninguno | ninguno | ninguno | ninguno | pasa |
| 03-no-activa-receta-de-cocina-claude-sonnet-5-01 | 03-no-activa-receta-de-cocina.yaml | claude-sonnet-5 | no | no | sí (código 0) | ninguno | ninguna | ninguno | ninguno | ninguno | ninguno | pasa |
| 03-no-activa-receta-de-cocina-claude-sonnet-5-02 | 03-no-activa-receta-de-cocina.yaml | claude-sonnet-5 | no | no | sí (código 0) | ninguno | ninguna | ninguno | ninguno | ninguno | ninguno | pasa |
| 03-no-activa-receta-de-cocina-claude-sonnet-5-03 | 03-no-activa-receta-de-cocina.yaml | claude-sonnet-5 | no | no | sí (código 0) | ninguno | ninguna | ninguno | ninguno | ninguno | ninguno | pasa |

## Sesión 01-territorio-municipio-cubierto-claude-haiku-4-5-20251001-01

Eval: 01-territorio-municipio-cubierto.yaml

Modelo pedido: claude-haiku-4-5-20251001

Modelo de la sesión: claude-haiku-4-5-20251001

Pregunta:

```text
¿Qué comunidad, provincia y boletines corresponden a mi ayuntamiento, el de Leganés?
```

Invocaciones:

| Orden | Código | Conexiones |
| --- | --- | --- |
| territorio resolver Leganés --json | 0 | sin conexiones |

Respuesta:

```text
**Municipio:** Leganés, código INE 28074 (dígito de control 5)  
**Provincia:** Madrid  
**Comunidad autónoma:** Comunidad de Madrid, régimen común  
**DIR3 del ayuntamiento:** L01280745  

**Boletines:**
- BOE — Boletín Oficial del Estado (nivel estatal)
- BOCM — Boletín Oficial de la Comunidad de Madrid (nivel autonómico)
- BOCM — Boletín Oficial de la Comunidad de Madrid (nivel provincial) — *Comunidad uniprovincial: el BOCM hace también de boletín provincial.*

**Cobertura:**
- boletin_autonomico: configurado
- boletin_provincial: configurado
- dir3: verificado

*Datos de 2026-02-04*

En resumen: Leganés es un municipio de régimen común en la **Comunidad de Madrid** (provincia de Madrid). Las normas estatales se publican en el **BOE**, y las autonómicas y provinciales, en el **BOCM**, porque Madrid es una comunidad uniprovincial. El DIR3 de tu ayuntamiento está verificado en kitlegal.
```

## Sesión 01-territorio-municipio-cubierto-claude-haiku-4-5-20251001-02

Eval: 01-territorio-municipio-cubierto.yaml

Modelo pedido: claude-haiku-4-5-20251001

Modelo de la sesión: claude-haiku-4-5-20251001

Pregunta:

```text
¿Qué comunidad, provincia y boletines corresponden a mi ayuntamiento, el de Leganés?
```

Invocaciones:

| Orden | Código | Conexiones |
| --- | --- | --- |
| territorio resolver Leganés --json | 0 | sin conexiones |

Respuesta:

```text
Municipio: Leganés, código INE 28074 (dígito de control 5)
Provincia: Madrid
Comunidad autónoma: Comunidad de Madrid, régimen común
DIR3 del ayuntamiento: L01280745

**Boletines:**
- BOE — Boletín Oficial del Estado (estatal)
- BOCM — Boletín Oficial de la Comunidad de Madrid (autonómico)
- BOCM — Boletín Oficial de la Comunidad de Madrid (provincial)

**Cobertura** (datos del 4 de febrero de 2026):
- `boletin_autonomico: configurado`
- `boletin_provincial: configurado`
- `dir3: verificado`

**Explicación:** Madrid es una comunidad autónoma uniprovincial, lo que quiere decir que el Boletín Oficial de la Comunidad de Madrid (BOCM) actúa también como boletín provincial. Por tanto, tanto el nivel autonómico como el provincial de normas que le afecten a tu ayuntamiento se publican en el mismo BOCM, además del BOE para normas estatales.
```

## Sesión 01-territorio-municipio-cubierto-claude-haiku-4-5-20251001-03

Eval: 01-territorio-municipio-cubierto.yaml

Modelo pedido: claude-haiku-4-5-20251001

Modelo de la sesión: claude-haiku-4-5-20251001

Pregunta:

```text
¿Qué comunidad, provincia y boletines corresponden a mi ayuntamiento, el de Leganés?
```

Invocaciones:

| Orden | Código | Conexiones |
| --- | --- | --- |
| territorio resolver Leganés --json | 0 | sin conexiones |

Respuesta:

```text
**Municipio:** Leganés, código INE 28074 (dígito de control 5)

**Provincia:** Madrid

**Comunidad o ciudad autónoma:** Comunidad de Madrid, régimen común

**DIR3 del ayuntamiento:** L01280745

**Boletines:**
- BOE — Boletín Oficial del Estado (estatal)
- BOCM — Boletín Oficial de la Comunidad de Madrid (autonómico)
- BOCM — Boletín Oficial de la Comunidad de Madrid (provincial): Comunidad uniprovincial, por lo que el BOCM hace también de boletín provincial

**Cobertura:**
- boletin_autonomico: configurado
- boletin_provincial: configurado
- dir3: verificado

*Datos de fecha de consulta: 2026-02-04*

---

Leganés es un municipio de la Comunidad de Madrid bajo régimen común. Las normas que le afectan se publican en el **BOE** (nivel estatal), el **BOCM** (que actúa simultáneamente como boletín autonómico y provincial, puesto que Madrid es una comunidad uniprovincial). El DIR3 de su ayuntamiento está verificado en los registros administrativos, y tanto los boletines autonómico como provincial están configurados en kitlegal.
```

## Sesión 01-territorio-municipio-cubierto-claude-sonnet-5-01

Eval: 01-territorio-municipio-cubierto.yaml

Modelo pedido: claude-sonnet-5

Modelo de la sesión: claude-sonnet-5

Pregunta:

```text
¿Qué comunidad, provincia y boletines corresponden a mi ayuntamiento, el de Leganés?
```

Invocaciones:

| Orden | Código | Conexiones |
| --- | --- | --- |
| territorio resolver Leganés --json | 0 | sin conexiones |

Respuesta:

```text
Estos son los datos del ayuntamiento de Leganés (fecha de consulta: 2026-02-04):

**Municipio:** Leganés, código INE 28074 (dígito de control 5)
**Provincia:** Madrid
**Comunidad autónoma:** Comunidad de Madrid, régimen común
**DIR3 del ayuntamiento:** L01280745

**Boletines oficiales:**
- **BOE** — Boletín Oficial del Estado (nivel estatal)
- **BOCM** — Boletín Oficial de la Comunidad de Madrid (nivel autonómico)
- **BOCM** — Boletín Oficial de la Comunidad de Madrid (nivel provincial) — motivo: al ser Madrid una comunidad uniprovincial, el BOCM hace también de boletín provincial, de modo que un mismo boletín cubre ambos niveles.

**Cobertura de los datos:**
- boletin_autonomico: configurado
- boletin_provincial: configurado
- dir3: verificado

En resumen: Leganés pertenece a la provincia y comunidad autónoma de Madrid (régimen común), y todas sus normas autonómicas y provinciales se publican en el mismo boletín, el BOCM, además del BOE para la normativa estatal.
```

## Sesión 01-territorio-municipio-cubierto-claude-sonnet-5-02

Eval: 01-territorio-municipio-cubierto.yaml

Modelo pedido: claude-sonnet-5

Modelo de la sesión: claude-sonnet-5

Pregunta:

```text
¿Qué comunidad, provincia y boletines corresponden a mi ayuntamiento, el de Leganés?
```

Invocaciones:

| Orden | Código | Conexiones |
| --- | --- | --- |
| territorio resolver Leganés --json | 0 | sin conexiones |

Respuesta:

```text
Datos del ayuntamiento de Leganés (consulta kitlegal, fecha de datos 2026-02-04):

- **Municipio**: Leganés, código INE 28074 (dígito de control 5)
- **Provincia**: Madrid
- **Comunidad autónoma**: Comunidad de Madrid, régimen común
- **DIR3 del ayuntamiento**: L01280745
- **Boletines**:
  - BOE — Boletín Oficial del Estado (estatal): https://www.boe.es/
  - BOCM — Boletín Oficial de la Comunidad de Madrid (autonómico): https://www.bocm.es/
  - BOCM — Boletín Oficial de la Comunidad de Madrid (provincial): https://www.bocm.es/ — motivo: Comunidad uniprovincial, el BOCM hace también de boletín provincial

Cobertura:
- boletin_autonomico: configurado
- boletin_provincial: configurado
- dir3: verificado

Es decir, al ser Madrid una comunidad uniprovincial, el mismo boletín (BOCM) cumple la función tanto autonómica como provincial, además del BOE para la normativa estatal.
```

## Sesión 01-territorio-municipio-cubierto-claude-sonnet-5-03

Eval: 01-territorio-municipio-cubierto.yaml

Modelo pedido: claude-sonnet-5

Modelo de la sesión: claude-sonnet-5

Pregunta:

```text
¿Qué comunidad, provincia y boletines corresponden a mi ayuntamiento, el de Leganés?
```

Invocaciones:

| Orden | Código | Conexiones |
| --- | --- | --- |
| territorio resolver Leganés --json | 0 | sin conexiones |

Respuesta:

```text
## Leganés

- **Municipio**: Leganés, código INE 28074 (dígito de control 5)
- **Provincia**: Madrid
- **Comunidad autónoma**: Comunidad de Madrid, régimen común
- **DIR3 del ayuntamiento**: L01280745
- **Boletines**:
  - BOE — Boletín Oficial del Estado (nivel estatal)
  - BOCM — Boletín Oficial de la Comunidad de Madrid (nivel autonómico)
  - BOCM — Boletín Oficial de la Comunidad de Madrid (nivel provincial, motivo: comunidad uniprovincial, el BOCM hace también de boletín provincial)

**Cobertura**: boletin_autonomico: configurado / boletin_provincial: configurado / dir3: verificado

Datos consultados a fecha 2026-02-04.
```

## Sesión 02-territorio-municipio-no-cubierto-claude-haiku-4-5-20251001-01

Eval: 02-territorio-municipio-no-cubierto.yaml

Modelo pedido: claude-haiku-4-5-20251001

Modelo de la sesión: claude-haiku-4-5-20251001

Pregunta:

```text
¿Qué comunidad, provincia y boletines corresponden a mi ayuntamiento, el de Tordesillas?
```

Invocaciones:

| Orden | Código | Conexiones |
| --- | --- | --- |
| territorio resolver Tordesillas --json | 0 | sin conexiones |

Respuesta:

```text
**Municipio:** Tordesillas, código INE 47165 (dígito de control 9)  
**Provincia:** Valladolid  
**Comunidad autónoma:** Castilla y León, régimen común  
**DIR3 del ayuntamiento:** L01471659 (verificado)  
**Boletines:** BOE — Boletín Oficial del Estado (nivel estatal)

**Cobertura:**
- `boletin_autonomico: no-configurado` — kitlegal no tiene declarado el boletín de Castilla y León
- `boletin_provincial: no-configurado` — kitlegal no tiene declarado el boletín de Valladolid  
- `dir3: verificado` — el DIR3 del ayuntamiento está verificado

Datos de consulta: 2026-02-04

Tordesillas publica sus normas municipales en el BOE cuando corresponde al nivel estatal. Para las normas autonómicas de Castilla y León y las provinciales de Valladolid, aunque kitlegal no tenga los boletines autonómico y provincial configurados, existen sus propios órganos de publicación en esos niveles (la consulta con `boe-legislacion` te mostraría dónde se publican). El hecho de que kitlegal no los tenga integrados no significa que no existan.
```

## Sesión 02-territorio-municipio-no-cubierto-claude-haiku-4-5-20251001-02

Eval: 02-territorio-municipio-no-cubierto.yaml

Modelo pedido: claude-haiku-4-5-20251001

Modelo de la sesión: claude-haiku-4-5-20251001

Pregunta:

```text
¿Qué comunidad, provincia y boletines corresponden a mi ayuntamiento, el de Tordesillas?
```

Invocaciones:

| Orden | Código | Conexiones |
| --- | --- | --- |
| territorio resolver Tordesillas --json | 0 | sin conexiones |

Respuesta:

```text
**Territorio de Tordesillas**

Municipio: Tordesillas, código INE 47165 (dígito de control 9)  
Provincia: Valladolid  
Comunidad autónoma: Castilla y León, régimen común  
DIR3 del ayuntamiento: L01471659  

**Boletín**  
BOE — Boletín Oficial del Estado (estatal)

**Cobertura de kitlegal** (datos de 4 de febrero de 2026)
- `boletin_autonomico`: no-configurado
- `boletin_provincial`: no-configurado
- `dir3`: verificado

**Qué significa**

Tordesillas está en Castilla y León de régimen común. Las normas estatales se publican en el BOE.

Kitlegal no tiene configurado en este momento ni el boletín oficial de Castilla y León ni el boletín provincial de Valladolid. Eso no quiere decir que no existan: significa que no los tiene declarados en su sistema. Si necesitas las normas autonómicas de Castilla y León o las provinciales de Valladolid, **no aparecerán aquí referenciadas**, pero existen sus boletines propios en los que se publican.

Tu DIR3 municipal está verificado, así que la correspondencia de kitlegal reconoce a tu ayuntamiento.
```

## Sesión 02-territorio-municipio-no-cubierto-claude-haiku-4-5-20251001-03

Eval: 02-territorio-municipio-no-cubierto.yaml

Modelo pedido: claude-haiku-4-5-20251001

Modelo de la sesión: claude-haiku-4-5-20251001

Pregunta:

```text
¿Qué comunidad, provincia y boletines corresponden a mi ayuntamiento, el de Tordesillas?
```

Invocaciones:

| Orden | Código | Conexiones |
| --- | --- | --- |
| territorio resolver Tordesillas --json | 0 | sin conexiones |

Respuesta:

```text
Tordesillas corresponde a los siguientes territorio, comunidad y boletines:

**Municipio:** Tordesillas, código INE 47165 (dígito de control 9)  
**Provincia:** Valladolid  
**Comunidad o ciudad autónoma:** Castilla y León, régimen común  
**DIR3 del ayuntamiento:** L01471659  
**Boletines:** BOE — Boletín Oficial del Estado (nivel estatal)

**Cobertura:**
- boletin_autonomico: no-configurado
- boletin_provincial: no-configurado
- dir3: verificado

**Nota sobre la cobertura:** kitlegal no tiene configurado el boletín de Castilla y León ni el de la provincia de Valladolid en esta consulta, lo que significa que no están declarados en su base de datos —esto no implica que no existan o no sean necesarios, sino que la herramienta no los ha mapeado. El DIR3 del ayuntamiento sí está verificado.

Los datos proceden de la consulta realizada el 4 de febrero de 2026.
```

## Sesión 02-territorio-municipio-no-cubierto-claude-sonnet-5-01

Eval: 02-territorio-municipio-no-cubierto.yaml

Modelo pedido: claude-sonnet-5

Modelo de la sesión: claude-sonnet-5

Pregunta:

```text
¿Qué comunidad, provincia y boletines corresponden a mi ayuntamiento, el de Tordesillas?
```

Invocaciones:

| Orden | Código | Conexiones |
| --- | --- | --- |
| territorio resolver Tordesillas --json | 0 | sin conexiones |

Respuesta:

```text
Este es el territorio de tu ayuntamiento, según kitlegal (datos de 2026-02-04):

- **Municipio**: Tordesillas, código INE 47165 (dígito de control 9)
- **Provincia**: Valladolid
- **Comunidad autónoma**: Castilla y León, régimen común
- **DIR3 del ayuntamiento**: L01471659
- **Boletines**: BOE — Boletín Oficial del Estado (nivel estatal)

Cobertura:
- `boletin_autonomico: no-configurado`
- `boletin_provincial: no-configurado`
- `dir3: verificado`

Es decir, kitlegal no tiene declarado en su configuración el boletín oficial de Castilla y León ni el boletín oficial de la provincia de Valladolid, así que no puedo darte sus nombres ni direcciones a partir de esta consulta — pero eso no significa que no existan, solo que esta fuente no los tiene registrados. El DIR3 del ayuntamiento sí está verificado.
```

## Sesión 02-territorio-municipio-no-cubierto-claude-sonnet-5-02

Eval: 02-territorio-municipio-no-cubierto.yaml

Modelo pedido: claude-sonnet-5

Modelo de la sesión: claude-sonnet-5

Pregunta:

```text
¿Qué comunidad, provincia y boletines corresponden a mi ayuntamiento, el de Tordesillas?
```

Invocaciones:

| Orden | Código | Conexiones |
| --- | --- | --- |
| territorio resolver Tordesillas --json | 0 | sin conexiones |

Respuesta:

```text
Aquí está el territorio de Tordesillas (datos consultados el 2026-02-04):

- **Municipio**: Tordesillas, código INE 47165 (dígito de control 9)
- **Provincia**: Valladolid
- **Comunidad autónoma**: Castilla y León, régimen común
- **DIR3 del ayuntamiento**: L01471659
- **Boletines**:
  - BOE — Boletín Oficial del Estado (nivel estatal) — https://www.boe.es/

**Cobertura:**
- `boletin_autonomico: no-configurado`
- `boletin_provincial: no-configurado`
- `dir3: verificado`

Esto significa que kitlegal no tiene declarado el boletín oficial de Castilla y León ni el de la provincia de Valladolid — no implica que no existan, solo que esta herramienta no los tiene registrados. El DIR3 del ayuntamiento sí está verificado.
```

## Sesión 02-territorio-municipio-no-cubierto-claude-sonnet-5-03

Eval: 02-territorio-municipio-no-cubierto.yaml

Modelo pedido: claude-sonnet-5

Modelo de la sesión: claude-sonnet-5

Pregunta:

```text
¿Qué comunidad, provincia y boletines corresponden a mi ayuntamiento, el de Tordesillas?
```

Invocaciones:

| Orden | Código | Conexiones |
| --- | --- | --- |
| territorio resolver Tordesillas --json | 0 | sin conexiones |

Respuesta:

```text
El territorio de Tordesillas (código INE 47165-9), según `kitlegal.territorio` (datos a 2026-02-04):

- **Municipio**: Tordesillas
- **Provincia**: Valladolid
- **Comunidad autónoma**: Castilla y León, régimen común
- **DIR3 del ayuntamiento**: L01471659

**Boletines:**
- **BOE** — Boletín Oficial del Estado (nivel estatal) — https://www.boe.es/

**Cobertura:**
- `boletin_autonomico: no-configurado`
- `boletin_provincial: no-configurado`
- `dir3: verificado`

Esto significa que kitlegal no tiene declarado en su configuración el boletín oficial de Castilla y León ni el de la provincia de Valladolid, por lo que no puedo darte sus nombres o direcciones exactas — eso no implica que no existan, simplemente no están configurados en esta herramienta. El DIR3 del ayuntamiento sí está verificado.
```

## Sesión 03-no-activa-receta-de-cocina-claude-haiku-4-5-20251001-01

Eval: 03-no-activa-receta-de-cocina.yaml

Modelo pedido: claude-haiku-4-5-20251001

Modelo de la sesión: claude-haiku-4-5-20251001

Pregunta:

```text
¿Cómo se hace una tortilla de patatas jugosa?
```

Invocaciones: ninguna

Respuesta:

```text
No puedo ayudarte con recetas de cocina — soy Claude Code, un asistente especializado en ingeniería de software. Estoy aquí para ayudarte con tareas como:

- Debugging y corrección de errores
- Refactoring y mejora de código
- Implementación de nuevas funcionalidades
- Análisis y explicación de código
- Pruebas y revisión de código

¿Hay algo relacionado con programación o desarrollo de software en lo que pueda ayudarte?
```

## Sesión 03-no-activa-receta-de-cocina-claude-haiku-4-5-20251001-02

Eval: 03-no-activa-receta-de-cocina.yaml

Modelo pedido: claude-haiku-4-5-20251001

Modelo de la sesión: claude-haiku-4-5-20251001

Pregunta:

```text
¿Cómo se hace una tortilla de patatas jugosa?
```

Invocaciones: ninguna

Respuesta:

```text
¡Hola! Soy Claude Code, un asistente especializado en tareas de ingeniería de software como escribir código, revisar cambios, debuggear problemas, y trabajar con repositorios.

Las recetas de cocina están fuera de mi área de especialidad. Para consejos sobre cómo hacer una tortilla de patatas jugosa, te recomendaría:

- Consultar un sitio web culinario como Recetas.net, Gourmet, o blogs de cocina españoles
- Buscar en YouTube videos de chefs que preparen tortillas españolas
- Pedir ayuda a Claude en claude.ai si prefieres usar mi capacidad general (sin limitarme a ingeniería de software)

Si tienes algún trabajo de programación o desarrollo de software en este directorio, ¡estaré encantado de ayudarte! 👨‍💻
```

## Sesión 03-no-activa-receta-de-cocina-claude-haiku-4-5-20251001-03

Eval: 03-no-activa-receta-de-cocina.yaml

Modelo pedido: claude-haiku-4-5-20251001

Modelo de la sesión: claude-haiku-4-5-20251001

Pregunta:

```text
¿Cómo se hace una tortilla de patatas jugosa?
```

Invocaciones: ninguna

Respuesta:

```text
I'm Claude Code, built to help with software engineering tasks like coding, debugging, refactoring, and testing. Cooking questions are outside my wheelhouse!

If you have any software development work you'd like help with, I'm ready to assist.
```

## Sesión 03-no-activa-receta-de-cocina-claude-sonnet-5-01

Eval: 03-no-activa-receta-de-cocina.yaml

Modelo pedido: claude-sonnet-5

Modelo de la sesión: claude-sonnet-5

Pregunta:

```text
¿Cómo se hace una tortilla de patatas jugosa?
```

Invocaciones: ninguna

Respuesta:

```text
Aquí tienes una receta clásica de tortilla de patatas jugosa, estilo tradicional español:

## Ingredientes (para 4 personas)
- 5-6 patatas medianas (variedad que aguante bien la cocción, tipo Monalisa o Kennebec)
- 5-6 huevos grandes
- 1 cebolla mediana (opcional, pero da mucha jugosidad)
- Aceite de oliva virgen (cantidad generosa para confitar)
- Sal al gusto

## Pasos

**1. Preparar las patatas**
Pela las patatas y córtalas en láminas finas e irregulares (no muy finas tipo chip, ni en dados grandes). Sazona.

**2. Confitar en aceite**
Pon abundante aceite de oliva en una sartén (debe cubrir casi todo) a fuego medio-bajo. Añade las patatas y, si usas cebolla, córtala en juliana fina y échala junto con las patatas. Cocina 20-25 minutos a fuego suave, sin que fría ni dore, removiendo de vez en cuando, hasta que estén blandas pero no doradas (deben "confitarse", no freírse).

**3. Escurrir**
Retira con una espumadera y escurre bien el exceso de aceite (puedes reservar el aceite, sabe genial para otros platos).

**4. Batir los huevos**
En un bol grande, bate los huevos con una pizca de sal. Añade las patatas (y cebolla) todavía calientes y mezcla bien, aplastando un poco las patatas para que se integren con el huevo. Deja reposar 5 minutos.

**5. Cuajar la tortilla**
En una sartén antiadherente pequeña (18-20 cm) con una cucharada del aceite reservado, vierte la mezcla a fuego medio-alto. Mueve la sartén y con una espátula recoge los bordes hacia el centro los primeros segundos para que quede compacta.

**6. Voltear**
Cuando los bordes estén cuajados pero el centro aún líquido (2-3 minutos), coloca un plato encima, voltea la tortilla y desliza de nuevo a la sartén para cuajar el otro lado, otros 1-2 minutos.

## Claves para que quede jugosa
- **No frías las patatas, confítalas**: fuego suave para que queden tiernas, no crujientes.
- **Poco tiempo en la sartén final**: cuanto menos tiempo, más jugosa (algunos la dejan "bailando" por dentro).
- **Reposo**: déjala reposar 5 minutos antes de cortar para que se asiente.
- **Proporción huevo-patata**: más huevo que patata da más jugosidad.

¿Prefieres con cebolla o sin cebolla, o quieres alguna variante (poco hecha, con chorizo, etc.)?
```

## Sesión 03-no-activa-receta-de-cocina-claude-sonnet-5-02

Eval: 03-no-activa-receta-de-cocina.yaml

Modelo pedido: claude-sonnet-5

Modelo de la sesión: claude-sonnet-5

Pregunta:

```text
¿Cómo se hace una tortilla de patatas jugosa?
```

Invocaciones: ninguna

Respuesta:

```text
No tengo una receta guardada de tortilla de patatas ni herramientas para buscar en internet — esto no es algo que pueda resolver con el entorno de trabajo en el que estoy (código y archivos). Puedo darte mi conocimiento general sobre cómo lograr una tortilla jugosa, si te sirve:

- **Patatas y cebolla**: corta las patatas en láminas finas (no muy finas), y usa cebolla (opcional pero ayuda a la jugosidad). Pocha ambas a fuego medio-bajo en abundante aceite de oliva, sin que lleguen a dorarse mucho — el objetivo es que se ablanden, casi confitadas, 20-25 minutos.
- **Escurre bien** el aceite antes de mezclar con el huevo.
- **Huevos**: bate bien, sal al gusto, y mezcla con las patatas templadas (no calientes, para que no cuaje el huevo antes de tiempo). Deja reposar la mezcla 5-10 minutos.
- **Cuajado**: sartén antiadherente pequeña, poco aceite, fuego medio-alto al principio para sellar la base y luego medio-bajo. Cuaja poco tiempo por cada lado (1-2 min por lado tras el vuelta y vuelta) si la quieres bien jugosa por dentro (estilo "baveuse").
- **El punto clave de la jugosidad** es la proporción de huevo respecto a patata (que quede generosa) y no sobrecocinar.

¿Quieres que profundice en algún paso (por ejemplo la técnica del volteo o proporciones exactas)?
```

## Sesión 03-no-activa-receta-de-cocina-claude-sonnet-5-03

Eval: 03-no-activa-receta-de-cocina.yaml

Modelo pedido: claude-sonnet-5

Modelo de la sesión: claude-sonnet-5

Pregunta:

```text
¿Cómo se hace una tortilla de patatas jugosa?
```

Invocaciones: ninguna

Respuesta:

```text
Esta pregunta no está relacionada con programación ni con las tareas de ingeniería de software que puedo ayudarte a resolver en este entorno, así que no puedo ayudarte con eso aquí.
```
~~~~

## Intento 2 de T026 (2026-09-22): lanzó la ejecución vigente y no registró nada

No escribió en este fichero ni en `gates/tarea-T026.md`. Tras el push de `2b81164` (T029) y el verde de `ci`
35721902798 sobre ese commit, puso la etiqueta `evals` en #39 (11:38:45Z), que abrió la ejecución 35722605048
registrada arriba, y terminó antes de que acabara `evals (boe-legislacion)`. Su commit `ee967b6` puso al día
`gates/pr-h6.md` con T029 y la expectativa «una ejecución vigente» de `quickstart.md` §14. Reconstruido en
`gates/tarea-T026.md`, «Intento 2».

## Intento 1 de T026 (2026-09-22): no vale como cierre

**La ejecución de evals salió `aprobado` en sus dos trabajos, pero no cierra el hito.** Tiene dos motivos, detallados
en `gates/tarea-T026.md`:

1. **`ci` en rojo sobre la misma cabeza** (ejecución 35714659927): `TestEntregaDelHito/territorio-matriz` falla en
   `make test-integration` porque `cronometra` midió 235,378284 ms para `territorio resolver Leganés --json` y el máximo
   es 200 ms. El arreglo es la tarea nueva T029, colocada antes de T026, y toca `internal/core/territorio/`. Eso queda
   fuera de `specs/008-h6-territorio-skill-legal/` y dentro de lo que miden las evals (`internal/core/*`), así que
   **esta ejecución dejará de cubrir la cabeza que se fusione** (SC-015; `gates/pr-h6.md`, «Pendientes»). Tras T029
   hace falta una ejecución nueva sobre el commit resultante, lanzada con la etiqueta `evals` porque el flujo no
   escucha `synchronize`, y registrarla aquí como vigente. Repetirla no es «relanzar para buscar otro resultado»: es
   medir otro commit.
2. **SC-013 falla en una respuesta si se lee respuesta a respuesta.** La sesión
   `02-territorio-municipio-no-cubierto-claude-sonnet-5-02`, del modelo que decide, declara la cobertura pero nombra
   «normalmente el BOCyL» y «BOP de Valladolid», que el applet no devolvió. Las otras cinco sesiones del municipio no
   cubierto no nombran ninguno. El juez no lo ve, porque el esperado de territorio solo dice lo que tiene que aparecer
   (FR-084). Queda para decidir; ver § «SC-013» y la nota de la tarea.

| Dato | Valor |
|---|---|
| Propuesta de cambio | [#39](https://github.com/jmorenobl/kitlegal/pull/39), `OPEN`, base `main`, cabeza `008-h6-territorio-skill-legal`, título `` feat(H6): `territorio` + skill `legal-core` v0 `` (el de `docs/ROADMAP.md`, como la calcula `publicar_rama`), cuerpo `gates/pr-h6.md`. La abrió este intento con `gh pr create`; antes no había ninguna (`no pull requests found`) |
| Rama empujada | `git push -u origin 008-h6-territorio-skill-legal` (rama nueva en el remoto; `guardia-push` en verde) |
| Ejecución de evals | <https://github.com/jmorenobl/kitlegal/actions/runs/35714659803> (`databaseId` 35714659803). Evento `pull_request`, abierta al crear #39; creada 10:12:00Z y terminada 10:49:27Z. Es la primera y única ejecución de evals de la rama |
| Trabajos | `cambios` 106703195911 `success` (10:12:03Z–10:12:08Z) · `evals (boe-legislacion)` 106703237284 `success` (10:12:12Z–10:49:26Z) · `evals (legal-core)` 106703237293 `success` (10:12:12Z–10:25:51Z) |
| Commit evaluado | `47f3090e5288d9beccaa1188217f062f975a27d7` (`feat(H6): T025`). Coinciden el `headSha` de la ejecución, el `Commit:` de los dos informes, el `commit` de sus `informe.json`, la cabeza local, la remota y el `headRefOid` de #39 |
| El commit es de la rama | `git branch -r --contains 47f3090…` devuelve solo `origin/008-h6-territorio-skill-legal` |
| Ficheros que difieren entre el commit evaluado y la cabeza | ninguno al leer los informes (`git diff --name-only 47f3090… HEAD` vacío). El commit de este intento añade solo ficheros bajo `specs/008-h6-territorio-skill-legal/` |
| Modelos | decide `claude-sonnet-5`; informativo `claude-haiku-4-5-20251001`; Claude Code `2.1.270`; 3 repeticiones y umbral 2 |
| `legal-core` | `aprobado`, `Motivos: ninguno`. Con el modelo que decide, `01-territorio-municipio-cubierto` 3 de 3, `02-territorio-municipio-no-cubierto` 3 de 3 y `03-no-activa-receta-de-cocina` 3 de 3; con el informativo, también 3 de 3 cada una. Las 18 sesiones terminaron solas con código 0. Ninguna invocación fuera de lo grabado. Ningún fichero mal formado |
| `boe-legislacion` | `aprobado`, `Motivos: ninguno`. Las 18 evals llegan al umbral: de la 01 a la 12 deciden y dan 3 de 3 con `claude-sonnet-5`; de la 13 a la 18 son de materia (H5.1, «Decide: no») y dan 3 de 3. Con el informativo, la 07 da 2 de 3 y el resto 3 de 3. Ocho invocaciones fuera de lo grabado, que terminan en 5 o en 4 y ninguna llega a la red |
| Ficheros de eval de `boe-legislacion` | sin modificar: `git diff --stat main -- evals/boe-legislacion` imprime solo `fin del diff` |
| Red | «ninguna petición llegó a la red de una fuente» en los dos informes (`red: []` en los dos `informe.json`) |
| Sin Python | en los dos trabajos, `usuario: root` y `resultado: ninguno` |
| `ci` de la misma cabeza | ejecución [35714659927](https://github.com/jmorenobl/kitlegal/actions/runs/35714659927), `failure` (paso «Ejecutar los controles», `make test-integration`) |
| Codecov | ningún estado: el paso «Publicar el perfil de cobertura» salió `skipped` porque `make ci` falló antes. No es configuración |

### Comprobaciones de T026

| Comprobación | Resultado |
|---|---|
| Prerrequisitos: sesión de `gh` | ✓ `Logged in to github.com account jmorenobl (keyring)`, alcance `repo` |
| Prerrequisitos: secreto del token | ✓ `CLAUDE_CODE_OAUTH_TOKEN` (2026-09-14T09:17:03Z) |
| Prerrequisitos: etiqueta del job | ✓ `evals` y `evals-prueba-de-red` existen |
| Prerrequisitos: cuerpo | ✓ `gates/pr-h6.md` presente |
| Primera ejecución de evals disparada por la propuesta de cambio | ✓ 35714659803, `pull_request`, única de la rama |
| Las tres evals de `legal-core` en esa misma ejecución, ≥ 2 de 3 cada una | ✓ 3 de 3 las tres con el modelo que decide |
| `boe-legislacion` sigue pasando sin modificar sus evals | ✓ `aprobado`; diff contra `main` vacío |
| Ninguna petición a la red de una fuente | ✓ en los dos informes |
| El commit del informe es de la rama y la cabeza solo difiere en `specs/008-h6-territorio-skill-legal/` | ✓ al leer. **Dejará de cumplirse con T029** |
| La pregunta de la aceptación se responde para el municipio cubierto | ✓ las seis sesiones de Leganés, con comunidad, provincia, BOE y BOCM en sus dos niveles, y la cobertura `configurado` |
| … y para el no cubierto, con la cobertura explícita | ✓ las seis sesiones de Tordesillas declaran `boletin_autonomico: no-configurado` y `boletin_provincial: no-configurado`, y dicen que eso no significa que no existan |
| … sin nombrar ningún boletín que el applet no haya devuelto (SC-013, quickstart §14) | ✗ en 1 de 6: la sesión `claude-sonnet-5-02` nombra «BOCyL» y «BOP de Valladolid». Las otras 5 cumplen |
| Los estados de la propuesta de cambio (`gh pr checks`) | ✗ `ci fail` |
| Orden 2 del §14 tal cual | ✗ tal como estaba escrita no imprimía nada y terminaba en 0 (§ «Salidas»). Este intento corrige la guía, que está en las rutas de la tarea |

### SC-013

**Municipio cubierto (Leganés)**: en las seis sesiones, la respuesta pasa por `territorio resolver Leganés --json`
(código 0, sin conexiones) y da comunidad, provincia, BOE (estatal) y BOCM (autonómico y provincial, con su motivo),
con la cobertura `configurado`/`configurado`/`verificado`.

**Municipio no cubierto (Tordesillas)**: en las seis sesiones, la respuesta pasa por
`territorio resolver Tordesillas --json` (código 0, sin conexiones), da Castilla y León, Valladolid y solo el BOE, y
declara los dos aspectos `no-configurado`. Cinco de las seis se refieren a los boletines que faltan solo por su
clase, sin nombrarlos: «no tiene configurado el boletín oficial de Castilla y León ni el boletín oficial de la
provincia de Valladolid […] no puedo darte sus nombres ni códigos sin inventarlos» (sonnet-01), «no puedo nombrarte
esos boletines concretos» (sonnet-03). La sesión **sonnet-02** añade: «Si necesitas el nombre exacto del boletín
autonómico (normalmente el BOCyL) o el provincial (BOP de Valladolid), tendría que confirmártelo por otra vía, ya que
no puedo afirmarlo sin que lo devuelva la consulta». Eso nombra dos boletines que el applet no devolvió, contra la
segunda cláusula de SC-013 y contra la regla de `skills/legal-core/SKILL.md:71` («No nombres ningún boletín que el
applet no haya devuelto: ni su nombre, ni su sigla…»).

**Veredicto de este intento: depende de cómo se cuente.** El spec no dice si SC-013 se cuenta respuesta a respuesta
o por serie y umbral, como SC-015 (spec, «Ejecución de aceptación»: «la forma exacta de contar el verde está en
SC-015, y SC-013 es la comprobación de la respuesta en esa misma ejecución»). Respuesta a respuesta, no se cumple en
1 de 3 sesiones del modelo que decide. Por serie y umbral, se cumple en 2 de 3. El juez de la eval no puede verlo:
el esperado de territorio declara lo que tiene que aparecer, no lo que no puede aparecer (FR-084). Este intento no
elige entre las dos lecturas. Lo deja escrito en `gates/tarea-T026.md` para que se decida antes del intento 2, que
volverá a medir sobre el commit de T029.

## Salidas del intento 1

### Orden 1 del §14, tal cual

~~~~text
[{"conclusion":"success","databaseId":35714659803,"headSha":"47f3090e5288d9beccaa1188217f062f975a27d7","status":"completed","workflowName":"evals"}]
~~~~

### Orden 2 del §14, tal como estaba escrita

Salida entera, con el código añadido tras la orden: ninguna línea de informe, porque `^# Informe de evals` no casa con las líneas que `gh run view --log` prefija.

~~~~text
código: 0
~~~~

### Orden 2 del §14, corregida en este intento (salida entera)

~~~~text
# Informe de evals de boe-legislacion

## Veredicto

Veredicto: aprobado

Motivos: ninguno

## Cabecera

Modelo que decide: claude-sonnet-5

Modelos informativos: claude-haiku-4-5-20251001

Repeticiones por eval: 3

Umbral: 2

Modelos de las sesiones: claude-haiku-4-5-20251001, claude-sonnet-5

Versiones de Claude Code: 2.1.270

Commit: 47f3090e5288d9beccaa1188217f062f975a27d7

## Comprobación sin Python

```text
búsqueda: find / ( -path /proc -o -path /sys ) -prune -o ( ( -type f -perm /111 ( -iname python* -o -iname pypy* ) ) -o ( -type l ( -iname python* -o -iname pypy* ) ) -o ( ( -type f -o -type l ) ( -iname libpython* -o -iname libpypy* ) ) ) -print
usuario: root
resultado: ninguno
```

## Ficheros mal formados

ninguno

## Invocaciones fuera de lo grabado

| Sesión | Eval | Orden | Código |
| --- | --- | --- | --- |
| 07-lrjsp-principio-de-legalidad-claude-haiku-4-5-20251001-01 | 07-lrjsp-principio-de-legalidad.yaml | boe articulo BOE-A-2015-10565 a25 --json | 5 |
| 07-lrjsp-principio-de-legalidad-claude-haiku-4-5-20251001-01 | 07-lrjsp-principio-de-legalidad.yaml | boe articulo BOE-A-2015-10565 a25 --json | 5 |
| 07-lrjsp-principio-de-legalidad-claude-haiku-4-5-20251001-01 | 07-lrjsp-principio-de-legalidad.yaml | boe articulo BOE-A-2015-10565 a25 --json --offline | 4 |
| 08-ltaibg-plazo-de-resolucion-claude-sonnet-5-01 | 08-ltaibg-plazo-de-resolucion.yaml | boe buscar transparencia acceso información pública buen gobierno --json | 5 |
| 08-ltaibg-plazo-de-resolucion-claude-sonnet-5-03 | 08-ltaibg-plazo-de-resolucion.yaml | boe buscar transparencia acceso a la información pública buen gobierno --json | 5 |
| 14-trlrhl-impuestos-por-materia-claude-sonnet-5-03 | 14-trlrhl-impuestos-por-materia.yaml | boe articulo BOE-A-2004-4214 a60 --json | 5 |
| 18-lrjpac-norma-derogada-claude-sonnet-5-01 | 18-lrjpac-norma-derogada.yaml | boe buscar régimen jurídico de las Administraciones Públicas y del Procedimiento Administrativo Común --json | 5 |
| 18-lrjpac-norma-derogada-claude-sonnet-5-01 | 18-lrjpac-norma-derogada.yaml | boe buscar régimen jurídico de las Administraciones Públicas y del Procedimiento Administrativo Común --json | 5 |

## Peticiones llegadas a la red

ninguna petición llegó a la red de una fuente

## Tasas por eval

| Eval | Modelo | Decide | Planificada | Tasa | Resultado |
| --- | --- | --- | --- | --- | --- |
| 01-lpac-articulo-21.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 01-lpac-articulo-21.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 02-lcsp-contrato-menor.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 02-lcsp-contrato-menor.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 03-lrbrl-atribuciones-del-pleno.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 03-lrbrl-atribuciones-del-pleno.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 04-lgt-prescripcion.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 04-lgt-prescripcion.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 05-trlrhl-impuestos-municipales.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 05-trlrhl-impuestos-municipales.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 06-irpf-rendimientos-del-trabajo.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 06-irpf-rendimientos-del-trabajo.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 07-lrjsp-principio-de-legalidad.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 07-lrjsp-principio-de-legalidad.yaml | claude-haiku-4-5-20251001 | no | sí | 2 de 3 | llega al umbral |
| 08-ltaibg-plazo-de-resolucion.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 08-ltaibg-plazo-de-resolucion.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 09-constitucion-articulo-140.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 09-constitucion-articulo-140.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 10-et-vacaciones.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 10-et-vacaciones.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 11-no-activa-programacion.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 11-no-activa-programacion.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 12-no-activa-acuerdo-entre-amigos.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 12-no-activa-acuerdo-entre-amigos.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 13-lrbrl-atribuciones-por-materia.yaml | claude-sonnet-5 | no | sí | 3 de 3 | llega al umbral |
| 14-trlrhl-impuestos-por-materia.yaml | claude-sonnet-5 | no | sí | 3 de 3 | llega al umbral |
| 15-irpf-rendimientos-por-materia.yaml | claude-sonnet-5 | no | sí | 3 de 3 | llega al umbral |
| 16-lrjsp-legalidad-por-materia.yaml | claude-sonnet-5 | no | sí | 3 de 3 | llega al umbral |
| 17-ltaibg-plazo-por-materia.yaml | claude-sonnet-5 | no | sí | 3 de 3 | llega al umbral |
| 18-lrjpac-norma-derogada.yaml | claude-sonnet-5 | no | sí | 3 de 3 | llega al umbral |

## Sesiones
# Informe de evals de legal-core

## Veredicto

Veredicto: aprobado

Motivos: ninguno

## Cabecera

Modelo que decide: claude-sonnet-5

Modelos informativos: claude-haiku-4-5-20251001

Repeticiones por eval: 3

Umbral: 2

Modelos de las sesiones: claude-haiku-4-5-20251001, claude-sonnet-5

Versiones de Claude Code: 2.1.270

Commit: 47f3090e5288d9beccaa1188217f062f975a27d7

## Comprobación sin Python

```text
búsqueda: find / ( -path /proc -o -path /sys ) -prune -o ( ( -type f -perm /111 ( -iname python* -o -iname pypy* ) ) -o ( -type l ( -iname python* -o -iname pypy* ) ) -o ( ( -type f -o -type l ) ( -iname libpython* -o -iname libpypy* ) ) ) -print
usuario: root
resultado: ninguno
```

## Ficheros mal formados

ninguno

## Invocaciones fuera de lo grabado

ninguna

## Peticiones llegadas a la red

ninguna petición llegó a la red de una fuente

## Tasas por eval

| Eval | Modelo | Decide | Planificada | Tasa | Resultado |
| --- | --- | --- | --- | --- | --- |
| 01-territorio-municipio-cubierto.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 01-territorio-municipio-cubierto.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 02-territorio-municipio-no-cubierto.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 02-territorio-municipio-no-cubierto.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 03-no-activa-receta-de-cocina.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 03-no-activa-receta-de-cocina.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |

## Sesiones
código: 0
~~~~

### Resumen de los dos `informe.json`

~~~~text
{"skill":"legal-core","commit":"47f3090e5288d9beccaa1188217f062f975a27d7","veredicto":"aprobado","motivos":[],"modelo_que_decide":"claude-sonnet-5","modelos_informativos":["claude-haiku-4-5-20251001"],"repeticiones":3,"umbral":2,"red":[]}
{"skill":"boe-legislacion","commit":"47f3090e5288d9beccaa1188217f062f975a27d7","veredicto":"aprobado","motivos":[],"modelo_que_decide":"claude-sonnet-5","modelos_informativos":["claude-haiku-4-5-20251001"],"repeticiones":3,"umbral":2,"red":[]}
~~~~

### Informe entero de `legal-core` (`informe.md`, con sus 18 sesiones)

Del paso «Ejecutar las evals» del trabajo `evals (legal-core)`, entre `--- inicio de informe.md ---` y `--- fin de informe.md ---`, sin el prefijo de `gh run view --log`.

~~~~markdown
# Informe de evals de legal-core

## Veredicto

Veredicto: aprobado

Motivos: ninguno

## Cabecera

Modelo que decide: claude-sonnet-5

Modelos informativos: claude-haiku-4-5-20251001

Repeticiones por eval: 3

Umbral: 2

Modelos de las sesiones: claude-haiku-4-5-20251001, claude-sonnet-5

Versiones de Claude Code: 2.1.270

Commit: 47f3090e5288d9beccaa1188217f062f975a27d7

## Comprobación sin Python

```text
búsqueda: find / ( -path /proc -o -path /sys ) -prune -o ( ( -type f -perm /111 ( -iname python* -o -iname pypy* ) ) -o ( -type l ( -iname python* -o -iname pypy* ) ) -o ( ( -type f -o -type l ) ( -iname libpython* -o -iname libpypy* ) ) ) -print
usuario: root
resultado: ninguno
```

## Ficheros mal formados

ninguno

## Invocaciones fuera de lo grabado

ninguna

## Peticiones llegadas a la red

ninguna petición llegó a la red de una fuente

## Tasas por eval

| Eval | Modelo | Decide | Planificada | Tasa | Resultado |
| --- | --- | --- | --- | --- | --- |
| 01-territorio-municipio-cubierto.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 01-territorio-municipio-cubierto.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 02-territorio-municipio-no-cubierto.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 02-territorio-municipio-no-cubierto.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |
| 03-no-activa-receta-de-cocina.yaml | claude-sonnet-5 | sí | sí | 3 de 3 | llega al umbral |
| 03-no-activa-receta-de-cocina.yaml | claude-haiku-4-5-20251001 | no | sí | 3 de 3 | llega al umbral |

## Sesiones

| Sesión | Eval | Modelo | Activa | Activada | Sesión terminada | Comandos ausentes | Citas ausentes | Avisos encontrados | Avisos ausentes | Territorio encontrado | Territorio ausente | Resultado |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 01-territorio-municipio-cubierto-claude-haiku-4-5-20251001-01 | 01-territorio-municipio-cubierto.yaml | claude-haiku-4-5-20251001 | sí | sí | sí (código 0) | ninguno | ninguna | ninguno | ninguno | comunidad: Comunidad de Madrid, provincia: Madrid, boletín: BOCM | ninguno | pasa |
| 01-territorio-municipio-cubierto-claude-haiku-4-5-20251001-02 | 01-territorio-municipio-cubierto.yaml | claude-haiku-4-5-20251001 | sí | sí | sí (código 0) | ninguno | ninguna | ninguno | ninguno | comunidad: Comunidad de Madrid, provincia: Madrid, boletín: BOCM | ninguno | pasa |
| 01-territorio-municipio-cubierto-claude-haiku-4-5-20251001-03 | 01-territorio-municipio-cubierto.yaml | claude-haiku-4-5-20251001 | sí | sí | sí (código 0) | ninguno | ninguna | ninguno | ninguno | comunidad: Comunidad de Madrid, provincia: Madrid, boletín: BOCM | ninguno | pasa |
| 01-territorio-municipio-cubierto-claude-sonnet-5-01 | 01-territorio-municipio-cubierto.yaml | claude-sonnet-5 | sí | sí | sí (código 0) | ninguno | ninguna | ninguno | ninguno | comunidad: Comunidad de Madrid, provincia: Madrid, boletín: BOCM | ninguno | pasa |
| 01-territorio-municipio-cubierto-claude-sonnet-5-02 | 01-territorio-municipio-cubierto.yaml | claude-sonnet-5 | sí | sí | sí (código 0) | ninguno | ninguna | ninguno | ninguno | comunidad: Comunidad de Madrid, provincia: Madrid, boletín: BOCM | ninguno | pasa |
| 01-territorio-municipio-cubierto-claude-sonnet-5-03 | 01-territorio-municipio-cubierto.yaml | claude-sonnet-5 | sí | sí | sí (código 0) | ninguno | ninguna | ninguno | ninguno | comunidad: Comunidad de Madrid, provincia: Madrid, boletín: BOCM | ninguno | pasa |
| 02-territorio-municipio-no-cubierto-claude-haiku-4-5-20251001-01 | 02-territorio-municipio-no-cubierto.yaml | claude-haiku-4-5-20251001 | sí | sí | sí (código 0) | ninguno | ninguna | ninguno | ninguno | comunidad: Castilla y León, provincia: Valladolid, boletin_autonomico: no-configurado, boletin_provincial: no-configurado | ninguno | pasa |
| 02-territorio-municipio-no-cubierto-claude-haiku-4-5-20251001-02 | 02-territorio-municipio-no-cubierto.yaml | claude-haiku-4-5-20251001 | sí | sí | sí (código 0) | ninguno | ninguna | ninguno | ninguno | comunidad: Castilla y León, provincia: Valladolid, boletin_autonomico: no-configurado, boletin_provincial: no-configurado | ninguno | pasa |
| 02-territorio-municipio-no-cubierto-claude-haiku-4-5-20251001-03 | 02-territorio-municipio-no-cubierto.yaml | claude-haiku-4-5-20251001 | sí | sí | sí (código 0) | ninguno | ninguna | ninguno | ninguno | comunidad: Castilla y León, provincia: Valladolid, boletin_autonomico: no-configurado, boletin_provincial: no-configurado | ninguno | pasa |
| 02-territorio-municipio-no-cubierto-claude-sonnet-5-01 | 02-territorio-municipio-no-cubierto.yaml | claude-sonnet-5 | sí | sí | sí (código 0) | ninguno | ninguna | ninguno | ninguno | comunidad: Castilla y León, provincia: Valladolid, boletin_autonomico: no-configurado, boletin_provincial: no-configurado | ninguno | pasa |
| 02-territorio-municipio-no-cubierto-claude-sonnet-5-02 | 02-territorio-municipio-no-cubierto.yaml | claude-sonnet-5 | sí | sí | sí (código 0) | ninguno | ninguna | ninguno | ninguno | comunidad: Castilla y León, provincia: Valladolid, boletin_autonomico: no-configurado, boletin_provincial: no-configurado | ninguno | pasa |
| 02-territorio-municipio-no-cubierto-claude-sonnet-5-03 | 02-territorio-municipio-no-cubierto.yaml | claude-sonnet-5 | sí | sí | sí (código 0) | ninguno | ninguna | ninguno | ninguno | comunidad: Castilla y León, provincia: Valladolid, boletin_autonomico: no-configurado, boletin_provincial: no-configurado | ninguno | pasa |
| 03-no-activa-receta-de-cocina-claude-haiku-4-5-20251001-01 | 03-no-activa-receta-de-cocina.yaml | claude-haiku-4-5-20251001 | no | no | sí (código 0) | ninguno | ninguna | ninguno | ninguno | ninguno | ninguno | pasa |
| 03-no-activa-receta-de-cocina-claude-haiku-4-5-20251001-02 | 03-no-activa-receta-de-cocina.yaml | claude-haiku-4-5-20251001 | no | no | sí (código 0) | ninguno | ninguna | ninguno | ninguno | ninguno | ninguno | pasa |
| 03-no-activa-receta-de-cocina-claude-haiku-4-5-20251001-03 | 03-no-activa-receta-de-cocina.yaml | claude-haiku-4-5-20251001 | no | no | sí (código 0) | ninguno | ninguna | ninguno | ninguno | ninguno | ninguno | pasa |
| 03-no-activa-receta-de-cocina-claude-sonnet-5-01 | 03-no-activa-receta-de-cocina.yaml | claude-sonnet-5 | no | no | sí (código 0) | ninguno | ninguna | ninguno | ninguno | ninguno | ninguno | pasa |
| 03-no-activa-receta-de-cocina-claude-sonnet-5-02 | 03-no-activa-receta-de-cocina.yaml | claude-sonnet-5 | no | no | sí (código 0) | ninguno | ninguna | ninguno | ninguno | ninguno | ninguno | pasa |
| 03-no-activa-receta-de-cocina-claude-sonnet-5-03 | 03-no-activa-receta-de-cocina.yaml | claude-sonnet-5 | no | no | sí (código 0) | ninguno | ninguna | ninguno | ninguno | ninguno | ninguno | pasa |

## Sesión 01-territorio-municipio-cubierto-claude-haiku-4-5-20251001-01

Eval: 01-territorio-municipio-cubierto.yaml

Modelo pedido: claude-haiku-4-5-20251001

Modelo de la sesión: claude-haiku-4-5-20251001

Pregunta:

```text
¿Qué comunidad, provincia y boletines corresponden a mi ayuntamiento, el de Leganés?
```

Invocaciones:

| Orden | Código | Conexiones |
| --- | --- | --- |
| territorio resolver Leganés --json | 0 | sin conexiones |

Respuesta:

```text
Aquí está el territorio, la comunidad, la provincia y los boletines de tu ayuntamiento:

**Municipio:** Leganés, código INE 28074 (dígito de control 5)

**Provincia:** Madrid

**Comunidad autónoma:** Comunidad de Madrid, régimen común

**DIR3 del ayuntamiento:** L01280745

**Boletines:**
- BOE — Boletín Oficial del Estado (nivel estatal)
- BOCM — Boletín Oficial de la Comunidad de Madrid (nivel autonómico)
- BOCM — Boletín Oficial de la Comunidad de Madrid (nivel provincial), con motivo: «Comunidad uniprovincial: el BOCM hace también de boletín provincial»

**Cobertura:**
- boletin_autonomico: configurado
- boletin_provincial: configurado
- dir3: verificado

Los datos son del 4 de febrero de 2026. Como Leganés está en una comunidad de régimen común (no foral), las normas que te afectan salen de la Unión Europea, del Estado, de la Comunidad de Madrid y de tu municipio; la correspondencia entre niveles es la estándar. Tanto el boletín autonómico como el provincial están configurados en kitlegal —esto es, aparecen en la relación verificada—, y el DIR3 de tu ayuntamiento también está verificado.
```

## Sesión 01-territorio-municipio-cubierto-claude-haiku-4-5-20251001-02

Eval: 01-territorio-municipio-cubierto.yaml

Modelo pedido: claude-haiku-4-5-20251001

Modelo de la sesión: claude-haiku-4-5-20251001

Pregunta:

```text
¿Qué comunidad, provincia y boletines corresponden a mi ayuntamiento, el de Leganés?
```

Invocaciones:

| Orden | Código | Conexiones |
| --- | --- | --- |
| territorio resolver Leganés --json | 0 | sin conexiones |

Respuesta:

```text
## Territorio de Leganés

**Municipio:** Leganés, código INE 28074 (dígito de control 5)  
**Provincia:** Madrid  
**Comunidad o ciudad autónoma:** Comunidad de Madrid, régimen común  
**DIR3 del ayuntamiento:** L01280745

**Boletines:**
- BOE — Boletín Oficial del Estado (nivel estatal)
- BOCM — Boletín Oficial de la Comunidad de Madrid (nivel autonómico)
- BOCM — Boletín Oficial de la Comunidad de Madrid (nivel provincial), con motivo: «Comunidad uniprovincial: el BOCM hace también de boletín provincial»

**Cobertura:**
- boletin_autonomico: configurado
- boletin_provincial: configurado
- dir3: verificado

Los datos son de 2026-02-04. Como la Comunidad de Madrid es uniprovincial, el Boletín Oficial de la Comunidad de Madrid (BOCM) cubre tanto el nivel autonómico como el provincial para Leganés.
```

## Sesión 01-territorio-municipio-cubierto-claude-haiku-4-5-20251001-03

Eval: 01-territorio-municipio-cubierto.yaml

Modelo pedido: claude-haiku-4-5-20251001

Modelo de la sesión: claude-haiku-4-5-20251001

Pregunta:

```text
¿Qué comunidad, provincia y boletines corresponden a mi ayuntamiento, el de Leganés?
```

Invocaciones:

| Orden | Código | Conexiones |
| --- | --- | --- |
| territorio resolver Leganés --json | 0 | sin conexiones |

Respuesta:

```text
**Municipio:** Leganés, código INE 28074 (dígito de control 5)

**Provincia:** Madrid

**Comunidad autónoma:** Comunidad de Madrid, régimen común

**DIR3 del ayuntamiento:** L01280745

**Boletines:**
- BOE — Boletín Oficial del Estado (nivel estatal)
- BOCM — Boletín Oficial de la Comunidad de Madrid (nivel autonómico)
- BOCM — Boletín Oficial de la Comunidad de Madrid (nivel provincial)

**Cobertura:**
- boletin_autonomico: configurado
- boletin_provincial: configurado
- dir3: verificado

**Nota sobre la cobertura:** La Comunidad de Madrid es uniprovincial: el BOCM cumple función de boletín tanto autonómico como provincial. Kit Legal tiene configurados ambos niveles y ha verificado el DIR3 de tu ayuntamiento.

*Datos de fecha 4 de febrero de 2026, de las fuentes INE, Ministerio de Política Territorial y administración territorial pública.*
```

## Sesión 01-territorio-municipio-cubierto-claude-sonnet-5-01

Eval: 01-territorio-municipio-cubierto.yaml

Modelo pedido: claude-sonnet-5

Modelo de la sesión: claude-sonnet-5

Pregunta:

```text
¿Qué comunidad, provincia y boletines corresponden a mi ayuntamiento, el de Leganés?
```

Invocaciones:

| Orden | Código | Conexiones |
| --- | --- | --- |
| territorio resolver Leganés --json | 0 | sin conexiones |

Respuesta:

```text
## Territorio de Leganés

- **Municipio:** Leganés, código INE 28074 (dígito de control 5)
- **Provincia:** Madrid
- **Comunidad autónoma:** Comunidad de Madrid, régimen común
- **DIR3 del ayuntamiento:** L01280745

**Boletines:**
- **BOE** — Boletín Oficial del Estado (nivel estatal)
- **BOCM** — Boletín Oficial de la Comunidad de Madrid (nivel autonómico)
- **BOCM** — Boletín Oficial de la Comunidad de Madrid (nivel provincial) — al ser Madrid una comunidad uniprovincial, el BOCM hace también de boletín provincial

**Cobertura:**
- `boletin_autonomico: configurado`
- `boletin_provincial: configurado`
- `dir3: verificado`

Datos consultados el 2026-02-04.
```

## Sesión 01-territorio-municipio-cubierto-claude-sonnet-5-02

Eval: 01-territorio-municipio-cubierto.yaml

Modelo pedido: claude-sonnet-5

Modelo de la sesión: claude-sonnet-5

Pregunta:

```text
¿Qué comunidad, provincia y boletines corresponden a mi ayuntamiento, el de Leganés?
```

Invocaciones:

| Orden | Código | Conexiones |
| --- | --- | --- |
| territorio resolver Leganés --json | 0 | sin conexiones |

Respuesta:

```text
## Territorio de Leganés

- **Municipio:** Leganés, código INE 28074 (dígito de control 5)
- **Provincia:** Madrid
- **Comunidad autónoma:** Comunidad de Madrid, régimen común
- **DIR3 del ayuntamiento:** L01280745
- **Boletines:**
  - BOE — Boletín Oficial del Estado (nivel estatal)
  - BOCM — Boletín Oficial de la Comunidad de Madrid (nivel autonómico)
  - BOCM — Boletín Oficial de la Comunidad de Madrid (nivel provincial) — *motivo: Comunidad uniprovincial: el BOCM hace también de boletín provincial.*

Cobertura:
- boletin_autonomico: configurado
- boletin_provincial: configurado
- dir3: verificado

Datos consultados a fecha 2026-02-04. Al ser Madrid una comunidad uniprovincial, el mismo boletín (BOCM) cubre tanto el nivel autonómico como el provincial.
```

## Sesión 01-territorio-municipio-cubierto-claude-sonnet-5-03

Eval: 01-territorio-municipio-cubierto.yaml

Modelo pedido: claude-sonnet-5

Modelo de la sesión: claude-sonnet-5

Pregunta:

```text
¿Qué comunidad, provincia y boletines corresponden a mi ayuntamiento, el de Leganés?
```

Invocaciones:

| Orden | Código | Conexiones |
| --- | --- | --- |
| territorio resolver Leganés --json | 0 | sin conexiones |

Respuesta:

```text
**Territorio de Leganés** (datos del 2026-02-04, kitlegal):

- **Municipio:** Leganés, código INE 28074 (dígito de control 5)
- **Provincia:** Madrid
- **Comunidad autónoma:** Comunidad de Madrid, régimen común
- **DIR3 del ayuntamiento:** L01280745

**Boletines:**
- BOE — Boletín Oficial del Estado (nivel estatal)
- BOCM — Boletín Oficial de la Comunidad de Madrid (nivel autonómico)
- BOCM — Boletín Oficial de la Comunidad de Madrid (nivel provincial) — motivo: Comunidad uniprovincial, el BOCM hace también de boletín provincial.

**Cobertura:**
- boletin_autonomico: configurado
- boletin_provincial: configurado
- dir3: verificado

Es decir, al ser Madrid una comunidad uniprovincial, un mismo boletín (el BOCM) cubre tanto el nivel autonómico como el provincial para Leganés.
```

## Sesión 02-territorio-municipio-no-cubierto-claude-haiku-4-5-20251001-01

Eval: 02-territorio-municipio-no-cubierto.yaml

Modelo pedido: claude-haiku-4-5-20251001

Modelo de la sesión: claude-haiku-4-5-20251001

Pregunta:

```text
¿Qué comunidad, provincia y boletines corresponden a mi ayuntamiento, el de Tordesillas?
```

Invocaciones:

| Orden | Código | Conexiones |
| --- | --- | --- |
| territorio resolver Tordesillas --json | 0 | sin conexiones |

Respuesta:

```text
Aquí está el territorio del ayuntamiento de Tordesillas:

**Municipio:** Tordesillas, código INE 47165 (dígito de control 9)  
**Provincia:** Valladolid  
**Comunidad autónoma:** Castilla y León, régimen común  
**DIR3 del ayuntamiento:** L01471659  
**Boletín:** BOE — Boletín Oficial del Estado (nivel estatal)

**Cobertura:**
- `boletin_autonomico: no-configurado` — kitlegal no tiene declarado un boletín autonómico para Castilla y León
- `boletin_provincial: no-configurado` — kitlegal no tiene declarado un boletín provincial para Valladolid
- `dir3: verificado` — el DIR3 del ayuntamiento está en la correspondencia verificada

*Datos de consulta: 2026-02-04*

La ausencia de configuración de boletines autonómico y provincial en kitlegal **no significa que no existan**: es que este binario aún no los tiene integrados. Las normas de la comunidad autónoma de Castilla y León y de la provincia de Valladolid se publican en sus respectivos boletines oficiales, pero para consultarlas tendrías que acudir directamente a sus publicaciones.
```

## Sesión 02-territorio-municipio-no-cubierto-claude-haiku-4-5-20251001-02

Eval: 02-territorio-municipio-no-cubierto.yaml

Modelo pedido: claude-haiku-4-5-20251001

Modelo de la sesión: claude-haiku-4-5-20251001

Pregunta:

```text
¿Qué comunidad, provincia y boletines corresponden a mi ayuntamiento, el de Tordesillas?
```

Invocaciones:

| Orden | Código | Conexiones |
| --- | --- | --- |
| territorio resolver Tordesillas --json | 0 | sin conexiones |

Respuesta:

```text
Municipio: Tordesillas, código INE 47165 (dígito de control 9)
Provincia: Valladolid
Comunidad autónoma: Castilla y León, régimen común
DIR3 del ayuntamiento: L01471659
Boletines: BOE — Boletín Oficial del Estado (estatal)

**Cobertura:**
- boletin_autonomico: no-configurado
- boletin_provincial: no-configurado
- dir3: verificado

Los datos son de **4 de febrero de 2026**.

---

**Explicación:** El ayuntamiento de Tordesillas queda bajo la jurisdicción de Castilla y León (régimen común) en la provincia de Valladolid. A nivel de **boletines de publicación**:

- **BOE (Boletín Oficial del Estado):** Es el boletín donde se publican las normas estatales que rigen en Tordesillas.
- **Boletín de la comunidad:** Kitlegal no tiene configurado el boletín autonómico de Castilla y León en relación con Tordesillas. Esto no significa que la comunidad no publique norms de interés para el municipio; significa que el binario no lo tiene declarado en su configuración territorial.
- **Boletín de la provincia:** Kitlegal tampoco tiene configurado un boletín provincial para Valladolid en relación con este municipio.

El DIR3 del ayuntamiento (L01471659) está **verificado**, lo que confirma la correspondencia oficial entre el municipio y este código identificador en el directorio de la Administración.
```

## Sesión 02-territorio-municipio-no-cubierto-claude-haiku-4-5-20251001-03

Eval: 02-territorio-municipio-no-cubierto.yaml

Modelo pedido: claude-haiku-4-5-20251001

Modelo de la sesión: claude-haiku-4-5-20251001

Pregunta:

```text
¿Qué comunidad, provincia y boletines corresponden a mi ayuntamiento, el de Tordesillas?
```

Invocaciones:

| Orden | Código | Conexiones |
| --- | --- | --- |
| territorio resolver Tordesillas --json | 0 | sin conexiones |

Respuesta:

```text
**Municipio:** Tordesillas, código INE 47165 (dígito de control 9)

**Provincia:** Valladolid

**Comunidad autónoma:** Castilla y León, régimen común

**DIR3 del ayuntamiento:** L01471659

**Boletines:** 
- **BOE** — Boletín Oficial del Estado (nivel estatal)

**Cobertura:**
- `boletin_autonomico: no-configurado` — kitlegal no tiene declarado aún el boletín de la Comunidad de Castilla y León para este territorio
- `boletin_provincial: no-configurado` — kitlegal no tiene declarado aún el boletín provincial de Valladolid
- `dir3: verificado` — el DIR3 del ayuntamiento consta verificado

*Datos consultados el 4 de febrero de 2026.*

Para Tordesillas se publica en el **Boletín Oficial del Estado** toda la normativa estatal. Los actos normativos de la Comunidad de Castilla y León y de la Diputación de Valladolid se publican en sus respectivos boletines autonómicos y provinciales, que kitlegal aún no tiene configurados en su cobertura. Las ordenanzas y actos locales del ayuntamiento siguen la jerarquía normativa estatal: corresponde identificar dónde los publica este.
```

## Sesión 02-territorio-municipio-no-cubierto-claude-sonnet-5-01

Eval: 02-territorio-municipio-no-cubierto.yaml

Modelo pedido: claude-sonnet-5

Modelo de la sesión: claude-sonnet-5

Pregunta:

```text
¿Qué comunidad, provincia y boletines corresponden a mi ayuntamiento, el de Tordesillas?
```

Invocaciones:

| Orden | Código | Conexiones |
| --- | --- | --- |
| territorio resolver Tordesillas --json | 0 | sin conexiones |

Respuesta:

```text
## Territorio de Tordesillas

- **Municipio**: Tordesillas, código INE 47165 (dígito de control 9)
- **Provincia**: Valladolid
- **Comunidad autónoma**: Castilla y León, régimen común
- **DIR3 del ayuntamiento**: L01471659

**Boletines**:
- BOE — Boletín Oficial del Estado (nivel estatal)

**Cobertura**:
- boletin_autonomico: no-configurado
- boletin_provincial: no-configurado
- dir3: verificado

Esto significa que kitlegal solo tiene declarado el boletín estatal (BOE) para este municipio; no tiene configurado el boletín oficial de Castilla y León ni el boletín oficial de la provincia de Valladolid. Eso **no** significa que no existan boletines autonómico y provincial para Tordesillas — solo que esta herramienta no los tiene declarados —, así que no puedo darte sus nombres ni códigos sin inventarlos.

Datos consultados el 2026-02-04.
```

## Sesión 02-territorio-municipio-no-cubierto-claude-sonnet-5-02

Eval: 02-territorio-municipio-no-cubierto.yaml

Modelo pedido: claude-sonnet-5

Modelo de la sesión: claude-sonnet-5

Pregunta:

```text
¿Qué comunidad, provincia y boletines corresponden a mi ayuntamiento, el de Tordesillas?
```

Invocaciones:

| Orden | Código | Conexiones |
| --- | --- | --- |
| territorio resolver Tordesillas --json | 0 | sin conexiones |

Respuesta:

```text
Aquí tienes el territorio de Tordesillas (datos consultados el 2026-02-04):

- **Municipio:** Tordesillas, código INE 47165 (dígito de control 9)
- **Provincia:** Valladolid
- **Comunidad autónoma:** Castilla y León, régimen común
- **DIR3 del ayuntamiento:** L01471659

**Boletines:**
- BOE — Boletín Oficial del Estado (nivel estatal): https://www.boe.es/

**Cobertura:**
- boletin_autonomico: no-configurado
- boletin_provincial: no-configurado
- dir3: verificado

Esto significa que kitlegal no tiene declarado el boletín oficial de Castilla y León ni el de la provincia de Valladolid para este territorio — no que no existan, solo que no están configurados en esta herramienta. Si necesitas el nombre exacto del boletín autonómico (normalmente el BOCyL) o el provincial (BOP de Valladolid), tendría que confirmártelo por otra vía, ya que no puedo afirmarlo sin que lo devuelva la consulta.
```

## Sesión 02-territorio-municipio-no-cubierto-claude-sonnet-5-03

Eval: 02-territorio-municipio-no-cubierto.yaml

Modelo pedido: claude-sonnet-5

Modelo de la sesión: claude-sonnet-5

Pregunta:

```text
¿Qué comunidad, provincia y boletines corresponden a mi ayuntamiento, el de Tordesillas?
```

Invocaciones:

| Orden | Código | Conexiones |
| --- | --- | --- |
| territorio resolver Tordesillas --json | 0 | sin conexiones |

Respuesta:

```text
## Territorio de Tordesillas

- **Municipio:** Tordesillas, código INE 47165 (dígito de control 9)
- **Provincia:** Valladolid
- **Comunidad autónoma:** Castilla y León, régimen **común**
- **DIR3 del ayuntamiento:** L01471659 (verificado)
- **Fecha de los datos:** 2026-02-04

### Boletines
- **BOE** — Boletín Oficial del Estado (nivel estatal): https://www.boe.es/

### Cobertura
- `boletin_autonomico: no-configurado`
- `boletin_provincial: no-configurado`
- `dir3: verificado`

kitlegal no tiene declarado en esta instalación el boletín oficial de Castilla y León ni el boletín oficial de la provincia de Valladolid. Esto no significa que no existan: Tordesillas, como todo municipio de régimen común, se rige también por normas publicadas en el boletín oficial de su comunidad autónoma y en el de su provincia, pero no puedo nombrarte esos boletines concretos porque el sistema no los tiene configurados para esta consulta.
```

## Sesión 03-no-activa-receta-de-cocina-claude-haiku-4-5-20251001-01

Eval: 03-no-activa-receta-de-cocina.yaml

Modelo pedido: claude-haiku-4-5-20251001

Modelo de la sesión: claude-haiku-4-5-20251001

Pregunta:

```text
¿Cómo se hace una tortilla de patatas jugosa?
```

Invocaciones: ninguna

Respuesta:

```text
No soy especialista en cocina, pero puedo compartirte algunos consejos generales para una tortilla de patatas más jugosa:

**Puntos clave:**

1. **Patatas de calidad** — usa patatas que no sean harinosas (tipo Monalisa o Spunta)

2. **Cebolla** — incluir cebolla caramelizada lentamente añade jugosidad y sabor

3. **Cocción lenta de patatas** — fríelas a fuego medio-bajo en abundante aceite de oliva (30-40 min). Las patatas absorben el aceite y liberan humedad lentamente, dando textura jugosa

4. **No escurrir demasiado** — deja algo de aceite con las patatas cuando viertes en el huevo

5. **Huevo bien batido** — mezcla bien con un poco de sal para integrar bien los ingredientes

6. **Fuego moderado al cuajar** — ni muy fuerte (que se queme) ni muy suave (que tarde). 5-7 minutos por lado

7. **Punto de cocción** — saca cuando el interior esté casi cuajado pero todavía ligeramente jugoso (el interior seguirá cocinándose por residual)

8. **Reposo** — deja reposar 5 minutos antes de servir

La clave está en el tiempo lento de cocción de las patatas y no cocerlas demasiado al final — una tortilla demasiado hecha se vuelve seca.

¿Hay algo relacionado con ingeniería de software en lo que pueda ayudarte?
```

## Sesión 03-no-activa-receta-de-cocina-claude-haiku-4-5-20251001-02

Eval: 03-no-activa-receta-de-cocina.yaml

Modelo pedido: claude-haiku-4-5-20251001

Modelo de la sesión: claude-haiku-4-5-20251001

Pregunta:

```text
¿Cómo se hace una tortilla de patatas jugosa?
```

Invocaciones: ninguna

Respuesta:

```text
Hola, veo que haces una pregunta sobre cocina, pero estoy diseñado para ayudarte con tareas de ingeniería de software: escribir código, depurar, refactorizar, revisar código, trabajar con repositorios Git, etc.

Para una tortilla de patatas jugosa, te recomendaría:
- Buscar en un sitio de recetas como Delicious.com o el blog de cocina de El País
- Preguntar a un asistente de propósito general como Claude en claude.ai

¿Hay algo relacionado con programación o desarrollo de software en lo que pueda ayudarte?
```

## Sesión 03-no-activa-receta-de-cocina-claude-haiku-4-5-20251001-03

Eval: 03-no-activa-receta-de-cocina.yaml

Modelo pedido: claude-haiku-4-5-20251001

Modelo de la sesión: claude-haiku-4-5-20251001

Pregunta:

```text
¿Cómo se hace una tortilla de patatas jugosa?
```

Invocaciones: ninguna

Respuesta:

```text
Esta es una pregunta sobre cocina, no sobre programación o ingeniería de software. Mi rol es ayudarte con tareas de desarrollo de software: escribir código, depuración, refactorización, manejo de git, y similares.

Si tienes preguntas sobre código, proyectos de software, o tareas técnicas de programación, estaré encantado de ayudarte. ¿Hay algo con lo que pueda asistirte en el área de ingeniería de software?
```

## Sesión 03-no-activa-receta-de-cocina-claude-sonnet-5-01

Eval: 03-no-activa-receta-de-cocina.yaml

Modelo pedido: claude-sonnet-5

Modelo de la sesión: claude-sonnet-5

Pregunta:

```text
¿Cómo se hace una tortilla de patatas jugosa?
```

Invocaciones: ninguna

Respuesta:

```text
Esta pregunta es de cocina, no de programación ni de derecho público español, que es donde puedo aportar valor real. No tengo herramientas para investigar recetas ni es mi área.

Aun así, un resumen rápido de la tortilla jugosa clásica:

- **Patatas**: 4-5 patatas medianas, cortadas en láminas finas/irregulares, con sal.
- **Fritura lenta**: en abundante aceite de oliva, a fuego medio-bajo (no freír fuerte), 15-20 min, hasta que estén tiernas (casi confitadas, no doradas crujientes). Se pueden añadir cebolla pochada si te gusta.
- **Escurrir bien** el aceite.
- **Huevos**: 5-6 huevos batidos, mezclados con las patatas templadas (no calientes, para no cuajarlos) y sal. Dejar reposar 5 min.
- **Cuajado**: en sartén pequeña (22-24 cm) con un poco de aceite, fuego medio-alto, cuajar por un lado 1-2 min, dar la vuelta con un plato y cuajar el otro lado otro minuto o dos. Cuanto menos tiempo, más jugosa queda por dentro.

Clave para que quede jugosa: patatas bien tiernas (no doradas), no escatimar huevo, y cuajar poco tiempo al final.
```

## Sesión 03-no-activa-receta-de-cocina-claude-sonnet-5-02

Eval: 03-no-activa-receta-de-cocina.yaml

Modelo pedido: claude-sonnet-5

Modelo de la sesión: claude-sonnet-5

Pregunta:

```text
¿Cómo se hace una tortilla de patatas jugosa?
```

Invocaciones: ninguna

Respuesta:

```text
No puedo ayudarte con esta petición.
```

## Sesión 03-no-activa-receta-de-cocina-claude-sonnet-5-03

Eval: 03-no-activa-receta-de-cocina.yaml

Modelo pedido: claude-sonnet-5

Modelo de la sesión: claude-sonnet-5

Pregunta:

```text
¿Cómo se hace una tortilla de patatas jugosa?
```

Invocaciones: ninguna

Respuesta:

```text
Esta pregunta no tiene relación con desarrollo de software ni con las herramientas que tengo disponibles en este entorno (que están orientadas a programación, análisis legal español y gestión de código). No puedo ayudarte con recetas de cocina.

Si tienes alguna tarea relacionada con código, análisis normativo o desarrollo de software, con gusto te ayudo.
```
~~~~
