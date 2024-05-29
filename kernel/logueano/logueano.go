package logueano

import (
	"fmt"
	"log"

	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

// -------------------------- LOG's --------------------------------------------------
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

func ErrorDecode() {

	fmt.Printf("Error al decodificar request body: ")
}
