$baseUrl = "https://qr-menu-staging.up.railway.app"
$loginBody = @{username = "trattoria_roma"; password = "RomaTest2026!"} | ConvertTo-Json
$loginResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
$token = $loginResp.data.token
$headers = @{"Authorization" = "Bearer $token"}

Write-Host "📍 Profilo ristorante:" -ForegroundColor Cyan
$profile = Invoke-RestMethod -Uri "$baseUrl/api/v1/restaurant/profile" -Headers $headers
$profile.data | ConvertTo-Json -Depth 5

Write-Host ""
Write-Host "🔍 Restaurant ID: $($profile.data.id)" -ForegroundColor Yellow
Write-Host "Active Menu ID: $($profile.data.active_menu_id)" -ForegroundColor Yellow
