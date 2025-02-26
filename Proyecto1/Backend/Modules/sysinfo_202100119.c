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

MODULE_LICENSE("GPL");
MODULE_DESCRIPTION("Monitor en tiempo real de CPU y RAM");
MODULE_AUTHOR("Samuel Isaí Muñoz Pereira");

static struct delayed_work cpu_work;
static struct workqueue_struct *cpu_queue;
static int cpu_usage = 0;  // Variable global para el uso del CPU

// Variables para almacenar la última lectura de CPU
static int prev_total = 0, prev_idle = 0;

// Función para leer `/proc/stat` y calcular el uso del CPU
static void update_cpu_usage(struct work_struct *work) {
    struct file *archivo;
    char lectura[256];
    loff_t pos = 0;
    int usuario, nice, system, idle, iowait, irq, softirq, steal, guest, guest_nice;
    int total, diff_idle, diff_total;
    ssize_t bytes_leidos;

    archivo = filp_open("/proc/stat", O_RDONLY, 0);
    if (IS_ERR(archivo)) {
        printk(KERN_ERR "Error al abrir /proc/stat\n");
        return;
    }

    memset(lectura, 0, sizeof(lectura));
    bytes_leidos = kernel_read(archivo, lectura, sizeof(lectura) - 1, &pos);
    filp_close(archivo, NULL);

    if (bytes_leidos <= 0) {
        printk(KERN_ERR "Error al leer /proc/stat\n");
        return;
    }

    lectura[bytes_leidos] = '\0';

    // Extraer valores de CPU
    if (sscanf(lectura, "cpu %d %d %d %d %d %d %d %d %d %d",
               &usuario, &nice, &system, &idle, &iowait, &irq, &softirq, &steal, &guest, &guest_nice) != 10) {
        printk(KERN_ERR "Error al parsear /proc/stat\n");
        return;
    }

    total = usuario + nice + system + idle + iowait + irq + softirq + steal + guest + guest_nice;
    diff_total = total - prev_total;
    diff_idle = idle - prev_idle;

    if (diff_total > 0) {
        cpu_usage = (100 * (diff_total - diff_idle)) / diff_total;
    }

    prev_total = total;
    prev_idle = idle;

    // Programar la próxima ejecución en 1 segundo
    queue_delayed_work(cpu_queue, &cpu_work, msecs_to_jiffies(1000));
}

// Función para obtener la memoria RAM (total, libre y en uso)
static void get_memory_usage(unsigned long *total, unsigned long *libre, unsigned long *uso) {
    // Obtener memoria total del sistema en MB
    *total = ((global_node_page_state(NR_INACTIVE_ANON) +
               global_node_page_state(NR_ACTIVE_ANON) +
               global_node_page_state(NR_INACTIVE_FILE) +
               global_node_page_state(NR_ACTIVE_FILE) +
               global_node_page_state(NR_UNEVICTABLE)) * PAGE_SIZE) / (1024 * 1024);

    *libre = (global_zone_page_state(NR_FREE_PAGES) * PAGE_SIZE) / (1024 * 1024);

    *uso = *total - *libre;
}

// Función para escribir en `/proc/sysinfo_202100119`
static int write_file(struct seq_file *archivo, void *v) {
    struct task_struct *task;
    unsigned long total_mem, free_mem, used_mem;

    // Obtener datos de memoria RAM
    get_memory_usage(&total_mem, &free_mem, &used_mem);

    seq_printf(archivo, "{\n");
    seq_printf(archivo, "  \"percentage_used\": %d,\n", cpu_usage);
    seq_printf(archivo, "  \"memory\": {\n");
    seq_printf(archivo, "    \"total\": %lu,\n", total_mem);
    seq_printf(archivo, "    \"free\": %lu,\n", free_mem);
    seq_printf(archivo, "    \"used\": %lu\n", used_mem);
    seq_printf(archivo, "  },\n");
    seq_printf(archivo, "  \"tasks\": [\n");

    for_each_process(task) {
        if (strcmp(task->parent->comm, "containerd-shim") == 0 || strcmp(task->parent->comm, "dockerd") == 0) {
            seq_printf(archivo, "    {\n");
            seq_printf(archivo, "      \"pid\": %d,\n", task->pid);
            seq_printf(archivo, "      \"name\": \"%s\",\n", task->comm);
            seq_printf(archivo, "      \"user\": %d,\n", task->cred->uid.val);
            seq_printf(archivo, "      \"father\": %d\n", task->parent->pid);
            seq_printf(archivo, "    },\n");
        }
    }

    seq_printf(archivo, "  ]\n}\n");
    return 0;
}

static int open_file(struct inode *inode, struct file *file) {
    return single_open(file, write_file, NULL);
}

static const struct proc_ops sysinfo_ops = {
    .proc_open = open_file,
    .proc_read = seq_read,
    .proc_lseek = seq_lseek,
    .proc_release = single_release,
};

static int _insert(void) {
    proc_create("sysinfo_202100119", 0, NULL, &sysinfo_ops);
    cpu_queue = create_workqueue("cpu_queue");
    INIT_DELAYED_WORK(&cpu_work, update_cpu_usage);
    queue_delayed_work(cpu_queue, &cpu_work, 0);
    printk(KERN_INFO "Se insertó el módulo sysinfo_202100119 con monitoreo en tiempo real\n");
    return 0;
}

static void _remove(void) {
    remove_proc_entry("sysinfo_202100119", NULL);
    cancel_delayed_work_sync(&cpu_work);
    destroy_workqueue(cpu_queue);
    printk(KERN_INFO "Módulo sysinfo_202100119 removido correctamente\n");
}

module_init(_insert);
module_exit(_remove);
