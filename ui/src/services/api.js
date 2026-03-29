const API_URL = import.meta.env.VITE_API_URL || "";

export async function api(method, endpoint, data = null, token = null) {
    const headers = { "Content-Type": "application/json" };
    if (token) headers["Authorization"] = `Bearer ${token}`;

    let res;
    try {
        res = await fetch(API_URL + endpoint, {
            method,
            headers,
            body: data ? JSON.stringify(data) : undefined,
        });
    } catch (e) {
        return {
            error: "Impossible de joindre le serveur. Vérifiez votre connexion.",
            networkError: true,
        };
    }

    let json = {};
    try {
        json = await res.json();
    } catch (e) {
        // fallback si le serveur ne renvoie pas de JSON
        json = {};
    }

    // Si status >= 400, on renvoie quand même l'objet pour le front
    if (!res.ok) {
        return json; // ex: { error: "invalid credentials" }
    }

    return json;
}
