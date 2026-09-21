// Package ids analiza los identificadores naturales del dominio: los que
// nombran de forma estable una cosa del mundo público y que el grafo usará
// como clave (docs/ADR/0014). H6 estrena dos, el código INE de municipio y el
// código DIR3 de un ayuntamiento; ELI, ECLI, CELEX y NIF entran con sus hitos,
// no antes (FR-034).
//
// Cada identificador es un tipo inmutable que solo se construye analizando:
// sus campos son privados, de modo que quien tiene un CodigoINE o un DIR3
// distinto del valor cero tiene uno bien formado. Analizar normaliza, y la
// normalización es idempotente: String() de lo analizado vuelve a analizarse
// al mismo valor y se escribe con la misma cadena (FR-032).
//
// Lo que garantiza:
//
//   - Las cifras se comprueban byte a byte, sin expresiones regulares: solo
//     valen las ASCII del 0 al 9, no las de otras escrituras (contrato de
//     identificadores §2).
//   - El código INE tiene la provincia entre 01 y 52 y el municipio entre 001
//     y 999; el DIR3 de ayuntamiento es L01 seguido de un código INE y de su
//     dígito de control.
//   - El dígito de control no se calcula: es dato oficial de la relación del
//     INE, y el paquete solo compara el declarado con el oficial que le da
//     quien resuelve (research.md D9).
//   - Toda entrada inválida devuelve un error de clase «argumentos», que el
//     kernel traduce a código 2, con la entrada entrecomillada y lo que tiene
//     de malo; nunca un panic ni un valor por omisión (FR-033).
//
// Es dominio puro: no hace entrada ni salida y no importa más paquete interno
// que internal/core/schema (regla R1 de docs/ROADMAP.md §2).
package ids
