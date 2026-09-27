# Revisión final de H19: lo que solo se arregla en la configuración de CI

El corrector no toca `.github/`: el workflow deshace cualquier cambio ahí. Lo que sigue queda para la persona que
revisa la propuesta de cambio; ninguno cambia el comportamiento del producto ni de `make ci`.

- .github/workflows/ci.yml:106-107 (trabajo `snapshot`): el comentario del paso «Construir el snapshot» dice «Los seis archivos y checksums.txt en dist/»; `make release` deja además los cuatro paquetes `.deb` y `.rpm`, y `checksums.txt` lista también `install.sh` (README.md y CONTRIBUTING.md ya lo dicen así).
- .github/workflows/ci.yml:46-50 (trabajo `ci`): el comentario de la caché de `setup-go` habla de «las cuatro herramientas de control»; desde H19 son cinco, con goreleaser (`tools/goreleaser/`).
- .github/workflows/nightly.yml:33-36: dice que el flujo no hace «ninguna invocación de `release` —que no entra ni en `ci` ni en `nightly` y falla a propósito hasta H6—»; desde H19 `make release` construye el snapshot y lo ejecuta el trabajo `snapshot` de `ci`.
- .github/workflows/evals.yml:55-56: el filtro de cambios que decide si se lanzan las evals no incluye `skills.go` (lo que el binario empotra), `internal/app/` (la composición del binario que las skills invocan desde el `PATH`) ni `internal/disco/`; una propuesta que solo toque esas rutas no lanza las evals aunque cambie lo que las skills ejecutan.
- .github/dependabot.yml:74-76: «Las acciones que usan los tres flujos… cubre `ci`, `nightly` y `codeql`»; tras `evals.yml` y `release.yml` son cinco (también anotado en gates/supuestos.md con los pendientes de docs/ROADMAP.md).
- .github/workflows/release.yml:122 (trabajo `humo`): `echo "version imprime «$primera» en lugar de…"` escribe una variable sin llaves delante de «»», la misma forma que hacía abortar `install.sh` con el `/bin/sh` de macOS en un locale UTF-8. Ahí no falla —el trabajo corre en `ubuntu-latest` con bash 5—, pero escribirla `«${primera}»` la dejaría a salvo si el paso se moviera a macOS.
