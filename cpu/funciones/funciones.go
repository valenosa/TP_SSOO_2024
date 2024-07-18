package funciones

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/logueano"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

//----------------------( VARIABLES )----------------------\\

// Contiene el pid del proceso que dispatch mandó a ejecutar (se usa para que el handler de la interrupción pueda chequear que el pid del proceso que mandó la interrupción sea el mismo que el pid del proceso que está en ejecución)
var PidEnEjecucion uint32

var HayInterrupcion bool = false

var RegistrosCPU structs.RegistrosUsoGeneral

var ConfigJson config.Cpu

// Es global porque la uso para "depositar" el motivo de desalojo del proceso (que a excepción de EXIT, es traído por una interrupción)
var MotivoDeDesalojo string

var Auxlogger *log.Logger

// ----------------------( TLB )----------------------\\
type pid = uint32
type pagina = uint32
type marco = uint32

// Estructura de la TLB.
// | El pid es el key, y el valor es otro mapa que tiene como key la página y como valor el marco.
type TLB map[pid]map[pagina]marco

type ElementoPrioridad struct {
	Pid    uint32
	Pagina uint32
}

// ----------------------( FUNCIONES TLB )----------------------\\

func (tlb TLB) initPID(pid uint32) {
	if tlb[pid] == nil {
		tlb[pid] = make(map[pagina]marco)
	}
}

func (tlb TLB) longitudTLB() int {

	sumatoria := 0

	for _, entradas := range tlb {
		sumatoria = sumatoria + len(entradas)
	}

	return sumatoria
}

// Valida si el TLBA está lleno
func (tlb TLB) Full() bool {
	return tlb.longitudTLB() >= ConfigJson.Number_Felling_tlb
}

// Hit or miss? I guess they never miss, huh?
func (tlb TLB) Hit(pid uint32, pagina uint32) (uint32, bool) {
	marco, encontrado := tlb[pid][pagina]
	return marco, encontrado
}

// ----------------------( MMU )----------------------\\

// Recibe una direccion lógica, devuelve una física y maneja el TLB
func TraduccionMMU(pid uint32, direccionLogica int, tlb *TLB, prioridadesTLB *[]ElementoPrioridad) (uint32, bool) {

	// Obtiene la página y el desplazamiento de la dirección lógica
	numeroDePagina, desplazamiento := ObtenerPaginayDesplazamiento(direccionLogica)

	// Obtiene el marco de la página
	marco, encontrado := ObtenerMarco(PidEnEjecucion, uint32(numeroDePagina), tlb, prioridadesTLB)

	// Si no se encontró el marco, se devuelve un error
	if !encontrado {
		fmt.Println("ERROR: Page Fault")
		return 0, false
	}

	// Calcula la dirección física

	pageSize := uint32(ConfigJson.Page_Size)

	desp := uint32(desplazamiento)

	direccionFisica := marco*pageSize + desp

	return direccionFisica, true
}

func ObtenerPaginayDesplazamiento(direccionLogica int) (int, int) {

	numeroDePagina := int(math.Floor(float64(direccionLogica) / float64(ConfigJson.Page_Size)))
	desplazamiento := direccionLogica - numeroDePagina*int(ConfigJson.Page_Size)

	return numeroDePagina, desplazamiento
}

// obtiene el marco de la pagina
func ObtenerMarco(pid uint32, pagina uint32, tlb *TLB, prioridadesTLB *[]ElementoPrioridad) (uint32, bool) {

	// Busca en la TLB
	marco, encontrado := (*tlb).Hit(pid, pagina)

	//^log obligatorio (3...4/6)
	logueano.TLBAccion(pid, encontrado, pagina)

	// Si no está en la TLB, busca en la tabla de páginas y de paso lo agrega
	if !encontrado {
		marco, encontrado = buscarEnMemoria(pid, pagina)

		//^log obligatorio (5/6)
		logueano.ObtenerMarcolg(pid, encontrado, pagina, marco)

		agregarEnTLB(pagina, marco, pid, tlb, prioridadesTLB)

	}
	// No se toma en cuenta el caso en el que no existe el marco
	return marco, encontrado
}

// TODO: Probar TLB, especificamente los algoritmos de remplazo
func agregarEnTLB(pagina uint32, marco uint32, pid uint32, tlb *TLB, prioridadesTLB *[]ElementoPrioridad) {
	if tlb.Full() {
		planificarTLB(pid, pagina, marco, tlb, prioridadesTLB)

	} else {
		(*tlb).initPID(pid)

		// agregar marco al TLB
		(*tlb)[pid][pagina] = marco
		// agregar a la lista de prioridades
		(*prioridadesTLB) = append((*prioridadesTLB), ElementoPrioridad{Pid: pid, Pagina: pagina})
	}
}

func planificarTLB(pid uint32, pagina uint32, marco uint32, tlb *TLB, prioridadesTLB *[]ElementoPrioridad) {
	switch ConfigJson.Algorithm_tlb {

	case "FIFO":
		algoritmoFifo(pid, pagina, marco, tlb, prioridadesTLB)

	case "LRU":
		algoritmoLru(pid, pagina, marco, tlb, prioridadesTLB)
	}
}

func algoritmoFifo(pid uint32, pagina uint32, marco uint32, tlb *TLB, prioridadesTLB *[]ElementoPrioridad) {
	_, paginaEncontrada := (*tlb)[pid][pagina]

	if !paginaEncontrada {
		// Elimina el primer elemento de la lista de prioridades
		delete((*tlb)[(*prioridadesTLB)[0].Pid], (*prioridadesTLB)[0].Pagina)
		(*prioridadesTLB) = (*prioridadesTLB)[1:]

		// Agrega el marco a la TLB
		(*tlb)[pid][pagina] = marco
		(*prioridadesTLB) = append((*prioridadesTLB), ElementoPrioridad{Pid: pid, Pagina: pagina})

	} else {

		(*tlb)[pid][pagina] = marco
	}
}

func algoritmoLru(pid uint32, pagina uint32, marco uint32, tlb *TLB, prioridadesTLB *[]ElementoPrioridad) {
	encontrado := false
	for posicion, entrada := range *prioridadesTLB {
		//si encuentro un elemento con el mismo pid y pagina
		if entrada.Pid == pid && entrada.Pagina == pagina {

			//Se elimina el elemento en la lista de prioridades
			(*prioridadesTLB) = append((*prioridadesTLB)[:posicion], (*prioridadesTLB)[posicion+1:]...)

			//Lo paso al final
			(*prioridadesTLB) = append((*prioridadesTLB), entrada)

			//Cambio el marco de la página en el TLB
			(*tlb)[pid][pagina] = marco

			encontrado = true
			break
		}
	}
	if !encontrado {
		algoritmoFifo(pid, pagina, marco, tlb, prioridadesTLB)
	}
}

func buscarEnMemoria(pid uint32, pagina uint32) (uint32, bool) {

	// Crea un cliente HTTP
	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/memoria/marco", ConfigJson.Ip_Memory, ConfigJson.Port_Memory)

	// Crea una nueva solicitud PUT
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logueano.Error(Auxlogger, err)
		return 0, false
	}

	// Agrega el PID y la PAGINA como params
	q := req.URL.Query()
	q.Add("pid", fmt.Sprint(pid))
	q.Add("pagina", fmt.Sprint(pagina))
	req.URL.RawQuery = q.Encode()

	// Establece el tipo de contenido de la solicitud
	req.Header.Set("Content-Type", "text/plain")

	// Realiza la solicitud al servidor de memoria
	respuesta, err := cliente.Do(req)
	if err != nil {
		logueano.Error(Auxlogger, err)
		return 0, false
	}

	// Verifica el código de estado de la respuesta
	if respuesta.StatusCode != http.StatusOK {
		return 0, false
	}

	// Lee el cuerpo de la respuesta
	marcoBytes, err := io.ReadAll(respuesta.Body)
	if err != nil {
		return 0, false
	}

	// Convierte el valor de la instrucción a un uint64 bits.
	valorInt64, err := strconv.ParseUint(string(marcoBytes), 10, 32)
	if err != nil {
		logueano.Error(Auxlogger, err)
		return 0, false
	}

	// Disminuye el valor de la instrucción en uno para ajustarlo al índice del slice de instrucciones.
	marcoEncontrado := uint32(valorInt64)

	return uint32(marcoEncontrado), true
}

//----------------------( FUNCIONES CICLO DE INSTRUCCION )----------------------\\

// Ejecuta un ciclo de instruccion.
func EjecutarCiclosDeInstruccion(PCB *structs.PCB, TLB *TLB, prioridadesTLB *[]ElementoPrioridad) {
	var cicloFinalizado bool = false

	// Itera el ciclo de instrucción si hay instrucciones a ejecutar y no hay interrupciones.
	for !HayInterrupcion && !cicloFinalizado {
		// Obtiene la próxima instrucción a ejecutar.
		instruccion := Fetch(PCB.PID, RegistrosCPU.PC)

		//^log obligatorio (1/6)
		logueano.FetchInstruccion(*PCB)

		// Decodifica y ejecuta la instrucción.
		DecodeAndExecute(PCB, instruccion, &RegistrosCPU.PC, &cicloFinalizado, TLB, prioridadesTLB)
	}
	HayInterrupcion = false // Resetea la interrupción

	// Actualiza los registros de uso general del PCB con los registros de la CPU.
	PCB.RegistrosUsoGeneral = RegistrosCPU
}

// Trae de memoria las instrucciones indicadas por el PC y el PID.
func Fetch(PID uint32, PC uint32) string {

	// Convierte el PID y el PC a string
	pid := strconv.FormatUint(uint64(PID), 10)
	pc := strconv.FormatUint(uint64(PC), 10)

	// Crea un cliente HTTP
	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/instrucciones", ConfigJson.Ip_Memory, ConfigJson.Port_Memory)

	// Crea una nueva solicitud GET
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ""
	}

	// Agrega el PID y el PC como params
	q := req.URL.Query()
	q.Add("PID", pid)
	q.Add("PC", pc)
	req.URL.RawQuery = q.Encode()

	// Establece el tipo de contenido de la solicitud
	req.Header.Set("Content-Type", "text/plain")

	// Realiza la solicitud al servidor de memoria
	respuesta, err := cliente.Do(req)
	if err != nil {
		return ""
	}

	// Verifica el código de estado de la respuesta
	if respuesta.StatusCode != http.StatusOK {
		return ""
	}

	// Lee el cuerpo de la respuesta
	bodyBytes, err := io.ReadAll(respuesta.Body)
	if err != nil {
		return ""
	}

	// Retorna las instrucciones obtenidas como una cadena de texto
	return string(bodyBytes)
}

// Ejecuta las instrucciones traidas de memoria.
// TODO: Todos los cambios de estado se deben relizar en kernel en base a el Motivo de Desalojo ( Kernel: administrarMotivoDesalojo() )
func DecodeAndExecute(PCB *structs.PCB, instruccion string, PC *uint32, cicloFinalizado *bool, TLB *TLB, prioridadesTLB *[]ElementoPrioridad) {

	// Mapa de registros para acceder a los registros de la CPU por nombre
	var registrosMap8 = map[string]*uint8{
		"AX": &RegistrosCPU.AX,
		"BX": &RegistrosCPU.BX,
		"CX": &RegistrosCPU.CX,
		"DX": &RegistrosCPU.DX,
	}

	var registrosMap32 = map[string]*uint32{
		"EAX": &RegistrosCPU.EAX,
		"EBX": &RegistrosCPU.EBX,
		"ECX": &RegistrosCPU.ECX,
		"EDX": &RegistrosCPU.EDX,
		"SI":  &RegistrosCPU.SI,
		"DI":  &RegistrosCPU.DI,
	}

	// Parsea las instrucciones de la cadena de instrucción
	variable := strings.Split(instruccion, " ")

	//^log obligatorio (2/6)
	logueano.EjecucionInstruccion(*PCB, variable)

	// Switch para determinar la operación a realizar según la instrucción
	switch variable[0] {
	case "SET":
		set(variable[1], variable[2], registrosMap8, registrosMap32, PC)

	case "SUM":
		sum(variable[1], variable[2], registrosMap8, registrosMap32)

	case "SUB":
		sub(variable[1], variable[2], registrosMap8, registrosMap32)

	case "JNZ":
		jnz(variable[1], variable[2], PC, registrosMap8, registrosMap32)

	case "MOV_IN":
		estado, valor, dirF := movIN(variable[1], variable[2], registrosMap8, registrosMap32, TLB, prioridadesTLB)

		if estado != "OK" {
			*cicloFinalizado = true
			PCB.Estado = "EXIT"       //TODO: Manejar en kernel
			MotivoDeDesalojo = estado //TODO: Manejar en kernel
			return
		} else {
			//^log obligatorio (6/6)
			logueano.LecturaEscritura(*PCB, "LEER", dirF, valor)
		}
	case "MOV_OUT":
		estado, valor, dirF := movOUT(variable[1], variable[2], registrosMap8, registrosMap32, TLB, prioridadesTLB)

		if estado != "OK" {
			*cicloFinalizado = true
			PCB.Estado = "EXIT"       //TODO: Manejar en kernel
			MotivoDeDesalojo = estado //TODO: Manejar en kernel
			return
		} else {
			//^log obligatorio (6/6)
			logueano.LecturaEscritura(*PCB, "ESCRIBIR", dirF, valor)
		}

	case "COPY_STRING":
		estado, valor, dirFR, dirFW := copyString(variable[1], TLB, prioridadesTLB)

		if estado != "OK" {
			*cicloFinalizado = true
			PCB.Estado = "EXIT"       //TODO: Manejar en kernel
			MotivoDeDesalojo = estado //TODO: Manejar en kernel
			return
		} else {
			//^log obligatorio (6/6)
			logueano.LecturaEscritura(*PCB, "LEER", dirFR, valor)
			logueano.LecturaEscritura(*PCB, "ESCRIBIR", dirFW, valor)
		}

	case "RESIZE":
		estado := resize(variable[1])
		if estado == "OUT OF MEMORY" {
			*cicloFinalizado = true
			MotivoDeDesalojo = estado
			return
		}

	case "WAIT":
		wait(variable[1], PCB, cicloFinalizado)

	case "SIGNAL":
		signal(variable[1], PCB, cicloFinalizado)

	case "IO_GEN_SLEEP":
		*cicloFinalizado = true
		MotivoDeDesalojo = "IO"
		go ioGenSleep(variable[1], variable[2], PCB.PID)

	case "IO_STDIN_READ":
		*cicloFinalizado = true
		MotivoDeDesalojo = "IO"
		go ioSTD(variable[1], variable[2], variable[3], registrosMap8, registrosMap32, PCB.PID, TLB, prioridadesTLB, "IO_STDIN_READ")

	case "IO_STDOUT_WRITE":
		*cicloFinalizado = true
		MotivoDeDesalojo = "IO"
		go ioSTD(variable[1], variable[2], variable[3], registrosMap8, registrosMap32, PCB.PID, TLB, prioridadesTLB, "IO_STDOUT_WRITE")

	case "IO_FS_CREATE":
		*cicloFinalizado = true
		MotivoDeDesalojo = "IO"
		go ioFSCreateOrDelete(variable[1], variable[2], PCB.PID, "IO_FS_CREATE")

		//!manejar el case de IO_FS_READ

	case "IO_FS_DELETE":
		*cicloFinalizado = true
		MotivoDeDesalojo = "IO"
		go ioFSCreateOrDelete(variable[1], variable[2], PCB.PID, "IO_FS_DELETE")

	case "IO_FS_TRUNCATE":
		*cicloFinalizado = true
		MotivoDeDesalojo = "IO"
		go ioFSTruncate(variable[1], variable[2], variable[3], PCB.PID, registrosMap8, registrosMap32)

	case "IO_FS_WRITE":
		*cicloFinalizado = true
		MotivoDeDesalojo = "IO"
		go ioFSRW(variable[1], variable[2], variable[3], variable[4], variable[5], PCB.PID, registrosMap8, registrosMap32, TLB, prioridadesTLB, "IO_FS_WRITE")

	case "IO_FS_READ":
		*cicloFinalizado = true
		MotivoDeDesalojo = "IO"
		go ioFSRW(variable[1], variable[2], variable[3], variable[4], variable[5], PCB.PID, registrosMap8, registrosMap32, TLB, prioridadesTLB, "IO_FS_READ")

	case "EXIT":
		*cicloFinalizado = true
		MotivoDeDesalojo = "EXIT"
		return

	default:
		fmt.Println("------")
	}

	// Incrementa el Program Counter para apuntar a la siguiente instrucción
	*PC++
}

//----------------------( FUNCIONES DE INSTRUCCIONES )----------------------\\

// Asigna al registro el valor pasado como parámetro.
func set(reg string, dato string, registroMap8 map[string]*uint8, registroMap32 map[string]*uint32, PC *uint32) {

	// Verifica si el registro a asignar es el PC
	if reg == "PC" {

		// Convierte el valor a un entero sin signo de 32 bits
		valorInt64, err := strconv.ParseUint(dato, 10, 32)
		if err != nil {
			logueano.Error(Auxlogger, err)
		}

		// Asigna el valor al PC (resta 1 ya que el PC se incrementará después de esta instrucción)
		*PC = uint32(valorInt64) - 1
		return
	}

	if reg == "AX" || reg == "BX" || reg == "CX" || reg == "DX" {

		// Obtiene el puntero al registro del mapa de registros
		registro, encontrado := registroMap8[reg]
		if !encontrado {
			logueano.Mensaje(Auxlogger, "Registro no encontrado")
			return
		}

		// Convierte el valor de string a entero
		valor, err := strconv.Atoi(dato)

		if err != nil {
			logueano.Error(Auxlogger, err)
		}

		// Asigna el nuevo valor al registro
		*registro = uint8(valor)
	} else {

		// Obtiene el puntero al registro del mapa de registros
		registro, encontrado := registroMap32[reg]
		if !encontrado {
			logueano.Mensaje(Auxlogger, "Registro no encontrado")
			return
		}

		// Convierte el valor de string a entero
		valor, err := strconv.Atoi(dato)

		if err != nil {
			logueano.Mensaje(Auxlogger, "Dato no valido")
		}

		// Asigna el nuevo valor al registro
		*registro = uint32(valor)
	}
}

// Suma al Registro Destino el Registro Origen y deja el resultado en el Registro Destino.
func sum(reg1 string, reg2 string, registroMap map[string]*uint8, registroMap32 map[string]*uint32) {

	// Verifica si existen los registros especificados en la instrucción.

	registro1 := extraerDatosDelRegistro(reg1, registroMap, registroMap32)
	registro2 := extraerDatosDelRegistro(reg2, registroMap, registroMap32)

	// Suma el valor del Registro Origen al Registro Destino.
	registro1 += registro2

	// Escribe el resultado en el Registro Destino.
	if reg1 == "AX" || reg1 == "BX" || reg1 == "CX" || reg1 == "DX" {
		*registroMap[reg1] = uint8(registro1)
	} else {
		*registroMap32[reg1] = registro1
	}
}

// Resta al Registro Destino el Registro Origen y deja el resultado en el Registro Destino.
func sub(reg1 string, reg2 string, registroMap map[string]*uint8, registroMap32 map[string]*uint32) {

	registro1 := extraerDatosDelRegistro(reg1, registroMap, registroMap32)
	registro2 := extraerDatosDelRegistro(reg2, registroMap, registroMap32)

	// Resta el valor del Registro Origen al Registro Destino.
	registro1 -= registro2

	// Escribe el resultado en el Registro Destino.
	if reg1 == "AX" || reg1 == "BX" || reg1 == "CX" || reg1 == "DX" {
		*registroMap[reg1] = uint8(registro1)
	} else {
		*registroMap32[reg1] = registro1
	}
}

// Si el valor del registro es distinto de cero, actualiza el PC al numero de instruccion pasada por parametro.
func jnz(reg string, valor string, PC *uint32, registroMap map[string]*uint8, registroMap32 map[string]*uint32) {

	// Verifica si existe el registro especificado en la instrucción.
	registro := extraerDatosDelRegistro(reg, registroMap, registroMap32)

	// Convierte el valor de la instrucción a un uint64 bits.
	valorInt64, err := strconv.ParseUint(valor, 10, 32)
	if err != nil {
		logueano.Error(Auxlogger, err)
		return
	}

	// Le resto uno al valor de la instrucción porque después se incrementa el PC al salir del switch.
	nuevoValor := uint32(valorInt64) - 1

	// Si el valor del registro es distinto de cero, actualiza el PC al nuevo valor.
	if registro != 0 {
		*PC = nuevoValor
	}
}

func resize(tamañoEnBytes string) string {
	// Convierte el PID y el PC a string
	pid := strconv.FormatUint(uint64(PidEnEjecucion), 10)

	// Crea un cliente HTTP
	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/memoria/resize", ConfigJson.Ip_Memory, ConfigJson.Port_Memory)

	// Crea una nueva solicitud GET
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return ""
	}

	// Agrega el PID y el PC como params
	q := req.URL.Query()
	q.Add("pid", pid)
	q.Add("size", tamañoEnBytes)
	req.URL.RawQuery = q.Encode()

	// Realiza la solicitud al servidor de memoria
	respuesta, err := cliente.Do(req)
	if err != nil {
		return ""
	}

	// Verifica el código de estado de la respuesta
	if respuesta.StatusCode != http.StatusOK {
		return ""
	}

	// Lee el cuerpo de la respuesta
	bodyBytes, err := io.ReadAll(respuesta.Body)
	if err != nil {
		return ""
	}

	// Retorna las instrucciones obtenidas como una cadena de texto
	return string(bodyBytes)
}

func movIN(registroDato string, registroDireccion string, registrosMap8 map[string]*uint8, registrosMap32 map[string]*uint32, TLB *TLB, prioridadesTLB *[]ElementoPrioridad) (string, string, string) {

	var direccionFisica uint32
	var encontrado bool
	var longitud string

	// Dir Logica a Fisica

	direccionFisica, encontrado = obtenerDireccionFisica(registroDireccion, registrosMap8, registrosMap32, TLB, prioridadesTLB)

	if !encontrado {
		logueano.Mensaje(Auxlogger, "Page Fault")
		return "PAGE FAULT", "", "" //?Es correcto esto?

	}

	// Obtiene longitud del registro de dato
	if registroDato == "AX" || registroDato == "BX" || registroDato == "CX" || registroDato == "DX" {
		longitud = "1"
	} else {
		longitud = "4"
	}

	// Crea un cliente HTTP
	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/memoria/movin", ConfigJson.Ip_Memory, ConfigJson.Port_Memory)

	// Crea una nueva solicitud GET
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return "", "", ""
	}

	//Parsea la direccion física de uint32 a string.
	direccionFisicaStr := strconv.FormatUint(uint64(direccionFisica), 10)
	pidEnEjecucionStr := strconv.FormatUint(uint64(PidEnEjecucion), 10)

	// Agrega el PID y el PC como params
	q := req.URL.Query()
	q.Add("pid", pidEnEjecucionStr)
	q.Add("dir", direccionFisicaStr)
	q.Add("size", longitud)
	req.URL.RawQuery = q.Encode()

	// Establece el tipo de contenido de la solicitud
	req.Header.Set("Content-Type", "text/plain")

	// Realiza la solicitud al servidor de memoria
	respuesta, err := cliente.Do(req)

	if err != nil {
		return "", "", ""
	}

	if respuesta.StatusCode == http.StatusNotFound {
		return "OUT OF MEMORY", "", ""
	}

	if respuesta.StatusCode != http.StatusOK {
		return "", "", ""
	}

	// Lee el cuerpo de la respuesta
	data, err := io.ReadAll(respuesta.Body)
	if err != nil {
		return "", "", ""
	}

	var dataStr string = string(data)

	escribirEnRegistro(registroDato, data, registrosMap8, registrosMap32)

	return "OK", dataStr, direccionFisicaStr
}

func escribirEnRegistro(registroDato string, data []byte, registrosMap8 map[string]*uint8, registrosMap32 map[string]*uint32) {
	if len(data) == 1 {
		*registrosMap8[registroDato] = uint8(data[0])
	} else {
		*registrosMap32[registroDato] = binary.BigEndian.Uint32(data)
	}
}

func extraerBytesDelRegistro(registroDato string, registrosMap8 map[string]*uint8, registrosMap32 map[string]*uint32) []byte {
	if registroDato == "AX" || registroDato == "BX" || registroDato == "CX" || registroDato == "DX" {
		return []byte{*registrosMap8[registroDato]}
	} else {
		data := make([]byte, 4)
		binary.BigEndian.PutUint32(data, *registrosMap32[registroDato])
		return data
	}
}

func extraerDatosDelRegistro(registroDato string, registrosMap8 map[string]*uint8, registrosMap32 map[string]*uint32) uint32 {
	if registroDato == "AX" || registroDato == "BX" || registroDato == "CX" || registroDato == "DX" {
		return uint32(*registrosMap8[registroDato])
	} else {
		return *registrosMap32[registroDato]
	}
}

func obtenerDireccionFisica(registroDireccion string, registrosMap8 map[string]*uint8, registrosMap32 map[string]*uint32, TLB *TLB, prioridadesTLB *[]ElementoPrioridad) (uint32, bool) {
	if registroDireccion == "AX" || registroDireccion == "BX" || registroDireccion == "CX" || registroDireccion == "DX" {
		return TraduccionMMU(PidEnEjecucion, int(*(registrosMap8[registroDireccion])), TLB, prioridadesTLB)
	}
	return TraduccionMMU(PidEnEjecucion, int(*(registrosMap32[registroDireccion])), TLB, prioridadesTLB)
}

func movOUT(registroDireccion string, registroDato string, registrosMap8 map[string]*uint8, registrosMap32 map[string]*uint32, TLB *TLB, prioridadesTLB *[]ElementoPrioridad) (string, string, string) {

	direccionFisica, encontrado := obtenerDireccionFisica(registroDireccion, registrosMap8, registrosMap32, TLB, prioridadesTLB)

	direccionFisicaStr := strconv.FormatUint(uint64(direccionFisica), 10)

	if !encontrado {
		logueano.Mensaje(Auxlogger, "Page Fault")
		return "PAGE FAULT", "", ""
	}

	valor := extraerBytesDelRegistro(registroDato, registrosMap8, registrosMap32)

	var valorStr string = string(valor)

	body, err := json.Marshal(structs.RequestMovOUT{Pid: PidEnEjecucion, Dir: direccionFisica, Data: valor})

	if err != nil {
		return "", "", ""
	}

	// Envía la solicitud de ejecucion a Kernel
	respuesta, err := config.Request(ConfigJson.Port_Memory, ConfigJson.Ip_Memory, "POST", "memoria/movout", body)
	if err != nil {
		logueano.Error(Auxlogger, err)
		return "", "", ""
	}

	if respuesta.StatusCode == http.StatusNotFound {
		return "INVALID_WRITE", "", ""
	}

	if respuesta.StatusCode != http.StatusOK {
		logueano.MensajeConFormato(Auxlogger, "Error : %d", respuesta.StatusCode)
		return "", "", ""
	}

	return "OK", valorStr, direccionFisicaStr
}

func wait(nombreRecurso string, PCB *structs.PCB, cicloFinalizado *bool) {

	//--------- REQUEST ---------

	//Creo estructura de request
	var requestRecurso = structs.RequestRecurso{
		PidSolicitante: PCB.PID,
		NombreRecurso:  nombreRecurso,
	}

	//Convierto request a JSON
	body, err := json.Marshal(requestRecurso)
	if err != nil {
		return
	}

	// Envía la solicitud de ejecución a Kernel
	respuesta, err := config.Request(ConfigJson.Port_Kernel, ConfigJson.Ip_Kernel, "POST", "wait", body)
	if err != nil {
		logueano.Error(Auxlogger, err)
		return
	}

	// Decodifica en formato JSON la request.
	var respWait string
	err = json.NewDecoder(respuesta.Body).Decode(&respWait)
	if err != nil {
		return
	}

	//--------- EJECUTA ---------

	switch respWait {
	case "OK: Recurso asignado":
		// Agrega el recurso a la lista de recursos retenidos por el proceso.
		PCB.Recursos = append(PCB.Recursos, nombreRecurso) // En base a esta lista se liberaran los recursos al finalizar el proceso
		return

	case "BLOQUEAR: Recurso no disponible":
		*cicloFinalizado = true
		MotivoDeDesalojo = "WAIT"
		//Bloquea el proceso
		return

	case "ERROR: Recurso no existe":
		*cicloFinalizado = true
		MotivoDeDesalojo = "ERROR: Recurso no existe"
		return
	}
}

func signal(nombreRecurso string, PCB *structs.PCB, cicloFinalizado *bool) {

	//Convierto request a JSON
	body, err := json.Marshal(nombreRecurso)
	if err != nil {
		return
	}

	// Envía la solicitud de ejecucion a Kernel
	respuesta, err := config.Request(ConfigJson.Port_Kernel, ConfigJson.Ip_Kernel, "POST", "signal", body)
	if err != nil {
		logueano.Error(Auxlogger, err)
		return
	}

	if respuesta.StatusCode != http.StatusOK {

		*cicloFinalizado = true
		MotivoDeDesalojo = "ERROR: Recurso no existe"
		return
	}

	// Elimina el recurso liberado de la lista de recursos retenidos por el proceso
	for i, recurso := range PCB.Recursos {
		if recurso == nombreRecurso {
			PCB.Recursos = append(PCB.Recursos[:i], PCB.Recursos[i+1:]...)
			return
		}
	}
}

// Envía una request a Kernel con el nombre de una interfaz y las unidades de trabajo a multiplicar.
func ioGenSleep(nombreInterfaz string, unitWorkTimeString string, PID uint32) {

	// Convierte el tiempo de trabajo de la unidad de cadena a entero.
	unitWorkTime, err := strconv.Atoi(unitWorkTimeString)
	if err != nil {
		return
	}

	//Creo estructura de request
	var requestEjecutarInstuccion = structs.RequestEjecutarInstruccionIO{
		PidDesalojado:  PID,
		NombreInterfaz: nombreInterfaz,
		Instruccion:    "IO_GEN_SLEEP",
		UnitWorkTime:   unitWorkTime,
	}

	//Convierto request a JSON
	body, err := json.Marshal(requestEjecutarInstuccion)
	if err != nil {
		return
	}

	// Envía la solicitud de ejecucion a Kernel
	config.Request(ConfigJson.Port_Kernel, ConfigJson.Ip_Kernel, "POST", "instruccionIO", body)
}

func ioSTD(nombreInterfaz string, regDir string, regTamaño string, registroMap8 map[string]*uint8, registroMap32 map[string]*uint32, PID uint32, tlb *TLB,
	prioridadesTLB *[]ElementoPrioridad, instruccionIO string) {

	//Extrae el tamaño de la instrucción
	tamaño := extraerDatosDelRegistro(regTamaño, registroMap8, registroMap32)

	//Traduce dirección lógica a física
	direccion, encontrado := obtenerDireccionFisica(regDir, registroMap8, registroMap32, tlb, prioridadesTLB)
	if !encontrado {
		logueano.Mensaje(Auxlogger, "No se pudo traducir el registro de dirección lógica a física.")
		return
	}

	//Crea una variable que contiene el cuerpo de la request.
	var requestEjecutarInstuccion = structs.RequestEjecutarInstruccionIO{
		PidDesalojado:  PID,
		NombreInterfaz: nombreInterfaz,
		Instruccion:    instruccionIO,
		Direccion:      direccion,
		Tamaño:         tamaño,
	}

	// Convierte request a JSON
	body, err := json.Marshal(requestEjecutarInstuccion)
	if err != nil {
		return
	}

	// Envía la solicitud de ejecucion a Kernel
	respuesta, err := config.Request(ConfigJson.Port_Kernel, ConfigJson.Ip_Kernel, "POST", "instruccionIO", body)
	if err != nil {
		logueano.Error(Auxlogger, err)
		return
	}

	respuestaBody, err := io.ReadAll(respuesta.Body)
	if err != nil {
		logueano.Error(Auxlogger, err)
		return
	}
	respuestaString := string(respuestaBody)

	fmt.Println(respuestaString)
}

func copyString(tamaño string, TLB *TLB, prioridadesTLB *[]ElementoPrioridad) (string, string, string, string) {

	direccionEscritura, encontrado := TraduccionMMU(PidEnEjecucion, int(RegistrosCPU.DI), TLB, prioridadesTLB)

	if !encontrado {
		logueano.Mensaje(Auxlogger, "Error: Page Fault")
		return "PAGE FAULT", "", "", ""
	}

	direccionLectura, encontrado := TraduccionMMU(PidEnEjecucion, int(RegistrosCPU.SI), TLB, prioridadesTLB)

	if !encontrado {
		logueano.Mensaje(Auxlogger, "Error: Page Fault")
		return "PAGE FAULT", "", "", ""
	}

	// Crea un cliente HTTP
	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/memoria/copystr", ConfigJson.Ip_Memory, ConfigJson.Port_Memory)

	// Crea una nueva solicitud GET
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", "", ""
	}

	//Parsea la direccion física de uint32 a string.
	pidEnEjecucionStr := strconv.FormatUint(uint64(PidEnEjecucion), 10)
	direccionEscrituraStr := strconv.FormatUint(uint64(direccionEscritura), 10)
	direccionLecturaStr := strconv.FormatUint(uint64(direccionLectura), 10)

	// Agrega el PID y las direcciones físicas
	q := req.URL.Query()
	q.Add("pid", pidEnEjecucionStr)
	q.Add("write", direccionEscrituraStr)
	q.Add("read", direccionLecturaStr)
	q.Add("size", tamaño)
	req.URL.RawQuery = q.Encode()

	// Establece el tipo de contenido de la solicitud
	req.Header.Set("Content-Type", "text/plain")

	// Realiza la solicitud al servidor de memoria
	respuesta, err := cliente.Do(req)

	if err != nil {
		return "", "", "", ""
	}

	if respuesta.StatusCode == http.StatusNotFound {
		return "INVALID_WRITE", "", "", ""
	}

	if respuesta.StatusCode != http.StatusOK {
		return "", "", "", ""
	}

	// Lee el cuerpo de la respuesta
	data, err := io.ReadAll(respuesta.Body)
	if err != nil {
		return "", "", "", ""
	}
	return "OK", (string(data)), direccionLecturaStr, direccionEscrituraStr
}

func ioFSCreateOrDelete(nombreInterfaz string, nombreArchivo string, PID uint32, instruccionIO string) {

	//Creo estructura de request
	var requestEjecutarInstuccion = structs.RequestEjecutarInstruccionIO{
		PidDesalojado:  PID,
		NombreInterfaz: nombreInterfaz,
		Instruccion:    instruccionIO,
		NombreArchivo:  nombreArchivo,
	}

	//Convierto request a JSON
	body, err := json.Marshal(requestEjecutarInstuccion)
	if err != nil {
		return
	}

	// Envía la solicitud de ejecucion a Kernel
	config.Request(ConfigJson.Port_Kernel, ConfigJson.Ip_Kernel, "POST", "instruccionIO", body)
}

func ioFSTruncate(nombreInterfaz string, nombreArchivo string, registroTamaño string, PID uint32, registroMap8 map[string]*uint8, registroMap32 map[string]*uint32) {

	//Extrae el tamaño de la instrucción
	tamaño := extraerDatosDelRegistro(registroTamaño, registroMap8, registroMap32)

	//Crea una variable que contiene el cuerpo de la request.
	var requestEjecutarInstuccion = structs.RequestEjecutarInstruccionIO{
		PidDesalojado:  PID,
		NombreInterfaz: nombreInterfaz,
		NombreArchivo:  nombreArchivo,
		Instruccion:    "IO_FS_TRUNCATE",
		Tamaño:         tamaño,
	}

	// Convierte request a JSON
	body, err := json.Marshal(requestEjecutarInstuccion)
	if err != nil {
		return
	}

	// Envía la solicitud de ejecucion a Kernel
	config.Request(ConfigJson.Port_Kernel, ConfigJson.Ip_Kernel, "POST", "instruccionIO", body)
}

func ioFSRW(nombreInterfaz string, nombreArchivo string, regDir string, regTamaño string,
	regPuntero string, PID uint32, registroMap8 map[string]*uint8, registroMap32 map[string]*uint32, tlb *TLB, prioridadesTLB *[]ElementoPrioridad, instruccionIO string) {

	//Extrae el tamaño de la instrucción
	tamaño := extraerDatosDelRegistro(regTamaño, registroMap8, registroMap32)

	puntero := extraerDatosDelRegistro(regPuntero, registroMap8, registroMap32)

	//Traduce dirección lógica a física
	direccion, encontrado := obtenerDireccionFisica(regDir, registroMap8, registroMap32, tlb, prioridadesTLB)
	if !encontrado {
		logueano.Mensaje(Auxlogger, "No se pudo traducir el registro de dirección lógica a física.")
		return
	}

	//Crea una variable que contiene el cuerpo de la request.
	var requestEjecutarInstuccion = structs.RequestEjecutarInstruccionIO{
		PidDesalojado:  PID,
		NombreInterfaz: nombreInterfaz,
		NombreArchivo:  nombreArchivo,
		Instruccion:    instruccionIO,
		Direccion:      direccion,
		Tamaño:         tamaño,
		PunteroArchivo: puntero,
	}

	// Convierte request a JSON
	body, err := json.Marshal(requestEjecutarInstuccion)
	if err != nil {
		return
	}

	// Envía la solicitud de ejecucion a Kernel
	config.Request(ConfigJson.Port_Kernel, ConfigJson.Ip_Kernel, "POST", "instruccionIO", body)
}
