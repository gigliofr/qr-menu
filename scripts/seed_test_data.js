// ===================================================================
// Script di Seed Database - Pulisce e Ricrea Dati di Test
// ===================================================================
// 
// Uso: mongosh "MONGODB_URI" --file scripts/seed_test_data.js
//

const colors = {
    reset: "\x1b[0m",
    red: "\x1b[31m",
    green: "\x1b[32m",
    yellow: "\x1b[33m",
    blue: "\x1b[34m",
    cyan: "\x1b[36m",
};

function log(message, color = colors.reset) {
    print(color + message + colors.reset);
}

// ===================================================================
// STEP 1: PULIZIA DATABASE
// ===================================================================

log("\n================================================", colors.cyan);
log("🧹 PULIZIA DATABASE", colors.cyan);
log("================================================\n", colors.cyan);

db = db.getSiblingDB("qr-menu");

log("📊 Conteggio documenti PRIMA della pulizia:", colors.yellow);
log("   Users:            " + db.users.countDocuments());
log("   Restaurants:      " + db.restaurants.countDocuments());
log("   Sessions:         " + db.sessions.countDocuments());
log("   Menus:            " + db.menus.countDocuments());
log("   Analytics Events: " + db.analytics_events.countDocuments());
log("");

log("⚠️  Attendere 3 secondi prima di pulire...", colors.yellow);
sleep(3000);

// Pulizia collezioni
const deleteResults = {
    users: db.users.deleteMany({}),
    restaurants: db.restaurants.deleteMany({}),
    sessions: db.sessions.deleteMany({}),
    menus: db.menus.deleteMany({}),
    analytics_events: db.analytics_events.deleteMany({})
};

log("✅ Collezioni pulite:", colors.green);
log("   Users:            " + deleteResults.users.deletedCount + " eliminati");
log("   Restaurants:      " + deleteResults.restaurants.deletedCount + " eliminati");
log("   Sessions:         " + deleteResults.sessions.deletedCount + " eliminati");
log("   Menus:            " + deleteResults.menus.deletedCount + " eliminati");
log("   Analytics Events: " + deleteResults.analytics_events.deletedCount + " eliminati");
log("");

// ===================================================================
// STEP 2: CREAZIONE UTENTE AMMINISTRATIVO
// ===================================================================

log("\n================================================", colors.cyan);
log("👤 CREAZIONE UTENTE AMMINISTRATIVO", colors.cyan);
log("================================================\n", colors.cyan);

// Password: "admin" (bcrypt hash verificato e testato)
// Hash generato e verificato con bcrypt cost 10
const adminPasswordHash = "$2a$10$dopU1ueHFSSkCmD78zuJCe1H0jBgtfnDp.pofNxNrleXL5SEGiCVK";

const adminUser = {
    _id: "admin_user_001",
    username: "admin",
    email: "admin@qrmenu.local",
    password_hash: adminPasswordHash,
    privacy_consent: true,
    marketing_consent: false,
    consent_date: new Date(),
    created_at: new Date(),
    last_login: new Date(),
    is_active: true
};

try {
    db.users.insertOne(adminUser);
    log("✅ Utente admin creato:", colors.green);
    log("   Username: admin", colors.green);
    log("   Email:    admin@qrmenu.local", colors.green);
    log("   Password: admin", colors.green);
    log("   User ID:  " + adminUser._id, colors.green);
} catch (error) {
    log("❌ Errore creazione utente: " + error.message, colors.red);
}

log("");

// ===================================================================
// STEP 3: CREAZIONE 4 RISTORANTI DI TEST
// ===================================================================

log("\n================================================", colors.cyan);
log("🏪 CREAZIONE 4 RISTORANTI DI TEST", colors.cyan);
log("================================================\n", colors.cyan);

const restaurants = [
    {
        _id: "rest_001",
        owner_id: adminUser._id,
        name: "Pizzeria Napoletana",
        description: "Autentica pizza napoletana con forno a legna. Impasto lievitato 48 ore.",
        address: "Via Roma 123, Napoli, 80100",
        phone: "+39 081 1234567",
        logo: "",
        active_menu_id: "",
        created_at: new Date(),
        is_active: true
    },
    {
        _id: "rest_002",
        owner_id: adminUser._id,
        name: "Trattoria Toscana",
        description: "Cucina tipica toscana con ingredienti biologici e a km zero.",
        address: "Piazza del Duomo 45, Firenze, 50122",
        phone: "+39 055 9876543",
        logo: "",
        active_menu_id: "",
        created_at: new Date(),
        is_active: true
    },
    {
        _id: "rest_003",
        owner_id: adminUser._id,
        name: "Sushi-Ya Tokyo",
        description: "Sushi giapponese autentico con chef certificato da Tokyo. Pesce freschissimo.",
        address: "Corso Venezia 88, Milano, 20121",
        phone: "+39 02 5551234",
        logo: "",
        active_menu_id: "",
        created_at: new Date(),
        is_active: true
    },
    {
        _id: "rest_004",
        owner_id: adminUser._id,
        name: "Burger House Americana",
        description: "Hamburger gourmet con carne 100% italiana e pane fatto in casa.",
        address: "Via Garibaldi 77, Roma, 00153",
        phone: "+39 06 7778888",
        logo: "",
        active_menu_id: "",
        created_at: new Date(),
        is_active: true
    }
];

try {
    const insertResult = db.restaurants.insertMany(restaurants);
    log("✅ Ristoranti creati: " + insertResult.insertedIds.length, colors.green);
    restaurants.forEach(rest => {
        log("   📍 " + rest.name + " (ID: " + rest._id + ")", colors.green);
    });
} catch (error) {
    log("❌ Errore creazione ristoranti: " + error.message, colors.red);
}

log("");

// ===================================================================
// STEP 4: CREAZIONE MENU PER OGNI RISTORANTE
// ===================================================================

log("\n================================================", colors.cyan);
log("📋 CREAZIONE MENU PER OGNI RISTORANTE", colors.cyan);
log("================================================\n", colors.cyan);

// MENU 1: Pizzeria Napoletana
const menu1 = {
    _id: "menu_001",
    restaurant_id: "rest_001",
    name: "Menu Pizzeria - Primavera 2026",
    description: "Le nostre specialità napoletane",
    is_active: true,
    created_at: new Date(),
    sections: [
        {
            id: "sec_pizza_001",
            name: "Pizze Classiche",
            description: "Le tradizionali pizze napoletane",
            items: [
                {
                    id: "item_001",
                    name: "Margherita",
                    description: "Pomodoro, mozzarella di bufala DOP, basilico fresco",
                    price: 8.00,
                    image: "",
                    available: true,
                    allergens: ["latte", "glutine"]
                },
                {
                    id: "item_002",
                    name: "Marinara",
                    description: "Pomodoro, aglio, origano, olio EVO",
                    price: 6.50,
                    image: "",
                    available: true,
                    allergens: ["glutine"]
                },
                {
                    id: "item_003",
                    name: "Diavola",
                    description: "Pomodoro, mozzarella, salame piccante, peperoncino",
                    price: 9.50,
                    image: "",
                    available: true,
                    allergens: ["latte", "glutine"]
                },
                {
                    id: "item_004",
                    name: "Quattro Stagioni",
                    description: "Pomodoro, mozzarella, funghi, carciofi, prosciutto, olive",
                    price: 11.00,
                    image: "",
                    available: true,
                    allergens: ["latte", "glutine"]
                }
            ]
        },
        {
            id: "sec_pizza_002",
            name: "Pizze Speciali",
            description: "Le nostre creazioni gourmet",
            items: [
                {
                    id: "item_005",
                    name: "Bufala e Pomodorini",
                    description: "Mozzarella di bufala, pomodorini del piennolo, basilico",
                    price: 13.00,
                    image: "",
                    available: true,
                    allergens: ["latte", "glutine"]
                },
                {
                    id: "item_006",
                    name: "Tartufo Nero",
                    description: "Mozzarella, funghi porcini, tartufo nero, parmigiano",
                    price: 16.00,
                    image: "",
                    available: true,
                    allergens: ["latte", "glutine"]
                }
            ]
        },
        {
            id: "sec_bevande_001",
            name: "Bevande",
            description: "Bibite e birre artigianali",
            items: [
                {
                    id: "item_007",
                    name: "Birra Artigianale IPA",
                    description: "Birra artigianale italiana 33cl",
                    price: 5.50,
                    image: "",
                    available: true,
                    allergens: ["glutine"]
                },
                {
                    id: "item_008",
                    name: "Acqua Naturale/Frizzante",
                    description: "Bottiglia 75cl",
                    price: 2.00,
                    image: "",
                    available: true,
                    allergens: []
                }
            ]
        }
    ]
};

// MENU 2: Trattoria Toscana
const menu2 = {
    _id: "menu_002",
    restaurant_id: "rest_002",
    name: "Menu Toscano - Stagione",
    description: "Sapori autentici della Toscana",
    is_active: true,
    created_at: new Date(),
    sections: [
        {
            id: "sec_antipasti_001",
            name: "Antipasti",
            description: "I nostri antipasti toscani",
            items: [
                {
                    id: "item_011",
                    name: "Crostini Toscani",
                    description: "Crostini di pane toscano con paté di fegatini",
                    price: 7.00,
                    image: "",
                    available: true,
                    allergens: ["glutine"]
                },
                {
                    id: "item_012",
                    name: "Tagliere di Salumi e Formaggi",
                    description: "Selezione di salumi e formaggi toscani DOP",
                    price: 15.00,
                    image: "",
                    available: true,
                    allergens: ["latte"]
                }
            ]
        },
        {
            id: "sec_primi_001",
            name: "Primi Piatti",
            description: "Pasta fresca fatta in casa",
            items: [
                {
                    id: "item_013",
                    name: "Pici Cacio e Pepe",
                    description: "Pici fatti a mano con pecorino toscano e pepe nero",
                    price: 12.00,
                    image: "",
                    available: true,
                    allergens: ["glutine", "latte"]
                },
                {
                    id: "item_014",
                    name: "Pappardelle al Cinghiale",
                    description: "Pappardelle con ragù di cinghiale del Chianti",
                    price: 14.00,
                    image: "",
                    available: true,
                    allergens: ["glutine"]
                },
                {
                    id: "item_015",
                    name: "Ribollita Toscana",
                    description: "Zuppa tradizionale con pane, cavolo nero e fagioli",
                    price: 10.00,
                    image: "",
                    available: true,
                    allergens: ["glutine"]
                }
            ]
        },
        {
            id: "sec_secondi_001",
            name: "Secondi Piatti",
            description: "Carne alla brace",
            items: [
                {
                    id: "item_016",
                    name: "Bistecca alla Fiorentina",
                    description: "Chianina IGP 1kg, cotta al sangue",
                    price: 45.00,
                    image: "",
                    available: true,
                    allergens: []
                },
                {
                    id: "item_017",
                    name: "Tagliata di Manzo",
                    description: "Con rucola, grana e pomodorini",
                    price: 18.00,
                    image: "",
                    available: true,
                    allergens: ["latte"]
                }
            ]
        }
    ]
};

// MENU 3: Sushi-Ya Tokyo
const menu3 = {
    _id: "menu_003",
    restaurant_id: "rest_003",
    name: "Sushi Menu - Stagione 2026",
    description: "Sushi e specialità giapponesi",
    is_active: true,
    created_at: new Date(),
    sections: [
        {
            id: "sec_nigiri_001",
            name: "Nigiri",
            description: "Nigiri di pesce fresco (2 pezzi)",
            items: [
                {
                    id: "item_021",
                    name: "Sake Nigiri",
                    description: "Salmone norvegese",
                    price: 4.50,
                    image: "",
                    available: true,
                    allergens: ["pesce"]
                },
                {
                    id: "item_022",
                    name: "Maguro Nigiri",
                    description: "Tonno rosso",
                    price: 6.00,
                    image: "",
                    available: true,
                    allergens: ["pesce"]
                },
                {
                    id: "item_023",
                    name: "Ebi Nigiri",
                    description: "Gambero rosso",
                    price: 5.00,
                    image: "",
                    available: true,
                    allergens: ["crostacei"]
                },
                {
                    id: "item_024",
                    name: "Unagi Nigiri",
                    description: "Anguilla con salsa teriyaki",
                    price: 5.50,
                    image: "",
                    available: true,
                    allergens: ["pesce", "soia"]
                }
            ]
        },
        {
            id: "sec_maki_001",
            name: "Maki Rolls",
            description: "Hosomaki e Uramaki (8 pezzi)",
            items: [
                {
                    id: "item_025",
                    name: "California Roll",
                    description: "Surimi, avocado, cetriolo, sesamo",
                    price: 7.00,
                    image: "",
                    available: true,
                    allergens: ["pesce", "sesamo"]
                },
                {
                    id: "item_026",
                    name: "Spicy Tuna Roll",
                    description: "Tonno piccante, cetriolo, mayo piccante",
                    price: 8.50,
                    image: "",
                    available: true,
                    allergens: ["pesce", "uova"]
                },
                {
                    id: "item_027",
                    name: "Dragon Roll",
                    description: "Gambero tempura, anguilla, avocado, salsa special",
                    price: 12.00,
                    image: "",
                    available: true,
                    allergens: ["crostacei", "pesce", "glutine"]
                }
            ]
        },
        {
            id: "sec_sashimi_001",
            name: "Sashimi",
            description: "Pesce crudo a fette (5 pezzi)",
            items: [
                {
                    id: "item_028",
                    name: "Sake Sashimi",
                    description: "Salmone norvegese",
                    price: 10.00,
                    image: "",
                    available: true,
                    allergens: ["pesce"]
                },
                {
                    id: "item_029",
                    name: "Maguro Sashimi",
                    description: "Tonno rosso",
                    price: 14.00,
                    image: "",
                    available: true,
                    allergens: ["pesce"]
                },
                {
                    id: "item_030",
                    name: "Mix Sashimi (15 pz)",
                    description: "Selezione dello chef",
                    price: 25.00,
                    image: "",
                    available: true,
                    allergens: ["pesce"]
                }
            ]
        }
    ]
};

// MENU 4: Burger House Americana
const menu4 = {
    _id: "menu_004",
    restaurant_id: "rest_004",
    name: "Burger Menu 2026",
    description: "Hamburger gourmet e specialità americane",
    is_active: true,
    created_at: new Date(),
    sections: [
        {
            id: "sec_burgers_001",
            name: "Classic Burgers",
            description: "I nostri classici burger (con patatine)",
            items: [
                {
                    id: "item_031",
                    name: "American Classic",
                    description: "180g di manzo, lattuga, pomodoro, cipolla, salsa burger",
                    price: 11.00,
                    image: "",
                    available: true,
                    allergens: ["glutine", "latte", "senape"]
                },
                {
                    id: "item_032",
                    name: "Cheeseburger Deluxe",
                    description: "180g manzo, doppio cheddar, bacon, cetrioli, salsa BBQ",
                    price: 13.50,
                    image: "",
                    available: true,
                    allergens: ["glutine", "latte", "senape"]
                },
                {
                    id: "item_033",
                    name: "Bacon Burger",
                    description: "180g manzo, bacon croccante, cheddar, cipolla caramellata",
                    price: 12.50,
                    image: "",
                    available: true,
                    allergens: ["glutine", "latte"]
                }
            ]
        },
        {
            id: "sec_burgers_002",
            name: "Gourmet Burgers",
            description: "Le nostre creazioni speciali",
            items: [
                {
                    id: "item_034",
                    name: "Truffle Burger",
                    description: "200g manzo Black Angus, brie, rucola, crema al tartufo",
                    price: 16.00,
                    image: "",
                    available: true,
                    allergens: ["glutine", "latte"]
                },
                {
                    id: "item_035",
                    name: "Mexican Burger",
                    description: "180g manzo, guacamole, jalapeños, cheddar, salsa messicana",
                    price: 14.00,
                    image: "",
                    available: true,
                    allergens: ["glutine", "latte"]
                },
                {
                    id: "item_036",
                    name: "Veggie Burger",
                    description: "Burger vegetale Beyond Meat, insalata, pomodoro, salsa vegan",
                    price: 12.00,
                    image: "",
                    available: true,
                    allergens: ["glutine", "soia"]
                }
            ]
        },
        {
            id: "sec_sides_001",
            name: "Contorni e Bevande",
            description: "Contorni e drink",
            items: [
                {
                    id: "item_037",
                    name: "Patatine Fritte",
                    description: "Porzione grande con salse a scelta",
                    price: 4.50,
                    image: "",
                    available: true,
                    allergens: []
                },
                {
                    id: "item_038",
                    name: "Onion Rings",
                    description: "Anelli di cipolla fritti (8 pz)",
                    price: 5.00,
                    image: "",
                    available: true,
                    allergens: ["glutine"]
                },
                {
                    id: "item_039",
                    name: "Coca-Cola / Sprite / Fanta",
                    description: "Lattina 33cl",
                    price: 3.00,
                    image: "",
                    available: true,
                    allergens: []
                },
                {
                    id: "item_040",
                    name: "Birra Americana",
                    description: "Budweiser / Corona 33cl",
                    price: 4.50,
                    image: "",
                    available: true,
                    allergens: ["glutine"]
                }
            ]
        }
    ]
};

const menus = [menu1, menu2, menu3, menu4];

try {
    const insertResult = db.menus.insertMany(menus);
    log("✅ Menu creati: " + insertResult.insertedIds.length, colors.green);
    
    // Aggiorna active_menu_id per ogni ristorante
    db.restaurants.updateOne({ _id: "rest_001" }, { $set: { active_menu_id: "menu_001" } });
    db.restaurants.updateOne({ _id: "rest_002" }, { $set: { active_menu_id: "menu_002" } });
    db.restaurants.updateOne({ _id: "rest_003" }, { $set: { active_menu_id: "menu_003" } });
    db.restaurants.updateOne({ _id: "rest_004" }, { $set: { active_menu_id: "menu_004" } });
    
    log("✅ Menu attivati per ogni ristorante", colors.green);
    log("");
    
    menus.forEach((menu, idx) => {
        const totalItems = menu.sections.reduce((sum, sec) => sum + sec.items.length, 0);
        log("   📋 " + menu.name + " → " + menu.sections.length + " sezioni, " + totalItems + " piatti", colors.green);
    });
    
} catch (error) {
    log("❌ Errore creazione menu: " + error.message, colors.red);
}

log("");

// ===================================================================
// STEP 5: CREAZIONE INDICI (se non esistono)
// ===================================================================

log("\n================================================", colors.cyan);
log("📇 VERIFICA INDICI DATABASE", colors.cyan);
log("================================================\n", colors.cyan);

try {
    // Users indices
    db.users.createIndex({ username: 1 }, { unique: true, name: "idx_username" });
    db.users.createIndex({ email: 1 }, { unique: true, name: "idx_email" });
    db.users.createIndex({ is_active: 1, last_login: -1 }, { name: "idx_active_login" });
    
    // Restaurants indices
    db.restaurants.createIndex({ owner_id: 1, is_active: 1 }, { name: "idx_owner_active" });
    db.restaurants.createIndex({ owner_id: 1, created_at: -1 }, { name: "idx_owner_created" });
    
    // Sessions indices
    db.sessions.createIndex({ user_id: 1, last_accessed: -1 }, { name: "idx_user_accessed" });
    db.sessions.createIndex({ restaurant_id: 1 }, { name: "idx_restaurant" });
    db.sessions.createIndex({ last_accessed: 1 }, { expireAfterSeconds: 2592000, name: "idx_ttl_sessions" }); // 30 days
    
    // Menus indices
    db.menus.createIndex({ restaurant_id: 1, is_active: 1 }, { name: "idx_restaurant_active" });
    
    log("✅ Indici verificati/creati", colors.green);
} catch (error) {
    log("⚠️  Indici già esistenti o errore: " + error.message, colors.yellow);
}

log("");

// ===================================================================
// STEP 6: STATISTICHE FINALI
// ===================================================================

log("\n================================================", colors.cyan);
log("📊 STATISTICHE FINALI", colors.cyan);
log("================================================\n", colors.cyan);

const finalStats = {
    users: db.users.countDocuments(),
    restaurants: db.restaurants.countDocuments(),
    sessions: db.sessions.countDocuments(),
    menus: db.menus.countDocuments(),
    analytics_events: db.analytics_events.countDocuments()
};

log("✅ Database popolato con successo!", colors.green);
log("");
log("👥 Users:            " + finalStats.users, colors.cyan);
log("🏪 Restaurants:      " + finalStats.restaurants, colors.cyan);
log("🔑 Sessions:         " + finalStats.sessions, colors.cyan);
log("📋 Menus:            " + finalStats.menus, colors.cyan);
log("📊 Analytics Events: " + finalStats.analytics_events, colors.cyan);
log("");

log("================================================", colors.cyan);
log("🎉 SEED COMPLETATO CON SUCCESSO!", colors.green);
log("================================================", colors.cyan);
log("");
log("🔐 CREDENZIALI AMMINISTRATORE:", colors.yellow);
log("   Username: admin", colors.yellow);
log("   Password: admin", colors.yellow);
log("   Email:    admin@qrmenu.local", colors.yellow);
log("");
log("🏪 RISTORANTI CREATI:", colors.yellow);
log("   1. Pizzeria Napoletana    (10 piatti)", colors.yellow);
log("   2. Trattoria Toscana      (7 piatti)", colors.yellow);
log("   3. Sushi-Ya Tokyo         (10 piatti)", colors.yellow);
log("   4. Burger House Americana (10 piatti)", colors.yellow);
log("");
log("🚀 Accedi a: http://localhost:8080/login", colors.cyan);
log("   Poi vai su /select-restaurant per scegliere un ristorante", colors.cyan);
log("");
