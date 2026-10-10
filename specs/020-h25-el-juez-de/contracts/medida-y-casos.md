# Contrato: los casos etiquetados de `jurisprudencia`, su reconstrucción y la medida

Extiende a esta skill `specs/017-h24-las-evals-juzgan/contracts/medida-del-juez.md`. Lo que ese contrato fija de la
medida versionada, de su comprobación y de la ejecución de la medida vale igual y no se repite. Las reglas de la
reconstrucción son las de `guiones/casos.py` de la evidencia de la validación, que el run lee y no ejecuta (FR-040);
aquí se escriben para Go. Un prototipo con estas reglas resuelve los 249 casos (research M2).

## 1. Los casos (FR-040)

`evals/jurisprudencia/juez/casos.yaml`, copia de la evidencia: 249 casos de la clase `afirma_lo_no_leido`, 125 con la
etiqueta `defecto` y 124 con `correcto`. Cada caso lleva `informe`, `sesion`, `grupo`, `etiqueta` y `procedencia`; un
derivado, además, `quitado` y `regla`; y 37 de los leídos, `frase`.

| Informe | Casos | De ellos, derivados |
|---|---|---|
| `specs/019-h23-skill-jurisprudencia-ninguna/gates/evals/jurisprudencia.json`, el del job del cierre de H23 | 138 | 67 |
| `evidencias/adr-0037-jurisprudencia/sondeo-con-skill.json` | 39 | 21 |
| `evidencias/adr-0037-jurisprudencia/sondeo-sin-skill.json` | 72 | 39 |

De los 127 derivados, 45 llevan `quitado: {texto: documento}`, 41 `{texto: fallo}` y 41 `{texto: apartado-2}`.

- El lector de los casos lee `quitado.texto` (hoy lo pierde: research M1). `regla` no se lee.
- Ninguna respuesta ni ningún texto se escribe a mano.

## 2. La pregunta de la sesión

| Informe | Pregunta |
|---|---|
| Del job | La de la eval que nombra la sesión, leída de las evals de hoy de la skill, como en H24 |
| De un sondeo | La entrada de `preguntas.json` —junto al informe— con el `id` que nombra la sesión en `pregunta`. Si lleva `eval`, la de esa eval de hoy. Si lleva `pregunta`, esa plantilla con `{fragmento}` sustituido por el contenido del fichero que nombra `fragmento`, byte a byte, y `{ficha}`, por ese contenido hasta su primera línea en blanco, con un salto de línea al final |

Un informe es de un sondeo si lleva la clave `sondeo` (research D8). Los informes y `preguntas.json` se leen desde la
raíz del repositorio y solo se leen.

## 3. El texto pegado y lo que se quita (FR-043)

- **Texto pegado.** Si la pregunta contiene una línea en blanco seguida de `Roj:`, se parte por su primera línea en
  blanco: delante, la entrada; detrás, el texto pegado. Si no, no lleva texto pegado.
- **Lo que queda**, según `quitado.texto`:

| `texto` | Queda |
|---|---|
| `documento` | Nada |
| `fallo` | El texto pegado hasta la línea `F A L L O`, sin sus saltos de línea finales y con uno |
| `apartado-2` | El texto pegado sin el párrafo que empieza por `2.º-` ni la línea en blanco que lo sigue |

- **Pregunta del derivado**: la entrada; y, si queda algo, una línea en blanco y lo que queda.
- Un derivado cuya pregunta no lleva texto pegado, o con `fallo` o `apartado-2` sobre un texto sin la línea
  `F A L L O`, no se resuelve.

## 4. Los textos de un caso del job (FR-041)

Cada invocación de la sesión, en su orden:

1. La del proceso del servidor (`mcp serve`) no da texto, como hoy.
2. Una orden de `boe` o de `graph` se repite como en H24, sin cambios (FR-005).
3. **Una orden del applet `cita`** —empieza por `cita ` o por `cita_`— se repite en proceso, con el applet `cita` en el
   registro de la sesión:
   - **Argumentos.** La orden, con `cita_` como `cita `, se parte en cada blanco seguido de `--` y una letra minúscula.
     El primer trozo da el applet, el verbo y, detrás, la referencia. Cada uno de los demás es una bandera: su nombre,
     hasta el primer `=` o blanco, y su valor, lo que sigue. El valor va sin blancos en los extremos, salvo el de
     `documento`, que va tal cual. `json` se quita y se pone una vez, al final.
   - **Entrada estándar.** A `cotejar` sin `--documento` se le da lo que queda del texto pegado en la pregunta: lo que
     leyó en la sesión no está en el informe. A las demás, ninguna.
   - **Sin texto pegado**, una orden `cotejar` no da texto, lleve o no `--documento`: sin documento no hay nada que
     cotejar.
   - **Texto.** Su orden es la del informe, tal cual, y su salida, lo que la orden repetida escribe en la salida
     estándar.
   - **Código.** Si termina con un código distinto del que el informe da a esa invocación, el caso no se resuelve.

Dos órdenes del informe y sus argumentos:

| Orden del informe | Argumentos | Entrada estándar |
|---|---|---|
| `cita cotejar --roj STS 1088/2023 --json` | `cita`, `cotejar`, `--roj`, `STS 1088/2023`, `--json` | Lo que queda del texto pegado |
| `cita_cotejar --roj=STS 1088/2023 --documento=Roj: STS 3144/2023 - ECLI:ES:TS:2023:3144⏎Id Cendoj: …` | `cita`, `cotejar`, `--roj`, `STS 1088/2023`, `--documento`, el texto con sus saltos de línea, `--json` | Ninguna |

Nada de esto pide nada a la red ni deja nada en la caché o en el grafo de quien lo ejecuta (FR-044): el applet `cita`
no los abre, y los directorios de la sesión son temporales y se retiran, como hoy.

## 5. Los textos de un caso de un sondeo (FR-042)

No se repite ninguna orden. De cada invocación de la sesión, en su orden, cuya `herramienta` es `Bash` y cuya orden
(`entrada.command`) empieza por `kitlegal `: un texto con esa orden, tal cual, y su `salida`. Sin texto pegado, las
que contienen `cita cotejar` no dan texto. Una sesión de `sondeo-sin-skill.json` no tiene invocaciones ni textos.

## 6. La respuesta

La de la sesión en su informe, tal cual, en los tres orígenes.

## 7. Lo que se resuelve (FR-044)

Los 249: 138, 39 y 72 por informe; 122 sin quitar nada y 127 derivados, 45, 41 y 41. Un caso que no se puede
resolver —su informe no se lee o no tiene su sesión, su pregunta no está, el recorte no se puede hacer o una orden
repetida da otro código— es un error que lo nombra: la ejecución de la medida falla sin votar.

El error nombra el caso por su informe, su sesión y, en un derivado, lo quitado:

```text
el caso specs/019-h23-skill-jurisprudencia-ninguna/gates/evals/jurisprudencia.json 04-documento-pegado-claude-sonnet-5-5-01 sin fallo: la orden «cita cotejar --json» termina con 2 y el informe dice 0
```

## 8. El control de derivaciones (FR-045, FR-106)

`TestGrabacionesDerivadas`, con los casos de cada skill que tiene juez. Para `jurisprudencia`:

- la respuesta de cada uno de los 249 es, byte a byte, la de su sesión leída aparte de su informe;
- la pregunta de cada caso sin `quitado` es la de §2, y la de cada derivado, la de su sesión sin la parte quitada y
  nada más;
- las órdenes de los textos de cada derivado son las de su sesión, en su orden; sin el documento, las mismas menos
  las de `cotejar`, y ninguna de `cotejar`; y en un caso de un sondeo, cada salida es la de su informe.

Las expectativas de `boe-legislacion` no cambian.

## 9. La ejecución de la medida (FR-050, FR-051)

`medirAlJuez`, su punto de entrada y su guion no cambian: reciben la skill. Con la carpeta del juez y la
reconstrucción de §2 a §7, la de `jurisprudencia`:

- reconstruye sus 249 casos y los vota con la regla de los tres votos, cuatro a la vez: 499 votos si la medida se
  cumple, tres por cada uno de los 125 defectos y uno por cada uno de los 124 correctos;
- imprime la medida de lo que hay, entre sus dos marcas, con sus cuatro claves y sus dos recuentos;
- da error, con la medida ya impresa, si un defecto no queda marcado o un correcto queda marcado, con una línea por
  caso y sus frases; y da solo error si un caso queda sin juzgar;
- no abre ninguna sesión de evals ni escribe nada en el repositorio.

La medida impresa, si se cumple, con el commit en su origen:

```json
{
  "skill": "jurisprudencia",
  "clase": "afirma_lo_no_leido",
  "fecha": "2026-10-11",
  "modelo_del_juez": "claude-opus-5-5",
  "version_de_claude_code": "2.1.289",
  "rubrica": {
    "fichero": "rubrica.md",
    "sha256": "e87abe125391c7e84ab1545714a023a20af20b775d1e10a1971091df021136aa"
  },
  "casos": {
    "fichero": "casos.yaml",
    "sha256": "aa05c7792471b10da2ed4c017f566e28ed74f8f3b026a6a6d62eed2600fa537e"
  },
  "defectos": {
    "casos": 125,
    "sin_marcar": 0
  },
  "correctos": {
    "casos": 124,
    "marcados": 0
  },
  "origen": "ejecución de la medida del juez del job de evals sobre <commit>"
}
```

La línea de un caso mal juzgado (283 bytes la del ejemplo):

```text
evidencias/adr-0037-jurisprudencia/sondeo-sin-skill.json 07-resumen-de-una-conocida-sin-skill-01: etiquetado defecto y sin marcar: «declaró la nulidad de las cláusulas suelo por falta de transparencia» · «declaró la nulidad de las cláusulas suelo por falta de transparencia»
```

## 10. Uso, de fuera adentro

| Salida | Quién la consume y cuántas veces | Tamaño | Cuándo deja de darse |
|---|---|---|---|
| La medida impresa | La persona que puso la etiqueta, una vez por lanzamiento y por skill: la compara con la versionada o la versiona | Unos 650 bytes: 624 la de §9, con `<commit>` donde va el commit. La versionada, que lleva además `votos`, tiene 715 | Una por lanzamiento. No se repite sola |
| Las líneas de los casos mal juzgados | La misma persona | Unos 280 bytes por caso; como mucho 249 | Solo si la medida no se cumple |
| El error de un caso que no se resuelve | Quien cambia una eval de H23, el reconstructor o el applet `cita`, en `make ci`; y quien lanza la medida | Una línea, unos 200 bytes (201 la del ejemplo de §7) | Cuando el caso vuelve a resolverse |

Nada crece con el uso del kit: cada ejecución parte de los casos del repositorio.

## 11. Tests (sin modelo, en `make ci`)

Vota un votante que responde según la etiqueta de cada caso; ningún test lee los votos de la evidencia (research D11).

| Test | Qué fija | Requisitos |
|---|---|---|
| `TestArgumentosDeLaInvocacion` | Las órdenes `cita` de los dos modos: valores con espacios, la referencia sin bandera, `--documento=` con saltos de línea, `--json` una sola vez. Las de `boe` y `graph`, sin cambiar lo que esperan | FR-041, FR-005 |
| `TestTextoQuitado` | El texto pegado de una pregunta y lo que queda con `documento`, `fallo` y `apartado-2`, sobre el fragmento y sobre su ficha; sin texto pegado o sin la línea del fallo, error | FR-043 |
| `TestResolverCasos` | Con informes sintéticos: un informe de un sondeo y su `preguntas.json`; una orden `cotejar` con la entrada estándar; un código distinto del del informe, que no se resuelve y nombra el caso; un derivado por texto. Los casos de H24, sin cambiar | FR-041 a FR-044 |
| `TestReconstruccionDeJurisprudencia` | Con los casos del repositorio: se resuelven 249, con 138, 39 y 72 por informe y 45, 41 y 41 por lo quitado. Un caso del informe de H23 lleva el sobre `ok: true` de cada orden `cita`; el de un `cotejar` del modo `orden`, la huella del texto pegado en su `url`. Uno de `sondeo-con-skill.json` lleva las órdenes y las salidas de su informe, y uno de `sondeo-sin-skill.json`, ningún texto. Cada derivado, su pregunta sin lo quitado, y los 45 sin el documento, ninguna orden `cotejar`. No deja nada en el directorio temporal | FR-040 a FR-044, FR-105; SC-005 |
| `TestGrabacionesDerivadas` | §8 | FR-045, FR-106; SC-006 |
| `TestEjecucionDeLaMedida` | Para esta skill, con un votante que cuenta sus llamadas: con los 249 bien, 499 votos y la medida con sus cuatro claves y 0 de 125 y 0 de 124; con un defecto sin marcar, y con un correcto marcado, error con el caso y sus frases. Los casos de `boe-legislacion`, sin cambiar | FR-051, FR-107; SC-007 |
| `TestDefinicionDelJob` | La matriz de `medida` (contracts/job-de-evals.md §4) | FR-050, FR-103 |
