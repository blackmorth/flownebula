import { useEffect, useState } from "react";
import SessionsView from "../views/SessionsView";
import Layout from "../components/Layout";
import { api } from "../services/api";

export default function Sessions() {
    const [loading, setLoading] = useState(true);
    const [sessions, setSessions] = useState([]);

    const token = localStorage.getItem("token");

    useEffect(() => {
        api("GET", "/sessions", null, token).then((res) => {
            setSessions(Array.isArray(res) ? res : []);
            setLoading(false);
        });
    }, []);

    return (
        <Layout>
            <SessionsView loading={loading} sessions={sessions} />
        </Layout>
    );
}
