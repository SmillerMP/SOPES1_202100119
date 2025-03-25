from locust import HttpUser, task
import random

class WeatherUser(HttpUser):
    @task
    def send_weather_data(self):
        
        # Descripciones y opciones de clima alineadas
        descriptions = [
            "Está lloviendo", "Está nublado", "Hace sol", "Hay tormenta", "Clima cálido", "Clima fresco", "Hay neblina", "Vientos fuertes", "Granizo", "Clima seco", "Clima húmedo", "Temperatura agradable", "Olas de calor", "Frío extremo", "Lluvias ligeras", "Lluvias torrenciales", "Cielo despejado", "Amanecer brillante", "Atardecer cálido", "Noche estrellada"
        ]
        
        wheather_options = [
            "Lluvioso", "Nubloso", "Soleado", "Tormentoso", "Cálido", "Fresco", "Neblinoso", "Ventoso", "Con granizo", "Seco", "Húmedo", "Agradable", "Caluroso", "Frío", "Lluvias ligeras", "Lluvias torrenciales", "Despejado", "Brillante", "Cálido", "Estrellado"
        ]
        
        countries = [
            "Guatemala", "Honduras", "El Salvador", "Nicaragua", "Costa Rica", "Panamá",
            "México", "Colombia", "Argentina", "Chile", "Perú", "Ecuador", "Venezuela",
            "Bolivia", "Paraguay", "Uruguay", "Cuba", "República Dominicana", "Puerto Rico",
            "España", "Brasil", "Canadá", "Estados Unidos", "Italia", "Francia", "Alemania",
            "Reino Unido", "Japón", "China", "India", "Australia", "Sudáfrica", "Egipto",
            "Rusia", "Corea del Sur", "Filipinas", "Indonesia", "Tailandia", "Vietnam",
            "Arabia Saudita", "Turquía", "Grecia", "Portugal", "Noruega", "Suecia", "Finlandia",
            "Dinamarca", "Polonia", "Hungría", "Austria", "Suiza", "Bélgica", "Países Bajos",
            "Irlanda", "Nueva Zelanda", "Singapur", "Malasia", "Pakistán", "Bangladés",
            "Nepal", "Sri Lanka", "Myanmar", "Camboya", "Laos", "Mongolia", "Kazajistán",
            "Uzbekistán", "Turkmenistán", "Irán", "Irak", "Siria", "Líbano", "Israel",
            "Jordania", "Kenia", "Nigeria", "Ghana", "Etiopía", "Tanzania", "Zimbabue",
            "Zambia", "Botsuana", "Namibia", "Mozambique", "Angola", "Argelia", "Marruecos",
            "Túnez", "Libia", "Sudán", "Somalia", "Madagascar", "Papúa Nueva Guinea",
            "Fiyi", "Samoa", "Tonga", "Kiribati", "Micronesia", "Maldivas", "Bhután"
        ]

        index = random.randint(0, len(descriptions) - 1)        

        payload = {
            "country": random.choice(countries),
            "weather": wheather_options[index],
            "description": descriptions[index]
        }

        self.client.post("/weather", json=payload)