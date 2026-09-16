# Contrato: la eval de la norma derogada, su norma y su grabación

FR-050 a FR-057, FR-060; US5; SC-006, SC-009. Decisiones: research.md D11 a D14, D16. Entidades: data-model §7 a §9.
Amplía, sin cambiar nada más, `specs/006-h5-skill-boe-legislacion/contracts/evals-y-grabaciones.md` §2, §3.1 y §3.2 y
`normas-y-referencias.md`.

## 1. Arnés de grabación y premisa de la verificación de identificadores

Dos cambios en ficheros de test de `internal/evals`, en la misma tarea y **antes** de la del manifiesto:

- `grabacion_test.go` (`//go:build grabacion`), método `pedir` de `grabacionDeEvals`: siembra la consulta desde
  `UnionDeGrabaciones()` en lugar de `[]string{GrabacionesDeH4}`, y su mensaje de error dice «desde las grabaciones de H4
  y de H5». El comentario de `TestGrabarEvals`, punto 1, pasa a decir que solo se pide a la fuente lo que no está grabado
  en ninguno de los dos conjuntos. Nada más cambia en el arnés ni en `scripts/grabar-evals.sh`.
- `grabaciones_test.go`, `otraNormaDeLaBusqueda`: recibe el manifiesto y la posición de la entrada, y elige el primer
  resultado de su búsqueda grabada que no es ninguna norma de la tabla **y cuyo título no empieza por el
  `titulo_empieza_por` de ninguna entrada del manifiesto**; su comentario lo dice. `TestIdentificadoresDeLasNormas` la
  llama con esos argumentos. Mientras el manifiesto no tenga la entrada de §2 sigue eligiendo `BOE-A-1992-26318`, que
  ninguna entrada resuelve todavía; con ella, elige `BOE-A-2018-12131` (research V14).

Con los dos cambios, `make ci` sigue en verde sobre el árbol de `main` (la premisa elige otra norma que tampoco resuelve
ninguna entrada) y el arnés compila con el lint (`run.build-tags` incluye `grabacion`).

## 2. Manifiesto (tarea `[datos]`, con pausa)

`testdata/evals/grabaciones.json` gana, al final de `normas`, esta entrada en una línea, con la forma de las demás:

```json
    {"busqueda": "procedimiento administrativo común", "titulo_empieza_por": "Ley 30/1992,", "bloques": ["a42"], "para": "eval 18-lrjpac-norma-derogada: artículo 42 de la Ley 30/1992, norma derogada con los avisos derogada y vigencia-agotada (FR-050 de H5.1); identificador de data/normas.yaml (FR-056 de H5.1); índice, metadatos y bloque de la cita (FR-074 de H5); la búsqueda, los metadatos y el bloque ya están en las grabaciones de H4, y solo el índice es nuevo"}
```

La tarea solo escribe esa entrada. `LeerManifiesto` la acepta: su prefijo no es igual a otro ni empieza por otro, y
repetir la búsqueda de la primera entrada no es un defecto (contrato de H5 §3.1, punto 5).

## 3. Lo que graba la persona en la pausa

Con el arnés de §1, `scripts/grabar-evals.sh` sirve desde lo grabado toda consulta de toda entrada salvo una: el índice
de `BOE-A-1992-26318` (research V15, V16, V18). Antes de pedirlo, `httpx` pide `robots.txt` y lo graba encima del de H5
(V11, V13), que `httpx.Replay` no usa.

| Fichero bajo `testdata/evals/boe.legislacion-consolidada/` | Qué hace la persona |
|---|---|
| `GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-1992-26318_texto_indice.json` (nuevo) | lo revisa y lo confirma |
| `GET_https_www.boe.es_robots.txt.json` (reescrito) | lo restaura |
| cualquier otro nuevo o modificado | no debería existir: restaura todo, no confirma nada, rechaza la pausa y anota la salida |

## 4. Procedimiento de la pausa

Desde la raíz del repositorio, con la tarea del manifiesto ya confirmada por el workflow:

```bash
scripts/grabar-evals.sh
git status --porcelain -- testdata/
git restore -- testdata/evals/boe.legislacion-consolidada/GET_https_www.boe.es_robots.txt.json
git status --porcelain -- testdata/
jq -r '.respuesta.cuerpo' testdata/evals/boe.legislacion-consolidada/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-1992-26318_texto_indice.json | jq -c '[.data[].bloque[] | select(.id == "a42") | {id, titulo}]'
go test -count=1 -run '^(TestManifiestoDeGrabaciones|TestGrabacionesSinSolape|TestIdentificadoresDeLasNormas|TestEvalsDelRepositorio)$' ./internal/evals/
git add testdata/evals/boe.legislacion-consolidada/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-1992-26318_texto_indice.json
git commit -m "test(H5.1): índice grabado de BOE-A-1992-26318"
```

Esperado: la grabación termina en `ok` y su registro lleva la línea `entrada 11 ("Ley 30/1992,"): BOE-A-1992-26318 «Ley
30/1992, de 26 de noviembre, …»`; el primer `git status` muestra solo `?? …BOE-A-1992-26318_texto_indice.json` y
` M …GET_https_www.boe.es_robots.txt.json` (y nada del manifiesto, ya confirmado); el segundo, solo el índice; el `jq`
da `[{"id":"a42","titulo":"Artículo 42"}]`; los tests terminan en `ok`; el commit contiene solo el índice. Después, la
persona aprueba la pausa. Si la búsqueda no resuelve la norma o el índice no lleva `a42`, restaura, rechaza y lo anota
(supuesto S5 de research). La red y `KITLEGAL_RECORD` solo los usa la persona, aquí; ningún job ni tarea los usa.

## 5. `data/normas.yaml` y `references/normas.md`

Al final de `normas`:

```yaml
  BOE-A-1992-26318:
    titulo: "Ley 30/1992, de 26 de noviembre, de Régimen Jurídico de las Administraciones Públicas y del Procedimiento Administrativo Común."
    rango: Ley
    abreviatura: LRJPAC
    materias:
      - régimen jurídico de las administraciones públicas
      - procedimiento administrativo
```

`make skills-sync` regenera `skills/boe-legislacion/references/normas.md`, que gana la fila de la norma en el orden que
aplica la generación. Ni la tabla ni la referencia dicen que la norma esté derogada (FR-056).

## 6. La eval

`evals/boe-legislacion/18-lrjpac-norma-derogada.yaml`:

```yaml
# Eval informativa (ADR 0016): norma derogada y con la vigencia agotada. La pregunta nombra la norma y el artículo y no
# dice nada de su vigencia: la respuesta tiene que citar el bloque y trasladar los dos avisos que emite el binario, cada
# uno con su forma fija. Nace informativa: promoverla a decisoria se decide con los datos de varias ejecuciones.
pregunta: "¿Qué dice el artículo 42 de la Ley 30/1992 sobre la obligación de resolver?"
activa: true
informativa: true
comandos:
  - applet: boe
    norma: BOE-A-1992-26318
    bloque: a42
citas:
  - norma: BOE-A-1992-26318
    bloque: a42
avisos:
  - derogada
  - vigencia-agotada
```

Se confirma en un commit **anterior** al primero que cambia `skills/boe-legislacion/SKILL.md` (FR-052).

## 7. Controles

| Test | Qué fija con H5.1 |
|---|---|
| `TestManifiestoDeGrabaciones/repositorio` | el manifiesto con la entrada nueva se lee sin error (once entradas) |
| `TestGrabacionesSinSolape` | el índice nuevo no tiene el nombre de ninguna grabación de H4 |
| `TestIdentificadoresDeLasNormas` | `BOE-A-1992-26318` lo resuelve exactamente la entrada nueva, con su título; los cuatro subtests negativos siguen fallando donde deben (premisa de §1) |
| `TestNormasDelRepositorio`, `TestSkillsDelRepositorio` | la tabla contra su esquema y `references/normas.md` sin deriva |
| `TestEvalsDelRepositorio/formato` | 18 evals bien formadas |
| `TestEvalsDelRepositorio/conjunto` | tamaño (18 ≤ 20), positivas (10 que deciden), no activación, informativas (la 18 es informativa y activa), materias distintas y el resto, sin defectos |
| `TestEvalsDelRepositorio/normas-conocidas` | `BOE-A-1992-26318` está en `data/normas.yaml` |
| `TestEvalsDelRepositorio/grabado` | el bloque `a42`, el índice y los metadatos de `BOE-A-1992-26318` se preparan desde lo grabado y se sirven con `--offline` con código 0 (FR-057) |

Demostración: sin el índice grabado, `grabado` falla con
`18-lrjpac-norma-derogada.yaml: la norma BOE-A-1992-26318 sin indice: …` (research V15; quickstart §8).
