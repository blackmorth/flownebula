import LoginView from "../views/LoginView";
import { useState } from "react";
import { api } from "../services/api";
import { useNavigate } from "react-router-dom";

export default function Login() {
    const nav = useNavigate();
    const [loading, setLoading] = useState(false);

    const handleLogin = async ({ email, password }) => {
        setLoading(true);

        // validation client
        const errors = {};
        if (!email) errors.email = "Email requis";
        if (!password) errors.password = "Mot de passe requis";

        if (Object.keys(errors).length) {
            setLoading(false);
            return { fieldErrors: errors };
        }

        // appel API
        const res = await api("POST", "/auth/login", { email, password });
        setLoading(false);

        if (res.fieldErrors) return res;
        if (res.error) return res;

        if (res.token) {
            localStorage.setItem("token", res.token);
            nav("/dashboard");
        }

        return res;
    };

    return <LoginView loading={loading} onLogin={handleLogin} />;
}
