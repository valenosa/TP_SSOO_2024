package structs

// =====================================| PROCESO |========================================================\\

//	Kernel -> Cliente
//
// Estructura que contiene el path del archivo que se utilizará como base para ejecutar un nuevo proceso y su PID asociado.
type BodyIniciarProceso struct {
	// Path del archivo que se utilizará como base para ejecutar un nuevo proceso
	PID  uint32 `json:"pid"`
	Path string `json:"path"`
}

type RequestIniciarProceso struct {
	Path string `json:"path"`
}

// Kernel -> Cliente
type ResponseEstadoProceso struct {
	// Path del archivo que se utilizará como base para ejecutar un nuevo proceso
	State string `json:"state"`
}

// CPU, Kernel.
// Estructura base de un proceso.
type PCB struct {
	PID     uint32
	Quantum uint16
	Estado  string
	RegistrosUsoGeneral
}

// CPU
// Estructura de los registros de uso general (para tener info del contexto de ejecución de cada PCB)
type RegistrosUsoGeneral struct {
	PC  uint32
	AX  uint8
	BX  uint8
	CX  uint8
	DX  uint8
	EAX uint32
	EBX uint32
	ECX uint32
	EDX uint32
	SI  uint32
	DI  uint32
}

//=====================================|  |=====================================\\

// Kernel, Memoria
// Estructura de respuesta al iniciar un proceso
type ResponseListarProceso struct {
	PID    uint32 `json:"pid"`
	Estado string `json:"estado"`
}

// CPU, Kernel
// Estructura de comunicacion al desalojar un proceso
type RespuestaDispatch struct {
	MotivoDeDesalojo string
	PCB              PCB
}

// ====================================| RECURSOS | ====================================\\

type Recurso struct {
	Instancias int
	ListaBlock []uint32
}

//=====================================| I/O |=====================================\\

// Estructura basica de InterfazIO que se guardara Kernel
type Interfaz struct {
	TipoInterfaz   string
	PuertoInterfaz int
}

// Estructura de comunicacion al conectar una interfaz (Contiene su nombre/identificador y lo necesario para validar en Kernel)
type RequestConectarInterfazIO struct {
	NombreInterfaz string
	Interfaz       Interfaz
}

// TODO: Implementar para DialFS
// Estructura de comunicacion entre CPU y Kernel para ejecutar una instruccion de I/O
type RequestEjecutarInstruccionIO struct {
	PidDesalojado     uint32
	NombreInterfaz    string
	Instruccion       string
	UnitWorkTime      int
	RegistroDireccion uint32
	RegistroTamaño    uint8
}

// Estructura que comunica Kernel con CPU y CPU con memoria para la instruccion STDIN.
type RequestInputSTDIN struct {
	TextoUsuario      string
	RegistroDireccion uint32
	//? Tambien deberia estar el pid?
}
