// Hook to handle fullscreen functionality
import { useState, useEffect, useCallback } from 'react'

export function useFullscreen() {
    const [isFullscreen, setIsFullscreen] = useState(false)

    // Check fullscreen status
    const checkFullscreen = useCallback(() => {
        setIsFullscreen(!!document.fullscreenElement)
    }, [])

    // Request fullscreen
    const enterFullscreen = useCallback(() => {
        const element = document.documentElement
        if (element.requestFullscreen) {
            element.requestFullscreen()
        } else if ((element as any).webkitRequestFullscreen) {
            (element as any).webkitRequestFullscreen()
        } else if ((element as any).msRequestFullscreen) {
            (element as any).msRequestFullscreen()
        }
    }, [])

    // Exit fullscreen
    const exitFullscreen = useCallback(() => {
        if (document.exitFullscreen) {
            document.exitFullscreen()
        } else if ((document as any).webkitExitFullscreen) {
            (document as any).webkitExitFullscreen()
        } else if ((document as any).msExitFullscreen) {
            (document as any).msExitFullscreen()
        }
    }, [])

    // Toggle fullscreen
    const toggleFullscreen = useCallback(() => {
        if (isFullscreen) {
            exitFullscreen()
        } else {
            enterFullscreen()
        }
    }, [isFullscreen, enterFullscreen, exitFullscreen])

    // Listen for fullscreen changes
    useEffect(() => {
        const handleFullscreenChange = () => {
            checkFullscreen()
        }

        document.addEventListener('fullscreenchange', handleFullscreenChange)
        document.addEventListener('webkitfullscreenchange', handleFullscreenChange)
        document.addEventListener('msfullscreenchange', handleFullscreenChange)

        return () => {
            document.removeEventListener('fullscreenchange', handleFullscreenChange)
            document.removeEventListener('webkitfullscreenchange', handleFullscreenChange)
            document.removeEventListener('msfullscreenchange', handleFullscreenChange)
        }
    }, [checkFullscreen])

    // Initial check
    useEffect(() => {
        checkFullscreen()
    }, [checkFullscreen])

    return {
        isFullscreen,
        toggleFullscreen,
        enterFullscreen,
        exitFullscreen
    }
}
