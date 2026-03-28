import {
    Box,
    Button,
    Code,
    Container,
    Heading,
    HStack,
    List,
    ListItem,
    Text,
    VStack,
} from "@chakra-ui/react";
import { Link as RouterLink } from "react-router-dom";

const quickStartCommands = `git clone <URL_DU_REPO>
cd flownebula
docker compose up --build`;

const devCommands = `git clone <URL_DU_REPO>
cd flownebula
docker compose -f docker-compose.dev.yml up --build`;

export default function LocalInstallGuideView() {
    return (
        <Box minH="100vh" bg="bg" color="text" py={12}>
            <Container maxW="4xl">
                <VStack align="stretch" spacing={8}>
                    <VStack align="stretch" spacing={3}>
                        <Heading size="2xl" color="primary">
                            Guide pas à pas — Installer & utiliser Flow Nebula en local
                        </Heading>
                        <Text fontSize="md" opacity={0.85}>
                            Cette page est pensée pour un premier lancement sur un PC portable,
                            sans configuration complexe. Tu peux démarrer en mode “tout prêt”
                            avec Docker, puis passer au mode développement.
                        </Text>
                        <HStack spacing={3} pt={2}>
                            <Button as={RouterLink} to="/" variant="outline" borderColor="primary" color="primary">
                                Retour à l&apos;accueil
                            </Button>
                            <Button as={RouterLink} to="/register" bg="primary" color="white" _hover={{ bg: "primaryHover" }}>
                                Créer un compte Nebula
                            </Button>
                        </HStack>
                    </VStack>

                    <Box borderTopWidth="1px" borderColor="border" />

                    <Section title="1) Pré-requis (5 min)">
                        <List spacing={2} pl={4} styleType="disc">
                            <ListItem>Installer Docker Desktop (Windows/macOS) ou Docker Engine + Docker Compose (Linux).</ListItem>
                            <ListItem>Vérifier que Docker est bien lancé.</ListItem>
                            <ListItem>Avoir Git pour cloner le projet.</ListItem>
                        </List>
                        <Text mt={3} opacity={0.85}>
                            Vérification rapide dans un terminal :
                        </Text>
                        <Code display="block" whiteSpace="pre" p={4} mt={2} rounded="md">
                            docker --version{"\n"}
                            docker compose version{"\n"}
                            git --version
                        </Code>
                    </Section>

                    <Section title="2) Lancement simple (mode démo local)">
                        <Text opacity={0.9}>
                            C&apos;est le chemin recommandé pour démarrer vite : tous les services
                            principaux se lancent ensemble (agent, php, server, ui, prometheus, grafana).
                        </Text>
                        <Code display="block" whiteSpace="pre" p={4} mt={3} rounded="md">
                            {quickStartCommands}
                        </Code>
                        <Text mt={3}>Une fois démarré, ouvre :</Text>
                        <List spacing={2} pl={4} styleType="disc" mt={2}>
                            <ListItem>UI Nebula : <Code>http://localhost:8081</Code></ListItem>
                            <ListItem>API server health : <Code>http://localhost:8080/health</Code></ListItem>
                            <ListItem>Prometheus : <Code>http://localhost:9090</Code></ListItem>
                            <ListItem>Grafana : <Code>http://localhost:3000</Code></ListItem>
                        </List>
                    </Section>

                    <Section title="3) Mode développement (UI chaude + backend live reload)">
                        <Text opacity={0.9}>
                            Utilise cette variante si tu veux coder l&apos;UI React en live (Vite)
                            et redémarrer automatiquement le serveur Go.
                        </Text>
                        <Code display="block" whiteSpace="pre" p={4} mt={3} rounded="md">
                            {devCommands}
                        </Code>
                        <Text mt={3}>Accès principal en dev :</Text>
                        <List spacing={2} pl={4} styleType="disc" mt={2}>
                            <ListItem>Entrée unifiée via Caddy : <Code>http://localhost:8081</Code></ListItem>
                            <ListItem>UI Vite directe : <Code>http://localhost:5173</Code></ListItem>
                        </List>
                    </Section>

                    <Section title="4) Créer ton premier flux de données profiling">
                        <List spacing={2} pl={4} styleType="disc">
                            <ListItem>Créer un utilisateur depuis l&apos;interface (bouton “Créer un compte Nebula”).</ListItem>
                            <ListItem>Se connecter pour accéder au dashboard.</ListItem>
                            <ListItem>Vérifier que l&apos;agent envoie bien des données au serveur.</ListItem>
                            <ListItem>Consulter les sessions et ouvrir le détail (call tree, timeline, etc.).</ListItem>
                        </List>
                    </Section>

                    <Section title="5) Commandes utiles au quotidien">
                        <Code display="block" whiteSpace="pre" p={4} rounded="md">
                            docker compose ps{"\n"}
                            docker compose logs -f server{"\n"}
                            docker compose logs -f agent{"\n"}
                            docker compose down
                        </Code>
                    </Section>

                    <Section title="6) Dépannage express">
                        <List spacing={2} pl={4} styleType="disc">
                            <ListItem>
                                <strong>Port déjà pris</strong> : adapte les ports dans
                                <Code ml={1}>docker-compose.yml</Code> ou stoppe le process conflictuel.
                            </ListItem>
                            <ListItem>
                                <strong>UI inaccessible</strong> : vérifier d&apos;abord
                                <Code ml={1}>docker compose ps</Code> puis
                                <Code ml={1}>docker compose logs -f ui</Code>.
                            </ListItem>
                            <ListItem>
                                <strong>Pas de données de session</strong> : inspecter
                                <Code ml={1}>docker compose logs -f agent</Code> et l&apos;état de l&apos;API
                                <Code ml={1}>/health</Code>.
                            </ListItem>
                        </List>
                    </Section>
                </VStack>
            </Container>
        </Box>
    );
}

function Section({ title, children }) {
    return (
        <VStack align="stretch" spacing={3}>
            <Heading size="md" color="primary">{title}</Heading>
            {children}
        </VStack>
    );
}
