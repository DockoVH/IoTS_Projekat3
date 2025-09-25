import pandas as pd
import numpy as np
import tensorflow as tf
import joblib
from sklearn.preprocessing import MinMaxScaler
from sklearn.model_selection import train_test_split
from tensorflow.keras import Model, Input
from tensorflow.keras.utils import to_categorical
from tensorflow.keras.layers import LSTM, Dense, Dropout

import time

def ucitaj_i_preprocesiraj(csv_putanja, scaler):
	print('Učitavanje i preprocesiranje podataka...')

	try:
		df = pd.read_csv(csv_putanja)
	except Exception as ex:
		print(f'Greška prilikom učitavanja fajla {csv_putanja}: {ex}')
		return None, None, None

	potrebne_kolone = ['id', 'vreme', 'temperatura', 'vlaznost_vazduha', 'pm2_5', 'pm10']
	nema_kolone = [col for col in potrebne_kolone if col not in df.columns]
	if nema_kolone:
		print(f'Neodstaju kolone: {nema_kolone}.')
		return None, None, None

	df = df[df.columns.intersection(potrebne_kolone)]

	df['vreme'] = pd.to_datetime(df['vreme'])
	df = df.sort_values('vreme')

	lenPocetak = len(df)
	df = df.drop_duplicates()
	print(f'Uklonjeno {lenPocetak - len(df)} duplikata.')
	lenSredina = len(df)
	df = df.dropna(subset=['temperatura', 'vlaznost_vazduha', 'pm2_5', 'pm10'])
	print(f'Uklonjeno {lenSredina - len(df)} redova sa nevalidnim vrednostima.')

	def get_kvalitet_vazduha(red):
		if red['pm2_5'] <= 11 and red['pm10'] <= 54:
			return 0 # DOBRO
		elif red['pm2_5'] <= 35 and red['pm10'] <= 154:
			return 1 # UMERENO
		elif red['pm2_5'] <= 55 and red['pm10'] <= 253:
			return 2 # LOŠE
		else:
			return 3 # VEOMA LOŠE

	df['kvalitet_vazduha'] = df.apply(get_kvalitet_vazduha, axis = 1)

	features = ['temperatura', 'vlaznost_vazduha', 'pm2_5', 'pm10']
	podaci = df[features].values
	labels_class = df['kvalitet_vazduha'].values
	labels_reg = df['temperatura'].values

	podaci_scaled = scaler.fit_transform(podaci)

	return podaci_scaled, labels_class, labels_reg

def treniraj(podaci, labels_class, labels_reg, lookback = 10, test_velicina = 0.2):
	def napravi_dataset(podaci, labels_class, labels_reg, lookback):
		X, y_class, y_reg = [], [], []
		for i in range(lookback, len(podaci)):
			X.append(podaci[i - lookback:i])
			y_class.append(labels_class[i])
			y_reg.append(labels_reg[i])

		return np.array(X), np.array(y_class), np.array(y_reg)

	X, y_class, y_reg = napravi_dataset(podaci, labels_class, labels_reg, lookback)

	X_train, X_test, y_class_train, y_class_test, y_reg_train, y_reg_test = train_test_split(X, y_class, y_reg, test_size = test_velicina, random_state = 42)

	#Za klasifikaciju ???
	y_class_train_cat = to_categorical(y_class_train, num_classes = 4)
	y_class_test_cat = to_categorical(y_class_test, num_classes = 4)

	ulaz = Input(shape = (lookback, X.shape[2]))
	x = LSTM(64, activation = 'tanh')(ulaz)
	x = Dropout(0.2)(x)
	zajednicki = Dense(32, activation = 'relu')(x)

	izlaz_reg = Dense(1, name='reg_izlaz')(zajednicki)
	izlaz_class = Dense(4, activation = 'softmax', name='class_izlaz')(zajednicki)

	model = Model(inputs = ulaz, outputs = [izlaz_reg, izlaz_class])

	model.compile(
		optimizer = 'adam',
		loss = {
			'reg_izlaz': 'mse',
			'class_izlaz': 'categorical_crossentropy'
		},
		metrics = {
			'reg_izlaz': ['mae'],
			'class_izlaz': ['accuracy']
		}
	)

	model.summary()

	model.fit(
		X_train,
		{
			'reg_izlaz': y_reg_train,
			'class_izlaz': y_class_train_cat
		},
		validation_data = (
			X_test,
			{
				'reg_izlaz': y_reg_test,
				'class_izlaz': y_class_test_cat
			}
		),
		epochs = 5,
		batch_size = 32
	)

	return model

def sacuvaj(model, scaler):
	try:
		model.save('Models/senzor_podaci_model.keras')
		joblib.dump(scaler, 'Models/scaler.pkl')
		return True
	except Exception as ex:
		print(f'Greška prilikom čuvanja modela {ex}')
	return False

def napravi_model():
	scaler = MinMaxScaler()

	podaci_scaled, labels_class, labels_reg = ucitaj_i_preprocesiraj('Datasetovi/18.csv', scaler)

	if podaci_scaled is None or labels_class is None or labels_reg is None:
		print('Model nije napravljen')
		return False

	model = treniraj(podaci_scaled, labels_class, labels_reg)
	return sacuvaj(model, scaler)


if __name__ == "__main__":
	napravi_model()
	print('Model uspešno napravljen.')
