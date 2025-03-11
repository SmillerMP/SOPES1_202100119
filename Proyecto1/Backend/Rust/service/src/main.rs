use serde::{Deserialize, Serialize};
use serde_json;
use std::collections::HashMap;
use std::fs::{self, OpenOptions};
use std::io::{BufReader, Write};
use std::process::Command;
use std::sync::{Arc, Mutex};
use std::thread;
use std::time::Duration;
use chrono::Utc;

#[derive(Deserialize)]
#[allow(dead_code)]
struct RamMemory {
    total: u32,
    free: u32,
    used: u32,
}

#[derive(Deserialize)]
#[allow(dead_code)]
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
#[allow(dead_code)]
struct SystemInfo {
    cpu_general: u32,
    ram_memory: RamMemory,
    containers: Vec<ContainersInfo>,
}

#[derive(Serialize, Deserialize, Clone)]
#[allow(dead_code)]
struct StoppedContainerLog {
    id: String,
    command: String,
    stopped_at: String, // ISO 8601 timestamp
    cpu_use: u32,
    ram_use: u32,
    io_use: u64,
    disk_use: u64,
}

fn stop_container(
    container_id: String, 
    container_command: String, 
    cpu_use: u32, 
    ram_use: u32, 
    io_use: u64, 
    disk_use: u64, 
    logs: Arc<Mutex<()>>
) {
    let timestamp = Utc::now().to_rfc3339(); 

    thread::spawn(move || {
        let _output = Command::new("docker")
            .arg("stop")
            .arg(&container_id)
            .output()
            .expect("failed to execute docker stop");

        println!("🛑 Stopped container: {}", container_id);

        let log_entry = StoppedContainerLog {
            id: container_id.clone(),
            command: container_command,
            stopped_at: timestamp,
            cpu_use,
            ram_use,
            io_use,
            disk_use,
        };

        save_stopped_container(&log_entry, logs);
    });
}

fn save_stopped_container(log_entry: &StoppedContainerLog, logs: Arc<Mutex<()>>) {
    let log_file = "stopped_containers.json";

    let _lock = logs.lock().expect("Failed to lock logs for writing"); 

    // Leer el archivo JSON actual (si existe)
    let mut logs_vec: Vec<StoppedContainerLog> = match fs::File::open(log_file) {
        Ok(file) => {
            let reader = BufReader::new(file);
            serde_json::from_reader(reader).unwrap_or_else(|_| Vec::new())
        }
        Err(_) => Vec::new(), // Si el archivo no existe, iniciar con un array vacío
    };

    // Agregar el nuevo registro
    logs_vec.push(log_entry.clone());

    // Escribir el array actualizado en formato JSON
    let json_data = serde_json::to_string_pretty(&logs_vec).expect("Failed to serialize logs");

    let mut file = OpenOptions::new()
        .write(true)
        .create(true)
        .truncate(true) 
        .open(log_file)
        .expect("Failed to open log file");

    file.write_all(json_data.as_bytes())
        .expect("Failed to write log entry");

    println!("📁 Contenedor detenido guardado en JSON: {}", log_entry.id);
}

fn get_container_command(container_id: &str) -> String {
    let output = Command::new("docker")
        .arg("inspect")
        .arg("--format={{join .Config.Cmd \" \"}}")
        .arg(container_id)
        .output();

    match output {
        Ok(output) if output.status.success() => {
            String::from_utf8_lossy(&output.stdout).trim().to_string()
        }
        _ => {
            println!("⚠️ Error obteniendo comando del contenedor {}", container_id);
            String::from("unknown")
        }
    }
}

fn monitor_containers() {
    let file_proc_path = "/proc/sysinfo_202100119";
    let logs = Arc::new(Mutex::new(())); // Inicializar Mutex para el JSON

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
                container.command = get_container_command(&container.id);

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
                    stop_container(
                        container.id.clone(), 
                        container.command.clone(), 
                        container.cpu_use, 
                        container.ram_use, 
                        container.io_use, 
                        container.disk_use, 
                        Arc::clone(&logs)
                    );
                }
            }
        }

        println!("🔄 Limpiando registros para la siguiente iteración...\n");

        thread::sleep(Duration::from_secs(20));
    }
}

fn main() {
    monitor_containers();
}
