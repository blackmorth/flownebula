import {
    Box,
    Button,
    Flex,
    Heading,
    Text,
    VStack,
    Image,
    HStack,
    SimpleGrid,
    Icon,
} from "@chakra-ui/react";
import { Link as RouterLink } from "react-router-dom";
import { FiCpu, FiActivity, FiServer, FiGitBranch } from "react-icons/fi";
import logo from "../assets/nebula-logo.svg";

export default function IndexView() {
    return (
        <Flex direction="column" align="center" minH="100vh" bg="bg" color="text" px={6} pt={20}>

            {/* Logo */}
            <Image src={logo} alt="Flow Nebula" height="90px" opacity={0.9} mb={6} />

            {/* Hero */}
            <VStack spacing={4} maxW="800px" textAlign="center">
                <Heading size="2xl" fontWeight="bold" color="primary">
                    Profiling & Monitoring distribués pour PHP
                </Heading>

                <Text fontSize="lg" opacity={0.8} maxW="650px">
                    Flow Nebula connecte un <strong>probe PHP natif</strong>, un <strong>agent Go</strong>&nbsp;
                    et un <strong>serveur centralisé</strong> pour offrir une visibilité complète sur vos performances,
                    avec un minimum overhead et sans dépendance réseau dans le hot‑path.
                </Text>

                <HStack spacing={4} mt={4} flexWrap="wrap" justify="center">
                    <Button
                        as={RouterLink}
                        to="/guide/local-install"
                        size="lg"
                        bg="primary"
                        color="white"
                        _hover={{ bg: "primaryHover" }}
                    >
                        Guide d&apos;installation locale
                    </Button>

                    <Button
                        as={RouterLink}
                        to="/register"
                        size="lg"
                        variant="outline"
                        borderColor="primary"
                        color="primary"
                        _hover={{ bg: "primary", color: "white" }}
                    >
                        Créer un compte
                    </Button>

                    <Button
                        as={RouterLink}
                        to="/login"
                        size="lg"
                        variant="ghost"
                        color="text.muted"
                        _hover={{ bg: "bg.subtle", color: "primary" }}
                    >
                        Voir la démo
                    </Button>
                </HStack>
            </VStack>

            {/* Pipeline */}
            <Box mt={20} maxW="900px" w="100%">
                <Heading size="md" mb={6} textAlign="center" color="primary">
                    Une architecture pensée pour la performance
                </Heading>

                <SimpleGrid columns={{ base: 1, md: 4 }} spacing={6}>
                    <Feature icon={FiCpu} title="Probe PHP (C)" desc="Minimum overhead. Activation instantanée via token." />
                    <Feature icon={FiGitBranch} title="Agent Go" desc="Validation des tokens, sessions, sampling." />
                    <Feature icon={FiServer} title="Serveur Nebula" desc="Stockage sécurisé des utilisateurs & sessions." />
                    <Feature icon={FiActivity} title="UI Nebula" desc="Visualisation claire, moderne et temps réel." />
                </SimpleGrid>
            </Box>

            {/* Footer */}
            <Text mt={20} opacity={0.4} fontSize="sm">
                © {new Date().getFullYear()} Flow Nebula — Open-source Profiling Platform
            </Text>
        </Flex>
    );
}

function Feature({ icon, title, desc }) {
    return (
        <VStack
            bg="bg.subtle"
            p={6}
            borderRadius="lg"
            border="1px solid"
            borderColor="border"
            spacing={3}
            shadow="md"
            _hover={{ shadow: "lg", transform: "translateY(-4px)" }}
            transition="all 0.2s"
            textAlign="center"
        >
            <Icon as={icon} boxSize={8} color="primary" />
            <Heading size="sm" color="primary">{title}</Heading>
            <Text opacity={0.8} fontSize="sm">{desc}</Text>
        </VStack>
    );
}
