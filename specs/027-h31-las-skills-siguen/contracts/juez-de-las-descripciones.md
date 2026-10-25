# Contrato: el juez de las descripciones en el job de evals (FR-025 a FR-034)

Quién lo consume: quien revisa una propuesta que cambia una descripción, en el `informe.md` y el `informe.json` de cada
trabajo `evals (<skill>)`, y el informe final del workflow, que lee sus `umbrales` (ADR 0029). Cuántas veces: una por
ejecución del job que juzga descripciones, una entrada por trabajo. Corre solo en el job y en la medida del juez; nunca
en `make ci` ni en un paso del run (FR-033): sus puntos de entrada llevan `//go:build evals`, y los tests de `make ci`
lo ejercen con un votante simulado o con el `claude` sustituto de `sustitutos_test.go`.

## 1. El juez

`juezDeLasDescripciones(raiz)` lee de `evidencias/adr-0046/` (research D2) y devuelve un `*Juez` con:

- `Clases`: una, `{Nombre: "habla_en_primera_o_segunda_persona", Decide: true}`;
- `Rubrica`, `Esquema`, `Casos`, `Medida`: el contenido entero de `rubrica.md`, `esquema.json`, `casos.yaml` y
  `medida.json`.

El modelo y la versión de Claude Code de sus votos son los del juez de las evals, `MODELO_DEL_JUEZ` y
`VERSION_DE_CLAUDE_CODE_DEL_JUEZ` de `.github/workflows/evals.yml` (hoy `claude-opus-5-5` y 2.1.289).

## 2. La decisión de cada trabajo

`TestDecisionDelTrabajo` añade al fichero de la decisión (`$RUNNER_TEMP/decision-de-evals.json`, el bloque `origen`
sin `sesiones_abiertas`) un bloque `descripcion` cuando la ejecución tiene propuesta (`-propuesta` y `-base` no
vacías) y algún `skills/<s>/SKILL.md` cumple `cambiaLaDescripcion` entre la base y la cabeza (research D3). Sin esa
condición, no hay bloque y el trabajo no hace nada de lo de este contrato.

Juzgada en esta ejecución:

```json
"descripcion": {
  "skill": "legal-core",
  "huella": "5d0c…64 cifras hex…",
  "juzgada": true
}
```

Reutilizada de una medida de `main` (solo con reutilización; research D4):

```json
"descripcion": {
  "skill": "boe-legislacion",
  "huella": "a91e…",
  "juzgada": false,
  "procedencia": {"ejecucion": 18923456789, "trabajo": 53012345678, "commit": "73aab57…", "propuesta": 189,
                  "fecha": "2026-10-23T09:12:44Z"},
  "juicio": {"marcada": false, "frase": "", "motivo": "Todas las frases hablan de la skill en tercera persona."}
}
```

`procedencia` tiene la forma de la de un resultado reutilizado de H30. La huella (data-model §4) es el SHA-256 de:

```text
skill boe-legislacion
descripcion <sha256 de la description>
rubrica <sha256 de evidencias/adr-0046/rubrica.md>
esquema <sha256 de evidencias/adr-0046/esquema.json>
casos <sha256 de evidencias/adr-0046/casos.yaml>
medida <sha256 de evidencias/adr-0046/medida.json>
juez claude-opus-5-5 2.1.289
```

## 3. El trabajo

Con el bloque, `ejecutarElJob`, en este orden:

1. antes de abrir ninguna sesión —ni de las evals ni del juez—, `comprobarLaMedida(juezDeLasDescripciones, modelo,
   versión)`; con alguna línea, escribe el informe sin medir, con cada línea como motivo precedida de
   `juez de las descripciones: `, y no abre nada (FR-026);
2. las sesiones de las evals, como hoy;
3. si `juzgada`, los votos de la descripción de su skill (§4, §5); si no, el juicio del bloque, sin abrir sesiones;
4. el informe (§6).

`scripts/evals.sh` pasa las banderas del juez (`-modelo-del-juez`, `-version-del-juez`, `-claude-del-juez`) en todo
trabajo, no solo con `evals/<skill>/juez/`: el paso que instala el Claude Code del juez ya corre en todos.

## 4. El mensaje de cada voto (FR-027)

`mensajeDeLaDescripcion(skill, descripcion)`, el de `mensaje_de` de `evidencias/adr-0046/guiones/juez.py`, con la
descripción que da `LeerFrontmatter` sin los blancos de los extremos (`esBlanco`). Para la skill `legal-core` y la
descripción `Punto de partida…`:

```text
<descripcion skill="legal-core">
Punto de partida…
</descripcion>

Responde a la pregunta de la rúbrica sobre esta descripción.
```

sin salto de línea al final. Va por la entrada estándar de `scripts/evals-voto.sh`, que abre la sesión con la rúbrica
como instrucciones y el esquema como forma de la respuesta.

## 5. Los votos (FR-028, FR-029, FR-032)

La regla de la `votacion` del juez de las evals con una clase que decide (research D6):

| Voto | Qué pasa |
|---|---|
| «si» con una frase que `fraseEsta` encuentra en la descripción | cuenta como «si»; se pide el siguiente |
| «si» con una frase que no está | nulo; se pide otra vez, una sola; si vuelve a ser nulo, su «si» no cuenta |
| «no» | la descripción no queda marcada; no se piden más |
| cortado por el tope (35 s más 5 de margen) | se pide otra vez, una sola; cortado otra vez, sin juzgar |
| sesión con error, sin la forma del esquema, o cortada por un límite de uso | sin repetición: sin juzgar |

Marcada con tres «si» unánimes con su frase, pedidos por orden; como mucho seis votos.

## 6. El informe

`informe.json` gana el bloque `descripcion` y el umbral. Juzgada y marcada:

```json
"descripcion": {
  "skill": "legal-core",
  "clase": "habla_en_primera_o_segunda_persona",
  "huella": "5d0c…",
  "juzgada": true,
  "marcada": true,
  "frase": "Úsala cuando se pregunte hasta cuándo se puede recurrir",
  "motivo": "Da una orden a quien la lee con un imperativo con un pronombre pegado.",
  "votos": [
    {"respuesta": "si", "frase": "Úsala cuando se pregunte hasta cuándo se puede recurrir", "motivo": "…"},
    {"respuesta": "si", "frase": "Úsala cuando se pregunte", "motivo": "…"},
    {"respuesta": "si", "frase": "Úsala cuando se pregunte hasta cuándo", "motivo": "…"}
  ]
}
```

`frase` y `motivo` son los del primer voto. Sin juzgar: `"marcada": false`, `"sin_juzgar": "voto 1: la sesión del
juez terminó con error: …"`. Reutilizada: `"juzgada": false`, `procedencia` y el juicio de la medida de `main`.

```json
{"nombre": "habla_en_primera_o_segunda_persona",
 "descripcion": "Descripciones marcadas por el juez de las descripciones, sobre las juzgadas en esta ejecución",
 "medida": 1, "total": 1, "comparacion": "<=", "umbral": 0, "cumple": false, "decide": true}
```

`medida` cuenta solo lo juzgado en la ejecución (0 o 1) y `total` es 1 si se juzgó y 0 si se reutilizó o quedó sin
juzgar; lo reutilizado no entra (FR-031; ADR 0045, punto 1). Con `cumple: false` y `decide: true`, el motivo es el de
hoy, «umbral habla_en_primera_o_segunda_persona: 1 de 1 (100,0 %), y tiene que ser ≤ 0,0 %», y el veredicto, `fallo`. La descripción sin
juzgar añade el motivo propio, que no cita ningún umbral:

```text
el juez de las descripciones dejó sin juzgar la descripción de legal-core: voto 1: la sesión del juez terminó con error: …
```

`informe.md` gana un apartado «La descripción»: la skill; juzgada o reutilizada, con la procedencia; si está marcada,
la frase y el motivo; si quedó sin juzgar, el motivo del voto. Sin el bloque en la decisión, el informe no lleva ni el
bloque ni el umbral ni el apartado (FR-030).

## 7. La medida del juez (FR-034)

`make evals-medir-juez SKILL=descripciones` (`scripts/evals-medir-juez.sh descripciones`), que corre en la entrada
`descripciones` del trabajo `medida` (etiqueta `evals-medir-juez` o entrada `medir_al_juez`), ejecuta
`TestMedidaDelJuezDeLasDescripciones`: vota cada uno de los 99 casos de `casos.yaml` con §4 y §5 y escribe en
`-salida`, que el guion imprime entre `--- inicio de medida.json ---` y `--- fin de medida.json ---`:

```json
{
  "clase": "habla_en_primera_o_segunda_persona",
  "fecha": "2026-10-26",
  "modelo_del_juez": "claude-opus-5-5",
  "version_de_claude_code": "2.1.289",
  "rubrica": {"fichero": "rubrica.md", "sha256": "db35645c843209838355887b95ba197ce7cdb6f06e1d397a2bd115e55a7efbc4"},
  "casos": {"fichero": "casos.yaml", "sha256": "b9b2384a6adf9ac4219e9c4f65f1db430ca48a54a53b15704c7b46e677ecf31a"},
  "defectos": {"casos": 63, "sin_marcar": 0},
  "correctos": {"casos": 36, "marcados": 0},
  "origen": "ejecución de la medida del juez de las descripciones del job de evals sobre <commit>"
}
```

y, en el registro, una línea por caso mal etiquetado (`<id>: etiquetado <etiqueta> y <marcado|sin marcar>: «frase»`) y
por caso sin juzgar (`<id>: sin juzgar: <motivo>`). No escribe nada en el repositorio: la versiona una persona.

## 8. Tamaño y uso, de fuera adentro

- **Votos**: como mucho seis por descripción y trabajo; con las tres cambiadas, como mucho 18 en la ejecución. Con una
  sola cambiada y las otras dos con juicio en `main`, como mucho seis.
- **Informe**: una entrada por trabajo, de unos 300 bytes sin los votos y unos 1,5 KB con seis votos; no depende de
  las preguntas hechas ni de las normas consultadas, sino de una descripción, que cambia solo con una propuesta.
- **Cuándo deja de darse cada señal**: la marca, cuando una propuesta cambia la descripción y el juicio nuevo no la
  marca; la reutilización, cuando cambia la descripción o el juez (otra huella) o cuando su medida de `main` sale de la
  ventana de H30 (10 medidas, 90 días); el fallo de la medida, cuando la medida vuelve a corresponder y a cumplirse, que
  solo cambia una persona fuera de un run; el bloque entero, en cada ejecución cuya propuesta no cambia ninguna
  descripción.
