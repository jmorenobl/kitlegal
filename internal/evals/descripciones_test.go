package evals

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
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

// casoDelMensaje es lo que los tests del mensaje y de la medida leen de un caso
// de casos.yaml de la evidencia: su id, su skill, su descripción y su etiqueta.
type casoDelMensaje struct {
	ID          string `yaml:"id"`
	Skill       string `yaml:"skill"`
	Descripcion string `yaml:"descripcion"`
	Etiqueta    string `yaml:"etiqueta"`
}

// casosDelMensaje es el documento casos.yaml de la evidencia, con lo que los
// tests del mensaje y de la medida leen de cada caso.
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
//   - el votante del guion de las evals, con el claude sustituto, recibe el
//     mensaje de la descripción por su entrada estándar, con la rúbrica y el
//     esquema de la evidencia;
//   - y la ejecución de la medida del juez de las descripciones, la de
//     TestMedidaDelJuezDeLasDescripciones (FR-034 de H31;
//     contracts/juez-de-las-descripciones.md §7 de H31; research D9 de H31),
//     vota con esa regla los casos de casos.yaml y da la medida con la forma de
//     medida.json, que con los casos de la evidencia y los votos que dicen lo
//     que dice su etiqueta es la versionada; una línea por caso que no da lo
//     que dice su etiqueta, con la medida; y, con algún caso sin juzgar, una
//     línea por cada uno y ninguna medida.
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

	t.Run("medida", func(t *testing.T) {
		t.Parallel()

		probarLaMedidaDelJuezDeLasDescripciones(t)
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
		{"skill", func(d *datosDeLaHuella) { d.skill = "normativa" }},
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
	skillDeLosVotos       = "normativa"
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

// formaDeLaMedidaDeLasDescripciones es la medida que da la ejecución de la
// medida del juez de las descripciones, con sus claves en el orden de
// contracts/juez-de-las-descripciones.md §7 de H31, las de medida.json de la
// evidencia sin skill y sin votos: con el día, el modelo, la versión, las dos
// huellas, los dos recuentos y el commit del origen.
const formaDeLaMedidaDeLasDescripciones = `{
  "clase": "habla_en_primera_o_segunda_persona",
  "fecha": "%s",
  "modelo_del_juez": "%s",
  "version_de_claude_code": "%s",
  "rubrica": {
    "fichero": "rubrica.md",
    "sha256": "%s"
  },
  "casos": {
    "fichero": "casos.yaml",
    "sha256": "%s"
  },
  "defectos": {
    "casos": %d,
    "sin_marcar": %d
  },
  "correctos": {
    "casos": %d,
    "marcados": %d
  },
  "origen": "ejecución de la medida del juez de las descripciones del job de evals sobre %s"
}
`

// votantePorMensaje es un votante simulado de la medida del juez de las
// descripciones, al que se llama desde varias gorrutinas a la vez: da, por
// orden, las grabaciones del mensaje con el que se pide cada voto, y anota
// cuántos se piden con cada uno. Un voto con un mensaje del que no queda
// ninguna grabación es un error del test.
type votantePorMensaje struct {
	t           *testing.T
	grabaciones map[string][]grabacion

	mutex   sync.Mutex
	pedidos map[string]int
}

// nuevoVotantePorMensaje es el votante simulado que da esas grabaciones de cada
// mensaje.
func nuevoVotantePorMensaje(t *testing.T, grabaciones map[string][]grabacion) *votantePorMensaje {
	t.Helper()

	return &votantePorMensaje{t: t, grabaciones: grabaciones, pedidos: make(map[string]int, len(grabaciones))}
}

// votar da la grabación siguiente del mensaje.
func (v *votantePorMensaje) votar(mensaje string) ([]byte, error) {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	dados := v.pedidos[mensaje]
	v.pedidos[mensaje] = dados + 1

	grabaciones := v.grabaciones[mensaje]
	if dados >= len(grabaciones) {
		v.t.Errorf("se pide el voto %d de un mensaje del que hay %d grabados:\n%s", dados+1, len(grabaciones), mensaje)

		return nil, errors.New("voto sin grabar")
	}

	return []byte(grabaciones[dados].salida), grabaciones[dados].err
}

// exigirTodos exige que se hayan pedido con cada mensaje todos sus votos
// grabados, ni uno más ni uno menos, y devuelve cuántos se han pedido.
func (v *votantePorMensaje) exigirTodos() int {
	v.t.Helper()

	v.mutex.Lock()
	defer v.mutex.Unlock()

	for mensaje, grabaciones := range v.grabaciones {
		assert.Equal(v.t, len(grabaciones), v.pedidos[mensaje], "los votos pedidos con el mensaje:\n%s", mensaje)
	}

	pedidos := 0
	for _, deUnMensaje := range v.pedidos {
		pedidos += deUnMensaje
	}

	return pedidos
}

// probarLaMedidaDelJuezDeLasDescripciones fija la ejecución de la medida del
// juez de las descripciones (contracts/juez-de-las-descripciones.md §7 de H31;
// research D9 de H31; FR-034), la que lanza TestMedidaDelJuezDeLasDescripciones,
// con votantes simulados que no abren ninguna sesión (FR-033; SC-014): con los
// casos de la evidencia, la medida versionada; con unos casos sintéticos, las
// líneas de los que no dan lo que dice su etiqueta, con la medida, y las de los
// que quedan sin juzgar, sin ella; y, sin juez, sin votante, sin casos a la vez,
// con un juez de otra clase o con casos que no tienen su forma, ningún voto ni
// ninguna medida.
func probarLaMedidaDelJuezDeLasDescripciones(t *testing.T) {
	t.Helper()

	t.Run("con-los-casos-de-la-evidencia", func(t *testing.T) {
		t.Parallel()

		probarLaMedidaConLosCasosDeLaEvidencia(t)
	})

	for _, caso := range casosDeLaMedidaSintetica() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			caso.probar(t)
		})
	}

	for _, caso := range medicionesDeLasDescripcionesQueNoVotan() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			caso.probar(t)
		})
	}
}

// probarLaMedidaConLosCasosDeLaEvidencia mide al juez de las descripciones del
// repositorio con sus 99 casos y un votante que dice de cada uno lo que dice su
// etiqueta: tres «si» de cada defecto, con la descripción entera como frase, y
// un «no» de cada correcto. La medida que da es la versionada salvo el día y el
// origen —las mismas huellas, el mismo modelo, la misma versión y los mismos
// recuentos, 63 de 63 y 0 de 36—, con sus claves en el orden del contrato y sin
// skill ni votos; comprobarLaMedida la da por buena con ese modelo y esa
// versión; y se piden 225 votos, cinco casos a la vez, sin ningún error.
func probarLaMedidaConLosCasosDeLaEvidencia(t *testing.T) {
	t.Helper()

	juez := juezDeLasDescripcionesEn(t, raizDelRepositorio)

	versionada, err := leerMedidaDelJuez([]byte(juez.Medida))
	require.NoError(t, err)

	casos, err := skills.ValidarDocumentoYAML[casosDelMensaje]([]byte(juez.Casos), nil)
	require.NoError(t, err)

	grabaciones := make(map[string][]grabacion, len(casos.Casos))
	etiquetas := make(map[string]string, len(casos.Casos))

	for _, caso := range casos.Casos {
		mensaje := mensajeDeLaDescripcion(caso.Skill, caso.Descripcion)
		if etiqueta, esta := etiquetas[mensaje]; esta {
			require.Equal(t, etiqueta, caso.Etiqueta, "premisa: los casos con el mensaje del caso %s tienen su etiqueta",
				caso.ID)
		}

		etiquetas[mensaje] = caso.Etiqueta

		if caso.Etiqueta == etiquetaDefecto {
			si := votoDeLaDescripcion(t, "si", caso.Descripcion, motivoDelPrimerVoto)
			grabaciones[mensaje] = append(grabaciones[mensaje], si, si, si)

			continue
		}

		grabaciones[mensaje] = append(grabaciones[mensaje], votoDeLaDescripcion(t, "no", "", motivoDelVotoQueNo))
	}

	votante := nuevoVotantePorMensaje(t, grabaciones)

	texto, err := medirAlJuezDeLasDescripciones(MedicionDelJuez{
		Juez: juez, Votar: votante.votar, ModeloDelJuez: versionada.ModeloDelJuez,
		VersionDelJuez: versionada.VersionDeClaudeCode, Concurrencia: concurrenciaDeLaMedidaDeLasDescripciones,
		Commit: commitDeLaMedicion, Fecha: instanteDeLaMedicion(),
	})
	require.NoError(t, err)

	exigirLaMedidaDeLasDescripciones(t, fmt.Sprintf(formaDeLaMedidaDeLasDescripciones, fechaDeLaMedidaDada,
		versionada.ModeloDelJuez, versionada.VersionDeClaudeCode, versionada.HuellaDeLaRubrica, versionada.HuellaDeLosCasos,
		defectosDeLaMedidaDeLasDescripciones, 0, correctosDeLaMedidaDeLasDescripciones, 0, commitDeLaMedicion), texto)
	assert.Equal(t, votosParaMarcar*defectosDeLaMedidaDeLasDescripciones+correctosDeLaMedidaDeLasDescripciones,
		votante.exigirTodos(), "tres votos por cada defecto y uno por cada correcto")

	conLaDada := *juez
	conLaDada.Medida = texto

	assert.Empty(t, comprobarLaMedida(&conLaDada, versionada.ModeloDelJuez, versionada.VersionDeClaudeCode),
		"la medida que da corresponde a lo que hay y se cumple")
}

// Los casos etiquetados sintéticos de la medida del juez de las descripciones,
// con la forma de casos.yaml de la evidencia: dos defectos y dos correctos, cada
// uno con una descripción de una línea, y las frases que los votos grabados
// citan de los tres primeros.
const (
	descripcionDelCaso1 = "Consulta y cita normativa consolidada del BOE. Úsala siempre que la respuesta dependa de una norma."
	descripcionDelCaso2 = "Calcula con kitlegal hasta qué día se puede recurrir una resolución."
	descripcionDelCaso3 = "Cita sentencias con su documento delante. Úsala cuando se pida una sentencia."
	descripcionDelCaso4 = "Identifica el territorio de un municipio y sus boletines."

	fraseDelCaso1 = "Úsala siempre"
	fraseDelCaso2 = "hasta qué día se puede recurrir"
	fraseDelCaso3 = "Úsala cuando se pida una sentencia"

	casosSinteticosDeLasDescripciones = "casos:\n" +
		"- id: '001'\n  skill: normativa\n  etiqueta: defecto\n  descripcion: '" + descripcionDelCaso1 + "'\n" +
		"- id: '002'\n  skill: legal-core\n  etiqueta: correcto\n  descripcion: '" + descripcionDelCaso2 + "'\n" +
		"- id: '003'\n  skill: jurisprudencia\n  etiqueta: defecto\n  descripcion: '" + descripcionDelCaso3 + "'\n" +
		"- id: '004'\n  skill: legal-core\n  etiqueta: correcto\n  descripcion: '" + descripcionDelCaso4 + "'\n"
)

// Los mensajes de los votos de cada caso sintético.
var (
	mensajeDelCaso1 = mensajeDeLaDescripcion("normativa", descripcionDelCaso1)
	mensajeDelCaso2 = mensajeDeLaDescripcion("legal-core", descripcionDelCaso2)
	mensajeDelCaso3 = mensajeDeLaDescripcion("jurisprudencia", descripcionDelCaso3)
	mensajeDelCaso4 = mensajeDeLaDescripcion("legal-core", descripcionDelCaso4)
)

// juezConLosCasosSinteticos es el juez de las descripciones del repositorio con
// esos casos en lugar de los de la evidencia.
func juezConLosCasosSinteticos(t *testing.T, casos string) *Juez {
	t.Helper()

	juez := *juezDeLasDescripcionesEn(t, raizDelRepositorio)
	juez.Casos = casos

	return &juez
}

// casoDeLaMedidaSintetica es un caso de la medida con los casos sintéticos: lo
// que el votante simulado da con el mensaje de cada caso, todos los votos que se
// le tienen que pedir; los dos recuentos que no dan lo de su etiqueta, o
// ninguna medida si sinMedida; y las líneas del error, en su orden.
type casoDeLaMedidaSintetica struct {
	nombre      string
	grabaciones func(t *testing.T) map[string][]grabacion
	sinMarcar   int
	marcados    int
	sinMedida   bool
	lineas      []string
}

// casosDeLaMedidaSintetica son los casos de la medida con los casos sintéticos:
// el defecto que no queda marcado y el correcto que sí, con la medida y la
// línea de cada uno, con las frases de sus «si»; y dos casos sin juzgar, uno en
// su primer voto y otro en el segundo, con la línea de cada uno, con su motivo,
// y sin medida.
func casosDeLaMedidaSintetica() []casoDeLaMedidaSintetica {
	return []casoDeLaMedidaSintetica{
		{
			nombre: "casos-que-no-dan-su-etiqueta",
			grabaciones: func(t *testing.T) map[string][]grabacion {
				t.Helper()

				si1 := votoDeLaDescripcion(t, "si", fraseDelCaso1, motivoDelPrimerVoto)
				si2 := votoDeLaDescripcion(t, "si", fraseDelCaso2, motivoDelSegundoVoto)
				si3 := votoDeLaDescripcion(t, "si", fraseDelCaso3, motivoDelTercerVoto)
				no := votoDeLaDescripcion(t, "no", "", motivoDelVotoQueNo)

				return map[string][]grabacion{
					mensajeDelCaso1: {si1, si1, no}, mensajeDelCaso2: {si2, si2, si2},
					mensajeDelCaso3: {si3, si3, si3}, mensajeDelCaso4: {no},
				}
			},
			sinMarcar: 1,
			marcados:  1,
			lineas: []string{
				"001: etiquetado defecto y sin marcar: «" + fraseDelCaso1 + "» · «" + fraseDelCaso1 + "»",
				"002: etiquetado correcto y marcado: «" + fraseDelCaso2 + "» · «" + fraseDelCaso2 + "» · «" +
					fraseDelCaso2 + "»",
			},
		},
		{
			nombre: "casos-sin-juzgar",
			grabaciones: func(t *testing.T) map[string][]grabacion {
				t.Helper()

				si1 := votoDeLaDescripcion(t, "si", fraseDelCaso1, motivoDelPrimerVoto)
				si3 := votoDeLaDescripcion(t, "si", fraseDelCaso3, motivoDelTercerVoto)
				conError := sesionDelJuezConError(t, errorDeLaSesionDelJuez)
				no := votoDeLaDescripcion(t, "no", "", motivoDelVotoQueNo)

				return map[string][]grabacion{
					mensajeDelCaso1: {si1, si1, si1}, mensajeDelCaso2: {conError},
					mensajeDelCaso3: {si3, conError}, mensajeDelCaso4: {no},
				}
			},
			sinMedida: true,
			lineas: []string{
				"002: sin juzgar: " + prefijoDeSesionConError + errorDeLaSesionDelJuez,
				"003: sin juzgar: voto 2: la sesión del juez terminó con error: " + errorDeLaSesionDelJuez,
			},
		},
	}
}

// probar mide al juez con los casos sintéticos, dos a la vez, con el votante
// simulado del caso, y exige sus líneas, unidas en el error; la medida, con las
// huellas de la rúbrica de la evidencia y de los casos sintéticos y los dos
// recuentos, o ninguna; y que se hayan pedido todos los votos grabados.
func (c casoDeLaMedidaSintetica) probar(t *testing.T) {
	t.Helper()

	juez := juezConLosCasosSinteticos(t, casosSinteticosDeLasDescripciones)
	votante := nuevoVotantePorMensaje(t, c.grabaciones(t))

	texto, err := medirAlJuezDeLasDescripciones(MedicionDelJuez{
		Juez: juez, Votar: votante.votar, ModeloDelJuez: modeloDelGuion, VersionDelJuez: versionDelGuion,
		Concurrencia: 2, Commit: commitDeLaMedicion, Fecha: instanteDeLaMedicion(),
	})

	require.EqualError(t, err, strings.Join(c.lineas, "\n"), "una línea por caso, en el orden de los casos")
	votante.exigirTodos()

	if c.sinMedida {
		assert.Empty(t, texto, "con un caso sin juzgar no hay medida")

		return
	}

	exigirLaMedidaDeLasDescripciones(t, fmt.Sprintf(formaDeLaMedidaDeLasDescripciones, fechaDeLaMedidaDada,
		modeloDelGuion, versionDelGuion, sha256DeTexto(juez.Rubrica), sha256DeTexto(casosSinteticosDeLasDescripciones),
		2, c.sinMarcar, 2, c.marcados, commitDeLaMedicion), texto)
}

// exigirLaMedidaDeLasDescripciones exige que la medida dada sea la esperada
// carácter a carácter, con su sangría, el orden de sus claves y su salto final,
// línea a línea: la forma es parte de lo que se versiona.
func exigirLaMedidaDeLasDescripciones(t *testing.T, esperada, dada string) {
	t.Helper()

	assert.Equal(t, slices.Collect(strings.Lines(esperada)), slices.Collect(strings.Lines(dada)),
		"la medida, con la forma de contracts/juez-de-las-descripciones.md §7 de H31")
}

// medicionDeLasDescripcionesQueNoVota es una medición de las descripciones que
// no llega a votar: lo que cambia en una medición que sí votaría y lo que dice
// su error.
type medicionDeLasDescripcionesQueNoVota struct {
	nombre  string
	cambiar func(medicion *MedicionDelJuez)
	dice    string
}

// medicionesDeLasDescripcionesQueNoVotan son las mediciones que no llegan a
// votar: sin juez, sin votante, sin casos a la vez, con un esquema que no
// compila, con un juez sin la clase de las descripciones, con unos casos que no
// son YAML y con un caso de otra etiqueta.
func medicionesDeLasDescripcionesQueNoVotan() []medicionDeLasDescripcionesQueNoVota {
	return []medicionDeLasDescripcionesQueNoVota{
		{
			nombre:  "sin-juez",
			cambiar: func(medicion *MedicionDelJuez) { medicion.Juez = nil },
			dice:    "no hay juez de las descripciones que medir",
		},
		{
			nombre:  "sin-votante",
			cambiar: func(medicion *MedicionDelJuez) { medicion.Votar = nil },
			dice:    "el juez de las descripciones no tiene votante",
		},
		{
			nombre:  "sin-casos-a-la-vez",
			cambiar: func(medicion *MedicionDelJuez) { medicion.Concurrencia = 0 },
			dice:    "los casos que se votan a la vez son 0 y tienen que ser al menos 1",
		},
		{
			nombre:  "con-un-esquema-que-no-compila",
			cambiar: func(medicion *MedicionDelJuez) { medicion.Juez.Esquema = `{"type": 3}` },
			dice:    "el esquema de la respuesta del juez no sirve para validar sus votos",
		},
		{
			nombre: "con-un-juez-de-otra-clase",
			cambiar: func(medicion *MedicionDelJuez) {
				medicion.Juez.Clases = []ClaseDelJuez{{Nombre: claseAfirmaLoNoLeido, Decide: true}}
			},
			dice: "el juez no tiene la clase " + claseHablaEnPrimeraOSegundaPersona + ", que decide",
		},
		{
			nombre:  "con-casos-que-no-son-yaml",
			cambiar: func(medicion *MedicionDelJuez) { medicion.Juez.Casos = "casos: [\n" },
			dice:    "los casos etiquetados del juez de las descripciones (evidencias/adr-0046/casos.yaml): ",
		},
		{
			nombre: "con-un-caso-de-otra-etiqueta",
			cambiar: func(medicion *MedicionDelJuez) {
				medicion.Juez.Casos = strings.Replace(casosSinteticosDeLasDescripciones, "etiqueta: correcto",
					"etiqueta: dudoso", 1)
			},
			dice: `el caso 2 (002) tiene la etiqueta "dudoso", que no es defecto ni correcto`,
		},
	}
}

// probar ejecuta la medición del caso con un votante que cuenta sus llamadas, y
// exige su error, ninguna medida y ningún voto.
func (m medicionDeLasDescripcionesQueNoVota) probar(t *testing.T) {
	t.Helper()

	var pedidos atomic.Int64

	medicion := MedicionDelJuez{
		Juez: juezConLosCasosSinteticos(t, casosSinteticosDeLasDescripciones),
		Votar: func(string) ([]byte, error) {
			pedidos.Add(1)

			return nil, errors.New("un voto que no se tenía que pedir")
		},
		ModeloDelJuez: modeloDelGuion, VersionDelJuez: versionDelGuion, Concurrencia: 2, Commit: commitDeLaMedicion,
		Fecha: instanteDeLaMedicion(),
	}
	m.cambiar(&medicion)

	texto, err := medirAlJuezDeLasDescripciones(medicion)

	require.ErrorContains(t, err, m.dice)
	assert.Empty(t, texto, "sin votar no hay medida")
	assert.Zero(t, pedidos.Load(), "no se pide ningún voto")
}
