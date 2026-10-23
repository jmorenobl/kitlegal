package evals

import (
	"cmp"
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// extensionDeUnaEval es la del fichero de una eval, la que su nombre en el
// conjunto no lleva y en el origen sí, como en las series del informe.
const extensionDeUnaEval = ".yaml"

// calculoDeLaDecision es lo que la decisión de un trabajo lee del checkout
// (data-model §5 y §6 de H30): la elección —el conjunto de cada eval, lo común y
// los árboles de la cabeza y de la base— y la huella de cada eval, una por
// conjunto y en su orden; ninguna si algún conjunto no se puede calcular, que
// las elige todas con sin_conjunto (FR-005).
type calculoDeLaDecision struct {
	eleccion calculoDeLaEleccion
	huellas  []string
}

// leerElCalculoDeLaDecision lee del checkout de la raíz lo que la decisión de
// un trabajo necesita (contracts/eleccion-y-huella.md §1 de H30): antes de nada
// exige los fijados de la huella; después, la elección, con el árbol de la base
// de fusión de base si no está vacía —con la suite completa no se elige nada—,
// y la huella de cada eval, si todos sus conjuntos se calculan. Es un error que
// un fijado no pueda ir en la huella o que el checkout no se pueda leer.
func leerElCalculoDeLaDecision(ctx context.Context, raiz, base string, fijados fijadosDeLaHuella) (
	calculoDeLaDecision, error,
) {
	if err := fijados.validar(); err != nil {
		return calculoDeLaDecision{}, err
	}

	eleccion, huella, err := leerElCheckout(ctx, raiz, base, fijados)
	if err != nil {
		return calculoDeLaDecision{}, err
	}

	calculo := calculoDeLaDecision{eleccion: eleccion}
	if eleccion.sinConjunto() != "" {
		return calculo, nil
	}

	calculo.huellas, err = huella.huellas()
	if err != nil {
		return calculoDeLaDecision{}, err
	}

	return calculo, nil
}

// leerElCalculoDeLaEleccion lee del checkout de la raíz lo que la elección de la
// apertura de una propuesta necesita: el conjunto de cada eval, lo común y los
// árboles de la cabeza y de la base de fusión de base. Es un error que el
// checkout no se pueda leer.
func leerElCalculoDeLaEleccion(ctx context.Context, raiz, base string) (calculoDeLaEleccion, error) {
	eleccion, _, err := leerElCheckout(ctx, raiz, base, fijadosDeLaHuella{})

	return eleccion, err
}

// leerElCheckout lee de la raíz, con un solo go list, el cálculo de la huella
// de las evals —el conjunto de cada una, el árbol de la cabeza, lo común y el
// resto de go.mod y go.sum— y, con él, el de la elección, con el árbol de la
// base de fusión de base y HEAD si base no está vacía (research D5 de H30).
func leerElCheckout(ctx context.Context, raiz, base string, fijados fijadosDeLaHuella) (
	calculoDeLaEleccion, calculoDeLaHuella, error,
) {
	grafo, err := leerGrafoDePaquetes(ctx, raiz)
	if err != nil {
		return calculoDeLaEleccion{}, calculoDeLaHuella{}, err
	}

	huella, err := leerElCalculoDeLaHuella(ctx, raiz, grafo, fijados)
	if err != nil {
		return calculoDeLaEleccion{}, calculoDeLaHuella{}, err
	}

	eleccion := calculoDeLaEleccion{conjuntos: huella.conjuntos, comun: huella.comun, cabeza: huella.cabeza}
	if base == "" {
		return eleccion, huella, nil
	}

	fusion, err := baseDeFusion(ctx, raiz, base)
	if err != nil {
		return calculoDeLaEleccion{}, calculoDeLaHuella{}, err
	}

	eleccion.base, err = leerArbolDeLaBase(ctx, raiz, fusion)
	if err != nil {
		return calculoDeLaEleccion{}, calculoDeLaHuella{}, err
	}

	return eleccion, huella, nil
}

// modoDeLaDecision es el modo de la ejecución que da el env de la tanda
// (contracts/job-de-evals.md §1 de H30): completa o reutilizacion. Cualquier
// otro valor es un error que lo nombra.
func modoDeLaDecision(valor string) (modoDelOrigen, error) {
	switch modo := modoDelOrigen(valor); modo {
	case origenDeLaSuiteCompleta, origenConReutilizacion:
		return modo, nil
	default:
		return "", fmt.Errorf("el modo de la ejecución %q no es %s ni %s", valor, origenDeLaSuiteCompleta,
			origenConReutilizacion)
	}
}

// eleccionConLaPlataforma es la elección de la propuesta con los ficheros que
// da la plataforma (FR-006 a FR-008 de H30) o, si no los da, la de todas con el
// motivo plataforma, sin rutas, con el error de su lectura: la orden y su
// salida de error, o el plazo del contexto agotado (research D17 de H30).
func eleccionConLaPlataforma(
	ctx context.Context, plataforma lecturasDeLaPlataforma, propuesta int, calculo calculoDeLaEleccion,
) (eleccionDeLaPropuesta, *errorDeLaPlataformaEnElOrigen) {
	cambios, err := plataforma.ficherosDeLaPropuesta(ctx, propuesta)
	if err != nil {
		fallida := errorDeLaLectura(ctx, err, ordenDeLosFicherosDeLaPropuesta,
			variableDeLaPropuesta+"="+strconv.Itoa(propuesta))

		return calculo.elegirSinLosFicheros(errorDeLaPlataforma{orden: fallida.Orden, causa: fallida.Error}), fallida
	}

	return calculo.elegir(cambios), nil
}

// eligeAlguna es la decisión de la apertura (FR-016 de H30): elige=si si la
// propuesta elige alguna eval —también si las elige todas por un conjunto
// incalculable o porque la plataforma no dio sus ficheros—, y elige=no si no
// elige ninguna, y entonces no corre ningún trabajo de evals.
func (e eleccionDeLaPropuesta) eligeAlguna() bool {
	return len(e.elegidas) > 0
}

// escribirLaApertura añade a la salida del paso, el GITHUB_OUTPUT del trabajo
// cambios, la línea elige=si o elige=no (contracts/job-de-evals.md §2 de H30).
func escribirLaApertura(ruta string, elige bool) error {
	valor := "no"
	if elige {
		valor = "si"
	}

	fichero, err := os.OpenFile(filepath.Clean(ruta), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err == nil {
		_, err = fmt.Fprintf(fichero, "elige=%s\n", valor)
		err = errors.Join(err, fichero.Close())
	}

	if err != nil {
		return fmt.Errorf("la elección de la apertura no se puede escribir en %s: %w", ruta, err)
	}

	return nil
}

// trabajoADecidir es lo que recibe la decisión del trabajo evals (<skill>) de
// una skill (contracts/job-de-evals.md §2 de H30): la skill, el modo de la
// ejecución, el número de la propuesta —sin uso en la suite completa—, lo que
// se leyó del checkout, la plataforma y el instante de la decisión, con el que
// la búsqueda compara la edad de cada medida de main y de cada procedencia.
type trabajoADecidir struct {
	skill      string
	modo       modoDelOrigen
	propuesta  int
	calculo    calculoDeLaDecision
	plataforma lecturasDeLaPlataforma
	instante   time.Time
}

// evalDelTrabajo es una eval de la skill del trabajo: su nombre en el
// conjunto, sin .yaml; su fichero, el nombre del origen; y su huella, vacía si
// algún conjunto no se puede calcular.
type evalDelTrabajo struct {
	nombre  string
	fichero string
	huella  string
}

// decidirElTrabajo es la decisión del trabajo, el bloque origen sin
// sesiones_abiertas (data-model §8 de H30; FR-012 a FR-014, FR-018), con una
// entrada por eval de la skill en orden de fichero:
//
//   - en la suite completa, todas medidas con suite_completa, sin leer nada de
//     la plataforma;
//   - con reutilización, en un plazo de 10 minutos para todo lo que lee de la
//     plataforma, las que elige la propuesta (eleccionConLaPlataforma) medidas
//     con su motivo y sus rutas, sin buscarlas; y cada una de las demás, la
//     reutilizada de la medida de main más reciente con su huella o, sin ella,
//     medida con sin_resultado_de_main (buscarEnLasMedidasDeMain). El error de
//     la plataforma es el de los ficheros de la propuesta, que no busca nada
//     porque las elige todas, o el que paró la búsqueda.
//
// Cada una lleva su huella, salvo sin conjunto, y el origen, la causa por la
// que algún conjunto no se pudo calcular.
func decidirElTrabajo(ctx context.Context, trabajo trabajoADecidir) *origenDeLasEvals {
	origen := &origenDeLasEvals{Modo: trabajo.modo, SinConjunto: trabajo.calculo.eleccion.sinConjunto()}
	evals := trabajo.calculo.evalsDeLaSkill(trabajo.skill)

	if trabajo.modo == origenConReutilizacion {
		origen.Evals, origen.ErrorDeLaPlataforma = trabajo.conReutilizacion(ctx, evals)
	} else {
		for _, eval := range evals {
			origen.Evals = append(origen.Evals, origenDeEval{
				Eval: eval.fichero, Huella: eval.huella, Medida: true, Motivo: motivoSuiteCompleta,
			})
		}
	}

	for _, eval := range origen.Evals {
		if eval.Medida {
			origen.EvalsMedidas++
		} else {
			origen.EvalsReutilizadas++
		}
	}

	return origen
}

// conReutilizacion son las entradas del origen de las evals del trabajo con
// reutilización, en su orden, y el error de la plataforma, si lo hubo.
func (t trabajoADecidir) conReutilizacion(
	ctx context.Context, evals []evalDelTrabajo,
) ([]origenDeEval, *errorDeLaPlataformaEnElOrigen) {
	ctx, cancelar := context.WithTimeout(ctx, plazoDeLaPlataforma)
	defer cancelar()

	eleccion, deLosFicheros := eleccionConLaPlataforma(ctx, t.plataforma, t.propuesta, t.calculo.eleccion)

	elegidas := map[string]evalElegida{}

	for _, elegida := range eleccion.elegidas {
		if elegida.skill == t.skill {
			elegidas[elegida.eval] = elegida
		}
	}

	var buscadas []evalABuscar

	for _, eval := range evals {
		if _, elegida := elegidas[eval.nombre]; !elegida {
			buscadas = append(buscadas, evalABuscar{fichero: eval.fichero, huella: eval.huella})
		}
	}

	halladas, deLaBusqueda := buscarEnLasMedidasDeMain(ctx, t.plataforma, t.skill, buscadas, t.instante)

	porFichero := make(map[string]origenDeEval, len(halladas))
	for _, hallada := range halladas {
		porFichero[hallada.Eval] = hallada
	}

	decididas := make([]origenDeEval, 0, len(evals))

	for _, eval := range evals {
		elegida, esElegida := elegidas[eval.nombre]
		if !esElegida {
			decididas = append(decididas, porFichero[eval.fichero])

			continue
		}

		decididas = append(decididas, origenDeEval{
			Eval: eval.fichero, Huella: eval.huella, Medida: true, Motivo: elegida.motivo, Rutas: elegida.rutas,
		})
	}

	return decididas, cmp.Or(deLosFicheros, deLaBusqueda)
}

// evalsDeLaSkill son las evals de la skill, en el orden de sus conjuntos, que
// es el de fichero, cada una con su huella si las hay.
func (c calculoDeLaDecision) evalsDeLaSkill(skill string) []evalDelTrabajo {
	var evals []evalDelTrabajo

	for indice, conjunto := range c.eleccion.conjuntos {
		if conjunto.skill != skill {
			continue
		}

		eval := evalDelTrabajo{nombre: conjunto.eval, fichero: conjunto.eval + extensionDeUnaEval}
		if c.huellas != nil {
			eval.huella = c.huellas[indice]
		}

		evals = append(evals, eval)
	}

	return evals
}

// lineasDeLaDecision son las del registro de la decisión de un trabajo
// (contracts/job-de-evals.md §2 de H30): la del error de la plataforma, si lo
// hubo, con la orden y su salida de error, delante; y una por eval, en su
// orden, como en informe.md (lineaDelOrigen): medida con su motivo o
// reutilizada con su procedencia.
func lineasDeLaDecision(origen *origenDeLasEvals) []string {
	var lineas []string

	if fallida := origen.ErrorDeLaPlataforma; fallida != nil {
		lineas = append(lineas, "error de la plataforma: "+fallida.Orden+": "+fallida.Error)
	}

	for _, eval := range origen.Evals {
		lineas = append(lineas, lineaDelOrigen(eval, origen.ErrorDeLaPlataforma != nil))
	}

	return lineas
}

// escribirElFicheroDeLaDecision escribe en la ruta el fichero de la decisión:
// el origen en JSON, sangrado como informe.json, sin sesiones_abiertas, que la
// decisión no tiene (contracts/informe-del-job.md §2 de H30).
func escribirElFicheroDeLaDecision(ruta string, origen *origenDeLasEvals) error {
	sinSesiones := *origen
	sinSesiones.SesionesAbiertas = nil

	codificado, err := json.Marshal(&sinSesiones, jsontext.WithIndent("  "))
	if err != nil {
		return fmt.Errorf("la decisión no se puede escribir en %s: %w", ruta, err)
	}

	if err := os.WriteFile(filepath.Clean(ruta), append(codificado, '\n'), 0o600); err != nil {
		return fmt.Errorf("la decisión no se puede escribir en %s: %w", ruta, err)
	}

	return nil
}

// leerElFicheroDeLaDecision lee el fichero de la decisión que escribió
// escribirElFicheroDeLaDecision. Es un error, que lo nombra, que no se pueda
// leer o que no sea JSON con su forma.
func leerElFicheroDeLaDecision(ruta string) (*origenDeLasEvals, error) {
	contenido, err := os.ReadFile(filepath.Clean(ruta))
	if err != nil {
		return nil, fmt.Errorf("la decisión de %s no se puede leer: %w", ruta, err)
	}

	var origen origenDeLasEvals
	if err := json.Unmarshal(contenido, &origen); err != nil {
		return nil, fmt.Errorf("la decisión de %s no se puede leer: %w", ruta, err)
	}

	return &origen, nil
}
