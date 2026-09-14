package skills_test

import (
	"encoding/json/jsontext"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/skills"
)

// Las dos marcas de la región de la tabla de comandos de SKILL.md, escritas aquí
// tal cual las fija data-model §2.
const (
	inicioDeLaTabla = "<!-- inicio de la tabla de comandos: generada desde --describe con make skills-sync, no editar -->"
	finDeLaTabla    = "<!-- fin de la tabla de comandos -->"
)

// plantillaDelDocumento es un documento de --describe con la forma del que emite
// cli.Describir: $defs con DatosError y las definiciones del caso; la entrada con
// los argumentos del verbo seguidos de las banderas globales; y la salida con el
// sobre y data condicionado a ok. Cada verbo de formato es una parte de
// describeDePrueba.
const plantillaDelDocumento = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$defs": {%s},
  "properties": {
    "entrada": {
      "properties": {%s},
      "additionalProperties": false,
      "type": "object",
      "required": %s
    },
    "salida": {
      "if": {"properties": {"ok": {"const": true}}},
      "then": {"properties": {"data": %s}},
      "else": {"properties": {"data": %s}},
      "properties": {
        "ok": {"type": "boolean"},
        "fuente": {"type": "string", "minLength": 1},
        "url": {"type": "string", "minLength": 1, "format": "uri"},
        "fecha_consulta": {"type": "string", "format": "date-time"},
        "hash": {"type": "string", "pattern": "^sha256:[0-9a-f]{64}$"},
        "data": true
      },
      "additionalProperties": false,
      "type": "object",
      "required": %s
    }
  },
  "title": %q,
  "description": %q
}`

// Las partes de un documento de prueba que son las de todo verbo de kitlegal.
const (
	// banderasGlobalesDePrueba son las ocho banderas globales, en el orden en que
	// --describe las declara detrás de los argumentos de cualquier verbo.
	banderasGlobalesDePrueba = `"json": {"type": "boolean"},
        "timeout": {"type": "string"},
        "offline": {"type": "boolean"},
        "dry-run": {"type": "boolean"},
        "describe": {"type": "boolean"},
        "no-graph": {"type": "boolean"},
        "asunto": {"type": "string"},
        "verbose": {"type": "boolean"}`

	// sobreDePrueba son las claves obligatorias del sobre.
	sobreDePrueba = `["ok", "fuente", "url", "fecha_consulta", "hash", "data"]`

	// falloDePrueba es el data de un sobre con ok falso.
	falloDePrueba = `{"$ref": "#/$defs/DatosError"}`

	// definicionDeLosDatosDeError es la forma común de error del kernel.
	definicionDeLosDatosDeError = `"DatosError": {
    "properties": {
      "clase": {"type": "string", "enum": ["argumentos", "no-encontrado", "inesperado"]},
      "mensaje": {"type": "string", "minLength": 1}
    },
    "additionalProperties": false,
    "type": "object",
    "required": ["clase", "mensaje"]
  }`
)

// describeDePrueba son las partes de un documento de --describe que cambian de
// un caso a otro, cada una como texto JSON salvo el título y la ayuda.
type describeDePrueba struct {
	// titulo y ayuda son title y description.
	titulo, ayuda string

	// argumentos son los miembros de entrada.properties que van delante de las
	// banderas, y banderas los que van detrás.
	argumentos, banderas string

	// obligatorios es entrada.required.
	obligatorios string

	// datos es then.properties.data, y fallo, else.properties.data.
	datos, fallo string

	// sobre es salida.required.
	sobre string

	// definiciones son los miembros de $defs que van detrás de DatosError.
	definiciones string
}

// documentoDeVerbo es el documento de un verbo sin argumentos y sin forma de
// data declarada, con las banderas, el sobre y el fallo de todo verbo.
func documentoDeVerbo(titulo, ayuda string) describeDePrueba {
	return describeDePrueba{
		titulo:       titulo,
		ayuda:        ayuda,
		banderas:     banderasGlobalesDePrueba,
		obligatorios: "[]",
		datos:        "true",
		fallo:        falloDePrueba,
		sobre:        sobreDePrueba,
	}
}

// contenido es el documento JSON.
func (d describeDePrueba) contenido() []byte {
	return fmt.Appendf(nil, plantillaDelDocumento,
		unirMiembros(definicionDeLosDatosDeError, d.definiciones),
		unirMiembros(d.argumentos, d.banderas),
		d.obligatorios, d.datos, d.fallo, d.sobre, d.titulo, d.ayuda)
}

// unirMiembros une los miembros de un objeto JSON que no están vacíos.
func unirMiembros(miembros ...string) string {
	return strings.Join(slices.DeleteFunc(miembros, func(miembro string) bool { return miembro == "" }), ", ")
}

// globalesDePrueba es el documento de un verbo sin argumentos, del que salen las
// banderas globales: el de cli.Describir con Argumentos nulo, sin applet ni verbo
// y, por eso, con el título vacío.
func globalesDePrueba() describeDePrueba {
	return documentoDeVerbo("", "")
}

// articuloDePrueba es el documento de boe articulo: dos argumentos obligatorios
// y data con la definición de un objeto.
func articuloDePrueba() describeDePrueba {
	documento := documentoDeVerbo("boe articulo",
		"Devuelve el texto vigente de un bloque de una norma, con los avisos de su vigencia.")
	documento.argumentos = `"norma": {"type": "string"}, "bloque": {"type": "string"}`
	documento.obligatorios = `["norma", "bloque"]`
	documento.datos = `{"$ref": "#/$defs/boe.Articulo"}`
	documento.definiciones = `"boe.Articulo": {
    "properties": {
      "norma": {"type": "string"},
      "bloque": {"type": "string"},
      "titulo": {"type": "string"},
      "tipo": {"type": "string"},
      "fecha_version": {"type": "string"},
      "fecha_vigencia": {"type": "string"},
      "norma_modificadora": {"type": "string"},
      "texto": {"type": "string"},
      "hash_texto": {"type": "string", "pattern": "^sha256:[0-9a-f]{64}$"},
      "avisos": {"items": {"$ref": "#/$defs/boe.Aviso"}, "type": "array"},
      "url": {"type": "string"},
      "url_eli": {"type": "string"}
    },
    "additionalProperties": false,
    "type": "object",
    "required": ["norma", "bloque", "titulo", "tipo", "fecha_version", "fecha_vigencia",
      "norma_modificadora", "texto", "hash_texto", "avisos", "url", "url_eli"]
  },
  "boe.Aviso": {
    "properties": {"codigo": {"type": "string"}, "texto": {"type": "string"}},
    "additionalProperties": false,
    "type": "object",
    "required": ["codigo", "texto"]
  }`

	return documento
}

// buscarDePrueba es el documento de boe buscar: un argumento obligatorio de
// varios valores y data con una lista de objetos. Las claves obligatorias de la
// definición van en otro orden que sus propiedades, que son las que dan el orden
// de las claves.
func buscarDePrueba() describeDePrueba {
	documento := documentoDeVerbo("boe buscar",
		"Busca normas consolidadas por las palabras de su título o con una consulta de la fuente.")
	documento.argumentos = `"texto": {"items": {"type": "string"}, "type": "array"}`
	documento.obligatorios = `["texto"]`
	documento.datos = `{"items": {"$ref": "#/$defs/boe.ResultadoDeBusqueda"}, "type": "array"}`
	documento.definiciones = `"boe.ResultadoDeBusqueda": {
    "properties": {
      "identificador": {"type": "string"},
      "titulo": {"type": "string"},
      "rango": {"type": "string"},
      "vigencia_agotada": {"type": "string"},
      "estado_consolidacion": {"type": "string"},
      "url": {"type": "string"}
    },
    "additionalProperties": false,
    "type": "object",
    "required": ["estado_consolidacion", "identificador", "rango", "titulo", "url", "vigencia_agotada"]
  }`

	return documento
}

// consultarDePrueba es el documento de un verbo con un argumento obligatorio,
// otro obligatorio de varios valores, uno opcional y otro opcional de varios
// valores; con una barra en la ayuda y en una clave de data.
func consultarDePrueba() describeDePrueba {
	documento := documentoDeVerbo("ejemplo consultar", "Consulta los bloques de una norma | o de varias.")
	documento.argumentos = `"norma": {"type": "string"},
        "bloques": {"items": {"type": "string"}, "type": "array"},
        "desde": {"type": "string"},
        "materias": {"items": {"type": "string"}, "type": "array"}`
	documento.obligatorios = `["norma", "bloques"]`
	documento.datos = `{"$ref": "#/$defs/ejemplo.Consulta"}`
	documento.definiciones = `"ejemplo.Consulta": {
    "properties": {"norma": {"type": "string"}, "con|barra": {"type": "string"}},
    "type": "object"
  }`

	return documento
}

// nadaDePrueba es el documento de un verbo cuyo data es una lista de objetos
// cuya definición no declara ninguna propiedad.
func nadaDePrueba() describeDePrueba {
	documento := documentoDeVerbo("ejemplo nada", "Devuelve una lista de objetos sin claves.")
	documento.datos = `{"items": {"$ref": "#/$defs/ejemplo.Nada"}, "type": "array"}`
	documento.definiciones = `"ejemplo.Nada": {"additionalProperties": false, "type": "object"}`

	return documento
}

// banderasEsperadas son las banderas globales de globalesDePrueba, en su orden.
func banderasEsperadas() []skills.Bandera {
	return []skills.Bandera{
		{Nombre: "json"},
		{Nombre: "timeout", ConValor: true},
		{Nombre: "offline"},
		{Nombre: "dry-run"},
		{Nombre: "describe"},
		{Nombre: "no-graph"},
		{Nombre: "asunto", ConValor: true},
		{Nombre: "verbose"},
	}
}

// sobreEsperado es el sobre de todo documento de prueba.
func sobreEsperado() skills.Sobre {
	return skills.Sobre{
		Claves:        []string{"ok", "fuente", "url", "fecha_consulta", "hash", "data"},
		ClavesDeFallo: []string{"clase", "mensaje"},
	}
}

// TestDescripcionDeVerbo fija LeerDescripcionDeVerbo (data-model §2.1; research.md
// D6): el applet y el verbo del título, la ayuda, los argumentos en su orden sin
// las banderas globales —las del documento de un verbo sin argumentos, por su
// nombre— con su obligatoriedad y si admiten varios valores, las banderas con si
// llevan valor, las claves de data de un objeto y de una lista en el orden de las
// propiedades de su definición, la forma sin declarar cuando data no tiene $ref, y
// el sobre; y cada documento que no permite describir el verbo, con el defecto
// que lo impide.
func TestDescripcionDeVerbo(t *testing.T) {
	t.Parallel()

	sinForma := documentoDeVerbo("ejemplo vacio", "No declara la forma de sus datos.")

	listaDeCadenas := documentoDeVerbo("ejemplo cadenas", "Devuelve una lista de cadenas.")
	listaDeCadenas.datos = `{"items": {"type": "string"}, "type": "array"}`

	opcionales := documentoDeVerbo("ejemplo opcionales", "Tiene argumentos opcionales.")
	opcionales.argumentos = `"norma": {"type": "string"},
        "desde": {"type": "string"},
        "materias": {"items": {"type": "string"}, "type": "array"}`
	opcionales.obligatorios = `["norma"]`

	pocasGlobales := globalesDePrueba()
	pocasGlobales.banderas = `"asunto": {"type": "string"}, "json": {"type": "boolean"}`

	casos := []struct {
		nombre    string
		documento describeDePrueba
		globales  describeDePrueba
		esperada  skills.DescripcionDeVerbo
	}{
		{
			nombre:    "objeto",
			documento: articuloDePrueba(),
			globales:  globalesDePrueba(),
			esperada: skills.DescripcionDeVerbo{
				Applet: "boe",
				Verbo:  "articulo",
				Hace:   "Devuelve el texto vigente de un bloque de una norma, con los avisos de su vigencia.",
				Argumentos: []skills.Argumento{
					{Nombre: "norma", Obligatorio: true},
					{Nombre: "bloque", Obligatorio: true},
				},
				Banderas: banderasEsperadas(),
				Devuelve: devuelveDelArticulo(),
				Sobre:    sobreEsperado(),
			},
		},
		{
			nombre:    "lista",
			documento: buscarDePrueba(),
			globales:  globalesDePrueba(),
			esperada: skills.DescripcionDeVerbo{
				Applet:     "boe",
				Verbo:      "buscar",
				Hace:       "Busca normas consolidadas por las palabras de su título o con una consulta de la fuente.",
				Argumentos: []skills.Argumento{{Nombre: "texto", Obligatorio: true, Varios: true}},
				Banderas:   banderasEsperadas(),
				Devuelve: skills.Devuelve{
					Forma:  skills.FormaListaDeObjetos,
					Claves: []string{"identificador", "titulo", "rango", "vigencia_agotada", "estado_consolidacion", "url"},
				},
				Sobre: sobreEsperado(),
			},
		},
		{
			nombre:    "sin-ref",
			documento: sinForma,
			globales:  globalesDePrueba(),
			esperada: skills.DescripcionDeVerbo{
				Applet:   "ejemplo",
				Verbo:    "vacio",
				Hace:     "No declara la forma de sus datos.",
				Banderas: banderasEsperadas(),
				Devuelve: skills.Devuelve{Forma: skills.FormaSinDeclarar},
				Sobre:    sobreEsperado(),
			},
		},
		{
			nombre:    "sin-ref-en-los-elementos-de-la-lista",
			documento: listaDeCadenas,
			globales:  globalesDePrueba(),
			esperada: skills.DescripcionDeVerbo{
				Applet:   "ejemplo",
				Verbo:    "cadenas",
				Hace:     "Devuelve una lista de cadenas.",
				Banderas: banderasEsperadas(),
				Devuelve: skills.Devuelve{Forma: skills.FormaSinDeclarar},
				Sobre:    sobreEsperado(),
			},
		},
		{
			nombre:    "objeto-sin-claves",
			documento: nadaDePrueba(),
			globales:  globalesDePrueba(),
			esperada: skills.DescripcionDeVerbo{
				Applet:   "ejemplo",
				Verbo:    "nada",
				Hace:     "Devuelve una lista de objetos sin claves.",
				Banderas: banderasEsperadas(),
				Devuelve: skills.Devuelve{Forma: skills.FormaListaDeObjetos},
				Sobre:    sobreEsperado(),
			},
		},
		{
			nombre:    "argumentos-opcionales-y-de-varios-valores",
			documento: opcionales,
			globales:  globalesDePrueba(),
			esperada: skills.DescripcionDeVerbo{
				Applet: "ejemplo",
				Verbo:  "opcionales",
				Hace:   "Tiene argumentos opcionales.",
				Argumentos: []skills.Argumento{
					{Nombre: "norma", Obligatorio: true},
					{Nombre: "desde"},
					{Nombre: "materias", Varios: true},
				},
				Banderas: banderasEsperadas(),
				Devuelve: skills.Devuelve{Forma: skills.FormaSinDeclarar},
				Sobre:    sobreEsperado(),
			},
		},
		{
			// Solo se quitan de los argumentos las banderas que declara el documento
			// de un verbo sin argumentos, por su nombre y no por su sitio, y las
			// banderas van en el orden de ese documento.
			nombre:    "sin-las-banderas-globales-por-su-nombre",
			documento: articuloDePrueba(),
			globales:  pocasGlobales,
			esperada: skills.DescripcionDeVerbo{
				Applet: "boe",
				Verbo:  "articulo",
				Hace:   "Devuelve el texto vigente de un bloque de una norma, con los avisos de su vigencia.",
				Argumentos: []skills.Argumento{
					{Nombre: "norma", Obligatorio: true},
					{Nombre: "bloque", Obligatorio: true},
					{Nombre: "timeout"},
					{Nombre: "offline"},
					{Nombre: "dry-run"},
					{Nombre: "describe"},
					{Nombre: "no-graph"},
					{Nombre: "verbose"},
				},
				Banderas: []skills.Bandera{{Nombre: "asunto", ConValor: true}, {Nombre: "json"}},
				Devuelve: devuelveDelArticulo(),
				Sobre:    sobreEsperado(),
			},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			descripcion, err := skills.LeerDescripcionDeVerbo(caso.documento.contenido(), caso.globales.contenido())
			require.NoError(t, err)
			assert.Equal(t, caso.esperada, descripcion)
		})
	}

	t.Run("sin-rama-then", func(t *testing.T) {
		t.Parallel()

		rama := `"then": {"properties": {"data": true}},`
		documento := strings.Replace(string(sinForma.contenido()), rama, "", 1)
		require.NotContains(t, documento, `"then"`, "el documento de prueba no tiene la rama then")

		descripcion, err := skills.LeerDescripcionDeVerbo([]byte(documento), globalesDePrueba().contenido())
		require.NoError(t, err)
		assert.Equal(t, skills.Devuelve{Forma: skills.FormaSinDeclarar}, descripcion.Devuelve)
	})

	t.Run("defectos", func(t *testing.T) {
		t.Parallel()

		probarDefectosDeLaDescripcion(t)
	})
}

// devuelveDelArticulo es lo que devuelve articuloDePrueba: un objeto con las
// claves de su definición, en el orden de sus propiedades.
func devuelveDelArticulo() skills.Devuelve {
	return skills.Devuelve{
		Forma: skills.FormaObjeto,
		Claves: []string{
			"norma", "bloque", "titulo", "tipo", "fecha_version", "fecha_vigencia",
			"norma_modificadora", "texto", "hash_texto", "avisos", "url", "url_eli",
		},
	}
}

// probarDefectosDeLaDescripcion comprueba que cada documento que no permite
// describir el verbo da su defecto, nombrando el documento y la parte que falla.
func probarDefectosDeLaDescripcion(t *testing.T) {
	t.Helper()

	conDocumento := func(cambiar func(*describeDePrueba)) []byte {
		documento := articuloDePrueba()
		cambiar(&documento)

		return documento.contenido()
	}

	conGlobales := func(cambiar func(*describeDePrueba)) []byte {
		documento := globalesDePrueba()
		cambiar(&documento)

		return documento.contenido()
	}

	articulo, globales := articuloDePrueba().contenido(), globalesDePrueba().contenido()

	casos := []struct {
		nombre              string
		documento, globales []byte
		error               string
	}{
		{
			nombre:    "titulo-sin-verbo",
			documento: conDocumento(func(d *describeDePrueba) { d.titulo = "boe" }),
			globales:  globales,
			error:     "documento de --describe «boe»: el título no tiene la forma «<applet> <verbo>»",
		},
		{
			nombre:    "titulo-vacio",
			documento: conDocumento(func(d *describeDePrueba) { d.titulo = "" }),
			globales:  globales,
			error:     "documento de --describe «»: el título no tiene la forma «<applet> <verbo>»",
		},
		{
			nombre:    "titulo-con-tres-palabras",
			documento: conDocumento(func(d *describeDePrueba) { d.titulo = "boe articulo a21" }),
			globales:  globales,
			error:     "documento de --describe «boe articulo a21»: el título no tiene la forma «<applet> <verbo>»",
		},
		{
			nombre:    "sin-entrada",
			documento: []byte(`{"title": "boe articulo", "properties": {"salida": {"required": ["ok"]}}}`),
			globales:  globales,
			error:     "documento de --describe «boe articulo»: falta properties.entrada",
		},
		{
			nombre:    "sin-salida",
			documento: []byte(`{"title": "boe articulo", "properties": {"entrada": {"properties": {}}}}`),
			globales:  globales,
			error:     "documento de --describe «boe articulo»: falta properties.salida",
		},
		{
			nombre:    "sobre-sin-claves",
			documento: conDocumento(func(d *describeDePrueba) { d.sobre = "[]" }),
			globales:  globales,
			error: "documento de --describe «boe articulo»: properties.salida.required no declara ninguna " +
				"clave del sobre",
		},
		{
			nombre:    "objeto-con-referencia-que-no-resuelve",
			documento: conDocumento(func(d *describeDePrueba) { d.datos = `{"$ref": "#/$defs/boe.Otro"}` }),
			globales:  globales,
			error: "documento de --describe «boe articulo»: properties.salida.then.properties.data apunta a " +
				"#/$defs/boe.Otro, que no es ninguna definición de $defs",
		},
		{
			nombre: "lista-con-referencia-externa",
			documento: conDocumento(func(d *describeDePrueba) {
				d.datos = `{"items": {"$ref": "otro.json#/$defs/boe.Articulo"}, "type": "array"}`
			}),
			globales: globales,
			error: "documento de --describe «boe articulo»: properties.salida.then.properties.data.items apunta a " +
				"otro.json#/$defs/boe.Articulo, que no es ninguna definición de $defs",
		},
		{
			nombre:    "fallo-sin-referencia",
			documento: conDocumento(func(d *describeDePrueba) { d.fallo = "true" }),
			globales:  globales,
			error: "documento de --describe «boe articulo»: properties.salida.else.properties.data no apunta a " +
				"ninguna definición de $defs",
		},
		{
			nombre:    "fallo-con-referencia-que-no-resuelve",
			documento: conDocumento(func(d *describeDePrueba) { d.fallo = `{"$ref": "#/$defs/OtroError"}` }),
			globales:  globales,
			error: "documento de --describe «boe articulo»: properties.salida.else.properties.data apunta a " +
				"#/$defs/OtroError, que no es ninguna definición de $defs",
		},
		{
			nombre: "fallo-sin-claves",
			documento: conDocumento(func(d *describeDePrueba) {
				d.fallo = `{"$ref": "#/$defs/boe.Vacio"}`
				d.definiciones += `, "boe.Vacio": {"type": "object"}`
			}),
			globales: globales,
			error: "documento de --describe «boe articulo»: la definición boe.Vacio de " +
				"properties.salida.else.properties.data no declara ninguna clave obligatoria",
		},
		{
			nombre:    "banderas-sin-entrada",
			documento: articulo,
			globales:  []byte(`{"properties": {"salida": {}}}`),
			error:     "documento de --describe de las banderas globales: falta properties.entrada",
		},
		{
			nombre:    "sin-ninguna-bandera",
			documento: articulo,
			globales:  conGlobales(func(d *describeDePrueba) { d.banderas = "" }),
			error: "documento de --describe de las banderas globales: properties.entrada.properties no declara " +
				"ninguna bandera",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			descripcion, err := skills.LeerDescripcionDeVerbo(caso.documento, caso.globales)
			require.EqualError(t, err, caso.error)
			assert.Zero(t, descripcion)
		})
	}

	lecturas := []struct {
		nombre              string
		documento, globales []byte
		prefijo             string
		causa               error
	}{
		{
			nombre:    "documento-que-no-es-json",
			documento: []byte(`{"title": "boe articulo"`),
			globales:  globales,
			prefijo:   "documento de --describe: no se puede leer: ",
			causa:     nil,
		},
		{
			nombre:    "documento-con-una-clave-repetida",
			documento: []byte(`{"title": "boe articulo", "title": "boe buscar"}`),
			globales:  globales,
			prefijo:   "documento de --describe: no se puede leer: ",
			causa:     jsontext.ErrDuplicateName,
		},
		{
			nombre:    "propiedades-que-no-son-un-objeto",
			documento: []byte(`{"title": "boe articulo", "properties": ["entrada", "salida"]}`),
			globales:  globales,
			prefijo:   "documento de --describe: no se puede leer: ",
			causa:     nil,
		},
		{
			nombre:    "propiedad-repetida",
			documento: []byte(`{"title": "boe articulo", "properties": {"entrada": {}, "entrada": {}}}`),
			globales:  globales,
			prefijo:   "documento de --describe: no se puede leer: ",
			causa:     jsontext.ErrDuplicateName,
		},
		{
			nombre:    "propiedad-que-no-es-un-esquema",
			documento: []byte(`{"title": "boe articulo", "properties": {"entrada": 5}}`),
			globales:  globales,
			prefijo:   "documento de --describe: no se puede leer: ",
			causa:     nil,
		},
		{
			nombre:    "propiedades-cortadas",
			documento: []byte(`{"title": "boe articulo", "properties": `),
			globales:  globales,
			prefijo:   "documento de --describe: no se puede leer: ",
			causa:     nil,
		},
		{
			nombre:    "banderas-que-no-son-json",
			documento: articulo,
			globales:  []byte(`[`),
			prefijo:   "documento de --describe de las banderas globales: no se puede leer: ",
			causa:     nil,
		},
	}

	for _, lectura := range lecturas {
		t.Run(lectura.nombre, func(t *testing.T) {
			t.Parallel()

			descripcion, err := skills.LeerDescripcionDeVerbo(lectura.documento, lectura.globales)
			require.Error(t, err)
			assert.True(t, strings.HasPrefix(err.Error(), lectura.prefijo), "error: %v", err)
			assert.Zero(t, descripcion)

			if lectura.causa != nil {
				require.ErrorIs(t, err, lectura.causa)
			}
		})
	}
}

// Los bytes de la tabla que se repiten en los casos de TestRenderizarTabla
// (contrato sincronizacion-y-comprobacion §3).
const (
	columnasDeLaTabla = "| Orden | Qué hace | Qué devuelve en `data` |\n|---|---|---|\n"

	filaDeBuscar = "| `scripts/boe buscar <texto>...` | Busca normas consolidadas por las palabras de su título o " +
		"con una consulta de la fuente. | lista de objetos con `identificador`, `titulo`, `rango`, " +
		"`vigencia_agotada`, `estado_consolidacion`, `url` |\n"

	filaDeArticulo = "| `scripts/boe articulo <norma> <bloque>` | Devuelve el texto vigente de un bloque de una " +
		"norma, con los avisos de su vigencia. | objeto con `norma`, `bloque`, `titulo`, `tipo`, " +
		"`fecha_version`, `fecha_vigencia`, `norma_modificadora`, `texto`, `hash_texto`, `avisos`, `url`, " +
		"`url_eli` |\n"

	lineaDeLasBanderasDePrueba = "Banderas comunes: `--json`, `--timeout <valor>`, `--offline`, `--dry-run`, " +
		"`--describe`, `--no-graph`, `--asunto <valor>`, `--verbose`.\n"

	pieDeLaTabla = "\n" +
		"Todas devuelven el sobre `ok`, `fuente`, `url`, `fecha_consulta`, `hash`, `data`; con `ok` falso, " +
		"`data` lleva `clase` y `mensaje`.\n" +
		"\n" +
		lineaDeLasBanderasDePrueba +
		"\n"
)

// describirDocumentos lee la descripción de cada documento, en su orden, con las
// banderas globales de globalesDePrueba.
func describirDocumentos(t *testing.T, documentos ...describeDePrueba) []skills.DescripcionDeVerbo {
	t.Helper()

	descripciones := make([]skills.DescripcionDeVerbo, 0, len(documentos))

	for _, documento := range documentos {
		descripcion, err := skills.LeerDescripcionDeVerbo(documento.contenido(), globalesDePrueba().contenido())
		require.NoError(t, err, documento.titulo)

		descripciones = append(descripciones, descripcion)
	}

	return descripciones
}

// TestRenderizarTabla fija los bytes exactos de RenderizarTabla sobre documentos
// de --describe de prueba (contrato sincronizacion-y-comprobacion §3; data-model
// §2.2; FR-032, FR-034): las filas de buscar y articulo del contrato; una sección
// por applet en el orden declarado, con sus filas en el orden de las
// descripciones y sin las de un applet no declarado; la sintaxis de un argumento
// obligatorio, de uno de varios valores y de los opcionales; la barra de la ayuda
// y de una clave escrita \|; data sin $ref y una lista de objetos sin claves; la
// misma salida en dos llamadas; y cada conjunto de descripciones que no permite
// escribir la tabla, con su defecto.
func TestRenderizarTabla(t *testing.T) {
	t.Parallel()

	vacio := documentoDeVerbo("ejemplo vacio", "No declara la forma de sus datos.")
	noDeclarado := documentoDeVerbo("otro listar", "No se presenta: su applet no está declarado.")

	unaClaveDeFallo := documentoDeVerbo("ejemplo vacio", "No declara la forma de sus datos.")
	unaClaveDeFallo.fallo = `{"$ref": "#/$defs/ejemplo.Error"}`
	unaClaveDeFallo.definiciones = `"ejemplo.Error": {"type": "object", "required": ["mensaje"]}`

	casos := []struct {
		nombre     string
		applets    []string
		documentos []describeDePrueba
		esperada   string
	}{
		{
			nombre:     "una-sola-clave-de-fallo",
			applets:    []string{"ejemplo"},
			documentos: []describeDePrueba{unaClaveDeFallo},
			esperada: "\n### `scripts/ejemplo`\n\n" + columnasDeLaTabla +
				"| `scripts/ejemplo vacio` | No declara la forma de sus datos. | sin forma declarada |\n" +
				"\n" +
				"Todas devuelven el sobre `ok`, `fuente`, `url`, `fecha_consulta`, `hash`, `data`; con `ok` " +
				"falso, `data` lleva `mensaje`.\n" +
				"\n" +
				lineaDeLasBanderasDePrueba +
				"\n",
		},
		{
			nombre:     "filas-del-contrato",
			applets:    []string{"boe"},
			documentos: []describeDePrueba{buscarDePrueba(), articuloDePrueba()},
			esperada:   "\n### `scripts/boe`\n\n" + columnasDeLaTabla + filaDeBuscar + filaDeArticulo + pieDeLaTabla,
		},
		{
			nombre:     "secciones-en-el-orden-declarado",
			applets:    []string{"ejemplo", "boe"},
			documentos: []describeDePrueba{buscarDePrueba(), vacio, noDeclarado, articuloDePrueba()},
			esperada: "\n### `scripts/ejemplo`\n\n" + columnasDeLaTabla +
				"| `scripts/ejemplo vacio` | No declara la forma de sus datos. | sin forma declarada |\n" +
				"\n### `scripts/boe`\n\n" + columnasDeLaTabla + filaDeBuscar + filaDeArticulo +
				pieDeLaTabla,
		},
		{
			nombre:     "sintaxis-y-escapes",
			applets:    []string{"ejemplo"},
			documentos: []describeDePrueba{consultarDePrueba(), vacio, nadaDePrueba()},
			esperada: "\n### `scripts/ejemplo`\n\n" + columnasDeLaTabla +
				"| `scripts/ejemplo consultar <norma> <bloques>... [--desde] [--materias]` | Consulta los " +
				"bloques de una norma \\| o de varias. | objeto con `norma`, `con\\|barra` |\n" +
				"| `scripts/ejemplo vacio` | No declara la forma de sus datos. | sin forma declarada |\n" +
				"| `scripts/ejemplo nada` | Devuelve una lista de objetos sin claves. | lista de objetos |\n" +
				pieDeLaTabla,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			descripciones := describirDocumentos(t, caso.documentos...)

			tabla, err := skills.RenderizarTabla(caso.applets, descripciones)
			require.NoError(t, err)
			assert.Equal(t, caso.esperada, string(tabla))

			otra, err := skills.RenderizarTabla(caso.applets, descripciones)
			require.NoError(t, err)
			assert.Equal(t, tabla, otra, "dos llamadas dan los mismos bytes")
		})
	}

	defectos := []struct {
		nombre  string
		applets []string
		cambiar func(descripciones []skills.DescripcionDeVerbo)
		error   string
	}{
		{
			nombre:  "sin-applets",
			applets: nil,
			cambiar: func([]skills.DescripcionDeVerbo) {},
			error:   "la tabla de comandos no declara ningún applet",
		},
		{
			nombre:  "applet-sin-verbos",
			applets: []string{"boe", "otro"},
			cambiar: func([]skills.DescripcionDeVerbo) {},
			error:   "el applet otro no tiene ningún verbo descrito",
		},
		{
			nombre:  "banderas-distintas",
			applets: []string{"boe"},
			cambiar: func(descripciones []skills.DescripcionDeVerbo) {
				descripciones[1].Banderas = descripciones[1].Banderas[1:]
			},
			error: "boe articulo no declara las mismas banderas globales que boe buscar",
		},
		{
			nombre:  "sobre-distinto",
			applets: []string{"boe"},
			cambiar: func(descripciones []skills.DescripcionDeVerbo) {
				descripciones[1].Sobre.ClavesDeFallo = []string{"clase"}
			},
			error: "boe articulo no declara el mismo sobre que boe buscar",
		},
	}

	for _, defecto := range defectos {
		t.Run(defecto.nombre, func(t *testing.T) {
			t.Parallel()

			descripciones := describirDocumentos(t, buscarDePrueba(), articuloDePrueba())
			defecto.cambiar(descripciones)

			tabla, err := skills.RenderizarTabla(defecto.applets, descripciones)
			require.EqualError(t, err, defecto.error)
			assert.Nil(t, tabla)
		})
	}
}

// TestSustituirRegion fija SustituirRegion (data-model §2; FR-032): sustituye
// solo lo que hay entre las dos marcas y deja el resto byte a byte igual, también
// con la región vacía, con las marcas en la primera y en la última línea y con la
// última sin salto final; y da un defecto, sin contenido, sin marcas, sin una de
// las dos, con una marca en más de una línea, con el fin antes del inicio o con
// una línea que no es exactamente la marca.
func TestSustituirRegion(t *testing.T) {
	t.Parallel()

	const (
		antes    = "---\nname: ejemplo\n---\n\n# Ejemplo\n\n## Comandos\n\n"
		despues  = "\n## Reglas\n\n1. Nunca inventar contenido legal.\n"
		nueva    = "\n### `scripts/boe`\n\nnueva\n\n"
		inicio   = inicioDeLaTabla + "\n"
		fin      = finDeLaTabla + "\n"
		sinFinal = finDeLaTabla
	)

	casos := []struct {
		nombre, contenido, region, esperado string
	}{
		{
			nombre:    "sustituye-solo-entre-marcas",
			contenido: antes + inicio + "\nvieja\n\n| a | b |\n" + fin + despues,
			region:    nueva,
			esperado:  antes + inicio + nueva + fin + despues,
		},
		{
			nombre:    "region-vacia",
			contenido: antes + inicio + fin + despues,
			region:    nueva,
			esperado:  antes + inicio + nueva + fin + despues,
		},
		{
			nombre:    "deja-la-region-vacia",
			contenido: antes + inicio + "vieja\n" + fin + despues,
			region:    "",
			esperado:  antes + inicio + fin + despues,
		},
		{
			nombre:    "marcas-en-la-primera-y-en-la-ultima-linea-sin-salto-final",
			contenido: inicio + "vieja\n" + sinFinal,
			region:    nueva,
			esperado:  inicio + nueva + sinFinal,
		},
		{
			nombre:    "la-misma-region-no-cambia-nada",
			contenido: antes + inicio + nueva + fin + despues,
			region:    nueva,
			esperado:  antes + inicio + nueva + fin + despues,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			sustituido, err := skills.SustituirRegion([]byte(caso.contenido), []byte(caso.region))
			require.NoError(t, err)
			assert.Equal(t, caso.esperado, string(sustituido))
		})
	}

	defectos := []struct {
		nombre, contenido, error string
	}{
		{
			nombre:    "sin-marcas",
			contenido: antes + despues,
			error:     "sin las marcas de la tabla de comandos",
		},
		{
			nombre:    "sin-inicio",
			contenido: antes + fin + despues,
			error:     "sin la marca de inicio de la tabla de comandos",
		},
		{
			nombre:    "sin-fin",
			contenido: antes + inicio + despues,
			error:     "sin la marca de fin de la tabla de comandos",
		},
		{
			nombre:    "lineas-que-no-son-exactamente-la-marca",
			contenido: " " + inicio + inicioDeLaTabla + " \n" + inicioDeLaTabla + "\r\n" + "texto " + inicio + fin,
			error:     "sin la marca de inicio de la tabla de comandos",
		},
		{
			nombre:    "dos-inicios",
			contenido: inicio + "vieja\n" + inicio + fin,
			error:     "la marca de inicio de la tabla de comandos aparece 2 veces, en las líneas 1 y 3",
		},
		{
			nombre:    "tres-fines",
			contenido: inicio + fin + fin + "texto\n" + sinFinal,
			error:     "la marca de fin de la tabla de comandos aparece 3 veces, en las líneas 2, 3 y 5",
		},
		{
			nombre:    "fin-antes-del-inicio",
			contenido: fin + "vieja\n" + inicio + despues,
			error: "la marca de fin de la tabla de comandos, en la línea 1, está antes que la de inicio, " +
				"en la línea 3",
		},
	}

	for _, defecto := range defectos {
		t.Run(defecto.nombre, func(t *testing.T) {
			t.Parallel()

			sustituido, err := skills.SustituirRegion([]byte(defecto.contenido), []byte(nueva))
			require.EqualError(t, err, defecto.error)
			assert.Nil(t, sustituido)
		})
	}
}
