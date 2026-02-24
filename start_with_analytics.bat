@echo off
REM Script per avviare il server QR Menu con Analytics Dashboard pronto al test

color 0A
echo.
echo ╔═══════════════════════════════════════════════════════════════╗
echo ║                                                               ║
echo ║        🍽️  QR MENU ENTERPRISE - ANALYTICS DASHBOARD  🍽️       ║
echo ║                                                               ║
echo ╚═══════════════════════════════════════════════════════════════╝
echo.

REM Verifica se il build esiste
if not exist qr-menu.exe (
    echo ⚠️  qr-menu.exe non trovato. Compilazione in corso...
    echo.
    go build -o qr-menu.exe .
    if errorlevel 1 (
        echo ❌ Errore nella compilazione
        pause
        exit /b 1
    )
)

echo ✅ Build confirmed
echo.

REM Crea le directory necessarie
if not exist logs mkdir logs
if not exist analytics mkdir analytics
if not exist static\qrcodes mkdir static\qrcodes
if not exist templates mkdir templates
if not exist storage mkdir storage

echo ✅ Directories created
echo.

REM Visualizza i parametri di avvio
echo 📋 Configurazione Server:
echo.
echo   Server:              http://localhost:8080
echo   Admin:               http://localhost:8080/admin
echo   Analytics:           http://localhost:8080/admin/analytics
echo   API Docs:            http://localhost:8080/api/v1/docs
echo   REST API:            http://localhost:8080/api/v1
echo   Health Check:        http://localhost:8080/api/v1/health
echo.

REM Avvia il server
echo 🚀 Avvio server in corso...
echo.

qr-menu.exe

REM Se il server si ferma, mostra un messaggio
echo.
echo ⛔ Server stoppato
echo.
pause
