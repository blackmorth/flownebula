import {
    Heading,
    Text,
    Badge,
    Spinner,
    HStack,
} from "@chakra-ui/react";
import { Table } from "@chakra-ui/react/table";

export default function SessionsView({ loading, sessions }) {
    return (
        <>
            <Heading mb={6}>Mes Sessions</Heading>

            {loading && (
                <HStack>
                    <Spinner />
                    <Text>Chargement des sessions…</Text>
                </HStack>
            )}

            {!loading && sessions.length === 0 && (
                <Text color="text.muted">Aucune session trouvée.</Text>
            )}

            {!loading && sessions.length > 0 && (
                <Table.Root
                    size="md"
                    variant="simple"
                    bg="bg.subtle"
                    p={6}
                    rounded="md"
                    shadow="lg"
                >
                    <Table.Header>
                        <Table.Row>
                            <Table.ColumnHeader>ID</Table.ColumnHeader>
                            <Table.ColumnHeader>Agent</Table.ColumnHeader>
                            <Table.ColumnHeader>Date</Table.ColumnHeader>
                        </Table.Row>
                    </Table.Header>

                    <Table.Body>
                        {sessions.map((s) => (
                            <Table.Row
                                key={s.id}
                                onClick={() => (window.location.href = `/sessions/${s.id}`)}
                                _hover={{ bg: "bg.muted", cursor: "pointer" }}
                            >
                                <Table.Cell>
                                    <Badge colorPalette="purple">{s.id}</Badge>
                                </Table.Cell>
                                <Table.Cell>{s.agent_id}</Table.Cell>
                                <Table.Cell>
                                    {new Date(s.created_at).toLocaleString()}
                                </Table.Cell>
                            </Table.Row>
                        ))}
                    </Table.Body>
                </Table.Root>
            )}
        </>
    );
}
