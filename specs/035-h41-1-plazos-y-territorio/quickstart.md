# Quickstart: validar H41.1

Desde la raíz del repositorio, con Go 1.27 (`go.mod`). Los escenarios 1 a 7 no usan modelo, credencial ni red, no
escriben en el índice ni en el historial de git ni en el árbol de trabajo: solo en la caché de compilación de Go y, el 5,
en `${TMPDIR:-/tmp}/h41-1-mutantes`, que borra al terminar. El 7, `make ci`, además pide red para `vuln` (`govulncheck`
consulta `https://vuln.go.dev`) y deja `coverage.out` y `coverage-integration.out` en la raíz, que `.gitignore` ignora
(research V12). Salen como se dice cuando el hito está hecho; antes, el 1 sale en verde (research V1) y los demás
nombran ficheros y tests que entrega el hito. Lo que comprueba cada uno, en contracts/.

## 1. Las skills caben en su tope y mandan leer como dice el plan

```sh
make skills-check
```

Verde: ningún cuerpo de `skills/` pasa de 11 000 caracteres, ninguna deriva tras `make skills-sync`, y
`TestRemisionesDelRepositorio` mira ya `plazos-y-territorio`: su regla da solo el fallo admitido de
`jerarquia_normativa.md`, y sus tres remisiones son las de la lista del plan (contracts/controles.md §2).

## 2. Lo que instala el binario

```sh
go test -count=1 -run '^TestEntregaDelHito$/^h41-(1-)?skills-instaladas$' ./internal/app/
```

Verde los dos guiones: el de H41.1 (contracts/aceptacion.md §1) y el de H41 con el bloque de FR-013 (§2). Construye los
binarios del arnés e2e en el temporal del test.

## 3. Los controles de `make ci` que toca el hito

```sh
go test -count=1 -run '^(TestSkillsDelProductoDesdeControl|TestVariantesDeControl|TestVariantesDePlazosYTerritorio|TestArnesDeLasFamilias|TestLaboratorioFueraDelProducto|TestRecetasDelSondeo)$' ./internal/evals/
go test -count=1 -run '^(TestRemisionesDelRepositorio|TestVerticalDeEnsayo|TestFrases)$' ./internal/skills/
```

Verde: `skills/plazos-y-territorio` es `-sin-condicion` con M1 y nada más, con el frontmatter de control; control y las
tres formas, sin cambiar un byte; las sesiones de `plazos-y-territorio-producto`, byte a byte `skills/plazos-y-territorio`;
las recetas del sondeo, como estaban; el cuerpo, con la línea de la vertical de ensayo, bajo 11 000 (contracts/controles.md
§1 a §4).

## 4. La skill frente a su variante, y su cuerpo

```sh
diff laboratorio/plazos-y-territorio/plazos-y-territorio-sin-condicion/skills/plazos-y-territorio/SKILL.md skills/plazos-y-territorio/SKILL.md
diff laboratorio/plazos-y-territorio/plazos-y-territorio-sin-condicion/skills/plazos-y-territorio/references/calculo_de_plazos.md skills/plazos-y-territorio/references/calculo_de_plazos.md
cmp laboratorio/plazos-y-territorio/plazos-y-territorio-sin-condicion/skills/plazos-y-territorio/references/administrativo.md skills/plazos-y-territorio/references/administrativo.md
cmp laboratorio/plazos-y-territorio/plazos-y-territorio-sin-condicion/skills/plazos-y-territorio/references/jerarquia_normativa.md skills/plazos-y-territorio/references/jerarquia_normativa.md
cmp laboratorio/plazos-y-territorio/plazos-y-territorio-sin-condicion/skills/plazos-y-territorio/references/municipios.md skills/plazos-y-territorio/references/municipios.md
awk 'f; NR>1 && /^---$/ && !f {f=1}' skills/plazos-y-territorio/SKILL.md | LC_ALL=en_US.UTF-8 wc -m
```

El primer `diff` sale con 1 y enseña solo el tramo `68,69c68` de contracts/skill.md §2; el segundo, con 1 y solo
`36a37,38`: una línea en blanco y la frase del 4 o 5 (§3); los tres `cmp` salen con 0 sin decir nada; y el recuento da
`10882` caracteres, la medida de `TestSkillsDelRepositorio` (medido en macOS; con la locale `C`, `wc -m` cuenta
bytes y da 11 052, así que la orden pide `en_US.UTF-8` instalada; research V2). Solo lee.

## 5. Cada control sale en rojo con una mutación de un solo cambio (SC-001)

Sobre una copia en el temporal, nunca sobre el árbol. Cada orden de `go test` marcada «rojo» sale con 1 y el mensaje de
contracts/controles.md §5; la última, en verde.

```sh
m="${TMPDIR:-/tmp}/h41-1-mutantes"
p=skills/plazos-y-territorio
v=laboratorio/plazos-y-territorio
rm -rf "$m" && mkdir -p "$m"
cp -R go.mod go.sum skills.go internal cmd skills laboratorio data evals schemas "$m"/

# a. un byte de calculo_de_plazos.md → rojo TestSkillsDelProductoDesdeControl
printf x >> "$m/$p/references/calculo_de_plazos.md"
(cd "$m" && go test -count=1 -run '^TestSkillsDelProductoDesdeControl$' ./internal/evals/)
cp "$p/references/calculo_de_plazos.md" "$m/$p/references/calculo_de_plazos.md"

# b. un byte de plazos-y-territorio-control → rojo TestVariantesDeControl
printf x >> "$m/$v/plazos-y-territorio-control/$p/references/municipios.md"
(cd "$m" && go test -count=1 -run '^TestVariantesDeControl$' ./internal/evals/)
cp "$v/plazos-y-territorio-control/$p/references/municipios.md" "$m/$v/plazos-y-territorio-control/$p/references/municipios.md"

# c. un byte de una forma de H41 → rojo TestVariantesDePlazosYTerritorio
printf x >> "$m/$v/plazos-y-territorio-lista-con-verbo/$p/SKILL.md"
(cd "$m" && go test -count=1 -run '^TestVariantesDePlazosYTerritorio$' ./internal/evals/)
cp "$v/plazos-y-territorio-lista-con-verbo/$p/SKILL.md" "$m/$v/plazos-y-territorio-lista-con-verbo/$p/SKILL.md"

# d. una palabra de una remisión → rojo TestRemisionesDelRepositorio («la remisión 3 no es la del plan»)
sed 's/el plazo sin calcular y la$/el plazo sin contar y la/' "$p/SKILL.md" > "$m/$p/SKILL.md"
(cd "$m" && go test -count=1 -run '^TestRemisionesDelRepositorio$' ./internal/skills/)
cp "$p/SKILL.md" "$m/$p/SKILL.md"

# e. 49 caracteres más en el cuerpo: 10 931, y 11 001 con la vertical → rojo TestVerticalDeEnsayo
printf 'xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\n' >> "$m/$p/SKILL.md"
(cd "$m" && go test -count=1 -run '^TestVerticalDeEnsayo$' ./internal/skills/)
# … y en verde TestSkillsDelRepositorio, porque sin la vertical cabe
(cd "$m" && go test -count=1 -run '^TestSkillsDelRepositorio$/^skills$' ./internal/app/)
cp "$p/SKILL.md" "$m/$p/SKILL.md"

# f. un byte al principio de un fichero de normativa y de jurisprudencia (la rama con desde vacío)
#    → rojo TestSkillsDelProductoDesdeControl/<skill>/desde_control
for s in normativa/references/con_herramientas.md jurisprudencia/references/documento_aportado.md; do
  { printf x; cat "skills/$s"; } > "$m/skills/$s"
  (cd "$m" && go test -count=1 -run '^TestSkillsDelProductoDesdeControl$' ./internal/evals/)
  cp "skills/$s" "$m/skills/$s"
done

# g. «lee » quitado de la orden de municipios.md, con el guion de H41 nuevo intacto
#    → rojo TestEntregaDelHito/h41-skills-instaladas en su línea 54 (el bloque de FR-013)
sed 's/llamada, lee \[municipios\.md\]/llamada, [municipios.md]/' "$p/SKILL.md" > "$m/$p/SKILL.md"
(cd "$m" && go test -count=1 -run '^TestEntregaDelHito$/^h41-skills-instaladas$' ./internal/app/)
cp "$p/SKILL.md" "$m/$p/SKILL.md"

rm -rf "$m"
```

Escribe solo en `$m`, que borra al final (research V7).

## 6. La guía y el registro de cambios (FR-009, FR-010)

```sh
grep -n 'hasta H41.1' docs/SKILLS.md
grep -n 'H41.1' CHANGELOG.md
awk '/^\*\*Cómo se remite\.\*\*/,/^El reparto \(ADR 0048\)/' docs/SKILLS.md
```

El primer `grep` sale con 1 sin nada; el segundo da la línea de la entrada de *Unreleased*; el `awk` imprime el apartado
«Cómo se remite» para leerlo contra FR-009: la línea de lo medido en `normativa` y `plazos-y-territorio`, sin el punto 2
del ADR 0052, y la viñeta de `jerarquia_normativa.md` como el fallo que el test admite (research D8). Solo lee.

## 7. Todo

```sh
make ci
```

## 8. La aceptación de la persona (fuera del run; SC-003)

Lo hace una persona antes de fusionar, con la credencial del sondeo y Claude Code en el `PATH`; abre 160 sesiones con
`claude-sonnet-5-5`, que consumen su suscripción, y escribe fuera del repositorio. El `DIRECTORIO` no puede existir antes.

```sh
V=$(claude --version | awk '{print $1}')
make evals-laboratorio VARIANTES=plazos-y-territorio-control,plazos-y-territorio-producto MODOS=orden,herramienta \
  SESIONES=40 MODELO=claude-sonnet-5-5 CLAUDE_CODE=$V CONCURRENCIA=5 DIRECTORIO=../kitlegal-sondeos/h41-1/plazos-y-territorio
make laboratorio-tabla DIRECTORIO=../kitlegal-sondeos/h41-1/plazos-y-territorio
```

La tabla tiene cuatro filas —dos variantes por dos modos—, 40 sesiones en cada una y diez por eval (02, 05, 10 y 11).
La persona la deja en `evidencias/sondeos-<fecha>/` con el modelo y la versión de Claude Code, y su lectura en la
propuesta de cambio. Vale si, en cada modo, `plazos-y-territorio-producto` no queda por debajo de
`plazos-y-territorio-control` —con intervalos de Wilson que no se tocan— en «Activan antes» ni en `administrativo.md`,
`calculo_de_plazos.md` y `municipios.md`, y si en `municipios.md` queda por encima en los dos modos, con intervalos que
no se tocan. Si no, no se fusiona y lo medido va a un ADR que dice qué medir después.
