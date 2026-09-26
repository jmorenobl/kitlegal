package instalacion

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"slices"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// SkillEmpotrada es una skill que el binario lleva dentro: el nombre de su
// directorio skills/<nombre>/, que tiene SKILL.md, y sus ficheros (data-model
// §1; FR-001, FR-004). La construye quien compone el binario a partir de lo
// empotrado y el dominio la recibe ya leída, como bytes: no lee nada del
// sistema de ficheros.
type SkillEmpotrada struct {
	// Nombre es el del directorio de la skill.
	Nombre string

	// Ficheros son SKILL.md y todo fichero bajo references/, en orden de
	// ruta.
	Ficheros []FicheroEmpotrado
}

// FicheroEmpotrado es un fichero de una skill empotrada, byte a byte lo que
// había en el árbol al compilar (FR-003).
type FicheroEmpotrado struct {
	// Ruta es la del fichero relativa al directorio de la skill, con /:
	// SKILL.md, references/normas.md.
	Ruta string

	// Contenido son sus bytes.
	Contenido []byte

	// Huella es la de Contenido, la que declara el manifiesto por cada
	// fichero instalado (HuellaDe).
	Huella string
}

// NuevoFicheroEmpotrado es el fichero de ruta y contenido dados con la huella
// de ese contenido. Guarda una copia de los bytes, de modo que la huella siga
// siendo la del contenido aunque quien lo llama cambie después los suyos.
func NuevoFicheroEmpotrado(ruta string, contenido []byte) FicheroEmpotrado {
	return FicheroEmpotrado{Ruta: ruta, Contenido: slices.Clone(contenido), Huella: HuellaDe(contenido)}
}

// HuellaDe es la huella de unos bytes tal como la declara el manifiesto: el
// prefijo del algoritmo seguido de los 64 dígitos hexadecimales en minúscula
// de su SHA-256, la misma forma que el hash del sobre (contracts/manifiesto.md
// §1).
func HuellaDe(contenido []byte) string {
	suma := sha256.Sum256(contenido)

	return schema.PrefijoHuella + hex.EncodeToString(suma[:])
}

// formaDeHuella es la de toda huella que el manifiesto admite, la del hash del
// sobre.
var formaDeHuella = regexp.MustCompile(schema.PatronHuella)

// esHuella dice si huella tiene la forma de las que produce HuellaDe.
func esHuella(huella string) bool {
	return formaDeHuella.MatchString(huella)
}

// maximoDelNombreDeSkill es la longitud máxima del nombre de una skill.
const maximoDelNombreDeSkill = 64

// formaDelNombreDeSkill es la del nombre de una skill (research.md V44): a-z,
// 0-9 y guiones, sin guion al principio, al final ni dos seguidos.
var formaDelNombreDeSkill = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// esNombreDeSkill dice si nombre tiene la forma del nombre de una skill, de 1 a
// 64 caracteres; como todos son ASCII, caracteres y bytes son los mismos.
func esNombreDeSkill(nombre string) bool {
	return len(nombre) <= maximoDelNombreDeSkill && formaDelNombreDeSkill.MatchString(nombre)
}
