package territorio

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

// El territorio sintético de los tests del dominio. Ningún test del paquete
// abre un fichero —la lista core de depguard le deniega os e io también en los
// _test.go (research.md V7, D27)—, así que las fuentes se escriben aquí, en
// Go, y se codifican en YAML con la misma biblioteca que las lee.
//
// Todo es inventado a propósito: los nombres, las comunidades y los boletines
// no son de ningún sitio, y los códigos INE no están en la relación congelada
// —ninguno lo está: la provincia existe y el municipio 99x no—, de modo que
// nada de esto puede tomarse por un dato real escrito de memoria. Las fechas y
// las procedencias son también distintas de las del repositorio, para que
// ningún test pase porque el dominio las tenga escritas.
const (
	fechaDeLaRelacion         = "2026-02-04"
	fechaDeLaCorrespondencia  = "2026-09-21"
	fechaDelEstado            = "2026-09-20"
	fechaDeLasComunidades     = "2026-09-20"
	fuenteDeLaRelacion        = "prueba.relacion"
	fuenteDeLaCorrespondencia = "prueba.correspondencia"
	fuenteDeLosNombres        = "prueba.nombres"
	fuenteDelEstado           = "prueba.estado"
)

// ficherosSinteticos devuelve un territorio completo y coherente, nuevo en
// cada llamada, para que cada caso cambie el suyo sin tocar el de los demás:
//
//   - la comunidad 01, uniprovincial y configurada entera, con su boletín
//     autonómico haciendo también de provincial y el motivo escrito;
//   - la 02, foral y sin ningún boletín configurado;
//   - la 03, de dos provincias, con solo el boletín autonómico configurado.
//
// Entre sus municipios hay un nombre con el artículo pospuesto, uno bilingüe,
// uno con diacríticos, dos que se llaman igual y dos que coinciden por una
// forma alternativa; cuatro tienen DIR3 verificado y los demás no.
func ficherosSinteticos() Ficheros {
	return Ficheros{
		Municipios: FicheroDeMunicipios{
			Fecha:  fechaDeLaRelacion,
			Source: fuenteDeLaRelacion,
			Municipios: map[string]FilaDeMunicipio{
				"28991": {DC: "5", Nombre: "Villaprueba", Provincia: "28", Comunidad: "01"},
				"28992": {DC: "0", Nombre: "Rozas de Prueba, Las", Provincia: "28", Comunidad: "01"},
				"31991": {DC: "3", Nombre: "Iruñeta/Pamploneta", Provincia: "31", Comunidad: "02"},
				"05991": {DC: "7", Nombre: "Villanueva", Provincia: "05", Comunidad: "03"},
				"05992": {DC: "1", Nombre: "Castro, El", Provincia: "05", Comunidad: "03"},
				"47991": {DC: "2", Nombre: "Villanueva", Provincia: "47", Comunidad: "03"},
				"47992": {DC: "9", Nombre: "Peñíscola del Río", Provincia: "47", Comunidad: "03"},
				"47993": {DC: "4", Nombre: "El Castro", Provincia: "47", Comunidad: "03"},
			},
		},
		DIR3: FicheroDeDIR3{
			Fecha:  fechaDeLaCorrespondencia,
			Source: fuenteDeLaCorrespondencia,
			Correspondencia: map[string]string{
				"28991": "L01289915",
				"31991": "L01319913",
				"05991": "L01059917",
				"47992": "L01479929",
			},
		},
		Estado: FicheroDeEstado{
			Fecha:  fechaDelEstado,
			Source: fuenteDelEstado,
			Boletin: BoletinDelEstado{
				Codigo: "BOEP",
				Nombre: "Boletín Oficial del Estado de Prueba",
				URL:    "https://estado.example/",
			},
		},
		Comunidades: map[string]FicheroDeComunidad{
			"01": {
				Fecha:      fechaDeLasComunidades,
				Source:     fuenteDeLosNombres,
				Codigo:     "01",
				Nombre:     "Comunidad Uniprovincial",
				Regimen:    "comun",
				Provincias: map[string]string{"28": "Provincia Única"},
				Boletines: &BoletinesConfigurados{
					Autonomico: &BoletinConfigurado{
						Codigo: "BOCU",
						Nombre: "Boletín Oficial de la Comunidad Uniprovincial",
						URL:    "https://uniprovincial.example/",
					},
					Provincial: &BoletinConfigurado{
						Codigo: "BOCU",
						Nombre: "Boletín Oficial de la Comunidad Uniprovincial",
						URL:    "https://uniprovincial.example/",
						Motivo: "Comunidad uniprovincial: su boletín hace también de boletín provincial.",
					},
				},
			},
			"02": {
				Fecha:      fechaDeLasComunidades,
				Source:     fuenteDeLosNombres,
				Codigo:     "02",
				Nombre:     "Comunidad Foral",
				Regimen:    "foral",
				Provincias: map[string]string{"31": "Provincia Foral"},
			},
			"03": {
				Fecha:      fechaDeLasComunidades,
				Source:     fuenteDeLosNombres,
				Codigo:     "03",
				Nombre:     "Comunidad Pluriprovincial",
				Regimen:    "comun",
				Provincias: map[string]string{"05": "Provincia Primera", "47": "Provincia Segunda"},
				Boletines: &BoletinesConfigurados{
					Autonomico: &BoletinConfigurado{
						Codigo: "BOCP",
						Nombre: "Boletín Oficial de la Comunidad Pluriprovincial",
						URL:    "https://pluriprovincial.example/",
					},
				},
			},
		},
	}
}

// fuentesDe codifica cada fichero en YAML, como los lee Cargar.
func fuentesDe(t *testing.T, ficheros Ficheros) Fuentes {
	t.Helper()

	comunidades := make(map[string][]byte, len(ficheros.Comunidades))
	for codigo, comunidad := range ficheros.Comunidades {
		comunidades[codigo] = enYAML(t, comunidad)
	}

	return Fuentes{
		Municipios:  enYAML(t, ficheros.Municipios),
		DIR3:        enYAML(t, ficheros.DIR3),
		Estado:      enYAML(t, ficheros.Estado),
		Comunidades: comunidades,
	}
}

// enYAML codifica un valor en YAML.
func enYAML(t *testing.T, valor any) []byte {
	t.Helper()

	codificado, err := yaml.Marshal(valor)
	require.NoError(t, err)

	return codificado
}

// cargarSintetico carga el territorio sintético, que tiene que cargar sin
// ningún defecto.
func cargarSintetico(t *testing.T) *Registro {
	t.Helper()

	return cargar(t, ficherosSinteticos())
}

// cargar carga unos ficheros que tienen que cargar sin ningún defecto.
func cargar(t *testing.T, ficheros Ficheros) *Registro {
	t.Helper()

	registro, err := Cargar(fuentesDe(t, ficheros))
	require.NoError(t, err)
	require.NotNil(t, registro)

	return registro
}

// cambiarMunicipio cambia la fila de un municipio de la relación sintética.
func cambiarMunicipio(f *Ficheros, codigo string, cambio func(*FilaDeMunicipio)) {
	fila := f.Municipios.Municipios[codigo]
	cambio(&fila)
	f.Municipios.Municipios[codigo] = fila
}

// cambiarComunidad cambia el fichero de una comunidad sintética.
func cambiarComunidad(f *Ficheros, codigo string, cambio func(*FicheroDeComunidad)) {
	comunidad := f.Comunidades[codigo]
	cambio(&comunidad)
	f.Comunidades[codigo] = comunidad
}

// relacionConClaveRepetida y correspondenciaConClaveRepetida son ficheros
// sintéticos escritos a mano con un municipio dos veces: el mapa de Go no puede
// representarlos, y es justo lo que el lector tiene que rechazar en lugar de
// quedarse con la última.
const (
	relacionConClaveRepetida = `fecha: "2026-02-04"
source: prueba.relacion
municipios:
  "28991": {dc: "5", nombre: "Villaprueba", provincia: "28", comunidad: "01"}
  "28991": {dc: "5", nombre: "Otra Villaprueba", provincia: "28", comunidad: "01"}
`
	correspondenciaConClaveRepetida = `fecha: "2026-09-21"
source: prueba.correspondencia
correspondencia:
  "28991": "L01289915"
  "31991": "L01319913"
  "28991": "L01289915"
`
)

// TestCargar fija lo que Cargar exige a las fuentes antes de construir el
// registro: que cada fichero se lea con las claves de su forma, con su fecha y
// su procedencia; que cada fila tenga lo que el territorio resuelto va a
// emitir, y los seis puntos de integridad entre ficheros de data-model §2.1.
// Un fallo es un error de carga que nombra el fichero y el dato, nunca un
// registro a medias; todos los defectos se dicen juntos (FR-044, FR-054,
// FR-055, SC-003, SC-008).
func TestCargar(t *testing.T) {
	t.Parallel()

	t.Run("completo", func(t *testing.T) {
		t.Parallel()

		registro, err := Cargar(fuentesDe(t, ficherosSinteticos()))
		require.NoError(t, err)
		require.NotNil(t, registro)
		assert.Len(t, registro.porCodigo, len(ficherosSinteticos().Municipios.Municipios),
			"el registro no tiene todos los municipios de la relación")
	})

	for _, caso := range casosDeCarga() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			compruebaDefectosDeCarga(t, caso, fuentesDe)
		})
	}
}

// casoDeCarga son unas fuentes con defectos y los que Cargar tiene que decir.
type casoDeCarga struct {
	nombre string
	// ficheros cambia el territorio sintético antes de codificarlo.
	ficheros func(*Ficheros)
	// fuentes cambia los bytes ya codificados, para lo que un valor de Go no
	// puede representar.
	fuentes func(*Fuentes)
	// defectos son los que el error tiene que decir, cada uno con su fichero
	// delante.
	defectos []string
}

// compruebaDefectosDeCarga codifica el territorio sintético de un caso con
// codificar, le aplica sus cambios y exige que Cargar no dé ningún registro y
// diga todos sus defectos.
func compruebaDefectosDeCarga(t *testing.T, caso casoDeCarga, codificar func(*testing.T, Ficheros) Fuentes) {
	t.Helper()

	ficheros := ficherosSinteticos()
	if caso.ficheros != nil {
		caso.ficheros(&ficheros)
	}

	fuentes := codificar(t, ficheros)
	if caso.fuentes != nil {
		caso.fuentes(&fuentes)
	}

	registro, err := Cargar(fuentes)
	require.Error(t, err, "unas fuentes con defectos no pueden cargar")
	assert.Nil(t, registro, "un error de carga no puede acompañarse de un registro a medias")

	for _, defecto := range caso.defectos {
		require.ErrorContains(t, err, defecto)
	}
}

// casosDeCarga son los casos de TestCargar, nuevos en cada llamada.
func casosDeCarga() []casoDeCarga {
	return []casoDeCarga{
		{
			nombre: "municipio-sin-provincia",
			ficheros: func(f *Ficheros) {
				f.Municipios.Municipios["29991"] = FilaDeMunicipio{
					DC: "1", Nombre: "Pueblo Sin Provincia", Provincia: "29", Comunidad: "01",
				}
			},
			defectos: []string{
				"data/territorio/municipios.yaml: el municipio 29991 es de la provincia 29, que no declara ninguna comunidad",
			},
		},
		{
			nombre: "provincia-en-dos-comunidades",
			ficheros: func(f *Ficheros) {
				cambiarComunidad(f, "02", func(c *FicheroDeComunidad) { c.Provincias["28"] = "Provincia Repetida" })
			},
			defectos: []string{
				"data/territorio/comunidades/02.yaml: declara la provincia 28, que ya declara " +
					"data/territorio/comunidades/01.yaml",
			},
		},
		{
			nombre: "comunidad-de-municipio-discrepante",
			ficheros: func(f *Ficheros) {
				cambiarMunicipio(f, "28991", func(m *FilaDeMunicipio) { m.Comunidad = "03" })
			},
			defectos: []string{
				"data/territorio/municipios.yaml: el municipio 28991 declara la comunidad 03 " +
					"(data/territorio/comunidades/03.yaml) y su provincia 28 es de la comunidad 01 " +
					"(data/territorio/comunidades/01.yaml)",
			},
		},
		{
			nombre: "dir3-de-municipio-inexistente",
			ficheros: func(f *Ficheros) {
				f.DIR3.Correspondencia["28999"] = "L01289991"
			},
			defectos: []string{
				"data/territorio/dir3.yaml: el municipio 28999 no está en data/territorio/municipios.yaml",
			},
		},
		{
			nombre: "dir3-incoherente",
			ficheros: func(f *Ficheros) {
				f.DIR3.Correspondencia["28991"] = "L01289916"
				f.DIR3.Correspondencia["28992"] = "L01289915"
			},
			defectos: []string{
				"data/territorio/dir3.yaml: el DIR3 L01289916 del municipio 28991 no es coherente con su " +
					"código INE y su dígito de control 5",
				"data/territorio/dir3.yaml: el DIR3 L01289915 del municipio 28992 no es coherente con su " +
					"código INE y su dígito de control 0",
			},
		},
		{
			nombre: "regimen-desconocido",
			ficheros: func(f *Ficheros) {
				cambiarComunidad(f, "02", func(c *FicheroDeComunidad) { c.Regimen = "especial" })
			},
			defectos: []string{`data/territorio/comunidades/02.yaml: el régimen "especial" no es comun ni foral`},
		},
		{
			nombre: "codigo-de-fichero-distinto",
			ficheros: func(f *Ficheros) {
				cambiarComunidad(f, "02", func(c *FicheroDeComunidad) { c.Codigo = "04" })
			},
			defectos: []string{
				`data/territorio/comunidades/02.yaml: el fichero es el de la comunidad 02 y declara el código "04"`,
			},
		},
		{
			nombre: "yaml-ilegible",
			fuentes: func(f *Fuentes) {
				f.Comunidades["02"] = []byte("codigo: [\n")
			},
			defectos: []string{"data/territorio/comunidades/02.yaml: no se puede leer: "},
		},
		{
			nombre: "clave-repetida",
			fuentes: func(f *Fuentes) {
				f.Municipios = []byte(relacionConClaveRepetida)
				f.DIR3 = []byte(correspondenciaConClaveRepetida)
			},
			defectos: []string{
				`data/territorio/municipios.yaml: municipios: la clave "28991" está repetida en las líneas 4 y 5`,
				`data/territorio/dir3.yaml: correspondencia: la clave "28991" está repetida en las líneas 4 y 6`,
			},
		},
		{
			nombre: "clave-repetida-en-la-raiz-o-en-una-fila",
			fuentes: func(f *Fuentes) {
				f.Estado = append(f.Estado, "fecha: \"2026-09-21\"\n"...)
				f.Comunidades["03"] = append(f.Comunidades["03"], "codigo: \"03\"\n"...)
				f.Municipios = []byte("fecha: \"2026-02-04\"\nsource: prueba.relacion\nmunicipios:\n" +
					"  \"28991\": {dc: \"5\", dc: \"5\", nombre: Villaprueba, provincia: \"28\", comunidad: \"01\"}\n")
			},
			defectos: []string{
				"data/territorio/estado.yaml: no se puede leer: ",
				`mapping key "fecha" already defined`,
				"data/territorio/comunidades/03.yaml: no se puede leer: ",
				`mapping key "codigo" already defined`,
				"data/territorio/municipios.yaml: municipios: la fila de 28991 no se puede leer: ",
				`mapping key "dc" already defined`,
			},
		},
		{
			nombre: "sin-filas",
			fuentes: func(f *Fuentes) {
				f.Municipios = []byte("fecha: \"2026-02-04\"\nsource: prueba.relacion\n")
				f.DIR3 = []byte("fecha: \"2026-09-21\"\nsource: prueba.correspondencia\ncorrespondencia: null\n")
			},
			defectos: []string{
				"data/territorio/municipios.yaml: municipios: no es un mapa de filas",
				"data/territorio/dir3.yaml: correspondencia: no es un mapa de filas",
			},
		},
		{
			nombre: "filas-ilegibles",
			fuentes: func(f *Fuentes) {
				f.Municipios = []byte("fecha: \"2026-02-04\"\nsource: prueba.relacion\nmunicipios:\n" +
					"  \"28991\": [5, Villaprueba]\n  ? [28992]\n  : {dc: \"0\"}\n")
				f.DIR3 = []byte("fecha: \"2026-09-21\"\nsource: prueba.correspondencia\nclave: desconocida\n")
			},
			defectos: []string{
				"data/territorio/municipios.yaml: municipios: la fila de 28991 no se puede leer: ",
				"data/territorio/municipios.yaml: municipios: la clave de la línea 5 no es un texto",
				"data/territorio/dir3.yaml: no se puede leer: ",
				"field clave not found",
			},
		},
		{
			nombre: "clave-desconocida",
			fuentes: func(f *Fuentes) {
				f.Comunidades["03"] = append(f.Comunidades["03"], "boletin: {}\n"...)
			},
			defectos: []string{"data/territorio/comunidades/03.yaml: no se puede leer: ", "field boletin not found"},
		},
		{
			nombre: "fichero-vacio",
			fuentes: func(f *Fuentes) {
				f.Estado = []byte(" \n")
			},
			defectos: []string{"data/territorio/estado.yaml: está vacío"},
		},
		{
			nombre: "fecha-ilegible",
			ficheros: func(f *Ficheros) {
				f.Municipios.Fecha = "2026-13-45"
			},
			defectos: []string{`data/territorio/municipios.yaml: la fecha "2026-13-45" no es una fecha AAAA-MM-DD`},
		},
		{
			nombre: "sin-source",
			ficheros: func(f *Ficheros) {
				f.DIR3.Source = ""
			},
			defectos: []string{"data/territorio/dir3.yaml: no declara su source"},
		},
		{
			nombre: "codigo-de-municipio-mal-formado",
			ficheros: func(f *Ficheros) {
				f.Municipios.Municipios["2899"] = FilaDeMunicipio{
					DC: "1", Nombre: "Pueblo Corto", Provincia: "28", Comunidad: "01",
				}
			},
			defectos: []string{`data/territorio/municipios.yaml: el código INE "2899" no es válido`},
		},
		{
			nombre: "digito-mal-formado",
			ficheros: func(f *Ficheros) {
				cambiarMunicipio(f, "28991", func(m *FilaDeMunicipio) { m.DC = "55" })
				cambiarMunicipio(f, "28992", func(m *FilaDeMunicipio) { m.DC = "" })
			},
			defectos: []string{
				`data/territorio/municipios.yaml: el municipio 28991 tiene el dígito de control "55", que no es una cifra`,
				`data/territorio/municipios.yaml: el municipio 28992 tiene el dígito de control "", que no es una cifra`,
			},
		},
		{
			nombre: "municipio-sin-nombre",
			ficheros: func(f *Ficheros) {
				cambiarMunicipio(f, "28991", func(m *FilaDeMunicipio) { m.Nombre = "" })
			},
			defectos: []string{"data/territorio/municipios.yaml: el municipio 28991 no tiene nombre"},
		},
		{
			nombre: "provincia-de-municipio-discrepante",
			ficheros: func(f *Ficheros) {
				cambiarMunicipio(f, "28991", func(m *FilaDeMunicipio) { m.Provincia = "05" })
			},
			defectos: []string{
				`data/territorio/municipios.yaml: el municipio 28991 declara la provincia "05" y su código es ` +
					"de la provincia 28",
			},
		},
		{
			nombre: "dir3-mal-formado",
			ficheros: func(f *Ficheros) {
				f.DIR3.Correspondencia["28991"] = "X01289915"
			},
			defectos: []string{`data/territorio/dir3.yaml: el código DIR3 "X01289915" no es válido`},
		},
		{
			nombre: "codigo-de-dir3-mal-formado",
			ficheros: func(f *Ficheros) {
				f.DIR3.Correspondencia["2899"] = "L01289915"
			},
			defectos: []string{`data/territorio/dir3.yaml: el código INE "2899" no es válido`},
		},
		{
			nombre: "boletin-incompleto",
			ficheros: func(f *Ficheros) {
				f.Estado.Boletin.URL = ""
				cambiarComunidad(f, "01", func(c *FicheroDeComunidad) { c.Boletines.Autonomico.Nombre = "" })
				cambiarComunidad(f, "03", func(c *FicheroDeComunidad) { c.Boletines.Autonomico.Codigo = "" })
			},
			defectos: []string{
				"data/territorio/estado.yaml: el boletín estatal no tiene url",
				"data/territorio/comunidades/01.yaml: el boletín autonomico no tiene nombre",
				"data/territorio/comunidades/03.yaml: el boletín autonomico no tiene codigo",
			},
		},
		{
			nombre: "comunidad-y-provincia-sin-nombre",
			ficheros: func(f *Ficheros) {
				cambiarComunidad(f, "02", func(c *FicheroDeComunidad) { c.Nombre = "" })
				cambiarComunidad(f, "03", func(c *FicheroDeComunidad) { c.Provincias["05"] = "" })
			},
			defectos: []string{
				"data/territorio/comunidades/02.yaml: la comunidad no tiene nombre",
				"data/territorio/comunidades/03.yaml: la provincia 05 no tiene nombre",
			},
		},
		{
			nombre: "defectos-en-varios-ficheros",
			ficheros: func(f *Ficheros) {
				f.Estado.Fecha = "ayer"
				cambiarComunidad(f, "03", func(c *FicheroDeComunidad) { c.Regimen = "" })
				cambiarMunicipio(f, "47992", func(m *FilaDeMunicipio) { m.Nombre = "" })
			},
			defectos: []string{
				`data/territorio/estado.yaml: la fecha "ayer" no es una fecha AAAA-MM-DD`,
				`data/territorio/comunidades/03.yaml: el régimen "" no es comun ni foral`,
				"data/territorio/municipios.yaml: el municipio 47992 no tiene nombre",
			},
		},
	}
}
