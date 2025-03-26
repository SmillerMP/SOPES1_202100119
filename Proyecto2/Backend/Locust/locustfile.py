from locust import HttpUser, task
import json

class WeatherUser(HttpUser):
    @task
    def send_weather_data(self):

        # deserializar la data

        with open ('./data/weather_data_10000.json', 'r') as file:
            data = json.load(file)

        # enviar la data
        self.client.post("/weather", json=data)