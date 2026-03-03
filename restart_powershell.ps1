# Script per completare setup di gcloud dopo l'installazione

Write-Host "=======================================" -ForegroundColor Cyan
Write-Host "Setup Google Cloud - Passo 1 completato" -ForegroundColor Green
Write-Host "=======================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "✅ Google Cloud SDK installato"
Write-Host ""
Write-Host "⚠️  IMPORTANTE: PowerShell deve essere riavviato per aggiornare il PATH"
Write-Host ""
Write-Host "Opzioni:"
Write-Host "1. Chiudi questa finestra PowerShell"
Write-Host "2. Apri una NUOVA finestra PowerShell"
Write-Host "3. Torna nella cartella: cd C:\Users\gigli\GoWs\qr-menu"
Write-Host "4. Verifica: gcloud --version"
Write-Host "5. Autentica: gcloud auth login"
Write-Host ""
Write-Host "Oppure esegui direttamente il prossimo script:" -ForegroundColor Yellow
Write-Host "powershell.exe .\cloud_run_deploy_step2.ps1"
Write-Host ""

# Opzionale: Chiedi se vuoi riavviare PowerShell automaticamente
$response = Read-Host "Riavviare PowerShell ora? (s/n)"
if ($response -eq "s") {
    Write-Host "Riavviando PowerShell..." -ForegroundColor Green
    Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd C:\Users\gigli\GoWs\qr-menu; .\complete_gcloud_setup.ps1"
    exit
}
