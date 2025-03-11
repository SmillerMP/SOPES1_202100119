#include <linux/fs.h>
#include <linux/init.h>
#include <linux/kernel.h>
#include <linux/module.h>
#include <linux/seq_file.h>
#include <linux/proc_fs.h>
#include <linux/delay.h>
#include <linux/workqueue.h>
#include <linux/sched/signal.h>
#include <linux/mm.h>
#include <linux/uaccess.h>
#include <linux/slab.h>
#include <linux/cgroup.h>

#define MAX_CONTAINERS 100
#define ID_CONTAINER_SIZE 12

MODULE_LICENSE("GPL");
MODULE_DESCRIPTION("Monitor en tiempo real de CPU y RAM");
MODULE_AUTHOR("Samuel Isaí Muñoz Pereira");

struct containers_struct_info
{
    char id[20];
    char name[128];
    char command[128];
    int pid;
    unsigned long long cpu_use;
    unsigned long memory_use;
    unsigned long long io_use;
    unsigned long long disk_use;
};

static struct delayed_work cpu_work;
static struct workqueue_struct *cpu_queue;
static int cpu_general_use = 0; // Variable global para el uso del CPU

// Variables para almacenar la última lectura de CPU
static int prev_total = 0, prev_idle = 0;

// Función para leer `/proc/stat` y calcular el uso del CPU
static void update_cpu_usage(struct work_struct *work)
{
    struct file *archivo;
    char lectura[256];
    loff_t pos = 0;
    int usuario, nice, system, idle, iowait, irq, softirq, steal, guest, guest_nice;
    int total, diff_idle, diff_total;
    ssize_t bytes_leidos;

    archivo = filp_open("/proc/stat", O_RDONLY, 0);
    if (IS_ERR(archivo))
    {
        printk(KERN_ERR "Error al abrir /proc/stat\n");
        return;
    }

    memset(lectura, 0, sizeof(lectura));
    bytes_leidos = kernel_read(archivo, lectura, sizeof(lectura) - 1, &pos);
    filp_close(archivo, NULL);

    if (bytes_leidos <= 0)
    {
        printk(KERN_ERR "Error al leer /proc/stat\n");
        return;
    }

    lectura[bytes_leidos] = '\0';

    // Extraer valores de CPU
    if (sscanf(lectura, "cpu %d %d %d %d %d %d %d %d %d %d",
               &usuario, &nice, &system, &idle, &iowait, &irq, &softirq, &steal, &guest, &guest_nice) != 10)
    {
        printk(KERN_ERR "Error al parsear /proc/stat\n");
        return;
    }

    total = usuario + nice + system + idle + iowait + irq + softirq + steal + guest + guest_nice;
    diff_total = total - prev_total;
    diff_idle = idle - prev_idle;

    if (diff_total > 0)
    {
        cpu_general_use = (100 * (diff_total - diff_idle)) / diff_total;
    }

    prev_total = total;
    prev_idle = idle;

    // Programar la próxima ejecución en 1 segundo
    queue_delayed_work(cpu_queue, &cpu_work, msecs_to_jiffies(1000));
}

// Función para obtener la memoria RAM (total, libre y en uso)
static void get_memory_usage_general(unsigned long *total, unsigned long *libre, unsigned long *uso)
{
    // Obtener memoria total del sistema en MB
    *total = ((global_node_page_state(NR_INACTIVE_ANON) +
               global_node_page_state(NR_ACTIVE_ANON) +
               global_node_page_state(NR_INACTIVE_FILE) +
               global_node_page_state(NR_ACTIVE_FILE) +
               global_node_page_state(NR_UNEVICTABLE)) *
              PAGE_SIZE) /
             (1024 * 1024);

    *libre = (global_zone_page_state(NR_FREE_PAGES) * PAGE_SIZE) / (1024 * 1024);

    *uso = *total - *libre;
}

// Función para obtener el ID del contenedor desde el cgroup del proceso
static const char *get_container_id(struct task_struct *task)
{
    struct cgroup *cgrp = task->cgroups->dfl_cgrp;
    if (cgrp && cgrp->kn)
    {

        const char *cgroup_name = cgrp->kn->name;
        if (strstr(cgroup_name, "docker"))
        {

            const char *prefix = "docker-";
            if (strncmp(cgroup_name, prefix, strlen(prefix)) == 0)
            {
                cgroup_name += strlen(prefix);
            }

            const char *suffix = ".scope";
            if (strstr(cgroup_name, suffix))
            {
                char *id_container = kstrdup(cgroup_name, GFP_KERNEL);
                if (id_container)
                {
                    char *suffix_pos = strstr(id_container, suffix);
                    if (suffix_pos)
                    {
                        *suffix_pos = '\0';
                    }
                    return id_container;
                }
            }
        }
    }
    return NULL; // No es un contenedor Docker
}

// Funcion para leer un archivo
static int read_file(const char *path, size_t size, char *buffer)
{
    struct file *file;
    char *kbuf;
    int ret = 0;

    file = filp_open(path, O_RDONLY, 0);
    if (IS_ERR(file))
    {
        printk(KERN_ERR "Error: no se ha podido abrir el archivo %s\n", path);
        return PTR_ERR(file);
    }

    kbuf = kmalloc(size, GFP_KERNEL);
    if (!kbuf)
    {
        ret = -ENOMEM;
        goto out;
    }

    ret = kernel_read(file, kbuf, size - 1, &file->f_pos);
    if (ret > 0)
    {
        kbuf[ret] = '\0';
        strncpy(buffer, kbuf, size);
    }
    else
    {
        printk(KERN_ERR "Error: no se ha podido abrir el archivo %s\n", path);
    }

    kfree(kbuf);
out:
    filp_close(file, NULL);
    return ret;
}

// Función para obtener el uso de CPU de un contenedor
static unsigned long long get_cpu_usage(const char *id_container)
{
    char path[256];
    char buffer[256];
    unsigned long long cpu_usage1 = 0, cpu_usage2 = 0;
    unsigned long long elapsed_time = 250000; // 250 ms

    // Primera lectura de usage_usec
    snprintf(path, sizeof(path), "/sys/fs/cgroup/system.slice/docker-%s.scope/cpu.stat", id_container);
    if (read_file(path, sizeof(buffer), buffer) > 0)
    {
        char *line = strstr(buffer, "usage_usec");
        if (line)
        {
            if (sscanf(line, "usage_usec %llu", &cpu_usage1) != 1)
            {
                printk(KERN_ERR "Error: problemas al convertir el uso de la CPU\n");
            }
        }
    }

    // sleep 250ms
    msleep(250);

    // Segunda lectura de usage_usec
    if (read_file(path, sizeof(buffer), buffer) > 0)
    {
        char *line = strstr(buffer, "usage_usec");
        if (line)
        {
            if (sscanf(line, "usage_usec %llu", &cpu_usage2) != 1)
            {
                printk(KERN_ERR "Error: problemas al convertir el uso de la CPU\n");
            }
        }
    }

    // difrenciar los dos valores de uso de CPU
    unsigned long long cpu_delta = cpu_usage2 - cpu_usage1;
    // porcentaje de uso de CPU
    unsigned long long cpu_percent = (cpu_delta * 100) / elapsed_time; //<- 250ms
    return cpu_percent;
}

// Función para obtener el uso de memoria de un contenedor
static unsigned long get_memory_usage(const char *id_container)
{
    char path[256];
    char buffer[64];
    unsigned long memory_use = 0;

    snprintf(path, sizeof(path), "/sys/fs/cgroup/system.slice/docker-%s.scope/memory.current", id_container);

    if (read_file(path, sizeof(buffer), buffer) > 0)
    {
        if (kstrtoul(buffer, 10, &memory_use) != 0)
        {
            printk(KERN_ERR "Error: problemas al convertir el uso de la memoria RAM\n");
        }
    }
    return memory_use / (1024 * 1024); // Convertir a MiB
}

// Función para obtener el uso de I/O de un contenedor
static unsigned long long get_io_usage(const char *container_id)
{
    char path[256], buffer[256];
    unsigned long long write_ops = 0;

    snprintf(path, sizeof(path), "/sys/fs/cgroup/system.slice/docker-%s.scope/io.stat", container_id);

    // Leer el archivo io.stat y extraer los wios
    if (read_file(path, sizeof(buffer), buffer) > 0)
    {
        char *wios_pos = strstr(buffer, "wios=");
        if (wios_pos)
        {
            wios_pos += strlen("wios=");
            if (sscanf(wios_pos, "%llu", &write_ops) != 1)
            {
                printk(KERN_ERR "Error: problemas al convertir wios\n");
            }
        }
    }

    return write_ops;
}

// Uso del disco de un contenedor
static unsigned long get_disk_usage(const char *container_id)
{
    char path[256], buffer[256];
    unsigned long long rbytes = 0, wbytes = 0;

    snprintf(path, sizeof(path), "/sys/fs/cgroup/system.slice/docker-%s.scope/io.stat", container_id);

    // Leer rbytes
    if (read_file(path, sizeof(buffer), buffer) > 0)
    {
        char *rbytes_pos = strstr(buffer, "rbytes=");
        if (rbytes_pos)
        {
            rbytes_pos += strlen("rbytes=");
            if (sscanf(rbytes_pos, "%llu", &rbytes) != 1)
            {
                printk(KERN_ERR "Error: problemas al convertir rbytes\n");
            }
        }

        // Leer wbytes
        char *wbytes_pos = strstr(buffer, "wbytes=");
        if (wbytes_pos)
        {
            wbytes_pos += strlen("wbytes=");
            if (sscanf(wbytes_pos, "%llu", &wbytes) != 1)
            {
                printk(KERN_ERR "Error: problemas al convertir wbytes\n");
            }
        }
    }
    // Calcular el total de bytes con los valores leidos
    unsigned long long total_bytes = rbytes + wbytes;
    return total_bytes / (1024 * 1024); // <- rotornar en MiB
}

static char *get_command_from_pid(pid_t pid)
{
    struct task_struct *task;
    char *cmd;

    task = get_pid_task(find_get_pid(pid), PIDTYPE_PID);
    if (!task)
        return "N/A"; // Si el proceso no existe, devolver "N/A"

    cmd = kstrdup(task->comm, GFP_KERNEL);

    put_task_struct(task); // Liberar la referencia al task_struct
    return cmd;
}

// Función para escribir en `/proc/sysinfo_202100119`
static int write_file(struct seq_file *archivo, void *v)
{
    unsigned long total_mem, free_mem, used_mem;
    struct task_struct *task;
    struct containers_struct_info *containers;
    int container_count = 0;

    // Obtener datos de memoria RAM general
    get_memory_usage_general(&total_mem, &free_mem, &used_mem);

    seq_printf(archivo, "{\n");
    seq_printf(archivo, "\t\"cpu_general\": %d,\n", cpu_general_use);
    seq_printf(archivo, "\t\"ram_memory\": {\n");
    seq_printf(archivo, "\t\t\"total\": %lu,\n", total_mem);
    seq_printf(archivo, "\t\t\"free\": %lu,\n", free_mem);
    seq_printf(archivo, "\t\t\"used\": %lu\n", used_mem);
    seq_printf(archivo, "\t},\n");

    // Reservar memoria para almacenar la información de los contenedores
    containers = kmalloc_array(MAX_CONTAINERS, sizeof(struct containers_struct_info), GFP_KERNEL);
    if (!containers)
    {
        printk(KERN_ERR "Error: problema con la asignacion de memoria\n");
        seq_printf(archivo, "\t\"Containers\": []\n");
        seq_printf(archivo, "}\n");
        return -ENOMEM;
    }

    seq_printf(archivo, "\t\"containers\": [\n");

    int first_container = 1; // Variable para controlar el primer contenedor

    for_each_process(task)
    {
        const char *id_container = get_container_id(task);
        if (id_container && strcmp(id_container, "docker.service") != 0)
        {
            int is_duplicate = 0;
            for (int i = 0; i < container_count; i++)
            {
                if (strncmp(containers[i].id, id_container, ID_CONTAINER_SIZE) == 0)
                {
                    is_duplicate = 1;
                    break;
                }
            }

            if (!is_duplicate && container_count < MAX_CONTAINERS)
            {
                struct containers_struct_info *container = &containers[container_count];

                // Copiar ID del contenedor
                strncpy(container->id, id_container, ID_CONTAINER_SIZE);
                container->id[ID_CONTAINER_SIZE] = '\0';

                snprintf(container->name, sizeof(container->name), "container_%s", container->id);
                container->pid = task->pid;

                // Obtener estadísticas
                char *command = get_command_from_pid(task->pid);
                container->cpu_use = get_cpu_usage(id_container);
                container->memory_use = get_memory_usage(id_container);
                container->io_use = get_io_usage(id_container);
                container->disk_use = get_disk_usage(id_container);

                // Copiar el comando
                if (command)
                {
                    strncpy(container->command, command, sizeof(container->command) - 1);
                    container->command[sizeof(container->command) - 1] = '\0';
                    kfree(command); // Liberar memoria después de copiar
                }
                else
                {
                    strncpy(container->command, "N/A", sizeof(container->command) - 1);
                    container->command[sizeof(container->command) - 1] = '\0';
                }

                // Escribir información del contenedor en formato JSON
                if (!first_container)
                {
                    seq_printf(archivo, ",\n"); // Coloca una coma antes del contenedor si no es el primero
                }
                first_container = 0; // Ya no es el primer contenedor

                seq_printf(archivo, "\t\t{\n");
                // seq_printf(archivo, "\t\t\t\"name\": \"%s\",\n", container->name);
                seq_printf(archivo, "\t\t\t\"id\": \"%s\",\n", container->id);
                seq_printf(archivo, "\t\t\t\"pid\": %d,\n", container->pid);
                seq_printf(archivo, "\t\t\t\"command\": \"%s\",\n", container->command);
                seq_printf(archivo, "\t\t\t\"cpu_use\": %llu,\n", container->cpu_use);
                seq_printf(archivo, "\t\t\t\"ram_use\": %lu,\n", container->memory_use);
                seq_printf(archivo, "\t\t\t\"io_use\": %llu,\n", container->io_use);
                seq_printf(archivo, "\t\t\t\"disk_use\": %llu\n", container->disk_use);
                seq_printf(archivo, "\t\t}");

                container_count++;
            }
        }
    }

    seq_printf(archivo, "\n\t]\n");
    seq_printf(archivo, "}\n");

    // Liberar memoria asignada
    kfree(containers);

    return 0;
}

static int open_file(struct inode *inode, struct file *file)
{
    return single_open(file, write_file, NULL);
}

static const struct proc_ops sysinfo_ops = {
    .proc_open = open_file,
    .proc_read = seq_read,
    .proc_lseek = seq_lseek,
    .proc_release = single_release,
};

static int _insert(void)
{
    proc_create("sysinfo_202100119", 0, NULL, &sysinfo_ops);
    cpu_queue = create_workqueue("cpu_queue");
    INIT_DELAYED_WORK(&cpu_work, update_cpu_usage);
    queue_delayed_work(cpu_queue, &cpu_work, 0);
    printk(KERN_INFO "Se insertó el módulo sysinfo_202100119 con monitoreo en tiempo real\n");
    return 0;
}

static void _remove(void)
{
    remove_proc_entry("sysinfo_202100119", NULL);
    cancel_delayed_work_sync(&cpu_work);
    destroy_workqueue(cpu_queue);
    printk(KERN_INFO "Módulo sysinfo_202100119 removido correctamente\n");
}

module_init(_insert);
module_exit(_remove);
