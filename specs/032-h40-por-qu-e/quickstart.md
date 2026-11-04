# Quickstart: validar H40

Cómo comprobar la entrega, desde la raíz del repositorio. Los escenarios 1 a 4 no usan red, modelo ni credencial: solo
leen el árbol y escriben en las cachés de Go y en temporales de `go test`, que borra al terminar; no cambian el árbol de
trabajo, el índice ni el historial. El 5 es `make ci`, con lo que declara. El 6 es el de la persona, fuera de cualquier
run. Contratos: contracts/variantes.md,
contracts/arnes.md, contracts/tabla.md y contracts/arnes-e2e.md.

## 1. Las variantes (historia 1; FR-001 a FR-007)

```sh
go test -count=1 -run '^(TestVariantesDeNormativa|TestLaboratorioFueraDelProducto)$' ./internal/evals/
diff -r skills/normativa laboratorio/normativa/normativa-control/skills/normativa
tail -n +17 laboratorio/normativa/normativa-remisiones-en-orden/skills/normativa/SKILL.md | wc -m
tail -n +17 laboratorio/normativa/normativa-cuerpo-corto/skills/normativa/SKILL.md | wc -m
tail -n +17 laboratorio/normativa/normativa-testigo-en-referencia/skills/normativa/SKILL.md | wc -m
```

Esperado: `ok` en los dos tests; `diff` sin salida y con código 0 (control es `skills/normativa`); y 11084, 3298 y
10973 caracteres de cuerpo, las cifras de research V2 y V3 con el `SKILL.md` de hoy (cambian si cambia
`skills/normativa`, y entonces el primer test dice qué variante hay que igualar).

## 2. El arnés, sin modelo (historia 2; FR-008, FR-010, FR-011)

```sh
go test -count=1 -run '^(TestArnesDeNormativa|TestArnesDelLaboratorio|TestRecetasDelSondeo)$' ./internal/evals/
```

Esperado: `ok`. `TestArnesDeNormativa` lanza el arnés con un sustituto de `claude` en temporales del test: con cada
petición inválida de contracts/arnes.md §5, ninguna sesión; con los dos lanzamientos de la aceptación con menos
sesiones, `normativa-control` en los dos modos y las otras tres variantes en herramienta, `SESIONES/4` por eval en cada
variante y modo, cada una con solo su variante instalada y el servidor solo en herramienta. `TestRecetasDelSondeo` ejecuta `make -n`, que solo imprime.

## 3. La tabla y la regla testigo (historia 3; FR-020 a FR-023, FR-030)

```sh
go test -count=1 -run '^(TestTablaDeNormativa|TestReglaTestigoDeNormativa|TestTablaDelLaboratorio|TestReglaTestigo)$' ./internal/evals/
make -s laboratorio-tabla DIRECTORIO=internal/evals/testdata/laboratorio/normativa/sondeo | cmp - internal/evals/testdata/laboratorio/normativa/tabla.md
make -s laboratorio-tabla DIRECTORIO=internal/evals/testdata/laboratorio/sondeo | cmp - internal/evals/testdata/laboratorio/tabla.md
```

Esperado: `ok`, y los dos `cmp` sin salida y con código 0: la tabla de las sesiones fijas de `normativa`, con las
columnas de contracts/tabla.md §3, y la de H39, igual que antes del hito. La tercera orden da ya hoy código 0
(research V13).

## 4. La suite de aceptación congelada (FR-020, FR-021, FR-023; SC-005, SC-007)

Activada por el workflow como `internal/app/testdata/script/h40-laboratorio-tabla.txtar`:

```sh
go test -count=1 -run '^TestEntregaDelHito$/^h40-laboratorio-tabla$' ./internal/app/
```

Esperado: `ok`. El guion ejecuta `laboratorio tabla` sobre las sesiones fijas de H39 de `main` y sobre las de
`normativa` que lleva dentro, y compara cada salida con su tabla esperada (contracts/arnes-e2e.md §2).

## 5. Todo `make ci`

```sh
make ci
```

Esperado: verde. Incluye lo anterior y comprueba que `skills/` no cambia (`make skills-check`,
`TestSkillsDelRepositorio`). Deja en la raíz `coverage.out` y `coverage-integration.out`, que `.gitignore` excluye, y
`make vuln` consulta la base de vulnerabilidades de `govulncheck`, como en cualquier `make ci`.

## 6. El sondeo de la persona (historia 4; FR-040 a FR-042)

Solo una persona, con la propuesta de cambio abierta y antes de fusionar; nunca un run (ADR 0032). Abre 200 sesiones con
modelo, consume la suscripción de quien lo lanza, pide la contraseña del llavero en macOS y crea los dos `DIRECTORIO`,
fuera del repositorio:

```sh
make evals-laboratorio VARIANTES=normativa-control MODOS=orden,herramienta SESIONES=40 \
  MODELO=claude-sonnet-5-5 CLAUDE_CODE=<versión instalada> CONCURRENCIA=5 DIRECTORIO=../kitlegal-sondeos/h40/a
make evals-laboratorio \
  VARIANTES=normativa-remisiones-en-orden,normativa-cuerpo-corto,normativa-testigo-en-referencia MODOS=herramienta \
  SESIONES=40 MODELO=claude-sonnet-5-5 CLAUDE_CODE=<versión instalada> CONCURRENCIA=5 DIRECTORIO=../kitlegal-sondeos/h40/b
make laboratorio-tabla DIRECTORIO=../kitlegal-sondeos/h40
```

`<versión instalada>` es el primer campo de `claude --version`; con otra, el arnés no abre nada. Esperado: cinco filas
de 40 sesiones con el modelo y la versión anotados. La persona deja la tabla en `evidencias/`, escribe con sus cifras el
ADR de FR-041 en `docs/ADR/` y añade los dos a la propuesta de cambio antes de fusionarla (FR-042, SC-009).
