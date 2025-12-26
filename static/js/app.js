// OAuth2 Test Tool - JavaScript Application

// ============================================
// 1. COPY TO CLIPBOARD FUNCTIONALITY
// ============================================
function copyToken(button) {
    const container = button.closest('.token-container');
    const textarea = container.querySelector('textarea, .token-field');

    // Select and copy
    textarea.select();
    textarea.setSelectionRange(0, 99999); // For mobile devices

    try {
        document.execCommand('copy');

        // Visual feedback
        const originalText = button.innerHTML;
        button.innerHTML = '✓ Copiado!';
        button.classList.add('copied');

        // Show toast notification
        showToast('Token copiado com sucesso!', 'success');

        setTimeout(() => {
            button.innerHTML = originalText;
            button.classList.remove('copied');
        }, 2000);
    } catch (err) {
        showToast('Erro ao copiar token', 'error');
    }

    // Deselect
    window.getSelection().removeAllRanges();
}

// ============================================
// 2. FORM VALIDATION
// ============================================
function validateForm(form) {
    let isValid = true;

    // Client ID
    const clientId = form.querySelector('#client_id');
    if (clientId && !clientId.value.trim()) {
        showError(clientId, 'Client ID é obrigatório');
        isValid = false;
    } else if (clientId) {
        clearError(clientId);
    }

    // Client Secret
    const clientSecret = form.querySelector('#client_secret');
    if (clientSecret && !clientSecret.value.trim()) {
        showError(clientSecret, 'Client Secret é obrigatório');
        isValid = false;
    } else if (clientSecret) {
        clearError(clientSecret);
    }

    // Redirect URI
    const redirectUri = form.querySelector('#redirect_uri');
    if (redirectUri && !isValidUrl(redirectUri.value)) {
        showError(redirectUri, 'URL inválida. Use formato: http://localhost:8080/callback');
        isValid = false;
    } else if (redirectUri) {
        clearError(redirectUri);
    }

    return isValid;
}

function showError(input, message) {
    const formGroup = input.closest('.form-group');
    if (!formGroup) return;

    formGroup.classList.add('error');

    let errorMsg = formGroup.querySelector('.error-message');
    if (!errorMsg) {
        errorMsg = document.createElement('div');
        errorMsg.className = 'error-message';
        input.parentNode.insertBefore(errorMsg, input.nextSibling);
    }
    errorMsg.textContent = message;
}

function clearError(input) {
    const formGroup = input.closest('.form-group');
    if (!formGroup) return;

    formGroup.classList.remove('error');

    const errorMsg = formGroup.querySelector('.error-message');
    if (errorMsg) {
        errorMsg.remove();
    }
}

function isValidUrl(string) {
    try {
        const url = new URL(string);
        return url.protocol === 'http:' || url.protocol === 'https:';
    } catch (_) {
        return false;
    }
}

// ============================================
// 3. TABLE SEARCH FUNCTIONALITY
// ============================================
function setupTableSearch() {
    const searchInput = document.querySelector('.table-search');
    if (!searchInput) return;

    const table = document.querySelector('.history-table');
    if (!table) return;

    const rows = table.querySelectorAll('tbody tr');

    searchInput.addEventListener('input', (e) => {
        const searchTerm = e.target.value.toLowerCase();

        rows.forEach(row => {
            const text = row.textContent.toLowerCase();
            row.style.display = text.includes(searchTerm) ? '' : 'none';
        });

        // Show "no results" message if needed
        const visibleRows = Array.from(rows).filter(row => row.style.display !== 'none');
        const tbody = table.querySelector('tbody');

        let noResultsRow = tbody.querySelector('.no-results-row');

        if (visibleRows.length === 0) {
            if (!noResultsRow) {
                noResultsRow = document.createElement('tr');
                noResultsRow.className = 'no-results-row';
                noResultsRow.innerHTML = '<td colspan="100%" style="text-align: center; padding: 2rem; color: #6c757d;">Nenhuma requisição encontrada</td>';
                tbody.appendChild(noResultsRow);
            }
        } else {
            if (noResultsRow) {
                noResultsRow.remove();
            }
        }
    });
}

// ============================================
// 4. TOAST NOTIFICATIONS
// ============================================
let toastContainer = null;
let toastIdCounter = 0;

function initToastContainer() {
    if (!toastContainer) {
        toastContainer = document.createElement('div');
        toastContainer.className = 'toast-container';
        document.body.appendChild(toastContainer);
    }
}

function showToast(message, type = 'info', duration = 3000) {
    initToastContainer();

    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    toast.id = `toast-${toastIdCounter++}`;

    const icon = type === 'success' ? '✓' : type === 'error' ? '✗' : 'ℹ';

    toast.innerHTML = `
        <div style="font-size: 1.25rem;">${icon}</div>
        <div style="flex: 1;">
            <div style="font-weight: 600; margin-bottom: 0.25rem;">${type === 'success' ? 'Sucesso' : type === 'error' ? 'Erro' : 'Info'}</div>
            <div style="font-size: 0.875rem;">${message}</div>
        </div>
        <button class="toast-close" onclick="closeToast('${toast.id}')">×</button>
    `;

    toastContainer.appendChild(toast);

    // Auto remove after duration
    setTimeout(() => {
        closeToast(toast.id);
    }, duration);
}

function closeToast(toastId) {
    const toast = document.getElementById(toastId);
    if (toast) {
        toast.style.animation = 'slideIn 0.3s ease reverse';
        setTimeout(() => {
            toast.remove();
        }, 300);
    }
}

// ============================================
// 5. DARK MODE TOGGLE
// ============================================
function initDarkMode() {
    // Check saved preference or system preference
    const savedMode = localStorage.getItem('darkMode');
    const systemPrefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;

    if (savedMode === 'true' || (savedMode === null && systemPrefersDark)) {
        document.body.classList.add('dark-mode');
    }

    // Create toggle button
    const toggleBtn = document.createElement('button');
    toggleBtn.className = 'dark-mode-toggle';
    toggleBtn.innerHTML = '🌙';
    toggleBtn.setAttribute('aria-label', 'Toggle dark mode');
    toggleBtn.onclick = toggleDarkMode;

    document.body.appendChild(toggleBtn);

    updateDarkModeButton();
}

function toggleDarkMode() {
    document.body.classList.toggle('dark-mode');
    const isDark = document.body.classList.contains('dark-mode');
    localStorage.setItem('darkMode', isDark);
    updateDarkModeButton();
}

function updateDarkModeButton() {
    const btn = document.querySelector('.dark-mode-toggle');
    if (btn) {
        btn.innerHTML = document.body.classList.contains('dark-mode') ? '☀️' : '🌙';
    }
}

// ============================================
// 6. ACTIVE NAVIGATION
// ============================================
function setActiveNav() {
    const currentPath = window.location.pathname;
    const navLinks = document.querySelectorAll('.nav-links a');

    navLinks.forEach(link => {
        link.classList.remove('active');

        const linkPath = new URL(link.href).pathname;

        if (currentPath === linkPath ||
            (currentPath.startsWith(linkPath) && linkPath !== '/')) {
            link.classList.add('active');
        }
    });
}

// ============================================
// 7. SYNTAX HIGHLIGHTING
// ============================================
function setupSyntaxHighlighting() {
    if (typeof hljs !== 'undefined') {
        document.querySelectorAll('pre code, .code-block').forEach((block) => {
            // Try to parse as JSON and format
            const text = block.textContent.trim();
            if (text.startsWith('{') || text.startsWith('[')) {
                try {
                    const json = JSON.parse(text);
                    block.textContent = JSON.stringify(json, null, 2);
                } catch (e) {
                    // Not valid JSON, keep as is
                }
            }

            hljs.highlightElement(block);
        });
    }
}

// ============================================
// 8. INITIALIZE ON PAGE LOAD
// ============================================
document.addEventListener('DOMContentLoaded', () => {
    // Set active navigation
    setActiveNav();

    // Form validation
    const form = document.querySelector('form[hx-post="/config"]');
    if (form) {
        // Real-time validation on blur
        ['#client_id', '#client_secret', '#redirect_uri'].forEach(selector => {
            const input = form.querySelector(selector);
            if (input) {
                input.addEventListener('blur', () => {
                    validateForm(form);
                });
            }
        });

        // Validate on submit (HTMX will handle the actual submit)
        form.addEventListener('submit', (e) => {
            if (!validateForm(form)) {
                e.preventDefault();
                showToast('Por favor, corrija os erros no formulário', 'error');
            }
        });
    }

    // Table search
    setupTableSearch();

    // Syntax highlighting
    setupSyntaxHighlighting();

    // Dark mode
    initDarkMode();

    // Initialize toast container
    initToastContainer();
});

// ============================================
// 9. HTMX EVENT HANDLERS
// ============================================
document.body.addEventListener('htmx:afterSwap', (event) => {
    // Re-run syntax highlighting after HTMX swap
    setupSyntaxHighlighting();

    // Re-apply active navigation
    setActiveNav();

    // Re-setup table search if needed
    setupTableSearch();
});

document.body.addEventListener('htmx:beforeRequest', (event) => {
    // You can add custom loading logic here if needed
    console.log('HTMX request starting...');
});

document.body.addEventListener('htmx:afterRequest', (event) => {
    // Check for success/error and show toast
    if (event.detail.successful) {
        // Only show success toast for specific endpoints
        const xhr = event.detail.xhr;
        if (xhr.status === 200 && event.detail.pathInfo.requestPath === '/config') {
            showToast('Configuração salva com sucesso!', 'success');
        }
    } else {
        showToast('Erro na requisição. Tente novamente.', 'error');
    }
});

// ============================================
// 10. UTILITY FUNCTIONS
// ============================================

// Add animation class to elements when they come into view
const observeAnimations = () => {
    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.classList.add('animate-in');
            }
        });
    }, { threshold: 0.1 });

    document.querySelectorAll('.card').forEach(card => {
        observer.observe(card);
    });
};

// Call after DOM loaded
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', observeAnimations);
} else {
    observeAnimations();
}
