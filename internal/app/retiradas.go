package app

import "github.com/jmorenobl/kitlegal/internal/core/instalacion"

// skillsRetiradas son los nombres de skill que este binario ya no empotra y,
// por cada uno, las skills de la entrega que heredan sus preguntas (H32, ADR
// 0048). install retira de cada ámbito la que está como la dejó kitlegal y
// avisa de la que no (contracts/instalacion-retiradas.md); TestSkillsRetiradas
// exige que ninguna se empotre y que cada heredera sí.
var skillsRetiradas = instalacion.Retiradas{
	"boe-legislacion": {"normativa"},
	"legal-core":      {"plazos-y-territorio"},
}
