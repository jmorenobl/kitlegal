# Contrato: la grabación de H36 y la raíz nueva de la unión (FR-015, FR-016)

## 1. La petición (tarea `[datos]`)

Al final de `normas` de `internal/evals/testdata/grabaciones.json` (`"fuente": "boe.legislacion-consolidada"`), una
entrada con la forma de `EntradaDelManifiesto`:

```json
{"busqueda": "enjuiciamiento civil", "titulo_empieza_por": "Ley 1/2000,", "bloques": ["a130", "a133"], "para": "data/plazos.yaml de H36: arts. 130 (días y horas inhábiles) y 133 (cómputo de los plazos) de la LEC, el cómputo civil (FR-014 a FR-016), con los ids de bloque de su índice grabado"}
```

La búsqueda y el prefijo son los de la entrada de la LEC del manifiesto de H5, así que la resuelve la búsqueda que ya
está grabada en el conjunto de H5 (research V16); `LeerManifiesto` la admite (no repite norma dentro del manifiesto de
H9) y `TestIdentificadoresDeLasNormas` sigue resolviendo la LEC por la entrada de H5. Los ids `a130` y `a133` son los
del índice grabado de `BOE-A-2000-323` (`testdata/evals/boe.legislacion-consolidada/…BOE-A-2000-323_texto_indice.json`);
si no lo fueran, `boe articulo` saldría con 3 y `TestGrabarEvals` fallaría nombrando la entrada.

En la misma tarea, `TestGrabarEvals` (`internal/evals/grabacion_test.go`, `//go:build grabacion`) construye el cliente
que graba con `httpx.ConRaizDeGrabacion(os.Getenv("KITLEGAL_EVIDENCIAS"))` cuando la variable no está vacía, y con la
raíz de su conjunto si lo está, como hoy. Nada más de la grabación cambia: lee los dos manifiestos, siembra cada
consulta desde `UnionDeGrabaciones` con `Preparar` y solo pide a la fuente lo que falta (V7, V8 de research). FR-017
no nombra `grabacion_test.go`: es una desviación declarada, por FR-016 (research D10).

## 2. Lo que graba `grabar_datos`

Tras el commit de la tarea, `scripts/workflow/grabar-datos.sh grabar H36` ejecuta `TestGrabar*` de `internal/evals` con
`KITLEGAL_RECORD=1` y `KITLEGAL_EVIDENCIAS=$PWD/evidencias/h36`. Lo único que falta son `boe articulo BOE-A-2000-323
a130` y `a133`; con ellos, el `robots.txt` que el cliente pide antes. Quedan en
`evidencias/h36/boe.legislacion-consolidada/` y en `evidencias/h36/manifiesto.json`, cada uno con su huella, y en
`gates/grabaciones.md`. Es el manifiesto nuevo de H36 que pide FR-016 (research D9). Si el BOE no responde, causa
mayor (S3 de research).

## 3. La raíz en la unión, el civil y su comprobación (tarea siguiente)

Una sola tarea, porque cada pieza sin las otras deja `make ci` en rojo o el civil sin comprobar:

- `internal/evals/grabaciones.go`: `GrabacionesDeH36 = "../../evidencias/h36/" + boe.NombreDeLaFuente`, y
  `UnionDeGrabaciones()` devuelve H4, H5, H9 y H36, en ese orden.
- `internal/evals/grabaciones_test.go`: la lista esperada de `TestGrabacionesSinSolape` gana `GrabacionesDeH36`
  (desviación declarada, research D10); `verificarPreceptos` añade, detrás de las de los tipos, las lecturas de los
  preceptos de los cómputos que no nombra ningún tipo, con una estructura YAML propia del test (research D11); la
  tabla `palabrasClaveDeLosPreceptos` gana, en `a130`, «24 de diciembre», «6 de enero» y «agosto», y en `a133`, «día
  siguiente», «veinticuatro horas», «se excluirán los inhábiles», «de fecha a fecha» y «prorrogados» (FR-015).
  `lecturasDeLosTipos` y `loQueLlevaALeerElRepositorio` no cambian: el civil no entra en lo que la skill lleva a leer
  ni en el oráculo de `TestLecturasDeLosDatos` y de `TestEvalsDelRepositorio/lo-que-lleva-a-leer` (FR-014; research
  V23).
- `internal/skills/normas_test.go`: en `rangosDeLasBusquedasGrabadas`, que recorre la unión en
  `TestEsquemaDeNormas/rangos-grabados`, la exigencia de alguna búsqueda grabada pasa de cada conjunto a la unión,
  porque el conjunto de H36 no tiene ninguna (§2); lo comparado no cambia (desviación declarada, research D10).
- `data/plazos.yaml`: el cómputo `civil` de contracts/datos-de-plazos.md §1.

Con eso, `TestPreceptosDeLosPlazos` comprueba que los dos bloques están grabados, que llevan sus palabras y que cada
`vigente_desde` de los cinco preceptos del civil es el de sus redacciones grabadas (SC-006). En rojo (mutantes de US6.2):
el del 133.1 con el 2010-05-04, y el del 130.2 con el 2015-10-07.

## 4. Lo que no cambia

`internal/evals/consultas.go`, `TestLecturasDeLosDatos`, el subtest `lo-que-lleva-a-leer` de `TestEvalsDelRepositorio`,
`TestManifiestoDeGrabaciones` (los dos manifiestos reales, con su mínimo de entradas), `TestDirectorioDeLosGrafosPrevios`
(la raíz de H36 está fuera de `testdata/evals/grafo-previo`), `TestIdentificadoresDeLasNormas`, `docs/SOURCES.md` y
ningún fichero de `data/festivos/`. `GrafosPrevios` sigue fuera de la unión.
