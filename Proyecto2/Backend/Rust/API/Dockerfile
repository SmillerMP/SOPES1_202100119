# Imagen de rust para docker
FROM rust:1.85-slim

# copiar el archivo de dependencias
COPY . ./app

# cambiar el directorio de trabajo
WORKDIR /app

# compilar el proyecto
RUN cargo build --release

# exponer el puerto 8000
EXPOSE 8000

# ejecutar el binario
CMD ["./target/release/API"]