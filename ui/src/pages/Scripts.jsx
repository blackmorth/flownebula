import { useState } from "react";
import Layout from "../components/Layout";
import ScriptsView from "../views/ScriptsView";
import { api } from "../services/api";

export default function Scripts() {
    const [loading, setLoading] = useState(false);
    const [result, setResult] = useState(null);

    const runScript = async ({ path, args }) => {
        setLoading(true);
        const token = localStorage.getItem("token");
        const res = await api("POST", "/scripts/run", { path, args }, token);
        setResult(res);
        setLoading(false);
        return res;
    };

    return (
        <Layout>
            <ScriptsView loading={loading} result={result} onRun={runScript} />
        </Layout>
    );
}
