# Evidencia de H6

Material de origen del que salen `data/territorio/dir3.yaml` y su verificación (ADR 0017, ADR 0018). Desde el
workflow 2.0.0 esta carpeta la escribe el paso `grabar_datos`; la de H6 es anterior y se trajo el 2026-09-25 desde la
copia local `~/.local/share/kitlegal/evidencias/h6/`, comprobando cada huella contra
`specs/008-h6-territorio-skill-legal/gates/verificacion-dir3.md`. `manifiesto.json` da, por fichero, su huella
SHA-256, su tamaño, la fila de `docs/SOURCES.md` de la que sale, la fecha de descarga y la dirección.

| Fichero | Fuente | Qué es |
|---|---|---|
| `rel.xls` | `mpt.rel` | Volcado de municipios del Registro de Entidades Locales (2026-09-21): el número de inscripción del que se deriva el DIR3 de cada ayuntamiento |
| `eatimes.xls` | `mpt.rel` | Volcado de entidades de ámbito territorial inferior al municipio (2026-09-22): el caso de FR-046 con entidades locales menores |

Falta, y por qué:

- **Las siete fichas del Punto de Acceso General** (`pag.directorio`) con las que se verificó la derivación contra DIR3
  real: su fila de `docs/SOURCES.md` está pendiente de revisión humana (términos de uso y licencia), y sin ella no se
  versionan (constitución, capa 3). Sus huellas están en `gates/verificacion-dir3.md` y la copia, fuera del
  repositorio.
- **`diccionario26.xlsx`, `cod_ccaa.htm` y `cod_provincia.htm` del INE**: se perdieron del directorio temporal de
  sesión antes de copiarlos. Sus huellas están en `gates/tarea-T001.md` y `gates/tarea-T003.md`; si una descarga
  nueva da la misma huella, el fichero se recupera aquí.
