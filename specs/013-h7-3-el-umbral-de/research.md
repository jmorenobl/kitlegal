# Research: H7.3 · El umbral de expresiones prohibidas decide, `boe-legislacion` sin el vocabulario que provoca el ruido, y evals en paralelo con un sondeo local

Modo desatendido: cada decisión se tomó con el «Criterio de decisión autónoma» de la constitución y lleva la alternativa
rechazada y por qué. Toda afirmación sobre Claude Code, GitHub Actions, `gh`, `gosec`, Go o el propio repositorio remite
a la tabla V, comprobada en local y sin red; lo que no se pudo comprobar así son los supuestos S1-S7. Nada de este
research lo hace una persona a mitad del run.

## V · Verificaciones en local

| Id | Qué | Dónde se comprobó |
|---|---|---|
| V1 | Claude Code **2.1.270** (la versión que fija el job, `VERSION_DE_CLAUDE_CODE`) está en la imagen Docker local `kitlegal-t036-a` que H5 construyó con `npm install -g @anthropic-ai/claude-code@2.1.270`: `claude --version` da `2.1.270 (Claude Code)`; binario `/usr/local/lib/node_modules/@anthropic-ai/claude-code/bin/claude.exe`. Se leyó con `docker run --rm --network none … grep -a -o` | imagen `kitlegal-t036-a:latest`, sin red |
| V2 | En 2.1.270, el evento de reintento: `subtype:"api_retry",attempt:…,max_retries:…,retry_delay_ms:…,error_status:…,error:e6t(…)`; `e6t`: 529 u `overloaded_error` → `"overloaded"`, **429 → `"rate_limit"`**, 401/403 → `"authentication_failed"`, ≥ 408 → `"server_error"`; reintentos por defecto `Kcs=10` (tope `Obt=15` para `CLAUDE_CODE_MAX_RETRIES`) | V1, `grep -a -o 'subtype:"api_retry"…'`, `'function e6t(…'`, `'…=10,…=300,…=15'` |
| V3 | En 2.1.270, el mensaje del límite de uso: plantilla ``You've hit your ${e}${n}${d}`` con `five_hour:"session limit"`, `seven_day:"weekly limit"`, `seven_day_opus:"Opus limit"`, `seven_day_sonnet:"Sonnet limit"`, `seven_day_overage_included:"Fable limit"`, `overage:"usage credit limit"`; y la lista propia de Claude Code de principios de mensaje de límite `G2n=["You've hit your","You've reached your","You're out of usage credits",…,"You're out of extra usage",…]`. El `result` de un error de la API es `subtype:"success"` con `is_error:Lo`, `Lo=Ue.isApiErrorMessage===!0` (verdadero para un mensaje de error de la API), `api_error_status:Ir` (`Ir=Ue.apiErrorStatus??null`) y el texto en `result`. Un 429 que no es de la cuenta dice `Server is temporarily limiting requests (not your usage limit)` | V1, `grep -a -o 'You.ve hit your …'`, `'var G2n=[…'`, `'subtype:"success",api_error_status…'` |
| V4 | En 2.1.270, `CLAUDE_CONFIG_DIR` es el directorio de configuración (`(s()??a(R(),".claude"))`, con `s` = `process.env.CLAUDE_CONFIG_DIR`); las skills de usuario son `Se()/skills` y solo se cargan con `userSettings` activo (`if(Sr("userSettings"))n.push({dir:r,scope:"user"})`); el `CLAUDE.md` de usuario es `Se()/CLAUDE.md`; el fichero global es `.claude${kB()}.json` junto a él; existe `CLAUDE_CODE_TMPDIR` («Point CLAUDE_CODE_TMPDIR at a private (0700) directory you own») | V1 |
| V5 | En 2.1.284 y 2.1.278 (`~/.local/share/claude/versions/`, binario Bun de un fichero; lo que ejecuta el sondeo en este Mac): lo mismo que V2-V4 y además: `--setting-sources` admite `user`, `project`, `local`; sin `user` no se cargan ni el `settings.json` ni las skills ni el `CLAUDE.md` de usuario, y los hooks y plugins salen de las fuentes activas; todo el estado (`.claude.json`, `projects`, `sessions`, `todos`, `shell-snapshots`, `statsig`, `.credentials.json`, `history.jsonl`…) cuelga del directorio de configuración; orden de credenciales `ANTHROPIC_AUTH_TOKEN` → `CLAUDE_CODE_OAUTH_TOKEN` → … → llavero/`.credentials.json`, y **`ANTHROPIC_API_KEY` tiene prioridad en `-p`**; con `CLAUDE_CODE_OAUTH_TOKEN` no se lee el almacén para el token; el servicio del llavero es `Claude Code-credentials` y, con `CLAUDE_CONFIG_DIR`, lleva el sufijo de los 8 primeros hex de `sha256(dir)`; en `-p` el diálogo de confianza se da por aceptado; `--bare` ignora `CLAUDE_CODE_OAUTH_TOKEN`; un 429 de suscriptor con las cabeceras del límite de la suscripción no se reintenta | offsets de byte del binario 2.1.284: @177244638, @179605545, @187560179, @180310723, @185444073, @189186907, @180359130, @180361834, @180383411, @179580139, @180266092, @186797608 |
| V6 | Esquema oficial de flujos de GitHub Actions (`workflow-v1.0`) que trae la extensión de VS Code `github.vscode-github-actions` 0.31.5 (`dist/server-node.js`; copia extraída en `/tmp/gha/chunk.txt`): `jobs.<id>.concurrency` admite `github, inputs, vars, needs, strategy, matrix` (@422946); `jobs.<id>.timeout-minutes` y `jobs.<id>.name` son `…-strategy-context` y admiten `matrix`; el `concurrency` de nivel de flujo solo `github, inputs, vars` (@422490); `group`: «the queued job or workflow will be `pending`. Any previously pending job or workflow in the concurrency group will be canceled. To also cancel any currently running job or workflow in the same concurrency group, specify `cancel-in-progress: true`» (@423380); `timeout-minutes` «Default: 360» (@414757); `include`: «the key:value pairs in the object will be added to each of the matrix combinations if none of the key:value pairs overwrite any of the original matrix values» | fichero de la extensión, `grep -b` |
| V7 | `gh` 2.101.0: `gh pr checks --json` da `bucket` con `pass`, `fail`, `pending`, `skipping`, `cancel`; el binario tiene `pr/checks.eliminateDuplicates`, pero su clave y su orden no se leen sin el fuente | `rtk proxy gh pr checks --help`; cadenas del binario |
| V8 | `gosec` v2.28.0 (el de `tools/golangci-lint`): G204 (`rules/subproc.go:62-94`) salta el argumento 0 si es un identificador que nombra un parámetro de la función y exige que **todo otro argumento** se resuelva a constante; un selector (`e.Guion`) no se resuelve nunca (`resolve.go:77-101`, `TryResolve` no trata `*ast.SelectorExpr`); G702 (`analyzers/commandinjection.go:24-49`) tiene por fuentes `os.Args`, `os.Getenv`, `*http.Request` y `*bufio.Reader/Scanner`, sin saneadores: ni los valores de `flag` ni `os.Environ()` son fuente | `$(go env GOMODCACHE)/github.com/securego/gosec/v2@v2.28.0/` |
| V8b | `gosec` v2.28.0, G301/G302/G306 (`rules/fileperms.go:55-112`): solo casan las funciones de paquete `os.WriteFile`, `os.OpenFile`, `os.Chmod`, `os.Mkdir` y `os.MkdirAll` (`MatchCallByPackage`), con 0600 o 0750 como máximo; los métodos de `*os.Root` no: los tests escriben los sustitutos ejecutables (0755) con `os.OpenRoot(t.TempDir())` y `root.WriteFile` | `$(go env GOMODCACHE)/github.com/securego/gosec/v2@v2.28.0/rules/fileperms.go` |
| V9 | `.golangci.yml`: `forbidigo` prohíbe `fmt.Print*`, `os.Stdout`, `os.Stderr` y `os.Exit` en todo el árbol, con excepciones solo para `^cmd/` y el ayudante e2e de `internal/app` | `.golangci.yml:229-262`, exclusiones `:396-410` |
| V10 | `go help testflag`: `-timeout` vale 10 min por omisión y `0` lo desactiva. `go help test`: en modo lista de paquetes, un test que pasa solo imprime la línea `ok` | `go help testflag`, `go help test` |
| V11 | `encoding/json/v2`: `omitzero` omite el valor cero de Go (un puntero nil) | `go doc encoding/json/v2` |
| V12 | Este Mac no tiene `timeout` en el `PATH` (`which timeout` → «timeout not found»): el tope no puede apoyarse en coreutils fuera de Linux | `which timeout` |
| V13 | `LeerTrazas` sobre un `traza/` vacío falla («no hay ningún fichero sin la línea clone…»): una sesión sin strace se lee con el informe del job como ilegible | `internal/evals/trazas.go`, `raiz()` |
| V14 | El informe final: `evals:<skill>:<nombre>` busca en `umbrales` el `nombre` que va tras el segundo `:`, rehace la comparación en coma flotante y da «comprobado por su control» solo con `decide: true`, `cumple: true` y coherente; `ci:<ruta>[:<Test>]` exige la ruta y `func <Test>(` o `<Test>:` en la cabeza y `make ci` en verde | `scripts/workflow/informe.sh:185-272`; forma de la celda en `scripts/workflow/comun.sh:106-112` |
| V15 | El cierre: `recoger_evals` toma el informe de cada comprobación de `evals` con `bucket` `pass` o `fail` cuyo nombre casa `^evals \(([a-z0-9-]+)\)$`; `medir` espera mientras haya alguna `pending` y cuenta `fail` y `cancel` como rojas | `scripts/workflow/cierre.sh:49-68, 112-122` |
| V16 | `release_test.go` lee los flujos con `go.yaml.in/yaml/v3`; solo `ci.yml` y `release.yml` se leen en estricto: `evals.yml` puede ganar claves sin romperlo | `release_test.go:18, 554-558, 1070-1165` (agente de búsqueda) |
| V17 | Medidas de los informes versionados de H7.1 y H7.2 con la comparación de `ExtraerExpresionesProhibidas` (misma expresión regular, prototipo fuera del repositorio, `/tmp/h73-proto/`), y de la prosa de `SKILL.md` con el algoritmo de D4 (`/tmp/h73-proto/prosa/`): v0.1.2, 14 párrafos y la fecha de la línea 191; una copia de `SKILL.md` con los cambios C1-C11 escritos como dice [contracts/skill-boe-legislacion.md](./contracts/skill-boe-legislacion.md) (`/tmp/h73-proto/SKILL-v013.md`), 0 párrafos, 0 fechas y 270 líneas | ver «Causa de raíz» y D2 |
| V18 | Una credencial que no sirve **no** sale con 0: con Claude Code 2.1.270 y un `CLAUDE_CODE_OAUTH_TOKEN` inválido, la orden de la sesión deja como último mensaje un `result` con `is_error` y «Failed to authenticate. API Error: 401 OAuth access token is invalid.», y `claude -p` sale con **código 1** (H5, V61, casos (2) y (4)). En `main`, antes de H7.3, `motivoSinTerminar` devolvía `código 1` antes de mirar el `result` (`internal/evals/sesion.go:386-403` de `main`), así que ese texto no llegaba al motivo; desde T003 lo lleva. El único transcript versionado con un `result` con `is_error` (`internal/evals/testdata/sesiones/leer-sesion/result-con-is-error/`, «API Error: 529 …») es de una sesión con código 0. En 2.1.284, el mensaje de un modelo que no existe es «There's an issue with the selected model (<id>). It may not exist or you may not have access to it.», que se construye con `Qo({content:…,error:"model_not_found"})`, la función de los mensajes de error de la API (`function Qo({content:e,apiError:n,…,error:h,…})`, que pone `isApiErrorMessage:!0`): llega, como el de la credencial, en un `result` con `is_error` (V3) | `specs/006-h5-skill-boe-legislacion/research.md` V61; `internal/evals/sesion.go`; `grep -r '"is_error":true' internal/evals/testdata`; binario 2.1.284, `grep -a -o` de las dos cadenas |

Supuestos no verificados:

- **S1** · Lo de V5 que no se releyó en 2.1.270 (el orden de credenciales, el sufijo del llavero, `--setting-sources` sobre
  hooks y plugins) vale igual en 2.1.270. Se releyó en 2.1.270 lo que el job necesita (V2-V4).
- **S2** · Claude Code escribe su estado de sesión bajo el directorio de configuración y los temporales bajo
  `TMPDIR`/`CLAUDE_CODE_TMPDIR` (V4, V5); fuera de ellos, nada que otra sesión escriba (no se leyó cada ruta del binario).
- **S3** · Un trabajo que espera por su grupo de concurrencia tiene su comprobación en un estado que `gh` pone en
  `pending`; y la clave de `eliminateDuplicates` (V7). El diseño de D11 no depende de ninguno de los dos.
- **S4** · El trabajo en espera no consume su `timeout-minutes` hasta que empieza.
- **S5** · El `/bin/bash` de macOS es bash 3.2: los guiones nuevos no usan nada posterior (ni `wait -n` ni arreglos
  asociativos), y el repartido de sesiones no vive en bash (D7).
- **S6** · `trap … EXIT` de bash se ejecuta también cuando el guion termina por un código distinto de 0 (`set -e`).
- **S7** · El efecto de `SKILL.md` v0.1.3 sobre las respuestas no se puede medir sin modelo: lo miden el job de cierre
  (SC-001) y el escenario del quickstart (SC-002).

## Causa de raíz del ruido en v0.1.2 (FR-014)

**Partida.** El research de H7.2 («Causa de raíz del ruido») halló un párrafo de transición que Sonnet 5 escribe tras la
última orden —`graph check`— y que `claude -p` entrega como principio de la respuesta, y lo atacó en posición, alcance y
motivo (v0.1.2). El cierre de H7.2 lo bajó de 34 de 51 a 10 de 51, y cambió su forma.

**Medida del cierre de H7.2** (`specs/012-h7-2-la-consulta-repetida/gates/evals/boe-legislacion.json`, V17):

| Medida | Sonnet 5 | Haiku 4.5 |
|---|---|---|
| Respuestas en evals que activan la skill | 51 | 30 |
| Última invocación: `graph check` | 51 de 51 | 30 de 30 |
| Con alguna expresión de la lista | 10 (41 sin ninguna, con la misma última orden) | 0 |
| Empiezan por «Sin hallazgos que trasladar.» | 8 (03-03, 06-02, 13-01, 13-02, 14-03, 15-01, 15-02, 15-03) | — |
| … seguidas de un anuncio («Redacto la respuesta.», «Ya puedo responder.», «Ya tengo todo lo necesario para responder.», «Con esto ya tengo la respuesta completa») | 6 de 8 | — |
| Empiezan por «No hay hallazgos, así que respondo con el texto vigente.» | 1 (14-02) | — |
| Copian el ejemplo de la forma, con su fecha inventada | 1 (19-02: «…por la de 20250101... en realidad las fechas del hallazgo son 20180309 y 20200206») | — |

**Dónde lo enseña `SKILL.md` v0.1.2, línea a línea** (fuera del código y de la región generada, V17):

| Líneas | Texto de v0.1.2 | Lo que repite la respuesta |
|---|---|---|
| 258-259 (regla 7) | «**Una comprobación con hallazgos no es un fallo.** … es un resultado, con hallazgos o sin ellos, … trasládalos como dice «Memoria de consultas»» | «Sin hallazgos que trasladar.» (8 de 10): la única frase de la skill que empareja el resultado vacío con el verbo |
| 260 (regla 7) | «responde igual con el texto de `kitlegal boe`» | «así que respondo con el texto vigente» (14-02) |
| 106 (paso 5) | «Cuando ya no quede nada por leer, y antes de redactar la respuesta, comprueba la memoria de consultas…» | «Redacto la respuesta.» (13-01, 13-02), «Ya tengo todo lo necesario…», «Con esto ya tengo la respuesta completa» |
| 113-114 (paso 5) | «Si da `version-obsoleta`, dilo…; si no, no digas nada de ella.» | nombra el caso vacío como algo que callar |
| 117-119 (paso 5) | «…tampoco para decir que no hay nada que decir… la respuesta no nombra la memoria de consultas, `kitlegal graph`…, los códigos de salida, los hallazgos, las clases del binario…, el JSON ni el sobre» | la enumeración enseña cada palabra que prohíbe |
| 121-123 (paso 5) | «Sin `version-obsoleta`, no digas nada de lo consultado antes…: la comprobación sin hallazgos no distingue…» | «sin hallazgos» otra vez, junto al caso vacío |
| 185-197 («Memoria de consultas») | «Trasládalo con su forma fija», «tal como las da el hallazgo», «`fuente-caducada` no se traslada», «Sin `version-obsoleta`, la respuesta no dice nada de la memoria de consultas» | «trasladar» y «hallazgo» ligados a la comprobación |
| 191 (ejemplo) | «…la redacción con fecha de vigencia 20161002, la que se consultó antes, ha sido sustituida por la de 20250101…» | 19-02 copia la frase con 20250101, una redacción que el BOE no tiene |
| 73-74, 74, 106, 113, 118, 178, 197, 256, 259, 261 | «memoria de consultas» (10 veces en la prosa, con las dos que remiten a la sección) | la forma dominante de H7.1 (24 de 34) |
| 76, 89, 118, 204, 239, 258 | «el código 4 o 5», «el código 3», «Códigos de salida», «código 0» | «código 0» (7 de 34 en H7.1) |

**Conclusión.** El modelo repite la palabra de la **prosa**, no la de la salida del binario: «trasladar» no está en
ninguna salida de `kitlegal`, y la pareja «hallazgos … trasladar» es literal de la regla 7; «respondo con el texto» es
la regla 7; «redacto la respuesta», el paso 5. La salida de `graph check --json` lleva la clave `hallazgos` (también la
región generada), pero Haiku 4.5 lee la misma salida en 30 de 30 sesiones y no la repite nunca, y en v0.1.1 la forma
dominante era «memoria de consultas», que no está en ninguna salida. Por eso el cambio va a la prosa (D1) y no a cómo
se lee la salida; la clave `data.hallazgos` se queda en código, donde el agente la lee (FR-010). Una prohibición más no
lo corrige: v0.1.2 ya enumeraba «los hallazgos» y al hacerlo enseñaba la palabra (FR-014).

**La comprobación en su sitio (FR-015).** `graph check` es la última orden en 51 de 51 sesiones de Sonnet 5, en las 41
limpias igual que en las 10 con ruido, y en 30 de 30 de Haiku 4.5, que no lo escribe nunca: el sitio es común a las
respuestas con ruido y a las limpias, y lo que distingue a las 10 es el vocabulario. **Decisión: la comprobación no
cambia de sitio** —una vez por norma citada, después de leer sus bloques y antes de la respuesta, como decidió H7.1—; lo
que cambia es que la orden ya no va unida a «antes de redactar la respuesta», que 2 de 10 repiten (C4). Alternativa
rechazada: moverla al final del paso 3; con varias normas seguiría siendo la última orden, y una lectura posterior a la
comprobación apagaría `version-obsoleta` (H7.1 FR 024) sin evidencia de que el sitio cause el ruido.

## Decisiones

### D1 · `SKILL.md` v0.1.3: quitar el vocabulario, no añadir prohibiciones (FR-010 a FR-017)

Once cambios, C1-C11, cada uno trazado a su fila de «Causa de raíz» (detalle y texto en
[contracts/skill-boe-legislacion.md](./contracts/skill-boe-legislacion.md)): la prosa llama a la comprobación «la
comprobación de la redacción» —«si la redacción ha cambiado desde una lectura anterior», palabras que la respuesta
puede usar—; ninguna frase nombra el caso vacío; la regla 7 dice solo qué hacer si la comprobación no termina con `0` y
cuándo va `⚠ REDACCIÓN MODIFICADA:`; el ejemplo lleva dos `AAAAMMDD`; los códigos van en código (`3`, `4`, `5`), sin la
palabra «código» delante; la sección «Memoria de consultas» pasa a «Redacción modificada desde una lectura anterior».
Se quedan la forma, las fechas tal como las da el binario, las citas, los avisos, los cinco pasos, las reglas 1 a 6 y
«Nada de otra conversación» (FR-016). «Traslada» se queda donde es de los avisos de vigencia (H5.1: paso 4, paso 5,
«Cómo se cita», regla 3) y sale de todo lo que es de la comprobación: en los cierres de H7.1 y H7.2, «que trasladar» va
siempre con «hallazgos» o «memoria de consultas» (V17), nunca con los avisos.

Alternativas rechazadas: **ampliar la enumeración** de lo que la respuesta no dice (es lo que enseñaba la palabra);
**cambiar la salida de `graph check`** (fuera de alcance, y la evidencia no la señala); **dejar la enumeración y quitar
solo la regla 7** (el paso 5 seguiría enseñando «hallazgos», «memoria de consultas», «códigos de salida» y «JSON»).

### D2 · La tercera familia, `anuncio` (FR-020, FR-021, FR-024)

Clave `anuncio` en la lista, con 14 expresiones, todas medidas en los cierres de H7.1 o H7.2 (V17):

`que trasladar`, `hace falta trasladar`, `ya puedo responder`, `con esto puedo responder`, `y puedo responder`,
`tengo todo lo necesario`, `tengo lo necesario`, `redacto la respuesta`, `respondo con el texto`,
`respondo con el contenido`, `ya tengo la respuesta`, `ya tengo el texto`, `así que respondo`, `sin redacciones cambiadas`.

Calibrado con las tres familias (V17): H7.1, **35 de 93** —02 (1), 03 (3), 04 (3), 05 (3), 06 (3), 07 (3), 08 (2), 09 (2),
13 (3), 14 (3), 15 (3), 16 (2), 17 (3) y 19 (1)—; H7.2, **10 de 93** —03 (1), 06 (1), 13 (2), 14 (2), 15 (3) y 19 (1)—;
exactamente el reparto de FR-021. La familia sola marca, en H7.1, 03 (3), 04 (3), 05 (1), 06 (2), 07 (2), 08 (1), 09 (1),
13 (2), 14 (2), 15 (3), 16 (2) y 17 (2) —24—, y en H7.2, 03 (1), 06 (1), 13 (2), 14 (2) y 15 (3) —9—, todas ya marcadas por
la maquinaria. Ninguna aparece en `testdata/` (`grep -r -i -F` de las 14, sin resultados); lo comprueba de verdad la
subprueba de los bloques (D3).

Rechazadas, con su medida: **«aquí tienes» y «aquí está»** marcarían 19-01 y 19-02 de H7.1, que la lista no marca
(respuestas limpias que empiezan por el texto); **«puedo responder» a secas** casaría con «no puedo responder con esta
fuente», que es lo que la regla 2 pide decir cuando la fuente falla; **«la respuesta:»** es demasiado corriente; y
**«sin avisos de vigencia» y parecidas**, por FR-020 (dicen derecho, no maquinaria). Formas no medidas («sin cambios de
redacción», «nada que decir») no entran: la lista se calibra con lo observado, y lo dicho con otras palabras sigue siendo
la limitación declarada.

### D3 · Calibrado y comprobaciones de H7.2 extendidos (FR-021, FR-022, FR-095)

La subprueba `expresiones-calibradas` de `TestEvalsDelRepositorio` pasa a leer los dos informes y a comparar, por eval
y por familia, las tres cuentas y, además, las respuestas con alguna expresión de la lista (la medida del umbral), con
los repartos de D2. `expresiones-en-los-bloques` y `expresiones-de-la-skill` aplican la lista entera, así que la tercera
familia entra sin tocarlas: basta con que `ExtraerExpresionesProhibidas` recorra las tres. Alternativa rechazada:
comparar solo el total por eval (FR-021 lo permitiría), porque una familia vaciada por error no se vería.

### D4 · La prosa de `SKILL.md`, comprobada en `make ci` (FR-091)

Subprueba nueva `prosa-de-la-skill` de `TestEvalsDelRepositorio`: de `skills/boe-legislacion/SKILL.md` quita los
bloques delimitados (```` ``` ````), la región generada (entre sus dos marcas) y cada tramo de código en línea
(sustituido por un espacio, para que no junte palabras), junta en un espacio los saltos de línea de cada párrafo —como
se lee; una línea en blanco o una que abre un elemento de lista empieza otro—, y aplica `ExtraerExpresionesProhibidas` con la lista del repositorio a cada párrafo; y busca en **todo** el
fichero la fecha `(?:^|[^0-9])[0-9]{4}(?:0[1-9]|1[0-2])(?:0[1-9]|[12][0-9]|3[01])(?:$|[^0-9])`. Falla nombrando la línea y la
expresión o la fecha. Juntar los saltos es necesario: en v0.1.2 «la memoria de ⏎ consultas» (líneas 73-74) escapaba a la
comparación línea a línea. Un test unitario de la extracción fija sus casos (código en línea, bloque, región, salto de
párrafo, frontmatter). Alternativa rechazada: un objetivo de `make` con `grep` (no vería los saltos de línea ni
reutilizaría la comparación del juez).

### D5 · `umbrales` en el informe (FR-001 a FR-008)

Tipo `Umbral` con los campos del contrato del ADR 0029 (`total` como puntero con `omitzero`, V11). Se construyen sin
ningún caso por skill:

- con lista de expresiones: uno `expresiones_prohibidas:<modelo>` por modelo del job, con `medida` y `total` del recuento
  por modelo (`recontarExpresiones`, ya sin las sesiones sin medir), `comparacion` `"<="`, `umbral` `0.05` (constante
  del paquete) y `decide` verdadero solo para el modelo que decide (FR-002, FR-004);
- con objetivo de duración: `duracion_de_las_sesiones`, `medida` los segundos, sin `total`, `"<="` y el objetivo, con
  `decide` verdadero (FR-051).

`cumple` es la comparación en `float64` (`medida/total`, 0 si `total` es 0). Cada umbral que decide y no se cumple da un
motivo al final de los de siempre; el veredicto sigue siendo la presencia de motivos. `informe.md` gana «## Umbrales»
detrás de «## Expresiones prohibidas por modelo». El objetivo de duración vive en la definición del job, junto a la
concurrencia de la que depende (D12), y llega al informe por bandera; el 5 % es de la lista, no del job.

Alternativas rechazadas: **un umbral por skill escrito en Go** (un caso especial por skill); **el objetivo en un fichero
de `evals/<skill>/`** (un fichero y un esquema más para un número que depende de la concurrencia); **derivar el
veredicto de `cumple` en `informe.sh`** (fuera de alcance y fuera del producto: ADR 0029, opción 1).

### D6 · Un límite no es una eval fallida (FR-033, FR-040 a FR-044)

`LeerSesion` lee además cada mensaje `system` con `subtype` `api_retry` (`attempt`, `max_retries`, `error`; uno sin esos
campos hace la sesión ilegible, como cualquier mensaje sin su forma), si el último mensaje es uno de ellos, y el texto
del último mensaje si es un `result` con `is_error`, sea cual sea el código de la sesión (V2, V3, V18). Con eso, sin
modelo y en este orden:

- **(a) mensaje del límite de uso**: el último mensaje es `result` con `is_error` y su texto empieza por
  `You've hit your`, `You've reached your` o `You're out of` —los principios de la lista propia de Claude Code para la
  cuenta de una suscripción (V3); los de organización o de puesto no llegan con `claude setup-token`—;
- **(b) reintentos agotados**: la sesión no terminó y su último `api_retry` es `rate_limit` con `attempt ≥ max_retries`;
- **(c) cortada durante reintentos**: el tope la cortó (124 o 137) y su último mensaje es un `api_retry` `rate_limit`.

Cualquiera de las tres deja la sesión **sin medir**, con el motivo `sin medir por límite de uso: <cuál>` (el de (a) con
el texto de Claude Code), fuera de la medida y del total de las expresiones, y su serie sin medir. Los reintentos por
`rate_limit` se cuentan por sesión y en total. Con alguna sesión sin medir, el veredicto es `fallo` con un solo motivo
que empieza por el prefijo fijo `de la ejecución, no de la skill: ` —el mismo que el de la duración—, que se distingue
sin modelo de los de la skill (FR-043). Tras (a) el repartidor no abre ninguna sesión más (D7); las que no abrió se
publican sin medir con «sin abrir tras el límite de uso».

El texto de ese `result` pasa también al motivo de la sesión sin terminar, **sea cual sea el código**: con 0, el motivo
`result con is_error` pasa a `result con is_error: <texto>`; con un código que no es del tope, `código <n>` pasa a
`código <n>: result con is_error: <texto>`; los del tope no cambian. Es lo que hace ver en el informe del job y en la
salida del sondeo una credencial caducada o revocada, o un modelo que no existe (FR-063, FR-065): con un token que no
sirve, `claude -p` sale con 1 y el `result` dice «Failed to authenticate. API Error: 401 OAuth access token is invalid.»
(V18), y el motivo de hoy, `código 1`, no dice la causa. Alternativa rechazada: añadir el texto solo con código 0 (el
caso medido de la credencial, con código 1, no lo llevaría nunca).

Alternativas rechazadas: **clasificar por `api_error_status` 429** (no distingue el límite de la cuenta del límite
temporal, que sí se reintenta y dice «not your usage limit», V3); **reintentar la sesión** o **esperar a que se reponga**
(fuera de alcance; FR-037); **inferir las no abiertas de las que faltan del plan** (el repartidor las conoce y las pasa:
nada que adivinar).

### D7 · El repartidor de sesiones, en Go y compartido por el job y el sondeo (FR-030 a FR-032, FR-036, FR-037, FR-044, FR-061)

`internal/evals` gana un repartidor: abre las sesiones del plan en su orden, como mucho `Concurrencia` a la vez; cada
una se prepara justo antes con `PrepararSesion` (lo que ejecutaba `TestPrepararSesion`: las evals de la skill y
`UnionDeGrabaciones()`), se abre con el guion de la sesión (D8) y, al terminar, se lee con `LeerSesion`; tras una de
tipo (a), no abre más y espera a las abiertas. Mide la duración (D14). Un error que impide abrir una sesión —de
preparación o de E/S— cierra el reparto, cierra las abiertas con el tope (D9) y vuelve con el error: el guion sale con
1, como hoy con una falta en la preparación. Una sesión que `LeerSesion` no puede leer no es ese error: el reparto
sigue y el informe la juzga como ilegible (T007; `gates/supuestos.md`). Lo que recibe el repartidor no es un contexto
sino un canal de interrupción, el `Done()` del contexto que las entradas cancelan con `SIGINT` y `SIGTERM` (T006;
`gates/supuestos.md`).

Alternativas rechazadas: **bash con `xargs -P` o `wait -n`** (no hay `wait -n` en bash 3.2, S5; la regla de parar tras
(a) necesita leer el transcript con el mismo código que el informe; y no se podría probar en `make ci` en macOS sin el
tope de coreutils, V12); **un `go test -run TestPrepararSesion` por sesión** desde el repartidor (un proceso `go test`
por sesión para llamar a una función que el repartidor ya tiene; `TestPrepararSesion` y `TestPlanDeSesiones` se retiran
porque ya nadie los ejecuta, y su lógica —`PlanDeEvals` y `PrepararSesion`— se queda); **`errgroup`**
(`golang.org/x/sync` no está entre las dependencias fijadas: canal como semáforo y `sync.WaitGroup`).

### D8 · El guion de la sesión, `scripts/evals-sesion.sh` (FR-031)

La orden de Claude Code se queda **igual que hoy**, carácter a carácter, en un guion de bash que el repartidor ejecuta
sin argumentos en `trabajo/` de la sesión: lee `../pregunta.txt` y `../modelo.txt`, y ejecuta
`claude -p "<pregunta>" --model <modelo> --output-format stream-json --verbose --max-turns 30 --no-session-persistence
--setting-sources user --settings '{"sandbox":{"enabled":false}}' --permission-mode bypassPermissions --disallowedTools
WebFetch WebSearch`, con `exec strace -ff -e trace=execve,connect,clone,clone3,fork,vfork -s 131072 -o ../traza/t --`
delante si `KITLEGAL_EVALS_TRAZA=si` (el job). Así `exec.CommandContext` recibe el guion como ejecutable y ningún
argumento, y G204 no salta (V8) si la orden la construye una función que recibe la ruta del guion como parámetro
(`exec.CommandContext(ctx, guion)`); lo mismo para `kitlegal skills install` en el sondeo, con la ruta del binario como
parámetro, y `go install -trimpath ./cmd/kitlegal`, todo constantes. Alternativas rechazadas: **la orden en Go** (la pregunta y el modelo son
variables: G204 exige constantes; pasarlos por la entrada estándar y `ANTHROPIC_MODEL` cambiaría la orden de la sesión,
FR-031); **`//nolint:gosec`** (constitución, «sin atajos»).

### D9 · El tope, en Go (FR-031, FR-062)

El repartidor abre el guion en su propio grupo de procesos (`SysProcAttr.Setpgid`); a los 240 s envía `TERM` al grupo y,
si no ha terminado 10 s después, `KILL`; el código que escribe en `codigo-de-la-sesion` es el de `timeout
--kill-after=10s 240s`: 124 si bastó `TERM`, 137 si hizo falta `KILL`, y el del proceso si terminó antes. Es lo que ya
lee `LeerSesion`. Vale igual en Linux y en macOS, donde no hay `timeout` (V12), y se prueba en `make ci` con un tope
corto. Las entradas del job y del sondeo cancelan el contexto con `SIGINT` y `SIGTERM`
(`signal.NotifyContext`), y el repartidor cierra las abiertas con la misma secuencia: una persona que interrumpe el
sondeo no deja sesiones vivas (FR-064). Alternativa rechazada: seguir con `timeout` en el job y otro tope en el sondeo
(dos caminos, uno sin probar en `make ci`).

### D10 · El estado de Claude Code, por sesión (FR-032, FR-063)

Cada sesión tiene su `claude/` (`CLAUDE_CONFIG_DIR`, V4) con `skills/<entrada>` como enlace simbólico absoluto a cada
entrada del directorio de skills instaladas —`$HOME/.claude/skills` en el job, tal como lo deja `make install`, y el del
`HOME` temporal del sondeo—, y su `tmp/` (0700) como `TMPDIR` y `CLAUDE_CODE_TMPDIR` (V4). Con `--setting-sources user`,
Claude Code carga solo las skills, el `settings.json` y el `CLAUDE.md` de ese directorio (V4, V5), y escribe ahí su
`.claude.json` y el resto de su estado (S2): dos sesiones no escriben en el mismo sitio, y ninguna ve la configuración
de quien lanza el sondeo. El llavero: con `CLAUDE_CONFIG_DIR`, el servicio lleva el sufijo del directorio y no es el de
la persona (V5), y la credencial es `CLAUDE_CODE_OAUTH_TOKEN`, que no lee el almacén. Alternativas rechazadas: **`--bare`**
(ignora `CLAUDE_CODE_OAUTH_TOKEN`, V5); **copiar las skills** (no sería la skill «tal como la deja `make install`»);
**`XDG_*_HOME` por sesión** (ninguna evidencia de que Claude Code escriba ahí en `-p`; S2).

### D11 · Una tanda por commit: el segundo disparo espera (FR-034)

`concurrency` en el trabajo `evals`, con `group: evals-${{ github.event.pull_request.head.sha || github.sha }}-${{
matrix.skill }}` y `cancel-in-progress: false` (V6). En el cierre, el disparo de apertura (`publicar`) y el de la
etiqueta (`medir`) caen en el mismo grupo por skill: el segundo queda `pending` hasta que el primero termina y después
abre su tanda, que ya no es simultánea. Ninguno se cancela ni se salta: `medir` espera mientras haya alguna `pending` y
lee informes de `pass` o `fail` (V15), y los dos miden el mismo commit. No depende de cómo agrupe `gh` las comprobaciones
repetidas (S3): si el cierre solo ve la segunda, espera a que termine; si ve las dos, lee las dos. La regla de V6
—un tercer disparo cancela al que espera— no tiene vía en el run: los únicos disparos sobre un commit son la apertura
(`publicar`, una vez) y la etiqueta (`medir`, una vez por ronda), y `medir` no vuelve hasta que termina lo que ve
—al menos la primera tanda—, así que el disparo de la ronda siguiente sobre el mismo commit llega con la segunda ya en
marcha o terminada: puede esperar detrás de ella, pero no hay ninguna pendiente que cancelar. Sobre el commit del
reparador, el grupo es otro.

Alternativas rechazadas: **saltar el segundo** (sin llamar a la API no se sabe si hay otro en curso; con un paso previo
que la llame, la comprobación saltada o vacía podría ocupar el nombre que lee el cierre, S3); **`cancel-in-progress:
true`** (cancela la tanda que corre: una comprobación `cancel`, roja para el cierre); **`concurrency` de nivel de
flujo** (no admite `matrix`, V6, y detendría también a la otra skill); **`queue: max`** (V6 lo describe, pero no se pudo
comprobar que la plataforma lo acepte). Coste asumido: el cierre mide dos veces el mismo commit, una tras otra, como
permite FR-034 («el plan elige si el segundo espera o se salta») y el caso límite del spec.

### D12 · Concurrencia y objetivo en la matriz, con el nombre del trabajo fijo (FR-030, FR-034, FR-051)

`strategy.matrix` conserva `skill: [boe-legislacion, legal-core]` y gana `include` con `concurrencia` (4 y 1) y
`objetivo_de_duracion` (900 y 0) por skill, que se añaden a su combinación sin crear otras (V6); `env` pasa
`CONCURRENCIA_DE_EVALS: ${{ matrix.concurrencia }}` y `OBJETIVO_DE_DURACION_DE_EVALS: ${{ matrix.objetivo_de_duracion
}}`. El trabajo lleva `name: evals (${{ matrix.skill }})` (admite `matrix`, V6): con claves nuevas en la matriz, el
nombre por omisión podría llevar sus valores, y el cierre reconoce el informe de cada skill por `evals (<skill>)` (V15).
El filtro del trabajo `cambios` gana `scripts/evals-sesion.sh`, que ahora forma parte de lo que las evals miden (D8).

### D13 · El tope de cada trabajo cubre su peor caso (FR-035)

`peor caso = fuera + ⌈N / C⌉ × (P + 240 s + 10 s)`, con lo medido en el cierre de H7.2 (`docs/USO.md`, 2026-09-29):

- `fuera` = **485 s**: lo más que tardó un trabajo fuera de sus sesiones (42 min 33 s − 34 min 28 s = 8 min 5 s), que
  incluye la preparación del runner (hasta 8 min), las comprobaciones previas, el plan y el informe;
- `P` = **22 s**: lo que tardó de media una sesión entera con su preparación (34 min 28 s / 93 = 22,2 s), cota de la
  preparación sola;
- `N`, las sesiones del plan con la prueba de red, calculadas con `PlanDeEvals` sobre las evals del repositorio y los
  modelos y repeticiones de la definición; `C`, la concurrencia de su skill.

`boe-legislacion`: 94 sesiones, 4 a la vez: 485 + 24 × 272 = 7 013 s (116,9 min). `legal-core` (la prueba de red contada
aunque no la lleve): 19 sesiones, 1 a la vez: 485 + 19 × 272 = 5 653 s (94,2 min). `timeout-minutes: 120` los cubre; la
comprobación de la definición (D15) lo recalcula y falla si una eval nueva lo deja corto. No es el control del objetivo
(FR-052).

### D14 · La duración de las sesiones (FR-050)

Los segundos desde que el repartidor empieza a preparar la primera sesión hasta que termina la última, **redondeados
hacia arriba** a un entero: 900,4 s se publican como 901 y no cumplen. Sin la preparación del runner, las comprobaciones
previas ni el informe. Con 4 a la vez se esperan unos 94 × 22 / 4 ≈ 520 s.

### D15 · La definición del job, comprobada en `make ci` (FR-094)

`TestDefinicionDelJob` lee `.github/workflows/evals.yml` con `go.yaml.in/yaml/v3` (V16) y falla si el trabajo `evals`
no tiene el `concurrency` de D11 con `cancel-in-progress: false` o si hay un `concurrency` de flujo; si su `name` no es
`evals (${{ matrix.skill }})`; si `include` no fija 4 y 1, 900 y 0, o `env` no los pasa; o si `timeout-minutes` no cubre
el peor caso de D13 de cada skill. El mismo lector da al sondeo la concurrencia de su skill (D16).

### D16 · El sondeo (FR-060 a FR-068)

- **Orden**: `make evals-sondeo SKILL=<skill> EVALS=<nn,nn,…> MODELO=<id> REPETICIONES=<n> [CONCURRENCIA=<n>]`, en `make
  help`. `SKILL`, `EVALS`, `MODELO` y `REPETICIONES` no tienen valor por defecto: cada uno cambia lo que se mide y un
  valor supuesto lo escondería; `CONCURRENCIA`, la del job para esa skill (FR-060), leída de la definición (D15).
- **Guion** `scripts/evals-sondeo.sh`: crea el directorio temporal (`mktemp -d`), lo borra con `trap … EXIT` también si
  termina con otro código (S6), ejecuta `TestSondeo` con su salida en un registro dentro de él y, si pasa, imprime
  `salida.txt`; si no, el registro por la salida de error y sale con 1. Imprimir desde el guion es obligado: en Go no
  hay escritor de la salida estándar fuera de `cmd/` (V9) y `go test` no enseña la salida de un test que pasa (V10).
- **`TestSondeo`** (etiqueta `evals`, `-timeout 0`): comprueba los argumentos nombrando el de `make` (FR-067) y
  `CLAUDE_CODE_OAUTH_TOKEN` no vacío (FR-063) antes de nada; construye el binario con `go install -trimpath
  ./cmd/kitlegal` y `GOBIN` en el temporal, e instala las skills con ese binario, `skills install -g --host claude` y
  `HOME` en el temporal (como `make install`, FR-062); abre el plan (las evals pedidas, ese modelo, esas repeticiones,
  sin modelos informativos ni prueba de red) con el repartidor sin traza; juzga cada sesión (D17); y escribe
  `salida.txt`.
- **Entorno de sus sesiones**: `PATH` (con `bin/` del temporal delante), `LANG`, `LC_ALL`, `LC_CTYPE`, `LC_MESSAGES`, `TERM`, `USER`, `LOGNAME`,
  `SHELL`, `TZ` y `CLAUDE_CODE_OAUTH_TOKEN` de quien lo lanza; `HOME` en el temporal; y lo de cada sesión (D10, proxy
  cerrado, `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1`, `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=0`). Ninguna otra variable:
  `ANTHROPIC_API_KEY` y `ANTHROPIC_AUTH_TOKEN` tendrían prioridad sobre la suscripción (V5). El job conserva su entorno
  entero, como hoy.
- **Salida** ([contracts/sondeo.md](./contracts/sondeo.md) §4): la primera línea dice que no es un veredicto; después lo
  que no comprueba, la tasa de cada serie, el recuento con el 5 % de referencia, las sesiones sin medir y las que no
  terminaron por otra causa, con su motivo. Nada por sesión de las expresiones. Sale con 0 si ha podido abrir y juzgar.

Alternativas rechazadas: **un `main` en `cmd/`** (un segundo ejecutable en el repositorio, principio VI, y una excepción
de lint para él); **`t.Output()` con `-v`** (la salida empezaría por `=== RUN`); **la comprobación de la credencial en
el guion** (no se probaría en `make ci` sin él; en Go se prueba con el entorno como dato); **validar la credencial
contra el servicio** (FR-063 lo excluye).

### D17 · El juicio del sondeo, sin traza (FR-061)

El sondeo juzga con `Juzgar` sobre una copia de la eval **sin `comandos` ni `prohibidos`** y una sesión sin
invocaciones: `Juzgar` deja entonces vacíos los comandos, los prohibidos, las invocaciones, lo fuera de lo grabado, las
otras fallidas y las llegadas a la red, y su juicio es «el de antes» en todo lo demás. Después, como el job:
`exigirElModeloPedido` y la clasificación de D6. Series y recuento, con `repartirEnSeries` y `recontarExpresiones` del
informe. Sin un juez propio.

### D18 · Los sustitutos de `claude` y `strace` en los tests (FR-068, FR-094, FR-096)

Guiones POSIX escritos por los tests en `t.TempDir()` (constantes del test, sin `testdata/`): `claude` elige su
transcript por el nombre de la sesión (`basename` de `..`) en un directorio que le da el test, anota lo que ve
—directorio de trabajo, `CLAUDE_CONFIG_DIR`, `TMPDIR`, `CLAUDE_CODE_TMPDIR`, `KITLEGAL_CACHE_DIR`, `HOME`, el
`kitlegal` que resuelve y si ve `ANTHROPIC_API_KEY`—, marca su llegada con un `mkdir` en un directorio común, marca
que está abierta con otro en un segundo directorio, que retira al terminar, y anota cuántas abiertas cuenta al llegar
(el número de sesiones abiertas a la vez), duerme lo que le digan (0 s por omisión), y sale con el código que le digan; `strace` escribe
una traza fija de un solo proceso en su `-o` y ejecuta lo que va tras `--`; y `go`, en el test del guion del sondeo,
escribe `salida.txt` en el `-temporal` que recibe y sale con el código que le digan. Se escriben con
`os.OpenRoot(t.TempDir())` y `root.WriteFile(…, 0o755)` (V8b). Nadie abre una sesión con modelo en `make ci`. Alternativas rechazadas: el binario de test como sustituto (necesita `os.Exit` y `os.Stdout`, V9); un programa en
`testdata/` compilado por el test (más código para lo mismo).

### D19 · Las entradas del job (FR-030, FR-050)

`scripts/evals.sh` conserva las comprobaciones previas, exige y valida `CONCURRENCIA_DE_EVALS` como las repeticiones,
deja de exigir `timeout` y sustituye el bucle y las dos órdenes de Go por una: `TestEjecucionDelJob` (etiqueta `evals`,
`-timeout 0`) planifica, reparte, juzga y escribe el informe con la duración, y falla con `fallo`. El guion sigue
comprobando que el informe se escribió y lo imprime entre sus marcas. `TestInformeDelJob`, `TestPlanDeSesiones`,
`TestPrepararSesion` y `plan.tsv` se retiran: nadie más los usa (`git grep`). Alternativa rechazada: dos órdenes,
sesiones e informe, con la duración en un fichero entre medias (un fichero y dos banderas más).

### D20 · Documentación que el hito deja falsa (FR-018, FR-097)

`CHANGELOG.md` (*Unreleased*): v0.1.3, la tercera familia, los umbrales que deciden, el job en paralelo con límites y
duración, y el sondeo. `CONTRIBUTING.md`, solo lo que deja de ser cierto: «Job de evals» (sesiones a la vez, las
variables nuevas, `umbrales`, sesiones sin medir, duración, reintentos, una tanda por commit, el tope) y la lista de
expresiones (tres familias, calibrado sobre H7.1 y H7.2, la prosa). El sondeo **no** se documenta fuera de `make help`,
del quickstart y de `CHANGELOG.md` (spec, «Fuera de alcance»). `internal/evals/doc.go`, lo que hace el paquete.

### D21 · Aceptación e2e: no aplica

El hito no cambia el binario (fuera de alcance; `TestDependenciasDelBinario` impide que `cmd/kitlegal` enlace
`internal/evals`): no hay comportamiento de `kitlegal` que un guion `testscript` describa. La aceptación es la del job
(SC-001) y la del quickstart (SC-002), y cada pieza entra con sus tests en `make ci`.

### D22 · Datos externos: ninguno

Ni fuente ni grabación nuevas; ningún manifiesto `grabaciones.json` ni test `TestGrabar*` cambia; `grabar_datos` no
tiene nada que grabar. La calibración usa los informes versionados de H7.1 y H7.2, que no se editan (FR-080).
