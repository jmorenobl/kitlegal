package territorio

import (
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Lo que fija la configuración lleva como source la ruta del fichero que lo
// fija (FR-005, data-model §2.3 y §2.5).
const (
	rutaDeLaUniprovincial   = "data/territorio/comunidades/01.yaml"
	rutaDeLaForal           = "data/territorio/comunidades/02.yaml"
	rutaDeLaPluriprovincial = "data/territorio/comunidades/03.yaml"
	rutaDelEstadoSintetico  = "data/territorio/estado.yaml"
)

// ochoClaves son las claves de primer nivel del territorio resuelto, en su
// orden (FR-006, data-model §2.5).
var ochoClaves = []string{
	"municipio", "codigo_ine", "provincia", "comunidad", "dir3", "regimen", "boletines", "cobertura",
}

// fecha analiza una fecha AAAA-MM-DD de los tests, a medianoche UTC.
func fecha(t *testing.T, texto string) time.Time {
	t.Helper()

	analizada, err := time.Parse(time.DateOnly, texto)
	require.NoError(t, err)

	return analizada
}

// resolver resuelve una consulta que tiene que dar un territorio.
func resolver(t *testing.T, registro *Registro, entrada string) Territorio {
	t.Helper()

	resuelto, err := registro.Resolver(entrada)
	require.NoError(t, err, "la consulta %q", entrada)

	return resuelto
}

// boletinEstatal es el boletín del estado sintético tal como sale en toda
// respuesta.
var boletinEstatal = Boletin{
	Nivel:  "estatal",
	Codigo: "BOEP",
	Nombre: "Boletín Oficial del Estado de Prueba",
	URL:    "https://estado.example/",
	Source: rutaDelEstadoSintetico,
}

// TestTerritorioResuelto fija el territorio resuelto de data-model §2.5: las
// ocho claves y ninguna más, sin omitempty; el source de cada dato; los
// boletines, siempre el estatal y solo los niveles configurados; la cobertura,
// con sus tres claves siempre y un vocabulario cerrado sin ningún valor que
// signifique «no existe», y la invariante del DIR3 no verificado (FR-006,
// FR-008, FR-020 a FR-023, FR-054, FR-055, SC-002, SC-003, SC-008).
func TestTerritorioResuelto(t *testing.T) {
	t.Parallel()

	registro := cargarSintetico(t)

	t.Run("las ocho claves", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, ochoClaves, clavesDe(reflect.TypeFor[Territorio]()),
			"las claves del territorio resuelto, en su orden")

		for codigo := range ficherosSinteticos().Municipios.Municipios {
			codificado, err := json.Marshal(resolver(t, registro, codigo))
			require.NoError(t, err)

			var claves map[string]json.RawMessage
			require.NoError(t, json.Unmarshal(codificado, &claves))
			assert.ElementsMatch(t, ochoClaves, slices.Collect(maps.Keys(claves)),
				"%s: ni una clave más ni una menos", codigo)
		}
	})

	t.Run("sin omitempty", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, etiquetasConOmitempty(reflect.TypeFor[Territorio](), "data"),
			"todo campo del territorio resuelto se emite siempre (research.md V13)")
	})

	t.Run("municipio cubierto", func(t *testing.T) {
		t.Parallel()

		esperado := Territorio{
			Municipio: Municipio{Nombre: "Villaprueba", Source: fuenteDeLaRelacion},
			CodigoINE: CodigoINE{Codigo: "28991", DigitoDeControl: "5", Source: fuenteDeLaRelacion},
			Provincia: Provincia{Codigo: "28", Nombre: "Provincia Única", Source: fuenteDeLosNombres},
			Comunidad: Comunidad{Codigo: "01", Nombre: "Comunidad Uniprovincial", Source: fuenteDeLosNombres},
			DIR3:      DIR3{Codigo: "L01289915", Source: fuenteDeLaCorrespondencia},
			Regimen:   Regimen{Valor: "comun", Source: rutaDeLaUniprovincial},
			Boletines: []Boletin{
				boletinEstatal,
				{
					Nivel:  "autonomico",
					Codigo: "BOCU",
					Nombre: "Boletín Oficial de la Comunidad Uniprovincial",
					URL:    "https://uniprovincial.example/",
					Source: rutaDeLaUniprovincial,
				},
				{
					Nivel:  "provincial",
					Codigo: "BOCU",
					Nombre: "Boletín Oficial de la Comunidad Uniprovincial",
					URL:    "https://uniprovincial.example/",
					Motivo: "Comunidad uniprovincial: su boletín hace también de boletín provincial.",
					Source: rutaDeLaUniprovincial,
				},
			},
			Cobertura: Cobertura{BoletinAutonomico: "configurado", BoletinProvincial: "configurado", DIR3: "verificado"},
			fecha:     fecha(t, fechaDeLaRelacion),
		}

		assert.Equal(t, esperado, resolver(t, registro, "Villaprueba"))
	})

	t.Run("municipio no cubierto", func(t *testing.T) {
		t.Parallel()

		esperado := Territorio{
			Municipio: Municipio{Nombre: "Iruñeta/Pamploneta", Source: fuenteDeLaRelacion},
			CodigoINE: CodigoINE{Codigo: "31991", DigitoDeControl: "3", Source: fuenteDeLaRelacion},
			Provincia: Provincia{Codigo: "31", Nombre: "Provincia Foral", Source: fuenteDeLosNombres},
			Comunidad: Comunidad{Codigo: "02", Nombre: "Comunidad Foral", Source: fuenteDeLosNombres},
			DIR3:      DIR3{Codigo: "L01319913", Source: fuenteDeLaCorrespondencia},
			Regimen:   Regimen{Valor: "foral", Source: rutaDeLaForal},
			Boletines: []Boletin{boletinEstatal},
			Cobertura: Cobertura{BoletinAutonomico: "no-configurado", BoletinProvincial: "no-configurado", DIR3: "verificado"},
			fecha:     fecha(t, fechaDeLaRelacion),
		}

		resuelto := resolver(t, registro, "Pamploneta")
		assert.Equal(t, esperado, resuelto, "los datos nacionales completos, el régimen y la cobertura parcial")

		codificado, err := json.Marshal(resuelto)
		require.NoError(t, err)

		for _, configurado := range []string{
			"BOCU", "BOCP", "Uniprovincial", "Pluriprovincial", "uniprovincial.example", "pluriprovincial.example",
		} {
			assert.NotContains(t, string(codificado), configurado,
				"fuera del territorio configurado no aparece ningún boletín de otro territorio (FR-021)")
		}
	})

	t.Run("solo los niveles configurados", func(t *testing.T) {
		t.Parallel()

		resuelto := resolver(t, registro, "Peñíscola del Río")
		assert.Equal(t, []Boletin{
			boletinEstatal,
			{
				Nivel:  "autonomico",
				Codigo: "BOCP",
				Nombre: "Boletín Oficial de la Comunidad Pluriprovincial",
				URL:    "https://pluriprovincial.example/",
				Source: rutaDeLaPluriprovincial,
			},
		}, resuelto.Boletines)
		assert.Equal(t,
			Cobertura{BoletinAutonomico: "configurado", BoletinProvincial: "no-configurado", DIR3: "verificado"},
			resuelto.Cobertura)
	})

	t.Run("DIR3 no verificado", func(t *testing.T) {
		t.Parallel()

		resuelto := resolver(t, registro, "Las Rozas de Prueba")
		assert.Equal(t, DIR3{}, resuelto.DIR3, "sin DIR3 verificado no se emite ningún código, tampoco derivado")
		assert.Equal(t, "no-verificado", resuelto.Cobertura.DIR3)
		assert.Equal(t, "28992", resuelto.CodigoINE.Codigo, "el municipio se resuelve igualmente (FR-023)")

		// La invariante, en todos los municipios del territorio sintético:
		// código vacío ⟺ source vacío ⟺ no verificado.
		verificados := 0

		for codigo := range ficherosSinteticos().Municipios.Municipios {
			territorio := resolver(t, registro, codigo)
			sinCodigo, sinSource := territorio.DIR3.Codigo == "", territorio.DIR3.Source == ""
			noVerificado := territorio.Cobertura.DIR3 == "no-verificado"
			assert.True(t, sinCodigo == sinSource && sinSource == noVerificado,
				"%s: DIR3 %+v con cobertura %q", codigo, territorio.DIR3, territorio.Cobertura.DIR3)

			if !noVerificado {
				verificados++
			}
		}

		assert.Equal(t, len(ficherosSinteticos().DIR3.Correspondencia), verificados,
			"verificados son exactamente los municipios de la correspondencia")
	})

	t.Run("cobertura con vocabulario cerrado", func(t *testing.T) {
		t.Parallel()

		vocabulario := AspectosDeCobertura()
		claves := make([]string, 0, len(vocabulario))
		valores := map[string]bool{}

		for _, aspecto := range vocabulario {
			claves = append(claves, aspecto.Clave)
			for _, valor := range aspecto.Valores {
				valores[valor] = true
			}
		}

		tipo := reflect.TypeFor[Cobertura]()
		assert.Equal(t, clavesDe(tipo), claves, "cada aspecto es una clave de la cobertura, en su orden")
		assert.ElementsMatch(t, []string{"configurado", "no-configurado", "verificado", "no-verificado"},
			slices.Collect(maps.Keys(valores)),
			"el vocabulario es exactamente este: ningún valor que signifique «no existe» (FR-022)")

		for indice, aspecto := range vocabulario {
			assert.Equal(t, aspecto.Valores, enumeradoDe(tipo.Field(indice)),
				"el enumerado de %s en el esquema y el vocabulario dicen lo mismo", aspecto.Clave)
		}

		vocabulario[0].Valores[0] = "cambiado"
		assert.Equal(t, "configurado", AspectosDeCobertura()[0].Valores[0],
			"cada llamada devuelve un vocabulario nuevo, que nadie puede cambiar desde fuera")
	})

	t.Run("vocabularios de régimen y nivel", func(t *testing.T) {
		t.Parallel()

		regimen, _ := reflect.TypeFor[Regimen]().FieldByName("Valor")
		assert.Equal(t, []string{"comun", "foral"}, enumeradoDe(regimen))

		nivel, _ := reflect.TypeFor[Boletin]().FieldByName("Nivel")
		assert.Equal(t, []string{"estatal", "autonomico", "provincial"}, enumeradoDe(nivel))
	})

	t.Run("cada dato cumple su etiqueta", func(t *testing.T) {
		t.Parallel()

		for codigo := range ficherosSinteticos().Municipios.Municipios {
			territorio := resolver(t, registro, codigo)
			assert.Empty(t, incumplimientos(reflect.ValueOf(territorio), "data"),
				"%s: la salida tiene que cumplir la descripción formal de sus etiquetas", codigo)
		}
	})

	t.Run("boletines nuevos en cada respuesta", func(t *testing.T) {
		t.Parallel()

		primero := resolver(t, registro, "Villaprueba")
		primero.Boletines[0].Codigo = "cambiado"

		assert.Equal(t, boletinEstatal, resolver(t, registro, "Villaprueba").Boletines[0],
			"cambiar una respuesta no cambia las siguientes")
	})
}

// TestFechaMasAntigua fija la fecha de una respuesta: la más antigua de las de
// los ficheros que la sostienen —la relación, la comunidad del municipio, el
// estado y, solo si trae el DIR3, la correspondencia—, a medianoche UTC, para
// que la cita nunca aparente más frescura que su parte más vieja y dos
// ejecuciones den lo mismo byte a byte (data-model §2.8, research.md D6).
func TestFechaMasAntigua(t *testing.T) {
	t.Parallel()

	t.Run("la más antigua de las dadas", func(t *testing.T) {
		t.Parallel()

		antigua, media, reciente := fecha(t, "2025-01-01"), fecha(t, "2025-06-01"), fecha(t, "2026-01-01")
		assert.Equal(t, antigua, fechaMasAntigua(antigua, media, reciente))
		assert.Equal(t, antigua, fechaMasAntigua(reciente, antigua, media))
		assert.Equal(t, antigua, fechaMasAntigua(media, reciente, antigua))
		assert.Equal(t, media, fechaMasAntigua(media))
	})

	casos := []struct {
		nombre string
		cambio func(*Ficheros)
		// fechas es la fecha que tiene que llevar la respuesta a cada consulta.
		fechas map[string]string
	}{
		{
			nombre: "la relación",
			fechas: map[string]string{"Villaprueba": fechaDeLaRelacion, "Las Rozas de Prueba": fechaDeLaRelacion},
		},
		{
			nombre: "la comunidad del municipio y no otra",
			cambio: func(f *Ficheros) {
				cambiarComunidad(f, "01", func(c *FicheroDeComunidad) { c.Fecha = "2025-12-31" })
			},
			fechas: map[string]string{"Villaprueba": "2025-12-31", "Pamploneta": fechaDeLaRelacion},
		},
		{
			nombre: "el estado",
			cambio: func(f *Ficheros) { f.Estado.Fecha = "2025-01-01" },
			fechas: map[string]string{"Villaprueba": "2025-01-01", "Pamploneta": "2025-01-01"},
		},
		{
			nombre: "la correspondencia, solo si trae el DIR3",
			cambio: func(f *Ficheros) { f.DIR3.Fecha = "2025-06-01" },
			fechas: map[string]string{"Villaprueba": "2025-06-01", "Las Rozas de Prueba": fechaDeLaRelacion},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			ficheros := ficherosSinteticos()
			if caso.cambio != nil {
				caso.cambio(&ficheros)
			}

			registro := cargar(t, ficheros)

			for entrada, esperada := range caso.fechas {
				dada := resolver(t, registro, entrada).Fecha()
				assert.Equal(t, fecha(t, esperada), dada, "la fecha de la respuesta a %q", entrada)
				assert.Equal(t, time.UTC, dada.Location(), "la fecha de la respuesta a %q es UTC", entrada)
				assert.True(t, dada.Equal(dada.Truncate(24*time.Hour)),
					"la fecha de la respuesta a %q es a medianoche", entrada)
			}
		})
	}
}

// clavesDe son las claves JSON de los campos exportados de un tipo estructura,
// en su orden.
func clavesDe(tipo reflect.Type) []string {
	var claves []string

	for campo := range tipo.Fields() {
		if campo.IsExported() {
			clave, _, _ := strings.Cut(campo.Tag.Get("json"), ",")
			claves = append(claves, clave)
		}
	}

	return claves
}

// etiquetasConOmitempty revisa un tipo y los que cuelgan de él y devuelve la
// ruta de cada campo exportado cuya etiqueta json lleva omitempty o no nombra
// su clave.
func etiquetasConOmitempty(tipo reflect.Type, ruta string) []string {
	switch tipo.Kind() {
	case reflect.Slice:
		return etiquetasConOmitempty(tipo.Elem(), ruta+"/*")
	case reflect.Struct:
		var malas []string

		for campo := range tipo.Fields() {
			if !campo.IsExported() {
				continue
			}

			clave, opciones, _ := strings.Cut(campo.Tag.Get("json"), ",")
			hija := ruta + "/" + clave
			if clave == "" || strings.Contains(opciones, "omitempty") {
				malas = append(malas, hija)
			}

			malas = append(malas, etiquetasConOmitempty(campo.Type, hija)...)
		}

		return malas
	default:
		return nil
	}
}

// enumeradoDe son los valores enum de la etiqueta jsonschema de un campo, en
// su orden.
func enumeradoDe(campo reflect.StructField) []string {
	var valores []string

	for parte := range strings.SplitSeq(campo.Tag.Get("jsonschema"), ",") {
		if valor, esEnum := strings.CutPrefix(parte, "enum="); esEnum {
			valores = append(valores, valor)
		}
	}

	return valores
}

// incumplimientos revisa un valor y devuelve, con su ruta, cada dato que no
// cumple la restricción que declara su etiqueta jsonschema: enum, pattern,
// minLength o minItems, que son las que el territorio resuelto usa y de las
// que --describe genera su esquema publicado.
func incumplimientos(valor reflect.Value, ruta string) []string {
	var fallos []string

	switch valor.Kind() {
	case reflect.Slice:
		for indice := range valor.Len() {
			fallos = append(fallos, incumplimientos(valor.Index(indice), ruta+"/"+strconv.Itoa(indice))...)
		}
	case reflect.Struct:
		for indice := range valor.NumField() {
			campo := valor.Type().Field(indice)
			if !campo.IsExported() {
				continue
			}

			clave, _, _ := strings.Cut(campo.Tag.Get("json"), ",")
			hija := ruta + "/" + clave
			fallos = append(fallos, incumplimientosDelCampo(campo.Tag.Get("jsonschema"), valor.Field(indice), hija)...)
			fallos = append(fallos, incumplimientos(valor.Field(indice), hija)...)
		}
	default:
		// Un escalar suelto no tiene etiqueta: la lleva el campo que lo contiene.
	}

	return fallos
}

// incumplimientosDelCampo comprueba un valor contra las restricciones de su
// etiqueta jsonschema.
func incumplimientosDelCampo(etiqueta string, valor reflect.Value, ruta string) []string {
	var fallos, enumerado []string

	for parte := range strings.SplitSeq(etiqueta, ",") {
		nombre, argumento, _ := strings.Cut(parte, "=")

		switch nombre {
		case "enum":
			enumerado = append(enumerado, argumento)
		case "pattern":
			if !regexp.MustCompile(argumento).MatchString(valor.String()) {
				fallos = append(fallos, fmt.Sprintf("%s: %q no casa con %s", ruta, valor.String(), argumento))
			}
		case "minLength", "minItems":
			if minimo, _ := strconv.Atoi(argumento); valor.Len() < minimo {
				fallos = append(fallos, fmt.Sprintf("%s: tiene %d y el mínimo es %d", ruta, valor.Len(), minimo))
			}
		}
	}

	if len(enumerado) > 0 && !slices.Contains(enumerado, valor.String()) {
		fallos = append(fallos, fmt.Sprintf("%s: %q no es ninguno de %q", ruta, valor.String(), enumerado))
	}

	return fallos
}
