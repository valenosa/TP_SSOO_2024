#!/bin/bash

# Códigos de color
ROJO='\033[0;31m'
VERDE='\033[0;32m'
MARRON='\033[0;33m'
AZUL='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
AMARILLO='\033[1;33m'
NEGRO='\033[0;30m'
GRIS_OSCURO='\033[1;30m'
GRIS_CLARO='\033[0;37m'
ROJO_CLARO='\033[1;31m'
VERDE_CLARO='\033[1;32m'
AZUL_CLARO='\033[1;34m'
PURPURA_CLARO='\033[1;35m'
CIAN_CLARO='\033[1;36m'
BLANCO='\033[1;37m'
NC='\033[0m' # No Color

#Las variables de entorno creadas son locales de la ejecución del script actual. Una vez que se deja de ejecutar, muere.
# si ejecutas source ./modificarIP.sh las variables de entorno son creadas localmente en la terminal, por lo que podes matar el script, y al levantarlo nuevamente siguen estando. Mueren cuando ripea la terminal.

# Los configs vienen pre-seteados en localhost y ya no usan el "caso default" que setea a localhost si la variable estaba vacía (se puede cambiar).
# Se puede escribir localhost o l 

modificar() {
    local var_name=$2
    local current_value=$(eval echo \$$var_name)
    echo -e "${GRIS_OSCURO}$2 Actual:${NC} $current_value"
    read -p "$(echo -e ${VERDE}$2 Nuevo:${NC} )" new_value
    #si no escribo nada, se mantiene el valor actual
    if [ -z "$new_value" ]; then
        new_value=$current_value
    fi
    #si escribo l, pone localhost
    if [ "$new_value" = "l" ]; then
      new_value="localhost"
    fi
    eval export $var_name=$new_value
}

# Lo hice para que con un comando se escriba todo, pero se podria hacer que se vaya escribiendo mientras vas ejecutando los comandos
escribir() {
  # Update parametro in archivo Kernel_DL.json 
sed -i "s/\"port\": .*,/\"port\": $KERNEL_PORT,/" Kernel_DL.json
sed -i "s/\"ip_memory\": .*,/\"ip_memory\": \"$MEM_HOST\",/" Kernel_DL.json
sed -i "s/\"port_memory\": .*,/\"port_memory\": $MEM_PORT,/" Kernel_DL.json
sed -i "s/\"ip_cpu\": .*,/\"ip_cpu\": \"$CPU_HOST\",/" Kernel_DL.json
sed -i "s/\"port_cpu\": .*,/\"port_cpu\": $CPU_PORT,/" Kernel_DL.json

# Update parametro in archivo Kernel_FS.json
sed -i "s/\"port\": .*,/\"port\": $KERNEL_PORT,/" Kernel_FS.json
sed -i "s/\"ip_memory\": .*,/\"ip_memory\": \"$MEM_HOST\",/" Kernel_FS.json
sed -i "s/\"port_memory\": .*,/\"port_memory\": $MEM_PORT,/" Kernel_FS.json
sed -i "s/\"ip_cpu\": .*,/\"ip_cpu\": \"$CPU_HOST\",/" Kernel_FS.json
sed -i "s/\"port_cpu\": .*,/\"port_cpu\": $CPU_PORT,/" Kernel_FS.json

# Update parametro in archivo Kernel_IO.json
sed -i "s/\"port\": .*,/\"port\": $KERNEL_PORT,/" Kernel_IO.json
sed -i "s/\"ip_memory\": .*,/\"ip_memory\": \"$MEM_HOST\",/" Kernel_IO.json
sed -i "s/\"port_memory\": .*,/\"port_memory\": $MEM_PORT,/" Kernel_IO.json
sed -i "s/\"ip_cpu\": .*,/\"ip_cpu\": \"$CPU_HOST\",/" Kernel_IO.json
sed -i "s/\"port_cpu\": .*,/\"port_cpu\": $CPU_PORT,/" Kernel_IO.json

# Update parametro in archivo Kernel_Mem.json
sed -i "s/\"port\": .*,/\"port\": $KERNEL_PORT,/" Kernel_Mem.json
sed -i "s/\"ip_memory\": .*,/\"ip_memory\": \"$MEM_HOST\",/" Kernel_Mem.json
sed -i "s/\"port_memory\": .*,/\"port_memory\": $MEM_PORT,/" Kernel_Mem.json
sed -i "s/\"ip_cpu\": .*,/\"ip_cpu\": \"$CPU_HOST\",/" Kernel_Mem.json
sed -i "s/\"port_cpu\": .*,/\"port_cpu\": $CPU_PORT,/" Kernel_Mem.json

# Update parametro in archivo Kernel_Plani.json
sed -i "s/\"port\": .*,/\"port\": $KERNEL_PORT,/" Kernel_Plani.json
sed -i "s/\"ip_memory\": .*,/\"ip_memory\": \"$MEM_HOST\",/" Kernel_Plani.json
sed -i "s/\"port_memory\": .*,/\"port_memory\": $MEM_PORT,/" Kernel_Plani.json
sed -i "s/\"ip_cpu\": .*,/\"ip_cpu\": \"$CPU_HOST\",/" Kernel_Plani.json
sed -i "s/\"port_cpu\": .*,/\"port_cpu\": $CPU_PORT,/" Kernel_Plani.json


# Update parametro in archivo Kernel_SE.json 
sed -i "s/\"port\": .*,/\"port\": $KERNEL_PORT,/" Kernel_SE.json
sed -i "s/\"ip_memory\": .*,/\"ip_memory\": \"$MEM_HOST\",/" Kernel_SE.json
sed -i "s/\"port_memory\": .*,/\"port_memory\": $MEM_PORT,/" Kernel_SE.json
sed -i "s/\"ip_cpu\": .*,/\"ip_cpu\": \"$CPU_HOST\",/" Kernel_SE.json
sed -i "s/\"port_cpu\": .*,/\"port_cpu\": $CPU_PORT,/" Kernel_SE.json
}

export MEM_HOST=localhost; 
export CPU_HOST=localhost; 
export KERNEL_PORT=8001; 
export MEM_PORT=8002;
export CPU_PORT=8003;

while true; do
    echo -e "${AMARILLO}1.${NC} Modificar IP Memoria"
    echo -e "${AMARILLO}2.${NC} Modificar IP CPU"
    echo -e "${AMARILLO}3.${NC} Modificar Puerto Kernel"
    echo -e "${AMARILLO}4.${NC} Modificar Puerto Memoria"
    echo -e "${AMARILLO}5.${NC} Modificar Puerto CPU"
    echo -e "${AMARILLO}p.${NC} Print Settings"
    echo -e "${AMARILLO}d.${NC} Default Settings"
    echo -e "${AMARILLO}e.${NC} Escribir Archivos"
    echo -e "${ROJO}s.${NC} Salir"
    read -p "$(echo -e ${AMARILLO}Opción:${NC} )" opcion

    case $opcion in
        1) modificar "IP" "MEM_HOST";;
        2) modificar "IP" "CPU_HOST" ;;
        3) modificar "Puerto" "KERNEL_PORT" ;;
        4) modificar "Puerto" "MEM_PORT" ;;
        5) modificar "Puerto" "CPU_PORT" ;;
        p) echo -e "${GRIS_OSCURO}IP Memoria:${NC} $MEM_HOST" 
           echo -e "${GRIS_OSCURO}IP CPU:${NC} $CPU_HOST"
           echo -e "${GRIS_OSCURO}Puerto Kernel:${NC} $KERNEL_PORT"
           echo -e "${GRIS_OSCURO}Puerto Memoria:${NC} $MEM_PORT"
           echo -e "${GRIS_OSCURO}Puerto CPU:${NC} $CPU_PORT";;
        d) export MEM_HOST=localhost; 
           export CPU_HOST=localhost; 
           export KERNEL_PORT=8001; 
           export MEM_PORT=8002;
           export CPU_PORT=8003; ;;
        e) escribir;;
        s) echo -e "${ROJO}Saliendo...${NC}"; break ;;
        *) echo -e "${ROJO}Opción no válida${NC}" ;;
    esac
    echo
done

