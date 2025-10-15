# IoTS_Projekat3

### DataManager
  * Servis zadužen za upravljanje podacima.

### Gateway
  * Ovaj servis predstavlja ulaznu tačku u sistem, on prima podatke sa senzora i prosledjuje ih dalje.

### SensorGenerator
  * Služi za simuliranje rada senzora. Čita podatke iz fajlova i prosledjuje ih na Gateway.

### EventManager
  * Prima podatke sa DataManager-a i prosledjuje ih na MqttClient servis ukoliko detektuje podatke sa prevelikim vrednostima.

### MqttClient
  * Prima podatke sa EventManager-a i prikazuje ih.

### Analytics
  * Prima podatke sa DataManager-a i analizira ih pomoću MLaaS servisa. Dobijene rezultate šalje MqttClient-u preko NATS broker-a.
### MLaaS
  * Vrši klasifikaciju i regresiju senzorskih podataka.
