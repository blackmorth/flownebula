import Layout from "../components/Layout";
import { Heading, Spinner, Text, Tabs, Box } from "@chakra-ui/react";
import { ChevronRight, TreePine, Flame, FileJson } from "lucide-react";

import OverviewTab from "./tabs/OverviewTab";
import CallTreeTab from "./tabs/CallTreeTab";
import FlamegraphTab from "./tabs/FlamegraphTab";
import RawTab from "./tabs/RawTab";

export default function SessionDetailView({ loading, session }) {
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

    return (
        <Layout>
            <Heading mb={6}>Session {session.id}</Heading>

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
                    <OverviewTab payload={payload} />
                </Tabs.Content>

                <Tabs.Content value="calltree">
                    <CallTreeTab payload={payload} />
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
