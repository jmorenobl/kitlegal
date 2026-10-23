package evals

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// lecturasDeLaPlataforma son las lecturas de la plataforma que necesitan la
// elección de una propuesta, la búsqueda de medidas de main y la comprobación de
// la suite completa de la release (FR-007, FR-010, FR-025; research D8, D15):
// todas de lectura. La implementa ghDeLaPlataforma, con gh; los tests de la
// lógica, en memoria.
type lecturasDeLaPlataforma interface {
	// ficherosDeLaPropuesta son los ficheros de la propuesta del número dado,
	// cada uno con su ruta y, si se renombró, la de antes (FR-007; research
	// S3).
	ficherosDeLaPropuesta(ctx context.Context, propuesta int) ([]cambioDeLaPropuesta, error)

	// propuestasFusionadas son las propuestas cerradas sobre main que se
	// fusionaron, de las 100 actualizadas más recientemente (FR-010; research
	// S2).
	propuestasFusionadas(ctx context.Context) ([]propuestaFusionada, error)

	// ejecucionesDelFlujo son las 200 ejecuciones más recientes del flujo evals
	// (FR-010; research S1).
	ejecucionesDelFlujo(ctx context.Context) ([]ejecucionDelFlujo, error)

	// ejecucionesDelCommit son las ejecuciones del flujo evals sobre el commit
	// dado, como mucho 100 (FR-025).
	ejecucionesDelCommit(ctx context.Context, commit string) ([]ejecucionDelFlujo, error)

	// trabajos son los trabajos de la ejecución dada, con sus pasos.
	trabajos(ctx context.Context, ejecucion int64) ([]trabajoDeLaEjecucion, error)

	// informeDelTrabajo es el informe.json que el trabajo dado de la ejecución
	// dada imprimió en su registro (research D8, V7).
	informeDelTrabajo(ctx context.Context, ejecucion, trabajo int64) ([]byte, error)
}

// propuestaFusionada es una propuesta cerrada sobre main que se fusionó: su
// número, el commit que era su cabeza al fusionarla y cuándo se fusionó (FR-010;
// research S2).
type propuestaFusionada struct {
	numero int
	cabeza string
	fusion time.Time
}

// ejecucionDelFlujo es una ejecución del flujo evals de gh run list --json
// databaseId,headSha,status,conclusion,createdAt (research S1).
type ejecucionDelFlujo struct {
	ID         int64     `json:"databaseId"`
	Commit     string    `json:"headSha"`
	Status     string    `json:"status"`
	Conclusion string    `json:"conclusion"`
	Creada     time.Time `json:"createdAt"`
}

// Las órdenes de gh con las que se lee la plataforma (contracts/job-de-evals.md
// §5), además de ordenDeLosTrabajos, la de la tanda: constantes, con cada valor
// en una variable del entorno de la orden y entre comillas, de modo que ningún
// dato entra en su texto (gosec G204), como las de la tanda. exec sustituye la
// shell por gh, que termina con su código.
const (
	// ordenDeLosFicherosDeLaPropuesta da los ficheros de la propuesta de
	// variableDeLaPropuesta, uno por línea, con su ruta y la de antes, o nada,
	// separadas por un tabulador.
	ordenDeLosFicherosDeLaPropuesta = `exec gh api --paginate "repos/{owner}/{repo}/pulls/$PROPUESTA/files" ` +
		`--jq '.[] | [.filename, (.previous_filename // "")] | @tsv'`

	// ordenDeLasPropuestasFusionadas da las propuestas cerradas sobre main que
	// se fusionaron, una por línea, con su número, su cabeza y la fecha de su
	// fusión, separados por tabuladores.
	ordenDeLasPropuestasFusionadas = `exec gh api ` +
		`"repos/{owner}/{repo}/pulls?state=closed&base=main&sort=updated&direction=desc&per_page=100" ` +
		`--jq '.[] | select(.merged_at != null) | [.number, .head.sha, .merged_at] | @tsv'`

	// ordenDeLasEjecucionesDelFlujo lista las ejecuciones del flujo.
	ordenDeLasEjecucionesDelFlujo = `exec gh run list --workflow evals.yml --limit 200 ` +
		`--json databaseId,headSha,status,conclusion,createdAt`

	// ordenDeLasEjecucionesDeUnCommit lista las ejecuciones del flujo sobre el
	// commit de variableDelCommitEvaluado, con más campos que la de la tanda.
	ordenDeLasEjecucionesDeUnCommit = `exec gh run list --workflow evals.yml --commit "$COMMIT_EVALUADO" --limit 100 ` +
		`--json databaseId,headSha,status,conclusion,createdAt`

	// ordenDelRegistroDeUnTrabajo da el registro del trabajo de
	// variableDelTrabajo en la ejecución de variableDeLaEjecucion.
	ordenDelRegistroDeUnTrabajo = `exec gh run view "$EJECUCION" --job "$TRABAJO" --log`
)

// Las variables del entorno con las que las órdenes de la plataforma reciben
// su valor, además de las de la tanda.
const (
	variableDeLaPropuesta = "PROPUESTA"
	variableDeLaEjecucion = "EJECUCION"
	variableDelTrabajo    = "TRABAJO"
)

// Las marcas entre las que el job imprime el informe.json de un trabajo en su
// registro (scripts/evals.sh), las que busca informe_del_trabajo de
// scripts/workflow/comun.sh.
const (
	marcaDeInicioDelInforme = "--- inicio de informe.json ---"
	marcaDeFinDelInforme    = "--- fin de informe.json ---"
)

// ordenDeBytesDelRegistro es la marca de orden de bytes con la que GitHub
// empieza cada trozo del registro de un trabajo, y que gh run view --log deja
// delante de la hora de esa línea.
const ordenDeBytesDelRegistro = "\xef\xbb\xbf"

// horaDelRegistro es la hora que gh run view --log pone delante del texto de
// cada línea, con el espacio que la sigue.
var horaDelRegistro = regexp.MustCompile(`^[0-9-]+T[0-9:.]+Z `)

// sinEscapesDeTSV quita los escapes de @tsv de jq de un campo: la barra
// invertida, el tabulador, el salto de línea y el retorno de carro.
var sinEscapesDeTSV = strings.NewReplacer(`\\`, `\`, `\t`, "\t", `\n`, "\n", `\r`, "\r")

// ghDeLaPlataforma son las lecturas de la plataforma con gh
// (contracts/job-de-evals.md §5), con el entorno dado: en el job, el del paso,
// con GH_TOKEN y GH_REPO.
type ghDeLaPlataforma struct {
	entorno []string
}

// ghDeLaPlataforma implementa las lecturas de la plataforma.
var _ lecturasDeLaPlataforma = ghDeLaPlataforma{}

func (g ghDeLaPlataforma) ficherosDeLaPropuesta(ctx context.Context, propuesta int) ([]cambioDeLaPropuesta, error) {
	salida, err := ejecutarGh(exec.CommandContext(ctx, "sh", "-c", ordenDeLosFicherosDeLaPropuesta), g.entorno,
		variableDeLaPropuesta+"="+strconv.Itoa(propuesta))
	if err != nil {
		return nil, err
	}

	filas, err := filasDeTSV(salida, 2)
	if err != nil {
		return nil, fmt.Errorf("los ficheros de la propuesta %d no se pueden leer: %w", propuesta, err)
	}

	cambios := make([]cambioDeLaPropuesta, 0, len(filas))
	for _, fila := range filas {
		cambios = append(cambios, cambioDeLaPropuesta{ruta: fila[0], anterior: fila[1]})
	}

	return cambios, nil
}

func (g ghDeLaPlataforma) propuestasFusionadas(ctx context.Context) ([]propuestaFusionada, error) {
	salida, err := ejecutarGh(exec.CommandContext(ctx, "sh", "-c", ordenDeLasPropuestasFusionadas), g.entorno)
	if err != nil {
		return nil, err
	}

	filas, err := filasDeTSV(salida, 3)
	if err != nil {
		return nil, fmt.Errorf("las propuestas fusionadas no se pueden leer: %w", err)
	}

	propuestas := make([]propuestaFusionada, 0, len(filas))

	for _, fila := range filas {
		numero, err := strconv.Atoi(fila[0])
		if err != nil {
			return nil, fmt.Errorf("las propuestas fusionadas no se pueden leer: el número %q: %w", fila[0], err)
		}

		fusion, err := time.Parse(time.RFC3339, fila[2])
		if err != nil {
			return nil, fmt.Errorf("las propuestas fusionadas no se pueden leer: la fusión de la %d: %w", numero, err)
		}

		propuestas = append(propuestas, propuestaFusionada{numero: numero, cabeza: fila[1], fusion: fusion})
	}

	return propuestas, nil
}

func (g ghDeLaPlataforma) ejecucionesDelFlujo(ctx context.Context) ([]ejecucionDelFlujo, error) {
	salida, err := ejecutarGh(exec.CommandContext(ctx, "sh", "-c", ordenDeLasEjecucionesDelFlujo), g.entorno)
	if err != nil {
		return nil, err
	}

	return leerLasEjecucionesDelFlujo(salida)
}

func (g ghDeLaPlataforma) ejecucionesDelCommit(ctx context.Context, commit string) ([]ejecucionDelFlujo, error) {
	salida, err := ejecutarGh(exec.CommandContext(ctx, "sh", "-c", ordenDeLasEjecucionesDeUnCommit), g.entorno,
		variableDelCommitEvaluado+"="+commit)
	if err != nil {
		return nil, err
	}

	return leerLasEjecucionesDelFlujo(salida)
}

func (g ghDeLaPlataforma) trabajos(ctx context.Context, ejecucion int64) ([]trabajoDeLaEjecucion, error) {
	salida, err := ejecutarGh(exec.CommandContext(ctx, "sh", "-c", ordenDeLosTrabajos), g.entorno,
		variableDeLaEjecucionAnterior+"="+strconv.FormatInt(ejecucion, 10))
	if err != nil {
		return nil, err
	}

	var trabajos trabajosDeLaEjecucion
	if err := json.Unmarshal(salida, &trabajos); err != nil {
		return nil, fmt.Errorf("los trabajos de la ejecución %d no se pueden leer: %w", ejecucion, err)
	}

	return trabajos.Jobs, nil
}

func (g ghDeLaPlataforma) informeDelTrabajo(ctx context.Context, ejecucion, trabajo int64) ([]byte, error) {
	registro, err := ejecutarGh(exec.CommandContext(ctx, "sh", "-c", ordenDelRegistroDeUnTrabajo), g.entorno,
		variableDeLaEjecucion+"="+strconv.FormatInt(ejecucion, 10), variableDelTrabajo+"="+strconv.FormatInt(trabajo, 10))
	if err != nil {
		return nil, err
	}

	informe, err := informeDelRegistro(registro)
	if err != nil {
		return nil, fmt.Errorf("el informe del trabajo %d de la ejecución %d: %w", trabajo, ejecucion, err)
	}

	return informe, nil
}

// leerLasEjecucionesDelFlujo lee la lista de gh run list --json
// databaseId,headSha,status,conclusion,createdAt.
func leerLasEjecucionesDelFlujo(lista []byte) ([]ejecucionDelFlujo, error) {
	var ejecuciones []ejecucionDelFlujo
	if err := json.Unmarshal(lista, &ejecuciones); err != nil {
		return nil, fmt.Errorf("la lista de ejecuciones del flujo no se puede leer: %w", err)
	}

	return ejecuciones, nil
}

// filasDeTSV son las filas de la salida de un --jq con @tsv, una por línea,
// cada una con el número de campos dado, separados por tabuladores y sin los
// escapes de @tsv. Una línea con otro número de campos es un error que la
// nombra.
func filasDeTSV(salida []byte, campos int) ([][]string, error) {
	var filas [][]string

	for linea := range strings.Lines(string(salida)) {
		fila := strings.Split(strings.TrimSuffix(linea, "\n"), "\t")
		if len(fila) != campos {
			return nil, fmt.Errorf("la línea %q tiene %d campos y no %d", linea, len(fila), campos)
		}

		for indice, campo := range fila {
			fila[indice] = sinEscapesDeTSV.Replace(campo)
		}

		filas = append(filas, fila)
	}

	return filas, nil
}

// informeDelRegistro es el informe.json que un trabajo imprimió en su
// registro, el de gh run view --log, como lo saca informe_del_trabajo de
// scripts/workflow/comun.sh: lo que hay entre las dos marcas del informe, línea
// a línea, con el texto de cada línea (textoDeLaLinea). Un registro sin la
// marca de inicio, sin la de fin tras ella o con la de inicio más de una vez es
// un error que lo dice; una marca de fin sin la de inicio delante no cuenta.
func informeDelRegistro(registro []byte) ([]byte, error) {
	var informe bytes.Buffer

	empezado, dentro := false, false

	for linea := range strings.Lines(string(registro)) {
		texto := textoDeLaLinea(strings.TrimSuffix(linea, "\n"))

		switch {
		case texto == marcaDeInicioDelInforme:
			if empezado {
				return nil, fmt.Errorf("el registro trae más de un informe: la marca %q aparece más de una vez",
					marcaDeInicioDelInforme)
			}

			empezado, dentro = true, true
		case texto == marcaDeFinDelInforme:
			dentro = false
		case dentro:
			informe.WriteString(texto + "\n")
		}
	}

	switch {
	case !empezado:
		return nil, fmt.Errorf("el registro no trae la marca %q", marcaDeInicioDelInforme)
	case dentro:
		return nil, fmt.Errorf("el registro no trae la marca %q tras la de inicio", marcaDeFinDelInforme)
	}

	return informe.Bytes(), nil
}

// textoDeLaLinea es el texto de una línea del registro de un trabajo sin lo
// que gh run view --log pone delante, como informe_del_trabajo de
// scripts/workflow/comun.sh: el trabajo y el paso, cada uno con su tabulador
// detrás —como cut -f3-, una línea sin tabulador queda entera y una con uno
// solo, vacía—, la marca de orden de bytes que empieza un trozo y la hora.
func textoDeLaLinea(linea string) string {
	texto := linea
	if _, resto, ok := strings.Cut(linea, "\t"); ok {
		_, texto, _ = strings.Cut(resto, "\t")
	}

	return horaDelRegistro.ReplaceAllLiteralString(strings.TrimPrefix(texto, ordenDeBytesDelRegistro), "")
}

// errorDeUnaOrden es el de una orden de gh que no terminó con 0 (data-model §5):
// la orden, con su texto constante; las variables, nombre=valor, con las que
// recibió sus valores; cómo terminó; y lo que escribió en sus dos salidas. La
// orden y su salida de error son lo que se escribe del error de la plataforma.
type errorDeUnaOrden struct {
	orden         string
	variables     []string
	causa         error
	salida        string
	salidaDeError string
}

// Error nombra la orden, con sus variables, y dice cómo terminó y lo que
// escribió en sus dos salidas.
func (e *errorDeUnaOrden) Error() string {
	orden := e.orden
	if len(e.variables) > 0 {
		orden += ", con " + strings.Join(e.variables, ", ")
	}

	return fmt.Sprintf("%s: %v\n%s%s", orden, e.causa, e.salida, e.salidaDeError)
}

// Unwrap es cómo terminó la orden.
func (e *errorDeUnaOrden) Unwrap() error {
	return e.causa
}

// ejecutarGh ejecuta la orden de gh dada —sh -c con su texto constante, el
// último de sus argumentos— con el entorno dado y, encima, las variables dadas,
// nombre=valor, que sustituyen a las que el entorno tenga con el mismo nombre,
// y devuelve lo que escribió en su salida estándar. Si no termina con 0,
// devuelve un *errorDeUnaOrden. La usan la tanda y las lecturas de la
// plataforma.
func ejecutarGh(orden *exec.Cmd, entorno []string, variables ...string) ([]byte, error) {
	var salida, errores bytes.Buffer

	orden.Env = sobreLaBase(entorno, variables)
	orden.Stdout = &salida
	orden.Stderr = &errores

	if err := orden.Run(); err != nil {
		return nil, &errorDeUnaOrden{
			orden:         orden.Args[len(orden.Args)-1],
			variables:     variables,
			causa:         err,
			salida:        salida.String(),
			salidaDeError: errores.String(),
		}
	}

	return salida.Bytes(), nil
}
