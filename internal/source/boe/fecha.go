package boe

import (
	"context"
	"fmt"
	"time"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// La lectura fechada de articulo, la de --fecha (H20; contracts/articulo-fecha.md
// §2 a §5; data-model.md §3 a §5; research.md D4 a D8): pide lo mismo que la de
// hoy, guarda todas las versiones del bloque en su propia entrada y elige entre
// ellas la que regía el día pedido. refs/boe.py no la tiene: solo da la última
// versión (nota [5]).
const (
	// motivoSinRedaccionVigente es el del 3 de un día anterior a la primera
	// vigencia del bloque: nombra el bloque, la norma, el día pedido y la
	// fecha_vigencia de la primera versión (FR-003; research.md D7).
	motivoSinRedaccionVigente = "el bloque %s de %s no tenía redacción vigente el %s: " +
		"su primera redacción tiene fecha de vigencia %s"
	// motivoSinFechasDeVigencia es el del 3 de un bloque sin versiones o con
	// alguna sin fecha_vigencia, para cualquier día pedido (FR-004; research.md
	// D7). Su forma lo distingue del de motivoSinRedaccionVigente y del de un
	// bloque que no está en la norma.
	motivoSinFechasDeVigencia = "el BOE no da la fecha de vigencia de cada redacción del bloque %s de %s: " +
		"no se puede saber cuál tenía el %s"
	// formaDeLaVigencia es la de fecha_vigencia, AAAAMMDD, como disposición de
	// time.Parse: lo que no se lee con ella no es un día del calendario.
	formaDeLaVigencia = "20060102"
)

// articuloEnUnaFecha resuelve articulo <norma> <bloque> --fecha <fecha>, los
// tres ya validados: dentro de invocar, las versiones del bloque con
// versionesDeLaLecturaFechada y, entre ellas, la del día pedido con
// elegirLaVersion. El resultado lleva la dirección del bloque en éxito, en
// ensayo, en los fallos del bloque y en los dos de la elección, y la de los
// metadatos en los suyos, como la lectura sin fecha (FR-007); en éxito y en los
// fallos de la elección, la fecha de consulta de las versiones. No emite nada al
// grafo.
func (f *Fuente) articuloEnUnaFecha(ctx context.Context, ec schema.Contexto, norma, bloque, fecha string,
) (schema.Resultado, error) {
	direccion := direccionDelBloque(norma, bloque)

	return f.invocar(ctx, ec, direccion, func(ctx context.Context, en *invocacion) (schema.Resultado, error) {
		versiones, err := versionesDeLaLecturaFechada(ctx, en, norma, bloque)

		switch {
		case err != nil:
			return resultadoDelError(err), err
		case len(versiones.ensayo) > 0:
			return resultadoDeLaConsulta(direccion, versiones), nil
		}

		articulo, err := elegirLaVersion(norma, bloque, fecha, versiones)
		if err != nil {
			return resultadoDelError(err), err
		}

		return resultadoDeLaConsulta(direccion,
			consultaResuelta[ArticuloFechado]{datos: articulo, fechaConsulta: versiones.fechaConsulta}), nil
	})
}

// versionesDeLaLecturaFechada son las versiones de un bloque de una norma, los
// dos ya validados, resueltas por resolverElBloque con su propia entrada
// (claveDeLasVersiones) y los mismos dos recursos que la lectura sin fecha, el
// bloque y los metadatos de su norma (FR-006): la entrada vigente se sirve sin
// pedir nada; sin ella, se piden y, solo si los dos se obtienen y se
// interpretan, se escribe antes de elegir, de modo que cualquier otro día del
// bloque se sirve después de ella (FR-007, FR-022). Nunca lee ni escribe la
// entrada de hoy del artículo (FR-020, FR-021).
func versionesDeLaLecturaFechada(ctx context.Context, en *invocacion, norma, bloque string,
) (consultaResuelta[versionesDelBloque], error) {
	return resolverElBloque(ctx, en, claveDeLasVersiones(norma, bloque), &normaDeLaInvocacion{identificador: norma},
		bloque, componerLasVersiones)
}

// componerLasVersiones compone los datos de la entrada de las versiones con lo
// leído del bloque —su título, su tipo y todas sus versiones, en su orden— y los
// avisos y el ELI de los metadatos de su norma (data-model.md §5).
func componerLasVersiones(leido bloqueLeido, metadatos Metadatos) versionesDelBloque {
	versiones := make([]versionGuardada, 0, len(leido.versiones))
	for _, version := range leido.versiones {
		versiones = append(versiones, versionGuardada{
			FechaVersion:      version.fechaVersion,
			FechaVigencia:     version.fechaVigencia,
			NormaModificadora: version.normaModificadora,
			Texto:             version.texto,
		})
	}

	return versionesDelBloque{
		Titulo:    leido.titulo,
		Tipo:      leido.tipo,
		Versiones: versiones,
		Avisos:    metadatos.Avisos,
		URLELI:    metadatos.URLELI,
	}
}

// elegirLaVersion compone el data de la lectura fechada con la versión del
// bloque que regía el día pedido, ya validado (data-model.md §3; research.md D4):
//
//  1. si el bloque no tiene versiones, o alguna no tiene una fecha_vigencia que
//     sea un día AAAAMMDD, no se puede situar ninguna: 3 con el mensaje de
//     FR-004, sea cual sea el día;
//  2. si no, la elegida es la última, en el orden del bloque, cuya fecha_vigencia
//     no es posterior al día pedido, comparadas como días del calendario y nunca
//     por fecha_publicacion (FR-002); el día en que una versión entra en vigor, es
//     esa; si no hay ninguna, 3 con el mensaje de FR-003;
//  3. y la redacción posterior es la última versión del bloque cuando no es la
//     elegida (FR-010, FR-013).
//
// Los dos 3 llevan la dirección del bloque y la fecha de consulta de las
// versiones, y ninguno lleva texto (research.md D7). El artículo de la elegida lo
// compone componerArticulo, como la lectura sin fecha el de la última: con los
// avisos y el ELI de la norma hoy, los guardados con las versiones (FR-011).
func elegirLaVersion(norma, bloque, fecha string, guardadas consultaResuelta[versionesDelBloque],
) (ArticuloFechado, error) {
	direccion := direccionDelBloque(norma, bloque)
	versiones := guardadas.datos.Versiones

	dia, err := time.Parse(time.DateOnly, fecha)
	if err != nil {
		// ValidarFecha ya la aceptó: no es un fallo de quien invoca.
		return ArticuloFechado{}, errorInesperado(direccion, guardadas.fechaConsulta, err,
			fmt.Sprintf("la fecha %q, ya validada, no se puede leer", fecha))
	}

	vigencias, situable := vigenciasDeLasVersiones(versiones)
	if !situable {
		return ArticuloFechado{}, errorDeNoEncontrado(direccion, guardadas.fechaConsulta, nil,
			fmt.Sprintf(motivoSinFechasDeVigencia, bloque, norma, fecha))
	}

	elegida := -1

	for posicion, vigencia := range vigencias {
		if !vigencia.After(dia) {
			elegida = posicion
		}
	}

	if elegida < 0 {
		return ArticuloFechado{}, errorDeNoEncontrado(direccion, guardadas.fechaConsulta, nil,
			fmt.Sprintf(motivoSinRedaccionVigente, bloque, norma, fecha, versiones[0].FechaVigencia))
	}

	version := versiones[elegida]
	leido := bloqueLeido{
		titulo:            guardadas.datos.Titulo,
		tipo:              guardadas.datos.Tipo,
		fechaVersion:      version.FechaVersion,
		fechaVigencia:     version.FechaVigencia,
		normaModificadora: version.NormaModificadora,
		texto:             version.Texto,
	}
	// componerArticulo solo toma de los metadatos los avisos y el ELI.
	delDia := Metadatos{Avisos: guardadas.datos.Avisos, URLELI: guardadas.datos.URLELI}

	fechado := ArticuloFechado{Articulo: componerArticulo(norma, bloque, leido, delDia), FechaPedida: fecha}

	if ultima := versiones[len(versiones)-1]; elegida != len(versiones)-1 {
		fechado.RedaccionPosterior = &RedaccionPosterior{
			FechaVigencia:     ultima.FechaVigencia,
			NormaModificadora: ultima.NormaModificadora,
		}
	}

	return fechado, nil
}

// vigenciasDeLasVersiones son las fechas de vigencia de las versiones, en su
// orden, como días del calendario, y dice si se pueden situar todas: un bloque
// sin versiones, o con alguna cuya fecha_vigencia falta, viene vacía o no es un
// día AAAAMMDD, no se puede situar (FR-004).
func vigenciasDeLasVersiones(versiones []versionGuardada) ([]time.Time, bool) {
	if len(versiones) == 0 {
		return nil, false
	}

	vigencias := make([]time.Time, 0, len(versiones))

	for _, version := range versiones {
		vigencia, err := time.Parse(formaDeLaVigencia, version.FechaVigencia)
		if err != nil {
			return nil, false
		}

		vigencias = append(vigencias, vigencia)
	}

	return vigencias, true
}
