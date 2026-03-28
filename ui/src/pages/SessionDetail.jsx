import { useEffect, useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import { api } from "../services/api";
import SessionDetailView from "../views/SessionDetailView";

export default function SessionDetail() {
    const { id } = useParams();
    const [loading, setLoading] = useState(true);
    const [session, setSession] = useState(null);
    const [sessions, setSessions] = useState([]);
    const [baselineId, setBaselineId] = useState("");

    useEffect(() => {
        const token = localStorage.getItem("token");

        Promise.all([
            api("GET", `/sessions/${id}`, null, token),
            api("GET", "/sessions", null, token),
        ]).then(([current, allSessions]) => {
            setSession(current?.error ? null : current);
            setSessions(Array.isArray(allSessions) ? allSessions : []);
            setLoading(false);
        });
    }, [id]);

    const baselineSession = useMemo(
        () => sessions.find((item) => String(item.id) === String(baselineId)) || null,
        [baselineId, sessions],
    );

    return (
        <SessionDetailView
            loading={loading}
            session={session}
            sessions={sessions}
            baselineId={baselineId}
            onChangeBaselineId={setBaselineId}
            baselineSession={baselineSession}
        />
    );
}
