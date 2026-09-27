package skills_test

import (
	"cmp"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/skills"
)

// skillMdCompleto es un SKILL.md con todas las claves que admite el estándar y
// un cuerpo con otra línea --- y un trozo que no es YAML válido: la lectura solo
// toca el bloque entre la primera línea --- y la siguiente.
const skillMdCompleto = "---\n" +
	"name: boe-legislacion\n" +
	"description: >-\n" +
	"  Consulta y cita normativa\n" +
	"  consolidada del BOE.\n" +
	"license: EUPL-1.2\n" +
	"allowed-tools:\n" +
	"  - Bash\n" +
	"  - Read\n" +
	"compatibility: Claude Code\n" +
	"metadata:\n" +
	"  kitlegal-applets: boe placsp\n" +
	"  kitlegal-referencias: normas\n" +
	"  version: \"1.0\"\n" +
	"---\n" +
	"# Consultar\n" +
	"\n" +
	"---\n" +
	"clave: [sin cerrar\n"

// Trozos de los SKILL.md sintéticos de TestLeerFrontmatter, cada uno con sus
// líneas completas: la línea que abre y cierra el frontmatter y las dos claves
// obligatorias, en las líneas 2 y 3.
const (
	delimitador      = "---\n"
	nameYDescription = "name: boe-legislacion\ndescription: d\n"
)

// TestLeerFrontmatter fija LeerFrontmatter (data-model §1.1 y §1.3; contrato
// sincronizacion-y-comprobacion §4): el frontmatter es el bloque YAML entre la
// primera línea --- y la siguiente; se lee con el lector común, así que una clave
// repetida —también dentro de metadata, y también cuando la repetición llega por
// un alias— es un defecto que nombra la clave y sus dos líneas, que son las de
// SKILL.md; cada valor que no tiene el tipo de su clave es un defecto con su ruta
// y su línea; y la declaración de kitlegal sale de metadata.
func TestLeerFrontmatter(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre      string
		contenido   string
		frontmatter skills.Frontmatter
		deKitlegal  skills.DeclaracionDeKitlegal
		error       string
		defecto     error
	}{
		{
			nombre:    "completo",
			contenido: skillMdCompleto,
			frontmatter: skills.Frontmatter{
				Claves:        []string{"name", "description", "license", "allowed-tools", "compatibility", "metadata"},
				Name:          "boe-legislacion",
				Description:   "Consulta y cita normativa consolidada del BOE.",
				Compatibility: "Claude Code",
				Metadata: map[string]string{
					"kitlegal-applets":     "boe placsp",
					"kitlegal-referencias": "normas",
					"version":              "1.0",
				},
			},
			deKitlegal: skills.DeclaracionDeKitlegal{Applets: []string{"boe", "placsp"}, Referencias: []string{"normas"}},
		},
		{
			nombre:      "allowed-tools-como-cadena",
			contenido:   delimitador + nameYDescription + "allowed-tools: Bash Read\n" + delimitador,
			frontmatter: frontmatterLeido("allowed-tools"),
		},
		{
			nombre:      "sin-metadata",
			contenido:   delimitador + nameYDescription + delimitador,
			frontmatter: frontmatterLeido(),
		},
		{
			nombre:    "metadata-sin-claves-de-kitlegal",
			contenido: delimitador + nameYDescription + "metadata:\n  version: \"1.0\"\n" + delimitador,
			frontmatter: func() skills.Frontmatter {
				leido := frontmatterLeido("metadata")
				leido.Metadata = map[string]string{"version": "1.0"}

				return leido
			}(),
		},
		{
			// Los nombres se leen tal cual, sin suponer ningún fichero de datos: el de
			// cada referencia lo declara la tabla de generadores, y leyes_vertebrales
			// sale de data/normas.yaml.
			nombre: "referencias-que-no-se-llaman-como-sus-datos",
			contenido: delimitador + nameYDescription +
				"metadata:\n  kitlegal-applets: territorio\n  kitlegal-referencias: leyes_vertebrales jerarquia_normativa\n" +
				delimitador,
			frontmatter: func() skills.Frontmatter {
				leido := frontmatterLeido("metadata")
				leido.Metadata = map[string]string{
					"kitlegal-applets":     "territorio",
					"kitlegal-referencias": "leyes_vertebrales jerarquia_normativa",
				}

				return leido
			}(),
			deKitlegal: skills.DeclaracionDeKitlegal{
				Applets:     []string{"territorio"},
				Referencias: []string{"leyes_vertebrales", "jerarquia_normativa"},
			},
		},
		{
			// La lectura no la rechaza: la rechaza ValidarFrontmatter.
			nombre:      "clave-que-no-es-del-estandar",
			contenido:   delimitador + nameYDescription + "user-invocable: true\n" + delimitador,
			frontmatter: frontmatterLeido("user-invocable"),
		},
		{
			nombre:    "sin-delimitadores",
			contenido: nameYDescription,
			error:     "sin frontmatter: la primera línea no es ---",
		},
		{
			nombre:    "vacio",
			contenido: "",
			error:     "sin frontmatter: la primera línea no es ---",
		},
		{
			nombre:    "primera-linea-con-texto-detras",
			contenido: "--- \n" + nameYDescription + delimitador,
			error:     "sin frontmatter: la primera línea no es ---",
		},
		{
			nombre:    "sin-cierre",
			contenido: delimitador + nameYDescription,
			error:     "frontmatter sin cierre: ninguna línea --- después de la primera",
		},
		{
			nombre:    "cierre-con-texto-detras",
			contenido: delimitador + nameYDescription + "--- fin\n",
			error:     "frontmatter sin cierre: ninguna línea --- después de la primera",
		},
		{
			nombre:    "solo-la-apertura-sin-salto",
			contenido: "---",
			error:     "frontmatter sin cierre: ninguna línea --- después de la primera",
		},
		{
			nombre:    "yaml-invalido",
			contenido: delimitador + "name: boe-legislacion\n  description: d\n" + delimitador,
			error:     "no es YAML válido: yaml: line 3: mapping values are not allowed in this context",
		},
		{
			nombre:    "no-es-un-mapa",
			contenido: delimitador + "- boe-legislacion\n" + delimitador,
			error:     "línea 2: el frontmatter no es un mapa",
			defecto:   &skills.DefectoEnElDocumento{Linea: 2, Motivo: "el frontmatter no es un mapa"},
		},
		{
			nombre:    "vacio-entre-delimitadores",
			contenido: delimitador + delimitador,
			error:     "línea 2: el frontmatter no es un mapa",
		},
		{
			nombre:    "name-que-no-es-una-cadena",
			contenido: delimitador + "name: 12\ndescription: d\n" + delimitador,
			error:     "name, línea 2: no es una cadena",
			defecto:   &skills.DefectoEnElDocumento{Ruta: []string{"name"}, Linea: 2, Motivo: "no es una cadena"},
		},
		{
			nombre:    "license-que-no-es-una-cadena",
			contenido: delimitador + nameYDescription + "license: true\n" + delimitador,
			error:     "license, línea 4: no es una cadena",
		},
		{
			nombre:    "allowed-tools-con-un-elemento-que-no-es-una-cadena",
			contenido: delimitador + nameYDescription + "allowed-tools:\n  - Bash\n  - 3\n" + delimitador,
			error:     "allowed-tools/1, línea 6: no es una cadena",
		},
		{
			nombre:    "allowed-tools-que-no-es-cadena-ni-lista",
			contenido: delimitador + nameYDescription + "allowed-tools:\n  Bash: Read\n" + delimitador,
			error:     "allowed-tools, línea 5: no es una cadena ni una lista de cadenas",
		},
		{
			nombre:    "metadata-que-no-es-un-mapa",
			contenido: delimitador + nameYDescription + "metadata: boe\n" + delimitador,
			error:     "metadata, línea 4: no es un mapa de cadena a cadena",
		},
		{
			nombre:    "metadata-con-un-valor-que-es-una-lista",
			contenido: delimitador + nameYDescription + "metadata:\n  kitlegal-applets:\n    - boe\n" + delimitador,
			error:     "metadata/kitlegal-applets, línea 6: no es una cadena",
			defecto: &skills.DefectoEnElDocumento{
				Ruta:   []string{"metadata", "kitlegal-applets"},
				Linea:  6,
				Motivo: "no es una cadena",
			},
		},
		{
			nombre: "metadata-con-valores-que-no-son-cadenas",
			contenido: delimitador + nameYDescription +
				"metadata:\n  kitlegal-applets: true\n  kitlegal-referencias: 3\n" + delimitador,
			error: "metadata/kitlegal-applets, línea 5: no es una cadena\n" +
				"metadata/kitlegal-referencias, línea 6: no es una cadena",
		},
		{
			nombre:    "name-dos-veces",
			contenido: delimitador + nameYDescription + "name: otra\n" + delimitador,
			error:     "name repetido en las líneas 2 y 4",
			defecto:   &skills.ClaveRepetida{Clave: "name", Lineas: [2]int{2, 4}},
		},
		{
			nombre: "kitlegal-applets-dos-veces",
			contenido: delimitador + nameYDescription +
				"metadata:\n  kitlegal-applets: boe\n  kitlegal-applets: placsp\n" + delimitador,
			error:   "metadata: kitlegal-applets repetido en las líneas 5 y 6",
			defecto: &skills.ClaveRepetida{Mapa: []string{"metadata"}, Clave: "kitlegal-applets", Lineas: [2]int{5, 6}},
		},
		{
			nombre:    "name-dos-veces-por-un-alias",
			contenido: delimitador + "&clave name: boe-legislacion\ndescription: d\n*clave : otra\n" + delimitador,
			error:     "name repetido en las líneas 2 y 4",
			defecto:   &skills.ClaveRepetida{Clave: "name", Lineas: [2]int{2, 4}},
		},
		{
			nombre: "kitlegal-applets-dos-veces-por-un-alias",
			contenido: delimitador + nameYDescription +
				"metadata:\n  &applets kitlegal-applets: boe\n  *applets : placsp\n" + delimitador,
			error:   "metadata: kitlegal-applets repetido en las líneas 5 y 6",
			defecto: &skills.ClaveRepetida{Mapa: []string{"metadata"}, Clave: "kitlegal-applets", Lineas: [2]int{5, 6}},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			frontmatter, err := skills.LeerFrontmatter([]byte(caso.contenido))
			if caso.error == "" {
				require.NoError(t, err)
				assert.Equal(t, caso.frontmatter, frontmatter)
				assert.Equal(t, caso.deKitlegal, frontmatter.DeclaracionDeKitlegal())

				return
			}

			require.EqualError(t, err, caso.error)
			assert.Zero(t, frontmatter, "un frontmatter con defectos no se entrega a medias")

			if caso.defecto != nil {
				assert.Equal(t, caso.defecto, primerDefecto(err))
			}
		})
	}
}

// frontmatterLeido es el Frontmatter de un SKILL.md con name boe-legislacion,
// description d y, detrás, las otras claves, sin metadata.
func frontmatterLeido(otrasClaves ...string) skills.Frontmatter {
	return skills.Frontmatter{
		Claves:      append([]string{"name", "description"}, otrasClaves...),
		Name:        "boe-legislacion",
		Description: "d",
	}
}

// primerDefecto es el primero de los errores unidos con errors.Join, o el propio
// error si no es una unión.
func primerDefecto(err error) error {
	var unidos interface{ Unwrap() []error }
	if errors.As(err, &unidos) {
		return unidos.Unwrap()[0]
	}

	return err
}

// Lo que TestValidarFrontmatter da por registrado y por escrito en el
// repositorio.
const skillDePrueba = "boe-legislacion"

// appletsRegistrados son los applets registrados de TestValidarFrontmatter.
var appletsRegistrados = []string{"boe", "placsp"}

// TestValidarFrontmatter fija ValidarFrontmatter (data-model §1.1 y §1.3;
// research.md D3 y D4; FR-040, SC-005): cada regla da un defecto que nombra la
// skill y el defecto, y sus límites exactos son válidos —name de 64 caracteres,
// description de 1024 contados en caracteres y no en bytes, compatibility de
// 500—; las claves del estándar se admiten y las demás no; cada applet declarado
// está registrado y sin repetir; y cada referencia declarada tiene generador en
// la tabla de generadores y el YAML de datos que esa tabla le declara, que no
// tiene por qué llamarse como ella, es un fichero regular del directorio de
// datos (research.md D20). Una referencia sin generador es un defecto aunque haya
// un YAML de datos de su nombre.
func TestValidarFrontmatter(t *testing.T) {
	t.Parallel()

	conDatos := t.TempDir()
	escribirFicheroDePrueba(t, filepath.Join(conDatos, "data", "normas.yaml"), "normas:\n")
	escribirFicheroDePrueba(t, filepath.Join(conDatos, "data", "tributos.yaml"), "tributos:\n")
	require.NoError(t, os.MkdirAll(filepath.Join(conDatos, "data", "jerarquia.yaml"), 0o750))

	sinDatos := t.TempDir()

	casos := []struct {
		nombre   string
		cambio   func(*skills.Frontmatter)
		skill    string
		raiz     string
		defectos []string
	}{
		{nombre: "valido"},
		{
			nombre: "solo-name-y-description",
			cambio: func(f *skills.Frontmatter) {
				f.Claves, f.Compatibility, f.Metadata = []string{"name", "description"}, "", nil
			},
			raiz: sinDatos,
		},
		{
			nombre: "metadata-sin-claves-de-kitlegal",
			cambio: func(f *skills.Frontmatter) { f.Metadata = map[string]string{"version": "1.0"} },
			raiz:   sinDatos,
		},
		{
			nombre: "name-de-64-caracteres",
			cambio: func(f *skills.Frontmatter) { f.Name = nombreDe(64) },
			skill:  nombreDe(64),
		},
		{
			nombre:   "name-de-65-caracteres",
			cambio:   func(f *skills.Frontmatter) { f.Name = nombreDe(65) },
			skill:    nombreDe(65),
			defectos: []string{nombreDe(65) + ": name de 65 caracteres (máximo 64)"},
		},
		{
			nombre:   "sin-name",
			cambio:   func(f *skills.Frontmatter) { f.Claves, f.Name = clavesSin(f.Claves, "name"), "" },
			defectos: []string{"boe-legislacion: falta name"},
		},
		{
			nombre: "name-vacio",
			cambio: func(f *skills.Frontmatter) { f.Name = "" },
			defectos: []string{
				"boe-legislacion: name vacío",
				`boe-legislacion: name "" distinto del nombre del directorio`,
			},
		},
		{
			nombre:   "name-distinto-del-directorio",
			cambio:   func(f *skills.Frontmatter) { f.Name = "boe" },
			defectos: []string{`boe-legislacion: name "boe" distinto del nombre del directorio`},
		},
		{
			nombre:   "name-con-mayuscula",
			cambio:   func(f *skills.Frontmatter) { f.Name = "Boe-legislacion" },
			skill:    "Boe-legislacion",
			defectos: []string{`Boe-legislacion: name "Boe-legislacion" con caracteres que no son a-z, 0-9 ni -`},
		},
		{
			nombre:   "name-con-otro-caracter",
			cambio:   func(f *skills.Frontmatter) { f.Name = "boe_legislación" },
			skill:    "boe_legislación",
			defectos: []string{`boe_legislación: name "boe_legislación" con caracteres que no son a-z, 0-9 ni -`},
		},
		{
			nombre:   "name-con-dos-guiones-seguidos",
			cambio:   func(f *skills.Frontmatter) { f.Name = "boe--legislacion" },
			skill:    "boe--legislacion",
			defectos: []string{`boe--legislacion: name "boe--legislacion" con un guion al principio, al final o dos seguidos`},
		},
		{
			nombre:   "name-empezando-por-guion",
			cambio:   func(f *skills.Frontmatter) { f.Name = "-boe" },
			skill:    "-boe",
			defectos: []string{`-boe: name "-boe" con un guion al principio, al final o dos seguidos`},
		},
		{
			nombre:   "name-terminando-en-guion",
			cambio:   func(f *skills.Frontmatter) { f.Name = "boe-" },
			skill:    "boe-",
			defectos: []string{`boe-: name "boe-" con un guion al principio, al final o dos seguidos`},
		},
		{
			nombre:   "sin-description",
			cambio:   func(f *skills.Frontmatter) { f.Claves, f.Description = clavesSin(f.Claves, "description"), "" },
			defectos: []string{"boe-legislacion: falta description"},
		},
		{
			nombre:   "description-vacia",
			cambio:   func(f *skills.Frontmatter) { f.Description = "" },
			defectos: []string{"boe-legislacion: description vacía"},
		},
		{
			nombre:   "description-solo-con-espacios",
			cambio:   func(f *skills.Frontmatter) { f.Description = " \n\t" },
			defectos: []string{"boe-legislacion: description vacía"},
		},
		{
			nombre: "description-de-1024-caracteres",
			cambio: func(f *skills.Frontmatter) { f.Description = strings.Repeat("á", 1024) },
		},
		{
			nombre:   "description-de-1025-caracteres",
			cambio:   func(f *skills.Frontmatter) { f.Description = strings.Repeat("á", 1025) },
			defectos: []string{"boe-legislacion: description de 1025 caracteres (máximo 1024)"},
		},
		{
			nombre:   "description-con-menor-que",
			cambio:   func(f *skills.Frontmatter) { f.Description = "Consulta <normas> del BOE." },
			defectos: []string{"boe-legislacion: description con < o >"},
		},
		{
			nombre:   "description-con-mayor-que",
			cambio:   func(f *skills.Frontmatter) { f.Description = "Consulta -> cita." },
			defectos: []string{"boe-legislacion: description con < o >"},
		},
		{
			nombre: "compatibility-de-500-caracteres",
			cambio: func(f *skills.Frontmatter) { f.Compatibility = strings.Repeat("é", 500) },
		},
		{
			nombre:   "compatibility-de-501-caracteres",
			cambio:   func(f *skills.Frontmatter) { f.Compatibility = strings.Repeat("é", 501) },
			defectos: []string{"boe-legislacion: compatibility de 501 caracteres (máximo 500)"},
		},
		{
			nombre:   "clave-no-admitida",
			cambio:   func(f *skills.Frontmatter) { f.Claves = append(f.Claves, "user-invocable") },
			defectos: []string{"boe-legislacion: clave no admitida: user-invocable"},
		},
		{
			nombre: "varios-applets",
			cambio: func(f *skills.Frontmatter) { f.Metadata["kitlegal-applets"] = "placsp boe" },
		},
		{
			nombre:   "applet-no-registrado",
			cambio:   func(f *skills.Frontmatter) { f.Metadata["kitlegal-applets"] = "boe cita" },
			defectos: []string{`boe-legislacion: metadata/kitlegal-applets: applet "cita" no registrado`},
		},
		{
			nombre:   "applet-repetido",
			cambio:   func(f *skills.Frontmatter) { f.Metadata["kitlegal-applets"] = "boe placsp boe" },
			defectos: []string{`boe-legislacion: metadata/kitlegal-applets: applet "boe" repetido`},
		},
		{
			nombre:   "applets-separados-por-dos-espacios",
			cambio:   func(f *skills.Frontmatter) { f.Metadata["kitlegal-applets"] = "boe  placsp" },
			defectos: []string{`boe-legislacion: metadata/kitlegal-applets: applet "" no registrado`},
		},
		{
			nombre: "referencia-que-no-se-llama-como-sus-datos",
			cambio: func(f *skills.Frontmatter) { f.Metadata["kitlegal-referencias"] = "normas leyes_vertebrales" },
		},
		{
			nombre:   "referencia-sin-generador",
			cambio:   func(f *skills.Frontmatter) { f.Metadata["kitlegal-referencias"] = "normas tributos" },
			defectos: []string{`boe-legislacion: metadata/kitlegal-referencias: "tributos" sin generador conocido`},
		},
		{
			nombre:   "referencia-cuyo-yaml-de-datos-no-es-un-fichero",
			cambio:   func(f *skills.Frontmatter) { f.Metadata["kitlegal-referencias"] = "jerarquia_normativa" },
			defectos: []string{`boe-legislacion: metadata/kitlegal-referencias: "jerarquia_normativa" sin data/jerarquia.yaml`},
		},
		{
			nombre:   "referencia-sin-directorio-de-datos",
			raiz:     sinDatos,
			defectos: []string{`boe-legislacion: metadata/kitlegal-referencias: "normas" sin data/normas.yaml`},
		},
		{
			nombre:   "referencia-sin-el-yaml-de-datos-de-otro-nombre",
			cambio:   func(f *skills.Frontmatter) { f.Metadata["kitlegal-referencias"] = "leyes_vertebrales" },
			raiz:     sinDatos,
			defectos: []string{`boe-legislacion: metadata/kitlegal-referencias: "leyes_vertebrales" sin data/normas.yaml`},
		},
		{
			nombre: "varios-defectos-en-orden",
			cambio: func(f *skills.Frontmatter) {
				f.Claves = append(clavesSin(f.Claves, "description"), "user-invocable")
				f.Name, f.Description = "boe", ""
				f.Metadata["kitlegal-applets"] = "cita"
				f.Metadata["kitlegal-referencias"] = "tributos"
			},
			defectos: []string{
				`boe-legislacion: name "boe" distinto del nombre del directorio`,
				"boe-legislacion: falta description",
				"boe-legislacion: clave no admitida: user-invocable",
				`boe-legislacion: metadata/kitlegal-applets: applet "cita" no registrado`,
				`boe-legislacion: metadata/kitlegal-referencias: "tributos" sin generador conocido`,
			},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			frontmatter := frontmatterValido()
			if caso.cambio != nil {
				caso.cambio(&frontmatter)
			}

			defectos, err := skills.ValidarFrontmatter(cmp.Or(caso.raiz, conDatos), cmp.Or(caso.skill, skillDePrueba),
				frontmatter, appletsRegistrados)
			require.NoError(t, err)
			assert.Equal(t, caso.defectos, textosDeLosDefectos(t, defectos, cmp.Or(caso.skill, skillDePrueba)))
		})
	}

	t.Run("directorio-de-datos-que-no-es-un-directorio", func(t *testing.T) {
		t.Parallel()

		raiz := t.TempDir()
		escribirFicheroDePrueba(t, filepath.Join(raiz, "data"), "no es un directorio\n")

		defectos, err := skills.ValidarFrontmatter(raiz, skillDePrueba, frontmatterValido(), appletsRegistrados)
		require.ErrorContains(t, err, "el directorio de datos "+filepath.Join(raiz, "data")+" no se puede listar")
		assert.Nil(t, defectos)
	})

	t.Run("sin-referencias-no-lista-los-datos", func(t *testing.T) {
		t.Parallel()

		raiz := t.TempDir()
		escribirFicheroDePrueba(t, filepath.Join(raiz, "data"), "no es un directorio\n")

		frontmatter := frontmatterValido()
		delete(frontmatter.Metadata, "kitlegal-referencias")

		defectos, err := skills.ValidarFrontmatter(raiz, skillDePrueba, frontmatter, appletsRegistrados)
		require.NoError(t, err)
		assert.Empty(t, defectos)
	})
}

// frontmatterValido es un Frontmatter nuevo, con su propio mapa de metadata, que
// cumple todas las reglas en el repositorio con datos de TestValidarFrontmatter.
func frontmatterValido() skills.Frontmatter {
	return skills.Frontmatter{
		Claves:        []string{"name", "description", "license", "allowed-tools", "compatibility", "metadata"},
		Name:          skillDePrueba,
		Description:   "Consulta y cita normativa consolidada del BOE.",
		Compatibility: "Claude Code",
		Metadata: map[string]string{
			"kitlegal-applets":     "boe",
			"kitlegal-referencias": "normas",
			"version":              "1.0",
		},
	}
}

// nombreDe es un nombre de skill con la forma del estándar y la longitud dada, de
// al menos cinco caracteres: boe- seguido de tantas a como falten.
func nombreDe(longitud int) string {
	return "boe-" + strings.Repeat("a", longitud-len("boe-"))
}

// clavesSin son las claves sin la dada, en su orden y en una copia.
func clavesSin(claves []string, retirada string) []string {
	return slices.DeleteFunc(slices.Clone(claves), func(clave string) bool { return clave == retirada })
}

// textosDeLosDefectos son los textos de los defectos, en su orden, tras exigir
// que cada uno nombre la skill; nil si no hay ninguno.
func textosDeLosDefectos(t *testing.T, defectos []*skills.DefectoDeSkill, skill string) []string {
	t.Helper()

	var textos []string

	for _, defecto := range defectos {
		assert.Equal(t, skill, defecto.Skill, "el defecto %q nombra la skill", defecto.Defecto)
		textos = append(textos, defecto.Error())
	}

	return textos
}
