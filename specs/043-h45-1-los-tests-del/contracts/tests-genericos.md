# Contrato: los cuatro tests, genéricos (FR-001 a FR-009)

Cada test: qué lee, qué compara, la forma de su mensaje con un ejemplo, y en qué rojo cae. Las líneas de código que se
citan entre comillas son las que fija este contrato: el quickstart (Q8) las usa para aplicar sus mutantes. Los mensajes
los lee quien escribe un paquete, persona o run, una vez por `make ci` en rojo; ninguna skill los pide.

## §1 · `TestCoberturaDelRepositorio` (`internal/skills/cobertura_test.go`; FR-001 a FR-004)

**Lee**: `data/cobertura.yaml` con `LeerCobertura` y `data/normas.yaml` con `normasDelRepositorioPorIdentificador` (sin
cambios).

**Forma**:

- `type tupla struct{ delPrecepto, apartado, queLaRige string }`, local al test; sin campo de procedimiento.
- `deLaBase []tupla`: una por fila de `ReguladosAparte`, en su orden, si su `vertical` es `paqueteDeLaBase` por la regla
  de hoy (la de `queLaRige` en `data/normas.yaml`; si está vacía, la de `delPrecepto`). El filtro va escrito en el test,
  sin `paqueteDeLaFila` ni `filasDeLaCobertura` (FR-002).
- `nueve := []tupla{…}`: las nueve de data-model §1, a mano y en su orden, con `const lpac = "BOE-A-2015-10565"`.
  Ninguna de las nueve cadenas de procedimiento de hoy queda en el cuerpo del test (FR-001).
- La comprobación, una subsecuencia (research D1):

  ```go
  casadas := 0
  for _, una := range nueve {
      if casadas < len(deLaBase) && deLaBase[casadas] == una {
          casadas++
      }
  }
  assert.Equal(t, len(deLaBase), casadas, "…")
  ```

- El comentario del test dice la regla: cada fila de lo administrativo leída es una de las nueve, en su orden; una fila
  que pasa a otro paquete deja de estar entre las leídas; que no falte ninguna fila ni cambie su texto lo comprueba
  `h32-paquetes` (FR-004).

**Mensaje** (con L06; 751 bytes medidos en el prototipo, research P1):

```text
../../data/cobertura.yaml: de sus filas de administrativo, [{BOE-A-2015-10565 DA 1ª.2.a BOE-A-2003-23186} {BOE-A-2015-10565 DA 1ª.2.b } … {BOE-A-2015-10565 112.4 }], solo las 5 primeras son, en su orden, de las nueve tuplas de contracts/datos-de-plazos.md §3 de H9, [{BOE-A-2015-10565 DA 1ª.2.a BOE-A-2003-23186} … {BOE-A-2015-10565 112.4 }]
```

Su tamaño lo dan las filas de lo regulado aparte de `data/cobertura.yaml` (hoy 9, unos 40 bytes por tupla y dos
listas), no el uso del kit: con meses de consultas es el mismo.

**En rojo**: L06 (fuera `data/cobertura.yaml:39`) → «solo las 5 primeras»; L07 (`LeerCobertura` con `Norma = nil`) →
`expected: 9`, `actual: 0`. **En verde**: con `main`; con L04 o L05 (los ve `h32-paquetes` con la sincronía hecha, y
sin ella `TestSkillsDelRepositorio`); en la copia con un paquete más.

## §2 · `mutantesDeLasOrdenes` (`internal/evals/conjunto_test.go`; FR-005)

**Forma**:

- `func mutantesDeLasOrdenes(t *testing.T) []mutanteDeUnControl`, con `t.Helper()`; la llamada de
  `TestControlesSobreLasReferenciasAMano` pasa `t` (`mutantesDeLasOrdenes(t)`).
- El defecto de `forma-en-una-referencia-con-otros-bloques` y de `forma-en-ninguna-referencia`:

  ```go
  sinSuFormaEnElSkillMd := []string{
      fmt.Sprintf("normativa: SKILL.md: línea %d, «%s»%s", lineaDeLaOrdenDelPasoDos(t), bashDelPasoDos, sinSuForma),
  }
  ```

- `func lineaDeLaOrdenDelPasoDos(t *testing.T) int`: lee `filepath.Join(raizDeLasSkills, directorioDeLasSkills,
  skillNormativa, ficheroDeLaSkill)` con `contenidoDelFichero`, recorre `strings.Lines` con su contador desde 1 y guarda
  el número de cada línea cuyo `strings.TrimSpace` es `bashDelPasoDos`; `require.Len(t, lineas, 1, "premisa: la orden de
  Bash del paso 2 está en una sola línea del SKILL.md de normativa")`. No usa `tramosDeCodigo` ni `ordenesDeLectura`.
- Ningún otro mutante cambia; el resto de la función, tal cual.

**Defecto esperado** (con `main`, byte a byte el de hoy):

```text
normativa: SKILL.md: línea 36, «kitlegal boe articulo BOE-A-1978-31229 a140 --json && kitlegal graph check BOE-A-1978-31229 a140 --json»: no la sigue su forma para PowerShell, «…; if ($LASTEXITCODE -eq 0) { … }», con el mismo …
```

En la copia con un paquete más, «línea 37». **En rojo**: P6; un control que da otra línea (`tramoDeCodigo{linea: numero
+ 1, …}` en `tramosDeCodigo`); la versión de `main` en la copia.

## §3 · `TestDefinicionDelJob/cuatro-evals-mas` (`internal/evals/definicion_test.go`; FR-006 a FR-008)

**Forma** (`probarCuatroEvalsMas`; el subtest conserva su nombre):

- Constantes: `skill = "plazos-y-territorio"`, `evalsDeLaCarpeta = 16`, `deDosEnDos = 2`. Sin `topeEnMinutos`,
  `evalsDeMas` ni `primeraCopia`, y sin la premisa del tope de 352.
- La definición, con `leerDefinicionDelJob(rutaDeLaDefinicionDelJob)`; la copia de las evals, con `os.CopyFS` en
  `t.TempDir()`, como hoy; la eval copiada, la primera de modo que decide de `LeerConjunto`, como hoy.
- En la carpeta de la skill de la copia, abierta con `os.OpenRoot`: `carpeta.Remove(eval.Fichero)` de cada eval de
  `LeerConjunto`, y `carpeta.WriteFile(fmt.Sprintf("%02d-copia-%s", numero+1, descripcion), contenido, 0o600)` para
  `numero` de 0 a 15; `require.Len(t, conLasCopias.Evals, evalsDeLaCarpeta, …)`.
- `peorCaso` comparado con `peorCasoEsperado`; `assert.GreaterOrEqual(t, tope, peor.Duracion(), …)` con la concurrencia
  del repositorio y `assert.Less(t, tope, peorADos.Duracion(), …)` con `plazos-y-territorio` a dos, con `tope :=
  time.Duration(leida.TopeEnMinutos) * time.Minute`.
- El comentario de la función dice lo que mide: 16 evals de modo, cabe a la concurrencia del repositorio y no a dos; que
  las del repositorio quepan es de `probarElTopeConLasEvalsDelRepositorio`.

**Mensajes**: «con 16 evals de plazos-y-territorio y la definición del repositorio, el tope cubre su peor caso, `<peor>`» y
«con 16 evals de plazos-y-territorio y la skill a dos, el tope no cubre su peor caso, `<peor>`», con `<peor>` la cuenta de
`peorCaso` (la fórmula de hoy, unos 200 bytes). Con T1: «"5h52m0s" is not greater than or equal to "7h47m49s"». 16 evals
fijas: no crece con el repositorio ni con el uso.

**En rojo**: T1 (`plazos-y-territorio` a dos en `.github/workflows/evals.yml:260` y en `ajustesDelTrabajo`); L29
(`Modos: []Modo{ModoOrden}` en `internal/evals/definicion.go:773`). **En verde**: con `main`; con evals de más en
`plazos-y-territorio` (entonces cae `/tope-con-las-evals-del-repositorio`, FR-008, que no cambia).

## §4 · `llevaNoCubierto` (`internal/evals/conjunto_test.go`; FR-009)

**Forma**:

```go
func llevaNoCubierto(eval Eval) bool {
    return eval.NoCubierto != NoCubiertoEsperado{} && !eval.NoCubierto.FueraDeLoDeclarado
}
```

Su comentario dice que deja fuera las de lo no declarado, que no llevan precepto. En `probarLosPreceptosDeLaCobertura`,
el mensaje del `require.NotEmpty` pasa a «%s tiene al menos una eval con no_cubierto de lo regulado aparte». El resto,
igual.

**Mensajes**: el de L33, el de hoy, una línea por eval (149 bytes la de la 07): «12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml:
no_cubierto.precepto BOE-A-2015-10565, bloque a999 no es el precepto de ninguna fila de data/cobertura.yaml». Crece con
las evals de lo regulado aparte de la skill, no con el uso.

**En rojo**: L33; todas las evals con `no_cubierto` de lo no declarado (`Should NOT be empty`); la versión de `main` en la
copia («95-fuera-de-lo-declarado.yaml: no_cubierto.precepto , bloque  no es…»).
