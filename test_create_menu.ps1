$baseUrl = "https://qr-menu-staging.up.railway.app"
$loginBody = @{username = "trattoria_roma"; password = "RomaTest2026!"} | ConvertTo-Json
$loginResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
$token = $loginResp.data.token
$headers = @{"Authorization" = "Bearer $token"; "Content-Type" = "application/json"}

Write-Host "🍝 Creazione menu di test..." -ForegroundColor Cyan

$menuData = @{
    name = "Menu Test BSON"
    description = "Menu per verificare i tag bson"
    meal_type = "lunch"
    categories = @(
        @{
            id = (New-Guid).ToString()
            name = "Primi Piatti"
            description = "Pasta fresca"
            items = @(
                @{
                    id = (New-Guid).ToString()
                    name = "Carbonara"
                    description = "Classica carbonara romana"
                    price = 12.50
                    category = "Primi Piatti"
                    available = $true
                },
                @{
                    id = (New-Guid).ToString()
                    name = "Amatriciana"
                    description = "Pasta con guanciale e pomodoro"
                    price = 11.50
                    category = "Primi Piatti"
                    available = $true
                }
            )
        }
    )
} | ConvertTo-Json -Depth 10

try {
    Write-Host "Invio richiesta POST..." -ForegroundColor Gray
    $createResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/menus" -Method Post -Body $menuData -Headers $headers
    Write-Host "✅ Menu creato con successo!" -ForegroundColor Green
    Write-Host "   ID: $($createResp.data.id)" -ForegroundColor White
    Write-Host "   Nome: $($createResp.data.name)" -ForegroundColor White
    
    Write-Host ""
    Write-Host "🔍 Verifica immediata: recupero menu..." -ForegroundColor Cyan
    $menusResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/menus" -Headers $headers
    Write-Host "Menu trovati: $($menusResp.data.Count)" -ForegroundColor Yellow
    
    if ($menusResp.data.Count -gt 0) {
        Write-Host "✅ MENU VISIBILE!" -ForegroundColor Green
        $menusResp.data | ForEach-Object {
            Write-Host "   - $($_.name)" -ForegroundColor White
            Write-Host "     ID: $($_.id)" -ForegroundColor Gray
            Write-Host "     Restaurant ID: $($_.restaurant_id)" -ForegroundColor Gray
            Write-Host "     Categorie: $($_.categories.Count)" -ForegroundColor Gray
        }
    } else {
        Write-Host "❌ Menu creato ma non visibile in GET!" -ForegroundColor Red
    }
} catch {
    Write-Host "❌ Errore nella creazione: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.ErrorDetails.Message) {
        Write-Host $_.ErrorDetails.Message -ForegroundColor Yellow
    }
}
