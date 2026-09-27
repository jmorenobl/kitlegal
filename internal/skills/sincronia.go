package skills

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
)

// Lo que forma la parte generada de una skill, además de la región de SKILL.md
// (data-model §1 y §4.2), y lo que una skill no lleva.
const (
	// carpetaDeLasReferencias es el directorio de las referencias generadas de
	// una skill, dentro del suyo.
	carpetaDeLasReferencias = "references"

	// extensionDeLasReferencias es la extensión de una referencia generada:
	// references/<nombre>.md sale del YAML de datos que la tabla de generadores
	// declara para <nombre>.
	extensionDeLasReferencias = ".md"

	// maximoDeLineas es el número de líneas de SKILL.md a partir del cual es un
	// defecto (FR-041: «300 líneas o más» falla).
	maximoDeLineas = 299

	// carpetaDeLosScripts es la entrada que una skill no lleva: la de los
	// enlaces al binario de la instalación anterior, que se retiró cuando las
	// skills pasaron a invocar kitlegal desde el PATH (ADR 0019; FR-082 de H19).
	carpetaDeLosScripts = "scripts"

	// defectoDeScripts es el defecto de una skill que la lleva
	// (contracts/skills-e-invocacion.md §3 de H19).
	defectoDeScripts = "una skill no lleva " + carpetaDeLosScripts + "/ (ADR 0019)"
)

// Los permisos con que Escribir crea lo que falta: un directorio, sin escritura
// para otros, y un fichero, solo para quien lo escribe. Un fichero que ya existe
// conserva los suyos.
const (
	permisosDeDirectorio fs.FileMode = 0o750
	permisosDeFichero    fs.FileMode = 0o600
)

// Regenerado es lo que Regenerar construye en memoria para las skills de un
// repositorio, sin escribir nada.
type Regenerado struct {
	// Skills son las skills del repositorio, en el orden de sus nombres.
	Skills []SkillRegenerada
}

// SkillRegenerada es lo que la sincronización genera para una skill desde su
// declaración de kitlegal, desde data/ y desde las descripciones de los verbos
// (research.md D4; FR-035), o los defectos que lo impiden.
type SkillRegenerada struct {
	// Nombre es el del directorio de la skill.
	Nombre string

	// Contenido son los bytes de SKILL.md con la región de la tabla de comandos
	// regenerada, o tal cual si la skill no declara applets.
	Contenido []byte

	// Referencias son las referencias generadas, una por cada nombre de
	// kitlegal-referencias y en su orden; nil si no declara ninguna.
	Referencias []Referencia

	// Defectos son los de la skill, en el orden frontmatter, líneas, región,
	// datos y scripts/; nil si no tiene ninguno. Con alguno, Contenido y
	// Referencias van vacíos: no hay nada regenerado que escribir ni con que
	// comparar.
	Defectos []*DefectoDeSkill
}

// Referencia es un fichero generado de references/ (data-model §4.2).
type Referencia struct {
	// Fichero es su nombre dentro de references/: <nombre>.md.
	Fichero string

	// Contenido son sus bytes.
	Contenido []byte
}

// Defectos son los defectos de todas las skills, en el orden de las skills; nil
// si no hay ninguno.
func (r Regenerado) Defectos() []*DefectoDeSkill {
	var defectos []*DefectoDeSkill

	for _, skill := range r.Skills {
		defectos = append(defectos, skill.Defectos...)
	}

	return defectos
}

// ClaseDeDeriva es la clase de una deriva (data-model §5).
type ClaseDeDeriva string

// Las clases de deriva de data-model §5.
const (
	// DerivaContenidoDistinto es la de un fichero generado —SKILL.md o una
	// referencia— que existe con otros bytes, o que no es un fichero regular.
	DerivaContenidoDistinto ClaseDeDeriva = "contenido-distinto"

	// DerivaFicheroAusente es la de un fichero generado que falta.
	DerivaFicheroAusente ClaseDeDeriva = "fichero-ausente"

	// DerivaFicheroSobrante es la de una entrada de references/ que no es
	// ninguna referencia declarada.
	DerivaFicheroSobrante ClaseDeDeriva = "fichero-sobrante"
)

// Deriva es una diferencia entre lo regenerado de una skill y lo que hay en su
// directorio (data-model §5; FR-042).
type Deriva struct {
	// Skill es el nombre del directorio de la skill.
	Skill string

	// Ruta es la del fichero dentro del directorio de la skill, con / como
	// separador: SKILL.md, references/normas.md.
	Ruta string

	// Clase es la clase de la deriva.
	Clase ClaseDeDeriva

	// Detalle precisa la clase cuando no lo dice todo —«no es un fichero
	// regular»—, o va vacío.
	Detalle string
}

// Error presenta la skill, la ruta y la clase, con el detalle entre paréntesis si
// lo hay, como «boe-legislacion: references/normas.md: contenido-distinto (no es
// un fichero regular)».
func (d *Deriva) Error() string {
	texto := d.Skill + ": " + d.Ruta + ": " + string(d.Clase)
	if d.Detalle == "" {
		return texto
	}

	return texto + " (" + d.Detalle + ")"
}

// Regenerar construye en memoria, sin escribir nada, lo que la sincronización
// genera para cada skill del repositorio de raíz (contrato
// sincronizacion-y-comprobacion §2; data-model §1 a §5; research.md D4 y D7;
// FR-034, FR-035), con las descripciones de los verbos de los applets
// registrados, de las que salen también los nombres de esos applets:
//
//   - el frontmatter de SKILL.md se lee y se valida (LeerFrontmatter y
//     ValidarFrontmatter), así que cada referencia declarada tiene fila en la
//     tabla de generadores y el YAML de datos que esa fila declara;
//   - con el frontmatter sin defectos, SKILL.md lleva en su región la tabla de
//     comandos de los applets que declara, si declara alguno, y cada referencia
//     declarada se genera con su fila desde su YAML de datos;
//   - SKILL.md, tal como quedaría, tiene que tener menos de 300 líneas;
//   - la skill no lleva scripts/ (defectoDeLosScripts), tampoco sin SKILL.md.
//
// Nada de eso sale de una lista escrita para una skill concreta: sale de la
// declaración de kitlegal de cada una y de data/. Los defectos de una skill van
// en el orden frontmatter, líneas, región, datos y scripts/, cada uno con la
// skill y lo que falla; con alguno, la skill no lleva nada regenerado. El error
// queda para lo que impide regenerar: un directorio de skills o de datos que no
// se puede listar, o una entrada que no se puede consultar o leer.
func Regenerar(raiz string, descripciones []DescripcionDeVerbo) (Regenerado, error) {
	nombres, err := Listar(raiz)
	if err != nil {
		return Regenerado{}, err
	}

	registrados := appletsDescritos(descripciones)
	regenerado := Regenerado{Skills: make([]SkillRegenerada, 0, len(nombres))}

	for _, nombre := range nombres {
		skill, err := regenerarSkill(raiz, nombre, registrados, descripciones)
		if err != nil {
			return Regenerado{}, err
		}

		regenerado.Skills = append(regenerado.Skills, skill)
	}

	return regenerado, nil
}

// appletsDescritos son los applets de las descripciones, cada uno una vez, en el
// orden en que aparece por primera vez: solo se describen los verbos de un applet
// registrado.
func appletsDescritos(descripciones []DescripcionDeVerbo) []string {
	var applets []string

	for _, descripcion := range descripciones {
		if !slices.Contains(applets, descripcion.Applet) {
			applets = append(applets, descripcion.Applet)
		}
	}

	return applets
}

// regenerarSkill construye lo regenerado de la skill de nombre, o sus defectos.
func regenerarSkill(raiz, nombre string, registrados []string, descripciones []DescripcionDeVerbo) (
	SkillRegenerada, error,
) {
	skill, err := Cargar(raiz, nombre)

	// Sin un SKILL.md que leer, Cargar da su defecto y no hay nada que
	// regenerar, pero scripts/ se mira igual.
	var deSkillMd *DefectoDeSkill

	sinSkillMd := errors.As(err, &deSkillMd)
	if err != nil && !sinSkillMd {
		return SkillRegenerada{}, err
	}

	deLosScripts, err := defectoDeLosScripts(raiz, nombre)
	if err != nil {
		return SkillRegenerada{}, err
	}

	if sinSkillMd {
		return SkillRegenerada{Nombre: nombre, Defectos: slices.Concat([]*DefectoDeSkill{deSkillMd}, deLosScripts)}, nil
	}

	deKitlegal, deFrontmatter, err := declaracionDeLaSkill(raiz, skill, registrados)
	if err != nil {
		return SkillRegenerada{}, err
	}

	regenerada := SkillRegenerada{Nombre: nombre, Contenido: skill.Contenido}

	var deLaRegion, deLosDatos []*DefectoDeSkill

	if len(deFrontmatter) == 0 {
		regenerada.Contenido, deLaRegion = contenidoRegenerado(skill, deKitlegal.Applets, descripciones)

		regenerada.Referencias, deLosDatos, err = referenciasGeneradas(raiz, nombre, deKitlegal.Referencias)
		if err != nil {
			return SkillRegenerada{}, err
		}
	}

	defectos := slices.Concat(deFrontmatter, defectoDeLasLineas(nombre, regenerada.Contenido), deLaRegion, deLosDatos,
		deLosScripts)
	if len(defectos) > 0 {
		return SkillRegenerada{Nombre: nombre, Defectos: defectos}, nil
	}

	return regenerada, nil
}

// defectoDeLosScripts es el defecto de la skill de nombre si lleva scripts/
// (ADR 0019; contracts/skills-e-invocacion.md §3 de H19; FR-082): cualquier
// entrada con ese nombre en su directorio, vista sin seguirla —un directorio,
// vacío o no, un fichero o un enlace, aunque no resuelva—. Nadie la retira: no
// hay nada regenerado que poner en su lugar (research.md D17 de H19).
func defectoDeLosScripts(raiz, skill string) ([]*DefectoDeSkill, error) {
	_, err := os.Lstat(filepath.Join(raiz, carpetaDeLasSkills, skill, carpetaDeLosScripts))

	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("%s: %s no se puede consultar: %w", skill, carpetaDeLosScripts, err)
	}

	return []*DefectoDeSkill{{Skill: skill, Defecto: defectoDeScripts}}, nil
}

// declaracionDeLaSkill lee y valida el frontmatter de la skill, y devuelve su
// declaración de kitlegal con los defectos del frontmatter: los de su lectura,
// con SKILL.md delante, sin declaración; o los de ValidarFrontmatter, entre ellos
// el de cada referencia declarada sin generador.
func declaracionDeLaSkill(raiz string, skill Skill, registrados []string) (
	DeclaracionDeKitlegal, []*DefectoDeSkill, error,
) {
	frontmatter, err := LeerFrontmatter(skill.Contenido)
	if err != nil {
		return DeclaracionDeKitlegal{}, defectosDelError(skill.Nombre, ficheroDeLaSkill, err), nil
	}

	defectos, err := ValidarFrontmatter(raiz, skill.Nombre, frontmatter, registrados)
	if err != nil {
		return DeclaracionDeKitlegal{}, nil, fmt.Errorf("%s: %w", skill.Nombre, err)
	}

	return frontmatter.DeclaracionDeKitlegal(), defectos, nil
}

// contenidoRegenerado es el SKILL.md de la skill con la tabla de comandos de los
// applets declarados en su región, o tal cual si no declara ninguno. Si la tabla
// no se puede escribir o la región no se puede sustituir, es el SKILL.md tal cual
// y el defecto que lo impide, con SKILL.md delante.
func contenidoRegenerado(skill Skill, applets []string, descripciones []DescripcionDeVerbo) (
	[]byte, []*DefectoDeSkill,
) {
	if len(applets) == 0 {
		return skill.Contenido, nil
	}

	tabla, err := RenderizarTabla(applets, descripciones)
	if err != nil {
		return skill.Contenido, defectosDelError(skill.Nombre, ficheroDeLaSkill, err)
	}

	sustituido, err := SustituirRegion(skill.Contenido, tabla)
	if err != nil {
		return skill.Contenido, defectosDelError(skill.Nombre, ficheroDeLaSkill, err)
	}

	return sustituido, nil
}

// referenciasGeneradas son las referencias declaradas por la skill, cada una
// generada con su fila de la tabla de generadores desde el YAML de datos que esa
// fila declara, en su orden, y un defecto por cada error de unos datos que no se
// pueden generar, con su YAML delante. Los defectos de un YAML de datos van una
// sola vez aunque salgan de él varias referencias. Solo se llama con el
// frontmatter sin defectos: cada referencia tiene fila y YAML de datos.
func referenciasGeneradas(raiz, skill string, nombres []string) ([]Referencia, []*DefectoDeSkill, error) {
	var (
		referencias []Referencia
		defectos    []*DefectoDeSkill
	)

	conDefectos := make(map[string]bool, len(nombres))

	for _, nombre := range nombres {
		generador := generadoresDeReferencias[nombre]
		if conDefectos[generador.datos] {
			continue
		}

		datos := carpetaDeLosDatos + "/" + generador.datos

		contenido, err := leerFichero(filepath.Join(raiz, carpetaDeLosDatos, generador.datos))
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %s no se puede leer: %w", skill, datos, err)
		}

		generada, err := generador.generar(contenido)
		if err != nil {
			conDefectos[generador.datos] = true
			defectos = append(defectos, defectosDelError(skill, datos, err)...)

			continue
		}

		referencias = append(referencias, Referencia{Fichero: nombre + extensionDeLasReferencias, Contenido: generada})
	}

	if len(defectos) > 0 {
		return nil, defectos, nil
	}

	return referencias, nil, nil
}

// defectoDeLasLineas es el defecto de un SKILL.md con 300 líneas o más.
func defectoDeLasLineas(skill string, contenido []byte) []*DefectoDeSkill {
	lineas := ContarLineas(contenido)
	if lineas <= maximoDeLineas {
		return nil
	}

	return []*DefectoDeSkill{{
		Skill:   skill,
		Defecto: fmt.Sprintf("%s tiene %d líneas (máximo %d)", ficheroDeLaSkill, lineas, maximoDeLineas),
	}}
}

// defectosDelError es un defecto de la skill por cada error que une err —o uno
// solo, si no une varios—, cada uno con el fichero delante.
func defectosDelError(skill, fichero string, err error) []*DefectoDeSkill {
	errores := []error{err}

	var unidos interface{ Unwrap() []error }
	if errors.As(err, &unidos) {
		errores = unidos.Unwrap()
	}

	defectos := make([]*DefectoDeSkill, 0, len(errores))
	for _, unido := range errores {
		defectos = append(defectos, &DefectoDeSkill{Skill: skill, Defecto: fichero + ": " + unido.Error()})
	}

	return defectos
}

// Comparar compara lo regenerado con el árbol del repositorio de raíz sin
// escribir nada (data-model §5; FR-042) y devuelve una Deriva por cada
// diferencia, skill a skill y, en cada una, en este orden: SKILL.md, las
// referencias declaradas y las entradas sobrantes de references/. Una skill con
// defectos no tiene nada regenerado con que comparar y no da ninguna. El error
// queda para lo que impide comparar: una entrada que no se puede consultar o
// leer, o un references/ que no es un directorio.
func Comparar(raiz string, regenerado Regenerado) ([]*Deriva, error) {
	arreglos, err := arreglosDe(raiz, regenerado)
	if err != nil {
		return nil, err
	}

	var derivas []*Deriva

	for _, arreglo := range arreglos {
		derivas = append(derivas, arreglo.deriva)
	}

	return derivas, nil
}

// Escribir deja el árbol del repositorio de raíz como lo regenerado
// (contrato sincronizacion-y-comprobacion §2; FR-034, FR-036): deshace cada
// deriva que da Comparar, y nada más. Escribe los ficheros generados que faltan o
// que difieren y retira las entradas sobrantes de references/. Lo que ya
// coincide no se toca, así que sobre un árbol sin derivas no escribe nada, ni
// siquiera un tiempo de modificación. No crea ningún scripts/ ni retira el que
// haya (ADR 0019; FR-080 y FR-082 de H19): una skill que lo lleva tiene un
// defecto.
//
// Si alguna skill tiene defectos no escribe nada, tampoco lo de las demás, y
// devuelve un error que une todos. Antes de escribir busca todas las derivas, de
// modo que un error al buscarlas no deja nada a medias.
func Escribir(raiz string, regenerado Regenerado) error {
	if defectos := regenerado.Defectos(); len(defectos) > 0 {
		errores := make([]error, 0, len(defectos))
		for _, defecto := range defectos {
			errores = append(errores, defecto)
		}

		return fmt.Errorf("no se escribe nada porque hay defectos: %w", errors.Join(errores...))
	}

	arreglos, err := arreglosDe(raiz, regenerado)
	if err != nil {
		return err
	}

	for _, arreglo := range arreglos {
		if err := arreglo.aplicar(); err != nil {
			return err
		}
	}

	return nil
}

// arreglo es una deriva con lo que Escribir necesita para deshacerla.
type arreglo struct {
	// deriva es la diferencia que deshace.
	deriva *Deriva

	// ruta es la del fichero en el sistema.
	ruta string

	// contenido son los bytes regenerados del fichero; vacío en lo sobrante.
	contenido []byte

	// retirar dice si lo que ocupa la ruta hay que retirarlo antes de escribir:
	// lo que no es un fichero regular donde va un fichero. Escribir en un enlace
	// escribiría en su destino.
	retirar bool
}

// aplicar deshace la deriva del arreglo en el árbol: retira lo sobrante y lo que
// hay que retirar y, salvo en lo sobrante, crea lo que falte del directorio que
// lo contiene y deja en la ruta el fichero.
func (a arreglo) aplicar() error {
	sobrante := a.deriva.Clase == DerivaFicheroSobrante

	if a.retirar || sobrante {
		if err := os.Remove(a.ruta); err != nil {
			return fmt.Errorf("%s: %s no se puede retirar: %w", a.deriva.Skill, a.deriva.Ruta, err)
		}
	}

	if sobrante {
		return nil
	}

	if err := crearDirectorio(filepath.Dir(a.ruta)); err != nil {
		return fmt.Errorf("%s: %s: %w", a.deriva.Skill, a.deriva.Ruta, err)
	}

	if err := os.WriteFile(a.ruta, a.contenido, permisosDeFichero); err != nil {
		return fmt.Errorf("%s: %s no se puede escribir: %w", a.deriva.Skill, a.deriva.Ruta, err)
	}

	return nil
}

// crearDirectorio crea el directorio y lo que falte de su ruta.
func crearDirectorio(directorio string) error {
	if err := os.MkdirAll(directorio, permisosDeDirectorio); err != nil {
		return fmt.Errorf("el directorio %s no se puede crear: %w", directorio, err)
	}

	return nil
}

// arreglosDe son los arreglos de las derivas de cada skill sin defectos, en el
// orden de Comparar.
func arreglosDe(raiz string, regenerado Regenerado) ([]arreglo, error) {
	var arreglos []arreglo

	for _, skill := range regenerado.Skills {
		if len(skill.Defectos) > 0 {
			continue
		}

		deLaSkill, err := arreglosDeLaSkill(filepath.Join(raiz, carpetaDeLasSkills, skill.Nombre), skill)
		if err != nil {
			return nil, err
		}

		arreglos = append(arreglos, deLaSkill...)
	}

	return arreglos, nil
}

// arreglosDeLaSkill son los arreglos de las derivas de la skill, cuyo directorio
// es el dado: SKILL.md, las referencias y las entradas sobrantes de references/.
func arreglosDeLaSkill(directorio string, skill SkillRegenerada) ([]arreglo, error) {
	enReferences, err := entradasDeLaCarpeta(directorio, skill.Nombre, carpetaDeLasReferencias)
	if err != nil {
		return nil, err
	}

	arreglos, err := arregloDelFichero(directorio, skill.Nombre, ficheroDeLaSkill, skill.Contenido)
	if err != nil {
		return nil, err
	}

	ficheros := make([]string, 0, len(skill.Referencias))

	for _, referencia := range skill.Referencias {
		deLaReferencia, err := arregloDelFichero(directorio, skill.Nombre,
			carpetaDeLasReferencias+"/"+referencia.Fichero, referencia.Contenido)
		if err != nil {
			return nil, err
		}

		arreglos = append(arreglos, deLaReferencia...)
		ficheros = append(ficheros, referencia.Fichero)
	}

	arreglos = append(arreglos, sobrantes(directorio, skill.Nombre, carpetaDeLasReferencias, enReferences, ficheros)...)

	return arreglos, nil
}

// entradasDeLaCarpeta son los nombres de las entradas de la carpeta del
// directorio de la skill, en orden de nombre, o ninguno si no existe. Una carpeta
// que no es un directorio —un fichero, o un enlace, que llevaría las escrituras
// fuera de la skill— es un error que la nombra.
func entradasDeLaCarpeta(directorio, skill, carpeta string) ([]string, error) {
	ruta := filepath.Join(directorio, carpeta)

	estado, err := os.Lstat(ruta)

	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("%s: %s no se puede consultar: %w", skill, carpeta, err)
	case !estado.IsDir():
		return nil, fmt.Errorf("%s: %s no es un directorio", skill, carpeta)
	}

	entradas, err := os.ReadDir(ruta)
	if err != nil {
		return nil, fmt.Errorf("%s: %s no se puede listar: %w", skill, carpeta, err)
	}

	nombres := make([]string, 0, len(entradas))
	for _, entrada := range entradas {
		nombres = append(nombres, entrada.Name())
	}

	return nombres, nil
}

// arregloDelFichero es el arreglo del fichero generado de la ruta dada dentro del
// directorio de la skill si no coincide con su contenido regenerado: ausente, que
// no es un fichero regular o con otros bytes. Si coincide, ninguno.
func arregloDelFichero(directorio, skill, relativa string, contenido []byte) ([]arreglo, error) {
	ruta := filepath.Join(directorio, filepath.FromSlash(relativa))
	deriva := &Deriva{Skill: skill, Ruta: relativa}
	arreglado := arreglo{deriva: deriva, ruta: ruta, contenido: contenido}

	estado, err := os.Lstat(ruta)

	switch {
	case errors.Is(err, fs.ErrNotExist):
		deriva.Clase = DerivaFicheroAusente

		return []arreglo{arreglado}, nil
	case err != nil:
		return nil, fmt.Errorf("%s: %s no se puede consultar: %w", skill, relativa, err)
	case !estado.Mode().IsRegular():
		deriva.Clase, deriva.Detalle, arreglado.retirar = DerivaContenidoDistinto, "no es un fichero regular", true

		return []arreglo{arreglado}, nil
	}

	actual, err := leerFichero(ruta)
	if err != nil {
		return nil, fmt.Errorf("%s: %s no se puede leer: %w", skill, relativa, err)
	}

	if bytes.Equal(actual, contenido) {
		return nil, nil
	}

	deriva.Clase = DerivaContenidoDistinto

	return []arreglo{arreglado}, nil
}

// sobrantes son los arreglos de las entradas de la carpeta del directorio de la
// skill que no son ninguno de los nombres esperados, en el orden de las entradas:
// cada una, un fichero sobrante.
func sobrantes(directorio, skill, carpeta string, entradas, esperados []string) []arreglo {
	var arreglos []arreglo

	for _, entrada := range entradas {
		if slices.Contains(esperados, entrada) {
			continue
		}

		arreglos = append(arreglos, arreglo{
			deriva: &Deriva{Skill: skill, Ruta: carpeta + "/" + entrada, Clase: DerivaFicheroSobrante},
			ruta:   filepath.Join(directorio, carpeta, entrada),
		})
	}

	return arreglos
}

// leerFichero lee el fichero de la ruta. filepath.Clean es lo que el control de
// rutas reconoce como saneado antes de abrir un fichero (gosec G304).
func leerFichero(ruta string) ([]byte, error) {
	return os.ReadFile(filepath.Clean(ruta))
}
