use serde::Deserialize;
use serde_json;
use std::fs;
use std::process::Command;

#[derive(Deserialize)]
struct RamMemory {
    total: u32,
    free: u32,
    used: u32,
}

#[derive(Deserialize)]
struct ContainersInfo {
    id: String,
    pid: u32,
    command: String,
    cpu_use: u32,
    ram_use: u32,
    io_use: u64,
    disk_use: u64
}

#[derive(Deserialize)]
struct SystemInfo {
    cpu_general: u32,
    ram_memory: RamMemory,
    containers: Vec<ContainersInfo>,
}

fn stop_container(container_id: &str) {
    let output = Command::new("docker")
        .arg("stop")
        .arg(container_id)
        .output()
        .expect("failed to execute docker stop");

    println!("status: {}", output.status);
    println!("stdout: {}", String::from_utf8_lossy(&output.stdout));
    println!("stderr: {}", String::from_utf8_lossy(&output.stderr));
}

fn get_container_name(container_id: &str) -> Result<String, Box<dyn std::error::Error>> {
    let output = Command::new("docker")
        .arg("inspect")
        .arg("--format={{join .Config.Cmd \" \"}}")
        .arg(container_id)
        .output()
        .expect("failed to execute docker inspect");

    if output.status.success() {
        let name = String::from_utf8_lossy(&output.stdout);
        Ok(name.trim().to_string())
    } else {
        Err(format!("Failed to get container name: {}", String::from_utf8_lossy(&output.stderr)).into())
    }
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
            println!("ID: {}", container.id);
            println!("PID: {}", container.pid);
            println!("Command: {}", container.command);
            println!("CPU usage: {}", container.cpu_use);
            println!("RAM usage: {}", container.ram_use);
            println!("I/O usage: {}", container.io_use);
            println!("Disk usage: {}", container.disk_use);
            match get_container_name(&container.id) {
                Ok(name) => println!("Name: {}", name),
                Err(e) => println!("Error getting name: {}", e),
            }
            println!("----------------------------");
        }
    }

    Ok(())
}


