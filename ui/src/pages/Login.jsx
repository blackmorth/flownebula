import { Box, Button, Flex, Heading, VStack, Image, Text, Link, Alert } from "@chakra-ui/react";
import { Fieldset, Field, Input } from "@chakra-ui/react"; // v3
import { useState } from "react";
import { useNavigate, Link as RouterLink } from "react-router-dom";
import { api } from "../services/api";
import NebulaBackground from "../components/ui/NebulaBackground";

export default function Login() {
    const nav = useNavigate();
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [error, setError] = useState("");
    const [fieldErrors, setFieldErrors] = useState({});

    async function submit(e) {
        e.preventDefault();
        setError("");
        setFieldErrors({});

        // validation client
        const errors = {};
        if (!email) errors.email = "Email requis";
        if (!password) errors.password = "Mot de passe requis";

        if (Object.keys(errors).length) {
            setFieldErrors(errors);
            return;
        }

        // appel API
        const res = await api("POST", "/auth/login", { email, password });

        // erreurs serveur
        if (res.fieldErrors) setFieldErrors(res.fieldErrors);
        if (res.error) {
            setError(res.error);
            return;
        }

        if (res.token) {
            localStorage.setItem("token", res.token);
            nav("/dashboard");
        }
    }

    return (
        <>
            <NebulaBackground />

            <Flex minH="100vh" align="center" justify="center" bg="transparent" color="text" px={4}>
                <Box
                    bg="bg.subtle"
                    p={10}
                    borderRadius="lg"
                    border="1px solid"
                    borderColor="border"
                    w="100%"
                    maxW="420px"
                    boxShadow="lg"
                >
                    <Flex justify="center" mb={6}>
                        <Image src="/nebula-logo.svg" alt="Flow Nebula" height="60px" />
                    </Flex>

                    <Heading size="lg" textAlign="center" mb={6} color="primary">
                        Connexion
                    </Heading>

                    {error && (
                        <Alert status="error" mb={4} borderRadius="md">
                            <Alert.Indicator />
                            {error}
                        </Alert>
                    )}

                    <form onSubmit={submit}>
                        <Fieldset.Root size="lg" maxW="100%">
                            <Fieldset.Content>

                                <Field.Root>
                                    <Field.Label>Email</Field.Label>
                                    <Input
                                        type="email"
                                        placeholder="ex: quentin@mail.com"
                                        value={email}
                                        onChange={e => setEmail(e.target.value)}
                                    />
                                    {fieldErrors.email && (
                                        <Field.ErrorText>{fieldErrors.email}</Field.ErrorText>
                                    )}
                                </Field.Root>

                                <Field.Root>
                                    <Field.Label>Mot de passe</Field.Label>
                                    <Input
                                        type="password"
                                        placeholder="Votre mot de passe"
                                        value={password}
                                        onChange={e => setPassword(e.target.value)}
                                    />
                                    {fieldErrors.password && (
                                        <Field.ErrorText>{fieldErrors.password}</Field.ErrorText>
                                    )}
                                </Field.Root>

                            </Fieldset.Content>

                            <Button
                                type="submit"
                                mt={4}
                                w="100%"
                                bg="primary"
                                color="white"
                                _hover={{ bg: "primary.400" }}
                            >
                                Se connecter
                            </Button>
                        </Fieldset.Root>
                    </form>

                    <Text textAlign="center" mt={6} opacity={0.8}>
                        Pas de compte ?{" "}
                        <Link as={RouterLink} to="/register" color="primary">
                            Créer un compte
                        </Link>
                    </Text>
                </Box>
            </Flex>
        </>
    );
}
