# Quickstart: validar H25

Guía para comprobar la entrega sobre la cabeza del hito. Todas las órdenes se dan desde la raíz del repositorio.
Contratos: [juez-de-jurisprudencia](./contracts/juez-de-jurisprudencia.md),
[medida-y-casos](./contracts/medida-y-casos.md), [evals-jurisprudencia](./contracts/evals-jurisprudencia.md),
[job-de-evals](./contracts/job-de-evals.md) y [skill-jurisprudencia](./contracts/skill-jurisprudencia.md).

Los escenarios 1 a 9 no abren ninguna sesión con modelo ni piden nada a ninguna fuente; la única red es la de las
herramientas de Go en `make ci` (`vuln`), en el 9. El 10 lo ejecuta el workflow y el 11, una persona.

**Lo que dejan en el árbol.** `make build` deja `bin/kitlegal`, y `make ci`, `coverage.out` y
`coverage-integration.out`: los tres los ignora git (`.gitignore`: `/bin/`, `/coverage.*`). Ningún escenario toca el
índice, el historial, `evidencias/` ni la caché o el grafo de la cuenta: las órdenes del binario llevan
`KITLEGAL_CACHE_DIR` en un directorio temporal, y los tests usan los suyos.

**Lo que vale ya hoy**, antes de implementar nada: el escenario 0 y las tres órdenes del binario del 3, porque el
applet `cita` no cambia. Los nombres de test nuevos los fija el plan.

## 0. Antes de empezar

```bash
go version
make build
export KITLEGAL_CACHE_DIR="$(mktemp -d)"
```

Esperado: Go 1.27 y `bin/kitlegal`.

## 1. La carpeta del juez y sus copias (US3)

```bash
ls evals/jurisprudencia/juez
cmp evidencias/adr-0037-jurisprudencia/rubrica.md evals/jurisprudencia/juez/rubrica.md
cmp evidencias/adr-0037-jurisprudencia/esquema.json evals/jurisprudencia/juez/esquema.json
cmp evidencias/adr-0037-jurisprudencia/casos.yaml evals/jurisprudencia/juez/casos.yaml
cmp evidencias/adr-0037-jurisprudencia/medida.json evals/jurisprudencia/juez/medida.json
go test -count=1 -run '^TestCopiasDelJuez$' ./internal/evals/
```

Esperado: cinco ficheros (`casos.yaml`, `clases.yaml`, `esquema.json`, `medida.json`, `rubrica.md`); los cuatro `cmp`
sin salida; y `ok`.

## 2. La medida versionada corresponde y se cumple (US3)

```bash
go test -count=1 -run '^(TestMedidaVersionada|TestEjecucionSinMedir)$' ./internal/evals/
```

Esperado: `ok`. Los dos tests recorren las dos skills con juez: la medida del repositorio no da ninguna línea, y con
cada una de las cuatro claves cambiada o cada recuento distinto de 0 en una copia, la comprobación da su línea y el job
termina en `fallo` sin abrir ninguna sesión.

## 3. Las diez evals y sus preguntas (US5)

```bash
ls evals/jurisprudencia
go test -count=1 -run '^(TestPreguntasDelSondeo|TestPreguntasConElFragmento|TestConjuntoDeEvals|TestEvalsDelRepositorio)$' ./internal/evals/
bin/kitlegal cita preparar --resolucion 241/2013 --fecha 2013-05-09 --json
bin/kitlegal cita preparar --texto "cláusula suelo transparencia" --json
bin/kitlegal cita cotejar --json < evidencias/adr-0036/ecli-es-ts-2023-3144-fragmento.txt
```

Esperado:

1. Diez ficheros de eval, de `01-existe-con-numero-y-fecha.yaml` a `10-doctrina-dada-por-hecha.yaml`, y la carpeta
   `juez`.
2. `ok`.
3. Código 0, `"direccion":"https://www.poderjudicial.es/search/indexAN.jsp"` y tres casillas: «Nº Resolución» con
   `241/2013` y «Fecha resolución», «Desde» y «Hasta», con `09/05/2013`. Es lo que espera la eval (g).
4. Código 0 y `"direccion":"https://www.poderjudicial.es/search/sentencias/cl%C3%A1usula%20suelo%20transparencia/1/AN"`.
   La eval (j) espera la dirección que devuelva la sesión.
5. Código 0, `"roj":"STS 3144/2023"`, `"ecli":"ECLI:ES:TS:2023:3144"` y
   `"url":"kitlegal:documento/sha256:4886e0c8ca9836527ec08d8732b50315640a76803033374af5287b2a8321ee27"`, la huella
   del fragmento. Son el ECLI y el ROJ de la cita que esperan las evals (h) e (i).

## 4. La reconstrucción de los 249 casos (US4)

```bash
go test -count=1 -run '^(TestArgumentosDeLaInvocacion|TestTextoQuitado|TestResolverCasos|TestReconstruccionDeJurisprudencia|TestGrabacionesDerivadas)$' ./internal/evals/
```

Esperado: `ok`. Se resuelven los 249 —138 del informe del cierre de H23, 39 y 72 de los dos sondeos—, sin red, sin
modelo y sin Python; cada respuesta es la de su informe byte a byte; y cada uno de los 127 derivados solo se
diferencia de su sesión en lo quitado.

## 5. La ejecución de la medida, con un votante de pega (US4)

```bash
go test -count=1 -run '^TestEjecucionDeLaMedida$' ./internal/evals/
```

Esperado: `ok`. Con los 249 bien, 499 votos y la medida con sus cuatro claves y 0 de 125 y 0 de 124; con un defecto
sin marcar o un correcto marcado, el error nombra el caso y sus frases. No abre ningún voto de verdad.

## 6. El voto, el mensaje y el informe (US1, US2)

```bash
go test -count=1 -run '^(TestVotoDelJuez|TestMensajeDelVoto|TestTextosDeLaSesion|TestUmbralesDeJurisprudencia|TestInformeConElJuez)$' ./internal/evals/
```

Esperado: `ok`. `TestUmbralesDeJurisprudencia` fija los doce umbrales, diez que deciden; que una respuesta marcada en
un modo da `fallo` con su sesión y sus tres frases; que 901 s de votos dan `fallo`; que `afirma_que_existe` no cambia
el veredicto; y que el voto se publica con `sentencia`.

## 7. El trabajo del job (US6)

```bash
grep -n -A 2 'skill: jurisprudencia' .github/workflows/evals.yml
go test -count=1 -run '^TestDefinicionDelJob$' ./internal/evals/
```

Esperado: dos entradas de `include` con `concurrencia: 4`, la del trabajo `evals` —con `objetivo_de_duracion: 0`— y la
del trabajo `medida`; y `ok`. El test recalcula 12 577 s y 15 505 s, bajo los topes de 352 y 269 minutos.

## 8. `jurisprudencia` v0.1 (US7)

```bash
wc -l skills/jurisprudencia/SKILL.md
grep -c CAPTCHA skills/jurisprudencia/SKILL.md
git diff --stat main -- skills/jurisprudencia/SKILL.md
make skills-check
```

Esperado: 197 líneas; `2`, las dos menciones del párrafo inicial; un fichero con 9 líneas añadidas y 7 quitadas, las
de los dos pasajes del diff del contrato; y `skills-check` en verde. Si una reparación del cierre tocó la skill, las
cifras son las de contracts/skill-jurisprudencia.md §6.

## 9. Todo junto

```bash
make ci
```

Esperado: en verde, con `schema-check`, `skills-check` y las reglas del conjunto.

## 10. El job de cierre (lo lanza el workflow)

El workflow empuja la rama, abre la propuesta de cambio, pone la etiqueta `evals` y guarda el informe de cada skill en
`gates/evals/`. Sobre el de esta skill:

```bash
jq '.veredicto, (.umbrales | length), ([.umbrales[] | select(.decide)] | length), .red' specs/020-h25-el-juez-de/gates/evals/jurisprudencia.json
jq -r '.umbrales[] | [.nombre, .medida, (.total // "-"), .cumple, .decide] | @tsv' specs/020-h25-el-juez-de/gates/evals/jurisprudencia.json
jq '.juez.sin_juzgar | length' specs/020-h25-el-juez-de/gates/evals/jurisprudencia.json
```

Esperado: `"aprobado"`, `12`, `10` y `[]`; los doce umbrales de contracts/juez-de-jurisprudencia.md §5, todos con
`cumple` verdadero —`afirma_lo_no_leido` con 0 de 30 en los dos modos y los dos de la medida con 0 de 125 y 0 de 124—;
y `0` respuestas sin juzgar. Los informes de `boe-legislacion` y de `legal-core`, también `"aprobado"`.

## 11. Después del run (una persona)

1. Poner la etiqueta `evals-medir-juez` en la propuesta de cambio, antes de fusionar. Corren
   `medida del juez (boe-legislacion)` y `medida del juez (jurisprudencia)`. El registro de cada uno imprime su medida
   entre `--- inicio de medida.json ---` y `--- fin de medida.json ---`: la de `jurisprudencia`, con 0 de 125 y 0 de
   124. Si no se cumple, no se fusiona.
2. Leer las respuestas con algún voto afirmativo y las de las cuatro evals nuevas, y anotar en `docs/USO.md` si el
   juez acertó y si alguna respuesta anuncia todavía un CAPTCHA.
