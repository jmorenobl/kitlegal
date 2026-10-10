# Data model: H25 · El juez de `jurisprudencia`

Lo que el hito añade o cambia sobre el modelo de H24 (`specs/017-h24-las-evals-juzgan/data-model.md`) y de H23. Lo que
no aparece aquí sigue como está. Ningún esquema de `schemas/` cambia.

## 1. La carpeta del juez de `jurisprudencia`

`evals/jurisprudencia/juez/`, cinco ficheros (FR-001, FR-002). Los lee `LeerConjunto` en un `Juez`, sin cambios.

| Fichero | De dónde sale | Regla |
|---|---|---|
| `rubrica.md`, `esquema.json`, `casos.yaml`, `medida.json` | Copia byte a byte de los de la evidencia de la validación del juez de esta skill | Idénticos a su original: `TestCopiasDelJuez` |
| `clases.yaml` | Se escribe en la carpeta | Contra `schemas/juez-clases.yaml.json`; sus nombres son las propiedades de `esquema.json` |

Clases: `afirma_lo_no_leido` (`decide: true`, `umbral: 0`) y `afirma_que_existe` (`decide: false`, `umbral: 0`).

## 2. El voto

`VotoDeClase` gana un campo, y lo que se lee de cada clase de un voto, el mismo (FR-013; research D2).

| Campo | Clave JSON | Cuándo está |
|---|---|---|
| `Precepto *string` | `precepto` | En la clase cuyo esquema lo tiene: hoy, `afirma_lo_no_leido` de `boe-legislacion` |
| `Sentencia *string` | `sentencia` | En la clase cuyo esquema lo tiene: `afirma_lo_no_leido` de `jurisprudencia` |

Los dos son `nil`, y sin clave en `informe.json`, donde el esquema no los tiene. El de un voto que dice no es vacío, y
su clave va igual. Ningún voto lleva los dos.

## 3. El caso etiquetado

`CasoEtiquetado` no cambia de campos. Cambia el tipo de `quitado`, que pasa a llamarse `Quitado` (research D3):

| Campo | Clave | Qué quita | De qué skill |
|---|---|---|---|
| `Norma`, `Bloque` | `norma`, `bloque` | El texto de ese bloque, de los textos de la sesión | `boe-legislacion` |
| `Texto` | `texto` | Una parte del texto pegado, de la pregunta: `documento`, `fallo` o `apartado-2` | `jurisprudencia` |

Un derivado lleva una de las dos formas. Su nombre en un error es `<informe> <sesión> sin <norma> <bloque>` o
`<informe> <sesión> sin <texto>`. El fichero lleva además `regla`, que no se lee.

`informe` nombra uno de dos tipos de informe, que se distinguen por su contenido (research D8):

| Tipo | Se reconoce por | Sesiones | De cada sesión |
|---|---|---|---|
| Del job de evals | No lleva `sondeo` | `evals[]` | `sesion`, `eval`, `respuesta`, `invocaciones[]` con `orden`, `llamada` y `codigo` |
| De un sondeo | Lleva `sondeo` | `sesiones[]` | `sesion`, `pregunta` (un id de `preguntas.json`), `respuesta`, `invocaciones[]` con `herramienta`, `entrada.command`, `salida` y `error` |

## 4. Las preguntas de un sondeo

`preguntas.json`, en la carpeta del informe del sondeo. Se lee, no se escribe.

| Clave | Qué es |
|---|---|
| `fragmento` | La ruta del fragmento, desde la raíz del repositorio |
| `preguntas[].id` | El id que nombra cada sesión del sondeo |
| `preguntas[].eval` | El fichero de la eval de la skill cuya pregunta es la de ese id |
| `preguntas[].pregunta` | Si no lleva `eval`, la plantilla: `{fragmento}` es el fragmento, byte a byte, y `{ficha}`, el fragmento hasta su primera línea en blanco, con un salto de línea al final |

## 5. El texto pegado y lo que se quita

De una pregunta (contracts/medida-y-casos.md §3):

- **Texto pegado**: si la pregunta lleva una línea en blanco seguida de `Roj:`, lo que sigue a su primera línea en
  blanco; si no, ninguno. Lo de antes es la **entrada**.
- **Ficha**: el texto pegado hasta su primera línea en blanco.
- **Lo que queda** tras quitar: con `documento`, nada; con `fallo`, lo anterior a la línea `F A L L O`, sin sus saltos
  de línea finales y con uno; con `apartado-2`, el texto sin el párrafo que empieza por `2.º-` ni la línea en blanco
  que lo sigue.
- **Pregunta del derivado**: la entrada y, si queda algo, una línea en blanco y lo que queda.

## 6. La sesión reconstruida

Se recuerda por informe, sesión y texto quitado (research D10): la pregunta, la respuesta y los textos.

| Origen | Pregunta | Respuesta | Textos |
|---|---|---|---|
| Informe del job, órdenes `boe` y `graph` | La de su eval | La del informe | Como en H24 |
| Informe del job, órdenes `cita` | La de su eval, sin lo quitado | La del informe | Cada orden `cita`, repetida en proceso: su orden tal cual y su salida estándar. Sin texto pegado, ninguna `cotejar`. Un código distinto del del informe, error |
| Informe de un sondeo | La de `preguntas.json`, compuesta y sin lo quitado | La del informe | De cada invocación de `Bash` cuya orden empieza por `kitlegal `, la orden y su `salida`. Sin texto pegado, ninguna con `cita cotejar` |

## 7. Los umbrales de `jurisprudencia`

Doce elementos, diez con `decide: true`, con el contrato del ADR 0029 (contracts/juez-de-jurisprudencia.md §5). No hay
tipo nuevo: son `Umbral`.

## 8. La definición del job

`DefinicionDelJob` y `TrabajoDeLaMedida` no cambian de campos. Cambian los valores que el repositorio fija y los que
`TestDefinicionDelJob` exige (contracts/job-de-evals.md):

| Qué | Hoy | Con H25 |
|---|---|---|
| `jobs.evals…include`, `jurisprudencia` | `concurrencia: 1`, `objetivo_de_duracion: 0` | `concurrencia: 4`, `objetivo_de_duracion: 0` |
| `jobs.medida.strategy.matrix.skill` | `[boe-legislacion]` | `[boe-legislacion, jurisprudencia]` |
| `jobs.medida…include`, `jurisprudencia` | no está | `concurrencia: 4` |
| Los dos `timeout-minutes` | 352 y 269 | los mismos |

## 9. Las evals

Diez ficheros, con el formato de eval de hoy (contracts/evals-jurisprudencia.md). Las reglas del conjunto: tamaño 10;
todas activan y declaran `sentencias`; por clase, 3, 2, 3, 1 y 1.

## 10. Ficheros

| Fichero | Cambio |
|---|---|
| `evals/jurisprudencia/juez/` | Nueva: cinco ficheros |
| `evals/jurisprudencia/07-…yaml` a `10-…yaml` | Nuevos |
| `skills/jurisprudencia/SKILL.md` | Dos pasajes |
| `.github/workflows/evals.yml` | La concurrencia, la matriz de `medida`, la descripción de la entrada y sus comentarios |
| `internal/evals/juez.go`, `informe.go` | El campo `sentencia` y su columna |
| `internal/evals/medida.go` | `Quitado`, el informe de un sondeo, las órdenes `cita`, el texto pegado |
| `internal/evals/conjunto.go` | Las reglas del conjunto con las diez |
| `internal/evals/doc.go` | Lo que dice de `jurisprudencia` |
| `internal/evals/*_test.go` | Los de «Controles mecánicos» del plan |
| `CHANGELOG.md`, `CONTRIBUTING.md`, `docs/JURISPRUDENCIA.md`, `docs/WORKFLOW.md` | FR-084, FR-090 a FR-093 |
