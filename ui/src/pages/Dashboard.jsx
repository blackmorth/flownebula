import DashboardView from "../views/DashboardView";
import { useEffect, useState } from "react";
import { api } from "../services/api";

export default function Dashboard() {
    const [loading, setLoading] = useState(true);
    const [sessions, setSessions] = useState([]);
    const [user, setUser] = useState(null);

    useEffect(() => {
        const token = localStorage.getItem("token");

        Promise.all([
            api("GET", "/auth/me", null, token),
            api("GET", "/sessions", null, token)
        ]).then(([me, sess]) => {
            setUser(me?.error ? null : me);
            setSessions(Array.isArray(sess) ? sess : []);
            setLoading(false);
        });
    }, []);

    return (
        <DashboardView
            loading={loading}
            user={user}
            sessions={sessions}
        />
    );
}
