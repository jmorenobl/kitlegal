package evals

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Las dos fechas de vigencia de la consulta repetida del quickstart: la de la
// redacción superada, la que dejó en el grafo la primera lectura, y la de la
// leída (FR-061; data-model §5).
const (
	fechaSuperadaDelQuickstart = "20180309"
	fechaLeidaDelQuickstart    = "20200206"
)

// Las líneas de las condiciones de la consulta repetida que no dependen de la
// respuesta, con el texto de contracts/comprobacion-del-quickstart.md §2.
const (
	primeraSinLaForma = "la primera respuesta no lleva \xe2\x9a\xa0 REDACCI\xc3\x93N MODIFICADA: en una l\xc3\xadnea con " +
		"20180309 y 20200206"
	segundaConLaForma = "la segunda respuesta lleva \xe2\x9a\xa0 REDACCI\xc3\x93N MODIFICADA:"
)

// TestCondicionesDeLaConsultaRepetida fija las condiciones con las que la
// comprobación del quickstart juzga las dos respuestas de la consulta repetida
// (contracts/comprobacion-del-quickstart.md §2 y §3; FR-061; US5.3), con la
// lista de expresiones prohibidas del repositorio y sesiones construidas aquí:
// una línea por lo que falla, en el orden del contrato, y ninguna si la primera
// lleva la forma de version-obsoleta en una línea con las dos fechas como
// palabras, la segunda no la lleva y ninguna lleva expresiones de la lista. Una
// conversación sin terminar se nombra con su motivo y su respuesta no se mira.
func TestCondicionesDeLaConsultaRepetida(t *testing.T) {
	t.Parallel()

	conjunto, err := LeerConjunto(evalsDelRepositorio)
	require.NoError(t, err)

	lista := listaDelRepositorio(t, conjunto)

	// La primera respuesta buena empieza por lo que se pregunta y traslada el
	// cambio con su forma y sus dos fechas; la segunda, sin cambio que trasladar,
	// es la respuesta con la cita.
	primeraBuena := respuestaSinLaCita118 + "\n\n" + redaccionModificadaDeLaLCSP +
		"\n\n[BOE-A-2017-12902, bloque a1-30]"
	segundaBuena := respuestaConLaCita118

	// Sin terminar, la respuesta de cada una fallaría todas sus condiciones: sin
	// la forma y con expresiones la primera, y con la forma y con expresiones la
	// segunda.
	primeraSinTerminar := sinTerminar(transicionDeLaMemoria + "\n\n" + respuestaConLaCita118)
	segundaSinTerminar := sinTerminar(redaccionModificadaDeLaLCSP + "\n\n" + loDichoEnOtraConversacion)

	// Desde H7.4, las palabras de la forma fija de la línea fuera de ella son de
	// la lista (contracts/lista-de-expresiones.md §1 y §3 de H7.4): la línea con
	// otras palabras, o con unas fechas que no son las de la forma, las lleva.
	conLaLineaFueraDeSuForma := "la primera respuesta lleva expresiones prohibidas: se consult\xc3\xb3 antes"

	casos := []struct {
		nombre           string
		primera, segunda Sesion
		lineas           []string
	}{
		{
			nombre:  "las-dos-buenas",
			primera: sesionTerminada(true, primeraBuena),
			segunda: sesionTerminada(true, segundaBuena),
		},
		{
			nombre:  "la-primera-sin-la-forma",
			primera: sesionTerminada(true, respuestaConLaCita118),
			segunda: sesionTerminada(true, segundaBuena),
			lineas:  []string{primeraSinLaForma},
		},
		{
			nombre: "la-forma-en-una-linea-y-las-fechas-en-otra",
			primera: sesionTerminada(true, "\xe2\x9a\xa0 REDACCI\xc3\x93N MODIFICADA: la redacci\xc3\xb3n que se "+
				"consult\xc3\xb3 antes ha sido sustituida.\nLa de 20180309 ha pasado a ser la de 20200206.\n\n"+
				respuestaConLaCita118),
			segunda: sesionTerminada(true, segundaBuena),
			lineas:  []string{primeraSinLaForma, conLaLineaFueraDeSuForma},
		},
		{
			nombre: "la-forma-con-una-sola-fecha-en-su-linea",
			primera: sesionTerminada(true, "\xe2\x9a\xa0 REDACCI\xc3\x93N MODIFICADA: la redacci\xc3\xb3n con "+
				"fecha de vigencia 20180309 ha sido sustituida.\nLa que se cita es la de 20200206.\n\n"+
				respuestaConLaCita118),
			segunda: sesionTerminada(true, segundaBuena),
			lineas:  []string{primeraSinLaForma},
		},
		{
			// Las fechas cuentan como palabras: dentro de otras cifras no están.
			nombre: "las-fechas-dentro-de-otras-cifras",
			primera: sesionTerminada(true, "\xe2\x9a\xa0 REDACCI\xc3\x93N MODIFICADA: la redacci\xc3\xb3n con "+
				"fecha de vigencia 201803091, la que se consult\xc3\xb3 antes, ha sido sustituida por la de "+
				"202002060, que es la que se cita.\n\n"+respuestaConLaCita118),
			segunda: sesionTerminada(true, segundaBuena),
			lineas:  []string{primeraSinLaForma, conLaLineaFueraDeSuForma},
		},
		{
			nombre:  "la-segunda-con-la-forma",
			primera: sesionTerminada(true, primeraBuena),
			segunda: sesionTerminada(true, primeraBuena),
			lineas:  []string{segundaConLaForma},
		},
		{
			nombre:  "una-expresion-en-la-primera",
			primera: sesionTerminada(true, "La comprobaci\xc3\xb3n con graph check no ha dado nada.\n\n"+primeraBuena),
			segunda: sesionTerminada(true, segundaBuena),
			lineas:  []string{"la primera respuesta lleva expresiones prohibidas: graph check"},
		},
		{
			nombre:  "una-expresion-en-la-segunda",
			primera: sesionTerminada(true, primeraBuena),
			segunda: sesionTerminada(true, loDichoEnOtraConversacion+"\n\n"+segundaBuena),
			lineas:  []string{"la segunda respuesta lleva expresiones prohibidas: te confirm\xc3\xa9"},
		},
		{
			nombre:  "la-primera-sin-terminar",
			primera: primeraSinTerminar,
			segunda: sesionTerminada(true, segundaBuena),
			lineas:  []string{"la primera conversaci\xc3\xb3n no termin\xc3\xb3: tope de 240 s agotado (c\xc3\xb3digo 124)"},
		},
		{
			nombre:  "la-segunda-sin-terminar",
			primera: sesionTerminada(true, primeraBuena),
			segunda: segundaSinTerminar,
			lineas:  []string{"la segunda conversaci\xc3\xb3n no termin\xc3\xb3: tope de 240 s agotado (c\xc3\xb3digo 124)"},
		},
		{
			// La primera sin la forma y con tres expresiones, separadas por «, »,
			// y la segunda con la forma: las líneas van por condición, no por
			// conversación.
			nombre:  "tres-fallos-a-la-vez",
			primera: sesionTerminada(true, transicionDeLaMemoria+"\n\n"+respuestaConLaCita118),
			segunda: sesionTerminada(true, primeraBuena),
			lineas: []string{
				primeraSinLaForma,
				segundaConLaForma,
				"la primera respuesta lleva expresiones prohibidas: memoria de consultas, hallazgos, " +
					"tengo todo lo necesario",
			},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			lineas := comprobarConsultaRepetida(caso.primera, caso.segunda, lista,
				fechaSuperadaDelQuickstart, fechaLeidaDelQuickstart)

			assert.Equal(t, caso.lineas, lineas)
		})
	}
}

// sinTerminar es la sesión con la respuesta dada que el tope cortó después de
// su mensaje result: tiene respuesta, pero no terminó.
func sinTerminar(respuesta string) Sesion {
	sesion := sesionTerminada(true, respuesta)
	sesion.Codigo = codigoDelTope
	sesion.Terminada = false
	sesion.Cortada = true
	sesion.MotivoSinTerminar = motivoDelTope

	return sesion
}
