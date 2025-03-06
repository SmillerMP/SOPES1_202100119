use serde::Deserialize;
use serde_json;
use std::fs;

#[derive(Deserialize)]
struct RamMemory {
    total: u32,
    free: u32,
    used: u32,
}

#[derive(Deserialize)]
struct ContainersInfo {
    name: String,
    pid: String,
    command: String,
    cpu_use: String,
    ram_use: String,
    io_use: String,
    disk_use: String
}

#[derive(Deserialize)]
struct SystemInfo {
    cpu_general: u32,
    ram_memory: RamMemory,
    containers: Vec<ContainersInfo>,
}

fn kill_container(pid: u32) {
    // Implementar
}



fn main() -> Result<(), Box<dyn std::error::Error>> {
    let file_proc_path = "/proc/sysinfo_202100119";

    let file_json_content =
        fs::read_to_string(file_proc_path).expect("Something went wrong reading the file");

    let system_info: SystemInfo = serde_json::from_str(&file_json_content)?;

    // Aquí usa los nombres en snake_case
    println!("CPU: {}", system_info.cpu_general);
    println!("RAM Total: {}", system_info.ram_memory.total);
    println!("RAM Free: {}", system_info.ram_memory.free);
    println!("RAM Used: {}", system_info.ram_memory.used);

    // Mostrar el arreglo de contenedores (aunque esté vacío por ahora)
    if system_info.containers.is_empty() {
        println!("No containers.");
    } else {
        println!("\n\nContainers:");
        for container in &system_info.containers {
            println!("Name: {}", container.name);
            println!("PID: {}", container.pid);
            println!("Command: {}", container.command);
            println!("CPU usage: {}", container.cpu_use);
            println!("RAM usage: {}", container.ram_use);
            println!("I/O usage: {}", container.io_use);
            println!("Disk usage: {}", container.disk_use);
            println!("----------------------------");
        }
    }

    Ok(())
}
