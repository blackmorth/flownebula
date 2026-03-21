// src/components/ui/FormField.jsx
import { VStack, Text, Input } from "@chakra-ui/react";

export default function FormField({ label, value, onChange, placeholder, type = "text", error }) {
    return (
        <VStack spacing={1} align="stretch" w="100%">
            <Text fontSize="sm" fontWeight="bold">{label}</Text>
            <Input
                type={type}
                value={value}
                onChange={onChange}
                placeholder={placeholder}
                w="100%"             // 🔹 Force la largeur du champ
                bg="bg.panel"
                borderColor="border"
                _focus={{
                    borderColor: "primary",
                    boxShadow: "0 0 0 1px var(--chakra-colors-primary)",
                }}
            />
            {error && <Text fontSize="xs" color="red.400">{error}</Text>}
        </VStack>
    );
}