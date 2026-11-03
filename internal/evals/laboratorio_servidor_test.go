package evals

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/app"
	"github.com/jmorenobl/kitlegal/internal/mcp"
	"github.com/jmorenobl/kitlegal/internal/mcp/mcptest"
)

// Lo que el filtro reescribe en TestServidorDelLaboratorio, además de la
// herramienta y el parámetro de la declaración del laboratorio
// (herramientaDelLaboratorio y parametroDelLaboratorio, los de la variante
// parametro; contracts/arnes.md §3.4 de H39): dos textos con comillas, comillas
// angulares, un salto de línea y los caracteres que el codificador del servidor
// escapa, para que se vea que el texto llega entero y como texto de JSON.
const (
	instruccionesDelLaboratorio = "Antes de la primera herramienta, activa la skill «articulos-del-boe».\n" +
		`Sin ella, no hay "cita" <ni> & respuesta.`
	descripcionDelLaboratorio = "Id del bloque, del campo id de boe_indice.\n" +
		`Cierra con la "regla" <de> & laboratorio.`
)

// entradaDelCliente es lo que TestServidorDelLaboratorio da al filtro como su
// entrada estándar: dos líneas, la última sin salto, que tienen que llegar al
// hijo byte a byte.
const entradaDelCliente = "primera línea del cliente\nsegunda, sin salto final"

// lineaDelSustituto es la que el kitlegal sustituto escribe en su salida de
// error, que tiene que llegar a la del filtro.
const lineaDelSustituto = "kitlegal sustituto: sirviendo\n"

// esperaDelFiltro es lo que TestServidorDelLaboratorio espera como mucho a que
// el filtro o el servidor en proceso terminen: uno que no termina es un fallo
// del test, y no un test que no acaba.
const esperaDelFiltro = 30 * time.Second

// TestServidorDelLaboratorio fija el filtro de laboratorio servidor
// (contracts/arnes.md §5 de H39; research.md D7 de H39; FR-021, FR-024, FR-029)
// con las líneas que escribe el servidor de producción en proceso a un cliente
// que se saluda y pide la lista de herramientas, y un kitlegal sustituto que las
// escribe tal cual cuando lo arrancan con mcp serve. Sin opciones, la salida es
// la del hijo byte a byte; con -instrucciones, solo cambia result.instructions
// de la respuesta a initialize; con -herramienta, -parametro y -descripcion,
// solo la description de ese parámetro de esa herramienta en la de tools/list.
// En todos, el hijo recibe mcp serve, el entorno del filtro y su entrada byte a
// byte, su salida de error llega a la del filtro y el filtro termina con su
// código. Una herramienta o un parámetro que la lista no trae, unos argumentos
// sin la forma de la orden y un kitlegal que no arranca son un defecto: su línea
// en la salida de error y 1.
func TestServidorDelLaboratorio(t *testing.T) {
	t.Parallel()

	lineas := lineasDelServidorDeProduccion(t)
	saludo := lineaConElResultado(t, lineas, "instructions")
	lista := lineaConElResultado(t, lineas, "tools")

	require.Equal(t, mcp.Instrucciones, miembro(t, decodificada(t, lineas[saludo]), "result")["instructions"],
		"premisa: el servidor de producción da las instrucciones de hoy (FR-029)")

	original := descripcionEnLaLista(t, decodificada(t, lineas[lista]))
	require.NotEmpty(t, original, "premisa: el parámetro %s de %s trae su ayuda como descripción",
		parametroDelLaboratorio, herramientaDelLaboratorio)

	t.Run("sin-opciones", func(t *testing.T) {
		t.Parallel()

		probarSinOpciones(t, lineas)
	})
	t.Run("con-instrucciones", func(t *testing.T) {
		t.Parallel()

		probarLaLineaReescrita(t, lineas, saludo,
			[]string{"-instrucciones", instruccionesDelLaboratorio},
			func(mensaje map[string]any) {
				miembro(t, mensaje, "result")["instructions"] = instruccionesDelLaboratorio
			})
	})
	t.Run("con-la-descripcion-del-parametro", func(t *testing.T) {
		t.Parallel()

		probarLaLineaReescrita(t, lineas, lista,
			[]string{
				"-herramienta", herramientaDelLaboratorio,
				"-parametro", parametroDelLaboratorio,
				"-descripcion", descripcionDelLaboratorio,
			},
			func(mensaje map[string]any) {
				propiedadEnLaLista(t, mensaje)["description"] = descripcionDelLaboratorio
			})
	})
	t.Run("codigo-del-hijo", func(t *testing.T) {
		t.Parallel()

		probarElCodigoDelHijo(t, lineas)
	})
	t.Run("lo-que-la-lista-no-trae", func(t *testing.T) {
		t.Parallel()

		probarLoQueLaListaNoTrae(t, lineas, lista)
	})
	t.Run("argumentos-sin-la-forma", func(t *testing.T) {
		t.Parallel()

		probarLosArgumentosSinLaForma(t, lineas)
	})
	t.Run("kitlegal-que-no-arranca", probarElKitlegalQueNoArranca)
}

// probarSinOpciones fija que, sin ninguna opción, el filtro no cambia nada: su
// salida es la del hijo byte a byte, también con un hijo cuyas líneas no son
// todas mensajes del protocolo ni terminan todas en salto de línea.
func probarSinOpciones(t *testing.T, lineas []string) {
	t.Helper()

	sustituto := escribirElServidorSustituto(t, unirLineas(lineas), 0)
	filtrado := sustituto.filtrar(t)

	require.Zerof(t, filtrado.codigo, "salida de error: %s", filtrado.errores)
	assert.Equal(t, unirLineas(lineas), filtrado.salida, "la salida del hijo, byte a byte")
	assert.Equal(t, lineaDelSustituto, filtrado.errores, "la salida de error del hijo y nada más")
	sustituto.exigirElHijo(t)

	extrana := unirLineas(lineas) + "no es JSON\n\n" + `{"result":{"instructions":` + "\n" + "sin salto final"
	sustituto = escribirElServidorSustituto(t, extrana, 0)
	filtrado = sustituto.filtrar(t, "-instrucciones", instruccionesDelLaboratorio)

	require.Zerof(t, filtrado.codigo, "salida de error: %s", filtrado.errores)
	assert.True(t, strings.HasSuffix(filtrado.salida, "no es JSON\n\n"+`{"result":{"instructions":`+"\n"+"sin salto final"),
		"las líneas que no son un mensaje pasan byte a byte, y la última, sin salto; la salida acaba en:\n%s",
		filtrado.salida[max(0, len(filtrado.salida)-200):])
}

// probarLaLineaReescrita fija que, con las opciones dadas, la salida tiene las
// mismas líneas que la del hijo, todas byte a byte salvo la del índice dado, y
// que esa, decodificada, es la original con el cambio que hace cambiar.
func probarLaLineaReescrita(
	t *testing.T, lineas []string, indice int, opciones []string, cambiar func(mensaje map[string]any),
) {
	t.Helper()

	sustituto := escribirElServidorSustituto(t, unirLineas(lineas), 0)
	filtrado := sustituto.filtrar(t, opciones...)

	require.Zerof(t, filtrado.codigo, "salida de error: %s", filtrado.errores)
	assert.Equal(t, lineaDelSustituto, filtrado.errores)
	sustituto.exigirElHijo(t)

	require.True(t, strings.HasSuffix(filtrado.salida, "\n"), "la última línea conserva su salto")
	salida := strings.Split(strings.TrimSuffix(filtrado.salida, "\n"), "\n")
	require.Len(t, salida, len(lineas), "una línea por cada una del hijo")

	for i, linea := range lineas {
		if i != indice {
			assert.Equalf(t, linea, salida[i], "la línea %d pasa byte a byte", i)

			continue
		}

		assert.NotEqual(t, linea, salida[i], "la línea %d cambia", i)

		esperado := decodificada(t, linea)
		cambiar(esperado)
		assert.Equal(t, esperado, decodificada(t, salida[i]), "la línea %d, decodificada, solo cambia en su campo", i)
	}
}

// probarElCodigoDelHijo fija que el filtro termina con el código con el que
// termina el hijo, con su salida entera.
func probarElCodigoDelHijo(t *testing.T, lineas []string) {
	t.Helper()

	const codigo = 4

	sustituto := escribirElServidorSustituto(t, unirLineas(lineas), codigo)
	filtrado := sustituto.filtrar(t)

	assert.Equal(t, codigo, filtrado.codigo)
	assert.Equal(t, unirLineas(lineas), filtrado.salida)
	assert.Equal(t, lineaDelSustituto, filtrado.errores)
}

// probarLoQueLaListaNoTrae fija que una herramienta o un parámetro que la
// respuesta a tools/list no trae son un defecto: el filtro escribe las líneas
// anteriores, no escribe la de la lista, dice en su salida de error qué falta y
// termina con 1.
func probarLoQueLaListaNoTrae(t *testing.T, lineas []string, lista int) {
	t.Helper()

	casos := []struct {
		nombre, herramienta, parametro, linea string
	}{
		{
			nombre:      "herramienta",
			herramienta: "boe_inexistente",
			parametro:   parametroDelLaboratorio,
			linea:       "laboratorio: la lista de herramientas no trae boe_inexistente\n",
		},
		{
			nombre:      "parametro",
			herramienta: herramientaDelLaboratorio,
			parametro:   "inexistente",
			linea:       "laboratorio: la herramienta boe_articulo no trae el parámetro inexistente\n",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			sustituto := escribirElServidorSustituto(t, unirLineas(lineas), 0)
			filtrado := sustituto.filtrar(t,
				"-herramienta", caso.herramienta, "-parametro", caso.parametro, "-descripcion", descripcionDelLaboratorio)

			assert.Equal(t, 1, filtrado.codigo)
			assert.Equal(t, lineaDelSustituto+caso.linea, filtrado.errores)
			assert.Equal(t, unirLineas(lineas[:lista]), filtrado.salida, "las líneas anteriores a la lista, y la lista no")
		})
	}
}

// probarLosArgumentosSinLaForma fija que unos argumentos sin la forma de la
// orden son un defecto antes de arrancar nada: la línea de uso, o la que dice
// qué texto no es UTF-8, en la salida de error, nada en la estándar, 1 y ningún
// hijo.
func probarLosArgumentosSinLaForma(t *testing.T, lineas []string) {
	t.Helper()

	const uso = "laboratorio: uso: laboratorio servidor -kitlegal <ruta> [-instrucciones <texto>] " +
		"[-herramienta <h> -parametro <p> -descripcion <texto>]\n"

	sustituto := escribirElServidorSustituto(t, unirLineas(lineas), 0)

	// Tras los casos, que corren en paralelo cuando esta función ya ha vuelto.
	t.Cleanup(func() {
		assert.NoFileExists(t, filepath.Join(sustituto.dir, anotacionDeLosArgumentos), "no se arranca ningún hijo")
	})

	casos := []struct {
		nombre string
		args   []string
		linea  string
	}{
		{nombre: "sin-argumentos", linea: uso},
		{nombre: "sin-kitlegal", args: []string{"-instrucciones", "x"}, linea: uso},
		{nombre: "kitlegal-vacio", args: []string{"-kitlegal", ""}, linea: uso},
		{nombre: "kitlegal-sin-valor", args: []string{"-kitlegal"}, linea: uso},
		{nombre: "argumento-de-mas", args: []string{"-kitlegal", sustituto.ruta, "mcp"}, linea: uso},
		{nombre: "bandera-ajena", args: []string{"-kitlegal", sustituto.ruta, "-otra", "x"}, linea: uso},
		{nombre: "ayuda", args: []string{"-h"}, linea: uso},
		{
			nombre: "herramienta-sola",
			args:   []string{"-kitlegal", sustituto.ruta, "-herramienta", herramientaDelLaboratorio},
			linea:  uso,
		},
		{
			nombre: "sin-descripcion",
			args: []string{
				"-kitlegal", sustituto.ruta,
				"-herramienta", herramientaDelLaboratorio, "-parametro", parametroDelLaboratorio,
			},
			linea: uso,
		},
		{
			nombre: "descripcion-sola",
			args:   []string{"-kitlegal", sustituto.ruta, "-descripcion", descripcionDelLaboratorio},
			linea:  uso,
		},
		{
			nombre: "instrucciones-que-no-son-utf8",
			args:   []string{"-kitlegal", sustituto.ruta, "-instrucciones", "activa\xff"},
			linea:  "laboratorio: el texto de -instrucciones no es UTF-8\n",
		},
		{
			nombre: "descripcion-que-no-es-utf8",
			args: []string{
				"-kitlegal", sustituto.ruta,
				"-herramienta", herramientaDelLaboratorio, "-parametro", parametroDelLaboratorio,
				"-descripcion", "id\xff",
			},
			linea: "laboratorio: el texto de -descripcion no es UTF-8\n",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			filtrado := filtrarCon(t, caso.args)

			assert.Equal(t, 1, filtrado.codigo)
			assert.Equal(t, caso.linea, filtrado.errores)
			assert.Empty(t, filtrado.salida)
		})
	}
}

// probarElKitlegalQueNoArranca fija que un kitlegal que no se puede arrancar es
// un defecto: una línea que lo nombra, nada en la salida estándar y 1.
func probarElKitlegalQueNoArranca(t *testing.T) {
	t.Parallel()

	ruta := filepath.Join(t.TempDir(), "no-existe", "kitlegal")
	filtrado := filtrarCon(t, []string{"-kitlegal", ruta})

	assert.Equal(t, 1, filtrado.codigo)
	assert.True(t, strings.HasPrefix(filtrado.errores, "laboratorio: "+ruta+" mcp serve no arranca: "),
		"la línea nombra la orden: %q", filtrado.errores)
	assert.Equal(t, 1, strings.Count(filtrado.errores, "\n"), "una sola línea")
	assert.Empty(t, filtrado.salida)
}

// Las anotaciones del kitlegal sustituto en su directorio: sus argumentos, su
// PATH y su entrada estándar. lineasDelHijo es el fichero con lo que escribe en
// su salida estándar.
const (
	anotacionDeLosArgumentos = "argumentos"
	anotacionDelPATH         = "path"
	anotacionDeLaEntrada     = "entrada"
	lineasDelHijo            = "lineas"
)

// servidorSustituto es un kitlegal sustituto: el directorio en el que anota lo
// que recibe y la ruta del ejecutable.
type servidorSustituto struct {
	dir, ruta string
}

// escribirElServidorSustituto escribe en un temporal un kitlegal que hace de
// kitlegal mcp serve: anota sus argumentos, su PATH y su entrada estándar
// entera, escribe lineaDelSustituto en su salida de error y la salida dada en la
// estándar, y termina con el código dado.
func escribirElServidorSustituto(t *testing.T, salida string, codigo int) servidorSustituto {
	t.Helper()

	dir := t.TempDir()

	raiz, err := os.OpenRoot(dir)
	require.NoError(t, err)

	defer func() { assert.NoError(t, raiz.Close()) }()

	require.NoError(t, raiz.WriteFile(lineasDelHijo, []byte(salida), 0o600))

	enDir := func(nombre string) string { return entreComillasSimples(filepath.Join(dir, nombre)) }
	guion := "#!/bin/sh\nset -eu\n" +
		`printf '%s\n' "$*" > ` + enDir(anotacionDeLosArgumentos) + "\n" +
		`printf '%s' "${PATH:-}" > ` + enDir(anotacionDelPATH) + "\n" +
		"cat > " + enDir(anotacionDeLaEntrada) + "\n" +
		"printf '%s' " + entreComillasSimples(lineaDelSustituto) + " >&2\n" +
		"cat " + enDir(lineasDelHijo) + "\n" +
		"exit " + strconv.Itoa(codigo) + "\n"

	require.NoError(t, escribirEjecutable(raiz, "kitlegal", guion))

	return servidorSustituto{dir: dir, ruta: filepath.Join(dir, "kitlegal")}
}

// filtrado es lo que da el filtro: su código, su salida y su salida de error.
type filtrado struct {
	codigo          int
	salida, errores string
}

// filtrar ejecuta el filtro con este sustituto como kitlegal y las opciones
// dadas.
func (s servidorSustituto) filtrar(t *testing.T, opciones ...string) filtrado {
	t.Helper()

	return filtrarCon(t, slices.Concat([]string{"-kitlegal", s.ruta}, opciones))
}

// filtrarCon ejecuta el filtro con los argumentos dados y entradaDelCliente
// como entrada, sin esperarlo más de esperaDelFiltro.
func filtrarCon(t *testing.T, args []string) filtrado {
	t.Helper()

	ctx, cancelar := context.WithTimeout(t.Context(), esperaDelFiltro)
	defer cancelar()

	var salida, errores bytes.Buffer

	codigo := ServidorDelLaboratorio(ctx, args, strings.NewReader(entradaDelCliente), &salida, &errores)

	return filtrado{codigo: codigo, salida: salida.String(), errores: errores.String()}
}

// exigirElHijo exige que el sustituto se haya arrancado con mcp serve, con el
// PATH del proceso del test, que el filtro hereda, y con la entrada del filtro
// byte a byte.
func (s servidorSustituto) exigirElHijo(t *testing.T) {
	t.Helper()

	raiz, err := os.OpenRoot(s.dir)
	require.NoError(t, err)

	defer func() { assert.NoError(t, raiz.Close()) }()

	leer := func(nombre string) string {
		contenido, err := raiz.ReadFile(nombre)
		require.NoError(t, err)

		return string(contenido)
	}

	assert.Equal(t, "mcp serve\n", leer(anotacionDeLosArgumentos), "los argumentos del hijo")
	assert.Equal(t, os.Getenv("PATH"), leer(anotacionDelPATH), "el hijo tiene el entorno del filtro")
	assert.Equal(t, entradaDelCliente, leer(anotacionDeLaEntrada), "la entrada llega al hijo byte a byte")
}

// lineasDelServidorDeProduccion son las líneas que escribe kitlegal mcp serve,
// con los applets de producción y atendido por app.Main en una gorrutina del
// test, a un cliente de la especificación anterior —el que las guarda— que se
// saluda, pide la lista de herramientas y cierra.
func lineasDelServidorDeProduccion(t *testing.T) []string {
	t.Helper()

	distribuido, err := app.RegistroDeProduccion("")
	require.NoError(t, err)

	// El applet mcp se registra aquí, con una tubería como entrada, en lugar
	// del de producción, que lee la entrada estándar del proceso.
	registro := &app.Registro{}

	for _, nombre := range distribuido.Nombres() {
		if nombre == appletDelServidor {
			continue
		}

		applet, registrado := distribuido.Buscar(nombre)
		require.True(t, registrado, nombre)
		require.NoError(t, registro.Registrar(applet))
	}

	entradaLee, entradaEscribe := io.Pipe()
	salidaLee, salidaEscribe := io.Pipe()

	require.NoError(t, registro.Registrar(app.AppletMCP(app.DependenciasDeMCP{
		Entrada: entradaLee,
		Version: versionDeDesarrollo,
	})))

	// errores y codigo los escribe la gorrutina del servidor: solo se leen
	// cuando terminado está cerrado.
	var (
		errores bytes.Buffer
		codigo  int
	)

	terminado := make(chan struct{})

	go func() {
		defer close(terminado)

		codigo = app.Main([]string{"kitlegal", appletDelServidor, verboDelServidor}, registro, salidaEscribe, &errores,
			versionDeDesarrollo, commitDeDesarrollo, fechaDeDesarrollo)

		// Lo que da un proceso al terminar: el final de su salida y una entrada
		// que ya no admite escrituras.
		assert.NoError(t, salidaEscribe.Close())
		assert.NoError(t, entradaLee.Close())
	}()

	// Pase lo que pase en el test, el servidor termina antes que él.
	t.Cleanup(func() {
		assert.NoError(t, entradaEscribe.Close())
		assert.NoError(t, salidaLee.Close())
		<-terminado
	})

	ctx, cancelar := context.WithTimeout(t.Context(), esperaDelFiltro)
	defer cancelar()

	sesion, err := mcptest.Abrir(ctx, salidaLee, entradaEscribe, true)
	require.NoError(t, err, "el saludo con el servidor")

	_, err = sesion.Herramientas(ctx)
	require.NoError(t, err, "la lista de herramientas")
	require.NoError(t, sesion.Cerrar(ctx), "el cierre de la sesión")

	select {
	case <-terminado:
		require.Zerof(t, codigo, "el servidor termina con 0; salida de error: %s", errores.String())
	case <-ctx.Done():
		require.FailNow(t, "el servidor no termina")
	}

	lineas := sesion.Lineas()
	require.NotEmpty(t, lineas, "el cliente anterior guarda lo que el servidor escribe")

	return lineas
}

// unirLineas es lo que escribe quien escribe esas líneas, cada una con su
// salto.
func unirLineas(lineas []string) string {
	var unidas strings.Builder

	for _, linea := range lineas {
		unidas.WriteString(linea + "\n")
	}

	return unidas.String()
}

// lineaConElResultado es el índice de la única línea que es un mensaje cuyo
// result tiene el miembro dado.
func lineaConElResultado(t *testing.T, lineas []string, nombre string) int {
	t.Helper()

	indice := -1

	for i, linea := range lineas {
		var mensaje struct {
			Resultado map[string]json.RawMessage `json:"result"`
		}

		if json.Unmarshal([]byte(linea), &mensaje) != nil {
			continue
		}

		if _, ok := mensaje.Resultado[nombre]; ok {
			require.Equalf(t, -1, indice, "premisa: una sola línea con result.%s", nombre)

			indice = i
		}
	}

	require.NotEqualf(t, -1, indice, "premisa: una línea con result.%s", nombre)

	return indice
}

// decodificada es la línea decodificada como un objeto de JSON.
func decodificada(t *testing.T, linea string) map[string]any {
	t.Helper()

	var mensaje map[string]any
	require.NoError(t, json.Unmarshal([]byte(linea), &mensaje), "la línea es un objeto de JSON: %s", linea)

	return mensaje
}

// miembro es el objeto al que lleva la ruta de nombres dada desde el objeto
// dado.
func miembro(t *testing.T, objeto map[string]any, ruta ...string) map[string]any {
	t.Helper()

	for _, nombre := range ruta {
		siguiente, ok := objeto[nombre].(map[string]any)
		require.Truef(t, ok, "%s es un objeto", nombre)

		objeto = siguiente
	}

	return objeto
}

// propiedadEnLaLista es, en la respuesta a tools/list decodificada, la
// propiedad de parametroDelLaboratorio en el inputSchema de
// herramientaDelLaboratorio.
func propiedadEnLaLista(t *testing.T, mensaje map[string]any) map[string]any {
	t.Helper()

	herramientas, ok := miembro(t, mensaje, "result")["tools"].([]any)
	require.True(t, ok, "result.tools es una lista")

	for _, herramienta := range herramientas {
		objeto, ok := herramienta.(map[string]any)
		require.True(t, ok, "cada herramienta es un objeto")

		if objeto["name"] == herramientaDelLaboratorio {
			return miembro(t, objeto, "inputSchema", "properties", parametroDelLaboratorio)
		}
	}

	require.FailNow(t, "la lista trae "+herramientaDelLaboratorio)

	return nil
}

// descripcionEnLaLista es la description de propiedadEnLaLista.
func descripcionEnLaLista(t *testing.T, mensaje map[string]any) string {
	t.Helper()

	descripcion, ok := propiedadEnLaLista(t, mensaje)["description"].(string)
	require.True(t, ok, "la descripción es un texto")

	return descripcion
}
