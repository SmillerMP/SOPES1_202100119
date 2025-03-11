use serde::Deserialize;
use serde_json;
use std::collections::HashMap;
use std::fs;
use std::process::Command;
use std::thread;
use std::time::Duration;

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
    disk_use: u64,
}

#[derive(Deserialize)]
struct SystemInfo {
    cpu_general: u32,
    ram_memory: RamMemory,
    containers: Vec<ContainersInfo>,
}

fn stop_container(container_id: String) {
    thread::spawn(move || {
        let output = Command::new("docker")
            .arg("stop")
            .arg(&container_id)
            .output()
            .expect("failed to execute docker stop");

        println!("🛑 Stopped container: {}", container_id);
        // println!("status: {}", output.status);
        // println!("stdout: {}", String::from_utf8_lossy(&output.stdout));
        // println!("stderr: {}", String::from_utf8_lossy(&output.stderr));
    });
}

fn get_container_command(container_id: &str) -> Result<String, Box<dyn std::error::Error>> {
    let output = Command::new("docker")
        .arg("inspect")
        .arg("--format={{join .Config.Cmd \" \"}}")
        .arg(container_id)
        .output()
        .expect("failed to execute docker inspect");

    if output.status.success() {
        let command = String::from_utf8_lossy(&output.stdout);
        Ok(command.trim().to_string())
    } else {
        Err(format!(
            "Failed to get container command: {}",
            String::from_utf8_lossy(&output.stderr)
        )
        .into())
    }
}

fn monitor_containers() {
    let file_proc_path = "/proc/sysinfo_202100119";

    loop {
        let mut stress_containers: HashMap<&str, String> = HashMap::new();
        let file_json_content = fs::read_to_string(file_proc_path)
            .expect("Something went wrong reading the file");

        let mut system_info: SystemInfo = serde_json::from_str(&file_json_content)
            .expect("Failed to parse JSON");

        println!("🔍 Analizando contenedores en ejecución...\n");

        if system_info.containers.is_empty() {
            println!("⚠️ No hay contenedores en ejecución.");
        } else {
            for container in &mut system_info.containers {
                // println!("🔹 ID: {}", container.id);

                match get_container_command(&container.id) {
                    Ok(command) => {
                        container.command = command.clone();
                        // println!("🔸 Command: {}", command);
                    }
                    Err(e) => println!("⚠️ Error obteniendo comando: {}", e),
                }

                let stress_types = ["--cpu", "--hdd", "--io", "--vm"];
                let mut should_stop = false;
                let mut container_type: Option<&str> = None;

                if container.command.contains("stress") {
                    for &stype in &stress_types {
                        if container.command.contains(stype) {
                            container_type = Some(stype);
                            break;
                        }
                    }

                    if let Some(stype) = container_type {
                        if stress_containers.contains_key(stype) {
                            println!(
                                "⚠️ Ya hay un contenedor registrado para '{}', deteniendo {}...",
                                stype, container.id
                            );
                            should_stop = true;
                        } else {
                            stress_containers.insert(stype, container.id.clone());
                            println!("✅ Registrado contenedor {} para '{}'", container.id, stype);
                        }
                    } else {
                        println!(
                            "❌ Contenedor {} ejecuta 'stress' sin un tipo reconocido, deteniéndolo...",
                            container.id
                        );
                        should_stop = true;
                    }
                }

                if should_stop {
                    stop_container(container.id.clone());
                }
            }
        }

        // Limpiar el `HashMap` para la siguiente iteración
        println!("🔄 Limpiando registros para la siguiente iteración...\n");

        // Esperar 10 segundos antes de la siguiente verificación
        thread::sleep(Duration::from_secs(20));
    }
}

fn main() {
    monitor_containers();
}
