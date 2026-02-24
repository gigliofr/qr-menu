# 🎯 Analytics Dashboard - Quick Start Guide

Questo documento spiega come accedere e utilizzare il nuovo **Analytics Dashboard** integrato in QR Menu Enterprise.

## 🚀 Avvio Veloce

### 1. Compilare e Avviare il Server
```bash
# Build
go build -o qr-menu.exe .

# Avvio con script helper
.\start_with_analytics.bat

# Oppure direttamente
.\qr-menu.exe
```

### 2. Accedere al Dashboard
1. Vai a `http://localhost:8080/admin`
2. Login con le tue credenziali (registrazione se necessario)
3. Clicca sul bottone **📊 Analytics** nella barra superiore
4. Accederai a `http://localhost:8080/admin/analytics`

## 📊 Cosa Puoi Vedere

### Stats Cards principali
- **📈 Visualizzazioni Totali**: Numero di volte che il tuo menu è stato visto
- **🔗 Condivisioni**: Quante volte è stato condiviso
- **📱 Scansioni QR**: Quante persone hanno scansionato il QR code
- **👥 Visitatori Unici**: Numero di device/IP diversi

### Grafici Interattivi
1. **Trend Visualizzazioni**: Line chart che mostra trend nel tempo
2. **Dispositivi**: Pie chart di Mobile/Desktop/Tablet
3. **Orari di Picco**: Bar chart dei migliori orari per il traffico

### Insights Dettagliati
- 🌍 **Top Paesi**: Da dove vengono i tuoi visitatori
- 🌐 **Browser**: Quali browser usano
- 🔗 **Condivisioni per Piattaforma**: WhatsApp vs Telegram vs Facebook
- ⏰ **Trend Orari**: Quando c'è più traffico

## 📱 Come Funziona il Tracking

### ✅ Visualizzazioni del Menu
Automaticamente tracciato quando qualcuno accede a:
```
GET /menu/{id}
```

**Dati raccolti:**
- Device type (Mobile/Desktop)
- Browser e OS
- Paese (da geolocalizzazione IP)
- Orario accesso

### ✅ Scansioni QR Code
Automaticamente tracciato quando scannerizzano il QR via:
```
GET /r/{username}
```

**Dati raccolti:**
- Data/ora scansione
- Device dell'utente
- Localizzazione approssimativa

### ✅ Condivisioni
Tracciato quando l'utente condivide il menu su:
- WhatsApp
- Telegram  
- Facebook
- Twitter
- Copia Link

## 🔄 Filtri e Periodi

### Selezione Periodo
La dashboard permette di filtrare per:
- **7 giorni**: Trend settimanale
- **30 giorni**: Trend mensile
- **90 giorni**: Trend trimestrale

### Come filtrare
```
GET /admin/analytics?days=30
GET /api/analytics?days=90
```

## 💾 Esportazione Dati

### Export PDF
Clicca il bottone **📄 Esporta PDF** per scaricare:
- Statistiche in formato report
- Grafici inclusi
- Perfetto per presentazioni

### Export CSV
Clicca il bottone **📊 Esporta CSV** per scaricare:
- Dati grezzi per analisi in Excel
- Compatibile con foglio di calcolo
- Ideale per approfondimenti

## 🔌 API REST per Analytics

### Ottenere i dati in JSON

```bash
# Dati ultimi 7 giorni
curl -H "Cookie: session=..." \
  http://localhost:8080/api/analytics?days=7

# Dati ultimi 30 giorni
curl -H "Cookie: session=..." \
  http://localhost:8080/api/analytics?days=30
```

### Risposta Esempio
```json
{
  "total_views": 1234,
  "unique_views": 567,
  "total_shares": 89,
  "qr_scans": 45,
  "daily_trend": [
    {
      "date": "2024-01-15",
      "views": 156,
      "qr_scans": 8
    }
  ],
  "device_stats": {
    "Mobile": 800,
    "Desktop": 300,
    "Tablet": 134
  },
  "share_breakdown": {
    "whatsapp": 45,
    "telegram": 20,
    "facebook": 15,
    "twitter": 5,
    "copy_link": 4
  }
}
```

### Tracciare Share Events Manualmente

```bash
curl -X POST http://localhost:8080/api/track/share \
  -H "Content-Type: application/json" \
  -d '{
    "menu_id": "abc123",
    "platform": "whatsapp"
  }'
```

Piattaforme supportate:
- `whatsapp`
- `telegram`
- `facebook`
- `twitter`
- `copy_link`

## 🔐 Sicurezza

- ✅ Dashboard privato (solo admin loggato)
- ✅ Dati tracciati per ristorante (non cross-restaurant leaks)
- ✅ IP anonymizzazione (non salva IP completo)
- ✅ HTTPS ready (configurable in produzione)

## 📊 Interpretazione Dati

### Indicatori Importanti

| Metrica | Cosa Significa | Target |
|---------|---|---|
| **Total Views** | Visite complessive | ↑ Aumentare |
| **Unique Views** | Clienti diversi | ↑ Aumentare |
| **QR Scans** | Clienti in loco | ↑ Aumentare |
| **Share Rate** | % che condivide | 5-10% |
| **Mobile %** | % traffico mobile | 70-80% |

### Analisi per Orari

- **Peak Hours**: Quando ordini sono massimi
- **Quiet Hours**: Quando promozionare
- **Dayparts**: Colazione/Pranzo/Cena

Esempio:
```
Lunedì ore 12:00-13:00 = 250 visite
Lunedì ore 19:00-20:00 = 340 visite
→ Cena è il momento più popolare
```

### Analisi per Device

```
Mobile:  80% (70 visite)
Desktop: 15% (13 visite)
Tablet:   5% ( 4 visite)
```

✅ Buono: >70% mobile (ristorante mobile-friendly)
⚠️ Attenzione: <50% mobile (verificare UX mobile)

## 🚀 Best Practices

### 1. Monitora Regolarmente
- Controlla analytics ogni giorno
- Nota trend correlati con promozioni
- Identifica sezionali vs evergreen items

### 2. Ottimizza basato su Dati
- Evidenzia menu items con alto traffico
- Aumenta prezzi sui piatti popolari
- Promuovi piatti underperforming

### 3. Sfrutta Peak Hours
- Invia notifiche all'ora di picco
- Offri promazioni strategiche
- Prepara staff di conseguenza

### 4. Analizza la Concorrenza
- Confronta con settore
- Mobile traffic: 70-85%
- Share rate: 5-15%
- Daily active users: 10-20% dei menu serviti

## 🔧 Troubleshooting

### Dashboard non carica
```
❌ Problema: "Non autorizzato"
✅ Soluzione: Accedi con /login

❌ Problema: Dati vuoti
✅ Soluzione: Aspetta qualche minuto, eventualmente accedi al menu

❌ Problema: Grafici non visibili
✅ Soluzione: Svuota cache browser (Ctrl+Shift+R)
```

### Analytics non traccia
```
❌ Problema: Views rimane 0
✅ Soluzione: 
   1. Apri il menu con /menu/{id}
   2. Anche in incognito
   3. Controlla logs in /logs/

❌ Problema: QR scans non registrati
✅ Soluzione: Scannerizza via /r/{username}
```

## 📈 Roadmap Futuro

### v2.0 (Prossimamente)
- [ ] Email report settimanale/mensile
- [ ] Notifiche per anomalie
- [ ] Comparazione con settimana precedente
- [ ] Previsioni (predictive analytics)

### v3.0 (Advanced)
- [ ] Integrazione Google Analytics
- [ ] Integrazione Segment
- [ ] Custom events tracking
- [ ] Goals e conversion tracking

### v4.0 (Enterprise)
- [ ] Grafana dashboard integration
- [ ] Data warehouse (BigQuery/Snowflake)
- [ ] ML anomaly detection
- [ ] Real-time alerts

## 💡 Tips & Tricks

### Massimizzare Dati Accurati
```bash
# Test manuale tracking
curl http://localhost:8080/menu/test-menu

# Simula share
curl -X POST http://localhost:8080/api/track/share \
  -H "Content-Type: application/json" \
  -d '{"menu_id":"test","platform":"whatsapp"}'

# Controlla dati
curl http://localhost:8080/api/analytics?days=1
```

### Creare Menu di Prova
```javascript
// Apri console browser (F12) e:
fetch('/menu/123', { credentials: 'include' })
  .then(r => r.text())
  .then(t => console.log('Menu loaded'))
```

### Bulk Testing
```bash
# Genera 100 views
for ($i=1; $i -le 100; $i++) {
  curl -s http://localhost:8080/menu/test > /dev/null
}
```

## 📞 Support & Documentation

- 📖 Full documentation: [ANALYTICS_IMPLEMENTATION.md](./ANALYTICS_IMPLEMENTATION.md)
- 🐛 Test script: [test_analytics.ps1](./test_analytics.ps1)
- 🚀 Launch script: [start_with_analytics.bat](./start_with_analytics.bat)
- 📝 GitHub Issues: [Report a bug](https://github.com/gigliofr/qr-menu/issues)

## ✨ Funzionalità Highlights

```
✅ Real-time tracking (no delays)
✅ Device detection (Mobile/Desktop/Tablet)
✅ Browser fingerprinting (Chrome/Safari/Firefox)
✅ Geolocation (city/country level)
✅ Timezone awareness
✅ Time-series analytics
✅ Interactive charts (Chart.js)
✅ Export capabilities (PDF/CSV)
✅ Multi-restaurant support
✅ Secure (auth required)
```

---

**Buon analisi! 📊**

Il tuo QR Menu Enterprise ora tiene traccia di tutto ciò che conta.
Usa questi dati per crescere e migliorare il tuo business. 🚀
