package db

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"qr-menu/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoClient wrapper per MongoDB Atlas con autenticazione X.509
type MongoClient struct {
	client *mongo.Client
	db     *mongo.Database
	ctx    context.Context
	cancel context.CancelFunc
}

var (
	// Instance globale del client MongoDB
	MongoInstance *MongoClient
)

// Connect crea connessione a MongoDB Atlas con certificato X.509
func Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Leggi certificato X.509 - supporta sia env var che file
	var certData []byte
	var err error

	// Opzione 1: Certificato come contenuto in variabile d'ambiente (per Railway/Cloud)
	certContent := os.Getenv("MONGODB_CERT_CONTENT")
	if certContent != "" {
		// Fix newlines: converte \n letterali in newlines reali
		certContent = strings.ReplaceAll(certContent, "\\n", "\n")
		certData = []byte(certContent)
		log.Println("✓ Certificato MongoDB caricato da MONGODB_CERT_CONTENT")
	} else {
		// Opzione 2: Certificato da file (per sviluppo locale)
		certPath := os.Getenv("MONGODB_CERT_PATH")
		if certPath == "" {
			return fmt.Errorf("nessun certificato MongoDB configurato: imposta MONGODB_CERT_CONTENT (contenuto) o MONGODB_CERT_PATH (path file)")
		}

		certData, err = ioutil.ReadFile(certPath)
		if err != nil {
			return fmt.Errorf("errore lettura certificato da %s: %v", certPath, err)
		}
		log.Printf("✓ Certificato MongoDB caricato da file: %s\n", certPath)
	}

	// Carica certificato per autenticazione X.509
	tlsConfig := &tls.Config{}

	// Parse certificato per client certificate authentication
	cert, err := tls.X509KeyPair(certData, certData)
	if err != nil {
		return fmt.Errorf("errore nel parsing del certificato: %v", err)
	}

	tlsConfig.Certificates = []tls.Certificate{cert}
	// Disabilita la verifica del server per MongoDB Atlas (usa il certificato self-signed)
	tlsConfig.InsecureSkipVerify = false

	// Connection string MongoDB Atlas
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		return fmt.Errorf("MONGODB_URI non configurato - imposta la connection string completa")
	}

	// Opzioni di connessione
	opts := options.Client().
		ApplyURI(mongoURI).
		SetTLSConfig(tlsConfig).
		SetServerSelectionTimeout(5 * time.Second)

	// Crea client
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return fmt.Errorf("errore connessione MongoDB: %v", err)
	}

	// Verifica connessione
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()

	if err := client.Ping(ctx2, nil); err != nil {
		return fmt.Errorf("errore ping MongoDB: %v", err)
	}

	// Salva istanza
	ctx3, cancel3 := context.WithTimeout(context.Background(), 30*time.Second)

	dbName := os.Getenv("MONGODB_DB_NAME")
	if dbName == "" {
		dbName = "qr-menu"
	}

	MongoInstance = &MongoClient{
		client: client,
		db:     client.Database(dbName),
		ctx:    ctx3,
		cancel: cancel3,
	}

	// Crea indici
	if err := MongoInstance.createIndexes(); err != nil {
		log.Printf("Avviso: errore nella creazione degli indici: %v", err)
	}

	log.Println("✓ Connesso a MongoDB Atlas")
	return nil
}

// Disconnect chiude la connessione
func (m *MongoClient) Disconnect() error {
	if m == nil || m.client == nil {
		return nil
	}

	m.cancel()
	if err := m.client.Disconnect(m.ctx); err != nil {
		return fmt.Errorf("errore disconnessione MongoDB: %v", err)
	}

	log.Println("✓ Disconnesso da MongoDB")
	return nil
}

// ==================== RESTAURANTS ====================

// CreateRestaurant salva un ristorante
func (m *MongoClient) CreateRestaurant(ctx context.Context, restaurant *models.Restaurant) error {
	coll := m.db.Collection("restaurants")
	_, err := coll.InsertOne(ctx, restaurant)
	if err != nil {
		return fmt.Errorf("errore insert restaurant: %v", err)
	}
	return nil
}

// GetRestaurantByID recupera un ristorante per ID
func (m *MongoClient) GetRestaurantByID(ctx context.Context, id string) (*models.Restaurant, error) {
	coll := m.db.Collection("restaurants")
	var restaurant models.Restaurant
	err := coll.FindOne(ctx, bson.M{"id": id}).Decode(&restaurant)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("errore find restaurant: %v", err)
	}
	return &restaurant, nil
}

// GetRestaurantByUsername recupera un ristorante per username
func (m *MongoClient) GetRestaurantByUsername(ctx context.Context, username string) (*models.Restaurant, error) {
	coll := m.db.Collection("restaurants")
	var restaurant models.Restaurant
	err := coll.FindOne(ctx, bson.M{"username": username}).Decode(&restaurant)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("errore find restaurant by username: %v", err)
	}
	return &restaurant, nil
}

// GetRestaurantByEmail recupera un ristorante per email
func (m *MongoClient) GetRestaurantByEmail(ctx context.Context, email string) (*models.Restaurant, error) {
	coll := m.db.Collection("restaurants")
	var restaurant models.Restaurant
	err := coll.FindOne(ctx, bson.M{"email": email}).Decode(&restaurant)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("errore find restaurant by email: %v", err)
	}
	return &restaurant, nil
}

// UpdateRestaurant aggiorna un ristorante
func (m *MongoClient) UpdateRestaurant(ctx context.Context, restaurant *models.Restaurant) error {
	coll := m.db.Collection("restaurants")
	result := coll.FindOneAndUpdate(ctx,
		bson.M{"id": restaurant.ID},
		bson.M{"$set": restaurant},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	if result.Err() != nil && result.Err() != mongo.ErrNoDocuments {
		return fmt.Errorf("errore update restaurant: %v", result.Err())
	}
	return nil
}

// GetAllRestaurants recupera tutti i ristoranti
func (m *MongoClient) GetAllRestaurants(ctx context.Context) ([]*models.Restaurant, error) {
	coll := m.db.Collection("restaurants")
	cursor, err := coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("errore find all restaurants: %v", err)
	}
	defer cursor.Close(ctx)

	var restaurants []*models.Restaurant
	if err = cursor.All(ctx, &restaurants); err != nil {
		return nil, fmt.Errorf("errore decode restaurants: %v", err)
	}
	return restaurants, nil
}

// ==================== MENUS ====================

// CreateMenu salva un menu
func (m *MongoClient) CreateMenu(ctx context.Context, menu *models.Menu) error {
	coll := m.db.Collection("menus")
	_, err := coll.InsertOne(ctx, menu)
	if err != nil {
		return fmt.Errorf("errore insert menu: %v", err)
	}
	return nil
}

// GetMenuByID recupera un menu per ID
func (m *MongoClient) GetMenuByID(ctx context.Context, id string) (*models.Menu, error) {
	coll := m.db.Collection("menus")
	var menu models.Menu
	err := coll.FindOne(ctx, bson.M{"id": id}).Decode(&menu)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("errore find menu: %v", err)
	}
	return &menu, nil
}

// GetMenusByRestaurantID recupera tutti i menu di un ristorante
func (m *MongoClient) GetMenusByRestaurantID(ctx context.Context, restaurantID string) ([]*models.Menu, error) {
	coll := m.db.Collection("menus")
	cursor, err := coll.Find(ctx, bson.M{"restaurant_id": restaurantID})
	if err != nil {
		return nil, fmt.Errorf("errore find menus: %v", err)
	}
	defer cursor.Close(ctx)

	var menus []*models.Menu
	if err = cursor.All(ctx, &menus); err != nil {
		return nil, fmt.Errorf("errore decode menus: %v", err)
	}
	return menus, nil
}

// UpdateMenu aggiorna un menu
func (m *MongoClient) UpdateMenu(ctx context.Context, menu *models.Menu) error {
	coll := m.db.Collection("menus")
	result := coll.FindOneAndUpdate(ctx,
		bson.M{"id": menu.ID},
		bson.M{"$set": menu},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	if result.Err() != nil && result.Err() != mongo.ErrNoDocuments {
		return fmt.Errorf("errore update menu: %v", result.Err())
	}
	return nil
}

// DeleteMenu elimina un menu
func (m *MongoClient) DeleteMenu(ctx context.Context, id string) error {
	coll := m.db.Collection("menus")
	result, err := coll.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return fmt.Errorf("errore delete menu: %v", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("menu non trovato")
	}
	return nil
}

// GetAllMenus recupera tutti i menu
func (m *MongoClient) GetAllMenus(ctx context.Context) ([]*models.Menu, error) {
	coll := m.db.Collection("menus")
	cursor, err := coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("errore find all menus: %v", err)
	}
	defer cursor.Close(ctx)

	var menus []*models.Menu
	if err = cursor.All(ctx, &menus); err != nil {
		return nil, fmt.Errorf("errore decode menus: %v", err)
	}
	return menus, nil
}

// ==================== SESSIONS ====================

// CreateSession salva una sessione
func (m *MongoClient) CreateSession(ctx context.Context, session *models.Session) error {
	coll := m.db.Collection("sessions")
	_, err := coll.InsertOne(ctx, session)
	if err != nil {
		return fmt.Errorf("errore insert session: %v", err)
	}
	return nil
}

// GetSessionByID recupera una sessione per ID
func (m *MongoClient) GetSessionByID(ctx context.Context, id string) (*models.Session, error) {
	coll := m.db.Collection("sessions")
	var session models.Session
	err := coll.FindOne(ctx, bson.M{"id": id}).Decode(&session)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("errore find session: %v", err)
	}
	return &session, nil
}

// GetSessionsByRestaurantID recupera tutte le sessioni di un ristorante
func (m *MongoClient) GetSessionsByRestaurantID(ctx context.Context, restaurantID string) ([]*models.Session, error) {
	coll := m.db.Collection("sessions")
	cursor, err := coll.Find(ctx, bson.M{
		"restaurant_id": restaurantID,
		"last_accessed": bson.M{
			"$gt": time.Now().Add(-24 * time.Hour),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("errore find sessions: %v", err)
	}
	defer cursor.Close(ctx)

	var sessions []*models.Session
	if err = cursor.All(ctx, &sessions); err != nil {
		return nil, fmt.Errorf("errore decode sessions: %v", err)
	}
	return sessions, nil
}

// UpdateSession aggiorna una sessione
func (m *MongoClient) UpdateSession(ctx context.Context, session *models.Session) error {
	coll := m.db.Collection("sessions")
	result := coll.FindOneAndUpdate(ctx,
		bson.M{"id": session.ID},
		bson.M{"$set": session},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	if result.Err() != nil && result.Err() != mongo.ErrNoDocuments {
		return fmt.Errorf("errore update session: %v", result.Err())
	}
	return nil
}

// DeleteSession elimina una sessione
func (m *MongoClient) DeleteSession(ctx context.Context, id string) error {
	coll := m.db.Collection("sessions")
	_, err := coll.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return fmt.Errorf("errore delete session: %v", err)
	}
	return nil
}

// DeleteExpiredSessions elimina sessioni scadute (>24h)
func (m *MongoClient) DeleteExpiredSessions(ctx context.Context) error {
	coll := m.db.Collection("sessions")
	_, err := coll.DeleteMany(ctx, bson.M{
		"last_accessed": bson.M{
			"$lt": time.Now().Add(-24 * time.Hour),
		},
	})
	if err != nil {
		return fmt.Errorf("errore delete expired sessions: %v", err)
	}
	return nil
}

// ==================== UTILITY ====================

// createIndexes crea gli indici necessari
func (m *MongoClient) createIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Indici per restaurants
	restColl := m.db.Collection("restaurants")
	restIndexModel := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "username", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "email", Value: 1}},
		},
	}
	if _, err := restColl.Indexes().CreateMany(ctx, restIndexModel); err != nil {
		return fmt.Errorf("errore creazione indici restaurants: %v", err)
	}

	// Indici per menus
	menuColl := m.db.Collection("menus")
	menuIndexModel := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "restaurant_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
		},
	}
	if _, err := menuColl.Indexes().CreateMany(ctx, menuIndexModel); err != nil {
		return fmt.Errorf("errore creazione indici menus: %v", err)
	}

	// Indici per sessions
	sessionColl := m.db.Collection("sessions")
	sessionIndexModel := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "restaurant_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "last_accessed", Value: -1}},
		},
		{
			Keys:    bson.D{{Key: "last_accessed", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(604800), // 7 giorni
		},
	}
	if _, err := sessionColl.Indexes().CreateMany(ctx, sessionIndexModel); err != nil {
		return fmt.Errorf("errore creazione indici sessions: %v", err)
	}

	// Indici per audit_logs
	auditColl := m.db.Collection("audit_logs")
	auditIndexModel := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "restaurant_id", Value: 1}, {Key: "timestamp", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "action", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "resource_type", Value: 1}},
		},
	}
	if _, err := auditColl.Indexes().CreateMany(ctx, auditIndexModel); err != nil {
		return fmt.Errorf("errore creazione indici audit_logs: %v", err)
	}

	// Indici per analytics_events
	analyticsColl := m.db.Collection("analytics_events")
	analyticsIndexModel := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "restaurant_id", Value: 1}, {Key: "timestamp", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "event_type", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "day_date", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
	}
	if _, err := analyticsColl.Indexes().CreateMany(ctx, analyticsIndexModel); err != nil {
		return fmt.Errorf("errore creazione indici analytics_events: %v", err)
	}

	return nil
}

// Ping verifica la connessione
func (m *MongoClient) Ping(ctx context.Context) error {
	return m.client.Ping(ctx, nil)
}
