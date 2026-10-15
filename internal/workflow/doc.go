// Package workflow solo tiene tests: ejerce los guiones del workflow `hito`
// —scripts/workflow/revision.sh, gate.sh, modelos.sh, informe.sh y
// scripts/claude-modelo.sh— sobre un repositorio de git en un directorio
// temporal, con las copias de esos guiones, la de workflow.yml y un Claude
// Code sustituto que no abre ninguna sesión.
//
// Nació con el workflow 2.7.0, por dos defectos de H26: la revisión final no
// se podía volver a juzgar entera con el ciclo cerrado (revision.sh rejuzgar),
// y nada registraba con qué modelo corrió de verdad cada sesión
// (gates/modelos.jsonl, ADR 0033). Los tests del cierre y del informe de las
// evals siguen en internal/evals, junto al código que comprueban.
//
// Necesitan bash, git y jq, con los que trabajan los guiones: donde falta
// alguno, el test se salta diciéndolo; la CI los tiene.
package workflow
