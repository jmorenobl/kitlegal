# Seguridad

## Cómo avisar de una vulnerabilidad

No abras una incidencia pública. Usa el aviso privado de GitHub: en la pestaña **Security** del repositorio,
**Report a vulnerability**. Solo lo ve quien mantiene kitlegal, y la conversación sigue ahí hasta que haya
arreglo y versión publicada; entonces se publica el aviso con el crédito que quieras.

Incluye lo que haga falta para reproducirlo: la versión (`kitlegal version`), el sistema, la orden exacta y qué
esperabas frente a qué pasó.

## Qué cubre

- El programa `kitlegal` y todo lo que hace en tu equipo: lo que lee y escribe (la caché, las skills que instala
  `kitlegal skills install`, su manifiesto) y lo que pide por la red.
- El instalador `scripts/install.sh` y la cadena de publicación: los archivos de cada release, sus checksums, la
  firma, el SBOM, la atestación de procedencia, el cask de Homebrew y el bucket de Scoop.
- Las skills, en lo que puedan hacer ejecutar a un agente.

Un error en el contenido legal —una cita mal resuelta, un aviso de vigencia que falta— no es una vulnerabilidad:
abre una incidencia normal.

## Versiones con soporte

Solo la última versión publicada. Los arreglos de seguridad salen en una versión nueva, no en parches de las
anteriores.
