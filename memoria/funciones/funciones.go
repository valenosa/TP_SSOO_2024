package funciones

import (
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/logueano"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

var ConfigJson config.Memoria

var Auxlogger *log.Logger

// Funciones auxiliares
// Toma de a un archivo a la vez y guarda las instrucciones en un map l
func GuardarInstrucciones(pid uint32, path string, memoriaInstrucciones map[uint32][]string) {
	path = ConfigJson.Instructions_Path + path
	data := ExtractInstructions(path)
	InsertData(pid, memoriaInstrucciones, data)
}

// Abre el archivo especificado por la ruta 'path' y guarda su contenido en un slice de bytes.
// Retorna el contenido del archivo como un slice de bytes.
func ExtractInstructions(path string) []byte {
	// Lee el archivo
	file, err := os.ReadFile(path)
	if err != nil {
		logueano.Error(Auxlogger, err)
		return nil
	}

	// Ahora 'file' es un slice de bytes con el contenido del archivo
	return file
}

// insertData inserta las instrucciones en la memoria de instrucciones asociadas al PID especificado
// e imprime las instrucciones guardadas en memoria junto con su PID correspondiente.
func InsertData(pid uint32, memoriaInstrucciones map[uint32][]string, data []byte) {
	// Separar las instrucciones por medio de tokens
	instrucciones := strings.Split(string(data), "\n")
	// Insertar las instrucciones en la memoria de instrucciones
	memoriaInstrucciones[pid] = instrucciones

	// Imprimir las instrucciones guardadas en memoria
	logueano.LeerInstrucciones(Auxlogger, memoriaInstrucciones, pid)
}

func AsignarTabla(pid uint32, tablaDePaginas map[uint32]structs.Tabla) {
	tablaDePaginas[pid] = structs.Tabla{}

	//^ log obligatorio (1/6)
	logueano.OperoConTablaDePaginas(pid, tablaDePaginas)
}

func BuscarMarco(pid uint32, pagina uint32, tablaDePaginas map[uint32]structs.Tabla) string {
	if len(tablaDePaginas[pid]) <= int(pagina) {
		return ""
	}

	marco := tablaDePaginas[pid][pagina]

	//^ log obligatorio (2/6)
	logueano.AccesoTabla(pid, pagina, marco)

	marcoStr := strconv.Itoa(marco)

	return marcoStr
}

func ObtenerPagina(pid uint32, direccionFisica uint32, tablaDePaginas map[uint32]structs.Tabla) int {

	marco := math.Floor(float64(direccionFisica) / float64(ConfigJson.Page_Size))

	for i := range tablaDePaginas[pid] {

		marcoActual := tablaDePaginas[pid][i]

		if uint32(marcoActual) == uint32(marco) {

			return i

		}
	}

	return -1
}

func tableHasNext(pid uint32, pagina uint32, tablaDePaginas map[uint32]structs.Tabla) bool {
	return len(tablaDePaginas[pid])-1 > int(pagina)
}

// Verifica si la pagina aun tiene espacio en memoria
func endOfPage(direccionFisica uint32) bool {
	//Si la direccion es multiplo del tamaño de pagina, es el fin de la pagina
	return direccionFisica%uint32(ConfigJson.Page_Size) == 0
}

func LiberarMarcos(marcosALiberar []int, bitMap []bool) {
	for _, marco := range marcosALiberar {
		bitMap[marco] = false
	}
}

func ReasignarPaginas(pid uint32, tablaDePaginas *map[uint32]structs.Tabla, bitMap []bool, size uint32) string {

	var accion string

	lenOriginal := len((*tablaDePaginas)[pid])

	cantidadDePaginas := int(math.Ceil(float64(size) / float64(ConfigJson.Page_Size)))

	//------------- CASO AGREGAR PAGINAS
	// Itera n cantidad de veces, siendo n la cantidad de paginas a agregar
	for len((*tablaDePaginas)[pid]) < cantidadDePaginas {

		// Por cada página a agregar, si no hay marcos disponibles, se devuelve un error OUT_OF_MEMORY
		outOfMemory := true

		// Recorre el bitMap buscando un marco desocupado
		for marco, ocupado := range bitMap {

			if !ocupado {
				// Guarda en la tabla de páginas del proceso el marco asignado a una página
				(*tablaDePaginas)[pid] = append((*tablaDePaginas)[pid], marco)
				// Marca el marco como ocupado
				bitMap[marco] = true

				// Notifica que por ahora no está OUT_OF_MEMORY
				outOfMemory = false
			}
		}

		//Si no hubo ningun marco desocupado para la página anterior, devuelve OUT_OF_MEMORY
		if outOfMemory {
			return "OUT_OF_MEMORY"
		}
	}

	accion = "Ampliar"

	//------------- CASO QUITAR PAGINAS
	if len((*tablaDePaginas)[pid]) > cantidadDePaginas {

		marcosALiberar := (*tablaDePaginas)[pid][cantidadDePaginas:]

		(*tablaDePaginas)[pid] = (*tablaDePaginas)[pid][:cantidadDePaginas]

		LiberarMarcos(marcosALiberar, bitMap)

		accion = "Reducir"
	}

	//^ log obligatorio ((3...4)/6)
	logueano.CambioDeTamaño(pid, lenOriginal, accion, tablaDePaginas)

	return "OK"
}

func LeerEnMemoria(pid uint32, tablaDePaginas map[uint32]structs.Tabla, pagina uint32, direccionFisica uint32, byteArraySize int, espacioUsuario *[]byte) ([]byte, string) {

	var dato []byte

	//^ log obligatorio (5/5)
	logueano.AccesoEspacioUsuario(pid, "LEER", direccionFisica, byteArraySize)

	// Itera sobre los bytes del dato recibido.
	for ; byteArraySize > 0; byteArraySize-- {

		// Lee el byte en la dirección física.
		dato = append(dato, (*espacioUsuario)[direccionFisica])

		// Incrementa la dirección
		direccionFisica++

		// Si la siguiente direccion fisica es endOfPage (ya no pertenece al marco en el que estamos escribiendo), hace cambio de página
		if endOfPage(direccionFisica) {
			// Si no se puede hacer el cambio de página, es OUT_OF_MEMORY
			if !cambioDePagina(&direccionFisica, pid, tablaDePaginas, &pagina) {
				return dato, "OUT_OF_MEMORY"
			}
		}
	}
	return dato, "OK"
}

// Escribe en memoria el dato recibido en la dirección física especificada.
func EscribirEnMemoria(pid uint32, tablaDePaginas map[uint32]structs.Tabla, pagina uint32, direccionFisica uint32, datoBytes []byte, espacioUsuario *[]byte) string {

	//^ log obligatorio (5/5)
	logueano.AccesoEspacioUsuario(pid, "ESCRIBIR", direccionFisica, len(datoBytes))

	// Itera sobre los bytes del dato recibido.
	for i := range datoBytes {

		// Escribe el byte en la dirección física.
		(*espacioUsuario)[direccionFisica] = datoBytes[i]

		// Incrementa la dirección
		direccionFisica++

		// Si la siguiente direccion fisica es endOfPage (ya no pertenece al marco en el que estamos escribiendo), hace cambio de página
		if endOfPage(direccionFisica) {
			// Si no se puede hacer el cambio de página, es OUT_OF_MEMORY
			if !cambioDePagina(&direccionFisica, pid, tablaDePaginas, &pagina) {
				return "OUT_OF_MEMORY"
			}
		}
	}

	return "OK"
}

func cambioDePagina(direccionFisica *uint32, pid uint32, tablasDePaginas map[uint32]structs.Tabla, pagina *uint32) bool {

	if tableHasNext(pid, *pagina, tablasDePaginas) {
		// Cambio la direccion fisica a la primera del siguitabla
		*pagina++
		*direccionFisica = uint32(((tablasDePaginas)[pid][*pagina]) * int(ConfigJson.Page_Size))
		return true
	}
	return false
}
