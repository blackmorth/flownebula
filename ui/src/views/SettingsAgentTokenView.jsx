import Layout from "../components/Layout";
import { Heading, Text, Spinner, Badge } from "@chakra-ui/react";

export default function SettingsAgentTokenView({ loading, token }) {
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

            <Text mt={4}>
                Token :{" "}
                <Badge colorScheme="purple">
                    {token ?? "Aucun token généré"}
                </Badge>
            </Text>
        </Layout>
    );
}
