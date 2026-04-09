import { createInjectionState } from '@vueuse/core'
import emblaCarouselVue from 'embla-carousel-vue'
import { onMounted, ref, watch } from 'vue'

const [useProvideCarousel, useInjectCarousel] = createInjectionState(
    (opts, plugins, orientation, emits) => {
        const [emblaNode, emblaApi] = emblaCarouselVue({
            ...opts,
            axis: orientation === 'horizontal' ? 'x' : 'y',
        }, plugins)

        const canScrollPrev = ref(false)
        const canScrollNext = ref(false)

        function scrollPrev() {
            emblaApi.value?.scrollPrev()
        }

        function scrollNext() {
            emblaApi.value?.scrollNext()
        }

        function onSelect(api) {
            canScrollPrev.value = api.canScrollPrev()
            canScrollNext.value = api.canScrollNext()
        }

        watch(emblaApi, (api) => {
            if (!api) return

            onSelect(api)
            api.on('select', onSelect)
            api.on('reInit', onSelect)

            emits('init-api', api)
        })

        return {
            carouselRef: emblaNode,
            carouselApi: emblaApi,
            canScrollPrev,
            canScrollNext,
            scrollPrev,
            scrollNext,
            orientation,
        }
    }
)

function useCarousel() {
    const carouselState = useInjectCarousel()
    if (!carouselState) {
        throw new Error('useCarousel must be used within a <Carousel />')
    }
    return carouselState
}

export { useCarousel, useProvideCarousel }
