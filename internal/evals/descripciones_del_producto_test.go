package evals

import (
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/skills"
)

// Este fichero protege las descripciones de las skills del producto, las de
// skills/ que empotra el binario (H41.4): que la description de cada una sea,
// byte a byte, su texto de la variante producto-descripciones-cortas (FR-001),
// que esté en un bloque >- de más de una línea (FR-002) y que no pase de 450
// caracteres ni la suma de todas de 1700 (FR-004). Lo corre make ci, sin red ni
// modelo (contracts/controles.md §1).

// La forma de la description de una skill de skills/ (FR-002): las líneas de la
// clave empiezan por aperturaDelBloqueDeLaDescripcion y son al menos
// lineasMinimasDelBloqueDeLaDescripcion, la de la clave y dos sangradas.
const (
	aperturaDelBloqueDeLaDescripcion      = claveDeLaDescripcion + " >-\n"
	lineasMinimasDelBloqueDeLaDescripcion = 3
)

// defectosDeLasDescripcionesDelProducto comprueba la description de cada skill
// de skills/ de la raíz (contracts/controles.md §1.1; research D7): por cada
// subdirectorio, en orden de nombre, como mucho una línea de cada clase, en este
// orden, la de su texto, que es el de la variante producto-descripciones-cortas
// (FR-001), la de su forma, un bloque >- de más de una línea (FR-002), y la de su
// longitud, de como mucho maximoDeUnaDescripcionCorta runas (FR-004); y al final,
// la de la suma de las de todas, de como mucho maximoDeLasDescripcionesCortas. La
// forma y la longitud se miden en toda skill, tenga o no texto en la variante.
// Los defectos de la variante no son de esta comprobación: los da
// TestDescripcionesCortas, y se descartan. El error, sin defectos, es el de un
// directorio o un fichero que no se puede leer o el de un frontmatter que no lee
// skills.LeerFrontmatter, con su ruta delante: un defecto inesperado (ADR 0023).
func defectosDeLasDescripcionesDelProducto(raiz string) ([]string, error) {
	textos, _, err := descripcionesCortas(raiz)
	if err != nil {
		return nil, err
	}

	arbol := os.DirFS(raiz)

	entradas, err := fs.ReadDir(arbol, directorioDeLasSkills)
	if err != nil {
		return nil, fmt.Errorf("%s: no se puede leer: %w", directorioDeLasSkills, err)
	}

	var (
		defectos []string
		suma     int
	)

	for _, entrada := range entradas {
		if !entrada.IsDir() {
			continue
		}

		skill := entrada.Name()

		deLaSkill, longitud, err := defectosDeLaDescripcionDeLaSkill(arbol, skill, textos)
		if err != nil {
			return nil, err
		}

		defectos = append(defectos, deLaSkill...)
		suma += longitud
	}

	if suma > maximoDeLasDescripcionesCortas {
		defectos = append(defectos, fmt.Sprintf("%s: las descripciones suman %d caracteres (máximo %d)",
			directorioDeLasSkills, suma, maximoDeLasDescripcionesCortas))
	}

	return defectos, nil
}

// defectosDeLaDescripcionDeLaSkill son las líneas de la description de la skill
// de skills/ del árbol, la de su texto contra el de la variante, textos, la de su
// forma y la de su longitud, en ese orden, y la longitud en runas de la
// description, la que suma defectosDeLasDescripcionesDelProducto. El error es el
// de su SKILL.md, que no se puede leer o cuyo frontmatter no lee
// skills.LeerFrontmatter, con su ruta delante.
func defectosDeLaDescripcionDeLaSkill(arbol fs.FS, skill string, textos map[string]string) ([]string, int, error) {
	directorio := path.Join(directorioDeLasSkills, skill)
	ruta := path.Join(directorio, ficheroDeLaSkill)

	contenido, err := fs.ReadFile(arbol, ruta)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: no se puede leer: %w", ruta, err)
	}

	frontmatter, err := skills.LeerFrontmatter(contenido)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: el frontmatter no se puede leer: %w", ruta, err)
	}

	var defectos []string

	texto, loTiene := textos[skill]

	switch {
	case !loTiene:
		defectos = append(defectos, fmt.Sprintf("%s: no tiene descripción en %s", directorio,
			carpetaDeLasDescripcionesCortas))
	case frontmatter.Description != texto:
		defectos = append(defectos, fmt.Sprintf("%s: la %s no es la de %s/%s%s", ruta, nombreDeLaClaveDeLaDescripcion,
			carpetaDeLasDescripcionesCortas, skill, extensionDeLaDescripcionCorta))
	}

	if !enUnBloqueDeMasDeUnaLinea(string(contenido)) {
		defectos = append(defectos, fmt.Sprintf("%s: la %s no está en un bloque >- de más de una línea", ruta,
			nombreDeLaClaveDeLaDescripcion))
	}

	longitud := utf8.RuneCountInString(frontmatter.Description)
	if longitud > maximoDeUnaDescripcionCorta {
		defectos = append(defectos, fmt.Sprintf("%s: la %s tiene %d caracteres (máximo %d)", ruta,
			nombreDeLaClaveDeLaDescripcion, longitud, maximoDeUnaDescripcionCorta))
	}

	return defectos, longitud, nil
}

// enUnBloqueDeMasDeUnaLinea dice si las líneas de la clave description del
// contenido de un SKILL.md, las que da descripcionYCuerpo, son un bloque >- de
// más de una línea (FR-002): empiezan por aperturaDelBloqueDeLaDescripcion y son
// al menos lineasMinimasDelBloqueDeLaDescripcion.
func enUnBloqueDeMasDeUnaLinea(contenido string) bool {
	deLaClave, _ := descripcionYCuerpo(contenido)
	if !strings.HasPrefix(deLaClave, aperturaDelBloqueDeLaDescripcion) {
		return false
	}

	lineas := 0
	for range strings.Lines(deLaClave) {
		lineas++
	}

	return lineas >= lineasMinimasDelBloqueDeLaDescripcion
}

// skillDeLaQuintaSkill es el SKILL.md de una quinta skill de skills/ que no tiene
// texto en la variante, el caso límite del spec (contracts/controles.md §1.3): su
// description, «Una skill de ensayo.», mide 20 caracteres.
const skillDeLaQuintaSkill = "---\nname: quinta\ndescription: >-\n  Una skill\n  de ensayo.\n---\n\n# Quinta\n"

// TestDescripcionesDelProducto comprueba las descripciones de skills/ contra la
// variante producto-descripciones-cortas y sus límites (FR-001, FR-002, FR-004;
// SC-001, SC-003; contracts/controles.md §1.2 y §1.3): la rama, sin ninguna
// línea, y cada mutante, la medida pasada de su umbral, con sus líneas exactas y
// en su orden (ADR 0029).
func TestDescripcionesDelProducto(t *testing.T) {
	t.Parallel()

	t.Run("la rama", func(t *testing.T) {
		t.Parallel()

		defectos, err := defectosDeLasDescripcionesDelProducto(raizDelRepositorio)
		require.NoError(t, err)

		compararLasLineasDeLosDefectos(t, nil, defectos)
	})

	casos := []struct {
		nombre    string
		mutante   func(t *testing.T, raiz string)
		esperadas []string
	}{
		{
			nombre: "un carácter añadido",
			mutante: func(t *testing.T, raiz string) {
				t.Helper()

				cambiarElSkill(t, raiz, "normativa", "lo dice la skill.", "lo dice la skill..")
			},
			esperadas: []string{
				"skills/normativa/SKILL.md: la description no es la de " +
					"laboratorio/producto-descripciones-cortas/normativa.txt",
			},
		},
		{
			nombre: "un carácter quitado",
			mutante: func(t *testing.T, raiz string) {
				t.Helper()

				cambiarElSkill(t, raiz, "consulta-juridica", "escrito presentar.", "escrito presentar")
			},
			esperadas: []string{
				"skills/consulta-juridica/SKILL.md: la description no es la de " +
					"laboratorio/producto-descripciones-cortas/consulta-juridica.txt",
			},
		},
		{
			nombre: "un carácter cambiado",
			mutante: func(t *testing.T, raiz string) {
				t.Helper()

				cambiarElSkill(t, raiz, "plazos-y-territorio", "tiene un municipio.", "tiene un municipio;")
			},
			esperadas: []string{
				"skills/plazos-y-territorio/SKILL.md: la description no es la de " +
					"laboratorio/producto-descripciones-cortas/plazos-y-territorio.txt",
			},
		},
		{
			nombre: "en un escalar de una línea",
			mutante: func(t *testing.T, raiz string) {
				t.Helper()

				ruta := filepath.Join(raiz, directorioDeLasSkills, "jurisprudencia", ficheroDeLaSkill)

				contenido, err := leerFichero(ruta)
				require.NoError(t, err)

				compuesto, err := skillConLaDescripcion(contenido, textoDeLaSkillEnLaVariante(t, raiz, "jurisprudencia"))
				require.NoError(t, err)
				require.NotEqual(t, string(contenido), string(compuesto), "el mutante cambia %s", ruta)

				require.NoError(t, escribirFichero(ruta, compuesto))
			},
			esperadas: []string{
				"skills/jurisprudencia/SKILL.md: la description no está en un bloque >- de más de una línea",
			},
		},
		{
			nombre: "en un bloque de una línea",
			mutante: func(t *testing.T, raiz string) {
				t.Helper()

				contenido, err := leerFichero(filepath.Join(raiz, directorioDeLasSkills, "plazos-y-territorio",
					ficheroDeLaSkill))
				require.NoError(t, err)

				deLaClave, _ := descripcionYCuerpo(string(contenido))
				enUnaLinea := "description: >-\n  " + textoDeLaSkillEnLaVariante(t, raiz, "plazos-y-territorio") + "\n"

				cambiarElSkill(t, raiz, "plazos-y-territorio", deLaClave, enUnaLinea)
			},
			esperadas: []string{
				"skills/plazos-y-territorio/SKILL.md: la description no está en un bloque >- de más de una línea",
			},
		},
		{
			nombre: "una de 451",
			mutante: func(t *testing.T, raiz string) {
				t.Helper()

				cambiarElSkill(t, raiz, "normativa", "lo dice la skill.", "lo dice la skill."+strings.Repeat("é", 25))
			},
			esperadas: []string{
				"skills/normativa/SKILL.md: la description no es la de " +
					"laboratorio/producto-descripciones-cortas/normativa.txt",
				"skills/normativa/SKILL.md: la description tiene 451 caracteres (máximo 450)",
				"skills: las descripciones suman 1715 caracteres (máximo 1700)",
			},
		},
		{
			nombre: "una suma de 1701",
			mutante: func(t *testing.T, raiz string) {
				t.Helper()

				cambiarElSkill(t, raiz, "plazos-y-territorio", "tiene un municipio.",
					"tiene un municipio."+strings.Repeat("é", 11))
			},
			esperadas: []string{
				"skills/plazos-y-territorio/SKILL.md: la description no es la de " +
					"laboratorio/producto-descripciones-cortas/plazos-y-territorio.txt",
				"skills: las descripciones suman 1701 caracteres (máximo 1700)",
			},
		},
		{
			nombre: "una quinta skill sin texto",
			mutante: func(t *testing.T, raiz string) {
				t.Helper()

				quinta := filepath.Join(raiz, directorioDeLasSkills, "quinta")
				require.NoError(t, os.Mkdir(quinta, 0o700))
				require.NoError(t, escribirFichero(filepath.Join(quinta, ficheroDeLaSkill), []byte(skillDeLaQuintaSkill)))
			},
			esperadas: []string{
				"skills/quinta: no tiene descripción en laboratorio/producto-descripciones-cortas",
				"skills: las descripciones suman 1710 caracteres (máximo 1700)",
			},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			raiz := copiaDeLasSkillsYLaVariante(t)
			caso.mutante(t, raiz)

			defectos, err := defectosDeLasDescripcionesDelProducto(raiz)
			require.NoError(t, err)

			compararLasLineasDeLosDefectos(t, caso.esperadas, defectos)
		})
	}
}

// copiaDeLasSkillsYLaVariante copia skills/ y la variante
// producto-descripciones-cortas del repositorio a un temporal del test, con sus
// rutas, y da la raíz de la copia, sobre la que se aplica un mutante.
func copiaDeLasSkillsYLaVariante(t *testing.T) string {
	t.Helper()

	raiz := t.TempDir()

	for _, carpeta := range []string{directorioDeLasSkills, carpetaDeLasDescripcionesCortas} {
		relativa := filepath.FromSlash(carpeta)

		require.NoError(t, os.CopyFS(filepath.Join(raiz, relativa), os.DirFS(filepath.Join(raizDelRepositorio, relativa))))
	}

	return raiz
}

// cambiarElSkill sustituye, en el SKILL.md de la skill de skills/ de la raíz, la
// primera aparición de de por a, y exige que el fichero cambie.
func cambiarElSkill(t *testing.T, raiz, skill, de, a string) {
	t.Helper()

	ruta := filepath.Join(raiz, directorioDeLasSkills, skill, ficheroDeLaSkill)

	contenido, err := leerFichero(ruta)
	require.NoError(t, err)

	cambiado := strings.Replace(string(contenido), de, a, 1)
	require.NotEqual(t, string(contenido), cambiado, "el mutante cambia %s", ruta)

	require.NoError(t, escribirFichero(ruta, []byte(cambiado)))
}

// consultaJuridica es la skill de skills/ que viene de la base de la rama salvo
// las líneas de la clave description (FR-003; contracts/controles.md §2), y
// baseDeLaRama, esa base abreviada, como la nombran sus defectos.
const (
	consultaJuridica = "consulta-juridica"
	baseDeLaRama     = "ab951c22"
)

// huellasDeConsultaJuridicaEnLaBase son los ficheros de skills/consulta-juridica/
// en la base de la rama, ab951c227e1fc8acac8bd6b23534e275e4a36a76 (git merge-base
// main HEAD; research V3), por su ruta relativa a la skill, con la SHA-256 en
// hexadecimal de su contenido; la del SKILL.md, de su contenido sin las líneas de
// la clave description, las que da descripcionYCuerpo (contracts/controles.md
// §2.1). Las rutas son las de git ls-tree -r --name-only ab951c22
// skills/consulta-juridica, y la huella se calculó sobre git show
// ab951c22:skills/consulta-juridica/SKILL.md dos veces, con la regla de
// lineasDelSkill y con awk y sha256sum. Se fijan aquí y no se leen con git
// porque el make ci del cierre no garantiza el historial (research D2).
var huellasDeConsultaJuridicaEnLaBase = map[string]string{
	ficheroDeLaSkill: "bd089909060d0fd82c1c6d2ee4ae148f6f0a1b2b7d7eefd476d164dc610a205a",
}

// defectoDeConsultaJuridica compara los ficheros de skills/consulta-juridica/ de
// la raíz con los de la base de la rama, huellasDeConsultaJuridicaEnLaBase, el
// SKILL.md sin las líneas de la clave description (FR-003; contracts/controles.md
// §2.2): en orden de ruta sobre la unión de las dos listas, da el primero que
// falta, sobra o difiere, o la cadena vacía. El error, sin defecto, es el de
// leerLosFicheros: un defecto inesperado (ADR 0023).
func defectoDeConsultaJuridica(raiz string) (string, error) {
	directorio := path.Join(directorioDeLasSkills, consultaJuridica)

	ficheros, err := leerLosFicheros(os.DirFS(raiz), directorio)
	if err != nil {
		return "", err
	}

	huellas := make(map[string]string, len(ficheros))

	for ruta, contenido := range ficheros {
		texto := string(contenido)
		if ruta == ficheroDeLaSkill {
			_, texto = descripcionYCuerpo(texto)
		}

		huellas[ruta] = sha256DeTexto(texto)
	}

	rutas := slices.AppendSeq(slices.Collect(maps.Keys(huellas)), maps.Keys(huellasDeConsultaJuridicaEnLaBase))
	slices.Sort(rutas)

	for _, ruta := range slices.Compact(rutas) {
		enLaRaiz, estaEnLaRaiz := huellas[ruta]
		enLaBase, estaEnLaBase := huellasDeConsultaJuridicaEnLaBase[ruta]
		completa := path.Join(directorio, ruta)

		switch {
		case !estaEnLaRaiz:
			return fmt.Sprintf("%s falta: está en la base de la rama (%s)", completa, baseDeLaRama), nil
		case !estaEnLaBase:
			return fmt.Sprintf("%s sobra: no está en la base de la rama (%s)", completa, baseDeLaRama), nil
		case enLaRaiz != enLaBase:
			return fmt.Sprintf("%s difiere de la base de la rama (%s)", completa, baseDeLaRama), nil
		}
	}

	return "", nil
}

// TestConsultaJuridicaDesdeLaBase comprueba que skills/consulta-juridica/ es la
// de la base de la rama salvo las líneas de la clave description (FR-003; SC-002;
// contracts/controles.md §2.3): la rama, sin defecto; cada mutante, un byte o un
// fichero de diferencia fuera de la clave, la medida pasada del umbral de 0, con
// su defecto exacto (ADR 0029); y el que solo cambia la clave, sin defecto.
func TestConsultaJuridicaDesdeLaBase(t *testing.T) {
	t.Parallel()

	t.Run("la rama", func(t *testing.T) {
		t.Parallel()

		defecto, err := defectoDeConsultaJuridica(raizDelRepositorio)
		require.NoError(t, err)

		require.Empty(t, defecto)
	})

	casos := []struct {
		nombre   string
		mutante  func(t *testing.T, raiz string)
		esperado string
	}{
		{
			nombre: "el cuerpo",
			mutante: func(t *testing.T, raiz string) {
				t.Helper()

				cambiarElSkill(t, raiz, consultaJuridica, "## Método", "## Métodos")
			},
			esperado: "skills/consulta-juridica/SKILL.md difiere de la base de la rama (ab951c22)",
		},
		{
			nombre: "name",
			mutante: func(t *testing.T, raiz string) {
				t.Helper()

				cambiarElSkill(t, raiz, consultaJuridica, "name: consulta-juridica", "name: consulta-juridico")
			},
			esperado: "skills/consulta-juridica/SKILL.md difiere de la base de la rama (ab951c22)",
		},
		{
			nombre: "un fichero de más",
			mutante: func(t *testing.T, raiz string) {
				t.Helper()

				referencias := filepath.Join(raiz, directorioDeLasSkills, consultaJuridica, "references")
				require.NoError(t, os.Mkdir(referencias, 0o700))
				require.NoError(t, escribirFichero(filepath.Join(referencias, "otra.md"), []byte("# Otra\n")))
			},
			esperado: "skills/consulta-juridica/references/otra.md sobra: no está en la base de la rama (ab951c22)",
		},
		{
			nombre: "un fichero de menos",
			mutante: func(t *testing.T, raiz string) {
				t.Helper()

				require.NoError(t, os.Remove(filepath.Join(raiz, directorioDeLasSkills, consultaJuridica, ficheroDeLaSkill)))
			},
			esperado: "skills/consulta-juridica/SKILL.md falta: está en la base de la rama (ab951c22)",
		},
		{
			nombre: "solo la clave",
			mutante: func(t *testing.T, raiz string) {
				t.Helper()

				cambiarElSkill(t, raiz, consultaJuridica, "Punto de entrada de toda", "Punto de entrada de cualquier")
			},
			esperado: "",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			raiz := copiaDeLasSkillsYLaVariante(t)
			caso.mutante(t, raiz)

			defecto, err := defectoDeConsultaJuridica(raiz)
			require.NoError(t, err)

			require.Equal(t, caso.esperado, defecto)
		})
	}
}

// textoDeLaSkillEnLaVariante es el texto de la skill en la variante
// producto-descripciones-cortas de la raíz, el de su fichero <skill>.txt sin el
// salto de línea final.
func textoDeLaSkillEnLaVariante(t *testing.T, raiz, skill string) string {
	t.Helper()

	texto, err := textoDeLaVariante(filepath.Join(raiz, filepath.FromSlash(carpetaDeLasDescripcionesCortas)),
		skill+extensionDeLaDescripcionCorta)
	require.NoError(t, err)

	return texto
}
