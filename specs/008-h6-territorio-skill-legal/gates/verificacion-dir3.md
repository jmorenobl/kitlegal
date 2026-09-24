# Verificación de la derivación INE → DIR3 (T003; reabierta en la revisión final, 2026-09-24)

Resuelve el supuesto **S2** del ADR 0017: «el DIR3 del ayuntamiento es `L` + el número de inscripción del REL
(`L01PPMMMDC`)». FR-046 pide verificar esa derivación **contra DIR3 real** en una muestra con un municipio fusionado
o renombrado, uno foral y uno con entidades locales menores; FR-047, registrar para cada uno el código derivado, el
código real y de dónde salió el real.

La primera versión de este registro (pausa de T003, 2026-09-21) no lo hacía: derivaba el DIR3 del número de
inscripción del REL y luego comprobaba que ese mismo número contenía el código INE y su dígito de control, es decir,
comparaba la entrada de la derivación consigo misma. Esa comprobación sigue abajo con su nombre —es el **criterio de
entrada** de FR-048, no la verificación— y la verificación contra DIR3 real es la de la sección siguiente.

## 1. Verificación contra DIR3 real (FR-046, FR-047)

**Fuente del código real**: la ficha de unidad orgánica del directorio del Punto de Acceso General
(`administracion.gob.es`), que publica el DIR3. Se pidió por el código derivado
(`https://administracion.gob.es/pagFront/espanaAdmon/directorioOrganigrama/fichaUnidadOrganica.htm?codigoUnidad=<código>`)
y se leyó de la respuesta **qué unidad tiene ese código**: la ficha muestra «Código de unidad orgánica: `<código>`» y,
en la cabecera de la misma tarjeta, el nombre de la unidad. Lo que hace la comprobación no circular es ese nombre: el
directorio dice que el código es el del **ayuntamiento de ese municipio**, y el municipio lo fija el código INE, no el
REL. Un código inexistente, o el de otra entidad, no habría devuelto «Ayuntamiento de» y el nombre del municipio.

| Municipio | Caso de FR-046 | Código INE | DIR3 derivado | DIR3 real | De dónde salió el real | Coincide |
|---|---|---|---|---|---|---|
| Leganés | Territorio de validación | `28074` | `L01280745` | `L01280745`, «Ayuntamiento de Leganés» | Ficha del PAG, 2026-09-24 05:12 GMT | sí |
| Oza-Cesuras | Fusionado (Oza dos Ríos + Cesuras) | `15902` | `L01159027` | `L01159027`, «Ayuntamiento de Oza Cesuras» | Ficha del PAG, 2026-09-24 05:13 GMT | sí |
| Cerdedo-Cotobade | Fusionado (Cerdedo + Cotobade) | `36902` | `L01369026` | `L01369026`, «Ayuntamiento de Cerdedo Cotobade» | Ficha del PAG, 2026-09-24 05:14 GMT | sí |
| Pamplona/Iruña | Foral (Navarra) y renombrado bilingüe | `31201` | `L01312016` | `L01312016`, «Ayuntamiento de Pamplona/Iruña» | Ficha del PAG, 2026-09-24 05:15 GMT | sí |
| Vitoria-Gasteiz | Foral (País Vasco) y con 61 entidades locales menores | `01059` | `L01010590` | `L01010590`, «Ayuntamiento de Vitoria-Gasteiz» | Ficha del PAG, 2026-09-24 05:16 GMT | sí |
| Riello | Con 37 entidades locales menores | `24132` | `L01241324` | `L01241324`, «Ayuntamiento de Riello» | Ficha del PAG, 2026-09-24 05:17 GMT | sí |
| Soba | Fila duplicada en el volcado del REL (y 27 entidades locales menores) | `39083` | `L01390839` | `L01390839`, «Ayuntamiento de Soba» | Ficha del PAG, 2026-09-24 05:18 GMT | sí |

**Siete de siete coinciden; cero discrepancias.** Los tres casos de FR-046 tienen ejemplo nombrado: fusionados,
Oza-Cesuras y Cerdedo-Cotobade (código `9xx`, el rango que el INE reserva a las fusiones); forales, Pamplona/Iruña y
Vitoria-Gasteiz; con entidades locales menores, Riello y Vitoria-Gasteiz. Soba se añadió por ser uno de los dos
municipios que el volcado del REL repite (abajo).

### Las entidades locales menores, desde el REL

El REL publica el volcado de entidades de ámbito territorial inferior al municipio en
`https://registroentidadeslocales.mpt.es/REL/frontend/export_data/file_export/export_excel/eatimes/all/all` —el
plural es `eatimes`; la pausa de T003 no dio con esa ruta y por eso dejó el caso sin ejemplo—. Descargado el
2026-09-22, 636.928 bytes, SHA-256 `298ec243b69346c9c260aefeb474a6758537f741fdeac65797d7daaf6b3b36c3`, con el mismo
lector BIFF8 que el volcado de municipios: **3.674 entidades**, cada una con su provincia y el nombre de su municipio
(columna `MUNICIPIO`; el volcado no trae el código INE del municipio, así que la cuenta es por provincia y nombre).
Vitoria-Gasteiz (Araba/Álava) tiene **61**, la que más de España; Riello (León), **37**; Soba (Cantabria), **27**.

### Cómo se consultó

- **Dentro de su `robots.txt`**: `Crawl-delay: 60`, `Request-rate: 1/1m` y `Visit-time: 0100-0645` GMT. Una petición
  por minuto, solo dentro de la ventana, con User-Agent `kitlegal/0.1 (+https://ventanillalegal.es/bot)`. La
  consulta empezó a las 05:11 y terminó a las 05:19 GMT. La visita previa a la portada del directorio falló por DNS
  y quedó anotada; las siete fichas se pidieron después y respondieron `HTTP 200`.
- **Sin fiarse del código HTTP.** Una primera tanda (2026-09-23, 01:03–01:09 GMT) dio siete `200` y solo tres páginas
  distintas: cinco eran la misma copia byte a byte, servida sin atender a `codigoUnidad`. Se descartó entera. En la
  tanda buena, cada respuesta se aceptó solo si contenía **su propio** código, y las siete son distintas entre sí.
- **Fuera del repositorio y sin versionar**, como el volcado del REL. Las fichas quedan identificadas por su hash:

  | Ficha | SHA-256 |
  |---|---|
  | `L01280745` | `5342395b5a8bacfb1c7744b50b6276eb1f7a0254b4aad937e9e9b7113fb01141` |
  | `L01159027` | `6ba160ac5db815b7927bdc8e279986048bcb6dd92e984e8e093950d8db9eecf5` |
  | `L01369026` | `c1a39ea72c3ff060bf4c4d4cc11aeab9df63962e7c5708358918929fe2654336` |
  | `L01312016` | `1a04ff4b205e295823c6c08a14ceded13c4e3cd86c589fa83d2d4e5c786f158f` |
  | `L01010590` | `95dea00d61e83b7e4017075b5f2811763d504c99ce2fab31328789572629b3e6` |
  | `L01241324` | `eb209ba00bf16913e3f1bef40b6c0040e44da898317dedb49918a5306fbc5204` |
  | `L01390839` | `3dd97014dcd6ec520a0f41ff4b77162d8d4fcbfb9bea1ce1fc44e6d250aa0a9c` |

### `L01…` no es `LA…`

`L01PPMMMDC` es el código de la **entidad local**: la raíz del organigrama del ayuntamiento en DIR3, la que la ficha
del PAG muestra como «Código de unidad orgánica» junto a «Ayuntamiento de …». Las unidades que cuelgan de él llevan
códigos de otra serie, `LA` más siete cifras; son los que un ayuntamiento suele publicar para facturación
electrónica. Las propias fichas lo muestran: la de Leganés enlaza 49 unidades dependientes con código `LA…`, la de
Pamplona/Iruña 22 y la de Vitoria-Gasteiz 20, y ninguna de las tres enlaza otro `L01…` que el suyo. Encontrar un
ayuntamiento con códigos `LA…` y sin ningún `L01…` en un directorio de facturación no contradice la derivación: son
series distintas, y el applet devuelve la de la entidad local.

## 2. Criterio de entrada: coherencia del REL con el INE (FR-048)

Volcado de municipios del REL descargado de la dirección que declara la fila `mpt.rel` de `docs/SOURCES.md`
(`…/export_excel/municipios/all/all`) el 2026-09-21, 2.063.872 bytes, SHA-256
`eeffc6d08d523a5367ee37b57b34119cb6ffe28dd4634d195729748ba406f28c`. Es un libro BIFF8 dentro de un contenedor OLE; se
leyó con la biblioteca estándar (contenedor CFB y registros BIFF, incluidas las cadenas compartidas y sus `CONTINUE`),
fuera del repositorio y sin versionar. Ni el producto ni `make ci` ganan dependencias: el fichero entra congelado.

**Forma del número.** `NUMERO_INSCRIPCION` viaja como número, así que pierde el cero inicial: las 8.134 filas traen 7
cifras. Rellenado a 8 (`01PPMMMDC`), el prefijo es `01` en las 8.134.

Esto no verifica el DIR3 —lo hace la sección 1—: dice a qué filas **se aplica la regla sin discrepancia**, que es lo
que FR-048 pide para entrar en el fichero. Se comprobó sobre la población entera:

| Comprobación | Resultado |
|---|---|
| Filas del REL con `NUMERO_INSCRIPCION` | 8.134 |
| Filas cuyo número, a 8 cifras, empieza por `01` | 8.134 (100 %) |
| Filas cuyo `PPMMM` existe en `data/territorio/municipios.yaml` | 8.134 (100 %) |
| Filas cuyo `DC` coincide con el dígito de control del INE | 8.134 (100 %), cero discrepancias |
| Municipios del INE sin fila en el REL | **0** |
| Claves con dos números de inscripción distintos | **0** |

Las 8.134 filas para 8.132 municipios son dos **duplicados exactos** del propio volcado —`18013` Alhama de Granada y
`39083` Soba, repetidos con datos idénticos, misma fecha de inscripción—, no dos municipios más. Se pliegan a una
entrada cada uno.

## Conclusión: **regla**, verificada contra DIR3 real, con cobertura completa

La derivación es una **regla**, no una tabla de excepciones: `DIR3 = "L" + "01" + <código INE de 5 cifras> + <dígito
de control del INE>`. La verifica la muestra de la sección 1 —siete municipios, los tres casos de FR-046, siete
coincidencias con el directorio oficial— y se aplica sin discrepancia a los **8.132 de 8.132** municipios de España
(sección 2). FR-048 separa las dos cosas: la muestra verifica la regla, no cada fila; y entra en el fichero todo
municipio al que la regla se aplica sin discrepancia. Por eso `cobertura.dir3: verificado` es cierto para los 8.132, y
ninguno queda fuera de cobertura por este motivo.

Aun siendo una regla, `data/territorio/dir3.yaml` se congela **en extenso**, con las 8.132 filas, como manda FR-048:
el fichero es el registro de lo verificado en esta fecha, no el resultado de recalcular la regla en ejecución. La
integridad que `Cargar` exige —`valor[3:8] == clave` y `valor[8] == dc`— se cumple por construcción.

## El dígito de control de Leganés en los artefactos

`L01280745`, no `L01280748`. Los artefactos del plan traían `8` escrito de memoria (`research.md` y `data-model.md`);
los corrigió la revisión final (`c481451`). `L01280748` sigue apareciendo solo como cadena de **forma** —bien formada,
sin nombrar a ningún municipio—: en los casos de `contracts/identificadores-ine-y-dir3.md`, en los tests de
`internal/core/ids` y `internal/skills` y en el corpus de fuzz. Ninguno lo presenta como el DIR3 de Leganés.
