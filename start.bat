@echo off
echo ========================================
echo    QR Menu System - Avvio Server
echo ========================================
echo.

REM Verifica se Go è installato
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo ERRORE: Go non è installato o non è nel PATH
    echo Scarica Go da: https://golang.org/dl/
    pause
    exit /b 1
)

echo [INFO] Controllo dipendenze...
go mod tidy

echo [INFO] Compilazione progetto...
go build -o qr-menu.exe .

if %errorlevel% neq 0 (
    echo [ERRORE] Compilazione fallita!
    pause
    exit /b 1
)

echo [INFO] Avvio del server QR Menu...
echo.
echo ========================================
echo  Server disponibile su:
echo  http://localhost:8080
echo  
echo  Interfaccia Admin:
echo  http://localhost:8080/admin
echo ========================================
echo.
echo Premi Ctrl+C per fermare il server
echo.

qr-menu.exe

pause