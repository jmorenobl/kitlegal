package bocm

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"maps"
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

// actualizarFestivos es la bandera con la que TestFestivosLocalesDerivados
// escribe data/festivos/locales/13.yaml desde lo grabado antes de compararlo,
// como -actualizar-golden (contracts/datos-de-plazos.md §1.3). Solo la usa la
// tarea [datos] que deriva el fichero, y lo escrito lo revisa una persona:
// ningún fichero de data/festivos/ se escribe a mano (FR-020).
var actualizarFestivos = flag.Bool("actualizar-festivos", false,
	"escribe data/festivos/locales/"+ComunidadDelBoletin+".yaml desde lo grabado antes de compararlo")

// Dónde están, relativos al directorio del paquete, el fichero que se deriva y
// el nacional con el que el calendario lo carga.
const (
	rutaDelFichero      = "../../../data/festivos/locales/" + ComunidadDelBoletin + ".yaml"
	rutaDeLosNacionales = "../../../data/festivos/nacionales-y-autonomicos.yaml"
)

// cabeceraDelFichero es lo que nombra su línea de fichero generado
// (contracts/datos-de-plazos.md §1.2).
var cabeceraDelFichero = plazos.Cabecera{
	Desde:   "evidencias/h9/" + NombreDeLaFuente + "/",
	Test:    "TestFestivosLocalesDerivados",
	Paquete: "internal/source/bocm",
}

// aniosEsperados son los dos años del fichero con sus disposiciones en el
// orden en que se publicaron (FR-020, FR-021; contracts/datos-de-plazos.md
// §1.2).
var aniosEsperados = []plazos.AnioLocal{
	{Anio: 2025, Disposiciones: []string{"BOCM-20241213-23", "BOCM-20241227-32", "BOCM-20250203-16", "BOCM-20250217-24"}},
	{Anio: 2026, Disposiciones: []string{"BOCM-20251212-34", "BOCM-20251229-19", "BOCM-20260108-18"}},
}

// nacionalesDelRepositorio es data/festivos/nacionales-y-autonomicos.yaml, el
// que T007 derivó, con el que el calendario carga el fichero local.
func nacionalesDelRepositorio(t *testing.T) []byte {
	t.Helper()

	contenido, err := os.ReadFile(rutaDeLosNacionales)
	require.NoError(t, err)

	return contenido
}

// festivosDe son las filas de festivos de un municipio en un año, en el orden
// del fichero.
func festivosDe(fichero plazos.FestivosLocales, municipio string, anio int) []plazos.FestivoLocal {
	var festivos []plazos.FestivoLocal

	for _, festivo := range fichero.Festivos {
		if festivo.Municipio == municipio && strings.HasPrefix(festivo.Fecha, fmt.Sprintf("%d-", anio)) {
			festivos = append(festivos, festivo)
		}
	}

	return festivos
}

// faltasDe son las faltas de un municipio en un año.
func faltasDe(fichero plazos.FestivosLocales, municipio string, anio int) []plazos.FaltaLocal {
	var faltas []plazos.FaltaLocal

	for _, falta := range fichero.Faltas {
		if falta.Municipio == municipio && falta.Anio == anio {
			faltas = append(faltas, falta)
		}
	}

	return faltas
}

// clonar copia el fichero derivado hasta sus listas, para cambiarlo sin tocar
// el original.
func clonar(fichero plazos.FestivosLocales) plazos.FestivosLocales {
	copia := plazos.FestivosLocales{
		Fuente:    fichero.Fuente,
		Comunidad: fichero.Comunidad,
		Anios:     make([]plazos.AnioLocal, 0, len(fichero.Anios)),
		Festivos:  slices.Clone(fichero.Festivos),
		Faltas:    slices.Clone(fichero.Faltas),
	}

	for _, anio := range fichero.Anios {
		copia.Anios = append(copia.Anios, plazos.AnioLocal{Anio: anio.Anio, Disposiciones: slices.Clone(anio.Disposiciones)})
	}

	return copia
}

// comprobacion acumula lo que FR-083 y SC-006 piden del fichero escrito.
type comprobacion struct {
	// codigos son los de los 179 municipios de la comunidad.
	codigos map[string]bool
	// disposiciones lleva de cada año del fichero a sus disposiciones.
	disposiciones map[int][]string
	// conFestivos y conFalta llevan de cada año y municipio a si el fichero le
	// da festivos o una falta.
	conFestivos, conFalta map[string]bool
	defectos              []error
}

// defecto anota uno.
func (c *comprobacion) defecto(formato string, argumentos ...any) {
	c.defectos = append(c.defectos, fmt.Errorf(formato, argumentos...))
}

// source exige que la fila lleve source y que sea una disposición de su año.
func (c *comprobacion) source(source string, anio int, fila string) {
	switch {
	case source == "":
		c.defecto("%s no lleva source", fila)
	case !slices.Contains(c.disposiciones[anio], source):
		c.defecto("%s lleva el source %s, que no es una disposición de %d", fila, source, anio)
	}
}

// municipio exige que el código sea el de un municipio de la comunidad.
func (c *comprobacion) municipio(codigo, fila string) {
	if !c.codigos[codigo] {
		c.defecto("%s lleva el municipio %q, que no es un municipio de la comunidad", fila, codigo)
	}
}

// clave es la de un año y un municipio.
func clave(anio int, municipio string) string {
	return fmt.Sprintf("%d/%s", anio, municipio)
}

// festivo comprueba una fila de festivos.
func (c *comprobacion) festivo(festivo plazos.FestivoLocal) {
	anio, _, _ := strings.Cut(festivo.Fecha, "-")
	fila := fmt.Sprintf("la fila del %s de %s", festivo.Fecha, festivo.Municipio)

	var numero int
	if _, err := fmt.Sscan(anio, &numero); err != nil {
		c.defecto("%s no lleva una fecha AAAA-MM-DD", fila)
	}

	c.source(festivo.Source, numero, fila)
	c.municipio(festivo.Municipio, fila)
	c.conFestivos[clave(numero, festivo.Municipio)] = true
}

// falta comprueba una falta: la de un municipio, con uno de sus tres motivos y
// sin línea; la de una línea, sin municipio, con su nombre y sin-municipio.
func (c *comprobacion) falta(falta plazos.FaltaLocal) {
	fila := fmt.Sprintf("la falta de %d de %q", falta.Anio, falta.Municipio+falta.Linea)

	c.source(falta.Source, falta.Anio, fila)

	if !slices.Contains(plazos.MotivosDeFalta(), falta.Motivo) {
		c.defecto("%s lleva el motivo %q, que no es un motivo de falta", fila, falta.Motivo)
	}

	if falta.Municipio == "" {
		if falta.Motivo != plazos.FaltaSinMunicipio {
			c.defecto("%s no lleva municipio y su motivo no es %s", fila, plazos.FaltaSinMunicipio)
		}

		if falta.Linea == "" {
			c.defecto("%s es de una línea sin nombre", fila)
		}

		return
	}

	c.municipio(falta.Municipio, fila)

	if falta.Motivo == plazos.FaltaSinMunicipio {
		c.defecto("%s lleva municipio y el motivo %s", fila, plazos.FaltaSinMunicipio)
	}

	if falta.Linea != "" {
		c.defecto("%s lleva municipio y línea", fila)
	}

	c.conFalta[clave(falta.Anio, falta.Municipio)] = true
}

// cobertura exige que cada municipio tenga en cada año festivos o una falta, y
// nunca las dos cosas (FR-022; SC-006).
func (c *comprobacion) cobertura(municipios []territorio.MunicipioConFormas) {
	for _, anio := range slices.Sorted(maps.Keys(c.disposiciones)) {
		for _, municipio := range municipios {
			codigo := municipio.Codigo.String()

			switch festivos, falta := c.conFestivos[clave(anio, codigo)], c.conFalta[clave(anio, codigo)]; {
			case !festivos && !falta:
				c.defecto("%s (%s) no tiene en %d ni festivos ni falta", municipio.Nombre, codigo, anio)
			case festivos && falta:
				c.defecto("%s (%s) tiene en %d festivos y una falta", municipio.Nombre, codigo, anio)
			}
		}
	}
}

// aniosIguales dice si las dos listas de años son la misma, disposición a
// disposición y en el mismo orden.
func aniosIguales(a, b []plazos.AnioLocal) bool {
	return slices.EqualFunc(a, b, func(x, y plazos.AnioLocal) bool {
		return x.Anio == y.Anio && slices.Equal(x.Disposiciones, y.Disposiciones)
	})
}

// comprobarFichero es lo que FR-083 y SC-006 piden del fichero escrito, además
// de ser igual a lo que sale de lo grabado: que se lee en su forma, que es de
// la fuente y de la comunidad del boletín y de los años del contrato con sus
// disposiciones en su orden, que cada fila lleva su source y es una
// disposición de su año, que cada municipio es de la comunidad, que cada uno
// de los 179 tiene en cada año festivos o una falta y nunca las dos cosas, que
// cada falta tiene la forma de su motivo y que el calendario lo carga.
// Devuelve todos los defectos juntos.
func comprobarFichero(contenido, nacionales []byte, municipios []territorio.MunicipioConFormas) error {
	var fichero plazos.FestivosLocales

	lector := yaml.NewDecoder(bytes.NewReader(contenido))
	lector.KnownFields(true)

	if err := lector.Decode(&fichero); err != nil {
		return fmt.Errorf("el fichero no se lee en su forma: %w", err)
	}

	c := &comprobacion{
		codigos:       make(map[string]bool, len(municipios)),
		disposiciones: map[int][]string{},
		conFestivos:   map[string]bool{},
		conFalta:      map[string]bool{},
	}

	for _, municipio := range municipios {
		c.codigos[municipio.Codigo.String()] = true
	}

	if fichero.Fuente != NombreDeLaFuente {
		c.defecto("la fuente es %q y no %q", fichero.Fuente, NombreDeLaFuente)
	}

	if fichero.Comunidad != ComunidadDelBoletin {
		c.defecto("la comunidad es %q y no %q", fichero.Comunidad, ComunidadDelBoletin)
	}

	if !aniosIguales(fichero.Anios, aniosEsperados) {
		c.defecto("los años son %v y el contrato da %v", fichero.Anios, aniosEsperados)
	}

	for _, anio := range fichero.Anios {
		c.disposiciones[anio.Anio] = anio.Disposiciones
	}

	for _, festivo := range fichero.Festivos {
		c.festivo(festivo)
	}

	for _, falta := range fichero.Faltas {
		c.falta(falta)
	}

	c.cobertura(municipios)

	if _, err := plazos.CargarCalendario(nacionales, map[string][]byte{ComunidadDelBoletin: contenido}); err != nil {
		c.defectos = append(c.defectos, err)
	}

	return errors.Join(c.defectos...)
}

// TestFestivosLocalesDerivados deriva data/festivos/locales/13.yaml desde las
// siete disposiciones grabadas y el territorio real (FR-020 a FR-024, FR-083;
// SC-006; research D13) y falla si el fichero del árbol no es igual byte a
// byte a lo derivado, si una fila no lleva su source, si uno de los 179
// municipios no tiene en 2025 o en 2026 festivos o una falta, o tiene las dos
// cosas, o si una línea de lo grabado no da ni festivos de un municipio ni una
// falta. Afirma los escenarios 2 y 3 de US3 y que el orden lo da la fecha de
// publicación y no el manifiesto. Con -actualizar-festivos escribe antes el
// fichero; sin la bandera no escribe nada. Un subtest demuestra que el
// comprobador no pasa en vacío.
func TestFestivosLocalesDerivados(t *testing.T) {
	t.Parallel()

	registro := registroDelRepositorio(t)
	manifiesto := manifiestoDelPaquete(t)
	municipios := registro.MunicipiosDe(ComunidadDelBoletin)
	require.Len(t, municipios, numeroDeMunicipios)

	nacionales := nacionalesDelRepositorio(t)

	fichero, err := Derivar(t.Context(), clienteDeLoGrabado(t), manifiesto, registro)
	require.NoError(t, err)

	derivado, err := plazos.EscribirFestivosLocales(fichero, cabeceraDelFichero)
	require.NoError(t, err)

	if *actualizarFestivos {
		require.NoError(t, os.MkdirAll(filepath.Dir(rutaDelFichero), 0o750))
		require.NoError(t, os.WriteFile(rutaDelFichero, derivado, 0o600))
	}

	t.Run("data-festivos", func(t *testing.T) {
		t.Parallel()

		escrito, err := os.ReadFile(rutaDelFichero)
		require.NoError(t, err, "el fichero lo escribe este test con -actualizar-festivos")

		require.NoError(t, comprobarFichero(escrito, nacionales, municipios))
		assert.Equal(t, string(derivado), string(escrito),
			"%s no es lo que sale de lo grabado: se regenera con -actualizar-festivos, nunca a mano", rutaDelFichero)
	})

	t.Run("el-comprobador-no-pasa-en-vacio", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, comprobarFichero(derivado, nacionales, municipios))
		probarElComprobador(t, fichero, nacionales, municipios)
	})

	t.Run("majadahonda-2025", func(t *testing.T) {
		t.Parallel()

		const majadahonda, correctora = "28080", "BOCM-20250203-16"

		festivos := festivosDe(fichero, majadahonda, 2025)
		assert.Equal(t, []plazos.FestivoLocal{
			{Fecha: "2025-09-15", Municipio: majadahonda, Source: correctora},
			{Fecha: "2025-11-25", Municipio: majadahonda, Source: correctora},
		}, festivos, "US3.2: la fiesta de septiembre es el 15 de la corrección, no el 14 de la principal")
		assert.NotContains(t, festivos, plazos.FestivoLocal{Fecha: "2025-09-14", Municipio: majadahonda, Source: "BOCM-20241213-23"})
		assert.Empty(t, faltasDe(fichero, majadahonda, 2025))
	})

	t.Run("brunete-2026", func(t *testing.T) {
		t.Parallel()

		const brunete, correctora = "28026", "BOCM-20251229-19"

		assert.Equal(t, []plazos.FestivoLocal{
			{Fecha: "2026-05-15", Municipio: brunete, Source: correctora},
			{Fecha: "2026-09-14", Municipio: brunete, Source: correctora},
		}, festivosDe(fichero, brunete, 2026), "US3.3: «no comunicado» en la principal, los de la corrección que los da")
		assert.Empty(t, faltasDe(fichero, brunete, 2026))
	})

	t.Run("cada-linea-de-lo-grabado", func(t *testing.T) {
		t.Parallel()

		c := nuevoCruce(municipios)
		sinMunicipio := 0

		for _, disposicion := range manifiesto.Disposiciones {
			for _, linea := range resolucionGrabada(t, disposicion.ID).Lineas {
				codigo, casa := c.municipioDe(linea.Nombre)
				if !casa {
					sinMunicipio++

					assert.Contains(t, fichero.Faltas, plazos.FaltaLocal{
						Anio: disposicion.Anio, Linea: linea.Nombre, Motivo: plazos.FaltaSinMunicipio, Source: disposicion.ID,
					}, "%s, «%s»: la línea no casa con un municipio y no deja falta", disposicion.ID, linea.Nombre)

					continue
				}

				conFestivos, conFalta := festivosDe(fichero, codigo, disposicion.Anio), faltasDe(fichero, codigo, disposicion.Anio)
				assert.True(t, len(conFestivos) > 0 || len(conFalta) > 0,
					"%s, «%s» (%s): la línea no da ni festivos ni falta", disposicion.ID, linea.Nombre, codigo)
			}
		}

		deLineas := 0

		for _, falta := range fichero.Faltas {
			if falta.Motivo == plazos.FaltaSinMunicipio {
				deLineas++
			}
		}

		assert.Equal(t, sinMunicipio, deLineas, "cada falta sin-municipio es una línea de lo grabado")
	})

	t.Run("el-orden-lo-da-la-publicacion", func(t *testing.T) {
		t.Parallel()

		desordenado := Manifiesto{Fuente: manifiesto.Fuente, Disposiciones: slices.Clone(manifiesto.Disposiciones)}
		slices.Reverse(desordenado.Disposiciones)

		otro, err := Derivar(t.Context(), clienteDeLoGrabado(t), desordenado, registro)
		require.NoError(t, err)

		escrito, err := plazos.EscribirFestivosLocales(otro, cabeceraDelFichero)
		require.NoError(t, err)
		assert.Equal(t, string(derivado), string(escrito), "el manifiesto al revés da el mismo fichero")
	})
}

// probarElComprobador cambia el fichero derivado de cada forma que FR-083 y
// SC-006 prohíben y comprueba que comprobarFichero lo rechaza nombrándolo.
func probarElComprobador(t *testing.T, fichero plazos.FestivosLocales, nacionales []byte,
	municipios []territorio.MunicipioConFormas,
) {
	t.Helper()

	deMunicipio := slices.IndexFunc(fichero.Faltas, func(f plazos.FaltaLocal) bool { return f.Municipio != "" })
	deLinea := slices.IndexFunc(fichero.Faltas, func(f plazos.FaltaLocal) bool { return f.Municipio == "" })
	require.GreaterOrEqual(t, deMunicipio, 0, "lo derivado tiene alguna falta de municipio")
	require.GreaterOrEqual(t, deLinea, 0, "lo derivado tiene alguna falta de línea")

	primero := fichero.Festivos[0].Municipio
	principal, correctora := aniosEsperados[0].Disposiciones[0], aniosEsperados[1].Disposiciones[0]

	casos := []struct {
		nombre string
		cambio func(*plazos.FestivosLocales)
		motivo string
	}{
		{"fila-sin-source", func(f *plazos.FestivosLocales) { f.Festivos[0].Source = "" }, "no lleva source"},
		{
			"fila-con-source-de-otro-año", func(f *plazos.FestivosLocales) { f.Festivos[0].Source = correctora },
			"no es una disposición de 2025",
		},
		{"falta-sin-source", func(f *plazos.FestivosLocales) { f.Faltas[deMunicipio].Source = "" }, "no lleva source"},
		{
			"municipio-de-otra-comunidad", func(f *plazos.FestivosLocales) { f.Festivos[0].Municipio = "15024" },
			"no es un municipio de la comunidad",
		},
		{"municipio-sin-nada", func(f *plazos.FestivosLocales) {
			f.Festivos = slices.DeleteFunc(f.Festivos, func(festivo plazos.FestivoLocal) bool {
				return festivo.Municipio == primero && strings.HasPrefix(festivo.Fecha, "2025-")
			})
		}, "no tiene en 2025 ni festivos ni falta"},
		{"municipio-con-las-dos-cosas", func(f *plazos.FestivosLocales) {
			f.Faltas = append(f.Faltas, plazos.FaltaLocal{
				Anio: 2025, Municipio: primero, Motivo: plazos.FaltaNoComunicado, Source: principal,
			})
		}, "tiene en 2025 festivos y una falta"},
		{
			"motivo-fuera-del-vocabulario", func(f *plazos.FestivosLocales) { f.Faltas[deMunicipio].Motivo = "otro" },
			"no es un motivo de falta",
		},
		{
			"falta-de-municipio-con-linea", func(f *plazos.FestivosLocales) { f.Faltas[deMunicipio].Linea = "Villa" },
			"lleva municipio y línea",
		},
		{"falta-de-municipio-sin-municipio", func(f *plazos.FestivosLocales) {
			f.Faltas[deLinea].Motivo = plazos.FaltaNoComunicado
		}, "no lleva municipio y su motivo no es sin-municipio"},
		{"linea-sin-nombre", func(f *plazos.FestivosLocales) { f.Faltas[deLinea].Linea = "" }, "una línea sin nombre"},
		{
			"fecha-que-no-es-fecha", func(f *plazos.FestivosLocales) { f.Festivos[0].Fecha = "2025-13-01" },
			"no es una fecha AAAA-MM-DD",
		},
		{"sin-un-año", func(f *plazos.FestivosLocales) { f.Anios = f.Anios[:1] }, "los años son"},
		{"otra-fuente", func(f *plazos.FestivosLocales) { f.Fuente = "boe.fiestas-laborales" }, "la fuente es"},
		{"otra-comunidad", func(f *plazos.FestivosLocales) { f.Comunidad = "12" }, "la comunidad es"},
	}

	for _, caso := range casos {
		cambiado := clonar(fichero)
		caso.cambio(&cambiado)

		contenido, err := plazos.EscribirFestivosLocales(cambiado, cabeceraDelFichero)
		require.NoError(t, err)

		require.ErrorContains(t, comprobarFichero(contenido, nacionales, municipios), caso.motivo, caso.nombre)
	}

	assert.ErrorContains(t, comprobarFichero([]byte("fuente: [\n"), nacionales, municipios), "no se lee en su forma")
}

// resolucionDePrueba es una disposición con ese identificador y esas líneas,
// leída como se lee una grabada, con su fecha de publicación.
func resolucionDePrueba(t *testing.T, id string, anio int, lineas ...string) publicada {
	t.Helper()

	resolucion, err := LeerDisposicion(documentoDePrueba(id, fmt.Sprintf("Fiestas para el año %d", anio), lineas...))
	require.NoError(t, err)

	publicacion, err := fechaDePublicacion(id)
	require.NoError(t, err)

	return publicada{resolucion: resolucion, publicacion: publicacion}
}

// TestDerivacionQueNoSale fija lo que hace fallar la derivación en lugar de dar
// un fichero incompleto o atribuido a quien no es —un manifiesto de otra
// fuente, con una disposición dos veces o con un identificador sin forma, una
// disposición que no está grabada, cuyo documento es otro o que es de otro año
// que el del manifiesto, un territorio sin la comunidad del boletín y una fecha
// que no existe en el año— y, con disposiciones sintéticas, las transiciones de
// FR-021 (data-model §2): cada línea de una posterior sustituye entera la del
// municipio, con su source; el municipio sin línea queda sin-linea con la
// última; y el orden lo da la fecha de publicación.
func TestDerivacionQueNoSale(t *testing.T) {
	t.Parallel()

	registro := registroDelRepositorio(t)
	grabado := manifiestoDelPaquete(t)

	t.Run("manifiesto", func(t *testing.T) {
		t.Parallel()

		primera := grabado.Disposiciones[0]

		casos := []struct {
			nombre     string
			manifiesto Manifiesto
			motivo     string
		}{
			{"otra-fuente", Manifiesto{Fuente: "boe.fiestas-laborales"}, "boe.fiestas-laborales"},
			{"repetida", Manifiesto{Fuente: NombreDeLaFuente, Disposiciones: []Disposicion{primera, primera}}, "dos veces"},
			{"identificador-sin-forma", Manifiesto{Fuente: NombreDeLaFuente, Disposiciones: []Disposicion{
				{ID: "BOCM-2024-23", Anio: primera.Anio, URL: primera.URL},
			}}, "no tiene la forma"},
			{"identificador-con-letras", Manifiesto{Fuente: NombreDeLaFuente, Disposiciones: []Disposicion{
				{ID: "BOCM-2024121a-23", Anio: primera.Anio, URL: primera.URL},
			}}, "no tiene la forma"},
			{"fecha-que-no-es-un-dia", Manifiesto{Fuente: NombreDeLaFuente, Disposiciones: []Disposicion{
				{ID: "BOCM-20241399-23", Anio: primera.Anio, URL: primera.URL},
			}}, "no es un día del calendario"},
			{"otro-documento", Manifiesto{Fuente: NombreDeLaFuente, Disposiciones: []Disposicion{
				{ID: "BOCM-20241213-24", Anio: primera.Anio, URL: primera.URL},
			}}, "es la disposición " + primera.ID},
			{"otro-año", Manifiesto{Fuente: NombreDeLaFuente, Disposiciones: []Disposicion{
				{ID: primera.ID, Anio: primera.Anio + 1, URL: primera.URL},
			}}, fmt.Sprint(primera.Anio)},
			{"sin-grabar", Manifiesto{Fuente: NombreDeLaFuente, Disposiciones: []Disposicion{
				{ID: "BOCM-20270101-1", Anio: 2027, URL: "https://www.bocm.es/boletin/CM_Orden_BOCM/2027/01/01/BOCM-20270101-1.xml"},
			}}, "BOCM-20270101-1"},
		}

		for _, caso := range casos {
			_, err := Derivar(t.Context(), clienteDeLoGrabado(t), caso.manifiesto, registro)
			assert.ErrorContains(t, err, caso.motivo, caso.nombre)
		}
	})

	t.Run("territorio-sin-la-comunidad", func(t *testing.T) {
		t.Parallel()

		fuentes := territorio.Fuentes{
			Municipios: []byte("fecha: \"2026-02-04\"\nsource: ine.municipios\nmunicipios:\n" +
				"  \"15024\": {dc: \"3\", nombre: \"Cerceda\", provincia: \"15\", comunidad: \"12\"}\n"),
			DIR3: []byte("fecha: \"2026-09-21\"\nsource: mpt.rel\ncorrespondencia: {}\n"),
			Estado: []byte("fecha: \"2026-09-20\"\nsource: data/territorio/estado.yaml\nboletin:\n  codigo: \"BOE\"\n" +
				"  nombre: \"Boletín Oficial del Estado\"\n  url: \"https://www.boe.es/\"\n"),
			Comunidades: map[string][]byte{"12": []byte("fecha: \"2026-09-20\"\nsource: ine.codigos-territoriales\n" +
				"codigo: \"12\"\nnombre: \"Galicia\"\nregimen: comun\nprovincias:\n  \"15\": \"A Coruña\"\n")},
		}

		sinMadrid, err := territorio.Cargar(fuentes)
		require.NoError(t, err)

		_, err = Derivar(t.Context(), clienteDeLoGrabado(t), grabado, sinMadrid)
		assert.ErrorContains(t, err, "ningún municipio de la comunidad "+ComunidadDelBoletin)
	})

	t.Run("transiciones", func(t *testing.T) {
		t.Parallel()

		c := nuevoCruce([]territorio.MunicipioConFormas{
			municipioSintetico(t, "28001", "villa"),
			municipioSintetico(t, "28002", "aldea"),
			municipioSintetico(t, "28003", "lugar"),
			municipioSintetico(t, "28004", "campo"),
		})

		principal := resolucionDePrueba(t, "BOCM-20250101-1", 2025,
			"Villa: no comunicado por el ayuntamiento.",
			"Aldea: 1 de mayo (Arriba) y 2 de mayo (Abajo).",
			"Lugar: 2 de febrero y 1 de enero.",
			"Barrio: 3 de marzo.")
		correctora := resolucionDePrueba(t, "BOCM-20250201-2", 2025,
			"Villa: 2 de mayo y 3 de junio.",
			"Aldea: 4 de julio.")

		conLaPrincipalSola, err := derivarAnio(2025, []publicada{principal}, c)
		require.NoError(t, err)
		assert.Equal(t, []string{"BOCM-20250101-1"}, conLaPrincipalSola.disposiciones)
		assert.Equal(t, []plazos.FestivoLocal{
			{Fecha: "2025-01-01", Municipio: "28003", Source: "BOCM-20250101-1"},
			{Fecha: "2025-02-02", Municipio: "28003", Source: "BOCM-20250101-1"},
		}, conLaPrincipalSola.festivos)
		assert.ElementsMatch(t, []plazos.FaltaLocal{
			{Anio: 2025, Municipio: "28001", Motivo: plazos.FaltaNoComunicado, Source: "BOCM-20250101-1"},
			{Anio: 2025, Municipio: "28002", Motivo: plazos.FaltaLineaSinLeer, Source: "BOCM-20250101-1"},
			{Anio: 2025, Municipio: "28004", Motivo: plazos.FaltaSinLinea, Source: "BOCM-20250101-1"},
			{Anio: 2025, Linea: "Barrio", Motivo: plazos.FaltaSinMunicipio, Source: "BOCM-20250101-1"},
		}, conLaPrincipalSola.faltas)

		for nombre, orden := range map[string][]publicada{
			"en-orden": {principal, correctora}, "al-reves": {correctora, principal},
		} {
			conLaCorreccion, err := derivarAnio(2025, orden, c)
			require.NoError(t, err, nombre)
			assert.Equal(t, []string{"BOCM-20250101-1", "BOCM-20250201-2"}, conLaCorreccion.disposiciones, nombre)
			assert.ElementsMatch(t, []plazos.FestivoLocal{
				{Fecha: "2025-01-01", Municipio: "28003", Source: "BOCM-20250101-1"},
				{Fecha: "2025-02-02", Municipio: "28003", Source: "BOCM-20250101-1"},
				{Fecha: "2025-05-02", Municipio: "28001", Source: "BOCM-20250201-2"},
				{Fecha: "2025-06-03", Municipio: "28001", Source: "BOCM-20250201-2"},
				{Fecha: "2025-07-04", Municipio: "28002", Source: "BOCM-20250201-2"},
			}, conLaCorreccion.festivos, nombre)
			assert.ElementsMatch(t, []plazos.FaltaLocal{
				{Anio: 2025, Municipio: "28004", Motivo: plazos.FaltaSinLinea, Source: "BOCM-20250201-2"},
				{Anio: 2025, Linea: "Barrio", Motivo: plazos.FaltaSinMunicipio, Source: "BOCM-20250101-1"},
			}, conLaCorreccion.faltas, nombre)
		}

		_, err = derivarAnio(2025, []publicada{resolucionDePrueba(t, "BOCM-20250101-1", 2025, "Villa: 29 de febrero.")}, c)
		require.ErrorContains(t, err, "29 de febrero")
		require.ErrorContains(t, err, "Villa")

		_, err = derivarAnio(2025, nil, c)
		assert.ErrorContains(t, err, "ninguna disposición")
	})
}
