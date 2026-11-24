package app

import (
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/plazos"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/core/territorio"
	"github.com/jmorenobl/kitlegal/internal/httpx"
	"github.com/jmorenobl/kitlegal/internal/skills"
	"github.com/jmorenobl/kitlegal/internal/source/bocm"
)

// Lo que compone el applet del caso 8 y lo que leen los casos de cobertura,
// relativo al directorio de este paquete (contracts/casos-de-plazos.md de H9;
// research.md D23 y D30 de H9).
const (
	// manifiestoDelBOCM es el de las disposiciones grabadas del BOCM.
	manifiestoDelBOCM = "../source/bocm/testdata/grabaciones.json"
	// grabadoDelBOCM es lo que grabó grabar_datos de esas disposiciones.
	grabadoDelBOCM = "../../evidencias/h9/" + bocm.NombreDeLaFuente
	// principalDeLosLocalesDe2026 es la resolución principal de las fiestas
	// locales de 2026 de la Comunidad de Madrid, la que da a Brunete como «no
	// comunicado» antes de que la corrija BOCM-20251229-19.
	principalDeLosLocalesDe2026 = "BOCM-20251212-34"
	// coberturaDelRepositorio es data/cobertura.yaml, que el binario no lleva.
	coberturaDelRepositorio = "../../data/cobertura.yaml"
	// tiposDelRepositorio es data/plazos.yaml, del que salen los tipos que
	// nombra el error del tipo que no existe.
	tiposDelRepositorio = "../../data/plazos.yaml"
)

// Las normas de los casos de cobertura, por su identificador: la LPAC, la LGT y
// el texto refundido de Tráfico, este último el que da su búsqueda grabada
// (research.md S5 de H9; TestPreceptosDeLosPlazos lo comprueba contra ella).
const (
	lpacDeLaTabla    = "BOE-A-2015-10565"
	lgtDeLaTabla     = "BOE-A-2003-23186"
	traficoDeLaTabla = "BOE-A-2015-11722"
)

// Los códigos de aviso de la tabla, escritos aquí otra vez y no tomados del
// código que se prueba (contracts/applet-plazos.md §5 de H9).
const (
	faltanNacionalesDeLaTabla = "faltan-festivos-nacionales-y-autonomicos"
	faltanLocalesDeLaTabla    = "faltan-festivos-locales"
	diasDelArt183DeLaTabla    = "dias-del-art-183-lopj"
	deUnaParteDeLaTabla       = "festivo-de-una-parte"
)

// plazoDeLaTabla es lo que la tabla compara de data: la fecha y los avisos, con
// su código y su texto, leídos como los lee quien llama y no con el tipo del
// dominio.
type plazoDeLaTabla struct {
	Fecha  string           `json:"fecha"`
	Avisos []avisoDeLaTabla `json:"avisos"`
}

// avisoDeLaTabla es un aviso de data.
type avisoDeLaTabla struct {
	Codigo string `json:"codigo"`
	Texto  string `json:"texto"`
}

// codigos son los de los avisos, en el orden en que llegan, o nil si no hay
// ninguno.
func (p plazoDeLaTabla) codigos() []string {
	var codigos []string
	for _, aviso := range p.Avisos {
		codigos = append(codigos, aviso.Codigo)
	}

	return codigos
}

// plazoDeLaSalida es el data de la salida con --json de un calcular que
// termina en 0, con un sobre de éxito y nada en la salida de error.
func plazoDeLaSalida(t *testing.T, res invocacionDePrueba) plazoDeLaTabla {
	t.Helper()

	require.Equal(t, 0, res.codigo, res.errores)
	assert.Empty(t, res.errores, "una salida correcta no escribe nada en la salida de error")

	var sobre struct {
		OK   bool           `json:"ok"`
		Data plazoDeLaTabla `json:"data"`
	}

	require.NoError(t, json.Unmarshal([]byte(res.salida), &sobre))
	require.True(t, sobre.OK, res.salida)

	return sobre.Data
}

// casoDePlazo es una fila con fecha de contracts/casos-de-plazos.md de H9: los
// argumentos de calcular como los da la skill —los municipios por su nombre—,
// la fecha calculada a mano y los códigos de aviso, en el orden de §5 del
// contrato del applet.
type casoDePlazo struct {
	nombre string
	orden  []string
	fecha  string
	avisos []string
	// conLaPrincipalSola compone el applet con los festivos locales que se
	// derivan de la resolución principal de 2026 sola (caso 8).
	conLaPrincipalSola bool
}

// casosDePlazos son los diecinueve casos con fecha de la tabla, cada uno con su
// número (FR-002 a FR-011, FR-024, FR-026, FR-082; SC-005). Su razonamiento
// está en la tabla.
func casosDePlazos() []casoDePlazo {
	return []casoDePlazo{
		{nombre: "01-eval-a", orden: []string{"reposicion", "2026-09-09", "Leganés", "Leganés"}, fecha: "2026-10-13"},
		{
			nombre: "02-eval-b", orden: []string{"reposicion", "2026-09-09", "Tordesillas", "Tordesillas"},
			fecha: "2026-10-09", avisos: []string{faltanLocalesDeLaTabla},
		},
		{nombre: "03-eval-e", orden: []string{"contencioso", "2026-06-15", "Madrid"}, fecha: "2026-09-15"},
		{
			nombre: "04-art-30-6-festivo-del-interesado", orden: []string{"alzada", "2026-09-09", "Leganés", "Madrid"},
			fecha: "2026-10-13",
		},
		{
			nombre: "05-art-30-6-festivo-de-la-sede", orden: []string{"reposicion", "2026-10-09", "Leganés", "Madrid"},
			fecha: "2026-11-10",
		},
		{
			nombre: "06-ultimo-dia-en-sabado", orden: []string{"reposicion", "2026-09-03", "Leganés", "Leganés"},
			fecha: "2026-10-05",
		},
		{
			nombre: "07-cambio-de-anio-hacia-2027",
			orden: []string{
				"generico", "2026-12-21", "Leganés", "Leganés", "--duracion", "10", "--unidad", "dias-habiles",
			},
			fecha: "2027-01-07", avisos: []string{faltanLocalesDeLaTabla},
		},
		{
			nombre: "08-no-comunicado-antes-de-corregirse", orden: []string{"reposicion", "2026-08-14", "Brunete", "Brunete"},
			fecha: "2026-09-14", avisos: []string{faltanLocalesDeLaTabla}, conLaPrincipalSola: true,
		},
		{
			nombre: "09-no-comunicado-ya-corregido", orden: []string{"reposicion", "2026-08-14", "Brunete", "Brunete"},
			fecha: "2026-09-15",
		},
		{
			nombre: "10-entidad-menor", orden: []string{"reposicion", "2026-09-09", "El Vellón", "El Vellón"},
			fecha: "2026-10-09",
		},
		{
			nombre: "11-silencio",
			orden:  []string{"silencio", "2026-07-09", "Leganés", "Leganés", "--duracion", "3", "--unidad", "meses"},
			fecha:  "2026-10-14",
		},
		{
			nombre: "12-dias-habiles",
			orden: []string{
				"generico", "2026-10-07", "Leganés", "Leganés", "--duracion", "5", "--unidad", "dias-habiles",
			},
			fecha: "2026-10-16",
		},
		{
			nombre: "13-dias-naturales",
			orden: []string{
				"generico", "2026-09-24", "Leganés", "Leganés", "--duracion", "15", "--unidad", "dias-naturales",
			},
			fecha: "2026-10-13",
		},
		{
			nombre: "14-mes-sin-dia-equivalente",
			orden:  []string{"generico", "2026-01-31", "Leganés", "Leganés", "--duracion", "1", "--unidad", "meses"},
			fecha:  "2026-03-02",
		},
		{
			nombre: "15-contencioso-con-el-art-183", orden: []string{"contencioso", "2026-10-27", "Madrid"},
			fecha: "2026-12-28", avisos: []string{diasDelArt183DeLaTabla},
		},
		{
			nombre: "16-contencioso-con-el-31-de-diciembre", orden: []string{"contencioso", "2026-10-31", "Madrid"},
			fecha: "2026-12-31", avisos: []string{diasDelArt183DeLaTabla},
		},
		{
			nombre: "17-anio-anterior-a-2025",
			orden: []string{
				"generico", "2024-12-30", "Leganés", "Leganés", "--duracion", "3", "--unidad", "dias-habiles",
			},
			fecha: "2025-01-03", avisos: []string{faltanNacionalesDeLaTabla, faltanLocalesDeLaTabla},
		},
		{
			nombre: "18-mes-que-cruza-a-2027", orden: []string{"reposicion", "2026-12-14", "Leganés", "Leganés"},
			fecha: "2027-01-14", avisos: []string{faltanLocalesDeLaTabla},
		},
		{
			nombre: "21-fiesta-sustituida-en-una-parte",
			orden:  []string{"reposicion", "2025-11-26", "Vielha e Mijaran", "Vielha e Mijaran"},
			fecha:  "2025-12-26", avisos: []string{faltanLocalesDeLaTabla, deUnaParteDeLaTabla},
		},
	}
}

// casoDeCobertura es una fila de cobertura de la tabla: el procedimiento que la
// ley regula aparte, buscado como dice buscada; el precepto de la LPAC que lo
// dice y la norma que lo rige, con el precepto que la hace aplicable.
type casoDeCobertura struct {
	nombre string
	// buscada dice, para el mensaje del test, cómo se busca la fila.
	buscada string
	// es dice si una fila de lo regulado aparte es la del caso.
	es        func(skills.ReguladoAparte) bool
	precepto  skills.PreceptoDeLaCobertura
	norma     string
	aplicable skills.PreceptoAplicable
}

// casosDeCobertura son los casos 19 y 20, las evals (c) y (d). El 19 busca su
// fila por su precepto y la norma que la rige, no por el texto de su
// procedimiento, que la tabla no fija (FR-016 de H45); el 20, por su texto.
func casosDeCobertura() []casoDeCobertura {
	tributos := skills.PreceptoDeLaCobertura{Norma: lpacDeLaTabla, Bloque: "daprimera", Apartado: "DA 1ª.2.a"}
	trafico := "procedimientos sancionadores en materia de tráfico y seguridad vial"

	return []casoDeCobertura{
		{
			nombre:    "19-eval-c-tributos",
			buscada:   "la del precepto " + tributos.Apartado + " de " + tributos.Norma + " regida por " + lgtDeLaTabla,
			es:        conSuPreceptoYSuNorma(tributos, lgtDeLaTabla),
			precepto:  tributos,
			norma:     lgtDeLaTabla,
			aplicable: skills.PreceptoAplicable{Bloque: "a97", Apartado: "97"},
		},
		{
			nombre:    "20-eval-d-trafico",
			buscada:   "la de «" + trafico + "»",
			es:        conSuProcedimiento(trafico),
			precepto:  skills.PreceptoDeLaCobertura{Norma: lpacDeLaTabla, Bloque: "daprimera", Apartado: "DA 1ª.2.c"},
			norma:     traficoDeLaTabla,
			aplicable: skills.PreceptoAplicable{Bloque: "a1", Apartado: "1"},
		},
	}
}

// conSuPreceptoYSuNorma dice si una fila lleva el precepto de la LPAC y la norma
// que la rige por su identificador.
func conSuPreceptoYSuNorma(precepto skills.PreceptoDeLaCobertura, norma string) func(skills.ReguladoAparte) bool {
	return func(fila skills.ReguladoAparte) bool {
		return fila.Precepto == precepto && fila.Norma != nil && fila.Norma.Identificador == norma
	}
}

// conSuProcedimiento dice si una fila describe el procedimiento con el texto.
func conSuProcedimiento(procedimiento string) func(skills.ReguladoAparte) bool {
	return func(fila skills.ReguladoAparte) bool {
		return fila.Procedimiento == procedimiento
	}
}

// TestCasosDePlazos comprueba la tabla de contracts/casos-de-plazos.md de H9,
// un subtest por caso (FR-082, SC-005): los diecinueve con fecha llaman al
// applet como lo llama la skill, con los municipios por su nombre y los
// ficheros del binario, y comparan la fecha y los códigos de aviso; el 8 lo
// compone con los festivos locales que deriva internal/source/bocm de la
// resolución principal de 2026 sola (research.md D23 y D30 de H9); el 19 y el
// 20 buscan su entrada en data/cobertura.yaml, leído con skills.LeerCobertura,
// el 19 por su precepto y su norma y el 20 por su procedimiento. Y el error del
// tipo que no existe, con su mensaje entero (elTipoQueNoExiste).
func TestCasosDePlazos(t *testing.T) {
	t.Parallel()

	empotrado := registroDePlazos(t, fuentesDePlazos(t))
	conLaPrincipalSola := registroDePlazos(t, fuentesConLaPrincipalSola(t))

	for _, caso := range casosDePlazos() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			registro := empotrado
			if caso.conLaPrincipalSola {
				registro = conLaPrincipalSola
			}

			plazo := plazoDeLaSalida(t, ordenDePlazos(t, registro, caso.orden...))

			assert.Equal(t, caso.fecha, plazo.Fecha, "%q", caso.orden)
			assert.Equal(t, caso.avisos, plazo.codigos(), "%q: los avisos, en el orden de sus códigos", caso.orden)
		})
	}

	cobertura := coberturaDeLaTabla(t)

	for _, caso := range casosDeCobertura() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			compruebaCasoDeCobertura(t, cobertura, caso)
		})
	}

	t.Run("tipo-que-no-existe", func(t *testing.T) {
		t.Parallel()

		res := ordenDePlazos(t, empotrado, "inventado", "2026-09-09", "Leganés", "Leganés")

		exigirSobreDeFallo(t, res, schema.ClaseArgumentos, 2, procedenciaDeUnFalloDePlazos)
		assert.Equal(t, elTipoQueNoExiste(t, "inventado"), datosDelSobre(t, sobreDelJSON(t, res.salida))["mensaje"])
	})
}

// elTipoQueNoExiste es el mensaje del error del tipo pedido que no es ninguno de
// data/plazos.yaml, como lo da el binario de main (§C.6 de H36; FR-017 y FR-026
// de H45): «el tipo "<pedido>" no es » y los códigos de sus tipos en el orden
// del fichero, separados por «, » y con « ni » delante del último. Los códigos
// salen del fichero leído con plazos.LeerTipos, no de una lista, para que un
// paquete más traiga los suyos; la enumeración se escribe aquí y no con la del
// dominio que se prueba.
func elTipoQueNoExiste(t *testing.T, pedido string) string {
	t.Helper()

	contenido, err := os.ReadFile(tiposDelRepositorio)
	require.NoError(t, err)

	tipos, err := plazos.LeerTipos(contenido)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(tipos), 2, "data/plazos.yaml lleva más de un tipo")

	codigos := make([]string, 0, len(tipos))
	for _, tipo := range tipos {
		codigos = append(codigos, tipo.Codigo)
	}

	ultimo := len(codigos) - 1

	return `el tipo "` + pedido + `" no es ` + strings.Join(codigos[:ultimo], ", ") + " ni " + codigos[ultimo]
}

// compruebaCasoDeCobertura exige que data/cobertura.yaml lleve una sola entrada
// del caso, con su precepto de la LPAC y la norma que lo rige.
func compruebaCasoDeCobertura(t *testing.T, cobertura skills.Cobertura, caso casoDeCobertura) {
	t.Helper()

	var entradas []skills.ReguladoAparte

	for _, entrada := range cobertura.ReguladosAparte {
		if caso.es(entrada) {
			entradas = append(entradas, entrada)
		}
	}

	require.Len(t, entradas, 1, "data/cobertura.yaml lleva una entrada, %s", caso.buscada)

	entrada := entradas[0]
	assert.Equal(t, caso.precepto, entrada.Precepto, "el precepto de la LPAC que la regula aparte")
	require.NotNil(t, entrada.Norma, "la entrada nombra la norma que la rige")
	assert.Equal(t, caso.norma, entrada.Norma.Identificador)
	assert.Equal(t, caso.aplicable, entrada.Norma.Aplicable, "el precepto que la hace aplicable")
}

// coberturaDeLaTabla es data/cobertura.yaml leído con skills.LeerCobertura.
func coberturaDeLaTabla(t *testing.T) skills.Cobertura {
	t.Helper()

	contenido, err := os.ReadFile(coberturaDelRepositorio)
	require.NoError(t, err)

	cobertura, err := skills.LeerCobertura(contenido)
	require.NoError(t, err)

	return cobertura
}

// fuentesConLaPrincipalSola son los ficheros del binario con los festivos
// locales de la Comunidad de Madrid que internal/source/bocm deriva, sin red y
// de lo grabado, de la resolución principal de 2026 sola, escritos con
// plazos.EscribirFestivosLocales: el fichero que habría antes de las
// correcciones (research.md D30 de H9).
func fuentesConLaPrincipalSola(t *testing.T) FicherosDePlazos {
	t.Helper()

	fuentes := fuentesDePlazos(t)

	registro, err := territorio.Cargar(fuentes.Territorio)
	require.NoError(t, err)

	contenido, err := os.ReadFile(manifiestoDelBOCM)
	require.NoError(t, err)

	var manifiesto bocm.Manifiesto
	require.NoError(t, json.Unmarshal(contenido, &manifiesto))

	indice := slices.IndexFunc(manifiesto.Disposiciones, func(d bocm.Disposicion) bool {
		return d.ID == principalDeLosLocalesDe2026
	})
	require.GreaterOrEqual(t, indice, 0, "el manifiesto del BOCM lista %s", principalDeLosLocalesDe2026)

	soloLaPrincipal := bocm.Manifiesto{
		Fuente:        manifiesto.Fuente,
		Disposiciones: []bocm.Disposicion{manifiesto.Disposiciones[indice]},
	}

	cliente, err := httpx.Replay(grabadoDelBOCM, httpx.ConFuente(bocm.NombreDeLaFuente))
	require.NoError(t, err)

	fichero, err := bocm.Derivar(t.Context(), cliente, soloLaPrincipal, registro)
	require.NoError(t, err)

	locales, err := plazos.EscribirFestivosLocales(fichero, plazos.Cabecera{
		Desde:   "evidencias/h9/" + bocm.NombreDeLaFuente + "/",
		Test:    "TestCasosDePlazos",
		Paquete: "internal/app",
	})
	require.NoError(t, err)

	fuentes.Locales[bocm.ComunidadDelBoletin] = locales

	return fuentes
}
