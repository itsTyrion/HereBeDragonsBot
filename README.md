# Here Be Dragons - Discord Bot

Ein einfacher Discord-Bot, der mit Go 1.25 und der [disgo](https://github.com/disgoorg/disgo) Bibliothek erstellt wurde.

## Funktionen
Nicht viele – entwickelt für einen privaten Gilden-Discord.
- Verfolgt Zahlen in Nachrichten für ein Hochzähl-Minispiel
- Purge Befehl zum Aufräumen (löschen) von Nachrichten


## Installation
1. Repository klonen
2. Bot bauen und starten (Optional mit Docker/Compose):
   ```bash
   go build
   ./HereBeDragons
   ```
3. Konfiguration anpassen (siehe unten)

## Docker
Der Bot kann auch mit Docker ausgeführt werden:

```bash
docker build -t here-be-dragons .
docker run -v $(pwd)/data:/app/data -e WORK_DIR=/app/data here-be-dragons
```

### Docker Compose
```bash
docker compose up -d --build
```

## Konfiguration
Beim ersten Start des Bots wird `config.json` im Arbeitsverzeichnis erstellt.
Das Arbeitsverzeichnis kann mit der Umgebungsvariable `WORK_DIR` angepasst werden.
Dort muss der Bot Token angegeben werden, welcher im [Discord Dev Portal](https://discord.com/developers/applications) erstellt werden muss.

Der Bot erstellt automatisch eine `state.json` Datei, um den aktuellen Zustand zu speichern.

## Lizenz
Dieses Projekt steht unter der MIT Lizenz.
