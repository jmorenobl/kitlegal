-- Versión 2 del esquema de world.db: las lecturas de cada bloque (H7.1
-- data-model §1, research.md D1). El grafo no guarda la historia de las
-- lecturas, sino, por bloque, las dos únicas que deciden algo: qué redacción
-- vio la última y cuál la anterior. Se aplica dentro de la transacción de la
-- entrega que encuentra la base en la versión 0 o 1, junto con la fila de
-- schema_version que la registra: entra entera o no entra nada (H7 FR-013).

CREATE TABLE lecturas (
    bloque   TEXT PRIMARY KEY NOT NULL REFERENCES nodes(id), -- el Bloque leído
    ultima   TEXT NOT NULL REFERENCES nodes(id),             -- la BloqueVersion que vio su última lectura
    anterior TEXT NOT NULL REFERENCES nodes(id)              -- la que vio la lectura anterior; en la primera, la misma
) STRICT;
