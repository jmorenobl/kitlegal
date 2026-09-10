# 0002 · SQLite sin cgo (`modernc.org/sqlite`)

- **Estado**: aceptada
- **Fecha**: 2026-09-10
- **Hito**: H0

## Contexto y problema

`kitlegal` necesita persistencia local en tres sitios distintos y por tres motivos distintos:

- **Caché** de respuestas de fuentes, con TTL por fuente, para no repetir peticiones contra el BOE o
  PLACSP y para que `--offline` pueda responder sin red.
- **Store** de los datos ya normalizados que un applet devuelve.
- **Grafo** legal sobre ELI: `~/.cache/kitlegal/world.db` para el mundo público y `.kitlegal/case.db`
  para el asunto en curso. Ahí sí hacen falta consultas relacionales de verdad —recorridos,
  comprobaciones de anomalías, `graph check`—, no solo lectura por clave.

El entorno de ejecución condiciona la elección más que el volumen de datos:

- La herramienta se instala en portátiles de los tres sistemas operativos y en sandboxes de agentes,
  donde no se puede exigir un servicio corriendo ni permisos de administrador.
- El release se cross-compila con goreleaser para darwin/linux/windows × arm64/amd64, y la gracia del
  multicall ([ADR 0001](0001-multicall.md)) es que se descarga un fichero y funciona.
- La integración continua debe poder ejecutar `make ci` sin instalar dependencias del sistema.

## Opciones consideradas

1. **`mattn/go-sqlite3`** (binding cgo del SQLite original). Es la implementación de referencia, la
   más rápida y la más probada. A cambio exige cgo: toolchain C presente en cada máquina que
   construya, cross-compilación con compilador cruzado por plataforma y adiós al binario estático.
2. **`modernc.org/sqlite`**: SQLite traducido a Go, sin cgo, misma semántica y mismo SQL.
3. **Un servidor externo** (PostgreSQL). Da concurrencia y herramientas de sobra, pero convierte una
   CLI local en un despliegue: nadie va a levantar un servidor para consultar un artículo del BOE.
4. **Ficheros planos o un almacén clave-valor** (JSON en disco, BoltDB). Sirve para la caché y poco
   más: el grafo necesita `JOIN`, índices y consultas ad-hoc, y reimplementarlos encima de un
   clave-valor es escribir media base de datos.

## Decisión

Se adopta la **opción 2**: `modernc.org/sqlite`, sin cgo, como única dependencia de persistencia
para caché, store y grafo.

- Los binarios se construyen y se publican con `CGO_ENABLED=0`; el resultado es estático y se
  cross-compila sin compilador C.
- Solo `internal/cache`, `internal/store` e `internal/graph` importan `modernc.org/sqlite` y
  `database/sql`. El resto del árbol trabaja contra las interfaces que esos paquetes exponen, de
  modo que el motor concreto queda encapsulado.
- Esquema versionado con migraciones embebidas (`embed` + tabla `schema_version`), WAL y `PRAGMA`
  seguros, iguales en todas las plataformas.

## Consecuencias

**A favor**

- Un binario estático por plataforma, sin toolchain C en la máquina del usuario ni en la de
  construcción; el release cross-compila sin infraestructura extra.
- La integración continua no necesita paquetes del sistema: el mismo `make ci` vale en local y en la
  plataforma, que es justamente el criterio de equivalencia que exige H0.
- Mismo motor, mismo SQL y mismo fichero en macOS, Linux y Windows: un `world.db` es legible desde
  cualquiera de ellos, y con cualquier cliente `sqlite3` estándar.
- Una sola dependencia de persistencia para tres usos, en vez de un almacén distinto por uso.

**En contra, y asumido**

- Rendimiento por debajo del SQLite en C. Es aceptable: los volúmenes son pequeños y el coste
  dominante de cualquier comando es la red contra la fuente, no la consulta local. Si algún día deja
  de serlo, se mide antes de reabrir nada.
- Se depende de un port menos usado que el original, con su propio ritmo de publicación de parches.
  Lo vigilan las actualizaciones semanales de dependencias y `make vuln` en cada propuesta de cambio
  y en el flujo nocturno.
- El binario es más grande: el motor entero viaja como código Go dentro del ejecutable.
- Las extensiones nativas de SQLite (por ejemplo, cargar un módulo compilado) quedan fuera de
  alcance. Nada de lo previsto en el proyecto las necesita.
