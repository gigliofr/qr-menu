// QR Menu - JavaScript Utility Functions

// Utility per le chiamate API
class MenuAPI {
    static async get(url) {
        try {
            const response = await fetch(url);
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            return await response.json();
        } catch (error) {
            console.error('Errore GET:', error);
            throw error;
        }
    }

    static async post(url, data) {
        try {
            const response = await fetch(url, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(data)
            });
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            return await response.json();
        } catch (error) {
            console.error('Errore POST:', error);
            throw error;
        }
    }
}

// Gestione dinamica form creazione menu
class MenuFormManager {
    constructor(formId) {
        this.form = document.getElementById(formId);
        this.categoryCount = 1;
        this.setupEventListeners();
    }

    setupEventListeners() {
        if (!this.form) return;

        // Event listener per aggiungere categoria
        const addCategoryBtn = document.getElementById('add-category-btn');
        if (addCategoryBtn) {
            addCategoryBtn.addEventListener('click', () => this.addCategory());
        }

        // Event listener per validazione form
        this.form.addEventListener('submit', (e) => this.validateForm(e));
    }

    addCategory() {
        this.categoryCount++;
        const container = document.getElementById('categories-container');
        if (!container) return;

        const categoryHtml = this.createCategoryHTML(this.categoryCount);
        container.insertAdjacentHTML('beforeend', categoryHtml);
        
        // Aggiunge eventi per la nuova categoria
        this.setupCategoryEvents(this.categoryCount);
    }

    createCategoryHTML(categoryIndex) {
        return `
            <div class="category-section" data-category="${categoryIndex}">
                <div class="category-header">
                    <h4>Categoria ${categoryIndex} 
                        <button type="button" class="remove-btn" onclick="menuFormManager.removeCategory(this)">✗ Rimuovi</button>
                    </h4>
                </div>
                <div class="form-group">
                    <label>Nome Categoria:</label>
                    <input type="text" name="category_name[]" placeholder="es: Primi Piatti" required>
                </div>
                <div class="form-group">
                    <label>Descrizione Categoria:</label>
                    <input type="text" name="category_description[]" placeholder="es: I nostri primi piatti della tradizione">
                </div>
                <h5>🍽️ Piatti di questa categoria:</h5>
                <div class="items-container" data-category="${categoryIndex}">
                    <div class="item-row">
                        <input type="text" placeholder="Nome piatto" name="item_name_${categoryIndex}[]">
                        <input type="text" placeholder="Descrizione" name="item_description_${categoryIndex}[]">
                        <input type="number" step="0.01" placeholder="Prezzo" name="item_price_${categoryIndex}[]">
                        <button type="button" class="remove-btn" onclick="menuFormManager.removeItem(this)">✗</button>
                    </div>
                </div>
                <button type="button" class="btn btn-secondary" onclick="menuFormManager.addItem(${categoryIndex})">➕ Aggiungi Piatto</button>
            </div>
        `;
    }

    setupCategoryEvents(categoryIndex) {
        // Aggiunge eventi per i nuovi elementi creati dinamicamente
        const categorySection = document.querySelector(`[data-category="${categoryIndex}"]`);
        if (!categorySection) return;

        // Animazione di entrata
        categorySection.classList.add('fade-in');
    }

    removeCategory(button) {
        if (confirm('Sicuro di voler rimuovere questa categoria?')) {
            button.closest('.category-section').remove();
        }
    }

    addItem(categoryIndex) {
        const itemsContainer = document.querySelector(`[data-category="${categoryIndex}"] .items-container`);
        if (!itemsContainer) return;

        const itemHtml = `
            <div class="item-row">
                <input type="text" placeholder="Nome piatto" name="item_name_${categoryIndex}[]">
                <input type="text" placeholder="Descrizione" name="item_description_${categoryIndex}[]">
                <input type="number" step="0.01" placeholder="Prezzo" name="item_price_${categoryIndex}[]">
                <button type="button" class="remove-btn" onclick="menuFormManager.removeItem(this)">✗</button>
            </div>
        `;
        itemsContainer.insertAdjacentHTML('beforeend', itemHtml);
    }

    removeItem(button) {
        button.closest('.item-row').remove();
    }

    validateForm(e) {
        const categories = document.querySelectorAll('input[name="category_name[]"]');
        let hasValidCategory = false;
        
        categories.forEach(cat => {
            if (cat.value.trim()) {
                hasValidCategory = true;
            }
        });
        
        if (!hasValidCategory) {
            e.preventDefault();
            this.showAlert('Inserisci almeno una categoria con un nome valido.', 'danger');
            return false;
        }

        return true;
    }

    showAlert(message, type = 'info') {
        // Crea e mostra un alert temporaneo
        const alertDiv = document.createElement('div');
        alertDiv.className = `alert alert-${type}`;
        alertDiv.textContent = message;
        
        // Inserisce l'alert in cima alla pagina
        document.body.insertBefore(alertDiv, document.body.firstChild);
        
        // Rimuove l'alert dopo 5 secondi
        setTimeout(() => {
            alertDiv.remove();
        }, 5000);
    }
}

// Gestione QR Code
class QRCodeManager {
    static async generateQR(menuId) {
        try {
            const response = await MenuAPI.post(`/api/menu/${menuId}/generate-qr`);
            return response;
        } catch (error) {
            console.error('Errore nella generazione QR code:', error);
            throw error;
        }
    }

    static displayQRCode(qrCodeUrl, menuUrl, containerId) {
        const container = document.getElementById(containerId);
        if (!container) return;

        container.innerHTML = `
            <div class="qr-display">
                <h3>QR Code generato con successo!</h3>
                <img src="${qrCodeUrl}" alt="QR Code Menu">
                <p>URL Menu: <a href="${menuUrl}" target="_blank">${menuUrl}</a></p>
                <p>I clienti possono scansionare questo QR code per visualizzare il menu.</p>
            </div>
        `;
    }
}

// Utility per il localStorage
class StorageManager {
    static set(key, value) {
        try {
            localStorage.setItem(key, JSON.stringify(value));
        } catch (error) {
            console.error('Errore nel salvataggio localStorage:', error);
        }
    }

    static get(key, defaultValue = null) {
        try {
            const item = localStorage.getItem(key);
            return item ? JSON.parse(item) : defaultValue;
        } catch (error) {
            console.error('Errore nella lettura localStorage:', error);
            return defaultValue;
        }
    }

    static remove(key) {
        try {
            localStorage.removeItem(key);
        } catch (error) {
            console.error('Errore nella rimozione localStorage:', error);
        }
    }
}

// Animazioni e effetti UI
class UIEffects {
    static fadeIn(element, duration = 300) {
        element.style.opacity = '0';
        element.style.display = 'block';
        
        let start = performance.now();
        
        function animate(currentTime) {
            const elapsed = currentTime - start;
            const progress = Math.min(elapsed / duration, 1);
            
            element.style.opacity = progress;
            
            if (progress < 1) {
                requestAnimationFrame(animate);
            }
        }
        
        requestAnimationFrame(animate);
    }

    static slideDown(element, duration = 300) {
        element.style.height = '0';
        element.style.overflow = 'hidden';
        element.style.display = 'block';
        
        const targetHeight = element.scrollHeight;
        let start = performance.now();
        
        function animate(currentTime) {
            const elapsed = currentTime - start;
            const progress = Math.min(elapsed / duration, 1);
            
            element.style.height = (targetHeight * progress) + 'px';
            
            if (progress >= 1) {
                element.style.height = 'auto';
                element.style.overflow = 'visible';
            } else {
                requestAnimationFrame(animate);
            }
        }
        
        requestAnimationFrame(animate);
    }

    static addLoadingSpinner(element) {
        const spinner = document.createElement('div');
        spinner.className = 'spinner';
        spinner.id = 'loading-spinner';
        element.appendChild(spinner);
    }

    static removeLoadingSpinner() {
        const spinner = document.getElementById('loading-spinner');
        if (spinner) {
            spinner.remove();
        }
    }
}

// Validazione form
class FormValidator {
    static validateEmail(email) {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return emailRegex.test(email);
    }

    static validatePrice(price) {
        return !isNaN(price) && price > 0;
    }

    static validateRequired(value) {
        return value && value.trim().length > 0;
    }

    static showFieldError(fieldElement, message) {
        // Rimuove errori precedenti
        this.clearFieldError(fieldElement);
        
        // Aggiunge classe di errore
        fieldElement.classList.add('error');
        
        // Crea messaggio di errore
        const errorDiv = document.createElement('div');
        errorDiv.className = 'field-error';
        errorDiv.textContent = message;
        
        // Inserisce dopo il campo
        fieldElement.parentNode.insertBefore(errorDiv, fieldElement.nextSibling);
    }

    static clearFieldError(fieldElement) {
        fieldElement.classList.remove('error');
        const errorDiv = fieldElement.parentNode.querySelector('.field-error');
        if (errorDiv) {
            errorDiv.remove();
        }
    }
}

// Inizializzazione globale
let menuFormManager;

document.addEventListener('DOMContentLoaded', function() {
    // Inizializza il manager del form se presente
    if (document.getElementById('menuForm')) {
        menuFormManager = new MenuFormManager('menuForm');
    }

    // Aggiunge animazioni di caricamento
    const cards = document.querySelectorAll('.card, .menu-card');
    cards.forEach((card, index) => {
        setTimeout(() => {
            card.classList.add('fade-in');
        }, index * 100);
    });

    // Gestione click sui menu items per effetti
    const menuItems = document.querySelectorAll('.menu-item');
    menuItems.forEach(item => {
        item.addEventListener('click', function() {
            this.style.transform = 'scale(0.98)';
            setTimeout(() => {
                this.style.transform = 'scale(1)';
            }, 150);
        });
    });

    // Salva timestamp dell ultima visita
    StorageManager.set('lastVisit', new Date().toISOString());
    
    console.log('QR Menu System inizializzato');
});

// Esporta per uso globale
window.MenuAPI = MenuAPI;
window.QRCodeManager = QRCodeManager;
window.StorageManager = StorageManager;
window.UIEffects = UIEffects;
window.FormValidator = FormValidator;