use actix_web::{middleware::Logger, web, App, HttpServer, Responder, HttpResponse};
use serde::{Deserialize, Serialize};
use dotenv::dotenv;
use env_logger;

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

// El endpoint GET  /
async fn hello() -> impl Responder {
    let greeting = Greeting {
        message: "Hello, World!".to_string()
    };

    HttpResponse::Ok().json(greeting) 
}

// Endpoint POST para recibir un arreglo de objetos JSON
async fn receive_weather(data: web::Json<Vec<WeatherInfo>>) -> impl Responder {
    // for weather_info in data.iter() {
    //     println!("Country: {}, Weather: {}, Description: {}", 
    //         weather_info.country, 
    //         weather_info.weather, 
    //         weather_info.description);
    // }
    HttpResponse::Ok().json(data.into_inner()) 
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    dotenv().ok();  // cargar las variables de entorno
    env_logger::init(); // middleware para el logeo de las peticiones

    HttpServer::new(|| {
        App::new()
            .wrap(Logger::default())
            .route("/", web::get().to(hello)) // Ruta GET
            .route("/weather", web::post().to(receive_weather)) // Ruta POST
    })
    .bind("0.0.0.0:8000")?
    .run()
    .await
}
