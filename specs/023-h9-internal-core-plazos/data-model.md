# Data model: H9

Las entidades del spec («Key Entities») con sus campos, reglas y relaciones. Los nombres Go son orientativos para la
tarea; los de las claves JSON y YAML los fijan los contratos.

## 1. Dominio: `internal/core/plazos`

### Calendario

Lo que cargan `data/festivos/` (contracts/datos-de-plazos.md §1). Inmutable tras cargar; lo usan a la vez varias
invocaciones.

| Campo | Qué es |
|---|---|
| nacionales | por año cargado (2025 a 2027): la disposición y las filas `{fecha, ambito, comunidades, parte?, source}` |
| locales | por comunidad con fichero (hoy `13`): por año cargado, sus disposiciones, los festivos por municipio `{fecha, source}` y las faltas por municipio `{motivo, source}` |

`CargarCalendario` recibe los bytes de los dos ficheros y los lee; lo que no se puede leer es un defecto `inesperado`
(código 1), la regla genérica, sin regla de carga propia: lo que llevan lo garantiza la derivación, que el test de
FR-083 compara byte a byte con lo grabado.

Consultas: `inhabil(día, municipios, cómputo)`; `festivosDe(día, municipios)` con su ámbito, municipio y `source`;
`cobertura(años, municipios)`; `partesDe(día, comunidad)`.

### Tipo

De `data/plazos.yaml` (`LeerTipos`): `codigo`, `nombre`, `computo` (`lpac`, `contencioso`), `duracion` opcional
(`cantidad` ≥ 1, `unidad` ∈ {`meses`, `dias-habiles`, `dias-naturales`}), `preceptos` (`norma`, `bloque`, `apartado`,
`fija`) y, en `generico`, `unidades` (unidad → precepto). Siete y solo siete.

### Petición

| Campo | Regla (FR-012; orden de comprobación) |
|---|---|
| tipo | uno de los siete |
| inicio | `AAAA-MM-DD`, día del calendario |
| duración, unidad | las dos en `silencio` y `generico`; ninguna en los demás; duración entero ≥ 1 |
| municipios | 2 en `lpac`, 1 en `contencioso`; el dominio comprueba cuántos, y el applet resuelve cada uno con `territorio.Registro.Resolver` y se lo da como `{codigo_ine, nombre, comunidad}` |

### Plazo (el `data` de `calcular`)

`fecha`, `tipo`, `preceptos`, `municipios`, `festivos`, `avisos`, `cobertura`: lo que enumera FR-001
(contracts/applet-plazos.md §4).

- **Festivo saltado**: `fecha`, `ambito` (`nacional`, `autonomico`, `local`), `municipio` solo si es local, `source`.
- **Aviso**: `codigo` (`faltan-festivos-nacionales-y-autonomicos`, `faltan-festivos-locales`, `dias-del-art-183-lopj`,
  `festivo-de-una-parte`) y `texto`, una línea con su forma fija.
- **Cobertura**: `ambito` (`nacional-y-autonomico`, `local`), `municipio` si es local, `desde` y `hasta` (años),
  `disposicion` si hay datos, `falta` si no hay festivos (`sin-cargar`, `sin-festivos-locales`, `no-comunicado`,
  `linea-sin-leer`, `sin-linea`).

Relaciones: los municipios vienen de `territorio` (código INE y comunidad); los festivos, del calendario; los
preceptos, del tipo. Nada se guarda: cada llamada calcula.

### Motivo de falta (FR-022)

Vocabulario cerrado del fichero local: `no-comunicado`, `linea-sin-leer`, `sin-linea` (de un municipio) y
`sin-municipio` (de una línea). `cobertura.falta` añade los dos que no salen de un fichero: `sin-cargar` (año sin
datos) y `sin-festivos-locales` (comunidad sin fichero local).

## 2. Derivación: `internal/source/boe/fiestas` e `internal/source/bocm`

### Disposición

`{id, anio, url}`, del manifiesto; ordenadas por fecha de publicación (la del id). Cuerpo, de `httpx.Replay` sobre
`evidencias/h9/`.

### Línea del BOCM

`{nombre, cuerpo, disposición}` → uno de: `festivos` (lista de fechas), `no-comunicado`, `linea-sin-leer`. Cruce →
municipio (código INE) o `sin-municipio`.

**Transiciones por año** (FR-021): cada municipio empieza sin línea; cada disposición, en su orden, sustituye la
entrada de los municipios que nombra; al final, sin línea → `sin-linea`.

| Entrada antes | Línea de la disposición | Entrada después |
|---|---|---|
| cualquiera | fechas | festivos, con el `source` de la disposición |
| cualquiera | «no comunicado» | `no-comunicado`, con su `source` |
| cualquiera | que no se lee | `linea-sin-leer`, con su `source` |
| cualquiera | ninguna | la misma |

### Fila del BOE

`{fecha, ambito, comunidades, parte?, source}` por cada fila del anexo y marca (research D14).

## 3. Evals: `internal/evals`

### Eval (claves nuevas)

| Campo | YAML | Regla |
|---|---|---|
| `Plazo.Fecha` | `plazo.fecha` | `AAAA-MM-DD` |
| `Plazo.Avisos` | `plazo.avisos` | códigos de `plazos.CodigosDeAviso`, sin repetir |
| `Plazo.Razonamiento` | `plazo.razonamiento` | no vacío |
| `NoCubierto.Norma` | `no_cubierto.norma` | `BOE-A-…`; nunca con `plazo` |
| `ComandoEsperado.Tipo` | `tipo` del comando `calcular` | no vacío |

Una eval de plazo o de cobertura es activa y no es sin binario ni servidor.

### Juicio

`ResultadoDeEval` gana, publicadas en el informe por sesión: `plazo_encontrado` (la línea con la fecha esperada),
`avisos_del_plazo_ausentes`, `no_cubierto_declarado` (la declaración con la norma esperada) y `lineas_del_plazo`
(fechas de las líneas `⚠ PLAZO:` de la respuesta). Cuentan en `cumpleLoEsperado`.

### Respuesta medida

`respuestaDelJuez` gana `dePlazo`, `sinFecha`, `deCobertura` y `sinDeclarar`, que pone `medirLosPlazos` en las
respuestas de los grupos de los modos con la eval y la respuesta de su sesión. `grupoJuzgado` gana
`deLasEvalsDePlazo`, `sinLaFecha`, `deLasEvalsDeCobertura` y `sinLaDeclaracion`.

### Umbral

`plazo_sin_fecha:<modelo>:<modo>` y `cobertura_sin_declarar:<modelo>:<modo>`, con el contrato del ADR 0029
(contracts/evals-legal-core.md §3).
