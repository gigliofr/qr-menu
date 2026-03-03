# Test automatico QR Menu su Railway
# Testa tutte le funzionalità principali dell'applicazione

$baseUrl = "https://qr-menu-staging.up.railway.app"
$username = "giglio.fr@gmail.com"
$password = "???Fr43V4l3"

Write-Host "`n=== TEST QR MENU APPLICATION ===" -ForegroundColor Cyan
Write-Host "Base URL: $baseUrl`n" -ForegroundColor Gray

# Test 1: Health Check
Write-Host "1. Testing Health Endpoint..." -ForegroundColor Yellow
try {
    $health = Invoke-RestMethod -Uri "$baseUrl/api/v1/health" -Method Get
    if ($health.data.database -eq "in-memory") {
        Write-Host "   ❌ FAIL: MongoDB not connected (using in-memory fallback)" -ForegroundColor Red
        Write-Host "      Database status: $($health.data.database)" -ForegroundColor Gray
    } else {
        Write-Host "   ✅ PASS: MongoDB connected" -ForegroundColor Green
    }
    Write-Host "   Version: $($health.data.version)" -ForegroundColor Gray
    Write-Host "   Services: $($health.data.services | ConvertTo-Json -Compress)" -ForegroundColor Gray
} catch {
    Write-Host "   ❌ FAIL: Health endpoint error: $_" -ForegroundColor Red
}

# Test 2: Home Page
Write-Host "`n2. Testing Home Page..." -ForegroundColor Yellow
try {
    $home = Invoke-WebRequest -Uri "$baseUrl/" -Method Get
    if ($home.StatusCode -eq 200) {
        Write-Host "   ✅ PASS: Home page loads (HTTP $($home.StatusCode))" -ForegroundColor Green
    }
} catch {
    Write-Host "   ❌ FAIL: Home page error: $_" -ForegroundColor Red
}

# Test 3: Login Page
Write-Host "`n3. Testing Login Page..." -ForegroundColor Yellow
try {
    $login = Invoke-WebRequest -Uri "$baseUrl/login" -Method Get
    if ($login.StatusCode -eq 200) {
        Write-Host "   ✅ PASS: Login page loads (HTTP $($login.StatusCode))" -ForegroundColor Green
    }
} catch {
    Write-Host "   ❌ FAIL: Login page error: $_" -ForegroundColor Red
}

# Test 4: Register Page
Write-Host "`n4. Testing Register Page..." -ForegroundColor Yellow
try {
    $register = Invoke-WebRequest -Uri "$baseUrl/register" -Method Get
    if ($register.StatusCode -eq 200) {
        Write-Host "   ✅ PASS: Register page loads (HTTP $($register.StatusCode))" -ForegroundColor Green
    }
} catch {
    Write-Host "   ❌ FAIL: Register page error: $_" -ForegroundColor Red
}

# Test 5: API Login (con le tue credenziali)
Write-Host "`n5. Testing API Login..." -ForegroundColor Yellow
try {
    $loginBody = @{
        username = $username
        password = $password
    } | ConvertTo-Json

    $session = [Microsoft.PowerShell.Commands.WebRequestSession]::new()
    $loginResponse = Invoke-WebRequest -Uri "$baseUrl/api/v1/auth/login" -Method Post `
        -Body $loginBody `
        -ContentType "application/json" `
        -SessionVariable session `
        -ErrorAction Stop

    if ($loginResponse.StatusCode -eq 200) {
        $loginData = $loginResponse.Content | ConvertFrom-Json
        Write-Host "   ✅ PASS: Login successful" -ForegroundColor Green
        Write-Host "      User: $($loginData.data.user.username)" -ForegroundColor Gray
        Write-Host "      Token: $($loginData.data.token.Substring(0, 20))..." -ForegroundColor Gray
        
        $token = $loginData.data.token
        $headers = @{
            "Authorization" = "Bearer $token"
        }

        # Test 6: Get Menus
        Write-Host "`n6. Testing Get Menus API..." -ForegroundColor Yellow
        try {
            $menus = Invoke-RestMethod -Uri "$baseUrl/api/v1/menus" -Method Get -Headers $headers
            Write-Host "   ✅ PASS: Menus API works" -ForegroundColor Green
            Write-Host "      Total menus: $($menus.data.Count)" -ForegroundColor Gray
        } catch {
            Write-Host "   ❌ FAIL: Get menus error: $_" -ForegroundColor Red
        }

        # Test 7: Get Restaurant Profile
        Write-Host "`n7. Testing Restaurant Profile API..." -ForegroundColor Yellow
        try {
            $profile = Invoke-RestMethod -Uri "$baseUrl/api/v1/restaurant/profile" -Method Get -Headers $headers
            Write-Host "   ✅ PASS: Restaurant profile API works" -ForegroundColor Green
            Write-Host "      Restaurant: $($profile.data.name)" -ForegroundColor Gray
        } catch {
            Write-Host "   ❌ FAIL: Get profile error: $_" -ForegroundColor Red
        }

        # Test 8: Create Test Menu
        Write-Host "`n8. Testing Create Menu API..." -ForegroundColor Yellow
        try {
            $newMenu = @{
                name = "Menu Test $(Get-Date -Format 'HHmmss')"
                description = "Menu di test automatico"
                categories = @(
                    @{
                        name = "Antipasti"
                        items = @(
                            @{
                                name = "Bruschetta"
                                description = "Pane tostato con pomodoro"
                                price = 5.50
                                available = $true
                            }
                        )
                    }
                    @{
                        name = "Primi"
                        items = @(
                            @{
                                name = "Spaghetti Carbonara"
                                description = "Pasta con uova, guanciale e pecorino"
                                price = 12.00
                                available = $true
                            }
                        )
                    }
                )
            } | ConvertTo-Json -Depth 10

            $createMenu = Invoke-RestMethod -Uri "$baseUrl/api/v1/menus" -Method Post `
                -Body $newMenu `
                -ContentType "application/json" `
                -Headers $headers
            
            Write-Host "   ✅ PASS: Menu created successfully" -ForegroundColor Green
            Write-Host "      Menu ID: $($createMenu.data._id)" -ForegroundColor Gray
            Write-Host "      Menu Name: $($createMenu.data.name)" -ForegroundColor Gray
            
            $menuId = $createMenu.data._id

            # Test 9: Get Single Menu
            Write-Host "`n9. Testing Get Single Menu API..." -ForegroundColor Yellow
            try {
                $menu = Invoke-RestMethod -Uri "$baseUrl/api/v1/menus/$menuId" -Method Get -Headers $headers
                Write-Host "   ✅ PASS: Get single menu works" -ForegroundColor Green
                Write-Host "      Categories: $($menu.data.categories.Count)" -ForegroundColor Gray
            } catch {
                Write-Host "   ❌ FAIL: Get single menu error: $_" -ForegroundColor Red
            }

        } catch {
            Write-Host "   ❌ FAIL: Create menu error: $_" -ForegroundColor Red
            Write-Host "      Details: $($_.Exception.Message)" -ForegroundColor Gray
        }

    }
} catch {
    Write-Host "   ❌ FAIL: Login failed - $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "      Questo è normale se MongoDB non è connesso." -ForegroundColor Gray
}

# Test 10: API Documentation
Write-Host "`n10. Testing API Documentation..." -ForegroundColor Yellow
try {
    $docs = Invoke-RestMethod -Uri "$baseUrl/api/v1/docs" -Method Get
    Write-Host "   ✅ PASS: API docs available" -ForegroundColor Green
    Write-Host "      API Version: $($docs.info.version)" -ForegroundColor Gray
} catch {
    Write-Host "   ❌ FAIL: API docs error: $_" -ForegroundColor Red
}

# Summary
Write-Host "`n=== TEST SUMMARY ===" -ForegroundColor Cyan
Write-Host "Application URL: $baseUrl" -ForegroundColor White
Write-Host "`nCritical Issues:" -ForegroundColor Yellow
Write-Host "  ⚠️  MongoDB not connected (using in-memory storage)" -ForegroundColor Red
Write-Host "     → Dati non persistenti (persi al restart)" -ForegroundColor Gray
Write-Host "     → Fix: Aggiorna MONGODB_URI in Railway Variables" -ForegroundColor Gray
Write-Host "`nNext Steps:" -ForegroundColor Yellow
Write-Host "  1. Fix MongoDB connection" -ForegroundColor White
Write-Host "  2. Verifica che i dati persistano dopo restart" -ForegroundColor White
Write-Host "  3. Testa upload immagini e generazione QR code" -ForegroundColor White

Write-Host "`n✅ Test completati!`n" -ForegroundColor Green
