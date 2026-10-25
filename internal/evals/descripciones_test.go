package evals

import (
	"encoding/json/v2"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/skills"
)

// La evidencia del juez de las descripciones (contracts/juez-de-las-descripciones.md
// §1 de H31; research D2 de H31): su carpeta desde la raíz del repositorio, con
// la barra de separador, como la nombran las líneas del control de identidad;
// su manifiesto; y la única clase del juez, que decide.
const (
	rutaDeLaEvidenciaDeLasDescripciones = "evidencias/adr-0046"
	manifiestoDeLaEvidenciaEnLosTests   = "manifiesto.json"
	claseHablaEnPrimeraOSegundaPersona  = "habla_en_primera_o_segunda_persona"
)

// plantillaDeMensajeDe es el mensaje del voto de una descripción construido
// como lo construye mensaje_de de evidencias/adr-0046/guiones/juez.py, con la
// skill y la descripción sin blancos a los lados: lo que el mensaje del job
// tiene que ser, byte a byte (FR-027 de H31; research D5 de H31).
const plantillaDeMensajeDe = "<descripcion skill=\"%s\">\n%s\n</descripcion>\n\n" +
	"Responde a la pregunta de la rúbrica sobre esta descripción."

// casoDelMensaje es lo que el test del mensaje lee de un caso de casos.yaml de
// la evidencia: su id, su skill y su descripción.
type casoDelMensaje struct {
	ID          string `yaml:"id"`
	Skill       string `yaml:"skill"`
	Descripcion string `yaml:"descripcion"`
}

// casosDelMensaje es el documento casos.yaml de la evidencia, con lo que el test
// del mensaje lee de cada caso.
type casosDelMensaje struct {
	Casos []casoDelMensaje `yaml:"casos"`
}

// carpetaDeLaEvidenciaEn es la carpeta de la evidencia del juez de las
// descripciones en esa raíz.
func carpetaDeLaEvidenciaEn(raiz string) string {
	return filepath.Join(raiz, filepath.FromSlash(rutaDeLaEvidenciaDeLasDescripciones))
}

// copiarLaEvidenciaDeLasDescripciones copia en un directorio temporal del test,
// en su carpeta de la evidencia, los cuatro ficheros del juez de las
// descripciones y el manifiesto de la evidencia, y devuelve ese directorio, que
// es la raíz de la copia. La evidencia del repositorio solo se lee.
func copiarLaEvidenciaDeLasDescripciones(t *testing.T) string {
	t.Helper()

	raiz := t.TempDir()
	carpeta := carpetaDeLaEvidenciaEn(raiz)
	require.NoError(t, os.MkdirAll(carpeta, 0o750))

	for _, fichero := range slices.Concat(ficherosCopiadosDelJuez, []string{manifiestoDeLaEvidenciaEnLosTests}) {
		contenido := contenidoDelFichero(t, filepath.Join(carpetaDeLaEvidenciaEn(raizDelRepositorio), fichero))
		require.NoError(t, os.WriteFile(filepath.Join(carpeta, fichero), contenido, 0o600))
	}

	return raiz
}

// juezDeLasDescripcionesEn es el juez de las descripciones de la evidencia de
// esa raíz, que tiene que poder leerse.
func juezDeLasDescripcionesEn(t *testing.T, raiz string) *Juez {
	t.Helper()

	juez, err := juezDeLasDescripciones(raiz)
	require.NoError(t, err, "el juez de las descripciones de %s", raiz)

	return juez
}

// identidadEn es lo que da el control de identidad sobre la evidencia de esa
// raíz, que tiene que poder leerse.
func identidadEn(t *testing.T, raiz string) []string {
	t.Helper()

	lineas, err := identidadDelJuezDeLasDescripciones(raiz)
	require.NoError(t, err, "el control de identidad de %s", raiz)

	return lineas
}

// TestJuezDeLasDescripciones fija el juez de las descripciones de FR-025 a
// FR-027 de H31 (contracts/juez-de-las-descripciones.md §1, §2 y §4 de H31;
// research D2, D4 y D5 de H31; SC-006):
//
//   - el juez lee, sin copia, la rúbrica, el esquema, los casos y la medida de
//     la evidencia del ADR 0046, con una sola clase, que decide, y no se lee
//     sin alguno de ellos ni con un esquema de otra clase;
//   - el control de identidad, el de make ci, no da nada sobre el repositorio
//     y, sobre una copia, da la línea de cada uno de los cuatro con un byte
//     cambiado y la del que el manifiesto no registra;
//   - el mensaje de cada voto es el de mensaje_de, byte a byte, para los casos
//     de casos.yaml, y quita de la descripción los blancos de los extremos que
//     quita str.strip();
//   - la huella del juicio es la del texto del contrato y cambia con cada uno
//     de sus datos;
//   - la descripción se vota con la regla de la votación del juez de las evals
//     y su clase única (FR-028, FR-029 y FR-032 de H31;
//     contracts/juez-de-las-descripciones.md §5 de H31; research D6 de H31;
//     SC-007): marcada con tres «si» unánimes con su frase en la descripción,
//     pedidos por orden; un «si» con una frase que no está, nulo y repetido una
//     vez; el voto cortado por el tope, repetido una vez, y cortado otra vez,
//     sin juzgar; la sesión con error, sin la forma del esquema o cortada por
//     un límite de uso, sin repetición y sin juzgar; como mucho seis votos; y
//     la frase y el motivo del primer voto que cuenta;
//   - y el votante del guion de las evals, con el claude sustituto, recibe el
//     mensaje de la descripción por su entrada estándar, con la rúbrica y el
//     esquema de la evidencia.
//
// Ninguno abre una sesión con modelo (FR-033; SC-014).
func TestJuezDeLasDescripciones(t *testing.T) {
	t.Parallel()

	t.Run("juez", func(t *testing.T) {
		t.Parallel()

		probarElJuezDeLasDescripciones(t)
	})

	t.Run("identidad", func(t *testing.T) {
		t.Parallel()

		probarLaIdentidadDelJuezDeLasDescripciones(t)
	})

	t.Run("mensaje", func(t *testing.T) {
		t.Parallel()

		probarElMensajeDeLaDescripcion(t)
	})

	t.Run("huella", func(t *testing.T) {
		t.Parallel()

		probarLaHuellaDelJuicio(t)
	})

	t.Run("votos", func(t *testing.T) {
		t.Parallel()

		probarLosVotosDeLaDescripcion(t)
	})

	t.Run("votante-del-guion", func(t *testing.T) {
		t.Parallel()

		probarLaDescripcionConElVotanteDelGuion(t)
	})
}

// probarElJuezDeLasDescripciones fija lo que el juez lee de la evidencia: la
// del repositorio da su clase y el contenido entero de los cuatro ficheros; sin
// cualquiera de ellos, o con un esquema que no es el de su clase, no hay juez,
// y el error nombra el fichero.
func probarElJuezDeLasDescripciones(t *testing.T) {
	t.Helper()

	t.Run("del-repositorio", func(t *testing.T) {
		t.Parallel()

		juez := juezDeLasDescripcionesEn(t, raizDelRepositorio)
		evidencia := carpetaDeLaEvidenciaEn(raizDelRepositorio)

		assert.Equal(t, []ClaseDelJuez{{Nombre: claseHablaEnPrimeraOSegundaPersona, Decide: true}}, juez.Clases)
		assert.Equal(t, []string{
			string(contenidoDelFichero(t, filepath.Join(evidencia, ficheroDeRubricaDelJuez))),
			string(contenidoDelFichero(t, filepath.Join(evidencia, ficheroDeEsquemaDelJuez))),
			string(contenidoDelFichero(t, filepath.Join(evidencia, ficheroDeCasosDelJuez))),
			string(contenidoDelFichero(t, filepath.Join(evidencia, ficheroDeMedidaDelJuez))),
		}, []string{juez.Rubrica, juez.Esquema, juez.Casos, juez.Medida})
	})

	for _, fichero := range ficherosCopiadosDelJuez {
		t.Run("sin-"+fichero, func(t *testing.T) {
			t.Parallel()

			raiz := copiarLaEvidenciaDeLasDescripciones(t)
			require.NoError(t, os.Remove(filepath.Join(carpetaDeLaEvidenciaEn(raiz), fichero)))

			juez, err := juezDeLasDescripciones(raiz)
			assert.Nil(t, juez)
			assert.ErrorContains(t, err, rutaDeLaEvidenciaDeLasDescripciones+"/"+fichero)
		})
	}

	esquemas := []struct{ nombre, contenido string }{
		{"esquema-de-otra-clase", `{"type":"object","properties":{"otra_clase":{"type":"object"}}}`},
		{"esquema-con-otra-clase-mas", `{"type":"object","properties":{"` + claseHablaEnPrimeraOSegundaPersona +
			`":{"type":"object"},"otra_clase":{"type":"object"}}}`},
		{"esquema-que-no-es-json", "{\n"},
	}

	for _, esquema := range esquemas {
		t.Run(esquema.nombre, func(t *testing.T) {
			t.Parallel()

			raiz := copiarLaEvidenciaDeLasDescripciones(t)
			escribirEnLaCarpeta(ficheroDeEsquemaDelJuez, esquema.contenido)(t, carpetaDeLaEvidenciaEn(raiz))

			juez, err := juezDeLasDescripciones(raiz)
			assert.Nil(t, juez)
			assert.ErrorContains(t, err, rutaDeLaEvidenciaDeLasDescripciones+"/"+ficheroDeEsquemaDelJuez)
		})
	}
}

// probarLaIdentidadDelJuezDeLasDescripciones es el control de identidad de
// FR-025 y SC-006 de H31: los cuatro ficheros de la evidencia del repositorio
// son los que registra su manifiesto, y sobre una copia lo ve fallar con un
// byte cambiado en cada uno de los cuatro, cuatro de cuatro, y con cada uno que
// el manifiesto no registra; sin manifiesto no hay identidad que comprobar.
func probarLaIdentidadDelJuezDeLasDescripciones(t *testing.T) {
	t.Helper()

	t.Run("del-repositorio", func(t *testing.T) {
		t.Parallel()

		lineas := identidadEn(t, raizDelRepositorio)
		assert.Empty(t, lineas, "ficheros de %s que no son los que registra su manifiesto:\n%s",
			rutaDeLaEvidenciaDeLasDescripciones, strings.Join(lineas, "\n"))
	})

	for _, fichero := range ficherosCopiadosDelJuez {
		t.Run("cambiado-"+fichero, func(t *testing.T) {
			t.Parallel()

			raiz := copiarLaEvidenciaDeLasDescripciones(t)
			require.Empty(t, identidadEn(t, raiz), "premisa: la copia sin tocar es la del manifiesto")

			ruta := filepath.Join(carpetaDeLaEvidenciaEn(raiz), fichero)
			registrada := sha256DeTexto(string(contenidoDelFichero(t, ruta)))

			cambiarUnByteDe(fichero)(t, carpetaDeLaEvidenciaEn(raiz))

			assert.Equal(t, []string{
				rutaDeLaEvidenciaDeLasDescripciones + "/" + fichero + ": su SHA-256 es " +
					sha256DeTexto(string(contenidoDelFichero(t, ruta))) +
					" y el manifiesto de la evidencia registra " + registrada,
			}, identidadEn(t, raiz))
		})

		t.Run("sin-registrar-"+fichero, func(t *testing.T) {
			t.Parallel()

			raiz := copiarLaEvidenciaDeLasDescripciones(t)
			quitarDelManifiesto(t, raiz, fichero)

			assert.Equal(t, []string{
				rutaDeLaEvidenciaDeLasDescripciones + "/" + fichero + ": el manifiesto de la evidencia no lo registra",
			}, identidadEn(t, raiz))
		})
	}

	t.Run("sin-manifiesto", func(t *testing.T) {
		t.Parallel()

		raiz := copiarLaEvidenciaDeLasDescripciones(t)
		require.NoError(t, os.Remove(filepath.Join(carpetaDeLaEvidenciaEn(raiz), manifiestoDeLaEvidenciaEnLosTests)))

		lineas, err := identidadDelJuezDeLasDescripciones(raiz)
		assert.Nil(t, lineas)
		assert.ErrorContains(t, err, rutaDeLaEvidenciaDeLasDescripciones+"/"+manifiestoDeLaEvidenciaEnLosTests)
	})
}

// quitarDelManifiesto quita del manifiesto de la evidencia de la copia de esa
// raíz la entrada de ese fichero, y deja las demás como estaban.
func quitarDelManifiesto(t *testing.T, raiz, fichero string) {
	t.Helper()

	ruta := filepath.Join(carpetaDeLaEvidenciaEn(raiz), manifiestoDeLaEvidenciaEnLosTests)

	var manifiesto struct {
		Ficheros []map[string]any `json:"ficheros"`
	}

	require.NoError(t, json.Unmarshal(contenidoDelFichero(t, ruta), &manifiesto))

	antes := len(manifiesto.Ficheros)
	manifiesto.Ficheros = slices.DeleteFunc(manifiesto.Ficheros, func(entrada map[string]any) bool {
		return entrada["fichero"] == fichero
	})
	require.Len(t, manifiesto.Ficheros, antes-1, "premisa: el manifiesto registra %s una vez", fichero)

	codificado, err := json.Marshal(manifiesto)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(ruta, codificado, 0o600))
}

// probarElMensajeDeLaDescripcion fija el mensaje del voto de FR-027 de H31: para
// cada caso de casos.yaml de la evidencia, el de mensaje_de, byte a byte, sin
// salto de línea al final; y, de la descripción, solo se quitan los blancos de
// los extremos que quita str.strip() —también U+001C a U+001F, que no quita
// strings.TrimSpace (research D5 de H31)—, no los de dentro.
func probarElMensajeDeLaDescripcion(t *testing.T) {
	t.Helper()

	ruta := filepath.Join(carpetaDeLaEvidenciaEn(raizDelRepositorio), ficheroDeCasosDelJuez)

	casos, err := skills.ValidarDocumentoYAML[casosDelMensaje](contenidoDelFichero(t, ruta), nil)
	require.NoError(t, err)
	require.NotEmpty(t, casos.Casos, "%s tiene casos", ruta)

	for _, caso := range casos.Casos {
		require.Equal(t, strings.TrimSpace(caso.Descripcion), caso.Descripcion,
			"premisa: la descripción del caso %s no tiene blancos a los lados", caso.ID)

		assert.Equal(t, fmt.Sprintf(plantillaDeMensajeDe, caso.Skill, caso.Descripcion),
			mensajeDeLaDescripcion(caso.Skill, caso.Descripcion), "el mensaje del caso %s", caso.ID)
	}

	assert.Equal(t,
		"<descripcion skill=\"legal-core\">\nPunto de partida…\n\n  Segunda línea.\n</descripcion>\n\n"+
			"Responde a la pregunta de la rúbrica sobre esta descripción.",
		mensajeDeLaDescripcion("legal-core", "\x1c \xc2\xa0\tPunto de partida…\n\n  Segunda línea.\n\x1f\r\n"))
}

// datosDeLaHuella son los datos de la huella del juicio de una descripción: la
// skill, su descripción, el juez y el modelo y la versión de Claude Code de sus
// votos.
type datosDeLaHuella struct {
	skill       string
	descripcion string
	juez        Juez
	modelo      string
	version     string
}

// huella es la huella del juicio con estos datos.
func (d datosDeLaHuella) huella() string {
	return huellaDelJuicio(d.skill, d.descripcion, &d.juez, d.modelo, d.version)
}

// probarLaHuellaDelJuicio fija la huella del juicio de contracts/juez-de-las-
// descripciones.md §2 de H31 (research D4 de H31): el SHA-256 del texto de sus
// siete líneas, en su orden y cada una terminada en salto de línea; y distinta
// con cada uno de sus ocho datos cambiado, de la de todos sin cambiar y de la
// de cualquier otro cambiado.
func probarLaHuellaDelJuicio(t *testing.T) {
	t.Helper()

	base := datosDeLaHuella{
		skill:       "legal-core",
		descripcion: "Punto de partida de las preguntas de derecho público español.",
		juez:        Juez{Rubrica: "la rúbrica", Esquema: "el esquema", Casos: "los casos", Medida: "la medida"},
		modelo:      "claude-opus-5-5",
		version:     "2.1.289",
	}

	texto := "skill legal-core\n" +
		"descripcion " + sha256DeTexto(base.descripcion) + "\n" +
		"rubrica " + sha256DeTexto(base.juez.Rubrica) + "\n" +
		"esquema " + sha256DeTexto(base.juez.Esquema) + "\n" +
		"casos " + sha256DeTexto(base.juez.Casos) + "\n" +
		"medida " + sha256DeTexto(base.juez.Medida) + "\n" +
		"juez claude-opus-5-5 2.1.289\n"

	assert.Equal(t, sha256DeTexto(texto), base.huella())

	cambios := []struct {
		dato   string
		cambio func(d *datosDeLaHuella)
	}{
		{"skill", func(d *datosDeLaHuella) { d.skill = "boe-legislacion" }},
		{"descripcion", func(d *datosDeLaHuella) { d.descripcion += " " }},
		{"rubrica", func(d *datosDeLaHuella) { d.juez.Rubrica += "." }},
		{"esquema", func(d *datosDeLaHuella) { d.juez.Esquema += "." }},
		{"casos", func(d *datosDeLaHuella) { d.juez.Casos += "." }},
		{"medida", func(d *datosDeLaHuella) { d.juez.Medida += "." }},
		{"modelo", func(d *datosDeLaHuella) { d.modelo = "claude-fable-5-1" }},
		{"version", func(d *datosDeLaHuella) { d.version = "2.1.290" }},
	}

	vistas := map[string]string{base.huella(): "ninguno"}

	for _, cambio := range cambios {
		cambiados := base
		cambio.cambio(&cambiados)

		huella := cambiados.huella()
		if anterior, vista := vistas[huella]; vista {
			t.Errorf("la huella con %s cambiado es la misma que con %s cambiado", cambio.dato, anterior)
		}

		vistas[huella] = cambio.dato
	}
}

// La descripción que juzgan los tests de los votos y lo que dicen de ella los
// votos grabados: tres frases que están en ella, una que no está, otra que solo
// está en el mensaje del voto, el motivo de cada voto y el texto del error de
// una sesión del juez. Cada voto que dice sí lleva su frase y su motivo, para
// que el juicio diga de qué voto son los suyos.
const (
	skillDeLosVotos       = "boe-legislacion"
	descripcionDeLosVotos = "Consulta y cita normativa consolidada del BOE.\n" +
		"Úsala siempre que la respuesta dependa de lo que dice una norma: sin leerla, no hay cita."

	fraseDelPrimerVoto            = "Úsala siempre que la respuesta dependa"
	fraseDelSegundoVoto           = "Úsala   siempre"
	fraseDelTercerVoto            = "*Úsala* siempre que la respuesta dependa de lo que dice una norma"
	fraseQueNoEstaEnLaDescripcion = "Úsame cuando lo necesites"
	fraseDelMensajeDelVoto        = "Responde a la pregunta de la rúbrica"

	motivoDelPrimerVoto  = "Da una orden a quien la lee con un imperativo con un pronombre pegado."
	motivoDelSegundoVoto = "Se dirige a quien la lee con un imperativo."
	motivoDelTercerVoto  = "Le dice a quien la lee cuándo usarla."
	motivoDelVotoNulo    = "Habla en primera persona."
	motivoDelVotoQueNo   = "Describe la skill en tercera persona y no se dirige a nadie."

	errorDeLaSesionDelJuez = "API Error: 500 Internal server error"
)

// votoDeLaDescripcion es la grabación de un voto del juez de las descripciones
// que llega a darse, con lo que dice de su única clase.
func votoDeLaDescripcion(t *testing.T, respuesta, frase, motivo string) grabacion {
	t.Helper()

	return votoGrabado(t, map[string]dicho{
		claseHablaEnPrimeraOSegundaPersona: {motivo: motivo, respuesta: respuesta, frase: frase},
	})
}

// sesionDelJuezConError es la grabación del voto cuya sesión del juez terminó
// con error, con ese texto en result, el de un límite de uso entre ellos: la
// salida que deja la sesión y el error del proceso, que sale con 1.
func sesionDelJuezConError(t *testing.T, texto string) grabacion {
	t.Helper()

	return grabacion{
		salida: `{"type":"result","subtype":"success","is_error":true,"duration_ms":412,"num_turns":1,"result":` +
			cadenaJSON(t, texto) + `,"session_id":"00000000-0000-4000-8000-000000000103"}` + "\n",
		err: errorDeProceso{codigo: 1, texto: "exit status 1"},
	}
}

// votoDado es lo que el juicio de una descripción guarda de un voto que se dio.
func votoDado(respuesta, frase, motivo string) VotoDeLaDescripcion {
	return VotoDeLaDescripcion{Respuesta: respuesta, Frase: frase, Motivo: motivo}
}

// casoDeVotosDeLaDescripcion es un caso de los votos de una descripción: lo que
// el votante simulado devuelve de cada voto, por orden, que son todos los que
// se le tienen que pedir, y el juicio que queda.
type casoDeVotosDeLaDescripcion struct {
	nombre      string
	grabaciones []grabacion
	juicio      JuicioDeLaDescripcion
}

// casosDeVotosDeLaDescripcion son los casos de los votos de una descripción,
// los de la tabla de contracts/juez-de-las-descripciones.md §5 de H31.
func casosDeVotosDeLaDescripcion(t *testing.T) []casoDeVotosDeLaDescripcion {
	t.Helper()

	primero := votoDeLaDescripcion(t, "si", fraseDelPrimerVoto, motivoDelPrimerVoto)
	segundo := votoDeLaDescripcion(t, "si", fraseDelSegundoVoto, motivoDelSegundoVoto)
	tercero := votoDeLaDescripcion(t, "si", fraseDelTercerVoto, motivoDelTercerVoto)
	nulo := votoDeLaDescripcion(t, "si", fraseQueNoEstaEnLaDescripcion, motivoDelVotoNulo)
	delMensaje := votoDeLaDescripcion(t, "si", fraseDelMensajeDelVoto, motivoDelVotoNulo)
	queNo := votoDeLaDescripcion(t, "no", "", motivoDelVotoQueNo)
	cortado := grabacion{err: fmt.Errorf("el voto se ha cortado antes de terminar: %w", errTopeDelVoto)}

	sinForma := juicioGrabado(t, map[string]dicho{
		claseHablaEnPrimeraOSegundaPersona: {motivo: motivoDelVotoQueNo, respuesta: "quizá"},
	})

	dadoPrimero := votoDado("si", fraseDelPrimerVoto, motivoDelPrimerVoto)
	dadoSegundo := votoDado("si", fraseDelSegundoVoto, motivoDelSegundoVoto)
	dadoTercero := votoDado("si", fraseDelTercerVoto, motivoDelTercerVoto)
	dadoNulo := votoDado("si", fraseQueNoEstaEnLaDescripcion, motivoDelVotoNulo)
	dadoDelMensaje := votoDado("si", fraseDelMensajeDelVoto, motivoDelVotoNulo)
	dadoQueNo := votoDado("no", "", motivoDelVotoQueNo)

	marcada := JuicioDeLaDescripcion{
		Marcada: true, Frase: fraseDelPrimerVoto, Motivo: motivoDelPrimerVoto,
		Votos: []VotoDeLaDescripcion{dadoPrimero, dadoSegundo, dadoTercero},
	}

	return []casoDeVotosDeLaDescripcion{
		{nombre: "tres-si", grabaciones: []grabacion{primero, segundo, tercero}, juicio: marcada},
		{
			nombre:      "si-si-no",
			grabaciones: []grabacion{primero, segundo, queNo},
			juicio: JuicioDeLaDescripcion{
				Motivo: motivoDelPrimerVoto, Votos: []VotoDeLaDescripcion{dadoPrimero, dadoSegundo, dadoQueNo},
			},
		},
		{
			nombre:      "primer-no",
			grabaciones: []grabacion{queNo},
			juicio:      JuicioDeLaDescripcion{Motivo: motivoDelVotoQueNo, Votos: []VotoDeLaDescripcion{dadoQueNo}},
		},
		{
			nombre:      "frase-que-no-esta-dos-veces",
			grabaciones: []grabacion{nulo, nulo},
			juicio:      JuicioDeLaDescripcion{Motivo: motivoDelVotoNulo, Votos: []VotoDeLaDescripcion{dadoNulo, dadoNulo}},
		},
		{
			nombre:      "frase-del-mensaje-dos-veces",
			grabaciones: []grabacion{delMensaje, delMensaje},
			juicio: JuicioDeLaDescripcion{
				Motivo: motivoDelVotoNulo, Votos: []VotoDeLaDescripcion{dadoDelMensaje, dadoDelMensaje},
			},
		},
		{
			nombre:      "frase-que-no-esta-y-tres-si",
			grabaciones: []grabacion{nulo, primero, segundo, tercero},
			juicio: JuicioDeLaDescripcion{
				Marcada: true, Frase: fraseDelPrimerVoto, Motivo: motivoDelPrimerVoto,
				Votos: []VotoDeLaDescripcion{dadoNulo, dadoPrimero, dadoSegundo, dadoTercero},
			},
		},
		{nombre: "cortado-y-tres-si", grabaciones: []grabacion{cortado, primero, segundo, tercero}, juicio: marcada},
		{
			nombre:      "cortado-dos-veces",
			grabaciones: []grabacion{cortado, cortado},
			juicio:      JuicioDeLaDescripcion{Votos: []VotoDeLaDescripcion{}, SinJuzgar: motivoDelTopeDosVecesUno},
		},
		{
			nombre:      "cortado-dos-veces-en-el-segundo",
			grabaciones: []grabacion{primero, cortado, cortado},
			juicio: JuicioDeLaDescripcion{
				Motivo: motivoDelPrimerVoto, Votos: []VotoDeLaDescripcion{dadoPrimero}, SinJuzgar: motivoDelTopeDosVecesDos,
			},
		},
		{
			nombre:      "sesion-con-error",
			grabaciones: []grabacion{sesionDelJuezConError(t, errorDeLaSesionDelJuez)},
			juicio: JuicioDeLaDescripcion{
				Votos: []VotoDeLaDescripcion{}, SinJuzgar: prefijoDeSesionConError + errorDeLaSesionDelJuez,
			},
		},
		{
			nombre:      "limite-de-uso",
			grabaciones: []grabacion{sesionDelJuezConError(t, limiteDeUsoDeLaSesion)},
			juicio: JuicioDeLaDescripcion{
				Votos: []VotoDeLaDescripcion{}, SinJuzgar: prefijoDeSesionConError + limiteDeUsoDeLaSesion,
			},
		},
		{
			nombre:      "sin-la-forma-del-esquema",
			grabaciones: []grabacion{{salida: cabeceraDeLaSalidaDelJuez + sinForma + pieDeLaSalidaDelJuez}},
			juicio:      JuicioDeLaDescripcion{Votos: []VotoDeLaDescripcion{}, SinJuzgar: prefijoDeSalidaSinForma + sinForma},
		},
		{
			nombre:      "sesion-con-error-en-el-segundo",
			grabaciones: []grabacion{primero, sesionDelJuezConError(t, errorDeLaSesionDelJuez)},
			juicio: JuicioDeLaDescripcion{
				Motivo: motivoDelPrimerVoto, Votos: []VotoDeLaDescripcion{dadoPrimero},
				SinJuzgar: "voto 2: la sesión del juez terminó con error: " + errorDeLaSesionDelJuez,
			},
		},
		{
			nombre:      "peor-caso",
			grabaciones: []grabacion{nulo, primero, cortado, segundo, nulo, tercero},
			juicio: JuicioDeLaDescripcion{
				Marcada: true, Frase: fraseDelPrimerVoto, Motivo: motivoDelPrimerVoto,
				Votos: []VotoDeLaDescripcion{dadoNulo, dadoPrimero, dadoSegundo, dadoNulo, dadoTercero},
			},
		},
	}
}

// probarLosVotosDeLaDescripcion vota la descripción de cada caso con la
// votación del juez de las descripciones del repositorio y un votante simulado,
// y exige su juicio y los votos que pide: los de sus grabaciones, ni uno más ni
// uno menos, que nunca son más de seis, cada uno con el mensaje de la
// descripción. Y fija las claves del juicio. Sin abrir ninguna sesión.
func probarLosVotosDeLaDescripcion(t *testing.T) {
	t.Helper()

	juez := juezDeLasDescripcionesEn(t, raizDelRepositorio)
	mensaje := mensajeDeLaDescripcion(skillDeLosVotos, descripcionDeLosVotos)

	for _, caso := range casosDeVotosDeLaDescripcion(t) {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			votante := &votanteGrabado{t: t, grabaciones: caso.grabaciones}

			votacion, err := nuevaVotacion(juez, votante.votar)
			require.NoError(t, err)

			assert.Equal(t, caso.juicio, votacion.juzgarLaDescripcion(skillDeLosVotos, descripcionDeLosVotos))
			assert.Len(t, votante.mensajes, len(caso.grabaciones), "los votos pedidos, ni uno más ni uno menos")
			assert.LessOrEqual(t, len(votante.mensajes), votosPorRespuestaComoMucho, "como mucho seis votos")

			for _, pedido := range votante.mensajes {
				assert.Equal(t, mensaje, pedido, "cada voto se pide con el mensaje de la descripción")
			}
		})
	}

	t.Run("forma", func(t *testing.T) {
		t.Parallel()

		marcada, err := json.Marshal(JuicioDeLaDescripcion{
			Marcada: true, Frase: fraseDelPrimerVoto, Motivo: motivoDelPrimerVoto,
			Votos: []VotoDeLaDescripcion{votoDado("si", fraseDelPrimerVoto, motivoDelPrimerVoto)},
		})
		require.NoError(t, err)
		assert.JSONEq(t, `{"marcada":true,"frase":"`+fraseDelPrimerVoto+`","motivo":"`+motivoDelPrimerVoto+`",`+
			`"votos":[{"respuesta":"si","frase":"`+fraseDelPrimerVoto+`","motivo":"`+motivoDelPrimerVoto+`"}]}`,
			string(marcada), "las claves del juicio de una descripción, las de data-model §4 de H31")

		sinJuzgar, err := json.Marshal(JuicioDeLaDescripcion{SinJuzgar: motivoDelTopeDosVecesUno})
		require.NoError(t, err)
		assert.JSONEq(t, `{"marcada":false,"frase":"","motivo":"","votos":[],"sin_juzgar":"`+motivoDelTopeDosVecesUno+`"}`,
			string(sinJuzgar), "sin_juzgar solo en la descripción que quedó sin juzgar, y votos siempre es una lista")
	})
}

// probarLaDescripcionConElVotanteDelGuion vota una descripción con la votación
// del juez de las descripciones del repositorio y el votante del guion de voto
// de las evals, con el sustituto de claude del juez delante en el PATH, que da
// tres «si» con una frase que está en ella: la descripción queda marcada, con un
// proceso del sustituto por voto, cada uno con el mensaje de la descripción por
// su entrada estándar, byte a byte, y con la rúbrica y el esquema de la
// evidencia del ADR 0046 en sus argumentos. Ninguna sesión con modelo.
func probarLaDescripcionConElVotanteDelGuion(t *testing.T) {
	t.Helper()

	grabada := votoDeLaDescripcion(t, "si", fraseDelPrimerVoto, motivoDelPrimerVoto).salida
	sustituto := escribirElClaudeDelJuez(t, map[string]string{salidaDelClaudeDelJuez: grabada})

	juez := juezDeLasDescripcionesEn(t, raizDelRepositorio)

	votar, retirar, err := nuevoVotanteDelGuion(t.Context(), ordenDelVotoConElSustituto(t, sustituto, juez))
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, retirar()) })

	votacion, err := nuevaVotacion(juez, votar)
	require.NoError(t, err)

	dado := votoDado("si", fraseDelPrimerVoto, motivoDelPrimerVoto)
	assert.Equal(t, JuicioDeLaDescripcion{
		Marcada: true, Frase: fraseDelPrimerVoto, Motivo: motivoDelPrimerVoto,
		Votos: []VotoDeLaDescripcion{dado, dado, dado},
	}, votacion.juzgarLaDescripcion(skillDeLosVotos, descripcionDeLosVotos))

	votos := sustituto.votos(t)
	require.Len(t, votos, votosParaMarcar, "un proceso del sustituto por voto")

	for _, voto := range votos {
		assert.Equal(t, mensajeDeLaDescripcion(skillDeLosVotos, descripcionDeLosVotos), voto.entrada,
			"el mensaje de la descripción llega por la entrada estándar, byte a byte")
		assert.Equal(t, []string{
			"-p", "--model", modeloDelJuezDelGuion, "--tools", "", "--strict-mcp-config", "--disable-slash-commands",
			"--no-session-persistence", "--system-prompt", juez.Rubrica, "--output-format", "json",
			"--json-schema", strings.TrimRight(juez.Esquema, "\n"),
		}, voto.argumentos, "la rúbrica y el esquema de la evidencia, con el guion de voto de las evals")
	}
}
