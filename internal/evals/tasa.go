package evals

import (
	"math/big"
	"sort"
)

// La regla de la tasa del umbral conjunto de una clase del juez que decide
// (FR-024, SC-010 de H30; data-model §9 y research D12 de H30; ADR 0038, punto
// 3): una ejecución sale en rojo sin motivo como mucho 1 de cada 50 veces, sea
// cual sea el número de respuestas medidas. La tasa de referencia, pref, se
// deriva del umbral de la clase en clases.yaml y de las respuestas que juzga la
// ejecución completa, y de ella sale el límite de marcadas para cualquier
// número de respuestas.
//
// Las probabilidades se comparan sin redondeo: con pref = q/D, D una diez mil,
// P(Bin(n, pref) > k) ≤ a/b es (b − a)·D^n ≤ b·S, con S la suma, de i = 0 a k,
// de C(n, i)·q^i·(D − q)^(n − i), todo enteros.
const (
	// diezmilesimasDeLaUnidad es D: pref se redondea hacia abajo a centésimas
	// de punto porcentual, que son diezmilésimas.
	diezmilesimasDeLaUnidad = 10000

	// rojosSinMotivo de cada ejecucionesDelCriterio es el criterio, 1/50.
	rojosSinMotivo         = 1
	ejecucionesDelCriterio = 50
)

// limiteDeLaTasa es la regla de la tasa de una clase que decide en un plan: lo
// que se deriva de su umbral y de las respuestas que juzga la ejecución
// completa (data-model §9 de H30).
type limiteDeLaTasa struct {
	// respuestas es N: las del modelo que decide que una ejecución completa del
	// plan da a juzgar en sus modos (respuestasQueSeJuzganEnSusModos).
	respuestas int

	// limite es K, el límite de hoy: el mayor número de marcadas de las N con
	// el que el umbral de la clase se cumple con la comparación del job.
	limite int

	// referencia es pref en diezmilésimas: la mayor con la que
	// P(Bin(N, pref) > K) ≤ 1/50.
	referencia int
}

// nuevoLimiteDeLaTasa es la regla de la tasa de una clase con el umbral dado
// en un plan que da a juzgar las respuestas dadas. P(Bin(N, p) > K) no baja al
// subir p, y con p = 0 es 0: la mayor diezmilésima que cumple el criterio se
// busca por bisección entre 0, que lo cumple siempre, y la unidad.
func nuevoLimiteDeLaTasa(umbral float64, respuestas int) limiteDeLaTasa {
	limite := mayorMedidaQueCumple(umbral, respuestas)

	primeraQueNoCumple := sort.Search(diezmilesimasDeLaUnidad+1, func(referencia int) bool {
		return !cumpleElCriterio(respuestas, referencia, limite)
	})

	return limiteDeLaTasa{respuestas: respuestas, limite: limite, referencia: primeraQueNoCumple - 1}
}

// para es k(n), el límite de marcadas con n respuestas medidas: el k más
// estricto con P(Bin(n, pref) > k) ≤ 1/50. Con k = n la probabilidad es 0, así
// que nunca pasa de n; y con todas las respuestas del plan es K, el límite de
// hoy (TestLimiteDeLaTasa).
func (l limiteDeLaTasa) para(respuestas int) int {
	cota := cotaDelCriterio(respuestas)
	hastaK := new(big.Int)

	for k := range respuestas {
		hastaK.Add(hastaK, terminoDeLaBinomial(respuestas, l.referencia, k))

		if dentroDelCriterio(cota, hastaK) {
			return k
		}
	}

	return respuestas
}

// cumpleElCriterio dice si P(Bin(n, q/D) > k) ≤ 1/50.
func cumpleElCriterio(respuestas, referencia, limite int) bool {
	hastaK := new(big.Int)

	for i := 0; i <= limite && i <= respuestas; i++ {
		hastaK.Add(hastaK, terminoDeLaBinomial(respuestas, referencia, i))
	}

	return dentroDelCriterio(cotaDelCriterio(respuestas), hastaK)
}

// cotaDelCriterio es (b − a)·D^n, el lado izquierdo del criterio con n
// respuestas.
func cotaDelCriterio(respuestas int) *big.Int {
	cota := new(big.Int).Exp(big.NewInt(diezmilesimasDeLaUnidad), big.NewInt(int64(respuestas)), nil)

	return cota.Mul(cota, big.NewInt(ejecucionesDelCriterio-rojosSinMotivo))
}

// dentroDelCriterio dice si (b − a)·D^n ≤ b·S, con S la suma hasta k de los
// términos de la binomial: si P(Bin(n, pref) > k) ≤ a/b.
func dentroDelCriterio(cota, hastaK *big.Int) bool {
	return cota.Cmp(new(big.Int).Mul(hastaK, big.NewInt(ejecucionesDelCriterio))) <= 0
}

// terminoDeLaBinomial es C(n, i)·q^i·(D − q)^(n − i): D^n veces P(Bin(n, q/D)
// = i).
func terminoDeLaBinomial(respuestas, referencia, i int) *big.Int {
	termino := new(big.Int).Binomial(int64(respuestas), int64(i))
	termino.Mul(termino, new(big.Int).Exp(big.NewInt(int64(referencia)), big.NewInt(int64(i)), nil))

	return termino.Mul(termino, new(big.Int).Exp(big.NewInt(int64(diezmilesimasDeLaUnidad-referencia)),
		big.NewInt(int64(respuestas-i)), nil))
}

// mayorMedidaQueCumple es el mayor número de marcadas, de 0 a total, con el que
// un umbral de esas respuestas se cumple con la comparación del job
// (compararUmbral): su proporción sobre total, 0 si total es 0, no pasa del
// umbral. 0 si ninguno se cumple, que con un umbral de 0 a 1, el que admite
// clases.yaml, no pasa.
func mayorMedidaQueCumple(umbral float64, total int) int {
	for marcadas := total; marcadas > 0; marcadas-- {
		if compararUmbral(Umbral{Medida: float64(marcadas), Total: &total, Umbral: umbral}).Cumple {
			return marcadas
		}
	}

	return 0
}

// respuestasQueSeJuzganEnSusModos son las respuestas que una ejecución
// completa del plan da a juzgar en sus modos: las de sus sesiones que se
// juzgan (seJuzga), del modelo que decide en las evals que activan la skill,
// sin las de las evals sin binario ni servidor, que no son de ningún modo y no
// cuentan en ningún umbral. Es N de la regla de la tasa: el total del umbral
// conjunto de una clase cuando se miden todas.
func (p PlanDeEvals) respuestasQueSeJuzganEnSusModos() int {
	respuestas := 0

	for _, sesion := range p.Sesiones() {
		if sesion.Modo != "" && seJuzga(p, sesion) {
			respuestas++
		}
	}

	return respuestas
}
