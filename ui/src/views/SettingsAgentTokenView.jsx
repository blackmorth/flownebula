import { useState } from "react";
import Layout from "../components/Layout";
import { Badge, Button, Heading, Spinner, Text, VStack } from "@chakra-ui/react";

export default function SettingsAgentTokenView({ loading, token }) {
    const [copyState, setCopyState] = useState("idle");

    async function copyToken() {
        if (!token) return;

        try {
            await navigator.clipboard.writeText(token);
            setCopyState("success");
            setTimeout(() => setCopyState("idle"), 2000);
        } catch (e) {
            setCopyState("error");
            setTimeout(() => setCopyState("idle"), 2500);
        }
    }

    if (loading) {
        return (
            <Layout>
                <Spinner size="lg" />
            </Layout>
        );
    }

    return (
        <Layout>
            <Heading>Agent Token</Heading>

            <VStack align="start" mt={4} gap={3}>
                <Text>
                    Token :{" "}
                    <Badge colorScheme="purple">
                        {token ?? "Aucun token généré"}
                    </Badge>
                </Text>

                <Button onClick={copyToken} disabled={!token} size="sm" variant="outline">
                    Copier le token
                </Button>

                {copyState === "success" && (
                    <Text fontSize="sm" color="green.500">Token copié ✅</Text>
                )}
                {copyState === "error" && (
                    <Text fontSize="sm" color="red.500">
                        Impossible de copier automatiquement (permissions navigateur).
                    </Text>
                )}
            </VStack>
        </Layout>
    );
}
