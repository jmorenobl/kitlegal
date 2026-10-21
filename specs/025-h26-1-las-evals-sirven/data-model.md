# Data model: H26.1

Lo que el hito añade o cambia en el formato de las evals, en las consultas de la caché de una sesión, en lo que se lee
de una sesión, en el juicio y en el informe del job. Todo vive en `internal/evals` salvo el esquema
(`schemas/eval.yaml.json`), las evals (`evals/`), el manifiesto de grabación (`internal/evals/testdata/grabaciones.json`)
y la skill (`skills/legal-core/SKILL.md`). Ningún tipo del binario cambia. Contratos: [formato de eval](./contracts/formato-de-eval.md),
[caché de la sesión](./contracts/cache-de-la-sesion.md), [juicio e informe](./contracts/juicio-e-informe.md) y
[skill legal-core](./contracts/skill-legal-core.md).

## 1. La eval (`Eval`, `internal/evals/formato.go`)

Dos campos opcionales nuevos, cada uno con su tipo:

| Campo Go | Clave YAML | Tipo | Requisitos |
|---|---|---|---|
| `Lecturas` | `lecturas` | `[]LecturaDeLaEval` | FR-006 a FR-008 |
| `SinServir` | `sin_servir` | `[]LecturaSinServir` | FR-012 a FR-014 |

- **`LecturaDeLaEval`** `{Norma, Bloque, Motivo string}` (`norma`, `bloque`, `motivo`): un bloque de una norma
  consolidada del BOE que la caché de las sesiones de la skill sirve, y por qué la eval lo provoca (clase (b) del ADR
  0043). No es un comando, ni una cita, ni da motivo, ni cuenta en ningún umbral (FR-007).
- **`LecturaSinServir`** `{Norma, Bloque, Verbo string}` (`norma` y `bloque`, o `norma` y `verbo`): un bloque, o el
  `indice` o los `metadatos` de una norma, que la caché de las sesiones de **esa** eval no sirve con ninguna fecha
  (FR-013). Lleva `Bloque` o `Verbo`, nunca los dos (el esquema).

Reglas (contracts/formato-de-eval.md §2):

1. El esquema: `lecturas`, lista no vacía; cada entrada con las tres claves y ninguna más, `norma` y `bloque` con las
   formas de `$defs/norma` y `$defs/bloque`, `motivo` con algún carácter que no es blanco. `sin_servir`, lista no vacía;
   cada entrada `{norma, bloque}` o `{norma, verbo}`, `verbo` `indice` o `metadatos`, sin más claves.
2. `LeerEval`, detrás del esquema: un bloque de `lecturas` que ya es un comando de bloque de la eval (misma norma y mismo
   bloque, con o sin `fecha`) o una de sus `citas`, y un bloque de `sin_servir` que es una de sus `citas`, hacen de la
   eval un fichero mal formado; el error empieza por el nombre del fichero y nombra la entrada.
3. Un bloque de `sin_servir` puede ser un comando de bloque de la eval: es lo que la eval quiere provocar (FR-016).

Métodos sin exportar que usan el juicio y la preparación:

- `LecturaSinServir.consulta(Invocacion) bool`: la invocación es del applet `boe` y consulta esa lectura —el bloque:
  verbo `articulo` o `articulos`, primer argumento de posición la norma y el bloque entre los demás, con o sin
  `--fecha` (`lecturaDeLosArgumentos`); la consulta de norma: ese verbo y esa norma (`esDeLaNorma`)—.
- `Eval.sinServir(Invocacion) bool`: alguna entrada de `SinServir` la consulta.
- `Eval.comandoSinServir(ComandoEsperado) bool`: el comando es de la forma bloque y su norma y su bloque son una entrada
  de bloque de `SinServir`.

## 2. Las consultas de la caché de una sesión (`internal/evals/consultas.go`)

`Consulta` y `Origen` no cambian. Dos puntos nuevos:

| Punto | Valor | De dónde sale | `Origen.Eval` |
|---|---|---|---|
| `PuntoLecturaDeLaEval` | `"lectura de la eval"` | cada entrada de `lecturas` (FR-006) | el fichero de la eval |
| `PuntoLoQueLlevaALeer` | `"lo que la skill lleva a leer"` | cada bloque y cada norma de `data/` (FR-001) | `data/plazos.yaml`, `data/cobertura.yaml` o `data/normas.yaml` |

- `ConsultasNecesarias(conjunto)` anota, detrás de las citas de cada eval, sus `lecturas` con `anotarBloque`: sin fecha y
  con `--fecha 9999-12-31`.
- `anotarLosDatos(dir) error`, del acumulador, lee `dir/plazos.yaml` con `plazos.LeerTipos`, `dir/cobertura.yaml` con
  `skills.LeerCobertura` y `dir/normas.yaml` con `skills.LeerNormas`, y anota, en este orden y sin repetir: cada bloque de
  los tipos (sus `Preceptos`, sus `Unidades` en el orden de `plazos.Unidades()` y sus `PreceptosDelComputo`), cada
  `Precepto` de la cobertura y el `Aplicable` de su `Norma`, todos con `anotarBloque`; y, de cada norma de esos bloques,
  de la cobertura y de `normas.yaml`, `indice` y `metadatos`. Ningún identificador escrito en el código (FR-002). Un
  fichero que no se lee o que su lector rechaza es un error, y entonces no anota nada.
- `consultasDeLaCache(evals, dir) ([]Consulta, error)`: las de `ConsultasNecesarias(evals)` y, detrás, las de
  `anotarLosDatos(dir)`, con el mismo acumulador: la invocación que ya estaba suma su origen.
- `sinSusLecturas(evals) []Eval` (`medida.go`): copias de las evals sin sus `lecturas`, para la base de la medida.
- `sinLoQueNoSeSirve(consultas, eval) []Consulta`: las consultas menos las que `eval.sinServir` reconoce, convertida cada
  una en la `Invocacion` que la ejecuta (applet, verbo, argumentos, `Consulta: true`).
- `datosDelRepositorio = "../../data"`: el `data/` del repositorio, relativo al paquete.

Quién prepara qué (contracts/cache-de-la-sesion.md §1):

| Caché | Consultas | Requisitos |
|---|---|---|
| La de una sesión del job o del sondeo (`prepararLaCache`) | `sinLoQueNoSeSirve(consultasDeLaCache(evals, datosDelRepositorio), eval)` | FR-001, FR-006, FR-013 |
| La base de la medida del juez (`leerLasEvalsDeHoy`, `baseVigente`) | `consultasDeLaCache(sinSusLecturas(evals), datosDelRepositorio)`: sin las `lecturas`, que las sesiones de los informes versionados no tenían (research D4) | FR-004 |

`Falta.String` nombra las nuevas desde su origen (`queFaltaDesde`):
`<fichero de la eval>: la lectura <norma> <bloque> sin su bloque: «…» terminó con código N: …` y
`data/plazos.yaml: lo que la skill lleva a leer <orden>: «…» terminó con código N: …` (FR-003).

## 3. Lo que se lee de una sesión (`internal/evals/sesion.go`)

| Campo | Tipo | Qué es | Requisito |
|---|---|---|---|
| `Llamada.SinSobre` | `bool` | el transcript trae su `tool_result`, es un error (`is_error` verdadero) y ninguno de los textos de su contenido es un objeto JSON con `ok` | FR-023 |
| `Sesion.OrdenesDeBash` | `[]Texto` | cada uso de Bash cuyo `input.command` lleva la palabra `kitlegal` y cuyo `tool_result` trae el transcript, con su orden tal cual y su salida, en el orden de sus `tool_use` | FR-025, FR-026 |

`Sesion.Textos` no cambia (lo que recibe el juez). `Llamada.Error` y `Llamada.Clase`, tampoco.

## 4. El reparto de una invocación en el juicio (`internal/evals/juzgar.go`)

Cada invocación de la sesión —de la traza o por una llamada— se informa en `invocaciones` y va, como mucho, a una de las
tres listas, por la primera fila que la describe:

| # | La invocación | `invocaciones[]` | Va a | Con |
|---|---|---|---|---|
| 1 | no consulta (ayuda, `--describe`, `--dry-run`, el proceso de `mcp serve`) | como hoy | ninguna | — |
| 2 | es una llamada sin sobre (`SinSobre`) | `codigo: null`, `motivo: "sin_sobre"` | `otras_fallidas` | `codigo: null`, `motivo: "sin_sobre"` |
| 3 | consulta una lectura del `sin_servir` de su eval | como hoy | `lecturas_sin_servir` | su `codigo` (también `null` o 0); y, si es una orden de salida no escrita (fila 6), `motivo` y `bash` |
| 4 | no tiene código (el tope la cortó; la llamada sin resultado) o terminó con 0 | como hoy | ninguna | — |
| 5 | terminó con 4 o 5 | como hoy | `fuera_de_lo_grabado` | su `codigo` |
| 6 | es de la traza, terminó con 1 y la enlaza FR-026 con un único uso de Bash cuyo resultado trae la línea de FR-025 | como hoy | `otras_fallidas` | `codigo: 1`, `motivo: "salida_no_escrita"`, `bash` |
| 7 | cualquier otro código | como hoy | `otras_fallidas` | su `codigo` |

**El enlace de FR-026.** Una invocación de la traza con código 1 se enlaza con un uso de `Sesion.OrdenesDeBash` si la
orden de ese uso contiene como palabras (`contieneComoPalabra`) el applet, el verbo y cada argumento de la invocación,
y alguna línea de su salida empieza por el prefijo de la escritura fallida (D8 de research); y si es el único uso de la
sesión que lo cumple para esa invocación. Ninguno o más de uno: fila 7 (FR-027).

**Lo que no cambia** (FR-028): si la sesión pasa, sus motivos y lo que satisface o ejecuta cada invocación, salvo FR-016.

**FR-016.** `satisface(invocacion, comando, admiteLaFallida)`, con `admiteLaFallida = eval.comandoSinServir(comando)`:
`consultoComoPide` admite entonces también el 4 y el 5. La lectura tiene que ser en la fecha del comando, como hoy.

## 5. El resultado de una sesión en el informe (`ResultadoDeEval`, `internal/evals/juzgar.go`)

| Campo | Clave JSON | Tipo | Requisito |
|---|---|---|---|
| `InvocacionInformada.Motivo` | `motivo` (se omite vacío) | `string` | FR-023 |
| `InvocacionFallida.Codigo` | `codigo` | `*int` (antes `int`) | FR-023, FR-015 |
| `InvocacionFallida.Motivo` | `motivo` (se omite vacío) | `string`: `sin_sobre` o `salida_no_escrita` | FR-023, FR-024 |
| `InvocacionFallida.Bash` | `bash` (se omite vacío) | `string`: la orden de Bash tal cual | FR-024 |
| `ResultadoDeEval.LecturasSinServir` | `lecturas_sin_servir`, detrás de `otras_fallidas` | `[]InvocacionFallida`, `[]` si no hay | FR-015 |

`FueraDeLoGrabado` ya no lleva ninguna lectura de `sin_servir`, así que tampoco el `fuera_de_lo_grabado` de la raíz del
informe, que las junta (FR-022); sus códigos son siempre 4 o 5 y `FueraDeLoGrabadoDelInforme.Codigo` sigue siendo un
entero. En `informe.md`, la fila de una invocación con `codigo: null` y `motivo` dice «sin sobre» en lugar de «sin código
(sesión cortada)».

## 6. La serie (`TasaDelInforme`, `internal/evals/informe.go`)

| Campo | Clave JSON | Tipo | Valor | Requisito |
|---|---|---|---|---|
| `SesionesConLecturaSinServir` | `sesiones_con_lectura_sin_servir` | `*int` | en la serie de una eval con `sin_servir`, cuántas de sus sesiones legibles tienen algo en `lecturas_sin_servir`, también 0; `null` en las demás | FR-015 |

## 7. Los umbrales nuevos (`Umbral`, `internal/evals/umbrales.go`)

Por cada modelo y cada modo del plan, detrás de todos los de hoy (research D12): `fuera_de_lo_grabado` en el informe
de toda skill, y `lectura_sin_intentar` solo en el de una skill con al menos una eval con un comando de bloque en su
`sin_servir` (hoy, `legal-core`), como `plazo_sin_fecha` solo en el de la que tiene evals de plazo. Las «sesiones
medidas» son las de research D11.

| Nombre | Medida | Total | Umbral | `decide` | Requisito |
|---|---|---|---|---|---|
| `fuera_de_lo_grabado:<modelo>:<modo>` | las lecturas de `fuera_de_lo_grabado` de esas sesiones | las sesiones medidas | `<=` 0 | `false` | FR-021, FR-022 |
| `lectura_sin_intentar:<modelo>:<modo>` | las sesiones de esas evals que no intentaron la lectura de un comando de bloque en su `sin_servir` | las sesiones medidas de las evals con un comando de bloque en su `sin_servir` | `<=` 0 | `true` solo con el modelo que decide | FR-033 |

Con el plan del job: `legal-core`, de 10 umbrales a 18 (de 10 que deciden a 12); `boe-legislacion` y `jurisprudencia`,
de 15 a 19 (los 11 que deciden, sin cambiar). El motivo de un `lectura_sin_intentar` que no se cumple, en el
veredicto, nombra sus sesiones separadas por «; », como el de `plazo_sin_fecha`.

## 8. Las evals nuevas y las que cambian (`evals/`)

| Fichero | Qué cambia | Requisito |
|---|---|---|
| `evals/legal-core/11-plazo-reposicion-sin-leer-el-precepto.yaml` | nueva: la pregunta, los comandos y el plazo de la 05; `a124` de `BOE-A-2015-10565` en `sin_servir`; sin `citas` ni `redaccion_fechada` | FR-017 |
| `evals/legal-core/12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml` | nueva: cuotas de la Seguridad Social, en Leganés y notificada el 9 de septiembre de 2026; `daprimera` en `sin_servir`; `prohibidos` con `plazos calcular`; `no_cubierto` con `BOE-A-2015-10565`; sin `citas` ni `comandos` | FR-018 |
| `evals/boe-legislacion/14-trlrhl-impuestos-por-materia.yaml` | `lecturas`: `a15` de `BOE-A-2004-4214` | FR-011 |
| `evals/boe-legislacion/15-irpf-rendimientos-por-materia.yaml` | `lecturas`: `a18` de `BOE-A-2006-20764` | FR-011 |
| `evals/boe-legislacion/16-lrjsp-legalidad-por-materia.yaml` | `lecturas`: `a26` y `a27` de `BOE-A-2015-10566` | FR-011 |

El texto exacto, en contracts/formato-de-eval.md §3.

## 9. El manifiesto de grabación (`internal/evals/testdata/grabaciones.json`)

Tres entradas nuevas en `normas`, con la forma de las de hoy (`busqueda`, `titulo_empieza_por`, `bloques`, `para`), las
búsquedas que ya grabó H5 (research V11): «régimen jurídico del sector público» / `Ley 40/2015,` / `a26`, `a27`;
«impuesto renta personas físicas» / `Ley 35/2006,` / `a18`; «reguladora de las haciendas locales» /
`Real Decreto Legislativo 2/2004,` / `a15` (FR-010). `grabar_datos` deja las cuatro respuestas en
`internal/evals/testdata/boe.legislacion-consolidada/`, y con ellas vuelve a grabar el `robots.txt` del BOE de ese
directorio, que `internal/httpx` pide antes que ellas (`gates/grabaciones.md`).
