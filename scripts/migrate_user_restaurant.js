// ================================================
// Script di Migrazione MongoDB
// Da: Restaurant (con auth) → User + Restaurant (separati)
// ================================================
//
// ISTRUZIONI D'USO:
// 1. Backup del database prima di eseguire:
//    mongodump --uri="mongodb+srv://..." --db=qr-menu --out=backup_pre_migration
//
// 2. Esegui lo script:
//    mongosh "mongodb+srv://..." --file migrate_user_restaurant.js
//
// 3. Verifica i risultati:
//    - Controlla il log finale
//    - Verifica che tutti gli utenti esistano: db.users.count()
//    - Verifica che i restaurant siano aggiornati: db.restaurants.findOne()
//
// ================================================

use("qr-menu"); // Cambia con il nome del tuo database

print("\n================================================");
print("🔄 MIGRAZIONE: Restaurant → User + Restaurant");
print("================================================\n");

// Contatori per statistiche
let stats = {
    totalRestaurants: 0,
    usersCreated: 0,
    restaurantsUpdated: 0,
    sessionsUpdated: 0,
    errors: 0,
    startTime: new Date()
};

// Verifica che la collezione restaurants esista
stats.totalRestaurants = db.restaurants.countDocuments({});
print(`📊 Trovati ${stats.totalRestaurants} ristoranti da migrare\n`);

if (stats.totalRestaurants === 0) {
    print("⚠️  Nessun ristorante trovato. Migrazione non necessaria.\n");
    quit();
}

// Verifica se esistono già utenti (migrazione già eseguita?)
const existingUsers = db.users.countDocuments({});
if (existingUsers > 0) {
    print(`⚠️  ATTENZIONE: Trovati ${existingUsers} utenti esistenti.`);
    print("   La migrazione potrebbe essere già stata eseguita.");
    print("   Vuoi continuare comunque? (Ctrl+C per annullare)\n");
    // Attendi 5 secondi per dare tempo di annullare
    sleep(5000);
}

print("🚀 Inizio migrazione...\n");

// Itera su tutti i ristoranti
db.restaurants.find({}).forEach(function(oldRestaurant) {
    try {
        print(`📍 Migrando: ${oldRestaurant.name || 'Senza nome'} (ID: ${oldRestaurant._id})`);
        
        // ⭐ STEP 1: Verifica se ha campi auth (username, email, password_hash)
        if (!oldRestaurant.username && !oldRestaurant.email) {
            print("   ⏩ Già migrato (nessun campo auth), skip\n");
            return; // Skip, già migrato
        }
        
        // ⭐ STEP 2: Genera nuovo ID per l'utente
        const userID = new ObjectId().toString();
        
        // ⭐ STEP 3: Crea documento User
        const newUser = {
            _id: userID,
            username: oldRestaurant.username || `user_${oldRestaurant._id}`,
            email: oldRestaurant.email || `${oldRestaurant._id}@migrated.local`,
            password_hash: oldRestaurant.password_hash || "",
            
            // GDPR: Assume consenso per utenti esistenti (pre-GDPR)
            privacy_consent: true,
            marketing_consent: false, // Conservative default
            consent_date: oldRestaurant.created_at || new Date(),
            
            created_at: oldRestaurant.created_at || new Date(),
            last_login: oldRestaurant.last_login || oldRestaurant.created_at || new Date(),
            is_active: oldRestaurant.is_active !== undefined ? oldRestaurant.is_active : true
        };
        
        // Inserisci User
        db.users.insertOne(newUser);
        stats.usersCreated++;
        print(`   ✅ User creato: ${newUser.username} (ID: ${userID})`);
        
        // ⭐ STEP 4: Aggiorna Restaurant
        const updateResult = db.restaurants.updateOne(
            { _id: oldRestaurant._id },
            {
                $set: {
                    owner_id: userID // ⭐ Link al nuovo User
                },
                $unset: {
                    // Rimuovi campi auth (ora in User)
                    username: "",
                    email: "",
                    password_hash: "",
                    role: "",
                    last_login: ""
                }
            }
        );
        
        if (updateResult.modifiedCount > 0) {
            stats.restaurantsUpdated++;
            print(`   ✅ Restaurant aggiornato con owner_id`);
        }
        
        // ⭐ STEP 5: Aggiorna sessioni esistenti per questo ristorante
        const sessionUpdateResult = db.sessions.updateMany(
            { restaurant_id: oldRestaurant._id },
            { $set: { user_id: userID } }
        );
        
        if (sessionUpdateResult.modifiedCount > 0) {
            stats.sessionsUpdated += sessionUpdateResult.modifiedCount;
            print(`   ✅ ${sessionUpdateResult.modifiedCount} sessioni aggiornate`);
        }
        
        print("   ✓ Completato\n");
        
    } catch (error) {
        stats.errors++;
        print(`   ❌ ERRORE: ${error.message}\n`);
    }
});

// ================================================
// VERIFICA POST-MIGRAZIONE
// ================================================

print("\n================================================");
print("📊 STATISTICHE MIGRAZIONE");
print("================================================\n");

print(`⏱️  Tempo totale: ${(new Date() - stats.startTime) / 1000}s\n`);
print(`📈 Ristoranti trovati:    ${stats.totalRestaurants}`);
print(`✅ Users creati:          ${stats.usersCreated}`);
print(`✅ Restaurants aggiornati: ${stats.restaurantsUpdated}`);
print(`✅ Sessioni aggiornate:   ${stats.sessionsUpdated}`);
print(`❌ Errori:                ${stats.errors}\n`);

// Verifica conteggi finali
const finalUserCount = db.users.countDocuments({});
const finalRestaurantCount = db.restaurants.countDocuments({});
const restaurantsWithOwnerID = db.restaurants.countDocuments({ owner_id: { $exists: true, $ne: "" } });
const restaurantsWithAuth = db.restaurants.countDocuments({ username: { $exists: true } });

print("================================================");
print("🔍 VERIFICA DATABASE");
print("================================================\n");
print(`👥 Utenti totali:                    ${finalUserCount}`);
print(`🏪 Ristoranti totali:                ${finalRestaurantCount}`);
print(`🔗 Ristoranti con owner_id:          ${restaurantsWithOwnerID}`);
print(`⚠️  Ristoranti con campi auth rimasti: ${restaurantsWithAuth}\n`);

if (restaurantsWithAuth > 0) {
    print("⚠️  WARNING: Alcuni ristoranti hanno ancora campi auth!");
    print("   Verifica manualmente con: db.restaurants.find({ username: { $exists: true } })\n");
}

if (stats.errors > 0) {
    print(`❌ Migrazione completata con ${stats.errors} errori`);
    print("   Controlla i log sopra per i dettagli\n");
} else if (restaurantsWithOwnerID === finalRestaurantCount) {
    print("✅ Migrazione completata con successo!");
    print("   Tutti i ristoranti hanno owner_id\n");
} else {
    print("⚠️  Migrazione parziale");
    print(`   ${finalRestaurantCount - restaurantsWithOwnerID} ristoranti senza owner_id\n`);
}

// ================================================
// CREAZIONE INDICI (Performance)
// ================================================

print("================================================");
print("🔧 CREAZIONE INDICI DATABASE");
print("================================================\n");

try {
    // Indici su users
    db.users.createIndex({ username: 1 }, { unique: true, name: "idx_username" });
    print("✅ Indice creato: users.username (unique)");
    
    db.users.createIndex({ email: 1 }, { unique: true, name: "idx_email" });
    print("✅ Indice creato: users.email (unique)");
    
    db.users.createIndex({ is_active: 1, last_login: -1 }, { name: "idx_active_login" });
    print("✅ Indice creato: users.is_active + last_login");
    
    // Indici su restaurants
    db.restaurants.createIndex({ owner_id: 1, is_active: 1 }, { name: "idx_owner_active" });
    print("✅ Indice creato: restaurants.owner_id + is_active");
    
    db.restaurants.createIndex({ owner_id: 1, created_at: -1 }, { name: "idx_owner_created" });
    print("✅ Indice creato: restaurants.owner_id + created_at");
    
    // Indici su sessions
    db.sessions.createIndex({ user_id: 1, last_accessed: -1 }, { name: "idx_user_session" });
    print("✅ Indice creato: sessions.user_id + last_accessed");
    
    db.sessions.createIndex({ restaurant_id: 1 }, { name: "idx_restaurant_session" });
    print("✅ Indice creato: sessions.restaurant_id");
    
    // Indice TTL per sessioni scadute (30 giorni)
    db.sessions.createIndex({ last_accessed: 1 }, { 
        expireAfterSeconds: 2592000, // 30 giorni
        name: "idx_session_ttl" 
    });
    print("✅ Indice TTL creato: sessions.last_accessed (30 giorni)");
    
    print("\n✅ Tutti gli indici creati con successo!\n");
    
} catch (error) {
    print(`\n❌ Errore nella creazione degli indici: ${error.message}\n`);
}

print("================================================");
print("📝 PROSSIMI PASSI");
print("================================================\n");
print("1. Verifica l'applicazione con i nuovi dati");
print("2. Testa login e registrazione");
print("3. Verifica selezione multi-ristorante");
print("4. Se tutto funziona, elimina il backup:\n");
print("   rm -rf backup_pre_migration\n");
print("================================================\n");
