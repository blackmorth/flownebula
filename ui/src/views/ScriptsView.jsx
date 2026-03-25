import { useState } from "react";
import { Badge, Box, Button, Heading, Input, Text, VStack } from "@chakra-ui/react";

export default function ScriptsView({ loading, result, onRun }) {
    const [path, setPath] = useState("test.php");
    const [args, setArgs] = useState("");

    const handleSubmit = async (e) => {
        e.preventDefault();
        const parsedArgs = args.trim() ? args.trim().split(/\s+/) : [];
        await onRun({ path, args: parsedArgs });
    };

    const data = result?.result;

    return (
        <VStack align="stretch" spacing={6}>
            <Heading>Lancer un script PHP</Heading>
            <Box as="form" onSubmit={handleSubmit} display="grid" gap={3} maxW="700px">
                <Input value={path} onChange={(e) => setPath(e.target.value)} placeholder="Chemin du script .php" />
                <Input value={args} onChange={(e) => setArgs(e.target.value)} placeholder="Arguments (optionnel)" />
                <Button type="submit" loading={loading} colorPalette="blue" width="fit-content">
                    Exécuter
                </Button>
            </Box>

            {result?.error && <Text color="red.400">Erreur: {result.error}</Text>}

            {data && (
                <Box border="1px solid" borderColor="border" rounded="md" p={4} bg="bg.subtle">
                    <Text mb={2}>Script: <strong>{data.script_path}</strong></Text>
                    <Text mb={2}>Code de sortie: <Badge colorPalette={data.exit_code === 0 ? "green" : "red"}>{data.exit_code}</Badge></Text>
                    {data.error && <Text mb={2} color="red.400">Erreur d'exécution: {data.error}</Text>}
                    <Text mb={2}>Sortie:</Text>
                    <Box as="pre" whiteSpace="pre-wrap" fontSize="sm" p={3} bg="blackAlpha.400" rounded="md">{data.output || "(aucune sortie)"}</Box>
                    {result?.session?.id && <Text mt={3} fontSize="sm">Session enregistrée: #{result.session.id}</Text>}
                </Box>
            )}
        </VStack>
    );
}
