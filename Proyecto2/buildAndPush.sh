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
# read -r 

pwdOriginial=$(pwd)
pwdConsumerRabbit=./Backend/Golang/Consumer/RabbitMQ
pwdConsumerKafka=./Backend/Golang/Consumer/Kafka
pwdGoGRPC=./Backend/Golang/API_gRPC
pwdServerGRPCKafka=./Backend/Golang/gRPC_Servers/Kafka
pwdServerGRPCRabbit=./Backend/Golang/gRPC_Servers/RabbitMQ
pwdRustApi=./Backend/Rust/API
pwdGrafana=./Backend/Grafana


# build y push de la imagen de Consumer RabbitMQ
cd "$pwdConsumerRabbit"
docker build -t sopesp2-consumer-rabbit:1.0 .
docker tag sopesp2-consumer-rabbit:1.0 smillermp/sopesp2-consumer-rabbit:1.0 
docker push smillermp/sopesp2-consumer-rabbit:1.0 
cd "$pwdOriginial"

cd "$pwdConsumerKafka"
docker build -t sopesp2-consumer-kafka:1.0 .
docker tag sopesp2-consumer-kafka:1.0 smillermp/sopesp2-consumer-kafka:1.0
docker push smillermp/sopesp2-consumer-kafka:1.0
cd "$pwdOriginial"

cd "$pwdGoGRPC"
docker build -t sopesp2-go-api-grpc:1.0 .
docker tag sopesp2-go-api-grpc:1.0 smillermp/sopesp2-go-api-grpc:1.0
docker push smillermp/sopesp2-go-api-grpc:1.0
cd "$pwdOriginial"

cd "$pwdServerGRPCKafka"
docker build -t sopesp2-grpc-server-kafka:1.0 .
docker tag sopesp2-grpc-server-kafka:1.0 smillermp/sopesp2-grpc-server-kafka:1.0
docker push smillermp/sopesp2-grpc-server-kafka:1.0
cd "$pwdOriginial"

cd "$pwdServerGRPCRabbit"
docker build -t sopesp2-grpc-server-rabbit:1.0 .
docker tag sopesp2-grpc-server-rabbit:1.0 smillermp/sopesp2-grpc-server-rabbit:1.0 
docker push smillermp/sopesp2-grpc-server-rabbit:1.0 
cd "$pwdOriginial"

cd "$pwdRustApi"
docker build -t sopesp2-rust-api:1.0 .
docker tag sopesp2-rust-api:1.0 smillermp/sopesp2-rust-api:1.0
docker push smillermp/sopesp2-rust-api:1.0
cd "$pwdOriginial"

cd "$pwdGrafana"
docker build -t sopesp2-grafana:1.0 .
docker tag sopesp2-grafana:1.0 smillermp/sopesp2-grafana:1.0
docker push smillermp/sopesp2-grafana:1.0
cd "$pwdOriginial"

cd "$pwdOriginial"
