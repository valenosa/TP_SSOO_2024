// Paquete para funciones de testeo
package test

import (
	"fmt"

	"github.com/sisoputnfrba/tp-golang/utils/APIs/kernel-cpu/planificacion"
	"github.com/sisoputnfrba/tp-golang/utils/APIs/kernel-memoria/proceso"
	"github.com/sisoputnfrba/tp-golang/utils/config"
)

//FUNCIONES PARA TESTEAR////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Utilizado para testear "IniciarConfiguracion()"
func Print(configJson config.Kernel) {

	fmt.Println("port: ", configJson.Port)
	fmt.Println("ip_memory: ", configJson.Ip_Memory)
	fmt.Println("port_memory: ", configJson.Port_Memory)
	fmt.Println("ip_cpu: ", configJson.Ip_CPU)
	fmt.Println("port_cpu: ", configJson.Port_CPU)
	fmt.Println("planning_algorithm: ", configJson.Planning_Algorithm)
	fmt.Println("quantum: ", configJson.Quantum)
	fmt.Println("resources: ", configJson.Resources)
	fmt.Println("resource_instances: ", configJson.Resource_Instances)
	fmt.Println("multiprogramming: ", configJson.Multiprogramming)
}

// Testea la conectividad con otros módulos
func Conectividad(configJson config.Kernel) {
	fmt.Println("\nIniciar Proceso:")
	proceso.Iniciar(configJson)
	proceso.Iniciar(configJson)
	proceso.Iniciar(configJson)
	proceso.Iniciar(configJson)
	fmt.Println("\nFinalizar Proceso:")
	proceso.Finalizar(configJson)
	fmt.Println("\nEstado Proceso:")
	proceso.Estado(configJson)
	fmt.Println("\nListar Procesos:")
	proceso.Listar(configJson)
	fmt.Println("\nDetener Planificación:")
	planificacion.Detener(configJson)
	fmt.Println("\nIniciar Planificación:")
	planificacion.Iniciar(configJson)
}
