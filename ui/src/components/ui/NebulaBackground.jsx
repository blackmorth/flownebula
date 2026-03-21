// src/components/ui/FormField.jsx
import { Box } from "@chakra-ui/react";

export default function NebulaBackground() {
    return (
        <Box
            position="fixed"
            inset="0"
            zIndex="-1"
            bgGradient="
                radial-gradient(circle at 20% 20%, rgba(138,77,255,0.25), transparent 60%),
                radial-gradient(circle at 80% 80%, rgba(100,50,200,0.15), transparent 70%),
                linear-gradient(180deg, #050510 0%, #0a0a15 100%)
              "
            _before={{
                content: '""',
                position: "absolute",
                inset: 0,
                backgroundImage: "url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0nMTAnIGhlaWdodD0nMTAnIHhtbG5zPSdodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2Zyc+PGNpcmNsZSBjeD0nNScgY3k9JzUnIHI9JzEnIGZpbGw9J3JnYmEoMjU1LDI1NSwyNTUsMC4wMyknIC8+PC9zdmc+')",
                opacity: 0.4,
            }}
        />
    );
}