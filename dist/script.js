/* ── theme ── */
const THEME_KEY = 'theme';

function applyTheme(theme) {
    document.documentElement.setAttribute('data-theme', theme);
    localStorage.setItem(THEME_KEY, theme);
    const btn = document.getElementById('theme-toggle');
    if (btn) btn.textContent = theme === 'light' ? '☾ Dark' : '☀ Light';
}

function toggleTheme() {
    const cur = document.documentElement.getAttribute('data-theme') || 'dark';
    applyTheme(cur === 'dark' ? 'light' : 'dark');
}

/* ── api base ── */
const API_BASE = window.location.origin.includes('localhost') && !window.location.port.includes('3000')
    ? 'http://localhost:3000'
    : '';

let uploadedHTML = '';
let scrapeMode = false; // true = URL scrape mode, false = file upload mode
const DOWNLOAD_STORAGE_KEY = 'downloadSettings';
const DOWNLOAD_DEFAULTS = {
    useFilePicker: true,
    formatName: 'formatted.html',
    zipName: 'extracted.zip',
    tsxName: 'project.zip',
    ejsName: 'project-ejs.zip',
};

const PICKER_TYPES = {
    html: [{ description: 'HTML File', accept: { 'text/html': ['.html'] } }],
    zip: [{ description: 'ZIP Archive', accept: { 'application/zip': ['.zip'] } }],
};

function supportsFilePicker() {
    return 'showSaveFilePicker' in window;
}

function sanitizeFilename(name, fallback) {
    const trimmed = (name || '').trim();
    if (!trimmed) {
        return fallback;
    }
    return trimmed.replace(/[\\\/<>:"|?*]+/g, '-');
}

function ensureExtension(name, extension) {
    if (!extension) {
        return name;
    }
    const lowerName = name.toLowerCase();
    const lowerExt = extension.toLowerCase();
    if (lowerName.endsWith(lowerExt)) {
        return name;
    }
    return `${name}${extension}`;
}

function resolveDownloadName(inputId, fallbackName, extension) {
    const input = document.getElementById(inputId);
    const rawName = input ? input.value : '';
    const sanitized = sanitizeFilename(rawName, fallbackName);
    return ensureExtension(sanitized, extension);
}

function loadDownloadSettings() {
    try {
        const stored = JSON.parse(localStorage.getItem(DOWNLOAD_STORAGE_KEY) || '{}');
        return { ...DOWNLOAD_DEFAULTS, ...stored };
    } catch (error) {
        return { ...DOWNLOAD_DEFAULTS };
    }
}

function storeDownloadSettings() {
    const settings = {
        useFilePicker: document.getElementById('download-picker-toggle')?.checked ?? false,
        formatName: document.getElementById('download-name-format')?.value ?? '',
        zipName: document.getElementById('download-name-zip')?.value ?? '',
        tsxName: document.getElementById('download-name-tsx')?.value ?? '',
        ejsName: document.getElementById('download-name-ejs')?.value ?? '',
    };
    try {
        localStorage.setItem(DOWNLOAD_STORAGE_KEY, JSON.stringify(settings));
    } catch (error) {
        console.warn('Failed to store download settings.', error);
    }
}

function initializeDownloadSettings() {
    const settings = loadDownloadSettings();
    const toggle = document.getElementById('download-picker-toggle');
    const hint = document.getElementById('download-picker-hint');
    const supportsPicker = supportsFilePicker();

    if (toggle) {
        toggle.checked = supportsPicker ? settings.useFilePicker : false;
        toggle.disabled = !supportsPicker;
        const toggleWrapper = toggle.closest('.download-toggle');
        if (toggleWrapper) {
            toggleWrapper.classList.toggle('disabled', !supportsPicker);
        }
    }

    if (hint) {
        hint.textContent = supportsPicker
            ? 'Use the save dialog to choose a location and edit the final name.'
            : 'Save dialog not supported in this browser. Use browser settings to choose a location.';
    }

    const fields = [
        { id: 'download-name-format', value: settings.formatName },
        { id: 'download-name-zip', value: settings.zipName },
        { id: 'download-name-tsx', value: settings.tsxName },
        { id: 'download-name-ejs', value: settings.ejsName },
    ];

    fields.forEach((field) => {
        const input = document.getElementById(field.id);
        if (input) {
            input.value = field.value;
        }
    });

    if (toggle) {
        toggle.addEventListener('change', storeDownloadSettings);
    }

    const inputs = document.querySelectorAll('.download-settings input');
    inputs.forEach((input) => {
        if (input.id !== 'download-picker-toggle') {
            input.addEventListener('input', storeDownloadSettings);
        }
    });
}

function shouldUseFilePicker() {
    const toggle = document.getElementById('download-picker-toggle');
    return !!(toggle && toggle.checked && supportsFilePicker());
}

function triggerBrowserDownload(blob, filename) {
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    a.click();
    URL.revokeObjectURL(url);
}

async function saveBlobWithPicker(blob, filename, pickerTypes) {
    if (!supportsFilePicker()) {
        return 'fallback';
    }

    try {
        const handle = await window.showSaveFilePicker({
            suggestedName: filename,
            types: pickerTypes,
            excludeAcceptAllOption: false,
        });
        const writable = await handle.createWritable();
        await writable.write(blob);
        await writable.close();
        return 'saved';
    } catch (error) {
        if (error && error.name === 'AbortError') {
            showToast('Save canceled', 'info');
            return 'canceled';
        }
        showToast('Save dialog failed, using browser download', 'error');
        return 'fallback';
    }
}

async function downloadBlob(blob, filename, pickerTypes) {
    if (shouldUseFilePicker()) {
        const result = await saveBlobWithPicker(blob, filename, pickerTypes);
        if (result !== 'fallback') {
            return result;
        }
    }

    triggerBrowserDownload(blob, filename);
    return 'downloaded';
}

function changeUploadButtonToFinish() {
    const uploadButton = document.querySelector('.button.upload');
    if (uploadButton) {
        const btn = uploadButton.querySelector('button');
        if (btn) {
            btn.textContent = 'Finish Operation';
        }
    }
}

function resetToInitialState() {
    const uploadButton = document.querySelector('.button.upload');
    const actionButtons = [
        document.querySelector('.button.first'),
        document.querySelector('.button.sec'),
        document.querySelector('.button.third'),
        document.querySelector('.button.fourth')
    ];

    if (uploadButton) {
        const btn = uploadButton.querySelector('button');
        if (btn) {
            btn.textContent = 'Upload';
        }
    }

    actionButtons.forEach(button => {
        if (button) {
            button.classList.remove('button-visible');
        }
    });

    uploadedHTML = '';

    // Also reset scrape mode URL indicator
    const scrapeArea = document.getElementById('scrape-input-area');
    if (scrapeArea) {
        scrapeArea.classList.remove('has-url');
        const scrapeBtn = scrapeArea.querySelector('#scrape-action-btn button');
        if (scrapeBtn) scrapeBtn.textContent = 'Scrape';
    }
    const urlInput = document.getElementById('scrape-url-input');
    if (urlInput) urlInput.value = '';
}

/* ── input mode toggle ── */
function setInputMode(mode) {
    scrapeMode = mode === 'scrape';
    const uploadBtn = document.getElementById('mode-upload-btn');
    const scrapeBtn = document.getElementById('mode-scrape-btn');
    const uploadButton = document.querySelector('.button.upload');
    const scrapeArea = document.getElementById('scrape-input-area');

    if (uploadBtn) uploadBtn.classList.toggle('mode-btn-active', !scrapeMode);
    if (scrapeBtn) scrapeBtn.classList.toggle('mode-btn-active', scrapeMode);
    if (uploadButton) uploadButton.style.display = scrapeMode ? 'none' : '';
    if (scrapeArea) scrapeArea.style.display = scrapeMode ? 'flex' : 'none';

    // Reset state when switching modes
    uploadedHTML = '';
    const actionButtons = [
        document.querySelector('.button.first'),
        document.querySelector('.button.sec'),
        document.querySelector('.button.third'),
        document.querySelector('.button.fourth')
    ];
    actionButtons.forEach(b => b && b.classList.remove('button-visible'));
    const uploadBtnEl = document.querySelector('.button.upload button');
    if (uploadBtnEl) uploadBtnEl.textContent = 'Upload';
}

function showActionButtons() {
    const actionButtons = [
        document.querySelector('.button.first'),
        document.querySelector('.button.sec'),
        document.querySelector('.button.third'),
        document.querySelector('.button.fourth')
    ];
    
    actionButtons.forEach(button => {
        if (button) {
            button.classList.add('button-visible');
        }
    });
}

function initializeButtonStates() {
    const uploadButton = document.querySelector('.button.upload');
    const actionButtons = [
        document.querySelector('.button.first'),
        document.querySelector('.button.sec'),
        document.querySelector('.button.third'),
        document.querySelector('.button.fourth')
    ];
    
    if (uploadButton) {
        uploadButton.classList.remove('button-hidden');
    }
    
    actionButtons.forEach(button => {
        if (button) {
            button.classList.remove('button-visible');
        }
    });
}

function showToast(message, type = 'info') {
    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;
    toast.style.cssText = `
        position: fixed;
        bottom: 20px;
        left: 50%;
        transform: translateX(-50%);
        padding: 12px 24px;
        color: white;
        z-index: 10000;
        font-family: monospace;
        font-size: 12px;
    `;
    document.body.appendChild(toast);
    
    let currentIndex = 0;
    const typingSpeed = 30;
    
    function typeCharacter() {
        if (currentIndex < message.length) {
            toast.textContent = message.substring(0, currentIndex + 1);
            currentIndex++;
            setTimeout(typeCharacter, typingSpeed);
        } else {
            setTimeout(() => {
                toast.style.opacity = '0';
                toast.style.transition = 'opacity 0.3s';
                setTimeout(() => toast.remove(), 300);
            }, 2000);
        }
    }
    
    typeCharacter();
}

function setButtonLoading(button, loading) {
    const btn = button.querySelector('button');
    if (loading) {
        btn.disabled = true;
        btn.style.opacity = '0.6';
        btn.textContent = btn.textContent + '...';
    } else {
        btn.disabled = false;
        btn.style.opacity = '1';
        btn.textContent = btn.textContent.replace('...', '');
    }
}

async function uploadFile() {
    const uploadButton = document.querySelector('.button.upload');
    const btn = uploadButton?.querySelector('button');
    
    if (btn && btn.textContent === 'Finish Operation') {
        resetToInitialState();
        showToast('Operation finished. Ready for new upload.', 'success');
        return;
    }
    
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = '.html,.htm';
    input.onchange = async (e) => {
        const file = e.target.files[0];
        if (!file) return;

        try {
            const text = await file.text();
            uploadedHTML = text;
            showToast(`File "${file.name}" uploaded successfully!`, 'success');
            
            changeUploadButtonToFinish();
            setTimeout(() => {
                showActionButtons();
            }, 200);
        } catch (error) {
            showToast('Error reading file: ' + error.message, 'error');
        }
    };
    input.click();
}

async function formatHTML() {
    if (!uploadedHTML) {
        showToast('Please upload an HTML file first', 'error');
        return;
    }

    const button = document.querySelector('.button.sec');
    setButtonLoading(button, true);

    try {
        const response = await fetch(`${API_BASE}/api/format`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ html: uploadedHTML }),
        });

        const data = await response.json();

        if (data.success) {
            const blob = new Blob([data.data], { type: 'text/html' });
            const filename = resolveDownloadName('download-name-format', DOWNLOAD_DEFAULTS.formatName, '.html');
            const result = await downloadBlob(blob, filename, PICKER_TYPES.html);
            if (result !== 'canceled') {
                showToast('HTML formatted and downloaded!', 'success');
            }
        } else {
            showToast(data.error || 'Formatting failed', 'error');
        }
    } catch (error) {
        showToast('Error: ' + error.message, 'error');
    } finally {
        setButtonLoading(button, false);
    }
}

async function exportZip() {
    if (!uploadedHTML) {
        showToast('Please upload an HTML file first', 'error');
        return;
    }

    const button = document.querySelector('.button.fourth');
    setButtonLoading(button, true);

    try {
        const response = await fetch(`${API_BASE}/api/export`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ html: uploadedHTML }),
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Export failed');
        }

        const blob = await response.blob();
        const filename = resolveDownloadName('download-name-zip', DOWNLOAD_DEFAULTS.zipName, '.zip');
        const result = await downloadBlob(blob, filename, PICKER_TYPES.zip);
        if (result !== 'canceled') {
            showToast('Files extracted and downloaded as ZIP!', 'success');
        }
    } catch (error) {
        showToast('Error: ' + error.message, 'error');
    } finally {
        setButtonLoading(button, false);
    }
}

async function exportTSXProject() {
    if (!uploadedHTML) {
        showToast('Please upload an HTML file first', 'error');
        return;
    }

    const button = document.querySelector('.button.first');
    setButtonLoading(button, true);

    try {
        const response = await fetch(`${API_BASE}/api/export-nodejs`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ html: uploadedHTML }),
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Export failed');
        }

        const blob = await response.blob();
        const contentDisposition = response.headers.get('Content-Disposition');
        const serverFilename = contentDisposition
            ? contentDisposition.split('filename=')[1]?.replace(/"/g, '') || 'project.zip'
            : 'project.zip';
        const filename = resolveDownloadName('download-name-tsx', serverFilename, '.zip');
        const result = await downloadBlob(blob, filename, PICKER_TYPES.zip);
        if (result !== 'canceled') {
            showToast('TSX project exported and downloaded!', 'success');
        }
    } catch (error) {
        showToast('Error: ' + error.message, 'error');
    } finally {
        setButtonLoading(button, false);
    }
}

async function scrapeAndExport(exportType) {
    const urlInput = document.getElementById('scrape-url-input');
    const rawURL = urlInput ? urlInput.value.trim() : '';
    if (!rawURL) {
        showToast('Please enter a URL to scrape', 'error');
        return;
    }

    const endpointMap = {
        zip: '/api/scrape',
        nodejs: '/api/scrape-nodejs',
        ejs: '/api/scrape-nodejs-ejs',
    };

    const endpoint = endpointMap[exportType];
    if (!endpoint) return;

    // Find the button that triggered this (the action button)
    const buttonSelectors = { zip: '.button.fourth', nodejs: '.button.first', ejs: '.button.third' };
    const button = document.querySelector(buttonSelectors[exportType]);
    if (button) setButtonLoading(button, true);

    try {
        const response = await fetch(`${API_BASE}${endpoint}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ url: rawURL }),
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Scrape failed');
        }

        const blob = await response.blob();
        const contentDisposition = response.headers.get('Content-Disposition');
        let defaultName = exportType === 'ejs' ? 'project-ejs.zip' : exportType === 'nodejs' ? 'project.zip' : 'extracted.zip';
        const serverFilename = contentDisposition
            ? contentDisposition.split('filename=')[1]?.replace(/"/g, '') || defaultName
            : defaultName;

        const nameInputIds = { zip: 'download-name-zip', nodejs: 'download-name-tsx', ejs: 'download-name-ejs' };
        const filename = resolveDownloadName(nameInputIds[exportType], defaultName, '.zip');
        const result = await downloadBlob(blob, filename || serverFilename, PICKER_TYPES.zip);
        if (result !== 'canceled') {
            showToast('Scraped and exported successfully!', 'success');
        }
    } catch (error) {
        showToast('Error: ' + error.message, 'error');
    } finally {
        if (button) setButtonLoading(button, false);
    }
}

async function exportEJSProject() {
    if (!uploadedHTML) {
        showToast('Please upload an HTML file first', 'error');
        return;
    }

    const button = document.querySelector('.button.third');
    setButtonLoading(button, true);

    try {
        const response = await fetch(`${API_BASE}/api/export-nodejs-ejs`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ html: uploadedHTML }),
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Export failed');
        }

        const blob = await response.blob();
        const contentDisposition = response.headers.get('Content-Disposition');
        const serverFilename = contentDisposition
            ? contentDisposition.split('filename=')[1]?.replace(/"/g, '') || 'project-ejs.zip'
            : 'project-ejs.zip';
        const filename = resolveDownloadName('download-name-ejs', serverFilename, '.zip');
        const result = await downloadBlob(blob, filename, PICKER_TYPES.zip);
        if (result !== 'canceled') {
            showToast('EJS project exported and downloaded!', 'success');
        }
    } catch (error) {
        showToast('Error: ' + error.message, 'error');
    } finally {
        setButtonLoading(button, false);
    }
}

document.addEventListener('DOMContentLoaded', () => {
    applyTheme(localStorage.getItem(THEME_KEY) || 'dark');
    initializeButtonStates();
    initializeDownloadSettings();

    const dropdowns = document.querySelectorAll('.dropdown');

    dropdowns.forEach(dropdown => {
        const dropdownTrigger = dropdown.querySelector('.dropdown-trigger');

        if (dropdownTrigger) {
            dropdownTrigger.addEventListener('click', (e) => {
                e.stopPropagation();
                dropdown.classList.toggle('active');
            });
        }

        const nestedTriggers = dropdown.querySelectorAll('.nested-trigger');
        nestedTriggers.forEach(trigger => {
            trigger.addEventListener('click', (e) => {
                e.stopPropagation();
                trigger.classList.toggle('active');
            });
        });
    });

    document.addEventListener('click', (e) => {
        dropdowns.forEach(dropdown => {
            if (!dropdown.contains(e.target)) {
                dropdown.classList.remove('active');
                const nestedTriggers = dropdown.querySelectorAll('.nested-trigger');
                nestedTriggers.forEach(trigger => {
                    trigger.classList.remove('active');
                });
            }
        });
    });

    // ── input mode toggle buttons ──
    const modeUploadBtn = document.getElementById('mode-upload-btn');
    const modeScrapeBtn = document.getElementById('mode-scrape-btn');
    if (modeUploadBtn) modeUploadBtn.addEventListener('click', () => setInputMode('upload'));
    if (modeScrapeBtn) modeScrapeBtn.addEventListener('click', () => setInputMode('scrape'));

    // ── scrape URL action button (shows export options after URL is set) ──
    const scrapeActionBtn = document.getElementById('scrape-action-btn');
    if (scrapeActionBtn) {
        scrapeActionBtn.addEventListener('click', (e) => {
            e.preventDefault();
            const urlInput = document.getElementById('scrape-url-input');
            if (!urlInput || !urlInput.value.trim()) {
                showToast('Please enter a URL', 'error');
                return;
            }
            // Show action buttons so user can choose export format
            showActionButtons();
            showToast('URL ready — choose an export format above', 'success');
        });
    }

    const uploadButton = document.querySelector('.button.upload');
    const formatButton = document.querySelector('.button.sec');
    const exportEJSButton = document.querySelector('.button.third');
    const exportZipButton = document.querySelector('.button.fourth');
    const exportTSXButton = document.querySelector('.button.first');

    if (uploadButton) {
        uploadButton.addEventListener('click', (e) => {
            e.preventDefault();
            uploadFile();
        });
    }

    if (formatButton) {
        formatButton.addEventListener('click', (e) => {
            e.preventDefault();
            formatHTML();
        });
    }

    if (exportZipButton) {
        exportZipButton.addEventListener('click', (e) => {
            e.preventDefault();
            if (scrapeMode) {
                scrapeAndExport('zip');
            } else {
                exportZip();
            }
        });
    }

    if (exportTSXButton) {
        exportTSXButton.addEventListener('click', (e) => {
            e.preventDefault();
            if (scrapeMode) {
                scrapeAndExport('nodejs');
            } else {
                exportTSXProject();
            }
        });
    }

    if (exportEJSButton) {
        exportEJSButton.addEventListener('click', (e) => {
            e.preventDefault();
            if (scrapeMode) {
                scrapeAndExport('ejs');
            } else {
                exportEJSProject();
            }
        });
    }
});
