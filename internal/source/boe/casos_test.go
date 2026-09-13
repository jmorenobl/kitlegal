package boe

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/httpx"
)

// Los datos de prueba que fija una persona en la pausa de la tarea [datos] del
// manifiesto, vistos desde este paquete (contrato esquemas-fixtures-y-controles
// §3 y §5; research.md D12). Los leen los tests que reproducen la fuente y
// ninguno los escribe.
const (
	// carpetaDeLasGrabaciones es la de las grabaciones reales de la fuente,
	// <raíz de grabación>/<fuente>: la que reproduce httpx.Replay.
	carpetaDeLasGrabaciones = "testdata/" + NombreDeLaFuente
	// ficheroDelManifiesto es la lista cerrada de los recursos grabados.
	ficheroDelManifiesto = "testdata/grabaciones.json"
	// carpetaDeLasReferencias es la de las referencias del diff de aceptación,
	// un fichero por artículo (FR-116).
	carpetaDeLasReferencias = "testdata/referencias"
)

// Los recursos que admite el manifiesto (contrato §3.1).
const (
	recursoDeBusqueda  = "busqueda"
	recursoDeIndice    = "indice"
	recursoDeMetadatos = "metadatos"
	recursoDeAnalisis  = "analisis"
	recursoDeBloque    = "bloque"
)

// aceptaDelBloque y aceptaDelResto son los formatos en que la fuente pide cada
// recurso: XML el bloque y JSON todo lo demás (contrato verbos-y-salidas §1 a §6;
// refs/boe.py 406-407). La reproducción no empareja por Accept; lo que decide es
// que la grabación reciba el cuerpo en el formato que leerá la fuente.
const (
	aceptaDelBloque = "application/xml"
	aceptaDelResto  = "application/json"
)

// Los artículos del diff de aceptación, lista cerrada del contrato §5 (SC-001;
// research.md D12): cuatro fijos, uno de ellos BOE-A-2015-10565 a21, y los dos
// candidatos de los recursos 7 y 11 del manifiesto, que los suplentes de §3.1
// pueden sustituir si al grabar no se cumplía S7.
//
// El cuarto fijo, BOE-A-1992-26318 a42, es el único con avisos de vigencia: los
// otros cinco los dan vacíos, así que sin él el diff nunca compararía el campo
// donde el porte se aparta de refs/boe.py con su forma {codigo, texto}.
var (
	articulosFijosDelDiff = []articuloDelDiff{
		{norma: "BOE-A-2015-10565", bloque: "a21"},
		{norma: "BOE-A-2017-12902", bloque: "a1-30"},
		{norma: "BOE-A-2017-12902", bloque: "da-3"},
		{norma: "BOE-A-1992-26318", bloque: "a42"},
	}
	candidatosDelDiff = []articuloDelDiff{
		{norma: "BOE-A-2015-10565", bloque: "a1"},
		{norma: "BOE-A-1985-5392", bloque: "a22"},
	}
	suplentesDelDiff = []articuloDelDiff{
		{norma: "BOE-A-2015-10565", bloque: "a5"},
		{norma: "BOE-A-1985-5392", bloque: "a1"},
	}
)

// referenciasDelDiff y normasDelDiff son los seis artículos de cuatro leyes de
// FR-116, que cumplen con holgura los cinco de tres del criterio de aceptación
// del hito (docs/ROADMAP.md, H4).
const (
	referenciasDelDiff = 6
	normasDelDiff      = 4
)

// Los golden de data (contrato esquemas-fixtures-y-controles §2; FR-112): los
// escribe TestGolden con su bandera en la tarea [datos] de los golden, los
// revisa una persona en su pausa y desde entonces TestGolden los compara byte a
// byte con lo que da cada caso.
const (
	// carpetaDeLosGolden es la de los golden, un fichero por caso.
	carpetaDeLosGolden = "testdata/golden"
	// extensionDeLosGolden es la del fichero de cada caso, <caso>.json.
	extensionDeLosGolden = ".json"
)

// casoDeGolden es un caso de la lista cerrada de golden: el nombre de su fichero,
// sin la extensión, y la consulta cuyo data y url fija.
type casoDeGolden struct {
	nombre   string
	consulta core.Consulta
}

// casosDeGolden son los trece casos del contrato §2, en su orden: los seis
// verbos, cada consulta sobre recursos del manifiesto, con los seis artículos del
// diff de aceptación entre ellos (FR-112, SC-005). Cada consulta va con los
// argumentos tal como los recibe el verbo, de modo que la búsqueda son las tres
// palabras que la persona escribe.
var casosDeGolden = []casoDeGolden{
	{
		nombre:   "buscar-procedimiento-administrativo-comun",
		consulta: ConsultaBuscar{Texto: []string{"procedimiento", "administrativo", "común"}},
	},
	{nombre: "buscar-sin-resultados", consulta: ConsultaBuscar{Texto: []string{"zzqxkwvjh"}}},
	{nombre: "indice-BOE-A-2015-10565", consulta: ConsultaIndice{Norma: "BOE-A-2015-10565"}},
	{nombre: "articulo-BOE-A-2015-10565-a21", consulta: ConsultaArticulo{Norma: "BOE-A-2015-10565", Bloque: "a21"}},
	{nombre: "articulo-BOE-A-2015-10565-a1", consulta: ConsultaArticulo{Norma: "BOE-A-2015-10565", Bloque: "a1"}},
	{nombre: "articulo-BOE-A-1985-5392-a22", consulta: ConsultaArticulo{Norma: "BOE-A-1985-5392", Bloque: "a22"}},
	{nombre: "articulo-BOE-A-2017-12902-a1-30", consulta: ConsultaArticulo{Norma: "BOE-A-2017-12902", Bloque: "a1-30"}},
	{nombre: "articulo-BOE-A-2017-12902-da-3", consulta: ConsultaArticulo{Norma: "BOE-A-2017-12902", Bloque: "da-3"}},
	{nombre: "articulo-BOE-A-1992-26318-a42", consulta: ConsultaArticulo{Norma: "BOE-A-1992-26318", Bloque: "a42"}},
	{
		nombre:   "articulos-BOE-A-2015-10565-a21-a22-a23",
		consulta: ConsultaArticulos{Norma: "BOE-A-2015-10565", Bloques: []string{"a21", "a22", "a23"}},
	},
	{nombre: "metadatos-BOE-A-2015-10565", consulta: ConsultaMetadatos{Norma: "BOE-A-2015-10565"}},
	{nombre: "metadatos-BOE-A-1992-26318", consulta: ConsultaMetadatos{Norma: "BOE-A-1992-26318"}},
	{nombre: "analisis-BOE-A-2015-10565", consulta: ConsultaAnalisis{Norma: "BOE-A-2015-10565"}},
}

// manifiestoDeLasGrabaciones es grabaciones.json: la fuente y la lista cerrada de
// sus recursos, sin dirección ni Accept, que construye el código (contrato §3.1).
type manifiestoDeLasGrabaciones struct {
	Fuente   string                 `json:"fuente"`
	Recursos []recursoDelManifiesto `json:"recursos"`
}

// recursoDelManifiesto es una entrada del manifiesto: la clase del recurso, sus
// argumentos y para qué se graba.
type recursoDelManifiesto struct {
	Recurso string `json:"recurso"`
	Texto   string `json:"texto"`
	Norma   string `json:"norma"`
	Bloque  string `json:"bloque"`
	Para    string `json:"para"`
}

// articuloDelDiff es un artículo del diff de aceptación: la norma y el bloque de
// cuya lectura hay una referencia generada con refs/boe.py sobre las grabaciones
// y revisada por una persona.
type articuloDelDiff struct {
	norma  string
	bloque string
}

// referenciaDelDiff es un fichero de las referencias con la forma del contrato
// §5. Los campos van por puntero para distinguir el que falta del que está vacío;
// omitempty solo sirve a los casos de prueba que escriben una referencia sin
// alguno de ellos.
type referenciaDelDiff struct {
	Norma              string         `json:"norma"`
	Bloque             string         `json:"bloque"`
	GrabacionBloque    string         `json:"grabacion_bloque"`
	GrabacionMetadatos string         `json:"grabacion_metadatos"`
	Campos             *camposDelDiff `json:"campos,omitempty"`
}

// camposDelDiff son los ocho campos que compara el diff de aceptación (FR-116):
// los avisos, como la lista ordenada de sus textos.
type camposDelDiff struct {
	Titulo            *campoDelDiff[string]   `json:"titulo,omitempty"`
	Tipo              *campoDelDiff[string]   `json:"tipo,omitempty"`
	FechaVersion      *campoDelDiff[string]   `json:"fecha_version,omitempty"`
	FechaVigencia     *campoDelDiff[string]   `json:"fecha_vigencia,omitempty"`
	NormaModificadora *campoDelDiff[string]   `json:"norma_modificadora,omitempty"`
	Texto             *campoDelDiff[string]   `json:"texto,omitempty"`
	Avisos            *campoDelDiff[[]string] `json:"avisos,omitempty"`
	URL               *campoDelDiff[string]   `json:"url,omitempty"`
}

// campoDelDiff es un campo de la referencia: su valor y las líneas de refs/boe.py
// que lo justifican.
type campoDelDiff[T any] struct {
	Valor *T     `json:"valor,omitempty"`
	BoePy string `json:"boe_py"`
}

// TestGrabacionesCompletas exige que cada recurso del manifiesto tenga su
// grabación: httpx.Replay sobre la carpeta de las grabaciones sirve la petición
// que el código construye para él, la misma con la que se grabó (FR-113, SC-005;
// contrato esquemas-fixtures-y-controles §3.2). No mira lo que responde la
// grabación: un 404 grabado también es una respuesta servida.
func TestGrabacionesCompletas(t *testing.T) {
	t.Parallel()

	manifiesto, err := leerManifiesto(ficheroDelManifiesto)
	require.NoError(t, err)

	cliente, err := httpx.Replay(carpetaDeLasGrabaciones, httpx.ConFuente(NombreDeLaFuente))
	require.NoError(t, err)

	for indice, recurso := range manifiesto.Recursos {
		t.Run(fmt.Sprintf("%02d %s", indice+1, recurso.nombre()), func(t *testing.T) {
			t.Parallel()

			peticion, err := recurso.peticion()
			require.NoError(t, err)

			_, err = cliente.Pedir(t.Context(), schema.Contexto{}, peticion)
			require.NoErrorf(t, err, "el recurso %d del manifiesto (%s) no tiene grabación", indice+1, recurso.Para)
		})
	}
}

// TestReferenciasCompletas exige las seis referencias del diff de aceptación tal
// como las deja una persona en la pausa del manifiesto (FR-116, SC-001, SC-005;
// contrato esquemas-fixtures-y-controles §5): de cuatro normas, una de ellas
// BOE-A-2015-10565-a21, de la lista cerrada o con sus suplentes, sin claves
// desconocidas, con los ocho campos y la línea de refs/boe.py de cada uno, y con
// grabaciones que existen. No compara ningún valor: eso lo hace
// TestArticuloCoincideConBoePy.
//
// El primer subtest comprueba las referencias reales; el resto demuestra sobre
// carpetas temporales que la comprobación no pasa en vacío.
func TestReferenciasCompletas(t *testing.T) {
	t.Parallel()

	t.Run("testdata-referencias", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, comprobarReferencias(carpetaDeLasReferencias, carpetaDeLasGrabaciones))
	})

	const primera = "BOE-A-2015-10565-a21.json: "

	delContrato := slices.Concat(articulosFijosDelDiff, candidatosDelDiff)

	casos := []struct {
		nombre    string
		articulos []articuloDelDiff
		// cambia estropea la referencia del primer artículo antes de escribirla, y
		// texto, el JSON que se escribe.
		cambia func(referencia *referenciaDelDiff)
		texto  func(contenido string) string
		// ajeno es un fichero que se deja además en la carpeta de las referencias.
		ajeno      string
		sinCarpeta bool
		// mensaje es un fragmento del error esperado; vacío, las referencias valen.
		mensaje string
	}{
		{nombre: "las-del-contrato", articulos: delContrato},
		{nombre: "con-los-dos-suplentes", articulos: slices.Concat(articulosFijosDelDiff, suplentesDelDiff)},
		{nombre: "sin-carpeta", sinCarpeta: true, mensaje: "la carpeta de las referencias no se puede leer"},
		{
			nombre:    "cinco",
			articulos: slices.Concat(articulosFijosDelDiff, candidatosDelDiff[:1]),
			mensaje:   "hay 5 referencias y tiene que haber 6",
		},
		{
			nombre:    "siete",
			articulos: slices.Concat(articulosFijosDelDiff, candidatosDelDiff, suplentesDelDiff[:1]),
			mensaje:   "hay 7 referencias y tiene que haber 6",
		},
		{
			nombre:    "sin-la-del-articulo-21",
			articulos: slices.Concat(articulosFijosDelDiff[1:], candidatosDelDiff, suplentesDelDiff[:1]),
			mensaje:   "falta la referencia BOE-A-2015-10565-a21.json",
		},
		{
			nombre:    "de-tres-normas",
			articulos: slices.Concat(articulosFijosDelDiff, candidatosDelDiff[:1], suplentesDelDiff[:1]),
			mensaje:   "son de 3 normas y tienen que ser de 4",
		},
		{
			nombre:    "fuera-de-la-lista",
			articulos: delContrato,
			ajeno:     "BOE-A-2015-10565-a7.json",
			mensaje:   "BOE-A-2015-10565-a7.json no es la referencia de ningún artículo del diff",
		},
		{
			nombre:    "fichero-que-no-es-una-referencia",
			articulos: delContrato,
			ajeno:     "LEEME.md",
			mensaje:   "LEEME.md no es la referencia de ningún artículo del diff",
		},
		{
			nombre:    "clave-desconocida",
			articulos: delContrato,
			texto: func(contenido string) string {
				return strings.Replace(contenido, `{"norma":`, `{"nota":"","norma":`, 1)
			},
			mensaje: `BOE-A-2015-10565-a21.json no tiene la forma esperada: json: unknown field "nota"`,
		},
		{
			nombre:    "clave-desconocida-en-un-campo",
			articulos: delContrato,
			texto: func(contenido string) string {
				return strings.Replace(contenido, `"boe_py":`, `"linea":"","boe_py":`, 1)
			},
			mensaje: `BOE-A-2015-10565-a21.json no tiene la forma esperada: json: unknown field "linea"`,
		},
		{
			nombre:    "de-otro-articulo",
			articulos: delContrato,
			cambia:    func(referencia *referenciaDelDiff) { referencia.Bloque = "a22" },
			mensaje:   primera + "es de BOE-A-2015-10565 a22 y su nombre dice BOE-A-2015-10565 a21",
		},
		{
			nombre:    "grabacion-que-no-existe",
			articulos: delContrato,
			cambia:    func(referencia *referenciaDelDiff) { referencia.GrabacionBloque = "GET_otra.json" },
			mensaje:   primera + "grabacion_bloque nombra GET_otra.json, que no es ninguna grabación",
		},
		{
			nombre:    "grabacion-con-carpeta",
			articulos: delContrato,
			cambia: func(referencia *referenciaDelDiff) {
				referencia.GrabacionMetadatos = filepath.Join("..", "grabaciones", grabacionDePruebaDeLosMetadatos)
			},
			mensaje: primera + "grabacion_metadatos es ",
		},
		{
			nombre:    "sin-campos",
			articulos: delContrato,
			cambia:    func(referencia *referenciaDelDiff) { referencia.Campos = nil },
			mensaje:   primera + "no tiene campos",
		},
		{
			nombre:    "sin-un-campo",
			articulos: delContrato,
			cambia:    func(referencia *referenciaDelDiff) { referencia.Campos.Texto = nil },
			mensaje:   primera + "falta el campo texto",
		},
		{
			nombre:    "campo-sin-valor",
			articulos: delContrato,
			cambia:    func(referencia *referenciaDelDiff) { referencia.Campos.Avisos.Valor = nil },
			mensaje:   primera + "el campo avisos no tiene valor",
		},
		{
			nombre:    "boe-py-vacio",
			articulos: delContrato,
			cambia:    func(referencia *referenciaDelDiff) { referencia.Campos.URL.BoePy = " " },
			mensaje:   primera + "el campo url no dice en boe_py",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			raiz := t.TempDir()
			grabaciones := filepath.Join(raiz, "grabaciones")
			referencias := filepath.Join(raiz, "referencias")

			for _, grabacion := range []string{grabacionDePruebaDelBloque, grabacionDePruebaDeLosMetadatos} {
				escribeDatoDePrueba(t, filepath.Join(grabaciones, grabacion), []byte("{}"))
			}

			if !caso.sinCarpeta {
				escribeReferenciasDePrueba(t, referencias, caso.articulos, caso.cambia, caso.texto)
			}

			if caso.ajeno != "" {
				escribeDatoDePrueba(t, filepath.Join(referencias, caso.ajeno), []byte("{}"))
			}

			err := comprobarReferencias(referencias, grabaciones)
			if caso.mensaje == "" {
				require.NoError(t, err)

				return
			}

			require.ErrorContains(t, err, caso.mensaje)
		})
	}
}

// leerManifiesto lee el manifiesto de las grabaciones y exige que sea de esta
// fuente y que liste algún recurso.
func leerManifiesto(ruta string) (manifiestoDeLasGrabaciones, error) {
	var manifiesto manifiestoDeLasGrabaciones
	if err := leerDatoDePrueba(ruta, &manifiesto); err != nil {
		return manifiestoDeLasGrabaciones{}, err
	}

	if manifiesto.Fuente != NombreDeLaFuente {
		return manifiestoDeLasGrabaciones{}, fmt.Errorf("%s es de la fuente %q y este paquete es %q",
			ruta, manifiesto.Fuente, NombreDeLaFuente)
	}

	if len(manifiesto.Recursos) == 0 {
		return manifiestoDeLasGrabaciones{}, fmt.Errorf("%s no lista ningún recurso", ruta)
	}

	return manifiesto, nil
}

// nombre es el del subtest del recurso: su clase y sus argumentos.
func (recurso recursoDelManifiesto) nombre() string {
	partes := []string{recurso.Recurso}

	for _, argumento := range []string{recurso.Norma, recurso.Bloque, recurso.Texto} {
		if argumento != "" {
			partes = append(partes, argumento)
		}
	}

	return strings.Join(partes, " ")
}

// peticion es la petición GET del recurso, la misma con la que se graba y con la
// que se reproduce: exige los argumentos de su clase y ningún otro, los valida
// como la fuente y construye la dirección con direcciones.go o busqueda.go.
func (recurso recursoDelManifiesto) peticion() (httpx.Peticion, error) {
	switch recurso.Recurso {
	case recursoDeBusqueda:
		if recurso.Norma != "" || recurso.Bloque != "" {
			return httpx.Peticion{}, errors.New("una búsqueda solo lleva texto")
		}

		direccion, err := direccionDeBusqueda(recurso.Texto)
		if err != nil {
			return httpx.Peticion{}, err
		}

		return httpx.Peticion{Metodo: "GET", URL: direccion, Acepta: aceptaDelResto}, nil

	case recursoDeIndice, recursoDeMetadatos, recursoDeAnalisis:
		if recurso.Texto != "" || recurso.Bloque != "" {
			return httpx.Peticion{}, errors.New("un índice, unos metadatos o un análisis solo llevan norma")
		}

		if err := ValidarNorma(recurso.Norma); err != nil {
			return httpx.Peticion{}, err
		}

		direccion := map[string]func(string) string{
			recursoDeIndice:    direccionDelIndice,
			recursoDeMetadatos: direccionDeLosMetadatos,
			recursoDeAnalisis:  direccionDelAnalisis,
		}[recurso.Recurso](recurso.Norma)

		return httpx.Peticion{Metodo: "GET", URL: direccion, Acepta: aceptaDelResto}, nil

	case recursoDeBloque:
		if recurso.Texto != "" {
			return httpx.Peticion{}, errors.New("un bloque solo lleva norma y bloque")
		}

		if err := errors.Join(ValidarNorma(recurso.Norma), ValidarBloque(recurso.Bloque)); err != nil {
			return httpx.Peticion{}, err
		}

		return httpx.Peticion{
			Metodo: "GET",
			URL:    direccionDelBloque(recurso.Norma, recurso.Bloque),
			Acepta: aceptaDelBloque,
		}, nil

	default:
		return httpx.Peticion{}, fmt.Errorf("recurso desconocido %q: el manifiesto admite %s, %s, %s, %s y %s",
			recurso.Recurso, recursoDeBusqueda, recursoDeIndice, recursoDeMetadatos, recursoDeAnalisis, recursoDeBloque)
	}
}

// fichero es el nombre del fichero de la referencia del artículo,
// <norma>-<bloque>.json.
func (articulo articuloDelDiff) fichero() string {
	return articulo.norma + "-" + articulo.bloque + ".json"
}

// comprobarReferencias exige en la carpeta de las referencias exactamente las del
// diff de aceptación: seis ficheros de artículos de la lista cerrada y nada más,
// los cuatro fijos entre ellos, de cuatro normas; y que cada uno tenga la forma
// del contrato y nombre grabaciones de la carpeta de las grabaciones.
func comprobarReferencias(referencias, grabaciones string) error {
	articulos, err := articulosDeLasReferencias(referencias)
	if err != nil {
		return err
	}

	if err := comprobarArticulosDelDiff(articulos); err != nil {
		return fmt.Errorf("%s: %w", referencias, err)
	}

	fallos := make([]error, 0, len(articulos))
	for _, articulo := range articulos {
		fallos = append(fallos, comprobarReferencia(referencias, grabaciones, articulo))
	}

	return errors.Join(fallos...)
}

// articulosDeLasReferencias devuelve el artículo de cada fichero de la carpeta de
// las referencias. Lo que no es el fichero de un artículo de la lista cerrada o
// de un suplente es un error que lo nombra.
func articulosDeLasReferencias(carpeta string) ([]articuloDelDiff, error) {
	entradas, err := os.ReadDir(filepath.Clean(carpeta))
	if err != nil {
		return nil, fmt.Errorf("la carpeta de las referencias no se puede leer: %w", err)
	}

	admitidos := slices.Concat(articulosFijosDelDiff, candidatosDelDiff, suplentesDelDiff)
	articulos := make([]articuloDelDiff, 0, len(entradas))

	var ajenos []error

	for _, entrada := range entradas {
		indice := slices.IndexFunc(admitidos, func(articulo articuloDelDiff) bool {
			return articulo.fichero() == entrada.Name()
		})
		if indice < 0 || !entrada.Type().IsRegular() {
			ajenos = append(ajenos, fmt.Errorf("%s: %s no es la referencia de ningún artículo del diff ni de un suplente",
				carpeta, entrada.Name()))

			continue
		}

		articulos = append(articulos, admitidos[indice])
	}

	if err := errors.Join(ajenos...); err != nil {
		return nil, err
	}

	return articulos, nil
}

// comprobarArticulosDelDiff exige que los artículos sean seis, con los cuatro
// fijos entre ellos, y de cuatro normas.
func comprobarArticulosDelDiff(articulos []articuloDelDiff) error {
	var fallos []error

	if len(articulos) != referenciasDelDiff {
		fallos = append(fallos, fmt.Errorf("hay %d referencias y tiene que haber %d", len(articulos), referenciasDelDiff))
	}

	for _, fijo := range articulosFijosDelDiff {
		if !slices.Contains(articulos, fijo) {
			fallos = append(fallos, fmt.Errorf("falta la referencia %s", fijo.fichero()))
		}
	}

	normas := make(map[string]bool, len(articulos))
	for _, articulo := range articulos {
		normas[articulo.norma] = true
	}

	if len(normas) != normasDelDiff {
		fallos = append(fallos, fmt.Errorf("las referencias son de %d normas y tienen que ser de %d",
			len(normas), normasDelDiff))
	}

	return errors.Join(fallos...)
}

// comprobarReferencia lee la referencia de un artículo con la forma del contrato,
// sin claves desconocidas, y exige que sea de ese artículo, que sus dos
// grabaciones existan y que tenga los ocho campos.
func comprobarReferencia(referencias, grabaciones string, articulo articuloDelDiff) error {
	ruta := filepath.Join(referencias, articulo.fichero())

	var referencia referenciaDelDiff
	if err := leerDatoDePrueba(ruta, &referencia); err != nil {
		return err
	}

	var deOtroArticulo error
	if referencia.Norma != articulo.norma || referencia.Bloque != articulo.bloque {
		deOtroArticulo = fmt.Errorf("es de %s %s y su nombre dice %s %s",
			referencia.Norma, referencia.Bloque, articulo.norma, articulo.bloque)
	}

	if err := errors.Join(
		deOtroArticulo,
		comprobarGrabacionNombrada(grabaciones, "grabacion_bloque", referencia.GrabacionBloque),
		comprobarGrabacionNombrada(grabaciones, "grabacion_metadatos", referencia.GrabacionMetadatos),
		referencia.Campos.comprobar(),
	); err != nil {
		return fmt.Errorf("%s: %w", ruta, err)
	}

	return nil
}

// comprobarGrabacionNombrada exige que la clave nombre, sin carpeta, un fichero de
// la carpeta de las grabaciones.
func comprobarGrabacionNombrada(grabaciones, clave, nombre string) error {
	if nombre == "" || filepath.Base(nombre) != nombre {
		return fmt.Errorf("%s es %q, que no es el nombre de un fichero de grabación", clave, nombre)
	}

	fichero, err := os.Stat(filepath.Join(grabaciones, nombre))
	if err != nil {
		return fmt.Errorf("%s nombra %s, que no es ninguna grabación: %w", clave, nombre, err)
	}

	if !fichero.Mode().IsRegular() {
		return fmt.Errorf("%s nombra %s, que no es un fichero de grabación", clave, nombre)
	}

	return nil
}

// comprobar exige los ocho campos del diff, cada uno con su valor y con las
// líneas de refs/boe.py que lo justifican.
func (campos *camposDelDiff) comprobar() error {
	if campos == nil {
		return errors.New("no tiene campos")
	}

	return errors.Join(
		campos.Titulo.comprobar("titulo"),
		campos.Tipo.comprobar("tipo"),
		campos.FechaVersion.comprobar("fecha_version"),
		campos.FechaVigencia.comprobar("fecha_vigencia"),
		campos.NormaModificadora.comprobar("norma_modificadora"),
		campos.Texto.comprobar("texto"),
		campos.Avisos.comprobar("avisos"),
		campos.URL.comprobar("url"),
	)
}

// comprobar exige que el campo esté, que tenga valor y que diga en boe_py qué
// líneas de refs/boe.py lo justifican. No mira el valor.
func (campo *campoDelDiff[T]) comprobar(nombre string) error {
	switch {
	case campo == nil:
		return fmt.Errorf("falta el campo %s", nombre)
	case campo.Valor == nil:
		return fmt.Errorf("el campo %s no tiene valor", nombre)
	case strings.TrimSpace(campo.BoePy) == "":
		return fmt.Errorf("el campo %s no dice en boe_py qué líneas de refs/boe.py lo justifican", nombre)
	default:
		return nil
	}
}

// leerDatoDePrueba lee en destino un fichero JSON de datos de prueba: la ruta pasa
// por filepath.Clean antes de abrirse (gosec G304), y no se admiten claves que
// destino no declare ni nada detrás del valor.
func leerDatoDePrueba(ruta string, destino any) error {
	contenido, err := os.ReadFile(filepath.Clean(ruta))
	if err != nil {
		return fmt.Errorf("%s no se puede leer: %w", ruta, err)
	}

	decodificador := json.NewDecoder(bytes.NewReader(contenido))
	decodificador.DisallowUnknownFields()

	if err := decodificador.Decode(destino); err != nil {
		return fmt.Errorf("%s no tiene la forma esperada: %w", ruta, err)
	}

	if _, err := decodificador.Token(); !errors.Is(err, io.EOF) {
		return fmt.Errorf("%s lleva algo detrás de su valor", ruta)
	}

	return nil
}

// Las grabaciones que nombran las referencias de los casos de prueba de
// TestReferenciasCompletas.
const (
	grabacionDePruebaDelBloque      = "GET_bloque.json"
	grabacionDePruebaDeLosMetadatos = "GET_metadatos.json"
)

// escribeReferenciasDePrueba deja en la carpeta una referencia válida por
// artículo; cambia estropea la del primero antes de serializarla, y texto, su
// JSON.
func escribeReferenciasDePrueba(t *testing.T, carpeta string, articulos []articuloDelDiff,
	cambia func(*referenciaDelDiff), texto func(string) string,
) {
	t.Helper()

	for indice, articulo := range articulos {
		referencia := referenciaDePrueba(articulo)
		if indice == 0 && cambia != nil {
			cambia(&referencia)
		}

		contenido, err := json.Marshal(referencia)
		require.NoError(t, err)

		if indice == 0 && texto != nil {
			contenido = []byte(texto(string(contenido)))
		}

		escribeDatoDePrueba(t, filepath.Join(carpeta, articulo.fichero()), contenido)
	}
}

// referenciaDePrueba es una referencia con la forma del contrato para un artículo,
// con valores que no pretenden ser los de ninguna norma.
func referenciaDePrueba(articulo articuloDelDiff) referenciaDelDiff {
	campo := func(valor, lineas string) *campoDelDiff[string] {
		return &campoDelDiff[string]{Valor: &valor, BoePy: lineas}
	}

	sinAvisos := []string{}

	return referenciaDelDiff{
		Norma:              articulo.norma,
		Bloque:             articulo.bloque,
		GrabacionBloque:    grabacionDePruebaDelBloque,
		GrabacionMetadatos: grabacionDePruebaDeLosMetadatos,
		Campos: &camposDelDiff{
			Titulo:            campo("Artículo de prueba", "106"),
			Tipo:              campo("articulo", "107"),
			FechaVersion:      campo("20151002", "116"),
			FechaVigencia:     campo("20161002", "117"),
			NormaModificadora: campo("", "118"),
			Texto:             campo("Texto de prueba.", "120, 132-136"),
			Avisos:            &campoDelDiff[[]string]{Valor: &sinAvisos, BoePy: "197-211"},
			URL:               campo(direccionPublicaDelBloque(articulo.norma, articulo.bloque), "430"),
		},
	}
}

// escribeDatoDePrueba crea un fichero de una carpeta temporal, y su carpeta, solo
// para su dueño (gosec G301 y G306), desde un auxiliar distinto del que lee (G703).
func escribeDatoDePrueba(t *testing.T, ruta string, contenido []byte) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(ruta), 0o750))
	require.NoError(t, os.WriteFile(ruta, contenido, 0o600))
}
