import {
    Flex,
    Box,
    VStack,
    HStack,
    Text,
    Link,
    Button,
    Image,
} from "@chakra-ui/react";
import { Link as RouterLink, useNavigate, useLocation } from "react-router-dom";
import logo from "../assets/nebula-logo.svg";

export default function AdminLayout({ children }) {
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
            {/* Sidebar Admin */}
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
                        <Image src={logo} alt="Flow Nebula Admin" />
                    </HStack>

                    {/* Admin Navigation */}
                    <VStack align="stretch" spacing={1}>
                        <Text
                            fontSize="xs"
                            textTransform="uppercase"
                            color="text.muted"
                            px={2}
                            mb={1}
                        >
                            Administration
                        </Text>
                        <NavItem to="/dashboard" label="User Mode" />
                        <NavItem to="/admin/users" label="Users" />
                        <NavItem to="/admin/agents" label="Agents" />
                        <NavItem to="/admin/sessions" label="Sessions" />
                        <NavItem to="/admin/metrics" label="Metrics" />
                        <NavItem to="/admin/logs" label="Logs" />
                        <NavItem to="/admin/settings" label="Settings" />
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

            {/* Main Admin Content */}
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
                    <Text fontSize="sm" color="text.muted">
                        Flow Nebula — Admin Console
                    </Text>
                </Flex>

                {/* Content */}
                <Box p={8}>{children}</Box>
            </Flex>
        </Flex>
    );
}
