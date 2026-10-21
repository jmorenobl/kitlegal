package evals

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/plazos"
	"github.com/jmorenobl/kitlegal/internal/skills"
)

// TestConsultasNecesarias fija ConsultasNecesarias (data-model §7.1; contrato
// evals-y-grabaciones §4): cada comando esperado en su forma de invocación, el
// índice y los metadatos de cada norma de las citas y de los comandos con
// identificador, y el bloque de cada cita; ninguna invocación repetida, en el
// orden en que aparece por primera vez y con cada eval y punto de los que sale,
// sin repetir tampoco ninguno.
//
// Desde H6, un comando de territorio no genera ninguna consulta que grabar —el
// applet no pide nada por red ni usa la caché—, ni solo ni junto a los comandos y
// las citas del BOE de la misma eval, que siguen generando las suyas (data-model
// §6.3 de H6; research D21; FR-043).
//
// Desde H7, tampoco genera ninguna el comando de comprobación, graph check, que
// lee el grafo de la sesión y no la caché preparada: la eval de la consulta
// repetida solo necesita las consultas de su bloque y de su norma, las mismas que
// la eval 01 (contrato evals-y-skill §1 de H7).
//
// Desde H23, tampoco generan ninguna los dos comandos de cita, con su ROJ, con
// texto o sin nada más: el applet no consulta ninguna fuente, y las evals de
// jurisprudencia no necesitan ninguna respuesta grabada
// (contracts/evals-jurisprudencia.md §2 de H23; FR-054).
//
// Desde H9, tampoco genera ninguna el comando calcular, plazos calcular: el
// applet no consulta nada, y la eval de plazo solo necesita las consultas del
// bloque que lee y cita; ni la eval de cobertura, que no espera comandos ni citas
// (contracts/evals-legal-core.md §1 de H9).
//
// Desde H20, cada lectura de un bloque, la de un comando esperado o la de una
// cita, va seguida de la misma con --fecha 9999-12-31, con los mismos orígenes:
// la caché preparada sirve así la lectura del bloque en cualquier fecha. El
// comando con fecha de una eval da las mismas dos lecturas que sin ella, nunca
// la de su fecha (contracts/evals-redaccion-fechada.md §4 de H20; research D18;
// FR-053).
func TestConsultasNecesarias(t *testing.T) {
	t.Parallel()

	const (
		lpac = "BOE-A-2015-10565"
		lcsp = "BOE-A-2017-12902"

		// otra es inventada —ninguna norma es del año 0000—: ConsultasNecesarias
		// solo copia los identificadores.
		otra = "BOE-A-0000-2"

		deTresFormas            = "01-lpac-tres-formas.yaml"
		deNoActivacion          = "02-no-activa-programacion.yaml"
		deArticulo22            = "03-lpac-articulo-22.yaml"
		deSoloTerritorio        = "04-territorio-municipio-cubierto.yaml"
		deTerritorioYArticulo21 = "05-territorio-y-articulo-21.yaml"
		deConsultaRepetida      = "06-lpac-articulo-21-redaccion-cambiada.yaml"
		deUnPlazo               = "07-plazo-reposicion-leganes.yaml"
		deUnaCobertura          = "08-cobertura-liquidacion-tributaria.yaml"
		deContratoMenorEn2019   = "09-lcsp-contrato-menor-en-2019.yaml"
		deArticulo21En2016      = "10-lpac-articulo-21-en-2016.yaml"
	)

	// fechada son los argumentos de la lectura del bloque con --fecha
	// 9999-12-31, escritos a mano.
	fechada := func(norma, bloque string) []string {
		return []string{norma, bloque, "--fecha", "9999-12-31"}
	}

	// Una eval con las tres formas de comando, dos normas —la LPAC, de un comando
	// y de sus citas; la otra, solo de una cita— y la cita del art. 21 dos veces,
	// que es además el bloque de su comando.
	tresFormas := Eval{
		Fichero:  deTresFormas,
		Pregunta: "¿Qué dice el art. 21 de la Ley 39/2015 y dónde se regula el procedimiento común?",
		Activa:   true,
		Comandos: []ComandoEsperado{
			{Applet: "boe", Norma: lpac, Bloque: "a21"},
			{Applet: "boe", Verbo: "indice", Norma: lpac},
			{Applet: "boe", Verbo: "buscar", Terminos: []string{"procedimiento", "común"}},
		},
		Citas: []CitaEsperada{
			{Norma: lpac, Bloque: "a21"},
			{Norma: otra, Bloque: "a5"},
			{Norma: lpac, Bloque: "a21"},
		},
	}
	comandoDeTresFormas := Origen{Eval: deTresFormas, Punto: PuntoComandoEsperado}
	normaDeTresFormas := Origen{Eval: deTresFormas, Punto: PuntoNormaDeLaEval}
	citaDeTresFormas := Origen{Eval: deTresFormas, Punto: PuntoCitaEsperada}

	// Dos evals de la LPAC, con una de no activación entre ellas, que comparten el
	// índice y los metadatos de la norma; la 03 espera además los metadatos como
	// comando.
	articulo21 := Eval{
		Fichero:  nombreDeEval,
		Pregunta: "¿qué dice el art. 21 de la Ley 39/2015?",
		Activa:   true,
		Comandos: []ComandoEsperado{{Applet: "boe", Norma: lpac, Bloque: "a21"}},
		Citas:    []CitaEsperada{{Norma: lpac, Bloque: "a21"}},
	}
	noActivacion := Eval{Fichero: deNoActivacion, Pregunta: "¿Cómo invierto una lista enlazada en Go?"}
	articulo22 := Eval{
		Fichero:  deArticulo22,
		Pregunta: "¿qué dice el art. 22 de la Ley 39/2015?",
		Activa:   true,
		Comandos: []ComandoEsperado{{Applet: "boe", Verbo: "metadatos", Norma: lpac}},
		Citas:    []CitaEsperada{{Norma: lpac, Bloque: "a22"}},
	}
	comandoDel21 := Origen{Eval: nombreDeEval, Punto: PuntoComandoEsperado}
	normaDel21 := Origen{Eval: nombreDeEval, Punto: PuntoNormaDeLaEval}
	citaDel21 := Origen{Eval: nombreDeEval, Punto: PuntoCitaEsperada}
	comandoDel22 := Origen{Eval: deArticulo22, Punto: PuntoComandoEsperado}
	normaDel22 := Origen{Eval: deArticulo22, Punto: PuntoNormaDeLaEval}
	citaDel22 := Origen{Eval: deArticulo22, Punto: PuntoCitaEsperada}

	// Una eval de territorio sin nada del BOE y otra que resuelve el municipio y
	// además lee y cita el art. 21.
	resolver := ComandoEsperado{Applet: "territorio", Verbo: "resolver", Municipio: "Leganés"}
	soloTerritorio := Eval{
		Fichero:    deSoloTerritorio,
		Pregunta:   "¿En qué boletines se publican las normas que afectan a Leganés?",
		Activa:     true,
		Comandos:   []ComandoEsperado{resolver},
		Territorio: TerritorioEsperado{Comunidad: "Comunidad de Madrid", Boletines: []string{"BOCM"}},
	}
	territorioYArticulo21 := Eval{
		Fichero:    deTerritorioYArticulo21,
		Pregunta:   "¿Qué dice el art. 21 de la Ley 39/2015 y en qué boletines publica Leganés?",
		Activa:     true,
		Comandos:   []ComandoEsperado{resolver, {Applet: "boe", Norma: lpac, Bloque: "a21"}},
		Citas:      []CitaEsperada{{Norma: lpac, Bloque: "a21"}},
		Territorio: TerritorioEsperado{Comunidad: "Comunidad de Madrid"},
	}
	comandoDeTerritorioYArticulo21 := Origen{Eval: deTerritorioYArticulo21, Punto: PuntoComandoEsperado}
	normaDeTerritorioYArticulo21 := Origen{Eval: deTerritorioYArticulo21, Punto: PuntoNormaDeLaEval}
	citaDeTerritorioYArticulo21 := Origen{Eval: deTerritorioYArticulo21, Punto: PuntoCitaEsperada}

	// La eval de la consulta repetida, con la comprobación del grafo entre sus
	// comandos, graph show prohibido y el grafo previo del mismo bloque.
	bloqueDel21 := ComandoEsperado{Applet: "boe", Norma: lpac, Bloque: "a21"}
	consultaRepetida := Eval{
		Fichero:     deConsultaRepetida,
		Pregunta:    "Ya te pregunté hace tiempo por el artículo 21 de la Ley 39/2015. ¿Qué dice ahora?",
		Activa:      true,
		Informativa: true,
		GrafoPrevio: GrafoPrevio{Grabaciones: "lcsp-a1-30-redaccion-original", Comandos: []ComandoEsperado{bloqueDel21}},
		Comandos:    []ComandoEsperado{bloqueDel21, {Applet: "graph", Verbo: "check"}},
		Prohibidos:  []ComandoProhibido{{Applet: "graph", Verbo: "show"}},
		Citas:       []CitaEsperada{{Norma: lpac, Bloque: "a21"}},
	}
	comandoDeConsultaRepetida := Origen{Eval: deConsultaRepetida, Punto: PuntoComandoEsperado}
	normaDeConsultaRepetida := Origen{Eval: deConsultaRepetida, Punto: PuntoNormaDeLaEval}
	citaDeConsultaRepetida := Origen{Eval: deConsultaRepetida, Punto: PuntoCitaEsperada}

	// Una eval de plazo, que resuelve el municipio, calcula el plazo y lee y cita
	// el art. 124, y una de cobertura, sin comandos ni citas y con plazos calcular
	// prohibido.
	dePlazo := Eval{
		Fichero:  deUnPlazo,
		Pregunta: "¿Hasta cuándo puedo presentar el recurso de reposición contra una resolución de Leganés?",
		Activa:   true,
		Comandos: []ComandoEsperado{
			resolver, {Applet: "plazos", Verbo: "calcular", Tipo: "reposicion"}, {Applet: "boe", Norma: lpac, Bloque: "a124"},
		},
		Citas: []CitaEsperada{{Norma: lpac, Bloque: "a124"}},
		Plazo: PlazoEsperado{Fecha: "2026-10-13", Avisos: []string{}, Razonamiento: "Un mes desde el día siguiente."},
	}
	deCobertura := Eval{
		Fichero:    deUnaCobertura,
		Pregunta:   "¿Hasta cuándo puedo recurrir una liquidación de la Agencia Tributaria?",
		Activa:     true,
		Prohibidos: []ComandoProhibido{{Applet: "plazos", Verbo: "calcular"}},
		NoCubierto: NoCubiertoEsperado{Norma: "BOE-A-2003-23186"},
	}
	comandoDelPlazo := Origen{Eval: deUnPlazo, Punto: PuntoComandoEsperado}
	normaDelPlazo := Origen{Eval: deUnPlazo, Punto: PuntoNormaDeLaEval}
	citaDelPlazo := Origen{Eval: deUnPlazo, Punto: PuntoCitaEsperada}

	// Dos evals con fecha, como las de H20: la de la LCSP lee y cita el bloque
	// del comando con fecha; la de la LPAC solo lo lee, sin citas, para que sus
	// dos lecturas salgan solo del comando. Ninguna de las dos fechas entra en
	// una consulta: la del 2016-06-01 sale con 3, que sería una falta.
	contratoMenorEn2019 := Eval{
		Fichero:  deContratoMenorEn2019,
		Pregunta: "¿qué exigía el artículo 118 de la LCSP para el expediente de un contrato menor el 1 de junio de 2019?",
		Activa:   true,
		Comandos: []ComandoEsperado{{Applet: "boe", Norma: lcsp, Bloque: "a1-30", Fecha: "2019-06-01"}},
		Citas:    []CitaEsperada{{Norma: lcsp, Bloque: "a1-30"}},
	}
	articulo21En2016 := Eval{
		Fichero:  deArticulo21En2016,
		Pregunta: "¿qué decía el artículo 21 de la Ley 39/2015 el 1 de junio de 2016?",
		Activa:   true,
		Comandos: []ComandoEsperado{{Applet: "boe", Norma: lpac, Bloque: "a21", Fecha: "2016-06-01"}},
	}
	comandoDel2019 := Origen{Eval: deContratoMenorEn2019, Punto: PuntoComandoEsperado}
	normaDel2019 := Origen{Eval: deContratoMenorEn2019, Punto: PuntoNormaDeLaEval}
	citaDel2019 := Origen{Eval: deContratoMenorEn2019, Punto: PuntoCitaEsperada}
	comandoDel2016 := Origen{Eval: deArticulo21En2016, Punto: PuntoComandoEsperado}
	normaDel2016 := Origen{Eval: deArticulo21En2016, Punto: PuntoNormaDeLaEval}

	casos := []struct {
		nombre    string
		conjunto  []Eval
		esperadas []Consulta
	}{
		{
			nombre:   "tres-formas-citas-repetidas-y-dos-normas",
			conjunto: []Eval{tresFormas},
			esperadas: []Consulta{
				{
					Applet: "boe", Verbo: "articulo", Argumentos: []string{lpac, "a21"},
					Origenes: []Origen{comandoDeTresFormas, citaDeTresFormas},
				},
				{
					Applet: "boe", Verbo: "articulo", Argumentos: fechada(lpac, "a21"),
					Origenes: []Origen{comandoDeTresFormas, citaDeTresFormas},
				},
				{
					Applet: "boe", Verbo: "indice", Argumentos: []string{lpac},
					Origenes: []Origen{comandoDeTresFormas, normaDeTresFormas},
				},
				{
					Applet: "boe", Verbo: "buscar", Argumentos: []string{"procedimiento", "común"},
					Origenes: []Origen{comandoDeTresFormas},
				},
				{Applet: "boe", Verbo: "metadatos", Argumentos: []string{lpac}, Origenes: []Origen{normaDeTresFormas}},
				{Applet: "boe", Verbo: "indice", Argumentos: []string{otra}, Origenes: []Origen{normaDeTresFormas}},
				{Applet: "boe", Verbo: "metadatos", Argumentos: []string{otra}, Origenes: []Origen{normaDeTresFormas}},
				{Applet: "boe", Verbo: "articulo", Argumentos: []string{otra, "a5"}, Origenes: []Origen{citaDeTresFormas}},
				{Applet: "boe", Verbo: "articulo", Argumentos: fechada(otra, "a5"), Origenes: []Origen{citaDeTresFormas}},
			},
		},
		{
			nombre:   "evals-que-comparten-consultas",
			conjunto: []Eval{articulo21, noActivacion, articulo22},
			esperadas: []Consulta{
				{
					Applet: "boe", Verbo: "articulo", Argumentos: []string{lpac, "a21"},
					Origenes: []Origen{comandoDel21, citaDel21},
				},
				{
					Applet: "boe", Verbo: "articulo", Argumentos: fechada(lpac, "a21"),
					Origenes: []Origen{comandoDel21, citaDel21},
				},
				{Applet: "boe", Verbo: "indice", Argumentos: []string{lpac}, Origenes: []Origen{normaDel21, normaDel22}},
				{
					Applet: "boe", Verbo: "metadatos", Argumentos: []string{lpac},
					Origenes: []Origen{normaDel21, comandoDel22, normaDel22},
				},
				{Applet: "boe", Verbo: "articulo", Argumentos: []string{lpac, "a22"}, Origenes: []Origen{citaDel22}},
				{Applet: "boe", Verbo: "articulo", Argumentos: fechada(lpac, "a22"), Origenes: []Origen{citaDel22}},
			},
		},
		{
			nombre:   "territorio-sin-consultas",
			conjunto: []Eval{soloTerritorio, noActivacion},
		},
		{
			nombre:   "territorio-junto-al-boe",
			conjunto: []Eval{territorioYArticulo21},
			esperadas: []Consulta{
				{
					Applet: "boe", Verbo: "articulo", Argumentos: []string{lpac, "a21"},
					Origenes: []Origen{comandoDeTerritorioYArticulo21, citaDeTerritorioYArticulo21},
				},
				{
					Applet: "boe", Verbo: "articulo", Argumentos: fechada(lpac, "a21"),
					Origenes: []Origen{comandoDeTerritorioYArticulo21, citaDeTerritorioYArticulo21},
				},
				{
					Applet: "boe", Verbo: "indice", Argumentos: []string{lpac},
					Origenes: []Origen{normaDeTerritorioYArticulo21},
				},
				{
					Applet: "boe", Verbo: "metadatos", Argumentos: []string{lpac},
					Origenes: []Origen{normaDeTerritorioYArticulo21},
				},
			},
		},
		{
			nombre:   "comprobacion-sin-consultas",
			conjunto: []Eval{consultaRepetida},
			esperadas: []Consulta{
				{
					Applet: "boe", Verbo: "articulo", Argumentos: []string{lpac, "a21"},
					Origenes: []Origen{comandoDeConsultaRepetida, citaDeConsultaRepetida},
				},
				{
					Applet: "boe", Verbo: "articulo", Argumentos: fechada(lpac, "a21"),
					Origenes: []Origen{comandoDeConsultaRepetida, citaDeConsultaRepetida},
				},
				{Applet: "boe", Verbo: "indice", Argumentos: []string{lpac}, Origenes: []Origen{normaDeConsultaRepetida}},
				{Applet: "boe", Verbo: "metadatos", Argumentos: []string{lpac}, Origenes: []Origen{normaDeConsultaRepetida}},
			},
		},
		{
			nombre:   "cita-sin-consultas",
			conjunto: conjuntoDeJurisprudencia(),
		},
		{
			nombre:   "calcular-sin-consultas",
			conjunto: []Eval{dePlazo, deCobertura},
			esperadas: []Consulta{
				{
					Applet: "boe", Verbo: "articulo", Argumentos: []string{lpac, "a124"},
					Origenes: []Origen{comandoDelPlazo, citaDelPlazo},
				},
				{
					Applet: "boe", Verbo: "articulo", Argumentos: fechada(lpac, "a124"),
					Origenes: []Origen{comandoDelPlazo, citaDelPlazo},
				},
				{Applet: "boe", Verbo: "indice", Argumentos: []string{lpac}, Origenes: []Origen{normaDelPlazo}},
				{Applet: "boe", Verbo: "metadatos", Argumentos: []string{lpac}, Origenes: []Origen{normaDelPlazo}},
			},
		},
		{
			nombre:   "comandos-con-fecha",
			conjunto: []Eval{contratoMenorEn2019, articulo21En2016},
			esperadas: []Consulta{
				{
					Applet: "boe", Verbo: "articulo", Argumentos: []string{lcsp, "a1-30"},
					Origenes: []Origen{comandoDel2019, citaDel2019},
				},
				{
					Applet: "boe", Verbo: "articulo", Argumentos: fechada(lcsp, "a1-30"),
					Origenes: []Origen{comandoDel2019, citaDel2019},
				},
				{Applet: "boe", Verbo: "indice", Argumentos: []string{lcsp}, Origenes: []Origen{normaDel2019}},
				{Applet: "boe", Verbo: "metadatos", Argumentos: []string{lcsp}, Origenes: []Origen{normaDel2019}},
				{Applet: "boe", Verbo: "articulo", Argumentos: []string{lpac, "a21"}, Origenes: []Origen{comandoDel2016}},
				{Applet: "boe", Verbo: "articulo", Argumentos: fechada(lpac, "a21"), Origenes: []Origen{comandoDel2016}},
				{Applet: "boe", Verbo: "indice", Argumentos: []string{lpac}, Origenes: []Origen{normaDel2016}},
				{Applet: "boe", Verbo: "metadatos", Argumentos: []string{lpac}, Origenes: []Origen{normaDel2016}},
			},
		},
		{
			nombre: "sin-evals",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.esperadas, ConsultasNecesarias(caso.conjunto))
		})
	}
}

// Los orígenes de lo que la skill lleva a leer, escritos a mano (data-model §2
// de H26.1).
var (
	origenDeLosTipos    = Origen{Eval: "data/plazos.yaml", Punto: "lo que la skill lleva a leer"}
	origenDeLaCobertura = Origen{Eval: "data/cobertura.yaml", Punto: "lo que la skill lleva a leer"}
	origenDeLasNormas   = Origen{Eval: "data/normas.yaml", Punto: "lo que la skill lleva a leer"}
)

// TestLecturasDeLosDatos fija lecturasDeLosDatos y consultasDeLaCache
// (data-model §2 de H26.1; contracts/cache-de-la-sesion.md §1 de H26.1; research
// D1 de H26.1; FR-001, FR-002): del data/ del repositorio, las mismas
// invocaciones que enumera el oráculo de los tests de los preceptos, ninguna
// repetida, cada una con el fichero de data/ del que sale y el punto de lo que
// la skill lleva a leer; detrás de las de una eval que lee uno de esos bloques,
// con el mismo acumulador, de modo que la invocación que ya estaba suma su
// origen; y un fichero que falta o que su lector rechaza es el error, que nombra
// su ruta, sin consultas.
func TestLecturasDeLosDatos(t *testing.T) {
	t.Parallel()

	t.Run("del-repositorio", probarLasLecturasDelRepositorio)
	t.Run("detras-de-las-evals", probarLasLecturasDetrasDeLasEvals)
	t.Run("ficheros-que-no-se-leen", probarLosDatosQueNoSeLeen)
}

// probarLasLecturasDelRepositorio es el subtest del-repositorio de
// TestLecturasDeLosDatos.
func probarLasLecturasDelRepositorio(t *testing.T) {
	t.Parallel()

	consultas, err := lecturasDeLosDatos(datosDelRepositorio)
	require.NoError(t, err)

	origenes := make(map[string][]Origen, len(consultas))
	ordenes := make([]string, 0, len(consultas))

	for _, consulta := range consultas {
		ordenes = append(ordenes, ordenDe(consulta))
		origenes[ordenDe(consulta)] = consulta.Origenes

		assert.NotEmpty(t, consulta.Origenes, "%s sale de algún fichero de data/", ordenDe(consulta))

		for _, origen := range consulta.Origenes {
			assert.Contains(t, []Origen{origenDeLosTipos, origenDeLaCobertura, origenDeLasNormas}, origen,
				"%s sale de un fichero de data/, como lo que la skill lleva a leer", ordenDe(consulta))
		}
	}

	esperadas := make([]string, 0, len(consultas))
	for _, consulta := range loQueLlevaALeerElRepositorio(t) {
		esperadas = append(esperadas, ordenDe(consulta))
	}

	assert.ElementsMatch(t, esperadas, ordenes, "las invocaciones que enumera el oráculo, cada una una vez")

	// Cada bloque y cada norma, con el fichero que la nombra entre sus orígenes.
	tipos, err := plazos.LeerTipos(contenidoDelFichero(t, tiposDelRepositorio))
	require.NoError(t, err)

	cobertura, err := skills.LeerCobertura(contenidoDelFichero(t, coberturaDelRepositorio))
	require.NoError(t, err)

	tabla, err := skills.LeerNormas(contenidoDelFichero(t, tablaDeNormasDelRepositorio))
	require.NoError(t, err)

	nombradas := map[Origen][]lecturaDeUnPrecepto{
		origenDeLosTipos:    lecturasDeLosTipos(tipos),
		origenDeLaCobertura: lecturasDeLaCobertura(cobertura),
	}
	for _, norma := range tabla {
		nombradas[origenDeLasNormas] = append(nombradas[origenDeLasNormas], lecturaDeUnPrecepto{norma: norma.Identificador})
	}

	for origen, lecturas := range nombradas {
		for _, lectura := range lecturas {
			ordenesDeLaLectura := []string{"boe indice " + lectura.norma, "boe metadatos " + lectura.norma}
			if lectura.bloque != "" {
				deLaLectura := "boe articulo " + lectura.norma + " " + lectura.bloque
				ordenesDeLaLectura = append(ordenesDeLaLectura, deLaLectura, deLaLectura+" --fecha 9999-12-31")
			}

			for _, orden := range ordenesDeLaLectura {
				assert.Contains(t, origenes[orden], origen, "%s sale, entre otros, de %s", orden, origen.Eval)
			}
		}
	}
}

// probarLasLecturasDetrasDeLasEvals es el subtest detras-de-las-evals de
// TestLecturasDeLosDatos, con una eval que espera el primer bloque que lleva a
// leer data/.
func probarLasLecturasDetrasDeLasEvals(t *testing.T) {
	t.Parallel()

	lecturas, err := lecturasDeLosDatos(datosDelRepositorio)
	require.NoError(t, err)
	require.NotEmpty(t, lecturas)

	primera := lecturas[0]
	require.Equal(t, "articulo", primera.Verbo, "data/ empieza por la lectura de un bloque")
	require.Len(t, primera.Argumentos, 2, "sin fecha")

	eval := Eval{
		Fichero:  "01-un-bloque-de-los-datos.yaml",
		Comandos: []ComandoEsperado{{Applet: "boe", Norma: primera.Argumentos[0], Bloque: primera.Argumentos[1]}},
	}
	necesarias := ConsultasNecesarias([]Eval{eval})

	consultas, err := consultasDeLaCache([]Eval{eval}, datosDelRepositorio)
	require.NoError(t, err)

	ordenes := make([]string, 0, len(consultas))
	for _, consulta := range consultas {
		ordenes = append(ordenes, ordenDe(consulta))
	}

	require.Len(t, consultas, len(slices.Compact(slices.Sorted(slices.Values(ordenes)))), "ninguna invocación repetida")

	// Primero las de la eval, en su orden; la que ya estaba suma, detrás de los
	// de la eval, el origen de data/.
	for posicion, necesaria := range necesarias {
		assert.Equal(t, ordenDe(necesaria), ordenes[posicion], "las de la eval van primero, en su orden")
	}

	assert.Equal(t, slices.Concat([]Origen{{Eval: eval.Fichero, Punto: PuntoComandoEsperado}}, primera.Origenes),
		consultas[0].Origenes, "%s suma el origen de data/ a los de la eval", ordenDe(primera))

	// Y detrás, las de data/ que la eval no tenía, cada una una vez.
	nuevas := slices.DeleteFunc(slices.Clone(lecturas), func(lectura Consulta) bool {
		return slices.ContainsFunc(necesarias, func(necesaria Consulta) bool { return ordenDe(necesaria) == ordenDe(lectura) })
	})
	assert.Len(t, consultas, len(necesarias)+len(nuevas))
}

// probarLosDatosQueNoSeLeen es el subtest ficheros-que-no-se-leen de
// TestLecturasDeLosDatos: con una copia de los tres ficheros de data/ en la que
// falta uno, o en la que uno es una lista y no un mapa, lecturasDeLosDatos y
// consultasDeLaCache dan el error que nombra la ruta, y ninguna consulta.
func probarLosDatosQueNoSeLeen(t *testing.T) {
	t.Parallel()

	ficheros := []string{"plazos.yaml", "cobertura.yaml", "normas.yaml"}
	casos := []struct {
		sufijo string
		falta  bool
	}{
		{sufijo: "-que-falta", falta: true},
		{sufijo: "-que-se-rechaza"},
	}

	for _, fichero := range ficheros {
		for _, caso := range casos {
			t.Run(strings.TrimSuffix(fichero, ".yaml")+caso.sufijo, func(t *testing.T) {
				t.Parallel()

				datos := t.TempDir()
				for _, copiado := range ficheros {
					require.NoError(t, os.WriteFile(filepath.Join(datos, copiado),
						contenidoDelFichero(t, filepath.Join(datosDelRepositorio, copiado)), 0o600))
				}

				ruta := filepath.Join(datos, fichero)
				if caso.falta {
					require.NoError(t, os.Remove(ruta))
				} else {
					require.NoError(t, os.WriteFile(ruta, []byte("- no es un mapa\n"), 0o600))
				}

				lecturas, err := lecturasDeLosDatos(datos)
				exigeElErrorDeLosDatos(t, err, ruta, caso.falta)
				assert.Nil(t, lecturas)

				consultas, err := consultasDeLaCache(nil, datos)
				exigeElErrorDeLosDatos(t, err, ruta, caso.falta)
				assert.Nil(t, consultas)
			})
		}
	}
}

// exigeElErrorDeLosDatos exige el error de lo que la skill lleva a leer que
// nombra la ruta: el del fichero que no se puede leer, si falta, o el del que
// su lector rechaza.
func exigeElErrorDeLosDatos(t *testing.T, err error, ruta string, falta bool) {
	t.Helper()

	if falta {
		require.ErrorIs(t, err, fs.ErrNotExist)
		require.ErrorContains(t, err, "lo que la skill lleva a leer: "+ruta+" no se puede leer: ")

		return
	}

	require.ErrorContains(t, err, "lo que la skill lleva a leer: "+ruta+": ")
}
