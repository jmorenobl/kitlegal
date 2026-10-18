# Contrato del applet `plazos`

Lo que fija este contrato lo comprueban los guiones de aceptación (`aceptacion/`), los golden de
`internal/app/testdata/plazos/`, `schemas/plazo.json` y los tests de `internal/core/plazos` e `internal/app`
(plan.md, «Controles mecánicos»). Decisiones: research D2 a D11.

## 1. Verbo y argumentos (FR-001, FR-012)

```text
kitlegal plazos calcular <tipo> <inicio> <municipios>... [--duracion <n>] [--unidad <unidad>]
```

| Argumento | Forma | Qué es |
|---|---|---|
| `tipo` | posición | `alzada`, `reposicion`, `contencioso`, `acceso-resolver`, `acceso-reclamar`, `silencio` o `generico` (research D3) |
| `inicio` | posición | `AAAA-MM-DD`: la de la notificación o la publicación; en `silencio`, la del acuerdo de iniciación o la de entrada de la solicitud en el registro del órgano competente. `plazos` no la deduce |
| `municipios` | posición, uno o más | por su nombre o su código INE, como en `territorio resolver`. En un plazo de la LPAC, dos: el del interesado y el de la sede del órgano, que pueden ser el mismo. En `contencioso`, uno: el de la sede del juzgado |
| `--duracion` | bandera | un entero positivo; solo, y siempre, en `silencio` y `generico` |
| `--unidad` | bandera | `meses`, `dias-habiles` o `dias-naturales`; solo, y siempre, en `silencio` y `generico` |

La herramienta MCP es `plazos_calcular`, generada desde `--describe` como las demás: `tipo`, `inicio`, `municipios`
(lista de cadenas), `duracion` y `unidad`, con `x-banderas: ["duracion", "unidad"]` (research V13). Ninguna bandera
propia más; hereda las ocho globales.

Los tipos, con su cómputo, su duración y su precepto (los nombra `data/plazos.yaml`; contracts/datos-de-plazos.md §2):

| Tipo | Cómputo | Duración | Precepto que la fija |
|---|---|---|---|
| `alzada` | `lpac` | 1 mes | art. 122.1 LPAC (`BOE-A-2015-10565`, `a122`) |
| `reposicion` | `lpac` | 1 mes | art. 124.1 LPAC (`a124`) |
| `contencioso` | `contencioso` | 2 meses | art. 46.1 LJCA (`BOE-A-1998-16718`, `a46`) y art. 128.2 LJCA (`a128`) |
| `acceso-resolver` | `lpac` | 1 mes | art. 20.1 LTAIBG (`BOE-A-2013-12887`, `a20`) |
| `acceso-reclamar` | `lpac` | 1 mes | art. 24.2 LTAIBG (`a24`) |
| `silencio` | `lpac` | la que se da | art. 21.3 LPAC (`a21`) y art. 24.4 LPAC (`a24`) |
| `generico` | `lpac` | la que se da | art. 30.2 LPAC en días hábiles y naturales, art. 30.4 en meses (`a30`) |

## 2. El cómputo (FR-002 a FR-011)

**LPAC** (`computo: lpac`; research D7). Inhábil: sábado, domingo, festivo nacional o autonómico (sin `parte`) de la
comunidad de alguno de los municipios, o festivo local de alguno. Se cuenta desde el día siguiente a `inicio`:

- días hábiles: la fecha es el enésimo día hábil;
- días naturales: el enésimo día y, si es inhábil, el primer hábil siguiente;
- meses: el mismo día del mes de vencimiento, o el último de ese mes si no lo hay, y, si es inhábil, el primer hábil
  siguiente.

`silencio`: el vencimiento del plazo máximo se calcula así con la duración y la unidad dadas, y la fecha que se da es el
día siguiente al vencimiento (research D8).

**Contencioso** (`computo: contencioso`; research D9). Un municipio, el de la sede. Si `inicio` o el día siguiente es
de agosto, el cómputo parte del 31 de agosto. Dos meses de fecha a fecha; si entre el día siguiente al punto de partida
y ese vencimiento cae algún día de agosto, el vencimiento pasa al mismo día del mes siguiente (o al último de ese mes).
Si el último día es sábado, domingo, fiesta nacional o festivo de la comunidad o de la localidad de la sede, los
inhábiles que enumera el art. 182.1 LOPJ grabado, se prorroga al primer día que no lo es. Los días del 24 de diciembre
al 6 de enero se cuentan, y el 24 y el 31 de diciembre no son inhábiles por serlo: el art. 182.1 grabado, en la
redacción de la LO 14/2022, ya no los enumera (research S4; T005, `gates/supuestos.md`).

**Festivos saltados** (`festivos`): en días hábiles, los festivos del tramo que no se contaron; en días naturales y en
meses, los de la prórroga. Nunca un sábado o un domingo, que no son festivos. Solo los hay de los años cargados.

## 3. El sobre (FR-013)

`fuente` `kitlegal.plazos`; `url` `kitlegal:applet/plazos`; `fecha_consulta`, la del reloj del kernel; `hash`, el de
`data`. Con la misma entrada y el mismo binario, `data` y `hash` son iguales, con `--offline` y sin él. Ninguna
operación de grafo, ninguna petición, ninguna escritura en la caché ni en el grafo.

## 4. `data` con el código 0 (FR-001)

| Clave | Tipo | Siempre | Qué es |
|---|---|---|---|
| `fecha` | cadena `AAAA-MM-DD` | sí | la que se da |
| `tipo` | cadena | sí | el pedido |
| `preceptos` | lista de `{norma, bloque, apartado}` | sí, uno o más | el que fija el plazo; en `contencioso`, también el 128.2 LJCA |
| `municipios` | lista de `{codigo_ine, nombre}` | sí | los que cuentan, una vez cada uno, en el orden dado, como los da `territorio resolver` |
| `festivos` | lista de `{fecha, ambito, municipio?, source}` | sí, puede ir vacía | los saltados, en orden de fecha; `ambito` es `nacional`, `autonomico` o `local`; `municipio` (código INE) solo en los locales |
| `avisos` | lista de `{codigo, texto}` | sí, puede ir vacía | §5, en el orden de los códigos de §5 y, dentro de uno, de fecha y de municipio |
| `cobertura` | lista de `{ambito, municipio?, desde, hasta, disposicion?, falta?}` | sí, una o más | por ámbito (`nacional-y-autonomico`, o `local` de cada municipio) y por tramo de años del tramo del plazo: `disposicion` es la última aplicada de ese año; `falta`, por qué no hay festivos (`sin-cargar`, `sin-festivos-locales`, `no-comunicado`, `linea-sin-leer`, `sin-linea`) |

Son las claves que enumera FR-001 y ninguna más: ni la fecha de inicio, ni la duración, ni el cómputo, que quien llama
ya tiene o lee en `references/recursos_y_plazos.md` (research D5).

Ejemplo, el caso (a), con `--json` (772 bytes con el `hash` entero; research M1):

```json
{"ok":true,"fuente":"kitlegal.plazos","url":"kitlegal:applet/plazos","fecha_consulta":"2026-10-18T09:00:00+02:00","hash":"sha256:…","data":{"fecha":"2026-10-13","tipo":"reposicion","preceptos":[{"norma":"BOE-A-2015-10565","bloque":"a124","apartado":"124.1"}],"municipios":[{"codigo_ine":"28074","nombre":"Leganés"}],"festivos":[{"fecha":"2026-10-09","ambito":"local","municipio":"28074","source":"BOCM-20251212-34"},{"fecha":"2026-10-12","ambito":"nacional","source":"BOE-A-2025-21667"}],"avisos":[],"cobertura":[{"ambito":"nacional-y-autonomico","desde":2026,"hasta":2026,"disposicion":"BOE-A-2025-21667"},{"ambito":"local","municipio":"28074","desde":2026,"hasta":2026,"disposicion":"BOCM-20260108-18"}]}}
```

El caso (b), Tordesillas, lleva `"festivos":[]`, un aviso y `{"ambito":"local","municipio":"47165","desde":2026,
"hasta":2026,"falta":"sin-festivos-locales"}` (858 bytes; M2).

## 5. Avisos (FR-010, FR-011, FR-026, FR-048)

Cada aviso solo se da si el tramo —del día siguiente a `inicio` hasta `fecha`— pasa por lo que lo causa. Su `texto` es
una línea con la forma fija `⚠ <ETIQUETA>:` (la de `internal/evals`, research V18):

| Código | Etiqueta | Texto |
|---|---|---|
| `faltan-festivos-nacionales-y-autonomicos` | `FALTAN FESTIVOS NACIONALES Y AUTONÓMICOS` | `⚠ FALTAN FESTIVOS NACIONALES Y AUTONÓMICOS: kitlegal no tiene los de <años>; la fecha se ha calculado sin ellos y puede ser anterior a la real. Los que tiene llegan a <año> (<disposición>).` |
| `faltan-festivos-locales` | `FALTAN FESTIVOS LOCALES` | `⚠ FALTAN FESTIVOS LOCALES: kitlegal no tiene los de <municipio> (<código INE>) de <años><motivo>; la fecha se ha calculado sin ellos y puede ser anterior a la real.` y, si la comunidad tiene festivos locales cargados, ` Los de <comunidad> llegan a <año> (<disposición>).` |
| `dias-del-art-183-lopj` | `DÍAS DEL ART. 183 LOPJ` | `⚠ DÍAS DEL ART. 183 LOPJ: el plazo pasa por días del 24 de diciembre al 6 de enero, que el art. 183 LOPJ declara inhábiles para las actuaciones judiciales; se han contado, salvo los inhábiles del art. 182.1 LOPJ, porque el art. 128.2 LJCA no los excluye del plazo para interponer el recurso. La fecha es la más temprana posible.` |
| `festivo-de-una-parte` | `FESTIVO DE UNA PARTE` | `⚠ FESTIVO DE UNA PARTE: el <fecha> es fiesta solo en <parte> (<comunidad>), según <source>; kitlegal no sabe si <municipio> está ahí y lo ha contado como hábil. La fecha es la más temprana posible.` |

`<años>` es un año («2027»), un tramo («de 2028 a 2030») o «anteriores a 2025»; `<motivo>`, vacío con `sin-cargar`, y
con los demás: «: no tiene festivos locales de <comunidad>» (`sin-festivos-locales`), «: el boletín los da como no
comunicados por el ayuntamiento» (`no-comunicado`), «: su línea del boletín no se lee sin interpretarla»
(`linea-sin-leer`) y «: el boletín no trae su línea» (`sin-linea`). Un aviso de cobertura por ámbito y municipio con
todos sus años; uno de una parte por cada festivo de una parte del tramo. Ninguno nombra una entidad menor (FR-024).

## 6. Errores (FR-012; ADR 0023)

Con `ok` falso, `data` es `{clase, mensaje}`. En este orden; el mensaje nombra la entrada:

| Defecto | Código | Clase |
|---|---|---|
| el tipo no es uno de §1 | 2 | `argumentos` |
| `inicio` no es `AAAA-MM-DD` o no es un día del calendario | 2 | `argumentos` |
| `--duracion` falta en `silencio` o `generico`, no es un entero positivo, o se da a otro tipo | 2 | `argumentos` |
| `--unidad` falta con `--duracion`, no es del vocabulario, o se da a otro tipo | 2 | `argumentos` |
| no son dos municipios en un plazo de la LPAC, o uno en `contencioso` | 2 | `argumentos` |
| un municipio que `territorio resolver` daría con 2: el mensaje de `territorio` de cada uno, tras «municipio «<pedido>»: », separados por « \| » (con candidatos, los da todos como `territorio`) | 2 | `argumentos` |
| todo lo demás vale y algún municipio no está en la relación | 3 | `no-encontrado` |
| `data/festivos/`, `data/plazos.yaml` o `data/territorio/` del binario no cargan | 1 | `inesperado` |

Nunca 4, 5 ni 6.

## 7. `--describe` y su golden (FR-085)

`kitlegal plazos calcular <tipo> <inicio> <municipios> --describe` emite el esquema con `entrada` y `salida`; su forma
canónica es `internal/app/testdata/plazos/describe.json`, y la salida, la parte `plazos calcular` de
`schemas/plazo.json` («Salida del verbo calcular del applet plazos. Generado desde --describe con make schema-check; no
editar.»). Los enumerados de `tipo`, `festivos[].ambito`, `avisos[].codigo`, `cobertura[].ambito` y
`cobertura[].falta` van en las etiquetas de los tipos de salida, y un test exige que digan lo
mismo que los vocabularios del dominio, como `TestTerritorioResuelto`.

## 8. Uso, de fuera adentro

| Quién | Cuántas veces | Qué escribe | Qué recibe | Cuándo deja de darse cada señal |
|---|---|---|---|---|
| `legal-core`, paso 9 | una por plazo que pide la pregunta: casi siempre una; dos si pide la reposición y el contencioso | la orden o la llamada: menos de 100 bytes | 0,77 a 0,86 KB en un plazo de un mes (M1, M2). Cada festivo saltado, unos 90 bytes; cada aviso, unos 250; cada entrada de `cobertura`, unos 100 | Cada llamada calcula de nuevo con lo que lleva el binario. Un aviso solo con lo que lo causa en el tramo de esa llamada: es información de esa fecha, no un recordatorio. El de los locales de 2027 de la Comunidad de Madrid deja de darse con la release que lleve el BOCM de 2027; el de una comunidad sin festivos locales, cuando entren los de su boletín (backlog «territorios»); el de los nacionales y autonómicos, cuando el tramo no sale de los años cargados; el de una parte, cuando el tramo no pasa por ese día; el del art. 183, cuando no pasa por el 24 de diciembre al 6 de enero |
| el agente, al conectar | una vez por conexión | — | el esquema de `plazos_calcular`: 531 bytes su `entrada` sin las globales (M3) | — |

**Peor caso, con meses de uso o sin ellos.** `plazos` no guarda nada ni lee la caché ni el grafo: tras cientos de
normas y miles de bloques consultados, la llamada siguiente devuelve lo mismo que la primera. Lo que puede crecer es la
duración pedida, y está acotado: festivos solo de 2025 a 2027; como estimación, no como control, unos 12 nacionales y
autonómicos más 2 locales por municipio y año, 48 en total con dos municipios, unos 4,3 KB; avisos, uno
por ámbito y municipio y uno por festivo de una parte de los años cargados, unos 8 KB en una comunidad con un festivo
por isla; `cobertura`, como mucho cinco tramos por ámbito, 1,5 KB. Menos de 15 KB por llamada.
