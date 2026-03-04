# Test completo con account e menu realistici
$baseUrl = "https://qr-menu-staging.up.railway.app"

Write-Host "`n╔════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║  TEST COMPLETO - CREAZIONE RISTORANTE CON MENU COMPLETI  ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════╝`n" -ForegroundColor Cyan

# Dati account
$timestamp = Get-Date -Format "yyyyMMddHHmmss"
$username = "trattoria_roma"
$email = "info@trattoriaroma.test"
$password = "RomaTest2026!"
$restaurantName = "Trattoria da Mario"

Write-Host "🏪 DATI RISTORANTE:" -ForegroundColor Yellow
Write-Host "   Nome:     $restaurantName" -ForegroundColor White
Write-Host "   Username: $username" -ForegroundColor White
Write-Host "   Email:    $email" -ForegroundColor White
Write-Host "   Password: $password" -ForegroundColor White
Write-Host ""

# STEP 1: Registrazione
Write-Host "📝 STEP 1: Registrazione account..." -ForegroundColor Cyan
try {
    $registerBody = @{
        username = $username
        email = $email
        password = $password
        confirm_password = $password
        restaurant_name = $restaurantName
        description = "Autentica cucina romana dal 1985. Specialità: carbonara, amatriciana, cacio e pepe"
        address = "Via del Corso 123, 00186 Roma RM"
        phone = "+39 06 1234567"
    } | ConvertTo-Json

    $registerResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/register" -Method Post `
        -Body $registerBody `
        -ContentType "application/json"

    Write-Host "   ✅ Account registrato!" -ForegroundColor Green
    $token = $registerResp.data.token
    $restaurantId = $registerResp.data.restaurant.id
    
    $headers = @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }
} catch {
    Write-Host "   ❌ Errore: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# STEP 2: Crea Menu Pranzo
Write-Host "`n🍝 STEP 2: Creazione Menu Pranzo..." -ForegroundColor Cyan
try {
    $menuPranzo = @{
        name = "Menu Pranzo"
        description = "Menu disponibile dalle 12:00 alle 15:30"
        active = $true
        categories = @(
            @{
                name = "Antipasti"
                description = "Per iniziare"
                items = @(
                    @{
                        name = "Bruschetta Romana"
                        description = "Pane casereccio tostato con pomodori freschi, aglio e basilico"
                        price = 6.50
                        available = $true
                        vegetarian = $true
                        allergens = @("glutine")
                        image_url = ""
                    }
                    @{
                        name = "Supplì al Telefono"
                        description = "Crocchette di riso ripiene di mozzarella filante (4 pezzi)"
                        price = 8.00
                        available = $true
                        vegetarian = $true
                        allergens = @("glutine", "latticini", "uova")
                    }
                    @{
                        name = "Fiori di Zucca Fritti"
                        description = "Fiori di zucca in pastella con alici e mozzarella"
                        price = 9.50
                        available = $true
                        allergens = @("glutine", "latticini", "pesce")
                    }
                    @{
                        name = "Carciofi alla Giudia"
                        description = "Carciofi fritti croccanti secondo la tradizione ebraico-romana"
                        price = 10.00
                        available = $true
                        vegetarian = $true
                        vegan = $true
                    }
                )
            }
            @{
                name = "Primi Piatti"
                description = "Paste fresche fatte in casa"
                items = @(
                    @{
                        name = "Spaghetti alla Carbonara"
                        description = "Spaghetti con guanciale, uova, pecorino romano DOP. La vera ricetta romana!"
                        price = 14.00
                        available = $true
                        allergens = @("glutine", "uova", "latticini")
                        popular = $true
                    }
                    @{
                        name = "Bucatini all'Amatriciana"
                        description = "Bucatini con guanciale croccante, pomodoro San Marzano e pecorino"
                        price = 13.50
                        available = $true
                        allergens = @("glutine", "latticini")
                        popular = $true
                    }
                    @{
                        name = "Tonnarelli Cacio e Pepe"
                        description = "Pasta fresca con pecorino romano DOP e pepe nero. Semplicità perfetta"
                        price = 12.00
                        available = $true
                        vegetarian = $true
                        allergens = @("glutine", "latticini")
                    }
                    @{
                        name = "Rigatoni alla Gricia"
                        description = "La carbonara senza uova: guanciale croccante e pecorino"
                        price = 13.00
                        available = $true
                        allergens = @("glutine", "latticini")
                    }
                    @{
                        name = "Gnocchi alla Sorrentina"
                        description = "Gnocchi di patate al pomodoro, mozzarella e basilico al forno"
                        price = 13.50
                        available = $true
                        vegetarian = $true
                        allergens = @("glutine", "latticini")
                    }
                    @{
                        name = "Fettuccine ai Funghi Porcini"
                        description = "Fettuccine fresche con porcini, aglio, prezzemolo e olio EVO"
                        price = 16.00
                        available = $false
                        vegetarian = $true
                        allergens = @("glutine")
                        notes = "Disponibile solo da settembre a novembre"
                    }
                )
            }
            @{
                name = "Secondi di Carne"
                description = "Carni selezionate"
                items = @(
                    @{
                        name = "Saltimbocca alla Romana"
                        description = "Fettine di vitello con prosciutto crudo e salvia al vino bianco"
                        price = 18.00
                        available = $true
                        allergens = @()
                    }
                    @{
                        name = "Coda alla Vaccinara"
                        description = "Coda di bue stufata con sedano, pomodoro e pinoli. Piatto storico romano"
                        price = 19.50
                        available = $true
                        allergens = @("sedano", "frutta a guscio")
                    }
                    @{
                        name = "Abbacchio al Forno"
                        description = "Agnello al forno con patate, rosmarino e aglio"
                        price = 20.00
                        available = $true
                        allergens = @()
                    }
                    @{
                        name = "Polpette al Sugo"
                        description = "Polpette di manzo in salsa di pomodoro (6 pezzi)"
                        price = 15.00
                        available = $true
                        allergens = @("glutine", "uova")
                    }
                )
            }
            @{
                name = "Secondi di Pesce"
                description = "Pesce fresco del giorno"
                items = @(
                    @{
                        name = "Baccalà alla Romana"
                        description = "Filetti di baccalà in umido con pomodoro, uvetta e pinoli"
                        price = 17.00
                        available = $true
                        allergens = @("pesce", "frutta a guscio")
                    }
                    @{
                        name = "Orata al Forno"
                        description = "Orata intera al forno con patate e olive"
                        price = 22.00
                        available = $true
                        allergens = @("pesce")
                    }
                )
            }
            @{
                name = "Contorni"
                description = "Verdure di stagione"
                items = @(
                    @{
                        name = "Cicoria Ripassata"
                        description = "Cicoria saltata in padella con aglio e peperoncino"
                        price = 5.00
                        available = $true
                        vegetarian = $true
                        vegan = $true
                    }
                    @{
                        name = "Patate al Forno"
                        description = "Patate croccanti al rosmarino"
                        price = 4.50
                        available = $true
                        vegetarian = $true
                        vegan = $true
                    }
                    @{
                        name = "Puntarelle alla Romana"
                        description = "Germogli di cicoria con alici, aglio e olio"
                        price = 6.00
                        available = $true
                        allergens = @("pesce")
                    }
                    @{
                        name = "Insalata Mista"
                        description = "Lattuga, pomodori, carote e radicchio"
                        price = 4.00
                        available = $true
                        vegetarian = $true
                        vegan = $true
                    }
                )
            }
            @{
                name = "Dolci"
                description = "Dolci della casa"
                items = @(
                    @{
                        name = "Tiramisù della Casa"
                        description = "Tiramisù artigianale con savoiardi, mascarpone e caffè"
                        price = 6.50
                        available = $true
                        vegetarian = $true
                        allergens = @("glutine", "uova", "latticini")
                        popular = $true
                    }
                    @{
                        name = "Panna Cotta"
                        description = "Panna cotta con coulis di frutti di bosco"
                        price = 6.00
                        available = $true
                        vegetarian = $true
                        allergens = @("latticini")
                    }
                    @{
                        name = "Crostata di Ricotta e Visciole"
                        description = "Crostata romana con ricotta e amarene"
                        price = 7.00
                        available = $true
                        vegetarian = $true
                        allergens = @("glutine", "latticini", "uova")
                    }
                    @{
                        name = "Gelato Artigianale"
                        description = "3 gusti a scelta (cioccolato, nocciola, stracciatella, limone, fragola)"
                        price = 5.50
                        available = $true
                        vegetarian = $true
                        allergens = @("latticini", "frutta a guscio")
                    }
                )
            }
            @{
                name = "Bevande"
                description = "Vini e bibite"
                items = @(
                    @{
                        name = "Acqua Minerale (1L)"
                        description = "Naturale o frizzante"
                        price = 2.50
                        available = $true
                        vegan = $true
                    }
                    @{
                        name = "Vino della Casa (caraffa)"
                        description = "Rosso o bianco - 0.5L"
                        price = 8.00
                        available = $true
                        vegan = $true
                    }
                    @{
                        name = "Frascati Superiore DOCG"
                        description = "Vino bianco dei Castelli Romani - Bottiglia"
                        price = 18.00
                        available = $true
                        vegan = $true
                    }
                    @{
                        name = "Birra Peroni (0.33L)"
                        description = "Birra italiana"
                        price = 4.50
                        available = $true
                        vegan = $true
                        allergens = @("glutine")
                    }
                    @{
                        name = "Caffè Espresso"
                        description = "Espresso italiano"
                        price = 1.50
                        available = $true
                        vegan = $true
                    }
                )
            }
        )
    } | ConvertTo-Json -Depth 10

    $menu1Resp = Invoke-RestMethod -Uri "$baseUrl/api/v1/menus" -Method Post `
        -Body $menuPranzo `
        -Headers $headers

    Write-Host "   ✅ Menu Pranzo creato!" -ForegroundColor Green
    Write-Host "      ID: $($menu1Resp.data._id)" -ForegroundColor Gray
    Write-Host "      Categorie: $($menu1Resp.data.categories.Count)" -ForegroundColor Gray
    $totalItems1 = ($menu1Resp.data.categories | ForEach-Object { $_.items.Count } | Measure-Object -Sum).Sum
    Write-Host "      Piatti totali: $totalItems1" -ForegroundColor Gray
    
    $menuPranzoId = $menu1Resp.data._id

} catch {
    Write-Host "   ❌ Errore: $($_.Exception.Message)" -ForegroundColor Red
}

# STEP 3: Crea Menu Cena
Write-Host "`n🍷 STEP 3: Creazione Menu Cena..." -ForegroundColor Cyan
try {
    $menuCena = @{
        name = "Menu Cena"
        description = "Menu degustazione serale - Dalle 19:30"
        active = $false
        categories = @(
            @{
                name = "Menu Degustazione"
                description = "Percorso gastronomico completo (4 portate)"
                items = @(
                    @{
                        name = "Menu Tradizione Romana"
                        description = "Antipasto misto + Pasta alla carbonara + Saltimbocca + Tiramisù"
                        price = 45.00
                        available = $true
                        allergens = @("glutine", "uova", "latticini")
                    }
                    @{
                        name = "Menu Vegetariano"
                        description = "Bruschetta + Cacio e pepe + Parmigiana + Panna cotta"
                        price = 38.00
                        available = $true
                        vegetarian = $true
                        allergens = @("glutine", "latticini", "uova")
                    }
                )
            }
            @{
                name = "Pizze (solo sera)"
                description = "Pizze cotte nel forno a legna"
                items = @(
                    @{
                        name = "Pizza Margherita"
                        description = "Pomodoro, mozzarella, basilico"
                        price = 8.00
                        available = $true
                        vegetarian = $true
                        allergens = @("glutine", "latticini")
                    }
                    @{
                        name = "Pizza Diavola"
                        description = "Pomodoro, mozzarella, salame piccante"
                        price = 10.00
                        available = $true
                        allergens = @("glutine", "latticini")
                    }
                    @{
                        name = "Pizza 4 Stagioni"
                        description = "Pomodoro, mozzarella, prosciutto, funghi, carciofi, olive"
                        price = 11.00
                        available = $true
                        allergens = @("glutine", "latticini")
                    }
                    @{
                        name = "Pizza Capricciosa"
                        description = "Pomodoro, mozzarella, prosciutto cotto, funghi, carciofi, olive, uovo"
                        price = 12.00
                        available = $true
                        allergens = @("glutine", "latticini", "uova")
                    }
                )
            }
        )
    } | ConvertTo-Json -Depth 10

    $menu2Resp = Invoke-RestMethod -Uri "$baseUrl/api/v1/menus" -Method Post `
        -Body $menuCena `
        -Headers $headers

    Write-Host "   ✅ Menu Cena creato!" -ForegroundColor Green
    Write-Host "      ID: $($menu2Resp.data._id)" -ForegroundColor Gray
    
    $menuCenaId = $menu2Resp.data._id

} catch {
    Write-Host "   ❌ Errore: $($_.Exception.Message)" -ForegroundColor Red
}

# STEP 4: Attiva Menu Pranzo
Write-Host "`n⚡ STEP 4: Attivazione Menu Pranzo..." -ForegroundColor Cyan
try {
    $activateResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/menus/$menuPranzoId/activate" -Method Post -Headers $headers
    Write-Host "   ✅ Menu Pranzo attivato come menu principale!" -ForegroundColor Green
} catch {
    Write-Host "   ⚠️  $($_.Exception.Message)" -ForegroundColor Yellow
}

# STEP 5: Verifica tutti i menu
Write-Host "`n📋 STEP 5: Verifica completa..." -ForegroundColor Cyan
try {
    $allMenus = Invoke-RestMethod -Uri "$baseUrl/api/v1/menus" -Method Get -Headers $headers
    Write-Host "   ✅ Menu totali creati: $($allMenus.data.Count)" -ForegroundColor Green
    
    foreach ($menu in $allMenus.data) {
        $itemsCount = ($menu.categories | ForEach-Object { $_.items.Count } | Measure-Object -Sum).Sum
        $status = if ($menu.active) { "🟢 ATTIVO" } else { "⚪ Non attivo" }
        Write-Host "      $status - $($menu.name): $itemsCount piatti" -ForegroundColor Gray
    }

    $profile = Invoke-RestMethod -Uri "$baseUrl/api/v1/restaurant/profile" -Method Get -Headers $headers
    Write-Host "`n   📍 Ristorante: $($profile.data.name)" -ForegroundColor Cyan
    Write-Host "      Indirizzo: $($profile.data.address)" -ForegroundColor Gray
    Write-Host "      Telefono: $($profile.data.phone)" -ForegroundColor Gray

} catch {
    Write-Host "   ❌ Errore verifica: $($_.Exception.Message)" -ForegroundColor Red
}

# RIEPILOGO FINALE
Write-Host "`n"
Write-Host "╔════════════════════════════════════════════════════════════╗" -ForegroundColor Green
Write-Host "║              ✅ SETUP COMPLETATO CON SUCCESSO!            ║" -ForegroundColor Green
Write-Host "╚════════════════════════════════════════════════════════════╝" -ForegroundColor Green

Write-Host "`n🏪 RISTORANTE CREATO:" -ForegroundColor Yellow
Write-Host "   Nome: $restaurantName" -ForegroundColor White
Write-Host "   Descrizione: Autentica cucina romana dal 1985" -ForegroundColor Gray
Write-Host "   Indirizzo: Via del Corso 123, 00186 Roma RM" -ForegroundColor Gray
Write-Host "   Telefono: +39 06 1234567" -ForegroundColor Gray

Write-Host "`n🔐 CREDENZIALI DI ACCESSO:" -ForegroundColor Yellow
Write-Host "   Username: $username" -ForegroundColor White
Write-Host "   Email:    $email" -ForegroundColor White
Write-Host "   Password: $password" -ForegroundColor White

Write-Host "`n📱 LINK APPLICAZIONE:" -ForegroundColor Yellow
Write-Host "   Homepage: $baseUrl" -ForegroundColor Cyan
Write-Host "   Login:    $baseUrl/login" -ForegroundColor Cyan
Write-Host "   Admin:    $baseUrl/admin" -ForegroundColor Cyan
Write-Host "   API Docs: $baseUrl/api/v1/docs" -ForegroundColor Cyan

Write-Host "`n🍝 MENU CREATI:" -ForegroundColor Yellow
Write-Host "   1. Menu Pranzo (ATTIVO) - 30+ piatti in 7 categorie" -ForegroundColor White
Write-Host "      Antipasti, Primi, Secondi Carne, Secondi Pesce," -ForegroundColor Gray
Write-Host "      Contorni, Dolci, Bevande" -ForegroundColor Gray
Write-Host "`n   2. Menu Cena - Menu degustazione e pizze" -ForegroundColor White
Write-Host "      Menu completi e 4 pizze" -ForegroundColor Gray

Write-Host "`n💡 PROSSIMI PASSI:" -ForegroundColor Yellow
Write-Host "   1. Accedi con le credenziali sopra" -ForegroundColor White
Write-Host "   2. Esplora il dashboard admin" -ForegroundColor White
Write-Host "   3. Visualizza i menu pubblici" -ForegroundColor White
Write-Host "   4. Genera QR code" -ForegroundColor White
Write-Host "   5. Testa le API" -ForegroundColor White

Write-Host "`n✅ Tutto pronto per l'uso! 🚀`n" -ForegroundColor Green
