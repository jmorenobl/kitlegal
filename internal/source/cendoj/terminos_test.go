package cendoj

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core"
)

// rutaDeSources es docs/SOURCES.md visto desde este paquete: la tabla de fuentes
// que revisa una persona y a la que este test ata lo que el adaptador declara.
const rutaDeSources = "../../../docs/SOURCES.md"

// Las celdas de la fila que tienen que coincidir con lo que el adaptador declara
// (contrato fuente-cendoj-y-grabacion §1).
const (
	columnaFuente   = "Fuente"
	columnaBase     = "Base"
	columnaTerminos = "Términos de uso"
	columnaRitmo    = "Ritmo"
	columnaFormato  = "Formato"
	columnaRevisado = "Revisado"
)

// envioDelFormulario es lo que lleva delante, en la celda «Formato», la ruta a
// la que se envía el formulario: su método.
const envioDelFormulario = "POST "

// cabeceraDeSources es la cabecera de la tabla de fuentes, celda a celda y en su
// orden.
var cabeceraDeSources = []string{
	columnaFuente, "Applet", columnaBase, "Licencia", columnaTerminos, "robots.txt", columnaRitmo, columnaFormato,
	columnaRevisado,
}

var (
	// separadorDeCelda es una celda de la fila que separa la cabecera de los
	// datos en una tabla de Markdown.
	separadorDeCelda = regexp.MustCompile(`^:?-+:?$`)
	// direccionEntreAngulos es una dirección escrita como <…>, la forma en que
	// las celdas «Base» y «Términos de uso» llevan la suya.
	direccionEntreAngulos = regexp.MustCompile(`<([^<>\s]+)>`)
	// tramoDeCodigo es un texto entre comillas invertidas, la forma en que la
	// celda «Formato» nombra la página, el formulario y sus campos de consulta.
	tramoDeCodigo = regexp.MustCompile("`([^`]+)`")
)

// loDeclarado es lo que el adaptador declara de cómo se consulta la fuente, que
// es lo que su fila de docs/SOURCES.md tiene que decir.
type loDeclarado struct {
	fuente     string
	intervalo  time.Duration
	terminos   core.Terminos
	base       string
	pagina     string
	formulario string
	campos     []string
}

// loQueDeclaraElAdaptador es lo de las constantes de terminos.go y de
// direcciones.go.
func loQueDeclaraElAdaptador() loDeclarado {
	return loDeclarado{
		fuente:     NombreDeLaFuente,
		intervalo:  IntervaloEntrePeticiones,
		terminos:   terminosDeUso,
		base:       baseDelBuscador,
		pagina:     paginaDelBuscador,
		formulario: formularioDeConsulta,
		campos:     []string{campoECLI, campoROJ, campoNumeroDeResolucion},
	}
}

// Las celdas de una fila como la de docs/SOURCES.md, con las que los subtests
// componen tablas en memoria.
const (
	celdaBase     = "<https://www.poderjudicial.es/search>"
	celdaTerminos = "<https://www.poderjudicial.es/search/indexAN.jsp>"
	celdaRitmo    = "`5s`"
	celdaFormato  = "HTML; sesión con cookie de `/search/indexAN.jsp` y formulario `POST /search/search.action` " +
		"con el campo `ECLI`, `ROJ` o `NUMERORESOLUCION` y fechas; sin autenticación"
	celdaRevisado = "2026-10-03"
)

// TestFuenteCoincideConSources ata la fila cendoj.jurisprudencia de
// docs/SOURCES.md a lo que el adaptador declara (FR-023, FR-114; contrato
// fuente-cendoj-y-grabacion §1): el ritmo de la fila es
// IntervaloEntrePeticiones; la dirección de sus términos de uso y su «Revisado»,
// los de terminosDeUso; su «Base», la base del buscador; y su «Formato» nombra
// la ruta de la página del buscador, la del formulario y, uno a uno, los campos
// de consulta, y ninguno más.
//
// El primer subtest compara la tabla real. Los demás demuestran sobre tablas en
// memoria que la comparación no pasa en vacío: cada constante cambiada y cada
// fila que no se puede leer dan su error.
func TestFuenteCoincideConSources(t *testing.T) {
	t.Parallel()

	t.Run("docs-SOURCES", func(t *testing.T) {
		t.Parallel()

		contenido, err := os.ReadFile(rutaDeSources)
		require.NoError(t, err)

		require.NoError(t, comprobarFilaDeLaFuente(string(contenido), loQueDeclaraElAdaptador()))
	})

	t.Run("la fila de las tablas en memoria coincide", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, comprobarFilaDeLaFuente(tablaDeSources(filaDeCendoj(nil)), loQueDeclaraElAdaptador()))
	})

	for _, caso := range constantesCambiadas() {
		t.Run("constante/"+caso.nombre, func(t *testing.T) {
			t.Parallel()

			declarado := loQueDeclaraElAdaptador()
			caso.cambiar(&declarado)

			err := comprobarFilaDeLaFuente(tablaDeSources(filaDeCendoj(nil)), declarado)

			for _, mensaje := range caso.mensajes {
				require.ErrorContains(t, err, mensaje)
			}
		})
	}

	for _, caso := range filasQueNoCoinciden() {
		t.Run("fila/"+caso.nombre, func(t *testing.T) {
			t.Parallel()

			err := comprobarFilaDeLaFuente(caso.tabla, loQueDeclaraElAdaptador())

			require.ErrorContains(t, err, caso.mensaje)
		})
	}
}

// constanteCambiada es una constante del adaptador que deja de decir lo que
// dice su fila, con los fragmentos del error que tiene que dar.
type constanteCambiada struct {
	nombre   string
	cambiar  func(*loDeclarado)
	mensajes []string
}

// constantesCambiadas son las del contrato §1, una a una: lo que se ve al
// cambiar cada constante de terminos.go y de direcciones.go.
func constantesCambiadas() []constanteCambiada {
	return []constanteCambiada{
		{
			nombre:   "el ritmo",
			cambiar:  func(d *loDeclarado) { d.intervalo = time.Second },
			mensajes: []string{"«Ritmo» es 5s y el adaptador declara 1s"},
		},
		{
			nombre:  "los términos de uso",
			cambiar: func(d *loDeclarado) { d.terminos.URL = "https://www.poderjudicial.es/aviso" },
			mensajes: []string{
				"«Términos de uso» es https://www.poderjudicial.es/search/indexAN.jsp y el adaptador declara " +
					"https://www.poderjudicial.es/aviso",
			},
		},
		{
			nombre: "el día de la revisión",
			cambiar: func(d *loDeclarado) {
				d.terminos.Revisados = time.Date(2026, time.October, 2, 0, 0, 0, 0, time.UTC)
			},
			mensajes: []string{"«Revisado» es 2026-10-03 y el adaptador declara 2026-10-02"},
		},
		{
			nombre:   "el día de la revisión, sin fijar",
			cambiar:  func(d *loDeclarado) { d.terminos.Revisados = time.Time{} },
			mensajes: []string{"«Revisado» es 2026-10-03 y el adaptador declara cero"},
		},
		{
			nombre:  "la base",
			cambiar: func(d *loDeclarado) { d.base = "https://www.poderjudicial.es/buscador" },
			mensajes: []string{
				"«Base» es https://www.poderjudicial.es/search y el adaptador declara " +
					"https://www.poderjudicial.es/buscador",
			},
		},
		{
			nombre:  "la página del buscador",
			cambiar: func(d *loDeclarado) { d.pagina = "https://www.poderjudicial.es/search/index.jsp" },
			mensajes: []string{
				"«Formato» da como página del buscador https://www.poderjudicial.es/search/indexAN.jsp y el " +
					"adaptador declara https://www.poderjudicial.es/search/index.jsp",
			},
		},
		{
			nombre:  "el formulario",
			cambiar: func(d *loDeclarado) { d.formulario = "https://www.poderjudicial.es/search/buscar.action" },
			mensajes: []string{
				"«Formato» da como formulario https://www.poderjudicial.es/search/search.action y el adaptador " +
					"declara https://www.poderjudicial.es/search/buscar.action",
			},
		},
		{
			nombre:  "el campo del ECLI",
			cambiar: func(d *loDeclarado) { d.campos[0] = "IDENTIFICADOR" },
			mensajes: []string{
				"«Formato» nombra el campo ECLI, que el adaptador no declara",
				"el adaptador declara el campo IDENTIFICADOR, que la celda «Formato» no nombra",
			},
		},
		{
			nombre:  "el campo del ROJ",
			cambiar: func(d *loDeclarado) { d.campos[1] = "roj" },
			mensajes: []string{
				"«Formato» nombra el campo ROJ, que el adaptador no declara",
				"el adaptador declara el campo roj, que la celda «Formato» no nombra",
			},
		},
		{
			nombre:  "el campo del número de resolución",
			cambiar: func(d *loDeclarado) { d.campos[2] = "NUMERORECURSO" },
			mensajes: []string{
				"«Formato» nombra el campo NUMERORESOLUCION, que el adaptador no declara",
				"el adaptador declara el campo NUMERORECURSO, que la celda «Formato» no nombra",
			},
		},
		{
			nombre:   "un campo de menos",
			cambiar:  func(d *loDeclarado) { d.campos = d.campos[:2] },
			mensajes: []string{"«Formato» nombra el campo NUMERORESOLUCION, que el adaptador no declara"},
		},
	}
}

// filaQueNoCoincide es una tabla en memoria cuya fila no se puede leer o no dice
// lo que el adaptador declara, con un fragmento del error que tiene que dar.
type filaQueNoCoincide struct {
	nombre  string
	tabla   string
	mensaje string
}

// filasQueNoCoinciden son las tablas que la comparación tiene que rechazar sin
// que cambie ninguna constante.
func filasQueNoCoinciden() []filaQueNoCoincide {
	con := func(columna, celda string) string {
		return tablaDeSources(filaDeCendoj(map[string]string{columna: celda}))
	}

	return []filaQueNoCoincide{
		{nombre: "sin tabla", tabla: "# Fuentes\n\nTexto sin tabla.", mensaje: "no tiene la tabla de fuentes"},
		{nombre: "sin fila", tabla: tablaDeSources(), mensaje: "ninguna fila"},
		{
			nombre:  "fila repetida",
			tabla:   tablaDeSources(filaDeCendoj(nil), filaDeCendoj(nil)),
			mensaje: "2 filas",
		},
		{
			nombre:  "celdas de menos",
			tabla:   tablaDeSources("| `" + NombreDeLaFuente + "` | `cita` |"),
			mensaje: "2 celdas",
		},
		{nombre: "ritmo ilegible", tabla: con(columnaRitmo, "cinco segundos"), mensaje: "«Ritmo» no es una duración"},
		{
			nombre:  "términos sin dirección",
			tabla:   con(columnaTerminos, "https://www.poderjudicial.es/search/indexAN.jsp"),
			mensaje: "«Términos de uso» lleva 0 direcciones",
		},
		{
			nombre:  "base sin dirección",
			tabla:   con(columnaBase, "https://www.poderjudicial.es/search"),
			mensaje: "«Base» lleva 0 direcciones",
		},
		{
			nombre:  "revisión pendiente",
			tabla:   con(columnaRevisado, "pendiente"),
			mensaje: "«Revisado» no es una fecha AAAA-MM-DD",
		},
		{
			nombre:  "revisión con otra forma",
			tabla:   con(columnaRevisado, "03/10/2026"),
			mensaje: "«Revisado» no es una fecha AAAA-MM-DD",
		},
		{
			nombre:  "revisión en la fecha cero",
			tabla:   con(columnaRevisado, "0001-01-01"),
			mensaje: "«Revisado» es la fecha cero",
		},
		{
			nombre:  "formato sin formulario",
			tabla:   con(columnaFormato, "HTML; cookie de `/search/indexAN.jsp`; campos `ECLI`, `ROJ` y `NUMERORESOLUCION`"),
			mensaje: "«Formato» nombra 0 rutas de formulario",
		},
		{
			nombre: "formato con dos formularios",
			tabla: con(columnaFormato, "HTML; cookie de `/search/indexAN.jsp`; formularios `POST /search/search.action` "+
				"y `POST /search/otro.action`; campos `ECLI`, `ROJ` y `NUMERORESOLUCION`"),
			mensaje: "«Formato» nombra 2 rutas de formulario",
		},
		{
			nombre:  "formato sin página",
			tabla:   con(columnaFormato, "HTML; formulario `POST /search/search.action`; campos `ECLI`, `ROJ` y `NUMERORESOLUCION`"),
			mensaje: "«Formato» nombra 0 rutas de página del buscador",
		},
		{
			nombre: "formato con otra ruta de formulario",
			tabla: con(columnaFormato, "HTML; cookie de `/search/indexAN.jsp`; formulario `POST /search/buscar.action`; "+
				"campos `ECLI`, `ROJ` y `NUMERORESOLUCION`"),
			mensaje: "«Formato» da como formulario https://www.poderjudicial.es/search/buscar.action",
		},
		{
			nombre: "formato sin un campo",
			tabla: con(columnaFormato, "HTML; cookie de `/search/indexAN.jsp`; formulario `POST /search/search.action`; "+
				"campos `ECLI`, ROJ y `NUMERORESOLUCION`"),
			mensaje: "el adaptador declara el campo ROJ, que la celda «Formato» no nombra",
		},
		{
			nombre: "formato con un campo más",
			tabla: con(columnaFormato, "HTML; cookie de `/search/indexAN.jsp`; formulario `POST /search/search.action`; "+
				"campos `ECLI`, `ROJ`, `NUMERORESOLUCION` y `NUMERORECURSO`"),
			mensaje: "«Formato» nombra el campo NUMERORECURSO, que el adaptador no declara",
		},
	}
}

// tablaDeSources es un documento con la tabla de fuentes y las filas que recibe,
// entre un título y un párrafo.
func tablaDeSources(filas ...string) string {
	lineas := []string{
		"# Fuentes",
		"",
		"| " + strings.Join(cabeceraDeSources, " | ") + " |",
		"|" + strings.Repeat("---|", len(cabeceraDeSources)),
	}

	lineas = append(lineas, filas...)
	lineas = append(lineas, "", "Texto que sigue a la tabla.")

	return strings.Join(lineas, "\n")
}

// filaDeCendoj es la fila de la fuente con las celdas de docs/SOURCES.md que el
// test compara, salvo las que cambia quien la pide, por su columna.
func filaDeCendoj(cambiadas map[string]string) string {
	celdas := map[string]string{
		columnaFuente:   "`" + NombreDeLaFuente + "`",
		columnaBase:     celdaBase,
		columnaTerminos: celdaTerminos,
		columnaRitmo:    celdaRitmo,
		columnaFormato:  celdaFormato,
		columnaRevisado: celdaRevisado,
	}

	fila := make([]string, 0, len(cabeceraDeSources))

	for _, columna := range cabeceraDeSources {
		celda, fijada := celdas[columna]
		if !fijada {
			celda = "por revisar"
		}

		if cambiada, loEsta := cambiadas[columna]; loEsta {
			celda = cambiada
		}

		fila = append(fila, celda)
	}

	return "| " + strings.Join(fila, " | ") + " |"
}

// comprobarFilaDeLaFuente localiza en la tabla de fuentes la fila cuya primera
// celda es el nombre de la fuente y comprueba que dice lo que el adaptador
// declara. Devuelve todas las divergencias juntas, cada una con su celda.
func comprobarFilaDeLaFuente(tabla string, declarado loDeclarado) error {
	celdas, err := filaDeLaFuente(tabla, declarado.fuente)
	if err != nil {
		return err
	}

	celda := func(columna string) string {
		return celdas[slices.Index(cabeceraDeSources, columna)]
	}

	return errors.Join(
		comprobarRitmo(celda(columnaRitmo), declarado.intervalo),
		comprobarDireccion(columnaTerminos, celda(columnaTerminos), declarado.terminos.URL),
		comprobarRevision(celda(columnaRevisado), declarado.terminos.Revisados),
		comprobarDireccion(columnaBase, celda(columnaBase), declarado.base),
		comprobarFormato(celda(columnaFormato), declarado),
	)
}

// filaDeLaFuente devuelve las celdas de la única fila de la tabla de fuentes cuya
// primera celda, sin comillas invertidas, es el nombre de la fuente. La tabla es
// la que empieza en su cabecera y llega hasta la primera línea que no es de
// tabla.
func filaDeLaFuente(tabla, fuente string) ([]string, error) {
	lineas := strings.Split(tabla, "\n")

	inicio := slices.IndexFunc(lineas, func(linea string) bool {
		celdas, esDeTabla := celdasDe(linea)

		return esDeTabla && slices.Equal(celdas, cabeceraDeSources)
	})
	if inicio < 0 {
		return nil, fmt.Errorf("docs/SOURCES.md no tiene la tabla de fuentes con la cabecera %s",
			strings.Join(cabeceraDeSources, ", "))
	}

	var encontradas [][]string

	for numero, linea := range lineas[inicio+1:] {
		celdas, esDeTabla := celdasDe(linea)
		if !esDeTabla {
			break
		}

		if slices.ContainsFunc(celdas, separadorDeCelda.MatchString) || strings.Trim(celdas[0], "`") != fuente {
			continue
		}

		if len(celdas) != len(cabeceraDeSources) {
			return nil, fmt.Errorf("docs/SOURCES.md, línea %d: la fila de %s tiene %d celdas y la tabla, %d",
				inicio+numero+2, fuente, len(celdas), len(cabeceraDeSources))
		}

		encontradas = append(encontradas, celdas)
	}

	switch len(encontradas) {
	case 1:
		return encontradas[0], nil
	case 0:
		return nil, fmt.Errorf("docs/SOURCES.md no tiene ninguna fila de %s", fuente)
	default:
		return nil, fmt.Errorf("docs/SOURCES.md tiene %d filas de %s y tiene que tener una", len(encontradas), fuente)
	}
}

// celdasDe parte una línea de tabla de Markdown en sus celdas, sin el espacio que
// las rodea, y dice si la línea es de tabla: empieza y termina por «|».
func celdasDe(linea string) ([]string, bool) {
	linea = strings.TrimSpace(linea)
	if len(linea) < 2 || !strings.HasPrefix(linea, "|") || !strings.HasSuffix(linea, "|") {
		return nil, false
	}

	celdas := strings.Split(linea[1:len(linea)-1], "|")
	for indice, celda := range celdas {
		celdas[indice] = strings.TrimSpace(celda)
	}

	return celdas, true
}

// comprobarRitmo exige que la celda «Ritmo», un literal de duración de Go entre
// comillas invertidas, sea el intervalo entre peticiones del adaptador.
func comprobarRitmo(celda string, intervalo time.Duration) error {
	ritmo, err := time.ParseDuration(strings.Trim(celda, "`"))
	if err != nil {
		return fmt.Errorf("la celda «%s» no es una duración de Go: %w", columnaRitmo, err)
	}

	if ritmo != intervalo {
		return fmt.Errorf("la celda «%s» es %s y el adaptador declara %s", columnaRitmo, ritmo, intervalo)
	}

	return nil
}

// comprobarDireccion exige que la celda lleve una sola dirección entre <…> y que
// sea la que el adaptador declara.
func comprobarDireccion(columna, celda, declarada string) error {
	encontradas := direccionEntreAngulos.FindAllStringSubmatch(celda, -1)
	if len(encontradas) != 1 {
		return fmt.Errorf("la celda «%s» lleva %d direcciones entre <…> y tiene que llevar una: %s",
			columna, len(encontradas), celda)
	}

	if encontradas[0][1] != declarada {
		return fmt.Errorf("la celda «%s» es %s y el adaptador declara %s", columna, encontradas[0][1], declarada)
	}

	return nil
}

// comprobarRevision exige que la celda «Revisado» sea una fecha AAAA-MM-DD que
// no es cero e igual al día de la revisión de los términos.
func comprobarRevision(celda string, revisados time.Time) error {
	fecha, err := time.Parse(time.DateOnly, celda)
	if err != nil {
		return fmt.Errorf("la celda «%s» no es una fecha AAAA-MM-DD: %w", columnaRevisado, err)
	}

	if fecha.IsZero() {
		return fmt.Errorf("la celda «%s» es la fecha cero, que no es ninguna revisión", columnaRevisado)
	}

	if !fecha.Equal(revisados) {
		declarada := "cero"
		if !revisados.IsZero() {
			declarada = revisados.Format(time.DateOnly)
		}

		return fmt.Errorf("la celda «%s» es %s y el adaptador declara %s", columnaRevisado, celda, declarada)
	}

	return nil
}

// comprobarFormato lee de la celda «Formato» lo que nombra entre comillas
// invertidas —la ruta de la página del buscador, la del formulario, detrás de
// su método, y los campos de consulta— y exige que sea lo que el adaptador
// declara. Las rutas son del sitio de la base, que ata la celda «Base».
func comprobarFormato(celda string, declarado loDeclarado) error {
	var paginas, formularios, campos []string

	for _, tramo := range tramoDeCodigo.FindAllStringSubmatch(celda, -1) {
		switch texto := tramo[1]; {
		case strings.HasPrefix(texto, envioDelFormulario):
			formularios = append(formularios, strings.TrimPrefix(texto, envioDelFormulario))
		case strings.HasPrefix(texto, "/"):
			paginas = append(paginas, texto)
		default:
			campos = append(campos, texto)
		}
	}

	return errors.Join(
		comprobarRuta("página del buscador", paginas, declarado.base, declarado.pagina),
		comprobarRuta("formulario", formularios, declarado.base, declarado.formulario),
		comprobarCampos(campos, declarado.campos),
	)
}

// comprobarRuta exige que la celda «Formato» nombre una sola ruta de esa clase y
// que, en el sitio de la base, sea la dirección que el adaptador declara.
func comprobarRuta(clase string, rutas []string, base, declarada string) error {
	if len(rutas) != 1 {
		return fmt.Errorf("la celda «%s» nombra %d rutas de %s y tiene que nombrar una", columnaFormato, len(rutas), clase)
	}

	sitio, err := url.Parse(base)
	if err != nil {
		return fmt.Errorf("la base que el adaptador declara no se puede interpretar: %w", err)
	}

	ruta, err := url.Parse(rutas[0])
	if err != nil {
		return fmt.Errorf("la celda «%s» nombra una ruta de %s que no se puede interpretar: %w", columnaFormato, clase, err)
	}

	if direccion := sitio.ResolveReference(ruta).String(); direccion != declarada {
		return fmt.Errorf("la celda «%s» da como %s %s y el adaptador declara %s",
			columnaFormato, clase, direccion, declarada)
	}

	return nil
}

// comprobarCampos exige que los campos de consulta que nombra la celda «Formato»
// y los que el adaptador declara sean los mismos, sin que sobre ni falte ninguno.
func comprobarCampos(nombrados, declarados []string) error {
	var divergencias []error

	for _, campo := range nombrados {
		if !slices.Contains(declarados, campo) {
			divergencias = append(divergencias, fmt.Errorf(
				"la celda «%s» nombra el campo %s, que el adaptador no declara", columnaFormato, campo))
		}
	}

	for _, campo := range declarados {
		if !slices.Contains(nombrados, campo) {
			divergencias = append(divergencias, fmt.Errorf(
				"el adaptador declara el campo %s, que la celda «%s» no nombra", campo, columnaFormato))
		}
	}

	return errors.Join(divergencias...)
}
