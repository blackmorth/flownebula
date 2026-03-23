import Layout from "../components/Layout";
import { Heading, Spinner, Text } from "@chakra-ui/react";

export default function SessionDetailView({ loading, session }) {
    if (loading) {
        return (
            <Layout>
                <Spinner size="lg" />
            </Layout>
        );
    }

    if (!session) {
        return (
            <Layout>
                <Text>Session introuvable.</Text>
            </Layout>
        );
    }

    return (
        <Layout>
            <Heading mb={6}>Session {session.id}</Heading>

            <Text>Agent : {session.agent_id ?? "—"}</Text>
            <Text>Date : {session.created_at ? new Date(session.created_at).toLocaleString() : "—"}</Text>
        </Layout>
    );
}
