import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { api } from "../services/api";
import SessionDetailView from "../views/SessionDetailView";

export default function SessionDetail() {
    const { id } = useParams();
    const [loading, setLoading] = useState(true);
    const [session, setSession] = useState(null);

    useEffect(() => {
        const token = localStorage.getItem("token");

        api("GET", `/sessions/${id}`, null, token).then(res => {
            setSession(res?.error ? null : res);
            setLoading(false);
        });
    }, [id]);

    return <SessionDetailView loading={loading} session={session} />;
}
