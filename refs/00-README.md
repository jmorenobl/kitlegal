# kitlegal: documentos semilla

Tres documentos de diseño para arrancar el proyecto. Léelos en este orden.

1. `mapa-sistema-legal-skills.md` — Qué es el sistema legal español (jerarquía, competencias, jurisdicción, vía administrativa), dónde está cada tipo de información (fuentes con semáforo de automatizabilidad) y el catálogo de skills agénticas con sus comandos.
2. `kitlegal-estructura-y-ecosistema.md` — Estructura de directorios del monorepo Go, convenciones del binario multicall `kitlegal`, buenas prácticas, packs por vertical y superficies de distribución.
3. `kitlegal-grafo.md` — El grafo legal (mundo + asunto): modelo sobre ELI, almacenamiento SQLite, alimentación automática por uso, consulta, validación (`graph check`) y detección de anomalías.

## Estado de partida

- No existe código previo. Nada de lo descrito está implementado.
- La única pieza existente es la skill `boe-fiscal` (Python, `scripts/boe.py` + `references/boe_api.md` + `references/normas_fiscales.md`), que sirve de patrón y se porta a Go como applet `boe`.
- Nombre fijado: `kitlegal` para repo, módulo (`github.com/jmorenobl/kitlegal`), binario y directorios (`.kitlegal/`, `~/.cache/kitlegal/`).

## Primer hito

`kitlegal boe articulo BOE-A-2015-10565 a21` funcionando en Go con caché SQLite, y la skill `boe-fiscal` migrada para invocar el binario en lugar de `boe.py`, con el mismo comportamiento que hoy. Todo lo demás se construye encima.

## Decisiones ya tomadas

- Skills sin código: `SKILL.md` + `references/` generadas desde `data/*.yaml`; `scripts/` son symlinks al binario.
- Toda salida en JSON con `fuente`, `url`, `fecha_consulta`, `hash`. Sin eso no hay cita.
- No se abstrae `internal/cli` a librería externa hasta que tres applets repitan el patrón.
- Solo se automatizan fuentes públicas; presentar escritos, notificaciones y cualquier acción con identidad requieren firma humana (exit code 6).
- CENDOJ: solo resolución de ECLI y metadatos; nunca ingesta masiva.
- El grafo del asunto (`.kitlegal/` en cwd) nunca se sincroniza ni exporta por defecto.

## Pendiente de verificar antes de codificar

- Identificadores BOE de la tabla de leyes vertebrales (`kitlegal boe buscar`).
- Endpoints actuales de BDNS (swagger) y nombres de los feeds Atom de PLACSP.
- Términos de uso de PETETE (DGT), DYCTEA (TEAC) y buscador HJ del TC.
