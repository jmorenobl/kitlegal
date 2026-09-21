# Verificación de la derivación INE → DIR3 (T003, pausa `[datos]` de 2026-09-21)

Resuelve el supuesto **S2** del ADR 0017: «el DIR3 del ayuntamiento es `L` + el número de inscripción del REL
(`L01PPMMMDC`)».

## Qué se hizo

Volcado de municipios del REL descargado de la dirección que declara la fila `mpt.rel` de `docs/SOURCES.md`
(`…/export_excel/municipios/all/all`), 2.063.872 bytes, SHA-256
`eeffc6d08d523a5367ee37b57b34119cb6ffe28dd4634d195729748ba406f28c`. Es un libro BIFF8 dentro de un contenedor OLE; se
leyó con la biblioteca estándar (contenedor CFB y registros BIFF, incluidas las cadenas compartidas y sus `CONTINUE`),
fuera del repositorio y sin versionar. Ni el producto ni `make ci` ganan dependencias: el fichero entra congelado.

**Forma del número.** `NUMERO_INSCRIPCION` viaja como número, así que pierde el cero inicial: las 8.134 filas traen 7
cifras. Rellenado a 8 (`01PPMMMDC`), el prefijo es `01` en las 8.134, y el DIR3 es `"L" + ese número`.

## La verificación fue exhaustiva, no muestral

FR-046 pide una muestra. Se comprobó la **población entera**, que es estrictamente más fuerte:

| Comprobación | Resultado |
|---|---|
| Filas del REL con `NUMERO_INSCRIPCION` | 8.134 |
| Filas cuyo número, a 8 cifras, empieza por `01` | 8.134 (100 %) |
| Filas cuyo `PPMMM` existe en `data/territorio/municipios.yaml` | 8.134 (100 %) |
| Filas cuyo `DC` **coincide** con el dígito de control del INE | 8.134 (100 %), **cero discrepancias** |
| Municipios del INE sin fila en el REL | **0** |
| Claves con dos números de inscripción distintos | **0** |

Las 8.134 filas para 8.132 municipios son dos **duplicados exactos** del propio volcado —`18013` Alhama de Granada y
`39083` Soba, repetidos con datos idénticos, misma fecha de inscripción—, no dos municipios más. Se pliegan a una
entrada cada uno.

## Conclusión: **regla**, y con cobertura completa

La derivación es una **regla**, no una tabla de excepciones: `DIR3 = "L" + "01" + <código INE de 5 cifras> + <dígito de
control del INE>`, sin una sola excepción en los 8.132 municipios de España. El dígito de control del REL es el mismo
que publica el INE, así que la integridad que `Cargar` exigirá en T006 —`valor[3:8] == clave` y `valor[8] == dc`— se
cumple por construcción y se comprobó fila a fila.

Aun siendo una regla, `data/territorio/dir3.yaml` se congela **en extenso**, con las 8.132 filas verificadas, como
manda FR-048: el fichero es el registro de lo verificado en esta fecha, no el resultado de recalcular la regla en
ejecución. Ningún municipio queda fuera de cobertura por este motivo.

## Los tres casos de FR-046

| Caso | Municipio | Código INE | `dc` (INE) | Nº inscripción (REL) | DIR3 derivado |
|---|---|---|---|---|---|
| Fusionado | Oza-Cesuras (Oza dos Ríos + Cesuras) | `15902` | `7` | `01159027` | `L01159027` |
| Fusionado | Cerdedo-Cotobade (Cerdedo + Cotobade) | `36902` | `6` | `01369026` | `L01369026` |
| Renombrado (bilingüe) | Pamplona/Iruña | `31201` | `6` | `01312016` | `L01312016` |
| Foral (Navarra) | Pamplona/Iruña | `31201` | `6` | `01312016` | `L01312016` |
| Foral (País Vasco) | Vitoria-Gasteiz | `01059` | `0` | `01010590` | `L01010590` |
| Territorio de validación | Leganés | `28074` | `5` | `01280745` | `L01280745` |

Los dos municipios fusionados llevan código `9xx`, el rango que el INE reserva a las fusiones, y derivan sin
excepción.

**El caso «con entidades locales menores» no tiene ejemplo nombrado, y es lo único que esta verificación no cubre como
FR-046 lo pide.** El REL no publica volcado de entidades de ámbito territorial inferior al municipio: de las rutas de
exportación se probaron doce y solo responden con datos `municipios`, `provincias`, `comarcas`, `mancomunidades` e
`islas`; las demás devuelven 33 bytes. Nombrar un municipio «con entidades locales menores» sin volcado que lo
sostenga sería escribirlo de memoria, que la batería de `tasks.md` prohíbe. **No debilita la conclusión**: el riesgo
que ese caso vigila —que un municipio con entidades menores derive distinto— queda excluido empíricamente porque la
comprobación abarcó los 8.132, no una muestra. Queda anotado para quien reabra FR-046.

## Aviso: el dígito de control de Leganés en los artefactos

`L01280745`, no `L01280748`. Confirma lo que ya avisó la pausa de T001: los artefactos traen `8` escrito de memoria en
cinco sitios (`research.md:128`, `:295`, `:298`; `data-model.md:226`, `:241`; y el caso de forma de
`contracts/identificadores-ine-y-dir3.md:88`, que solo ilustra la forma). Ninguno se ha copiado a `data/` ni a
`testdata/`; los corrige la tarea que toque cada fichero.
