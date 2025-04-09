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

function ctrl_c() {
    echo -e "\n${cYellow}[!] Saliendo del script... Buen Viaje [!]${cNone}"
    exit 1
}

trap ctrl_c INT

clear
echo -e "${cCyan}[!] Build a Imagenes y pusheo [!]${cNone}"
echo -e "${cRed}[!] Presiona Enter para continuar [!]${cNone}"
read -r 

pwdOriginial=$(pwd)
pwdConsumerRabbit=./Backend/Golang/Costumers/RabbitMQ

# build y push de la imagen de Consumer RabbitMQ
cd "$pwdConsumerRabbit"
docker build -t sopesp2-consumer-rabbit:1.0 .
docker tag sopesp2-consumer-rabbit:1.0 smillermp/sopesp2-consumer-rabbit:1.0 
docker push smillermp/sopesp2-consumer-rabbit:1.0 


cd "$pwdOriginial"
