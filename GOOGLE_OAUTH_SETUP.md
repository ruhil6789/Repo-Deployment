# Google OAuth Setup Guide

## Step 1: Create Google OAuth Credentials

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Navigate to **APIs & Services** > **Credentials**
4. Click **Create Credentials** > **OAuth client ID**
5. If prompted, configure the OAuth consent screen:
   - User Type: **External** (for testing) or **Internal** (for G Suite)
   - App name: **Deploy Platform**
   - User support email: Your email
   - Developer contact: Your email
   - Click **Save and Continue**
   - Scopes: Add `openid`, `profile`, `email`
   - Click **Save and Continue**
   - Test users: Add your email (for testing)
   - Click **Save and Continue**

6. Create OAuth Client ID:
   - Application type: **Web application**
   - Name: **Deploy Platform Web Client**
   - Authorized JavaScript origins:
     - `http://localhost:8080` (for development)
     - `https://yourdomain.com` (for production)
   - Authorized redirect URIs:
     - `http://localhost:8080/auth/google/callback` (for development)
     - `https://yourdomain.com/auth/google/callback` (for production)
   - Click **Create**

7. Copy the **Client ID** and **Client Secret**

## Step 2: Add to .env File

Add these to your `.env` file:

```env
GOOGLE_CLIENT_ID=your-google-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_CALLBACK_URL=http://localhost:8080/auth/google/callback
```

## Step 3: Restart Server

Restart your server to load the new configuration:

```bash
go run cmd/api/main.go
```

## Testing

1. Go to `http://localhost:8080/login`
2. Click **Continue with Google**
3. Sign in with your Google account
4. You should be redirected to the dashboard

## Troubleshooting

- **"redirect_uri_mismatch"**: Make sure the redirect URI in Google Console exactly matches `GOOGLE_CALLBACK_URL`
- **"invalid_client"**: Check that `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` are correct
- **"access_denied"**: Make sure you've added your email as a test user in the OAuth consent screen
