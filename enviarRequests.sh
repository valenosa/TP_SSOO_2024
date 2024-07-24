#!/bin/bash

# Códigos de color
ROJO='\033[0;31m'
VERDE='\033[0;32m'
AMARILLO='\033[0;33m'
AZUL='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

iniciar_proceso() {
    echo -e "${VERDE}Iniciando proceso${NC}"
    read -p "$(echo -e ${AMARILLO}PID:${NC} )" pid
    read -p "$(echo -e ${AMARILLO}Path:${NC} )" path
    curl --location --request PUT http://$KERNEL_HOST:$KERNEL_PORT/process \
           --header 'Content-Type: application/json' \
           --data '{
           "pid": $pid,
           "path": "$path"
           }'
}

finalizar_proceso() {
    echo -e "${VERDE}Finalizando proceso${NC}"
    read -p "$(echo -e ${AMARILLO}PID:${NC} )" pid
    curl --location --request DELETE http://$KERNEL_HOST:$KERNEL_PORT/process/$pid
}

estado_proceso() {
    echo -e "${VERDE}Estado del proceso${NC}"
    read -p "$(echo -e ${AMARILLO}PID:${NC} )" pid
    curl --location --request GET http://$KERNEL_HOST:$KERNEL_PORT/process/$pid
}

listar_procesos() {
    echo -e "${VERDE}Listar procesos${NC}"
    curl --location --request GET http://$KERNEL_HOST:$KERNEL_PORT/process
}

iniciar_plani() {
    echo -e ${VERDE}"Iniciando planificación${NC}"
    curl --location --request PUT http://$KERNEL_HOST:$KERNEL_PORT/plani
}
detener_plani() {
    echo -e ${VERDE}"Deteniendo planificación${NC}"
    curl --location --request DELETE http://$KERNEL_HOST:$KERNEL_PORT/plani
}


print_kernel_host() {
    echo -e "${VERDE}Kernel Host:${NC} $KERNEL_HOST"
}

print_kernel_port() {
    echo -e "${VERDE}Kernel Port:${NC} $KERNEL_PORT"
}


while true; do
    echo -e "${AMARILLO}1.${NC} Iniciar Proceso"
    echo -e "${AMARILLO}2.${NC} Finalizar Proceso"
    echo -e "${AMARILLO}3.${NC} Estado Proceso"
    echo -e "${AMARILLO}4.${NC} Listar Procesos"
    echo -e "${AMARILLO}5.${NC} Iniciar Planificación"
    echo -e "${AMARILLO}6.${NC} Detener Planificación"
    echo -e "${AMARILLO}k.${NC}  Kernel Data"
    echo -e "${ROJO}s.${NC} Salir"
    echo 
    read -p "$(echo -e ${AMARILLO}Opción:${NC} )" opcion

    case $opcion in
        1) iniciar_proceso;;
        2) finalizar_proceso ;;
        3) estado_proceso ;;
        4) listar_procesos ;;
        5) iniciar_plani ;;
        6) detener_plani ;;
        k) print_kernel_host; print_kernel_port ;;
        s) echo -e "${ROJO}Saliendo...${NC}"; break ;;
        *) echo -e "${ROJO}Opción no válida${NC}" ;;
    esac
    echo
done