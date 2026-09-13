package boe

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				},
			},
			{
				nombre: "la última versión sin atributos da la fecha original y lo demás vacío",
				cuerpo: respuestaConBloque(`<bloque id="a1"><version><p>Texto.</p></version></bloque>`),
				bloque: bloqueLeido{fechaVersion: "original", texto: "Texto."},
			},
			{
				nombre: "una fecha de publicación vacía presente es la fecha, no la original",
				cuerpo: respuestaConBloque(`<bloque titulo="Artículo 1">
<version id_norma="BOE-A-2015-10565" fecha_publicacion="" fecha_vigencia=""><p>Texto.</p></version>
</bloque>`),
				bloque: bloqueLeido{titulo: "Artículo 1", normaModificadora: "BOE-A-2015-10565", texto: "Texto."},
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
				},
			},
			{
				nombre: "la cola de una versión anterior no entra en el texto de la vigente",
				cuerpo: respuestaConBloque(`<bloque titulo="T">
<version fecha_publicacion="20151002"><p>Antigua.</p></version>
Entre versiones.
<version fecha_publicacion="20181206"><p>Vigente.</p></version>
</bloque>
Cola del bloque.`),
				bloque: bloqueLeido{titulo: "T", fechaVersion: "20181206", texto: "Vigente."},
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
				bloque: bloqueLeido{titulo: "T", fechaVersion: "20151002", texto: "Uno.\nDos.\nTres."},
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
				},
			},
			{
				nombre: "un comentario no aporta texto ni corta el que lo rodea",
				cuerpo: respuestaConBloque(`<bloque titulo="T"><version fecha_publicacion="20151002">` +
					`<p>Obliga<!-- corte -->ción de resolver.</p></version></bloque>`),
				bloque: bloqueLeido{titulo: "T", fechaVersion: "20151002", texto: "Obligación de resolver."},
			},
			{
				nombre: "una instrucción de proceso no aporta texto ni corta el que la rodea",
				cuerpo: respuestaConBloque(`<bloque titulo="T"><version fecha_publicacion="20151002">` +
					`<p>Obliga<?proceso algo?>ción de resolver.</p></version></bloque>`),
				bloque: bloqueLeido{titulo: "T", fechaVersion: "20151002", texto: "Obligación de resolver."},
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
				},
			},
			{
				nombre: "las referencias de carácter a tabulador, salto y retorno de un atributo se conservan",
				cuerpo: respuestaConBloque(`<bloque titulo="a&#10;b&#9;c&#13;d" tipo='x&#xA;"y"'>` +
					`<version fecha_publicacion="20151002"><p>Texto.</p></version></bloque>`),
				bloque: bloqueLeido{titulo: "a\nb\tc\rd", tipo: "x\n\"y\"", fechaVersion: "20151002", texto: "Texto."},
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
				bloque: bloqueLeido{titulo: "T", fechaVersion: "20151002", texto: "Uno   y\tuno.\nDos."},
			},
			{
				nombre: "la codificación ISO-8859-1 declarada no cuenta: el cuerpo se lee como UTF-8",
				cuerpo: "<?xml version=\"1.0\" encoding=\"ISO-8859-1\"?>\n<response><data>" +
					"<bloque titulo=\"Artículo 1\"><version fecha_publicacion=\"20151002\"><p>Obligación.</p></version></bloque>" +
					"</data></response>",
				bloque: bloqueLeido{titulo: "Artículo 1", fechaVersion: "20151002", texto: "Obligación."},
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
