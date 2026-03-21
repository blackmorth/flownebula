import { useEffect, useState } from "react";
import { api } from "../services/api";
import Layout from "../components/Layout";
import {
    Box,
    Heading,
    Text,
    Badge,
    Table,
} from "@chakra-ui/react";

export default function Dashboard() {
    const [sessions, setSessions] = useState([]);

    useEffect(() => {
        const token = localStorage.getItem("token");
        api("GET", "/sessions", null, token).then(res => {
            setSessions(res || []);
        });
    }, []);

    return (
        <Layout>
            <Heading mb={6} color="primary">
                Nebula Dashboard
            </Heading>

            {sessions.length === 0 && (
                <Text fontSize="lg" color="fg.muted">
                    Aucune session pour le moment.
                </Text>
            )}

            {sessions.length > 0 && (
                <Table.Root bg="bg.subtle" p={6} rounded="md" shadow="lg">
                    <Table variant="simple">
                        <Table.Header>
                            <Table.Row>
                                <Table.ColumnHeader>ID</Table.ColumnHeader>
                                <Table.ColumnHeader>Agent</Table.ColumnHeader>
                                <Table.ColumnHeader>Date</Table.ColumnHeader>
                            </Table.Row>
                        </Table.Header>

                        <Table.Body>
                            {sessions.map(s => (
                                <Table.Row key={s.id}>
                                    <Table.Cell>
                                        <Badge colorScheme="purple">{s.id}</Badge>
                                    </Table.Cell>
                                    <Table.Cell>{s.agent_id}</Table.Cell>
                                    <Table.Cell>{new Date(s.created_at).toLocaleString()}</Table.Cell>
                                </Table.Row>
                            ))}
                        </Table.Body>
                    </Table>
                </Table.Root>
            )}
        </Layout>
    );
}