#include <linux/fs.h>
#include <linux/init.h>
#include <linux/kernel.h>
#include <linux/module.h>
#include <linux/seq_file.h>
#include <linux/stat.h>
#include <linux/string.h>
#include <linux/uaccess.h>
#include <linux/mm.h>
#include <linux/sysinfo.h>
#include <linux/sched/signal.h>
#include <linux/proc_fs.h>
#include <asm/uaccess.h>

MODULE_LICENSE("GPL");
MODULE_DESCRIPTION("Monitor de CPU y RAM");
MODULE_AUTHOR("Samuel Isaí Muñoz Pereira");

// Función para calcular el porcentaje de CPU usado
static int calcularPorcentajeCPU(void) {
    struct file *archivo;
    char lectura[256];
    int usuario, nice, idle, iowait, irq, softirq, steal, guest, guest_nice;
    int total, porcentaje;

    archivo = filp_open("/proc/stat", O_RDONLY, 0);
    if (IS_ERR(archivo)) {
        printk(KERN_INFO "Error al abrir el archivo /proc/stat\n");
        return -1;
    }

    memset(lectura, 0, 256);
    kernel_read(archivo, lectura, 256, &archivo->f_pos);

    sscanf(lectura, "cpu %d %d %d %d %d %d %d %d %d", &usuario, &nice, &idle, &iowait, &irq, &softirq, &steal, &guest, &guest_nice);

    total = usuario + nice + idle + iowait + irq + softirq + steal + guest + guest_nice;
    porcentaje = (total - idle) * 100 / total;

    filp_close(archivo, NULL);
    return porcentaje;
}

// Función para calcular el porcentaje de memoria usado por un proceso
static int calcularPorcentajeRAM(pid_t pid) {
    struct file *archivo;
    char buf[512];
    int memoria_total, vmrss;
    char *linea;
    int ret;

    // Obtener la memoria total del sistema
    struct sysinfo info;
    si_meminfo(&info);
    memoria_total = info.totalram * info.mem_unit;

    archivo = filp_open("/proc/1/status", O_RDONLY, 0);  // Abrir el archivo status de un proceso
    if (IS_ERR(archivo)) {
        printk(KERN_INFO "Error al abrir el archivo /proc/%d/status\n", pid);
        return -1;
    }

    memset(buf, 0, 512);
    kernel_read(archivo, buf, 512, &archivo->f_pos);

    // Buscar la línea que contiene "VmRSS"
    linea = strstr(buf, "VmRSS:");
    if (linea) {
        ret = sscanf(linea, "VmRSS: %d kB", &vmrss);
        if (ret != 1) {
            printk(KERN_INFO "Error al leer VmRSS para el proceso %d\n", pid);
            filp_close(archivo, NULL);
            return -1;
        }

        filp_close(archivo, NULL);
        // Calcular el porcentaje de memoria utilizado por el proceso
        return (vmrss * 100) / (memoria_total / 1024);  // Convertir la memoria total a KB para comparar
    }

    filp_close(archivo, NULL);
    printk(KERN_INFO "No se encontró VmRSS para el proceso %d\n", pid);
    return -1;
}

static int write_file(struct seq_file *archivo, void *v) {
    struct task_struct *task;
    int porcentaje_cpu = calcularPorcentajeCPU();
    if (porcentaje_cpu == -1) {
        seq_printf(archivo, "{\"error\": \"Error al leer el archivo\"}\n");
        return -1;
    }

    seq_printf(archivo, "{\n  \"percentage_used\": %d,\n  \"tasks\": [\n", porcentaje_cpu);

    for_each_process(task) {
        if (strcmp(task->parent->comm, "containerd-shim") == 0 || strcmp(task->parent->comm, "dockerd") == 0) {
            // Este proceso pertenece a un contenedor Docker
            int porcentaje_memoria = calcularPorcentajeRAM(task->pid);
            if (porcentaje_memoria == -1) {
                porcentaje_memoria = 0;  // Si no se puede calcular, poner 0
            }

            seq_printf(archivo, "    {\n");
            seq_printf(archivo, "      \"pid\": %d,\n", task->pid);
            seq_printf(archivo, "      \"name\": \"%s\",\n", task->comm);
            seq_printf(archivo, "      \"user\": %d,\n", task->cred->uid.val);
            seq_printf(archivo, "      \"father\": %d,\n", task->parent->pid);
            seq_printf(archivo, "      \"memory_percentage\": %d\n", porcentaje_memoria);
            seq_printf(archivo, "    },\n");
        }
    }

    seq_printf(archivo, "\n  ]\n}\n");
    return 0;
}

static int al_abrir(struct inode *inode, struct file *file) {
    return single_open(file, write_file, NULL);
}

static const struct proc_ops sysinfo_ops = {
    .proc_open = al_abrir,
    .proc_read = seq_read,
    .proc_lseek = seq_lseek,
    .proc_release = single_release,
};

static int _insert(void) {
    proc_create("cpu_test", 0, NULL, &sysinfo_ops);
    printk(KERN_INFO "Se insertó el módulo sysinfo_202100119\n");
    return 0;
}

static void _remove(void) {
    remove_proc_entry("cpu_test", NULL);
    printk(KERN_INFO "Módulo sysinfo_202100119 removido correctamente\n");
}

module_init(_insert);
module_exit(_remove);
