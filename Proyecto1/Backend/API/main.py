from fastapi import FastAPI
import json

app = FastAPI()


@app.get("/")
def read_root():
    return {"Hello": "World"}

@app.get("/json_file")
def read_json():

    # leer el archivo json
    with open('stopped_containers.json', 'r') as file:
        data = json.load(file)

    return data