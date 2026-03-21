import React from "react";
import ReactDOM from "react-dom/client";
import { ChakraProvider } from "@chakra-ui/react";
import { ThemeProvider } from "next-themes";
import { BrowserRouter } from "react-router-dom";
import { system } from "./components/ui/theme";
import App from "./App.jsx";

ReactDOM.createRoot(document.getElementById("root")).render(
    <React.StrictMode>
        <ChakraProvider value={system}>
            <ThemeProvider
                attribute="class"
                defaultTheme="dark"
                enableSystem={false}
            >
                <BrowserRouter>
                    <App />
                </BrowserRouter>
            </ThemeProvider>
        </ChakraProvider>
    </React.StrictMode>
);