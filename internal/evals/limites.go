package evals

import "strings"

// ClaseDeLimite es la clase por la que una sesión queda sin medir por un límite
// de uso de la cuenta, o SinLimite si se mide (data-model §3 de H7.3; research.md
// D6 de H7.3).
type ClaseDeLimite int

const (
	// SinLimite es la sesión que se mide como cualquier otra, también la que
	// termina tras reintentos de los que se recupera (FR-041 de H7.3).
	SinLimite ClaseDeLimite = iota
	// LimiteMensajeDeUso es la clase (a): el último mensaje es el result con
	// is_error del mensaje del límite de uso. Es la única tras la que no se abre
	// ninguna sesión más (FR-044 de H7.3).
	LimiteMensajeDeUso
	// LimiteReintentosAgotados es la clase (b): la sesión no terminó y su último
	// reintento es por rate_limit y el último que Claude Code iba a hacer.
	LimiteReintentosAgotados
	// LimiteCortadaDuranteReintentos es la clase (c): el tope cortó la sesión
	// cuando su último mensaje era un reintento por rate_limit.
	LimiteCortadaDuranteReintentos
)

// Descripciones de las clases de límite que publica el informe, la de la sesión
// que el repartidor no abrió tras el mensaje del límite de uso, y el principio
// del motivo de una sesión sin medir (data-model §3 de H7.3).
const (
	descripcionDelMensajeDeUso       = "mensaje del límite de uso: "
	descripcionDeReintentosAgotados  = "reintentos por rate_limit agotados"
	descripcionDeCortadaEnReintentos = "cortada por el tope durante reintentos por rate_limit"
	descripcionSinAbrir              = "sin abrir tras el límite de uso"
	motivoSinMedir                   = "sin medir por límite de uso: "
)

// principiosDelLimiteDeUso son los principios del texto con que Claude Code dice
// que la cuenta de una suscripción ha llegado a su límite de uso: los de su
// propia lista que llegan con el token de claude setup-token (research.md V3 y
// D6 de H7.3).
var principiosDelLimiteDeUso = [...]string{"You've hit your", "You've reached your", "You're out of"}

// LimiteDeUso es lo que la clasificación dice de una sesión: su clase y, si no
// es SinLimite, su descripción.
type LimiteDeUso struct {
	// Clase es la primera que aplica de data-model §3 de H7.3.
	Clase ClaseDeLimite

	// Descripcion es la clase como la publica el informe, en sin_medir de la
	// eval y en el motivo de sesiones_sin_medir: «mensaje del límite de uso:
	// <texto>» con el texto de Claude Code, «reintentos por rate_limit
	// agotados» o «cortada por el tope durante reintentos por rate_limit»;
	// vacía con SinLimite (contrato informe-del-job §3 de H7.3).
	Descripcion string
}

// Motivo es el único motivo de la sesión sin medir, «sin medir por límite de
// uso: <descripción>», o vacío si la sesión se mide (data-model §3 de H7.3).
func (l LimiteDeUso) Motivo() string {
	if l.Clase == SinLimite {
		return ""
	}

	return motivoSinMedir + l.Descripcion
}

// dejarSinMedir deja sin medir el resultado de una sesión que un límite de uso
// de la cuenta no dejó terminar: no pasa, lleva la descripción de su clase en
// SinMedir y el motivo del límite como único motivo, porque lo que no llegó a
// hacer no es un defecto de la skill (FR-040 de H7.3). Lo observado de la sesión
// —la respuesta, las invocaciones, las llegadas a la red— se informa igual. Con
// SinLimite no cambia nada.
func (r *ResultadoDeEval) dejarSinMedir(limite LimiteDeUso) {
	if limite.Clase == SinLimite {
		return
	}

	r.SinMedir = limite.Descripcion
	r.Motivos = []string{limite.Motivo()}
	r.Pasa = false
}

// ClasificarElLimite dice, sin modelo y con lo que LeerSesion lee del
// transcript, si la sesión queda sin medir por un límite de uso de la cuenta, con
// la primera clase que aplica en el orden de data-model §3 de H7.3: (a) el
// mensaje del límite de uso, (b) los reintentos por rate_limit agotados y (c) el
// tope durante reintentos por rate_limit. Los reintentos por otra causa —la
// sobrecarga, un 529— no son un límite de la cuenta, y una sesión que termina se
// mide aunque haya reintentado (FR-040 y FR-041 de H7.3).
func ClasificarElLimite(sesion Sesion) LimiteDeUso {
	for _, principio := range principiosDelLimiteDeUso {
		if strings.HasPrefix(sesion.ErrorDelResultado, principio) {
			return LimiteDeUso{Clase: LimiteMensajeDeUso, Descripcion: descripcionDelMensajeDeUso + sesion.ErrorDelResultado}
		}
	}

	if len(sesion.Reintentos) == 0 {
		return LimiteDeUso{}
	}

	ultimo := sesion.Reintentos[len(sesion.Reintentos)-1]
	if ultimo.Error != errorDeLimiteDeRitmo {
		return LimiteDeUso{}
	}

	switch {
	case !sesion.Terminada && ultimo.Intento >= ultimo.Maximo:
		return LimiteDeUso{Clase: LimiteReintentosAgotados, Descripcion: descripcionDeReintentosAgotados}
	case sesion.Cortada && sesion.TerminaEnReintento:
		return LimiteDeUso{Clase: LimiteCortadaDuranteReintentos, Descripcion: descripcionDeCortadaEnReintentos}
	default:
		return LimiteDeUso{}
	}
}
