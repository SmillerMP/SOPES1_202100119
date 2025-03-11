# Usa una imagen base de Python
FROM python:3.13.2-slim

# Establece el directorio de trabajo
WORKDIR /app

# Instalar libmagic para que funcione python-magic
RUN apt-get update && apt-get install -y libmagic1 && rm -rf /var/lib/apt/lists/*

# Copia los archivos de la aplicación al contenedor
COPY . /app

# Instala las dependencias necesarias
RUN pip install --no-cache-dir -r /app/requirements.txt

# Expón el puerto que usara fastapi
EXPOSE 8000

# Comando para iniciar la aplicación con fastapi
CMD ["fastapi", "run", "main.py", "--host", "0.0.0.0", "--port", "8000"]
