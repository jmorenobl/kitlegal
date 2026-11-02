package evals

import (
	"fmt"
	"math/big"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// probabilidadDelCriterio es la del criterio del ADR 0038, punto 3, que
// mantiene FR-024 de H30: una ejecución sale en rojo sin motivo como mucho 1 de
// cada 50 veces. El test la escribe aparte de tasa.go, como racional exacto.
var probabilidadDelCriterio = big.NewRat(1, 50)

// TestLimiteDeLaTasa fija la regla de la tasa del umbral conjunto de la clase
// que decide (FR-024, SC-010, SC-013 de H30; data-model §9; research D12), con
// la definición del job y las evals del repositorio: de cada skill del job con
// juez, N son las respuestas del modelo que decide que su plan da a juzgar en
// los dos modos, contadas aquí aparte —las evals que activan la skill y tienen
// binario o servidor, por las repeticiones del job y por los dos modos— y
// comparadas con las que da el plan; y de su clase que decide, con el umbral de
// su clases.yaml:
//
//   - pref es la mayor tasa en diezmilésimas (centésimas de punto porcentual)
//     con P(Bin(N, pref) > K) ≤ 1/50, con K el mayor entero con K/N ≤ umbral,
//     la comparación del job: con ella se cumple y con la siguiente diezmilésima
//     ya no;
//   - para cada n de 1 a N, P(Bin(n, pref) > k(n)) ≤ 1/50 y, si k(n) > 0,
//     P(Bin(n, pref) > k(n) − 1) > 1/50: k(n) es el límite más estricto que
//     cumple el criterio;
//   - y k(N) es K, el límite de hoy.
//
// Las probabilidades son racionales exactos de math/big. El t.Log de cada skill
// da pref, N y k(N), las cifras del ADR (T016 de H30).
//
// Y el mutante de SC-010, en memoria y sin tocar el fichero: el clases.yaml de
// normativa con umbral: 0.017 baja pref, no deja ningún k(n) por encima
// del de hoy y da k(N) = 2 (2 de 120), el mayor entero con K/N ≤ 0.017.
func TestLimiteDeLaTasa(t *testing.T) {
	t.Parallel()

	delJob, err := leerDefinicionDelJob(rutaDeLaDefinicionDelJob)
	require.NoError(t, err)

	conJuez := 0

	for _, skill := range delJob.Skills {
		conjunto, err := LeerConjunto(filepath.Join(directorioDeEvals, skill))
		require.NoError(t, err)

		if conjunto.Juez == nil {
			continue
		}

		conJuez++

		respuestas := respuestasDelPlanDelJob(t, delJob, conjunto)

		for _, clase := range conjunto.Juez.Clases {
			if !clase.Decide {
				continue
			}

			t.Run(skill+"/"+clase.Nombre, func(t *testing.T) {
				t.Parallel()

				limite := nuevoLimiteDeLaTasa(clase.Umbral, respuestas)
				exigirElLimiteDeLaTasa(t, limite, clase.Umbral, respuestas)

				t.Logf("%s, %s (umbral %s): pref = %s, N = %d, k(N) = %d", skill, clase.Nombre,
					strconv.FormatFloat(clase.Umbral, 'f', -1, 64), tasaEscrita(limite.referencia), respuestas,
					limite.para(respuestas))
			})
		}
	}

	require.NotZero(t, conJuez, "alguna skill del job tiene juez")

	t.Run("estrechar-el-umbral-estrecha-el-limite", func(t *testing.T) {
		t.Parallel()

		exigirElMutanteDelUmbral(t, delJob)
	})
}

// respuestasDelPlanDelJob son las respuestas del modelo que decide que el plan
// del job de la skill da a juzgar en los dos modos, contadas aparte: cada eval
// que activa la skill y no es de las sin binario ni servidor, por las
// repeticiones del job y por los dos modos. Exige que el plan dé las mismas.
func respuestasDelPlanDelJob(t *testing.T, delJob DefinicionDelJob, conjunto Conjunto) int {
	t.Helper()

	const modosDelJob = 2

	queSeJuzgan := 0

	for _, eval := range conjunto.Evals {
		if eval.Activa && !eval.SinBinarioNiServidor {
			queSeJuzgan++
		}
	}

	respuestas := queSeJuzgan * delJob.Repeticiones * modosDelJob

	plan := PlanDeEvals{
		Evals:               conjunto.Evals,
		ModeloQueDecide:     delJob.ModeloQueDecide,
		ModelosInformativos: delJob.ModelosInformativos,
		Repeticiones:        delJob.Repeticiones,
		Modos:               []Modo{ModoOrden, ModoHerramienta},
	}
	require.NoError(t, plan.Comprobar())
	require.Equal(t, respuestas, plan.respuestasQueSeJuzganEnSusModos(),
		"el plan da a juzgar en sus modos las respuestas del modelo que decide de las evals que activan la skill")

	return respuestas
}

// exigirElLimiteDeLaTasa exige la regla de TestLimiteDeLaTasa al límite de una
// clase con su umbral y las N respuestas de su plan.
func exigirElLimiteDeLaTasa(t *testing.T, limite limiteDeLaTasa, umbral float64, respuestas int) {
	t.Helper()

	deHoy := mayorMarcadasDentro(umbral, respuestas)

	assert.Equal(t, respuestas, limite.respuestas, "N")
	assert.Equal(t, deHoy, limite.limite, "K, el mayor entero con K/N ≤ umbral")

	referencia := tasaEnDiezmilesimas(limite.referencia)
	assert.LessOrEqual(t, colaDeLaBinomial(respuestas, referencia, deHoy).Cmp(probabilidadDelCriterio), 0,
		"P(Bin(%d, %s) > %d) ≤ 1/50", respuestas, tasaEscrita(limite.referencia), deHoy)

	if limite.referencia < diezmilesimasDeLaUnidad {
		siguiente := tasaEnDiezmilesimas(limite.referencia + 1)
		assert.Positive(t, colaDeLaBinomial(respuestas, siguiente, deHoy).Cmp(probabilidadDelCriterio),
			"con %s, la diezmilésima siguiente, P(Bin(%d, p) > %d) ya pasa de 1/50: pref es la mayor",
			tasaEscrita(limite.referencia+1), respuestas, deHoy)
	}

	for n := 1; n <= respuestas; n++ {
		k := limite.para(n)

		assert.LessOrEqual(t, colaDeLaBinomial(n, referencia, k).Cmp(probabilidadDelCriterio), 0,
			"P(Bin(%d, %s) > %d) ≤ 1/50", n, tasaEscrita(limite.referencia), k)

		if k > 0 {
			assert.Positive(t, colaDeLaBinomial(n, referencia, k-1).Cmp(probabilidadDelCriterio),
				"P(Bin(%d, %s) > %d) pasa de 1/50: k(%d) = %d es el más estricto", n,
				tasaEscrita(limite.referencia), k-1, n, k)
		}
	}

	assert.Equal(t, deHoy, limite.para(respuestas), "k(N) es el límite de hoy, %d de %d", deHoy, respuestas)
}

// exigirElMutanteDelUmbral exige el mutante de SC-010: el clases.yaml de
// normativa, leído del repositorio y con el umbral de su clase que decide
// cambiado en memoria a 0.017, da un límite con pref más baja, ningún k(n) por
// encima del del umbral vigente y k(N) más estricto, el de 0.017: 2 de 120.
func exigirElMutanteDelUmbral(t *testing.T, delJob DefinicionDelJob) {
	t.Helper()

	const (
		skill     = "normativa"
		estrecho  = 0.017
		escritoEn = "umbral: "
	)

	conjunto, err := LeerConjunto(filepath.Join(directorioDeEvals, skill))
	require.NoError(t, err)
	require.NotNil(t, conjunto.Juez, "%s tiene juez", skill)

	respuestas := respuestasDelPlanDelJob(t, delJob, conjunto)

	contenido := string(contenidoDelFichero(t, filepath.Join(directorioDeEvals, skill, clasesEnElJuez)))
	vigente := unicaClaseQueDecide(t, conjunto.Juez.Clases)

	deHoy := escritoEn + strconv.FormatFloat(vigente.Umbral, 'f', -1, 64)
	require.Equal(t, 1, strings.Count(contenido, deHoy), "el umbral de la clase que decide está escrito una vez: %s", deHoy)

	mutadas, err := leerClasesDelJuez([]byte(strings.Replace(contenido, deHoy,
		escritoEn+strconv.FormatFloat(estrecho, 'f', -1, 64), 1)))
	require.NoError(t, err)

	mutada := unicaClaseQueDecide(t, mutadas)
	require.InDelta(t, estrecho, mutada.Umbral, 0, "el mutante lleva el umbral estrechado")

	original := nuevoLimiteDeLaTasa(vigente.Umbral, respuestas)
	limite := nuevoLimiteDeLaTasa(mutada.Umbral, respuestas)

	exigirElLimiteDeLaTasa(t, limite, mutada.Umbral, respuestas)

	assert.Less(t, limite.referencia, original.referencia, "con el umbral estrechado, pref baja")
	assert.Less(t, limite.para(respuestas), original.para(respuestas), "con el umbral estrechado, k(N) baja")

	for n := 1; n <= respuestas; n++ {
		assert.LessOrEqual(t, limite.para(n), original.para(n), "con el umbral estrechado, k(%d) no sube", n)
	}

	t.Logf("%s con umbral %s: pref = %s, N = %d, k(N) = %d", skill, strconv.FormatFloat(estrecho, 'f', -1, 64),
		tasaEscrita(limite.referencia), respuestas, limite.para(respuestas))
}

// unicaClaseQueDecide es la única clase que decide de las dadas.
func unicaClaseQueDecide(t *testing.T, clases []ClaseDelJuez) ClaseDelJuez {
	t.Helper()

	posicion := slices.IndexFunc(clases, func(clase ClaseDelJuez) bool { return clase.Decide })
	require.GreaterOrEqual(t, posicion, 0, "hay una clase que decide")
	require.False(t, slices.ContainsFunc(clases[posicion+1:], func(clase ClaseDelJuez) bool { return clase.Decide }),
		"solo una clase decide")

	return clases[posicion]
}

// mayorMarcadasDentro es el mayor m de 0 a total con m/total ≤ umbral, con la
// comparación en coma flotante del job (compararUmbral), escrito aquí aparte de
// tasa.go.
func mayorMarcadasDentro(umbral float64, total int) int {
	for m := total; m > 0; m-- {
		if float64(m)/float64(total) <= umbral {
			return m
		}
	}

	return 0
}

// tasaEnDiezmilesimas es la tasa de las diezmilésimas dadas, como racional.
func tasaEnDiezmilesimas(diezmilesimas int) *big.Rat {
	return big.NewRat(int64(diezmilesimas), diezmilesimasDeLaUnidad)
}

// tasaEscrita es la tasa de las diezmilésimas dadas en tanto por ciento, con
// dos decimales y coma, como la escribe el ADR: 85 es «0,85 %».
func tasaEscrita(diezmilesimas int) string {
	return fmt.Sprintf("%d,%02d %%", diezmilesimas/100, diezmilesimas%100)
}

// colaDeLaBinomial es P(Bin(n, p) > k), en racionales exactos: 1 menos la suma,
// de i = 0 a k, de C(n, i)·p^i·(1 − p)^(n − i).
func colaDeLaBinomial(n int, p *big.Rat, k int) *big.Rat {
	uno := big.NewRat(1, 1)
	complementaria := new(big.Rat).Sub(uno, p)
	hastaK := new(big.Rat)

	for i := 0; i <= k && i <= n; i++ {
		termino := new(big.Rat).SetInt(new(big.Int).Binomial(int64(n), int64(i)))
		termino.Mul(termino, potenciaRacional(p, i))
		termino.Mul(termino, potenciaRacional(complementaria, n-i))
		hastaK.Add(hastaK, termino)
	}

	return new(big.Rat).Sub(uno, hastaK)
}

// potenciaRacional es r elevado a e, con e ≥ 0.
func potenciaRacional(r *big.Rat, e int) *big.Rat {
	exponente := big.NewInt(int64(e))

	return new(big.Rat).SetFrac(new(big.Int).Exp(r.Num(), exponente, nil), new(big.Int).Exp(r.Denom(), exponente, nil))
}
