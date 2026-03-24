import { Box } from "@chakra-ui/react";

export default function RawTab({ payload }) {
    return (
        <Box as="pre" p={4} bg="gray.800" color="gray.100" borderRadius="md" overflowX="auto">
            {JSON.stringify(payload, null, 2)}
        </Box>
    );
}
