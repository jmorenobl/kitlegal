# T003 — el esquema lo cierra el ejecutor; el volcado del REL, la correspondencia y el registro de la muestra los escribe la persona en la pausa (intento 1)

**Estado**: marcada `[X]` con `make ci` en verde. `schemas/territorio-dir3.yaml.json` está escrito y verificado.
`data/territorio/dir3.yaml` y `gates/verificacion-dir3.md` **no existen todavía**, y la fila `mpt.rel` de
`docs/SOURCES.md` sigue con su «Fecha del fichero» genérica: los tres los escribe una persona en la pausa humana que
este mismo commit dispara, con el manifiesto de abajo («Lo que hace la persona en la pausa»).

## Por qué el reparto es este

Es el mismo de T001 y T002, y esta tarea lo dice tres veces en su propia línea: «todo lo escribe y lo decide una
persona en la pausa», «el ejecutor no descarga, no deriva y no inventa ningún código» y, en el contrato de datos §4,
«**El ejecutor desatendido no descarga nada**». Lo que el ejecutor puede cerrar es el esquema; el material sale del
volcado del REL, que no está en el repositorio y que FR-043 y ADR 0017 prohíben pedir por red.

La mecánica del bucle obliga además a marcar la tarea para que la pausa llegue (`gates/tarea-T001.md`, «Por qué el
intento 1 no podía llegar nunca a la pausa»): `clasificar_datos` solo dispara `gate_humano_datos` cuando ve un fichero
nuevo bajo `schemas/` en el diff **ya commiteado**, y `estado_cierre` solo commitea una tarea marcada `[X]`. Dejarla
`[ ]` a la espera del volcado haría imposible la pausa en la que ese volcado entra.

**Qué decide la persona, no el ejecutor** (contrato de datos §4; spec *Clarifications* Q1; D19): si la muestra revela
discrepancias más allá de casos aislados, la derivación deja de ser regla y **pasa a ser tabla**. Esa conclusión es la
última línea de `verificacion-dir3.md` y no la puede adelantar nadie sin la muestra delante.

## Lo que entra en este commit

### `schemas/territorio-dir3.yaml.json`

| Clave | Forma |
|---|---|
| raíz | `additionalProperties: false`; obligatorias `fecha`, `source`, `correspondencia` |
| `fecha` | texto con `pattern` `^[0-9]{4}-[0-9]{2}-[0-9]{2}$` |
| `source` | texto no vacío (el identificador lo comprobará contra `docs/SOURCES.md` el subtest `fuentes` de T007) |
| `correspondencia` | objeto con `propertyNames.pattern` `^[0-9]{5}$` |
| cada valor | texto con `pattern` `^L01[0-9]{6}$` |

Es la forma que fija la línea de la tarea y el ejemplo de data-model §3.2, con la misma disposición que
`schemas/territorio-municipios.yaml.json` (T001). Tres decisiones de forma que los artefactos no escribían:

1. **`correspondencia` no lleva `minProperties`.** FR-048 deja fuera del fichero todo municipio no verificado, así que
   el número de filas no es una propiedad del formato; un fichero con pocas filas —o con ninguna, si la muestra
   tumbara la regla— sigue siendo un fichero válido, y lo que diga de la cobertura lo dirá el applet. Exigir un mínimo
   aquí sería una restricción no especificada que además podría bloquear un caso legítimo.
2. **El valor no tiene `$defs` propio.** Es un escalar con su `pattern`, no un objeto: un `$defs` de una línea solo
   añadiría indirección (en `territorio-municipios.yaml.json` sí lo hay porque la fila es un objeto de cinco columnas).
3. **La letra solo se acepta en mayúscula.** El fichero congelado guarda la forma canónica, que es la que `DIR3.String()`
   devolverá (data-model §1.2: la letra «se analiza sin distinguir mayúscula y minúscula y se **normaliza a
   mayúscula**»). Ver el aviso para T004 y T007 al final de esta nota.

## Verificación del esquema (este intento)

Ningún control mira todavía este esquema —eso llega en T007—, así que se verificó con un **arnés temporal**
(`internal/skills/territorio_dir3_arnes_temporal_test.go`, borrado antes de `make ci` y no commiteado; patrón de T001 y
T002) sobre el lector común, `skills.CompilarEsquema` + `skills.ValidarDocumentoYAML`. **43 subtests, todos en verde,
ninguno saltado**, primero en rojo por no existir el esquema:

- **Acepta**: el documento mínimo de data-model §3.2 entrecomillado; uno con varias filas; y uno con
  `correspondencia: {}`, que es la decisión 1 de arriba puesta a prueba.
- **Rechaza**: 28 documentos, cada uno con su defecto —fecha sin comillas (timestamp), con forma mala, vacía y como
  número; raíz sin `fecha`, sin `source`, sin `correspondencia` y con clave de más; `source` vacío y como número; clave
  de cuatro y de seis cifras, con letras y sin comillas; `correspondencia` como lista y como texto; valor en minúscula,
  sin la letra, corto, largo, con letras, con espacio final, vacío, de otro nivel (`L02…`), como número y como mapa;
  clave repetida en la raíz y dentro de `correspondencia`—. Cada uno se lee dos veces, en el tipo del fichero y en
  `map[string]any`, para que el rechazo sea del esquema y nunca de la lectura final en el tipo.
- **Mutantes** (10, en memoria, cada uno una copia del esquema con **una** palabra clave quitada): para cada copia se
  exige el **conjunto exacto** de casos que pasa a aceptar, y que siga aceptando los tres válidos. Cada restricción
  vigila lo suyo y nada más:

  | Restricción quitada | Casos que pasa a aceptar |
  |---|---|
  | raíz `additionalProperties` | `clave-de-mas-en-la-raiz` |
  | raíz `required` | `sin-fecha`, `sin-source`, `sin-correspondencia` |
  | `fecha.pattern` | `fecha-sin-comillas`, `fecha-con-forma-mala`, `fecha-vacia` |
  | `fecha.type` | `fecha-como-numero` |
  | `source.minLength` | `source-vacio` |
  | `source.type` | `source-como-numero` |
  | `correspondencia.type` | `correspondencia-como-lista`, `correspondencia-como-texto` |
  | `correspondencia.propertyNames` | `clave-de-cuatro-cifras`, `clave-de-seis-cifras`, `clave-con-letras` |
  | valor `pattern` | los ocho valores mal formados, incluido `L02…` |
  | valor `type` | `valor-como-numero`, `valor-como-mapa` |

  Tres casos los rechaza el **lector** antes de validar —clave sin comillas (mapa con claves que no son texto) y las
  dos claves repetidas (V37)— y por eso ningún mutante los hace pasar: el arnés lo exige explícitamente.

**Ninguna de las restricciones queda sin caso propio**, al contrario que el `minLength` de `url` de T002: las diez
sostienen al menos un rechazo.

**Los códigos del arnés son sintéticos** (`"99999"`, `"98765"`, `L01999990`…). La tarea prohíbe al ejecutor derivar o
inventar un DIR3, y el ejemplo de data-model §3.2 arrastra además el dígito que T001 corrigió (`gates/tarea-T001.md`,
«el dígito de control de Leganés es **5**, no 8»). Lo verificado es la forma del esquema, no ningún dato del REL.

`make ci` en primer plano tras borrar el arnés: **en verde** (exit 0).

## Lo que hace la persona en la pausa

Todo fuera del repositorio, salvo los tres ficheros que escribe dentro (contrato de datos §4, pasos 1 a 6).

1. **Descargar el volcado del REL** por la dirección de su fila de `docs/SOURCES.md`:
   <https://registroentidadeslocales.mpt.es/REL/frontend/export_data/file_export/export_excel/municipios/all/all>
   (`.xls` BIFF8, ~2 MB, campo `NUMERO_INSCRIPCION` con la forma `01PPMMMDC`). La relación del INE **ya no hace falta
   descargarla**: `data/territorio/municipios.yaml` la trae congelada desde la pausa de T001, con el dígito de control
   oficial (`dc`) de cada uno de los 8.132 municipios.
2. **Derivar el DIR3** de cada ayuntamiento anteponiendo `L` al número de inscripción, y quedarse **solo** con las
   filas coherentes: el `PPMMM` del número de inscripción es el código INE de la fila y su última cifra es el `dc` que
   declara `municipios.yaml` para ese código (criterio de entrada de FR-048 y de la *Clarification* Q1). Un municipio
   sin fila en el REL, o con número incoherente, **queda fuera del fichero** y lo declarará `cobertura` en T009.
3. **Verificar la derivación contra DIR3 real** en una muestra que incluya al menos los tres casos de FR-046: un
   municipio **fusionado o renombrado**, uno de **régimen foral** y uno **con entidades locales menores**. Los tres los
   elige la persona con el volcado delante; el repositorio solo puede acotar el segundo:

   | Caso de la muestra | Lo que el repositorio ya sabe |
   |---|---|
   | Régimen foral | Las dos comunidades forales de `data/territorio/comunidades/`: **15** (Comunidad Foral de Navarra, provincia `31`, 272 municipios) y **16** (País Vasco, provincias `01` Araba/Álava, `20` Gipuzkoa y `48` Bizkaia, 252 municipios). Cualquier código INE que empiece por `31`, `01`, `20` o `48` sirve |
   | Fusionado o renombrado | Nada: el repositorio guarda el nombre vigente, no la historia de las variaciones. Sale del material del INE o del propio REL |
   | Con entidades locales menores | Nada: el hito deja las entidades locales menores **fuera de alcance** como entidad propia (spec, *Fuera de alcance*). Aquí solo interesa que el **municipio** que las tiene derive bien su DIR3 |

   El «DIR3 real» de cada fila sale de una consulta pública hecha **fuera del repositorio** —y se nombra en la columna
   «De dónde salió el real»—, nunca de un volcado masivo de DIR3: ninguno es descargable sin WAF, credenciales ni Red
   SARA, y sortearlos está fuera de alcance (ADR 0017, decisión 4; nota al pie `[^rel-dir3]` de `docs/SOURCES.md`).
4. **Escribir `data/territorio/dir3.yaml`**, con una línea por municipio verificado y **todo entrecomillado**:

   ```yaml
   fecha: "AAAA-MM-DD"   # la del volcado del REL
   source: mpt.rel
   correspondencia:
     "PPMMM": "L01PPMMMD"
   ```

   **Las comillas no son de estilo** (comprobado con el lector común, `gates/tarea-T001.md`): un `fecha: 2026-09-20`
   sin comillas se resuelve como `!!timestamp` y se normaliza a `"2026-09-20T00:00:00Z"`, que no casa con el `pattern`;
   una clave `28074:` sin comillas es un entero y se rechaza como «mapa con claves que no son texto». El valor, en
   mayúscula: `l01…` lo rechaza el esquema.
5. **Escribir `specs/008-h6-territorio-skill-legal/gates/verificacion-dir3.md`** con la tabla del contrato de datos §5
   —una fila por municipio de la muestra— y la conclusión:

   ```markdown
   | Municipio | Código INE | DIR3 derivado | DIR3 real | De dónde salió el real | Coincide |
   |---|---|---|---|---|---|
   ```

   La **conclusión** dice cuál de las dos cosas pasa (FR-048, D19): *regla confirmada*, y entonces cuántas filas entran
   y cuántas quedan fuera; o *la derivación pasa a tabla*, si hay discrepancias más allá de casos aislados.
6. **Actualizar la fila `mpt.rel` de `docs/SOURCES.md`** con la fecha del volcado en «Fecha del fichero» (hoy dice
   «Volcado continuo (generado en la fecha de ejecución de la tarea `[datos]`)») y, si la revisión de términos se
   repite ese día, con el «Revisado» correspondiente. **`scripts/verify-sources.sh` no se toca**: su único caso sigue
   siendo `boe articulo` —comprobado en este intento—, y FR-049 prohíbe expresamente añadir uno para estas fuentes,
   que no se piden en red y que el flujo nocturno no puede vigilar.

### Herencia de T002, para decidir en esta misma pausa

`gates/tarea-T002.md` («Pendiente, no bloqueante») dejó anotado que `docs/SOURCES.md` **no tiene fila** para las dos
tablas de códigos del INE (`cod_ccaa.htm` y `cod_provincia.htm`, bajo `/daco/`) de las que salen los nombres de las 19
comunidades y las 52 provincias, y que esa fila entra «en la tarea que ya declara esa ruta (T003)». El ejecutor no la
escribe: una fila de `docs/SOURCES.md` la revisa una persona, que anota en «Revisado» el día en que leyó los términos
y el `robots.txt` (cabecera del propio documento; FR-049). Queda aquí para que la persona la añada —o decida que la
fila `ine.municipios` ya la cubre, por ser la misma publicadora, la misma licencia CC BY 4.0 y la misma ruta
permitida— en el mismo commit de la pausa, que ya toca ese fichero.

### Comprobaciones antes de aprobar la pausa

1. `data/territorio/dir3.yaml` valida contra `schemas/territorio-dir3.yaml.json` con el lector común (en T007 lo hará
   `make ci`).
2. Toda clave de `correspondencia` existe en `data/territorio/municipios.yaml`, y su valor cumple
   `valor[3:8] == clave` y `valor[8] == dc` de esa fila: es exactamente la integridad que `Cargar` exigirá en T006
   (data-model §2.1, punto 2) y hacerla fallar aquí es más barato que en T007.
3. `verificacion-dir3.md` tiene una fila por municipio de la muestra con las seis columnas, cubre los tres casos de
   FR-046 y termina con la conclusión (SC-008).
4. La fila `mpt.rel` de `docs/SOURCES.md` declara la fecha del volcado, y `scripts/verify-sources.sh` no ha cambiado.
5. Confirmarlo en la rama del hito (`git add data/territorio/dir3.yaml docs/SOURCES.md specs/008-h6-territorio-skill-legal/gates/verificacion-dir3.md && git commit`)
   y aprobar la pausa. T004 arranca desde ese commit y no depende de este material.

## Aviso para T004 y T007: mayúscula en el fichero, las dos formas en la entrada

El `pattern` del valor, `^L01[0-9]{6}$`, **solo acepta la mayúscula**, mientras que `AnalizarDIR3` aceptará la letra en
minúscula y la normalizará (data-model §1.2, FR-034). No es una discrepancia, pero el subtest `gramaticas` de T007
—que «exige que los `pattern` de los tres esquemas acepten y rechacen exactamente lo mismo que los analizadores»
(contrato de datos §2, research.md V38)— tiene que comparar el `pattern` contra la **forma canónica**, la que devuelve
`DIR3.String()`, y no contra el conjunto de entradas que `AnalizarDIR3` admite. Lo contrario obligaría a relajar el
esquema del fichero congelado, que es justo lo que la línea de la tarea fija.

## Notas

- No se ejecutó ninguna descarga, ninguna grabación ni ninguna orden de red en esta tarea, y no se escribió ningún
  código DIR3 derivado de un municipio real.
- Los mutantes se comprobaron **en memoria**: no queda ningún fichero suelto de esta tarea fuera del repositorio.
- Sigue en pie la mejora de proceso anotada en T001 y repetida en T002: la línea de una tarea `[datos]` cuya entrega
  completa una persona debería decir la secuencia explícita —«el ejecutor marca la tarea con el esquema; la persona
  añade el material en la pausa y lo confirma en la rama»— en lugar de la cláusula «todo lo escribe y lo decide una
  persona en la pausa», que sin esta nota se lee como una orden de no marcar. Afecta también a T015.

## Resuelto en la pausa (2026-09-21)

`data/territorio/dir3.yaml` (8.132 filas), `gates/verificacion-dir3.md` y las dos filas de `docs/SOURCES.md`.
`scripts/verify-sources.sh` **no se ha tocado**.

**S2 queda resuelto: la derivación es una regla, sin una sola excepción.** La verificación no fue muestral sino
exhaustiva sobre los 8.132 municipios: el `DC` del REL coincide con el dígito de control del INE en el 100 % de las
filas, ningún municipio del INE se queda sin fila en el REL, y las 8.134 filas del volcado son 8.132 municipios más
dos duplicados exactos. Detalle y cifras en `gates/verificacion-dir3.md`.

De los tres casos de FR-046, dos quedan con ejemplo nombrado (fusionado: Oza-Cesuras y Cerdedo-Cotobade; foral:
Pamplona/Iruña y Vitoria-Gasteiz). El tercero, «con entidades locales menores», **no**: el REL no publica volcado de
entidades de ámbito inferior al municipio —doce rutas de exportación probadas, solo responden `municipios`,
`provincias`, `comarcas`, `mancomunidades` e `islas`— y nombrar uno sin fuente sería escribirlo de memoria. La
comprobación exhaustiva lo cubre de hecho; queda anotado en el registro para quien reabra FR-046.

### `docs/SOURCES.md`

1. La fila `mpt.rel` gana la fecha del volcado: **2026-09-21**.
2. Se añade la fila **`ine.codigos-territoriales`** para `cod_ccaa.htm` y `cod_provincia.htm`, de donde salieron los
   nombres de las 19 comunidades y las 52 provincias en la pausa de T002 (la herencia que esta tarea tenía que
   decidir). Se declara en vez de darla por cubierta por `ine.municipios`: es otra dirección y sostiene otro dato. La
   licencia, los términos y la ruta `/daco/` son los mismos que Jorge revisó el 2026-09-18 para `ine.municipios`, así
   que la revisión de fondo no cambia; el «Revisado» dice 2026-09-21 porque es el día en que se comprobaron para esta
   fila.

El `source` de los 19 ficheros de comunidad **sigue siendo `ine.municipios`**, como fija `data-model` §3.4: los
códigos y las provincias salen de ahí, y la procedencia de los nombres queda escrita en `gates/tarea-T002.md` y en
esta fila nueva. Cambiarlo contradiría el artefacto sin gate que lo apruebe.

### Comprobado antes de aprobar

Con el lector real en un `_test.go` temporal borrado antes de `make ci`: `dir3.yaml` valida contra su esquema; las
8.132 claves existen en `municipios.yaml`; y para cada fila `valor[3:8] == clave` y `valor[8] == dc`, que es la
integridad que `Cargar` exigirá en T006. `make ci` en verde.

## Reabierto en la revisión final (2026-09-24)

Los dos jueces rechazaron la verificación de esta pausa (`revision-a.json` y `revision-b.json`, criterios `f` y `h`), y
con razón: la «comprobación exhaustiva» comparaba el número de inscripción del REL con el código INE y su dígito de
control, es decir, la entrada de la derivación consigo misma, y el registro no tenía las columnas «DIR3 real», «De
dónde salió el real» y «Coincide» que pedía el paso 3 de este manifiesto. Además, el caso «con entidades locales
menores» sí tenía fuente: el volcado `…/export_excel/eatimes/all/all` del REL, que esta pausa no encontró.

`gates/verificacion-dir3.md` se reescribió con la verificación que faltaba: siete municipios —los tres casos de FR-046,
entidades locales menores incluidas— contra las fichas de unidad orgánica del directorio del Punto de Acceso General,
siete coincidencias. La comprobación REL↔INE se conserva como lo que es, el criterio de entrada de FR-048. La
conclusión no cambia —regla, 8.132 de 8.132—, pero ahora descansa en una verificación contra DIR3 real. `data/` no
cambia.
