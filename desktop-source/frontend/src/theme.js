import { ref, watch, nextTick } from 'vue'

const isDark = ref(localStorage.getItem('theme') === 'dark')

// View Transition compatible theme toggle
const toggleTheme = (event) => {
    const isAppearanceTransition = document.startViewTransition
        && !window.matchMedia('(prefers-reduced-motion: reduce)').matches

    if (!isAppearanceTransition || !event) {
        isDark.value = !isDark.value
        return
    }

    const x = event.clientX
    const y = event.clientY
    const endRadius = Math.hypot(
        Math.max(x, innerWidth - x),
        Math.max(y, innerHeight - y)
    )

    const transition = document.startViewTransition(async () => {
        isDark.value = !isDark.value
        // Wait for the DOM to update
        await nextTick()
    })

    transition.ready.then(() => {
        const clipPath = [
            `circle(0px at ${x}px ${y}px)`,
            `circle(${endRadius}px at ${x}px ${y}px)`
        ]
        document.documentElement.animate(
            {
                clipPath: clipPath,
            },
            {
                duration: 400,
                easing: 'ease-out',
                pseudoElement: '::view-transition-new(root)',
            }
        )
    })
}

watch(isDark, (val) => {
    localStorage.setItem('theme', val ? 'dark' : 'light')
    if (val) {
        document.documentElement.classList.add('dark')
    } else {
        document.documentElement.classList.remove('dark')
    }
}, { immediate: true })

export { isDark, toggleTheme }
