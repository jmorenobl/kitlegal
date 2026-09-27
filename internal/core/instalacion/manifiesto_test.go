package instalacion_test

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
)

// canonicoDeReferencia es, byte a byte, el manifiesto de manifiestoDeReferencia
// en su forma canónica (contracts/manifiesto.md §2): claves de todo objeto en
// orden, sangrado de dos espacios, «": "» entre clave y valor y un salto de
// línea final. Declara las tres formas de una skill: con su entrada de host en
// enlace, en copia y sin entrada de host.
const canonicoDeReferencia = `{
  "skills": {
    "boe-legislacion": {
      "ficheros": {
        "boe-legislacion/SKILL.md": "` + huellaDeA + `",
        "boe-legislacion/references/normas.md": "` + huellaDeB + `"
      },
      "hosts": {
        "claude": {
          "modo": "enlace",
          "ruta": ".claude/skills/boe-legislacion"
        }
      },
      "version": "v0.1.0"
    },
    "legal-core": {
      "ficheros": {
        "legal-core/SKILL.md": "` + huellaDeC + `",
        "legal-core/references/jerarquia_normativa.md": "` + huellaDeABC + `",
        "legal-core/references/leyes_vertebrales.md": "` + huellaVacia + `"
      },
      "hosts": {
        "claude": {
          "ficheros": {
            ".claude/skills/legal-core/SKILL.md": "` + huellaDeC + `",
            ".claude/skills/legal-core/references/jerarquia_normativa.md": "` + huellaDeABC + `",
            ".claude/skills/legal-core/references/leyes_vertebrales.md": "` + huellaVacia + `"
          },
          "modo": "copia",
          "ruta": ".claude/skills/legal-core"
        }
      },
      "version": "v0.1.0"
    },
    "otra-skill": {
      "ficheros": {
        "otra-skill/SKILL.md": "` + huellaDeA + `"
      },
      "version": "v0.0.9"
    }
  },
  "version": "v0.1.0"
}
`

// compactoDeReferencia es el mismo manifiesto sin blancos y con las claves de
// cada objeto en otro orden: legible, pero no canónico.
const compactoDeReferencia = `{"version":"v0.1.0","skills":{` +
	`"otra-skill":{"version":"v0.0.9","ficheros":{"otra-skill/SKILL.md":"` + huellaDeA + `"}},` +
	`"legal-core":{"version":"v0.1.0","hosts":{"claude":{"ruta":".claude/skills/legal-core","modo":"copia",` +
	`"ficheros":{".claude/skills/legal-core/references/leyes_vertebrales.md":"` + huellaVacia + `",` +
	`".claude/skills/legal-core/references/jerarquia_normativa.md":"` + huellaDeABC + `",` +
	`".claude/skills/legal-core/SKILL.md":"` + huellaDeC + `"}}},` +
	`"ficheros":{"legal-core/references/leyes_vertebrales.md":"` + huellaVacia + `",` +
	`"legal-core/references/jerarquia_normativa.md":"` + huellaDeABC + `",` +
	`"legal-core/SKILL.md":"` + huellaDeC + `"}},` +
	`"boe-legislacion":{"version":"v0.1.0","hosts":{"claude":{"ruta":".claude/skills/boe-legislacion",` +
	`"modo":"enlace"}},"ficheros":{"boe-legislacion/references/normas.md":"` + huellaDeB + `",` +
	`"boe-legislacion/SKILL.md":"` + huellaDeA + `"}}}}`

// manifiestoDeReferencia es lo que se lee de canonicoDeReferencia, nuevo en
// cada llamada.
func manifiestoDeReferencia() instalacion.Manifiesto {
	return instalacion.Manifiesto{
		Version: "v0.1.0",
		Skills: map[string]instalacion.SkillDeclarada{
			"boe-legislacion": {
				Version: "v0.1.0",
				Ficheros: map[string]string{
					"boe-legislacion/SKILL.md":             huellaDeA,
					"boe-legislacion/references/normas.md": huellaDeB,
				},
				Claude: &instalacion.EntradaDeHost{Ruta: ".claude/skills/boe-legislacion", Modo: instalacion.ModoEnlace},
			},
			"legal-core": {
				Version: "v0.1.0",
				Ficheros: map[string]string{
					"legal-core/SKILL.md":                          huellaDeC,
					"legal-core/references/jerarquia_normativa.md": huellaDeABC,
					"legal-core/references/leyes_vertebrales.md":   huellaVacia,
				},
				Claude: &instalacion.EntradaDeHost{
					Ruta: ".claude/skills/legal-core",
					Modo: instalacion.ModoCopia,
					Ficheros: map[string]string{
						".claude/skills/legal-core/SKILL.md":                          huellaDeC,
						".claude/skills/legal-core/references/jerarquia_normativa.md": huellaDeABC,
						".claude/skills/legal-core/references/leyes_vertebrales.md":   huellaVacia,
					},
				},
			},
			"otra-skill": {
				Version:  "v0.0.9",
				Ficheros: map[string]string{"otra-skill/SKILL.md": huellaDeA},
			},
		},
	}
}

// Piezas legibles con las que se escriben los manifiestos de las filas: cada
// fila cambia una sola cosa respecto de un manifiesto legible.
const (
	// ficherosDeLegalCore es un miembro ficheros legible de legal-core.
	ficherosDeLegalCore = `{"legal-core/SKILL.md": "` + huellaDeC + `"}`
	// enlaceDeLegalCore es una entrada de host legible de legal-core.
	enlaceDeLegalCore = `{"modo": "enlace", "ruta": ".claude/skills/legal-core"}`
	// entradaDeLegalCore es una entrada de skill legible de legal-core.
	entradaDeLegalCore = `{"ficheros": ` + ficherosDeLegalCore + `, "version": "v0.1.0"}`
	// entradaSinFicheros es una entrada de skill legible de cualquier nombre.
	entradaSinFicheros = `{"ficheros": {}, "version": "v0.1.0"}`
)

// conVersion es el manifiesto sin skills cuyo miembro version es el literal
// JSON version, escrito tal cual.
func conVersion(version string) string {
	return `{"skills": {}, "version": ` + version + `}`
}

// conSkills es el manifiesto de la versión v0.1.0 cuyo miembro skills es el
// literal JSON skills.
func conSkills(skills string) string {
	return `{"skills": ` + skills + `, "version": "v0.1.0"}`
}

// conNombre es el manifiesto de una sola skill, legible salvo por su nombre,
// que va tal cual entre comillas.
func conNombre(nombre string) string {
	return conSkills(`{"` + nombre + `": ` + entradaSinFicheros + `}`)
}

// conLegalCore es el manifiesto de una sola skill, legal-core, cuya entrada es
// el literal JSON entrada.
func conLegalCore(entrada string) string {
	return conSkills(`{"legal-core": ` + entrada + `}`)
}

// conFicheros es el de legal-core sin hosts y con el miembro ficheros literal.
func conFicheros(ficheros string) string {
	return conLegalCore(`{"ficheros": ` + ficheros + `, "version": "v0.1.0"}`)
}

// conFichero es el de legal-core con un único fichero declarado, de ruta y
// huella dadas, escritas tal cual entre comillas.
func conFichero(ruta, huella string) string {
	return conFicheros(`{"` + ruta + `": "` + huella + `"}`)
}

// conHosts es el de legal-core con ficheros legibles y el miembro hosts
// literal.
func conHosts(hosts string) string {
	return conLegalCore(`{"ficheros": ` + ficherosDeLegalCore + `, "hosts": ` + hosts + `, "version": "v0.1.0"}`)
}

// conClaude es el de legal-core con la entrada de host claude literal.
func conClaude(claude string) string {
	return conHosts(`{"claude": ` + claude + `}`)
}

// conCopia es el de legal-core con una copia de host cuyo miembro ficheros es
// el literal JSON ficheros.
func conCopia(ficheros string) string {
	return conClaude(`{"ficheros": ` + ficheros + `, "modo": "copia", "ruta": ".claude/skills/legal-core"}`)
}

// conFicheroDeLaCopia es el de legal-core con una copia de host de un único
// fichero, de ruta y huella dadas.
func conFicheroDeLaCopia(ruta, huella string) string {
	return conCopia(`{"` + ruta + `": "` + huella + `"}`)
}

// TestLeerManifiesto fija la lectura estricta de kitlegal.json
// (contracts/manifiesto.md §1 y §3; research.md D10): lo que respeta la forma
// se lee a un Manifiesto, y toda desviación —no ser JSON, un miembro
// desconocido, un nombre repetido en cualquier objeto, datos tras el
// documento o cualquier regla de la tabla del contrato— es un manifiesto
// ilegible, con su tipo y sin manifiesto (FR-035). Una fila por regla.
func TestLeerManifiesto(t *testing.T) {
	t.Parallel()

	t.Run("legibles", probarManifiestosLegibles)
	t.Run("ilegibles por el documento", probarIlegiblesPorElDocumento)
	t.Run("ilegibles por la forma", probarIlegiblesPorLaForma)
}

// probarManifiestosLegibles lee cada documento que respeta la forma y exige
// exactamente el manifiesto que declara: cada objeto del documento, también
// vacío, como un mapa; la entrada de host, si la hay; sus ficheros solo en
// copia; y las cadenas ya sin escapes JSON.
func probarManifiestosLegibles(t *testing.T) {
	t.Parallel()

	nombreMaximo := strings.Repeat("a", 64)

	casos := []struct {
		nombre     string
		contenido  string
		manifiesto instalacion.Manifiesto
	}{
		{
			nombre:     "el mínimo, sin ninguna skill",
			contenido:  `{"skills": {}, "version": "dev"}`,
			manifiesto: instalacion.Manifiesto{Version: "dev", Skills: map[string]instalacion.SkillDeclarada{}},
		},
		{
			nombre:     "la forma de referencia",
			contenido:  canonicoDeReferencia,
			manifiesto: manifiestoDeReferencia(),
		},
		{
			nombre:     "compacto y con las claves en otro orden",
			contenido:  compactoDeReferencia,
			manifiesto: manifiestoDeReferencia(),
		},
		{
			nombre:     "con blancos antes y después del documento",
			contenido:  "\n \t" + `{"skills": {}, "version": "dev"}` + "\n\n \t",
			manifiesto: instalacion.Manifiesto{Version: "dev", Skills: map[string]instalacion.SkillDeclarada{}},
		},
		{
			nombre: "versiones sin forma de SemVer, que solo no pueden estar vacías ni llevar controles",
			contenido: `{"skills": {"legal-core": {"ficheros": {}, "version": "desarrollo local ñ"}},` +
				` "version": "0.1.0"}`,
			manifiesto: instalacion.Manifiesto{Version: "0.1.0", Skills: map[string]instalacion.SkillDeclarada{
				"legal-core": {Version: "desarrollo local ñ", Ficheros: map[string]string{}},
			}},
		},
		{
			nombre: "nombres de skill en los extremos de su forma",
			contenido: conSkills(`{"a": ` + entradaSinFicheros + `, "0": ` + entradaSinFicheros +
				`, "a-1-b": ` + entradaSinFicheros + `, "` + nombreMaximo + `": ` + entradaSinFicheros + `}`),
			manifiesto: instalacion.Manifiesto{Version: "v0.1.0", Skills: map[string]instalacion.SkillDeclarada{
				"a":          {Version: "v0.1.0", Ficheros: map[string]string{}},
				"0":          {Version: "v0.1.0", Ficheros: map[string]string{}},
				"a-1-b":      {Version: "v0.1.0", Ficheros: map[string]string{}},
				nombreMaximo: {Version: "v0.1.0", Ficheros: map[string]string{}},
			}},
		},
		{
			nombre: "elementos de ruta que empiezan por punto sin ser . ni ..",
			contenido: conFicheros(`{"legal-core/.oculto": "` + huellaDeA + `", "legal-core/references/..md": "` +
				huellaDeB + `", "legal-core/...": "` + huellaDeC + `"}`),
			manifiesto: instalacion.Manifiesto{Version: "v0.1.0", Skills: map[string]instalacion.SkillDeclarada{
				"legal-core": {Version: "v0.1.0", Ficheros: map[string]string{
					"legal-core/.oculto":         huellaDeA,
					"legal-core/references/..md": huellaDeB,
					"legal-core/...":             huellaDeC,
				}},
			}},
		},
		{
			nombre: "cadenas con escapes JSON",
			contenido: `{"skills": {"legal-core": {"ficheros": {"legal-core\/SKILL.md": "` + huellaDeC +
				`"}, "version": "v0.1.0+abc"}}, "version": "v0.1.0"}`,
			manifiesto: instalacion.Manifiesto{Version: "v0.1.0", Skills: map[string]instalacion.SkillDeclarada{
				"legal-core": {Version: "v0.1.0+abc", Ficheros: map[string]string{"legal-core/SKILL.md": huellaDeC}},
			}},
		},
		{
			nombre:    "copia de host sin ficheros declarados",
			contenido: conCopia(`{}`),
			manifiesto: instalacion.Manifiesto{Version: "v0.1.0", Skills: map[string]instalacion.SkillDeclarada{
				"legal-core": {
					Version:  "v0.1.0",
					Ficheros: map[string]string{"legal-core/SKILL.md": huellaDeC},
					Claude: &instalacion.EntradaDeHost{
						Ruta: ".claude/skills/legal-core", Modo: instalacion.ModoCopia, Ficheros: map[string]string{},
					},
				},
			}},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			manifiesto, err := instalacion.LeerManifiesto([]byte(caso.contenido))
			require.NoError(t, err)
			assert.Equal(t, caso.manifiesto, manifiesto)
		})
	}
}

// falloDeJSON es cómo no es JSON con la forma del manifiesto un documento
// ilegible por el documento: por la gramática o por el significado.
type falloDeJSON int

const (
	// sintactico es el documento que no respeta la gramática de JSON, lleva
	// datos tras el primer valor o repite un nombre en un objeto.
	sintactico falloDeJSON = iota + 1
	// semantico es el JSON que no tiene la forma de Go del manifiesto: un
	// miembro desconocido o un valor de otro tipo.
	semantico
)

// probarIlegiblesPorElDocumento lee documentos que no son JSON con la forma
// del manifiesto (contracts/manifiesto.md §3, regla 3) y exige de cada uno un
// manifiesto ilegible que envuelve el error de encoding/json/v2 que lo
// rechaza, con su causa y, en los de significado, el valor donde falla.
func probarIlegiblesPorElDocumento(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		contenido string
		fallo     falloDeJSON
		// causa, si no es nil, es el error centinela que tiene que envolver.
		causa error
		// puntero, si no está vacío, es el valor JSON donde falla uno de
		// significado.
		puntero string
		// motivo, si no está vacío, es un fragmento del mensaje.
		motivo string
	}{
		// La gramática.
		{nombre: "no es JSON", contenido: "esto no es JSON", fallo: sintactico},
		{nombre: "vacío", contenido: "", fallo: sintactico},
		{nombre: "solo blancos", contenido: " \n\t", fallo: sintactico},
		{nombre: "objeto sin cerrar", contenido: `{"skills": {}, "version": "v0.1.0"`, fallo: sintactico},
		{nombre: "coma final", contenido: `{"skills": {}, "version": "v0.1.0",}`, fallo: sintactico},
		{nombre: "comentario", contenido: conVersion(`"v0.1.0"`) + " // nota", fallo: sintactico},
		{nombre: "comillas simples", contenido: `{'skills': {}, 'version': 'v0.1.0'}`, fallo: sintactico},
		{nombre: "UTF-8 inválido", contenido: conVersion("\"v0.1.0\xff\""), fallo: sintactico},

		// Datos tras el documento.
		{nombre: "un objeto tras el documento", contenido: conVersion(`"v0.1.0"`) + "\n{}", fallo: sintactico},
		{nombre: "el documento dos veces", contenido: conVersion(`"v0.1.0"`) + conVersion(`"v0.1.0"`), fallo: sintactico},

		// Un nombre repetido, en cada objeto del manifiesto.
		{
			nombre:    "version repetida",
			contenido: `{"skills": {}, "version": "v0.1.0", "version": "v0.1.0"}`,
			fallo:     sintactico, causa: jsontext.ErrDuplicateName,
		},
		{
			nombre:    "version repetida tras un escape que la esconde",
			contenido: `{"skills": {}, "version": "v0.1.0", "version": "v0.2.0"}`,
			fallo:     sintactico, causa: jsontext.ErrDuplicateName,
		},
		{
			nombre:    "skill repetida",
			contenido: conSkills(`{"legal-core": ` + entradaDeLegalCore + `, "legal-core": ` + entradaDeLegalCore + `}`),
			fallo:     sintactico, causa: jsontext.ErrDuplicateName,
		},
		{
			nombre:    "miembro repetido en la entrada de una skill",
			contenido: conLegalCore(`{"ficheros": {}, "version": "v0.1.0", "version": "v0.2.0"}`),
			fallo:     sintactico, causa: jsontext.ErrDuplicateName,
		},
		{
			nombre: "fichero repetido",
			contenido: conFicheros(`{"legal-core/SKILL.md": "` + huellaDeC + `", "legal-core/SKILL.md": "` +
				huellaDeA + `"}`),
			fallo: sintactico, causa: jsontext.ErrDuplicateName,
		},
		{
			nombre:    "host repetido",
			contenido: conHosts(`{"claude": ` + enlaceDeLegalCore + `, "claude": ` + enlaceDeLegalCore + `}`),
			fallo:     sintactico, causa: jsontext.ErrDuplicateName,
		},
		{
			nombre:    "miembro repetido en la entrada de host",
			contenido: conClaude(`{"modo": "enlace", "modo": "copia", "ruta": ".claude/skills/legal-core"}`),
			fallo:     sintactico, causa: jsontext.ErrDuplicateName,
		},

		// Un miembro desconocido, en cada objeto de campos fijos: ni fechas,
		// ni usuarios, ni máquinas, ni rutas absolutas (FR-032).
		{
			nombre:    "miembro desconocido arriba",
			contenido: `{"fecha": "2026-09-26", "skills": {}, "version": "v0.1.0"}`,
			fallo:     semantico, causa: json.ErrUnknownName,
		},
		{
			nombre:    "miembro desconocido en la entrada de una skill",
			contenido: conLegalCore(`{"ficheros": {}, "usuario": "ana", "version": "v0.1.0"}`),
			fallo:     semantico, causa: json.ErrUnknownName,
		},
		{
			nombre: "miembro desconocido en la entrada de host",
			contenido: conClaude(`{"maquina": "portatil", "modo": "enlace", ` +
				`"ruta": ".claude/skills/legal-core"}`),
			fallo: semantico, causa: json.ErrUnknownName,
		},
		{
			nombre:    "una clave conocida con otras mayúsculas",
			contenido: `{"Skills": {}, "version": "v0.1.0"}`,
			fallo:     semantico, causa: json.ErrUnknownName,
		},

		// Un valor de otro tipo.
		{nombre: "el documento es una lista", contenido: `[]`, fallo: semantico},
		{nombre: "el documento es una cadena", contenido: `"v0.1.0"`, fallo: semantico},
		{nombre: "el documento es un número", contenido: `1`, fallo: semantico},
		{nombre: "version que no es una cadena", contenido: conVersion(`1`), fallo: semantico, puntero: "/version"},
		{
			nombre:    "skills null",
			contenido: `{"skills": null, "version": "v0.1.0"}`,
			fallo:     semantico, puntero: "/skills", motivo: "no es un objeto JSON",
		},
		{
			nombre:    "skills que es una lista",
			contenido: `{"skills": [], "version": "v0.1.0"}`,
			fallo:     semantico, puntero: "/skills", motivo: "no es un objeto JSON",
		},
		{
			nombre:    "entrada de skill que es una lista",
			contenido: conSkills(`{"legal-core": []}`),
			fallo:     semantico, puntero: "/skills/legal-core",
		},
		{
			nombre:    "ficheros null",
			contenido: conFicheros(`null`),
			fallo:     semantico, puntero: "/skills/legal-core/ficheros", motivo: "no es un objeto JSON",
		},
		{
			nombre:    "ficheros que es una lista",
			contenido: conFicheros(`["legal-core/SKILL.md"]`),
			fallo:     semantico, puntero: "/skills/legal-core/ficheros", motivo: "no es un objeto JSON",
		},
		{nombre: "huella que no es una cadena", contenido: conFicheros(`{"legal-core/SKILL.md": 1}`), fallo: semantico},
		{
			nombre:    "hosts null",
			contenido: conHosts(`null`),
			fallo:     semantico, puntero: "/skills/legal-core/hosts", motivo: "no es un objeto JSON",
		},
		{
			nombre:    "ficheros de la copia null",
			contenido: conCopia(`null`),
			fallo:     semantico, puntero: "/skills/legal-core/hosts/claude/ficheros", motivo: "no es un objeto JSON",
		},
		{
			nombre:    "modo que no es una cadena",
			contenido: conClaude(`{"modo": true, "ruta": ".claude/skills/legal-core"}`),
			fallo:     semantico, puntero: "/skills/legal-core/hosts/claude/modo",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			err := compruebaIlegible(t, caso.contenido, "no es un documento JSON con la forma del manifiesto")

			if caso.causa != nil {
				require.ErrorIs(t, err, caso.causa)
			}

			if caso.motivo != "" {
				require.ErrorContains(t, err, caso.motivo)
			}

			switch caso.fallo {
			case sintactico:
				var sintaxis *jsontext.SyntacticError
				require.ErrorAs(t, err, &sintaxis, "el rechazo no es de la gramática de JSON")
			case semantico:
				var significado *json.SemanticError
				require.ErrorAs(t, err, &significado, "el rechazo no es del significado del JSON")

				if caso.puntero != "" {
					assert.Equal(t, caso.puntero, string(significado.JSONPointer), "el valor donde falla")
				}
			}
		})
	}
}

// probarIlegiblesPorLaForma lee documentos JSON con la forma de Go del
// manifiesto que incumplen una regla de la tabla de contracts/manifiesto.md §1
// (§3, regla 4) y exige de cada uno un manifiesto ilegible que nombra esa
// regla y que no es un error de encoding/json/v2: el documento se leyó.
func probarIlegiblesPorLaForma(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre    string
		contenido string
		// motivo es el fragmento del mensaje que nombra la regla incumplida.
		motivo string
	}{
		// version, arriba.
		{nombre: "sin version", contenido: `{"skills": {}}`, motivo: "version: está vacía"},
		{nombre: "version vacía", contenido: conVersion(`""`), motivo: "version: está vacía"},
		{nombre: "version null", contenido: conVersion(`null`), motivo: "version: está vacía"},
		{nombre: "el documento es null", contenido: `null`, motivo: "version: está vacía"},
		{nombre: "version con un salto de línea", contenido: conVersion(`"v0.1.0\n"`), motivo: "lleva un carácter de control"},
		{nombre: "version con un tabulador", contenido: conVersion(`"v0.1.0\tx"`), motivo: "lleva un carácter de control"},
		{nombre: "version con NUL", contenido: conVersion(`"v0.1.0\u0000"`), motivo: "lleva un carácter de control"},
		{nombre: "version con DEL", contenido: conVersion("\"v0.1.0\x7f\""), motivo: "lleva un carácter de control"},
		{nombre: "version con un control C1", contenido: conVersion(`"v0.1.0\u0085"`), motivo: "lleva un carácter de control"},

		// skills.
		{nombre: "sin skills", contenido: `{"version": "v0.1.0"}`, motivo: "falta skills"},

		// El nombre de una skill.
		{nombre: "nombre vacío", contenido: conNombre(""), motivo: "no tiene la forma de un nombre de skill"},
		{nombre: "nombre con mayúsculas", contenido: conNombre("Legal-core"), motivo: "no tiene la forma de un nombre de skill"},
		{nombre: "nombre con guion bajo", contenido: conNombre("legal_core"), motivo: "no tiene la forma de un nombre de skill"},
		{nombre: "nombre con punto", contenido: conNombre("legal.core"), motivo: "no tiene la forma de un nombre de skill"},
		{nombre: "nombre con barra", contenido: conNombre("legal/core"), motivo: "no tiene la forma de un nombre de skill"},
		{nombre: "nombre con blanco", contenido: conNombre("legal core"), motivo: "no tiene la forma de un nombre de skill"},
		{nombre: "nombre ..", contenido: conNombre(".."), motivo: "no tiene la forma de un nombre de skill"},
		{nombre: "nombre fuera de ASCII", contenido: conNombre("legislación"), motivo: "no tiene la forma de un nombre de skill"},
		{nombre: "guion al principio", contenido: conNombre("-legal"), motivo: "no tiene la forma de un nombre de skill"},
		{nombre: "guion al final", contenido: conNombre("legal-"), motivo: "no tiene la forma de un nombre de skill"},
		{nombre: "dos guiones seguidos", contenido: conNombre("legal--core"), motivo: "no tiene la forma de un nombre de skill"},
		{
			nombre:    "nombre de 65 caracteres",
			contenido: conNombre(strings.Repeat("a", 65)),
			motivo:    "no tiene la forma de un nombre de skill",
		},

		// La version de una skill.
		{
			nombre:    "skill sin version",
			contenido: conLegalCore(`{"ficheros": ` + ficherosDeLegalCore + `}`),
			motivo:    `skill "legal-core": version: está vacía`,
		},
		{
			nombre:    "skill con un control en la version",
			contenido: conLegalCore(`{"ficheros": ` + ficherosDeLegalCore + `, "version": "v0.1.0\r"}`),
			motivo:    `skill "legal-core": version: "v0.1.0\r" lleva un carácter de control`,
		},
		{
			nombre:    "entrada de skill null",
			contenido: conSkills(`{"legal-core": null}`),
			motivo:    `skill "legal-core": version: está vacía`,
		},

		// Los ficheros de una skill: rutas relativas, limpias, con / y bajo
		// el directorio de la skill.
		{
			nombre:    "sin ficheros",
			contenido: conLegalCore(`{"version": "v0.1.0"}`),
			motivo:    `skill "legal-core": falta ficheros`,
		},
		{
			nombre:    "ruta vacía",
			contenido: conFichero("", huellaDeC),
			motivo:    `skill "legal-core": ficheros: una ruta está vacía`,
		},
		{
			nombre:    "ruta absoluta",
			contenido: conFichero("/legal-core/SKILL.md", huellaDeC),
			motivo:    `skill "legal-core": ficheros: la ruta "/legal-core/SKILL.md" es absoluta`,
		},
		{
			nombre:    "ruta absoluta con la raíz de la instalación",
			contenido: conFichero("/home/ana/proyecto/.agents/skills/legal-core/SKILL.md", huellaDeC),
			motivo:    "es absoluta",
		},
		{
			nombre:    "ruta con barras invertidas",
			contenido: conFichero(`legal-core\\references\\normas.md`, huellaDeC),
			motivo:    `skill "legal-core": ficheros: la ruta "legal-core\\references\\normas.md" lleva una barra invertida`,
		},
		{
			nombre:    "ruta que sube con barras invertidas",
			contenido: conFichero(`legal-core/references\\..\\..\\fuera.md`, huellaDeC),
			motivo:    "lleva una barra invertida",
		},
		{nombre: "componente vacío", contenido: conFichero("legal-core//SKILL.md", huellaDeC), motivo: "no está limpia"},
		{nombre: "componente .", contenido: conFichero("legal-core/./SKILL.md", huellaDeC), motivo: "no está limpia"},
		{nombre: "componente ..", contenido: conFichero("legal-core/../SKILL.md", huellaDeC), motivo: "no está limpia"},
		{
			nombre:    "ruta que sale de la skill",
			contenido: conFichero("legal-core/references/../../boe-legislacion/SKILL.md", huellaDeC),
			motivo:    "no está limpia",
		},
		{nombre: "ruta que termina en ..", contenido: conFichero("legal-core/..", huellaDeC), motivo: "no está limpia"},
		{nombre: "ruta que termina en .", contenido: conFichero("legal-core/.", huellaDeC), motivo: "no está limpia"},
		{nombre: "barra final", contenido: conFichero("legal-core/references/", huellaDeC), motivo: "no está limpia"},
		{nombre: "empieza por ..", contenido: conFichero("../legal-core/SKILL.md", huellaDeC), motivo: "no está limpia"},
		{nombre: "empieza por .", contenido: conFichero("./legal-core/SKILL.md", huellaDeC), motivo: "no está limpia"},
		{
			nombre:    "ruta de otra skill",
			contenido: conFichero("boe-legislacion/SKILL.md", huellaDeC),
			motivo:    `skill "legal-core": ficheros: la ruta "boe-legislacion/SKILL.md" no empieza por "legal-core/"`,
		},
		{
			nombre:    "ruta de una skill cuyo nombre empieza igual",
			contenido: conFichero("legal-core-extra/SKILL.md", huellaDeC),
			motivo:    `no empieza por "legal-core/"`,
		},
		{nombre: "ruta que es la skill", contenido: conFichero("legal-core", huellaDeC), motivo: `no empieza por "legal-core/"`},
		{nombre: "ruta sin la skill", contenido: conFichero("SKILL.md", huellaDeC), motivo: `no empieza por "legal-core/"`},

		// Las huellas.
		{
			nombre:    "huella vacía",
			contenido: conFichero("legal-core/SKILL.md", ""),
			motivo:    `skill "legal-core": ficheros: la huella de "legal-core/SKILL.md" no es sha256: seguido de`,
		},
		{
			nombre:    "huella null",
			contenido: conFicheros(`{"legal-core/SKILL.md": null}`),
			motivo:    "no es sha256: seguido de",
		},
		{
			nombre:    "huella sin prefijo",
			contenido: conFichero("legal-core/SKILL.md", strings.TrimPrefix(huellaDeC, "sha256:")),
			motivo:    "no es sha256: seguido de",
		},
		{
			nombre:    "prefijo en mayúsculas",
			contenido: conFichero("legal-core/SKILL.md", "SHA256:"+strings.TrimPrefix(huellaDeC, "sha256:")),
			motivo:    "no es sha256: seguido de",
		},
		{
			nombre:    "otro algoritmo",
			contenido: conFichero("legal-core/SKILL.md", "md5:"+strings.TrimPrefix(huellaDeC, "sha256:")),
			motivo:    "no es sha256: seguido de",
		},
		{nombre: "huella corta", contenido: conFichero("legal-core/SKILL.md", "sha256:ABCDEF"), motivo: "no es sha256: seguido de"},
		{
			nombre:    "hexadecimales en mayúsculas",
			contenido: conFichero("legal-core/SKILL.md", "sha256:"+strings.ToUpper(huellaDeC[len("sha256:"):])),
			motivo:    "no es sha256: seguido de",
		},
		{
			nombre:    "63 hexadecimales",
			contenido: conFichero("legal-core/SKILL.md", huellaDeC[:len(huellaDeC)-1]),
			motivo:    "no es sha256: seguido de",
		},
		{
			nombre:    "65 hexadecimales",
			contenido: conFichero("legal-core/SKILL.md", huellaDeC+"0"),
			motivo:    "no es sha256: seguido de",
		},
		{
			nombre:    "un carácter que no es hexadecimal",
			contenido: conFichero("legal-core/SKILL.md", huellaDeC[:len(huellaDeC)-1]+"g"),
			motivo:    "no es sha256: seguido de",
		},
		{
			nombre:    "un blanco tras el prefijo",
			contenido: conFichero("legal-core/SKILL.md", "sha256: "+huellaDeC[len("sha256:"):]),
			motivo:    "no es sha256: seguido de",
		},
		{
			nombre:    "un salto de línea tras la huella",
			contenido: conFichero("legal-core/SKILL.md", huellaDeC+`\n`),
			motivo:    "no es sha256: seguido de",
		},

		// hosts: se omite sin entradas y su única clave es claude.
		{nombre: "hosts sin entradas", contenido: conHosts(`{}`), motivo: `skill "legal-core": hosts no tiene ninguna entrada`},
		{
			nombre:    "otro host",
			contenido: conHosts(`{"codex": ` + enlaceDeLegalCore + `}`),
			motivo:    `skill "legal-core": hosts: el host "codex" no se admite`,
		},
		{
			nombre:    "claude con mayúscula",
			contenido: conHosts(`{"Claude": ` + enlaceDeLegalCore + `}`),
			motivo:    `el host "Claude" no se admite`,
		},
		{
			nombre:    "claude y otro host",
			contenido: conHosts(`{"antigravity": ` + enlaceDeLegalCore + `, "claude": ` + enlaceDeLegalCore + `}`),
			motivo:    `el host "antigravity" no se admite`,
		},

		// La entrada de host claude: su ruta exacta y su modo.
		{
			nombre:    "entrada de host null",
			contenido: conClaude(`null`),
			motivo:    `skill "legal-core": hosts: claude: la ruta es "" y tiene que ser ".claude/skills/legal-core"`,
		},
		{nombre: "sin ruta", contenido: conClaude(`{"modo": "enlace"}`), motivo: `hosts: claude: la ruta es ""`},
		{
			nombre:    "ruta absoluta con la raíz del ámbito",
			contenido: conClaude(`{"modo": "enlace", "ruta": "/home/ana/proyecto/.claude/skills/legal-core"}`),
			motivo:    `hosts: claude: la ruta es "/home/ana/proyecto/.claude/skills/legal-core"`,
		},
		{
			nombre:    "ruta de otra skill",
			contenido: conClaude(`{"modo": "enlace", "ruta": ".claude/skills/boe-legislacion"}`),
			motivo:    `hosts: claude: la ruta es ".claude/skills/boe-legislacion"`,
		},
		{
			nombre:    "ruta con barra final",
			contenido: conClaude(`{"modo": "enlace", "ruta": ".claude/skills/legal-core/"}`),
			motivo:    `hosts: claude: la ruta es ".claude/skills/legal-core/"`,
		},
		{
			nombre:    "ruta que empieza por ./",
			contenido: conClaude(`{"modo": "enlace", "ruta": "./.claude/skills/legal-core"}`),
			motivo:    `hosts: claude: la ruta es "./.claude/skills/legal-core"`,
		},
		{
			nombre:    "ruta del directorio neutro",
			contenido: conClaude(`{"modo": "enlace", "ruta": ".agents/skills/legal-core"}`),
			motivo:    `hosts: claude: la ruta es ".agents/skills/legal-core"`,
		},
		{
			nombre:    "sin modo",
			contenido: conClaude(`{"ruta": ".claude/skills/legal-core"}`),
			motivo:    `skill "legal-core": hosts: claude: el modo es "" y tiene que ser "enlace" o "copia"`,
		},
		{
			nombre:    "modo con mayúscula",
			contenido: conClaude(`{"modo": "Enlace", "ruta": ".claude/skills/legal-core"}`),
			motivo:    `hosts: claude: el modo es "Enlace"`,
		},
		{
			nombre:    "modo desconocido",
			contenido: conClaude(`{"modo": "symlink", "ruta": ".claude/skills/legal-core"}`),
			motivo:    `hosts: claude: el modo es "symlink"`,
		},

		// Los ficheros de una entrada de host: solo y obligatorios en copia.
		{
			nombre:    "enlace con ficheros vacíos",
			contenido: conClaude(`{"ficheros": {}, "modo": "enlace", "ruta": ".claude/skills/legal-core"}`),
			motivo:    `skill "legal-core": hosts: claude: ficheros solo va en modo copia`,
		},
		{
			nombre: "enlace con ficheros",
			contenido: conClaude(`{"ficheros": {".claude/skills/legal-core/SKILL.md": "` + huellaDeC +
				`"}, "modo": "enlace", "ruta": ".claude/skills/legal-core"}`),
			motivo: "ficheros solo va en modo copia",
		},
		{
			nombre:    "copia sin ficheros",
			contenido: conClaude(`{"modo": "copia", "ruta": ".claude/skills/legal-core"}`),
			motivo:    `skill "legal-core": hosts: claude: falta ficheros, obligatorio en modo copia`,
		},
		{
			nombre:    "fichero de la copia con la ruta del directorio neutro",
			contenido: conFicheroDeLaCopia("legal-core/SKILL.md", huellaDeC),
			motivo: `skill "legal-core": hosts: claude: ficheros: la ruta "legal-core/SKILL.md" no empieza por ` +
				`".claude/skills/legal-core/"`,
		},
		{
			nombre:    "fichero de la copia de otra skill",
			contenido: conFicheroDeLaCopia(".claude/skills/boe-legislacion/SKILL.md", huellaDeC),
			motivo:    `no empieza por ".claude/skills/legal-core/"`,
		},
		{
			nombre:    "fichero de la copia que es la copia",
			contenido: conFicheroDeLaCopia(".claude/skills/legal-core", huellaDeC),
			motivo:    `no empieza por ".claude/skills/legal-core/"`,
		},
		{
			nombre:    "fichero de la copia con ruta absoluta",
			contenido: conFicheroDeLaCopia("/home/ana/.claude/skills/legal-core/SKILL.md", huellaDeC),
			motivo:    `hosts: claude: ficheros: la ruta "/home/ana/.claude/skills/legal-core/SKILL.md" es absoluta`,
		},
		{
			nombre:    "fichero de la copia que sale de ella",
			contenido: conFicheroDeLaCopia(".claude/skills/legal-core/../boe-legislacion/SKILL.md", huellaDeC),
			motivo:    "hosts: claude: ficheros: la ruta",
		},
		{
			nombre:    "fichero de la copia con barra invertida",
			contenido: conFicheroDeLaCopia(`.claude/skills/legal-core/references\\normas.md`, huellaDeC),
			motivo:    "lleva una barra invertida",
		},
		{
			nombre:    "huella de la copia que no es sha256",
			contenido: conFicheroDeLaCopia(".claude/skills/legal-core/SKILL.md", "sha256:ABCDEF"),
			motivo:    `hosts: claude: ficheros: la huella de ".claude/skills/legal-core/SKILL.md" no es sha256:`,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			err := compruebaIlegible(t, caso.contenido, caso.motivo)

			var sintaxis *jsontext.SyntacticError
			assert.NotErrorAs(t, err, &sintaxis, "el documento es JSON válido y se tuvo que leer")

			var significado *json.SemanticError
			assert.NotErrorAs(t, err, &significado, "el documento tiene la forma de Go y se tuvo que leer")
		})
	}
}

// compruebaIlegible lee contenido y exige un manifiesto ilegible: el error
// tiene el tipo instalacion.ManifiestoIlegible —el que reconocen install,
// list, doctor y el aviso—, su mensaje dice que el manifiesto es ilegible y
// contiene motivo, y no sale ningún manifiesto. Devuelve el error.
func compruebaIlegible(t *testing.T, contenido, motivo string) error {
	t.Helper()

	manifiesto, err := instalacion.LeerManifiesto([]byte(contenido))
	require.Error(t, err, "se lee %q", contenido)

	var ilegible *instalacion.ManifiestoIlegible
	require.ErrorAs(t, err, &ilegible, "el rechazo de %q no es un manifiesto ilegible", contenido)
	require.ErrorContains(t, err, motivo, "el mensaje no nombra la regla incumplida")
	assert.Zero(t, manifiesto, "un manifiesto ilegible no da ningún manifiesto")
	assert.True(t, strings.HasPrefix(err.Error(), "manifiesto ilegible: "), "el mensaje %q", err.Error())

	return err
}

// compruebaBytesCanonicos exige que escrito sea, byte a byte, esperado. No es
// una comparación de JSON: la forma canónica es el contrato
// (contracts/manifiesto.md §2), y otro sangrado u otro orden de las claves,
// que una comparación de JSON daría por buenos, son otro manifiesto en disco
// y un cambio en el proyecto que lo versiona (FR-032, FR-033).
func compruebaBytesCanonicos(t *testing.T, esperado string, escrito []byte, mensaje ...any) {
	t.Helper()

	assert.Equal(t, esperado, string(escrito), mensaje...)
}

// TestManifiestoCanonico fija la escritura de kitlegal.json
// (contracts/manifiesto.md §2; FR-032, FR-033): los bytes son función del
// contenido y de nada más —ni del orden en que se llenaron los mapas, ni de la
// ruta o la máquina de la instalación—, en la forma canónica; releer y
// reescribir no cambia un byte; y lo que no respeta la forma, que no se
// podría volver a leer, no se escribe.
func TestManifiestoCanonico(t *testing.T) {
	t.Parallel()

	t.Run("la forma canónica", probarFormaCanonica)
	t.Run("dos instalaciones iguales en rutas distintas", probarDosInstalacionesIguales)
	t.Run("releer y reescribir no cambia un byte", probarReleerYReescribir)
	t.Run("objetos sin entradas", probarObjetosSinEntradas)
	t.Run("lo que no respeta la forma no se escribe", probarLoQueNoSeEscribe)
}

// probarFormaCanonica escribe el manifiesto de referencia y exige
// canonicoDeReferencia byte a byte, y que un carácter de HTML no se escape.
func probarFormaCanonica(t *testing.T) {
	t.Parallel()

	contenido, err := manifiestoDeReferencia().Bytes()
	require.NoError(t, err)
	compruebaBytesCanonicos(t, canonicoDeReferencia, contenido)

	html := instalacion.Manifiesto{Version: "dev<&>", Skills: map[string]instalacion.SkillDeclarada{}}
	contenido, err = html.Bytes()
	require.NoError(t, err)
	compruebaBytesCanonicos(t, "{\n  \"skills\": {},\n  \"version\": \"dev<&>\"\n}\n", contenido)
}

// probarDosInstalacionesIguales construye el manifiesto de una misma
// instalación como lo harían dos máquinas, cada una con sus mapas llenados en
// un orden distinto, muchas veces, y exige siempre los mismos bytes. La ruta
// de la instalación no puede cambiarlos porque no puede entrar en el
// manifiesto: lo que la lleve no se escribe (probarLoQueNoSeEscribe).
func probarDosInstalacionesIguales(t *testing.T) {
	t.Parallel()

	for vuelta := range 50 {
		una, err := manifiestoLlenadoEnOrden(false).Bytes()
		require.NoError(t, err)

		otra, err := manifiestoLlenadoEnOrden(true).Bytes()
		require.NoError(t, err)

		compruebaBytesCanonicos(t, canonicoDeReferencia, una, "vuelta %d", vuelta)
		compruebaBytesCanonicos(t, string(una), otra, "vuelta %d", vuelta)
	}
}

// manifiestoLlenadoEnOrden es el manifiesto de referencia con cada mapa
// llenado entrada a entrada en el orden de manifiestoDeReferencia o, si
// inverso, en el contrario.
func manifiestoLlenadoEnOrden(inverso bool) instalacion.Manifiesto {
	referencia := manifiestoDeReferencia()

	enOrden := func(claves []string) []string {
		if inverso {
			slices.Reverse(claves)
		}

		return claves
	}

	llenar := func(origen map[string]string) map[string]string {
		destino := make(map[string]string, len(origen))
		for _, clave := range enOrden(slices.Sorted(maps.Keys(origen))) {
			destino[clave] = origen[clave]
		}

		return destino
	}

	manifiesto := instalacion.Manifiesto{Version: referencia.Version, Skills: map[string]instalacion.SkillDeclarada{}}
	for _, nombre := range enOrden(slices.Sorted(maps.Keys(referencia.Skills))) {
		skill := referencia.Skills[nombre]
		skill.Ficheros = llenar(skill.Ficheros)

		if skill.Claude != nil {
			host := *skill.Claude
			if host.Ficheros != nil {
				host.Ficheros = llenar(host.Ficheros)
			}

			skill.Claude = &host
		}

		manifiesto.Skills[nombre] = skill
	}

	return manifiesto
}

// probarReleerYReescribir lee lo escrito y lo vuelve a escribir varias veces
// sin que cambie un byte ni el manifiesto; y un documento legible que no
// está en la forma canónica se escribe en ella a la primera y ya no cambia.
func probarReleerYReescribir(t *testing.T) {
	t.Parallel()

	escrito, err := manifiestoDeReferencia().Bytes()
	require.NoError(t, err)

	for vuelta := range 3 {
		leido, err := instalacion.LeerManifiesto(escrito)
		require.NoError(t, err, "vuelta %d", vuelta)
		require.Equal(t, manifiestoDeReferencia(), leido, "vuelta %d", vuelta)

		reescrito, err := leido.Bytes()
		require.NoError(t, err, "vuelta %d", vuelta)
		compruebaBytesCanonicos(t, string(escrito), reescrito, "vuelta %d", vuelta)
	}

	for _, legible := range []string{compactoDeReferencia, "\n" + canonicoDeReferencia + "\n\t"} {
		leido, err := instalacion.LeerManifiesto([]byte(legible))
		require.NoError(t, err)

		escrito, err := leido.Bytes()
		require.NoError(t, err)
		compruebaBytesCanonicos(t, canonicoDeReferencia, escrito)
	}
}

// probarObjetosSinEntradas escribe manifiestos con objetos sin entradas y
// exige la forma del contrato: skills y los ficheros de una skill van
// siempre, aunque estén vacíos; los ficheros de una entrada de host van en
// copia aunque estén vacíos y faltan en enlace; y lo escrito se vuelve a leer.
func probarObjetosSinEntradas(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre     string
		manifiesto instalacion.Manifiesto
		escrito    string
	}{
		{
			nombre:     "sin skills",
			manifiesto: instalacion.Manifiesto{Version: "v0.1.0"},
			escrito:    "{\n  \"skills\": {},\n  \"version\": \"v0.1.0\"\n}\n",
		},
		{
			nombre: "una skill sin ficheros y con una copia sin ficheros",
			manifiesto: instalacion.Manifiesto{Version: "v0.1.0", Skills: map[string]instalacion.SkillDeclarada{
				"legal-core": {
					Version: "v0.1.0",
					Claude:  &instalacion.EntradaDeHost{Ruta: ".claude/skills/legal-core", Modo: instalacion.ModoCopia},
				},
			}},
			escrito: "{\n  \"skills\": {\n    \"legal-core\": {\n      \"ficheros\": {},\n      \"hosts\": {\n" +
				"        \"claude\": {\n          \"ficheros\": {},\n          \"modo\": \"copia\",\n" +
				"          \"ruta\": \".claude/skills/legal-core\"\n        }\n      },\n" +
				"      \"version\": \"v0.1.0\"\n    }\n  },\n  \"version\": \"v0.1.0\"\n}\n",
		},
		{
			nombre: "un enlace con un mapa de ficheros vacío",
			manifiesto: instalacion.Manifiesto{Version: "v0.1.0", Skills: map[string]instalacion.SkillDeclarada{
				"legal-core": {
					Version:  "v0.1.0",
					Ficheros: map[string]string{},
					Claude: &instalacion.EntradaDeHost{
						Ruta: ".claude/skills/legal-core", Modo: instalacion.ModoEnlace, Ficheros: map[string]string{},
					},
				},
			}},
			escrito: "{\n  \"skills\": {\n    \"legal-core\": {\n      \"ficheros\": {},\n      \"hosts\": {\n" +
				"        \"claude\": {\n          \"modo\": \"enlace\",\n" +
				"          \"ruta\": \".claude/skills/legal-core\"\n        }\n      },\n" +
				"      \"version\": \"v0.1.0\"\n    }\n  },\n  \"version\": \"v0.1.0\"\n}\n",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			escrito, err := caso.manifiesto.Bytes()
			require.NoError(t, err)
			compruebaBytesCanonicos(t, caso.escrito, escrito)

			leido, err := instalacion.LeerManifiesto(escrito)
			require.NoError(t, err)

			reescrito, err := leido.Bytes()
			require.NoError(t, err)
			compruebaBytesCanonicos(t, caso.escrito, reescrito)
		})
	}
}

// probarLoQueNoSeEscribe exige que Bytes no escriba nada de lo que
// LeerManifiesto no podría volver a leer: cada manifiesto de la tabla cambia
// una sola cosa respecto del de referencia y sale con un error, sin bytes. Las
// rutas absolutas son las de dos instalaciones en raíces distintas, la local
// de un proyecto y la global de un HOME: ninguna llega al manifiesto (FR-032).
// El error no es un manifiesto ilegible: es el escritor el que se niega.
func probarLoQueNoSeEscribe(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre  string
		cambiar func(manifiesto *instalacion.Manifiesto)
	}{
		{nombre: "version vacía", cambiar: func(m *instalacion.Manifiesto) { m.Version = "" }},
		{nombre: "version con un control", cambiar: func(m *instalacion.Manifiesto) { m.Version = "v0.1.0\n" }},
		{nombre: "version con UTF-8 inválido", cambiar: func(m *instalacion.Manifiesto) { m.Version = "v0.1.0\xff" }},
		{
			nombre: "nombre de skill sin forma",
			cambiar: func(m *instalacion.Manifiesto) {
				m.Skills["Legal-core"] = m.Skills["otra-skill"]
			},
		},
		{
			nombre: "skill sin version",
			cambiar: func(m *instalacion.Manifiesto) {
				skill := m.Skills["otra-skill"]
				skill.Version = ""
				m.Skills["otra-skill"] = skill
			},
		},
		{
			nombre: "fichero con la ruta absoluta de un proyecto",
			cambiar: func(m *instalacion.Manifiesto) {
				m.Skills["otra-skill"].Ficheros["/home/ana/proyecto/.agents/skills/otra-skill/SKILL.md"] = huellaDeA
			},
		},
		{
			nombre: "fichero con la ruta absoluta de un HOME",
			cambiar: func(m *instalacion.Manifiesto) {
				m.Skills["otra-skill"].Ficheros["/Users/luis/.agents/skills/otra-skill/SKILL.md"] = huellaDeA
			},
		},
		{
			nombre: "fichero de otra skill",
			cambiar: func(m *instalacion.Manifiesto) {
				m.Skills["otra-skill"].Ficheros["legal-core/SKILL.md"] = huellaDeA
			},
		},
		{
			nombre: "fichero con UTF-8 inválido en la ruta",
			cambiar: func(m *instalacion.Manifiesto) {
				m.Skills["otra-skill"].Ficheros["otra-skill/\xff.md"] = huellaDeA
			},
		},
		{
			nombre: "huella sin forma",
			cambiar: func(m *instalacion.Manifiesto) {
				m.Skills["otra-skill"].Ficheros["otra-skill/SKILL.md"] = "sha256:ABCDEF"
			},
		},
		{
			nombre: "host con la ruta absoluta de un proyecto",
			cambiar: func(m *instalacion.Manifiesto) {
				m.Skills["boe-legislacion"].Claude.Ruta = "/home/ana/proyecto/.claude/skills/boe-legislacion"
			},
		},
		{
			nombre: "host con la ruta absoluta de un HOME",
			cambiar: func(m *instalacion.Manifiesto) {
				m.Skills["boe-legislacion"].Claude.Ruta = "/Users/luis/.claude/skills/boe-legislacion"
			},
		},
		{
			nombre: "fichero de la copia con la ruta absoluta de un HOME",
			cambiar: func(m *instalacion.Manifiesto) {
				m.Skills["legal-core"].Claude.Ficheros["/Users/luis/.claude/skills/legal-core/SKILL.md"] = huellaDeC
			},
		},
		{nombre: "host sin modo", cambiar: func(m *instalacion.Manifiesto) { m.Skills["boe-legislacion"].Claude.Modo = "" }},
		{
			nombre:  "host con otro modo",
			cambiar: func(m *instalacion.Manifiesto) { m.Skills["boe-legislacion"].Claude.Modo = "symlink" },
		},
		{
			nombre: "enlace con ficheros",
			cambiar: func(m *instalacion.Manifiesto) {
				m.Skills["boe-legislacion"].Claude.Ficheros = map[string]string{
					".claude/skills/boe-legislacion/SKILL.md": huellaDeA,
				}
			},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			manifiesto := manifiestoDeReferencia()
			caso.cambiar(&manifiesto)

			escrito, err := manifiesto.Bytes()
			require.ErrorContains(t, err, "manifiesto")
			assert.Nil(t, escrito, "lo que no se puede escribir no da bytes")

			var ilegible *instalacion.ManifiestoIlegible
			assert.NotErrorAs(t, err, &ilegible, "escribir no es leer: el error no es un manifiesto ilegible")
		})
	}
}

// FuzzLeerManifiesto lee cualquier contenido y exige: ningún panic; o un
// manifiesto ilegible con su tipo y sin manifiesto, o un manifiesto que se
// escribe sin error, cuya forma canónica se vuelve a leer al mismo manifiesto
// y se reescribe sin cambiar un byte (FR-035, FR-144; contracts/manifiesto.md
// §2 y §3). El corpus va aquí, sin ficheros de corpus.
func FuzzLeerManifiesto(f *testing.F) {
	semillas := []string{
		canonicoDeReferencia,
		compactoDeReferencia,
		`{"skills": {}, "version": "dev"}`,
		conCopia(`{}`),
		conHosts(`{}`),
		conFichero("legal-core/../SKILL.md", huellaDeC),
		conFichero("legal-core/SKILL.md", "sha256:ABCDEF"),
		`{"skills": {}, "version": "v0.1.0", "version": "v0.1.0"}`,
		conVersion(`"v0.1.0"`) + "\n{}",
		conVersion("\"v0.1.0\xff\""),
		`null`,
		`[]`,
		"esto no es JSON",
		"",
	}

	for _, semilla := range semillas {
		f.Add([]byte(semilla))
	}

	f.Fuzz(func(t *testing.T, contenido []byte) {
		manifiesto, err := instalacion.LeerManifiesto(contenido)
		if err != nil {
			var ilegible *instalacion.ManifiestoIlegible
			require.ErrorAs(t, err, &ilegible, "el rechazo no es un manifiesto ilegible")
			require.Zero(t, manifiesto, "un manifiesto ilegible no da ningún manifiesto")

			return
		}

		escrito, err := manifiesto.Bytes()
		require.NoError(t, err, "lo que se lee no se puede escribir")

		releido, err := instalacion.LeerManifiesto(escrito)
		require.NoError(t, err, "lo que se escribe no se puede leer: %q", escrito)
		require.Equal(t, manifiesto, releido, "releer lo escrito cambia el manifiesto")

		reescrito, err := releido.Bytes()
		require.NoError(t, err)
		require.Equal(t, string(escrito), string(reescrito), "reescribir lo releído cambia un byte")
	})
}
