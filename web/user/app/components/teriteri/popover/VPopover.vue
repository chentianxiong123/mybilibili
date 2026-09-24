<template>
    <div @mouseleave="handleMouseLeave" style="position: relative; display: inline;">
        <div @mouseenter="handleMouseEnter" @click="handleClick" style="position: relative;" ref="vPopRef">
            <slot name="reference"></slot>
        </div>
        <div class="v-popover" :class="'to-' + placement" :style="popStyle">
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
            }
        },
        methods: {
            show() {
                this.popoverDisplay = "";
                this.isPopoverShow = true;
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
    position: absolute;
    transition: .3s;
    z-index: 3000;
}

.v-popover-content {
    background-color: #fff;
    border-radius: 8px;
    box-shadow: 0 0 30px rgba(0,0,0,.1);
    border: 1px solid var(--line_regular);
}

.to-bottom {
    top: 100%;
    left: 50%;
    padding-top: 5px;
    transform: translate3d(-50%,0,0);
}

.to-right {
    top: 50%;
    left: 100%;
    padding-left: 5px;
    transform: translate3d(0,-50%,0);
}

.to-top {
    bottom: 100%;
    left: 50%;
    padding-bottom: 5px;
    transform: translate3d(-50%,-100%,0);
}

.to-left {
    top: 50%;
    right: 100%;
    padding-right: 5px;
    transform: translate3d(-100%,-50%,0);
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
    0% { opacity: 0; transform: translate3d(-50%,-5px,0); }
    100% { opacity: 1; transform: translate3d(-50%,0,0); }
}

@keyframes fade-out-bottom {
    0% { opacity: 1; transform: translate3d(-50%,0,0); }
    100% { opacity: 0; transform: translate3d(-50%,-5px,0); }
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
    0% { opacity: 0; transform: translate3d(-5px,-50%,0); }
    100% { opacity: 1; transform: translate3d(0,-50%,0); }
}

@keyframes fade-out-right {
    0% { opacity: 1; transform: translate3d(0,-50%,0); }
    100% { opacity: 0; transform: translate3d(-5px,-50%,0); }
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
    0% { opacity: 0; transform: translate3d(-50%,5px,0); }
    100% { opacity: 1; transform: translate3d(-50%,0,0); }
}

@keyframes fade-out-top {
    0% { opacity: 1; transform: translate3d(-50%,0,0); }
    100% { opacity: 0; transform: translate3d(-50%,5px,0); }
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
    0% { opacity: 0; transform: translate3d(5px,-50%,0); }
    100% { opacity: 1; transform: translate3d(0,-50%,0); }
}

@keyframes fade-out-left {
    0% { opacity: 1; transform: translate3d(0,-50%,0); }
    100% { opacity: 0; transform: translate3d(5px,-50%,0); }
}
</style>
