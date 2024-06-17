package logueano

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

// Configuro el log principal
// -------------------------------------------------------------------------

func Logger() {
	logFile, err := os.OpenFile("log.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
}

// Config de logs auxiliares
// -------------------------------------------------------------------------
func AuxLogger() *log.Logger {
	auxLogFile, err := os.OpenFile("logAux.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	mw := io.MultiWriter(os.Stdout, auxLogFile)
	return log.New(mw, "AUX: ", log.LstdFlags)
}

var AuxLog *log.Logger = AuxLogger()

// -------------------------- == LOG's CPU == --------------------------------------------------

// -------------------------- == LOG's E/S == --------------------------------------------------

// -------------------------- == LOG's KERNEL == --------------------------------------------------
// log obligatorio (1/6)
func LogNuevoProceso(nuevoPCB structs.PCB) {

	log.Printf("Se crea el proceso %d en estado %s", nuevoPCB.PID, nuevoPCB.Estado)
}

// log obligatorio (2/6)
func CambioDeEstado(pcb_estado_viejo string, pcb structs.PCB) {

	log.Printf("PID: %d - Estado anterior: %s - Estado actual: %s", pcb.PID, pcb_estado_viejo, pcb.Estado)

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

//LUEGO IMPLEMENTAR EN NUESTRO ARCHIVO NO OFICIAL DE LOGS ----------------------------

// TODO: Implementar para blockedMap.
func PidsBlock(blockQueue []structs.PCB) {
	var pids []uint32
	//Recorre la lista BLOCK y guarda sus PIDs
	for _, pcb := range blockQueue {
		pids = append(pids, pcb.PID)
	}

	fmt.Printf("Cola Block 'blockQueue' : %v", pids)
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

// -------------------------- == LOG's MEMORIA == --------------------------------------------------

// log obligatorio (1/5)
func OperoConTablaDePaginas(pid uint32, tablaDePaginas map[uint32]structs.Tabla) {

	log.Printf("PID: %d- Tamaño: %d\n", pid, len(tablaDePaginas[pid]))
}

// log obligatorio (2/5)
func AccesoTabla(pid uint32, pagina uint32, marco int) {

	log.Printf("PID: %d - Pagina: %d - Marco: %d\n", pid, pagina, marco)
}

// log obligatorio (5/5)
func AccesoEspacioUsuario(pid uint32, accion string, direccionFisica uint32, byteArraySize int) {

	log.Printf("PID: %d - Accion: %s - Direccion Fisica: %d - Tamaño: %d\n", pid, accion, direccionFisica, byteArraySize)
}

func LeerInstrucciones(memoriaInstrucciones map[uint32][]string) {

	// Imprimir las instrucciones guardadas en memoria
	AuxLog.Println("Instrucciones guardadas en memoria: ")
	for pid, instrucciones := range memoriaInstrucciones {
		AuxLog.Printf("PID: %d\n", pid)
		for _, instruccion := range instrucciones {
			fmt.Println(instruccion)
		}
		fmt.Println()
	}
}
