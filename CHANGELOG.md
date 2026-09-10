# Changelog

Todo cambio de comportamiento visible de `kitlegal` se registra aquí.

El formato sigue [Keep a Changelog](https://keepachangelog.com/es-ES/1.1.0/) y el proyecto se adhiere al
[versionado semántico](https://semver.org/lang/es/). Mientras el mayor sea `0` —lo será hasta la primera
release, que es H6 (`v0.1.0`)— un cambio incompatible sube el **menor**. Este fichero se mantiene **a
mano** hasta ese hito: cada propuesta de cambio añade su entrada bajo *Unreleased* en el mismo cambio que
introduce el comportamiento, y al publicar una versión esa sección se cierra bajo su número y su fecha y
se abre una nueva vacía.

## [Unreleased]

Todavía no hay ninguna versión publicada. Lo que sigue es lo que aporta **H0 — esqueleto del repositorio
y sus controles**: un repositorio sin fuentes legales todavía, pero blindado, para que cualquier línea de
Go que entre después atraviese los mismos gates.

### Añadido

- **Binario `kitlegal`** con un único verbo, `version`, que imprime en tres líneas de la salida estándar
  la versión, el commit y la fecha de construcción, deja vacía la salida de error y termina con código
  `0`. La versión sale de `git describe --tags --always --dirty` y el commit coincide carácter a carácter
  con la revisión construida. Cualquier otra invocación —sin verbo, con un verbo desconocido o con un
  argumento sobrante— escribe una línea de uso en la salida de error y termina con código `2` («args» en
  la tabla de códigos de salida estables del proyecto).
- **`Makefile` como única superficie de invocación de los controles**, con `make ci` como veredicto del
  repositorio: la misma orden que se ejecuta en local es la que ejecutan el gancho de pre-commit y la
  integración continua, y ningún control se aplica por otra vía. `make ci` no modifica ningún fichero
  versionado. `make build` y `make install` compilan sin cgo, con `-trimpath` e inyectando los datos de
  construcción; `make help` es el objetivo por defecto.
- **Ocho controles activos dentro de `make ci`**: formato en modo verificación (`gofumpt` y `goimports`),
  análisis estático (`golangci-lint`, con `gosec` y `govet` incluidos), tests unitarios con detector de
  carreras y perfil de cobertura, vulnerabilidades conocidas (`govulncheck`), validación contra esquemas,
  detección de secretos (`gitleaks`), integridad de los módulos (`go mod verify`, raíz y herramientas) y
  dependencias saneadas (`go mod tidy -diff`). `make check-tools` comprueba los prerrequisitos como
  dependencia de las demás órdenes.
- **Cadena de herramientas reproducible y sin instalación previa**: los cuatro controles con binario
  propio —`golangci-lint`, `govulncheck`, `gitleaks` y `lefthook`— se construyen solos desde la versión
  fijada en `tools/<herramienta>/go.mod`, y `go.mod` declara la directiva `toolchain` que el `Makefile`
  exporta como `GOTOOLCHAIN`, de modo que todas las órdenes usan el mismo parche de Go que la integración
  continua. Los dos únicos prerrequisitos son una cadena Go 1.21 o superior y `git`.
- **Ganchos de pre-commit** (`make hooks`, con `lefthook`): cada commit corrige el formato y vuelve a
  preparar lo corregido, y ejecuta `lint-fast`, `secrets` y `mod-tidy-check`. Es un subconjunto rápido, no
  el veredicto: la autoridad final sigue siendo la integración continua.
- **Umbrales de cobertura bloqueantes** (`codecov.yml`): ≥ 70 % global y ≥ 85 % en el dominio interno,
  ambos declarados como estado que falla y no como información.
- **Flujos de la plataforma**: `ci` en cada propuesta de cambio y en cada push a la rama principal,
  `nightly` a diario sobre la rama principal —ambos se limitan a preparar el entorno y ejecutar `make
  ci`—, `codeql` semanal como segundo análisis de seguridad y Dependabot semanal sobre el módulo raíz, los
  cuatro módulos de herramienta y las acciones de los flujos.
- **Documentación fundacional**: `LICENSE` (Apache-2.0), `README.md`, `CONTRIBUTING.md` —ritual por hito,
  estructura de propuesta de cambio, Conventional Commits, versionado semántico, catálogo de controles y
  justificación obligatoria de toda dependencia nueva—, este `CHANGELOG.md` y los cuatro ADR fundacionales
  en `docs/ADR/`: multicall, SQLite sin cgo, CENDOJ no masivo y frontera humana.

Cinco órdenes existen ya pero reciben su contenido en un hito posterior y ninguna miente sobre ello:
`test-integration`, `test-e2e` (H1), `schema-check` (H4 y H11), `skills-sync` (H5) y `release`, que falla
con código distinto de `0` hasta H6 por ser la única con efectos externos. Los applets de fuentes (`boe`,
`placsp`, `bdns`…) llegan en los hitos siguientes, en el orden de `docs/ROADMAP.md`.
