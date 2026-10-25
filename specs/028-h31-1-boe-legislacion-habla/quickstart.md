# Quickstart: validar H31.1

Escenarios para comprobar la entrega desde la raíz del repositorio, en la rama del hito, sin red ni modelo. Todos son
de lectura: no cambian el árbol de trabajo, el índice ni el historial, salvo el escenario 9, que deja los ficheros de
cobertura que `make ci` escribe siempre (`coverage.out` y `coverage-integration.out`, ignorados por `.gitignore`).
Hacen falta Go 1.27.3 (`toolchain` de `go.mod`), `git`, `jq` y la rama `main` local. Lo que esperan es el estado tras
el hito; entre paréntesis, lo que dan hoy. El sondeo de SC-012 no está aquí: lo lanza una persona fuera del run.

## 1. El cuerpo cabe (FR-005, SC-002)

```bash
awk 'NR>1 && /^---$/ {f=1; next} f' skills/boe-legislacion/SKILL.md | wc -m
```

Espera un número de 11 000 o menos: 10 993 con el texto de contracts/skill-boe-legislacion-v0-3-1.md (hoy, 10 992).
`wc -m` cuenta caracteres en una locale UTF-8, como el gate de `make skills-check` (escenario 5), que es el que decide y
sale en rojo con 11 001; con otra locale, `wc -m` cuenta bytes y da más.

## 2. Los seis ejemplos, en el cuerpo (FR-001, SC-003)

```bash
grep -oF -e '«el artículo no lleva avisos de vigencia»' -e '«la lectura no trajo avisos»' \
  -e '«el art. N, que es…»' -e '«el art. N sobre <materia>»' \
  -e '«los artículos que regulan <materia> (N y M)»' -e '«no puedo decir qué <regla> fija»' \
  skills/boe-legislacion/SKILL.md | wc -l
```

Espera `6`, cada ejemplo entero en una línea (hoy, `1`: «el art. N, que es…»; «no puedo decir qué <regla> fija» está
partido entre dos líneas y los otros cuatro no están). Las reglas de §2 del contrato, el paso 3 y el paso 5 los coteja
la revisión final.

## 3. Las líneas añadidas no nombran las normas de las evals por materia (FR-004)

```bash
git diff main -- skills/boe-legislacion/SKILL.md | grep '^+' \
  | grep -cE 'BOE-A-2004-4214|BOE-A-2006-20764|BOE-A-2013-12887|Ley 35/2006|Real Decreto Legislativo 2/2004|Ley 19/2013|IRPF|TRLRHL'
```

Espera `0` (y `grep` sale con 1).

## 4. El frontmatter, con su descripción, no cambia (FR-008)

```bash
git diff -U0 main -- skills/boe-legislacion/SKILL.md | grep -cE '^@@ -([1-9]|1[0-5])[ ,]'
```

Espera `0`: ningún trozo del diff empieza en las líneas 1 a 15 de `main`, las del frontmatter.

## 5. Los gates y los controles de contenido (FR-005 a FR-007, SC-001)

```bash
make skills-check
```

Espera `ok` en `internal/app`, `internal/skills` e `internal/evals`: el cuerpo, el frontmatter, cada referencia
enlazada, `prosa-de-la-skill`, `ordenes-para-powershell`, `linea-sin-consulta`, `formas-de-la-redaccion-fechada`,
`avisos-de-la-skill`, `hallazgos-de-la-skill`, `lineas-de-aviso-de-las-referencias`, `ordenes-de-ejemplo`,
`normas-nombradas`, `sin-instrucciones-de-evals` y las `lecturas`.

## 6. Las lecturas, grabadas y servidas sin red (FR-009 a FR-011, SC-004, SC-008)

```bash
jq -r '.normas[] | select(.titulo_empieza_por=="Ley 35/2006," or .titulo_empieza_por=="Real Decreto Legislativo 2/2004,")
  | "\(.titulo_empieza_por) \(.bloques | join(" "))"' internal/evals/testdata/grabaciones.json
```

Espera `Ley 35/2006, a18 a7 a19 a20 dadecimoctava` y `Real Decreto Legislativo 2/2004, a15 a60 a92 a100 a104` (hoy,
`a18` y `a15`).

```bash
git ls-files internal/evals/testdata/boe.legislacion-consolidada \
  | grep -cE '20764_texto_bloque_(a7|a19|a20|dadecimoctava)\.json$|4214_texto_bloque_(a60|a92|a100|a104)\.json$'
```

Espera `8`, las grabaciones de `grabar_datos` (hoy, `0`).

```bash
go test -count=1 -run '^TestEvalsDelRepositorio$/^(lecturas|grabado)$' ./internal/evals/
go test -count=1 -run '^TestCacheDeLasSesiones$' ./internal/evals/
go test -count=1 -run '^TestLecturaDeLaEvalSinGrabar$' ./internal/evals/
```

Espera `ok` en las tres: las lecturas de cada eval son las del contrato, la unión de grabaciones las sirve, cada bloque
nuevo sale con 0 con `--offline` desde la caché de una sesión de su eval, sin fecha y con `--fecha 9999-12-31`, y una
lectura sin grabar sigue dando su falta.

## 7. El tope del trabajo de evals (FR-013, SC-005)

```bash
go test -count=1 -run '^TestDefinicionDelJob$' ./internal/evals/
grep -n 'timeout-minutes: 352' .github/workflows/evals.yml
```

Espera `ok` y la línea 276: 352 minutos, 21 120 s, cubren el peor caso de `boe-legislacion`, 20 117 s, que las lecturas
no cambian (research D10).

## 8. La guía, el CHANGELOG y la frase de CONTRIBUTING (FR-012, FR-014)

```bash
grep -c 'brazo-D' docs/SKILLS.md
awk '/^## \[Unreleased\]/{f=1; next} /^## \[/{f=0} f' CHANGELOG.md | grep -c 'boe-legislacion` v0.3.1'
grep -c 'regla 7 sin su punto final' CONTRIBUTING.md
```

Espera `1`, `1` y `0` (hoy, `0`, `0` y `1`).

## 9. `make ci` (SC-001)

```bash
make ci
```

Espera `ci: todos los controles en verde`. Deja `coverage.out` y `coverage-integration.out`, ignorados.

## 10. Lo que mide el cierre (SC-007 a SC-011)

No se ejecuta en local: el workflow lanza el job de evals sobre la cabeza de la propuesta y el informe final lee, sin
modelo, los umbrales de plan.md («Controles de umbral»). Lo que mide una persona antes de fusionar (SC-012) es el sondeo
de la eval 15 con el arnés de `evidencias/sondeos-20261025/`.
