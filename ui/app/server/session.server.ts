import { createCookieSessionStorage, redirect } from '@remix-run/node';

// Session configuration
const sessionStorage = createCookieSessionStorage({
  cookie: {
    name: 'vault0_session',
    httpOnly: true,
    path: '/',
    sameSite: 'lax',
    secrets: [process.env.SESSION_SECRET || 's3cr3t'],
    secure: process.env.NODE_ENV === 'production',
  },
});

// Session keys
const USER_SESSION_KEY = 'userId';
const TOKEN_SESSION_KEY = 'token';

// Get the user session
export async function getSession(request: Request) {
  const cookie = request.headers.get('Cookie');
  return sessionStorage.getSession(cookie);
}

// Create a new session with user info
export async function createUserSession({
  userId,
  token,
  redirectTo,
}: {
  userId: string;
  token: string;
  redirectTo: string;
}) {
  const session = await sessionStorage.getSession();
  session.set(USER_SESSION_KEY, userId);
  session.set(TOKEN_SESSION_KEY, token);
  
  return redirect(redirectTo, {
    headers: {
      'Set-Cookie': await sessionStorage.commitSession(session),
    },
  });
}

// Get the user ID from session
export async function getUserId(request: Request): Promise<string | null> {
  const session = await getSession(request);
  const userId = session.get(USER_SESSION_KEY);
  return userId || null;
}

// Get the auth token from session
export async function getToken(request: Request): Promise<string | null> {
  const session = await getSession(request);
  const token = session.get(TOKEN_SESSION_KEY);
  return token || null;
}

// Ensure user is authenticated
export async function requireUserId(request: Request): Promise<string> {
  const userId = await getUserId(request);
  
  if (!userId) {
    throw redirect('/login');
  }
  
  return userId;
}

// Ensure token is available
export async function requireToken(request: Request): Promise<string> {
  const token = await getToken(request);
  
  if (!token) {
    throw redirect('/login');
  }
  
  return token;
}

// Log out the user
export async function logout(request: Request) {
  const session = await getSession(request);
  
  return redirect('/', {
    headers: {
      'Set-Cookie': await sessionStorage.destroySession(session),
    },
  });
} 