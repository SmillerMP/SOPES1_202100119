use actix_web::{middleware::Logger, web, App, HttpServer, Responder, HttpResponse};
use serde::{Deserialize, Serialize};
use dotenv::dotenv;
use env_logger;
use reqwest;
use std::env;

// Estructura del JSON
#[derive(Deserialize, Serialize)]
struct WeatherInfo {
    country: String,
    weather: String,
    description: String,
}

// Estructura de la respuesta JSON
#[derive(Deserialize, Serialize)]
struct Greeting {
    message: String
}

// Estructura de la respuesta a enviar al cliente
#[derive(Deserialize, Serialize)]
struct Response {
    message: String
}

// El endpoint GET  /
async fn hello() -> impl Responder {
    let greeting = Greeting {
        message: "Hello, World!".to_string()
    };

    HttpResponse::Ok().json(greeting) 
}

// Endpoint POST para recibir un arreglo de objetos JSON
async fn receive_weather(data: web::Json<Vec<WeatherInfo>>) -> impl Responder {
    // Obtener la URL de la API de Golang desde las variables de entorno
    let golang_api_url = env::var("GOLANG_API_URL").unwrap_or_else(|_| "http://api_golang:8010/weather".to_string());

    // Hacer petición a la API de Golang
    let client = reqwest::Client::new();
    let response = client
        .post(&golang_api_url)
        .json(&*data) // Enviar el arreglo de objetos JSON
        .send()
        .await;

    match response {
        Ok(res) => {
            // Si la petición fue exitosa
            if res.status().is_success() {
                HttpResponse::Ok().json(Response {
                    message: "OK API Rust".to_string(),
                })
            } else {
                // Si hubo un error con la API externa
                HttpResponse::InternalServerError().json(Response {
                    message: "Error en la API de Golang".to_string(),
                })
            }
        },
        Err(_) => {
            // Si hubo un error al realizar la petición
            HttpResponse::InternalServerError().json(Response {
                message: "Error realizando la petición".to_string(),
            })
        }
    }
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    dotenv().ok();  // cargar las variables de entorno
    env_logger::init(); // middleware para el logeo de las peticiones

    // Obtener el puerto de la API de Rust desde las variables de entorno
    let rust_api_port = env::var("API_PORT").unwrap_or_else(|_| "8000".to_string());

    HttpServer::new(|| {
        App::new()
            .wrap(Logger::default()) // Middleware de logging
            .route("/", web::get().to(hello)) // Ruta GET
            .route("/weather", web::post().to(receive_weather)) // Ruta POST
    })
    .bind(format!("0.0.0.0:{}", rust_api_port))? // Usar el puerto desde la variable de entorno
    .run()
    .await
}
