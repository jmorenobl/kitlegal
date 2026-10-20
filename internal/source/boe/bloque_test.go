package boe

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
	"github.com/jmorenobl/kitlegal/internal/httpx"
)

// Envoltorio de la respuesta de un bloque tal como lo entrega la API (la
// grabación de a21): el estado va fuera del bloque y su texto no entra en el
// del artículo.
const (
	cabeceraDeLaRespuesta = "<?xml version=\"1.0\" encoding=\"utf-8\"?>\n<response>\n  <status>\n" +
		"    <code>200</code>\n    <text>ok</text>\n  </status>\n  <data>\n"
	pieDeLaRespuesta = "\n  </data>\n</response>\n"
)

// TestLeerBloque fija la lectura del XML de un bloque con el recorrido de
// ElementTree (data-model.md §3.2, research.md D6): el cuerpo es UTF-8 válido
// (X1) y la codificación que declara no cuenta (X2); tiene un único elemento raíz
// (X3); el bloque es el primer descendiente bloque sin espacio de nombres que no
// es la raíz (X4); título y tipo son sus atributos, y fecha, vigencia y norma
// modificadora, los de la última versión de primer nivel, con la fecha original
// si falta o no hay versiones (X5, X6; FR-010, FR-011); el texto junta CharData,
// CDATA y la cola del elemento sin que comentarios ni instrucciones de proceso lo
// corten (X7); y las líneas se recortan con el espacio en blanco de Python y se
// descartan las vacías (X8).
//
// Además guarda todas las versiones de primer nivel, en el orden de la
// respuesta, cada una con los campos y el texto que la lectura da hoy de la
// última (H20, data-model.md §2; research.md D4 y M1): las de los tres bloques
// grabados de H4, con sus fechas y sus normas escritas aquí, y ninguna en un
// bloque sin versiones (FR-002, FR-010, FR-012).
//
// Todo lo que no se puede leer es «fuente no disponible» con un mensaje literal
// que dice qué no se pudo interpretar y sin dirección ni instante, que pone quien
// pidió al envolverlo (contrato errores-y-codigos, fila 16; FR-014): nunca el
// cuerpo ni un recorte suyo. Las entradas son XML escrito en la tabla.
func TestLeerBloque(t *testing.T) {
	t.Parallel()

	t.Run("lo que se lee del bloque", func(t *testing.T) {
		t.Parallel()

		casos := []struct {
			nombre string
			cuerpo string
			bloque bloqueLeido
		}{
			{
				nombre: "con versiones, la última es la vigente, con sus atributos y solo su texto",
				cuerpo: respuestaConBloque(`    <bloque id="a21" tipo="precepto" titulo="Artículo 21">
      <version id_norma="BOE-A-2015-10565" fecha_publicacion="20151002" fecha_vigencia="20161002">
        <p class="articulo">Artículo 21. Obligación de resolver.</p>
        <p class="parrafo">Texto original.</p>
      </version>
      <version id_norma="BOE-A-2018-16673" fecha_publicacion="20181206" fecha_vigencia="20181207">
        <p class="articulo">Artículo 21. Obligación de resolver.</p>
        <p class="parrafo">Texto modificado.</p>
      </version>
    </bloque>`),
				bloque: bloqueLeido{
					titulo:            "Artículo 21",
					tipo:              "precepto",
					fechaVersion:      "20181206",
					fechaVigencia:     "20181207",
					normaModificadora: "BOE-A-2018-16673",
					texto:             "Artículo 21. Obligación de resolver.\nTexto modificado.",
					versiones: []versionLeida{
						{
							fechaVersion:      "20151002",
							fechaVigencia:     "20161002",
							normaModificadora: "BOE-A-2015-10565",
							texto:             "Artículo 21. Obligación de resolver.\nTexto original.",
						},
						{
							fechaVersion:      "20181206",
							fechaVigencia:     "20181207",
							normaModificadora: "BOE-A-2018-16673",
							texto:             "Artículo 21. Obligación de resolver.\nTexto modificado.",
						},
					},
				},
			},
			{
				nombre: "sin versiones, el texto de todo el bloque, la fecha original y lo demás vacío",
				cuerpo: respuestaConBloque(`    <bloque id="pr" tipo="preambulo" titulo="Preámbulo">
      <p class="parrafo">Primer párrafo.</p>
      <p class="parrafo">Segundo párrafo.</p>
    </bloque>`),
				bloque: bloqueLeido{
					titulo:       "Preámbulo",
					tipo:         "preambulo",
					fechaVersion: "original",
					texto:        "Primer párrafo.\nSegundo párrafo.",
				},
			},
			{
				nombre: "la última versión sin fecha de publicación da la fecha original",
				cuerpo: respuestaConBloque(`<bloque tipo="precepto" titulo="Artículo 1">
<version id_norma="BOE-A-2015-10565" fecha_publicacion="20151002" fecha_vigencia="20161002"><p>Antigua.</p></version>
<version id_norma="BOE-A-2018-16673" fecha_vigencia="20181207"><p>Vigente.</p></version>
</bloque>`),
				bloque: bloqueLeido{
					titulo:            "Artículo 1",
					tipo:              "precepto",
					fechaVersion:      "original",
					fechaVigencia:     "20181207",
					normaModificadora: "BOE-A-2018-16673",
					texto:             "Vigente.",
					versiones: []versionLeida{
						{fechaVersion: "20151002", fechaVigencia: "20161002", normaModificadora: "BOE-A-2015-10565", texto: "Antigua."},
						{fechaVersion: "original", fechaVigencia: "20181207", normaModificadora: "BOE-A-2018-16673", texto: "Vigente."},
					},
				},
			},
			{
				nombre: "la última versión sin atributos da la fecha original y lo demás vacío",
				cuerpo: respuestaConBloque(`<bloque id="a1"><version><p>Texto.</p></version></bloque>`),
				bloque: bloqueLeido{
					fechaVersion: "original",
					texto:        "Texto.",
					versiones:    []versionLeida{{fechaVersion: "original", texto: "Texto."}},
				},
			},
			{
				nombre: "una fecha de publicación vacía presente es la fecha, no la original",
				cuerpo: respuestaConBloque(`<bloque titulo="Artículo 1">
<version id_norma="BOE-A-2015-10565" fecha_publicacion="" fecha_vigencia=""><p>Texto.</p></version>
</bloque>`),
				bloque: bloqueLeido{
					titulo:            "Artículo 1",
					normaModificadora: "BOE-A-2015-10565",
					texto:             "Texto.",
					versiones:         []versionLeida{{normaModificadora: "BOE-A-2015-10565", texto: "Texto."}},
				},
			},
			{
				nombre: "la cola de la última versión entra en su texto, también tras un comentario",
				cuerpo: respuestaConBloque(`<bloque titulo="T">
<version fecha_publicacion="20151002"><p>Dentro.</p>
</version>
Cola de la versión.
<!-- nota -->
sigue la cola.
</bloque>`),
				bloque: bloqueLeido{
					titulo:       "T",
					fechaVersion: "20151002",
					texto:        "Dentro.\nCola de la versión.\nsigue la cola.",
					versiones: []versionLeida{
						{fechaVersion: "20151002", texto: "Dentro.\nCola de la versión.\nsigue la cola."},
					},
				},
			},
			{
				nombre: "la cola de una versión anterior entra en el suyo y no en el de la vigente",
				cuerpo: respuestaConBloque(`<bloque titulo="T">
<version fecha_publicacion="20151002"><p>Antigua.</p></version>
Entre versiones.
<version fecha_publicacion="20181206"><p>Vigente.</p></version>
</bloque>
Cola del bloque.`),
				bloque: bloqueLeido{
					titulo:       "T",
					fechaVersion: "20181206",
					texto:        "Vigente.",
					versiones: []versionLeida{
						{fechaVersion: "20151002", texto: "Antigua.\nEntre versiones."},
						{fechaVersion: "20181206", texto: "Vigente."},
					},
				},
			},
			{
				nombre: "sin versiones, la cola del bloque entra en su texto hasta la etiqueta siguiente",
				cuerpo: respuestaConBloque(`<bloque titulo="T">
<p>Dentro.</p>
</bloque>
Cola del bloque.
<otro>No entra.</otro>`),
				bloque: bloqueLeido{titulo: "T", fechaVersion: "original", texto: "Dentro.\nCola del bloque."},
			},
			{
				nombre: "CRLF y retorno suelto son saltos de línea",
				cuerpo: cabeceraDeLaRespuesta + "<bloque titulo=\"T\">\r\n<version fecha_publicacion=\"20151002\">\r\n" +
					"<p>Uno.</p>\r\n<p>Dos.\rTres.</p>\r\n</version>\r\n</bloque>" + pieDeLaRespuesta,
				bloque: bloqueLeido{
					titulo:       "T",
					fechaVersion: "20151002",
					texto:        "Uno.\nDos.\nTres.",
					versiones:    []versionLeida{{fechaVersion: "20151002", texto: "Uno.\nDos.\nTres."}},
				},
			},
			{
				nombre: "CDATA aporta su contenido literal, unido al texto que lo rodea",
				cuerpo: respuestaConBloque(`<bloque titulo="T"><version fecha_publicacion="20151002">
<p><![CDATA[a < b && c > d]]></p>
<p>Antes <![CDATA[<dentro>]]> después</p>
</version></bloque>`),
				bloque: bloqueLeido{
					titulo:       "T",
					fechaVersion: "20151002",
					texto:        "a < b && c > d\nAntes <dentro> después",
					versiones:    []versionLeida{{fechaVersion: "20151002", texto: "a < b && c > d\nAntes <dentro> después"}},
				},
			},
			{
				nombre: "un comentario no aporta texto ni corta el que lo rodea",
				cuerpo: respuestaConBloque(`<bloque titulo="T"><version fecha_publicacion="20151002">` +
					`<p>Obliga<!-- corte -->ción de resolver.</p></version></bloque>`),
				bloque: bloqueLeido{
					titulo:       "T",
					fechaVersion: "20151002",
					texto:        "Obligación de resolver.",
					versiones:    []versionLeida{{fechaVersion: "20151002", texto: "Obligación de resolver."}},
				},
			},
			{
				nombre: "una instrucción de proceso no aporta texto ni corta el que la rodea",
				cuerpo: respuestaConBloque(`<bloque titulo="T"><version fecha_publicacion="20151002">` +
					`<p>Obliga<?proceso algo?>ción de resolver.</p></version></bloque>`),
				bloque: bloqueLeido{
					titulo:       "T",
					fechaVersion: "20151002",
					texto:        "Obligación de resolver.",
					versiones:    []versionLeida{{fechaVersion: "20151002", texto: "Obligación de resolver."}},
				},
			},
			{
				nombre: "las entidades y las referencias de carácter se resuelven en atributos y texto",
				cuerpo: respuestaConBloque(`<bloque tipo="precepto" titulo="Ley &quot;39&quot; &amp; &apos;40&apos;">` +
					`<version fecha_publicacion="20151002"><p>&lt;a&gt; &amp; &#233;&#xE9; &#x1F4DC;</p></version></bloque>`),
				bloque: bloqueLeido{
					titulo:       `Ley "39" & '40'`,
					tipo:         "precepto",
					fechaVersion: "20151002",
					texto:        "<a> & éé \U0001F4DC",
					versiones:    []versionLeida{{fechaVersion: "20151002", texto: "<a> & éé \U0001F4DC"}},
				},
			},
			{
				nombre: "tabulador, salto y retorno literales de un atributo pasan a espacio, sin recortarlo",
				cuerpo: respuestaConBloque("<bloque tipo=\"pre\tcepto\" titulo=\" Artículo\n\t21\r\nbis\rter \">" +
					"<version fecha_publicacion=\"2015\n1002\" fecha_vigencia='2016\t1002' id_norma=\"BOE-A-2015-\r\n10565\">" +
					"<p>Texto.</p></version></bloque>"),
				bloque: bloqueLeido{
					titulo:            " Artículo  21 bis ter ",
					tipo:              "pre cepto",
					fechaVersion:      "2015 1002",
					fechaVigencia:     "2016 1002",
					normaModificadora: "BOE-A-2015- 10565",
					texto:             "Texto.",
					versiones: []versionLeida{
						{fechaVersion: "2015 1002", fechaVigencia: "2016 1002", normaModificadora: "BOE-A-2015- 10565", texto: "Texto."},
					},
				},
			},
			{
				nombre: "las referencias de carácter a tabulador, salto y retorno de un atributo se conservan",
				cuerpo: respuestaConBloque(`<bloque titulo="a&#10;b&#9;c&#13;d" tipo='x&#xA;"y"'>` +
					`<version fecha_publicacion="20151002"><p>Texto.</p></version></bloque>`),
				bloque: bloqueLeido{
					titulo:       "a\nb\tc\rd",
					tipo:         "x\n\"y\"",
					fechaVersion: "20151002",
					texto:        "Texto.",
					versiones:    []versionLeida{{fechaVersion: "20151002", texto: "Texto."}},
				},
			},
			{
				nombre: "un bloque anidado en otro: cuenta el primero, y las versiones del interior no son suyas",
				cuerpo: respuestaConBloque(`    <bloque id="t1" tipo="encabezado" titulo="TÍTULO I">
      <p>Encabezado.</p>
      <bloque id="a1" tipo="precepto" titulo="Artículo 1">
        <version fecha_publicacion="20151002"><p>Artículo interior.</p></version>
      </bloque>
    </bloque>`),
				bloque: bloqueLeido{
					titulo:       "TÍTULO I",
					tipo:         "encabezado",
					fechaVersion: "original",
					texto:        "Encabezado.\nArtículo interior.",
				},
			},
			{
				nombre: "una versión que no es de primer nivel no cuenta",
				cuerpo: respuestaConBloque(`<bloque titulo="T"><div>
<version fecha_publicacion="20151002"><p>Anidada.</p></version>
</div></bloque>`),
				bloque: bloqueLeido{titulo: "T", fechaVersion: "original", texto: "Anidada."},
			},
			{
				nombre: "raíz bloque con otro bloque debajo: cuenta el de debajo",
				cuerpo: `<bloque titulo="Raíz"><bloque titulo="Hijo"><p>Texto.</p></bloque></bloque>`,
				bloque: bloqueLeido{titulo: "Hijo", fechaVersion: "original", texto: "Texto."},
			},
			{
				nombre: "una versión o un atributo con espacio de nombres no cuentan",
				cuerpo: respuestaConBloque(`<bloque xmlns:x="urn:x" x:titulo="No." titulo="Sí.">` +
					`<x:version fecha_publicacion="20151002"><p>Texto.</p></x:version></bloque>`),
				bloque: bloqueLeido{titulo: "Sí.", fechaVersion: "original", texto: "Texto."},
			},
			{
				nombre: "líneas recortadas con el espacio en blanco de Python, sin las vacías y con el interior intacto",
				cuerpo: respuestaConBloque("<bloque titulo=\"T\"><version fecha_publicacion=\"20151002\">\n" +
					"<p>\u00a0 Uno   y\tuno.\u3000</p>\n\n \t \n<p>\u2003Dos.\u0085</p>\n</version></bloque>"),
				bloque: bloqueLeido{
					titulo:       "T",
					fechaVersion: "20151002",
					texto:        "Uno   y\tuno.\nDos.",
					versiones:    []versionLeida{{fechaVersion: "20151002", texto: "Uno   y\tuno.\nDos."}},
				},
			},
			{
				nombre: "la codificación ISO-8859-1 declarada no cuenta: el cuerpo se lee como UTF-8",
				cuerpo: "<?xml version=\"1.0\" encoding=\"ISO-8859-1\"?>\n<response><data>" +
					"<bloque titulo=\"Artículo 1\"><version fecha_publicacion=\"20151002\"><p>Obligación.</p></version></bloque>" +
					"</data></response>",
				bloque: bloqueLeido{
					titulo:       "Artículo 1",
					fechaVersion: "20151002",
					texto:        "Obligación.",
					versiones:    []versionLeida{{fechaVersion: "20151002", texto: "Obligación."}},
				},
			},
			{
				nombre: "una codificación desconocida declarada tampoco cuenta",
				cuerpo: "<?xml version=\"1.0\" encoding=\"x-desconocida\"?><response><data>" +
					"<bloque titulo=\"Artículo 1\"><p>Obligación.</p></bloque></data></response>",
				bloque: bloqueLeido{titulo: "Artículo 1", fechaVersion: "original", texto: "Obligación."},
			},
			{
				nombre: "espacio en blanco, comentarios e instrucciones de proceso fuera de la raíz",
				cuerpo: " \r\n\t<!-- antes --><?proceso antes?>\n<response><data>" +
					"<bloque titulo=\"T\"><p>Texto.</p></bloque></data></response>\n<!-- después -->\n<?proceso después?>\n",
				bloque: bloqueLeido{titulo: "T", fechaVersion: "original", texto: "Texto."},
			},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				bloque, err := leerBloque([]byte(caso.cuerpo))

				require.NoError(t, err)
				assert.Equal(t, caso.bloque, bloque)
			})
		}
	})

	t.Run("las versiones de los bloques grabados, en el orden de la respuesta", func(t *testing.T) {
		t.Parallel()

		// Las fechas y la norma de cada versión, escritas a mano desde los
		// atributos de las grabaciones de H4 (research.md M1).
		casos := []struct {
			articulo  articuloDelDiff
			versiones []versionLeida
		}{
			{
				articulo: articuloDelDiff{norma: "BOE-A-1985-5392", bloque: "a22"},
				versiones: []versionLeida{
					{fechaVersion: "19850403", fechaVigencia: "19850423", normaModificadora: "BOE-A-1985-5392"},
					{fechaVersion: "19960608", fechaVigencia: "19960609", normaModificadora: "BOE-A-1996-13000"},
					{fechaVersion: "19970415", fechaVigencia: "19970416", normaModificadora: "BOE-A-1997-7879"},
					{fechaVersion: "19990422", fechaVigencia: "19990512", normaModificadora: "BOE-A-1999-8932"},
					{fechaVersion: "20031217", fechaVigencia: "20040101", normaModificadora: "BOE-A-2003-23103"},
					{fechaVersion: "20070529", fechaVigencia: "20070701", normaModificadora: "BOE-A-2007-10701"},
					{fechaVersion: "20071031", fechaVigencia: "20080430", normaModificadora: "BOE-A-2007-18874"},
					{fechaVersion: "20080626", fechaVigencia: "20080627", normaModificadora: "BOE-A-2008-10792"},
					{fechaVersion: "20151031", fechaVigencia: "20151031", normaModificadora: "BOE-A-2015-11723"},
				},
			},
			{
				articulo: articuloDelDiff{norma: "BOE-A-2017-12902", bloque: "a1-30"},
				versiones: []versionLeida{
					{fechaVersion: "20171109", fechaVigencia: "20180309", normaModificadora: "BOE-A-2017-12902"},
					{fechaVersion: "20200205", fechaVigencia: "20200206", normaModificadora: "BOE-A-2020-1651"},
				},
			},
			{
				articulo: articuloDelDiff{norma: "BOE-A-2015-10565", bloque: "a21"},
				versiones: []versionLeida{
					{fechaVersion: "20151002", fechaVigencia: "20161002", normaModificadora: "BOE-A-2015-10565"},
				},
			},
		}

		for _, caso := range casos {
			t.Run(caso.articulo.norma+" "+caso.articulo.bloque, func(t *testing.T) {
				t.Parallel()

				cuerpo := cuerpoGrabadoDelBloque(t, caso.articulo)

				bloque, err := leerBloque(cuerpo)
				require.NoError(t, err)

				assert.Equal(t, caso.versiones, versionesSinTexto(bloque.versiones))
				compruebaVigenteComoHoy(t, caso.articulo, bloque)
				compruebaTextoDeCadaVersion(t, cuerpo, bloque.versiones)
			})
		}
	})

	t.Run("una respuesta sin versiones deja la lista vacía", func(t *testing.T) {
		t.Parallel()

		bloque, err := leerBloque([]byte(respuestaConBloque(`<bloque id="pr" tipo="preambulo" titulo="Preámbulo">` +
			`<p class="parrafo">Primer párrafo.</p></bloque>`)))

		require.NoError(t, err)
		assert.Empty(t, bloque.versiones)
		assert.Equal(t, "Primer párrafo.", bloque.texto)
	})

	t.Run("lo que no se puede leer es fuente no disponible", func(t *testing.T) {
		t.Parallel()

		const (
			noUTF8         = "el cuerpo no es UTF-8 válido"
			ilegible       = "el cuerpo no es XML legible"
			sinRaiz        = "el XML no tiene elemento raíz"
			variasRaices   = "el XML tiene más de un elemento raíz"
			fueraDeRaiz    = "el XML tiene texto fuera del elemento raíz"
			conDeclaracion = "el XML trae una declaración <!…> que la lectura no interpreta"
			sinBloque      = "el XML no tiene ningún elemento bloque por debajo de la raíz"
		)

		casos := []struct {
			nombre  string
			cuerpo  string
			mensaje string
		}{
			{nombre: "HTML de mantenimiento, el sintético bloque-ilegible", cuerpo: "<html><body>Mantenimiento<br></body></html>", mensaje: ilegible},
			{nombre: "XML cortado", cuerpo: cabeceraDeLaRespuesta + `<bloque titulo="T"><version>`, mensaje: ilegible},
			{nombre: "cierre que no casa", cuerpo: `<response><data></response>`, mensaje: ilegible},
			{nombre: "entidad no definida", cuerpo: respuestaConBloque(`<bloque titulo="T"><p>a&nbsp;b</p></bloque>`), mensaje: ilegible},
			{nombre: "carácter no permitido", cuerpo: respuestaConBloque(`<bloque titulo="T"><p>a&#1;b</p></bloque>`), mensaje: ilegible},
			{nombre: "atributo sin comillas", cuerpo: respuestaConBloque(`<bloque titulo=T><p>Texto.</p></bloque>`), mensaje: ilegible},
			{nombre: "cuerpo vacío", cuerpo: "", mensaje: sinRaiz},
			{nombre: "solo la declaración, espacio en blanco y un comentario", cuerpo: "<?xml version=\"1.0\"?>\n<!-- nada -->\n", mensaje: sinRaiz},
			{
				nombre:  "dos raíces, aunque la primera traiga el bloque",
				cuerpo:  "<response><data><bloque titulo=\"T\"><p>Texto.</p></bloque></data></response>\n<response/>",
				mensaje: variasRaices,
			},
			{nombre: "JSON, como el de los demás verbos", cuerpo: `{"data": []}`, mensaje: fueraDeRaiz},
			{nombre: "texto tras la raíz", cuerpo: respuestaConBloque(`<bloque titulo="T"><p>Texto.</p></bloque>`) + "fin", mensaje: fueraDeRaiz},
			{nombre: "CDATA de espacio en blanco tras la raíz", cuerpo: "<response/><![CDATA[ ]]>", mensaje: fueraDeRaiz},
			{nombre: "referencia a un espacio tras la raíz", cuerpo: "<response/>&#32;", mensaje: fueraDeRaiz},
			{
				nombre:  "declaración de tipo de documento",
				cuerpo:  "<!DOCTYPE response>\n<response><data><bloque titulo=\"T\"><p>Texto.</p></bloque></data></response>",
				mensaje: conDeclaracion,
			},
			{
				nombre:  "declaración dentro de un elemento",
				cuerpo:  respuestaConBloque(`<bloque titulo="T"><!ENTITY x "y"><p>Texto.</p></bloque>`),
				mensaje: conDeclaracion,
			},
			{
				nombre:  "sin bloque, el sintético bloque-sin-elemento",
				cuerpo:  `<?xml version="1.0" encoding="utf-8"?><response><status><code>200</code><text>ok</text></status><data></data></response>`,
				mensaje: sinBloque,
			},
			{
				nombre:  "raíz bloque sin otro bloque debajo",
				cuerpo:  `<bloque titulo="Raíz"><version fecha_publicacion="20151002"><p>Texto.</p></version></bloque>`,
				mensaje: sinBloque,
			},
			{
				nombre:  "bloque con espacio de nombres",
				cuerpo:  `<response xmlns="urn:boe"><data><bloque titulo="T"><p>Texto.</p></bloque></data></response>`,
				mensaje: sinBloque,
			},
			{nombre: "un octeto que no es UTF-8 en el texto", cuerpo: respuestaConBloque("<bloque titulo=\"T\"><p>Obligaci\xf3n.</p></bloque>"), mensaje: noUTF8},
			{
				nombre:  "ISO-8859-1 declarada y cuerpo en ISO-8859-1: no se transcodifica",
				cuerpo:  "<?xml version=\"1.0\" encoding=\"ISO-8859-1\"?><response><data><bloque titulo=\"Art\xedculo 1\"/></data></response>",
				mensaje: noUTF8,
			},
		}

		for _, caso := range casos {
			t.Run(caso.nombre, func(t *testing.T) {
				t.Parallel()

				bloque, err := leerBloque([]byte(caso.cuerpo))

				assert.Zero(t, bloque)
				exigirIlegible(t, err, caso.mensaje)
			})
		}
	})

	t.Run("una etiqueta de apertura que no se puede volver a leer es fuente no disponible", func(t *testing.T) {
		t.Parallel()

		// La lectura solo vuelve a leer etiquetas que el decodificador ya aceptó;
		// estas no lo son, y el fallo tiene que declararse igual.
		for _, etiqueta := range []string{"", "texto", "</bloque>", `<bloque titulo=>`} {
			t.Run(etiqueta, func(t *testing.T) {
				t.Parallel()

				atributos, err := atributosDeLaEtiqueta([]byte(etiqueta))

				assert.Nil(t, atributos)
				exigirIlegible(t, err, "el cuerpo no es XML legible")
			})
		}
	})
}

// respuestaConBloque es la respuesta de la API con el XML de la tabla dentro de
// data.
func respuestaConBloque(bloque string) string {
	return cabeceraDeLaRespuesta + bloque + pieDeLaRespuesta
}

// cuerpoGrabadoDelBloque es el cuerpo XML de la grabación del bloque del
// artículo, pedido como lo pide la fuente al cliente de reproducción.
func cuerpoGrabadoDelBloque(t *testing.T, articulo articuloDelDiff) []byte {
	t.Helper()

	cliente := clienteDeReproduccion(t, carpetaDeLasGrabaciones, time.Now)

	respuesta, err := cliente.Pedir(t.Context(), schema.Contexto{}, httpx.Peticion{
		Metodo: "GET",
		URL:    direccionDelBloque(articulo.norma, articulo.bloque),
		Acepta: aceptaDelBloque,
	})
	require.NoError(t, err)

	return respuesta.Cuerpo
}

// versionesSinTexto son las versiones con sus fechas y su norma y sin el texto,
// que se comprueba aparte.
func versionesSinTexto(versiones []versionLeida) []versionLeida {
	sinTexto := make([]versionLeida, 0, len(versiones))

	for _, version := range versiones {
		version.texto = ""
		sinTexto = append(sinTexto, version)
	}

	return sinTexto
}

// compruebaVigenteComoHoy exige que la última versión sea lo que la lectura da
// hoy del bloque, campo a campo, y que su texto sea el de la referencia del diff
// de aceptación de H4, que derivó refs/boe.py (FR-012).
func compruebaVigenteComoHoy(t *testing.T, articulo articuloDelDiff, bloque bloqueLeido) {
	t.Helper()

	require.NotEmpty(t, bloque.versiones)

	vigente := versionLeida{
		fechaVersion:      bloque.fechaVersion,
		fechaVigencia:     bloque.fechaVigencia,
		normaModificadora: bloque.normaModificadora,
		texto:             bloque.texto,
	}
	assert.Equal(t, vigente, bloque.versiones[len(bloque.versiones)-1])
	assert.Equal(t, textoDeLaReferencia(t, articulo), bloque.texto)
}

// compruebaTextoDeCadaVersion exige que cada versión sea la que la lectura da
// como vigente del mismo cuerpo cuando se le quitan las demás versiones, que es
// como se midieron en research.md M1: su texto normalizado igual que hoy el de la
// última.
func compruebaTextoDeCadaVersion(t *testing.T, cuerpo []byte, versiones []versionLeida) {
	t.Helper()

	for posicion, version := range versiones {
		sola, err := leerBloque(cuerpoConUnaVersion(t, cuerpo, len(versiones), posicion))
		require.NoError(t, err)

		assert.NotEmpty(t, version.texto, "versión %d", posicion)
		assert.Equal(t, []versionLeida{version}, sola.versiones, "versión %d", posicion)
		assert.Equal(t, version.texto, sola.texto, "versión %d", posicion)
	}
}

// cuerpoConUnaVersion es el cuerpo de una grabación sin las versiones que no son
// la de esa posición: cada elemento version de las grabaciones lleva atributos y
// ninguno anida otro, así que cada «<version » abre uno, que termina en el primer
// «</version>» que le sigue. Lo que va entre versiones, y tras la última, se
// queda.
func cuerpoConUnaVersion(t *testing.T, cuerpo []byte, versiones, posicion int) []byte {
	t.Helper()

	partes := strings.Split(string(cuerpo), "<version ")
	require.Len(t, partes, versiones+1)

	var conUna strings.Builder

	conUna.WriteString(partes[0])

	for actual, parte := range partes[1:] {
		if actual == posicion {
			conUna.WriteString("<version " + parte)

			continue
		}

		_, cola, cerrada := strings.Cut(parte, "</version>")
		require.True(t, cerrada)
		conUna.WriteString(cola)
	}

	return []byte(conUna.String())
}
