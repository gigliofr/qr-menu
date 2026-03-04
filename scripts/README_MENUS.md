# Come Creare i Menu per i Ristoranti

## ✅ Menu Già Creati

I menu sono già stati creati ed eseguiti con lo script `fix_menus.go`. Puoi accedere all'app per vederli:
- https://qr-menu-staging.up.railway.app/login
- Username: `admin` / Password: `admin`

## 📁 File Disponibili

- **`fix_menus.go`** - Script Go per creare/ricreare menu (già eseguito)
- **`fix_menus.js`** - Versione JavaScript mongosh (alternativa)
- **`fix_menus.ps1`** - PowerShell wrapper per fix_menus.js
- **`menus_data.json`** - Export JSON per MongoDB Compass/Atlas

---

## Opzioni per Modificare o Ricreare Menu

### Opzione 1: MongoDB Compass (CONSIGLIATO per modifiche manuali)

1. Apri MongoDB Compass e connettiti al tuo cluster
2. Seleziona il database `qr-menu`  
3. Seleziona la collection `menus`
4. Clicca su "ADD DATA" → "Import JSON or CSV"
5. Seleziona il file **`scripts/menus_data.json`** ✅

### Opzione 2: Usa lo script Go (per ricreare da zero)

```powershell
# Dalla cartella principale del progetto
$env:MONGODB_CERT_CONTENT = Get-Content "C:\path\to\cert.pem" -Raw
$env:MONGODB_URI = "mongodb+srv://..."
go run scripts/fix_menus.go
```

Questo elimina i menu esistenti e li ricrea.

### Opzione 3: MongoDB Atlas Web Interface

1. Vai su https://cloud.mongodb.com
2. Accedi al tuo cluster
3. Clicca su "Browse Collections"
4. Seleziona database `qr-menu` → collection `menus`
5. Clicca "Insert Document"
6. Copia e incolla il contenuto di ogni menu da **`scripts/menus_data.json`** ✅

### Opzione 4: Tramite l'interfaccia Web dell'App

1. Accedi su https://qr-menu-staging.up.railway.app/login
2. Username: `admin` Password: `admin`
3. Seleziona un ristorante
4. Vai su "Menu" → "Crea Nuovo Menu"
5. Compila il form e aggiungi categorie e piatti

⚠️ **NOTA**: L'opzione 4 è lenta se devi creare molti piatti. Meglio usare Compass, lo script Go, o Atlas.

## Struttura Menu

Ogni menu ha questa struttura:

```javascript
{
  "id": "menu_001",
  "restaurant_id": "rest_001",
  "name": "Menu Pizzeria - Primavera 2026",
  "description": "Le nostre specialità napoletane",
  "meal_type": "dinner",
  "is_active": true,
  "is_completed": true,
  "created_at": ISODate(),
  "updated_at": ISODate(),
  "categories": [
    {
      "id": "cat_001",
      "name": "Pizze Classiche",
      "description": "Le tradizionali pizze napoletane",
      "items": [
        {
          "id": "item_001",
          "name": "Margherita",
          "description": "Pomodoro, mozzarella di bufala DOP, basilico",
          "price": 8.00,
          "category": "Pizze Classiche",
          "available": true,
          "image_url": ""
        }
      ]
    }
  ]
}
```

## Generare  QR Code

Una volta che i menu sono nel database:

1. Accedi all'admin del ristorante
2. Vai su "Menu"
3. Seleziona il menu attivo
4. Clicca su "Genera QR Code"
5. Il QR Code sarà salvato e visibile nella pagina del menu

Il QR Code punterà all'URL pubblico del menu che i clienti possono scansionare.
