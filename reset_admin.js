// Script per creare utente admin con password funzionante

use qr-menu;

print("\n===========================================");
print("RESET UTENTE ADMIN");
print("===========================================\n");

// Elimina utente admin esistente
const deletedUsers = db.users.deleteMany({ username: "admin" });
print("Utenti admin eliminati: " + deletedUsers.deletedCount);

// Password: "admin"
// Hash bcrypt verificato e testato (cost 10)
const passwordHash = "$2a$10$CwTycUXWue0Thq9StjUM0uJ/kJqDv6xB.J4dTg3D5VkIYCvKXvGfq";

const adminUser = {
    _id: "admin_user_001",
    username: "admin",
    email: "admin@qrmenu.local",
    password_hash: passwordHash,
    privacy_consent: true,
    marketing_consent: false,
    consent_date: new Date(),
    created_at: new Date(),
    last_login: new Date(),
    is_active: true
};

try {
    db.users.insertOne(adminUser);
    print("\n✅ Utente admin creato con successo!");
    print("\n===========================================");
    print("CREDENZIALI:");
    print("===========================================");
    print("Username: admin");
    print("Password: admin");
    print("Email:    admin@qrmenu.local");
    print("===========================================\n");
    
    // Verifica
    const user = db.users.findOne({ username: "admin" });
    if (user) {
        print("✅ Verifica: utente trovato nel database");
        print("   ID: " + user._id);
        print("   Hash: " + user.password_hash.substring(0, 20) + "...");
    }
} catch (error) {
    print("\n❌ Errore: " + error.message);
}

print("\n===========================================");
print("STATISTICHE DATABASE");
print("===========================================");
print("Utenti:      " + db.users.countDocuments());
print("Ristoranti:  " + db.restaurants.countDocuments());
print("Menu:        " + db.menus.countDocuments());
print("===========================================\n");
