# GUIDA RAPIDA - QR Menu System 🍽️

## ✅ IL TUO PROGETTO È PRONTO!

Il sistema QR Menu è stato creato con successo nella directory `qr-menu/` ed è completamente funzionale.

## 🚀 COME UTILIZZARLO

### 1. Avvio Rapido
```bash
# Doppio click su questo file:
start.bat

# Oppure da terminale:
.\qr-menu.exe
```

### 2. Accesso Interfacce  
- **Admin Panel**: http://localhost:8080/admin
- **Crea Menu**: http://localhost:8080/admin/menu/create
- **API**: http://localhost:8080/api/menus

### 3. Test Automatico
```powershell
# Testa tutte le funzionalità:
.\test_api.ps1

# Apri tutte le interfacce:
.\open_interfaces.ps1
```

## 🍽️ FLUSSO DI UTILIZZO

1. **Crea Menu**: Vai all'admin panel → "Crea Nuovo Menu"
2. **Aggiungi Categorie**: Antipasti, Primi, Secondi, etc.
3. **Inserisci Piatti**: Nome, descrizione, prezzo per ogni piatto
4. **Completa Menu**: Clicca "Completa Menu" per generare il QR code
5. **Condividi**: Stampa/mostra il QR code ai clienti

## 📋 FUNZIONALITÀ IMPLEMENTATE

✅ **Interfaccia Web Completa**
- Dashboard amministrazione
- Form creazione menu intuitivo  
- Modifica menu esistenti
- Visualizzazione menu pubblici responsive

✅ **Generazione QR Code Automatica**
- QR code generato al completamento menu
- Accesso diretto tramite scansione
- File PNG scaricabili

✅ **API REST Completa**  
- GET /api/menus (lista menu)
- POST /api/menu (crea menu)
- GET /api/menu/{id} (dettagli menu)
- POST /api/menu/{id}/generate-qr (genera QR)

✅ **Storage Persistente** 
- Salvataggio automatico in file JSON
- Caricamento menu all'avvio
- Gestione file QR code

✅ **Design Responsive**
- Ottimizzato per mobile/tablet
- Menu pubblici eleganti  
- Interfaccia admin user-friendly

## 📁 STRUTTURA PROGETTO

```
qr-menu/
├── main.go              # Server principale
├── go.mod               # Dipendenze
├── qr-menu.exe          # Eseguibile compilato
├── start.bat            # Script avvio
├── test_api.ps1         # Test automatico
├── open_interfaces.ps1  # Apri interfacce
├── README.md            # Documentazione completa
├── models/              # Strutture dati
├── handlers/            # Logica server
├── templates/           # Template HTML
├── static/              # CSS, JS, QR codes
├── storage/             # Menu salvati (JSON)
└── examples/            # Menu di esempio
```

## 🔧 PERSONALIZZAZIONI FACILI

### Cambiare Porta
```bash
set PORT=3000
.\qr-menu.exe
```

### Aggiungere Stili
Modifica: `static/css/style.css`

### Personalizzare Template  
Modifica i file in: `templates/`

## 📱 ESEMPI D'USO

### Per Ristorante
1. Crea categorie: Antipasti, Primi, Secondi, Dolci, Bevande
2. Inserisci i tuoi piatti con prezzi
3. Genera QR code
4. Stampa e posiziona sui tavoli

### Per Pizzeria  
1. Categorie: Pizze Classiche, Pizze Speciali, Bevande
2. Dettagli ingredienti nelle descrizioni
3. QR code sul bancone/tavoli

### Per Bar
1. Categorie: Colazioni, Aperitivi, Caffetteria  
2. Orari disponibilità nelle descrizioni
3. QR code al bancone

## 🆘 RISOLUZIONE PROBLEMI

**Server non si avvia?**
- Controlla che Go sia installato: `go version`
- Verifica porta libera: cambia PORT

**Template non si caricano?** 
- Controlla che la cartella `templates/` esista
- Riavvia il server: Ctrl+C poi `.\qr-menu.exe`

**QR Code non si genera?**
- Controlla permessi cartella `static/qrcodes/`
- Verifica che il menu sia "completato"

## 🎯 PROSSIMI STEP SUGGERITI

1. **Test Real-World**: Crea menu del tuo ristorante
2. **Personalizzazione**: Modifica colori/stili  
3. **Deploy**: Metti online con Heroku/AWS
4. **Backup**: Salva cartella `storage/` regolarmente

## 🏆 SUCCESSO!

Il tuo sistema QR Menu è completamente operativo e pronto per l'uso professionale!

**Per supporto**: Consulta README.md per documentazione completa

---
*Sistema creato: Febbraio 2026*  
*Tecnologie: Go, HTML5, CSS3, JavaScript*  
*Librerie: Gorilla Mux, go-qrcode*