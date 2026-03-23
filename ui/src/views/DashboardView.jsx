import Layout from "../components/Layout";
import {
    Box, Heading, Text, Badge, Button, SimpleGrid,
    CardHeader, CardBody, Spinner, Stack
} from "@chakra-ui/react";
import { Card } from "@chakra-ui/react"
import { Table } from "@chakra-ui/react";

export default function DashboardView({ loading, user, sessions }) {
    if (loading || !user) {
        return (
            <Layout>
                <Spinner size="lg" />
            </Layout>
        );
    }

    const roles = Array.isArray(user.roles) ? user.roles : [];

    return (
        <Layout>
            <Heading mb={6} color="primary">
                Nebula Dashboard
            </Heading>

            <SimpleGrid columns={{ base: 1, md: 2 }} spacing={6}>
                <Card.Root shadow="lg">
                    <Card.Header>
                        <Heading size="md">Agent Status</Heading>
                    </Card.Header>

                    <Card.Body>
                        <Text mb={2}>
                            Statut :{" "}
                            <Badge
                                colorPalette={user.agent_enabled ? "green" : "red"}
                                variant="solid"
                            >
                                {user.agent_enabled ? "Activé" : "Désactivé"}
                            </Badge>
                        </Text>

                        <Text mb={4}>
                            Token :{" "}
                            <Badge colorPalette="purple">
                                {user.agent_token ? "Disponible" : "Aucun"}
                            </Badge>
                        </Text>

                        <Button
                            colorScheme="blue"
                            onClick={() => (window.location.href = "/settings/token")}
                        >
                            Voir le token
                        </Button>
                    </Card.Body>
                </Card.Root>
                {/* Agent Status */}
                {/*<Card shadow="lg">
                    <CardHeader>
                        <Heading size="md">Agent Status</Heading>
                    </CardHeader>
                    <CardBody>
                        <Text mb={2}>
                            Statut :{" "}
                            <Badge colorScheme={user.agent_enabled ? "green" : "red"}>
                                {user.agent_enabled ? "Activé" : "Désactivé"}
                            </Badge>
                        </Text>

                        <Text mb={4}>
                            Token :{" "}
                            <Badge colorScheme="purple">
                                {user.agent_token ? "Disponible" : "Aucun"}
                            </Badge>
                        </Text>

                        <Button
                            colorScheme="blue"
                            onClick={() => (window.location.href = "/settings/token")}
                        >
                            Voir le token
                        </Button>
                    </CardBody>
                </Card>

                {/* User Info */}
                {/* <Card shadow="lg">
                    <CardHeader>
                        <Heading size="md">Mon Compte</Heading>
                    </CardHeader>
                    <CardBody>
                        <Text>Email : {user.email}</Text>
                        <Text mt={2}>
                            Roles :{" "}
                            {Array.isArray(user?.roles) && user.roles.map(r => (
                                <Badge key={r} colorScheme="purple" mr={2}>{r}</Badge>
                            ))}
                        </Text>
                    </CardBody>
                </Card>*/}
            </SimpleGrid>

            {/* Recent Sessions */}
            <Box mt={10}>
                <Heading size="md" mb={4}>
                    Sessions Récentes
                </Heading>

                {sessions.length === 0 && (
                    <Text color="fg.muted">Aucune session pour le moment.</Text>
                )}

                {sessions.length > 0 && (
                    <Table.Root size="md" variant="simple" bg="bg.subtle" p={6} rounded="md" shadow="lg">
                        <Table.Header>
                            <Table.Row>
                                <Table.ColumnHeader>ID</Table.ColumnHeader>
                                <Table.ColumnHeader>Agent</Table.ColumnHeader>
                                <Table.ColumnHeader>Date</Table.ColumnHeader>
                            </Table.Row>
                        </Table.Header>

                        <Table.Body>
                            {sessions.map(s => (
                                <Table.Row
                                    key={s.id}
                                    onClick={() => (window.location.href = `/sessions/${s.id}`)}
                                    _hover={{ bg: "bg.muted", cursor: "pointer" }}
                                >
                                    <Table.Cell>
                                        <Badge colorScheme="purple">{s.id}</Badge>
                                    </Table.Cell>
                                    <Table.Cell>{s.agent_id}</Table.Cell>
                                    <Table.Cell>{new Date(s.created_at).toLocaleString()}</Table.Cell>
                                </Table.Row>
                            ))}
                        </Table.Body>
                    </Table.Root>

                )}
            </Box>

            {/* Quick Actions */}
            <Box mt={10}>
                <Heading size="md" mb={4}>
                    Actions Rapides
                </Heading>

                <Stack direction="row" spacing={4}>
                    <Button onClick={() => (window.location.href = "/sessions")}>
                        Voir les sessions
                    </Button>

                    <Button onClick={() => (window.location.href = "/settings/token")}>
                        Agent Token
                    </Button>

                    {Array.isArray(user?.roles) && user.roles.includes("ROLE_ADMIN") && (
                        <Button colorScheme="red" onClick={() => (window.location.href = "/admin/users")}>
                            Admin Panel
                        </Button>
                    )}

                </Stack>
            </Box>
        </Layout>
    );
}