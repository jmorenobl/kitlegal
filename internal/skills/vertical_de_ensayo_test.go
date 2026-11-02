package skills_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/app"
	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/evals"
	"github.com/jmorenobl/kitlegal/internal/render"
	"github.com/jmorenobl/kitlegal/internal/skills"
)

// verticalDeEnsayo es la vertical de ensayo, relativa al directorio de este
// paquete, con la forma del repositorio: data/, skills/ y evals/
// (contracts/paquetes.md §5 de H32).
const verticalDeEnsayo = "testdata/vertical-de-ensayo"

// carpetasCopiadas son las del repositorio que el test copia, relativas a su
// raíz: las que la sincronía lee y la de las evals de la skill de la eval de la
// vertical. Sus huellas son las que el test anota del repositorio.
var carpetasCopiadas = []string{"skills", "data", "evals/normativa"}

// lineaDelEnsayo es la del enlace a la referencia de la vertical que la
// sincronía escribe entre las marcas de los paquetes de cada skill de capacidad
// que la tiene: 69 caracteres y su salto (contracts/paquetes.md §2).
const lineaDelEnsayo = "- Si la pregunta es de lo ensayo, lee [ensayo](references/ensayo.md).\n"

// lineaDelEnsayoEnLaEntrada es la de la vertical en la lista de paquetes de la
// skill de entrada, con las tres skills de capacidad que tienen su referencia
// (contracts/paquetes.md §4).
const lineaDelEnsayoEnLaEntrada = "- lo ensayo: `jurisprudencia`, `normativa`, `plazos-y-territorio`.\n"

// filaDelEnsayo es el principio de la fila de la cobertura de la vertical en la
// tabla generada de su referencia de plazos-y-territorio, y citaDelEnsayo, la
// forma legible de su precepto, en la primera norma, que no tiene abreviatura
// (contracts/paquetes.md §3).
const (
	filaDelEnsayo    = "| procedimientos de ensayo del molde de los paquetes | ⚠ NO CUBIERTO: "
	citaDelEnsayo    = "(art. 1 Ley 1/1900 [BOE-A-1900-1, bloque a1])"
	evalDelEnsayo    = "24-ensayo-norma-del-paquete.yaml"
	primeraDelEnsayo = "BOE-A-1900-1"
)

// TestVerticalDeEnsayo prueba el molde de los paquetes de materia (FR-024,
// SC-005; contracts/paquetes.md §5 de H32): la vertical de ensayo, instalada en
// una copia de skills/, data/ y evals/normativa/ del repositorio, pasa la
// sincronía y la comprobación de make skills-sync y make skills-check sin
// defectos ni derivas, su eval se carga, y para entrar el diff de la copia no
// tiene ningún .go ni cambia ningún SKILL.md fuera de las marcas que genera la
// sincronía; las huellas de lo copiado del repositorio —skills/ y data/ entre
// ello— no cambian. Su mutante, la línea del enlace a ensayo.md escrita a mano en
// el cuerpo de normativa, como la escribiría un paquete sin sincronía, pone en
// rojo la comparación fuera de las marcas.
func TestVerticalDeEnsayo(t *testing.T) {
	t.Parallel()

	delRepositorio := huellasDe(t, arbolDelRepositorio)
	raiz := copiaDelRepositorio(t)

	instalarLaVertical(t, raiz)

	descripciones := descripcionesDeProduccion(t, raiz)

	regenerado, err := skills.Regenerar(raiz, descripciones)
	require.NoError(t, err)
	require.Empty(t, textosDeLosFallos(regenerado.Defectos()), "la vertical se sincroniza sin defectos")
	require.NoError(t, skills.Escribir(raiz, regenerado))

	regenerado, err = skills.Regenerar(raiz, descripciones)
	require.NoError(t, err)

	derivas, err := skills.Comparar(raiz, regenerado)
	require.NoError(t, err)
	assert.Empty(t, textosDeLosFallos(regenerado.Defectos()), "la comprobación no da defectos")
	assert.Empty(t, textosDeLosFallos(derivas), "la comprobación no da derivas")

	comprobarLoSincronizado(t, raiz)

	eval, err := evals.LeerEval(evalDelEnsayo,
		[]byte(leerFicheroDePrueba(t, filepath.Join(raiz, "evals", "normativa", evalDelEnsayo))))
	require.NoError(t, err)
	assert.Contains(t, eval.Citas, evals.CitaEsperada{Norma: primeraDelEnsayo, Bloque: "a1"},
		"la eval cita la primera norma de la vertical")

	cambiados := ficherosCambiados(delRepositorio, huellasDe(t, raiz))
	assert.NotEmpty(t, cambiados, "la vertical cambia la copia")
	assert.Empty(t, ficherosGo(cambiados), "el diff de la copia no tiene ningún .go")
	assert.Empty(t, skillsCambiadasFueraDeLasMarcas(t, raiz),
		"ningún SKILL.md cambia fuera de las marcas de los paquetes")
	assert.Equal(t, delRepositorio, huellasDe(t, arbolDelRepositorio),
		"skills/, data/ y evals/normativa/ del repositorio no cambian")

	t.Run("mutante", func(t *testing.T) {
		t.Parallel()

		mutada := copiaDelRepositorio(t)
		instalarLaVertical(t, mutada)

		// Un molde que pide editar el cuerpo: la línea del enlace, a mano, junto a
		// las marcas y fuera de ellas.
		cambiarFicheroDePrueba(t, rutaDeSkill(mutada, "normativa", "SKILL.md"), finDeLosPaquetes,
			finDeLosPaquetes+lineaDelEnsayo)

		assert.Equal(t, []string{"skills/normativa/SKILL.md"}, skillsCambiadasFueraDeLasMarcas(t, mutada))
	})
}

// comprobarLoSincronizado comprueba que la sincronía ha enlazado la vertical en
// la copia de raíz: la línea de su enlace entre las marcas de cada skill de
// capacidad, la suya en la lista de la skill de entrada y la fila de su
// cobertura en la región, vacía en la vertical, de su referencia de
// plazos-y-territorio. Sin ellas, lo demás del test pasaría en vacío.
func comprobarLoSincronizado(t *testing.T, raiz string) {
	t.Helper()

	for _, skill := range []string{"jurisprudencia", "normativa", "plazos-y-territorio"} {
		_, dentro := partirPorLasMarcas(leerFicheroDePrueba(t, rutaDeSkill(raiz, skill, "SKILL.md")))
		assert.Contains(t, strings.Split(dentro, "\n"), strings.TrimSuffix(lineaDelEnsayo, "\n"),
			"%s enlaza la referencia de la vertical entre sus marcas", skill)
	}

	_, dentro := partirPorLasMarcas(leerFicheroDePrueba(t, rutaDeSkill(raiz, "consulta-juridica", "SKILL.md")))
	assert.Contains(t, strings.Split(dentro, "\n"), strings.TrimSuffix(lineaDelEnsayoEnLaEntrada, "\n"),
		"la skill de entrada lista la vertical")

	referencia := leerFicheroDePrueba(t, rutaDeSkill(raiz, "plazos-y-territorio", "references", "ensayo.md"))
	fila, hay := filaQueEmpiezaPor(referencia, filaDelEnsayo)
	require.True(t, hay, "la región de la referencia de plazos-y-territorio genera la fila de la vertical")
	assert.Contains(t, fila, citaDelEnsayo, "la fila cita el precepto en la forma legible")
}

// filaQueEmpiezaPor es la primera línea del texto que empieza por el principio
// dado, sin su salto, y si la hay.
func filaQueEmpiezaPor(texto, principio string) (string, bool) {
	for linea := range strings.Lines(texto) {
		if strings.HasPrefix(linea, principio) {
			return strings.TrimSuffix(linea, "\n"), true
		}
	}

	return "", false
}

// textosDeLosFallos son los textos de los defectos o de las derivas, uno por
// cada uno, con los que el test dice cuáles son si hay alguno.
func textosDeLosFallos[E error](fallos []E) []string {
	textos := make([]string, 0, len(fallos))
	for _, fallo := range fallos {
		textos = append(textos, fallo.Error())
	}

	return textos
}

// copiaDelRepositorio es una raíz temporal con una copia de las carpetas
// copiadas del repositorio.
func copiaDelRepositorio(t *testing.T) string {
	t.Helper()

	raiz := t.TempDir()

	for _, carpeta := range carpetasCopiadas {
		require.NoError(t, os.CopyFS(filepath.Join(raiz, filepath.FromSlash(carpeta)),
			os.DirFS(filepath.Join(arbolDelRepositorio, filepath.FromSlash(carpeta)))))
	}

	return raiz
}

// instalarLaVertical instala la vertical de ensayo en la copia de raíz como lo
// haría quien escribe un paquete: añade sus normas a data/normas.yaml y su fila a
// data/cobertura.yaml, y copia sus ficheros de skills/ y de evals/, que no pueden
// pisar ninguno de la copia (os.CopyFS no sobrescribe).
func instalarLaVertical(t *testing.T, raiz string) {
	t.Helper()

	anadirEntradas(t, raiz, "normas.yaml", "normas")
	anadirEntradas(t, raiz, "cobertura.yaml", "regulados_aparte")

	for _, carpeta := range []string{"skills", "evals"} {
		require.NoError(t, os.CopyFS(filepath.Join(raiz, carpeta), os.DirFS(filepath.Join(verticalDeEnsayo, carpeta))))
	}
}

// anadirEntradas añade al fichero de data/ de la copia de raíz las entradas del
// de la vertical: lo que este lleva tras la línea de su clave, que en los dos es
// la única de primer nivel y va al final.
func anadirEntradas(t *testing.T, raiz, fichero, clave string) {
	t.Helper()

	_, entradas, hay := strings.Cut(leerFicheroDePrueba(t, filepath.Join(verticalDeEnsayo, "data", fichero)),
		"\n"+clave+":\n")
	require.True(t, hay, "data/%s de la vertical lleva sus entradas bajo %s:", fichero, clave)

	ruta := filepath.Join(raiz, "data", fichero)
	copiado := leerFicheroDePrueba(t, ruta)
	require.True(t, strings.HasSuffix(copiado, "\n"), "data/%s del repositorio termina en un salto de línea", fichero)

	escribirFicheroDePrueba(t, ruta, copiado+entradas)
}

// descripcionesDeProduccion son las descripciones de los verbos de los applets
// que declaran las skills del árbol de raíz, con el registro de producción, como
// las lee make skills-sync (TestSkillsDelRepositorio, en internal/app): lo que
// emite cli.Describir para cada verbo, con el cli.Verbo en que el kernel lo
// traduce para --describe, y para un verbo sin argumentos, que da las banderas
// globales. La orden con --describe no sirve aquí, porque pide los argumentos de
// posición del verbo. Si el kernel lo tradujera de otro modo, las tablas de
// comandos regeneradas en la copia no serían las del repositorio y la
// comprobación daría su deriva.
func descripcionesDeProduccion(t *testing.T, raiz string) []skills.DescripcionDeVerbo {
	t.Helper()

	registro, err := app.RegistroDeProduccion("")
	require.NoError(t, err)

	var globales, erroresDeGlobales bytes.Buffer
	require.NoError(t, cli.Describir(render.Nuevo(&globales, &erroresDeGlobales), cli.Verbo{}),
		erroresDeGlobales.String())

	var descripciones []skills.DescripcionDeVerbo

	for _, nombre := range appletsDeLasSkills(t, raiz) {
		applet, registrado := registro.Buscar(nombre)
		require.True(t, registrado, "el applet %s está registrado", nombre)

		for _, verbo := range applet.Verbos() {
			require.NotNil(t, verbo.Argumentos, "%s %s tiene argumentos que describir", nombre, verbo.Nombre)

			var salida, errores bytes.Buffer

			require.NoError(t, cli.Describir(render.Nuevo(&salida, &errores), cli.Verbo{
				Applet:     applet.Nombre(),
				Verbo:      verbo.Nombre,
				Ayuda:      verbo.Descripcion,
				Argumentos: verbo.Argumentos(),
				Salida:     verbo.Salida,
			}), "%s %s --describe: %s", nombre, verbo.Nombre, errores.String())

			descripcion, err := skills.LeerDescripcionDeVerbo(salida.Bytes(), globales.Bytes())
			require.NoError(t, err)

			descripciones = append(descripciones, descripcion)
		}
	}

	return descripciones
}

// appletsDeLasSkills son los applets que declara en kitlegal-applets alguna
// skill del árbol de raíz, cada uno una vez y en el orden en que aparece por
// primera vez.
func appletsDeLasSkills(t *testing.T, raiz string) []string {
	t.Helper()

	nombres, err := skills.Listar(raiz)
	require.NoError(t, err)

	var applets []string

	for _, nombre := range nombres {
		skill, err := skills.Cargar(raiz, nombre)
		require.NoError(t, err)

		frontmatter, err := skills.LeerFrontmatter(skill.Contenido)
		require.NoError(t, err)

		for _, applet := range frontmatter.DeclaracionDeKitlegal().Applets {
			if !slices.Contains(applets, applet) {
				applets = append(applets, applet)
			}
		}
	}

	return applets
}

// ficherosCambiados son las rutas del diff entre dos árboles, por sus huellas:
// las de los ficheros que solo están en uno de los dos o que en los dos tienen
// huellas distintas, en orden.
func ficherosCambiados(antes, despues map[string]string) []string {
	var cambiados []string

	for ruta, huella := range antes {
		if despues[ruta] != huella {
			cambiados = append(cambiados, ruta)
		}
	}

	for ruta := range despues {
		if _, estaba := antes[ruta]; !estaba {
			cambiados = append(cambiados, ruta)
		}
	}

	slices.Sort(cambiados)

	return cambiados
}

// ficherosGo son las rutas .go de las dadas, en su orden.
func ficherosGo(rutas []string) []string {
	var deGo []string

	for _, ruta := range rutas {
		if path.Ext(ruta) == ".go" {
			deGo = append(deGo, ruta)
		}
	}

	return deGo
}

// skillsCambiadasFueraDeLasMarcas son las rutas, relativas a la raíz, de los
// SKILL.md de la copia de raíz que no son los del repositorio fuera de las
// marcas de los paquetes de §2 y §4 de contracts/paquetes.md. Un paquete no
// añade ni quita skills.
func skillsCambiadasFueraDeLasMarcas(t *testing.T, raiz string) []string {
	t.Helper()

	delRepositorio, err := skills.Listar(arbolDelRepositorio)
	require.NoError(t, err)

	deLaCopia, err := skills.Listar(raiz)
	require.NoError(t, err)
	require.Equal(t, delRepositorio, deLaCopia, "la copia tiene las skills del repositorio")

	var cambiadas []string

	for _, nombre := range delRepositorio {
		original, _ := partirPorLasMarcas(leerFicheroDePrueba(t, rutaDeSkill(arbolDelRepositorio, nombre, "SKILL.md")))
		copiado, _ := partirPorLasMarcas(leerFicheroDePrueba(t, rutaDeSkill(raiz, nombre, "SKILL.md")))

		if copiado != original {
			cambiadas = append(cambiadas, path.Join("skills", nombre, "SKILL.md"))
		}
	}

	return cambiadas
}

// partirPorLasMarcas separa un SKILL.md en lo que va fuera de las marcas de los
// paquetes, con las líneas de las marcas, y lo que va entre ellas.
func partirPorLasMarcas(contenido string) (fuera, dentro string) {
	var (
		deFuera, deDentro strings.Builder
		enLaRegion        bool
	)

	for linea := range strings.Lines(contenido) {
		switch linea {
		case inicioDeLosPaquetes, inicioDeLosPaquetesDeLaBase:
			enLaRegion = true

			deFuera.WriteString(linea)
		case finDeLosPaquetes, finDeLosPaquetesDeLaBase:
			enLaRegion = false

			deFuera.WriteString(linea)
		default:
			if enLaRegion {
				deDentro.WriteString(linea)
			} else {
				deFuera.WriteString(linea)
			}
		}
	}

	return deFuera.String(), deDentro.String()
}

// huellasDe son las huellas SHA-256 de cada fichero de las carpetas copiadas del
// árbol de raíz —el repositorio o su copia—, por su ruta relativa a ella.
func huellasDe(t *testing.T, raiz string) map[string]string {
	t.Helper()

	arbol := os.DirFS(raiz)
	huellas := make(map[string]string)

	for _, carpeta := range carpetasCopiadas {
		require.NoError(t, fs.WalkDir(arbol, carpeta, func(ruta string, entrada fs.DirEntry, err error) error {
			if err != nil || entrada.IsDir() {
				return err
			}

			contenido, err := fs.ReadFile(arbol, ruta)
			if err != nil {
				return err
			}

			suma := sha256.Sum256(contenido)
			huellas[ruta] = hex.EncodeToString(suma[:])

			return nil
		}))
	}

	return huellas
}
