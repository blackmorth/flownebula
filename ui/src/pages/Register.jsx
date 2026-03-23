import RegisterView from "../views/RegisterView";
import { useState } from "react";
import { api } from "../services/api";
import { useNavigate } from "react-router-dom";

export default function Register() {
    const nav = useNavigate();
    const [loading, setLoading] = useState(false);

    const handleRegister = async ({ email, password, confirm }) => {
        setLoading(true);

        // validation client
        const errors = {};
        if (!email) errors.email = "Email requis";
        if (!password) errors.password = "Mot de passe requis";
        if (password !== confirm) errors.confirm = "Les mots de passe ne correspondent pas";

        if (Object.keys(errors).length) {
            setLoading(false);
            return { fieldErrors: errors };
        }

        // appel API
        const res = await api("POST", "/auth/register", { email, password });
        setLoading(false);

        if (res.token) {
            localStorage.setItem("token", res.token);
            nav("/dashboard");
        }

        return res;
    };

    return <RegisterView loading={loading} onRegister={handleRegister} />;
}
