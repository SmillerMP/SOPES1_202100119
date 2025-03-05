#!/bin/bash

# Colores
cNone='\033[00m'
cRed='\033[01;31m'
cGreen='\033[01;32m'
cYellow='\033[01;33m'
cPurple='\033[01;35m'
cCyan='\033[01;36m'
cWhite='\033[01;37m'
cBold='\033[1m'
cUnderline='\033[4m'

echo -e "${cGreen}[!] Corriendo el script de stress en contenedores${cNone}"


# Función para crear contenedores de manera Random
function container_random() {
    if [[ $1 -ge 1 && $1 -le 4 ]]; then
        case_number=$1
    else
        case_number=$(($RANDOM % 4 + 1))
    fi

    unique_id=$(date +%s%N)
    echo "Creando contenedor con ID: $unique_id tipo de estrés: $case_number"

    case $case_number in
        1)  # CPU Stress (Limitar a 20% de un núcleo)
            docker run -d --name CPU_stress-$unique_id --cpus="0.20" containerstack/alpine-stress stress --cpu 1 &> /dev/null
            ;;
        2)  # Memory Stress (usar 2 instancias de stress de 256MB c/u, reducir a 0.1 núcleo)
            docker run -d --name Memory_stress-$unique_id --cpus="0.1" containerstack/alpine-stress stress --vm 2 &> /dev/null
            ;;
        3)  # Disk I/O Stress (Reducir la carga de I/O)
            docker run -d --name Disk_stress-$unique_id --memory="64M" containerstack/alpine-stress stress --io 1 &> /dev/null
            ;;
        4)  # HDD Stress (Limitar uso de disco a 16MB)
            docker run -d --name Hdd_stress-$unique_id --cpus="0.1" containerstack/alpine-stress stress --hdd 1 --hdd-bytes 16M &> /dev/null
            ;;
    esac
}


for i in {1..10}; do
    #  Se asegura que se cree un contenedor de cada tipo de estrés
    if [[ $i -ge 8 && $i -le 10 ]]; then
        if [ $(docker ps --filter "name=CPU_stress" -q | wc -l) -eq 0 ]; then
            container_random 1
        elif [ $(docker ps --filter "name=Memory_stress" -q | wc -l) -eq 0 ]; then
            container_random 2
        elif [ $(docker ps --filter "name=Disk_stress" -q | wc -l) -eq 0 ]; then
            container_random 3
        elif [ $(docker ps --filter "name=Hdd_stress" -q | wc -l) -eq 0 ]; then
            container_random 4
        else
            container_random 0
        fi

    else
        container_random 0
    fi
done
