package logueano

import (
	"io"
	"log"
	"os"

	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

// -------------------------- == LOG's CONFIG == -----------------------------------------------------------

// Configuro el log principal
func Logger(path string) {
	logFile, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
}

// Config de logs auxiliares

// Inicializa y devuelve un nuevo Logger para un módulo específico
func NewLogger(modulo string) (*log.Logger, error) {

	// Archivo de log auxiliar
	auxLogFile, err := os.OpenFile(modulo+"Aux.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644) //"logs/"+modulo+"/"+

	if err != nil {
		return nil, err
	}

	// Logger auxiliar
	auxLogger := log.New(auxLogFile, "", log.LstdFlags)

	auxLogger.SetOutput(auxLogFile)

	return auxLogger, nil
}

func InitAuxLog(modulo string) *log.Logger {
	var err error
	auxLogger, err := NewLogger(modulo)
	if err != nil {
		panic(err)
	}

	return auxLogger
}

// -------------------------- == LOG's AUXILIARES GRALES == -----------------------------------------------------------
func Error(auxLog *log.Logger, err error) {

	auxLog.Println("Error: ", err)
}

func Mensaje(auxLog *log.Logger, mensaje string) {

	auxLog.Println(mensaje)
}

func MensajeConFormato(auxLog *log.Logger, mensaje string, args ...interface{}) {

	auxLog.Printf(mensaje, args...)
}

// -------------------------- == LOG's CPU == -----------------------------------------------------------
// log obligatorio (1/6)
func FetchInstruccion(PID uint32, PC uint32) {
	log.Printf("PID: %d - FETCH - Program Counter: %d", PID, PC)
}

// log obligatorio (2/6)
func EjecucionInstruccion(pcb structs.PCB, variable []string) {

	log.Println("PID: ", pcb.PID, " - Ejecutando: ", variable[0], " - Parametros: ", variable[1:])
}

// log obligatorio (3...4/6)
func TLBAccion(pid uint32, encontrado bool, pagina uint32) {

	if encontrado {
		log.Printf("PID: %d - TLB HIT - Pagina %d", pid, pagina)
	} else {
		log.Printf("PID: %d - TLB MISS - Pagina %d", pid, pagina)
	}
}

// log obligatorio (5/6)
func ObtenerMarcolg(pid uint32, encontrado bool, pagina uint32, marco uint32) {
	if encontrado {
		log.Printf("PID: %d - OBTENER MARCO - Página: %d - Marco: %d", pid, pagina, marco)
	}
}

// log obligatorio (6/6)
func LecturaEscritura(pcb structs.PCB, accion string, direccionFisica string, valor []byte) {

	log.Printf("PID: %d - Accion: %s - Direccion Fisica: %s - Valor: %v", pcb.PID, accion, direccionFisica, valor)
}

//log´s auxiliares------------------------------------------------------

// -------------------------- == LOG's E/S == -----------------------------------------------------------
// log obligatorio (1/6)
func Operacion(pid uint32, operacion string) {
	log.Printf("PID: %d - Operacion %s", pid, operacion)
}

// log obligatorio (2/6)
func CrearArchivo(pid uint32, nombre string) {
	log.Printf("PID: %d - Crear Archivo %s", pid, nombre)
}

// log obligatorio (3/6)
func EliminarArchivo(pid uint32, nombre string) {
	log.Printf("PID: %d - Eliminar Archivo %s", pid, nombre)
}

// log obligatorio (4/6)
func TruncarArchivo(pid uint32, nombre string, tamaño uint32) {
	log.Printf("PID: %d - Truncar Archivo %s - Tamaño: %d", pid, nombre, tamaño)
}

// log obligatorio (5...6/6)
func LeerEscribirArchivo(pid uint32, accion string, nombre string, tamaño int, puntero uint32) {
	if accion == "LEER" {
		log.Printf("PID: %d - Leer Archivo %s - Tamaño a Leer: %d - Puntero Archivo: %d", pid, nombre, tamaño, puntero)
	} else {
		log.Printf("PID: %d - Escribir Archivo %s - Tamaño a Escribir: %d - Puntero Archivo: %d", pid, nombre, tamaño, puntero)
	}
}

// -------------------------- == LOG's KERNEL == -----------------------------------------------------------
// log obligatorio (1/6)
func NuevoProceso(nuevoPCB structs.PCB) {

	log.Printf("Se crea el proceso %d en estado %s", nuevoPCB.PID, nuevoPCB.Estado)
}

func CambioDeEstado(pcb_estado_anterior string, pcb_estado_nuevo string, pid uint32) {

	log.Printf("PID: %d - Estado anterior: %s - Estado actual: %s", pid, pcb_estado_anterior, pcb_estado_nuevo)
}

// log obligatorio (3/6)
func PidsReady(readyQueue []structs.PCB) {
	var pids []uint32
	//Recorre la lista READY y guarda sus PIDs
	for _, pcb := range readyQueue {
		pids = append(pids, pcb.PID)
	}

	log.Printf("Cola Ready 'ListaREADY' : %v", pids)
}

// log obligatorio (4/6)
func FinDeProceso(pid uint32, motivoDeFinalizacion string) {

	log.Printf("Finaliza el proceso: %d - Motivo: %s", pid, motivoDeFinalizacion)
}

// log obligatorio (5/6)
func FinDeQuantum(pcb structs.PCB) {

	log.Printf("PID: %d - Desalojado por fin de Quantum", pcb.PID)
}

// log obligatorio (6/6)
func MotivoBloqueo(pid uint32, motivo string) {

	log.Printf("PID: %d - Bloqueado por: %s", pid, motivo)
}

//log´s auxiliares------------------------------------------------------

func PidsBlock(auxLog *log.Logger, blockedQueue map[uint32]structs.PCB) {
	var pids []uint32
	//Recorre la lista BLOCKED y guarda sus PIDs
	for _, pcb := range blockedQueue {
		pids = append(pids, pcb.PID)
	}

	auxLog.Printf("Cola Blocked 'MapBLOCK' : %v", pids)

}

func PidsNew(auxLog *log.Logger, newQueue []structs.PCB) {
	var pids []uint32
	//Recorre la lista NEW y guarda sus PIDs
	for _, pcb := range newQueue {
		pids = append(pids, pcb.PID)
	}

	auxLog.Printf("Cola New 'ListaNEW' : %v", pids)
}

func PidsExit(auxLog *log.Logger, exitQueue []structs.PCB) {
	var pids []uint32
	//Recorre la lista EXIT y guarda sus PIDs
	for _, pcb := range exitQueue {
		pids = append(pids, pcb.PID)
	}

	auxLog.Printf("Cola Exit 'ListaEXIT' : %v", pids)
}

func PidsReadyPrioritarios(auxLog *log.Logger, pcb structs.PCB) {

	auxLog.Println("Se agregó el proceso", pcb.PID, "a la cola de READY_PRIORITARIO")
}

//* ==================================| LOG's MEMORIA |==================================\\

// log obligatorio (1/5)
func OperoConTablaDePaginas(pid uint32, tablaDePaginas map[uint32]structs.Tabla) {

	log.Printf("PID: %d- Tamaño: %d\n", pid, len(tablaDePaginas[pid]))
}

// log obligatorio (2/5)
func AccesoTabla(pid uint32, pagina uint32, marco int) {

	log.Printf("PID: %d - Pagina: %d - Marco: %d\n", pid, pagina, marco)
}

// log obligatorio ((3...4)/6)
func CambioDeTamaño(pid uint32, lenOriginal int, accion string, tablaDePaginas *map[uint32]structs.Tabla) {

	log.Printf("PID: %d - Tamaño Actual: %d - Tamaño a %s: %d\n", pid, lenOriginal, accion, len((*tablaDePaginas)[pid]))
}

// log obligatorio (5/5)
func AccesoEspacioUsuario(pid uint32, accion string, direccionFisica uint32, byteArraySize int) {

	log.Printf("PID: %d - Accion: %s - Direccion Fisica: %d - Tamaño: %d\n", pid, accion, direccionFisica, byteArraySize)
}

// log´s auxiliares------------------------------------------------------
func LeerInstrucciones(auxLog *log.Logger, memoriaInstrucciones map[uint32][]string, pid uint32) {

	// Imprimir las instrucciones guardadas en memoria
	auxLog.Println("Instrucciones guardadas en memoria: ")
	auxLog.Printf("PID: %d\n", pid)
	for _, instruccion := range memoriaInstrucciones[pid] {
		auxLog.Println(instruccion)
	}
}
