# kitlegal: documentos semilla

Tres documentos de diseño para arrancar el proyecto. Léelos en este orden.

**Enfoque (2026-09-11).** El producto son las skills; el binario Go es la capa de herramientas deterministas que usan. Todo es genérico para cualquier municipio de España y se valida primero en la Comunidad de Madrid con Leganés; los demás territorios llegan después como datos. Ver la constitución (principios VIII y IX) y los ADR 0008 y 0009; ante conflicto, prevalecen sobre estos documentos. Las skills nacen genéricas y las verticales (fiscal, laboral, mercantil…) se especializan después sobre la base (ADR 0012): donde estos documentos ponen `boe-fiscal` como primera skill, léase `boe-legislacion`. El orden de construcción es por tiempo hasta el uso (ADR 0013): actuar en el municipio va antes que consultar sus fuentes, y la fiscalización y la vigilancia son backlog. El grafo llega en tres piezas (ADR 0014): donde `kitlegal-grafo.md` dice que «llega cuando las skills del municipio ya existen» y que cada applet implementa `Emit(ctx) []GraphOp`, léase que la memoria (G0) llega en H7 justo después de `territorio`, el asunto (G1) en H10, el instrumento (G2) en el backlog, y que las operaciones viajan en el `Resultado` del applet con su `Procedencia` como `source`. El grafo es índice y validador, nunca fuente del texto que se cita.

1. `mapa-sistema-legal-skills.md` — Qué es el sistema legal español (jerarquía, competencias, jurisdicción, vía administrativa), dónde está cada tipo de información (fuentes con semáforo de automatizabilidad) y el catálogo de skills agénticas con sus comandos.
2. `kitlegal-estructura-y-ecosistema.md` — Estructura del monorepo (skills, packs y datos; binario multicall `kitlegal` como capa de herramientas), buenas prácticas, packs por vertical y superficies de distribución.
3. `kitlegal-grafo.md` — El grafo legal (mundo + asunto): modelo sobre ELI, almacenamiento SQLite, alimentación automática por uso, consulta, validación (`graph check`) y detección de anomalías.

## Estado de partida

- No existe código previo. Nada de lo descrito está implementado.
- La única pieza existente es la skill `boe-fiscal` (Python, `scripts/boe.py` + `references/boe_api.md` + `references/normas_fiscales.md`), que sirve de patrón: `boe.py` se porta a Go como applet `boe` y su protocolo, generalizado a cualquier materia, es el de la skill base `boe-legislacion`. `boe-fiscal` se migra después como primera vertical (ADR 0012).
- `boe.py` está aquí como cuarto documento semilla: **`refs/boe.py`**, copia congelada del 2026-05-14 (la versión viva de la skill, con los TTL de caché afinados a 7 días para `indice`, `articulo` y `analisis`). Es material de lectura, no código del producto: no se mantiene y no entra en ningún gate. Tampoco se ejecuta dentro del repositorio, con una única excepción: la pausa `[datos]` de H4 que deriva las referencias del diff de aceptación, donde un guion de un solo uso —fuera del repositorio, sin versionar y sin modificar este fichero— lo importa y sustituye su única puerta a la red por la lectura de las grabaciones, para que la referencia salga de la lógica original y de los mismos bytes que ven los tests, y no de la lectura que alguien haga de este fuente. Ni el producto, ni los tests, ni `make ci` necesitan Python. Su dependencia `network_utils.robust_request` **no se porta**: reintentos, ritmo y `robots.txt` los cubre `internal/httpx` (H2).
- Nombre fijado: `kitlegal` para repo, módulo (`github.com/jmorenobl/kitlegal`), binario y directorios (`.kitlegal/`, `~/.cache/kitlegal/`).

## Primer hito

`kitlegal boe articulo BOE-A-2015-10565 a21` funcionando en Go con caché SQLite, y la skill genérica `boe-legislacion` consultando y citando cualquier norma consolidada del BOE con el binario. Todo lo demás se construye encima; `boe-fiscal` se migra sobre esa base como primera vertical, cuando se quiera.

## Decisiones ya tomadas

- Skills primero: cada hito entrega o mejora una skill medible con evals; solo va a Go lo que exige determinismo o verificabilidad.
- Skills sin código: `SKILL.md` + `references/` generadas desde `data/*.yaml`; `scripts/` son symlinks al binario.
- Genericidad territorial: nada se particulariza para un municipio; datos de municipio desde registros nacionales, lo territorial por comunidad o boletín, y cobertura declarada fuera del territorio de validación.
- Toda salida en JSON con `fuente`, `url`, `fecha_consulta`, `hash`. Sin eso no hay cita.
- No se abstrae `internal/cli` a librería externa hasta que tres applets repitan el patrón.
- Solo se automatizan fuentes públicas; presentar escritos, notificaciones y cualquier acción con identidad requieren firma humana (exit code 6).
- CENDOJ: solo resolución de ECLI y metadatos; nunca ingesta masiva.
- El grafo del asunto (`.kitlegal/` en cwd) nunca se sincroniza ni exporta por defecto.

## Pendiente de verificar antes de codificar

- Identificadores BOE de la tabla de leyes vertebrales (`kitlegal boe buscar`).
- Endpoints actuales de BDNS (swagger) y nombres de los feeds Atom de PLACSP.
- Relación de municipios del INE, inventario DIR3 y festivos locales de la Comunidad de Madrid (formato, licencia, actualización).
- Términos de uso de PETETE (DGT), DYCTEA (TEAC) y buscador HJ del TC.
