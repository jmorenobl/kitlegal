# Quickstart: validar H33 en la cabeza de la rama

Escenarios para comprobar la entrega cuando el run ha terminado. Todos se ejecutan desde la raíz del repositorio, sin
red (más allá de la de las herramientas de Go para el caché de módulos) y sin sesiones con modelo; ninguno cambia el
árbol de trabajo, el índice ni el historial: los tests escriben en `t.TempDir()` y el escenario 2 trabaja en un
directorio temporal que borra al final. Antes de la entrega, los escenarios 1 a 6 salen en rojo: es lo esperado.

**Requisitos**: el Go del `go.mod`, `git`, `grep`, `awk`, `wc` y un locale UTF-8 para el escenario 6.

## 1. La aceptación e2e (FR-025; historia 5)

```sh
go test ./internal/app/ -run 'TestEntregaDelHito/h33-cobertura-instalada$' -count=1 -v
```

Esperado: `--- PASS: TestEntregaDelHito/h33-cobertura-instalada`. El guion es el de contracts/aceptacion.md §2.

## 2. Lo que se instala, a mano (FR-004, FR-007, FR-025)

```sh
dir=$(mktemp -d)
go build -o "$dir/kitlegal" ./cmd/kitlegal
(cd "$dir" && ./kitlegal skills install --dir destino --json > /dev/null)
grep -nE '^⚠ NO (CUBIERTO|VERIFICADO): ' "$dir/destino/consulta-juridica/SKILL.md" "$dir/destino/plazos-y-territorio/SKILL.md"
grep -A5 '^<!-- lo que cubre la base' "$dir/destino/consulta-juridica/SKILL.md"
grep -cE '^\| .* \| ⚠ NO CUBIERTO: .*: (se rige por|lo regula aparte)' "$dir/destino/plazos-y-territorio/references/administrativo.md"
rm -rf "$dir"
```

Esperado: tres líneas —la plantilla en los dos `SKILL.md`, idéntica, y la marca solo en `consulta-juridica`—; la marca
de inicio, las cuatro líneas `- **Normativa**`, `- **Jurisprudencia**`, `- **Plazos**`, `- **Territorio**` y la marca de
fin; y `9`. Todo queda dentro de `$dir`, que se borra.

## 3. Los datos y las regiones (FR-001 a FR-003, FR-006, FR-007, FR-011, FR-023; SC-001, SC-002, SC-008)

```sh
go test ./internal/skills/ -run 'TestLeerCobertura|TestRegionDeLoCubierto|TestRegiones|TestRegionDeLaCobertura|TestVerticalDeEnsayo' -count=1
go test ./internal/app/ -run 'TestSkillsDelRepositorio' -count=1
```

Esperado: `ok` en los dos. `TestSkillsDelRepositorio` incluye `deriva` (0 ficheros que cambiaría `make skills-sync`) y
`formas-compartidas` (la plantilla, la misma línea en las dos skills).

## 4. Las evals, sin modelo (FR-012 a FR-019; SC-006 a SC-008)

```sh
go test ./internal/evals/ -run 'TestFormatoDeLaCobertura|TestFormatoDePlazos|TestJuzgarLaDeclaracionDeNoCubierto|TestJuzgarPlazos|TestActivadaPor|TestSkillsNombradasEnTambienLaResuelven|TestEvalsDelRepositorio|TestUmbralesDeLasSkills|TestCacheDeLasSesiones|TestDefinicionDelJob|TestMatrizDeLasPropuestas' -count=1
```

Esperado: `ok`. `TestCacheDeLasSesiones` sirve el art. 59 del ET en la fecha 2026-11-02 desde lo grabado, sin red.

## 5. Las suites congeladas, con sus cuatro cambios y ninguno más (FR-020, FR-021; SC-009)

```sh
go test ./internal/evals/ -run 'TestSkillsDelProductoDesdeControl|TestConsultaJuridicaDesdeLaBase' -count=1
go test ./internal/app/ -run 'TestEntregaDelHito/h32-(paquetes|capacidades)$' -count=1
git diff --stat main -- specs/030-h32-arquitectura-de-skills/aceptacion/
```

Esperado: `ok` en los dos `go test` y ninguna línea del `git diff`.

## 6. Los cuerpos (FR-008 a FR-010; SC-005)

```sh
awk 'BEGIN{n=0} /^---$/ && n<2 {n++; next} n>=2' skills/consulta-juridica/SKILL.md | wc -m
awk 'BEGIN{n=0} /^---$/ && n<2 {n++; next} n>=2' skills/plazos-y-territorio/SKILL.md | wc -m
```

Esperado: `4481` y `10686`.

## 7. Lo grabado (FR-018)

```sh
git log --oneline -1 -- internal/evals/testdata/boe.legislacion-consolidada/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-11430_texto_bloque_a59.json
grep -c '"titulo_empieza_por": "Real Decreto Legislativo 2/2015,", "bloques": \["a59"\]' internal/evals/testdata/grabaciones.json
```

Esperado: el commit del paso `grabar_datos` que lo grabó, y `1`.

## 8. Todo `make ci` (SC-009)

```sh
antes=$(mktemp)
git status --porcelain > "$antes"
make ci
git status --porcelain | diff "$antes" -
rm -f "$antes"
```

Esperado: `make ci` termina con 0 y el `diff` no da nada (lo que `make ci` deja en disco está en `.gitignore`).

## 9. La documentación (FR-022)

```sh
grep -cE 'fuera_de_lo_declarado|tambien_la_resuelven' CONTRIBUTING.md
grep -c 'las dos formas' docs/SKILLS.md
grep -c 'en prosa' docs/SKILLS.md
grep -c '15 013 s en consulta-juridica' .github/workflows/evals.yml
grep -c 'H33' docs/HISTORIAL.md
```

Esperado: al menos `2`, `0`, `0`, `1` y al menos `1` (hoy, `0`, `2`, `1`, `0` y `0`: `grep -c` sale con 1 cuando
cuenta 0). Lo demás de FR-022 (`docs/USO.md`, `CHANGELOG.md`, la página de la web) lo lee la revisión en el diff.

## 10. La medición del cierre (FR-024; SC-004)

No se ejecuta aquí: la hace el workflow tras la revisión final, con el job de evals sobre la cabeza, y su resultado
llega en el informe final. Se cumple si los umbrales de «Controles de umbral» con `evals:` salen con `cumple: true`, en
particular `cobertura_sin_declarar` de `consulta-juridica` y de `plazos-y-territorio` y `sin_activar` de
`consulta-juridica`, en los dos modos.
