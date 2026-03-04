$baseUrl = "https://qr-menu-staging.up.railway.app"

Write-Host "⏳ Aspettando deploy Railway..." -ForegroundColor Cyan
Start-Sleep -Seconds 30

Write-Host "🔍 Test endpoint di debug..." -ForegroundColor Cyan
Write-Host ""

# Login
$loginBody = @{username = "trattoria_roma"; password = "RomaTest2026!"} | ConvertTo-Json
$loginResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
$token = $loginResp.data.token
$headers = @{"Authorization" = "Bearer $token"}

Write-Host "✅ Login OK" -ForegroundColor Green
Write-Host ""

# Chiama endpoint di debug
Write-Host "📋 Chiamando /api/v1/debug/menus..." -ForegroundColor Yellow
try {
    $debugResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/debug/menus" -Headers $headers
    
    Write-Host ""
    Write-Host "=" * 70 -ForegroundColor Cyan
    Write-Host "DEBUG INFO" -ForegroundColor Cyan
    Write-Host "=" * 70 -ForegroundColor Cyan
    Write-Host ""
    
    $data = $debugResp.data
    
    Write-Host "Restaurant ID: $($data.restaurant_id)" -ForegroundColor White
    Write-Host "Total menus in DB: $($data.total_menus_in_db)" -ForegroundColor Yellow
    Write-Host "Filtered menus count: $($data.filtered_menus_count)" -ForegroundColor Yellow
    Write-Host ""
    
    if ($data.total_menus_in_db -gt 0) {
        Write-Host "📄 Menu nel database (RAW):" -ForegroundColor Cyan
        Write-Host ""
        $data.all_menus_raw | ForEach-Object {
            Write-Host "  Menu: $($_['name'])" -ForegroundColor White
            Write-Host "    ID: $($_['id'])" -ForegroundColor Gray
            Write-Host "    Restaurant ID: $($_['restaurant_id'])" -ForegroundColor Gray
            
            # Verifica se il restaurant_id corrisponde
            if ($_['restaurant_id'] -eq $data.restaurant_id) {
                Write-Host "    ✅ Restaurant ID CORRISPONDE" -ForegroundColor Green
            } else {
                Write-Host "    ❌ Restaurant ID NON CORRISPONDE!" -ForegroundColor Red
                Write-Host "       Expected: $($data.restaurant_id)" -ForegroundColor Yellow
                Write-Host "       Got:      $($_['restaurant_id'])" -ForegroundColor Yellow
            }
            
            # Mostra tutti i campi
            Write-Host "    Campi presenti nel documento:" -ForegroundColor Gray
            $_.Keys | ForEach-Object {
                Write-Host "      - $_" -ForegroundColor DarkGray
            }
            Write-Host ""
        }
    } else {
        Write-Host "❌ Nessun menu nel database!" -ForegroundColor Red
        Write-Host "   Devi creare un menu prima." -ForegroundColor Yellow
    }
    
    Write-Host "=" * 70 -ForegroundColor Cyan
    
    if ($data.filtered_menus_count -gt 0) {
        Write-Host ""
        Write-Host "✅ LA QUERY FUNZIONA! Menu filtrati trovati: $($data.filtered_menus_count)" -ForegroundColor Green
    } elseif ($data.total_menus_in_db -gt 0) {
        Write-Host ""
        Write-Host "❌ PROBLEMA CONFERMATO: La query filtrata non funziona" -ForegroundColor Red
        Write-Host "   Ci sono menu nel DB ma la query per restaurant_id non li trova" -ForegroundColor Yellow
    } else {
        Write-Host ""
        Write-Host "⚠️  Non ci sono menu nel database per testare" -ForegroundColor Yellow
        Write-Host "   Creo un menu di test..." -ForegroundColor Cyan
        
        $menuData = @{
            name = "Menu Debug $(Get-Date -Format 'HH:mm:ss')"
            description = "Test"
            meal_type = "lunch"
            categories = @(
                @{
                    id = (New-Guid).ToString()
                    name = "Test"
                    description = "Test"
                    items = @(
                        @{
                            id = (New-Guid).ToString()
                            name = "Piatto Test"
                            description = "Test"
                            price = 10.0
                            category = "Test"
                            available = $true
                        }
                    )
                }
            )
        } | ConvertTo-Json -Depth 10
        
        $createResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/menus" -Method Post -Body $menuData -Headers $headers -ContentType "application/json"
        Write-Host "   ✅ Menu creato: $($createResp.data.name)" -ForegroundColor Green
        Write-Host ""
        Write-Host "   Riprova ora l'endpoint di debug:" -ForegroundColor Cyan
        Write-Host "   .\test_debug_endpoint.ps1" -ForegroundColor Gray
    }
    
} catch {
    Write-Host "❌ Errore: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.ErrorDetails.Message) {
        Write-Host $_.ErrorDetails.Message -ForegroundColor Yellow
    }
}

Write-Host ""
