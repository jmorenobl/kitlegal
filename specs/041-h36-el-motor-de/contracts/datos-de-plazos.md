# Contrato: `data/plazos.yaml`, los ficheros de prueba y `LeerTipos`

La forma y las reglas son las de spec §C.1 y §C.2. Este contrato escribe completo lo que §C.1 resume y fija cómo se
prueba cada regla. Lo demás de `specs/023-h9-internal-core-plazos/contracts/datos-de-plazos.md` sigue en vigor
(FR-018).

## 1. `data/plazos.yaml` de la rama

Los preceptos, sus bloques, sus `fija` y sus `vigente_desde` de hoy no cambian; cada tipo gana `da` detrás de
`duracion` (si la lleva) y delante de `preceptos` o `unidades` (§C.1); cada cómputo gana `municipios`, `unidades` y,
si los tiene, `periodos`; y entra el civil. El precepto de un periodo no lleva `bloque`, así que `bloque: a128,` sigue
saliendo una sola vez y los fragmentos que sustituyen los tests de H9 y de H26.1 siguen casando (spec §C.1).

```yaml
# Tipos de plazo de `kitlegal plazos calcular` y los cómputos con que se cuentan. Lo lee plazos.LeerTipos, el binario y
# el generador de la región recursos_y_plazos de la referencia de cada paquete, hoy
# skills/plazos-y-territorio/references/administrativo.md.
# Cada tipo nombra su cómputo, dice qué fecha da (da: vencimiento, o dia-siguiente, el día siguiente al vencimiento),
# lleva la duración que fija o la recibe con una unidad de su cómputo, y los preceptos que lo fundan, cada uno con lo
# que fija. Cada cómputo dice los municipios que cuentan, con el texto que nombra a cada uno, las unidades que admite,
# sus periodos del año (inhabil, no-corre o aviso, cada uno con su precepto) y los preceptos con que se cuenta
# cualquier plazo suyo (ADR 0061). Cada precepto lleva también vigente_desde: el primer día en que su apartado dice lo
# que dice en la última redacción grabada, con la que está escrito este fichero; plazos no da la fecha de un plazo que
# empieza antes de ese día en alguno de sus preceptos (ADR 0042). Sin texto de ninguna norma: cada precepto se nombra
# por su identificador, su bloque y su apartado, y TestPreceptosDeLosPlazos (internal/evals) comprueba contra lo
# grabado del BOE que cada bloque existe, que el apartado que fija la duración la dice en palabras, que cada precepto
# lleva sus palabras clave y que vigente_desde es el que dan las redacciones grabadas.
tipos:
  - codigo: alzada
    nombre: recurso de alzada
    computo: lpac
    duracion: {cantidad: 1, unidad: meses}
    da: vencimiento
    preceptos:
      - {norma: BOE-A-2015-10565, bloque: a122, apartado: "122.1", fija: duracion, vigente_desde: "2016-10-02"}
  - codigo: reposicion
    nombre: recurso potestativo de reposición
    computo: lpac
    duracion: {cantidad: 1, unidad: meses}
    da: vencimiento
    preceptos:
      - {norma: BOE-A-2015-10565, bloque: a124, apartado: "124.1", fija: duracion, vigente_desde: "2016-10-02"}
  - codigo: contencioso
    nombre: recurso contencioso-administrativo
    computo: contencioso
    duracion: {cantidad: 2, unidad: meses}
    da: vencimiento
    preceptos:
      - {norma: BOE-A-1998-16718, bloque: a46, apartado: "46.1", fija: duracion, vigente_desde: "1998-12-14"}
      - {norma: BOE-A-1998-16718, bloque: a128, apartado: "128.2", fija: agosto, vigente_desde: "1998-12-14"}
  - codigo: acceso-resolver
    nombre: plazo para resolver una solicitud de acceso a la información pública
    computo: lpac
    duracion: {cantidad: 1, unidad: meses}
    da: vencimiento
    preceptos:
      - {norma: BOE-A-2013-12887, bloque: a20, apartado: "20.1", fija: duracion, vigente_desde: "2014-12-10"}
  - codigo: acceso-reclamar
    nombre: reclamación contra la resolución de una solicitud de acceso
    computo: lpac
    duracion: {cantidad: 1, unidad: meses}
    da: vencimiento
    preceptos:
      - {norma: BOE-A-2013-12887, bloque: a24, apartado: "24.2", fija: duracion, vigente_desde: "2014-12-10"}
  - codigo: silencio
    nombre: día en que se produce el silencio administrativo
    computo: lpac
    da: dia-siguiente
    preceptos:
      - {norma: BOE-A-2015-10565, bloque: a21, apartado: "21.3", fija: inicio, vigente_desde: "2016-10-02"}
      - {norma: BOE-A-2015-10565, bloque: a24, apartado: "24.4", fija: silencio, vigente_desde: "2016-10-02"}
  - codigo: generico
    nombre: plazo genérico
    computo: lpac
    da: vencimiento
    unidades:
      dias-habiles: {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.2", fija: computo, vigente_desde: "2016-10-02"}
      dias-naturales: {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.2", fija: computo, vigente_desde: "2016-10-02"}
      meses: {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.4", fija: computo, vigente_desde: "2016-10-02"}
computos:
  - codigo: lpac
    municipios: [el del interesado, el de la sede del órgano]
    unidades: [meses, dias-habiles, dias-naturales]
    preceptos:
      - {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.2", fija: habiles, vigente_desde: "2016-10-02"}
      - {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.3", fija: inicio, vigente_desde: "2016-10-02"}
      - {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.4", fija: meses, vigente_desde: "2016-10-02"}
      - {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.5", fija: prorroga, vigente_desde: "2016-10-02"}
      - {norma: BOE-A-2015-10565, bloque: a30, apartado: "30.6", fija: municipios, vigente_desde: "2016-10-02"}
  - codigo: contencioso
    municipios: [el de la sede del juzgado]
    unidades: [meses]
    periodos:
      - {desde: "08-01", hasta: "08-31", efecto: no-corre, precepto: {norma: BOE-A-1998-16718, apartado: "128.2"}}
      - desde: "12-24"
        hasta: "01-06"
        efecto: aviso
        precepto: {norma: BOE-A-1985-12666, apartado: "183"}
        aviso:
          codigo: dias-del-art-183-lopj
          etiqueta: DÍAS DEL ART. 183 LOPJ
          texto: "el plazo pasa por días del 24 de diciembre al 6 de enero, que el art. 183 LOPJ declara inhábiles para las actuaciones judiciales; se han contado, salvo los inhábiles del art. 182.1 LOPJ, porque el art. 128.2 LJCA no los excluye del plazo para interponer el recurso. La fecha es la más temprana posible."
    preceptos:
      - {norma: BOE-A-1985-12666, bloque: acientoochentaydos, apartado: "182.1", fija: inhabiles, vigente_desde: "2022-12-23"}
      - {norma: BOE-A-1985-12666, bloque: acientoochentaytres, apartado: "183", fija: aviso, vigente_desde: "2022-12-23"}
      - {norma: BOE-A-1985-12666, bloque: acientoochentaycinco, apartado: "185.1", fija: meses, vigente_desde: "1985-07-03"}
      - {norma: BOE-A-1985-12666, bloque: acientoochentaycinco, apartado: "185.2", fija: prorroga, vigente_desde: "1985-07-03"}
      - {norma: BOE-A-1889-4763, bloque: art5, apartado: "5.1", fija: meses, vigente_desde: "1974-07-29"}
  - codigo: civil
    municipios: [el de la sede del tribunal]
    unidades: [meses, dias-habiles]
    periodos:
      - {desde: "08-01", hasta: "08-31", efecto: inhabil, precepto: {norma: BOE-A-2000-323, apartado: "130.2"}}
      - {desde: "12-24", hasta: "01-06", efecto: inhabil, precepto: {norma: BOE-A-2000-323, apartado: "130.2"}}
    preceptos:
      - {norma: BOE-A-2000-323, bloque: a130, apartado: "130.2", fija: inhabiles, vigente_desde: "2022-12-23"}
      - {norma: BOE-A-2000-323, bloque: a133, apartado: "133.1", fija: expiracion, vigente_desde: "2001-01-08"}
      - {norma: BOE-A-2000-323, bloque: a133, apartado: "133.2", fija: habiles, vigente_desde: "2010-05-04"}
      - {norma: BOE-A-2000-323, bloque: a133, apartado: "133.3", fija: meses, vigente_desde: "2001-01-08"}
      - {norma: BOE-A-2000-323, bloque: a133, apartado: "133.4", fija: prorroga, vigente_desde: "2010-05-04"}
```

El texto del aviso es, byte a byte, el de `textoDelArt183` de hoy (`internal/core/plazos/avisos.go:46-48`; spec §C.5);
lo fija el golden del caso 15 (contracts/aceptacion.md §2). La copia del contrato de `internal/core/plazos`
(`plazosDelContrato`, `calcular_test.go`) pasa a ser este fichero sin la cabecera (FR-017).

**El civil entra cuando hay con qué comprobarlo.** `TestPreceptosDeLosPlazos` lee los cómputos sin tipos (FR-015) y
necesita grabados `a130` y `a133`, así que el cómputo `civil` entra en `data/plazos.yaml` en la tarea que añade la raíz
de H36 a la unión, después de la grabación (contracts/grabacion.md §3).

## 2. Las reglas de `LeerTipos` y sus defectos

Cada defecto es un `defectoDeDatos` de `data/plazos.yaml`, sin clase de usuario (el kernel lo da como `inesperado`,
código 1), y nombra el tipo o el cómputo que lo tiene. Se acumulan todos, en el orden del fichero, con `errors.Join`,
como hoy. Los textos son la forma; los tests exigen que el defecto nombre el tipo o el cómputo y lo que falla, no sus
bytes.

| Regla (§C.2) | Forma del defecto | Copia rota de su subtest (sobre la copia del contrato) |
|---|---|---|
| 1 · el cómputo del tipo existe | `el tipo "<tipo>" nombra el cómputo "<cómputo>", que no está en computos` | `alzada` con `computo: inexistente` |
| 2 · la unidad de la duración la admite su cómputo | `la unidad "<unidad>" de la duración del tipo "<tipo>" no la admite su cómputo, "<cómputo>"` | `contencioso` con `unidad: dias-habiles` |
| 3 · las `unidades` del tipo son las de su cómputo, sin `duracion` | `el tipo "<tipo>" no lleva la unidad "<unidad>" de su cómputo, "<cómputo>"`, `el tipo "<tipo>" lleva la unidad "<unidad>", que su cómputo, "<cómputo>", no admite`, `el tipo "<tipo>" lleva unidades y duración` | `generico` sin la entrada `meses` |
| 4 · el precepto de cada periodo es uno del cómputo o de cada tipo suyo | `el precepto del periodo <n> del cómputo "<cómputo>", <norma> <apartado>, no está entre los del cómputo ni entre los de cada uno de sus tipos` | el periodo de agosto de `contencioso` con `apartado: "128.3"` |
| 5 · un `no-corre` es un mes entero en un cómputo de solo `meses` | `el periodo <n> del cómputo "<cómputo>" no corre y no es un mes entero de 30 o 31 días`, `el periodo <n> del cómputo "<cómputo>" no corre y el cómputo admite otras unidades que meses` | el de agosto de `contencioso` con `hasta: "08-30"` |
| 6 · dos periodos no comparten ningún día | `los periodos <n> y <m> del cómputo "<cómputo>" comparten el <MM-DD>` | el segundo de `civil` con `desde: "08-31"` |
| 7 · el aviso de un periodo es único y no es del motor | `el aviso "<código>" del periodo <n> del cómputo "<cómputo>" repite el código de otro aviso`, `… repite la etiqueta …`, `… es un aviso del motor` | el aviso de `contencioso` con `codigo: festivo-de-una-parte` |
| 8 · gramática de los códigos | `el código "<código>" del tipo <n>` (o `del cómputo <n>`, o `del aviso del periodo <n> del cómputo "<cómputo>"`) `no es una o más palabras de minúsculas y cifras unidas por guiones` | `alzada` con `codigo: Alzada` |
| 9 · la forma de un cómputo y de sus periodos | `el cómputo "<cómputo>" no lleva uno ni dos municipios`, `… lleva un municipio vacío`, `… no lleva unidades`, `… repite la unidad "<u>"`, `… lleva la unidad "<u>", que no es meses, dias-habiles ni dias-naturales`, `el periodo <n> del cómputo "<cómputo>" no tiene <desde, hasta, efecto o precepto>`, `… tiene un <desde o hasta>, "<x>", que no es MM-DD de un día de todos los años`, `… tiene un efecto, "<x>", que no es inhabil, no-corre ni aviso`, `… es de aviso y no lleva su aviso`, `… no es de aviso y lleva un aviso`, `el aviso del periodo <n> del cómputo "<cómputo>" no tiene <código, etiqueta o texto>`, `… tiene una etiqueta con dos puntos`, `… tiene un salto de línea` | `lpac` con tres municipios |
| 10 · `da` | `el tipo "<tipo>" no lleva da`, `el tipo "<tipo>" da "<x>", que no es vencimiento ni dia-siguiente` | `alzada` sin su línea `da` |

Siguen los defectos de hoy que no dependen de un código (tipo sin código o repetido, sin nombre, sin preceptos; duración
menor que 1, unidad fuera del vocabulario o sin precepto que la fije; cómputo repetido o sin preceptos; precepto sin uno
de sus cinco datos o con un `vigente_desde` que no es `AAAA-MM-DD`). Desaparecen los que dependen de los siete códigos
y de los dos cómputos: «no es uno de los siete», «falta el tipo», «no es lpac ni contencioso», «falta el cómputo», «fija
su duración y no la lleva», «no fija su duración y lleva una», «lleva unidades, que solo lleva generico».

`TestLeerTipos` lleva, además de un subtest por regla: la copia del contrato, que da los siete tipos sin defecto y el
civil sin tipos (FR-013); y una copia con los defectos de las reglas 1 y 10 a la vez, cuyo error los da los dos (US5.3).
Que el `data/plazos.yaml` real carga y ningún tipo lleva el civil lo comprueba `TestComputoCivil` en `internal/app`
(US2.5, US5.1).

## 3. `internal/core/plazos/testdata/computo-de-ensayo.yaml` (FR-022; spec §C.10)

```yaml
# Cómputo de ensayo de H36 (spec §C.10): un cómputo inventado, que no es de ninguna materia, con sus preceptos
# inventados de la norma BOE-A-0000-1. Lo lee LeerTipos en TestComputoDeEnsayo, sobre la rejilla de las propiedades, y
# sus códigos los busca TestNingunCodigoDePlazosEnGo en el Go del binario.
tipos:
  - codigo: ensayo-fijo
    nombre: tipo de ensayo con duración fija
    computo: ensayo
    duracion: {cantidad: 5, unidad: dias-habiles}
    da: dia-siguiente
    preceptos:
      - {norma: BOE-A-0000-1, bloque: a1, apartado: "1.1", fija: duracion, vigente_desde: "2000-01-01"}
  - codigo: ensayo-libre
    nombre: tipo de ensayo sin duración
    computo: ensayo
    da: vencimiento
    preceptos:
      - {norma: BOE-A-0000-1, bloque: a1, apartado: "1.2", fija: computo, vigente_desde: "2000-01-01"}
computos:
  - codigo: ensayo
    municipios: [el del lugar del ensayo]
    unidades: [dias-habiles, dias-naturales]
    periodos:
      - {desde: "04-06", hasta: "04-10", efecto: inhabil, precepto: {norma: BOE-A-0000-1, apartado: "2.1"}}
      - desde: "10-19"
        hasta: "10-23"
        efecto: aviso
        precepto: {norma: BOE-A-0000-1, apartado: "2.2"}
        aviso:
          codigo: periodo-de-ensayo
          etiqueta: PERIODO DE ENSAYO
          texto: "el plazo pasa por el periodo de ensayo, del 19 al 23 de octubre; sus días se han contado."
    preceptos:
      - {norma: BOE-A-0000-1, bloque: a2, apartado: "2.1", fija: inhabiles, vigente_desde: "2000-01-01"}
      - {norma: BOE-A-0000-1, bloque: a2, apartado: "2.2", fija: aviso, vigente_desde: "2000-01-01"}
```

`TestComputoDeEnsayo` (`internal/core/plazos/ensayo_test.go`) lo empotra con `//go:embed`, lo lee con `LeerTipos` y
calcula E1 a E13 de §C.10 con el calendario de `nuevaRejilla` y los municipios de `municipiosDeLaRejilla` (`81001`
salvo E11, `82001`, y E13, `81001` y `81002`). Cada fila compara la fecha, los códigos de los avisos en su orden y, en E7
y E8, el texto del aviso; E12 y E13, el código 2 (clase `argumentos`) y el mensaje de §C.10.

## 4. `internal/core/plazos/testdata/tipos-de-prueba-civil.yaml` (FR-023)

```yaml
# Tipos de prueba de H36 (FR-023): cuentan con el cómputo civil de data/plazos.yaml, sin duración y con da:
# vencimiento. Ningún tipo de data/plazos.yaml cuenta con el civil, que es de H9.1. Los tests sustituyen la línea
# tipos: de su fichero base por este fichero entero (research D7).
tipos:
  - codigo: prueba-civil-dias
    nombre: tipo de prueba del cómputo civil en días hábiles
    computo: civil
    da: vencimiento
    preceptos:
      - {norma: BOE-A-2000-323, bloque: a133, apartado: "133.2", fija: habiles, vigente_desde: "2010-05-04"}
  - codigo: prueba-civil-meses
    nombre: tipo de prueba del cómputo civil en meses
    computo: civil
    da: vencimiento
    preceptos:
      - {norma: BOE-A-2000-323, bloque: a133, apartado: "133.3", fija: meses, vigente_desde: "2001-01-08"}
```

`TestComputoCivil` (`internal/app/plazos_civil_test.go`) lo combina con `data/plazos.yaml` (`fuentesDePlazos(t).Tipos`),
carga el calendario de `data/festivos/` del binario, resuelve Madrid (y, en C8, Leganés) con el registro de territorio y
calcula cada fila de §C.9 con `plazos.Calcular`: la fecha, los festivos saltados (fecha y ámbito, y el municipio y la
disposición en el local de C10), los códigos de los avisos o el código, la clase y el mensaje. La sede es Madrid
(`28079`) en todos, salvo los dos municipios de C8.

`TestPropiedadesDelComputo` lo combina con la copia del contrato y gana, sobre la misma rejilla y con un municipio cada
vez, las tres propiedades de FR-024: en días hábiles con `prueba-civil-dias` y en meses con `prueba-civil-meses`, para
cada inicio de 2025 a 2027 y cada duración de 1 a 40. El oráculo de la rejilla gana «día de un periodo `inhabil` del
civil» (agosto y del 24 de diciembre al 6 de enero), que se escribe en el test sin pasar por el motor, como los
festivos de la rejilla.
