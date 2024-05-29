package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/kernel/funciones"
	"github.com/sisoputnfrba/tp-golang/kernel/logueano"
	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

//*======================================| MAIN |=======================================\\

func main() {

	// Se declara una variable para almacenar la configuración del Kernel.
	var configJson config.Kernel

	// Extrae info de config.json
	config.Iniciar("config.json", &configJson)

	// Testea la conectividad con otros modulos
	//Conectividad(configJson)

	// Configura el logger
	config.Logger("Kernel.log")

	// ======== HandleFunctions ========
	http.HandleFunc("POST /interfazConectada", handlerIniciarInterfaz)
	http.HandleFunc("POST /instruccion", handlerInstrucciones)

	//inicio el servidor de Kern
	go config.IniciarServidor(configJson.Port)

	fmt.Printf("Antes del test")

	// Espera a que haya una interfaz conectada.

	// Ahora que el servidor está en ejecución y hay una interfaz conectada, se puede iniciar el ciclo de instrucción.
	testCicloDeInstruccion(configJson)

	fmt.Printf("Despues del test")

}

//*======================================| HANDLERS |=======================================\\

// Recibe una interfazConectada y la agrega al map de interfaces conectadas.
func handlerIniciarInterfaz(w http.ResponseWriter, r *http.Request) {

	// Se crea una variable para almacenar la interfaz recibida en la solicitud.
	var requestInterfaz structs.RequestInterfaz

	// Se decodifica el cuerpo de la solicitud en formato JSON.
	err := json.NewDecoder(r.Body).Decode(&requestInterfaz)

	// Maneja el error en la decodificación.
	if err != nil {
		logueano.ErrorDecode()
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Imprime la solicitud
	fmt.Println("Request path:", requestInterfaz)

	//Guarda la interfazConectada en la lista de interfaces conectadas.
	funciones.InterfacesConectadas[requestInterfaz.NombreInterfaz] = requestInterfaz.Interfaz

	// Envía una señal al canal 'hayInterfaz' para indicar que hay una nueva interfaz conectada.

}

// TODO: implementar para los demás tipos de interfaces (cambiar tipos de datos en request y body)
func handlerInstrucciones(w http.ResponseWriter, r *http.Request) {

	// Se crea una variable para almacenar la instrucción recibida en la solicitud.
	var request structs.InstruccionIO

	// Se decodifica el cuerpo de la solicitud en formato JSON.
	err := json.NewDecoder(r.Body).Decode(&request)

	// Maneja el error de la decodificación
	if err != nil {
		logueano.ErrorDecode()
		return
	}

	// Imprime la solicitud
	fmt.Println("Request path:", request)

	// Busca la interfaz conectada en el mapa de funciones.InterfacesConectadas.
	interfazConectada, encontrado := funciones.InterfacesConectadas[request.NombreInterfaz]

	// Si no se encontró la interfazConectada de la request, se desaloja el structs.
	if !encontrado {
		funciones.DesalojarProceso(request.PidDesalojado, "EXIT")
		fmt.Println("Interfaz no conectada.")
		return
	}

	//Verifica que la instruccion sea compatible con el tipo de interfazConectada.
	isValid := funciones.ValidarInstruccion(interfazConectada.TipoInterfaz, request.Instruccion)

	// Si la instrucción no es compatible, se desaloja el proceso y se marca como "EXIT".
	if !isValid {

		funciones.DesalojarProceso(request.PidDesalojado, "EXIT")
		fmt.Println("Interfaz incompatible.")
		return
	}

	// Agrega el Proceso a la cola de bloqueados de la interfazConectada.
	interfazConectada.QueueBlock = append(interfazConectada.QueueBlock, request.PidDesalojado)
	funciones.InterfacesConectadas[request.NombreInterfaz] = interfazConectada

	// Prepara la interfazConectada para enviarla en el body.
	body, err := json.Marshal(request.UnitWorkTime)

	// Maneja los errores al crear el body.
	if err != nil {
		fmt.Printf("error codificando body: %s", err.Error())
		return
	}

	// Envía la instrucción a ejecutar a la interfazConectada (Puerto).
	respuesta := config.Request(interfazConectada.PuertoInterfaz, "localhost", "POST", request.Instruccion, body)

	// Verifica que no hubo error en la request
	if respuesta == nil {
		return
	}

	// Si la interfazConectada pudo ejecutar la instrucción, pasa el Proceso a READY.
	if respuesta.StatusCode == http.StatusOK {
		// Pasa el proceso a READY y lo quita de la lista de bloqueados.
		funciones.DesalojarProceso(request.PidDesalojado, "READY")
		return
	}
}

//*======================================| FUNC de TESTEO |=======================================\\
// !ESTO NO SE MIGRÓ A NINGÚN PAQUETE.
// Testea la conectividad con otros módulos

func testConectividad(configJson config.Kernel) {
	fmt.Println("\nIniciar Proceso:")
	funciones.IniciarProceso(configJson, "path")
	funciones.IniciarProceso(configJson, "path")
	funciones.IniciarProceso(configJson, "path")
	funciones.IniciarProceso(configJson, "path")
	fmt.Println("\nFinalizar Proceso:")
	funciones.FinalizarProceso(configJson)
	fmt.Println("\nEstado Proceso:")
	funciones.EstadoProceso(configJson)
	fmt.Println("\nListar Procesos:")
	funciones.ListarProceso(configJson)
	fmt.Println("\nDetener Planificación:")
	funciones.DetenerPlanificacion(configJson)
	fmt.Println("\nIniciar Planificación:")
	funciones.IniciarPlanificacion(configJson)
}

func testPlanificacion(configJson config.Kernel) {

	printList := func() {
		fmt.Println("readyQueue:")
		var ready []uint32
		for _, pcb := range funciones.ReadyQueue {
			ready = append(ready, pcb.PID)
		}
		fmt.Println(ready)
	}

	//
	fmt.Printf("\nSe crean 2 procesos-------------\n\n")
	for i := 0; i < 2; i++ {
		path := "procesos" + strconv.Itoa(funciones.Counter) + ".txt"
		funciones.IniciarProceso(configJson, path)
	}

	fmt.Printf("\nSe testea el planificador-------------\n\n")
	funciones.Planificador(configJson)
	printList()

	fmt.Printf("\nSe crean 2 procesos-------------\n\n")
	for i := 0; i < 2; i++ {
		path := "proceso" + strconv.Itoa(funciones.Counter) + ".txt"
		funciones.IniciarProceso(configJson, path)
	}
}

func testCicloDeInstruccion(configJson config.Kernel) {

	fmt.Printf("\nSe crean 1 proceso-------------\n\n")
	funciones.IniciarProceso(configJson, "proceso_test")

	fmt.Printf("\nSe testea el planificador-------------\n\n")
	funciones.Planificador(configJson)
}
