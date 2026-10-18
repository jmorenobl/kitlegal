//go:build grabacion

package evals

import (
	"bytes"
	"encoding/json/v2"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/app"
	"github.com/jmorenobl/kitlegal/internal/cache"
	"github.com/jmorenobl/kitlegal/internal/httpx"
	"github.com/jmorenobl/kitlegal/internal/source/boe"
)

// valorQueGraba es el único valor de la variable de grabación que la enciende
// en internal/httpx (contrato evals-y-grabaciones §3.2 de H5; research.md D11
// de H5).
const valorQueGraba = "1"

// conjuntosQueSeGraban son los conjuntos que llena TestGrabarEvals, en el orden
// en que los graba: el de H5 y el de H9 (research D15 de H9). La raíz de
// grabación de cada uno es el directorio que lo contiene, donde está también su
// manifiesto: httpx escribe bajo ella <fuente>/, así que lo grabado queda
// exactamente donde lo leen Preparar y los tests del paquete (research.md V46
// de H5).
func conjuntosQueSeGraban() []string {
	return []string{GrabacionesDeH5, GrabacionesDeH9}
}

// TestGrabarEvals graba contra la fuente real las respuestas que necesitan las
// evals de las skills y la verificación de data/normas.yaml, de
// data/plazos.yaml y de data/cobertura.yaml, según los dos manifiestos de
// grabación —el de H5 y el de H9—, ejecutando en proceso el applet boe
// (contrato evals-y-grabaciones §3.2 de H5; FR-023, FR-074 de H5; FR-030,
// FR-031, FR-052 de H9; research.md D11 de H5 y D15 de H9):
//
//  1. cada consulta se sirve primero, en una caché temporal, desde las
//     grabaciones de H4 y de H5 con Preparar, y solo lo que da una falta se
//     graba: solo se pide a la fuente lo que no está grabado en ninguno de los
//     dos conjuntos. Lo de H5 ya está grabado, así que solo escribe en la raíz
//     de H9;
//  2. por cada entrada del manifiesto, boe buscar con su búsqueda y la elección
//     del único resultado cuyo título empieza por su prefijo; después, de esa
//     norma, boe metadatos, boe indice y boe articulo por cada bloque, todas con
//     --json. Un código distinto de 0 falla nombrando la entrada y la orden;
//  3. cada identificador resuelto y su título quedan en el registro del test,
//     de donde la persona los copia a data/normas.yaml y a las evals, y de
//     donde la tarea de data/cobertura.yaml toma el del texto refundido de la
//     Ley sobre Tráfico, que ningún fichero escribe de memoria (research S5 de
//     H9).
//
// Toca la red. Solo lo ejecutan scripts/grabar-evals.sh, a mano y en la pausa
// de la tarea [datos] del manifiesto de H5 (contrato §3.3), y el paso
// grabar_datos del workflow (scripts/workflow/grabar-datos.sh), sin modelo,
// tras el commit de la tarea del manifiesto de H9; los dos ponen la etiqueta
// grabacion y la variable de grabación. Sin la variable, o con un manifiesto
// que LeerManifiesto rechaza, falla antes de pedir nada.
func TestGrabarEvals(t *testing.T) {
	if valor := os.Getenv(httpx.VariableGrabacion); valor != valorQueGraba {
		t.Fatalf("TestGrabarEvals pide a la fuente real y graba lo que responde: solo se ejecuta con %s=%s, "+
			"desde scripts/grabar-evals.sh o scripts/workflow/grabar-datos.sh (la variable vale %q)",
			httpx.VariableGrabacion, valorQueGraba, valor)
	}

	// Los dos manifiestos se leen, y el cliente de cada raíz se construye, antes
	// de pedir nada: un manifiesto que no se lee no graba nada. Un cliente por
	// raíz, porque la raíz de grabación es del cliente, y un solo Ritmo para
	// los dos: el ritmo por sitio es de cada cliente, y uno por raíz empezaría
	// sin esperar su turno respecto del otro.
	ritmo := httpx.NuevoRitmo(boe.IntervaloEntrePeticiones)

	pendientes := make([]raizQueSeGraba, 0, len(conjuntosQueSeGraban()))

	for _, conjunto := range conjuntosQueSeGraban() {
		raiz := filepath.Dir(conjunto)
		rutaDelManifiesto := rutaDelManifiestoJuntoA(conjunto)

		contenido, err := leerFichero(rutaDelManifiesto)
		require.NoErrorf(t, err, "el manifiesto de grabación %s no se puede leer", rutaDelManifiesto)

		manifiesto, err := LeerManifiesto(contenido)
		require.NoErrorf(t, err, "el manifiesto de grabación %s", rutaDelManifiesto)

		cliente, err := httpx.New(
			httpx.ConFuente(boe.NombreDeLaFuente),
			httpx.ConRaizDeGrabacion(raiz),
			httpx.ConRitmo(ritmo),
		)
		require.NoError(t, err)

		pendientes = append(pendientes, raizQueSeGraba{raiz: raiz, manifiesto: manifiesto, cliente: cliente})
	}

	// httpx.New lee la variable de grabación una sola vez, al construir el
	// cliente, y la grabación queda fija en su cadena. httpx.Replay, en cambio,
	// rechaza construirse con la variable encendida (FR-043), y Preparar lo
	// construye para cada consulta: con ella encendida, ninguna consulta se
	// sembraría desde las grabaciones de H4, todas se pedirían a la fuente y lo
	// que H4 ya grabó se volvería a grabar en H5 con el mismo nombre (contrato
	// §3.2, paso 1, y §3.4). Por eso, ya construido el cliente de cada raíz, que
	// es el que graba, se apaga para el resto del test; t.Setenv exige que el
	// test no sea paralelo.
	t.Setenv(httpx.VariableGrabacion, "")

	for _, pendiente := range pendientes {
		t.Logf("raíz de grabación %s", pendiente.raiz)

		nuevaGrabacionDeEvals(t, pendiente.cliente).grabarNormas(t, pendiente.manifiesto)
	}
}

// raizQueSeGraba es una raíz de grabación lista para grabar: su manifiesto, ya
// leído, y el cliente que escribe bajo ella.
type raizQueSeGraba struct {
	raiz       string
	manifiesto Manifiesto
	cliente    *httpx.Cliente
}

// grabacionDeEvals es lo que comparten las invocaciones de una raíz de
// TestGrabarEvals: la caché temporal, que Preparar siembra desde las
// grabaciones de H4 y de H5, y el registro con el applet boe sobre el cliente
// que graba y esa misma caché.
type grabacionDeEvals struct {
	dirCache string
	registro *app.Registro
}

// nuevaGrabacionDeEvals crea la caché temporal y registra, en el valor cero de
// app.Registro, el applet boe con el cliente que graba y esa caché: la
// composición de Preparar con este cliente en lugar de httpx.Replay (contrato
// evals-y-grabaciones §3.2 y §5.1).
func nuevaGrabacionDeEvals(t *testing.T, cliente *httpx.Cliente) grabacionDeEvals {
	t.Helper()

	dirCache := t.TempDir()

	var registro app.Registro
	require.NoError(t, registro.Registrar(app.AppletBoe(app.DependenciasDeBoe{
		Cliente: func(*slog.Logger) (*httpx.Cliente, error) { return cliente, nil },
		Cache:   []cache.Opcion{cache.ConDirectorio(dirCache)},
	})))

	return grabacionDeEvals{dirCache: dirCache, registro: &registro}
}

// grabarNormas pasa por las entradas del manifiesto en su orden y, por cada
// una, resuelve la norma y pide sus metadatos, su índice y cada bloque (punto 2
// de TestGrabarEvals).
func (g grabacionDeEvals) grabarNormas(t *testing.T, manifiesto Manifiesto) {
	t.Helper()

	for indice, entrada := range manifiesto.Normas {
		nombre := nombrarEntrada(indice, entrada)

		norma := g.resolver(t, nombre, entrada)
		t.Logf("%s: %s «%s»", nombre, norma.Identificador, norma.Titulo)

		consultas := []Consulta{
			consultaDeBoe(verboMetadatos, norma.Identificador),
			consultaDeBoe(verboIndice, norma.Identificador),
		}
		for _, bloque := range entrada.Bloques {
			consultas = append(consultas, consultaDeBoe(verboArticulo, norma.Identificador, bloque))
		}

		for _, consulta := range consultas {
			g.pedir(t, nombre, consulta)
		}
	}
}

// resolver busca la norma de la entrada con boe buscar y devuelve el único
// resultado cuyo título empieza por su prefijo, comparado byte a byte con
// strings.HasPrefix como en LeerManifiesto. Ninguno, o más de uno, falla
// nombrando la entrada, la orden y los títulos de la búsqueda.
func (g grabacionDeEvals) resolver(t *testing.T, nombre string, entrada EntradaDelManifiesto) boe.ResultadoDeBusqueda {
	t.Helper()

	busqueda := consultaDeBoe(verboBuscar, entrada.Busqueda)

	var sobre struct {
		Data []boe.ResultadoDeBusqueda `json:"data"`
	}

	require.NoErrorf(t, json.Unmarshal(g.pedir(t, nombre, busqueda), &sobre),
		"%s: la salida de «%s» no es un sobre con los resultados de la búsqueda", nombre, ordenDe(busqueda))

	var elegidos []boe.ResultadoDeBusqueda

	titulos := make([]string, 0, len(sobre.Data))

	for _, resultado := range sobre.Data {
		titulos = append(titulos, resultado.Titulo)

		if strings.HasPrefix(resultado.Titulo, entrada.TituloEmpiezaPor) {
			elegidos = append(elegidos, resultado)
		}
	}

	if len(elegidos) != 1 {
		t.Fatalf("%s: %d resultados de «%s» tienen un título que empieza por el prefijo, y tiene que ser uno solo; "+
			"títulos: %q", nombre, len(elegidos), ordenDe(busqueda), titulos)
	}

	return elegidos[0]
}

// pedir siembra la consulta en la caché desde las grabaciones de H4 y de H5
// con Preparar y después la ejecuta con app.Main y --json sobre el cliente que
// graba, y devuelve su salida estándar. Lo que Preparar sirve ya está en la
// caché y no llega a la fuente; lo que da como falta es una consulta que no está
// grabada en ninguno de los dos conjuntos, que la invocación pide a la fuente y
// graba, o que falla con su código. Por eso las faltas no se miran, y el error
// de Preparar, que es lo que impide preparar, sí.
//
// Se siembra consulta a consulta, justo antes de invocarla, para que no venza la
// vigencia de lo sembrado mientras se piden otras a la fuente: pedirlo de nuevo
// grabaría en H5 un fichero con el nombre de uno de H4 (contrato §3.4). Un código
// distinto de 0 falla nombrando la entrada, la orden, el código y el mensaje.
func (g grabacionDeEvals) pedir(t *testing.T, nombre string, consulta Consulta) []byte {
	t.Helper()

	_, err := Preparar(g.dirCache, UnionDeGrabaciones(), []Consulta{consulta})
	require.NoErrorf(t, err, "%s: «%s» no se puede servir desde las grabaciones de H4 y de H5", nombre,
		ordenDe(consulta))

	argv := slices.Concat([]string{programaDeLasConsultas, consulta.Applet, consulta.Verbo}, consulta.Argumentos,
		[]string{"--json"})

	var salida, errores bytes.Buffer

	codigo := app.Main(argv, g.registro, &salida, &errores,
		sinDatosDeConstruccion, sinDatosDeConstruccion, sinDatosDeConstruccion)
	if codigo != 0 {
		falta := Falta{Consulta: consulta, Codigo: codigo, Mensaje: strings.TrimSuffix(errores.String(), "\n")}
		t.Fatalf("%s: %s", nombre, falta)
	}

	return salida.Bytes()
}

// consultaDeBoe es la invocación del applet boe con el verbo y los argumentos.
func consultaDeBoe(verbo string, argumentos ...string) Consulta {
	return Consulta{Applet: appletDeLasNormas, Verbo: verbo, Argumentos: argumentos}
}
