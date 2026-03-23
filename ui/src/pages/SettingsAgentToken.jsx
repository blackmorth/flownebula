import SettingsAgentTokenView from "../views/SettingsAgentTokenView";
import { useEffect, useState } from "react";
import { api } from "../services/api";

export default function SettingsAgentToken() {
    const [loading, setLoading] = useState(true);
    const [token, setToken] = useState(null);

    useEffect(() => {
        const t = localStorage.getItem("token");

        api("GET", "/auth/me", null, t).then(res => {
            setToken(res?.agent_token ?? null);
            setLoading(false);
        });
    }, []);

    return <SettingsAgentTokenView loading={loading} token={token} />;
}
