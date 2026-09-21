# T007 — tarea mal delimitada: el subtest `gramaticas` exige cambiar tres esquemas y T007 no puede tocar `schemas/` (intento 1)

**Estado**: sin marcar (`[ ]`), **redelimitada**. En el árbol no queda ningún fichero de código de T007: solo cambian
`tasks.md` y esta nota, ambos en el directorio del hito. Si quedara algo, el guardián de la tarea siguiente (T027, que
solo declara los tres esquemas) lo rechazaría por estar fuera de sus rutas. T007 entera está implementada y verificada
en un clon desechable (sección «Verificación del arreglo»), así que el intento 2 puede partir de ahí.

## Causa

El subtest `TestTerritorioDelRepositorio/gramaticas` exige que «los `pattern` de los tres esquemas acepten y rechacen
exactamente lo mismo que `AnalizarCodigoINE` y `AnalizarDIR3`», igual que el contrato de identificadores §2 y
research.md V38. Los patrones que fijaron T001 y T003 (y T002 en la comunidad) no son esas gramáticas:

| Sitio | Patrón actual | Gramática de `ids` | Discrepancia |
|---|---|---|---|
| clave de `municipios` (municipios) | `^[0-9]{5}$` | provincia `01`-`52`, municipio `001`-`999` | el esquema acepta `00074`, `53001`, `99999` y `28000` |
| `provincia` de cada fila (municipios) | `^[0-9]{2}$` | `01`-`52` | el esquema acepta `00`, `53` y `99` |
| clave de `correspondencia` (dir3) | `^[0-9]{5}$` | como la clave de municipios | lo mismo que ella |
| valor de `correspondencia` (dir3) | `^L01[0-9]{6}$` | `[Ll]01` + código INE en rango + dígito | el esquema rechaza `l01280748` y acepta `L01000740`, `L01530011` y `L01280000` |
| clave de `provincias` (comunidad) | `^[0-9]{2}$` | `01`-`52` | el esquema acepta `00`, `53` y `99` |

El `dc` de cada fila (`^[0-9]$`) sí coincide con el dígito de `AnalizarCodigoINEConDigito`. Los «tres esquemas» del
subtest son, por tanto, los de municipios, DIR3 y comunidad: los tres que llevan un código INE, una provincia o un DIR3.

Con T007 entera en un clon y los esquemas **tal como están** en la rama, el subtest queda en rojo exactamente por esas
filas (salida de `go test -run 'TestTerritorioDelRepositorio/gramaticas'`, resumida):

```text
--- FAIL: TestTerritorioDelRepositorio/gramaticas/provincia_de_una_comunidad
      el esquema rechaza "00" / "53" / "99" en provincia de una comunidad, como internal/core/ids   (An error is expected but got nil)
--- FAIL: TestTerritorioDelRepositorio/gramaticas/provincia_de_un_municipio
      el esquema rechaza "00" / "53" / "99" en provincia de un municipio, como internal/core/ids
--- FAIL: TestTerritorioDelRepositorio/gramaticas/clave_de_municipios
      el esquema rechaza "00074" / "53001" / "99999" / "28000" en clave de municipios, como internal/core/ids
--- FAIL: TestTerritorioDelRepositorio/gramaticas/clave_de_la_correspondencia
      el esquema rechaza "00074" / "53001" / "99999" / "28000" en clave de la correspondencia, como internal/core/ids
--- FAIL: TestTerritorioDelRepositorio/gramaticas/DIR3_de_la_correspondencia
      el esquema acepta "l01280748" en DIR3 de la correspondencia, como internal/core/ids   (Received unexpected error)
      el esquema rechaza "L01000740" / "L01530011" / "L01280000" en DIR3 de la correspondencia, como internal/core/ids
```

Cumplirlo exige cambiar esos cinco patrones, y T007 no lleva `[datos]`: el guardián rechaza cualquier cambio bajo
`schemas/`. Relajar el subtest —elegir casos en los que coincidan, o comparar solo la forma y no los rangos— sería un
atajo que la batería de `tasks.md` prohíbe y vaciaría el control de V38.

## Decisión

- **Tarea nueva T027 `[datos]`, colocada en `tasks.md` justo antes de T007** (el bucle toma la primera línea `- [ ]` en
  el orden del fichero, no por número, como la T038 de H4). Solo cambia los cinco patrones de los tres esquemas y
  alinea data-model §3.1 y §3.2, que los describen. No toca código ni datos y deja `make ci` en verde por sí sola:
  hasta T007 ningún lector mira esos esquemas. Al modificar esquemas existentes, **provoca pausa humana**, que revisa
  los tres juntos. Es tarea aparte y no una excepción dentro de T007 porque es divisible —en verde antes que T007— y
  `tasks.md` solo admite código en una tarea `[datos]` cuando un control existente lo hace inseparable (D16, D28).
- **Igualdad literal, `[Ll]` incluido.** El DIR3 acepta la letra en mayúscula y en minúscula, como `AnalizarDIR3` y
  como la gramática del contrato de identificadores §2 (`^[Ll]01[0-9]{6}$`, normalizado a mayúscula); el fichero
  congelado lo escribe en mayúscula y el territorio resuelto lo emite siempre en mayúscula. La alternativa —que el
  esquema acepte solo la forma canónica y el subtest compare «lo que el analizador acepta y escribe así»— es defendible,
  pero añade una regla al control que el contrato no enuncia; la decide la persona en la pausa de T027 si la prefiere.
- **T007 declara además `internal/skills/export_test.go`** (y gana `TestCompilarEsquemaDelTerritorioDesdeUnaRuta`):
  el compilador de los cuatro esquemas lee cada uno por una ruta constante que nunca falla, y la forma que el paquete
  ya usa para fijar esos errores es exponerlo en `export_test.go`, como `CompilarEsquemaDeNormas`. Sin eso, las dos
  ramas de error quedarían sin test.
- `tasks.md`: las cuentas de tareas `[datos]` (once) y de pausas (nueve), las excepciones de rebanadas verticales, la
  trazabilidad (FR-030 a FR-034, FR-044, FR-086), la dependencia **T004 → T027 → T007**, los ficheros existentes que se
  tocan, la historia US7 y el criterio e de la rúbrica.

## Verificación del arreglo

Sobre un clon desechable de la base de este intento (`2fc10f8`, `git clone` a `/tmp/kitlegal-t007`, fuera del
repositorio y sin versionar), con los patrones de T027 y T007 completa:

- Patrones aplicados: código INE `^(0[1-9]|[1-4][0-9]|5[0-2])(00[1-9]|0[1-9][0-9]|[1-9][0-9]{2})$`; provincia
  `^(0[1-9]|[1-4][0-9]|5[0-2])$`; DIR3 `^[Ll]01(0[1-9]|[1-4][0-9]|5[0-2])(00[1-9]|0[1-9][0-9]|[1-9][0-9]{2})[0-9]$`.
  Los ficheros congelados reales validan con ellos (subtest `esquema` en verde): todos sus códigos están en rango.
- `make -C /tmp/kitlegal-t007 ci` → **código 0**, con la línea `ci: todos los controles en verde` y
  `TestTerritorioDelRepositorio` en la orden de `skills-check`. Cobertura de `internal/skills`, 98,5 %; `data`, 0 %,
  como prevé plan.md (sin test propio: lo ejercen `TestRegistroDeProduccion` y el e2e desde T010).
- Los nueve subtests en verde sobre el corpus real; el paquete tarda ≈ 7 s con `-race` (leer la relación contra su
  esquema cuesta ≈ 5,7 s por la búsqueda de claves repetidas del lector común, que es cuadrática) y ≈ 1,3 s sin él.
- Sondas negativas, cada una en rojo nombrando lo que falla: la `comunidad` de la primera fila cambiada a `99`
  (`integridad`, con el municipio y las dos comunidades), una `Ø` delante de un nombre (`pliegue-cubre-el-corpus`),
  `boletines` añadidos a la comunidad `01` (`solo-madrid-configurada`: `[]string{"01", "13"}`) y el fichero de la
  comunidad `19` retirado (el test entero, antes de ningún subtest: 18 de 19).

## Para el intento 2

El prototipo verificado está en `/tmp/kitlegal-t007` (puede no existir ya; es orientativo, no normativo):
`data/datos.go`, `internal/skills/territorio.go`, `internal/skills/territorio_test.go`,
`internal/skills/export_test.go` y el `-run` de `skills-check` en el `Makefile`. Lo que costó ponerlo en verde:

- `gocyclo` (> 13) cuenta los cierres de `t.Run` dentro de `TestTerritorioDelRepositorio`: una función por subtest,
  recorridas desde una tabla, y un valor `corpusDelRepositorio` con tres `sync.OnceValues` (leer, cargar y resolver
  todos los municipios) que comparten los subtests.
- `misspell` toma `calcular` e `informacion` por erratas: no usarlos sueltos.
- `testifylint` (`require-error`) no admite `assert.NoError`/`assert.ErrorAs` en el bucle de casos de `gramaticas`: un
  subtest por valor, con `require`.
- La línea de un defecto de `propertyNames` en `municipios` es la del nodo del mapa (la de su primera fila), no la de
  la clave `municipios:`.
- `raizDelRepositorio` ya existe en `instalacion_test.go` bajo la etiqueta `integration`: con otro nombre, o
  `make test-integration` no compila.
- Los ficheros reales se leen por ruta relativa (`../../data/territorio/…`), sin importar el paquete `data`, y la
  exigencia de los cuatro ficheros y las 19 comunidades va en la preparación común, antes de ningún subtest.
