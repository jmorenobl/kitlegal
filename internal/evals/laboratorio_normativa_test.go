package evals

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestVariantesDeNormativa es el control de las variantes declaradas de
// normativa del laboratorio (FR-003 a FR-006, SC-001 y SC-002 de H40; FR-021;
// contracts/variantes.md §3 a §5 y §7 de H40; data-model §1 a §3 de H40; research
// D5 y D6 de H40): hay una declaración por cada variante versionada de la familia
// que no es la de control, en su orden; cada una de esas tres es, byte a byte, la
// que se calcula desde normativa-control con su declaración, y el cuerpo de su
// SKILL.md remite una vez a cada referencia y a nada más; las referencias de la
// familia son los ficheros de references/ de normativa-control, en orden de
// nombre; y ningún nombre de la familia es el de una variante de H39. Los
// directorios de la familia y normativa-control, que ya no se compara con
// skills/normativa, los comprueba TestVariantesDeControl (FR-020).
func TestVariantesDeNormativa(t *testing.T) {
	t.Parallel()

	repositorio := os.DirFS(raizDelRepositorio)

	t.Run("variantes declaradas", func(t *testing.T) {
		t.Parallel()

		declaradas := make([]string, 0, len(declaracionesDeNormativa))
		for _, declarada := range declaracionesDeNormativa {
			declaradas = append(declaradas, declarada.nombre)
		}

		assert.Equal(t, familiaDeNormativa.nombres()[1:], declaradas,
			"hay una declaración por cada variante de la familia de normativa que no es la de control, en su orden")

		for _, declarada := range declaracionesDeNormativa {
			t.Run(declarada.nombre, func(t *testing.T) {
				t.Parallel()

				comprobarLaVarianteDeclarada(t, repositorio, familiaDeNormativa, declarada)
			})
		}
	})

	t.Run("remisiones de las variantes declaradas", func(t *testing.T) {
		t.Parallel()

		for _, declarada := range declaracionesDeNormativa {
			comprobarLasRemisiones(t, repositorio, declarada.nombre)
		}
	})

	t.Run("referencias de la familia", func(t *testing.T) {
		t.Parallel()

		referencias := path.Join(familiaDeNormativa.directorioDeLaSkill(varianteDeControlDeNormativa),
			directorioDeLasReferencias)

		entradas, err := fs.ReadDir(repositorio, referencias)
		require.NoError(t, err)

		ficheros := make([]string, 0, len(entradas))
		for _, entrada := range entradas {
			ficheros = append(ficheros, entrada.Name())
		}

		assert.Equal(t, ficheros, familiaDeNormativa.referencias,
			"las referencias de la familia de normativa son los ficheros de %s, en orden de nombre", referencias)
	})

	t.Run("nombres que no son de H39", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, varianteDeControlDeNormativa, familiaDeNormativa.nombres()[0],
			"la primera variante de la familia de normativa es la de control")

		declarada, err := leerLaDeclaracionDelRepositorio(raizDelRepositorio)
		require.NoError(t, err)
		require.NotEmpty(t, declarada.Variantes, "la declaración del laboratorio de H39 tiene variantes")

		for _, deH39 := range declarada.Variantes {
			assert.NotContains(t, familiaDeNormativa.nombres(), deH39.Nombre,
				"%s es el nombre de una variante de %s y de una de la familia de normativa", deH39.Nombre,
				path.Join(carpetaDelLaboratorio, ficheroDeLasVariantes))
		}
	})
}

// leerLosFicheros lee los ficheros del árbol bajo ese directorio, por su ruta
// relativa a él. Lo que no es un directorio ni un fichero normal —un enlace
// simbólico, que os.CopyFS copiaría como enlace (research V10)— es un error con
// su ruta.
func leerLosFicheros(arbol fs.FS, directorio string) (map[string][]byte, error) {
	ficheros := map[string][]byte{}

	err := fs.WalkDir(arbol, directorio, func(ruta string, entrada fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entrada.IsDir() {
			return nil
		}

		relativa, dentro := strings.CutPrefix(ruta, directorio+"/")
		if !dentro || !entrada.Type().IsRegular() {
			return fmt.Errorf("%s no es un fichero normal dentro de un directorio", ruta)
		}

		contenido, err := fs.ReadFile(arbol, ruta)
		if err != nil {
			return err
		}

		ficheros[relativa] = contenido

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("los ficheros de %s: %w", directorio, err)
	}

	return ficheros, nil
}

// primeraDiferencia es la primera ruta, en orden, en que los ficheros obtenidos
// no son los esperados, con su motivo: «difiere», si los dos la tienen con bytes
// distintos; «falta», si solo está en los esperados; y «sobra», si solo está en
// los obtenidos. Sin diferencias, las dos son vacías.
func primeraDiferencia(esperados, obtenidos map[string][]byte) (ruta, motivo string) {
	rutas := slices.Sorted(maps.Keys(esperados))
	rutas = append(rutas, slices.Collect(maps.Keys(obtenidos))...)
	slices.Sort(rutas)

	for _, ruta := range slices.Compact(rutas) {
		esperado, loEspera := esperados[ruta]
		obtenido, loTiene := obtenidos[ruta]

		switch {
		case !loTiene:
			return ruta, "falta"
		case !loEspera:
			return ruta, "sobra"
		case !bytes.Equal(esperado, obtenido):
			return ruta, "difiere"
		}
	}

	return "", ""
}

// Lo que separa la declaración de una variante del texto de un SKILL.md
// (contracts/variantes.md §3 y §4 de H40).
const (
	// marcaDeLaVineta es la con la que empieza la línea de una viñeta.
	marcaDeLaVineta = "- "
	// sangriaDeLaVineta es la de las líneas que siguen a la de una viñeta y son
	// suyas.
	sangriaDeLaVineta = "  "
	// inicioDeUnaRemision es el principio de una remisión del cuerpo a una
	// referencia, `](references/<f>.md)`.
	inicioDeUnaRemision = "](" + directorioDeLasReferencias + "/"
	// lineasAntesDeLaRegla son las de la referencia de H39 que preceden al
	// párrafo de su regla: el título y la línea vacía (research V4 de H40).
	lineasAntesDeLaRegla = 2
)

// sustitucionDeLaVariante es una diferencia declarada de una variante sobre un
// fichero de control (data-model §3 de H40): antes está una sola vez en el
// fichero de control; despues son los bytes que lo sustituyen en la variante; e
// hito es el texto que fija el hito, el de despues sin la marca de la viñeta y
// con cada salto de línea y la sangría de la línea que sigue contados como un
// espacio.
type sustitucionDeLaVariante struct {
	// fichero es la ruta del fichero, relativa a la skill de la variante.
	fichero string
	antes   string
	despues string
	hito    string
}

// reglaDelCuerpoCorto es la que da el cuerpo del SKILL.md de
// normativa-cuerpo-corto sobre el de control (contracts/variantes.md §4 de H40).
type reglaDelCuerpoCorto struct {
	// enteras son los encabezados de las secciones que se quedan enteras.
	enteras []string
	// renombrada es el encabezado de la sección de la que se quedan solo sus
	// líneas anteriores a la primera viñeta, la viñeta y las líneas vacías que
	// siguen a la primera viñeta; renombre, cómo se escribe su encabezado.
	renombrada string
	renombre   string
	// vineta es el principio de la viñeta que se queda de la renombrada.
	vineta string
}

// anadidoDeLaVariante es lo que sigue en la variante a un fichero de control
// (contracts/variantes.md §5 de H40): un salto de línea, el encabezado, una línea
// vacía y el párrafo de la regla, leído de la referencia de H39 que lo lleva.
type anadidoDeLaVariante struct {
	// fichero es la ruta del fichero, relativa a la skill de la variante.
	fichero    string
	encabezado string
	// regla es la ruta, desde la raíz del repositorio, de la referencia de cuyo
	// párrafo sale el añadido.
	regla string
}

// declaracionDeLaVariante es lo que se aplica a normativa-control para calcular
// la variante de ese nombre (data-model §3 de H40); lo que no declara es el de
// control.
type declaracionDeLaVariante struct {
	nombre        string
	sustituciones []sustitucionDeLaVariante
	cuerpoCorto   *reglaDelCuerpoCorto
	anadido       *anadidoDeLaVariante
}

// declaracionesDeNormativa son las de las variantes de la familia de normativa
// que no son la de control, en su orden, con los bytes de contracts/variantes.md
// §3 a §5 de H40 (research D5 y D6 de H40).
var declaracionesDeNormativa = []declaracionDeLaVariante{
	{
		nombre: "normativa-remisiones-en-orden",
		sustituciones: []sustitucionDeLaVariante{
			{
				fichero: ficheroDeLaSkillDeLaVariante,
				antes: "Sigue los pasos en orden, cada operación con `kitlegal … --json` desde el `PATH` o, si la " +
					"tienes, con su\nherramienta, tras leer [con_herramientas](references/con_herramientas.md).\n",
				despues: "Antes de la primera orden o llamada, lee [con_herramientas](references/con_herramientas.md). " +
					"Sigue los pasos en orden,\ncada operación con `kitlegal … --json` desde el `PATH` o, si la " +
					"tienes, con su herramienta.\n",
				hito: "Antes de la primera orden o llamada, lee [con_herramientas](references/con_herramientas.md). " +
					"Sigue los pasos en orden, cada operación con `kitlegal … --json` desde el `PATH` o, si la " +
					"tienes, con su herramienta.",
			},
			{
				fichero: ficheroDeLaSkillDeLaVariante,
				antes: "Toma el `BOE-A-…` de la tabla de [normas](references/normas.md); si no está, búscala con\n" +
					"[buscar_la_norma](references/buscar_la_norma.md).\n",
				despues: "Lee [normas](references/normas.md) y toma el `BOE-A-…` de su tabla; si no está, lee\n" +
					"[buscar_la_norma](references/buscar_la_norma.md) y búscala como dice.\n",
				hito: "Lee [normas](references/normas.md) y toma el `BOE-A-…` de su tabla; si no está, lee " +
					"[buscar_la_norma](references/buscar_la_norma.md) y búscala como dice.",
			},
			{
				fichero: ficheroDeLaSkillDeLaVariante,
				antes: "- Sin fecha —con ella, [redaccion_en_una_fecha](references/redaccion_en_una_fecha.md)—, " +
					"lee cada bloque",
				despues: "- Con una fecha en la pregunta, lee [redaccion_en_una_fecha](references/redaccion_en_una_fecha.md) " +
					"antes de leer el\n  bloque. Sin fecha, lee cada bloque",
				hito: "Con una fecha en la pregunta, lee [redaccion_en_una_fecha](references/redaccion_en_una_fecha.md) " +
					"antes de leer el bloque. Sin fecha, lee cada bloque",
			},
			{
				fichero: ficheroDeLaSkillDeLaVariante,
				antes: "- Varios bloques a la vez: [varios_bloques](references/varios_bloques.md). Si una orden no " +
					"termina con `0`, antes de\n  seguir: [lecturas_que_fallan](references/lecturas_que_fallan.md).\n",
				despues: "- Para leer varios bloques a la vez, lee antes [varios_bloques](references/varios_bloques.md). " +
					"Si una orden no\n  termina con `0`, lee [lecturas_que_fallan](references/lecturas_que_fallan.md) " +
					"antes de seguir.\n",
				hito: "Para leer varios bloques a la vez, lee antes [varios_bloques](references/varios_bloques.md). " +
					"Si una orden no termina con `0`, lee [lecturas_que_fallan](references/lecturas_que_fallan.md) " +
					"antes de seguir.",
			},
			{
				fichero: ficheroDeLaSkillDeLaVariante,
				antes:   ";\nrepásalas, y las dudas, en [como_se_cita](references/como_se_cita.md).",
				despues: ".\nAntes de escribir la respuesta, lee [como_se_cita](references/como_se_cita.md) y repasa " +
					"con ella las\ncitas.",
				hito: ". Antes de escribir la respuesta, lee [como_se_cita](references/como_se_cita.md) y repasa con " +
					"ella las citas.",
			},
		},
	},
	{
		nombre: "normativa-cuerpo-corto",
		cuerpoCorto: &reglaDelCuerpoCorto{
			enteras: []string{
				"## Protocolo",
				"### 1. Identificar la norma",
				"### 2. Leer",
				"## Cómo se cita",
				"## Redacción en una fecha",
				"## Redacción modificada",
			},
			renombrada: "### 4. Responder",
			renombre:   "### 3. Responder",
			vineta:     "- Cada afirmación sobre el contenido de una norma lleva su cita",
		},
	},
	{
		nombre: "normativa-testigo-en-referencia",
		sustituciones: []sustitucionDeLaVariante{
			{
				fichero: ficheroDeLaSkillDeLaVariante,
				antes:   "[con_herramientas](references/con_herramientas.md).",
				despues: "[con_herramientas](references/con_herramientas.md), que dice también cómo termina la respuesta.",
				hito:    "[con_herramientas](references/con_herramientas.md), que dice también cómo termina la respuesta.",
			},
		},
		anadido: &anadidoDeLaVariante{
			fichero:    path.Join(directorioDeLasReferencias, "con_herramientas.md"),
			encabezado: "## Cómo termina la respuesta",
			regla: path.Join(carpetaDelLaboratorio, carpetaDeLasVariantes, "referencia-paso",
				carpetaDeLasSkillsDeLaVariante, skillDelLaboratorio, directorioDeLasReferencias,
				referenciaDelLaboratorio),
		},
	},
}

// comprobarLaVarianteDeclarada: el «después» de cada sustitución de la
// declaración, normalizado, es el texto del hito; la declaración se aplica sobre
// la variante de control de la familia, la primera; y el directorio de la
// variante es, ruta a ruta y byte a byte, la que se calcula así, sin nada más
// (FR-003 a FR-006 de H40; FR-021 y FR-023); falla con la sustitución cuyo
// «después» no es el hito, con lo que ya no se aplica sobre control, o con la
// ruta del primer fichero, en orden, que difiere, falta o sobra.
func comprobarLaVarianteDeclarada(t *testing.T, repositorio fs.FS, familia familiaDelLaboratorio,
	declarada declaracionDeLaVariante,
) {
	t.Helper()

	control := familia.nombres()[0]

	for indice, sustitucion := range declarada.sustituciones {
		assert.Equal(t, sustitucion.hito, textoDelHito(sustitucion.despues),
			"%s: el «después» de la sustitución %d de %s, sin la marca de la viñeta y con cada salto de línea y "+
				"su sangría como un espacio, no es el texto del hito", declarada.nombre, indice+1, sustitucion.fichero)
	}

	esperada, err := calcularLaVariante(repositorio, familia, declarada)
	require.NoError(t, err, "%s: su declaración ya no se aplica sobre %s", declarada.nombre, control)

	esperados := make(map[string][]byte, len(esperada))
	for relativa, contenido := range esperada {
		esperados[path.Join(carpetaDeLasSkillsDeLaVariante, familia.skill, relativa)] = contenido
	}

	directorio := familia.directorioDeLaVariante(declarada.nombre)

	obtenidos, err := leerLosFicheros(repositorio, directorio)
	require.NoError(t, err)

	relativa, motivo := primeraDiferencia(esperados, obtenidos)
	assert.Empty(t, motivo, "%s %s: la variante es la que da su declaración sobre %s, ruta a ruta y byte a byte",
		path.Join(directorio, relativa), motivo, control)
}

// comprobarLasRemisiones: el cuerpo del SKILL.md de la variante de ese nombre
// remite una vez a cada referencia de la familia y a nada más
// (contracts/variantes.md §7 de H40).
func comprobarLasRemisiones(t *testing.T, repositorio fs.FS, nombre string) {
	t.Helper()

	ruta := path.Join(familiaDeNormativa.directorioDeLaSkill(nombre), ficheroDeLaSkillDeLaVariante)

	contenido, err := fs.ReadFile(repositorio, ruta)
	require.NoError(t, err)

	_, cuerpo, err := frontmatterYCuerpo(string(contenido))
	require.NoError(t, err, ruta)

	assert.Equal(t, len(familiaDeNormativa.referencias), strings.Count(cuerpo, inicioDeUnaRemision),
		"%s: el cuerpo tiene una remisión a cada referencia de la familia y ninguna más", ruta)

	for _, referencia := range familiaDeNormativa.referencias {
		assert.Equal(t, 1, strings.Count(cuerpo, inicioDeUnaRemision+referencia+")"),
			"%s: el cuerpo tiene una remisión a %s", ruta, referencia)
	}
}

// textoDelHito es el texto de un «después» como lo fija el hito (FR-003 de H40):
// sin la marca de la viñeta con la que empieza y con cada salto de línea, y la
// sangría de la línea que lo sigue, como un espacio. El salto con el que termina
// el «después», que no sigue ninguna línea, no cuenta.
func textoDelHito(despues string) string {
	lineas := strings.Split(strings.TrimSuffix(strings.TrimPrefix(despues, marcaDeLaVineta), "\n"), "\n")
	for indice := 1; indice < len(lineas); indice++ {
		lineas[indice] = strings.TrimLeft(lineas[indice], " ")
	}

	return strings.Join(lineas, " ")
}

// calcularLaVariante son los ficheros de la variante que da la declaración sobre
// los de la variante de control de la familia, la primera, por su ruta relativa a
// la skill (data-model §3 de H40; data-model §6): sus sustituciones, fichero a
// fichero; la regla del cuerpo corto, en el SKILL.md; y el añadido, al final de
// su fichero. Lo que no declara es el de control. Lo que ya no se aplica sobre
// control es un error con su motivo.
func calcularLaVariante(repositorio fs.FS, familia familiaDelLaboratorio,
	declarada declaracionDeLaVariante,
) (map[string][]byte, error) {
	variante, err := leerLosFicheros(repositorio, familia.directorioDeLaSkill(familia.nombres()[0]))
	if err != nil {
		return nil, err
	}

	porFichero := map[string][]sustitucionDeLaVariante{}
	for _, sustitucion := range declarada.sustituciones {
		porFichero[sustitucion.fichero] = append(porFichero[sustitucion.fichero], sustitucion)
	}

	for _, fichero := range slices.Sorted(maps.Keys(porFichero)) {
		sustituido, err := aplicarLasSustituciones(string(variante[fichero]), porFichero[fichero])
		if err != nil {
			return nil, fmt.Errorf("%s: %w", fichero, err)
		}

		variante[fichero] = []byte(sustituido)
	}

	if regla := declarada.cuerpoCorto; regla != nil {
		corto, err := aplicarElCuerpoCorto(string(variante[ficheroDeLaSkillDeLaVariante]), *regla)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", ficheroDeLaSkillDeLaVariante, err)
		}

		variante[ficheroDeLaSkillDeLaVariante] = []byte(corto)
	}

	if anadido := declarada.anadido; anadido != nil {
		parrafo, err := parrafoDelAnadido(repositorio, anadido.regla)
		if err != nil {
			return nil, err
		}

		variante[anadido.fichero] = []byte(string(variante[anadido.fichero]) + "\n" + anadido.encabezado + "\n\n" +
			parrafo)
	}

	return variante, nil
}

// aplicarLasSustituciones es el contenido con cada «antes» sustituido por su
// «después»; ninguna toca un byte fuera de su «antes». Un «antes» que no está una
// sola vez en el contenido, o dos que se solapan, es un error con la sustitución,
// por su número en el fichero, y cuántas veces está.
func aplicarLasSustituciones(contenido string, sustituciones []sustitucionDeLaVariante) (string, error) {
	type tramo struct {
		numero, inicio, fin int
		despues             string
	}

	tramos := make([]tramo, 0, len(sustituciones))

	for indice, sustitucion := range sustituciones {
		if veces := strings.Count(contenido, sustitucion.antes); veces != 1 {
			return "", fmt.Errorf("la sustitución %d ya no se aplica: su «antes» %q está %d veces, y tiene que estar una",
				indice+1, sustitucion.antes, veces)
		}

		inicio := strings.Index(contenido, sustitucion.antes)
		tramos = append(tramos, tramo{
			numero: indice + 1, inicio: inicio, fin: inicio + len(sustitucion.antes), despues: sustitucion.despues,
		})
	}

	slices.SortFunc(tramos, func(a, b tramo) int { return a.inicio - b.inicio })

	var sustituido strings.Builder

	hecho := 0

	for _, sustitucion := range tramos {
		if sustitucion.inicio < hecho {
			return "", fmt.Errorf("la sustitución %d ya no se aplica: su «antes» se solapa con el de otra",
				sustitucion.numero)
		}

		sustituido.WriteString(contenido[hecho:sustitucion.inicio])
		sustituido.WriteString(sustitucion.despues)
		hecho = sustitucion.fin
	}

	sustituido.WriteString(contenido[hecho:])

	return sustituido.String(), nil
}

// seccionDelCuerpo es un encabezado del cuerpo de un SKILL.md —una línea que
// empieza por «## » o por «### »— con las líneas que lo siguen hasta el
// siguiente; la primera del cuerpo es el preámbulo, sin encabezado
// (contracts/variantes.md §4 de H40).
type seccionDelCuerpo struct {
	// encabezado es la línea del encabezado, sin su salto.
	encabezado string
	// lineas son las de la sección, con su salto, empezando por la del
	// encabezado.
	lineas []string
}

// seccionesDelCuerpo son las secciones del cuerpo, en su orden, empezando por el
// preámbulo.
func seccionesDelCuerpo(cuerpo string) []seccionDelCuerpo {
	secciones := []seccionDelCuerpo{{}}

	for linea := range strings.Lines(cuerpo) {
		if strings.HasPrefix(linea, "## ") || strings.HasPrefix(linea, "### ") {
			secciones = append(secciones, seccionDelCuerpo{encabezado: strings.TrimSuffix(linea, "\n")})
		}

		ultima := &secciones[len(secciones)-1]
		ultima.lineas = append(ultima.lineas, linea)
	}

	return secciones
}

// aplicarElCuerpoCorto es el SKILL.md con el mismo frontmatter y, de su cuerpo,
// en su orden, el preámbulo, las secciones enteras de la regla y lo que se queda
// de la renombrada, hasta su última línea no vacía, con su salto
// (contracts/variantes.md §4 de H40). Un encabezado de la regla que no está una
// sola vez en el cuerpo es un error con el encabezado y cuántas veces está.
func aplicarElCuerpoCorto(contenido string, regla reglaDelCuerpoCorto) (string, error) {
	frontmatter, cuerpo, err := frontmatterYCuerpo(contenido)
	if err != nil {
		return "", err
	}

	secciones := seccionesDelCuerpo(cuerpo)

	for _, encabezado := range append(slices.Clone(regla.enteras), regla.renombrada) {
		veces := 0

		for _, seccion := range secciones[1:] {
			if seccion.encabezado == encabezado {
				veces++
			}
		}

		if veces != 1 {
			return "", fmt.Errorf("la regla del cuerpo corto ya no se aplica: el encabezado %q está %d veces en "+
				"el cuerpo, y tiene que estar una", encabezado, veces)
		}
	}

	corto := strings.Join(secciones[0].lineas, "")

	for _, seccion := range secciones[1:] {
		switch {
		case slices.Contains(regla.enteras, seccion.encabezado):
			corto += strings.Join(seccion.lineas, "")
		case seccion.encabezado == regla.renombrada:
			queda, err := loQueQuedaDeLaRenombrada(seccion, regla)
			if err != nil {
				return "", err
			}

			corto += queda
		}
	}

	return frontmatter + strings.TrimRight(corto, "\n") + "\n", nil
}

// loQueQuedaDeLaRenombrada es lo que se queda de la sección renombrada: su
// encabezado como lo escribe la regla, sus líneas anteriores a la primera viñeta,
// la viñeta de la regla —su línea y las que la siguen con la sangría de una
// viñeta— y las líneas vacías que siguen a la primera viñeta. Una viñeta de la
// regla que no está una sola vez en la sección es un error con la viñeta y
// cuántas veces está.
func loQueQuedaDeLaRenombrada(seccion seccionDelCuerpo, regla reglaDelCuerpoCorto) (string, error) {
	queda := regla.renombre + "\n"
	trasLaPrimera, enLaDeLaRegla, veces := false, false, 0

	for _, linea := range seccion.lineas[1:] {
		switch {
		case strings.HasPrefix(linea, marcaDeLaVineta):
			trasLaPrimera = true
			enLaDeLaRegla = strings.HasPrefix(linea, regla.vineta)

			if enLaDeLaRegla {
				veces++
			}
		case !strings.HasPrefix(linea, sangriaDeLaVineta):
			enLaDeLaRegla = false
		}

		if !trasLaPrimera || enLaDeLaRegla || linea == "\n" {
			queda += linea
		}
	}

	if veces != 1 {
		return "", fmt.Errorf("la regla del cuerpo corto ya no se aplica: la viñeta %q está %d veces en la "+
			"sección %q, y tiene que estar una", regla.vineta, veces, seccion.encabezado)
	}

	return queda, nil
}

// parrafoDelAnadido es el párrafo de la regla testigo de la referencia de H39 de
// esa ruta: lo que sigue a sus dos primeras líneas, una sola línea que termina en
// la del sello, con su salto (contracts/variantes.md §5 de H40; research V4 de
// H40). Si no lo es, el añadido ya no se aplica.
func parrafoDelAnadido(repositorio fs.FS, ruta string) (string, error) {
	contenido, err := fs.ReadFile(repositorio, ruta)
	if err != nil {
		return "", fmt.Errorf("el párrafo del añadido: %w", err)
	}

	lineas := slices.Collect(strings.Lines(string(contenido)))
	if len(lineas) != lineasAntesDeLaRegla+1 || !strings.HasSuffix(lineas[lineasAntesDeLaRegla], lineaDelSello+"\n") {
		return "", fmt.Errorf("el añadido ya no se aplica: lo que sigue a las %d primeras líneas de %s no es una "+
			"línea que termina en %q, con su salto", lineasAntesDeLaRegla, ruta, lineaDelSello)
	}

	return lineas[lineasAntesDeLaRegla], nil
}

// frontmatterYCuerpo parte un SKILL.md en su frontmatter —desde la primera línea,
// la del delimitador, hasta la que lo cierra, las dos incluidas— y su cuerpo, las
// líneas que siguen (contracts/variantes.md §1 de H40).
func frontmatterYCuerpo(contenido string) (string, string, error) {
	delimitador := delimitadorDelFrontmatterDeLaVariante + "\n"

	resto, abre := strings.CutPrefix(contenido, delimitador)
	if !abre {
		return "", "", fmt.Errorf("el SKILL.md no empieza por la línea %q", delimitadorDelFrontmatterDeLaVariante)
	}

	fin := strings.Index(resto, "\n"+delimitador)
	if fin < 0 {
		return "", "", fmt.Errorf("el SKILL.md no cierra su frontmatter con la línea %q",
			delimitadorDelFrontmatterDeLaVariante)
	}

	corte := len(delimitador) + fin + 1 + len(delimitador)

	return contenido[:corte], contenido[corte:], nil
}

// respuestasDeLaReglaTestigoDeNormativa es el directorio de las respuestas fijas
// a las evals del sondeo de la familia de normativa con las que
// TestReglaTestigoDeNormativa comprueba el guion de la regla testigo, cada una
// con su veredicto en el nombre como en H39 (contracts/tabla.md §6 de H40).
const respuestasDeLaReglaTestigoDeNormativa = "testdata/laboratorio/normativa/respuestas"

// nombresDeLasRespuestasDeNormativa son las respuestas fijas de la regla testigo
// a las evals del sondeo de normativa, cada una a la eval que dice
// contracts/tabla.md §6 de H40: la que la cumple, a la 04, y una por cada forma
// de no cumplirla que enumera FR-030 —sin la línea del sello, a la 05; la línea
// en negrita, a la 15; entre comillas, a la 22; en código, a la 04; y antes del
// final, a la 05— (research D11 de H40).
var nombresDeLasRespuestasDeNormativa = []string{
	"cumple.txt",
	"no-cumple-sin-sello.txt",
	"no-cumple-sello-en-negrita.txt",
	"no-cumple-sello-entre-comillas.txt",
	"no-cumple-sello-en-codigo.txt",
	"no-cumple-sello-no-al-final.txt",
}

// evalDeLaSesionDeNormativaDeH32 es la eval de la respuesta de la skill
// normativa de kitlegal con la que se comprueba la regla de su familia: la 05,
// la del artículo 59 del TRLRHL, en la primera sesión de Sonnet en el modo orden
// del informe de evals de normativa que versionó H32, el mismo que lee
// TestReglaTestigo (research V18 de H40). Es la de control en el sentido del
// spec: lo que responde normativa sin la regla.
const evalDeLaSesionDeNormativaDeH32 = "05-trlrhl-impuestos-municipales.yaml"

// TestReglaTestigoDeNormativa comprueba el guion de la regla testigo de H39, sin
// cambios, con respuestas a las evals del sondeo de normativa (FR-030 y SC-004
// de H40; contracts/tabla.md §6 y §7 de H40; research D11 y V18 de H40): el
// directorio de las respuestas fijas tiene exactamente sus seis respuestas, el
// guion da a cada una el veredicto de su nombre, y la respuesta de la skill
// normativa de kitlegal a la eval 05, leída del informe versionado de H32, no la
// cumple.
func TestReglaTestigoDeNormativa(t *testing.T) {
	t.Parallel()

	t.Run("respuestas fijas", comprobarLasRespuestasDeNormativa)
	t.Run("respuesta de normativa a la eval 05", comprobarLaRespuestaDeNormativaDeH32)
}

// comprobarLasRespuestasDeNormativa: el directorio tiene exactamente las seis
// respuestas fijas de normativa, y el guion da a cada una, en su subtest, el
// veredicto de su nombre.
func comprobarLasRespuestasDeNormativa(t *testing.T) {
	t.Parallel()

	entradas, err := os.ReadDir(respuestasDeLaReglaTestigoDeNormativa)
	require.NoError(t, err, "el directorio %s de las respuestas fijas se lee", respuestasDeLaReglaTestigoDeNormativa)

	nombres := make([]string, len(entradas))
	for i, entrada := range entradas {
		nombres[i] = entrada.Name()
	}

	require.ElementsMatch(t, nombresDeLasRespuestasDeNormativa, nombres,
		"%s tiene las respuestas fijas de la regla testigo y ninguna más", respuestasDeLaReglaTestigoDeNormativa)

	for _, nombre := range nombresDeLasRespuestasDeNormativa {
		t.Run(nombre, func(t *testing.T) {
			t.Parallel()

			respuesta := string(contenidoDelFichero(t, path.Join(respuestasDeLaReglaTestigoDeNormativa, nombre)))
			require.NotEmpty(t, strings.TrimSpace(respuesta), "la respuesta fija %s tiene texto", nombre)

			assert.Equal(t, veredictoDelNombre(t, nombre), reglaTestigo(respuesta),
				"el guion da a %s el veredicto de su nombre", nombre)
		})
	}
}

// comprobarLaRespuestaDeNormativaDeH32: la respuesta de la sesión de H32 a la
// eval 05, leída del informe versionado, es la de una sesión de esa eval que
// activó la skill normativa y tiene respuesta, y no cumple la regla.
func comprobarLaRespuestaDeNormativaDeH32(t *testing.T) {
	t.Parallel()

	nombre := nombreDeSesion(evalDeLaSesionDeNormativaDeH32, modeloDeLaSesionDeH32, ModoOrden,
		numeroDeLaSesionDeH32, false)
	sesion := sesionDelInformeVersionado(t, informeDeNormativaDeH32, nombre)

	require.Equal(t, evalDeLaSesionDeNormativaDeH32, sesion.Eval, "la sesión %s es de la eval 05", nombre)
	require.True(t, sesion.Activada, "la sesión %s activó la skill normativa", nombre)
	require.NotEmpty(t, strings.TrimSpace(sesion.Respuesta), "la sesión %s tiene respuesta", nombre)

	assert.False(t, reglaTestigo(sesion.Respuesta),
		"la respuesta de la skill normativa en la sesión %s no cumple la regla testigo", nombre)
}

// sesionDelInformeVersionado es la sesión del nombre dado en el informe
// versionado de la ruta, que la tiene una vez y solo una.
func sesionDelInformeVersionado(t *testing.T, ruta, nombre string) entradaDelInforme {
	t.Helper()

	var informe struct {
		Evals []entradaDelInforme `json:"evals"`
	}
	require.NoError(t, json.Unmarshal(contenidoDelFichero(t, ruta), &informe), "%s es un informe del job de evals", ruta)

	sesiones := slices.DeleteFunc(informe.Evals, func(entrada entradaDelInforme) bool {
		return entrada.Sesion != nombre
	})
	require.Len(t, sesiones, 1, "el informe %s tiene una sesión %s", ruta, nombre)

	return sesiones[0]
}
