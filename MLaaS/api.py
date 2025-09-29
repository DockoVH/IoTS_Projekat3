import joblib
import numpy as np
from tensorflow.keras.models import load_model
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from model import napravi_model

def ucitaj_model(putanja):
	model, scaler = None, None

	try:
		model = load_model(f'{putanja}/senzor_podaci_model.keras')
	except Exception as ex:
		print(f'Greška prilikom učitavanja modela: {ex}')

	try:
		scaler = joblib.load(f'{putanja}/scaler.pkl')
	except Exception as ex:
		print(f'Greška prilikom učitavanja scaler-a: {ex}')

	return model, scaler


model, scaler = ucitaj_model('./Models')

if model is None or scaler is None:
	print('Model ne postoji. Pravljenje novog modela...')
	napravi_model()
	model, scaler = ucitaj_model('./Models')

if model is None or scaler is None:
	exit(1)

app = FastAPI(title = 'Senzor podaci predikcija API')
KLASE_KVALITETA_VAZDUHA = ['DOBRO', 'UMERENO', 'LOŠE', 'VEOMA LOŠE']

class SenzorPodaci(BaseModel):
	Podaci: list

@app.post('/predict')
async def predict(ulaz: SenzorPodaci):
	print(ulaz)
	try:
		podaci = np.array(ulaz.Podaci)

		if len(podaci.shape) != 2 or podaci.shape[1] != 4:
			raise HTTPException(status_code = 400, detail = 'Očekuje se matrica podataka oblika (10, 4)')

		podaci_scaled = scaler.transform(podaci)
		X_new = podaci_scaled.reshape(1, podaci_scaled.shape[0], podaci_scaled.shape[1])

		pred_temp, pred_class_probs = model.predict(X_new)
		pred_temp = float(pred_temp[0][0])
		pred_class_idx = int(np.argmax(pred_class_probs, axis = 1)[0])
		pred_kvalitet = KLASE_KVALITETA_VAZDUHA[pred_class_idx]

		return {
			'predvidjena_temperatura': pred_temp,
			'kvalitet_vazduha': {
				'klasa': pred_kvalitet,
				'verovatnoce': pred_class_probs[0].tolist()
			}
		}
	except Exception as ex:
		if ex is HTTPException:
			raise ex
		else:
			raise HTTPException(status_code = 400, detail = str(ex))
