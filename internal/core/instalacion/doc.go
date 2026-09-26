// Package instalacion decide todo lo que el applet skills hace con las skills
// que el binario lleva empotradas: qué es de quién en el disco, qué se
// escribe, se enlaza o se retira y en qué orden, el manifiesto kitlegal.json,
// los hallazgos de doctor con la orden que arregla cada uno, la igualdad de
// versiones y el aviso sin red de que lo instalado es de otra versión
// (docs/ADR/0019; research.md D1 de H19).
//
// El paquete no hace entrada ni salida. Recibe las skills empotradas ya
// leídas, como bytes, de quien compone el binario, y ve el disco a través de
// los puertos que declara él mismo y que implementa un adaptador:
//
//   - Disco, solo lectura y sin seguir nunca un enlace simbólico: examina una
//     ruta (ausente, directorio, fichero, enlace u otra cosa, con el destino
//     literal del enlace y si resuelve), calcula la huella de un fichero
//     regular, lo lee y enumera las entradas de un directorio real.
//   - Escritor, las operaciones sueltas de la fase de aplicación: crear un
//     directorio, escribir un fichero de forma atómica, retirar una entrada y
//     enlazar.
//   - Enlazador, si el sistema de ficheros de un directorio admite enlaces
//     simbólicos y crear uno.
//
// El dominio planifica sobre lo que el Disco le enseña, y Aplicar lleva el
// plan a cabo a través del Escritor, fase a fase; el adaptador ejecuta
// operaciones sueltas y no decide nada.
//
// Lo que garantiza:
//
//   - Todo conflicto se detecta antes de escribir nada, cada entrada cae como
//     mucho en una clase, y lo que no puso ahí el binario no se sobrescribe
//     ni se retira (FR-040 a FR-043, FR-047).
//   - Por debajo de la raíz del ámbito nada se lee, se escribe ni se retira a
//     través de un enlace simbólico, y lo que no es un fichero regular no se
//     abre (FR-028).
//   - Las mismas skills sobre el mismo disco dan el mismo plan, el mismo
//     manifiesto byte a byte y los mismos hallazgos en el mismo orden; sin
//     nada que cambiar, el plan está vacío y el disco queda igual (FR-033,
//     FR-045, FR-066).
//   - Tras un fallo a mitad de la aplicación, volver a planificar no da ningún
//     conflicto y completa la instalación (FR-044).
//   - Dos versiones son iguales si y solo si, quitada a cada una una v inicial,
//     son la misma cadena byte a byte; ninguna se ordena, y el aviso solo
//     compara si la versión del binario tiene la forma de SemVer 2.0.0 (FR-073,
//     FR-077).
//   - El aviso nunca es un error: lo que no permite decidir no tiene ningún
//     efecto (FR-072).
//   - Una invocación inválida es un error de clase «argumentos»; un conflicto
//     de install, un hallazgo de doctor o un ámbito ilegible, de clase
//     «inesperado»; y ninguna entrada provoca un panic (data-model §9).
//
// Es dominio puro: no importa os, io, io/fs, log ni ningún adaptador —tampoco
// internal/disco ni lo empotrado en la raíz del módulo— y no abre ninguna
// conexión (regla R1 de docs/ROADMAP.md §2; research.md D32 de H19).
package instalacion
