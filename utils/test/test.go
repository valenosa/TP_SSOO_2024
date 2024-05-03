// Paquete para funciones de testeo
package test

import (
	"fmt"

	"github.com/sisoputnfrba/tp-golang/utils/APIs/kernel-cpu/planificacion"
	"github.com/sisoputnfrba/tp-golang/utils/APIs/kernel-memoria/proceso"
	"github.com/sisoputnfrba/tp-golang/utils/config"
)

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
