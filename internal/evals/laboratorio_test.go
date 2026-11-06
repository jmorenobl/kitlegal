package evals

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"maps"
	"os"
	"path"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal"
	"github.com/jmorenobl/kitlegal/internal/app"
	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/mcp"
	"github.com/jmorenobl/kitlegal/internal/skills"
)

// respuestasDeLaReglaTestigo es el directorio de las respuestas fijas con las
// que TestReglaTestigo comprueba el guion de la regla testigo, cada una con su
// veredicto en el nombre (contracts/laboratorio.md §3 de H39).
const respuestasDeLaReglaTestigo = "testdata/laboratorio/respuestas"

// Los principios del nombre de una respuesta fija, que dicen su veredicto: la
// que empieza por el de no cumplir no cumple la regla y la que empieza por el de
// cumplir la cumple.
const (
	principioDeLaQueCumple   = "cumple"
	principioDeLaQueNoCumple = "no-cumple"
)

// nombresDeLasRespuestasFijas son las respuestas fijas de la regla testigo: una
// que la cumple con el sello al final, otra con líneas vacías detrás, y una por
// cada forma de no cumplirla que la regla distingue —el sello no aparece,
// aparece y no es la última línea, es la última con formato y es la última con
// otro texto— (FR-014 de H39; data-model §5 de H39).
var nombresDeLasRespuestasFijas = []string{
	"cumple.txt",
	"cumple-con-lineas-en-blanco.txt",
	"no-cumple-sin-sello.txt",
	"no-cumple-sello-no-al-final.txt",
	"no-cumple-sello-en-negrita.txt",
	"no-cumple-sello-con-texto.txt",
}

// La respuesta de una skill de kitlegal a una pregunta de las evals de
// laboratorio: la de la primera sesión de Sonnet en el modo orden de la eval
// del artículo 140 de la Constitución, la 01 del laboratorio, en el informe de
// evals de normativa que versionó H32 (research V20 de H39). El informe está
// versionado, solo se lee y no se edita.
const (
	informeDeNormativaDeH32 = "../../specs/030-h32-arquitectura-de-skills/gates/evals/normativa.json"
	evalDeLaSesionDeH32     = "09-constitucion-articulo-140.yaml"
	modeloDeLaSesionDeH32   = "claude-sonnet-5-5"
	numeroDeLaSesionDeH32   = 1
)

// skillsDeKitlegal es el directorio de las skills de kitlegal, desde el del
// paquete: las que empotra el binario, con su SKILL.md y sus referencias.
const skillsDeKitlegal = "../../skills"

// TestReglaTestigo comprueba el guion de la regla testigo (FR-013, FR-014 y
// SC-003 de H39; contracts/laboratorio.md §3 de H39): acierta el veredicto de
// cada respuesta fija y el de la respuesta de una skill de kitlegal, que no la
// cumple, y la línea del sello es inventada para el laboratorio, así que no
// aparece en ninguna skill ni referencia de kitlegal ni en las instrucciones del
// servidor MCP.
func TestReglaTestigo(t *testing.T) {
	t.Parallel()

	t.Run("respuestas fijas", comprobarLasRespuestasFijas)
	t.Run("el párrafo de la regla", comprobarElParrafoDeLaRegla)
	t.Run("respuesta de una skill de kitlegal", comprobarLaRespuestaDeKitlegal)
	t.Run("sello fuera de kitlegal", comprobarElSelloFueraDeKitlegal)
}

// comprobarLasRespuestasFijas: el directorio tiene exactamente las respuestas
// fijas, y el guion da a cada una el veredicto de su nombre.
func comprobarLasRespuestasFijas(t *testing.T) {
	t.Parallel()

	entradas, err := os.ReadDir(respuestasDeLaReglaTestigo)
	require.NoError(t, err, "el directorio %s de las respuestas fijas se lee", respuestasDeLaReglaTestigo)

	nombres := make([]string, 0, len(entradas))
	for _, entrada := range entradas {
		nombres = append(nombres, entrada.Name())
	}

	require.ElementsMatch(t, nombresDeLasRespuestasFijas, nombres,
		"%s tiene las respuestas fijas de la regla testigo y ninguna más", respuestasDeLaReglaTestigo)

	for _, nombre := range nombresDeLasRespuestasFijas {
		cumple := veredictoDelNombre(t, nombre)
		respuesta := string(contenidoDelFichero(t, path.Join(respuestasDeLaReglaTestigo, nombre)))

		require.NotEmpty(t, strings.TrimSpace(respuesta), "la respuesta fija %s tiene texto", nombre)
		assert.Equal(t, cumple, reglaTestigo(respuesta), "el guion da a %s el veredicto de su nombre", nombre)
	}
}

// veredictoDelNombre es el veredicto que dice el nombre de una respuesta fija:
// cumple si empieza por el principio de la que cumple y no cumple si empieza
// por el de la que no. Un nombre sin ninguno de los dos no dice veredicto.
func veredictoDelNombre(t *testing.T, nombre string) bool {
	t.Helper()

	switch {
	case strings.HasPrefix(nombre, principioDeLaQueNoCumple+"-"):
		return false
	case strings.HasPrefix(nombre, principioDeLaQueCumple):
		return true
	default:
		require.Failf(t, "respuesta fija sin veredicto",
			"el nombre %s no empieza por %q ni por %q", nombre, principioDeLaQueCumple, principioDeLaQueNoCumple)

		return false
	}
}

// comprobarElParrafoDeLaRegla: el párrafo de la regla termina con la línea del
// sello que pide, y el párrafo mismo, copiado como última línea, no la cumple.
func comprobarElParrafoDeLaRegla(t *testing.T) {
	t.Parallel()

	assert.True(t, strings.HasSuffix(parrafoDeLaRegla, ": "+lineaDelSello),
		"el párrafo de la regla dice exactamente la línea del sello: %q", parrafoDeLaRegla)
	assert.False(t, reglaTestigo(parrafoDeLaRegla+"\n"),
		"la regla copiada como última línea lleva el sello con otro texto y no la cumple")
}

// comprobarLaRespuestaDeKitlegal: la respuesta de la sesión de H32, leída del
// informe versionado, es la de una sesión de la eval del artículo 140 que activó
// la skill normativa, y no cumple la regla.
func comprobarLaRespuestaDeKitlegal(t *testing.T) {
	t.Parallel()

	var informe struct {
		Evals []entradaDelInforme `json:"evals"`
	}
	require.NoError(t, json.Unmarshal(contenidoDelFichero(t, informeDeNormativaDeH32), &informe),
		"%s es un informe del job de evals", informeDeNormativaDeH32)

	nombre := nombreDeSesion(evalDeLaSesionDeH32, modeloDeLaSesionDeH32, ModoOrden, numeroDeLaSesionDeH32, false)

	var sesiones []entradaDelInforme
	for _, entrada := range informe.Evals {
		if entrada.Sesion == nombre {
			sesiones = append(sesiones, entrada)
		}
	}

	require.Len(t, sesiones, 1, "el informe %s tiene una sesión %s", informeDeNormativaDeH32, nombre)
	sesion := sesiones[0]
	require.Equal(t, evalDeLaSesionDeH32, sesion.Eval, "la sesión %s es de la eval del artículo 140", nombre)
	require.True(t, sesion.Activada, "la sesión %s activó la skill normativa", nombre)
	require.NotEmpty(t, strings.TrimSpace(sesion.Respuesta), "la sesión %s tiene respuesta", nombre)

	assert.False(t, reglaTestigo(sesion.Respuesta),
		"la respuesta de la skill normativa en la sesión %s no cumple la regla testigo", nombre)
}

// comprobarElSelloFueraDeKitlegal: la línea del sello no aparece en ningún
// fichero de las skills de kitlegal —ni en un SKILL.md ni en una referencia— ni
// en las instrucciones que da el servidor MCP. Lee al menos un SKILL.md y una
// referencia, para no pasar en vacío.
func comprobarElSelloFueraDeKitlegal(t *testing.T) {
	t.Parallel()

	skills := os.DirFS(skillsDeKitlegal)
	skillsLeidas, referenciasLeidas := 0, 0

	err := fs.WalkDir(skills, ".", func(ruta string, entrada fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entrada.IsDir() {
			return nil
		}

		contenido, err := fs.ReadFile(skills, ruta)
		if err != nil {
			return err
		}

		assert.NotContains(t, string(contenido), lineaDelSello,
			"el fichero %s de las skills de kitlegal no lleva la línea del sello", path.Join(skillsDeKitlegal, ruta))

		switch {
		case path.Base(ruta) == "SKILL.md":
			skillsLeidas++
		case path.Base(path.Dir(ruta)) == "references":
			referenciasLeidas++
		}

		return nil
	})
	require.NoError(t, err, "las skills de kitlegal en %s se leen", skillsDeKitlegal)

	assert.Positive(t, skillsLeidas, "se lee al menos un SKILL.md de %s", skillsDeKitlegal)
	assert.Positive(t, referenciasLeidas, "se lee al menos una referencia de %s", skillsDeKitlegal)
	assert.NotContains(t, mcp.Instrucciones, lineaDelSello,
		"las instrucciones del servidor MCP no llevan la línea del sello")
}

// Lo que la declaración del laboratorio dice de sus skills, su herramienta, su
// parámetro y su referencia (contracts/laboratorio.md §2 de H39; data-model §2
// de H39).
const (
	skillDelLaboratorio       = "articulos-del-boe"
	entradaDelLaboratorio     = "consulta-legal"
	herramientaDelLaboratorio = "boe_articulo"
	parametroDelLaboratorio   = "bloque"
	referenciaDelLaboratorio  = "cierre.md"
)

// Las variantes que el control nombra por sí mismas: la de control, con la que
// se compara cada una, y la que saca las herramientas de la descripción.
const (
	varianteDeControl       = "control"
	varianteSinHerramientas = "descripcion-sin-herramientas"
)

// Los textos de las variantes que no son el párrafo de la regla
// (contracts/laboratorio.md §5 de H39): la frase que la variante de las
// instrucciones añade a las de kitlegal, y el título de la referencia de las tres
// variantes que la llevan.
const (
	fraseDeLasInstrucciones = "Antes de la primera herramienta de kitlegal, activa la skill articulos-del-boe."
	tituloDeLaReferencia    = "# Cómo termina la respuesta"
)

// varianteEsperada es una variante de la declaración del laboratorio tal como la
// fijan FR-020 a FR-027 de H39 (data-model §2 de H39): su nombre, sus modos y sus
// partes, y la parte donde lleva el párrafo de la regla testigo.
type varianteEsperada struct {
	nombre string
	modos  []Modo
	cambia []parteDeUnaVariante
	regla  parteDeUnaVariante
}

// variantesEsperadas son las ocho variantes del laboratorio, en el orden de la
// declaración (FR-020 a FR-027 de H39; contracts/laboratorio.md §2 de H39): la
// regla va en el cuerpo de la de control y de las que no lo cambian, en la
// referencia de las tres que la llevan y en el texto del parámetro de la que lo
// cambia.
var variantesEsperadas = []varianteEsperada{
	{nombre: varianteDeControl, modos: []Modo{ModoOrden, ModoHerramienta}, cambia: []parteDeUnaVariante{}, regla: parteCuerpo},
	{
		nombre: "instrucciones", modos: []Modo{ModoHerramienta},
		cambia: []parteDeUnaVariante{parteInstrucciones}, regla: parteCuerpo,
	},
	{
		nombre: "entrada", modos: []Modo{ModoOrden, ModoHerramienta},
		cambia: []parteDeUnaVariante{parteEntrada}, regla: parteCuerpo,
	},
	{
		nombre: "referencia-secundaria", modos: []Modo{ModoOrden, ModoHerramienta},
		cambia: []parteDeUnaVariante{parteCuerpo, parteReferencia}, regla: parteReferencia,
	},
	{
		nombre: "parametro", modos: []Modo{ModoHerramienta},
		cambia: []parteDeUnaVariante{parteCuerpo, parteParametro}, regla: parteParametro,
	},
	{
		nombre: "referencia-paso", modos: []Modo{ModoOrden, ModoHerramienta},
		cambia: []parteDeUnaVariante{parteCuerpo, parteReferencia}, regla: parteReferencia,
	},
	{
		nombre: "referencia-en-el-punto", modos: []Modo{ModoOrden, ModoHerramienta},
		cambia: []parteDeUnaVariante{parteCuerpo, parteReferencia}, regla: parteReferencia,
	},
	{
		nombre: varianteSinHerramientas, modos: []Modo{ModoOrden, ModoHerramienta},
		cambia: []parteDeUnaVariante{parteDescripcion}, regla: parteCuerpo,
	},
}

// declaracionEsperada es la declaración del laboratorio de
// contracts/laboratorio.md §2 de H39, con las variantes de variantesEsperadas.
func declaracionEsperada() declaracionDelLaboratorio {
	declarada := declaracionDelLaboratorio{
		Skill:       skillDelLaboratorio,
		Entrada:     entradaDelLaboratorio,
		Herramienta: herramientaDelLaboratorio,
		Parametro:   parametroDelLaboratorio,
		Referencia:  referenciaDelLaboratorio,
	}

	for _, esperada := range variantesEsperadas {
		declarada.Variantes = append(declarada.Variantes,
			varianteDelLaboratorio{Nombre: esperada.nombre, Modos: esperada.modos, Cambia: esperada.cambia})
	}

	return declarada
}

// TestVariantesDelLaboratorio es el control de la skill de laboratorio y sus
// variantes (FR-011, FR-020 a FR-028 y SC-004 de H39; contracts/laboratorio.md
// §1, §2, §4, §5 y §7 de H39; data-model §2 y §3 de H39): el lector estricto de
// la declaración; la declaración, con las ocho variantes en orden, sus modos y
// sus partes, y lo que nombra; cada variante frente a la de control, parte a
// parte, sin ningún fichero fuera de una parte; la regla testigo donde la pone
// cada variante; los textos que la sesión anuncia en lugar de los de kitlegal;
// lo que nombran las descripciones y la skill de entrada; y que las skills de
// cada variante tienen la forma de las de kitlegal.
func TestVariantesDelLaboratorio(t *testing.T) {
	t.Parallel()

	t.Run("lector estricto", comprobarElLectorDeLaDeclaracion)

	declarada, err := leerLaDeclaracionDelRepositorio(raizDelRepositorio)
	require.NoError(t, err)

	registro, err := app.RegistroDeProduccion("")
	require.NoError(t, err)

	leidas := map[string]partesDeUnaVarianteLeidas{}

	for _, variante := range declarada.Variantes {
		leidas[variante.Nombre], err = leerLasPartesDeLaVariante(
			directorioDeLaVariante(raizDelRepositorio, variante.Nombre), declarada)
		require.NoError(t, err)
	}

	require.Contains(t, leidas, varianteDeControl, "la declaración tiene la variante de control")

	t.Run("declaración", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, declaracionEsperada(), declarada)
	})

	t.Run("lo que nombra", func(t *testing.T) {
		t.Parallel()

		comprobarLoQueNombraLaDeclaracion(t, registro, declarada)
	})

	t.Run("partes", func(t *testing.T) {
		t.Parallel()

		comprobarLasPartes(t, declarada, leidas)
	})

	t.Run("regla testigo", func(t *testing.T) {
		t.Parallel()

		comprobarDondeVaLaRegla(t, leidas)
	})

	t.Run("textos de la sesión", func(t *testing.T) {
		t.Parallel()

		comprobarLosTextosDeLaSesion(t, registro, declarada, leidas)
	})

	t.Run("descripciones y entrada", func(t *testing.T) {
		t.Parallel()

		comprobarLoQueNombranLasSkills(t, registro, declarada)
	})

	t.Run("forma de las skills de kitlegal", func(t *testing.T) {
		t.Parallel()

		comprobarLaFormaDelProducto(t, declarada)
	})
}

// declaracionMinima es una declaración del laboratorio bien formada, la más
// corta, sobre la que comprobarElLectorDeLaDeclaracion cambia una sola cosa.
const declaracionMinima = "skill: s\nentrada: e\nherramienta: boe_articulo\nparametro: bloque\nreferencia: cierre.md\n" +
	"variantes:\n  - {nombre: control, modos: [orden], cambia: []}\n"

// comprobarElLectorDeLaDeclaracion: el lector de la declaración lee una
// declaración bien formada y rechaza, en la raíz o en una variante, una clave
// repetida, una clave que la declaración no tiene y una que le falta, y un modo o
// una parte fuera de su vocabulario (data-model §2 de H39).
func comprobarElLectorDeLaDeclaracion(t *testing.T) {
	t.Parallel()

	leida, err := leerLaDeclaracionDelLaboratorio([]byte(declaracionMinima))
	require.NoError(t, err)
	assert.Equal(t, declaracionDelLaboratorio{
		Skill: "s", Entrada: "e", Herramienta: "boe_articulo", Parametro: "bloque", Referencia: "cierre.md",
		Variantes: []varianteDelLaboratorio{
			{Nombre: "control", Modos: []Modo{ModoOrden}, Cambia: []parteDeUnaVariante{}},
		},
	}, leida)

	variante := "  - {nombre: control, modos: [orden], cambia: []}\n"
	casos := []struct {
		nombre, documento string
	}{
		{nombre: "clave repetida en la raíz", documento: "skill: t\n" + declaracionMinima},
		{nombre: "clave desconocida en la raíz", documento: "otra: x\n" + declaracionMinima},
		{nombre: "clave que falta", documento: strings.Replace(declaracionMinima, "referencia: cierre.md\n", "", 1)},
		{
			nombre:    "clave repetida en una variante",
			documento: strings.Replace(declaracionMinima, variante, "  - {nombre: control, nombre: otra, modos: [orden], cambia: []}\n", 1),
		},
		{
			nombre:    "clave desconocida en una variante",
			documento: strings.Replace(declaracionMinima, variante, "  - {nombre: control, modos: [orden], cambia: [], otra: x}\n", 1),
		},
		{
			nombre:    "modo desconocido",
			documento: strings.Replace(declaracionMinima, "modos: [orden]", "modos: [otro]", 1),
		},
		{
			nombre:    "parte desconocida",
			documento: strings.Replace(declaracionMinima, "cambia: []", "cambia: [otra]", 1),
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			require.NotEqual(t, declaracionMinima, caso.documento, "el caso cambia la declaración")

			_, err := leerLaDeclaracionDelLaboratorio([]byte(caso.documento))
			require.Error(t, err)
		})
	}
}

// comprobarLoQueNombraLaDeclaracion: la skill de laboratorio y la de entrada no
// son skills de kitlegal; la herramienta es una de las que anuncia el servidor
// MCP de kitlegal; y el parámetro es una propiedad de su entrada (data-model §2
// de H39).
func comprobarLoQueNombraLaDeclaracion(t *testing.T, registro *app.Registro, declarada declaracionDelLaboratorio) {
	t.Helper()

	deKitlegal, err := skills.Listar(raizDelRepositorio)
	require.NoError(t, err)
	require.NotEmpty(t, deKitlegal, "se leen las skills de kitlegal")

	assert.NotContains(t, deKitlegal, declarada.Skill, "la skill de laboratorio no es una skill de kitlegal")
	assert.NotContains(t, deKitlegal, declarada.Entrada, "la skill de entrada no es una skill de kitlegal")
	assert.Contains(t, app.NombresDeHerramientas(registro), declarada.Herramienta,
		"la herramienta del laboratorio la anuncia el servidor MCP de kitlegal")

	descripcionDelParametro(t, registro, declarada.Herramienta, declarada.Parametro)
}

// descripcionDelParametro es la descripción que el servidor MCP de kitlegal da al
// parámetro de la herramienta, la que la sesión de las demás variantes anuncia:
// la de la propiedad de ese nombre en la entrada de la herramienta, construida
// con el mismo generador que usa el servidor (cli.EsquemasDeHerramienta) sobre el
// verbo del registro de producción que la herramienta nombra, `<applet>_<verbo>`.
// Falla si el parámetro no es una propiedad de la entrada o no tiene descripción.
func descripcionDelParametro(t *testing.T, registro *app.Registro, herramienta, parametro string) string {
	t.Helper()

	delApplet, delVerbo, conSeparador := strings.Cut(herramienta, separadorDeHerramienta)
	require.True(t, conSeparador, "la herramienta %s es <applet>%s<verbo>", herramienta, separadorDeHerramienta)

	applet, registrado := registro.Buscar(delApplet)
	require.True(t, registrado, "el applet %s está en el registro de producción", delApplet)

	posicion := slices.IndexFunc(applet.Verbos(), func(verbo app.Verbo) bool { return verbo.Nombre == delVerbo })
	require.NotEqual(t, -1, posicion, "el verbo %s está en el applet %s", delVerbo, delApplet)

	verbo := applet.Verbos()[posicion]

	entrada, _, err := cli.EsquemasDeHerramienta(cli.Verbo{
		Applet:     applet.Nombre(),
		Verbo:      verbo.Nombre,
		Ayuda:      verbo.Descripcion,
		Argumentos: verbo.Argumentos(),
		Salida:     verbo.Salida,
	})
	require.NoError(t, err)

	var esquema struct {
		Propiedades map[string]struct {
			Descripcion string `json:"description"`
		} `json:"properties"`
	}

	require.NoError(t, json.Unmarshal(entrada, &esquema))

	propiedad, esPropiedad := esquema.Propiedades[parametro]
	require.True(t, esPropiedad, "%s es una propiedad de la entrada de %s", parametro, herramienta)
	require.NotEmpty(t, propiedad.Descripcion, "la propiedad %s de la entrada de %s tiene descripción",
		parametro, herramienta)

	return propiedad.Descripcion
}

// comprobarLasPartes es FR-028 de H39: cada variante frente a la de control,
// parte a parte, idéntica byte a byte en las partes que no declara y distinta en
// las que declara, sin ningún fichero fuera de una parte. La de control solo
// tiene la descripción y el cuerpo de su SKILL.md (data-model §3 de H39).
func comprobarLasPartes(t *testing.T, declarada declaracionDelLaboratorio,
	leidas map[string]partesDeUnaVarianteLeidas,
) {
	t.Helper()

	control := leidas[varianteDeControl]
	assert.ElementsMatch(t, []parteDeUnaVariante{parteDescripcion, parteCuerpo}, slices.Collect(maps.Keys(control.partes)),
		"la variante de control solo tiene la descripción y el cuerpo de su SKILL.md")

	for _, variante := range declarada.Variantes {
		deLaVariante := leidas[variante.Nombre]

		assert.Empty(t, deLaVariante.fuera, "la variante %s no tiene ningún fichero fuera de sus partes", variante.Nombre)

		for _, parte := range partesDeLasVariantes {
			if slices.Contains(variante.Cambia, parte) {
				assert.NotEqual(t, control.partes[parte], deLaVariante.partes[parte],
					"la variante %s declara %s y la cambia", variante.Nombre, parte)

				continue
			}

			assert.Equal(t, control.partes[parte], deLaVariante.partes[parte],
				"la variante %s no declara %s y la tiene idéntica a la de control", variante.Nombre, parte)
		}
	}
}

// comprobarDondeVaLaRegla: el párrafo de la regla testigo está en la parte donde
// lo pone cada variante, y la línea del sello no está en el cuerpo de las que lo
// sacan de él (data-model §3 de H39).
func comprobarDondeVaLaRegla(t *testing.T, leidas map[string]partesDeUnaVarianteLeidas) {
	t.Helper()

	for _, esperada := range variantesEsperadas {
		deLaVariante := leidas[esperada.nombre]

		assert.True(t, algunoContiene(deLaVariante.partes[esperada.regla], parrafoDeLaRegla),
			"la variante %s lleva el párrafo de la regla en %s", esperada.nombre, esperada.regla)

		if esperada.regla != parteCuerpo {
			assert.False(t, algunoContiene(deLaVariante.partes[parteCuerpo], lineaDelSello),
				"la variante %s no lleva la línea del sello en el cuerpo", esperada.nombre)
		}
	}
}

// algunoContiene dice si alguno de los ficheros de una parte contiene el texto.
func algunoContiene(ficheros map[string]string, texto string) bool {
	for _, contenido := range ficheros {
		if strings.Contains(contenido, texto) {
			return true
		}
	}

	return false
}

// comprobarLosTextosDeLaSesion: lo que añade cada variante que cambia uno de los
// textos de la sesión o la referencia (contracts/laboratorio.md §5 de H39). Las
// instrucciones son las de kitlegal, un espacio y la frase que pide activar la
// skill, que la nombra; el texto del parámetro, la descripción que kitlegal le da,
// un espacio y el párrafo de la regla; y la referencia, su único fichero, con el
// título, una línea vacía y el párrafo de la regla. Cada uno, una línea con su
// salto final.
func comprobarLosTextosDeLaSesion(t *testing.T, registro *app.Registro, declarada declaracionDelLaboratorio,
	leidas map[string]partesDeUnaVarianteLeidas,
) {
	t.Helper()

	instrucciones := mcp.Instrucciones + " " + fraseDeLasInstrucciones + "\n"
	parametro := descripcionDelParametro(t, registro, declarada.Herramienta, declarada.Parametro) + " " +
		parrafoDeLaRegla + "\n"
	referencia := path.Join(carpetaDeLasSkillsDeLaVariante, declarada.Skill, carpetaDeLasReferenciasDeLaSkill,
		declarada.Referencia)

	assert.Contains(t, fraseDeLasInstrucciones, declarada.Skill, "las instrucciones nombran la skill de laboratorio")

	textos := 0

	for _, variante := range declarada.Variantes {
		partes := leidas[variante.Nombre].partes

		if slices.Contains(variante.Cambia, parteInstrucciones) {
			textos++

			assert.Equal(t, map[string]string{ficheroDeLasInstrucciones: instrucciones}, partes[parteInstrucciones],
				"las instrucciones de la variante %s", variante.Nombre)
		}

		if slices.Contains(variante.Cambia, parteParametro) {
			textos++

			assert.Equal(t, map[string]string{ficheroDelParametro: parametro}, partes[parteParametro],
				"el texto del parámetro de la variante %s", variante.Nombre)
		}

		if slices.Contains(variante.Cambia, parteReferencia) {
			textos++

			assert.Equal(t, map[string]string{referencia: tituloDeLaReferencia + "\n\n" + parrafoDeLaRegla + "\n"},
				partes[parteReferencia], "la referencia de la variante %s", variante.Nombre)
		}
	}

	assert.Equal(t, 5, textos, "una variante cambia las instrucciones, una el parámetro y tres la referencia")
}

// comprobarLoQueNombranLasSkills: la descripción de la skill de control nombra
// alguna herramienta de kitlegal, y la de la variante sin herramientas ninguna, ni
// una orden `kitlegal <applet>`; y el cuerpo de la skill de entrada nombra la
// skill de laboratorio (data-model §3 de H39; FR-020, FR-022 y FR-027 de H39).
func comprobarLoQueNombranLasSkills(t *testing.T, registro *app.Registro, declarada declaracionDelLaboratorio) {
	t.Helper()

	herramientas := app.NombresDeHerramientas(registro)
	require.NotEmpty(t, herramientas)

	nombraAlguna := func(texto string) bool {
		return slices.ContainsFunc(herramientas, func(herramienta string) bool {
			return strings.Contains(texto, herramienta)
		})
	}

	deControl := frontmatterDeLaSkill(t, varianteDeControl, declarada.Skill)
	assert.True(t, nombraAlguna(deControl.Description), "la descripción de control nombra alguna herramienta")

	sinHerramientas := frontmatterDeLaSkill(t, varianteSinHerramientas, declarada.Skill)
	assert.False(t, nombraAlguna(sinHerramientas.Description), "la descripción de %s no nombra ninguna herramienta",
		varianteSinHerramientas)

	for _, applet := range registro.Nombres() {
		assert.NotContains(t, sinHerramientas.Description, programaDeLasConsultas+" "+applet,
			"la descripción de %s no nombra la orden del applet %s", varianteSinHerramientas, applet)
	}

	for _, variante := range declarada.Variantes {
		if !slices.Contains(variante.Cambia, parteEntrada) {
			continue
		}

		contenido := skillMdDeLaVariante(t, variante.Nombre, declarada.Entrada)

		_, cuerpo, cerrado := strings.Cut(strings.TrimPrefix(string(contenido), "---\n"), "\n---\n")
		require.True(t, cerrado, "el SKILL.md de %s en la variante %s cierra su frontmatter", declarada.Entrada,
			variante.Nombre)
		assert.Contains(t, cuerpo, declarada.Skill, "el cuerpo de la skill de entrada nombra la de laboratorio")
	}
}

// skillMdDeLaVariante es el contenido del SKILL.md de esa skill en esa variante.
func skillMdDeLaVariante(t *testing.T, variante, skill string) []byte {
	t.Helper()

	contenido, err := fs.ReadFile(os.DirFS(directorioDeLaVariante(raizDelRepositorio, variante)),
		path.Join(carpetaDeLasSkillsDeLaVariante, skill, ficheroDeLaSkillDeLaVariante))
	require.NoError(t, err)

	return contenido
}

// frontmatterDeLaSkill es el frontmatter del SKILL.md de esa skill en esa
// variante.
func frontmatterDeLaSkill(t *testing.T, variante, skill string) skills.Frontmatter {
	t.Helper()

	frontmatter, err := skills.LeerFrontmatter(skillMdDeLaVariante(t, variante, skill))
	require.NoError(t, err)

	return frontmatter
}

// comprobarLaFormaDelProducto es FR-011 de H39: las skills de cada variante —la
// de laboratorio y, en la que la cambia, la de entrada— tienen la forma que
// make ci exige a las de kitlegal. Su frontmatter lleva name y description y
// nada más, y skills.Regenerar y skills.Comparar sobre el directorio de la
// variante, como TestSkillsDelRepositorio sobre la raíz, no dan ningún defecto
// —cuerpo de como mucho 11 000 caracteres, descripción de como mucho 900,
// references/ plano y enlazado— ni ninguna deriva. Sin applets declarados no hay
// tabla que generar, ni descripciones de verbos que dar (research V11 de H39).
func comprobarLaFormaDelProducto(t *testing.T, declarada declaracionDelLaboratorio) {
	t.Helper()

	for _, variante := range declarada.Variantes {
		directorio := directorioDeLaVariante(raizDelRepositorio, variante.Nombre)

		esperadas := []string{declarada.Skill}
		if slices.Contains(variante.Cambia, parteEntrada) {
			esperadas = append(esperadas, declarada.Entrada)
		}

		regenerado, err := skills.Regenerar(directorio, nil)
		require.NoError(t, err, variante.Nombre)

		var nombres []string
		for _, skill := range regenerado.Skills {
			nombres = append(nombres, skill.Nombre)
		}

		assert.ElementsMatch(t, esperadas, nombres, "las skills de la variante %s", variante.Nombre)
		assert.Empty(t, regenerado.Defectos(), "las skills de la variante %s no tienen defectos", variante.Nombre)

		derivas, err := skills.Comparar(directorio, regenerado)
		require.NoError(t, err, variante.Nombre)
		assert.Empty(t, derivas, "las skills de la variante %s no tienen deriva", variante.Nombre)

		for _, skill := range esperadas {
			assert.Equal(t, []string{"name", "description"}, frontmatterDeLaSkill(t, variante.Nombre, skill).Claves,
				"el frontmatter de %s en la variante %s lleva name y description, sin metadata", skill, variante.Nombre)
		}
	}
}

// evalDelLaboratorioEsperada es una eval del laboratorio tal como la fija
// data-model §4 de H39: su fichero, la lectura que pide y la eval de normativa
// de la que toma la pregunta, el comando y la cita (research V19 de H39).
type evalDelLaboratorioEsperada struct {
	fichero, norma, bloque, deNormativa string
}

// evalsDelLaboratorioEsperadas son las cuatro evals del laboratorio, en orden de
// fichero (data-model §4 de H39; contracts/laboratorio.md §6 de H39).
var evalsDelLaboratorioEsperadas = []evalDelLaboratorioEsperada{
	{
		fichero: "01-constitucion-articulo-140.yaml", norma: "BOE-A-1978-31229", bloque: "a140",
		deNormativa: "09-constitucion-articulo-140.yaml",
	},
	{
		fichero: "02-lgt-prescripcion.yaml", norma: "BOE-A-2003-23186", bloque: "a66",
		deNormativa: "04-lgt-prescripcion.yaml",
	},
	{
		fichero: "03-lrjsp-principio-de-legalidad.yaml", norma: "BOE-A-2015-10566", bloque: "a25",
		deNormativa: "07-lrjsp-principio-de-legalidad.yaml",
	},
	{
		fichero: "04-et-vacaciones.yaml", norma: "BOE-A-2015-11430", bloque: "a38",
		deNormativa: "10-et-vacaciones.yaml",
	},
}

// TestEvalsDelLaboratorio es el control de las evals del laboratorio (FR-012 de
// H39; contracts/laboratorio.md §6 y §7 de H39; data-model §4 de H39):
// LeerConjunto las lee bien formadas y son cuatro, cada una con la pregunta, el
// comando y la cita de su eval de normativa; cada comando es una lectura de la
// herramienta de la declaración, la que satisface una llamada a esa herramienta
// con su norma y su bloque; y la unión de grabaciones de hoy sirve todas sus
// consultas, primero al preparar la caché y después sin red. El laboratorio no
// graba nada: lo que una eval pida y la grabación no tenga es un defecto de la
// eval.
func TestEvalsDelLaboratorio(t *testing.T) {
	t.Parallel()

	evalsDelLaboratorio := directorioDeLasEvalsDelLaboratorio(raizDelRepositorio)

	conjunto, err := LeerConjunto(evalsDelLaboratorio)
	require.NoError(t, err)

	t.Run("bien formadas", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, conjunto.MalFormados, "cada fichero de %s es una eval bien formada", evalsDelLaboratorio)
		assert.Nil(t, conjunto.Juez, "%s no tiene juez", evalsDelLaboratorio)
		assert.Zero(t, conjunto.Prohibidas, "%s no tiene lista de expresiones prohibidas", evalsDelLaboratorio)
		assert.Len(t, conjunto.Evals, len(evalsDelLaboratorioEsperadas), "%s tiene las cuatro evals",
			evalsDelLaboratorio)
	})

	t.Run("las de normativa", func(t *testing.T) {
		t.Parallel()

		comprobarLasEvalsDeNormativa(t, conjunto.Evals)
	})

	t.Run("lecturas de la herramienta", func(t *testing.T) {
		t.Parallel()

		comprobarLasLecturasDeLaHerramienta(t, conjunto.Evals)
	})

	t.Run("grabado", func(t *testing.T) {
		t.Parallel()

		consultas := ConsultasNecesarias(conjunto.Evals)
		require.NotEmpty(t, consultas, "las evals de %s necesitan alguna consulta", evalsDelLaboratorio)

		dirCache := t.TempDir()

		preparadas, err := Preparar(dirCache, UnionDeGrabaciones(), consultas)
		require.NoError(t, err)
		assert.Empty(t, preparadas, "consultas de las evals de %s sin su respuesta en las grabaciones %s:\n%s",
			evalsDelLaboratorio, strings.Join(UnionDeGrabaciones(), " y "), presentarFaltas(preparadas))

		comprobadas, err := ComprobarSinRed(dirCache, consultas)
		require.NoError(t, err)
		assert.Empty(t, comprobadas, "consultas de las evals de %s que la caché preparada no sirve sin red:\n%s",
			evalsDelLaboratorio, presentarFaltas(comprobadas))
	})
}

// comprobarLasEvalsDeNormativa: las evals del laboratorio son las de
// evalsDelLaboratorioEsperadas, en su orden, y cada una tiene la pregunta, la
// activación, los comandos y las citas de su eval de normativa, con un solo
// comando, la lectura de su norma y su bloque, y una sola cita, la de ese bloque
// (data-model §4 de H39; research V19 de H39).
func comprobarLasEvalsDeNormativa(t *testing.T, evals []Eval) {
	t.Helper()

	deNormativa, err := LeerConjunto(evalsDeNormativa)
	require.NoError(t, err)

	porFichero := map[string]Eval{}
	for _, eval := range deNormativa.Evals {
		porFichero[eval.Fichero] = eval
	}

	ficheros := make([]string, 0, len(evals))
	for _, eval := range evals {
		ficheros = append(ficheros, eval.Fichero)
	}

	esperados := make([]string, 0, len(evalsDelLaboratorioEsperadas))
	for _, esperada := range evalsDelLaboratorioEsperadas {
		esperados = append(esperados, esperada.fichero)
	}

	require.Equal(t, esperados, ficheros, "las evals del laboratorio, en orden de fichero")

	for posicion, esperada := range evalsDelLaboratorioEsperadas {
		eval := evals[posicion]

		origen, existe := porFichero[esperada.deNormativa]
		require.True(t, existe, "la eval %s de %s existe", esperada.deNormativa, evalsDeNormativa)

		assert.True(t, eval.Activa, "%s es una eval de activación", eval.Fichero)
		assert.Equal(t, []ComandoEsperado{{Applet: appletDeLasNormas, Norma: esperada.norma, Bloque: esperada.bloque}},
			eval.Comandos, "%s pide la lectura de %s %s", eval.Fichero, esperada.norma, esperada.bloque)
		assert.Equal(t, []CitaEsperada{{Norma: esperada.norma, Bloque: esperada.bloque}}, eval.Citas,
			"%s cita %s %s", eval.Fichero, esperada.norma, esperada.bloque)

		assert.Equal(t, origen.Pregunta, eval.Pregunta, "%s tiene la pregunta de %s", eval.Fichero, origen.Fichero)
		assert.Equal(t, origen.Activa, eval.Activa, "%s tiene la activación de %s", eval.Fichero, origen.Fichero)
		assert.Equal(t, origen.Comandos, eval.Comandos, "%s tiene los comandos de %s", eval.Fichero, origen.Fichero)
		assert.Equal(t, origen.Citas, eval.Citas, "%s tiene las citas de %s", eval.Fichero, origen.Fichero)
	}
}

// comprobarLasLecturasDeLaHerramienta: cada comando de cada eval del laboratorio
// es una lectura de la herramienta de la declaración (data-model §2 y §4 de H39;
// FR-012 de H39): lo satisface, como lo juzga el juicio de una sesión en modo
// herramienta, la llamada a esa herramienta con la norma y el bloque del comando
// que termina sin error.
func comprobarLasLecturasDeLaHerramienta(t *testing.T, evals []Eval) {
	t.Helper()

	declarada, err := leerLaDeclaracionDelRepositorio(raizDelRepositorio)
	require.NoError(t, err)

	verbos, err := verbosDeLasHerramientas()
	require.NoError(t, err)

	verbo, anunciada := verbos[declarada.Herramienta]
	require.True(t, anunciada, "la herramienta %s la anuncia el servidor MCP de kitlegal", declarada.Herramienta)

	for _, eval := range evals {
		require.NotEmpty(t, eval.Comandos, "%s pide alguna lectura", eval.Fichero)

		for _, comando := range eval.Comandos {
			argumentos, err := json.Marshal(map[string]string{"norma": comando.Norma, "bloque": comando.Bloque})
			require.NoError(t, err)

			llamada := Llamada{Herramienta: declarada.Herramienta, Argumentos: argumentos, ConResultado: true}

			assert.True(t, satisface(invocacionDeLaLlamada(llamada, verbo).Invocacion, comando, false),
				"el comando %s de %s es una lectura de %s", textoDelComando(comando), eval.Fichero,
				declarada.Herramienta)
		}
	}
}

// TestLaboratorioFueraDelProducto comprueba que el laboratorio no es parte de
// lo que kitlegal empotra, instala ni mide (FR-010 y SC-005 de H39;
// contracts/laboratorio.md §1 y §7 de H39; research V9 y V10 de H39; FR-001 y
// SC-003 de H40; contracts/variantes.md §7 de H40; research D12 de H40; FR-023 y
// SC-008; contracts/variantes.md §2): ni lo que el binario empotra ni lo que deja
// skills install --dir en un directorio vacío tiene por tramo de una ruta un
// nombre del laboratorio —el de una skill de las variantes de H39 o el de un
// directorio de variante de la carpeta de una familia del producto— ni los bytes
// de un fichero de laboratorio/ que no sea copia de uno de skills/; la definición
// del job de evals no ejecuta ninguna skill con uno de esos nombres ni hay evals
// de kitlegal con él; y ninguna ruta de laboratorio/ tiene por tramo el nombre de
// la variante del producto de una familia, que no se versiona, ni mide, así que
// una propuesta que solo cambia el laboratorio no elige ninguna skill de
// kitlegal.
func TestLaboratorioFueraDelProducto(t *testing.T) {
	t.Parallel()

	laboratorio := leerElContenidoDelLaboratorio(t)

	t.Run("empotrado", func(t *testing.T) {
		t.Parallel()

		comprobarQueNadaEsDelLaboratorio(t, kitlegal.Skills(), laboratorio, "lo empotrado")
	})

	t.Run("instalado", func(t *testing.T) {
		t.Parallel()

		directorio := t.TempDir()

		var salida, errores bytes.Buffer

		codigo := app.Arrancar([]string{"kitlegal", "skills", "install", "--dir", directorio},
			app.RegistroDeProduccion, &salida, &errores, versionDeDesarrollo, commitDeDesarrollo, fechaDeDesarrollo)
		require.Zero(t, codigo, "skills install --dir %s:\n%s", directorio, errores.String())

		comprobarQueNadaEsDelLaboratorio(t, os.DirFS(directorio), laboratorio, "lo instalado con --dir")
	})

	t.Run("job de evals", func(t *testing.T) {
		t.Parallel()

		comprobarElJobDeEvals(t, laboratorio)
	})

	t.Run("rutas que miden", func(t *testing.T) {
		t.Parallel()

		comprobarLasRutasDelLaboratorio(t, laboratorio)
	})
}

// La versión, el commit y la fecha con los que TestLaboratorioFueraDelProducto
// arranca el binario: los de cmd/kitlegal compilado sin -ldflags. La versión,
// sin forma SemVer, es la de un binario de desarrollo, que instala y no compara,
// así que el aviso de versión no lee el disco (FR-073 de H19).
const (
	versionDeDesarrollo = "dev"
	commitDeDesarrollo  = "none"
	fechaDeDesarrollo   = "unknown"
)

// contenidoDelLaboratorio es lo que TestLaboratorioFueraDelProducto busca fuera
// del laboratorio: los nombres que no pueden ser un tramo de una ruta de lo que
// kitlegal empotra o instala, en orden y sin repetir; sus ficheros, por su ruta
// desde la raíz del repositorio; y los que no son copia de uno de skills/, por su
// contenido.
type contenidoDelLaboratorio struct {
	nombres      []string
	rutas        []string
	porContenido map[string]string
}

// leerElContenidoDelLaboratorio lee de laboratorio/ sus nombres —los de las
// skills de todas las variantes de H39, los de variantes/*/skills/*, también los
// de una variante que la declaración no tuviera, y los de los directorios de la
// carpeta de cada familia del producto, también los de una variante que la
// familia no tuviera— y todos sus ficheros (research D12 de H40; contracts/
// variantes.md §2). Exige las dos skills de la declaración de H39, la variante de
// control de cada familia del producto y algún fichero, para que ninguna
// comprobación pase en vacío.
func leerElContenidoDelLaboratorio(t *testing.T) contenidoDelLaboratorio {
	t.Helper()

	repositorio := os.DirFS(raizDelRepositorio)

	deLasVariantes, err := fs.Glob(repositorio, path.Join(carpetaDelLaboratorio, carpetaDeLasVariantes, "*",
		carpetaDeLasSkillsDeLaVariante, "*"))
	require.NoError(t, err)

	skillsDeH39 := make([]string, 0, len(deLasVariantes))
	for _, deUnaVariante := range deLasVariantes {
		skillsDeH39 = append(skillsDeH39, path.Base(deUnaVariante))
	}

	require.Subset(t, skillsDeH39, []string{skillDelLaboratorio, entradaDelLaboratorio},
		"las variantes de %s tienen las dos skills de la declaración", carpetaDelLaboratorio)

	nombres := skillsDeH39

	for _, familia := range familiasDelProducto {
		deLaFamilia, err := fs.ReadDir(repositorio, familia.carpeta)
		require.NoError(t, err)

		variantes := make([]string, 0, len(deLaFamilia))
		for _, entrada := range deLaFamilia {
			if entrada.IsDir() {
				variantes = append(variantes, entrada.Name())
			}
		}

		require.Contains(t, variantes, familia.nombres()[0], "%s tiene la variante de control", familia.carpeta)

		nombres = slices.Concat(nombres, variantes)
	}

	slices.Sort(nombres)

	laboratorio := contenidoDelLaboratorio{nombres: slices.Compact(nombres), porContenido: map[string]string{}}

	err = fs.WalkDir(repositorio, carpetaDelLaboratorio, func(ruta string, entrada fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !entrada.Type().IsRegular() {
			return nil
		}

		contenido, err := fs.ReadFile(repositorio, ruta)
		if err != nil {
			return err
		}

		laboratorio.rutas = append(laboratorio.rutas, ruta)

		copia, err := esCopiaDeUnFicheroDeSkills(repositorio, ruta, contenido)
		if err != nil {
			return err
		}

		if !copia {
			laboratorio.porContenido[string(contenido)] = ruta
		}

		return nil
	})
	require.NoError(t, err)
	require.NotEmpty(t, laboratorio.rutas, "%s tiene algún fichero", carpetaDelLaboratorio)

	return laboratorio
}

// esCopiaDeUnFicheroDeSkills dice si el fichero de laboratorio/ de esa ruta es
// copia de uno de skills/: su ruta tiene un tramo …/skills/<s>/<r> y
// skills/<s>/<r> es un fichero con sus mismos bytes (research D12 de H40). Nada
// distingue la copia de su original, así que sus bytes no dicen que un fichero
// que kitlegal empotra o instala sea del laboratorio.
func esCopiaDeUnFicheroDeSkills(repositorio fs.FS, ruta string, contenido []byte) (bool, error) {
	tramos := strings.Split(ruta, "/")

	for i := range len(tramos) - 2 {
		if tramos[i] != carpetaDeLasSkillsDeLaVariante {
			continue
		}

		original := path.Join(append([]string{directorioDeLasSkills}, tramos[i+1:]...)...)

		delOriginal, err := fs.Stat(repositorio, original)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}

		if err != nil {
			return false, err
		}

		if !delOriginal.Mode().IsRegular() {
			continue
		}

		delProducto, err := fs.ReadFile(repositorio, original)
		if err != nil {
			return false, err
		}

		if bytes.Equal(delProducto, contenido) {
			return true, nil
		}
	}

	return false, nil
}

// comprobarQueNadaEsDelLaboratorio: ninguna ruta del árbol tiene por tramo un
// nombre del laboratorio, ni ningún fichero suyo tiene los bytes de un fichero
// del laboratorio que no sea copia de uno de skills/. Exige algún fichero, para
// no pasar en vacío.
func comprobarQueNadaEsDelLaboratorio(t *testing.T, arbol fs.FS, laboratorio contenidoDelLaboratorio,
	donde string,
) {
	t.Helper()

	var ficheros int

	err := fs.WalkDir(arbol, ".", func(ruta string, entrada fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		for _, componente := range strings.Split(ruta, "/") {
			assert.NotContains(t, laboratorio.nombres, componente,
				"%s: %s tiene por tramo un nombre del laboratorio", donde, ruta)
		}

		if !entrada.Type().IsRegular() {
			return nil
		}

		contenido, err := fs.ReadFile(arbol, ruta)
		if err != nil {
			return err
		}

		ficheros++

		delLaboratorio, igual := laboratorio.porContenido[string(contenido)]
		assert.False(t, igual, "%s: %s tiene los bytes de %s", donde, ruta, delLaboratorio)

		return nil
	})
	require.NoError(t, err)
	require.Positive(t, ficheros, "%s tiene algún fichero", donde)
}

// comprobarElJobDeEvals: la definición del job de evals no ejecuta ninguna skill
// con un nombre del laboratorio —no está en la matriz ni en el include del
// trabajo evals ni en los del trabajo de la medida del juez— y evals/ no tiene
// ninguna carpeta con él (FR-010 de H39; research D4 de H39; research D12 de
// H40).
func comprobarElJobDeEvals(t *testing.T, laboratorio contenidoDelLaboratorio) {
	t.Helper()

	job, err := leerDefinicionDelJob(rutaDeLaDefinicionDelJob)
	require.NoError(t, err)

	ejecutadas := skillsQueEjecutaElJob(job)
	require.NotEmpty(t, ejecutadas, "el job de evals ejecuta alguna skill")

	entradas, err := fs.ReadDir(os.DirFS(raizDelRepositorio), directorioDeLasEvals)
	require.NoError(t, err)

	carpetas := make([]string, 0, len(entradas))
	for _, entrada := range entradas {
		carpetas = append(carpetas, entrada.Name())
	}

	require.NotEmpty(t, carpetas, "%s/ tiene las evals de kitlegal", directorioDeLasEvals)

	for _, nombre := range laboratorio.nombres {
		assert.NotContains(t, ejecutadas, nombre, "el job de evals ejecuta la skill %s, un nombre del laboratorio",
			nombre)
		assert.NotContains(t, carpetas, nombre, "%s/ tiene evals de %s, un nombre del laboratorio",
			directorioDeLasEvals, nombre)
	}
}

// skillsQueEjecutaElJob son las skills que nombra la definición del job de
// evals: las de la matriz y el include del trabajo evals y, si lo tiene, las del
// trabajo de la medida del juez.
func skillsQueEjecutaElJob(job DefinicionDelJob) []string {
	ejecutadas := slices.Concat(job.Skills, slices.Collect(maps.Keys(job.PorSkill)))

	if job.Medida != nil {
		ejecutadas = slices.Concat(ejecutadas, job.Medida.Skills,
			slices.Collect(maps.Keys(job.Medida.ConcurrenciaPorSkill)))
	}

	return ejecutadas
}

// comprobarLasRutasDelLaboratorio: ninguna ruta de laboratorio/ tiene por tramo
// el nombre de la variante del producto de una familia, que es skills/<skill> y
// no se versiona (FR-022; contracts/variantes.md §2), ni mide (rutaQueMide;
// research V9 de H39). Sin paquetes, porque los de los conjuntos solo quitan de lo
// que mide los tests y los fixtures de su carpeta: la ruta que no mide sin ninguno
// no mide con ninguno.
func comprobarLasRutasDelLaboratorio(t *testing.T, laboratorio contenidoDelLaboratorio) {
	t.Helper()

	sinVersionar := make([]string, 0, len(familiasDelProducto))
	for _, familia := range familiasDelProducto {
		sinVersionar = append(sinVersionar, familia.producto)
	}

	for _, ruta := range laboratorio.rutas {
		for _, tramo := range strings.Split(ruta, "/") {
			assert.NotContains(t, sinVersionar, tramo,
				"%s tiene por tramo la variante del producto %s, que no se versiona", ruta, tramo)
		}

		assert.False(t, rutaQueMide(ruta, nil), "%s mide: un cambio suyo elegiría skills de kitlegal", ruta)
	}
}
