# 0022 · Licencia: EUPL-1.2

- **Estado**: aceptada
- **Fecha**: 2026-09-27
- **Hito**: transversal (tras publicar v0.1.1). Sustituye la Apache-2.0 con la que nació el repositorio en H0.

## Contexto y problema

kitlegal nació con Apache-2.0, la opción que `refs/kitlegal-estructura-y-ecosistema.md` §6 recomendaba por adopción.
Es una herramienta pública sobre fuentes públicas españolas y europeas, pensada también para administraciones, y una
licencia permisiva deja que cualquiera la cierre en un producto propio sin devolver las mejoras.

## Opciones consideradas

1. **Mantener Apache-2.0.** Máxima adopción; nada obliga a publicar las versiones modificadas.
2. **GPL-3.0 o AGPL-3.0.** Copyleft fuerte, pero redactadas para el derecho estadounidense y con la versión inglesa
   como única oficial.
3. **EUPL-1.2.** La licencia de software libre que publica la Comisión Europea: copyleft que alcanza también la
   comunicación por red, redactada para el derecho europeo, con el mismo valor en todas sus versiones lingüísticas
   oficiales (también en castellano), y con un apéndice de licencias compatibles (GPL, AGPL, MPL, LGPL, EPL…) que
   evita el choque entre copylefts. La usan muchas administraciones europeas para su propio software.

## Decisión

Opción 3: EUPL-1.2. `LICENSE` lleva el texto oficial en inglés (el de la lista de licencias SPDX, que es el que
reconocen GitHub y pkg.go.dev); la versión en castellano tiene el mismo valor. El identificador SPDX `EUPL-1.2` va en
el cask, el bucket y los paquetes de la release.

## Consecuencias

- Quien distribuya kitlegal modificado, o lo ofrezca como servicio, tiene que publicar su código bajo la EUPL (o una
  licencia compatible de su apéndice). Usarlo tal cual, o como herramienta de un agente, no obliga a nada.
- Las dependencias (MIT, BSD, Apache-2.0) son compatibles: se pueden combinar en una obra bajo EUPL.
- Todo el código tiene un único autor; las otras contribuciones son actualizaciones automáticas de versiones de
  Dependabot, que no aportan obra propia. El cambio no necesita el consentimiento de nadie más. Las versiones ya publicadas (v0.1.1) siguen bajo Apache-2.0 para quien las tenga.
