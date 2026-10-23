# Contrato: la release exige la suite completa del commit etiquetado

`.github/workflows/release.yml` (FR-025; research D14; data-model §11).

## 1. El trabajo `suite-completa`

```yaml
  suite-completa:
    runs-on: ubuntu-24.04
    permissions:
      contents: read
      actions: read
    env:
      GH_TOKEN: ${{ github.token }}
      GH_REPO: ${{ github.repository }}
      COMMIT_ETIQUETADO: ${{ github.sha }}
      ETIQUETA: ${{ github.ref_name }}
    steps:
      - uses: actions/checkout@v7          # el commit etiquetado: de él salen las evals de cada skill
        with:
          persist-credentials: false       # como publicar: git no necesita el token
      - uses: actions/setup-go@v7
        with:
          go-version-file: go.mod
      - name: Comprobar la suite completa de evals del commit etiquetado
        run: >-
          go test -tags evals -count=1 -timeout 15m -v -run '^TestSuiteCompletaDelCommit$' ./internal/evals/
          -args -commit "$COMMIT_ETIQUETADO" -etiqueta "$ETIQUETA"
```

`publicar` lleva `needs: [suite-completa]`; `humo` y `catalogo` siguen detrás de `publicar`. Si el trabajo falla, no se
publica la release, ni el cask, ni el manifiesto del bucket, ni el catálogo. Las claves son las que admite
`trabajoDelFlujo` de `release_test.go`.

## 2. La comprobación

Para el commit *c* (data-model §11): de las ejecuciones del flujo `evals` sobre *c* (`gh run list --commit`), terminadas
y de la más reciente a la más antigua, vale la primera con `headSha` *c*, todos los trabajos `evals (<skill>)` de las
skills con evals en *c* en `success`, y en el informe de cada una `origen.modo: completa` con cada eval de
`evals/<skill>/` en *c* medida y ninguna reutilizada. Sale bien con la primera que vale. Si ninguna vale, falla con este
mensaje (una línea por ejecución mirada, con el primer motivo):

```text
release: el commit 0f3c2a9e8d7b6c5a4f3e2d1c0b9a8f7e6d5c4b3a no tiene una ejecución del job de evals de la suite completa, en verde y sin evals reutilizadas.
- ejecución 18734029311: el trabajo evals (legal-core) terminó en failure
- ejecución 18733990002: el informe de boe-legislacion es de una ejecución con reutilización (3 evals reutilizadas)
Lanza la suite completa sobre la etiqueta y, cuando termine en verde, vuelve a ejecutar la release:
  gh workflow run evals.yml --ref v0.9.0
```

Sin ninguna ejecución sobre *c*, la lista dice `- ninguna ejecución del flujo evals sobre el commit`. Unas 3 a 12 líneas,
que lee quien empuja la etiqueta, una vez por etiqueta; deja de darse cuando el commit tiene su ejecución completa en
verde.

## 3. Controles de `make ci`

- `release_test.go`, subtest `suite-completa` de `TestConfiguracionDeLaRelease`: el trabajo existe con su paso y sus
  permisos, el paso pasa `${{ github.sha }}` como commit, y `publicar` depende de él. Mutantes: sin el trabajo, sin la
  dependencia, con otro commit en el paso.
- `internal/evals/suite_completa_test.go`, `TestComprobarLaSuiteCompleta`, con la plataforma en memoria: falla sin
  ejecución, con otro `headSha`, con un trabajo en `failure`, con un informe que falta, con una eval reutilizada, con
  modo `reutilizacion`, con una eval del commit que no está en el informe; sale bien con una completa y en verde.
