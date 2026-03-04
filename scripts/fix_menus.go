package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"qr-menu/db"

	"go.mongodb.org/mongo-driver/bson"
)

type MenuItem struct {
	ID          string  `bson:"id"`
	Name        string  `bson:"name"`
	Description string  `bson:"description"`
	Price       float64 `bson:"price"`
	Category    string  `bson:"category"`
	Available   bool    `bson:"available"`
	ImageURL    string  `bson:"image_url"`
}

type MenuCategory struct {
	ID          string     `bson:"id"`
	Name        string     `bson:"name"`
	Description string     `bson:"description"`
	Items       []MenuItem `bson:"items"`
}

type Menu struct {
	ID           string         `bson:"id"`
	RestaurantID string         `bson:"restaurant_id"`
	Name         string         `bson:"name"`
	Description  string         `bson:"description"`
	MealType     string         `bson:"meal_type"`
	IsActive     bool           `bson:"is_active"`
	IsCompleted  bool           `bson:"is_completed"`
	CreatedAt    time.Time      `bson:"created_at"`
	UpdatedAt    time.Time      `bson:"updated_at"`
	Categories   []MenuCategory `bson:"categories"`
}

func main() {
	fmt.Println("\n================================================")
	fmt.Println("🔧 FIX MENU - Correzione Struttura Menu")
	fmt.Println("================================================\n")

	// Connetti a MongoDB usando il package db
	fmt.Println("🔌 Connessione a MongoDB...")
	if err := db.Connect(); err != nil {
		log.Fatal("❌ Errore connessione MongoDB:", err)
	}
	defer db.MongoInstance.Disconnect()
	
	fmt.Println("✅ Connesso a MongoDB\n")

	menusColl := db.MongoInstance.DB.Collection("menus")
	restaurantsColl := db.MongoInstance.DB.Collection("restaurants")
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Elimina menu esistenti
	fmt.Println("🗑️  Rimozione menu vecchi...")
	deleteResult, err := menusColl.DeleteMany(ctx, bson.M{})
	if err != nil {
		log.Fatal("❌ Errore eliminazione menu:", err)
	}
	fmt.Printf("✅ Menu eliminati: %d\n\n", deleteResult.DeletedCount)

	// Crea menu corretti
	menus := createMenus()

	fmt.Println("📋 Inserimento nuovi menu...")
	var menuDocs []interface{}
	for _, m := range menus {
		menuDocs = append(menuDocs, m)
	}

	insertResult, err := menusColl.InsertMany(ctx, menuDocs)
	if err != nil {
		log.Fatal("❌ Errore inserimento menu:", err)
	}
	fmt.Printf("✅ Menu creati: %d\n\n", len(insertResult.InsertedIDs))

	// Aggiorna active_menu_id per ogni ristorante
	fmt.Println("🔗 Collegamento menu ai ristoranti...")
	updates := map[string]string{
		"rest_001": "menu_001",
		"rest_002": "menu_002",
		"rest_003": "menu_003",
		"rest_004": "menu_004",
	}

	for restID, menuID := range updates {
		_, err := restaurantsColl.UpdateOne(
			ctx,
			bson.M{"_id": restID},
			bson.M{"$set": bson.M{"active_menu_id": menuID}},
		)
		if err != nil {
			log.Printf("⚠️  Errore aggiornamento ristorante %s: %v", restID, err)
		}
	}

	fmt.Println("✅ Menu attivati per ogni ristorante\n")

	// Stampa riepilogo
	for _, menu := range menus {
		totalItems := 0
		for _, cat := range menu.Categories {
			totalItems += len(cat.Items)
		}
		fmt.Printf("   📋 %s\n", menu.Name)
		fmt.Printf("      → %d categorie, %d piatti\n", len(menu.Categories), totalItems)
	}

	fmt.Println("\n✅ Fix completato!")
	fmt.Println("================================================\n")
}

func createMenus() []Menu {
	now := time.Now()

	// MENU 1: Pizzeria Napoletana
	menu1 := Menu{
		ID:           "menu_001",
		RestaurantID: "rest_001",
		Name:         "Menu Pizzeria - Primavera 2026",
		Description:  "Le nostre specialità napoletane",
		MealType:     "dinner",
		IsActive:     true,
		IsCompleted:  true,
		CreatedAt:    now,
		UpdatedAt:    now,
		Categories: []MenuCategory{
			{
				ID:          "cat_001",
				Name:        "Pizze Classiche",
				Description: "Le tradizionali pizze napoletane",
				Items: []MenuItem{
					{ID: "item_001", Name: "Margherita", Description: "Pomodoro, mozzarella di bufala DOP, basilico", Price: 8.00, Category: "Pizze Classiche", Available: true},
					{ID: "item_002", Name: "Marinara", Description: "Pomodoro, aglio, origano, olio EVO", Price: 6.50, Category: "Pizze Classiche", Available: true},
					{ID: "item_003", Name: "Diavola", Description: "Pomodoro, mozzarella, salame piccante", Price: 9.50, Category: "Pizze Classiche", Available: true},
				},
			},
			{
				ID:          "cat_002",
				Name:        "Pizze Speciali",
				Description: "Le nostre creazioni gourmet",
				Items: []MenuItem{
					{ID: "item_004", Name: "Bufala e Pomodorini", Description: "Mozzarella di bufala, pomodorini del piennolo", Price: 13.00, Category: "Pizze Speciali", Available: true},
					{ID: "item_005", Name: "Tartufo Nero", Description: "Mozzarella, funghi porcini, tartufo nero", Price: 16.00, Category: "Pizze Speciali", Available: true},
				},
			},
			{
				ID:          "cat_003",
				Name:        "Antipasti",
				Description: "Per iniziare",
				Items: []MenuItem{
					{ID: "item_006", Name: "Bruschette Miste", Description: "Pomodoro, olive, funghi", Price: 7.00, Category: "Antipasti", Available: true},
				},
			},
			{
				ID:          "cat_004",
				Name:        "Bevande",
				Description: "Bibite e birre",
				Items: []MenuItem{
					{ID: "item_007", Name: "Acqua Minerale", Description: "Naturale o frizzante 1L", Price: 2.50, Category: "Bevande", Available: true},
					{ID: "item_008", Name: "Birra Peroni", Description: "Bottiglia 66cl", Price: 5.00, Category: "Bevande", Available: true},
				},
			},
		},
	}

	// MENU 2: Trattoria Toscana
	menu2 := Menu{
		ID:           "menu_002",
		RestaurantID: "rest_002",
		Name:         "Menu Toscano - Stagionale",
		Description:  "I sapori della tradizione toscana",
		MealType:     "lunch",
		IsActive:     true,
		IsCompleted:  true,
		CreatedAt:    now,
		UpdatedAt:    now,
		Categories: []MenuCategory{
			{
				ID:          "cat_005",
				Name:        "Primi Piatti",
				Description: "Pasta fresca fatta in casa",
				Items: []MenuItem{
					{ID: "item_009", Name: "Pappardelle al Cinghiale", Description: "Pasta fresca con ragù di cinghiale", Price: 14.00, Category: "Primi Piatti", Available: true},
					{ID: "item_010", Name: "Ribollita", Description: "Zuppa di pane e verdure", Price: 10.00, Category: "Primi Piatti", Available: true},
					{ID: "item_011", Name: "Pici Cacio e Pepe", Description: "Pasta tipica senese", Price: 12.00, Category: "Primi Piatti", Available: true},
				},
			},
			{
				ID:          "cat_006",
				Name:        "Secondi Piatti",
				Description: "Carni alla brace",
				Items: []MenuItem{
					{ID: "item_012", Name: "Bistecca alla Fiorentina", Description: "Chianina 1kg (per 2 persone)", Price: 45.00, Category: "Secondi Piatti", Available: true},
					{ID: "item_013", Name: "Arista al Forno", Description: "Maiale con rosmarino e patate", Price: 18.00, Category: "Secondi Piatti", Available: true},
				},
			},
			{
				ID:          "cat_007",
				Name:        "Contorni",
				Description: "Verdure di stagione",
				Items: []MenuItem{
					{ID: "item_014", Name: "Fagioli all'Uccelletto", Description: "Fagioli con pomodoro e salvia", Price: 6.00, Category: "Contorni", Available: true},
					{ID: "item_015", Name: "Patate al Forno", Description: "Con rosmarino", Price: 5.00, Category: "Contorni", Available: true},
				},
			},
			{
				ID:          "cat_008",
				Name:        "Dolci",
				Description: "I nostri dessert",
				Items: []MenuItem{
					{ID: "item_016", Name: "Tiramisù", Description: "Ricetta tradizionale", Price: 7.00, Category: "Dolci", Available: true},
					{ID: "item_017", Name: "Cantucci e Vin Santo", Description: "Biscotti toscani con vino dolce", Price: 8.00, Category: "Dolci", Available: true},
				},
			},
		},
	}

	// MENU 3: Sushi-Ya Tokyo
	menu3 := Menu{
		ID:           "menu_003",
		RestaurantID: "rest_003",
		Name:         "Menu Sushi - Akira Selection",
		Description:  "Autenticità giapponese a Roma",
		MealType:     "dinner",
		IsActive:     true,
		IsCompleted:  true,
		CreatedAt:    now,
		UpdatedAt:    now,
		Categories: []MenuCategory{
			{
				ID:          "cat_009",
				Name:        "Nigiri",
				Description: "Pesce fresco su riso",
				Items: []MenuItem{
					{ID: "item_018", Name: "Nigiri Salmone", Description: "2 pezzi", Price: 6.00, Category: "Nigiri", Available: true},
					{ID: "item_019", Name: "Nigiri Tonno", Description: "2 pezzi", Price: 7.50, Category: "Nigiri", Available: true},
					{ID: "item_020", Name: "Nigiri Gambero Rosso", Description: "2 pezzi", Price: 8.00, Category: "Nigiri", Available: true},
				},
			},
			{
				ID:          "cat_010",
				Name:        "Maki",
				Description: "Rotolini di riso",
				Items: []MenuItem{
					{ID: "item_021", Name: "California Roll", Description: "Surimi, avocado, cetriolo - 8 pezzi", Price: 9.00, Category: "Maki", Available: true},
					{ID: "item_022", Name: "Spicy Tuna Roll", Description: "Tonno piccante, cetriolo - 8 pezzi", Price: 11.00, Category: "Maki", Available: true},
					{ID: "item_023", Name: "Philadelphia Roll", Description: "Salmone, formaggio - 8 pezzi", Price: 10.00, Category: "Maki", Available: true},
				},
			},
			{
				ID:          "cat_011",
				Name:        "Ramen",
				Description: "Zuppe tradizionali",
				Items: []MenuItem{
					{ID: "item_024", Name: "Shoyu Ramen", Description: "Brodo di soia, maiale, uovo marinato", Price: 14.00, Category: "Ramen", Available: true},
					{ID: "item_025", Name: "Miso Ramen", Description: "Brodo di miso, verdure, tofu", Price: 13.00, Category: "Ramen", Available: true},
				},
			},
			{
				ID:          "cat_012",
				Name:        "Bevande",
				Description: "Drink giapponesi",
				Items: []MenuItem{
					{ID: "item_026", Name: "Sake Caldo", Description: "Vino di riso giapponese", Price: 8.00, Category: "Bevande", Available: true},
					{ID: "item_027", Name: "Birra Asahi", Description: "Bottiglia 50cl", Price: 6.00, Category: "Bevande", Available: true},
				},
			},
		},
	}

	// MENU 4: Burger House Americana
	menu4 := Menu{
		ID:           "menu_004",
		RestaurantID: "rest_004",
		Name:         "Burger House Menu - Classic American",
		Description:  "I migliori burger di Milano",
		MealType:     "dinner",
		IsActive:     true,
		IsCompleted:  true,
		CreatedAt:    now,
		UpdatedAt:    now,
		Categories: []MenuCategory{
			{
				ID:          "cat_013",
				Name:        "Burgers",
				Description: "Carne 100% Black Angus",
				Items: []MenuItem{
					{ID: "item_028", Name: "Classic Burger", Description: "Carne, lattuga, pomodoro, cipolla, salse", Price: 12.00, Category: "Burgers", Available: true},
					{ID: "item_029", Name: "Cheeseburger Deluxe", Description: "Doppia carne, doppio cheddar, bacon", Price: 16.00, Category: "Burgers", Available: true},
					{ID: "item_030", Name: "BBQ Bacon Burger", Description: "Carne, bacon, cipolle caramellate, BBQ", Price: 14.50, Category: "Burgers", Available: true},
					{ID: "item_031", Name: "Veggie Burger", Description: "Burger vegetale, insalata, pomodoro", Price: 11.00, Category: "Burgers", Available: true},
				},
			},
			{
				ID:          "cat_014",
				Name:        "Contorni",
				Description: "I nostri sides",
				Items: []MenuItem{
					{ID: "item_032", Name: "Patatine Fritte", Description: "Croccanti e dorate", Price: 4.50, Category: "Contorni", Available: true},
					{ID: "item_033", Name: "Onion Rings", Description: "Anelli di cipolla fritti - 8 pezzi", Price: 5.50, Category: "Contorni", Available: true},
					{ID: "item_034", Name: "Chicken Wings", Description: "Alette piccanti - 6 pezzi", Price: 7.00, Category: "Contorni", Available: true},
				},
			},
			{
				ID:          "cat_015",
				Name:        "Dessert",
				Description: "Dolci americani",
				Items: []MenuItem{
					{ID: "item_035", Name: "New York Cheesecake", Description: "Con frutti di bosco", Price: 6.50, Category: "Dessert", Available: true},
					{ID: "item_036", Name: "Brownie al Cioccolato", Description: "Con gelato alla vaniglia", Price: 6.00, Category: "Dessert", Available: true},
				},
			},
			{
				ID:          "cat_016",
				Name:        "Bevande",
				Description: "Soft drinks e birre",
				Items: []MenuItem{
					{ID: "item_037", Name: "Milkshake", Description: "Vaniglia, cioccolato o fragola", Price: 5.50, Category: "Bevande", Available: true},
					{ID: "item_038", Name: "Coca-Cola", Description: "Bottiglia 50cl", Price: 3.50, Category: "Bevande", Available: true},
					{ID: "item_039", Name: "Birra Budweiser", Description: "Bottiglia 33cl", Price: 5.00, Category: "Bevande", Available: true},
				},
			},
		},
	}

	return []Menu{menu1, menu2, menu3, menu4}
}
