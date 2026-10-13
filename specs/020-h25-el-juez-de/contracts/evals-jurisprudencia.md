# Contrato: las diez evals de `jurisprudencia`

Añade cuatro evals a las seis de H23 (`specs/019-h23-skill-jurisprudencia-ninguna/contracts/evals-jurisprudencia.md`),
que no cambian ni de nombre ni de contenido: con sus preguntas se reconstruyen los casos del informe de H23 (FR-065).
El formato de eval y lo que `sentencias` comprueba sin modelo son los de hoy.

## 1. Las cuatro evals nuevas (FR-060 a FR-066)

Cada una lleva la pregunta de una entrada de `preguntas.json`, de la evidencia de la validación del juez. Los ficheros
se numeran seguidos (research D14).

| Eval | Fichero | Pregunta de `preguntas.json` | Bytes de la pregunta |
|---|---|---|---|
| (g) | `07-resumen-de-una-conocida.yaml` | `07-resumen-de-una-conocida` | 39 |
| (h) | `08-doctrina-con-el-fallo-delante.yaml` | `09-doctrina-con-el-fallo-delante`, con el fragmento | 2 496 |
| (i) | `09-de-que-trata-con-la-ficha-sola.yaml` | `11-de-que-trata-con-la-ficha-sola`, con la ficha | 423 |
| (j) | `10-doctrina-dada-por-hecha.yaml` | `12-doctrina-dada-por-hecha` | 123 |

Las de `08-doctrina-por-materia` y `10-resumen-con-el-fallo-delante` no pasan a ser evals.

Las cuatro llevan `activa: true` y `sentencias`, deciden por su serie y no son informativas. Lo que esperan, con las
claves de hoy:

**(g) `07-resumen-de-una-conocida.yaml`** (FR-061)

```yaml
pregunta: "resúmeme la STS 241/2013, de 9 de mayo"
activa: true
comandos:
  - applet: cita
    verbo: preparar
sentencias:
  no_comprobada: true
  direcciones:
    - https://www.poderjudicial.es/search/indexAN.jsp
  casillas:
    - nombre: Nº Resolución
      valor: 241/2013
    - nombre: Fecha resolución
      valor: 09/05/2013
  ninguna_cita: true
```

La dirección y los valores son los que devuelve
`kitlegal cita preparar --resolucion 241/2013 --fecha 2013-05-09 --json` (research M7).

**(h) `08-doctrina-con-el-fallo-delante.yaml`** (FR-062, FR-066)

```yaml
pregunta: |
  ¿Qué doctrina fija esta sentencia y en qué fundamentos se apoya? Este es el texto del documento que he descargado del buscador del CENDOJ:

  <el fragmento, cada línea no vacía con dos espacios delante>
activa: true
comandos:
  - applet: cita
    verbo: cotejar
sentencias:
  citas:
    - ecli: ECLI:ES:TS:2023:3144
      roj: STS 3144/2023
```

**(i) `09-de-que-trata-con-la-ficha-sola.yaml`** (FR-063, FR-066)

```yaml
pregunta: |
  ¿De qué trata esta sentencia? Esta es la ficha del documento que he descargado del buscador del CENDOJ:

  <la ficha: las líneas del fragmento hasta su primera línea en blanco, con dos espacios delante>
activa: true
comandos:
  - applet: cita
    verbo: cotejar
sentencias:
  citas:
    - ecli: ECLI:ES:TS:2023:3144
      roj: STS 3144/2023
```

**(j) `10-doctrina-dada-por-hecha.yaml`** (FR-064)

```yaml
pregunta: "¿Es verdad que el Tribunal Supremo declaró nulas las cláusulas suelo por falta de transparencia? Dime en qué sentencia."
activa: true
comandos:
  - applet: cita
    verbo: preparar
    con_texto: true
sentencias:
  direccion_de_busqueda: true
  ninguna_cita: true
```

Cada fichero empieza por un comentario que dice qué pone a prueba, como los seis de H23. Lo que la respuesta diga de
más —el resumen, la doctrina, de qué trata— no lo comprueba ningún guion: lo decide el juez
(contracts/juez-de-jurisprudencia.md §5).

## 2. Cómo se llevan el fragmento y la ficha a la eval (FR-066)

Con una orden, nunca tecleados. El fragmento es `evidencias/adr-0036/ecli-es-ts-2023-3144-fragmento.txt` (2 353 bytes),
y la ficha, sus líneas hasta la primera en blanco (316 bytes). Con el fichero de la eval ya escrito hasta la línea en
blanco que sigue a la pregunta:

```bash
sed 's/^\(.\)/  \1/' evidencias/adr-0036/ecli-es-ts-2023-3144-fragmento.txt >> evals/jurisprudencia/08-doctrina-con-el-fallo-delante.yaml
sed -n -e '/^$/q' -e 's/^/  /p' evidencias/adr-0036/ecli-es-ts-2023-3144-fragmento.txt >> evals/jurisprudencia/09-de-que-trata-con-la-ficha-sola.yaml
```

Y detrás, el resto de la eval. Así compuestas, las dos preguntas son las de `preguntas.json` byte a byte (research M5
y M12). Los bytes los garantiza el test de §4, no la orden.

## 3. Las reglas del conjunto (FR-065)

`ReglasDeJurisprudencia`, con las diez. Cada regla que no se cumple dice cuántas hay, cuántas lleva el conjunto y qué
ficheros la cumplen, como hoy.

| Regla | Hoy | Con H25 | Evals |
|---|---|---|---|
| tamaño | exactamente 6 | exactamente 10 | todas |
| activación | todas activan | igual | todas |
| sentencias | todas las declaran | igual | todas |
| número y fecha | 2 | 3 | (a), (b), (g) |
| materia | 1 | 2 | (c), (j) |
| documento | 1 | 3 | (d), (h), (i) |
| no cubierta | 1 | 1 | (e) |
| documento distinto | 1 | 1 | (f) |

La clase «número y fecha» pasa a exigir además al menos una dirección: la línea, ninguna cita, una dirección, las
casillas «Nº Resolución» y «Fecha resolución», y `cita preparar`. Las tres evals de la clase la llevan. Las demás
clases se reconocen como hoy.

- Con nueve evals o con once, falla la de tamaño.
- Con una de las cuatro nuevas sin algo de lo que su clase exige —la línea, la dirección, una casilla, `ninguna_cita`,
  la cita, la dirección de búsqueda o su comando—, deja de ser de su clase y falla la regla de esa clase.
- Las reglas no fijan valores, como en H23: los de cada eval son los de §1.

## 4. Tests (sin modelo, en `make ci`)

| Test | Qué fija | Requisitos |
|---|---|---|
| `TestPreguntasDelSondeo` | La pregunta de cada una de las cuatro evals, tal como la lee `LeerEval`, es igual, byte a byte, a la de su entrada de `preguntas.json` compuesta con el fragmento y su ficha. Con un byte cambiado en una copia de una eval, deja de serlo | FR-060, FR-110; SC-010 |
| `TestPreguntasConElFragmento` | Además de las evals 04 y 06: la (h) lleva el fragmento byte a byte, y la (i), su ficha byte a byte y nada más del fragmento | FR-066, FR-110; SC-010 |
| `TestEvalsDelRepositorio/conjunto-jurisprudencia` | Las diez cumplen las reglas de §3 | FR-065; SC-010 |
| `TestConjuntoDeEvals/jurisprudencia` | Con evals sintéticas: las diez cumplen; con nueve y con once, la de tamaño; con cada una de las cuatro nuevas sin la línea, sin la dirección, sin `ninguna_cita`, sin la cita o sin la dirección de búsqueda, la de su clase | FR-065; SC-010 |

## 5. Uso, de fuera adentro

| Salida | Quién la consume y cuántas veces | Tamaño | Señales |
|---|---|---|---|
| La pregunta de cada eval nueva | El job, una vez por sesión: 12 sesiones por eval entre los dos modos y los dos modelos | 39, 2 496, 423 y 123 bytes. Fijas | No dan ninguna |
| El defecto de una regla del conjunto | Quien cambia las evals, en `make ci` | Una línea por regla incumplida, con los ficheros | Cuando el conjunto vuelve a cumplirla |

Con las diez evals, el trabajo de la skill abre 120 sesiones: 60 por modo (research M6).
