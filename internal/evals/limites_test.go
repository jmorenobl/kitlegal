package evals

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Textos con los que termina el transcript de una sesión en los casos de
// TestClasificarElLimite: el mensaje del límite de uso de la cuenta con cada uno
// de sus tres principios (research.md V3), el de un 429 que no es de la cuenta y
// el del último reintento por rate_limit.
const (
	textoDelLimiteDeSesion  = "You've hit your session limit · resets 5pm"
	textoDelLimiteAlcanzado = "You've reached your usage limit"
	textoSinUsoExtra        = "You're out of extra usage · resets 5pm"
	textoDelLimiteTemporal  = "API Error: Server is temporarily limiting requests (not your usage limit)"
	textoDelError429        = `API Error: 429 {"type":"error","error":{"type":"rate_limit_error","message":"Rate limited"}}`
	mensajeAssistantDeTexto = `{"type":"assistant","message":{"content":[{"type":"text","text":"Consulto el BOE."}]}}` + "\n"
)

// TestClasificarElLimite fija la clasificación sin modelo de data-model §3, en su
// orden, sobre transcripts sintéticos que el propio test escribe en t.TempDir()
// y lee con LeerSesion (research.md D6; FR-040, FR-041, FR-093; SC-007): (a) el
// mensaje del límite de uso con cada uno de sus tres principios, y no el de un
// límite temporal del servidor ni una respuesta que empieza igual; (b) los
// reintentos por rate_limit agotados en una sesión que no terminó, y no los que
// quedan por agotar, ni los de sobrecarga, ni si el último reintento no es de
// rate_limit, ni si el último sale bien y la sesión termina; (c) la sesión que
// corta el tope, con 124 o 137, cuando su último mensaje es un reintento por
// rate_limit, y no por sobrecarga ni si hubo algo detrás; y medida, sin clase,
// la que termina tras reintentos de los que se recupera. Cada clase lleva su
// descripción y su motivo «sin medir por límite de uso: <descripción>».
func TestClasificarElLimite(t *testing.T) {
	t.Parallel()

	reintentosAgotados := mensajeDeReintento(9, 10, 429, "rate_limit") + mensajeDeReintento(10, 10, 429, "rate_limit")
	cortadaEnReintentos := mensajeDeReintento(1, 10, 429, "rate_limit") + mensajeDeReintento(2, 10, 429, "rate_limit")

	casos := []struct {
		nombre      string
		transcript  string
		codigo      int
		clase       ClaseDeLimite
		descripcion string
		terminada   bool
	}{
		{
			nombre:      "mensaje-del-limite-de-sesion",
			transcript:  mensajeResultConError(t, textoDelLimiteDeSesion),
			codigo:      1,
			clase:       LimiteMensajeDeUso,
			descripcion: "mensaje del límite de uso: You've hit your session limit · resets 5pm",
		},
		{
			nombre:      "mensaje-del-limite-alcanzado",
			transcript:  mensajeResultConError(t, textoDelLimiteAlcanzado),
			codigo:      1,
			clase:       LimiteMensajeDeUso,
			descripcion: "mensaje del límite de uso: You've reached your usage limit",
		},
		{
			nombre:      "mensaje-sin-uso-extra",
			transcript:  mensajeResultConError(t, textoSinUsoExtra),
			clase:       LimiteMensajeDeUso,
			descripcion: "mensaje del límite de uso: You're out of extra usage · resets 5pm",
		},
		{
			nombre:     "limite-temporal-del-servidor",
			transcript: mensajeResultConError(t, textoDelLimiteTemporal),
			codigo:     1,
			clase:      SinLimite,
		},
		{
			nombre: "respuesta-que-empieza-como-el-limite",
			transcript: `{"type":"result","subtype":"success","is_error":false,"result":` +
				cadenaJSON(t, textoDelLimiteDeSesion) + `}` + "\n",
			clase:     SinLimite,
			terminada: true,
		},
		{
			nombre:      "mensaje-del-limite-antes-que-reintentos-agotados",
			transcript:  reintentosAgotados + mensajeResultConError(t, textoDelLimiteDeSesion),
			codigo:      1,
			clase:       LimiteMensajeDeUso,
			descripcion: "mensaje del límite de uso: You've hit your session limit · resets 5pm",
		},
		{
			nombre:      "reintentos-agotados",
			transcript:  reintentosAgotados + mensajeResultConError(t, textoDelError429),
			codigo:      1,
			clase:       LimiteReintentosAgotados,
			descripcion: "reintentos por rate_limit agotados",
		},
		{
			nombre:     "terminada-tras-el-ultimo-reintento",
			transcript: reintentosAgotados + mensajeAssistantDeTexto + mensajeResultCorrecto,
			clase:      SinLimite,
			terminada:  true,
		},
		{
			nombre:     "reintentos-por-agotar",
			transcript: mensajeDeReintento(3, 10, 429, "rate_limit") + mensajeResultConError(t, textoDelError429),
			codigo:     1,
			clase:      SinLimite,
		},
		{
			nombre:     "reintentos-agotados-por-sobrecarga",
			transcript: mensajeDeReintento(10, 10, 529, "overloaded") + mensajeResultConError(t, textoDelError529),
			codigo:     1,
			clase:      SinLimite,
		},
		{
			nombre: "reintentos-agotados-y-despues-por-sobrecarga",
			transcript: mensajeDeReintento(10, 10, 429, "rate_limit") + mensajeDeReintento(1, 10, 529, "overloaded") +
				mensajeResultConError(t, textoDelError529),
			codigo: 1,
			clase:  SinLimite,
		},
		{
			nombre:      "reintentos-agotados-antes-que-cortada",
			transcript:  reintentosAgotados,
			codigo:      124,
			clase:       LimiteReintentosAgotados,
			descripcion: "reintentos por rate_limit agotados",
		},
		{
			nombre:      "cortada-durante-reintentos",
			transcript:  cortadaEnReintentos,
			codigo:      124,
			clase:       LimiteCortadaDuranteReintentos,
			descripcion: "cortada por el tope durante reintentos por rate_limit",
		},
		{
			nombre:      "cortada-con-kill-durante-reintentos",
			transcript:  cortadaEnReintentos,
			codigo:      137,
			clase:       LimiteCortadaDuranteReintentos,
			descripcion: "cortada por el tope durante reintentos por rate_limit",
		},
		{
			nombre:     "cortada-durante-reintentos-por-sobrecarga",
			transcript: mensajeDeReintento(1, 10, 429, "rate_limit") + mensajeDeReintento(2, 10, 529, "overloaded"),
			codigo:     124,
			clase:      SinLimite,
		},
		{
			nombre:     "cortada-tras-reintentos",
			transcript: cortadaEnReintentos + mensajeAssistantDeTexto,
			codigo:     124,
			clase:      SinLimite,
		},
		{
			nombre:     "terminada-tras-reintentos",
			transcript: cortadaEnReintentos + mensajeAssistantDeTexto + mensajeResultCorrecto,
			clase:      SinLimite,
			terminada:  true,
		},
		{
			nombre:     "sin-reintentos",
			transcript: mensajeResultCorrecto,
			clase:      SinLimite,
			terminada:  true,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			sesion, err := LeerSesion(escribirSesionConCodigo(t, mensajeInit+caso.transcript, caso.codigo))
			require.NoError(t, err)
			require.Equal(t, caso.terminada, sesion.Terminada, "premisa: la sesión termina, o no, con su result")

			limite := ClasificarElLimite(sesion)

			assert.Equal(t, LimiteDeUso{Clase: caso.clase, Descripcion: caso.descripcion}, limite)

			if caso.clase == SinLimite {
				assert.Empty(t, limite.Motivo(), "una sesión medida no lleva el motivo de las que no se miden")

				return
			}

			assert.Equal(t, "sin medir por límite de uso: "+caso.descripcion, limite.Motivo())
		})
	}
}
