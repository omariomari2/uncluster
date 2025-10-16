// Application state
let currentOutput = '';
let currentSuggestions = [];

// DOM elements
const htmlInput = document.getElementById('htmlInput');
const outputCode = document.getElementById('outputCode');
const processBtn = document.getElementById('processBtn');
const copyBtn = document.getElementById('copyBtn');
const downloadBtn = document.getElementById('downloadBtn');
const clearBtn = document.getElementById('clearBtn');
const fileInput = document.getElementById('fileInput');
const suggestionsSection = document.getElementById('suggestionsSection');
const suggestionsContent = document.getElementById('suggestionsContent');
const toggleSuggestions = document.getElementById('toggleSuggestions');
const loadingOverlay = document.getElementById('loadingOverlay');
const toast = document.getElementById('toast');

// Event listeners
document.addEventListener('DOMContentLoaded', function() {
    processBtn.addEventListener('click', processHTML);
    copyBtn.addEventListener('click', copyToClipboard);
    downloadBtn.addEventListener('click', downloadOutput);
    clearBtn.addEventListener('click', clearInput);
    fileInput.addEventListener('change', handleFileUpload);
    toggleSuggestions.addEventListener('click', toggleSuggestionsPanel);
    
    // Auto-resize textarea
    htmlInput.addEventListener('input', autoResizeTextarea);
    
    // Handle mode changes
    document.querySelectorAll('input[name="mode"]').forEach(radio => {
        radio.addEventListener('change', function() {
            if (this.value === 'jsx') {
                suggestionsSection.style.display = 'block';
            } else {
                suggestionsSection.style.display = 'none';
            }
        });
    });
});

// Process HTML based on selected mode
async function processHTML() {
    const html = htmlInput.value.trim();
    if (!html) {
        showToast('Please enter some HTML content', 'error');
        return;
    }

    const mode = document.querySelector('input[name="mode"]:checked').value;
    
    showLoading(true);
    
    try {
        let response;
        if (mode === 'format') {
            response = await fetch('/api/format', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ html: html })
            });
        } else if (mode === 'jsx') {
            response = await fetch('/api/convert', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ html: html })
            });
        }

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const result = await response.json();
        
        if (!result.success) {
            throw new Error(result.error || 'Processing failed');
        }

        currentOutput = result.data;
        outputCode.textContent = currentOutput;
        
        // Enable output buttons
        copyBtn.disabled = false;
        downloadBtn.disabled = false;
        
        // If JSX mode, also get component suggestions
        if (mode === 'jsx') {
            await getComponentSuggestions(html);
        } else {
            suggestionsSection.style.display = 'none';
        }
        
        showToast('Processing completed successfully!');
        
    } catch (error) {
        console.error('Error processing HTML:', error);
        showToast('Error processing HTML: ' + error.message, 'error');
        outputCode.textContent = 'Error: ' + error.message;
    } finally {
        showLoading(false);
    }
}

// Get component suggestions for JSX mode
async function getComponentSuggestions(html) {
    try {
        const response = await fetch('/api/analyze', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ html: html })
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const result = await response.json();
        
        if (!result.success) {
            throw new Error(result.error || 'Analysis failed');
        }

        currentSuggestions = result.suggestions || [];
        displaySuggestions(currentSuggestions);
        
    } catch (error) {
        console.error('Error getting suggestions:', error);
        // Don't show error toast for suggestions as it's not critical
        suggestionsContent.innerHTML = '<p>Unable to analyze components</p>';
    }
}

// Display component suggestions
function displaySuggestions(suggestions) {
    if (!suggestions || suggestions.length === 0) {
        suggestionsContent.innerHTML = '<p>No component suggestions found</p>';
        return;
    }

    suggestionsContent.innerHTML = suggestions.map(suggestion => `
        <div class="suggestion-card">
            <h3>${suggestion.name}</h3>
            <p>${suggestion.description}</p>
            <div class="suggestion-meta">
                <span>Tag: ${suggestion.tagName}</span>
                <span>Count: ${suggestion.count}</span>
                ${Object.keys(suggestion.attributes).length > 0 ? `<span>Props: ${Object.keys(suggestion.attributes).length}</span>` : ''}
            </div>
            <div class="suggestion-code">${escapeHtml(suggestion.jsxCode)}</div>
        </div>
    `).join('');
}

// Copy output to clipboard
async function copyToClipboard() {
    if (!currentOutput) {
        showToast('No output to copy', 'error');
        return;
    }

    try {
        await navigator.clipboard.writeText(currentOutput);
        showToast('Copied to clipboard!');
    } catch (error) {
        console.error('Error copying to clipboard:', error);
        showToast('Failed to copy to clipboard', 'error');
    }
}

// Download output as file
function downloadOutput() {
    if (!currentOutput) {
        showToast('No output to download', 'error');
        return;
    }

    const mode = document.querySelector('input[name="mode"]:checked').value;
    const extension = mode === 'jsx' ? 'jsx' : 'html';
    const filename = `formatted.${extension}`;
    
    const blob = new Blob([currentOutput], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    
    showToast(`Downloaded as ${filename}`);
}

// Clear input
function clearInput() {
    htmlInput.value = '';
    outputCode.textContent = '';
    currentOutput = '';
    copyBtn.disabled = true;
    downloadBtn.disabled = true;
    suggestionsSection.style.display = 'none';
    suggestionsContent.innerHTML = '';
    autoResizeTextarea();
}

// Handle file upload
function handleFileUpload(event) {
    const file = event.target.files[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = function(e) {
        htmlInput.value = e.target.result;
        autoResizeTextarea();
    };
    reader.readAsText(file);
}

// Auto-resize textarea based on content
function autoResizeTextarea() {
    htmlInput.style.height = 'auto';
    htmlInput.style.height = htmlInput.scrollHeight + 'px';
}

// Toggle suggestions panel
function toggleSuggestionsPanel() {
    const isVisible = suggestionsSection.style.display !== 'none';
    suggestionsSection.style.display = isVisible ? 'none' : 'block';
    toggleSuggestions.textContent = isVisible ? 'Show Suggestions' : 'Collapse';
}

// Show/hide loading overlay
function showLoading(show) {
    loadingOverlay.style.display = show ? 'flex' : 'none';
}

// Show toast notification
function showToast(message, type = 'success') {
    toast.textContent = message;
    toast.className = `toast ${type}`;
    toast.style.display = 'block';
    
    setTimeout(() => {
        toast.style.display = 'none';
    }, 3000);
}

// Escape HTML for display
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Handle keyboard shortcuts
document.addEventListener('keydown', function(e) {
    // Ctrl/Cmd + Enter to process
    if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
        e.preventDefault();
        processHTML();
    }
    
    // Ctrl/Cmd + S to save (download)
    if ((e.ctrlKey || e.metaKey) && e.key === 's') {
        e.preventDefault();
        if (!downloadBtn.disabled) {
            downloadOutput();
        }
    }
});
