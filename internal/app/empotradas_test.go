package app

import (
	"io/fs"
	"path"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal"
	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
)

// TestSkillsEmpotradasDelBinario fija que lo que el binario lleva dentro se lee
// a una skill por cada directorio de skills/ con SKILL.md, en orden de nombre y
// con cada fichero, su ruta dentro de la skill, sus bytes tal cual y su huella
// (contracts/skills-e-invocacion.md §1; FR-001, FR-003, FR-004). Que lo empotrado
// sea el árbol lo fija TestSkillsEmpotradas, en la raíz del módulo.
func TestSkillsEmpotradasDelBinario(t *testing.T) {
	t.Parallel()

	empotradas, err := skillsEmpotradas()
	require.NoError(t, err)

	nombres := make([]string, 0, len(empotradas))
	for _, skill := range empotradas {
		nombres = append(nombres, skill.Nombre)
	}

	require.Subset(t, nombres, skillsExigidas, "lo empotrado no tiene las skills de hoy")
	assert.True(t, slices.IsSorted(nombres), "las skills no van en orden de nombre: %v", nombres)

	for _, skill := range empotradas {
		require.NotEmpty(t, skill.Ficheros, "la skill %s no tiene ficheros", skill.Nombre)
		assert.Equal(t, ficheroDeLaSkill, skill.Ficheros[0].Ruta, "el primero de %s no es su SKILL.md", skill.Nombre)
		assert.True(t, slices.IsSortedFunc(skill.Ficheros, porRuta), "los ficheros de %s no van en orden de ruta",
			skill.Nombre)

		for _, fichero := range skill.Ficheros {
			empotrado, err := fs.ReadFile(kitlegal.Skills(), path.Join(carpetaDeSkills, skill.Nombre, fichero.Ruta))
			require.NoError(t, err)
			assert.Equal(t, instalacion.NuevoFicheroEmpotrado(fichero.Ruta, empotrado), fichero)
		}
	}
}

// TestSkillsEmpotradasDe fija la lectura de lo empotrado sobre árboles
// sintéticos, con lo que el embebido real no da: una skill sin SKILL.md, rutas
// cuyo orden no es el del recorrido y cada rama de error.
func TestSkillsEmpotradasDe(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string
		arbol  fs.FS
		quiere []instalacion.SkillEmpotrada
	}{
		{
			// Las skills van en orden de nombre y sus ficheros en orden de ruta, que
			// no es el del recorrido: este da references/a/b.md antes que
			// references/a-b.md, que va antes, porque «-» es menor que «/».
			nombre: "en-orden-de-nombre-y-de-ruta",
			arbol: fstest.MapFS{
				"skills/zeta/SKILL.md":                {Data: []byte("zeta")},
				"skills/alfa/SKILL.md":                {Data: []byte("alfa")},
				"skills/alfa/references/a/b.md":       {Data: []byte("a/b")},
				"skills/alfa/references/a-b.md":       {Data: []byte("a-b")},
				"skills/alfa/references/.oculto":      {Data: []byte("oculto")},
				"skills/alfa/references/a/c/hondo.md": {Data: []byte("hondo")},
			},
			quiere: []instalacion.SkillEmpotrada{
				{Nombre: "alfa", Ficheros: []instalacion.FicheroEmpotrado{
					instalacion.NuevoFicheroEmpotrado("SKILL.md", []byte("alfa")),
					instalacion.NuevoFicheroEmpotrado("references/.oculto", []byte("oculto")),
					instalacion.NuevoFicheroEmpotrado("references/a-b.md", []byte("a-b")),
					instalacion.NuevoFicheroEmpotrado("references/a/b.md", []byte("a/b")),
					instalacion.NuevoFicheroEmpotrado("references/a/c/hondo.md", []byte("hondo")),
				}},
				{Nombre: "zeta", Ficheros: []instalacion.FicheroEmpotrado{
					instalacion.NuevoFicheroEmpotrado("SKILL.md", []byte("zeta")),
				}},
			},
		},
		{
			// Un directorio con references/ y sin SKILL.md no es una skill (FR-004),
			// ni lo es una entrada de skills/ que no es un directorio.
			nombre: "sin-skill-md-no-entra",
			arbol: fstest.MapFS{
				"skills/beta/references/normas.md": {Data: []byte("normas")},
				"skills/LEEME.md":                  {Data: []byte("léeme")},
				"skills/gamma/SKILL.md":            {Data: []byte("gamma")},
			},
			quiere: []instalacion.SkillEmpotrada{
				{Nombre: "gamma", Ficheros: []instalacion.FicheroEmpotrado{
					instalacion.NuevoFicheroEmpotrado("SKILL.md", []byte("gamma")),
				}},
			},
		},
		{
			// Un SKILL.md que no es un fichero tampoco hace una skill.
			nombre: "skill-md-que-es-un-directorio",
			arbol:  fstest.MapFS{"skills/delta/SKILL.md/dentro": {Data: []byte("dentro")}},
		},
		{
			// De una skill solo entran SKILL.md y references/ (FR-001): ni scripts/
			// ni un fichero que se llama references.
			nombre: "solo-skill-md-y-references",
			arbol: fstest.MapFS{
				"skills/epsilon/SKILL.md":            {Data: []byte("epsilon")},
				"skills/epsilon/scripts/boe":         {Data: []byte("boe")},
				"skills/epsilon/otro.md":             {Data: []byte("otro")},
				"skills/epsilon/references-no.md":    {Data: []byte("no")},
				"skills/epsilon/references/norma.md": {Data: []byte("norma")},
			},
			quiere: []instalacion.SkillEmpotrada{
				{Nombre: "epsilon", Ficheros: []instalacion.FicheroEmpotrado{
					instalacion.NuevoFicheroEmpotrado("SKILL.md", []byte("epsilon")),
					instalacion.NuevoFicheroEmpotrado("references/norma.md", []byte("norma")),
				}},
			},
		},
		{
			nombre: "skills-vacio",
			arbol:  fstest.MapFS{"skills": {Mode: fs.ModeDir}},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			empotradas, err := skillsEmpotradasDe(caso.arbol)
			require.NoError(t, err)
			assert.Equal(t, caso.quiere, empotradas)
		})
	}
}

// TestSkillsEmpotradasDeConErrores fija que un árbol que no se puede leer es un
// error que nombra lo que falló, y no una skill que falta o un fichero de menos.
func TestSkillsEmpotradasDeConErrores(t *testing.T) {
	t.Parallel()

	arbol := fstest.MapFS{
		"skills/alfa/SKILL.md":            {Data: []byte("alfa")},
		"skills/alfa/references/norma.md": {Data: []byte("norma")},
	}

	casos := []struct {
		nombre   string
		arbol    fs.FS
		mensajes []string
	}{
		{
			nombre:   "sin-skills",
			arbol:    fstest.MapFS{},
			mensajes: []string{"app: no se puede listar skills en lo empotrado: ", "file does not exist"},
		},
		{
			nombre:   "skill-md-que-no-se-examina",
			arbol:    arbolQueFallaEn{FS: arbol, ruta: "skills/alfa/SKILL.md"},
			mensajes: []string{"app: no se puede examinar skills/alfa/SKILL.md en lo empotrado: ", "permission denied"},
		},
		{
			nombre:   "referencia-que-no-se-lee",
			arbol:    arbolQueFallaEn{FS: arbol, ruta: "skills/alfa/references/norma.md"},
			mensajes: []string{"app: no se puede leer la skill alfa en lo empotrado: ", "skills/alfa/references/norma.md"},
		},
		{
			nombre:   "references-que-no-se-lista",
			arbol:    arbolQueFallaEn{FS: arbol, ruta: "skills/alfa/references"},
			mensajes: []string{"app: no se puede leer la skill alfa en lo empotrado: ", "skills/alfa/references"},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			empotradas, err := skillsEmpotradasDe(caso.arbol)
			require.Error(t, err)

			for _, mensaje := range caso.mensajes {
				assert.Contains(t, err.Error(), mensaje)
			}

			assert.Nil(t, empotradas)
		})
	}
}

// arbolQueFallaEn es el árbol que envuelve salvo en una ruta, que no deja abrir:
// ni examinarla, ni leerla, ni listarla. Solo expone Open, de modo que fs.Stat,
// fs.ReadFile y fs.ReadDir pasan todos por él.
type arbolQueFallaEn struct {
	fs.FS

	ruta string
}

func (a arbolQueFallaEn) Open(nombre string) (fs.File, error) {
	if nombre == a.ruta || strings.HasPrefix(nombre, a.ruta+"/") {
		return nil, &fs.PathError{Op: "open", Path: nombre, Err: fs.ErrPermission}
	}

	return a.FS.Open(nombre)
}
