import Layout from "../components/Layout";
import { Badge, Box, Heading, Spinner, Text, VStack } from "@chakra-ui/react";

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

    const payload = session.payload || {};

    return (
        <Layout>
            <VStack align="stretch" spacing={4}>
                <Heading>Session #{session.id}</Heading>
                <Text>Script: <strong>{payload.script_path || "-"}</strong></Text>
                <Text>Code de sortie: <Badge colorPalette={payload.exit_code === 0 ? "green" : "red"}>{payload.exit_code}</Badge></Text>
                {payload.error && <Text color="red.400">Erreur: {payload.error}</Text>}
                <Text>Sortie:</Text>
                <Box as="pre" whiteSpace="pre-wrap" p={4} rounded="md" bg="bg.subtle" border="1px solid" borderColor="border">
                    {payload.output || "(aucune sortie)"}
                </Box>
            </VStack>
        </Layout>
    );
}
