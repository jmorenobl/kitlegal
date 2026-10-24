# Contrato: los controles de contenido leen la skill entera (FR-016 a FR-019)

Quién lo consume: quien ejecuta `make ci`; una línea por defecto, que nombra la skill y el fichero. Cuándo deja de
darse: cuando el texto se corrige. No crece con el uso: lee el texto de las skills del repositorio.

«La skill entera» es su `SKILL.md` y sus referencias a mano (`skills.ReferenciasAMano`, contracts/gates-de-skills-check.md
§3), cada una por separado y con su ruta. Las referencias generadas no se leen, salvo donde ya se leían (control 8).
Cada control se aplica en las mismas skills que hoy.

| # | Control (subtest) | Dónde | Skills | Qué lee y qué exige, desde H31 |
|---|---|---|---|---|
| 1 | Prosa sin expresiones prohibidas ni fechas AAAAMMDD con cifras (`prosa-de-la-skill`) | `internal/evals/conjunto_test.go`, `TestEvalsDelRepositorio` | `boe-legislacion` | `defectosDeLaProsa` sobre cada fichero; ninguno lleva una expresión fuera del código ni una fecha con cifras. La prosa no vacía se sigue exigiendo al `SKILL.md` |
| 2 | La línea sin consulta (`linea-sin-consulta`) | ídem | `boe-legislacion`, `legal-core` | La línea, sola en un bloque `text` del `SKILL.md`, como hoy: en una referencia sola, rojo |
| 3 | Las órdenes para PowerShell (`ordenes-para-powershell`) | ídem | `boe-legislacion` | Cada orden de lectura o de comprobación de cada fichero lleva detrás, en ese fichero, su forma para PowerShell; una orden de Bash por verbo de lectura (`articulo`, `articulos`) en la skill entera |
| 4 | Las formas de la redacción fechada (`formas-de-la-redaccion-fechada`) | ídem | `boe-legislacion` | Cada una de las tres en un bloque `text` del `SKILL.md`, como hoy |
| 5 | Las formas de aviso (`avisos-de-la-skill`) | ídem | `boe-legislacion` | `ComprobarFormasDeAviso` sobre el `SKILL.md`, como hoy |
| 6 | Las formas de hallazgo (`hallazgos-de-la-skill`) | ídem | `boe-legislacion` | `ComprobarFormasDeHallazgo` sobre el `SKILL.md`, como hoy |
| 7 | Las líneas `⚠` de las referencias (`lineas-de-aviso-de-las-referencias`, nuevo) | ídem | `boe-legislacion`, `legal-core` | Cada línea de una referencia a mano que empieza, tras blancos, un marcador de lista (`-`, `*`, `<n>.`) y énfasis (`*`, `_`, `` ` ``), por `⚠` es, sin los blancos de los extremos, una línea del `SKILL.md` |
| 8 | Normas nombradas (`normas-nombradas`) | `internal/app/skills_test.go`, `TestSkillsDelRepositorio` | todas | Cada «Ley N/AAAA», «LO…», «RD…» de cada fichero empieza un título de `data/normas.yaml` |
| 9 | Sin instrucciones de las evals (`sin-instrucciones-de-evals`) | ídem | todas | Como hoy, `SKILL.md` sin la región y todo `references/*.md`; su mutante, también sobre una referencia a mano |
| 10 | Las órdenes de ejemplo que leen del BOE (`ordenes-de-ejemplo`, nuevo) | `internal/evals/conjunto_test.go`, `TestEvalsDelRepositorio` | las que tienen evals | Cada orden concreta de cada fichero —`kitlegal boe`, un verbo y al menos un argumento, sin marcador— sale con 0 con `--offline` sobre la caché de una sesión de su skill |

## Detalle de los controles nuevos

### 7 · Las líneas `⚠` de las referencias (FR-018)

Una forma fija la exigen los controles 2, 4, 5 y 6 en el `SKILL.md`, y las evals la comprueban sin modelo en la
respuesta. En una referencia a mano solo puede repetirse como la escribe el cuerpo: así una variante mal escrita —una
errata, otra puntuación, otro marcador— no le enseña al agente una forma que la eval no reconoce. Ejemplo, con el
cuerpo de `boe-legislacion` que lleva `⚠ REDACCIÓN MODIFICADA: <cita>: la redacción con fecha de vigencia <fecha>, …`:

- en verde, ninguna línea `⚠` en la referencia, o esa misma línea;
- en rojo, `⚠ REDACCION MODIFICADA: <cita>: …` (sin tilde), con el defecto
  `boe-legislacion: references/redaccion_modificada.md:7: la línea ⚠ no es ninguna de SKILL.md: «⚠ REDACCION MODIFICADA: …»`.

### 10 · Las órdenes de ejemplo (FR-019)

- **Qué se extrae**: de cada fichero, en los tramos de código que ya recorre `ordenesDeLectura` (cada línea de un
  bloque delimitado y cada tramo de código en línea fuera de ellos), cada aparición de `kitlegal boe` con lo que la
  sigue en el tramo hasta el primer `&&`, `;` o `|`, sin `--json`. Es una orden concreta si lo que queda lleva un verbo
  y al menos un argumento, y ningún marcador `<…>`, `...` ni `…`; lo demás no se ejecuta. La mención del verbo sin
  argumentos no es una orden, como ya fija `TestOrdenesParaPowerShell` («El texto de kitlegal boe articulo en la
  prosa no es una orden»): «`kitlegal boe articulo` o `articulos`» de la regla 6 (B123), «con
  `kitlegal boe articulo`» de references/lecturas_que_fallan.md (B115) o «Si `kitlegal boe` falla» de la regla 2.
- **El extractor**: `ordenesDeEjemplo` (`internal/evals/conjunto_test.go`, junto a `ordenesDeLectura`), que fija
  `TestOrdenesDeEjemplo` sobre Markdown escrito en el test: la orden en un bloque, en su forma de Bash (hasta `&&`) y
  en la de PowerShell (hasta `;`), y en código en línea, cada una con su línea y su `--fecha <día>`; y ninguna orden de
  la que lleva un marcador, de la mención del verbo sin argumentos ni de la que está en la prosa.
- **Contra qué**: `cacheDeUnaSesion(t, evals/<skill>, <fichero>)`, la caché que prepara `PrepararSesion` para la primera
  eval de `evals/<skill>/`, por nombre, que no declara `sin_servir`; `ComprobarSinRed` ejecuta cada una con
  `--offline` y da una `Falta` por cada una que no sale con 0.
- **Lo que sirve hoy** (research V3, V4): `indice`, `metadatos` y `articulo` con y sin `--fecha` de las normas de
  `data/normas.yaml`; ningún `buscar` ni `analisis`.
- **El defecto**: `boe-legislacion: SKILL.md: «boe buscar régimen jurídico del sector público» terminó con código 4: …`
  (el texto de la `Falta`).
- **Sin pasar en vacío**: el subtest exige al menos una orden en el repositorio (hoy, las de `boe-legislacion`).
- **Invocaciones**: una por aparición de una orden concreta. Con el reparto de contracts/skills-reescritas.md, cuatro,
  todas de `boe-legislacion`: en el cuerpo, `boe indice BOE-A-2015-10565` y `boe articulo BOE-A-2015-10565 a21` en su
  forma de Bash y en la de PowerShell; en references/redaccion_en_una_fecha.md,
  `boe articulo BOE-A-1985-5392 a22 --fecha 2005-06-15` (cinco si la reescritura pone el ejemplo de `metadatos`, que
  también se sirve). Las de `buscar`, `articulos`, la tabla de comandos y las de `legal-core` llevan marcadores;
  `jurisprudencia` no tiene ninguna. Hoy, sobre el texto de `main`, son seis, y las de `buscar` y `analisis` (l. 48 y
  88) salen con 4 (research V3): son las dos que cambia la reescritura (D17).

## Mutantes (SC-005)

Un test de `make ci`, `TestControlesSobreLasReferenciasAMano` (`internal/evals/conjunto_test.go`), copia la skill a
`t.TempDir()` y, por cada control 1 a 7 y 10, mueve su regla del `SKILL.md` a una referencia a mano de la copia o la
escribe en ella, y exige el defecto: la expresión prohibida, la fecha AAAAMMDD, la orden sin su forma para PowerShell,
la línea sin consulta movida, cada forma de aviso, de hallazgo y de la redacción fechada movida, una línea `⚠` mal
escrita y la orden de ejemplo que la caché no sirve. Los de los controles 8 y 9 (la norma que no está en
`data/normas.yaml`, la instrucción de las evals) van en `TestSkillsDelRepositorio`, sobre `copiaConLaSkill`. Con
cualquiera sin detectar, el test sale en rojo.
