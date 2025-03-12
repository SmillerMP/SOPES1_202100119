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

trap ctrl_c INT

function ctrl_c() {
    echo -e "\n${cYellow}${cBold}[!] Deteniendo los contenedores...${cNone}"
    docker stop $(docker ps -aq) # Detiene todos los contenedores
    echo -e "\n\n${cRed}${cBold}[!] Deteniendo los contenedores...${cNone}"
    exit 1
}

echo -e "${cWhite}${cBold}[!] Ejecutando el Proyecto Completo ${cNone}\n"

echo -e "${cWhite}${cBold}[!] Compilando el servicio de Rust${cNone}"
cargo build --manifest-path ./Backend/Rust/service/Cargo.toml --release
echo "[ ]" > stopped_containers.json
echo -e "\n\n${cWhite}${cBold}[!] Ejecutando el servicio de Rust${cNone}"
./Backend/Rust/service/target/release/service ./Backend/Scripting/contenerized.sh 

