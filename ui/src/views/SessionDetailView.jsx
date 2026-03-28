import Layout from "../components/Layout";
import {
    Heading,
    Spinner,
    Text,
    Tabs,
    Box,
    SimpleGrid,
    Badge,
    VStack,
    HStack,
} from "@chakra-ui/react";
import { ChevronRight, TreePine, Flame, FileJson, Clock3 } from "lucide-react";

import OverviewTab from "./tabs/OverviewTab";
import CallTreeTab from "./tabs/CallTreeTab";
import FlamegraphTab from "./tabs/FlamegraphTab";
import RawTab from "./tabs/RawTab";
import CallFlowTab from "./tabs/CallFlowTab.jsx";
import TimelineTab from "./tabs/TimelineTab.jsx";

function parseTags(tags) {
    try {
        if (!tags) return {};
        if (typeof tags === "object") return tags;
        return JSON.parse(tags);
    } catch {
        return {};
    }
}

function buildExecutionContext(session) {
    const payload = session?.payload || {};
    const tags = parseTags(session?.tags);

    return {
        service: session?.service || payload?.service || tags?.service,
        endpoint: session?.endpoint || payload?.endpoint || tags?.endpoint,
        release: session?.release || payload?.release || tags?.release,
        environment: payload?.environment || tags?.environment || tags?.env,
        traceId: payload?.trace_id || payload?.traceId || tags?.trace_id || tags?.traceId,
        requestId: payload?.request_id || payload?.requestId || tags?.request_id || tags?.requestId,
        commitSha: payload?.commit_sha || payload?.commitSha || tags?.commit_sha || tags?.commitSha,
    };
}

export default function SessionDetailView({
    loading,
    session,
    sessions = [],
    baselineId = "",
    onChangeBaselineId,
    baselineSession = null,
}) {
    if (loading) {
        return (
            <Layout>
                <Spinner size="lg" />
            </Layout>
        );
    }

    if (!session) {
        return (
            <Layout>
                <Text>Session introuvable.</Text>
            </Layout>
        );
    }

    const payload = session.payload;
    const context = buildExecutionContext(session);

    const baselineCandidates = sessions.filter((item) => item.id !== session.id);

    return (
        <Layout>
            <Heading mb={6}>Session {session.id}</Heading>

            <VStack align="stretch" gap={4} mb={5}>
                <Box borderWidth="1px" rounded="md" p={4}>
                    <HStack justify="space-between" mb={3}>
                        <Text fontWeight="bold">Contexte d'exécution</Text>
                        <Badge colorPalette="purple">Agent {session.agent_id || "N/A"}</Badge>
                    </HStack>
                    <SimpleGrid columns={{ base: 1, md: 3 }} gap={3}>
                        <Box><Text fontSize="xs" color="gray.500">Service</Text><Text>{context.service || "—"}</Text></Box>
                        <Box><Text fontSize="xs" color="gray.500">Endpoint</Text><Text>{context.endpoint || "—"}</Text></Box>
                        <Box><Text fontSize="xs" color="gray.500">Release</Text><Text>{context.release || "—"}</Text></Box>
                        <Box><Text fontSize="xs" color="gray.500">Environment</Text><Text>{context.environment || "—"}</Text></Box>
                        <Box><Text fontSize="xs" color="gray.500">Trace ID</Text><Text>{context.traceId || "—"}</Text></Box>
                        <Box><Text fontSize="xs" color="gray.500">Request ID</Text><Text>{context.requestId || "—"}</Text></Box>
                        <Box><Text fontSize="xs" color="gray.500">Commit SHA</Text><Text>{context.commitSha || "—"}</Text></Box>
                        <Box><Text fontSize="xs" color="gray.500">Créée le</Text><Text>{new Date(session.created_at).toLocaleString()}</Text></Box>
                    </SimpleGrid>
                </Box>

                <Box borderWidth="1px" rounded="md" p={4}>
                    <Text fontWeight="bold" mb={2}>Comparaison baseline</Text>
                    <Text fontSize="sm" color="gray.500" mb={1}>Sélectionnez une session baseline</Text>
                    <Box as="select"
                        value={baselineId}
                        onChange={(event) => onChangeBaselineId?.(event.target.value)}
                        width="100%"
                        borderWidth="1px"
                        borderRadius="md"
                        px={3}
                        py={2}
                    >
                        <option value="">Aucune baseline sélectionnée</option>
                        {baselineCandidates.map((item) => (
                            <option key={item.id} value={item.id}>
                                {`#${item.id} • ${item.service || "service?"} • ${new Date(item.created_at).toLocaleString()}`}
                            </option>
                        ))}
                    </Box>
                </Box>
            </VStack>

            <Tabs.Root defaultValue="overview" variant="enclosed">
                <Tabs.List mb={4}>
                    <Tabs.Trigger value="overview">
                        <ChevronRight size={16} />
                        Overview
                    </Tabs.Trigger>

                    <Tabs.Trigger value="calltree">
                        <TreePine size={16} />
                        Call Tree
                    </Tabs.Trigger>

                    <Tabs.Trigger value="callflow">
                        <ChevronRight size={16} />
                        Call Flow
                    </Tabs.Trigger>
                    <Tabs.Trigger value="timeline">
                        <Clock3 size={16} />
                        Timeline
                    </Tabs.Trigger>

                    <Tabs.Trigger value="flamegraph">
                        <Flame size={16} />
                        Flamegraph
                    </Tabs.Trigger>

                    <Tabs.Trigger value="raw">
                        <FileJson size={16} />
                        Raw
                    </Tabs.Trigger>
                </Tabs.List>

                <Tabs.Content value="overview">
                    <OverviewTab payload={payload} baselinePayload={baselineSession?.payload || null} />
                </Tabs.Content>

                <Tabs.Content value="calltree">
                    <CallTreeTab payload={payload} />
                </Tabs.Content>

                <Tabs.Content value="callflow">
                    <CallFlowTab payload={payload} />
                </Tabs.Content>
                <Tabs.Content value="timeline">
                    <TimelineTab payload={payload} />
                </Tabs.Content>

                <Tabs.Content value="flamegraph">
                    <FlamegraphTab payload={payload} />
                </Tabs.Content>

                <Tabs.Content value="raw">
                    <RawTab payload={payload} />
                </Tabs.Content>
            </Tabs.Root>
        </Layout>
    );
}
