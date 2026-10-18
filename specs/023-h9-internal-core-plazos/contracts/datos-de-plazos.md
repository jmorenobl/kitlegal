# Contrato de los datos de `plazos`

`data/festivos/` lo escribe el código desde lo grabado (ADR 0039; FR-020 a FR-026); `data/plazos.yaml` y
`data/cobertura.yaml` los fija este plan desde el spec (FR-008, FR-031) y un test los comprueba contra lo grabado del
BOE (FR-030, FR-031), como `data/normas.yaml`. Decisiones: research D12 a D17.

## 1. `data/festivos/` (FR-020 a FR-026)

Dos ficheros, cada uno con una primera línea de comentario de fichero generado
(`# generado desde evidencias/h9/<fuente>/ por <Test> (<paquete>), no editar`), sus claves en este orden y sus listas
ordenadas como se dice. Los escribe `plazos.EscribirFestivos…` en su forma canónica y los lee
`plazos.CargarCalendario`. Lo que llevan lo garantiza la derivación, que el test de FR-083 compara byte a byte con lo
grabado; un fichero empotrado que no se puede leer sigue la regla genérica, defecto `inesperado` (1), sin regla de carga
propia.

### 1.1 `data/festivos/nacionales-y-autonomicos.yaml`

```yaml
# generado desde evidencias/h9/boe.fiestas-laborales/ por TestFestivosNacionalesDerivados (internal/source/boe/fiestas), no editar
fuente: boe.fiestas-laborales
anios:
  - {anio: 2025, disposicion: BOE-A-2024-21316}
  - {anio: 2026, disposicion: BOE-A-2025-21667}
  - {anio: 2027, disposicion: BOE-A-2026-21254}
festivos:
  - {fecha: "2026-10-12", ambito: nacional, comunidades: ["01", "02", …, "19"], source: BOE-A-2025-21667}
  - {fecha: "…", ambito: autonomico, comunidades: ["07"], source: …}
  - {fecha: "…", ambito: autonomico, comunidades: ["05"], parte: "<la parte que nombra la resolución>", source: …}
```

- Una fila por fecha, ámbito y parte, con las comunidades (códigos de `data/territorio/comunidades/`, ordenados) en las
  que la resolución la marca con ese ámbito; ordenadas por fecha, ámbito y parte. `source` es la resolución de su año.
- `ambito`: `nacional` («***» o «**») o `autonomico` («*»). `parte`, solo si la resolución limita la fiesta a una parte
  de la comunidad (FR-025); esa fila no hace inhábil el día (FR-011).
- Lo que comprueba la derivación: las 19 comunidades tienen festivos cada año, porque FR-020 pide los de todas las que
  trae cada resolución y una columna que no se lee no puede quedar sin rastro.

### 1.2 `data/festivos/locales/13.yaml`

```yaml
# generado desde evidencias/h9/bocm.fiestas-locales/ por TestFestivosLocalesDerivados (internal/source/bocm), no editar
fuente: bocm.fiestas-locales
comunidad: "13"
anios:
  - {anio: 2025, disposiciones: [BOCM-20241213-23, BOCM-20241227-32, BOCM-20250203-16, BOCM-20250217-24]}
  - {anio: 2026, disposiciones: [BOCM-20251212-34, BOCM-20251229-19, BOCM-20260108-18]}
festivos:
  - {fecha: "2026-08-14", municipio: "28074", source: BOCM-20251212-34}
  - {fecha: "2026-10-09", municipio: "28074", source: BOCM-20251212-34}
faltas:
  - {anio: 2026, municipio: "28162", motivo: linea-sin-leer, source: BOCM-20251212-34}
  - {anio: 2026, linea: "Cerceda", motivo: sin-municipio, source: BOCM-20251212-34}
```

- `festivos`: una fila por fecha y municipio (código INE), ordenadas por fecha y municipio; `source` es la última
  disposición del año que fija la línea del municipio.
- `faltas`: por año, cada municipio de la comunidad sin festivos (`no-comunicado`, `linea-sin-leer`, `sin-linea`) y
  cada línea que no casa con un único municipio (`sin-municipio`, con `linea`, el nombre tal como lo escribe el
  boletín); ordenadas por año, municipio y línea. `source`, la disposición de la línea, o la última del año con
  `sin-linea`.
- Cada uno de los 179 municipios de la comunidad 13 de `data/territorio/` tiene en cada año de `anios` festivos o una
  falta, y nunca las dos cosas; cada línea de cada disposición da festivos de un municipio o una falta (FR-022).
- Una comunidad sin fichero en `data/festivos/locales/` no tiene festivos locales cargados: `cobertura` dice
  `sin-festivos-locales` y avisa.

### 1.3 La derivación (FR-020 a FR-025)

| Fichero | Paquete | Lee | Regla |
|---|---|---|---|
| `nacionales-y-autonomicos.yaml` | `internal/source/boe/fiestas` | las tres resoluciones de `evidencias/h9/boe.fiestas-laborales/`, con `httpx.Replay` | research D14; S1, S2 |
| `locales/13.yaml` | `internal/source/bocm` | las siete disposiciones de `evidencias/h9/bocm.fiestas-locales/`, con `httpx.Replay`, y `data/territorio/` | research D13 |

La línea del BOCM (research D13, V8): `.⇥<nombre>[:] <cuerpo>[.]`; el nombre, hasta los dos puntos o, sin ellos, hasta
la primera cifra; el cuerpo, «no comunicado por el ayuntamiento» (sin distinguir mayúsculas) o fechas
`<d>[ y <d>] [de ]<mes>` unidas por « y », con blancos y puntos sueltos tolerados; lo demás, `linea-sin-leer`. El año es
el de la disposición en el manifiesto. El cruce, solo con los municipios de la comunidad del boletín: forma compacta
(sin «de», «del», «la», «las», «los», «el» ni blancos) igual a la de una forma de un único municipio; si ninguno, el
nombre que empieza por todas las palabras de una forma de un único municipio seguidas de «de» o «del»; si no,
`sin-municipio`.

Cada test de derivación escribe su fichero con `-actualizar-festivos` (como `-actualizar-golden`) y, sin la bandera,
falla si el fichero no es byte a byte lo que sale de lo grabado.

## 2. `data/plazos.yaml` (FR-008, FR-030)

Lo lee `plazos.LeerTipos` (binario y generador de `references/recursos_y_plazos.md`, research D17):

```yaml
tipos:
  - codigo: alzada
    nombre: recurso de alzada
    computo: lpac
    duracion: {cantidad: 1, unidad: meses}
    preceptos:
      - {norma: BOE-A-2015-10565, bloque: a122, apartado: "122.1", fija: duracion}
  - codigo: reposicion
    nombre: recurso potestativo de reposición
    computo: lpac
    duracion: {cantidad: 1, unidad: meses}
    preceptos:
      - {norma: BOE-A-2015-10565, bloque: a124, apartado: "124.1", fija: duracion}
  - codigo: contencioso
    nombre: recurso contencioso-administrativo
    computo: contencioso
    duracion: {cantidad: 2, unidad: meses}
    preceptos:
      - {norma: BOE-A-1998-16718, bloque: a46, apartado: "46.1", fija: duracion}
      - {norma: BOE-A-1998-16718, bloque: a128, apartado: "128.2", fija: agosto}
  - codigo: acceso-resolver
    nombre: plazo para resolver una solicitud de acceso a la información pública
    computo: lpac
    duracion: {cantidad: 1, unidad: meses}
    preceptos:
      - {norma: BOE-A-2013-12887, bloque: a20, apartado: "20.1", fija: duracion}
  - codigo: acceso-reclamar
    nombre: reclamación contra la resolución de una solicitud de acceso
    computo: lpac
    duracion: {cantidad: 1, unidad: meses}
    preceptos:
      - {norma: BOE-A-2013-12887, bloque: a24, apartado: "24.2", fija: duracion}
  - codigo: silencio
    nombre: día en que se produce el silencio administrativo
    computo: lpac
    preceptos:
      - {norma: BOE-A-2015-10565, bloque: a21, apartado: "21.3", fija: inicio}
      - {norma: BOE-A-2015-10565, bloque: a24, apartado: "24.4", fija: silencio}
  - codigo: generico
    nombre: plazo genérico
    computo: lpac
    unidades:
      dias-habiles: {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.2", fija: computo}
      dias-naturales: {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.2", fija: computo}
      meses: {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.4", fija: computo}
```

`LeerTipos` exige los siete códigos y ninguno más, `duracion` en los cinco que la fijan y en ningún otro, `unidades`
solo en `generico`, y que `fija: duracion` esté en un precepto de cada tipo con duración.

## 3. `data/cobertura.yaml` (FR-031)

Lo lee `skills.LeerCobertura`, con `schemas/cobertura.yaml.json` y el lector de `internal/skills`
(`ValidarDocumentoYAML`), como `normas.yaml` y `jerarquia.yaml`:

```yaml
regulados_aparte:
  - procedimiento: aplicación de los tributos en materia tributaria y aduanera y su revisión en vía administrativa
    precepto: {norma: BOE-A-2015-10565, bloque: daprimera, apartado: "DA 1ª.2.a"}
    norma: {identificador: BOE-A-2003-23186, nombre: Ley 58/2003, General Tributaria, aplicable: {bloque: a97, apartado: "97"}}
  - procedimiento: gestión, inspección, liquidación, recaudación, impugnación y revisión en materia de Seguridad Social y desempleo
    precepto: {norma: BOE-A-2015-10565, bloque: daprimera, apartado: "DA 1ª.2.b"}
  - procedimiento: procedimientos sancionadores en materia tributaria y aduanera
    precepto: {…, apartado: "DA 1ª.2.c"}
  - procedimiento: procedimientos sancionadores en el orden social
    precepto: {…, apartado: "DA 1ª.2.c"}
  - procedimiento: procedimientos sancionadores en materia de tráfico y seguridad vial
    precepto: {…, apartado: "DA 1ª.2.c"}
    norma: {identificador: <el que da la búsqueda grabada>, nombre: texto refundido de la Ley sobre Tráfico, Circulación de Vehículos a Motor y Seguridad Vial, aplicable: {bloque: a1, apartado: "1"}}
  - procedimiento: procedimientos sancionadores en materia de extranjería
    precepto: {…, apartado: "DA 1ª.2.c"}
  - procedimiento: actuaciones y procedimientos en materia de extranjería y asilo
    precepto: {…, apartado: "DA 1ª.2.d"}
  - procedimiento: procedimientos de impugnación que una ley establece en sustitución del recurso de alzada
    precepto: {norma: BOE-A-2015-10565, bloque: a112, apartado: "112.2"}
  - procedimiento: reclamaciones económico-administrativas
    precepto: {norma: BOE-A-2015-10565, bloque: a112, apartado: "112.4"}
```

Una entrada sin `norma` no la toma de ningún otro sitio (FR-031): la declaración nombra el precepto de la LPAC. El
bloque `a1` del texto refundido de Tráfico es el único de este contrato que no sale de un índice grabado (research S8).

## 4. Lo que se graba y cómo se comprueba (FR-027, FR-030, FR-031, FR-052; research D15, D16)

### 4.1 Manifiestos

| Manifiesto | `fuente` | Qué pide | Test |
|---|---|---|---|
| `internal/source/boe/fiestas/testdata/grabaciones.json` | `boe.fiestas-laborales` | `https://www.boe.es/diario_boe/txt.php?id=<id>` de `BOE-A-2024-21316` (2025), `BOE-A-2025-21667` (2026) y `BOE-A-2026-21254` (2027), a 1 s | `TestGrabarFiestasLaborales` |
| `internal/source/bocm/testdata/grabaciones.json` | `bocm.fiestas-locales` | `https://www.bocm.es/boletin/CM_Orden_BOCM/AAAA/MM/DD/<id>.xml` de las siete disposiciones de FR-020, con su año, a 10 s | `TestGrabarFiestasLocales` |
| `internal/evals/testdata/grabaciones.json` | `boe.legislacion-consolidada` | con la forma del de H5 (`busqueda`, `titulo_empieza_por`, `bloques`, `para`): LPAC (`a24`, `a30`, `a112`, `a122`, `a124`, `daprimera`), LJCA (`a46`, `a128`), LOPJ (`acientoochentaydos`, `acientoochentaytres`, `acientoochentaycinco`), LTAIBG (`a24`), LGT (`a97`), Código Civil (`art5`) y el texto refundido de Tráfico (`a1`, con su búsqueda) | `TestGrabarEvals` |

Los dos primeros escriben en `$KITLEGAL_EVIDENCIAS` (`evidencias/h9/`); el tercero, en
`internal/evals/testdata/boe.legislacion-consolidada/`, solo lo que no está ya en H4 ni en H5.

Cada id de bloque es el que da el índice grabado de su norma (research V20, V29), no uno compuesto con la forma de otra:
la LOPJ escribe sus artículos con palabras («Artículo ciento ochenta y dos») y el Código Civil como `art<n>`. El índice
del texto refundido de Tráfico no está grabado: su `a1` es un supuesto (research S8).

### 4.2 `TestPreceptosDeLosPlazos` (`internal/evals`)

Con la caché preparada desde `UnionDeGrabaciones()` —H4, H5 y la raíz nueva—, lee cada bloque con `boe articulo` en
proceso y comprueba:

1. cada norma y cada bloque de `data/plazos.yaml` y de `data/cobertura.yaml` está grabado (FR-030, FR-031);
2. en cada precepto con `fija: duracion`, el párrafo de su apartado (el que empieza por «<n>.» del número tras el punto,
   o el bloque entero si el apartado no lo tiene) lleva la duración en palabras: «un mes» con 1 mes, «dos meses» con 2
   (FR-030);
3. cada `norma.identificador` de `data/cobertura.yaml` es el que resuelve la búsqueda grabada de su entrada del
   manifiesto, como `TestIdentificadoresDeLasNormas` (FR-031);
4. las palabras clave de cada precepto del cómputo y de la cobertura, que fija esta tabla:

| Precepto | Bloque | Palabras clave |
|---|---|---|
| art. 30 LPAC | `BOE-A-2015-10565` `a30` | «sábados», «domingos», «días hábiles», «mismo día», «último día del mes», «prorrogado» |
| art. 21 LPAC | `a21` | «entrada en el registro» |
| art. 24 LPAC | `a24` | «vencimiento del plazo máximo» |
| DA 1ª LPAC | `daprimera` | «aplicación de los tributos», «Seguridad Social», «tráfico y seguridad vial», «extranjería y asilo» |
| art. 112 LPAC | `a112` | «recurso de alzada», «reclamaciones económico-administrativas» |
| art. 128 LJCA | `BOE-A-1998-16718` `a128` | «agosto» |
| art. 182 LOPJ | `BOE-A-1985-12666` `acientoochentaydos` | «sábados y domingos», «24 y 31 de diciembre», «fiesta nacional», «festivos a efectos laborales» |
| art. 183 LOPJ | `acientoochentaytres` | «24 de diciembre», «6 de enero» |
| art. 185 LOPJ | `acientoochentaycinco` | «Código Civil», «prorrogado» |
| art. 5 CC | `BOE-A-1889-4763` `art5` | «de fecha a fecha» |
| art. 97 LGT | `BOE-A-2003-23186` `a97` | «aplicación de los tributos» |
| art. 1 del texto refundido de Tráfico | `a1` | «sancion» |
| art. 24 LTAIBG | `BOE-A-2013-12887` `a24` | «un mes», «Consejo de Transparencia y Buen Gobierno» |

Una palabra que falta hace fallar el test nombrando el precepto y la palabra: lo que el plan recuerda del texto (research
S4) lo confirma o lo rechaza lo grabado, y el código se ajusta a lo grabado.
