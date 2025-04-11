# Este archivo es el encargado de construir la imagen de docker para el proyecto
FROM golang:1.24-alpine AS build

# Cambiar el directorio de trabajo
WORKDIR /app

# Copiar el archivo de dependencias
COPY go.mod go.sum ./

# Descargar las dependencias
RUN go mod download

#  Copiar el proyecto
COPY . .

#  Compilar el proyecto
RUN go build -o app .

# Imagen de alpine para docker
FROM alpine:latest

# Cambiar el directorio de trabajo
WORKDIR /root/

# Copiar el binario del contenedor de compilación
COPY --from=build /app/app .

# Exponer el puerto 8010
EXPOSE 8010

# Ejecutar el binario
CMD ["./app"]

