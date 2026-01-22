const API_BASE = window.location.origin.includes('localhost') && !window.location.port.includes('3000')
    ? 'http://localhost:3000'
    : '';

let uploadedHTML = '';

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
        document.querySelector('.button.fifth'),
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
}

function showActionButtons() {
    const actionButtons = [
        document.querySelector('.button.first'),
        document.querySelector('.button.fifth'),
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
        document.querySelector('.button.fifth'),
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
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = 'formatted.html';
            a.click();
            URL.revokeObjectURL(url);
            showToast('HTML formatted and downloaded!', 'success');
        } else {
            showToast(data.error || 'Formatting failed', 'error');
        }
    } catch (error) {
        showToast('Error: ' + error.message, 'error');
    } finally {
        setButtonLoading(button, false);
    }
}

async function convertToJSX() {
    if (!uploadedHTML) {
        showToast('Please upload an HTML file first', 'error');
        return;
    }

    const button = document.querySelector('.button.third');
    setButtonLoading(button, true);

    try {
        const response = await fetch(`${API_BASE}/api/convert`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ html: uploadedHTML }),
        });

        const data = await response.json();

        if (data.success) {
            const blob = new Blob([data.data], { type: 'text/jsx' });
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = 'converted.jsx';
            a.click();
            URL.revokeObjectURL(url);
            showToast('HTML converted to JSX and downloaded!', 'success');
        } else {
            showToast(data.error || 'Conversion failed', 'error');
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
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = 'extracted.zip';
        a.click();
        URL.revokeObjectURL(url);
        showToast('Files extracted and downloaded as ZIP!', 'success');
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
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        const contentDisposition = response.headers.get('Content-Disposition');
        const filename = contentDisposition
            ? contentDisposition.split('filename=')[1]?.replace(/"/g, '') || 'project.zip'
            : 'project.zip';
        a.href = url;
        a.download = filename;
        a.click();
        URL.revokeObjectURL(url);
        showToast('TSX project exported and downloaded!', 'success');
    } catch (error) {
        showToast('Error: ' + error.message, 'error');
    } finally {
        setButtonLoading(button, false);
    }
}

async function exportEJSProject() {
    if (!uploadedHTML) {
        showToast('Please upload an HTML file first', 'error');
        return;
    }

    const button = document.querySelector('.button.fifth');
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
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        const contentDisposition = response.headers.get('Content-Disposition');
        const filename = contentDisposition
            ? contentDisposition.split('filename=')[1]?.replace(/"/g, '') || 'project-ejs.zip'
            : 'project-ejs.zip';
        a.href = url;
        a.download = filename;
        a.click();
        URL.revokeObjectURL(url);
        showToast('EJS project exported and downloaded!', 'success');
    } catch (error) {
        showToast('Error: ' + error.message, 'error');
    } finally {
        setButtonLoading(button, false);
    }
}

document.addEventListener('DOMContentLoaded', () => {
    initializeButtonStates();
    
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
    
    const uploadButton = document.querySelector('.button.upload');
    const formatButton = document.querySelector('.button.sec');
    const convertButton = document.querySelector('.button.third');
    const exportZipButton = document.querySelector('.button.fourth');
    const exportTSXButton = document.querySelector('.button.first');
    const exportEJSButton = document.querySelector('.button.fifth');

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

    if (convertButton) {
        convertButton.addEventListener('click', (e) => {
            e.preventDefault();
            convertToJSX();
        });
    }

    if (exportZipButton) {
        exportZipButton.addEventListener('click', (e) => {
            e.preventDefault();
            exportZip();
        });
    }

    if (exportTSXButton) {
        exportTSXButton.addEventListener('click', (e) => {
            e.preventDefault();
            exportTSXProject();
        });
    }

    if (exportEJSButton) {
        exportEJSButton.addEventListener('click', (e) => {
            e.preventDefault();
            exportEJSProject();
        });
    }
});
