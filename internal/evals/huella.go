package evals

import (
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"slices"
	"strings"
	"unicode"
)

// La huella de una eval (FR-009; data-model §6) es el sha256 en hexadecimal de
// un texto canónico, en UTF-8, con una línea por cosa que la elige, terminada
// en salto de línea; dentro de cada bloque, las líneas ordenadas por bytes; los
// bloques, en el orden de estos campos (contracts/eleccion-y-huella.md §4;
// research D6 y D7). No lleva fechas, permisos, rutas absolutas ni el commit:
// da lo mismo en cualquier directorio.
const (
	// cabeceraDeLaHuella es la primera línea del texto, con la versión de su
	// forma.
	cabeceraDeLaHuella = "huella-de-eval v1"

	// campoModeloQueDecide y campoClaudeCode son el modelo que decide y la
	// versión de Claude Code de las sesiones.
	campoModeloQueDecide = "modelo_que_decide"
	campoClaudeCode      = "claude_code"

	// campoJuez es el modelo del juez y la versión de Claude Code de sus votos,
	// solo en las evals de una skill con juez; la rúbrica, los casos y la
	// medida entran como rutas de su conjunto, parte (f).
	campoJuez = "juez"

	// campoGo y campoToolchain son los de go.mod.
	campoGo        = "go"
	campoToolchain = "toolchain"

	// campoModulo es cada módulo del conjunto y de lo común, con su versión.
	campoModulo = "modulo"

	// campoRestoDeGoMod y campoRestoDeGoSum son el sha256 de lo demás de go.mod
	// y de go.sum, iguales en todas las evals (research D7).
	campoRestoDeGoMod = "go_mod_resto"
	campoRestoDeGoSum = "go_sum_resto"

	// campoDescripcion es el sha256 del description de cada skill (FR-003).
	campoDescripcion = "descripcion"

	// campoRuta es cada fichero que mide del conjunto, de lo común o sin
	// asignar, con el sha256 de su contenido.
	campoRuta = "ruta"

	nombreDelGoSum = "go.sum"
)

// Los campos de go mod edit -json que no entran en el resto de go.mod: Go y
// Toolchain tienen sus líneas, y de Require salen los módulos de algún conjunto
// o de lo común, que tienen las suyas (research D7).
const (
	campoGoDeGoMod        = "Go"
	campoToolchainDeGoMod = "Toolchain"
	campoRequireDeGoMod   = "Require"
)

// fijadosDeLaHuella son los modelos y las versiones que el job fija en
// evals.yml y llegan por banderas (contracts/eleccion-y-huella.md §1).
type fijadosDeLaHuella struct {
	modeloQueDecide   string
	claudeCode        string
	modeloDelJuez     string
	claudeCodeDelJuez string
}

// validar dice por qué los fijados no pueden ir en el texto de la huella: uno
// vacío o con espacios lo haría ambiguo, y dos fijados distintos darían la
// misma huella.
func (f fijadosDeLaHuella) validar() error {
	fijados := []struct{ nombre, valor string }{
		{"el modelo que decide", f.modeloQueDecide},
		{"la versión de Claude Code de las sesiones", f.claudeCode},
		{"el modelo del juez", f.modeloDelJuez},
		{"la versión de Claude Code del juez", f.claudeCodeDelJuez},
	}

	var errores []error

	for _, fijado := range fijados {
		if fijado.valor == "" || strings.ContainsFunc(fijado.valor, unicode.IsSpace) {
			errores = append(errores,
				fmt.Errorf("%s no puede ir vacío ni con espacios en la huella: %q", fijado.nombre, fijado.valor))
		}
	}

	return errors.Join(errores...)
}

// restoDeGo es lo que la huella lee de go.mod y de go.sum, además de go,
// toolchain y los requeridos, que están en el árbol (research D7).
type restoDeGo struct {
	// goMod son los campos de go mod edit -json, sin Go, Toolchain ni Require.
	goMod map[string]jsontext.Value

	// requeridos son las entradas de Require, en su orden.
	requeridos []requeridoDeGoMod

	// lineasDeGoSum son las líneas de go.sum que no están vacías, en su orden.
	lineasDeGoSum []string
}

// requeridoDeGoMod es una entrada de Require de go mod edit -json: su módulo
// y, tal cual, sus demás campos.
type requeridoDeGoMod struct {
	Path  string
	Resto jsontext.Value `json:",unknown"`
}

// leerRestoDeGo lee de la raíz go.mod, con go mod edit -json, y go.sum, que
// puede no estar: un módulo sin dependencias no lo tiene. Es un error que no
// se puedan leer.
func leerRestoDeGo(ctx context.Context, raiz string) (restoDeGo, error) {
	disco := os.DirFS(raiz)

	goMod, err := fs.ReadFile(disco, nombreDelGoMod)
	if err != nil {
		return restoDeGo{}, fmt.Errorf("%s no se puede leer: %w", nombreDelGoMod, err)
	}

	enJSON, err := goModEnJSON(ctx, raiz, goMod)
	if err != nil {
		return restoDeGo{}, err
	}

	resto := restoDeGo{}
	if err := json.Unmarshal(enJSON, &resto.goMod); err != nil {
		return restoDeGo{}, fmt.Errorf("la salida de go mod edit -json no se puede leer: %w", err)
	}

	if requeridos, estan := resto.goMod[campoRequireDeGoMod]; estan {
		if err := json.Unmarshal(requeridos, &resto.requeridos); err != nil {
			return restoDeGo{}, fmt.Errorf("los requeridos de go mod edit -json no se pueden leer: %w", err)
		}
	}

	for _, campo := range []string{campoGoDeGoMod, campoToolchainDeGoMod, campoRequireDeGoMod} {
		delete(resto.goMod, campo)
	}

	goSum, err := fs.ReadFile(disco, nombreDelGoSum)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return restoDeGo{}, fmt.Errorf("%s no se puede leer: %w", nombreDelGoSum, err)
	}

	for linea := range strings.Lines(string(goSum)) {
		if linea = strings.TrimSpace(linea); linea != "" {
			resto.lineasDeGoSum = append(resto.lineasDeGoSum, linea)
		}
	}

	return resto, nil
}

// sumaDelGoMod es el sha256 del JSON de go mod edit -json sin Go, Toolchain
// ni las entradas de Require de los módulos dados, en la forma canónica de RFC
// 8785, que no depende del orden de sus campos ni de sus espacios.
func (r restoDeGo) sumaDelGoMod(modulos map[string]bool) (string, error) {
	campos := maps.Clone(r.goMod)
	if campos == nil {
		campos = map[string]jsontext.Value{}
	}

	quedan := slices.DeleteFunc(slices.Clone(r.requeridos), func(requerido requeridoDeGoMod) bool {
		return modulos[requerido.Path]
	})

	if len(quedan) > 0 {
		requeridos, err := json.Marshal(quedan)
		if err != nil {
			return "", fmt.Errorf("los requeridos que quedan de %s: %w", nombreDelGoMod, err)
		}

		campos[campoRequireDeGoMod] = requeridos
	}

	resto, err := json.Marshal(campos)
	if err != nil {
		return "", fmt.Errorf("el resto de %s: %w", nombreDelGoMod, err)
	}

	canonico := jsontext.Value(resto)
	if err := canonico.Canonicalize(); err != nil {
		return "", fmt.Errorf("el resto de %s en su forma canónica: %w", nombreDelGoMod, err)
	}

	return huellaSHA256(canonico), nil
}

// sumaDelGoSum es el sha256 de las líneas de go.sum cuyo módulo, el primer
// campo, no es de los dados, ordenadas por bytes y cada una terminada en salto
// de línea.
func (r restoDeGo) sumaDelGoSum(modulos map[string]bool) string {
	var lineas []string

	for _, linea := range r.lineasDeGoSum {
		if modulo, _, _ := strings.Cut(linea, " "); !modulos[modulo] {
			lineas = append(lineas, linea+"\n")
		}
	}

	slices.Sort(lineas)

	return huellaSHA256([]byte(strings.Join(lineas, "")))
}

// calculoDeLaHuella es lo que lee la huella de las evals de un árbol: los
// fijados, el conjunto de cada eval y lo común, calculados sobre la cabeza, el
// árbol de la cabeza y el resto de su go.mod y su go.sum.
type calculoDeLaHuella struct {
	fijados   fijadosDeLaHuella
	conjuntos []dependenciasDeLaEval
	comun     loComun
	cabeza    arbolDelRepositorio
	resto     restoDeGo
}

// leerElCalculoDeLaHuella lee de la raíz, con su grafo de paquetes, lo que la
// huella necesita: el conjunto de cada eval, el árbol de la cabeza con las
// rutas que miden, lo común y el resto de go.mod y go.sum. Un conjunto que no
// se puede calcular no es un error aquí: lo es su huella.
func leerElCalculoDeLaHuella(ctx context.Context, raiz string, grafo grafoDePaquetes, fijados fijadosDeLaHuella) (
	calculoDeLaHuella, error,
) {
	conjuntos, err := calcularDependencias(os.DirFS(raiz), grafo)
	if err != nil {
		return calculoDeLaHuella{}, err
	}

	cabeza, err := leerArbolDeLaCabeza(ctx, raiz, paquetesDeLosConjuntos(conjuntos))
	if err != nil {
		return calculoDeLaHuella{}, err
	}

	comun, err := grafo.loComun(cabeza.modulos)
	if err != nil {
		return calculoDeLaHuella{}, err
	}

	resto, err := leerRestoDeGo(ctx, raiz)
	if err != nil {
		return calculoDeLaHuella{}, err
	}

	return calculoDeLaHuella{fijados: fijados, conjuntos: conjuntos, comun: comun, cabeza: cabeza, resto: resto}, nil
}

// huellas son las de las evals, una por conjunto y en su orden: el sha256 de
// su texto en 64 caracteres hexadecimales. No hay huellas si alguna eval no
// tiene conjunto o un fijado no puede ir en el texto (textos).
func (c calculoDeLaHuella) huellas() ([]string, error) {
	textos, err := c.textos()
	if err != nil {
		return nil, err
	}

	huellas := make([]string, len(textos))
	for indice, texto := range textos {
		huellas[indice] = huellaSHA256([]byte(texto))
	}

	return huellas, nil
}

// textos son los textos canónicos de las huellas de las evals, uno por
// conjunto y en su orden. Es un error que un fijado esté vacío o lleve
// espacios, o que el conjunto de alguna eval no se pueda calcular: sin él no
// se sabe qué rutas son de esa eval ni cuáles no son de ningún conjunto, y con
// sin_conjunto ninguna lleva huella (contracts/informe-del-job.md §2). El
// error nombra cada eval sin conjunto.
func (c calculoDeLaHuella) textos() ([]string, error) {
	if err := c.fijados.validar(); err != nil {
		return nil, err
	}

	var errores []error

	for _, conjunto := range c.conjuntos {
		if conjunto.err != nil {
			errores = append(errores, conjunto.err)
		}
	}

	if err := errors.Join(errores...); err != nil {
		return nil, fmt.Errorf("sin el conjunto de cada eval no hay huella: %w", err)
	}

	deAlguno := c.modulosDeAlgunConjunto()

	sumaDelGoMod, err := c.resto.sumaDelGoMod(deAlguno)
	if err != nil {
		return nil, err
	}

	restos := []string{lineaDe(campoRestoDeGoMod, sumaDelGoMod), lineaDe(campoRestoDeGoSum, c.resto.sumaDelGoSum(deAlguno))}
	descripciones := c.lineasDeLasDescripciones()
	deTodas, porConjunto := c.lineasDeLasRutas()

	textos := make([]string, len(c.conjuntos))

	for indice, conjunto := range c.conjuntos {
		var juez []string
		if tieneJuez(conjunto) {
			juez = append(juez, lineaDe(campoJuez, c.fijados.modeloDelJuez, c.fijados.claudeCodeDelJuez))
		}

		textos[indice] = textoDeLaHuella(
			[]string{cabeceraDeLaHuella},
			[]string{lineaDe(campoModeloQueDecide, c.fijados.modeloQueDecide)},
			[]string{lineaDe(campoClaudeCode, c.fijados.claudeCode)},
			juez,
			[]string{lineaDe(campoGo, c.cabeza.versionDeGo)},
			[]string{lineaDe(campoToolchain, c.cabeza.toolchain)},
			c.lineasDeLosModulos(conjunto),
			restos,
			descripciones,
			slices.Concat(deTodas, porConjunto[indice]),
		)
	}

	return textos, nil
}

// textoDeLaHuella es el texto de los bloques, en su orden, con las líneas de
// cada uno ordenadas por bytes y terminadas en salto de línea.
func textoDeLaHuella(bloques ...[]string) string {
	var texto strings.Builder

	for _, bloque := range bloques {
		for _, linea := range slices.Sorted(slices.Values(bloque)) {
			texto.WriteString(linea)
			texto.WriteByte('\n')
		}
	}

	return texto.String()
}

// lineaDe es la línea de un campo con sus valores, separados por un espacio.
func lineaDe(campo string, valores ...string) string {
	return strings.Join(slices.Concat([]string{campo}, valores), " ")
}

// tieneJuez dice si la skill de la eval tiene juez: si su conjunto lleva la
// carpeta evals/<skill>/juez/ por la parte (f).
func tieneJuez(conjunto dependenciasDeLaEval) bool {
	return slices.Contains(conjunto.entradas, entradaDeDependencia{
		clase: entradaDirectorio,
		ruta:  directorioDeLasEvals + "/" + conjunto.skill + "/" + carpetaDelJuez + "/",
		parte: parteDelJuez,
	})
}

// modulosDeAlgunConjunto son los módulos de algún conjunto o de lo común: los
// que tienen su línea en la huella y no entran en los restos.
func (c calculoDeLaHuella) modulosDeAlgunConjunto() map[string]bool {
	modulos := map[string]bool{}

	for modulo := range c.comun.modulos {
		modulos[modulo] = true
	}

	for _, conjunto := range c.conjuntos {
		for modulo := range conjunto.modulos {
			modulos[modulo] = true
		}
	}

	return modulos
}

// lineasDeLosModulos son las de los módulos del conjunto y de lo común, cada
// uno una vez, con su versión en el go.mod de la cabeza, que es la que deja de
// estar en el resto (research D7), o con la que da go list si go.mod no lo
// requiere.
func (c calculoDeLaHuella) lineasDeLosModulos(conjunto dependenciasDeLaEval) []string {
	versiones := map[string]string{}
	maps.Copy(versiones, c.comun.modulos)
	maps.Copy(versiones, conjunto.modulos)

	lineas := make([]string, 0, len(versiones))

	for modulo, version := range versiones {
		if requerida, requerido := c.cabeza.modulos[modulo]; requerido {
			version = requerida
		}

		lineas = append(lineas, lineaDe(campoModulo, modulo, version))
	}

	return lineas
}

// lineasDeLasDescripciones son las de la descripción de cada skill de la
// cabeza, con el sha256 de su description.
func (c calculoDeLaHuella) lineasDeLasDescripciones() []string {
	lineas := make([]string, 0, len(c.cabeza.descripciones))
	for skill, descripcion := range c.cabeza.descripciones {
		lineas = append(lineas, lineaDe(campoDescripcion, skill, huellaSHA256([]byte(descripcion))))
	}

	return lineas
}

// lineasDeLasRutas son las líneas ruta de los ficheros que miden de la
// cabeza, con el sha256 de su contenido: las de todas las evals, de los que
// son de lo común o de ningún conjunto (sin asignar), y, por conjunto, las de
// los demás que son suyos. go.mod y go.sum no tienen línea: entran por sus
// restos y sus versiones.
func (c calculoDeLaHuella) lineasDeLasRutas() ([]string, [][]string) {
	var deTodas []string

	porConjunto := make([][]string, len(c.conjuntos))

	for ruta, suma := range c.cabeza.ficheros {
		if slices.Contains(ficherosQueMiden, ruta) {
			continue
		}

		lineaDeLaRuta := lineaDe(campoRuta, ruta, suma)

		var suyos []int

		for indice, conjunto := range c.conjuntos {
			if conjunto.pertenece(ruta) {
				suyos = append(suyos, indice)
			}
		}

		if esDeLoComun(ruta) || len(suyos) == 0 {
			deTodas = append(deTodas, lineaDeLaRuta)

			continue
		}

		for _, indice := range suyos {
			porConjunto[indice] = append(porConjunto[indice], lineaDeLaRuta)
		}
	}

	return deTodas, porConjunto
}
