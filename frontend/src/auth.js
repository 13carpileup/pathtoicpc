const authTokenKey = "pathtoicpc.authToken";
const authUserKey = "pathtoicpc.authUser";

export function getStoredAuthToken() {
  return localStorage.getItem(authTokenKey) || "";
}

export function getStoredUser() {
  try {
    return JSON.parse(localStorage.getItem(authUserKey));
  } catch {
    return null;
  }
}

export function storeUser(user) {
  localStorage.setItem(authUserKey, JSON.stringify(user));
}

export function authHeaders(token) {
  return token
    ? {
        Authorization: `Bearer ${token}`
      }
    : {};
}

export function jsonHeaders(token) {
  return {
    "Content-Type": "application/json",
    ...authHeaders(token)
  };
}

export async function authRequest(path, payload) {
  const response = await fetch(path, {
    method: "POST",
    headers: jsonHeaders(),
    body: JSON.stringify(payload)
  });

  const data = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error(data.error || "Authentication request failed.");
  }

  return data;
}

export function saveSession(data, setToken, setUser) {
  localStorage.setItem(authTokenKey, data.token);
  storeUser(data.user);
  setToken(data.token);
  setUser(data.user);
}

export function clearSession(setToken, setUser) {
  localStorage.removeItem(authTokenKey);
  localStorage.removeItem(authUserKey);
  setToken("");
  setUser(null);
}
