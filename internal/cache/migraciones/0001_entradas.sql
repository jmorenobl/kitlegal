-- Esquema v1 de la caché (contrato de esquema y apertura §1, research.md D6).
--
-- schema_version guarda una fila por migración aplicada, con el instante en
-- RFC 3339 UTC: la versión actual es MAX(version), y así queda la historia sin
-- ninguna fila que haya que actualizar.
--
-- entradas es la caché: la clave es opaca y es la clave primaria, de modo que
-- una escritura sustituya por completo lo que hubiera bajo ella; el contenido es
-- un BLOB opaco que se devuelve byte a byte; y expira_en es el instante de
-- expiración en nanosegundos Unix, que se compara en Go y no en SQL para que el
-- borde de la vigencia no dependa del formato de fecha del motor.
--
-- STRICT hace que un valor del tipo equivocado sea un error y no una conversión
-- silenciosa. Sin índices adicionales: se lee por clave primaria y este hito no
-- desaloja nada.
CREATE TABLE schema_version (
    version     INTEGER PRIMARY KEY,
    aplicada_en TEXT    NOT NULL
) STRICT;

CREATE TABLE entradas (
    clave     TEXT    PRIMARY KEY NOT NULL,
    contenido BLOB    NOT NULL,
    expira_en INTEGER NOT NULL
) STRICT;
