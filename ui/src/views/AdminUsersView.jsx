import {
    Heading,
    Text,
    Badge,
    Button,
    HStack,
    Spinner,
} from "@chakra-ui/react";
import { Table } from "@chakra-ui/react/table";

export default function AdminUsersView({
                                           loading,
                                           users,
                                           enableAgent,
                                           disableAgent,
                                           regenerateToken,
                                       }) {
    return (
        <>
            <Heading mb={6}>Users</Heading>

            {loading && (
                <HStack>
                    <Spinner />
                    <Text>Chargement des utilisateurs…</Text>
                </HStack>
            )}

            {!loading && users.length === 0 && (
                <Text color="text.muted">Aucun utilisateur trouvé.</Text>
            )}

            {!loading && users.length > 0 && (
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
                            <Table.ColumnHeader>Email</Table.ColumnHeader>
                            <Table.ColumnHeader>Roles</Table.ColumnHeader>
                            <Table.ColumnHeader>Agent</Table.ColumnHeader>
                            <Table.ColumnHeader>Actions</Table.ColumnHeader>
                        </Table.Row>
                    </Table.Header>

                    <Table.Body>
                        {users.map((u) => (
                            <Table.Row key={u.id}>
                                <Table.Cell>
                                    <Badge colorPalette="purple">{u.id}</Badge>
                                </Table.Cell>

                                <Table.Cell>{u.email}</Table.Cell>

                                <Table.Cell>
                                    <HStack>
                                        {u.roles.map((r) => (
                                            <Badge key={r} colorScheme="blue">
                                                {r}
                                            </Badge>
                                        ))}
                                    </HStack>
                                </Table.Cell>

                                <Table.Cell>
                                    <Badge
                                        colorPalette={u.agent_enabled ? "green" : "red"}
                                        variant="solid"
                                    >
                                        {u.agent_enabled ? "Enabled" : "Disabled"}
                                    </Badge>
                                </Table.Cell>

                                <Table.Cell>
                                    <HStack spacing={2}>
                                        {u.agent_enabled ? (
                                            <Button
                                                size="sm"
                                                colorScheme="red"
                                                onClick={() => disableAgent(u.id)}
                                            >
                                                Disable
                                            </Button>
                                        ) : (
                                            <Button
                                                size="sm"
                                                colorScheme="green"
                                                onClick={() => enableAgent(u.id)}
                                            >
                                                Enable
                                            </Button>
                                        )}

                                        <Button
                                            size="sm"
                                            colorScheme="purple"
                                            onClick={() => regenerateToken(u.id)}
                                        >
                                            Regenerate Token
                                        </Button>
                                    </HStack>
                                </Table.Cell>
                            </Table.Row>
                        ))}
                    </Table.Body>
                </Table.Root>
            )}
        </>
    );
}
