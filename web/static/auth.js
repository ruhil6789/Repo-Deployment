// Auth page JavaScript
document.addEventListener('DOMContentLoaded', () => {
    const loginTab = document.getElementById('loginTab');
    const registerTab = document.getElementById('registerTab');
    const loginForm = document.getElementById('loginForm');
    const registerForm = document.getElementById('registerForm');
    const loginError = document.getElementById('loginError');
    const registerError = document.getElementById('registerError');

    // Tab switching
    loginTab.addEventListener('click', () => {
        loginTab.classList.add('text-blue-600', 'border-b-2', 'border-blue-600');
        loginTab.classList.remove('text-gray-500');
        registerTab.classList.remove('text-blue-600', 'border-b-2', 'border-blue-600');
        registerTab.classList.add('text-gray-500');
        loginForm.classList.remove('hidden');
        registerForm.classList.add('hidden');
    });

    registerTab.addEventListener('click', () => {
        registerTab.classList.add('text-blue-600', 'border-b-2', 'border-blue-600');
        registerTab.classList.remove('text-gray-500');
        loginTab.classList.remove('text-blue-600', 'border-b-2', 'border-blue-600');
        loginTab.classList.add('text-gray-500');
        registerForm.classList.remove('hidden');
        loginForm.classList.add('hidden');
    });

    // Login form
    loginForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        loginError.classList.add('hidden');
        
        const formData = new FormData(loginForm);
        const data = {
            email: formData.get('email'),
            password: formData.get('password')
        };

        try {
            const response = await fetch('/api/auth/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(data)
            });

            const result = await response.json();
            if (response.ok) {
                localStorage.setItem('token', result.token);
                localStorage.setItem('user', JSON.stringify(result.user));
                window.location.href = '/dashboard';
            } else {
                loginError.textContent = result.error || 'Login failed';
                loginError.classList.remove('hidden');
            }
        } catch (error) {
            loginError.textContent = 'Network error. Please try again.';
            loginError.classList.remove('hidden');
        }
    });

    // Register form
    registerForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        registerError.classList.add('hidden');
        
        const formData = new FormData(registerForm);
        const data = {
            username: formData.get('username'),
            email: formData.get('email'),
            password: formData.get('password')
        };

        try {
            const response = await fetch('/api/auth/register', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(data)
            });

            const result = await response.json();
            if (response.ok) {
                localStorage.setItem('token', result.token);
                localStorage.setItem('user', JSON.stringify(result.user));
                window.location.href = '/dashboard';
            } else {
                registerError.textContent = result.error || 'Registration failed';
                registerError.classList.remove('hidden');
            }
        } catch (error) {
            registerError.textContent = 'Network error. Please try again.';
            registerError.classList.remove('hidden');
        }
    });
});
