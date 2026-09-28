-- Versión 1 del esquema de world.db, el grafo del mundo (H7: data-model §3,
-- research.md D12). Se aplica dentro de la transacción que la pide, junto con
-- la fila de schema_version que la registra: entra entera o no entra nada
-- (FR-013). Todas las tablas son STRICT, de modo que un valor de otro tipo que
-- el declarado es un fallo y no una conversión silenciosa.

CREATE TABLE schema_version (
    version     INTEGER PRIMARY KEY,
    aplicada_en TEXT    NOT NULL
) STRICT;

CREATE TABLE nodes (
    id           TEXT    PRIMARY KEY NOT NULL,
    type         TEXT    NOT NULL,
    props        TEXT    NOT NULL,   -- datos identificativos de la última observación, JSON canónico (RFC 8785)
    first_seen   TEXT    NOT NULL,   -- fecha_consulta de la primera observación, tal como la escribió su sobre
    first_source TEXT    NOT NULL,   -- fuente y url de la primera: solo para el desempate de FR-023
    first_url    TEXT    NOT NULL,
    last_seen    TEXT    NOT NULL,   -- fecha_consulta de la última observación
    source       TEXT    NOT NULL,   -- fuente de la última
    url          TEXT    NOT NULL,   -- url de la última
    ttl          INTEGER             -- vigencia de la última en segundos; NULL si no se declaró
) STRICT;
CREATE INDEX nodes_type_source ON nodes(type, source);

CREATE TABLE edges (
    src          TEXT    NOT NULL REFERENCES nodes(id),
    rel          TEXT    NOT NULL,
    dst          TEXT    NOT NULL REFERENCES nodes(id),
    first_seen   TEXT    NOT NULL,
    first_source TEXT    NOT NULL,
    first_url    TEXT    NOT NULL,
    last_seen    TEXT    NOT NULL,
    source       TEXT    NOT NULL,
    url          TEXT    NOT NULL,
    ttl          INTEGER,
    PRIMARY KEY (src, rel, dst)
) STRICT;
CREATE INDEX edges_dst_rel ON edges(dst, rel);

CREATE TABLE texts (
    hash       TEXT PRIMARY KEY NOT NULL,   -- sha256:<hex>, el hash_texto del sobre
    body       TEXT NOT NULL,               -- nunca sale por ningún verbo (FR-070)
    fetched_at TEXT NOT NULL,               -- fecha_consulta de la observación más antigua
    source     TEXT NOT NULL,
    url        TEXT NOT NULL
) STRICT;
