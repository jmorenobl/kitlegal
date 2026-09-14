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

// Lo que forma la parte generada de una skill, además de la región de SKILL.md y
// de scripts/ (data-model §1 y §4.2).
const (
	// carpetaDeLasReferencias es el directorio de las referencias generadas de
	// una skill, dentro del suyo.
	carpetaDeLasReferencias = "references"

	// extensionDeLasReferencias es la extensión de una referencia generada:
	// references/<nombre>.md sale de data/<nombre>.yaml.
	extensionDeLasReferencias = ".md"

	// maximoDeLineas es el número de líneas de SKILL.md a partir del cual es un
	// defecto (FR-041: «300 líneas o más» falla).
	maximoDeLineas = 299
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

	// Enlaces son los enlaces esperados de scripts/ (EnlacesEsperados).
	Enlaces []Enlace

	// Defectos son los de la skill, en el orden frontmatter, líneas, región y
	// datos; nil si no tiene ninguno. Con alguno, Contenido, Referencias y
	// Enlaces van vacíos: no hay nada regenerado que escribir ni con que
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

	// DerivaEnlaceAusente es la de un enlace esperado de scripts/ que falta.
	DerivaEnlaceAusente ClaseDeDeriva = "enlace-ausente"

	// DerivaEnlaceSobrante es la de una entrada de scripts/ que no es ningún
	// enlace esperado.
	DerivaEnlaceSobrante ClaseDeDeriva = "enlace-sobrante"

	// DerivaEnlaceConOtroDestino es la de un enlace esperado de scripts/ que
	// apunta a otro sitio o que no es un enlace simbólico.
	DerivaEnlaceConOtroDestino ClaseDeDeriva = "enlace-con-otro-destino"
)

// Deriva es una diferencia entre lo regenerado de una skill y lo que hay en su
// directorio (data-model §5; FR-042).
type Deriva struct {
	// Skill es el nombre del directorio de la skill.
	Skill string

	// Ruta es la del fichero o el enlace dentro del directorio de la skill, con /
	// como separador: SKILL.md, references/normas.md, scripts/boe.
	Ruta string

	// Clase es la clase de la deriva.
	Clase ClaseDeDeriva

	// Detalle precisa la clase cuando no lo dice todo —«no es un fichero
	// regular», «no es un enlace simbólico» o «apunta a <destino>»—, o va vacío.
	Detalle string
}

// Error presenta la skill, la ruta y la clase, con el detalle entre paréntesis si
// lo hay, como «boe-legislacion: scripts/boe: enlace-con-otro-destino (apunta a
// ../../../bin/kitlegal)».
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
//     ValidarFrontmatter), y cada referencia declarada tiene que tener un
//     generador conocido;
//   - con el frontmatter sin defectos, SKILL.md lleva en su región la tabla de
//     comandos de los applets que declara, si declara alguno; cada referencia
//     declarada se genera desde su data/<nombre>.yaml; y los enlaces esperados
//     son los de EnlacesEsperados;
//   - SKILL.md, tal como quedaría, tiene que tener menos de 300 líneas.
//
// Nada de eso sale de una lista escrita para una skill concreta: sale de la
// declaración de kitlegal de cada una y de data/. Los defectos de una skill van
// en el orden frontmatter, líneas, región y datos, cada uno con la skill y lo que
// falla; con alguno, la skill no lleva nada regenerado. El error queda para lo
// que impide regenerar: un directorio de skills o de datos que no se puede
// listar, o un fichero que no se puede leer.
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
	var sinSkillMd *DefectoDeSkill

	skill, err := Cargar(raiz, nombre)

	switch {
	case errors.As(err, &sinSkillMd):
		return SkillRegenerada{Nombre: nombre, Defectos: []*DefectoDeSkill{sinSkillMd}}, nil
	case err != nil:
		return SkillRegenerada{}, err
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

		regenerada.Enlaces = EnlacesEsperados(deKitlegal)
	}

	defectos := slices.Concat(deFrontmatter, defectoDeLasLineas(nombre, regenerada.Contenido), deLaRegion, deLosDatos)
	if len(defectos) > 0 {
		return SkillRegenerada{Nombre: nombre, Defectos: defectos}, nil
	}

	return regenerada, nil
}

// declaracionDeLaSkill lee y valida el frontmatter de la skill, y devuelve su
// declaración de kitlegal con los defectos del frontmatter: los de su lectura,
// con SKILL.md delante, sin declaración; o los de ValidarFrontmatter seguidos de
// uno por cada referencia declarada sin generador conocido.
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

	deKitlegal := frontmatter.DeclaracionDeKitlegal()
	ruta := presentarRuta([]string{claveMetadata, claveKitlegalReferencias})

	for _, referencia := range deKitlegal.Referencias {
		if _, conocido := generadorDeReferencia(referencia); !conocido {
			defectos = append(defectos, &DefectoDeSkill{
				Skill:   skill.Nombre,
				Defecto: fmt.Sprintf("%s: %q sin generador conocido", ruta, referencia),
			})
		}
	}

	return deKitlegal, defectos, nil
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
// generada desde su YAML de datos, en su orden, y un defecto por cada error de
// unos datos que no se pueden generar, con su YAML delante. Solo se llama con el
// frontmatter sin defectos: cada referencia tiene generador y YAML de datos.
func referenciasGeneradas(raiz, skill string, nombres []string) ([]Referencia, []*DefectoDeSkill, error) {
	var (
		referencias []Referencia
		defectos    []*DefectoDeSkill
	)

	for _, nombre := range nombres {
		generar, _ := generadorDeReferencia(nombre)
		datos := carpetaDeLosDatos + "/" + nombre + extensionDeLosDatos

		contenido, err := leerFichero(filepath.Join(raiz, carpetaDeLosDatos, nombre+extensionDeLosDatos))
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %s no se puede leer: %w", skill, datos, err)
		}

		generada, err := generar(contenido)
		if err != nil {
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

// generadorDeReferencia es la función que genera la referencia de nombre desde
// el contenido de su YAML de datos, y si la hay (data-model §1.3): hoy, solo la
// de las normas.
func generadorDeReferencia(nombre string) (func(datos []byte) ([]byte, error), bool) {
	switch nombre {
	case nombreDeLasNormas:
		return generarNormas, true
	default:
		return nil, false
	}
}

// generarNormas es references/normas.md desde el contenido de data/normas.yaml.
func generarNormas(datos []byte) ([]byte, error) {
	normas, err := LeerNormas(datos)
	if err != nil {
		return nil, err
	}

	return RenderizarNormas(normas), nil
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
// referencias declaradas, las entradas sobrantes de references/, los enlaces
// esperados y las entradas sobrantes de scripts/. Una skill con defectos no tiene
// nada regenerado con que comparar y no da ninguna. El error queda para lo que
// impide comparar: una entrada que no se puede consultar o leer, o un references/
// o un scripts/ que no es un directorio.
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
// que difieren, crea los enlaces que faltan, corrige los que apuntan a otro sitio
// o no son un enlace, y retira las entradas sobrantes de references/ y de
// scripts/. Lo que ya coincide no se toca, así que sobre un árbol sin derivas no
// escribe nada, ni siquiera un tiempo de modificación.
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

	// ruta es la del fichero o el enlace en el sistema.
	ruta string

	// contenido son los bytes regenerados de un fichero, y destino el destino
	// literal de un enlace; cada uno vacío en las demás clases.
	contenido []byte
	destino   string

	// retirar dice si lo que ocupa la ruta hay que retirarlo antes de escribir:
	// lo que no es un fichero regular donde va un fichero, y un enlace con otro
	// destino, o lo que no es un enlace, donde va un enlace. Escribir en un
	// enlace escribiría en su destino.
	retirar bool
}

// aplicar deshace la deriva del arreglo en el árbol.
func (a arreglo) aplicar() error {
	if a.retirar || a.deriva.Clase == DerivaFicheroSobrante || a.deriva.Clase == DerivaEnlaceSobrante {
		if err := os.Remove(a.ruta); err != nil {
			return fmt.Errorf("%s: %s no se puede retirar: %w", a.deriva.Skill, a.deriva.Ruta, err)
		}
	}

	switch a.deriva.Clase {
	case DerivaFicheroSobrante, DerivaEnlaceSobrante:
		return nil
	case DerivaEnlaceAusente, DerivaEnlaceConOtroDestino:
		if err := crearDirectorio(filepath.Dir(a.ruta)); err != nil {
			return fmt.Errorf("%s: %s: %w", a.deriva.Skill, a.deriva.Ruta, err)
		}

		if err := os.Symlink(a.destino, a.ruta); err != nil {
			return fmt.Errorf("%s: %s no se puede enlazar: %w", a.deriva.Skill, a.deriva.Ruta, err)
		}
	default:
		if err := crearDirectorio(filepath.Dir(a.ruta)); err != nil {
			return fmt.Errorf("%s: %s: %w", a.deriva.Skill, a.deriva.Ruta, err)
		}

		if err := os.WriteFile(a.ruta, a.contenido, permisosDeFichero); err != nil {
			return fmt.Errorf("%s: %s no se puede escribir: %w", a.deriva.Skill, a.deriva.Ruta, err)
		}
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
// es el dado: SKILL.md, las referencias, las entradas sobrantes de references/,
// los enlaces y las entradas sobrantes de scripts/.
func arreglosDeLaSkill(directorio string, skill SkillRegenerada) ([]arreglo, error) {
	enReferences, err := entradasDeLaCarpeta(directorio, skill.Nombre, carpetaDeLasReferencias)
	if err != nil {
		return nil, err
	}

	enScripts, err := entradasDeLaCarpeta(directorio, skill.Nombre, carpetaDeLosScripts)
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

	arreglos = append(arreglos, sobrantes(directorio, skill.Nombre, carpetaDeLasReferencias, enReferences, ficheros,
		DerivaFicheroSobrante)...)

	nombres := make([]string, 0, len(skill.Enlaces))

	for _, enlace := range skill.Enlaces {
		delEnlace, err := arregloDelEnlace(directorio, skill.Nombre, enlace)
		if err != nil {
			return nil, err
		}

		arreglos = append(arreglos, delEnlace...)
		nombres = append(nombres, enlace.Nombre)
	}

	arreglos = append(arreglos, sobrantes(directorio, skill.Nombre, carpetaDeLosScripts, enScripts, nombres,
		DerivaEnlaceSobrante)...)

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

// arregloDelEnlace es el arreglo del enlace esperado dentro de scripts/ del
// directorio de la skill si no está como se espera: ausente, que no es un enlace
// simbólico o con otro destino. Si está, ninguno.
func arregloDelEnlace(directorio, skill string, enlace Enlace) ([]arreglo, error) {
	relativa := carpetaDeLosScripts + "/" + enlace.Nombre
	ruta := filepath.Join(directorio, carpetaDeLosScripts, enlace.Nombre)
	deriva := &Deriva{Skill: skill, Ruta: relativa, Clase: DerivaEnlaceConOtroDestino}
	arreglado := arreglo{deriva: deriva, ruta: ruta, destino: enlace.Destino, retirar: true}

	estado, err := os.Lstat(ruta)

	switch {
	case errors.Is(err, fs.ErrNotExist):
		deriva.Clase, arreglado.retirar = DerivaEnlaceAusente, false

		return []arreglo{arreglado}, nil
	case err != nil:
		return nil, fmt.Errorf("%s: %s no se puede consultar: %w", skill, relativa, err)
	case estado.Mode()&fs.ModeSymlink == 0:
		deriva.Detalle = "no es un enlace simbólico"

		return []arreglo{arreglado}, nil
	}

	destino, err := os.Readlink(ruta)
	if err != nil {
		return nil, fmt.Errorf("%s: %s no se puede leer como enlace: %w", skill, relativa, err)
	}

	if destino == enlace.Destino {
		return nil, nil
	}

	deriva.Detalle = "apunta a " + destino

	return []arreglo{arreglado}, nil
}

// sobrantes son los arreglos, con la clase dada, de las entradas de la carpeta
// del directorio de la skill que no son ninguno de los nombres esperados, en el
// orden de las entradas.
func sobrantes(directorio, skill, carpeta string, entradas, esperados []string, clase ClaseDeDeriva) []arreglo {
	var arreglos []arreglo

	for _, entrada := range entradas {
		if slices.Contains(esperados, entrada) {
			continue
		}

		arreglos = append(arreglos, arreglo{
			deriva: &Deriva{Skill: skill, Ruta: carpeta + "/" + entrada, Clase: clase},
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
