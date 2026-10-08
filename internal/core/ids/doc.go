// Package ids analiza los identificadores naturales del dominio: los que
// nombran de forma estable una cosa del mundo público y que el grafo usa como
// clave (docs/ADR/0014). H6 estrenó dos, el código INE de municipio y el
// código DIR3 de un ayuntamiento, y H23 añade los dos de una resolución
// judicial, el ECLI y el ROJ; ELI, CELEX y NIF entran con sus hitos, no antes
// (FR-034 de H6).
//
// Cada identificador es un tipo inmutable que solo se construye analizando:
// sus campos son privados, de modo que quien tiene un CodigoINE, un DIR3, un
// ECLI o un ROJ distinto del valor cero tiene uno bien formado. Analizar
// normaliza, y la normalización es idempotente: String() de lo analizado
// vuelve a analizarse al mismo valor y se escribe con la misma cadena (FR-032
// de H6). El ECLI y el ROJ no tienen nada que normalizar: se aceptan solo como
// se escriben, y String() devuelve la entrada.
//
// Lo que garantiza:
//
//   - Las cifras y las letras se comprueban byte a byte, sin expresiones
//     regulares: solo valen las cifras ASCII del 0 al 9 y las letras ASCII de
//     la A a la Z, no las de otras escrituras (contrato de identificadores §2
//     de H6).
//   - El código INE tiene la provincia entre 01 y 52 y el municipio entre 001
//     y 999; el DIR3 de ayuntamiento es L01 seguido de un código INE y de su
//     dígito de control.
//   - El dígito de control no se calcula: es dato oficial de la relación del
//     INE, y el paquete solo compara el declarado con el oficial que le da
//     quien resuelve (research.md D9 de H6).
//   - El ECLI es el de una resolución española, ECLI:ES:<órgano>:<año>:<número>,
//     en mayúsculas y sin blancos; el de otro país se rechaza diciendo que no
//     es español. El ROJ son unas siglas en mayúsculas, un espacio, el número,
//     una barra y el año. Ninguno de los dos se recorta ni se pasa a
//     mayúsculas, y ni el órgano ni las siglas se buscan en ninguna lista
//     (FR-005 de H23).
//   - El ECLI y el ROJ de una misma resolución se deducen el uno del otro en
//     una sola pareja, la de las sentencias del Tribunal Supremo: el ECLI de
//     órgano TS con el número solo de cifras y el ROJ de siglas STS, con el
//     número y el año trasladados carácter a carácter. En cualquier otro caso
//     no se deduce nada, y lo que no se deduce es el valor cero, nunca un
//     identificador compuesto a medias (FR-012 de H23).
//   - Toda entrada inválida devuelve un error de clase «argumentos», que el
//     kernel traduce a código 2, con la entrada entrecomillada y lo que tiene
//     de malo; nunca un panic ni un valor por omisión (FR-033 de H6, FR-006 de
//     H23).
//
// Es dominio puro: no hace entrada ni salida y no importa más paquete interno
// que internal/core/schema (regla R1 de docs/ROADMAP.md §2).
package ids
