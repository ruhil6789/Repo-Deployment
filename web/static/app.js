// Dashboard JavaScript - Vercel-like UI
const API_BASE = '/api';

// Check authentication
function checkAuth() {
    const token = localStorage.getItem('token');
    if (!token) {
        const urlParams = new URLSearchParams(window.location.search);
        const urlToken = urlParams.get('token');
        if (urlToken) {
            localStorage.setItem('token', urlToken);
            window.history.replaceState({}, document.title, window.location.pathname);
            return urlToken;
        }
        window.location.href = '/login';
        return null;
    }
    return token;
}

// API request helper
async function apiRequest(endpoint, options = {}) {
    const token = localStorage.getItem('token');
    const headers = {
        'Content-Type': 'application/json',
        ...options.headers
    };
    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }
    
    try {
        const response = await fetch(`${API_BASE}${endpoint}`, {
            ...options,
            headers
        });

        if (response.status === 401) {
            localStorage.removeItem('token');
            localStorage.removeItem('user');
            window.location.href = '/login';
            return null;
        }

        if (!response.ok) {
            let errorData;
            try {
                errorData = await response.json();
            } catch {
                errorData = { error: response.statusText };
            }
            console.error(`API Error for ${endpoint}:`, errorData.error || response.statusText, 'Status:', response.status);
            return null;
        }

        try {
            return await response.json();
        } catch (error) {
            console.error(`JSON parse error for ${endpoint}:`, error);
            return null;
        }
    } catch (error) {
        console.error(`Network error for ${endpoint}:`, error);
        return null;
    }
}

// Load user info
async function loadUser() {
    let user = JSON.parse(localStorage.getItem('user') || '{}');
    
    if (!user.username && !user.email) {
        try {
            const userData = await apiRequest('/profile');
            if (userData) {
                user = {
                    id: userData.user_id,
                    username: userData.username,
                    email: userData.email,
                    avatar_url: userData.avatar_url
                };
                localStorage.setItem('user', JSON.stringify(user));
            }
        } catch (error) {
            console.error('Failed to load user info:', error);
        }
    }
    
    const usernameEl = document.getElementById('username');
    if (usernameEl) {
        usernameEl.textContent = user.username || user.email || 'User';
    }
}

// Get status badge HTML (Vercel style)
function getStatusBadge(status) {
    const statusConfig = {
        'ready': { label: 'Ready', class: 'status-ready', text: 'text-green-400' },
        'deployed': { label: 'Ready', class: 'status-ready', text: 'text-green-400' },
        'building': { label: 'Building', class: 'status-building', text: 'text-blue-400' },
        'deploying': { label: 'Deploying', class: 'status-building', text: 'text-blue-400' },
        'failed': { label: 'Error', class: 'status-error', text: 'text-red-400' },
        'pending': { label: 'Pending', class: 'status-pending', text: 'text-yellow-400' }
    };
    
    const config = statusConfig[status?.toLowerCase()] || { label: status, class: 'status-pending', text: 'text-gray-400' };
    
    return `
        <span class="inline-flex items-center text-xs font-medium ${config.text}">
            <span class="status-dot ${config.class}"></span>
            ${config.label}
        </span>
    `;
}

// Load projects (Vercel style)
async function loadProjects() {
    try {
        console.log('Loading projects...');
        const projects = await apiRequest('/projects');
        console.log('Projects response:', projects);
        
        const container = document.getElementById('projects');
        if (!container) {
            console.error('Projects container not found');
            return;
        }

        if (!projects) {
            container.innerHTML = `
                <div class="col-span-full text-center py-12">
                    <p class="text-red-400 mb-2">Error loading projects</p>
                    <p class="text-gray-500 text-sm">Check console for details</p>
                </div>
            `;
            return;
        }

        if (!Array.isArray(projects)) {
            console.error('Projects is not an array:', typeof projects, projects);
            container.innerHTML = `
                <div class="col-span-full text-center py-12">
                    <p class="text-red-400">Invalid data format</p>
                </div>
            `;
            return;
        }

        if (projects.length === 0) {
            container.innerHTML = `
                <div class="col-span-full text-center py-12">
                    <div class="mb-4">
                        <svg class="w-16 h-16 text-gray-700 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"/>
                        </svg>
                    </div>
                    <h3 class="text-lg font-medium text-white mb-2">No projects yet</h3>
                    <p class="text-gray-400 text-sm mb-4">Get started by connecting a GitHub repository</p>
                    <button id="addProjectBtn" class="px-4 py-2 bg-white text-black rounded hover:bg-gray-100 text-sm font-medium">
                        Add New Project
                    </button>
                </div>
            `;
            return;
        }

        console.log('Rendering projects:', projects);

        container.innerHTML = projects.map(project => {
            const deployments = project.deployments || project.Deployments || [];
            const latestDeployment = deployments.find(d => (d.hostname || d.Hostname)) || deployments[0];
            const hostname = latestDeployment?.hostname || latestDeployment?.Hostname || '';
            const liveUrl = hostname ? `http://${hostname}` : null;
            const status = latestDeployment?.status || 'pending';
            
            // Count deployments
            let deploymentCount = deployments.length;
            
            return `
                <div class="project-card bg-gray-900 border border-gray-800 rounded-lg p-6 cursor-pointer hover:border-gray-700">
                    <div class="flex items-start justify-between mb-4">
                        <div class="flex-1">
                            <h3 class="text-lg font-semibold text-white mb-1">${project.name || 'Unnamed Project'}</h3>
                            <p class="text-sm text-gray-400">${project.repo_owner || ''}/${project.repo_name || ''}</p>
                        </div>
                        ${liveUrl ? `
                            <a href="${liveUrl}" target="_blank" 
                               class="ml-2 px-3 py-1.5 bg-green-500/10 border border-green-500/20 rounded text-green-400 hover:bg-green-500/20 text-xs font-medium flex items-center"
                               onclick="event.stopPropagation()">
                                <svg class="w-3 h-3 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"/>
                                </svg>
                                Visit
                            </a>
                        ` : ''}
                    </div>
                    
                    <div class="flex items-center justify-between pt-4 border-t border-gray-800">
                        <div class="flex items-center space-x-4">
                            ${getStatusBadge(status)}
                            <span class="text-xs text-gray-500">${deploymentCount} deployment${deploymentCount !== 1 ? 's' : ''}</span>
                        </div>
                        ${project.repo_url ? `
                            <a href="${project.repo_url}" target="_blank" 
                               class="text-gray-400 hover:text-white text-xs"
                               onclick="event.stopPropagation()">
                                <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
                                    <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
                                </svg>
                            </a>
                        ` : ''}
                    </div>
                </div>
            `;
        }).join('');
    } catch (error) {
        console.error('Error loading projects:', error);
        const container = document.getElementById('projects');
        if (container) {
            container.innerHTML = `
                <div class="col-span-full text-center py-12">
                    <p class="text-red-400">Error loading projects</p>
                    <p class="text-gray-500 text-sm mt-2">${error.message}</p>
                </div>
            `;
        }
    }
}

// Load deployments (Vercel style)
async function loadDeployments() {
    try {
        const deployments = await apiRequest('/deployments');
        if (!deployments) {
            console.error('Failed to load deployments');
            return;
        }

        const container = document.getElementById('deployments');
        if (!container) {
            console.error('Deployments container not found');
            return;
        }

        if (deployments.length === 0) {
            container.innerHTML = `
                <div class="text-center py-12 border border-gray-800 rounded-lg bg-gray-900/50">
                    <p class="text-gray-400 text-sm">No deployments yet. Push code to trigger a deployment.</p>
                </div>
            `;
            return;
        }

        container.innerHTML = deployments.slice(0, 10).map(deployment => {
            const date = new Date(deployment.created_at).toLocaleString();
            const commitShort = deployment.commit_sha?.substring(0, 7) || 'N/A';
            const hostname = deployment.hostname || '';
            const status = deployment.status || 'pending';
            const projectName = deployment.project?.name || 'Unknown Project';
            
            return `
                <div class="bg-gray-900 border border-gray-800 rounded-lg p-4 hover:border-gray-700 transition-colors">
                    <div class="flex items-center justify-between">
                        <div class="flex-1 min-w-0">
                            <div class="flex items-center space-x-3 mb-2">
                                <h3 class="text-sm font-medium text-white truncate">${projectName}</h3>
                                ${getStatusBadge(status)}
                            </div>
                            <p class="text-xs text-gray-400 truncate mb-2">${deployment.commit_msg || 'No commit message'}</p>
                            <div class="flex items-center space-x-4 text-xs text-gray-500">
                                <span>${deployment.branch || 'main'}</span>
                                <span class="font-mono">${commitShort}</span>
                                <span>${date}</span>
                            </div>
                            ${hostname ? `
                                <div class="mt-3">
                                    <a href="http://${hostname}" target="_blank" 
                                       class="inline-flex items-center text-xs text-green-400 hover:text-green-300">
                                        <svg class="w-3 h-3 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"/>
                                        </svg>
                                        ${hostname}
                                    </a>
                                </div>
                            ` : ''}
                        </div>
                    </div>
                </div>
            `;
        }).join('');
    } catch (error) {
        console.error('Error loading deployments:', error);
        const container = document.getElementById('deployments');
        if (container) {
            container.innerHTML = `
                <div class="text-center py-8 border border-gray-800 rounded-lg bg-gray-900/50">
                    <p class="text-red-400 text-sm">Error loading deployments</p>
                </div>
            `;
        }
    }
}

async function refreshData() {
    await Promise.all([loadProjects(), loadDeployments()]);
}

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    if (!checkAuth()) return;
    
    loadUser();
    refreshData();

    // Logout button
    document.getElementById('logoutBtn')?.addEventListener('click', () => {
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        window.location.href = '/login';
    });

    // Refresh button
    document.getElementById('refreshBtn')?.addEventListener('click', refreshData);

    // Add project button
    document.getElementById('addProjectBtn')?.addEventListener('click', () => {
        alert('To add a project:\n1. Push code to a GitHub repository\n2. Configure the webhook URL in GitHub\n3. The project will be created automatically');
    });

    // Auto-refresh every 10 seconds
    setInterval(refreshData, 10000);
});

