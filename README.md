# Introduction

This project implements a distributed system simulation for a fridge manufactorer.

<ins>Developers:</ins>

- Dennis Appel
- Marius Ebert
- Jens Niederreiter

<ins>Architecture:</ins> **Microservices**

<ins>Technologies: </ins>

- Programming language: **Go**
- Database: **MongoDB**
- API: **REST**
- Message broker: **RabbitMQ**
- Container: **Docker**





Language: German

School: University of Applied Sciences Mittelhessen, Gießen, Germany (*Technische Hochschule Mittelhessen*)



____

# EFridge

**WICHTIG: Zwar ist die Architektur komplett und die services implementiert, jeodch verhindern einige Bugs ein ausführen der Gesamtanwendung. Fixes, die für die Demo nötig sind, werden gegebenenfalls in einem unabhaengigen Branch gemacht und in einem entsprechenden Merge Request aufgefuehrt um diese klar von der Abgabe zur Deadline zur differenzieren.**

## Architektur
Bei unserer Lösung handelt es sich um eine Microservice Architektur. Kern der Architektur ist eine Service Library, die eine einfache Wiederverwendung der Basiskomponenten ermöglicht. Die Services werden in Form von Docker Images deployt. Die gesamte Anwendung kann mittels einer einzelnen [docker-compose](docker-compose.yml) Datei hochgefahren werden. Eine Besonderheit ist, dass es nur eine einzelne Binary gibt, die in der Lage ist jeden Service zu starten. Eigentlich werden in einer solchen Architektur die Services in einzelne Binaries unterteilt. Wir haben uns für diesen Weg entschieden um das Projekt übersichtlicher zu halten.

Eine genauere Beschreibung ist der [Architekturübersicht](Architekturübersicht.png)Architekturübersicht.png zu entnehmen.

## Starten
Die gesamte Anwendung kann mit Docker in den folgenden Schritten gestartet werden:
* Bauen des Service Images `docker-compose build`
* Ausführen `docker-compose up`

## Usage
Wie bereits oben erwähnt funktionieren die Services nicht einwandfrei, weshalb ein kompletter durchlauf nicht funktioniert. Dennoch wird folgend die theoretische Nutzung der Anwendung beschrieben:

### Customer
Bevor eine Order erstellt werden kann muss ein Kunde angelegt werden. Dies kann z.B. mit folgendem curl request gemacht werden:
```
curl --location --request POST '127.0.0.1:8080' \
--header 'Content-Type: application/json' \
--data-raw '{
	"firstName": "Max",
    "lastName": "Mustermann",
    "address": {
    	"country": "Germany",
    	"city": "Musterstadt",
    	"zip": "123456",
    	"address": "Musterstr 1"
    }
}'
```

### Order
Eine Order benötigt eine valide Kunden ID. Diese wird vom Customer Service nach dem erstellen eines neuen Kunden zurückgeliefert. Anschließend kann wie folgt eine Order erstelt werden:
```
curl --location --request POST '127.0.0.1:8081' \
--header 'Content-Type: application/json' \
--data-raw '{
	"customer": "5f05c865368b37098bd87aea",
    "items": [
    	"123",
    	"123"
    ]
}'
```
Hier müssten die Item IDs dem Model Service entnommen werden. Dies funktioniert zu diesem Zeitpunkt leider nicht.

### Teile Updates
Teile updates können wie folgt durchgeführt werden:
```
curl --location --request POST '127.0.0.1:8082' \
--header 'Content-Type: application/json' \
--data-raw '{
	"id": "partID",
	"price": 42
}'
```

Model informationen können wie folgt erfragt werden:
```
curl --location --request GET '127.0.0.1:8082'
```
Auch hier ist mit fehlferhalten zu rechnen.

### KPI
KPI können auf drei Arten erfragt werden: (1) die neusten Einträge, (2) der neuste Eintrag einer bestimmten Fabrik und (3) die letzten n Einträge einer bestimmten Fabrik.
```
curl --location --request GET '127.0.0.1:8083'

curl --location --request GET '127.0.0.1:8083/usa'

curl --location --request GET '127.0.0.1:8083/china/2'
```

Auch hier ist fehlferhalten zu erwarten.

### Ticket
Die ticket id wird vom post request zurück gegeben
```
curl --location --request POST '127.0.0.1:8084' \
--header 'Content-Type: application/json' \
--data-raw '{
	"text": "Ihre Frage"
}'

curl --location --request GET '127.0.0.1:8084/<ticketid>'
```
Hier ist ebenfalls mit fehlferhalten zu rechnen.