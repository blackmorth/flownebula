import { createSystem, defaultConfig } from "@chakra-ui/react";

export const system = createSystem(defaultConfig, {
    theme: {
        tokens: {
            colors: {
                nebula: {
                    100: { value: "#C9A7FF" },
                    200: { value: "#A87BFF" },
                    300: { value: "#8A4DFF" },
                    400: { value: "#6A1BFF" },
                    500: { value: "#4B00E6" },
                },

                // 🔥 palette perf (ultra utile pour profiler)
                perf: {
                    low: { value: "#22c55e" },   // vert
                    medium: { value: "#f59e0b" },// orange
                    high: { value: "#ef4444" },  // rouge
                },
            },

            fonts: {
                heading: { value: "Space Grotesk, sans-serif" },
                body: { value: "Inter, sans-serif" },
                mono: { value: "JetBrains Mono, monospace" },
            },

            shadows: {
                glow: { value: "0 0 20px rgba(138,77,255,0.35)" },
            },
        },

        semanticTokens: {
            colors: {
                bg: {
                    value: {
                        _light: "#ffffff",
                        _dark: "#0b0618",
                    },
                },

                "bg.subtle": {
                    value: {
                        _light: "#f8f7ff",
                        _dark: "#140d2e",
                    },
                },

                "bg.panel": {
                    value: {
                        _light: "#ffffff",
                        _dark: "#1a1238",
                    },
                },

                text: {
                    value: {
                        _light: "#1a1a1a",
                        _dark: "#f5f5f5",
                    },
                },

                "text.muted": {
                    value: {
                        _light: "#666",
                        _dark: "#aaa",
                    },
                },

                border: {
                    value: {
                        _light: "#e5e7eb",
                        _dark: "#2a1f55",
                    },
                },

                primary: {
                    value: {
                        _light: "{colors.nebula.400}",
                        _dark: "{colors.nebula.300}",
                    },
                },

                // 🔥 glow dynamique
                "primary.glow": {
                    value: {
                        _light: "0 0 0 rgba(0,0,0,0)",
                        _dark: "0 0 25px rgba(138,77,255,0.6)",
                    },
                },
            },
        },
    },
});