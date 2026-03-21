import {
    Box,
    Flex,
    VStack,
    Text,
    Link,
    Button,
    HStack,
    Image,
} from "@chakra-ui/react";
import { Link as RouterLink, useNavigate, useLocation } from "react-router-dom";

export default function Layout({ children }) {
    const navigate = useNavigate();
    const location = useLocation();

    const logout = () => {
        // 1. Supprime le token
        localStorage.removeItem("token");

        // 2. Redirige vers la page d'accueil
        navigate("/", { replace: true }); // replace pour ne pas pouvoir revenir en arrière
    };

    const NavItem = ({ to, label }) => {
        const isActive = location.pathname.startsWith(to);

        return (
            <Link
                as={RouterLink}
                to={to}
                w="full"
                px={3}
                py={2}
                rounded="md"
                fontWeight={isActive ? "semibold" : "normal"}
                bg={isActive ? "bg.panel" : "transparent"}
                color={isActive ? "primary" : "text.muted"}
                _hover={{
                    bg: "bg.panel",
                    color: "primary",
                }}
                transition="all 0.2s"
            >
                {label}
            </Link>
        );
    };

    return (
        <Flex minH="100vh" bg="bg" color="text">
            {/* 🌌 Sidebar */}
            <Flex
                w="260px"
                direction="column"
                justify="space-between"
                bg="bg.subtle"
                borderRight="1px solid"
                borderColor="border"
                p={4}
            >
                <VStack align="stretch" spacing={6}>
                    {/* Logo */}
                    <HStack px={2}>
                        <Image src="/nebula-logo.svg" alt="Flow Nebula" />
                    </HStack>

                    {/* Nav */}
                    <VStack align="stretch" spacing={1}>
                        <NavItem to="/dashboard" label="Dashboard" />
                        <NavItem to="/sessions" label="Sessions" />
                        <NavItem to="/settings" label="Settings" />
                    </VStack>
                </VStack>

                {/* Logout */}
                <Button
                    size="sm"
                    variant="ghost"
                    color="text.muted"
                    _hover={{
                        bg: "bg.panel",
                        color: "primary",
                    }}
                    onClick={logout}
                >
                    Déconnexion
                </Button>
            </Flex>

            {/* 📊 Main */}
            <Flex direction="column" flex="1">
                {/* Header */}
                <Flex
                    h="64px"
                    px={6}
                    align="center"
                    justify="space-between"
                    borderBottom="1px solid"
                    borderColor="border"
                    bg="bg.subtle"
                >
                    {/* Left */}
                    <Text fontSize="sm" color="text.muted">
                        Monitoring & Profiling Platform
                    </Text>

                    {/* Right */}
                    {/*<HStack spacing={3}>
                        <Box
                            w="8px"
                            h="8px"
                            bg="perf.low"
                            rounded="full"
                        />
                        <Text fontSize="xs" color="text.muted">
                            Server OK
                        </Text>
                    </HStack>*/}
                </Flex>

                {/* Content */}
                <Box p={8}>
                    {children}
                </Box>
            </Flex>
        </Flex>
    );
}