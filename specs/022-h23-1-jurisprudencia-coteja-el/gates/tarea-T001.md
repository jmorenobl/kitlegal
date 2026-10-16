# T001 · nota de la tarea (intento 1)

La tarea quedó en verde en este intento. Esta nota recoge el rojo que pide la tarea antes del código, y la verificación.

## Rojo con el código de producción de antes

Con los casos nuevos de `TestUmbralesDeJurisprudencia` y su arnés ya escritos, y sin tocar todavía `informe.go`,
`umbrales.go` ni `doc.go` (`piden-el-cotejo` quedó fuera un momento, porque llama a `pideElCotejo`, que no compila sin
el código):

```text
go test -count=1 -run '^TestUmbralesDeJurisprudencia$' ./internal/evals/
```

- `FAIL`, compilando, por aserciones: los catorce casos de la tabla fallan y `el-elemento-del-contrato` pasa.
- En los catorce, `require.Len` da `should have 15 item(s), but has 13` («los quince de
  contracts/umbral-documento-sin-cotejar.md §1 de H23.1»), y `exigirUmbrales` da la lista sin ningún
  `documento_sin_cotejar`.
- En los casos nuevos, `exigirMotivosDeLaRaiz` espera el motivo `umbral documento_sin_cotejar:…` y no lo hay: el informe
  no da el motivo que pone el veredicto en `fallo`. La aserción del veredicto (`fallo` frente a `aprobado`) no llega a
  ejecutarse, porque `require.Len` corta el caso antes.

## Rojo de `piden-el-cotejo`

Con el código escrito, quitando un momento la 09 de la lista esperada:

```text
go test -count=1 -run '^TestEvalsDelRepositorio$/^piden-el-cotejo$' ./internal/evals/
--- FAIL: TestEvalsDelRepositorio/piden-el-cotejo
    Error: Not equal:
      - "08-doctrina-con-el-fallo-delante.yaml"
      + "08-doctrina-con-el-fallo-delante.yaml",
      + "09-de-que-trata-con-la-ficha-sola.yaml"
    Messages: las evals de ../../evals/jurisprudencia que piden el cotejo
```

Con la 09 de vuelta en la lista, pasa.

## Verificación

- quickstart.md §1, con `-v`: `ok`, quince subtests en `PASS`.
- quickstart.md §2, primera orden: `ok`, y `piden-el-cotejo` en `PASS`.
- quickstart.md §3: la primera orden da `ok`; la segunda, sin salida; la tercera, solo los cinco ficheros de la tarea
  (`skills/jurisprudencia/SKILL.md` es de T003).
- `golangci-lint run ./internal/evals/`: `0 issues.`; `golangci-lint fmt --diff ./internal/evals/...`: sin diferencias.
- `grep -n -i 'trece\|nueve'` en los ficheros Go del paquete: lo que queda de «trece» y «nueve» es de
  `boe-legislacion`, del cierre de H26 (`como-en-el-cierre-de-h26` y `umbralesDeJurisprudenciaComoEnH26`) o de otra
  cosa (cifras, sondeo). Se corrigieron dos frases de `informe_test.go`: «la lista de los trece» del caso
  `tres-con-si-en-afirma-que-existe`, que pasa a quince, y la cabecera de `como-en-el-cierre-de-h26`, que ahora dice
  que esos trece son los de entonces.
- `make ci`, en primer plano: código de salida 0 (`make ci > /dev/null && echo CI-VERDE` dio `CI-VERDE`), con
  `internal/evals` en `ok` y gitleaks sin fugas.

Además de lo que pide la tarea, `el-elemento-del-contrato` compara el elemento de contracts §2, tal cual y con sus
claves en orden, con `umbralDocumentoSinCotejar`, el ayudante escrito a mano con el que los casos comparan el informe.
