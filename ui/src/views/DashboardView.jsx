import Layout from "../components/Layout";
import { Box, Heading, Text, Badge, Spinner } from "@chakra-ui/react";
import { Table } from "@chakra-ui/react/table";

export default function DashboardView({ loading, user, sessions }) {
    if (loading || !user) {
        return (
            <Layout>
                <Spinner size="lg" />
            </Layout>
        );
    }

    return (
        <Layout>
            <Heading mb={4}>Tableau de bord</Heading>
            <Text mb={8}>Connecté en tant que <strong>{user.email}</strong>.</Text>

            <Box>
                <Heading size="md" mb={4}>Dernières exécutions</Heading>
                {sessions.length === 0 ? (
                    <Text color="text.muted">Aucune session pour le moment.</Text>
                ) : (
                    <Table.Root size="md" variant="simple" bg="bg.subtle" p={6} rounded="md" shadow="lg">
                        <Table.Header>
                            <Table.Row>
                                <Table.ColumnHeader>ID</Table.ColumnHeader>
                                <Table.ColumnHeader>Type</Table.ColumnHeader>
                                <Table.ColumnHeader>Date</Table.ColumnHeader>
                            </Table.Row>
                        </Table.Header>
                        <Table.Body>
                            {sessions.map((s) => (
                                <Table.Row key={s.id} onClick={() => (window.location.href = `/sessions/${s.id}`)} _hover={{ bg: "bg.muted", cursor: "pointer" }}>
                                    <Table.Cell><Badge colorPalette="purple">{s.id}</Badge></Table.Cell>
                                    <Table.Cell>{s.agent_id || "local-php"}</Table.Cell>
                                    <Table.Cell>{new Date(s.created_at).toLocaleString()}</Table.Cell>
                                </Table.Row>
                            ))}
                        </Table.Body>
                    </Table.Root>
                )}
            </Box>
        </Layout>
    );
}
