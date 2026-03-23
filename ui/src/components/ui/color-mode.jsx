import { IconButton } from '@chakra-ui/react'
import { useTheme } from 'next-themes'
import * as React from 'react'
import { LuMoon, LuSun } from 'react-icons/lu'

export function useColorMode() {
    const { resolvedTheme, setTheme } = useTheme()

    const toggleColorMode = () => {
        setTheme(resolvedTheme === 'dark' ? 'light' : 'dark')
    }

    return {
        colorMode: resolvedTheme,
        toggleColorMode,
    }
}

export function ColorModeButton(props) {
    const { toggleColorMode, colorMode } = useColorMode()

    return (
        <IconButton
            onClick={toggleColorMode}
            variant="ghost"
            aria-label="Toggle color mode"
            size="sm"
            {...props}
        >
            {colorMode === "dark" ? <LuMoon /> : <LuSun />}
        </IconButton>
    )
}