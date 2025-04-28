FROM ubuntu:latest

# Crea un directorio dentro del contenedor
RUN mkdir /data

# Copia tu archivo JSON al contenedor
COPY weather_data_100.json /data/

# Definir el comando para que se ejecute al iniciar el contenedor
CMD ["cat", "/data/weather_data_100.json"]
