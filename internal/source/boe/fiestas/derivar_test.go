package fiestas

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"

	"github.com/jmorenobl/kitlegal/internal/core/plazos"
	"github.com/jmorenobl/kitlegal/internal/core/territorio"
)

// actualizarFestivos es la bandera con la que TestFestivosNacionalesDerivados
// escribe data/festivos/nacionales-y-autonomicos.yaml desde lo grabado antes de
// compararlo, como -actualizar-golden (contracts/datos-de-plazos.md §1.3). Solo
// la usa la tarea [datos] que deriva el fichero, y lo escrito lo revisa una
// persona: ningún fichero de data/festivos/ se escribe a mano (FR-020).
var actualizarFestivos = flag.Bool("actualizar-festivos", false,
	"escribe data/festivos/nacionales-y-autonomicos.yaml desde lo grabado antes de compararlo")

// Dónde están, relativos al directorio del paquete, el fichero que se deriva y
// el territorio con el que se cruza la cabecera de cada resolución.
const (
	rutaDelFichero    = "../../../../data/festivos/nacionales-y-autonomicos.yaml"
	raizDelTerritorio = "../../../../data/territorio/"
)

// cabeceraDelFichero es lo que nombra su línea de fichero generado
// (contracts/datos-de-plazos.md §1.1).
var cabeceraDelFichero = plazos.Cabecera{
	Desde:   "evidencias/h9/" + NombreDeLaFuente + "/",
	Test:    "TestFestivosNacionalesDerivados",
	Paquete: "internal/source/boe/fiestas",
}

// aniosDeLasResoluciones son los años de las tres resoluciones de FR-020.
var aniosDeLasResoluciones = []int{2025, 2026, 2027}

// registroDelRepositorio es el territorio de data/territorio/, leído del árbol y
// cargado con territorio.Cargar, como lo carga el binario.
func registroDelRepositorio(t *testing.T) *territorio.Registro {
	t.Helper()

	leer := func(ruta string) []byte {
		contenido, err := os.ReadFile(filepath.Clean(filepath.Join(raizDelTerritorio, ruta)))
		require.NoError(t, err)

		return contenido
	}

	fuentes := territorio.Fuentes{
		Municipios:  leer("municipios.yaml"),
		DIR3:        leer("dir3.yaml"),
		Estado:      leer("estado.yaml"),
		Comunidades: map[string][]byte{},
	}

	entradas, err := os.ReadDir(filepath.Join(raizDelTerritorio, "comunidades"))
	require.NoError(t, err)

	for _, entrada := range entradas {
		codigo, esYAML := strings.CutSuffix(entrada.Name(), ".yaml")
		require.True(t, esYAML, "data/territorio/comunidades/%s no es el fichero de una comunidad", entrada.Name())

		fuentes.Comunidades[codigo] = leer(filepath.Join("comunidades", entrada.Name()))
	}

	registro, err := territorio.Cargar(fuentes)
	require.NoError(t, err)

	return registro
}

// comprobarFichero es lo que FR-083 y SC-006 piden del fichero escrito, además
// de ser igual a lo que sale de lo grabado: que se lee en su forma, que es de
// la fuente y de los años del manifiesto, que cada fila lleva su source y es la
// resolución de su año, que cada comunidad del registro tiene festivos de toda
// ella en cada año (contracts/datos-de-plazos.md §1.1) y que el calendario lo
// carga. Devuelve todos los defectos juntos.
func comprobarFichero(contenido []byte, manifiesto Manifiesto, comunidades []territorio.Comunidad) error {
	var fichero plazos.FestivosNacionales

	lector := yaml.NewDecoder(bytes.NewReader(contenido))
	lector.KnownFields(true)

	if err := lector.Decode(&fichero); err != nil {
		return fmt.Errorf("el fichero no se lee en su forma: %w", err)
	}

	var defectos []error

	if fichero.Fuente != NombreDeLaFuente {
		defectos = append(defectos, fmt.Errorf("la fuente es %q y no %q", fichero.Fuente, NombreDeLaFuente))
	}

	disposicionDe := map[string]string{}
	esperados := make([]plazos.AnioNacional, 0, len(manifiesto.Disposiciones))

	for _, disposicion := range manifiesto.Disposiciones {
		disposicionDe[fmt.Sprint(disposicion.Anio)] = disposicion.ID
		esperados = append(esperados, plazos.AnioNacional{Anio: disposicion.Anio, Disposicion: disposicion.ID})
	}

	if !slices.Equal(fichero.Anios, esperados) {
		defectos = append(defectos, fmt.Errorf("los años son %v y el manifiesto da %v", fichero.Anios, esperados))
	}

	for _, festivo := range fichero.Festivos {
		anio, _, _ := strings.Cut(festivo.Fecha, "-")

		switch {
		case festivo.Source == "":
			defectos = append(defectos, fmt.Errorf("la fila del %s (%s) no lleva source", festivo.Fecha, festivo.Ambito))
		case festivo.Source != disposicionDe[anio]:
			defectos = append(defectos, fmt.Errorf("la fila del %s lleva el source %s, que no es la resolución de %s",
				festivo.Fecha, festivo.Source, anio))
		}
	}

	for _, disposicion := range manifiesto.Disposiciones {
		for _, comunidad := range comunidades {
			if !conFestivosDeToda(fichero.Festivos, disposicion.Anio, comunidad.Codigo) {
				defectos = append(defectos, fmt.Errorf("%s (%s) no tiene festivos de toda ella en %d",
					comunidad.Nombre, comunidad.Codigo, disposicion.Anio))
			}
		}
	}

	if _, err := plazos.CargarCalendario(contenido, nil); err != nil {
		defectos = append(defectos, err)
	}

	return errors.Join(defectos...)
}

// conFestivosDeToda dice si alguna fila del año, sin parte, es de la comunidad.
func conFestivosDeToda(festivos []plazos.FestivoNacional, anio int, comunidad string) bool {
	return slices.ContainsFunc(festivos, func(festivo plazos.FestivoNacional) bool {
		return strings.HasPrefix(festivo.Fecha, fmt.Sprintf("%d-", anio)) && festivo.Parte == "" &&
			slices.Contains(festivo.Comunidades, comunidad)
	})
}

// TestFestivosNacionalesDerivados deriva data/festivos/nacionales-y-autonomicos.yaml
// desde las tres resoluciones grabadas y el territorio real (FR-020, FR-025,
// FR-083; SC-006; research D14) y falla si el fichero del árbol no es igual
// byte a byte a lo derivado, si una fila no lleva su source o si una de las
// comunidades no tiene festivos en 2025, 2026 o 2027. Con -actualizar-festivos
// escribe antes el fichero; sin la bandera no escribe nada. El último subtest
// demuestra que el comprobador no pasa en vacío.
func TestFestivosNacionalesDerivados(t *testing.T) {
	t.Parallel()

	registro := registroDelRepositorio(t)
	manifiesto := manifiestoDelPaquete(t)

	anios := make([]int, 0, len(manifiesto.Disposiciones))
	for _, disposicion := range manifiesto.Disposiciones {
		anios = append(anios, disposicion.Anio)
	}

	require.Equal(t, aniosDeLasResoluciones, anios, "el manifiesto lleva las tres resoluciones de FR-020")
	require.Len(t, registro.Comunidades(), numeroDeComunidades)

	fichero, err := Derivar(t.Context(), clienteDeLoGrabado(t), manifiesto, registro)
	require.NoError(t, err)

	derivado, err := plazos.EscribirFestivosNacionales(fichero, cabeceraDelFichero)
	require.NoError(t, err)

	if *actualizarFestivos {
		require.NoError(t, os.MkdirAll(filepath.Dir(rutaDelFichero), 0o750))
		require.NoError(t, os.WriteFile(rutaDelFichero, derivado, 0o600))
	}

	t.Run("data-festivos", func(t *testing.T) {
		t.Parallel()

		escrito, err := os.ReadFile(rutaDelFichero)
		require.NoError(t, err, "el fichero lo escribe este test con -actualizar-festivos")

		require.NoError(t, comprobarFichero(escrito, manifiesto, registro.Comunidades()))
		assert.Equal(t, string(derivado), string(escrito),
			"%s no es lo que sale de lo grabado: se regenera con -actualizar-festivos, nunca a mano", rutaDelFichero)
	})

	t.Run("el-comprobador-no-pasa-en-vacio", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, comprobarFichero(derivado, manifiesto, registro.Comunidades()))
		probarElComprobador(t, fichero, manifiesto, registro.Comunidades())
	})
}

// probarElComprobador cambia el fichero derivado de cada forma que FR-083 y
// SC-006 prohíben y comprueba que comprobarFichero lo rechaza nombrándolo.
func probarElComprobador(t *testing.T, fichero plazos.FestivosNacionales, manifiesto Manifiesto,
	comunidades []territorio.Comunidad,
) {
	t.Helper()

	casos := []struct {
		nombre string
		cambio func(*plazos.FestivosNacionales)
		motivo string
	}{
		{"fila-sin-source", func(f *plazos.FestivosNacionales) { f.Festivos[0].Source = "" }, "no lleva source"},
		{
			"source-de-otro-año", func(f *plazos.FestivosNacionales) { f.Festivos[0].Source = "BOE-A-2026-21254" },
			"no es la resolución de 2025",
		},
		{"comunidad-sin-festivos", func(f *plazos.FestivosNacionales) {
			for indice, festivo := range f.Festivos {
				if strings.HasPrefix(festivo.Fecha, "2026-") {
					f.Festivos[indice].Comunidades = slices.DeleteFunc(slices.Clone(festivo.Comunidades),
						func(codigo string) bool { return codigo == "18" })
				}
			}
		}, "Ceuta (18) no tiene festivos de toda ella en 2026"},
		{"sin-un-año", func(f *plazos.FestivosNacionales) { f.Anios = f.Anios[:2] }, "los años son"},
		{"otra-fuente", func(f *plazos.FestivosNacionales) { f.Fuente = "bocm.fiestas-locales" }, "la fuente es"},
	}

	for _, caso := range casos {
		cambiado := plazos.FestivosNacionales{
			Fuente:   fichero.Fuente,
			Anios:    slices.Clone(fichero.Anios),
			Festivos: slices.Clone(fichero.Festivos),
		}
		caso.cambio(&cambiado)

		contenido, err := plazos.EscribirFestivosNacionales(cambiado, cabeceraDelFichero)
		require.NoError(t, err)

		require.ErrorContains(t, comprobarFichero(contenido, manifiesto, comunidades), caso.motivo, caso.nombre)
	}

	assert.ErrorContains(t, comprobarFichero([]byte("fuente: [\n"), manifiesto, comunidades), "no se lee en su forma")
}

// TestCruceConLasComunidades fija el cruce de cada nombre de la cabecera de las
// tres resoluciones grabadas con su comunidad de data/territorio/, sin ninguna
// lista de alias (research D14, S2; FR-022): la forma compacta igual; si
// ninguna, las mismas palabras con las abreviadas —«Com.», «C.»— por la que
// empieza como ellas; si ninguna, el nombre que lleva delante del otro su
// título unido por «de» —«Ciudad de Ceuta», «Principado de Asturias»—. Un
// nombre que no casa con ninguna, o con más de una, no se atribuye.
func TestCruceConLasComunidades(t *testing.T) {
	t.Parallel()

	comunidades := registroDelRepositorio(t).Comunidades()

	t.Run("cabeceras-grabadas", func(t *testing.T) {
		t.Parallel()

		esperadas := map[string]string{
			"Andalucía": "01", "Aragón": "02", "Asturias": "03", "Illes Balears": "04", "Canarias": "05",
			"Cantabria": "06", "Castilla y León": "07", "Castilla-La Mancha": "08", "Cataluña": "09",
			"Comunitat Valenciana": "10", "Extremadura": "11", "Galicia": "12", "Com. Madrid": "13",
			"Com. de Madrid": "13", "Región Murcia": "14", "C. Foral Navarra": "15", "País Vasco": "16",
			"La Rioja": "17", "Ciudad de Ceuta": "18", "Ciudad de Melilla": "19",
		}

		for _, anio := range aniosDeLasResoluciones {
			for _, columna := range resolucionGrabada(t, anio).Cabecera {
				codigo, conocida := esperadas[columna.Nombre]
				require.True(t, conocida, "%d: «%s» es un nombre de la cabecera que el test no conoce", anio, columna.Nombre)

				comunidad, err := comunidadDeLaColumna(columna.Nombre, comunidades)
				require.NoError(t, err)
				assert.Equal(t, codigo, comunidad.Codigo, "%d: «%s»", anio, columna.Nombre)
			}
		}
	})

	t.Run("sin-una-sola-comunidad", func(t *testing.T) {
		t.Parallel()

		inventadas := func(nombres ...string) []territorio.Comunidad {
			de := make([]territorio.Comunidad, 0, len(nombres))
			for indice, nombre := range nombres {
				de = append(de, territorio.Comunidad{Codigo: fmt.Sprintf("%02d", indice+1), Nombre: nombre})
			}

			return de
		}

		casos := []struct {
			nombre, columna string
			comunidades     []territorio.Comunidad
			motivo          string
		}{
			{"ninguna", "Castilla", comunidades, "no casa con ninguna"},
			{"vacia", "", comunidades, "no casa con ninguna"},
			{"dos-por-la-forma-compacta", "Rioja", inventadas("La Rioja", "Rioja"), "con más de una"},
			{"dos-por-la-abreviatura", "Com. Norte", inventadas("Comunidad Norte", "Comarca Norte"), "con más de una"},
			{"dos-por-el-titulo", "Norte", inventadas("Ciudad de Norte", "Villa de Norte"), "con más de una"},
			{"titulo-sin-de", "Norte", inventadas("Ciudad Norte"), "no casa con ninguna"},
		}

		for _, caso := range casos {
			_, err := comunidadDeLaColumna(caso.columna, caso.comunidades)
			require.ErrorContains(t, err, caso.motivo, caso.nombre)
			assert.ErrorContains(t, err, "«"+caso.columna+"»", caso.nombre)
		}
	})
}

// TestDerivacionQueNoSale fija lo que hace fallar la derivación en lugar de dar
// un fichero incompleto o atribuido a quien no es: un manifiesto de otra fuente
// o con un año dos veces, una resolución que no está grabada o cuyo anexo es de
// otro año que el del manifiesto, y, sobre resoluciones ya leídas, un nombre de
// la cabecera que no casa, dos que casan con la misma comunidad y una comunidad
// sin festivos de toda ella (contracts/datos-de-plazos.md §1.1; research D14,
// S2). Fija también cómo se juntan las filas: una por fecha, ámbito y parte.
func TestDerivacionQueNoSale(t *testing.T) {
	t.Parallel()

	registro := registroDelRepositorio(t)
	grabado := manifiestoDelPaquete(t)

	t.Run("manifiesto", func(t *testing.T) {
		t.Parallel()

		casos := []struct {
			nombre     string
			manifiesto Manifiesto
			motivo     string
		}{
			{"otra-fuente", Manifiesto{Fuente: "bocm.fiestas-locales"}, "bocm.fiestas-locales"},
			{"año-repetido", Manifiesto{Fuente: NombreDeLaFuente, Disposiciones: []Disposicion{
				grabado.Disposiciones[0], {ID: "BOE-A-2025-21667", Anio: 2025, URL: grabado.Disposiciones[1].URL},
			}}, "dos veces"},
			{"anexo-de-otro-año", Manifiesto{Fuente: NombreDeLaFuente, Disposiciones: []Disposicion{
				{ID: grabado.Disposiciones[0].ID, Anio: 2026, URL: grabado.Disposiciones[0].URL},
			}}, "2025"},
			{"sin-grabar", Manifiesto{Fuente: NombreDeLaFuente, Disposiciones: []Disposicion{
				{ID: "BOE-A-2027-1", Anio: 2028, URL: "https://www.boe.es/diario_boe/txt.php?id=BOE-A-2027-1"},
			}}, "BOE-A-2027-1"},
		}

		for _, caso := range casos {
			_, err := Derivar(t.Context(), clienteDeLoGrabado(t), caso.manifiesto, registro)
			assert.ErrorContains(t, err, caso.motivo, caso.nombre)
		}
	})

	t.Run("resolucion-leida", probarFestivosDe)
}

// probarFestivosDe deriva las filas de resoluciones sintéticas ya leídas, con
// dos comunidades inventadas.
func probarFestivosDe(t *testing.T) {
	t.Parallel()

	nacional, autonomico := plazos.AmbitoNacional, plazos.AmbitoAutonomico
	comunidades := []territorio.Comunidad{{Codigo: "02", Nombre: "Aragón"}, {Codigo: "05", Nombre: "Canarias"}}
	resolucion := Resolucion{
		Anio:     2026,
		Cabecera: []Columna{{Nombre: "Canarias", Notas: []int{1}}, {Nombre: "Aragón"}},
		Filas: []Fila{
			{Fecha: "2026-01-01", Ambitos: []plazos.Ambito{nacional, nacional}},
			{Fecha: "2026-04-23", Ambitos: []plazos.Ambito{"", autonomico}},
			{Fecha: "2026-04-23", Ambitos: []plazos.Ambito{autonomico, ""}},
		},
		Partes: []FiestaDeUnaParte{{Fecha: "2026-09-24", Parte: "El Hierro", Columna: 0, Nota: 1}},
	}

	festivos, err := festivosDe(resolucion, "BOE-A-2025-21667", comunidades)
	require.NoError(t, err)
	assert.Equal(t, []plazos.FestivoNacional{
		{Fecha: "2026-01-01", Ambito: nacional, Comunidades: []string{"02", "05"}, Source: "BOE-A-2025-21667"},
		{Fecha: "2026-04-23", Ambito: autonomico, Comunidades: []string{"02", "05"}, Source: "BOE-A-2025-21667"},
		{
			Fecha: "2026-09-24", Ambito: autonomico, Comunidades: []string{"05"}, Parte: "El Hierro",
			Source: "BOE-A-2025-21667",
		},
	}, festivos)

	casos := []struct {
		nombre string
		cambio func(*Resolucion)
		motivo string
	}{
		{"nombre-que-no-casa", func(r *Resolucion) { r.Cabecera[1].Nombre = "Atlántida" }, "«Atlántida»"},
		{"dos-con-la-misma", func(r *Resolucion) { r.Cabecera[1].Nombre = "Canarias" }, "las dos con Canarias (05)"},
		{"comunidad-sin-festivos", func(r *Resolucion) {
			r.Filas = []Fila{{Fecha: "2026-04-23", Ambitos: []plazos.Ambito{"", autonomico}}}
		}, "Canarias (05)"},
	}

	for _, caso := range casos {
		cambiada := resolucion
		cambiada.Cabecera = slices.Clone(resolucion.Cabecera)
		caso.cambio(&cambiada)

		_, err := festivosDe(cambiada, "BOE-A-2025-21667", comunidades)
		assert.ErrorContains(t, err, caso.motivo, caso.nombre)
	}
}
