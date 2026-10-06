//go:build evals

package evals

import (
	"context"
	"errors"
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// directorioDeEvalsDeLasSkills es el de las evals de las skills en la raíz del
// repositorio, relativo al directorio de este paquete, que es donde go test
// ejecuta los tests (research.md V46): las de cada skill están en su
// subdirectorio.
const directorioDeEvalsDeLasSkills = "../../evals"

// Banderas con las que scripts/evals.sh invoca, tras -args, la ejecución del job
// (contracts/ejecucion-del-job.md §1 y §5 de H7.3). -skill, -repeticiones y
// -concurrencia son también del sondeo, que recibe las dos últimas tal como las
// escribe quien lo lanza, sin comprobar, y la concurrencia vacía para la del
// job: por eso son texto. Con una bandera entera, un valor que no lo es haría
// fallar go test antes del sondeo, sin el error que nombra su argumento de make
// (FR-067 de H7.3); el job las convierte, porque scripts/evals.sh ya las ha
// comprobado.
var (
	banderaSkill        = flag.String("skill", "", "skill evaluada")
	banderaQueDecide    = flag.String("modelo-que-decide", "", "modelo del job cuyas sesiones deciden el veredicto")
	banderaInformativos = flag.String("modelos-informativos", "",
		"modelos informativos del job, separados por comas; vacío si no hay ninguno")
	banderaRepeticiones = flag.String("repeticiones", "", "sesiones que se abren de cada eval con cada modelo")
	banderaUmbral       = flag.Int("umbral", 0, "sesiones de una serie que tienen que pasar para que la serie pase")
	banderaConcurrencia = flag.String("concurrencia", "",
		"sesiones que se abren a la vez como mucho; en el sondeo, vacía es la del job para la skill")
	banderaPruebaDeRed = flag.Bool("prueba-de-red", false,
		"añade la sesión de la primera eval con el texto de la prueba de red")
	banderaObjetivo = flag.Int("objetivo-de-duracion", 0,
		"segundos que el job admite para sus sesiones; 0, sin objetivo")
	banderaSkills   = flag.String("skills", "", "directorio de las skills instaladas, el que deja make install")
	banderaKitlegal = flag.String("kitlegal", "",
		"ruta absoluta de kitlegal: la orden del servidor del modo herramienta, y el directorio que sale del PATH "+
			"de las sesiones que no son del modo orden")
	banderaCommit    = flag.String("commit", "", "commit evaluado")
	banderaSinPython = flag.String("sin-python", "", "ruta de sin-python.txt")
	banderaSesiones  = flag.String("sesiones", "", "directorio en el que se crea el de cada sesión")
	banderaInforme   = flag.String("informe", "", "directorio en el que se escriben informe.md e informe.json")
)

// Banderas con las que scripts/evals.sh invoca además, tras -args, la ejecución
// del job de una skill con juez (contracts/job-de-evals.md §2 de H24): el modelo
// del juez y la versión de Claude Code de sus votos, los fijados en la
// definición del job, y el claude de esa versión, que no es el de las sesiones.
// Con una skill sin juez no las pasa ni se miran. Son también, con -skill,
// -concurrencia, -commit y -salida, las de la ejecución de la medida del juez,
// con las que la invoca scripts/evals-medir-juez.sh (contracts/job-de-evals.md
// §3 de H24).
var (
	banderaModeloDelJuez  = flag.String("modelo-del-juez", "", "id del modelo del juez de la skill")
	banderaVersionDelJuez = flag.String("version-del-juez", "", "versión de Claude Code de los votos del juez")
	banderaClaudeDelJuez  = flag.String("claude-del-juez", "",
		"ruta absoluta del claude con el que vota el juez: su directorio va delante en el PATH de sus votos")
)

// Banderas con las que scripts/evals-sondeo.sh invoca, tras -args, el sondeo,
// además de -skill, -repeticiones y -concurrencia (contracts/sondeo.md §2 y §3
// de H7.3; data-model §8 de H7.3): las evals y el modelo, tal como los escribe
// quien lo lanza, y el temporal que crea el guion.
var (
	banderaEvals    = flag.String("evals", "", "números de las evals del sondeo, de dos cifras y separados por comas")
	banderaModelo   = flag.String("modelo", "", "modelo con el que se abren las sesiones del sondeo")
	banderaTemporal = flag.String("temporal", "", "directorio del sondeo, que crea y borra su guion")
)

// Banderas con las que el paso decidir del trabajo tanda invoca, tras -args, la
// decisión de la tanda, además de -commit (contracts/tanda-del-job.md §1 y §3
// de H7.4): su ejecución del flujo y su GITHUB_OUTPUT. -salida es además la
// bandera con la que scripts/evals-medir-juez.sh dice a la ejecución de la
// medida del juez en qué fichero la escribe (contracts/medida-del-juez.md §7 de
// H24).
var (
	banderaEjecucion = flag.String("ejecucion", "", "databaseId de la ejecución del flujo evals que decide si mide")
	banderaSalida    = flag.String("salida", "",
		"en la decisión de la tanda, fichero al que se añade la línea medir=si o medir=no; en la medida del juez, "+
			"fichero en el que se escribe la medida")
)

// Banderas con las que el quickstart invoca, tras -args, la comprobación de la
// consulta repetida (contracts/comprobacion-del-quickstart.md §1).
var (
	banderaPrimera       = flag.String("primera", "", "directorio de la primera conversación de la consulta repetida")
	banderaSegunda       = flag.String("segunda", "", "directorio de la segunda conversación de la consulta repetida")
	banderaFechaSuperada = flag.String("fecha-superada", "",
		"fecha de vigencia de la redacción superada que la primera respuesta lleva en la línea de la forma")
	banderaFechaLeida = flag.String("fecha-leida", "",
		"fecha de vigencia de la redacción leída que la primera respuesta lleva en la línea de la forma")
)

// TestEjecucionDelJob es la ejecución del job de evals de una skill, entera y en
// una sola orden (contracts/ejecucion-del-job.md §1 y §5 de H7.3; research.md
// D19 de H7.3). Desde H24 reúne lo de sus banderas y lo ejecuta con
// ejecutarElJob (contracts/job-de-evals.md §2 de H24):
//
//  1. lee las evals de la skill de la raíz del repositorio y compone el plan de
//     sesiones con los modelos, las repeticiones y la prueba de red de sus
//     banderas, en el modo orden y en el modo herramienta
//     (contracts/evals-en-dos-modos.md §2.1 de H21). Falla si el directorio
//     tiene algún fichero mal formado: el guion ya lo ha comprobado antes, y
//     planificar sobre un conjunto incompleto abriría menos sesiones sin
//     decirlo;
//  2. con una skill que tiene juez, exige -modelo-del-juez, -version-del-juez y
//     -claude-del-juez y prepara el votante que abre cada voto con
//     scripts/evals-voto.sh (votanteDelJuez); con una skill sin juez no los
//     mira ni hay votante;
//  3. ejecutarElJob comprueba, con una skill que tiene juez, que su medida
//     versionada corresponde y se cumple con ese modelo y esa versión; si no,
//     escribe el informe del instrumento sin medir sin abrir ninguna sesión, de
//     evals o del juez (FR-043 de H24);
//  4. si la medida corresponde y se cumple, o la skill no tiene juez, reparte
//     las sesiones del plan con ejecutarSesiones, como mucho -concurrencia a la
//     vez, bajo strace, con scripts/evals-sesion.sh, el entorno del job debajo
//     del de cada sesión, las skills instaladas de -skills, el binario de
//     -kitlegal, que es el servidor del modo herramienta y lo que sale del PATH
//     de las sesiones que no son del modo orden, y el tope de 240 s con su
//     margen de 10 s. ejecutarElJob lo llama una vez por tanda, con
//     ejecutarPorTandas —la del modo orden, la del modo herramienta y la de las
//     evals sin binario ni servidor—, y mide cada una; si una acaba con sesiones
//     sin abrir por el límite de uso, las siguientes no se abren
//     (contracts/evals-en-dos-modos.md §2.2 de H21). SIGINT y SIGTERM cierran
//     las abiertas con la secuencia del tope, y el test falla con el error que
//     nombra cada una, sin escribir el informe (FR-037 de H7.3);
//  5. escribe informe.md e informe.json con EscribirInforme: las mismas evals y
//     los mismos modos, las sesiones, las que el repartidor no abrió tras el
//     mensaje del límite de uso y las duraciones que midió, en segundos
//     redondeados hacia arriba —la suma de las tres tandas y la de la tanda de
//     cada modo—, frente al objetivo de -objetivo-de-duracion (FR-044, FR-050 y
//     FR-051 de H7.3; FR-043 de H21); y, con una skill que tiene juez, el juicio
//     de sus respuestas, votadas como mucho -concurrencia a la vez, sin votar
//     ningún caso etiquetado (FR-044 de H24).
//
// Falla con un error o con el veredicto fallo, nombrando sus motivos: así el job
// sale en rojo por una serie que no pasa en un modo, por un umbral que decide y
// no se cumple en un modo, por la duración de una tanda, por sesiones sin medir
// (FR-043 de H7.3; FR-044 de H21), por una respuesta que el juez marca o deja
// sin juzgar o por el instrumento sin medir (FR-043 de H24). Solo lo ejecuta
// scripts/evals.sh, porque abre sesiones con modelo; lo que decide lo fijan
// TestPlan, TestPlanEnDosModos, los tests del repartidor, TestEjecutarPorTandas,
// TestInforme, TestInformeEnDosModos, TestUmbralesDelInforme,
// TestInformeConElJuez, TestEjecucionSinMedir y TestEjecutarElJob.
func TestEjecucionDelJob(t *testing.T) {
	t.Parallel()

	exigirBanderas(t, "skill", "modelo-que-decide", "repeticiones", "umbral", "concurrencia", "skills", "kitlegal",
		"commit", "sin-python", "sesiones", "informe")

	repeticiones := enteroDeLaBandera(t, "repeticiones")
	concurrencia := enteroDeLaBandera(t, "concurrencia")

	evals := filepath.Join(directorioDeEvalsDeLasSkills, *banderaSkill)

	conjunto, err := LeerConjunto(evals)
	require.NoError(t, err)
	require.Empty(t, conjunto.MalFormados, "el directorio de evals no tiene ficheros mal formados")

	// El repartidor ejecuta el guion en el directorio de trabajo de cada sesión:
	// su ruta relativa a este paquete no le serviría.
	guion, err := filepath.Abs(guionDeLaSesion)
	require.NoError(t, err)

	senales, dejarDeEscuchar := signal.NotifyContext(t.Context(), syscall.SIGINT, syscall.SIGTERM)
	defer dejarDeEscuchar()

	// El repartidor no recibe un contexto sino el Done() del que las señales
	// cancelan (research.md D9 de H7.3), que se toma aquí, fuera del cierre.
	interrupcion := senales.Done()

	var votar Votante
	if conjunto.Juez != nil {
		votar = votanteDelJuez(senales, t, conjunto.Juez)
	}

	informe, err := ejecutarElJob(EjecucionDelJob{
		Skill: *banderaSkill,
		Evals: evals,
		Plan: PlanDeEvals{
			Evals:               conjunto.Evals,
			ModeloQueDecide:     *banderaQueDecide,
			ModelosInformativos: separarLosModelos(*banderaInformativos),
			Repeticiones:        repeticiones,
			Modos:               []Modo{ModoOrden, ModoHerramienta},
			PruebaDeRed:         *banderaPruebaDeRed,
		},
		AbrirLaTanda: func(tanda []SesionPlanificada) (EjecucionDeSesiones, error) {
			return ejecutarSesiones(interrupcion, SesionesAEjecutar{
				Plan:          tanda,
				Concurrencia:  concurrencia,
				Evals:         evals,
				Sesiones:      *banderaSesiones,
				Skills:        *banderaSkills,
				Guion:         guion,
				Binario:       *banderaKitlegal,
				Entorno:       os.Environ(),
				Traza:         true,
				Tope:          topeDeUnaSesion,
				MargenDelTope: margenDelTopeDeUnaSesion,
			})
		},
		Sesiones:            *banderaSesiones,
		Destino:             *banderaInforme,
		Umbral:              *banderaUmbral,
		Commit:              *banderaCommit,
		SinPython:           *banderaSinPython,
		ObjetivoDeDuracion:  *banderaObjetivo,
		Votar:               votar,
		ModeloDelJuez:       *banderaModeloDelJuez,
		VersionDelJuez:      *banderaVersionDelJuez,
		ConcurrenciaDelJuez: concurrencia,
	})
	require.NoError(t, err)
	require.Equalf(t, VeredictoAprobado, informe.Veredicto, "motivos del veredicto:\n%s", strings.Join(informe.Motivos, "\n"))
}

// votanteDelJuez da el votante del juez de la skill en el job y en la ejecución
// de su medida, el que abre cada voto con scripts/evals-voto.sh
// (nuevoVotanteDelGuion; contracts/job-de-evals.md §2 y contracts/juez-y-voto.md
// §4 de H24), y deja para el final del test la retirada de su directorio. Exige
// las tres banderas del juez, que scripts/evals.sh pasa con una skill que lo
// tiene y scripts/evals-medir-juez.sh, siempre. El PATH de los votos es el de
// quien lo lanza con el directorio de -claude-del-juez delante (pathDelJuez),
// su credencial es la suya, y su contexto, el de las señales del punto de
// entrada, de modo que SIGINT y SIGTERM cortan los votos abiertos.
func votanteDelJuez(senales context.Context, t *testing.T, juez *Juez) Votante {
	t.Helper()

	exigirBanderas(t, "modelo-del-juez", "version-del-juez", "claude-del-juez")

	// Cada voto se ejecuta en su propio directorio: la ruta del guion relativa a
	// este paquete no le serviría.
	guion, err := filepath.Abs(guionDelVoto)
	require.NoError(t, err)

	entorno := os.Environ()

	path, err := pathDelJuez(*banderaClaudeDelJuez, valorEnElEntorno(entorno, variableDelPATH))
	require.NoError(t, err)

	votar, retirar, err := nuevoVotanteDelGuion(senales, ordenDelVoto{
		juez:        juez,
		modelo:      *banderaModeloDelJuez,
		guion:       guion,
		path:        path,
		suscripcion: valorEnElEntorno(entorno, variableDeLaSuscripcion),
	})
	require.NoError(t, err)

	t.Cleanup(func() { assert.NoError(t, retirar()) })

	return votar
}

// enteroDeLaBandera es el entero del valor de la bandera de la ejecución del
// job, que scripts/evals.sh ha comprobado antes de invocarla.
func enteroDeLaBandera(t *testing.T, nombre string) int {
	t.Helper()

	entero, err := strconv.Atoi(flag.Lookup(nombre).Value.String())
	require.NoErrorf(t, err, "la bandera -%s es un entero", nombre)

	return entero
}

// TestMedidaDelJuez es la ejecución de la medida del juez de una skill, entera
// y en una sola orden (contracts/medida-del-juez.md §7 y
// contracts/job-de-evals.md §3 de H24; FR-050 a FR-054 de H24):
//
//  1. exige sus banderas —-skill, -modelo-del-juez, -version-del-juez,
//     -claude-del-juez, -concurrencia, -commit y -salida— y lee el juez de la
//     skill de sus evals de la raíz del repositorio. Falla si el directorio
//     tiene algún fichero mal formado: con la carpeta del juez mal formada no
//     hay juez que medir, y la reconstrucción de los casos necesita las evals;
//  2. prepara el votante que abre cada voto con scripts/evals-voto.sh
//     (votanteDelJuez), con el directorio de -claude-del-juez delante en su
//     PATH y el contexto de SIGINT y SIGTERM, que cortan los votos abiertos:
//     sus casos quedan sin juzgar;
//  3. llama a medirAlJuez con el reconstructor de la skill, el modelo y la
//     versión de sus banderas, que van a la medida tal cual, -concurrencia
//     casos a la vez, el commit y el instante en el que se lanza. No comprueba
//     la medida versionada ni abre ninguna sesión de evals (FR-043, FR-051);
//  4. escribe en el fichero de -salida el texto de la medida, si medirAlJuez da
//     alguno: es el que su guion imprime entre sus dos marcas (FR-052). Con
//     algún caso sin juzgar no hay texto y no escribe nada (FR-053). No escribe
//     en ningún otro sitio, tampoco en el repositorio (FR-054);
//  5. falla con el error de medirAlJuez, que nombra cada caso: el defecto que
//     no queda marcado y el correcto que queda marcado, con la medida ya
//     escrita, o el que queda sin juzgar, con su motivo.
//
// Solo lo ejecuta scripts/evals-medir-juez.sh, porque abre sesiones con modelo;
// lo que decide lo fija TestEjecucionDeLaMedida, y la orden que lo ejecuta y lo
// que su guion hace con el fichero de -salida, TestGuionDeLaMedida.
func TestMedidaDelJuez(t *testing.T) {
	t.Parallel()

	exigirBanderas(t, "skill", "modelo-del-juez", "version-del-juez", "claude-del-juez", "concurrencia", "commit",
		"salida")

	concurrencia := enteroDeLaBandera(t, "concurrencia")

	evals := filepath.Join(directorioDeEvalsDeLasSkills, *banderaSkill)

	conjunto, err := LeerConjunto(evals)
	require.NoError(t, err)
	require.Empty(t, conjunto.MalFormados, "el directorio de evals no tiene ficheros mal formados")

	senales, dejarDeEscuchar := signal.NotifyContext(t.Context(), syscall.SIGINT, syscall.SIGTERM)
	defer dejarDeEscuchar()

	// Sin juez no hay con qué preparar el votante: medirAlJuez dice que la
	// skill no lo tiene.
	var votar Votante
	if conjunto.Juez != nil {
		votar = votanteDelJuez(senales, t, conjunto.Juez)
	}

	texto, err := medirAlJuez(nuevoReconstructor(evals), MedicionDelJuez{
		Skill:          *banderaSkill,
		Juez:           conjunto.Juez,
		Votar:          votar,
		ModeloDelJuez:  *banderaModeloDelJuez,
		VersionDelJuez: *banderaVersionDelJuez,
		Concurrencia:   concurrencia,
		Commit:         *banderaCommit,
		Fecha:          time.Now(),
	})

	if texto != "" {
		require.NoError(t, os.WriteFile(*banderaSalida, []byte(texto), 0o600))
	}

	require.NoError(t, err)
}

// TestSondeo es el sondeo local de unas evals de una skill, entero y en una sola
// orden (contracts/sondeo.md §2, §3 y §6 de H7.3; research.md D16 de H7.3):
// con sondear,
//
//  1. comprueba, antes de construir nada, los argumentos de -skill, -evals,
//     -modelo, -repeticiones y -concurrencia, con los errores que nombran su
//     argumento de make, y CLAUDE_CODE_OAUTH_TOKEN (FR-063 y FR-067 de H7.3);
//  2. construye el binario del árbol de trabajo e instala sus skills en el
//     temporal de -temporal, con prepararElArbol (FR-062 de H7.3);
//  3. abre las sesiones de las evals pedidas con el repartidor, sin traza, con
//     scripts/evals-sesion.sh y el entorno de quien lo lanza que ven las
//     sesiones del sondeo. SIGINT y SIGTERM cierran las abiertas con la
//     secuencia del tope, y el test falla con el error que nombra cada una;
//  4. las juzga con el código del job, sin lo que el job lee de la traza;
//  5. con una skill que tiene juez, juzga con él sus respuestas, con el votante
//     que abre cada voto con scripts/evals-voto.sh (votanteDelSondeo): el modelo
//     del juez es el de la definición del job, y el claude de sus votos, el del
//     PATH de quien lo lanza, el mismo que abre las sesiones
//     (contracts/informe-del-job.md §8 de H24; FR-075 de H24). SIGINT y SIGTERM
//     cortan los votos abiertos. Su salida dice si la medida versionada del juez
//     corresponde a ese Claude Code, sin comprobarla para decidir nada (FR-076
//     de H24).
//
// Escribe la salida en salida.txt del temporal, que su guion imprime, y ningún
// informe ni veredicto: falla solo con un error, sean cuales sean las tasas y
// lo que el juez marque (FR-066 de H7.3). Con un error de uso —un argumento que no vale o la
// credencial que falta, un errorDeUso—, no falla: escribe su mensaje, con un
// salto de línea final, en uso.txt del temporal, que su guion imprime solo en
// la salida de error, y no escribe salida.txt (contracts/sondeo.md §2 de H7.4;
// FR-080 de H7.4). Con cualquier otro error falla, y su guion imprime el
// registro de go test (FR-081 de H7.4). Sin -temporal falla antes de nada,
// porque el sondeo escribiría en el directorio de este paquete. Solo lo ejecuta
// scripts/evals-sondeo.sh, porque abre sesiones con modelo; lo que decide lo
// fijan TestComprobarElSondeo, TestSondear, TestJuicioDelSondeo y
// TestSalidaDelSondeo, y la orden que lo ejecuta, TestGuionDelSondeo.
func TestSondeo(t *testing.T) {
	t.Parallel()

	exigirBanderas(t, "temporal")

	// El repartidor ejecuta el guion en el directorio de trabajo de cada sesión:
	// su ruta relativa a este paquete no le serviría.
	guion, err := filepath.Abs(guionDeLaSesion)
	require.NoError(t, err)

	// Cada voto se ejecuta en su propio directorio, como cada sesión.
	delVoto, err := filepath.Abs(guionDelVoto)
	require.NoError(t, err)

	senales, dejarDeEscuchar := signal.NotifyContext(t.Context(), syscall.SIGINT, syscall.SIGTERM)
	defer dejarDeEscuchar()

	salida, err := sondear(senales.Done(), SondeoAEjecutar{
		Argumentos: ArgumentosDelSondeo{
			Skill:        *banderaSkill,
			Evals:        *banderaEvals,
			Modelo:       *banderaModelo,
			Repeticiones: *banderaRepeticiones,
			Concurrencia: *banderaConcurrencia,
		},
		Entorno:                  os.Environ(),
		EvalsDeLasSkills:         directorioDeEvalsDeLasSkills,
		RutaDeLaDefinicionDelJob: rutaDeLaDefinicionDelJob,
		Temporal:                 *banderaTemporal,
		Guion:                    guion,
		PrepararElArbol:          prepararElArbol,
		NuevoVotante: func(juez *Juez, modelo string) (Votante, error) {
			return votanteDelSondeo(senales, t, delVoto, juez, modelo)
		},
	})

	var uso *errorDeUso
	if errors.As(err, &uso) {
		require.NoError(t, os.WriteFile(filepath.Join(*banderaTemporal, ficheroDelUsoDelSondeo), []byte(uso.Error()+"\n"),
			0o600))

		return
	}

	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(*banderaTemporal, ficheroDeLaSalidaDelSondeo), []byte(salida), 0o600))
}

// votanteDelSondeo da el votante del juez de la skill en el sondeo, el que abre
// cada voto con el guion del voto de esa ruta, que es absoluta
// (nuevoVotanteDelGuion; contracts/informe-del-job.md §8 y
// contracts/juez-y-voto.md §4 de H24), y deja para el final del test la
// retirada de su directorio. El juez y el id de su modelo los da sondear: el de
// la carpeta de evals de la skill y el de la definición del job. El PATH de los
// votos es el de quien lanza el sondeo, sin nada delante —su claude es el del
// equipo, el que abre las sesiones—, su credencial es la suya, y su contexto, el
// de las señales del punto de entrada, de modo que SIGINT y SIGTERM cortan los
// votos abiertos.
func votanteDelSondeo(senales context.Context, t *testing.T, guion string, juez *Juez, modelo string) (Votante, error) {
	t.Helper()

	entorno := os.Environ()

	votar, retirar, err := nuevoVotanteDelGuion(senales, ordenDelVoto{
		juez:        juez,
		modelo:      modelo,
		guion:       guion,
		path:        valorEnElEntorno(entorno, variableDelPATH),
		suscripcion: valorEnElEntorno(entorno, variableDeLaSuscripcion),
	})
	if err != nil {
		return nil, err
	}

	t.Cleanup(func() { assert.NoError(t, retirar()) })

	return votar, nil
}

// TestTandaDelCommit decide si la ejecución del flujo evals de -ejecucion mide
// el commit de -commit (contracts/tanda-del-job.md §2 y §3 de H7.4; research.md
// D15 y D16 de H7.4; FR-070 y FR-071 de H7.4): consulta con gh las ejecuciones
// del flujo sobre el commit y los trabajos de cada anterior sin terminar, decide
// con esperarLaDecision y, mientras alguna no ha decidido, espera de verdad
// entre consulta y consulta; registra qué ejecuciones anteriores miró en cada
// consulta y por qué mide o no; y añade a -salida la línea medir=si o medir=no.
// Falla solo con un error: el de una orden de gh, que la nombra con lo que
// escribió, o el de -salida. No abre sesiones ni escribe fuera de -salida. Solo
// lo ejecuta el paso decidir del trabajo tanda, porque consulta la API de
// GitHub; lo que decide lo fijan las subpruebas segundo-disparo y
// estado-de-la-tanda de TestDefinicionDelJob, y las órdenes, lo que registra y
// lo que escribe, TestOrdenesDeLaTanda, TestRegistroDeLaTanda y
// TestSalidaDeLaTanda.
func TestTandaDelCommit(t *testing.T) {
	t.Parallel()

	exigirBanderas(t, "commit", "ejecucion", "salida")

	propia, err := strconv.ParseInt(*banderaEjecucion, 10, 64)
	require.NoErrorf(t, err, "la bandera -ejecucion es el databaseId de una ejecuci\xc3\xb3n")

	gh := ghDeLaTanda{entorno: os.Environ(), commit: *banderaCommit}

	consultas := 0

	var miradas []ejecucionDelCommit

	consultar := func(ctx context.Context) ([]ejecucionDelCommit, error) {
		ejecuciones, err := consultarLasEjecuciones(ctx, propia, gh.listar, gh.verLosTrabajos)
		if err != nil {
			return nil, err
		}

		consultas++
		miradas = ejecuciones
		t.Log(registroDeLaConsulta(consultas, propia, ejecuciones))

		return ejecuciones, nil
	}

	decision, err := esperarLaDecision(t.Context(), propia, consultar, time.Now, dormir)
	require.NoError(t, err)

	t.Log(motivoDeLaDecision(propia, miradas, decision))
	require.NoError(t, escribirLaDecision(*banderaSalida, decision.mide))
}

// TestComprobarConsultaRepetida comprueba las dos respuestas de la consulta
// repetida del quickstart §6 con comprobarConsultaRepetida: las conversaciones
// de -primera y -segunda, leídas con LeerSesion como las lee el job, y las fechas
// de -fecha-superada y -fecha-leida. Falla con las conversaciones que no se
// pueden leer o con una línea por condición que falla, una por renglón; si no
// falla ninguna, lo registra (contracts/comprobacion-del-quickstart.md; FR-061).
// Solo lo ejecuta la persona en el quickstart, porque necesita las dos
// conversaciones con modelo (FR-062); lo que decide lo fija
// TestCondicionesDeLaConsultaRepetida. Desde H24 no lee la lista de expresiones
// prohibidas de ninguna skill, que ya no juzga ninguna respuesta (research D14
// de H24): -skill, que el quickstart pasaba para leerla, ya no hace falta.
func TestComprobarConsultaRepetida(t *testing.T) {
	t.Parallel()

	exigirBanderas(t, "primera", "segunda", "fecha-superada", "fecha-leida")

	primera := leerConversacion(t, "primera", *banderaPrimera)
	segunda := leerConversacion(t, "segunda", *banderaSegunda)

	if t.Failed() {
		t.FailNow()
	}

	if lineas := comprobarConsultaRepetida(primera, segunda, *banderaFechaSuperada,
		*banderaFechaLeida); len(lineas) > 0 {
		t.Fatal(strings.Join(lineas, "\n"))
	}

	t.Logf("se cumplen las dos condiciones: la forma con %s y %s en la primera respuesta, y sin ella en la segunda",
		*banderaFechaSuperada, *banderaFechaLeida)
}

// leerConversacion lee con LeerSesion la conversación del directorio dado. Si
// no se puede leer, lo anota nombrándola por su ordinal y con el error, y deja
// seguir para que se lea también la otra.
func leerConversacion(t *testing.T, ordinal, dir string) Sesion {
	t.Helper()

	sesion, err := LeerSesion(dir)
	if err != nil {
		t.Errorf("la %s conversación no se ha podido leer: %v", ordinal, err)
	}

	return sesion
}

// exigirBanderas falla, nombrando cada una, si alguna de las banderas no tiene
// valor: con una ruta vacía, PrepararSesion o EscribirInforme leerían y
// escribirían en el directorio de este paquete en lugar de en el de la
// ejecución. Las mira todas antes de fallar, para que quien lanza la orden sin
// ellas sepa de una vez las que faltan.
func exigirBanderas(t *testing.T, nombres ...string) {
	t.Helper()

	faltan := false

	for _, nombre := range nombres {
		if !assert.NotEmptyf(t, flag.Lookup(nombre).Value.String(), "falta la bandera -%s", nombre) {
			faltan = true
		}
	}

	if faltan {
		t.FailNow()
	}
}
