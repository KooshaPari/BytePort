<script lang="ts">
    import { goto } from '$app/navigation';
    import tempUser from '../+layout.svelte';
    import { initializeUser, setUser, user } from '../../stores/user';
    import type { User } from '../../stores/user';

    let newUser: User;
    let Error: string = '';

    // Get base URL with correct port
    const getBaseUrl = () => {
        const url = new URL(window.location.href);
        url.port = '8081';
        // Use 0.0.0.0 for better mobile compatibility
        url.hostname = '0.0.0.0';
        return url.origin;
    };

    async function signUpUser() {
        let newUser = {
            Name: document.forms['regUser']['name'].value,
            Email: document.forms['regUser']['email'].value,
            Password: document.forms['regUser']['password'].value
        };

        const { Name, Email, Password } = newUser;
        try {
            const baseUrl = getBaseUrl();
            const response = await fetch(`${baseUrl}/signup`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ Name, Email, Password }),
                credentials: 'include'
            });

            console.log('Response Status:', response.status);
            console.log('Response OK:', response.ok);

            const data = await response.json();

            if (response.ok) {
                console.log('Signup successful:', data);
                await initializeUser();
                goto('/fts');
            } else {
                Error = data.message || data.error || 'An unknown error occurred';
                console.log('Signup failed:', Error);
            }
        } catch (err) {
            console.error('Error during signup:', err);
            Error = 'An error occurred during signup.';
        }
    }
</script>