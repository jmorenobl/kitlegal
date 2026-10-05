# Contrato: la definición del job, sus guiones y sus topes

Lo que cambia en `.github/workflows/evals.yml`, en `scripts/` y en el `Makefile` (FR-050, FR-090 a FR-093, FR-112;
SC-012). Lo comprueba `TestDefinicionDelJob` en `make ci`.

## 1. Lo que se fija (FR-090, FR-091)

En `env` del trabajo `evals`, junto a `MODELO_DE_EVALS` y `VERSION_DE_CLAUDE_CODE`:

```yaml
      MODELO_DEL_JUEZ: claude-opus-5-5
      VERSION_DE_CLAUDE_CODE_DEL_JUEZ: 2.1.289
```

`VERSION_DE_CLAUDE_CODE` sigue siendo la de las sesiones y no cambia: subir una no toca la otra.

## 2. El trabajo `evals`: dos Claude Code

El paso que instala Claude Code instala además el del juez en un prefijo aparte y deja su ruta a los pasos siguientes:

```bash
npm install -g "@anthropic-ai/claude-code@${VERSION_DE_CLAUDE_CODE}"
claude --version
npm install --prefix "$RUNNER_TEMP/claude-del-juez" "@anthropic-ai/claude-code@${VERSION_DE_CLAUDE_CODE_DEL_JUEZ}"
"$RUNNER_TEMP/claude-del-juez/node_modules/.bin/claude" --version
echo "CLAUDE_DEL_JUEZ=$RUNNER_TEMP/claude-del-juez/node_modules/.bin/claude" >> "$GITHUB_ENV"
```

- Que `npm install --prefix <dir>` deja el ejecutable en `<dir>/node_modules/.bin/` es un supuesto sin verificar
  (research S1): si no, el paso falla en su cuarta línea, antes de abrir ninguna sesión.
- El paso «Retirar Python del runner» añade `$CLAUDE_DEL_JUEZ` a lo que el job usa, para fallar si se lo lleva.
- `scripts/evals.sh`, con una skill que tiene `evals/<skill>/juez/`, exige `MODELO_DEL_JUEZ`,
  `VERSION_DE_CLAUDE_CODE_DEL_JUEZ` y un `CLAUDE_DEL_JUEZ` ejecutable antes de la primera sesión, y los pasa a
  `TestEjecucionDelJob` con `-modelo-del-juez`, `-version-del-juez` y `-claude-del-juez`. Con una skill sin juez no
  los mira.
- `TestEjecucionDelJob` llama a `ejecutarElJob` (`ejecucion.go`): comprueba la medida, y solo si corresponde y se
  cumple abre las sesiones por tandas, como hoy, y escribe el informe con el votante de `scripts/evals-voto.sh`
  ([juez-y-voto.md](./juez-y-voto.md) §4), cuyo `PATH` lleva delante el directorio de `CLAUDE_DEL_JUEZ`.

## 3. La ejecución de la medida (FR-050)

Dos formas de lanzarla, y ninguna más:

| Forma | En `evals.yml` |
|---|---|
| La etiqueta `evals-medir-juez` en una propuesta de cambio | `on.pull_request.types` ya incluye `labeled` |
| El flujo lanzado a mano con su entrada | `on.workflow_dispatch.inputs.medir_al_juez`, booleana, `false` por omisión |

Un trabajo nuevo, `medida`:

```yaml
  medida:
    name: medida del juez (${{ matrix.skill }})
    if: github.event.label.name == 'evals-medir-juez' || inputs.medir_al_juez == true
    strategy:
      fail-fast: false
      matrix:
        skill: [boe-legislacion]
        include:
          - skill: boe-legislacion
            concurrencia: 4
    runs-on: ubuntu-24.04
    timeout-minutes: 269
    permissions:
      contents: read
    env:
      MODELO_DEL_JUEZ: claude-opus-5-5
      VERSION_DE_CLAUDE_CODE_DEL_JUEZ: 2.1.289
      CONCURRENCIA_DE_EVALS: ${{ matrix.concurrencia }}
      SKILL_EVALUADA: ${{ matrix.skill }}
      COMMIT_EVALUADO: ${{ github.event.pull_request.head.sha || github.sha }}
```

Sus pasos: el código del commit evaluado, Go, el Claude Code del juez (las tres últimas líneas de §2) y `make
evals-medir-juez SKILL="$SKILL_EVALUADA"` con `CLAUDE_CODE_OAUTH_TOKEN`. No depende de `tanda`, no instala `strace` ni
las skills, y no retira Python: no abre sesiones de evals.

Para que el despacho con la entrada no abra además las sesiones de evals, la condición de `tanda` pasa de
`github.event_name == 'workflow_dispatch'` a `(github.event_name == 'workflow_dispatch' && inputs.medir_al_juez !=
true)`. La etiqueta `evals-medir-juez` ya no entra en `tanda`: no es ninguna de las suyas, y `cambios` no corre con
`labeled`.

| Disparo | `tanda` y `evals` | `medida` |
|---|---|---|
| Abrir o reabrir una propuesta | como hoy | no |
| Etiqueta `evals` o `evals-prueba-de-red` | sí | no |
| Etiqueta `evals-medir-juez` | no | sí |
| Despacho sin la entrada | sí | no |
| Despacho con `medir_al_juez` | no | sí |

`make evals-medir-juez SKILL=<skill>` ejecuta `scripts/evals-medir-juez.sh <skill>`: exige las variables de la
tabla de arriba, `CLAUDE_DEL_JUEZ` y `CLAUDE_CODE_OAUTH_TOKEN`; ejecuta `TestMedidaDelJuez`, sin límite de tiempo de
`go test`; imprime `medida.json` entre sus marcas si se escribió
([medida-del-juez.md](./medida-del-juez.md) §7); y sale con el código del test. No forma parte de `make ci`.

## 4. Los topes (FR-092)

`TestDefinicionDelJob` recalcula los dos peores casos con las evals y los casos del repositorio y falla si un
`timeout-minutes` no lo cubre.

| Término | Valor | De dónde |
|---|---|---|
| Sesiones de `boe-legislacion` | 14 357 s | El de hoy: `485 s + (⌈97/4⌉ + ⌈96/4⌉ + ⌈6/4⌉) × 272 s` (`definicion.go`) |
| Instalación del Claude Code del juez | 60 s | Cota sin medir (research S3) |
| Un voto, como mucho | 40 s | 35 s de tope y 5 s de margen ([juez-y-voto.md](./juez-y-voto.md) §4) |
| Votos por respuesta, como mucho | 6 | Tres votos, cada uno con su repetición por nulo |
| Respuestas a la vez | 4 | La `concurrencia` de la skill |

- **`evals`**: `14 357 + 60 + (⌈54×6/4⌉ + ⌈54×6/4⌉ + ⌈3×6/4⌉) × 40 = 14 417 + 167 × 40 = 21 097 s`, 351,6 min:
  `timeout-minutes: 352` (hoy, 240). Las respuestas de cada grupo salen del plan: las sesiones del modelo que decide
  en las evals que activan la skill, por modo, y las de la eval sin binario ni servidor.
- **`medida`**: `485 + 60 + ⌈259×6/4⌉ × 40 = 545 + 389 × 40 = 16 105 s`, 268,4 min: `timeout-minutes: 269`.
- Los dos caben en las 6 horas de un trabajo de un runner de GitHub (research S4).
- Con una skill sin juez, el peor caso de su trabajo no tiene los términos del juez.

## 5. `TestDefinicionDelJob` (FR-112)

En `del-repositorio`, y en `sinteticas` con la definición cambiada en cada punto, que da su fallo:

| Comprueba | Falla si |
|---|---|
| `jobs.evals.env.MODELO_DEL_JUEZ` | falta; no tiene la forma de un id completo (`^claude-[a-z]+(-[0-9]+)+$`), como un alias; o es igual a `MODELO_DE_EVALS` |
| `jobs.evals.env.VERSION_DE_CLAUDE_CODE_DEL_JUEZ` | falta o no es `<n>.<n>.<n>`; o el paso de instalación no instala `@anthropic-ai/claude-code@${VERSION_DE_CLAUDE_CODE_DEL_JUEZ}` en su prefijo, o instala el de las sesiones con ella |
| `jobs.medida.env` | su modelo o su versión no son los de `jobs.evals.env` |
| `jobs.medida.if` | no es exactamente la condición de §3 |
| `jobs.medida.needs` | existe |
| `jobs.tanda.if` y `jobs.evals.if` | nombran `evals-medir-juez`, o `tanda` admite el despacho con `medir_al_juez` |
| `on.workflow_dispatch.inputs.medir_al_juez` | falta o su valor por omisión no es `false` |
| Los dos `timeout-minutes` | no cubren su peor caso (§4) |

`leerDefinicionDelJob` lee para ello el trabajo `medida`, las entradas del despacho y el `run` del paso de
instalación; el sondeo toma de ella el modelo del juez, como toma la concurrencia.

## 6. Lo que un run no hace (FR-045, FR-093)

Ninguna tarea ejecuta `make evals`, `make evals-sondeo`, `make evals-medir-juez`, `scripts/evals*.sh`,
`TestEjecucionDelJob`, `TestMedidaDelJuez` ni `TestSondeo`: abren sesiones con modelo. Los tests de `make ci` usan
votantes de salidas grabadas y el `claude` sustituto. El workflow solo pone la etiqueta `evals`.
