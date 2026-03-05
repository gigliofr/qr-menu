// ===================================================================
// Script di Migrazione Username Ristoranti (one-shot, idempotente)
// ===================================================================
//
// Uso:
// mongosh "MONGODB_URI" --tls --tlsCertificateKeyFile "path/to/cert.pem" --file scripts/migrate_restaurant_usernames.js
//

use("qr-menu");

print("\n================================================");
print("🔧 MIGRAZIONE USERNAME RISTORANTI");
print("================================================\n");

const stats = {
    totalRestaurants: 0,
    alreadyValid: 0,
    updated: 0,
    skippedNoName: 0,
    errors: 0,
    startTime: new Date()
};

function normalizeUsername(name) {
    let username = (name || "").toString().trim().toLowerCase();
    username = username.replace(/[^a-z0-9]+/g, "-");
    username = username.replace(/^-+|-+$/g, "");

    if (!username) {
        username = "ristorante";
    }

    if (username.length > 40) {
        username = username.substring(0, 40).replace(/-+$/g, "");
    }

    if (!username) {
        username = "ristorante";
    }

    return username;
}

function isUsernameTaken(username, currentRestaurantId) {
    return db.restaurants.countDocuments({
        _id: { $ne: currentRestaurantId },
        username: username
    }) > 0;
}

function generateUniqueUsername(base, currentRestaurantId) {
    let candidate = base;
    let counter = 1;

    while (isUsernameTaken(candidate, currentRestaurantId)) {
        counter += 1;
        const suffix = `-${counter}`;
        let trimmedBase = base;

        if (trimmedBase.length + suffix.length > 50) {
            trimmedBase = trimmedBase.substring(0, 50 - suffix.length).replace(/-+$/g, "");
            if (!trimmedBase) {
                trimmedBase = "ristorante";
            }
        }

        candidate = `${trimmedBase}${suffix}`;

        if (counter > 5000) {
            throw new Error("Troppi tentativi per generare username univoco");
        }
    }

    return candidate;
}

stats.totalRestaurants = db.restaurants.countDocuments({});
print(`📊 Ristoranti totali: ${stats.totalRestaurants}`);

if (stats.totalRestaurants === 0) {
    print("⚠️  Nessun ristorante trovato. Nulla da migrare.\n");
    quit();
}

const cursor = db.restaurants.find({});

cursor.forEach((restaurant) => {
    try {
        const currentUsername = (restaurant.username || "").toString().trim();

        if (!restaurant.name || !restaurant.name.toString().trim()) {
            stats.skippedNoName++;
            print(`⚠️  Skip ristorante senza nome (ID: ${restaurant._id})`);
            return;
        }

        const base = normalizeUsername(restaurant.name);
        const desired = generateUniqueUsername(base, restaurant._id);

        if (currentUsername === desired) {
            stats.alreadyValid++;
            return;
        }

        const result = db.restaurants.updateOne(
            { _id: restaurant._id },
            { $set: { username: desired } }
        );

        if (result.modifiedCount > 0) {
            stats.updated++;
            print(`✅ ${restaurant.name} → ${desired}`);
        }
    } catch (error) {
        stats.errors++;
        print(`❌ Errore su ID ${restaurant._id}: ${error.message}`);
    }
});

print("\n================================================");
print("🔍 VERIFICA FINALE");
print("================================================\n");

const totalWithUsername = db.restaurants.countDocuments({
    username: { $exists: true, $nin: [null, ""] }
});

const duplicates = db.restaurants.aggregate([
    { $match: { username: { $exists: true, $nin: [null, ""] } } },
    { $group: { _id: "$username", count: { $sum: 1 } } },
    { $match: { count: { $gt: 1 } } },
    { $count: "duplicateCount" }
]).toArray();

const duplicateCount = duplicates.length > 0 ? duplicates[0].duplicateCount : 0;

print(`🏪 Totali:                 ${stats.totalRestaurants}`);
print(`✅ Aggiornati:            ${stats.updated}`);
print(`ℹ️  Già validi:           ${stats.alreadyValid}`);
print(`⚠️  Senza nome (skip):    ${stats.skippedNoName}`);
print(`❌ Errori:                ${stats.errors}`);
print(`🔗 Con username valorizzato: ${totalWithUsername}`);
print(`🧪 Username duplicati:     ${duplicateCount}`);
print(`⏱️  Tempo totale:          ${(new Date() - stats.startTime) / 1000}s\n`);

print("================================================");
print("🧱 INDICE UNICO username");
print("================================================\n");

try {
    db.restaurants.createIndex(
        { username: 1 },
        {
            unique: true,
            name: "idx_restaurants_username_unique",
            partialFilterExpression: {
                username: { $type: "string", $gt: "" }
            }
        }
    );
    print("✅ Indice creato/verificato: restaurants.username UNIQUE (partial)");
} catch (error) {
    print(`❌ Errore creazione indice: ${error.message}`);
}

print("\n================================================");
if (stats.errors === 0 && duplicateCount === 0) {
    print("🎉 Migrazione completata con successo");
} else {
    print("⚠️  Migrazione completata con warning/errori");
}
print("================================================\n");
