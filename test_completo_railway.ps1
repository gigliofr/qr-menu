# Test completo QR Menu - Registrazione nuovo account e test funzionalità
# Attende che MongoDB sia connesso prima di procedere

$baseUrl = "https://qr-menu-staging.up.railway.app"

Write-Host "`n=== QR MENU - TEST COMPLETO CON REGISTRAZIONE ===" -ForegroundColor Cyan
Write-Host "Base URL: $baseUrl`n" -ForegroundColor Gray

# Genera dati random per account di test
$timestamp = Get-Date -Format "yyyyMMddHHmmss"
$testUsername = "test_user_$timestamp"
$testEmail = "test.user.$timestamp@qrmenu.test"
$testPassword = "TestPass123!"
$restaurantName = "Ristorante Test $timestamp"

Write-Host "📋 Dati account di test generati:" -ForegroundColor Yellow
Write-Host "   Username: $testUsername" -ForegroundColor White
Write-Host "   Email: $testEmail" -ForegroundColor White
Write-Host "   Password: $testPassword" -ForegroundColor White
Write-Host "   Ristorante: $restaurantName`n" -ForegroundColor White

# Step 1: Attendi MongoDB connessione
Write-Host "1. Verifica connessione MongoDB..." -ForegroundColor Yellow
$maxRetries = 10
$retryCount = 0
$mongoConnected = $false

while ($retryCount -lt $maxRetries -and -not $mongoConnected) {
    try {
        $health = Invoke-RestMethod -Uri "$baseUrl/api/v1/health" -Method Get -ErrorAction Stop
        if ($health.data.database -eq "connected" -or $health.data.database -eq "mongodb") {
            Write-Host "   ✅ MongoDB connesso!" -ForegroundColor Green
            $mongoConnected = $true
        } else {
            Write-Host "   ⏳ Tentativo $($retryCount + 1)/$maxRetries - Status: $($health.data.database)" -ForegroundColor Gray
            Start-Sleep -Seconds 3
            $retryCount++
        }
    } catch {
        Write-Host "   ⚠️  Errore health check, riprovo..." -ForegroundColor Gray
        Start-Sleep -Seconds 3
        $retryCount++
    }
}

if (-not $mongoConnected) {
    Write-Host "   ❌ MongoDB non connesso dopo $maxRetries tentativi" -ForegroundColor Red
    Write-Host "   Continuo comunque con i test (potrebbero fallire)...`n" -ForegroundColor Yellow
}

# Step 2: Registrazione nuovo account
Write-Host "`n2. Registrazione nuovo account..." -ForegroundColor Yellow
try {
    $registerBody = @{
        username = $testUsername
        email = $testEmail
        password = $testPassword
        confirm_password = $testPassword
        name = $restaurantName
        description = "Ristorante di test automatico creato per verificare funzionalità"
        address = "Via Test 123, Roma"
        phone = "+39 06 1234567"
    } | ConvertTo-Json

    $registerResponse = Invoke-WebRequest -Uri "$baseUrl/register" -Method Post `
        -Body $registerBody `
        -ContentType "application/json" `
        -ErrorAction Stop

    if ($registerResponse.StatusCode -eq 200 -or $registerResponse.StatusCode -eq 302) {
        Write-Host "   ✅ Account registrato con successo!" -ForegroundColor Green
    }
} catch {
    Write-Host "   ❌ Errore registrazione: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.ErrorDetails.Message) {
        Write-Host "      Details: $($_.ErrorDetails.Message)" -ForegroundColor Gray
    }
}

# Step 3: Login con nuovo account
Write-Host "`n3. Login con account di test..." -ForegroundColor Yellow
Start-Sleep -Seconds 2
try {
    $loginBody = @{
        username = $testEmail
        password = $testPassword
    } | ConvertTo-Json

    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/login" -Method Post `
        -Body $loginBody `
        -ContentType "application/json" `
        -ErrorAction Stop

    Write-Host "   ✅ Login successful!" -ForegroundColor Green
    Write-Host "      User: $($loginResponse.data.user.username)" -ForegroundColor Gray
    Write-Host "      Role: $($loginResponse.data.user.role)" -ForegroundColor Gray
    
    $token = $loginResponse.data.token
    $userId = $loginResponse.data.user._id
    
    $headers = @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }

    # Step 4: Verifica profilo ristorante
    Write-Host "`n4. Recupero profilo ristorante..." -ForegroundColor Yellow
    try {
        $profile = Invoke-RestMethod -Uri "$baseUrl/api/v1/restaurant/profile" -Method Get -Headers $headers
        Write-Host "   ✅ Profilo recuperato" -ForegroundColor Green
        Write-Host "      Nome: $($profile.data.name)" -ForegroundColor Gray
        Write-Host "      Owner: $($profile.data.owner_id)" -ForegroundColor Gray
    } catch {
        Write-Host "   ❌ Errore profilo: $($_.Exception.Message)" -ForegroundColor Red
    }

    # Step 5: Crea primo menu di test
    Write-Host "`n5. Creazione menu 'Pranzo Italiano'..." -ForegroundColor Yellow
    try {
        $menu1 = @{
            name = "Menu Pranzo Italiano"
            description = "Specialità italiane per il pranzo"
            active = $true
            categories = @(
                @{
                    name = "Antipasti"
                    description = "Antipasti della casa"
                    items = @(
                        @{
                            name = "Bruschetta al Pomodoro"
                            description = "Pane tostato con pomodorini freschi, basilico e olio EVO"
                            price = 6.50
                            available = $true
                            allergens = @("glutine")
                        }
                        @{
                            name = "Carpaccio di Bresaola"
                            description = "Bresaola della Valtellina con rucola, grana e limone"
                            price = 9.00
                            available = $true
                        }
                    )
                }
                @{
                    name = "Primi Piatti"
                    description = "Paste fresche e risotti"
                    items = @(
                        @{
                            name = "Spaghetti alla Carbonara"
                            description = "Spaghetti con guanciale, uova, pecorino romano"
                            price = 12.00
                            available = $true
                            allergens = @("glutine", "uova", "latticini")
                        }
                        @{
                            name = "Risotto ai Funghi Porcini"
                            description = "Risotto mantecato con funghi porcini freschi"
                            price = 14.00
                            available = $true
                            vegetarian = $true
                        }
                    )
                }
                @{
                    name = "Secondi Piatti"
                    description = "Carni e pesci"
                    items = @(
                        @{
                            name = "Tagliata di Manzo"
                            description = "Tagliata di manzo con rucola e pomodorini"
                            price = 18.00
                            available = $true
                        }
                    )
                }
                @{
                    name = "Dolci"
                    description = "Dessert della casa"
                    items = @(
                        @{
                            name = "Tiramisù"
                            description = "Tiramisù artigianale con savoiardi e mascarpone"
                            price = 6.00
                            available = $true
                            allergens = @("glutine", "uova", "latticini")
                        }
                    )
                }
            )
        } | ConvertTo-Json -Depth 10

        $menu1Response = Invoke-RestMethod -Uri "$baseUrl/api/v1/menus" -Method Post `
            -Body $menu1 `
            -Headers $headers
        
        Write-Host "   ✅ Menu creato!" -ForegroundColor Green
        Write-Host "      ID: $($menu1Response.data._id)" -ForegroundColor Gray
        Write-Host "      Nome: $($menu1Response.data.name)" -ForegroundColor Gray
        Write-Host "      Categorie: $($menu1Response.data.categories.Count)" -ForegroundColor Gray
        
        $menuId1 = $menu1Response.data._id

        # Step 6: Crea secondo menu
        Write-Host "`n6. Creazione menu 'Cena Gourmet'..." -ForegroundColor Yellow
        $menu2 = @{
            name = "Menu Cena Gourmet"
            description = "Menu degustazione della sera"
            active = $false
            categories = @(
                @{
                    name = "Aperitivi"
                    items = @(
                        @{
                            name = "Selezione Formaggi e Salumi"
                            description = "Tagliere di formaggi e salumi selezionati"
                            price = 15.00
                            available = $true
                        }
                    )
                }
                @{
                    name = "Piatti Principali"
                    items = @(
                        @{
                            name = "Filetto al Vino Rosso"
                            description = "Filetto di manzo con riduzione di vino rosso"
                            price = 28.00
                            available = $true
                        }
                        @{
                            name = "Branzino al Forno"
                            description = "Branzino intero al forno con patate"
                            price = 24.00
                            available = $true
                        }
                    )
                }
            )
        } | ConvertTo-Json -Depth 10

        $menu2Response = Invoke-RestMethod -Uri "$baseUrl/api/v1/menus" -Method Post `
            -Body $menu2 `
            -Headers $headers
        
        Write-Host "   ✅ Menu creato!" -ForegroundColor Green
        Write-Host "      ID: $($menu2Response.data._id)" -ForegroundColor Gray
        $menuId2 = $menu2Response.data._id

        # Step 7: Lista tutti i menu
        Write-Host "`n7. Recupero lista menu..." -ForegroundColor Yellow
        $menusList = Invoke-RestMethod -Uri "$baseUrl/api/v1/menus" -Method Get -Headers $headers
        Write-Host "   ✅ Menu totali: $($menusList.data.Count)" -ForegroundColor Green
        foreach ($m in $menusList.data) {
            Write-Host "      - $($m.name) (ID: $($m._id))" -ForegroundColor Gray
        }

        # Step 8: Recupera singolo menu
        Write-Host "`n8. Recupero dettagli menu specifico..." -ForegroundColor Yellow
        $menuDetail = Invoke-RestMethod -Uri "$baseUrl/api/v1/menus/$menuId1" -Method Get -Headers $headers
        Write-Host "   ✅ Menu recuperato: $($menuDetail.data.name)" -ForegroundColor Green
        Write-Host "      Categorie totali: $($menuDetail.data.categories.Count)" -ForegroundColor Gray
        $totalItems = ($menuDetail.data.categories | ForEach-Object { $_.items.Count } | Measure-Object -Sum).Sum
        Write-Host "      Piatti totali: $totalItems" -ForegroundColor Gray

        # Step 9: Aggiorna menu
        Write-Host "`n9. Aggiornamento menu..." -ForegroundColor Yellow
        $updateMenu = @{
            name = "Menu Pranzo Italiano (Aggiornato)"
            description = "Specialità italiane - Aggiornato oggi!"
            active = $true
        } | ConvertTo-Json

        $updateResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/menus/$menuId1" -Method Put `
            -Body $updateMenu `
            -Headers $headers
        Write-Host "   ✅ Menu aggiornato!" -ForegroundColor Green

        # Step 10: Attiva menu
        Write-Host "`n10. Attivazione menu come menu principale..." -ForegroundColor Yellow
        try {
            $activateResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/menus/$menuId1/activate" -Method Post -Headers $headers
            Write-Host "   ✅ Menu attivato!" -ForegroundColor Green
        } catch {
            Write-Host "   ⚠️  $($_.Exception.Message)" -ForegroundColor Yellow
        }

        # Step 11: Test tracking analytics
        Write-Host "`n11. Test tracking analytics (share)..." -ForegroundColor Yellow
        try {
            $trackBody = @{
                menu_id = $menuId1
                event = "share"
                platform = "test"
            } | ConvertTo-Json
            
            Invoke-RestMethod -Uri "$baseUrl/api/track/share" -Method Post `
                -Body $trackBody `
                -ContentType "application/json"
            Write-Host "   ✅ Event tracked!" -ForegroundColor Green
        } catch {
            Write-Host "   ⚠️  Analytics: $($_.Exception.Message)" -ForegroundColor Yellow
        }

        # Step 12: Test analytics dashboard
        Write-Host "`n12. Recupero analytics..." -ForegroundColor Yellow
        try {
            $analytics = Invoke-RestMethod -Uri "$baseUrl/api/analytics" -Method Get -Headers $headers
            Write-Host "   ✅ Analytics recuperate" -ForegroundColor Green
        } catch {
            Write-Host "   ⚠️  $($_.Exception.Message)" -ForegroundColor Yellow
        }

    } catch {
        Write-Host "   ❌ Errore: $($_.Exception.Message)" -ForegroundColor Red
        if ($_.ErrorDetails.Message) {
            Write-Host "      Details: $($_.ErrorDetails.Message)" -ForegroundColor Gray
        }
    }

} catch {
    Write-Host "   ❌ Login failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Summary
Write-Host "`n" + "="*70 -ForegroundColor Cyan
Write-Host "=== RIEPILOGO TEST ===" -ForegroundColor Cyan
Write-Host "="*70 -ForegroundColor Cyan

Write-Host "`n📱 URL Applicazione:" -ForegroundColor Yellow
Write-Host "   $baseUrl" -ForegroundColor White

Write-Host "`n🔐 CREDENZIALI ACCOUNT DI TEST:" -ForegroundColor Yellow
Write-Host "   Username: $testUsername" -ForegroundColor White
Write-Host "   Email:    $testEmail" -ForegroundColor White  
Write-Host "   Password: $testPassword" -ForegroundColor White

Write-Host "`n🏪 Ristorante:" -ForegroundColor Yellow
Write-Host "   Nome: $restaurantName" -ForegroundColor White

Write-Host "`n🔗 Link Utili:" -ForegroundColor Yellow
Write-Host "   Login:    $baseUrl/login" -ForegroundColor White
Write-Host "   Admin:    $baseUrl/admin" -ForegroundColor White
Write-Host "   API Docs: $baseUrl/api/v1/docs" -ForegroundColor White

Write-Host "`n✅ Test completato!`n" -ForegroundColor Green
