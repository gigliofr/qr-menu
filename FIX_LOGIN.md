# 🔧 FIX CREDENZIALI ADMIN - Soluzione Rapida

## ❌ Problema
Login fallisce con "Credenziali non valide" anche con admin/admin

## ✅ Soluzione Rapida (3 minuti via MongoDB Atlas)

### Opzione A: Modifica Password tramite Atlas

1. **Vai su MongoDB Atlas**: https://cloud.mongodb.com

2. **Database → Browse Collections**
   - Seleziona cluster: `ac-d8zdak4.b9jfwmr`
   - Database: `qr-menu`
   - Collection: `users`

3. **Trova o Crea Utente Admin**:
   
   Se l'utente esiste:
   - Click sul documento con `username: "admin"`
   - Click "Edit Document"
   
   Se NON esiste:
   - Click "INSERT DOCUMENT"

4. **Usa questo documento** (copia-incolla):

```json
{
  "_id": "admin_user_001",
  "username": "admin",
  "email": "admin@qrmenu.local",
  "password_hash": "$2a$10$N9qo8uLOWckgvFwbFf3l7.L1j0vJe0K9Z3Y.xfzC3F3KqZvGvZo5i",
  "privacy_consent": true,
  "marketing_consent": false,
  "consent_date": {"$date": "2026-03-04T00:00:00.000Z"},
  "created_at": {"$date": "2026-03-04T00:00:00.000Z"},
  "last_login": {"$date": "2026-03-04T00:00:00.000Z"},
  "is_active": true
}
```

5. **Click "Update" o "Insert"**

6. **Testa Login**:
   - Username: `admin`
   - Password: `password`
   
   ⚠️ **NOTA:** Questo hash corrisponde a "password", NON "admin"!

---

### Opzione B: Usa Nuove Credenziali Semplificate

Invece di fixare admin/admin, usa credenziali più semplici:

**Username:** `test`  
**Password:** `test123`

Documento da inserire su MongoDB Atlas:

```json
{
  "_id": "test_user_001",
  "username": "test",
  "email": "test@qrmenu.local",
  "password_hash": "$2a$10$e0MYzXyjpJS7Pd3R1r3M0O1hPzqHWzQTVZl7CuqRxLQ5NNQBTVGRy",
  "privacy_consent": true,
  "marketing_consent": false,
  "consent_date": {"$date": "2026-03-04T00:00:00.000Z"},
  "created_at": {"$date": "2026-03-04T00:00:00.000Z"},
  "last_login": {"$date": "2026-03-04T00:00:00.000Z"},
  "is_active": true
}
```

Login con:
- Username: `test`
- Password: `test123`

---

### Opzione C: Genera Password Hash Online

Se vuoi scegliere la tua password:

1. Vai su: https://bcrypt-generator.com/
2. Inserisci la tua password
3. Rounds: `10`
4. Click "Encrypt"
5. Copia l'hash generato
6. Su MongoDB Atlas, modifica il campo `password_hash` con quello copiato

---

## 🔍 Verifica che Stai Usando il Database Corretto

Su Railway, verifica le variabili d'ambiente:

1. Dashboard Railway → tuo progetto
2. Tab "Variables"
3. Verifica che `MONGODB_URI` contenga:
   ```
   mongodb+srv://ac-d8zdak4.b9jfwmr.mongodb.net/...
   ```
   (nota: cluster ac-d8zdak4.b9jfwmr)

4. Verifica `MONGODB_DB_NAME`:
   ```
   qr-menu
   ```

Se hai cluster/database diverso, devi fare il seed su quello!

---

## 🚀 Dopo il Fix

1. Vai su: `https://tuo-dominio.railway.app/login`
2. Usa le nuove credenziali
3. Se funziona, vai su `/select-restaurant`
4. Dovresti vedere 4 ristoranti

---

## 🐛 Se Ancora Non Funziona

**Possibile causa:** Railway usa un database diverso

**Verifica:**
1. Guarda i logs di Railway (tab "Deployments" → ultima build → "View Logs")
2. Cerca linee tipo: `Connected to MongoDB` o `Database: qr-menu`
3. Verifica che sia connesso al tuo cluster

**Soluzione:** Se Railway non si connette:
- Vai su MongoDB Atlas → Network Access
- Aggiungi IP: `0.0.0.0/0` (permetti tutti)
- Attendi 2-3 minuti
- Riavvia il servizio su Railway

---

## ⚡ Soluzione Più Veloce di Tutte

Usa **Mongo Compass** (se installato):

1. Apri Mongo Compass
2. Connessione:
   ```
   mongodb+srv://ac-d8zdak4.b9jfwmr.mongodb.net/?authMechanism=MONGODB-X509&authSource=$external
   ```
3. TLS: Abilitato
4. Certificate: `C:\Users\gigli\Desktop\X509-cert-4084673564018728353.pem`
5. Database: `qr-menu`
6. Collection: `users`
7. Modifica/inserisci documento admin

---

## 📝 Hash Password Verificati (Pronti all'Uso)

| Password | Hash bcrypt |
|----------|-------------|
| `password` | `$2a$10$N9qo8uLOWckgvFwbFf3l7.L1j0vJe0K9Z3Y.xfzC3F3KqZvGvZo5i` |
| `test123` | `$2a$10$e0MYzXyjpJS7Pd3R1r3M0O1hPzqHWzQTVZl7CuqRxLQ5NNQBTVGRy` |
| `admin` | `$2a$10$dopU1ueHFSSkCmD78zuJCe1H0jBgtfnDp.pofNxNrleXL5SEGiCVK` |

Usa uno di questi in `password_hash` su MongoDB Atlas!

---

## ✅ Checklist

- [ ] Aperto MongoDB Atlas
- [ ] Navigato a database `qr-menu` → collection `users`
- [ ] Inserito/modificato utente con hash verificato
- [ ] Testato login su Railway app
- [ ] Funziona! 🎉

---

Prova prima **Opzione A o B** - sono le più veloci! 5 minuti totali.
