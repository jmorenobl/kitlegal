# Contrato: la eval de la consulta repetida, su derivada y lo que se retira

FR-001 a FR-004, FR-010 a FR-014, FR-020, FR-021, FR-030, FR-081; SC-005, SC-007. Decisiones en research D12-D15.

## 1. La eval

`evals/boe-legislacion/19-lcsp-contrato-menor-redaccion-cambiada.yaml`, con este contenido:

```yaml
# Eval informativa (ADR 0016): la consulta repetida sobre un artículo que el BOE modificó de verdad. El grafo de la
# sesión ya registró una lectura del art. 118 de la LCSP que vio su redacción original (vigencia 20180309), derivada de
# la respuesta grabada quitándole solo la redacción posterior; la caché sirve la grabada, con la vigente (20200206, Real
# Decreto-ley 3/2020). La sesión tiene que leer el bloque con kitlegal boe articulo, comprobar con kitlegal graph check y
# la norma, no pedir graph show, citar el bloque y trasladar el cambio de redacción con la forma fija
# ⚠ REDACCIÓN MODIFICADA:.
pregunta: "Hace tiempo te pregunté qué exige el artículo 118 de la LCSP para el expediente de un contrato menor. ¿Qué dice ahora?"
activa: true
informativa: true
grafo_previo:
  grabaciones: lcsp-a1-30-redaccion-original
  comandos:
    - applet: boe
      norma: BOE-A-2017-12902
      bloque: a1-30
comandos:
  - applet: boe
    norma: BOE-A-2017-12902
    bloque: a1-30
  - applet: graph
    verbo: check
    norma: BOE-A-2017-12902
prohibidos:
  - applet: graph
    verbo: show
citas:
  - norma: BOE-A-2017-12902
    bloque: a1-30
hallazgos:
  - version-obsoleta
```

Valida contra `schemas/eval.yaml.json` sin cambiarlo. El conjunto queda en 19 ficheros de eval con exactamente 10
positivas que deciden (01-10), informativas 13-19 y las reglas sin cambios (FR-030). El informe declara junto a su tasa
la forma `⚠ REDACCIÓN MODIFICADA:` (`formas`, desde H7.1; FR-004). Se juzga con la lista (contracts/lista-y-juicio.md
§4): una respuesta con la forma y sin expresiones pasa si hace lo demás (FR-003).

## 2. La derivada

`testdata/evals/grafo-previo/lcsp-a1-30-redaccion-original/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2017-12902_texto_bloque_a1-30.json`
(tarea `[datos]`).

**Derivación** (función de `internal/app/grafo_test.go`): de la grabación H4 del mismo nombre, en
`internal/source/boe/testdata/boe.legislacion-consolidada/`, y la fecha 20180309:

1. lee la grabación en un tipo local con los campos del formato de `httpx` en su orden (`formato`, `grabado_en`,
   `peticion{metodo, url, cabeceras}`, `respuesta{estado, cabeceras, cuerpo}`);
2. en `respuesta.cuerpo`, localiza con `encoding/xml` las `<version>` hijas del `<bloque>` y quita los bytes desde el
   final de `</version>` de la de `fecha_vigencia="20180309"` hasta el final del `</version>` de la última (las
   redacciones posteriores y el blanco que las precede);
3. escribe con el codificador de las grabaciones: sangrado de dos espacios, sin escapar HTML, salto de línea final.

Premisa del test: derivar con la fecha de la última redacción (no quita nada) devuelve la grabación byte a byte; sin
ella, la derivación podría cambiar el formato y la comparación de §3.1 no diría nada de la grabación.

**Escritura**: `go test -count=1 -run '^TestGrabacionesDerivadas$' ./internal/app/ -args -actualizar-derivadas`
escribe cada derivada del grafo previo desde su grabación antes de comprobar. Sin la bandera, solo comprueba.

## 3. El control de derivaciones

La comprobación de cada derivada la decide su carpeta, no la clase de la entrada que la declara. Las declaraciones se
separan por carpeta:

- **las del e2e** (`derivadasDelE2E`): `grabacionesDerivadas()`, con sus entradas de hoy, su clase y su comprobación, sin
  cambios (FR-013);
- **las del grafo previo** (`grafosPreviosDeLasEvals`): una lista de un tipo propio, la **derivada del grafo previo**,
  con su subcarpeta, su fichero, los argumentos de `boe` (`articulo BOE-A-2017-12902 a1-30`) y la fecha de vigencia de
  la redacción que da (20180309). El tipo no lleva comprobación propia —las comprobaciones 1 y 2 las aplica el test a
  toda entrada de la lista— y ninguna entrada de otra clase cabe en ella.

`TestGrabacionesDerivadas` compara carpeta a carpeta: los ficheros de `derivadasDelE2E` con las entradas del e2e, y los
de `grafosPreviosDeLasEvals` con las del grafo previo, en los dos sentidos. Una entrada de otra clase con carpeta de grafo
previo —`versionDelArticulo21`, que se queda, admite cualquier carpeta— hace fallar el test nombrándola: es una entrada
del e2e sin fichero en su carpeta, y su fichero, uno del grafo previo sin entrada. Toda derivada bajo
`grafosPreviosDeLasEvals` pasa, por tanto, las comprobaciones 1 y 2:

1. **Reproducible y sin otro cambio** (FR-010): el fichero es, byte a byte, la derivación de §2 sobre su grabación.
2. **Una redacción de la grabada** (FR-011): servida en lugar de la grabación, `boe articulo … --json` da un `Articulo`
   igual, campo a campo, a exactamente uno de los que da `boe` sobre la grabación reducida a cada una de sus
   redacciones (§2 con la fecha de cada `<version>`), y ese es el de la fecha declarada. Con la derivada de H7.2:
   `fecha_vigencia` 20180309, `fecha_version` 20171109, `norma_modificadora` `BOE-A-2017-12902`, el texto original y su
   `hash_texto`, tal como los trae la grabada (SC-005).
3. **Lo inventado no pasa** (FR-012): un test arma en `t.TempDir()`, desde la grabación, tres derivadas —la fecha
   20151002 con el texto original; un párrafo de más al final de la redacción original; y las dos cosas, como la de la
   eval 19 retirada— y exige que la comprobación 2 falle con cada una nombrando su fichero y diciendo que la redacción
   que da no es ninguna de las de la grabada. La huella la calcula `boe` desde el texto: no existe una derivada con una
   huella distinta y el mismo texto.

Las derivadas del e2e (`version-posterior`, `version-ulterior`, `sin-eli`, `eli-sin-segmento`) siguen con su clase y su
comprobación, y todo fichero de las dos carpetas sigue teniendo su comprobación, y toda comprobación su fichero (FR-013):
la comparación carpeta a carpeta lo sigue exigiendo, cada fichero con la declaración de su carpeta.

**Cuándo entra la comparación carpeta a carpeta**: con la retirada de `lpac-a21-version-anterior` (§5), que es hoy una
entrada del e2e (`versionDelArticulo21`) con carpeta de grafo previo y la haría fallar. Hasta entonces, la entrada nueva
se comprueba con 1 y 2 y la comparación es la de hoy, las dos carpetas juntas; así ninguna tarea queda en rojo.

## 4. La preparación de la eval dice lo que la sesión verá (FR-002)

La subprueba `grafo-previo` de `TestEvalsDelRepositorio` (`compruebaElGrafoPrevio`) sigue exigiendo lo de H7 —el
directorio existe, la preparación no da faltas y deja un `BloqueVersion` por comando— y además, para cada eval con
`grafo_previo`:

1. anota la `fecha_vigencia` del `BloqueVersion` que dejó el grafo previo;
2. lee en proceso cada bloque de sus comandos con `boe articulo <norma> <bloque> --json` sobre `UnionDeGrabaciones()`,
   con una caché temporal y entregando al grafo de la sesión, como la sesión;
3. ejecuta `graph check <norma> <bloques> --json` por cada norma de sus comandos;
4. exige código 0 en las dos, que las clases de `data.hallazgos` sean exactamente las de `hallazgos` de la eval y que
   cada `version-obsoleta` lleve en `fecha_vigencia` la del punto 1 y en `fecha_vigencia_reciente` la que dio el punto 2.

Con la eval 19: un `version-obsoleta`, 20180309 → 20200206 (FR-002, US2.2). Sin red: todo sale de las grabaciones.

## 5. Lo que se retira y lo que se adapta

| Se retira | Cómo se comprueba |
|---|---|
| `evals/boe-legislacion/19-lpac-articulo-21-redaccion-cambiada.yaml` | no existe |
| `testdata/evals/grafo-previo/lpac-a21-version-anterior/` (tarea `[datos]`) | no existe; `TestGrabacionesDerivadas` exige que toda carpeta tenga su comprobación |
| Su entrada en `grabacionesDerivadas()` y `parrafoDeLaVersionAnterior` | `git grep` sin coincidencias en código ni tests (quickstart §4) |

`versionDelArticulo21` se queda (la usan `version-posterior` y `version-ulterior`).

Tests que se adaptan (FR-021), sin cambiar lo que comprueban ni desactivar nada:

- `internal/evals/formato_test.go`: los documentos sintéticos con `grabaciones: lpac-a21-version-anterior` pasan a
  `grabaciones: lcsp-a1-30-redaccion-original`.
- `internal/evals/consultas_test.go`: el grafo previo sintético, igual.
- `internal/evals/juzgar_test.go`: `ficheroDeLaConsultaRepetida` pasa a `19-lcsp-contrato-menor-redaccion-cambiada.yaml`
  y la eval sintética de la consulta repetida (`evalDeLaConsultaRepetida` y los juicios que la usan), a la forma de la
  nueva: norma `BOE-A-2017-12902`, bloque `a1-30`, la comprobación con la norma.
- `internal/app/grafo_test.go`: la entrada retirada sale de `grabacionesDerivadas()` y su comentario deja de nombrarla;
  con ella, `TestGrabacionesDerivadas` pasa a comparar carpeta a carpeta (§3).
