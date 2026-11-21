package instalacion_test

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
)

// Las dos líneas nuevas con los binarios de las pruebas, v0.2.0 sobre una
// instalación v0.1.0, con los bytes de contracts/aviso-y-reinstalacion.md §3:
// la de FR-006, cuando reinstala, y la de FR-009, cuando lo ensaya.
const (
	reinstaladaDeLaV1 = "aviso: kitlegal v0.2.0 ha actualizado las skills globales, que eran de v0.1.0; " +
		"se aplican en la próxima conversación"
	ensayadaDeLaV1 = "aviso: kitlegal v0.2.0 se actualizarían las skills globales, que eran de v0.1.0; " +
		"se aplican en la próxima conversación"
)

// lineaEnsayada es la línea de FR-009 con la versión del binario y la que
// nombra la línea de hoy (data-model §5).
func lineaEnsayada(binario, nombra string) string {
	return "aviso: kitlegal " + binario + " se actualizarían las skills globales, que eran de " + nombra +
		"; se aplican en la próxima conversación"
}

// candadoDePrueba es el puerto TomarCandado sobre un candado en memoria: se
// toma si está libre, y liberar lo deja libre otra vez. Cuenta las veces que se
// pide y las que se libera; ocupado de partida es el candado de otra
// invocación, que no se suelta.
type candadoDePrueba struct {
	ocupado   bool
	pedidos   int
	liberados int
}

// tomar es el puerto TomarCandado: con el candado ocupado, no lo toma y
// liberar es nil, como pide data-model §3.
func (c *candadoDePrueba) tomar() (func(), bool) {
	c.pedidos++

	if c.ocupado {
		return nil, false
	}

	c.ocupado = true

	return func() {
		c.ocupado = false
		c.liberados++
	}, true
}

// reinstalacionEn es la Reinstalacion de una invocación sobre d con el binario
// de version, el HOME de las pruebas, sus skills empotradas, los nombres
// retirados de retiradas, un creador de enlaces que funciona en todas partes y
// el candado de prueba.
func reinstalacionEn(
	d *discoEnMemoria, version string, retiradas instalacion.Retiradas, candado *candadoDePrueba,
) instalacion.Reinstalacion {
	enlazador := nuevoEnlazadorHonesto(d, admiteSiempre)

	return instalacion.Reinstalacion{
		Disco:      d,
		Enlazador:  enlazador,
		Escritor:   &escritorEnMemoria{disco: d, enlazador: enlazador},
		Candado:    candado.tomar,
		Home:       homeDePrueba,
		Version:    version,
		Empotradas: empotradasDePrueba(),
		Retiradas:  retiradas,
	}
}

// instalarGlobal deja d como lo deja kitlegal skills install -g <nombres> del
// binario de version, sin --host: el pedido de ValidarInvocacion, el plan de
// Planificar sin ningún conflicto y Aplicar sin ningún fallo (FR-005).
func instalarGlobal(t *testing.T, d *discoEnMemoria, version string, retiradas instalacion.Retiradas, nombres ...string) {
	t.Helper()

	pedido, err := instalacion.ValidarInvocacion(
		instalacion.Invocacion{Skills: nombres, Global: true}, homeDePrueba, empotradasDePrueba())
	require.NoError(t, err, "la invocación de install -g")

	enlazador := nuevoEnlazadorHonesto(d, admiteSiempre)

	plan, err := instalacion.Planificar(d, enlazador, pedido, empotradasDePrueba(), retiradas, version)
	require.NoError(t, err, "install -g no encuentra ningún conflicto")

	_, err = instalacion.Aplicar(plan, &escritorEnMemoria{disco: d, enlazador: enlazador})
	require.NoError(t, err, "install -g aplica su plan")
}

// rutaGlobal es la ruta de rel bajo el HOME de las pruebas.
func rutaGlobal(rel string) string {
	return path.Join(homeDePrueba, rel)
}

// globalDeLaV1 son las dos skills de las pruebas en el global, como las dejó
// install -g de un binario v0.1.0 que empotraba otro SKILL.md de normativa,
// enlazadas en claude: el global sin tocar, con algo que reescribir. Cada caso
// cambia una cosa antes de escribir su manifiesto.
func globalDeLaV1(d *discoEnMemoria) *instalada {
	d.directorio(rutaGlobal(".claude/skills"))

	return enGlobal(d, versionDePrueba).
		deOtroBinario("normativa", "SKILL.md", "# normativa de la v0.1.0\n").
		enlazar("normativa").
		enlazar("plazos-y-territorio")
}

// retiradaPorNormativa es la lista de nombres retirados de los casos de la
// retirada: antigua, cuyas preguntas hereda normativa.
func retiradaPorNormativa() instalacion.Retiradas {
	return instalacion.Retiradas{"antigua": {"normativa"}}
}

// globalConAntigua es el global de un binario v0.1.0 con plazos-y-territorio,
// enlazada en claude, y antigua, de un nombre que este binario retira, intacta.
func globalConAntigua(d *discoEnMemoria) *instalada {
	d.directorio(rutaGlobal(".claude/skills"))

	return instalarEn(d, ambitoGlobal(d.t, homeDePrueba), "plazos-y-territorio").
		enlazar("plazos-y-territorio").
		deAntes("antigua")
}

// TestReglaDeVersiones fija la regla de versiones de FR-004 (FR-019;
// research.md D6 y D7) por lo único que se ve de ella, si reinstala: con
// Avisar(Ensayar), que no toma el candado ni escribe, sobre un global sin
// tocar de las versiones de cada fila. Da la línea de FR-009 si toda versión
// comparada —la del manifiesto y la de cada skill que declara y el binario
// empotra— tiene la forma de SemVer, ninguna va después de la del binario en
// la precedencia de SemVer 2.0.0 y alguna va antes; y si no, la de hoy, o
// ninguna si hoy no hay ninguna (FR-007). En cada fila, el disco queda igual y
// el candado no se pide.
func TestReglaDeVersiones(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre string
		// manifiesto es la versión del manifiesto y de cada skill que declara.
		manifiesto string
		// skill, si no es vacía, es la de plazos-y-territorio.
		skill   string
		binario string
		// ensaya dice si la línea es la de FR-009; si no, la de hoy.
		ensaya bool
		// nombra es la versión que nombra la línea de hoy; vacía, sin línea.
		nombra string
	}{
		{nombre: "una menor", manifiesto: "v0.1.0", binario: "v0.2.0", ensaya: true, nombra: "v0.1.0"},
		{nombre: "la v inicial no cuenta", manifiesto: "0.1.0", binario: "v0.2.0", ensaya: true, nombra: "0.1.0"},
		{nombre: "el parche", manifiesto: "0.2.0", binario: "0.2.1", ensaya: true, nombra: "0.2.0"},
		{nombre: "el núcleo, número a número", manifiesto: "1.9.9", binario: "2.0.0", ensaya: true, nombra: "1.9.9"},
		{nombre: "el núcleo por su valor y no por sus bytes", manifiesto: "0.9.0", binario: "0.10.0", ensaya: true, nombra: "0.9.0"},
		{nombre: "el núcleo por su valor, al revés", manifiesto: "0.10.0", binario: "0.9.0", nombra: "0.10.0"},
		{
			nombre: "números de cualquier longitud", manifiesto: "0.1.99999999999999999999",
			binario: "0.1.100000000000000000000", ensaya: true, nombra: "0.1.99999999999999999999",
		},
		{
			nombre: "números de cualquier longitud, al revés", manifiesto: "0.1.100000000000000000000",
			binario: "0.1.99999999999999999999", nombra: "0.1.100000000000000000000",
		},
		{nombre: "una prerelease es anterior a su final", manifiesto: "0.2.0-rc.1", binario: "0.2.0", ensaya: true, nombra: "0.2.0-rc.1"},
		{nombre: "una final no es anterior a su prerelease", manifiesto: "0.2.0", binario: "0.2.0-rc.1", nombra: "0.2.0"},
		{nombre: "una prerelease del núcleo siguiente", manifiesto: "0.1.0", binario: "0.2.0-rc.1", ensaya: true, nombra: "0.1.0"},
		{nombre: "más identificadores van después", manifiesto: "1.0.0-alpha", binario: "1.0.0-alpha.1", ensaya: true, nombra: "1.0.0-alpha"},
		{nombre: "menos identificadores van antes", manifiesto: "1.0.0-alpha.1", binario: "1.0.0-alpha", nombra: "1.0.0-alpha.1"},
		{nombre: "los numéricos por su valor", manifiesto: "1.0.0-2", binario: "1.0.0-11", ensaya: true, nombra: "1.0.0-2"},
		{nombre: "los numéricos por su valor, al revés", manifiesto: "1.0.0-11", binario: "1.0.0-2", nombra: "1.0.0-11"},
		{nombre: "un numérico antes que un alfanumérico", manifiesto: "1.0.0-1", binario: "1.0.0-alpha", ensaya: true, nombra: "1.0.0-1"},
		{nombre: "un alfanumérico después de un numérico", manifiesto: "1.0.0-alpha", binario: "1.0.0-1", nombra: "1.0.0-alpha"},
		{nombre: "los alfanuméricos por sus bytes ASCII", manifiesto: "1.0.0-Beta", binario: "1.0.0-alpha", ensaya: true, nombra: "1.0.0-Beta"},
		{nombre: "los alfanuméricos por sus bytes, al revés", manifiesto: "1.0.0-beta", binario: "1.0.0-alpha", nombra: "1.0.0-beta"},
		{nombre: "un guion dentro de la prerelease", manifiesto: "1.0.0-rc-1", binario: "1.0.0-rc-2", ensaya: true, nombra: "1.0.0-rc-1"},
		{nombre: "solo los metadatos de construcción", manifiesto: "0.1.0+abc", binario: "0.1.0", nombra: "0.1.0+abc"},
		{nombre: "solo los metadatos de construcción, al revés", manifiesto: "0.1.0", binario: "0.1.0+abc", nombra: "0.1.0"},
		{nombre: "los metadatos no ordenan una prerelease", manifiesto: "1.0.0-rc.1+b", binario: "1.0.0-rc.1+a", nombra: "1.0.0-rc.1+b"},
		{nombre: "los metadatos no ordenan, con una menor", manifiesto: "0.1.0+zzz", binario: "0.2.0+aaa", ensaya: true, nombra: "0.1.0+zzz"},
		{nombre: "un binario más viejo", manifiesto: "v0.2.0", binario: "v0.1.0", nombra: "v0.2.0"},
		{
			nombre: "una skill más nueva que el binario con el manifiesto más viejo", manifiesto: "v0.1.0", skill: "v0.3.0",
			binario: "v0.2.0", nombra: "v0.1.0",
		},
		{
			nombre: "una skill más vieja con el manifiesto al día", manifiesto: "v0.2.0", skill: "v0.1.0",
			binario: "v0.2.0", ensaya: true, nombra: "v0.1.0",
		},
		{nombre: "un manifiesto sin forma", manifiesto: "0.1", binario: "v0.2.0", nombra: "0.1"},
		{nombre: "una skill sin forma", manifiesto: "v0.1.0", skill: "desarrollo", binario: "v0.2.0", nombra: "v0.1.0"},
		{nombre: "un binario sin forma: hoy no avisa", manifiesto: "v0.1.0", binario: "desarrollo"},
		{nombre: "la misma versión: hoy no avisa", manifiesto: "0.2.0", binario: "v0.2.0"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			d := nuevoDiscoEnMemoria(t)
			d.directorio(temporal)

			global := enGlobal(d, caso.manifiesto)
			if caso.skill != "" {
				global.skillDeVersion("plazos-y-territorio", caso.skill)
			}

			global.escribir()

			antes := d.instantanea()
			candado := &candadoDePrueba{}

			linea, hay := reinstalacionEn(d, caso.binario, nil, candado).Avisar(instalacion.Ensayar)

			esperada := ""

			switch {
			case caso.nombra == "":
			case caso.ensaya:
				esperada = lineaEnsayada(caso.binario, caso.nombra)
			default:
				esperada = avisoGlobal(caso.nombra, caso.binario)
			}

			assert.Equal(t, esperada, linea)
			assert.Equal(t, esperada != "", hay, "hay línea si y solo si se da")
			assert.Equal(t, antes, d.instantanea(), "ensayar no escribe nada (FR-009, FR-013)")
			assert.Zero(t, candado.pedidos, "ensayar no pide el candado (FR-009)")
		})
	}
}

// casoDeReinstalacion es un estado del disco sobre el que se pide el aviso con
// un permiso y lo que tiene que salir: la línea, el disco y el candado.
type casoDeReinstalacion struct {
	nombre string
	// preparar deja el disco del caso, que empieza con el directorio de
	// trabajo y el temporal vacíos.
	preparar func(d *discoEnMemoria)
	// retiradas son los nombres que retira el binario; nil, ninguno.
	retiradas instalacion.Retiradas
	// permiso es el de la invocación; sin él, Reinstalar.
	permiso *instalacion.PermisoDelAviso
	// ocupado dice si otra invocación tiene el candado.
	ocupado bool
	// reinstala son las skills de install -g cuyo estado deja la
	// reinstalación; nil si el disco no cambia.
	reinstala []string
	// esperada es la línea; vacía, ninguna.
	esperada string
	// pideElCandado dice si se cumplen FR-003 (a) a (c) y la invocación pide el
	// candado.
	pideElCandado bool
}

// conPermiso es la dirección de un permiso, la del campo de un caso.
func conPermiso(permiso instalacion.PermisoDelAviso) *instalacion.PermisoDelAviso {
	return &permiso
}

// TestReinstalacion fija lo que hace el aviso con cada permiso sobre los
// estados del spec (contracts/aviso-y-reinstalacion.md §2; data-model §6;
// FR-003, FR-005 a FR-007, FR-009, FR-013, FR-014; SC-001, SC-002, SC-006),
// con el binario v0.2.0. En cada caso:
//
//   - la línea es exactamente la esperada, y hay línea si y solo si se da;
//   - el disco queda como lo deja kitlegal skills install -g de las skills
//     del pedido (Planificar y Aplicar sobre un disco gemelo) o sin cambios;
//   - el candado se pide si y solo si se cumplen FR-003 (a) a (c) y con
//     Reinstalar, y cada candado tomado se libera una vez;
//   - no se abre ni se lista nada que no sea un fichero regular o un
//     directorio real (FR-028 de H19);
//   - y si reinstaló, la consulta siguiente no da ninguna línea ni pide el
//     candado: las versiones ya son iguales.
func TestReinstalacion(t *testing.T) {
	t.Parallel()

	grupos := []struct {
		nombre string
		casos  []casoDeReinstalacion
	}{
		{nombre: "reinstala", casos: casosQueReinstalan()},
		{nombre: "lo tocado", casos: casosDeLoTocado()},
		{nombre: "el ámbito local", casos: casosDelAmbitoLocal()},
		{nombre: "el permiso y el candado", casos: casosDelPermisoYElCandado()},
	}

	for _, grupo := range grupos {
		t.Run(grupo.nombre, func(t *testing.T) {
			t.Parallel()

			for _, caso := range grupo.casos {
				t.Run(caso.nombre, func(t *testing.T) {
					t.Parallel()

					probarCasoDeReinstalacion(t, caso)
				})
			}
		})
	}
}

// probarCasoDeReinstalacion pide el aviso sobre el disco del caso y lo compara
// con lo esperado.
func probarCasoDeReinstalacion(t *testing.T, caso casoDeReinstalacion) {
	t.Helper()

	d := nuevoDiscoEnMemoria(t)
	d.directorio(temporal)
	caso.preparar(d)

	esperado := d.instantanea()

	if caso.reinstala != nil {
		gemelo := nuevoDiscoEnMemoria(t)
		gemelo.directorio(temporal)
		caso.preparar(gemelo)
		instalarGlobal(t, gemelo, versionNueva, caso.retiradas, caso.reinstala...)

		esperado = gemelo.instantanea()
	}

	permiso := instalacion.Reinstalar
	if caso.permiso != nil {
		permiso = *caso.permiso
	}

	candado := &candadoDePrueba{ocupado: caso.ocupado}

	linea, hay := reinstalacionEn(d, versionNueva, caso.retiradas, candado).Avisar(permiso)

	assert.Equal(t, caso.esperada, linea)
	assert.Equal(t, caso.esperada != "", hay, "hay línea si y solo si se da")
	assert.Equal(t, esperado, d.instantanea(), "el disco, el de install -g del pedido o sin cambios (FR-005, FR-013)")
	exigirDiscoRespetado(t, d, directorioDeTrabajo)
	exigirDiscoRespetado(t, d, homeDePrueba)

	pedidos := 0
	if caso.pideElCandado {
		pedidos = 1
	}

	assert.Equal(t, pedidos, candado.pedidos, "el candado, solo con FR-003 (a) a (c) y Reinstalar (FR-010)")

	if !caso.ocupado {
		assert.Equal(t, candado.pedidos, candado.liberados, "cada candado tomado se libera una vez")
	}

	if caso.reinstala == nil {
		return
	}

	siguiente := &candadoDePrueba{}

	linea, hay = reinstalacionEn(d, versionNueva, caso.retiradas, siguiente).Avisar(instalacion.Reinstalar)

	assert.Empty(t, linea, "la consulta siguiente no da ninguna línea (SC-001)")
	assert.False(t, hay)
	assert.Equal(t, esperado, d.instantanea(), "la consulta siguiente no escribe nada")
	assert.Zero(t, siguiente.pedidos, "la consulta siguiente no pide el candado")
}

// todasLasDePrueba son las dos skills empotradas de las pruebas, el pedido de
// un global que las declara a las dos.
func todasLasDePrueba() []string {
	return []string{"normativa", "plazos-y-territorio"}
}

// casosQueReinstalan son los estados en los que se cumplen FR-003 (a) a (d):
// la reinstalación deja el global como install -g del pedido y da la línea de
// FR-006 (US1).
func casosQueReinstalan() []casoDeReinstalacion {
	return []casoDeReinstalacion{
		{
			nombre:        "el global sin tocar",
			preparar:      func(d *discoEnMemoria) { globalDeLaV1(d).escribir() },
			reinstala:     todasLasDePrueba(),
			esperada:      reinstaladaDeLaV1,
			pideElCandado: true,
		},
		{
			nombre: "un fichero que falta no es un conflicto",
			preparar: func(d *discoEnMemoria) {
				globalDeLaV1(d).escribir()
				d.retirar(rutaGlobal(".agents/skills/plazos-y-territorio/references/leyes_vertebrales.md"))
			},
			reinstala:     todasLasDePrueba(),
			esperada:      reinstaladaDeLaV1,
			pideElCandado: true,
		},
		{
			nombre: "una skill de otra versión con el manifiesto al día",
			preparar: func(d *discoEnMemoria) {
				d.directorio(rutaGlobal(".claude/skills"))
				enGlobal(d, versionNueva).enlazar("normativa").skillDeVersion("normativa", versionDePrueba).escribir()
			},
			reinstala:     todasLasDePrueba(),
			esperada:      reinstaladaDeLaV1,
			pideElCandado: true,
		},
		{
			nombre:        "una retirada intacta se retira y entra su heredera",
			preparar:      func(d *discoEnMemoria) { globalConAntigua(d).escribir() },
			retiradas:     retiradaPorNormativa(),
			reinstala:     todasLasDePrueba(),
			esperada:      reinstaladaDeLaV1,
			pideElCandado: true,
		},
		{
			nombre: "una heredera ya declarada se pide una vez",
			preparar: func(d *discoEnMemoria) {
				globalDeLaV1(d).deAntes("antigua").escribir()
			},
			retiradas:     retiradaPorNormativa(),
			reinstala:     todasLasDePrueba(),
			esperada:      reinstaladaDeLaV1,
			pideElCandado: true,
		},
		{
			nombre: "solo lo declarado: una empotrada que no se declara no se instala",
			preparar: func(d *discoEnMemoria) {
				instalarEn(d, ambitoGlobal(d.t, homeDePrueba), "plazos-y-territorio").escribir()
			},
			reinstala:     []string{"plazos-y-territorio"},
			esperada:      reinstaladaDeLaV1,
			pideElCandado: true,
		},
	}
}

// casosDeLoTocado son los estados con algo que install -g no reescribiría sin
// --force, o en los que no se cumple FR-004: la línea de hoy, nada escrito y
// el candado sin pedir (US2; FR-005, FR-007; SC-002).
func casosDeLoTocado() []casoDeReinstalacion {
	hoy := avisoGlobal(versionDePrueba, versionNueva)

	return []casoDeReinstalacion{
		{
			nombre: "un fichero editado",
			preparar: func(d *discoEnMemoria) {
				globalDeLaV1(d).escribir()
				d.fichero(rutaGlobal(".agents/skills/normativa/SKILL.md"), "# normativa, editada\n")
			},
			esperada: hoy,
		},
		{
			nombre: "un fichero ajeno",
			preparar: func(d *discoEnMemoria) {
				globalDeLaV1(d).
					olvidar("plazos-y-territorio", "plazos-y-territorio/references/jerarquia_normativa.md").
					escribir()
			},
			esperada: hoy,
		},
		{
			nombre: "una entrada de host ajena",
			preparar: func(d *discoEnMemoria) {
				globalDeLaV1(d).escribir()
				d.retirar(rutaGlobal(".claude/skills/normativa"))
				d.fichero(temporal+"/fuera/SKILL.md", "# de otro\n")
				d.enlace(rutaGlobal(".claude/skills/normativa"), temporal+"/fuera")
			},
			esperada: hoy,
		},
		{
			nombre: "una retirada «sin retirar»",
			preparar: func(d *discoEnMemoria) {
				globalConAntigua(d).escribir()
				d.fichero(rutaGlobal(".agents/skills/antigua/SKILL.md"), "# antigua, editada\n")
			},
			retiradas: retiradaPorNormativa(),
			esperada:  hoy,
		},
		{
			nombre: "un binario más viejo",
			preparar: func(d *discoEnMemoria) {
				d.directorio(rutaGlobal(".claude/skills"))
				enGlobal(d, "v0.3.0").enlazar("normativa").escribir()
			},
			esperada: avisoGlobal("v0.3.0", versionNueva),
		},
		{
			nombre: "solo los metadatos de construcción",
			preparar: func(d *discoEnMemoria) {
				enGlobal(d, "0.2.0+abc").escribir()
			},
			esperada: avisoGlobal("0.2.0+abc", versionNueva),
		},
		{
			nombre: "ninguna skill que pedir",
			preparar: func(d *discoEnMemoria) {
				instalarEn(d, ambitoGlobal(d.t, homeDePrueba)).noEmpotrada("de-otra-rama", versionDePrueba).escribir()
			},
			esperada: hoy,
		},
		{
			nombre:   "las versiones ya iguales",
			preparar: func(d *discoEnMemoria) { enGlobal(d, versionNueva).escribir() },
		},
	}
}

// casosDelAmbitoLocal son los de un manifiesto local, que el aviso compara
// antes que el global: el global no se reinstala nunca (US3; FR-003 a,
// FR-014; «Edge Cases»).
func casosDelAmbitoLocal() []casoDeReinstalacion {
	return []casoDeReinstalacion{
		{
			nombre: "un manifiesto local viejo: la línea de hoy, sin -g",
			preparar: func(d *discoEnMemoria) {
				globalDeLaV1(d).escribir()
				enLocal(d, versionDePrueba).escribir()
			},
			esperada: avisoLocal(versionDePrueba, versionNueva),
		},
		{
			nombre: "un local al día con el global viejo: ninguna línea",
			preparar: func(d *discoEnMemoria) {
				globalDeLaV1(d).escribir()
				enLocal(d, versionNueva).escribir()
			},
		},
		{
			nombre: "un local ilegible con el global viejo: ninguna línea",
			preparar: func(d *discoEnMemoria) {
				globalDeLaV1(d).escribir()
				d.fichero(".agents/skills/kitlegal.json", "{")
			},
		},
		{
			nombre: "un ./.agents que no es un directorio con el global viejo: ninguna línea",
			preparar: func(d *discoEnMemoria) {
				globalDeLaV1(d).escribir()
				d.fichero(".agents", "no es un directorio")
			},
		},
	}
}

// casosDelPermisoYElCandado son los del global sin tocar con cada permiso que
// no reinstala y con el candado de otra invocación (FR-002, FR-007, FR-009,
// FR-010, FR-013; SC-006).
func casosDelPermisoYElCandado() []casoDeReinstalacion {
	sinTocar := func(d *discoEnMemoria) { globalDeLaV1(d).escribir() }
	hoy := avisoGlobal(versionDePrueba, versionNueva)

	return []casoDeReinstalacion{
		{
			nombre:   "solo avisar: la línea de hoy, sin candado",
			preparar: sinTocar,
			permiso:  conPermiso(instalacion.SoloAvisar),
			esperada: hoy,
		},
		{
			nombre:   "ensayar: la línea de FR-009, sin candado ni escritura",
			preparar: sinTocar,
			permiso:  conPermiso(instalacion.Ensayar),
			esperada: ensayadaDeLaV1,
		},
		{
			nombre: "ensayar con un fichero editado: la línea de hoy",
			preparar: func(d *discoEnMemoria) {
				globalDeLaV1(d).escribir()
				d.fichero(rutaGlobal(".agents/skills/normativa/SKILL.md"), "# normativa, editada\n")
			},
			permiso:  conPermiso(instalacion.Ensayar),
			esperada: hoy,
		},
		{
			nombre:   "solo avisar con las versiones iguales: ninguna línea",
			preparar: func(d *discoEnMemoria) { enGlobal(d, versionNueva).escribir() },
			permiso:  conPermiso(instalacion.SoloAvisar),
		},
		{
			nombre:        "el candado de otra invocación: la línea de hoy, sin cambios",
			preparar:      sinTocar,
			ocupado:       true,
			esperada:      hoy,
			pideElCandado: true,
		},
	}
}
