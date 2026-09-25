package evals

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
func TestConsultasNecesarias(t *testing.T) {
	t.Parallel()

	const (
		lpac = "BOE-A-2015-10565"

		// otra es inventada —ninguna norma es del año 0000—: ConsultasNecesarias
		// solo copia los identificadores.
		otra = "BOE-A-0000-2"

		deTresFormas            = "01-lpac-tres-formas.yaml"
		deNoActivacion          = "02-no-activa-programacion.yaml"
		deArticulo22            = "03-lpac-articulo-22.yaml"
		deSoloTerritorio        = "04-territorio-municipio-cubierto.yaml"
		deTerritorioYArticulo21 = "05-territorio-y-articulo-21.yaml"
	)

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
				{Applet: "boe", Verbo: "indice", Argumentos: []string{lpac}, Origenes: []Origen{normaDel21, normaDel22}},
				{
					Applet: "boe", Verbo: "metadatos", Argumentos: []string{lpac},
					Origenes: []Origen{normaDel21, comandoDel22, normaDel22},
				},
				{Applet: "boe", Verbo: "articulo", Argumentos: []string{lpac, "a22"}, Origenes: []Origen{citaDel22}},
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
