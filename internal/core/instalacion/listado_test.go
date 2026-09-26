package instalacion_test

import (
	"encoding/json"
	"reflect"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// casoDeAmbitoIlegible es un ámbito que ni list ni doctor pueden leer (FR-013,
// FR-027, FR-035): su disco, la invocación y la única entrada que nombran.
type casoDeAmbitoIlegible struct {
	nombre string
	// preparar deja el disco del caso, que empieza con el directorio de
	// trabajo vacío y el temporal.
	preparar func(d *discoEnMemoria)
	// invocacion es la de list o doctor; se valida con HOME=homeDePrueba.
	invocacion instalacion.Invocacion
	// motivo es la entrada que se nombra, con su clase.
	motivo instalacion.Conflicto
	// noExaminadas son rutas de las que no se puede examinar nada, ni ellas
	// ni lo que cuelga de ellas.
	noExaminadas []string
}

// casosDeAmbitoIlegible son los de las tres primeras filas de data-model
// §4.1, que list y doctor comprueban igual que install: una guarda del ámbito
// que no es un directorio real, un manifiesto ilegible —también un
// kitlegal.json que es un enlace, aunque apunte a un manifiesto válido— y, con
// --dir, un manifiesto con entradas de host. Con cualquiera de ellos no se
// sabe qué es de quién y no se examina ninguna skill.
func casosDeAmbitoIlegible() []casoDeAmbitoIlegible {
	instaladas := func(d *discoEnMemoria) {
		instalarLocal(d, "boe-legislacion", "legal-core").enlazar("legal-core").escribir()
	}
	sinExaminar := []string{".agents/skills/boe-legislacion", ".agents/skills/legal-core", ".claude"}

	return []casoDeAmbitoIlegible{
		{
			nombre:   ".agents es un fichero",
			preparar: func(d *discoEnMemoria) { d.fichero(".agents", "no soy un directorio") },
			motivo:   conflicto(noEsDirectorio, ".agents"),
		},
		{
			nombre: ".agents es un enlace a un directorio con una instalación",
			preparar: func(d *discoEnMemoria) {
				instalarEn(d, instalacion.NuevoAmbitoDir(temporal+"/dotfiles/skills"), "legal-core").escribir()
				d.enlace(".agents", temporal+"/dotfiles")
			},
			motivo:       conflicto(noEsDirectorio, ".agents"),
			noExaminadas: []string{temporal + "/dotfiles"},
		},
		{
			nombre:   ".agents/skills es un enlace colgando",
			preparar: func(d *discoEnMemoria) { d.enlace(".agents/skills", "no-existe") },
			motivo:   conflicto(noEsDirectorio, ".agents/skills"),
		},
		{
			nombre:   ".agents/skills es una tubería",
			preparar: func(d *discoEnMemoria) { d.tuberia(".agents/skills") },
			motivo:   conflicto(noEsDirectorio, ".agents/skills"),
		},
		{
			nombre: "kitlegal.json es un enlace a un manifiesto válido",
			preparar: func(d *discoEnMemoria) {
				instalarEn(d, instalacion.NuevoAmbitoDir("fuera"), "legal-core").escribir()
				instalarLocal(d, "legal-core")
				d.enlace(".agents/skills/kitlegal.json", "../../fuera/kitlegal.json")
			},
			motivo:       conflicto(ilegible, ".agents/skills/kitlegal.json"),
			noExaminadas: append([]string{"fuera"}, sinExaminar...),
		},
		{
			nombre:   "kitlegal.json es un directorio",
			preparar: func(d *discoEnMemoria) { d.directorio(".agents/skills/kitlegal.json") },
			motivo:   conflicto(ilegible, ".agents/skills/kitlegal.json"),
		},
		{
			nombre:   "kitlegal.json es una tubería",
			preparar: func(d *discoEnMemoria) { d.tuberia(".agents/skills/kitlegal.json") },
			motivo:   conflicto(ilegible, ".agents/skills/kitlegal.json"),
		},
		{
			nombre: "kitlegal.json no respeta su forma",
			preparar: func(d *discoEnMemoria) {
				instaladas(d)
				d.fichero(".agents/skills/kitlegal.json", `{"version": "v0.1.0"}`)
			},
			motivo:       conflicto(ilegible, ".agents/skills/kitlegal.json"),
			noExaminadas: sinExaminar,
		},
		{
			nombre: "kitlegal.json no se puede leer",
			preparar: func(d *discoEnMemoria) {
				instaladas(d)
				d.fallar(opLeer, ".agents/skills/kitlegal.json", errInyectado)
			},
			motivo:       conflicto(ilegible, ".agents/skills/kitlegal.json"),
			noExaminadas: sinExaminar,
		},
		{
			nombre:     "la ruta de --dir es un fichero",
			preparar:   func(d *discoEnMemoria) { d.fichero("destino", "no soy un directorio") },
			invocacion: instalacion.Invocacion{Dir: texto("destino")},
			motivo:     conflicto(noEsDirectorio, "destino"),
		},
		{
			nombre: "la ruta de --dir es un enlace a un directorio con una instalación",
			preparar: func(d *discoEnMemoria) {
				instalarEn(d, instalacion.NuevoAmbitoDir("fuera"), "legal-core").escribir()
				d.enlace("destino", "fuera")
			},
			invocacion:   instalacion.Invocacion{Dir: texto("destino")},
			motivo:       conflicto(noEsDirectorio, "destino"),
			noExaminadas: []string{"fuera"},
		},
		{
			nombre:       "--dir sobre el directorio neutro de una instalación local con entradas de host",
			preparar:     instaladas,
			invocacion:   instalacion.Invocacion{Dir: texto(".agents/skills")},
			motivo:       conflicto(conEntradasDeHost, ".agents/skills/kitlegal.json"),
			noExaminadas: sinExaminar,
		},
		{
			nombre: "-g con $HOME/.agents enlazado desde un repositorio de dotfiles",
			preparar: func(d *discoEnMemoria) {
				instalarEn(d, instalacion.NuevoAmbitoDir(temporal+"/dotfiles/skills"), "legal-core").escribir()
				d.directorio(homeDePrueba)
				d.enlace(homeDePrueba+"/.agents", temporal+"/dotfiles")
			},
			invocacion:   global,
			motivo:       conflicto(noEsDirectorio, homeDePrueba+"/.agents"),
			noExaminadas: []string{temporal + "/dotfiles"},
		},
	}
}

// leerElAmbito es list o doctor sobre el disco en el ámbito, que exige lo
// propio de su verbo de lo que devuelve con el error y devuelve el error.
type leerElAmbito func(t *testing.T, d *discoEnMemoria, ambito instalacion.Ambito) error

// probarAmbitosIlegibles comprueba, en cada caso de casosDeAmbitoIlegible, que
// leer, list o doctor, termina con el error del ámbito ilegible del verbo, sin
// tocar el disco y sin examinar nada por debajo de lo que nombra.
func probarAmbitosIlegibles(t *testing.T, verbo string, leer leerElAmbito) {
	t.Helper()

	for _, caso := range casosDeAmbitoIlegible() {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			d := nuevoDiscoEnMemoria(t)
			d.directorio(temporal)
			caso.preparar(d)

			pedido, err := instalacion.ValidarInvocacion(caso.invocacion, homeDePrueba, empotradasDePrueba())
			require.NoError(t, err, "la invocación del caso")

			antes := d.instantanea()

			exigirAmbitoIlegible(t, leer(t, d, pedido.Ambito), verbo, caso.motivo)
			assert.Equal(t, antes, d.instantanea(), "%s no cambia nada (FR-062, FR-068)", verbo)
			exigirDiscoRespetado(t, d, raizVigilada(pedido.Ambito))
			exigirNoExaminadas(t, d, caso.noExaminadas...)
		})
	}
}

// exigirAmbitoIlegible exige que err sea el error del ámbito ilegible del
// verbo que nombra motivo: de clase «inesperado», que sale con código 1, y con
// el mensaje de una línea de contracts/applet-skills.md §5, «skills <verbo>:
// <clase>: <ruta>».
func exigirAmbitoIlegible(t *testing.T, err error, verbo string, motivo instalacion.Conflicto) {
	t.Helper()

	var rechazo *instalacion.AmbitoIlegible
	require.ErrorAs(t, err, &rechazo, "el ámbito ilegible llega en el error tipado")
	assert.Equal(t, motivo, rechazo.Motivo())

	var conClase schema.ConClase
	require.ErrorAs(t, err, &conClase, "el ámbito ilegible declara su clase")
	assert.Equal(t, schema.ClaseInesperado, conClase.Clase())

	assert.Equal(t, "skills "+verbo+": "+string(motivo.Clase)+": "+motivo.Ruta, err.Error())
}

// casoDeListado es una fila de TestListar: un disco, una invocación y lo que
// list tiene que dar.
type casoDeListado struct {
	nombre string
	// preparar deja el disco del caso, que empieza con el directorio de
	// trabajo vacío y el temporal.
	preparar func(d *discoEnMemoria)
	// invocacion es la de list; se valida con HOME=homeDePrueba.
	invocacion instalacion.Invocacion
	// esperado es el data de list.
	esperado instalacion.Listado
	// noExaminadas son rutas de las que no se puede examinar nada, ni ellas
	// ni lo que cuelga de ellas.
	noExaminadas []string
}

// TestListar fija list (contracts/applet-skills.md §4.2; FR-060 a FR-062;
// SC-011): el directorio neutro del ámbito, si hay manifiesto, su versión, nula
// sin él, y cada skill declarada en orden de nombre, con su ruta, la versión
// de su instalación, si el binario la empotra y sus entradas de host con su modo,
// también las que el binario no empotra (FR-036). En cada fila, list solo
// examina las guardas del ámbito y el manifiesto, sin seguir ningún enlace, y
// no cambia nada en el disco. Con un ámbito que no se puede leer, un error que
// lo nombra y ninguna skill; con un fallo de entrada y salida, ese fallo.
func TestListar(t *testing.T) {
	t.Parallel()

	for _, caso := range casosDeListado(t) {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			probarCasoDeListado(t, caso)
		})
	}

	t.Run("ámbito ilegible", func(t *testing.T) {
		t.Parallel()

		probarAmbitosIlegibles(t, "list", func(t *testing.T, d *discoEnMemoria, ambito instalacion.Ambito) error {
			t.Helper()

			listado, err := instalacion.Listar(d, ambito, empotradasDePrueba())
			assert.Equal(t, instalacion.Listado{}, listado, "ninguna skill")

			return err
		})
	})

	t.Run("fallos de entrada y salida", probarFallosAlListar)
	t.Run("la salida", probarSalidaDeList)
}

// probarCasoDeListado lista en el disco del caso y lo compara con lo esperado.
func probarCasoDeListado(t *testing.T, caso casoDeListado) {
	t.Helper()

	d := nuevoDiscoEnMemoria(t)
	d.directorio(temporal)
	caso.preparar(d)

	pedido, err := instalacion.ValidarInvocacion(caso.invocacion, homeDePrueba, empotradasDePrueba())
	require.NoError(t, err, "la invocación del caso")

	antes := d.instantanea()

	listado, err := instalacion.Listar(d, pedido.Ambito, empotradasDePrueba())
	require.NoError(t, err)

	assert.Equal(t, caso.esperado, listado)
	assert.Equal(t, antes, d.instantanea(), "list no escribe nada (FR-062)")
	exigirDiscoRespetado(t, d, raizVigilada(pedido.Ambito))
	exigirNoExaminadas(t, d, caso.noExaminadas...)

	leibles := slices.Concat(pedido.Ambito.Guardas(), []string{pedido.Ambito.RutaDelManifiesto()})
	for _, pedida := range d.accesos {
		assert.True(t, slices.ContainsFunc(leibles, func(ruta string) bool { return absoluta(ruta) == pedida.ruta }),
			"list solo examina las guardas del ámbito y el manifiesto: %s %s", pedida.operacion, pedida.ruta)
	}
}

// casosDeListado son los de TestListar con un ámbito que se puede leer.
func casosDeListado(t *testing.T) []casoDeListado {
	t.Helper()

	deHome := ambitoGlobal(t, homeDePrueba)
	enDestino := instalacion.NuevoAmbitoDir("destino")

	return []casoDeListado{
		{
			nombre:   "sin .agents",
			preparar: func(*discoEnMemoria) {},
			esperado: sinManifiesto(ambitoLocal),
		},
		{
			nombre:   "sin .agents/skills",
			preparar: func(d *discoEnMemoria) { d.directorio(".agents") },
			esperado: sinManifiesto(ambitoLocal),
		},
		{
			nombre: "el directorio neutro sin kitlegal.json, con una skill dentro",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "legal-core")
				d.directorio(".claude")
			},
			esperado:     sinManifiesto(ambitoLocal),
			noExaminadas: []string{".agents/skills/legal-core", ".claude"},
		},
		{
			nombre:   "un manifiesto que no declara ninguna skill",
			preparar: func(d *discoEnMemoria) { instalarLocal(d).escribir() },
			esperado: listadoDe(ambitoLocal, versionDePrueba),
		},
		{
			nombre: "las dos, enlazadas en el host",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "boe-legislacion", "legal-core").enlazar("boe-legislacion").enlazar("legal-core").escribir()
			},
			esperado: listadoDe(ambitoLocal, versionDePrueba,
				listadaEn(ambitoLocal, "boe-legislacion", versionDePrueba, true, instalacion.ModoEnlace),
				listadaEn(ambitoLocal, "legal-core", versionDePrueba, true, instalacion.ModoEnlace)),
			noExaminadas: []string{".agents/skills/boe-legislacion", ".agents/skills/legal-core", ".claude"},
		},
		{
			nombre: "una en copia, otra sin host, y cada una de su versión",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "boe-legislacion", "legal-core").copiar("boe-legislacion").
					skillDeVersion("legal-core", versionVieja).escribir()
			},
			esperado: listadoDe(ambitoLocal, versionDePrueba,
				listadaEn(ambitoLocal, "boe-legislacion", versionDePrueba, true, instalacion.ModoCopia),
				listadaEn(ambitoLocal, "legal-core", versionVieja, true)),
			noExaminadas: []string{".claude"},
		},
		{
			nombre: "una skill que el binario no empotra, de otra versión",
			preparar: func(d *discoEnMemoria) {
				instalarLocal(d, "legal-core").noEmpotrada("otra-skill", versionVieja).escribir()
			},
			esperado: listadoDe(ambitoLocal, versionDePrueba,
				listadaEn(ambitoLocal, "legal-core", versionDePrueba, true),
				listadaEn(ambitoLocal, "otra-skill", versionVieja, false, instalacion.ModoEnlace)),
			noExaminadas: []string{".agents/skills/otra-skill", ".claude"},
		},
		{
			nombre: "-g: rutas absolutas",
			preparar: func(d *discoEnMemoria) {
				instalarEn(d, deHome, "boe-legislacion", "legal-core").enlazar("legal-core").escribir()
			},
			invocacion: global,
			esperado: listadoDe(deHome, versionDePrueba,
				listadaEn(deHome, "boe-legislacion", versionDePrueba, true),
				listadaEn(deHome, "legal-core", versionDePrueba, true, instalacion.ModoEnlace)),
		},
		{
			nombre:     "-g con un HOME que no existe",
			preparar:   func(*discoEnMemoria) {},
			invocacion: global,
			esperado:   sinManifiesto(deHome),
		},
		{
			nombre:     "--dir, con la ruta limpia",
			preparar:   func(d *discoEnMemoria) { instalarEn(d, enDestino, "legal-core").escribir() },
			invocacion: instalacion.Invocacion{Dir: texto("destino/")},
			esperado: listadoDe(enDestino, versionDePrueba,
				listadaEn(enDestino, "legal-core", versionDePrueba, true)),
		},
		{
			nombre:     "--dir que no existe",
			preparar:   func(*discoEnMemoria) {},
			invocacion: instalacion.Invocacion{Dir: texto("otro/destino")},
			esperado:   sinManifiesto(instalacion.NuevoAmbitoDir("otro/destino")),
		},
	}
}

// sinManifiesto es lo que da list en un ámbito sin manifiesto: su directorio
// neutro, manifiesto no, la versión nula y ninguna skill (FR-061).
func sinManifiesto(ambito instalacion.Ambito) instalacion.Listado {
	return instalacion.Listado{Directorio: ambito.Neutro(), Skills: []instalacion.SkillListada{}}
}

// listadoDe es lo que da list en un ámbito con un manifiesto de esa versión
// que declara esas skills.
func listadoDe(ambito instalacion.Ambito, version string, skills ...instalacion.SkillListada) instalacion.Listado {
	return instalacion.Listado{
		Directorio: ambito.Neutro(),
		Manifiesto: true,
		Version:    &version,
		Skills:     append([]instalacion.SkillListada{}, skills...),
	}
}

// listadaEn es la skill nombre del ámbito tal como la da list, declarada de
// version, empotrada o no y, por cada modo, con su entrada en el host claude
// en ese modo.
func listadaEn(
	ambito instalacion.Ambito, nombre, version string, empotrada bool, modos ...instalacion.Modo,
) instalacion.SkillListada {
	enlaces := []instalacion.Enlace{}
	for _, modo := range modos {
		enlaces = append(enlaces, instalacion.Enlace{Host: "claude", Ruta: ambito.RutaDeHost(nombre), Modo: modo})
	}

	return instalacion.SkillListada{
		Nombre:    nombre,
		Ruta:      ambito.RutaDeSkill(nombre),
		Version:   version,
		Empotrada: empotrada,
		Enlaces:   enlaces,
	}
}

// probarFallosAlListar exige que un fallo al examinar una guarda o el
// manifiesto se devuelva tal cual: no es un ámbito ilegible, que sí es no
// poder leer el manifiesto (FR-035).
func probarFallosAlListar(t *testing.T) {
	t.Parallel()

	for _, ruta := range []string{".agents", ".agents/skills", ".agents/skills/kitlegal.json"} {
		t.Run(ruta, func(t *testing.T) {
			t.Parallel()

			d := nuevoDiscoEnMemoria(t)
			instalarLocal(d, "legal-core").escribir()
			d.fallar(opExaminar, ruta, errInyectado)

			listado, err := instalacion.Listar(d, ambitoLocal, empotradasDePrueba())
			require.ErrorIs(t, err, errInyectado)
			assert.Equal(t, instalacion.Listado{}, listado)

			var rechazo *instalacion.AmbitoIlegible
			assert.NotErrorAs(t, err, &rechazo, "un fallo de entrada y salida no es un ámbito ilegible")
		})
	}
}

// probarSalidaDeList fija la salida de list en JSON, la de
// contracts/applet-skills.md §4.2, con la biblioteca con la que la escribe el
// kernel: sus claves en español y en su orden, la versión nula sin manifiesto
// y una lista vacía como [], nunca null. Y que la etiqueta jsonschema de la
// versión la declara nula, de la que --describe saca el esquema (research.md
// D15).
func probarSalidaDeList(t *testing.T) {
	t.Parallel()

	listados := map[string]instalacion.Listado{
		`{"directorio":".agents/skills","manifiesto":false,"version":null,"skills":[]}`: sinManifiesto(ambitoLocal),
		`{"directorio":".agents/skills","manifiesto":true,"version":"v0.1.0","skills":[` +
			`{"nombre":"boe-legislacion","ruta":".agents/skills/boe-legislacion","version":"v0.1.0","empotrada":true,` +
			`"enlaces":[{"host":"claude","ruta":".claude/skills/boe-legislacion","modo":"enlace"}]},` +
			`{"nombre":"otra-skill","ruta":".agents/skills/otra-skill","version":"v0.0.9","empotrada":false,` +
			`"enlaces":[]}]}`: listadoDe(ambitoLocal, versionDePrueba,
			listadaEn(ambitoLocal, "boe-legislacion", versionDePrueba, true, instalacion.ModoEnlace),
			listadaEn(ambitoLocal, "otra-skill", versionVieja, false)),
	}

	for esperado, listado := range listados {
		salida, err := json.Marshal(listado)
		require.NoError(t, err)
		assert.Equal(t, esperado, string(salida), "las claves en el orden del contrato")
	}

	version, hay := reflect.TypeFor[instalacion.Listado]().FieldByName("Version")
	require.True(t, hay)
	assert.Equal(t, "nullable,minLength=1", version.Tag.Get("jsonschema"), "la versión, nula sin manifiesto (FR-053)")
}
