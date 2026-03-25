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
import { FiCpu, FiActivity, FiServer, FiTerminal } from "react-icons/fi";
import logo from "../assets/nebula-logo.svg";

export default function IndexView() {
    return (
        <Flex direction="column" align="center" minH="100vh" bg="bg" color="text" px={6} pt={20}>
            <Image src={logo} alt="Flow Nebula" height="90px" opacity={0.9} mb={6} />

            <VStack spacing={4} maxW="800px" textAlign="center">
                <Heading size="2xl" fontWeight="bold" color="primary">
                    Lancez vos scripts PHP et visualisez la sortie
                </Heading>

                <Text fontSize="lg" opacity={0.8} maxW="650px">
                    Version minimale: installation des outils, création de compte, connexion,
                    exécution de scripts PHP depuis l'UI et historique des sorties.
                </Text>

                <HStack spacing={4} mt={4}>
                    <Button as={RouterLink} to="/register" size="lg" bg="primary" color="white" _hover={{ bg: "primaryHover" }}>
                        S'inscrire
                    </Button>

                    <Button as={RouterLink} to="/login" size="lg" variant="outline" borderColor="primary" color="primary" _hover={{ bg: "primary", color: "white" }}>
                        Se connecter
                    </Button>
                </HStack>
            </VStack>

            <Box mt={20} maxW="900px" w="100%">
                <Heading size="md" mb={6} textAlign="center" color="primary">
                    Workflow simplifié
                </Heading>

                <SimpleGrid columns={{ base: 1, md: 4 }} spacing={6}>
                    <Feature icon={FiServer} title="Serveur API" desc="Authentification et sessions." />
                    <Feature icon={FiTerminal} title="Runner PHP" desc="Exécution directe de scripts PHP." />
                    <Feature icon={FiActivity} title="Historique" desc="Sorties conservées par session." />
                    <Feature icon={FiCpu} title="UI" desc="Vue simple pour lancer et consulter." />
                </SimpleGrid>
            </Box>

            <Text mt={20} opacity={0.4} fontSize="sm">
                © {new Date().getFullYear()} Flow Nebula
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
