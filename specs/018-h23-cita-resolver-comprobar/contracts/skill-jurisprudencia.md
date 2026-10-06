# Contrato · La skill `jurisprudencia` v0

El texto propuesto de `skills/jurisprudencia/SKILL.md` está entero en
[skill-jurisprudencia.prototipo.md](./skill-jurisprudencia.prototipo.md): 197 líneas con la tabla de comandos que
generará `make skills-sync` (medido con `wc -l`; el límite es 299). La tarea que entrega la skill parte de él. No se ha
probado con ningún modelo: lo miden las evals del cierre.

## 1. Qué hay en `skills/jurisprudencia/`

- `SKILL.md`, con `name: jurisprudencia`, su `description` y `metadata.kitlegal-applets: cita`. Sin
  `kitlegal-referencias` y sin carpeta `references/`: el hito no le da ningún dato de `data/` (spec, «Assumptions»).
  Sin `scripts/`.
- Genérica: ni vertical, ni caso para un territorio o un órgano (FR-060). `boe-legislacion` y `legal-core` no cambian
  (FR-069).

## 2. Cada parte del texto y su requisito

| Parte del prototipo | Requisito |
|---|---|
| «Protocolo», las dos formas de pedir | ADR 0035; FR-060 |
| 1 · Qué hay que resolver; de una en una | FR-061 |
| 2 · Con qué se resuelve cada referencia; el segundo intento con la fecha; sin fecha no se consulta | FR-061, FR-063 |
| 2 · «Su ECLI y su ROJ: por el ECLI» | FR-061 (precisión del plan, research D29): hace determinista la eval (d) |
| 3 · Leer el resultado: una, más de una, página completa, cada código | FR-062, FR-063, FR-067, FR-068; FR-014 |
| 4 · Lo que no se ha leído no se resume | FR-064 (sin control en este hito: FR-085) |
| 5 · Una pregunta por materia | FR-065 |
| 6 · Una sentencia del Tribunal Constitucional | FR-066 |
| «Cómo se cita» | FR-062 |
| «Lo que no se ha podido comprobar» | FR-063 |
| «Comandos» y su tabla | FR-060; ADR 0035 |
| «Reglas» | FR-063, FR-064, FR-068 |

## 3. Lo que el prototipo no dice, y por qué

- **Operadores del buscador**: ninguno. El repositorio no describe ninguno (spec, «Assumptions»). La tarea que escribe
  `SKILL.md` tiene delante la grabación de la página del buscador; si esa página documenta operadores o los valores de
  sus campos de jurisdicción y de órgano, el paso 5 los nombra tal como constan en ella, y nada que no conste.
- **La dirección del buscador del Tribunal Constitucional** es `https://hj.tribunalconstitucional.es/`, escrita aquí y
  en ningún otro sitio del producto; no se ha comprobado (research S6). La eval (e) la toma de aquí y un test lo exige
  ([evals-jurisprudencia.md §2](./evals-jurisprudencia.md)).
- **Nada de las evals, del job, de modelos ni de la caché de las sesiones**: una skill no nombra cómo se mide.

## 4. Controles en `make ci`

| Control | Qué falla |
|---|---|
| `make skills-check` (`TestSkillsDelRepositorio`) | 300 líneas o más; frontmatter inválido; un applet declarado que no está en el registro; la tabla distinta de la generada (FR-060; SC-009) |
| `TestOrdenesDeLasSkillsEmpotradas` | una orden de la tabla que no está en el registro o una herramienta que el servidor no anuncia |
| `TestTablaDeComandosCoincideConLaGramatica` | una fila cuya sintaxis rechaza la gramática del verbo; con una bandera propia, que la fila no la escriba como bandera |
| `TestEvalsDelRepositorio`, subpruebas `formas-de-jurisprudencia` y `direcciones-de-la-skill` | `SKILL.md` sin la forma de la cita o sin la línea `⚠ SENTENCIA NO COMPROBADA:` tal como las extrae el job; una dirección de una eval que no está en `SKILL.md` |

`skillsExigidas` gana `jurisprudencia`: sin ella en `skills/`, esos controles no pasan en vacío.

Dos controles de las otras skills **no** miran esta, y lo que cambia para que sea así está en research D33:

- **La forma para PowerShell.** `TestOrdenesParaPowerShell` y la subprueba `ordenes-para-powershell` solo leen el
  `SKILL.md` de `boe-legislacion` y la pareja `kitlegal boe <verbo> … && kitlegal graph check …`. No cubren ninguna
  orden de `jurisprudencia`, que no encadena órdenes —el prototipo no lleva ningún `&&`—, y el spec no pide ese control
  para esta skill.
- **La línea `⚠ SIN CONSULTA AL BOE: …`.** La subprueba `linea-sin-consulta` la exige a las skills cuyo conjunto de
  evals lleva la eval sin binario ni servidor. El de `jurisprudencia` no la lleva: sin herramienta ni binario, la skill
  responde con su línea `⚠ SENTENCIA NO COMPROBADA:` y el motivo «no se ha podido consultar» (FR-063), y el prototipo
  no lleva la del BOE.

Y la skill es la primera sin referencias (§1): dos ayudantes de `TestSkillsDelRepositorio` que daban por hecho que toda
skill las declara cambian para que sus casos de frontmatter y de regeneración se ejecuten también sobre ella.

## 5. Uso, de fuera adentro

| Qué | Quién lo recibe | Cuánto | Cuándo deja de darse |
|---|---|---|---|
| `SKILL.md` | el modelo, una vez por conversación en que se activa | 197 líneas | — |
| Consultas a `cita resolver` | la skill las pide de una en una | ECLI o ROJ dados como tales: 1 por sentencia. Cita con número y fecha: 1 si se encuentra, 2 si no. Sin fecha, por materia o del Tribunal Constitucional: 0 | cada una responde a su referencia |
| La cita | quien pregunta | una línea por sentencia comprobada: 71 caracteres más la dirección | no es una señal |
| La línea `⚠ SENTENCIA NO COMPROBADA:` | quien pregunta | una por referencia no comprobada de esa respuesta, del orden de 150 caracteres | no se repite en otra respuesta salvo que se pregunte otra vez por esa referencia y siga sin comprobarse |
| La declaración de que hay más de una resolución o una página completa | quien pregunta | una frase | solo con esa entrega |
