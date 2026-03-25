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
import logo from "../assets/nebula-logo.svg";

export default function Layout({ children }) {
    const navigate = useNavigate();
    const location = useLocation();

    const logout = () => {
        localStorage.removeItem("token");
        navigate("/", { replace: true });
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
                _hover={{ bg: "bg.panel", color: "primary" }}
                transition="all 0.2s"
            >
                {label}
            </Link>
        );
    };

    return (
        <Flex minH="100vh" bg="bg" color="text">
            <Flex w="260px" direction="column" justify="space-between" bg="bg.subtle" borderRight="1px solid" borderColor="border" p={4}>
                <VStack align="stretch" spacing={6}>
                    <HStack px={2}>
                        <Image src={logo} alt="Flow Nebula" />
                    </HStack>

                    <VStack align="stretch" spacing={1}>
                        <NavItem to="/dashboard" label="Dashboard" />
                        <NavItem to="/scripts" label="Scripts PHP" />
                        <NavItem to="/sessions" label="Historique" />
                    </VStack>
                </VStack>

                <Button size="sm" variant="ghost" color="text.muted" _hover={{ bg: "bg.panel", color: "primary" }} onClick={logout}>
                    Déconnexion
                </Button>
            </Flex>

            <Flex direction="column" flex="1">
                <Flex h="64px" px={6} align="center" justify="space-between" borderBottom="1px solid" borderColor="border" bg="bg.subtle">
                    <Text fontSize="sm" color="text.muted">Exécution locale de scripts PHP</Text>
                </Flex>

                <Box p={8}>{children}</Box>
            </Flex>
        </Flex>
    );
}
