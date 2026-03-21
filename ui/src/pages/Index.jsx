import { Box, Button, Flex, Heading, Text, VStack, Image, HStack } from "@chakra-ui/react";
import { Link as RouterLink } from "react-router-dom";
import { useEffect } from "react";
import { useNavigate } from "react-router-dom";

export default function IndexPage() {
    const navigate = useNavigate();

    useEffect(() => {
        if (localStorage.getItem("token")) {
            navigate("/dashboard");
        }
    }, []);
    return (
        <Flex
            direction="column"
            align="center"
            justify="center"
            minH="100vh"
            bg="bg"
            color="text"
            px={6}
        >
            {/* Logo */}
            <Image
                src="/nebula-logo.svg"
                alt="Flow Nebula"
                height="90px"
                mb={6}
            />

            {/* Hero */}
            <VStack spacing={4} maxW="700px" textAlign="center">
                <Heading size="2xl" fontWeight="700" color="primary">
                    Flow Nebula
                </Heading>

                <Text fontSize="xl" opacity={0.8}>
                    La plateforme moderne de <strong>profiling</strong> et <strong>monitoring</strong>
                    pour vos applications Go, pensée pour la performance, la clarté et la scalabilité.
                </Text>

                <HStack spacing={4} mt={4}>
                    <Button
                        as={RouterLink}
                        to="/register"
                        size="lg"
                        bg="primary"
                        color="white"
                        _hover={{ bg: "primaryHover" }}
                    >
                        Commencer
                    </Button>

                    <Button
                        as={RouterLink}
                        to="/login"
                        size="lg"
                        variant="outline"
                        borderColor="primary"
                        color="primary"
                        _hover={{ bg: "primary", color: "white" }}
                    >
                        Se connecter
                    </Button>
                </HStack>
            </VStack>

            {/* Features */}
            <Flex mt={20} gap={12} wrap="wrap" justify="center">
                <Feature title="Monitoring en temps réel" desc="CPU, RAM, sessions, agents — tout en un coup d'œil." />
                <Feature title="Profiling avancé" desc="Analysez vos performances avec précision et simplicité." />
                <Feature title="Sessions intelligentes" desc="Chaque session est historisée, visualisée et exploitable." />
            </Flex>

            {/* Footer */}
            <Text mt={20} opacity={0.4} fontSize="sm">
                © {new Date().getFullYear()} Flow Nebula — Crafted for developers
            </Text>
        </Flex>
    );
}

function Feature({ title, desc }) {
    return (
        <VStack
            bg="bg.subtle"
            p={6}
            borderRadius="lg"
            border="1px solid"
            borderColor="border"
            maxW="260px"
            spacing={3}
        >
            <Heading size="md" color="primary">{title}</Heading>
            <Text opacity={0.8}>{desc}</Text>
        </VStack>
    );
}
