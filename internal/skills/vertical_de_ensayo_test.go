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

// lineaDeLoCubiertoDelEnsayo es la de la fila de lo cubierto de la vertical en la
// región de lo que cubre la base de la skill de entrada: la segunda de los
// plazos, detrás de las del repositorio (contracts/datos-y-regiones.md §6 de
// H33).
const lineaDeLoCubiertoDelEnsayo = "- **Plazos**: los de ensayo del molde de los paquetes."

// skillsConLaVertical son las skills de capacidad que tienen la referencia de la
// vertical, ensayo.md, y que enlazan su línea entre las marcas de los paquetes.
var skillsConLaVertical = []string{"jurisprudencia", "normativa", "plazos-y-territorio"}

// referenciaDeLosPlazosDelEnsayo es la referencia de la vertical en
// plazos-y-territorio, relativa a la raíz: la que lleva las regiones de sus tipos
// y de su cobertura.
var referenciaDeLosPlazosDelEnsayo = path.Join("skills", "plazos-y-territorio", "references", "ensayo.md")

// Las dos filas de la vertical que la sincronía da a su paquete aunque algo suyo
// sea de lo administrativo (contracts/sincronia-de-los-paquetes.md §3 y §5 de
// H45): el principio de la fila de su tipo, que se cuenta con el cómputo de la
// LPAC, en la tabla de los recursos y plazos; y el procedimiento de su fila de lo
// regulado aparte con su precepto en la LPAC y la norma que la rige, la segunda de
// la vertical, con el motivo de su línea, que nombra esa norma y el precepto que
// la deja aparte, como hoy (FR-007).
const (
	filaDelTipoDelEnsayo   = "| plazo de ensayo del molde de los paquetes | `ensayo` | "
	procedimientoConNorma  = "procedimientos de ensayo que la ley común deja aparte"
	motivoDeLaFilaConNorma = "se rige por Ley 2/1900, de ensayo del molde de los paquetes de materia (BOE-A-1900-2), " +
		"no por el procedimiento común (DA 1ª.2.c LPAC [BOE-A-2015-10565, bloque daprimera])"
)

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
// SC-005; contracts/paquetes.md §5 de H32; contracts/sincronia-de-los-paquetes.md
// §5 de H45): la vertical de ensayo, instalada en una copia de skills/, data/ y
// evals/normativa/ del repositorio, pasa la sincronía y la comprobación de make
// skills-sync y make skills-check sin defectos ni derivas, su eval se carga, y
// para entrar el diff de la copia no tiene ningún .go ni cambia ningún SKILL.md
// fuera de las marcas que genera la sincronía; las huellas de lo copiado del
// repositorio —skills/ y data/ entre ello— no cambian. Su tipo y su fila con
// norma salen solo en su referencia de plazos-y-territorio (FR-002, FR-005,
// SC-002); el sitio de los demás paquetes lo comprueba el presupuesto del cuerpo
// sin sus líneas (ADR 0063, punto 1). Su mutante, la línea del enlace a ensayo.md escrita a
// mano en el cuerpo de normativa, como la escribiría un paquete sin sincronía,
// pone en rojo la comparación fuera de las marcas; y sin-la-region, la referencia
// de la vertical sin la región de la cobertura, da el defecto de su fila con
// norma (FR-006).
func TestVerticalDeEnsayo(t *testing.T) {
	t.Parallel()

	delRepositorio := huellasDe(t, arbolDelRepositorio)
	raiz := copiaDelRepositorio(t)
	ultimaDeLoCubierto := ultimaLineaDeLoCubierto(t, raiz)

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

	comprobarLoSincronizado(t, raiz, ultimaDeLoCubierto)
	comprobarElReparto(t, raiz)

	eval, err := evals.LeerEval(evalDelEnsayo,
		[]byte(leerFicheroDePrueba(t, filepath.Join(raiz, "evals", "normativa", evalDelEnsayo))))
	require.NoError(t, err)
	assert.Contains(t, eval.Citas, evals.CitaEsperada{Norma: primeraDelEnsayo, Bloque: "a1"},
		"la eval cita la primera norma de la vertical")

	cambiados := ficherosCambiados(delRepositorio, huellasDe(t, raiz))
	assert.NotEmpty(t, cambiados, "la vertical cambia la copia")
	assert.Empty(t, ficherosGo(cambiados), "el diff de la copia no tiene ningún .go")
	assert.Empty(t, skillsCambiadasFueraDeLasMarcas(t, raiz),
		"ningún SKILL.md cambia fuera de las marcas de los paquetes y de lo cubierto")
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

	t.Run("sin-la-region", func(t *testing.T) {
		t.Parallel()

		sinLaRegion := copiaDelRepositorio(t)
		instalarLaVertical(t, sinLaRegion)

		// La referencia de la vertical de plazos-y-territorio, sin las marcas de la
		// región de la cobertura: su paquete deja de tenerla, y la fila con norma, que
		// es de él aunque su precepto sea de lo administrativo, no la genera ninguna.
		cambiarFicheroDePrueba(t, filepath.Join(sinLaRegion, filepath.FromSlash(referenciaDeLosPlazosDelEnsayo)),
			inicioDeLaCobertura+finDeLaCobertura, "")

		regenerado, err := skills.Regenerar(sinLaRegion, descripcionesDeProduccion(t, sinLaRegion))
		require.NoError(t, err)
		assert.Contains(t, textosDeLosFallos(regenerado.Defectos()), "plazos-y-territorio: data/cobertura.yaml: la fila "+
			"de «"+procedimientoConNorma+"» no la genera ninguna región cobertura_por_materia")
	})
}

// ultimaLineaDeLoCubierto es la última línea de la región de lo que cubre la base
// de la skill de entrada de la copia de raíz, que la lleva con alguna línea.
func ultimaLineaDeLoCubierto(t *testing.T, raiz string) string {
	t.Helper()

	cubierto := lineasEntreLasMarcasDeLoCubierto(t,
		leerFicheroDePrueba(t, rutaDeSkill(raiz, "consulta-juridica", "SKILL.md")))
	require.NotEmpty(t, cubierto, "la región de lo cubierto de la skill de entrada de la copia")

	return cubierto[len(cubierto)-1]
}

// comprobarLoSincronizado comprueba que la sincronía ha enlazado la vertical en
// la copia de raíz: la línea de su enlace entre las marcas de cada skill de
// capacidad, la suya en la lista de la skill de entrada, la de su fila de lo
// cubierto como última de la región de lo que cubre la base de la skill de
// entrada, detrás de la que era la última de la copia antes de instalarla —dos
// filas de una capacidad dan dos líneas (FR-007 de H33)—, y la fila de su
// cobertura sin norma en la región, vacía en la vertical, de su referencia de
// plazos-y-territorio. Sin ellas, lo demás del test pasaría en vacío.
func comprobarLoSincronizado(t *testing.T, raiz, ultimaDeAntes string) {
	t.Helper()

	for _, skill := range skillsConLaVertical {
		_, dentro := partirPorLasMarcas(leerFicheroDePrueba(t, rutaDeSkill(raiz, skill, "SKILL.md")))
		assert.Contains(t, strings.Split(dentro, "\n"), strings.TrimSuffix(lineaDelEnsayo, "\n"),
			"%s enlaza la referencia de la vertical entre sus marcas", skill)
	}

	entrada := leerFicheroDePrueba(t, rutaDeSkill(raiz, "consulta-juridica", "SKILL.md"))
	_, dentro := partirPorLasMarcas(entrada)
	assert.Contains(t, strings.Split(dentro, "\n"), strings.TrimSuffix(lineaDelEnsayoEnLaEntrada, "\n"),
		"la skill de entrada lista la vertical")

	cubierto := lineasEntreLasMarcasDeLoCubierto(t, entrada)
	require.GreaterOrEqual(t, len(cubierto), 2, "la región de lo cubierto de la skill de entrada")
	assert.Equal(t, lineaDeLoCubiertoDelEnsayo, cubierto[len(cubierto)-1],
		"la fila de lo cubierto de la vertical es la última línea de la región")
	assert.Equal(t, ultimaDeAntes, cubierto[len(cubierto)-2],
		"la fila de la vertical va detrás de la que era la última de la región")

	referencia := leerFicheroDePrueba(t, filepath.Join(raiz, filepath.FromSlash(referenciaDeLosPlazosDelEnsayo)))
	fila, hay := filaQueEmpiezaPor(referencia, filaDelEnsayo)
	require.True(t, hay, "la región de la referencia de plazos-y-territorio genera la fila de la vertical")
	assert.Contains(t, fila, citaDelEnsayo, "la fila cita el precepto en la forma legible")
}

// comprobarElReparto comprueba que la sincronía ha dado a la vertical, en la
// copia de raíz, lo suyo que se apoya en lo administrativo (contracts/
// sincronia-de-los-paquetes.md §5 de H45; FR-002, FR-005, FR-007, SC-002): la
// fila de su tipo, que se cuenta con el cómputo de la LPAC, en la región
// recursos_y_plazos de su referencia de plazos-y-territorio y en ninguna otra
// referencia de la copia; y su fila de lo regulado aparte con su precepto en la
// LPAC, en la región de la cobertura de esa misma referencia, con la línea de hoy,
// y en ninguna otra, tampoco en ninguna línea de administrativo.md, la del
// paquete de su precepto.
func comprobarElReparto(t *testing.T, raiz string) {
	t.Helper()

	referencia := leerFicheroDePrueba(t, filepath.Join(raiz, filepath.FromSlash(referenciaDeLosPlazosDelEnsayo)))

	_, hay := filaQueEmpiezaPor(entreLasMarcas(t, referencia, inicioDeLosPlazos, finDeLosPlazos), filaDelTipoDelEnsayo)
	assert.True(t, hay, "la región recursos_y_plazos de la referencia de la vertical genera la fila de su tipo")
	assert.Equal(t, []string{referenciaDeLosPlazosDelEnsayo}, referenciasQueContienen(t, raiz, filaDelTipoDelEnsayo),
		"la fila del tipo de ensayo sale en la referencia de la vertical y en ninguna otra")

	fila, hay := filaQueEmpiezaPor(entreLasMarcas(t, referencia, inicioDeLaCobertura, finDeLaCobertura),
		"| "+procedimientoConNorma+" | ")
	if assert.True(t, hay, "la región de la cobertura de la referencia de la vertical genera la fila con norma") {
		assert.Contains(t, fila, motivoDeLaFilaConNorma,
			"la línea nombra la norma que rige la fila y el precepto que la deja aparte")
	}

	assert.Equal(t, []string{referenciaDeLosPlazosDelEnsayo}, referenciasQueContienen(t, raiz, procedimientoConNorma),
		"la fila con norma sale en la referencia de la vertical y en ninguna otra, tampoco en administrativo.md")
}

// entreLasMarcas es lo que hay en el contenido entre la línea de la marca de
// inicio y la de fin de una región, cada una con su salto y una vez cada una.
func entreLasMarcas(t *testing.T, contenido, inicio, fin string) string {
	t.Helper()

	require.Equal(t, 1, strings.Count(contenido, inicio), "una marca %q", inicio)
	require.Equal(t, 1, strings.Count(contenido, fin), "una marca %q", fin)

	_, tras, _ := strings.Cut(contenido, inicio)
	entre, _, conFin := strings.Cut(tras, fin)
	require.True(t, conFin, "la marca %q detrás de la de inicio", fin)

	return entre
}

// referenciasQueContienen son las rutas, relativas a la raíz y en orden, de las
// referencias de las skills de la copia de raíz con alguna línea que contiene el
// texto dado, que es de una sola línea.
func referenciasQueContienen(t *testing.T, raiz, texto string) []string {
	t.Helper()

	rutas, err := fs.Glob(os.DirFS(raiz), "skills/*/references/*.md")
	require.NoError(t, err)

	var con []string

	for _, ruta := range rutas {
		if strings.Contains(leerFicheroDePrueba(t, filepath.Join(raiz, filepath.FromSlash(ruta))), texto) {
			con = append(con, ruta)
		}
	}

	return con
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
// haría quien escribe un paquete: añade sus normas a data/normas.yaml, sus tipos
// de plazo a data/plazos.yaml y sus filas de lo cubierto y de lo regulado aparte a
// data/cobertura.yaml, y copia sus ficheros de skills/ y de evals/, que no pueden
// pisar ninguno de la copia (os.CopyFS no sobrescribe).
func instalarLaVertical(t *testing.T, raiz string) {
	t.Helper()

	anadirEntradas(t, raiz, "normas.yaml", "normas")
	anadirEntradas(t, raiz, "plazos.yaml", "tipos")
	anadirEntradas(t, raiz, "cobertura.yaml", "cubierto", "regulados_aparte")

	for _, carpeta := range []string{"skills", "evals"} {
		require.NoError(t, os.CopyFS(filepath.Join(raiz, carpeta), os.DirFS(filepath.Join(verticalDeEnsayo, carpeta))))
	}
}

// anadirEntradas añade al fichero de data/ de la copia de raíz las entradas de
// cada clave del de la vertical, en su orden, al final del bloque de esa misma
// clave en la copia (research D17 de H33): cubierto va en medio de
// data/cobertura.yaml, y una segunda clave al final la rechazaría el lector. Las
// entradas de una clave, en los dos ficheros, son las líneas sangradas que siguen
// a la suya, que es de primer nivel; su bloque acaba en la primera línea que no
// lo es.
func anadirEntradas(t *testing.T, raiz, fichero string, claves ...string) {
	t.Helper()

	deLaVertical := leerFicheroDePrueba(t, filepath.Join(verticalDeEnsayo, "data", fichero))
	ruta := filepath.Join(raiz, "data", fichero)
	copiado := leerFicheroDePrueba(t, ruta)
	require.True(t, strings.HasSuffix(copiado, "\n"), "data/%s del repositorio termina en un salto de línea", fichero)

	for _, clave := range claves {
		_, entradas := bloqueDeLaClave(t, deLaVertical, clave)
		require.NotEmpty(t, entradas, "data/%s de la vertical lleva sus entradas bajo %s:", fichero, clave)

		fin, _ := bloqueDeLaClave(t, copiado, clave)
		copiado = copiado[:fin] + entradas + copiado[fin:]
	}

	escribirFicheroDePrueba(t, ruta, copiado)
}

// bloqueDeLaClave es, en el texto YAML, el byte en que acaba el bloque de la
// clave de primer nivel —el que sigue a su última línea sangrada— y las líneas
// sangradas de ese bloque, cada una con su salto. El texto lleva la clave en una
// línea, una sola vez.
func bloqueDeLaClave(t *testing.T, texto, clave string) (int, string) {
	t.Helper()

	linea := clave + ":\n"
	require.Equal(t, 1, strings.Count("\n"+texto, "\n"+linea), "la clave %s, una vez en su línea", clave)

	comienzo := strings.Index("\n"+texto, "\n"+linea) + len(linea)
	fin := comienzo

	for resto := range strings.Lines(texto[comienzo:]) {
		if !strings.HasPrefix(resto, " ") {
			break
		}

		fin += len(resto)
	}

	return fin, texto[comienzo:fin]
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
// marcas de los paquetes de §2 y §4 de contracts/paquetes.md y de las de lo que
// cubre la base (contracts/datos-y-regiones.md §3 de H33). Un paquete no añade
// ni quita skills.
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
// paquetes y de lo que cubre la base, con las líneas de las marcas, y lo que va
// entre ellas.
func partirPorLasMarcas(contenido string) (fuera, dentro string) {
	var (
		deFuera, deDentro strings.Builder
		enLaRegion        bool
	)

	for linea := range strings.Lines(contenido) {
		switch linea {
		case inicioDeLosPaquetes, inicioDeLosPaquetesDeLaBase, inicioDeLoCubierto:
			enLaRegion = true

			deFuera.WriteString(linea)
		case finDeLosPaquetes, finDeLosPaquetesDeLaBase, finDeLoCubierto:
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
