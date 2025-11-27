// Global variables
let currentPage = 1;
let totalPages = 1;

// DOM elements
const elements = {
    initBtn: document.getElementById('initBtn'),
    filterBtn: document.getElementById('filterBtn'),
    clearBtn: document.getElementById('clearBtn'),
    truncateBtn: document.getElementById('truncateBtn'),
    refreshBtn: document.getElementById('refreshBtn'),
    logsTableBody: document.getElementById('logsTableBody'),
    alertContainer: document.getElementById('alertContainer'),
    paginationContainer: document.getElementById('paginationContainer'),
    contentModal: new bootstrap.Modal(document.getElementById('contentModal')),
    contentDisplay: document.getElementById('contentDisplay'),
    // Stats
    totalRecords: document.getElementById('totalRecords'),
    currentPageDisplay: document.getElementById('currentPage'),
    pageSize: document.getElementById('pageSize'),
    totalPagesDisplay: document.getElementById('totalPages')
};

// Event listeners
document.addEventListener('DOMContentLoaded', () => {
    elements.initBtn.addEventListener('click', initializeData);
    elements.filterBtn.addEventListener('click', applyFilters);
    elements.clearBtn.addEventListener('click', clearFilters);
    elements.truncateBtn.addEventListener('click', truncateDatabase);
    elements.refreshBtn.addEventListener('click', loadLogs);

    // Initial load
    loadLogs();
});

// API functions
async function apiCall(url, options = {}) {
    try {
        const response = await fetch(url, {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            ...options
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        return await response.json();
    } catch (error) {
        console.error('API call failed:', error);
        showAlert('API call failed: ' + error.message, 'danger');
        throw error;
    }
}

// Initialize data
async function initializeData() {
    const recordCount = document.getElementById('recordCount').value;
    const contentSize = document.getElementById('contentSize').value;

    elements.initBtn.disabled = true;
    elements.initBtn.innerHTML = '<i class="fas fa-spinner fa-spin me-2"></i>Initializing...';

    try {
        const result = await apiCall('/api/initialize', {
            method: 'POST',
            body: JSON.stringify({
                record_count: parseInt(recordCount),
                content_size: contentSize
            })
        });

        showAlert(`Successfully initialized ${result.record_count} records in ${result.duration}`, 'success');
        loadLogs();
    } catch (error) {
        showAlert('Failed to initialize data', 'danger');
    } finally {
        elements.initBtn.disabled = false;
        elements.initBtn.innerHTML = '<i class="fas fa-play me-2"></i>Initialize';
    }
}

// Load logs with filters
async function loadLogs(page = 1) {
    const filters = getFilters();
    const params = new URLSearchParams({
        page: page,
        limit: 50,
        ...filters
    });

    let endpoint = '/api/logs';
    if (filters.search_term) {
        endpoint = '/api/search/partial';
    }

    try {
        const result = await apiCall(`${endpoint}?${params}`);
        displayLogs(result.data);
        updatePagination(result.page, result.total_pages, result.total);
        updateStats(result.total, result.page, result.limit, result.total_pages);
        currentPage = result.page;
        totalPages = result.total_pages;

        if (result.query_duration) {
            showAlert(`Query completed in ${result.query_duration}`, 'info');
        }
    } catch (error) {
        console.error('Failed to load logs:', error);
    }
}

// Apply filters
function applyFilters() {
    currentPage = 1;
    loadLogs(1);
}

// Clear filters
function clearFilters() {
    document.getElementById('userId').value = '';
    document.getElementById('domain').value = '';
    document.getElementById('createdAt').value = '';
    document.getElementById('createdAtTo').value = '';
    document.getElementById('contentLike').value = '';
    document.getElementById('searchType').value = 'fulltext';

    currentPage = 1;
    loadLogs(1);
}

// Get current filters
function getFilters() {
    const filters = {};

    const userId = document.getElementById('userId').value;
    if (userId) filters.user_id = userId;

    const domain = document.getElementById('domain').value;
    if (domain) filters.domain = domain;

    const createdAt = document.getElementById('createdAt').value;
    if (createdAt) filters.created_at = createdAt;

    const createdAtTo = document.getElementById('createdAtTo').value;
    if (createdAtTo) filters.created_at_to = createdAtTo;

    const contentLike = document.getElementById('contentLike').value;
    const searchType = document.getElementById('searchType').value;

    if (contentLike) {
        if (searchType === 'partial') {
            filters.search_term = contentLike;
        } else {
            filters.content_like = contentLike;
        }
    }

    return filters;
}

// Display logs in table
function displayLogs(logs) {
    if (!logs || logs.length === 0) {
        elements.logsTableBody.innerHTML = `
            <tr>
                <td colspan="6" class="text-center">No data available. Click "Initialize" to generate sample data.</td>
            </tr>
        `;
        return;
    }

    elements.logsTableBody.innerHTML = logs.map(log => `
        <tr>
            <td><code>${log.id.substring(0, 8)}...</code></td>
            <td><code>${log.user_id.substring(0, 8)}...</code></td>
            <td>${log.domain}</td>
            <td><span class="badge bg-secondary">${log.action}</span></td>
            <td>${formatDate(log.created_at)}</td>
            <td>
                <button class="btn btn-sm btn-outline-primary" onclick="showContent(${JSON.stringify(log.content).replace(/"/g, '&quot;')})">
                    <i class="fas fa-eye"></i> View
                </button>
            </td>
        </tr>
    `).join('');
}

// Show content in modal
function showContent(content) {
    elements.contentDisplay.textContent = JSON.stringify(content, null, 2);
    elements.contentModal.show();
}

// Update pagination
function updatePagination(current, total, totalRecords) {
    if (total <= 1) {
        elements.paginationContainer.innerHTML = '';
        return;
    }

    let pagination = '<nav><ul class="pagination justify-content-center">';

    // Previous button
    pagination += `
        <li class="page-item ${current === 1 ? 'disabled' : ''}">
            <a class="page-link" href="#" onclick="loadLogs(${current - 1}); return false;">
                <i class="fas fa-chevron-left"></i> Previous
            </a>
        </li>
    `;

    // Page numbers
    const startPage = Math.max(1, current - 2);
    const endPage = Math.min(total, current + 2);

    if (startPage > 1) {
        pagination += `<li class="page-item"><a class="page-link" href="#" onclick="loadLogs(1); return false;">1</a></li>`;
        if (startPage > 2) {
            pagination += `<li class="page-item disabled"><a class="page-link" href="#">...</a></li>`;
        }
    }

    for (let i = startPage; i <= endPage; i++) {
        pagination += `
            <li class="page-item ${i === current ? 'active' : ''}">
                <a class="page-link" href="#" onclick="loadLogs(${i}); return false;">${i}</a>
            </li>
        `;
    }

    if (endPage < total) {
        if (endPage < total - 1) {
            pagination += `<li class="page-item disabled"><a class="page-link" href="#">...</a></li>`;
        }
        pagination += `<li class="page-item"><a class="page-link" href="#" onclick="loadLogs(${total}); return false;">${total}</a></li>`;
    }

    // Next button
    pagination += `
        <li class="page-item ${current === total ? 'disabled' : ''}">
            <a class="page-link" href="#" onclick="loadLogs(${current + 1}); return false;">
                Next <i class="fas fa-chevron-right"></i>
            </a>
        </li>
    `;

    pagination += '</ul></nav>';
    elements.paginationContainer.innerHTML = pagination;
}

// Update statistics
function updateStats(total, page, limit, totalPages) {
    elements.totalRecords.textContent = total.toLocaleString();
    elements.currentPageDisplay.textContent = page;
    elements.pageSize.textContent = limit;
    elements.totalPagesDisplay.textContent = totalPages;
}

// Truncate database
async function truncateDatabase() {
    if (!confirm('Are you sure you want to delete all log records? This action cannot be undone.')) {
        return;
    }

    elements.truncateBtn.disabled = true;
    elements.truncateBtn.innerHTML = '<i class="fas fa-spinner fa-spin me-2"></i>Truncating...';

    try {
        const result = await apiCall('/api/truncate', {
            method: 'DELETE'
        });

        showAlert(`Successfully truncated database. ${result.rows_affected} records deleted.`, 'success');
        loadLogs();
    } catch (error) {
        showAlert('Failed to truncate database', 'danger');
    } finally {
        elements.truncateBtn.disabled = false;
        elements.truncateBtn.innerHTML = '<i class="fas fa-trash me-2"></i>Truncate Database';
    }
}

// Show alert message
function showAlert(message, type = 'info') {
    const alertHtml = `
        <div class="alert alert-${type} alert-dismissible fade show" role="alert">
            <i class="fas fa-${getAlertIcon(type)} me-2"></i>
            ${message}
            <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
        </div>
    `;

    elements.alertContainer.innerHTML = alertHtml;

    // Auto-dismiss after 5 seconds
    setTimeout(() => {
        const alert = elements.alertContainer.querySelector('.alert');
        if (alert) {
            const bsAlert = new bootstrap.Alert(alert);
            bsAlert.close();
        }
    }, 5000);
}

// Get appropriate icon for alert type
function getAlertIcon(type) {
    const icons = {
        success: 'check-circle',
        danger: 'exclamation-triangle',
        warning: 'exclamation-circle',
        info: 'info-circle'
    };
    return icons[type] || 'info-circle';
}

// Format date
function formatDate(dateString) {
    const date = new Date(dateString);
    return date.toLocaleString();
}