<template>
    <div @mouseleave="handleMouseLeave" style="position: relative;">
        <div @mouseenter="handleMouseEnter" @click="handleClick" style="position: relative;" ref="vPopRef">
            <slot name="reference"></slot>
        </div>
        <Teleport to="body">
            <div class="v-popover" :class="'to-' + placement" :style="mergedPopStyle" ref="vPopEl">
                <div
                    class="v-popover-content"
                    ref="vPopCon"
                    :class="isPopoverShow ? 'popShow-' + placement : 'popHide-' + placement"
                    :style="{ display: popoverDisplay }"
                    @mouseenter="handlePopoverEnter"
                    @mouseleave="handlePopoverLeave"
                >
                    <slot name="content"></slot>
                </div>
            </div>
        </Teleport>
    </div>
</template>

<script lang="ts">
let inTimer;

    export default {
        name: "VPopover",
        props: {
            placement: {
                type: String,
                default() {
                    return "bottom";
                }
            },
            trigger: {
                type: String,
                default() {
                    return "hover";
                }
            },
            popStyle: {
                type: String,
                default() {
                    return "";
                }
            }
        },
        data() {
            return {
                popoverDisplay: "none",
                isPopoverShow: false,
                computedStyle: "",
            }
        },
        computed: {
            mergedPopStyle() {
                return `${this.computedStyle};${this.popStyle}`;
            }
        },
        methods: {
            computePosition() {
                const ref = this.$refs.vPopRef as HTMLElement | undefined;
                if (!ref) return;
                const rect = ref.getBoundingClientRect();
                const pop = this.$refs.vPopEl as HTMLElement | undefined;
                const pw = pop ? pop.offsetWidth : 0;
                const ph = pop ? pop.offsetHeight : 0;
                const gap = 5;
                let top = '0px', left = '0px';
                if (this.placement === 'bottom') {
                    top = `${rect.bottom + gap}px`;
                    left = `${rect.left + rect.width / 2 - pw / 2}px`;
                } else if (this.placement === 'top') {
                    top = `${rect.top - ph - gap}px`;
                    left = `${rect.left + rect.width / 2 - pw / 2}px`;
                } else if (this.placement === 'right') {
                    top = `${rect.top + rect.height / 2 - ph / 2}px`;
                    left = `${rect.right + gap}px`;
                } else if (this.placement === 'left') {
                    top = `${rect.top + rect.height / 2 - ph / 2}px`;
                    left = `${rect.left - pw - gap}px`;
                }
                this.computedStyle = `position: fixed; top: ${top}; left: ${left}; z-index: 3000;`;
            },
            show() {
                this.popoverDisplay = "";
                this.isPopoverShow = true;
                this.$nextTick(() => {
                    this.computePosition();
                });
            },
            hide() {
                this.isPopoverShow = false;
                setTimeout(() => {
                    this.popoverDisplay = "none";
                }, 300);
            },

            handleMouseEnter() {
                if (this.trigger === "hover") {
                    clearTimeout(inTimer);
                    inTimer = setTimeout(() => {
                        this.show();
                    }, 100);
                }
            },
            handleMouseLeave() {
                if (this.trigger === "hover") {
                    clearTimeout(inTimer);
                    this.hide();
                }
            },
            handlePopoverEnter() {
                if (this.trigger === "hover") {
                    clearTimeout(inTimer);
                }
            },
            handlePopoverLeave() {
                if (this.trigger === "hover") {
                    inTimer = setTimeout(() => {
                        this.hide();
                    }, 200);
                }
            },
            handleClick() {
                if (this.trigger === "click") {
                    if (this.isPopoverShow) {
                        this.hide();
                    } else {
                        this.show();
                    }
                }
            },
            handleOutsideClick(event) {
                const vPopRef = this.$refs.vPopRef;
                const vPopCon = this.$refs.vPopCon;
                if (vPopRef && !vPopRef.contains(event.target) && vPopCon && !vPopCon.contains(event.target)) {
                    this.hide();
                }
            },
        },
        mounted() {
            if (this.trigger === 'click') {
                window.addEventListener("click", this.handleOutsideClick);
            }
        },
        beforeUnmount() {
            clearTimeout(inTimer);
            if (this.trigger === 'click') {
                window.removeEventListener("click", this.handleOutsideClick);
            }
        }
    }
</script>

<style scoped>
.v-popover {
    transition: .3s;
}

.v-popover-content {
    background-color: #fff;
    border-radius: 8px;
    box-shadow: 0 0 30px rgba(0,0,0,.1);
    border: 1px solid var(--line_regular);
}

.popHide-bottom {
    animation: fade-out-bottom 0.2s ease-out forwards;
    transform-origin: top;
}

.popShow-bottom {
    animation: fade-in-bottom 0.2s ease-out forwards;
    transform-origin: top;
}

@keyframes fade-in-bottom {
    0% { opacity: 0; transform: translateY(-5px); }
    100% { opacity: 1; transform: translateY(0); }
}

@keyframes fade-out-bottom {
    0% { opacity: 1; transform: translateY(0); }
    100% { opacity: 0; transform: translateY(-5px); }
}

.popHide-right {
    animation: fade-out-right 0.2s ease-out forwards;
    transform-origin: left;
}

.popShow-right {
    animation: fade-in-right 0.2s ease-out forwards;
    transform-origin: left;
}

@keyframes fade-in-right {
    0% { opacity: 0; transform: translateX(-5px); }
    100% { opacity: 1; transform: translateX(0); }
}

@keyframes fade-out-right {
    0% { opacity: 1; transform: translateX(0); }
    100% { opacity: 0; transform: translateX(-5px); }
}

.popHide-top {
    animation: fade-out-top 0.2s ease-out forwards;
    transform-origin: bottom;
}

.popShow-top {
    animation: fade-in-top 0.2s ease-out forwards;
    transform-origin: bottom;
}

@keyframes fade-in-top {
    0% { opacity: 0; transform: translateY(5px); }
    100% { opacity: 1; transform: translateY(0); }
}

@keyframes fade-out-top {
    0% { opacity: 1; transform: translateY(0); }
    100% { opacity: 0; transform: translateY(5px); }
}

.popHide-left {
    animation: fade-out-left 0.2s ease-out forwards;
    transform-origin: right;
}

.popShow-left {
    animation: fade-in-left 0.2s ease-out forwards;
    transform-origin: right;
}

@keyframes fade-in-left {
    0% { opacity: 0; transform: translateX(5px); }
    100% { opacity: 1; transform: translateX(0); }
}

@keyframes fade-out-left {
    0% { opacity: 1; transform: translateX(0); }
    100% { opacity: 0; transform: translateX(5px); }
}
</style>
