import { useState } from "react";
import { useNavigate, Link as RouterLink } from "react-router-dom";
import { api } from "../services/api";
import { Box, Button, Flex, Heading, Image, Text, Link, Alert } from "@chakra-ui/react";
import { Fieldset, Field, Input } from "@chakra-ui/react";
import NebulaBackground from "../components/ui/NebulaBackground";


export default function Register() {
    const nav = useNavigate();
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [confirm, setConfirm] = useState("");
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
        if (password !== confirm) errors.confirm = "Les mots de passe ne correspondent pas";
        if (Object.keys(errors).length) {
            setFieldErrors(errors);
            return;
        }

        // appel API
        const res = await api("POST", "/auth/register", { email, password });

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
                    Créer un compte
                </Heading>

                {error && <Alert status="error" mb={4} borderRadius="md"><Alert.Indicator />{error}</Alert>}

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
                                {fieldErrors.email && <Field.ErrorText>{fieldErrors.email}</Field.ErrorText>}
                            </Field.Root>

                            <Field.Root>
                                <Field.Label>Mot de passe</Field.Label>
                                <Input
                                    type="password"
                                    placeholder="Votre mot de passe"
                                    value={password}
                                    onChange={e => setPassword(e.target.value)}
                                />
                                {fieldErrors.password && <Field.ErrorText>{fieldErrors.password}</Field.ErrorText>}
                            </Field.Root>

                            <Field.Root>
                                <Field.Label>Confirmer le mot de passe</Field.Label>
                                <Input
                                    type="password"
                                    placeholder="Confirmez le mot de passe"
                                    value={confirm}
                                    onChange={e => setConfirm(e.target.value)}
                                />
                                {fieldErrors.confirm && <Field.ErrorText>{fieldErrors.confirm}</Field.ErrorText>}
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
                            S'inscrire
                        </Button>
                    </Fieldset.Root>
                </form>

                <Text textAlign="center" mt={6} opacity={0.8}>
                    Déjà un compte ?{" "}
                    <Link as={RouterLink} to="/login" color="primary">
                        Se connecter
                    </Link>
                </Text>
            </Box>
        </Flex>
        </>
    );
}