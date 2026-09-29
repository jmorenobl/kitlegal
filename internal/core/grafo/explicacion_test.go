package grafo_test

import (
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/grafo"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Los ids de ejemplo que solo usan las explicaciones: otra norma y otro
// bloque de la de ejemplo.
const (
	idOtraNorma            = "eli/es/l/2015/10/01/40"
	identificadorOtraNorma = "BOE-A-2015-10566"
	citaDelBloque          = "[" + identificadorLPAC + ", bloque a21]"
)

// Las versiones del bloque a21 del e2e del cambio de versión: la grabada, que
// se consultó el lunes, y la derivada, de fecha de vigencia posterior, que se
// consultó el martes (contracts/arnes-e2e.md §3).
var (
	grabada  = versionDeEjemplo{fechaVigencia: "20161002", letra: "a", consulta: lunes}
	derivada = versionDeEjemplo{fechaVigencia: "20250101", letra: "b", consulta: martes}
)

// grafoDelBloque es el que deja consultar el bloque a21 con la versión grabada
// y con la derivada: la norma, el bloque y las dos versiones, cada uno con la
// vigencia que declaró su consulta, y las aristas que los unen (contracts/
// emision.md §1). No lleva filas de lecturas, así que la redacción vista del
// bloque es la de RedaccionVistaSinLecturas: la derivada, la de observación
// más reciente (H7.1 FR-026).
func grafoDelBloque(vigencia time.Duration) grafo.Instantanea {
	return grafo.Instantanea{
		Nodos: []grafo.NodoDeInstantanea{
			conVigenciaDe(normaLPAC(lunes), vigencia),
			conVigenciaDe(bloqueDeLaLPAC(idBloque, "a21", lunes), vigencia),
			grabada.conVigencia(vigencia).nodo(),
			derivada.conVigencia(vigencia).nodo(),
		},
		Aristas: []schema.Arista{
			arista(idNorma, grafo.RelacionTieneParte, idBloque),
			arista(idBloque, grafo.RelacionTieneVersion, grabada.id()),
			arista(idBloque, grafo.RelacionTieneVersion, derivada.id()),
		},
	}
}

// leidaLaDerivada es la misma instantánea con la fila de lecturas del bloque
// a21 que dejan sus dos consultas: la primera vio la grabada y la segunda, la
// derivada (data-model §4).
func leidaLaDerivada(instantanea grafo.Instantanea) grafo.Instantanea {
	primera := grafo.LecturasDeBloque{Bloque: idBloque, Ultima: grabada.id(), Anterior: grabada.id()}
	instantanea.Lecturas = []grafo.LecturasDeBloque{primera.Leida(derivada.id())}

	return instantanea
}

// consultaCaducada es la explicación de una fuente-caducada con la plantilla
// de contracts/applet-graph.md §5, que el primer caso de
// probarPlantillaDeFuenteCaducada escribe entera.
func consultaCaducada(cita string, observacion grafo.Procedencia, segundos int64, caducidad string) string {
	return "La consulta de " + cita + " a " + observacion.Fuente + " en " + observacion.URL + " del " +
		observacion.FechaConsulta + " ten\xc3\xada una vigencia de " + strconv.FormatInt(segundos, 10) +
		" s y caduc\xc3\xb3 el " + caducidad + "."
}

// TestExplicaciones fija las explicaciones de los hallazgos de `graph check`
// (FR-061, FR-063, FR-066; contracts/applet-graph.md §5; data-model §6;
// research.md D18): las dos plantillas literales, con cada valor copiado
// carácter a carácter de lo guardado; el instante de caducidad en el
// desplazamiento de la fecha de consulta, escrito con RFC3339Nano; y la cita de
// cada tipo de nodo que recibe un hallazgo, o su id cuando faltan los datos que
// la forman.
func TestExplicaciones(t *testing.T) {
	t.Parallel()

	t.Run("plantilla de version-obsoleta", probarPlantillaDeVersionObsoleta)
	t.Run("plantilla de fuente-caducada", probarPlantillaDeFuenteCaducada)
	t.Run("instante de caducidad", probarInstanteDeCaducidad)
	t.Run("cita de cada tipo de nodo", probarCitas)
}

// probarPlantillaDeVersionObsoleta fija la explicación de una versión
// superada: la cita del bloque, las dos fechas de vigencia y la url y la fecha
// de consulta de la última observación de la versión que vio la última
// lectura, tal como se guardaron, también con otro desplazamiento y con los «&»
// de la url.
func probarPlantillaDeVersionObsoleta(t *testing.T) {
	t.Parallel()

	hallazgos, err := grafo.Comprobar(leidaLaDerivada(grafoDelBloque(0)), elMartes)
	require.NoError(t, err)
	require.Len(t, hallazgos, 1)
	assert.Equal(t, "La versi\xc3\xb3n de [BOE-A-2015-10565, bloque a21] con fecha de vigencia 20161002 "+
		"est\xc3\xa1 superada por la de fecha de vigencia 20250101, observada en "+
		"https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565/texto/bloque/a21 "+
		"el 2026-09-29T12:00:00Z.", hallazgos[0].Explicacion)

	instantanea := leidaLaDerivada(grafoDelBloque(0))
	instantanea.Nodos[3].UltimaObservacion = grafo.Procedencia{
		Fuente: fuenteCorta, URL: urlA22 + "?a=1&b=2", FechaConsulta: "2026-09-29T14:00:00.250+02:00",
	}

	hallazgos, err = grafo.Comprobar(instantanea, elMartes)
	require.NoError(t, err)
	require.Len(t, hallazgos, 1)
	assert.Equal(t, "La versi\xc3\xb3n de "+citaDelBloque+" con fecha de vigencia 20161002 est\xc3\xa1 superada "+
		"por la de fecha de vigencia 20250101, observada en "+urlA22+"?a=1&b=2 el 2026-09-29T14:00:00.250+02:00.",
		hallazgos[0].Explicacion)
}

// probarPlantillaDeFuenteCaducada fija la explicación de una consulta
// caducada de cada nodo del bloque que la recibe —la Norma, el Bloque y su
// redacción vista, la derivada—: la cita, la fuente, la url y la fecha de
// consulta de su última observación, la vigencia en segundos y el instante en
// que caducó. La grabada, que no es la redacción vista, no recibe ningún
// hallazgo aunque su consulta también haya caducado (H7.1 FR-030).
func probarPlantillaDeFuenteCaducada(t *testing.T) {
	t.Parallel()

	hallazgos, err := grafo.Comprobar(grafoDelBloque(unaSemana), dentroDeUnMes)
	require.NoError(t, err)

	explicaciones := map[string]string{}

	for _, hallazgo := range hallazgos {
		assert.NotEqual(t, grabada.id(), hallazgo.ID, "la grabada no es la redacci\xc3\xb3n vista")

		if hallazgo.Clase == grafo.ClaseFuenteCaducada {
			explicaciones[hallazgo.ID] = hallazgo.Explicacion
		}
	}

	delLunes := grafo.Procedencia{Fuente: fuenteBOE, URL: urlA21, FechaConsulta: lunes}
	delMartes := grafo.Procedencia{Fuente: fuenteBOE, URL: urlA21, FechaConsulta: martes}

	assert.Equal(t, map[string]string{
		idNorma: "La consulta de BOE-A-2015-10565 a boe.legislacion-consolidada en " +
			"https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-2015-10565/texto/bloque/a21 " +
			"del 2026-09-28T12:00:00Z ten\xc3\xada una vigencia de 604800 s y caduc\xc3\xb3 el 2026-10-05T12:00:00Z.",
		idBloque:      consultaCaducada(citaDelBloque, delLunes, 604800, "2026-10-05T12:00:00Z"),
		derivada.id(): consultaCaducada(citaDelBloque, delMartes, 604800, "2026-10-06T12:00:00Z"),
	}, explicaciones)
}

// probarInstanteDeCaducidad fija el instante en que caduca una consulta: su
// fecha de consulta más la vigencia, en el mismo desplazamiento aunque entre
// medias cambie la hora de alguna zona, con Z si la fecha lleva Z y con la
// fracción de segundo sin ceros a la derecha, solo si no es cero; la fecha de
// consulta se copia tal cual se guardó (FR-066; research.md D18).
func probarInstanteDeCaducidad(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		fecha     string
		vigencia  time.Duration
		caducidad string
	}{
		{"con medio segundo", "2026-09-28T12:00:00.5+02:00", unaSemana, "2026-10-05T12:00:00.5+02:00"},
		{"con doce centesimas", "2026-09-28T12:00:00.12+02:00", unaSemana, "2026-10-05T12:00:00.12+02:00"},
		{"en UTC", "2026-09-28T12:00:00Z", unaSemana, "2026-10-05T12:00:00Z"},
		{"con un cero a la derecha", "2026-09-28T12:00:00.120+02:00", unaSemana, "2026-10-05T12:00:00.12+02:00"},
		{"con fraccion cero", "2026-09-28T12:00:00.000Z", unaSemana, "2026-10-05T12:00:00Z"},
		{"a traves de un cambio de hora", "2026-10-20T12:00:00+02:00", unaSemana, "2026-10-27T12:00:00+02:00"},
		{"con un desplazamiento negativo", "2026-09-28T12:00:00-03:30", unaHora, "2026-09-28T13:00:00-03:30"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			nodo := conVigenciaDe(normaLPAC(caso.fecha), caso.vigencia)

			hallazgos, err := grafo.Comprobar(deNodos(nodo), dentroDeUnMes)
			require.NoError(t, err)
			require.Len(t, hallazgos, 1)
			assert.Equal(t, consultaCaducada(identificadorLPAC, nodo.UltimaObservacion,
				int64(caso.vigencia/time.Second), caso.caducidad), hallazgos[0].Explicacion)
		})
	}
}

// casoDeCita es un grafo, la cita que llevan, en sus explicaciones, los nodos
// que se nombran, y los nodos que no reciben ningún hallazgo.
type casoDeCita struct {
	nombre      string
	instantanea grafo.Instantanea
	citas       map[string]string
	sinHallazgo []string
}

// probarCitas fija la cita de cada tipo de nodo que recibe un hallazgo
// (data-model §6; H7.1 FR-030): el identificador de una Norma; el de la Norma
// de la arista eli:has_part que llega a un Bloque y su bloque; la de su Bloque,
// por la arista eli:has_version que llega a una BloqueVersion que es la
// redacción vista de un bloque; y el id del nodo si falta alguno de esos datos,
// si no es texto o está vacío, si la arista no llega de un nodo de ese tipo o
// si llegan dos y la cita no es una sola. Una BloqueVersion que no es la
// redacción vista de ningún bloque, un Municipio y un Organo no reciben ninguno.
func probarCitas(t *testing.T) {
	t.Parallel()

	for _, caso := range casosDeCita() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			exigirCitas(t, caso)
		})
	}
}

// casosDeCita son los de probarCitas, cada uno sobre el grafo del bloque con
// sus consultas caducadas. Sin filas de lecturas, la redacción vista del bloque
// a21 es la derivada, así que la cita de una BloqueVersion se prueba sobre
// ella, y la grabada no recibe ningún hallazgo; en «una BloqueVersion de dos
// Bloques», que prueba la cita de la grabada, el caso añade la fila de lecturas
// que la hace la redacción vista del bloque a22.
func casosDeCita() []casoDeCita {
	municipio := conVigenciaDe(observado(idMunicipio, grafo.TipoMunicipio,
		map[string]any{grafo.DatoCodigoINE: "28074"}, lunes), unaSemana)
	organo := conVigenciaDe(observado(idOrgano, grafo.TipoOrgano, map[string]any{grafo.DatoDIR3: idOrgano}, lunes), unaSemana)
	otraNorma := conVigenciaDe(observado(idOtraNorma, grafo.TipoNorma,
		map[string]any{grafo.DatoIdentificador: identificadorOtraNorma}, lunes), unaSemana)
	bloqueA22 := conVigenciaDe(bloqueDeLaLPAC(idBloqueA22, "a22", lunes), unaSemana)
	sinCitas := map[string]string{idNorma: idNorma, idBloque: idBloque, derivada.id(): derivada.id()}
	laGrabada := []string{grabada.id()}
	laGrabadaYElMunicipio := []string{grabada.id(), idMunicipio}

	return []casoDeCita{
		{"cada tipo con sus datos", grafoDelBloque(unaSemana), map[string]string{
			idNorma: identificadorLPAC, idBloque: citaDelBloque, derivada.id(): citaDelBloque,
		}, laGrabada},
		{"una Norma sin identificador", conDatos(idNorma, map[string]any{}), sinCitas, laGrabada},
		{
			"un identificador que no es texto", conDatos(idNorma, map[string]any{grafo.DatoIdentificador: float64(10565)}),
			sinCitas, laGrabada,
		},
		{"un identificador vacio", conDatos(idNorma, map[string]any{grafo.DatoIdentificador: ""}), sinCitas, laGrabada},
		{"un Bloque sin bloque", conDatos(idBloque, map[string]any{}), map[string]string{
			idNorma: identificadorLPAC, idBloque: idBloque, derivada.id(): derivada.id(),
		}, laGrabada},
		{"un bloque vacio", conDatos(idBloque, map[string]any{grafo.DatoBloque: ""}), map[string]string{
			idBloque: idBloque, derivada.id(): derivada.id(),
		}, laGrabada},
		{"un Bloque sin su Norma", sinArista(arista(idNorma, grafo.RelacionTieneParte, idBloque)), map[string]string{
			idNorma: identificadorLPAC, idBloque: idBloque, derivada.id(): derivada.id(),
		}, laGrabada},
		{
			"un Bloque de dos Normas",
			conMas(grafoDelBloque(unaSemana), []grafo.NodoDeInstantanea{otraNorma},
				arista(idOtraNorma, grafo.RelacionTieneParte, idBloque)),
			map[string]string{idOtraNorma: identificadorOtraNorma, idBloque: idBloque, derivada.id(): derivada.id()},
			laGrabada,
		},
		{
			"un Bloque que cuelga de un nodo que no es una Norma",
			conMas(sinArista(arista(idNorma, grafo.RelacionTieneParte, idBloque)), []grafo.NodoDeInstantanea{municipio},
				arista(idMunicipio, grafo.RelacionTieneParte, idBloque)),
			map[string]string{idBloque: idBloque, derivada.id(): derivada.id()},
			laGrabadaYElMunicipio,
		},
		{
			"un Bloque de una Norma y de un nodo que no lo es",
			conMas(grafoDelBloque(unaSemana), []grafo.NodoDeInstantanea{municipio},
				arista(idMunicipio, grafo.RelacionTieneParte, idBloque)),
			map[string]string{idBloque: citaDelBloque, derivada.id(): citaDelBloque},
			laGrabadaYElMunicipio,
		},
		{
			"una BloqueVersion sin su Bloque",
			sinArista(arista(idBloque, grafo.RelacionTieneVersion, grabada.id())),
			map[string]string{derivada.id(): citaDelBloque},
			laGrabada,
		},
		{
			"una BloqueVersion de dos Bloques",
			conFila(conMas(grafoDelBloque(unaSemana), []grafo.NodoDeInstantanea{bloqueA22},
				arista(idNorma, grafo.RelacionTieneParte, idBloqueA22),
				arista(idBloqueA22, grafo.RelacionTieneVersion, grabada.id())),
				grafo.LecturasDeBloque{Bloque: idBloqueA22, Ultima: grabada.id(), Anterior: grabada.id()}),
			map[string]string{
				grabada.id(): grabada.id(), derivada.id(): citaDelBloque, idBloqueA22: "[" + identificadorLPAC + ", bloque a22]",
			},
			nil,
		},
		{
			"una BloqueVersion que cuelga de un nodo que no es un Bloque",
			conMas(sinArista(arista(idBloque, grafo.RelacionTieneVersion, grabada.id())), nil,
				arista(idNorma, grafo.RelacionTieneVersion, grabada.id())),
			map[string]string{derivada.id(): citaDelBloque},
			laGrabada,
		},
		{
			"los tipos sin cita",
			conMas(grafoDelBloque(unaSemana), []grafo.NodoDeInstantanea{municipio, organo},
				arista(idOrgano, grafo.RelacionPerteneceA, idMunicipio)),
			map[string]string{derivada.id(): citaDelBloque},
			[]string{grabada.id(), idMunicipio, idOrgano},
		},
	}
}

// conFila es la misma instantánea con una fila de lecturas más.
func conFila(instantanea grafo.Instantanea, fila grafo.LecturasDeBloque) grafo.Instantanea {
	instantanea.Lecturas = append(slices.Clone(instantanea.Lecturas), fila)

	return instantanea
}

// conDatos es el grafo del bloque, con sus consultas caducadas, con otros
// datos en el nodo de ese id.
func conDatos(id string, datos map[string]any) grafo.Instantanea {
	instantanea := grafoDelBloque(unaSemana)

	for i := range instantanea.Nodos {
		if instantanea.Nodos[i].ID == id {
			instantanea.Nodos[i].Datos = datos
		}
	}

	return instantanea
}

// sinArista es el grafo del bloque, con sus consultas caducadas, sin esa
// arista.
func sinArista(quitada schema.Arista) grafo.Instantanea {
	instantanea := grafoDelBloque(unaSemana)
	aristas := instantanea.Aristas[:0]

	for _, arista := range instantanea.Aristas {
		if arista != quitada {
			aristas = append(aristas, arista)
		}
	}

	instantanea.Aristas = aristas

	return instantanea
}

// exigirCitas exige que cada hallazgo de un nodo nombrado en el caso empiece
// su explicación con la cita del caso, que cada nodo nombrado tenga alguno y
// que ninguno sea de un nodo que el caso dice que no lo recibe.
func exigirCitas(t *testing.T, caso casoDeCita) {
	t.Helper()

	hallazgos, err := grafo.Comprobar(caso.instantanea, dentroDeUnMes)
	require.NoError(t, err)

	citados := map[string]bool{}

	for _, hallazgo := range hallazgos {
		assert.NotContains(t, caso.sinHallazgo, hallazgo.ID, "no recibe ning\xc3\xban hallazgo")

		cita, nombrado := caso.citas[hallazgo.ID]
		if !nombrado {
			continue
		}

		inicio := "La consulta de " + cita + " a "
		if hallazgo.Clase == grafo.ClaseVersionObsoleta {
			inicio = "La versi\xc3\xb3n de " + cita + " con fecha de vigencia "
		}

		assert.Equal(t, inicio, hallazgo.Explicacion[:min(len(inicio), len(hallazgo.Explicacion))], hallazgo.ID)

		citados[hallazgo.ID] = true
	}

	assert.Len(t, citados, len(caso.citas))
}
