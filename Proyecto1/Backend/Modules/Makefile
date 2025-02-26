obj-m += sysinfo_202100119.o

# Definir la carpeta de salida
OUT_DIR := ./build

# Crear la carpeta de salida si no existe
$(shell mkdir -p $(OUT_DIR))

# Modificar las reglas de compilación para redirigir los archivos generados
all:
	make -C /lib/modules/$(shell uname -r)/build M=$(PWD) modules
	# Mover los archivos generados a la carpeta de salida
	mv $(PWD)/*.o $(OUT_DIR)/ 
	mv $(PWD)/*.ko $(OUT_DIR)/
	mv $(PWD)/*.mod* $(OUT_DIR)/
	mv $(PWD)/*.order $(OUT_DIR)/
	mv $(PWD)/*.symvers $(OUT_DIR)/
	mv $(PWD)/.* $(OUT_DIR)/

clean:
	make -C /lib/modules/$(shell uname -r)/build M=$(PWD) clean
	# Limpiar la carpeta de salida también
	rm -rf $(OUT_DIR)
