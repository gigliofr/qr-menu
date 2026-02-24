# Test delle API QR Menu System
# Esegui: .\test_api.ps1

$baseUrl = "http://localhost:8080"
$headers = @{ 'Content-Type' = 'application/json' }

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "    Test QR Menu System API" -ForegroundColor Cyan  
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Test 1: Verificare che il server sia attivo
Write-Host "Test 1: Verifica connessione server..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "$baseUrl/api/menus" -Method GET
    if ($response.StatusCode -eq 200) {
        Write-Host "✅ Server attivo e raggiungibile" -ForegroundColor Green
    }
} catch {
    Write-Host "❌ Server non raggiungible. Assicurati che sia avviato su $baseUrl" -ForegroundColor Red
    Write-Host "   Avvia il server con: .\start.bat" -ForegroundColor Yellow
    exit 1
}

# Test 2: Creare un menu di test
Write-Host "Test 2: Creazione menu di test..." -ForegroundColor Yellow  

$menuData = @{
    restaurant_id = "Ristorante Test API"
    name = "Menu di Test"
    description = "Menu creato tramite API per test"
    categories = @(
        @{
            id = "cat-antipasti"
            name = "Antipasti"
            description = "I nostri gustosi antipasti"
            items = @(
                @{
                    id = "item-bruschetta"
                    name = "Bruschetta al Pomodoro"
                    description = "Pane tostato con pomodori freschi e basilico"
                    price = 7.50
                    category = "Antipasti"
                    available = $true
                },
                @{
                    id = "item-salumi"
                    name = "Tagliere di Salumi e Formaggi"
                    description = "Selezione di salumi e formaggi locali"
                    price = 12.00
                    category = "Antipasti"
                    available = $true
                }
            )
        },
        @{
            id = "cat-primi"
            name = "Primi Piatti"
            description = "I nostri primi della tradizione"
            items = @(
                @{
                    id = "item-pasta"
                    name = "Spaghetti alla Carbonara"
                    description = "Pasta fresca con uova, guanciale e pecorino"
                    price = 11.00
                    category = "Primi Piatti"
                    available = $true
                },
                @{
                    id = "item-risotto"
                    name = "Risotto ai Porcini"
                    description = "Risotto cremoso con funghi porcini"
                    price = 13.50
                    category = "Primi Piatti"
                    available = $true
                }
            )
        }
    )
} | ConvertTo-Json -Depth 10

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/menu" -Method POST -Body $menuData -Headers $headers
    $menuId = $response.id
    Write-Host "✅ Menu creato con successo! ID: $menuId" -ForegroundColor Green
    
    # Test 3: Generare QR Code
    Write-Host "Test 3: Generazione QR Code..." -ForegroundColor Yellow
    
    $qrResponse = Invoke-RestMethod -Uri "$baseUrl/api/menu/$menuId/generate-qr" -Method POST -Headers $headers
    if ($qrResponse.success) {
        Write-Host "✅ QR Code generato con successo!" -ForegroundColor Green
        Write-Host "   URL Menu: $($qrResponse.menu_url)" -ForegroundColor Cyan
        Write-Host "   QR Code: $($qrResponse.qr_code_url)" -ForegroundColor Cyan
    }
    
    # Test 4: Recuperare il menu
    Write-Host "Test 4: Recupero menu creato..." -ForegroundColor Yellow
    
    $retrievedMenu = Invoke-RestMethod -Uri "$baseUrl/api/menu/$menuId" -Method GET
    Write-Host "✅ Menu recuperato con successo!" -ForegroundColor Green
    Write-Host "   Nome: $($retrievedMenu.name)" -ForegroundColor Cyan
    Write-Host "   Categorie: $($retrievedMenu.categories.Length)" -ForegroundColor Cyan
    Write-Host "   Completato: $($retrievedMenu.is_completed)" -ForegroundColor Cyan
    
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "           Test Completati!" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "🌐 Accedi all'interfaccia web:" -ForegroundColor White
    Write-Host "   Admin: $baseUrl/admin" -ForegroundColor Cyan
    Write-Host "   Menu Pubblico: $baseUrl/menu/$menuId" -ForegroundColor Cyan
    Write-Host ""
    
} catch {
    Write-Host "❌ Errore nella creazione del menu: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}