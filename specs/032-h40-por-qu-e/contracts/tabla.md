# Contrato: la tabla del sondeo de las variantes de `normativa` y la regla testigo

FR-020 a FR-023, FR-030; SC-004, SC-005, SC-007. Research D9 a D11; data-model §1, §5, §6. Lo que no se nombra aquí es
el guion de H39 tal cual (`specs/031-h39-c-omo-se/contracts/tabla.md`).

## 1. Cómo se lanza y de qué familia es

```sh
make laboratorio-tabla DIRECTORIO=<ruta>     # go run ./cmd/laboratorio tabla <ruta>
```

Sin red, modelo ni credencial, como en H39: `<ruta>` es un sondeo si tiene `sondeo.json` y, si no, se leen sus
subdirectorios que lo tienen, en orden de nombre. **La familia de la tabla es la de la primera variante que lee**
(sondeos y variantes en orden de nombre): si es una de las cuatro de `normativa`, la tabla es la de §3 y §4; si no, la
de H39, con su comportamiento de hoy, byte a byte (FR-023). En una tabla de `normativa`, un directorio de variante que
no es de `normativa` —también uno de H39— es un defecto: `laboratorio: <ruta> no es ninguna variante de normativa` en la
salida de error, código 1 y sin tabla; y un modo que no es de la variante, el defecto de H39 (`… no es ningún modo de la
variante …`). Lo demás que el guion no sabe leer, como en H39: defecto `inesperado`, código 1, sin tabla (ADR 0023).

## 2. Qué cuenta, por sesión

Las reglas de H39 (contracts/tabla.md §2 de H39; FR-020), con la skill y las referencias de la familia:

- **Turno**, **primera herramienta** (de cada modo) y **terminada**: las de H39, sin cambios. Una sesión que no termina,
  o que termina sin respuesta, va a «Sin terminar» y no entra en ningún `n` (FR-021).
- **Activa antes**: un `tool_use` de `Skill` con `input.skill` igual a `normativa` en un turno anterior al de la primera
  herramienta de kitlegal, o activa y sin ninguna herramienta; en el mismo turno, no.
- **Abre la referencia `<f>`**, por cada una de las ocho de la familia: un `tool_use` de `Read` cuyo `input.file_path` es
  `normativa/references/<f>` o termina en `/normativa/references/<f>`, entero o por tramos, o de `Bash` cuya
  `input.command` contiene `references/<f>`. `Grep` y `Glob` no la abren. Su turno es el de la primera apertura; abierta
  varias veces, cuenta una.
- **Cumple**: `reglaTestigo(Respuesta)` (§6). Se cuenta en todas, y solo se escribe en `normativa-testigo-en-referencia`.

## 3. La tabla

Una fila por variante, modo, modelo y versión de Claude Code, en el orden de la declaración de la familia, `orden` antes
que `herramienta`, y modelo y versión por orden de texto. 16 columnas: las seis de H39; «Activan antes»; una por cada
referencia, en orden de nombre de fichero; y «Cumplen la regla».

- Recuento: `<k> de <n> · <p> % [<lo>–<hi> %]`, con un decimal, coma decimal y el intervalo de Wilson al 95 %
  (z = 1,96) recortado a [0, 100] antes de redondear, como en H39.
- Celda de una referencia: su recuento y, si alguna sesión la abre, ` · turno <mediana> (<mínimo>–<máximo>)` del turno
  de la primera apertura de las que la abren, con la mediana entera o con «,5»; si ninguna la abre, solo el recuento
  (`0 de <n> · …`).
- «Cumplen la regla»: el recuento en `normativa-testigo-en-referencia`; vacía en las otras tres, no a cero.
- Con `n` = 0: vacías todas salvo variante, modo, modelo, versión, «Sin terminar» y `n`.

Ejemplo, con las sesiones fijas del guion de aceptación (contracts/arnes-e2e.md §2; Wilson en research V21):

```text
# Sondeo de las variantes de normativa

| Variante | Modo | Modelo | Claude Code | Sin terminar | n | Activan antes | buscar_la_norma.md | como_se_cita.md | con_herramientas.md | lecturas_que_fallan.md | normas.md | redaccion_en_una_fecha.md | redaccion_modificada.md | varios_bloques.md | Cumplen la regla |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| normativa-control | orden | claude-sonnet-5-5 | 2.1.317 | 0 | 1 | 1 de 1 · 100,0 % [20,7–100,0 %] | 0 de 1 · 0,0 % [0,0–79,3 %] | 0 de 1 · 0,0 % [0,0–79,3 %] | 0 de 1 · 0,0 % [0,0–79,3 %] | 0 de 1 · 0,0 % [0,0–79,3 %] | 1 de 1 · 100,0 % [20,7–100,0 %] · turno 2 (2–2) | 0 de 1 · 0,0 % [0,0–79,3 %] | 0 de 1 · 0,0 % [0,0–79,3 %] | 0 de 1 · 0,0 % [0,0–79,3 %] |  |
| normativa-control | herramienta | claude-sonnet-5-5 | 2.1.317 | 0 | 1 | 0 de 1 · 0,0 % [0,0–79,3 %] | 0 de 1 · 0,0 % [0,0–79,3 %] | 0 de 1 · 0,0 % [0,0–79,3 %] | 0 de 1 · 0,0 % [0,0–79,3 %] | 0 de 1 · 0,0 % [0,0–79,3 %] | 0 de 1 · 0,0 % [0,0–79,3 %] | 0 de 1 · 0,0 % [0,0–79,3 %] | 0 de 1 · 0,0 % [0,0–79,3 %] | 0 de 1 · 0,0 % [0,0–79,3 %] |  |
| normativa-cuerpo-corto | herramienta | claude-sonnet-5-5 | 2.1.317 | 1 | 0 |  |  |  |  |  |  |  |  |  |  |
| normativa-testigo-en-referencia | herramienta | claude-sonnet-5-5 | 2.1.317 | 0 | 2 | 2 de 2 · 100,0 % [34,2–100,0 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | 1 de 2 · 50,0 % [9,5–90,5 %] · turno 2 (2–2) | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | 0 de 2 · 0,0 % [0,0–65,8 %] | 1 de 2 · 50,0 % [9,5–90,5 %] |
```

## 4. El detalle

Detrás, una línea vacía, `## Detalle`, otra línea vacía y una tabla con una fila por sesión, en el orden de las filas y,
dentro, por ruta, relativa a `<ruta>`:

```text
| Sesión | Terminada | Activa antes | buscar_la_norma.md | como_se_cita.md | con_herramientas.md | lecturas_que_fallan.md | normas.md | redaccion_en_una_fecha.md | redaccion_modificada.md | varios_bloques.md | Cumple |
|---|---|---|---|---|---|---|---|---|---|---|---|
| h40-a/normativa-control/orden/04-lgt-prescripcion-01 | sí | sí | no | no | no | no | turno 2 | no | no | no |  |
| h40-a/normativa-control/herramienta/05-trlrhl-impuestos-municipales-01 | sí | no | no | no | no | no | no | no | no | no |  |
| h40-a/normativa-cuerpo-corto/herramienta/15-irpf-rendimientos-por-materia-01 | no: tope de 240 s agotado (código 124) |  |  |  |  |  |  |  |  |  |  |
| h40-a/normativa-testigo-en-referencia/herramienta/04-lgt-prescripcion-01 | sí | sí | no | no | turno 2 | no | no | no | no | no | sí |
| h40-a/normativa-testigo-en-referencia/herramienta/05-trlrhl-impuestos-municipales-01 | sí | sí | no | no | no | no | no | no | no | no | no |
```

Por referencia, `no` o `turno <t>`; «Cumple», `sí` o `no` solo en `normativa-testigo-en-referencia`; una sesión sin
terminar, su motivo y las demás celdas vacías. Título, líneas vacías y separadores, como en H39.

## 5. Uso, de fuera adentro

La lee la persona una vez, para escribir el ADR (FR-041): en la aceptación, cinco filas de 400 a 600 bytes —las tres
completas del ejemplo de §3 suman 1212, y con `n` = 40 cada recuento gana unos 7 bytes y cada turno 17— más la cabecera
de 269 (unos 3 KB), y el detalle de las 200 sesiones, a unos 140 bytes cada una (las cinco del ejemplo de §4 suman 685),
unos 28 KB. Su tamaño lo fijan las
variantes, los modos y las sesiones pedidas: como mucho cinco filas por modelo y versión en una familia de `normativa`, y
no cambia con cientos de normas y miles de bloques consultados ni con nada que se acumule. No es una señal: no avisa de
nada, deja de leerse cuando el ADR está escrito, y un sondeo nuevo da otra tabla. Ninguna skill del producto la pide.

## 6. La regla testigo (FR-030)

El guion es `reglaTestigo` de H39, sin cambios: cumple la respuesta cuya última línea no vacía, sin espacios ni `\r` a los
lados, es exactamente `Sello de lectura: KL-7Q4`. Respuestas fijas a las evals del sondeo, en
`internal/evals/testdata/laboratorio/normativa/respuestas/`, con el veredicto en el nombre como en H39:

| Fichero | Eval a la que responde | Última línea no vacía |
|---|---|---|
| `cumple.txt` | 04 (art. 66 LGT) | `Sello de lectura: KL-7Q4` |
| `no-cumple-sin-sello.txt` | 05 (art. 59 TRLRHL) | otra línea; el sello no aparece |
| `no-cumple-sello-en-negrita.txt` | 15 (art. 17 LIRPF) | `**Sello de lectura: KL-7Q4**` |
| `no-cumple-sello-entre-comillas.txt` | 22 (art. 118 LCSP en 2019) | `«Sello de lectura: KL-7Q4»` |
| `no-cumple-sello-en-codigo.txt` | 04 (art. 66 LGT) | `` `Sello de lectura: KL-7Q4` `` |
| `no-cumple-sello-no-al-final.txt` | 05 (art. 59 TRLRHL) | otra línea; el sello va en una línea anterior |

Y la respuesta de la skill del producto a la eval 05: la de la sesión `05-trlrhl-impuestos-municipales-claude-sonnet-5-5-01`
del informe versionado de H32 (`specs/030-h32-arquitectura-de-skills/gates/evals/normativa.json`), leída del informe,
que no cumple (research V18).

## 7. Controles

- **`TestTablaDeNormativa`** (`internal/evals/laboratorio_tabla_test.go`, en `make ci`; FR-020 a FR-022, SC-005): sobre
  las sesiones fijas de `internal/evals/testdata/laboratorio/normativa/sondeo/` —los dos modos; una referencia abierta dos
  veces, con `Read` por tramos, y otra con `Bash`; `Grep` sobre una referencia; referencias que ninguna sesión abre;
  `Skill` y la primera herramienta en el mismo turno; una sesión sin ninguna herramienta que activa; una sesión sin
  terminar; una fila sin ninguna sesión terminada; una sesión que cumple la regla testigo y otra que no— la salida es,
  byte a byte, `internal/evals/testdata/laboratorio/normativa/tabla.md`. Y sobre sondeos que el test monta en un temporal
  con 40 copias de sesiones fijas, las celdas de 0, 20 y 40 de 40: `[0,0–8,8 %]`, `[35,2–64,8 %]` y `[91,2–100,0 %]`.
  Con un directorio de una variante de H39 en un sondeo de `normativa`, el defecto de §1, código 1 y sin tabla: una tabla
  tiene las columnas de una sola familia (FR-010).
- **`TestReglaTestigoDeNormativa`** (`internal/evals/laboratorio_normativa_test.go`, en `make ci`; FR-030, SC-004): el
  directorio de §6 tiene exactamente esas seis respuestas, cada una con el veredicto de su nombre, y la respuesta de H32,
  el de no cumplir.
- **`TestTablaDelLaboratorio`** y **`TestReglaTestigo`** (H39): sin cambiar ningún valor esperado (FR-023).
- **El guion congelado** `h40-laboratorio-tabla.txtar` (contracts/arnes-e2e.md §2; FR-020, FR-021, FR-023, SC-005,
  SC-007).
