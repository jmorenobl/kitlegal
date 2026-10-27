# Data model: H31.1

El hito no añade tipos ni estructuras de datos al binario ni al arnés. Las entidades son ficheros del repositorio y
sus partes, con las reglas que las comprueban.

## 1. El `SKILL.md` de `boe-legislacion` v0.3.1

| Parte | Qué es | Regla | Lo comprueba |
|---|---|---|---|
| Frontmatter | `name`, `description`, `metadata` (`kitlegal-applets: boe graph`, `kitlegal-referencias: normas`), líneas 1 a 15 | igual al de `main` y al de la skill medida (FR-010) | `h31-1-sitio-en-el-cuerpo`; `TestSkillsDelRepositorio` |
| Cuerpo | desde la línea 16: lo que el agente carga entero al activar la skill | ≤ 11 000 runas y < 500 líneas: 10 996 y 156 (FR-005) | gate de H31 (`defectosDelCuerpo`) |
| Pasaje medido | paso 3, primera viñeta del paso 4, paso 5, paso 6 | idéntico, línea a línea, a la skill medida (FR-002) | `h31-1-responder`, `h31-1-lee-lo-que-nombra` |
| Forma fija | cada línea de un bloque `text` y las tres etiquetas de aviso | en el `SKILL.md` tal cual (FR-003) | `h31-1-sitio-en-el-cuerpo`; controles de §1 de `docs/SKILLS.md` |
| Tabla de comandos | entre `<!-- inicio de la tabla de comandos…-->` y `<!-- fin de la tabla de comandos -->` | la generada, igual a la de la skill medida (FR-004) | `h31-1-sitio-en-el-cuerpo`; `make skills-check` |
| Encabezado | cada línea `## ` y `### ` (14) | mismo texto y orden (FR-006): las referencias remiten a ellos | `h31-1-sitio-en-el-cuerpo` |
| Regla | `<n>. **<encabezado>** <texto>`, siete | mismo número y encabezado (FR-006); el texto, reescrito o no (contrato §2) | `h31-1-sitio-en-el-cuerpo`; revisión final |
| Texto reescrito | T1 a T15 de `contracts/skill-boe-legislacion-v0-3-1.md` §2 | lo quitado lo dice otro sitio del cuerpo o el agente no lo haría mal (FR-006) | revisión final |

Un texto es **de una clase de pregunta** si solo se aplica cuando se da una condición de la pregunta o de una lectura;
lo demás es **lo que toda respuesta necesita** (spec, «Contexto»). En v0.3.1 ningún texto de una clase sale a una
referencia (contrato §3).

## 2. Una referencia de `references/`

`references/<nombre>.md`, plano. **A mano** (sin la cabecera «generado desde»): `buscar_la_norma`, `como_se_cita`,
`con_herramientas`, `lecturas_que_fallan`, `redaccion_en_una_fecha`, `redaccion_modificada`, `varios_bloques`.
**Generada** desde `data/normas.yaml` por `make skills-sync`: `normas`. Regla de v0.3.1: cada fichero es igual al de
la skill medida y no hay ninguno más (FR-001, FR-007); cada uno, enlazado desde el `SKILL.md` (gate de H31).

## 3. Una lectura de una eval (H26.1)

`{norma, bloque, motivo}` en la lista `lecturas` de `evals/boe-legislacion/<nn>-….yaml`. Un bloque que la caché de
las sesiones sirve porque la eval lo provoca; no es comando ni cita, ni cuenta en ningún umbral. Reglas: el bloque es
un id del índice grabado de su norma; la unión de grabaciones lo sirve sin red, sin fecha y con `--fecha 9999-12-31`;
el `motivo` dice por qué (FR-011). Se añaden nueve (contrato de lecturas §1).

## 4. Una entrada del manifiesto de grabaciones

`{busqueda, titulo_empieza_por, bloques, para}` en `internal/evals/testdata/grabaciones.json`, con
`"fuente": "boe.legislacion-consolidada"`. `TestGrabarEvals` resuelve la norma con la búsqueda y graba lo que falta
de sus metadatos, su índice y cada bloque. Se amplían dos entradas (contrato de lecturas §2).

## 5. Una grabación

`internal/evals/testdata/boe.legislacion-consolidada/GET_<url con _>.json`, una por petición GET, que escribe
`internal/httpx` en `grabar_datos` y que la caché de las sesiones y los tests leen sin red. Ocho nuevas.

## 6. Un guion de la suite de aceptación

`<feature_dir>/aceptacion/h31-1-<nombre>.txtar`, tres (`contracts/aceptacion.md`). Su ciclo es el del workflow
(`scripts/workflow/aceptacion.sh`): T001 lo escribe → `rojo-primero` exige que falle con el código de hoy por una
aserción → `congelar` guarda su huella y el guardián de diff lo protege → `activar` lo copia a
`internal/app/testdata/script/` al terminar el bucle → `TestEntregaDelHito` lo ejecuta en `make ci`.
