package logueano

import (
	"fmt"
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
	auxLog.Println(err)
}

func Mensaje(auxLog *log.Logger, mensaje string) {
	auxLog.Println(mensaje)
}

// -------------------------- == LOG's CPU == -----------------------------------------------------------

//log´s auxiliares------------------------------------------------------

// -------------------------- == LOG's E/S == -----------------------------------------------------------

//log´s auxiliares------------------------------------------------------

// -------------------------- == LOG's KERNEL == -----------------------------------------------------------
// log obligatorio (1/6)
func LogNuevoProceso(nuevoPCB structs.PCB) {

	log.Printf("Se crea el proceso %d en estado %s", nuevoPCB.PID, nuevoPCB.Estado)
}

// log obligatorio (2/6)
func CambioDeEstado(pcb_estado_viejo string, pcb structs.PCB) {

	log.Printf("PID: %d - Estado anterior: %s - Estado actual: %s", pcb.PID, pcb_estado_viejo, pcb.Estado)

}

func CambioDeEstadoInverso(pcb structs.PCB, pcb_estado_nuevo string) {

	log.Printf("PID: %d - Estado anterior: %s - Estado actual: %s", pcb.PID, pcb.Estado, pcb_estado_nuevo)

}

// log obligatorio (3/6)
func PidsReady(readyQueue []structs.PCB) {
	var pids []uint32
	//Recorre la lista READY y guarda sus PIDs
	for _, pcb := range readyQueue {
		pids = append(pids, pcb.PID)
	}

	log.Printf("Cola Ready 'readyQueue' : %v", pids)
}

// log obligatorio (4/6)
func FinDeProceso(pcb structs.PCB, motivoDeFinalizacion string) {

	log.Printf("Finaliza el proceso: %d - Motivo: %s", pcb.PID, motivoDeFinalizacion)

}

// log obligatorio (5/6)
func FinDeQuantum(pcb structs.PCB) {

	log.Printf("PID: %d - Desalojado por fin de Quantum", pcb.PID)
}

//log´s auxiliares------------------------------------------------------

// TODO: Implementar para blockedMap.
func PidsBlock(auxLog *log.Logger, blockedQueue map[uint32]structs.PCB) {
	var pids []uint32
	//Recorre la lista BLOCKED y guarda sus PIDs
	for _, pcb := range blockedQueue {
		pids = append(pids, pcb.PID)
	}

	auxLog.Printf("Cola Blocked 'blockedQueue' : %v", pids)

}

// log para el manejo de listas EXEC
func PidsExec(ExecQueue []structs.PCB) {
	var pids []uint32
	//Recorre la lista EXEC y guarda sus PIDs
	for _, pcb := range ExecQueue {
		pids = append(pids, pcb.PID)
	}

	fmt.Printf("Cola Executing 'ExecQueue' : %v", pids)
}

func EsperaNuevosProcesos() {

	fmt.Println("Esperando nuevos procesos...")

}

func IndicarPath(auxLog *log.Logger, path string) {

	auxLog.Printf("Path: %s\n", path)
}

func IndicarPID(auxLog *log.Logger, pid uint32) {
	auxLog.Printf("PID: %d\n", pid)
}

func PidsReadyPrioritarios(auxLog *log.Logger, pcb structs.PCB) {

	auxLog.Println("Se agregó el proceso", pcb.PID, "a la cola de READY_PRIORITARIO")
}

func EnviarInterrupcion(auxLog *log.Logger, tipoDeInterrupcion string) {

	auxLog.Printf("Interrupción tipo %s enviada correctamente.\n", tipoDeInterrupcion)
}

// -------------------------- == LOG's MEMORIA == -----------------------------------------------------------

// log obligatorio (1/5)
func OperoConTablaDePaginas(pid uint32, tablaDePaginas map[uint32]structs.Tabla) {

	log.Printf("PID: %d- Tamaño: %d\n", pid, len(tablaDePaginas[pid]))
}

// log obligatorio (2/5)
func AccesoTabla(pid uint32, pagina uint32, marco int) {

	log.Printf("PID: %d - Pagina: %d - Marco: %d\n", pid, pagina, marco)
}

// log obligatorio ((3...4)/6)
func CambioDeTamaño(pid uint32, lenOriginal int, accion string, tablaDePaginas *map[uint32]structs.Tabla) string {

	log.Printf("PID: %d - Tamaño Actual: %d - Tamaño a %s: %d\n", pid, lenOriginal, accion, len((*tablaDePaginas)[pid]))

	return "OK"
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
