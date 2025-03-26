# Imagen de rust para docker
FROM rust:1.85-slim

# Instalar pkg-config y los paquetes de desarrollo de OpenSSL
RUN apt-get update && apt-get install -y \
    pkg-config \
    libssl-dev \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# cambiar el directorio de trabajo
WORKDIR /app

# copiar el archivo de dependencias
COPY . .

# compilar el proyecto
RUN cargo build --release

# exponer el puerto 8000
EXPOSE 8000

# ejecutar el binario
CMD ["./target/release/API"]