package territorio

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
	"slices"
	"time"

	"go.yaml.in/yaml/v3"

	"github.com/jmorenobl/kitlegal/internal/core/ids"
)

// Dónde vive cada fichero congelado, relativo a la raíz del repositorio. Son
// los nombres con los que Cargar señala cada defecto y, de lo que fija la
// configuración, el source que emite el territorio resuelto (FR-005,
// data-model §2.3 y §2.5).
const (
	rutaDeMunicipios          = "data/territorio/municipios.yaml"
	rutaDeDIR3                = "data/territorio/dir3.yaml"
	rutaDelEstado             = "data/territorio/estado.yaml"
	carpetaDeComunidades      = "data/territorio/comunidades/"
	extensionDeLasComunidades = ".yaml"
)

// Los dos regímenes de una comunidad (data-model §2.2).
const (
	regimenComun = "comun"
	regimenForal = "foral"
)

// Fuentes son los cuatro ficheros congelados de data/territorio/ tal como
// están escritos, ya leídos por quien los empaqueta: el dominio recibe bytes y
// no un sistema de ficheros (research.md D3).
type Fuentes struct {
	// Municipios es data/territorio/municipios.yaml.
	Municipios []byte
	// DIR3 es data/territorio/dir3.yaml.
	DIR3 []byte
	// Estado es data/territorio/estado.yaml.
	Estado []byte
	// Comunidades son los ficheros de data/territorio/comunidades/, indexados
	// por el código que da nombre a cada uno.
	Comunidades map[string][]byte
}

// Ficheros son los cuatro ficheros ya decodificados, en la forma que validan
// sus esquemas (data-model §3). Sus tipos se exportan para que quien valida
// los ficheros del repositorio contra su esquema lo haga sin repetirlos
// (FR-044, research.md D27).
type Ficheros struct {
	// Municipios es la relación de municipios del INE.
	Municipios FicheroDeMunicipios
	// DIR3 es la correspondencia verificada INE→DIR3.
	DIR3 FicheroDeDIR3
	// Estado es lo nacional que no es de ninguna comunidad.
	Estado FicheroDeEstado
	// Comunidades son las comunidades y ciudades autónomas, por código.
	Comunidades map[string]FicheroDeComunidad
}

// FicheroDeMunicipios es data/territorio/municipios.yaml, que valida
// schemas/territorio-municipios.yaml.json (data-model §3.1).
type FicheroDeMunicipios struct {
	// Fecha es la de referencia de la relación del INE, AAAA-MM-DD.
	Fecha string `yaml:"fecha"`
	// Source es el identificador de la fila de docs/SOURCES.md.
	Source string `yaml:"source"`
	// Municipios son las filas, indexadas por código INE de cinco cifras.
	Municipios map[string]FilaDeMunicipio `yaml:"municipios"`
}

// FilaDeMunicipio es un municipio de la relación del INE: con su código, que es
// la clave de la fila, los cinco datos que exige FR-040.
type FilaDeMunicipio struct {
	// DC es el dígito de control oficial, una cifra.
	DC string `yaml:"dc"`
	// Nombre es el nombre oficial.
	Nombre string `yaml:"nombre"`
	// Provincia es el código de su provincia, dos cifras.
	Provincia string `yaml:"provincia"`
	// Comunidad es el código de su comunidad, dos cifras.
	Comunidad string `yaml:"comunidad"`
}

// FicheroDeDIR3 es data/territorio/dir3.yaml, que valida
// schemas/territorio-dir3.yaml.json (data-model §3.2).
type FicheroDeDIR3 struct {
	// Fecha es la del volcado del REL, AAAA-MM-DD.
	Fecha string `yaml:"fecha"`
	// Source es el identificador de la fila de docs/SOURCES.md.
	Source string `yaml:"source"`
	// Correspondencia lleva del código INE al DIR3 de su ayuntamiento, solo
	// en las filas verificadas (FR-048).
	Correspondencia map[string]string `yaml:"correspondencia"`
}

// FicheroDeEstado es data/territorio/estado.yaml, que valida
// schemas/territorio-estado.yaml.json (data-model §3.3).
type FicheroDeEstado struct {
	// Fecha es la de su redacción, AAAA-MM-DD.
	Fecha string `yaml:"fecha"`
	// Source es el del propio fichero.
	Source string `yaml:"source"`
	// Boletin es el boletín estatal.
	Boletin BoletinDelEstado `yaml:"boletin"`
}

// BoletinDelEstado es el boletín estatal que fija estado.yaml.
type BoletinDelEstado struct {
	// Codigo es el del boletín.
	Codigo string `yaml:"codigo"`
	// Nombre es el oficial.
	Nombre string `yaml:"nombre"`
	// URL es su dirección pública.
	URL string `yaml:"url"`
}

// FicheroDeComunidad es data/territorio/comunidades/<código>.yaml, que valida
// schemas/territorio-comunidad.yaml.json (data-model §3.4).
type FicheroDeComunidad struct {
	// Fecha es la de su redacción, AAAA-MM-DD.
	Fecha string `yaml:"fecha"`
	// Source es el de sus nombres.
	Source string `yaml:"source"`
	// Codigo es el de la comunidad, el mismo que da nombre al fichero.
	Codigo string `yaml:"codigo"`
	// Nombre es el oficial.
	Nombre string `yaml:"nombre"`
	// Regimen es comun o foral, dato nacional de las 19 (FR-055).
	Regimen string `yaml:"regimen"`
	// Provincias lleva del código de cada provincia a su nombre.
	Provincias map[string]string `yaml:"provincias"`
	// Boletines son los configurados, o nil si la comunidad no tiene ninguno:
	// ausente es territorio no configurado (FR-051, FR-053).
	Boletines *BoletinesConfigurados `yaml:"boletines"`
}

// BoletinesConfigurados son los boletines que una comunidad tiene
// configurados; un nivel ausente no lo está.
type BoletinesConfigurados struct {
	// Autonomico es el boletín autonómico, o nil.
	Autonomico *BoletinConfigurado `yaml:"autonomico"`
	// Provincial es el boletín provincial, o nil.
	Provincial *BoletinConfigurado `yaml:"provincial"`
}

// BoletinConfigurado es un boletín que fija la configuración de una
// comunidad.
type BoletinConfigurado struct {
	// Codigo es el del boletín.
	Codigo string `yaml:"codigo"`
	// Nombre es el oficial.
	Nombre string `yaml:"nombre"`
	// URL es su dirección pública.
	URL string `yaml:"url"`
	// Motivo es la razón por la que un mismo boletín cubre dos niveles, o
	// vacío (FR-052).
	Motivo string `yaml:"motivo"`
}

// Cargar decodifica las fuentes, comprueba su forma y su integridad y
// construye el registro (data-model §2.1).
//
// Cada fichero tiene que leerse como un documento YAML con las claves de su
// forma y ninguna más, con su fecha AAAA-MM-DD y su source, y cada dato que el
// territorio resuelto va a emitir tiene que estar: un código INE y un DIR3
// que la gramática de internal/core/ids acepta, un dígito de control de una
// cifra, los nombres y los boletines configurados completos. Y entre ficheros,
// los seis puntos de integridad:
//
//  1. toda provincia citada por un municipio la declara una comunidad, y
//     ninguna la declaran dos;
//  2. todo municipio de la correspondencia está en la relación, y su DIR3 es
//     coherente con su código INE y su dígito de control;
//  3. ningún código de municipio, de provincia ni de comunidad se repite —la
//     clave repetida se rechaza ya al leer cada fichero—;
//  4. el régimen de toda comunidad es comun o foral;
//  5. el fichero de cada comunidad declara dentro el código que le da nombre;
//  6. la comunidad de cada municipio es la de la comunidad que declara su
//     provincia, que es el camino autoritativo. Lo mismo, y por la misma
//     razón, la provincia de cada municipio es la de su código INE.
//
// Con cualquier defecto no hay registro: el error los dice todos, cada uno con
// su fichero y su dato, y no lleva clase de usuario (contrato del applet §7).
func Cargar(fuentes Fuentes) (*Registro, error) {
	ficheros, err := decodificar(fuentes)
	if err != nil {
		return nil, err
	}

	var c carga

	registro := c.construir(ficheros)
	if err := errors.Join(c.defectos...); err != nil {
		return nil, err
	}

	return registro, nil
}

// rutaDeComunidad es la ruta del fichero de una comunidad.
func rutaDeComunidad(codigo string) string {
	return carpetaDeComunidades + codigo + extensionDeLasComunidades
}

// decodificar lee cada fichero en su tipo y devuelve juntos los defectos de
// todos los que no se leen.
func decodificar(fuentes Fuentes) (Ficheros, error) {
	var defectos []error

	municipios, err := decodificarRelacion(fuentes.Municipios)
	defectos = append(defectos, err)

	dir3, err := decodificarCorrespondencia(fuentes.DIR3)
	defectos = append(defectos, err)

	estado, err := decodificarFichero[FicheroDeEstado](rutaDelEstado, fuentes.Estado)
	defectos = append(defectos, err)

	comunidades := make(map[string]FicheroDeComunidad, len(fuentes.Comunidades))

	for _, codigo := range slices.Sorted(maps.Keys(fuentes.Comunidades)) {
		comunidad, err := decodificarFichero[FicheroDeComunidad](rutaDeComunidad(codigo), fuentes.Comunidades[codigo])
		defectos = append(defectos, err)
		comunidades[codigo] = comunidad
	}

	if err := errors.Join(defectos...); err != nil {
		return Ficheros{}, err
	}

	return Ficheros{Municipios: municipios, DIR3: dir3, Estado: estado, Comunidades: comunidades}, nil
}

// decodificarFichero lee un fichero en su tipo. Una clave que el tipo no
// declara es un defecto, no algo que se descarta en silencio: una errata en el
// nombre de un nivel de boletín lo daría por no configurado. La clave repetida
// la rechaza el propio lector.
func decodificarFichero[T any](ruta string, contenido []byte) (T, error) {
	var fichero T

	if len(bytes.TrimSpace(contenido)) == 0 {
		return fichero, &defectoDeCarga{fichero: ruta, motivo: "está vacío"}
	}

	lector := yaml.NewDecoder(bytes.NewReader(contenido))
	lector.KnownFields(true)

	if err := lector.Decode(&fichero); err != nil {
		return fichero, &defectoDeCarga{fichero: ruta, motivo: "no se puede leer: " + err.Error()}
	}

	return fichero, nil
}

// relacionSinFilas y correspondenciaSinFilas son la cabecera de la relación y
// la de la correspondencia, con su mapa de filas en un nodo que el lector de
// YAML no decodifica: las filas se leen aparte (decodificarFicheroDeFilas). El
// resto del fichero sí lo decodifica el lector, con sus claves conocidas.
type (
	relacionSinFilas struct {
		Fecha      string    `yaml:"fecha"`
		Source     string    `yaml:"source"`
		Municipios yaml.Node `yaml:"municipios"`
	}
	correspondenciaSinFilas struct {
		Fecha           string    `yaml:"fecha"`
		Source          string    `yaml:"source"`
		Correspondencia yaml.Node `yaml:"correspondencia"`
	}
)

// decodificarRelacion lee data/territorio/municipios.yaml.
func decodificarRelacion(contenido []byte) (FicheroDeMunicipios, error) {
	relacion, filas, err := decodificarFicheroDeFilas[relacionSinFilas](
		rutaDeMunicipios, "municipios", contenido, leerFilaDeMunicipio)

	return FicheroDeMunicipios{Fecha: relacion.Fecha, Source: relacion.Source, Municipios: filas}, err
}

// decodificarCorrespondencia lee data/territorio/dir3.yaml.
func decodificarCorrespondencia(contenido []byte) (FicheroDeDIR3, error) {
	correspondencia, filas, err := decodificarFicheroDeFilas[correspondenciaSinFilas](
		rutaDeDIR3, "correspondencia", contenido, leerFilaDeDIR3)

	return FicheroDeDIR3{Fecha: correspondencia.Fecha, Source: correspondencia.Source, Correspondencia: filas}, err
}

// decodificarFilas decodifica con el lector de YAML el nodo de un mapa de
// filas, una por clave, y busca la clave repetida fila a fila (filasLeidas).
// Cada fila la decodifica el lector, que en ella sí busca la clave repetida:
// son unas pocas. Lo que no hace el lector al decodificar un nodo es rechazar
// dentro de la fila una clave que el tipo no declara; no se pierde nada,
// porque la carga exige cada dato de la fila —una columna mal escrita es una
// que falta— y la clave de más la rechaza el esquema del fichero en make ci
// (FR-044). Los defectos de todas las filas se dicen juntos, cada uno con su
// línea o con su clave.
func decodificarFilas[V any](ruta, clave string, nodo *yaml.Node) (map[string]V, error) {
	if nodo.Kind != yaml.MappingNode {
		return nil, defectoDeFilas(ruta, clave, "no es un mapa de filas")
	}

	filas := nuevasFilasLeidas[V](ruta, clave, len(nodo.Content)/2)

	for indice := 0; indice+1 < len(nodo.Content); indice += 2 {
		nombre, valor := nodo.Content[indice], nodo.Content[indice+1]

		if nombre.Kind != yaml.ScalarNode {
			filas.defecto("la clave de la línea %d no es un texto", nombre.Line)

			continue
		}

		if !filas.nueva(nombre.Value, nombre.Line) {
			continue
		}

		var fila V
		if err := valor.Decode(&fila); err != nil {
			filas.defecto("la fila de %s no se puede leer: %s", nombre.Value, err.Error())

			continue
		}

		filas.filas[nombre.Value] = fila
	}

	return filas.resultado()
}

// carga comprueba los ficheros ya decodificados mientras construye el
// registro, y acumula sus defectos en el orden de los ficheros y, dentro de
// cada uno, en el de sus códigos. Lo que tiene un defecto se sigue
// construyendo en lo posible, para que un defecto no arrastre a otros que no
// lo son; el registro solo se entrega sin ninguno.
type carga struct {
	defectos []error
}

// defecto anota un defecto de un fichero.
func (c *carga) defecto(fichero, formato string, argumentos ...any) {
	c.defectos = append(c.defectos, &defectoDeCarga{fichero: fichero, motivo: fmt.Sprintf(formato, argumentos...)})
}

// rechazo anota como defecto de un fichero el rechazo de un dato por la
// gramática de su identificador. Guarda su texto y no el error: su clase es de
// usuario y el defecto no lo es (defectoDeCarga).
func (c *carga) rechazo(fichero string, err error) {
	c.defectos = append(c.defectos, &defectoDeCarga{fichero: fichero, motivo: err.Error()})
}

// construir comprueba los ficheros y construye con ellos el registro.
func (c *carga) construir(ficheros Ficheros) *Registro {
	registro := &Registro{
		relacion:        c.origen(rutaDeMunicipios, ficheros.Municipios.Fecha, ficheros.Municipios.Source),
		correspondencia: c.origen(rutaDeDIR3, ficheros.DIR3.Fecha, ficheros.DIR3.Source),
		estado:          c.origen(rutaDelEstado, ficheros.Estado.Fecha, ficheros.Estado.Source),
	}

	estatal := ficheros.Estado.Boletin
	registro.estatal = c.boletin(rutaDelEstado, nivelEstatal, BoletinConfigurado{
		Codigo: estatal.Codigo, Nombre: estatal.Nombre, URL: estatal.URL,
	})

	provincias := c.provincias(ficheros.Comunidades)
	municipios := c.municipios(ficheros.Municipios.Municipios, provincias)
	c.correspondencia(ficheros.DIR3.Correspondencia, ficheros.Municipios.Municipios, municipios)

	registro.indexar(slices.Collect(maps.Values(municipios)))

	return registro
}

// origen comprueba la fecha y el source de un fichero.
func (c *carga) origen(fichero, fecha, source string) origen {
	analizada, err := time.Parse(time.DateOnly, fecha)
	if err != nil {
		c.defecto(fichero, "la fecha %q no es una fecha AAAA-MM-DD", fecha)
	}

	if source == "" {
		c.defecto(fichero, "no declara su source")
	}

	return origen{fecha: analizada, source: source}
}

// boletin comprueba que un boletín configurado lleva lo que el territorio
// resuelto emite de él y lo devuelve en esa forma, con la ruta del fichero
// que lo fija como source.
func (c *carga) boletin(fichero, nivel string, configurado BoletinConfigurado) Boletin {
	for _, campo := range []struct{ clave, valor string }{
		{"codigo", configurado.Codigo}, {"nombre", configurado.Nombre}, {"url", configurado.URL},
	} {
		if campo.valor == "" {
			c.defecto(fichero, "el boletín %s no tiene %s", nivel, campo.clave)
		}
	}

	return Boletin{
		Nivel:  nivel,
		Codigo: configurado.Codigo,
		Nombre: configurado.Nombre,
		URL:    configurado.URL,
		Motivo: configurado.Motivo,
		Source: fichero,
	}
}

// provincias comprueba los ficheros de comunidad y devuelve las provincias que
// declaran, por código, cada una con su comunidad (puntos 1, 4 y 5).
func (c *carga) provincias(comunidades map[string]FicheroDeComunidad) map[string]*provinciaRegistrada {
	provincias := map[string]*provinciaRegistrada{}

	for _, clave := range slices.Sorted(maps.Keys(comunidades)) {
		fichero := comunidades[clave]
		comunidad := c.comunidad(clave, fichero)

		for _, codigo := range slices.Sorted(maps.Keys(fichero.Provincias)) {
			if declarada, ya := provincias[codigo]; ya {
				c.defecto(comunidad.ruta, "declara la provincia %s, que ya declara %s", codigo, declarada.comunidad.ruta)

				continue
			}

			nombre := fichero.Provincias[codigo]
			if nombre == "" {
				c.defecto(comunidad.ruta, "la provincia %s no tiene nombre", codigo)
			}

			provincias[codigo] = &provinciaRegistrada{codigo: codigo, nombre: nombre, comunidad: comunidad}
		}
	}

	return provincias
}

// comunidad comprueba el fichero de una comunidad (puntos 4 y 5). La comunidad
// se registra con el código que da nombre a su fichero, aunque declare otro:
// ese defecto ya queda dicho, y así no arrastra a los de sus municipios.
func (c *carga) comunidad(clave string, fichero FicheroDeComunidad) *comunidadRegistrada {
	ruta := rutaDeComunidad(clave)

	if fichero.Codigo != clave {
		c.defecto(ruta, "el fichero es el de la comunidad %s y declara el código %q", clave, fichero.Codigo)
	}

	if fichero.Nombre == "" {
		c.defecto(ruta, "la comunidad no tiene nombre")
	}

	if fichero.Regimen != regimenComun && fichero.Regimen != regimenForal {
		c.defecto(ruta, "el régimen %q no es %s ni %s", fichero.Regimen, regimenComun, regimenForal)
	}

	comunidad := &comunidadRegistrada{
		codigo:  clave,
		nombre:  fichero.Nombre,
		regimen: fichero.Regimen,
		ruta:    ruta,
		origen:  c.origen(ruta, fichero.Fecha, fichero.Source),
	}

	if boletines := fichero.Boletines; boletines != nil {
		comunidad.autonomico = c.boletinConfigurado(ruta, nivelAutonomico, boletines.Autonomico)
		comunidad.provincial = c.boletinConfigurado(ruta, nivelProvincial, boletines.Provincial)
	}

	return comunidad
}

// boletinConfigurado comprueba un nivel de boletín de una comunidad, si está
// configurado; si no, es nil.
func (c *carga) boletinConfigurado(fichero, nivel string, configurado *BoletinConfigurado) *Boletin {
	if configurado == nil {
		return nil
	}

	boletin := c.boletin(fichero, nivel, *configurado)

	return &boletin
}

// municipios comprueba las filas de la relación y devuelve, por código, los
// municipios que no tienen ningún defecto (puntos 1 y 6).
func (c *carga) municipios(
	filas map[string]FilaDeMunicipio,
	provincias map[string]*provinciaRegistrada,
) map[string]*municipioRegistrado {
	municipios := make(map[string]*municipioRegistrado, len(filas))

	for _, clave := range slices.Sorted(maps.Keys(filas)) {
		if municipio := c.municipio(clave, filas[clave], provincias); municipio != nil {
			municipios[clave] = municipio
		}
	}

	return municipios
}

// municipio comprueba una fila de la relación. Con algún defecto devuelve nil.
func (c *carga) municipio(
	clave string,
	fila FilaDeMunicipio,
	provincias map[string]*provinciaRegistrada,
) *municipioRegistrado {
	codigo, err := ids.AnalizarCodigoINE(clave)
	if err != nil {
		c.rechazo(rutaDeMunicipios, err)

		return nil
	}

	antes := len(c.defectos)

	if len(fila.DC) != 1 || !esCifra(fila.DC[0]) {
		c.defecto(rutaDeMunicipios, "el municipio %s tiene el dígito de control %q, que no es una cifra", clave, fila.DC)
	}

	if fila.Nombre == "" {
		c.defecto(rutaDeMunicipios, "el municipio %s no tiene nombre", clave)
	}

	if fila.Provincia != codigo.Provincia() {
		c.defecto(rutaDeMunicipios, "el municipio %s declara la provincia %q y su código es de la provincia %s",
			clave, fila.Provincia, codigo.Provincia())
	}

	provincia, declarada := provincias[codigo.Provincia()]

	switch {
	case !declarada:
		c.defecto(rutaDeMunicipios, "el municipio %s es de la provincia %s, que no declara ninguna comunidad",
			clave, codigo.Provincia())
	case fila.Comunidad != provincia.comunidad.codigo:
		c.defecto(rutaDeMunicipios, "el municipio %s declara la comunidad %s (%s) y su provincia %s es de la "+
			"comunidad %s (%s)", clave, fila.Comunidad, rutaDeComunidad(fila.Comunidad), provincia.codigo,
			provincia.comunidad.codigo, provincia.comunidad.ruta)
	}

	if len(c.defectos) > antes {
		return nil
	}

	return &municipioRegistrado{codigo: codigo, digito: fila.DC, nombre: fila.Nombre, provincia: provincia}
}

// esCifra dice si el byte es una cifra ASCII, del 0 al 9.
func esCifra(b byte) bool {
	return '0' <= b && b <= '9'
}

// correspondencia comprueba la correspondencia INE→DIR3 contra la relación y
// asigna a cada municipio su DIR3 verificado (punto 2). Que un municipio esté
// en la relación se mira en sus filas, no entre los municipios sin defectos,
// para no decir que falta uno que solo tiene otro defecto ya dicho.
func (c *carga) correspondencia(
	filas map[string]string,
	relacion map[string]FilaDeMunicipio,
	municipios map[string]*municipioRegistrado,
) {
	for _, clave := range slices.Sorted(maps.Keys(filas)) {
		codigo, err := ids.AnalizarCodigoINE(clave)
		if err != nil {
			c.rechazo(rutaDeDIR3, err)

			continue
		}

		if _, esta := relacion[clave]; !esta {
			c.defecto(rutaDeDIR3, "el municipio %s no está en %s", clave, rutaDeMunicipios)

			continue
		}

		dir3, err := ids.AnalizarDIR3(filas[clave])
		if err != nil {
			c.rechazo(rutaDeDIR3, err)

			continue
		}

		if municipio := municipios[clave]; municipio != nil {
			c.ayuntamiento(municipio, codigo, dir3)
		}
	}
}

// ayuntamiento asigna a un municipio su DIR3 si es coherente con su código
// INE y su dígito de control; si no, lo anota como defecto.
func (c *carga) ayuntamiento(municipio *municipioRegistrado, codigo ids.CodigoINE, dir3 ids.DIR3) {
	if dir3.CodigoINE() != codigo || dir3.Digito() != municipio.digito[0] {
		c.defecto(rutaDeDIR3, "el DIR3 %s del municipio %s no es coherente con su código INE y su dígito de "+
			"control %s", dir3, codigo, municipio.digito)

		return
	}

	municipio.dir3 = &dir3
}
