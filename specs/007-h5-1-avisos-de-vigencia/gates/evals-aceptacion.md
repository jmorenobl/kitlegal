# Evidencia: ejecución de aceptación de H5.1

Registro de la tarea T014 (quickstart §11; contrato de la ejecución de aceptación; data-model §10). Tomado el
2026-09-16 sobre la propuesta de cambio [#34](https://github.com/jmorenobl/kitlegal/pull/34), abierta hacia `main` con
el título `feat(H5.1): Avisos de vigencia en las evals` y el cuerpo de `gates/pr-h5.1.md`.

> **Repetición pendiente** (anotado por la corrección de la revisión final, ronda 1, 2026-09-16). La ejecución de
> apertura de abajo es la última registrada, pero **ya no cubre la cabeza**: `7ae7a05` cambia `CHANGELOG.md`,
> `CONTRIBUTING.md`, `README.md`, `internal/evals/conjunto.go` e `internal/evals/formato.go`, fuera de
> `specs/007-h5-1-avisos-de-vigencia/`. Por el contrato de la ejecución de aceptación §6 se repite por etiqueta
> (quickstart §11.6) tras la revisión aprobada y el push, y su evidencia se añade aquí como sección nueva y vigente
> (`gates/revision-pendiente.md`).

## Ejecución de apertura (vigente hasta `7ae7a05`)

| Dato | Valor |
|---|---|
| Ruta de identificación | primera ejecución de `evals` de la rama (apertura): evento `pull_request`, `workflowName` `evals`, la más antigua por `createdAt` (quickstart §11.3) |
| Enlace | https://github.com/jmorenobl/kitlegal/actions/runs/35148840549 |
| `databaseId` | `35148840549` |
| `headSha` | `8272bc8c9c1d27000dec2b95fdd0bedc542119ae` (la cabeza de la rama al abrir la propuesta; `headRefOid` de #34) |
| Conclusión | `success` (`gh run watch --exit-status` → `código 0`); jobs `cambios` `success` y `evals` `success` |
| Duración del paso «Ejecutar las evals» | 36 min 6 s: primera línea `2026-09-16T20:53:39.5795121Z`, última `2026-09-16T21:29:45.8534254Z` |

### Supuestos de plataforma (research §S)

- **S1**: abrir la propuesta creó la ejecución `35148840549` (`workflowName` `evals`, evento `pull_request`, creada
  `2026-09-16T20:49:16Z`), la primera y única de `evals` de la rama; su job `cambios` terminó en `success` y el job
  `evals` se ejecutó.
- **S2**: su `headSha` es la cabeza de la rama al abrir la propuesta y coincide con el `commit` del informe (segunda
  salida de §11.5).
- **S3**: el registro de `gh run view --log` llevó las cuatro marcas de `scripts/evals.sh` y la extracción de §11.4
  funcionó sin cambios.
- **S4**: 90 sesiones (54 con `claude-sonnet-5`, 36 con `claude-haiku-4-5-20251001`) en 36 min 6 s, unos 24 s por
  sesión, dentro de `timeout-minutes: 120`.

### Lectura del informe (quickstart §11.4)

```text
{"commit":"8272bc8c9c1d27000dec2b95fdd0bedc542119ae","veredicto":"aprobado","modelo_que_decide":"claude-sonnet-5","repeticiones":3,"umbral":2}
paso Ejecutar las evals, primera y última línea:
2026-09-16T20:53:39.5795121Z
2026-09-16T21:29:45.8534254Z
```

### Resumen del informe: programa de `jq` (quickstart §11.5, primera orden; código 0)

```json
{
  "commit": "8272bc8c9c1d27000dec2b95fdd0bedc542119ae",
  "veredicto": "aprobado",
  "red": [],
  "modelo_que_decide": "claude-sonnet-5",
  "tasa": [
    {
      "planificada": true,
      "decide": false,
      "sesiones": 3,
      "pasan": 3,
      "pasa": true
    }
  ],
  "sesiones": [
    {
      "sesion": "18-lrjpac-norma-derogada-claude-sonnet-5-01",
      "pasa": true,
      "avisos_encontrados": [
        "derogada",
        "vigencia-agotada"
      ],
      "avisos_ausentes": [],
      "motivos": []
    },
    {
      "sesion": "18-lrjpac-norma-derogada-claude-sonnet-5-02",
      "pasa": true,
      "avisos_encontrados": [
        "derogada",
        "vigencia-agotada"
      ],
      "avisos_ausentes": [],
      "motivos": []
    },
    {
      "sesion": "18-lrjpac-norma-derogada-claude-sonnet-5-03",
      "pasa": true,
      "avisos_encontrados": [
        "derogada",
        "vigencia-agotada"
      ],
      "avisos_ausentes": [],
      "motivos": []
    }
  ]
}
true
```

### Commit y cabeza (quickstart §11.5, segunda orden; código 0)

```text
commit evaluado: 8272bc8c9c1d27000dec2b95fdd0bedc542119ae (el del informe)
ficheros cambiados entre el commit evaluado y la cabeza:

todos bajo specs/007-h5-1-avisos-de-vigencia/
```

### Repetición por etiqueta (quickstart §11.6)

No aplicaba al registrarla T014: ningún commit posterior a la ejecución cambiaba nada fuera de
`specs/007-h5-1-avisos-de-vigencia/` (la lista de ficheros cambiados entre el commit evaluado y la cabeza estaba vacía).
La etiqueta `evals` no se puso. Aplica desde `7ae7a05` (aviso del principio).

### Estados de la propuesta de cambio

`gh pr checks 007-h5-1-avisos-de-vigencia` con todo en `pass`: `ci` (4 min 49 s), `cambios`, `evals`,
`codecov/project` (96,42 %, objetivo 70 %), `codecov/project/internal/core` (90,36 %, objetivo 85 %),
`codecov/project/internal/cli` (98,09 %, objetivo 90 %) y `codecov/patch` (100,00 % del diff, objetivo 96,38 %); los
estados de Codecov traen cifra, no un verde vacío.
