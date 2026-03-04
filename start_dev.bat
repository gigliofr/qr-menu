@echo off
REM ================================================
REM Script di Avvio Sviluppo - QR Menu
REM ================================================

echo.
echo === Configurazione MongoDB ===
echo.

REM Configurazione MongoDB Atlas
set MONGODB_URI=mongodb+srv://ac-d8zdak4.b9jfwmr.mongodb.net/?authMechanism=MONGODB-X509^&authSource=$external^&retryWrites=true^&w=majority
set MONGODB_CERT_PATH=C:\Users\gigli\Desktop\X509-cert-4084673564018728353.pem
set MONGODB_DB_NAME=qr-menu

echo [OK] MONGODB_URI configurato
echo [OK] MONGODB_CERT_PATH configurato
echo [OK] MONGODB_DB_NAME: %MONGODB_DB_NAME%
echo.

REM Verifica certificato
if exist "%MONGODB_CERT_PATH%" (
    echo [OK] Certificato trovato
) else (
    echo [ERRORE] Certificato NON trovato: %MONGODB_CERT_PATH%
    pause
    exit /b 1
)

echo.
echo ================================================
echo Credenziali Admin:
echo   Username: admin
echo   Password: admin123
echo ================================================
echo.
echo URL: http://localhost:8080/login
echo.
echo Avvio applicazione...
echo.

REM Avvia applicazione
qr-menu.exe

pause
